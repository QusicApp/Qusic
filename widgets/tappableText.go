package widgets

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
)

type Text struct {
	base widget.BaseWidget

	*canvas.Text

	OnTapped func()
}

func NewText(text string, color color.Color, tapped func()) *Text {
	t := &Text{
		Text:     canvas.NewText(text, color),
		OnTapped: tapped,
	}
	return t
}

func (t *Text) CreateRenderer() fyne.WidgetRenderer {
	t.base.ExtendBaseWidget(t)
	return widget.NewSimpleRenderer(t.Text)
}

func (t *Text) Tapped(*fyne.PointEvent) {
	if t.OnTapped != nil {
		t.OnTapped()
	}
}
