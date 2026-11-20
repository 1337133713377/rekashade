//go:build windows

package display

import (
	"encoding/binary"
	"errors"
	"math"
	"strings"
	"sync"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"

	"github.com/1337133713377/rekashade/internal/win"
)

// Vibrance is "digital vibrance" as exposed by the GPU driver (NVIDIA
// Control Panel / AMD Software). It is applied by the display hardware, so
// it also works in exclusive-fullscreen games and costs zero performance.
type vibranceBackend interface {
	name() string
	has(device string) bool
	// set applies pct (100 = driver default, 200 = driver maximum) and
	// returns the part it could not apply as a saturation factor (1 = all
	// applied) that the caller may hand to the colour matrix instead.
	set(device string, pct float64) (float64, error)
	restore(device string) error
	originals() map[string]int
	close()
}

// ---------------------------------------------------------------------------
// NVIDIA (nvapi64.dll — ships with every GeForce driver)
// ---------------------------------------------------------------------------

const (
	nvInitialize               = 0x0150E828
	nvUnload                   = 0xD22BDD7E
	nvEnumNvidiaDisplay        = 0x9ABDD40D
	nvGetDisplayName           = 0x22A78B05
	nvGetDVCInfo               = 0x4085DE45
	nvSetDVCLevel              = 0x172409B4
	nvDVCInfoSize              = 16
	nvDVCInfoVersion    uint32 = nvDVCInfoSize | 1<<16
)

type nvDVCInfo struct {
	Version uint32
	Current uint32
	Min     uint32
	Max     uint32
	_       [16]byte // slack in case a driver writes a newer, larger struct
}

type nvDisplay struct {
	handle   uintptr
	min, max uint32
	orig     uint32
	last     int64
}

type nvidia struct {
	mu     sync.Mutex
	get    uintptr
	setFn  uintptr
	unload uintptr
	disp   map[string]*nvDisplay
	orig   map[string]int
}

func nvOK(r uintptr) bool { return int32(r) == 0 }

func loadNvidia(saved map[string]int) (*nvidia, error) {
	dll := windows.NewLazySystemDLL("nvapi64.dll")
	if err := dll.Load(); err != nil {
		return nil, err
	}
	qi := dll.NewProc("nvapi_QueryInterface")
	if err := qi.Find(); err != nil {
		return nil, err
	}
	q := func(id uint32) uintptr { r, _, _ := qi.Call(uintptr(id)); return r }
	initFn, enum, nameFn := q(nvInitialize), q(nvEnumNvidiaDisplay), q(nvGetDisplayName)
	n := &nvidia{get: q(nvGetDVCInfo), setFn: q(nvSetDVCLevel), unload: q(nvUnload), disp: map[string]*nvDisplay{}, orig: map[string]int{}}
	if initFn == 0 || enum == 0 || nameFn == 0 || n.get == 0 || n.setFn == 0 {
		return nil, errors.New("nvapi: required functions not exported by this driver")
	}
	if r, _, _ := syscall.SyscallN(initFn); !nvOK(r) {
		return nil, errors.New("nvapi: initialisation failed")
	}
	for i := uint32(0); i < 16; i++ {
		var h uintptr
		if r, _, _ := syscall.SyscallN(enum, uintptr(i), uintptr(unsafe.Pointer(&h))); !nvOK(r) {
			break
		}
		var buf [64]byte
		if r, _, _ := syscall.SyscallN(nameFn, h, uintptr(unsafe.Pointer(&buf[0]))); !nvOK(r) {
			continue
		}
		dev := strings.ToUpper(cString(buf[:]))
		info := nvDVCInfo{Version: nvDVCInfoVersion}
		if r, _, _ := syscall.SyscallN(n.get, h, 0, uintptr(unsafe.Pointer(&info))); !nvOK(r) || info.Max <= info.Min {
			continue
		}
		d := &nvDisplay{handle: h, min: info.Min, max: info.Max, orig: info.Current, last: -1}
		if v, ok := saved[dev]; ok && v >= int(info.Min) && v <= int(info.Max) {
			d.orig = uint32(v) // a previous session crashed while tuned
		}
		n.disp[dev] = d
		n.orig[dev] = int(d.orig)
	}
	if len(n.disp) == 0 {
		n.close()
		return nil, errors.New("nvapi: no display supports digital vibrance")
	}
	return n, nil
}

func (n *nvidia) name() string { return "nvidia" }

func (n *nvidia) has(device string) bool {
	n.mu.Lock()
	defer n.mu.Unlock()
	_, ok := n.disp[strings.ToUpper(device)]
	return ok
}

