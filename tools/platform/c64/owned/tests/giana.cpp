#include "drive-support.h"
#include "../debugger.h"
#include <deque>
static uint32_t hash(std::span<const uint8_t> bytes){uint32_t h=2166136261;for(auto v:bytes)h=(h^v)*16777619;return h;}
static uint64_t eventHash(const Debugger& d){uint64_t h=1469598103934665603ull;for(const auto& e:d.events())for(uint64_t v:{e.time,uint64_t(e.actor),uint64_t(e.levels),uint64_t(e.hostPC),uint64_t(e.drivePC),uint64_t(e.halfTrack),uint64_t(e.bit)})h=(h^v)*1099511628211ull;return h;}
int main(int argc,char** argv){
 assert(argc==6);auto machine=std::make_unique<System>();auto& s=*machine;
 load(argv[1],s.board.basic);load(argv[2],s.board.kernal);load(argv[3],s.board.chars);load(argv[4],s.drive.rom);
 std::ifstream f(argv[5],std::ios::binary);assert(f);std::vector<uint8_t> disk((std::istreambuf_iterator<char>(f)),{});assert(s.drive.mount(disk));s.power();run(s,3000000);type(s,"LOAD\"LADER\",8,1\r");
 Debugger d(s,1024);d.breakpoint(Processor::Drive,0x046c);d.resume();auto stop=d.run(40000000);
 assert(stop.reason==StopReason::Breakpoint&&stop.processor==Processor::Drive&&stop.address==0x046c);
 // The game's M-W commands uploaded these bytes through real DOS/IEC; no host
 // writes to emulated RAM. Compare the source buffer with the drive destination.
 assert(std::equal(s.drive.state.ram.begin()+0x400,s.drive.state.ram.begin()+0x600,s.board.state.ram.begin()+0xcdee));
 assert(d.disassemble(Processor::Drive,0x046c).text=="LDA #$C8");const auto uploadHash=hash(std::span(s.drive.state.ram).subspan(0x400,512));
 const auto paused=std::make_unique<SystemSnapshot>(s.save());assert(d.run(10000).edges==0&&s.drive.state.clocks==paused->drive.clocks&&s.board.state.cycles==paused->board.cycles&&s.drive.state.position==paused->drive.position);
 const auto retired=s.drive.state.cpu.retired;assert(d.step(Processor::Drive).reason==StopReason::Step&&s.drive.state.cpu.a==0xc8&&s.drive.state.cpu.retired==retired+1&&s.board.state.cycles>paused->board.cycles);
 d.breakpoint(Processor::Drive,0x046c,false);d.breakpoint(Processor::Drive,0x04df);d.resume();stop=d.run(2000000);assert(stop.reason==StopReason::Breakpoint&&stop.address==0x04df);
 assert(d.disassemble(Processor::Drive,0x04df).text=="TAX");assert(d.disassemble(Processor::Drive,0x04e0).text=="BIT $1800");
 // A repeatable investigation: stop before a byte is sent, inspect/step its
 // sender, capture both machines' IEC transitions, restore and reproduce them.
 const auto byteStart=std::make_unique<SystemSnapshot>(s.save());d.breakpoint(Processor::Drive,0x04df,false);d.clearEvents();d.resume();assert(d.run(1000).reason==StopReason::Budget);const auto trace=eventHash(d);const auto ram=hash(s.board.state.ram);const auto endDrive=hash(s.drive.state.ram);assert(!d.events().empty()&&!d.droppedEvents());
 std::cout<<"INSPECT drive $04DF: "<<d.disassemble(Processor::Drive,0x04df).text<<"; $04E0: "<<d.disassemble(Processor::Drive,0x04e0).text<<"; half-track="<<unsigned(byteStart->drive.halfTrack)<<" bit="<<byteStart->drive.position<<'\n';
 unsigned shown=0;for(const auto& e:d.events()){if(shown++==8)break;std::cout<<"IEC time="<<e.time<<" actor="<<unsigned(e.actor)<<" changed="<<unsigned(e.changed)<<" levels="<<unsigned(e.levels)<<" hostPC="<<std::hex<<e.hostPC<<" drivePC="<<e.drivePC<<std::dec<<" half-track="<<unsigned(e.halfTrack)<<" bit="<<e.bit<<'\n';}
 d.restore(*byteStart);d.resume();assert(d.run(1000).reason==StopReason::Budget);assert(eventHash(d)==trace&&hash(s.board.state.ram)==ram&&hash(s.drive.state.ram)==endDrive);
 d.restore(*byteStart);d.clearEvents();d.resume();
 std::deque<uint8_t> pending;unsigned sent=0,received=0,stores=0;bool loaded=false;
 for(uint64_t edge=0;edge<40000000;edge++){
  const auto p=s.nextProcessor();const auto& cpu=d.cpu(p);
  if(cpu.stage==Stage::Fetch){
   // $04DF TAX starts the drive's sender. $CDED RTS ends the four-sample
   // receiver. Their values must match, including length/address framing.
   if(p==Processor::Drive&&cpu.pc==0x04df){pending.push_back(cpu.a);++sent;}
   if(p==Processor::C64&&cpu.pc==0xcded){assert(!pending.empty()&&pending.front()==cpu.a);pending.pop_front();++received;}
   if(p==Processor::C64&&cpu.pc==0xcd9f){loaded=true;break;}
  }
  assert(d.run(1).reason==StopReason::Budget);
  if(p==Processor::C64){const auto& b=s.board.state.lastBus;if(b.valid&&b.write&&(b.pc==0xcd08||b.pc==0xcd75))++stores;}
 }
 assert(loaded&&pending.empty()&&sent==20168&&received==20168);
 const auto payload=hash(std::span(s.board.state.ram).subspan(0x0801,20086));assert(payload==1795265874u);assert(stores==20086);assert(!s.drive.state.media.changed());
 // Reaching the authentic unpacker confirms the caller completes its startup
 // sequence after the transfer. Decompression/gameplay are a separate claim.
 d.breakpoint(Processor::C64,0x0810);d.resume();stop=d.run(4000000);assert(stop.reason==StopReason::Breakpoint&&stop.processor==Processor::C64&&stop.address==0x0810&&d.peek(Processor::C64,0x0810)==0x78);
 std::cout<<"PASS Giana G64 custom loader: uploaded 512 bytes ("<<uploadHash<<"); 20168 IEC bytes; 20086 payload stores/hash="<<payload<<"; unpacker entry; drive breakpoint/step and replay="<<trace<<'\n';
}
