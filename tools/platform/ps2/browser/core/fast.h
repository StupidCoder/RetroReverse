#pragma once

// A stream commonly repeats thousands of primitives with identical state.
// Retain exactly the reference census (including its labels) without formatting
// and hashing those labels again for every primitive. Nothing is deferred, so
// inspecting or saving between any two draws sees the same counts.
inline void ps2_GS_noteFeatures(ps2_GS*gs,uint64_t p) {
  auto&cache=gs->rrFeatures;
  if(cache.owner!=gs->drawCensus.p){
    cache.entries={};
    cache.owner=gs->drawCensus.p;
  }
  const unsigned context=(p>>9)&1;
  const std::array<uint64_t,6> key{
    p,gs->reg[6+context],gs->reg[0x47+context],
    gs->reg[0x4c + context],gs->reg[0x4e + context],0};
  auto&entry=cache.entries[context];
  if(entry.valid&&entry.key==key){
    for(unsigned i=0;i<entry.size;i++)++*entry.counts[i];
    return;
  }
  entry={};entry.key=key;cache.recording=&entry;
  try {ps2_GS_noteFeatures_reference(gs,p);}
  catch(...){cache.recording=nullptr;throw;}
  cache.recording=nullptr;cache.owner=gs->drawCensus.p;entry.valid=true;
}

// The within-page swizzle never changes. Resolve its several table lookups
// once, leaving page selection and one offset lookup per texture sample.
inline uint32_t ps2_addrPSMT8(uint32_t bp,uint32_t bw,uint32_t x,uint32_t y) {
  static const auto offsets=[]{
    std::array<uint16_t,128*64> a{};
    for(unsigned y=0;y<64;y++)for(unsigned x=0;x<128;x++)
      a[y*128+x]=ps2_addrPSMT8_reference(0,2,x,y);
    return a;
  }();
  const uint32_t page=(y>>6)*std::max(1u,bw>>1)+(x>>7);
  return bp*256+page*8192+offsets[(y&63)*128+(x&127)];
}
inline std::tuple<uint32_t,uint32_t> ps2_addrPSMT4(uint32_t bp,uint32_t bw,uint32_t x,uint32_t y) {
  static const auto offsets=[]{
    std::array<uint16_t,128*128> a{};
    for(unsigned y=0;y<128;y++)for(unsigned x=0;x<128;x++){
      auto[byte,nibble]=ps2_addrPSMT4_reference(0,2,x,y);
      a[y*128+x]=byte*2+nibble;
    }
    return a;
  }();
  const uint32_t page=(y>>7)*std::max(1u,bw>>1)+(x>>7);
  const uint32_t offset=offsets[(y&127)*128+(x&127)];
  return {bp*256+page*8192+(offset>>1),offset&1};
}

inline uint32_t rrGSPage32Offset(uint32_t x,uint32_t y){
  static const auto offsets=[]{
    std::array<uint16_t,64*32> a{};
    for(unsigned y=0;y<32;y++)for(unsigned x=0;x<64;x++)
      a[y*64+x]=ps2_addrPSMCT32_reference(0,1,x,y);
    return a;
  }();
  return offsets[(y&31)*64+(x&63)];
}
inline uint32_t ps2_addrPSMCT32(uint32_t bp,uint32_t bw,uint32_t x,uint32_t y){
  uint32_t page=(y>>5)*std::max(1u,bw)+(x>>6);
  return bp*256+page*8192+rrGSPage32Offset(x,y);
}
inline uint32_t ps2_addrPSMZ32(uint32_t bp,uint32_t bw,uint32_t x,uint32_t y){
  uint32_t page=(y>>5)*std::max(1u,bw)+(x>>6);
  return bp*256+page*8192+(rrGSPage32Offset(x,y)^0x1800);
}

// Keep diagnostic formatting out of the normal sampling function so LLVM can
// inline its small paths. Only addresses are cached: every texel and palette
// value is still read live, including when textures alias the render target.
inline uint32_t ps2_gsSampler_at(ps2_gsSampler*s,int32_t u,int32_t v){
  if(s->probe)return ps2_gsSampler_at_reference(s,u,v);
  uint32_t x=ps2_wrapTexel(u,s->w,s->wms,s->minu,s->maxu),
           y=ps2_wrapTexel(v,s->h,s->wmt,s->minv,s->maxv);
  auto*g=s->gs;const auto&t=s->tex;auto*ram=g->vram.p;
  uint32_t a=0,idx=0;
  switch(t.psm){
    case 0:case 1:case 0x1b:case 0x24:case 0x2c:{
      a=ps2_addrPSMCT32(t.tbp,t.tbw,x,y);
      if(uint64_t(a)+4>uint64_t(g->vram.n))return ps2_gsSampler_at_reference(s,u,v);
      uint32_t raw;std::memcpy(&raw,ram+a,4);
      if constexpr(std::endian::native==std::endian::big)raw=__builtin_bswap32(raw);
      if(t.psm==0)return raw;
      if(t.psm==1){
        uint32_t rgb=raw&0xffffff;uint64_t texa=g->reg[ps2_gsTEXA];
        uint32_t alpha=(texa&(1<<15))&&rgb==0?0:uint32_t(texa)&255;
        return rgb|(alpha<<24);
      }
      idx=raw>>24;
      if(t.psm==0x24)idx&=15;
      if(t.psm==0x2c)idx>>=4;
      break;
    }
    case 2:case 0xa:
      a=ps2_addrPSMCT16(t.tbp,t.tbw,x,y,t.psm==0xa);
      if(uint64_t(a)+2>uint64_t(g->vram.n))return ps2_gsSampler_at_reference(s,u,v);
      return ps2_GS_expand16(g,uint32_t(ram[a])|(uint32_t(ram[a+1])<<8));
    case 0x13:
      a=ps2_addrPSMT8(t.tbp,t.tbw,x,y);
      if(a>=uint64_t(g->vram.n))return 0;
      idx=ram[a];break;
    case 0x14:{
      auto[byte,nibble]=ps2_addrPSMT4(t.tbp,t.tbw,x,y);
      if(byte>=uint64_t(g->vram.n))return 0;
      idx=(ram[byte]>>(4*nibble))&15;break;
    }
    default:return ps2_gsSampler_at_reference(s,u,v);
  }
  uint32_t entry=ps2_GS_clutEntry(g,&s->tex,idx);
  if(t.psm==0x13&&!(entry&0xffffff)&&!s->parallel&&g->t8Dumped<16)
    return ps2_gsSampler_at_reference(s,u,v);
  return entry;
}
