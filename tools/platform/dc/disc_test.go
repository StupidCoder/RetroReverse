package dc

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"
)

// rawSectorFor frames 2048 bytes of user data as a raw Mode-1 sector at an
// absolute LBA: sync, BCD MSF (150-frame pregap), mode byte, data, zeroed
// EDC/ECC.
func rawSectorFor(lba int, data []byte) []byte {
	s := make([]byte, rawSector)
	copy(s, []byte{0x00, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0x00})
	msf := lba + 150
	toBCD := func(v int) byte { return byte(v/10<<4 | v%10) }
	s[12] = toBCD(msf / (60 * 75))
	s[13] = toBCD(msf / 75 % 60)
	s[14] = toBCD(msf % 75)
	s[15] = 1
	copy(s[16:], data)
	return s
}

// buildSyntheticGD lays out a three-track image the way the Crazy Taxi dump
// is laid out: a small low-density data track at LBA 0, a truncated audio
// track, and a high-density area at LBA 45000 whose ISO 9660 extents are
// absolute. It returns the cue path.
func buildSyntheticGD(t *testing.T) string {
	t.Helper()
	const session = 45000
	var bin []byte

	// Track 1: two sectors at LBA 0.
	for i := 0; i < 2; i++ {
		bin = append(bin, rawSectorFor(i, []byte("low-density"))...)
	}
	// Track 2: audio, deliberately short (the TOC length lies about it).
	audioOff := len(bin)
	bin = append(bin, make([]byte, 2352)...)

	// Track 3: the high-density area.
	dataOff := len(bin)
	sec := map[int][]byte{}

	ip := make([]byte, 2048)
	pad := func(off int, s string, n int) {
		copy(ip[off:off+n], []byte(s + string(make([]byte, 0))))
		for i := len(s); i < n; i++ {
			ip[off+i] = ' '
		}
	}
	pad(0x00, "SEGA SEGAKATANA", 16)
	pad(0x10, "SEGA TESTWORKS", 16)
	pad(0x20, "0000 GD-ROM1/1", 16)
	pad(0x30, " U", 8)
	pad(0x38, "0799A10", 8)
	pad(0x40, "MK-00000", 10)
	pad(0x4A, "V9.999", 6)
	pad(0x50, "20260729", 16)
	pad(0x60, "1ST_READ.BIN", 16)
	pad(0x70, "SEGA TESTWORKS", 16)
	pad(0x80, "SYNTHETIC DISC", 128)
	sec[session] = ip

	pvd := make([]byte, 2048)
	pvd[0] = 1
	copy(pvd[1:6], "CD001")
	pvd[6] = 1
	copy(pvd[40:], "SYNTHETIC ")
	binary.LittleEndian.PutUint32(pvd[80:], 30) // session-relative size, as shipped
	binary.LittleEndian.PutUint16(pvd[128:], 2048)
	root := pvd[156 : 156+34]
	root[0] = 34
	binary.LittleEndian.PutUint32(root[2:], uint32(session+20)) // absolute!
	binary.LittleEndian.PutUint32(root[10:], 2048)
	root[25] = 2
	root[32] = 1
	sec[session+16] = pvd

	dir := make([]byte, 2048)
	name := "1ST_READ.BIN;1"
	rec := dir[0 : 33+len(name)+1]
	rec[0] = byte(len(rec))
	binary.LittleEndian.PutUint32(rec[2:], uint32(session+21)) // absolute
	binary.LittleEndian.PutUint32(rec[10:], 20)
	rec[32] = byte(len(name))
	copy(rec[33:], name)
	sec[session+20] = dir

	file := make([]byte, 2048)
	copy(file, "SYNTHETIC BOOTSTRAP!")
	sec[session+21] = file

	for lba := session; lba <= session+21; lba++ {
		d := sec[lba]
		if d == nil {
			d = make([]byte, 2048)
		}
		bin = append(bin, rawSectorFor(lba, d)...)
	}

	dir2 := t.TempDir()
	binPath := filepath.Join(dir2, "synthetic.bin")
	if err := os.WriteFile(binPath, bin, 0o644); err != nil {
		t.Fatal(err)
	}
	// The audio length in the TOC deliberately overstates the stored bytes,
	// like the real dump: only the # offsets may be believed.
	cue := "CD_ROM\n" +
		"// Track 1\nTRACK MODE1_RAW\nNO COPY\nDATAFILE \"synthetic.bin\" 00:08:00 // length in bytes: 4704\n" +
		"// Track 2\nTRACK AUDIO\nNO COPY\nDATAFILE \"synthetic.bin\" #" + itoa(audioOff) + " 09:52:00 // length in bytes: 104428800\n" +
		"// Track 3\nTRACK MODE1_RAW\nNO COPY\nDATAFILE \"synthetic.bin\" #" + itoa(dataOff) + " 112:02:00\n"
	cuePath := filepath.Join(dir2, "synthetic.cue")
	if err := os.WriteFile(cuePath, []byte(cue), 0o644); err != nil {
		t.Fatal(err)
	}
	return cuePath
}

func itoa(v int) string {
	if v == 0 {
		return "0"
	}
	var b []byte
	for v > 0 {
		b = append([]byte{byte('0' + v%10)}, b...)
		v /= 10
	}
	return string(b)
}

func TestSyntheticDisc(t *testing.T) {
	d, err := OpenDisc(buildSyntheticGD(t))
	if err != nil {
		t.Fatal(err)
	}
	if d.DataLBA() != 45000 {
		t.Fatalf("DataLBA=%d, want 45000 (anchored from the sector header, not the TOC)", d.DataLBA())
	}
	if d.IP.BootFile != "1ST_READ.BIN" || d.IP.Title != "SYNTHETIC DISC" {
		t.Fatalf("IP.BIN parsed as %+v", d.IP)
	}
	if got := d.Tracks[0].StartLBA; got != 0 {
		t.Fatalf("track 1 LBA=%d, want 0", got)
	}
	// The absolute-extent quirk: the volume resolves the file through
	// absolute LBAs served by the disc.
	data, err := d.Vol.ReadFile("1ST_READ.BIN")
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "SYNTHETIC BOOTSTRAP!" {
		t.Fatalf("boot file reads %q", data)
	}
	e, err := d.Vol.Resolve("1ST_READ.BIN")
	if err != nil {
		t.Fatal(err)
	}
	if e.Block != 45021 {
		t.Fatalf("boot extent LBA=%d, want the absolute 45021", e.Block)
	}
	// Raw sector access for the GD syscall path.
	b, err := d.ReadSector(45021)
	if err != nil {
		t.Fatal(err)
	}
	if string(b[:20]) != "SYNTHETIC BOOTSTRAP!" {
		t.Fatalf("ReadSector(45021) reads %q", b[:20])
	}
	// An LBA in no data track is an error, not a zero sector.
	if _, err := d.ReadSector(20000); err == nil {
		t.Fatalf("ReadSector(20000) succeeded; the audio gap must not read as data")
	}
}
