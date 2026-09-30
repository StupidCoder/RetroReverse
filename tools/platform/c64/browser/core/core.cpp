#include <cstdint>
#include <cstdio>
#include <cstring>
#include <algorithm>
#include <memory>
#include <string>
#include <cstddef>
#include <array>
#include <vector>
#include <sstream>
#include "observe.h"
#include "../../../../browser/core/memory.h"
#include "../../../../browser/core/profile.h"
static uint64_t profileCycle=0;
#define RR_PROFILE_BEGIN bool rrSample=(++profileCycle%1021)==0; rrprof::Scope rrTick(0,"Bus / glue remainder",rrSample)
#define RR_PROFILE_CALL(id,name,expr) ([&](){rrprof::Scope clock(id,name,rrSample);return (expr);}())
#define CHIPS_IMPL
#include "../vendor/chips/chips_common.h"
#include "../vendor/chips/m6502.h"
#include "../vendor/chips/m6526.h"
#include "../vendor/chips/m6569.h"
#include "../vendor/chips/m6581.h"
#include "../vendor/chips/kbd.h"
#include "../vendor/chips/mem.h"
#include "../vendor/chips/clk.h"
#include "../vendor/chips/m6522.h"
#include "../vendor/systems/c1530.h"
#include "../vendor/systems/c1541.h"
#include "../vendor/systems/c64.h"
#include "core.h"

