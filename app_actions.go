package main

import (
	"strings"
	"time"

	"github.com/1337133713377/rekashade/internal/config"
	"github.com/1337133713377/rekashade/internal/host"
	"github.com/1337133713377/rekashade/internal/watch"
	"github.com/1337133713377/rekashade/internal/win"
)

const (
	vkShift   = 0x10
	vkControl = 0x11
	vkMenu    = 0x12
)

// ---------------------------------------------------------------------------
// Key binds — polled, never hooked: the game still receives every key.
// ---------------------------------------------------------------------------

func (a *App) keyLoop() {
	t := time.NewTicker(15 * time.Millisecond)
	defer t.Stop()
	prev := map[int]bool{}
	for {
		select {
		case <-a.stopCh:
			return
		case <-t.C:
		}
		if a.bindsPaused.Load() {
			continue
		}
		safely(func() { a.pollKeys(prev) })
	}
}

// safely runs f and swallows a panic so a background loop keeps going.
func safely(f func()) {
	defer func() { _ = recover() }()
	f()
}

func (a *App) pollKeys(prev map[int]bool) {
	a.mu.Lock()
	binds := a.cfg.Binds
	a.mu.Unlock()
	if len(binds) == 0 {
		return
	}
	var ctrl, alt, shift, mods bool
	for i, b := range binds {
		if b.Key <= 0 {
			continue
		}
		down := win.KeyDown(b.Key)
		if b.Action == "holdHide" {
			if down != prev[i] {
				a.setHoldHidden(down)
			}
			prev[i] = down
			continue
		}
		if !mods {
			ctrl, alt, shift, mods = win.KeyDown(vkControl), win.KeyDown(vkMenu), win.KeyDown(vkShift), true
		}
		match := down && ctrl == b.Ctrl && alt == b.Alt && shift == b.Shift
		if match && !prev[i] {
			go a.runAction(b.Action)
		}
		prev[i] = match
	}
}

func (a *App) runAction(action string) {
	switch action {
	case "toggleTuning":
		a.updateConfig(func(c *config.Config) { c.Tuning.Enabled = !c.Tuning.Enabled })
	case "toggleCrosshair":
		a.updateConfig(func(c *config.Config) { c.Overlay.Enabled = !c.Overlay.Enabled })
	case "nextPreset":
		a.cyclePreset(1)
	case "prevPreset":
		a.cyclePreset(-1)
	case "toggleStretch":
		a.toggleStretch()
	case "showApp":
		a.ShowWindow()
	case "panic":
		a.Panic()
	}
}

// Panic turns everything off at once.
func (a *App) Panic() {
	a.stopStretchCountdown()
	a.suppressAutoStretch()
	_ = a.revertStretch(true)
	a.updateConfig(func(c *config.Config) {
		c.Tuning.Enabled = false
		c.Overlay.Enabled = false
	})
}

func (a *App) cyclePreset(dir int) {
	dev := a.displayUnderForeground()
	if dev == "" {
		return
	}
	a.updateConfig(func(c *config.Config) {
		if len(c.Presets) == 0 {
			return
		}
		cur := c.DisplayTuningFor(dev).PresetID
		idx := -1
		for i, p := range c.Presets {
			if p.ID == cur {
				idx = i
				break
			}
		}
		next := (idx + dir + len(c.Presets)) % len(c.Presets)
		if idx < 0 && dir < 0 {
			next = len(c.Presets) - 1
		}
		c.ApplyPreset(dev, c.Presets[next])
		c.Tuning.Enabled = true
	})
}

// ---------------------------------------------------------------------------
// Watch loop: foreground window, per-game presets, auto stretch, keep-alive
// ---------------------------------------------------------------------------

type gameRule struct {
	exe      string
	presetID string
	device   string
}

type watchRuntime struct {
	fg          watch.Foreground
	rule        *gameRule
	lastFgHWND  uintptr
	keepAliveAt time.Time
}

