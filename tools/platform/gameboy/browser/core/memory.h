#pragma once
#include "../../../../browser/core/memory.h"
inline void rrMemoryRegions(){
 using namespace rrmem;regions.clear();auto*m=machine;supported=false;
 add("ram","Work RAM","ram",m->wram.data(),8192,"[49152,57344]",0xc000,"[{\"cpu\":49152,\"offset\":0,\"size\":8192},{\"cpu\":57344,\"offset\":0,\"size\":7680}]");
 add("vram","Video RAM","ram",m->vram.data(),8192,"[32768]",0x8000);
 add("oam","Sprite attributes","ram",m->oam.data(),160,"[65024]",0xfe00);
 add("hram","High RAM","ram",m->hram.data(),127,"[65408]",0xff80);
 unsigned ramSize=m->rom[0x149]==1?2048:m->rom[0x149]==2?8192:m->rom[0x149]==3?32768:0;
 for(unsigned b=0;b<(ramSize+8191)/8192;b++)add("save-"+std::to_string(b),"Cartridge RAM bank "+std::to_string(b),"ram",m->extram.data()+b*8192,std::min(8192u,ramSize-b*8192),m->ramEnable&&unsigned(m->mode?m->ramBank:0)==b?"[40960]":"[]");
 auto bank0=m->rom[0x147]!=0&&m->mode==1?(m->ramBank<<5)%m->nbanks:0;
 for(int b=0;b<m->nbanks;b++){
  std::string aliases="[";if(b==bank0)aliases+="0";if(b==m->romBank){if(aliases.size()>1)aliases+=",";aliases+="16384";}aliases+="]";
  add("rom-"+std::to_string(b),"ROM bank "+std::to_string(b),"rom",m->rom.p+b*16384,16384,aliases,b*16384);
 }
}
#include "../../../../browser/core/memory-api.inc"
