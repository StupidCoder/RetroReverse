#include "host.h"
static xbox_Machine*machine=nullptr;
static std::string errorText,reply;
static std::vector<uint8_t>pixels;
static uint64_t frames=0;
static void rrFlip(xbox_Machine*m){++frames;m->StopRequested=true;}
extern "C" {
int rr_init_file(double n){try{if(machine||!std::isfinite(n)||n<32768||n>16ull*1024*1024*1024)throw std::runtime_error("Invalid Xbox disc initialization");machine=rrBoot(uint64_t(n));machine->OnFlip=rrFlip;return 1;}catch(const std::exception&e){errorText=e.what();return 0;}}
int rr_run(uint32_t n){try{if(!machine||!n||n>1000000)throw std::runtime_error("Invalid execution slice");auto was=frames;auto[r,done]=xbox_Machine_Run(machine,n);if(frames!=was)rrCollectTextures(machine->pgraph);if(machine->CPU->Halted)throw std::runtime_error(machine->CPU->HaltReason);return done;}catch(const std::exception&e){errorText=e.what();return -1;}}
void rr_pad(uint32_t b,int x,int y){if(!machine)return;xbox_Machine_SetPadButtons(machine,0,b&255);for(int i=0;i<8;i++)xbox_Machine_SetPadAnalog(machine,0,i,b&(1u<<(8+i))?255:0);xbox_Machine_SetPadAxis(machine,0,0,std::clamp(x,-80,80)*32767/80);xbox_Machine_SetPadAxis(machine,0,1,std::clamp(y,-80,80)*32767/80);}
const char*rr_error(){return errorText.c_str();}
const char*rr_status(){if(!machine)return "{}";auto s=rrSurface(machine);std::ostringstream o;o<<"{\"steps\":"<<machine->CPU->Steps<<",\"frames\":"<<frames<<",\"seconds\":"<<double(xbox_Machine_guestMs(machine))/1000<<",\"pc\":"<<machine->CPU->IP<<",\"width\":"<<s.w<<",\"height\":"<<s.h<<"}";reply=o.str();return reply.c_str();}
const char*rr_proof(){auto f=rrFrame(machine);std::ostringstream o;o<<"{\"steps\":"<<machine->CPU->Steps<<",\"frames\":"<<frames<<",\"pc\":"<<machine->CPU->IP<<",\"tick\":"<<machine->tick<<",\"draws\":"<<machine->pgraph->Draws<<",\"ram\":"<<rrHash(machine->RAM.p,machine->RAM.n)<<",\"rgba\":"<<rrHash(f.data(),f.size())<<"}";reply=o.str();return reply.c_str();}
const char*rr_profile(){return rrprof::json();}
uint8_t*rr_frame(){rrprof::Scope t(7,"Display conversion");pixels=rrFrame(machine);return pixels.data();}
}
#include "state-fields.h"
static std::vector<std::shared_ptr<void>>stateOwners;
static void rrBindCPU(xbox_Machine*m,x86_CPU*c){c->bus=m;c->SegResolve=[m](uint16_t v){return xbox_Machine_resolveSel(m,v);};c->OnStep=[m](x86_CPU*c){xbox_Machine_onStep(m,c);};c->TSCFunc=[m](){return xbox_Machine_guestTSC(m);};c->PortIn=[m](uint16_t p,int64_t n){return xbox_Machine_portIn(m,p,n);};c->PortOut=[m](uint16_t p,int64_t n,uint32_t v){xbox_Machine_portOut(m,p,n,v);};}
static void rrBind(xbox_Machine*m){rrBindCPU(m,m->CPU);rrBindCPU(m,&m->isrSaved);for(auto*t:m->threads)rrBindCPU(m,&t->ctx);m->pgraph->m=m;m->OnFlip=rrFlip;}
static void stateWrite(rrstate::Archive&a){a.header(16,1);a(machine);a.raw(machine->RAM.p,64*1024*1024);a(frames);}
static void stateRead(rrstate::Archive&a){a.header(16,1);xbox_Machine*next=nullptr;a(next);if(!next||!next->CPU||!next->pgraph||next->threads.n>4096||next->push.running)throw std::runtime_error("Invalid Xbox state");a.charge(64*1024*1024);next->RAM=Slice<uint8_t>::make(64*1024*1024);a.raw(next->RAM.p,next->RAM.n);uint64_t nextFrames{};a(nextFrames);a.finish();next->XBE=machine->XBE;next->Disc=machine->Disc;rrBind(next);machine=next;frames=nextFrames;rrTextureEntries.clear();rrTextures.clear();stateOwners=std::move(a.owned);}
#include "../../../../browser/state/api.inc"
#include "capture.h"
