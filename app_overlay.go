package main

import (
	"bytes"
	"encoding/base64"
	"errors"
	"image/png"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/1337133713377/rekashade/internal/display"
	"github.com/1337133713377/rekashade/internal/watch"
)

// The UI renders the crosshair/reticle on a canvas (single source of truth,
// pixel-identical to the preview) and sends it here as a PNG. We convert it
// to premultiplied BGRA and hand it to the layered overlay window.

type overlayRuntime struct {
	mu          sync.Mutex
	png         []byte
	hasImage    bool
	holdHidden  bool
	fullscreen  bool
	visible     bool
	lastTargetX int
	lastTargetY int
	haveTarget  bool
	saveTimer   *time.Timer
	dirty       bool
}

// OverlayStatus is shown in the UI.
type OverlayStatus struct {
	Visible    bool   `json:"visible"`
	Reason     string `json:"reason"` // why it is hidden: off | noimage | hold | notfullscreen | ""
	HasImage   bool   `json:"hasImage"`
	HoldHidden bool   `json:"holdHidden"`
}

const maxOverlaySide = 1024

func decodeOverlay(pngBytes []byte) ([]byte, int, int, error) {
	cfgImg, err := png.DecodeConfig(bytes.NewReader(pngBytes))
	if err != nil {
		return nil, 0, 0, err
	}
	if cfgImg.Width <= 0 || cfgImg.Height <= 0 || cfgImg.Width > maxOverlaySide || cfgImg.Height > maxOverlaySide {
		return nil, 0, 0, errors.New("overlay image must be at most 1024×1024 pixels")
	}
	img, err := png.Decode(bytes.NewReader(pngBytes))
	if err != nil {
		return nil, 0, 0, err
	}
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	out := make([]byte, w*h*4)
	i := 0
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			r, g, bl, al := img.At(x, y).RGBA() // already premultiplied, 16-bit
			out[i], out[i+1], out[i+2], out[i+3] = byte(bl>>8), byte(g>>8), byte(r>>8), byte(al>>8)
			i += 4
		}
	}
	return out, w, h, nil
}

// SetOverlayImage receives the rendered crosshair as a PNG data URL plus the
// pixel (anchor) that must sit exactly on the aim point.
func (a *App) SetOverlayImage(dataURL string, anchorX, anchorY int) error {
	raw := dataURL
	if i := strings.Index(raw, ","); i >= 0 && strings.HasPrefix(raw, "data:") {
		raw = raw[i+1:]
	}
	pngBytes, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		return err
	}
	if err := a.pushOverlayPNG(pngBytes, anchorX, anchorY); err != nil {
		return err
	}
	a.ov.mu.Lock()
	a.ov.png = pngBytes
	a.ov.dirty = true
	if a.ov.saveTimer == nil {
		a.ov.saveTimer = time.AfterFunc(800*time.Millisecond, func() { a.ov.flush(a) })
	} else {
		a.ov.saveTimer.Reset(800 * time.Millisecond)
	}
	a.ov.mu.Unlock()
	return nil
}

// flush writes the last overlay bitmap to disk (debounced).
func (o *overlayRuntime) flush(a *App) {
	o.mu.Lock()
	data, dirty := o.png, o.dirty
	o.dirty = false
	o.mu.Unlock()
	if !dirty || a.store == nil || len(data) == 0 {
		return
	}
	_ = os.WriteFile(a.store.OverlayPath(), data, 0o644)
}

func (a *App) pushOverlayPNG(pngBytes []byte, anchorX, anchorY int) error {
	return a.pushOverlay(pngBytes, anchorX, anchorY, true)
}

func (a *App) pushOverlay(pngBytes []byte, anchorX, anchorY int, refresh bool) error {
	bgra, w, h, err := decodeOverlay(pngBytes)
	if err != nil {
		return err
	}
	if anchorX < 0 || anchorX >= w || anchorY < 0 || anchorY >= h {
		anchorX, anchorY = w/2, h/2
	}
	if a.host != nil {
		a.host.SetOverlayImage(bgra, w, h, anchorX, anchorY)
	}
	a.ov.mu.Lock()
	a.ov.hasImage = true
	a.ov.mu.Unlock()
	if refresh {
		a.updateOverlayVisibility()
	}
	return nil
}

