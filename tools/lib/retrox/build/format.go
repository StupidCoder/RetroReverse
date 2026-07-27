package build

import (
	"encoding/json"
	"fmt"
	"strings"
)

// marshalPretty renders Retro-X JSON for humans: objects are indented one key
// per line, but arrays of scalars (tile cells, matrices, durations, palettes)
// stay on a single line, and arrays of small arrays (steps, points) inline
// when short — a big nested table (block tiles, collision profiles) gets one
// row per line instead of one NUMBER per line.
func marshalPretty(v any) ([]byte, error) {
	compact, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	root, err := parseNode(json.NewDecoder(strings.NewReader(string(compact))))
	if err != nil {
		return nil, err
	}
	var sb strings.Builder
	renderNode(&sb, root, 0)
	sb.WriteByte('\n')
	return []byte(sb.String()), nil
}

const (
	nodeScalar = iota
	nodeArray
	nodeObject
)

type jnode struct {
	kind int
	raw  string // scalar: literal JSON text
	arr  []jnode
	keys []string
	vals []jnode
}

// parseNode consumes one JSON value from the decoder, preserving key order
// and exact number formatting.
func parseNode(dec *json.Decoder) (jnode, error) {
	dec.UseNumber()
	return parseValue(dec)
}

func parseValue(dec *json.Decoder) (jnode, error) {
	tok, err := dec.Token()
	if err != nil {
		return jnode{}, err
	}
	switch t := tok.(type) {
	case json.Delim:
		switch t {
		case '{':
			n := jnode{kind: nodeObject}
			for dec.More() {
				kt, err := dec.Token()
				if err != nil {
					return jnode{}, err
				}
				key, ok := kt.(string)
				if !ok {
					return jnode{}, fmt.Errorf("object key is %T", kt)
				}
				val, err := parseValue(dec)
				if err != nil {
					return jnode{}, err
				}
				n.keys = append(n.keys, key)
				n.vals = append(n.vals, val)
			}
			if _, err := dec.Token(); err != nil { // closing '}'
				return jnode{}, err
			}
			return n, nil
		case '[':
			n := jnode{kind: nodeArray}
			for dec.More() {
				el, err := parseValue(dec)
				if err != nil {
					return jnode{}, err
				}
				n.arr = append(n.arr, el)
			}
			if _, err := dec.Token(); err != nil { // closing ']'
				return jnode{}, err
			}
			return n, nil
		}
		return jnode{}, fmt.Errorf("unexpected delim %v", t)
	case json.Number:
		return jnode{kind: nodeScalar, raw: t.String()}, nil
	case string:
		b, err := json.Marshal(t)
		if err != nil {
			return jnode{}, err
		}
		return jnode{kind: nodeScalar, raw: string(b)}, nil
	case bool:
		if t {
			return jnode{kind: nodeScalar, raw: "true"}, nil
		}
		return jnode{kind: nodeScalar, raw: "false"}, nil
	case nil:
		return jnode{kind: nodeScalar, raw: "null"}, nil
	}
	return jnode{}, fmt.Errorf("unexpected token %T", tok)
}

// inlineLimit is the longest an array-of-arrays may be to stay on one line.
const inlineLimit = 120

func containsObject(n jnode) bool {
	switch n.kind {
	case nodeObject:
		return true
	case nodeArray:
		for _, el := range n.arr {
			if containsObject(el) {
				return true
			}
		}
	}
	return false
}

func scalarsOnly(n jnode) bool {
	for _, el := range n.arr {
		if el.kind != nodeScalar {
			return false
		}
	}
	return true
}

// inline renders a node without any newlines.
func inline(n jnode) string {
	switch n.kind {
	case nodeScalar:
		return n.raw
	case nodeArray:
		parts := make([]string, len(n.arr))
		for i, el := range n.arr {
			parts[i] = inline(el)
		}
		return "[" + strings.Join(parts, ", ") + "]"
	default:
		parts := make([]string, len(n.keys))
		for i, k := range n.keys {
			kb, _ := json.Marshal(k)
			parts[i] = string(kb) + ": " + inline(n.vals[i])
		}
		return "{" + strings.Join(parts, ", ") + "}"
	}
}

func renderNode(sb *strings.Builder, n jnode, depth int) {
	ind := strings.Repeat(" ", depth)
	sub := strings.Repeat(" ", depth+1)
	switch n.kind {
	case nodeScalar:
		sb.WriteString(n.raw)
	case nodeArray:
		if len(n.arr) == 0 {
			sb.WriteString("[]")
			return
		}
		if !containsObject(n) {
			// Scalar-only arrays always inline; arrays of arrays inline when short.
			if s := inline(n); scalarsOnly(n) || len(s)+depth <= inlineLimit {
				sb.WriteString(s)
				return
			}
			// Long nested table: one row per line, rows themselves inline.
			sb.WriteString("[\n")
			for i, el := range n.arr {
				sb.WriteString(sub)
				renderNode(sb, el, depth+1)
				if i < len(n.arr)-1 {
					sb.WriteByte(',')
				}
				sb.WriteByte('\n')
			}
			sb.WriteString(ind + "]")
			return
		}
		sb.WriteString("[\n")
		for i, el := range n.arr {
			sb.WriteString(sub)
			renderNode(sb, el, depth+1)
			if i < len(n.arr)-1 {
				sb.WriteByte(',')
			}
			sb.WriteByte('\n')
		}
		sb.WriteString(ind + "]")
	default:
		if len(n.keys) == 0 {
			sb.WriteString("{}")
			return
		}
		sb.WriteString("{\n")
		for i, k := range n.keys {
			kb, _ := json.Marshal(k)
			sb.WriteString(sub + string(kb) + ": ")
			renderNode(sb, n.vals[i], depth+1)
			if i < len(n.keys)-1 {
				sb.WriteByte(',')
			}
			sb.WriteByte('\n')
		}
		sb.WriteString(ind + "}")
	}
}
