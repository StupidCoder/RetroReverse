#include "../../platform/nds/browser/core/api.cpp"
#include <cassert>
int main(){
 auto spi=dsmachine_newSPI();assert(le_Uint16(sub(spi->firmware,32,34))==0x7fc0);
 auto s=Slice<uint8_t>{1,2,3};auto out=fullSub(s,0,0,0);out=append(out,uint8_t(9));assert(s[0]==1&&out[0]==9);
 dsmachine_card cd;cd.rom=Slice<uint8_t>::make(8192);cd.rom[0x1234]=0x5a;cd.cmd={0xb7,0,0,0x12,0x34,0,0,0};dsmachine_card_start(&cd,7u<<24);assert(cd.buf.n==4&&cd.buf[0]==0x5a);
 // Fixed-point edge interpolation rounds by up to one coordinate unit.
 auto poly=Slice<dsmachine_gxVertex>{{12,-12,0,10,0,0,0,0,0},{12,12,0,10,0,0,0,0,0},{-12,12,0,10,0,0,0,0,0},{-12,-12,0,10,0,0,0,0,0}};auto clipped=dsmachine_clipPolygon(poly);assert(clipped.n>=3);for(auto v:clipped)for(int i=0;i<6;i++)assert(dsmachine_planeDist(v,i)>=-1);assert(poly[0].x==12&&poly[0].y==-12);
 machine=arenaNew(dsmachine_Machine{});machine->ARM9=arenaNew(dsmachine_core{});machine->ARM9->cpu=arenaNew(arm_CPU{});machine->gpu2d=dsmachine_newGPU2D();machine->gpu3d=dsmachine_newGPU3D();machine->gpu2d->a.m=machine;machine->gpu2d->b.m=machine;
 assert(rr_capture_begin());dsmachine_engine_fillWhite(&machine->gpu2d->a);dsmachine_engine_fillBlack(&machine->gpu2d->b);assert(rr_capture_end());assert(rrcapture::trace.shadow==rrcapture::trace.final);rrreplay::replay.begin();while(!rr_replay_seek(rrreplay::replay.steps.size())){}auto live=frame(machine);auto p=rr_replay_frame();assert(std::equal(live.begin(),live.end(),p));assert(rrcapture::trace.overflow==0);
}
