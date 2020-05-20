package cview

import (
	"sync"

	"github.com/gdamore/tcell"
)

// flexItem holds layout options for one item.
type Window struct {
	*Box
	root    Primitive // The item to be positioned. May be nil for an empty item.
	ZIndex  int
	focus   bool // Whether or not this item attracts the layout's focus.
	manager *WindowManager
}

func (w *Window) SetRoot(root Primitive) {
	w.root = root
}

func (w *Window) Draw(screen tcell.Screen) {
	w.Box.Draw(screen)
	if w.root != nil {
		x, y, width, height := w.GetInnerRect()
		w.root.SetRect(x, y, width, height)
		w.root.Draw(screen)
	}
}

// Focus is called when this primitive receives focus.
func (w *Window) Focus(delegate func(p Primitive)) {
	delegate(w.root)
}

// HasFocus returns whether or not this primitive has focus.
func (w *Window) HasFocus() bool {
	return w.root.GetFocusable().HasFocus()
}

/* func (w *Window) SetRect(x, y, width, height int) {
	mx, my, _, _ := w.manager.GetInnerRect()
	w.Box.SetRect(mx+x, my+y, width, height)
	//w.root.SetRect(w.GetInnerRect())
} */
func (w *Window) Center(horizontal, vertical bool) *Window {
	_, _, mw, mh := w.manager.GetInnerRect()
	_, _, wh, ww := w.Box.GetRect()
	var wx, wy int
	if horizontal {
		wx = (mw - ww) / 2
	}
	if vertical {
		wy = (mh - wh) / 2
	}
	w.SetRect(wx, wy, ww, wh)
	return w
}

func (w *Window) MouseHandler() func(action MouseAction, event *tcell.EventMouse, setFocus func(p Primitive)) (consumed bool, capture Primitive) {
	return w.WrapMouseHandler(func(action MouseAction, event *tcell.EventMouse, setFocus func(p Primitive)) (consumed bool, capture Primitive) {
		return w.root.MouseHandler()(action, event, setFocus)
	})
}

type WindowManager struct {
	*Box

	// The windows to be positioned.
	windows []*Window

	// If set to true, the WindowManager will use the entire screen as its available space
	// instead its box dimensions.
	fullScreen             bool
	mouseWindow            *Window
	lastMouseX, lastMouseY int
	draggedWindow          *Window

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

// AddItem adds a new item to the container. The "fixedSize" argument is a width
// or height that may not be changed by the layout algorithm. A value of 0 means
// that its size is flexible and may be changed. The "proportion" argument
// defines the relative size of the item compared to other flexible-size items.
// For example, items with a proportion of 2 will be twice as large as items
// with a proportion of 1. The proportion must be at least 1 if fixedSize == 0
// (ignored otherwise).
//
// If "focus" is set to true, the item will receive focus when the Flex
// primitive receives focus. If multiple items have the "focus" flag set to
// true, the first one will receive focus.
//
// You can provide a nil value for the primitive. This will still consume screen
// space but nothing will be drawn.
func (wm *WindowManager) AddItem(item Primitive, focus bool) *WindowManager {
	wm.Lock()
	defer wm.Unlock()
	window := &Window{root: item, focus: focus, manager: wm, Box: NewBox().SetBackgroundColor(tcell.ColorDefault)}
	window.SetBorder(true)
	wm.windows = append(wm.windows, window)
	return wm
}

// RemoveItem removes all items for the given primitive from the container,
// keeping the order of the remaining items intact.
func (wm *WindowManager) RemoveItem(p Primitive) *WindowManager {
	wm.Lock()
	defer wm.Unlock()

	for index := len(wm.windows) - 1; index >= 0; index-- {
		if wm.windows[index].root == p {
			wm.windows = append(wm.windows[:index], wm.windows[index+1:]...)
		}
	}
	return wm
}

func (wm *WindowManager) GetWindow(p Primitive) *Window {
	for _, window := range wm.windows {
		if window.root == p {
			return window
		}
	}
	return nil
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

		if wm.draggedWindow != nil {
			switch action {
			case MouseLeftUp:
				wm.draggedWindow.SetTitle("Stop")
				wm.draggedWindow = nil
			case MouseMove:
				wm.draggedWindow.SetTitle("Dragging")
				x, y := event.Position()
				_, _, ww, wh := wm.draggedWindow.GetRect()
				wm.draggedWindow.Box.SetRect(x-wm.lastMouseX, y-wm.lastMouseY, ww, wh)
				return true, nil
			}
		}

		// Pass mouse events along to the first child item that takes it.
		for i := len(wm.windows) - 1; i >= 0; i-- {
			window := wm.windows[i]
			if !window.InRect(event.Position()) {
				continue
			}

			if action == MouseLeftDown {
				if !window.HasFocus() {
					setFocus(window)
				}
				wx, wy, ww, wh := window.GetRect()
				x, y := event.Position()
				if x == wx || x == wx+ww-1 || y == wy || y == wy+wh-1 {
					wm.draggedWindow = window
					wm.lastMouseX = x - wx
					wm.lastMouseY = y - wy
					window.SetTitle("Drag")
					return true, nil
				}
			}

			return window.MouseHandler()(action, event, setFocus)
		}

		return
	})
}
