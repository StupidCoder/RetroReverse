#pragma once
// Borrow memory for a single bus operation. No shared_ptr traffic per byte, and
// no cached host pointers that could outlive CP15/WRAM remapping or a state load.
namespace rrdsfast {
struct Span { uint8_t* p=nullptr; uint64_t n=0; };
inline Span at(const Slice<uint8_t>&s,uint32_t i,uint64_t limit=UINT32_MAX) {
  return i<uint64_t(s.n)?Span{s.p+i,std::min(uint64_t(s.n)-i,limit)}:Span{};
}
inline Span memory(dsmachine_core*c,uint32_t a) {
  if(a>=0x02000000 && a<0x03000000) {
    if(c->dtcm && a>=c->dtcmBase && a<uint32_t(c->dtcmBase+c->dtcm.n))
      return at(c->dtcm,a-c->dtcmBase);
    auto off=(a-0x02000000)&0x3fffff;
    uint64_t limit=0x400000-off;
    if(c->dtcm && a<c->dtcmBase && c->dtcmBase<uint32_t(c->dtcmBase+c->dtcm.n))
      limit=std::min(limit,uint64_t(c->dtcmBase-a));
    return at(c->m->ram,off,limit);
  }
  if(a>=0x03000000 && a<0x03800000) {
    auto off=a-0x03000000;auto mode=c->m->wramcnt&3;
    if(c->arm9) {
      if(mode==3)return {};
      if(mode==0)return at(c->m->swram,off&32767,32768-(off&32767));
      return at(c->m->swram,(mode==1?16384:0)+(off&16383),16384-(off&16383));
    }
    if(mode==0)return at(c->wram7,a&65535,65536-(a&65535));
    if(mode==3)return at(c->m->swram,off&32767,32768-(off&32767));
    return at(c->m->swram,(mode==2?16384:0)+(off&16383),16384-(off&16383));
  }
  // Device dispatch has priority over TCM, as in the reference bus.
  if(a>>24==4 || a>>24==6)return {};
  if(c->arm9) {
    if(a>=c->itcmBase && a<uint32_t(c->itcmBase+c->itcm.n))
      return at(c->itcm,a-c->itcmBase,0x1000000-(a&0xffffff));
    uint64_t limit=2048-(a&2047);
    if(a<c->itcmBase && c->itcmBase<uint32_t(c->itcmBase+c->itcm.n))limit=std::min(limit,uint64_t(c->itcmBase-a));
    if(a>>24==5)return at(c->m->pal,a&2047,limit);
    if(a>>24==7)return at(c->m->oam,a&2047,limit);
  } else {
    if(a<uint64_t(c->low.n))return at(c->low,a,0x1000000-(a&0xffffff));
    if(a>=0x03800000 && a<0x04000000)return at(c->wram7,a&65535,65536-(a&65535));
  }
  return {};
}
inline uint8_t* byte(dsmachine_core*c,uint32_t a){return memory(c,a).p;}
template<unsigned N> inline uint8_t* span(dsmachine_core*c,uint32_t a) {
  auto s=memory(c,a);return s.n>=N?s.p:nullptr;
}
template<class T> inline T load(uint8_t*p){T v;std::memcpy(&v,p,sizeof v);if constexpr(std::endian::native==std::endian::big){if constexpr(sizeof(T)==4)v=__builtin_bswap32(v);else if constexpr(sizeof(T)==2)v=__builtin_bswap16(v);}return v;}
template<class T> inline void store(uint8_t*p,T v){if constexpr(std::endian::native==std::endian::big){if constexpr(sizeof(T)==4)v=__builtin_bswap32(v);else if constexpr(sizeof(T)==2)v=__builtin_bswap16(v);}std::memcpy(p,&v,sizeof v);}
}
uint8_t dsmachine_bus_Read(dsmachine_bus*b,uint32_t a){if(!b->c->m->OnRead)if(auto p=rrdsfast::byte(b->c,a))return *p;return dsmachine_bus_ReadReference(b,a);}
uint16_t dsmachine_bus_Read16(dsmachine_bus*b,uint32_t a){if(!b->c->m->OnRead)if(auto p=rrdsfast::span<2>(b->c,a))return rrdsfast::load<uint16_t>(p);return dsmachine_bus_Read16Reference(b,a);}
uint32_t dsmachine_bus_Read32(dsmachine_bus*b,uint32_t a){if(!b->c->m->OnRead)if(auto p=rrdsfast::span<4>(b->c,a))return rrdsfast::load<uint32_t>(p);return dsmachine_bus_Read32Reference(b,a);}
void dsmachine_bus_Write(dsmachine_bus*b,uint32_t a,uint8_t v){if(!b->c->m->OnWrite)if(auto p=rrdsfast::byte(b->c,a)){*p=v;return;}dsmachine_bus_WriteReference(b,a,v);}
void dsmachine_bus_Write16(dsmachine_bus*b,uint32_t a,uint16_t v){if(!b->c->m->OnWrite)if(auto p=rrdsfast::span<2>(b->c,a)){rrdsfast::store(p,v);return;}dsmachine_bus_Write16Reference(b,a,v);}
void dsmachine_bus_Write32(dsmachine_bus*b,uint32_t a,uint32_t v){if(!b->c->m->OnWrite)if(auto p=rrdsfast::span<4>(b->c,a)){rrdsfast::store(p,v);return;}dsmachine_bus_Write32Reference(b,a,v);}
