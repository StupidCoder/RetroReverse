package knowledge

import "testing"

func TestPreparedStarts(t *testing.T) {
	if len(fort(t).PreparedStarts) != 4 {
		t.Fatal("missing prepared lessons")
	}
	p := fort(t)
	for _, id := range []string{"terrain", "enemy-ai"} {
		old, owned := p.PreparedStarts[id], p.PreparedStarts[id+"-owned"]
		if old.Core == owned.Core || old.SHA256 == owned.SHA256 || old.Target != owned.Target || owned.PC != old.PC {
			t.Fatal("owned lessons must bind distinct core/state identities and the same reviewed target", id)
		}
	}
	for name, change := range map[string]func(map[string]any){
		"missing target": func(m map[string]any) { at(m, "preparedStarts", "terrain", "target")["id"] = "missing" },
		"wrong hash":     func(m map[string]any) { at(m, "preparedStarts", "terrain")["sha256"] = "abc" },
		"arbitrary script": func(m map[string]any) {
			at(m, "preparedStarts", "terrain")["actions"] = []any{map[string]any{"script": "run()"}}
		},
		"held input": func(m map[string]any) {
			at(m, "preparedStarts", "terrain")["actions"] = []any{map[string]any{"key": []any{65, 1}}}
		},
		"cycle cap": func(m map[string]any) {
			at(m, "preparedStarts", "terrain")["actions"] = []any{map[string]any{"run": 200000000}, map[string]any{"run": 200000000}, map[string]any{"run": 200000000}}
		},
	} {
		t.Run(name, func(t *testing.T) {
			if _, e := Parse(changed(t, change)); e == nil {
				t.Fatal("accepted invalid preparation")
			}
		})
	}
}
