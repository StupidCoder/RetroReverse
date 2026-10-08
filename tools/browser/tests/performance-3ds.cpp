#include "../../platform/n3ds/browser/core/api.cpp"
#include <cassert>

int main(){
 machine=arenaNew(n3ds_Machine{});machine->CPU=arm_NewCPU(machine);machine->gpu=n3ds_newGPU(machine);machine->SingleThreaded=true;
 auto g=machine->gpu;auto r=n3ds_Machine_mapRegion(machine,"fixture",0x14000000,Slice<uint8_t>::make(65536));
 for(int i=0;i<6;i++)r->data[i]=i%3;
 g->Regs[0x200]=r->base>>3;g->Regs[0x228]=6;g->Code[0]=34u<<26; // END, no outputs.
 auto counters=[&](){return std::array<int64_t,6>{g->Draws,g->PixelsDrawn,g->DepthKilled,g->RejectedTris,g->ZeroAreaTris,g->CulledTris};};
 const auto before=counters();assert(rr_perf_enable(0));n3ds_GPU_draw(g,true);const auto reference=counters();
 const auto outputs=std::vector<n3ds_vsOut>(g->outs.p,g->outs.p+g->outs.n);
 g->Draws=before[0];g->PixelsDrawn=before[1];g->DepthKilled=before[2];g->RejectedTris=before[3];g->ZeroAreaTris=before[4];g->CulledTris=before[5];
 assert(rr_perf_enable(2));n3ds_GPU_draw(g,true);assert(counters()==reference);assert(!machine->CPU->Halted);
 assert(rrperf::programs.size()==1);const auto v=rrperf::programs.begin()->second;
 assert(v.vertices==6&&v.unique==3&&v.samples==3&&v.draws==1);
 assert(v.fetchMs>=0&&v.shaderMs>=0&&v.clipMs>=0);
 for(int i=0;i<6;i++){assert(outputs[i].pos==g->outs[i].pos);assert(outputs[i].color==g->outs[i].color);assert(outputs[i].uv==g->outs[i].uv);}
 for(int i=0;i<6;i++)assert(r->data[i]==i%3);
 assert(!machine->OnRead&&!machine->OnWrite&&!machine->OnPixel&&!machine->Profile);
 const auto oldHash=rrperf::programs.begin()->first;g->Code[0]=33u<<26;g->Code[1]=34u<<26;n3ds_GPU_invalidateShaders(g);n3ds_GPU_draw(g,true);
 assert(rrperf::programs.size()==2&&rrperf::shaderHash(g)!=oldHash);
 n3ds_fbState fb{};fb.width=fb.height=64;fb.colorAddr=r->base+0x1000;fb.colorMask=15;
 n3ds_lightState light{};n3ds_tevState tev{};n3ds_rasterTri tri{};
 tri.v0.x=tri.v0.y=.5f;tri.v1.x=63.5f;tri.v1.y=.5f;tri.v2.x=.5f;tri.v2.y=63.5f;tri.area=3969;tri.maxX=tri.maxY=64;
 for(auto*p:{&tri.v0,&tri.v1,&tri.v2}){p->iw=1;p->col={1,1,1,1};}
 rrgpu::enabled=true;n3ds_GPU_fill(g,&fb,&light,&tev,Slice<n3ds_rasterTri>{tri});
 assert(rrperf::fallbacks.size()==1);const auto f=rrperf::fallbacks.begin()->second;
 assert(f.draws==1&&f.triangles==1&&f.work==4096&&f.ms>=0&&f.pixels>0);
 assert(std::string(rr_perf_stats()).find("\"vertexTimesEstimated\":true")!=std::string::npos);
 assert(!rr_perf_enable(1025));assert(rr_perf_enable(0));assert(rrperf::programs.empty()&&rrperf::fallbacks.empty());
 n3ds_GPU_draw(g,true);assert(rrperf::programs.empty());
 std::cout<<"3DS optional diagnostics: counts, sampled phases, shader epochs, fallback work and unchanged vertex results pass\n";
}
