package main

import (
	"fmt"

	"github.com/epiclabs-io/winman"
	"github.com/rivo/tview"
)

func calculator() *winman.WindowBase {

	value := []float64{0.0, 0.0}
	i := 0
	decimal := 1.0
	op := ' '
	display := tview.NewTextView().
		SetText("0.").
		SetTextAlign(tview.AlignRight)

	keyPressed := func(char rune) {
		if char >= '0' && char <= '9' {
			digit := (float64)(char - '0')
			if decimal == 1.0 {
				value[i] = value[i]*10 + digit
			} else {
				value[i] = value[i] + digit*decimal
				decimal /= 10
			}
		} else {
			switch char {
			case '.':
				if decimal == 1.0 {
					decimal = decimal / 10
				}
			case '=':
				if i == 1 {
					switch op {
					case '+':
						value[0] = value[0] + value[1]
					case '-':
						value[0] = value[0] - value[1]
					case 'x':
						value[0] = value[0] * value[1]
					case '/':
						if value[1] == 0.0 {
							display.SetText("Err")
							value[0] = 0.0
						} else {
							value[0] = value[0] / value[1]
						}
					}
					i = 0
					decimal = 1.0
				} else {
					value[0] = 0.0
				}
				op = ' '
			default:
				op = char
				i = 1
				decimal = 1.0
				value[1] = 0
			}
		}
		display.SetText(fmt.Sprintf("%g", value[i]))
	}

	newCalcButton := func(char rune) *tview.Button {
		return tview.NewButton(string(char)).SetSelectedFunc(func() {
			keyPressed(char)
		})
	}

	grid := tview.NewGrid().
		SetRows(2, 0, 0, 0, 0).
		SetColumns(0, 0, 0, 0).
		SetBorders(true).
		AddItem(display, 0, 0, 1, 4, 2, 0, false)

	buttons := []rune{'7', '8', '9', '/', '4', '5', '6', 'x', '1', '2', '3', '-', '0', '.', '=', '+'}

	for i, b := range buttons {
		row := 1 + i/4
		col := i % 4
		grid.AddItem(newCalcButton(b), row, col, 1, 1, 1, 1, true)
	}

	wnd := winman.NewWindow().SetRoot(grid)
	wnd.AddButton(&winman.Button{
		Symbol:    'X',
		Alignment: winman.ButtonLeft,
		OnClick:   func() { wnd.Hide() },
	})
	wnd.SetRect(0, 0, 30, 15)
	wnd.Draggable = true
	wnd.Resizable = true

	return wnd
}
