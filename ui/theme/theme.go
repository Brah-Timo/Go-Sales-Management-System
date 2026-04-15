// Package theme fournit un thème personnalisé pour Gestion Commerciale Pro
// avec support de la police Amiri (arabe) + NotoSans (latin) et un design moderne.
package theme

import (
	_ "embed"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// ── Polices embarquées ────────────────────────────────────────────────────────
// Amiri : police arabe haute qualité (static, GSUB complet)

//go:embed fonts/Amiri-Regular.ttf
var amiriRegular []byte

//go:embed fonts/Amiri-Bold.ttf
var amiriBold []byte

//go:embed fonts/Amiri-Italic.ttf
var amiriItalic []byte

//go:embed fonts/Amiri-BoldItalic.ttf
var amiriBoldItalic []byte

// ── Palette de couleurs ───────────────────────────────────────────────────────

var (
	// Primaire – bleu marine algérien
	ColorPrimary      = color.RGBA{R: 0x1a, G: 0x3a, B: 0x6e, A: 0xff} // #1a3a6e
	ColorPrimaryLight = color.RGBA{R: 0x25, G: 0x53, B: 0x9e, A: 0xff} // #25539e
	ColorPrimaryDark  = color.RGBA{R: 0x0d, G: 0x21, B: 0x42, A: 0xff} // #0d2142

	// Accent – or algérien
	ColorAccent = color.RGBA{R: 0xf0, G: 0xb4, B: 0x29, A: 0xff} // #f0b429

	// Surfaces
	ColorBackground = color.RGBA{R: 0xf0, G: 0xf4, B: 0xf8, A: 0xff}
	ColorSurface    = color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}
	ColorSidebarBg  = color.RGBA{R: 0x12, G: 0x28, B: 0x4f, A: 0xff}

	// Statuts
	ColorSuccess = color.RGBA{R: 0x27, G: 0xae, B: 0x60, A: 0xff}
	ColorDanger  = color.RGBA{R: 0xe7, G: 0x4c, B: 0x3c, A: 0xff}
	ColorWarning = color.RGBA{R: 0xf3, G: 0x9c, B: 0x12, A: 0xff}
	ColorInfo    = color.RGBA{R: 0x29, G: 0x80, B: 0xb9, A: 0xff}

	// Texte
	ColorTextPrimary   = color.RGBA{R: 0x1a, G: 0x1a, B: 0x2e, A: 0xff}
	ColorTextSecondary = color.RGBA{R: 0x5a, G: 0x6a, B: 0x85, A: 0xff}
)

// ── GestionTheme ──────────────────────────────────────────────────────────────

// GestionTheme est le thème personnalisé de l'application
type GestionTheme struct{}

var _ fyne.Theme = (*GestionTheme)(nil)

// Color retourne les couleurs du thème
func (t GestionTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return ColorBackground
	case theme.ColorNameButton:
		return ColorPrimaryLight
	case theme.ColorNameDisabledButton:
		return color.RGBA{R: 0xcc, G: 0xcc, B: 0xcc, A: 0xff}
	case theme.ColorNameDisabled:
		return color.RGBA{R: 0x99, G: 0x99, B: 0x99, A: 0xff}
	case theme.ColorNameError:
		return ColorDanger
	case theme.ColorNameFocus:
		return ColorAccent
	case theme.ColorNameForeground:
		return ColorTextPrimary
	case theme.ColorNameHover:
		return color.RGBA{R: 0xe8, G: 0xf0, B: 0xfe, A: 0xff}
	case theme.ColorNameInputBackground:
		return ColorSurface
	case theme.ColorNameInputBorder:
		return color.RGBA{R: 0xb0, G: 0xc4, B: 0xde, A: 0xff}
	case theme.ColorNameMenuBackground:
		return ColorSurface
	case theme.ColorNameOverlayBackground:
		return ColorSurface
	case theme.ColorNamePlaceHolder:
		return ColorTextSecondary
	case theme.ColorNamePressed:
		return ColorPrimaryDark
	case theme.ColorNamePrimary:
		return ColorPrimary
	case theme.ColorNameScrollBar:
		return color.RGBA{R: 0xb0, G: 0xbe, B: 0xd4, A: 0xff}
	case theme.ColorNameSeparator:
		return color.RGBA{R: 0xd8, G: 0xe2, B: 0xef, A: 0xff}
	case theme.ColorNameShadow:
		return color.RGBA{R: 0x1a, G: 0x3a, B: 0x6e, A: 0x28}
	case theme.ColorNameSuccess:
		return ColorSuccess
	case theme.ColorNameWarning:
		return ColorWarning
	case theme.ColorNameHeaderBackground:
		return ColorPrimary
	case theme.ColorNameSelection:
		return color.RGBA{R: 0xf0, G: 0xb4, B: 0x29, A: 0x55}
	}
	return theme.DefaultTheme().Color(name, variant)
}

// Font retourne les polices Amiri (Arabic + Latin)
// Amiri est une police static avec GSUB complet → l'arabe s'affiche correctement
func (t GestionTheme) Font(style fyne.TextStyle) fyne.Resource {
	if style.Monospace {
		return theme.DefaultTheme().Font(style)
	}
	if style.Bold && style.Italic {
		return fyne.NewStaticResource("Amiri-BoldItalic.ttf", amiriBoldItalic)
	}
	if style.Bold {
		return fyne.NewStaticResource("Amiri-Bold.ttf", amiriBold)
	}
	if style.Italic {
		return fyne.NewStaticResource("Amiri-Italic.ttf", amiriItalic)
	}
	return fyne.NewStaticResource("Amiri-Regular.ttf", amiriRegular)
}

// Icon retourne les icônes Fyne par défaut
func (t GestionTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

// Size retourne les tailles du thème
func (t GestionTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNameText:
		return 14
	case theme.SizeNameHeadingText:
		return 24
	case theme.SizeNameSubHeadingText:
		return 17
	case theme.SizeNameCaptionText:
		return 12
	case theme.SizeNameInnerPadding:
		return 8
	case theme.SizeNamePadding:
		return 10
	case theme.SizeNameScrollBar:
		return 8
	case theme.SizeNameScrollBarSmall:
		return 4
	case theme.SizeNameSeparatorThickness:
		return 1
	case theme.SizeNameInputBorder:
		return 2
	case theme.SizeNameLineSpacing:
		return 5
	}
	return theme.DefaultTheme().Size(name)
}
