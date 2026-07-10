// audiodump parses Pilotwings 64's music resources and reports their structure,
// validating every pointer against the sample table. It ties the three pieces
// together: the UVSX ".CTL"/".TBL" bank pair inside the IFF archive, the second
// instrument bank the game DMAs from the post-archive region at cart 0x62D460,
// and the "S1" sequence bank at 0x618B70 (31 songs). None of these offsets is
// guessed — they come from the loader-trace (bootoracle -dmalog) that shows the
// running game fetch exactly these cart addresses.
//
//	audiodump -image ROM
package main

import (
	"flag"
	"fmt"
	"os"

	"retroreverse.com/games/pilotwings-64-n64/extract/pwad"
	"retroreverse.com/tools/platform/n64"
	"retroreverse.com/tools/platform/n64/audio"
)

// Cart offsets the loader trace proved the game reads (see the package doc).
const (
	seqBankOff   = 0x618B70 // sequence bank ("S1"), 31 songs
	seqBankEnd   = 0x62D460 // the music bank starts where the last song ends
	musicBankOff = 0x62D460 // second ALBankFile ("B1"), the music instrument bank
	musicTblOff  = 0x6314D0 // the music bank's own VADPCM sample table, right after its ".CTL"
)

func main() {
	image := flag.String("image", "", "cartridge image (.z64/.v64/.n64)")
	flag.Parse()
	if *image == "" {
		fmt.Fprintln(os.Stderr, "audiodump: -image is required")
		os.Exit(2)
	}
	rom, err := n64.Load(*image)
	if err != nil {
		die(err)
	}

	// --- UVSX: the ".CTL"/".TBL" bank pair from the archive ---
	ar, err := pwad.Open(rom.Data)
	if err != nil {
		die(err)
	}
	var ctl, tbl []byte
	for _, i := range ar.ByType("UVSX") {
		f, err := ar.Resource(i)
		if err != nil {
			die(err)
		}
		for _, c := range f.Chunks {
			b, err := ar.Data(c)
			if err != nil {
				die(err)
			}
			switch c.Tag {
			case ".CTL":
				ctl = b
			case ".TBL":
				tbl = b
			}
		}
	}
	if ctl == nil || tbl == nil {
		die(fmt.Errorf("UVSX .CTL/.TBL not found"))
	}
	fmt.Printf("UVSX: .CTL %d bytes, .TBL %d bytes\n", len(ctl), len(tbl))
	uvsx, err := audio.ParseBankFile(ctl)
	if err != nil {
		die(err)
	}
	reportBank("UVSX .CTL", uvsx, tbl)

	// --- the post-archive music instrument bank at 0x62D460 ---
	mbank, err := audio.ParseBankFile(rom.Data[musicBankOff:])
	if err != nil {
		die(fmt.Errorf("music bank @0x%x: %w", musicBankOff, err))
	}
	// The music bank has its own VADPCM sample table at musicTblOff, right after
	// its ".CTL" — the offset that reproduces every stored loop state.
	fmt.Printf("\nmusic .TBL at cart 0x%x\n", musicTblOff)
	reportBank(fmt.Sprintf("music bank @0x%x", musicBankOff), mbank, rom.Data[musicTblOff:])

	// --- the "S1" sequence bank at 0x618B70 ---
	sb, err := audio.ParseSeqBank(rom.Data[seqBankOff:seqBankEnd])
	if err != nil {
		die(fmt.Errorf("seq bank @0x%x: %w", seqBankOff, err))
	}
	fmt.Printf("\nsequence bank @0x%x: %d songs\n", seqBankOff, len(sb.Songs))
	for _, s := range sb.Songs {
		fmt.Printf("  song %2d: %6d bytes, channels %v\n", s.Index, len(s.Data), s.ActiveTracks())
	}
}

