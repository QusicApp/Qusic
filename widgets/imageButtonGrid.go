package widgets

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type ImageButtonGrid struct {
	widget.DisableableWidget
	Image           *canvas.Image
	OnTapped        func()
	OnTappedOutside func()

	buttonGrid *fyne.Container

	focused bool
}

func (img *ImageButtonGrid) CreateRenderer() fyne.WidgetRenderer {
	img.ExtendBaseWidget(img)
	button := &RoundedButton{
		Icon:     theme.MediaPlayIcon(),
		Color:    color.White,
		OnTapped: img.OnTapped,
	}
	img.buttonGrid = container.NewGridWithRows(5,
		layout.NewSpacer(), layout.NewSpacer(), layout.NewSpacer(), layout.NewSpacer(),
		container.NewGridWithColumns(5,
			layout.NewSpacer(), layout.NewSpacer(), layout.NewSpacer(), layout.NewSpacer(),
			button))
	img.buttonGrid.Hide()
	stack := container.NewStack(img.Image, img.buttonGrid)

	return widget.NewSimpleRenderer(stack)
}

func (img *ImageButtonGrid) MouseIn(*desktop.MouseEvent) {
	img.focused = true
	img.buttonGrid.Show()
}

func (img *ImageButtonGrid) MouseMoved(*desktop.MouseEvent) {

}

func (img *ImageButtonGrid) MouseOut() {
	img.focused = false
	img.buttonGrid.Hide()
}

func (img *ImageButtonGrid) Tapped(*fyne.PointEvent) {
	if !img.Disabled() && img.OnTappedOutside != nil {
		img.OnTappedOutside()
	}
}