func (a *App) watchLoop() {
	t := time.NewTicker(250 * time.Millisecond)
	defer t.Stop()
	w := &watchTickState{}
	for {
		select {
		case <-a.stopCh:
			return
		case <-t.C:
		}
		w.tick++
		safely(func() { a.watchTick(w) })
	}
}

type watchTickState struct {
	tick  int
	procs map[uint32]string
}

func (a *App) watchTick(w *watchTickState) {
	tick := w.tick
	cfg := a.config()
	rulesOn := cfg.Tuning.Enabled && hasEnabledRule(cfg.Rules)
	autoStretch := cfg.Resolution.AutoApply && cfg.Resolution.Confirmed && len(cfg.Resolution.AutoGames) > 0
	if (rulesOn || autoStretch) && (w.procs == nil || tick%4 == 0) {
		w.procs = watch.Processes()
	} else if !rulesOn && !autoStretch {
		w.procs = nil
	}
	procs := w.procs

	fg := watch.GetForeground(procs, a.selfPID)
	a.mu.Lock()
	changed := fg.HWND != a.watch.lastFgHWND
	a.watch.lastFgHWND = fg.HWND
	if !cfg.Overlay.FollowWindow || changed {
		a.watch.fg = fg
	} else {
		a.watch.fg.Exe = fg.Exe
		a.watch.fg.Monitor = fg.Monitor
		a.watch.fg.Fullscreen = fg.Fullscreen
	}
	a.mu.Unlock()

	if cfg.Overlay.OnlyFullscreen {
		a.setFullscreenFocus(fg.Fullscreen && fg.PID != a.selfPID)
	}
	if cfg.Overlay.FollowWindow && changed && cfg.Overlay.Enabled && a.host != nil {
		x, y := a.overlayTarget()
		a.host.SetOverlayTarget(x, y)
	}

	// Per-game presets.
	if rulesOn || a.activeRule() != "" {
		a.evalRules(cfg, fg)
	}

	// Auto stretch while a listed game is running (checked every second).
	if tick%4 == 0 {
		a.evalAutoStretch(cfg, procs, autoStretch)
	}

	// Games switching to exclusive fullscreen may reset the gamma ramp:
	// re-check shortly after focus changes and every few seconds.
	if cfg.Tuning.Enabled && a.engine != nil {
		now := time.Now()
		a.mu.Lock()
		if changed {
			a.watch.keepAliveAt = now.Add(1500 * time.Millisecond)
		}
		due := !a.watch.keepAliveAt.IsZero() && now.After(a.watch.keepAliveAt)
		a.mu.Unlock()
		if due || tick%12 == 0 {
			a.mu.Lock()
			a.watch.keepAliveAt = time.Time{}
			a.mu.Unlock()
			a.engine.KeepAlive()
		}
	}
}

func hasEnabledRule(rules []config.GameRule) bool {
	for _, r := range rules {
		if r.Enabled && r.Exe != "" {
			return true
		}
	}
	return false
}

func (a *App) evalRules(cfg config.Config, fg watch.Foreground) {
	var match *gameRule
	if cfg.Tuning.Enabled && fg.Exe != "" {
		for _, r := range cfg.Rules {
			if r.Enabled && r.Exe == fg.Exe {
				if _, ok := cfg.PresetByID(r.PresetID); ok {
					match = &gameRule{exe: r.Exe, presetID: r.PresetID, device: fg.Monitor}
				}
				break
			}
		}
	}
	a.mu.Lock()
	cur := a.watch.rule
	same := (cur == nil && match == nil) || (cur != nil && match != nil && *cur == *match)
	if !same {
		a.watch.rule = match
	}
	a.mu.Unlock()
	if !same {
		a.applyTuning()
		a.emitStatus()
	}
}

