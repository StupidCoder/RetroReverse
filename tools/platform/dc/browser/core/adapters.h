#pragma once
arm_CPU* arm_NewCPU(dc_armBus*b){auto c=arenaNew(arm_CPU{});c->bus=c->wide=b;arm_CPU_Reset(c);return c;}
uint32_t arm_CPU_read16(arm_CPU*c,uint32_t a){if(!(a&1))return dc_armBus_Read16(*c->bus,a);return dc_armBus_Read(*c->bus,a)|uint32_t(dc_armBus_Read(*c->bus,a+1))<<8;}
uint32_t arm_CPU_read32aligned(arm_CPU*c,uint32_t a){return dc_armBus_Read32(*c->bus,a&~3u);}
void arm_CPU_write16(arm_CPU*c,uint32_t a,uint32_t v){if(!(a&1)){dc_armBus_Write16(*c->bus,a,v);return;}dc_armBus_Write(*c->bus,a,v);dc_armBus_Write(*c->bus,a+1,v>>8);}
void arm_CPU_write32aligned(arm_CPU*c,uint32_t a,uint32_t v){dc_armBus_Write32(*c->bus,a&~3u,v);}
arm_CPU*arm_NewCPU(dc_armBus b){return arm_NewCPU(arenaNew(b));}
sh4_CPU*sh4_NewCPU(dc_Machine*m){auto*c=arenaNew(sh4_CPU{});c->bus=c->fetcher=m;sh4_CPU_Reset(c);return c;}
uint16_t sh4_CPU_fetchInstr(sh4_CPU*c,uint32_t a){if(a>=0xe0000000){sh4_CPU_Halt(c,"instruction fetch from P4 at %08X",a);return 9;}return dc_Machine_Fetch16(c->bus,a&0x1fffffffu);}
