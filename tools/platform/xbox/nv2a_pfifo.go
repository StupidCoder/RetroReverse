package xbox

// nv2a_pfifo.go is the NV2A's PFIFO DMA pusher: the command-stream front end that reads
// the push buffer the title builds in RAM and turns it into a sequence of (subchannel,
// method, argument) writes for the graphics engine (PGRAPH, nv2a_pgraph.go). It is the
// direct analogue of the PICA200 command processor (n3ds/gpu.go Execute): walk a buffer,
// most words latch state, a handful are triggers, and anything structurally unknown
// halts loudly rather than guessing.
//
// The pusher is driven by the pair of pointers the title programs into the channel: it
// consumes commands from DMA_GET up to DMA_PUT. On the Xbox the runtime kicks the GPU by
// writing DMA_PUT (through the USER channel alias 0xFD800040); that write is where we run
// the pusher (nvWrite calls runPusher). The DMA object the channel is bound to
// (CACHE1_DMA_INSTANCE) has base 0 and a ~128 MB limit, so the GET/PUT values are plain
// physical addresses — the pusher reads command words straight out of RAM.
//
// Command-word format (NV2A / envytools; the platform's, not any game's):
//
//	increasing methods     (word & 0xE0030003) == 0x00000000
//	non-increasing methods (word & 0xE0030003) == 0x40000000
//	   method  = word        & 0x1FFF   (byte offset within the object)
//	   subchan = (word >> 13) & 0x7
//	   count   = (word >> 18) & 0x7FF   (data words that follow)
//	old jump                (word & 0xE0000003) == 0x20000000 -> get = word & 0x1FFFFFFF
//	new jump                (word & 3) == 1                   -> get = word & 0xFFFFFFFC
//	call                    (word & 3) == 2                   -> push get, get = word & ~3
//	return                   word == 0x00020000                -> get = subroutine return
//
// A method's `count` data words follow it; for increasing methods the method offset
// advances by 4 per data word (writing consecutive registers), for non-increasing it
// stays put (streaming data into one register, e.g. the inline vertex-data FIFO).

import "fmt"

// pusherState is the DMA pusher's running decode state, kept on the Machine so a push
// that ends mid-method (the game splits a batch across two DMA_PUT writes) resumes
// cleanly on the next kick.
type pusherState struct {
	method    uint32 // current method byte-offset within the bound object
	subchan   uint32 // current subchannel (0..7)
	count     uint32 // remaining data words for the current method
	nonInc    bool   // method offset does not advance between data words
	subReturn uint32 // saved GET for a CALL (one level, like the hardware)
	subActive bool
	running   bool // guard against a PUT write re-entering the pusher
}

// maxPushWords bounds one pusher run so a corrupt jump (or a buffer that jumps to itself)
// halts instead of spinning the oracle forever. A real frame's chain is millions of words
// at most; this is a generous backstop.
const maxPushWords = 64 << 20

// runPusher consumes the push buffer from DMA_GET to DMA_PUT, dispatching each decoded
// method to the graphics engine. It advances DMA_GET as it goes (so a title polling GET
// sees the work drained) and returns when GET reaches PUT or the CPU halts.
func (m *Machine) runPusher() {
	p := &m.push
	if p.running {
		return // re-entrancy guard: a method handler must not kick the pusher
	}
	p.running = true
	defer func() { p.running = false }()

	// Texture decodes cached during the previous run may be stale: between kicks the
	// CPU owns memory and may have rewritten texture data. Within one run the GPU
	// sees a consistent snapshot, so the cache lives exactly that long.
	if len(m.pgraph.texCache) > 0 {
		m.pgraph.texCache = map[texKey]*texImage{}
	}

	if nvTrace {
		fmt.Printf("PUSH run: GET=%08X PUT=%08X\n", m.nv.dmaGet, m.nv.dmaPut)
	}
	words := 0
	for m.nv.dmaGet != m.nv.dmaPut {
		if words++; words > maxPushWords {
			m.CPU.Halt("nv2a: pusher exceeded %d words (GET=%08X PUT=%08X) — runaway push buffer",
				maxPushWords, m.nv.dmaGet, m.nv.dmaPut)
			return
		}
		word := m.read32(m.nv.dmaGet)
		if nvTrace && p.count == 0 {
			fmt.Printf("  PUSH @%08X cmd=%08X\n", m.nv.dmaGet, word)
		}
		m.nv.dmaGet += 4

		if p.count > 0 {
			// A data word for the method in progress.
			m.pgraphMethod(p.subchan, p.method, word)
			if !p.nonInc {
				p.method += 4
			}
			p.count--
			if m.CPU.Halted {
				return
			}
			continue
		}

		switch {
		case word&0xE0030003 == 0x00000000 && word != 0:
			// increasing methods (a bare 0 word is a NOP, handled by the default below)
			p.method = word & 0x1FFF
			p.subchan = (word >> 13) & 7
			p.count = (word >> 18) & 0x7FF
			p.nonInc = false
		case word&0xE0030003 == 0x40000000:
			// non-increasing methods
			p.method = word & 0x1FFF
			p.subchan = (word >> 13) & 7
			p.count = (word >> 18) & 0x7FF
			p.nonInc = true
		case word&0xE0000003 == 0x20000000:
			// old-style jump (28-bit target)
			m.nv.dmaGet = word & 0x1FFFFFFF
		case word&3 == 1:
			// new-style jump
			m.nv.dmaGet = word & 0xFFFFFFFC
		case word&3 == 2:
			// call — one subroutine level, like the hardware
			if p.subActive {
				m.CPU.Halt("nv2a: nested push-buffer CALL at GET=%08X (word %08X)", m.nv.dmaGet-4, word)
				return
			}
			p.subReturn = m.nv.dmaGet
			p.subActive = true
			m.nv.dmaGet = word & 0xFFFFFFFC
		case word == 0x00020000:
			// return
			if !p.subActive {
				m.CPU.Halt("nv2a: push-buffer RETURN with no active CALL at GET=%08X", m.nv.dmaGet-4)
				return
			}
			m.nv.dmaGet = p.subReturn
			p.subActive = false
		case word == 0:
			// NOP: a zero command word (padding / an empty ring slot). Skip.
		default:
			m.CPU.Halt("nv2a: unrecognised push-buffer command %08X at GET=%08X", word, m.nv.dmaGet-4)
			return
		}
	}
}
