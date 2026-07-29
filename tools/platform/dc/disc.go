package dc

// disc.go opens a GD-ROM rip: a cdrdao-style .cue naming one .bin of raw
// 2352-byte sectors (cue.go), the high-density data area anchored at the
// absolute LBA its own first sector header names (45000 on every retail GD),
// and an ISO 9660 volume whose directory extents are absolute — the reader
// therefore serves absolute LBAs, and the PVD is opened at session+16 via
// iso9660.OpenVolumeAt.
//
// A .gdi rip is a different TOC syntax over the same sector framing; teaching
// cue.go its three-column format is the natural extension when such a dump
// arrives.

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"retroreverse.com/tools/lib/iso9660"
)

const rawSector = 2352

// Disc is an opened GD-ROM image.
type Disc struct {
	Path   string
	Tracks []Track
	IP     IPBin
	Vol    *iso9660.Volume

	ra   io.ReaderAt
	size int64
	data Track // the high-density data track the filesystem lives in
	md5  string
}

// OpenDisc opens a .cue (cdrdao TOC) and its .bin.
func OpenDisc(cuePath string) (*Disc, error) {
	text, err := os.ReadFile(cuePath)
	if err != nil {
		return nil, err
	}
	bin, tracks, err := parseCue(string(text))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", cuePath, err)
	}
	f, err := os.Open(filepath.Join(filepath.Dir(cuePath), bin))
	if err != nil {
		return nil, err
	}
	st, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, err
	}
	d, err := openDisc(f, st.Size(), tracks)
	if err != nil {
		f.Close()
		return nil, err
	}
	d.Path = cuePath
	return d, nil
}

// openDisc anchors the tracks against the actual file and opens the volume.
func openDisc(ra io.ReaderAt, size int64, tracks []Track) (*Disc, error) {
	d := &Disc{Tracks: tracks, ra: ra, size: size}
	// Lengths from the next track's offset; the TOC's own lengths lie on
	// truncated dumps.
	for i := range d.Tracks {
		end := size
		if i+1 < len(d.Tracks) {
			end = d.Tracks[i+1].FileOffset
		}
		d.Tracks[i].Length = end - d.Tracks[i].FileOffset
	}
	// Anchor every data track at the absolute LBA its first sector header
	// declares (BCD MSF, 150-frame pregap offset).
	for i := range d.Tracks {
		t := &d.Tracks[i]
		if !t.IsData() {
			continue
		}
		var hdr [16]byte
		if _, err := ra.ReadAt(hdr[:], t.FileOffset); err != nil {
			return nil, fmt.Errorf("disc: reading track %d header: %w", t.Number, err)
		}
		lba, err := headerLBA(hdr[:])
		if err != nil {
			return nil, fmt.Errorf("disc: track %d: %w", t.Number, err)
		}
		t.StartLBA = lba
		d.data = *t // the last data track is the high-density area
	}
	if d.data.Mode == "" {
		return nil, fmt.Errorf("disc: no data track in TOC")
	}

	vol, err := iso9660.OpenVolumeAt(d, d.data.StartLBA+16)
	if err != nil {
		return nil, err
	}
	d.Vol = vol

	sec0, err := d.ReadBlock(d.data.StartLBA)
	if err != nil {
		return nil, err
	}
	ip, err := parseIPBin(sec0)
	if err != nil {
		return nil, err
	}
	d.IP = ip
	return d, nil
}

// headerLBA decodes a raw sector's 16-byte header: 12-byte sync, BCD
// minute/second/frame, mode.
func headerLBA(hdr []byte) (int, error) {
	sync := []byte{0x00, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0x00}
	if string(hdr[:12]) != string(sync) {
		return 0, fmt.Errorf("no raw-sector sync (got % X)", hdr[:12])
	}
	m, err1 := bcd(hdr[12])
	s, err2 := bcd(hdr[13])
	f, err3 := bcd(hdr[14])
	if err1 != nil || err2 != nil || err3 != nil {
		return 0, fmt.Errorf("bad BCD MSF % X", hdr[12:15])
	}
	return (m*60+s)*75 + f - 150, nil
}

func bcd(b byte) (int, error) {
	if b>>4 > 9 || b&0xF > 9 {
		return 0, fmt.Errorf("bad BCD byte %02X", b)
	}
	return int(b>>4)*10 + int(b&0xF), nil
}

// track finds the track containing an absolute LBA.
func (d *Disc) track(lba int) (Track, bool) {
	for _, t := range d.Tracks {
		if !t.IsData() || t.StartLBA < 0 {
			continue
		}
		if n := int(t.Length / rawSector); lba >= t.StartLBA && lba < t.StartLBA+n {
			return t, true
		}
	}
	return Track{}, false
}

// ReadBlock serves iso9660.BlockSource: 2048 user bytes by *absolute* LBA,
// de-framed from the raw sector.
func (d *Disc) ReadBlock(n int) ([]byte, error) {
	t, ok := d.track(n)
	if !ok {
		return nil, fmt.Errorf("disc: LBA %d is not in any data track", n)
	}
	buf := make([]byte, iso9660.BlockSize)
	off := t.FileOffset + int64(n-t.StartLBA)*rawSector + 16
	if _, err := d.ra.ReadAt(buf, off); err != nil {
		return nil, fmt.Errorf("disc: reading LBA %d: %w", n, err)
	}
	return buf, nil
}

// ReadSector is ReadBlock under the name the GD-ROM syscall layer uses.
func (d *Disc) ReadSector(lba int) ([]byte, error) { return d.ReadBlock(lba) }

// DataLBA is the absolute LBA the high-density data area starts at.
func (d *Disc) DataLBA() int { return d.data.StartLBA }

// MD5 hashes the whole .bin, caching the answer; it pins savestates to the
// disc they were taken on.
func (d *Disc) MD5() (string, error) {
	if d.md5 != "" {
		return d.md5, nil
	}
	h := md5.New()
	if _, err := io.Copy(h, io.NewSectionReader(d.ra, 0, d.size)); err != nil {
		return "", err
	}
	d.md5 = hex.EncodeToString(h.Sum(nil))
	return d.md5, nil
}

// BootFilePath is the boot binary's name as an ISO path (the IP.BIN field is
// space-padded and versionless).
func (d *Disc) BootFilePath() string { return strings.TrimSpace(d.IP.BootFile) }
