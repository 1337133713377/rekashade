//go:build windows

// Package optimizer applies a handful of well-known, reversible Windows
// settings that help games feel smoother. Every tweak records the original
// value before changing anything, so "Undo" puts the system back exactly as
// it was. None of them require administrator rights and none touch a game.
package optimizer

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"

	"github.com/1337133713377/rekashade/internal/win"
)

// Status is one tweak as shown in the UI.
type Status struct {
	ID        string `json:"id"`
	On        bool   `json:"on"`
	Supported bool   `json:"supported"`
	CanUndo   bool   `json:"canUndo"`
	Error     string `json:"error,omitempty"`
}

// Backups maps a tweak id to its saved original values (JSON).
type Backups map[string]json.RawMessage

type tweak struct {
	id        string
	supported func() bool
	get       func() (bool, error)
	apply     func() (any, error) // returns the backup to keep
	undo      func(raw json.RawMessage) error
}

var tweaks []tweak

func osBuild() uint32 {
	return windows.RtlGetVersion().BuildNumber
}

func always() bool { return true }

// IDs lists the tweaks in display order.
func IDs() []string {
	ids := make([]string, len(tweaks))
	for i, t := range tweaks {
		ids[i] = t.id
	}
	return ids
}

func find(id string) (tweak, bool) {
	for _, t := range tweaks {
		if t.id == id {
			return t, true
		}
	}
	return tweak{}, false
}

func status(t tweak, b Backups) Status {
	s := Status{ID: t.id, Supported: t.supported()}
	if !s.Supported {
		return s
	}
	on, err := t.get()
	s.On = on
	if err != nil {
		s.Error = err.Error()
	}
	_, s.CanUndo = b[t.id]
	return s
}

// List returns the current state of every tweak (read from the system).
func List(b Backups) []Status {
	out := make([]Status, 0, len(tweaks))
	for _, t := range tweaks {
		out = append(out, status(t, b))
	}
	return out
}

// Set turns a tweak on (optimise) or off (undo). b is updated in place.
func Set(id string, on bool, b Backups) (Status, error) {
	t, ok := find(id)
	if !ok {
		return Status{}, fmt.Errorf("unknown tweak %q", id)
	}
	if !t.supported() {
		return status(t, b), errors.New("not available on this version of Windows")
	}
	if on {
		cur, _ := t.get()
		if !cur {
			backup, err := t.apply()
			if err != nil {
				return status(t, b), err
			}
			if _, exists := b[id]; !exists {
				raw, _ := json.Marshal(backup)
				b[id] = raw
			}
		}
	} else {
		raw, exists := b[id]
		if !exists {
			return status(t, b), errors.New("nothing to undo: this was already set before Reka Shade")
		}
		if err := t.undo(raw); err != nil {
			return status(t, b), err
		}
		delete(b, id)
	}
	return status(t, b), nil
}

// ---------------------------------------------------------------------------
// Registry helpers
// ---------------------------------------------------------------------------

type regValue struct {
	Path   string `json:"path"`
	Name   string `json:"name"`
	Exists bool   `json:"exists"`
	Kind   uint32 `json:"kind"`
	Str    string `json:"str,omitempty"`
	Num    uint64 `json:"num,omitempty"`
}

func readReg(path, name string) regValue {
	v := regValue{Path: path, Name: name}
	k, err := registry.OpenKey(registry.CURRENT_USER, path, registry.QUERY_VALUE)
	if err != nil {
		return v
	}
	defer k.Close()
	if s, kind, err := k.GetStringValue(name); err == nil {
		v.Exists, v.Kind, v.Str = true, kind, s
		return v
	}
	if n, kind, err := k.GetIntegerValue(name); err == nil {
		v.Exists, v.Kind, v.Num = true, kind, n
	}
	return v
}

