namespace bridge {
bool beginJob(int mode,int target,int next){
 if(!initialized||mode<0||mode>5||target<0||target>65535)return false;
 Job j;j.active=true;j.mode=mode;j.target=target;j.space=space(uint16_t(target));
 j.skip=next&&boundary()&&board.state.cpu.pc==target;
 if((mode==1||mode==3||mode==4)&&!boundary())return false;
 if(mode==3||mode==4){
  if(interrupt())return false;j.returnSP=board.state.cpu.s;
  if(mode==3){
   if(board.peek(board.state.cpu.pc)!=0x20)j.mode=1;
   else j.target=uint16_t(board.state.cpu.pc+3);
  }else{
   const auto sp=board.state.cpu.s;
   j.target=uint16_t((board.state.ram[0x100+uint8_t(sp+1)]|(board.state.ram[0x100+uint8_t(sp+2)]<<8))+1);
   const auto call=uint16_t(j.target-3);
   if(space(call)!=0||board.peek(call)!=0x20)return false;
   j.calls=1;j.returnSP=uint8_t(sp+2);
  }
  j.space=space(uint16_t(j.target));
 }
 job=j;writeEvent="null";return true;
}
}
extern "C" {
int rr_debug_begin(int mode,int target,int next){if(mode>4)return 0;return beginJob(mode,target,next);}
int rr_debug_watch(int address,int value,int mask,int writer){
 if(address<2||address>65535||space(uint16_t(address))!=0||value<0||value>255||mask<0||mask>255||(value&mask)!=value||writer < -1||writer>65535)return 0;
 if(!beginJob(5,address,0))return 0;job.value=value;job.mask=mask;job.writer=writer;return 1;
}
const char* rr_debug_event(){return writeEvent.c_str();}
int rr_debug_run(int cycles){
 if(!initialized||!job.active||cycles<0||cycles>10000){error="Invalid debugger job or execution budget.";return -1;}
 if(job.result)return job.result;
 rrprof::Scope profile(1,"Owned debugger");stopReason=0;
 for(int done=0;;done++){
  if(fault())return job.result=6;
  if(boundary()){
   if(job.inFlight){
    job.inFlight=false;
    if(job.mode==1)return job.result=job.entry?3:2;
    if(job.mode==3||job.mode==4){
     if(job.entry||job.opcode==0)++job.interrupts;
     else if(job.opcode==0x40&&job.interrupts>0)--job.interrupts;
     else if(!job.interrupts){
      if(job.opcode==0x20)++job.calls;
      if(job.opcode==0x60&&--job.calls==0){
       if(board.state.cpu.pc!=job.target||board.state.cpu.s!=job.returnSP)return job.result=14;
       return job.result=space(uint16_t(job.target))==job.space?12:5;
      }
     }
    }
   }
   if(job.mode==0)return job.result=1;
   if(job.mode==2&&!job.skip&&!interrupt()&&board.state.cpu.pc==job.target)return job.result=space(uint16_t(job.target))==job.space?4:5;
  }
  if(job.mode==5&&space(uint16_t(job.target))!=job.space)return job.result=5;
  if(done==cycles)return 0;
  const bool wasBoundary=boundary();const bool entry=interrupt();
  const bool ok=advance();const auto& bus=board.state.lastBus;
  // A pending SYNC held by RDY has not started an instruction. Capture only
  // the opcode actually consumed by the CPU, after the VIC computes BA/AEC.
  if(wasBoundary&&bus.valid&&bus.ready&&board.state.cpu.stage!=Stage::Fetch){job.inFlight=true;job.entry=entry;job.opcode=bus.value;job.skip=false;}
  if(!ok)return job.result=fault()?6:-1;
  if(job.mode==5&&bus.valid&&bus.write&&bus.address==job.target&&space(bus.address,true)==0&&(bus.value&job.mask)==job.value&&(job.writer<0||ctx.pc==job.writer)){
   writeEvent="{\"address\":"+std::to_string(bus.address)+",\"value\":"+std::to_string(bus.value)+",\"pc\":"+std::to_string(ctx.pc)+",\"cycle\":\""+std::to_string(board.state.cycles)+"\"}";
   return job.result=13;
  }
 }
}
const char* rr_debug_snapshot(int address){
 static std::string out;
 if(address < -1||address>65535)return "{\"error\":\"Invalid CPU address\"}";
 const auto& c=board.state.cpu;const bool at=boundary();if(address<0)address=at?c.pc:ctx.pc;
 std::ostringstream s;s<<"{\"boundary\":"<<(at?"true":"false")<<",\"nextPC\":"<<(at?std::to_string(c.pc):"null")
  <<",\"lastExecutedPC\":"<<(lastExecuted>=0&&lastExecutedCycle==board.state.cycles?std::to_string(lastExecuted):"null")<<",\"cycle\":\""<<board.state.cycles<<"\",\"bank\":"<<unsigned(board.port())
  <<",\"interruptPending\":"<<(interrupt()?"true":"false")<<",\"stalled\":"<<((!board.state.vic.ba||!board.state.vic.aec)?"true":"false")
  <<",\"registers\":{\"A\":"<<unsigned(c.a)<<",\"X\":"<<unsigned(c.x)<<",\"Y\":"<<unsigned(c.y)<<",\"S\":"<<unsigned(c.s)<<",\"P\":"<<unsigned(c.p)<<"},\"address\":"<<address<<",\"bytes\":[";
 for(unsigned i=0;i<256;i++){if(i)s<<',';const auto a=uint16_t(address+i);const auto sp=space(a);if(sp==1||sp==2)s<<"null";else s<<unsigned(board.peek(a));}
 s<<"],\"mapping\":[";for(unsigned i=0;i<256;i++){if(i)s<<',';s<<space(uint16_t(address+i));}s<<"]}";out=s.str();return out.c_str();
}
int rr_debug_edit(int count){
 if(!initialized||!boundary()||interrupt()||count<1||count>256)return 0;
 for(int i=0;i<count;i++){
  const uint16_t a=uint16_t(input[i*4]|unsigned(input[i*4+1])<<8);
  if(a<2||space(a)!=0||board.state.ram[a]!=input[i*4+2])return 0;
  for(int j=0;j<i;j++)if(a==(input[j*4]|unsigned(input[j*4+1])<<8))return 0;
 }
 for(int i=0;i<count;i++){const auto a=uint16_t(input[i*4]|unsigned(input[i*4+1])<<8);board.state.ram[a]=input[i*4+3];render::edited(a);}
 // The owned CPU has a pending fetch, so an edit at PC needs no synthetic
 // prefetch replacement. The actual next read consumes the patched byte.
 invalidateDebug();return 1;
}
}
