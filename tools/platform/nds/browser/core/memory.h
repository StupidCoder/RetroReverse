#pragma once
#include "../../../../browser/core/memory.h"
// Direct backing storage only: never invoke CPU or device read handlers.
inline void rrMemoryRegions(){
 using namespace rrmem;regions.clear();supported=false;
 auto*m=machine;if(!m)return;
 if(m->ram.n)add("ram","Main RAM","ram",m->ram.p,m->ram.n,"[33554432]",33554432);
 if(m->swram.n)add("swram","Shared work RAM (physical)","ram",m->swram.p,m->swram.n,"[]",0);
 if(m->pal.n)add("palette","Palette RAM","ram",m->pal.p,m->pal.n,"[83886080]",83886080);
 if(m->oam.n)add("oam","Sprite attributes","ram",m->oam.p,m->oam.n,"[117440512]",117440512);
 if(m->ARM9->itcm.n)add("itcm","ARM9 instruction TCM","ram",m->ARM9->itcm.p,m->ARM9->itcm.n,"[]",0);
 if(m->ARM9->dtcm.n)add("dtcm","ARM9 data TCM","ram",m->ARM9->dtcm.p,m->ARM9->dtcm.n,"[]",0);
 if(m->ARM7->wram7.n)add("wram7","ARM7 work RAM","ram",m->ARM7->wram7.p,m->ARM7->wram7.n,"[58720256]",58720256);
 for(unsigned b=0;b<9;b++){auto v=m->vram->bank[b];if(v.n)add("vram-"+std::to_string(b),"Video RAM bank "+std::string(1,char('A'+b))+" (physical)","ram",v.p,v.n);}
}
#include "../../../../browser/core/memory-api.inc"
