// binsurvey sweeps every .bin member of every room arc on the disc and
// reports which of the 21 section slots are used, what GX texture formats
// appear, which attribute masks the meshes carry, and what the materials'
// stage lists reference — the reconnaissance for finishing the .bin decode.
package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"

	"retroreverse.com/games/luigis-mansion-gc/extract/lm"
	"retroreverse.com/tools/platform/gc"
)

func main() {
	image := flag.String("image", "", "disc image")
	flag.Parse()
	d, err := gc.Open(*image)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer d.Close()

	secFiles := map[int]map[string]bool{} // section -> files using it
	fmtCount := map[uint8]int{}
	maskCount := map[uint16]int{}
	palSamplers := 0
	multiStage := map[string]int{}
	files := 0

	for _, f := range d.FST.Entries {
		if f.Dir || !strings.HasPrefix(f.Path, "/Iwamoto/") || !strings.HasSuffix(f.Path, ".arc") {
			continue
		}
		b, err := d.Read(f.Offset, int(f.Size))
		if err != nil {
			continue
		}
		if len(b) >= 4 && string(b[:4]) == "Yay0" {
			if b, err = lm.Yay0(b); err != nil {
				continue
			}
		}
		members, err := lm.RARC(b)
		if err != nil {
			continue
		}
		for _, mem := range members {
			if !strings.HasSuffix(mem.Name, ".bin") || len(mem.Data) < 0x60 {
				continue
			}
			files++
			tag := f.Path + ":" + mem.Name
			u32 := func(o int) uint32 { return binary.BigEndian.Uint32(mem.Data[o:]) }
			u16 := func(o int) int { return int(binary.BigEndian.Uint16(mem.Data[o:])) }
			var offs [21]uint32
			for i := range offs {
				offs[i] = u32(0xC + i*4)
				if offs[i] != 0 {
					if secFiles[i] == nil {
						secFiles[i] = map[string]bool{}
					}
					secFiles[i][tag] = true
				}
			}
			end := func(sec int) uint32 {
				best := uint32(len(mem.Data))
				for _, o := range offs {
					if o > offs[sec] && o < best {
						best = o
					}
				}
				return best
			}
			// texture formats
			if t := offs[0]; t != 0 {
				firstData := uint32(len(mem.Data)) - t
				for o := t; int(o)+12 <= len(mem.Data) && o-t+12 <= firstData; o += 12 {
					w, h := u16(int(o)), u16(int(o)+2)
					data := u32(int(o) + 8)
					if data < firstData {
						firstData = data
					}
					if o-t >= firstData || w == 0 || h == 0 || w > 1024 || h > 1024 {
						break
					}
					fmtCount[mem.Data[o+4]]++
				}
			}
			// sampler palettes
			if s := offs[1]; s != 0 {
				for o := s; o+20 <= end(1); o += 20 {
					if int(int16(binary.BigEndian.Uint16(mem.Data[o+2:]))) >= 0 {
						palSamplers++
					}
				}
			}
			// mesh attr masks: the runtime's word is the u32 at +4 (1<<GXAttr bits)
			if ms := offs[11]; ms != 0 {
				tableEnd := end(11) - ms
				for o := ms; o-ms+24 <= tableEnd; o += 24 {
					dlRel := u32(int(o) + 12)
					if dlRel < tableEnd {
						tableEnd = dlRel
					}
					if o-ms >= tableEnd {
						break
					}
					mask := u32(int(o) + 4)
					if mask > 0xFFFF {
						fmt.Printf("  BIG mask %#x in %s\n", mask, tag)
						mask = 0xFFFF
					}
					maskCount[uint16(mask)]++
					if mem.Data[o+11] != 0 {
						fmt.Printf("  NBT byte %d in %s\n", mem.Data[o+11], tag)
					}
				}
			}
			// material stage lists: count stage entries != -1 beyond stage 0
			if mt := offs[10]; mt != 0 {
				for o := mt; o+40 <= end(10); o += 40 {
					n := 0
					for st := 0; st < 8; st++ {
						if int16(binary.BigEndian.Uint16(mem.Data[o+8+uint32(st)*2:])) != -1 {
							n++
						}
					}
					if n > 1 {
						multiStage[tag]++
					}
				}
			}
		}
	}

	fmt.Printf("%d .bin files\n\nsection usage:\n", files)
	for i := 0; i < 21; i++ {
		if len(secFiles[i]) > 0 {
			ex := ""
			for t := range secFiles[i] {
				ex = t
				break
			}
			fmt.Printf("  [%2d] %4d files   e.g. %s\n", i, len(secFiles[i]), ex)
		}
	}
	fmt.Printf("\nGX texture formats: ")
	var fk []int
	for k := range fmtCount {
		fk = append(fk, int(k))
	}
	sort.Ints(fk)
	for _, k := range fk {
		fmt.Printf("0x%X:%d  ", k, fmtCount[uint8(k)])
	}
	fmt.Printf("\nsamplers with palette != -1: %d\n\nattr masks: ", palSamplers)
	var mk []int
	for k := range maskCount {
		mk = append(mk, int(k))
	}
	sort.Ints(mk)
	for _, k := range mk {
		fmt.Printf("0x%03X:%d  ", k, maskCount[uint16(k)])
	}
	fmt.Println()
	if len(multiStage) > 0 {
		fmt.Printf("\nmaterials with >1 stage (%d files):\n", len(multiStage))
		var keys []string
		for k := range multiStage {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for i, k := range keys {
			if i >= 10 {
				fmt.Printf("  … %d more\n", len(keys)-10)
				break
			}
			fmt.Printf("  %s: %d materials\n", k, multiStage[k])
		}
	}
}
