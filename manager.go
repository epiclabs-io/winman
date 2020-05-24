package winman

import (
	"sync"

	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

type WindowEdge int16

// Available mouse actions.
const (
	WindowEdgeNone WindowEdge = iota
	WindowEdgeTop
	WindowEdgeRight
	WindowEdgeBottom
	WindowEdgeLeft
	WindowEdgeBottomRight
	WindowEdgeBottomLeft
)

const WindowZTop = -1
const WindowZBottom = 0

const minWindowWidth = 3
const minWindowHeight = 3

// flexItem holds layout options for one item.

type Manager struct {
	*tview.Box

	// The windows to be positioned.
	windows []*Window

	mouseWindow              *Window
	dragOffsetX, dragOffsetY int
	draggedWindow            *Window
	draggedEdge              WindowEdge
	modalWindow              *Window
	sync.Mutex
}

func NewWindowManager() *Manager {
	wm := &Manager{
		Box: tview.NewBox(),
	}
	return wm
}

// NewWindow creates a new window in this window manager
func (wm *Manager) NewWindow() *Window {
	window := NewWindow()
	window.manager = wm
	return window
}

func (wm *Manager) Show(window *Window) *Manager {
	wm.Lock()
	defer wm.Unlock()
	for _, wnd := range wm.windows {
		if wnd == window {
			return wm
		}
	}
	window.manager = wm
	wm.windows = append(wm.windows, window)
	return wm
}

func (wm *Manager) ShowModal(window *Window) *Manager {
	wm.Show(window)
	wm.Lock()
	defer wm.Unlock()
	wm.modalWindow = window
	return wm
}

func (wm *Manager) Hide(window *Window) *Manager {
	wm.Lock()
	defer wm.Unlock()
	if window == wm.modalWindow {
		wm.modalWindow = nil
	}
	for i, wnd := range wm.windows {
		if wnd == window {
			wm.windows = append(wm.windows[:i], wm.windows[i+1:]...)
			break
		}
	}
	return wm
}

func (wm *Manager) FindPrimitive(p tview.Primitive) *Window {
	wm.Lock()
	defer wm.Unlock()
	for _, window := range wm.windows {
		if window.root == p {
			return window
		}
	}
	return nil
}

func (wm *Manager) WindowCount() int {
	wm.Lock()
	defer wm.Unlock()
	return len(wm.windows)
}

func (wm *Manager) Window(i int) *Window {
	wm.Lock()
	defer wm.Unlock()
	if i < 0 || i >= len(wm.windows) {
		return nil
	}
	return wm.windows[i]
}

func (wm *Manager) getZ(window *Window) int {
	for i, wnd := range wm.windows {
		if wnd == window {
			return i
		}
	}
	return -1
}
func (wm *Manager) GetZ(window *Window) int {
	wm.Lock()
	defer wm.Unlock()
	return wm.getZ(window)
}

func (wm *Manager) setZ(window *Window, newZ int) {
	oldZ := wm.getZ(window)
	lenW := len(wm.windows)
	if oldZ == -1 {
		return
	}

	if newZ < 0 || newZ >= lenW {
		newZ = lenW - 1
	}

	newWindows := make([]*Window, lenW)
	for i, j := 0, 0; i < lenW; j++ {
		if j == oldZ {
			j++
		}
		if i == newZ {
			j--
		} else {
			newWindows[i] = wm.windows[j]
		}
		i++
	}

	newWindows[newZ] = window
	wm.windows = newWindows
}

func (wm *Manager) SetZ(window *Window, newZ int) *Manager {
	wm.Lock()
	defer wm.Unlock()
	wm.setZ(window, newZ)
	return wm
}

// Focus is called when this primitive receives focus.
func (wm *Manager) Focus(delegate func(p tview.Primitive)) {
	wm.Lock()
	lenW := len(wm.windows)
	if lenW > 0 {
		window := wm.windows[lenW-1]
		wm.Unlock()
		delegate(window)
		return
	}
	wm.Unlock()
}

func (wm *Manager) SetRect(x, y, width, height int) {
	wm.Box.SetRect(x, y, width, height)
	wm.Lock()
	defer wm.Unlock()
	for _, window := range wm.windows {
		wx, wy, ww, wh := window.GetRect()
		window.SetRect(x+wx, y+wy, ww, wh)
	}
}

// HasFocus returns whether or not this primitive has focus.
func (wm *Manager) HasFocus() bool {
	wm.Lock()
	defer wm.Unlock()

	for i := len(wm.windows) - 1; i >= 0; i-- {
		window := wm.windows[i]
		if window.HasFocus() {
			return true
		}
	}
	return false
}

// Draw draws this primitive onto the screen.
func (wm *Manager) Draw(screen tcell.Screen) {
	wm.Box.Draw(screen)

	wm.Lock()
	defer wm.Unlock()

	topWindowIndex := len(wm.windows) - 1
	for i := topWindowIndex; i >= 0; i-- {
		window := wm.windows[i]
		if window.HasFocus() && i != topWindowIndex {
			wm.setZ(window, WindowZTop)
			break
		}
	}

	for _, window := range wm.windows {
		mx, my, mw, mh := wm.GetInnerRect()
		x, y, w, h := window.GetRect()
		if x < mx {
			x = mx
		}
		if y < my {
			y = my
		}

		if w < minWindowWidth {
			w = minWindowWidth
		}
		if h < minWindowHeight {
			h = minWindowHeight
		}

		if w > mw || window.maximized {
			w = mw
			x = mx
		}
		if h > mh || window.maximized {
			h = mh
			y = my
		}

		if x+w > mx+mw {
			x = mx + mw - w
		}

		if y+h > my+mh {
			y = my + mh - h
		}

		window.SetRect(x, y, w, h)
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
				if wm.draggedEdge == WindowEdgeTop && wm.draggedWindow.Draggable {
					wm.draggedWindow.SetRect(x-wm.dragOffsetX, y-wm.dragOffsetY, ww, wh)
				} else {
					if wm.draggedWindow.Resizable {
						switch wm.draggedEdge {
						case WindowEdgeRight:
							wm.draggedWindow.SetRect(wx, wy, x-wx+1, wh)
						case WindowEdgeBottom:
							wm.draggedWindow.SetRect(wx, wy, ww, y-wy+1)
						case WindowEdgeLeft:
							wm.draggedWindow.SetRect(x, wy, ww+wx-x, wh)
						case WindowEdgeBottomRight:
							wm.draggedWindow.SetRect(wx, wy, x-wx+1, y-wy+1)
						case WindowEdgeBottomLeft:
							wm.draggedWindow.SetRect(x, wy, ww+wx-x, y-wy+1)
						}
					}
				}
				wm.Unlock()
				return true, nil
			}
		}

		var windows []*Window
		if wm.modalWindow != nil {
			windows = []*Window{wm.modalWindow}
		} else {
			windows = wm.windows
		}

		// Pass mouse events along to the first child item that takes it.
		for i := len(windows) - 1; i >= 0; i-- {
			window := windows[i]
			if !window.InRect(event.Position()) {
				continue
			}

			if action == tview.MouseLeftDown && window.border {
				if !window.HasFocus() {
					setFocus(window)
				}
				wx, wy, ww, wh := window.GetRect()
				x, y := event.Position()
				wm.draggedEdge = WindowEdgeNone
				switch {
				case y == wy+wh-1:
					switch {
					case x == wx:
						wm.draggedEdge = WindowEdgeBottomLeft
					case x == wx+ww-1:
						wm.draggedEdge = WindowEdgeBottomRight
					default:
						wm.draggedEdge = WindowEdgeBottom
					}
				case x == wx:
					wm.draggedEdge = WindowEdgeLeft
				case x == wx+ww-1:
					wm.draggedEdge = WindowEdgeRight
				case y == wy:
					wm.draggedEdge = WindowEdgeTop
				}
				if wm.draggedEdge != WindowEdgeNone {
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
