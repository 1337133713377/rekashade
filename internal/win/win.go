//go:build windows

// Package win holds the raw Win32 bindings Reka Shade needs.
//
// Every call in this package talks to Windows itself (user32, gdi32,
// shell32, kernel32, ...). Nothing here opens, reads, writes or hooks
// another process — that is the whole point of the app: games are never
// touched, so anti-cheat software has nothing to object to.
package win

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	user32   = windows.NewLazySystemDLL("user32.dll")
	gdi32    = windows.NewLazySystemDLL("gdi32.dll")
	kernel32 = windows.NewLazySystemDLL("kernel32.dll")
	shell32  = windows.NewLazySystemDLL("shell32.dll")

	pEnumDisplayDevicesW         = user32.NewProc("EnumDisplayDevicesW")
	pEnumDisplaySettingsExW      = user32.NewProc("EnumDisplaySettingsExW")
	pChangeDisplaySettingsExW    = user32.NewProc("ChangeDisplaySettingsExW")
	pGetDisplayConfigBufferSizes = user32.NewProc("GetDisplayConfigBufferSizes")
	pQueryDisplayConfig          = user32.NewProc("QueryDisplayConfig")
	pSetDisplayConfig            = user32.NewProc("SetDisplayConfig")
	pDisplayConfigGetDeviceInfo  = user32.NewProc("DisplayConfigGetDeviceInfo")
	pMonitorFromWindow           = user32.NewProc("MonitorFromWindow")
	pGetMonitorInfoW             = user32.NewProc("GetMonitorInfoW")
	pGetForegroundWindow         = user32.NewProc("GetForegroundWindow")
	pGetWindowThreadProcessId    = user32.NewProc("GetWindowThreadProcessId")
	pGetWindowRect               = user32.NewProc("GetWindowRect")
	pGetClientRect               = user32.NewProc("GetClientRect")
	pClientToScreen              = user32.NewProc("ClientToScreen")
	pGetClassNameW               = user32.NewProc("GetClassNameW")
	pIsWindowVisible             = user32.NewProc("IsWindowVisible")
	pIsIconic                    = user32.NewProc("IsIconic")
	pIsWindow                    = user32.NewProc("IsWindow")
	pGetShellWindow              = user32.NewProc("GetShellWindow")
	pGetDesktopWindow            = user32.NewProc("GetDesktopWindow")
	pEnumWindows                 = user32.NewProc("EnumWindows")
	pGetAsyncKeyState            = user32.NewProc("GetAsyncKeyState")
	pSystemParametersInfoW       = user32.NewProc("SystemParametersInfoW")
	pRegisterClassExW            = user32.NewProc("RegisterClassExW")
	pCreateWindowExW             = user32.NewProc("CreateWindowExW")
	pDefWindowProcW              = user32.NewProc("DefWindowProcW")
	pDestroyWindow               = user32.NewProc("DestroyWindow")
	pShowWindow                  = user32.NewProc("ShowWindow")
	pSetWindowPos                = user32.NewProc("SetWindowPos")
	pUpdateLayeredWindow         = user32.NewProc("UpdateLayeredWindow")
	pGetMessageW                 = user32.NewProc("GetMessageW")
	pTranslateMessage            = user32.NewProc("TranslateMessage")
	pDispatchMessageW            = user32.NewProc("DispatchMessageW")
	pPostMessageW                = user32.NewProc("PostMessageW")
	pPostQuitMessage             = user32.NewProc("PostQuitMessage")
	pRegisterWindowMessageW      = user32.NewProc("RegisterWindowMessageW")
	pSetWindowDisplayAffinity    = user32.NewProc("SetWindowDisplayAffinity")
	pGetDC                       = user32.NewProc("GetDC")
	pReleaseDC                   = user32.NewProc("ReleaseDC")
	pCreatePopupMenu             = user32.NewProc("CreatePopupMenu")
	pAppendMenuW                 = user32.NewProc("AppendMenuW")
	pTrackPopupMenu              = user32.NewProc("TrackPopupMenu")
	pDestroyMenu                 = user32.NewProc("DestroyMenu")
	pSetForegroundWindow         = user32.NewProc("SetForegroundWindow")
	pGetCursorPos                = user32.NewProc("GetCursorPos")
	pFindWindowW                 = user32.NewProc("FindWindowW")
	pSendMessageTimeoutW         = user32.NewProc("SendMessageTimeoutW")
	pSetTimer                    = user32.NewProc("SetTimer")
	pKillTimer                   = user32.NewProc("KillTimer")
	pDestroyIcon                 = user32.NewProc("DestroyIcon")
	pLoadIconW                   = user32.NewProc("LoadIconW")
	pShowWindowAsync             = user32.NewProc("ShowWindowAsync")
	pBringWindowToTop            = user32.NewProc("BringWindowToTop")

	pCreateDCW          = gdi32.NewProc("CreateDCW")
	pDeleteDC           = gdi32.NewProc("DeleteDC")
	pGetDeviceGammaRamp = gdi32.NewProc("GetDeviceGammaRamp")
	pSetDeviceGammaRamp = gdi32.NewProc("SetDeviceGammaRamp")
	pCreateCompatibleDC = gdi32.NewProc("CreateCompatibleDC")
	pCreateDIBSection   = gdi32.NewProc("CreateDIBSection")
	pSelectObject       = gdi32.NewProc("SelectObject")
	pDeleteObject       = gdi32.NewProc("DeleteObject")

	pRtlMoveMemory            = kernel32.NewProc("RtlMoveMemory")
	pGetUserDefaultUILanguage = kernel32.NewProc("GetUserDefaultUILanguage")

	pShellNotifyIconW = shell32.NewProc("Shell_NotifyIconW")
	pExtractIconExW   = shell32.NewProc("ExtractIconExW")
)

