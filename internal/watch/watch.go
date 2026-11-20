//go:build windows

// Package watch answers "which game is running / in front?" without ever
// touching a game: it reads the system process list (the same list Task
// Manager shows) and the foreground window's metadata. No process handle is
// opened, no memory is read, no hook is installed.
package watch

import (
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"

	"github.com/1337133713377/rekashade/internal/win"
)

// Processes returns pid → lower-case exe name for all running processes.
func Processes() map[uint32]string {
	out := map[uint32]string{}
	snap, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return out
	}
	defer windows.CloseHandle(snap)
	var pe windows.ProcessEntry32
	pe.Size = uint32(unsafe.Sizeof(pe))
	if err := windows.Process32First(snap, &pe); err != nil {
		return out
	}
	for {
		out[pe.ProcessID] = strings.ToLower(windows.UTF16ToString(pe.ExeFile[:]))
		if err := windows.Process32Next(snap, &pe); err != nil {
			break
		}
	}
	return out
}

// Foreground describes the window the user is looking at.
type Foreground struct {
	HWND       uintptr
	PID        uint32
	Exe        string
	Fullscreen bool   // covers its whole monitor (borderless/fullscreen game)
	Monitor    string // GDI device of the monitor it is on
	Client     win.RECT
	HasClient  bool
}

var shellClasses = map[string]bool{
	"progman": true, "workerw": true, "shell_traywnd": true, "shell_secondarytraywnd": true,
	"windows.ui.core.corewindow": true, "multitaskingviewframe": true, "xamlexplorerhostislandwindow": true,
}

// GetForeground inspects the foreground window. procs may be nil.
func GetForeground(procs map[uint32]string, self uint32) Foreground {
	var f Foreground
	hwnd := win.GetForegroundWindow()
	if hwnd == 0 {
		return f
	}
	f.HWND = hwnd
	f.PID = win.WindowPID(hwnd)
	if procs != nil {
		f.Exe = procs[f.PID]
	}
	if f.PID == self || win.IsIconic(hwnd) || hwnd == win.ShellWindow() || hwnd == win.DesktopWindow() {
		return f
	}
	if shellClasses[strings.ToLower(win.ClassName(hwnd))] {
		return f
	}
	f.Client, f.HasClient = win.ClientRectOnScreen(hwnd)
	dev, mon, ok := win.MonitorDeviceOfWindow(hwnd)
	if !ok {
		return f
	}
	f.Monitor = dev
	if wr, ok := win.GetWindowRect(hwnd); ok {
		f.Fullscreen = wr.Left <= mon.Left && wr.Top <= mon.Top && wr.Right >= mon.Right && wr.Bottom >= mon.Bottom
	}
	return f
}

// VisibleApps lists exe names that currently own a visible top-level window
// (used to offer "pick a running game" in the UI).
func VisibleApps(self uint32) []string {
	procs := Processes()
	seen := map[string]bool{}
	var out []string
	win.EnumWindows(func(hwnd uintptr) bool {
		if !win.IsWindowVisible(hwnd) {
			return true
		}
		r, ok := win.GetWindowRect(hwnd)
		if !ok || r.Right-r.Left < 200 || r.Bottom-r.Top < 150 {
			return true
		}
		pid := win.WindowPID(hwnd)
		exe := procs[pid]
		if pid == self || exe == "" || seen[exe] || exe == "explorer.exe" || exe == "applicationframehost.exe" ||
			exe == "textinputhost.exe" || exe == "searchhost.exe" || exe == "startmenuexperiencehost.exe" || exe == "shellexperiencehost.exe" {
			return true
		}
		seen[exe] = true
		out = append(out, exe)
		return true
	})
	return out
}
