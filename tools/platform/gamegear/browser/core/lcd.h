#pragma once
#include "../../../../browser/handheld/common.h"
inline uint32_t ggColor(const gamegear_Machine*m,unsigned entry){auto&v=m->VDP;uint16_t c=v.CRAM[entry*2]|uint16_t(v.CRAM[entry*2+1])<<8;return 0xff000000|uint32_t(c&15)*17|(uint32_t((c>>4)&15)*17<<8)|(uint32_t((c>>8)&15)*17<<16);}
inline uint32_t ggBits(const gamegear_Machine*m,unsigned at){uint32_t b=0;for(unsigned i=0;i<4;i++)b|=uint32_t(m->VDP.VRAM[(at+i)&0x3fff])<<(8*i);return b;}
inline unsigned ggTexel(uint32_t bits,unsigned bit){return ((bits>>bit)&1)|((bits>>(bit+7))&2)|((bits>>(bit+14))&4)|((bits>>(bit+21))&8);}
// Capture addresses describe the VDP's separate address spaces.
struct GGPixelSource {uint32_t map=0,mapValue=0,spriteY=0,spriteYValue=0,spriteXT=0,spriteXTValue=0;};
template<class Event,class Dot>
inline uint8_t rrGGRenderLine(const gamegear_Machine*m,int line,uint8_t scrollXLatched,uint8_t scrollYLatched,Event event,Dot dot){
 if(line<0||line>=192)return 0;const auto&v=m->VDP;uint8_t status=0;
 int y=line-24;bool visible=y>=0&&y<144;
 std::array<uint8_t,256>bg{},priority{},claimed{};unsigned nt=(v.Regs[2]&14)<<10;
 if(visible)event("Background scanline",y,bool(v.Regs[1]&64));
 for(int sx=0;sx<256;sx++){
  unsigned scrollX=(v.Regs[0]&64)&&line<16?0:scrollXLatched;unsigned bx=(sx-scrollX)&255,by=((v.Regs[0]&128)&&sx>=192?line:line+scrollYLatched)%224;
  unsigned entry=(nt+(by/8)*64+(bx/8)*2)&0x3fff;uint16_t name=v.VRAM[entry]|uint16_t(v.VRAM[(entry+1)&0x3fff])<<8;
  unsigned tx=bx&7,ty=by&7;if(name&0x200)tx=7-tx;if(name&0x400)ty=7-ty;unsigned address=(name&511)*32+ty*4;uint32_t bits=ggBits(m,address);unsigned color=ggTexel(bits,7-tx),pal=(name&0x800?16:0)+color;bg[sx]=color;priority[sx]=(name&0x1000)&&color;
  bool shown=(v.Regs[1]&64)&&(!(v.Regs[0]&32)||sx>=8);if(!shown)pal=16+(v.Regs[7]&15);
  if(visible)dot(sx-48,y,ggColor(m,pal),shown?0x10000+address:0,bits,0x14000+pal*2,v.CRAM[pal*2]|uint16_t(v.CRAM[pal*2+1])<<8,color,tx,ty,1,GGPixelSource{shown?0x10000+entry:0,name});
 }
 if(!(v.Regs[1]&64))return 0;
 unsigned sat=(v.Regs[5]&0x7e)<<7;int height=v.Regs[1]&2?16:8,scale=v.Regs[1]&1?2:1,count=0;
 for(int i=0;i<64;i++){
  uint8_t yy=v.VRAM[(sat+i)&0x3fff];if(yy==0xd0)break;int sy=int(yy)+1;if(sy>=240)sy-=256;int row=line-sy;if(row<0||row>=height*scale)continue;
  if(++count>8){status|=0x40;break;}row/=scale;
  int sx=v.VRAM[(sat+128+i*2)&0x3fff]-(v.Regs[0]&8?8:0);unsigned tile=v.VRAM[(sat+129+i*2)&0x3fff]|(v.Regs[6]&4?256:0);if(height==16)tile&=~1u;unsigned address=tile*32+row*4;uint32_t bits=ggBits(m,address);
  if(visible)event("Sprite scanline",y,true,i,0x10000+sat+i);
  for(int px=0;px<8*scale;px++){
   int x=sx+px;if(x<0||x>=256)continue;unsigned color=ggTexel(bits,7-px/scale),pal=16+color;uint8_t flags=1;
   if(!color)flags=4;else if(claimed[x]){flags=8;status|=0x20;}else{claimed[x]=1;if(priority[x])flags=2;}
   if(visible)dot(x-48,y,ggColor(m,pal),0x10000+address,bits,0x14000+pal*2,v.CRAM[pal*2]|uint16_t(v.CRAM[pal*2+1])<<8,color,px/scale,row,flags,GGPixelSource{0,0,0x10000+((sat+i)&0x3fff),yy,0x10000+((sat+128+i*2)&0x3fff),uint32_t(v.VRAM[(sat+128+i*2)&0x3fff])|(uint32_t(v.VRAM[(sat+129+i*2)&0x3fff])<<8)});
  }
 }
 return status;
}

#include "timing.h"
#include "raster-state.h"
inline void rrGGLine(gamegear_Machine*m,int line,uint8_t scrollX,uint8_t scrollY){
 if(line<0||line>=192)return;rrprof::Scope measured(1,"VDP tiles and sprites");
 if(line==0)rrhh::clear(ggColor(m,16+(m->VDP.Regs[7]&15)));
 int y=line-24;if(y>=0&&y<144)rrgg::recordStart(m,y,scrollX,scrollY);
 struct Events {void operator()(const char*k,int y,bool source=true,int i=-1,uint32_t descriptor=0)const{rrhh::event(k,y,source,i,descriptor);}};
 struct Dots {void operator()(int x,int y,uint32_t color,uint32_t source,uint32_t bits,uint32_t pal,uint32_t palette,unsigned texel,int u,int v,uint8_t flags,const GGPixelSource&)const{rrhh::dot(x,y,color,source,bits,pal,palette,texel,u,v,flags);}};
 m->VDP.status|=rrGGRenderLine(m,line,scrollX,scrollY,Events{},Dots{});
 if(y>=0&&y<144)rrgg::recordEnd(y);
 rrcapture::trace.current=0;
}
inline void rrGGLine(gamegear_Machine*m,int line){rrGGLine(m,line,m->VDP.Regs[8],m->VDP.Regs[9]);}
