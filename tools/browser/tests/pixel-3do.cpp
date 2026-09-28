#include "../../platform/threedo/browser/core/host.h"
#include <cassert>
int main(){auto m=threedo_NewMachine();threedo_gfxBitmap bm{0x200000,2,2};threedo_Machine_blendPixel(m,bm,0,0,0x7c00,73,0x1f00,0);assert(m->vram[0]==0x7c&&m->vram[1]==0);
 // Equal source/destination blend: red destination plus green source, divided by two.
 threedo_Machine_blendPixel(m,bm,0,0,0x03e0,73,0x1f81,0);uint16_t out=(uint16_t(m->vram[0])<<8)|m->vram[1];assert(out==((15<<10)|(15<<5)));
 threedo_Machine_blendPixel(m,bm,1,1,0x001f,73,0x1f00,0);assert(m->vram[6]==0&&m->vram[7]==31); // interleaved LR-form layout

 // Render an interleaved offscreen buffer through the actual cel renderer.
 auto&t=rrcapture::trace;t.begin(m->vram.p,m->vram.n,0x200000);t.event(0,0,"{\"kind\":\"offscreen copy\"}");
 uint32_t ccb=0x1000;auto w=[&](uint32_t off,uint32_t v){threedo_Machine_write32(m,ccb+off,v);};
 w(16,0);w(20,0);w(24,1<<20);w(28,0);w(32,0);w(36,1<<16);w(40,0);w(44,0);w(48,0x1f001f00);w(52,6);w(56,0x801);
 threedo_gfxBitmap dest{0x200100,2,2};
 m->OnPixel=[&](uint32_t x,uint32_t y,threedo_PixelEvent e){auto a=dest.buf+(y/2)*dest.w*4+x*4+(y&1)*2;t.record(a,rrcapture::little(m->vram.p+a-0x200000,2),2,0,0,t.current,e.Drawn?1:4);};
 assert(threedo_Machine_drawOneCel(m,dest,ccb,threedo_ccbCCBPre,0x200000,0));assert(!t.writes.empty());
 auto copied=t.writes.front();assert(copied.source==0x200000);assert(copied.sourceValue==rrcapture::little(m->vram.p,2));
 m->vram[0]=0;m->vram[1]=0;t.record(0x200000,0,2,0,0);t.end(m->vram.p,m->vram.n);
 assert(t.pixel(copied.source,2,copied.sourceBefore,copied.sourceValue).find("\"complete\":true")!=std::string::npos);
}
