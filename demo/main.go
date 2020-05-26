// Demo code for the Form primitive.
package main

import (
	"fmt"
	"strconv"

	"github.com/epiclabs-io/winman"
	"github.com/rivo/tview"
)

func calculator(wm *winman.Manager) winman.Window {

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
		Symbol:       'X',
		Alignment:    winman.ButtonLeft,
		ClickHandler: func() { wm.Hide(wnd) },
	})
	wnd.SetRect(0, 0, 30, 15)
	wnd.Draggable = true
	wnd.Resizable = true

	return wnd
}

func main() {

	app := tview.NewApplication()
	wm := winman.NewWindowManager()

	modalWindowMessage := tview.NewTextView().SetText("\nChanges have been saved").SetTextAlign(tview.AlignCenter)
	modalWindowContent := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(modalWindowMessage, 0, 1, false)
	modalWindow := winman.NewWindow().SetRoot(modalWindowContent)
	modalWindowButton := tview.NewButton("OK").SetSelectedFunc(func() { wm.Hide(modalWindow) })
	modalWindowContent.AddItem(modalWindowButton, 1, 0, true)
	modalWindow.SetTitle("Confirmation")
	modalWindow.SetRect(4, 2, 30, 6)
	modalWindow.Draggable = true

	var createForm func() winman.Window
	var counter = 0

	setFocus := func(p tview.Primitive) {
		go app.QueueUpdateDraw(func() {
			app.SetFocus(p)
		})
	}

	setZ := func(wnd *winman.WindowBase, newZ int) {
		go app.QueueUpdateDraw(func() {
			newTopWindow := wm.Window(wm.WindowCount() - 2)
			if newTopWindow != nil {
				app.SetFocus(newTopWindow)
				wm.SetZ(wnd, newZ)
			}
		})
	}

	createForm = func() winman.Window {
		counter++
		form := tview.NewForm()
		window := winman.NewWindow().SetRoot(form)
		window.Draggable = true
		window.Resizable = true

		form.AddDropDown("Title", []string{"Mr.", "Ms.", "Mrs.", "Dr.", "Prof."}, 0, nil).
			AddInputField("First name", "", 20, nil, nil).
			AddPasswordField("Password", "", 10, '*', nil).
			AddCheckbox("Draggable", window.Draggable, func(checked bool) {
				window.Draggable = checked
			}).
			AddCheckbox("Resizable", window.Draggable, func(checked bool) {
				window.Resizable = checked
			}).
			AddCheckbox("Border", window.Draggable, func(checked bool) {
				window.SetBorder(checked)
			}).
			AddInputField("Z-Index", "", 20, func(text string, char rune) bool {
				return char >= '0' && char <= '9'
			}, nil).
			AddButton("Set Z", func() {
				zIndexField := form.GetFormItemByLabel("Z-Index").(*tview.InputField)
				z, _ := strconv.Atoi(zIndexField.GetText())
				setZ(window, z)
			}).
			AddButton("New", func() {
				newWnd := createForm()
				wm.Show(newWnd)
				setFocus(newWnd)
			}).
			AddButton("Modal", func() {
				newWnd := createForm()
				wm.ShowModal(newWnd)
				setFocus(newWnd)
			}).
			AddButton("Save", func() {
				wm.ShowModal(modalWindow).Center(modalWindow)
				setFocus(modalWindow)
			}).
			AddButton("Calc", func() {
				calc := calculator(wm)
				wm.Show(calc).Center(calc)
				setFocus(calc)
			}).
			AddButton("Quit", func() {
				app.Stop()
			})

		title := fmt.Sprintf("Window%d", counter)
		window.SetBorder(true).SetTitle(title).SetTitleAlign(tview.AlignCenter)
		window.SetRect(2+counter*2, 2+counter, 50, 30)
		window.AddButton(&winman.Button{
			Symbol:    'X',
			Alignment: winman.ButtonLeft,
			ClickHandler: func() {
				wm.Hide(window)
				setFocus(wm)
			},
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
		wm.Show(window)
		return window
	}

	for i := 0; i < 10; i++ {
		createForm()
	}

	if err := app.SetRoot(wm, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}
