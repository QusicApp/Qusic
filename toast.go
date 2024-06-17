package main

import (
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

var toastContainer *fyne.Container

func showToast(text string) {
	if toastContainer == nil {
		return
	}
	toastContainer.Objects[1].(*widget.RichText).ParseMarkdown("## " + text)
	toastContainer.Refresh()
	toastContainer.Show()
	go func() {
		time.Sleep(2 * time.Second)
		toastContainer.Hide()
	}()
}
