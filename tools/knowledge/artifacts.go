package knowledge

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"image/png"
	"os"
	"path/filepath"
	"strings"
)

// VerifyArtifacts checks curated outputs; it does not claim to regenerate a GLB
// from a BCH. That remains the researched game exporter named in the evidence.
func (p *Package) VerifyArtifacts(siteRoot string) error {
	root, e := filepath.EvalSymlinks(siteRoot)
	if e != nil {
		return e
	}
	root, e = filepath.Abs(root)
	if e != nil {
		return e
	}
	for id, a := range p.Assets {
		for _, f := range a.Artifacts {
			if !assetPath(f.Path) {
				return fmt.Errorf("unsafe artifact path")
			}
			file, e := filepath.EvalSymlinks(filepath.Join(root, filepath.FromSlash(f.Path)))
			if e != nil {
				return e
			}
			rel, e := filepath.Rel(root, file)
			if e != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
				return fmt.Errorf("artifact escapes site root")
			}
			st, e := os.Stat(file)
			if e != nil || !st.Mode().IsRegular() || uint64(st.Size()) != f.Size {
				return fmt.Errorf("asset %s: artifact size/type mismatch", id)
			}
			n, h, e := HashFile(file)
			if e != nil {
				return e
			}
			if n != f.Size || h != f.SHA256 {
				return fmt.Errorf("asset %s: artifact hash mismatch", id)
			}
			b, e := os.ReadFile(file)
			if e != nil {
				return e
			}
			if e = validateArtifact(f.Format, b); e != nil {
				return fmt.Errorf("asset %s: %w", id, e)
			}
		}
	}
	return nil
}
func validateArtifact(format string, b []byte) error {
	if format == "png" {
		c, e := png.DecodeConfig(bytes.NewReader(b))
		if e != nil {
			return e
		}
		if c.Width > 4096 || c.Height > 4096 {
			return fmt.Errorf("preview dimensions exceed limit")
		}
		_, e = png.Decode(bytes.NewReader(b))
		return e
	}
	if format != "glb" || len(b) < 20 || string(b[:4]) != "glTF" || binary.LittleEndian.Uint32(b[4:]) != 2 || int(binary.LittleEndian.Uint32(b[8:])) != len(b) {
		return fmt.Errorf("invalid GLB header")
	}
	jsonSeen := false
	binSeen := false
	for off := 12; off < len(b); {
		if off+8 > len(b) {
			return fmt.Errorf("truncated GLB chunk")
		}
		n := uint64(binary.LittleEndian.Uint32(b[off:]))
		kind := string(b[off+4 : off+8])
		off += 8
		if n%4 != 0 || n > uint64(len(b)-off) {
			return fmt.Errorf("invalid GLB chunk length")
		}
		chunk := b[off : off+int(n)]
		if !jsonSeen {
			if kind != "JSON" || !json.Valid(chunk) {
				return fmt.Errorf("GLB requires initial JSON chunk")
			}
			var doc map[string]any
			if e := json.Unmarshal(chunk, &doc); e != nil {
				return e
			}
			asset, ok := doc["asset"].(map[string]any)
			if !ok || asset["version"] != "2.0" {
				return fmt.Errorf("GLB asset version")
			}
			for _, group := range []string{"buffers", "images"} {
				if entries, ok := doc[group].([]any); ok {
					for _, entry := range entries {
						m, ok := entry.(map[string]any)
						if !ok {
							return fmt.Errorf("invalid GLB resource")
						}
						if _, ok = m["uri"]; ok {
							return fmt.Errorf("curated GLB must embed resources")
						}
					}
				}
			}
			jsonSeen = true
		} else if kind != "BIN\x00" || binSeen {
			return fmt.Errorf("unsupported/duplicate GLB chunk")
		} else {
			binSeen = true
		}
		off += int(n)
	}
	return nil
}
