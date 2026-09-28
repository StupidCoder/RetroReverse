#include "amiga.h"
namespace rramiga {
uint8_t Machine::ciaRead(unsigned n, unsigned r) {
  auto &c = cia[n];
  switch (r) {
  case 0: {
    uint8_t input = 255;
    if (n == 0) {
      if (selected) {
        if (motor)
          input &= ~32;
        if (cylinder == 0)
          input &= ~16;
        input &= ~8;
        if (diskChanged)
          input &= ~4;
      }
      if (mouseLeft)
        input &= ~64;
      if (buttons & 16)
        input &= ~128;
    }
    return (c.r[0] & c.r[2]) | (input & ~c.r[2]);
  }
  case 1:
    return (c.r[1] & c.r[3]) | ~c.r[3];
  case 4:
    return c.ta;
  case 5:
    return c.ta >> 8;
  case 6:
    return c.tb;
  case 7:
    return c.tb >> 8;
  case 8:
  case 9:
  case 10: {
    if (r == 10 && !c.todLatch) {
      c.latch = c.tod;
      c.todLatch = true;
    }
    auto v = (c.todLatch ? c.latch : c.tod) >> ((r - 8) * 8);
    if (r == 8)
      c.todLatch = false;
    return v;
  }
  case 13: {
    auto v = c.pending | ((c.pending & c.mask) ? 128 : 0);
    c.pending = 0;
    return v;
  }
  default:
    return c.r[r];
  }
}
void Machine::ciaWrite(unsigned n, unsigned r, uint8_t v) {
  auto &c = cia[n];
  switch (r) {
  case 0:
  case 2:
    c.r[r] = v;
    if (n == 0)
      overlay = (c.r[0] | ~c.r[2]) & 1;
    break;
  case 1:
  case 3:
    c.r[r] = v;
    if (n == 1)
      diskPort(c.r[1] | ~c.r[3]);
    break;
  case 4:
    c.la = (c.la & 0xff00) | v;
    break;
  case 5:
    c.la = (c.la & 255) | (v << 8);
    if (!(c.r[14] & 1))
      c.ta = c.la;
    if (c.r[14] & 8) {
      c.ta = c.la;
      c.r[14] |= 1;
    }
    break;
  case 6:
    c.lb = (c.lb & 0xff00) | v;
    break;
  case 7:
    c.lb = (c.lb & 255) | (v << 8);
    if (!(c.r[15] & 1))
      c.tb = c.lb;
    if (c.r[15] & 8) {
      c.tb = c.lb;
      c.r[15] |= 1;
    }
    break;
  case 8:
  case 9:
  case 10: {
    auto &counter = c.r[15] & 128 ? c.alarm : c.tod;
    counter =
        (counter & ~(255u << ((r - 8) * 8))) | (uint32_t(v) << ((r - 8) * 8));
    if (!(c.r[15] & 128))
      c.todStop = r != 8;
    break;
  }
  case 13:
    if (v & 128)
      c.mask |= v & 31;
    else
      c.mask &= ~v;
    break;
  case 14:
  case 15:
    if (v & 16) {
      if (r == 14)
        c.ta = c.la;
      else
        c.tb = c.lb;
    }
    c.r[r] = v & ~16;
    break;
  default:
    c.r[r] = v;
    break;
  }
  if (c.pending & c.mask)
    request(n ? 0x2000 : 8);
}
void Machine::ciaTick() {
  for (unsigned n = 0; n < 2; n++) {
    auto &c = cia[n];
    bool underA = false;
    if ((c.r[14] & 0x21) == 1) {
      if (c.ta == 0) {
        c.ta = c.la;
        c.pending |= 1;
        underA = true;
        if (c.r[14] & 8)
          c.r[14] &= ~1;
      } else
        c.ta--;
    }
    unsigned mode = (c.r[15] >> 5) & 3;
    if ((c.r[15] & 1) && (mode == 0 || (mode >= 2 && underA))) {
      if (c.tb == 0) {
        c.tb = c.lb;
        c.pending |= 2;
        if (c.r[15] & 8)
          c.r[15] &= ~1;
      } else
        c.tb--;
    }
    if (c.pending & c.mask)
      request(n ? 0x2000 : 8);
  }
  if (keyboardWait > 0)
    keyboardWait--;
  if (!keyboardWait && !keys.empty() && !(cia[0].r[14] & 64)) {
    cia[0].r[12] = keys.front();
    keys.pop_front();
    cia[0].pending |= 8;
    keyboardWait = 20000;
    if (cia[0].mask & 8)
      request(8);
  }
}
void Machine::todTick(unsigned n) {
  auto &c = cia[n];
  if (!c.todStop) {
    c.tod = (c.tod + 1) & 0xffffff;
    if (c.tod == c.alarm) {
      c.pending |= 4;
      if (c.mask & 4)
        request(n ? 0x2000 : 8);
    }
  }
}
void Machine::key(uint8_t raw, bool down) {
  uint8_t code = raw | (down ? 0 : 128);
  keys.push_back(uint8_t(~((code << 1) | (code >> 7))));
}
void Machine::diskPort(uint8_t v) {
  bool next = !(v & 8);
  if (next && !selected)
    motor = !(v & 128);
  if (next && (diskControl & 1) && !(v & 1)) {
    cylinder = std::clamp(cylinder + (v & 2 ? -1 : 1), 0, 79);
    diskChanged = false;
  }
  selected = next;
  side = (v & 4) ? 0 : 1;
  diskControl = v;
}
static void put32(std::array<uint16_t, 544> &a, unsigned at, uint32_t v) {
  a[at] = v >> 16;
  a[at + 1] = v;
}
void Machine::prepareTrack() {
  rrprof::Scope measured(4, "Floppy MFM encoding");
  loadedTrack = cylinder * 2 + side;
  track.assign(6334, 0xaaaa);
  unsigned out = 166;
  for (unsigned sector = 0; sector < 11; sector++) {
    std::array<uint16_t, 544> a{};
    a[0] = a[1] = 0xaaaa;
    a[2] = a[3] = 0x4489;
    uint32_t info =
        0xff000000 | (loadedTrack << 16) | (sector << 8) | (11 - sector);
    put32(a, 4, (info >> 1) & 0x55555555);
    put32(a, 6, info & 0x55555555);
    uint32_t header = 0, data = 0;
    for (unsigned j = 4; j < 24; j += 2)
      header ^= (uint32_t(a[j]) << 16) | a[j + 1];
    put32(a, 24, (header >> 1) & 0x55555555);
    put32(a, 26, header & 0x55555555);
    for (unsigned i = 0; i < 128; i++) {
      unsigned offset = (loadedTrack * 11 + sector) * 512 + i * 4;
      uint32_t v = uint32_t(disk[offset]) << 24 |
                   uint32_t(disk[offset + 1]) << 16 |
                   uint32_t(disk[offset + 2]) << 8 | disk[offset + 3];
      uint32_t odd = (v >> 1) & 0x55555555, even = v & 0x55555555;
      put32(a, 32 + i * 2, odd);
      put32(a, 288 + i * 2, even);
      data ^= odd ^ even;
    }
    put32(a, 28, (data >> 1) & 0x55555555);
    put32(a, 30, data & 0x55555555);
    unsigned previous = 1;
    for (unsigned i = 4; i < 544; i++) {
      uint16_t v = a[i] & 0x5555;
      for (int bit = 14; bit >= 0; bit -= 2) {
        unsigned b = (v >> bit) & 1;
        if (!(previous | b))
          v |= 1u << (bit + 1);
        previous = b;
      }
      a[i] = v;
    }
    std::copy(a.begin(), a.end(), track.begin() + out);
    out += 544;
  }
  trackIndex %= track.size();
}
void Machine::diskTick() {
  if (!motor || !selected)
    return;
  diskClock += 2;
  if (diskClock < 224)
    return;
  diskClock -= 224;
  if (loadedTrack != cylinder * 2 + side)
    prepareTrack();
  uint16_t word = track[trackIndex++];
  if (trackIndex >= track.size()) {
    trackIndex = 0;
    cia[1].pending |= 16;
    if (cia[1].mask & 16)
      request(0x2000);
  }
  bool matched = word == reg[0x7e / 2];
  if (matched)
    request(0x1000);
  if (!diskRemaining || (dma & 0x210) != 0x210)
    return;
  if (!diskSync) {
    if (matched)
      diskSync = true;
    return;
  }
  auto previous = hardwareEvent;
  if (rrcapture::trace.active) {
    if (!diskEvent)
      diskEvent = rrcapture::trace.event(
          cycles, pc,
          "{\"kind\":\"Floppy DMA\",\"origin\":\"disk\",\"track\":" +
              std::to_string(loadedTrack) + "}");
    hardwareEvent = diskEvent;
  }
  chipWrite(diskPtr, word);
  diskPtr += 2;
  setPtr(0x20, diskPtr);
  hardwareEvent = previous;
  if (--diskRemaining == 0) {
    diskEvent = 0;
    diskReads++;
    request(2);
    if (logging)
      std::fprintf(stderr, "disk read %llu track %d PC=%06x\n",
                   (unsigned long long)diskReads, loadedTrack, pc);
  }
}
} // namespace rramiga
