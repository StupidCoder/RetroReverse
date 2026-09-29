#include "../../platform/gba/browser/core/api.cpp"
#include <cassert>
#include <iostream>
int main(){
 auto*p=rr_input(512);assert(p);std::fill(p,p+512,0);p[0]=0xfe;p[1]=0xff;p[2]=0xff;p[3]=0xea;assert(rr_init(512));auto*m=machine;gbamachine_bus b{m};
 // Bitmap BG2, identity affine mapping. A pixel's final color must retain
 // the real VRAM sample as an intermediate source and replay to the same frame.
 m->io[0]=3|0x400;m->io[0x20]=0x100;m->io[0x26]=0x100;
 for(int i=0;i<240*160;i++)gbamachine_bus_Write16(&b,0x06000000+i*2,uint16_t(i&31));
 assert(rr_capture_begin());for(int y=0;y<160;y++){m->ppu.bg2x=0;m->ppu.bg2y=y*256;gbamachine_ppu_renderLine(&m->ppu,m,y);}assert(rr_capture_end());
 assert(rrcapture::trace.overflow==0);assert(m->screen[1]==gbamachine_rgb15(1));
 auto before=std::string(rr_proof());rr_replay_begin();auto n=rrreplay::replay.steps.size();assert(n>160);while(!rr_replay_seek(n)){}auto*frame=rr_frame();std::vector<uint8_t>final(frame,frame+240*160*4);assert(!std::memcmp(final.data(),rr_replay_frame(),final.size()));assert(std::string(rr_pixel(1,0)).find("\"complete\":true")!=std::string::npos);while(!rr_replay_seek(0)){}assert(std::memcmp(final.data(),rr_replay_frame(),final.size()));assert(before==rr_proof());
 // Blending, forced blank and a frame stopped/restored between scanlines.
 m->ppu.bg2x=0;m->ppu.bg2y=0;m->pal[0]=31;m->pal[1]=0;m->io[0x50]=0x2444;m->io[0x52]=0x0808;gbamachine_ppu_renderLine(&m->ppu,m,2);assert(m->screen[2*240]==gbamachine_rgb15(gbamachine_blend(0,31,8,8)));
 auto size=rr_state_save();assert(size);std::vector<uint8_t>s(rr_state_data(),rr_state_data()+size);assert(rr_run(1000)>=0);auto proof=std::string(rr_proof());std::memcpy(rr_state_input(size),s.data(),size);assert(rr_state_load(size));assert(rr_run(1000)>=0);assert(proof==rr_proof());std::memcpy(rr_state_input(size-1),s.data(),size-1);assert(!rr_state_load(size-1));assert(proof==rr_proof());
 m=machine;m->io[0]=128;assert(rr_capture_begin());gbamachine_ppu_renderLine(&m->ppu,m,0);assert(rr_capture_end());assert(m->screen[0]==0xffffffff);assert(std::string(rr_pixel(0,0)).find("\"complete\":true")!=std::string::npos);
 // BIOS RegisterRamReset writes must participate in historical memory too.
 m->pal[0]=3;m->vram[10]=9;m->oam[20]=7;m->cpu->Thumb=true;m->cpu->R[0]=0x1c;
 assert(rr_capture_begin());assert(m->cpu->SWI(m->cpu,1));assert(rr_capture_end());
 assert(rrcapture::trace.shadow[rrgba::palBase]==0&&rrcapture::trace.shadow[rrgba::vramBase+10]==0&&rrcapture::trace.shadow[rrgba::oamBase+20]==0);
 assert(std::string(rr_source(rrgba::vramBase+10,1,UINT32_MAX,0)).find("\"complete\":true")!=std::string::npos);
 // Tile/character bank, palette and display control all follow replay memory.
 m=machine;b.m=m;m->io[8]=0;m->vram[0x4000]=0;m->pal[2]=0;
 assert(rr_capture_begin());
 gbamachine_Machine_ioWrite16(m,0x04000008,4);m->Steps++;
 gbamachine_bus_Write16(&b,0x06004000,0x21);m->Steps++;
 gbamachine_bus_Write16(&b,0x05000002,31);assert(rr_capture_end());
 rr_replay_begin();assert(rr_tileset_size()==0x1a000);auto*tiles=rr_tileset_data();assert(tiles[8]==0&&tiles[8192+0x4000]==0&&tiles[4098]==0);
 auto stateSize=rr_state_save();std::vector<uint8_t>checkpoint(rr_state_data(),rr_state_data()+stateSize);
 while(!rr_replay_seek(rrreplay::replay.steps.size())){}tiles=rr_tileset_data();assert(tiles[8]==4&&tiles[8192+0x4000]==0x21&&tiles[4098]==31);
 while(!rr_replay_seek(0)){}tiles=rr_tileset_data();assert(tiles[8]==0&&tiles[8192+0x4000]==0&&tiles[4098]==0);
 assert(rr_state_save()==stateSize&&!memcmp(checkpoint.data(),rr_state_data(),stateSize));
 std::cout<<"GBA pixel, replay, blending and state checks pass\n";
}
