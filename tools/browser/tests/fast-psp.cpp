#include "../../platform/psp/browser/core/api.cpp"
#include <cassert>
#include <random>
#include <iostream>
static std::vector<std::tuple<uint32_t,uint32_t,unsigned>>writes;
static void record(psp_Machine*,uint32_t a,uint32_t v,unsigned n){writes.emplace_back(a,v,n);}
int main(){
 auto*m=psp_NewMachine();std::mt19937 random(0x505350);std::vector<uint8_t>seed(65536),expected(65536);for(auto&b:seed)b=random();
 rrObserveWrite=record;
 for(unsigned i=0;i<6000;i++){
  psp_geState s{};s.fbLow=0x04000000;s.fbStride=32;s.fbFmt=i%4;s.zLow=i%7?0x2000:8;s.zStride=32;
  s.stencilOn=random()%2;s.stFunc=random()%8;s.stRef=random()%256;s.stMask=random()%256;s.stSFail=random()%6;s.stZFail=random()%6;s.stZPass=random()%6;
  s.zTestOn=random()%2;s.zNoWrite=random()%2;s.zFunc=random()%8;s.clearOn=i%13==0;s.clearDepth=random()%2;
  s.blendOn=random()%2;s.blendSrc=random()%11;s.blendDst=random()%11;s.blendEq=random()%6;s.blendFixA=random();s.blendFixB=random();
  if(i%3==0){s.blendOn=true;s.blendSrc=2;s.blendDst=3;s.blendEq=0;}
  s.alphaTestOn=random()%2;s.alphaFunc=random()%8;s.alphaTestMask=random()%256;s.alphaRef=random()%256;
  s.maskRGB=i%3==0?random()&0xffffff:i%17==0?0xffffff:0;s.maskA=i%5==0?random()%256:i%17==0?255:0;
  s.scX0=2;s.scY0=2;s.scX1=20;s.scY1=20;int x=random()%24,y=random()%24;float z=int(random()%66000)-100;
  uint8_t r=random(),g=random(),b=random(),a=random();std::vector<psp_PixelEvent>events;
  m->OnPixel=[&](int64_t,int64_t,psp_PixelEvent e){events.push_back(e);};
  std::memcpy(m->vram.p,seed.data(),seed.size());writes.clear();psp_Machine_putPixel_Reference(m,&s,x,y,z,r,g,b,a);auto referenceWrites=writes;auto referenceEvents=events;std::memcpy(expected.data(),m->vram.p,expected.size());
  std::memcpy(m->vram.p,seed.data(),seed.size());writes.clear();events.clear();{rrFragmentScope scope(m,&s);assert(scope.context.emit);psp_Machine_putPixel(m,&s,x,y,z,r,g,b,a);}
  assert(std::equal(expected.begin(),expected.end(),m->vram.p));assert(writes==referenceWrites);assert(events.size()==referenceEvents.size());
  for(size_t j=0;j<events.size();j++){auto e=events[j],r=referenceEvents[j];assert(e.Drawn==r.Drawn&&e.ScissorReject==r.ScissorReject&&e.AlphaReject==r.AlphaReject&&e.StencilReject==r.StencilReject&&e.ZReject==r.ZReject&&e.MaskReject==r.MaskReject&&e.R==r.R&&e.G==r.G&&e.B==r.B&&e.A==r.A);}
 }
 // Bilinear SIMD lanes must retain the exact scalar channel rounding.
 for(int j=0;j<65536;j++)m->ram[j]=random();
 for(unsigned i=0;i<20000;i++){
  psp_geState s{};s.texAddr=0x08000000;s.texStride=s.texW=s.texH=32;s.texFmt=i%6;s.texSwizzle=i%2;s.texLinear=true;s.texFunc=i%5;s.texUseA=i%2;s.texDouble=i%7==0;s.texMaxLvl=i%4;s.texLodBias=float(int(i%9)-4)/16;s.texWrapU=i%2;s.texWrapV=i%3;s.clutFmt=0xff00;s.texEnvCol=random();
  for(auto&b:s.clut)b=random();float u=float(int(random()%2048)-1024)/99,v=float(int(random()%2048)-1024)/133,rho=float(random()%500)/31;uint8_t r=random(),g=random(),b=random(),a=random();
  rrTextureScope scope(m,&s);assert(psp_modTex(m,&s,u,v,rho,r,g,b,a)==psp_modTex_Reference(m,&s,u,v,rho,r,g,b,a));
 }
 // Raster setup must preserve floating-point operation order, texture filtering,
 // winding, fog and the ordered writes used by capture replay.
 for(unsigned i=0;i<160;i++){
  psp_geState s{};s.fbLow=0x04000000;s.fbStride=512;s.fbFmt=i%4;s.scX1=479;s.scY1=271;s.texEnable=true;s.texAddr=0x08000000;s.texStride=s.texW=s.texH=32;s.texFmt=i%6;s.texSwizzle=i%2;s.texLinear=i%3;s.texFunc=i%5;s.texUseA=i%2;s.texDouble=i%7==0;s.texMaxLvl=i%2;s.texWrapU=i%2;s.texWrapV=i%3;s.clutFmt=0xff00;
  for(auto&b:s.clut)b=random();for(int j=0;j<65536;j++)m->ram[j]=random();
  auto vertex=[&](){psp_vert v{};v.x=float(random()%240)/8-2;v.y=float(random()%240)/8-2;v.z=random()%65536;v.u=float(random()%256)/64-1;v.v=float(random()%256)/64-1;v.invW=float(1+random()%32)/8;v.clip=i%2;v.r=random();v.g=random();v.b=random();v.a=random();v.fog=float(random()%256)/255;return v;};auto a=vertex(),b=vertex(),c=vertex();s.fogOn=i%2;s.fogColor=random();m->OnPixel={};
  std::memcpy(m->vram.p,seed.data(),seed.size());writes.clear();{rrTextureScope ts(m,&s);rrFragmentScope fs(m,&s);psp_Machine_rasterTri_Reference(m,&s,a,b,c);}auto want=writes;std::memcpy(expected.data(),m->vram.p,expected.size());
  std::memcpy(m->vram.p,seed.data(),seed.size());writes.clear();{rrTextureScope ts(m,&s);rrFragmentScope fs(m,&s);psp_Machine_rasterTri(m,&s,a,b,c);}assert(writes==want);assert(std::equal(expected.begin(),expected.end(),m->vram.p));
 }
 for(unsigned format=0;format<4;format++){m->fbAddr=0x04000000;m->fbWidth=512;m->fbFormat=format;assert(rrFrame(m)==rrFrameReference(m));}
 rrObserveWrite=nullptr;std::cout<<"PSP: 6000 fragments, 20000 filters, 160 triangles and four scanout formats match reference\n";
}
