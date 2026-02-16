package widget

import (
	"github.com/hajimehoshi/ebiten/v2"

	"status-widget/internal/font"
	"status-widget/internal/theme"
)

// drawTitleBar renders the title bar
func (w *Widget) drawTitleBar(screen *ebiten.Image) {
	// Draw title text with padding
	titleX := theme.PaddingOuter
	titleY := theme.PaddingOuter + 16 // Baseline adjustment
	font.DrawColoredText(screen, w.titleText, w.fontManager.TitleFace(), titleX, titleY, w.colors.PrimaryText)

	// Draw CPU and memory stats on separate lines in top right corner
	// CPU: ~18 chars (approx 126px), MEM: ~15 chars (approx 105px)
	// Position from right edge with proper spacing to ensure rendering
	// CPU text width approx 84px, MEM text width approx 84px
	// Total needed: 84 + 28 + 8 + 8 + 10 = 138px (within 280px)
	statsX := w.width - theme.PaddingOuter - 90
	statsY := theme.PaddingOuter + 16

	// Label "CPU" - first line
	cpuLabelText := "CPU"
	font.DrawColoredText(screen, cpuLabelText, w.fontManager.TitleFace(), statsX, statsY, w.colors.PrimaryText)

	// CPU value - first line (10px right of "CPU" label)
	cpuText := w.formatCPUText()
	cpuX := statsX + len(cpuLabelText)*7 + 10
	cpuY := statsY
	font.DrawColoredText(screen, cpuText, w.fontManager.TitleFace(), cpuX, cpuY, w.colors.PrimaryText)

	// Label "MEM" - second line, 14px below CPU (line spacing)
	memLabelText := "MEM"
	font.DrawColoredText(screen, memLabelText, w.fontManager.TitleFace(), statsX, statsY+14, w.colors.PrimaryText)

	// MEM value - second line (10px right of "MEM" label)
	memText := w.formatMEMText()
	memX := statsX + len(memLabelText)*7 + 10
	memY := statsY + 14
	font.DrawColoredText(screen, memText, w.fontManager.TitleFace(), memX, memY, w.colors.PrimaryText)
}

// drawMessages renders the message list
func (w *Widget) drawMessages(screen *ebiten.Image) {
	// Starting position for messages with section gap
	// Stats area: 24px (title) + 8px (gap) + 28px (CPU line) + 14px (MEM line) = 74px
	// Extra 4px padding + 14px line height = 78px
	x := theme.PaddingOuter
	y := theme.TitleBarHeight + theme.SectionGap + 4 + 14 // 24+8+4+14 = 50px

	// Draw each message
	for i, msg := range w.messages {
		// Alternate colors for visual interest
		textColor := w.colors.PrimaryText
		if i%2 == 1 {
			textColor = w.colors.DimText
		}

		font.DrawColoredText(screen, msg, w.fontManager.BodyFace(), x, y, textColor)
		y += 18 // Line height
	}
}
