package debug

import (
	"fmt"
	"sort"
	"sync"
)

// The platform registry.
//
// An adapter registers itself in an init function; the debugger imports the adapters
// it wants and then opens games by name. This is the only way round: the tools module
// cannot import the games, and debug cannot import the adapters (they import it), so
// the binary is what brings the two together.

// OpenSpec is everything needed to open a game.
type OpenSpec struct {
	// Image is the ROM, disc or game directory.
	Image string

	// Profile carries the handful of things a platform cannot work out for itself and
	// that are properties of the *game*, not the machine — the boot executable inside
	// a PSP ISO, the interrupt handler a PSX game expects the BIOS to have installed.
	// It comes from the game's debug.json (see Library), so no game knowledge lives in
	// the tools module.
	Profile map[string]string
}

// Get reads a profile key.
func (s OpenSpec) Get(key string) string { return s.Profile[key] }

type opener func(OpenSpec) (Target, error)

var (
	regMu sync.RWMutex
	reg   = map[string]opener{}
)

// Register adds a platform's opener. Adapters call it from init.
func Register(platform string, open func(OpenSpec) (Target, error)) {
	regMu.Lock()
	defer regMu.Unlock()
	if _, dup := reg[platform]; dup {
		panic("debug: platform " + platform + " registered twice")
	}
	reg[platform] = open
}

// Open opens a game on a registered platform.
func Open(platform string, spec OpenSpec) (Target, error) {
	regMu.RLock()
	open, ok := reg[platform]
	regMu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("debug: no adapter for platform %q (registered: %v)", platform, Platforms())
	}
	return open(spec)
}

// Platforms lists the registered platforms.
func Platforms() []string {
	regMu.RLock()
	defer regMu.RUnlock()
	out := make([]string, 0, len(reg))
	for p := range reg {
		out = append(out, p)
	}
	sort.Strings(out)
	return out
}

// Registered reports whether a platform has an adapter.
func Registered(platform string) bool {
	regMu.RLock()
	defer regMu.RUnlock()
	_, ok := reg[platform]
	return ok
}
