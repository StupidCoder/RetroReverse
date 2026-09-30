package knowledge

import (
	"bytes"
	"encoding/json"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf16"
)

func assetFixture(t *testing.T, media []byte, locations map[string]Location, assets map[string]Asset) *Package {
	t.Helper()
	p := Package{SchemaVersion: 1, ID: "fixture", Revision: 1, Game: Game{Name: "Synthetic fixture"}, Releases: map[string]Release{"test": {Label: "Generated media", Media: []Media{{Role: "rom", Size: uint64(len(media)), SHA256: SHA256(media)}}}}, Sources: map[string]Source{"test": {Path: "tools/knowledge/assets_test.go"}}, Evidence: map[string]Evidence{"test": {Status: "confirmed", Source: "test", Description: "Synthetic bytes generated in test"}}, Locations: locations, Assets: assets}
	p.Game.System.ID = "3ds"
	p.Game.System.Name = "Fixture"
	b, e := json.Marshal(p)
	if e != nil {
		t.Fatal(e)
	}
	var obj map[string]any
	json.Unmarshal(b, &obj)
	for key, v := range obj {
		if v == nil {
			delete(obj, key)
		}
	}
	for _, v := range obj["locations"].(map[string]any) {
		m := v.(map[string]any)
		if m["kind"] == "sectors" {
			if _, ok := m["first"]; !ok {
				m["first"] = 0
			}
			if _, ok := m["payloadOffset"]; !ok {
				m["payloadOffset"] = 0
			}
		}
	}
	b, _ = json.Marshal(obj)
	q, e := Parse(b)
	if e != nil {
		t.Fatal(e)
	}
	return q
}
func sampleAsset(location string, b []byte) Asset {
	return Asset{Label: "Fixture asset", Description: "Synthetic fixture", Category: "data", Format: "bin", Decoder: "bytes", Version: 1, Sources: []AssetSource{{location, uint64(len(b))}}, Size: uint64(len(b)), SHA256: SHA256(b), Dependencies: []AssetDependency{}, Artifacts: []Artifact{}, Releases: []string{"test"}, Evidence: []string{"test"}}
}
func openFixture(t *testing.T, p *Package, b []byte) *Resolver {
	t.Helper()
	file := filepath.Join(t.TempDir(), "media")
	if e := os.WriteFile(file, b, 0600); e != nil {
		t.Fatal(e)
	}
	r, e := OpenResolver(p, "test", map[string]string{"rom": file})
	if e != nil {
		t.Fatal(e)
	}
	t.Cleanup(r.Close)
	return r
}
func TestAssetROMFragmentsAndIdentity(t *testing.T) {
	b := []byte{0, 1, 2, 3, 4, 5, 6, 7}
	want := []byte{1, 2, 6, 7}
	a := sampleAsset("a", want)
	a.Sources = []AssetSource{{"a", 2}, {"b", 2}}
	p := assetFixture(t, b, map[string]Location{"rom": {Kind: "image", Role: "rom", Offset: "0x0"}, "a": {Kind: "relative", Base: "rom", Offset: "0x1", Length: 2}, "b": {Kind: "relative", Base: "rom", Offset: "0x6", Length: 2}}, map[string]Asset{"data": a})
	r := openFixture(t, p, b)
	x, e := r.ResolveAsset("data")
	if e != nil || !bytes.Equal(x.Data, want) || len(x.Chain) != 4 || x.Chain[3].Offset != 6 {
		t.Fatal(x, e)
	}
	wrong := append([]byte{}, b...)
	wrong[0] ^= 1
	file := filepath.Join(t.TempDir(), "wrong")
	os.WriteFile(file, wrong, 0600)
	if rr, e := OpenResolver(p, "test", map[string]string{"rom": file}); e == nil {
		rr.Close()
		t.Fatal("wrong image accepted")
	}
	if _, e := OpenResolver(p, "test", nil); e == nil {
		t.Fatal("missing role accepted")
	}
	r.remaining = 1
	if _, e := r.ResolveAsset("data"); e == nil {
		t.Fatal("read budget ignored")
	}
}
func literalYaz(b []byte) []byte {
	out := make([]byte, 16)
	copy(out, "Yaz0")
	be.PutUint32(out[4:], uint32(len(b)))
	for len(b) > 0 {
		n := len(b)
		if n > 8 {
			n = 8
		}
		out = append(out, 255)
		out = append(out, b[:n]...)
		b = b[n:]
	}
	return out
}
func syntheticSARC(name string, b []byte) []byte {
	data := (56 + len(name) + 1 + 3) &^ 3
	out := make([]byte, data+len(b))
	copy(out, "SARC")
	le.PutUint16(out[4:], 20)
	le.PutUint16(out[6:], 0xfeff)
	le.PutUint32(out[8:], uint32(len(out)))
	le.PutUint32(out[12:], uint32(data))
	copy(out[20:], "SFAT")
	le.PutUint16(out[24:], 12)
	le.PutUint16(out[26:], 1)
	le.PutUint32(out[28:], 101)
	var hash uint32
	for _, x := range []byte(name) {
		hash = hash*101 + uint32(x)
	}
	le.PutUint32(out[32:], hash)
	le.PutUint32(out[36:], 1<<24)
	le.PutUint32(out[44:], uint32(len(b)))
	copy(out[48:], "SFNT")
	le.PutUint16(out[52:], 8)
	copy(out[56:], name)
	copy(out[data:], b)
	return out
}
func namedLocation(kind, parent, decoder, path string, b []byte) Location {
	return Location{Kind: kind, Parent: parent, Decoder: decoder, Version: 1, Path: path, Size: uint64(len(b)), SHA256: SHA256(b)}
}
func TestAssetYazSARCAndBudgets(t *testing.T) {
	want := []byte("a repeated model resource")
	arc := syntheticSARC("model.bin", want)
	media := literalYaz(arc)
	loc := map[string]Location{"rom": {Kind: "image", Role: "rom", Offset: "0x0"}, "decoded": namedLocation("transform", "rom", "yaz0", "", arc), "member": namedLocation("member", "decoded", "sarc", "model.bin", want)}
	p := assetFixture(t, media, loc, map[string]Asset{"model": sampleAsset("member", want)})
	r := openFixture(t, p, media)
	x, e := r.ResolveAsset("model")
	if e != nil || !bytes.Equal(x.Data, want) || x.Chain[1].Kind != "transform" || x.Chain[2].Extents[0].Length != uint64(len(want)) {
		t.Fatal(x, e)
	}
	for _, mode := range []string{"member-path", "member-hash", "decoded-size", "allocation", "entries"} {
		t.Run(mode, func(t *testing.T) {
			r := openFixture(t, p, media)
			switch mode {
			case "member-path":
				l := p.Locations["member"]
				l.Path = "missing"
				old := p.Locations["member"]
				p.Locations["member"] = l
				defer func() { p.Locations["member"] = old }()
			case "member-hash":
				l := p.Locations["member"]
				l.SHA256 = strings.Repeat("0", 64)
				old := p.Locations["member"]
				p.Locations["member"] = l
				defer func() { p.Locations["member"] = old }()
			case "decoded-size":
				l := p.Locations["decoded"]
				l.Size++
				old := p.Locations["decoded"]
				p.Locations["decoded"] = l
				defer func() { p.Locations["decoded"] = old }()
			case "allocation":
				r.allocated = 128 << 20
			case "entries":
				r.entries = 65536
			}
			if _, e := r.ResolveAsset("model"); e == nil {
				t.Fatal("invalid recipe accepted")
			}
		})
	}
	for _, b := range [][]byte{media[:15], media[:17], append(append([]byte{}, media[:16]...), 0, 0, 0), append([]byte("Yaz0\xff\xff\xff\xff"), make([]byte, 8)...)} {
		r := &Resolver{remaining: 2 << 30}
		_, e := r.transform(byteView(b), loc["decoded"])
		if e == nil {
			t.Fatal("malformed Yaz0 accepted")
		}
	}
	// Actual overlap-copy run, and the exact output limit (not a post-allocation check).
	b := make([]byte, 16)
	copy(b, "Yaz0")
	be.PutUint32(b[4:], 6)
	b = append(b, 0x80, 'A', 0x30, 0)
	r = &Resolver{remaining: 100}
	v, e := r.transform(byteView(b), Location{Decoder: "yaz0", Size: 6})
	if e != nil {
		t.Fatal(e)
	}
	out, _ := r.read(v, 0, 6)
	if string(out) != "AAAAAA" {
		t.Fatal(string(out))
	}
}
func byteView(b []byte) assetView {
	return assetView{uint64(len(b)), func(off uint64, out []byte) error { copy(out, b[off:off+uint64(len(out))]); return nil }}
}
func isoRecord(name []byte, lba, size uint32, flags byte) []byte {
	n := 33 + len(name)
	if n%2 != 0 {
		n++
	}
	b := make([]byte, n)
	b[0] = byte(n)
	le.PutUint32(b[2:], lba)
	be.PutUint32(b[6:], lba)
	le.PutUint32(b[10:], size)
	be.PutUint32(b[14:], size)
	b[25] = flags
	le.PutUint16(b[28:], 1)
	be.PutUint16(b[30:], 1)
	b[32] = byte(len(name))
	copy(b[33:], name)
	return b
}
func syntheticISO() []byte {
	b := make([]byte, 25*2048)
	p := b[16*2048:]
	p[0] = 1
	copy(p[1:], "CD001")
	p[6] = 1
	le.PutUint32(p[80:], 25)
	be.PutUint32(p[84:], 25)
	le.PutUint16(p[128:], 2048)
	be.PutUint16(p[130:], 2048)
	copy(p[156:], isoRecord([]byte{0}, 20, 2048, 2))
	d := b[20*2048:]
	x := isoRecord([]byte("MODEL.BIN;1"), 22, 3, 128)
	copy(d, x)
	copy(d[len(x):], isoRecord([]byte("MODEL.BIN;1"), 24, 3, 0))
	copy(b[22*2048:], "abc")
	copy(b[24*2048:], "def")
	return b
}
func TestRawSectorsAndNoncontiguousISO(t *testing.T) {
	cooked := syntheticISO()
	raw := bytes.Repeat([]byte{0xaa}, 25*2352)
	for i := 0; i < 25; i++ {
		copy(raw[i*2352+16:], cooked[i*2048:(i+1)*2048])
	}
	want := []byte("abcdef")
	p := assetFixture(t, raw, map[string]Location{"raw": {Kind: "image", Role: "rom", Offset: "0x0"}, "payload": {Kind: "sectors", Parent: "raw", Count: 25, Stride: 2352, PayloadOffset: 16, PayloadSize: 2048}, "file": namedLocation("file", "payload", "iso9660", "MODEL.BIN;1", want)}, map[string]Asset{"model": sampleAsset("file", want)})
	r := openFixture(t, p, raw)
	x, e := r.ResolveAsset("model")
	if e != nil || !bytes.Equal(x.Data, want) || len(x.Chain[2].Extents) != 2 || x.Chain[1].Stride != 2352 {
		t.Fatal(x, e)
	}
	// Both-endian mismatch and duplicate final records must not pick the first hit.
	for _, mode := range []string{"endian", "duplicate", "truncated", "interleave"} {
		b := append([]byte{}, cooked...)
		switch mode {
		case "endian":
			b[16*2048+84] ^= 1
		case "duplicate":
			d := b[20*2048:]
			n := int(d[0])
			n += int(d[n])
			copy(d[n:], isoRecord([]byte("MODEL.BIN;1"), 24, 3, 0))
		case "truncated":
			b[20*2048] = 255
		case "interleave":
			b[20*2048+26] = 1
		}
		r := &Resolver{remaining: 1 << 20}
		if _, _, e := r.iso(byteView(b), "MODEL.BIN;1"); e == nil {
			t.Fatal("accepted", mode)
		}
	}
}
func syntheticRomFS(file []byte) []byte {
	// Three 512-byte IVFC levels; level 3 first. No authenticity claim: test image hash pins bytes.
	n := 512
	for n < 128+len(file) {
		n += 512
	}
	b := make([]byte, 512+n+1024)
	copy(b, "IVFC")
	le.PutUint32(b[4:], 0x10000)
	le.PutUint32(b[8:], 32)
	for i := 0; i < 3; i++ {
		le.PutUint64(b[20+i*24:], 512)
		le.PutUint32(b[28+i*24:], 9)
	}
	le.PutUint64(b[68:], uint64(n))
	h := b[512:]
	le.PutUint32(h, 40)
	le.PutUint32(h[12:], 40)
	le.PutUint32(h[16:], 24)
	le.PutUint32(h[28:], 64)
	le.PutUint32(h[32:], 64)
	le.PutUint32(h[36:], 128)
	d := h[40:]
	le.PutUint32(d[4:], 0xffffffff)
	le.PutUint32(d[8:], 0xffffffff)
	le.PutUint32(d[12:], 0)
	f := h[64:]
	le.PutUint32(f[4:], 0xffffffff)
	le.PutUint64(f[16:], uint64(len(file)))
	u := utf16.Encode([]rune("Model.szs"))
	le.PutUint32(f[28:], uint32(len(u)*2))
	for i, v := range u {
		le.PutUint16(f[32+i*2:], v)
	}
	copy(h[128:], file)
	return b
}
func TestRomFSContainerChain(t *testing.T) {
	want := []byte("MODEL")
	sarc := syntheticSARC("model.bin", want)
	yaz := literalYaz(sarc)
	rom := syntheticRomFS(yaz)
	part := make([]byte, 512+len(rom))
	copy(part[256:], "NCCH")
	part[0x18f] = 4
	le.PutUint32(part[0x1b0:], 1)
	le.PutUint32(part[0x1b4:], uint32(len(rom)/512))
	copy(part[512:], rom)
	cart := make([]byte, 512+len(part))
	copy(cart[256:], "NCSD")
	le.PutUint32(cart[0x120:], 1)
	le.PutUint32(cart[0x124:], uint32(len(part)/512))
	copy(cart[512:], part)
	loc := map[string]Location{"rom": {Kind: "image", Role: "rom", Offset: "0x0"}, "partition": namedLocation("member", "rom", "ncsd", "partition0", part), "fs": namedLocation("member", "partition", "ncch", "romfs", rom), "file": namedLocation("file", "fs", "romfs", "Model.szs", yaz), "decoded": namedLocation("transform", "file", "yaz0", "", sarc), "model": namedLocation("member", "decoded", "sarc", "model.bin", want)}
	p := assetFixture(t, cart, loc, map[string]Asset{"model": sampleAsset("model", want)})
	r := openFixture(t, p, cart)
	x, e := r.ResolveAsset("model")
	if e != nil || !bytes.Equal(x.Data, want) || len(x.Chain) != 6 {
		t.Fatal(x, e)
	}
	for _, mode := range []string{"cycle", "bounds", "name", "block"} {
		b := append([]byte{}, rom...)
		switch mode {
		case "cycle":
			le.PutUint32(b[512+64+4:], 0)
		case "bounds":
			le.PutUint64(b[512+64+16:], 0xffffffffffffffff)
		case "name":
			le.PutUint32(b[512+64+28:], 0xffffffff)
		case "block":
			le.PutUint32(b[76:], 63)
		}
		r := &Resolver{remaining: 1 << 20}
		if _, _, e := r.romfs(byteView(b), "Model.szs"); e == nil {
			t.Fatal("accepted", mode)
		}
	}
}
func TestIndexedPreviewAndArtifactIdentity(t *testing.T) {
	pal := make([]byte, 768)
	pal[3] = 255
	pixels := []byte{1, 0, 0, 1}
	b := append(append([]byte{}, pal...), pixels...)
	pa := sampleAsset("palette", pal)
	pa.Category = "palette"
	a := sampleAsset("pixels", pixels)
	a.Decoder = "indexed8"
	a.Parameters = &struct {
		Width  uint64 `json:"width"`
		Height uint64 `json:"height"`
	}{2, 2}
	a.Dependencies = []AssetDependency{{"palette", "palette"}}
	p := assetFixture(t, b, map[string]Location{"palette": {Kind: "image", Role: "rom", Offset: "0x0", Length: 768}, "pixels": {Kind: "image", Role: "rom", Offset: "0x300", Length: 4}}, map[string]Asset{"palette": pa, "picture": a})
	r := openFixture(t, p, b)
	x, e := r.ResolveAsset("picture")
	if e != nil {
		t.Fatal(e)
	}
	im, e := png.Decode(bytes.NewReader(x.Data))
	if e != nil {
		t.Fatal(e)
	}
	red, green, blue, _ := im.At(0, 0).RGBA()
	if red != 65535 || green != 0 || blue != 0 {
		t.Fatal("incorrect pixel")
	}
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "public/fixture"), 0755)
	file := filepath.Join(dir, "public/fixture/picture.png")
	os.WriteFile(file, x.Data, 0600)
	a.Artifacts = []Artifact{{"public/fixture/picture.png", "png", uint64(len(x.Data)), SHA256(x.Data)}}
	p.Assets["picture"] = a
	if e = p.VerifyArtifacts(dir); e != nil {
		t.Fatal(e)
	}
	os.WriteFile(file, []byte("wrong"), 0600)
	if e = p.VerifyArtifacts(dir); e == nil {
		t.Fatal("wrong artifact accepted")
	}
}
func TestAssetValidationAndCatalog(t *testing.T) {
	p, e := Read("../../games/captain-toad-treasure-tracker-3ds/knowledge.json")
	if e != nil {
		t.Fatal(e)
	}
	for _, edit := range []func(map[string]any){
		func(m map[string]any) { at(m, "locations", "kinopio-archive")["path"] = "../escape" },
		func(m map[string]any) { at(m, "locations", "kinopio-archive")["path"] = strings.Repeat("x", 1025) },
		func(m map[string]any) { at(m, "locations", "kinopio-decoded")["size"] = 67108865 },
		func(m map[string]any) { at(m, "locations", "kinopio-decoded")["decoder"] = "shell" },
		func(m map[string]any) { at(m, "locations", "kinopio-decoded")["version"] = 2 },
		func(m map[string]any) { at(m, "locations", "application")["parent"] = "kinopio-bch" },
		func(m map[string]any) { at(m, "assets", "kinopio")["size"] = 2 },
		func(m map[string]any) {
			at(m, "assets", "kinopio")["dependencies"] = []any{map[string]any{"asset": "kinopio", "role": "data"}}
		},
		func(m map[string]any) {
			at(m, "assets", "kinopio")["dependencies"] = []any{map[string]any{"asset": "missing", "role": "data"}}
		},
		func(m map[string]any) { at(m, "assets", "kinopio")["releases"] = []any{"wrong"} },
		func(m map[string]any) {
			at(m, "assets", "kinopio")["artifacts"] = []any{map[string]any{"path": "public/../evil.glb", "format": "glb", "size": 1, "sha256": strings.Repeat("0", 64)}}
		},
	} {
		var m map[string]any
		json.Unmarshal(p.Raw, &m)
		edit(m)
		b, _ := json.Marshal(m)
		if _, e := Parse(b); e == nil {
			t.Fatal("invalid asset accepted", string(b))
		}
	}
	a := p.Assets["kinopio"]
	a.Label = "<script>alert(1)</script>"
	p.Assets["kinopio"] = a
	b, e := ExportCatalog([]*Package{p})
	if e != nil || bytes.Contains(b, []byte("<script>")) || !bytes.Contains(b, []byte("&lt;script&gt;")) {
		t.Fatal("catalog escape", e)
	}
}

