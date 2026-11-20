//go:build windows

// Package startup manages the per-user "start with Windows" entry.
package startup

import (
	"errors"
	"os"
	"path/filepath"

	"golang.org/x/sys/windows/registry"
)

const runKey = `Software\Microsoft\Windows\CurrentVersion\Run`
const valueName = "RekaShade"

// Set adds or removes Reka Shade from the user's startup programs.
func Set(enabled, minimized bool) error {
	k, _, err := registry.CreateKey(registry.CURRENT_USER, runKey, registry.SET_VALUE|registry.QUERY_VALUE)
	if err != nil {
		return err
	}
	defer k.Close()
	if !enabled {
		err := k.DeleteValue(valueName)
		if errors.Is(err, registry.ErrNotExist) {
			return nil
		}
		return err
	}
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	exe, _ = filepath.EvalSymlinks(exe)
	cmd := `"` + exe + `"`
	if minimized {
		cmd += " --minimized"
	}
	if cur, _, err := k.GetStringValue(valueName); err == nil && cur == cmd {
		return nil
	}
	return k.SetStringValue(valueName, cmd)
}

// Enabled reports whether the startup entry exists.
func Enabled() bool {
	k, err := registry.OpenKey(registry.CURRENT_USER, runKey, registry.QUERY_VALUE)
	if err != nil {
		return false
	}
	defer k.Close()
	_, _, err = k.GetStringValue(valueName)
	return err == nil
}
