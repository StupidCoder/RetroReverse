#pragma once
#include "../../../../browser/core/memory.h"
// Direct backing storage only: never invoke CPU or device read handlers.
inline void rrMemoryRegions(){
 using namespace rrmem;regions.clear();supported=false;
 auto*m=machine;if(!m)return;
 for(unsigned i=0;i<m->regions.n;i++){auto*r=m->regions[i];if(r&&r->data.n)add("region-"+std::to_string(i),"Mapped "+r->name,"ram",r->data.p,r->data.n,"["+std::to_string(r->base)+"]",r->base);}
}
#include "../../../../browser/core/memory-api.inc"
