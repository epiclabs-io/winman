package winman

type ButtonSide int16

const (
	ButtonLeft = iota
	ButtonRight
)

type Button struct {
	Symbol       rune
	offsetX      int
	offsetY      int
	Alignment    ButtonSide
	ClickHandler func()
}
