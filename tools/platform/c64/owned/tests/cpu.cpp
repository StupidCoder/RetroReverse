#include "../cpu.h"
#include <array>
#include <cassert>
#include <cstdio>
#include <cstdlib>
#include <cstring>
#include <iostream>
#include <vector>
using namespace rr::c64;
struct Transaction {uint16_t address;uint8_t data;bool write;};
struct Machine {
 Cpu cpu;std::array<uint8_t,65536> ram{};std::vector<Transaction> log;
 void tick(bool irq=false,bool nmi=false,bool rdy=true){auto b=cpu.bus();const auto value=b.write?b.data:ram[b.address];log.push_back({b.address,value,b.write});if(b.write)ram[b.address]=value;cpu.tick(value,irq,nmi,rdy);}
 unsigned instruction(){const auto start=cpu.state.clocks;do{tick();assert(!cpu.faulted());assert(cpu.state.clocks-start<20);}while(!cpu.boundary());return unsigned(cpu.state.clocks-start);}
};
static void require(bool condition,const char* message){if(!condition){std::cerr<<message<<'\n';std::abort();}}
static void unit(){
 Machine m;m.cpu.start(0x200);m.cpu.state.s=0xfd;
 // LDA #$81 / STA $D000 / ASL $D000: both old/new writes must hit the device.
 const uint8_t code[]={0xa9,0x81,0x8d,0,0xd0,0x0e,0,0xd0};std::memcpy(m.ram.data()+0x200,code,sizeof(code));
 assert(m.instruction()==2&&m.cpu.state.a==0x81);assert(m.instruction()==4&&m.ram[0xd000]==0x81);
 m.log.clear();assert(m.instruction()==6&&m.ram[0xd000]==2);assert(m.log[4].write&&m.log[4].data==0x81);assert(m.log[5].write&&m.log[5].data==2);
 // A read stall repeats the exposed transaction; an NMOS write cannot stall.
 m.cpu.start(0x200);m.log.clear();m.tick(false,false,false);m.tick(false,false,false);assert(m.cpu.state.pc==0x200);assert(m.instruction()==2);assert(m.cpu.state.clocks==4);
 m.tick();m.tick();m.tick();assert(m.cpu.bus().write);m.tick(false,false,false);assert(m.cpu.boundary()&&m.ram[0xd000]==0x81);
 // Replay a partially completed RMW, including its pending old-value write.
 m.tick();m.tick();m.tick();m.tick();assert(m.cpu.bus().write);const auto saved=m.cpu.state;const auto ram=m.ram;
 m.log.clear();m.tick();m.tick();const auto forward=m.log;const auto end=m.cpu.state;m.cpu.state=saved;m.ram=ram;m.log.clear();m.tick();m.tick();
 assert(m.cpu.state.pc==end.pc&&m.cpu.state.p==end.p&&m.cpu.state.clocks==end.clocks);assert(m.log.size()==forward.size());for(size_t i=0;i<m.log.size();i++)assert(m.log[i].address==forward[i].address&&m.log[i].data==forward[i].data&&m.log[i].write==forward[i].write);
 // Reset is a sequence of reads, including three stack reads; it never writes.
 m.cpu.start(0x4567);m.cpu.state.s=0;m.ram[0xfffc]=0x34;m.ram[0xfffd]=0x12;m.cpu.reset();m.log.clear();for(int i=0;i<7;i++)m.tick();assert(m.cpu.boundary()&&m.cpu.state.pc==0x1234&&m.cpu.state.s==0xfd);for(const auto& b:m.log)assert(!b.write);assert(m.log[2].address==0x100&&m.log[3].address==0x1ff&&m.log[4].address==0x1fe);
 // CLI defers a level IRQ for one instruction. IRQ pushes PC without increment.
 m.cpu.start(0x200);m.cpu.state.s=0xfd;m.ram[0x200]=0x58;m.ram[0x201]=0xea;m.ram[0x202]=0xea;m.ram[0xfffe]=0;m.ram[0xffff]=3;
 m.tick(true);m.tick(true);assert(!m.cpu.state.interruptPending);m.tick(true);m.tick(true);assert(m.cpu.state.interruptPending);for(int i=0;i<7;i++)m.tick(true);
 assert(m.cpu.state.pc==0x300&&m.cpu.boundary()&&m.cpu.state.s==0xfa);assert(m.ram[0x1fd]==2&&m.ram[0x1fc]==2&&!(m.ram[0x1fb]&B));
 // NMI edge is captured during RDY stalls; a held NMI line does not retrigger.
 m.cpu.start(0x200);m.ram[0x200]=0xea;m.ram[0xfffa]=0;m.ram[0xfffb]=4;m.tick(false,true,false);assert(m.cpu.state.nmiPending);m.tick(false,true);m.tick(false,true);for(int i=0;i<7;i++)m.tick(false,true);assert(m.cpu.state.pc==0x400&&!m.cpu.state.nmiPending);
 // Unsupported instructions stop explicitly, rather than silently acting as NOPs.
 m.cpu.start(0x200);m.ram[0x200]=2;m.tick();assert(m.cpu.faulted());
 std::cout<<"PASS owned NMOS CPU: RMW bus writes, RDY, mid-cycle replay, reset, IRQ/NMI and unsupported-opcode stop\n";
}
// Cycle expectations from NESdev/Visual6502 interrupt polling and hijacking
// traces, independent of this engine's control flow. Inputs are phi2 samples.
static void interruptTests(){
 for(unsigned asserted=0;asserted<4;asserted++){
  Machine m;m.cpu.start(0x200);m.cpu.state.p=U;m.ram[0x200]=0xad;m.ram[0x201]=0;m.ram[0x202]=4;
  for(unsigned i=0;i<4;i++)m.tick(i==asserted);assert(m.cpu.state.interruptPending==(asserted==2));
 }
 for(auto op:{0x58,0x78}){Machine m;m.cpu.start(0x200);m.cpu.state.p=U|(op==0x58?I:0);m.ram[0x200]=uint8_t(op);m.tick(true);m.tick(true);assert(m.cpu.state.interruptPending==(op==0x78));}
 for(bool initialMask:{false,true}){Machine m;m.cpu.start(0x200);m.cpu.state.p=U|(initialMask?I:0);m.cpu.state.s=0xfc;m.ram[0x200]=0x28;m.ram[0x1fd]=U|(initialMask?0:I);for(int i=0;i<4;i++)m.tick(true);assert(m.cpu.state.interruptPending==!initialMask);}
 {Machine m;m.cpu.start(0x200);m.cpu.state.s=0xfa;m.ram[0x200]=0x40;m.ram[0x1fb]=U;m.ram[0x1fc]=0;m.ram[0x1fd]=4;for(int i=0;i<6;i++)m.tick(true);assert(m.cpu.state.pc==0x400&&m.cpu.state.interruptPending);}
 for(bool taken:{false,true})for(bool crossing:{false,true})for(unsigned asserted=0;asserted<4;asserted++){
  Machine m;const uint16_t pc=crossing?0x2fd:0x200;m.cpu.start(pc);m.cpu.state.p=U|(taken?0:Z);m.ram[pc]=0xd0;m.ram[pc+1]=2;
  const unsigned cycles=taken?(crossing?4:3):2;for(unsigned i=0;i<cycles;i++)m.tick(i==asserted);
  assert(m.cpu.state.interruptPending==(asserted==0||(taken&&crossing&&asserted==2)));
 }
 for(unsigned edge=0;edge<7;edge++){
  Machine m;m.cpu.start(0x200);m.cpu.state.s=0xfd;m.ram[0x200]=0;m.ram[0xfffa]=0;m.ram[0xfffb]=4;m.ram[0xfffe]=0;m.ram[0xffff]=3;m.ram[0x300]=m.ram[0x400]=0xea;
  for(unsigned i=0;i<7;i++)m.tick(false,i==edge);assert(m.cpu.state.pc==(edge<4?0x400:0x300));assert(m.ram[0x1fc]==2&&(m.ram[0x1fb]&B));
  if(edge>=4){assert(!m.cpu.state.interruptPending);m.tick();m.tick();assert(m.cpu.state.interruptPending);}
 }
 // Reset cancels an in-flight write transaction; reset cycles must all read.
 {Machine m;m.cpu.start(0x200);m.cpu.state.a=0x55;m.ram[0x200]=0x8d;m.ram[0x201]=0;m.ram[0x202]=4;m.tick();m.tick();m.tick();assert(m.cpu.bus().write);m.cpu.reset();for(int i=0;i<7;i++){assert(!m.cpu.bus().write);m.tick();}assert(m.ram[0x400]==0);}
 // A held NMI is one edge even when the CPU resumes from a read stall.
 {Machine m;m.cpu.start(0x200);m.ram[0x200]=m.ram[0x400]=m.ram[0x401]=0xea;m.ram[0xfffa]=0;m.ram[0xfffb]=4;for(int i=0;i<3;i++)m.tick(false,true,false);m.tick(false,true);m.tick(false,true);for(int i=0;i<7;i++)m.tick(false,true);m.tick(false,true);m.tick(false,true);assert(m.cpu.state.pc==0x401&&!m.cpu.state.interruptPending);}
 // Seven silicon-dependent encodings fail explicitly; JAM is separately identified.
 for(auto op:{0x8b,0xab,0x93,0x9f,0x9b,0x9c,0x9e}){Machine m;m.cpu.start(0x200);m.ram[0x200]=uint8_t(op);m.tick();assert(m.cpu.state.stage==Stage::Fault);}
 {Machine m;m.cpu.start(0x200);m.ram[0x200]=2;m.tick();assert(m.cpu.state.stage==Stage::Jam);m.cpu.reset();for(int i=0;i<7;i++)m.tick();assert(m.cpu.boundary());}
 std::cout<<"PASS CPU timing: phi2 polling, CLI/SEI/PLP/RTI, branch polling windows, NMI/BRK hijacking, RDY/NMI and mid-write reset\n";
}
static uint32_t number(FILE* f,unsigned n){uint32_t v=0;for(unsigned i=0;i<n;i++){int c=std::fgetc(f);require(c!=EOF,"Truncated vectors");v|=uint32_t(c)<<(i*8);}return v;}
static CpuState registers(FILE* f){CpuState s;s.pc=number(f,2);s.s=number(f,1);s.a=number(f,1);s.x=number(f,1);s.y=number(f,1);s.p=number(f,1);return s;}
static void vectors(const char* path){
 FILE* f=std::fopen(path,"rb");require(f,"Cannot open independent test vectors");require(number(f,4)==0x31565043,"Bad vector format");const unsigned count=number(f,4);uint64_t clocks=0;uint32_t digest=2166136261u;const auto hash=[&](uint8_t b){digest=(digest^b)*16777619u;};
 for(unsigned test=0;test<count;test++){
  const auto opcode=number(f,1),index=number(f,2);Machine m;const auto initial=registers(f),expected=registers(f);m.cpu.start(initial.pc);auto& s=m.cpu.state;s.s=initial.s;s.a=initial.a;s.x=initial.x;s.y=initial.y;s.p=initial.p;
  for(unsigned n=number(f,2);n;n--){const auto address=number(f,2);m.ram[address]=number(f,1);}
  std::vector<Transaction> memory;for(unsigned n=number(f,2);n;n--){const auto a=number(f,2),v=number(f,1);memory.push_back({uint16_t(a),uint8_t(v),false});}
  const unsigned cycles=number(f,1);bool ok=true;
  for(unsigned i=0;i<cycles;i++){
   const unsigned address=number(f,2),value=number(f,1),write=number(f,1);const auto b=m.cpu.bus();const auto observed=b.write?b.data:m.ram[b.address];
   if(b.address!=address||observed!=value||b.write!=bool(write)){
    std::fprintf(stderr,"%02X vector %u cycle %u: actual %04X %02X %c expected %04X %02X %c\n",opcode,index,i,b.address,observed,b.write?'W':'R',address,value,write?'W':'R');ok=false;
   }hash(uint8_t(b.address));hash(uint8_t(b.address>>8));hash(observed);hash(b.write);m.tick();
  }
  if(!m.cpu.boundary()||s.pc!=expected.pc||s.s!=expected.s||s.a!=expected.a||s.x!=expected.x||s.y!=expected.y||s.p!=expected.p){
   std::fprintf(stderr,"%02X vector %u registers: actual pc=%04X s=%02X a=%02X x=%02X y=%02X p=%02X; expected pc=%04X s=%02X a=%02X x=%02X y=%02X p=%02X\n",opcode,index,s.pc,s.s,s.a,s.x,s.y,s.p,expected.pc,expected.s,expected.a,expected.x,expected.y,expected.p);ok=false;
  }
  for(const auto& b:memory)if(m.ram[b.address]!=b.data)ok=false;
  require(ok,"Independent vector failed");for(auto b:{uint8_t(s.pc),uint8_t(s.pc>>8),s.s,s.a,s.x,s.y,s.p})hash(b);clocks+=cycles;
 }
 require(std::fgetc(f)==EOF,"Unexpected trailing vector data");std::fclose(f);std::cout<<"PASS "<<count<<" independent instruction vectors / "<<clocks<<" bus cycles; trace digest "<<std::hex<<digest<<std::dec<<"\n";
}
int main(int argc,char** argv){unit();interruptTests();if(argc==2)vectors(argv[1]);return 0;}
