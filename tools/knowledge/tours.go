package knowledge

import (
	"fmt"
	"strings"
)

// Tours are data-only: current paused state, sequential stops, bounded RAM and
// architecture-specific predicates. No expressions or executable scripts.
type TourCondition struct {
	Mode     string          `json:"mode,omitempty"`
	Register string          `json:"register,omitempty"`
	Location string          `json:"location,omitempty"`
	PC       *uint64         `json:"pc,omitempty"`
	Address  *uint64         `json:"address,omitempty"`
	Value    uint64          `json:"value,omitempty"`
	Mask     *uint64         `json:"mask,omitempty"`
	All      []TourCondition `json:"all,omitempty"`
	Any      []TourCondition `json:"any,omitempty"`
}
type TourGuard struct {
	Location string `json:"location,omitempty"`
	Address  uint64 `json:"address"`
	Bytes    []byte `json:"bytes"`
}
type TourCapture struct {
	Location string `json:"location,omitempty"`
	Address  uint64 `json:"address"`
	Length   uint64 `json:"length"`
}
type TourStop struct {
	ID          string          `json:"id"`
	Title       string          `json:"title"`
	Explanation string          `json:"explanation"`
	Until       TourCondition   `json:"until"`
	Assertions  []TourCondition `json:"assertions"`
	CycleBudget uint64          `json:"cycleBudget"`
	WallBudget  uint64          `json:"wallBudget"`
	HitCount    uint64          `json:"hitCount"`
	Capture     []TourCapture   `json:"capture"`
	EventBudget uint64          `json:"eventBudget"`
	Layout      struct {
		Address   uint64 `json:"address"`
		GameState bool   `json:"gameState"`
	} `json:"layout"`
}
type Tour struct {
	Iteration *struct {
		Counter uint64 `json:"counter"`
	} `json:"iteration,omitempty"`
	Title        string        `json:"title"`
	Description  string        `json:"description"`
	Releases     []string      `json:"releases"`
	Evidence     []string      `json:"evidence"`
	StartPolicy  string        `json:"startPolicy"`
	Requirements string        `json:"requirements"`
	Guards       []TourGuard   `json:"guards"`
	Start        TourCondition `json:"start"`
	Stops        []TourStop    `json:"stops"`
}

func validateCondition(c TourCondition, depth int, nodes *int) error {
	*nodes++
	if depth > 8 || *nodes > 64 {
		return fmt.Errorf("tour predicate exceeds depth/node budget")
	}
	if c.Mask != nil && c.Value & ^*c.Mask != 0 {
		return fmt.Errorf("tour predicate value exceeds mask")
	}
	for _, children := range [][]TourCondition{c.All, c.Any} {
		for _, x := range children {
			if e := validateCondition(x, depth+1, nodes); e != nil {
				return e
			}
		}
	}
	return nil
}
func (p *Package) validateTours() error {
	for id, t := range p.Tours {
		limit := uint64(65536)
		if p.Game.System.ID == "dos" {
			limit = 68157440
		}
		if p.Game.System.ID == "3do" {
			limit = 0x300000
		}
		if t.Iteration != nil && p.Game.System.ID != "3do" {
			return fmt.Errorf("iteration binding requires 3DO")
		}
		if p.Game.System.ID != "c64" && p.Game.System.ID != "dos" && p.Game.System.ID != "3do" {
			return fmt.Errorf("tour %s: unsupported tour platform", id)
		}
		if e := p.evidence(t.Evidence); e != nil {
			return e
		}
		for _, r := range t.Releases {
			if _, ok := p.Releases[r]; !ok {
				return fmt.Errorf("tour %s: unknown release %s", id, r)
			}
		}
		for _, g := range t.Guards {
			if g.Address+uint64(len(g.Bytes)) > limit {
				return fmt.Errorf("tour guard exceeds RAM")
			}
		}
		conditions := []TourCondition{t.Start}
		seen := map[string]bool{}
		for _, s := range t.Stops {
			if seen[s.ID] {
				return fmt.Errorf("duplicate tour stop %s", s.ID)
			}
			seen[s.ID] = true
			conditions = append(conditions, s.Until)
			conditions = append(conditions, s.Assertions...)
			total := uint64(0)
			for i, r := range s.Capture {
				total += r.Length
				if r.Address+r.Length > limit || total > 16384 {
					return fmt.Errorf("tour capture exceeds RAM/budget")
				}
				for _, previous := range s.Capture[:i] {
					if r.Location == "" && previous.Location == "" && r.Address < previous.Address+previous.Length && previous.Address < r.Address+r.Length {
						return fmt.Errorf("overlapping tour capture ranges")
					}
				}
			}
		}
		for _, g := range t.Guards {
			if g.Location != "" {
				if p.Game.System.ID != "dos" {
					return fmt.Errorf("symbolic tour guards require DOS")
				}
				for _, r := range t.Releases {
					s, e := p.Resolve(g.Location, r)
					if e != nil {
						return e
					}
					if uint64(len(g.Bytes)) > s.Limit-s.Offset {
						return fmt.Errorf("guard exceeds location")
					}
				}
			}
		}
		for _, s := range t.Stops {
			for _, c := range s.Capture {
				if c.Location != "" {
					if p.Game.System.ID != "dos" {
						return fmt.Errorf("symbolic capture ranges require DOS")
					}
					for _, r := range t.Releases {
						loc, e := p.Resolve(c.Location, r)
						if e != nil {
							return e
						}
						if c.Length > loc.Limit-loc.Offset {
							return fmt.Errorf("capture exceeds location")
						}
					}
				}
			}
		}
		for _, c := range conditions {
			if e := p.checkCondition(c, t.Releases, limit); e != nil {
				return e
			}
			nodes := 0
			if e := validateCondition(c, 0, &nodes); e != nil {
				return e
			}
		}
	}
	return nil
}

func (p *Package) checkCondition(c TourCondition, releases []string, limit uint64) error {
	nodes := 0
	if e := validateCondition(c, 0, &nodes); e != nil {
		return e
	}
	if p.Game.System.ID != "dos" && p.Game.System.ID != "3do" && (c.Location != "" || c.Register != "" || c.Mode != "") {
		return fmt.Errorf("DOS condition on another platform")
	}
	if p.Game.System.ID == "3do" && (c.Location != "" || c.Mode != "" && c.Mode != "arm32" || c.Register != "" && c.Register != "CPSR" && !strings.HasPrefix(c.Register, "R")) {
		return fmt.Errorf("unsupported ARM condition")
	}
	if p.Game.System.ID == "dos" && (c.Mode == "arm32" || strings.HasPrefix(c.Register, "R") || c.Register == "CPSR") {
		return fmt.Errorf("ARM condition on DOS")
	}

	if c.Location != "" {
		for _, r := range releases {
			if _, e := p.Resolve(c.Location, r); e != nil {
				return e
			}
		}
	}
	if c.PC != nil && *c.PC >= limit || c.Address != nil && *c.Address >= limit {
		return fmt.Errorf("tour address out of bounds")
	}
	for _, xs := range [][]TourCondition{c.All, c.Any} {
		for _, x := range xs {
			if e := p.checkCondition(x, releases, limit); e != nil {
				return e
			}
		}
	}
	return nil
}
