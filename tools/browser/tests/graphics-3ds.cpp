#include "../../platform/n3ds/browser/core/api.cpp"
#include <cassert>
int main(int argc,char**argv){
 machine=arenaNew(n3ds_Machine{});machine->CPU=arm_NewCPU(machine);machine->gpu=n3ds_newGPU(machine);
 auto r=n3ds_Machine_mapRegion(machine,"test",0x14000000,Slice<uint8_t>::make(2*1024*1024));
 uint32_t seed=0x3d500002;for(auto&b:r->data){seed^=seed<<13;seed^=seed>>17;seed^=seed<<5;b=seed;}
 rr_graphics_begin();
 for(uint32_t width=0;width<3;width++)for(uint32_t offset=0;offset<4;offset++)
  n3ds_Machine_gxMemoryFill(machine,r->base+offset,0xa153e70b,r->base+offset+259,width<<8);
 for(uint32_t gap:{0u,3u,19u})n3ds_Machine_gxTextureCopy(machine,r->base,r->base+8193,384,48|(gap<<16),24|((gap+1)<<16));
 // Source and destination overlap, including an alias at a different address.
 n3ds_Machine_gxTextureCopy(machine,r->base,r->base+1,128,0,0);
 n3ds_Machine_mapRegion(machine,"alias",0x15000000,r->data);
 n3ds_Machine_gxTextureCopy(machine,r->base,0x15000001,128,0,0);
 for(uint32_t fmt=0;fmt<5;fmt++)for(uint32_t flip=0;flip<2;flip++)for(uint32_t dw:{24u,40u})
  n3ds_Machine_gxDisplayTransfer(machine,r->base+16384,r->base+65537,32|(24<<16),dw|(32<<16),(fmt<<12)|flip);
 auto g=machine->gpu;g->Regs[0x116]=3;g->Regs[0x115]=1;
 n3ds_fbState fb{};fb.width=fb.height=32;fb.colorAddr=r->base+0x20000;fb.depthAddr=r->base+0x30000;
 n3ds_lightState ls{};auto tv=n3ds_GPU_tevstate(g);
 n3ds_rasterTri tri{};tri.v0.x=.5;tri.v0.y=.5;tri.v1.x=24.5;tri.v1.y=.5;tri.v2.x=.5;tri.v2.y=24.5;
 tri.v0.iw=tri.v1.iw=tri.v2.iw=1;tri.area=576;tri.maxX=tri.maxY=25;
 auto tris=Slice<n3ds_rasterTri>{tri,tri};
 for(uint32_t op=0;op<8;op++)for(uint32_t mask:{255u,90u}){
  g->Regs[0x105]=0xffa50001|(mask<<8);g->Regs[0x106]=op;
  n3ds_GPU_fill(g,&fb,&ls,&tv,tris);
 }
 assert(!machine->CPU->Halted);assert(rrgpu::dropped==0);
 uint32_t n=rr_graphics_end();auto snapshot=rrgpu::stream;
 n3ds_Machine_gxMemoryFill(machine,r->base,0,r->base+r->data.n,0x200);
 assert(snapshot==rrgpu::stream); // Inputs are owned, not borrowed guest views.
 assert(n>8);if(argc>1){std::ofstream f(argv[1],std::ios::binary);f.write((char*)rr_graphics_data(),n);}
 std::cout<<"3DS immutable GX stream: fill widths, gaps, overlap/aliases, display formats and source overwrite pass\n";
}
