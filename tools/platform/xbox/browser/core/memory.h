#pragma once
#include "../../../../browser/core/memory.h"
// Direct backing storage only: never invoke CPU or device read handlers.
inline void rrMemoryRegions(){
 using namespace rrmem;regions.clear();supported=false;
 auto*m=machine;if(!m)return;
 if(m->RAM.n)add("ram","Unified CPU and GPU RAM","ram",m->RAM.p,m->RAM.n,"[0,2147483648]",0);
}
#include "../../../../browser/core/memory-api.inc"
