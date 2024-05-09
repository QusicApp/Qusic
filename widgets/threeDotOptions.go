package widgets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type ThreeDotOptions struct {
	widget.BaseWidget

	OnTapped func()

	circle0, circle1, circle2 *canvas.Circle
}

func NewThreeDotOptions(onTapped func()) *ThreeDotOptions {
	return &ThreeDotOptions{
		circle0: canvas.NewCircle(theme.ForegroundColor()),
		circle1: canvas.NewCircle(theme.ForegroundColor()),
		circle2: canvas.NewCircle(theme.ForegroundColor()),

		OnTapped: onTapped,
	}
}

func (t *ThreeDotOptions) CreateRenderer() fyne.WidgetRenderer {
	t.ExtendBaseWidget(t)

	return widget.NewSimpleRenderer(
		container.NewGridWithColumns(3, t.circle0, t.circle1, t.circle2),
	)
}

func (t *ThreeDotOptions) Hide() {
	t.circle0.Hide()
	t.circle1.Hide()
	t.circle2.Hide()
}

func (t *ThreeDotOptions) Show() {
	t.circle0.Show()
	t.circle1.Show()
	t.circle2.Show()
}

func (t *ThreeDotOptions) Tapped(*fyne.PointEvent) {
	if t.OnTapped != nil {
		t.OnTapped()
	}
}

func (*ThreeDotOptions) Cursor() desktop.Cursor {
	return desktop.PointerCursor
}
