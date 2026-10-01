#include "../state/fields.h"
#include <cassert>
#include <fstream>
#include <iostream>
#include <memory>
using namespace rr::c64;

static StateIdentity identity(){
 StateIdentity id;unsigned n=1;
 for(auto* digest:{&id.core,&id.basic,&id.kernal,&id.characters,&id.driveRom,&id.tape,&id.disk})for(auto& b:*digest)b=uint8_t(n++);
 id.configuration=7;return id;
}
static void setup(System& s){
 s.board.kernal[0x1ffc]=0;s.board.kernal[0x1ffd]=2;s.drive.rom[0x3ffc]=0;s.drive.rom[0x3ffd]=2;
 std::vector<uint8_t> disk(174848);for(size_t i=0;i<disk.size();i++)disk[i]=uint8_t(i*13+(i>>8));assert(s.drive.mount(disk,false));
 std::vector<uint8_t> tap(2020,2);std::copy_n("C64-TAPE-RAW",12,tap.begin());std::fill(tap.begin()+12,tap.begin()+20,0);tap[16]=0xd0;tap[17]=7;assert(s.board.loadTape(tap));s.power();
 const uint8_t host[]={0xe6,0x20,0xa5,0x20,0x8d,0x00,0xdd,0x4c,0x00,0x02};
 const uint8_t drive[]={0xe6,0x20,0xa5,0x20,0x8d,0x00,0x18,0x4c,0x00,0x02};
 std::copy(std::begin(host),std::end(host),s.board.state.ram.begin()+0x200);std::copy(std::begin(drive),std::end(drive),s.drive.state.ram.begin()+0x200);
 s.board.state.cia2.ddra=0x38;s.drive.state.serial.ddrb=0x1a;s.drive.state.disk.write(2,0x6f);s.drive.state.disk.write(0,4);
 s.board.state.ddr=0x2f;s.board.state.data=7;s.board.play(true);s.board.key(2,3,true);s.board.state.joy1=16;
 s.board.state.vic.write(0x11,0x1b);s.board.state.vic.write(0x16,8);s.board.state.vic.write(0x18,0x14);s.board.state.vic.write(0x21,6);
 s.board.state.sid.write(14,0xfd);s.board.state.sid.write(15,0x87);s.board.state.sid.write(18,0x81);
 s.board.state.cia1.write(4,17);s.board.state.cia1.write(5,0);s.board.state.cia1.write(14,0x11);
}
static uint32_t crc(std::span<const uint8_t> bytes){uint32_t c=~0u;for(auto b:bytes){c^=b;for(unsigned i=0;i<8;i++)c=(c>>1)^((c&1)?0xedb88320u:0);}return ~c;}
// Deliberately bypass the production validator to exercise a corrupt but
// checksum-valid input, including vector bounds and invalid enum values.
static std::vector<uint8_t> unchecked(SystemSnapshot& s,const StateIdentity& id){
 Archive a;a.header(0x4f343643,1);uint8_t kind=1;auto copy=id;a(kind,copy,s);auto sum=crc(a.bytes);a(sum);return a.bytes;
}
static void reject(System& s,const std::vector<uint8_t>& bytes,const StateIdentity& id){
 const auto before=saveState(s,id);std::string error;assert(!loadState(s,bytes,id,error)&&!error.empty());assert(saveState(s,id)==before);
}
int main(int argc,char** argv){
 const auto id=identity();auto machine=std::make_unique<System>();auto& s=*machine;setup(s);std::string error;
 uint64_t digest=1469598103934665603ull;unsigned cases=0;
 for(unsigned edges:{0,1,2,7,8,13,14,19,37,66,67,1001,40000}){
  setup(s);for(unsigned i=0;i<edges;i++)assert(s.tickEdge());
  const auto saved=saveState(s,id);for(auto b:saved)digest=(digest^b)*1099511628211ull;
  assert(s.run(1003));const auto end=saveState(s,id);assert(loadState(s,saved,id,error)&&error.empty());assert(saveState(s,id)==saved);
  assert(s.run(1003));assert(saveState(s,id)==end);++cases;
 }
 // Round-trip a dirty disk during writing, not merely a read-only source image.
 s.drive.state.mediaChange=0;s.drive.state.disk.write(12,0xc0);s.drive.state.disk.write(1,0xa5);assert(s.run(1000));assert(s.drive.state.media.changed());
 const auto dirty=saveState(s,id);assert(s.run(2000));const auto written=saveState(s,id);assert(loadState(s,dirty,id,error));assert(s.run(2000));assert(saveState(s,id)==written);
 // No C++ object layouts are used: the WASM test imports the bytes generated
 // by the native executable and resumes the identical in-flight disk writer.
 if(argc==3){
  if(std::string(argv[1])=="--write"){std::ofstream f(argv[2],std::ios::binary);f.write(reinterpret_cast<const char*>(dirty.data()),dirty.size());assert(f);}
  else{assert(std::string(argv[1])=="--read");std::ifstream f(argv[2],std::ios::binary);assert(f);std::vector<uint8_t> native((std::istreambuf_iterator<char>(f)),{});assert(native==dirty);assert(loadState(s,native,id,error));assert(s.run(2000));assert(saveState(s,id)==written);}
 }
 // Board-only transport shares the exact hardware fields; no fake drive is
 // needed to resume a tape lesson. Cross-kind loads are rejected atomically.
 const auto board=saveState(s.board,id);assert(s.board.run(33));const auto boardEnd=saveState(s.board,id);assert(loadState(s.board,board,id,error));assert(s.board.run(33));assert(saveState(s.board,id)==boardEnd);
 reject(s,board,id);const auto beforeBoard=saveState(s.board,id);assert(!loadState(s.board,dirty,id,error));assert(saveState(s.board,id)==beforeBoard);
 assert(loadState(s,dirty,id,error));
 auto mismatch=id;
 for(auto* value:{&mismatch.core,&mismatch.basic,&mismatch.kernal,&mismatch.characters,&mismatch.driveRom,&mismatch.tape,&mismatch.disk}){(*value)[0]^=1;reject(s,dirty,mismatch);(*value)[0]^=1;}
 mismatch.configuration^=1;reject(s,dirty,mismatch);
 reject(s,{},id);reject(s,std::vector<uint8_t>(dirty.begin(),dirty.end()-1),id);
 auto broken=dirty;broken[broken.size()/2]^=128;reject(s,broken,id);broken=dirty;broken.push_back(0);reject(s,broken,id);
 const auto renew=[&](){const auto sum=crc(std::span(broken).first(broken.size()-4));for(unsigned i=0;i<4;i++)broken[broken.size()-4+i]=uint8_t(sum>>(8*i));};
 broken=dirty;broken[8]=2;renew();reject(s,broken,id); // Unsupported version, with a valid checksum.
 broken=dirty;broken.erase(broken.end()-5);renew();reject(s,broken,id); // Truncated body, not merely damaged checksum.
 auto state=std::make_unique<SystemSnapshot>(s.save());
 const auto malformed=[&](auto change){*state=s.save();change(*state);reject(s,unchecked(*state,id),id);};
 malformed([](auto& x){x.board.cpu.stage=Stage(255);});malformed([](auto& x){x.drive.cpu.mode=Mode(255);});
 malformed([](auto& x){x.board.vic.raster=312;});malformed([](auto& x){x.board.vic.sprites[0].pixel=25;});
 malformed([](auto& x){x.board.sid.voice[0].phase=0x1000000;});malformed([](auto& x){x.board.cia1.a.queue=4;});
 malformed([](auto& x){x.board.tape.pulse=0xffffffff;});malformed([](auto& x){x.board.tape.remaining=0xffffffff;});
 malformed([](auto& x){x.drive.halfTrack=1;});malformed([](auto& x){x.drive.lastHalf=86;});
 malformed([](auto& x){x.drive.media.speeds[0][0]=4;});malformed([](auto& x){x.drive.media.trackCount=43;});
 malformed([](auto& x){x.drive.serial.shiftBits=8;});malformed([](auto& x){x.iec.hostSampleCycle=x.board.cycles+2;});
 // Allocation abuse: replace the first track count in the serialized payload,
 // then renew the checksum. Locate it with an independently serialized prefix.
 Archive prefix;prefix.header(0x4f343643,1);uint8_t kind=1;auto key=id;prefix(kind,key,state->board,state->drive.cpu,state->drive.lastBus,state->drive.serial,state->drive.disk,state->drive.ram);
 broken=dirty;const auto at=prefix.bytes.size();for(unsigned i=0;i<4;i++)broken[at+i]=255;
 renew();reject(s,broken,id);
 // G64 can contain actual half-track data and a speed change within a track.
 // Preserve both instead of serializing just a flattened D64 sector image.
 setup(s);s.drive.state.media.tracks[1]={0xa5,0x33,0x78,0xe1};s.drive.state.media.speeds[1]={0,1,2,3};
 std::vector<uint8_t> g64;assert(s.drive.state.media.exportG64(g64));assert(s.drive.mount(g64,false));
 assert(s.drive.state.media.g64&&s.drive.state.media.speeds[1][3]==3);
 s.drive.state.halfTrack=s.drive.state.lastHalf=3;s.drive.state.mediaChange=0;
 for(unsigned i=0;i<17;i++)assert(s.tickEdge());const auto raw=saveState(s,id);assert(s.run(500));const auto rawEnd=saveState(s,id);
 s.drive.eject();assert(loadState(s,raw,id,error));assert(s.run(500));assert(saveState(s,id)==rawEnd);
 std::cout<<"PASS portable state: "<<cases<<" dual-clock phases, tape/VIC/SID replay, dirty disk continuation, board/system isolation, identity/corruption/semantic rejection, digest="<<digest<<" bytes="<<dirty.size()<<'\n';
}
