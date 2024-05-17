package widgets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type Paged fyne.Container

func NewPaged(mainPage fyne.CanvasObject) *Paged {
	return (*Paged)(container.NewStack(mainPage))
}

func (p *Paged) Container() *fyne.Container {
	return (*fyne.Container)(p)
}

func (p *Paged) SetPage(page fyne.CanvasObject) {
	for _, obj := range p.Objects {
		obj.Hide()
	}
	cont := p.Container()
	cont.Add(
		container.NewBorder(container.NewHBox(
			&widget.Button{
				Icon:       theme.MediaSkipPreviousIcon(),
				Importance: widget.LowImportance,

				OnTapped: func() {
					p.Objects = p.Objects[:len(p.Objects)-1]
					p.Objects[len(p.Objects)-1].Show()

					cont.Refresh()
				},
			},
		), nil, nil, nil, page),
	)
}
