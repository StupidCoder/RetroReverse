package xbox

// kernel_data.go handles xboxkrnl's *data* exports. Not every import is a function: the
// kernel also exports variables and structures by ordinal — the OBJECT_TYPE blocks the
// object manager compares against, the debugger-present flags, the running tick/time
// counters, and the console's version/hardware/key structures. A title imports these
// exactly like a function (an IAT slot), but then *dereferences* the slot to read the
// data rather than CALLing through it. Pointing such a slot at a code trap makes the
// first `MOV EAX,[slot]; MOV EAX,[EAX]` fault — which is precisely how these were found
// (an out-of-range read inside the trap region).
//
// So patchThunks routes each import two ways: a data-export ordinal gets a populated
// block in the kernel band and its slot points there; every other ordinal gets the code
// trap. The set of data exports is standard kernel ABI. Values are the plausible ones a
// retail console presents; a title mostly passes the OBJECT_TYPE pointers straight back
// to Nt*/Ob* calls and reads the version/flags for feature checks, so exact contents
// matter far less than the pointer being valid and the struct being coherent.

// dataExportSize reports whether an ordinal needs an explicit, populated data block
// (rather than the default: a code trap whose dereference reads back zero — see
// machine.go's Read). Classifying an ordinal as data here is only warranted when a zero
// value would send the boot down the wrong path AND the ordinal's identity is certain;
// guessing wrong is worse than a zero (a *function* wrongly pointed at a data block is
// CALLed straight into that block and executes garbage — the bug that motivated the
// zero-on-deref default). So this set stays empty until a concrete need with a verified
// ordinal appears; everything else is a code trap that reads back zero when used as
// data.
func dataExportSize(ord uint16) (int, bool) {
	return 0, false
}

// initDataExport populates a data-export block with plausible retail values.
func (m *Machine) initDataExport(ord uint16, addr uint32) {
	switch ord {
	case 84: // KdDebuggerEnabled -> FALSE
		m.Write(addr, 0)
	case 85: // KdDebuggerNotPresent -> TRUE
		m.Write(addr, 1)
	case 152: // KeTimeIncrement -> ~1 ms in 100 ns units (0x2710)
		m.write32(addr, 0x2710)
	case 323:
		// XboxKrnlVersion is read as a { USHORT Major, Minor, Build, Qfe } struct;
		// present the retail 1.0.5838.1 kernel so an explicit version check passes.
		m.write16(addr+0, 1)    // Major
		m.write16(addr+2, 0)    // Minor
		m.write16(addr+4, 5838) // Build
		m.write16(addr+6, 1)    // Qfe
	case 320, 321, 322, 324, 325:
		// Hardware-info / key blocks read as flags and key material. Leave them zeroed
		// rather than inventing feature bits — a fabricated flag steers title logic down
		// paths the real value would not (the HLE-fiction-becomes-evidence trap). The
		// boot's config query for setting 0x11 runs on the zero-flags path.
	case 344: // IdexChannelObject — a nonzero, otherwise-zeroed object is enough
	}
	if ord == 151 {
		m.tickCountAddr = addr // KeTickCount: keep it live (schedTick updates it)
	}
	if ord == 149 {
		m.systemTimeAddr = addr
	}
}
