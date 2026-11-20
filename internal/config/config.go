// Package config defines Reka Shade's settings file and its defaults.
//
// The same JSON shape is used by the UI (frontend/src/lib/types.ts). Fields
// only the UI cares about (theme, crosshair design, profile...) are kept as
// raw JSON so the Go side round-trips them untouched.
package config

import (
	_ "embed"
	"encoding/json"
	"math"
	"strings"
)

// SchemaVersion is bumped when the file format changes incompatibly.
const SchemaVersion = 1

//go:embed defaults.json
var defaultsJSON []byte

// DisplayParams are the per-display tuning values (hardware gamma ramp +
// GPU vibrance).
type DisplayParams struct {
	Vibrance    float64 `json:"vibrance"`    // %, 0..200, 100 = neutral
	Brightness  float64 `json:"brightness"`  // %, 50..150
	Contrast    float64 `json:"contrast"`    // %, 50..150
	Gamma       float64 `json:"gamma"`       // 0.4..2.5, 1 = neutral
	Temperature float64 `json:"temperature"` // Kelvin 3500..10000, 6500 = neutral
	Red         float64 `json:"red"`         // %, 0..200
	Green       float64 `json:"green"`
	Blue        float64 `json:"blue"`
}

// ColorParams is a full look: global hue/saturation + per-display values.
type ColorParams struct {
	Hue        float64 `json:"hue"`        // degrees, -180..180
	Saturation float64 `json:"saturation"` // %, 0..300, 100 = neutral
	DisplayParams
}

// Preset is a named look.
type Preset struct {
	ID      string      `json:"id"`
	Name    string      `json:"name"`
	Builtin bool        `json:"builtin,omitempty"`
	Tags    []string    `json:"tags,omitempty"`
	Params  ColorParams `json:"params"`
}

// DisplayTuning is what is applied to one display.
type DisplayTuning struct {
	PresetID string `json:"presetId"`
	DisplayParams
}

// Tuning is the display-colour section. Hue and saturation are global
// because the Windows colour engine applies one matrix to every screen.
type Tuning struct {
	Enabled    bool                     `json:"enabled"`
	Hue        float64                  `json:"hue"`
	Saturation float64                  `json:"saturation"`
	Displays   map[string]DisplayTuning `json:"displays"`
}

// Overlay is where and when the crosshair/reticle bitmap is shown.
type Overlay struct {
	Enabled         bool   `json:"enabled"`
	Mode            string `json:"mode"`    // "crosshair" | "reticle"
	Display         string `json:"display"` // GDI device name, "" = primary
	OffsetX         int    `json:"offsetX"`
	OffsetY         int    `json:"offsetY"`
	FollowWindow    bool   `json:"followWindow"`
	OnlyFullscreen  bool   `json:"onlyFullscreen"`
	HideFromCapture bool   `json:"hideFromCapture"`
}

// Bind maps a key combination to an action.
type Bind struct {
	Action string `json:"action"`
	Key    int    `json:"key"` // Windows virtual-key code, 0 = unbound
	Ctrl   bool   `json:"ctrl"`
	Alt    bool   `json:"alt"`
	Shift  bool   `json:"shift"`
}

// Resolution is the stretched-resolution setup.
type Resolution struct {
	Display      string   `json:"display"`
	Width        int      `json:"width"`
	Height       int      `json:"height"`
	Refresh      int      `json:"refresh"`
	Scaling      string   `json:"scaling"`   // "stretch" | "aspect" | "center"
	Confirmed    bool     `json:"confirmed"` // user kept this mode once
	RevertOnExit bool     `json:"revertOnExit"`
	AutoApply    bool     `json:"autoApply"`
	AutoGames    []string `json:"autoGames"`
}

// GameRule switches to a preset while a given game is in focus.
type GameRule struct {
	Exe      string `json:"exe"`
	PresetID string `json:"presetId"`
	Enabled  bool   `json:"enabled"`
}

// Settings are app-level preferences.
type Settings struct {
	StartWithWindows bool `json:"startWithWindows"`
	StartMinimized   bool `json:"startMinimized"`
	CloseToTray      bool `json:"closeToTray"`
	AutoUpdate       bool `json:"autoUpdate"`
	AutoInstall      bool `json:"autoInstall"`
	ApplyOnStartup   bool `json:"applyOnStartup"`
	RestoreOnExit    bool `json:"restoreOnExit"`
	GPUFeatures      bool `json:"gpuFeatures"`
}

