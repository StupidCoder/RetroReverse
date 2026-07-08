// Archive-resident files (Part V §2): internal file IDs with bit 15 set resolve
// through the archive table the resolver at $020186C0 walks — `CMP id, #$8000`,
// then over a 13-entry descriptor array at ARM9 $0208ECF4 (stride $14, loop bound
// #$D at $0201874C): each descriptor carries the archive's flagged-ID range at
// +$08/+$0A (u16 first/end) and its filesystem path ("/ARCHIVE/arc0.narc", …);
// the member index is `id - first`. The shipped ranges step by $400:
// $8000=arc0, $8400=en1, $8800..$9400=vs1-4, $9800=c2d, $9C00=ar1, $A000+=per-language.
package sm64ds

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"retroreverse.com/tools/platform/nds"
)

const arcTableOff = 0x8ACF4 // ARM9 file offset of the descriptor array ($0208ECF4)

// ArchiveRef is one resolved flagged file ID.
type ArchiveRef struct {
	Archive string // e.g. "arc0"
	Member  int
}

// Stem names the extracted model file for a flagged ID, e.g. "arc0_21".
func (r ArchiveRef) Stem() string { return fmt.Sprintf("%s_%d", r.Archive, r.Member) }

// ResolveArchiveID maps a bit-15-flagged internal file ID to its archive+member
// via the game's descriptor table.
func (ls *LevelSet) ResolveArchiveID(id int) (ArchiveRef, bool) {
	if id < 0x8000 {
		return ArchiveRef{}, false
	}
	for i := 0; i < 13; i++ {
		e := arcTableOff + i*0x14
		first := int(le.Uint16(ls.arm9[e+8:]))
		end := int(le.Uint16(ls.arm9[e+10:]))
		if id < first || id >= end {
			continue
		}
		// the descriptor carries two string pointers: a short name and the full
		// "/ARCHIVE/<name>.narc" path — use the path, keyed by its basename
		for _, off := range []int{0xC, 0x10} {
			p := le.Uint32(ls.arm9[e+off:])
			if p < 0x02004000 || p >= 0x02004000+uint32(len(ls.arm9)) {
				continue
			}
			o := int(p - 0x02004000)
			q := o
			for q < len(ls.arm9) && ls.arm9[q] != 0 && q-o < 32 {
				q++
			}
			s := string(ls.arm9[o:q])
			if strings.HasPrefix(s, "/ARCHIVE/") && strings.HasSuffix(s, ".narc") {
				return ArchiveRef{Archive: strings.TrimSuffix(filepath.Base(s), ".narc"), Member: id - first}, true
			}
		}
		return ArchiveRef{}, false
	}
	return ArchiveRef{}, false
}

// ArchiveMember returns a member's (decompressed) bytes.
func (ls *LevelSet) ArchiveMember(ref ArchiveRef) ([]byte, error) {
	d, err := readExtractedFile(ls.extDir, "files/ARCHIVE/"+ref.Archive+".narc")
	if err != nil {
		return nil, err
	}
	files, err := nds.ParseNARCFiles(d)
	if err != nil {
		return nil, err
	}
	if ref.Member < 0 || ref.Member >= len(files) {
		return nil, fmt.Errorf("sm64ds: %s member %d out of range", ref.Archive, ref.Member)
	}
	m := files[ref.Member].Data
	if len(m) > 4 && string(m[:4]) == "LZ77" {
		m = nds.Decompress(m[4:])
	}
	return m, nil
}

// PlausibleBMD sanity-checks the fixed header before a full decode (archive
// members are a mixed bag of models, animations and textures).
func PlausibleBMD(d []byte) bool {
	if len(d) < 0x40 {
		return false
	}
	if le.Uint32(d) > 16 { // scale shift
		return false
	}
	for _, o := range []int{4, 0xC, 0x14, 0x1C, 0x24} {
		n, p := le.Uint32(d[o:]), le.Uint32(d[o+4:])
		if n > 512 || p > uint32(len(d)) {
			return false
		}
	}
	return true
}

func readExtractedFile(extDir, rel string) ([]byte, error) {
	return os.ReadFile(filepath.Join(extDir, filepath.FromSlash(strings.TrimPrefix(rel, "/"))))
}
