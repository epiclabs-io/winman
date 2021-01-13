package winman

import (
	"sync"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// WindowEdge enumerates the different window edges and corners
type WindowEdge int16

// Different window edges
const (
	EdgeNone WindowEdge = iota
	EdgeTop
	EdgeRight
	EdgeBottom
	EdgeLeft
	EdgeBottomRight
	EdgeBottomLeft
)

// WindowZTop is used with SetZ to move a window to the top
const WindowZTop = -1

// WindowZBottom is used with SetZ to move a window to the bottom
const WindowZBottom = 0

// MinWindowWidth sets the minimum width a window can have as part of a window manager
var MinWindowWidth = 3

// MinWindowHeight sets the minimum height a window can have as part of a window manager
var MinWindowHeight = 3

// inRect returns true if the given coordinates are within the window
func inRect(wnd Window, x, y int) bool {
	return NewRect(wnd.GetRect()).Contains(x, y)
}

// Manager represents a Window Manager primitive
type Manager struct {
	*tview.Box

	// The windows to be positioned.
	windows Stack

	dragOffsetX, dragOffsetY int
	draggedWindow            Window
	draggedEdge              WindowEdge
	sync.Mutex
}

// NewWindowManager returns a ready to use window manager
func NewWindowManager() *Manager {
	wm := &Manager{
		Box: tview.NewBox(),
	}
	return wm
}

// NewWindow creates a new (hidden) window and adds it to this window manager
func (wm *Manager) NewWindow() *WindowBase {
	wnd := NewWindow()
	wm.AddWindow(wnd)
	return wnd
}

// AddWindow adds the given window to the window manager
func (wm *Manager) AddWindow(window Window) *Manager {
	wm.Lock()
	defer wm.Unlock()
	wm.windows.Push(window)
	return wm
}

// RemoveWindow removes the given window from this window manager
func (wm *Manager) RemoveWindow(window Window) *Manager {
	wm.Lock()
	defer wm.Unlock()
	wm.windows.Remove(window)
	return wm
}

// Center centers the given window relative to the window manager
func (wm *Manager) Center(window Window) *Manager {
	mx, my, mw, mh := wm.GetInnerRect()
	_, _, width, height := window.GetRect()
	x := mx + (mw-width)/2
	y := my + (mh-height)/2
	window.SetRect(x, y, width, height)
	return wm
}

// WindowCount returns the number of windows managed by this window manager
func (wm *Manager) WindowCount() int {
	wm.Lock()
	defer wm.Unlock()
	return len(wm.windows)
}

// Window returns the window at the given z index
func (wm *Manager) Window(z int) Window {
	wm.Lock()
	defer wm.Unlock()
	wnd, _ := wm.windows.Item(z).(Window)
	return wnd
}

func (wm *Manager) getZ(window Window) int {
	return wm.windows.IndexOf(window)
}

// GetZ returns the z index of the given window
// returns -1 if the given window is not part of this manager
func (wm *Manager) GetZ(window Window) int {
	wm.Lock()
	defer wm.Unlock()
	return wm.getZ(window)
}

func (wm *Manager) setZ(window Window, newZ int) {
	wm.windows.Move(window, newZ)
}

// SetZ moves the given window to the given z index
// The special constants WindowZTop and WindowZBottom can be used
func (wm *Manager) SetZ(window Window, newZ int) *Manager {
	wm.Lock()
	defer wm.Unlock()
	wm.setZ(window, newZ)
	return wm
}

// Focus is called when this primitive receives focus
// implements tview.Primitive.Focus
func (wm *Manager) Focus(delegate func(p tview.Primitive)) {
	wm.Lock()

	window, _ := wm.windows.Find(func(wi interface{}) bool {
		return wi.(Window).IsVisible()
	}).(Window)

	if window != nil {
		wm.Unlock()
		window.Focus(delegate)
		return
	}
	wm.Unlock()
}

// HasFocus returns whether or not this primitive has focus.
// implements tview.Focusable
func (wm *Manager) HasFocus() bool {
	wm.Lock()
	defer wm.Unlock()
	// iterate over all windows. If any has focus, then the
	// this window manager has focus.
	return nil != wm.windows.Find(func(wi interface{}) bool {
		return wi.(Window).HasFocus()
	})
}

// Draw draws this primitive onto the screen.
// implements tview.Primitive.Draw
func (wm *Manager) Draw(screen tcell.Screen) {
	wm.Box.Draw(screen)

	wm.Lock()
	defer wm.Unlock()

	// Ensure that the window with focus has the highest Z-index:
	topWindowIndex := len(wm.windows) - 1
	for i := topWindowIndex; i >= 0; i-- {
		window := wm.windows[i].(Window)
		if window.IsVisible() && window.HasFocus() {
			if i < topWindowIndex {
				wm.setZ(window, WindowZTop) // move focused window on top
			}
			break
		}
	}

	// make sure windows are not out of bounds, too small,
	// or too big to fit within the window manager:
	for _, wndItem := range wm.windows {
		window := wndItem.(Window)
		if !window.IsVisible() {
			continue
		}
		mx, my, mw, mh := wm.GetInnerRect()
		x, y, w, h := window.GetRect()

		// Avoid window overflowing on the left:
		if x < mx {
			x = mx
		}

		// Avoid window overflowing on the top:
		if y < my {
			y = my
		}

		// Fix window that is too narrow:
		if w < MinWindowWidth {
			w = MinWindowWidth
		}

		// Fix window that is too short:
		if h < MinWindowHeight {
			h = MinWindowHeight
		}

		// reduce windows that are too wide,
		// or fix size if the window is maximized
		if w > mw || window.IsMaximized() {
			w = mw
			x = mx
		}

		// reduce windows that are too tall,
		// or fix size if the window is maximized
		if h > mh || window.IsMaximized() {
			h = mh
			y = my
		}

		// Avoid window overflowing the right edge
		if x+w > mx+mw {
			x = mx + mw - w
		}

		// Avoid window overflowing the bottom edge
		if y+h > my+mh {
			y = my + mh - h
		}

		// reposition window to the new coordinates:
		window.SetRect(x, y, w, h)

		// now we can draw it
		window.Draw(screen)
	}
}

// MouseHandler returns the mouse handler for this primitive.
// implements tview.Primitive.MouseHandler
func (wm *Manager) MouseHandler() func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
	return wm.WrapMouseHandler(func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
		// ignore mouse events out of the bounds of the window manager
		if !wm.InRect(event.Position()) {
			return false, nil
		}
		wm.Lock()

		// check if there is an active drag operation:
		if wm.draggedWindow != nil {
			switch action {
			case tview.MouseLeftUp:
				wm.draggedWindow = nil // if the button is released, stop the drag operation
			case tview.MouseMove:
				x, y := event.Position()
				wx, wy, ww, wh := wm.draggedWindow.GetRect()
				// depending if the drag operation is on the top or edges, either move the window or resize
				if wm.draggedEdge == EdgeTop && wm.draggedWindow.IsDraggable() {
					wm.draggedWindow.SetRect(x-wm.dragOffsetX, y-wm.dragOffsetY, ww, wh) // move window
				} else {
					// resize window pulling from the corresponding edge
					if wm.draggedWindow.IsResizable() {
						switch wm.draggedEdge {
						case EdgeRight:
							wm.draggedWindow.SetRect(wx, wy, x-wx+1, wh)
						case EdgeBottom:
							wm.draggedWindow.SetRect(wx, wy, ww, y-wy+1)
						case EdgeLeft:
							wm.draggedWindow.SetRect(x, wy, ww+wx-x, wh)
						case EdgeBottomRight:
							wm.draggedWindow.SetRect(wx, wy, x-wx+1, y-wy+1)
						case EdgeBottomLeft:
							wm.draggedWindow.SetRect(x, wy, ww+wx-x, y-wy+1)
						}
					}
				}
				wm.Unlock()
				return true, nil
			}
		}

		lastModal := false
		// Pass mouse events along to the window with highest Z
		// that is hit by the mouse
		// Stop if the last window was a modal.
		for i := len(wm.windows) - 1; i >= 0 && !lastModal; i-- {
			window := wm.windows[i].(Window)
			if !window.IsVisible() { // skip hidden windows
				continue
			}

			// if this is a modal window, then don't give a chance for
			// other windows to get mouse events
			lastModal = window.IsModal() // if true, will exit loop on the next iteration

			x, y := event.Position()
			if !inRect(window, x, y) {
				// skip this window since it is not hit
				continue
			}

			if action == tview.MouseLeftDown && window.HasBorder() {
				// initiate a drag operation
				if !window.HasFocus() {
					setFocus(window)
				}
				wx, wy, ww, wh := window.GetRect()
				wm.draggedEdge = EdgeNone
				switch {
				case y == wy+wh-1:
					switch {
					case x == wx:
						wm.draggedEdge = EdgeBottomLeft
					case x == wx+ww-1:
						wm.draggedEdge = EdgeBottomRight
					default:
						wm.draggedEdge = EdgeBottom
					}
				case x == wx:
					wm.draggedEdge = EdgeLeft
				case x == wx+ww-1:
					wm.draggedEdge = EdgeRight
				case y == wy:
					wm.draggedEdge = EdgeTop
				}
				if wm.draggedEdge != EdgeNone {
					// drag detected. Remember where the drag operation started
					wm.draggedWindow = window
					wm.dragOffsetX = x - wx
					wm.dragOffsetY = y - wy
					wm.Unlock()
					return true, nil
				}
			}
			wm.Unlock()
			// no drag operation detected.
			// pass the mouse events to the window itself.
			return window.MouseHandler()(action, event, setFocus)
		}
		wm.Unlock()

		return
	})
}

// InputHandler returns a handler which receives key events when it has focus.
func (wm *Manager) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return wm.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		wm.Lock()
		// Pass key events along to the window with highest Z that is visible and has focus
		var window Window
		for i := len(wm.windows) - 1; i >= 0; i-- {
			window = wm.windows[i].(Window)
			if window.HasFocus() {
				break
			}
			window = nil
		}
		wm.Unlock()
		if window != nil {
			inputHandler := window.InputHandler()
			if inputHandler != nil {
				inputHandler(event, setFocus)
			}
		}
	})
}
