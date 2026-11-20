//go:build windows

package display

import (
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"

	"github.com/1337133713377/rekashade/internal/colorx"
	"github.com/1337133713377/rekashade/internal/config"
	"github.com/1337133713377/rekashade/internal/host"
	"github.com/1337133713377/rekashade/internal/win"
)

// ---------------------------------------------------------------------------
// Windows colour matrix (Magnification API full-screen colour effect — the
// same engine as Settings › Accessibility › Colour filters). It is applied
// by the desktop compositor to everything on every screen.
// ---------------------------------------------------------------------------

var (
	magDLL       = windows.NewLazySystemDLL("magnification.dll")
	pMagInit     = magDLL.NewProc("MagInitialize")
	pMagUninit   = magDLL.NewProc("MagUninitialize")
	pMagSetColor = magDLL.NewProc("MagSetFullscreenColorEffect")
)

type magnifier struct {
	host   *host.Host
	inited bool
	last   [25]float32
	has    bool
}

// apply must be called with the engine lock held; it hops to the host thread
// because the Magnification runtime is bound to the thread that started it.
func (m *magnifier) apply(mat colorx.Mat3) error {
	var err error
	m.host.Call(func() {
		if mat.IsIdentity() {
			if m.inited {
				id := colorx.MagEffect(colorx.Identity3())
				pMagSetColor.Call(uintptr(unsafe.Pointer(&id)))
				pMagUninit.Call()
				m.inited, m.has = false, false
			}
			return
		}
		if !m.inited {
			if e := magDLL.Load(); e != nil {
				err = fmt.Errorf("Windows colour engine unavailable: %v", e)
				return
			}
			if r, _, e := pMagInit.Call(); r == 0 {
				err = fmt.Errorf("MagInitialize: %v", e)
				return
			}
			m.inited = true
		}
		eff := colorx.MagEffect(mat)
		if m.has && eff == m.last {
			return
		}
		if r, _, e := pMagSetColor.Call(uintptr(unsafe.Pointer(&eff))); r == 0 {
			err = fmt.Errorf("MagSetFullscreenColorEffect: %v", e)
			return
		}
		m.last, m.has = eff, true
	})
	return err
}

// ---------------------------------------------------------------------------
// Engine
// ---------------------------------------------------------------------------

// GammaStatus reports how a display's ramp was applied.
type GammaStatus struct {
	OK       bool    `json:"ok"`
	Limited  bool    `json:"limited"`
	Strength float64 `json:"strength"` // fraction of the requested change Windows accepted
	Error    string  `json:"error,omitempty"`
}

// Status is the engine state shown in the UI.
type Status struct {
	Active      bool                   `json:"active"`
	Matrix      string                 `json:"matrix"` // off | on | error
	MatrixError string                 `json:"matrixError,omitempty"`
	Gamma       map[string]GammaStatus `json:"gamma"`
	Vibrance    map[string]string      `json:"vibrance"` // nvidia | amd | matrix | off
	GPU         []string               `json:"gpu"`
	GPUError    string                 `json:"gpuError,omitempty"`
}

// Engine applies tunings to the displays.
type Engine struct {
	mu       sync.Mutex
	mag      magnifier
	gpus     []vibranceBackend
	gpuErr   string
	applied  map[string]colorx.Ramp // ramp currently loaded per display
	wanted   map[string]colorx.Ramp // last ramp requested per display
	baseline map[string]colorx.Ramp // what was loaded before us (calibration)
	gamma    map[string]GammaStatus
	status   Status
}

// NewEngine creates the engine. ignoreBaseline should be true when the
// previous session ended while tuned (the loaded ramp is ours, not a
// calibration to preserve).
func NewEngine(h *host.Host) *Engine {
	return &Engine{
		mag:      magnifier{host: h},
		applied:  map[string]colorx.Ramp{},
		wanted:   map[string]colorx.Ramp{},
		baseline: map[string]colorx.Ramp{},
		gamma:    map[string]GammaStatus{},
		status:   Status{Matrix: "off", Gamma: map[string]GammaStatus{}, Vibrance: map[string]string{}},
	}
}

// CaptureBaselines records each display's current ramp so that turning
// tuning off restores the user's own calibration instead of a flat ramp.
func (e *Engine) CaptureBaselines(devices []string, ignore bool) {
	e.mu.Lock()
	defer e.mu.Unlock()
	for _, d := range devices {
		key := strings.ToUpper(d)
		if _, ok := e.baseline[key]; ok {
			continue
		}
		id := colorx.IdentityRamp()
		if ignore {
			e.baseline[key] = id
			continue
		}
		r, err := win.GetGammaRamp(d)
		if err != nil || !monotonic((*colorx.Ramp)(r)) {
			e.baseline[key] = id
			continue
		}
		e.baseline[key] = colorx.Ramp(*r)
	}
}

func monotonic(r *colorx.Ramp) bool {
	for c := 0; c < 3; c++ {
		if r[c][255] < 0x4000 { // nearly black ramp: not a sane calibration
			return false
		}
		for i := 1; i < 256; i++ {
			if r[c][i] < r[c][i-1] {
				return false
			}
		}
	}
	return true
}

