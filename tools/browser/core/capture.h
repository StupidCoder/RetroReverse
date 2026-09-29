#pragma once
#include <algorithm>
#include <cstdint>
#include <sstream>
#include <string>
#include <unordered_map>
#include <vector>
#ifndef RR_CAPTURE_WRITE_CAP
#define RR_CAPTURE_WRITE_CAP 1048576
#endif
#ifndef RR_CAPTURE_EVENT_CAP
#define RR_CAPTURE_EVENT_CAP 16384
#endif
#ifndef RR_CAPTURE_META_CAP
#define RR_CAPTURE_META_CAP (16 * 1024 * 1024)
#endif
namespace rrcapture {
struct Event {
  uint64_t clock;
  uint32_t pc;
  std::string detail;
  uint32_t resource = 0, sourceBefore = 0;
};
struct Write {
  uint64_t clock = 0;
  uint32_t address = 0, before = 0, after = 0, event = 0, pc = 0, source = 0,
           palette = 0, texel = 0;
  uint32_t sourceValue = 0, paletteValue = 0, sourceBefore = 0,
           paletteBefore = 0;
  int32_t u = 0, v = 0, depth = 0;
  uint8_t size = 0,
          flags = 0; // 1 written, 2 depth reject, 4 alpha/transparent reject, 8 repeated polygon ID, 16 stencil, 32 scissor, 64 write mask
};
struct Recorder {
  static constexpr size_t WRITE_CAP = RR_CAPTURE_WRITE_CAP, EVENT_CAP = RR_CAPTURE_EVENT_CAP,
                          META_CAP = RR_CAPTURE_META_CAP;
  bool active = false, valid = false, rendering = false;
  uint32_t base = 0, current = 0, overflow = 0;
  uint32_t source = 0, palette = 0, texel = 0, sourceValue = 0,
           paletteValue = 0, sourceBefore = 0, paletteBefore = 0;
  int32_t u = 0, v = 0;
  size_t metadata = 0;
  std::vector<uint8_t> initial, shadow, final;
  std::vector<Write> writes;
  std::vector<Event> events;
  std::vector<std::vector<uint8_t>> resources;
  std::unordered_map<uint32_t, std::vector<uint32_t>> resourceHashes;
  uint32_t resource(const uint8_t *p, size_t n) {
    if (!active)
      return 0;
    uint32_t h = 2166136261u;
    for (size_t i = 0; i < n; i++)
      h = (h ^ p[i]) * 16777619u;
    for (auto id : resourceHashes[h])
      if (resources[id - 1].size() == n &&
          std::equal(resources[id - 1].begin(), resources[id - 1].end(), p))
        return id;
    if (n > META_CAP - metadata) {
      overflow++;
      return 0;
    }
    metadata += n;
    resources.emplace_back(p, p + n);
    auto id = uint32_t(resources.size());
    resourceHashes[h].push_back(id);
    return id;
  }
  void begin(const uint8_t *p, size_t n, uint32_t b = 0) {
    base = b;
    initial.assign(p, p + n);
    shadow = initial;
    final.clear();
    writes.clear();
    events.clear();
    resources.clear();
    resourceHashes.clear();
    current = overflow = 0;
    metadata = 0;
    source = palette = texel = 0;
    u = v = 0;
    active = true;
    valid = false;
  }
  uint32_t event(uint64_t clock, uint32_t pc, std::string detail,
                 uint32_t resourceId = 0) {
    if (!active)
      return 0;
    if (events.size() >= EVENT_CAP || detail.size() > META_CAP - metadata) {
      overflow++;
      current = 0;
      return 0;
    }
    metadata += detail.size();
    events.push_back(
        {clock, pc, std::move(detail), resourceId, uint32_t(writes.size())});
    return current = events.size();
  }
  uint32_t value(const std::vector<uint8_t> &b, uint32_t address,
                 int size) const {
    if (address < base || uint64_t(address - base) + size > b.size())
      return 0;
    uint32_t n = 0;
    for (int i = 0; i < size; i++)
      n |= uint32_t(b[address - base + i]) << (i * 8);
    return n;
  }
  void record(uint32_t address, uint32_t after, int size, uint64_t clock,
              uint32_t pc, uint32_t eventId = 0, uint8_t flags = 1,
              int32_t depth = 0) {
    if (!active || address < base || size < 1 || size > 4 ||
        uint64_t(address - base) + size > shadow.size())
      return;
    if (writes.size() >= WRITE_CAP) {
      overflow++;
      return;
    }
    writes.push_back({clock, address, value(shadow, address, size), after,
                      eventId, pc, source, palette, texel, sourceValue,
                      paletteValue, sourceBefore, paletteBefore, u, v, depth,
                      uint8_t(size), flags});
    if (flags & 1)
      for (int i = 0; i < size; i++)
        shadow[address - base + i] = after >> (i * 8);
  }
  void end(const uint8_t *p, size_t n) {
    active = false;
    valid = true;
    final.assign(p, p + n);
  }
  std::string info() const {
    std::ostringstream o;
    o << "{\"events\":" << events.size() << ",\"writes\":" << writes.size()
      << ",\"overflow\":" << overflow << ",\"bytes\":"
      << (initial.size() + shadow.size() + final.size() +
          writes.size() * sizeof(Write) + metadata)
      << "}";
    return o.str();
  }
  std::string resourceJSON(uint32_t id, uint32_t offset) const {
    if (!id || id > resources.size())
      return "{\"error\":\"No source snapshot\"}";
    const auto &b = resources[id - 1];
    if (offset >= b.size())
      return "{\"error\":\"Outside source snapshot\"}";
    std::ostringstream o;
    o << "{\"offset\":" << offset << ",\"size\":" << b.size() << ",\"bytes\":[";
    for (size_t i = offset; i < std::min<size_t>(offset + 32, b.size()); i++) {
      if (i != offset)
        o << ',';
      o << unsigned(b[i]);
    }
    o << "]}";
    return o.str();
  }
  std::string pixel(uint32_t address, int size, uint32_t before = 0xffffffff,
                    uint32_t sourceExpected = 0) const {
    if (!valid || address < base || size < 1 || size > 4 ||
        uint64_t(address - base) + size > final.size())
      return "{\"error\":\"No captured memory at this pixel\"}";
    std::ostringstream o;
    auto expected = before == 0xffffffff ? value(final, address, size)
                                         : sourceExpected,
         observed = value(shadow, address, size);
    if (before != 0xffffffff) {
      observed = value(initial, address, size);
      for (size_t i = 0; i < std::min<size_t>(before, writes.size()); i++) {
        const auto &w = writes[i];
        if (!(w.flags & 1))
          continue;
        for (int j = 0; j < w.size; j++)
          if (w.address + j >= address && w.address + j < address + size) {
            auto shift = (w.address + j - address) * 8;
            observed = (observed & ~(255u << shift)) |
                       (((w.after >> (j * 8)) & 255) << shift);
          }
      }
    }
    o << "{\"address\":" << address << ",\"size\":" << size
      << ",\"initial\":" << value(initial, address, size)
      << ",\"final\":" << expected << ",\"reconstructed\":" << observed
      << ",\"complete\":"
      << (!overflow && expected == observed ? "true" : "false")
      << ",\"overflow\":" << overflow << ",\"contributors\":[";
    size_t count = 0, matched = 0;
    for (size_t i = 0; i < std::min<size_t>(before, writes.size()); i++) {
      const auto &w = writes[i];
      if (uint64_t(w.address) + w.size <= address ||
          w.address >= uint64_t(address) + size)
        continue;
      matched++;
      if (count >= 512)
        continue;
      if (count++)
        o << ',';
      o << "{\"id\":" << i + 1 << ",\"event\":" << w.event
        << ",\"clock\":" << w.clock << ",\"pc\":" << w.pc
        << ",\"address\":" << w.address << ",\"size\":" << unsigned(w.size)
        << ",\"before\":" << w.before << ",\"after\":" << w.after
        << ",\"drawn\":" << ((w.flags & 1) ? "true" : "false")
        << ",\"depthRejected\":" << ((w.flags & 2) ? "true" : "false")
        << ",\"alphaRejected\":" << ((w.flags & 4) ? "true" : "false")
        << ",\"idRejected\":" << ((w.flags & 8) ? "true" : "false")
        << ",\"stencilRejected\":" << ((w.flags & 16) ? "true" : "false")
        << ",\"scissorRejected\":" << ((w.flags & 32) ? "true" : "false")
        << ",\"maskRejected\":" << ((w.flags & 64) ? "true" : "false")
        << ",\"depth\":" << w.depth << ",\"sourceAddress\":" << w.source
        << ",\"paletteAddress\":" << w.palette << ",\"texel\":" << w.texel
        << ",\"sourceValue\":" << w.sourceValue
        << ",\"paletteValue\":" << w.paletteValue
        << ",\"sourceBefore\":" << w.sourceBefore
        << ",\"paletteBefore\":" << w.paletteBefore << ",\"u\":" << w.u
        << ",\"v\":" << w.v;
      if (w.event && w.event <= events.size()) {
        const auto &e = events[w.event - 1];
        o << ",\"submissionPC\":" << e.pc << ",\"submissionClock\":" << e.clock
          << ",\"command\":" << e.detail
          << ",\"sourceSnapshot\":" << e.resource;
      } else
        o << ",\"command\":{\"kind\":\"CPU / HLE memory write\"}";
      o << '}';
    }
    o << "],\"contributorCount\":" << matched
      << ",\"truncated\":" << (matched > count ? "true" : "false")
      << ",\"ancestry\":\"History before this capture is not recorded. Source "
         "addresses alone are not writer ancestry.\"}";
    return o.str();
  }
};
inline Recorder trace;
struct RenderScope {
  bool before;
  RenderScope() : before(trace.rendering) { trace.rendering = trace.active; }
  ~RenderScope() { trace.rendering = before; }
};
inline uint32_t little(const uint8_t *p, int n) {
  uint32_t v = 0;
  for (int i = 0; i < n; i++)
    v |= uint32_t(p[i]) << (i * 8);
  return v;
}
} // namespace rrcapture
