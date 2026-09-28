r4300_CPU *r4300_NewCPU(n64_Machine *m) {
  auto c = new r4300_CPU{};
  c->bus = m;
  c->fetch = [m](uint32_t a) { return n64_Machine_Fetch32(m, a); };
  r4300_CPU_Reset(c);
  return c;
}
n64_Machine *n64_NewMachine(n64_ROM *rom) {
  auto m = new n64_Machine{};
  m->RDRAM = Slice<uint8_t>::make(4 * 1024 * 1024);
  m->DMEM = Slice<uint8_t>::make(4096);
  m->IMEM = Slice<uint8_t>::make(4096);
  m->PIF = Slice<uint8_t>::make(64);
  m->ROM = rom->Data;
  m->EEPROM = Slice<uint8_t>::make(512);
  m->CPU = r4300_NewCPU(m);
  n64_pi_init(&m->pi);
  n64_vi_init(&m->vi);
  n64_ai_init(&m->ai);
  n64_si_init(&m->si);
  m->sp[n64_spStatus] = n64_spStatusHalt;
  m->Controllers[0].Present = true;
  return m;
}
// Avoid constructing/ref-counting temporary slice views for every CPU word fetch.
inline uint32_t rawBE(const uint8_t *p) {
  return uint32_t(p[0]) << 24 | uint32_t(p[1]) << 16 | uint32_t(p[2]) << 8 | p[3];
}
uint32_t n64_Machine_Fetch32(n64_Machine *m, uint32_t a) {
  a &= ~3u;
  if (a < uint32_t(m->RDRAM.n))
    return rawBE(m->RDRAM.p + a);
  if (a >= 0x04000000 && a < 0x04002000)
    return rawBE((a < 0x04001000 ? m->DMEM.p : m->IMEM.p) + (a & 4095));
  auto [b, off] = n64_Machine_backing(m, a);
  if (b)
    return be_Uint32(sub(b, off, b.n));
  return n64_Machine_ioRead(m, a);
}
uint32_t n64_Machine_Read32(n64_Machine *m, uint32_t a) {
  a &= ~3u;
  auto v = n64_Machine_Fetch32(m, a);
  if (!m->hookMuted && m->OnRead && a >= m->RWatchLo && a < m->RWatchHi)
    m->OnRead(a, v, n64_Machine_pc(m));
  return v;
}
void n64_Machine_Write32(n64_Machine *m, uint32_t a, uint32_t v) {
  a &= ~3u;
  if (!m->hookMuted && m->OnWrite && a >= m->WatchLo && a < m->WatchHi)
    m->OnWrite(a, v, n64_Machine_pc(m));
  if (a < uint32_t(m->RDRAM.n)) {
    auto p = m->RDRAM.p + a;
    p[0] = v >> 24;
    p[1] = v >> 16;
    p[2] = v >> 8;
    p[3] = v;
    return;
  }
  auto [b, off] = n64_Machine_backing(m, a);
  if (b) {
    if (a >= n64_cartBase && a < n64_cartEnd) {
      n64_Machine_note(m, "write 0x%08X to the cartridge at 0x%08X (ignored)", v, a);
      return;
    }
    be_PutUint32(sub(b, off, b.n), v);
    return;
  }
  n64_Machine_ioWrite(m, a, v);
}
