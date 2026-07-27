// Package build assembles a Retro-X game tree. Exporters decode their game's
// formats, hand the results to a Builder, and never write manifest or
// document JSON by hand — the schema structs are the single source of truth,
// and Write() validates the finished tree before it is considered done.
package build

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"retroreverse.com/tools/lib/retrox/curation"
	"retroreverse.com/tools/lib/retrox/schema"
)

type Builder struct {
	OutDir string
	man    schema.Manifest
	// documents to write at Write() time, path (relative to OutDir) → value
	docs map[string]any
	// warnings collected along the way (curation gaps etc.)
	Warnings []string
}

// New starts a game tree at outDir (…/site/public/<gameID>).
func New(outDir, gameID string) *Builder {
	return &Builder{
		OutDir: outDir,
		man: schema.Manifest{
			Header: schema.NewHeader(),
			ID:     gameID,
		},
		docs: map[string]any{},
	}
}

func (b *Builder) warnf(format string, args ...any) {
	b.Warnings = append(b.Warnings, fmt.Sprintf(format, args...))
}

// --- game metadata ---

func (b *Builder) SetTitle(t string)              { b.man.Title = t }
func (b *Builder) SetPlatform(p string)           { b.man.Platform = p }
func (b *Builder) SetYear(y int)                  { b.man.Year = y }
func (b *Builder) SetDescription(d string)        { b.man.Description = d }
func (b *Builder) SetDisplay(d schema.Display)    { b.man.Display = d }
func (b *Builder) AddDoc(page schema.DocPage)     { b.man.Docs = append(b.man.Docs, page) }

// --- assets ---

func (b *Builder) addAsset(a schema.Asset) *schema.Asset {
	b.man.Assets = append(b.man.Assets, a)
	return &b.man.Assets[len(b.man.Assets)-1]
}

// AddLevel registers a level asset and queues its document.
func (b *Builder) AddLevel(a schema.Asset, doc *schema.Level) {
	a.Category = schema.CategoryLevel
	if a.File == "" {
		a.File = "levels/" + a.ID + ".json"
	}
	doc.Header = schema.NewHeader()
	b.docs[a.File] = doc
	b.addAsset(a)
}

// AddObject registers an object asset and queues its document.
func (b *Builder) AddObject(a schema.Asset, doc *schema.Object) {
	a.Category = schema.CategoryObject
	if a.File == "" {
		a.File = "objects/" + a.ID + ".json"
	}
	doc.Header = schema.NewHeader()
	b.docs[a.File] = doc
	b.addAsset(a)
}

// AddMedia registers a music/sfx/picture/video asset whose payload the
// exporter has already written (or will write before Write()).
func (b *Builder) AddMedia(a schema.Asset) {
	if a.Category == "" || a.File == "" {
		b.warnf("media asset %q needs category and file", a.ID)
	}
	b.addAsset(a)
}

// AddSideDoc queues an auxiliary Retro-X document (cutscene script, camera
// track, collision shapes) at a path relative to the game root. The header is
// stamped when the document embeds schema.Header.
func (b *Builder) AddSideDoc(relPath string, doc any) {
	switch d := doc.(type) {
	case *schema.Script:
		d.Header = schema.NewHeader()
	case *schema.CameraTrack:
		d.Header = schema.NewHeader()
	case *schema.Shapes:
		d.Header = schema.NewHeader()
	}
	b.docs[relPath] = doc
}

// Path returns an absolute path under the game root, creating parent
// directories — exporters use it to write payloads (PNGs, GLBs, MP3s).
func (b *Builder) Path(rel ...string) (string, error) {
	p := filepath.Join(append([]string{b.OutDir}, rel...)...)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return "", err
	}
	return p, nil
}

// CreateFile creates a payload file under the game root.
func (b *Builder) CreateFile(rel ...string) (*os.File, error) {
	p, err := b.Path(rel...)
	if err != nil {
		return nil, err
	}
	return os.Create(p)
}

// --- curation ---

