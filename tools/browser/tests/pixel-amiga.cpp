#include "../../platform/amiga/browser/core/api.cpp"
#include <cassert>
#include <iostream>
using namespace rramiga;
static void nextFrame() {
  auto f = machine->frames;
  while (machine->frames == f)
    assert(rr_run(10000) > 0);
}
static void resetTest() {
  std::vector<uint8_t> rom(262144), disk(901120);
  auto put = [&](unsigned a, uint16_t v) {
    rom[a] = v >> 8;
    rom[a + 1] = v;
  };
  put(0, 7);
  put(2, 0);
  put(4, 0xfc);
  put(6, 8); // Reset SP and PC.
  for (auto [a, v] :
       std::initializer_list<std::pair<unsigned, uint16_t>>{{8, 0x23fc},
                                                            {10, 0x1234},
                                                            {12, 0x5678},
                                                            {14, 0},
                                                            {16, 0x1000},
                                                            {18, 0x60fe}})
    put(a, v);
  std::memcpy(rr_firmware(rom.size()), rom.data(), rom.size());
  std::memcpy(rr_input(disk.size()), disk.data(), disk.size());
  assert(rr_init(disk.size()));
}
int main() {
  resetTest();
  assert(rr_run(100) > 0);
  auto &m = *machine;
  assert(m.chip16(0x1000) == 0x1234 && m.chip16(0x1002) == 0x5678);
  // ROM overlay is controlled through CIAA's output and direction bits.
  assert(m.read16(0) == 7);
  m.ciaWrite(0, 2, 3);
  m.ciaWrite(0, 0, 2);
  m.write16(0, 0xa55a);
  assert(m.read16(0) == 0xa55a);
  assert(m.read16(0x80000) == 0xa55a);
  m.write32(0xc00000, 0x87654321);
  assert(m.read32(0xc00000) == 0x87654321);
  // Timer A one-shot, latched ICR and binary TOD latch.
  m.ciaWrite(0, 4, 1);
  m.ciaWrite(0, 5, 0);
  m.ciaWrite(0, 13, 0x81);
  m.ciaWrite(0, 14, 0x19);
  m.ciaTick();
  m.ciaTick();
  assert(m.ciaRead(0, 13) == 0x81);
  assert(!(m.cia[0].r[14] & 1));
  assert(m.ciaRead(0, 13) == 0);
  m.cia[1].tod = 0x12ffff;
  assert(m.ciaRead(1, 10) == 0x12);
  m.todTick(1);
  assert(m.ciaRead(1, 9) == 0xff);
  assert(m.ciaRead(1, 8) == 0xff);
  assert(m.ciaRead(1, 9) == 0);
  m.cia[0].mask = 8;
  m.key(0x20, true);
  m.keyboardWait = 0;
  m.ciaTick();
  assert(m.cia[0].r[12] == uint8_t(~0x40));
  assert(m.ciaRead(0, 13) & 8);
  m.buttons = 1 | 8 | 16;
  assert(m.customRead(0xc) == 0x103);
  assert(!(m.ciaRead(0, 0) & 128));
  m.mouseX = 3;
  m.mouseY = 2;
  m.customWrite(0x36, 0x8490);
  assert(m.customRead(0xa) == 0x8693);
  // Read every MFM-encoded data word back from a patterned, media-free ADF.
  for (unsigned i = 0; i < m.disk.size(); i++)
    m.disk[i] = (i * 37 + i / 512) & 255;
  m.cylinder = 17;
  m.side = 1;
  m.prepareTrack();
  auto longAt = [&](unsigned at) {
    return uint32_t(m.track[at]) << 16 | m.track[at + 1];
  };
  for (unsigned s = 0; s < 11; s++) {
    unsigned base = 166 + s * 544;
    assert(m.track[base + 2] == 0x4489);
    uint32_t checksum = 0;
    auto header = ((longAt(base + 4) & 0x55555555) << 1) |
                  (longAt(base + 6) & 0x55555555);
    assert(header == (0xff230000 | (s << 8) | (11 - s)));
    for (unsigned i = 0; i < 128; i++) {
      auto odd = longAt(base + 32 + i * 2) & 0x55555555,
           even = longAt(base + 288 + i * 2) & 0x55555555;
      checksum ^= odd ^ even;
      auto at = (35 * 11 + s) * 512 + i * 4;
      uint32_t expected = uint32_t(m.disk[at]) << 24 |
                          uint32_t(m.disk[at + 1]) << 16 |
                          uint32_t(m.disk[at + 2]) << 8 | m.disk[at + 3];
      assert((odd << 1 | even) == expected);
    }
    assert((((longAt(base + 28) & 0x55555555) << 1) |
            (longAt(base + 30) & 0x55555555)) == checksum);
  }
  m.motor = m.selected = true;
  m.dma = 0x210;
  m.adkcon = 0;
  m.customWrite(0x20, 0);
  m.customWrite(0x22, 0x4000);
  m.customWrite(0x24, 0x8002);
  assert(m.diskRemaining == 0);
  m.customWrite(0x24, 0x8002);
  assert(m.diskRemaining == 2);
  for (int i = 0; i < 224; i++)
    m.diskTick();
  assert(!m.diskRemaining && m.diskReads == 1 && (m.intreq & 2));
  // All Boolean minterms, descending copy and cross-word source shift.
  m.motor = false;
  m.reg[0x44 / 2] = m.reg[0x46 / 2] = 0xffff;
  for (unsigned op = 0; op < 256; op++) {
    m.setPtr(0x54, 0x5000);
    m.reg[0x40 / 2] = 0x100 | op;
    m.reg[0x42 / 2] = 0;
    m.reg[0x74 / 2] = 0xf0f0;
    m.reg[0x72 / 2] = 0xcccc;
    m.reg[0x70 / 2] = 0xaaaa;
    m.reg[0x58 / 2] = 65;
    m.blitter();
    uint16_t expected = 0;
    for (unsigned bit = 0; bit < 16; bit++) {
      unsigned index = (((0xf0f0 >> bit) & 1) << 2) |
                       (((0xcccc >> bit) & 1) << 1) | ((0xaaaa >> bit) & 1);
      expected |= ((op >> index) & 1) << bit;
    }
    assert(m.chip16(0x5000) == expected);
  }
  m.chipWrite(0x6000, 0x1234);
  m.chipWrite(0x6002, 0x5678);
  m.reg[0x40 / 2] = 0x49f0;
  m.reg[0x42 / 2] = 0;
  m.reg[0x58 / 2] = 66;
  m.setPtr(0x50, 0x6000);
  m.setPtr(0x54, 0x7000);
  m.blitter();
  assert(m.chip16(0x7000) == 0x123 && m.chip16(0x7002) == 0x4567);
  m.reg[0x40 / 2] = 0x9f0;
  m.reg[0x42 / 2] = 2;
  m.setPtr(0x50, 0x6002);
  m.setPtr(0x54, 0x7002);
  m.blitter();
  assert(m.chip16(0x7000) == 0x1234 && m.chip16(0x7002) == 0x5678);
  // Copper WAIT must not pass the post-255 wait before the beam wraps.
  m.dma = 0x280;
  m.line = 255;
  m.beam = 444;
  m.copperPC = 0x8000;
  m.copperStopped = false;
  m.copperDelay = 0;
  uint16_t copper[] = {0xffdf, 0xfffe, 0x0501, 0xfffe,
                       0x0180, 0x0f00, 0xffff, 0xfffe};
  for (unsigned i = 0; i < 8; i++)
    m.chipWrite(0x8000 + i * 2, copper[i]);
  m.reg[0x180 / 2] = 0;
  m.tick(32);
  assert(m.line == 256 && m.reg[0x180 / 2] == 0);
  m.tick(454 * 5 + 32);
  assert(m.reg[0x180 / 2] == 0xf00);
  // Pixel-source writes, palette history, rejected sprites and reversible
  // replay.
  resetTest();
  auto &v = *machine;
  v.overlay = false;
  v.dma = 0x320;
  v.line = 44;
  v.reg[0x8e / 2] = 0x2c81;
  v.reg[0x90 / 2] = 0x2cc1;
  v.reg[0x92 / 2] = 0x38;
  v.reg[0x94 / 2] = 0xd0;
  v.reg[0x100 / 2] = 0x1000;
  v.bplPointer[0] = 0x10000;
  v.reg[0x182 / 2] = 0xf00;
  v.reg[0x1a2 / 2] = 0x00f;
  v.reg[0x104 / 2] = 1;
  v.sprites[0] = {.pointer = 0x20004,
                  .pos = 0x2c40,
                  .ctl = 1,
                  .a = 0xffff,
                  .b = 0,
                  .active = true,
                  .header = true}; // x=0
  assert(rr_capture_begin());
  v.chipWrite(0x10000, 0xffff);
  v.chipWrite(0x20000, 0xffff);
  v.chipWrite(0x20002, 0);
  v.renderLine();
  assert(v.draw[0] == 0xffff0000);
  v.reg[0x104 / 2] = 0;
  captureWrite(registerBase + 0x104, 0, 2, 0, 0);
  v.chipWrite(v.bplPointer[0], 0xffff);
  v.line = 45;
  v.renderLine();
  assert(v.draw[W] == 0xff0000ff);
  v.screen = v.draw;
  assert(rr_capture_end());
  assert(!rrcapture::trace.overflow);
  assert(rrcapture::trace.shadow == rrcapture::trace.final);
  assert(std::string(rr_pixel(0, 0)).find("Sprite scanline") !=
         std::string::npos);
  rr_replay_begin();
  while (!rr_replay_seek(rrreplay::replay.steps.size())) {
  }
  assert(std::memcmp(rr_frame(), rr_replay_frame(), W * H * 4) == 0);
  while (!rr_replay_seek(1)) {
  }
  assert(std::memcmp(rr_frame(), rr_replay_frame(), W * H * 4) != 0);
  // Portable state restores exactly, including CPU prefetch and device phase.
  resetTest();
  nextFrame();
  auto n = rr_state_save();
  assert(n);
  auto saved = stateOutput;
  for (int i = 0; i < 5; i++)
    nextFrame();
  auto proof = std::string(rr_proof());
  std::memcpy(rr_state_input(n), saved.data(), n);
  assert(rr_state_load(n));
  for (int i = 0; i < 5; i++)
    nextFrame();
  assert(std::string(rr_proof()) == proof);
  rr_state_input(12);
  assert(!rr_state_load(12));
  assert(std::string(rr_proof()) == proof);
  std::cout << "Amiga 68000, CIA, MFM, floppy DMA, blitter, Copper wrap, pixel "
               "sources, replay and portable states passed\n";
}
