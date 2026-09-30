package knowledge

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"reflect"
	"regexp"
	"sort"
	"strings"
)

// Schema is a standard JSON Schema document. The private evaluator implements only
// the keywords used by this embedded schema; it is not a general schema service.
//
//go:embed schema/game-knowledge-v1.schema.json
var Schema []byte

const MaxPackageBytes = 8 << 20

// Decode rejects duplicate keys, excess nesting and trailing values before schema validation.
func decode(data []byte) (any, error) {
	if len(data) > MaxPackageBytes {
		return nil, fmt.Errorf("package exceeds 8 MiB")
	}
	d := json.NewDecoder(bytes.NewReader(data))
	d.UseNumber()
	var read func(int) (any, error)
	read = func(depth int) (any, error) {
		if depth > 64 {
			return nil, fmt.Errorf("JSON nesting exceeds 64")
		}
		t, e := d.Token()
		if e != nil {
			return nil, e
		}
		switch t {
		case json.Delim('{'):
			m := map[string]any{}
			for d.More() {
				k, e := d.Token()
				if e != nil {
					return nil, e
				}
				key, ok := k.(string)
				if !ok {
					return nil, fmt.Errorf("invalid key")
				}
				if _, exists := m[key]; exists {
					return nil, fmt.Errorf("duplicate key %q", key)
				}
				v, e := read(depth + 1)
				if e != nil {
					return nil, e
				}
				m[key] = v
			}
			_, e = d.Token()
			return m, e
		case json.Delim('['):
			a := []any{}
			for d.More() {
				v, e := read(depth + 1)
				if e != nil {
					return nil, e
				}
				a = append(a, v)
			}
			_, e = d.Token()
			return a, e
		default:
			return t, nil
		}
	}
	v, e := read(0)
	if e != nil {
		return nil, e
	}
	if _, e = d.Token(); e != io.EOF {
		return nil, fmt.Errorf("trailing JSON data")
	}
	return v, nil
}
func number(v any) float64 {
	switch n := v.(type) {
	case json.Number:
		f, _ := n.Float64()
		return f
	case float64:
		return n
	}
	return math.NaN()
}
func keys[V any](m map[string]V) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
func schemaCheck(s map[string]any, v any, path string, root map[string]any) error {
	fail := func(s string) error { return fmt.Errorf("%s: %s", path, s) }
	if ref, ok := s["$ref"].(string); ok {
		return schemaCheck(root["$defs"].(map[string]any)[strings.TrimPrefix(ref, "#/$defs/")].(map[string]any), v, path, root)
	}
	if c, ok := s["const"]; ok && !reflect.DeepEqual(c, v) {
		return fail(fmt.Sprintf("expected constant %v", c))
	}
	if a, ok := s["enum"].([]any); ok {
		found := false
		for _, x := range a {
			found = found || reflect.DeepEqual(x, v)
		}
		if !found {
			return fail("value is not in enum")
		}
	}
	if a, ok := s["oneOf"].([]any); ok {
		n := 0
		for _, x := range a {
			if schemaCheck(x.(map[string]any), v, path, root) == nil {
				n++
			}
		}
		if n != 1 {
			return fail("expected exactly one supported variant (unsupported kinds/operations fail closed)")
		}
	}
	if typ, ok := s["type"].(string); ok {
		valid := false
		switch typ {
		case "object":
			_, valid = v.(map[string]any)
		case "array":
			_, valid = v.([]any)
		case "string":
			_, valid = v.(string)
		case "boolean":
			_, valid = v.(bool)
		case "integer":
			n, ok := v.(json.Number)
			valid = ok && !math.IsInf(number(n), 0) && math.Trunc(number(n)) == number(n)
		}
		if !valid {
			return fail("expected " + typ)
		}
	}
	if x, ok := v.(string); ok {
		if p, ok := s["pattern"].(string); ok && !regexp.MustCompile(p).MatchString(x) {
			return fail("invalid string pattern")
		}
		if n, ok := s["minLength"]; ok && len([]rune(x)) < int(number(n)) {
			return fail("string is too short")
		}
	}
	if n, ok := v.(json.Number); ok {
		f := number(n)
		if min, ok := s["minimum"]; ok && f < number(min) {
			return fail("below minimum")
		}
		if max, ok := s["maximum"]; ok && f > number(max) {
			return fail("above maximum")
		}
	}
	if a, ok := v.([]any); ok {
		if n, ok := s["minItems"]; ok && len(a) < int(number(n)) {
			return fail("too few items")
		}
		if n, ok := s["maxItems"]; ok && len(a) > int(number(n)) {
			return fail("too many items")
		}
		seen := map[string]bool{}
		for i, x := range a {
			if s["uniqueItems"] == true {
				b, _ := json.Marshal(x)
				if seen[string(b)] {
					return fail("duplicate array item")
				}
				seen[string(b)] = true
			}
			if item, ok := s["items"].(map[string]any); ok {
				if e := schemaCheck(item, x, fmt.Sprintf("%s[%d]", path, i), root); e != nil {
					return e
				}
			}
		}
	}
	if m, ok := v.(map[string]any); ok {
		if n, ok := s["minProperties"]; ok && len(m) < int(number(n)) {
			return fail("too few properties")
		}
		if n, ok := s["maxProperties"]; ok && len(m) > int(number(n)) {
			return fail("unsupported or excessive properties")
		}
		if a, ok := s["required"].([]any); ok {
			for _, k := range a {
				if _, ok := m[k.(string)]; !ok {
					return fail("missing " + k.(string))
				}
			}
		}
		props, _ := s["properties"].(map[string]any)
		for _, k := range keys(m) {
			x := m[k]
			if names, ok := s["propertyNames"].(map[string]any); ok {
				if e := schemaCheck(names, k, path+".<key>", root); e != nil {
					return e
				}
			}
			p, known := props[k]
			if !known {
				if s["additionalProperties"] == false {
					return fail("unknown field " + k)
				}
				p = s["additionalProperties"]
			}
			if sub, ok := p.(map[string]any); ok {
				if e := schemaCheck(sub, x, path+"."+k, root); e != nil {
					return e
				}
			}
		}
	}
	return nil
}

