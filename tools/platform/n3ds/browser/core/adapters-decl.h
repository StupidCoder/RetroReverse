#pragma once
template<class...A>void arm_CPU_Halt(arm_CPU*c,std::string f,A...args){c->Halted=true;c->HaltReason=go_fmt_Sprintf(f,args...);}
arm_CPU* arm_NewCPU(n3ds_Machine*);
uint32_t arm_CPU_read16(arm_CPU*,uint32_t);
uint32_t arm_CPU_read32aligned(arm_CPU*,uint32_t);
void arm_CPU_write16(arm_CPU*,uint32_t,uint32_t);
void arm_CPU_write32aligned(arm_CPU*,uint32_t,uint32_t);
inline int64_t n3ds_maxWorkers=1;
n3ds_workPool* n3ds_GPU_pool(n3ds_GPU*);
void n3ds_workPool_run(n3ds_workPool*,int64_t,std::function<void(int64_t)>);
void n3ds_Machine_Close(n3ds_Machine*);

uint32_t n3ds_Machine_Read16(n3ds_Machine*,uint32_t);
uint32_t n3ds_Machine_Read32(n3ds_Machine*,uint32_t);
void n3ds_Machine_Write16(n3ds_Machine*,uint32_t,uint32_t);
void n3ds_Machine_Write32(n3ds_Machine*,uint32_t,uint32_t);

inline uint32_t rr3dsCirclePad=0;
