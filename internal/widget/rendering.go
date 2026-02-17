package widget

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"status-widget/internal/font"
	"status-widget/internal/theme"
)

// drawTitleBar renders the title bar
func (w *Widget) drawTitleBar(screen *ebiten.Image) {
	// Draw title text with uniform padding
	titleX := theme.PaddingOuter
	titleY := theme.PaddingOuter // Same padding as left side
	font.DrawColoredText(screen, w.titleText, w.fontManager.TitleFace(), titleX, titleY, w.colors.PrimaryText)
}

// drawMessages renders the message list
func (w *Widget) drawMessages(screen *ebiten.Image) {
	// Starting position for messages
	x := theme.PaddingOuter
	y := theme.TitleBarHeight + theme.SectionGap + 10

	// Draw each message
	for i, msg := range w.messages {
		face := w.fontManager.BodyFace()

		if i == 1 {
			// Draw usage line with gradient color (green at 0% to red at 100%)
			usageColor := w.getUsageColorGradient()
			font.DrawColoredText(screen, msg, face, x, y, usageColor)
		} else {
			// Alternate colors for visual interest on other lines
			textColor := w.colors.PrimaryText
			if i%2 == 1 {
				textColor = w.colors.DimText
			}
			font.DrawColoredText(screen, msg, face, x, y, textColor)
		}

		// Draw colored status circle for first message
		if i == 0 {
			circleColor := w.getZaiStatusColor()
			textWidth, _ := text.Measure(msg, face, 0)
			circleX := float32(x) + float32(textWidth) + 12
			circleY := float32(y) + 11
			vector.FillCircle(screen, circleX, circleY, 8, color.RGBA{0, 0, 0, 255}, true)
			vector.FillCircle(screen, circleX, circleY, 7, circleColor, true)
		}

		y += 22 // Line height for larger font
	}
}

// Draw CPU and memory stats at bottom right (status bar style)
func (w *Widget) drawBottomStatusBar(screen *ebiten.Image) {
	// Create a face that's between SmallFace (14px) and BodyFace (16px) - 11px
	statusFace := &text.GoTextFace{
		Source: w.fontManager.SmallFace().Source,
		Size:   11,
	}

	// Format status text: "CPU 0.5% | MEM 45MB"
	statusText := fmt.Sprintf("CPU %s | MEM %s", w.formatCPUText(), w.formatMEMText())

	// Measure text to position at bottom right
	textWidth, textHeight := text.Measure(statusText, statusFace, 0)

	// Position at bottom right with padding
	padding := float32(theme.PaddingOuter)
	x := float32(w.width) - float32(textWidth) - padding
	y := float32(w.height) - float32(textHeight) - padding - 6

	// Draw with dim text color
	font.DrawColoredText(screen, statusText, statusFace, int(x), int(y), w.colors.DimText)
}

// getUsageColorGradient returns a color from green (0%) to red (100%)
func (w *Widget) getUsageColorGradient() color.Color {
	percent := w.usagePercent
	if percent < 0 {
		percent = 0
	}
	if percent > 100 {
		percent = 100
	}

	// Normalize to 0-1 range
	t := percent / 100.0

	// Interpolate from green (0, 255, 0) to red (255, 0, 0)
	// At t=0: green, At t=1: red
	r := uint8(255 * t)
	g := uint8(255 * (1 - t))
	b := uint8(0)

	return color.RGBA{r, g, b, 255}
}

// getZaiStatusColor returns appropriate color for Z.ai status
func (w *Widget) getZaiStatusColor() color.Color {
	switch w.zaiStatus {
	case "ok":
		return w.colors.StatusGreen
	case "low":
		return w.colors.StatusYellow
	case "warn":
		return w.colors.StatusOrange
	case "crit":
		return w.colors.StatusRed
	case "none":
		return w.colors.StatusBlack
	default:
		return w.colors.PrimaryText
	}
}

// Draw Task Manager comparison values at bottom (for debugging)
func (w *Widget) drawTaskManagerComparison(screen *ebiten.Image) {
	// Check if Task Manager values were fetched
	if w.taskManagerCPU == "" {
		return // No data yet
	}

	// Format comparison text
	// Line 1: Widget CPU vs Task Manager CPU
	// Line 2: Widget MEM vs Task Manager MEM
	widgetCPU := w.formatCPUText()
	widgetMEM := w.formatMEMText()
	line1 := fmt.Sprintf("Widget: %s | TM: %s", widgetCPU, w.taskManagerCPU)
	line2 := fmt.Sprintf("Widget: %s | TM: %.2f MB", widgetMEM, w.taskManagerMEM)

	// Position at bottom, below process ID
	x := theme.PaddingOuter
	y := float32(w.height) - 50 // 50px from bottom

	// Draw comparison text with dim color
	font.DrawColoredText(screen, line1, w.fontManager.SmallFace(), x, int(y), w.colors.DimText)
	font.DrawColoredText(screen, line2, w.fontManager.SmallFace(), x, int(y)+12, w.colors.DimText)
}
