// Package export turns Luigi's Mansion's decoded formats into web-ready
// assets: .mdl/.bin models as (skinned) GLBs, .anm/.key animations as glTF
// clips, and .scd/.sco demo cameras as baked per-frame tracks. It is the
// engine behind both cmd/lmtool (single-asset inspection) and cmd/webexport
// (the Retro-X game tree).
package export

import (
	"fmt"
	"strings"

	"retroreverse.com/games/luigis-mansion-gc/extract/lm"
	"retroreverse.com/tools/platform/gc"
)

// Source is an open disc image.
type Source struct {
	Disc *gc.Disc
}

func Open(image string) (*Source, error) {
	d, err := gc.Open(image)
	if err != nil {
		return nil, err
	}
	return &Source{Disc: d}, nil
}

func (s *Source) Close() error { return s.Disc.Close() }

// File reads a disc file, transparently undoing Yay0.
func (s *Source) File(path string) ([]byte, error) {
	for _, f := range s.Disc.FST.Entries {
		if !f.Dir && strings.EqualFold(f.Path, path) {
			b, err := s.Disc.Read(f.Offset, int(f.Size))
			if err != nil {
				return nil, err
			}
			if len(b) >= 4 && string(b[:4]) == "Yay0" {
				return lm.Yay0(b)
			}
			return b, nil
		}
	}
	return nil, fmt.Errorf("no file %q on the disc", path)
}

// Archive reads a disc file and parses the RARC archive inside it.
func (s *Source) Archive(path string) ([]lm.RARCFile, error) {
	b, err := s.File(path)
	if err != nil {
		return nil, err
	}
	return lm.RARC(b)
}

// List returns the disc paths matching prefix+suffix (case-sensitive prefix,
// as stored), sorted by the FST's own order.
func (s *Source) List(prefix, suffix string) []string {
	var out []string
	for _, f := range s.Disc.FST.Entries {
		if !f.Dir && strings.HasPrefix(f.Path, prefix) && strings.HasSuffix(f.Path, suffix) {
			out = append(out, f.Path)
		}
	}
	return out
}

// Member finds one archive member by name (case-insensitive).
func Member(files []lm.RARCFile, name string) *lm.RARCFile {
	for i := range files {
		if strings.EqualFold(files[i].Name, name) {
			return &files[i]
		}
	}
	return nil
}
