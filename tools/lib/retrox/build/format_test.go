package build

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestMarshalPrettyInlinesScalarArrays(t *testing.T) {
	v := map[string]any{
		"cells": []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
	}
	out, err := marshalPretty(v)
	if err != nil {
		t.Fatal(err)
	}
	s := string(out)
	if !strings.Contains(s, `"cells": [1, 2, 3, 4, 5, 6, 7, 8, 9, 10]`) {
		t.Fatalf("scalar array not inlined:\n%s", s)
	}
}

func TestMarshalPrettyScalarArrayAlwaysOneLine(t *testing.T) {
	big := make([]int, 5000)
	for i := range big {
		big[i] = i % 97
	}
	out, err := marshalPretty(map[string]any{"cells": big})
	if err != nil {
		t.Fatal(err)
	}
	// One line for the key+array, plus the braces: 3 lines total.
	if got := strings.Count(string(out), "\n"); got != 3 {
		t.Fatalf("big scalar array split over lines (%d newlines):\n%.200s...", got, out)
	}
}

func TestMarshalPrettyNestedTableRowPerLine(t *testing.T) {
	rows := make([][]int, 4)
	for i := range rows {
		rows[i] = make([]int, 32)
		for j := range rows[i] {
			rows[i][j] = 100000 + i*100 + j // long enough to exceed the inline limit
		}
	}
	out, err := marshalPretty(map[string]any{"tiles": rows})
	if err != nil {
		t.Fatal(err)
	}
	s := string(out)
	// {, "tiles": [, one line per row, ], } + trailing newline.
	if strings.Count(s, "\n") != 4+len(rows) {
		t.Fatalf("nested table not row-per-line:\n%s", s)
	}
	if !strings.Contains(s, "[100000, 100001") {
		t.Fatalf("row not inlined:\n%s", s)
	}
}

func TestMarshalPrettySmallNestedInline(t *testing.T) {
	out, err := marshalPretty(map[string]any{"steps": [][]int{{0, 300}, {1, 16}}})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(out), `"steps": [[0, 300], [1, 16]]`) {
		t.Fatalf("small nested array not inlined:\n%s", out)
	}
}

func TestMarshalPrettyRoundTrips(t *testing.T) {
	src := map[string]any{
		"format": "retro-x", "version": 1,
		"nested": map[string]any{"a": []any{1.5, "x", true, nil}},
		"objs":   []any{map[string]any{"id": 1}, map[string]any{"id": 2}},
		"empty":  []any{}, "emptyObj": map[string]any{},
	}
	out, err := marshalPretty(src)
	if err != nil {
		t.Fatal(err)
	}
	var back any
	if err := json.Unmarshal(out, &back); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, out)
	}
	want, _ := json.Marshal(src)
	got, _ := json.Marshal(back)
	// Compare canonical forms (map ordering normalised by Marshal).
	var w1, g1 any
	_ = json.Unmarshal(want, &w1)
	_ = json.Unmarshal(got, &g1)
	if string(mustJSON(w1)) != string(mustJSON(g1)) {
		t.Fatalf("round-trip mismatch:\n%s\nvs\n%s", want, got)
	}
}

func mustJSON(v any) []byte {
	b, _ := json.Marshal(v)
	return b
}

func TestMarshalPrettyPreservesKeyOrder(t *testing.T) {
	type doc struct {
		Format  string `json:"format"`
		Version int    `json:"version"`
		Zed     int    `json:"zed"`
		Alpha   int    `json:"alpha"`
	}
	out, err := marshalPretty(doc{Format: "retro-x", Version: 1, Zed: 1, Alpha: 2})
	if err != nil {
		t.Fatal(err)
	}
	s := string(out)
	if strings.Index(s, `"format"`) > strings.Index(s, `"zed"`) ||
		strings.Index(s, `"zed"`) > strings.Index(s, `"alpha"`) {
		t.Fatalf("key order not preserved:\n%s", s)
	}
}
