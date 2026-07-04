// Package mkds holds the Mario-Kart-DS-specific asset plumbing shared by the
// extraction commands: loading NSBMD models and NSBTX textures whether they sit as
// loose files (the menu models) or inside a course's NARC archives, where the
// course geometry lives in <name>.carc and its textures in the sibling
// <name>Tex.carc.
package mkds

import (
	"os"
	"path/filepath"
	"strings"

	"retroreverse.com/tools/nds"
	"retroreverse.com/tools/nds/nitro"
)

// LoadModels returns every model in path: a .nsbmd directly, or every BMD0
// sub-file of a .carc/NARC.
func LoadModels(path string) ([]nitro.Model, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	data = nds.Decompress(data)
	if len(data) >= 4 && string(data[:4]) == "BMD0" {
		return nitro.ParseNSBMD(data)
	}
	files, err := nds.ParseNARC(data)
	if err != nil {
		return nil, err
	}
	var out []nitro.Model
	for _, f := range files {
		if len(f) >= 4 && string(f[:4]) == "BMD0" {
			ms, err := nitro.ParseNSBMD(f)
			if err != nil {
				continue
			}
			out = append(out, ms...)
		}
	}
	return out, nil
}

// LoadTextures gathers the texture set for the models at path: BTX0 blocks in the
// file/archive itself, the same-stem .nsbtx, every sibling .nsbtx, and — the course
// convention — the sibling "<stem>Tex.carc" archive.
func LoadTextures(path string) map[string]nitro.Texture {
	texs := map[string]nitro.Texture{}
	addFile := func(p string) {
		data, err := os.ReadFile(p)
		if err != nil {
			return
		}
		addBlob(texs, data)
	}
	// the file/archive itself (embedded BTX0s)
	addFile(path)
	stem := strings.TrimSuffix(path, filepath.Ext(path))
	// same-stem texture files and the course Tex archive
	addFile(stem + ".nsbtx")
	addFile(stem + "Tex.carc")
	// any loose sibling .nsbtx (menu model dirs)
	if ents, err := os.ReadDir(filepath.Dir(path)); err == nil {
		for _, e := range ents {
			if strings.HasSuffix(strings.ToLower(e.Name()), ".nsbtx") {
				addFile(filepath.Join(filepath.Dir(path), e.Name()))
			}
		}
	}
	return texs
}

// addBlob collects the textures of every BTX0 in data (a bare NSBTX or a NARC).
func addBlob(texs map[string]nitro.Texture, data []byte) {
	data = nds.Decompress(data)
	add := func(b []byte) {
		ts, err := nitro.DecodeNSBTX(b)
		if err != nil {
			return
		}
		for _, t := range ts {
			if _, dup := texs[t.Name]; !dup {
				texs[t.Name] = t
			}
		}
	}
	if len(data) >= 4 && string(data[:4]) == "BTX0" {
		add(data)
		return
	}
	if files, err := nds.ParseNARC(data); err == nil {
		for _, f := range files {
			if len(f) >= 4 && string(f[:4]) == "BTX0" {
				add(f)
			}
		}
	}
}
