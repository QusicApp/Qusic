package widgets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

type TappableRichText struct {
	widget.RichText
	OnTapped func(*fyne.PointEvent)
}

func NewRichText(s ...widget.RichTextSegment) *TappableRichText {
	t := new(TappableRichText)
	t.ExtendBaseWidget(t)
	t.Segments = s

	return t
}

func (r *TappableRichText) Tapped(e *fyne.PointEvent) {
	if r.OnTapped != nil {
		r.OnTapped(e)
	}
}