// ---------------------------------------------------------------------------
// Constants
// ---------------------------------------------------------------------------

const (
	DISPLAY_DEVICE_ATTACHED_TO_DESKTOP = 0x00000001
	DISPLAY_DEVICE_PRIMARY_DEVICE      = 0x00000004
	DISPLAY_DEVICE_MIRRORING_DRIVER    = 0x00000008

	ENUM_CURRENT_SETTINGS  = 0xFFFFFFFF
	ENUM_REGISTRY_SETTINGS = 0xFFFFFFFE

	DM_POSITION           = 0x00000020
	DM_BITSPERPEL         = 0x00040000
	DM_PELSWIDTH          = 0x00080000
	DM_PELSHEIGHT         = 0x00100000
	DM_DISPLAYFLAGS       = 0x00200000
	DM_DISPLAYFREQUENCY   = 0x00400000
	DM_DISPLAYFIXEDOUTPUT = 0x20000000

	DMDFO_DEFAULT = 0
	DMDFO_STRETCH = 1
	DMDFO_CENTER  = 2

	CDS_UPDATEREGISTRY = 0x00000001
	CDS_TEST           = 0x00000002

	DISP_CHANGE_SUCCESSFUL  = 0
	DISP_CHANGE_RESTART     = 1
	DISP_CHANGE_FAILED      = -1
	DISP_CHANGE_BADMODE     = -2
	DISP_CHANGE_NOTUPDATED  = -3
	DISP_CHANGE_BADFLAGS    = -4
	DISP_CHANGE_BADPARAM    = -5
	DISP_CHANGE_BADDUALVIEW = -6

	QDC_ONLY_ACTIVE_PATHS = 0x00000002

	SDC_APPLY                       = 0x00000080
	SDC_USE_SUPPLIED_DISPLAY_CONFIG = 0x00000020
	SDC_SAVE_TO_DATABASE            = 0x00000200
	SDC_ALLOW_CHANGES               = 0x00000400

	DISPLAYCONFIG_DEVICE_INFO_GET_SOURCE_NAME           = 1
	DISPLAYCONFIG_DEVICE_INFO_GET_TARGET_NAME           = 2
	DISPLAYCONFIG_DEVICE_INFO_GET_TARGET_PREFERRED_MODE = 3

	DISPLAYCONFIG_SCALING_IDENTITY               = 1
	DISPLAYCONFIG_SCALING_CENTERED               = 2
	DISPLAYCONFIG_SCALING_STRETCHED              = 3
	DISPLAYCONFIG_SCALING_ASPECTRATIOCENTEREDMAX = 4
	DISPLAYCONFIG_SCALING_CUSTOM                 = 5
	DISPLAYCONFIG_SCALING_PREFERRED              = 128

	MONITOR_DEFAULTTONULL    = 0
	MONITOR_DEFAULTTOPRIMARY = 1
	MONITOR_DEFAULTTONEAREST = 2

	WS_POPUP          = 0x80000000
	WS_EX_LAYERED     = 0x00080000
	WS_EX_TRANSPARENT = 0x00000020
	WS_EX_TOPMOST     = 0x00000008
	WS_EX_TOOLWINDOW  = 0x00000080
	WS_EX_NOACTIVATE  = 0x08000000

	SW_HIDE           = 0
	SW_SHOWNORMAL     = 1
	SW_SHOWNOACTIVATE = 4
	SW_SHOW           = 5
	SW_SHOWNA         = 8
	SW_RESTORE        = 9

	SWP_NOSIZE         = 0x0001
	SWP_NOMOVE         = 0x0002
	SWP_NOACTIVATE     = 0x0010
	SWP_SHOWWINDOW     = 0x0040
	SWP_HIDEWINDOW     = 0x0080
	SWP_NOOWNERZORDER  = 0x0200
	SWP_NOSENDCHANGING = 0x0400

	ULW_ALPHA    = 0x00000002
	AC_SRC_OVER  = 0x00
	AC_SRC_ALPHA = 0x01

	BI_RGB         = 0
	DIB_RGB_COLORS = 0

	WM_NULL            = 0x0000
	WM_DESTROY         = 0x0002
	WM_CLOSE           = 0x0010
	WM_QUERYENDSESSION = 0x0011
	WM_ENDSESSION      = 0x0016
	WM_SETTINGCHANGE   = 0x001A
	WM_CONTEXTMENU     = 0x007B
	WM_DISPLAYCHANGE   = 0x007E
	WM_NCHITTEST       = 0x0084
	WM_COMMAND         = 0x0111
	WM_TIMER           = 0x0113
	WM_LBUTTONUP       = 0x0202
	WM_LBUTTONDBLCLK   = 0x0203
	WM_RBUTTONUP       = 0x0205
	WM_POWERBROADCAST  = 0x0218
	WM_APP             = 0x8000

	PBT_APMRESUMESUSPEND   = 0x0007
	PBT_APMRESUMEAUTOMATIC = 0x0012

	WDA_NONE               = 0x00000000
	WDA_EXCLUDEFROMCAPTURE = 0x00000011

	NIM_ADD      = 0x00000000
	NIM_MODIFY   = 0x00000001
	NIM_DELETE   = 0x00000002
	NIF_MESSAGE  = 0x00000001
	NIF_ICON     = 0x00000002
	NIF_TIP      = 0x00000004
	NIF_INFO     = 0x00000010
	NIF_SHOWTIP  = 0x00000080
	NIIF_INFO    = 0x00000001
	NIIF_NOSOUND = 0x00000010

	MF_STRING      = 0x00000000
	MF_GRAYED      = 0x00000001
	MF_CHECKED     = 0x00000008
	MF_POPUP       = 0x00000010
	MF_SEPARATOR   = 0x00000800
	MFT_RADIOCHECK = 0x00000200

	TPM_RIGHTBUTTON = 0x0002
	TPM_NONOTIFY    = 0x0080
	TPM_RETURNCMD   = 0x0100

	SPIF_UPDATEINIFILE = 0x01
	SPIF_SENDCHANGE    = 0x02

	SMTO_ABORTIFHUNG = 0x0002
	HWND_BROADCAST   = 0xFFFF
)

