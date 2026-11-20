//go:build windows

// Package host runs a dedicated Win32 UI thread that owns:
//
//   - a hidden "host" window receiving system broadcasts (display changes,
//     resume from sleep, Explorer restarts, installer close requests);
//   - the crosshair overlay: a layered, click-through, top-most window that
//     simply sits above the desktop like any other window (no hooking, no
//     injection — the game never knows it exists);
//   - the notification-area (tray) icon and its menu.
//
// Win32 windows belong to the thread that created them, so every window
// operation is marshalled onto this thread with Post/Call.
package host

import (
	"errors"
	"os"
	"runtime"
	"sync"
	"unsafe"

	"golang.org/x/sys/windows"

	"github.com/1337133713377/rekashade/internal/win"
)

const (
	HostClass    = "RekaShadeHost"
	overlayClass = "RekaShadeOverlay"

	wmTask = win.WM_APP + 1
	wmTray = win.WM_APP + 2

	timerTopmost = 1
	timerFollow  = 2
)

// MenuItem describes one entry of the tray menu.
type MenuItem struct {
	ID        int
	Label     string
	Checked   bool
	Radio     bool
	Disabled  bool
	Separator bool
	Children  []MenuItem
}

// Callbacks are invoked on the host thread; keep them short (post work to
// goroutines if needed).
type Callbacks struct {
	TrayClick     func()
	TrayMenu      func() []MenuItem
	MenuCommand   func(id int)
	DisplayChange func()
	Resume        func()
	CloseRequest  func()
	EndSession    func()
	FollowTick    func()
}

// Host is the native UI thread.
type Host struct {
	cb      Callbacks
	tasks   chan func()
	hwnd    uintptr
	tid     uint32
	ready   chan error
	stopped chan struct{}

	taskbarCreated uint32
	trayOn         bool
	trayTip        string
	traySmall      windows.Handle
	trayLarge      windows.Handle

	ov overlay
}

var (
	current   *Host
	wndProcCb uintptr
	cbOnce    sync.Once
)

// Start launches the host thread and waits until its windows exist.
func Start(cb Callbacks) (*Host, error) {
	h := &Host{cb: cb, tasks: make(chan func(), 4096), ready: make(chan error, 1), stopped: make(chan struct{})}
	current = h
	go h.run()
	if err := <-h.ready; err != nil {
		return nil, err
	}
	return h, nil
}

func (h *Host) run() {
	runtime.LockOSThread()
	defer close(h.stopped)
	h.tid = windows.GetCurrentThreadId()

	cbOnce.Do(func() { wndProcCb = windows.NewCallback(wndProc) })
	var inst windows.Handle
	_ = windows.GetModuleHandleEx(0, nil, &inst)

	for _, class := range []string{HostClass, overlayClass} {
		wc := win.WndClassEx{WndProc: wndProcCb, Instance: inst, ClassName: win.UTF16(class)}
		wc.Size = uint32(unsafe.Sizeof(wc))
		if err := win.RegisterClassEx(&wc); err != nil {
			h.ready <- err
			return
		}
	}

	hwnd, err := win.CreateWindowEx(win.WS_EX_TOOLWINDOW, HostClass, "Reka Shade", win.WS_POPUP, 0, 0, 0, 0, 0, 0, uintptr(inst))
	if err != nil {
		h.ready <- err
		return
	}
	h.hwnd = hwnd
	h.taskbarCreated = win.RegisterWindowMessage("TaskbarCreated")

	ovh, err := win.CreateWindowEx(win.WS_EX_LAYERED|win.WS_EX_TRANSPARENT|win.WS_EX_TOPMOST|win.WS_EX_TOOLWINDOW|win.WS_EX_NOACTIVATE,
		overlayClass, "Reka Shade Overlay", win.WS_POPUP, 0, 0, 1, 1, 0, 0, uintptr(inst))
	if err != nil {
		h.ready <- err
		return
	}
	h.ov.hwnd = ovh
	win.SetTimer(h.hwnd, timerTopmost, 1500)
	h.ready <- nil

	var msg win.Msg
	for {
		r := win.GetMessage(&msg)
		if r == 0 || r == -1 {
			break
		}
		win.TranslateMessage(&msg)
		win.DispatchMessage(&msg)
	}
}

