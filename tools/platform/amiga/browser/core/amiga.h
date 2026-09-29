#pragma once
#include "evidence.h"
#include <algorithm>
#include <array>
#include <cstdint>
#include <cstdio>
#include <cstring>
#include <deque>
#include <stdexcept>
#include <string>
#include <vector>
extern "C" {
#include "../vendor/musashi/m68k.h"
}
namespace rramiga {
constexpr unsigned W = 640, H = 256, CHIP = 512 * 1024, SLOW = 512 * 1024;
struct CIA {
  std::array<uint8_t, 16> r{};
  uint16_t ta = 65535, tb = 65535, la = 65535, lb = 65535;
  uint32_t tod = 0, alarm = 0, latch = 0;
  uint8_t pending = 0, mask = 0;
  bool todLatch = false, todStop = false;
};
struct Sprite {
  uint32_t pointer = 0;
  uint16_t pos = 0, ctl = 0, a = 0, b = 0;
  bool active = false, header = false;
};
struct ColorChange { unsigned beam = 0, index = 0; uint16_t value = 0; };
struct RenderProbe;
struct Machine {
  std::array<uint8_t, CHIP> ram{};
  std::array<uint8_t, SLOW> slow{};
  std::vector<uint8_t> rom, disk;
  std::array<uint16_t, 256> reg{};
  std::array<CIA, 2> cia{};
  std::array<Sprite, 8> sprites{};
  std::array<uint32_t, W * H> draw{}, screen{};
  uint64_t cycles = 0, steps = 0, frames = 0, diskReads = 0, blits = 0,
           copperMoves = 0;
  unsigned beam = 0, line = 0, ciaClock = 0;
  int cpuDebt = 0;
  bool overlay = true, blitZero = true, blitBusy = false;
  uint64_t blitDone = 0;
  uint16_t dma = 0, intena = 0, intreq = 0, adkcon = 0;
  uint32_t copperPC = 0;
  bool copperStopped = true;
  int copperDelay = 0;
  uint32_t diskPtr = 0;
  uint16_t diskRemaining = 0, diskArm = 0;
  std::vector<uint16_t> track;
  unsigned trackIndex = 0, diskClock = 0;
  int cylinder = 0, side = 0, loadedTrack = -1;
  bool motor = false, selected = false, diskChanged = true, diskSync = false;
  uint8_t diskControl = 255;
  uint32_t buttons = 0;
  uint8_t mouseX = 0, mouseY = 0;
  bool mouseLeft = false, mouseRight = false;
  std::deque<uint8_t> keys;
  int keyboardWait = 0;
  std::array<uint32_t, 6> bplPointer{};
  bool displayY = false;
  int lastRender = -1;
  bool palettePrepared = false;
  std::array<uint16_t, 32> linePalette{};
  std::vector<ColorChange> lineColors;
  std::array<uint32_t, 4> audioPtr{}, audioCount{}, audioClock{};
  uint32_t pc = 0;
  bool logging = false;
  void reset(const std::vector<uint8_t> &firmware,
             const std::vector<uint8_t> &image);
  uint8_t read8(uint32_t a);
  uint16_t read16(uint32_t a);
  uint32_t read32(uint32_t a);
  void write8(uint32_t a, uint8_t v);
  void write16(uint32_t a, uint16_t v);
  void write32(uint32_t a, uint32_t v);
  uint16_t chip16(uint32_t a) const {
    rrmem::access(cycles,0,a&(CHIP-1),ram[a&(CHIP-1)]|uint32_t(ram[(a+1)&(CHIP-1)])<<8,2,9,pc,a);
    return uint16_t(ram[a & (CHIP - 1)]) << 8 | ram[(a + 1) & (CHIP - 1)];
  }
  void chipWrite(uint32_t a, uint16_t v) {
    a &= CHIP - 1;
    rrmem::access(cycles,0,a,v>>8,1,10,pc,a);
    rrmem::access(cycles,0,(a+1)&(CHIP-1),v&255,1,10,pc,a+1);
    captureWrite(a, swapped(v), 2, cycles, pc, hardwareEvent);
    ram[a] = v >> 8;
    ram[(a + 1) & (CHIP - 1)] = v;
  }
  uint16_t customRead(unsigned r);
  void customWrite(unsigned r, uint16_t v, bool copper = false);
  uint8_t ciaRead(unsigned n, unsigned r);
  void ciaWrite(unsigned n, unsigned r, uint8_t v);
  void updateIRQ();
  void request(unsigned mask) {
    intreq |= mask;
    updateIRQ();
  }
  void tick(unsigned cpuCycles);
  void ciaTick();
  void todTick(unsigned n);
  void copperTick();
  void diskTick();
  void prepareTrack();
  void diskPort(uint8_t value);
  void blitter();
  void renderLine(bool preview = false, uint32_t *background = nullptr,
                  uint32_t *spriteLayer = nullptr, int plane = -1,
                  RenderProbe *probe = nullptr);
  void beginPalette();
  void spriteLine();
  unsigned run(unsigned budget);
  void key(uint8_t raw, bool down);
  uint32_t color(unsigned n) const {
    unsigned c = reg[(0x180 / 2) + (n & 31)];
    return 0xff000000 | ((c >> 8) & 15) * 17 | (((c >> 4) & 15) * 17 << 8) |
           ((c & 15) * 17 << 16);
  }
  uint32_t ptr(unsigned r) const {
    return (uint32_t(reg[r / 2]) << 16 | reg[r / 2 + 1]) & 0x1ffffe;
  }
  void setPtr(unsigned r, uint32_t p) {
    reg[r / 2] = p >> 16;
    reg[r / 2 + 1] = p;
    captureWrite(registerBase + r, swapped(p >> 16), 2, cycles, pc,
                 hardwareEvent);
    captureWrite(registerBase + r + 2, swapped(p), 2, cycles, pc,
                 hardwareEvent);
  }
};
inline Machine *active = nullptr;
} // namespace rramiga
