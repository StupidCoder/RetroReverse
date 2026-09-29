#include "../../platform/gameboy/browser/core/api.cpp"
#include <cassert>
#include <iostream>
#include <string>
static void contains(const char*text,const char*part){assert(std::string(text).find(part)!=std::string::npos);}
int main(){
 std::vector<uint8_t>rom(32768);rom[0x100]=0x18;rom[0x101]=0xfe;
 memcpy(rr_input(rom.size()),rom.data(),rom.size());assert(rr_init(rom.size()));auto*m=machine;
 m->io[0x40]=0x93;m->io[0x47]=m->io[0x48]=m->io[0x49]=0xe4;m->io[0x42]=m->io[0x43]=0;
 m->io[0x4a]=72;m->io[0x4b]=7;
 for(int y=0;y<8;y++){m->vram[y*2]=255;m->vram[16+y*2+1]=255;m->vram[32+y*2]=m->vram[32+y*2+1]=255;}
 for(int i=0;i<1024;i++){m->vram[0x1800+i]=i%2;m->vram[0x1c00+i]=i/32==1?0:2;}
 m->oam[0]=32;m->oam[1]=8;m->oam[2]=2;m->oam[3]=128;
 m->oam[4]=32;m->oam[5]=8;m->oam[6]=1;
 auto write=[&](uint16_t address,uint8_t value){rrhh::video.pc=0x4567;rrhh::video.steps++;gameboy_Machine_Write(m,address,value);};
 assert(rr_capture_begin());
 for(int y=0;y<144;y++){
  if(y==48)write(0xff43,8);
  if(y==72)write(0xff40,0xf3);
  if(y==80)write(0xff40,0xd3);
  if(y==90)write(0xff40,0xf3);
  if(y==112)write(0x9c40,1);
  if(y==120)write(0xfe01,40);
  rrhh::video.steps++;rrGBLine(m,y);
 }
 rrhh::present();assert(rr_capture_end());auto proof=std::string(rr_proof());
 auto size=rr_state_save();auto saved=stateOutput;
 contains(rr_raster_info(),"\"line\":48,\"changes\":1");
 contains(rr_raster_seek(47),"\"name\":\"SCX\",\"address\":65347,\"value\":0");
 assert(rrgb::pictures[0][0]==0xffaaaaaa);assert(rrgb::pictures[2][48*160]==0);
 contains(rr_raster_seek(48),"\"pc\":17767"); // actual register writer, not rendering-time PC
 assert(rrgb::pictures[0][0]==0xff555555); // frozen source has shifted everywhere
 assert(rrgb::pictures[2][0]==0xffaaaaaa); // earlier output keeps the original scroll
 assert(rrgb::pictures[2][48*160]==0xff555555);
 contains(rr_raster_pixel(2,0,49),"has not been drawn");
 contains(rr_raster_pixel(0,0,0),"\"mapValue\":1");
 rr_raster_seek(16);assert(rrgb::pictures[1][16*160]==0xff000000);
 assert(rrgb::pictures[2][16*160]==0xffaaaaaa); // BG hides the winning sprite
 contains(rr_raster_pixel(1,0,16),"\"flags\":2");contains(rr_raster_pixel(1,0,16),"\"flags\":8");
 rr_raster_seek(90);assert(rrgb::lines[90].windowLine==8);
 assert(rrgb::pictures[0][90*160]==0xffaaaaaa); // window uses internal row 8, not y-WY=18
 contains(rr_raster_pixel(2,0,90),"\"complete\":true");
 contains(rr_raster_seek(112),"\"mapWrites\":1");contains(rr_raster_seek(120),"\"objectWrites\":1");
 // Frozen snapshots survive changes to live memory after the capture.
 m->vram.fill(0);m->oam.fill(0);m->io[0x43]=123;auto changedProof=std::string(rr_proof());
 rr_raster_seek(47);assert(rrgb::pictures[0][0]==0xffaaaaaa);
 rr_raster_seek(143);assert(!memcmp(rr_raster_frame(2),rr_frame(),160*144*4));
 assert(std::string(rr_proof())==changedProof);
 // Restore, then prove inspection preserves the complete machine state too.
 memcpy(rr_state_input(size),saved.data(),size);assert(rr_state_load(size));assert(std::string(rr_proof())==proof);
 for(int y:{0,72,90,120,143,48,0,143}){rr_raster_seek(y);contains(rr_raster_pixel(2,24,y),"\"complete\":true");}
 assert(rr_state_save()==size&&stateOutput==saved);
 assert(rr_capture_begin());rrhh::clear(0xffffffff);rrhh::present();assert(rr_capture_end());
 contains(rr_raster_info(),"\"lines\":[]");contains(rr_raster_info(),"\"complete\":false");contains(rr_raster_seek(0),"\"error\"");contains(rr_raster_pixel(0,0,0),"\"error\"");
 std::cout<<"Game Boy raster: historical scroll, window counter, sprite priority, source addresses, partial output, immutable seeks and blank LCD pass\n";
}
