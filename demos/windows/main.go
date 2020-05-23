// Demo code for the Form primitive.
package main

import (
	"fmt"
	"strconv"

	"gitlab.com/tslocum/cview"
	"gitlab.com/tslocum/cview/winman"
)

func calculator() *winman.Window {

	value := []float64{0.0, 0.0}
	i := 0
	decimal := 1.0
	op := ' '
	display := cview.NewTextView().
		SetText("0.").
		SetTextAlign(cview.AlignRight)

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

	newCalcButton := func(char rune) *cview.Button {
		return cview.NewButton(string(char)).SetSelectedFunc(func() {
			keyPressed(char)
		})
	}

	grid := cview.NewGrid().
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
		Symbol:       'X',
		Alignment:    winman.ButtonLeft,
		ClickHandler: func() { wnd.Hide() },
	})
	wnd.SetRect(0, 0, 30, 15)
	wnd.Draggable = true
	wnd.Resizable = true

	return wnd
}

func main() {

	app := cview.NewApplication()
	wm := winman.NewWindowManager()

	modalWindowMessage := cview.NewTextView().SetText("\nChanges have been saved").SetTextAlign(cview.AlignCenter)
	modalWindowContent := cview.NewFlex().
		SetDirection(cview.FlexRow).
		AddItem(modalWindowMessage, 0, 1, false)
	modalWindow := winman.NewWindow().SetRoot(modalWindowContent)
	modalWindowButton := cview.NewButton("OK").SetSelectedFunc(func() { wm.Hide(modalWindow) })
	modalWindowContent.AddItem(modalWindowButton, 1, 0, true)
	modalWindow.SetTitle("Confirmation")
	modalWindow.SetRect(4, 2, 30, 6)
	modalWindow.Draggable = true

	var createForm func(int) *winman.Window
	createForm = func(i int) *winman.Window {

		form := cview.NewForm()
		window := wm.NewWindow().SetRoot(form)
		window.Draggable = true
		window.Resizable = true

		form.AddDropDown("Title", []string{"Mr.", "Ms.", "Mrs.", "Dr.", "Prof."}, 0, nil).
			AddInputField("First name", "", 20, nil, nil).
			AddPasswordField("Password", "", 10, '*', nil).
			AddCheckBox("", "Draggable", window.Draggable, func(checked bool) {
				window.Draggable = checked
			}).
			AddCheckBox("", "Resizable", window.Draggable, func(checked bool) {
				window.Resizable = checked
			}).
			AddCheckBox("", "Border", window.Draggable, func(checked bool) {
				window.SetBorder(checked)
			}).
			AddInputField("Z-Index", "", 20, func(text string, char rune) bool {
				return char >= '0' && char <= '9'
			}, nil).
			AddButton("Set Z", func() {
				zIndexField := form.GetFormItemByLabel("Z-Index").(*cview.InputField)
				z, _ := strconv.Atoi(zIndexField.GetText())
				wm.SetZ(window, z)
				app.SetFocus(wm.Window(wm.WindowCount() - 1))
			}).
			AddButton("New", func() {
				app.SetFocus(createForm(i + 1).Show())
			}).
			AddButton("Modal", func() {
				app.SetFocus(createForm(i + 1).ShowModal())
			}).
			AddButton("Save", func() {
				app.SetFocus(modalWindow.ShowModal().Center())
			}).
			AddButton("Calc", func() {
				calc := calculator()
				wm.Show(calc)
				calc.Center()
				app.SetFocus(calc)
				window.SetTitle(fmt.Sprintf("%t", calc.HasFocus()))
			}).
			AddButton("Quit", func() {
				app.Stop()
			})

		title := fmt.Sprintf("Window%d", i)
		window.SetBorder(true).SetTitle(title).SetTitleAlign(cview.AlignCenter)
		window.SetRect(2+i*2, 2+i, 50, 30)
		window.AddButton(&winman.Button{
			Symbol:       'X',
			Alignment:    winman.ButtonLeft,
			ClickHandler: func() { window.Hide() },
		})
		var maxMinButton *winman.Button
		maxMinButton = &winman.Button{
			Symbol:    '▴',
			Alignment: winman.ButtonRight,
			ClickHandler: func() {
				if window.IsMaximized() {
					window.Restore()
					maxMinButton.Symbol = '▴'
				} else {
					window.Maximize()
					maxMinButton.Symbol = '▾'
				}
			},
		}
		window.AddButton(maxMinButton)
		window.Show()
		return window
	}

	for i := 0; i < 10; i++ {
		createForm(i)
	}

	if err := app.SetRoot(wm, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}
