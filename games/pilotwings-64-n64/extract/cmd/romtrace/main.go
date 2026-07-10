// romtrace builds the RDRAM-to-cartridge asset map and verifies it.
//
// Input is a call log from a scratch boot (bootoracle -calllog 8022A760
// -calllog 80231A20): every call to the game's loader (dst, src, len — src
// below 0x80000000 is a cartridge offset fetched by 4 KiB-chunked PI DMA,
// src with bit 31 set is a RAM-to-RAM byte copy) and to its MIO0
// decompressor (fixed bounce buffer 0xDA808 in, 0x3DA800 out).
//
// The chain per asset is: load(0xDA800, cartOff, n) -> mio0(0xDA808, 0x3DA800)
// -> load(finalAddr, 0x3DA800+sliceOff, sliceLen). romtrace replays that chain
// from the ROM alone — our own mio0 package does the decompression — and
// compares every copied slice byte-for-byte against an RDRAM snapshot, so the
// map is proven rather than inferred.
//
// Usage:
//
//	romtrace -image ROM -calllog FILE -ram SNAP.bin [-lo ADDR -hi ADDR]
package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"

	"retroreverse.com/games/pilotwings-64-n64/extract/mio0"
	"retroreverse.com/tools/platform/n64"
)

const (
	loaderPC = 0x8022A760
	mio0PC   = 0x80231A20

	bouncePhys = 0x000DA800 // compressed blob staging
	outKSEG0   = 0x803DA800 // decompressor output buffer
)

type event struct {
	pc, field, a0, a1, a2, ra uint32
}

type slice struct {
	dst, off, length uint32
	blob             uint32 // cart offset of the MIO0 container, 0 = direct load
	direct           uint32 // cart offset for a direct (uncompressed) load
	field            uint32
}

func main() {
	image := flag.String("image", "", "cartridge image")
	callLog := flag.String("calllog", "", "call log from bootoracle -calllog 8022A760 -calllog 80231A20")
	ramFile := flag.String("ram", "", "RDRAM snapshot to verify against")
	lo := flag.Uint64("lo", 0, "only report copies whose destination is at or above this address")
	hi := flag.Uint64("hi", 0x400000, "only report copies whose destination is below this address")
	flag.Parse()

	rom, err := n64.Load(*image)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	ram, err := os.ReadFile(*ramFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	events, err := readLog(*callLog)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// Replay the chain: remember the last blob staged at the bounce buffer;
	// a decompress binds it to the output buffer; copies out of the output
	// buffer inherit it.
	var lastStage uint32 // cart offset most recently loaded to the bounce buffer
	var curBlob uint32   // cart offset of the blob now sitting decompressed in the output buffer
	var slices []slice
	for _, e := range events {
		switch e.pc {
		case mio0PC:
			curBlob = lastStage
		case loaderPC:
			if e.ra == 0x8022A7D4 { // the loader's own 4 KiB chunking
				continue
			}
			dstPhys := e.a0 &^ 0x80000000
			switch {
			case e.a1 < 0x80000000 && dstPhys == bouncePhys:
				lastStage = e.a1
			case e.a1 >= outKSEG0 && e.a1 < outKSEG0+0x100000:
				slices = append(slices, slice{
					dst: dstPhys, off: e.a1 - outKSEG0, length: e.a2,
					blob: curBlob, field: e.field,
				})
			case e.a1 < 0x80000000 && e.a2 > 4:
				// A direct cartridge load to its final address: verbatim bytes.
				slices = append(slices, slice{
					dst: dstPhys, length: e.a2, direct: e.a1, field: e.field,
				})
			}
		}
	}

	// Decompress each referenced blob once, from the ROM.
	blobs := map[uint32][]byte{}
	for _, s := range slices {
		if s.blob == 0 || blobs[s.blob] != nil {
			continue
		}
		off := int(s.blob) // cart offset == ROM file offset
		if off+16 > len(rom.Data) {
			fmt.Fprintf(os.Stderr, "blob %07X out of ROM range\n", s.blob)
			continue
		}
		// 8-byte outer container (fourCC + payload size), then the MIO0 image.
		out, err := mio0.Decompress(rom.Data[off+8:])
		if err != nil {
			fmt.Fprintf(os.Stderr, "blob %07X (%q): %v\n", s.blob, rom.Data[off:off+4], err)
			continue
		}
		blobs[s.blob] = out
	}

	// Verify each slice against the snapshot. A later load may overwrite an
	// earlier one at the same address; report each copy on its own line and
	// let the field column tell the story.
	sort.SliceStable(slices, func(i, j int) bool { return slices[i].dst < slices[j].dst })
	fmt.Printf("%-9s %-9s %-6s %-28s %-6s %s\n", "RDRAM", "end", "field", "source", "slice", "snapshot match")
	var okBytes, totBytes int
	for _, s := range slices {
		if uint64(s.dst) < *lo || uint64(s.dst) >= *hi {
			continue
		}
		var src []byte
		var srcDesc string
		if s.blob != 0 {
			b := blobs[s.blob]
			if b == nil || int(s.off+s.length) > len(b) {
				fmt.Printf("%08X  %08X  %5d  blob %07X+%04X  UNAVAILABLE\n", s.dst, s.dst+s.length, s.field, s.blob, s.off)
				continue
			}
			src = b[s.off : s.off+s.length]
			srcDesc = fmt.Sprintf("%s %07X (mio0) +%04X", string(rom.Data[s.blob:s.blob+4]), s.blob, s.off)
		} else {
			if int(s.direct)+int(s.length) > len(rom.Data) {
				continue
			}
			src = rom.Data[s.direct : s.direct+s.length]
			srcDesc = fmt.Sprintf("cart %07X (direct)", s.direct)
		}
		same := 0
		for i := range src {
			if int(s.dst)+i < len(ram) && ram[s.dst+uint32(i)] == src[i] {
				same++
			}
		}
		okBytes += same
		totBytes += len(src)
		verdict := "IDENTICAL"
		if same != len(src) {
			verdict = fmt.Sprintf("%d/%d", same, len(src))
		}
		fmt.Printf("%08X  %08X  %5d  %-28s %5X  %s\n", s.dst, s.dst+s.length, s.field, srcDesc, s.length, verdict)
	}
	fmt.Printf("\n%d/%d bytes identical across %d copies\n", okBytes, totBytes, len(slices))
}

func readLog(path string) ([]event, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var out []event
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 1024*1024), 1024*1024)
	for sc.Scan() {
		line := sc.Text()
		if !strings.HasPrefix(line, "call ") {
			continue
		}
		var e event
		if _, err := fmt.Sscanf(line, "call %x field=%d a0=%x a1=%x a2=%x a3=%x ra=%x",
			&e.pc, &e.field, &e.a0, &e.a1, &e.a2, new(uint32), &e.ra); err != nil {
			continue
		}
		out = append(out, e)
	}
	return out, sc.Err()
}
