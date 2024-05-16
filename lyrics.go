package main

import (
	"fmt"
	"qusic/logger"
	"qusic/lyrics"
	"qusic/widgets"
	"slices"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	fynesyncedlyrics "github.com/dweymouth/fyne-lyrics"
)

var lyricsTxt *fynesyncedlyrics.LyricsViewer
var lyricsRect *canvas.Rectangle

const lyricsTxtDefault = "# There are no lyrics available for silence, there are simply no words"

func lyricsPage(w fyne.Window) fyne.CanvasObject {
	lyricsTxt = fynesyncedlyrics.NewLyricsViewer()
	lyricsRect = canvas.NewRectangle(theme.BackgroundColor())

	page := container.NewStack(lyricsRect, lyricsTxt)

	editor := lyricsEditorPage(w, page)
	editor.Hide()

	return container.NewBorder(container.NewHBox(
		layout.NewSpacer(),

		&widget.Button{
			Importance: widget.LowImportance,
			Icon:       theme.DocumentCreateIcon(),
			OnTapped: func() {
				cs := player.CurrentSong()
				if cs == nil {
					return
				}
				if !editor.Visible() {
					if cs.Lyrics.LyricSource != "LRCLIB" {
						dialog.NewError(fmt.Errorf("This feature is only available for the LRCLIB lyrics provider"), w).Show()
						return
					}
					editor.Show()
					page.Hide()
				} else {
					editor.Hide()
					page.Show()
				}
			},
		},
	), nil, nil, nil, page, editor)
}

func syncedLyricsEditorPage() fyne.CanvasObject {
	lines := container.NewVBox()
	lyricsAddButton := widget.NewButtonWithIcon("", theme.ContentAddIcon(), func() {
		i := len(lines.Objects)

		dur := player.TimePosition()

		lyricsEditorSyncedData = append(lyricsEditorSyncedData, lyrics.SyncedLyric{
			At: dur,
		})
		removeButton := widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {
			lyricsEditorSyncedData = slices.Delete(lyricsEditorSyncedData, i, i+1)
		})
		entry := widget.NewEntry()
		entry.OnChanged = func(s string) {
			lyricsEditorSyncedData[i].Lyric = s
		}
		time := widget.NewLabel(durString(dur))

		border := container.NewBorder(nil, nil, container.NewHBox(removeButton, time), nil, entry)
		lines.Add(border)
	})
	return container.NewVBox(
		lines, container.NewHBox(lyricsAddButton),
	)
}

func plainLyricsEditorPage() fyne.CanvasObject {
	entry := widget.NewMultiLineEntry()
	entry.SetText("Input the plain lyrics for this song")
	entry.TextStyle.Monospace = true
	entry.OnChanged = func(s string) {
		lyricsEditorPlainData = s
	}
	return entry
}

var (
	lyricsEditorPlainData  string
	lyricsEditorSyncedData []lyrics.SyncedLyric
)

func lyricsEditorPage(w fyne.Window, page fyne.CanvasObject) fyne.CanvasObject {
	tabs := container.NewAppTabs(
		container.NewTabItem("Synced", syncedLyricsEditorPage()),
		container.NewTabItem("Plain", plainLyricsEditorPage()),
	)

	button := &widget.Button{
		Importance: widget.LowImportance,
		Icon:       theme.NavigateBackIcon(),
	}
	cs := *player.CurrentSong()
	top := container.NewPadded(container.NewHBox(
		button,
		layout.NewSpacer(),
		&widgets.Button{
			Button: widget.Button{
				Importance: widget.HighImportance,
				Text:       "Publish",
				OnTapped: func() {
					d := dialog.NewConfirm("Publishing lyrics", "These lyrics are uploaded to https://lrclib.net, not saved locally, are you sure you want to continue?",
						func(b bool) {
							if b {
								status := widget.NewRichTextFromMarkdown("## status: verifying")
								d := dialog.NewCustomWithoutButtons("Publishing lyrics", container.NewCenter(
									status,
								), w)
								d.Show()

								token, err := lyrics.NewLRCLIBPublishToken()
								if err != nil {
									status.ParseMarkdown("## error verifying, check logs")
									d.SetButtons([]fyne.CanvasObject{
										widgets.NewButton("OK", func() {
											d.Hide()
										}),
									})
									logger.Errorf("error getting lrclib token: %v", err)
									return
								}

								status.ParseMarkdown("## status: publishing")
								err = lyrics.PublishSongLRCLIB(
									cs.Name,
									cs.Artists[0].Name,
									cs.Album.Name,
									cs.Duration,
									lyricsEditorPlainData,
									lyricsEditorSyncedData,
									token,
								)
								if err != nil {
									status.ParseMarkdown("## error publishing, check logs")
									d.SetButtons([]fyne.CanvasObject{
										widgets.NewButton("OK", func() {
											d.Hide()
										}),
									})
									logger.Errorf("error publishing lrclib song: %v", err)
									return
								}

								status.ParseMarkdown("## Published song successfully!")
								d.SetButtons([]fyne.CanvasObject{
									widgets.NewButton("OK", func() {
										d.Hide()
										button.OnTapped()
									}),
								})
							}
						}, w)

					d.Show()
					d.SetConfirmText("Publish")
					d.SetDismissText("Cancel")
				},
			},
		},
	))
	border := container.NewBorder(container.NewStack(top, container.NewCenter(widget.NewRichTextFromMarkdown("# Lyrics Editor"))), nil, nil, nil, tabs)

	button.OnTapped = func() {
		border.Hide()
		page.Show()
	}

	return border
}
