package tray

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
)

// ICO file structures (little-endian). Field names must be exported for
// encoding/binary, which skips unexported fields.
type iconDir struct {
	Reserved uint16 // Must be 0
	Type     uint16 // 1 = icon
	Count    uint16 // Number of images
}

type iconDirEntry struct {
	Width      uint8  // 0 means 256
	Height     uint8  // 0 means 256
	ColorCount uint8  // 0 for 32bpp
	Reserved   uint8  // Must be 0
	Planes     uint16 // Must be 1
	BitCount   uint16 // Bits per pixel
	BytesInRes uint32 // Size of the image data
	Offset     uint32 // Offset from file start
}

type bitmapInfoHeader struct {
	Size          uint32
	Width         int32
	Height        int32 // 2x icon height: XOR mask + AND mask
	Planes        uint16
	BitCount      uint16
	Compression   uint32 // BI_RGB = 0
	SizeImage     uint32
	XPelsPerMeter int32
	YPelsPerMeter int32
	ClrUsed       uint32
	ClrImportant  uint32
}

// EncodeICO converts img into single-image .ico file bytes using an
// uncompressed 32bpp BMP entry with straight alpha, which Windows tray
// icons (CreateIconFromResourceEx) always accept.
func EncodeICO(img image.Image) ([]byte, error) {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	if w == 0 || h == 0 || w > 256 || h > 256 {
		return nil, fmt.Errorf("unsupported icon size %dx%d", w, h)
	}

	// XOR mask: bottom-up rows of BGRA pixels. ICO wants straight
	// (non-premultiplied) alpha; image.RGBA from ReadPixels is premultiplied,
	// so convert through the NRGBA color model.
	rowLen := 4 * w
	xor := make([]byte, rowLen*h)
	for y := 0; y < h; y++ {
		srcY := h - 1 - y // BMP rows are stored bottom-up
		for x := 0; x < w; x++ {
			c := color.NRGBAModel.Convert(img.At(b.Min.X+x, b.Min.Y+srcY)).(color.NRGBA)
			i := y*rowLen + 4*x
			xor[i], xor[i+1], xor[i+2], xor[i+3] = c.B, c.G, c.R, c.A
		}
	}

	// AND mask: 1bpp rows padded to 32-bit boundaries. All zeros: opacity is
	// carried by the alpha channel, so the AND mask never forces transparency.
	andRow := ((w + 31) / 32) * 4
	and := make([]byte, andRow*h)

	bmp := &bitmapInfoHeader{
		Size:     uint32(binary.Size(bitmapInfoHeader{})),
		Width:    int32(w),
		Height:   int32(2 * h),
		Planes:   1,
		BitCount: 32,
	}
	imageSize := uint32(binary.Size(bmp)) + uint32(len(xor)) + uint32(len(and))

	var buf bytes.Buffer
	dir := iconDir{Type: 1, Count: 1}
	dataOffset := uint32(binary.Size(dir)) + uint32(binary.Size(iconDirEntry{}))
	entry := iconDirEntry{
		Width:      uint8(w),
		Height:     uint8(h),
		Planes:     1,
		BitCount:   32,
		BytesInRes: imageSize,
		Offset:     dataOffset,
	}
	for _, v := range []interface{}{&dir, &entry, bmp} {
		if err := binary.Write(&buf, binary.LittleEndian, v); err != nil {
			return nil, fmt.Errorf("failed to encode ICO header: %w", err)
		}
	}
	buf.Write(xor)
	buf.Write(and)

	return buf.Bytes(), nil
}
