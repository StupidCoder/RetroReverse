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
 // General integer fragment tails, fed by exact reference interpolation and
 // sampling. Exercise six stages, delayed combiner-buffer updates, alpha tests,
 // all blend equations/factors, logic operations and partial color masks.
 g->Regs[0x105]=0;fb.width=fb.height=64;g->Regs[0x82]=8|(8<<16);g->Regs[0x85]=(r->base+0x80000)>>3;g->Regs[0x83]=0;
 tri.v0.x=.1f;tri.v0.y=.3f;tri.v1.x=63.7f;tri.v1.y=.2f;tri.v2.x=.4f;tri.v2.y=63.8f;tri.maxX=tri.maxY=64;
 tri.v0.iw=.7f;tri.v1.iw=1.3f;tri.v2.iw=.9f;tri.area=n3ds_edgeFn(tri.v0.x,tri.v0.y,tri.v1.x,tri.v1.y,tri.v2.x,tri.v2.y);
 tri.v0.col={.1f,.7f,.3f,.9f};tri.v1.col={.8f,.2f,.4f,.3f};tri.v2.col={.2f,.9f,.6f,.5f};tri.v0.uv[0]={-.3f,.2f};tri.v1.uv[0]={1.2f,.1f};tri.v2.uv[0]={.1f,1.4f};
 auto rnd=[&](){seed^=seed<<13;seed^=seed>>17;seed^=seed<<5;return seed;};
 std::array<uint8_t,9> sources{0,1,2,3,4,5,13,14,15};std::array<uint8_t,8> combines{0,1,2,3,4,5,8,9};
 for(bool raster:{false,true})for(uint32_t trial=0;trial<96;trial++){
  rrgpu::rasterEnabled=raster;
  fb.colorMask=trial%16;tv.texEnable=1;g->Regs[0x8e]=trial%14;g->Regs[0x83]=((trial%4)<<12)|(((trial/4)%4)<<8);
  tv.bufColor={int32_t(rnd()%256),int32_t(rnd()%256),int32_t(rnd()%256),int32_t(rnd()%256)};tv.alphaTest=trial%2;tv.alphaFunc=(trial/2)%8;tv.alphaRef=rnd()%256;
  for(auto&s:tv.stages){for(int j=0;j<3;j++){s.colr[j]={sources[rnd()%9],uint8_t(rnd()%16)};s.alph[j]={sources[rnd()%9],uint8_t(rnd()%8)};}s.combC=combines[rnd()%8];s.combA=combines[rnd()%8];s.scaleC=rnd()%4;s.scaleA=rnd()%4;s.updC=rnd()%2;s.updA=rnd()%2;s.konst={int32_t(rnd()%256),int32_t(rnd()%256),int32_t(rnd()%256),int32_t(rnd()%256)};}
  g->Regs[0x100]=trial<80?256:0;g->Regs[0x101]=(trial%5)|(((trial/5)%5)<<8)|((trial%15)<<16)|(((trial+3)%15)<<20)|(((trial+7)%15)<<24)|(((trial+11)%15)<<28);g->Regs[0x102]=trial%16;g->Regs[0x103]=rnd();
  n3ds_GPU_invalidateTextures(g,r->base+0x80000,4096);
  n3ds_GPU_fill(g,&fb,&ls,&tv,Slice<n3ds_rasterTri>{tri,tri});
 }
 // Fractional/negative edges, shared edges, constant-color rounding, extreme
 // reciprocal W, subnormal attributes, all three texture units and empty bins.
 rrgpu::rasterEnabled=true;tv={};tv.texEnable=7;fb.colorMask=15;g->Regs[0x100]=0;g->Regs[0x102]=3;
 for(uint32_t trial=0;trial<64;trial++){
  fb.width=trial%4==2?72:64;fb.height=trial%4==2?40:64;tri.maxX=fb.width;tri.maxY=fb.height;
  for(int u=0;u<3;u++){auto[dim,param,addr,fmt]=n3ds_texUnitRegs(u);g->Regs[dim]=(trial%8==0&&u==2)?0:(uint32_t(8+u*8)<<16)|(8+u*8);g->Regs[param]=((trial%8)<<12)|(((trial/8)%8)<<8);g->Regs[addr]=(trial%8==7&&u==0?fb.colorAddr:r->base+0x80000+u*0x10000)>>3;g->Regs[fmt]=(trial+u)%14;}
  // PICA texture-0 type is normal, including when using other register flags.
  for(auto&s:tv.stages){s={};for(int j=0;j<3;j++){s.colr[j].src=s.alph[j].src=trial%4?3+trial%3:0;}}
  tri.v0.x=trial%2?-.5f:.5f;tri.v0.y=.5f;tri.v1.x=63.5f;tri.v1.y=.5f;tri.v2.x=.5f;tri.v2.y=63.5f;
  tri.area=n3ds_edgeFn(tri.v0.x,tri.v0.y,tri.v1.x,tri.v1.y,tri.v2.x,tri.v2.y);
  int i=0;for(auto*v:{&tri.v0,&tri.v1,&tri.v2}){
   v->iw=trial%3==0?(i==0?0x1p-20f:i==1?0x1p20f:1.f):1.f;
   v->col=trial%4==0?std::array<float,4>{1.f,1.f,1.f,1.f}:std::array<float,4>{-1.f,0x1p-149f,.5f,1.1f};
   for(int u=0;u<3;u++)v->uv[u]={float(i-1)*(trial%2?1023.5f:1.f)+float(u)*.125f,float(1-i)+float(u)*0x1p-149f};i++;
  }
  auto second=tri;second.v0.x=63.5f;second.v0.y=63.5f;second.v1=tri.v2;second.v2=tri.v1;second.area=n3ds_edgeFn(second.v0.x,second.v0.y,second.v1.x,second.v1.y,second.v2.x,second.v2.y);
  auto empty=tri;empty.minX=empty.maxX=3;
  n3ds_GPU_invalidateTextures(g,r->base,2*1024*1024);
  n3ds_GPU_fill(g,&fb,&ls,&tv,Slice<n3ds_rasterTri>{tri,empty,second,tri});
 }
 assert(!machine->CPU->Halted);assert(rrgpu::dropped==0);
 uint32_t n=rr_graphics_end();auto snapshot=rrgpu::stream;
 n3ds_Machine_gxMemoryFill(machine,r->base,0,r->base+r->data.n,0x200);
 assert(snapshot==rrgpu::stream); // Inputs are owned, not borrowed guest views.
 assert(n>8);if(argc>1){std::ofstream f(argv[1],std::ios::binary);f.write((char*)rr_graphics_data(),n);}
 std::cout<<"3DS immutable GX stream: fill widths, gaps, overlap/aliases, display formats and source overwrite pass\n";
}
