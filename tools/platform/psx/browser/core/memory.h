#pragma once
#include "../../../../browser/core/memory.h"
// Direct backing storage only: never invoke CPU or device read handlers.
inline void rrMemoryRegions(){
 using namespace rrmem;regions.clear();supported=false;
 auto*m=machine.get();if(!m)return;
 if(m->ram.size())add("ram","Main RAM","ram",m->ram.data(),m->ram.size(),"[0,2147483648,2684354560]",0);
 if(m->scratch.size())add("scratch","Scratchpad","ram",m->scratch.data(),m->scratch.size(),"[528482304,2675965952,3212836864]",528482304);
 if(m->gpu.vram.size()*2)add("vram","GPU VRAM","ram",reinterpret_cast<const uint8_t*>(m->gpu.vram.data()),m->gpu.vram.size()*2,"[]",0);
}
#include "../../../../browser/core/memory-api.inc"
