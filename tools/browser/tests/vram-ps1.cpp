#include "../../platform/psx/browser/core/gpu-inspection.h"
#include <cassert>
#include <iostream>
int main(){
 GPU g;auto&t=rrcapture::trace;
 g.vram[400*1024+1]=31;g.vram[256*1024+64]=0x4321;
 t.begin((uint8_t*)g.vram.data(),g.vram.size()*2);rrps1::begin(g);
 g.onCommand=[&](const auto&w){auto s=rrps1::command(g,w);rrps1::record(t.event(1,0x80001000,"{}"),s);};
 g.onPixel=[&](int x,int y,u16 v){t.record((y*1024+x)*2,v,2,1,0,t.current);};
 auto packet=[&](std::initializer_list<u32>w){for(auto x:w)g.gp0(x);return t.current;};
 auto mode=packet({0xe1000261});assert(rrps1::last.depth==0&&rrps1::last.dither==1&&rrps1::last.blend==3);
 auto mask=packet({0xe6000003});assert(rrps1::last.forceMask==1&&rrps1::last.checkMask==1);
 // A polygon's packet overrides page/depth before its first sample. The
 // observer must expose the new value even though the hook precedes execution.
 auto poly4=packet({0x24808080,0x000a000a,(400u<<6)<<16,0x000a000b,0x00110001,0x000b000a,0x00000100});
 assert(rrps1::last.pageX==64&&rrps1::last.pageY==256&&rrps1::last.depth==0&&rrps1::last.clut==400<<6);
 assert(rrps1::last.vertices==3&&rrps1::last.uv[1][0]==1&&g.vram[10*1024+10]==31);
 auto gouraud8=packet({0x3c808080,0x00280028,(401u<<6)<<16,0x00808080,0x00280029,0x00820001,0x00808080,0x00290028,0x00000100,0x00808080,0x00290029,0x00000101});
 assert(rrps1::last.depth==1&&rrps1::last.pageX==128&&rrps1::last.gouraud&&rrps1::last.vertices==4);
 auto direct=packet({0xe100010f});assert(rrps1::last.depth==2&&rrps1::last.pageX==960&&rrps1::last.dither==0);
 auto window=packet({0xe2008421});assert(rrps1::last.maskX==1&&rrps1::last.maskY==1&&rrps1::last.windowX==1&&rrps1::last.windowY==1);
 packet({0xe300140a});packet({0xe4005028});auto offset=packet({0xe53ff7ff});assert(rrps1::last.offsetX==-1&&rrps1::last.offsetY==-2);
 auto rectangle=packet({0x65808080,0x00140014,0x64000706,0x00090008});assert(rrps1::last.raw&&rrps1::last.uv[2][0]==14&&rrps1::last.uv[2][1]==16);
 auto upload=packet({0xa0000000,0x01000040,0x00010001,0x0000abcd});
 t.end((uint8_t*)g.vram.data(),g.vram.size()*2);rrreplay::replay.begin();auto live=g.vram;auto commands=g.commands;
 auto seek=[&](unsigned event){auto&r=rrreplay::replay;for(unsigned i=0;i<r.steps.size();i++)if(r.steps[i].event==event){while(!r.seek(i+1)){}return;}assert(false);};
 seek(poly4);assert(rrps1::selected().depth==0&&rrps1::selected().pageY==256);assert(t.value(rrreplay::replay.memory,(256*1024+64)*2,2)==0x4321);
 seek(gouraud8);assert(rrps1::selected().depth==1&&rrps1::selected().clut==401<<6);
 seek(upload);assert(t.value(rrreplay::replay.memory,(256*1024+64)*2,2)==0xabcd);
 seek(poly4);assert(t.value(rrreplay::replay.memory,(256*1024+64)*2,2)==0x4321);
 seek(mode);assert(rrps1::selected().clut==-1&&rrps1::selected().dither==1);
 seek(mask);assert(rrps1::selected().checkMask==1);
 seek(window);assert(rrps1::selected().windowY==1);
 seek(rectangle);assert(rrps1::selected().depth==2&&rrps1::selected().raw);
 while(!rrreplay::replay.seek(0)){}assert(rrps1::selected().dither==-1&&rrps1::selected().clut==-1);
 assert(g.vram==live&&g.commands==commands);
 std::cout<<"PS1 GPU inspection: packet page/depth/CLUT overrides, no-write state commands, UV layouts, historical VRAM, backwards seeks and immutable inspection pass\n";
}
