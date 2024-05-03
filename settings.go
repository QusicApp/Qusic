package main

import (
	"fmt"
	"qusic/logger"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

var settingsDialog *dialog.CustomDialog

type Preferences map[string]any

func (p Preferences) Bool(key string) bool {
	v, ok := p[key]
	if !ok {
		v := fyne.CurrentApp().Preferences().Bool(key)
		p[key] = v
		return v
	}
	b, _ := v.(bool)
	return b
}

func (p Preferences) SetBool(key string, value bool) {
	p[key] = value
	fyne.CurrentApp().Preferences().SetBool(key, value)
}

var preferences = make(Preferences)

func settingsGeneralTab() fyne.CanvasObject {
	enableDiscordRPC := widget.NewCheck("Discord RPC", func(b bool) {
		if b == true && !preferences.Bool("discord_rpc") {
			logger.Inf("Connecting to Discord RPC: ")
			fmt.Println(rpc.Connect())
		}
		preferences.SetBool("discord_rpc", b)
	})
	enableDiscordRPC.SetChecked(preferences.Bool("discord_rpc"))
	hideApp := widget.NewCheck("Hide app instead of closing", func(b bool) {
		preferences.SetBool("hide_app", true)
	})
	hideApp.SetChecked(preferences.Bool("hide_app"))

	return container.NewVBox(enableDiscordRPC, hideApp)
}

func settings(w fyne.Window) {
	doneButton := widget.NewButton("Done", func() {
		settingsDialog.Hide()
	})

	tabs := container.NewAppTabs(
		container.NewTabItem("General", settingsGeneralTab()),
	)

	b := container.NewBorder(nil, container.NewGridWithColumns(5,
		layout.NewSpacer(),
		layout.NewSpacer(),
		doneButton,
	), nil, nil, tabs)

	settingsDialog = dialog.NewCustomWithoutButtons("Settings", b, w)
	settingsDialog.Resize(fyne.NewSize(float32(resolution.Width)/2, float32(resolution.Height)/2))

	settingsDialog.Show()
}
