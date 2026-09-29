#include "../../platform/ps2/browser/core/api.cpp"
#include <cassert>
#include <random>

int main(){
  std::mt19937 random(0x592);
  ps2_GS fast{},reference{};
  for(unsigned i=0;i<2000;i++){
    // Long repeated runs, alternating contexts, register changes, map rehash
    // and replacement all have to retain the reference's exact counters.
    unsigned context=i&1;uint64_t p=0x10|0x20|0x40|(context<<9);
    if(i%7==0)p^=0x10;
    if(i%11==0){
      for(unsigned reg:{6u,0x47u,0x4cu,0x4eu})
        fast.reg[reg+context]=reference.reg[reg+context]=(uint64_t(random())<<32)|random();
    }
    for(int j=0;j<5;j++){
      ps2_GS_noteFeatures(&fast,p);ps2_GS_noteFeatures_reference(&reference,p);
    }
    if(i==1000){fast.drawCensus={};reference.drawCensus={};}
    assert(*fast.drawCensus.p==*reference.drawCensus.p);
  }
  for(uint32_t bw:{0u,1u,2u,3u,16u,63u})
    for(uint32_t y=0;y<256;y++)for(uint32_t x=0;x<256;x++){
      assert(ps2_addrPSMT8(23,bw,x,y)==ps2_addrPSMT8_reference(23,bw,x,y));
      assert(ps2_addrPSMT4(23,bw,x,y)==ps2_addrPSMT4_reference(23,bw,x,y));
      assert(ps2_addrPSMCT32(23,bw,x,y)==ps2_addrPSMCT32_reference(23,bw,x,y));
      assert(ps2_addrPSMZ32(23,bw,x,y)==ps2_addrPSMZ32_reference(23,bw,x,y));
    }
  for(int i=0;i<10000;i++){
    uint32_t bp=random(),bw=random(),x=random(),y=random();
    assert(ps2_addrPSMT8(bp,bw,x,y)==ps2_addrPSMT8_reference(bp,bw,x,y));
    assert(ps2_addrPSMT4(bp,bw,x,y)==ps2_addrPSMT4_reference(bp,bw,x,y));
    assert(ps2_addrPSMCT32(bp,bw,x,y)==ps2_addrPSMCT32_reference(bp,bw,x,y));
    assert(ps2_addrPSMZ32(bp,bw,x,y)==ps2_addrPSMZ32_reference(bp,bw,x,y));
  }
  ps2_Machine m{};fast.m=&m;fast.vram=Slice<uint8_t>::make(4*1024*1024);fast.t8Dumped=16;
  for(auto&b:fast.vram)b=random();
  for(auto&c:fast.clut)c=random();
  for(uint32_t psm:{0u,1u,2u,0xau,0x13u,0x14u,0x1bu,0x24u,0x2cu})
    for(int i=0;i<5000;i++){
      ps2_gsSampler s{};s.gs=&fast;s.w=s.h=128;s.tex.psm=psm;
      s.tex.tbp=random()%0x4000;s.tex.tbw=random()%64;s.tex.cpsm=i%3==0?2:i%3==1?0xa:0;s.tex.csa=random()%32;
      s.wms=random()%4;s.wmt=random()%4;s.minu=random()%128;s.maxu=128+random()%128;s.minv=random()%128;s.maxv=128+random()%128;
      fast.reg[ps2_gsTEXA]=(uint64_t(random())<<32)|random();
      int u=int(random()%2048)-1024,v=int(random()%2048)-1024;
      assert(ps2_gsSampler_at(&s,u,v)==ps2_gsSampler_at_reference(&s,u,v));
      // Read-after-write must not reuse decoded texel or palette data.
      fast.clut[random()%512]=random();fast.vram[random()%fast.vram.n]=random();
      assert(ps2_gsSampler_at(&s,u,v)==ps2_gsSampler_at_reference(&s,u,v));
    }
}
