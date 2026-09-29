#pragma once
#include "../../../../browser/core/memory.h"
inline void rrMemoryRegions(){
 using namespace rrmem;regions.clear();auto*m=machine;supported=false;
 add("ewram","External work RAM","ram",m->ewram.p,m->ewram.n,"[33554432]",0x02000000);
 add("iwram","Internal work RAM","ram",m->iwram.p,m->iwram.n,"[50331648]",0x03000000);
 add("vram","Video RAM","ram",m->vram.p,m->vram.n,"[100663296]",0x06000000);
 add("palette","Palette RAM","ram",m->pal.p,m->pal.n,"[83886080]",0x05000000);
 add("oam","Object attributes","ram",m->oam.p,m->oam.n,"[117440512]",0x07000000);
 if(m->eeprom.present)add("eeprom","Cartridge EEPROM","ram",m->eeprom.data.p,m->eeprom.data.n);
 for(unsigned off=0;off<m->rom.n;off+=65536){auto n=std::min(65536u,unsigned(m->rom.n)-off);add("rom-"+std::to_string(off/65536),"Cartridge ROM +"+std::to_string(off),"rom",m->rom.p+off,n,"["+std::to_string(0x08000000+off)+","+std::to_string(0x0a000000+off)+(off>=0x1000000&&m->eeprom.present?"":","+std::to_string(0x0c000000+off))+"]",off);}
}
#include "../../../../browser/core/memory-api.inc"
