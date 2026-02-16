# Status Widget - Agent Guidelines

## Project Overview
Matrix-styled desktop system status widget built with Go and Ebitengine v2. Displays real-time CPU/memory usage with draggable, always-on-top UI.

**Stack**: Go 1.25+ | Ebitengine v2.9.8 | golang.org/x/image

---

## Build Commands

```bash
# Build executable
go build -o status-widget.exe .

# Run directly
go run .

# Clean build artifacts
go clean

# Format all Go files
go fmt ./...

# Vet code for common issues
go vet ./...

# Run tests (none currently exist)
go test ./...

# Run single test (when tests exist)
go test -run TestFunctionName ./internal/widget
```

---

## Code Style Guidelines

### Import Organization
Imports must be grouped in three sections with blank lines between:

```go
import (
    // Standard library
    "fmt"
    "image/color"
    "time"

    // Local packages (status-widget/internal/...)
    "status-widget/internal/font"
    "status-widget/internal/theme"

    // Third-party packages
    "github.com/hajimehoshi/ebiten/v2"
)
```

Use `goimports` to auto-format imports: `goimports -w .`

### Naming Conventions

- **Exported**: PascalCase (`NewWidget`, `MatrixColors`, `Update`)
- **Private**: camelCase (`updateDrag`, `cpuUsage`, `fontManager`)
- **Constants**: PascalCase (`DefaultTitle`, `TitleBarHeight`)
- **Interfaces**: Single-method interfaces end with `-er` (`Drawer`, `Updater`)

### Struct Design

Group fields logically with section comments:

```go
type Widget struct {
    // Drag state
    dragging    bool
    dragStartX  int

    // Content
    messages  []string
    titleText string

    // Resources
    fontManager *font.Manager
}
```

### Documentation

- Exported types require package-level doc comments
- Exported functions require godoc comments
- Inline comments explain "why", not "what"
- Document edge cases and invariants

```go
// Widget manages the main application state and rendering.
type Widget struct {
    // dragging indicates whether widget is currently being moved by user.
    dragging bool
}

// Update handles the game loop, processing input and updating state.
// Returns ebiten.Termination if ESC key is pressed.
func (w *Widget) Update() error {
    // Check for ESC key to exit
    if ebiten.IsKeyPressed(ebiten.KeyEscape) {
        return ebiten.Termination
    }
    return nil
}
```

### Error Handling

- Use `error` return values for recoverable conditions
- Use `log.Fatal()` only in main() for unrecoverable errors
- Always check errors: `if err != nil { return err }`
- Wrap errors with context when useful: `fmt.Errorf("failed to load font: %w", err)`

### Formatting

- Use `gofmt` standard formatting (enforced by tooling)
- Run `go fmt ./...` before committing
- Max line length: 120 characters (soft limit, prefer ~80-100)
- One blank line between functions, two between major sections

### Constants and Magic Numbers

Define all dimensions and colors in `internal/theme/`:

```go
const (
    MinWidgetWidth = 280
    TitleBarHeight = 24
    PaddingOuter   = 10
)

// Never hardcode:
// vector.DrawFilledRect(screen, 0, 0, 280, 100, ...) // ❌
// Instead:
// vector.DrawFilledRect(screen, 0, 0, float32(MinWidgetWidth), 100, ...) // ✅
```

---

## Project Structure

```
status-widget/
├── main.go                      # Entry point, window config, positioning
├── internal/
│   ├── widget/
│   │   ├── widget.go            # Core widget struct, New(), Update(), Draw(), Layout()
│   │   ├── rendering.go         # Drawing functions (drawTitleBar, drawMessages)
│   │   ├── system_stats.go      # System monitoring (CPU, memory)
│   │   └── zai_api.go           # Z.ai API integration (fetching, parsing, formatting)
│   ├── theme/matrix.go          # Color palette & dimension constants
│   └── font/font.go             # Font management & text rendering
└── assets/fonts/                # Custom font files (if added)
```

**Guidelines**:
- Put application logic in `internal/` packages (no external imports)
- External code in `main.go` only
- Constants and theme in `internal/theme/`
- Reusable UI components in `internal/widget/`
- Separate concerns: API, rendering, system stats, and core logic in different files

---

## Ebitengine-Specific Patterns

### Game Loop Pattern

```go
// Update() - Process input, update state (no rendering)
func (w *Widget) Update() error {
    // Handle input
    // Update game state
    return nil
}

// Draw() - Render only (no state changes)
func (w *Widget) Draw(screen *ebiten.Image) {
    // Pure rendering using current state
}

// Layout() - Return dimensions
func (w *Widget) Layout(outsideWidth, outsideHeight int) (int, int) {
    return w.width, w.height
}
```

### Window Configuration

```go
// Window positioning - account for taskbar
monitorWidth, monitorHeight := ebiten.Monitor().Size()
taskBarHeight := 48 // Windows taskbar height
initialX := monitorWidth - widgetWidth - 16
initialY := monitorHeight - widgetHeight - taskBarHeight - 16
ebiten.SetWindowPosition(initialX, initialY)

// Always-on-top & transparency
ebiten.SetWindowFloating(true)
opts := &ebiten.RunGameOptions{
    ScreenTransparent: true,
    SkipTaskbar:       true,
}
```

---

## Testing Strategy

Currently no tests exist. When adding tests:

- Test files: `widget_test.go` (same package as code)
- Table-driven tests for multiple scenarios
- Mock Ebiten interfaces where needed
- Test file: `go test -run TestWidgetUpdate ./internal/widget`

---

## Performance Notes

- Widget updates system stats every second (see `updateSystemStats()`)
- Keep last 20 messages max to prevent unbounded growth
- Use `runtime.ReadMemStats()` sparingly (expensive call)
- Ebiten runs at 60 FPS by default

---

## Cursor/Copilot Rules

No existing rules found. Follow the guidelines in this file.
