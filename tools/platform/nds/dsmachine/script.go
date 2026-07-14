package dsmachine

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

// Input scripts: a timed sequence of button presses and stylus movements, played
// into the machine's hardware as the frames go by.
//
// This is what "the oracle plays the game" means in practice, and the reason it is a
// script rather than a set of held keys is the press EDGE. A DS game does not ask
// "is A down", it asks "did A go down since last frame" — and a button held from
// reset produces exactly one edge, at reset, before the game is listening. Hold the
// stylus on SM64DS's "TOUCH TO START" from frame zero and the title screen waits for
// ever, because the touch it is waiting for already happened.
//
// The format is one event per line: a frame number, an action, and its arguments.
//
//	60   touch 128,96     # put the stylus down at (128,96)
//	66   release          # lift it
//	120  press a,start    # hold A and Start
//	126  press            # release everything
//	200  touch 40,150
//	206  release
//
// Blank lines and everything after a '#' are ignored.

type inputEvent struct {
	frame  uint64
	keys   uint32
	touch  bool
	x, y   int
	setKey bool // this event sets the button state
	setTch bool // this event sets the stylus state
}

// Script is a parsed input script, ready to be attached to a machine.
type Script struct{ events []inputEvent }

// LoadScript reads an input script from a file.
func LoadScript(path string) (*Script, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	s := &Script{}
	sc := bufio.NewScanner(f)
	for line := 1; sc.Scan(); line++ {
		t := sc.Text()
		if i := strings.IndexByte(t, '#'); i >= 0 {
			t = t[:i]
		}
		fields := strings.Fields(t)
		if len(fields) == 0 {
			continue
		}
		frame, err := strconv.ParseUint(fields[0], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("%s:%d: want a frame number, got %q", path, line, fields[0])
		}
		if len(fields) < 2 {
			return nil, fmt.Errorf("%s:%d: no action", path, line)
		}
		ev := inputEvent{frame: frame}
		switch fields[1] {
		case "touch":
			if len(fields) < 3 {
				return nil, fmt.Errorf("%s:%d: touch needs X,Y", path, line)
			}
			if _, err := fmt.Sscanf(fields[2], "%d,%d", &ev.x, &ev.y); err != nil {
				return nil, fmt.Errorf("%s:%d: touch wants X,Y", path, line)
			}
			ev.touch, ev.setTch = true, true
		case "release":
			ev.touch, ev.setTch = false, true
		case "press":
			ev.setKey = true
			if len(fields) >= 3 {
				m, ok := ParseKeys(fields[2])
				if !ok {
					return nil, fmt.Errorf("%s:%d: unknown button in %q", path, line, fields[2])
				}
				ev.keys = m
			}
		default:
			return nil, fmt.Errorf("%s:%d: unknown action %q", path, line, fields[1])
		}
		s.events = append(s.events, ev)
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	sort.SliceStable(s.events, func(i, j int) bool { return s.events[i].frame < s.events[j].frame })
	return s, nil
}

// Play attaches the script to the machine: its events are applied at the top of the
// frames they name, as those frames arrive.
func (m *Machine) Play(s *Script) {
	if s == nil {
		return
	}
	next := 0
	prev := m.OnFrame
	m.OnFrame = func() {
		f := m.Frame()
		for next < len(s.events) && s.events[next].frame <= f {
			ev := s.events[next]
			if ev.setKey {
				m.SetKeys(ev.keys)
			}
			if ev.setTch {
				m.SetTouch(ev.x, ev.y, ev.touch)
			}
			next++
		}
		if prev != nil {
			prev()
		}
	}
}
