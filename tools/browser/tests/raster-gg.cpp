#include "../../platform/gamegear/browser/core/api.cpp"
#include <cassert>
#include <iostream>
static void contains(const char*text,const char*part){assert(std::string(text).find(part)!=std::string::npos);}
static void reg(unsigned n,uint8_t value){gamegear_Machine_Out(machine,0xbf,value);gamegear_Machine_Out(machine,0xbf,0x80|n);}
static void vram(unsigned a,uint8_t value){gamegear_Machine_Out(machine,0xbf,a);gamegear_Machine_Out(machine,0xbf,0x40|(a>>8));gamegear_Machine_Out(machine,0xbe,value);}
static void frame(){auto n=rrhh::video.frames;while(rrhh::video.frames==n)assert(rr_run(10000)>0);}
int main(){
 std::vector<uint8_t>rom(32768);rom[0x100]=0x18;rom[0x101]=0xfe;memcpy(rr_input(rom.size()),rom.data(),rom.size());assert(rr_init(rom.size()));auto*m=machine;auto&v=m->VDP;
 v.Regs[0]=4;v.Regs[1]=64;v.Regs[2]=14;v.Regs[5]=0x7e;
 v.CRAM[2]=15;v.CRAM[5]=15;v.CRAM[6]=0xf0;v.CRAM[38]=0xf0;
 for(int y=0;y<8;y++){v.VRAM[y*4]=255;v.VRAM[32+y*4+1]=255;v.VRAM[64+y*4]=v.VRAM[64+y*4+1]=255;}
 for(int i=0;i<896;i++)v.VRAM[0x3800+i*2]=i%2;
 std::fill_n(v.VRAM.begin()+0x3f00,64,0xd0);v.VRAM[0x3f00]=39;v.VRAM[0x3f80]=48;v.VRAM[0x3f81]=2;v.VRAM[0x3f01]=39;v.VRAM[0x3f82]=48;v.VRAM[0x3f83]=1;
 v.VRAM[0x3800+5*64+6*2+1]=0x10; // High-priority BG hides the winning sprite.
 assert(rr_capture_begin());
 for(int line=0;line<192;line++){
  rrhh::video.steps++;rrhh::video.pc=0x4567;
  if(line==72)reg(8,8);
  if(line==96){gamegear_Machine_Out(m,0xbf,2);gamegear_Machine_Out(m,0xbf,0xc0);gamegear_Machine_Out(m,0xbe,0xf0);gamegear_Machine_Out(m,0xbe,0);}
  if(line==136)vram(0x3800+17*64+5*2,0);
  if(line==144)vram(0x3f80,80);
  rrGGLine(m,line);
 }
 rrhh::present();assert(rr_capture_end());assert(rrcapture::trace.shadow==rrcapture::trace.final);contains(rr_raster_info(),"\"complete\":true");
 contains(rr_raster_seek(47),"\"scrollX\":0");assert(rrgg::pictures[0][0]==0xff0000ff);assert(rrgg::pictures[2][48*160]==0);
 contains(rr_raster_seek(48),"\"pc\":17767");assert(rrgg::pictures[0][0]==0xffff0000);assert(rrgg::pictures[2][0]==0xff0000ff);assert(rrgg::pictures[2][48*160]==0xffff0000);
 contains(rr_raster_pixel(2,0,49),"has not been drawn");contains(rr_raster_pixel(0,0,0),"\"role\":\"Tilemap entry\"");
 rr_raster_seek(16);assert(rrgg::pictures[1][16*160]==0xff00ff00);assert(rrgg::pictures[2][16*160]==0xff0000ff);contains(rr_raster_pixel(1,0,16),"\"flags\":2");contains(rr_raster_pixel(1,0,16),"\"flags\":8");contains(rr_raster_pixel(1,0,16),"\"role\":\"Sprite X / tile\"");
 contains(rr_raster_seek(72),"\"colorWrites\":1");contains(rr_raster_seek(112),"\"mapWrites\":1");contains(rr_raster_seek(120),"\"objectWrites\":1");
 auto size=rr_state_save();assert(size);auto saved=stateOutput;
 for(int y:{143,0,72,16,48,90,143}){rr_raster_seek(y);contains(rr_raster_pixel(2,24,y),"\"complete\":true");}
 assert(!memcmp(rr_raster_frame(2),rr_frame(),160*144*4));assert(rr_state_save()==size&&stateOutput==saved);
 v.VRAM.fill(0);v.CRAM.fill(0);v.Regs.fill(0);rr_raster_seek(47);assert(rrgg::pictures[0][0]==0xff0000ff);
 memcpy(rr_state_input(size),saved.data(),size);assert(rr_state_load(size));contains(rr_raster_seek(0),"\"error\"");
 assert(rr_capture_begin());assert(rr_capture_end());contains(rr_raster_info(),"\"complete\":false");contains(rr_raster_info(),"\"lines\":[]");

 // Run real Z80 code: every sixteenth line IRQ acknowledges BF and writes R8.
 m=machine;auto*c=m->CPU;*c={};c->bus=m;c->SP=0xd000;c->PC=0x100;m->ram.fill(0);
 rrhh::init(0xff000000);rrgg::timing={};rrgg::timing.lineCounter=15;
 m->VDP.Regs[0]=0x14;m->VDP.Regs[1]=64;m->VDP.Regs[8]=0;m->VDP.Regs[9]=0;m->VDP.Regs[10]=15;m->VDP.status=0;
 uint8_t main[]={0xfb,0x76,0x18,0xfd};memcpy(m->rom.p+0x100,main,sizeof main);
 uint8_t irq[]={0xf5,0xdb,0xbf,0x3a,0,0xc0,0x3c,0x32,0,0xc0,0xd3,0xbf,0x3e,0x88,0xd3,0xbf,0xf1,0xfb,0xc9};memcpy(m->rom.p+0x38,irq,sizeof irq);
 frame();assert(rr_capture_begin());frame();assert(rr_capture_end());contains(rr_raster_info(),"\"complete\":true");
 bool split=false,writer=false;for(int y=1;y<144;y++)if(rrgg::lines[y].scrollX!=rrgg::lines[y-1].scrollX){assert((y+24)%16==0);split=true;}
 for(const auto&w:rrcapture::trace.writes)if(w.address==0x14048){assert(w.pc==0x46);writer=true;}
 assert(split&&writer);assert(rrgg::timing.cycles>=2*rrgg::frameCycles-70*228);
 assert(rr_state_save());saved=stateOutput;for(int y=0;y<144;y++){rr_raster_seek(y);contains(rr_raster_pixel(2,48,y),"\"complete\":true");}assert(rr_state_save()&&stateOutput==saved);
 assert(!memcmp(rr_raster_frame(2),rr_frame(),160*144*4));assert(rrcapture::trace.shadow==rrcapture::trace.final);
 std::cout<<"Game Gear raster: real Z80 line-IRQ scroll splits, historical layers/palettes/tilemap/sprites, CPU writers, partial output and immutable seeks pass\n";
}
