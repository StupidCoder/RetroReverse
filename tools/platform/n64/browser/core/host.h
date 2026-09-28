#pragma once
#include "generated.cpp"
inline void destroy(n64_Machine *m) {
  if (!m)
    return;
  delete m->CPU;
  delete m->RSP;
  delete m;
}
inline uint32_t crc32(const uint8_t *p, size_t n) {
  uint32_t c = ~0u;
  for (size_t i = 0; i < n; i++) {
    c ^= p[i];
    for (int j = 0; j < 8; j++)
      c = (c >> 1) ^ (0xedb88320u & (0u - (c & 1)));
  }
  return ~c;
}
inline n64_Machine *boot(Slice<uint8_t> rom) {
  if (rom.n < 4096)
    throw std::runtime_error("ROM too short");
  if (rom[0] == 0x37 && rom[1] == 0x80) {
    for (int64_t i = 0; i + 1 < rom.n; i += 2)
      std::swap(rom[i], rom[i + 1]);
  } else if (rom[0] == 0x40 && rom[1] == 0x12) {
    for (int64_t i = 0; i + 3 < rom.n; i += 4) {
      std::swap(rom[i], rom[i + 3]);
      std::swap(rom[i + 1], rom[i + 2]);
    }
  }
  if (be_Uint32(rom) != 0x80371240)
    throw std::runtime_error("Unrecognized ROM byte order");
  const auto cic = crc32(rom.p + 0x40, 0xfc0);
  if (cic != 0x90bb6cb5 && cic != 0xb531bde6)
    throw std::runtime_error("Unsupported IPL3 boot code: this core currently supports CIC 6102 and libdragon IPL3");
  const bool pal = std::string("DFIPSUXY").find(char(rom[0x3e])) != std::string::npos;
  n64_ROM r{rom, {}};
  auto m = n64_NewMachine(&r);
  gcopy(m->DMEM, sub(rom, 0, 4096));
  m->ri[0x0c] = 0x14;
  n64_Machine_writePhys32(m, 0x318, 4 * 1024 * 1024);
  auto c = m->CPU;
  r4300_CPU_Reset(c);
  c->COP0[12] = 0x34000000;
  for (auto [reg, val] : std::initializer_list<std::pair<int, uint64_t>>{{3, 0},
                                                                         {11, 0xa4000040},
                                                                         {19, 0},
                                                                         {20, pal ? 0u : 1u},
                                                                         {21, 0},
                                                                         {22, 0x3f},
                                                                         {23, 0},
                                                                         {29, 0xa4001ff0},
                                                                         {31, 0xa4001550}})
    r4300_CPU_SetReg(c, reg, val);
  r4300_CPU_SetPC(c, 0xa4000040);
  m->noSpin = true;
  return m;
}
inline uint32_t hashBytes(const uint8_t *p, size_t n) {
  uint32_t h = 2166136261;
  for (size_t i = 0; i < n; i++)
    h = (h ^ p[i]) * 16777619;
  return h;
}
struct Hasher {
  uint32_t h = 2166136261;
  void add(uint64_t v) {
    for (int i = 0; i < 8; i++) {
      h = (h ^ uint8_t(v)) * 16777619;
      v >>= 8;
    }
  }
};
inline std::string proof(n64_Machine *m, uint64_t steps) {
  Hasher c, r;
  for (auto v : m->CPU->R)
    c.add(v);
  c.add(m->CPU->HI);
  c.add(m->CPU->LO);
  c.add(m->CPU->PC);
  c.add(m->CPU->nextPC);
  for (auto v : m->CPU->COP0)
    c.add(v);
  for (auto v : m->CPU->FGR)
    c.add(v);
  c.add(m->CPU->FCR31);
  c.add(m->CPU->Steps);
  if (m->RSP) {
    for (auto v : m->RSP->R)
      r.add(v);
    for (auto v : m->RSP->V)
      for (auto x : v)
        r.add(x);
    for (auto v : m->RSP->Acc)
      r.add(v);
    r.add(m->RSP->VCO);
    r.add(m->RSP->VCC);
    r.add(m->RSP->VCE);
    r.add(m->RSP->PC);
    r.add(m->RSP->Steps);
  }
  std::ostringstream s;
  s << "{\"steps\":" << steps << ",\"pc\":" << uint32_t(m->CPU->PC) << ",\"cpu\":" << c.h
    << ",\"rsp\":" << r.h << ",\"ram\":" << hashBytes(m->RDRAM.p, m->RDRAM.n)
    << ",\"dmem\":" << hashBytes(m->DMEM.p, 4096) << ",\"imem\":" << hashBytes(m->IMEM.p, 4096)
    << ",\"rspSteps\":" << m->rspSteps << ",\"rdpWords\":" << m->rdpWords
    << ",\"origin\":" << n64_Machine_Origin(m) << ",\"width\":" << n64_Machine_Width(m)
    << ",\"polls\":" << m->ContPolls << "}";
  return s.str();
}
inline uint32_t height(n64_Machine *m) {
  auto vs = m->vi.Regs[0x28], ys = m->vi.Regs[0x34] & 4095;
  auto start = (vs >> 16) & 1023, end = vs & 1023;
  if (end <= start || ys == 0)
    return 240;
  auto h = (end - start) / 2 * ys >> 10;
  return h == 0 || h > 512 ? 240 : h;
}
inline std::vector<uint8_t> frame(n64_Machine *m) {
  auto w = n64_Machine_Width(m), h = height(m), o = n64_Machine_Origin(m),
       t = n64_Machine_PixelType(m);
  if (w == 0 || w > 1024)
    w = 320;
  std::vector<uint8_t> p(w * h * 4);
  for (uint32_t i = 0; i < w * h; i++) {
    p[i * 4 + 3] = 255;
    if (!o)
      continue;
    uint32_t a = o + i * (t == 2 ? 2 : 4);
    if (a + 3 >= m->RDRAM.n)
      continue;
    if (t == 2) {
      auto c = n64_fromRGBA16(uint16_t(m->RDRAM[a]) << 8 | m->RDRAM[a + 1]);
      p[i * 4] = c.R;
      p[i * 4 + 1] = c.G;
      p[i * 4 + 2] = c.B;
    } else if (t == 3) {
      p[i * 4] = m->RDRAM[a];
      p[i * 4 + 1] = m->RDRAM[a + 1];
      p[i * 4 + 2] = m->RDRAM[a + 2];
    }
  }
  return p;
}
