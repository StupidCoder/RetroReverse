#include "../../platform/n64/browser/core/host.h"
#include <cassert>
int main(){n64_Machine m{};r4300_CPU cpu{};m.CPU=&cpu;cpu.bus=&m;m.RDRAM=Slice<uint8_t>::make(4*1024*1024);auto&r=m.rdp;r.Color={0,2,4,0x1000};r.Mask=0x2000;n64_combineInputs in{};n64_PixelEvent last{};m.OnPixel=[&](uint32_t,uint32_t,n64_PixelEvent e){last=e;};
 r.OtherModes=n64_omZCompare;n64_Machine_drawPixel(&m,0,0,&in,0,true);assert(last.ZReject&&!last.Drawn);
 r.OtherModes=n64_omAlphaCompare;r.BlendColor=1;n64_Machine_drawPixel(&m,0,0,&in,0,false);assert(last.AlphaReject&&!last.Drawn);
 r.OtherModes=0;r.BlendColor=0;n64_Machine_drawPixel(&m,0,0,&in,0,false);assert(last.Drawn&&!last.ZReject&&!last.AlphaReject);
 r.TMem[0]=1;r.TMem[0x808]=0xf8;r.TMem[0x809]=1;n64_tile tile{};tile.Format=n64_fmtCI;tile.Size=n64_size8;tile.Line=1;tile.SH=4;tile.TH=4;auto [c,ok]=n64_Machine_texelAt(&m,&tile,0,0);assert(ok&&c.R==248&&c.G==0&&c.B==0);
}
