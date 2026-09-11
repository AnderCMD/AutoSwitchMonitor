//go:build darwin

package ddc

import (
	"fmt"
	"os/exec"
	"strconv"
	"time"
)

// The DDC/CI channel (DisplayPort AUX, or I2C forwarded by a USB-C→HDMI
// hub/adapter) is prone to transient communication failures even on
// setups that generally work fine. Retrying a few times with a short
// pause avoids reporting an error for a single passing glitch.
const (
	setInputAttempts = 3
	setInputRetryGap = 300 * time.Millisecond
)

// macOS has no public Apple API for DDC/CI: DDC/CI itself on Apple
// Silicon relies on private frameworks (DisplayServices/AVService) that
// Apple can change without notice between macOS versions. Instead of
// reimplementing those private bindings here (fragile and high
// maintenance), we delegate to command-line tools already proven by the
// community:
//
//   - m1ddc (Apple Silicon, via USB-C/DisplayPort Alt Mode):
//     brew install m1ddc   (https://github.com/waydabber/m1ddc)
//   - ddcctl (Intel Mac):
//     brew install ddcctl  (https://github.com/kfix/ddcctl)
//
// setInput tries m1ddc first and falls back to ddcctl if it's not installed.
func setInput(vcpCode int) error {
	if path, err := exec.LookPath("m1ddc"); err == nil {
		return runWithRetry(path, "set", "input", strconv.Itoa(vcpCode))
	}

	if path, err := exec.LookPath("ddcctl"); err == nil {
		return runWithRetry(path, "-d", "1", "-i", strconv.Itoa(vcpCode))
	}

	return fmt.Errorf("neither m1ddc nor ddcctl was found in PATH; install one with Homebrew " +
		"(brew install waydabber/m1ddc/m1ddc, or brew install ddcctl) to control the monitor input")
}

func runWithRetry(path string, args ...string) error {
	var lastErr error
	var lastOut []byte
	for attempt := 1; attempt <= setInputAttempts; attempt++ {
		out, err := exec.Command(path, args...).CombinedOutput()
		if err == nil {
			return nil
		}
		lastErr, lastOut = err, out
		if attempt < setInputAttempts {
			time.Sleep(setInputRetryGap)
		}
	}
	return fmt.Errorf("%s failed after %d attempts: %w (%s)%s",
		path, setInputAttempts, lastErr, lastOut, adapterHint)
}

// adapterHint is appended to the error when DDC/CI fails: on Apple Silicon
// (without a physical HDMI port) the DDC channel only works if the
// USB-C→HDMI hub/cable forwards it, and many cheap single-port adapters
// don't, even though the video itself looks fine. See README.md § macOS.
const adapterHint = " — if this always fails (not just occasionally), your " +
	"USB-C→HDMI cable/adapter may not be forwarding the DDC/CI channel: try a " +
	"multiport USB-C hub instead of a simple adapter, or a real DisplayPort " +
	"connection if your monitor has one (see README.md)"
