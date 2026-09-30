package knowledge

import "fmt"

// M4 is deliberately data-only: current paused state, sequential stops, C64
// physical RAM predicates and mapped RAM PC targets. No expressions or scripts.
type TourCondition struct {
	PC      *uint64         `json:"pc,omitempty"`
	Address *uint64         `json:"address,omitempty"`
	Value   uint64          `json:"value,omitempty"`
	Mask    *uint64         `json:"mask,omitempty"`
	All     []TourCondition `json:"all,omitempty"`
	Any     []TourCondition `json:"any,omitempty"`
}
type TourGuard struct {
	Address uint64 `json:"address"`
	Bytes   []byte `json:"bytes"`
}
type TourCapture struct {
	Address uint64 `json:"address"`
	Length  uint64 `json:"length"`
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
		if p.Game.System.ID != "c64" {
			return fmt.Errorf("tour %s: only C64 is implemented", id)
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
			if g.Address+uint64(len(g.Bytes)) > 65536 {
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
				if r.Address+r.Length > 65536 || total > 16384 {
					return fmt.Errorf("tour capture exceeds RAM/budget")
				}
				for _, previous := range s.Capture[:i] {
					if r.Address < previous.Address+previous.Length && previous.Address < r.Address+r.Length {
						return fmt.Errorf("overlapping tour capture ranges")
					}
				}
			}
		}
		for _, c := range conditions {
			nodes := 0
			if e := validateCondition(c, 0, &nodes); e != nil {
				return e
			}
		}
	}
	return nil
}
