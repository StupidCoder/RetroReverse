#pragma once
#include "machine.h"
#include <sstream>
struct Hash {
  u32 v = 2166136261u;
  void byte(u8 b) { v = (v ^ b) * 16777619u; }
  void word(u32 w) {
    for (int i = 0; i < 4; i++)
      byte(w >> (i * 8));
  }
  template <class T> void bytes(const T &a) {
    for (auto b : a)
      byte(u8(b));
  }
  template <class T> void words(const T &a) {
    for (auto b : a)
      word(u32(b));
  }
};
inline std::string quote(const std::string &s) {
  std::string out = "\"";
  for (u8 c : s) {
    if (c == '"' || c == '\\') {
      out += '\\';
      out += char(c);
    } else if (c < 32) {
      char b[7];
      snprintf(b, sizeof b, "\\u%04x", c);
      out += b;
    } else
      out += char(c);
  }
  return out + '"';
}
inline std::string proof(Machine &m) {
  Hash ram, vram, frame, cpu, gte, scratch;
  ram.bytes(m.ram);
  scratch.bytes(m.scratch);
  for (u16 v : m.gpu.vram) {
    vram.byte(v);
    vram.byte(v >> 8);
  }
  frame.words(m.gpu.frame());
  auto &c = m.cpu;
  cpu.words(c.r);
  cpu.words(c.out);
  cpu.word(c.hi);
  cpu.word(c.lo);
  cpu.word(c.pc);
  cpu.word(c.next);
  cpu.words(c.cop0);
  cpu.word(c.cur);
  cpu.word(c.ldReg);
  cpu.word(c.ldVal);
  cpu.word(c.delay);
  cpu.word(c.pendingDelay);
  cpu.word(c.branchAddr);
  gte.words(m.gte.data);
  gte.words(m.gte.ctrl);
  std::ostringstream s;
  s << "{\"steps\":" << c.steps << ",\"pc\":" << c.pc << ",\"ram\":" << ram.v
    << ",\"scratch\":" << scratch.v << ",\"vram\":" << vram.v << ",\"frame\":" << frame.v
    << ",\"cpu\":" << cpu.v << ",\"gte\":" << gte.v << ",\"irq\":" << m.irqStat
    << ",\"mask\":" << m.irqMask << ",\"vblank\":" << m.vblankAcc
    << ",\"commands\":" << m.gpu.commands << ",\"commandHash\":" << m.gpu.commandHash
    << ",\"pad\":" << m.buttons << ",\"x\":" << m.read32(0x800801a4)
    << ",\"z\":" << m.read32(0x800801ac) << ",\"speed\":" << m.read32(0x80080234)
    << ",\"steering\":" << m.read16(0x80080208) << ",\"throttle\":" << m.read16(0x8008025c) << "}";
  return s.str();
}
