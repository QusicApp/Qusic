package main

import (
	"net/url"
	"qusic/logger"
	"qusic/preferences"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

var settingsDialog *dialog.CustomDialog

func settingsGeneralTab() fyne.CanvasObject {
	discordRPCConnectingText := widget.NewLabel("connecting...")
	discordRPCConnectingText.Hide()

	enableDiscordRPC := widget.NewCheck("Discord RPC", nil)
	enableDiscordRPC.SetChecked(preferences.Preferences.Bool("discord_rpc"))
	enableDiscordRPC.OnChanged = func(b bool) {
		if b == true && !preferences.Preferences.Bool("discord_rpc") {
			logger.Inf("Connecting to Discord RPC: ")
			discordRPCConnectingText.Show()
			logger.Println(rpc.Connect())
			discordRPCConnectingText.Hide()
		}
		preferences.Preferences.SetBool("discord_rpc", b)
	}

	hideApp := widget.NewCheck("Hide app instead of closing", func(b bool) {
		preferences.Preferences.SetBool("hide_app", b)
	})
	hideApp.SetChecked(preferences.Preferences.Bool("hide_app"))

	hardwareAcceleration := widget.NewCheck("Hardware acceleration", func(b bool) {
		preferences.Preferences.SetBool("hardware_acceleration", b)
	})
	hardwareAcceleration.SetChecked(preferences.Preferences.Bool("hardware_acceleration"))

	return container.NewVBox(
		container.NewHBox(enableDiscordRPC, discordRPCConnectingText),
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

var lsources = map[string]int{
	"lrclib":        0,
	"spotify":       1,
	"ytmusicsynced": 2,
	"ytmusicplain":  3,
	"youtubesub":    4,
	"genius":        5,
	"lyrics.ovh":    6,
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
		preferences.Preferences.SetString("source", s)
	}
	sel.SetSelectedIndex(sources[preferences.Preferences.StringWithFallback("source", "ytmusic")])

	ytmusicSV := widget.NewCheck("Show video results", func(b bool) {
		preferences.Preferences.SetBool("ytmusic.show_video_results", b)
	})
	ytmusicSV.SetChecked(preferences.Preferences.Bool("ytmusic.show_video_results"))

	lyricsSel := widget.NewSelect([]string{
		"LRCLIB (Synced)",
		"Spotify (Synced) (Requires Spotify Premium)",
		"YouTube Music (Synced)",
		"YouTube Music",
		"YouTube Subtitles (Synced)",
		"Genius",
		"Lyrics.ovh",
	}, nil)
	lyricsSel.OnChanged = func(s string) {
		i := lyricsSel.SelectedIndex()
		switch i {
		case 0:
			s = "lrclib"
		case 1:
			s = "spotify"
		case 2:
			s = "ytmusicsynced"
		case 3:
			s = "ytmusicplain"
		case 4:
			s = "youtubesub"
		case 5:
			s = "genius"
		case 6:
			s = "lyrics.ovh"
		default:
			return
		}
		preferences.Preferences.SetString("lyrics.source", s)
	}
	lyricsSel.SetSelectedIndex(lsources[preferences.Preferences.StringWithFallback("lyrics.source", "lrclib")])

	spotifysp_dcEntry := widget.NewEntry()
	spotifysp_dcEntry.SetPlaceHolder("spotify sp_dc")

	spotifysp_dcCheck := widget.NewButton("verify", nil)

	spotifysp_dcCheck.OnTapped = func() {
		preferences.Preferences.SetString("spotify.sp_dc", spotifysp_dcEntry.Text)
		spotifyClient.Cookie_sp_dc = spotifysp_dcEntry.Text
		ok := spotifyClient.Ok(true)
		if ok {
			spotifysp_dcCheck.Importance = widget.SuccessImportance
			spotifysp_dcCheck.SetText("valid")
		} else {
			spotifysp_dcCheck.Importance = widget.DangerImportance
			spotifysp_dcCheck.SetText("invalid")
			preferences.Preferences.SetString("spotify.sp_dc", "")
			spotifyClient.Cookie_sp_dc = ""
		}
	}
	spotifysp_dcEntry.OnChanged = func(s string) {
		spotifysp_dcCheck.SetText("verify")
		spotifysp_dcCheck.Enable()
	}
	spotifysp_dcEntry.SetText(preferences.Preferences.String("spotify.sp_dc"))

	spotifyDYT := widget.NewCheck("Download songs from YouTube", func(b bool) {
		preferences.Preferences.SetBool("spotify.download_yt", b)
	})
	spotifyDYT.SetChecked(preferences.Preferences.Bool("spotify.download_yt"))

	discl := widget.NewLabel("Only required for lyrics and decrypting songs. If not provided, songs will be downloaded from YouTube")
	discl.Wrapping = fyne.TextWrapBreak
	return container.NewVBox(
		widget.NewRichTextFromMarkdown("# Sources"),
		container.NewBorder(nil, nil, widget.NewLabel("Selected Music Source"), nil, container.NewGridWithColumns(3, sel)),
		container.NewBorder(nil, nil, widget.NewLabel("Selected Lyric Source"), nil, container.NewGridWithColumns(3, lyricsSel)),

		widget.NewRichTextFromMarkdown("## YouTube Music"),
		ytmusicSV,

		widget.NewRichTextFromMarkdown("## Spotify"),
		spotifyDYT,
		container.NewBorder(nil, nil, widget.NewLabel("Cookie"), nil, container.NewGridWithColumns(3, container.NewBorder(nil, nil, nil, spotifysp_dcCheck, spotifysp_dcEntry))),
		discl,
	)
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

	size := settingsDialog.MinSize()
	size.Width *= 1.5
	size.Height *= 1.2
	settingsDialog.Resize(size)

	settingsDialog.Show()
}
