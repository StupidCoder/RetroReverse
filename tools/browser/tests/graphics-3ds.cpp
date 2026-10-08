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
  if(trial%4==0){
   // Identity stages must still advance the delayed buffer and honor updates.
   for(int i:{1,2}){auto&s=tv.stages[i];s.colr[0]=s.alph[0]={15,0};s.combC=s.combA=s.scaleC=s.scaleA=0;}
   tv.stages[3].colr[0]=tv.stages[3].alph[0]={13,0};
  }
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
 // Unlit depth: all comparisons, Z/W buffers, write masks, alpha rejection,
 // ordered overdraw, clipped depth, subnormals and a texture alias of depth RAM.
 tv={};tv.texEnable=1;g->Regs[0x82]=8|(8<<16);g->Regs[0x83]=0;g->Regs[0x8e]=0;
 for(auto&s:tv.stages){s.colr[0]=s.alph[0]={3,0};}
 fb.width=fb.height=64;fb.depthTest=true;tri.maxX=tri.maxY=64;
 tri.v0.x=.5f;tri.v0.y=.5f;tri.v1.x=63.5f;tri.v1.y=.5f;tri.v2.x=.5f;tri.v2.y=63.5f;tri.area=63.f*63.f;
 for(uint32_t trial=0;trial<96;trial++){
  fb.depthFunc=trial%8;fb.depthWr=(trial/8)%2;fb.depthZBuffer=(trial/16)%2;fb.colorMask=trial%16;
  fb.depthScale=trial<64?1.f:-.75f;fb.depthOff=trial<64?0.f:.25f;
  if(trial>=88){fb.depthScale=-0.f;fb.depthOff=-0.f;} // Signed zero quantizes to zero.
  tv.alphaTest=trial>=32;tv.alphaFunc=trial%8;tv.alphaRef=127;
  auto depth=rrgpu::range(machine,fb.depthAddr,64*64*4);
  for(uint32_t i=0;i<64*64;i++){uint32_t z=trial%3?0x7fffff:rnd()&0xffffff;depth[i*4]=z;depth[i*4+1]=z>>8;depth[i*4+2]=z>>16;depth[i*4+3]=rnd();}
  int vertex=0;for(auto*v:{&tri.v0,&tri.v1,&tri.v2}){v->iw=trial%3?1.f:(vertex==0?.7f:vertex==1?1.3f:.9f);v->z=trial<32?.5f:trial<64?float(vertex)-.5f:(vertex==0?0x1p-149f:float(vertex));v->col={1.f,.5f,.25f,1.f};v->uv[0]={float(vertex)*.4f,float(2-vertex)*.3f};vertex++;}
  auto second=tri;second.v0.z+=.25f;second.v1.z-=.5f;second.v2.z+=.125f;
  g->Regs[0x85]=(trial%4? r->base+0x80000:fb.depthAddr)>>3;
  n3ds_GPU_invalidateTextures(g,r->base,2*1024*1024);
  // Depth-rejected draws must not create speculative texture-cache entries.
  const n3ds_texKey key{g->Regs[0x85]<<3,0,8,8};
  {auto cold=rrgpu::Operation::draw(g,&fb,&ls,&tv,Slice<n3ds_rasterTri>{tri,second,tri});assert(!cold.target);assert(!std::get<1>(lookup(g->texCache,key)));}
  assert(std::get<1>(n3ds_GPU_texture(g,key.addr,key.fmt,key.w,key.h)));
  n3ds_GPU_fill(g,&fb,&ls,&tv,Slice<n3ds_rasterTri>{tri,second,tri});
 }
 // Directional lighting includes the demo's full normal-map + D0/D1/Fresnel
 // combination. Keep every TEV lighting channel observable in these fixtures.
 fb.depthAddr=r->base+0x30000;fb.depthScale=1;fb.depthOff=0;fb.depthZBuffer=true;
 tv={};tv.texEnable=7;fb.colorMask=15;
 for(int u=0;u<3;u++){auto[dim,param,addr,fmt]=n3ds_texUnitRegs(u);g->Regs[dim]=8|(8<<16);g->Regs[param]=0;g->Regs[addr]=(r->base+0x80000+u*0x10000)>>3;g->Regs[fmt]=0;}
 for(uint32_t trial=0;trial<96;trial++){
  ls={};ls.enabled=true;ls.count=1;ls.env=trial%8;ls.ambient={.1f,.05f,.15f};
  ls.bumpMode=(trial/8)%4;ls.bumpSel=trial%3;ls.noBumpRenorm=(trial/4)%2;
  ls.primaryAlpha=trial%2;ls.secondAlpha=true;ls.clampHighlight=(trial/2)%2;
  ls.noD0=trial%5==0;ls.noD1=trial%7==0;ls.noFR=trial%11==0;ls.noRR=trial%13==0;ls.noRG=trial%17==0;ls.noRB=trial%19==0;
  ls.lutAbs=rnd();ls.lutIn=0;ls.lutScale=0;
  for(int field:{0,4,12,16,20,24}){ls.lutIn|=((trial+field/4)%8)<<field;ls.lutScale|=((trial+field/4)%8)<<field;}
  auto&light=ls.lights[0];light.directional=true;light.twoSided=trial%3==0;
  light.pos=trial%7?std::array<float,3>{.4f,-.25f,.9f}:std::array<float,3>{0,0,0};
  light.spotDir={-.8f,.3f,.1f};light.diffuse={.8f,.7f,.6f};light.ambient={.2f,.15f,.1f};light.specular0={.25f,.5f,.75f};light.specular1={.4f,.3f,.2f};
  if(trial==6){light.pos={-1.f,-0.f,-0.f};ls.lutAbs=0;ls.lutIn=0x03333033;light.diffuse={};light.ambient={};light.specular0={.01f,.02f,.03f};light.specular1={.04f,.05f,.06f};ls.ambient={};}
  for(int t=0;t<7;t++)for(int j=0;j<256;j++){g->LUT[t][j]=float(rnd()%4096)/4095;g->LUTDiff[t][j]=float(int(rnd()%4096)-2048)/4095;}
  int k=0;for(auto*v:{&tri.v0,&tri.v1,&tri.v2}){
   v->z=.2f+float(k)*.1f;v->iw=k==0?.7f:k==1?1.3f:.9f;
   v->quat=trial%6?std::array<float,4>{float(k)*.25f,-.3f,.5f,.7f}:std::array<float,4>{0,0,0,0};
   v->view=trial%8?std::array<float,3>{float(k)-.5f,.3f,-1.f}:std::array<float,3>{0,0,0};
   if(trial>=88){v->quat={0x1p-70f,-0x1p-74f,0x1p-72f,0x1p-71f};v->view={0x1p-70f,0x1p-71f,-0x1p-72f};}
   if(trial>=80&&trial<88){v->view={-0x1p20f,0x1p18f,0x1p19f};v->quat={0x1p20f,-0x1p19f,0x1p18f,-0x1p17f};}
   k++;
  }
  for(auto&stage:tv.stages){stage={};stage.colr[0]=stage.alph[0]={15,0};}
  tv.stages[0].colr[0]=tv.stages[0].alph[0]={uint8_t(1+trial%2),0};
  if(trial%3==0){tv.stages[1].combC=2;tv.stages[1].colr[0]={15,0};tv.stages[1].colr[1]={uint8_t(2-trial%2),0};}
  fb.depthFunc=trial%4==0?4:1;fb.depthWr=trial%2;tv.alphaTest=trial%3==1;tv.alphaFunc=trial%8;tv.alphaRef=127;
  auto depth=rrgpu::range(machine,fb.depthAddr,64*64*4);for(int j=0;j<64*64;j++){uint32_t word=(rnd()&0xff000000)|0x7fffff;std::memcpy(depth+j*4,&word,4);}
  for(int u=0;u<3;u++){auto[dim,param,addr,fmt]=n3ds_texUnitRegs(u);assert(std::get<1>(n3ds_GPU_texture(g,g->Regs[addr]<<3,0,8,8)));}
  auto operation=rrgpu::Operation::draw(g,&fb,&ls,&tv,Slice<n3ds_rasterTri>{tri,tri});assert(operation.target&&operation.kind==8);operation.record=false;
  // Only the fill below belongs in the immutable stream.
  n3ds_GPU_fill(g,&fb,&ls,&tv,Slice<n3ds_rasterTri>{tri,tri});
  // Same specialization key, different light colors, direction and LUT values.
  // Cached kernels must continue to read all of these from the current packet.
  light.diffuse={.1f,.2f,.3f};light.specular0={.4f,.3f,.2f};light.pos={-.7f,.2f,.4f};
  for(int t=0;t<7;t++)for(int j=0;j<256;j++){g->LUT[t][j]=float((j+t*17)%256)/255;g->LUTDiff[t][j]=-.25f;}
  n3ds_GPU_fill(g,&fb,&ls,&tv,Slice<n3ds_rasterTri>{tri,tri});
 }
 for(int reason=0;reason<6;reason++){
  auto invalid=ls;if(reason==0)invalid.count=2;if(reason==1)invalid.shadow=true;if(reason==2)invalid.env=8;if(reason==3)invalid.lights[0].directional=false;if(reason==4)invalid.lights[0].distAtten=true;if(reason==5)invalid.lights[0].geo1=true;
  assert(!rrgpu::Operation::draw(g,&fb,&invalid,&tv,Slice<n3ds_rasterTri>{tri,tri}).target);
 }
 // General stencil: every compare, fail/depth-fail/pass operation and write
 // mask, with alpha before stencil, optional depth, lighting and ordered overdraw.
 for(bool lit:{false,true})for(uint32_t trial=0;trial<192;trial++){
  fb.depthAddr=r->base+0x30000;fb.width=fb.height=64;fb.depthTest=trial%3!=0;fb.depthWr=(trial/3)%2;
  fb.depthFunc=(trial/8)%8;fb.depthZBuffer=trial%2;fb.depthScale=1;fb.depthOff=0;fb.colorMask=(trial/4)%16;
  const uint32_t reference=(trial%4==0?0:trial%4==1?255:trial%4==2?0x5a:0xa5),mask=trial%5?255:0x5a,writeMask=trial%7?255:0xa5;
  g->Regs[0x105]=1|((trial%8)<<4)|(writeMask<<8)|(reference<<16)|(mask<<24);
  g->Regs[0x106]=((trial/8)%8)|(((trial/16+3)%8)<<4)|(((trial/24+5)%8)<<8);g->Regs[0x115]=trial%5!=0;g->Regs[0x116]=3;
  tv={};tv.texEnable=trial%2;tv.alphaTest=true;tv.alphaFunc=(trial/3)%8;tv.alphaRef=127;
  for(auto&stage:tv.stages){stage={};stage.colr[0]=stage.alph[0]={15,0};}
  tv.stages[0].colr[0]={uint8_t(lit?1:0),0};tv.stages[0].alph[0]={0,0};
  ls={};ls.enabled=lit;ls.count=1;ls.env=0;ls.noD0=ls.noD1=ls.noFR=ls.noRR=ls.noRG=ls.noRB=true;
  auto&light=ls.lights[0];light.directional=true;light.pos={.25f,.5f,1};light.diffuse={.75f,.5f,.25f};light.ambient={.1f,.2f,.3f};
  tri.maxX=tri.maxY=64;tri.v0.x=.5f;tri.v0.y=.5f;tri.v1.x=63.5f;tri.v1.y=.5f;tri.v2.x=.5f;tri.v2.y=63.5f;tri.area=63.f*63.f;
  int k=0;for(auto*v:{&tri.v0,&tri.v1,&tri.v2}){v->iw=k==0?.7f:k==1?1.3f:.9f;v->z=.2f+float(k)*.15f;v->col={.2f,.6f,.9f,float(k)*.5f};v->quat={.1f,.2f,.3f,.9f};v->view={float(k),.5f,-1};k++;}
  auto second=tri;second.v0.z+=.4f;second.v1.z+=.4f;second.v2.z+=.4f;
  auto depth=rrgpu::range(machine,fb.depthAddr,64*64*4);
  for(uint32_t j=0;j<64*64;j++){const uint32_t stencil=j%5==0?0:j%5==1?255:j%5==2?reference:rnd()%256;uint32_t word=(stencil<<24)|(j%4==0?0:j%4==1?0xffffff:0x7fffff);std::memcpy(depth+j*4,&word,4);}
  g->Regs[0x85]=(trial%2?fb.depthAddr:r->base+0x80000)>>3;g->Regs[0x82]=8|(8<<16);g->Regs[0x83]=0;g->Regs[0x8e]=0;
  n3ds_GPU_invalidateTextures(g,r->base,2*1024*1024);
  if(tv.texEnable){const n3ds_texKey key{g->Regs[0x85]<<3,0,8,8};auto cold=rrgpu::Operation::draw(g,&fb,&ls,&tv,Slice<n3ds_rasterTri>{tri,second,tri});assert(!cold.target);assert(!std::get<1>(lookup(g->texCache,key)));assert(std::get<1>(n3ds_GPU_texture(g,key.addr,key.fmt,key.w,key.h)));}
  {auto op=rrgpu::Operation::draw(g,&fb,&ls,&tv,Slice<n3ds_rasterTri>{tri,second,tri});assert(op.target&&op.kind==(lit?10:9));op.record=false;}
  n3ds_GPU_fill(g,&fb,&ls,&tv,Slice<n3ds_rasterTri>{tri,second,tri});
  if(lit&&trial<8){
   // Identical lighting key across kinds 8/10; a compiled stencil kernel must
   // never be reused for an ordinary depth draw (or the reverse).
   auto cfg=g->Regs[0x105];g->Regs[0x105]=0;auto testing=fb.depthTest;fb.depthTest=true;
   n3ds_GPU_fill(g,&fb,&ls,&tv,Slice<n3ds_rasterTri>{tri,second,tri});g->Regs[0x105]=cfg;fb.depthTest=testing;
  }
 }
 g->Regs[0x105]=1;g->Regs[0x115]=1;g->Regs[0x116]=3;fb.depthTest=false;ls={};
 // Even without depth testing, stencil/color aliases require ordered Reference.
 for(uint32_t depth:{fb.colorAddr,fb.colorAddr+4,0x15000000+(fb.colorAddr-r->base)}){fb.depthAddr=depth;assert(!rrgpu::Operation::draw(g,&fb,&ls,&tv,Slice<n3ds_rasterTri>{tri,tri}).target);}
 fb.depthAddr=r->base+0x30000;g->Regs[0x116]=2;assert(!rrgpu::Operation::draw(g,&fb,&ls,&tv,Slice<n3ds_rasterTri>{tri,tri}).target);
 g->Regs[0x105]=0;g->Regs[0x116]=3;fb.depthTest=true;
 ls={};
 // Overlapping depth/color, including different virtual addresses aliasing the
 // same bytes, cannot execute independently per pixel. Stay in Reference.
 for(uint32_t depth:{fb.colorAddr,fb.colorAddr+4,0x15000000+(fb.colorAddr-r->base)}){
  fb.depthAddr=depth;auto op=rrgpu::Operation::draw(g,&fb,&ls,&tv,Slice<n3ds_rasterTri>{tri,tri});assert(!op.target);
 }
 assert(!machine->CPU->Halted);assert(rrgpu::dropped==0);
 uint32_t n=rr_graphics_end();auto snapshot=rrgpu::stream;
 n3ds_Machine_gxMemoryFill(machine,r->base,0,r->base+r->data.n,0x200);
 assert(snapshot==rrgpu::stream); // Inputs are owned, not borrowed guest views.
 assert(n>8);if(argc>1){std::ofstream f(argv[1],std::ios::binary);f.write((char*)rr_graphics_data(),n);}
 std::cout<<"3DS immutable GX stream: fill widths, gaps, overlap/aliases, display formats and source overwrite pass\n";
}
