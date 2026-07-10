package psx

// pad.go parses scripted controller input. A press spec is a comma-separated
// list of BUTTON@STEP:HOLD entries — press BUTTON once the run has executed
// STEP instructions and release it HOLD instructions later. Overlapping holds
// are OR-combined so chords work. The result is a PadScript for Machine.

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// padNames maps script button names to their active-low mask bit.
var padNames = map[string]uint16{
	"select": PadSelect, "start": PadStart,
	"up": PadUp, "right": PadRight, "down": PadDown, "left": PadLeft,
	"l2": PadL2, "r2": PadR2, "l1": PadL1, "r1": PadR1,
	"triangle": PadTriangle, "circle": PadCircle, "cross": PadCross, "square": PadSquare,
}

// ParsePress turns "start@380000000:400000,right@390000000:400000" into a
// time-ordered pad schedule. STEP and HOLD are decimal, or hex with a 0x/$
// prefix.
func ParsePress(spec string) ([]PadEvent, error) {
	type edge struct {
		step uint64
		bit  uint16
		down bool
	}
	var edges []edge
	for _, part := range strings.Split(spec, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		at := strings.IndexByte(part, '@')
		col := strings.LastIndexByte(part, ':')
		if at < 0 || col < at {
			return nil, fmt.Errorf("want BUTTON@STEP:HOLD, got %q", part)
		}
		bit, ok := padNames[strings.ToLower(part[:at])]
		if !ok {
			return nil, fmt.Errorf("unknown button %q", part[:at])
		}
		step, err := parseCount(part[at+1 : col])
		if err != nil {
			return nil, fmt.Errorf("bad step in %q", part)
		}
		hold, err := parseCount(part[col+1:])
		if err != nil {
			return nil, fmt.Errorf("bad hold in %q", part)
		}
		edges = append(edges, edge{step, bit, true}, edge{step + hold, bit, false})
	}
	sort.Slice(edges, func(i, j int) bool { return edges[i].step < edges[j].step })
	var script []PadEvent
	held := uint16(0)
	for _, e := range edges {
		if e.down {
			held |= e.bit
		} else {
			held &^= e.bit
		}
		buttons := PadReleased &^ held
		if n := len(script); n > 0 && script[n-1].AtStep == e.step {
			script[n-1].Buttons = buttons // coalesce simultaneous edges
		} else {
			script = append(script, PadEvent{AtStep: e.step, Buttons: buttons})
		}
	}
	return script, nil
}

func parseCount(s string) (uint64, error) {
	if strings.HasPrefix(s, "0x") || strings.HasPrefix(s, "$") {
		v, err := strconv.ParseUint(strings.TrimPrefix(strings.TrimPrefix(s, "$"), "0x"), 16, 64)
		return v, err
	}
	return strconv.ParseUint(s, 10, 64)
}

// PadBufAddr reports the port-1 pad buffer the game registered with InitPad
// (0 until it does) — instrumentation for pad-flow tracing.
func (m *Machine) PadBufAddr() uint32 { return m.padBuf }
