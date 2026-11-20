//go:build windows

// Package display talks to Windows' display stack: monitors, modes, the
// hardware gamma ramp, the system colour matrix and GPU driver vibrance.
// All of it acts on the display output — never on a game.
package display

import (
	"fmt"
	"sort"
	"strings"

	"github.com/1337133713377/rekashade/internal/win"
)

// Mode is a resolution and the refresh rates available for it.
type Mode struct {
	Width  int   `json:"width"`
	Height int   `json:"height"`
	Rates  []int `json:"rates"`
}

// Info describes one active display.
type Info struct {
	ID           string `json:"id"`     // GDI device name, e.g. \\.\DISPLAY1
	Number       int    `json:"number"` // 1-based, primary first
	Label        string `json:"label"`  // "Display 1"
	Monitor      string `json:"monitor"`
	Adapter      string `json:"adapter"`
	Vendor       string `json:"vendor"` // nvidia | amd | intel | other
	Primary      bool   `json:"primary"`
	X            int    `json:"x"`
	Y            int    `json:"y"`
	Width        int    `json:"width"`
	Height       int    `json:"height"`
	Refresh      int    `json:"refresh"`
	NativeWidth  int    `json:"nativeWidth"`
	NativeHeight int    `json:"nativeHeight"`
	Scaling      string `json:"scaling"`
	Modes        []Mode `json:"modes"`
	Vibrance     string `json:"vibrance"` // "nvidia" | "amd" | "" (no driver vibrance)
}

func vendorFromID(id string) string {
	u := strings.ToUpper(id)
	switch {
	case strings.Contains(u, "VEN_10DE"):
		return "nvidia"
	case strings.Contains(u, "VEN_1002"), strings.Contains(u, "VEN_1022"):
		return "amd"
	case strings.Contains(u, "VEN_8086"):
		return "intel"
	}
	return "other"
}

func scalingName(s uint32) string {
	switch s {
	case win.DISPLAYCONFIG_SCALING_STRETCHED:
		return "stretch"
	case win.DISPLAYCONFIG_SCALING_ASPECTRATIOCENTEREDMAX:
		return "aspect"
	case win.DISPLAYCONFIG_SCALING_CENTERED:
		return "center"
	case win.DISPLAYCONFIG_SCALING_IDENTITY:
		return "identity"
	case win.DISPLAYCONFIG_SCALING_CUSTOM:
		return "custom"
	}
	return "preferred"
}

type targetInfo struct {
	friendly     string
	nativeW      int
	nativeH      int
	scaling      uint32
	hasPreferred bool
}

func targets() map[string]targetInfo {
	out := map[string]targetInfo{}
	paths, _, err := win.QueryActiveDisplayConfig()
	if err != nil {
		return out
	}
	for i := range paths {
		p := &paths[i]
		name := strings.ToUpper(win.SourceGDIName(p))
		if name == "" {
			continue
		}
		if _, dup := out[name]; dup {
			continue // cloned outputs: keep the first
		}
		ti := targetInfo{friendly: win.TargetFriendlyName(p), scaling: p.Target.Scaling}
		ti.nativeW, ti.nativeH, ti.hasPreferred = win.TargetPreferred(p)
		out[name] = ti
	}
	return out
}

// ListModes returns the 32-bit modes a display supports, largest first.
func ListModes(device string) []Mode {
	type key struct{ w, h int }
	rates := map[key]map[int]bool{}
	for i := uint32(0); i < 4096; i++ {
		dm, ok := win.EnumDisplaySettings(device, i)
		if !ok {
			break
		}
		if dm.BitsPerPel != 32 || dm.PelsWidth < 640 || dm.PelsHeight < 480 {
			continue
		}
		k := key{int(dm.PelsWidth), int(dm.PelsHeight)}
		if rates[k] == nil {
			rates[k] = map[int]bool{}
		}
		if dm.DisplayFrequency > 1 {
			rates[k][int(dm.DisplayFrequency)] = true
		}
	}
	modes := make([]Mode, 0, len(rates))
	for k, r := range rates {
		m := Mode{Width: k.w, Height: k.h}
		for hz := range r {
			m.Rates = append(m.Rates, hz)
		}
		sort.Sort(sort.Reverse(sort.IntSlice(m.Rates)))
		modes = append(modes, m)
	}
	sort.Slice(modes, func(i, j int) bool {
		if modes[i].Width*modes[i].Height != modes[j].Width*modes[j].Height {
			return modes[i].Width*modes[i].Height > modes[j].Width*modes[j].Height
		}
		return modes[i].Width > modes[j].Width
	})
	return modes
}

// Enumerate lists the displays attached to the desktop, primary first.
func Enumerate(vibranceFor func(device string) string) []Info {
	tgt := targets()
	var list []Info
	for i := uint32(0); i < 64; i++ {
		dd, ok := win.EnumDisplayDevices("", i, 0)
		if !ok {
			break
		}
		if dd.StateFlags&win.DISPLAY_DEVICE_ATTACHED_TO_DESKTOP == 0 || dd.StateFlags&win.DISPLAY_DEVICE_MIRRORING_DRIVER != 0 {
			continue
		}
		dev := win.FromUTF16(dd.DeviceName[:])
		dm, ok := win.EnumDisplaySettings(dev, win.ENUM_CURRENT_SETTINGS)
		if !ok {
			continue
		}
		info := Info{
			ID:      dev,
			Adapter: win.FromUTF16(dd.DeviceString[:]),
			Vendor:  vendorFromID(win.FromUTF16(dd.DeviceID[:])),
			Primary: dd.StateFlags&win.DISPLAY_DEVICE_PRIMARY_DEVICE != 0,
			X:       int(dm.PositionX), Y: int(dm.PositionY),
			Width: int(dm.PelsWidth), Height: int(dm.PelsHeight),
			Refresh: int(dm.DisplayFrequency),
			Modes:   ListModes(dev),
			Scaling: "preferred",
		}
		if t, ok := tgt[strings.ToUpper(dev)]; ok {
			info.Monitor = t.friendly
			info.Scaling = scalingName(t.scaling)
			if t.hasPreferred {
				info.NativeWidth, info.NativeHeight = t.nativeW, t.nativeH
			}
		}
		if info.Monitor == "" {
			if mon, ok := win.EnumDisplayDevices(dev, 0, 0); ok {
				info.Monitor = win.FromUTF16(mon.DeviceString[:])
			}
		}
		if info.NativeWidth == 0 && len(info.Modes) > 0 {
			info.NativeWidth, info.NativeHeight = info.Modes[0].Width, info.Modes[0].Height
		}
		if vibranceFor != nil {
			info.Vibrance = vibranceFor(dev)
		}
		list = append(list, info)
	}
	sort.SliceStable(list, func(i, j int) bool {
		if list[i].Primary != list[j].Primary {
			return list[i].Primary
		}
		if list[i].X != list[j].X {
			return list[i].X < list[j].X
		}
		return list[i].Y < list[j].Y
	})
	for i := range list {
		list[i].Number = i + 1
		list[i].Label = fmt.Sprintf("Display %d", i+1)
	}
	return list
}

// Find returns the display with the given id, or the primary one.
func Find(list []Info, id string) (Info, bool) {
	for _, d := range list {
		if strings.EqualFold(d.ID, id) {
			return d, true
		}
	}
	for _, d := range list {
		if d.Primary {
			return d, id == ""
		}
	}
	if len(list) > 0 {
		return list[0], false
	}
	return Info{}, false
}
