#include "../../platform/n64/browser/core/host.h"
#include <cassert>
int main() {
 n64_Machine m{};r4300_CPU cpu{};m.CPU=&cpu;cpu.bus=&m;
 m.RDRAM=Slice<uint8_t>::make(4*1024*1024);auto&r=m.rdp;
 r.Color={0,2,4,0x1000};
 // Pass primitive RGB/alpha through both combiner cycles.
 r.Combine=(8ULL<<52)|(16ULL<<47)|(7ULL<<44)|(7ULL<<41)|(8ULL<<28)|(3ULL<<15)|(7ULL<<12)|(3ULL<<9)|
 (8ULL<<37)|(16ULL<<32)|(8ULL<<24)|(3ULL<<6)|(7ULL<<21)|(7ULL<<18)|(7ULL<<3)|3;
 n64_combineInputs in{};in.Prim={200,40,80,0};in.Shade={255,255,255,255};
 for(uint64_t cycle:{0,1})for(uint32_t size:{2,3}){
  r.Color.Size=size;
  auto mux=1ULL<<(cycle?20:22);
  for(auto b:{0ULL,1ULL})for(auto alpha:{0u,64u,128u,255u}){
   r.OtherModes=(cycle<<52)|mux|(b<<(cycle?16:18))|n64_omAntialias|n64_omImageRead|n64_omAlphaCvgSel;
   in.Prim.A=alpha;
   n64_Machine_writePixel(&m,1,1,0,200,240,255);n64_Machine_drawPixel(&m,1,1,&in,0,false);
   auto c=n64_Machine_readPixel(&m,1,1);assert(c.R==200&&c.G==40&&c.B==80);
  }
  for(uint64_t mode:{uint64_t(n64_omForceBlend),uint64_t(n64_omAntialias|n64_omAlphaCvgSel|n64_omCvgTimesAlpha)}){
   r.OtherModes=(cycle<<52)|mux|n64_omImageRead|mode;in.Prim.A=128;
   n64_Machine_writePixel(&m,1,1,0,200,0,255);n64_Machine_drawPixel(&m,1,1,&in,0,false);
   auto c=n64_Machine_readPixel(&m,1,1);assert(c.R>=96&&c.R<=104&&c.G>=112&&c.G<=120);
  }
  in.Prim.A=0;n64_Machine_writePixel(&m,1,1,0,200,0,255);n64_Machine_drawPixel(&m,1,1,&in,0,false);
  auto c=n64_Machine_readPixel(&m,1,1);assert(c.R==0&&c.G==200);
 }
 r.OtherModes=n64_omCvgTimesAlpha|n64_omAlphaCvgSel;
 auto [a,c]=n64_rdp_pixelCoverage(&r,255,8);assert(a==255&&c==8);
 puts("PASS: N64 opaque coverage, transparent blending and cutouts in 1/2-cycle, 16/32-bit buffers");
}
