// decode is a standalone, pure-Go reimplementation of c/zzz's decryption: it
// turns an encrypted Marble Madness loader file (c/xxx, or c/MarbleMadness!.dat)
// back into a plain AmigaDOS hunk file. No emulator needed -- the copy-protection
// inputs c/zzz reads from live Kickstart 1.2 are baked in (captured via the
// FS-UAE GDB stub): every CPU exception/TRAP vector page byte = $FC, the launcher
// process tc_ExceptCode page = $FC and tc_TrapCode page = $FF.
//
// Format: the first longword ($000003F3 HUNK_HEADER) and every per-block TYPE
// longword are stored plaintext; HUNK_SYMBOL names and resident-library names
// are plaintext; every other structural field and every CODE/DATA body longword
// is XORed with an additive lagged-Fibonacci keystream over a 55-long key table.
//
// Usage: decode disk.adf c/xxx out.hunk
package main

import (
	"encoding/binary"
	"fmt"
	"os"

	"stupidcoder.com/tools/amiga/adf"
)

// --- key table: seed (x31 hash) + the copy-protection perturbation ---
func keyTable() []uint32 {
	t := make([]uint32, 55)
	t[0] = 0x57319753
	for i := uint32(1); i < 55; i++ {
		t[i] = t[i-1]*31 + i
	}
	// sub_DAA protection with the captured Kickstart-1.2 inputs.
	const vec = 0xFC // (vector>>16)&0xFF, identical for all of $8..$BC
	for i := 10; i <= 16; i++ {
		t[i] ^= vec // loop1: vectors $8..$20
	}
	for i := 10; i <= 14; i++ {
		t[i] ^= vec // loop2: vectors $28..$38
	}
	for i := 32; i <= 47; i++ {
		t[i] ^= vec // loop3: vectors $80..$BC
	}
	t[30] ^= 0xFC // byte_hi(tc_ExceptCode=$00FC2FB4)
	t[31] ^= 0x0F // byte_hi(tc_TrapCode=$00FF4B6A) >> 4
	return t
}

type dec struct {
	in  []byte
	out []byte
	pos int
	t   []uint32
	p, q int
}

func (d *dec) ks() uint32 { // sub_$EAC lagged-Fibonacci
	d.t[d.p] += d.t[d.q]
	d.p = (d.p + 1) % 55
	d.q = (d.q + 2) % 55
	return d.t[d.p]
}

func (d *dec) raw() uint32 { // plaintext longword (TYPE marker): copied as-is
	v := binary.BigEndian.Uint32(d.in[d.pos:])
	d.pos += 4
	return v
}

func (d *dec) xor() uint32 { // encrypted structural longword: decrypt + write
	c := binary.BigEndian.Uint32(d.in[d.pos:])
	pl := c ^ d.ks()
	binary.BigEndian.PutUint32(d.out[d.pos:], pl)
	d.pos += 4
	return pl
}

func (d *dec) rawN(longs uint32) { d.pos += int(longs) * 4 } // plaintext name bytes

const (
	hUNIT   = 0x3E7
	hNAME   = 0x3E8
	hCODE   = 0x3E9
	hDATA   = 0x3EA
	hBSS    = 0x3EB
	hRELOC32 = 0x3EC
	hSYMBOL = 0x3F0
	hDEBUG  = 0x3F1
	hEND    = 0x3F2
	hHEADER = 0x3F3
)

func main() {
	if len(os.Args) != 4 {
		fmt.Fprintln(os.Stderr, "usage: decode disk.adf innerpath out.hunk")
		os.Exit(2)
	}
	img, err := os.ReadFile(os.Args[1])
	chk(err)
	vol, err := adf.Open(img)
	chk(err)
	in, err := vol.ReadFile(os.Args[2])
	chk(err)

	d := &dec{in: in, out: append([]byte(nil), in...), t: keyTable(), p: 0, q: 27}

	magic := d.raw() & 0x3FFFFFFF
	if magic != hHEADER {
		fmt.Fprintf(os.Stderr, "not a HUNK_HEADER (got $%X)\n", magic)
		os.Exit(1)
	}
	// HUNK_HEADER: resident-library list, then table size/first/last/sizes.
	for {
		n := d.xor()
		if n == 0 {
			break
		}
		d.rawN(n) // library name (plaintext)
	}
	d.xor()             // table_size
	first := d.xor()
	last := d.xor()
	for i := first; i <= last; i++ {
		d.xor() // hunk size (top 2 bits = mem flags)
	}

	// per-block loop
	for d.pos+4 <= len(in) {
		typ := d.raw() & 0x3FFFFFFF
		switch typ {
		case hCODE, hDATA:
			size := d.xor() & 0x3FFFFFFF
			for j := uint32(0); j < size; j++ {
				d.xor()
			}
		case hBSS:
			d.xor()
		case hRELOC32:
			for {
				cnt := d.xor()
				if cnt == 0 {
					break
				}
				d.xor() // target hunk
				for j := uint32(0); j < cnt; j++ {
					d.xor() // offset
				}
			}
		case hSYMBOL:
			for {
				n := d.xor()
				if n == 0 {
					break
				}
				d.rawN(n) // symbol name (plaintext)
				d.xor()   // value
			}
		case hDEBUG:
			n := d.xor()
			for j := uint32(0); j < n; j++ {
				d.xor()
			}
		case hEND:
			// hunk terminator, continue to next hunk's first block
		case hUNIT, hNAME:
			// unit/name marker followed by name
			n := d.xor()
			d.rawN(n)
		default:
			fmt.Fprintf(os.Stderr, "stopped: unknown block type $%X at byte %d (ks words used: p=%d)\n", typ, d.pos-4, d.p)
			goto done
		}
	}
done:
	chk(os.WriteFile(os.Args[3], d.out, 0644))
	fmt.Printf("decoded %d bytes -> %s (consumed %d/%d)\n", len(d.out), os.Args[3], d.pos, len(in))
}

func chk(e error) {
	if e != nil {
		panic(e)
	}
}
