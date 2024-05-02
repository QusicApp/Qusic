package widgets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

type Button struct {
	widget.Button
}

func NewButton(text string, tapped func()) *Button {
	t := new(Button)
	t.ExtendBaseWidget(t)
	t.Text = text
	t.OnTapped = tapped

	return t
}

func NewButtonWithIcon(text string, icon fyne.Resource, tapped func()) *Button {
	t := new(Button)
	t.ExtendBaseWidget(t)
	t.Text = text
	t.OnTapped = tapped
	t.Icon = icon

	return t
}

func (r *Button) Cursor() desktop.Cursor {
	return desktop.PointerCursor
}
