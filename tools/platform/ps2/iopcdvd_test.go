package ps2

// iopcdvd_test.go pins the disc drive on the three points where being wrong was silent.
//
// Every one of them presented as the same thing: a game that read a sector correctly, put it
// in exactly the right place, and then said it could not find the file.

import (
	"encoding/binary"
	"testing"

	"retroreverse.com/tools/lib/iso9660"
)

// fakeDisc is a two-block volume: a valid Primary Volume Descriptor at LBA 16, and a sector
// whose contents we can recognise anywhere.
type fakeDisc struct{}

const (
	fakeDiscMarkerLBA = 4321
	fakeDiscMarker    = "this is the sector that was asked for"
)

func (fakeDisc) ReadBlock(n int) ([]byte, error) {
	b := make([]byte, iso9660.BlockSize)
	switch n {
	case 16:
		b[0] = 1
		copy(b[1:], "CD001")
		b[156] = 34                                  // the root directory record's length
		binary.LittleEndian.PutUint32(b[158:], 261)  // its extent
		binary.LittleEndian.PutUint32(b[166:], 2048) // its size
		binary.LittleEndian.PutUint32(b[80:], 1000)  // the volume's size, in blocks
	case fakeDiscMarkerLBA:
		copy(b, fakeDiscMarker)
	}
	return b, nil
}

func newDriveOverAFakeDisc(t *testing.T) (*Machine, *IOP) {
	t.Helper()
	vol, err := iso9660.OpenVolume(fakeDisc{})
	if err != nil {
		t.Fatalf("the fake disc is not a disc: %v", err)
	}
	m := NewMachine()
	m.SetVolume(vol)
	return m, m.StartIOP()
}

func TestTheDriveIsAByteDeviceAndTheWordPathMustNotTouchIt(t *testing.T) {
	// The S-command register at 0x2016 and its parameter FIFO at 0x2017 are two bytes of one
	// 32-bit word. The IOP's ordinary register path composes a word, merges a byte into it and
	// writes the whole word back — which for this device would push a phantom parameter into
	// the FIFO every time a command was issued, and re-issue the command every time a parameter
	// was pushed.
	//
	// CDVDMAN only ever uses `sb` and `lbu` here, so the drive is wired to the bus as the byte
	// device it is. This is the test that says so: two parameters go in, and the command sees
	// exactly those two and no others.
	m, p := newDriveOverAFakeDisc(t)
	_ = m

	p.Write(cdvdBase+cdvdSStatus, 0xAA) // a parameter
	p.Write(cdvdBase+cdvdSStatus, 0xBB) // and another
	p.Write(cdvdBase+cdvdSCommand, cdvdSCmdReady)

	if got := p.cdvd.sCommand; got != cdvdSCmdReady {
		t.Fatalf("the command register holds 0x%02X, not the command that was written", got)
	}
	// The command consumed its parameters, and the two it consumed were the two that were
	// pushed. A word-wide store would have left a third behind — the old contents of the
	// status byte, merged back in as data.
	if n := len(p.cdvd.sParams); n != 0 {
		t.Errorf("%d parameter(s) left over after the command ran: the FIFO is being written "+
			"by something other than the guest", n)
	}
}

func TestTheResultFIFOSaysWhenItIsEmpty(t *testing.T) {
	// Bit 6 of 0x2017 is the only thing driving CDVDMAN's two collect loops: the one that
	// drains stale bytes before a command and the one that reads the answer after it. A drive
	// that never says "empty" is one the module waits on for ever, and it did — 91,697 times
	// round a four-instruction loop, which from every other angle looks like a busy machine.
	_, p := newDriveOverAFakeDisc(t)

	if st := p.Read(cdvdBase + cdvdSStatus); st&cdvdSStatusNoData == 0 {
		t.Fatal("an idle drive with nothing to say does not report an empty result FIFO")
	}

	// S-command 5 answers with one byte. Reading it must empty the FIFO again.
	p.Write(cdvdBase+cdvdSCommand, cdvdSCmdReady)
	if st := p.Read(cdvdBase + cdvdSStatus); st&cdvdSStatusNoData != 0 {
		t.Fatal("the drive has an answer waiting and says its result FIFO is empty")
	}
	if v := p.Read(cdvdBase + cdvdSResult); v != 0 {
		t.Errorf("the drive answered 0x%02X to \"have you finished?\"; CDVDMAN polls for zero", v)
	}
	if st := p.Read(cdvdBase + cdvdSStatus); st&cdvdSStatusNoData == 0 {
		t.Error("the result FIFO was drained and still does not report itself empty")
	}
}

func TestADVDSectorCarriesItsOwnNumber(t *testing.T) {
	// The drive hands over whole 2064-byte physical sectors, and CDVDMAN checks the number in
	// each one's header against the sector it asked for: it reads bytes 1..3 as a big-endian
	// number, adds -0x30000, and expects the LBA back (CDVDMAN+0x3960).
	//
	// So the header does not hold the LBA. It holds the LBA plus 0x30000 — where a DVD's data
	// area begins — and a header of zeroes fails that check with the LBA negated, which is what
	// the drive did until this was found. The user data begins twelve bytes in.
	_, p := newDriveOverAFakeDisc(t)
	c := p.cdvd

	c.readSectors(fakeDiscMarkerLBA, 1)

	if n := len(c.data); n != cdvdRawSectorBytes {
		t.Fatalf("the drive staged %d bytes for one sector, not %d", n, cdvdRawSectorBytes)
	}
	sec := c.data

	// The number CDVDMAN will compute, computed its way.
	id := uint32(sec[1])<<16 | uint32(sec[2])<<8 | uint32(sec[3])
	if lba := id - cdvdDVDDataStart; lba != fakeDiscMarkerLBA {
		t.Errorf("CDVDMAN reads this sector's header as sector %d; it asked for %d",
			int32(lba), fakeDiscMarkerLBA)
	}
	if sec[0]&1 != 0 {
		t.Error("the header claims layer 1 on a single-layer disc")
	}
	if got := string(sec[cdvdSectorHeader : cdvdSectorHeader+len(fakeDiscMarker)]); got != fakeDiscMarker {
		t.Errorf("the user data is not twelve bytes into the sector; found %q", got)
	}
}

func TestTheReadErrorRegisterIsZeroWhenTheReadWorked(t *testing.T) {
	// 0x2006 reads back as the error the last command ended with, which is nothing like the
	// transfer mode that is written to it. CDVDMAN's interrupt handler files this byte, "what
	// was the last error" hands back exactly it, and a sector read fails unless it is zero.
	//
	// It used to answer with the command number. Every read on this disc therefore came back
	// "failed with error 8" — 8 being the read command — and the game reported that a file it
	// had just fetched correctly could not be found.
	_, p := newDriveOverAFakeDisc(t)

	params := make([]byte, 11)
	binary.LittleEndian.PutUint32(params[0:], fakeDiscMarkerLBA)
	binary.LittleEndian.PutUint32(params[4:], 1)
	for _, b := range params {
		p.Write(cdvdBase+cdvdNStatus, b)
	}
	p.Write(cdvdBase+cdvdNCommand, cdvdNCmdReadDVD)

	if e := p.Read(cdvdBase + cdvdNMode); e != 0 {
		t.Errorf("the drive read the sector and reports error %d", e)
	}
}
