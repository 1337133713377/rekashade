package main

import (
	"errors"
	"sync"
	"time"

	"github.com/1337133713377/rekashade/internal/config"
	"github.com/1337133713377/rekashade/internal/display"
)

// Stretched resolution: the desktop itself switches to e.g. 1440×1080 and
// the GPU scales it to fill the panel. Games are never touched — they just
// see a different desktop resolution, exactly like changing it in Windows
// Settings. A new mode must be kept within 15 seconds or it reverts, so a
// mode the monitor cannot show can never leave you with a black screen.

const confirmSeconds = 15

type stretchRuntime struct {
	mu          sync.Mutex
	active      bool
	auto        bool
	device      string
	prevScaling uint32
	width       int
	height      int
	refresh     int
	pending     int // seconds left to confirm, 0 = none
	cancel      chan struct{}
	lastError   string
	suppressed  bool // user turned auto stretch off until the game exits
}

// StretchStatus is shown in the UI.
type StretchStatus struct {
	Active  bool   `json:"active"`
	Auto    bool   `json:"auto"`
	Device  string `json:"device"`
	Width   int    `json:"width"`
	Height  int    `json:"height"`
	Refresh int    `json:"refresh"`
	Pending int    `json:"pending"`
	Error   string `json:"error,omitempty"`
}

func (a *App) stretchStatus() StretchStatus {
	s := &a.stretch
	s.mu.Lock()
	defer s.mu.Unlock()
	return StretchStatus{Active: s.active, Auto: s.auto, Device: s.device, Width: s.width, Height: s.height,
		Refresh: s.refresh, Pending: s.pending, Error: s.lastError}
}

func (a *App) stretchActive() bool {
	a.stretch.mu.Lock()
	defer a.stretch.mu.Unlock()
	return a.stretch.active
}

// ResolutionRequest is a mode to try.
type ResolutionRequest struct {
	Display string `json:"display"`
	Width   int    `json:"width"`
	Height  int    `json:"height"`
	Refresh int    `json:"refresh"`
	Scaling string `json:"scaling"`
}

// TestResolution applies a mode and starts the keep/revert countdown.
func (a *App) TestResolution(req ResolutionRequest) (StretchStatus, error) {
	if req.Width < 640 || req.Height < 480 {
		return a.stretchStatus(), errors.New("resolution too small")
	}
	dev := req.Display
	if d, ok := display.Find(a.GetDisplays(), dev); ok {
		dev = d.ID
	}
	if dev == "" {
		return a.stretchStatus(), errors.New("no display found")
	}
	if err := a.applyStretch(dev, req.Width, req.Height, req.Refresh, req.Scaling, false); err != nil {
		return a.stretchStatus(), err
	}
	a.updateConfig(func(c *config.Config) {
		r := c.Resolution
		same := r.Display == dev && r.Width == req.Width && r.Height == req.Height && r.Refresh == req.Refresh && r.Scaling == req.Scaling
		c.Resolution.Confirmed = r.Confirmed && same
		c.Resolution.Display, c.Resolution.Width, c.Resolution.Height = dev, req.Width, req.Height
		c.Resolution.Refresh, c.Resolution.Scaling = req.Refresh, req.Scaling
	})
	a.startStretchCountdown()
	return a.stretchStatus(), nil
}

// ConfirmResolution keeps the mode being tested.
func (a *App) ConfirmResolution() error {
	a.stopStretchCountdown()
	if !a.stretchActive() {
		return errors.New("no resolution to keep")
	}
	a.updateConfig(func(c *config.Config) { c.Resolution.Confirmed = true })
	a.emitStatus()
	return nil
}

// RevertResolution goes back to the normal Windows resolution now.
func (a *App) RevertResolution() error {
	a.stopStretchCountdown()
	a.suppressAutoStretch()
	return a.revertStretch(true)
}

