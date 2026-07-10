package uvtx

import (
	"os"
	"testing"

	"retroreverse.com/games/pilotwings-64-n64/extract/pwad"
	"retroreverse.com/tools/platform/n64"
)

const romPath = "../../image/Pilotwings 64 (USA).z64"

// all decodes every UVTX in the archive, alongside whether it shipped raw.
func all(t *testing.T) (texs []*Texture, raw int) {
	t.Helper()
	if _, err := os.Stat(romPath); err != nil {
		t.Skip("cartridge image not present")
	}
	rom, err := n64.Load(romPath)
	if err != nil {
		t.Fatal(err)
	}
	a, err := pwad.Open(rom.Data)
	if err != nil {
		t.Fatal(err)
	}
	for _, i := range a.ByType("UVTX") {
		f, err := a.Resource(i)
		if err != nil {
			t.Fatal(err)
		}
		for _, c := range f.Chunks {
			tag := c.Tag
			if c.Compressed() {
				tag = c.InnerTag
			}
			if tag != "COMM" {
				continue
			}
			if !c.Compressed() {
				raw++
			}
			data, err := a.Data(c)
			if err != nil {
				t.Fatal(err)
			}
			tex, err := Decode(data)
			if err != nil {
				t.Fatalf("UVTX %d: %v", i, err)
			}
			texs = append(texs, tex)
		}
	}
	return texs, raw
}

// Every texture in the archive must decode. There is no fallback path: a
// texture that cannot be read is an error, not a white square. (Every GLB
// exported before 2026-07-10 shipped with white squares and looked plausible.)
func TestEveryTextureDecodes(t *testing.T) {
	texs, raw := all(t)
	if got, want := len(texs), 463; got != want {
		t.Fatalf("decoded %d textures, want %d", got, want)
	}
	if got, want := raw, 23; got != want {
		t.Errorf("%d textures ship uncompressed, want %d", got, want)
	}
	for i, tex := range texs {
		if tex.Image == nil || tex.Width == 0 || tex.Height == 0 {
			t.Fatalf("texture %d decoded to nothing", i)
		}
		if tex.DataSize > maxDataSize {
			t.Errorf("texture %d: %d texel bytes exceeds the game's own clamp", i, tex.DataSize)
		}
	}
}

// The format census. A misread of the material template would redistribute
// textures between formats, so pinning the counts catches it.
func TestFormatCensus(t *testing.T) {
	texs, _ := all(t)
	want := map[string]int{"RGBA16": 255, "I4": 103, "IA8": 56, "I8": 20, "IA4": 20, "IA16": 9}
	got := map[string]int{}
	for _, tex := range texs {
		got[tex.Format()]++
	}
	for f, n := range want {
		if got[f] != n {
			t.Errorf("%s: %d textures, want %d", f, got[f], n)
		}
	}
	for f, n := range got {
		if _, ok := want[f]; !ok {
			t.Errorf("unexpected format %s (%d textures)", f, n)
		}
	}
}

// No shipped texture is paletted, so no UVTX needs a TLUT — a fact the decoder
// leans on. If one ever did, its palette would have to come from somewhere.
func TestNoPalettedTextures(t *testing.T) {
	texs, _ := all(t)
	for i, tex := range texs {
		if tex.Fmt == 2 {
			t.Fatalf("texture %d is CI%d — the archive was assumed palette-free", i, 4<<tex.Siz)
		}
	}
}
