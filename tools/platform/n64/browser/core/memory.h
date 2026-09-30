#pragma once
#include "../../../../browser/core/memory.h"
// Direct backing storage only: never invoke CPU or device read handlers.
inline void rrMemoryRegions(){
 using namespace rrmem;regions.clear();supported=false;
 auto*m=machine;if(!m)return;
 if(m->RDRAM.n)add("ram","RDRAM","ram",m->RDRAM.p,m->RDRAM.n,"[0,2147483648,2684354560]",0);
 if(m->DMEM.n)add("dmem","RSP data memory","ram",m->DMEM.p,m->DMEM.n,"[2751463424]",67108864);
 if(m->IMEM.n)add("imem","RSP instruction memory","ram",m->IMEM.p,m->IMEM.n,"[2751467520]",67112960);
 if(m->PIF.n)add("pif","PIF RAM","ram",m->PIF.p,m->PIF.n,"[532678592,3217033152]",532678592);
 if(m->EEPROM.n)add("eeprom","Cartridge EEPROM","ram",m->EEPROM.p,m->EEPROM.n,"[]",0);
 if(m->ROM.n)add("rom","Cartridge ROM","rom",m->ROM.p,m->ROM.n,"[268435456,2952790016]",268435456);
}
#include "../../../../browser/core/memory-api.inc"
