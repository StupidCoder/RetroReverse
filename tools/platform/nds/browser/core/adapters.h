#pragma once
arm_CPU* arm_NewCPU(dsmachine_bus*b){auto c=arenaNew(arm_CPU{});c->bus=c->wide=b;arm_CPU_Reset(c);return c;}
uint32_t arm_CPU_read16(arm_CPU*c,uint32_t a){if(!(a&1))return dsmachine_bus_Read16(c->bus,a);return dsmachine_bus_Read(c->bus,a)|uint32_t(dsmachine_bus_Read(c->bus,a+1))<<8;}
uint32_t arm_CPU_read32aligned(arm_CPU*c,uint32_t a){return dsmachine_bus_Read32(c->bus,a&~3u);}
void arm_CPU_write16(arm_CPU*c,uint32_t a,uint32_t v){if(!(a&1)){dsmachine_bus_Write16(c->bus,a,v);return;}dsmachine_bus_Write(c->bus,a,v);dsmachine_bus_Write(c->bus,a+1,v>>8);}
void arm_CPU_write32aligned(arm_CPU*c,uint32_t a,uint32_t v){dsmachine_bus_Write32(c->bus,a&~3u,v);}
template<class...V>void dsmachine_geom_emit(dsmachine_geom*g,dsmachine_gpu3d*gp,V...v){dsmachine_geom_emit(g,gp,Slice<dsmachine_gxVertex>{v...});}

#include "capture-hooks.h"
