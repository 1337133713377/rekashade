package main

import (
	"context"
	"os"
	"reflect"
	"sync"
	"sync/atomic"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"

	"github.com/1337133713377/rekashade/internal/config"
	"github.com/1337133713377/rekashade/internal/display"
	"github.com/1337133713377/rekashade/internal/host"
	"github.com/1337133713377/rekashade/internal/startup"
	"github.com/1337133713377/rekashade/internal/updater"
	"github.com/1337133713377/rekashade/internal/win"
)

const (
	mainWindowClass = "RekaShadeWindow"
	minWindowWidth  = 940
	minWindowHeight = 600
)

// App is bound to the UI: every exported method is callable from JavaScript
// as window.go.main.App.<Method>.
type App struct {
	ctx   context.Context
	args  launchArgs
	store *config.Store

	mu       sync.Mutex
	cfg      config.Config
	state    config.State
	displays []display.Info

	applyMu   sync.Mutex    // serialises pushes to the display stack
	applyKick chan struct{} // coalesces tuning updates while sliders move
	host      *host.Host
	engine    *display.Engine

	ov      overlayRuntime
	stretch stretchRuntime
	upd     updateRuntime
	watch   watchRuntime

	selfPID      uint32
	ready        chan struct{} // closed once startup() has finished
	stopCh       chan struct{}
	quitting     atomic.Bool
	shutdownOnce sync.Once
	trayHinted   atomic.Bool
	bindsPaused  atomic.Bool
}

// NewApp loads settings; native work starts in startup().
func NewApp(args launchArgs) *App {
	a := &App{args: args, stopCh: make(chan struct{}), ready: make(chan struct{}), applyKick: make(chan struct{}, 1), selfPID: uint32(os.Getpid())}
	go a.tuningLoop()
	store, err := config.NewStore()
	if err == nil {
		a.store = store
		a.cfg = store.Load()
		a.state = store.LoadState()
	} else {
		a.cfg = config.Default()
		a.state = config.State{}
	}
	if !a.cfg.Settings.ApplyOnStartup {
		a.cfg.Tuning.Enabled = false
	}
	return a
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	// Wails runs OnStartup in parallel with the page loading: the UI's first
	// calls wait on this so they never see an empty display list.
	defer func() {
		close(a.ready)
		a.emit("rs:displays", a.GetDisplays())
		a.emitStatus()
	}()
	go updater.Cleanup()

	h, err := host.Start(host.Callbacks{
		TrayClick:     func() { go a.ShowWindow() },
		TrayMenu:      a.trayMenu,
		MenuCommand:   a.onMenuCommand,
		DisplayChange: a.onDisplayChange,
		Resume:        a.onResume,
		CloseRequest:  a.QuitApp,
		EndSession:    func() { a.cleanup(true) },
		FollowTick:    a.onFollowTick,
	})
	if err == nil {
		a.host = h
	}
	a.engine = display.NewEngine(h)

	a.mu.Lock()
	cfg := a.cfg
	a.mu.Unlock()

	a.probeGPU(cfg)
	a.refreshDisplays()

	ids := a.displayIDs()
	a.engine.CaptureBaselines(ids, a.state.TuningWasOn)

	// A previous session died while a stretched resolution was active.
	if a.state.Stretch.Active && a.state.Stretch.Device != "" {
		_ = display.RestoreMode(a.state.Stretch.Device, a.state.Stretch.PrevScaling)
		a.state.Stretch = config.StretchState{}
		a.refreshDisplays()
	}

	a.applyTuning()
	a.loadOverlayImage() // positions first, then shows (no flash at 0,0)

	if a.host != nil {
		a.host.ShowTray("Reka Shade")
	}
	_ = startup.Set(cfg.Settings.StartWithWindows, cfg.Settings.StartMinimized)

	a.mu.Lock()
	a.state.LastVersion = version
	a.saveStateLocked()
	a.mu.Unlock()

	go a.keyLoop()
	go a.watchLoop()
	if cfg.Settings.AutoUpdate {
		go func() {
			select {
			case <-time.After(4 * time.Second):
				a.autoUpdateCheck()
			case <-a.stopCh:
			}
		}()
	}
}

