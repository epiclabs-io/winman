package cview

import (
	"fmt"
	"sync"

	"github.com/gdamore/tcell"
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

type WindowButtonSide int16

const (
	WindowButtonLeft = iota
	WindowButtonRight
)

const minWindowWidth = 3
const minWindowHeight = 3

type WindowButton struct {
	Symbol       rune
	offsetX      int
	offsetY      int
	Alignment    int
	ClickHandler func()
}

// flexItem holds layout options for one item.
type Window struct {
	*Box
	root          Primitive // The item to be positioned. May be nil for an empty item.
	manager       *WindowManager
	buttons       []*WindowButton
	restoreX      int
	restoreY      int
	restoreWidth  int
	restoreHeight int
	maximized     bool
	Draggable     bool
	Resizable     bool
}

func (w *Window) SetRoot(root Primitive) {
	w.root = root
}

func (w *Window) Draw(screen tcell.Screen) {
	if w.Box.HasFocus() && !w.HasFocus() {
		w.Box.Blur()
	}
	w.Box.Draw(screen)

	if w.root != nil {
		x, y, width, height := w.GetInnerRect()
		w.root.SetRect(x, y, width, height)
		w.root.Draw(NewClipRegion(screen, x, y, width, height))
	}

	x, y, width, height := w.GetRect()
	screen = NewClipRegion(screen, x, y, width, height)
	for _, button := range w.buttons {
		buttonX, buttonY := button.offsetX+x, button.offsetY+y
		if button.offsetX < 0 {
			buttonX += width
		}
		if button.offsetY < 0 {
			buttonY += height
		}

		//screen.SetContent(buttonX, buttonY, button.Symbol, nil, tcell.StyleDefault.Foreground(tcell.ColorYellow))
		Print(screen, Escape(fmt.Sprintf("[%c]", button.Symbol)), buttonX-1, buttonY, 9, 0, tcell.ColorYellow)
	}
}

func (w *Window) Show() {
	w.manager.Show(w)
}

func (w *Window) Hide() {
	w.manager.Hide(w)
}

func (w *Window) Maximize() {
	w.restoreX, w.restoreY, w.restoreHeight, w.restoreWidth = w.GetRect()
	w.SetRect(w.manager.GetInnerRect())
	w.maximized = true
}

func (w *Window) Restore() {
	w.SetRect(w.restoreX, w.restoreY, w.restoreHeight, w.restoreWidth)
	w.maximized = false
}

// Focus is called when this primitive receives focus.
func (w *Window) Focus(delegate func(p Primitive)) {
	delegate(w.root)
	w.Box.Focus(nil)
}

func (w *Window) Blur() {
	w.root.Blur()
	w.Box.Blur()
}

func (w *Window) IsMaximized() bool {
	return w.maximized
}

// HasFocus returns whether or not this primitive has focus.
func (w *Window) HasFocus() bool {
	return w.root.GetFocusable().HasFocus()
}

func (w *Window) MouseHandler() func(action MouseAction, event *tcell.EventMouse, setFocus func(p Primitive)) (consumed bool, capture Primitive) {
	return w.WrapMouseHandler(func(action MouseAction, event *tcell.EventMouse, setFocus func(p Primitive)) (consumed bool, capture Primitive) {
		if action == MouseLeftClick {
			x, y := event.Position()
			wx, wy, width, _ := w.GetRect()
			if y == wy {
				for _, button := range w.buttons {
					if button.offsetX >= 0 && x == wx+button.offsetX || button.offsetX < 0 && x == wx+width+button.offsetX {
						if button.ClickHandler != nil {
							button.ClickHandler()
						}
						return true, nil
					}
				}
			}
		}
		return w.root.MouseHandler()(action, event, setFocus)
	})
}

func (w *Window) AddButton(button *WindowButton) *Window {
	w.buttons = append(w.buttons, button)

	offsetLeft, offsetRight := 2, -3
	for _, button := range w.buttons {
		if button.Alignment == AlignRight {
			button.offsetX = offsetRight
			offsetRight -= 3
		} else {
			button.offsetX = offsetLeft
			offsetLeft += 3
		}
	}

	return w
}

type WindowManager struct {
	*Box

	// The windows to be positioned.
	windows []*Window

	// If set to true, the WindowManager will use the entire screen as its available space
	// instead its box dimensions.
	fullScreen               bool
	mouseWindow              *Window
	dragOffsetX, dragOffsetY int
	draggedWindow            *Window
	draggedEdge              WindowEdge
	modalWindow              *Window
	sync.Mutex
}

// NewFlex returns a new flexbox layout container with no primitives and its
// direction set to FlexColumn. To add primitives to this layout, see AddItem().
// To change the direction, see SetDirection().
//
// Note that Box, the superclass of Flex, will have its background color set to
// transparent so that any nil flex items will leave their background unchanged.
// To clear a Flex's background before any items are drawn, set it to the
// desired color:
//
//   flex.SetBackgroundColor(cview.Styles.PrimitiveBackgroundColor)
func NewWindowManager() *WindowManager {
	wm := &WindowManager{
		Box: NewBox().SetBackgroundColor(tcell.ColorDefault),
	}
	wm.focus = wm
	return wm
}

// SetFullScreen sets the flag which, when true, causes the flex layout to use
// the entire screen space instead of whatever size it is currently assigned to.
func (wm *WindowManager) SetFullScreen(fullScreen bool) *WindowManager {
	wm.Lock()
	defer wm.Unlock()

	wm.fullScreen = fullScreen
	return wm
}

// NewWindow creates a new window in this window manager
func (wm *WindowManager) NewWindow(root Primitive, focus bool) *Window {
	window := &Window{
		root:    root,
		manager: wm,
		Box:     NewBox().SetBackgroundColor(tcell.ColorDefault),
	}
	window.restoreX, window.restoreY, window.restoreHeight, window.restoreWidth = window.GetRect()
	window.SetBorder(true)
	return window
}

func (wm *WindowManager) Show(window *Window) *WindowManager {
	wm.Lock()
	defer wm.Unlock()
	for _, wnd := range wm.windows {
		if wnd == window {
			return wm
		}
	}
	wm.windows = append(wm.windows, window)
	return wm
}

func (wm *WindowManager) ShowModal(window *Window) *WindowManager {
	wm.Show(window)
	wm.Lock()
	defer wm.Unlock()
	wm.modalWindow = window
	return wm
}

func (wm *WindowManager) Hide(window *Window) *WindowManager {
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

func (wm *WindowManager) FindPrimitive(p Primitive) *Window {
	wm.Lock()
	defer wm.Unlock()
	for _, window := range wm.windows {
		if window.root == p {
			return window
		}
	}
	return nil
}

func (wm *WindowManager) BringToFront(window *Window) *WindowManager {
	wm.Lock()
	defer wm.Unlock()
	for i, wnd := range wm.windows {
		if wnd == window {
			wm.windows = append(append(wm.windows[:i], wm.windows[i+1:]...), window)
			break
		}
	}
	return wm
}

func (wm *WindowManager) SendToBack(window *Window) *WindowManager {
	wm.Lock()
	defer wm.Unlock()
	for i, wnd := range wm.windows {
		if wnd == window {
			wm.windows = append([]*Window{window}, append(wm.windows[:i], wm.windows[i+1:]...)...)
			break
		}
	}
	return wm
}

// Draw draws this primitive onto the screen.
func (wm *WindowManager) Draw(screen tcell.Screen) {
	wm.Box.Draw(screen)

	wm.Lock()
	defer wm.Unlock()

	// Calculate size and position of the items.

	// Do we use the entire screen?
	if wm.fullScreen {
		width, height := screen.Size()
		wm.SetRect(0, 0, width, height)
	}

	lenW := len(wm.windows)
	if lenW > 1 {
		for i, window := range wm.windows {
			if window.HasFocus() && i != lenW-1 {
				wm.windows = append(append(wm.windows[:i], wm.windows[i+1:]...), window)
				break
			}
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

// Focus is called when this primitive receives focus.
func (wm *WindowManager) Focus(delegate func(p Primitive)) {
	wm.Lock()
	if len(wm.windows) > 0 {
		window := wm.windows[0]
		wm.Unlock()
		delegate(window)
		return
	}
	wm.Unlock()
}

func (wm *WindowManager) SetRect(x, y, width, height int) {
	wm.Box.SetRect(x, y, width, height)
	wm.Lock()
	defer wm.Unlock()
	for _, window := range wm.windows {
		_, _, windowWidth, windowHeight := window.GetRect()
		window.SetRect(x+window.x, y+window.y, windowWidth, windowHeight)
	}
}

// HasFocus returns whether or not this primitive has focus.
func (wm *WindowManager) HasFocus() bool {
	wm.Lock()
	defer wm.Unlock()

	for i := len(wm.windows) - 1; i >= 0; i-- {
		window := wm.windows[i]
		if window.GetFocusable().HasFocus() {
			return true
		}
	}
	return false
}

// MouseHandler returns the mouse handler for this primitive.
func (wm *WindowManager) MouseHandler() func(action MouseAction, event *tcell.EventMouse, setFocus func(p Primitive)) (consumed bool, capture Primitive) {
	return wm.WrapMouseHandler(func(action MouseAction, event *tcell.EventMouse, setFocus func(p Primitive)) (consumed bool, capture Primitive) {
		if !wm.InRect(event.Position()) {
			return false, nil
		}
		wm.Lock()

		if wm.draggedWindow != nil {
			switch action {
			case MouseLeftUp:
				wm.draggedWindow = nil
			case MouseMove:
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

			if action == MouseLeftDown {
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
