package widget

import (
	"image"
	"image/color"
	"log"
	"time"

	"status-widget/internal/font"
	"status-widget/internal/theme"
	"status-widget/internal/win"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// Widget is the main structure for the status widget
type Widget struct {
	// Drag state
	dragging    bool
	dragStartX  int
	dragStartY  int
	dragWindowX int
	dragWindowY int

	// Window visibility (hidden = moved off-screen)
	hidden    bool
	restoreX  int
	restoreY  int
	toggleReq chan struct{}

	// Content
	messages  []string
	titleText string

	// Resources
	fontManager *font.Manager

	// Theme
	colors *theme.MatrixColors

	// Dimensions
	width  int
	height int

	// Widget monitoring
	cpuUsage     float64 // CPU usage percentage
	memUsageMB   float64 // Memory usage in MB
	lastCPUTime  time.Time
	lastCPUUsage float64
	lastUpdate   time.Time

	// Z.ai API monitoring
	apiKey        string
	zaiData       *ZaiResponse
	zaiStatus     string    // "ok", "low", "warn", "crit", "none"
	usagePercent  float64   // 0-100% for color gradient
	resetTime     time.Time // API quota reset time
	lastApiUpdate time.Time

	// Taskbar icon state (last rendered availability, -1 = none yet)
	taskbarIconPct int

	// External integrations
	trayUpdater func(icon image.Image, tooltip string) // System tray refresh callback
	quitChan    <-chan struct{}                        // External quit request (e.g. tray menu)

	// Task Manager comparison values (for debugging)
	taskManagerCPU    string
	taskManagerMEM    float64
	lastTaskMgrUpdate time.Time
}

// New creates a new widget instance
func New() *Widget {
	apiKey, _ := getApiKey()
	w := &Widget{
		titleText:   theme.DefaultTitle,
		apiKey:      apiKey,
		messages:    []string{"> Initializing...", "> Loading API status..."},
		colors:      theme.NewMatrixColors(),
		fontManager: font.NewManager(),
		width:       theme.DefaultWidgetWidth,
		height:      theme.DefaultWidgetHeight,
		toggleReq:   make(chan struct{}, 1),
	}
	// Initial API fetch
	w.updateZaiAPI()

	// Mark taskbar icon as not yet rendered; it is generated on the first
	// Update() frame once a graphics context exists.
	w.taskbarIconPct = taskbarIconNone
	return w
}

// Update handles the game update loop
func (w *Widget) Update() error {
	// Check for external quit request (e.g. tray menu Quit)
	select {
	case <-w.quitChan:
		return ebiten.Termination
	default:
	}

	// Check for ESC key to exit (tap-safe: IsKeyJustPressed catches presses
	// shorter than one frame, unlike the polled IsKeyPressed)
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		log.Printf("DEBUG: ESC pressed, shift=%v hidden=%v", ebiten.IsKeyPressed(ebiten.KeyShift), w.hidden)
		if ebiten.IsKeyPressed(ebiten.KeyShift) {
			return ebiten.Termination
		}

		// ESC alone hides the widget; double-click the tray icon to bring
		// it back
		if !w.hidden {
			w.setHidden(true)
		}
	}

	// Apply pending visibility toggles requested from the tray
	select {
	case <-w.toggleReq:
		w.setHidden(!w.hidden)
	default:
	}

	w.updateDrag()

	// Update system stats every second
	if time.Since(w.lastUpdate) >= time.Second {
		w.updateSystemStats()
		w.lastUpdate = time.Now()
	}

	// Update Task Manager comparison stats every 5 seconds (PowerShell is expensive)
	if time.Since(w.lastTaskMgrUpdate) >= 5*time.Second {
		w.lastTaskMgrUpdate = time.Now()
	}

	// Update Z.ai API status every 30 seconds
	if time.Since(w.lastApiUpdate) >= 30*time.Second {
		w.updateZaiAPI()
		w.lastApiUpdate = time.Now()
	}

	// Refresh taskbar icon and title when availability changes
	w.updateTaskbarIcon()

	return nil
}