// compose maps a tuning ramp through the display's baseline ramp.
func compose(base, t *colorx.Ramp) colorx.Ramp {
	var out colorx.Ramp
	for c := 0; c < 3; c++ {
		for i := 0; i < 256; i++ {
			v := float64(t[c][i]) / 65535 * 255
			lo := int(math.Floor(v))
			if lo >= 255 {
				out[c][i] = base[c][255]
				continue
			}
			f := v - float64(lo)
			out[c][i] = uint16(math.Round(float64(base[c][lo])*(1-f) + float64(base[c][lo+1])*f))
		}
	}
	return out
}

// InitGPU loads the NVIDIA / AMD driver interfaces for digital vibrance.
// saved holds original levels from a session that did not exit cleanly.
func (e *Engine) InitGPU(saved map[string]int) {
	e.mu.Lock()
	defer e.mu.Unlock()
	var errs []string
	if n, err := loadNvidia(saved); err == nil {
		e.gpus = append(e.gpus, n)
	} else if !errors.Is(err, windows.ERROR_MOD_NOT_FOUND) && !strings.Contains(err.Error(), "The specified module could not be found") {
		errs = append(errs, err.Error())
	}
	if a, err := loadAMD(saved); err == nil {
		e.gpus = append(e.gpus, a)
	} else if !errors.Is(err, windows.ERROR_MOD_NOT_FOUND) && !strings.Contains(err.Error(), "The specified module could not be found") {
		errs = append(errs, err.Error())
	}
	e.gpuErr = strings.Join(errs, "; ")
	e.status.GPU = e.gpuNames()
	e.status.GPUError = e.gpuErr
}

// ReloadGPU re-probes drivers (after a display change handles may be stale).
func (e *Engine) ReloadGPU() {
	e.mu.Lock()
	saved := map[string]int{}
	for _, g := range e.gpus {
		for k, v := range g.originals() {
			saved[k] = v
		}
		g.close()
	}
	e.gpus = nil
	e.mu.Unlock()
	e.InitGPU(saved)
}

func (e *Engine) gpuNames() []string {
	var n []string
	for _, g := range e.gpus {
		n = append(n, g.name())
	}
	return n
}

func (e *Engine) backend(device string) vibranceBackend {
	for _, g := range e.gpus {
		if g.has(device) {
			return g
		}
	}
	return nil
}

// VibranceBackend returns "nvidia", "amd" or "" for a display.
func (e *Engine) VibranceBackend(device string) string {
	e.mu.Lock()
	defer e.mu.Unlock()
	if b := e.backend(device); b != nil {
		return b.name()
	}
	return ""
}

// Originals returns the vibrance levels to restore on exit.
func (e *Engine) Originals() map[string]int {
	e.mu.Lock()
	defer e.mu.Unlock()
	out := map[string]int{}
	for _, g := range e.gpus {
		for k, v := range g.originals() {
			out[k] = v
		}
	}
	return out
}

func (e *Engine) setRamp(device string, target colorx.Ramp) GammaStatus {
	key := strings.ToUpper(device)
	base, ok := e.baseline[key]
	if !ok {
		base = colorx.IdentityRamp()
	}
	final := compose(&base, &target)
	if prev, ok := e.wanted[key]; ok && prev == final {
		if gs, ok := e.gamma[key]; ok {
			return gs // unchanged request: nothing to do (or already refused)
		}
	}
	e.wanted[key] = final
	gs := GammaStatus{}
	if err := win.SetGammaRamp(device, (*win.GammaRamp)(&final)); err == nil {
		e.applied[key] = final
		gs = GammaStatus{OK: true, Strength: 1}
	} else {
		// Windows refuses ramps that stray too far from linear unless the
		// "extended gamma range" registry switch is on. Find the strongest
		// version of the requested look it will accept.
		lo, hi := 0.0, 1.0
		var best *colorx.Ramp
		for i := 0; i < 6; i++ {
			mid := (lo + hi) / 2
			r := colorx.LerpRamp(&base, &final, mid)
			if win.SetGammaRamp(device, (*win.GammaRamp)(&r)) == nil {
				lo, best = mid, &r
			} else {
				hi = mid
			}
		}
		if best != nil {
			// The last call may have been a rejected one: put the best back.
			_ = win.SetGammaRamp(device, (*win.GammaRamp)(best))
			e.applied[key] = *best
			gs = GammaStatus{OK: true, Limited: true, Strength: math.Round(lo*100) / 100}
		} else {
			gs = GammaStatus{OK: false, Error: err.Error()}
		}
	}
	e.gamma[key] = gs
	return gs
}

