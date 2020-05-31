// Demo code for the Form primitive.
package main

import (
	"fmt"
	"strconv"

	"github.com/epiclabs-io/winman"
	"github.com/rivo/tview"
)

func main() {

	app := tview.NewApplication()
	wm := winman.NewWindowManager()

	quitMsgBox := MsgBox("Confirmation", "Really quit?", []string{"Yes", "No"}, func(clicked string) {
		if clicked == "Yes" {
			app.Stop()
		}
	})
	wm.AddWindow(quitMsgBox)

	calc := calculator()
	wm.AddWindow(calc)

	setFocus := func(p tview.Primitive) {
		go app.QueueUpdateDraw(func() {
			app.SetFocus(p)
		})
	}

	var createForm func(modal bool) *winman.WindowBase
	var counter = 0

	setZ := func(wnd *winman.WindowBase, newZ int) {
		go app.QueueUpdateDraw(func() {
			newTopWindow := wm.Window(wm.WindowCount() - 2)
			if newTopWindow != nil {
				app.SetFocus(newTopWindow)
				wm.SetZ(wnd, newZ)
			}
		})
	}

	createForm = func(modal bool) *winman.WindowBase {
		counter++
		form := tview.NewForm()
		window := winman.NewWindow().
			SetRoot(form).
			SetResizable(true).
			SetDraggable(true).
			SetModal(modal)

		quit := func() {
			if wm.WindowCount() == 3 {
				quitMsgBox.Show()
				wm.Center(quitMsgBox)
				setFocus(quitMsgBox)
			} else {
				wm.RemoveWindow(window)
				setFocus(wm)
			}
		}

		form.AddDropDown("Title", []string{"Mr.", "Ms.", "Mrs.", "Dr.", "Prof."}, 0, nil).
			AddInputField("First name", "", 20, nil, nil).
			AddPasswordField("Password", "", 10, '*', nil).
			AddCheckbox("Draggable", window.IsDraggable(), func(checked bool) {
				window.SetDraggable(checked)
			}).
			AddCheckbox("Resizable", window.IsResizable(), func(checked bool) {
				window.SetResizable(checked)
			}).
			AddCheckbox("Modal", window.Modal, func(checked bool) {
				window.SetModal(checked)
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
				newWnd := createForm(false).Show()
				wm.AddWindow(newWnd)
				setFocus(newWnd)
			}).
			AddButton("Modal", func() {
				newWnd := createForm(true).Show()
				newWnd.Modal = true
				wm.AddWindow(newWnd)
				setFocus(newWnd)
			}).
			AddButton("Calc", func() {
				calc.Show()
				wm.Center(calc)
				setFocus(calc)
			}).
			AddButton("Close", quit)

		title := fmt.Sprintf("Window%d", counter)
		window.SetBorder(true).SetTitle(title).SetTitleAlign(tview.AlignCenter)
		window.SetRect(2+counter*2, 2+counter, 50, 30)
		window.AddButton(&winman.Button{
			Symbol:    'X',
			Alignment: winman.ButtonLeft,
			OnClick:   quit,
		})

		var maxMinButton *winman.Button
		maxMinButton = &winman.Button{
			Symbol:    '▴',
			Alignment: winman.ButtonRight,
			OnClick: func() {
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
		wm.AddWindow(window)
		return window
	}

	for i := 0; i < 1; i++ {
		createForm(false).Show()
	}

	if err := app.SetRoot(wm, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}
