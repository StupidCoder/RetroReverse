#pragma once
template<class... A>void arm_CPU_Halt(arm_CPU*c,std::string f,A...args){c->Halted=true;c->HaltReason=go_fmt_Sprintf(f,args...);}
arm_CPU* arm_NewCPU(gbamachine_bus*);
uint32_t arm_CPU_read16(arm_CPU*,uint32_t);
uint32_t arm_CPU_read32aligned(arm_CPU*,uint32_t);
void arm_CPU_write16(arm_CPU*,uint32_t,uint32_t);
void arm_CPU_write32aligned(arm_CPU*,uint32_t,uint32_t);
template<class...A>void gbamachine_Machine_note(gbamachine_Machine*m,std::string f,A...a){auto v=go_fmt_Sprintf(f,a...);if(!get(m->logSeen,v)){m->logSeen[v]=true;m->Log=append(m->Log,v);}}
void rrGBAMemWrite(gbamachine_Machine*,const Slice<uint8_t>&,uint32_t);

void rrGBARegisterWrite(gbamachine_Machine*,uint32_t,uint16_t);
