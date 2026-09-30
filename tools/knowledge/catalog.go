package knowledge

import (
	"bytes"
	"fmt"
	"html/template"
	"sort"
	"strings"
)

type catalogPackage struct {
	ID, Name, Platform, Hash string
	Revision                 uint64
	Assets                   []catalogAsset
	Functions                []catalogFunction
}
type catalogFunction struct {
	ID         string
	Definition Definition
	Location   Location
	Evidence   []string
}
type catalogAsset struct {
	ID       string
	Asset    Asset
	Sources  []string
	Evidence []string
}

// ExportCatalog is an escaped, script-free static index. Links point only at
// validated curated paths. No game media or extracted source bytes are included.
func ExportCatalog(ps []*Package) ([]byte, error) {
	ps = append([]*Package{}, ps...)
	sort.Slice(ps, func(i, j int) bool { return ps[i].ID < ps[j].ID })
	seen := map[string]bool{}
	rows := []catalogPackage{}
	for _, p := range ps {
		if seen[p.ID] {
			return nil, fmt.Errorf("duplicate package")
		}
		seen[p.ID] = true
		row := catalogPackage{ID: p.ID, Name: p.Game.Name, Platform: p.Game.System.Name, Hash: p.Hash, Revision: p.Revision}
		for _, id := range keys(p.Functions) {
			f := p.Functions[id]
			item := catalogFunction{ID: id, Definition: f, Location: p.Locations[f.Entry]}
			for _, eid := range f.Evidence {
				e := p.Evidence[eid]
				item.Evidence = append(item.Evidence, e.Status+": "+e.Description+" "+e.Limitations+" Source: "+p.Sources[e.Source].Path)
			}
			row.Functions = append(row.Functions, item)
		}
		for _, id := range keys(p.Assets) {
			a := p.Assets[id]
			item := catalogAsset{ID: id, Asset: a}
			for _, s := range a.Sources {
				var chain []string
				key := s.Location
				for depth := 0; key != "" && depth < 16; depth++ {
					l := p.Locations[key]
					label := key + " (" + l.Kind
					if l.Decoder != "" {
						label += " · " + l.Decoder + " v1"
					}
					if l.Path != "" {
						label += " · " + l.Path
					}
					label += ")"
					chain = append([]string{label}, chain...)
					key = l.Parent
					if l.Kind == "relative" {
						key = l.Base
					}
				}
				item.Sources = append(item.Sources, strings.Join(chain, " → "))
			}
			for _, id := range a.Evidence {
				e := p.Evidence[id]
				item.Evidence = append(item.Evidence, e.Status+": "+e.Description+" "+e.Limitations+" Source: "+p.Sources[e.Source].Path)
			}
			row.Assets = append(row.Assets, item)
		}
		rows = append(rows, row)
	}
	var b bytes.Buffer
	e := template.Must(template.New("catalog").Parse(catalogTemplate)).Execute(&b, rows)
	return b.Bytes(), e
}

const catalogTemplate = `<!doctype html>
<html lang="en"><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>Game knowledge · RetroReverse</title>
<style>body{max-width:960px;margin:3rem auto;padding:0 1.2rem;background:#11151a;color:#e5e9ef;font:16px/1.6 system-ui}a{color:#9ccfff}article{border-top:1px solid #526171;margin-top:2rem}code,dd{overflow-wrap:anywhere}dt{font-weight:bold}dd{margin-left:1rem}li{margin-bottom:.6rem}small{color:#bbc5d1}</style>
<a href="/">RetroReverse</a><h1>Game knowledge</h1><p>Research packages connect exact game releases to documented code, state and assets. See the <a href="https://github.com/StupidCoder/RetroReverse/blob/main/tools/browser/docs/code-m9/README.md">release capabilities and schema policy</a>. This catalog describes the packages; it has not verified any media on your computer. Resolve source bytes locally with <code>knowledgeasset</code>. Preview links open curated exports, not live game memory.</p>
{{range .}}{{$pkg := .ID}}<section id="{{.ID}}"><h2>{{.Name}}</h2><p>{{.Platform}} · revision {{.Revision}}</p><small>Package {{.ID}} · SHA-256 {{.Hash}}</small><h3>Identified functions</h3>{{range .Functions}}<article id="{{$pkg}}--function--{{.ID}}"><h4>{{.Definition.Label}}</h4><p><code>{{.ID}}</code> · {{.Definition.ISA}} · location <code>{{.Definition.Entry}}</code> ({{.Location.Kind}} {{.Location.Address}})</p><p>{{.Definition.Applicability}}</p><p>Release IDs: {{range .Definition.Releases}}{{.}} {{end}}</p>{{range .Evidence}}<p><small>{{.}}</small></p>{{end}}</article>{{else}}<p>No functions cataloged yet.</p>{{end}}<h3>Identified assets</h3>{{range .Assets}}<article id="{{$pkg}}--{{.ID}}"><h3>{{.Asset.Label}}</h3><p>{{.Asset.Description}}</p><dl><dt>Source format / export decoder</dt><dd>{{.Asset.Format}} / {{.Asset.Decoder}} v{{.Asset.Version}}</dd><dt>Exact source identity</dt><dd>{{.Asset.Size}} bytes · {{.Asset.SHA256}}</dd><dt>Release IDs</dt><dd>{{range .Asset.Releases}}{{.}} {{end}}</dd><dt>Resolution chain</dt><dd>{{range .Sources}}<p>{{.}}</p>{{end}}</dd></dl>{{if .Asset.Dependencies}}<h4>Dependencies</h4><ul>{{range .Asset.Dependencies}}<li>{{.Role}}: <a href="#{{$pkg}}--{{.Asset}}">{{.Asset}}</a></li>{{end}}</ul>{{end}}<ul>{{range .Asset.Artifacts}}<li><a href="/{{.Path}}">Curated {{.Format}} export</a> · {{.Size}} bytes<br><small>SHA-256 {{.SHA256}}</small></li>{{end}}</ul>{{range .Evidence}}<p><small>{{.}}</small></p>{{end}}</article>{{else}}<p>Code and state knowledge; no static assets cataloged yet.</p>{{end}}</section>{{end}}</html>
`
