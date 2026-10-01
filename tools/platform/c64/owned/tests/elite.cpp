#include "../state.h"
#include <algorithm>
#include <cassert>
#include <fstream>
#include <iostream>
#include <memory>
using namespace rr::c64;
static std::vector<uint8_t> file(const char* path){std::ifstream f(path,std::ios::binary);assert(f);return {std::istreambuf_iterator<char>(f),{}};}
static void run(Board& b,unsigned clocks){assert(b.run(clocks));}
int main(int argc,char** argv){
 assert(argc==5);auto board=std::make_unique<Board>();auto& b=*board;
 const auto basic=file(argv[1]),kernal=file(argv[2]),chars=file(argv[3]),tape=file(argv[4]);
 assert(basic.size()==8192&&kernal.size()==8192&&chars.size()==4096);
 std::copy(basic.begin(),basic.end(),b.basic.begin());std::copy(kernal.begin(),kernal.end(),b.kernal.begin());std::copy(chars.begin(),chars.end(),b.chars.begin());assert(b.loadTape(tape));b.power();
 // Independent wire-format oracle for the FIRST block only. The documented
 // initial format is a start pulse + eight MSB-first pulses (384/744 cycles).
 // Later self-modification changes this format; it is not extrapolated here.
 std::array<uint8_t,57> decoded{};
 for(unsigned i=0;i<decoded.size();i++){
  const auto at=0xd7a5+i*9;assert(tape.at(at)==0x5d);
  for(unsigned bit=1;bit<=8;bit++){const auto p=tape.at(at+bit);assert(p==0x30||p==0x5d);decoded[i]=uint8_t((decoded[i]<<1)|(p==0x5d));}
 }
 const uint8_t header[]={0x16,0x03,0x34,0x03,0x00};assert(std::equal(std::begin(header),std::end(header),decoded.begin()));
 run(b,3000000);
 // Physical LOAD + RETURN; the tape's IRQ-vector overwrite autostarts itself.
 for(auto key:std::array<std::array<unsigned,2>,5>{{{5,2},{4,6},{1,2},{2,2},{0,1}}}){b.key(key[0],key[1],true);run(b,100000);b.key(key[0],key[1],false);run(b,100000);}
 b.play(true);bool entered=false;
 for(unsigned i=0;i<80000000;i++){if(b.state.cpu.stage==Stage::Fetch&&b.state.cpu.pc==0x378){entered=true;break;}run(b,1);}assert(entered);
 const auto entryCycle=b.state.cycles;const auto entryPulse=b.state.tape.pulse;
 const StateIdentity identity{};std::string error; // Harness pins the input hashes.
 const auto entry=saveState(b,identity);
 auto firstBlock=[&](){
  unsigned stores=0;uint32_t hash=2166136261;
  for(unsigned i=0;i<5000000&&stores<52;i++){
   run(b,1);const auto& bus=b.state.lastBus;
   if(bus.valid&&bus.write&&bus.pc==0x3b3){assert(bus.address==0x300+stores&&bus.value==decoded[5+stores]);hash=(hash^bus.value)*16777619;++stores;}
  }
  assert(stores==52);return hash;
 };
 const auto hash=firstBlock();const auto end=saveState(b,identity);const auto endCycle=b.state.cycles;const auto endPulse=b.state.tape.pulse;
 assert(loadState(b,entry,identity,error));assert(firstBlock()==hash&&saveState(b,identity)==end);
 // Stop at the asserted boundary. This does not claim a full Elite load or
 // game/object-slot compatibility merely because the first block now agrees.
 std::cout<<"PASS authentic Elite TAP first block: 52 stores $0300-$0333, wire oracle hash="<<hash<<"; entry cycle="<<entryCycle<<" pulse="<<entryPulse<<"; end cycle="<<endCycle<<" pulse="<<endPulse<<"; portable replay; later stages unvalidated\n";
}
