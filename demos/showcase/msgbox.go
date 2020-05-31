package main

import (
	"github.com/epiclabs-io/winman"
	"github.com/rivo/tview"
)

// MsgBox creates a new modal message box
func MsgBox(title, text string, buttons []string, callback func(clicked string)) *winman.WindowBase {

	msgBox := winman.NewWindow()
	message := tview.NewTextView().SetText(text).SetTextAlign(tview.AlignCenter)
	buttonBar := tview.NewFlex().
		SetDirection(tview.FlexColumn)

	content := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(nil, 0, 1, false).
		AddItem(message, 0, 1, false).
		AddItem(buttonBar, 1, 0, true)

	msgBox.SetRoot(content)
	msgBox.SetTitle(title).
		SetRect(4, 2, 30, 6)
	msgBox.Draggable = true
	msgBox.Modal = true

	for _, buttonText := range buttons {
		button := func(buttonText string) *tview.Button {
			return tview.NewButton(buttonText).SetSelectedFunc(func() {
				msgBox.Hide()
				callback(buttonText)
			})
		}(buttonText)
		buttonBar.AddItem(button, 0, 1, true)
	}

	return msgBox
}