// Post runs f on the host thread without waiting.
func (h *Host) Post(f func()) {
	if h == nil {
		return
	}
	select {
	case h.tasks <- f:
		win.PostMessage(h.hwnd, wmTask, 0, 0)
	default: // queue full: the thread is wedged; dropping beats deadlocking
	}
}

// Call runs f on the host thread and waits for it.
func (h *Host) Call(f func()) {
	if h == nil {
		return
	}
	if windows.GetCurrentThreadId() == h.tid {
		f()
		return
	}
	done := make(chan struct{})
	h.Post(func() { defer close(done); f() })
	<-done
}

// Stop removes the tray icon, destroys the windows and ends the thread.
func (h *Host) Stop() {
	if h == nil {
		return
	}
	h.Call(func() {
		h.removeTray()
		win.KillTimer(h.hwnd, timerTopmost)
		win.KillTimer(h.hwnd, timerFollow)
		if h.ov.hwnd != 0 {
			win.DestroyWindow(h.ov.hwnd)
			h.ov.hwnd = 0
		}
		win.DestroyWindow(h.hwnd)
		win.PostQuitMessage(0)
	})
	<-h.stopped
}

func (h *Host) drainTasks() {
	for {
		select {
		case f := <-h.tasks:
			func() {
				defer func() { _ = recover() }()
				f()
			}()
		default:
			return
		}
	}
}

func wndProc(hwnd uintptr, msg uint32, wParam, lParam uintptr) uintptr {
	h := current
	if h == nil {
		return win.DefWindowProc(hwnd, msg, wParam, lParam)
	}
	if hwnd == h.ov.hwnd && hwnd != 0 {
		if msg == win.WM_NCHITTEST {
			return win.HTTRANSPARENT
		}
		return win.DefWindowProc(hwnd, msg, wParam, lParam)
	}
	switch msg {
	case wmTask:
		h.drainTasks()
		return 0
	case wmTray:
		switch uint32(lParam & 0xFFFF) {
		case win.WM_LBUTTONUP, win.WM_LBUTTONDBLCLK:
			if h.cb.TrayClick != nil {
				h.cb.TrayClick()
			}
		case win.WM_RBUTTONUP, win.WM_CONTEXTMENU:
			h.showMenu()
		}
		return 0
	case win.WM_TIMER:
		switch wParam {
		case timerTopmost:
			h.ov.reassertTopmost()
		case timerFollow:
			if h.cb.FollowTick != nil {
				h.cb.FollowTick()
			}
		}
		return 0
	case win.WM_DISPLAYCHANGE:
		if h.cb.DisplayChange != nil {
			go h.cb.DisplayChange()
		}
	case win.WM_POWERBROADCAST:
		if wParam == win.PBT_APMRESUMEAUTOMATIC || wParam == win.PBT_APMRESUMESUSPEND {
			if h.cb.Resume != nil {
				go h.cb.Resume()
			}
		}
		return 1
	case win.WM_CLOSE:
		// Sent by the installer/uninstaller to ask a running copy to exit.
		if h.cb.CloseRequest != nil {
			go h.cb.CloseRequest()
		}
		return 0
	case win.WM_QUERYENDSESSION:
		return 1
	case win.WM_ENDSESSION:
		if wParam != 0 && h.cb.EndSession != nil {
			h.cb.EndSession()
		}
		return 0
	default:
		if h.taskbarCreated != 0 && msg == h.taskbarCreated && h.trayOn {
			h.trayOn = false
			h.addTray()
			return 0
		}
	}
	return win.DefWindowProc(hwnd, msg, wParam, lParam)
}

