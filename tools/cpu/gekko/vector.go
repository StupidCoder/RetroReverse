package gekko

// vector.go defines the per-instruction test-vector format, shared by the generator
// (tools/cmd/gekkovec) and the test that runs the committed suite (vectors_test.go).
//
// A vector is a complete, self-contained claim about one instruction: this state, this
// word, that state. Nothing in it is derived from the interpreter — the generator
// computes the expected side from tools/cpu/gekko/ref, a second implementation that does
// not import this package — so a vector that passes is evidence, not a tautology.
//
// DontCare is the honest part. Two instructions (divw by zero, and 0x80000000 / -1) leave
// their destination register *architecturally undefined* while still defining the
// overflow flags. A suite that asserted a result for them would be asserting a fact the
// architecture declines to state; a suite that skipped them entirely would lose the flag
// check, which is real. So the register is named don't-care and the flags are asserted.

// Case is one vector.
type Case struct {
	Op uint32 `json:"op"`

	// The initial state. Only the fields an instruction reads need be present.
	//
	// Floating-point values are BIT PATTERNS, not JSON numbers. JSON cannot encode an
	// infinity or a NaN at all, and a suite whose entire purpose is bit-exactness has no
	// business round-tripping its expected results through a decimal representation. So a
	// register is two uint64s, and the comparison is on the bits.
	GPR  map[string]uint32    `json:"gpr,omitempty"`
	FPR  map[string][2]uint64 `json:"fpr,omitempty"`
	GQR  map[string]uint32    `json:"gqr,omitempty"`
	Mem  map[string]uint32    `json:"mem,omitempty"` // byte address (hex) -> byte value
	XER  uint32               `json:"xer,omitempty"`
	CR   uint32               `json:"cr,omitempty"`
	HID2 uint32               `json:"hid2,omitempty"`

	// The expected final state. Only what the instruction is expected to change.
	OutGPR map[string]uint32    `json:"out_gpr,omitempty"`
	OutFPR map[string][2]uint64 `json:"out_fpr,omitempty"`
	OutMem map[string]uint32    `json:"out_mem,omitempty"`
	OutXER uint32               `json:"out_xer,omitempty"`
	OutCR  uint32               `json:"out_cr,omitempty"`

	// CheckXER and CheckCR say whether those registers are part of the claim. They are
	// explicit because zero is a legitimate expected value, and "expected 0" and "not
	// checked" must not be the same thing.
	CheckXER bool `json:"check_xer,omitempty"`
	CheckCR  bool `json:"check_cr,omitempty"`

	// DontCare names registers the architecture leaves undefined for this case.
	DontCare []string `json:"dontcare,omitempty"`

	// Note explains a case whose point is not obvious from its numbers.
	Note string `json:"note,omitempty"`
}

// Suite is the whole committed set, keyed by mnemonic.
type Suite map[string][]Case
