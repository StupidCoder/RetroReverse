#pragma once
#include "../../../../browser/core/memory.h"
// Direct backing storage only: never invoke CPU or device read handlers.
inline void rrMemoryRegions(){
 using namespace rrmem;regions.clear();supported=false;
 auto*m=machine;if(!m)return;
 if(m->ram.n)add("ram","Main RAM","ram",m->ram.p,m->ram.n,"[134217728,2281701376,2818572288]",134217728);
 if(m->vram.n)add("vram","Video RAM","ram",m->vram.p,m->vram.n,"[67108864,1140850688]",67108864);
 if(m->scratch.n)add("scratch","Scratchpad","ram",m->scratch.p,m->scratch.n,"[65536]",65536);
}
#include "../../../../browser/core/memory-api.inc"
