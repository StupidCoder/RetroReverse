#include "host.h"
static ps2_Machine*machine=nullptr;
static std::string errorText,reply;
static Slice<uint8_t>pixels,input;
static int64_t width=640,height=448;
extern "C" {
uint8_t*rr_input(uint32_t n){input=Slice<uint8_t>::make(n);return input.p;}
int rr_bios(uint32_t n){if(!machine||n!=uint32_t(input.n)||(n!=4*1024*1024&&n!=8*1024*1024)){errorText="PS2 BIOS must be 4 or 8 MiB";return 0;}machine->bios=input;return 1;}
int rr_init_file(double n){try{if(machine)throw std::runtime_error("Reset requires a new worker");machine=rrBoot(uint64_t(n));machine->OnVBlank=[](ps2_Machine*m){m->StopRequested=true;};return 1;}catch(const std::exception&e){errorText=e.what();return 0;}}
int rr_boot(){try{auto e=ps2_Machine_RebootIOP(machine);if(e)throw std::runtime_error(e.text);return 1;}catch(const std::exception&e){errorText=e.what();return 0;}}
int rr_run(uint32_t n){try{if(!machine||!n||n>1000000)throw std::runtime_error("Invalid execution slice");auto r=ps2_Machine_Run(machine,n);if(r.Reason!="step budget exhausted"&&r.Reason!="stop requested")throw std::runtime_error(r.Reason);return r.Steps;}catch(const std::exception&e){errorText=e.what();return -1;}}
void rr_pad(uint32_t b,int x,int y){if(machine){ps2_Machine_SetPadButtons(machine,b);ps2_Machine_SetPadStick(machine,128+std::clamp(x,-80,80)*127/80,128-std::clamp(y,-80,80)*127/80,128,128);}}
const char*rr_error(){return errorText.c_str();}
const char*rr_status(){if(!machine)return "{}";std::ostringstream s;s<<"{\"steps\":"<<machine->steps<<",\"frames\":"<<machine->vblanks<<",\"seconds\":"<<double(machine->vblanks)/60<<",\"pc\":"<<machine->CPU->PC<<",\"width\":"<<width<<",\"height\":"<<height<<"}";reply=s.str();return reply.c_str();}
uint8_t*rr_frame(){rrprof::Scope timing(7,"GS scanout");auto[b,w,h]=ps2_Machine_GSFrame(machine);if(w>0&&h>0&&w<=2048&&h<=2048){pixels=b;width=w;height=h;}else{pixels=Slice<uint8_t>::make(640*448*4);width=640;height=448;}return pixels.p;}
const char*rr_proof(){auto m=machine;rr_frame();std::ostringstream s;s<<"{\"steps\":"<<m->steps<<",\"frames\":"<<m->vblanks<<",\"pc\":"<<m->CPU->PC<<",\"ram\":"<<rrHash(m->ram.p,m->ram.n)<<",\"vram\":"<<(m->gs?rrHash(m->gs->vram.p,m->gs->vram.n):0)<<",\"rgba\":"<<rrHash(pixels.p,pixels.n)<<"}";reply=s.str();return reply.c_str();}
const char*rr_profile(){return rrprof::json();}
}
#include "state-fields.h"
static std::vector<std::shared_ptr<void>>stateOwners;
static void rrBind(ps2_Machine*m){
 m->CPU->bus=m;m->CPU->fetch=[m](uint32_t a){return ps2_Machine_Fetch32(m,a);};m->CPU->Syscall=[m](r5900_CPU*c){return ps2_Machine_handleSyscall(m,c);};
 m->OnVBlank=[](ps2_Machine*m){m->StopRequested=true;};m->SingleThreaded=true;
 if(m->gs)m->gs->m=m;
 for(int i=0;i<2;i++)if(auto v=m->vifs[i]){v->m=m;if(!v->vu||v->micro.n!=(i?16384:4096)||v->data.n!=v->micro.n)throw std::runtime_error("Invalid VIF state");v->vu->Micro=v->micro;v->vu->Data=v->data;if(i==1)v->vu->XGKick=[v](uint32_t a){ps2_vif_xgkick(v,a);};else m->CPU->COP2=v->vu;}
 if(auto p=m->IOP){if(!p->CPU||p->spr.n!=1024)throw std::runtime_error("Invalid IOP state");p->ps2=m;p->ram=m->iopRAM;p->CPU->bus=p;p->CPU->Syscall=[p](mips_CPU*c){return ps2_IOP_handleSyscall(p,c);};
  for(auto&[key,index]:*p->bound.p){auto l=get(ps2_goLibraries,key.library);if(!l||index>=uint64_t(p->calls.n)||!p->calls[index])throw std::runtime_error("Invalid IOP call binding");p->calls[index]->fn=get(l->funcs,key.id).fn;}
 }
}
static void stateWrite(rrstate::Archive&a){a.header(12,2);a(machine);a.raw(machine->ram.p,32*1024*1024);}
static void stateRead(rrstate::Archive&a){
 a.header(12,2);ps2_Machine*next=nullptr;a(next);
 if(!next||!next->CPU||next->iopRAM.n!=2*1024*1024||next->spram.n!=16384||next->rrVblAcc>=1000000||next->rrIopAcc>=8||(next->gs&&next->gs->vram.n!=4*1024*1024))throw std::runtime_error("Invalid PS2 state");
 a.charge(32*1024*1024);next->ram=Slice<uint8_t>::make(32*1024*1024);a.raw(next->ram.p,next->ram.n);a.finish();
 next->vol=machine->vol;next->exe=machine->exe;rrBind(next);machine=next;stateOwners=std::move(a.owned);
}
#include "../../../../browser/state/api.inc"

#include "capture.h"

#include "memory.h"
