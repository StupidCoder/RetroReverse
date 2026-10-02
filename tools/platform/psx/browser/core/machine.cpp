#include "machine.h"
u8 Machine::read(u32 a) {
  a &= 0x1fffffff;
  if (a < 0x800000)
    return ram[a & 0x1fffff];
  if (a >= 0x1f800000 && a < 0x1f800400)
    return scratch[a - 0x1f800000];
  if (a >= 0x1f801800 && a <= 0x1f801803)
    return cd.read(a - 0x1f801800);
  if (a >= 0x1f801000 && a < 0x1f803000)
    return u8(ioRead(a & ~3u) >> ((a & 3) * 8));
  if (a >= 0x1f000000 && a < 0x1f800000)
    return 255;
  return 0;
}
void Machine::write(u32 a, u8 v) {
  if (a == 0xfffe0130)
    return;
  a &= 0x1fffffff;
  if (a < 0x800000) {
    ram[a & 0x1fffff] = v;
    return;
  }
  if (a >= 0x1f800000 && a < 0x1f800400) {
    scratch[a - 0x1f800000] = v;
    return;
  }
  if (a >= 0x1f801800 && a <= 0x1f801803) {
    cd.write(*this, a - 0x1f801800, v);
    return;
  }
  if (a >= 0x1f801000 && a < 0x1f803000) {
    u32 base = a & ~3u, shift = (a & 3) * 8;
    auto &r = io[(base - 0x1f801000) / 4];
    r = (r & ~(255u << shift)) | (u32(v) << shift);
    if ((a & 3) == 3 || ((base == 0x1f801070 || base == 0x1f801074 || base == 0x1f801114) && (a & 3) == 1))
      ioEffect(base, r);
  }
}
u32 Machine::ioRead(u32 base) {
  switch (base) {
  case 0x1f801070:
    return irqStat;
  case 0x1f801074:
    return irqMask;
  case 0x1f801814:
    return gpu.status();
  case 0x1f801810:
    return gpu.read();
  case 0x1f801110:
    if (io[0x114 / 4] & 0x100) return io[0x110 / 4] & 65535;
    [[fallthrough]];
  case 0x1f801100:
  case 0x1f801120:
    timer += 0x100;
    return timer & 65535;
  case 0x1f8010f4: {
    u32 v = (io[(base - 0x1f801000) / 4] & 0xffffff) | (dmaFlags << 24);
    if ((v & (1 << 15)) || ((v & (1 << 23)) && (dmaFlags & ((v >> 16) & 127))))
      v |= 0x80000000;
    return v;
  }
  default:
    return io[(base - 0x1f801000) / 4];
  }
}
void Machine::ioEffect(u32 base, u32 w) {
  switch (base) {
  case 0x1f801070:
    irqStat &= w;
    return;
  case 0x1f801074:
    irqMask = w;
    return;
  case 0x1f801810:
    gpu.gp0(w);
    return;
  case 0x1f801814:
    gpu.gp1(w);
    return;
  case 0x1f801114:
    io[0x110 / 4] = 0;
    return;
  case 0x1f8010f4:
    dmaFlags &= ~((w >> 24) & 127);
    return;
  }
  if (base >= 0x1f801080 && base < 0x1f801100 && (base & 15) == 8 && (w & 0x1000000)) {
    u32 ch = (base - 0x1f801080) >> 4, i = (base - 0x1f801000) / 4, madr = io[i - 2] & 0x1fffff,
        bcr = io[i - 1];
    if (ch == 2)
      dmaGPU(madr, bcr, w);
    else if (ch == 3)
      cd.dma(*this, madr, bcr);
    io[i] = w & ~0x1000000u;
    irq(3);
    dmaFlags |= 1 << ch;
  }
}
void Machine::dmaGPU(u32 addr, u32 bcr, u32 chcr) {
  if (((chcr >> 9) & 3) == 2) {
    std::vector<u32> ws;
    ws.reserve(256);
    for (int i = 0; i < 0x40000; i++) {
      addr &= 0x1fffff;
      u32 header = read32(addr), n = header >> 24;
      ws.clear();
      for (u32 j = 0; j < n; j++) {
        addr = (addr + 4) & 0x1fffff;
        ws.push_back(read32(addr));
      }
      gpu.feedNode(ws);
      u32 next = header & 0xffffff;
      if (next & 0x800000)
        return;
      addr = next;
    }
    throw std::runtime_error("GPU DMA linked list exceeded 262,144 nodes");
  }
  u32 n = bcr & 65535;
  if ((bcr >> 16) > 1)
    n *= bcr >> 16;
  if (n > 4 * 1024 * 1024)
    throw std::runtime_error("GPU DMA word budget exceeded");
  for (u32 i = 0; i < n; i++) {
    if (chcr & 1)
      gpu.gp0(read32(addr));
    else
      write32(addr, gpu.readWord());
    addr = (addr + 4) & 0x1fffff;
  }
}
void Machine::boot(std::shared_ptr<Disc> d, u32 handler) {
  auto exe = d->boot();
  if (exe.size() < 2048 || memcmp(exe.data(), "PS-X EXE", 8))
    throw std::runtime_error("Invalid PS-X EXE");
  u32 base = le32(exe.data() + 24) & 0x1fffff, size = le32(exe.data() + 28);
  if (u64(base) + size > ram.size() || u64(size) + 2048 > exe.size())
    throw std::runtime_error("EXE text overruns RAM or file");
  disc = std::move(d);
  isrHandler = handler;
  memcpy(ram.data() + base, exe.data() + 2048, size);
  cpu.setPC(le32(exe.data() + 16));
  cpu.seed(28, le32(exe.data() + 20));
  u32 sp = le32(exe.data() + 48);
  sp = sp ? sp + le32(exe.data() + 52) : 0x801fff00;
  cpu.seed(29, sp);
  cpu.seed(30, sp);
  cpu.seed(4, 1);
  cpu.cop0[12] |= 1 << 10;
}
u32 Machine::run(u32 budget, bool stopField) {
  u32 count = 0, intercepts = 0;
  bool stopPending = false;
  while (count < budget && !cpu.halted) {
    bool boundary = false;
    // Approximate NTSC HBlank ticks using the existing synthetic field clock.
    // Timer reads must not consume time: VSync(1) uses this for upload deadlines.
    if ((io[0x114 / 4] & 0x100) &&
        (vblankAcc + 1) * 263 / 250000 != vblankAcc * 263 / 250000)
      io[0x110 / 4] = (io[0x110 / 4] + 1) & 65535;
    if (++vblankAcc >= 250000) {
      vblankAcc = 0;
      irq(0);
      writePad();
      fields++;
      boundary = true;
    }
    if (stopField && boundary)
      stopPending = true;
    cd.tick(*this);
    cpu.interrupt((irqStat & irqMask) && !isr.active);
    u32 p = cpu.pc & 0x1fffffff;
    if (p == 0xa0 || p == 0xb0 || p == 0xc0) {
      bios(p == 0xa0 ? 'A' : p == 0xb0 ? 'B' : 'C');
      if (++intercepts > 100000)
        throw std::runtime_error("BIOS intercept budget exceeded");
      continue;
    }
    if (p == 0x80) {
      exception();
      if (++intercepts > 100000)
        throw std::runtime_error("Exception budget exceeded");
      continue;
    }
    if (p == 0xe0) {
      returnISR();
      continue;
    }
    if (!cpu.pc)
      throw std::runtime_error("Guest returned to address zero");
    if (stopPending)
      return count;
    cpu.step(*this);
    count++;
  }
  if (cpu.halted)
    throw std::runtime_error(cpu.error);
  return count;
}
#include "bios.inc"
#include "cd.inc"
#include "cpu.inc"