func (a *App) domReady(ctx context.Context) { a.fitWindowToScreen(ctx) }

// fitWindowToScreen shrinks the window when the screen (with Windows display
// scaling) is smaller than the default size, so nothing ends up off-screen;
// pages scroll instead.
func (a *App) fitWindowToScreen(ctx context.Context) {
	screens, err := runtime.ScreenGetAll(ctx)
	if err != nil || len(screens) == 0 {
		return
	}
	scr := screens[0]
	for _, sc := range screens {
		if sc.IsCurrent {
			scr = sc
			break
		}
	}
	sw, sh := scr.Size.Width, scr.Size.Height
	if sw <= 0 || sh <= 0 {
		return
	}
	w, h := runtime.WindowGetSize(ctx)
	maxW, maxH := sw-40, sh-72 // leave room for the taskbar
	if w <= maxW && h <= maxH {
		return
	}
	if w > maxW {
		w = max(maxW, minWindowWidth)
	}
	if h > maxH {
		h = max(maxH, minWindowHeight)
	}
	runtime.WindowSetSize(ctx, w, h)
	runtime.WindowCenter(ctx)
}

func (a *App) beforeClose(ctx context.Context) bool {
	if a.quitting.Load() {
		return false
	}
	a.mu.Lock()
	toTray := a.cfg.Settings.CloseToTray && a.host != nil
	lang := a.langLocked()
	a.mu.Unlock()
	if toTray {
		runtime.WindowHide(ctx)
		if !a.trayHinted.Swap(true) {
			a.host.Notify(tr(lang, "hiddenTitle"), tr(lang, "hidden"))
		}
		return true
	}
	a.quitting.Store(true)
	return false
}

func (a *App) shutdown(ctx context.Context) { a.cleanup(false) }

// cleanup restores the system. fromHost is true when called on the host
// thread (Windows session ending), where the host must not be stopped.
func (a *App) cleanup(fromHost bool) {
	a.shutdownOnce.Do(func() {
		close(a.stopCh)
		a.mu.Lock()
		cfg := a.cfg
		a.mu.Unlock()

		a.stopStretchCountdown()
		if a.stretchActive() && (cfg.Resolution.RevertOnExit || fromHost) {
			_ = a.revertStretch(false)
		}
		if a.engine != nil && fromHost {
			// Session ending: restore what outlives the process (gamma,
			// vibrance) without waiting on the host thread we are running on.
			if cfg.Settings.RestoreOnExit {
				a.engine.ResetHardwareBestEffort(1500 * time.Millisecond)
			}
		} else if a.engine != nil {
			if cfg.Settings.RestoreOnExit {
				a.engine.Close()
				a.mu.Lock()
				a.state.TuningWasOn = false
				a.state.Vibrance = map[string]int{}
				a.mu.Unlock()
			}
		}
		if a.host != nil {
			a.host.SetOverlayVisible(false)
			if !fromHost {
				a.host.Stop()
			}
		}
		a.ov.flush(a)
		if a.store != nil {
			a.store.Save(cfg)
			_ = a.store.Flush()
		}
		a.mu.Lock()
		a.saveStateLocked()
		a.mu.Unlock()
	})
}

func (a *App) saveStateLocked() {
	if a.store != nil {
		_ = a.store.SaveState(a.state)
	}
}

func (a *App) config() config.Config {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.cfg
}

func (a *App) emit(name string, data any) {
	if a.ctx != nil {
		runtime.EventsEmit(a.ctx, name, data)
	}
}

// toast asks the UI to show a localised message.
func (a *App) toast(kind, key string, params map[string]any) {
	a.emit("rs:toast", map[string]any{"kind": kind, "key": key, "params": params})
}

func (a *App) emitConfig() { a.emit("rs:config", a.config()) }

func (a *App) emitStatus() { a.emit("rs:status", a.GetStatus()) }

