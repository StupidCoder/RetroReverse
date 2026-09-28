#include "amiga.h"
namespace rramiga {
void Machine::spriteLine() {
  if ((dma & 0x220) != 0x220)
    return;
  for (unsigned i = 0; i < 8; i++) {
    auto &s = sprites[i];
    unsigned end = (s.ctl >> 8) | ((s.ctl & 2) << 7);
    if (s.header && line == end)
      s.header = false;
    if (!s.header) {
      s.pos = chip16(s.pointer);
      s.ctl = chip16(s.pointer + 2);
      s.pointer += 4;
      s.header = true;
    }
    unsigned start = (s.pos >> 8) | ((s.ctl & 4) << 6);
    end = (s.ctl >> 8) | ((s.ctl & 2) << 7);
    s.active = line >= start && line < end;
    if (s.active) {
      s.a = chip16(s.pointer);
      s.b = chip16(s.pointer + 2);
      s.pointer += 4;
    }
  }
}
void Machine::renderLine() {
  rrprof::Scope measured(1, "Bitplane and sprite scanout");
  auto &t = rrcapture::trace;
  unsigned start = reg[0x8e / 2] >> 8, stop = reg[0x90 / 2] >> 8;
  if (!(stop & 128))
    stop += 256;
  if (line == start)
    displayY = true;
  if (line == stop)
    displayY = false;
  bool enabled = displayY && (dma & 0x300) == 0x300;
  unsigned planes = std::min<unsigned>(6, (reg[0x100 / 2] >> 12) & 7);
  bool hires = reg[0x100 / 2] & 0x8000;
  unsigned ddfStart = reg[0x92 / 2] & 0xf8, ddfStop = reg[0x94 / 2] & 0xf8;
  unsigned words = ddfStop >= ddfStart ? (ddfStop - ddfStart) / 8 + 1 : 0;
  if (hires)
    words *= 2;
  words = std::min(words, 64u);
  std::array<std::array<uint16_t, 64>, 6> data{};
  auto pointers = bplPointer;
  if (enabled) {
    for (unsigned p = 0; p < planes; p++) {
      for (unsigned x = 0; x < words; x++)
        data[p][x] = chip16(bplPointer[p] + x * 2);
      bplPointer[p] += words * 2 + int16_t(reg[(p & 1 ? 0x10a : 0x108) / 2]);
      setPtr(0xe0 + p * 4, bplPointer[p]);
    }
  }
  int y = int(line) - 44;
  if (y < 0 || y >= int(H))
    return;
  if (y == 0) {
    draw.fill(color(0));
    if (t.active) {
      auto e =
          t.event(cycles, pc,
                  "{\"kind\":\"Display frame start\",\"origin\":\"display\"}");
      for (unsigned i = 0; i < W * H; i++)
        captureWrite(frameBase + i * 4, draw[i], 4, cycles, pc, e);
    }
  }
  uint32_t backgroundEvent = 0;
  if (t.active) {
    std::ostringstream s;
    s << "{\"kind\":\"Bitplane scanline\",\"origin\":\"display\",\"scanline\":"
      << line << ",\"bplcon0\":" << reg[0x100 / 2]
      << ",\"scroll\":" << reg[0x102 / 2] << ",\"fetchStart\":" << ddfStart
      << ",\"hasSource\":" << (enabled && planes ? "true" : "false")
      << ",\"planePointers\":[";
    for (unsigned p = 0; p < planes; p++) {
      if (p)
        s << ',';
      s << (pointers[p] & (CHIP - 1));
    }
    s << "],\"planeWords\":[";
    for (unsigned p = 0; p < planes; p++) {
      if (p)
        s << ',';
      s << '[';
      for (unsigned w = 0; w < words; w++) {
        if (w)
          s << ',';
        s << data[p][w];
      }
      s << ']';
    }
    s << "]}";
    backgroundEvent = t.event(cycles, pc, s.str());
  }
  int x0 = int(reg[0x8e / 2] & 255) - 0x81,
      x1 = int(reg[0x90 / 2] & 255) + 256 - 0x81;
  bool dual = reg[0x100 / 2] & 0x400, ham = reg[0x100 / 2] & 0x800;
  uint32_t held = color(0);
  std::array<uint8_t, W> playfield{}, claimed{};
  for (unsigned x = 0; x < W; x++) {
    int lowX = x / 2;
    unsigned index = 0, pf = 0;
    uint32_t rgba = color(0);
    int sample = -1;
    if (enabled && lowX >= x0 && lowX < x1) {
      for (unsigned p = 0; p < planes; p++) {
        int pos = (hires ? int(x) : lowX) -
                  (int(ddfStart) - 0x38) * (hires ? 4 : 2) -
                  ((reg[0x102 / 2] >> (p & 1 ? 4 : 0)) & 15);
        if (!p)
          sample = pos;
        if (pos >= 0 && unsigned(pos) < words * 16)
          index |= ((data[p][pos / 16] >> (15 - (pos & 15))) & 1) << p;
      }
      if (dual) {
        unsigned a = (index & 1) | ((index >> 1) & 2) | ((index >> 2) & 4),
                 b = ((index >> 1) & 1) | ((index >> 2) & 2) |
                     ((index >> 3) & 4);
        bool second = b && (!a || (reg[0x104 / 2] & 64));
        index = second ? b + 8 : a;
        pf = second ? 2 : (a ? 1 : 0);
        rgba = color(index);
      } else if (ham && planes == 6) {
        unsigned nib = (index & 15) * 17;
        switch (index >> 4) {
        case 0:
          held = color(index);
          break;
        case 1:
          held = (held & 0xff00ffff) | (nib << 16);
          break;
        case 2:
          held = (held & 0xffffff00) | nib;
          break;
        case 3:
          held = (held & 0xffff00ff) | (nib << 8);
          break;
        }
        rgba = held;
        pf = index ? 1 : 0;
      } else {
        rgba = color(index);
        if (planes == 6 && (index & 32))
          rgba = 0xff000000 | ((rgba & 0xfefefe) >> 1);
        pf = index ? 1 : 0;
      }
    }
    if (t.active) {
      bool valid = sample >= 0 && unsigned(sample) < words * 16;
      t.source = valid ? (pointers[0] + (sample / 16) * 2) & (CHIP - 1) : 0;
      t.sourceValue = valid ? swapped(data[0][sample / 16]) : 0;
      t.palette = registerBase + 0x180 + (index & 31) * 2;
      t.paletteValue = swapped(reg[(0x180 / 2) + (index & 31)]);
      t.texel = index;
      t.u = x;
      t.v = y;
      t.sourceBefore = t.paletteBefore = t.writes.size();
      t.record(frameBase + (y * W + x) * 4, rgba, 4, cycles, pc,
               backgroundEvent);
    }
    draw[y * W + x] = rgba;
    playfield[x] = pf;
  }
  if ((dma & 0x220) == 0x220) {
    for (unsigned n = 0; n < 8; n++) {
      auto &s = sprites[n];
      if (!s.active)
        continue;
      bool attached =
          (n % 2 == 0) && sprites[n + 1].active && (sprites[n + 1].ctl & 128);
      int sx = int(s.pos & 255) * 2 + (s.ctl & 1) - 0x81;
      uint32_t event = 0;
      if (t.active) {
        std::ostringstream o;
        o << "{\"kind\":\"Sprite scanline\",\"origin\":\"sprite\",\"scanline\":"
          << line << ",\"sprite\":" << n
          << ",\"attached\":" << (attached ? "true" : "false")
          << ",\"hasSource\":true}";
        event = t.event(cycles, pc, o.str());
      }
      for (int i = 0; i < 32; i++) {
        int x = sx * 2 + i;
        if (x < 0 || x >= int(W))
          continue;
        int bit = 15 - i / 2;
        unsigned value = ((s.a >> bit) & 1) | (((s.b >> bit) & 1) << 1),
                 palette = 16 + (n / 2) * 4 + value;
        if (attached) {
          auto &other = sprites[n + 1];
          value |= ((other.a >> bit) & 1) << 2 | ((other.b >> bit) & 1) << 3;
          palette = 16 + value;
        }
        unsigned pf = playfield[x], threshold = pf == 2
                                                    ? (reg[0x104 / 2] >> 3) & 7
                                                    : reg[0x104 / 2] & 7;
        uint8_t flags = 1;
        if (!value)
          flags = 4;
        else if (claimed[x])
          flags = 8;
        else {
          claimed[x] = 1;
          if (pf && n / 2 >= threshold)
            flags = 2;
        }
        auto rgba = color(palette);
        if (t.active) {
          t.source = (s.pointer - 4) & (CHIP - 1);
          t.sourceValue = swapped(s.a) | (uint32_t(swapped(s.b)) << 16);
          t.palette = registerBase + 0x180 + palette * 2;
          t.paletteValue = swapped(reg[0x180 / 2 + palette]);
          t.sourceBefore = t.paletteBefore = t.writes.size();
          t.texel = value;
          t.u = i / 2;
          t.v = y;
          t.record(frameBase + (y * W + x) * 4, rgba, 4, cycles, pc, event,
                   flags);
        }
        if (flags & 1)
          draw[y * W + x] = rgba;
      }
      if (attached)
        n++;
    }
  }
}
} // namespace rramiga
