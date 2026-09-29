#pragma once
#include "amiga.h"
namespace rramiga {
struct Source {std::string role; uint32_t address=0,value=0,before=0; unsigned size=2;};
struct Candidate {std::string kind; int object=-1; uint8_t flags=1; uint32_t color=0; std::vector<Source> sources;};
struct RenderProbe {int x=0,y=0,anchor=-1; uint32_t before=0; bool frozen=false; std::vector<Candidate> candidates;};
namespace inspect {
struct Line {
 std::array<uint16_t,256> regs{};
 std::array<uint32_t,6> pointers{};
 std::array<Sprite,8> sprites{};
 std::array<uint32_t,8> spriteBefore{};
 std::array<uint16_t,32> palette{};
 std::vector<ColorChange> colors;
 uint64_t frame=0,clock=0;
 uint32_t before=0,changeStart=0,changeEnd=0,paletteBefore=0;
 unsigned dma=0,raster=0;
 bool display=false,prepared=false,valid=false;
};
struct Change {uint64_t frame,clock;unsigned line,beam,reg;uint16_t oldValue,value;uint32_t pc,copperPC,before;};
struct BlitWord {
 std::array<uint32_t,4> address{};
 uint32_t before=0,write=0;
 uint16_t rawA=0,rawB=0,a=0,b=0,c=0,d=0,oldD=0;
 bool written=false;
};
struct Blit {
 uint64_t frame=0,clock=0;
 uint32_t event=0,pc=0,copperPC=0,before=0,end=0;
 unsigned line=0,beam=0,width=0,height=0,first=0,count=0;
 uint16_t con0=0,con1=0,firstMask=0,lastMask=0;
 std::array<uint32_t,4> pointers{};
 std::array<int,4> mod{};
 bool complete=true;
};
inline std::array<Line,H> lines;
inline std::vector<Change> changes;
inline std::vector<Blit> blits;
inline std::vector<BlitWord> words;
inline std::array<std::array<uint32_t,W*H>,3> pictures{};
inline std::array<uint32_t,W*H> output{};
inline std::vector<uint8_t> memory;
inline std::array<uint32_t,8> spriteCutoff{};
inline unsigned cursor=0,lineChangeStart=0,linePaletteBefore=0,frameChangeStart=0;
inline int selected=-1,selectedBlit=-1;
inline bool overflow=false;
inline constexpr size_t maxWords=(16*1024*1024)/sizeof(BlitWord);
inline void begin(){for(auto&l:lines)l.valid=false;changes.clear();blits.clear();words.clear();memory.clear();cursor=0;selected=selectedBlit=-1;overflow=false;lineChangeStart=linePaletteBefore=frameChangeStart=0;spriteCutoff.fill(0);}
inline void lineBegin(unsigned line){if(rrcapture::trace.active){if(line==0)frameChangeStart=changes.size();lineChangeStart=changes.size();linePaletteBefore=rrcapture::trace.writes.size();}}
inline void change(const Machine&m,unsigned r,uint16_t old,uint16_t v,uint32_t copperPC){
 if(!rrcapture::trace.active)return;
 if(changes.size()>=32768){overflow=true;return;}
 changes.push_back({m.frames,m.cycles,m.line,m.beam,r,old,v,m.pc,copperPC,uint32_t(rrcapture::trace.writes.size())});
}
inline void line(const Machine&m){
 if(!rrcapture::trace.active||m.line<44||m.line>=44+H)return;
 auto&l=lines[m.line-44];l.regs=m.reg;l.pointers=m.bplPointer;l.sprites=m.sprites;l.spriteBefore=spriteCutoff;l.dma=m.dma;l.display=m.displayY;
 l.palette=m.linePalette;l.colors=m.lineColors;l.prepared=m.palettePrepared;l.frame=m.frames;l.clock=m.cycles;l.before=rrcapture::trace.writes.size();
 l.raster=m.line;l.changeStart=m.line==44?frameChangeStart:lineChangeStart;l.changeEnd=changes.size();l.paletteBefore=linePaletteBefore;l.valid=true;
}
inline bool complete(){return rrcapture::trace.valid&&!rrcapture::trace.overflow&&!overflow&&std::all_of(lines.begin(),lines.end(),[](const Line&l){return l.valid;});}
inline void seekMemory(unsigned before){
 const auto&t=rrcapture::trace;
 if(memory.empty()){memory.assign(t.initial.begin(),t.initial.begin()+std::min<size_t>(CHIP,t.initial.size()));cursor=0;}
 before=std::min<unsigned>(before,t.writes.size());
 auto apply=[&](const rrcapture::Write&w,uint32_t value){if((w.flags&1)&&w.address<CHIP)for(unsigned b=0;b<w.size&&w.address+b<CHIP;b++)memory[w.address+b]=value>>(8*b);};
 while(cursor>before){auto&w=t.writes[--cursor];apply(w,w.before);}
 while(cursor<before){auto&w=t.writes[cursor++];apply(w,w.after);}
}
inline uint32_t paletteCutoff(const Line&l,unsigned index,int x){
 uint32_t before=l.paletteBefore;
 for(unsigned i=l.changeStart;i<l.changeEnd;i++){const auto&c=changes[i];if(c.line==l.raster&&c.reg==0x180+index*2&&int(c.beam)*2-258<=x)before=c.before;}
 return before;
}
inline int blitBegin(const Machine&m,unsigned width,unsigned height){
 if(!rrcapture::trace.active)return -1;
 if(blits.size()>=4096){overflow=true;return -1;}
 Blit b;b.frame=m.frames;b.clock=m.cycles;b.line=m.line;b.beam=m.beam;b.pc=m.pc;b.event=hardwareEvent;b.before=rrcapture::trace.writes.size();
 b.con0=m.reg[0x40/2];b.con1=m.reg[0x42/2];b.firstMask=m.reg[0x44/2];b.lastMask=m.reg[0x46/2];b.width=width;b.height=height;b.first=words.size();
 b.pointers={m.ptr(0x50),m.ptr(0x4c),m.ptr(0x48),m.ptr(0x54)};
 b.mod={int16_t(m.reg[0x64/2])&~1,int16_t(m.reg[0x62/2])&~1,int16_t(m.reg[0x60/2])&~1,int16_t(m.reg[0x66/2])&~1};
 if(!changes.empty()&&changes.back().reg==0x58)b.copperPC=changes.back().copperPC;
 blits.push_back(b);return blits.size()-1;
}
inline void blitWord(int op,const BlitWord&w){if(op<0)return;if(words.size()>=maxWords){overflow=true;blits[op].complete=false;return;}words.push_back(w);blits[op].count++;}
inline size_t bytes(){size_t n=sizeof(lines)+sizeof(output)+words.size()*sizeof(BlitWord)+blits.size()*sizeof(Blit)+changes.size()*sizeof(Change);for(auto&l:lines)n+=l.colors.size()*sizeof(ColorChange);return n;}
}
}
