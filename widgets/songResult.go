package widgets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type SongResult struct {
	widget.BaseWidget
	Name, Artist, DurationString string
	Image                        *canvas.Image
	OnTapped                     func()
}

func (card *SongResult) CreateRenderer() fyne.WidgetRenderer {
	card.ExtendBaseWidget(card)

	card.Image.FillMode = canvas.ImageFillContain

	c := container.NewHBox(
		&ImageButton{Image: card.Image, OnTapped: card.OnTapped},
		widget.NewRichText(
			&widget.TextSegment{
				Text:  card.Name,
				Style: widget.RichTextStyle{TextStyle: fyne.TextStyle{Bold: true}},
			},
			&widget.TextSegment{
				Text:  card.Artist,
				Style: widget.RichTextStyle{ColorName: theme.ColorNamePlaceHolder},
			},
		),
		layout.NewSpacer(),
		widget.NewLabelWithStyle(card.DurationString, fyne.TextAlignTrailing, fyne.TextStyle{
			Monospace: true,
		}),
	)

	return widget.NewSimpleRenderer(c)
}
