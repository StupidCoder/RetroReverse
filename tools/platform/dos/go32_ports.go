package dos

// I/O-port model for the go32 protected-mode machine. A DJGPP game talks to the
// PC hardware the same way a real-mode one does — through IN/OUT — so the flat
// machine needs the same handful of device models the real-mode Machine has. The
// load-bearing one for Quake is the 8254 PIT: Sys_FloatTime reads timer counter 0
// (latched via port 0x43, read lo-then-hi from 0x40) together with the BIOS tick
// dword at 0040:006C, and reconstructs wall-clock seconds as
//
//	time = (bios_tick*65536 + (65536 - counter)) / 1193180
//
// so both the counter and the BIOS tick must advance in lockstep, or the game's
// clock stands still and it re-renders frame 0 forever. We derive elapsed PIT
// ticks from the CPU's instruction count — the oracle has no real wall clock, and
// tying game time to instructions retired keeps a run deterministic and replayable.

// pitBiosTickAddr is the BIOS timer-tick counter (a dword at 0040:006C); the game
// reads its low word as the coarse part of the clock.
const pitBiosTickAddr = 0x46C

// pitInstrsPerTick maps retired instructions to 8254 ticks (the timer runs at
// 1.193180 MHz). It sets how fast game time flows relative to emulated work; the
// exact value is not physical — it only needs to advance time smoothly enough that
// the demo plays and frames differ. ~1.19 M ticks/s over this ratio puts a frame's
// worth of rendering at a small fraction of a second, so the game progresses at a
// sane pace instead of skipping.
const pitInstrsPerTick = 32

type pitState struct {
	reload    uint16 // counter-0 reload latch (0 means the full 65536)
	writeHi   bool   // next OUT 0x40 writes the high reload byte
	latched   uint16 // value captured by a latch command
	haveLatch bool   // a latch is pending to be read out
	readHi    bool   // next IN 0x40 returns the high counter byte
}

// pitTotalTicks is the number of 8254 ticks that have elapsed, derived from retired
// instructions. counter 0 and the BIOS tick are both slices of this one clock.
func (p *PM) pitTotalTicks() uint64 { return p.CPU.Steps / pitInstrsPerTick }

// pitCounter returns the live counter-0 value: a down-counter from reload to 1 that
// reloads at the top, exactly what the game samples to interpolate within a tick.
func (p *PM) pitCounter() uint16 {
	reload := uint32(p.pit.reload)
	if reload == 0 {
		reload = 0x10000
	}
	e := uint32(p.pitTotalTicks() % uint64(reload))
	return uint16((reload - e) & 0xFFFF)
}

// pitSyncBiosTick writes the coarse tick count to 0040:006C so a direct memory read
// of it stays consistent with the counter the game latches in the same breath.
func (p *PM) pitSyncBiosTick() { p.w32(pitBiosTickAddr, uint32(p.pitTotalTicks()/0x10000)) }

func (p *PM) portIn(port uint16, size int) uint32 {
	switch port {
	case 0x40: // PIT counter 0 — latched value, low byte then high byte
		p.pitSyncBiosTick()
		v := p.pit.latched
		if !p.pit.haveLatch {
			v = p.pitCounter()
		}
		var b byte
		if p.pit.readHi {
			b = byte(v >> 8)
			p.pit.readHi = false
			p.pit.haveLatch = false // both halves consumed; go live again
		} else {
			b = byte(v)
			p.pit.readHi = true
		}
		return uint32(b)
	case 0x3DA, 0x3BA: // VGA input status #1: toggle retrace so wait-loops progress
		p.retrace = !p.retrace
		if p.retrace {
			return 0x09 // display-disabled | vertical-retrace
		}
		return 0x00
	case 0x60: // keyboard data: nothing pending
		return 0
	case 0x64: // keyboard status: input/output buffers empty (ready)
		return 0
	case 0x20, 0x21, 0xA0, 0xA1: // PIC
		return 0
	default:
		return widthMask8(size) // absent device reads as all-ones
	}
}

func (p *PM) portOut(port uint16, size int, v uint32) {
	b := byte(v)
	switch port {
	case 0x43: // PIT control word
		if b>>6 == 0 { // counter 0 selected
			if (b>>4)&3 == 0 { // access field 00 = counter-latch command
				p.pit.latched = p.pitCounter()
				p.pit.haveLatch = true
				p.pit.readHi = false
			} else { // mode / access programming (Quake writes 0x34: RW lo/hi, mode 2)
				p.pit.writeHi = false
				p.pit.readHi = false
			}
		}
		p.pitSyncBiosTick()
	case 0x40: // PIT counter 0 reload, low byte then high byte
		if p.pit.writeHi {
			p.pit.reload = (p.pit.reload & 0x00FF) | uint16(b)<<8
			p.pit.writeHi = false
		} else {
			p.pit.reload = (p.pit.reload & 0xFF00) | uint16(b)
			p.pit.writeHi = true
		}
	case 0x3C8: // VGA DAC write index (register = b, cursor at b*3)
		p.dacIndex = int(b) * 3
	case 0x3C9: // VGA DAC data: R,G,B 6-bit components, auto-advancing
		p.Pal[p.dacIndex%768] = b
		p.dacIndex++
	}
	// Everything else (PIC EOI, VGA sequencer/CRTC, DMA) is accepted and discarded.
}
