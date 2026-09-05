package main

import (
	"image"
	"log"

	"status-widget/internal/tray"
	"status-widget/internal/widget"

	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
	// Create widget instance
	w := widget.New()

	// System tray (notification area) icon showing the available quota, next
	// to the battery/wifi indicators. The tray loop runs on its own goroutine;
	// the widget pushes icon updates through the tray updater callback.
	w.SetTrayUpdater(func(icon image.Image, tooltip string) {
		iconICO, err := tray.EncodeICO(icon)
		if err != nil {
			log.Printf("tray icon encode failed: %v", err)
			return
		}
		tray.Set(iconICO, tooltip)
	})
	w.SetQuitChannel(tray.QuitRequested())

	// Double-clicking the tray icon toggles the widget window visibility
	// (ESC hides it; see widget.Update)
	tray.SetOnDoubleClick(w.ToggleVisible)

	go tray.Run()

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

	// Configure transparent window. SkipTaskbar is disabled on purpose: the
	// widget shows in the Windows taskbar with a dynamic icon displaying the
	// available Z.ai quota percentage (see widget.updateTaskbarIcon).
	opts := &ebiten.RunGameOptions{
		ScreenTransparent: true,
	}

	// Run game
	err := ebiten.RunGameWithOptions(w, opts)

	// Remove the tray icon on any exit path (ESC+Shift or tray Quit)
	tray.Quit()

	if err != nil {
		log.Fatal(err)
	}
}
