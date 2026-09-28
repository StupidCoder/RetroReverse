#include "../../../../browser/core/profile.h"
#pragma once
#include "gte.h"
#include <functional>
#include <vector>
// Software GPU port of gpu.go / gpu_raster.go, including their documented approximations.
struct GPU {
  static constexpr int W = 1024, H = 512;
  std::vector<u16> vram = std::vector<u16>(W * H);
  std::vector<u32> fifo;
  int need = 0;
  int imgX = 0, imgY = 0, imgW = 0, imgH = 0, imgCurX = 0, imgCurY = 0;
  i64 imgPx = 0;
  int rdX = 0, rdY = 0, rdW = 0, rdH = 0, rdCurX = 0, rdCurY = 0;
  i64 rdPx = 0;
  int drawL = 0, drawT = 0, drawR = W, drawB = H, offX = 0, offY = 0, texPageX = 0, texPageY = 0,
      texDepth = 0;
  int texWinMX = 0, texWinMY = 0, texWinOX = 0, texWinOY = 0, dispX = 0, dispY = 0, dispW = 320,
      dispH = 240;
  bool dispEnabled = true;
  u32 gp0Read = 0, statField = 0;
  u64 commands = 0, words = 0, pixels = 0;
  u32 commandHash = 2166136261u;
  // Optional debugger hooks are host state, cleared on snapshot restore.
  std::function<void(const std::vector<u32> &)> onCommand;
  std::function<void(int, int, u16)> onPixel;
  void store(int x, int y, u16 v) {
    vram[y * W + x] = v;
    pixels++;
    if (onPixel)
      onPixel(x, y, v);
  }
  u32 status() {
    statField ^= 0x80000000;
    return 0x1c000000 | statField;
  }
  u32 readWord() {
    u32 w = 0;
    for (int i = 0; i < 2 && rdPx > 0; i++) {
      int x = (rdX + rdCurX) & 1023, y = (rdY + rdCurY) & 511;
      w |= u32(vram[y * W + x]) << (16 * i);
      rdPx--;
      if (++rdCurX >= rdW) {
        rdCurX = 0;
        rdCurY++;
      }
    }
    return gp0Read = w;
  }
  u32 read() { return rdPx > 0 ? readWord() : gp0Read; }
  static int sext11(u32 v) { return (v & 0x400) ? int(v) - 0x800 : int(v); }
  static u16 color15(u32 v) {
    return ((v & 255) >> 3) | (((v >> 8) & 255) >> 3 << 5) | (((v >> 16) & 255) >> 3 << 10);
  }
  static u32 rgba(u16 v) {
    return ((v & 31) << 3) | (((v >> 5) & 31) << 11) | (((v >> 10) & 31) << 19) | 0xff000000;
  }
  void gp1(u32 w) {
    switch (w >> 24) {
    case 0:
      dispEnabled = false;
      [[fallthrough]];
    case 1:
      need = imgPx = 0;
      fifo.clear();
      break;
    case 3:
      dispEnabled = !(w & 1);
      break;
    case 5:
      dispX = w & 1023;
      dispY = (w >> 10) & 511;
      break;
    case 8: {
      constexpr int widths[] = {256, 320, 512, 640};
      dispW = (w & 4) ? 368 : widths[w & 3];
      dispH = (w & 32) ? 480 : 240;
      break;
    }
    }
  }
  static int length(u8 op) {
    if (op == 2)
      return 3;
    if (op >= 0x20 && op <= 0x3f) {
      int n = op & 8 ? 4 : 3;
      return n + (op & 16 ? n : 1) + (op & 4 ? n : 0);
    }
    if (op >= 0x40 && op <= 0x5f)
      return op & 8 ? 4 : 3;
    if (op >= 0x60 && op <= 0x7f)
      return 2 + (op & 4 ? 1 : 0) + ((op & 24) == 0 ? 1 : 0);
    if (op >= 0x80 && op <= 0x9f)
      return 4;
    if (op >= 0xa0 && op <= 0xdf)
      return 3;
    return 1;
  }
  void gp0(u32 w) {
    words++;
    if (imgPx) {
      for (int i = 0; i < 2 && imgPx > 0; i++) {
        store((imgX + imgCurX) & 1023, (imgY + imgCurY) & 511, u16(w >> (i * 16)));
        imgPx--;
        if (++imgCurX >= imgW) {
          imgCurX = 0;
          imgCurY++;
        }
      }
      return;
    }
    if (!need) {
      need = length(w >> 24);
      fifo.clear();
    }
    fifo.push_back(w);
    if (--need == 0)
      exec();
  }
  void feedNode(const std::vector<u32> &ws) {
    if (!need && !imgPx && ws.size() >= 2 && !(ws[0] >> 24))
      return;
    for (u32 w : ws)
      gp0(w);
  }
  void setting(u8 op) {
    u32 w = fifo[0];
    switch (op) {
    case 0xe1:
      texPageX = (w & 15) * 64;
      texPageY = ((w >> 4) & 1) * 256;
      texDepth = (w >> 7) & 3;
      break;
    case 0xe2:
      texWinMX = w & 31;
      texWinMY = (w >> 5) & 31;
      texWinOX = (w >> 10) & 31;
      texWinOY = (w >> 15) & 31;
      break;
    case 0xe3:
      drawL = w & 1023;
      drawT = (w >> 10) & 1023;
      break;
    case 0xe4:
      drawR = (w & 1023) + 1;
      drawB = ((w >> 10) & 1023) + 1;
      break;
    case 0xe5:
      offX = sext11(w & 2047);
      offY = sext11((w >> 11) & 2047);
      break;
    }
  }
  void exec() {
    rrprof::Scope clock(2,"GPU / software rasterizer");
    commands++;
    for (u32 w : fifo)
      for (int j = 0; j < 4; j++)
        commandHash = (commandHash ^ u8(w >> (8 * j))) * 16777619u;
    if (onCommand)
      onCommand(fifo);
    u8 op = fifo[0] >> 24;
    if (op == 2) {
      u16 c = color15(fifo[0]);
      int x = fifo[1] & 1023, y = (fifo[1] >> 16) & 511, w = fifo[2] & 1023,
          h = (fifo[2] >> 16) & 511;
      for (int py = y; py < y + h && py < H; py++)
        for (int px = x; px < x + w && px < W; px++)
          store(px, py, c);
    } else if (op >= 0x20 && op <= 0x3f)
      polygon(op);
    else if (op >= 0x60 && op <= 0x7f)
      rect(op);
    else if (op >= 0x80 && op <= 0x9f) {
      int sx = fifo[1] & 1023, sy = (fifo[1] >> 16) & 511, dx = fifo[2] & 1023,
          dy = (fifo[2] >> 16) & 511, w = fifo[3] & 1023, h = (fifo[3] >> 16) & 511;
      for (int y = 0; y < h; y++)
        for (int x = 0; x < w; x++)
          store((dx + x) & 1023, (dy + y) & 511, vram[((sy + y) & 511) * W + ((sx + x) & 1023)]);
    } else if (op >= 0xa0 && op <= 0xbf) {
      imgX = fifo[1] & 1023;
      imgY = (fifo[1] >> 16) & 511;
      imgW = fifo[2] & 65535;
      imgH = fifo[2] >> 16;
      if (!imgW)
        imgW = W;
      if (!imgH)
        imgH = H;
      imgCurX = imgCurY = 0;
      imgPx = i64(imgW) * imgH;
    } else if (op >= 0xc0 && op <= 0xdf) {
      rdX = fifo[1] & 1023;
      rdY = (fifo[1] >> 16) & 511;
      rdW = fifo[2] & 65535;
      rdH = fifo[2] >> 16;
      if (!rdW)
        rdW = W;
      if (!rdH)
        rdH = H;
      rdCurX = rdCurY = 0;
      rdPx = i64(rdW) * rdH;
    } else if (op >= 0xe1 && op <= 0xe6)
      setting(op);
    need = 0;
  }
  struct Vert {
    int x = 0, y = 0, r = 0, g = 0, b = 0, u = 0, v = 0;
  };
  static i64 edge(int ax, int ay, int bx, int by, int px, int py) {
    return i64(bx - ax) * (py - ay) - i64(by - ay) * (px - ax);
  }
  static u16 modulate(u16 t, int r, int g, int b) {
    if (!(r | g | b))
      return t;
    return std::min((t & 31) * r >> 7, 31) | (std::min(((t >> 5) & 31) * g >> 7, 31) << 5) |
           (std::min(((t >> 10) & 31) * b >> 7, 31) << 10) | (t & 0x8000);
  }
  u16 texel(int u, int v, u32 clut) const {
    u &= 255;
    v &= 255;
    u = (u & ~(texWinMX * 8)) | ((texWinOX & texWinMX) * 8);
    v = (v & ~(texWinMY * 8)) | ((texWinOY & texWinMY) * 8);
    int cb = ((clut >> 6) & 511) * W + (clut & 63) * 16;
    if (texDepth < 2) {
      int div = texDepth ? 2 : 4, shift = texDepth ? (u & 1) * 8 : (u & 3) * 4,
          mask = texDepth ? 255 : 15;
      u16 w = vram[((texPageY + v) & 511) * W + ((texPageX + u / div) & 1023)];
      return vram[(cb + ((w >> shift) & mask)) & (W * H - 1)];
    }
    return vram[((texPageY + v) & 511) * W + ((texPageX + u) & 1023)];
  }
  void tri(Vert a, Vert b, Vert c, bool textured, u32 clut) {
    i64 area = edge(a.x, a.y, b.x, b.y, c.x, c.y);
    if (!area)
      return;
    if (area < 0) {
      std::swap(a, c);
      area = -area;
    }
    auto clamp = [](int v, int lo, int hi) { return std::max(lo, std::min(v, hi)); };
    int rx = std::min(drawR - 1, W - 1), by = std::min(drawB - 1, H - 1);
    if (drawL >= W || drawT >= H)
      return; // Invalid clip rectangles cannot address host memory.
    int minX = clamp(std::min({a.x, b.x, c.x}), drawL, rx),
        maxX = clamp(std::max({a.x, b.x, c.x}), drawL, rx),
        minY = clamp(std::min({a.y, b.y, c.y}), drawT, by),
        maxY = clamp(std::max({a.y, b.y, c.y}), drawT, by);
    for (int y = minY; y <= maxY; y++)
      for (int x = minX; x <= maxX; x++) {
        i64 w0 = edge(b.x, b.y, c.x, c.y, x, y), w1 = edge(c.x, c.y, a.x, a.y, x, y),
            w2 = edge(a.x, a.y, b.x, b.y, x, y);
        if (w0 < 0 || w1 < 0 || w2 < 0)
          continue;
        int r = int((w0 * a.r + w1 * b.r + w2 * c.r) / area),
            g = int((w0 * a.g + w1 * b.g + w2 * c.g) / area),
            bb = int((w0 * a.b + w1 * b.b + w2 * c.b) / area);
        u16 px;
        if (textured) {
          int u = int((w0 * a.u + w1 * b.u + w2 * c.u) / area),
              v = int((w0 * a.v + w1 * b.v + w2 * c.v) / area);
          u16 t = texel(u, v, clut);
          if (!t)
            continue;
          px = modulate(t, r, g, bb);
        } else
          px = (r >> 3) | ((g >> 3) << 5) | ((bb >> 3) << 10);
        store(x, y, px);
      }
  }
  void polygon(u8 op) {
    int nv = (op & 8) ? 4 : 3, idx = (op & 16) ? 0 : 1;
    bool shaded = op & 16, textured = op & 4;
    Vert vs[4];
    u32 clut = 0, tpage = 0;
    for (int i = 0; i < nv; i++) {
      u32 col = shaded ? fifo[idx++] : fifo[0], xy = fifo[idx++];
      auto &v = vs[i];
      v.r = col & 255;
      v.g = (col >> 8) & 255;
      v.b = (col >> 16) & 255;
      v.x = i16(xy) + offX;
      v.y = i16(xy >> 16) + offY;
      if (textured) {
        u32 uv = fifo[idx++];
        v.u = uv & 255;
        v.v = (uv >> 8) & 255;
        if (i == 0)
          clut = uv >> 16;
        if (i == 1)
          tpage = uv >> 16;
      }
    }
    if (textured) {
      texPageX = (tpage & 15) * 64;
      texPageY = ((tpage >> 4) & 1) * 256;
      texDepth = (tpage >> 7) & 3;
    }
    tri(vs[0], vs[1], vs[2], textured, clut);
    if (nv == 4)
      tri(vs[1], vs[2], vs[3], textured, clut);
  }
  void rect(u8 op) {
    u32 col = fifo[0], xy = fifo[1], clut = 0;
    int x0 = i16(xy) + offX, y0 = i16(xy >> 16) + offY, idx = 2, u0 = 0, v0 = 0, w = 1, h = 1;
    bool textured = op & 4;
    if (textured) {
      u32 uv = fifo[idx++];
      u0 = uv & 255;
      v0 = (uv >> 8) & 255;
      clut = uv >> 16;
    }
    switch (op & 24) {
    case 0:
      w = fifo[idx] & 65535;
      h = fifo[idx] >> 16;
      break;
    case 16:
      w = h = 8;
      break;
    case 24:
      w = h = 16;
      break;
    }
    u16 flat = color15(col);
    int left = std::max({x0, drawL, 0}), right = std::min({x0 + w, drawR, W}),
        top = std::max({y0, drawT, 0}), bottom = std::min({y0 + h, drawB, H});
    for (int py = top; py < bottom; py++)
      for (int px = left; px < right; px++) {
        u16 out = flat;
        if (textured) {
          out = texel(u0 + px - x0, v0 + py - y0, clut);
          if (!out)
            continue;
          if (!(op & 1))
            out = modulate(out, col & 255, (col >> 8) & 255, (col >> 16) & 255);
        }
        store(px, py, out);
      }
  }
  std::vector<u32> frame(bool draw = false) const {
    std::vector<u32> out(dispW * dispH);
    for (int y = 0; y < dispH; y++)
      for (int x = 0; x < dispW; x++)
        out[y * dispW + x] = rgba(
            vram[((y + (draw ? offY : dispY)) & 511) * W + ((x + (draw ? offX : dispX)) & 1023)]);
    return out;
  }
};
