package knowledge

import "fmt"

type PreparationAction struct {
	Run      uint64   `json:"run,omitempty"`
	Key      []uint64 `json:"key,omitempty"`
	Joystick []uint64 `json:"joystick,omitempty"`
	Play     *uint64  `json:"play,omitempty"`
	Until    *uint64  `json:"until,omitempty"`
	Budget   uint64   `json:"budget,omitempty"`
}
type PreparedStart struct {
	Layout      string   `json:"layout,omitempty"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Releases    []string `json:"releases"`
	Evidence    []string `json:"evidence"`
	Target      struct {
		Kind string `json:"kind"`
		ID   string `json:"id"`
	} `json:"target"`
	Core     string              `json:"core"`
	Firmware []string            `json:"firmware"`
	SHA256   string              `json:"sha256"`
	PC       uint64              `json:"pc"`
	Actions  []PreparationAction `json:"actions"`
}

func (p *Package) validatePreparedStarts() error {
	for id, r := range p.PreparedStarts {
		if p.Game.System.ID != "c64" {
			return fmt.Errorf("prepared start %s: only verified C64 recipes supported", id)
		}
		if e := p.evidence(r.Evidence); e != nil {
			return e
		}
		var releases []string
		if r.Target.Kind == "tour" {
			releases = p.Tours[r.Target.ID].Releases
		} else {
			releases = p.Experiments[r.Target.ID].Releases
		}
		for _, v := range r.Releases {
			if _, ok := p.Releases[v]; !ok || !contains(releases, v) {
				return fmt.Errorf("prepared start %s: missing target/release", id)
			}
		}
		total := uint64(0)
		keys := map[uint64]uint64{}
		sticks := map[uint64]uint64{}
		for _, a := range r.Actions {
			total += a.Run + a.Budget
			if len(a.Key) > 0 {
				if a.Key[1] > 1 {
					return fmt.Errorf("invalid key state")
				}
				keys[a.Key[0]] = a.Key[1]
			}
			if len(a.Joystick) > 0 {
				if a.Joystick[0] < 1 || a.Joystick[0] > 2 || a.Joystick[1] > 31 {
					return fmt.Errorf("invalid joystick")
				}
				sticks[a.Joystick[0]] = a.Joystick[1]
			}
		}
		if total > 500000000 {
			return fmt.Errorf("prepared start %s: cycle budget exceeded", id)
		}
		for _, v := range keys {
			if v != 0 {
				return fmt.Errorf("prepared start leaves a key held")
			}
		}
		for _, v := range sticks {
			if v != 0 {
				return fmt.Errorf("prepared start leaves a joystick held")
			}
		}
	}
	return nil
}
