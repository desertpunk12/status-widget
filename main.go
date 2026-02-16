package main

import (
	"log"

	"status-widget/internal/widget"

	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
	// Create widget instance
	w := widget.New()

	// Configure window
	ebiten.SetWindowSize(w.Layout(0, 0))

	// Get monitor dimensions and position widget at bottom right, just above taskbar
	monitorWidth, monitorHeight := ebiten.Monitor().Size()
	widgetWidth, widgetHeight := w.Layout(0, 0)
	taskBarHeight := 48 // Approximate Windows taskbar/clock area height

	// Position at bottom right with margins
	// Keep widget away from right and bottom edges of screen
	initialX := monitorWidth - widgetWidth - 16
	initialY := monitorHeight - widgetHeight - taskBarHeight - 16
	ebiten.SetWindowPosition(initialX, initialY)

	ebiten.SetWindowDecorated(false) // Remove OS window decorations
	ebiten.SetWindowFloating(true)   // Always-on-top
	ebiten.SetWindowTitle("Status Widget")

	// Configure transparent window
	opts := &ebiten.RunGameOptions{
		ScreenTransparent: true,
		SkipTaskbar:       true,
	}

	// Run game
	if err := ebiten.RunGameWithOptions(w, opts); err != nil {
		log.Fatal(err)
	}
}
