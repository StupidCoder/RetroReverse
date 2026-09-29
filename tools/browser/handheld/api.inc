// Shared browser ABI; platform host.h provides Machine, boot, advance and memory.
static Machine*machine=nullptr;
static std::vector<uint8_t>input;
static std::string errorText,reply;
extern "C" {
uint8_t*rr_input(uint32_t n){try{if(n>4*1024*1024+512)throw std::runtime_error("Cartridge exceeds 4 MiB");input.resize(n);return input.data();}catch(const std::exception&e){errorText=e.what();return nullptr;}}
int rr_init(uint32_t n){try{if(machine)throw std::runtime_error("Reset requires a new worker");if(n!=input.size())throw std::runtime_error("Cartridge length mismatch");machine=rrBoot(input);errorText.clear();return 1;}catch(const std::exception&e){errorText=e.what();return 0;}}
int rr_run(uint32_t n){try{if(!machine||!n||n>1000000)throw std::runtime_error("Invalid execution slice");rrprof::Scope measured(0,HH_CPU_NAME);auto frame=rrhh::video.frames;unsigned done=0;while(done<n&&rrhh::video.frames==frame){if(machine->CPU->Halted)throw std::runtime_error(machine->CPU->HaltReason);rrhh::video.pc=machine->CPU->PC;rrhh::video.steps++;rrAdvance(machine);done++;}return done;}catch(const std::exception&e){errorText=e.what();return -1;}}
void rr_pad(uint32_t buttons){if(machine)rrPad(machine,buttons);}
const char*rr_error(){return errorText.c_str();}
const char*rr_status(){std::ostringstream s;s<<"{\"steps\":"<<rrhh::video.steps<<",\"frames\":"<<rrhh::video.frames<<",\"pc\":"<<(machine?machine->CPU->PC:0)<<",\"seconds\":"<<rrSeconds(machine)<<",\"width\":160,\"height\":144}";reply=s.str();return reply.c_str();}
const char*rr_profile(){return rrprof::json();}
uint8_t*rr_frame(){return reinterpret_cast<uint8_t*>(rrhh::video.screen.data());}
const char*rr_proof(){auto mem=rrMemoryImage(machine);std::ostringstream s;s<<"{\"steps\":"<<rrhh::video.steps<<",\"frames\":"<<rrhh::video.frames<<",\"pc\":"<<machine->CPU->PC<<",\"memory\":"<<rrhh::hash(mem.data(),mem.size())<<",\"rgba\":"<<rrhh::hash(rr_frame(),rrhh::pixels*4)<<"}";reply=s.str();return reply.c_str();}
int rr_capture_begin(){try{
#ifdef HH_CAPTURE_BEGIN
HH_CAPTURE_BEGIN;
#endif
auto b=rrMemoryImage(machine);rrcapture::trace.begin(b.data(),b.size());return 1;}catch(const std::exception&e){errorText=e.what();return 0;}}
int rr_capture_end(){try{auto b=rrMemoryImage(machine);rrcapture::trace.end(b.data(),b.size());return 1;}catch(const std::exception&e){errorText=e.what();return 0;}}
const char*rr_capture_info(){reply=rrcapture::trace.info();
#ifdef HH_CAPTURE_BYTES
reply.pop_back();reply+=",\"rasterBytes\":"+std::to_string(HH_CAPTURE_BYTES)+"}";
#endif
return reply.c_str();}
const char*rr_pixel(int x,int y){uint32_t a=x>=0&&x<160&&y>=0&&y<144?rrhh::frameBase+(y*160+x)*4:UINT32_MAX;reply=rrcapture::trace.pixel(a,4);return reply.c_str();}
const char*rr_source(uint32_t a,int size,uint32_t before,uint32_t expected){reply=rrcapture::trace.pixel(a,size,before,expected);return reply.c_str();}
const char*rr_resource(uint32_t id,uint32_t offset){reply=rrcapture::trace.resourceJSON(id,offset);return reply.c_str();}
const char*rr_replay_begin(){rrreplay::replay.begin();reply=rrreplay::replay.info();return reply.c_str();}
int rr_replay_seek(uint32_t n){return rrreplay::replay.seek(n);}
const char*rr_replay_info(){reply=rrreplay::replay.info();return reply.c_str();}
uint32_t rr_replay_for_write(uint32_t id){return rrreplay::replay.forWrite(id);}
uint8_t*rr_replay_frame(){auto&b=rrreplay::replay.memory;return b.size()>=rrhh::memorySize?b.data()+rrhh::frameBase:nullptr;}
}
#include HH_STATE_FILE
#include "../state/api.inc"