func (a *App) evalAutoStretch(cfg config.Config, procs map[uint32]string, enabled bool) {
	running := false
	if enabled {
		want := map[string]bool{}
		for _, g := range cfg.Resolution.AutoGames {
			want[g] = true
		}
		for _, exe := range procs {
			if want[exe] {
				running = true
				break
			}
		}
	}
	a.stretch.mu.Lock()
	if !running {
		a.stretch.suppressed = false
	}
	active, auto, pending, suppressed := a.stretch.active, a.stretch.auto, a.stretch.pending, a.stretch.suppressed
	a.stretch.mu.Unlock()
	switch {
	case running && !active && !suppressed:
		r := cfg.Resolution
		if err := a.applyStretch(r.Display, r.Width, r.Height, r.Refresh, r.Scaling, true); err != nil {
			a.toast("error", "toast.resFailed", map[string]any{"error": err.Error()})
		}
	case !running && active && auto && pending == 0:
		_ = a.revertStretch(true)
	}
}

// RunningApps lists programs with a visible window (to pick games from).
func (a *App) RunningApps() []string { return watch.VisibleApps(a.selfPID) }

// ---------------------------------------------------------------------------
// Tray menu
// ---------------------------------------------------------------------------

const (
	cmdOpen = 1 + iota
	cmdTuning
	cmdCrosshair
	cmdStretch
	cmdQuit
	cmdPresetBase = 100
)

var trayStrings = map[string]map[string]string{
	"en": {
		"open": "Open Reka Shade", "tuning": "Display tuning", "crosshair": "Crosshair",
		"stretch": "Stretched resolution", "presets": "Presets", "quit": "Quit",
		"hiddenTitle": "Reka Shade is still running",
		"hidden":      "Your tuning and crosshair stay active. Click the tray icon to come back.",
	},
	"fr": {
		"open": "Ouvrir Reka Shade", "tuning": "Réglage de l'affichage", "crosshair": "Viseur",
		"stretch": "Résolution étirée", "presets": "Préréglages", "quit": "Quitter",
		"hiddenTitle": "Reka Shade tourne toujours",
		"hidden":      "Vos réglages et votre viseur restent actifs. Cliquez sur l'icône pour revenir.",
	},
}

func tr(lang, key string) string {
	if m, ok := trayStrings[lang]; ok {
		if s, ok := m[key]; ok {
			return s
		}
	}
	return trayStrings["en"][key]
}

func (a *App) trayMenu() []host.MenuItem {
	a.mu.Lock()
	cfg := a.cfg
	lang := a.langLocked()
	a.mu.Unlock()
	dev := a.displayUnderForeground()
	cur := cfg.DisplayTuningFor(dev).PresetID
	var presets []host.MenuItem
	for i, p := range cfg.Presets {
		if i >= 40 {
			break
		}
		presets = append(presets, host.MenuItem{ID: cmdPresetBase + i, Label: strings.ReplaceAll(p.Name, "&", "&&"), Radio: true, Checked: p.ID == cur && cfg.Tuning.Enabled})
	}
	return []host.MenuItem{
		{ID: cmdOpen, Label: tr(lang, "open")},
		{Separator: true},
		{ID: cmdTuning, Label: tr(lang, "tuning"), Checked: cfg.Tuning.Enabled},
		{ID: cmdCrosshair, Label: tr(lang, "crosshair"), Checked: cfg.Overlay.Enabled},
		{ID: cmdStretch, Label: tr(lang, "stretch"), Checked: a.stretchActive()},
		{Label: tr(lang, "presets"), Children: presets},
		{Separator: true},
		{ID: cmdQuit, Label: tr(lang, "quit")},
	}
}

func (a *App) onMenuCommand(id int) {
	switch {
	case id == cmdOpen:
		a.ShowWindow()
	case id == cmdTuning:
		a.runAction("toggleTuning")
	case id == cmdCrosshair:
		a.runAction("toggleCrosshair")
	case id == cmdStretch:
		a.toggleStretch()
	case id == cmdQuit:
		a.QuitApp()
	case id >= cmdPresetBase:
		idx := id - cmdPresetBase
		dev := a.displayUnderForeground()
		a.updateConfig(func(c *config.Config) {
			if idx < len(c.Presets) && dev != "" {
				c.ApplyPreset(dev, c.Presets[idx])
				c.Tuning.Enabled = true
			}
		})
	}
}