// Fail closed if this repository's schema starts using a keyword the evaluator
// does not implement. This prevents a later schema edit from silently weakening validation.
func auditSchema(s map[string]any, root map[string]any) error {
	allowed := map[string]bool{}
	for _, k := range strings.Fields("$schema $id title description $defs $ref type const enum oneOf properties required additionalProperties propertyNames items minItems maxItems uniqueItems minProperties maxProperties minLength pattern minimum maximum") {
		allowed[k] = true
	}
	for k, v := range s {
		if !allowed[k] {
			return fmt.Errorf("unsupported embedded schema keyword %s", k)
		}
		switch k {
		case "$ref":
			ref, ok := v.(string)
			if !ok || !strings.HasPrefix(ref, "#/$defs/") {
				return fmt.Errorf("unsupported schema reference")
			}
			if _, ok := root["$defs"].(map[string]any)[strings.TrimPrefix(ref, "#/$defs/")]; !ok {
				return fmt.Errorf("missing schema definition")
			}
		case "pattern":
			if _, e := regexp.Compile(v.(string)); e != nil {
				return e
			}
		case "properties", "$defs":
			for _, sub := range v.(map[string]any) {
				if e := auditSchema(sub.(map[string]any), root); e != nil {
					return e
				}
			}
		case "oneOf":
			for _, sub := range v.([]any) {
				if e := auditSchema(sub.(map[string]any), root); e != nil {
					return e
				}
			}
		case "items", "propertyNames", "additionalProperties":
			if sub, ok := v.(map[string]any); ok {
				if e := auditSchema(sub, root); e != nil {
					return e
				}
			}
		}
	}
	return nil
}
