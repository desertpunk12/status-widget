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

	// Status indicator colors (colored circles)
	StatusGreen  color.Color
	StatusYellow color.Color
	StatusOrange color.Color
	StatusRed    color.Color
	StatusBlack  color.Color
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

		// Status indicator colors
		StatusGreen:  color.RGBA{45, 189, 0, 255},  // #2DBD00
		StatusYellow: color.RGBA{255, 214, 0, 255}, // #FFD600
		StatusOrange: color.RGBA{255, 153, 0, 255}, // #FF9900
		StatusRed:    color.RGBA{255, 48, 0, 255},  // #FF3000
		StatusBlack:  color.RGBA{26, 26, 26, 255},  // #1A1A1A
	}
}

// Widget dimensions
const (
	MinWidgetWidth = 340
	TitleBarHeight = 32
	PaddingOuter   = 12
	PaddingInner   = 8
	SectionGap     = 10
	FontSizeTitle  = 18
	FontSizeBody   = 16
	FontSizeSmall  = 14

	// Margins from screen edges
	MarginRight  = 16 // Keep widget away from right screen edge
	MarginBottom = 16 // Keep widget away from bottom screen edge
)

// Widget title text
const (
	DefaultTitle = "Z.AI API STATUS"
)
