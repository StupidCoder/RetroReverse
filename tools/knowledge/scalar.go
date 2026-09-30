package knowledge

import (
	"encoding/hex"
	"fmt"
	"math/big"
)

// Scalar preserves the exact numeric value. An unknown enum is not invalid data
// and is never coerced to the first known label. Structured live decoding is M3.
type Scalar struct {
	Value string `json:"value"`
	Raw   string `json:"raw"`
	Label string `json:"label,omitempty"`
	Known bool   `json:"known"`
}

func (p *Package) DecodeScalar(id string, bytes []byte) (Scalar, error) {
	t, ok := p.Types[id]
	if !ok {
		return Scalar{}, fmt.Errorf("unknown type %s", id)
	}
	base := t
	if t.Kind == "enum" {
		base = p.Types[t.Storage]
	} else if t.Kind != "integer" {
		return Scalar{}, fmt.Errorf("scalar decoding only supports integer/enum in M1")
	}
	if uint64(len(bytes)) != base.Bytes {
		return Scalar{}, fmt.Errorf("wrong scalar byte count")
	}
	data := append([]byte(nil), bytes...)
	if base.Endian == "little" {
		for i, j := 0, len(data)-1; i < j; i, j = i+1, j-1 {
			data[i], data[j] = data[j], data[i]
		}
	}
	n := new(big.Int).SetBytes(data)
	if base.Signed && data[0]&128 != 0 {
		n.Sub(n, new(big.Int).Lsh(big.NewInt(1), uint(base.Bytes*8)))
	}
	value := n.String()
	s := Scalar{Value: value, Raw: hex.EncodeToString(bytes), Known: true}
	if t.Kind == "enum" {
		s.Label, s.Known = t.Values[value]
	}
	return s, nil
}
