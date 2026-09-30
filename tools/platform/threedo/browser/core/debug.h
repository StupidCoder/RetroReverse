#pragma once
// ARMv3 32-bit mode only. Counters distinguish scheduler units, instructions and displays.
static int rrDebugMode=0;static uint32_t rrDebugTarget=0;static bool rrDebugSkip=false;
static int32_t rrDebugTask=-1;static std::vector<uint8_t> rrBacking;
static int rrPeek(uint32_t a){if(a<0x200000)return machine->dram[a];if(a<0x300000)return machine->vram[a-0x200000];if(a>=0x400000&&a<0x800000)return machine->imem[a-0x400000];return -1;}
static int32_t rrTask(){return threedo_Machine_curTask(machine)->num;}
static bool rrHLE(uint32_t pc){return pc>=threedo_hleBase&&pc<threedo_hleBase+threedo_hleSize;}
static std::string rrQuote(std::string text){std::string out="\"";for(unsigned char c:text){if(c=='"'||c=='\\')out+='\\';if(c>=32)out+=c;}return out+'"';}
extern "C" {
double rr_cycle(){return double(totalSteps);}
uint32_t rr_ram_size(){return 0x300000;}
uint8_t*rr_ram(){rrBacking.resize(0x300000);memcpy(rrBacking.data(),machine->dram.p,0x200000);memcpy(rrBacking.data()+0x200000,machine->vram.p,0x100000);return rrBacking.data();}
const char*rr_debug_snapshot(int32_t address){auto*c=machine->CPU;uint32_t pc=c->R[15],at=address<0?pc:uint32_t(address);std::ostringstream o;o<<"{\"architecture\":\"arm60\",\"mode\":\"arm32\",\"boundary\":true,\"cycle\":\""<<totalSteps<<"\",\"retiredInstructions\":\""<<c->Instrs<<"\",\"frame\":"<<machine->frame<<",\"field\":"<<machine->vblank<<",\"audioTime\":"<<machine->audioTime<<",\"task\":"<<rrTask()<<",\"hle\":"<<(rrHLE(pc)?"true":"false")<<",\"movieHLE\":"<<(movieActive(machine)?"true":"false")<<",\"bank\":"<<c->Mode<<",\"nextPC\":"<<pc<<",\"address\":"<<at<<",\"registers\":{";
for(int i=0;i<16;i++){if(i)o<<',';o<<"\"R"<<i<<"\":"<<c->R[i];}o<<",\"CPSR\":"<<arm60_CPU_CPSR(c)<<"},\"bytes\":[";for(int i=0;i<256;i++){if(i)o<<',';auto b=rrPeek(at+i);if(b<0)o<<"null";else o<<b;}o<<"],\"mapping\":[";for(int i=0;i<256;i++){if(i)o<<',';o<<(rrPeek(at+i)<0?1:0);}o<<"],\"instructions\":[";
for(int i=0;i<32;i++){uint32_t a=at+i*4;if(i)o<<',';o<<"{\"address\":"<<a<<",\"length\":4,\"bytes\":[";std::array<uint8_t,4>b{};bool ok=true;for(int j=0;j<4;j++){if(j)o<<',';int v=rrPeek(a+j);if(v<0){ok=false;o<<"null";}else{o<<v;b[j]=v;}}Slice<uint8_t> code;code.p=b.data();code.n=code.c=4;auto text=ok?arm60_Decode(code,a).Text:rrHLE(a)?"Portfolio HLE entry (no guest opcode)":"Unavailable storage";o<<"],\"text\":"<<rrQuote(text)<<"}";}o<<"]}";reply=o.str();return reply.c_str();}
int rr_debug_begin(int mode,uint32_t target,int next){if(!machine||mode<0||mode>2||(mode==2&&(target>=0x800000||(target&3)))||movieActive(machine))return 0;rrDebugMode=mode;rrDebugTarget=target;rrDebugSkip=next&&machine->CPU->R[15]==target;rrDebugTask=-1;return 1;}
void rr_debug_bind(int32_t task){rrDebugTask=task;if(task>=0&&task!=rrTask())rrDebugSkip=false;}
int rr_debug_run(uint32_t budget){try{if(!machine||!budget||budget>1000)return 7;if(movieActive(machine))return 7;if(rrDebugMode==0)return 1;auto startTask=rrTask();auto instr=machine->CPU->Instrs;auto steps=runContext.steps;int stopped=0;bool hle=false,swi=false;
rrSchedulerBoundary=[&](threedo_Machine*m,uint32_t pc){if(rrDebugMode==1&&rrTask()!=startTask&&m->CPU->Instrs==instr&&runContext.steps==steps){stopped=9;return true;}if(rrDebugMode==2&&pc==rrDebugTarget&&(rrDebugTask<0||rrTask()==rrDebugTask)){if(rrDebugSkip)rrDebugSkip=false;else{stopped=4;return true;}}if(rrDebugMode==1){hle=rrHLE(pc);swi=rrPeek(pc)>=0&&(rrPeek(pc)&15)==15&&arm60_CPU_cond(m->CPU,rrPeek(pc)>>4); }return false;};
rrSchedulerClockBase=totalSteps-runContext.steps;auto beforeFrame=machine->frame;auto result=threedo_Machine_RunSlice(machine,rrDebugMode==1?1:budget,runContext);rrSchedulerBoundary={};totalSteps+=result.Steps;presentationSeconds+=(machine->frame-beforeFrame)/30.0;if(stopped)return stopped;
if(result.Reason=="stop requested"){runContext={};return rrDebugMode==1?(machine->CPU->Instrs>instr?2:11):0;}if(result.Reason=="movie pending")return 7;if(result.Reason!="step budget reached"){errorText=result.Reason;return 6;}if(rrDebugMode==1)return hle?8:machine->CPU->Instrs>instr?(swi?10:2):9;return 0;
}catch(const std::exception&e){rrSchedulerBoundary={};errorText=e.what();return -1;}}
}
