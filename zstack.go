package cview

import (
	"sync"

	"github.com/gdamore/tcell"
)

// flexItem holds layout options for one item.
type zStackItem struct {
	Item   Primitive // The item to be positioned. May be nil for an empty item.
	ZIndex int
	Focus  bool // Whether or not this item attracts the layout's focus.
}

type ZStack struct {
	*Box

	// The items to be positioned.
	items []*zStackItem

	// If set to true, ZStack will use the entire screen as its available space
	// instead its box dimensions.
	fullScreen bool

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
func NewZStack() *ZStack {
	z := &ZStack{
		Box: NewBox().SetBackgroundColor(tcell.ColorDefault),
	}
	z.focus = z
	return z
}

// SetFullScreen sets the flag which, when true, causes the flex layout to use
// the entire screen space instead of whatever size it is currently assigned to.
func (z *ZStack) SetFullScreen(fullScreen bool) *ZStack {
	z.Lock()
	defer z.Unlock()

	z.fullScreen = fullScreen
	return z
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
func (z *ZStack) AddItem(item Primitive, focus bool) *ZStack {
	z.Lock()
	defer z.Unlock()

	z.items = append(z.items, &zStackItem{Item: item, Focus: focus})
	return z
}

// RemoveItem removes all items for the given primitive from the container,
// keeping the order of the remaining items intact.
func (z *ZStack) RemoveItem(p Primitive) *ZStack {
	z.Lock()
	defer z.Unlock()

	for index := len(z.items) - 1; index >= 0; index-- {
		if z.items[index].Item == p {
			z.items = append(z.items[:index], z.items[index+1:]...)
		}
	}
	return z
}

// Draw draws this primitive onto the screen.
func (z *ZStack) Draw(screen tcell.Screen) {
	z.Box.Draw(screen)

	z.Lock()
	defer z.Unlock()

	// Calculate size and position of the items.

	// Do we use the entire screen?
	if z.fullScreen {
		width, height := screen.Size()
		z.SetRect(0, 0, width, height)
	}

	// How much space can we distribute?
	x, y, width, height := z.GetInnerRect()

	for _, item := range z.items {
		item.Item.SetRect(x, y, width, height)

		//		if item.Item.GetFocusable().HasFocus() {
		//			defer item.Item.Draw(screen)
		//		} else {
		item.Item.Draw(screen)
		//		}
	}
}

// Focus is called when this primitive receives focus.
func (z *ZStack) Focus(delegate func(p Primitive)) {
	z.Lock()

	for i := len(z.items) - 1; i >= 0; i-- {
		item := z.items[i]
		if item.Focus {
			z.Unlock()
			delegate(item.Item)
			return
		}
	}

	z.Unlock()
}

// HasFocus returns whether or not this primitive has focus.
func (z *ZStack) HasFocus() bool {
	z.Lock()
	defer z.Unlock()

	for i := len(z.items) - 1; i >= 0; i-- {
		item := z.items[i]
		if item.Item.GetFocusable().HasFocus() {
			return true
		}
	}
	return false
}

// MouseHandler returns the mouse handler for this primitive.
func (z *ZStack) MouseHandler() func(action MouseAction, event *tcell.EventMouse, setFocus func(p Primitive)) (consumed bool, capture Primitive) {
	return z.WrapMouseHandler(func(action MouseAction, event *tcell.EventMouse, setFocus func(p Primitive)) (consumed bool, capture Primitive) {
		if !z.InRect(event.Position()) {
			return false, nil
		}

		// Pass mouse events along to the first child item that takes it.
		for i := len(z.items) - 1; i >= 0; i-- {
			item := z.items[i]
			consumed, capture = item.Item.MouseHandler()(action, event, setFocus)
			if consumed {
				return
			}
		}

		return
	})
}
