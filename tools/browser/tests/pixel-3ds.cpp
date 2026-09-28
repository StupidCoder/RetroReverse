#include "../../platform/n3ds/browser/core/api.cpp"
#include <cassert>
int main(){
 assert(n3ds_f24bits(0x3f0000)==1);assert(n3ds_toF24(std::numeric_limits<float>::quiet_NaN())==0);
 assert(go_utf16_Decode(Slice<uint16_t>{0xd83d,0xde00})[0]==0x1f600);
 machine=arenaNew(n3ds_Machine{});machine->CPU=arm_NewCPU(machine);machine->gpu=n3ds_newGPU(machine);auto r=n3ds_Machine_mapRegion(machine,"vram",0x1f000000,Slice<uint8_t>::make(240*400*4));machine->screenFB[0]={0,0x1f000000,0x1f000000,240*4,0,0,true};
 assert(rr_capture_begin());{rr3ds::Command command(machine,"GX memory fill",r->base,0,0,0);n3ds_Machine_WriteWord(machine,r->base+239*4,0xff332211);}assert(rr_capture_end());assert(rrcapture::trace.shadow==rrcapture::trace.final);assert(rr3ds::pixelAddress(0,0)==239*4);
 rr_replay_begin();while(!rr_replay_seek(rrreplay::replay.steps.size())){}auto pixels=rr3ds::display(rrreplay::replay.memory);assert(pixels[0]==255&&pixels[1]==51&&pixels[2]==34&&pixels[3]==255);assert(rrcapture::trace.overflow==0);
}
