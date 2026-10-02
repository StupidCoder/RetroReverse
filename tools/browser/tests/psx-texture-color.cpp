// Regression: raw polygons ignore vertex colours; modulated polygons retain tint.
#include "../../platform/psx/browser/core/gpu.h"
#include <cassert>
int main() {
 for (u32 depth : {0u, 1u, 2u})
  for (u32 op : {0x24u, 0x25u, 0x2cu, 0x2du, 0x34u, 0x35u, 0x3cu, 0x3du})
   for (u16 texel : {0u, 0x8000u, 0xc210u}) {
    GPU g;
    g.vram[256*1024+512] = texel;
    if (depth < 2) {
     g.vram[256*1024+512] = 1;
     g.vram[300*1024+1] = texel;
    }
    u32 coords[] = {10|10<<16, 18|10<<16, 10|18<<16, 18|18<<16};
    int nv = op & 8 ? 4 : 3;
    g.vram[11*1024+11] = g.vram[17*1024+17] = 0x1234;
    g.gp0(op<<24 | 0x2040ff);
    for (int i=0; i<nv; ++i) {
     if (i && (op&16)) g.gp0(0x2040ff);
     g.gp0(coords[i]);
     g.gp0(i==0 ? (300*64)<<16 : i==1 ? (8|16|depth<<7)<<16 : 0);
    }
    u16 want = !texel ? 0x1234 : op&1 ? texel : texel==0xc210 ? 0x911f : 0x8000;
    assert(g.vram[11*1024+11] == want);
    if(nv==4) assert(g.vram[17*1024+17] == want);
   }
 puts("PASS: PS1 raw/modulated triangles and quads, flat/Gouraud, all texture depths, transparency");
}
