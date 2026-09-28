#include "amiga.h"
namespace rramiga {
static int ack(int) { return M68K_INT_ACK_AUTOVECTOR; }
static void instruction(unsigned pc) {
  active->pc = pc;
  active->steps++;
}
static void cpuReset() { active->overlay = true; }
void Machine::reset(const std::vector<uint8_t> &firmware,
                    const std::vector<uint8_t> &image) {
  if (firmware.size() != 262144 && firmware.size() != 524288)
    throw std::runtime_error("Kickstart must be 256 or 512 KiB");
  if (image.size() != 901120)
    throw std::runtime_error(
        "This Amiga core accepts standard 880 KiB ADF disks");
  *this = Machine{};
  rom = firmware;
  disk = image;
  active = this;
  draw.fill(0xff000000);
  screen = draw;
  cia[0].r[0] = cia[1].r[0] = cia[0].r[1] = cia[1].r[1] = 255;
  reg[0x7e / 2] = 0x4489;
  reg[0x44 / 2] = reg[0x46 / 2] = 65535;
  m68k_init();
  m68k_set_cpu_type(M68K_CPU_TYPE_68000);
  m68k_set_int_ack_callback(ack);
  m68k_set_instr_hook_callback(instruction);
  m68k_set_reset_instr_callback(cpuReset);
  m68k_pulse_reset();
}
uint8_t Machine::read8(uint32_t a) {
  a &= 0xffffff;
  if (a < 0x200000) {
    if (overlay && a < 0x80000)
      return rom[a % rom.size()];
    return ram[a & (CHIP - 1)];
  }
  if (a >= 0xf80000)
    return rom[(a - 0xf80000) % rom.size()];
  if (a >= 0xc00000 && a < 0xc00000 + SLOW)
    return slow[a - 0xc00000];
  if ((a & 0xfff001) == 0xbfe001)
    return ciaRead(0, (a >> 8) & 15);
  if ((a & 0xfff001) == 0xbfd000)
    return ciaRead(1, (a >> 8) & 15);
  if ((a & 0xff0000) == 0xdf0000) {
    auto v = customRead(a & 0x1fe);
    return a & 1 ? v : v >> 8;
  }
  return 255;
}
uint16_t Machine::read16(uint32_t a) {
  if ((a & 0xff0000) == 0xdf0000)
    return customRead(a & 0x1fe);
  return uint16_t(read8(a)) << 8 | read8(a + 1);
}
uint32_t Machine::read32(uint32_t a) {
  return uint32_t(read16(a)) << 16 | read16(a + 2);
}
void Machine::write8(uint32_t a, uint8_t v) {
  a &= 0xffffff;
  if (a < 0x200000) {
    captureWrite(a & (CHIP - 1), v, 1, cycles, pc);
    ram[a & (CHIP - 1)] = v;
    return;
  }
  if (a >= 0xc00000 && a < 0xc00000 + SLOW) {
    captureWrite(CHIP + a - 0xc00000, v, 1, cycles, pc);
    slow[a - 0xc00000] = v;
    return;
  }
  if ((a & 0xfff001) == 0xbfe001) {
    ciaWrite(0, (a >> 8) & 15, v);
    return;
  }
  if ((a & 0xfff001) == 0xbfd000) {
    ciaWrite(1, (a >> 8) & 15, v);
    return;
  }
  if ((a & 0xff0000) == 0xdf0000)
    customWrite(a & 0x1fe, uint16_t(v) * 0x101);
}
void Machine::write16(uint32_t a, uint16_t v) {
  if ((a & 0xff0000) == 0xdf0000) {
    customWrite(a & 0x1fe, v);
    return;
  }
  write8(a, v >> 8);
  write8(a + 1, v);
}
void Machine::write32(uint32_t a, uint32_t v) {
  write16(a, v >> 16);
  write16(a + 2, v);
}
void Machine::updateIRQ() {
  unsigned pending = (intena & 0x4000) ? intena & intreq : 0;
  unsigned level = 0;
  if (pending & 0x0007)
    level = 1;
  if (pending & 0x0008)
    level = 2;
  if (pending & 0x0070)
    level = 3;
  if (pending & 0x0780)
    level = 4;
  if (pending & 0x1800)
    level = 5;
  if (pending & 0x2000)
    level = 6;
  m68k_set_irq(level);
}
uint16_t Machine::customRead(unsigned r) {
  switch (r) {
  case 2:
    return dma | (blitBusy ? 0x4000 : 0) | (blitZero ? 0x2000 : 0);
  case 4:
    return (line >> 8) & 1;
  case 6:
    return (line << 8) | (beam >> 1);
  case 0xa:
    if (logging)
      std::fprintf(stderr, "JOY0 frame=%llu PC=%06x value=%02x%02x\n",
                   (unsigned long long)frames, pc, mouseY, mouseX);
    return uint16_t(mouseY) << 8 | mouseX;
  case 0xc: {
    bool u = buttons & 1, d = buttons & 2, l = buttons & 4, r = buttons & 8;
    return (l << 9) | ((u ^ l) << 8) | (r << 1) | (d ^ r);
  }
  case 0xe:
    return 0;
  case 0x10:
    return adkcon;
  case 0x12:
  case 0x14:
    return 0xffff;
  case 0x16:
    return 0xffff ^ (mouseRight ? 0x400 : 0);
  case 0x18:
    return 0x3000;
  case 0x1a:
    return (motor && selected ? 0x8000 : 0) | (diskRemaining ? 0x4000 : 0) |
           (track.empty() ? 0 : track[trackIndex % track.size()] & 255) |
           (diskSync ? 0x1000 : 0);
  case 0x1c:
    return intena;
  case 0x1e:
    return intreq;
  case 0x7c:
    return 0xffff;
  default:
    return 0xffff;
  }
}
static void setclear(uint16_t &d, uint16_t v) {
  if (v & 0x8000)
    d |= v & 0x7fff;
  else
    d &= ~v;
}
void Machine::customWrite(unsigned r, uint16_t v, bool copper) {
  if (r >= 0x200)
    return;
  if (logging && !copper && (r == 0x24 || r == 0x96 || r == 0x9a))
    std::fprintf(stderr, "%llu PC=%06x custom %03x=%04x\n",
                 (unsigned long long)frames, pc, r, v);
  if (logging && r >= 0x140 && r < 0x148)
    std::fprintf(stderr, "SPR0 %03x=%04x PC=%06x\n", r, v, pc);
  switch (r) {
  case 0x96:
    setclear(dma, v);
    dma &= 0x7ff;
    break;
  case 0x9a:
    setclear(intena, v);
    intena &= 0x7fff;
    updateIRQ();
    break;
  case 0x9c:
    setclear(intreq, v);
    intreq &= 0x7fff;
    updateIRQ();
    break;
  case 0x9e:
    setclear(adkcon, v);
    break;
  case 0x36:
    mouseX = (mouseX & 3) | (v & 0xfc);
    mouseY = (mouseY & 3) | ((v >> 8) & 0xfc);
    break;
  case 0x88:
  case 0x8a:
    copperPC = ptr(r == 0x88 ? 0x80 : 0x84);
    copperStopped = false;
    copperDelay = 0;
    break;
  case 0x24:
    if (!(v & 0x8000)) {
      diskRemaining = 0;
      diskArm = 0;
    } else if (diskArm & 0x8000) {
      diskRemaining = v & 0x3fff;
      diskPtr = ptr(0x20);
      diskClock = 0;
      diskSync = !(adkcon & 0x400);
      diskArm = 0;
      if (v & 0x4000)
        throw std::runtime_error("ADF writes are not implemented");
    } else
      diskArm = v;
    break;
  default:
    reg[r / 2] = v;
    captureWrite(registerBase + r, swapped(v), 2, cycles, pc, hardwareEvent);
    if (r == 0x58)
      blitter();
    if (r >= 0xe0 && r <= 0xf6)
      bplPointer[(r - 0xe0) / 4] = ptr(0xe0 + ((r - 0xe0) / 4) * 4);
    if (r >= 0x120 && r < 0x140) {
      unsigned n = (r - 0x120) / 4;
      sprites[n].pointer = ptr(0x120 + n * 4);
      sprites[n].active = sprites[n].header = false;
    }
    break;
  }
}
void Machine::tick(unsigned n) {
  cycles += n;
  ciaClock += n;
  while (ciaClock >= 10) {
    ciaClock -= 10;
    ciaTick();
  }
  for (unsigned i = 0; i < n; i += 2) {
    beam += 2;
    copperTick();
    diskTick();
    for (unsigned ch = 0; ch < 4; ch++)
      if ((dma & (0x200 | (1 << ch))) == (0x200 | (1 << ch))) {
        if (audioClock[ch] > 2)
          audioClock[ch] -= 2;
        else {
          audioClock[ch] = std::max<unsigned>(4, reg[(0xa6 + ch * 16) / 2]) * 4;
          if (!audioCount[ch]) {
            audioPtr[ch] = ptr(0xa0 + ch * 16);
            audioCount[ch] = reg[(0xa4 + ch * 16) / 2];
            if (!audioCount[ch])
              audioCount[ch] = 65536;
            request(0x80 << ch);
          }
          audioPtr[ch] += 2;
          audioCount[ch]--;
        }
      }
    if (blitBusy && cycles >= blitDone) {
      blitBusy = false;
      request(0x40);
    }
    if (beam >= 454) {
      beam -= 454;
      renderLine();
      line++;
      todTick(1);
      spriteLine();
      if (line == 312) {
        line = 0;
        frames++;
        screen = draw;
        displayY = false;
        lastRender = -1;
        copperPC = ptr(0x80);
        copperStopped = false;
        request(0x20);
        todTick(0);
      }
    }
  }
}
unsigned Machine::run(unsigned budget) {
  rrprof::Scope measured(0, "68000 and chipset scheduling");
  unsigned spent = 0;
  auto start = frames;
  while (spent < budget && frames == start) {
    int n = m68k_execute(8);
    if (n <= 0)
      n = 8;
    tick(n);
    spent += n;
  }
  return spent;
}
} // namespace rramiga
extern "C" {
unsigned m68k_read_memory_8(unsigned a) { return rramiga::active->read8(a); }
unsigned m68k_read_memory_16(unsigned a) { return rramiga::active->read16(a); }
unsigned m68k_read_memory_32(unsigned a) { return rramiga::active->read32(a); }
void m68k_write_memory_8(unsigned a, unsigned v) {
  rramiga::active->write8(a, v);
}
void m68k_write_memory_16(unsigned a, unsigned v) {
  rramiga::active->write16(a, v);
}
void m68k_write_memory_32(unsigned a, unsigned v) {
  rramiga::active->write32(a, v);
}
unsigned m68k_read_disassembler_8(unsigned a) { return m68k_read_memory_8(a); }
unsigned m68k_read_disassembler_16(unsigned a) {
  return m68k_read_memory_16(a);
}
unsigned m68k_read_disassembler_32(unsigned a) {
  return m68k_read_memory_32(a);
}
}