// Apply pushes a tuning to every display. It is cheap to call repeatedly:
// only what changed reaches Windows/the driver.
func (e *Engine) Apply(t config.Tuning, devices []string) Status {
	e.mu.Lock()
	defer e.mu.Unlock()
	if !t.Enabled {
		e.resetLocked()
		return e.snapshot()
	}
	st := Status{Active: true, Gamma: map[string]GammaStatus{}, Vibrance: map[string]string{}, GPU: e.gpuNames(), GPUError: e.gpuErr}
	vibFactor := 1.0
	for _, dev := range devices {
		dt, ok := t.Displays[dev]
		if !ok {
			for k, v := range t.Displays {
				if strings.EqualFold(k, dev) {
					dt, ok = v, true
					break
				}
			}
		}
		p := config.NeutralDisplay()
		if ok {
			p = config.SanitizeDisplay(dt.DisplayParams)
		}
		tone := colorx.Tone{Brightness: p.Brightness, Contrast: p.Contrast, Gamma: p.Gamma, Temperature: p.Temperature, Red: p.Red, Green: p.Green, Blue: p.Blue}
		st.Gamma[dev] = e.setRamp(dev, colorx.BuildRamp(tone))

		factor := p.Vibrance / 100
		if b := e.backend(dev); b != nil {
			rem, err := b.set(dev, p.Vibrance)
			if err == nil {
				st.Vibrance[dev] = b.name()
				factor = rem
			} else {
				st.Vibrance[dev] = "matrix"
			}
		} else if math.Abs(p.Vibrance-100) > 0.5 {
			st.Vibrance[dev] = "matrix"
		} else {
			st.Vibrance[dev] = "off"
		}
		// The colour matrix is global, so keep the strongest fallback.
		if math.Abs(math.Log(math.Max(factor, 0.01))) > math.Abs(math.Log(math.Max(vibFactor, 0.01))) {
			vibFactor = factor
		}
	}
	mat := colorx.ColorMatrix(t.Hue, t.Saturation*vibFactor)
	if err := e.mag.apply(mat); err != nil {
		st.Matrix, st.MatrixError = "error", err.Error()
	} else if mat.IsIdentity() {
		st.Matrix = "off"
	} else {
		st.Matrix = "on"
	}
	e.status = st
	return e.snapshot()
}

func (e *Engine) resetLocked() {
	for key, base := range e.baseline {
		if _, touched := e.applied[key]; touched {
			b := base
			_ = win.SetGammaRamp(key, (*win.GammaRamp)(&b))
		}
	}
	e.applied = map[string]colorx.Ramp{}
	e.wanted = map[string]colorx.Ramp{}
	e.gamma = map[string]GammaStatus{}
	for _, g := range e.gpus {
		for dev := range g.originals() {
			_ = g.restore(dev)
		}
	}
	_ = e.mag.apply(colorx.Identity3())
	e.status = Status{Matrix: "off", Gamma: map[string]GammaStatus{}, Vibrance: map[string]string{}, GPU: e.gpuNames(), GPUError: e.gpuErr}
}

// Reset restores everything Reka Shade changed on the displays.
func (e *Engine) Reset() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.resetLocked()
}

// ResetHardwareBestEffort restores gamma ramps and driver vibrance without
// touching the colour matrix (which needs the host thread). Used when the
// Windows session ends: it never blocks for long, even if another goroutine
// is mid-apply.
func (e *Engine) ResetHardwareBestEffort(wait time.Duration) {
	deadline := time.Now().Add(wait)
	for !e.mu.TryLock() {
		if time.Now().After(deadline) {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	defer e.mu.Unlock()
	for key, base := range e.baseline {
		if _, touched := e.applied[key]; touched {
			b := base
			_ = win.SetGammaRamp(key, (*win.GammaRamp)(&b))
		}
	}
	for _, g := range e.gpus {
		for dev := range g.originals() {
			_ = g.restore(dev)
		}
	}
}

// Close resets and releases driver interfaces.
func (e *Engine) Close() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.resetLocked()
	for _, g := range e.gpus {
		g.close()
	}
	e.gpus = nil
}

// KeepAlive re-uploads ramps that something else (a game switching to
// exclusive fullscreen, a driver reset, sleep) replaced behind our back.
func (e *Engine) KeepAlive() {
	e.mu.Lock()
	defer e.mu.Unlock()
	for key, want := range e.applied {
		cur, err := win.GetGammaRamp(key)
		if err != nil {
			continue
		}
		if (*colorx.Ramp)(cur).Distance(&want) > 512 {
			w := want
			_ = win.SetGammaRamp(key, (*win.GammaRamp)(&w))
		}
	}
}

// Invalidate forgets cached state so the next Apply re-uploads everything
// (used after display changes and resume from sleep).
func (e *Engine) Invalidate() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.applied = map[string]colorx.Ramp{}
	e.wanted = map[string]colorx.Ramp{}
	e.gamma = map[string]GammaStatus{}
	e.mag.has = false
}

// Status returns the last applied state.
func (e *Engine) Status() Status {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.snapshot()
}

func (e *Engine) snapshot() Status {
	s := e.status
	s.Gamma = map[string]GammaStatus{}
	for k, v := range e.status.Gamma {
		s.Gamma[k] = v
	}
	s.Vibrance = map[string]string{}
	for k, v := range e.status.Vibrance {
		s.Vibrance[k] = v
	}
	s.GPU = append([]string(nil), e.status.GPU...)
	sort.Strings(s.GPU)
	return s
}
