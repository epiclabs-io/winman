package winman

import (
	"fmt"

	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

// Window interface defines primitives that can be managed by the Window Manager
type Window interface {
	tview.Primitive
	// IsModal defines this window as modal. When a window is modal, input cannot go to other windows
	IsModal() bool

	// HasFocus returns true when this window has the focus
	HasFocus() bool

	// IsMaximized returns true when the window is maximized and takes all the space available to the
	// Window Manager
	IsMaximized() bool

	// IsResizable returns true when the window can be resized by the user
	IsResizable() bool

	// IsDraggable returns true when the window can be moved by the user by dragging
	// the title bar
	IsDraggable() bool

	// IsVisible returns true when the window has to be drawn and can receive focus
	IsVisible() bool

	// HasBorder returns true if the window must have a border
	HasBorder() bool
}

// WindowBase defines a basic window
type WindowBase struct {
	*tview.Box
	root        tview.Primitive // The item to be positioned. May be nil for an empty item.
	buttons     []*Button
	border      bool
	restoreRect Rect
	maximized   bool
	Draggable   bool
	Resizable   bool
	Modal       bool
	Visible     bool
}

// NewWindow creates a new window in this window manager
func NewWindow() *WindowBase {
	window := &WindowBase{
		Box: tview.NewBox(),
	}
	window.restoreRect = NewRect(window.GetRect())
	window.SetBorder(true)
	return window
}

func (w *WindowBase) SetRoot(root tview.Primitive) *WindowBase {
	w.root = root
	return w
}

func (w *WindowBase) GetRoot() tview.Primitive {
	return w.root
}

func (w *WindowBase) SetModal(modal bool) *WindowBase {
	w.Modal = modal
	return w
}

func (w *WindowBase) IsModal() bool {
	return w.Modal
}

func (w *WindowBase) HasBorder() bool {
	return w.border
}

// SetBorder sets the flag indicating whether or not the box should have a
// border.
func (w *WindowBase) SetBorder(show bool) *WindowBase {
	w.border = show
	w.Box.SetBorder(show)
	return w
}

func (w *WindowBase) IsDraggable() bool {
	return w.Draggable
}

func (w *WindowBase) SetDraggable(draggable bool) *WindowBase {
	w.Draggable = draggable
	return w
}

func (w *WindowBase) IsResizable() bool {
	return w.Resizable
}

func (w *WindowBase) SetResizable(resizable bool) *WindowBase {
	w.Resizable = resizable
	return w
}

func (w *WindowBase) IsVisible() bool {
	return w.Visible
}

func (w *WindowBase) Show() *WindowBase {
	w.Visible = true
	return w
}

func (w *WindowBase) Hide() *WindowBase {
	w.Visible = false
	return w
}

func (w *WindowBase) Draw(screen tcell.Screen) {
	if w.HasFocus() {
		w.Box.Focus(nil)
	} else {
		w.Box.Blur()
	}
	w.Box.Draw(screen)

	if w.root != nil {
		x, y, width, height := w.GetInnerRect()
		w.root.SetRect(x, y, width, height)
		w.root.Draw(NewClipRegion(screen, x, y, width, height))
	}

	if w.border {
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
			tview.Print(screen, tview.Escape(fmt.Sprintf("[%c]", button.Symbol)), buttonX-1, buttonY, 9, 0, tcell.ColorYellow)
		}
	}
}

func (w *WindowBase) Maximize() *WindowBase {
	w.restoreRect = NewRect(w.GetRect())
	w.maximized = true
	return w
}

func (w *WindowBase) IsMaximized() bool {
	return w.maximized
}

func (w *WindowBase) Restore() *WindowBase {
	w.SetRect(w.restoreRect.Rect())
	w.maximized = false
	return w
}

// Focus is called when this primitive receives focus.
func (w *WindowBase) Focus(delegate func(p tview.Primitive)) {
	if w.root != nil {
		delegate(w.root)
	} else {
		delegate(w.Box)
	}
	w.Visible = true
}

// HasFocus returns whether or not this primitive has focus.
func (w *WindowBase) HasFocus() bool {
	if !w.Visible {
		return false
	}
	if w.root != nil {
		return w.root.GetFocusable().HasFocus()
	} else {
		return w.Box.HasFocus()
	}
}

func (w *WindowBase) MouseHandler() func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
	return w.WrapMouseHandler(func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
		if action == tview.MouseLeftClick {
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
		if w.root != nil {
			return w.root.MouseHandler()(action, event, setFocus)
		}
		return w.Box.MouseHandler()(action, event, setFocus)
	})
}

func (w *WindowBase) AddButton(button *Button) *WindowBase {
	w.buttons = append(w.buttons, button)

	offsetLeft, offsetRight := 2, -3
	for _, button := range w.buttons {
		if button.Alignment == ButtonRight {
			button.offsetX = offsetRight
			offsetRight -= 3
		} else {
			button.offsetX = offsetLeft
			offsetLeft += 3
		}
	}

	return w
}

func (w *WindowBase) GetButton(i int) *Button {
	if i < 0 || i >= len(w.buttons) {
		return nil
	}
	return w.buttons[i]
}

func (w *WindowBase) ButtonCount() int {
	return len(w.buttons)
}
