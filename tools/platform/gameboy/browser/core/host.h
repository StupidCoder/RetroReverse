#pragma once
#include "generated.cpp"
using Machine=gameboy_Machine;
#define HH_CAPTURE_BEGIN rrgb::begin()
#define HH_CAPTURE_BYTES sizeof(rrgb::lines)
#define HH_CPU_NAME "SM83 CPU and timers"
inline Machine*rrBoot(const std::vector<uint8_t>&input){
 if(input.size()<32768||input.size()>2097152||input.size()%16384)throw std::runtime_error("Game Boy image must contain 32 KiB–2 MiB in complete ROM banks");
 if(input[0x143]==0xc0)throw std::runtime_error("Game Boy Color-only cartridges are not supported by the DMG core");
 auto mapper=input[0x147];if(mapper>3)throw std::runtime_error("This DMG core supports ROM-only and MBC1 cartridges");
 auto rom=Slice<uint8_t>::make(input.size());std::memcpy(rom.p,input.data(),input.size());auto*m=gameboy_NewMachine(rom);m->io[0x47]=0xfc;m->io[0x48]=m->io[0x49]=255;rrhh::init(0xffffffff);rrGBStat(m);return m;
}
inline void rrAdvance(Machine*m){gameboy_Machine_Step(m);}
inline void rrPad(Machine*m,uint32_t buttons){auto old=gameboy_Machine_joyp(m);m->Buttons=buttons;if((old&~gameboy_Machine_joyp(m))&15)m->io[15]|=16;}
inline double rrSeconds(Machine*m){return m?double(m->Cycles)/4194304:0;}
inline std::vector<uint8_t>rrMemoryImage(Machine*m){std::vector<uint8_t>b(rrhh::memorySize);std::memcpy(b.data()+0x8000,m->vram.data(),8192);std::memcpy(b.data()+0xc000,m->wram.data(),8192);std::memcpy(b.data()+0xfe00,m->oam.data(),160);std::memcpy(b.data()+0xff80,m->hram.data(),127);std::memcpy(b.data()+0x10000,m->extram.data(),32768);
 for(unsigned i:{0x40,0x42,0x43,0x45,0x46,0x47,0x48,0x49,0x4a,0x4b})b[0xff00+i]=m->io[i];rrhh::includeFrame(b);return b;
}
inline void rrAfterWrite(Machine*m,uint16_t a,uint8_t value){if(!rrcapture::trace.active)return;
 if(a>=0x8000&&a<0xa000)rrhh::memoryWrite(a,m->vram[a-0x8000]);
 else if(a>=0xa000&&a<0xc000&&m->ramEnable){unsigned off=gameboy_Machine_ramOff(m,a);rrhh::memoryWrite(0x10000+off,m->extram[off]);}
 else if(a>=0xc000&&a<0xfe00)rrhh::memoryWrite(0xc000+(a&0x1fff),m->wram[a&0x1fff]);
 else if(a>=0xfe00&&a<0xfea0)rrhh::memoryWrite(a,m->oam[a-0xfe00]);
 else if(a>=0xff80&&a<0xffff)rrhh::memoryWrite(a,m->hram[a-0xff80]);
 else if(a==0xff40||(a>=0xff42&&a<=0xff4b&&a!=0xff44))rrhh::memoryWrite(a,m->io[a&127]);
 if(a==0xff46)for(unsigned i=0;i<160;i++)rrhh::memoryWrite(0xfe00+i,m->oam[i]);
}
