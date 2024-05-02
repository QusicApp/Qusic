package main

import (
	"qusic/widgets"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
)

var lyricsTxt *widgets.TappableRichText
var lyricsScroll *container.Scroll

func lyricsPage() fyne.CanvasObject {
	lyricsTxt = widgets.NewRichText()
	lyricsTxt.Wrapping = fyne.TextWrapBreak

	lyricsScroll = container.NewVScroll(lyricsTxt)
	return lyricsScroll
}
