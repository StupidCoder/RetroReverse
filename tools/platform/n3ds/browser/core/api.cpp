#include "host.h"
static n3ds_Machine*machine=nullptr;
static Slice<uint8_t>upload;
static std::string errorText,reply;
static std::vector<uint8_t>pixels;
extern "C"{
uint8_t*rr_input(uint32_t n){try{if(n<512||n>1024u*1024*1024)throw std::runtime_error("Invalid 3DS cartridge size (maximum 1 GiB)");upload.owner=std::shared_ptr<uint8_t[]>(new uint8_t[n]);upload.p=upload.owner.get();upload.n=upload.c=n;return upload.p;}catch(const std::exception&e){errorText=e.what();return nullptr;}}
int rr_init(uint32_t n){try{if(n!=upload.n)throw std::runtime_error("Cartridge upload mismatch");auto[m,e]=n3ds_NewMachine(upload);if(e)throw std::runtime_error(e.text);machine=m;rrgpu::enabled=false;rrgpu::operations={};rrgpu::accelerated={};rrgpu::committedBytes=0;machine->SingleThreaded=true;errorText.clear();return 1;}catch(const std::exception&e){errorText=e.what();return 0;}}
int rr_run(uint32_t n){try{if(!machine||n>1000000)throw std::runtime_error("Invalid 3DS execution slice");rrprof::Scope t(0,"ARM11 / Horizon scheduler");auto ran=n3ds_Machine_RunFrames(machine,1,n);if(machine->CPU->Halted)throw std::runtime_error(machine->CPU->HaltReason);rrCollectTextures(machine->gpu);return ran;}catch(const std::exception&e){errorText=e.what();return -1;}}
void rr_pad(uint32_t buttons,int x,int y){if(machine){machine->hidButtons=buttons;if(!x)x=((buttons&16)?80:0)-((buttons&32)?80:0);if(!y)y=((buttons&64)?80:0)-((buttons&128)?80:0);int16_t cx=std::clamp(x,-80,80)*156/80,cy=std::clamp(y,-80,80)*156/80;rr3dsCirclePad=uint16_t(cx)|(uint32_t(uint16_t(cy))<<16);}}
void rr_touch(int x,int y,int down){if(machine)n3ds_Machine_SetTouch(machine,std::clamp(x,0,319),std::clamp(y,0,239),down!=0);}
const char*rr_error(){return errorText.c_str();}
const char*rr_status(){if(!machine)return "{}";std::ostringstream s;s<<"{\"steps\":"<<machine->instrs<<",\"frames\":"<<machine->vblankCount<<",\"inputSeconds\":"<<double(machine->instrs)/(double(n3ds_stepsPerFrame)*60)<<",\"pc\":"<<machine->CPU->R[15]<<",\"width\":400,\"height\":480,\"draws\":"<<machine->gpu->Draws<<"}";reply=s.str();return reply.c_str();}
const char*rr_proof(){reply=proof(machine);return reply.c_str();}
const char*rr_profile(){return rrprof::json();}
int rr_perf_enable(uint32_t stride){return rrperf::enable(stride);}
const char*rr_perf_stats(){reply=rrperf::json();return reply.c_str();}
uint8_t*rr_frame(){pixels=frame(machine);return pixels.data();}
}
#include "state.h"
static std::vector<std::shared_ptr<void>>stateOwners;
static void stateWrite(rrstate::Archive&a){a.header(6,1);a(machine,rr3dsCirclePad);}
static void stateRead(rrstate::Archive&a){a.header(6,1);n3ds_Machine*next=nullptr;uint32_t circle=0;a(next,circle);a.finish();rebindState(next,machine->romfs,machine->romfsRaw);machine=next;rr3dsCirclePad=circle;stateOwners=std::move(a.owned);}
#include "../../../../browser/state/api.inc"
#include "replay.h"
extern "C"{
int rr_capture_begin(){try{auto b=rr3ds::memory(machine,true);rrcapture::trace.begin(b.data(),b.size());machine->WatchLo=0;machine->WatchHi=UINT32_MAX;
 machine->OnWrite=[](uint32_t a,uint32_t v,uint32_t pc){rr3ds::write(machine,a,v,pc);};
 machine->OnPICACmd=[](n3ds_PICAWrite w){rr3ds::clean();std::ostringstream s;s<<"{\"kind\":\"PICA register write\",\"register\":"<<w.Reg<<",\"mask\":"<<unsigned(w.Mask)<<",\"value\":"<<w.Value<<",\"offset\":"<<w.Off<<"}";rrcapture::trace.event(machine->instrs,machine->CPU->R[15],s.str());};
 machine->OnPixel=[](uint32_t x,uint32_t y,n3ds_PixelEvent ev){if(ev.Drawn)return;auto g=machine->gpu;auto fb=n3ds_GPU_fbstate(g);auto a=rr3ds::address(fb.colorAddr+n3ds_tiledOffset(x,y,fb.width));uint32_t value=uint32_t(ev.A)|uint32_t(ev.B)<<8|uint32_t(ev.G)<<16|uint32_t(ev.R)<<24;rrcapture::trace.record(a,value,4,machine->instrs,machine->CPU->R[15],rrcapture::trace.current,(ev.ZReject?2:0)|(ev.AlphaReject?4:0)|(ev.StencilReject?16:0));};return 1;}catch(const std::exception&e){errorText=e.what();return 0;}}
int rr_capture_end(){try{machine->OnWrite={};machine->OnPICACmd={};machine->OnPixel={};rr3ds::screenGeometry(machine);auto b=rr3ds::memory(machine);rrcapture::trace.end(b.data(),b.size());return 1;}catch(const std::exception&e){errorText=e.what();return 0;}}
const char*rr_capture_info(){reply=rrcapture::trace.info();return reply.c_str();}
const char*rr_pixel(int x,int y){auto a=rr3ds::pixelAddress(x,y);if(a==UINT32_MAX)return "{\"error\":\"Outside captured screen or screen not presented\"}";reply=rrcapture::trace.pixel(a,rr3ds::screens[y/240].bpp);reply.pop_back();reply+=",\"displayFormat\":"+std::to_string(rr3ds::screens[y/240].format)+"}";return reply.c_str();}
const char*rr_source(uint32_t a,int size,uint32_t before,uint32_t expected){reply=rrcapture::trace.pixel(a,size,before,expected);return reply.c_str();}
const char*rr_resource(uint32_t id,uint32_t offset){reply=rrcapture::trace.resourceJSON(id,offset);return reply.c_str();}
const char*rr_replay_begin(){rrreplay::replay.begin();rr3ds::prepareReplay();reply=rr3ds::replayInfo();return reply.c_str();}
int rr_replay_seek(uint32_t step){return rrreplay::replay.seek(step);}
const char*rr_replay_info(){reply=rr3ds::replayInfo();return reply.c_str();}
uint32_t rr_replay_for_write(uint32_t id){return rrreplay::replay.forWrite(id);}
uint8_t*rr_replay_frame(){pixels=rr3ds::replayDisplay();return pixels.data();}
}

