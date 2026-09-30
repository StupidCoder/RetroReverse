package knowledge

// Bounded reader adapters follow the repository's n3ds and iso9660 layouts.
// They deliberately do not call full-image parsers: metadata traversal and all
// allocations must remain within the resolver's limits, even on malformed media.
import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strings"
	"unicode/utf16"
)

var le = binary.LittleEndian
var be = binary.BigEndian

func rangeView(v assetView, off, n uint64) (assetView, []SourceExtent, error) {
	x, e := v.slice(off, n)
	return x, []SourceExtent{{off, n}}, e
}
func (r *Resolver) selectMember(v assetView, l Location) (assetView, []SourceExtent, error) {
	bad := func(s string) (assetView, []SourceExtent, error) {
		return assetView{}, nil, fmt.Errorf("%s: %s", l.Decoder, s)
	}
	switch l.Decoder {
	case "ncsd":
		b, e := r.read(v, 0, 512)
		if e != nil {
			return v, nil, e
		}
		if string(b[256:260]) != "NCSD" || len(l.Path) != 10 || !strings.HasPrefix(l.Path, "partition") || l.Path[9] < '0' || l.Path[9] > '7' {
			return bad("invalid header/partition name")
		}
		i := int(l.Path[9] - '0')
		off := uint64(le.Uint32(b[0x120+i*8:])) * 512
		n := uint64(le.Uint32(b[0x124+i*8:])) * 512
		if n == 0 {
			return bad("empty partition")
		}
		return rangeView(v, off, n)
	case "ncch":
		b, e := r.read(v, 0, 512)
		if e != nil {
			return v, nil, e
		}
		if string(b[256:260]) != "NCCH" || b[0x18f]&4 == 0 || b[0x18e] > 8 {
			return bad("unsupported/encrypted NCCH")
		}
		field := 0x1b0
		if l.Path == "exefs" {
			field = 0x1a0
		} else if l.Path != "romfs" {
			return bad("unknown region")
		}
		unit := uint64(512) << b[0x18e]
		return rangeView(v, uint64(le.Uint32(b[field:]))*unit, uint64(le.Uint32(b[field+4:]))*unit)
	case "sarc":
		return r.sarc(v, l.Path)
	case "romfs":
		return r.romfs(v, l.Path)
	case "iso9660":
		return r.iso(v, l.Path)
	}
	return bad("unsupported decoder")
}
func (r *Resolver) sarc(v assetView, name string) (assetView, []SourceExtent, error) {
	b, e := r.read(v, 0, v.size)
	if e != nil {
		return v, nil, e
	}
	bad := func() (assetView, []SourceExtent, error) {
		return v, nil, fmt.Errorf("invalid SARC structure while selecting %q", name)
	}
	if len(b) < 40 || string(b[:4]) != "SARC" || le.Uint16(b[4:]) != 20 || le.Uint16(b[6:]) != 0xfeff || uint64(le.Uint32(b[8:])) != v.size || string(b[20:24]) != "SFAT" || le.Uint16(b[24:]) != 12 {
		return bad()
	}
	n := int(le.Uint16(b[26:]))
	data := uint64(le.Uint32(b[12:]))
	sfnt := 32 + n*16
	if sfnt+8 > len(b) || string(b[sfnt:sfnt+4]) != "SFNT" || le.Uint16(b[sfnt+4:]) != 8 || data < uint64(sfnt+8) || data > v.size {
		return bad()
	}
	key := le.Uint32(b[28:])
	matches := 0
	var off, size uint64
	names := map[string]bool{}
	for i := 0; i < n; i++ {
		if e := r.entry(); e != nil {
			return v, nil, e
		}
		node := b[32+i*16:]
		attrs := le.Uint32(node[4:])
		at := uint64(sfnt+8) + uint64(attrs&0xffffff)*4
		if attrs>>24 != 1 || at >= data {
			return bad()
		}
		end := bytes.IndexByte(b[at:data], 0)
		if end < 0 || end > 1024 {
			return bad()
		}
		s := string(b[at : at+uint64(end)])
		if names[s] {
			return v, nil, resolutionError("ambiguous", "duplicate SARC member: "+s)
		}
		if !assetPath(s) {
			return bad()
		}
		names[s] = true
		var hash uint32
		for _, c := range []byte(s) {
			hash = hash*key + uint32(c)
		}
		if hash != le.Uint32(node) {
			return bad()
		}
		a, z := uint64(le.Uint32(node[8:])), uint64(le.Uint32(node[12:]))
		if z < a || z > v.size-data {
			return bad()
		}
		if s == name {
			matches++
			off = data + a
			size = z - a
		}
	}
	if matches != 1 {
		return v, nil, resolutionError("unavailable", "SARC member not found: "+name)
	}
	return rangeView(v, off, size)
}
func (r *Resolver) transform(v assetView, l Location) (assetView, error) {
	if l.Decoder != "yaz0" || l.Size > 64<<20 || r.allocated+l.Size > 128<<20 {
		return v, fmt.Errorf("transform allocation/decoder limit")
	}
	b, e := r.read(v, 0, v.size)
	if e != nil {
		return v, e
	}
	if len(b) < 16 || string(b[:4]) != "Yaz0" || uint64(be.Uint32(b[4:])) != l.Size {
		return v, fmt.Errorf("Yaz0 size/header mismatch")
	}
	r.allocated += l.Size
	out := make([]byte, 0, int(l.Size))
	src := 16
	for uint64(len(out)) < l.Size {
		if src >= len(b) {
			return v, fmt.Errorf("truncated Yaz0 control")
		}
		ctrl := b[src]
		src++
		for i := 0; i < 8 && uint64(len(out)) < l.Size; i++ {
			if ctrl&128 != 0 {
				if src >= len(b) {
					return v, fmt.Errorf("truncated Yaz0 literal")
				}
				out = append(out, b[src])
				src++
			} else {
				if src+2 > len(b) {
					return v, fmt.Errorf("truncated Yaz0 reference")
				}
				x, y := b[src], b[src+1]
				src += 2
				dist := int(x&15)<<8 | int(y)
				dist++
				n := int(x>>4) + 2
				if x>>4 == 0 {
					if src >= len(b) {
						return v, fmt.Errorf("truncated Yaz0 length")
					}
					n = int(b[src]) + 18
					src++
				}
				if dist > len(out) || n > int(l.Size)-len(out) {
					return v, fmt.Errorf("invalid Yaz0 reference")
				}
				for j := 0; j < n; j++ {
					out = append(out, out[len(out)-dist])
				}
			}
			ctrl <<= 1
		}
	}
	return assetView{uint64(len(out)), func(off uint64, b []byte) error {
		if off > uint64(len(out)) || uint64(len(b)) > uint64(len(out))-off {
			return fmt.Errorf("decoded read outside bounds")
		}
		copy(b, out[off:])
		return nil
	}}, nil
}
func (r *Resolver) romfs(v assetView, name string) (assetView, []SourceExtent, error) {
	bad := func() (assetView, []SourceExtent, error) {
		return v, nil, fmt.Errorf("invalid RomFS structure while selecting %q", name)
	}
	h, e := r.read(v, 0, 96)
	if e != nil {
		return v, nil, e
	}
	if string(h[:4]) != "IVFC" || le.Uint32(h[4:]) != 0x10000 {
		return bad()
	}
	master := uint64(le.Uint32(h[8:]))
	logs := []uint32{le.Uint32(h[28:]), le.Uint32(h[52:]), le.Uint32(h[76:])}
	for _, x := range logs {
		if x < 9 || x > 20 {
			return bad()
		}
	}
	align := func(n uint64, log uint32) uint64 { return (n + (1 << log) - 1) &^ ((1 << log) - 1) }
	l3 := align(96+master, logs[2])
	size3 := le.Uint64(h[68:])
	size1 := le.Uint64(h[20:])
	size2 := le.Uint64(h[44:])
	if master > v.size || size3 > v.size || size1 > v.size || size2 > v.size {
		return bad()
	}
	l1 := align(l3+size3, logs[0])
	l2 := align(l1+size1, logs[1])
	if align(l2+size2, logs[1]) != v.size {
		return bad()
	}
	hdr, e := r.read(v, l3, 40)
	if e != nil {
		return v, nil, e
	}
	if le.Uint32(hdr) != 40 {
		return bad()
	}
	table := func(offField int) ([]byte, error) {
		off, n := uint64(le.Uint32(hdr[offField:])), uint64(le.Uint32(hdr[offField+4:]))
		if n > 16<<20 || off > size3 || n > size3-off {
			return nil, fmt.Errorf("RomFS metadata exceeds bounds")
		}
		return r.read(v, l3+off, n)
	}
	dirs, e := table(12)
	if e != nil {
		return v, nil, e
	}
	files, e := table(28)
	if e != nil {
		return v, nil, e
	}
	data := uint64(le.Uint32(hdr[36:]))
	if data > size3 {
		return bad()
	}
	readName := func(b []byte, off uint32, head int) (string, error) {
		if uint64(off)+uint64(head) > uint64(len(b)) {
			return "", fmt.Errorf("RomFS metadata entry outside table")
		}
		n := le.Uint32(b[int(off)+head-4:])
		if n > 2048 || n%2 != 0 || uint64(off)+uint64(head)+uint64(n) > uint64(len(b)) {
			return "", fmt.Errorf("RomFS name bounds")
		}
		s := b[int(off)+head : int(off)+head+int(n)]
		u := make([]uint16, len(s)/2)
		for i := range u {
			u[i] = le.Uint16(s[i*2:])
		}
		text := string(utf16.Decode(u))
		back := utf16.Encode([]rune(text))
		if len(back) != len(u) {
			return "", fmt.Errorf("invalid UTF16")
		}
		for i := range u {
			if u[i] != back[i] {
				return "", fmt.Errorf("invalid UTF16")
			}
		}
		if !assetPath(text) || strings.Contains(text, "/") {
			return "", fmt.Errorf("invalid RomFS name")
		}
		return text, nil
	}
	dir := uint32(0)
	parts := strings.Split(name, "/")
	if len(parts) > 32 {
		return bad()
	}
	for index, part := range parts {
		if uint64(dir)+24 > uint64(len(dirs)) {
			return bad()
		}
		last := index == len(parts)-1
		ptr := le.Uint32(dirs[dir+8:])
		buf, head := dirs, 24
		if last {
			ptr = le.Uint32(dirs[dir+12:])
			buf, head = files, 32
		}
		seen := map[uint32]bool{}
		names := map[string]bool{}
		found := false
		var match uint32
		for ptr != 0xffffffff {
			if seen[ptr] {
				return bad()
			}
			seen[ptr] = true
			if e := r.entry(); e != nil {
				return v, nil, e
			}
			s, e := readName(buf, ptr, head)
			if e != nil {
				return v, nil, e
			}
			if names[s] {
				return v, nil, resolutionError("ambiguous", "duplicate RomFS name: "+s)
			}
			names[s] = true
			if le.Uint32(buf[ptr:]) != dir {
				return bad()
			}
			if s == part {
				found = true
				match = ptr
			}
			ptr = le.Uint32(buf[ptr+4:])
		}
		if !found {
			return v, nil, resolutionError("unavailable", "RomFS path not found: "+name)
		}
		if last {
			off, n := le.Uint64(files[match+8:]), le.Uint64(files[match+16:])
			if off > size3-data || n > size3-data-off {
				return bad()
			}
			return rangeView(v, l3+data+off, n)
		}
		dir = match
	}
	return bad()
}

