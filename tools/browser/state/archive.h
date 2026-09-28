#pragma once
#include <array>
#include <bit>
#include <cstdint>
#include <cstring>
#include <deque>
#include <map>
#include <memory>
#include <stdexcept>
#include <string>
#include <typeindex>
#include <unordered_map>
#include <vector>
// Field-wise little-endian format. No padding, native pointers, or WASM heap
// dump.
namespace rrstate {
constexpr uint64_t limit = 128ull * 1024 * 1024;
struct Archive {
  bool reading = false;
  std::vector<uint8_t> bytes;
  size_t pos = 0;
  uint64_t allocation = 0;
  struct Ref {
    void *p;
    std::type_index type;
  };
  std::unordered_map<const void *, uint32_t> ids;
  std::vector<Ref> refs;
  std::vector<std::shared_ptr<void>> owned;
  Archive() = default;
  Archive(const uint8_t *p, size_t n) : reading(true), bytes(p, p + n) {
    if (n > limit)
      throw std::runtime_error("State too large");
  }
  void charge(uint64_t n) {
    if (n > limit - allocation)
      throw std::runtime_error("State allocation limit");
    allocation += n;
  }
  void raw(uint8_t *p, size_t n) {
    if (reading) {
      if (n > bytes.size() - pos)
        throw std::runtime_error("Truncated state");
      std::memcpy(p, bytes.data() + pos, n);
      pos += n;
    } else {
      if (n > limit - bytes.size())
        throw std::runtime_error("State size limit");
      bytes.insert(bytes.end(), p, p + n);
    }
  }
  template <class T>
    requires(std::is_arithmetic_v<T> || std::is_enum_v<T>)
  void one(T &v) {
    if constexpr (std::is_floating_point_v<T>) {
      if constexpr (sizeof(T) == 4) {
        auto b = std::bit_cast<uint32_t>(v);
        one(b);
        v = std::bit_cast<T>(b);
      } else {
        auto b = std::bit_cast<uint64_t>(v);
        one(b);
        v = std::bit_cast<T>(b);
      }
    } else {
      constexpr size_t n = std::is_same_v<T, bool> ? 1 : sizeof(T);
      uint64_t u = uint64_t(v);
      uint8_t b[n];
      for (size_t i = 0; i < n; i++)
        b[i] = u >> (i * 8);
      raw(b, n);
      if (reading) {
        u = 0;
        for (size_t i = 0; i < n; i++)
          u |= uint64_t(b[i]) << (i * 8);
        if constexpr (std::is_same_v<T, bool>)
          if (u > 1)
            throw std::runtime_error("Invalid boolean");
        v = T(u);
      }
    }
  }
  uint32_t count(size_t n) {
    uint32_t c = n;
    one(c);
    if (c > 16 * 1024 * 1024)
      throw std::runtime_error("Invalid state count");
    return c;
  }
  void one(std::string &v) {
    auto n = count(v.size());
    if (reading) {
      charge(n);
      v.resize(n);
    }
    raw((uint8_t *)v.data(), n);
  }
  template <class T, size_t N> void one(T (&v)[N]) {
    for (auto &x : v)
      one(x);
  }
  template <class T, size_t N> void one(std::array<T, N> &v) {
    for (auto &x : v)
      one(x);
  }
  template <class T> void one(std::vector<T> &v) {
    auto n = count(v.size());
    if (reading) {
      charge(uint64_t(n) * sizeof(T));
      v.resize(n);
    }
    for (auto &x : v)
      one(x);
  }
  template <class T> void one(std::deque<T> &v) {
    auto n = count(v.size());
    if (reading) {
      charge(uint64_t(n) * sizeof(T));
      v.resize(n);
    }
    for (auto &x : v)
      one(x);
  }
  template <class T> void one(T *&p) {
    uint32_t id = 0;
    if (!reading && p) {
      auto [it, fresh] = ids.emplace(p, ids.size() + 1);
      id = it->second;
      one(id);
      bool isNew = fresh;
      one(isNew);
      if (fresh)
        one(*p);
      return;
    }
    one(id);
    if (!id) {
      p = nullptr;
      return;
    }
    bool fresh = false;
    one(fresh);
    if (fresh) {
      if (id != refs.size() + 1 || id > 100000)
        throw std::runtime_error("Invalid object identity");
      charge(sizeof(T));
      auto obj = std::make_shared<T>();
      p = obj.get();
      refs.push_back({p, typeid(T)});
      owned.push_back(obj);
      one(*p);
    } else {
      if (id > refs.size() || refs[id - 1].type != typeid(T))
        throw std::runtime_error("Invalid object reference");
      p = (T *)refs[id - 1].p;
    }
  }
  template <class T>
    requires(!std::is_arithmetic_v<T> && !std::is_enum_v<T>)
  void one(T &v) {
    stateFields(*this, v);
  }
  template <class... T> void operator()(T &...v) { (one(v), ...); }
  void header(uint32_t platform, uint32_t version) {
    uint32_t magic = 0x53525231, p = platform, v = version;
    (*this)(magic, p, v);
    if (magic != 0x53525231 || p != platform || v != version)
      throw std::runtime_error("Incompatible core state");
  }
  void finish() {
    if (reading && pos != bytes.size())
      throw std::runtime_error("Trailing state bytes");
  }
};
} // namespace rrstate
