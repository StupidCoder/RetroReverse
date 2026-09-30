#pragma once
#include "../../../../browser/core/memory.h"
// Direct backing storage only: never invoke CPU or device read handlers.
inline void rrMemoryRegions(){
 using namespace rrmem;regions.clear();supported=false;
 auto*m=machine;if(!m)return;
 if(m->ram.n)add("ram","Emotion Engine RAM","ram",m->ram.p,m->ram.n,"[0,2147483648,2684354560]",0);
 if(m->spram.n)add("scratch","EE scratchpad","ram",m->spram.p,m->spram.n,"[1879048192]",1879048192);
 if(m->iopRAM.n)add("iop","IOP RAM","ram",m->iopRAM.p,m->iopRAM.n,"[]",0);
 if(m->bios.n)add("bios","Loaded BIOS","rom",m->bios.p,m->bios.n,"[532676608,3217031168]",532676608);
 if(m->gs)if(m->gs->vram.n)add("vram","GS video RAM","ram",m->gs->vram.p,m->gs->vram.n,"[]",0);
 for(unsigned b=0;b<2;b++){auto*v=m->vifs[b];if(!v)continue;auto base=0x11000000u+b*0x8000u;
  if(v->micro.n)add("vu-micro-"+std::to_string(b),"VU"+std::to_string(b)+" microcode","ram",v->micro.p,v->micro.n,"["+std::to_string(base)+"]",base);
  if(v->data.n)add("vu-data-"+std::to_string(b),"VU"+std::to_string(b)+" data","ram",v->data.p,v->data.n,"["+std::to_string(base+0x4000)+"]",base+0x4000);
 }
}
#include "../../../../browser/core/memory-api.inc"
