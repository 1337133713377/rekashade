//go:build windows

package display

import (
	"errors"
	"fmt"
	"strings"

	"github.com/1337133713377/rekashade/internal/win"
)

// ScalingCode converts "stretch" | "aspect" | "center" to DISPLAYCONFIG_SCALING.
func ScalingCode(s string) uint32 {
	switch s {
	case "aspect":
		return win.DISPLAYCONFIG_SCALING_ASPECTRATIOCENTEREDMAX
	case "center":
		return win.DISPLAYCONFIG_SCALING_CENTERED
	}
	return win.DISPLAYCONFIG_SCALING_STRETCHED
}

func fixedOutput(s string) uint32 {
	switch s {
	case "center":
		return win.DMDFO_CENTER
	case "aspect":
		return win.DMDFO_DEFAULT
	}
	return win.DMDFO_STRETCH
}

// CurrentScaling returns the DISPLAYCONFIG_SCALING value of a display.
func CurrentScaling(device string) (uint32, error) {
	paths, _, err := win.QueryActiveDisplayConfig()
	if err != nil {
		return 0, err
	}
	for i := range paths {
		if strings.EqualFold(win.SourceGDIName(&paths[i]), device) {
			return paths[i].Target.Scaling, nil
		}
	}
	return 0, errors.New("display not found")
}

// SetScaling tells Windows/the GPU how to fit a lower resolution on the panel
// (stretched, black bars, or centred) — the same switch as "Perform scaling
// on GPU" in the driver control panel. The change is not saved permanently.
func SetScaling(device string, scaling uint32) error {
	paths, modes, err := win.QueryActiveDisplayConfig()
	if err != nil {
		return err
	}
	found := false
	for i := range paths {
		if strings.EqualFold(win.SourceGDIName(&paths[i]), device) {
			if paths[i].Target.Scaling == scaling {
				return nil
			}
			paths[i].Target.Scaling = scaling
			found = true
		}
	}
	if !found {
		return errors.New("display not found")
	}
	return win.SetDisplayConfig(paths, modes, win.SDC_APPLY|win.SDC_USE_SUPPLIED_DISPLAY_CONFIG|win.SDC_ALLOW_CHANGES)
}

// ApplyMode switches a display to w×h@hz until it is reverted (the mode is
// not written to the registry, so a reboot always brings the native mode
// back). scaling is "stretch", "aspect" or "center".
func ApplyMode(device string, w, h, hz int, scaling string) error {
	cur, ok := win.EnumDisplaySettings(device, win.ENUM_CURRENT_SETTINGS)
	if !ok {
		return fmt.Errorf("cannot read the current mode of %s", device)
	}
	if hz <= 0 {
		// "Best available": the highest rate this mode supports, otherwise
		// Windows may fall back to 60 Hz.
		for _, m := range ListModes(device) {
			if m.Width == w && m.Height == h && len(m.Rates) > 0 {
				hz = m.Rates[0]
				break
			}
		}
	}
	dm := cur
	dm.PelsWidth, dm.PelsHeight = uint32(w), uint32(h)
	dm.Fields = win.DM_PELSWIDTH | win.DM_PELSHEIGHT
	if hz > 0 {
		dm.DisplayFrequency = uint32(hz)
		dm.Fields |= win.DM_DISPLAYFREQUENCY
	}
	withScaling := dm
	withScaling.DisplayFixedOutput = fixedOutput(scaling)
	withScaling.Fields |= win.DM_DISPLAYFIXEDOUTPUT

	code := win.ChangeDisplaySettings(device, &withScaling, 0)
	if code != win.DISP_CHANGE_SUCCESSFUL {
		// Some drivers reject DM_DISPLAYFIXEDOUTPUT: retry without it.
		code = win.ChangeDisplaySettings(device, &dm, 0)
	}
	if err := win.DispChangeError(code); err != nil {
		return err
	}
	// Ask Windows for the matching GPU scaling mode as well (best effort:
	// some setups only honour the driver control panel).
	_ = SetScaling(device, ScalingCode(scaling))
	return nil
}

// RestoreMode returns a display to the mode saved in the registry (the
// user's normal Windows resolution) and restores its previous scaling.
func RestoreMode(device string, prevScaling uint32) error {
	err := win.DispChangeError(win.ChangeDisplaySettings(device, nil, 0))
	if prevScaling != 0 {
		_ = SetScaling(device, prevScaling)
	}
	return err
}

// RegistryMode returns the display's saved (normal) resolution.
func RegistryMode(device string) (w, h, hz int, ok bool) {
	dm, ok := win.EnumDisplaySettings(device, win.ENUM_REGISTRY_SETTINGS)
	if !ok {
		return 0, 0, 0, false
	}
	return int(dm.PelsWidth), int(dm.PelsHeight), int(dm.DisplayFrequency), true
}

// CurrentMode returns the display's current resolution.
func CurrentMode(device string) (w, h, hz int, ok bool) {
	dm, ok := win.EnumDisplaySettings(device, win.ENUM_CURRENT_SETTINGS)
	if !ok {
		return 0, 0, 0, false
	}
	return int(dm.PelsWidth), int(dm.PelsHeight), int(dm.DisplayFrequency), true
}