// HWND_TOPMOST is (HWND)-1.
var HWND_TOPMOST = ^uintptr(0)

// HTTRANSPARENT is the WM_NCHITTEST result (-1) that makes a window click-through.
var HTTRANSPARENT = ^uintptr(0)

// ---------------------------------------------------------------------------
// Structures. Sizes are asserted at compile time where layout matters.
// ---------------------------------------------------------------------------

type RECT struct{ Left, Top, Right, Bottom int32 }

type POINT struct{ X, Y int32 }

type SIZE struct{ CX, CY int32 }

// DisplayDevice mirrors DISPLAY_DEVICEW.
type DisplayDevice struct {
	Cb           uint32
	DeviceName   [32]uint16
	DeviceString [128]uint16
	StateFlags   uint32
	DeviceID     [128]uint16
	DeviceKey    [128]uint16
}

// DevMode mirrors DEVMODEW (display flavour of the unions).
type DevMode struct {
	DeviceName         [32]uint16
	SpecVersion        uint16
	DriverVersion      uint16
	Size               uint16
	DriverExtra        uint16
	Fields             uint32
	PositionX          int32
	PositionY          int32
	DisplayOrientation uint32
	DisplayFixedOutput uint32
	Color              int16
	Duplex             int16
	YResolution        int16
	TTOption           int16
	Collate            int16
	FormName           [32]uint16
	LogPixels          uint16
	BitsPerPel         uint32
	PelsWidth          uint32
	PelsHeight         uint32
	DisplayFlags       uint32
	DisplayFrequency   uint32
	ICMMethod          uint32
	ICMIntent          uint32
	MediaType          uint32
	DitherType         uint32
	Reserved1          uint32
	Reserved2          uint32
	PanningWidth       uint32
	PanningHeight      uint32
}

