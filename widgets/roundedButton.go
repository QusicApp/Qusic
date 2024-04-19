package widgets

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type RoundedButton struct {
	widget.DisableableWidget
	Icon     fyne.Resource
	Color    color.Color
	OnTapped func()
}

func (button *RoundedButton) CreateRenderer() fyne.WidgetRenderer {
	button.ExtendBaseWidget(button)

	res := theme.NewThemedResource(button.Icon)
	res.ColorName = theme.ColorNameButton

	circle := canvas.NewCircle(button.Color)
	circle.FillColor = button.Color

	return widget.NewSimpleRenderer(
		container.NewStack(circle, container.NewCenter(widget.NewIcon(res))),
	)
}

func (button *RoundedButton) MinSize() fyne.Size {
	b := &widget.Button{
		Importance: widget.LowImportance,
		OnTapped:   button.OnTapped,
		Icon:       button.Icon,
	}
	return b.MinSize()
}

func (button *RoundedButton) Cursor() desktop.Cursor {
	return desktop.PointerCursor
}

func (button *RoundedButton) Tapped(*fyne.PointEvent) {
	if button.Disabled() {
		return
	}
	if button.OnTapped != nil {
		button.OnTapped()
	}
}
