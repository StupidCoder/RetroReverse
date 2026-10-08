#include "../../platform/n3ds/browser/core/api.cpp"
#include <cassert>

int main(){
 auto m=arenaNew(n3ds_Machine{});m->CPU=arm_NewCPU(m);auto c=m->CPU;
 n3ds_Machine_mapRegion(m,"zero",0,Slice<uint8_t>::make(4096));
 auto ram=n3ds_Machine_mapRegion(m,"ram",0x10000,Slice<uint8_t>::make(12288));
 n3ds_Machine_mapRegion(m,"alias",0x20000,ram->data);
 n3ds_Machine_mapRegion(m,"partial",0x30003,Slice<uint8_t>::make(13));
 n3ds_Machine_mapRegion(m,"high",0xfffffff0,Slice<uint8_t>::make(8));
 auto wide=arenaNew(n3ds_Machine{});wide->CPU=arm_NewCPU(wide);
 n3ds_Machine_mapRegion(wide,"different wide bus",0x10000,Slice<uint8_t>::make(12288));
 using Event=std::array<uint32_t,4>;std::vector<Event> events;
 auto reset=[&](){uint32_t seed=0x3d5a1234;for(auto b:{m,wide})for(auto r:b->regions)for(auto&v:r->data){seed^=seed<<13;seed^=seed>>17;seed^=seed<<5;v=seed;}events.clear();};
 auto bytes=[&](){std::vector<uint8_t> out;for(auto b:{m,wide})for(auto r:b->regions)out.insert(out.end(),r->data.p,r->data.p+r->data.n);return out;};
 std::vector<uint32_t> addresses;
 for(uint32_t edge:{0u,4096u,0x10000u,0x11000u,0x12000u,0x13000u,0x20000u,0x22000u,0x23000u,0x30003u,0x30010u,0xfffffff0u,0xfffffff8u})for(int d=-4;d<=4;d++)addresses.push_back(edge+d);
 auto check=[&](){
  for(auto arch:{arm_V4T,arm_V5TE,arm_V6K})for(auto bus:{static_cast<n3ds_Machine*>(nullptr),m,wide}){
   c->Arch=arch;c->wide=bus;c->R[15]=0x10400;
   for(auto a:addresses){
    auto reads=[&](){return std::array<uint32_t,3>{arm_CPU_read16(c,a),arm_CPU_read32aligned(c,a),arm_CPU_read32(c,a)};};
    reset();rr_graphics_enable(0);auto expected=reads();auto observed=events;
    reset();rr_graphics_enable(1);assert(reads()==expected);assert(events==observed);
    for(unsigned kind=0;kind<3;kind++)for(uint32_t value:{0u,0xffffffffu,0x12345678u,0x80000001u}){
     auto write=[&](){if(kind==0)arm_CPU_write16(c,a,value);else if(kind==1)arm_CPU_write32aligned(c,a,value);else arm_CPU_write32(c,a,value);};
     reset();rr_graphics_enable(0);write();auto expectedBytes=bytes();observed=events;
     reset();rr_graphics_enable(1);write();assert(bytes()==expectedBytes);assert(events==observed);
    }
   }
  }
 };
 check();rr_graphics_enable(0);assert(!rrarm::ordinary(m,0x10000,4));rr_graphics_enable(1);
 rrarm::memoryEnabled=false;assert(!rrarm::ordinary(m,0x10000,4));rrarm::memoryEnabled=true;
 assert(rrarm::ordinary(m,0x10001,4)==ram->data.p+1);
 assert(!rrarm::ordinary(m,0x10fff,4));assert(!rrarm::ordinary(m,0x30003,2));assert(!rrarm::ordinary(m,UINT32_MAX,4));
 auto replacement=n3ds_Machine_mapRegion(m,"replacement",0x21000,Slice<uint8_t>::make(4096));
 assert(rrarm::ordinary(m,0x21000,4)==replacement->data.p);check();
 for(auto b:{m,wide}){
  b->RWatchLo=b->WatchLo=0;b->RWatchHi=b->WatchHi=UINT32_MAX;
  b->OnRead=[&](uint32_t a,uint32_t v,uint32_t pc){events.push_back({0,a,v,pc});};
  b->OnWrite=[&](uint32_t a,uint32_t v,uint32_t pc){events.push_back({1,a,v,pc});};
 }
 check();assert(!rrarm::ordinary(m,0x10000,4));
 m->OnRead={};m->OnWrite={};wide->OnRead={};wide->OnWrite={};
 c->wide=nullptr;c->Arch=arm_V6K;m->HidTrace=true;m->hidSharedAddr=0x10000;
 for(bool fast:{false,true}){
  rr_graphics_enable(fast);m->hidReadHist={};m->hidReadPC={};
  arm_CPU_read32(c,0x10001);assert(m->hidReadHist[0]==3&&m->hidReadHist[4]==1);
  assert(m->hidReadPC[0]==arm_CPU_PC(c));assert(!rrarm::ordinary(m,0x10000,4));
 }
 m->HidTrace=false;
 // Exercise actual ARM/Thumb fetch, load/store and flags, including crossing a
 // page. Self-modifying code must be visible at the very next instruction.
 auto state=[&](){return std::array<uint64_t,24>{c->R[0],c->R[1],c->R[2],c->R[3],c->R[4],c->R[5],c->R[6],c->R[7],c->R[8],c->R[9],c->R[10],c->R[11],c->R[12],c->R[13],c->R[14],c->R[15],uint64_t(c->N),uint64_t(c->Z),uint64_t(c->C),uint64_t(c->V),uint64_t(c->Thumb),c->Instrs,uint64_t(c->Halted),c->cur};};
 for(bool thumb:{false,true})for(uint32_t addr:{0x10001u,0x10ffeu,0x10fffu,0x21001u}){
  auto run=[&](bool fast){
   reset();*c=*arm_NewCPU(m);c->Arch=arm_V6K;c->Thumb=thumb;c->R[15]=0x12000;c->R[1]=addr;
   if(thumb){unsigned i=0;for(uint32_t op:{0x6808u,0x3001u,0x6008u,0x880au,0x804au})n3ds_Machine_Write16(m,0x12000+2*i++,op);}
   else {unsigned i=0;for(uint32_t op:{0xe5910000u,0xe2900001u,0xe5810000u,0xe1d120b0u,0xe1c120b2u})n3ds_Machine_Write32(m,0x12000+4*i++,op);}
   rr_graphics_enable(fast);std::vector<std::array<uint64_t,24>> states;
   for(int i=0;i<5;i++){arm_CPU_Step(c);assert(!c->Halted);states.push_back(state());}
   c->R[15]=0x12000;
   if(thumb)arm_CPU_write16(c,0x12000,0x205a);else arm_CPU_write32(c,0x12000,0xe3a0005a);
   arm_CPU_Step(c);assert(c->R[0]==0x5a);states.push_back(state());
   return std::make_pair(states,bytes());
  };
  auto expected=run(false);assert(run(true)==expected);
 }
 // Discarding the index uses the original linear lookup, with no retained RAM
 // pointer surviving a remap or a save-state's reconstructed machine.
 n3ds_Machine_clearPages(m);rr_graphics_enable(1);assert(!rrarm::ordinary(m,0x10000,4));check();
 std::cout<<"3DS ARM memory: Reference parity for alignment, wrap, aliases, remaps, wide bus, watched byte events, HID and self-modifying ARM/Thumb instructions\n";
}
