#pragma once
#include <array>


// During a short capture, retain CPU RAM writes before the blit's source is
// known. Only the discovered producer pages are promoted to the shared replay.
// The high bit separates modeled VGA planes/registers from guest RAM.
namespace rrdos {
constexpr uint32_t VIDEO = 0x80000000u, NONE = UINT32_MAX;
constexpr uint32_t VIDEO_BYTES = 0xe0340, PACKED_RAM = 0x100000;
#ifndef RR_DOS_RAW_CAP
#define RR_DOS_RAW_CAP (8 * 1024 * 1024)
#endif
constexpr size_t RAW_CAP = RR_DOS_RAW_CAP, PAGE_CAP = 4096;
struct Write {
  uint64_t clock;
  uint32_t address, value, pc, source, readBefore;
  uint8_t size;
};
struct Region { uint32_t guest, packed, size; };
struct Pixel { uint32_t source = NONE, copy = NONE, write = 0; };
inline std::vector<Write> raw;
inline std::vector<uint8_t> initialRAM, initialVideo;
inline std::vector<Region> regions;
inline std::array<Pixel, 64000> pixels;
inline uint64_t startClock = 0, lastVideoClock = 0;
inline uint32_t bursts = 0, dropped = 0, producerPixels = 0, rawCount = 0;
inline bool collecting = false, inBurst = false, timedOut = false, reveal = false;
inline uint32_t copySource = NONE;
inline std::array<uint32_t, 4> copySources{}, copyDestinations{};
inline uint32_t copyValue = 0, copyReadBefore = 0;
inline int copySize = 0;

inline void observe(uint64_t clock) {
  if (collecting && inBurst && clock - lastVideoClock >= 20000) {
    ++bursts;
    inBurst = false;
  }
}
inline void record(uint32_t address, uint8_t value, uint64_t clock, uint32_t pc,
                   bool literalCopy) {
  if (!collecting) return;
  observe(clock);
  if ((address & VIDEO) && (address & ~VIDEO) >= 0xa0000 &&
      (address & ~VIDEO) < 0xe0000) {
    inBurst = true;
    lastVideoClock = clock;
  }
  uint32_t source = literalCopy ? copySource : NONE;
  if (!raw.empty()) {
    auto &w = raw.back();
    if (w.clock == clock && w.pc == pc && w.size < 4 &&
        address == w.address + w.size && (address & ~3u) == (w.address & ~3u) &&
        ((source == NONE && w.source == NONE) ||
         (source != NONE && w.source != NONE && source == w.source + w.size && copyReadBefore == w.readBefore))) {
      w.value |= uint32_t(value) << (8 * w.size++);
      return;
    }
  }
  if (raw.size() >= RAW_CAP) { ++dropped; return; }
  raw.push_back({clock, address, value, pc, source, copyReadBefore, 1});
}
inline uint32_t packed(uint32_t guest) {
  auto it = std::upper_bound(regions.begin(), regions.end(), guest,
      [](uint32_t a, const Region &r) { return a < r.guest; });
  if (it == regions.begin()) return NONE;
  const auto &r = *--it;
  return guest - r.guest < r.size ? r.packed + guest - r.guest : NONE;
}
inline uint32_t guest(uint32_t address) {
  for (const auto &r : regions)
    if (address >= r.packed && address - r.packed < r.size)
      return r.guest + address - r.packed;
  return address;
}
} // namespace rrdos
