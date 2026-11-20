package colorx

import (
	"math"
	"testing"
)

func TestNeutralToneIsIdentity(t *testing.T) {
	r := BuildRamp(NeutralTone())
	if !r.IsIdentity() {
		id := IdentityRamp()
		t.Fatalf("neutral tone should give identity ramp, distance=%d", r.Distance(&id))
	}
}

func TestRampIsMonotonic(t *testing.T) {
	cases := []Tone{
		{Brightness: 150, Contrast: 150, Gamma: 2.5, Temperature: 3500, Red: 200, Green: 0, Blue: 50},
		{Brightness: 50, Contrast: 50, Gamma: 0.4, Temperature: 10000, Red: 0, Green: 100, Blue: 200},
	}
	for _, tone := range cases {
		r := BuildRamp(tone)
		for c := 0; c < 3; c++ {
			for i := 1; i < 256; i++ {
				if r[c][i] < r[c][i-1] {
					t.Fatalf("ramp not monotonic at c=%d i=%d", c, i)
				}
			}
		}
	}
}

func TestTemperatureNeutralAt6500(t *testing.T) {
	g := TemperatureGains(6500)
	if g != [3]float64{1, 1, 1} {
		t.Fatalf("6500K should be neutral, got %v", g)
	}
	warm := TemperatureGains(4000)
	if !(warm[0] > warm[2]) {
		t.Fatalf("4000K should be warm (more red than blue): %v", warm)
	}
	cool := TemperatureGains(9000)
	if !(cool[2] > cool[0]) {
		t.Fatalf("9000K should be cool (more blue than red): %v", cool)
	}
}

func TestColorMatrixIdentity(t *testing.T) {
	if !ColorMatrix(0, 100).IsIdentity() {
		t.Fatal("hue 0 / saturation 100 must be identity")
	}
	if ColorMatrix(0, 190).IsIdentity() {
		t.Fatal("saturation 190 must not be identity")
	}
}

func TestSaturationPreservesGray(t *testing.T) {
	m := ColorMatrix(25, 180)
	for _, v := range []float64{0.1, 0.5, 0.9} {
		for row := 0; row < 3; row++ {
			out := (m[row][0] + m[row][1] + m[row][2]) * v
			if math.Abs(out-v) > 1e-3 {
				t.Fatalf("gray %v drifted to %v on row %d", v, out, row)
			}
		}
	}
}

func TestMagEffectIsTransposed(t *testing.T) {
	m := Saturate(0) // grayscale: every output = luma weights
	e := MagEffect(m)
	// Row 0 (input R) should carry the red luma weight into all outputs.
	for out := 0; out < 3; out++ {
		if math.Abs(float64(e[out])-0.213) > 1e-6 {
			t.Fatalf("expected 0.213 at [0][%d], got %v", out, e[out])
		}
	}
	if e[18] != 1 || e[24] != 1 {
		t.Fatal("alpha/homogeneous diagonal must be 1")
	}
}

func TestLerpRamp(t *testing.T) {
	a := IdentityRamp()
	b := BuildRamp(Tone{Brightness: 140, Contrast: 120, Gamma: 1.6, Temperature: 6500, Red: 100, Green: 100, Blue: 100})
	half := LerpRamp(&a, &b, 0.5)
	if half.Distance(&a) == 0 || half.Distance(&b) == 0 {
		t.Fatal("lerp should land between both ramps")
	}
	full := LerpRamp(&a, &b, 1)
	if full.Distance(&b) != 0 {
		t.Fatal("lerp(1) must equal target")
	}
}
