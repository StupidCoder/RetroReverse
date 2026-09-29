#pragma once
#include "../../../../browser/core/replay.h"

namespace rrdos {
inline std::vector<uint8_t> compactMemory(const std::vector<uint8_t> &video,
                                        const uint8_t *ram) {
  size_t n = regions.empty() ? VIDEO_BYTES : regions.back().packed + regions.back().size;
  std::vector<uint8_t> out(n);
  std::copy(video.begin(), video.end(), out.begin());
  for (const auto &r : regions) memcpy(out.data() + r.packed, ram + r.guest, r.size);
  return out;
}
inline void finish(const Slice<uint8_t> &ram, const std::vector<uint8_t> &video) {
  collecting = false; rawCount = raw.size();
  // The chipset's write path has already resolved VGA plane selection.
  std::vector<Pixel> origins(0x40000);
  for (uint32_t i = 0; i < raw.size(); ++i) {
    const auto &w = raw[i];
    if (!(w.address & VIDEO)) continue;
    for (unsigned b = 0; b < w.size; ++b) {
      auto a = (w.address & ~VIDEO) + b;
      if (a >= 0xa0000 && a < 0xe0000)
        origins[a - 0xa0000] = {w.source == NONE ? NONE : w.source + b, i, 0};
    }
  }
  // Track exact contributing bytes; pages are only a compact storage unit.
  // Backward MOVS edges also retain intermediate copies' source histories.
  std::vector<uint64_t> needed((ram.n + 63) / 64);
  std::vector<bool> pages((ram.n + 4095) / 4096);
  size_t pageCount = 0;
  auto has = [&](uint32_t a) { return a < uint32_t(ram.n) && (needed[a / 64] & (1ull << (a % 64))); };
  auto mark = [&](uint32_t a) {
    if (a >= uint32_t(ram.n) || (a >= 0xa0000 && a < 0xb0000)) return;
    auto page = a / 4096;
    if (!pages[page]) {
      if (pageCount == PAGE_CAP) { ++dropped; return; }
      pages[page] = true; ++pageCount;
    }
    needed[a / 64] |= 1ull << (a % 64);
  };
  pixels.fill(Pixel{});
  for (int y = 0; y < 200; ++y) for (int x = 0; x < 320; ++x) {
    auto a = rrAddress(video, x, y);
    auto &p = pixels[y * 320 + x];
    p = origins[a - 0xa0000];
    if (p.source != NONE) mark(p.source);
  }
  for (auto it = raw.rbegin(); it != raw.rend(); ++it) {
    if (it->address & VIDEO || it->source == NONE) continue;
    for (unsigned b = 0; b < it->size; ++b) if (has(it->address + b)) mark(it->source + b);
  }
  regions.clear(); uint32_t next = PACKED_RAM;
  for (uint32_t i = 0; i < pages.size(); ++i) if (pages[i]) {
    uint32_t size = std::min<uint32_t>(4096, ram.n - i * 4096);
    if (!regions.empty() && regions.back().guest + regions.back().size == i * 4096)
      regions.back().size += size;
    else regions.push_back({i * 4096, next, size});
    next += size;
  }
  auto initial = compactMemory(initialVideo, initialRAM.data());
  auto &t = rrcapture::trace;
  t.begin(initial.data(), initial.size());
  // prefix[i] counts replay writes before raw write i.
  std::vector<uint32_t> prefix(raw.size() + 1);
  uint32_t event = 0; uint64_t eventClock = UINT64_MAX; uint32_t eventPC = NONE;
  for (uint32_t i = 0; i < raw.size(); ++i) {
    prefix[i] = t.writes.size(); const auto &w = raw[i];
    bool videoWrite = w.address & VIDEO;
    for (unsigned b = 0; b < w.size;) {
      if (!videoWrite && !has(w.address + b)) { ++b; continue; }
      auto first = b++;
      while (b < w.size && (videoWrite || has(w.address + b))) ++b;
      auto size = b - first;
      auto a = videoWrite ? (w.address & ~VIDEO) + first : packed(w.address + first);
      if (a == NONE) continue;
      if (videoWrite && (eventClock != w.clock || eventPC != w.pc)) {
        eventClock = w.clock; eventPC = w.pc;
        event = t.event(w.clock, w.pc, w.source == NONE ?
          "{\"kind\":\"x86 VGA write\"}" : "{\"kind\":\"MOVS to VGA\"}");
      }
      auto source = w.source == NONE ? NONE : packed(w.source + first);
      t.source = source == NONE ? 0 : source;
      t.sourceBefore = prefix[std::min(w.readBefore, i)];
      t.sourceValue = w.value >> (first * 8);
      if (size < 4) t.sourceValue &= (1u << (size * 8)) - 1;
      // RAM writes have no GPU event: every CPU instruction is a replay step.
      t.record(a, t.sourceValue, size, w.clock, w.pc, videoWrite ? event : 0);
    }
  }
  prefix.back() = t.writes.size();
  producerPixels = 0;
  for (auto &p : pixels) {
    if (p.source == NONE || packed(p.source) == NONE || p.copy == NONE) { p.source = NONE; continue; }
    p.write = prefix[p.copy + 1];
    if (p.write > prefix[p.copy]) ++producerPixels;
    else p.source = NONE;
  }
  auto final = compactMemory(video, ram.p);
  t.end(final.data(), final.size()); t.overflow += dropped;
  std::vector<Write>().swap(raw);
  std::vector<uint8_t>().swap(initialRAM);
  std::vector<uint8_t>().swap(initialVideo);
}
inline std::vector<uint8_t> replayFrame() {
  auto &r = rrreplay::replay;
  auto out = rrFrame(r.memory);
  std::vector<uint8_t> touched;
  if (reveal) {
    touched.resize(r.memory.size());
    for (uint32_t i = 0; i < r.writeCursor; ++i) {
      const auto &w = rrcapture::trace.writes[i];
      for (unsigned b = 0; b < w.size; ++b) touched[w.address + b] = 1;
    }
  }
  for (int y = 0; y < 200; ++y) for (int x = 0; x < 320; ++x) {
    const auto &p = pixels[y * 320 + x];
    if (p.source == NONE || r.writeCursor >= p.write) continue;
    // Project the final blit's source into screen coordinates until presentation.
    // Freeze at the copy so later RAM reuse cannot change the displayed pixel.
    auto a = packed(p.source); auto pi = r.memory[a];
    for (int k = 0; k < 3; ++k) out[(y * 320 + x) * 4 + k] = (r.memory[0xe0000 + pi * 3 + k] << 2) / (reveal && !touched[a] ? 5 : 1);
  }
  return out;
}
} // namespace rrdos

