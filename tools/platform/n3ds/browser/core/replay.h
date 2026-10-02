#pragma once
#include "../../../../browser/core/replay.h"
namespace rr3ds {
// Project the final display copies back onto their tiled render targets. A
// display transfer is one replay step; showing only its destination hides every
// preceding draw. Freeze each source at the copy so later buffer reuse cannot
// leak into the captured image.
inline std::vector<uint32_t> displayCopies;
inline void prepareReplay() {
  auto &t = rrcapture::trace;
  displayCopies.assign(400 * 480, UINT32_MAX);
  std::unordered_map<uint32_t, uint32_t> visible;
  for (int y = 0; y < 480; ++y) for (int x = 0; x < 400; ++x) {
    auto a = pixelAddress(x, y);
    if (a == UINT32_MAX) continue;
    for (uint32_t b = 0; b < screens[y / 240].bpp; ++b)
      visible[a + b] = y * 400 + x;
  }
  for (uint32_t i = 0; i < t.writes.size(); ++i) {
    const auto &w = t.writes[i];
    if (!(w.flags & 1)) continue;
    for (uint32_t b = 0; b < w.size; ++b) {
      auto it = visible.find(w.address + b);
      if (it == visible.end()) continue;
      auto pixel = it->second;
      bool copy = w.event && w.event <= t.events.size() &&
        t.events[w.event - 1].detail.find("\"kind\":\"GX display transfer\"") != std::string::npos &&
        w.address == pixelAddress(pixel % 400, pixel / 400) &&
        w.size == screens[pixel / (400 * 240)].bpp &&
        uint64_t(w.source) + 4 <= t.initial.size();
      displayCopies[pixel] = copy ? i : UINT32_MAX;
    }
  }
}
inline std::vector<uint8_t> replayDisplay() {
  auto &r = rrreplay::replay;
  auto &t = rrcapture::trace;
  auto out = display(r.memory);
  for (uint32_t i = 0; i < displayCopies.size(); ++i) {
    auto id = displayCopies[i];
    if (id == UINT32_MAX) continue;
    const auto &w = t.writes[id];
    // Once the transfer occurs, the ordinary destination includes its format
    // conversion and any subsequent writes exactly.
    if (r.writeCursor > id) continue;
    auto v = r.writeCursor >= w.sourceBefore ? w.sourceValue : t.value(r.memory, w.source, 4);
    uint32_t red = v >> 24, green = (v >> 16) & 255, blue = (v >> 8) & 255;
    auto format = screens[i / (400 * 240)].format;
    if (format == 2 || format == 3) {
      green = format == 2 ? ((green >> 2) << 2 | (green >> 2) >> 4) : ((green >> 3) << 3 | (green >> 3) >> 2);
      // Match the display decoder's bit replication (rather than rescaling).
      red = (v >> 27) << 3 | (v >> 27) >> 2;
      blue = ((v >> 11) & 31) << 3 | ((v >> 11) & 31) >> 2;
    } else if (format == 4) {
      red = (red >> 4) * 17; green = (green >> 4) * 17; blue = (blue >> 4) * 17;
    }
    out[i * 4] = red; out[i * 4 + 1] = green; out[i * 4 + 2] = blue;
  }
  return out;
}
inline std::string replayInfo() {
  auto s = rrreplay::replay.info();
  s.pop_back();
  return s + ",\"surface\":\"Tiled render targets mapped through captured display transfers\"}";
}
}