func writeReg(v regValue) error {
	k, _, err := registry.CreateKey(registry.CURRENT_USER, v.Path, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer k.Close()
	if !v.Exists {
		err := k.DeleteValue(v.Name)
		if errors.Is(err, registry.ErrNotExist) {
			return nil
		}
		return err
	}
	switch v.Kind {
	case registry.SZ:
		return k.SetStringValue(v.Name, v.Str)
	case registry.EXPAND_SZ:
		return k.SetExpandStringValue(v.Name, v.Str)
	case registry.QWORD:
		return k.SetQWordValue(v.Name, v.Num)
	default:
		return k.SetDWordValue(v.Name, uint32(v.Num))
	}
}

func dword(path, name string, n uint32) regValue {
	return regValue{Path: path, Name: name, Exists: true, Kind: registry.DWORD, Num: uint64(n)}
}

func regNum(path, name string, def uint64) uint64 {
	v := readReg(path, name)
	if !v.Exists {
		return def
	}
	if v.Kind == registry.SZ {
		var n uint64
		fmt.Sscan(v.Str, &n)
		return n
	}
	return v.Num
}

// regTweak builds a tweak that writes DWORD values and restores them on undo.
func regTweak(id string, supported func() bool, isOn func() bool, values []regValue, after func()) tweak {
	return tweak{
		id:        id,
		supported: supported,
		get:       func() (bool, error) { return isOn(), nil },
		apply: func() (any, error) {
			var backup []regValue
			for _, v := range values {
				backup = append(backup, readReg(v.Path, v.Name))
			}
			for _, v := range values {
				if err := writeReg(v); err != nil {
					for _, b := range backup {
						_ = writeReg(b)
					}
					return nil, err
				}
			}
			if after != nil {
				after()
			}
			return backup, nil
		},
		undo: func(raw json.RawMessage) error {
			var backup []regValue
			if err := json.Unmarshal(raw, &backup); err != nil {
				return err
			}
			var firstErr error
			for _, b := range backup {
				if err := writeReg(b); err != nil && firstErr == nil {
					firstErr = err
				}
			}
			if after != nil {
				after()
			}
			return firstErr
		},
	}
}

// ---------------------------------------------------------------------------
// Power plan (powrprof.dll — same as Control Panel › Power Options)
// ---------------------------------------------------------------------------

var (
	powrprof              = windows.NewLazySystemDLL("powrprof.dll")
	pPowerGetActiveScheme = powrprof.NewProc("PowerGetActiveScheme")
	pPowerSetActiveScheme = powrprof.NewProc("PowerSetActiveScheme")
	pPowerDuplicateScheme = powrprof.NewProc("PowerDuplicateScheme")

	guidHighPerf = windows.GUID{Data1: 0x8c5e7fda, Data2: 0xe8bf, Data3: 0x4a96, Data4: [8]byte{0x9a, 0x85, 0xa6, 0xe2, 0x3a, 0x8c, 0x63, 0x5c}}
	guidUltimate = windows.GUID{Data1: 0xe9a42b02, Data2: 0xd5df, Data3: 0x448d, Data4: [8]byte{0xaa, 0x00, 0x03, 0xf1, 0x47, 0x49, 0xeb, 0x61}}
	guidBalanced = windows.GUID{Data1: 0x381b4222, Data2: 0xf694, Data3: 0x41f0, Data4: [8]byte{0x96, 0x85, 0xff, 0x5b, 0xb2, 0x60, 0xdf, 0x2e}}
)

func activeScheme() (windows.GUID, error) {
	var p *windows.GUID
	r, _, _ := pPowerGetActiveScheme.Call(0, uintptr(unsafe.Pointer(&p)))
	if r != 0 || p == nil {
		return windows.GUID{}, fmt.Errorf("PowerGetActiveScheme: %v", windows.Errno(r))
	}
	g := *p
	windows.LocalFree(windows.Handle(unsafe.Pointer(p)))
	return g, nil
}

func setScheme(g windows.GUID) error {
	r, _, _ := pPowerSetActiveScheme.Call(0, uintptr(unsafe.Pointer(&g)))
	if r != 0 {
		return fmt.Errorf("PowerSetActiveScheme: %v", windows.Errno(r))
	}
	return nil
}

type powerBackup struct {
	Previous string `json:"previous"`
	Created  string `json:"created,omitempty"`
}

var createdHighPerf windows.GUID

func powerTweak() tweak {
	return tweak{
		id:        "powerPlan",
		supported: always,
		get: func() (bool, error) {
			g, err := activeScheme()
			if err != nil {
				return false, err
			}
			return g == guidHighPerf || g == guidUltimate || (createdHighPerf != windows.GUID{} && g == createdHighPerf), nil
		},
		apply: func() (any, error) {
			prev, err := activeScheme()
			if err != nil {
				return nil, err
			}
			b := powerBackup{Previous: prev.String()}
			if err := setScheme(guidHighPerf); err != nil {
				// Hidden on many laptops (Modern Standby): make a copy of it.
				var p *windows.GUID
				r, _, _ := pPowerDuplicateScheme.Call(0, uintptr(unsafe.Pointer(&guidHighPerf)), uintptr(unsafe.Pointer(&p)))
				if r != 0 || p == nil {
					return nil, fmt.Errorf("the High performance plan is not available: %v", windows.Errno(r))
				}
				created := *p
				windows.LocalFree(windows.Handle(unsafe.Pointer(p)))
				if err := setScheme(created); err != nil {
					return nil, err
				}
				createdHighPerf = created
				b.Created = created.String()
			}
			return b, nil
		},
		undo: func(raw json.RawMessage) error {
			var b powerBackup
			_ = json.Unmarshal(raw, &b)
			g, err := windows.GUIDFromString(b.Previous)
			if err != nil {
				g = guidBalanced
			}
			return setScheme(g)
		},
	}
}

// ---------------------------------------------------------------------------
// SystemParametersInfo-based tweaks (applied live, saved to the profile)
// ---------------------------------------------------------------------------

const (
	spiGetMouse               = 0x0003
	spiSetMouse               = 0x0004
	spiGetFilterKeys          = 0x0032
	spiSetFilterKeys          = 0x0033
	spiGetToggleKeys          = 0x0034
	spiSetToggleKeys          = 0x0035
	spiGetStickyKeys          = 0x003A
	spiSetStickyKeys          = 0x003B
	spiGetAnimation           = 0x0048
	spiSetAnimation           = 0x0049
	spiGetClientAreaAnimation = 0x1042
	spiSetClientAreaAnimation = 0x1043
	hotkeyActive              = 0x00000004
	winIni                    = win.SPIF_UPDATEINIFILE | win.SPIF_SENDCHANGE
)

func mouseTweak() tweak {
	return tweak{
		id:        "mouseAccel",
		supported: always,
		get: func() (bool, error) {
			var m [3]int32
			if err := win.SystemParametersInfo(spiGetMouse, 0, unsafe.Pointer(&m), 0); err != nil {
				return false, err
			}
			return m[2] == 0, nil // m[2] = "Enhance pointer precision"
		},
		apply: func() (any, error) {
			var m [3]int32
			if err := win.SystemParametersInfo(spiGetMouse, 0, unsafe.Pointer(&m), 0); err != nil {
				return nil, err
			}
			off := [3]int32{0, 0, 0}
			return m, win.SystemParametersInfo(spiSetMouse, 0, unsafe.Pointer(&off), winIni)
		},
		undo: func(raw json.RawMessage) error {
			var m [3]int32
			if err := json.Unmarshal(raw, &m); err != nil {
				return err
			}
			return win.SystemParametersInfo(spiSetMouse, 0, unsafe.Pointer(&m), winIni)
		},
	}
}

type accessFlags struct {
	Sticky uint32 `json:"sticky"`
	Filter uint32 `json:"filter"`
	Toggle uint32 `json:"toggle"`
}

type stickyKeys struct{ Size, Flags uint32 }
type toggleKeys struct{ Size, Flags uint32 }
type filterKeys struct{ Size, Flags, Wait, Delay, Repeat, Bounce uint32 }

func readAccess() (stickyKeys, filterKeys, toggleKeys, error) {
	s := stickyKeys{Size: 8}
	f := filterKeys{Size: 24}
	t := toggleKeys{Size: 8}
	if err := win.SystemParametersInfo(spiGetStickyKeys, s.Size, unsafe.Pointer(&s), 0); err != nil {
		return s, f, t, err
	}
	if err := win.SystemParametersInfo(spiGetFilterKeys, f.Size, unsafe.Pointer(&f), 0); err != nil {
		return s, f, t, err
	}
	if err := win.SystemParametersInfo(spiGetToggleKeys, t.Size, unsafe.Pointer(&t), 0); err != nil {
		return s, f, t, err
	}
	return s, f, t, nil
}

func writeAccess(a accessFlags) error {
	s, f, t, err := readAccess()
	if err != nil {
		return err
	}
	s.Flags, f.Flags, t.Flags = a.Sticky, a.Filter, a.Toggle
	if err := win.SystemParametersInfo(spiSetStickyKeys, s.Size, unsafe.Pointer(&s), winIni); err != nil {
		return err
	}
	if err := win.SystemParametersInfo(spiSetFilterKeys, f.Size, unsafe.Pointer(&f), winIni); err != nil {
		return err
	}
	return win.SystemParametersInfo(spiSetToggleKeys, t.Size, unsafe.Pointer(&t), winIni)
}

func stickyTweak() tweak {
	return tweak{
		id:        "stickyKeys",
		supported: always,
		get: func() (bool, error) {
			s, _, _, err := readAccess()
			return s.Flags&hotkeyActive == 0, err
		},
		apply: func() (any, error) {
			s, f, t, err := readAccess()
			if err != nil {
				return nil, err
			}
			orig := accessFlags{s.Flags, f.Flags, t.Flags}
			return orig, writeAccess(accessFlags{s.Flags &^ hotkeyActive, f.Flags &^ hotkeyActive, t.Flags &^ hotkeyActive})
		},
		undo: func(raw json.RawMessage) error {
			var a accessFlags
			if err := json.Unmarshal(raw, &a); err != nil {
				return err
			}
			return writeAccess(a)
		},
	}
}

type animBackup struct {
	MinAnimate int32 `json:"minAnimate"`
	ClientArea int32 `json:"clientArea"`
}

func animationTweak() tweak {
	read := func() (animBackup, error) {
		ai := struct{ Size, MinAnimate int32 }{Size: 8}
		if err := win.SystemParametersInfo(spiGetAnimation, 8, unsafe.Pointer(&ai), 0); err != nil {
			return animBackup{}, err
		}
		var ca int32 = 1
		_ = win.SystemParametersInfo(spiGetClientAreaAnimation, 0, unsafe.Pointer(&ca), 0)
		return animBackup{MinAnimate: ai.MinAnimate, ClientArea: ca}, nil
	}
	write := func(b animBackup) error {
		ai := struct{ Size, MinAnimate int32 }{Size: 8, MinAnimate: b.MinAnimate}
		if err := win.SystemParametersInfo(spiSetAnimation, 8, unsafe.Pointer(&ai), winIni); err != nil {
			return err
		}
		_ = win.SystemParametersInfoValue(spiSetClientAreaAnimation, 0, uintptr(b.ClientArea), winIni)
		return nil
	}
	return tweak{
		id:        "animations",
		supported: always,
		get: func() (bool, error) {
			b, err := read()
			return b.MinAnimate == 0, err
		},
		apply: func() (any, error) {
			b, err := read()
			if err != nil {
				return nil, err
			}
			return b, write(animBackup{MinAnimate: 0, ClientArea: 0})
		},
		undo: func(raw json.RawMessage) error {
			var b animBackup
			if err := json.Unmarshal(raw, &b); err != nil {
				return err
			}
			return write(b)
		},
	}
}

// ---------------------------------------------------------------------------
// Windowed-game optimisations (Windows 11): a key in a semicolon list
// ---------------------------------------------------------------------------

const dxPrefsPath = `Software\Microsoft\DirectX\UserGpuPreferences`
const dxPrefsName = "DirectXUserGlobalSettings"

func dxSetting(s, key string) (string, bool) {
	for _, part := range strings.Split(s, ";") {
		kv := strings.SplitN(strings.TrimSpace(part), "=", 2)
		if len(kv) == 2 && strings.EqualFold(kv[0], key) {
			return kv[1], true
		}
	}
	return "", false
}

func dxWith(s, key, val string) string {
	var parts []string
	replaced := false
	for _, part := range strings.Split(s, ";") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		kv := strings.SplitN(part, "=", 2)
		if strings.EqualFold(kv[0], key) {
			part, replaced = key+"="+val, true
		}
		parts = append(parts, part)
	}
	if !replaced {
		parts = append(parts, key+"="+val)
	}
	return strings.Join(parts, ";") + ";"
}

