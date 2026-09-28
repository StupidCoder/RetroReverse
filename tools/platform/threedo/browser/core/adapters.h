// Word access preserves byte watchpoints and the original I/O fallback.
inline uint8_t *wordMemory(threedo_Machine *m, uint32_t a) {
  if (a < 0x200000)
    return m->dram.p + a;
  if (a >= 0x200000 && a < 0x300000)
    return m->vram.p + a - 0x200000;
  if (a >= 0x400000 && a < 0x800000)
    return m->imem.p + a - 0x400000;
  return nullptr;
}
uint32_t arm60_CPU_read32aligned(arm60_CPU *c, uint32_t a) {
  a &= ~3u;
  auto m = c->bus;
  if (!m->OnRead)
    if (auto p = wordMemory(m, a))
      return uint32_t(p[0]) << 24 | uint32_t(p[1]) << 16 | uint32_t(p[2]) << 8 | p[3];
  return uint32_t(threedo_Machine_Read(m, a)) << 24 |
         uint32_t(threedo_Machine_Read(m, a + 1)) << 16 |
         uint32_t(threedo_Machine_Read(m, a + 2)) << 8 | threedo_Machine_Read(m, a + 3);
}
void arm60_CPU_write32(arm60_CPU *c, uint32_t a, uint32_t v) {
  a &= ~3u;
  auto m = c->bus;
  if (!m->OnWrite)
    if (auto p = wordMemory(m, a)) {
      p[0] = v >> 24;
      p[1] = v >> 16;
      p[2] = v >> 8;
      p[3] = v;
      return;
    }
  for (int i = 0; i < 4; i++)
    threedo_Machine_Write(m, a + i, v >> (24 - i * 8));
}
