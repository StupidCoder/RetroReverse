package dc

// cue.go parses the cdrdao table-of-contents format a GD-ROM rip commonly
// ships as: a .cue-suffixed file that is actually cdrdao TOC syntax, naming
// one .bin and a track list of the form
//
//	CD_ROM
//	// Track 1
//	TRACK MODE1_RAW
//	NO COPY
//	DATAFILE "game.bin" 00:08:00 // length in bytes: 1411200
//	TRACK AUDIO
//	DATAFILE "game.bin" #1058400 09:52:00
//	TRACK MODE1_RAW
//	DATAFILE "game.bin" #2295552 112:02:00
//
// Only two facts in it are trusted: each track's mode, and each DATAFILE's
// "#offset" starting byte (absent on the first track, which starts at 0). The
// MSF after the offset is the track's *length*, and on the dump this package
// was built against it overstates a truncated audio track — so lengths are
// derived from the next track's offset and the file size instead, and each
// data track's absolute LBA comes from its own first sector header (see
// disc.go), never from the TOC.

import (
	"fmt"
	"strconv"
	"strings"
)

// Track is one TOC entry, located but not yet anchored: StartLBA is filled in
// by the disc open path for raw data tracks.
type Track struct {
	Number     int
	Mode       string // "MODE1_RAW" or "AUDIO"
	FileOffset int64  // starting byte in the .bin
	Length     int64  // bytes to the next track (or EOF)
	StartLBA   int    // absolute LBA of the first sector (data tracks)
}

// IsData reports whether the track carries 2352-byte framed data sectors.
func (t Track) IsData() bool { return strings.HasPrefix(t.Mode, "MODE1") }

// parseCue reads cdrdao TOC text, returning the referenced .bin filename and
// the track list with offsets resolved. Lengths are left for the caller to
// derive against the real file size.
func parseCue(text string) (bin string, tracks []Track, err error) {
	for ln, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if i := strings.Index(line, "//"); i >= 0 {
			line = strings.TrimSpace(line[:i])
		}
		if line == "" {
			continue
		}
		f := strings.Fields(line)
		switch f[0] {
		case "TRACK":
			if len(f) < 2 {
				return "", nil, fmt.Errorf("cue line %d: TRACK without a mode", ln+1)
			}
			tracks = append(tracks, Track{Number: len(tracks) + 1, Mode: f[1], StartLBA: -1})
		case "DATAFILE", "FILE", "AUDIOFILE":
			if len(tracks) == 0 {
				return "", nil, fmt.Errorf("cue line %d: %s before any TRACK", ln+1, f[0])
			}
			name, rest, ok := quoted(line)
			if !ok {
				return "", nil, fmt.Errorf("cue line %d: no quoted filename", ln+1)
			}
			if bin == "" {
				bin = name
			} else if bin != name {
				return "", nil, fmt.Errorf("cue line %d: second data file %q (only single-file rips are handled)", ln+1, name)
			}
			t := &tracks[len(tracks)-1]
			for _, tok := range strings.Fields(rest) {
				if strings.HasPrefix(tok, "#") {
					off, err := strconv.ParseInt(tok[1:], 10, 64)
					if err != nil {
						return "", nil, fmt.Errorf("cue line %d: bad offset %q", ln+1, tok)
					}
					t.FileOffset = off
				}
				// Anything else (the MSF length) is deliberately ignored.
			}
		}
	}
	if bin == "" || len(tracks) == 0 {
		return "", nil, fmt.Errorf("cue: no tracks found")
	}
	return bin, tracks, nil
}

// quoted extracts the first double-quoted string from line, returning it and
// the remainder of the line after it.
func quoted(line string) (s, rest string, ok bool) {
	i := strings.IndexByte(line, '"')
	if i < 0 {
		return "", "", false
	}
	j := strings.IndexByte(line[i+1:], '"')
	if j < 0 {
		return "", "", false
	}
	return line[i+1 : i+1+j], line[i+2+j:], true
}
