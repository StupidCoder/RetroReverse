#pragma once
#include "generated.cpp"
#include "slice.h"
#include "movie.h"
inline RunContext runContext;
struct DiscChunk {
  int64_t base;
  Slice<uint8_t> bytes;
  uint64_t used;
};
inline uint64_t discSize = 0, discBytesRead = 0, discReads = 0, discClock = 0, totalSteps = 0;
inline std::function<void(uint64_t, uint8_t *, size_t)> discRead;
inline std::vector<DiscChunk> discCache;
inline Slice<uint8_t> readDisc(uint64_t offset, size_t n) {
  if (offset > discSize || n > discSize - offset)
    throw std::runtime_error("Disc read beyond image");
  auto b = Slice<uint8_t>::make(n);
  { rrprof::Scope clock(4,"Disc I/O"); discRead(offset, b.p, n); }
  discBytesRead += n;
  discReads++;
  return b;
}
std::tuple<Slice<uint8_t>, Error> threedo_Volume_block(threedo_Volume *v, int64_t n) {
  if (n < 0 || n >= v->nsect)
    return {{}, {"Disc sector out of range"}};
  auto base = (n / 1024) * 1024;
  DiscChunk *chunk = nullptr;
  for (auto &c : discCache)
    if (c.base == base) {
      chunk = &c;
      break;
    }
  if (!chunk) {
    auto size = std::min<int64_t>(1024, v->nsect - base) * v->stride;
    auto b = readDisc(base * v->stride, size);
    if (discCache.size() == 8) {
      chunk = &*std::min_element(discCache.begin(), discCache.end(),
                                 [](auto &a, auto &b) { return a.used < b.used; });
      *chunk = {base, b, 0};
    } else {
      discCache.push_back({base, b, 0});
      chunk = &discCache.back();
    }
  }
  chunk->used = ++discClock;
  auto off = (n - base) * v->stride + v->dataOff;
  return {sub(chunk->bytes, off, off + 2048), {}};
}
#include "fileio.h"
inline threedo_Volume *openDisc() {
  auto v = arenaNew(threedo_Volume{});
  auto head = readDisc(0, 32);
  if (discSize >= 2352 && discSize % 2352 == 0 && head[0] == 0 && head[1] == 255 && head[11] == 0) {
    v->stride = 2352;
    v->dataOff = head[15] == 1 ? 16 : head[15] == 2 ? 24 : 0;
    if (!v->dataOff)
      throw std::runtime_error("Unsupported CD sector mode");
  } else if (discSize >= 2048 && discSize % 2048 == 0) {
    v->stride = 2048;
  } else
    throw std::runtime_error("Invalid disc geometry");
  v->nsect = discSize / v->stride;
  auto [label, e] = threedo_Volume_block(v, 0);
  if (e)
    throw std::runtime_error(e.text);
  if (label[0] != 1 || std::string((char *)label.p + 1, 5) != "ZZZZZ")
    throw std::runtime_error("Opera volume label missing");
  v->Comment = threedo_trimName(sub(label, 8, 40));
  v->Label = threedo_trimName(sub(label, 40, 72));
  v->ID = be_Uint32(sub(label, 72, label.n));
  v->blockSize = be_Uint32(sub(label, 76, label.n));
  if (v->blockSize != 2048)
    throw std::runtime_error("Unsupported Opera block size");
  v->rootBlocks = be_Uint32(sub(label, 0x58, label.n));
  v->rootBlock = be_Uint32(sub(label, 0x64, label.n));
  return v;
}
std::tuple<threedo_Volume *, Error> threedo_Open(Slice<uint8_t> data) {
  discSize = data.n;
  discRead = [data](uint64_t o, uint8_t *p, size_t n) { std::memcpy(p, data.p + o, n); };
  try {
    return {openDisc(), {}};
  } catch (const std::exception &e) {
    return {{}, {e.what()}};
  }
}
inline threedo_Machine *boot(bool nfsProfile = true) {
  arenaClear();
  runContext = {};
  presentationSeconds=movieSeconds=0;
  fileEntries.clear();
  discCache.clear();
  discClock = discBytesRead = discReads = totalSteps = 0;
  auto vol = openDisc();
  auto [prog, err] = threedo_Volume_ReadFile(vol, "LaunchMe");
  if (err)
    throw std::runtime_error(err.text);
  auto [a, e] = threedo_ParseAIF(prog);
  if (e)
    throw std::runtime_error(e.text);
  auto m = threedo_NewMachine();
  m->PaceFields = true;
  m->StallTolerance = 8;
  m->MovieHLE = nfsProfile;
  m->NoStreams = false;
  threedo_Machine_SetVolume(m, vol);
  if (nfsProfile) threedo_Machine_SetVBLMirror(m, 0x42734);
  threedo_Machine_LoadAIF(m, a);
  m->OnDisplay = [](threedo_Machine *m, uint64_t, uint32_t) { m->StopRequested = true; };
  return m;
}
inline uint64_t runSlice(threedo_Machine *m, uint64_t budget) {
  rrprof::Scope clock(0,"ARM60 / scheduler remainder");
  if(presentMovie(m))return 0;
  const auto oldFrame=m->frame;
  // Normal execution keeps bounded diagnostic history; guest state is untouched.
  if(m->SWICalls.n>65536)m->SWICalls=sub(m->SWICalls,m->SWICalls.n-32768,m->SWICalls.n);
  if(m->KernelCalls.n>65536)m->KernelCalls=sub(m->KernelCalls,m->KernelCalls.n-32768,m->KernelCalls.n);
  rrSchedulerClockBase=totalSteps-runContext.steps;
  auto r=threedo_Machine_RunSlice(m,budget,runContext);
  totalSteps+=r.Steps;
  presentationSeconds+=(m->frame-oldFrame)/30.0;
  if(r.Reason!="step budget reached"&&r.Reason!="movie pending") {
    runContext={};
    if(r.Reason!="stop requested") throw std::runtime_error(r.Reason);
  }
  return r.Steps;
}
inline void nextFrame(threedo_Machine *m) {
  auto before=m->frame;uint64_t steps=0;
  while(m->frame==before){steps+=runSlice(m,10000);if(steps>400000000)throw std::runtime_error("No new display within budget");}
}
inline uint32_t hashBytes(const uint8_t *p, size_t n) {
  uint32_t h = 2166136261;
  for (size_t i = 0; i < n; i++)
    h = (h ^ p[i]) * 16777619;
  return h;
}
struct Hasher {
  uint32_t h = 2166136261;
  void add(uint64_t v) {
    for (int i = 0; i < 8; i++) {
      h = (h ^ uint8_t(v)) * 16777619;
      v >>= 8;
    }
  }
};
inline std::vector<uint8_t> frame(threedo_Machine *m) {
  std::vector<uint8_t> p(320 * 240 * 4);
  for (uint32_t y = 0; y < 240; y++)
    for (uint32_t x = 0; x < 320; x++) {
      uint32_t i = (y * 320 + x) * 4,
               a = m->displayBuf - 0x200000 + (y >> 1) * 320 * 4 + x * 4 + (y & 1) * 2;
      if (a + 1 >= m->vram.n)
        continue;
      p[i + 3] = 255;
      auto c = threedo_rgb555(uint16_t(m->vram[a]) << 8 | m->vram[a + 1]);
      p[i] = c.R;
      p[i + 1] = c.G;
      p[i + 2] = c.B;
    }
  return p;
}
inline std::string proof(threedo_Machine *m) {
  Hasher h;
  for (auto v : m->CPU->R)
    h.add(v);
  h.add(arm60_CPU_CPSR(m->CPU));
  h.add(m->CPU->Instrs);
  auto p = frame(m);
  std::ostringstream s;
  s << "{\"steps\":" << totalSteps << ",\"frame\":" << m->frame << ",\"pc\":" << m->CPU->R[15]
    << ",\"cpu\":" << h.h << ",\"dram\":" << hashBytes(m->dram.p, m->dram.n)
    << ",\"vram\":" << hashBytes(m->vram.p, m->vram.n)
    << ",\"imem\":" << hashBytes(m->imem.p, m->imem.n)
    << ",\"pixels\":" << hashBytes(p.data(), p.size()) << ",\"display\":" << m->displayBuf
    << ",\"instructions\":" << m->CPU->Instrs << "}";
  return s.str();
}
#ifndef __EMSCRIPTEN__
inline void nativeDisc(const char *path) {
  auto f = std::make_shared<std::ifstream>(path, std::ios::binary);
  if (!*f)
    throw std::runtime_error("Cannot open disc");
  f->seekg(0, std::ios::end);
  discSize = f->tellg();
  discRead = [f](uint64_t o, uint8_t *p, size_t n) {
    f->seekg(o);
    f->read((char *)p, n);
    if (size_t(f->gcount()) != n)
      throw std::runtime_error("Short disc read");
  };
}
#endif