// loopOK reports whether one looping wavetable decodes to the decoder state the
// game stored at its loop point. The state is the decoded VADPCM frame that
// contains the loop start (aligned down to the 16-sample frame boundary).
func loopOK(w *audio.WaveTable, tbl []byte) (checked, ok bool) {
	if w == nil || w.Loop == nil || w.Book == nil {
		return false, false
	}
	frameStart := int(w.Loop.Start) - int(w.Loop.Start)%16
	if frameStart < 0 || int(w.Base) < 0 || int(w.Base)+int(w.Len) > len(tbl) {
		return false, false
	}
	pcm := audio.DecodeADPCM(tbl[w.Base:int(w.Base)+int(w.Len)], w.Book)
	if frameStart+16 > len(pcm) {
		return false, false
	}
	for i := 0; i < 16; i++ {
		if int16(pcm[frameStart+i]) != w.Loop.State[i] {
			return true, false
		}
	}
	return true, true
}

// eachLoopingWave calls fn for the first distinct looping wavetable of each base
// in the bank, stopping early if fn returns false.
func eachLoopingWave(bf *audio.BankFile, fn func(*audio.WaveTable) bool) {
	seen := map[int32]bool{}
	for _, b := range bf.Banks {
		insts := append([]*audio.Instrument{b.Percussion}, b.Instruments...)
		for _, in := range insts {
			if in == nil {
				continue
			}
			for _, s := range in.Sounds {
				w := s.Wave
				if w == nil || w.Loop == nil || w.Book == nil || seen[w.Base] {
					continue
				}
				seen[w.Base] = true
				if !fn(w) {
					return
				}
			}
		}
	}
}

// loopMatches counts looping samples reproducing their stored state, stopping
// after `limit` are found (limit<=0 means all) — a fast locator during search.
func loopMatches(bf *audio.BankFile, tbl []byte) int {
	matched := 0
	eachLoopingWave(bf, func(w *audio.WaveTable) bool {
		if _, ok := loopOK(w, tbl); ok {
			matched++
		}
		return true
	})
	return matched
}

// verifyLoops reports the bit-exact loop-state check over a whole bank.
func verifyLoops(name string, bf *audio.BankFile, tbl []byte) {
	checked, matched := 0, 0
	eachLoopingWave(bf, func(w *audio.WaveTable) bool {
		if c, ok := loopOK(w, tbl); c {
			checked++
			if ok {
				matched++
			}
		}
		return true
	})
	fmt.Printf("    VADPCM loop-state check: %d/%d looping samples reproduce the stored decoder state\n", matched, checked)
}

// reportBank prints a bank's shape and checks every ADPCM wavetable lands inside
// the sample table — the proof the two halves belong together.
func reportBank(name string, bf *audio.BankFile, tbl []byte) {
	for bi, b := range bf.Banks {
		nInst, nSound, nWave, oob := 0, 0, 0, 0
		var minBase, maxEnd int32 = 1 << 30, 0
		visit := func(in *audio.Instrument) {
			if in == nil {
				return
			}
			nInst++
			for _, s := range in.Sounds {
				nSound++
				if s.Wave == nil {
					continue
				}
				nWave++
				if s.Wave.Base < minBase {
					minBase = s.Wave.Base
				}
				if e := s.Wave.Base + s.Wave.Len; e > maxEnd {
					maxEnd = e
				}
				if s.Wave.Base < 0 || int(s.Wave.Base+s.Wave.Len) > len(tbl) {
					oob++
				}
			}
		}
		visit(b.Percussion)
		for _, in := range b.Instruments {
			visit(in)
		}
		fmt.Printf("%s bank %d: %d instruments, %d sounds, %d waves, rate %d Hz, percussion=%v\n",
			name, bi, nInst, nSound, nWave, b.SampleRate, b.Percussion != nil)
		fmt.Printf("    .TBL span used: 0x%x..0x%x of 0x%x   out-of-range waves: %d\n",
			minBase, maxEnd, len(tbl), oob)
	}
	verifyLoops(name, bf, tbl)
}

func die(err error) {
	fmt.Fprintln(os.Stderr, "audiodump:", err)
	os.Exit(1)
}
