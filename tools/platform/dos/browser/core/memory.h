#pragma once
#include "../../../../browser/core/memory.h"
// Direct backing storage only: never invoke CPU or device read handlers.
inline void rrMemoryRegions(){
 using namespace rrmem;regions.clear();supported=false;
 if(!realMachine&&!protectedMachine)return;
 auto mem=realMachine?realMachine->Mem:protectedMachine->Mem;
 add("ram","Main RAM backing store","ram",mem.p,mem.n,"[0]");
 if(realMachine){for(unsigned b=0;b<4;b++)add("vga-"+std::to_string(b),"VGA plane "+std::to_string(b),"ram",realMachine->vga->planes[b].data(),65536);add("palette","VGA palette","ram",realMachine->io->Pal.data(),768);}
 else add("palette","VGA palette","ram",protectedMachine->Pal.data(),768);
}
#include "../../../../browser/core/memory-api.inc"
