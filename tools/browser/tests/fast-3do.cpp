#include "../../platform/threedo/browser/core/host.h"
#include <cassert>
#include <random>
int main(){
 std::mt19937 random(12345);
 auto bytes=Slice<uint8_t>::make(73);for(auto&v:bytes)v=random();
 for(int start=0;start<bytes.n*8+32;start++)for(int n=0;n<=64;n++){
  threedo_bitReader reader{bytes,start};uint32_t expected=0;
  for(int i=0;i<n;i++){auto at=start+i;expected=(expected<<1)|((at/8<bytes.n)?((bytes[at/8]>>(7-at%8))&1):0);}
  assert(threedo_bitReader_read(&reader,n)==expected);assert(reader.pos==start+n);
 }
 auto m=threedo_NewMachine();uint32_t watched=0;
 m->OnWrite=[&](uint32_t,uint32_t,uint32_t){watched++;};m->WatchLo=0x200000;m->WatchHi=0x300000;
 threedo_Machine_flashClearRange(m,0x200003,9,0x1234);assert(watched==8);
 auto before=std::vector<uint8_t>(m->vram.begin(),m->vram.end());
 m->OnWrite={};std::fill(m->vram.begin(),m->vram.end(),0);
 threedo_Machine_flashClearRange(m,0x200003,9,0x1234);assert(std::equal(before.begin(),before.end(),m->vram.begin()));
 // Identity/no-blend stores retain byte watchpoint delivery when enabled.
 threedo_gfxBitmap bm{0x200000,2,2};m->OnWrite=[&](auto...){watched++;};watched=0;
 threedo_Machine_blendPixel(m,bm,0,0,0x7c00,73,0x1f00,0);assert(watched==2);
 m->OnWrite={};threedo_Machine_blendPixel(m,bm,1,0,0x7c00,73,0x1f00,0);
 assert(m->vram[0]==m->vram[4]&&m->vram[1]==m->vram[5]);
 std::cout<<"3DO token boundaries, truncated reads, clear spans and watched-write fallbacks passed\n";
}
