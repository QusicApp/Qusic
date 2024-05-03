package widgets

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type TextSegment struct {
	Style widget.RichTextStyle
	Text  string

	OnTapped func()
}

// Inline should return true if this text can be included within other elements, or false if it creates a new block.
func (t *TextSegment) Inline() bool {
	return t.Style.Inline
}

// Textual returns the content of this segment rendered to plain text.
func (t *TextSegment) Textual() string {
	return t.Text
}

// Visual returns the graphical elements required to render this segment.
func (t *TextSegment) Visual() fyne.CanvasObject {
	obj := NewText(t.Text, t.color(), t.OnTapped)

	t.Update(obj)
	return obj
}

// Update applies the current state of this text segment to an existing visual.
func (t *TextSegment) Update(o fyne.CanvasObject) {
	obj := o.(*Text)
	obj.Text.Text = t.Text
	obj.Color = t.color()
	obj.Alignment = t.Style.Alignment
	obj.TextStyle = t.Style.TextStyle
	obj.TextSize = t.size()
	obj.Refresh()
}

// Select tells the segment that the user is selecting the content between the two positions.
func (t *TextSegment) Select(begin, end fyne.Position) {
	// no-op: this will be added when we progress to editor
}

// SelectedText should return the text representation of any content currently selected through the Select call.
func (t *TextSegment) SelectedText() string {
	// no-op: this will be added when we progress to editor
	return ""
}

// Unselect tells the segment that the user is has cancelled the previous selection.
func (t *TextSegment) Unselect() {
	// no-op: this will be added when we progress to editor
}

func (t *TextSegment) color() color.Color {
	if t.Style.ColorName != "" {
		return theme.Color(t.Style.ColorName)
	}

	return theme.ForegroundColor()
}

func (t *TextSegment) size() float32 {
	if t.Style.SizeName != "" {
		return theme.Size(t.Style.SizeName)
	}

	return theme.TextSize()
}