// MonitorInfoEx mirrors MONITORINFOEXW.
type MonitorInfoEx struct {
	CbSize  uint32
	Monitor RECT
	Work    RECT
	Flags   uint32
	Device  [32]uint16
}

type LUID struct {
	LowPart  uint32
	HighPart int32
}

// PathSourceInfo mirrors DISPLAYCONFIG_PATH_SOURCE_INFO.
type PathSourceInfo struct {
	AdapterID   LUID
	ID          uint32
	ModeInfoIdx uint32
	StatusFlags uint32
}

// PathTargetInfo mirrors DISPLAYCONFIG_PATH_TARGET_INFO.
type PathTargetInfo struct {
	AdapterID        LUID
	ID               uint32
	ModeInfoIdx      uint32
	OutputTechnology uint32
	Rotation         uint32
	Scaling          uint32
	RefreshNum       uint32
	RefreshDen       uint32
	ScanLineOrdering uint32
	TargetAvailable  int32
	StatusFlags      uint32
}

// PathInfo mirrors DISPLAYCONFIG_PATH_INFO.
type PathInfo struct {
	Source PathSourceInfo
	Target PathTargetInfo
	Flags  uint32
}

// ModeInfo mirrors DISPLAYCONFIG_MODE_INFO (the union is kept opaque).
type ModeInfo struct {
	InfoType  uint32
	ID        uint32
	AdapterID LUID
	Data      [48]byte
}

// DeviceInfoHeader mirrors DISPLAYCONFIG_DEVICE_INFO_HEADER.
type DeviceInfoHeader struct {
	Type      uint32
	Size      uint32
	AdapterID LUID
	ID        uint32
}

// SourceDeviceName mirrors DISPLAYCONFIG_SOURCE_DEVICE_NAME.
type SourceDeviceName struct {
	Header            DeviceInfoHeader
	ViewGdiDeviceName [32]uint16
}

// TargetDeviceName mirrors DISPLAYCONFIG_TARGET_DEVICE_NAME.
type TargetDeviceName struct {
	Header                    DeviceInfoHeader
	Flags                     uint32
	OutputTechnology          uint32
	EdidManufactureID         uint16
	EdidProductCodeID         uint16
	ConnectorInstance         uint32
	MonitorFriendlyDeviceName [64]uint16
	MonitorDevicePath         [128]uint16
}

// TargetPreferredMode mirrors DISPLAYCONFIG_TARGET_PREFERRED_MODE. The C
// struct aligns targetMode on 8 bytes (it starts with a UINT64), hence _pad.
type TargetPreferredMode struct {
	Header     DeviceInfoHeader
	Width      uint32
	Height     uint32
	_pad       uint32
	TargetMode [48]byte
}

// WndClassEx mirrors WNDCLASSEXW.
type WndClassEx struct {
	Size       uint32
	Style      uint32
	WndProc    uintptr
	ClsExtra   int32
	WndExtra   int32
	Instance   windows.Handle
	Icon       windows.Handle
	Cursor     windows.Handle
	Background windows.Handle
	MenuName   *uint16
	ClassName  *uint16
	IconSm     windows.Handle
}

// Msg mirrors MSG.
type Msg struct {
	Hwnd     windows.HWND
	Message  uint32
	WParam   uintptr
	LParam   uintptr
	Time     uint32
	Pt       POINT
	LPrivate uint32
}

// NotifyIconData mirrors NOTIFYICONDATAW (Vista+ layout).
type NotifyIconData struct {
	CbSize          uint32
	Wnd             windows.HWND
	ID              uint32
	Flags           uint32
	CallbackMessage uint32
	Icon            windows.Handle
	Tip             [128]uint16
	State           uint32
	StateMask       uint32
	Info            [256]uint16
	Version         uint32
	InfoTitle       [64]uint16
	InfoFlags       uint32
	GuidItem        windows.GUID
	BalloonIcon     windows.Handle
}