// Config is the whole settings file.
type Config struct {
	Version    int        `json:"version"`
	Language   string     `json:"language"` // "auto" | "en" | "fr"
	Tuning     Tuning     `json:"tuning"`
	Presets    []Preset   `json:"presets"`
	Overlay    Overlay    `json:"overlay"`
	Binds      []Bind     `json:"binds"`
	Resolution Resolution `json:"resolution"`
	Rules      []GameRule `json:"rules"`
	Settings   Settings   `json:"settings"`

	// UI-owned sections, preserved verbatim.
	Crosshair json.RawMessage `json:"crosshair,omitempty"`
	Reticle   json.RawMessage `json:"reticle,omitempty"`
	Theme     json.RawMessage `json:"theme,omitempty"`
	Profile   json.RawMessage `json:"profile,omitempty"`
	UI        json.RawMessage `json:"ui,omitempty"`
}

type defaultsFile struct {
	Presets []Preset `json:"presets"`
	Binds   []Bind   `json:"binds"`
}

func loadDefaults() defaultsFile {
	var d defaultsFile
	if err := json.Unmarshal(defaultsJSON, &d); err != nil {
		panic("config: invalid embedded defaults.json: " + err.Error())
	}
	return d
}

// BuiltinPresets returns the presets shipped with the app.
func BuiltinPresets() []Preset { return loadDefaults().Presets }

// DefaultBinds returns the default key binds.
func DefaultBinds() []Bind { return loadDefaults().Binds }

// NeutralDisplay returns per-display values that change nothing.
func NeutralDisplay() DisplayParams {
	return DisplayParams{Vibrance: 100, Brightness: 100, Contrast: 100, Gamma: 1, Temperature: 6500, Red: 100, Green: 100, Blue: 100}
}

// Default returns a fresh configuration.
func Default() Config {
	d := loadDefaults()
	return Config{
		Version:  SchemaVersion,
		Language: "auto",
		Tuning: Tuning{
			Enabled: false, Hue: 0, Saturation: 100,
			Displays: map[string]DisplayTuning{},
		},
		Presets: d.Presets,
		Overlay: Overlay{Enabled: false, Mode: "crosshair"},
		Binds:   d.Binds,
		Resolution: Resolution{
			Width: 1440, Height: 1080, Scaling: "stretch", RevertOnExit: true, AutoGames: []string{},
		},
		Rules: []GameRule{},
		Settings: Settings{
			StartWithWindows: false, StartMinimized: true, CloseToTray: true,
			AutoUpdate: true, AutoInstall: false, ApplyOnStartup: true,
			RestoreOnExit: true, GPUFeatures: true,
		},
	}
}

func clamp(v, lo, hi, def float64) float64 {
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return def
	}
	return math.Max(lo, math.Min(hi, v))
}

// SanitizeDisplay clamps per-display values into their legal ranges.
func SanitizeDisplay(p DisplayParams) DisplayParams {
	return DisplayParams{
		Vibrance:    clamp(p.Vibrance, 0, 200, 100),
		Brightness:  clamp(p.Brightness, 50, 150, 100),
		Contrast:    clamp(p.Contrast, 50, 150, 100),
		Gamma:       clamp(p.Gamma, 0.4, 2.5, 1),
		Temperature: clamp(p.Temperature, 3500, 10000, 6500),
		Red:         clamp(p.Red, 0, 200, 100),
		Green:       clamp(p.Green, 0, 200, 100),
		Blue:        clamp(p.Blue, 0, 200, 100),
	}
}

