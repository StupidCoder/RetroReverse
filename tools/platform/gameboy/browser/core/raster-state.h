#pragma once
namespace rrgb {
struct Line {
 std::array<uint8_t,8192> vram{};
 std::array<uint8_t,160> oam{};
 std::array<uint8_t,128> io{};
 std::array<uint32_t,160> output{};
 uint32_t before=0,end=0;uint64_t clock=0;
 int windowLine=0;bool valid=false;
};
inline std::array<Line,144> lines;
inline int selected=-1;
inline void begin(){selected=-1;for(auto&line:lines)line.valid=false;}
inline void recordStart(const gameboy_Machine*m,int y,int windowLine){
 if(!rrcapture::trace.active)return;
 if(y==0)begin();
 auto&line=lines[y];line.vram=m->vram;line.oam=m->oam;line.io=m->io;
 line.before=rrcapture::trace.writes.size();line.clock=rrhh::video.steps;
 line.windowLine=windowLine;line.valid=true;
}
inline void recordEnd(int y){
 if(!rrcapture::trace.active)return;
 auto&line=lines[y];line.end=rrcapture::trace.writes.size();
 std::copy_n(rrhh::video.draw.data()+y*160,160,line.output.data());
}
}
