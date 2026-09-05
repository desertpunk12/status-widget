package widget

import (
	"fmt"
	"image"
	"math"
	"strconv"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// The Windows taskbar icon is a ring gauge: the filled arc and its color
// (green -> red) represent the available Z.ai quota percentage, with the
// number labeled inside on larger sizes. The tray icon instead shows only
// the plain number (no ring, no % sign): at tray scale a ring shrinks the
// digits below legibility, so the number fills the whole icon and its color
// (green -> red) carries the level instead.
//
// Sizes used for the window/taskbar icon; the tray renders its own icon at
// the middle size (good balance of detail and small-icon downscaling).
var taskbarIconSizes = []int{16, 32, 48}

// trayIconSize is the render size passed to the system tray updater.
const trayIconSize = 32

// taskbarIconNone marks "no icon rendered yet" so the first frame always
// generates one, even when availability is genuinely 0%.
const taskbarIconNone = -1

// updateTaskbarIcon refreshes the Windows taskbar icon and window title
// whenever the available quota percentage changes. Cheap to call every
// frame; regeneration only happens on change.
func (w *Widget) updateTaskbarIcon() {
	avail := w.availablePercent()
	if avail == w.taskbarIconPct {
		return
	}
	w.taskbarIconPct = avail

	icons := make([]image.Image, 0, len(taskbarIconSizes))
	for _, size := range taskbarIconSizes {
		icons = append(icons, w.renderTaskbarIcon(size, avail))
	}
	ebiten.SetWindowIcon(icons)

	// The taskbar shows the title next to the icon; keep the number visible
	// there too, since icon text is tiny at taskbar scale.
	ebiten.SetWindowTitle(fmt.Sprintf("Z.AI %d%%", avail))

	// Refresh the system tray (notification area) icon + tooltip. The tray
	// icon is number-only; the % sign lives in the tooltip.
	trayIcon := w.renderTrayIcon(trayIconSize, avail)
	if w.trayUpdater != nil && trayIcon != nil {
		w.trayUpdater(trayIcon, fmt.Sprintf("Z.AI: %d%% available", avail))
	}
}

// renderTrayIcon draws the availability as a plain number in the usage
// gradient color (green when plenty remains, red when low) on a fully
// transparent background. The digits are scaled to fill the icon so they
// stay legible at tray scale, regardless of digit count (0-100).
// Must be called from the game loop (Update), since offscreen images and
// ReadPixels require an initialized graphics context.
func (w *Widget) renderTrayIcon(size int, avail int) image.Image {
	label := strconv.Itoa(avail)
	face := w.fontManager.BoldFace(float64(size))
	if face == nil {
		return nil
	}

	// Pass 1: draw the number on a scratch image large enough to avoid
	// clipping, then find the actual ink bounds. Font metrics alone cannot
	// tell where the glyphs land; pixels do.
	tw, th := text.Measure(label, face, 0)
	scratchW := int(math.Ceil(tw)) + 2
	scratchH := int(math.Ceil(th)) + 2
	scratch := ebiten.NewImage(scratchW, scratchH)
	defer scratch.Deallocate()
	text.Draw(scratch, label, face, &text.DrawOptions{})
	pixels := make([]byte, 4*scratchW*scratchH)
	scratch.ReadPixels(pixels)
	l, t, r, b := inkBounds(pixels, scratchW, scratchH)
	if l >= r || t >= b {
		return nil // No visible glyphs (should not happen for digits)
	}

	// Pass 2: redraw scaled so the ink box fills the icon with a small
	// margin, and centered.
	inkW := float64(r - l)
	inkH := float64(b - t)
	scale := math.Min(float64(size)/(inkW*1.08), float64(size)/(inkH*1.08))
	marginX := (float64(size) - scale*inkW) / 2
	marginY := (float64(size) - scale*inkH) / 2

	img := ebiten.NewImage(size, size)
	defer img.Deallocate()
	op := &text.DrawOptions{}
	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate(marginX-scale*float64(l), marginY-scale*float64(t))
	op.ColorScale.ScaleWithColor(w.getUsageColorGradient())
	text.Draw(img, label, face, op)

	// Copy to a plain RGBA image; ReadPixels returns premultiplied alpha,
	// which matches image.RGBA's pixel format expected by the tray encoder.
	final := make([]byte, 4*size*size)
	img.ReadPixels(final)
	return &image.RGBA{
		Pix:    final,
		Stride: 4 * size,
		Rect:   image.Rect(0, 0, size, size),
	}
}

// inkBounds returns the bounding box (inclusive min, exclusive max) of
// non-transparent pixels in a premultiplied RGBA8 pixel buffer.
func inkBounds(pixels []byte, w, h int) (l, t, r, b int) {
	l, t = w, h
	r, b = 0, 0
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if pixels[(y*w+x)*4+3] != 0 {
				if x < l {
					l = x
				}
				if x >= r {
					r = x + 1
				}
				if y < t {
					t = y
				}
				if y >= b {
					b = y + 1
				}
			}
		}
	}
	return
}

// renderTaskbarIcon draws the availability ring gauge into a new image of
// the given pixel size. Must be called from the game loop (Update), since
// offscreen images and ReadPixels require an initialized graphics context.
func (w *Widget) renderTaskbarIcon(size int, avail int) image.Image {
	img := ebiten.NewImage(size, size)
	defer img.Deallocate()

	clr := w.getUsageColorGradient()
	cx := float32(size) / 2
	cy := float32(size) / 2
	stroke := float32(size) / 8
	if stroke < 2 {
		stroke = 2
	}
	radius := cx - stroke/2

	// Dim track ring for the unfilled portion of the gauge
	vector.StrokeCircle(img, cx, cy, radius, stroke, w.colors.DimText, true)

	// Filled arc proportional to availability, starting at 12 o'clock
	if avail > 0 {
		start := -math.Pi / 2
		end := start + (float64(avail)/100)*2*math.Pi
		var arc vector.Path
		arc.Arc(cx, cy, radius, float32(start), float32(end), vector.Clockwise)
		strokeOp := &vector.StrokeOptions{
			Width:   stroke,
			LineCap: vector.LineCapRound,
		}
		drawOp := &vector.DrawPathOptions{AntiAlias: true}
		drawOp.ColorScale.ScaleWithColor(clr)
		vector.StrokePath(img, &arc, strokeOp, drawOp)
	}

	// Percentage label inside larger icons (text is unreadable at 16px)
	if size >= 32 {
		label := fmt.Sprintf("%d", avail)
		face := w.fontManager.IconFace()
		if size >= 48 {
			face = w.fontManager.BodyFace()
		}
		if face != nil {
			labelWidth, labelHeight := text.Measure(label, face, 0)
			op := &text.DrawOptions{}
			op.GeoM.Translate(float64(cx)-labelWidth/2, float64(cy)-labelHeight/2)
			op.ColorScale.ScaleWithColor(w.colors.IconText)
			text.Draw(img, label, face, op)
		}
	}

	// Copy to a plain RGBA image; ReadPixels returns premultiplied alpha,
	// which matches image.RGBA's pixel format expected by SetWindowIcon.
	pixels := make([]byte, 4*size*size)
	img.ReadPixels(pixels)
	return &image.RGBA{
		Pix:    pixels,
		Stride: 4 * size,
		Rect:   image.Rect(0, 0, size, size),
	}
}
