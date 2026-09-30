package knowledge

import (
	"encoding/json"
	"os"
	"testing"
)

func Test3DOKnowledge(t *testing.T) {
	b, err := os.ReadFile("../../games/need-for-speed-3do/knowledge.json")
	if err != nil {
		t.Fatal(err)
	}
	p, err := Parse(b)
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Tours["one-iteration"].Stops) != 17 || len(p.LiveGuards) < 14 {
		t.Fatal("missing verified loop or layout guards")
	}
	for name, change := range map[string]func(map[string]any){
		"guard outside RAM":   func(m map[string]any) { m["liveGuards"].([]any)[0].(map[string]any)["address"] = 0x300000 },
		"counter outside RAM": func(m map[string]any) { at(m, "tours", "one-iteration", "iteration")["counter"] = 0x2fffff },
		"x86 register": func(m map[string]any) {
			at(m, "tours", "one-iteration")["start"] = map[string]any{"register": "EAX", "value": 0}
		},
		"x86 mode": func(m map[string]any) { at(m, "tours", "one-iteration")["start"] = map[string]any{"mode": "real16"} },
	} {
		t.Run(name, func(t *testing.T) {
			var m map[string]any
			if err := json.Unmarshal(b, &m); err != nil {
				t.Fatal(err)
			}
			change(m)
			v, _ := json.Marshal(m)
			if _, err := Parse(v); err == nil {
				t.Fatal("accepted invalid ARM knowledge")
			}
		})
	}
	for name, change := range map[string]func(map[string]any){
		"ARM register": func(m map[string]any) {
			at(m, "tours", "terrain-runs")["start"] = map[string]any{"register": "R4", "value": 0}
		},
		"iteration":   func(m map[string]any) { at(m, "tours", "terrain-runs")["iteration"] = map[string]any{"counter": 0} },
		"live guards": func(m map[string]any) { m["liveGuards"] = []any{map[string]any{"address": 0, "bytes": []any{0}}} },
	} {
		t.Run("C64 rejects "+name, func(t *testing.T) {
			if _, err := Parse(changed(t, change)); err == nil {
				t.Fatal("accepted ARM-only capability on C64")
			}
		})
	}
}
