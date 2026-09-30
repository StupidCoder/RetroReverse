package knowledge

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"os"
)

// ResolutionError distinguishes identity failure from absent/ambiguous storage.
type ResolutionError struct {
	Status  string
	Message string
}

func (e *ResolutionError) Error() string           { return e.Status + ": " + e.Message }
func resolutionError(status, message string) error { return &ResolutionError{status, message} }
func ResolutionStatus(e error) string {
	if e == nil {
		return "resolved"
	}
	var r *ResolutionError
	if errors.As(e, &r) {
		return r.Status
	}
	return "unavailable"
}

// Immutable, bounded random-access views keep large media off the heap.
type assetView struct {
	size uint64
	read func(uint64, []byte) error
}

func (v assetView) slice(off, n uint64) (assetView, error) {
	if off > v.size || n > v.size-off {
		return assetView{}, fmt.Errorf("source range out of bounds")
	}
	return assetView{n, func(o uint64, b []byte) error {
		if o > n || uint64(len(b)) > n-o {
			return fmt.Errorf("read outside view")
		}
		return v.read(off+o, b)
	}}, nil
}

type SourceExtent struct {
	Offset uint64 `json:"offset"`
	Length uint64 `json:"length"`
}
type ResolutionStep struct {
	Extents       []SourceExtent `json:"extents,omitempty"`
	Location      string         `json:"location"`
	Kind          string         `json:"kind"`
	Parent        string         `json:"parent,omitempty"`
	Decoder       string         `json:"decoder,omitempty"`
	Path          string         `json:"path,omitempty"`
	Offset        uint64         `json:"offset,omitempty"`
	Size          uint64         `json:"size"`
	SHA256        string         `json:"sha256,omitempty"`
	First         uint64         `json:"first,omitempty"`
	Count         uint64         `json:"count,omitempty"`
	Stride        uint64         `json:"stride,omitempty"`
	PayloadOffset uint64         `json:"payloadOffset,omitempty"`
	PayloadSize   uint64         `json:"payloadSize,omitempty"`
}
type AssetResult struct {
	Sources        []AssetSource    `json:"sources"`
	DependencyRole string           `json:"dependencyRole,omitempty"`
	Status         string           `json:"status"`
	Package        string           `json:"package"`
	Revision       uint64           `json:"revision"`
	PackageSHA256  string           `json:"packageSHA256"`
	Release        string           `json:"release"`
	Asset          string           `json:"asset"`
	SHA256         string           `json:"sha256"`
	Size           uint64           `json:"size"`
	Format         string           `json:"format"`
	OutputSHA256   string           `json:"outputSHA256"`
	Chain          []ResolutionStep `json:"chain"`
	Dependencies   []AssetResult    `json:"dependencies,omitempty"`
	Data           []byte           `json:"-"`
}

// Resolver owns verified open handles. Close it after use; do not modify local media
// during a session. Every selected member/transform is independently hash-pinned.
type Resolver struct {
	p         *Package
	release   string
	files     map[string]*os.File
	media     map[string]assetView
	cache     map[string]assetView
	chain     map[string][]ResolutionStep
	remaining uint64
	allocated uint64
	entries   int
}

