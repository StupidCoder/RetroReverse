#include "host.h"
static dc_Machine*machine=nullptr;
static std::string errorText,reply;
static std::vector<uint8_t>pixels;
extern "C" {
int rr_disc_track(uint32_t number,int32_t lba,double offset,double length,int data){if(machine||discTracks.n>=99||number<1||number>99||offset<0||length<=0)return 0;discTracks=append(discTracks,dc_Track{int64_t(number),data?"MODE1_RAW":"AUDIO",int64_t(offset),int64_t(length),lba});return 1;}
int rr_init_file(double n){try{if(machine)throw std::runtime_error("Reset requires a new worker");machine=rrBoot(uint64_t(n));machine->OnDisplay=[](uint64_t){machine->StopRequested=true;};return 1;}catch(const std::exception&e){errorText=e.what();return 0;}}
int rr_run(uint32_t n){try{if(!machine||!n||n>1000000)throw std::runtime_error("Invalid execution slice");auto before=machine->Instrs;auto r=dc_Machine_Run(machine,n,dc_RunConfig{{},true});if(r.Reason!="steps"&&r.Reason!="stop requested")throw std::runtime_error(r.Reason);return machine->Instrs-before;}catch(const std::exception&e){errorText=e.what();return -1;}}
void rr_pad(uint32_t b,int x,int y){if(machine){machine->Pad.Buttons=uint16_t(~b);machine->Pad.JoyX=128+std::clamp(x,-80,80)*127/80;machine->Pad.JoyY=128+std::clamp(y,-80,80)*127/80;machine->Pad.LT=(b&(1<<16))?255:0;machine->Pad.RT=(b&(1<<17))?255:0;}}
const char*rr_error(){return errorText.c_str();}
const char*rr_status(){if(!machine)return "{}";auto[w,h]=rrDimensions(machine);std::ostringstream s;s<<"{\"steps\":"<<machine->Instrs<<",\"frames\":"<<machine->Fields<<",\"seconds\":"<<double(machine->Fields)/60<<",\"pc\":"<<machine->CPU->PC<<",\"width\":"<<w<<",\"height\":"<<h<<"}";reply=s.str();return reply.c_str();}
const char*rr_proof(){auto m=machine;auto f=rrFrame(m);std::ostringstream s;s<<"{\"steps\":"<<m->Instrs<<",\"frames\":"<<m->Fields<<",\"pc\":"<<m->CPU->PC<<",\"ram\":"<<rrHash(m->RAM.p,m->RAM.n)<<",\"vram\":"<<rrHash(m->VRAM.p,m->VRAM.n)<<",\"aica\":"<<rrHash(m->AICARAM.p,m->AICARAM.n)<<",\"rgba\":"<<rrHash(f.data(),f.size())<<"}";reply=s.str();return reply.c_str();}
const char*rr_profile(){return rrprof::json();}
uint8_t*rr_frame(){rrprof::Scope t(7,"Display conversion");pixels=rrFrame(machine);return pixels.data();}
}
#include "state-fields.h"
static std::vector<std::shared_ptr<void>>stateOwners;
static void rrBind(dc_Machine*m){m->CPU->bus=m->CPU->fetcher=m;m->ARM->bus=m->ARM->wide=arenaNew(dc_armBus{m});if(m->bios)m->bios->m=m;m->OnDisplay=[](uint64_t){machine->StopRequested=true;};}
static void stateWrite(rrstate::Archive&a){a.header(14,1);a(machine);}
static void stateRead(rrstate::Archive&a){a.header(14,1);dc_Machine*next=nullptr;a(next);a.finish();if(!next||!next->CPU||!next->ARM||next->RAM.n!=16777216||next->VRAM.n!=8388608||next->AICARAM.n!=2097152||next->Flash.n!=262144||next->TAFrame.n>32*1024*1024||next->TAClosed.n>32*1024*1024)throw std::runtime_error("Invalid Dreamcast state");next->Disc=machine->Disc;rrBind(next);machine=next;stateOwners=std::move(a.owned);}
#include "../../../../browser/state/api.inc"

#include "capture.h"
