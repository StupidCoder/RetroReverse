package gc

// disc.go reads a GameCube disc image.
//
// The image is raw: no sector headers, no error-correction layer, no ISO 9660. It is
// the drive's linear address space written straight to a file, and everything the
// machine reads is at the byte offset the disc's own structures name. Four regions
// matter, and they sit at fixed offsets in the first ten kilobytes:
//
//	0x0000  boot.bin    the disc header: game ID, magic, and where everything else is
//	0x0440  bi2.bin     the disc's second header: country, debug flags, memory sizing
//	0x2440  apploader   the loader, as PowerPC code, with a header naming its entry
//	                    (its body is what the console's IPL runs; see ipl.go)
//	varies  the DOL      the game's executable      (boot.bin names the offset)
//	varies  the FST      the filesystem             (boot.bin names the offset)
//
// The apploader is the reason this machine needs no BIOS. The console's boot ROM does
// not know how to read a game; it knows how to read *this*, and this knows how to read
// the game. Everything the loader then does — finding the executable, walking the
// filesystem, asking the drive for sectors — is the disc's own code, and we can watch
// it happen.

import (
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

// Where the fixed structures live.
const (
	bootBinOff   = 0x0000 // the disc header
	bootBinSize  = 0x0440
	bi2Off       = 0x0440 // the second header
	bi2Size      = 0x2000
	apploaderOff = 0x2440 // the loader's header; its code follows at +0x20

	// GameMagic marks a GameCube disc, at offset 0x1C of the header. (A Wii disc
	// carries a different magic at 0x18 and leaves this word zero, which is how the
	// two are told apart.)
	GameMagic = 0xC2339F3D
)

// Header is boot.bin: the disc's own account of itself.
type Header struct {
	GameID     string // "GLME01": four-byte game code + two-byte maker code
	DiscID     byte   // which disc of a multi-disc set
	Version    byte
	Streaming  bool
	Title      string // the game's name, as the disc gives it
	DOLOffset  uint32 // where the executable is
	FSTOffset  uint32 // where the filesystem is
	FSTSize    uint32
	FSTMaxSize uint32 // the largest FST across a multi-disc set — what to reserve
	FSTAddr    uint32 // where the loader is asked to put the FST in memory
	UserOffset uint32
	UserSize   uint32
}

// Apploader is the header of the disc's loader.
//
// Size and TrailerSize are two halves of one blob: the console reads Size+TrailerSize
// bytes of code and data starting at apploaderOff+0x20, puts them at ApploaderBase,
// and jumps to Entry. Why the split exists is the loader's own business; to the machine
// they are one contiguous copy.
type Apploader struct {
	Date        string // "2001/08/09" — the build date, in ASCII
	Entry       uint32
	Size        uint32
	TrailerSize uint32
}

// Body is the number of bytes of the loader to copy into memory.
func (a Apploader) Body() int { return int(a.Size) + int(a.TrailerSize) }

// Disc is an opened disc image.
type Disc struct {
	Path      string
	Size      int64
	Header    Header
	Apploader Apploader
	FST       *FST

	f   *os.File
	md5 string // computed on demand: hashing 1.4 GB is not free
}

// Open reads a disc image's headers and filesystem. The image stays open: a GameCube
// disc is well over a gigabyte, and the machine streams from it exactly as the drive
// does rather than holding it in memory.
func Open(path string) (*Disc, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	st, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, err
	}
	d := &Disc{Path: path, Size: st.Size(), f: f}

	boot, err := d.Read(bootBinOff, bootBinSize)
	if err != nil {
		f.Close()
		return nil, fmt.Errorf("reading the disc header: %w", err)
	}
	if got := be32(boot[0x1C:]); got != GameMagic {
		f.Close()
		return nil, fmt.Errorf("not a GameCube disc: magic at 0x1C is %#08X, want %#08X", got, GameMagic)
	}
	d.Header = Header{
		GameID:     string(boot[0x00:0x06]),
		DiscID:     boot[0x06],
		Version:    boot[0x07],
		Streaming:  boot[0x08] != 0,
		Title:      cstr(boot[0x20:0x400]),
		DOLOffset:  be32(boot[0x420:]),
		FSTOffset:  be32(boot[0x424:]),
		FSTSize:    be32(boot[0x428:]),
		FSTMaxSize: be32(boot[0x42C:]),
		FSTAddr:    be32(boot[0x430:]),
		UserOffset: be32(boot[0x434:]),
		UserSize:   be32(boot[0x438:]),
	}

	ap, err := d.Read(apploaderOff, 0x20)
	if err != nil {
		f.Close()
		return nil, fmt.Errorf("reading the apploader header: %w", err)
	}
	d.Apploader = Apploader{
		Date:        cstr(ap[0x00:0x10]),
		Entry:       be32(ap[0x10:]),
		Size:        be32(ap[0x14:]),
		TrailerSize: be32(ap[0x18:]),
	}

	fst, err := d.Read(int64(d.Header.FSTOffset), int(d.Header.FSTSize))
	if err != nil {
		f.Close()
		return nil, fmt.Errorf("reading the filesystem: %w", err)
	}
	d.FST, err = ParseFST(fst)
	if err != nil {
		f.Close()
		return nil, err
	}
	return d, nil
}

// Close releases the image.
func (d *Disc) Close() error { return d.f.Close() }

// Read returns n bytes at a byte offset. A read that runs off the end of the image is
// an error rather than a short read: the drive would have failed, and a caller that
// silently accepted zeroes would be building on sand.
func (d *Disc) Read(off int64, n int) ([]byte, error) {
	if off < 0 || n < 0 || off+int64(n) > d.Size {
		return nil, fmt.Errorf("read of %d bytes at %#x is outside the %d-byte image", n, off, d.Size)
	}
	b := make([]byte, n)
	if _, err := d.f.ReadAt(b, off); err != nil {
		return nil, err
	}
	return b, nil
}

// ApploaderCode is the loader's body: the PowerPC the console's boot ROM runs.
func (d *Disc) ApploaderCode() ([]byte, error) {
	return d.Read(apploaderOff+0x20, d.Apploader.Body())
}

// DOL reads and parses the game's executable.
func (d *Disc) DOL() (*DOL, error) {
	// The header names the executable's offset but not its length; the length is
	// whatever its own segment table reaches.
	hdr, err := d.Read(int64(d.Header.DOLOffset), dolHeaderSize)
	if err != nil {
		return nil, err
	}
	n := dolLength(hdr)
	b, err := d.Read(int64(d.Header.DOLOffset), n)
	if err != nil {
		return nil, err
	}
	return ParseDOL(b)
}

// MD5 is the image's hash, computed on first use.
func (d *Disc) MD5() (string, error) {
	if d.md5 != "" {
		return d.md5, nil
	}
	if _, err := d.f.Seek(0, io.SeekStart); err != nil {
		return "", err
	}
	h := md5.New()
	if _, err := io.Copy(h, d.f); err != nil {
		return "", err
	}
	d.md5 = fmt.Sprintf("%x", h.Sum(nil))
	return d.md5, nil
}

func be32(b []byte) uint32 { return binary.BigEndian.Uint32(b) }

// cstr reads a NUL-terminated string out of a fixed-width field.
func cstr(b []byte) string {
	for i, c := range b {
		if c == 0 {
			return string(b[:i])
		}
	}
	return string(b)
}
