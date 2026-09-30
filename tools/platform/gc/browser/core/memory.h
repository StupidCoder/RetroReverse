#pragma once
#include "../../../../browser/core/memory.h"
// Direct backing storage only: never invoke CPU or device read handlers.
inline void rrMemoryRegions(){
 using namespace rrmem;regions.clear();supported=false;
 auto*m=machine;if(!m)return;
 if(m->RAM.n)add("ram","Main RAM (MEM1)","ram",m->RAM.p,m->RAM.n,"[0,2147483648,3221225472]",0);
 if(m->ARAM.n)add("aram","Audio RAM","ram",m->ARAM.p,m->ARAM.n,"[]",0);
}
#include "../../../../browser/core/memory-api.inc"
