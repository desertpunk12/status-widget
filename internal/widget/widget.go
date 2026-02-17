package widget

import (
	"image/color"
	"time"

	"status-widget/internal/font"
	"status-widget/internal/theme"

	"github.com/hajimehoshi/ebiten/v2"
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
		width:       180,
		height:      180,
	}
	// Initial API fetch
	w.updateZaiAPI()
	return w
}

// Update handles the game update loop
func (w *Widget) Update() error {
	// Check for ESC key to exit
	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		return ebiten.Termination
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

	return nil
}

// Draw renders the widget
func (w *Widget) Draw(screen *ebiten.Image) {
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