extern "C" {
int rr_capture_begin(){try{
  auto &ram = realMachine ? realMachine->Mem : protectedMachine->Mem;
  rrdos::initialRAM.assign(ram.p, ram.p + ram.n); rrdos::initialVideo = rrVideo();
  rrcapture::trace.begin(rrdos::initialVideo.data(), rrdos::initialVideo.size());
  rrdos::raw.clear(); rrdos::regions.clear(); rrdos::pixels.fill(rrdos::Pixel{});
  rrdos::bursts = rrdos::dropped = rrdos::producerPixels = rrdos::rawCount = 0; rrdos::reveal = false;
  rrdos::startClock = rrCPU()->Steps; rrdos::inBurst = rrdos::timedOut = false;
  rrdos::copySize = 0; rrdos::copySource = rrdos::NONE; rrdos::collecting = true;
  return 1;
}catch(const std::exception&e){errorText=e.what();return 0;}}
int rr_capture_end(){try{
  rrdos::finish(realMachine ? realMachine->Mem : protectedMachine->Mem, rrVideo());return 1;
}catch(const std::exception&e){rrdos::collecting=false;rrcapture::trace.active=false;errorText=e.what();return 0;}}
const char*rr_capture_info(){
  rrdos::observe(rrCPU()->Steps);
  if(rrdos::collecting)rrdos::timedOut = rrCPU()->Steps-rrdos::startClock >= rrFramePeriod()*140;
  reply=rrcapture::trace.info();reply.pop_back();
  reply+=",\"ready\":"+std::string(rrdos::bursts>=2||rrdos::timedOut||rrdos::dropped?"true":"false")+
    ",\"vgaWriteWindow\":true,\"renderBuffers\":true,\"producerPixels\":"+std::to_string(rrdos::producerPixels)+
    ",\"bursts\":"+std::to_string(rrdos::bursts)+",\"timedOut\":"+(rrdos::timedOut?"true":"false")+
    ",\"rawWrites\":"+std::to_string(rrdos::collecting ? rrdos::raw.size() : rrdos::rawCount)+",\"memoryRegions\":[";
  bool comma=false;for(const auto&r:rrdos::regions){if(comma)reply+=',';comma=true;reply+="{\"guest\":"+std::to_string(r.guest)+",\"packed\":"+std::to_string(r.packed)+",\"size\":"+std::to_string(r.size)+"}";}
  reply+="]}";return reply.c_str();
}
const char*rr_pixel(int x,int y){
  auto&t=rrcapture::trace;if(!t.valid||x<0||y<0||x>=320||y>=200){reply="{\"error\":\"No captured pixel\"}";return reply.c_str();}
  auto vga=rrAddress(t.final,x,y);const auto&p=rrdos::pixels[y*320+x];
  auto a=p.source==rrdos::NONE?vga:rrdos::packed(p.source);
  uint32_t before=p.source==rrdos::NONE?UINT32_MAX:p.write-1;
  reply=t.pixel(a,1,before,t.final[vga]);reply.pop_back();
  auto palette=0xe0000+t.final[vga]*3;
  reply+=",\"displayFormat\":\"VGA palette index\",\"vgaAddress\":"+std::to_string(vga)+
    ",\"producer\":"+(p.source==rrdos::NONE?std::string("false"):std::string("true"))+
    ",\"guestAddress\":"+std::to_string(rrdos::guest(a))+
    ",\"paletteAddress\":"+std::to_string(palette)+",\"displayRGBA\": ["+
    std::to_string(t.final[palette]<<2)+","+std::to_string(t.final[palette+1]<<2)+","+std::to_string(t.final[palette+2]<<2)+",255]}";
  return reply.c_str();
}
const char*rr_source(uint32_t a,int n,uint32_t before,uint32_t expected){reply=rrcapture::trace.pixel(a,n,before,expected);return reply.c_str();}
const char*rr_resource(uint32_t id,uint32_t off){reply=rrcapture::trace.resourceJSON(id,off);return reply.c_str();}
const char*rr_replay_info();
const char*rr_replay_begin(){rrreplay::replay.begin();return rr_replay_info();}
int rr_replay_seek(uint32_t n){return rrreplay::replay.seek(n);}
const char*rr_replay_info(){reply=rrreplay::replay.info();if(rrdos::producerPixels){reply.pop_back();reply+=",\"surface\":\"RAM render buffers mapped through the captured VGA copies\",\"writerPC\":true}";}return reply.c_str();}
void rr_replay_view(int reveal){rrdos::reveal = reveal != 0;}
uint32_t rr_replay_for_write(uint32_t id){return rrreplay::replay.forWrite(id);}
uint8_t*rr_replay_frame(){pixels=rrdos::replayFrame();return pixels.data();}
}
