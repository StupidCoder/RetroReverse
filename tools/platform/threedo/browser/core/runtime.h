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
struct threedo_Machine;
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
// Go's append(slice, other...) copies the range in bulk and permits overlap.
// Appending one byte at a time needlessly copied shared_ptr owners for every
// byte of every movie file, making a small guest READ a multi-second host call.
template <class T> Slice<T> append(Slice<T> s, Slice<T> v) {
  const auto count=v.n;
  if(count>s.c-s.n){
    int64_t capacity=std::max<int64_t>(8,s.c);
    while(capacity<s.n+count) capacity*=2;
    auto z=Slice<T>::make(s.n,capacity);gcopy(z,s);s=z;
  }
  if(count){
    if constexpr(std::is_trivially_copyable_v<T>)std::memmove(s.p+s.n,v.p,count*sizeof(T));
    else {std::vector<T> copy(v.begin(),v.end());std::copy(copy.begin(),copy.end(),s.p+s.n);}
  }
  s.n+=count;return s;
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
template <class T> std::ostream &operator<<(std::ostream &o, const Slice<T> &s) {
  o << "[";
  for (int64_t i = 0; i < s.n; i++) {
    if (i)
      o << " ";
    o << s[i];
  }
  return o << "]";
}
// Go object lifetimes are retained for a session and reclaimed together on reset.
inline std::vector<std::function<void()>> arena;
template <class T> T *arenaNew(T v) {
  auto p = new T(std::move(v));
  arena.push_back([p]() { delete p; });
  return p;
}
inline void arenaClear() {
  for (auto i = arena.rbegin(); i != arena.rend(); ++i)
    (*i)();
  arena.clear();
}
struct Error {
  std::string text;
  explicit operator bool() const { return !text.empty(); }
};
inline std::ostream &operator<<(std::ostream &s, const Error &e) {
  return s << e.text;
}
#include <cctype>
#include <iomanip>
template <class T> std::string formatOne(std::string spec, const T &v) {
  std::ostringstream s;
  char kind = spec.back();
  if (kind == 'x' || kind == 'X')
    s << std::hex;
  if (kind == 'X')
    s << std::uppercase;
  if (spec.find('-') != std::string::npos)
    s << std::left;
  if (spec.size() > 1 && spec[0] == '0')
    s << std::setfill('0');
  int width = 0;
  for (char c : spec) {
    if (c >= '0' && c <= '9')
      width = width * 10 + c - '0';
  }
  if (width > 0 && width < 10000)
    s << std::setw(width);
  if constexpr (std::is_same_v<T, bool>) {
    s << (v ? "true" : "false");
  } else if constexpr (std::is_integral_v<T>) {
    if (kind == 'c')
      s << char(v);
    else
      s << +v;
  } else if constexpr (std::is_same_v<T, std::string>) {
    if (kind == 'q')
      s << std::quoted(v);
    else
      s << v;
  } else {
    s << v;
  }
  return s.str();
}
inline std::string go_fmt_Sprintf(std::string f) {
  std::string s;
  for (size_t i = 0; i < f.size(); i++) {
    s += f[i];
    if (f[i] == '%' && i + 1 < f.size() && f[i + 1] == '%')
      i++;
  }
  return s;
}
template <class T, class... A> std::string go_fmt_Sprintf(std::string f, T v, A... a) {
  size_t p = 0;
  std::string pre;
  while (p < f.size()) {
    if (f[p] != '%') {
      pre += f[p++];
      continue;
    }
    if (p + 1 < f.size() && f[p + 1] == '%') {
      pre += '%';
      p += 2;
      continue;
    }
    size_t e = p + 1;
    while (e < f.size() && !std::isalpha((unsigned char)f[e]))
      e++;
    if (e == f.size())
      return pre + f.substr(p);
    return pre + formatOne(f.substr(p + 1, e - p), v) + go_fmt_Sprintf(f.substr(e + 1), a...);
  }
  return pre;
}
template <class... A> Error go_fmt_Errorf(std::string f, A... a) {
  return {go_fmt_Sprintf(f, a...)};
}
template <class K, class V, class Q> void removeKey(Map<K, V> &m, Q k) {
  m.p->erase(k);
}
template <class T, class U> Slice<T> append(Slice<T> s, U v) {
  return append(s, T(v));
}
template <class T, class U, class... A> Slice<T> append(Slice<T> s, U v, A... a) {
  return append(append(s, T(v)), a...);
}
template <> inline Slice<uint8_t> cast<Slice<uint8_t>, std::string>(std::string s) {
  auto b = Slice<uint8_t>::make(s.size());
  std::copy(s.begin(), s.end(), b.p);
  return b;
}
inline int64_t go_bits_OnesCount32(uint32_t n) {
  return std::popcount(n);
}
inline uint16_t be_Uint16(Slice<uint8_t> s) {
  return uint16_t(s[0]) << 8 | s[1];
}
inline void be_PutUint16(Slice<uint8_t> s, uint16_t v) {
  s[0] = v >> 8;
  s[1] = v;
}
inline std::string go_strings_ToLower(std::string s) {
  for (auto &c : s)
    c = std::tolower((unsigned char)c);
  return s;
}
inline bool go_strings_EqualFold(std::string a, std::string b) {
  return go_strings_ToLower(a) == go_strings_ToLower(b);
}
inline bool go_strings_Contains(std::string a, std::string b) {
  return a.find(b) != std::string::npos;
}
inline bool go_strings_HasPrefix(std::string a, std::string b) {
  return a.starts_with(b);
}
inline int64_t go_strings_LastIndexByte(std::string a, uint8_t b) {
  auto i = a.rfind(char(b));
  return i == std::string::npos ? -1 : int64_t(i);
}
inline std::string go_strings_TrimLeft(std::string s, std::string chars) {
  auto p = s.find_first_not_of(chars);
  return p == std::string::npos ? "" : s.substr(p);
}
inline std::string go_strings_TrimRight(std::string s, std::string chars) {
  auto p = s.find_last_not_of(chars);
  return p == std::string::npos ? "" : s.substr(0, p + 1);
}
inline std::string go_strings_Trim(std::string s, std::string c) {
  return go_strings_TrimRight(go_strings_TrimLeft(s, c), c);
}
inline std::string go_strings_TrimSpace(std::string s) {
  return go_strings_Trim(s, " \t\n\r\v\f");
}
inline Slice<std::string> go_strings_Split(std::string s, std::string sep) {
  Slice<std::string> o;
  size_t p = 0;
  while (true) {
    auto e = s.find(sep, p);
    if (e == std::string::npos) {
      o = append(o, s.substr(p));
      break;
    }
    o = append(o, s.substr(p, e - p));
    p = e + sep.size();
  }
  return o;
}
template <class T, class F> void go_sort_Slice(Slice<T> s, F less) {
  for (int64_t i = 1; i < s.n; i++)
    for (int64_t j = i; j > 0 && less(j, j - 1); j--)
      std::swap(s[j], s[j - 1]);
}
inline void go_sort_Strings(Slice<std::string> s) {
  std::sort(s.begin(), s.end());
}
using time_Time = int64_t;
inline time_Time go_time_Now() {
  return std::chrono::duration_cast<std::chrono::nanoseconds>(
             std::chrono::steady_clock::now().time_since_epoch())
      .count();
}
inline int64_t go_time_Since(time_Time t) {
  return go_time_Now() - t;
}
inline bool time_Time_IsZero(time_Time t) {
  return t == 0;
}
struct color_RGBA {
  uint8_t R{}, G{}, B{}, A{};
};
struct image_Point {
  int64_t X{}, Y{};
};
struct image_Rectangle {
  image_Point Min{}, Max{};
};
struct image_RGBA {
  Slice<uint8_t> Pix;
  int64_t Stride;
  image_Rectangle Rect;
};
inline image_Rectangle go_image_Rect(int64_t x0, int64_t y0, int64_t x1, int64_t y1) {
  return {{x0, y0}, {x1, y1}};
}
inline image_RGBA *go_image_NewRGBA(image_Rectangle r) {
  auto w = r.Max.X - r.Min.X, h = r.Max.Y - r.Min.Y;
  if (w < 0 || h < 0 || w > 4096 || h > 4096)
    throw std::runtime_error("invalid image size");
  return arenaNew(image_RGBA{Slice<uint8_t>::make(w * h * 4), w * 4, r});
}
inline int64_t image_RGBA_PixOffset(image_RGBA *p, int64_t x, int64_t y) {
  return (y - p->Rect.Min.Y) * p->Stride + (x - p->Rect.Min.X) * 4;
}
inline void image_RGBA_SetRGBA(image_RGBA *p, int64_t x, int64_t y, color_RGBA c) {
  if (x < p->Rect.Min.X || x >= p->Rect.Max.X || y < p->Rect.Min.Y || y >= p->Rect.Max.Y)
    return;
  auto i = image_RGBA_PixOffset(p, x, y);
  p->Pix[i] = c.R;
  p->Pix[i + 1] = c.G;
  p->Pix[i + 2] = c.B;
  p->Pix[i + 3] = c.A;
}