// ---------------------------------------------------------------------------
// Tray
// ---------------------------------------------------------------------------

// ShowTray adds (or updates) the notification-area icon.
func (h *Host) ShowTray(tip string) {
	h.Post(func() {
		h.trayTip = tip
		if h.trayOn {
			nid := h.nid()
			nid.Flags = win.NIF_TIP | win.NIF_SHOWTIP
			win.ShellNotifyIcon(win.NIM_MODIFY, &nid)
			return
		}
		h.addTray()
	})
}

// Notify shows a balloon/toast from the tray icon.
func (h *Host) Notify(title, text string) {
	h.Post(func() {
		if !h.trayOn {
			return
		}
		nid := h.nid()
		nid.Flags = win.NIF_INFO
		copy(nid.InfoTitle[:63], windows.StringToUTF16(title))
		copy(nid.Info[:255], windows.StringToUTF16(text))
		nid.InfoFlags = win.NIIF_INFO | win.NIIF_NOSOUND
		win.ShellNotifyIcon(win.NIM_MODIFY, &nid)
	})
}

func (h *Host) nid() win.NotifyIconData {
	var nid win.NotifyIconData
	nid.CbSize = uint32(unsafe.Sizeof(nid))
	nid.Wnd = windows.HWND(h.hwnd)
	nid.ID = 1
	return nid
}

func (h *Host) addTray() {
	if h.traySmall == 0 {
		exe, _ := os.Executable()
		h.traySmall, h.trayLarge = win.ExeIcon(exe)
	}
	nid := h.nid()
	nid.Flags = win.NIF_MESSAGE | win.NIF_ICON | win.NIF_TIP | win.NIF_SHOWTIP
	nid.CallbackMessage = wmTray
	nid.Icon = h.traySmall
	copy(nid.Tip[:127], windows.StringToUTF16(h.trayTip))
	h.trayOn = win.ShellNotifyIcon(win.NIM_ADD, &nid)
}

func (h *Host) removeTray() {
	if !h.trayOn {
		return
	}
	nid := h.nid()
	win.ShellNotifyIcon(win.NIM_DELETE, &nid)
	h.trayOn = false
}

func buildMenu(items []MenuItem) uintptr {
	m := win.CreatePopupMenu()
	for _, it := range items {
		switch {
		case it.Separator:
			win.AppendMenu(m, win.MF_SEPARATOR, 0, "")
		case len(it.Children) > 0:
			sub := buildMenu(it.Children)
			win.AppendMenu(m, win.MF_POPUP, sub, it.Label)
		default:
			flags := uint32(win.MF_STRING)
			if it.Checked {
				flags |= win.MF_CHECKED
			}
			if it.Radio {
				flags |= win.MFT_RADIOCHECK
			}
			if it.Disabled {
				flags |= win.MF_GRAYED
			}
			win.AppendMenu(m, flags, uintptr(it.ID), it.Label)
		}
	}
	return m
}

func (h *Host) showMenu() {
	if h.cb.TrayMenu == nil {
		return
	}
	menu := buildMenu(h.cb.TrayMenu())
	defer win.DestroyMenu(menu) // also destroys submenus
	win.SetForegroundWindow(h.hwnd)
	pt := win.CursorPos()
	cmd := win.TrackPopupMenu(menu, win.TPM_RETURNCMD|win.TPM_RIGHTBUTTON|win.TPM_NONOTIFY, pt.X, pt.Y, h.hwnd)
	win.PostMessage(h.hwnd, win.WM_NULL, 0, 0)
	if cmd != 0 && h.cb.MenuCommand != nil {
		go h.cb.MenuCommand(cmd)
	}
}

// ---------------------------------------------------------------------------
// Overlay
// ---------------------------------------------------------------------------

type overlay struct {
	hwnd     uintptr
	bgra     []byte
	w, h     int32
	ax, ay   int32 // anchor (the pixel that must sit on the target point)
	cx, cy   int32 // target point on screen
	visible  bool
	uploaded bool
	excluded bool
}

