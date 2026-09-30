#pragma once
// Opt-in instruction-boundary debugging. Safe peeks bypass VGA/device handlers.
static bool rrDebugHookPending=false,rrDebugActive=false,rrDebugSkip=false;
static int rrDebugMode=0;static uint32_t rrDebugTarget=0;static int64_t rrDebugTargetMode=0;
static int64_t rrDebugTargetCS=-1;static uint32_t rrDebugTargetBase=0;
static uint32_t rrPC(){auto*c=rrCPU();return x86_CPU_linear(c,x86_CPU_segBase(c,1),c->IP);}
static int rrPeek(uint32_t a){auto mem=realMachine?realMachine->Mem:protectedMachine->Mem;if(a>=mem.n||(a>=0xa0000&&a<0xc0000))return -1;return mem[a];}
static std::string rrQuote(const std::string&s){std::string o="\"";for(unsigned char c:s){if(c=='"'||c=='\\')o+='\\';if(c>=32)o+=c;}return o+'"';}
static x86_Inst rrDecode(uint32_t address,bool current=false,uint32_t offset=0){auto*c=rrCPU();std::array<uint8_t,15> bytes{};Slice<uint8_t>b;b.p=bytes.data();b.n=b.c=15;int count=0;for(;count<15;count++){uint32_t a=current?x86_CPU_linear(c,x86_CPU_segBase(c,1),c->IP+offset+count):address+count;auto v=rrPeek(a);if(v<0)break;b[count]=v;}auto code=sub(b,0,count);x86_dec decoder{};decoder.mem=code;decoder.addr=current?x86_CPU_ipMask(c,c->IP+offset):address;decoder.defSize=c->Mode?32:16;auto result=x86_dec_finish(&decoder,code,decoder.addr,x86_dec_run(&decoder));if(current&&result.HasTarget&&result.Mnem!="CALLF"&&result.Mnem!="JMPF"){auto old=go_fmt_Sprintf(std::string("$%08X"),result.Target);result.Target=x86_CPU_linear(c,x86_CPU_segBase(c,1),result.Target);auto pos=result.Text.rfind(old);if(pos!=std::string::npos)result.Text.replace(pos,old.size(),go_fmt_Sprintf(std::string("$%08X"),result.Target));}if(result.Len>15){result.Len=1;result.Mnem=".byte";result.Text="Unsupported instruction length";}return result;}
// Returns interrupt delivery separately, before executing the handler. Remember
// the consumed device hook across debug/normal execution and serialized states.
static int rrStep(bool debug){
 auto*c=rrCPU();if(c->Halted)return 6;
 if(!debug&&!rrDebugHookPending){x86_CPU_Step(c);if(rrdos::collecting)rrdos::observe(c->Steps);return c->Halted?6:2;}
 if(!rrDebugHookPending){auto ip=c->IP;auto cs=c->Seg[1];auto mode=c->Mode;if(c->OnStep)c->OnStep(c);rrDebugHookPending=true;if(c->Halted)return 6;if(debug&&(ip!=c->IP||cs!=c->Seg[1]||mode!=c->Mode))return 3;}
 if(debug){
  if(rrDebugMode==2&&rrPC()==rrDebugTarget){if(c->Mode!=rrDebugTargetMode||(rrDebugTargetCS>=0&&(c->Seg[1]!=rrDebugTargetCS||x86_CPU_segBase(c,1)!=rrDebugTargetBase)))return 5;if(!rrDebugSkip)return 4;}
  bool rep=false,wide=c->Mode!=0;int op=-1;
  for(unsigned i=0;i<15;i++){op=rrPeek(x86_CPU_linear(c,x86_CPU_segBase(c,1),c->IP+i));if(op<0)return 7;
   if(op==0x67){wide=c->Mode==0;continue;}if(op==0xf2||op==0xf3){rep=true;continue;}
   if(op==0x26||op==0x2e||op==0x36||op==0x3e||op==0x64||op==0x65||op==0x66||op==0xf0){op=-1;continue;}break;
  }
  if(op<0||op==0xf2||op==0xf3||op==0x67)return 7;
  if(rep&&wide&&c->Regs[1]>65536&&((op>=0xa4&&op<=0xaf)||(op>=0x6c&&op<=0x6f)))return 7;
 }
 rrDebugSkip=false;auto hook=c->OnStep;c->OnStep={};rrDebugHookPending=false;
 try{x86_CPU_Step(c);}catch(...){c->OnStep=hook;throw;}c->OnStep=hook;
 if(rrdos::collecting)rrdos::observe(c->Steps);
 return c->Halted?6:2;
}
extern "C" {
double rr_cycle(){return rrCPU()?double(rrCPU()->Steps):0;}
uint8_t*rr_ram(){return realMachine?realMachine->Mem.p:protectedMachine?protectedMachine->Mem.p:nullptr;}
uint32_t rr_ram_size(){return realMachine?realMachine->Mem.n:protectedMachine?protectedMachine->Mem.n:0;}
const char*rr_debug_snapshot(int32_t requested){
 auto*c=rrCPU();if(!c)return "{}";uint32_t at=requested<0?rrPC():uint32_t(requested);std::ostringstream o;
 o<<"{\"architecture\":\"x86\",\"mode\":"<<rrQuote(c->Mode?"protected32":"real16")<<",\"boundary\":true,\"nextPC\":"<<rrPC()<<",\"cycle\":\""<<c->Steps<<"\",\"bank\":"<<c->Mode<<",\"address\":"<<at<<",\"loadSegment\":"<<(realMachine?realMachine->loadSeg:0)<<",\"ip\":"<<c->IP<<",\"cs\":"<<c->Seg[1]<<",\"csBase\":"<<x86_CPU_segBase(c,1)<<",\"hookPending\":"<<(rrDebugHookPending?"true":"false")<<",\"registers\":{";
 const char*names[]={"EAX","ECX","EDX","EBX","ESP","EBP","ESI","EDI"};for(int i=0;i<8;i++){if(i)o<<',';o<<rrQuote(names[i])<<':'<<c->Regs[i];}o<<",\"FLAGS\":"<<x86_CPU_EFlags(c)<<"},\"segments\":[";
 for(int i=0;i<6;i++){if(i)o<<',';o<<"{\"selector\":"<<c->Seg[i]<<",\"base\":"<<x86_CPU_segBase(c,i)<<'}';}o<<"],\"bytes\":[";
 for(int i=0;i<256;i++){if(i)o<<',';auto b=rrPeek(at+i);if(b<0)o<<"null";else o<<b;}o<<"],\"mapping\":[";for(int i=0;i<256;i++){if(i)o<<',';o<<(rrPeek(at+i)<0?2:0);}o<<"],\"instructions\":[";
 for(unsigned off=0,rows=0;off<128&&rows<32;rows++){auto address=requested<0?x86_CPU_linear(c,x86_CPU_segBase(c,1),c->IP+off):at+off;auto d=rrDecode(address,requested<0,off);if(rows)o<<',';o<<"{\"address\":"<<address<<",\"length\":"<<d.Len<<",\"text\":"<<rrQuote(d.Text)<<",\"supported\":"<<(d.Mnem==".byte"?"false":"true")<<",\"bytes\":[";for(int j=0;j<d.Len;j++){if(j)o<<',';auto v=rrPeek(requested<0?x86_CPU_linear(c,x86_CPU_segBase(c,1),c->IP+off+j):address+j);if(v<0)o<<"null";else o<<v;}o<<"]}";if(d.Mnem==".byte"||!d.Len)break;off+=d.Len;}
 o<<"]}";reply=o.str();return reply.c_str();
}
int rr_debug_begin(int mode,uint32_t target,int next){auto*c=rrCPU();if(!c||mode<0||mode>2||rrPeek(target)<0)return 0;rrDebugActive=true;rrDebugMode=mode;rrDebugTarget=target;rrDebugSkip=next&&rrPC()==target;rrDebugTargetMode=c->Mode;rrDebugTargetCS=-1;rrDebugTargetBase=x86_CPU_segBase(c,1);return 1;}
void rr_debug_bind(int selector,uint32_t base){rrDebugTargetCS=selector;rrDebugTargetBase=base;}
int rr_debug_run(uint32_t steps){try{if(!rrDebugActive||steps>10000||!steps)return -1;if(rrDebugMode==0)return 1;auto*c=rrCPU();auto end=c->Steps+steps;
 while(c->Steps<end){if(rrDebugMode==2&&rrPC()==rrDebugTarget&&!rrDebugSkip){if(c->Mode!=rrDebugTargetMode||(rrDebugTargetCS>=0&&(c->Seg[1]!=rrDebugTargetCS||x86_CPU_segBase(c,1)!=rrDebugTargetBase)))return 5;return 4;}auto result=rrStep(true);if(result==3&&rrDebugMode==2)continue;if(result!=2)return result;if(rrDebugMode==1)return 2;}return 0;
 }catch(const std::exception&e){errorText=e.what();return -1;}}
}