namespace {
constexpr int MAX_PULSES=2*1024*1024, EVENT_CAP=8192;
static c64_t machine;
struct Context {uint64_t cycles=0,sequence=0;uint32_t pulse=0,remaining=0,frames=0;uint16_t pc=0;bool play=false,synthetic=false;uint32_t generation=0;};
static Context ctx;
struct Checkpoint {c64_t machine;Context ctx;uint32_t version;bool valid;};
static Checkpoint checkpoints[16];
static uint8_t input[2*1024*1024];
static uint32_t durations[MAX_PULSES],pulse_count=0,generation=0;
struct Event {uint64_t cycle,seq;uint32_t pulse;uint16_t pc,address,timer;uint8_t value,kind,a,x,y,p,bank;};
static Event events[EVENT_CAP];static uint32_t event_count=0,dropped=0;
static unsigned traceMask=31;
static bool tracing=false;static uint16_t trace_lo=0,trace_hi=65535;
static uint32_t pixels[392*272];static float audio[8192];static int audio_count=0;
static char error[256],status[2048],event_json[EVENT_CAP*230];
static void fail(const char* message){snprintf(error,sizeof(error),"%s",message);}
static void record(uint8_t kind,uint16_t address,uint8_t value){
 if(!tracing||!(traceMask&(1u<<kind)))return;
 if(event_count==EVENT_CAP){dropped++;return;}
 events[event_count++]={ctx.cycles,ctx.sequence++,ctx.pulse,ctx.pc,address,machine.cia_1.ta.counter,value,kind,machine.cpu.A,machine.cpu.X,machine.cpu.Y,machine.cpu.P,machine.cpu_port};
}
static void samples(const float* data,int n,void*){int take=std::min(n,8192-audio_count);memcpy(audio+audio_count,data,take*sizeof(float));audio_count+=take;}
static uint32_t hash(const uint8_t* p,size_t n){uint32_t h=2166136261u;for(size_t i=0;i<n;i++)h=(h^p[i])*16777619u;return h;}
static bool initialized=false;
static int stopReason=0;
static void invalidateDebug();
}
#include "observe.inc"
#include "raster.inc"
#include "memory.inc"
extern "C" {
const char* rr_profile(){return rrprof::json(true);}
uint8_t* rr_input(){return input;}
const char* rr_error(){return error;}
uint8_t* rr_ram(){return machine.ram;}
int rr_bus_flags(){return (machine.pins&M6502_SYNC?1:0)|(machine.pins&M6502_RW?2:0);}
uint32_t rr_bus(){return uint32_t(machine.pins&0xffffff);}
int rr_width(){return 392;} int rr_height(){return 272;}
int rr_init(int b,int k,int c){
 if(b!=8192||k!=8192||c!=4096){fail("Expected BASIC 8192, KERNAL 8192 and character ROM 4096 bytes.");return 0;}
 c64_desc_t desc={};desc.roms.basic={input,8192};desc.roms.kernal={input+8192,8192};desc.roms.chars={input+16384,4096};desc.audio.sample_rate=44100;desc.audio.num_samples=256;desc.audio.callback={samples,nullptr};
 c64_init(&machine,&desc);invalidateDebug();ctx={};observation::reset();tracing=false;traceMask=31;ctx.generation=++generation;event_count=dropped=0;audio_count=0;error[0]=0;initialized=true;pulse_count=0;for(auto& cp:checkpoints)cp.valid=false;return 1;
}
int rr_tape(int size){
 if(!initialized){fail("Initialize the machine first.");return 0;}
 if(size<20||size>int(sizeof(input))||memcmp(input,"C64-TAPE-RAW",12)||input[12]>1||input[13]!=0||input[14]!=0){fail("Expected a PAL C64 TAP v0 or v1 image.");return 0;}
 uint32_t bytes=uint32_t(input[16])|(uint32_t(input[17])<<8)|(uint32_t(input[18])<<16)|(uint32_t(input[19])<<24);
 if(bytes==0||bytes!=uint32_t(size-20)){fail("TAP length does not match its header.");return 0;}
 // Validate before replacing the active tape.
 uint32_t count=0;for(int i=20;i<size;){uint32_t n=input[i++]*8;if(!n){if(input[12]==0)n=2048;else{if(i+3>size){fail("Truncated TAP extended pulse.");return 0;}n=input[i]|(input[i+1]<<8)|(input[i+2]<<16);i+=3;}}if(n==0||++count>MAX_PULSES){fail("Empty pulse or tape exceeds pulse capacity.");return 0;}}
 count=0;for(int i=20;i<size;){uint32_t n=input[i++]*8;if(!n){if(input[12]==0)n=2048;else{n=input[i]|(input[i+1]<<8)|(input[i+2]<<16);i+=3;}}durations[count++]=n;}
 pulse_count=count;ctx.pulse=0;ctx.remaining=durations[0];ctx.generation=++generation;for(auto& cp:checkpoints)cp.valid=false;return 1;
}
int rr_prepare(int pc,int pulse){
 if(!initialized||pulse<0||uint32_t(pulse)>=pulse_count||pc<0||pc>65535){fail("Invalid prepared segment.");return 0;}
 invalidateDebug();memcpy(machine.ram,input,65536);ctx.synthetic=true;ctx.pulse=pulse;ctx.remaining=durations[pulse];
 // Explicit prototype bootstrap; no CPU instructions or ROM calls are replaced during execution.
 machine.cpu.io_ddr=0x2f;machine.cpu.io_out=0x37;machine.cpu.io_pins=0x37;_c64_cpu_port_out(0x37,&machine);
 machine.cpu.PC=pc;machine.cpu.P=M6502_IF;machine.cpu.S=0xfd;machine.cpu.brk_flags=0;machine.cpu.irq_pip=0;machine.cpu.nmi_pip=0;
 machine.pins=M6502_SYNC|M6502_RW;M6502_SET_ADDR(machine.pins,pc);M6502_SET_DATA(machine.pins,machine.ram[pc]);ctx.pc=pc;return 1;
}
void rr_play(int down){ctx.play=down!=0;if(down)machine.cas_port&=~C64_CASPORT_SENSE;else machine.cas_port|=C64_CASPORT_SENSE;}
void rr_key(int code,int down){if(down)c64_key_down(&machine,code);else c64_key_up(&machine,code);}
void rr_joystick(int port,int mask){if(port==1)machine.joy_joy1_mask=mask&31;else if(port==2)machine.joy_joy2_mask=mask&31;}
void rr_trace_mask(int mask){traceMask=unsigned(mask)&31;}
void rr_trace(int enabled,int lo,int hi){traceMask=31;tracing=enabled!=0;trace_lo=lo;trace_hi=hi;event_count=dropped=0;}
int rr_stop_reason(){return stopReason;}
int rr_run(int ticks,int stop_kind,int target){
 stopReason=0;
 if(!initialized||ticks<0||ticks>1000000){fail("Execution slice must be 0–1000000 PAL cycles.");return -1;}
 audio_count=0;int done=0;bool leftBoundary=!(machine.pins&M6502_SYNC);
 for(;done<ticks;){
  if(ctx.synthetic&&(machine.pins&M6502_SYNC)){
   const auto pc=M6502_GET_ADDR(machine.pins);
   if((pc>=0xe000&&(machine.cpu_port&2))||(pc>=0xa000&&pc<0xc000&&(machine.cpu_port&3)==3)){
    snprintf(error,sizeof(error),"Prepared segment reached ROM at $%04X. Supply real ROMs and reset for authentic boot.",pc);return -1;
   }
  }
  machine.cas_port&=~C64_CASPORT_READ;
  bool edge=false;
  if(ctx.play&&!(machine.cas_port&C64_CASPORT_MOTOR)&&ctx.pulse<pulse_count){
   if(ctx.remaining>0)--ctx.remaining;
   if(ctx.remaining==0){machine.cas_port|=C64_CASPORT_READ;ctx.pulse++;if(ctx.pulse<pulse_count)ctx.remaining=durations[ctx.pulse];edge=true;}
  }
  if(machine.pins&M6502_SYNC){ctx.pc=M6502_GET_ADDR(machine.pins);if(!(machine.pins&M6502_RDY))observation::instruction();}
  uint16_t previous_line=machine.vic.rs.v_count;
  if(observation::capturing)observation::beginTick();
  machine.pins=_c64_tick(&machine,machine.pins);ctx.cycles++;done++;
  if(previous_line>machine.vic.rs.v_count)ctx.frames++;
  if(ctx.cycles%985==0)kbd_update(&machine.kbd,1000);
  const auto pins=machine.pins;const uint16_t address=M6502_GET_ADDR(pins);const uint8_t value=M6502_GET_DATA(pins);
  if((pins&M6502_RW)&&!(pins&M6502_SYNC))observation::read(address,value);
  if(edge){record(0,0,0);rrmem::access(ctx.cycles,5,ctx.pulse-1,durations[ctx.pulse-1],1,17,ctx.pc,0);}
  if(rrmem::active&&(pins&M6502_RW)&&!(pins&M6502_RDY)){
   const int space=observation::cpuSpace(address);unsigned region=0,off=address;
   if(space==1){region=4;off=address&1023;}
   else if(space==3){region=address>=0xe000?2:address>=0xd000?3:1;off=address-(region==2?0xe000:region==3?0xd000:0xa000);}
   if(space!=2)rrmem::access(ctx.cycles,region,off,value,1,pins&M6502_SYNC?4:1,ctx.pc,address);
  }
  const bool selected=ctx.pc>=trace_lo&&ctx.pc<=trace_hi;
  if(selected){if(pins&M6502_SYNC)record(1,address,value);else if(!(pins&M6502_RW))record(3,address,value);else if(address>=0xdc00&&address<=0xddff)record(2,address,value);else if(observation::traceReads)record(4,address,value);}
  if(!(pins&M6502_SYNC))leftBoundary=true;
  if((stop_kind==6&&leftBoundary&&(pins&M6502_SYNC))||(stop_kind==7&&previous_line>machine.vic.rs.v_count)||(stop_kind==1&&edge)||(stop_kind==5&&edge&&ctx.pulse==uint32_t(target))||(stop_kind==2&&(pins&M6502_SYNC)&&address==target)||(stop_kind==3&&!(pins&M6502_RW)&&address==target)||(stop_kind==4&&!(pins&M6502_RW)&&ctx.pc==target)){stopReason=stop_kind;break;}
 }
 return done;
}
int rr_checkpoint(int slot){if(slot<0||slot>=16||!initialized)return 0;auto& cp=checkpoints[slot];cp.version=c64_save_snapshot(&machine,&cp.machine);cp.ctx=ctx;cp.valid=true;if(!observation::saved[slot])observation::saved[slot]=std::make_unique<observation::History>();*observation::saved[slot]=observation::history;return 1;}
int rr_restore(int slot){if(slot<0||slot>=16)return 0;auto& cp=checkpoints[slot];if(!cp.valid||cp.ctx.generation!=ctx.generation){fail("Checkpoint belongs to a different machine or tape generation.");return 0;}if(!c64_load_snapshot(&machine,cp.version,&cp.machine))return 0;invalidateDebug();ctx=cp.ctx;observation::history=*observation::saved[slot];observation::capturing=false;event_count=dropped=0;audio_count=0;return 1;}
int rr_corrupt(int pulse,int duration){if(pulse<0||uint32_t(pulse)>=pulse_count||duration<=0)return 0;durations[pulse]=duration;ctx.generation=++generation;return 1;}
uint32_t* rr_frame(){const auto* palette=static_cast<const uint32_t*>(m6569_palette().ptr);for(int y=0;y<272;y++)for(int x=0;x<392;x++)pixels[y*392+x]=palette[machine.fb[y*M6569_FRAMEBUFFER_WIDTH+x]&15];return pixels;}
float* rr_audio(){return audio;}int rr_audio_count(){return audio_count;}
const char* rr_status(){
 snprintf(status,sizeof(status),"{\"schema\":1,\"cycle\":%llu,\"pc\":%u,\"a\":%u,\"x\":%u,\"y\":%u,\"s\":%u,\"p\":%u,\"pulse\":%u,\"pulseCount\":%u,\"remaining\":%u,\"frames\":%u,\"raster\":%u,\"timer\":%u,\"bank\":%u,\"motor\":%s,\"prepared\":%s,\"ramHash\":%u,\"frameHash\":%u,\"events\":%u,\"dropped\":%u,\"generation\":%u}",static_cast<unsigned long long>(ctx.cycles),ctx.pc,machine.cpu.A,machine.cpu.X,machine.cpu.Y,machine.cpu.S,machine.cpu.P,ctx.pulse,pulse_count,ctx.remaining,ctx.frames,machine.vic.rs.v_count,machine.cia_1.ta.counter,machine.cpu_port,(machine.cas_port&C64_CASPORT_MOTOR)?"false":"true",ctx.synthetic?"true":"false",hash(machine.ram,65536),hash(machine.fb,sizeof(machine.fb)),event_count,dropped,ctx.generation);return status;
}
const char* rr_events(){
 size_t pos=0;event_json[pos++]='[';
 for(uint32_t i=0;i<event_count;i++){auto& e=events[i];pos+=snprintf(event_json+pos,sizeof(event_json)-pos,"%s{\"id\":%llu,\"cycle\":%llu,\"pulse\":%u,\"kind\":%u,\"pc\":%u,\"address\":%u,\"value\":%u,\"timer\":%u,\"a\":%u,\"x\":%u,\"y\":%u,\"p\":%u,\"bank\":%u}",i?",":"",(unsigned long long)e.seq,(unsigned long long)e.cycle,e.pulse,e.kind,e.pc,e.address,e.value,e.timer,e.a,e.x,e.y,e.p,e.bank);}
 event_json[pos++]=']';event_json[pos]=0;return event_json;
}
}
#include "debug.inc"
#include "../../../../browser/state/archive.h"
#include "state-fields.h"
static void contextFields(rrstate::Archive&a,Context&v){a(v.cycles,v.sequence,v.pulse,v.remaining,v.frames,v.pc,v.play,v.synthetic);}
static void stateWrite(rrstate::Archive&a){a.header(1,1);a(machine);contextFields(a,ctx);}
static void stateRead(rrstate::Archive&a){
 a.header(1,1);auto next=std::make_unique<c64_t>(machine);Context context=ctx;a(*next);contextFields(a,context);a.finish();
 if(context.pulse>pulse_count||next->audio.num_samples<1||next->audio.num_samples>C64_MAX_AUDIO_SAMPLES||next->audio.sample_pos<0||next->audio.sample_pos>=next->audio.num_samples)throw std::runtime_error("Invalid C64 state");
 invalidateDebug();machine=*next;ctx=context;_c64_update_memory_map(&machine);observation::reset();tracing=false;audio_count=0;event_count=dropped=0;for(auto&cp:checkpoints)cp.valid=false;
}
#include "../../../../browser/state/api.inc"
extern "C" const char* rr_capture_info(){static std::string s;s="{\"rasterBytes\":"+std::to_string(sizeof(raster::lines)+sizeof(raster::charROM)+sizeof(raster::usedMemory)+(raster::initialWriters?sizeof(*raster::initialWriters):0)+raster::writes.size()*sizeof(observation::Write))+",\"events\":"+std::to_string(observation::refCount)+",\"writes\":"+std::to_string(observation::graphCount)+",\"overflow\":"+std::to_string(observation::overflow)+",\"bytes\":"+std::to_string(sizeof(observation::captured)+observation::refCount*sizeof(observation::Ref)+observation::graphCount*sizeof(observation::Graphics)+observation::controlCount*sizeof(observation::Control))+"}";return s.c_str();}
