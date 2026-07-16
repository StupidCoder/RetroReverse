package vu

// state.go is the vector unit's snapshot surface: every register and every stage of the
// in-flight flag pipeline, in a form gob can carry. The program and data memories are NOT
// here — the VIF owns those slices (it fills them with MPG and UNPACK) and hands them to
// New, so the layer that snapshots the VIF snapshots the memories and rebuilds the VU over
// them.
//
// The flag pipeline is the part that would be easy to drop and expensive to have dropped:
// microcode reads the flags of the upper instruction four pairs back, so a VU restored
// with an empty pipeline runs its next four pairs against flags that are not the ones the
// scheduler counted on — a cull branch that goes the other way, once, right after every
// resume.

// FlagVals is one pipeline stage's worth of flag state, exported for the snapshot.
type FlagVals struct {
	Mac    uint16
	Status uint16
	Clip   uint32
}

// State is a VU's registers and pipeline, complete.
type State struct {
	VF  [32][4]uint32
	VI  [16]uint16
	ACC [4]float32
	Q   float32
	P   float32
	I   float32
	R   uint32

	PC uint32

	Mac    uint16
	Status uint16
	Clip   uint32

	FlagPipe           [4]FlagVals
	VisMac, VisStatus  uint16
	VisClip            uint32

	Top, ITop uint16
	CMSAR0    uint16
	Steps     uint64
}

// Snapshot captures the unit.
func (v *VU) Snapshot() State {
	s := State{
		VF: v.VF, VI: v.VI, ACC: v.ACC,
		Q: v.Q, P: v.P, I: v.I, R: v.R,
		PC:  v.PC,
		Mac: v.Mac, Status: v.Status, Clip: v.Clip,
		VisMac: v.visMac, VisStatus: v.visStatus, VisClip: v.visClip,
		Top: v.Top, ITop: v.ITop, CMSAR0: v.CMSAR0, Steps: v.Steps,
	}
	for i, f := range v.flagPipe {
		s.FlagPipe[i] = FlagVals{Mac: f.mac, Status: f.status, Clip: f.clip}
	}
	return s
}

// Restore puts a snapshot back. The memories and the host callbacks (XGKick, StartVU1)
// are the caller's wiring and are left as they are.
func (v *VU) Restore(s State) {
	v.VF, v.VI, v.ACC = s.VF, s.VI, s.ACC
	v.Q, v.P, v.I, v.R = s.Q, s.P, s.I, s.R
	v.PC = s.PC
	v.Mac, v.Status, v.Clip = s.Mac, s.Status, s.Clip
	v.visMac, v.visStatus, v.visClip = s.VisMac, s.VisStatus, s.VisClip
	v.Top, v.ITop, v.CMSAR0, v.Steps = s.Top, s.ITop, s.CMSAR0, s.Steps
	for i, f := range s.FlagPipe {
		v.flagPipe[i] = flagVals{mac: f.Mac, status: f.Status, clip: f.Clip}
	}
}
