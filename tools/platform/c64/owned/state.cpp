#include "state/fields.h"
#include <algorithm>
#include <limits>
#include <memory>

namespace rr::c64 {
namespace {
constexpr size_t MaxBytes=12*1024*1024;
static_assert(std::is_nothrow_copy_assignable_v<BoardState>);
static_assert(std::is_nothrow_move_assignable_v<DriveState>);
static_assert(std::is_nothrow_copy_assignable_v<IecState>);
// Corruption check, not authentication. Prepared lessons additionally pin a
// SHA-256 of the entire checkpoint in the knowledge package.
uint32_t checksum(std::span<const uint8_t> bytes){
 static const auto table=[](){std::array<uint32_t,256> t{};for(uint32_t i=0;i<256;i++){auto c=i;for(unsigned b=0;b<8;b++)c=(c>>1)^((c&1)?0xedb88320u:0);t[i]=c;}return t;}();
 uint32_t c=0xffffffff;for(auto v:bytes)c=table[(c^v)&255]^(c>>8);return ~c;
}
void require(bool value,const char* message){if(!value)throw std::runtime_error(message);}
void valid(const CpuState& s){
 require(s.op<=Op::TYA&&s.mode<=Mode::ind&&s.stage<=Stage::Fault,"Invalid CPU microstate");
}
void valid(const Cia& s){
 for(const auto* t:{&s.a,&s.b})require(t->queue<=3&&t->loadDelay<=2,"Invalid CIA timer pipeline");
 require(s.bits<=8&&s.todDivider<=5&&s.flags<=31&&s.mask<=31,"Invalid CIA state");
}
void valid(const VicFetch& s){
 require(s.kind<=VicAccess::Sprite&&s.phase<=2&&s.color<=15,"Invalid VIC fetch");
 if(s.kind==VicAccess::Matrix||s.kind==VicAccess::Graphics)require(s.slot<40,"Invalid VIC cell");
 if(s.kind==VicAccess::Pointer||s.kind==VicAccess::Sprite)require(s.slot<8,"Invalid VIC sprite");
}
void valid(const BoardState& s,const Board& media){
 valid(s.cpu);valid(s.cia1);valid(s.cia2);
 require(s.cycles<std::numeric_limits<uint64_t>::max()/1000000-2,"C64 clock exceeds scheduler range");
 require(s.todPhase<985248,"Invalid TOD phase");
 const auto& t=s.tape;require(t.pulse<=media.pulses.size(),"Tape position outside attached medium");
 require(t.pulse==media.pulses.size()?t.remaining==0:t.remaining>0&&t.remaining<=media.pulses[t.pulse],"Invalid tape pulse phase");
 const auto& v=s.vic;
 require(v.raster<312&&v.compare<512&&v.vc<1024&&v.base<1024&&v.cycle<=63&&v.row<8&&v.index<=40&&v.fetchCount<=2,"Invalid VIC pipeline");
 require(v.flags<16&&v.mask<16,"Invalid VIC interrupt state");
 for(const auto& c:v.cells){valid(c.matrix);valid(c.graphics);require(c.color<16,"Invalid cell color");}
 for(const auto& p:v.sprites){for(const auto& f:p.bytes)valid(f);require(p.pixel<=24&&p.shift<=0xffffff,"Invalid sprite shifter");}
 for(const auto& f:v.fetches)valid(f);
 require(std::all_of(v.pixels.begin(),v.pixels.end(),[](auto p){return p<16;}),"Invalid framebuffer palette index");
 require(std::all_of(v.regs.begin()+0x20,v.regs.begin()+0x2f,[](auto p){return p<16;}),"Invalid VIC palette register");
 require(std::all_of(s.color.begin(),s.color.end(),[](auto p){return p<16;}),"Invalid color RAM");
 for(const auto& voice:s.sid.voice){
  require(voice.phase<=0xffffff&&voice.noise<=0x7fffff&&voice.stage<=Envelope::Release&&voice.rate<0x8000&&voice.noiseDelay<=2,"Invalid SID pipeline");
  require(voice.divider==1||voice.divider==2||voice.divider==4||voice.divider==8||voice.divider==16||voice.divider==30,"Invalid SID divider");
 }
}
void valid(const Via& s){require(s.shiftBits<8&&s.t1Delay<=1&&s.t2Delay<=1&&s.ifr<128&&s.ier<128,"Invalid VIA pipeline");}
void valid(const SystemSnapshot& s,const Board& media){
 valid(s.board,media);const auto& d=s.drive;valid(d.cpu);valid(d.serial);valid(d.disk);
 require(d.clocks<std::numeric_limits<uint64_t>::max()/1000000-2,"Drive clock exceeds scheduler range");
 require(d.halfTrack>=2&&d.halfTrack<=85&&d.lastHalf>=2&&d.lastHalf<=85&&d.phase<4&&d.mediaPhase<16&&d.bitPhase<16&&d.bits<8&&d.ones<=10,"Invalid drive mechanics");
 require(d.rotation<200000&&d.mediaChange<=200000&&d.device>=8&&d.device<=11,"Invalid drive state");
 const auto& disk=d.media;
 require(disk.g64?disk.trackCount<=42:(disk.trackCount==0||disk.trackCount==35||disk.trackCount==40),"Invalid disk geometry");
 for(unsigned i=0;i<84;i++){
  require(disk.tracks[i].size()<=65535&&disk.speeds[i].size()==disk.tracks[i].size(),"Invalid disk track size");
  require(i<disk.trackCount*2||disk.tracks[i].empty(),"Track outside disk geometry");
  require(std::all_of(disk.speeds[i].begin(),disk.speeds[i].end(),[](auto v){return v<4;}),"Invalid disk speed zone");
 }
 require(s.iec.hostPulls<8&&s.iec.drivePulls<8&&s.iec.last.actor<=1&&s.iec.last.changed<8&&s.iec.last.levels<8&&s.iec.last.hostPulls<8&&s.iec.last.drivePulls<8,"Invalid IEC state");
 // Samples may be from this or the pending edge, and must not become a future
 // sample that the scheduler silently reuses for an unrelated CPU clock.
 require(s.iec.hostSampleCycle<=s.board.cycles+1&&s.iec.driveSampleCycle<=d.clocks+1,"Invalid IEC sample phase");
}
void header(Archive& a,uint8_t kind,const StateIdentity& expected){
 a.header(0x4f343643,1);uint8_t storedKind=kind;a(storedKind);
 require(storedKind==kind,"Wrong checkpoint machine type");
 auto identity=expected;a(identity);require(identity==expected,"Checkpoint core, firmware, media or configuration mismatch");
}
template<class T>std::vector<uint8_t> encode(T& state,const Board& media,const StateIdentity& identity,uint8_t kind){
 valid(state,media);Archive a;header(a,kind,identity);a(state);
 require(a.bytes.size()+4<=MaxBytes,"Checkpoint too large");auto crc=checksum(a.bytes);a(crc);return std::move(a.bytes);
}
template<class T>std::unique_ptr<T> decode(std::span<const uint8_t> bytes,const Board& media,const StateIdentity& identity,uint8_t kind){
 require(bytes.size()>=4&&bytes.size()<=MaxBytes,"Invalid checkpoint size");
 const auto n=bytes.size()-4;uint32_t stored=0;for(unsigned i=0;i<4;i++)stored|=uint32_t(bytes[n+i])<<(8*i);
 require(stored==checksum(bytes.first(n)),"Checkpoint checksum mismatch");
 Archive a(bytes.data(),n);header(a,kind,identity);auto next=std::make_unique<T>();a(*next);a.finish();valid(*next,media);return next;
}
}
std::vector<uint8_t> saveState(const Board& board,const StateIdentity& identity){
 auto state=std::make_unique<BoardState>(board.state);return encode(*state,board,identity,0);
}
std::vector<uint8_t> saveState(const System& system,const StateIdentity& identity){
 auto state=std::make_unique<SystemSnapshot>(system.save());return encode(*state,system.board,identity,1);
}
bool loadState(Board& board,std::span<const uint8_t> bytes,const StateIdentity& identity,std::string& error){
 try{auto next=decode<BoardState>(bytes,board,identity,0);board.state=*next;error.clear();return true;}
 catch(const std::exception& e){error=e.what();return false;}
}
bool loadState(System& system,std::span<const uint8_t> bytes,const StateIdentity& identity,std::string& error){
 try{auto next=decode<SystemSnapshot>(bytes,system.board,identity,1);
  // Vector moves do not allocate: once validation succeeds, commit cannot
  // fail halfway through replacing the live disk's tracks.
  system.board.state=next->board;system.drive.state=std::move(next->drive);system.iec=next->iec;error.clear();return true;
 }catch(const std::exception& e){error=e.what();return false;}
}
}
