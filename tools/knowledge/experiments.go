package knowledge

import "fmt"

type ExperimentEdit struct {
	Kind        string `json:"kind"`
	Address     uint64 `json:"address"`
	Before      []byte `json:"before"`
	After       []byte `json:"after"`
	Explanation string `json:"explanation"`
}
type ExperimentStage struct {
	Title      string           `json:"title"`
	Require    []TourCondition  `json:"require"`
	Edits      []ExperimentEdit `json:"edits"`
	Cycles     uint64           `json:"cycles"`
	PC         uint64           `json:"pc"`
	Budget     uint64           `json:"budget"`
	Assertions []TourCondition  `json:"assertions"`
}
type Experiment struct {
	Title        string            `json:"title"`
	Description  string            `json:"description"`
	Requirements string            `json:"requirements"`
	Releases     []string          `json:"releases"`
	Evidence     []string          `json:"evidence"`
	Start        TourCondition     `json:"start"`
	Guards       []TourGuard       `json:"guards"`
	Setup        []ExperimentStage `json:"setup"`
	Invariants   []TourCondition   `json:"invariants"`
	Patch        []ExperimentEdit  `json:"patch"`
	Duration     uint64            `json:"duration"`
	WallBudget   uint64            `json:"wallBudget"`
	ProbePC      uint64            `json:"probePC"`
	Inputs       []struct {
		Cycle   uint64 `json:"cycle"`
		Buttons uint64 `json:"buttons"`
	} `json:"inputs"`
	Watches []struct {
		ID      string `json:"id"`
		Label   string `json:"label"`
		Address uint64 `json:"address"`
	} `json:"watches"`
	Observations []struct {
		ID    string        `json:"id"`
		Label string        `json:"label"`
		When  TourCondition `json:"when"`
	} `json:"observations"`
	Interpretation string `json:"interpretation"`
}

func validateEdits(edits []ExperimentEdit) error {
	seen := map[uint64]bool{}
	for _, e := range edits {
		if len(e.Before) != len(e.After) || e.Address+uint64(len(e.Before)) > 65536 {
			return fmt.Errorf("edit length/range mismatch")
		}
		for i := range e.Before {
			a := e.Address + uint64(i)
			if seen[a] {
				return fmt.Errorf("overlapping edits")
			}
			seen[a] = true
		}
	}
	if len(seen) > 256 {
		return fmt.Errorf("edit transaction exceeds 256 bytes")
	}
	return nil
}
func (p *Package) validateExperiments() error {
	for id, e := range p.Experiments {
		if p.Game.System.ID != "c64" {
			return fmt.Errorf("experiment %s: only C64 checked RAM/code edits supported", id)
		}
		if err := p.evidence(e.Evidence); err != nil {
			return err
		}
		for _, r := range e.Releases {
			if _, ok := p.Releases[r]; !ok {
				return fmt.Errorf("unknown experiment release")
			}
		}
		conditions := append([]TourCondition{e.Start}, e.Invariants...)
		budget := uint64(0)
		for _, s := range e.Setup {
			if err := validateEdits(s.Edits); err != nil {
				return err
			}
			budget += s.Cycles + s.Budget
			conditions = append(conditions, s.Require...)
			conditions = append(conditions, s.Assertions...)
		}
		if budget > 15000000 {
			return fmt.Errorf("preparation exceeds cycle budget")
		}
		if err := validateEdits(e.Patch); err != nil {
			return err
		}
		for _, g := range e.Guards {
			if g.Location != "" || g.Address < 2 || g.Address+uint64(len(g.Bytes)) > 65536 {
				return fmt.Errorf("invalid experiment guard")
			}
		}
		seen := map[string]bool{}
		for _, w := range e.Watches {
			if seen[w.ID] {
				return fmt.Errorf("duplicate watch")
			}
			seen[w.ID] = true
		}
		seen = map[string]bool{}
		for _, o := range e.Observations {
			if seen[o.ID] {
				return fmt.Errorf("duplicate observation")
			}
			seen[o.ID] = true
			conditions = append(conditions, o.When)
		}
		for _, c := range conditions {
			if err := p.checkCondition(c, e.Releases, 65536); err != nil {
				return err
			}
		}
		for i, v := range e.Inputs {
			if v.Cycle >= e.Duration || i == 0 && v.Cycle != 0 || i > 0 && v.Cycle <= e.Inputs[i-1].Cycle {
				return fmt.Errorf("input schedule must start at zero, increase and precede duration")
			}
		}
	}
	return nil
}
