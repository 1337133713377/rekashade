package config

import (
	"encoding/json"
	"testing"
)

func TestDefaultsLoad(t *testing.T) {
	c := Default()
	if len(c.Presets) < 5 {
		t.Fatalf("expected built-in presets, got %d", len(c.Presets))
	}
	if _, ok := c.PresetByID("stock"); !ok {
		t.Fatal("stock preset missing")
	}
	if len(c.Binds) == 0 {
		t.Fatal("default binds missing")
	}
}

func TestSanitizeClampsAndRepairs(t *testing.T) {
	raw := `{"language":"de","tuning":{"enabled":true,"hue":999,"saturation":-5,
	  "displays":{"\\\\.\\DISPLAY1":{"presetId":"x","vibrance":900,"gamma":0,"temperature":1,"red":100,"green":100,"blue":100,"brightness":100,"contrast":100}}},
	  "resolution":{"width":5,"height":5,"scaling":"zoom"},
	  "rules":[{"exe":"C:\\Games\\Valorant.EXE","presetId":"stock","enabled":true}],
	  "theme":{"accent":"violet"}}`
	c, err := Parse([]byte(raw))
	if err != nil {
		t.Fatal(err)
	}
	if c.Language != "auto" {
		t.Errorf("language not repaired: %q", c.Language)
	}
	if c.Tuning.Hue != 180 || c.Tuning.Saturation != 0 {
		t.Errorf("hue/sat not clamped: %v %v", c.Tuning.Hue, c.Tuning.Saturation)
	}
	d := c.Tuning.Displays[`\\.\DISPLAY1`]
	if d.Vibrance != 200 || d.Gamma != 0.4 || d.Temperature != 3500 {
		t.Errorf("display params not clamped: %+v", d)
	}
	if c.Resolution.Width != 1440 || c.Resolution.Scaling != "stretch" {
		t.Errorf("resolution not repaired: %+v", c.Resolution)
	}
	if c.Rules[0].Exe != "valorant.exe" {
		t.Errorf("exe not normalised: %q", c.Rules[0].Exe)
	}
	var theme map[string]string
	if err := json.Unmarshal(c.Theme, &theme); err != nil || theme["accent"] != "violet" {
		t.Errorf("UI-owned section not preserved: %s", c.Theme)
	}
}

func TestApplyPreset(t *testing.T) {
	c := Default()
	p, _ := c.PresetByID("ultra-vivid")
	c.ApplyPreset(`\\.\DISPLAY2`, p)
	if c.Tuning.Saturation != p.Params.Saturation {
		t.Fatal("global saturation not set from preset")
	}
	if got := c.DisplayTuningFor(`\\.\DISPLAY2`); got.PresetID != "ultra-vivid" || got.Vibrance != p.Params.Vibrance {
		t.Fatalf("display tuning not set: %+v", got)
	}
}

func TestNormalizeExe(t *testing.T) {
	cases := map[string]string{
		"FortniteClient-Win64-Shipping.exe": "fortniteclient-win64-shipping.exe",
		`D:\Steam\cs2`:                      "cs2.exe",
		"  ":                                "",
	}
	for in, want := range cases {
		if got := NormalizeExe(in); got != want {
			t.Errorf("NormalizeExe(%q) = %q, want %q", in, got, want)
		}
	}
}
