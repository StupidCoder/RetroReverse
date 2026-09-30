// Debugging is opt-in: the normal rr_run loop has no debugger hooks.
namespace debug {
static int mode=0,target=0,targetSpace=0,result=0;
static bool skip=false,inFlight=false,entry=false;
static int instruction=-1,last=-1;
static uint64_t lastCycle=~uint64_t(0);
static bool boundary(){return (machine.pins&M6502_SYNC)!=0;}
static bool interrupt(){return (machine.cpu.irq_pip&0x400)||(machine.cpu.nmi_pip&0xfc00)||(machine.pins&M6502_RES);}
}
namespace {static void invalidateDebug(){debug::last=-1;debug::lastCycle=~uint64_t(0);debug::result=0;debug::inFlight=false;}}
extern "C" int rr_debug_begin(int mode,int target,int next_match){
 if(!initialized||mode<0||mode>2||target<0||target>65535)return 0;
 debug::mode=mode;debug::target=target;debug::targetSpace=observation::cpuSpace(target);debug::result=0;debug::inFlight=false;
 debug::skip=next_match&&debug::boundary()&&M6502_GET_ADDR(machine.pins)==target;
 // A step from a partial instruction must first explicitly normalize.
 if(mode==1&&!debug::boundary())return 0;
 return 1;
}
extern "C" int rr_debug_run(int cycles){
 using namespace debug;
 if(cycles<0||cycles>10000)return -1;
 if(result)return result;
 for(int i=0;;i++){
  if(boundary()){
   if(inFlight){
    last=entry?-1:instruction;lastCycle=ctx.cycles;inFlight=false;
    if(mode==1)return result=entry?3:2;
   }
   if(mode==0)return result=1;
   if(mode==2&&!skip&&!interrupt()&&M6502_GET_ADDR(machine.pins)==target)return result=observation::cpuSpace(target)==targetSpace?4:5;
  }
  const int opcode=machine.cpu.IR>>3;
  if(!boundary()&&(machine.cpu.IR&7)==1&&(opcode==0x02||opcode==0x12||opcode==0x22||opcode==0x32||opcode==0x42||opcode==0x52||opcode==0x62||opcode==0x72||opcode==0x92||opcode==0xb2||opcode==0xd2||opcode==0xf2))return result=6;
  if(i==cycles)return 0;
  if(boundary()&&!(machine.pins&M6502_RDY)){
   inFlight=true;entry=interrupt();instruction=M6502_GET_ADDR(machine.pins);skip=false;
  }
  if(rr_run(1,0,0)<0)return result=-1;
 }
}
extern "C" const char* rr_debug_snapshot(int address){
 static std::string out;
 const bool boundary=debug::boundary();
 const int pc=boundary?M6502_GET_ADDR(machine.pins):-1;
 if(address<0)address=boundary?pc:ctx.pc;
 if(address>65535)return "{\"error\":\"Invalid CPU address\"}";
 std::ostringstream s;
 s<<"{\"boundary\":"<<(boundary?"true":"false")<<",\"nextPC\":"<<(boundary?std::to_string(pc):"null")
  <<",\"lastExecutedPC\":"<<(debug::lastCycle==ctx.cycles&&debug::last>=0?std::to_string(debug::last):"null")
  <<",\"cycle\":\""<<ctx.cycles<<"\",\"bank\":"<<int(machine.cpu_port)
  <<",\"interruptPending\":"<<(debug::interrupt()?"true":"false")<<",\"stalled\":"<<((machine.pins&M6502_RDY)?"true":"false")
  <<",\"registers\":{\"A\":"<<int(machine.cpu.A)<<",\"X\":"<<int(machine.cpu.X)<<",\"Y\":"<<int(machine.cpu.Y)<<",\"S\":"<<int(machine.cpu.S)<<",\"P\":"<<int(machine.cpu.P)<<"},\"address\":"<<address<<",\"bytes\":[";
 for(int i=0;i<256;i++){
  const uint16_t a=uint16_t(address+i);const int space=observation::cpuSpace(a);
  if(i)s<<',';
  // Color RAM has an open-bus high nibble; CPU ports and device registers are
  // deliberately unavailable instead of performing a destructive live read.
  if(space==1||space==2)s<<"null";else s<<int(mem_rd(&machine.mem_cpu,a));
 }
 s<<"],\"mapping\":[";
 for(int i=0;i<256;i++){if(i)s<<',';s<<observation::cpuSpace(uint16_t(address+i));}
 s<<"]}";out=s.str();return out.c_str();
}
