// Package knowledge validates the implemented knowledge profile and compiles read-only
// annotations. It does not run guest code, evaluate applicability text or tours.
package knowledge

import (
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/url"
	"os"
	"strconv"
	"strings"
	"unicode/utf8"
)

type Space struct {
	Kind      string `json:"kind"`
	Processor string `json:"processor,omitempty"`
	Region    string `json:"region,omitempty"`
	Size      uint64 `json:"size"`
}
type Location struct {
	Kind    string `json:"kind"`
	Space   string `json:"space,omitempty"`
	Address string `json:"address,omitempty"`
	Offset  string `json:"offset,omitempty"`
	Role    string `json:"role,omitempty"`
	Base    string `json:"base,omitempty"`
}
type Field struct {
	Name        string `json:"name"`
	Offset      uint64 `json:"offset"`
	Type        string `json:"type"`
	Description string `json:"description,omitempty"`
}
type Bits struct {
	Offset uint64 `json:"offset"`
	Width  uint64 `json:"width"`
	Label  string `json:"label"`
}
type Type struct {
	Kind         string            `json:"kind"`
	Bytes        uint64            `json:"bytes,omitempty"`
	Signed       bool              `json:"signed,omitempty"`
	Endian       string            `json:"endian,omitempty"`
	Storage      string            `json:"storage,omitempty"`
	Values       map[string]string `json:"values,omitempty"`
	Bits         map[string]Bits   `json:"bits,omitempty"`
	FractionBits uint64            `json:"fractionBits,omitempty"`
	Size         uint64            `json:"size,omitempty"`
	Fields       []Field           `json:"fields,omitempty"`
	Element      string            `json:"element,omitempty"`
	Count        uint64            `json:"count,omitempty"`
	Stride       uint64            `json:"stride,omitempty"`
}
type Source struct {
	Path        string `json:"path"`
	Description string `json:"description,omitempty"`
}
type Evidence struct {
	Status      string `json:"status"`
	Source      string `json:"source"`
	Description string `json:"description"`
	Limitations string `json:"limitations,omitempty"`
}
type Occupancy struct {
	Path      []string `json:"path"`
	NotEquals string   `json:"notEquals"`
}
type Definition struct {
	Occupancy        *Occupancy `json:"occupancy,omitempty"`
	RelatedFunctions []string   `json:"relatedFunctions,omitempty"`
	Label            string     `json:"label"`
	Description      string     `json:"description,omitempty"`
	Releases         []string   `json:"releases"`
	Evidence         []string   `json:"evidence"`
	Applicability    string     `json:"applicability"`
	Location         string     `json:"location,omitempty"`
	Length           uint64     `json:"length,omitempty"`
	MemoryLabel      bool       `json:"memoryLabel,omitempty"`
	Type             string     `json:"type,omitempty"`
	Entry            string     `json:"entry,omitempty"`
	ISA              string     `json:"isa,omitempty"`
}
type Annotation struct {
	Function string   `json:"function"`
	Offset   string   `json:"offset"`
	Text     string   `json:"text"`
	Evidence []string `json:"evidence"`
}
type Game struct {
	Name   string `json:"name"`
	System struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"system"`
}
type Package struct {
	SchemaVersion int                   `json:"schemaVersion"`
	ID            string                `json:"id"`
	Revision      uint64                `json:"revision"`
	Game          Game                  `json:"game"`
	Releases      map[string]Release    `json:"releases"`
	Sources       map[string]Source     `json:"sources"`
	Evidence      map[string]Evidence   `json:"evidence"`
	Spaces        map[string]Space      `json:"spaces"`
	Locations     map[string]Location   `json:"locations"`
	Types         map[string]Type       `json:"types"`
	Regions       map[string]Definition `json:"regions"`
	State         map[string]Definition `json:"state"`
	Functions     map[string]Definition `json:"functions"`
	Tours         map[string]Tour       `json:"tours"`
	Annotations   map[string]Annotation `json:"annotations"`
	Raw           json.RawMessage       `json:"-"`
	Hash          string                `json:"-"`
}

func Read(path string) (*Package, error) {
	f, e := os.Open(path)
	if e != nil {
		return nil, e
	}
	defer f.Close()
	b, e := io.ReadAll(io.LimitReader(f, MaxPackageBytes+1))
	if e != nil {
		return nil, e
	}
	return Parse(b)
}
func Parse(b []byte) (*Package, error) {
	if !utf8.Valid(b) {
		return nil, fmt.Errorf("invalid UTF-8 package")
	}
	v, e := decode(b)
	if e != nil {
		return nil, e
	}
	schema, e := decode(Schema)
	if e != nil {
		return nil, e
	}
	root := schema.(map[string]any)
	if e = auditSchema(root, root); e != nil {
		return nil, e
	}
	if e = schemaCheck(root, v, "$", root); e != nil {
		return nil, e
	}
	var p Package
	if e = json.Unmarshal(b, &p); e != nil {
		return nil, e
	}
	p.Raw = append([]byte(nil), b...)
	p.Hash = SHA256(b)
	if e = p.validate(); e != nil {
		return nil, e
	}
	return &p, nil
}
func hexnum(s string) (uint64, error) {
	n, e := strconv.ParseUint(strings.TrimPrefix(s, "0x"), 16, 64)
	if e != nil {
		return 0, fmt.Errorf("invalid/overflowing address %q", s)
	}
	return n, nil
}

type Span struct {
	Kind   string
	Region string
	Space  string
	Role   string
	Offset uint64
	Limit  uint64
}

// Resolve is static only. CPU spans describe logical addresses, never guessed physical aliases.
func (p *Package) Resolve(id, release string) (Span, error) {
	if _, ok := p.Releases[release]; !ok {
		return Span{}, fmt.Errorf("unknown release %s", release)
	}
	return p.resolve(id, release, map[string]bool{}, 0)
}
func (p *Package) resolve(id, release string, seen map[string]bool, depth int) (Span, error) {
	if seen[id] || depth >= 16 {
		return Span{}, fmt.Errorf("location cycle/depth at %s", id)
	}
	seen[id] = true
	defer delete(seen, id)
	l, ok := p.Locations[id]
	if !ok {
		return Span{}, fmt.Errorf("unknown location %s", id)
	}
	if l.Kind == "relative" {
		s, e := p.resolve(l.Base, release, seen, depth+1)
		if e != nil {
			return s, e
		}
		n, e := hexnum(l.Offset)
		if e != nil {
			return s, e
		}
		if n >= s.Limit-s.Offset {
			return s, fmt.Errorf("relative location exceeds space")
		}
		s.Offset += n
		return s, nil
	}
	s := Span{Kind: l.Kind, Space: l.Space, Role: l.Role}
	var e error
	if l.Kind == "cpu" {
		s.Offset, e = hexnum(l.Address)
	} else {
		s.Offset, e = hexnum(l.Offset)
	}
	if e != nil {
		return s, e
	}
	if l.Kind == "image" {
		r, ok := p.Releases[release]
		if !ok {
			return s, fmt.Errorf("unknown release %s", release)
		}
		found := false
		for _, m := range r.Media {
			if m.Role == l.Role {
				s.Limit = m.Size
				found = true
			}
		}
		if !found {
			return s, fmt.Errorf("missing raw media role %s in %s", l.Role, release)
		}
	} else {
		x, ok := p.Spaces[l.Space]
		if !ok || x.Kind != l.Kind {
			return s, fmt.Errorf("wrong/unknown space %s", l.Space)
		}
		s.Limit = x.Size
		s.Region = x.Region
	}
	if s.Offset >= s.Limit {
		return s, fmt.Errorf("location %s exceeds space", id)
	}
	return s, nil
}
func (p *Package) typeSize(id string, seen map[string]bool, cache map[string]uint64) (result uint64, err error) {
	if size, ok := cache[id]; ok {
		return size, nil
	}
	defer func() {
		if err == nil {
			cache[id] = result
		}
	}()
	if seen[id] || len(seen) >= 16 {
		return 0, fmt.Errorf("type cycle/depth at %s", id)
	}
	seen[id] = true
	defer delete(seen, id)
	t, ok := p.Types[id]
	if !ok {
		return 0, fmt.Errorf("unknown type %s", id)
	}
	switch t.Kind {
	case "integer":
		return t.Bytes, nil
	case "enum", "bitfield", "fixed":
		base, ok := p.Types[t.Storage]
		if !ok || base.Kind != "integer" {
			return 0, fmt.Errorf("%s storage must name an integer type", id)
		}
		size := base.Bytes
		if t.Kind == "fixed" && t.FractionBits >= size*8 {
			return 0, fmt.Errorf("fixed fraction exceeds storage")
		}
		if t.Kind == "enum" {
			max := new(big.Int).Lsh(big.NewInt(1), uint(size*8))
			min := big.NewInt(0)
			if base.Signed {
				max.Rsh(max, 1)
				min.Neg(new(big.Int).Set(max))
			}
			max.Sub(max, big.NewInt(1))
			for v := range t.Values {
				n, ok := new(big.Int).SetString(v, 10)
				if !ok || n.Cmp(min) < 0 || n.Cmp(max) > 0 {
					return 0, fmt.Errorf("enum %s value exceeds storage", id)
				}
			}
		}
		if t.Kind == "bitfield" {
			used := uint64(0)
			for _, b := range t.Bits {
				if b.Offset >= size*8 || b.Width > size*8-b.Offset {
					return 0, fmt.Errorf("bitfield exceeds storage")
				}
				mask := ^uint64(0)
				if b.Width < 64 {
					mask = (uint64(1) << b.Width) - 1
				}
				mask <<= b.Offset
				if used&mask != 0 {
					return 0, fmt.Errorf("overlapping bitfield")
				}
				used |= mask
			}
		}
		return size, nil
	case "array":
		n, e := p.typeSize(t.Element, seen, cache)
		if e != nil {
			return 0, e
		}
		if t.Stride < n || t.Count > 9007199254740991/t.Stride {
			return 0, fmt.Errorf("invalid array stride/extent")
		}
		return t.Count * t.Stride, nil
	case "struct":
		names := map[string]bool{}
		type extent struct{ a, b uint64 }
		spans := []extent{}
		for _, f := range t.Fields {
			if names[f.Name] {
				return 0, fmt.Errorf("duplicate field %s", f.Name)
			}
			names[f.Name] = true
			n, e := p.typeSize(f.Type, seen, cache)
			if e != nil {
				return 0, e
			}
			if f.Offset > t.Size || n > t.Size-f.Offset {
				return 0, fmt.Errorf("field %s exceeds struct", f.Name)
			}
			for _, s := range spans {
				if f.Offset < s.b && s.a < f.Offset+n {
					return 0, fmt.Errorf("overlapping struct fields")
				}
			}
			spans = append(spans, extent{f.Offset, f.Offset + n})
		}
		return t.Size, nil
	}
	return 0, fmt.Errorf("unsupported type %s", t.Kind)
}
func (p *Package) TypeSize(id string) (uint64, error) {
	return p.typeSize(id, map[string]bool{}, map[string]uint64{})
}
func safeSource(path string) bool {
	u, e := url.Parse(path)
	return e == nil && u.Scheme == "" && u.Host == "" && !strings.Contains(path, "\\") && !strings.HasPrefix(path, "//") && !strings.Contains(u.Path, "..")
}
func (p *Package) evidence(ids []string) error {
	for _, id := range ids {
		if _, ok := p.Evidence[id]; !ok {
			return fmt.Errorf("unknown evidence %s", id)
		}
	}
	return nil
}
func (p *Package) validate() error {
	for _, id := range keys(p.Sources) {
		if !safeSource(p.Sources[id].Path) {
			return fmt.Errorf("unsafe source path %s", id)
		}
	}
	for _, id := range keys(p.Evidence) {
		if _, ok := p.Sources[p.Evidence[id].Source]; !ok {
			return fmt.Errorf("unknown source on evidence %s", id)
		}
	}
	for _, id := range keys(p.Releases) {
		r := p.Releases[id]
		seen := map[string]bool{}
		for _, m := range r.Media {
			if seen[m.Role] {
				return fmt.Errorf("duplicate media role %s", m.Role)
			}
			seen[m.Role] = true
		}
		if r.FileSet != nil {
			_, h, e := Manifest(r.FileSet.Members, r.FileSet.PathPolicy)
			if e != nil {
				return e
			}
			if h != r.FileSet.SHA256 {
				return fmt.Errorf("file-set hash mismatch in %s", id)
			}
		}
	}
	// Reject overlapping release identities: otherwise matching could depend on selected extras.
	ids := keys(p.Releases)
	for i, a := range ids {
		for _, b := range ids[i+1:] {
			x, y := p.Releases[a], p.Releases[b]
			if len(x.Media) > 0 && len(y.Media) > 0 {
				conflict := false
				for _, m := range x.Media {
					for _, n := range y.Media {
						if m.Role == n.Role && (m.Size != n.Size || m.SHA256 != n.SHA256) {
							conflict = true
						}
					}
				}
				if !conflict {
					return fmt.Errorf("ambiguous raw release identities %s and %s", a, b)
				}
			}
			if x.FileSet != nil && y.FileSet != nil && x.FileSet.PathPolicy == y.FileSet.PathPolicy && x.FileSet.SHA256 == y.FileSet.SHA256 {
				return fmt.Errorf("duplicate file-set releases")
			}
		}
	}
	for _, id := range keys(p.Types) {
		if _, e := p.TypeSize(id); e != nil {
			return fmt.Errorf("types.%s: %w", id, e)
		}
	}
	for _, id := range keys(p.Locations) {
		for _, release := range ids {
			if _, e := p.Resolve(id, release); e != nil {
				return fmt.Errorf("locations.%s: %w", id, e)
			}
		}
	}
	for _, collection := range []struct {
		name string
		data map[string]Definition
	}{{"regions", p.Regions}, {"state", p.State}, {"functions", p.Functions}} {
		for _, id := range keys(collection.data) {
			d := collection.data[id]
			if e := p.evidence(d.Evidence); e != nil {
				return e
			}
			loc := d.Location
			n := d.Length
			if collection.name == "state" {
				var e error
				n, e = p.TypeSize(d.Type)
				if e != nil {
					return e
				}
			}
			for _, function := range d.RelatedFunctions {
				f, ok := p.Functions[function]
				if !ok {
					return fmt.Errorf("unknown related function %s", function)
				}
				for _, r := range d.Releases {
					if !contains(f.Releases, r) {
						return fmt.Errorf("related function release mismatch")
					}
				}
			}
			if d.Occupancy != nil {
				t := p.Types[d.Type]
				if t.Kind != "array" {
					return fmt.Errorf("occupancy requires an array")
				}
				t = p.Types[t.Element]
				for _, field := range d.Occupancy.Path {
					found := false
					if t.Kind == "struct" {
						for _, f := range t.Fields {
							if f.Name == field {
								t = p.Types[f.Type]
								found = true
								break
							}
						}
					}
					if !found {
						return fmt.Errorf("invalid occupancy field %s", field)
					}
				}
				if t.Kind == "enum" {
					t = p.Types[t.Storage]
				}
				if t.Kind != "integer" {
					return fmt.Errorf("occupancy must compare integer storage")
				}
				n, ok := new(big.Int).SetString(d.Occupancy.NotEquals, 10)
				max := new(big.Int).Lsh(big.NewInt(1), uint(t.Bytes*8))
				min := big.NewInt(0)
				if t.Signed {
					max.Rsh(max, 1)
					min.Neg(new(big.Int).Set(max))
				}
				max.Sub(max, big.NewInt(1))
				if !ok || n.Cmp(min) < 0 || n.Cmp(max) > 0 {
					return fmt.Errorf("occupancy value exceeds storage")
				}
			}
			if collection.name == "functions" {
				loc = d.Entry
				n = 1
			}
			for _, r := range d.Releases {
				if _, ok := p.Releases[r]; !ok {
					return fmt.Errorf("unknown release %s", r)
				}
				s, e := p.Resolve(loc, r)
				if e != nil {
					return e
				}
				if n > s.Limit-s.Offset {
					return fmt.Errorf("%s.%s exceeds resolved space", collection.name, id)
				}
				if d.MemoryLabel && s.Kind != "physical" {
					return fmt.Errorf("memory label %s requires an explicit physical location", id)
				}
			}
		}
	}
	for _, id := range keys(p.Annotations) {
		a := p.Annotations[id]
		f, ok := p.Functions[a.Function]
		if !ok {
			return fmt.Errorf("unknown function %s", a.Function)
		}
		if e := p.evidence(a.Evidence); e != nil {
			return e
		}
		n, e := hexnum(a.Offset)
		if e != nil {
			return e
		}
		for _, r := range f.Releases {
			s, e := p.Resolve(f.Entry, r)
			if e != nil {
				return e
			}
			if n >= s.Limit-s.Offset {
				return fmt.Errorf("annotation exceeds space")
			}
		}
	}
	return p.validateTours()
}