#include "memory.h"

extern "C"{
const char*rr_graphics_scanout(){
 if(machine->OnRead||machine->OnWrite)return "{\"complete\":false}";
 std::ostringstream s;s<<"{\"complete\":true,\"screens\":[";
 for(int i=0;i<2;i++){
  if(i)s<<',';auto r=i?machine->lastXferBottom:machine->lastXferTop;auto fb=n3ds_Machine_Scanout(machine,i);
  if(fb.Valid&&fb.AddrLeft){r.dst=fb.AddrLeft;auto bpp=n3ds_fbBPP(fb.Format);if(bpp){r.bpp=bpp;r.format=fb.Format&7;if(fb.Stride)r.stride=fb.Stride/bpp;}if(!r.w||!r.h){r.w=240;r.h=i?320:400;}}
  if(!r.dst){s<<"null";continue;}
  uint64_t n=(uint64_t(r.h?r.h-1:0)*r.stride+r.w)*r.bpp;auto p=rrgpu::range(machine,r.dst,n);
  if(!p||!n||r.format>4||!r.bpp||r.w>1024||r.h>1024)return "{\"complete\":false}";
  s<<"{\"pointer\":"<<uintptr_t(p)<<",\"bytes\":"<<n<<",\"width\":"<<r.w<<",\"height\":"<<r.h<<",\"stride\":"<<r.stride<<",\"format\":"<<r.format<<",\"bpp\":"<<r.bpp<<"}";
 }s<<"]}";reply=s.str();return reply.c_str();
}
void rr_graphics_enable(int value){rrgpu::enabled=value!=0;}
const char*rr_graphics_stats(){std::ostringstream s;s<<"{\"operations\":[";for(int i=1;i<=4;i++){if(i>1)s<<',';s<<rrgpu::operations[i];}s<<"],\"accelerated\":[";for(int i=1;i<=4;i++){if(i>1)s<<',';s<<rrgpu::accelerated[i];}s<<"],\"bytes\":"<<rrgpu::committedBytes<<"}";reply=s.str();return reply.c_str();}
void rr_graphics_begin(){rrgpu::begin();}
uint32_t rr_graphics_end(){rrgpu::recording=false;return rrgpu::stream.size();}
uint8_t*rr_graphics_data(){return rrgpu::stream.data();}
const char*rr_graphics_info(){reply="{\"schema\":1,\"dropped\":"+std::to_string(rrgpu::dropped)+",\"referenceDraws\":"+std::to_string(rrgpu::draws)+"}";return reply.c_str();}
}
