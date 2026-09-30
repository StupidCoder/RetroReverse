package knowledge

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func fortBytes(t *testing.T) []byte {
	t.Helper()
	b, e := os.ReadFile("../../games/fort-apocalypse-c64/knowledge.json")
	if e != nil {
		t.Fatal(e)
	}
	return b
}
func fort(t *testing.T) *Package {
	t.Helper()
	p, e := Parse(fortBytes(t))
	if e != nil {
		t.Fatal(e)
	}
	return p
}
func at(m map[string]any, ks ...string) map[string]any {
	for _, k := range ks {
		m = m[k].(map[string]any)
	}
	return m
}
func changed(t *testing.T, f func(map[string]any)) []byte {
	t.Helper()
	var m map[string]any
	if e := json.Unmarshal(fortBytes(t), &m); e != nil {
		t.Fatal(e)
	}
	f(m)
	b, e := json.Marshal(m)
	if e != nil {
		t.Fatal(e)
	}
	return b
}
func TestFortAndDeterministicExport(t *testing.T) {
	p := fort(t)
	uw, e := Read("../../games/ultima-underworld-pc/knowledge.json")
	if e != nil {
		t.Fatal(e)
	}
	a, e := ExportJS([]*Package{p, uw})
	if e != nil {
		t.Fatal(e)
	}
	b, e := ExportJS([]*Package{uw, p})
	if e != nil || !bytes.Equal(a, b) {
		t.Fatal("nondeterministic export")
	}
	actual, e := os.ReadFile("../../site/emulators/knowledge-data.js")
	if e != nil || !bytes.Equal(a, actual) {
		t.Fatal("stale generated knowledge-data.js; run knowledgeexport")
	}
	d, e := p.BrowserData()
	if e != nil {
		t.Fatal(e)
	}
	if len(d.Releases) != 1 || len(d.Releases[0].Labels) != 13 {
		t.Fatal("missing migrated Fort labels")
	}
	for _, l := range d.Releases[0].Labels {
		if l.Region != "ram" && l.Region != "tape" {
			t.Fatal("wrong physical region")
		}
		if !strings.Contains(l.Description, "Applicability:") {
			t.Fatal("phase caveat lost")
		}
	}
	if _, e = ExportJS([]*Package{p, p}); e == nil {
		t.Fatal("duplicate package accepted")
	}
}
func TestRejectInvalidPackages(t *testing.T) {
	tests := map[string]func(map[string]any){
		"unknown field":        func(m map[string]any) { m["typo"] = true },
		"unknown nested field": func(m map[string]any) { at(m, "game")["execute"] = "bad" },
		"wrong version":        func(m map[string]any) { m["schemaVersion"] = 2 },
		"hash": func(m map[string]any) {
			at(m, "releases", "reference-pal")["media"].([]any)[0].(map[string]any)["sha256"] = "ABC"
		},
		"duplicate roles": func(m map[string]any) {
			r := at(m, "releases", "reference-pal")
			a := r["media"].([]any)
			r["media"] = append(a, a[0])
		},
		"ambiguous releases":   func(m map[string]any) { r := at(m, "releases"); r["other"] = r["reference-pal"] },
		"future operation":     func(m map[string]any) { m["tours"] = map[string]any{"bad": map[string]any{"execute": "anything"}} },
		"unknown kind":         func(m map[string]any) { at(m, "locations", "terrain")["kind"] = "javascript" },
		"noncanonical address": func(m map[string]any) { at(m, "locations", "terrain")["offset"] = "0x0503" },
		"address overflow":     func(m map[string]any) { at(m, "locations", "terrain")["offset"] = "0x10000000000000000" },
		"out of bounds":        func(m map[string]any) { at(m, "regions", "terrain")["length"] = 65536 },
		"location cycle": func(m map[string]any) {
			at(m, "locations")["terrain"] = map[string]any{"kind": "relative", "base": "terrain", "offset": "0x0"}
		},
		"one past end relative": func(m map[string]any) {
			at(m, "locations")["past"] = map[string]any{"kind": "relative", "base": "tape-payload-source", "offset": "0x37205"}
		},
		"bad space":              func(m map[string]any) { at(m, "locations", "terrain")["space"] = "cpu" },
		"unknown release":        func(m map[string]any) { at(m, "regions", "terrain")["releases"] = []string{"missing"} },
		"unknown evidence":       func(m map[string]any) { at(m, "regions", "terrain")["evidence"] = []string{"missing"} },
		"unknown source":         func(m map[string]any) { at(m, "evidence", "architecture")["source"] = "missing" },
		"script source":          func(m map[string]any) { at(m, "sources", "architecture")["path"] = "javascript:alert(1)" },
		"encoded traversal":      func(m map[string]any) { at(m, "sources", "architecture")["path"] = "/%2e%2e/private" },
		"empty evidence":         func(m map[string]any) { at(m, "regions", "terrain")["evidence"] = []string{} },
		"unknown location":       func(m map[string]any) { at(m, "regions", "terrain")["location"] = "missing" },
		"unknown type":           func(m map[string]any) { at(m, "state", "enemy-mode")["type"] = "missing" },
		"enum width":             func(m map[string]any) { at(m, "types", "enemy-mode", "values")["256"] = "too big" },
		"negative unsigned enum": func(m map[string]any) { at(m, "types", "enemy-mode", "values")["-1"] = "negative" },
		"unknown function":       func(m map[string]any) { at(m, "annotations", "upward-probe-caution")["function"] = "missing" },
		"label logical alias":    func(m map[string]any) { at(m, "regions", "terrain")["location"] = "enemy-mode" },
		"missing applicability":  func(m map[string]any) { delete(at(m, "regions", "terrain"), "applicability") },
	}
	for name, f := range tests {
		t.Run(name, func(t *testing.T) {
			if _, e := Parse(changed(t, f)); e == nil {
				t.Fatal("accepted invalid package")
			}
		})
	}
}
func TestStrictJSON(t *testing.T) {
	for _, b := range [][]byte{[]byte(`{"schemaVersion":1,"schemaVersion":1}`), append(fortBytes(t), []byte(` {}`)...), bytes.Repeat([]byte(" "), MaxPackageBytes+1), []byte{0xff}} {
		if _, e := Parse(b); e == nil {
			t.Fatal("accepted invalid JSON")
		}
	}
	raw := changed(t, func(m map[string]any) { m["extensions"] = map[string]any{"test:deep": nil} })
	raw = bytes.Replace(raw, []byte(`"test:deep":null`), []byte(`"test:deep":`+strings.Repeat("[", 70)+"0"+strings.Repeat("]", 70)), 1)
	if _, e := Parse(raw); e == nil {
		t.Fatal("accepted excessive depth")
	}
	if e := auditSchema(map[string]any{"not": map[string]any{}}, nil); e == nil {
		t.Fatal("unimplemented schema keyword ignored")
	}
}
func TestRelativeAndTypes(t *testing.T) {
	data := changed(t, func(m map[string]any) {
		at(m, "locations")["second-byte"] = map[string]any{"kind": "relative", "base": "terrain", "offset": "0x1"}
		ts := at(m, "types")
		ts["u16"] = map[string]any{"kind": "integer", "bytes": 2, "signed": false, "endian": "little"}
		ts["record"] = map[string]any{"kind": "struct", "size": 4, "fields": []any{map[string]any{"name": "value", "offset": 0, "type": "u16"}, map[string]any{"name": "flags", "offset": 2, "type": "u8"}}}
		ts["records"] = map[string]any{"kind": "array", "element": "record", "count": 2, "stride": 4}
	})
	p, e := Parse(data)
	if e != nil {
		t.Fatal(e)
	}
	s, e := p.Resolve("second-byte", "reference-pal")
	if e != nil || s.Offset != 0x504 {
		t.Fatal(s, e)
	}
	n, e := p.TypeSize("records")
	if e != nil || n != 8 {
		t.Fatal(n, e)
	}
	for _, bad := range []Type{{Kind: "array", Element: "record", Count: 2, Stride: 3}, {Kind: "array", Element: "bad", Count: 2, Stride: 8}, {Kind: "struct", Size: 1, Fields: []Field{{Name: "x", Offset: 0, Type: "u16"}}}, {Kind: "struct", Size: 4, Fields: []Field{{Name: "x", Type: "u16"}, {Name: "y", Type: "u16"}}}, {Kind: "fixed", Storage: "u8", FractionBits: 8}, {Kind: "bitfield", Storage: "u8", Bits: map[string]Bits{"bad": {Offset: 7, Width: 2}}}} {
		p.Types["bad"] = bad
		if _, e := p.TypeSize("bad"); e == nil {
			t.Fatal("accepted invalid layout", bad)
		}
	}
	p = fort(t)
	v, e := p.DecodeScalar("enemy-mode", []byte{99})
	if e != nil || v.Known || v.Value != "99" || v.Label != "" {
		t.Fatal(v, e)
	}
	v, e = p.DecodeScalar("enemy-mode", []byte{3})
	if e != nil || !v.Known || v.Label != "Hunting" {
		t.Fatal(v, e)
	}
	p.Types["signed"] = Type{Kind: "integer", Bytes: 8, Signed: true, Endian: "big"}
	v, e = p.DecodeScalar("signed", []byte{0x80, 0, 0, 0, 0, 0, 0, 0})
	if e != nil || v.Value != "-9223372036854775808" {
		t.Fatal(v, e)
	}
}
func TestM0IdentityVectors(t *testing.T) {
	b, e := os.ReadFile("../browser/tests/fixtures/code-m0/contracts.json")
	if e != nil {
		t.Fatal(e)
	}
	var vectors struct {
		Identity struct {
			ValidPaths   []struct{ Input, Policy, Canonical string }
			InvalidPaths []string
			Collision    []string
			Members      []struct{ Path, Hex, SHA256 string }
		}
	}
	if e = json.Unmarshal(b, &vectors); e != nil {
		t.Fatal(e)
	}
	for _, p := range vectors.Identity.ValidPaths {
		s, e := CanonicalPath(p.Input, p.Policy)
		if e != nil || s != p.Canonical {
			t.Fatal(s, e)
		}
	}
	for _, p := range vectors.Identity.InvalidPaths {
		if _, e := CanonicalPath(p, "exact"); e == nil {
			t.Fatal("accepted path", p)
		}
	}
	members := []Member{{Path: "B.BIN", Size: 0, SHA256: SHA256(nil)}, {Path: "A.BIN", Size: 3, SHA256: SHA256([]byte("abc"))}}
	golden, e := os.ReadFile("../browser/results/code-m0/contract-vectors.json")
	if e != nil {
		t.Fatal(e)
	}
	var expected struct{ Manifest, ManifestSHA256 string }
	if e = json.Unmarshal(golden, &expected); e != nil {
		t.Fatal(e)
	}
	data, h, e := Manifest(members, "ascii-insensitive")
	if e != nil || string(data) != expected.Manifest || h != expected.ManifestSHA256 {
		t.Fatal(string(data), h, e)
	}
	members[0], members[1] = members[1], members[0]
	again, _, _ := Manifest(members, "ascii-insensitive")
	if !bytes.Equal(data, again) {
		t.Fatal("order dependent")
	}
	for _, name := range []string{"DATA/model.bin", "data/MODEL.BIN"} {
		members = append(members, Member{Path: name, SHA256: SHA256(nil)})
	}
	if _, _, e = Manifest(members, "ascii-insensitive"); e == nil {
		t.Fatal("collision accepted")
	}
	unicode := []Member{{Path: "line\u2028&<.bin", SHA256: SHA256(nil)}}
	data, _, e = Manifest(unicode, "exact")
	if e != nil || bytes.Contains(data, []byte(`\u2028`)) || bytes.Contains(data, []byte(`\u0026`)) {
		t.Fatal("not canonical literal UTF8", string(data), e)
	}
}
func TestReleaseMatching(t *testing.T) {
	p := fort(t)
	path := filepath.Join(t.TempDir(), "test.bin")
	if e := os.WriteFile(path, []byte("abc"), 0600); e != nil {
		t.Fatal(e)
	}
	p.Releases = map[string]Release{"test": {Label: "Synthetic", Media: []Media{{Role: "rom", Size: 3, SHA256: SHA256([]byte("abc"))}}}}
	if r, e := p.MatchMedia(map[string]string{"rom": path}); e != nil || r != "test" {
		t.Fatal(r, e)
	}
	if _, e := p.MatchMedia(map[string]string{"wrong-role": path}); e == nil {
		t.Fatal("wrong role matched")
	}
	if e := os.WriteFile(path, []byte("abd"), 0600); e != nil {
		t.Fatal(e)
	}
	if _, e := p.MatchMedia(map[string]string{"rom": path}); e == nil {
		t.Fatal("wrong hash matched")
	}
	ms := []Member{{Path: "A.BIN", Size: 3, SHA256: SHA256([]byte("abc"))}}
	_, h, _ := Manifest(ms, "ascii-insensitive")
	p.Releases = map[string]Release{"files": {FileSet: &FileSet{Algorithm: "rr-media-set-v1", PathPolicy: "ascii-insensitive", Members: ms, SHA256: h}}}
	if r, e := p.MatchFileSet(append(ms, Member{Path: "extra", SHA256: SHA256(nil)})); e != nil || r != "files" {
		t.Fatal(r, e)
	}
	before := append([]Member(nil), ms...)
	if !reflect.DeepEqual(ms, before) {
		t.Fatal("mutated identity")
	}
	if _, e := p.MatchFileSet([]Member{{Path: "a.bin", Size: 3, SHA256: SHA256([]byte("abd"))}}); e == nil {
		t.Fatal("wrong file-set hash matched")
	}
	p.Releases["duplicate"] = p.Releases["files"]
	if _, e := p.MatchFileSet(ms); e == nil {
		t.Fatal("ambiguous file set matched")
	}
}

