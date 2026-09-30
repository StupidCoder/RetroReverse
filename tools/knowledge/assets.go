package knowledge

import (
	"fmt"
	"path"
	"strings"
	"unicode/utf8"
)

// Asset recipes are data. Decoder implementations and versions are fixed in Go.
type AssetSource struct {
	Location string `json:"location"`
	Length   uint64 `json:"length"`
}
type AssetDependency struct {
	Asset string `json:"asset"`
	Role  string `json:"role"`
}
type Artifact struct {
	Path   string `json:"path"`
	Format string `json:"format"`
	Size   uint64 `json:"size"`
	SHA256 string `json:"sha256"`
}
type Asset struct {
	Label       string        `json:"label"`
	Description string        `json:"description"`
	Category    string        `json:"category"`
	Format      string        `json:"format"`
	Decoder     string        `json:"decoder"`
	Version     int           `json:"version"`
	Sources     []AssetSource `json:"sources"`
	Size        uint64        `json:"size"`
	SHA256      string        `json:"sha256"`
	Parameters  *struct {
		Width  uint64 `json:"width"`
		Height uint64 `json:"height"`
	} `json:"parameters,omitempty"`
	Dependencies []AssetDependency `json:"dependencies"`
	Artifacts    []Artifact        `json:"artifacts"`
	Releases     []string          `json:"releases"`
	Evidence     []string          `json:"evidence"`
}

type DecoderInfo struct {
	ID      string `json:"id"`
	Version int    `json:"version"`
	Kind    string `json:"kind"`
}

func Decoders() []DecoderInfo {
	return []DecoderInfo{{"ncsd", 1, "member"}, {"ncch", 1, "member"}, {"sarc", 1, "member"}, {"romfs", 1, "file"}, {"iso9660", 1, "file"}, {"yaz0", 1, "transform"}, {"bytes", 1, "asset"}, {"indexed8", 1, "asset"}}
}
func registered(id, kind string, version int) bool {
	for _, d := range Decoders() {
		if d.ID == id && d.Kind == kind && d.Version == version {
			return true
		}
	}
	return false
}
func assetPath(s string) bool {
	if !utf8.ValidString(s) {
		return false
	}
	for _, r := range s {
		if r < 32 || r == 127 {
			return false
		}
	}
	return s != "" && len(s) <= 1024 && path.Clean(s) == s && !strings.HasPrefix(s, "/") && !strings.ContainsAny(s, "\\:\x00") && s != "." && s != ".." && !strings.HasPrefix(s, "../")
}
func (p *Package) resolveAssetLocation(l Location, release string, seen map[string]bool, depth int) (Span, error) {
	s, e := p.resolve(l.Parent, release, seen, depth+1)
	if e != nil {
		return s, e
	}
	if s.Kind != "image" && s.Kind != "asset" {
		return s, fmt.Errorf("asset chain requires immutable media")
	}
	n := s.Limit - s.Offset
	if l.Kind == "sectors" {
		if l.PayloadOffset >= l.Stride || l.PayloadSize > l.Stride-l.PayloadOffset || l.First > n/l.Stride || l.Count > (n/l.Stride)-l.First {
			return s, fmt.Errorf("sector geometry exceeds parent")
		}
		return Span{Kind: "asset", Limit: l.Count * l.PayloadSize}, nil
	}
	if !registered(l.Decoder, l.Kind, l.Version) {
		return s, fmt.Errorf("unsupported %s decoder %s version %d", l.Kind, l.Decoder, l.Version)
	}
	if l.Kind != "transform" && !assetPath(l.Path) {
		return s, fmt.Errorf("unsafe member path")
	}
	if l.Kind == "transform" && l.Size > 64<<20 {
		return s, fmt.Errorf("decoded size exceeds 64 MiB")
	}
	return Span{Kind: "asset", Limit: l.Size}, nil
}
func (p *Package) validateAssets() error {
	for id, a := range p.Assets {
		if !registered(a.Decoder, "asset", a.Version) {
			return fmt.Errorf("asset %s: unknown decoder", id)
		}
		if e := p.evidence(a.Evidence); e != nil {
			return e
		}
		var total uint64
		for _, s := range a.Sources {
			total += s.Length
			for _, r := range a.Releases {
				span, e := p.Resolve(s.Location, r)
				if e != nil {
					return e
				}
				if (span.Kind != "image" && span.Kind != "asset") || s.Length > span.Limit-span.Offset {
					return fmt.Errorf("asset %s: invalid source extent", id)
				}
			}
		}
		if total != a.Size {
			return fmt.Errorf("asset %s: fragment sizes do not equal asset size", id)
		}
		palettes := 0
		for _, d := range a.Dependencies {
			target, ok := p.Assets[d.Asset]
			if !ok {
				return fmt.Errorf("unknown asset dependency %s", d.Asset)
			}
			for _, r := range a.Releases {
				if !contains(target.Releases, r) {
					return fmt.Errorf("asset dependency release mismatch")
				}
			}
			if d.Role == "palette" {
				palettes++
				if target.Size != 768 || target.Category != "palette" || target.Decoder != "bytes" {
					return fmt.Errorf("indexed palette must contain 768 RGB bytes")
				}
			}
		}
		if a.Decoder == "indexed8" {
			if a.Parameters == nil || a.Parameters.Width*a.Parameters.Height != a.Size || palettes != 1 {
				return fmt.Errorf("indexed8 requires dimensions and one RGB palette")
			}
		} else if a.Parameters != nil {
			return fmt.Errorf("bytes decoder has no parameters")
		}
		for _, f := range a.Artifacts {
			if !assetPath(f.Path) || strings.ContainsAny(f.Path, "?#%") || !strings.HasPrefix(f.Path, "public/"+p.ID+"/") || path.Ext(f.Path) != "."+f.Format {
				return fmt.Errorf("invalid asset artifact path")
			}
		}
	}
	visits := 0
	var walk func(string, map[string]bool, int) error
	walk = func(id string, seen map[string]bool, depth int) error {
		if seen[id] || depth >= 16 {
			return fmt.Errorf("asset dependency cycle/depth")
		}
		visits++
		if visits > 65536 {
			return fmt.Errorf("dependency validation budget exceeded")
		}
		seen[id] = true
		defer delete(seen, id)
		for _, d := range p.Assets[id].Dependencies {
			if e := walk(d.Asset, seen, depth+1); e != nil {
				return e
			}
		}
		return nil
	}
	for id := range p.Assets {
		if e := walk(id, map[string]bool{}, 0); e != nil {
			return e
		}
	}
	return nil
}
