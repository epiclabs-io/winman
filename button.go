package winman

// ButtonSide determines the alignment of a window button
type ButtonSide int16

const (
	// ButtonLeft will set the button to be drawn on the left
	ButtonLeft = iota
	// ButtonRight will set the button to be drawn on the right
	ButtonRight
)

// Button represents a button on the window title bar
type Button struct {
	Symbol    rune // icon for the button
	offsetX   int  // where the button is drawn
	offsetY   int
	Alignment ButtonSide // alignment of the button, left or right
	OnClick   func()     // callback to be invoked when the button is clicked
}
