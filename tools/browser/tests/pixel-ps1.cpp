#include "../../platform/psx/browser/core/gpu.h"
#include <cassert>
int main(){GPU g;g.vram[0]=1;g.vram[1025]=31;auto&t=rrcapture::trace;t.begin((uint8_t*)g.vram.data(),g.vram.size()*2);g.onCommand=[&](auto&){t.event(1,0x80001000,"{\"kind\":\"synthetic GPU command\"}");};g.onPixel=[&](int x,int y,u16 v){t.record((y*1024+x)*2,v,2,1,0,t.current);};
 for(u32 w:{0x65000000u,0x000a000au,64u<<16,0x00010001u})g.gp0(w);
 assert(g.vram[10*1024+10]==31);auto sample=t.writes.back();assert(sample.source==0&&sample.palette==2050&&sample.sourceValue==1&&sample.paletteValue==31);
 for(u32 w:{0x0200ff00u,0x000a0000u,0x00010010u})g.gp0(w);assert(g.vram[10*1024+10]==0x3e0);
 for(u32 w:{0x80000000u,0x000a000au,0x00140014u,0x00010001u})g.gp0(w);assert(g.vram[20*1024+20]==0x3e0);auto copy=t.writes.back();
 for(u32 w:{0xa0000000u,0u,0x00010001u,0u})g.gp0(w);
 auto before=g.vram[10*1024+10];for(u32 w:{0x65000000u,0x000a000au,64u<<16,0x00010001u})g.gp0(w);assert(g.vram[10*1024+10]==before);assert(t.writes.back().flags==4);
 t.end((uint8_t*)g.vram.data(),g.vram.size()*2);assert(t.pixel(copy.source,2,copy.sourceBefore,copy.sourceValue).find("\"complete\":true")!=std::string::npos);
 assert(t.pixel(0,2,sample.sourceBefore,sample.sourceValue).find("\"complete\":true")!=std::string::npos); // source was overwritten later
}
