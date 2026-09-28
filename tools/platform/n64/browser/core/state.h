#pragma once
#include "state-fields.h"
inline void rebindState(n64_Machine*m,Slice<uint8_t>rom){
 if(!m->CPU||m->RDRAM.n!=4*1024*1024||m->DMEM.n!=4096||m->IMEM.n!=4096||m->PIF.n!=64||m->rdp.Pending.n>65536)throw std::runtime_error("Invalid N64 machine state");
 m->ROM=rom;m->CPU->bus=m;
 if(m->RSP){m->RSP->regs=m;m->RSP->DMEM=m->DMEM;m->RSP->IMEM=m->IMEM;}
}