func (n *nvidia) level(d *nvDisplay, lvl int64) error {
	if d.last == lvl {
		return nil
	}
	if r, _, _ := syscall.SyscallN(n.setFn, d.handle, 0, uintptr(lvl)); !nvOK(r) {
		return errors.New("nvapi: SetDVCLevel failed")
	}
	d.last = lvl
	return nil
}

func (n *nvidia) set(device string, pct float64) (float64, error) {
	n.mu.Lock()
	defer n.mu.Unlock()
	d, ok := n.disp[strings.ToUpper(device)]
	if !ok {
		return pct / 100, errors.New("nvapi: display not found")
	}
	rem := 1.0
	if pct < 100 { // NVIDIA's DVC cannot go below its default
		rem = pct / 100
		pct = 100
	}
	lvl := int64(d.min) + int64(math.Round((pct-100)/100*float64(d.max-d.min)))
	return rem, n.level(d, lvl)
}

func (n *nvidia) restore(device string) error {
	n.mu.Lock()
	defer n.mu.Unlock()
	d, ok := n.disp[strings.ToUpper(device)]
	if !ok {
		return nil
	}
	return n.level(d, int64(d.orig))
}

func (n *nvidia) originals() map[string]int { return n.orig }

func (n *nvidia) close() {
	if n.unload != 0 {
		syscall.SyscallN(n.unload)
	}
}

func cString(b []byte) string {
	for i, c := range b {
		if c == 0 {
			return string(b[:i])
		}
	}
	return string(b)
}

// ---------------------------------------------------------------------------
// AMD (atiadlxx.dll — ADL, ships with every Radeon driver)
// ---------------------------------------------------------------------------

const (
	adlColorSaturation   = 1 << 2
	adlDisplayConnected  = 0x1
	adlDisplayMapped     = 0x2
	adlAdapterInfoSize   = 1572 // sizeof(AdapterInfo) on Windows
	adlDisplayInfoSize   = 552  // sizeof(ADLDisplayInfo)
	adlAdapterIndexOff   = 4
	adlDisplayNameOff    = 536
	adlPresentOff        = 792
	adlDispLogicalIdxOff = 0
	adlDispLogicalAdpOff = 8
	adlDispInfoValueOff  = 548
)

var (
	adlMallocOnce sync.Once
	adlMallocCb   uintptr
)

func adlMalloc() uintptr {
	adlMallocOnce.Do(func() {
		adlMallocCb = windows.NewCallback(func(size uintptr) uintptr {
			p, err := windows.LocalAlloc(0x40 /* LPTR */, uint32(size))
			if err != nil {
				return 0
			}
			return p
		})
	})
	return adlMallocCb
}

type amdDisplay struct {
	adapter, display int32
	min, max, def    int32
	orig, last       int32
	hasLast          bool
}

type amd struct {
	mu      sync.Mutex
	ctx     uintptr
	get     *windows.LazyProc
	setP    *windows.LazyProc
	destroy *windows.LazyProc
	disp    map[string]*amdDisplay
	orig    map[string]int
}

func adlOK(r uintptr) bool { return int32(r) >= 0 }

func le32(b []byte, off int) int32 { return int32(binary.LittleEndian.Uint32(b[off:])) }

