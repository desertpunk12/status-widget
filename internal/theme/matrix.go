package theme

import "image/color"

// MatrixColors contains the color palette for the Matrix-style widget
type MatrixColors struct {
	Background  color.Color
	TitleBG     color.Color
	PrimaryText color.Color
	DimText     color.Color
	Border      color.Color
	Highlight   color.Color
	Warning     color.Color
}

// NewMatrixColors returns the Matrix color palette
func NewMatrixColors() *MatrixColors {
	return &MatrixColors{
		Background:  color.RGBA{0, 5, 0, 255},      // #000500
		TitleBG:     color.RGBA{0, 17, 0, 230},     // #001100 with 90% opacity
		PrimaryText: color.RGBA{0, 255, 65, 255},   // #00FF41
		DimText:     color.RGBA{0, 143, 17, 255},   // #008F11
		Border:      color.RGBA{0, 214, 57, 255},   // #00D639
		Highlight:   color.RGBA{175, 255, 74, 255}, // #AFFF4A
		Warning:     color.RGBA{255, 51, 0, 255},   // #FF3300
	}
}

// Widget dimensions
const (
	MinWidgetWidth = 280
	TitleBarHeight = 24
	PaddingOuter   = 10
	PaddingInner   = 6
	SectionGap     = 8
	FontSizeTitle  = 14
	FontSizeBody   = 13
	FontSizeSmall  = 11

	// Margins from screen edges
	MarginRight  = 16 // Keep widget away from right screen edge
	MarginBottom = 16 // Keep widget away from bottom screen edge
)

// Widget title text
const (
	DefaultTitle = "MATRIX WIDGET"
)
