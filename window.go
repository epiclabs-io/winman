package winman

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// Window interface defines primitives that can be managed by the Window Manager
type Window interface {
	tview.Primitive
	// IsModal defines this window as modal. When a window is modal, input cannot go to other windows
	IsModal() bool

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
	root        tview.Primitive // The item contained in the window
	buttons     []*Button       // window buttons on the title bar
	border      bool            // whether to render a border
	restoreRect Rect            // store previous coordinates after restoring from maximize
	maximized   bool            // whether the window is maximized to the entire window manager area
	Draggable   bool            //whether this window can be dragged around with the mouse
	Resizable   bool            // whether this window is user-resizable
	Modal       bool            // whether this window is modal
	Visible     bool            // whether this window is rendered
}

// NewWindow creates a new window
func NewWindow() *WindowBase {
	window := &WindowBase{
		Box: tview.NewBox(), // initialize underlying box
	}
	window.restoreRect = NewRect(window.GetRect())
	window.SetBorder(true)
	return window
}

// SetRoot sets the main content of the window
func (w *WindowBase) SetRoot(root tview.Primitive) *WindowBase {
	w.root = root
	return w
}

// GetRoot returns the primitive that represents the main content of the window
func (w *WindowBase) GetRoot() tview.Primitive {
	return w.root
}

// SetModal makes this window modal. A modal window captures all input
func (w *WindowBase) SetModal(modal bool) *WindowBase {
	w.Modal = modal
	return w
}

// IsModal returns true if this window is modal
func (w *WindowBase) IsModal() bool {
	return w.Modal
}

// HasBorder returns true if this window has a border
// windows without border cannot be resized or dragged by the user
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

// IsDraggable returns true if this window can be dragged by the user
func (w *WindowBase) IsDraggable() bool {
	return w.Draggable
}

// SetDraggable sets if this window can be dragged by the user
func (w *WindowBase) SetDraggable(draggable bool) *WindowBase {
	w.Draggable = draggable
	return w
}

// IsResizable returns true if the user may resize this window
func (w *WindowBase) IsResizable() bool {
	return w.Resizable
}

// SetResizable sets if this window can be resized
func (w *WindowBase) SetResizable(resizable bool) *WindowBase {
	w.Resizable = resizable
	return w
}

// SetTitle sets the window title
func (w *WindowBase) SetTitle(text string) *WindowBase {
	w.Box.SetTitle(text)
	return w
}

// IsVisible returns true if this window is rendered and may
// get focus
func (w *WindowBase) IsVisible() bool {
	return w.Visible
}

// Show makes the window visible
func (w *WindowBase) Show() *WindowBase {
	w.Visible = true
	return w
}

// Hide hides this window
func (w *WindowBase) Hide() *WindowBase {
	w.Visible = false
	return w
}

// Draw draws this primitive on to the screen
func (w *WindowBase) Draw(screen tcell.Screen) {
	if w.HasFocus() { // if the window has focus, make sure the underlying box shows a thicker border
		w.Box.Focus(nil)
	} else {
		w.Box.Blur()
	}
	w.Box.Draw(screen) // draw the window frame

	// draw the underlying root primitive within the window bounds
	if w.root != nil {
		x, y, width, height := w.GetInnerRect()
		w.root.SetRect(x, y, width, height)
		w.root.Draw(NewClipRegion(screen, x, y, width, height))
	}

	// draw the window border
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

			// render the window title buttons
			tview.Print(screen, tview.Escape(fmt.Sprintf("[%c]", button.Symbol)), buttonX-1, buttonY, 9, 0, tcell.ColorYellow)
		}
	}
}

// Maximize signals the window manager to resize this window to the maximum size available
func (w *WindowBase) Maximize() *WindowBase {
	w.restoreRect = NewRect(w.GetRect())
	w.maximized = true
	return w
}

// IsMaximized returns true if this window is maximized
func (w *WindowBase) IsMaximized() bool {
	return w.maximized
}

// Restore restores the window to the size it had before maximizing
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
		return w.root.HasFocus()
	}
	return w.Box.HasFocus()
}

// MouseHandler returns a mouse handler for this primitive
func (w *WindowBase) MouseHandler() func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
	return w.WrapMouseHandler(func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
		if action == tview.MouseLeftClick {
			x, y := event.Position()
			wx, wy, width, _ := w.GetRect()

			// check if any window button was pressed
			// if the window does not have border, it cannot receive button events
			if y == wy && w.border {
				for _, button := range w.buttons {
					if button.offsetX >= 0 && x == wx+button.offsetX || button.offsetX < 0 && x == wx+width+button.offsetX {
						if button.OnClick != nil {
							button.OnClick()
						}
						return true, nil
					}
				}
			}
		}
		// pass on clicks to the root primitive, if any
		if w.root != nil {
			return w.root.MouseHandler()(action, event, setFocus)
		}
		return w.Box.MouseHandler()(action, event, setFocus)
	})
}

// InputHandler returns a handler which receives key events when it has focus.
func (w *WindowBase) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	if w.root != nil {
		return w.root.InputHandler()
	}
	return nil
}

// AddButton adds a new window button to the title bar
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

// GetButton returns the given button
func (w *WindowBase) GetButton(i int) *Button {
	if i < 0 || i >= len(w.buttons) {
		return nil
	}
	return w.buttons[i]
}

// ButtonCount returns the number of buttons in the window title bar
func (w *WindowBase) ButtonCount() int {
	return len(w.buttons)
}