// updateConfig mutates the config from the Go side (hotkeys, tray, rules),
// persists it and tells the UI.
func (a *App) updateConfig(f func(c *config.Config)) config.Config {
	a.mu.Lock()
	old := a.cfg
	next := a.cfg.Clone()
	f(&next)
	next.Sanitize()
	a.cfg = next
	if a.store != nil {
		a.store.Save(next)
	}
	a.mu.Unlock()
	a.applyChanges(old, next)
	a.emit("rs:config", next)
	return next
}

// tuningLoop applies the latest tuning whenever it changes. Slider drags
// send many updates per second; if the display is slow to accept a gamma
// ramp, intermediate values are skipped instead of piling up.
func (a *App) tuningLoop() {
	for {
		select {
		case <-a.stopCh:
			return
		case <-a.applyKick:
			safely(a.applyTuning)
			a.emitStatus()
		}
	}
}

func (a *App) scheduleTuning() {
	select {
	case a.applyKick <- struct{}{}:
	default: // an update is already pending; it will pick up the latest values
	}
}

func (a *App) applyChanges(old, next config.Config) {
	if !reflect.DeepEqual(old.Tuning, next.Tuning) || !reflect.DeepEqual(old.Presets, next.Presets) || !reflect.DeepEqual(old.Rules, next.Rules) {
		a.scheduleTuning()
	}
	if old.Overlay != next.Overlay {
		a.applyOverlay()
	}
	if old.Settings.StartWithWindows != next.Settings.StartWithWindows || old.Settings.StartMinimized != next.Settings.StartMinimized {
		_ = startup.Set(next.Settings.StartWithWindows, next.Settings.StartMinimized)
		invalidateSystemFlags()
	}
	if old.Settings.GPUFeatures != next.Settings.GPUFeatures && next.Settings.GPUFeatures {
		go func() {
			a.mu.Lock()
			a.state.GPUProbeFailed = false
			a.mu.Unlock()
			a.probeGPU(next)
			a.refreshDisplays()
			a.applyTuning()
			a.emit("rs:displays", a.GetDisplays())
		}()
	}
	a.emitStatus()
}

// ---------------------------------------------------------------------------
// Bindings: bootstrap & config
// ---------------------------------------------------------------------------

// Bootstrap is the first call the UI makes.
type Bootstrap struct {
	Version        string         `json:"version"`
	Config         config.Config  `json:"config"`
	Displays       []display.Info `json:"displays"`
	Status         AppStatus      `json:"status"`
	FirstRun       bool           `json:"firstRun"`
	Updated        bool           `json:"updated"`
	SystemLanguage string         `json:"systemLanguage"`
	WindowsBuild   int            `json:"windowsBuild"`
	DataDir        string         `json:"dataDir"`
	Mock           bool           `json:"mock"`
}

// waitReady blocks until startup() is done (bounded, so the UI never hangs).
func (a *App) waitReady() {
	select {
	case <-a.ready:
	case <-time.After(20 * time.Second):
	}
}

func (a *App) Bootstrap() Bootstrap {
	a.waitReady()
	a.mu.Lock()
	cfg := a.cfg
	a.mu.Unlock()
	b := Bootstrap{
		Version: version, Config: cfg, Displays: a.GetDisplays(), Status: a.GetStatus(),
		Updated: a.args.updated, SystemLanguage: systemLanguage(),
		WindowsBuild: int(windows.RtlGetVersion().BuildNumber),
	}
	if a.store != nil {
		b.FirstRun = a.store.FirstRun
		b.DataDir = a.store.Dir
	}
	return b
}

// SaveConfig receives the whole settings object from the UI.
func (a *App) SaveConfig(next config.Config) error {
	next.Sanitize()
	a.mu.Lock()
	old := a.cfg
	// "Confirmed" belongs to the Go side: it is only set by ConfirmResolution
	// and is lost as soon as the mode changes.
	r0, r1 := old.Resolution, next.Resolution
	next.Resolution.Confirmed = r0.Confirmed && r0.Display == r1.Display && r0.Width == r1.Width &&
		r0.Height == r1.Height && r0.Refresh == r1.Refresh && r0.Scaling == r1.Scaling
	a.cfg = next
	if a.store != nil {
		a.store.Save(next)
	}
	a.mu.Unlock()
	a.applyChanges(old, next)
	return nil
}

