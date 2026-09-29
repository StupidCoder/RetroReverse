#pragma once
namespace rrgg {
struct Line {
 std::array<uint8_t,16384> vram{};
 std::array<uint8_t,64> cram{};
 std::array<uint8_t,16> regs{};
 std::array<uint32_t,160> output{};
 uint32_t before=0,end=0;uint64_t clock=0;
 uint8_t scrollX=0,scrollY=0,lineCounter=0;bool linePending=false,framePending=false,valid=false;
};
inline std::array<Line,144> lines;
inline int selected=-1;
inline void begin(){selected=-1;for(auto&line:lines)line.valid=false;}
inline void recordStart(const gamegear_Machine*m,int y,uint8_t scrollX,uint8_t scrollY){
 if(!rrcapture::trace.active)return;if(y==0)begin();
 auto&line=lines[y];line.vram=m->VDP.VRAM;line.cram=m->VDP.CRAM;line.regs=m->VDP.Regs;
 line.before=rrcapture::trace.writes.size();line.clock=timing.cycles;line.scrollX=scrollX;line.scrollY=scrollY;
 line.lineCounter=timing.lineCounter;line.linePending=timing.linePending;line.framePending=m->VDP.status&128;line.valid=true;
}
inline void recordEnd(int y){if(!rrcapture::trace.active)return;auto&line=lines[y];line.end=rrcapture::trace.writes.size();std::copy_n(rrhh::video.draw.data()+y*160,160,line.output.data());}
}
