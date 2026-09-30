package knowledge

import (
	"encoding/json"
	"os"
	"testing"
)

func TestM5ModuleAndBufferValidation(t *testing.T) {
	b, e := os.ReadFile("../../games/ultima-underworld-pc/knowledge.json")
	if e != nil {
		t.Fatal(e)
	}
	if _, e = Parse(b); e != nil {
		t.Fatal(e)
	}
	for name, change := range map[string]func(map[string]any){
		"bad relocation": func(m map[string]any) {
			at(m, "modules", "raster")["relocations"] = []any{map[string]any{"offset": 65534, "relativeSegment": 0}}
		},
		"unknown module":   func(m map[string]any) { at(m, "locations", "span")["module"] = "missing" },
		"buffer too large": func(m map[string]any) { at(m, "buffers", "chunky")["stride"] = 1024 },
		"buffer condition": func(m map[string]any) { at(m, "buffers", "chunky")["when"] = map[string]any{"location": "missing"} },
		"wrong entry":      func(m map[string]any) { at(m, "releases", "reference-installed")["executable"] = "other.exe" },
		"MZ paragraph":     func(m map[string]any) { delete(at(m, "modules", "raster"), "paragraph") },
		"signature bounds": func(m map[string]any) { at(m, "modules", "geometry")["size"] = 100 },
	} {
		t.Run(name, func(t *testing.T) {
			var m map[string]any
			json.Unmarshal(b, &m)
			change(m)
			v, _ := json.Marshal(m)
			if _, e := Parse(v); e == nil {
				t.Fatal("accepted invalid module profile")
			}
		})
	}
}
