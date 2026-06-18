// webexport renders Turrican's disk-streamed level maps for the companion
// website's tile viewer. For each world it writes a tile atlas PNG (the world's
// 32x32 tiles in its palette) and, for each scene in that world, a JSON file with
// the column-major map flattened to a row-major cell grid. A meta.json lists every
// scene. The viewer resolves the flip flag (cell >= ntiles -> tile cell-128,
// horizontally flipped) just like extract/cmd/map.
//
// Usage: webexport [-o dir] [Turrican.adf]
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

	"turrican/extract/decrunch"
)

const (
	blockBase  = 0x1B980
	levelTable = 0x46A
	numWorlds  = 5
	tileSide   = 32
	tileBytes  = tileSide * 4 * (tileSide / 8) // 512
	atlasCols  = 16
)

type jsonLevel struct {
	World  int    `json:"world"`
	Scene  int    `json:"scene"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
	NTiles int    `json:"ntiles"`
	Atlas  string `json:"atlas"`
	Cells  []int  `json:"cells"` // row-major, raw map bytes (>=ntiles = flipped tile-128)
}

type metaLevel struct {
	Name  string `json:"name"`
	File  string `json:"file"`
	Atlas string `json:"atlas"`
}

func main() {
	out := flag.String("o", "site/public/turrican", "output directory")
	flag.Parse()
	adfPath := flag.Arg(0)
	if adfPath == "" {
		adfPath = "Turrican (Amiga)/Turrican.adf"
	}
	if err := run(adfPath, *out); err != nil {
		fmt.Fprintln(os.Stderr, "webexport:", err)
		os.Exit(1)
	}
}

func run(adfPath, outDir string) error {
	adf, err := os.ReadFile(adfPath)
	if err != nil {
		return err
	}
	res, err := decrunch.Decrunch(mainBlob(adf))
	if err != nil {
		return err
	}
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}

	var meta []metaLevel
	for w := 0; w < numWorlds; w++ {
		t := levelTable + w*8
		off := int(binary.BigEndian.Uint32(res.Data[t:]))
		length := int(binary.BigEndian.Uint32(res.Data[t+4:]))
		block, err := decrunch.DecrunchBlock(adf[off : off+length])
		if err != nil {
			return fmt.Errorf("world %d: %w", w, err)
		}

		be16 := func(o int) int { return int(binary.BigEndian.Uint16(block[o:])) }
		be32 := func(o int) int { return int(binary.BigEndian.Uint32(block[o:])) }
		at := func(addr int) int { return addr - blockBase }

		pal := readPalette(block, at(be32(0x08)))
		tableOff := at(be32(0x00))
		nTiles := be32(tableOff) / 4

		atlasName := fmt.Sprintf("atlas%d.png", w)
		if err := writeAtlas(filepath.Join(outDir, atlasName), block, tableOff, nTiles, pal); err != nil {
			return err
		}

		nScenes := be16(0x14) // header+$14 high word = scene count
		for s := 0; s < nScenes; s++ {
			descOff := at(be32(0x16 + s*4))
			mapOff := at(be32(descOff + 0x00))
			width := be16(descOff + 0x04)
			height := be16(descOff + 0x06)
			if width <= 0 || height <= 0 || mapOff+width*height > len(block) {
				return fmt.Errorf("world %d scene %d: bad map %dx%d", w, s, width, height)
			}
			cells := make([]int, width*height)
			for col := 0; col < width; col++ {
				for row := 0; row < height; row++ {
					cells[row*width+col] = int(block[mapOff+col*height+row]) // col-major -> row-major
				}
			}
			file := fmt.Sprintf("world%d_scene%d.json", w, s)
			if err := writeJSON(filepath.Join(outDir, file), jsonLevel{
				World: w, Scene: s, Width: width, Height: height,
				NTiles: nTiles, Atlas: atlasName, Cells: cells,
			}); err != nil {
				return err
			}
			meta = append(meta, metaLevel{
				Name: fmt.Sprintf("World %d · Scene %d", w+1, s+1),
				File: file, Atlas: atlasName,
			})
			fmt.Printf("world %d scene %d: %dx%d, %d tiles -> %s\n", w, s, width, height, nTiles, file)
		}
	}
	return writeJSON(filepath.Join(outDir, "meta.json"), map[string]any{"levels": meta})
}

func writeAtlas(path string, block []byte, tableOff, nTiles int, pal color.Palette) error {
	rows := (nTiles + atlasCols - 1) / atlasCols
	img := image.NewPaletted(image.Rect(0, 0, atlasCols*tileSide, rows*tileSide), pal)
	for n := 0; n < nTiles; n++ {
		off := tableOff + int(binary.BigEndian.Uint32(block[tableOff+n*4:]))
		if off+tileBytes > len(block) {
			break
		}
		ox, oy := (n%atlasCols)*tileSide, (n/atlasCols)*tileSide
		for y := 0; y < tileSide; y++ {
			var planes [4]uint32
			for p := 0; p < 4; p++ {
				planes[p] = binary.BigEndian.Uint32(block[off+(y*4+p)*4:])
			}
			for x := 0; x < tileSide; x++ {
				var v uint8
				for p := 0; p < 4; p++ {
					v |= uint8((planes[p]>>(31-uint(x)))&1) << uint(p)
				}
				img.SetColorIndex(ox+x, oy+y, v)
			}
		}
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

func readPalette(block []byte, off int) color.Palette {
	pal := make(color.Palette, 16)
	for i := range pal {
		c := binary.BigEndian.Uint16(block[off+i*2:])
		pal[i] = color.RGBA{R: uint8((c>>8)&0xF) * 17, G: uint8((c>>4)&0xF) * 17, B: uint8(c&0xF) * 17, A: 255}
	}
	return pal
}

func writeJSON(path string, v any) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}

func mainBlob(adf []byte) []byte {
	const off = 0x2C00
	n := int(binary.BigEndian.Uint32(adf[off:]))
	return adf[off : off+n]
}