func TestFileSetPackageValidation(t *testing.T) {
	members := []Member{{Path: "DATA/ABC", Size: 3, SHA256: SHA256([]byte("abc"))}}
	_, hash, _ := Manifest(members, "ascii-insensitive")
	root := map[string]any{"schemaVersion": 1, "id": "synthetic-dos", "revision": 1,
		"game":     map[string]any{"name": "Synthetic", "system": map[string]any{"id": "dos", "name": "DOS"}},
		"releases": map[string]any{"files": map[string]any{"label": "Fixture", "fileSet": FileSet{Algorithm: "rr-media-set-v1", PathPolicy: "ascii-insensitive", Members: members, SHA256: hash}}}}
	data, _ := json.Marshal(root)
	p, e := Parse(data)
	if e != nil {
		t.Fatal(e)
	}
	if r, e := p.MatchFileSet(members); e != nil || r != "files" {
		t.Fatal(r, e)
	}
	files := root["releases"].(map[string]any)["files"].(map[string]any)
	s := files["fileSet"].(FileSet)
	s.SHA256 = strings.Repeat("0", 64)
	files["fileSet"] = s
	data, _ = json.Marshal(root)
	if _, e := Parse(data); e == nil {
		t.Fatal("accepted inconsistent manifest hash")
	}
}

