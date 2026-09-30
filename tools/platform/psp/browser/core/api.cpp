#include "host.h"
static psp_Machine*machine=nullptr;
static BlockSource disc;
static std::string errorText,reply;
static std::vector<uint8_t>pixels;
extern "C"{
int rr_init_file(double n){try{if(machine)throw std::runtime_error("Reset requires a new worker");disc.mount(n);machine=rrBoot(&disc);errorText.clear();return 1;}catch(const std::exception&e){errorText=e.what();return 0;}}
int rr_run(uint32_t n){try{if(!machine||!n||n>1000000)throw std::runtime_error("Invalid PSP execution slice");auto r=psp_Machine_Run(machine,n);if(r.Reason!="budget reached")throw std::runtime_error(r.Reason);return r.Steps;}catch(const std::exception&e){errorText=e.what();return -1;}}
void rr_pad(uint32_t buttons,int x,int y){if(machine){machine->pad=buttons;rrAnalogX=128+std::clamp(x,-80,80)*127/80;rrAnalogY=128-std::clamp(y,-80,80)*127/80;}}
const char*rr_error(){return errorText.c_str();}
const char*rr_status(){if(!machine)return "{}";std::ostringstream s;s<<"{\"steps\":"<<rrSteps<<",\"frames\":"<<rrFrames<<",\"inputSeconds\":"<<double(machine->vblanks)/60<<",\"seconds\":"<<double(machine->vblanks)/60<<",\"pc\":"<<machine->CPU->PC<<",\"width\":480,\"height\":272,\"discReads\":"<<disc.reads<<",\"discBytes\":"<<disc.readBytes<<"}";reply=s.str();return reply.c_str();}
const char*rr_proof(){auto m=machine;auto fb=rrFrame(m);std::ostringstream s;s<<"{\"steps\":"<<rrSteps<<",\"frames\":"<<rrFrames<<",\"pc\":"<<m->CPU->PC<<",\"ram\":"<<rrHash(m->ram.p,m->ram.n)<<",\"vram\":"<<rrHash(m->vram.p,m->vram.n)<<",\"rgba\":"<<rrHash(fb.data(),fb.size())<<"}";reply=s.str();return reply.c_str();}
const char*rr_profile(){return rrprof::json();}
uint8_t*rr_frame(){rrprof::Scope t(5,"Display conversion");pixels=rrFrame(machine);return pixels.data();}
}
#include "state.h"
static std::vector<std::shared_ptr<void>>stateOwners;
static void stateWrite(rrstate::Archive&a){a.header(7,1);a(machine,rrSteps,rrFrames,rrUntilVBlank,rrAnalogX,rrAnalogY);a.raw(machine->ram.p,33554432);}
static void stateRead(rrstate::Archive&a){a.header(7,1);psp_Machine*next=nullptr;uint64_t steps,frames,clock;uint8_t x,y;a(next,steps,frames,clock,x,y);if(!next||!clock||clock>1000000)throw std::runtime_error("Invalid PSP clock state");a.charge(33554432);next->ram=Slice<uint8_t>::make(33554432);a.raw(next->ram.p,33554432);a.finish();rebindState(next,machine->vol);machine=next;rrSteps=steps;rrFrames=frames;rrUntilVBlank=clock;rrAnalogX=x;rrAnalogY=y;stateOwners=std::move(a.owned);}
#include "../../../../browser/state/api.inc"
#include "capture.h"

#include "memory.h"
