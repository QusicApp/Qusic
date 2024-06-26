package widgets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type SongInfo struct {
	widget.BaseWidget
	Name, Artist string
	Image        *canvas.Image
}

func (songInfo *SongInfo) CreateRenderer() fyne.WidgetRenderer {
	songInfo.ExtendBaseWidget(songInfo)

	songInfo.Image.FillMode = canvas.ImageFillContain
	songInfo.Image.SetMinSize(fyne.NewSize(64, 64))

	return widget.NewSimpleRenderer(container.NewBorder(nil, nil, songInfo.Image, nil, &widget.RichText{
		Segments: []widget.RichTextSegment{&widget.TextSegment{
			Text:  songInfo.Name,
			Style: widget.RichTextStyle{TextStyle: fyne.TextStyle{Bold: true}},
		},
			&widget.TextSegment{
				Text:  songInfo.Artist,
				Style: widget.RichTextStyle{ColorName: theme.ColorNamePlaceHolder},
			}},
		Wrapping: fyne.TextWrapBreak,
	}))
}
