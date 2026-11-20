// Package colorx holds the colour maths used by the native engine.
//
// The exact same formulas live in frontend/src/lib/color.ts so that the
// before/after preview in the UI matches what Windows really applies.
// If you change a formula here, change it there too (tests guard both).
package colorx

import "math"

// Tone is the per-display part of a tuning: everything that can be done by
// the display's hardware gamma ramp (a per-channel 1D look-up table).
type Tone struct {
	Brightness  float64 // percent, 100 = neutral
	Contrast    float64 // percent, 100 = neutral
	Gamma       float64 // exponent, 1 = neutral
	Temperature float64 // Kelvin, 6500 = neutral
	Red         float64 // percent, 100 = neutral
	Green       float64
	Blue        float64
}

// NeutralTone returns a tone that leaves the image untouched.
func NeutralTone() Tone {
	return Tone{Brightness: 100, Contrast: 100, Gamma: 1, Temperature: 6500, Red: 100, Green: 100, Blue: 100}
}

// Ramp is a gamma ramp in the WORD[3][256] layout Windows expects.
type Ramp [3][256]uint16

// IdentityRamp returns the linear ramp (no change).
func IdentityRamp() Ramp {
	var r Ramp
	for i := 0; i < 256; i++ {
		v := uint16(i * 257)
		r[0][i], r[1][i], r[2][i] = v, v, v
	}
	return r
}

func clamp01(x float64) float64 {
	if x < 0 {
		return 0
	}
	if x > 1 {
		return 1
	}
	return x
}

// toneBase applies gamma, contrast and brightness (channel independent).
func toneBase(t Tone, x float64) float64 {
	g := t.Gamma
	if g < 0.1 || math.IsNaN(g) {
		g = 1
	}
	y := math.Pow(x, 1/g)
	y = (y-0.5)*(t.Contrast/100) + 0.5
	b := (t.Brightness - 100) / 100
	y = y*(1+b*0.6) + math.Max(b, 0)*0.12*(1-y)
	return y
}

// KelvinRGB approximates the colour of a black body (Tanner Helland's fit).
// Returned channels are in 0..1.
func KelvinRGB(k float64) (r, g, b float64) {
	t := k / 100
	if t <= 66 {
		r = 255
		g = 99.4708025861*math.Log(t) - 161.1195681661
	} else {
		r = 329.698727446 * math.Pow(t-60, -0.1332047592)
		g = 288.1221695283 * math.Pow(t-60, -0.0755148492)
	}
	switch {
	case t >= 66:
		b = 255
	case t <= 19:
		b = 0
	default:
		b = 138.5177312231*math.Log(t-10) - 305.0447927307
	}
	c := func(v float64) float64 { return math.Min(255, math.Max(0, v)) / 255 }
	return c(r), c(g), c(b)
}

// TemperatureGains returns per-channel multipliers relative to 6500K,
// normalised so the strongest channel is 1 (nothing clips).
func TemperatureGains(k float64) [3]float64 {
	if k <= 0 || math.Abs(k-6500) < 1 {
		return [3]float64{1, 1, 1}
	}
	r, g, b := KelvinRGB(k)
	r0, g0, b0 := KelvinRGB(6500)
	gains := [3]float64{r / r0, g / g0, b / b0}
	m := math.Max(gains[0], math.Max(gains[1], gains[2]))
	for i := range gains {
		gains[i] /= m
	}
	return gains
}

// ChannelGains combines the RGB sliders with the colour temperature.
func ChannelGains(t Tone) [3]float64 {
	tg := TemperatureGains(t.Temperature)
	return [3]float64{t.Red / 100 * tg[0], t.Green / 100 * tg[1], t.Blue / 100 * tg[2]}
}

// Curve evaluates the tone curve of channel c (0=R,1=G,2=B) at x in 0..1.
func Curve(t Tone, c int, x float64) float64 {
	return clamp01(toneBase(t, x) * ChannelGains(t)[c])
}

