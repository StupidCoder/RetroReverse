#pragma once
template<class... A>void arm_CPU_Halt(arm_CPU*c,std::string f,A...args){c->Halted=true;c->HaltReason=go_fmt_Sprintf(f,args...);}
arm_CPU* arm_NewCPU(dsmachine_bus*);
uint32_t arm_CPU_read16(arm_CPU*,uint32_t);
uint32_t arm_CPU_read32aligned(arm_CPU*,uint32_t);
void arm_CPU_write16(arm_CPU*,uint32_t,uint32_t);
void arm_CPU_write32aligned(arm_CPU*,uint32_t,uint32_t);
template<class...A>void dsmachine_Machine_note(dsmachine_Machine*m,std::string f,A...args){auto s=go_fmt_Sprintf(f,args...);if(!get(m->logSeen,s)){m->logSeen[s]=true;m->Log=append(m->Log,s);}}
template<class...A>Error dsmachine_errf(std::string f,A...args){return {go_fmt_Sprintf(f,args...)};}
