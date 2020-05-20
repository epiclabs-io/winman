// Demo code for the Form primitive.
package main

import (
	"fmt"

	"gitlab.com/tslocum/cview"
)

func main() {
	app := cview.NewApplication()
	wm := cview.NewWindowManager()
	createForm := func(i int) {
		form := cview.NewForm().
			AddDropDown("Title", []string{"Mr.", "Ms.", "Mrs.", "Dr.", "Prof."}, 0, nil).
			AddInputField("First name", "", 20, nil, nil).
			AddInputField("Last name", "", 20, nil, nil).
			AddPasswordField("Password", "", 10, '*', nil).
			AddCheckBox("", "Age 18+", false, nil).
			AddButton("Save", nil).
			AddButton("Quit", func() {
				app.Stop()
			})

		wm.AddItem(form, true)
		window := wm.GetWindow(form)
		window.SetBorder(true).SetTitle(fmt.Sprintf("Form%d", i)).SetTitleAlign(cview.AlignCenter)
		window.SetRect(2+i*2, 2+i, 50, 30)
		//window.Center(true, true)
		/* 		window.SetMouseCapture(func(action cview.MouseAction, event *tcell.EventMouse) (cview.MouseAction, *tcell.EventMouse) {
			x, y := event.Position()
			window.SetTitle(fmt.Sprintf("Form%d - %d, %d", i, x, y))
			app.Draw()
			return action, event
		}) */
	}

	for i := 0; i < 10; i++ {
		createForm(i)
	}

	if err := app.SetRoot(wm, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}
