#pragma once
#include <algorithm>
#include <array>
#include <cstdint>
using u8 = uint8_t;
using u16 = uint16_t;
using u32 = uint32_t;
using u64 = uint64_t;
using i8 = int8_t;
using i16 = int16_t;
using i32 = int32_t;
using i64 = int64_t;
// Translation of tools/cpu/mips/gte.go. Keep its supported operations and flags.
struct GTE {
  std::array<u32, 32> data{}, ctrl{};
  static i32 s16(u32 v) { return i16(v); }
  static u32 lzc(u32 v) {
    u32 top = v & 0x80000000, n = 0;
    for (int i = 0; i < 32; i++, v <<= 1) {
      if ((v & 0x80000000) != top)
        break;
      n++;
    }
    return n;
  }
  u32 read(u32 r) const {
    r &= 31;
    if (r == 15)
      return data[14];
    if (r == 28 || r == 29)
      return std::clamp(s16(data[9]) >> 7, 0, 31) | (std::clamp(s16(data[10]) >> 7, 0, 31) << 5) |
             (std::clamp(s16(data[11]) >> 7, 0, 31) << 10);
    return data[r];
  }
  void write(u32 r, u32 v) {
    r &= 31;
    if (r == 15) {
      data[12] = data[13];
      data[13] = data[14];
      data[14] = v;
    } else if (r == 28) {
      for (int i = 0; i < 3; i++)
        data[9 + i] = ((v >> (i * 5)) & 31) * 128;
      data[28] = v & 0x7fff;
    } else {
      data[r] = v;
      if (r == 30)
        data[31] = lzc(v);
    }
  }
  i32 mat(int base, int r, int c) const {
    int i = r * 3 + c;
    return s16(ctrl[base + i / 2] >> ((i & 1) * 16));
  }
  std::array<i32, 3> vec(int n) const {
    return {s16(data[n * 2]), s16(data[n * 2] >> 16), s16(data[n * 2 + 1])};
  }
  i32 ir(int n) const { return s16(data[8 + n]); }
  void flag(u32 b) { ctrl[31] |= b; }
  void macCheck(int i, i64 v) {
    if (v > (i64(1) << 43) - 1)
      flag(1u << (31 - i));
    else if (v < -(i64(1) << 43))
      flag(1u << (28 - i));
  }
  void setIR(int i, i32 v, bool lm) {
    i32 lo = lm ? 0 : -32768;
    if (v > 32767) {
      v = 32767;
      flag((1u << 24) >> (i - 1));
    } else if (v < lo) {
      v = lo;
      flag((1u << 24) >> (i - 1));
    }
    data[8 + i] = u32(v);
  }
  void macIR(int i, i64 v, int sf, bool lm) {
    macCheck(i, v);
    i32 m = i32(v >> sf);
    data[24 + i] = u32(m);
    setIR(i, m, lm);
  }
  void mac0(i64 v) {
    if (v > INT32_MAX)
      flag(1 << 16);
    else if (v < INT32_MIN)
      flag(1 << 15);
    data[24] = u32(v);
  }
  i32 clamp(i32 v, i32 lo, i32 hi, u32 bit) {
    if (v < lo) {
      flag(bit);
      return lo;
    }
    if (v > hi) {
      flag(bit);
      return hi;
    }
    return v;
  }
  u32 divide(u32 h, u32 z);
  void rtp(int n, int sf, bool lm, bool depth) {
    auto v = vec(n);
    i64 full = 0;
    for (int r = 0; r < 3; r++) {
      full = i64(i32(ctrl[5 + r])) * 4096;
      for (int c = 0; c < 3; c++)
        full += i64(mat(0, r, c)) * v[c];
      macIR(r + 1, full, sf, lm);
    }
    data[16] = data[17];
    data[17] = data[18];
    data[18] = data[19];
    data[19] = clamp(i32(full >> 12), 0, 65535, 1 << 18);
    u32 d = divide(ctrl[26] & 65535, data[19] & 65535);
    i64 x = i64(d) * ir(1) + i32(ctrl[24]), y = i64(d) * ir(2) + i32(ctrl[25]);
    mac0(x);
    mac0(y);
    i32 sx = clamp(i32(x >> 16), -1024, 1023, 1 << 14),
        sy = clamp(i32(y >> 16), -1024, 1023, 1 << 13);
    data[12] = data[13];
    data[13] = data[14];
    data[14] = u16(sx) | (u32(u16(sy)) << 16);
    if (depth) {
      i64 z = i64(d) * s16(ctrl[27]) + i32(ctrl[28]);
      mac0(z);
      data[8] = clamp(i32(z >> 12), 0, 4096, 1 << 12);
    }
  }
  void mvmva(u32 cmd, int sf, bool lm) {
    int mx = (cmd >> 17) & 3, vs = (cmd >> 15) & 3, cv = (cmd >> 13) & 3;
    auto v = vs < 3 ? vec(vs) : std::array<i32, 3>{ir(1), ir(2), ir(3)};
    int base = mx == 0 ? 0 : mx == 1 ? 8 : 16;
    for (int r = 0; r < 3; r++) {
      i64 x = cv < 3 ? i64(i32(ctrl[(cv == 0 ? 5 : cv == 1 ? 13 : 21) + r])) * 4096 : 0;
      for (int c = 0; c < 3; c++)
        x += i64(mat(base, r, c)) * v[c];
      macIR(r + 1, x, sf, lm);
    }
  }
  void lightColor(int sf, bool lm) {
    std::array<i32, 3> v{ir(1), ir(2), ir(3)};
    for (int r = 0; r < 3; r++) {
      i64 x = i64(i32(ctrl[13 + r])) * 4096;
      for (int c = 0; c < 3; c++)
        x += i64(mat(16, r, c)) * v[c];
      macIR(r + 1, x, sf, lm);
    }
  }
  void primary(int sf, bool lm, bool depth) {
    std::array<i64, 3> mac;
    for (int i = 0; i < 3; i++)
      mac[i] = i64((data[6] >> (8 * i)) & 255) * ir(i + 1) * 16;
    if (depth)
      for (int i = 0; i < 3; i++) {
        i64 diff = i64(i32(ctrl[21 + i])) * 4096 - mac[i];
        setIR(i + 1, i32(diff >> sf), false);
        mac[i] += i64(ir(i + 1)) * ir(0);
      }
    for (int i = 0; i < 3; i++)
      macIR(i + 1, mac[i], sf, lm);
  }
  void pushColor() {
    u32 v = data[6] & 0xff000000;
    for (int i = 0; i < 3; i++)
      v |= u32(clamp(i32(data[25 + i]) >> 4, 0, 255, 1u << (21 - i))) << (i * 8);
    data[20] = data[21];
    data[21] = data[22];
    data[22] = v;
  }
  void normal(int n, int sf, bool lm, bool pri, bool dep) {
    auto v = vec(n);
    for (int r = 0; r < 3; r++) {
      i64 x = 0;
      for (int c = 0; c < 3; c++)
        x += i64(mat(8, r, c)) * v[c];
      macIR(r + 1, x, sf, lm);
    }
    lightColor(sf, lm);
    if (pri)
      primary(sf, lm, dep);
    pushColor();
  }
  void interpolate(int sf, bool lm, std::array<i64, 3> start) {
    for (int i = 0; i < 3; i++) {
      i64 diff = i64(i32(ctrl[21 + i])) * 4096 - start[i];
      setIR(i + 1, i32(diff >> sf), false);
      macIR(i + 1, start[i] + i64(ir(i + 1)) * ir(0), sf, lm);
    }
    pushColor();
  }
  void command(u32 cmd) {
    ctrl[31] = 0;
    int sf = (cmd & (1 << 19)) ? 12 : 0;
    bool lm = cmd & (1 << 10);
    switch (cmd & 63) {
    case 1:
      rtp(0, sf, lm, true);
      break;
    case 0x30:
      for (int n = 0; n < 3; n++)
        rtp(n, sf, lm, n == 2);
      break;
    case 6: {
      i64 x0 = s16(data[12]), y0 = s16(data[12] >> 16), x1 = s16(data[13]),
          y1 = s16(data[13] >> 16), x2 = s16(data[14]), y2 = s16(data[14] >> 16);
      mac0(x0 * y1 + x1 * y2 + x2 * y0 - x0 * y2 - x1 * y0 - x2 * y1);
      break;
    }
    case 0x12:
      mvmva(cmd, sf, lm);
      break;
    case 0x2d:
    case 0x2e: {
      bool four = (cmd & 63) == 0x2e;
      i64 sum = 0;
      for (int i = four ? 16 : 17; i <= 19; i++)
        sum += data[i] & 65535;
      i64 m = sum * s16(ctrl[four ? 30 : 29]);
      mac0(m);
      data[7] = clamp(i32(m >> 12), 0, 65535, 1 << 18);
      break;
    }
    case 0x1e:
      normal(0, sf, lm, false, false);
      break;
    case 0x20:
      for (int n = 0; n < 3; n++)
        normal(n, sf, lm, false, false);
      break;
    case 0x1b:
      normal(0, sf, lm, true, false);
      break;
    case 0x3f:
      for (int n = 0; n < 3; n++)
        normal(n, sf, lm, true, false);
      break;
    case 0x13:
      normal(0, sf, lm, true, true);
      break;
    case 0x16:
      for (int n = 0; n < 3; n++)
        normal(n, sf, lm, true, true);
      break;
    case 0x1c:
    case 0x14:
      lightColor(sf, lm);
      primary(sf, lm, (cmd & 63) == 0x14);
      pushColor();
      break;
    case 0x11:
      interpolate(sf, lm, {i64(ir(1)) * 4096, i64(ir(2)) * 4096, i64(ir(3)) * 4096});
      break;
    case 0x10:
      interpolate(sf, lm,
                  {i64(data[6] & 255) * 65536, i64((data[6] >> 8) & 255) * 65536,
                   i64((data[6] >> 16) & 255) * 65536});
      break;
    default:
      break; // Same documented unmodelled operations as Go.
    }
    if (ctrl[31] & 0x7f87e000)
      ctrl[31] |= 0x80000000;
  }
};

