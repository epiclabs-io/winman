package winman

import "fmt"

// Rect represents rectangular coordinates
type Rect struct {
	X int // x coordinate
	Y int // y coordinate
	W int // width
	H int // height
}

// NewRect instantiates a new Rect with the given coordinates
func NewRect(x, y, w, h int) Rect {
	return Rect{x, y, w, h}
}

// String implements Stringer
func (r Rect) String() string {
	return fmt.Sprintf("{(%d, %d) %dx%d}", r.X, r.Y, r.W, r.H)
}

// Contains returns true if the given coordinates are within this rectangle
func (r Rect) Contains(x, y int) bool {
	return x >= r.X && x < r.X+r.W && y >= r.Y && y < r.Y+r.H

}

// Rect returns the four coordinates of the rectangle: x, y, width and height
func (r *Rect) Rect() (int, int, int, int) {
	return r.X, r.Y, r.W, r.H
}
