#include "core.h"
#include "../state.h"
#include "../../../../browser/state/archive.h"
#include "../../../../browser/core/memory.h"
#include "../../../../browser/core/profile.h"
#include <algorithm>
#include <memory>
#include <sstream>
#include <map>
#include <tuple>
#include <cstdio>
namespace driveDebug {void reset();}
namespace render {void reset();void beforeTick();void afterTick(const rr::c64::CpuState&,int,uint8_t);void edited(uint16_t);}

namespace bridge {
using namespace rr::c64;
System machine;
Board& board=machine.board;
struct Context {
 uint16_t pc=0;
 uint64_t instructionCycle=0,sequence=0;
 bool synthetic=false;
 std::array<uint8_t,3> instruction{};
 uint32_t interruptDepth=0;
 std::array<bool,256> keys{};
};
Context ctx;
static void stateFields(rrstate::Archive& a,Context& s){a(s.pc,s.instructionCycle,s.synthetic,s.keys,s.instruction,s.interruptDepth);}
bool initialized=false,driveEnabled=false,bound=false;
uint32_t generation=0;
StateIdentity identity;
std::array<uint8_t,2*1024*1024> input{};
std::array<uint32_t,392*272> pixels{};
std::string error,status,eventsJSON;
int stopReason=0;
struct Event {uint64_t cycle,id;uint32_t pulse;uint16_t pc,address,timer;uint8_t kind,value,a,x,y,p,bank;};
std::vector<Event> events;
uint32_t dropped=0;
bool tracing=false,traceReads=false;
unsigned traceMask=31;
uint16_t traceLow=0,traceHigh=65535;
struct Checkpoint {uint32_t generation=0;std::vector<uint8_t> bytes;};
std::array<Checkpoint,16> checkpoints;
std::vector<uint8_t> stateInput,stateOutput;
constexpr uint32_t MaxState=12*1024*1024+4096;
struct Job {
 int mode=0,target=0,space=0,result=0,calls=0,interrupts=0,returnSP=0;
 int value=0,mask=0,writer=-1;
 bool active=false,skip=false,inFlight=false,entry=false;
 uint8_t opcode=0;
};
Job job;
std::string writeEvent="null";
int lastExecuted=-1;
uint64_t lastExecutedCycle=0;
bool boundary(){return board.state.cpu.stage==Stage::Fetch;}
bool interrupt(){return board.state.cpu.interruptPending;}
bool fault(){return board.state.cpu.stage==Stage::Fault||board.state.cpu.stage==Stage::Jam;}
void invalidateDebug(){job={};writeEvent="null";lastExecuted=-1;lastExecutedCycle=0;}
void clearObservation(){driveDebug::reset();render::reset();events.clear();dropped=0;ctx.sequence=0;tracing=false;rrmem::events.clear();rrmem::dropped=0;rrmem::active=false;invalidateDebug();}
void mediaChanged(){++generation;bound=false;stateOutput.clear();for(auto& c:checkpoints)c={};clearObservation();}
int space(uint16_t a,bool writing=false){
 if(a<2)return 2;
 const auto p=board.port();
 if(a>=0xd000&&a<0xe000&&(p&3)&&(p&4))return a<0xd800||a>=0xdc00?2:1;
 if(!writing&&((a>=0xe000&&(p&2))||(a>=0xa000&&a<0xc000&&(p&3)==3)||(a>=0xd000&&a<0xe000&&(p&3)&&!(p&4))))return 3;
 return 0;
}
uint32_t hash(std::span<const uint8_t> bytes){uint32_t h=2166136261;for(auto b:bytes)h=(h^b)*16777619;return h;}
void record(uint8_t kind,uint16_t address,uint8_t value){
 if(!tracing||!(traceMask&(1u<<kind)))return;
 if(events.size()==8192){++dropped;return;}
 const auto& s=board.state;const auto& c=s.cpu;
 events.push_back({s.cycles,ctx.sequence++,s.tape.pulse,ctx.pc,address,s.cia1.a.counter,kind,value,c.a,c.x,c.y,c.p,board.port()});
}
// Called immediately around the one C64 edge. Drive edges preceding it share
// the same scheduler but must not be mistaken for a C64 memory transaction.
bool tick(){
 if(fault()){error="C64 CPU is halted.";return false;}
 while(driveEnabled&&machine.nextProcessor()==Processor::Drive)if(!machine.tickEdge()){error="1541 CPU halted on a JAM or unsupported opcode.";return false;}
 const auto before=board.state.cpu;const auto pulse=board.state.tape.pulse;
 const int mapped=space(before.bus.address,before.bus.write);
 if(before.stage==Stage::Fetch){ctx.pc=before.pc;ctx.instructionCycle=board.state.cycles;}
 const auto old=!before.bus.write?uint8_t(0):mapped==0?board.state.ram[before.bus.address]:mapped==1?board.state.color[before.bus.address&1023]:board.peek(before.bus.address);render::beforeTick();
 const bool ok=driveEnabled?machine.tickEdge():board.tick();
 render::afterTick(before,mapped,old);
 const auto& s=board.state;const auto& b=s.lastBus;
 if(before.retired!=s.cpu.retired){lastExecuted=ctx.pc;lastExecutedCycle=s.cycles;}
 if(b.valid&&(b.write||b.ready)){
  unsigned region=0,offset=b.address;
  if(mapped==1){region=4;offset=b.address&1023;}
  if(mapped==3){region=b.address>=0xe000?2:b.address>=0xd000?3:1;offset=b.address-(region==2?0xe000:region==3?0xd000:0xa000);}
  if(mapped!=2)rrmem::access(s.cycles,region,offset,b.value,1,b.write?2:b.sync?4:1,ctx.pc,b.address);
  if(ctx.pc>=traceLow&&ctx.pc<=traceHigh){
   if(b.write)record(3,b.address,b.value);
   else if(b.sync)record(1,b.address,b.value);
   else if(b.address>=0xdc00&&b.address<=0xddff)record(2,b.address,b.value);
   else if(traceReads)record(4,b.address,b.value);
  }
 }
 if(pulse!=s.tape.pulse){record(0,0,0);rrmem::access(s.cycles,driveEnabled?7:5,pulse,pulse<board.pulses.size()?board.pulses[pulse]:0,1,17,ctx.pc,0);}
 if(!ok)error="C64 CPU halted on a JAM or unsupported opcode.";
 return ok;
}
bool advance(){
 if(ctx.synthetic&&boundary()&&space(board.state.cpu.pc)==3){error="Synthetic program reached ROM; reset for an authentic boot.";return false;}
 return tick();
}
void applyKeys(){
 board.state.keys.fill(0);board.state.restore=ctx.keys[255];
 const char* columns[]={"\x01\r\x09\xf7\xf1\xf3\xf5\x0a","3WA4ZSE\x00","5RD6CFTX","7YG8BHUV","9IJ0MKON","+PL-.:@,","\\*;\x0c\x00=^/",("1_\x0e" "2 \x0fQ\x03")};
 for(unsigned code=1;code<255;code++)if(ctx.keys[code]){
  unsigned key=code;bool shift=false;
  if(key>='a'&&key<='z')key-=32;
  if(key>=242&&key<=248&&!(key&1)){key--;shift=true;}
  if(key==16){key=1;shift=true;}if(key==2){key=12;shift=true;}if(key==7){key=3;shift=true;}
  const std::string shifted="!\"#$%&'()<>?[]";const std::string base="123456789,./:;";
  const auto at=shifted.find(char(key));if(at!=std::string::npos){key=uint8_t(base[at]);shift=true;}
  for(unsigned col=0;col<8;col++)for(unsigned row=0;row<8;row++)if(uint8_t(columns[col][row])==key)board.key(col,row,true);
  if(shift)board.key(1,7,true);
 }
}
std::vector<uint8_t> snapshot(){
 auto hardware=driveEnabled?saveState(machine,identity):saveState(board,identity);
 rrstate::Archive a;a.header(0x41343643,2);a(driveEnabled,ctx,hardware);auto checksum=hash(a.bytes);a(checksum);return std::move(a.bytes);
}
bool restore(std::span<const uint8_t> bytes){
 try{
  if(bytes.size()<4||bytes.size()>MaxState)throw std::runtime_error("Invalid adapter state size");
  const auto n=bytes.size()-4;uint32_t checksum=0;for(unsigned i=0;i<4;i++)checksum|=uint32_t(bytes[n+i])<<(i*8);
  if(checksum!=hash(bytes.first(n)))throw std::runtime_error("Adapter state checksum mismatch");
  rrstate::Archive a(bytes.data(),n);a.header(0x41343643,2);bool drive=false;Context next;a(drive,next);
  if(drive!=driveEnabled)throw std::runtime_error("Checkpoint drive configuration mismatch");
  const auto count=a.count(0);if(count!=a.bytes.size()-a.pos)throw std::runtime_error("Invalid hardware state length");
  const std::span<const uint8_t> hardware(a.bytes.data()+a.pos,count);
  // Hardware load validates and commits atomically. All adapter parsing and
  // checks precede it; remaining assignments cannot allocate or fail.
  if(!(driveEnabled?loadState(machine,hardware,identity,error):loadState(board,hardware,identity,error)))return false;
  ctx=next;clearObservation();stopReason=0;error.clear();return true;
 }catch(const std::exception& e){error=e.what();return false;}
}
}
using namespace bridge;
extern "C" {
const char* rr_capabilities(){return R"({"core":"owned-c64","development":false,"execution":true,"debugger":true,"memoryActivity":true,"portableState":true,"stateIdentityBinding":true,"renderCapture":true,"pixelProvenance":true,"audio":false})";}
uint8_t* rr_input(){return input.data();}
const char* rr_error(){return error.c_str();}
const char* rr_profile(){return rrprof::json();}
uint8_t* rr_ram(){return board.state.ram.data();}
int rr_width(){return 392;}int rr_height(){return 272;}
int rr_init(int b,int k,int c){
 if(b!=8192||k!=8192||c!=4096){error="Expected BASIC 8192, KERNAL 8192 and character ROM 4096 bytes.";return 0;}
 std::copy_n(input.begin(),8192,board.basic.begin());std::copy_n(input.begin()+8192,8192,board.kernal.begin());std::copy_n(input.begin()+16384,4096,board.chars.begin());
 board.pulses.clear();machine.drive.eject();machine.drive.rom.fill(0);machine.power();ctx={};identity={};driveEnabled=false;initialized=true;tracing=traceReads=false;traceMask=31;stopReason=0;mediaChanged();error.clear();return 1;
}
int rr_tape(int size){
 if(!initialized||size<20||size>int(input.size())||!board.loadTape(std::span(input).first(size))){error="Invalid PAL TAP v0/v1 image, or machine not initialized.";return 0;}
 mediaChanged();error.clear();return 1;
}
int rr_drive_rom(int size){
 if(!initialized||board.state.cycles||size!=16384){error="Load the 16 KiB 1541 firmware before starting execution.";return 0;}
 std::copy_n(input.begin(),16384,machine.drive.rom.begin());driveEnabled=true;machine.drive.power();mediaChanged();error.clear();return 1;
}
int rr_disk(int size,int protect){
 if(!initialized||!driveEnabled||size<0||size>int(input.size())||!machine.drive.mount(std::span(input).first(size),protect!=0)){error="Invalid D64/G64 image, or 1541 firmware missing.";return 0;}
 mediaChanged();error.clear();return 1;
}
int rr_prepare(int pc,int pulse){
 if(!initialized||pc<0||pc>65535||pulse<0||unsigned(pulse)>=board.pulses.size()){error="Invalid synthetic start.";return 0;}
 std::copy_n(input.begin(),65536,board.state.ram.begin());Cpu cpu;cpu.start(uint16_t(pc));cpu.state.s=0xfd;board.state.cpu=cpu.state;
 board.state.ddr=0x2f;board.state.data=0x37;board.state.tape.pulse=pulse;board.state.tape.remaining=board.pulses[pulse];ctx.synthetic=true;ctx.pc=pc;ctx.instructionCycle=board.state.cycles;
 clearObservation();error.clear();return 1;
}
void rr_play(int down){board.play(down!=0);}
void rr_key(int code,int down){if(code>0&&code<256){ctx.keys[code]=down!=0;applyKeys();}}
void rr_joystick(int port,int mask){if(port==1)board.state.joy1=mask&31;else if(port==2)board.state.joy2=mask&31;}
void rr_trace(int enabled,int lo,int hi){tracing=enabled!=0;traceLow=uint16_t(lo);traceHigh=uint16_t(hi);traceMask=31;events.clear();dropped=0;}
void rr_trace_mask(int mask){traceMask=unsigned(mask)&31;}
void rr_trace_reads(int value){traceReads=value!=0;}
int rr_stop_reason(){return stopReason;}
double rr_cycle(){return double(board.state.cycles);}
double rr_previous(){return double(ctx.instructionCycle);}
int rr_run(int ticks,int kind,int target){
 stopReason=0;
 if(!initialized||ticks<0||ticks>1000000||kind<0||kind>7||target<0||(kind>=2&&kind<=4&&target>65535)){error="Invalid execution slice or stop condition.";return -1;}
 rrprof::Scope profile(0,driveEnabled?"Owned C64 + 1541":"Owned C64");
 const auto frame=board.state.vic.frames;bool left=!boundary();
 for(int done=0;done<ticks;){
  const auto pulse=board.state.tape.pulse;if(!advance())return -1;++done;
  const auto& bus=board.state.lastBus;const bool edge=pulse!=board.state.tape.pulse;
  if(!boundary())left=true;
  if((kind==1&&edge)||(kind==2&&boundary()&&board.state.cpu.pc==target)||(kind==3&&bus.valid&&bus.write&&bus.address==target)||(kind==4&&bus.valid&&bus.write&&ctx.pc==target)||(kind==5&&edge&&board.state.tape.pulse==unsigned(target))||(kind==6&&left&&boundary())||(kind==7&&board.state.vic.frames!=frame)){stopReason=kind;return done;}
 }
 return ticks;
}
// The public bus view follows the pending transaction (as the existing browser
// ABI does); actual completed accesses live in the memory/trace recorders.
uint32_t rr_bus(){const auto& b=board.state.cpu.bus;return b.address|(uint32_t(b.write?b.data:board.peek(b.address))<<16);}
int rr_bus_flags(){const auto& b=board.state.cpu.bus;return (b.sync?1:0)|(b.write?0:2);}
uint32_t* rr_frame(){
 static constexpr uint32_t palette[]={0xff000000,0xffffffff,0xff383381,0xffc8ce75,0xff973c8e,0xff4dac56,0xff9b2c2e,0xff71f1ed,0xff29508e,0xff003855,0xff716cc4,0xff4a4a4a,0xff7b7b7b,0xff9fffa9,0xffeb6d70,0xffb2b2b2};
 for(unsigned y=0;y<272;y++)for(unsigned x=0;x<392;x++)pixels[y*392+x]=palette[board.state.vic.pixels[(y+16)*504+(x+496)%504]&15];return pixels.data();
}
float* rr_audio(){return nullptr;}int rr_audio_count(){return 0;}
const char* rr_status(){
 const auto& s=board.state;const auto& c=s.cpu;std::ostringstream o;
 o<<"{\"schema\":1,\"cycle\":"<<s.cycles<<",\"pc\":"<<ctx.pc<<",\"a\":"<<unsigned(c.a)<<",\"x\":"<<unsigned(c.x)<<",\"y\":"<<unsigned(c.y)<<",\"s\":"<<unsigned(c.s)<<",\"p\":"<<unsigned(c.p)
  <<",\"pulse\":"<<s.tape.pulse<<",\"pulseCount\":"<<board.pulses.size()<<",\"remaining\":"<<s.tape.remaining<<",\"frames\":"<<s.vic.frames<<",\"raster\":"<<s.vic.raster<<",\"timer\":"<<s.cia1.a.counter<<",\"bank\":"<<unsigned(board.port())
  <<",\"motor\":"<<(board.motor()?"true":"false")<<",\"prepared\":"<<(ctx.synthetic?"true":"false")<<",\"ramHash\":"<<hash(s.ram)<<",\"frameHash\":"<<hash(s.vic.pixels)<<",\"events\":"<<events.size()<<",\"dropped\":"<<dropped<<",\"generation\":"<<generation<<"}";status=o.str();return status.c_str();
}
const char* rr_drive_status(){static std::string out;const auto& d=machine.drive.state;std::ostringstream o;o<<"{\"enabled\":"<<(driveEnabled?"true":"false")<<",\"cycle\":\""<<d.clocks<<"\",\"pc\":"<<d.cpu.pc<<",\"halfTrack\":"<<unsigned(d.halfTrack)<<",\"g64\":"<<(d.media.g64?"true":"false")<<",\"track\":"<<(d.media.g64?double(d.halfTrack-2):double(d.halfTrack-2)/2)<<",\"bitCount\":"<<d.media.tracks[d.halfTrack-2].size()*8<<",\"bit\":"<<d.position<<",\"motor\":"<<(d.motor?"true":"false")<<",\"led\":"<<(d.led?"true":"false")<<",\"dirty\":"<<(d.media.changed()?"true":"false")<<"}";out=o.str();return out.c_str();}
const char* rr_events(){std::ostringstream o;o<<'[';bool comma=false;for(const auto& e:events){if(comma)o<<',';comma=true;o<<"{\"id\":"<<e.id<<",\"cycle\":"<<e.cycle<<",\"pulse\":"<<e.pulse<<",\"kind\":"<<unsigned(e.kind)<<",\"pc\":"<<e.pc<<",\"address\":"<<e.address<<",\"value\":"<<unsigned(e.value)<<",\"timer\":"<<e.timer<<",\"a\":"<<unsigned(e.a)<<",\"x\":"<<unsigned(e.x)<<",\"y\":"<<unsigned(e.y)<<",\"p\":"<<unsigned(e.p)<<",\"bank\":"<<unsigned(e.bank)<<'}';}o<<']';eventsJSON=o.str();return eventsJSON.c_str();}
int rr_checkpoint(int slot){if(!initialized||slot<0||slot>=16)return 0;try{auto bytes=snapshot();checkpoints[slot]={generation,std::move(bytes)};return 1;}catch(const std::exception& e){error=e.what();return 0;}}
int rr_restore(int slot){if(slot<0||slot>=16||checkpoints[slot].bytes.empty()||checkpoints[slot].generation!=generation){error="Checkpoint belongs to another machine or media generation.";return 0;}return restore(checkpoints[slot].bytes);}
int rr_corrupt(int pulse,int duration){if(pulse<0||unsigned(pulse)>=board.pulses.size()||duration<=0||duration>0xffffff)return 0;board.pulses[pulse]=duration;if(board.state.tape.pulse==unsigned(pulse))board.state.tape.remaining=duration;mediaChanged();return 1;}
int rr_state_bind(){
 if(!initialized)return 0;unsigned at=0;for(auto* d:{&identity.core,&identity.basic,&identity.kernal,&identity.characters,&identity.driveRom,&identity.tape,&identity.disk}){std::copy_n(input.begin()+at,32,d->begin());at+=32;}
 identity.configuration=uint32_t(input[at])|(uint32_t(input[at+1])<<8)|(uint32_t(input[at+2])<<16)|(uint32_t(input[at+3])<<24);bound=true;stateOutput.clear();return 1;
}
uint8_t* rr_state_input(uint32_t n){try{if(n>MaxState)throw std::runtime_error("State exceeds adapter limit");stateInput.resize(n);return stateInput.data();}catch(const std::exception& e){error=e.what();return nullptr;}}
uint32_t rr_state_save(){try{if(!initialized||!bound)throw std::runtime_error("Bind the verified session identities before saving");stateOutput=snapshot();return uint32_t(stateOutput.size());}catch(const std::exception& e){error=e.what();stateOutput.clear();return 0;}}
uint8_t* rr_state_data(){return stateOutput.data();}
int rr_state_load(uint32_t n){if(!initialized||!bound||n!=stateInput.size()){error="State length or identity binding missing";return 0;}if(!restore(stateInput))return 0;for(auto& c:checkpoints)c={};return 1;}
}
#include "debug.inc"
#include "memory.inc"

#include "render.inc"

#include "drive-debug.inc"
