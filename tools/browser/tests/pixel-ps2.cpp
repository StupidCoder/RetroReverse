#include "../../platform/ps2/browser/core/api.cpp"
#include <cassert>
int main(){
 RRSpinSet set;for(int i=0;i<20;i++)set.add(i%6);assert(set.size()==6);for(int i=0;i<1000;i++)set.add(i);assert(set.size()==7);
 machine=ps2_NewMachine();auto m=machine;rrBind(m);auto g=ps2_Machine_ensureGS(m);
 // Two overlapping GS sprites with a depth test: the farther green sprite
 // loses, and the nearer blue sprite wins. A true replay retains both writes.
 ps2_GS_write(g,ps2_gsFRAME1,1<<16);ps2_GS_write(g,ps2_gsSCISSOR1,63ull<<16|31ull<<48);ps2_GS_write(g,ps2_gsZBUF1,8);ps2_GS_write(g,ps2_gsTEST1,(1<<16)|(3<<17));ps2_GS_write(g,ps2_gsPRMODECONT,1);ps2_GS_write(g,ps2_gsPRIM,6);
 m->io[ps2_gsPMODE]=2;m->io[ps2_gsDISPFB2]=1<<9;m->io[ps2_gsDISPLAY2+4]=63|(31<<12);
 auto sprite=[&](uint32_t color,uint32_t z){ps2_GS_write(g,ps2_gsRGBAQ,color);ps2_GS_write(g,ps2_gsXYZ2,uint64_t(z)<<32);ps2_GS_write(g,ps2_gsXYZ2,uint64_t(z)<<32|uint64_t(16*8)<<16|16*8);};
 assert(rr_capture_begin());sprite(0x800000ff,1000);sprite(0x8000ff00,500);sprite(0x80ff0000,1500);assert(rr_capture_end());assert(rrcapture::trace.shadow==rrcapture::trace.final);assert(!rrcapture::trace.overflow);assert(std::string(rr_pixel(1,1)).find("complete\":true")!=std::string::npos);rr_frame();std::vector<uint8_t>final(pixels.begin(),pixels.end());rr_replay_begin();while(!rr_replay_seek(rrreplay::replay.steps.size())){}assert(std::equal(final.begin(),final.end(),rr_replay_frame()));while(!rr_replay_seek(1)){}auto p=rr_replay_frame()+(width+1)*4;assert(p[0]==255&&p[2]==0);
 // Temporary texture samplers must not accumulate in the machine-lifetime arena.
 auto arenaBefore=arena.size();for(int i=0;i<100000;i++){auto sampler=ps2_GS_sampler(g,16);assert(sampler);}assert(arena.size()==arenaBefore);
 // Tiny MIPS loop crossing a VBlank in unequal slices preserves timer phase.
 m->CPU->PC=0x80010000;m->CPU->nextPC=0x80010004;ps2_Machine_Write32(m,0x10000,0x08004000);ps2_Machine_Write32(m,0x10004,0);m->OnVBlank={};m->rrVblAcc=999950;m->noIdleSkip=true;
 auto n=rr_state_save();assert(n);auto saved=stateOutput;for(int i=0;i<10;i++)assert(rr_run(13)>=0);auto proof=std::string(rr_proof());auto acc=m->rrVblAcc;std::memcpy(rr_state_input(saved.size()),saved.data(),saved.size());assert(rr_state_load(saved.size()));machine->OnVBlank={};assert(rr_run(130)>=0);assert(std::string(rr_proof())==proof&&machine->rrVblAcc==acc);rr_state_input(12);assert(!rr_state_load(12));assert(std::string(rr_proof())==proof);
}
