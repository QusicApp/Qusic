package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
)

func settings(w fyne.Window) {
	b := container.NewBorder(nil, nil, nil, nil)
	d := dialog.NewCustomWithoutButtons("Settings", b, w)

	d.Show()
}
