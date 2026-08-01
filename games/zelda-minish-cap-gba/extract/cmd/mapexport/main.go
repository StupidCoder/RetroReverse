// mapexport lifts the backgrounds of whatever scene the machine is currently
// showing: for each enabled text-mode BG layer it writes the tilemap (raw and
// as JSON), the character data as a tile sheet, and the whole map rendered at
// its native size — plus the palette and a manifest naming every register the
// decode depended on.
//
// This is the VRAM-side export. It is the honest first step and not the last
// one: what it lifts is the scene the game has ALREADY decompressed into video
// memory, so it is bounded by what is on screen now, and a room the game has
// not loaded does not exist to it. Finding the ROM-side compressed map format
// (Part IV) is what turns this into "every room"; until then this is the
// ground truth that a ROM-side decoder will be checked against.
//
//	mapexport -state FILE [-out DIR]
//	mapexport -keys SCRIPT -frames N [-out DIR]
package main

import (
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"retroreverse.com/tools/platform/gba"
	"retroreverse.com/tools/platform/gba/gbamachine"
)

func die(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, "mapexport: "+format+"\n", a...)
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

// layerInfo is one BG's entry in the manifest.
type layerInfo struct {
	BG         int    `json:"bg"`
	BGCNT      string `json:"bgcnt"`
	Priority   int    `json:"priority"`
	CharBase   string `json:"char_base"`
	ScreenBase string `json:"screen_base"`
	BPP        int    `json:"bpp"`
	Mosaic     bool   `json:"mosaic"`
	SizeTiles  string `json:"size_tiles"`
	ScrollX    int    `json:"scroll_x"`
	ScrollY    int    `json:"scroll_y"`
	TilesUsed  int    `json:"distinct_tiles_used"`
	MaxTile    int    `json:"max_tile_index"`
}

type manifest struct {
	Frame   uint64      `json:"frame"`
	DISPCNT string      `json:"dispcnt"`
	Mode    int         `json:"mode"`
	Layers  []layerInfo `json:"layers"`
	Note    string      `json:"note"`
}

