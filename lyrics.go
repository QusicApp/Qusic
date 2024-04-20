package main

import (
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func lyricsPage() fyne.CanvasObject {
	txt := widget.NewRichText()
	passed, _ := player.TimePosition(false)
	passedD := time.Duration(passed) * time.Second
	for _, lyric := range player.CurrentSong().SongInfo.SyncedLyrics {
		seg := &widget.TextSegment{Text: lyric.Lyric, Style: widget.RichTextStyle{SizeName: theme.SizeNameHeadingText}}
		if lyric.At <= passedD {
			seg.Style.TextStyle.Bold = true
		}
		txt.Segments = append(txt.Segments, seg)
	}
	txt.Wrapping = fyne.TextWrapBreak
	return container.NewVScroll(txt)
}
