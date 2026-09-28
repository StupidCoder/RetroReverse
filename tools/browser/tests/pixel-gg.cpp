#include "../../platform/gamegear/browser/core/api.cpp"
#include <cassert>
#include <iostream>
int main(){
 std::vector<uint8_t>rom(32768);rom[0]=0x3e;rom[1]=0x55;rom[2]=0x32;rom[3]=0x00;rom[4]=0xc0;rom[5]=0x18;rom[6]=0xfe;std::memcpy(rr_input(rom.size()),rom.data(),rom.size());assert(rr_init(rom.size()));assert(rr_run(10)==10);assert(machine->ram[0]==0x55);
 auto*m=machine;auto&v=m->VDP;v.Regs[0]=4;v.Regs[1]=64;v.Regs[2]=14;v.Regs[5]=0x7e;
 v.CRAM[2]=15;v.CRAM[18*2+1]=15;
 for(int y=0;y<8;y++){v.VRAM[y*4]=255;v.VRAM[32+y*4+1]=255;}
 for(int i=0;i<64;i++)v.VRAM[0x3f00+i]=0xd0;v.VRAM[0x3f00]=23;v.VRAM[0x3f80]=48;v.VRAM[0x3f81]=1;
 rrGGLine(m,24);assert(rrhh::video.draw[0]==0xffff0000); // blue sprite, little-endian RGBA
 unsigned nt=0x3800+3*64+6*2;v.VRAM[nt+1]=0x10;rrGGLine(m,24);assert(rrhh::video.draw[0]==0xff0000ff);v.VRAM[nt+1]=0;
 for(int i=0;i<9;i++){v.VRAM[0x3f00+i]=23;v.VRAM[0x3f80+i*2]=i<8?0:48;v.VRAM[0x3f81+i*2]=1;}rrGGLine(m,24);assert(rrhh::video.draw[0]==0xff0000ff&&(v.status&0x40));
 for(int i=0;i<64;i++)v.VRAM[0x3f00+i]=0xd0;
 assert(rr_capture_begin());gamegear_Machine_Out(m,0xbf,0);gamegear_Machine_Out(m,0xbf,0x40);gamegear_Machine_Out(m,0xbe,0x7f);for(int y=0;y<192;y++)rrGGLine(m,y);rrhh::present();assert(rr_capture_end());assert(rrcapture::trace.shadow==rrcapture::trace.final);assert(!rrcapture::trace.overflow);
 auto evidence=std::string(rr_pixel(0,0));assert(evidence.find("sourceAddress\":65536")!=std::string::npos);rr_replay_begin();while(!rr_replay_seek(rrreplay::replay.steps.size())){}assert(std::memcmp(rr_frame(),rr_replay_frame(),rrhh::pixels*4)==0);while(!rr_replay_seek(1)){}assert(std::memcmp(rr_frame(),rr_replay_frame(),rrhh::pixels*4)!=0);
 auto n=rr_state_save();assert(n);auto saved=stateOutput;for(int i=0;i<10;i++)assert(rr_run(13)==13);auto proof=std::string(rr_proof());std::memcpy(rr_state_input(saved.size()),saved.data(),saved.size());assert(rr_state_load(saved.size()));assert(rr_run(130)==130);assert(std::string(rr_proof())==proof);rr_state_input(12);assert(!rr_state_load(12));assert(std::string(rr_proof())==proof);
 gamegear_Machine_Write(machine,0xfffe,0);assert(gamegear_Machine_Read(machine,0x4000)==rom[0]);assert(gamegear_Machine_Read(machine,0)==rom[0]);
 std::cout<<"Game Gear CPU, mapper, VDP ports, sprite priority/limit, source history, replay and states passed\n";
}
