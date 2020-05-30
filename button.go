package winman

// ButtonSide determines the alignment of a window button
type ButtonSide int16

const (
	ButtonLeft  = iota // the button will be drawn on the left
	ButtonRight        // the button will be drawn on the right
)

type Button struct {
	Symbol    rune // icon for the button
	offsetX   int  // where the button is drawn
	offsetY   int
	Alignment ButtonSide // alignment of the button, left or right
	OnClick   func()     // callback to be invoked when the button is clicked
}
