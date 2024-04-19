package widgets

import "image/color"

func colorOpacity(c color.Color, o float32) color.Color {
	r, g, b, _ := c.RGBA()
	return color.RGBA{byte(r), byte(g), byte(b), uint8(o * 255)}
}