// Draw renders the widget
func (w *Widget) Draw(screen *ebiten.Image) {
	// Hidden window is off-screen; skip rendering entirely
	if w.hidden {
		return
	}

	// Draw semi-transparent black background (opacity ~85%)
	vector.DrawFilledRect(screen, 0, 0, float32(w.width), float32(w.height), color.RGBA{0, 0, 0, 220}, false)

	// Draw title bar
	w.drawTitleBar(screen)

	// Draw messages
	w.drawMessages(screen)

	// Draw CPU and memory stats at bottom right (status bar style)
	w.drawBottomStatusBar(screen)
}

// Layout returns the widget dimensions
func (w *Widget) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return w.width, w.height
}

// AddMessage adds a new message to the widget
func (w *Widget) AddMessage(msg string) {
	w.messages = append(w.messages, msg)
	// Keep only the last 20 messages
	if len(w.messages) > 20 {
		w.messages = w.messages[len(w.messages)-20:]
	}
}

// SetTrayUpdater registers a callback invoked from the game loop whenever
// the system tray icon should refresh. The icon image carries straight or
// premultiplied alpha; the receiver handles format conversion.
func (w *Widget) SetTrayUpdater(fn func(icon image.Image, tooltip string)) {
	w.trayUpdater = fn
}

// SetQuitChannel registers a channel whose closure requests termination,
// e.g. from the system tray Quit menu item.
func (w *Widget) SetQuitChannel(ch <-chan struct{}) {
	w.quitChan = ch
}

// ToggleVisible requests the window to be shown/hidden. Safe to call from
// any goroutine (e.g. the tray thread); the switch is applied on the next
// Update tick to keep window state confined to the game loop.
func (w *Widget) ToggleVisible() {
	select {
	case w.toggleReq <- struct{}{}:
	default: // A toggle is already pending; drop duplicates
	}
}

// setHidden hides or shows the window. Preferred path is a true Win32 hide
// (SW_HIDE): like Discord, the window vanishes entirely, taskbar button
// included, and reappears at its old spot on show. If the Win32 handle
// cannot be resolved, it falls back to parking the window off-screen.
// Must be called from the game loop.
func (w *Widget) setHidden(hidden bool) {
	if hidden == w.hidden {
		return
	}
	w.dragging = false // Cancel any in-flight drag

	hwnd, hwerr := win.GameWindow()
	log.Printf("DEBUG: setHidden(%v) hwnd=%d err=%v", hidden, hwnd, hwerr)
	if hwerr == nil {
		if hidden {
			win.Hide(hwnd)
		} else {
			win.Show(hwnd)
		}
	} else if hidden {
		w.restoreX, w.restoreY = ebiten.WindowPosition()
		ebiten.SetWindowPosition(theme.WindowHiddenPos, theme.WindowHiddenPos)
	} else {
		ebiten.SetWindowPosition(w.restoreX, w.restoreY)
	}
	w.hidden = hidden
}

// updateDrag handles the drag-to-move functionality
func (w *Widget) updateDrag() {
	cx, cy := ebiten.CursorPosition()

	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		if !w.dragging {
			// Start dragging - store initial cursor position
			w.dragging = true
			w.dragStartX = cx
			w.dragStartY = cy
			w.dragWindowX, w.dragWindowY = ebiten.WindowPosition()
		} else if w.dragging {
			// Continue dragging - calculate delta from initial position
			deltaX := cx - w.dragStartX
			deltaY := cy - w.dragStartY
			newX := w.dragWindowX + deltaX
			newY := w.dragWindowY + deltaY
			ebiten.SetWindowPosition(newX, newY)
			// Update stored window position for next frame
			w.dragWindowX = newX
			w.dragWindowY = newY
		}
	} else {
		w.dragging = false
	}
}