func OpenResolver(p *Package, release string, paths map[string]string) (*Resolver, error) {
	r := &Resolver{p: p, release: release, files: map[string]*os.File{}, media: map[string]assetView{}, cache: map[string]assetView{}, chain: map[string][]ResolutionStep{}, remaining: 2 << 30}
	fail := func(e error) (*Resolver, error) { r.Close(); return nil, e }
	rel, ok := p.Releases[release]
	if !ok || len(rel.Media) == 0 {
		return fail(fmt.Errorf("static resolver requires a raw-media release"))
	}
	if len(paths) != len(rel.Media) {
		return fail(fmt.Errorf("supply exactly the release's media roles"))
	}
	var total uint64
	for _, m := range rel.Media {
		total += m.Size
		if total > 16<<30 {
			return fail(fmt.Errorf("total media exceeds 16 GiB"))
		}
		if m.Size > 16<<30 {
			return fail(fmt.Errorf("media exceeds 16 GiB limit"))
		}
		f, e := os.Open(paths[m.Role])
		if e != nil {
			return fail(e)
		}
		r.files[m.Role] = f
		st, e := f.Stat()
		if e != nil || !st.Mode().IsRegular() || uint64(st.Size()) != m.Size {
			return fail(resolutionError("mismatch", "media size/type mismatch: "+m.Role))
		}
		h := sha256.New()
		if _, e = io.Copy(h, io.NewSectionReader(f, 0, int64(m.Size))); e != nil {
			return fail(e)
		}
		if hex.EncodeToString(h.Sum(nil)) != m.SHA256 {
			return fail(resolutionError("mismatch", "media hash mismatch: "+m.Role))
		}
		size := m.Size
		r.media[m.Role] = assetView{size, func(off uint64, b []byte) error {
			if off > size || uint64(len(b)) > size-off {
				return fmt.Errorf("media read exceeds bounds")
			}
			_, e := f.ReadAt(b, int64(off))
			return e
		}}
	}
	return r, nil
}
func (r *Resolver) Close() {
	for _, f := range r.files {
		f.Close()
	}
}
func (r *Resolver) read(v assetView, off, n uint64) ([]byte, error) {
	if n > 64<<20 || n > r.remaining || off > v.size || n > v.size-off {
		return nil, fmt.Errorf("read range/budget exceeded")
	}
	r.remaining -= n
	b := make([]byte, int(n))
	if e := v.read(off, b); e != nil {
		return nil, e
	}
	return b, nil
}
func (r *Resolver) entry() error {
	r.entries++
	if r.entries > 65536 {
		return fmt.Errorf("directory entry budget exceeded")
	}
	return nil
}
func (r *Resolver) hash(v assetView) (string, error) {
	h := sha256.New()
	for off := uint64(0); off < v.size; {
		n := uint64(65536)
		if n > v.size-off {
			n = v.size - off
		}
		b, e := r.read(v, off, n)
		if e != nil {
			return "", e
		}
		h.Write(b)
		off += n
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
func (r *Resolver) location(id string, depth int) (assetView, []ResolutionStep, error) {
	if depth >= 16 {
		return assetView{}, nil, fmt.Errorf("location depth exceeded")
	}
	if v, ok := r.cache[id]; ok {
		return v, r.chain[id], nil
	}
	l, ok := r.p.Locations[id]
	if !ok {
		return assetView{}, nil, fmt.Errorf("unknown location")
	}
	var v assetView
	var chain []ResolutionStep
	var e error
	step := ResolutionStep{Location: id, Kind: l.Kind, Parent: l.Parent, Decoder: l.Decoder, Path: l.Path, SHA256: l.SHA256}
	switch l.Kind {
	case "image":
		v = r.media[l.Role]
		off, _ := hexnum(l.Offset)
		n := l.Length
		if n == 0 && off <= v.size {
			n = v.size - off
		}
		v, e = v.slice(off, n)
		step.Path = l.Role
		step.Offset = off
	case "relative":
		v, chain, e = r.location(l.Base, depth+1)
		if e == nil {
			off, _ := hexnum(l.Offset)
			n := l.Length
			if n == 0 && off <= v.size {
				n = v.size - off
			}
			v, e = v.slice(off, n)
			step.Parent = l.Base
			step.Offset = off
		}
	case "sectors", "file", "member", "transform":
		v, chain, e = r.location(l.Parent, depth+1)
		if e != nil {
			break
		}
		if l.Kind == "sectors" {
			parent := v
			n := l.Count * l.PayloadSize
			step.First = l.First
			step.Count = l.Count
			step.Stride = l.Stride
			step.PayloadOffset = l.PayloadOffset
			step.PayloadSize = l.PayloadSize
			v = assetView{n, func(off uint64, b []byte) error {
				if off > n || uint64(len(b)) > n-off {
					return fmt.Errorf("sector read outside bounds")
				}
				for len(b) > 0 {
					within := off % l.PayloadSize
					take := int(l.PayloadSize - within)
					if take > len(b) {
						take = len(b)
					}
					physical := (l.First+off/l.PayloadSize)*l.Stride + l.PayloadOffset + within
					if e := parent.read(physical, b[:take]); e != nil {
						return e
					}
					off += uint64(take)
					b = b[take:]
				}
				return nil
			}}
		} else if l.Kind == "transform" {
			v, e = r.transform(v, l)
		} else {
			v, step.Extents, e = r.selectMember(v, l)
		}
		if e == nil && l.Kind != "sectors" {
			if v.size != l.Size {
				e = resolutionError("mismatch", "member/transform size mismatch")
			} else {
				var hash string
				hash, e = r.hash(v)
				if e == nil && hash != l.SHA256 {
					e = resolutionError("mismatch", "member/transform hash mismatch")
				}
			}
		}
	default:
		e = fmt.Errorf("location is not immutable asset storage")
	}
	if e != nil {
		return v, nil, fmt.Errorf("location %s: %w", id, e)
	}
	step.Size = v.size
	chain = append(append([]ResolutionStep{}, chain...), step)
	r.cache[id] = v
	r.chain[id] = chain
	return v, chain, nil
}
func (r *Resolver) ResolveAsset(id string) (AssetResult, error) {
	return r.asset(id, map[string]bool{}, 0)
}
func (r *Resolver) asset(id string, seen map[string]bool, depth int) (result AssetResult, err error) {
	defer func() {
		if err != nil {
			result.Status = ResolutionStatus(err)
		}
	}()
	if e := r.entry(); e != nil {
		return AssetResult{}, e
	}
	a, ok := r.p.Assets[id]
	out := AssetResult{Status: "unavailable", Package: r.p.ID, Revision: r.p.Revision, PackageSHA256: r.p.Hash, Release: r.release, Asset: id}
	if !ok || !contains(a.Releases, r.release) || seen[id] || depth >= 16 {
		return out, resolutionError("inactive", "unknown/inactive asset or dependency cycle")
	}
	seen[id] = true
	defer delete(seen, id)
	if a.Size > 64<<20 || r.allocated+a.Size > 128<<20 {
		return out, fmt.Errorf("asset allocation budget exceeded")
	}
	r.allocated += a.Size
	data := make([]byte, 0, int(a.Size))
	for _, s := range a.Sources {
		v, chain, e := r.location(s.Location, 0)
		if e != nil {
			return out, e
		}
		b, e := r.read(v, 0, s.Length)
		if e != nil {
			return out, e
		}
		data = append(data, b...)
		out.Chain = append(out.Chain, chain...)
	}
	out.SHA256 = SHA256(data)
	out.Size = uint64(len(data))
	if out.SHA256 != a.SHA256 {
		return out, resolutionError("mismatch", "asset hash mismatch")
	}
	var palette []byte
	for _, d := range a.Dependencies {
		x, e := r.asset(d.Asset, seen, depth+1)
		if e != nil {
			return out, e
		}
		if d.Role == "palette" {
			palette = x.Data
		}
		x.DependencyRole = d.Role
		x.Data = nil
		out.Dependencies = append(out.Dependencies, x)
	}
	out.Sources = a.Sources
	out.Data = data
	out.Format = a.Format
	if a.Decoder == "indexed8" {
		if len(palette) != 768 || a.Parameters == nil {
			return out, fmt.Errorf("invalid palette")
		}
		w, h := int(a.Parameters.Width), int(a.Parameters.Height)
		pal := make(color.Palette, 256)
		for i := range pal {
			pal[i] = color.NRGBA{palette[i*3], palette[i*3+1], palette[i*3+2], 255}
		}
		im := image.NewPaletted(image.Rect(0, 0, w, h), pal)
		copy(im.Pix, data)
		var b bytes.Buffer
		if e := png.Encode(&b, im); e != nil {
			return out, e
		}
		out.Data = b.Bytes()
		out.Format = "png"
	}
	out.OutputSHA256 = SHA256(out.Data)
	out.Status = "resolved"
	return out, nil
}
