package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"golang.org/x/sys/windows"

	"github.com/1337133713377/rekashade/internal/config"
	"github.com/1337133713377/rekashade/internal/optimizer"
	"github.com/1337133713377/rekashade/internal/updater"
)

// ---------------------------------------------------------------------------
// Updates
// ---------------------------------------------------------------------------

type updateRuntime struct {
	mu   sync.Mutex
	last *updater.Check
	busy bool
}

// CheckForUpdates asks GitHub for a newer release.
func (a *App) CheckForUpdates() (*updater.Check, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()
	res, err := updater.CheckLatest(ctx, version)
	if err != nil {
		return nil, err
	}
	a.upd.mu.Lock()
	a.upd.last = &res
	a.upd.mu.Unlock()
	a.emitStatus()
	return &res, nil
}

func (a *App) autoUpdateCheck() {
	res, err := a.CheckForUpdates()
	if err != nil || res == nil || !res.Available {
		return
	}
	cfg := a.config()
	if cfg.Settings.AutoInstall {
		if err := a.InstallUpdate(); err != nil {
			a.toast("error", "toast.updateFailed", map[string]any{"error": err.Error()})
		}
		return
	}
	a.toast("info", "toast.updateAvailable", map[string]any{"version": res.Latest.Version})
	if a.host != nil {
		a.mu.Lock()
		lang := a.langLocked()
		a.mu.Unlock()
		title := "Reka Shade " + res.Latest.Version
		msg := "A new version is ready. Open Reka Shade › About to install it."
		if lang == "fr" {
			msg = "Une nouvelle version est prête. Ouvrez Reka Shade › À propos pour l'installer."
		}
		a.host.Notify(title, msg)
	}
}

// InstallUpdate downloads, verifies and installs the latest release, then
// restarts the app.
func (a *App) InstallUpdate() error {
	a.upd.mu.Lock()
	if a.upd.busy {
		a.upd.mu.Unlock()
		return errors.New("an update is already in progress")
	}
	last := a.upd.last
	a.upd.busy = true
	a.upd.mu.Unlock()
	defer func() {
		a.upd.mu.Lock()
		a.upd.busy = false
		a.upd.mu.Unlock()
		a.emitStatus()
	}()
	if last == nil || !last.Available {
		res, err := a.CheckForUpdates()
		if err != nil {
			return err
		}
		if !res.Available {
			return errors.New("you already have the latest version")
		}
		last = res
	}
	a.emitStatus()
	dir := filepath.Join(os.TempDir(), "RekaShadeUpdate")
	lastEmit := time.Time{}
	path, err := updater.Download(context.Background(), last.Latest, dir, func(done, total int64) {
		if time.Since(lastEmit) > 120*time.Millisecond || done == total {
			lastEmit = time.Now()
			a.emit("rs:update-progress", map[string]int64{"done": done, "total": total})
		}
	})
	if err != nil {
		return err
	}
	if err := updater.Install(path, last.Latest.Version); err != nil {
		return err
	}
	a.QuitApp()
	return nil
}

// ---------------------------------------------------------------------------
// Optimizer
// ---------------------------------------------------------------------------

func (a *App) backups() optimizer.Backups {
	a.mu.Lock()
	defer a.mu.Unlock()
	b := optimizer.Backups{}
	for k, v := range a.state.Optimizer {
		b[k] = v
	}
	return b
}

func (a *App) saveBackups(b optimizer.Backups) {
	a.mu.Lock()
	a.state.Optimizer = map[string]json.RawMessage(b)
	a.saveStateLocked()
	a.mu.Unlock()
}

// OptimizerList reads the real state of every tweak from Windows.
func (a *App) OptimizerList() []optimizer.Status { return optimizer.List(a.backups()) }

// OptimizerSet applies (on) or undoes (off) one tweak.
func (a *App) OptimizerSet(id string, on bool) (optimizer.Status, error) {
	b := a.backups()
	s, err := optimizer.Set(id, on, b)
	a.saveBackups(b)
	return s, err
}

// OptimizerAll applies or undoes every tweak.
func (a *App) OptimizerAll(on bool) []optimizer.Status {
	b := a.backups()
	for _, id := range optimizer.IDs() {
		_, _ = optimizer.Set(id, on, b)
	}
	a.saveBackups(b)
	return optimizer.List(b)
}

// CleanTemp removes old files from the user's temp folder.
func (a *App) CleanTemp() optimizer.CleanResult { return optimizer.CleanTemp() }

// ---------------------------------------------------------------------------
// Settings import / export & diagnostics
// ---------------------------------------------------------------------------

// ExportSettings saves the settings to a file the user picks.
func (a *App) ExportSettings() (string, error) {
	path, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "Export Reka Shade settings",
		DefaultFilename: "RekaShade-settings.json",
		Filters:         []runtime.FileFilter{{DisplayName: "Reka Shade settings (*.json)", Pattern: "*.json"}},
	})
	if err != nil || path == "" {
		return "", err
	}
	b, err := json.MarshalIndent(a.config(), "", "  ")
	if err != nil {
		return "", err
	}
	return path, os.WriteFile(path, b, 0o644)
}

// ImportSettings reads a settings file; the UI then saves it with SaveConfig.
func (a *App) ImportSettings() (*config.Config, error) {
	path, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title:   "Import Reka Shade settings",
		Filters: []runtime.FileFilter{{DisplayName: "Reka Shade settings (*.json)", Pattern: "*.json"}},
	})
	if err != nil || path == "" {
		return nil, err
	}
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if info.Size() > 8<<20 {
		return nil, errors.New("file too large")
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	cfg, err := config.Parse(b)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

// OpenDataFolder opens the settings folder in Explorer.
func (a *App) OpenDataFolder() error {
	if a.store == nil {
		return errors.New("no settings folder")
	}
	return windows.ShellExecute(0, windows.StringToUTF16Ptr("open"), windows.StringToUTF16Ptr(a.store.Dir), nil, nil, 1)
}

// Diagnostics returns a plain-text report for bug reports.
func (a *App) Diagnostics() string {
	v := windows.RtlGetVersion()
	st := a.GetStatus()
	b, _ := json.MarshalIndent(map[string]any{
		"version":  version,
		"windows":  fmt.Sprintf("%d.%d.%d", v.MajorVersion, v.MinorVersion, v.BuildNumber),
		"displays": a.GetDisplays(),
		"status":   st,
	}, "", "  ")
	return string(b)
}