// loadOverlayImage restores the last crosshair at startup so it shows up
// before the UI has even loaded (the image is always drawn centred).
func (a *App) loadOverlayImage() {
	var data []byte
	if a.store != nil {
		data, _ = os.ReadFile(a.store.OverlayPath())
	}
	if len(data) > 0 {
		if cfg, err := png.DecodeConfig(bytes.NewReader(data)); err == nil {
			if a.pushOverlay(data, cfg.Width/2, cfg.Height/2, false) == nil {
				a.ov.mu.Lock()
				a.ov.png = data
				a.ov.mu.Unlock()
			}
		}
	}
	a.applyOverlay() // target first, visibility last
}

// overlayTarget computes where the aim point is on screen.
func (a *App) overlayTarget() (int, int) {
	a.mu.Lock()
	cfg := a.cfg.Overlay
	list := a.displays
	fg := a.watch.fg
	a.mu.Unlock()

	if cfg.FollowWindow {
		if fg.HasClient && fg.PID != a.selfPID && fg.HWND != 0 {
			c := fg.Client
			if c.Right-c.Left >= 320 && c.Bottom-c.Top >= 240 {
				x, y := int(c.Left+(c.Right-c.Left)/2)+cfg.OffsetX, int(c.Top+(c.Bottom-c.Top)/2)+cfg.OffsetY
				a.ov.mu.Lock()
				a.ov.lastTargetX, a.ov.lastTargetY, a.ov.haveTarget = x, y, true
				a.ov.mu.Unlock()
				return x, y
			}
		}
		a.ov.mu.Lock()
		if a.ov.haveTarget {
			x, y := a.ov.lastTargetX, a.ov.lastTargetY
			a.ov.mu.Unlock()
			return x, y
		}
		a.ov.mu.Unlock()
	}
	d, _ := display.Find(list, cfg.Display)
	return d.X + d.Width/2 + cfg.OffsetX, d.Y + d.Height/2 + cfg.OffsetY
}

func (a *App) applyOverlay() {
	if a.host == nil {
		return
	}
	cfg := a.config().Overlay
	a.host.SetOverlayCaptureExcluded(cfg.HideFromCapture)
	if !cfg.FollowWindow {
		a.ov.mu.Lock()
		a.ov.haveTarget = false
		a.ov.mu.Unlock()
	}
	x, y := a.overlayTarget()
	a.host.SetOverlayTarget(x, y)
	a.host.SetFollowTimer(cfg.Enabled && cfg.FollowWindow)
	a.updateOverlayVisibility()
}

func (a *App) onFollowTick() {
	// Runs on the host thread every 50 ms while "follow active window" is on.
	fg := watch.GetForeground(nil, a.selfPID)
	a.mu.Lock()
	exe := a.watch.fg.Exe
	a.watch.fg = fg
	a.watch.fg.Exe = exe
	a.mu.Unlock()
	x, y := a.overlayTarget()
	a.host.SetOverlayTarget(x, y)
}

func (a *App) setHoldHidden(h bool) {
	a.ov.mu.Lock()
	a.ov.holdHidden = h
	a.ov.mu.Unlock()
	a.updateOverlayVisibility()
}

func (a *App) setFullscreenFocus(fs bool) {
	a.ov.mu.Lock()
	changed := a.ov.fullscreen != fs
	a.ov.fullscreen = fs
	a.ov.mu.Unlock()
	if changed {
		a.updateOverlayVisibility()
	}
}

func (a *App) overlayStatus() OverlayStatus {
	cfg := a.config().Overlay
	a.ov.mu.Lock()
	defer a.ov.mu.Unlock()
	s := OverlayStatus{Visible: a.ov.visible, HasImage: a.ov.hasImage, HoldHidden: a.ov.holdHidden}
	switch {
	case !cfg.Enabled:
		s.Reason = "off"
	case !a.ov.hasImage:
		s.Reason = "noimage"
	case a.ov.holdHidden:
		s.Reason = "hold"
	case cfg.OnlyFullscreen && !a.ov.fullscreen:
		s.Reason = "notfullscreen"
	}
	return s
}

func (a *App) updateOverlayVisibility() {
	if a.host == nil {
		return
	}
	cfg := a.config().Overlay
	a.ov.mu.Lock()
	vis := cfg.Enabled && a.ov.hasImage && !a.ov.holdHidden && (!cfg.OnlyFullscreen || a.ov.fullscreen)
	changed := vis != a.ov.visible
	a.ov.visible = vis
	a.ov.mu.Unlock()
	if changed {
		a.host.SetOverlayVisible(vis)
		go a.emitStatus()
	}
}
