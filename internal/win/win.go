//go:build windows

// Package win provides minimal Win32 control over the ebitengine (GLFW)
// game window: a true hide/show that also removes the taskbar button,
// Discord-style, instead of just minimizing or moving the window off-screen.
package win

import (
	"fmt"
	"sync"
	"unsafe"

	"golang.org/x/sys/windows"
)

// glfwClassName is the window class GLFW registers for game windows on
// Windows; used to pick our game window among the process's windows
// (the systray helper window uses a different class).
const glfwClassName = "GLFW30"

const (
	swHide = 0 // SW_HIDE: hide window and remove its taskbar button
	swShow = 5 // SW_SHOW: show at previous position and activate
)

var (
	user32                    = windows.NewLazySystemDLL("user32.dll")
	pEnumWindows              = user32.NewProc("EnumWindows")
	pGetClassNameW            = user32.NewProc("GetClassNameW")
	pGetWindowThreadProcessId = user32.NewProc("GetWindowThreadProcessId")
	pShowWindowAsync          = user32.NewProc("ShowWindowAsync")
)

var (
	mu   sync.Mutex
	hwnd uintptr // cached game window handle

	// Enumeration state, only accessed while mu is held (EnumWindows runs
	// its callback synchronously on the calling thread)
	enumPid   uint32
	enumFound uintptr
)

// GameWindow returns the HWND of the GLFW game window belonging to the
// current process. The result is cached after the first successful lookup;
// lookups also succeed while the window is hidden.
func GameWindow() (uintptr, error) {
	mu.Lock()
	defer mu.Unlock()

	if hwnd != 0 {
		return hwnd, nil
	}

	enumPid = uint32(windows.GetCurrentProcessId())
	enumFound = 0
	pEnumWindows.Call(windows.NewCallback(enumProc), 0)
	if enumFound == 0 {
		return 0, fmt.Errorf("GLFW game window not found for current process")
	}
	hwnd = enumFound
	return hwnd, nil
}

// enumProc matches top-level windows of our PID with the GLFW class name.
func enumProc(h uintptr, _ uintptr) uintptr {
	var wpid uint32
	pGetWindowThreadProcessId.Call(h, uintptr(unsafe.Pointer(&wpid)))
	if wpid != enumPid {
		return 1 // Different process; keep enumerating
	}

	var buf [64]uint16
	n, _, _ := pGetClassNameW.Call(h, uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
	if n > 0 && windows.UTF16ToString(buf[:n]) == glfwClassName {
		enumFound = h
		return 0 // Found; stop enumerating
	}
	return 1
}

// Hide hides the window and removes its taskbar button (SW_HIDE).
//
// ShowWindowAsync is used instead of ShowWindow on purpose: in Ebitengine's
// multithreaded mode, Update/Draw run on the game thread while the GLFW
// window belongs to the UI thread, which only pumps messages between
// frames. ShowWindow on an active window performs a synchronous SendMessage
// to the owning thread; called from the game thread this deadlocks both
// threads until Windows kills the app as "not responding" (AppHangB1).
// ShowWindowAsync posts the request and returns immediately; the window is
// hidden when the UI thread pumps messages on the next frame.
func Hide(h uintptr) {
	pShowWindowAsync.Call(h, swHide)
}

// Show reveals a hidden window at its previous position and activates it
// (SW_SHOW). Like Hide, this uses ShowWindowAsync to avoid a cross-thread
// SendMessage deadlock with the UI thread (activation of the window also
// involves synchronous messages).
func Show(h uintptr) {
	pShowWindowAsync.Call(h, swShow)
}
