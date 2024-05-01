package main

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

type myTheme struct{}

var _ fyne.Theme = (*myTheme)(nil)

func (m myTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	if name == theme.ColorNameBackground && !apprunning {
		return color.Transparent
	}
	return theme.DefaultTheme().Color(name, variant)
}

func (m myTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (m myTheme) Font(style fyne.TextStyle) fyne.Resource {
	if style.Monospace {
		return theme.DefaultTheme().Font(style)
	}
	if style.Bold {
		if style.Italic {
			return resourceRobotoBoldItalicTtf
		}
		return resourceRobotoBoldTtf
	}
	if style.Italic {
		return resourceRobotoItalicTtf
	}
	return resourceRobotoRegularTtf
}

func (m myTheme) Size(name fyne.ThemeSizeName) float32 {
	return theme.DefaultTheme().Size(name)
}