// BitmapInfoHeader mirrors BITMAPINFOHEADER.
type BitmapInfoHeader struct {
	Size          uint32
	Width         int32
	Height        int32
	Planes        uint16
	BitCount      uint16
	Compression   uint32
	SizeImage     uint32
	XPelsPerMeter int32
	YPelsPerMeter int32
	ClrUsed       uint32
	ClrImportant  uint32
}

// BitmapInfo mirrors BITMAPINFO with a single (unused) palette entry.
type BitmapInfo struct {
	Header BitmapInfoHeader
	Colors [1]uint32
}

// BlendFunction mirrors BLENDFUNCTION.
type BlendFunction struct {
	BlendOp             byte
	BlendFlags          byte
	SourceConstantAlpha byte
	AlphaFormat         byte
}

// Compile-time layout checks: each line fails to compile if the Go struct
// size drifts from the documented Win32 size.
var (
	_ [840 - unsafe.Sizeof(DisplayDevice{})]byte
	_ [unsafe.Sizeof(DisplayDevice{}) - 840]byte
	_ [220 - unsafe.Sizeof(DevMode{})]byte
	_ [unsafe.Sizeof(DevMode{}) - 220]byte
	_ [104 - unsafe.Sizeof(MonitorInfoEx{})]byte
	_ [unsafe.Sizeof(MonitorInfoEx{}) - 104]byte
	_ [72 - unsafe.Sizeof(PathInfo{})]byte
	_ [unsafe.Sizeof(PathInfo{}) - 72]byte
	_ [64 - unsafe.Sizeof(ModeInfo{})]byte
	_ [unsafe.Sizeof(ModeInfo{}) - 64]byte
	_ [84 - unsafe.Sizeof(SourceDeviceName{})]byte
	_ [unsafe.Sizeof(SourceDeviceName{}) - 84]byte
	_ [420 - unsafe.Sizeof(TargetDeviceName{})]byte
	_ [unsafe.Sizeof(TargetDeviceName{}) - 420]byte
	_ [80 - unsafe.Sizeof(TargetPreferredMode{})]byte
	_ [unsafe.Sizeof(TargetPreferredMode{}) - 80]byte
	_ [80 - unsafe.Sizeof(WndClassEx{})]byte
	_ [unsafe.Sizeof(WndClassEx{}) - 80]byte
	_ [48 - unsafe.Sizeof(Msg{})]byte
	_ [unsafe.Sizeof(Msg{}) - 48]byte
	_ [976 - unsafe.Sizeof(NotifyIconData{})]byte
	_ [unsafe.Sizeof(NotifyIconData{}) - 976]byte
	_ [40 - unsafe.Sizeof(BitmapInfoHeader{})]byte
	_ [unsafe.Sizeof(BitmapInfoHeader{}) - 40]byte
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// UTF16 converts s to a NUL-terminated UTF-16 pointer (nil for "").
func UTF16(s string) *uint16 {
	if s == "" {
		return nil
	}
	p, err := windows.UTF16PtrFromString(s)
	if err != nil {
		return nil
	}
	return p
}

// FromUTF16 converts a fixed UTF-16 buffer to a Go string.
func FromUTF16(b []uint16) string { return windows.UTF16ToString(b) }

// CopyFromPtr copies len(dst) bytes from native memory at src.
func CopyFromPtr(dst []byte, src uintptr) {
	if len(dst) == 0 || src == 0 {
		return
	}
	pRtlMoveMemory.Call(uintptr(unsafe.Pointer(&dst[0])), src, uintptr(len(dst)))
}

// CopyToPtr copies src into native memory at dst.
func CopyToPtr(dst uintptr, src []byte) {
	if len(src) == 0 || dst == 0 {
		return
	}
	pRtlMoveMemory.Call(dst, uintptr(unsafe.Pointer(&src[0])), uintptr(len(src)))
}

// UserUILanguage returns the primary language id of the user's UI language
// (e.g. 0x0C for French, 0x09 for English).
func UserUILanguage() uint16 {
	r, _, _ := pGetUserDefaultUILanguage.Call()
	return uint16(r) & 0x3FF
}
