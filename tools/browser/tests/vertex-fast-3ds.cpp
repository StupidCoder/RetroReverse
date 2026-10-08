#include "../../platform/n3ds/browser/core/api.cpp"
#include <cassert>

int main(){
 machine=arenaNew(n3ds_Machine{});machine->CPU=arm_NewCPU(machine);machine->gpu=n3ds_newGPU(machine);machine->SingleThreaded=true;
 auto g=machine->gpu;auto ram=n3ds_Machine_mapRegion(machine,"vertices",0x14000000,Slice<uint8_t>::make(65536));
 uint32_t seed=0x3d500d01;for(auto&b:ram->data){seed^=seed<<13;seed^=seed>>17;seed^=seed<<5;b=seed;}
 g->Regs[0x200]=ram->base>>3;g->Regs[0x228]=30;g->Regs[0x203]=0x1000;g->Regs[0x204]=0x210c0;g->Regs[0x205]=(5u<<28)|(80u<<16);
 g->Regs[0x4f]=1;g->Regs[0x50]=0x0b0a0908; // Route o0 to color; leave position clipped.
 g->Opdesc[0]=15|(0x1bu<<5)|(0x1bu<<14)|(0x1bu<<23);
 g->Code[0]=(0x13u<<26);g->Code[1]=0x22u<<26;
 for(int a=0;a<12;a++)g->fixedVal[a]={float(a),-float(a),0x1p-149f,-0.f};
 auto counters=[&](){return std::array<int64_t,6>{g->Draws,g->PixelsDrawn,g->DepthKilled,g->RejectedTris,g->ZeroAreaTris,g->CulledTris};};
 auto reset=[&](){g->Draws=g->PixelsDrawn=g->DepthKilled=g->RejectedTris=g->ZeroAreaTris=g->CulledTris=0;machine->CPU->Halted=false;};
 auto check=[&](bool indexed,bool expectReuse){
  reset();rr_graphics_enable(0);rr_perf_enable(0);n3ds_GPU_draw(g,indexed);assert(!machine->CPU->Halted);
  auto expected=counters();auto outputs=std::vector<n3ds_vsOut>(g->outs.p,g->outs.p+g->Regs[0x228]);
  for(auto mode:{0,1,2}){
   rrvertex::fetchEnabled=mode!=0;rrvertex::reuseEnabled=mode!=1;
   reset();rr_graphics_enable(1);rr_perf_enable(1);n3ds_GPU_draw(g,indexed);assert(!machine->CPU->Halted);assert(counters()==expected);
   assert(!std::memcmp(outputs.data(),g->outs.p,outputs.size()*sizeof(n3ds_vsOut)));
   const auto&v=rrperf::programs.begin()->second;assert(v.vertices==g->Regs[0x228]);
   assert(v.cacheHits==uint64_t(expectReuse&&mode!=1?g->Regs[0x228]-7:0));
  }
  rrvertex::fetchEnabled=rrvertex::reuseEnabled=true;
 };
 for(bool wide:{false,true}){
  g->Regs[0x227]=wide?0x80000000u:0;
  for(unsigned i=0;i<30;i++){if(wide){ram->data[i*2]=i%7;ram->data[i*2+1]=0;}else ram->data[i]=i%7;}
  // Every scalar format/component count, padding, repeated attributes, input
  // mappings and fixed inputs are compared against the actual generated path.
  for(uint32_t f=0;f<16;f++)for(uint32_t attr=0;attr<3;attr++){
   g->Regs[0x201]=f|(((f+5)&15)<<4)|(((f+11)&15)<<8);g->Regs[0x202]=(f%2?1u<<attr:0)<<16;g->Regs[0x2bb]=attr;
   check(true,true);
  }
  // A later draw sees writes, uniform edits, output mapping and shader edits.
  g->Float[0]={.25f,-.5f,.75f,1.f};g->Code[0]=0x20u<<12;n3ds_GPU_invalidateShaders(g);check(true,true);
  for(int i=0;i<80;i++)ram->data[0x1000+i]^=255;
  g->Float[0][0]=.5f;g->Regs[0x50]=0x08090a0b;check(true,true);
  g->Code[0]=0x13u<<26;n3ds_GPU_invalidateShaders(g);
  g->Regs[0x202]=1u<<19;g->Regs[0x2bb]=3;
  g->fixedVal[3]={-0.f,0x1p-149f,std::numeric_limits<float>::quiet_NaN(),std::numeric_limits<float>::infinity()};check(true,true);
  g->Regs[0x2bb]=0;
  g->Regs[0x205]=5u<<28; // Maximum indices with zero-stride ordinary attributes.
  for(unsigned i=0;i<30;i++){auto v=(wide?65535u:255u)-i%7;if(wide){ram->data[i*2]=v;ram->data[i*2+1]=v>>8;}else ram->data[i]=v;}
  check(true,true);rrvertex::nextGeneration=UINT32_MAX;check(true,true);
  for(unsigned i=0;i<30;i++){if(wide){ram->data[i*2]=i%7;ram->data[i*2+1]=0;}else ram->data[i]=i%7;}
  g->Regs[0x205]=(5u<<28)|(80u<<16);
 }
 g->Regs[0x22a]=1;check(false,false);
 // Physical RAM aliases remain valid; changed page mappings never read through
 // an old region pointer, and an indexed span across regions stays in Reference.
 n3ds_Machine_mapRegion(machine,"alias",0x15000000,ram->data);g->Regs[0x200]=0x15000000>>3;check(true,true);
 auto replacement=n3ds_Machine_mapRegion(machine,"replacement",0x15002000,Slice<uint8_t>::make(4096));
 assert(!rrvertex::ordinary(machine,0x15001ff0,32));assert(rrvertex::ordinary(machine,0x15002000,32)==replacement->data.p);
 g->Regs[0x203]=0x1ffc;g->Regs[0x205]=(5u<<28);check(true,false);
 g->Regs[0x200]=ram->base>>3;g->Regs[0x203]=0x1000;g->Regs[0x205]=(5u<<28)|(80u<<16);
 unsigned reads=0;machine->OnRead=[&](auto...){reads++;};machine->RWatchLo=0;machine->RWatchHi=UINT32_MAX;
 check(true,false);assert(reads>0);machine->OnRead={};
 machine->OnWrite=[](auto...){};check(true,false);machine->OnWrite={};
 machine->OnPixel=[](auto...){};check(true,false);machine->OnPixel={};
 machine->HidTrace=true;check(true,false);machine->HidTrace=false;
 machine->Profile=true;check(true,false);machine->Profile=false;
 machine->SingleThreaded=false;check(true,false);machine->SingleThreaded=true;
 {rrvertex::Draw observed(g,true,ram->base,0,0,{},{},ram->base,true,30,true);assert(!observed.eligible);}
 // Wraparound and nonindexed partial mappings are never direct RAM spans.
 assert(!rrvertex::ordinary(machine,UINT32_MAX-3,8));assert(!rrvertex::ordinary(machine,0x17000000,4));
 // Shader errors are reached at the same vertex before any output is cached.
 g->Code[0]=0x1cu<<26;n3ds_GPU_invalidateShaders(g);
 reset();rr_graphics_enable(0);n3ds_GPU_draw(g,true);assert(machine->CPU->Halted);auto error=machine->CPU->HaltReason;
 reset();rr_graphics_enable(1);n3ds_GPU_draw(g,true);assert(machine->CPU->Halted&&machine->CPU->HaltReason==error);
 std::cout<<"3DS vertex reuse/direct fetch: exact outputs, mappings, formats, draw invalidation, observation and errors pass\n";
}
