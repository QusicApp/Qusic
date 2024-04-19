package widgets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type ImageButton struct {
	widget.DisableableWidget
	Image    *canvas.Image
	OnTapped func()

	icon *widget.Icon
	rect *canvas.Rectangle

	focused bool
}

func (img *ImageButton) CreateRenderer() fyne.WidgetRenderer {
	img.ExtendBaseWidget(img)
	img.icon = widget.NewIcon(theme.MediaPlayIcon())
	img.icon.Hide()

	img.rect = canvas.NewRectangle(colorOpacity(theme.DisabledButtonColor(), 0.5))

	img.rect.Resize(img.Image.Size())
	img.rect.Hide()

	img.Image.FillMode = canvas.ImageFillContain

	stack := container.NewStack(img.Image, img.rect, img.icon)

	return widget.NewSimpleRenderer(stack)
}

func (img *ImageButton) Tapped(*fyne.PointEvent) {
	if !img.Disabled() && img.OnTapped != nil {
		img.OnTapped()
	}
}

func (img *ImageButton) Cursor() desktop.Cursor {
	return desktop.PointerCursor
}

func (img *ImageButton) MouseIn(*desktop.MouseEvent) {
	img.focused = true
	img.icon.Show()
	img.rect.Show()
}

func (img *ImageButton) MouseMoved(*desktop.MouseEvent) {

}

func (img *ImageButton) MouseOut() {
	img.focused = false
	img.icon.Hide()
	img.rect.Hide()
}
