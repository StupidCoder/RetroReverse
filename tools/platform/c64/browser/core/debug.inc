// Debugging is opt-in: the normal rr_run loop has no debugger hooks.
namespace debug {
static int mode=0,target=0,targetSpace=0,result=0;
static bool skip=false,inFlight=false,entry=false;
static int instruction=-1,last=-1, fetched=0,calls=0,interrupts=0,returnSP=0;
static int watchValue=0,watchMask=0,watchWriter=-1;
static std::string event="null";
static uint64_t lastCycle=~uint64_t(0);
static bool boundary(){return (machine.pins&M6502_SYNC)!=0;}
static bool interrupt(){return (machine.cpu.irq_pip&0x400)||(machine.cpu.nmi_pip&0xfc00)||(machine.pins&M6502_RES);}
}
namespace {static void invalidateDebug(){debug::last=-1;debug::lastCycle=~uint64_t(0);debug::result=0;debug::inFlight=false;debug::event="null";}}
extern "C" int rr_debug_begin(int mode,int target,int next_match){
 if(!initialized||mode<0||mode>5||target<0||target>65535)return 0;
 debug::mode=mode;debug::target=target;debug::targetSpace=observation::cpuSpace(target);debug::result=0;debug::inFlight=false;debug::event="null";debug::calls=0;debug::interrupts=0;
 debug::skip=next_match&&debug::boundary()&&M6502_GET_ADDR(machine.pins)==target;
 // A step from a partial instruction must first explicitly normalize.
 if((mode==1||mode==3||mode==4)&&!debug::boundary())return 0;
 if(mode==3||mode==4){
  if(debug::interrupt())return 0;
  debug::returnSP=machine.cpu.S;
  if(mode==3){
   if(M6502_GET_DATA(machine.pins)!=0x20){debug::mode=1;return 1;}
   debug::target=uint16_t(M6502_GET_ADDR(machine.pins)+3);
  }else{
   const uint8_t sp=machine.cpu.S;
   debug::target=uint16_t((machine.ram[0x100+uint8_t(sp+1)]|(machine.ram[0x100+uint8_t(sp+2)]<<8))+1);
   // Only balanced conventional JSR frames are supported. This is not an inferred call stack.
   const uint16_t call=uint16_t(debug::target-3);
   if(observation::cpuSpace(call)>0||mem_rd(&machine.mem_cpu,call)!=0x20)return 0;
   debug::calls=1;debug::returnSP=uint8_t(sp+2);
  }
  debug::targetSpace=observation::cpuSpace(debug::target);
 }
 return 1;
}
extern "C" int rr_debug_watch(int address,int value,int mask,int writer){
 if(address<2||address>65535||observation::cpuSpace(address)!=0||value<0||value>255||mask<0||mask>255||(value&mask)!=value||writer < -1||writer>65535)return 0;
 if(!rr_debug_begin(5,address,0))return 0;
 debug::watchValue=value;debug::watchMask=mask;debug::watchWriter=writer;return 1;
}
extern "C" const char* rr_debug_event(){return debug::event.c_str();}
extern "C" int rr_debug_run(int cycles){
 using namespace debug;
 if(cycles<0||cycles>10000)return -1;
 if(result)return result;
 for(int i=0;;i++){
  if(boundary()){
   if(inFlight){
    last=entry?-1:instruction;lastCycle=ctx.cycles;inFlight=false;
    if(mode==1)return result=entry?3:2;
    if(mode==3||mode==4){
     if(entry||fetched==0x00)interrupts++;
     else if(fetched==0x40&&interrupts>0)interrupts--;
     else if(!interrupts){
      if(fetched==0x20)calls++;
      if(fetched==0x60&&--calls==0){
       if(M6502_GET_ADDR(machine.pins)!=target||machine.cpu.S!=returnSP)return result=14;
       return result=observation::cpuSpace(target)==targetSpace?12:5;
      }
     }
    }
   }
   if(mode==0)return result=1;
   if(mode==2&&!skip&&!interrupt()&&M6502_GET_ADDR(machine.pins)==target)return result=observation::cpuSpace(target)==targetSpace?4:5;
  }
  const int opcode=machine.cpu.IR>>3;
  if(!boundary()&&(machine.cpu.IR&7)==1&&(opcode==0x02||opcode==0x12||opcode==0x22||opcode==0x32||opcode==0x42||opcode==0x52||opcode==0x62||opcode==0x72||opcode==0x92||opcode==0xb2||opcode==0xd2||opcode==0xf2))return result=6;
  if(i==cycles)return 0;
  if(boundary()&&!(machine.pins&M6502_RDY)){
   inFlight=true;entry=interrupt();instruction=M6502_GET_ADDR(machine.pins);fetched=M6502_GET_DATA(machine.pins);skip=false;
  }
  if(mode==5&&observation::cpuSpace(target)!=targetSpace)return result=5;
  if(rr_run(1,0,0)<0)return result=-1;
  if(mode==5&&!(machine.pins&M6502_RW)&&M6502_GET_ADDR(machine.pins)==target&&observation::cpuSpace(target)==0&&(M6502_GET_DATA(machine.pins)&watchMask)==watchValue&&(watchWriter<0||ctx.pc==watchWriter)){
   event="{\"address\":"+std::to_string(target)+",\"value\":"+std::to_string(M6502_GET_DATA(machine.pins))+",\"pc\":"+std::to_string(ctx.pc)+",\"cycle\":\""+std::to_string(ctx.cycles)+"\"}";
   return result=13;
  }
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
// Experiment transaction: input records are u16le address, expected byte, new byte.
// Check the entire batch before touching RAM. Ports, ROM and I/O are excluded.
extern "C" int rr_debug_edit(int count){
 if(!initialized||!debug::boundary()||debug::interrupt()||count<1||count>256)return 0;
 for(int i=0;i<count;i++){
  const unsigned a=input[i*4]|unsigned(input[i*4+1])<<8;
  if(a<2||observation::cpuSpace(a)!=0||machine.ram[a]!=input[i*4+2])return 0;
  for(int j=0;j<i;j++)if(a==(input[j*4]|unsigned(input[j*4+1])<<8))return 0;
 }
 for(int i=0;i<count;i++)machine.ram[input[i*4]|unsigned(input[i*4+1])<<8]=input[i*4+3];
 // SYNC already holds the fetched byte; a patch at this PC must replace it too.
 const uint16_t pc=M6502_GET_ADDR(machine.pins);
 if(observation::cpuSpace(pc)==0)M6502_SET_DATA(machine.pins,machine.ram[pc]);
 invalidateDebug();return 1;
}
