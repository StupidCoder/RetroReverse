package knowledge

import "testing"

func TestM4Tours(t *testing.T) {
	if len(fort(t).Tours) != 1 {
		t.Fatal("missing Fort tour")
	}
	for name, change := range map[string]func(map[string]any){
		"unknown release":   func(m map[string]any) { at(m, "tours", "terrain-runs")["releases"] = []any{"missing"} },
		"code injection":    func(m map[string]any) { at(m, "tours", "terrain-runs")["script"] = "alert(1)" },
		"unsupported start": func(m map[string]any) { at(m, "tours", "terrain-runs")["startPolicy"] = "patch-ram" },
		"deep predicate": func(m map[string]any) {
			c := map[string]any{"pc": 1}
			for i := 0; i < 10; i++ {
				c = map[string]any{"all": []any{c}}
			}
			at(m, "tours", "terrain-runs")["start"] = c
		},
		"mask mismatch": func(m map[string]any) {
			at(m, "tours", "terrain-runs")["start"] = map[string]any{"address": 5, "value": 128, "mask": 127}
		},
		"guard overrun": func(m map[string]any) {
			at(m, "tours", "terrain-runs")["guards"] = []any{map[string]any{"address": 65535, "bytes": []any{1, 2}}}
		},
		"capture overlap": func(m map[string]any) {
			s := at(m, "tours", "terrain-runs")["stops"].([]any)[0].(map[string]any)
			s["capture"] = []any{map[string]any{"address": 100, "length": 10}, map[string]any{"address": 105, "length": 10}}
		},
		"cycle budget": func(m map[string]any) {
			at(m, "tours", "terrain-runs")["stops"].([]any)[0].(map[string]any)["cycleBudget"] = 0
		},
		"duplicate stop": func(m map[string]any) {
			s := at(m, "tours", "terrain-runs")["stops"].([]any)
			s[1].(map[string]any)["id"] = s[0].(map[string]any)["id"]
		},
	} {
		t.Run(name, func(t *testing.T) {
			if _, e := Parse(changed(t, change)); e == nil {
				t.Fatal("accepted invalid tour")
			}
		})
	}
}
