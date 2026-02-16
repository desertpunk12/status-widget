# Status Widget

An always-on-top, draggable, semi-transparent TUI-style status widget built with Go and Ebitengine.

## Features

- ✅ **Always-on-top** - Stays above all other windows
- ✅ **Semi-transparent background** - 85% opacity black background
- ✅ **Draggable** - Click and drag anywhere to move (smooth mouse following)
- ✅ **Matrix/terminal style** - Classic green-on-black aesthetic
- ✅ **Custom messages** - Display status updates and notifications
- ✅ **Widget resource monitoring** - Real-time CPU & memory usage of the widget
- ✅ **Auto-updating** - Stats refresh every second
- ✅ **Smart positioning** - Starts at bottom-right, just above the taskbar/clock
- ✅ **Compact design** - Optimized sizing with proper margins for aesthetics
- ✅ **Proper alignment** - CPU and MEM stats displayed cleanly with labels and aligned values

## Building

```bash
go build
```

## Running

```bash
./status-widget.exe  # Windows
./status-widget      # Linux/macOS
```

## What You'll See

The widget will appear at the **bottom-right corner of your screen**, just above the Windows taskbar/clock, with margins from screen edges:

```
┌─────────────────────────────┐
│ MATRIX WIDGET              │
│ CPU    0.5%              │
│ MEM     8 MB               │
├─────────────────────────────┤
│ > System ready             │
│ > Widget initialized       │
└─────────────────────────────┘
```

**Compact dimensions:** 280×180 pixels with optimized spacing and proper margins.

**Stats display:**
- CPU and MEM on separate lines with clean formatting
- Labels ("CPU", "MEM") and values properly aligned
- Clean format: "0.5%" and "8 MB" (no redundant percentage or "MEM" prefix)
- Properly positioned within widget bounds (no cut-off)
- 78px vertical gap between stats and content
- 10px horizontal gap after labels and values

## Usage

- **Drag widget**: Click and hold anywhere on the widget to move it
- **Exit**: Press `ESC` to quit

## Project Structure

```
status-widget/
├── main.go                 # Entry point
├── internal/
│   ├── widget/
│   │   └── widget.go       # Core widget logic
│   ├── theme/
│   │   └── matrix.go       # Matrix color palette
│   └── font/
│       └── font.go         # Font management
```

## Customization

To customize messages, modify the `messages` slice in `internal/widget/widget.go`:

```go
messages: []string{
    "> System ready",
    "> Widget initialized",
    // Add your custom messages here
},
```

To change colors, modify the color values in `internal/theme/matrix.go`.

## Requirements

- Go 1.26+
- Windows, macOS, or Linux

## License

MIT License - See LICENSE file for details.