// ApplyCuration merges editorial inputs: game metadata (never overriding
// values the exporter set explicitly), doc pages (copied into docs/), asset
// descriptions, and object prose as fallback object-asset descriptions.
func (b *Builder) ApplyCuration(c *curation.Curation) error {
	if c == nil {
		return nil
	}
	g := c.Game
	if b.man.Title == "" {
		b.man.Title = g.Title
	}
	if b.man.Platform == "" {
		b.man.Platform = g.Platform
	}
	if b.man.Year == 0 {
		b.man.Year = g.Year
	}
	if b.man.Description == "" {
		b.man.Description = g.Description
	}
	if g.Logo != "" {
		if err := b.copyFile(g.Logo, "logo.png"); err != nil {
			return err
		}
		b.man.Logo = "logo.png"
	}
	for _, d := range g.Docs {
		if c.DocsDir == "" {
			b.warnf("curation declares doc %q but has no docs/ directory", d.ID)
			continue
		}
		rel := "docs/" + filepath.Base(d.File)
		if err := b.copyFile(filepath.Join(c.DocsDir, d.File), rel); err != nil {
			return err
		}
		b.AddDoc(schema.DocPage{ID: d.ID, Title: d.Title, File: rel})
	}

	byID := map[string]*schema.Asset{}
	for i := range b.man.Assets {
		byID[b.man.Assets[i].ID] = &b.man.Assets[i]
	}
	for id, desc := range c.Assets {
		a := byID[id]
		if a == nil {
			b.warnf("curation describes unknown asset %q", id)
			continue
		}
		a.Description = desc
	}
	for id, info := range c.Objects {
		a := byID[id]
		if a == nil {
			b.warnf("curation objects.json names unknown object %q", id)
			continue
		}
		if a.Category != schema.CategoryObject {
			b.warnf("curation objects.json entry %q is not an object asset", id)
			continue
		}
		if a.Description == "" {
			a.Description = info.Text
		}
	}
	return nil
}

func (b *Builder) copyFile(src, destRel string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := b.CreateFile(destRel)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}

// --- output ---

// Write emits every queued document plus manifest.json, updates the parent
// index.json, and validates the finished tree. Validation errors fail the
// build; warnings are returned via b.Warnings.
func (b *Builder) Write() error {
	if b.man.Title == "" {
		return fmt.Errorf("builder: the game needs a title (exporter or curation)")
	}
	for rel, doc := range b.docs {
		if err := b.writeJSON(rel, doc); err != nil {
			return err
		}
	}
	if err := b.writeJSON("manifest.json", &b.man); err != nil {
		return err
	}
	if err := b.updateIndex(); err != nil {
		return err
	}

	issues := schema.ValidateGame(os.DirFS(b.OutDir), b.man.ID)
	var errs []string
	for _, iss := range issues {
		if iss.Level == "error" {
			errs = append(errs, iss.String())
		} else {
			b.warnf("%s", iss.String())
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("builder: the exported tree is invalid:\n  %s", strings.Join(errs, "\n  "))
	}
	return nil
}

func (b *Builder) writeJSON(rel string, v any) error {
	p, err := b.Path(rel)
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(v, "", " ")
	if err != nil {
		return fmt.Errorf("%s: %w", rel, err)
	}
	return os.WriteFile(p, append(data, '\n'), 0o644)
}

// updateIndex maintains <parent>/index.json: adds this game if absent, keeps
// the list sorted for deterministic output.
func (b *Builder) updateIndex() error {
	root := filepath.Dir(b.OutDir)
	idxPath := filepath.Join(root, "index.json")
	idx := schema.Index{Header: schema.NewHeader()}
	if data, err := os.ReadFile(idxPath); err == nil {
		if err := json.Unmarshal(data, &idx); err != nil {
			return fmt.Errorf("%s: %w", idxPath, err)
		}
	}
	found := false
	for _, g := range idx.Games {
		if g == b.man.ID {
			found = true
			break
		}
	}
	if !found {
		idx.Games = append(idx.Games, b.man.ID)
	}
	sort.Strings(idx.Games)
	idx.Header = schema.NewHeader()
	data, err := json.MarshalIndent(&idx, "", " ")
	if err != nil {
		return err
	}
	return os.WriteFile(idxPath, append(data, '\n'), 0o644)
}
