#pragma once
#include "../../../../browser/handheld/common.h"
inline uint32_t ggColor(gamegear_Machine*m,unsigned entry){auto&v=m->VDP;uint16_t c=v.CRAM[entry*2]|uint16_t(v.CRAM[entry*2+1])<<8;return 0xff000000|uint32_t(c&15)*17|(uint32_t((c>>4)&15)*17<<8)|(uint32_t((c>>8)&15)*17<<16);}
inline uint32_t ggBits(gamegear_Machine*m,unsigned at){uint32_t b=0;for(unsigned i=0;i<4;i++)b|=uint32_t(m->VDP.VRAM[(at+i)&0x3fff])<<(8*i);return b;}
inline unsigned ggTexel(uint32_t bits,unsigned bit){return ((bits>>bit)&1)|((bits>>(bit+7))&2)|((bits>>(bit+14))&4)|((bits>>(bit+21))&8);}
inline void rrGGLine(gamegear_Machine*m,int line){
 if(line<0||line>=192)return;rrprof::Scope measured(1,"VDP tiles and sprites");auto&v=m->VDP;
 if(line==0)rrhh::clear(ggColor(m,16+(v.Regs[7]&15)));
 int y=line-24;bool visible=y>=0&&y<144;
 std::array<uint8_t,256>bg{},priority{},claimed{};unsigned nt=(v.Regs[2]&14)<<10;
 if(visible)rrhh::event("Background scanline",y,bool(v.Regs[1]&64));
 for(int sx=0;sx<256;sx++){
  unsigned scrollX=(v.Regs[0]&64)&&line<16?0:v.Regs[8];unsigned bx=(sx-scrollX)&255,by=((v.Regs[0]&128)&&sx>=192?line:line+v.Regs[9])%224;
  unsigned entry=(nt+(by/8)*64+(bx/8)*2)&0x3fff;uint16_t name=v.VRAM[entry]|uint16_t(v.VRAM[(entry+1)&0x3fff])<<8;
  unsigned tx=bx&7,ty=by&7;if(name&0x200)tx=7-tx;if(name&0x400)ty=7-ty;unsigned address=(name&511)*32+ty*4;uint32_t bits=ggBits(m,address);unsigned color=ggTexel(bits,7-tx),pal=(name&0x800?16:0)+color;bg[sx]=color;priority[sx]=(name&0x1000)&&color;
  bool shown=(v.Regs[1]&64)&&(!(v.Regs[0]&32)||sx>=8);if(!shown)pal=16+(v.Regs[7]&15);
  if(visible)rrhh::dot(sx-48,y,ggColor(m,pal),shown?0x10000+address:0,bits,0x14000+pal*2,v.CRAM[pal*2]|uint16_t(v.CRAM[pal*2+1])<<8,color,tx,ty);
 }
 if(!(v.Regs[1]&64))return;
 unsigned sat=(v.Regs[5]&0x7e)<<7;int height=v.Regs[1]&2?16:8,scale=v.Regs[1]&1?2:1,count=0;
 for(int i=0;i<64;i++){
  uint8_t yy=v.VRAM[(sat+i)&0x3fff];if(yy==0xd0)break;int sy=int(yy)+1;if(sy>=240)sy-=256;int row=line-sy;if(row<0||row>=height*scale)continue;
  if(++count>8){v.status|=0x40;break;}row/=scale;
  int sx=v.VRAM[(sat+128+i*2)&0x3fff]-(v.Regs[0]&8?8:0);unsigned tile=v.VRAM[(sat+129+i*2)&0x3fff]|(v.Regs[6]&4?256:0);if(height==16)tile&=~1u;unsigned address=tile*32+row*4;uint32_t bits=ggBits(m,address);
  if(visible)rrhh::event("Sprite scanline",y,true,i,0x10000+sat+i);
  for(int px=0;px<8*scale;px++){
   int x=sx+px;if(x<0||x>=256)continue;unsigned color=ggTexel(bits,7-px/scale),pal=16+color;uint8_t flags=1;
   if(!color)flags=4;else if(claimed[x]){flags=8;v.status|=0x20;}else{claimed[x]=1;if(priority[x])flags=2;}
   if(visible)rrhh::dot(x-48,y,ggColor(m,pal),0x10000+address,bits,0x14000+pal*2,v.CRAM[pal*2]|uint16_t(v.CRAM[pal*2+1])<<8,color,px/scale,row,flags);
  }
 }
 rrcapture::trace.current=0;
}
