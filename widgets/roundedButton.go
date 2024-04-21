package widgets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type RoundedButton struct {
	widget.DisableableWidget
	Icon fyne.Resource

	icon *widget.Icon

	OnTapped func()
}

func (button *RoundedButton) CreateRenderer() fyne.WidgetRenderer {
	button.ExtendBaseWidget(button)

	res := theme.NewThemedResource(button.Icon)
	res.ColorName = theme.ColorNameForeground

	circle := canvas.NewCircle(theme.ButtonColor())

	button.icon = widget.NewIcon(res)

	return widget.NewSimpleRenderer(
		container.NewStack(circle, container.NewCenter(button.icon)),
	)
}

func (button *RoundedButton) SetIcon(icon fyne.Resource) {
	button.Icon = icon
	res := theme.NewThemedResource(button.Icon)
	res.ColorName = theme.ColorNameForeground

	button.icon.SetResource(res)
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
