#include "host.h"
static dsmachine_Machine* machine=nullptr;
static Slice<uint8_t> upload,cartridge;
static std::string errorText,reply;
static std::vector<uint8_t> pixels;
extern "C" {
uint8_t*rr_input(uint32_t n){try{if(n<512||n>512*1024*1024)throw std::runtime_error("Invalid DS cartridge size");upload.owner=std::shared_ptr<uint8_t[]>(new uint8_t[n]);upload.p=upload.owner.get();upload.n=upload.c=n;return upload.p;}catch(const std::exception&e){errorText=e.what();return nullptr;}}
int rr_init(uint32_t n){try{if(n!=upload.n)throw std::runtime_error("Cartridge upload mismatch");auto[rom,e]=nds_Open(upload);if(e)throw std::runtime_error(e.text);machine=dsmachine_New(rom,0);cartridge=upload;bindFrameBoundary(machine);errorText.clear();return 1;}catch(const std::exception&e){errorText=e.what();return 0;}}
int rr_run(uint32_t n){try{if(!machine||n>50000)throw std::runtime_error("Invalid DS execution slice");rrprof::Scope t(0,"ARM9 / ARM7 and scheduler");auto start=machine->Steps;auto r=dsmachine_Machine_RunFrames(machine,1,n,64);if(machine->ARM9->cpu->Halted||machine->ARM7->cpu->Halted)throw std::runtime_error(r.Reason);return int(machine->Steps-start);}catch(const std::exception&e){errorText=e.what();return -1;}}
void rr_pad(uint32_t buttons,int,int){if(machine)dsmachine_Machine_SetKeys(machine,buttons);}
void rr_touch(int x,int y,int down){if(machine)dsmachine_Machine_SetTouch(machine,std::clamp(x,0,255),std::clamp(y,0,191),down!=0);}
const char*rr_error(){return errorText.c_str();}
const char*rr_status(){if(!machine)return "{}";std::ostringstream s;s<<"{\"steps\":"<<machine->Steps<<",\"frames\":"<<machine->vid.frames<<",\"inputSeconds\":"<<double(machine->Steps)/(4288.0*262*60)<<",\"pc\":"<<machine->ARM9->cpu->R[15]<<",\"pc7\":"<<machine->ARM7->cpu->R[15]<<",\"width\":256,\"height\":384,\"commands\":"<<machine->gpu3d->count<<"}";reply=s.str();return reply.c_str();}
const char*rr_proof(){reply=proof(machine);return reply.c_str();}
const char*rr_profile(){return rrprof::json();}
uint8_t*rr_frame(){pixels=frame(machine);return pixels.data();}
}
#include "state.h"
static std::vector<std::shared_ptr<void>> stateOwners;
static void stateWrite(rrstate::Archive&a){rrDSStateVersion=2;a.header(5,2);a(machine);}
static void stateRead(rrstate::Archive&a){if(a.bytes.size()<12)throw std::runtime_error("Truncated DS state");rrDSStateVersion=uint32_t(a.bytes[8])|uint32_t(a.bytes[9])<<8|uint32_t(a.bytes[10])<<16|uint32_t(a.bytes[11])<<24;if(rrDSStateVersion!=1&&rrDSStateVersion!=2)throw std::runtime_error("Unsupported DS state version");a.header(5,rrDSStateVersion);dsmachine_Machine*next=nullptr;a(next);a.finish();rebindState(next,cartridge);bindFrameBoundary(next);machine=next;stateOwners=std::move(a.owned);}
#include "../../../../browser/state/api.inc"
static bool captureSwap=false;
extern "C" {
int rr_capture_begin(){try{auto b=rrds::memory(machine);rrcapture::trace.begin(b.data(),b.size());rrds::events.clear();rrds::render3DEvents.clear();
 machine->OnPoly=[](int64_t cmd){auto&t=rrcapture::trace;rrds::clean();auto m=machine;for(auto&p:m->gpu3d->geom.polys)if(p.cmd==cmd){std::ostringstream s;s<<"{\"kind\":\"GX polygon rasterization\",\"command\":"<<cmd<<",\"POLYGON_ATTR\":"<<p.attr<<",\"TEXIMAGE_PARAM\":"<<p.texParam<<",\"PLTT_BASE\":"<<p.pltt<<",\"vertices\":"<<p.verts.n<<",\"wBuffer\":"<<(p.wbuffer?"true":"false")<<"}";rrds::render3DEvents.insert(t.event(m->Steps,m->ARM9->cpu->R[15],s.str()));break;}};
 machine->OnPixel=[](int64_t x,int64_t y,dsmachine_PixelEvent e){auto&t=rrcapture::trace;auto m=machine;uint32_t v=rrds::frag({e.R,e.G,e.B,e.A});t.record(rrds::plane3D+(y*256+x)*4,v,4,m->Steps,m->ARM9->cpu->R[15],t.current,(e.Drawn?1:0)|(e.ZReject?2:0)|(e.AlphaReject?4:0));};return 1;}catch(const std::exception&e){errorText=e.what();return 0;}}
int rr_capture_end(){machine->OnPoly={};machine->OnPixel={};captureSwap=machine->gpu2d->swap;auto b=rrds::memory(machine);rrcapture::trace.end(b.data(),b.size());return 1;}
const char*rr_capture_info(){reply=rrcapture::trace.info();return reply.c_str();}
const char*rr_pixel(int x,int y){if(x<0||x>=256||y<0||y>=384)return "{\"error\":\"Outside captured display\"}";int engine=(y<192)?!captureSwap:captureSwap;reply=rrcapture::trace.pixel(engine*rrds::planeSize+((y%192)*256+x)*4,4);return reply.c_str();}
const char*rr_source(uint32_t address,int size,uint32_t before,uint32_t expected){reply=rrcapture::trace.pixel(address,size,before,expected);return reply.c_str();}
const char*rr_resource(uint32_t id,uint32_t offset){reply=rrcapture::trace.resourceJSON(id,offset);return reply.c_str();}
}
#include "../../../../browser/core/replay.h"
// Follow the actual draw target until the final step, which always shows the LCDs.
static bool replay3DTarget(){auto&r=rrreplay::replay;return r.cursor>0&&r.cursor<r.steps.size()&&rrds::render3DEvents.count(r.steps[r.cursor-1].event);}
extern "C" {
const char*rr_replay_begin(){rrreplay::replay.begin();reply=rrreplay::replay.info();return reply.c_str();}
int rr_replay_seek(uint32_t step){return rrreplay::replay.seek(step);}
const char*rr_replay_info(){reply=rrreplay::replay.info();reply.pop_back();reply+=replay3DTarget()?",\"surface\":\"3D render target before 2D composition\"}":",\"surface\":\"Composited LCD outputs\"}";return reply.c_str();}
uint32_t rr_replay_for_write(uint32_t id){return rrreplay::replay.forWrite(id);}
uint8_t*rr_replay_frame(){auto&r=rrreplay::replay;pixels.resize(rrds::planeSize*2);if(r.memory.size()<rrds::total)return nullptr;auto first=captureSwap?0:rrds::planeSize;std::memcpy(pixels.data(),r.memory.data()+first,rrds::planeSize);std::memcpy(pixels.data()+rrds::planeSize,r.memory.data()+rrds::planeSize-first,rrds::planeSize);if(replay3DTarget())std::memcpy(pixels.data()+(captureSwap?0:rrds::planeSize),r.memory.data()+rrds::plane3D,rrds::planeSize);return pixels.data();}
}
