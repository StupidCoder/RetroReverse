#include "vic.h"
namespace rr::c64 {
void Vic::rasterIrq(){const bool match=raster==compare;if(match&&!irqMatch)flags|=1;irqMatch=match;}
uint8_t Vic::peek(uint8_t reg)const{
 reg&=63;switch(reg){case 0x11:return (regs[reg]&127)|((raster&256)?128:0);case 0x12:return uint8_t(raster);
 case 0x16:return regs[reg]|0xc0;case 0x18:return regs[reg]|1;case 0x19:return flags|0x70|(irq()?128:0);case 0x1a:return mask|0xf0;
 case 0x1e:return collisionSprites;case 0x1f:return collisionGraphics;default:return reg>=0x2f?255:reg>=0x20?regs[reg]|0xf0:regs[reg];}
}
uint8_t Vic::read(uint8_t reg){const auto value=peek(reg);if((reg&63)==0x1e)collisionSprites=0;if((reg&63)==0x1f)collisionGraphics=0;return value;}
void Vic::write(uint8_t reg,uint8_t value){
 reg&=63;if(reg>=0x2f||reg==0x13||reg==0x14||reg==0x1e||reg==0x1f)return;
 if(reg==0x19){flags&=uint8_t(~value);return;}if(reg==0x1a){mask=value&15;return;}
 if(reg==0x11)compare=(compare&255)|((value&128)<<1);if(reg==0x12)compare=(compare&256)|value;
 if(reg==0x17)for(unsigned i=0;i<8;i++)if(!(value&(1<<i)))sprites[i].advance=true;
 regs[reg]=reg>=0x20?value&15:value;
 if(reg==0x11||reg==0x12)rasterIrq();
}
void Vic::tick(std::span<const uint8_t,65536> ram,std::span<const uint8_t,4096> chars,std::span<const uint8_t,1024> color,uint16_t bank,VicObserver* observer){
 ++clocks;fetchCount=0;if(++cycle==64){cycle=1;if(++raster==312){raster=0;++frames;}}
 if(cycle==1){if(!raster){base=0;den=false;}else rasterIrq();}if(!raster&&cycle==2)rasterIrq();
 if(raster==0x30&&(regs[0x11]&16))den=true;
 bad=den&&raster>=0x30&&raster<=0xf7&&(raster&7)==(regs[0x11]&7);if(bad)display=true;
 if(cycle==14){vc=base;index=0;if(bad)row=0;}
 for(unsigned i=0;i<8;i++){
  auto& s=sprites[i];const auto bit=1<<i;if(!(regs[0x17]&bit))s.advance=true;
  if((cycle==55||cycle==56)&&(regs[0x15]&bit)&&uint8_t(raster)==regs[2*i+1]&&!s.dma){s.dma=true;s.base=0;s.advance=true;}
  if(cycle==56&&(regs[0x17]&bit)&&s.dma)s.advance=!s.advance;
  if(cycle==58){s.mc=s.base;if(s.dma&&uint8_t(raster)==regs[2*i+1])s.display=true;else if(!s.dma)s.display=false;}
  if(cycle==16){if(s.advance)s.base=s.mc;if(s.base==63)s.dma=false;}
 }
 const auto steal=[&](unsigned c){if(bad&&c>=15&&c<=54)return true;for(unsigned i=0;i<8;i++){const unsigned start=(57+2*i)%63+1;if(sprites[i].dma&&(c==start||c==start+1))return true;}return false;};
 aec=!steal(cycle);ba=true;for(unsigned ahead=0;ahead<=3;ahead++)if(steal((cycle-1+ahead)%63+1))ba=false;
 const auto fetch=[&](uint16_t address,VicAccess kind,uint8_t phase,uint8_t slot=0){
  VicFetch f;f.cycle=clocks;f.address=uint16_t(bank|(address&0x3fff));f.kind=kind;f.phase=phase;f.slot=slot;
  f.rom=(f.address&0x7000)==0x1000;f.value=f.rom?chars[f.address&4095]:ram[f.address];
  if(kind==VicAccess::Matrix)f.color=color[address&1023]&15;
  fetches[fetchCount++]=f;if(observer)observer->fetch(f);return f;
 };
 // Sprite pointer/data slots: 0..2 at 58/60/62; 3..7 at 1/3/5/7/9.
 int sprite=-1;bool second=false;
 if(cycle>=58){sprite=(cycle-58)/2;second=(cycle-58)&1;}
 else if(cycle<=10){sprite=3+(cycle-1)/2;second=!(cycle&1);}
 if(sprite>=0){auto& s=sprites[unsigned(sprite)];
  if(!second){s.pointer=fetch(uint16_t(((regs[0x18]&0xf0)<<6)|0x3f8|sprite),VicAccess::Pointer,1,uint8_t(sprite)).value;
   if(s.dma)s.bytes[0]=fetch(uint16_t((s.pointer<<6)|s.mc++),VicAccess::Sprite,2,uint8_t(sprite));
  }else if(s.dma){s.bytes[1]=fetch(uint16_t((s.pointer<<6)|s.mc++),VicAccess::Sprite,1,uint8_t(sprite));s.bytes[2]=fetch(uint16_t((s.pointer<<6)|s.mc++),VicAccess::Sprite,2,uint8_t(sprite));s.mc&=63;
   s.shift=(uint32_t(s.bytes[0].value)<<16)|(uint32_t(s.bytes[1].value)<<8)|s.bytes[2].value;s.pixel=0;s.repeat=0;s.active=false;
  }else fetch(0x3fff,VicAccess::Idle,1);
 }else if(cycle>=11&&cycle<=15)fetch(uint16_t(0x3f00|refresh--),VicAccess::Refresh,1);
 else if(cycle>=16&&cycle<=55){const unsigned column=cycle-16;auto& cell=cells[column];
  uint16_t address=(regs[0x11]&64)?0x39ff:0x3fff;
  if(display){if(regs[0x11]&32)address=uint16_t(((regs[0x18]&8)<<10)|(vc<<3)|row);
   else address=uint16_t(((regs[0x18]&14)<<10)|(uint16_t(cell.matrix.value)<<3)|row);
   if(regs[0x11]&64)address&=uint16_t(~0x600);
  }
  cell.graphics=fetch(address,VicAccess::Graphics,1,uint8_t(column));cell.code=display?cell.matrix.value:0;cell.color=display?cell.matrix.color:0;
  if(display){vc=(vc+1)&1023;++index;}
 }else fetch(0x3fff,VicAccess::Idle,1);
 if(bad&&cycle>=15&&cycle<=54){const auto column=cycle-15;cells[column].matrix=fetch(uint16_t(((regs[0x18]&0xf0)<<6)|vc),VicAccess::Matrix,2,uint8_t(column));}
 draw(observer);
 if(cycle==58){if(row==7){base=vc;display=bad;}if(display)row=(row+1)&7;}
}
void Vic::draw(VicObserver* observer){
 const unsigned start=((cycle-1)*8+396)%Width; // Common eight-dot output delay removed.
 const unsigned left=regs[0x16]&8?24:31,right=regs[0x16]&8?344:335;
 const unsigned top=regs[0x11]&8?51:55,bottom=regs[0x11]&8?251:247;
 for(unsigned dot=0;dot<8;dot++){
  const unsigned x=(start+dot)%Width;
  if(x==right)border=true;
  if(x==left){if(raster==bottom)verticalBorder=true;if(raster==top&&(regs[0x11]&16))verticalBorder=false;if(!verticalBorder)border=false;}
  uint8_t ink=regs[0x21];bool foreground=false;
  const int gx=int(x)-24-(regs[0x16]&7);
  if(gx>=0&&gx<320&&!verticalBorder){const auto& cell=cells[unsigned(gx)/8];const unsigned bit=7-(unsigned(gx)&7);const auto data=cell.graphics.value;
   const bool ecm=regs[0x11]&64,bitmap=regs[0x11]&32,multi=(regs[0x16]&16)&&(bitmap||(cell.color&8));
   if(multi){const auto pair=(data>>((bit/2)*2))&3;foreground=pair>=2;
    if(bitmap){const uint8_t colors[]={regs[0x21],uint8_t(cell.code>>4),uint8_t(cell.code&15),cell.color};ink=colors[pair];}
    else{const uint8_t colors[]={regs[0x21],regs[0x22],regs[0x23],uint8_t(cell.color&7)};ink=colors[pair];}
   }else{foreground=(data>>bit)&1;ink=foreground?(bitmap?cell.code>>4:((regs[0x16]&16)?cell.color&7:cell.color)):(bitmap?cell.code&15:regs[0x21+(ecm?(cell.code>>6):0)]);}
   if(ecm&&(bitmap||(regs[0x16]&16)))ink=0;
  }
  const uint8_t background=ink;std::array<uint8_t,8> positions{};
  uint8_t hits=0,spriteInk=0;int winner=-1;
  for(unsigned i=0;i<8;i++){auto& s=sprites[i];const auto bit=1<<i;positions[i]=s.pixel;
   const unsigned sx=regs[2*i]|((regs[0x10]&bit)?256:0);
   if(s.display&&x==sx&&s.pixel==0)s.active=true;
   if(!s.active||!s.display||s.pixel>=24)continue;
   const bool multi=regs[0x1c]&bit;const unsigned value=multi?(s.shift>>(22-(s.pixel/2)*2))&3:(s.shift>>(23-s.pixel))&1;
   if(value){hits|=uint8_t(bit);if(winner<0){winner=int(i);spriteInk=multi?(value==1?regs[0x25]:value==2?regs[0x27+i]:regs[0x26]):regs[0x27+i];}}
   if(++s.repeat==((regs[0x1d]&bit)?2:1)){s.repeat=0;if(++s.pixel==24)s.active=false;}
  }
  if(hits&&(hits&(hits-1))){if(!collisionSprites)flags|=4;collisionSprites|=hits;}
  if(hits&&foreground){if(!collisionGraphics)flags|=2;collisionGraphics|=hits;}
  if(winner>=0&&(!(regs[0x1b]&(1<<winner))||!foreground))ink=spriteInk;
  pixels[raster*Width+x]=border?regs[0x20]:ink;
  if(observer)observer->pixel(*this,{x,pixels[raster*Width+x],uint8_t(border?regs[0x20]:background),spriteInk,hits,winner,foreground,border,positions});
 }
 if(cycle==63){if(raster==bottom)verticalBorder=true;if(raster==top&&(regs[0x11]&16))verticalBorder=false;}
}
}
