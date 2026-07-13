package psp

// The GE register file persists across display lists — rasterList says so itself:
// "a frame's FBP/matrix setup list conditions the draw lists that follow, so the
// state lives on the machine". A savestate that does not carry it therefore does not
// restore the machine that was saved: the restored GE rebuilds itself from engine
// defaults, and the frame buffer pointer, the matrices, the texture and blend state
// the game established in an earlier list are simply gone until it happens to re-send
// them — which, for the state it programs once at engine init, is never.
//
// These two tests say that in the two ways it can be said: structurally (the restored
// register file equals the saved one) and behaviourally (a machine restored mid-render
// goes on to draw the same pixels as the machine it was cloned from).

import (
	"bytes"
	"reflect"
	"strings"
	"testing"
)

// TestGEWireCoversGEState is the guard on the mirror in ge_wire.go. geWire is generated
// from geState by reflection, and nothing makes the two stay in step — so a field added
// to the register file and not to the mirror would vanish from every savestate, without
// a word. That is the bug this whole file is about; this test refuses to let it recur.
func TestGEWireCoversGEState(t *testing.T) {
	st, wire := reflect.TypeOf(geState{}), reflect.TypeOf(geWire{})
	if st.NumField() != wire.NumField() {
		t.Fatalf("geState has %d fields, geWire mirrors %d: regenerate the mirror (see ge_wire.go)",
			st.NumField(), wire.NumField())
	}
	for i := 0; i < st.NumField(); i++ {
		f, w := st.Field(i), wire.Field(i)
		if want := strings.ToUpper(f.Name[:1]) + f.Name[1:]; w.Name != want {
			t.Errorf("field %d: geState.%s should mirror to geWire.%s, found geWire.%s", i, f.Name, want, w.Name)
		}
		if f.Type != w.Type {
			t.Errorf("field %d (%s): geState has %s, geWire has %s", i, f.Name, f.Type, w.Type)
		}
	}
}

// bootedWithGE boots the game and runs until the GE has executed a list, so there is
// a register file worth saving. It returns the machine and the steps it took.
func bootedWithGE(t *testing.T) (*Machine, uint64) {
	t.Helper()
	im, err := OpenImage(testImagePath(t))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { im.Close() })
	mod, err := im.LoadExecutable("PSP_GAME/SYSDIR/EBOOT.BIN")
	if err != nil {
		t.Fatal(err)
	}
	m := NewMachine()
	m.SetImageHash("ge-test")
	if err := m.LoadModule(mod); err != nil {
		t.Fatal(err)
	}

	// Run in slices until the GE has drawn something. The machine is deterministic, so
	// the step count that got us here is reproducible — the second machine reruns it.
	var steps uint64
	const slice = 2_000_000
	for i := 0; i < 40; i++ {
		res := m.Run(slice)
		steps += res.Steps
		if m.geSt != nil && len(m.GeLists) > 0 {
			return m, steps
		}
		if res.Steps < slice {
			t.Fatalf("run stopped early before the GE drew anything: %s", res)
		}
	}
	t.Skip("the game did not submit a GE display list within the step budget")
	return nil, 0
}

// TestSaveStateCarriesGERegisters is the structural statement: what the GE holds after
// a restore must be what it held when the state was taken.
func TestSaveStateCarriesGERegisters(t *testing.T) {
	m, _ := bootedWithGE(t)

	s := m.SaveState()
	m2 := NewMachine()
	m2.SetImageHash("ge-test")
	if err := m2.LoadState(s); err != nil {
		t.Fatal(err)
	}

	if m2.geSt == nil {
		t.Fatal("the restored machine has no GE register file: the savestate dropped it")
	}
	if *m2.geSt != *m.geSt {
		t.Error("the restored GE register file differs from the saved one")
		if m2.geSt.fbLow != m.geSt.fbLow || m2.geSt.fbHigh != m.geSt.fbHigh {
			t.Errorf("  frame buffer: restored %02X:%06X, saved %02X:%06X",
				m2.geSt.fbHigh, m2.geSt.fbLow, m.geSt.fbHigh, m.geSt.fbLow)
		}
		if m2.geSt.world != m.geSt.world {
			t.Errorf("  world matrix: restored %v, saved %v", m2.geSt.world, m.geSt.world)
		}
		if m2.geSt.texAddr != m.geSt.texAddr {
			t.Errorf("  texture: restored %08X, saved %08X", m2.geSt.texAddr, m.geSt.texAddr)
		}
	}

	// The snapshot must be independent of the machine that made it: framedbg restores
	// one snapshot repeatedly while the live machine runs on, so a restore that handed
	// back the live register file would let the present rewrite the past.
	m.geSt.texAddr ^= 0xFFFF
	m3 := NewMachine()
	m3.SetImageHash("ge-test")
	if err := m3.LoadState(s); err != nil {
		t.Fatal(err)
	}
	if m3.geSt.texAddr == m.geSt.texAddr {
		t.Error("the snapshot aliases the live GE register file instead of copying it")
	}
}

// TestRestoredMachineDrawsTheSame is the behavioural statement, and the one that
// matters: clone a machine mid-render and both must go on to draw the same pixels.
// VRAM, not the visible framebuffer — the off-screen render targets are where a
// wrong FBP or a stale matrix shows up first.
func TestRestoredMachineDrawsTheSame(t *testing.T) {
	m, _ := bootedWithGE(t)

	s := m.SaveState()
	m2 := NewMachine()
	m2.SetImageHash("ge-test")
	if err := m2.LoadState(s); err != nil {
		t.Fatal(err)
	}

	// Both machines now run the same number of instructions from the same state.
	const more = 4_000_000
	r1 := m.Run(more)
	r2 := m2.Run(more)

	if r1.Steps != r2.Steps || r1.PC != r2.PC {
		t.Errorf("the restored machine diverged: live stopped %s; restored stopped %s", r1, r2)
	}
	if !bytes.Equal(m.vram, m2.vram) {
		n := 0
		first := -1
		for i := range m.vram {
			if m.vram[i] != m2.vram[i] {
				if first < 0 {
					first = i
				}
				n++
			}
		}
		t.Errorf("the restored machine drew different pixels: %d of %d VRAM bytes differ, first at 0x%X (VRAM 0x%08X)",
			n, len(m.vram), first, vramBase+uint32(first))
	}
}
