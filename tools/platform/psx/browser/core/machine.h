#pragma once
#include "cpu.h"
#include "disc.h"
#include "gpu.h"
#include <cstdio>
#include <deque>
#include <memory>
struct Machine;
struct CD {
  struct Pending {
    int delay;
    u8 cause;
    std::vector<u8> response;
    bool read = false;
  };
  u8 index = 0, stat = 2, mode = 0, irqEnable = 0, irqFlags = 0;
  std::vector<u8> params, data;
  std::deque<u8> response;
  size_t dataPos = 0;
  int loc = 0, readLBA = 0;
  bool reading = false;
  std::deque<Pending> queue;
  u64 commands = 0;
  u8 read(u32 port);
  void write(Machine &, u32, u8);
  void command(Machine &, u8);
  void tick(Machine &);
  void deliver(Machine &);
  void loadData(Machine &);
  void dma(Machine &, u32, u32);
  void enqueue(int delay, u8 cause, std::vector<u8> r, bool read = false) {
    if (queue.size() >= 256)
      throw std::runtime_error("CD response queue capacity exceeded");
    queue.push_back({delay, cause, std::move(r), read});
  }
  void ack() { enqueue(1000, 3, {stat}); }
  void second() { enqueue(20000, 2, {stat}); }
  void queueRead() { enqueue(15000, 1, {stat}, true); }
};
struct Machine {
  CPU cpu;
  GTE gte;
  GPU gpu;
  CD cd;
  std::shared_ptr<Disc> disc;
  std::vector<u8> ram = std::vector<u8>(2 * 1024 * 1024), scratch = std::vector<u8>(1024);
  std::array<u32, 2048> io{};
  u32 irqStat = 0, irqMask = 0, timer = 0, dmaFlags = 0;
  u64 vblankAcc = 0, fields = 0;
  u32 isrChain = 0, isrHandler = 0;
  struct ISR {
    bool active = false;
    u32 pc = 0;
    std::array<u32, 32> r{};
    u32 hi = 0, lo = 0;
  } isr;
  u32 heapPtr = 0, heapEnd = 0, nextEvent = 0, randSeed = 0, padBuf = 0;
  bool padActive = false;
  u16 buttons = 65535;
  std::string tty;
  std::vector<std::string> diagnostics;
  u8 read(u32);
  void write(u32, u8);
  u32 ioRead(u32);
  void ioEffect(u32, u32);
  void dmaGPU(u32, u32, u32);
  u32 read16(u32 a) {
    u32 lo = read(a), hi = read(a + 1);
    return lo | (hi << 8);
  }
  u32 read32(u32 a) {
    u32 p = a & 0x1fffffff;
    if (p < 0x800000 && (p & 0x1fffff) <= 0x1ffffc)
      return le32(ram.data() + (p & 0x1fffff));
    u32 b0 = read(a), b1 = read(a + 1), b2 = read(a + 2), b3 = read(a + 3);
    return b0 | (b1 << 8) | (b2 << 16) | (b3 << 24);
  }
  void write16(u32 a, u32 v) {
    write(a, u8(v));
    write(a + 1, u8(v >> 8));
  }
  void write32(u32 a, u32 v) {
    u32 p = a & 0x1fffffff;
    if (p < 0x800000 && (p & 0x1fffff) <= 0x1ffffc) {
      for (int i = 0; i < 4; i++)
        ram[(p & 0x1fffff) + i] = u8(v >> (8 * i));
      return;
    }
    for (int i = 0; i < 4; i++)
      write(a + i, u8(v >> (8 * i)));
  }
  void irq(int bit) { irqStat |= 1u << bit; }
  void writePad() {
    if (padActive && padBuf) {
      write(padBuf, 0);
      write(padBuf + 1, 0x41);
      write(padBuf + 2, u8(buttons));
      write(padBuf + 3, u8(buttons >> 8));
    }
  }
  void boot(std::shared_ptr<Disc>, u32 handler);
  u32 run(u32 budget, bool stopField = false);
  void bios(char);
  u32 service(char, u32);
  void exception();
  void rfeTo(u32 p) {
    cpu.rfe();
    cpu.setPC(p);
  }
  bool dispatch(u32);
  void returnISR();
  std::string cstr(u32 a) {
    std::string s;
    for (unsigned i = 0; i < 4096; i++) {
      u8 c = read(a + i);
      if (!c)
        break;
      s += char(c);
    }
    return s;
  }
  void note(std::string s) {
    if (diagnostics.size() < 128 &&
        std::find(diagnostics.begin(), diagnostics.end(), s) == diagnostics.end())
      diagnostics.push_back(std::move(s));
  }
  void appendTTY(const std::string &s) {
    if (tty.size() < 65536)
      tty.append(s, 0, 65536 - tty.size());
  }
  u32 alloc(u32 size) {
    if (!heapPtr) {
      heapPtr = 0x80180000;
      heapEnd = 0x80200000;
    }
    size = (size + 3) & ~3u;
    u32 p = heapPtr;
    if (u64(p) + size > heapEnd)
      return 0;
    heapPtr += size;
    return p;
  }
  void printfGuest(u32);
};
