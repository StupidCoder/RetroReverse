// Package curation loads a game's editorial inputs from
// games/<slug>/curation/ (see STANDARDS §4.3): the hand-authored text that
// accompanies extracted data — game description, per-asset descriptions,
// per-object prose, and the technical write-up pages. The exporter merges
// these into the Retro-X output; the curation directory is the source of
// truth and site/public is never hand-edited.
package curation

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Curation struct {
	Game    GameMeta
	Assets  map[string]string     // asset id → description
	Objects map[string]ObjectInfo // object asset id → richer prose
	DocsDir string                // "" when the game has no docs pages
}

type GameMeta struct {
	Title       string    `json:"title,omitempty"`
	Platform    string    `json:"platform,omitempty"`
	Year        int       `json:"year,omitempty"`
	Description string    `json:"description,omitempty"`
	Logo        string    `json:"logo,omitempty"` // path relative to the curation dir
	Docs        []DocPage `json:"docs,omitempty"`
}

type DocPage struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	File  string `json:"file"` // relative to curation/docs/
}

type ObjectInfo struct {
	Title string `json:"title,omitempty"`
	Text  string `json:"text"`
}

// Load reads games/<slug>/curation/. A missing directory or missing files are
// not errors — extraction must work without curation — but malformed JSON is.
func Load(dir string) (*Curation, error) {
	c := &Curation{
		Assets:  map[string]string{},
		Objects: map[string]ObjectInfo{},
	}
	if st, err := os.Stat(dir); err != nil || !st.IsDir() {
		return c, nil
	}
	if err := loadInto(filepath.Join(dir, "game.json"), &c.Game); err != nil {
		return nil, err
	}
	if c.Game.Logo != "" {
		c.Game.Logo = filepath.Join(dir, c.Game.Logo)
	}
	if err := loadInto(filepath.Join(dir, "assets.json"), &c.Assets); err != nil {
		return nil, err
	}
	if err := loadInto(filepath.Join(dir, "objects.json"), &c.Objects); err != nil {
		return nil, err
	}
	if docs := filepath.Join(dir, "docs"); dirExists(docs) {
		c.DocsDir = docs
	}
	return c, nil
}

func loadInto(path string, v any) error {
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if err := json.Unmarshal(b, v); err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}
	return nil
}

func dirExists(p string) bool {
	st, err := os.Stat(p)
	return err == nil && st.IsDir()
}
