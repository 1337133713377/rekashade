//go:build windows

package win

import (
	"errors"
	"fmt"
	"sync"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

// ---------------------------------------------------------------------------
// Displays
// ---------------------------------------------------------------------------

// EnumDisplayDevices wraps EnumDisplayDevicesW. device may be "" for adapters.
func EnumDisplayDevices(device string, index uint32, flags uint32) (DisplayDevice, bool) {
	var dd DisplayDevice
	dd.Cb = uint32(unsafe.Sizeof(dd))
	r, _, _ := pEnumDisplayDevicesW.Call(uintptr(unsafe.Pointer(UTF16(device))), uintptr(index), uintptr(unsafe.Pointer(&dd)), uintptr(flags))
	return dd, r != 0
}

// EnumDisplaySettings wraps EnumDisplaySettingsExW.
func EnumDisplaySettings(device string, mode uint32) (DevMode, bool) {
	var dm DevMode
	dm.Size = uint16(unsafe.Sizeof(dm))
	r, _, _ := pEnumDisplaySettingsExW.Call(uintptr(unsafe.Pointer(UTF16(device))), uintptr(mode), uintptr(unsafe.Pointer(&dm)), 0)
	return dm, r != 0
}

// ChangeDisplaySettings wraps ChangeDisplaySettingsExW. A nil mode restores
// the mode stored in the registry for that display.
func ChangeDisplaySettings(device string, dm *DevMode, flags uint32) int32 {
	r, _, _ := pChangeDisplaySettingsExW.Call(uintptr(unsafe.Pointer(UTF16(device))), uintptr(unsafe.Pointer(dm)), 0, uintptr(flags), 0)
	return int32(r)
}

// DispChangeError converts a DISP_CHANGE_* code to an error (nil on success).
func DispChangeError(code int32) error {
	switch code {
	case DISP_CHANGE_SUCCESSFUL:
		return nil
	case DISP_CHANGE_RESTART:
		return errors.New("Windows needs a restart to use this mode")
	case DISP_CHANGE_BADMODE:
		return errors.New("this mode is not supported by the display or driver")
	case DISP_CHANGE_FAILED:
		return errors.New("the display driver refused this mode")
	case DISP_CHANGE_NOTUPDATED:
		return errors.New("Windows could not save the display settings")
	case DISP_CHANGE_BADFLAGS, DISP_CHANGE_BADPARAM:
		return errors.New("invalid display settings")
	case DISP_CHANGE_BADDUALVIEW:
		return errors.New("this mode is not allowed on a DualView system")
	}
	return fmt.Errorf("display change failed (code %d)", code)
}

// QueryActiveDisplayConfig returns the active display paths and modes.
func QueryActiveDisplayConfig() ([]PathInfo, []ModeInfo, error) {
	for attempt := 0; attempt < 3; attempt++ {
		var np, nm uint32
		r, _, _ := pGetDisplayConfigBufferSizes.Call(QDC_ONLY_ACTIVE_PATHS, uintptr(unsafe.Pointer(&np)), uintptr(unsafe.Pointer(&nm)))
		if r != 0 {
			return nil, nil, syscall.Errno(r)
		}
		if np == 0 {
			return nil, nil, nil
		}
		paths := make([]PathInfo, np)
		modes := make([]ModeInfo, nm+1)
		r, _, _ = pQueryDisplayConfig.Call(QDC_ONLY_ACTIVE_PATHS,
			uintptr(unsafe.Pointer(&np)), uintptr(unsafe.Pointer(&paths[0])),
			uintptr(unsafe.Pointer(&nm)), uintptr(unsafe.Pointer(&modes[0])), 0)
		if r == uintptr(windows.ERROR_INSUFFICIENT_BUFFER) {
			continue // topology changed between the two calls
		}
		if r != 0 {
			return nil, nil, syscall.Errno(r)
		}
		return paths[:np], modes[:nm], nil
	}
	return nil, nil, errors.New("display configuration keeps changing")
}

// SetDisplayConfig wraps SetDisplayConfig with caller-supplied paths/modes.
func SetDisplayConfig(paths []PathInfo, modes []ModeInfo, flags uint32) error {
	var pp, pm uintptr
	if len(paths) > 0 {
		pp = uintptr(unsafe.Pointer(&paths[0]))
	}
	if len(modes) > 0 {
		pm = uintptr(unsafe.Pointer(&modes[0]))
	}
	r, _, _ := pSetDisplayConfig.Call(uintptr(len(paths)), pp, uintptr(len(modes)), pm, uintptr(flags))
	if r != 0 {
		return syscall.Errno(r)
	}
	return nil
}

// SourceGDIName returns the GDI device name ("\\.\DISPLAY1") of a path's source.
func SourceGDIName(p *PathInfo) string {
	var sn SourceDeviceName
	sn.Header.Type = DISPLAYCONFIG_DEVICE_INFO_GET_SOURCE_NAME
	sn.Header.Size = uint32(unsafe.Sizeof(sn))
	sn.Header.AdapterID = p.Source.AdapterID
	sn.Header.ID = p.Source.ID
	r, _, _ := pDisplayConfigGetDeviceInfo.Call(uintptr(unsafe.Pointer(&sn)))
	if r != 0 {
		return ""
	}
	return FromUTF16(sn.ViewGdiDeviceName[:])
}

// TargetFriendlyName returns the monitor's friendly name (from EDID).
func TargetFriendlyName(p *PathInfo) string {
	var tn TargetDeviceName
	tn.Header.Type = DISPLAYCONFIG_DEVICE_INFO_GET_TARGET_NAME
	tn.Header.Size = uint32(unsafe.Sizeof(tn))
	tn.Header.AdapterID = p.Target.AdapterID
	tn.Header.ID = p.Target.ID
	r, _, _ := pDisplayConfigGetDeviceInfo.Call(uintptr(unsafe.Pointer(&tn)))
	if r != 0 {
		return ""
	}
	return FromUTF16(tn.MonitorFriendlyDeviceName[:])
}

// TargetPreferred returns the monitor's preferred (native) resolution.
func TargetPreferred(p *PathInfo) (int, int, bool) {
	var pm TargetPreferredMode
	pm.Header.Type = DISPLAYCONFIG_DEVICE_INFO_GET_TARGET_PREFERRED_MODE
	pm.Header.Size = uint32(unsafe.Sizeof(pm))
	pm.Header.AdapterID = p.Target.AdapterID
	pm.Header.ID = p.Target.ID
	r, _, _ := pDisplayConfigGetDeviceInfo.Call(uintptr(unsafe.Pointer(&pm)))
	if r != 0 || pm.Width == 0 || pm.Height == 0 {
		return 0, 0, false
	}
	return int(pm.Width), int(pm.Height), true
}

// MonitorDeviceOfWindow returns the GDI device name and rect of the monitor
// a window is (mostly) on.
func MonitorDeviceOfWindow(hwnd uintptr) (string, RECT, bool) {
	hmon, _, _ := pMonitorFromWindow.Call(hwnd, MONITOR_DEFAULTTONEAREST)
	if hmon == 0 {
		return "", RECT{}, false
	}
	var mi MonitorInfoEx
	mi.CbSize = uint32(unsafe.Sizeof(mi))
	r, _, _ := pGetMonitorInfoW.Call(hmon, uintptr(unsafe.Pointer(&mi)))
	if r == 0 {
		return "", RECT{}, false
	}
	return FromUTF16(mi.Device[:]), mi.Monitor, true
}

// ---------------------------------------------------------------------------
// Gamma ramps (GPU look-up table — applied by the display hardware)
// ---------------------------------------------------------------------------

// GammaRamp is the WORD[3][256] layout used by Get/SetDeviceGammaRamp.
type GammaRamp [3][256]uint16

func displayDC(device string) (uintptr, error) {
	hdc, _, err := pCreateDCW.Call(uintptr(unsafe.Pointer(UTF16("DISPLAY"))), uintptr(unsafe.Pointer(UTF16(device))), 0, 0)
	if hdc == 0 {
		return 0, fmt.Errorf("CreateDC(%s): %v", device, err)
	}
	return hdc, nil
}

// SetGammaRamp uploads a gamma ramp to a display.
func SetGammaRamp(device string, ramp *GammaRamp) error {
	hdc, err := displayDC(device)
	if err != nil {
		return err
	}
	defer pDeleteDC.Call(hdc)
	r, _, e := pSetDeviceGammaRamp.Call(hdc, uintptr(unsafe.Pointer(ramp)))
	if r == 0 {
		return fmt.Errorf("SetDeviceGammaRamp: %v", e)
	}
	return nil
}

// GetGammaRamp reads the gamma ramp currently loaded for a display.
func GetGammaRamp(device string) (*GammaRamp, error) {
	hdc, err := displayDC(device)
	if err != nil {
		return nil, err
	}
	defer pDeleteDC.Call(hdc)
	var ramp GammaRamp
	r, _, e := pGetDeviceGammaRamp.Call(hdc, uintptr(unsafe.Pointer(&ramp)))
	if r == 0 {
		return nil, fmt.Errorf("GetDeviceGammaRamp: %v", e)
	}
	return &ramp, nil
}

// ---------------------------------------------------------------------------
// Windows & input
// ---------------------------------------------------------------------------

func GetForegroundWindow() uintptr {
	r, _, _ := pGetForegroundWindow.Call()
	return r
}

// WindowPID returns the id of the process owning hwnd. It only reads window
// metadata kept by Windows — no handle to that process is ever opened.
func WindowPID(hwnd uintptr) uint32 {
	var pid uint32
	pGetWindowThreadProcessId.Call(hwnd, uintptr(unsafe.Pointer(&pid)))
	return pid
}

func GetWindowRect(hwnd uintptr) (RECT, bool) {
	var r RECT
	ok, _, _ := pGetWindowRect.Call(hwnd, uintptr(unsafe.Pointer(&r)))
	return r, ok != 0
}

// ClientRectOnScreen returns the client area of hwnd in screen coordinates.
func ClientRectOnScreen(hwnd uintptr) (RECT, bool) {
	var r RECT
	ok, _, _ := pGetClientRect.Call(hwnd, uintptr(unsafe.Pointer(&r)))
	if ok == 0 {
		return r, false
	}
	pt := POINT{0, 0}
	ok, _, _ = pClientToScreen.Call(hwnd, uintptr(unsafe.Pointer(&pt)))
	if ok == 0 {
		return r, false
	}
	return RECT{pt.X, pt.Y, pt.X + r.Right, pt.Y + r.Bottom}, true
}

func ClassName(hwnd uintptr) string {
	var buf [256]uint16
	n, _, _ := pGetClassNameW.Call(hwnd, uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
	if n == 0 {
		return ""
	}
	return FromUTF16(buf[:n])
}

func IsWindowVisible(hwnd uintptr) bool {
	r, _, _ := pIsWindowVisible.Call(hwnd)
	return r != 0
}

func IsIconic(hwnd uintptr) bool {
	r, _, _ := pIsIconic.Call(hwnd)
	return r != 0
}

func IsWindow(hwnd uintptr) bool {
	r, _, _ := pIsWindow.Call(hwnd)
	return r != 0
}

func ShellWindow() uintptr {
	r, _, _ := pGetShellWindow.Call()
	return r
}

func DesktopWindow() uintptr {
	r, _, _ := pGetDesktopWindow.Call()
	return r
}

var (
	enumWindowsMu sync.Mutex
	enumWindowsCb uintptr
	enumWindowsFn func(hwnd uintptr) bool
)

// EnumWindows calls fn for each top-level window (serialised; not re-entrant).
func EnumWindows(fn func(hwnd uintptr) bool) {
	enumWindowsMu.Lock()
	defer enumWindowsMu.Unlock()
	if enumWindowsCb == 0 {
		enumWindowsCb = windows.NewCallback(func(hwnd, _ uintptr) uintptr {
			if enumWindowsFn != nil && enumWindowsFn(hwnd) {
				return 1
			}
			return 0
		})
	}
	enumWindowsFn = fn
	pEnumWindows.Call(enumWindowsCb, 0)
	enumWindowsFn = nil
}

// KeyDown reports whether a virtual key is physically held right now. This
// polls the global async key state; it installs no hook and never swallows
// the key, so the game still receives every press.
func KeyDown(vk int) bool {
	r, _, _ := pGetAsyncKeyState.Call(uintptr(vk))
	return r&0x8000 != 0
}

// SystemParametersInfo wraps SystemParametersInfoW.
func SystemParametersInfo(action, param uint32, pv unsafe.Pointer, winIni uint32) error {
	r, _, e := pSystemParametersInfoW.Call(uintptr(action), uintptr(param), uintptr(pv), uintptr(winIni))
	if r == 0 {
		return fmt.Errorf("SystemParametersInfo(0x%X): %v", action, e)
	}
	return nil
}

// SystemParametersInfoValue is SystemParametersInfo with a value (not a
// pointer) in pvParam, as some SPI_SET* actions require.
func SystemParametersInfoValue(action, param uint32, value uintptr, winIni uint32) error {
	r, _, e := pSystemParametersInfoW.Call(uintptr(action), uintptr(param), value, uintptr(winIni))
	if r == 0 {
		return fmt.Errorf("SystemParametersInfo(0x%X): %v", action, e)
	}
	return nil
}

// BroadcastSettingChange notifies top-level windows that a setting changed.
func BroadcastSettingChange(area string) {
	var result uintptr
	pSendMessageTimeoutW.Call(HWND_BROADCAST, WM_SETTINGCHANGE, 0, uintptr(unsafe.Pointer(UTF16(area))), SMTO_ABORTIFHUNG, 200, uintptr(unsafe.Pointer(&result)))
}

// ---------------------------------------------------------------------------
// Window plumbing used by the host thread
// ---------------------------------------------------------------------------

func RegisterClassEx(wc *WndClassEx) error {
	r, _, e := pRegisterClassExW.Call(uintptr(unsafe.Pointer(wc)))
	if r == 0 {
		return fmt.Errorf("RegisterClassEx: %v", e)
	}
	return nil
}

func CreateWindowEx(exStyle uint32, class, title string, style uint32, x, y, w, h int32, parent, menu, instance uintptr) (uintptr, error) {
	r, _, e := pCreateWindowExW.Call(uintptr(exStyle), uintptr(unsafe.Pointer(UTF16(class))), uintptr(unsafe.Pointer(UTF16(title))),
		uintptr(style), uintptr(x), uintptr(y), uintptr(w), uintptr(h), parent, menu, instance, 0)
	if r == 0 {
		return 0, fmt.Errorf("CreateWindowEx(%s): %v", class, e)
	}
	return r, nil
}

func DefWindowProc(hwnd uintptr, msg uint32, wParam, lParam uintptr) uintptr {
	r, _, _ := pDefWindowProcW.Call(hwnd, uintptr(msg), wParam, lParam)
	return r
}

func DestroyWindow(hwnd uintptr) { pDestroyWindow.Call(hwnd) }

func ShowWindow(hwnd uintptr, cmd int32) { pShowWindow.Call(hwnd, uintptr(cmd)) }

func ShowWindowAsync(hwnd uintptr, cmd int32) { pShowWindowAsync.Call(hwnd, uintptr(cmd)) }

func BringWindowToTop(hwnd uintptr) { pBringWindowToTop.Call(hwnd) }

func SetWindowPos(hwnd, after uintptr, x, y, w, h int32, flags uint32) bool {
	r, _, _ := pSetWindowPos.Call(hwnd, after, uintptr(x), uintptr(y), uintptr(w), uintptr(h), uintptr(flags))
	return r != 0
}

func GetMessage(msg *Msg) int32 {
	r, _, _ := pGetMessageW.Call(uintptr(unsafe.Pointer(msg)), 0, 0, 0)
	return int32(r)
}

func TranslateMessage(msg *Msg) { pTranslateMessage.Call(uintptr(unsafe.Pointer(msg))) }

func DispatchMessage(msg *Msg) { pDispatchMessageW.Call(uintptr(unsafe.Pointer(msg))) }

func PostMessage(hwnd uintptr, msg uint32, wParam, lParam uintptr) bool {
	r, _, _ := pPostMessageW.Call(hwnd, uintptr(msg), wParam, lParam)
	return r != 0
}

func PostQuitMessage(code int32) { pPostQuitMessage.Call(uintptr(code)) }

func RegisterWindowMessage(name string) uint32 {
	r, _, _ := pRegisterWindowMessageW.Call(uintptr(unsafe.Pointer(UTF16(name))))
	return uint32(r)
}

// SetWindowDisplayAffinity hides a window from screenshots, recordings and
// streams (WDA_EXCLUDEFROMCAPTURE, Windows 10 2004+).
func SetWindowDisplayAffinity(hwnd uintptr, affinity uint32) bool {
	r, _, _ := pSetWindowDisplayAffinity.Call(hwnd, uintptr(affinity))
	return r != 0
}

func SetTimer(hwnd, id uintptr, ms uint32) { pSetTimer.Call(hwnd, id, uintptr(ms), 0) }

func KillTimer(hwnd, id uintptr) { pKillTimer.Call(hwnd, id) }

func FindWindow(class, title string) uintptr {
	r, _, _ := pFindWindowW.Call(uintptr(unsafe.Pointer(UTF16(class))), uintptr(unsafe.Pointer(UTF16(title))))
	return r
}

func SetForegroundWindow(hwnd uintptr) bool {
	r, _, _ := pSetForegroundWindow.Call(hwnd)
	return r != 0
}

func CursorPos() POINT {
	var pt POINT
	pGetCursorPos.Call(uintptr(unsafe.Pointer(&pt)))
	return pt
}

// ---------------------------------------------------------------------------
// Layered (per-pixel alpha) windows
// ---------------------------------------------------------------------------

// UpdateLayeredBitmap pushes a premultiplied BGRA bitmap to a layered window
// and moves it to (x, y) in one atomic call.
func UpdateLayeredBitmap(hwnd uintptr, x, y, w, h int32, bgra []byte) error {
	if w <= 0 || h <= 0 || len(bgra) < int(w*h*4) {
		return errors.New("invalid overlay bitmap")
	}
	screen, _, _ := pGetDC.Call(0)
	if screen == 0 {
		return errors.New("GetDC failed")
	}
	defer pReleaseDC.Call(0, screen)
	mem, _, _ := pCreateCompatibleDC.Call(screen)
	if mem == 0 {
		return errors.New("CreateCompatibleDC failed")
	}
	defer pDeleteDC.Call(mem)

	bi := BitmapInfo{Header: BitmapInfoHeader{
		Size: uint32(unsafe.Sizeof(BitmapInfoHeader{})), Width: w, Height: -h, // top-down
		Planes: 1, BitCount: 32, Compression: BI_RGB,
	}}
	var bits uintptr
	bmp, _, _ := pCreateDIBSection.Call(mem, uintptr(unsafe.Pointer(&bi)), DIB_RGB_COLORS, uintptr(unsafe.Pointer(&bits)), 0, 0)
	if bmp == 0 || bits == 0 {
		return errors.New("CreateDIBSection failed")
	}
	defer pDeleteObject.Call(bmp)
	CopyToPtr(bits, bgra[:w*h*4])
	old, _, _ := pSelectObject.Call(mem, bmp)
	defer pSelectObject.Call(mem, old)

	dst := POINT{x, y}
	size := SIZE{w, h}
	src := POINT{0, 0}
	blend := BlendFunction{BlendOp: AC_SRC_OVER, SourceConstantAlpha: 255, AlphaFormat: AC_SRC_ALPHA}
	r, _, e := pUpdateLayeredWindow.Call(hwnd, screen, uintptr(unsafe.Pointer(&dst)), uintptr(unsafe.Pointer(&size)),
		mem, uintptr(unsafe.Pointer(&src)), 0, uintptr(unsafe.Pointer(&blend)), ULW_ALPHA)
	if r == 0 {
		return fmt.Errorf("UpdateLayeredWindow: %v", e)
	}
	return nil
}

// ---------------------------------------------------------------------------
// Tray icon & menus
// ---------------------------------------------------------------------------

func ShellNotifyIcon(msg uint32, nid *NotifyIconData) bool {
	r, _, _ := pShellNotifyIconW.Call(uintptr(msg), uintptr(unsafe.Pointer(nid)))
	return r != 0
}

// ExeIcon extracts the small (16px) and large (32px) icon embedded in exe.
func ExeIcon(exe string) (small, large windows.Handle) {
	pExtractIconExW.Call(uintptr(unsafe.Pointer(UTF16(exe))), 0, uintptr(unsafe.Pointer(&large)), uintptr(unsafe.Pointer(&small)), 1)
	if small == 0 && large == 0 {
		r, _, _ := pLoadIconW.Call(0, 32512) // IDI_APPLICATION
		return windows.Handle(r), windows.Handle(r)
	}
	return small, large
}

func DestroyIcon(h windows.Handle) {
	if h != 0 {
		pDestroyIcon.Call(uintptr(h))
	}
}

func CreatePopupMenu() uintptr {
	r, _, _ := pCreatePopupMenu.Call()
	return r
}

func AppendMenu(menu uintptr, flags uint32, id uintptr, label string) {
	pAppendMenuW.Call(menu, uintptr(flags), id, uintptr(unsafe.Pointer(UTF16(label))))
}

func TrackPopupMenu(menu uintptr, flags uint32, x, y int32, hwnd uintptr) int {
	r, _, _ := pTrackPopupMenu.Call(menu, uintptr(flags), uintptr(x), uintptr(y), 0, hwnd, 0)
	return int(r)
}

func DestroyMenu(menu uintptr) { pDestroyMenu.Call(menu) }
