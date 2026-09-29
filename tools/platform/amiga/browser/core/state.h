#pragma once
#include "../../../../browser/state/archive.h"
#include "inspection.h"
#include <memory>
extern "C" unsigned rr_cpu_state(unsigned *, int);
namespace rramiga {
inline void stateFields(rrstate::Archive &a, ColorChange &c) { a(c.beam,c.index,c.value); }
inline void stateFields(rrstate::Archive &a, CIA &c) {
  a(c.r, c.ta, c.tb, c.la, c.lb, c.tod, c.alarm, c.latch, c.pending, c.mask,
    c.todLatch, c.todStop);
}
inline void stateFields(rrstate::Archive &a, Sprite &s) {
  a(s.pointer, s.pos, s.ctl, s.a, s.b, s.active, s.header);
}
inline void stateFields(rrstate::Archive &a, Machine &m) {
  a(m.ram, m.slow, m.reg, m.cia, m.sprites, m.draw, m.screen, m.cycles, m.steps,
    m.frames, m.diskReads, m.blits, m.copperMoves);
  a(m.beam, m.line, m.ciaClock, m.cpuDebt, m.overlay, m.blitZero, m.blitBusy,
    m.blitDone, m.dma, m.intena, m.intreq, m.adkcon);
  a(m.copperPC, m.copperStopped, m.copperDelay, m.diskPtr, m.diskRemaining,
    m.diskArm, m.track, m.trackIndex, m.diskClock, m.cylinder, m.side,
    m.loadedTrack);
  a(m.motor, m.selected, m.diskChanged, m.diskSync, m.diskControl, m.buttons,
    m.mouseX, m.mouseY, m.mouseLeft, m.mouseRight, m.keyboardWait);
  a(m.bplPointer, m.displayY, m.lastRender, m.audioPtr, m.audioCount,
    m.audioClock, m.pc);
  std::vector<uint8_t> keys(m.keys.begin(), m.keys.end());
  a(keys);
  if (a.reading)
    m.keys = {keys.begin(), keys.end()};
}
inline std::vector<uint8_t> save() {
  rrstate::Archive a;
  a.header(10, 2);
  std::array<uint32_t, 61> cpu{};
  if (rr_cpu_state(cpu.data(), 0) != cpu.size())
    throw std::runtime_error("CPU state schema");
  a(cpu, *active);
  a(active->palettePrepared,active->linePalette,active->lineColors);
  return std::move(a.bytes);
}
inline void restore(const uint8_t *p, size_t n) {
  rrstate::Archive a(p, n);
  bool legacy=n>=12 && p[8]==1;
  a.header(10, legacy?1:2);
  std::array<uint32_t, 61> cpu{};
  auto next = std::make_unique<Machine>();
  a(cpu, *next);
  if(!legacy)a(next->palettePrepared,next->linePalette,next->lineColors);
  a.finish();
  if(next->lineColors.size()>454)throw std::runtime_error("Invalid palette history");
  unsigned previous=0;for(auto&c:next->lineColors){if(c.beam>454||c.index>=32||c.beam<previous)throw std::runtime_error("Invalid palette position");previous=c.beam;}
  if (next->beam >= 454 || next->line >= 312 || next->ciaClock >= 10 ||
      next->cylinder < 0 || next->cylinder > 79 || next->side < 0 ||
      next->side > 1 || next->track.size() > 6400 ||
      (!next->track.empty() && next->trackIndex >= next->track.size()) ||
      next->keys.size() > 1024)
    throw std::runtime_error("Invalid Amiga clock or disk state");
  if (cpu[40] > 0xffffff || cpu[44] > 4 || cpu[45] != 0 || cpu[51] > 0x700 ||
      cpu[52] > 0x700 || cpu[53] > 3)
    throw std::runtime_error("Invalid 68000 state");
  next->rom = std::move(active->rom);
  next->disk = std::move(active->disk);
  next->logging = active->logging;
  *active = std::move(*next);
  rr_cpu_state(cpu.data(), 1);
  inspect::begin();
}
} // namespace rramiga
