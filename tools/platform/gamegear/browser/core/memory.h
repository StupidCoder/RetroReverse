#pragma once
#include "../../../../browser/core/memory.h"
inline void rrMemoryRegions(){
 using namespace rrmem;regions.clear();supported=true;auto*m=machine;
 add("ram","Work RAM","ram",m->ram.data(),8192,"[49152,57344]",0xc000);
 add("vram","Video RAM","ram",m->VDP.VRAM.data(),16384);
 add("cram","Color RAM","ram",m->VDP.CRAM.data(),64);
 for(int b=0;b<m->nbanks;b++){
  std::string aliases="[";for(int s=0;s<3;s++)if(m->slot[s]==b){if(aliases.size()>1)aliases+=",";aliases+=std::to_string(s*16384);}aliases+="]";
  std::string mappings="[";
  auto map=[&](unsigned cpu,unsigned off,unsigned size){if(mappings.size()>1)mappings+=",";mappings+="{\"cpu\":"+std::to_string(cpu)+",\"offset\":"+std::to_string(off)+",\"size\":"+std::to_string(size)+"}";};
  if(b==0)map(0,0,1024);
  for(int s=0;s<3;s++)if(m->slot[s]==b)map(s*16384+(s==0?1024:0),s==0?1024:0,s==0?15360:16384);
  mappings+="]";
  add("rom-"+std::to_string(b),"ROM bank "+std::to_string(b),"rom",m->rom.p+b*16384,16384,aliases,b*16384,mappings);
 }
}
#include "../../../../browser/core/memory-api.inc"
