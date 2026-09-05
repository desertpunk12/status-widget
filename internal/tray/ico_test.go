package tray

import (
	"image"
	"image/color"
	"runtime"
	"testing"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	user32                  = windows.NewLazySystemDLL("user32.dll")
	procCreateIconFromResEx = user32.NewProc("CreateIconFromResourceEx")
	procDestroyIcon         = user32.NewProc("DestroyIcon")
)

// TestEncodeICOHeader verifies the binary layout of the generated ICO file.
func TestEncodeICOHeader(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 32, 32))
	img.Set(0, 0, color.RGBA{0, 255, 65, 255}) // Matrix green
	data, err := EncodeICO(img)
	if err != nil {
		t.Fatalf("EncodeICO failed: %v", err)
	}

	if len(data) < 22 {
		t.Fatalf("ICO too short: %d bytes", len(data))
	}
	if data[0] != 0 || data[1] != 0 { // Reserved
		t.Errorf("reserved bytes not zero: %x %x", data[0], data[1])
	}
	if data[2] != 1 || data[3] != 0 { // Type = icon
		t.Errorf("type != icon: %x %x", data[2], data[3])
	}
	if data[4] != 1 || data[5] != 0 { // Count = 1
		t.Errorf("count != 1: %x %x", data[4], data[5])
	}
	if data[6] != 32 || data[7] != 32 { // Entry width/height
		t.Errorf("entry size != 32x32: %d %d", data[6], data[7])
	}
}

// TestEncodeICOLoadableByWindows asks the OS to decode the generated ICO,
// proving the byte format is a valid icon resource.
func TestEncodeICOLoadableByWindows(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows-only test")
	}

	img := image.NewRGBA(image.Rect(0, 0, 32, 32))
	for y := 0; y < 32; y++ {
		for x := 0; x < 32; x++ {
			img.Set(x, y, color.RGBA{uint8(x * 8), uint8(y * 8), 65, 255})
		}
	}
	data, err := EncodeICO(img)
	if err != nil {
		t.Fatalf("EncodeICO failed: %v", err)
	}

	// CreateIconFromResourceEx expects the resource payload (bitmap header +
	// masks), i.e. the .ico file without its 22-byte directory header
	headerSize := 6 + 16
	if len(data) <= headerSize {
		t.Fatalf("ICO payload missing: %d bytes", len(data))
	}
	resource := data[headerSize:]

	hIcon, _, _ := procCreateIconFromResEx.Call(
		uintptr(unsafe.Pointer(&resource[0])),
		uintptr(len(resource)),
		1,          // fIcon = TRUE
		0x00030000, // icon resource version
		0, 0,       // desired size 0 = use native size
		0, // LR_DEFAULTCOLOR
	)
	if hIcon == 0 {
		t.Fatal("CreateIconFromResourceEx returned NULL handle")
	}
	procDestroyIcon.Call(hIcon)
}