// Fuzz the bounded format entry points without private copyrighted fixtures.
func FuzzAssetFormats(f *testing.F) {
	f.Add(syntheticSARC("a", []byte{1}))
	f.Add(literalYaz([]byte{1, 2, 3}))
	f.Add(syntheticRomFS([]byte{1}))
	f.Add(syntheticISO())
	f.Fuzz(func(t *testing.T, b []byte) {
		if len(b) > 65536 {
			return
		}
		r := &Resolver{remaining: 1 << 20}
		v := byteView(b)
		r.sarc(v, "a")
		r.romfs(v, "a")
		r.iso(v, "a")
		r.transform(v, Location{Decoder: "yaz0", Size: 32})
	})
}

func TestResolutionStatusesAndArtifactEscape(t *testing.T) {
	b := []byte{1, 2, 3}
	p := assetFixture(t, b, map[string]Location{"rom": {Kind: "image", Role: "rom", Offset: "0x0"}}, map[string]Asset{"data": sampleAsset("rom", b)})
	r := openFixture(t, p, b)
	x, e := r.ResolveAsset("missing")
	if e == nil || x.Status != "inactive" {
		t.Fatal(x, e)
	}
	a := p.Assets["data"]
	a.SHA256 = strings.Repeat("0", 64)
	p.Assets["data"] = a
	x, e = r.ResolveAsset("data")
	if e == nil || x.Status != "mismatch" {
		t.Fatal(x, e)
	}
	iso := syntheticISO()
	d := iso[20*2048:]
	n := int(d[0])
	n += int(d[n])
	copy(d[n:], isoRecord([]byte("MODEL.BIN;1"), 24, 3, 0))
	r = &Resolver{remaining: 1 << 20}
	_, _, e = r.iso(byteView(iso), "MODEL.BIN;1")
	if ResolutionStatus(e) != "ambiguous" {
		t.Fatal(e)
	}
	root := t.TempDir()
	outside := filepath.Join(t.TempDir(), "a.png")
	os.WriteFile(outside, b, 0600)
	os.MkdirAll(filepath.Join(root, "public/fixture"), 0755)
	if e = os.Symlink(outside, filepath.Join(root, "public/fixture/a.png")); e != nil {
		t.Fatal(e)
	}
	a.Artifacts = []Artifact{{"public/fixture/a.png", "png", 3, SHA256(b)}}
	p.Assets["data"] = a
	if e = p.VerifyArtifacts(root); e == nil || !strings.Contains(e.Error(), "escapes") {
		t.Fatal("symlink escape", e)
	}
}
