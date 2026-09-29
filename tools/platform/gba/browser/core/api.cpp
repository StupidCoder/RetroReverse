#include "host.h"
static gbamachine_Machine*machine=nullptr;
static Slice<uint8_t>romInput;
static std::string errorText,reply;
static std::vector<uint8_t>pixels;
extern "C" {
uint8_t*rr_input(uint32_t n){try{if(n<192||n>32*1024*1024)throw std::runtime_error("GBA images must be 192 bytes to 32 MiB");romInput=Slice<uint8_t>::make(n);return romInput.p;}catch(const std::exception&e){errorText=e.what();return nullptr;}}
int rr_init(uint32_t n){try{if(machine||n!=romInput.n)throw std::runtime_error("Invalid GBA initialization");gba_ROM rom;rom.Data=romInput;machine=gbamachine_New(&rom);return 1;}catch(const std::exception&e){errorText=e.what();return 0;}}
int rr_run(uint32_t n){try{if(!machine||!n||n>1000000)throw std::runtime_error("Invalid execution slice");auto before=machine->Steps;gbamachine_Machine_RunFrames(machine,1,n);if(machine->cpu->Halted)throw std::runtime_error(machine->cpu->HaltReason);return machine->Steps-before;}catch(const std::exception&e){errorText=e.what();return -1;}}
void rr_pad(uint32_t b){if(machine)gbamachine_Machine_SetKeys(machine,b);}
const char*rr_error(){return errorText.c_str();}
const char*rr_status(){if(!machine)return "{}";std::ostringstream s;s<<"{\"steps\":"<<machine->Steps<<",\"frames\":"<<machine->vid.frames<<",\"seconds\":"<<double(machine->vid.frames)*280896/16777216<<",\"pc\":"<<machine->cpu->R[15]<<",\"width\":240,\"height\":160}";reply=s.str();return reply.c_str();}
const char*rr_proof(){auto m=machine;auto f=rrFrame(m);std::ostringstream s;s<<"{\"steps\":"<<m->Steps<<",\"instrs\":"<<m->cpu->Instrs<<",\"frames\":"<<m->vid.frames<<",\"pc\":"<<m->cpu->R[15]<<",\"ram\":"<<rrHash(m->ewram.p,m->ewram.n)<<",\"iwram\":"<<rrHash(m->iwram.p,m->iwram.n)<<",\"vram\":"<<rrHash(m->vram.p,m->vram.n)<<",\"rgba\":"<<rrHash(f.data(),f.size())<<"}";reply=s.str();return reply.c_str();}
const char*rr_profile(){return rrprof::json();}
uint8_t*rr_frame(){rrprof::Scope t(7,"Display conversion");pixels=rrFrame(machine);return pixels.data();}
}
#include "state-fields.h"
static std::vector<std::shared_ptr<void>>stateOwners;
static void rrBind(gbamachine_Machine*m){auto*b=arenaNew(gbamachine_bus{m});m->cpu->bus=m->cpu->wide=b;m->cpu->SWI=gbamachine_biosSWI(m);}
static void stateWrite(rrstate::Archive&a){a.header(13,1);a(machine);}
static void stateRead(rrstate::Archive&a){a.header(13,1);gbamachine_Machine*next=nullptr;a(next);a.finish();if(!next||!next->cpu||!next->apu||next->ewram.n!=262144||next->iwram.n!=32768||next->vram.n!=98304||next->pal.n!=1024||next->oam.n!=1024||next->vid.line<0||next->vid.line>=228)throw std::runtime_error("Invalid GBA state");next->rom=romInput;rrBind(next);machine=next;stateOwners=std::move(a.owned);}
#include "../../../../browser/state/api.inc"

#include "capture.h"
