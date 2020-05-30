package winman

import "fmt"

type Rect struct {
	X int
	Y int
	W int
	H int
}

func NewRect(x, y, w, h int) Rect {
	return Rect{x, y, w, h}
}

func (r Rect) String() string {
	return fmt.Sprintf("{(%d, %d) %dx%d}", r.X, r.Y, r.W, r.H)
}

func (r Rect) Contains(x, y int) bool {
	return x >= r.X && x < r.X+r.W && y >= r.Y && y < r.Y+r.H

}

func (r *Rect) Rect() (int, int, int, int) {
	return r.X, r.Y, r.W, r.H
}
