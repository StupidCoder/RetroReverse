package n3ds

import (
	"fmt"
	"sort"
	"strings"
)

// hid.go models the HID (Human Interface Device) service's shared memory — the
// page through which the input driver publishes controller, touch-screen and
// motion state, and the application polls it each frame. The service hands the
// application a shared-memory block and a set of interrupt events (GetIPCHandles,
// ipc_services.go); this file backs that block with a live pad-state ring the
// oracle can drive, so the game can be played past dialogs and menus.
//
// The layout is not guessed: DumpHIDReads (driven by -hidtrace) records which
// offsets the game actually reads, and the ring below is filled at exactly those
// offsets. Following [[trace-pointers-not-heuristics]], the shape comes from the
// game's own polling, verified by the game advancing when a button is asserted.

// hidButton bit assignments — the pad bitmask the game tests. Names map to the
// 3DS face/shoulder/d-pad buttons; the values are the driver's bit positions,
// confirmed by which bit the game polls when waiting on a dialog's OK.
const (
	hidA      = 1 << 0
	hidB      = 1 << 1
	hidSelect = 1 << 2
	hidStart  = 1 << 3
	hidRight  = 1 << 4
	hidLeft   = 1 << 5
	hidUp     = 1 << 6
	hidDown   = 1 << 7
	hidR      = 1 << 8
	hidL      = 1 << 9
	hidX      = 1 << 10
	hidY      = 1 << 11
)

var hidButtonNames = map[string]uint32{
	"a": hidA, "b": hidB, "x": hidX, "y": hidY,
	"l": hidL, "r": hidR,
	"up": hidUp, "down": hidDown, "left": hidLeft, "right": hidRight,
	"start": hidStart, "select": hidSelect,
}

// SetKeys parses a comma-separated button list into the held-button mask the HID
// shared memory will publish. Empty names are ignored; an unknown name is an
// error so a typo does not silently inject nothing.
func (m *Machine) SetKeys(list string) error {
	var mask uint32
	for _, name := range strings.Split(list, ",") {
		name = strings.TrimSpace(strings.ToLower(name))
		if name == "" {
			continue
		}
		bit, ok := hidButtonNames[name]
		if !ok {
			return fmt.Errorf("unknown key %q (valid: a,b,x,y,l,r,up,down,left,right,start,select)", name)
		}
		mask |= bit
	}
	m.hidButtons = mask
	return nil
}

// HID shared-memory section layout, read off the game's own sample-copy loop
// (disassembled at 0x00297700): three sections the render loop polls each frame.
// Each section is a header — two 64-bit timestamps at +0x00 and +0x08 (the driver
// bumps them every sample; the game compares them against its saved copy to count
// how many new entries arrived) and the latest-entry index (u32) at +0x10 — then
// an 8-slot ring of 0x10-byte entries starting at +0x28. A PadData entry's first
// word is the current button mask (the game derives keys-down itself by diffing
// against the previous poll); the remaining words are the circle-pad reading. The
// touch and accel sections share the header shape and are published as fresh
// no-input samples (a real console reports "nothing touched / neutral motion").
const (
	hidPadBase   = 0x000 // PadData
	hidTouchBase = 0x0A8 // TouchData
	hidAccelBase = 0x108 // Accelerometer
	hidHdrIndex  = 0x10  // u32 latest-entry index within a section
	hidEntries   = 0x28  // ring start within a section
	hidEntSz     = 0x10  // one sample
	hidRingLen   = 8
)

// updateHIDShared publishes one fresh input sample into the HID shared memory —
// the job the HID module does each frame on hardware. It advances the ring index
// and the section timestamps (so the game's "how many new samples?" check sees
// one), and writes the current button mask into the newest entry. The game
// computes press/release edges itself by diffing successive polls, so holding a
// button (cur set every frame) yields exactly one keys-down edge on the first
// frame after it changes. Called from the VBlank heartbeat, so it is paced by
// retired instructions and rides the savestate deterministically.
func (m *Machine) updateHIDShared() {
	if m.hidSharedAddr == 0 {
		return
	}
	idx := m.hidRingIdx
	tick := m.vblankCount + 1 // strictly increasing, and > the all-zero pre-injection tick
	cur := m.hidButtons
	m.hidPrevButtons = cur

	writeHeader := func(sb uint32) {
		base := m.hidSharedAddr + sb
		m.WriteWord(base+0x00, uint32(tick))
		m.WriteWord(base+0x04, uint32(tick>>32))
		m.WriteWord(base+0x08, uint32(tick))
		m.WriteWord(base+0x0C, uint32(tick>>32))
		m.WriteWord(base+hidHdrIndex, idx)
	}

	// PadData: header + the newest entry carrying the held buttons.
	writeHeader(hidPadBase)
	ent := m.hidSharedAddr + hidPadBase + hidEntries + idx*hidEntSz
	m.WriteWord(ent+0x0, cur)
	m.WriteWord(ent+0x4, 0)
	m.WriteWord(ent+0x8, 0)
	m.WriteWord(ent+0xC, 0) // circle pad neutral

	// Touch and accelerometer: fresh headers, entries left zero (no input).
	writeHeader(hidTouchBase)
	writeHeader(hidAccelBase)

	m.hidRingIdx = (idx + 1) & (hidRingLen - 1)

	// Signal the pad interrupt event (event 0 of GetIPCHandles) so a game blocked
	// on a new HID sample wakes; harmless when it polls the shared memory directly.
	if len(m.hidEvents) > 0 {
		if obj := m.handles[m.hidEvents[0]]; obj != nil {
			obj.signal = true
			if m.signalObject(obj) {
				m.reschedule = true
			}
		}
	}
}

// DumpHIDReads prints the histogram of reads within the HID shared-memory block,
// so the offsets the game polls (the ring header, the newest PadData entry, the
// touch/circle state) can be identified before the ring is modelled.
func (m *Machine) DumpHIDReads() {
	fmt.Printf("\nHID shared-memory reads (base 0x%08X):\n", m.hidSharedAddr)
	if len(m.hidReadHist) == 0 {
		fmt.Println("  (none — the game did not read the HID block during this run)")
		return
	}
	offs := make([]uint32, 0, len(m.hidReadHist))
	for o := range m.hidReadHist {
		offs = append(offs, o)
	}
	sort.Slice(offs, func(i, j int) bool { return offs[i] < offs[j] })
	for _, o := range offs {
		fmt.Printf("  +0x%03X  %d reads (last reader pc=0x%08X)\n", o, m.hidReadHist[o], m.hidReadPC[o])
	}
}
