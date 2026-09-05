package font

import (
	"bytes"
	"image/color"
	"log"
	"os"
	"path/filepath"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"

	"status-widget/internal/theme"
)

// Manager handles font loading and provides text faces
type Manager struct {
	fontSource *text.GoTextFaceSource
	faces      map[float64]*text.GoTextFace // size -> face

	// boldSource backs BoldFace; nil falls back to the regular source.
	boldSource *text.GoTextFaceSource
	boldFaces  map[float64]*text.GoTextFace // size -> face
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
			boldFaces:  make(map[float64]*text.GoTextFace),
		}
	}

	// Create font source from TrueType data
	fontSource, err := text.NewGoTextFaceSource(bytes.NewReader(fontData))
	if err != nil {
		log.Printf("Failed to create font source: %v", err)
		return &Manager{
			fontSource: nil,
			faces:      make(map[float64]*text.GoTextFace),
			boldFaces:  make(map[float64]*text.GoTextFace),
		}
	}

	return &Manager{
		fontSource: fontSource,
		faces:      make(map[float64]*text.GoTextFace),

		// Bold is optional: BoldFace falls back to the regular source when
		// no bold system font is installed.
		boldSource: loadBoldFontSource(),
		boldFaces:  make(map[float64]*text.GoTextFace),
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

// loadBoldFontSource loads a bold font for small-size text like the tray
// number, where thin regular strokes become unreadable. Returns nil when no
// bold font is found; callers treat that as "use the regular font".
func loadBoldFontSource() *text.GoTextFaceSource {
	fontPaths := []string{
		// Windows 10/11: Segoe UI Bold
		filepath.Join(os.Getenv("SystemRoot"), "Fonts", "segoeuib.ttf"),
		// Windows: Arial Bold
		filepath.Join(os.Getenv("SystemRoot"), "Fonts", "arialbd.ttf"),
	}

	for _, fontPath := range fontPaths {
		if _, err := os.Stat(fontPath); err != nil {
			continue
		}
		fontData, err := os.ReadFile(fontPath)
		if err != nil {
			continue
		}
		source, err := text.NewGoTextFaceSource(bytes.NewReader(fontData))
		if err != nil {
			log.Printf("Failed to create bold font source %s: %v", fontPath, err)
			return nil
		}
		return source
	}
	return nil
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

// TitleFace returns the face for title text (18px)
func (m *Manager) TitleFace() *text.GoTextFace {
	return m.getFace(18)
}

// BodyFace returns the face for body text (16px)
func (m *Manager) BodyFace() *text.GoTextFace {
	return m.getFace(16)
}

// SmallFace returns the face for small text (14px)
func (m *Manager) SmallFace() *text.GoTextFace {
	return m.getFace(14)
}

// StatusFace returns the face for status bar text (11px)
func (m *Manager) StatusFace() *text.GoTextFace {
	return m.getFace(theme.FontSizeStatus)
}

// IconFace returns the face for icon labels (9px)
func (m *Manager) IconFace() *text.GoTextFace {
	return m.getFace(theme.FontSizeIcon)
}

// BoldFace returns a bold face of the specified size, used where regular
// strokes are too thin (e.g. the tray number). Falls back to the regular
// font when no bold system font is available.
func (m *Manager) BoldFace(size float64) *text.GoTextFace {
	if m.boldSource == nil {
		return m.getFace(size)
	}
	if face, ok := m.boldFaces[size]; ok {
		return face
	}
	face := &text.GoTextFace{
		Source: m.boldSource,
		Size:   size,
	}
	m.boldFaces[size] = face
	return face
}

// DrawColoredText draws text with the specified color at the given position
func DrawColoredText(screen *ebiten.Image, str string, face *text.GoTextFace, x, y int, clr color.Color) {
	if face == nil {
		// Skip rendering if no face available
		return
	}

	op := &text.DrawOptions{}
	op.GeoM.Translate(float64(x), float64(y))
	op.ColorScale.ScaleWithColor(clr)

	text.Draw(screen, str, face, op)
}
