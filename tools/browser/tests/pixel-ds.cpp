#include "../../platform/nds/browser/core/api.cpp"
#include <cassert>
int main(){
 auto spi=dsmachine_newSPI();assert(le_Uint16(sub(spi->firmware,32,34))==0x7fc0);
 auto s=Slice<uint8_t>{1,2,3};auto out=fullSub(s,0,0,0);out=append(out,uint8_t(9));assert(s[0]==1&&out[0]==9);
 dsmachine_card cd;cd.rom=Slice<uint8_t>::make(8192);cd.rom[0x1234]=0x5a;cd.cmd={0xb7,0,0,0x12,0x34,0,0,0};dsmachine_card_start(&cd,7u<<24);assert(cd.buf.n==4&&cd.buf[0]==0x5a);
 // Fixed-point edge interpolation rounds by up to one coordinate unit.
 auto poly=Slice<dsmachine_gxVertex>{{12,-12,0,10,0,0,0,0,0},{12,12,0,10,0,0,0,0,0},{-12,12,0,10,0,0,0,0,0},{-12,-12,0,10,0,0,0,0,0}};auto clipped=dsmachine_clipPolygon(poly);assert(clipped.n>=3);for(auto v:clipped)for(int i=0;i<6;i++)assert(dsmachine_planeDist(v,i)>=-1);assert(poly[0].x==12&&poly[0].y==-12);
 machine=arenaNew(dsmachine_Machine{});machine->ARM9=arenaNew(dsmachine_core{});machine->ARM9->cpu=arenaNew(arm_CPU{});machine->gpu2d=dsmachine_newGPU2D();machine->gpu3d=dsmachine_newGPU3D();machine->gpu2d->a.m=machine;machine->gpu2d->b.m=machine;
 // Verify the renderer's AABBGGRR surface at its 2D boundary, including alpha.
 auto&rast=machine->gpu3d->rast;dsmachine_raster_reset(&rast);
 rast.col[0]={63,0,0,31};rast.col[1]={0,63,0,31};rast.col[2]={0,0,63,31};rast.col[3]={63,63,63,0};
 dsmachine_raster_publish(&rast);machine->gpu2d->a.threeD=rast.frame;
 dsmachine_engine_threeDLine(&machine->gpu2d->a,0);
 assert(machine->gpu2d->a.bg[0][0]==31&&machine->gpu2d->a.bg[0][1]==(31<<5)&&machine->gpu2d->a.bg[0][2]==(31<<10));
 assert(machine->gpu2d->a.a3D[0]==31&&!machine->gpu2d->a.bgOK[0][3]);
 // Per-texel alpha is independent of polygon alpha, including depth writes.
 machine->vram=dsmachine_newVRAM();dsmachine_vram_setCNT(machine->vram,0,0x83);dsmachine_vram_setCNT(machine->vram,4,0x83);
 machine->vram->bank[4][0]=0xff;machine->vram->bank[4][1]=0x7f;
 dsmachine_gxPolygon alphaPoly{};alphaPoly.attr=31u<<16|1u<<24;alphaPoly.texParam=6u<<26;
 auto alphaState=dsmachine_gpu3d_polyState(machine->gpu3d,machine,&alphaPoly,9);
 dsmachine_rvert fragment{};fragment.iw=1;fragment.depth=100;fragment.r=fragment.g=fragment.b=63;
 rast.col[0]={0,0,63,31};rast.depth[0]=1000;rast.transID[0]=0xff;machine->vram->bank[0][0]=15<<3;
 dsmachine_gpu3d_shade(machine->gpu3d,machine,&alphaState,&alphaPoly,fragment,0);
 assert(rast.col[0].r==31&&rast.col[0].g==31&&rast.col[0].b==63&&rast.col[0].a==31&&rast.depth[0]==1000);
 dsmachine_gpu3d_shade(machine->gpu3d,machine,&alphaState,&alphaPoly,fragment,0);assert(rast.col[0].r==31);
 machine->vram->bank[0][0]=31<<3;dsmachine_gpu3d_shade(machine->gpu3d,machine,&alphaState,&alphaPoly,fragment,0);assert(rast.col[0].r==63&&rast.depth[0]==100);
 assert(rr_capture_begin());dsmachine_engine_fillWhite(&machine->gpu2d->a);dsmachine_engine_fillBlack(&machine->gpu2d->b);assert(rr_capture_end());assert(rrcapture::trace.shadow==rrcapture::trace.final);rrreplay::replay.begin();while(!rr_replay_seek(rrreplay::replay.steps.size())){}auto live=frame(machine);auto p=rr_replay_frame();assert(std::equal(live.begin(),live.end(),p));assert(rrcapture::trace.overflow==0);
 // A static completed LCD must not hide drawing into the separate 3D target.
 for(bool swap:{false,true}){
  machine->gpu2d->swap=swap;assert(rr_capture_begin());
  machine->gpu3d->regs[dsmachine_regCLEARCOLOR]=31u<<16;
  dsmachine_gpu3d_clear(machine->gpu3d,machine,0);
  auto clear=rrcapture::trace.current;
  auto draw=rrcapture::trace.event(1,0,"{\"kind\":\"GX polygon rasterization\"}");rrds::render3DEvents.insert(draw);
  rrcapture::trace.record(rrds::plane3D,0xff0000ff,4,1,0,draw);
  rrcapture::trace.event(2,0,"{\"kind\":\"2D composition\"}");
  machine->gpu2d->threeD=Slice<uint32_t>::make(49152);std::fill(machine->gpu2d->threeD.begin(),machine->gpu2d->threeD.end(),0xff000000);machine->gpu2d->threeD[0]=0xff0000ff;
  assert(rr_capture_end());assert(rrcapture::trace.shadow==rrcapture::trace.final);rr_replay_begin();
  auto seekEvent=[&](uint32_t event){auto&r=rrreplay::replay;for(uint32_t i=0;i<r.steps.size();i++)if(r.steps[i].event==event){while(!rr_replay_seek(i+1)){}return;}assert(false);};
  seekEvent(clear);auto pixel=rr_replay_frame()+(swap?0:rrds::planeSize);assert(pixel[0]==0&&pixel[1]==0&&pixel[2]==0&&pixel[3]==255);
  seekEvent(draw);pixel=rr_replay_frame()+(swap?0:rrds::planeSize);assert(pixel[0]==255&&pixel[1]==0&&pixel[2]==0&&pixel[3]==255);
  assert(std::string(rr_replay_info()).find("3D render target")!=std::string::npos);
  while(!rr_replay_seek(rrreplay::replay.steps.size())){}live=frame(machine);assert(std::equal(live.begin(),live.end(),rr_replay_frame()));
  seekEvent(clear);pixel=rr_replay_frame()+(swap?0:rrds::planeSize);assert(pixel[0]==0); // backward seek
 }
}
