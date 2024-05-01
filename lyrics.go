package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

var lyricsTxt *widget.RichText
var lyricsScroll *container.Scroll

func lyricsPage() fyne.CanvasObject {
	lyricsTxt = widget.NewRichText()
	lyricsTxt.Wrapping = fyne.TextWrapBreak

	lyricsScroll = container.NewVScroll(lyricsTxt)
	return lyricsScroll
}
