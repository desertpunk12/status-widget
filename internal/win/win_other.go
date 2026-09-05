//go:build !windows

// Package win is a no-op on non-Windows platforms; the widget falls back to
// parking the window off-screen when hiding.
package win

import "errors"

var errUnsupported = errors.New("window control unsupported on this platform")

// GameWindow always fails on non-Windows platforms.
func GameWindow() (uintptr, error) {
	return 0, errUnsupported
}

// Hide is a no-op on non-Windows platforms.
func Hide(h uintptr) {}

// Show is a no-op on non-Windows platforms.
func Show(h uintptr) {}
