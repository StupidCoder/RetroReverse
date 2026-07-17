package gc

// stackprof.go answers "what is the CPU doing?" on an executable with no symbol table.
//
// profile.go times the machine's subsystems and reports the Gekko as a derived remainder. When
// that remainder is the biggest bucket — and in this game's gameplay it is, at 180 ms of a
// 340 ms field — the profile has nothing further to say, and the instrument that usually takes
// over cannot help either: A PC HISTOGRAM OF THIS GAME IS A FLAT SHEET. Four fifths of the
// intro's instructions are the three addresses of the idle loop (idle.go), and once the
// fast-forward removes those, gameplay's hottest single address is 1.25% and the next forty
// are within a factor of two of it. There is no hot spot to find. There never was one.
//
// A stack is what the histogram is missing, and it costs nothing to obtain: PowerPC keeps its
// frames in a linked list in memory, so the whole call chain is readable at any instant with no
// debug information at all (see backtraceFrom, and read its comment before trusting a frame —
// the walker reported a phantom frame 0 for the life of the file). Sampling that chain turns a
// flat sheet of addresses into "8.7% of gameplay is one linear search, called from here".
//
// WHAT THIS MEASURES IS INSTRUCTIONS, NOT TIME, and that is deliberate. A wall-clock sampler
// would measure this emulator — its Go allocator, its host's scheduler, whatever else the
// machine is doing — where an instruction-paced sampler measures the GUEST, and reports the
// same numbers on a slow laptop, under -race, and on a resumed savestate. It is also
// deterministic, which a time-based sampler can never be: two runs of the same savestate give
// the same profile, so a difference between two runs is a difference in the game.
//
// The sampling period wants to be coprime with whatever the game does periodically, or the
// sampler locks onto a harmonic and reports a phase rather than a profile — hence a prime
// default rather than a round one.

import (
	"fmt"
	"sort"
	"strings"
)

// StackSample is one distinct call stack and how many samples landed in it.
type StackSample struct {
	Stack []uint32 // the leaf PC first, then its callers outward
	Count int
}

// stackProf is the sampler's state.
type stackProf struct {
	every uint64 // sample once every this many instructions; 0 = off
	depth int    // how many caller frames to keep
	next  uint64 // the instruction count at which to sample again

	stacks  map[string]*StackSample
	samples int
}

// stackProfDefaultEvery is prime so the sampler does not lock onto a periodic guest.
const stackProfDefaultEvery = 199

// SetStackProfile turns the sampler on. every is the sampling period in instructions (0 for a
// sensible default), depth is how many caller frames to keep beyond the leaf. Passing a
// negative every turns it off and discards what was collected.
//
// The cost when it is off is one compare against a field per instruction, which is the same
// price the rest of the run loop's optional work pays.
func (m *Machine) SetStackProfile(every uint64, depth int) {
	if depth <= 0 {
		depth = 8
	}
	if every == 0 {
		every = stackProfDefaultEvery
	}
	m.stack = stackProf{
		every:  every,
		depth:  depth,
		next:   m.Instrs + every,
		stacks: map[string]*StackSample{},
	}
}

// StopStackProfile turns the sampler off, leaving what it collected readable.
func (m *Machine) StopStackProfile() { m.stack.every = 0 }

// sampleStack takes one sample. It is called from the run loop when the instruction counter
// reaches the next sampling point.
//
// A SKIPPED IDLE LOOP MUST NOT BE SAMPLED, and it is not: idleSkip advances Instrs past the
// sampling point without executing anything, so the loop it skipped never appears here. That is
// the honest answer — those instructions were not run — but it means this profile is of the
// work the machine actually interpreted, not of the console's wall clock. In gameplay the two
// are the same thing, because nothing is skipped.
func (m *Machine) sampleStack(pc uint32) {
	m.stack.next = m.Instrs + m.stack.every
	m.stack.samples++

	frames := make([]uint32, 0, m.stack.depth+1)
	frames = append(frames, pc)
	bt := m.Backtrace()
	for i := 0; i < len(bt) && i < m.stack.depth; i++ {
		frames = append(frames, bt[i])
	}

	var b strings.Builder
	for i, f := range frames {
		if i > 0 {
			b.WriteByte(' ')
		}
		fmt.Fprintf(&b, "%08X", f)
	}
	k := b.String()
	if s := m.stack.stacks[k]; s != nil {
		s.Count++
		return
	}
	m.stack.stacks[k] = &StackSample{Stack: frames, Count: 1}
}

// StackProfile reports the distinct stacks sampled, most frequent first, and the total number
// of samples taken.
func (m *Machine) StackProfile() (samples []StackSample, total int) {
	for _, s := range m.stack.stacks {
		samples = append(samples, *s)
	}
	sort.Slice(samples, func(i, j int) bool {
		if samples[i].Count != samples[j].Count {
			return samples[i].Count > samples[j].Count
		}
		return samples[i].Stack[0] < samples[j].Stack[0] // stable across runs
	})
	return samples, m.stack.samples
}

// StackProfileByCaller folds the samples by the frame at the given depth — 0 being the leaf PC,
// 1 its caller, and so on — and reports each distinct address with its share.
//
// This is the view that finds structure. A leaf histogram of this game says nothing because the
// work is spread over thousands of addresses; folding at depth 1 or 2 gathers those back into
// the functions and phases that called them.
func (m *Machine) StackProfileByCaller(depth int) (samples []StackSample, total int) {
	by := map[uint32]int{}
	for _, s := range m.stack.stacks {
		if depth >= len(s.Stack) {
			continue
		}
		by[s.Stack[depth]] += s.Count
	}
	for a, n := range by {
		samples = append(samples, StackSample{Stack: []uint32{a}, Count: n})
	}
	sort.Slice(samples, func(i, j int) bool {
		if samples[i].Count != samples[j].Count {
			return samples[i].Count > samples[j].Count
		}
		return samples[i].Stack[0] < samples[j].Stack[0]
	})
	return samples, m.stack.samples
}

// StackProfileString renders the top n stacks.
func (m *Machine) StackProfileString(n int) string {
	samples, total := m.StackProfile()
	if total == 0 {
		return "  (no stack samples; is the profiler on, and did the run cover any instructions?)\n"
	}
	var b strings.Builder
	fmt.Fprintf(&b, "%d stack samples, every %d instructions — %d distinct stacks\n",
		total, m.stack.every, len(samples))
	for i, s := range samples {
		if i >= n {
			break
		}
		fmt.Fprintf(&b, "%6.2f%%  ", float64(s.Count)/float64(total)*100)
		for j, f := range s.Stack {
			if j > 0 {
				b.WriteString(" < ")
			}
			fmt.Fprintf(&b, "0x%08X", f)
		}
		b.WriteByte('\n')
	}
	return b.String()
}