inline u32 GTE::divide(u32 h, u32 z) {
  static constexpr u8 table[257] = {
      0xFF, 0xFD, 0xFB, 0xF9, 0xF7, 0xF5, 0xF3, 0xF1, 0xEF, 0xEE, 0xEC, 0xEA, 0xE8, 0xE6, 0xE4,
      0xE3, 0xE1, 0xDF, 0xDD, 0xDC, 0xDA, 0xD8, 0xD6, 0xD5, 0xD3, 0xD1, 0xD0, 0xCE, 0xCD, 0xCB,
      0xC9, 0xC8, 0xC6, 0xC5, 0xC3, 0xC1, 0xC0, 0xBE, 0xBD, 0xBB, 0xBA, 0xB8, 0xB7, 0xB5, 0xB4,
      0xB2, 0xB1, 0xB0, 0xAE, 0xAD, 0xAB, 0xAA, 0xA9, 0xA7, 0xA6, 0xA4, 0xA3, 0xA2, 0xA0, 0x9F,
      0x9E, 0x9C, 0x9B, 0x9A, 0x99, 0x97, 0x96, 0x95, 0x94, 0x92, 0x91, 0x90, 0x8F, 0x8D, 0x8C,
      0x8B, 0x8A, 0x89, 0x87, 0x86, 0x85, 0x84, 0x83, 0x82, 0x81, 0x7F, 0x7E, 0x7D, 0x7C, 0x7B,
      0x7A, 0x79, 0x78, 0x77, 0x75, 0x74, 0x73, 0x72, 0x71, 0x70, 0x6F, 0x6E, 0x6D, 0x6C, 0x6B,
      0x6A, 0x69, 0x68, 0x67, 0x66, 0x65, 0x64, 0x63, 0x62, 0x61, 0x60, 0x5F, 0x5E, 0x5D, 0x5D,
      0x5C, 0x5B, 0x5A, 0x59, 0x58, 0x57, 0x56, 0x55, 0x54, 0x53, 0x53, 0x52, 0x51, 0x50, 0x4F,
      0x4E, 0x4D, 0x4D, 0x4C, 0x4B, 0x4A, 0x49, 0x48, 0x48, 0x47, 0x46, 0x45, 0x44, 0x43, 0x43,
      0x42, 0x41, 0x40, 0x3F, 0x3F, 0x3E, 0x3D, 0x3C, 0x3C, 0x3B, 0x3A, 0x39, 0x39, 0x38, 0x37,
      0x36, 0x36, 0x35, 0x34, 0x33, 0x33, 0x32, 0x31, 0x31, 0x30, 0x2F, 0x2E, 0x2E, 0x2D, 0x2C,
      0x2C, 0x2B, 0x2A, 0x2A, 0x29, 0x28, 0x28, 0x27, 0x26, 0x26, 0x25, 0x24, 0x24, 0x23, 0x22,
      0x22, 0x21, 0x20, 0x20, 0x1F, 0x1E, 0x1E, 0x1D, 0x1D, 0x1C, 0x1B, 0x1B, 0x1A, 0x19, 0x19,
      0x18, 0x18, 0x17, 0x16, 0x16, 0x15, 0x15, 0x14, 0x14, 0x13, 0x12, 0x12, 0x11, 0x11, 0x10,
      0x0F, 0x0F, 0x0E, 0x0E, 0x0D, 0x0D, 0x0C, 0x0C, 0x0B, 0x0A, 0x0A, 0x09, 0x09, 0x08, 0x08,
      0x07, 0x07, 0x06, 0x06, 0x05, 0x05, 0x04, 0x04, 0x03, 0x03, 0x02, 0x02, 0x01, 0x01, 0x00,
      0x00, 0x00,
  };
  if (h >= z * 2) {
    flag(1 << 17);
    return 0x1ffff;
  }
  unsigned shift = 0;
  for (u32 v = z; (v & 0x8000) == 0; v <<= 1)
    shift++;
  u32 n = h << shift, d = z << shift, u = u32(table[(d - 0x7fc0) >> 7]) + 0x101;
  d = (0x2000080 - d * u) >> 8;
  d = (0x80 + d * u) >> 8;
  return u32(std::min<u64>((u64(n) * d + 0x8000) >> 16, 0x1ffff));
}