// BuildRamp turns a tone into a monotonic gamma ramp.
func BuildRamp(t Tone) Ramp {
	var r Ramp
	gains := ChannelGains(t)
	for i := 0; i < 256; i++ {
		base := toneBase(t, float64(i)/255)
		for c := 0; c < 3; c++ {
			r[c][i] = uint16(math.Round(clamp01(base*gains[c]) * 65535))
		}
	}
	for c := 0; c < 3; c++ { // Windows rejects non-monotonic ramps
		for i := 1; i < 256; i++ {
			if r[c][i] < r[c][i-1] {
				r[c][i] = r[c][i-1]
			}
		}
	}
	return r
}

// LerpRamp blends from a to b by f (0 = a, 1 = b). Used to back off when
// Windows refuses a ramp that strays too far from linear.
func LerpRamp(a, b *Ramp, f float64) Ramp {
	var r Ramp
	for c := 0; c < 3; c++ {
		for i := 0; i < 256; i++ {
			r[c][i] = uint16(math.Round(float64(a[c][i]) + (float64(b[c][i])-float64(a[c][i]))*f))
		}
	}
	return r
}

// IsIdentity reports whether a ramp is (almost) linear.
func (r *Ramp) IsIdentity() bool {
	id := IdentityRamp()
	return r.Distance(&id) <= 64
}

// Distance returns the largest per-entry difference between two ramps.
func (r *Ramp) Distance(o *Ramp) int {
	worst := 0
	for c := 0; c < 3; c++ {
		for i := 0; i < 256; i++ {
			d := int(r[c][i]) - int(o[c][i])
			if d < 0 {
				d = -d
			}
			if d > worst {
				worst = d
			}
		}
	}
	return worst
}

// Mat3 is a 3x3 colour matrix in column-vector form: out = M · [r g b].
type Mat3 [3][3]float64

// Identity3 returns the identity matrix.
func Identity3() Mat3 { return Mat3{{1, 0, 0}, {0, 1, 0}, {0, 0, 1}} }

// Saturate is the SVG/CSS "saturate" matrix (s = 1 is neutral).
func Saturate(s float64) Mat3 {
	return Mat3{
		{0.213 + 0.787*s, 0.715 - 0.715*s, 0.072 - 0.072*s},
		{0.213 - 0.213*s, 0.715 + 0.285*s, 0.072 - 0.072*s},
		{0.213 - 0.213*s, 0.715 - 0.715*s, 0.072 + 0.928*s},
	}
}

// HueRotate is the SVG/CSS "hue-rotate" matrix.
func HueRotate(deg float64) Mat3 {
	a := deg * math.Pi / 180
	c, s := math.Cos(a), math.Sin(a)
	return Mat3{
		{0.213 + c*0.787 - s*0.213, 0.715 - c*0.715 - s*0.715, 0.072 - c*0.072 + s*0.928},
		{0.213 - c*0.213 + s*0.143, 0.715 + c*0.285 + s*0.140, 0.072 - c*0.072 - s*0.283},
		{0.213 - c*0.213 - s*0.787, 0.715 - c*0.715 + s*0.715, 0.072 + c*0.928 + s*0.072},
	}
}

// Mul returns a·b (apply b first, then a).
func Mul(a, b Mat3) Mat3 {
	var m Mat3
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			for k := 0; k < 3; k++ {
				m[i][j] += a[i][k] * b[k][j]
			}
		}
	}
	return m
}

// IsIdentity reports whether m is (almost) the identity.
func (m Mat3) IsIdentity() bool {
	id := Identity3()
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			if math.Abs(m[i][j]-id[i][j]) > 1e-4 {
				return false
			}
		}
	}
	return true
}

// ColorMatrix builds the global colour matrix from hue (degrees) and
// saturation (percent, 100 = neutral).
func ColorMatrix(hueDeg, saturationPct float64) Mat3 {
	return Mul(HueRotate(hueDeg), Saturate(saturationPct/100))
}

// MagEffect converts m to the 5x5 row-vector layout of MAGCOLOREFFECT
// (row = input channel, column = output channel, last row = offsets).
func MagEffect(m Mat3) [25]float32 {
	var e [25]float32
	for in := 0; in < 3; in++ {
		for out := 0; out < 3; out++ {
			e[in*5+out] = float32(m[out][in])
		}
	}
	e[3*5+3] = 1 // alpha passthrough
	e[4*5+4] = 1 // homogeneous coordinate
	return e
}
