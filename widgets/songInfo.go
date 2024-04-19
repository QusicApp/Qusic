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

	info := container.NewHBox(songInfo.Image, widget.NewRichText(
		&widget.TextSegment{
			Text:  songInfo.Name,
			Style: widget.RichTextStyle{TextStyle: fyne.TextStyle{Bold: true}},
		},
		&widget.TextSegment{
			Text:  songInfo.Artist,
			Style: widget.RichTextStyle{ColorName: theme.ColorNamePlaceHolder},
		},
	))

	rect := canvas.NewRectangle(theme.DisabledButtonColor())
	rect.CornerRadius = 10

	return widget.NewSimpleRenderer(container.NewStack(rect, info))
}
