package widgets

import (
	"net/url"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type HyperlinkSegment struct {
	Style widget.RichTextStyle
	Text  string
	URL   *url.URL

	// OnTapped overrides the default `fyne.OpenURL` call when the link is tapped
	//
	// Since: 2.4
	OnTapped func()
}

// Inline returns true as hyperlinks are inside other elements.
func (h *HyperlinkSegment) Inline() bool {
	return true
}

// Textual returns the content of this segment rendered to plain text.
func (h *HyperlinkSegment) Textual() string {
	return h.Text
}

// Visual returns the hyperlink widget required to render this segment.
func (h *HyperlinkSegment) Visual() fyne.CanvasObject {
	link := widget.NewHyperlink(h.Text, h.URL)
	link.Alignment = h.Style.Alignment
	link.OnTapped = h.OnTapped
	return link //&fyne.Container{Layout: &unpadTextWidgetLayout{}, Objects: []fyne.CanvasObject{link}}
}

// Update applies the current state of this hyperlink segment to an existing visual.
func (h *HyperlinkSegment) Update(o fyne.CanvasObject) {
	link := o.(*widget.Hyperlink)
	link.Text = h.Text
	link.URL = h.URL
	link.Alignment = h.Style.Alignment
	link.OnTapped = h.OnTapped

	link.Resize(fyne.NewSquareSize(h.size()))

	link.Refresh()
}

func (t *HyperlinkSegment) size() float32 {
	if t.Style.SizeName != "" {
		return fyne.CurrentApp().Settings().Theme().Size(t.Style.SizeName)
	}

	return theme.TextSize()
}

// Select tells the segment that the user is selecting the content between the two positions.
func (h *HyperlinkSegment) Select(begin, end fyne.Position) {
	// no-op: this will be added when we progress to editor
}

// SelectedText should return the text representation of any content currently selected through the Select call.
func (h *HyperlinkSegment) SelectedText() string {
	// no-op: this will be added when we progress to editor
	return ""
}

// Unselect tells the segment that the user is has cancelled the previous selection.
func (h *HyperlinkSegment) Unselect() {
	// no-op: this will be added when we progress to editor
}