func windowedTweak() tweak {
	return tweak{
		id:        "windowedBoost",
		supported: func() bool { return osBuild() >= 22000 },
		get: func() (bool, error) {
			v := readReg(dxPrefsPath, dxPrefsName)
			val, ok := dxSetting(v.Str, "SwapEffectUpgradeEnable")
			return ok && val == "1", nil
		},
		apply: func() (any, error) {
			orig := readReg(dxPrefsPath, dxPrefsName)
			next := regValue{Path: dxPrefsPath, Name: dxPrefsName, Exists: true, Kind: registry.SZ, Str: dxWith(orig.Str, "SwapEffectUpgradeEnable", "1")}
			return orig, writeReg(next)
		},
		undo: func(raw json.RawMessage) error {
			var orig regValue
			if err := json.Unmarshal(raw, &orig); err != nil {
				return err
			}
			return writeReg(orig)
		},
	}
}

// ---------------------------------------------------------------------------
// Temp folder cleaner
// ---------------------------------------------------------------------------

// CleanResult summarises a temp clean-up.
type CleanResult struct {
	Files   int   `json:"files"`
	Bytes   int64 `json:"bytes"`
	Skipped int   `json:"skipped"`
}

// CleanTemp deletes files older than a day from the user's own temp folder.
// Files that are in use are simply skipped.
func CleanTemp() CleanResult {
	var res CleanResult
	root := os.TempDir()
	cutoff := time.Now().Add(-24 * time.Hour)
	var dirs []string
	_ = filepath.WalkDir(root, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if p == root {
			return nil
		}
		if d.Type()&os.ModeSymlink != 0 {
			return nil // never follow links out of the temp folder
		}
		if d.IsDir() {
			dirs = append(dirs, p)
			return nil
		}
		info, err := d.Info()
		if err != nil || info.ModTime().After(cutoff) {
			return nil
		}
		if err := os.Remove(p); err != nil {
			res.Skipped++
			return nil
		}
		res.Files++
		res.Bytes += info.Size()
		return nil
	})
	for i := len(dirs) - 1; i >= 0; i-- { // deepest first; only empty ones go
		if info, err := os.Stat(dirs[i]); err == nil && info.ModTime().Before(cutoff) {
			_ = os.Remove(dirs[i])
		}
	}
	return res
}