func main() {
	rom := flag.String("rom", "../Legend of Zelda, The - The Minish Cap (USA).gba", "cartridge image")
	state := flag.String("state", "", "savestate to export from")
	keys := flag.String("keys", "", "input script, if booting instead of loading a state")
	frames := flag.Uint64("frames", 0, "frames to run before exporting")
	out := flag.String("out", "export", "output directory")
	verify := flag.Bool("verify", false, "read the exported files back, recompose them, and compare against the machine's own background render")
	flag.Parse()

	data, err := os.ReadFile(*rom)
	if err != nil {
		die("%v", err)
	}
	cart, err := gba.Parse(data)
	if err != nil {
		die("%v", err)
	}
	m := gbamachine.New(cart)
	if *state != "" {
		if err := m.LoadState(*state); err != nil {
			die("%v", err)
		}
	}
	if *keys != "" {
		applyKeys(m, *keys)
	}
	if *frames > 0 {
		res := m.RunFrames(*frames, 999999999999)
		fmt.Println("run:", res.Reason)
	}
	if err := os.MkdirAll(*out, 0o755); err != nil {
		die("%v", err)
	}

	vram := m.Snapshot(0x06000000, 0x18000)
	palRAM := m.Snapshot(0x05000000, 0x400)
	pal := gba.ParsePalette(palRAM)

	dispcnt := m.Reg(0x000)
	mode := int(dispcnt & 7)
	if mode > 2 {
		die("display is in bitmap mode %d — there are no tilemaps to export", mode)
	}

	// The palette, as a 16x16 swatch grid (one row per 16-colour bank).
	swatch := image.NewRGBA(image.Rect(0, 0, 16*8, 32*8))
	for i := 0; i < 512 && i < len(pal); i++ {
		c := pal.RGBA(i)
		for y := 0; y < 8; y++ {
			for x := 0; x < 8; x++ {
				swatch.SetRGBA((i%16)*8+x, (i/16)*8+y, c)
			}
		}
	}
	writePNG(filepath.Join(*out, "palette.png"), swatch)
	os.WriteFile(filepath.Join(*out, "palette.bin"), palRAM, 0o644)

	man := manifest{
		Frame: m.Frame(), DISPCNT: fmt.Sprintf("0x%04X", dispcnt), Mode: mode,
		Note: "VRAM-side export: the scene the game had decompressed into video memory at this frame.",
	}

	// The screen as the PPU composed it, for provenance: an exported layer that
	// does not appear in this picture is a decode that went somewhere else.
	writePNG(filepath.Join(*out, "screen.png"), m.Screen())

	for bg := 0; bg < 4; bg++ {
		if dispcnt&(1<<uint(8+bg)) == 0 {
			continue
		}
		// Mode 1's BG2 and mode 2's BGs are AFFINE, a different map format;
		// this exporter covers the text-mode layers only.
		if (mode == 1 && bg == 2) || mode == 2 {
			fmt.Printf("bg%d: affine layer, not exported (text-mode only)\n", bg)
			continue
		}
		cnt := m.Reg(0x008 + 2*uint32(bg))
		layer := gba.DecodeBGCNT(cnt)
		bpp := 4
		if layer.EightBpp {
			bpp = 8
		}
		used, maxTile := layer.TilesUsed(vram)

		base := filepath.Join(*out, fmt.Sprintf("bg%d", bg))
		writePNG(base+"-map.png", layer.Render(vram, pal))
		banks, multiBank := layer.TilePalBanks(vram)
		// The sheet worth looking at is the tiles this map actually references,
		// each in its own palette bank; the full character range is mostly art
		// the current scene never mentions.
		sheet, order := gba.TileSheetUsed(vram, pal, layer.CharBase, bpp, used, banks)
		writePNG(base+"-tiles.png", sheet)
		writePNG(base+"-tiles-all.png", gba.TileSheetBanked(vram, pal, layer.CharBase, maxTile+1, bpp, banks))
		oj, _ := json.Marshal(order)
		os.WriteFile(base+"-tiles-order.json", oj, 0o644)

		// The tilemap itself, raw and decoded. The raw block is what the game's
		// own loader wrote; the JSON is the same thing readable.
		mapBytes := make([]byte, layer.WidthTiles*layer.HeightTiles*2)
		entries := make([][]int, layer.HeightTiles)
		for ty := 0; ty < layer.HeightTiles; ty++ {
			entries[ty] = make([]int, layer.WidthTiles)
			for tx := 0; tx < layer.WidthTiles; tx++ {
				e := layer.MapEntryAt(vram, tx, ty)
				entries[ty][tx] = e.Tile
				var raw uint16
				raw = uint16(e.Tile) | uint16(e.PalBank)<<12
				if e.HFlip {
					raw |= 1 << 10
				}
				if e.VFlip {
					raw |= 1 << 11
				}
				binary.LittleEndian.PutUint16(mapBytes[(ty*layer.WidthTiles+tx)*2:], raw)
			}
		}
		os.WriteFile(base+"-map.bin", mapBytes, 0o644)
		j, _ := json.Marshal(entries)
		os.WriteFile(base+"-map.json", j, 0o644)

		// The character data as the game stored it, so a ROM-side decoder has
		// something byte-exact to be checked against.
		tileBytes := (maxTile + 1) * (bpp * 8)
		if end := int(layer.CharBase) + tileBytes; end <= len(vram) {
			os.WriteFile(base+"-tiles.bin", vram[layer.CharBase:end], 0o644)
		}

		man.Layers = append(man.Layers, layerInfo{
			BG: bg, BGCNT: fmt.Sprintf("0x%04X", cnt), Priority: layer.Priority,
			CharBase:   fmt.Sprintf("0x%05X", layer.CharBase),
			ScreenBase: fmt.Sprintf("0x%05X", layer.ScrBase),
			BPP:        bpp, Mosaic: layer.Mosaic,
			SizeTiles: fmt.Sprintf("%dx%d", layer.WidthTiles, layer.HeightTiles),
			ScrollX:   int(m.Reg(0x010 + 4*uint32(bg))),
			ScrollY:   int(m.Reg(0x012 + 4*uint32(bg))),
			TilesUsed: len(used), MaxTile: maxTile,
		})
		fmt.Printf("bg%d: %s %dbpp, char 0x%05X, screen 0x%05X, %d distinct tiles (max %d), scroll %d,%d, %d tiles used with >1 palette bank\n",
			bg, fmt.Sprintf("%dx%d", layer.WidthTiles, layer.HeightTiles), bpp,
			layer.CharBase, layer.ScrBase, len(used), maxTile,
			m.Reg(0x010+4*uint32(bg)), m.Reg(0x012+4*uint32(bg)), multiBank)
	}

	if *verify {
		verifyExport(m, *out, man)
	}

	mj, _ := json.MarshalIndent(man, "", "  ")
	os.WriteFile(filepath.Join(*out, "manifest.json"), mj, 0o644)
	fmt.Printf("wrote %d layer(s) to %s/\n", len(man.Layers), *out)
}

