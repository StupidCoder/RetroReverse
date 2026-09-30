package knowledge

import "testing"

func TestExperiments(t *testing.T) {
	if len(fort(t).Experiments) != 1 {
		t.Fatal("missing Fort experiment")
	}
	for name, change := range map[string]func(map[string]any){
		"script": func(m map[string]any) { at(m, "experiments", "upward-probe")["script"] = "alert(1)" },
		"register edit": func(m map[string]any) {
			at(m, "experiments", "upward-probe")["patch"].([]any)[0].(map[string]any)["kind"] = "register"
		},
		"wrong release": func(m map[string]any) { at(m, "experiments", "upward-probe")["releases"] = []any{"missing"} },
		"range": func(m map[string]any) {
			at(m, "experiments", "upward-probe")["patch"].([]any)[0].(map[string]any)["address"] = 65536
		},
		"ports": func(m map[string]any) {
			at(m, "experiments", "upward-probe")["patch"].([]any)[0].(map[string]any)["address"] = 1
		},
		"length": func(m map[string]any) {
			at(m, "experiments", "upward-probe")["patch"].([]any)[0].(map[string]any)["after"] = []any{1, 2}
		},
		"overlap": func(m map[string]any) {
			e := at(m, "experiments", "upward-probe")
			p := e["patch"].([]any)
			e["patch"] = append(p, p[0])
		},
		"schedule": func(m map[string]any) {
			at(m, "experiments", "upward-probe")["inputs"].([]any)[0].(map[string]any)["cycle"] = 1
		},
		"observation depth": func(m map[string]any) {
			c := map[string]any{"pc": 1}
			for i := 0; i < 10; i++ {
				c = map[string]any{"all": []any{c}}
			}
			at(m, "experiments", "upward-probe")["observations"].([]any)[0].(map[string]any)["when"] = c
		},
		"duplicate watch": func(m map[string]any) {
			w := at(m, "experiments", "upward-probe")["watches"].([]any)
			w[1].(map[string]any)["id"] = w[0].(map[string]any)["id"]
		},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := Parse(changed(t, change)); err == nil {
				t.Fatal("accepted invalid experiment")
			}
		})
	}
}