func init() {
	const (
		gameBar     = `Software\Microsoft\GameBar`
		gameConfig  = `System\GameConfigStore`
		gameDVR     = `Software\Microsoft\Windows\CurrentVersion\GameDVR`
		push        = `Software\Microsoft\Windows\CurrentVersion\PushNotifications`
		notifSet    = `Software\Microsoft\Windows\CurrentVersion\Notifications\Settings`
		bgAccess    = `Software\Microsoft\Windows\CurrentVersion\BackgroundAccessApplications`
		search      = `Software\Microsoft\Windows\CurrentVersion\Search`
		personalize = `Software\Microsoft\Windows\CurrentVersion\Themes\Personalize`
	)
	tweaks = []tweak{
		powerTweak(),
		regTweak("gameMode", always,
			func() bool { return regNum(gameBar, "AutoGameModeEnabled", 1) != 0 },
			[]regValue{dword(gameBar, "AutoGameModeEnabled", 1), dword(gameBar, "AllowAutoGameMode", 1)}, nil),
		regTweak("gameDVR", always,
			func() bool {
				return regNum(gameConfig, "GameDVR_Enabled", 1) == 0 && regNum(gameDVR, "AppCaptureEnabled", 1) == 0
			},
			[]regValue{dword(gameConfig, "GameDVR_Enabled", 0), dword(gameDVR, "AppCaptureEnabled", 0), dword(gameDVR, "HistoricalCaptureEnabled", 0)}, nil),
		windowedTweak(),
		mouseTweak(),
		stickyTweak(),
		regTweak("notifications", always,
			func() bool { return regNum(push, "ToastEnabled", 1) == 0 },
			[]regValue{dword(push, "ToastEnabled", 0), dword(notifSet, "NOC_GLOBAL_SETTING_TOASTS_ENABLED", 0)},
			func() { win.BroadcastSettingChange("Policy") }),
		regTweak("backgroundApps", always,
			func() bool { return regNum(bgAccess, "GlobalUserDisabled", 0) == 1 },
			[]regValue{dword(bgAccess, "GlobalUserDisabled", 1), dword(search, "BackgroundAppGlobalToggle", 0)}, nil),
		regTweak("transparency", always,
			func() bool { return regNum(personalize, "EnableTransparency", 1) == 0 },
			[]regValue{dword(personalize, "EnableTransparency", 0)},
			func() { win.BroadcastSettingChange("ImmersiveColorSet") }),
		animationTweak(),
	}
}
