#pragma once
#include "../../../../browser/core/capture.h"
#include "../../../../browser/core/profile.h"
#include "../../../../browser/core/replay.h"
#include "../../../../browser/core/memory.h"
namespace rramiga {
constexpr uint32_t registerBase = 0x100000, frameBase = 0x110000,
                   captureSize = frameBase + 640 * 256 * 4;
inline uint32_t hardwareEvent = 0, diskEvent = 0;
inline uint16_t swapped(uint16_t v) { return (v >> 8) | (v << 8); }
inline void captureWrite(uint32_t a, uint32_t value, unsigned size,
                         uint64_t clock, uint32_t pc, uint32_t event = 0) {
  auto &t = rrcapture::trace;
  if (!t.active)
    return;
  t.source = t.palette = t.texel = t.sourceValue = t.paletteValue = 0;
  t.u = t.v = 0;
  t.sourceBefore = t.paletteBefore = t.writes.size();
  t.record(a, value, size, clock, pc, event);
}
struct EvidenceScope {
  uint32_t previous;
  EvidenceScope(uint64_t clock, uint32_t pc, std::string detail)
      : previous(hardwareEvent) {
    hardwareEvent = rrcapture::trace.event(clock, pc, std::move(detail));
  }
  ~EvidenceScope() { hardwareEvent = previous; }
};
} // namespace rramiga
