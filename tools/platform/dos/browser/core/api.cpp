#include "host.h"
static std::string errorText,reply;
static std::vector<uint8_t>pixels;
static Slice<uint8_t>pathInput;
static std::array<bool,128>heldKeys{};
static uint32_t lastPad{};
#include "debug.h"
extern "C" {
uint8_t*rr_input(uint32_t n){if(n>4096)return nullptr;pathInput=Slice<uint8_t>::make(n);return pathInput.p;}
int rr_file(int id,uint32_t n,double size){try{if(rrCPU()||n!=pathInput.n||!n||id<0||!std::isfinite(size)||size<0||size>4ull*1024*1024*1024)throw std::runtime_error("Invalid local file");auto name=rrfiles::path(std::string((char*)pathInput.p,n));if(rrfiles::sources.contains(name))throw std::runtime_error("Duplicate game path");
#ifdef __EMSCRIPTEN__
rrfiles::sources[name]={uint64_t(size),[id](uint64_t at,uint8_t*out,size_t n){rrprof::Scope t(6,"Local game file reads");auto got=rrReadGameFile(id,out,at,n);if(got<0)throw std::runtime_error("Cannot read selected game file");return size_t(got);}};
#else
(void)id;(void)size;
#endif
return 1;}catch(const std::exception&e){errorText=e.what();return 0;}}
int rr_init(uint32_t n,int compatibility){try{if(rrCPU()||n!=pathInput.n)throw std::runtime_error("Invalid DOS executable selection");rrBoot(std::string((char*)pathInput.p,n),compatibility);return 1;}catch(const std::exception&e){errorText=e.what();return 0;}}
int rr_run(uint32_t n){try{auto*c=rrCPU();if(!c||!n||n>1000000)throw std::runtime_error("Invalid execution slice");auto start=c->Steps,end=std::min(start+n,(rrFrames()+1)*rrFramePeriod());rrprof::Scope t(0,"x86 CPU and devices");while(c->Steps<end&&!c->Halted){if(rrDebugHookPending)rrStep(false);else{x86_CPU_Step(c);if(rrdos::collecting)rrdos::observe(c->Steps);}}if(c->Halted)throw std::runtime_error(c->HaltReason);return c->Steps-start;}catch(const std::exception&e){errorText=e.what();return -1;}}
static constexpr int padCodes[]={0x48,0x50,0x4b,0x4d,0x1c,0x39,0x1d,0x01};
static bool padHeld(uint32_t code){for(int i=0;i<8;i++)if(padCodes[i]==code&&(lastPad&(1<<i)))return true;return false;}
static void sendKey(uint32_t code,bool down){uint8_t sc=code|(down?0:0x80);if(protectedMachine)dos_PM_EnqueueScancode(protectedMachine,sc);else realMachine->keyEvents=append(realMachine->keyEvents,dos_injEvent{dos_injKey,sc,0,0,0});}
void rr_key(uint32_t code,int down){if(!rrCPU()||code>=128||heldKeys[code]==bool(down))return;bool was=heldKeys[code]||padHeld(code);heldKeys[code]=down;bool now=heldKeys[code]||padHeld(code);if(was!=now)sendKey(code,now);}
void rr_pad(uint32_t b){if(!rrCPU())return;auto old=lastPad;lastPad=b;for(int i=0;i<8;i++)if(((old^b)&(1<<i))&&!heldKeys[padCodes[i]])sendKey(padCodes[i],bool(b&(1<<i)));}
void rr_mouse(int x,int y,int buttons){if(realMachine){dos_Machine_MoveMouseTo(realMachine,std::clamp(x,0,319),std::clamp(y,0,199));dos_Machine_SetMouseButtons(realMachine,buttons&3);}}
const char*rr_error(){return errorText.c_str();}
const char*rr_status(){auto*c=rrCPU();if(!c)return "{}";std::ostringstream o;o<<"{\"steps\":"<<c->Steps<<",\"frames\":"<<rrFrames()<<",\"seconds\":"<<double(c->Steps)/rrFramePeriod()/70<<",\"pc\":"<<rrPC()<<",\"width\":320,\"height\":200,\"protectedMode\":"<<(protectedMachine?"true":"false")<<"}";reply=o.str();return reply.c_str();}
const char*rr_proof(){auto*c=rrCPU();auto&mem=realMachine?realMachine->Mem:protectedMachine->Mem;auto f=rrFrame();std::ostringstream o;o<<"{\"steps\":"<<c->Steps<<",\"frames\":"<<rrFrames()<<",\"pc\":"<<c->IP<<",\"ram\":"<<rrHash(mem.p,mem.n)<<",\"rgba\":"<<rrHash(f.data(),f.size())<<"}";reply=o.str();return reply.c_str();}
const char*rr_profile(){return rrprof::json();}
uint8_t*rr_frame(){rrprof::Scope t(7,"VGA display conversion");pixels=rrFrame();return pixels.data();}
}
#include "state-fields.h"
inline void stateFields(rrstate::Archive&a,os_File&f){a(f.name,f.offset,f.writable,f.closed);if(a.reading&&(rrfiles::path(f.name)!=f.name||f.offset<0))throw std::runtime_error("Invalid virtual file handle");}
static std::vector<std::shared_ptr<void>>stateOwners;
static void rrBind(dos_Machine*m){auto*c=m->CPU;c->bus=m;c->IntHook=[m](x86_CPU*c,uint8_t n){return dos_Machine_handleInt(m,c,n);};c->OnStep=[m](x86_CPU*c){dos_Machine_onStep(m,c);};c->PortIn=[m](uint16_t p,int64_t n){return dos_Machine_portIn(m,p,n);};c->PortOut=[m](uint16_t p,int64_t n,uint32_t v){dos_Machine_portOut(m,p,n,v);};}
static void rrBind(dos_PM*m){auto*c=m->CPU;c->bus=m;c->IntHook=[m](x86_CPU*c,uint8_t n){return dos_PM_handleInt(m,c,n);};c->OnStep=[m](x86_CPU*c){dos_PM_PumpInput(m,c);};c->SegResolve=[m](uint16_t s){return dos_PM_resolveSel(m,s);};c->PortIn=[m](uint16_t p,int64_t n){return dos_PM_portIn(m,p,n);};c->PortOut=[m](uint16_t p,int64_t n,uint32_t v){dos_PM_portOut(m,p,n,v);};}
static void stateWrite(rrstate::Archive&a){a.header(15,2);a(realMachine,protectedMachine);auto&mem=realMachine?realMachine->Mem:protectedMachine->Mem;a.raw(mem.p,mem.n);a(heldKeys,lastPad);auto n=a.count(rrfiles::overlay.size());for(auto&[key,value]:rrfiles::overlay){auto k=key;a(k,value);}for(auto set:{rrfiles::deleted,rrfiles::dirs}){auto n=a.count(set.size());for(auto k:set)a(k);}a(rrDebugHookPending);}
static void stateRead(rrstate::Archive&a){uint32_t version=a.bytes.size()>=12?uint32_t(a.bytes[8]):0;if(version!=1&&version!=2)throw std::runtime_error("Unsupported DOS state version");a.header(15,version);dos_Machine*r=nullptr;dos_PM*p=nullptr;a(r,p);if(bool(r)==bool(p)||bool(r)!=bool(realMachine))throw std::runtime_error("State machine mode mismatch");if(r&&(!r->CPU||!r->io||!r->vga))throw std::runtime_error("Invalid DOS state");if(p&&!p->CPU)throw std::runtime_error("Invalid DOS state");auto&mem=r?r->Mem:p->Mem;auto n=r?1048576:65*1048576;a.charge(n);mem=Slice<uint8_t>::make(n);a.raw(mem.p,n);std::array<bool,128>keys{};uint32_t pad{};a(keys,pad);std::map<std::string,Slice<uint8_t>>overlay;auto count=a.count(0);for(uint32_t i=0;i<count;i++){std::string k;Slice<uint8_t>v;a(k,v);if(rrfiles::path(k)!=k||overlay.contains(k))throw std::runtime_error("Invalid virtual save path");overlay[k]=v;}std::set<std::string>deleted,dirs;for(auto*set:{&deleted,&dirs}){auto count=a.count(0);for(uint32_t i=0;i<count;i++){std::string k;a(k);if(rrfiles::path(k)!=k)throw std::runtime_error("Invalid virtual path");set->insert(k);}}bool pending=false;if(version==2)a(pending);a.finish();if(r)rrBind(r);else rrBind(p);realMachine=r;protectedMachine=p;rrDOSReal=r;rrDOSProtected=p;rrfiles::overlay=std::move(overlay);rrfiles::deleted=std::move(deleted);rrfiles::dirs=std::move(dirs);heldKeys=keys;lastPad=pad;rrDebugHookPending=pending;rrDebugActive=false;stateOwners=std::move(a.owned);}
#include "../../../../browser/state/api.inc"
#include "capture.h"

#include "memory.h"
