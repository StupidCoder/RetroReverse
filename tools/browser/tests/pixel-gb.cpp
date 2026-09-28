#include "../../platform/gameboy/browser/core/api.cpp"
#include <cassert>
#include <iostream>
int main(){
 std::vector<uint8_t>rom(32768);rom[0x100]=0x3e;rom[0x101]=0x55;rom[0x102]=0xea;rom[0x103]=0x00;rom[0x104]=0xc0;rom[0x105]=0x18;rom[0x106]=0xfe;
 std::memcpy(rr_input(rom.size()),rom.data(),rom.size());assert(rr_init(rom.size()));assert(rr_run(10)==10);assert(machine->wram[0]==0x55);
 auto*m=machine;m->io[0x40]=0x93;m->io[0x47]=m->io[0x48]=m->io[0x49]=0xe4;
 for(int y=0;y<8;y++){m->vram[y*2]=255;m->vram[16+y*2+1]=255;m->vram[32+y*2]=m->vram[32+y*2+1]=255;}
 m->oam[0]=16;m->oam[1]=8;m->oam[2]=1;m->oam[3]=128;
 m->oam[4]=16;m->oam[5]=8;m->oam[6]=2;m->oam[7]=0;
 rrGBLine(m,0);assert(rrhh::video.draw[0]==0xffaaaaaa); // BG hides the winning object and masks the second.
 m->oam[3]=0;rrGBLine(m,0);assert(rrhh::video.draw[0]==0xff555555);
 m->oam.fill(0);for(int i=0;i<11;i++){m->oam[i*4]=16;m->oam[i*4+1]=i<10?0:8;m->oam[i*4+2]=1;}rrGBLine(m,0);assert(rrhh::video.draw[0]==0xffaaaaaa);
 m->io[0x40]=0xf1;m->io[0x4a]=0;m->io[0x4b]=7;m->vram[0x1c00]=2;rrGBLine(m,0);assert(rrhh::video.draw[0]==0xff000000&&rrhh::video.windowLine==1);
 m->io[0x40]=0x91;assert(rr_capture_begin());gameboy_Machine_Write(m,0x8000,0x7f);for(int y=0;y<144;y++)rrGBLine(m,y);rrhh::present();assert(rr_capture_end());assert(rrcapture::trace.shadow==rrcapture::trace.final);assert(!rrcapture::trace.overflow);
 auto evidence=std::string(rr_pixel(0,0));assert(evidence.find("sourceAddress\":32768")!=std::string::npos);
 rr_replay_begin();while(!rr_replay_seek(rrreplay::replay.steps.size())){}assert(std::memcmp(rr_frame(),rr_replay_frame(),rrhh::pixels*4)==0);while(!rr_replay_seek(1)){}assert(std::memcmp(rr_frame(),rr_replay_frame(),rrhh::pixels*4)!=0);
 // State continuation and failure are independent of execution slice size.
 auto n=rr_state_save();assert(n);auto saved=stateOutput;for(int i=0;i<10;i++)assert(rr_run(13)==13);auto proof=std::string(rr_proof());std::memcpy(rr_state_input(saved.size()),saved.data(),saved.size());assert(rr_state_load(saved.size()));assert(rr_run(130)==130);assert(std::string(rr_proof())==proof);rr_state_input(12);assert(!rr_state_load(12));assert(std::string(rr_proof())==proof);
 // ROM-only writes cannot change banking; MBC1 mode 1 changes the low window.
 gameboy_Machine_Write(machine,0x2000,0);assert(machine->romBank==1);
 auto cart=Slice<uint8_t>::make(1048576);cart[0x147]=1;cart[32*16384]=0x5a;auto*banked=gameboy_NewMachine(cart);gameboy_Machine_Write(banked,0x4000,1);gameboy_Machine_Write(banked,0x6000,1);assert(gameboy_Machine_Read(banked,0)==0x5a);
 std::cout<<"Game Boy CPU, window, object priority/limit, source history, replay and states passed\n";
}
