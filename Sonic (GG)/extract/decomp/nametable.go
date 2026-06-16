package decomp

// LoadRLE reimplements the name-table loader the engine calls at $0502 — a small
// run-length codec distinct from the $0406 tile codec. The stream is a list of tile
// bytes: normally one literal entry each, but when a tile byte equals the one just
// written it switches to a run — the duplicate is followed by a count byte and the
// tile is emitted that many more times. An $FF tile terminates. Every emitted entry
// is the pair (tile, hi), where hi is the shared high byte the engine takes from
// $D20F (palette select + priority bits). It returns the decoded name-table entries
// as 2-byte little-endian words (tile, hi), ready to drop into a name table.
//
// fileOff is the already-resolved file offset of the source; srcLen bounds how many
// source bytes to consume (the engine passes this in BC), independent of the $FF
// terminator — decoding stops at whichever comes first.
// LoadMapRLE reimplements the level-map decompressor at $0A73 — the codec that fills
// the row-major block-index window at RAM $C000 (16 rows x 256 cols = 4096 bytes) at
// level load. It is the same run-length family as LoadRLE ($0502) but for a raw byte
// stream rather than name-table entries, and differs in three ways:
//
//   - Output is one byte per entry (a block index), not a (tile,hi) pair.
//   - A run's count byte of $00 means 256, not 0 — the engine repeats with DJNZ, whose
//     B=0 loops the full 256 times.
//   - There is no $FF terminator: decoding is bounded solely by srcLen (the engine's BC,
//     which it decrements once per source byte consumed and stops at zero).
//
// As in LoadRLE the first byte of a run-pair is emitted as a literal and the run emits
// `count` MORE copies, and after a run the "previous byte" is re-armed to a value that
// can never match (the $0A76 CPL) so the byte immediately after a run starts fresh.
// Verified byte-perfect against the live decompressor's output (Green Hills, Zone 0):
// source bank 5 file $17430, srcLen $0786 (1926 bytes) -> 4096 bytes, 2.13x.
func LoadMapRLE(rom []byte, fileOff, srcLen int) []byte {
	out := []byte{}
	hl := fileOff
	end := fileOff + srcLen
	bc := srcLen
	rd := func(i int) byte { // bounds-safe source read
		if i >= 0 && i < len(rom) {
			return rom[i]
		}
		return 0
	}
	for {
		prev := ^rd(hl) // $0A76: prev = ~src[HL], a value the next byte can't match
		litDone := false
		for { // $0A7B literal loop
			a := rd(hl)
			if a == prev { // -> run
				break
			}
			out = append(out, a)
			prev = a
			hl++
			bc--
			if bc == 0 || hl >= end {
				litDone = true
				break
			}
		}
		if litDone {
			break
		}
		bc-- // $0A8E DEC BC (the duplicate byte)
		if bc == 0 {
			break
		}
		a := rd(hl)
		hl++
		n := int(rd(hl)) // count byte
		if n == 0 {
			n = 256
		}
		for k := 0; k < n; k++ {
			out = append(out, a)
		}
		hl++ // past the count byte
		bc--
		if bc == 0 || hl >= end {
			break
		}
	}
	return out
}

func LoadRLE(rom []byte, fileOff, srcLen int, hi byte) []byte {
	out := []byte{}
	i := fileOff
	end := fileOff + srcLen
	prev := ^rom[i] // sentinel: the first byte can never match, so it is a literal
	for i < end && i < len(rom) {
		b := rom[i]
		switch {
		case b == prev: // run: this duplicate + the next byte (a count) repeat the tile
			i++
			if i >= len(rom) || i >= end {
				return out
			}
			n := int(rom[i])
			for k := 0; k < n; k++ {
				out = append(out, b, hi)
			}
			i++
			if i < len(rom) {
				prev = ^rom[i] // re-arm the sentinel so the next byte is a literal
			}
		case b == 0xFF: // end of stream
			return out
		default: // literal
			out = append(out, b, hi)
			prev = b
			i++
		}
	}
	return out
}
