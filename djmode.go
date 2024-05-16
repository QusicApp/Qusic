package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type Wheel struct {
	widget.BaseWidget
	img *canvas.Image
}

func (w *Wheel) CreateRenderer() fyne.WidgetRenderer {
	w.ExtendBaseWidget(w)
	return widget.NewSimpleRenderer(w.img)
}

func djModePage() fyne.CanvasObject {
	speedSlider := widget.NewSlider(-2, 2)
	speedSlider.OnChanged = func(f float64) {
		switch f {
		case -2:
			f = 0.25
		case -1:
			f = 0.5
		case 0:
			f = 1
		case 1:
			f = 1.5
		case 2:
			f = 2
		}
		player.SetSpeed(f)
	}

	vinyl := canvas.NewImageFromFile("vinyl.png")
	vinyl.FillMode = canvas.ImageFillOriginal

	wheel := &Wheel{img: vinyl}

	bottom := container.NewGridWithColumns(3,
		container.NewBorder(container.NewCenter(widget.NewRichTextFromMarkdown("## Speed")), nil, widget.NewLabel("x0.25"), widget.NewLabel("x2"), speedSlider),
	)
	border := container.NewBorder(nil, bottom, nil, nil, container.NewCenter(wheel))

	return border
}
