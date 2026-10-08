#define RR_SHADER_TEST
#include "../../platform/n3ds/browser/core/api.cpp"
#include <cassert>
static bool refuse=false;
int main(){
 machine=arenaNew(n3ds_Machine{});machine->CPU=arm_NewCPU(machine);machine->gpu=n3ds_newGPU(machine);machine->SingleThreaded=true;rrgpu::enabled=true;
 auto g=machine->gpu;auto ram=n3ds_Machine_mapRegion(machine,"vertices",0x14000000,Slice<uint8_t>::make(4096));
 for(int i=0;i<12;i++)ram->data[i]=i%4;
 for(int i=0;i<64;i++)ram->data[64+i]=i;
 g->Code[0]=0x13u<<26;g->Code[1]=0x22u<<26;g->Opdesc[0]=15|(0x1bu<<5);g->Regs[0x4f]=1;g->Regs[0x50]=0x0b0a0908;
 rrshader::readyTest=[](uint32_t){return true;};
 rrshader::runTest=[](uint32_t,const rrshader::Values*in,rrshader::Values*out,uint32_t count,n3ds_GPU*g){
  assert(count==4);for(uint32_t i=0;i<count;i++){assert(n3ds_GPU_shaderRun(g,const_cast<rrshader::Values*>(&in[i]),&out[i],g->Regs[0x2ba]&4095));if(refuse&&i==1)return false;}return true;
 };
 auto outs=Slice<n3ds_vsOut>::make(12);auto loaders=Slice<n3ds_loaderBuf>{{64,0,1,4}};auto components=Slice<int64_t>{0};
 auto draw=[&](){return rrvertex::Draw(g,true,ram->base,13,0,loaders,components,ram->base,false,12,false);};
 {
  rr_perf_enable(1);rrperf::Draw profile(g,true);auto fast=draw();assert(rrshader::batch(fast,0,12,0,1,outs));
  for(int i=0;i<12;i++){for(int c=0;c<4;c++)assert(outs[i].color[c]==float((i%4)*4+c));}
  assert(profile.compiled==4&&profile.cacheHits==8&&profile.unique==4&&profile.count==12);
 }
 assert(rrperf::programs.begin()->second.compiled==4);
 auto before=std::vector<n3ds_vsOut>(outs.p,outs.p+outs.n);
 {auto fast=draw();auto generation=fast.generation;refuse=true;assert(!rrshader::batch(fast,0,12,0,1,outs));assert(fast.generation!=generation);assert(!fast.reuse(0,0,outs));assert(!memcmp(before.data(),outs.p,outs.n*sizeof(n3ds_vsOut)));}
 refuse=false;
 // Every canonical input stays live; shader cache keys include entry, code and
 // descriptors and compare full bytes even if the diagnostic hash collides.
 auto initial=rrshader::program(g);g->Float[0][0]=4;g->Bool=1;assert(rrshader::program(g)==initial);
 g->Code[0]=0x13u<<26|32<<12;n3ds_GPU_invalidateShaders(g);auto h=rrperf::shaderHash(g);rrshader::programs[initial-1].hash=h;
 auto changed=rrshader::program(g);assert(changed!=initial);
 {auto fast=draw();assert(rrshader::batch(fast,0,12,0,1,outs));assert(outs[0].color[0]==4);}
 g->Float[0][0]=7;{auto fast=draw();assert(rrshader::batch(fast,0,12,0,1,outs));assert(outs[0].color[0]==7);}
 g->Opdesc[0]^=1;n3ds_GPU_invalidateShaders(g);assert(rrshader::program(g)!=changed);
 g->Regs[0x2ba]=1;assert(rrshader::program(g)!=changed);
 rrshader::invalidate();assert(rrshader::program(g)==rrshader::lastID);
 // Eligibility remains identical to D1, including observation and RAM guards.
 machine->OnRead=[](auto...){};{auto fast=draw();assert(!rrshader::batch(fast,0,12,0,1,outs));}machine->OnRead={};
 {auto fast=draw();assert(!rrshader::batch(fast,0,32769,0,1,outs));}
 std::cout<<"3DS compiled batch: transaction rollback, exact mapping, reuse, live uniforms, cache identity and observation pass\n";
}
