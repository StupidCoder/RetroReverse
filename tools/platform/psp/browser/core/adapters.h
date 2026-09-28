#pragma once
inline std::tuple<Slice<uint8_t>,Error>psp_cbcDecryptZero(Slice<uint8_t>key,Slice<uint8_t>src){return rrcrypto::cbc(key,src);}
inline uint8_t*rrMemory(psp_Machine*m,uint32_t a,uint32_t n){a&=0x1fffffff;auto at=[&](Slice<uint8_t>&s,uint32_t base)->uint8_t*{uint32_t off=a-base;return off<=uint32_t(s.n)&&n<=uint32_t(s.n)-off?s.p+off:nullptr;};if(auto*p=at(m->ram,0x08000000))return p;if(auto*p=at(m->vram,0x04000000))return p;return at(m->scratch,0x10000);}
inline uint32_t rrRead(psp_Machine*m,uint32_t a,unsigned n){if(!m->OnRead)if(auto*p=rrMemory(m,a,n)){uint32_t v=0;std::memcpy(&v,p,n);return v;}uint32_t v=0;for(unsigned i=0;i<n;i++)v|=uint32_t(psp_Machine_Read(m,a+i))<<(8*i);return v;}
inline void rrWrite(psp_Machine*m,uint32_t a,uint32_t v,unsigned n){if(rrObserveWrite||!m->OnWrite)if(auto*p=rrMemory(m,a,n)){if(rrObserveWrite)rrObserveWrite(m,a,v,n);std::memcpy(p,&v,n);return;}for(unsigned i=0;i<n;i++)psp_Machine_Write(m,a+i,v>>(8*i));}
inline uint32_t psp_Machine_read32(psp_Machine*m,uint32_t a){return rrRead(m,a,4);}
inline void psp_Machine_write16(psp_Machine*m,uint32_t a,uint16_t v){rrWrite(m,a,v,2);}
inline void psp_Machine_write32(psp_Machine*m,uint32_t a,uint32_t v){rrWrite(m,a,v,4);}
inline uint32_t allegrex_CPU_read16(allegrex_CPU*c,uint32_t a){return rrRead(c->bus,a,2);}
inline uint32_t allegrex_CPU_read32(allegrex_CPU*c,uint32_t a){return rrRead(c->bus,a,4);}
inline void allegrex_CPU_write16(allegrex_CPU*c,uint32_t a,uint32_t v){rrWrite(c->bus,a,v,2);}
inline void allegrex_CPU_write32(allegrex_CPU*c,uint32_t a,uint32_t v){rrWrite(c->bus,a,v,4);}
inline allegrex_CPU*allegrex_NewCPU(psp_Machine*m){auto*c=arenaNew(allegrex_CPU{});c->bus=m;allegrex_CPU_Reset(c);return c;}
template<class...A>uint32_t psp_Machine_callGuest(psp_Machine*m,uint32_t entry,A...args){auto saved=allegrex_CPU_SaveState(m->CPU);allegrex_CPU_SetPC(m->CPU,entry);uint32_t a[]={uint32_t(args)...};for(size_t i=0;i<sizeof...(args);i++)allegrex_CPU_SetReg(m->CPU,4+i,a[i]);uint32_t sp=saved.R[29];if(!sp||sp==psp_intrExitAddr)sp=psp_stackTop;allegrex_CPU_SetReg(m->CPU,29,(sp-0x800)&~15);allegrex_CPU_SetReg(m->CPU,31,psp_intrExitAddr);for(int i=0;i<4000000&&m->CPU->PC!=psp_intrExitAddr&&!m->CPU->Halted&&!m->Halted;i++)allegrex_CPU_Step(m->CPU);auto v=allegrex_CPU_Reg(m->CPU,2);bool halt=m->CPU->Halted;auto reason=m->CPU->HaltReason;allegrex_CPU_LoadState(m->CPU,saved);if(halt){m->CPU->Halted=true;m->CPU->HaltReason=reason;}return v;}

inline uint16_t psp_u16(psp_Machine*m,uint32_t a){return rrRead(m,a,2);}

#include "texture.h"
