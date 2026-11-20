//go:build windows

package updater

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

// Install replaces the running executable with the downloaded one and starts
// it. Windows lets a running exe be renamed (not deleted), so the old file is
// moved aside to "<exe>.old" and removed by the new version on its next start.
// The caller must quit right after Install returns nil.
func Install(downloaded, newVersion string) error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	exe, _ = filepath.EvalSymlinks(exe)
	old := exe + ".old"
	_ = os.Remove(old)
	if err := os.Rename(exe, old); err != nil {
		return fmt.Errorf("cannot replace %s (is the folder read-only?): %w", filepath.Base(exe), err)
	}
	if err := copyFile(downloaded, exe); err != nil {
		_ = os.Remove(exe)
		_ = os.Rename(old, exe) // roll back
		return err
	}
	_ = os.Remove(downloaded)
	updateUninstallVersion(exe, newVersion)

	cmd := exec.Command(exe, "--updated", "--wait-pid", strconv.Itoa(os.Getpid()))
	cmd.Dir = filepath.Dir(exe)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: false}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("update installed but restart failed: %w", err)
	}
	return nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o755)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return err
	}
	return out.Close()
}

// updateUninstallVersion refreshes "Apps & features" when installed by the
// per-user installer.
func updateUninstallVersion(exe, version string) {
	const base = `Software\Microsoft\Windows\CurrentVersion\Uninstall`
	k, err := registry.OpenKey(registry.CURRENT_USER, base, registry.ENUMERATE_SUB_KEYS)
	if err != nil {
		return
	}
	names, _ := k.ReadSubKeyNames(-1)
	k.Close()
	dir := strings.ToLower(filepath.Dir(exe))
	for _, n := range names {
		sk, err := registry.OpenKey(registry.CURRENT_USER, base+`\`+n, registry.QUERY_VALUE|registry.SET_VALUE)
		if err != nil {
			continue
		}
		icon, _, _ := sk.GetStringValue("DisplayIcon")
		if icon != "" && strings.HasPrefix(strings.ToLower(strings.Trim(icon, `"`)), dir) {
			_ = sk.SetStringValue("DisplayVersion", version)
		}
		sk.Close()
	}
}

// Cleanup removes the previous exe left behind by an update.
func Cleanup() {
	exe, err := os.Executable()
	if err != nil {
		return
	}
	old := exe + ".old"
	for i := 0; i < 20; i++ {
		if err := os.Remove(old); err == nil || os.IsNotExist(err) {
			return
		}
		time.Sleep(500 * time.Millisecond)
	}
}

// WaitForExit blocks until process pid exits (or the timeout elapses). Used
// right after an update so the new copy does not race the old one. This is
// our own previous process, never a game.
func WaitForExit(pid int, timeout time.Duration) {
	h, err := windows.OpenProcess(windows.SYNCHRONIZE, false, uint32(pid))
	if err != nil {
		return
	}
	defer windows.CloseHandle(h)
	_, _ = windows.WaitForSingleObject(h, uint32(timeout/time.Millisecond))
}
