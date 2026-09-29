#pragma once
#include "generated.cpp"
using Machine=gamegear_Machine;
#define HH_CPU_NAME "Z80 CPU and ports"
#define HH_CAPTURE_BEGIN rrgg::begin()
#define HH_CAPTURE_BYTES sizeof(rrgg::lines)
inline Machine*rrBoot(const std::vector<uint8_t>&input){
 unsigned skip=input.size()%16384==512?512:0,n=input.size()-skip;if(n<16384||n>4194304||n%16384)throw std::runtime_error("Game Gear image must contain 16 KiB–4 MiB in complete ROM banks (optional 512-byte header)");
 auto rom=Slice<uint8_t>::make(n);std::memcpy(rom.p,input.data()+skip,n);auto*m=gamegear_NewMachine(rom);for(auto&bank:m->slot)bank%=m->nbanks;rrhh::init(0xff000000);rrgg::timing={};m->VDP.Regs[10]=1;return m;
}
namespace rrgg {
inline void lineStart(Machine*m){
 auto&v=m->VDP;unsigned line=rrhh::video.phase/lineCycles;v.line=counter(line);
 timing.scrollX=v.Regs[8];if(line==261)timing.scrollY=v.Regs[9];
 // F4 advances the V counter and accounts for the line just completed,
 // including the pre-display line. A zero reload interrupts every line.
 if(line<=192){if(timing.lineCounter==0){timing.lineCounter=v.Regs[10];timing.linePending=true;}else timing.lineCounter--;}
 else timing.lineCounter=v.Regs[10];
 if(line==192){v.status|=128;rrhh::present();}
 irq(m);
 if(line<192)rrGGLine(m,line,timing.scrollX,timing.scrollY);
}
inline void tick(Machine*m,unsigned cycles){
 auto&h=rrhh::video;
 if(!timing.started){timing.started=true;lineStart(m);}
 while(cycles){unsigned n=std::min(cycles,lineCycles-h.phase%lineCycles);h.phase+=n;timing.cycles+=n;cycles-=n;
  if(h.phase%lineCycles==0){if(h.phase==frameCycles)h.phase=0;lineStart(m);}
 }
}
}
inline void rrAdvance(Machine*m){
 rrgg::tick(m,0);rrgg::instructionCycles=rrgg::duration(m);rrgg::elapsedCycles=0;
 gamegear_Machine_step(m);
 rrgg::tick(m,rrgg::instructionCycles-rrgg::elapsedCycles);rrgg::instructionCycles=rrgg::ioCycles=rrgg::elapsedCycles=0;
}
inline void rrPad(Machine*m,uint32_t buttons){m->PadDC=~buttons;m->Pad00=buttons&128?0x7f:0xff;}
inline double rrSeconds(Machine*){return double(rrgg::timing.cycles)/3579545;}
inline std::vector<uint8_t>rrMemoryImage(Machine*m){std::vector<uint8_t>b(rrhh::memorySize);std::memcpy(b.data()+0xc000,m->ram.data(),8192);std::memcpy(b.data()+0x10000,m->VDP.VRAM.data(),16384);std::memcpy(b.data()+0x14000,m->VDP.CRAM.data(),64);std::memcpy(b.data()+0x14040,m->VDP.Regs.data(),16);rrhh::includeFrame(b);return b;}
inline void rrAfterWrite(Machine*m,uint16_t a,uint8_t){if(a>=0xc000)rrmem::access(rrgg::timing.cycles,0,a&8191,m->ram[a&8191],1,2,rrhh::video.pc,a);if(a>=0xc000)rrhh::memoryWrite(0xc000+(a&0x1fff),m->ram[a&0x1fff]);}
inline void rrAfterPort(Machine*m,uint16_t port,uint8_t value){auto&v=m->VDP;
 if(uint8_t(port)==0xbe){auto a=uint16_t(v.addr-1);if(v.code==3)rrhh::memoryWrite(0x14000+(a&63),v.CRAM[a&63]);else rrhh::memoryWrite(0x10000+(a&16383),v.VRAM[a&16383]);}
 else if(uint8_t(port)==0xbf&&!v.latched&&v.code==2)rrhh::memoryWrite(0x14040+(value&15),v.Regs[value&15]);
}

inline void rrAfterRead(Machine*m,uint16_t a,uint8_t v){
 if(!rrmem::active)return;
 auto [off,ram]=gamegear_FileOffset(m->slot,a);
 if(ram)rrmem::access(rrgg::timing.cycles,0,a&8191,v,1,1,rrhh::video.pc,a);
 else if(off<m->rom.n)rrmem::access(rrgg::timing.cycles,3+off/16384,off%16384,v,1,1,rrhh::video.pc,a);
}
