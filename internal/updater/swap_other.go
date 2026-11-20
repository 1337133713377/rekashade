//go:build !windows

package updater

import (
	"errors"
	"time"
)

// Install is only implemented on Windows.
func Install(downloaded, newVersion string) error {
	return errors.New("self-update is only supported on Windows")
}

// Cleanup is a no-op outside Windows.
func Cleanup() {}

// WaitForExit is a no-op outside Windows.
func WaitForExit(pid int, timeout time.Duration) {}
