#pragma once
#include "../../../../browser/core/memory.h"
// Direct backing storage only: never invoke CPU or device read handlers.
inline void rrMemoryRegions(){
 using namespace rrmem;regions.clear();supported=false;
 auto*m=machine;if(!m)return;
 if(m->dram.n)add("ram","Main DRAM","ram",m->dram.p,m->dram.n,"[0]",0);
 if(m->vram.n)add("vram","Video RAM","ram",m->vram.p,m->vram.n,"[2097152]",2097152);
 if(m->imem.n)add("items","Simulated kernel item memory","ram",m->imem.p,m->imem.n,"[4194304]",4194304);
}
#include "../../../../browser/core/memory-api.inc"
