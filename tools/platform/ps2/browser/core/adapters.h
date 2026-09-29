#pragma once
inline r5900_CPU*r5900_NewCPU(ps2_Machine*m){auto*c=arenaNew(r5900_CPU{});c->bus=m;c->fetch=[m](uint32_t a){return ps2_Machine_Fetch32(m,a);};r5900_CPU_Reset(c);return c;}
inline mips_CPU*mips_NewCPU(ps2_IOP*p){auto*c=arenaNew(mips_CPU{});c->bus=p;mips_CPU_Reset(c);return c;}
