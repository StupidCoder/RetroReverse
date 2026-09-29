#include "../../platform/gc/browser/core/api.cpp"
#include <cassert>
int main(){
 machine=arenaNew(gc_Machine{});auto m=machine;m->RAM=Slice<uint8_t>::make(24*1024*1024);m->ARAM=Slice<uint8_t>::make(16*1024*1024);m->CPU=gekko_NewCPU(m);rrBind(m);gc_gpu_init(&m->gpu);auto*g=&m->gpu;
 // Two overlapping actual rasterized triangles, then an RGB→YUY2 display copy
 // and a clear. Pixel history must refer to color before that final clear.
 g->BP[0xf3]=(7<<16)|(7<<19);g->BP[0x41]=8;g->BP[0x49]=0;g->BP[0x4a]=639|(479<<10);g->BP[0x4b]=0x100000>>5;m->vi.TFBL=0x100000;
 assert(rr_capture_begin());gc_tevState tev{};tev.seed[0]={255,0,0,255};gc_rasterTri tri{};tri.v0.x=0;tri.v0.y=0;tri.v1.x=40;tri.v1.y=0;tri.v2.x=0;tri.v2.y=40;tri.v0.invW=tri.v1.invW=tri.v2.invW=1;tri.area=gc_edge(0,0,40,0,0,40);tri.minX=tri.minY=0;tri.maxX=tri.maxY=40;
 auto tris=Slice<gc_rasterTri>::make(1);tris[0]=tri;rrconsole::command(1,0x100,"{\"kind\":\"red triangle\"}");gc_gpu_fill(g,m,&tev,tris);assert(g->EFB[1+640]==0xff000000u||g->EFB[1+640]==0xff0000ffu);
 tev.seed[0]={0,255,0,255};rrconsole::command(2,0x104,"{\"kind\":\"green triangle\"}");gc_gpu_fill(g,m,&tev,tris);
 rrconsole::command(3,0x108,"{\"kind\":\"display copy\"}");gc_gpu_copyDisplay(g,m,(1<<14)|(1<<11));assert(rr_capture_end());assert(!rrcapture::trace.overflow);assert(rrgc::hasCopy);auto evidence=std::string(rr_pixel(1,1));assert(evidence.find("green triangle")!=std::string::npos);assert(evidence.find("complete\":true")!=std::string::npos);
 auto final=rrFrame(m);rr_replay_begin();while(!rr_replay_seek(rrreplay::replay.steps.size())){}assert(std::equal(final.begin(),final.end(),rr_replay_frame()));while(!rr_replay_seek(1)){}auto p=rr_replay_frame()+(641*4);assert(p[0]==255&&p[1]==0);
 // Native serialization retains nil optional buffers and resumes the exact CPU.
 m->noIdle=true;m->CPU->PC=0x100;gc_Machine_setRAM32(m,0x100,0x48000000);auto n=rr_state_save();assert(n);auto saved=stateOutput;for(int i=0;i<10;i++)assert(rr_run(10)>=0);auto proof=std::string(rr_proof());std::memcpy(rr_state_input(saved.size()),saved.data(),saved.size());assert(rr_state_load(saved.size()));assert(rr_run(100)>=0);assert(std::string(rr_proof())==proof);rr_state_input(12);assert(!rr_state_load(12));assert(std::string(rr_proof())==proof);
}
