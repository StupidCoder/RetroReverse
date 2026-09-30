package knowledge

import "fmt"

type ModuleSignature struct {
	Offset uint64 `json:"offset"`
	Bytes  []byte `json:"bytes"`
}
type ModuleRelocation struct {
	Offset          uint64 `json:"offset"`
	RelativeSegment uint64 `json:"relativeSegment"`
}
type Module struct {
	Label       string             `json:"label"`
	Mode        string             `json:"mode"`
	Kind        string             `json:"kind"`
	Paragraph   *uint64            `json:"paragraph"`
	Size        uint64             `json:"size"`
	Signatures  []ModuleSignature  `json:"signatures"`
	Relocations []ModuleRelocation `json:"relocations"`
}
type Buffer struct {
	When     TourCondition `json:"when"`
	Label    string        `json:"label"`
	Location string        `json:"location"`
	Width    uint64        `json:"width"`
	Height   uint64        `json:"height"`
	Stride   uint64        `json:"stride"`
	Palette  string        `json:"palette"`
	Releases []string      `json:"releases"`
	Evidence []string      `json:"evidence"`
}

func (p *Package) validateModules() error {
	for _, r := range p.Releases {
		if r.Executable != "" {
			found := false
			if r.FileSet != nil {
				for _, m := range r.FileSet.Members {
					if m.Path == r.Executable {
						found = true
					}
				}
			}
			if !found || p.Game.System.ID != "dos" {
				return fmt.Errorf("executable must be a pinned DOS file-set member")
			}
		}
	}

	for id, m := range p.Modules {
		if p.Game.System.ID != "dos" {
			return fmt.Errorf("modules require DOS")
		}
		if (m.Kind == "mz") != (m.Paragraph != nil) {
			return fmt.Errorf("module %s: only MZ modules declare relative paragraphs", id)
		}
		for _, s := range m.Signatures {
			if s.Offset+uint64(len(s.Bytes)) > m.Size {
				return fmt.Errorf("signature exceeds module")
			}
		}
		for _, r := range m.Relocations {
			if r.Offset+2 > m.Size {
				return fmt.Errorf("relocation exceeds module")
			}
		}
	}
	for _, b := range p.Buffers {
		if e := p.checkCondition(b.When, b.Releases, 68157440); e != nil {
			return e
		}
		if p.Game.System.ID != "dos" || b.Stride < b.Width || b.Stride*b.Height > 65536 {
			return fmt.Errorf("invalid buffer bounds/platform")
		}
		if e := p.evidence(b.Evidence); e != nil {
			return e
		}
		for _, r := range b.Releases {
			s, e := p.Resolve(b.Location, r)
			if e != nil {
				return e
			}
			if s.Limit-s.Offset < b.Stride*b.Height {
				return fmt.Errorf("buffer exceeds location")
			}
		}
	}
	return nil
}
