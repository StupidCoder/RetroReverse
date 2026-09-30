package knowledge

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"unicode/utf8"

	"golang.org/x/text/unicode/norm"
)

type Media struct {
	Role   string `json:"role"`
	Size   uint64 `json:"size"`
	SHA256 string `json:"sha256"`
	MD5    string `json:"md5,omitempty"`
}
type Member struct {
	Path   string `json:"path"`
	Size   uint64 `json:"size"`
	SHA256 string `json:"sha256"`
}
type FileSet struct {
	Algorithm  string   `json:"algorithm"`
	PathPolicy string   `json:"pathPolicy"`
	Members    []Member `json:"members"`
	SHA256     string   `json:"sha256"`
}
type Release struct {
	Label   string   `json:"label"`
	Media   []Media  `json:"media,omitempty"`
	FileSet *FileSet `json:"fileSet,omitempty"`
}

func CanonicalPath(path, policy string) (string, error) {
	if policy != "exact" && policy != "ascii-insensitive" {
		return "", fmt.Errorf("unknown path policy %q", policy)
	}
	if !utf8.ValidString(path) {
		return "", fmt.Errorf("invalid UTF-8 path")
	}
	path = norm.NFC.String(strings.ReplaceAll(path, "\\", "/"))
	if len(path) >= 2 && path[1] == ':' {
		return "", fmt.Errorf("drive path rejected")
	}
	for _, r := range path {
		if r < 32 || r == 127 {
			return "", fmt.Errorf("control character in path")
		}
	}
	for _, p := range strings.Split(path, "/") {
		if p == "" || p == "." || p == ".." {
			return "", fmt.Errorf("unsafe path %q", path)
		}
	}
	if policy == "ascii-insensitive" {
		path = strings.Map(func(r rune) rune {
			if r >= 'A' && r <= 'Z' {
				return r + 32
			}
			return r
		}, path)
	}
	return path, nil
}
func SHA256(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }
func validHash(s string) bool {
	b, e := hex.DecodeString(s)
	return e == nil && len(b) == 32 && s == strings.ToLower(s)
}

// compactJSON uses literal Unicode, including U+2028/29, matching JSON.stringify.
func compactJSON(v any) ([]byte, error) {
	var b bytes.Buffer
	e := json.NewEncoder(&b)
	e.SetEscapeHTML(false)
	if err := e.Encode(v); err != nil {
		return nil, err
	}
	out := bytes.TrimSuffix(b.Bytes(), []byte("\n"))
	out = bytes.ReplaceAll(out, []byte(`\u2028`), []byte("\u2028"))
	out = bytes.ReplaceAll(out, []byte(`\u2029`), []byte("\u2029"))
	return out, nil
}
func Manifest(members []Member, policy string) ([]byte, string, error) {
	if len(members) == 0 {
		return nil, "", fmt.Errorf("empty file set")
	}
	type row struct {
		path string
		size uint64
		hash string
	}
	rows := []row{}
	seen := map[string]bool{}
	for _, m := range members {
		p, e := CanonicalPath(m.Path, policy)
		if e != nil {
			return nil, "", e
		}
		if seen[p] {
			return nil, "", fmt.Errorf("duplicate canonical path %q", p)
		}
		seen[p] = true
		if !validHash(m.SHA256) || m.Size > 9007199254740991 {
			return nil, "", fmt.Errorf("invalid member identity")
		}
		rows = append(rows, row{p, m.Size, m.SHA256})
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].path < rows[j].path })
	a := make([][]any, len(rows))
	for i, r := range rows {
		a[i] = []any{r.path, r.size, r.hash}
	}
	b, e := compactJSON(a)
	if e != nil {
		return nil, "", e
	}
	b = append([]byte("rr-media-set-v1\n"), b...)
	return b, SHA256(b), nil
}
func HashFile(path string) (uint64, string, error) {
	f, e := os.Open(path)
	if e != nil {
		return 0, "", e
	}
	defer f.Close()
	h := sha256.New()
	n, e := io.Copy(h, f)
	return uint64(n), hex.EncodeToString(h.Sum(nil)), e
}

// MatchMedia requires all declared roles, with no substitution between logical and raw identities.
func (p *Package) MatchMedia(files map[string]string) (string, error) {
	matches := []string{}
	actual := map[string]Media{}
	for role, path := range files {
		n, h, e := HashFile(path)
		if e != nil {
			return "", e
		}
		actual[role] = Media{Role: role, Size: n, SHA256: h}
	}
	for _, id := range keys(p.Releases) {
		r := p.Releases[id]
		if len(r.Media) == 0 {
			continue
		}
		match := true
		for _, m := range r.Media {
			x, ok := actual[m.Role]
			match = match && ok && x.Size == m.Size && x.SHA256 == m.SHA256
		}
		if match {
			matches = append(matches, id)
		}
	}
	if len(matches) != 1 {
		return "", fmt.Errorf("raw media matched %d releases (expected one)", len(matches))
	}
	return matches[0], nil
}

// MatchFileSet accepts extra files only outside each release's declared identity subset.
func (p *Package) MatchFileSet(members []Member) (string, error) {
	matches := []string{}
	for _, id := range keys(p.Releases) {
		r := p.Releases[id]
		if r.FileSet == nil {
			continue
		}
		s := r.FileSet
		if _, _, e := Manifest(members, s.PathPolicy); e != nil {
			return "", e
		}
		actual := map[string]Member{}
		for _, m := range members {
			k, _ := CanonicalPath(m.Path, s.PathPolicy)
			actual[k] = m
		}
		selected := []Member{}
		for _, m := range s.Members {
			k, _ := CanonicalPath(m.Path, s.PathPolicy)
			if a, ok := actual[k]; ok {
				selected = append(selected, a)
			}
		}
		if len(selected) != len(s.Members) {
			continue
		}
		_, h, e := Manifest(selected, s.PathPolicy)
		if e != nil {
			return "", e
		}
		if h == s.SHA256 {
			matches = append(matches, id)
		}
	}
	if len(matches) != 1 {
		return "", fmt.Errorf("file set matched %d releases (expected one)", len(matches))
	}
	return matches[0], nil
}