// verifyExport is the acceptance test for everything above, and it deliberately
// works from the FILES rather than from the structs that wrote them: it reads
// the exported layer PNGs back off disk, recomposes them with the scroll and
// priority the manifest recorded, and compares the result against the machine's
// own background-only render. Re-rendering from the in-memory decode would
// prove only that the decode agrees with itself — the repository has shipped an
// unloadable asset that way before.
func verifyExport(m *gbamachine.Machine, dir string, man manifest) {
	want := m.RenderLayers(false) // backgrounds only: the exporter emits no sprites

	type layerImg struct {
		img      *image.RGBA
		priority int
		sx, sy   int
	}
	var layers []layerImg
	for _, li := range man.Layers {
		f, err := os.Open(filepath.Join(dir, fmt.Sprintf("bg%d-map.png", li.BG)))
		if err != nil {
			die("verify: %v", err)
		}
		src, err := png.Decode(f)
		f.Close()
		if err != nil {
			die("verify: %v", err)
		}
		rgba, ok := src.(*image.RGBA)
		if !ok {
			rgba = image.NewRGBA(src.Bounds())
			for y := src.Bounds().Min.Y; y < src.Bounds().Max.Y; y++ {
				for x := src.Bounds().Min.X; x < src.Bounds().Max.X; x++ {
					r, g, b, a := src.At(x, y).RGBA()
					rgba.Set(x, y, color.RGBA{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8), uint8(a >> 8)})
				}
			}
		}
		layers = append(layers, layerImg{rgba, li.Priority, li.ScrollX, li.ScrollY})
	}
	// Lower priority number draws in front; equal priority resolves by BG index,
	// which is the order the manifest already lists them in.
	sort.SliceStable(layers, func(i, j int) bool { return layers[i].priority < layers[j].priority })

	backdrop := want.RGBAAt(0, 0) // sampled below where nothing is opaque
	pf, err := os.Open(filepath.Join(dir, "palette.png"))
	if err == nil {
		if pimg, err := png.Decode(pf); err == nil {
			r, g, b, a := pimg.At(0, 0).RGBA()
			backdrop = color.RGBA{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8), uint8(a >> 8)}
		}
		pf.Close()
	}

	got := image.NewRGBA(want.Bounds())
	for y := 0; y < 160; y++ {
		for x := 0; x < 240; x++ {
			out := backdrop
			for _, l := range layers {
				w, h := l.img.Bounds().Dx(), l.img.Bounds().Dy()
				px, py := (x+l.sx)%w, (y+l.sy)%h
				if c := l.img.RGBAAt(px, py); c.A != 0 {
					out = c
					break // front-most opaque layer wins
				}
			}
			got.SetRGBA(x, y, out)
		}
	}
	writePNG(filepath.Join(dir, "verify-composed.png"), got)
	writePNG(filepath.Join(dir, "verify-reference.png"), want)

	diff := 0
	for y := 0; y < 160; y++ {
		for x := 0; x < 240; x++ {
			if got.RGBAAt(x, y) != want.RGBAAt(x, y) {
				diff++
			}
		}
	}
	total := 240 * 160
	fmt.Printf("verify: recomposed the exported PNGs — %d/%d pixels differ (%.2f%%)\n",
		diff, total, 100*float64(diff)/float64(total))
	if diff > 0 {
		fmt.Println("verify: see verify-composed.png vs verify-reference.png")
	}
}

// applyKeys installs a frame-scripted input sequence (same syntax as bootoracle).
func applyKeys(m *gbamachine.Machine, spec string) {
	names := map[string]uint16{"a": 1, "b": 2, "select": 4, "start": 8,
		"right": 16, "left": 32, "up": 64, "down": 128, "r": 256, "l": 512}
	type press struct {
		frame uint64
		mask  uint16
	}
	var script []press
	for _, part := range strings.Split(spec, ",") {
		fs := strings.SplitN(part, ":", 2)
		if len(fs) != 2 {
			die("bad -keys entry %q", part)
		}
		f, err := strconv.ParseUint(fs[0], 10, 64)
		if err != nil {
			die("bad frame in %q", part)
		}
		var mask uint16
		for _, k := range strings.Split(fs[1], "+") {
			mask |= names[strings.ToLower(strings.TrimSpace(k))]
		}
		script = append(script, press{f, mask})
	}
	sort.Slice(script, func(i, j int) bool { return script[i].frame < script[j].frame })
	m.OnFrame = func() {
		f := m.Frame()
		var mask uint16
		for _, p := range script {
			if f >= p.frame && f < p.frame+20 {
				mask |= p.mask
			}
		}
		m.SetKeys(mask)
	}
}
