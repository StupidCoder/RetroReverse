#include "../../platform/nds/browser/core/api.cpp"
#include <cassert>

int main() {
 dsmachine_Machine m{}; dsmachine_core c9{},c7{};arm_CPU cpu9{},cpu7{};
 m.ARM9=&c9;m.ARM7=&c7;c9.cpu=&cpu9;c7.cpu=&cpu7;c9.m=c7.m=&m;c9.arm9=true;
 m.ram=Slice<uint8_t>::make(4194304);m.swram=Slice<uint8_t>::make(32768);
 m.pal=Slice<uint8_t>::make(2048);m.oam=Slice<uint8_t>::make(2048);
 c9.itcm=Slice<uint8_t>::make(32768);c9.itcmBase=0x01ff8000;
 c9.dtcm=Slice<uint8_t>::make(16384);c9.dtcmBase=0x027c0000;
 c7.low=Slice<uint8_t>::make(16384);c7.wram7=Slice<uint8_t>::make(65536);
 m.vram=dsmachine_newVRAM();
 for(auto*s:{&m.ram,&m.swram,&m.pal,&m.oam,&c9.itcm,&c9.dtcm,&c7.low,&c7.wram7})
   for(int64_t i=0;i<s->n;i++)(*s)[i]=uint8_t(i*19+(i>>8));
 std::vector<uint32_t> addresses;
 for(uint32_t edge:{0u,0x4000u,0x01ff8000u,0x02000000u,0x02400000u,0x027c0000u,0x027c4000u,0x03000000u,0x03004000u,0x03008000u,0x03800000u,0x03810000u,0x05000000u,0x05000800u,0x07000000u,0x07000800u,0xfffffffcu})
   for(int d=-4;d<=4;d++)addresses.push_back(edge+d);
 for(int mode=0;mode<4;mode++)for(auto*c:{&c9,&c7}) {
   m.wramcnt=mode;dsmachine_bus b{c};
   for(auto a:addresses) {
     auto expected=[&](unsigned n){uint32_t v=0;for(unsigned i=0;i<n;i++)v|=uint32_t(dsmachine_bus_read(&b,a+i))<<(8*i);return v;};
     assert(dsmachine_bus_Read(&b,a)==expected(1));
     assert(dsmachine_bus_Read16(&b,a)==(a>>24==4?dsmachine_bus_Read16Reference(&b,a):expected(2)));
     assert(dsmachine_bus_Read32(&b,a)==(a>>24==4?dsmachine_bus_Read32Reference(&b,a):expected(4)));
     auto old=expected(4);dsmachine_bus_Write32(&b,a,0x563412ab);auto fast=expected(4);
     for(unsigned i=0;i<4;i++)dsmachine_bus_WriteReference(&b,a+i,old>>(8*i));
     for(unsigned i=0;i<4;i++)dsmachine_bus_WriteReference(&b,a+i,0x563412ab>>(8*i));
     assert(fast==expected(4));
   }
 }
 dsmachine_bus b{&c9};std::vector<uint32_t> observed;
 m.OnRead=[&](bool,uint32_t a,uint8_t,uint32_t){observed.push_back(a);};
 dsmachine_bus_Read32(&b,0x02000000);assert((observed==std::vector<uint32_t>{0x02000000,0x02000001,0x02000002,0x02000003}));
 observed.clear();m.OnWrite=[&](bool,uint32_t a,uint8_t,uint32_t){observed.push_back(a);};
 dsmachine_bus_Write32(&b,0x02000000,0x12345678);assert((observed==std::vector<uint32_t>{0x02000000,0x02000001,0x02000002,0x02000003}));
 m.OnRead={};m.OnWrite={};cpu9.bus=cpu9.wide=&b;cpu9.Coproc=dsmachine_cp15(&c9);
 cpu9.R[15]=0x02000000;dsmachine_bus_Write32(&b,0x02000000,0xee070f90);
 dsmachine_bus_Write32(&b,0x02000004,0xe2800001);
 dsmachine_Machine_runQuantum(&m,&c9,64,{},{});
 assert(c9.wfi&&cpu9.Instrs==1&&cpu9.R[15]==0x02000004);
 rrstate::Archive saved;auto original=&c9;saved(original);
 rrstate::Archive restored(saved.bytes.data(),saved.bytes.size());dsmachine_core*copy=nullptr;restored(copy);restored.finish();assert(copy->wfi);
 c9.ie=c9.if_=1;c9.ime=false;dsmachine_Machine_deliver(&m,&c9);assert(c9.wfi);
 c9.ime=true;cpu9.IRQDisable=true;dsmachine_Machine_deliver(&m,&c9);assert(!c9.wfi&&cpu9.R[15]==0x02000004);
 dsmachine_Machine_runQuantum(&m,&c9,1,{},{});assert(cpu9.R[0]==1);
 std::cout<<"DS bus mirrors, TCM boundaries, WRAM modes, unaligned access and watched byte order passed\n";
}
