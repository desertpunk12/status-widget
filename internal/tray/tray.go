// Package tray integrates the widget with the Windows notification area
// (system tray), showing the available Z.ai quota as a dynamic icon next
// to the battery, wifi and volume indicators.
package tray

import (
	"sync"

	"github.com/energye/systray"
)

var (
	mu      sync.Mutex
	running bool
	iconICO []byte // Latest icon/tooltip, applied on ready if set early
	tooltip string

	quitOnce sync.Once
	quitCh   = make(chan struct{})
)

// Set updates the tray icon and tooltip. Safe from any goroutine; values
// supplied before Run starts are applied once the tray becomes ready.
// Empty icons are ignored. systray calls are serialized under mu to avoid
// concurrent temp-file writes inside getlantern's Windows backend.
func Set(iconBytes []byte, tip string) {
	if len(iconBytes) == 0 {
		return
	}

	mu.Lock()
	defer mu.Unlock()

	iconICO = iconBytes
	tooltip = tip
	if running {
		systray.SetIcon(iconBytes)
		systray.SetTooltip(tip)
	}
}

// QuitRequested returns a channel that is closed when the user selects
// Quit from the tray menu (or the tray is shut down).
func QuitRequested() <-chan struct{} {
	return quitCh
}

// Quit removes the tray icon and stops its event loop. After Quit returns,
// the tray cannot be restarted.
func Quit() {
	systray.Quit()
}

// SetOnDoubleClick registers a callback invoked when the tray icon is
// double-clicked. Must be called before Run.
func SetOnDoubleClick(fn func()) {
	systray.SetOnDClick(func(_ systray.IMenu) {
		fn()
	})
}

// Run starts the notification area icon and blocks until Quit is called.
// Run must be called exactly once, from its own goroutine.
func Run() {
	systray.Run(onReady, onExit)
}

// onReady applies the latest icon/tooltip (if any yet) and builds the menu.
func onReady() {
	mu.Lock()
	running = true
	iconBytes, tip := iconICO, tooltip
	if len(iconBytes) > 0 {
		systray.SetIcon(iconBytes)
		systray.SetTooltip(tip)
	}
	mu.Unlock()

	quitItem := systray.AddMenuItem("Quit", "Quit Status Widget")
	quitItem.Click(func() {
		requestQuit()
	})
}

// onExit fires after the tray is removed, on any exit path.
func onExit() {
	requestQuit()
}

func requestQuit() {
	quitOnce.Do(func() { close(quitCh) })
}
