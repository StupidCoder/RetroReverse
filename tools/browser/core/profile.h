#pragma once
#include <array>
#include <chrono>
#include <sstream>
#include <string>
#ifdef __EMSCRIPTEN__
#include <emscripten/emscripten.h>
#endif
// Cumulative exclusive wall-time buckets. Parent scopes subtract every child,
// including recursive same-bucket work. No instruction-count approximations.
namespace rrprof {
inline std::array<double, 8> ms{}, longest{};
inline std::array<unsigned, 8> longestTag{};
inline std::array<const char *, 8> names{};
inline double now() {
#ifdef __EMSCRIPTEN__
  // Read the browser's monotonic millisecond clock directly. chrono's WASI
  // clock path converts it to 64-bit nanoseconds and back through JS BigInt.
  return emscripten_get_now();
#else
  return std::chrono::duration<double, std::milli>(
             std::chrono::steady_clock::now().time_since_epoch())
      .count();
#endif
}
struct Scope;
inline Scope *current = nullptr;
struct Scope {
  Scope *parent = nullptr;
  int bucket;
  double start = 0, children = 0;
  bool enabled;
  unsigned tag;
  Scope(int b, const char *name, bool on = true, unsigned id = 0)
      : bucket(b), enabled(on), tag(id) {
    if (on) {
      names[b] = name;
      parent = current;
      current = this;
      start = now();
    }
  }
  ~Scope() {
    if (enabled) {
      double elapsed = now() - start;
      ms[bucket] += elapsed - children;
      if (elapsed > longest[bucket]) {
        longest[bucket] = elapsed;
        longestTag[bucket] = tag;
      }
      if (parent)
        parent->children += elapsed;
      current = parent;
    }
  }
};
inline const char *json(bool sampled = false) {
  static std::string result;
  std::ostringstream s;
  s << "{\"sampled\":" << (sampled ? "true" : "false") << ",\"buckets\":[";
  bool comma = false;
  for (int i = 0; i < 8; i++)
    if (names[i]) {
      if (comma)
        s << ",";
      comma = true;
      s << "{\"name\":\"" << names[i] << "\",\"ms\":" << ms[i]
        << ",\"longestMs\":" << longest[i] << ",\"tag\":" << longestTag[i]
        << "}";
    }
  s << "]}";
  result = s.str();
  return result.c_str();
}
} // namespace rrprof
