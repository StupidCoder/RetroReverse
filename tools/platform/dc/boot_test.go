package dc

import (
	"encoding/binary"
	"testing"
)

func TestBootLoadsAndPoisons(t *testing.T) {
	d, err := OpenDisc(buildSyntheticGD(t))
	if err != nil {
		t.Fatal(err)
	}
	m := NewMachine(d)
	if err := m.Boot(); err != nil {
		t.Fatal(err)
	}
	if got := string(m.RAM[0x10000:0x10014]); got != "SYNTHETIC BOOTSTRAP!" {
		t.Fatalf("boot binary at 8C010000 reads %q", got)
	}
	if m.CPU.PC != 0x8C010000 {
		t.Fatalf("PC=%08X, want 8C010000", m.CPU.PC)
	}
	if v := m.ram32(0xBC); v != trapGdrom {
		t.Fatalf("gdrom vector holds %08X, want %08X", v, uint32(trapGdrom))
	}
	// Unplanted low RAM is poison, not zero and not garbage: F00D + offset.
	if v := m.ram32(0x2000); v != 0xF00D2000 {
		t.Fatalf("low RAM at 8C002000 = %08X, want the F00D2000 poison", v)
	}
}

// TestGDSyscall drives the documented command-queue protocol against the trap
// PCs the vectors point at: SEND files, CHECK says processing, EXEC_SERVER
// does the work, CHECK delivers — the sector lands in the game's buffer.
func TestGDSyscall(t *testing.T) {
	d, err := OpenDisc(buildSyntheticGD(t))
	if err != nil {
		t.Fatal(err)
	}
	m := NewMachine(d)
	if err := m.Boot(); err != nil {
		t.Fatal(err)
	}
	c := m.CPU

	call := func(vec uint32, r4, r5, r6, r7 uint32) uint32 {
		c.R[4], c.R[5], c.R[6], c.R[7] = r4, r5, r6, r7
		c.PR = 0x8C010000 // anywhere sane to return to
		c.SetPC(m.ram32(vec & 0xFFFF))
		if r := m.Run(1, RunConfig{NoSpin: true}); r.Reason == "halt: "+c.HaltReason && c.Halted {
			t.Fatalf("syscall halted: %s", c.HaltReason)
		}
		return c.R[0]
	}

	// params block: read 1 sector, FAD 45021+150, into 8C020000.
	params := uint32(0x8C030000)
	binary.LittleEndian.PutUint32(m.RAM[0x30000:], 45021+150)
	binary.LittleEndian.PutUint32(m.RAM[0x30004:], 1)
	binary.LittleEndian.PutUint32(m.RAM[0x30008:], 0x8C020000)

	id := call(0x8C0000BC, gdCmdPIORead, params, 0, 0) // SEND_COMMAND
	if id == 0 {
		t.Fatalf("SEND_COMMAND returned id 0")
	}
	status := uint32(0x8C030100)
	if got := call(0x8C0000BC, id, status, 0, 1); got != 1 {
		t.Fatalf("CHECK before EXEC_SERVER = %d, want 1 (processing)", got)
	}
	call(0x8C0000BC, 0, 0, 0, 2) // EXEC_SERVER
	if got := call(0x8C0000BC, id, status, 0, 1); got != 2 {
		t.Fatalf("CHECK after EXEC_SERVER = %d, want 2 (done)", got)
	}
	if got := string(m.RAM[0x20000:0x20014]); got != "SYNTHETIC BOOTSTRAP!" {
		t.Fatalf("PIOREAD delivered %q", got)
	}
}

// TestMapleDMA builds a two-frame command table (DEVINFO to port A, DEVINFO
// to port B) and starts the DMA: port A answers as a controller, port B as an
// empty socket, and the DMA-end interrupt fires.
func TestMapleDMA(t *testing.T) {
	m := NewMachine(nil)
	tbl, recvA, recvB := uint32(0x1000), uint32(0x2000), uint32(0x2100)
	put := func(off, v uint32) { binary.LittleEndian.PutUint32(m.RAM[off:], v) }
	// Frame 1: DEVINFO (cmd 1) to port A main (AP 0x20), no payload.
	put(tbl+0, 0)
	put(tbl+4, 0x0C000000+recvA)
	put(tbl+8, 1|0x20<<8|0x00<<16)
	// Frame 2 (last): DEVINFO to port B main (AP 0x60).
	put(tbl+12, 1<<31)
	put(tbl+16, 0x0C000000+recvB)
	put(tbl+20, 1|0x60<<8|0x00<<16)

	m.Write32(sbMDEN, 1)
	m.Write32(sbMDSTAR, 0x0C000000+tbl)
	m.Write32(sbMDST, 1)

	respA := binary.LittleEndian.Uint32(m.RAM[recvA:])
	if respA&0xFF != 5 {
		t.Fatalf("port A DEVINFO response cmd=%d, want 5", respA&0xFF)
	}
	// The FT word goes out raw (wire bytes 00 00 00 01): the game's driver
	// byte-reverses it and tests the CONTROLLER bit in the low byte — the
	// pre-swapped send made Crazy Taxi's per-field gate reject the pad the
	// moment enumeration succeeded.
	if fn := binary.LittleEndian.Uint32(m.RAM[recvA+4:]); fn != 0x01000000 {
		t.Fatalf("port A function mask %08X, want 01000000 (controller, wire order)", fn)
	}
	if respB := binary.LittleEndian.Uint32(m.RAM[recvB:]); respB != 0xFFFFFFFF {
		t.Fatalf("port B answered %08X, want FFFFFFFF (empty socket)", respB)
	}
	if m.Holly.ISTNRM&istMapleDMA == 0 {
		t.Fatalf("Maple DMA end interrupt not raised")
	}
}

// TestGETCONDReflectsPad: an injected button press shows up active-low in the
// condition response.
func TestGETCONDReflectsPad(t *testing.T) {
	m := NewMachine(nil)
	m.Pad.Buttons = 0xFFFF &^ PadStart // Start held
	tbl, recv := uint32(0x1000), uint32(0x2000)
	put := func(off, v uint32) { binary.LittleEndian.PutUint32(m.RAM[off:], v) }
	put(tbl+0, 1<<31)
	put(tbl+4, 0x0C000000+recv)
	put(tbl+8, 9|0x20<<8|0x00<<16|1<<24) // GET CONDITION (command 9), one payload word
	put(tbl+12, 0x00000001)              // the function queried, bus byte order

	m.Write32(sbMDEN, 1)
	m.Write32(sbMDSTAR, 0x0C000000+tbl)
	m.Write32(sbMDST, 1)

	cond := binary.LittleEndian.Uint32(m.RAM[recv+8:])
	if cond&uint32(PadStart) != 0 {
		t.Fatalf("condition %08X: Start bit still set, want cleared (active low)", cond)
	}
	if cond&uint32(PadA) == 0 {
		t.Fatalf("condition %08X: A reads pressed, but nothing pressed it", cond)
	}
}
