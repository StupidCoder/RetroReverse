package main

// models.go exports one textured GLB per selectable car. A car is the
// composite the car-select carousel draws (0x8001D2F4): the body object, a
// canopy/cockpit object, an underbody object, and one axle object drawn at
// the rear (the car origin) and again offset by the car table's wheelbase
// halfword. The 16-byte-per-car table at 0x80056B40 supplies the object
// indices; the name list at the head of the text segment supplies the names,
// in the same order.

import (
	"fmt"
	"os"
	"path/filepath"

	"retroreverse.com/games/ridge-racer-psx/extract/rr"
)

const (
	carTableOff = 0x80056B40 - 0x80010000 // car table in the text segment
	carCount    = 13                      // 12 selectable + the hidden 13th
)

// car is one decoded car-table entry.
type car struct {
	Name      string
	Body      int   // h1: the display body object
	LOD       int   // h2: the reduced in-race body
	Canopy    int   // h3: canopy family base (axle +1, shadow +2, underbody +3)
	Wheelbase int16 // h5: Z offset of the second axle
}

// carTable decodes the executable's car table and name list.
func carTable(text []byte) ([]car, error) {
	if len(text) < carTableOff+carCount*16 {
		return nil, fmt.Errorf("text segment too short for the car table")
	}
	names := carNames(text)
	cars := make([]car, carCount)
	for i := range cars {
		e := carTableOff + i*16
		cars[i] = car{
			Name:      names[i],
			Body:      int(int16(u16(text, e+2))),
			LOD:       int(int16(u16(text, e+4))),
			Canopy:    int(int16(u16(text, e+6))),
			Wheelbase: int16(u16(text, e+10)),
		}
	}
	return cars, nil
}

// carNames reads the 13 NUL-terminated car names that open the text segment.
func carNames(text []byte) [carCount]string {
	var names [carCount]string
	off := 0
	for i := 0; i < carCount; i++ {
		for off < len(text) && text[off] == 0 {
			off++
		}
		start := off
		for off < len(text) && text[off] != 0 {
			off++
		}
		names[i] = string(text[start:off])
	}
	return names
}

func u16(b []byte, off int) uint16 { return uint16(b[off]) | uint16(b[off+1])<<8 }

// addObject feeds every record of one object into the mesh at a world offset.
func addObject(b *meshBuilder, o *rr.Object, off [3]int32) {
	for _, q := range o.FT {
		b.AddTextured(q.V, q.UV, q.TPage, q.CLUT, off, 0)
	}
	for _, q := range o.FT8 {
		b.AddTextured(q.V, q.UV, q.TPage, q.CLUT, off, 0)
	}
	for _, q := range o.F {
		b.AddFlat(q.V, q.RGB, off)
	}
	for _, q := range o.GT {
		b.AddTextured(q.V, q.UV, q.TPage, q.CLUT, off, 0)
	}
	for _, q := range o.GT8 {
		b.AddTextured(q.V, q.UV, q.TPage, q.CLUT, off, 0)
	}
	for _, q := range o.G {
		b.AddFlat(q.V, q.RGB, off)
	}
}

func exportCars(a *assets, out string) ([]ModelIndex, error) {
	cars, err := carTable(a.exe)
	if err != nil {
		return nil, err
	}
	dir := filepath.Join(out, "models")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	var models []ModelIndex
	for i, c := range cars {
		b := newMeshBuilder(a.vrams[0])
		addObject(b, &a.objs[c.Body], [3]int32{0, 0, 0})
		addObject(b, &a.objs[c.Canopy], [3]int32{0, 0, 0})
		addObject(b, &a.objs[c.Canopy+3], [3]int32{0, 0, 0}) // underbody
		axle := &a.objs[c.Canopy+1]
		addObject(b, axle, [3]int32{0, 0, 0})                       // rear axle
		addObject(b, axle, [3]int32{0, 0, int32(c.Wheelbase)})      // front axle
		file := fmt.Sprintf("models/car-%02d.glb", i+1)
		if err := b.Write(filepath.Join(out, file)); err != nil {
			return nil, err
		}
		fmt.Fprintf(os.Stderr, "[models] %2d/%d %s (%s, %d verts)\n",
			i+1, len(cars), filepath.Base(file), c.Name, len(b.verts))
		models = append(models, ModelIndex{Name: c.Name, File: file, Kind: "mesh3d"})
	}
	return models, nil
}