// ErrNoImage is returned when the overlay has nothing to show yet.
var ErrNoImage = errors.New("no overlay image")

func (o *overlay) push() error {
	if o.hwnd == 0 || len(o.bgra) == 0 {
		return ErrNoImage
	}
	err := win.UpdateLayeredBitmap(o.hwnd, o.cx-o.ax, o.cy-o.ay, o.w, o.h, o.bgra)
	o.uploaded = err == nil
	return err
}

func (o *overlay) reassertTopmost() {
	if o.visible && o.hwnd != 0 {
		win.SetWindowPos(o.hwnd, win.HWND_TOPMOST, 0, 0, 0, 0, win.SWP_NOMOVE|win.SWP_NOSIZE|win.SWP_NOACTIVATE|win.SWP_NOOWNERZORDER|win.SWP_NOSENDCHANGING)
	}
}

// SetOverlayImage replaces the overlay bitmap (premultiplied BGRA, w*h*4).
func (h *Host) SetOverlayImage(bgra []byte, w, hgt, anchorX, anchorY int) {
	h.Post(func() {
		h.ov.bgra, h.ov.w, h.ov.h = bgra, int32(w), int32(hgt)
		h.ov.ax, h.ov.ay = int32(anchorX), int32(anchorY)
		if h.ov.visible {
			_ = h.ov.push()
		} else {
			h.ov.uploaded = false
		}
	})
}

// SetOverlayTarget moves the overlay so its anchor sits on (x, y).
func (h *Host) SetOverlayTarget(x, y int) {
	h.Post(func() {
		if h.ov.cx == int32(x) && h.ov.cy == int32(y) && h.ov.uploaded {
			return
		}
		h.ov.cx, h.ov.cy = int32(x), int32(y)
		if h.ov.visible && h.ov.uploaded {
			win.SetWindowPos(h.ov.hwnd, win.HWND_TOPMOST, h.ov.cx-h.ov.ax, h.ov.cy-h.ov.ay, 0, 0,
				win.SWP_NOSIZE|win.SWP_NOACTIVATE|win.SWP_NOOWNERZORDER|win.SWP_NOSENDCHANGING)
		} else if h.ov.visible {
			_ = h.ov.push()
		}
	})
}

// SetOverlayVisible shows or hides the overlay.
func (h *Host) SetOverlayVisible(v bool) {
	h.Post(func() {
		if v == h.ov.visible {
			return
		}
		h.ov.visible = v
		if !v {
			win.ShowWindow(h.ov.hwnd, win.SW_HIDE)
			return
		}
		if !h.ov.uploaded && h.ov.push() != nil {
			h.ov.visible = false
			return
		}
		win.ShowWindow(h.ov.hwnd, win.SW_SHOWNOACTIVATE)
		h.ov.reassertTopmost()
	})
}

// SetOverlayCaptureExcluded hides the overlay from screenshots and streams.
func (h *Host) SetOverlayCaptureExcluded(ex bool) {
	h.Post(func() {
		if ex == h.ov.excluded {
			return
		}
		aff := uint32(win.WDA_NONE)
		if ex {
			aff = win.WDA_EXCLUDEFROMCAPTURE
		}
		if win.SetWindowDisplayAffinity(h.ov.hwnd, aff) {
			h.ov.excluded = ex
		}
	})
}

// SetFollowTimer enables the fast timer used to track the active window.
func (h *Host) SetFollowTimer(on bool) {
	h.Post(func() {
		if on {
			win.SetTimer(h.hwnd, timerFollow, 50)
		} else {
			win.KillTimer(h.hwnd, timerFollow)
		}
	})
}

// OverlayHWND returns the overlay window handle (for "is this us?" checks).
func (h *Host) OverlayHWND() uintptr { return h.ov.hwnd }

// HostHWND returns the hidden host window handle.
func (h *Host) HostHWND() uintptr { return h.hwnd }