func TestSharedTypeGraphAndSupportedLayouts(t *testing.T) {
	p := fort(t)
	p.Types["bits"] = Type{Kind: "bitfield", Storage: "u8", Bits: map[string]Bits{"low": {Offset: 0, Width: 3}, "high": {Offset: 3, Width: 5}}}
	p.Types["q4"] = Type{Kind: "fixed", Storage: "u8", FractionBits: 4}
	for _, id := range []string{"bits", "q4"} {
		if n, e := p.TypeSize(id); e != nil || n != 1 {
			t.Fatal(id, n, e)
		}
	}
	// Reusing the same nested type must not expand validation exponentially.
	prev := "u8"
	size := uint64(1)
	for i := 0; i < 12; i++ {
		id := strings.Repeat("x", i+1)
		p.Types[id] = Type{Kind: "struct", Size: size * 2, Fields: []Field{{Name: "a", Type: prev}, {Name: "b", Offset: size, Type: prev}}}
		prev = id
		size *= 2
	}
	if n, e := p.TypeSize(prev); e != nil || n != 4096 {
		t.Fatal(n, e)
	}
}

func TestM3StateRelationsAndOccupancy(t *testing.T) {
	fixture := func(m map[string]any) {
		types := at(m, "types")
		types["slot"] = map[string]any{"kind": "struct", "size": 2, "fields": []any{map[string]any{"name": "type", "offset": 0, "type": "u8"}, map[string]any{"name": "flags", "offset": 1, "type": "u8"}}}
		types["slots"] = map[string]any{"kind": "array", "element": "slot", "count": 2, "stride": 2}
		d := at(m, "state", "enemy-mode")
		d["type"] = "slots"
		d["occupancy"] = map[string]any{"path": []string{"type"}, "notEquals": "0"}
	}
	if _, e := Parse(changed(t, fixture)); e != nil {
		t.Fatal(e)
	}
	for _, mutate := range []func(map[string]any){
		func(m map[string]any) { at(m, "state", "enemy-mode")["relatedFunctions"] = []string{"missing"} },
		func(m map[string]any) { at(m, "state", "enemy-mode", "occupancy")["path"] = []string{"missing"} },
		func(m map[string]any) { at(m, "state", "enemy-mode", "occupancy")["notEquals"] = "256" },
		func(m map[string]any) { at(m, "state", "enemy-mode")["type"] = "slot" },
	} {
		if _, e := Parse(changed(t, func(m map[string]any) { fixture(m); mutate(m) })); e == nil {
			t.Fatal("invalid occupancy/relation accepted")
		}
	}
}

func TestM3FinalByteState(t *testing.T) {
	_, e := Parse(changed(t, func(m map[string]any) { at(m, "locations", "enemy-mode")["address"] = "0xffff" }))
	if e != nil {
		t.Fatal("one-byte watch at final CPU address must fit:", e)
	}
}
