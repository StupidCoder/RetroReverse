#include "inspection.h"
namespace rramiga {
void Machine::copperTick() {
  if (copperStopped || (dma & 0x280) != 0x280)
    return;
  if (copperDelay > 0) {
    copperDelay--;
    return;
  }
  uint16_t a = chip16(copperPC), b = chip16(copperPC + 2);
  if (!(a & 1)) {
    unsigned r = a & 0x1fe;
    if (r < ((reg[0x2e / 2] & 2) ? 0x40 : 0x80)) {
      copperStopped = true;
      return;
    }
    rrprof::Scope measured(3, "Copper register moves");
    std::string detail;
    if (rrcapture::trace.active) {
      std::ostringstream s;
      s << "{\"kind\":\"Copper MOVE\",\"origin\":\"copper\",\"copperPC\":"
        << copperPC << ",\"scanline\":" << line << ",\"beam\":" << beam << ",\"register\":" << r << ",\"value\":" << b << "}";
      detail = s.str();
    }
    EvidenceScope scope(cycles, pc, std::move(detail));
    copperPC = (copperPC + 4) & 0x1ffffe;
    customWrite(r, b, true);
    copperMoves++;
    copperDelay = 3;
    return;
  }
  unsigned vp = line & 255, hp = beam / 2;
  unsigned mask = (b & 0x7ffe) | 0x8000;
  unsigned pos = (vp << 8) | (hp & 0xfe);
  bool ready = (pos & mask) >= (a & mask) && ((b & 0x8000) || !blitBusy);
  if (b & 1) {
    copperPC += ready ? 8 : 4;
    copperDelay = 3;
  } else if (ready) {
    copperPC += 4;
    copperDelay = 5;
  }
}
static uint16_t minterm(uint16_t a, uint16_t b, uint16_t c, unsigned op) {
  uint16_t v = 0;
  for (unsigned i = 0; i < 8; i++)
    if (op & (1 << i))
      v |= (i & 4 ? a : uint16_t(~a)) & (i & 2 ? b : uint16_t(~b)) &
           (i & 1 ? c : uint16_t(~c));
  return v;
}
void Machine::blitter() {
  rrprof::Scope measured(2, "Blitter");
  std::string detail;
  if (rrcapture::trace.active) {
    std::ostringstream s;
    s << "{\"kind\":\"Blitter operation\",\"origin\":\"blitter\",\"con0\":"
      << reg[0x40 / 2] << ",\"blit\":" << inspect::blits.size() << ",\"con1\":" << reg[0x42 / 2]
      << ",\"copperPC\":" << (!inspect::changes.empty()&&inspect::changes.back().reg==0x58?inspect::changes.back().copperPC:0)
      << ",\"size\":" << reg[0x58 / 2] << ",\"sourceA\":" << ptr(0x50)
      << ",\"sourceB\":" << ptr(0x4c) << ",\"sourceC\":" << ptr(0x48)
      << ",\"destination\":" << ptr(0x54) << "}";
    detail = s.str();
  }
  EvidenceScope evidence(cycles, pc, std::move(detail));
  unsigned con0 = reg[0x40 / 2], con1 = reg[0x42 / 2],
           width = reg[0x58 / 2] & 63, height = reg[0x58 / 2] >> 6;
  if (!width)
    width = 64;
  if (!height)
    height = 1024;
  int inspected = inspect::blitBegin(*this, width, height);
  unsigned ash = con0 >> 12, bsh = con1 >> 12;
  blits++;
  blitZero = true;
  uint32_t ap = ptr(0x50), bp = ptr(0x4c), cp = ptr(0x48), dp = ptr(0x54);
  int am = int16_t(reg[0x64 / 2]) & ~1, bm = int16_t(reg[0x62 / 2]) & ~1,
      cm = int16_t(reg[0x60 / 2]) & ~1, dm = int16_t(reg[0x66 / 2]) & ~1;
  uint16_t ad = reg[0x74 / 2], bd = reg[0x72 / 2], cd = reg[0x70 / 2];
  if (con1 & 1) {
    int error = int16_t(reg[0x52 / 2]), bit = ash;
    bool sign = con1 & 0x40, onedot = false;
    unsigned oct = (con1 >> 2) & 7;
    auto moveX = [&](int dx) {
      bit += dx;
      if (bit < 0) {
        bit += 16;
        cp -= 2;
        dp -= 2;
      }
      if (bit >= 16) {
        bit -= 16;
        cp += 2;
        dp += 2;
      }
    };
    auto moveY = [&](int dy) {
      cp += dy * cm;
      dp += dy * cm;
      onedot = false;
    };
    // Table 6-3: x-major octants are 6,7,5,4; y-major are 1,3,2,0.
    bool xMajor = oct >= 4;
    int dx = (oct == 3 || oct == 7 || oct == 5 || oct == 2) ? -1 : 1,
        dy = (oct == 6 || oct == 1 || oct == 3 || oct == 7) ? -1 : 1;
    for (unsigned y = 0; y < height; y++) {
      uint16_t a = ad >> bit, b = ((bd >> ((bsh + y) & 15)) & 1) ? 65535 : 0,
               c = con0 & 0x200 ? chip16(cp) : cd;
      inspect::BlitWord observed;
      if (inspected >= 0) { observed.address={0,0,cp & (CHIP-1),dp & (CHIP-1)}; observed.before=rrcapture::trace.writes.size(); observed.a=a; observed.b=b; observed.c=c; observed.rawA=ad; observed.rawB=bd; observed.oldD=chip16(dp); }
      uint16_t d = minterm(a, b, c, con0 & 255);
      if (!(con1 & 2) || !onedot) {
        if (con0 & 0x100)
          chipWrite(dp, d);
        if (d)
          blitZero = false;
      }
      if (inspected >= 0) { observed.d=d; observed.written=(con0&0x100)&&(!(con1&2)||!onedot); observed.write=observed.written?rrcapture::trace.writes.size():0; inspect::blitWord(inspected,observed); }
      onedot = true;
      bool diagonal = !sign;
      error += sign ? bm : am;
      sign = error < 0;
      if (diagonal) {
        if (xMajor)
          moveY(dy);
        else
          moveX(dx);
      }
      if (xMajor)
        moveX(dx);
      else
        moveY(dy);
    }
    reg[0x52 / 2] = error;
    reg[0x40 / 2] = (con0 & 0xfff) | (bit << 12);
    reg[0x42 / 2] = (con1 & ~0x40) | (sign ? 0x40 : 0);
    for (unsigned r : {0x52, 0x40, 0x42})
      captureWrite(registerBase + r, swapped(reg[r / 2]), 2, cycles, pc,
                   hardwareEvent);
  } else {
    bool desc = con1 & 2;
    int step = desc ? -2 : 2;
    uint16_t prevA = 0, prevB = 0;
    for (unsigned y = 0; y < height; y++) {
      bool carry = con1 & 4;
      for (unsigned x = 0; x < width; x++) {
        uint16_t a = con0 & 0x800 ? chip16(ap) : ad,
                 b = con0 & 0x400 ? chip16(bp) : bd,
                 c = con0 & 0x200 ? chip16(cp) : cd;
        inspect::BlitWord observed;
        if (inspected >= 0) { observed.address={ap & (CHIP-1),bp & (CHIP-1),cp & (CHIP-1),dp & (CHIP-1)}; observed.before=rrcapture::trace.writes.size(); observed.rawA=a; observed.rawB=b; observed.c=c; observed.oldD=chip16(dp); }
        if (x == 0)
          a &= reg[0x44 / 2];
        if (x + 1 == width)
          a &= reg[0x46 / 2];
        uint16_t sa = desc ? uint16_t((uint32_t(a) << ash) |
                                      (ash ? prevA >> (16 - ash) : 0))
                           : uint16_t((uint32_t(prevA) << 16 | a) >> ash);
        uint16_t sb = desc ? uint16_t((uint32_t(b) << bsh) |
                                      (bsh ? prevB >> (16 - bsh) : 0))
                           : uint16_t((uint32_t(prevB) << 16 | b) >> bsh);
        prevA = a;
        prevB = b;
        uint16_t d = minterm(sa, sb, c, con0 & 255);
        if (con1 & 0x18) {
          for (unsigned bit = 0; bit < 16; bit++) {
            bool original = d & (1 << bit);
            if (carry) {
              if (con1 & 0x10)
                d ^= 1 << bit;
              else
                d |= 1 << bit;
            }
            if (original)
              carry = !carry;
          }
        }
        if (inspected >= 0) { observed.a=sa; observed.b=sb; observed.d=d; }
        if (d)
          blitZero = false;
        if (con0 & 0x100) {
          chipWrite(dp, d);
          dp += step;
        }
        if (inspected >= 0) { observed.written=con0&0x100; observed.write=observed.written?rrcapture::trace.writes.size():0; inspect::blitWord(inspected,observed); }
        if (con0 & 0x800)
          ap += step;
        if (con0 & 0x400)
          bp += step;
        if (con0 & 0x200)
          cp += step;
      }
      if (con0 & 0x800)
        ap += desc ? -am : am;
      if (con0 & 0x400)
        bp += desc ? -bm : bm;
      if (con0 & 0x200)
        cp += desc ? -cm : cm;
      if (con0 & 0x100)
        dp += desc ? -dm : dm;
    }
  }
  setPtr(0x50, ap);
  setPtr(0x4c, bp);
  setPtr(0x48, cp);
  setPtr(0x54, dp);
  if (inspected >= 0) inspect::blits[inspected].end=rrcapture::trace.writes.size();
  blitBusy = true;
  blitDone = cycles + uint64_t(width) * height * 8;
}
} // namespace rramiga
