package main

import (
	"github.com/epiclabs-io/winman"
	"github.com/rivo/tview"
)

func main() {

	app := tview.NewApplication()
	wm := winman.NewWindowManager()

	content := tview.NewTextView().
		SetText("Hello, world!").       // set content of the text view
		SetTextAlign(tview.AlignCenter) // align text to the center of the text view

	window := wm.NewWindow(). // create new window and add it to the window manager
					Show().                   // make window visible
					SetRoot(content).         // have the text view above be the content of the window
					SetDraggable(true).       // make window draggable around the screen
					SetResizable(true).       // make the window resizable
					SetTitle("Hi!").          // set the window title
					AddButton(&winman.Button{ // create a button with an X to close the application
			Symbol:  'X',
			OnClick: func() { app.Stop() }, // close the application
		})

	window.SetRect(5, 5, 30, 10) // place the window

	// now, execute the application:
	if err := app.SetRoot(wm, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}
