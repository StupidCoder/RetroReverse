#pragma once
#include <algorithm>
#include <any>
#include <array>
#include <bit>
#include <chrono>
#include <cmath>
#include <cstdint>
#include <cstring>
#include <fstream>
#include <functional>
#include <iostream>
#include <limits>
#include <memory>
#include <sstream>
#include <stdexcept>
#include <string>
#include <tuple>
#include <unordered_map>
#include <vector>
struct n64_Machine;
template <class T> struct Slice {
  std::shared_ptr<T[]> owner;
  T *p = nullptr;
  int64_t n = 0, c = 0;
  Slice() = default;
  Slice(std::initializer_list<T> v) {
    *this = make(v.size());
    std::copy(v.begin(), v.end(), p);
  }
  static Slice make(int64_t n, int64_t cap = -1) {
    if (cap < 0)
      cap = n;
    if (n < 0 || cap < n || cap > 134217728)
      throw std::runtime_error("invalid slice allocation");
    Slice s;
    s.owner = std::shared_ptr<T[]>(new T[cap]{});
    s.p = s.owner.get();
    s.n = n;
    s.c = cap;
    return s;
  }
  T &operator[](int64_t i) const {
    if (i < 0 || i >= n)
      throw std::runtime_error("slice bounds");
    return p[i];
  }
  explicit operator bool() const { return p != nullptr; }
  T *begin() const { return p; }
  T *end() const { return p + n; }
};
template <class T> int64_t len(const T &x) {
  return x.size();
}
template <class T> int64_t len(const Slice<T> &x) {
  return x.n;
}
template <class T> Slice<T> sub(Slice<T> s, int64_t a, int64_t b) {
  if (a < 0 || b < a || b > s.c)
    throw std::runtime_error("slice bounds");
  s.p += a;
  s.n = b - a;
  s.c -= a;
  return s;
}
template <class T, size_t N> Slice<T> sub(std::array<T, N> &a, int64_t lo, int64_t hi) {
  if (lo < 0 || hi < lo || hi > N)
    throw std::runtime_error("array bounds");
  Slice<T> s;
  s.p = a.data() + lo;
  s.n = hi - lo;
  s.c = N - lo;
  return s;
}
inline std::string sub(std::string s, int64_t a, int64_t b) {
  return s.substr(a, b - a);
}
template <class T, class U> int64_t gcopy(T d, U s) {
  auto n = std::min(len(d), len(s));
  for (int64_t i = 0; i < n; i++)
    d[i] = s[i];
  return n;
}
template <class T> Slice<T> append(Slice<T> s, T v) {
  if (s.n == s.c) {
    auto z = Slice<T>::make(s.n, std::max<int64_t>(8, s.n * 2));
    gcopy(z, s);
    s = z;
  }
  s.p[s.n++] = v;
  return s;
}
template <class T> Slice<T> append(Slice<T> s, Slice<T> v) {
  for (auto x : v)
    s = append(s, x);
  return s;
}
template <class K, class V> struct Map {
  std::shared_ptr<std::unordered_map<K, V>> p = std::make_shared<std::unordered_map<K, V>>();
  Map() = default;
  Map(std::initializer_list<std::pair<const K, V>> v)
      : p(std::make_shared<std::unordered_map<K, V>>(v)) {}
  V &operator[](K k) { return (*p)[k]; }
  explicit operator bool() const { return bool(p); }
  auto begin() { return p->begin(); }
  auto end() { return p->end(); }
  auto size() const { return p->size(); }
};
template <class K, class V, class Q> V get(const Map<K, V> &m, Q k) {
  auto i = m.p->find(k);
  return i == m.p->end() ? V{} : i->second;
}
template <class K, class V, class Q> auto lookup(Map<K, V> m, Q k) {
  auto i = m.p->find(k);
  return std::make_tuple(i == m.p->end() ? V{} : i->second, i != m.p->end());
}
template <class T, class U> constexpr T cast(U v) {
  if constexpr (std::is_integral_v<T> && std::is_floating_point_v<U>) {
    if (!std::isfinite(v) || v < static_cast<long double>(std::numeric_limits<T>::min()) ||
        v >= static_cast<long double>(std::numeric_limits<T>::max()) + 1)
      return std::numeric_limits<T>::min();
  }
  return static_cast<T>(v);
}
template <class T> T cast(Slice<uint8_t> v) {
  if constexpr (std::is_same_v<T, std::string>)
    return std::string((char *)v.p, v.n);
  else
    return T{};
}
template <> inline std::string cast<std::string, char>(char v) {
  return std::string(1, v);
}
template <> inline std::string cast<std::string, int32_t>(int32_t v) {
  return std::string(1, char(v));
}
template <class T, class A, class B> constexpr T shl(A a, B b) {
  if (b < 0)
    throw std::runtime_error("negative shift");
  if (uint64_t(b) >= sizeof(T) * 8)
    return 0;
  return T(std::make_unsigned_t<T>(a) << b);
}
template <class T, class A, class B> constexpr T shr(A a, B b) {
  if (b < 0)
    throw std::runtime_error("negative shift");
  if (uint64_t(b) >= sizeof(T) * 8)
    return T(a) < 0 ? T(-1) : T(0);
  return T(a) >> b;
}
template <class T, class A, class B> T divi(A aa, B bb) {
  T a = aa, b = bb;
  if (!b)
    throw std::runtime_error("division by zero");
  if constexpr (std::is_signed_v<T>) {
    if (a == std::numeric_limits<T>::min() && b == -1)
      return a;
  }
  return a / b;
}
template <class T, class A, class B> T modi(A aa, B bb) {
  T a = aa, b = bb;
  if (!b)
    throw std::runtime_error("modulo by zero");
  if constexpr (std::is_signed_v<T>) {
    if (a == std::numeric_limits<T>::min() && b == -1)
      return 0;
  }
  return a % b;
}
template <class A, class B> auto gmin(A a, B b) {
  using T = std::common_type_t<A, B>;
  return std::min(T(a), T(b));
}
template <class A, class B> auto gmax(A a, B b) {
  using T = std::common_type_t<A, B>;
  return std::max(T(a), T(b));
}
template <class... A> std::string go_fmt_Sprintf(std::string f, A... a) {
  std::ostringstream s;
  s << f;
  ((s << " " << a), ...);
  return s.str();
}
template <class... A> std::string go_fmt_Errorf(std::string f, A... a) {
  return go_fmt_Sprintf(f, a...);
}
inline uint32_t be_Uint32(Slice<uint8_t> s) {
  return uint32_t(s[0]) << 24 | uint32_t(s[1]) << 16 | uint32_t(s[2]) << 8 | s[3];
}
inline uint64_t be_Uint64(Slice<uint8_t> s) {
  return uint64_t(be_Uint32(s)) << 32 | be_Uint32(sub(s, 4, s.n));
}
inline void be_PutUint32(Slice<uint8_t> s, uint32_t v) {
  for (int i = 0; i < 4; i++)
    s[i] = v >> (24 - i * 8);
}
inline auto go_bits_Mul64(uint64_t a, uint64_t b) {
  __uint128_t v = __uint128_t(a) * b;
  return std::make_tuple(uint64_t(v >> 64), uint64_t(v));
}
inline float go_math_Float32frombits(uint32_t x) {
  return std::bit_cast<float>(x);
}
inline double go_math_Float64frombits(uint64_t x) {
  return std::bit_cast<double>(x);
}
inline uint32_t go_math_Float32bits(float x) {
  return std::bit_cast<uint32_t>(x);
}
inline uint64_t go_math_Float64bits(double x) {
  return std::bit_cast<uint64_t>(x);
}
inline double go_math_NaN() {
  return std::numeric_limits<double>::quiet_NaN();
}
inline double go_math_Inf(int s) {
  return s < 0 ? -INFINITY : INFINITY;
}
inline bool go_math_IsNaN(double x) {
  return std::isnan(x);
}
inline bool go_math_IsInf(double x, int s) {
  return std::isinf(x) && (s == 0 || (s < 0 ? x < 0 : x > 0));
}
inline bool go_math_Signbit(double x) {
  return std::signbit(x);
}
inline double go_math_Copysign(double x, double s) {
  return std::copysign(x, s);
}
inline double go_math_Trunc(double x) {
  return std::trunc(x);
}
inline double go_math_Ceil(double x) {
  return std::ceil(x);
}
inline double go_math_Floor(double x) {
  return std::floor(x);
}
inline double go_math_RoundToEven(double x) {
  return std::nearbyint(x);
}
inline double go_math_Sqrt(double x) {
  return std::sqrt(x);
}
inline double go_math_Abs(double x) {
  return std::abs(x);
}
inline double go_math_Nextafter(double x, double y) {
  return std::nextafter(x, y);
}
inline float go_math_Nextafter32(float x, float y) {
  return std::nextafter(x, y);
}
inline double go_math_FMA(double a, double b, double c) {
  return std::fma(a, b, c);
}
template <class F> struct Defer {
  F f;
  ~Defer() { f(); }
};
template <class F> auto defer(F f) {
  return Defer<F>{f};
}
