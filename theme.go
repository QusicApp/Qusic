package main

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

type myTheme struct{}

var _ fyne.Theme = (*myTheme)(nil)

var BackgroundColor = color.RGBA{18, 18, 18, 255}
var OverlayBackgroundColor = color.RGBA{25, 25, 25, 255}

func (m myTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return BackgroundColor
	case theme.ColorNamePrimary:
		return m.Color(theme.ColorNameForeground, variant)
	case theme.ColorNameOverlayBackground:
		return OverlayBackgroundColor
	case theme.ColorNameHyperlink:
		return m.Color(theme.ColorNameForeground, variant)
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
	if style.Monospace {
		return resourceRobotoMonoLightTtf
	}
	return resourceRobotoRegularTtf
}

func (m myTheme) Size(name fyne.ThemeSizeName) float32 {
	//if name == theme.SizeNameInputRadius {
	//	return 20
	//}
	if name == theme.SizeNamePadding {
		return 5
	}
	return theme.DefaultTheme().Size(name)
}
