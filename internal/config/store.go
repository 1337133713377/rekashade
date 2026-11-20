package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Store persists the configuration and the app's private state on disk:
//
//	%APPDATA%\RekaShade\settings.json  – user settings (Config)
//	%APPDATA%\RekaShade\state.json     – things only the app writes (State)
//	%APPDATA%\RekaShade\overlay.png    – last rendered crosshair/reticle
type Store struct {
	Dir string

	mu       sync.Mutex
	timer    *time.Timer
	pending  *Config
	lastErr  error
	FirstRun bool
}

// NewStore creates the settings directory if needed.
func NewStore() (*Store, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}
	dir := filepath.Join(base, "RekaShade")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	return &Store{Dir: dir}, nil
}

func (s *Store) settingsPath() string { return filepath.Join(s.Dir, "settings.json") }
func (s *Store) statePath() string    { return filepath.Join(s.Dir, "state.json") }

// OverlayPath is where the last overlay bitmap is cached.
func (s *Store) OverlayPath() string { return filepath.Join(s.Dir, "overlay.png") }

// Load reads settings.json, falling back to defaults when it is missing or
// unreadable (a corrupt file is kept aside as settings.corrupt.json).
func (s *Store) Load() Config {
	b, err := os.ReadFile(s.settingsPath())
	if errors.Is(err, os.ErrNotExist) {
		s.FirstRun = true
		return Default()
	}
	cfg := Default()
	if err == nil {
		err = json.Unmarshal(b, &cfg)
	}
	if err != nil {
		_ = os.WriteFile(filepath.Join(s.Dir, "settings.corrupt.json"), b, 0o644)
		cfg = Default()
	}
	cfg.Sanitize()
	return cfg
}

// Parse validates an imported settings file.
func Parse(b []byte) (Config, error) {
	cfg := Default()
	if err := json.Unmarshal(b, &cfg); err != nil {
		return Config{}, fmt.Errorf("not a Reka Shade settings file: %w", err)
	}
	cfg.Sanitize()
	return cfg, nil
}

// Save writes the settings after a short debounce (sliders fire often).
func (s *Store) Save(cfg Config) {
	s.mu.Lock()
	defer s.mu.Unlock()
	c := cfg
	s.pending = &c
	if s.timer == nil {
		s.timer = time.AfterFunc(400*time.Millisecond, s.flushPending)
	} else {
		s.timer.Reset(400 * time.Millisecond)
	}
}

func (s *Store) flushPending() {
	s.mu.Lock()
	p := s.pending
	s.pending = nil
	s.mu.Unlock()
	if p != nil {
		if err := writeJSONAtomic(s.settingsPath(), p); err != nil {
			s.mu.Lock()
			s.lastErr = err
			s.mu.Unlock()
		}
	}
}

// Flush writes any pending settings immediately.
func (s *Store) Flush() error {
	s.mu.Lock()
	if s.timer != nil {
		s.timer.Stop()
	}
	s.mu.Unlock()
	s.flushPending()
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.lastErr
}

// LoadState reads state.json.
func (s *Store) LoadState() State {
	st := State{Optimizer: map[string]json.RawMessage{}, Vibrance: map[string]int{}}
	b, err := os.ReadFile(s.statePath())
	if err == nil {
		_ = json.Unmarshal(b, &st)
	}
	if st.Optimizer == nil {
		st.Optimizer = map[string]json.RawMessage{}
	}
	if st.Vibrance == nil {
		st.Vibrance = map[string]int{}
	}
	return st
}

// SaveState writes state.json immediately.
func (s *Store) SaveState(st State) error { return writeJSONAtomic(s.statePath(), st) }

// MarkerPath is a file that exists only while GPU drivers are being probed;
// if it is still there on the next start, the previous probe crashed.
func (s *Store) MarkerPath() string { return filepath.Join(s.Dir, "gpu-probe.lock") }

func writeJSONAtomic(path string, v any) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// StretchState remembers what to go back to after a stretched resolution.
type StretchState struct {
	Active      bool   `json:"active"`
	Device      string `json:"device"`
	PrevScaling uint32 `json:"prevScaling"`
}

// State is data only the app writes: undo information and crash recovery.
type State struct {
	Optimizer      map[string]json.RawMessage `json:"optimizer"`
	Vibrance       map[string]int             `json:"vibrance"` // original driver levels
	Stretch        StretchState               `json:"stretch"`
	TuningWasOn    bool                       `json:"tuningWasOn"`
	LastVersion    string                     `json:"lastVersion"`
	GPUProbeFailed bool                       `json:"gpuProbeFailed"`
}
