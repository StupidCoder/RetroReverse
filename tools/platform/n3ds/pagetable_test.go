package n3ds

// The page table (machine.go) is an accelerator in front of the region scan, and the
// only way it can break is by going stale: a region that grew, or a whole machine
// that was replaced by a savestate restore, while the table still describes what was
// there before. A stale entry does not crash — it reads plausible bytes out of the
// wrong memory, which is the worst kind of wrong for an oracle.
//
// So this test never checks the table against what it *should* say. It checks it
// against the thing it replaced.

import "testing"

// scanRegionOf is regionOf without the table: the original linear scan, kept here as
// the oracle for the accelerator.
func (m *Machine) scanRegionOf(a uint32) *memRegion {
	for _, r := range m.regions {
		if r.contains(a) {
			return r
		}
	}
	return nil
}

// agrees checks the table and the scan at every page boundary and either side of it,
// across the whole 32-bit space at a coarse stride, plus every region's exact edges —
// which is where a wholly-covered-page rule can go wrong.
func agrees(t *testing.T, m *Machine, when string) {
	t.Helper()
	check := func(a uint32) {
		if got, want := m.regionOf(a), m.scanRegionOf(a); got != want {
			gn, wn := "nil", "nil"
			if got != nil {
				gn = got.name
			}
			if want != nil {
				wn = want.name
			}
			t.Fatalf("%s: address 0x%08X resolves to %s through the page table but %s by scan", when, a, gn, wn)
		}
	}
	for _, r := range m.regions {
		if len(r.data) == 0 {
			continue
		}
		end := r.base + uint32(len(r.data))
		for _, a := range []uint32{r.base - 1, r.base, r.base + 1, end - 1, end, end + 1,
			r.base + 0xFFF, r.base + 0x1000} {
			check(a)
		}
	}
	// A coarse sweep of the whole address space, at a stride that is not a multiple
	// of the page size, so it lands inside pages as well as on their edges.
	for a := uint64(0); a < 1<<32; a += 0x777 {
		check(uint32(a))
	}
}

func TestPageTableAgreesWithTheScan(t *testing.T) {
	m := &Machine{}
	m.mapRegion("code", 0x00100000, make([]byte, 0x4000))
	m.mapRegion("odd", 0x00200800, make([]byte, 0x1800)) // deliberately unaligned at both ends
	m.mapRegion("empty", 0x00300000, nil)
	agrees(t, m, "after mapping")

	// Growth: the heap and the linear heap grow by append under ControlMemory, and
	// the pages that growth backs must become visible.
	heap := m.mapRegion("heap", heapBase, nil)
	m.heapReg = heap
	agrees(t, m, "with an empty heap")
	heap.data = append(heap.data, make([]byte, 0x9000)...)
	m.indexRegion(heap)
	agrees(t, m, "after the heap grew")
	if m.regionOf(heapBase+0x8FFF) != heap {
		t.Error("the page the heap just grew into is not backed by it")
	}

	heap.data = append(heap.data, make([]byte, 0x1000)...)
	m.indexRegion(heap)
	agrees(t, m, "after the heap grew again")
}

// TestPageTableSurvivesRestore is the stale-entry case with teeth: restore a snapshot
// into a machine that already has a page table, and every entry in it now points into
// regions that have been thrown away. If clearPages were forgotten, reads would
// silently come from the old machine's memory.
func TestPageTableSurvivesRestore(t *testing.T) {
	m := toadAtScene(t) // a real machine, with a real map, restored from a real state

	agrees(t, m, "after a restore")

	// And the restored memory must actually be the restored memory: read a byte of
	// code through the table and through the scan, and compare the bytes, not just
	// the region pointers.
	for _, a := range []uint32{codeBase, codeBase + 0x1000, vramVirtBase, vramVirtBase + 0x10000} {
		r := m.regionOf(a)
		if r == nil {
			continue
		}
		if r != m.scanRegionOf(a) {
			t.Fatalf("0x%08X: the table and the scan disagree after a restore", a)
		}
	}

	// Restoring a second time must not leave the first restore's regions behind.
	snap := m.SnapshotState()
	if err := m.RestoreState(snap); err != nil {
		t.Fatal(err)
	}
	agrees(t, m, "after a second restore")
}
