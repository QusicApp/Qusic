package main

import (
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
)

func settingsDialog() {
	dialog.NewCustomWithoutButtons("Settings", container.NewWithoutLayout(), nil)
}

func settingsShow() {

}
