package main

import (
	"net/url"
	"qusic/logger"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

var settingsDialog *dialog.CustomDialog

type Preferences map[string]any

func (p Preferences) StringWithFallback(key string, fallback string) string {
	v, ok := p[key]
	if !ok {
		v := fyne.CurrentApp().Preferences().StringWithFallback(key, fallback)
		p[key] = v
		return v
	}
	b, _ := v.(string)
	return b
}

func (p Preferences) SetString(key string, value string) {
	p[key] = value
	fyne.CurrentApp().Preferences().SetString(key, value)
}

func (p Preferences) String(key string) string {
	v, ok := p[key]
	if !ok {
		v := fyne.CurrentApp().Preferences().String(key)
		p[key] = v
		return v
	}
	b, _ := v.(string)
	return b
}

func (p Preferences) BoolWithFallback(key string, fallback bool) bool {
	v, ok := p[key]
	if !ok {
		v := fyne.CurrentApp().Preferences().BoolWithFallback(key, fallback)
		p[key] = v
		return v
	}
	b, _ := v.(bool)
	return b
}

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
			logger.Println(rpc.Connect())
		}
		preferences.SetBool("discord_rpc", b)
	})
	enableDiscordRPC.SetChecked(preferences.Bool("discord_rpc"))

	hideApp := widget.NewCheck("Hide app instead of closing", func(b bool) {
		preferences.SetBool("hide_app", true)
	})
	hideApp.SetChecked(preferences.Bool("hide_app"))

	hardwareAcceleration := widget.NewCheck("Hardware acceleration", func(b bool) {
		preferences.SetBool("hardware_acceleration", true)
	})
	hardwareAcceleration.SetChecked(preferences.Bool("hardware_acceleration"))

	return container.NewVBox(
		enableDiscordRPC,
		hideApp,
		hardwareAcceleration,
		widget.NewRichTextFromMarkdown("## Links"),
		container.NewHBox(
			widget.NewHyperlink("GitHub Repository", &url.URL{
				Scheme: "https",
				Host:   "github.com",
				Path:   "oq-x/qusic",
			}),
			canvas.NewCircle(theme.ForegroundColor()),
			widget.NewHyperlink("Discord Server", &url.URL{
				Scheme: "https",
				Host:   "discord.gg",
				Path:   "naVkn4NSXx",
			}),
		),
	)
}

func settingsLogTab() fyne.CanvasObject {
	log := widget.NewLabelWithData(logger.Log.Binding)
	log.TextStyle.Monospace = true
	log.Wrapping = fyne.TextWrapBreak

	errors := widget.NewLabelWithData(logger.Errors.Binding)
	errors.TextStyle.Monospace = true
	errors.Wrapping = fyne.TextWrapBreak
	tabs := container.NewAppTabs(
		container.NewTabItem("Log", log),
		container.NewTabItem("Errors", errors),
	)

	return tabs
}

var sources = map[string]int{
	"ytmusic": 0,
	"spotify": 1,
}

func settingsSourcesTab() fyne.CanvasObject {
	sel := widget.NewSelect([]string{
		"YouTube Music",
		"Spotify",
	}, nil)
	sel.OnChanged = func(s string) {
		i := sel.SelectedIndex()
		switch i {
		case 0:
			s = "ytmusic"
			player.Source = youtubeSource
		case 1:
			s = "spotify"
			player.Source = spotifySource
		default:
			return
		}
		preferences.SetString("source", s)
	}
	sel.SetSelectedIndex(sources[preferences.StringWithFallback("source", "ytmusic")])

	ytmusicSV := widget.NewCheck("Show video results", func(b bool) {
		preferences.SetBool("ytmusic.show_video_results", b)
	})
	ytmusicSV.SetChecked(preferences.BoolWithFallback("ytmusic.show_video_results", true))

	lyricsSel := widget.NewSelect([]string{
		"LRCLIB (Synced)",
		"YouTube Music",
		"Genius",
	}, nil)
	lyricsSel.OnChanged = func(s string) {
		i := lyricsSel.SelectedIndex()
		switch i {
		case 0:
			s = "lrclib"
		case 1:
			s = "ytmusic"
		case 2:
			s = "genius"
		default:
			return
		}
		preferences.SetString("lyrics.source", s)
	}
	lyricsSel.SetSelectedIndex(lsources[preferences.StringWithFallback("lyrics.source", "lrclib")])

	return container.NewVBox(
		container.NewBorder(nil, nil, widget.NewLabel("Selected Source"), nil, container.NewGridWithColumns(3, sel)),
		container.NewBorder(nil, nil, widget.NewLabel("Selected Lyric Source"), nil, container.NewGridWithColumns(3, lyricsSel)),

		widget.NewRichTextFromMarkdown("# YouTube Music"),
		ytmusicSV,
	)
}

var lsources = map[string]int{
	"lrclib":  0,
	"ytmusic": 1,
	"genius":  2,
}

func settings(w fyne.Window) {
	doneButton := widget.NewButton("Done", func() {
		settingsDialog.Hide()
	})

	tabs := container.NewAppTabs(
		container.NewTabItem("General", settingsGeneralTab()),
		container.NewTabItem("Sources", settingsSourcesTab()),
		container.NewTabItem("Log", settingsLogTab()),
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