// ISO 9660 uses exact on-disc identifiers, including ;1. Joliet/Rock Ridge,
// interleaving and extended attributes are rejected; multi-extent files are joined.
func (r *Resolver) iso(v assetView, name string) (assetView, []SourceExtent, error) {
	bad := func() (assetView, []SourceExtent, error) {
		return v, nil, fmt.Errorf("invalid/unsupported/ambiguous ISO9660 path %q", name)
	}
	p, e := r.read(v, 16*2048, 2048)
	if e != nil {
		return v, nil, e
	}
	if p[0] != 1 || string(p[1:6]) != "CD001" || p[6] != 1 || le.Uint16(p[128:]) != 2048 || be.Uint16(p[130:]) != 2048 {
		return bad()
	}
	blocks := uint64(le.Uint32(p[80:]))
	if be.Uint32(p[84:]) != uint32(blocks) || blocks*2048 > v.size {
		return bad()
	}
	record := func(b []byte) (SourceExtent, byte, string, error) {
		if len(b) < 34 || int(b[0]) != len(b) || int(b[32])+33 > len(b) || b[1] != 0 || b[26] != 0 || b[27] != 0 || b[25]&0x7c != 0 {
			return SourceExtent{}, 0, "", fmt.Errorf("unsupported ISO record")
		}
		off, n := le.Uint32(b[2:]), le.Uint32(b[10:])
		if off != be.Uint32(b[6:]) || n != be.Uint32(b[14:]) || uint64(off)*2048 > blocks*2048 || uint64(n) > blocks*2048-uint64(off)*2048 {
			return SourceExtent{}, 0, "", fmt.Errorf("invalid ISO extent")
		}
		return SourceExtent{uint64(off) * 2048, uint64(n)}, b[25], string(b[33 : 33+int(b[32])]), nil
	}
	if p[156] < 34 || 156+int(p[156]) > len(p) {
		return bad()
	}
	current, flags, _, e := record(p[156 : 156+int(p[156])])
	if e != nil || flags&2 == 0 {
		return bad()
	}
	parts := strings.Split(name, "/")
	if len(parts) > 32 {
		return bad()
	}
	for pi, part := range parts {
		if current.Length > 16<<20 {
			return bad()
		}
		b, e := r.read(v, current.Offset, current.Length)
		if e != nil {
			return v, nil, e
		}
		var ext []SourceExtent
		var lastFlags byte
		complete := false
		for pos := 0; pos < len(b); {
			if b[pos] == 0 {
				pos = (pos/2048 + 1) * 2048
				continue
			}
			n := int(b[pos])
			if n > 2048-pos%2048 || pos+n > len(b) {
				return bad()
			}
			if e := r.entry(); e != nil {
				return v, nil, e
			}
			x, f, s, e := record(b[pos : pos+n])
			if e != nil {
				return v, nil, e
			}
			pos += n
			if s == part {
				if complete {
					return v, nil, resolutionError("ambiguous", "duplicate ISO member: "+name)
				}
				if len(ext) > 0 && f&2 != lastFlags&2 {
					return bad()
				}
				ext = append(ext, x)
				lastFlags = f
				complete = f&128 == 0
			} else if len(ext) > 0 && !complete {
				return bad()
			}
		}
		if len(ext) == 0 {
			return v, nil, resolutionError("unavailable", "ISO member not found: "+name)
		}
		if !complete {
			return bad()
		}
		if pi < len(parts)-1 {
			if len(ext) != 1 || lastFlags&2 == 0 {
				return bad()
			}
			current = ext[0]
			continue
		}
		if lastFlags&2 != 0 {
			return bad()
		}
		var size uint64
		for _, x := range ext {
			size += x.Length
		}
		parent := v
		result := assetView{size, func(off uint64, b []byte) error {
			if off > size || uint64(len(b)) > size-off {
				return fmt.Errorf("ISO read bounds")
			}
			for _, x := range ext {
				if off >= x.Length {
					off -= x.Length
					continue
				}
				take := uint64(len(b))
				if take > x.Length-off {
					take = x.Length - off
				}
				if e := parent.read(x.Offset+off, b[:take]); e != nil {
					return e
				}
				b = b[take:]
				off = 0
				if len(b) == 0 {
					return nil
				}
			}
			return nil
		}}
		return result, ext, nil
	}
	return bad()
}
