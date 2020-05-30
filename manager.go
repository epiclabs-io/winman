package winman

import (
	"sync"

	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

type WindowEdge int16

// Available mouse actions.
const (
	EdgeNone WindowEdge = iota
	EdgeTop
	EdgeRight
	EdgeBottom
	EdgeLeft
	EdgeBottomRight
	EdgeBottomLeft
)

const WindowZTop = -1
const WindowZBottom = 0

var MinWindowWidth = 3
var MinWindowHeight = 3

// flexItem holds layout options for one item.

func inRect(wnd Window, x, y int) bool {
	return NewRect(wnd.GetRect()).Contains(x, y)
}

type Manager struct {
	*tview.Box

	// The windows to be positioned.
	windows Stack

	dragOffsetX, dragOffsetY int
	draggedWindow            Window
	draggedEdge              WindowEdge
	sync.Mutex
}

func NewWindowManager() *Manager {
	wm := &Manager{
		Box: tview.NewBox(),
	}
	return wm
}

func (wm *Manager) NewWindow() *WindowBase {
	wnd := NewWindow()
	wm.AddWindow(wnd)
	return wnd
}

func (wm *Manager) AddWindow(window Window) *Manager {
	wm.Lock()
	defer wm.Unlock()
	wm.windows.Push(window)
	return wm
}

func (wm *Manager) RemoveWindow(window Window) *Manager {
	wm.Lock()
	defer wm.Unlock()
	wm.windows.Remove(window)
	return wm
}

func (wm *Manager) Center(window Window) *Manager {
	mx, my, mw, mh := wm.GetInnerRect()
	x, y, width, height := window.GetRect()
	x = mx + (mw-width)/2
	y = my + (mh-height)/2
	window.SetRect(x, y, width, height)
	return wm
}

func (wm *Manager) WindowCount() int {
	wm.Lock()
	defer wm.Unlock()
	return len(wm.windows)
}

func (wm *Manager) Window(i int) Window {
	wm.Lock()
	defer wm.Unlock()
	wnd, _ := wm.windows.Item(i).(Window)
	return wnd
}

func (wm *Manager) getZ(window Window) int {
	return wm.windows.IndexOf(window)
}
func (wm *Manager) GetZ(window Window) int {
	wm.Lock()
	defer wm.Unlock()
	return wm.getZ(window)
}

func (wm *Manager) setZ(window Window, newZ int) {
	wm.windows.Move(window, newZ)
}

func (wm *Manager) SetZ(window Window, newZ int) *Manager {
	wm.Lock()
	defer wm.Unlock()
	wm.setZ(window, newZ)
	return wm
}

// Focus is called when this primitive receives focus.
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

		// reduce window that are too wide,
		// or fix size if the window is maximized
		if w > mw || window.IsMaximized() {
			w = mw
			x = mx
		}

		// reduce window that are too tall,
		// or fix size if the window is maximized
		if h > mh || window.IsMaximized() {
			h = mh
			y = my
		}

		// Avoid window overflowing on the right
		if x+w > mx+mw {
			x = mx + mw - w
		}

		// Avoid window overflowing on the bottom
		if y+h > my+mh {
			y = my + mh - h
		}

		// reposition window to the fixed coordinates:
		window.SetRect(x, y, w, h)

		// now we can draw it
		window.Draw(screen)
	}
}

// MouseHandler returns the mouse handler for this primitive.
func (wm *Manager) MouseHandler() func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
	return wm.WrapMouseHandler(func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
		if !wm.InRect(event.Position()) {
			return false, nil
		}
		wm.Lock()

		if wm.draggedWindow != nil {
			switch action {
			case tview.MouseLeftUp:
				wm.draggedWindow = nil
			case tview.MouseMove:
				x, y := event.Position()
				wx, wy, ww, wh := wm.draggedWindow.GetRect()
				if wm.draggedEdge == EdgeTop && wm.draggedWindow.IsDraggable() {
					wm.draggedWindow.SetRect(x-wm.dragOffsetX, y-wm.dragOffsetY, ww, wh)
				} else {
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
					wm.draggedWindow = window
					wm.dragOffsetX = x - wx
					wm.dragOffsetY = y - wy
					wm.Unlock()
					return true, nil
				}
			}
			wm.Unlock()
			return window.MouseHandler()(action, event, setFocus)
		}
		wm.Unlock()

		return
	})
}
