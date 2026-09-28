#pragma once
#include "../../../../browser/handheld/common.h"
inline uint32_t gbShade(uint8_t palette,unsigned texel){uint32_t c=255-85*((palette>>(texel*2))&3);return 0xff000000|c*0x010101;}
inline void rrGBLine(gameboy_Machine*m,int y){
 if(y<0||y>=144)return;rrprof::Scope measured(1,"LCD tiles and sprites");auto&v=rrhh::video;
 auto lcd=m->io[0x40];if(y==0){v.windowLine=0;rrhh::clear(0xffffffff);}
 std::array<uint8_t,160>bg{},claimed{};
 bool bgOn=lcd&1,window=bgOn&&(lcd&0x20)&&y>=m->io[0x4a]&&m->io[0x4b]<=166;
 int wx=int(m->io[0x4b])-7;
 rrhh::event("Background / window scanline",y,bgOn);
 for(int x=0;x<160;x++){
  bool win=window&&x>=wx;int bx=win?x-wx:(x+m->io[0x43])&255,by=win?v.windowLine:(y+m->io[0x42])&255;
  uint32_t nt=(lcd&(win?0x40:8))?0x1c00:0x1800;uint8_t idx=m->vram[nt+(by/8)*32+bx/8];
  uint32_t address=(lcd&0x10)?idx*16:0x1000+int8_t(idx)*16;address+=(by&7)*2;unsigned bit=7-(bx&7);
  uint16_t bits=m->vram[address]|uint16_t(m->vram[address+1])<<8;unsigned color=bgOn?((bits>>bit)&1)|((bits>>(bit+7))&2):0;bg[x]=color;
  rrhh::dot(x,y,bgOn?gbShade(m->io[0x47],color):0xffffffff,bgOn?0x8000+address:0,bits,0xff47,m->io[0x47],color,bx&7,by&7);
 }
 if(window&&wx<160)v.windowLine++;
 if(!(lcd&2))return;
 int height=lcd&4?16:8;std::array<int,10>sprites{};int count=0;
 for(int i=0;i<40&&count<10;i++){int sy=int(m->oam[i*4])-16;if(y>=sy&&y<sy+height)sprites[count++]=i;}
 // DMG object priority: X first, then OAM index. The winning object blocks
 // lower-priority objects even when its BG-priority bit hides it behind BG.
 std::stable_sort(sprites.begin(),sprites.begin()+count,[&](int a,int b){return m->oam[a*4+1]<m->oam[b*4+1];});
 for(int n=0;n<count;n++){int i=sprites[n],sx=int(m->oam[i*4+1])-8,sy=int(m->oam[i*4])-16;uint8_t tile=m->oam[i*4+2],attr=m->oam[i*4+3];int row=y-sy;if(attr&64)row=height-1-row;if(height==16)tile&=254;uint32_t address=tile*16+row*2;uint16_t bits=m->vram[address]|uint16_t(m->vram[address+1])<<8;unsigned pal=attr&16?0x49:0x48;rrhh::event("Object scanline",y,true,i,0xfe00+i*4);
  for(int px=0;px<8;px++){int x=sx+px;if(x<0||x>=160)continue;unsigned bit=attr&32?px:7-px,color=((bits>>bit)&1)|((bits>>(bit+7))&2);uint8_t flags=1;if(!color)flags=4;else if(claimed[x])flags=8;else{claimed[x]=1;if((attr&128)&&bg[x])flags=2;}rrhh::dot(x,y,gbShade(m->io[pal],color),0x8000+address,bits,0xff00+pal,m->io[pal],color,7-bit,row,flags);}
 }
 rrcapture::trace.current=0;
}
inline void rrGBStat(gameboy_Machine*m){auto&v=rrhh::video;unsigned ly=m->io[0x44];unsigned mode=!(m->io[0x40]&128)?0:ly>=144?1:m->lcdDot<80?2:m->lcdDot<252?3:0;
 m->io[0x41]=(m->io[0x41]&0xf8)|mode|(ly==m->io[0x45]?4:0);bool irq=(m->io[0x40]&128)&&(((m->io[0x41]&0x40)&&(m->io[0x41]&4))||(mode==0&&(m->io[0x41]&8))||(mode==1&&(m->io[0x41]&16))||(mode==2&&(m->io[0x41]&32)));
 if(irq&&!v.statIRQ)m->io[0x0f]|=2;v.statIRQ=irq;
}
inline void gameboy_Machine_tick(gameboy_Machine*m,int64_t cycles){
 auto&v=rrhh::video;m->Cycles+=cycles;m->divCounter=(m->divCounter+cycles)&65535;m->io[4]=m->divCounter>>8;
 if(m->io[7]&4){static constexpr int periods[]={1024,16,64,256};int period=periods[m->io[7]&3];m->timaCounter+=cycles;while(m->timaCounter>=period){m->timaCounter-=period;if(++m->io[5]==0){m->io[5]=m->io[6];m->io[15]|=4;}}}
 if(!(m->io[0x40]&128)){m->lcdDot=0;m->io[0x44]=0;v.windowLine=0;v.offDots+=cycles;if(v.offDots>=70224){v.offDots-=70224;rrhh::clear(0xffffffff);rrhh::present();}rrGBStat(m);return;}
 v.offDots=0;int old=m->lcdDot;m->lcdDot+=cycles;if(m->io[0x44]<144&&old<252&&m->lcdDot>=252)rrGBLine(m,m->io[0x44]);
 if(m->lcdDot>=456){m->lcdDot-=456;if(++m->io[0x44]>=154)m->io[0x44]=0;if(m->io[0x44]==144){m->io[15]|=1;rrhh::present();}}
 rrGBStat(m);
}
inline void gameboy_Machine_writeIO(gameboy_Machine*m,uint16_t a,uint8_t value){
 if(a==0xff41)m->io[0x41]=(value&0x78)|(m->io[0x41]&7)|0x80;
 else if(a==0xff40){bool was=m->io[0x40]&128;m->io[0x40]=value;if(was!=bool(value&128)){m->lcdDot=0;m->io[0x44]=0;rrhh::video.windowLine=0;rrhh::video.statIRQ=false;}}
 else gameboy_Machine_writeIO_Reference(m,a,value);
 if(a==0xff40||a==0xff41||a==0xff45)rrGBStat(m);
}
