package winman

import "github.com/gdamore/tcell/v2"

// ClipRegion implements tcell.Screen and only allows setting content within
// a defined region
type ClipRegion struct {
	tcell.Screen
	x      int
	y      int
	width  int
	height int
	style  tcell.Style
}

// NewClipRegion Creates a new clipped screen with the given rectangular coordinates
func NewClipRegion(screen tcell.Screen, x, y, width, height int) *ClipRegion {
	return &ClipRegion{
		Screen: screen,
		x:      x,
		y:      y,
		width:  width,
		height: height,
		style:  tcell.StyleDefault,
	}
}

// InRect returns true if the given coordinates are within this clipped region
func (cr *ClipRegion) InRect(x, y int) bool {
	return !(x < cr.x || y < cr.y || x >= cr.x+cr.width || y >= cr.y+cr.height)
}

// Fill implements tcell.Screen.Fill
func (cr *ClipRegion) Fill(ch rune, style tcell.Style) {
	for x := cr.x; x < cr.width; x++ {
		for y := cr.y; y < cr.height; y++ {
			cr.SetContent(x, y, ch, nil, style)
		}
	}
}

// SetCell is an older API, and will be removed.  Please use
// SetContent instead; SetCell is implemented in terms of SetContent.
func (cr *ClipRegion) SetCell(x int, y int, style tcell.Style, ch ...rune) {
	if len(ch) > 0 {
		cr.SetContent(x, y, ch[0], ch[1:], style)
	} else {
		cr.SetContent(x, y, ' ', nil, style)
	}
}

// SetContent sets the contents of the given cell location.  If
// the coordinates are out of range, then the operation is ignored.
//
// The first rune is the primary non-zero width rune.  The array
// that follows is a possible list of combining characters to append,
// and will usually be nil (no combining characters.)
//
// The results are not displayed until Show() or Sync() is called.
//
// Note that wide (East Asian full width) runes occupy two cells,
// and attempts to place character at next cell to the right will have
// undefined effects.  Wide runes that are printed in the
// last column will be replaced with a single width space on output.
func (cr *ClipRegion) SetContent(x int, y int, mainc rune, combc []rune, style tcell.Style) {
	if cr.InRect(x, y) {
		cr.Screen.SetContent(x, y, mainc, combc, style)
	}
}

// ShowCursor is used to display the cursor at a given location.
// If the coordinates -1, -1 are given or are otherwise outside the
// dimensions of the screen, the cursor will be hidden.
func (cr *ClipRegion) ShowCursor(x int, y int) {
	if cr.InRect(x, y) {
		cr.Screen.ShowCursor(x, y)
	}
}

// Clear clears the clipped region
func (cr *ClipRegion) Clear() {
	cr.Fill(' ', cr.style)
}
