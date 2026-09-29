#include "host.h"
static gc_Machine*machine=nullptr;
static std::string errorText,reply;
static std::vector<uint8_t>pixels;
extern "C" {
int rr_init_file(double n){try{if(machine)throw std::runtime_error("Reset requires a new worker");machine=rrBoot(uint64_t(n));machine->OnDisplay=[](gc_Machine*m){m->StopRequested=true;};return 1;}catch(const std::exception&e){errorText=e.what();return 0;}}
int rr_run(uint32_t n){try{if(!machine||!n||n>1000000)throw std::runtime_error("Invalid execution slice");auto r=gc_Machine_Run(machine,n);if(r.Reason!="step budget exhausted"&&r.Reason!="stop requested")throw std::runtime_error(r.Reason);return r.Steps;}catch(const std::exception&e){errorText=e.what();return -1;}}
void rr_pad(uint32_t b,int x,int y){if(machine){gc_Machine_SetPadButtons(machine,0,b);gc_Machine_SetPadStick(machine,0,128+std::clamp(x,-80,80)*96/80,128+std::clamp(y,-80,80)*96/80);}}
const char*rr_error(){return errorText.c_str();}
const char*rr_status(){if(!machine)return "{}";std::ostringstream s;s<<"{\"steps\":"<<machine->Instrs<<",\"frames\":"<<machine->vi.Field<<",\"seconds\":"<<double(machine->vi.Field)/60<<",\"pc\":"<<machine->CPU->PC<<",\"width\":640,\"height\":480}";reply=s.str();return reply.c_str();}
const char*rr_proof(){auto m=machine;auto f=rrFrame(m);std::ostringstream s;s<<"{\"steps\":"<<m->Instrs<<",\"frames\":"<<m->vi.Field<<",\"pc\":"<<m->CPU->PC<<",\"ram\":"<<rrHash(m->RAM.p,m->RAM.n)<<",\"rgba\":"<<rrHash(f.data(),f.size())<<"}";reply=s.str();return reply.c_str();}
const char*rr_profile(){return rrprof::json();}
uint8_t*rr_frame(){rrprof::Scope t(7,"Display conversion");pixels=rrFrame(machine);return pixels.data();}
}
#include "state-fields.h"
static std::vector<std::shared_ptr<void>>stateOwners;
static void rrBind(gc_Machine*m){
 m->CPU->bus=m->CPU->fetcher=m;m->CPU->SC=[m](gekko_CPU*c){return gc_Machine_handleSyscall(m,c);};
 if(m->dsp.Core)m->dsp.Core->bus={m};
 m->OnDisplay=[](gc_Machine*m){m->StopRequested=true;};m->SingleThreaded=true;
}
static void stateWrite(rrstate::Archive&a){a.header(11,2);a(machine);a.raw(machine->RAM.p,24*1024*1024);a.raw(machine->ARAM.p,16*1024*1024);}
static void stateRead(rrstate::Archive&a){
 a.header(11,2);gc_Machine*next=nullptr;a(next);
 if(!next||!next->CPU||next->gpu.EFB.n>640*528||next->gpu.ZBuf.n>640*528||next->gpu.Tlut.n>1024*1024)throw std::runtime_error("Invalid GameCube state");
 a.charge(40*1024*1024);next->RAM=Slice<uint8_t>::make(24*1024*1024);next->ARAM=Slice<uint8_t>::make(16*1024*1024);
 a.raw(next->RAM.p,next->RAM.n);a.raw(next->ARAM.p,next->ARAM.n);a.finish();
 next->disc=machine->disc;next->discMD5=machine->discMD5;rrBind(next);machine=next;stateOwners=std::move(a.owned);
}
#include "../../../../browser/state/api.inc"

#include "capture.h"
