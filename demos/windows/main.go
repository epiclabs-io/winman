// Demo code for the Form primitive.
package main

import (
	"fmt"

	"gitlab.com/tslocum/cview"
)

func main() {

	app := cview.NewApplication()
	wm := cview.NewWindowManager()

	modalWindowMessage := cview.NewTextView().SetText("\nChanges have been saved").SetTextAlign(cview.AlignCenter)
	modalWindowContent := cview.NewFlex().
		SetDirection(cview.FlexRow).
		AddItem(modalWindowMessage, 0, 1, false)
	modalWindow := wm.NewWindow(modalWindowContent, true)
	modalWindowButton := cview.NewButton("OK").SetSelectedFunc(func() { modalWindow.Hide() })
	modalWindowContent.AddItem(modalWindowButton, 1, 0, true)
	modalWindow.SetTitle("Confirmation")
	modalWindow.SetRect(4, 2, 30, 6)
	modalWindow.Draggable = true

	var createForm func(int) *cview.Window
	createForm = func(i int) *cview.Window {

		form := cview.NewForm()
		window := wm.NewWindow(form, true)
		window.Draggable = true
		window.Resizable = true

		form.AddDropDown("Title", []string{"Mr.", "Ms.", "Mrs.", "Dr.", "Prof."}, 0, nil).
			AddInputField("First name", "", 20, nil, nil).
			AddInputField("Last name", "", 20, nil, nil).
			AddPasswordField("Password", "", 10, '*', nil).
			AddCheckBox("", "Draggable", window.Draggable, func(checked bool) {
				window.Draggable = checked
			}).
			AddCheckBox("", "Resizable", window.Draggable, func(checked bool) {
				window.Resizable = checked
			}).
			AddButton("Save", func() {
				modalWindow.ShowModal()
				modalWindow.Center()
				app.SetFocus(modalWindow)
			}).
			AddButton("Quit", func() {
				app.Stop()
			}).
			AddButton("New Window", func() {
				app.SetFocus(createForm(i + 1).Show())
			}).
			AddButton("New Modal", func() {
				app.SetFocus(createForm(i + 1).ShowModal())
			})

		title := fmt.Sprintf("Window%d", i)
		window.SetBorder(true).SetTitle(title).SetTitleAlign(cview.AlignCenter)
		window.SetRect(2+i*2, 2+i, 50, 30)
		window.AddButton(&cview.WindowButton{
			Symbol:       'X',
			Alignment:    cview.AlignLeft,
			ClickHandler: func() { window.Hide() },
		})
		var maxMinButton *cview.WindowButton
		maxMinButton = &cview.WindowButton{
			Symbol:    '▴',
			Alignment: cview.AlignRight,
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
