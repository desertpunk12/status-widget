package font

import (
	"bytes"
	"image/color"
	"log"
	"os"
	"path/filepath"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

// Manager handles font loading and provides text faces
type Manager struct {
	fontSource *text.GoTextFaceSource
	faces      map[float64]*text.GoTextFace // size -> face
}

// NewManager creates a new font manager and loads emoji-compatible fonts
func NewManager() *Manager {
	// Try to load system emoji font (Segoe UI Emoji on Windows)
	fontData, err := loadSystemEmojiFont()
	if err != nil {
		log.Printf("Failed to load emoji font, using fallback: %v", err)
		// Use built-in text rendering as fallback
		return &Manager{
			fontSource: nil,
			faces:      make(map[float64]*text.GoTextFace),
		}
	}

	// Create font source from TrueType data
	fontSource, err := text.NewGoTextFaceSource(bytes.NewReader(fontData))
	if err != nil {
		log.Printf("Failed to create font source: %v", err)
		return &Manager{
			fontSource: nil,
			faces:      make(map[float64]*text.GoTextFace),
		}
	}

	return &Manager{
		fontSource: fontSource,
		faces:      make(map[float64]*text.GoTextFace),
	}
}

// loadSystemEmojiFont attempts to load Windows emoji-compatible fonts
func loadSystemEmojiFont() ([]byte, error) {
	// Try common Windows emoji fonts in order of preference
	fontPaths := []string{
		// Windows 11: Segoe UI Emoji
		filepath.Join(os.Getenv("SystemRoot"), "Fonts", "segoeuiemj.ttf"),
		// Windows 10/11: Segoe UI Symbol
		filepath.Join(os.Getenv("SystemRoot"), "Fonts", "seguisym.ttf"),
		// Windows: Arial Unicode MS
		filepath.Join(os.Getenv("SystemRoot"), "Fonts", "arialuni.ttf"),
		// Windows: Arial (standard font)
		filepath.Join(os.Getenv("SystemRoot"), "Fonts", "arial.ttf"),
		// Windows: Segoe UI (standard Windows 10+ font)
		filepath.Join(os.Getenv("SystemRoot"), "Fonts", "segoeui.ttf"),
	}

	for _, fontPath := range fontPaths {
		if _, err := os.Stat(fontPath); err == nil {
			log.Printf("Loading font: %s", fontPath)
			return os.ReadFile(fontPath)
		}
	}

	return nil, os.ErrNotExist
}

// getFace returns or creates a face of the specified size
func (m *Manager) getFace(size float64) *text.GoTextFace {
	if m.fontSource == nil {
		return nil
	}

	if face, ok := m.faces[size]; ok {
		return face
	}

	face := &text.GoTextFace{
		Source: m.fontSource,
		Size:   size,
	}
	m.faces[size] = face
	return face
}

// TitleFace returns the face for title text (14px)
func (m *Manager) TitleFace() *text.GoTextFace {
	return m.getFace(14)
}

// BodyFace returns the face for body text (13px)
func (m *Manager) BodyFace() *text.GoTextFace {
	return m.getFace(13)
}

// SmallFace returns the face for small text (11px)
func (m *Manager) SmallFace() *text.GoTextFace {
	return m.getFace(11)
}

// DrawColoredText draws text with the specified color at the given position
// Emojis are rendered in white to preserve their original colors
func DrawColoredText(screen *ebiten.Image, str string, face *text.GoTextFace, x, y int, clr color.Color) {
	if face == nil {
		// Skip rendering if no face available
		return
	}

	// Check if string contains emoji - emoji need white color to preserve original colors
	drawColor := clr
	for _, r := range str {
		// Emoji ranges: Miscellaneous Symbols and Pictographs (0x1F300+)
		// Also includes the specific emoji we use
		if r >= 0x1F300 || (r >= 0x1F600 && r <= 0x1F64F) || r == '⏳' {
			drawColor = color.White
			break
		}
	}

	op := &text.DrawOptions{}
	op.GeoM.Translate(float64(x), float64(y))
	op.ColorScale.ScaleWithColor(drawColor)

	text.Draw(screen, str, face, op)
}
