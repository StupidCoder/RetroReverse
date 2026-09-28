#pragma once
#include "generated.cpp"
using Machine=gamegear_Machine;
#define HH_CPU_NAME "Z80 CPU and ports"
inline Machine*rrBoot(const std::vector<uint8_t>&input){
 unsigned skip=input.size()%16384==512?512:0,n=input.size()-skip;if(n<16384||n>4194304||n%16384)throw std::runtime_error("Game Gear image must contain 16 KiB–4 MiB in complete ROM banks (optional 512-byte header)");
 auto rom=Slice<uint8_t>::make(n);std::memcpy(rom.p,input.data()+skip,n);auto*m=gamegear_NewMachine(rom);for(auto&bank:m->slot)bank%=m->nbanks;rrhh::init(0xff000000);return m;
}
inline void rrAdvance(Machine*m){auto&h=rrhh::video;
 if(h.phase<20000){int line=h.phase*262/20000;if(line!=h.lastLine){if(h.lastLine>=0)rrGGLine(m,h.lastLine);h.lastLine=line;}m->VDP.line=line&255;if(m->VDP.line>=192)m->VDP.status|=128;}
 else if(h.phase==20000){m->VDP.line=192;m->VDP.status|=128;if(m->VDP.Regs[1]&32)z80_CPU_RequestIRQ(m->CPU,true);}
 gamegear_Machine_step(m);if(++h.phase==30000){h.phase=0;h.lastLine=-1;z80_CPU_RequestIRQ(m->CPU,false);rrhh::present();}
}
inline void rrPad(Machine*m,uint32_t buttons){m->PadDC=~buttons;m->Pad00=buttons&128?0x7f:0xff;}
inline double rrSeconds(Machine*){return (double(rrhh::video.frames)+double(rrhh::video.phase)/30000)/60;}
inline std::vector<uint8_t>rrMemoryImage(Machine*m){std::vector<uint8_t>b(rrhh::memorySize);std::memcpy(b.data()+0xc000,m->ram.data(),8192);std::memcpy(b.data()+0x10000,m->VDP.VRAM.data(),16384);std::memcpy(b.data()+0x14000,m->VDP.CRAM.data(),64);std::memcpy(b.data()+0x14040,m->VDP.Regs.data(),16);rrhh::includeFrame(b);return b;}
inline void rrAfterWrite(Machine*m,uint16_t a,uint8_t){if(a>=0xc000)rrhh::memoryWrite(0xc000+(a&0x1fff),m->ram[a&0x1fff]);}
inline void rrAfterPort(Machine*m,uint16_t port,uint8_t value){auto&v=m->VDP;
 if(uint8_t(port)==0xbe){auto a=uint16_t(v.addr-1);if(v.code==3)rrhh::memoryWrite(0x14000+(a&63),v.CRAM[a&63]);else rrhh::memoryWrite(0x10000+(a&16383),v.VRAM[a&16383]);}
 else if(uint8_t(port)==0xbf&&!v.latched&&v.code==2)rrhh::memoryWrite(0x14040+(value&15),v.Regs[value&15]);
}
