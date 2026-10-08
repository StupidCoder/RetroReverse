#pragma once
#include <bit>
namespace rrgpu {extern bool enabled;}
namespace rrarm {
// No cached host pointers: every access checks the current indexed page. This
// preserves aliases, remapping and self-modifying instructions without epochs.
inline bool memoryEnabled=true;
inline uint8_t* ordinary(n3ds_Machine*m,uint32_t address,uint32_t width){
 if(!memoryEnabled||!rrgpu::enabled||m->OnRead||m->OnWrite||m->HidTrace||!m->pages||uint64_t(address>>12)>=uint64_t(m->pages.n)||(address&4095)+width>4096)return nullptr;
 auto r=m->pages[address>>12];
 if(!r||address<r->base||uint64_t(address-r->base)+width>uint64_t(r->data.n))return nullptr;
 return r->data.p+(address-r->base);
}
template<class T>inline T load(const uint8_t*p){
 if constexpr(std::endian::native==std::endian::little){T value;std::memcpy(&value,p,sizeof(T));return value;}
 else {T value=0;for(size_t i=0;i<sizeof(T);i++)value|=T(p[i])<<(i*8);return value;}
}
template<class T>inline void store(uint8_t*p,T value){
 if constexpr(std::endian::native==std::endian::little)std::memcpy(p,&value,sizeof(T));
 else for(size_t i=0;i<sizeof(T);i++)p[i]=value>>(i*8);
}
// The generic ARM model also supports a separate wide bus. Only bypass it if
// it is absent or names the same machine as the ordinary byte bus.
inline uint8_t* word(arm_CPU*c,uint32_t a){return (!c->wide||c->wide==c->bus)?ordinary(c->bus,a,4):nullptr;}
}
