#include "../../platform/gc/browser/core/api.cpp"
#include <cassert>
#include <random>

int main(){
  std::mt19937 random(0x6c);
  gc_Machine m{};m.RAM=Slice<uint8_t>::make(24*1024*1024);
  for(unsigned a=0;a<4096;a++)m.RAM[a]=random();
  for(unsigned a=0;a<4092;a++)assert(gc_Machine_Fetch32(&m,a)==gc_Machine_Fetch32_reference(&m,a));
  for(unsigned a=24*1024*1024-8;a<=24*1024*1024;a++)
    assert(gc_Machine_Fetch32(&m,a)==gc_Machine_Fetch32_reference(&m,a));
  for(int i=0;i<4000;i++){
    gekko_CPU actual{};actual.MSR=random();actual.PC=random();
    for(auto&b:actual.IBAT)for(auto&w:b)w=random();
    for(auto&b:actual.DBAT)for(auto&w:b)w=random();
    actual.LC.Enabled_=(i%3)==0;actual.LC.Base=0xe0000000;
    uint32_t ea=i%5==0?0xe0000020:random();bool insn=i&1,store=i&2;
    auto expected=actual;
    assert(gekko_CPU_Translate(&actual,ea,store,insn)==gekko_CPU_Translate_reference(&expected,ea,store,insn));
    assert(actual.Halted==expected.Halted&&actual.HaltReason==expected.HaltReason);
  }
  // Random complete TEV programs exercise operand routing, swaps, all compare
  // widths, destination registers, bias/scale and alpha rejection.
  gc_gpu g{};
  for(int i=0;i<12000;i++){
    for(auto&v:g.BP)v=random();
    if(i&1)g.BP[0xf3]=(7<<16)|(7<<19); // also compare fully shaded colors
    for(auto&v:g.TevColorReg)for(auto&w:v)w=random();
    for(auto&v:g.TevKonstReg)for(auto&w:v)w=random();
    // Use a valid I8 texture for every map; include textured and untextured stages.
    for(unsigned map=0;map<8;map++){
      unsigned bank=map<4?0:0x20,index=map&3;
      g.BP[0x88+bank+index]=(1<<20)|15|(15<<10);
      g.BP[0x94+bank+index]=0;
    }
    auto t=gc_gpu_tevstate(&g);
    std::array<std::array<uint8_t,4>,2>ras{};
    std::array<gc_texCoord,8>tc{};
    for(auto&c:ras)for(auto&v:c)v=random();
    for(auto&v:tc){v.s=float(int(random()%400)-200)/100;v.t=float(int(random()%400)-200)/100;}
    assert(gc_gpu_shade(&g,&m,&t,&ras,&tc)==gc_gpu_shade_reference(&g,&m,&t,&ras,&tc));
  }
}
