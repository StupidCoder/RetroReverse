// roomexport decodes a room straight out of the ROM — no emulator, no running
// game — and checks the result against the VRAM export the oracle produced.
//
// The starting area is four LZ77 streams (addresses traced with bootoracle
// -declog, and the expander at $0801AB40 read with disarm):
//
//	$08385E68 -> the room's metatile-ID grid
//	$08381E9C -> the metatile table, 8 bytes (2x2 tile entries) per id
//	$0836E448 -> the BG2 tileset (character data)
//	$08370C18 -> the BG1 tileset
//
// A room is NOT stored as a tilemap: it is a grid of metatile ids, each
// expanding to a 2x2 block of tile entries. That is why searching a ROM for
// something tilemap-shaped finds nothing.
//
//	roomexport [-verify DIR] [-out DIR]
package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"

	"retroreverse.com/tools/platform/gba"
)

func die(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, "roomexport: "+format+"\n", a...)
	os.Exit(2)
}

func writePNG(path string, img image.Image) {
	f, err := os.Create(path)
	if err != nil {
		die("%v", err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		die("%v", err)
	}
}

// stream decompresses one LZ77 stream at a cartridge address.
func stream(rom []byte, addr uint32, what string) []byte {
	off := gba.ROMOffset(addr)
	if off >= len(rom) {
		die("%s: address %08X is past the end of the ROM", what, addr)
	}
	out, err := gba.LZ77Decompress(rom[off:])
	if err != nil {
		die("%s at %08X: %v", what, addr, err)
	}
	fmt.Printf("%-16s %08X -> %6d bytes\n", what, addr, len(out))
	return out
}

func main() {
	romPath := flag.String("rom", "../Legend of Zelda, The - The Minish Cap (USA).gba", "cartridge image")
	out := flag.String("out", "room", "output directory")
	verify := flag.String("verify", "", "a mapexport directory to check the ROM decode against")
	widthMeta := flag.Int("width", 63, "room width in metatiles (the ROM array is packed at this width; the game copies it into a 64-wide RAM buffer)")
	// The starting area ships as two layers, each its own id grid + metatile
	// table, and they pair up: the terrain layer's ids only make sense with the
	// terrain layer's table. Crossing them decodes to plausible-looking garbage.
	// Three 16 KiB tileset streams land back to back in VRAM at 0x00000,
	// 0x04000 and 0x08000. A 4bpp tile index is 10 bits, so a layer addresses
	// 1024 tiles = 32 KiB = TWO streams from its character base. Handing a
	// renderer only the first one silently draws garbage for every index above
	// 511 — which looks like a decode bug and is not one.
	listAddr := flag.String("list", "081034D0", "the room's asset list (tilesets + metatile-id grids)")
	metaAddr := flag.String("mlist", "0810279C", "the metatile-table asset list (shared between areas)")
	tlistAddr := flag.String("tlist", "08100E88", "the tileset asset list (a room list often replaces only the banks that changed)")
	scan := flag.Bool("scan", false, "enumerate every asset list in the ROM that loads a room map, and exit")
	ts0 := flag.String("tiles0", "0836E448", "tileset stream at VRAM 0x00000")
	ts1 := flag.String("tiles1", "08370C18", "tileset stream at VRAM 0x04000")
	ts2 := flag.String("tiles2", "083734F0", "tileset stream at VRAM 0x08000")
	flag.Parse()

	rom, err := os.ReadFile(*romPath)
	if err != nil {
		die("%v", err)
	}
	hex := func(s string) uint32 {
		var v uint32
		if _, err := fmt.Sscanf(s, "%x", &v); err != nil {
			die("bad address %q", s)
		}
		return v
	}

	if err := os.MkdirAll(*out, 0o755); err != nil {
		die("%v", err)
	}

	// The palette is not part of the compressed set; borrow the one the oracle
	// captured so the render is in the game's own colours.
	var pal gba.Palette
	if *verify != "" {
		if p, err := os.ReadFile(filepath.Join(*verify, "palette.bin")); err == nil {
			pal = gba.ParsePalette(p)
		}
	}
	if pal == nil {
		pal = make(gba.Palette, 512)
		for i := range pal {
			g := uint16(i & 31)
			pal[i] = g | g<<5 | g<<10
		}
	}

	if *scan {
		lists := ScanRoomLists(rom)
		fmt.Printf("%d asset lists load a room map:\n", len(lists))
		for i, a := range lists {
			l, err := ParseAssetList(rom, a)
			if err != nil {
				continue
			}
			var ids, tsets int
			for _, e := range l {
				if e.Dst == 0x02025EB4 || e.Dst == 0x0200B654 {
					ids++
				}
				if e.Dst>>24 == 6 {
					tsets++
				}
			}
			fmt.Printf("  [%3d] %08X  %d entries (%d id grids, %d tilesets)\n", i, a, len(l), ids, tsets)
		}
		return
	}

	roomList, err := ParseAssetList(rom, hex(*listAddr))
	if err != nil {
		die("room list: %v", err)
	}
	metaList, err := ParseAssetList(rom, hex(*metaAddr))
	if err != nil {
		die("metatile list: %v", err)
	}
	tilesetList, err := ParseAssetList(rom, hex(*tlistAddr))
	if err != nil {
		die("tileset list: %v", err)
	}
	assets, err := Resolve(roomList, metaList, tilesetList)
	if err != nil {
		die("%v", err)
	}
	fmt.Printf("room list %s: ids %08X/%08X, tables %08X/%08X, tilesets %08X/%08X/%08X\n",
		*listAddr, assets.IDs[0], assets.IDs[1], assets.Tables[0], assets.Tables[1],
		assets.Tilesets[0], assets.Tilesets[1], assets.Tilesets[2])

	// Decompress the three tileset streams once; each layer takes the pair its
	// character base spans.
	vram := make([]byte, 0, 3*16384)
	for i, a := range assets.Tilesets {
		if a == 0 {
			vram = append(vram, make([]byte, 16384)...) // this room does not replace this bank
			continue
		}
		vram = append(vram, stream(rom, a, fmt.Sprintf("tileset %d", i))...)
	}
	_ = ts0
	_ = ts1
	_ = ts2

	var rendered []*image.RGBA
	for _, l := range []struct {
		name          string
		ids, tab      uint32
		charBaseIndex int // which 16 KiB block this layer's character base sits at
	}{
		{"bg2", assets.IDs[0], assets.Tables[0], 0},
		{"bg1", assets.IDs[1], assets.Tables[1], 1},
	} {
		if l.ids == 0 || l.tab == 0 {
			fmt.Printf("%s: not present in this room\n", l.name)
			continue
		}
		ids := stream(rom, l.ids, l.name+" ids")
		table := stream(rom, l.tab, l.name+" table")
		chars := vram[l.charBaseIndex*16384:]

		room := gba.Room{WidthMeta: *widthMeta, HeightMeta: len(ids) / 2 / *widthMeta, Table: table}
		room.IDs = make([]uint16, len(ids)/2)
		for i := range room.IDs {
			room.IDs[i] = binary.LittleEndian.Uint16(ids[i*2:])
		}
		tiles, wt, ht, flagged := room.Expand()
		fmt.Printf("%s: %dx%d metatiles -> %dx%d tiles (%d ids above the 0x%X flag)\n",
			l.name, room.WidthMeta, room.HeightMeta, wt, ht, flagged, gba.MetatileFlag)

		raw := make([]byte, len(tiles)*2)
		for i, v := range tiles {
			binary.LittleEndian.PutUint16(raw[i*2:], v)
		}
		os.WriteFile(filepath.Join(*out, l.name+"-room-map.bin"), raw, 0o644)
		os.WriteFile(filepath.Join(*out, l.name+"-room-tiles.bin"), chars, 0o644)
		img := gba.RenderTiles(tiles, wt, ht, chars, pal, 0, 4)
		rendered = append(rendered, img)
		writePNG(filepath.Join(*out, l.name+"-room.png"), img)
		if *verify != "" {
			verifyAgainstVRAM(*verify, l.name, tiles, wt, ht)
		}
	}
	// The composite: BG1 over BG2, which is how the game stacks them (BG1 is the
	// overlay Link walks behind). BG2 alone shows dark blocks under every tree —
	// those are the shadow tiles the canopy covers, not a decode error.
	if len(rendered) == 2 {
		b := rendered[0].Bounds()
		comp := image.NewRGBA(b)
		draw.Draw(comp, b, rendered[0], b.Min, draw.Src)
		draw.Draw(comp, b, rendered[1], b.Min, draw.Over)
		writePNG(filepath.Join(*out, "room-composite.png"), comp)
	}
	fmt.Printf("wrote %s/{bg2,bg1}-room.png and room-composite.png\n", *out)
}

// verifyAgainstVRAM slides the oracle's VRAM window over the ROM-decoded map
// and reports the best alignment.
//
// It scores only the entries the game had ACTUALLY UPLOADED. A 32x32
// screenblock is taller than the 20 visible tile rows, so its bottom rows are
// still zero when the export is taken; counting those as mismatches turns a
// perfect decode into "71.9%", which reads exactly like a decoder that is
// nearly right — the most expensive kind of wrong answer. Unwritten entries are
// reported separately rather than hidden.
func verifyAgainstVRAM(dir, layer string, tiles []uint16, wt, ht int) {
	want, err := os.ReadFile(filepath.Join(dir, layer+"-map.bin"))
	if err != nil {
		fmt.Println("verify: nothing to compare against:", err)
		return
	}
	const w, h = 32, 32
	bestX, bestY, best, bestWritten := -1, -1, -1, 0
	for oy := 0; oy+h <= ht; oy++ {
		for ox := 0; ox+w <= wt; ox++ {
			match, written := 0, 0
			for y := 0; y < h; y++ {
				for x := 0; x < w; x++ {
					e := binary.LittleEndian.Uint16(want[(y*w+x)*2:])
					if e == 0 {
						continue // the game has not uploaded this cell
					}
					written++
					if e == tiles[(oy+y)*wt+ox+x] {
						match++
					}
				}
			}
			if match > best {
				best, bestWritten, bestX, bestY = match, written, ox, oy
			}
		}
	}
	pct := 0.0
	if bestWritten > 0 {
		pct = 100 * float64(best) / float64(bestWritten)
	}
	fmt.Printf("verify %s: best alignment at room tile (%d,%d): %d/%d uploaded entries match (%.1f%%), %d of %d cells not yet uploaded\n",
		layer, bestX, bestY, best, bestWritten, pct, w*h-bestWritten, w*h)
	if best == bestWritten && bestWritten > 0 {
		fmt.Printf("verify %s: EXACT — the ROM decode reproduces every tile the game uploaded\n", layer)
	}
}