// Sanitize repairs anything out of range so a hand-edited or corrupted file
// can never push extreme values to the display.
func (c *Config) Sanitize() {
	def := Default()
	c.Version = SchemaVersion
	switch c.Language {
	case "auto", "en", "fr":
	default:
		c.Language = "auto"
	}
	c.Tuning.Hue = clamp(c.Tuning.Hue, -180, 180, 0)
	c.Tuning.Saturation = clamp(c.Tuning.Saturation, 0, 300, 100)
	if c.Tuning.Displays == nil {
		c.Tuning.Displays = map[string]DisplayTuning{}
	}
	for k, d := range c.Tuning.Displays {
		d.DisplayParams = SanitizeDisplay(d.DisplayParams)
		c.Tuning.Displays[k] = d
	}
	if len(c.Presets) == 0 {
		c.Presets = def.Presets
	}
	seen := map[string]bool{}
	presets := c.Presets[:0]
	for _, p := range c.Presets {
		if p.ID == "" || seen[p.ID] {
			continue
		}
		seen[p.ID] = true
		p.Params.Hue = clamp(p.Params.Hue, -180, 180, 0)
		p.Params.Saturation = clamp(p.Params.Saturation, 0, 300, 100)
		p.Params.DisplayParams = SanitizeDisplay(p.Params.DisplayParams)
		if strings.TrimSpace(p.Name) == "" {
			p.Name = p.ID
		}
		presets = append(presets, p)
	}
	c.Presets = presets
	if c.Overlay.Mode != "reticle" {
		c.Overlay.Mode = "crosshair"
	}
	c.Overlay.OffsetX = int(clamp(float64(c.Overlay.OffsetX), -2000, 2000, 0))
	c.Overlay.OffsetY = int(clamp(float64(c.Overlay.OffsetY), -2000, 2000, 0))
	if c.Binds == nil {
		c.Binds = def.Binds
	}
	// Make sure every known action has an entry (older files may lack some).
	have := map[string]bool{}
	for _, b := range c.Binds {
		have[b.Action] = true
	}
	for _, b := range def.Binds {
		if !have[b.Action] {
			c.Binds = append(c.Binds, Bind{Action: b.Action})
		}
	}
	for i := range c.Binds {
		if c.Binds[i].Key < 0 || c.Binds[i].Key > 0xFE {
			c.Binds[i].Key = 0
		}
	}
	switch c.Resolution.Scaling {
	case "stretch", "aspect", "center":
	default:
		c.Resolution.Scaling = "stretch"
	}
	if c.Resolution.Width < 320 || c.Resolution.Width > 16384 || c.Resolution.Height < 200 || c.Resolution.Height > 16384 {
		c.Resolution.Width, c.Resolution.Height, c.Resolution.Refresh = 1440, 1080, 0
		c.Resolution.Confirmed = false
	}
	if c.Resolution.Refresh < 0 || c.Resolution.Refresh > 1000 {
		c.Resolution.Refresh = 0
	}
	if c.Resolution.AutoGames == nil {
		c.Resolution.AutoGames = []string{}
	}
	for i, g := range c.Resolution.AutoGames {
		c.Resolution.AutoGames[i] = NormalizeExe(g)
	}
	if c.Rules == nil {
		c.Rules = []GameRule{}
	}
	for i := range c.Rules {
		c.Rules[i].Exe = NormalizeExe(c.Rules[i].Exe)
	}
}

// NormalizeExe lower-cases an executable name and strips any folder.
func NormalizeExe(s string) string {
	s = strings.TrimSpace(strings.ToLower(s))
	if i := strings.LastIndexAny(s, `\/`); i >= 0 {
		s = s[i+1:]
	}
	if s != "" && !strings.HasSuffix(s, ".exe") {
		s += ".exe"
	}
	return s
}

// PresetByID finds a preset.
func (c *Config) PresetByID(id string) (Preset, bool) {
	for _, p := range c.Presets {
		if p.ID == id {
			return p, true
		}
	}
	return Preset{}, false
}

// DisplayTuningFor returns the tuning of a display (neutral if unset).
func (c *Config) DisplayTuningFor(device string) DisplayTuning {
	if d, ok := c.Tuning.Displays[device]; ok {
		return d
	}
	return DisplayTuning{PresetID: "stock", DisplayParams: NeutralDisplay()}
}

// ApplyPreset puts a preset on a display (and sets the global hue/saturation).
func (c *Config) ApplyPreset(device string, p Preset) {
	if c.Tuning.Displays == nil {
		c.Tuning.Displays = map[string]DisplayTuning{}
	}
	c.Tuning.Hue = p.Params.Hue
	c.Tuning.Saturation = p.Params.Saturation
	c.Tuning.Displays[device] = DisplayTuning{PresetID: p.ID, DisplayParams: SanitizeDisplay(p.Params.DisplayParams)}
}

// Clone deep-copies the configuration.
func (c Config) Clone() Config {
	b, _ := json.Marshal(c)
	var out Config
	_ = json.Unmarshal(b, &out)
	return out
}
