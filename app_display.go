package main

import (
	"os"
	"strings"
	"time"

	"golang.org/x/sys/windows"

	"github.com/1337133713377/rekashade/internal/config"
	"github.com/1337133713377/rekashade/internal/display"
)

// probeGPU loads the NVIDIA/AMD vibrance interfaces. A marker file guards
// against a buggy driver crashing us: if a probe never finished, the next
// start skips GPU features and tells the user.
func (a *App) probeGPU(cfg config.Config) {
	if a.engine == nil || !cfg.Settings.GPUFeatures {
		return
	}
	a.mu.Lock()
	failed := a.state.GPUProbeFailed
	a.mu.Unlock()
	if failed || a.store == nil {
		return
	}
	marker := a.store.MarkerPath()
	if _, err := os.Stat(marker); err == nil {
		_ = os.Remove(marker)
		a.mu.Lock()
		a.state.GPUProbeFailed = true
		a.saveStateLocked()
		a.mu.Unlock()
		return
	}
	_ = os.WriteFile(marker, []byte(time.Now().Format(time.RFC3339)), 0o644)
	a.mu.Lock()
	saved := a.state.Vibrance
	a.mu.Unlock()
	a.engine.InitGPU(saved)
	_ = os.Remove(marker)
	a.mu.Lock()
	if len(a.state.Vibrance) == 0 {
		a.state.Vibrance = a.engine.Originals()
	}
	a.saveStateLocked()
	a.mu.Unlock()
}

func (a *App) refreshDisplays() {
	var vib func(string) string
	if a.engine != nil {
		vib = a.engine.VibranceBackend
	}
	list := display.Enumerate(vib)
	a.mu.Lock()
	a.displays = list
	a.mu.Unlock()
}

func (a *App) displayIDs() []string {
	a.mu.Lock()
	defer a.mu.Unlock()
	ids := make([]string, len(a.displays))
	for i, d := range a.displays {
		ids[i] = d.ID
	}
	return ids
}

// GetDisplays returns the current list of displays.
func (a *App) GetDisplays() []display.Info {
	a.mu.Lock()
	defer a.mu.Unlock()
	out := make([]display.Info, len(a.displays))
	copy(out, a.displays)
	return out
}

// RefreshDisplays re-scans monitors (e.g. after plugging one in).
func (a *App) RefreshDisplays() []display.Info {
	a.refreshDisplays()
	if a.engine != nil {
		a.engine.CaptureBaselines(a.displayIDs(), false)
	}
	a.applyTuning()
	return a.GetDisplays()
}

// effectiveTuning merges the saved tuning with an active per-game rule.
func (a *App) effectiveTuning() config.Tuning {
	a.mu.Lock()
	defer a.mu.Unlock()
	t := a.cfg.Tuning
	t.Displays = map[string]config.DisplayTuning{}
	for k, v := range a.cfg.Tuning.Displays {
		t.Displays[k] = v
	}
	r := a.watch.rule
	if t.Enabled && r != nil {
		if p, ok := a.cfg.PresetByID(r.presetID); ok {
			t.Hue, t.Saturation = p.Params.Hue, p.Params.Saturation
			dev := r.device
			if dev == "" {
				if d, ok := display.Find(a.displays, ""); ok {
					dev = d.ID
				}
			}
			t.Displays[dev] = config.DisplayTuning{PresetID: p.ID, DisplayParams: p.Params.DisplayParams}
		}
	}
	return t
}

func (a *App) applyTuning() {
	if a.engine == nil {
		return
	}
	a.applyMu.Lock()
	defer a.applyMu.Unlock()
	t := a.effectiveTuning()
	a.engine.Apply(t, a.displayIDs())
	a.mu.Lock()
	if t.Enabled && !a.state.TuningWasOn {
		a.state.TuningWasOn = true
		a.saveStateLocked()
	} else if !t.Enabled && a.state.TuningWasOn {
		a.state.TuningWasOn = false
		a.saveStateLocked()
	}
	a.mu.Unlock()
}

func (a *App) activeRule() string {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.watch.rule != nil && a.cfg.Tuning.Enabled {
		return a.watch.rule.exe
	}
	return ""
}

func (a *App) onDisplayChange() {
	time.Sleep(600 * time.Millisecond) // let Windows settle
	before := strings.Join(a.displayIDs(), "|")
	a.refreshDisplays()
	if a.engine != nil {
		if strings.Join(a.displayIDs(), "|") != before {
			// A monitor was plugged or unplugged: driver handles are stale.
			a.engine.ReloadGPU()
			a.refreshDisplays()
		}
		a.engine.CaptureBaselines(a.displayIDs(), false)
		a.engine.Invalidate()
	}
	a.applyTuning()
	a.applyOverlay()
	a.emit("rs:displays", a.GetDisplays())
	a.emitStatus()
}

func (a *App) onResume() {
	time.Sleep(1500 * time.Millisecond)
	if a.engine != nil {
		a.engine.Invalidate()
	}
	a.applyTuning()
	a.applyOverlay()
}

// ResetDisplays turns tuning off and restores the displays immediately.
func (a *App) ResetDisplays() {
	a.updateConfig(func(c *config.Config) { c.Tuning.Enabled = false })
}

// EnableExtendedGamma lifts Windows' gamma-ramp safety limit (asks for
// administrator rights once; takes effect after signing out or rebooting).
func (a *App) EnableExtendedGamma() error {
	return windows.ShellExecute(0, windows.StringToUTF16Ptr("runas"), windows.StringToUTF16Ptr("reg.exe"),
		windows.StringToUTF16Ptr(`add "HKLM\SOFTWARE\Microsoft\Windows NT\CurrentVersion\ICM" /v GdiIcmGammaRange /t REG_DWORD /d 256 /f`),
		nil, 0)
}

// displayUnderForeground returns the GDI device the active window is on.
func (a *App) displayUnderForeground() string {
	a.mu.Lock()
	dev := a.watch.fg.Monitor
	list := a.displays
	a.mu.Unlock()
	if dev != "" {
		for _, d := range list {
			if strings.EqualFold(d.ID, dev) {
				return d.ID
			}
		}
	}
	if d, ok := display.Find(list, ""); ok {
		return d.ID
	}
	return ""
}