// AppStatus is everything live the UI displays.
type AppStatus struct {
	Display        display.Status `json:"display"`
	Overlay        OverlayStatus  `json:"overlay"`
	Stretch        StretchStatus  `json:"stretch"`
	Rule           string         `json:"rule"`
	Update         *updater.Check `json:"update,omitempty"`
	Updating       bool           `json:"updating"`
	ExtendedGamma  bool           `json:"extendedGamma"`
	Startup        bool           `json:"startup"`
	GPUProbeFailed bool           `json:"gpuProbeFailed"`
}

var sysFlags struct {
	sync.Mutex
	at            time.Time
	extendedGamma bool
	startup       bool
}

// systemFlags reads two registry values, cached for a few seconds because
// GetStatus runs on every settings change.
func systemFlags() (extGamma, startupOn bool) {
	sysFlags.Lock()
	defer sysFlags.Unlock()
	if time.Since(sysFlags.at) > 3*time.Second {
		sysFlags.extendedGamma = extendedGammaEnabled()
		sysFlags.startup = startup.Enabled()
		sysFlags.at = time.Now()
	}
	return sysFlags.extendedGamma, sysFlags.startup
}

func invalidateSystemFlags() {
	sysFlags.Lock()
	sysFlags.at = time.Time{}
	sysFlags.Unlock()
}

func (a *App) GetStatus() AppStatus {
	ext, run := systemFlags()
	s := AppStatus{
		Overlay:       a.overlayStatus(),
		Stretch:       a.stretchStatus(),
		Rule:          a.activeRule(),
		ExtendedGamma: ext,
		Startup:       run,
	}
	if a.engine != nil {
		s.Display = a.engine.Status()
	}
	a.upd.mu.Lock()
	s.Update, s.Updating = a.upd.last, a.upd.busy
	a.upd.mu.Unlock()
	a.mu.Lock()
	s.GPUProbeFailed = a.state.GPUProbeFailed
	a.mu.Unlock()
	return s
}

// ---------------------------------------------------------------------------
// Bindings: window
// ---------------------------------------------------------------------------

func (a *App) ShowWindow() {
	if a.ctx == nil {
		return
	}
	runtime.WindowShow(a.ctx)
	runtime.WindowUnminimise(a.ctx)
	if hwnd := win.FindWindow(mainWindowClass, ""); hwnd != 0 {
		win.ShowWindowAsync(hwnd, win.SW_RESTORE)
		win.SetForegroundWindow(hwnd)
	}
	a.emitStatus()
}

func (a *App) WindowMinimize() { runtime.WindowMinimise(a.ctx) }

// WindowClose is the title-bar close button: hide to tray or quit.
func (a *App) WindowClose() {
	if a.beforeClose(a.ctx) {
		return
	}
	a.QuitApp()
}

func (a *App) QuitApp() {
	a.quitting.Store(true)
	if a.ctx != nil {
		runtime.Quit(a.ctx)
	}
}

// OpenURL opens a link in the default browser (http/https only).
func (a *App) OpenURL(url string) {
	if len(url) > 8 && (url[:8] == "https://" || url[:7] == "http://") {
		runtime.BrowserOpenURL(a.ctx, url)
	}
}

// PauseBinds disables hotkeys while the UI records a new key.
func (a *App) PauseBinds(paused bool) { a.bindsPaused.Store(paused) }

func systemLanguage() string {
	if win.UserUILanguage() == 0x0C {
		return "fr"
	}
	return "en"
}

func (a *App) langLocked() string {
	if a.cfg.Language == "fr" || a.cfg.Language == "en" {
		return a.cfg.Language
	}
	return systemLanguage()
}

func extendedGammaEnabled() bool {
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Windows NT\CurrentVersion\ICM`, registry.QUERY_VALUE)
	if err != nil {
		return false
	}
	defer k.Close()
	v, _, err := k.GetIntegerValue("GdiIcmGammaRange")
	return err == nil && v >= 256
}