// suppressAutoStretch stops auto stretch from re-applying right after the
// user switched it off by hand while the game is still running.
func (a *App) suppressAutoStretch() {
	a.stretch.mu.Lock()
	if a.stretch.active && a.stretch.auto {
		a.stretch.suppressed = true
	}
	a.stretch.mu.Unlock()
}

func (a *App) applyStretch(dev string, w, h, hz int, scaling string, auto bool) error {
	s := &a.stretch
	s.mu.Lock()
	if s.active && s.device != dev {
		_ = display.RestoreMode(s.device, s.prevScaling)
		s.active = false
	}
	if !s.active {
		s.prevScaling, _ = display.CurrentScaling(dev)
	}
	err := display.ApplyMode(dev, w, h, hz, scaling)
	if err != nil {
		s.lastError = err.Error()
		s.mu.Unlock()
		a.emitStatus()
		return err
	}
	s.active, s.auto, s.device, s.width, s.height, s.refresh, s.lastError = true, auto, dev, w, h, hz, ""
	prev := s.prevScaling
	s.mu.Unlock()

	a.mu.Lock()
	a.state.Stretch = config.StretchState{Active: true, Device: dev, PrevScaling: prev}
	a.saveStateLocked()
	a.mu.Unlock()
	a.afterModeChange()
	return nil
}

func (a *App) revertStretch(notify bool) error {
	s := &a.stretch
	s.mu.Lock()
	if !s.active {
		s.mu.Unlock()
		return nil
	}
	err := display.RestoreMode(s.device, s.prevScaling)
	s.active, s.auto, s.pending = false, false, 0
	if err != nil {
		s.lastError = err.Error()
	}
	s.mu.Unlock()

	a.mu.Lock()
	a.state.Stretch = config.StretchState{}
	a.saveStateLocked()
	a.mu.Unlock()
	if notify {
		a.afterModeChange()
	}
	return err
}

func (a *App) afterModeChange() {
	time.Sleep(400 * time.Millisecond)
	a.refreshDisplays()
	if a.engine != nil {
		a.engine.Invalidate()
	}
	a.applyTuning()
	a.applyOverlay()
	a.emit("rs:displays", a.GetDisplays())
	a.emitStatus()
}

func (a *App) startStretchCountdown() {
	s := &a.stretch
	s.mu.Lock()
	if s.cancel != nil {
		close(s.cancel)
	}
	cancel := make(chan struct{})
	s.cancel = cancel
	s.pending = confirmSeconds
	s.mu.Unlock()
	a.emitStatus()

	go func() {
		t := time.NewTicker(time.Second)
		defer t.Stop()
		for {
			select {
			case <-cancel:
				return
			case <-a.stopCh:
				return
			case <-t.C:
			}
			s.mu.Lock()
			if s.cancel != cancel {
				s.mu.Unlock()
				return
			}
			s.pending--
			left := s.pending
			if left <= 0 {
				s.cancel = nil
			}
			s.mu.Unlock()
			if left <= 0 {
				_ = a.revertStretch(true)
				a.toast("info", "toast.resReverted", nil)
				return
			}
			a.emitStatus()
		}
	}()
}

func (a *App) stopStretchCountdown() {
	s := &a.stretch
	s.mu.Lock()
	if s.cancel != nil {
		close(s.cancel)
		s.cancel = nil
	}
	s.pending = 0
	s.mu.Unlock()
}

// toggleStretch is the hotkey/tray action.
func (a *App) toggleStretch() {
	if a.stretchActive() {
		a.suppressAutoStretch()
		_ = a.revertStretch(true)
		return
	}
	cfg := a.config().Resolution
	if !cfg.Confirmed {
		a.ShowWindow()
		a.toast("info", "toast.resTestFirst", nil)
		return
	}
	if err := a.applyStretch(cfg.Display, cfg.Width, cfg.Height, cfg.Refresh, cfg.Scaling, false); err != nil {
		a.toast("error", "toast.resFailed", map[string]any{"error": err.Error()})
	}
}
