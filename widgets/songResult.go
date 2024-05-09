package widgets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type SongResult struct {
	widget.BaseWidget
	Name, Artist, DurationString string
	Image                        *canvas.Image
	OnTapped                     func()
	OptionsOnTapped              func()

	Options *ThreeDotOptions
}

func (card *SongResult) MouseIn(*desktop.MouseEvent) {
	card.Options.Show()
}

func (card *SongResult) MouseMoved(*desktop.MouseEvent) {
}

func (card *SongResult) MouseOut() {
	card.Options.Hide()
}

func (card *SongResult) CreateRenderer() fyne.WidgetRenderer {
	if card.Options == nil {
		card.Options = NewThreeDotOptions(card.OptionsOnTapped)
		card.Options.Hide()
	}
	card.ExtendBaseWidget(card)

	card.Image.FillMode = canvas.ImageFillContain

	text := widget.NewRichText(
		&widget.TextSegment{
			Text:  card.Name,
			Style: widget.RichTextStyle{TextStyle: fyne.TextStyle{Bold: true}},
		},
		&widget.TextSegment{
			Text:  card.Artist,
			Style: widget.RichTextStyle{ColorName: theme.ColorNamePlaceHolder},
		},
	)
	text.Truncation = fyne.TextTruncateEllipsis

	c := container.NewBorder(
		nil,
		nil,
		&ImageButton{Image: card.Image, OnTapped: card.OnTapped},
		container.NewVBox(layout.NewSpacer(),
			container.NewBorder(nil, nil,
				widget.NewLabelWithStyle(card.DurationString, fyne.TextAlignTrailing, fyne.TextStyle{
					Monospace: true,
				}),
				card.Options,
			), layout.NewSpacer()),
		text,
	)

	return widget.NewSimpleRenderer(c)
}
