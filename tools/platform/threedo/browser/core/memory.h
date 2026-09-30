#pragma once
#include "../../../../browser/core/memory.h"
// Direct backing reads; activity is CPU/bus writes, not direct HLE bulk stores.
inline void rrMemoryRegions(){using namespace rrmem;regions.clear();supported=true;auto*m=machine;if(!m)return;add("ram","Main DRAM (bus writes only)","ram",m->dram.p,m->dram.n,"[0]",0);add("vram","Video RAM (bus writes only)","ram",m->vram.p,m->vram.n,"[2097152]",2097152);add("items","Simulated kernel item memory","ram",m->imem.p,m->imem.n,"[4194304]",4194304);}
#define rr_activity_begin rr_base_activity_begin
#define rr_activity_end rr_base_activity_end
#include "../../../../browser/core/memory-api.inc"
#undef rr_activity_begin
#undef rr_activity_end
static std::vector<std::pair<uint32_t,uint32_t>> rrSelected;
static std::function<void(uint32_t,uint32_t,uint32_t)> rrOldWrite;
static uint32_t rrOldLo,rrOldHi;
static bool rrObserving=false;
extern "C" {
void rr_activity_clear_selection(){rrSelected.clear();}
int rr_activity_select(uint32_t a,uint32_t n){if(rrSelected.size()>=32||!n||a>=0x300000||n>0x300000-a)return 0;rrSelected.push_back({a,n});return 1;}
void rr_activity_end();
void rr_activity_begin(uint32_t mask){if(rrObserving)rr_activity_end();rr_base_activity_begin(mask);rrObserving=true;rrOldWrite=machine->OnWrite;rrOldLo=machine->WatchLo;rrOldHi=machine->WatchHi;machine->WatchLo=0;machine->WatchHi=0x300000;machine->OnWrite=[](uint32_t a,uint32_t v,uint32_t pc){if(rrOldWrite&&a>=rrOldLo&&a<rrOldHi)rrOldWrite(a,v,pc);if(!rrSelected.empty()&&!std::any_of(rrSelected.begin(),rrSelected.end(),[a](auto r){return a>=r.first&&a-r.first<r.second;}))return;rrmem::access(rrSchedulerClockBase+runContext.steps,a<0x200000?0:1,a<0x200000?a:a-0x200000,v,1,2,pc,a);};}
void rr_activity_end(){rr_base_activity_end();if(!rrObserving)return;rrObserving=false;machine->OnWrite=rrOldWrite;machine->WatchLo=rrOldLo;machine->WatchHi=rrOldHi;rrOldWrite={};}
}
