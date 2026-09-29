#pragma once
#include "inspection.h"
#include <memory>
namespace rramiga::inspect {
inline std::unique_ptr<Machine> scratch;
inline std::array<std::vector<uint32_t>,6> blitPictures;
inline int planeView=-1;
inline const char*regName(unsigned r){
 switch(r){case 0x100:return "BPLCON0";case 0x102:return "BPLCON1";case 0x104:return "BPLCON2";case 0x108:return "BPL1MOD";case 0x10a:return "BPL2MOD";case 0x8e:return "DIWSTRT";case 0x90:return "DIWSTOP";case 0x92:return "DDFSTRT";case 0x94:return "DDFSTOP";case 0x58:return "BLTSIZE";default:return "Register";}
}
inline std::string label(unsigned r){if(r>=0x180&&r<0x1c0)return "COLOR"+std::to_string((r-0x180)/2);if(r>=0xe0&&r<0xf8)return "BPL"+std::to_string((r-0xe0)/4+1)+(r&2?"PTL":"PTH");return regName(r);}
inline bool videoRegister(unsigned r){return (r>=0x180&&r<0x1c0)||(r>=0xe0&&r<0x10c)||(r>=0x8e&&r<=0x94)||(r>=0x120&&r<0x180);}
inline const char*kind(const Blit&b){return b.con1&1?"Line":(b.con0&255)==0xca?"Cookie cut":(b.con0&255)==0xf0?"Copy A":(b.con0&255)==0xcc?"Copy B":(b.con0&255)==0?"Clear":"Boolean blit";}
inline unsigned countChanges(const Line&l){unsigned n=0;for(unsigned i=l.changeStart;i<l.changeEnd;i++)if(videoRegister(changes[i].reg))n++;return n;}
inline void loadLine(const Line&l,int y,bool frozen=false){
 if(!scratch)scratch=std::make_unique<Machine>();auto&m=*scratch;
 m.reg=l.regs;m.bplPointer=l.pointers;m.sprites=l.sprites;m.dma=l.dma;m.displayY=l.display;m.line=y+44;
 m.linePalette=l.palette;m.lineColors=l.colors;m.palettePrepared=l.prepared;
 if(frozen){m.palettePrepared=false;m.lineColors.clear();}
}
inline void frozenRow(const Line&l,int anchor,int y){
 loadLine(l,y,true);auto&m=*scratch;
 unsigned words=((l.regs[0x94/2]&0xf8)>=(l.regs[0x92/2]&0xf8))?((l.regs[0x94/2]&0xf8)-(l.regs[0x92/2]&0xf8))/8+1:0;
 if(l.regs[0x100/2]&0x8000)words*=2;words=std::min(words,64u);
 unsigned top=l.regs[0x8e/2]>>8,bottom=l.regs[0x90/2]>>8;if(!(bottom&128))bottom+=256;
 int sampled=std::clamp(anchor+44,int(top),int(bottom));
 for(unsigned p=0;p<6;p++)m.bplPointer[p]=l.pointers[p]+(y+44-sampled)*(int(words*2)+int16_t(l.regs[(p&1?0x10a:0x108)/2]));
 for(unsigned i=0;i<8;i++){
  auto&s=m.sprites[i];unsigned start=(s.pos>>8)|((s.ctl&4)<<6),end=(s.ctl>>8)|((s.ctl&2)<<7);
  uint32_t base=s.pointer-(l.sprites[i].active?unsigned(anchor+44-start+1)*4:0);
  s.active=unsigned(y+44)>=start&&unsigned(y+44)<end;
  if(s.active){s.pointer=base+(y+44-start+1)*4;s.a=m.chip16(s.pointer-4);s.b=m.chip16(s.pointer-2);}
 }
}
inline void sourcesJSON(std::ostream&o,const std::vector<Source>&sources){
 o<<'[';bool comma=false;for(auto&s:sources){if(comma)o<<',';comma=true;o<<"{\"role\":\""<<s.role<<"\",\"address\":"<<s.address<<",\"size\":"<<s.size<<",\"value\":"<<s.value<<",\"before\":"<<s.before<<'}';}o<<']';
}
inline std::string info(){
 std::ostringstream o;o<<"{\"available\":"<<(rrcapture::trace.valid?"true":"false")<<",\"complete\":"<<(complete()?"true":"false")<<",\"bytes\":"<<bytes()<<",\"lines\":[";bool comma=false;
 if(rrcapture::trace.valid)for(unsigned y=0;y<H;y++)if(lines[y].valid){if(comma)o<<',';comma=true;o<<"{\"line\":"<<y<<",\"changes\":"<<countChanges(lines[y])<<'}';}
 o<<"],\"blits\":[";comma=false;if(rrcapture::trace.valid)for(unsigned i=0;i<blits.size();i++){auto&b=blits[i];if(comma)o<<',';comma=true;o<<"{\"index\":"<<i<<",\"kind\":\""<<kind(b)<<"\",\"frame\":"<<b.frame<<",\"line\":"<<b.line<<",\"width\":"<<b.width*16<<",\"height\":"<<b.height<<",\"destination\":"<<b.pointers[3]<<'}';}o<<"]}";return o.str();
}
inline std::string seek(int y){
 if(!rrcapture::trace.valid||y<0||y>=int(H)||!lines[y].valid)return "{\"error\":\"Capture a complete display first.\"}";
 selected=y;auto&l=lines[y];seekMemory(l.before);loadLine(l,y);std::copy(memory.begin(),memory.end(),scratch->ram.begin());
 for(auto&p:pictures)p.fill(0);
 for(int row=0;row<int(H);row++){frozenRow(l,y,row);scratch->renderLine(true,pictures[0].data(),pictures[1].data(),planeView);}
 for(int row=0;row<=y;row++)if(lines[row].valid)std::copy_n(output.begin()+row*W,W,pictures[2].begin()+row*W);
 std::ostringstream o;o<<"{\"line\":"<<y<<",\"raster\":"<<y+44<<",\"frame\":"<<l.frame<<",\"complete\":"<<(complete()?"true":"false")<<",\"planes\":"<<std::min(6,(l.regs[0x100/2]>>12)&7)<<",\"mode\":\""<<(l.regs[0x100/2]&0x400?"Dual playfield":l.regs[0x100/2]&0x800?"HAM6":((l.regs[0x100/2]>>12)&7)==6?"Extra half-bright":"Indexed color")<<"\",\"registers\":[";bool comma=false;
 for(unsigned r:{0x100,0x102,0x104,0x108,0x10a,0x8e,0x90,0x92,0x94}){if(comma)o<<',';comma=true;o<<"{\"name\":\""<<label(r)<<"\",\"value\":"<<l.regs[r/2]<<'}';}
 o<<"],\"pointers\":[";for(unsigned p=0;p<6;p++){if(p)o<<',';o<<(l.pointers[p]&(CHIP-1));}
 o<<"],\"palette\":[";for(unsigned p=0;p<32;p++){if(p)o<<',';o<<l.regs[0x180/2+p];}
 o<<"],\"changes\":[";comma=false;for(unsigned i=l.changeStart;i<l.changeEnd;i++){auto&c=changes[i];if(!videoRegister(c.reg))continue;if(comma)o<<',';comma=true;o<<"{\"name\":\""<<label(c.reg)<<"\",\"register\":"<<c.reg<<",\"line\":"<<c.line<<",\"before\":"<<c.oldValue<<",\"value\":"<<c.value<<",\"beam\":"<<c.beam<<",\"x\":"<<int(c.beam)*2-258<<",\"pc\":"<<c.pc<<",\"copperPC\":"<<c.copperPC<<",\"cutoff\":"<<c.before<<'}';}o<<"]}";return o.str();
}
inline std::string pixel(int panel,int x,int y){
 if(selected<0||!rrcapture::trace.valid||panel<0||panel>2||x<0||x>=int(W)||y<0||y>=int(H))return "{\"error\":\"Select a recorded line.\"}";
 if(panel==2&&y>selected)return "{\"error\":\"This line has not been drawn yet.\"}";
 int anchor=panel==2?y:selected;auto&l=lines[anchor];if(!l.valid)return "{\"error\":\"Missing scanline.\"}";
 seekMemory(l.before);loadLine(l,y);std::copy(memory.begin(),memory.end(),scratch->ram.begin());
 if(panel<2)frozenRow(l,anchor,y);
 RenderProbe p;p.x=x;p.y=y;p.anchor=anchor;p.before=l.before;p.frozen=panel<2;
 scratch->renderLine(true,nullptr,nullptr,panel==0?planeView:-1,&p);
 std::ostringstream o;o<<"{\"panel\":"<<panel<<",\"x\":"<<x<<",\"y\":"<<y<<",\"stateLine\":"<<anchor<<",\"preview\":"<<(panel<2?"true":"false")<<",\"complete\":"<<(complete()&&(panel!=2||scratch->draw[y*W+x]==output[y*W+x])?"true":"false")<<",\"ham\":"<<((l.regs[0x100/2]&0x800)?"true":"false")<<",\"candidates\":[";bool comma=false;
 for(auto&c:p.candidates){if((panel==0&&c.object>=0)||(panel==1&&c.object<0))continue;if(comma)o<<',';comma=true;o<<"{\"kind\":\""<<c.kind<<"\",\"object\":"<<c.object<<",\"flags\":"<<unsigned(c.flags)<<",\"color\":"<<c.color<<",\"sources\":";sourcesJSON(o,c.sources);o<<'}';}o<<"]}";return o.str();
}
inline std::string blitSeek(int index){
 if(!rrcapture::trace.valid||index<0||unsigned(index)>=blits.size())return "{\"error\":\"No captured blit.\"}";
 selectedBlit=index;auto&b=blits[index];unsigned width=b.con1&1?16:b.width*16,height=b.height;
 for(unsigned p=0;p<4;p++)blitPictures[p].assign(width*height,0);
 for(unsigned i=0;i<b.count;i++){auto&w=words[b.first+i];unsigned row=b.con1&1?i:i/b.width,column=b.con1&1?0:i%b.width;if(row>=height)break;
  if((b.con1&2)&&!(b.con1&1)){row=height-1-row;column=b.width-1-column;}
  for(unsigned bit=0;bit<16;bit++){unsigned at=row*width+column*16+bit;uint16_t values[]={w.a,w.b,w.c,w.d};for(unsigned p=0;p<4;p++)blitPictures[p][at]=(values[p]&(0x8000>>bit))?(p==0?0xff777d28:0xffeae6dd):0xff463228;}
 }
 // These previews use the final frame's DMA layout and the memory at this blit.
 // They expose buffer production, not an assertion that it was visible then.
 for(unsigned stage=0;stage<2;stage++){
  seekMemory(stage?b.end:b.before);if(!scratch)scratch=std::make_unique<Machine>();std::copy(memory.begin(),memory.end(),scratch->ram.begin());
  auto&image=blitPictures[4+stage];image.assign(W*H,0);
  for(unsigned y=0;y<H;y++)if(lines[y].valid){loadLine(lines[y],y);scratch->dma&=~0x20;scratch->renderLine(true,image.data());}
 }
 std::ostringstream o;o<<"{\"index\":"<<index<<",\"count\":"<<blits.size()<<",\"kind\":\""<<kind(b)<<"\",\"width\":"<<width<<",\"height\":"<<height<<",\"frame\":"<<b.frame<<",\"line\":"<<b.line<<",\"beam\":"<<b.beam<<",\"pc\":"<<b.pc<<",\"copperPC\":"<<b.copperPC<<",\"con0\":"<<b.con0<<",\"con1\":"<<b.con1<<",\"firstMask\":"<<b.firstMask<<",\"lastMask\":"<<b.lastMask<<",\"complete\":"<<(complete()&&b.complete?"true":"false")<<",\"pointers\":[";
 for(unsigned p=0;p<4;p++){if(p)o<<',';o<<(b.pointers[p]&(CHIP-1));}o<<"],\"modulos\":[";for(unsigned p=0;p<4;p++){if(p)o<<',';o<<b.mod[p];}o<<"]}";return o.str();
}
inline std::string blitPixel(int x,int y){
 if(selectedBlit<0||!rrcapture::trace.valid)return "{\"error\":\"Select a blit first.\"}";auto&b=blits[selectedBlit];unsigned width=b.con1&1?16:b.width*16;
 if(x<0||unsigned(x)>=width||y<0||unsigned(y)>=b.height)return "{\"error\":\"Outside the blit.\"}";
 unsigned row=y,col=x/16;if((b.con1&2)&&!(b.con1&1)){row=b.height-1-row;col=b.width-1-col;}
 unsigned at=b.con1&1?row:row*b.width+col;if(at>=b.count)return "{\"error\":\"This word was not captured.\"}";
 auto&w=words[b.first+at];std::vector<Source> src;
 const uint16_t values[]={w.rawA,w.rawB,w.c};
 for(unsigned p=0;p<3;p++){
  // Line drawing synthesizes A and B from the data registers even when the
  // corresponding channel enable bits are set; no A/B memory fetch occurs.
  bool enabled=(b.con0&(0x800>>p))&&(!(b.con1&1)||p==2);unsigned addr=enabled?w.address[p]:registerBase+(p==0?0x74:p==1?0x72:0x70);
  src.push_back({std::string(1,'A'+p)+(enabled?" fetched word":" data register"),addr,swapped(values[p]),w.before});
  if(p<2&&at&&((p?b.con1:b.con0)>>12)&&!(b.con1&1)&&enabled){auto&prev=words[b.first+at-1];src.push_back({std::string(1,'A'+p)+" previous word (shift carry)",prev.address[p],swapped(p?prev.rawB:prev.rawA),prev.before});}
 }
 src.push_back({"Destination before",w.address[3],swapped(w.oldD),w.before});
 if(w.written)src.push_back({"Destination result",w.address[3],swapped(w.d),w.write});
 std::ostringstream o;o<<"{\"x\":"<<x<<",\"y\":"<<y<<",\"word\":"<<at<<",\"bit\":"<<15-(x&15)<<",\"a\":"<<w.a<<",\"b\":"<<w.b<<",\"c\":"<<w.c<<",\"d\":"<<w.d<<",\"oldD\":"<<w.oldD<<",\"written\":"<<(w.written?"true":"false")<<",\"sources\":";sourcesJSON(o,src);o<<'}';return o.str();
}
}
extern "C" {
const char*rr_raster_info(){reply=rramiga::inspect::info();return reply.c_str();}
const char*rr_raster_seek(int y){reply=rramiga::inspect::seek(y);return reply.c_str();}
void rr_raster_plane(int plane){rramiga::inspect::planeView=std::clamp(plane,-1,5);}
uint8_t*rr_raster_frame(int p){return p>=0&&p<3?reinterpret_cast<uint8_t*>(rramiga::inspect::pictures[p].data()):nullptr;}
const char*rr_raster_pixel(int p,int x,int y){reply=rramiga::inspect::pixel(p,x,y);return reply.c_str();}
const char*rr_blit_seek(int index){reply=rramiga::inspect::blitSeek(index);return reply.c_str();}
uint8_t*rr_blit_frame(int p){return p>=0&&p<6?reinterpret_cast<uint8_t*>(rramiga::inspect::blitPictures[p].data()):nullptr;}
const char*rr_blit_pixel(int x,int y){reply=rramiga::inspect::blitPixel(x,y);return reply.c_str();}
}
