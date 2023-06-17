package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
	"image/color"
	"strings"
)

type TestTheme struct {
	regular, bold, italic, boldItalic, monospace fyne.Resource
}

func (t *TestTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	return theme.DefaultTheme().Color(name, variant)
}

func (t *TestTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (m *TestTheme) Font(style fyne.TextStyle) fyne.Resource {
	if style.Monospace {
		return m.monospace
	}
	if style.Bold {
		if style.Italic {
			return m.boldItalic
		}
		return m.bold
	}
	if style.Italic {
		return m.italic
	}
	return m.regular
}

func (m *TestTheme) Size(name fyne.ThemeSizeName) float32 {
	return theme.DefaultTheme().Size(name)
}

func (t *TestTheme) SetFonts(regularFontPath string, monoFontPath string) {
	t.regular = theme.TextFont()
	t.bold = theme.TextBoldFont()
	t.italic = theme.TextItalicFont()
	t.boldItalic = theme.TextBoldItalicFont()
	t.monospace = theme.TextMonospaceFont()

	if regularFontPath != "" {
		t.regular = LoadCustomFont(regularFontPath, "Regular", t.regular)
		t.bold = LoadCustomFont(regularFontPath, "Bold", t.bold)
		t.italic = LoadCustomFont(regularFontPath, "Italic", t.italic)
		t.boldItalic = LoadCustomFont(regularFontPath, "BoldItalic", t.boldItalic)
	}
	if monoFontPath != "" {
		t.monospace = LoadCustomFont(monoFontPath, "Regular", t.monospace)
	} else {
		t.monospace = t.regular
	}
}

func LoadCustomFont(env, variant string, fallback fyne.Resource) fyne.Resource {
	variantPath := strings.Replace(env, "Regular", variant, -1)

	res, err := fyne.LoadResourceFromPath(variantPath)
	if err != nil {
		fyne.LogError("Error loading specified font", err)
		return fallback
	}

	return res
}
