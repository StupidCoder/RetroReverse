#pragma once
#include "../../../../browser/core/memory.h"
// Direct backing storage only: never invoke CPU or device read handlers.
inline void rrMemoryRegions(){
 using namespace rrmem;regions.clear();supported=false;
 auto*m=machine;if(!m)return;
 if(m->RAM.n)add("ram","Main RAM","ram",m->RAM.p,m->RAM.n,"[201326592,2348810240,2885681152]",201326592);
 if(m->VRAM.n)add("vram","Video RAM (64-bit texture order)","ram",m->VRAM.p,m->VRAM.n,"[67108864,2751463424]",67108864);
 if(m->AICARAM.n)add("aica","AICA sound RAM","ram",m->AICARAM.p,m->AICARAM.n,"[8388608,2692743168]",8388608);
 if(m->Flash.n)add("flash","Flash storage","ram",m->Flash.p,m->Flash.n,"[2097152]",2097152);
}
#include "../../../../browser/core/memory-api.inc"