func loadAMD(saved map[string]int) (*amd, error) {
	dll := windows.NewLazySystemDLL("atiadlxx.dll")
	if err := dll.Load(); err != nil {
		return nil, err
	}
	create := dll.NewProc("ADL2_Main_Control_Create")
	numGet := dll.NewProc("ADL2_Adapter_NumberOfAdapters_Get")
	infoGet := dll.NewProc("ADL2_Adapter_AdapterInfo_Get")
	dispGet := dll.NewProc("ADL2_Display_DisplayInfo_Get")
	a := &amd{
		get: dll.NewProc("ADL2_Display_Color_Get"), setP: dll.NewProc("ADL2_Display_Color_Set"),
		destroy: dll.NewProc("ADL2_Main_Control_Destroy"), disp: map[string]*amdDisplay{}, orig: map[string]int{},
	}
	for _, p := range []*windows.LazyProc{create, numGet, infoGet, dispGet, a.get, a.setP, a.destroy} {
		if err := p.Find(); err != nil {
			return nil, err
		}
	}
	if r, _, _ := create.Call(adlMalloc(), 1, uintptr(unsafe.Pointer(&a.ctx))); !adlOK(r) || a.ctx == 0 {
		return nil, errors.New("adl: initialisation failed")
	}
	var n int32
	if r, _, _ := numGet.Call(a.ctx, uintptr(unsafe.Pointer(&n))); !adlOK(r) || n <= 0 || n > 256 {
		a.close()
		return nil, errors.New("adl: no adapters")
	}
	infos := make([]byte, int(n)*adlAdapterInfoSize)
	for i := 0; i < int(n); i++ {
		binary.LittleEndian.PutUint32(infos[i*adlAdapterInfoSize:], adlAdapterInfoSize)
	}
	if r, _, _ := infoGet.Call(a.ctx, uintptr(unsafe.Pointer(&infos[0])), uintptr(len(infos))); !adlOK(r) {
		a.close()
		return nil, errors.New("adl: adapter info failed")
	}
	for i := 0; i < int(n); i++ {
		ai := infos[i*adlAdapterInfoSize : (i+1)*adlAdapterInfoSize]
		if le32(ai, adlPresentOff) == 0 {
			continue
		}
		adapter := le32(ai, adlAdapterIndexOff)
		dev := strings.ToUpper(cString(ai[adlDisplayNameOff : adlDisplayNameOff+256]))
		if dev == "" || a.disp[dev] != nil {
			continue
		}
		var count int32
		var ptr uintptr
		if r, _, _ := dispGet.Call(a.ctx, uintptr(adapter), uintptr(unsafe.Pointer(&count)), uintptr(unsafe.Pointer(&ptr)), 0); !adlOK(r) || ptr == 0 {
			continue
		}
		if count <= 0 || count > 64 {
			windows.LocalFree(windows.Handle(ptr))
			continue
		}
		raw := make([]byte, int(count)*adlDisplayInfoSize)
		win.CopyFromPtr(raw, ptr)
		windows.LocalFree(windows.Handle(ptr))
		for j := 0; j < int(count); j++ {
			di := raw[j*adlDisplayInfoSize : (j+1)*adlDisplayInfoSize]
			val := le32(di, adlDispInfoValueOff)
			if val&adlDisplayConnected == 0 || val&adlDisplayMapped == 0 || le32(di, adlDispLogicalAdpOff) != adapter {
				continue
			}
			disp := le32(di, adlDispLogicalIdxOff)
			var cur, def, mn, mx, step int32
			r, _, _ := a.get.Call(a.ctx, uintptr(adapter), uintptr(disp), adlColorSaturation,
				uintptr(unsafe.Pointer(&cur)), uintptr(unsafe.Pointer(&def)), uintptr(unsafe.Pointer(&mn)),
				uintptr(unsafe.Pointer(&mx)), uintptr(unsafe.Pointer(&step)))
			if !adlOK(r) || mx <= mn {
				continue
			}
			d := &amdDisplay{adapter: adapter, display: disp, min: mn, max: mx, def: def, orig: cur}
			if v, ok := saved[dev]; ok && v >= int(mn) && v <= int(mx) {
				d.orig = int32(v)
			}
			a.disp[dev] = d
			a.orig[dev] = int(d.orig)
			break
		}
	}
	if len(a.disp) == 0 {
		a.close()
		return nil, errors.New("adl: no display exposes saturation")
	}
	return a, nil
}

func (a *amd) name() string { return "amd" }

func (a *amd) has(device string) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	_, ok := a.disp[strings.ToUpper(device)]
	return ok
}

func (a *amd) value(d *amdDisplay, v int32) error {
	if d.hasLast && d.last == v {
		return nil
	}
	if r, _, _ := a.setP.Call(a.ctx, uintptr(d.adapter), uintptr(d.display), adlColorSaturation, uintptr(v)); !adlOK(r) {
		return errors.New("adl: Display_Color_Set failed")
	}
	d.last, d.hasLast = v, true
	return nil
}

func (a *amd) set(device string, pct float64) (float64, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	d, ok := a.disp[strings.ToUpper(device)]
	if !ok {
		return pct / 100, errors.New("adl: display not found")
	}
	var v float64
	if pct >= 100 {
		v = float64(d.def) + (pct-100)/100*float64(d.max-d.def)
	} else {
		v = float64(d.def) - (100-pct)/100*float64(d.def-d.min)
	}
	return 1, a.value(d, int32(math.Round(v)))
}

func (a *amd) restore(device string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	d, ok := a.disp[strings.ToUpper(device)]
	if !ok {
		return nil
	}
	return a.value(d, d.orig)
}

func (a *amd) originals() map[string]int { return a.orig }

func (a *amd) close() {
	if a.ctx != 0 {
		a.destroy.Call(a.ctx)
		a.ctx = 0
	}
}
