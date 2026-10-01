#include "../state.h"
#include <cassert>
#include <fstream>
#include <iostream>
#include <memory>
#include <string>
using namespace rr::c64;
static std::vector<uint8_t> file(const char* path){std::ifstream f(path,std::ios::binary);assert(f);return {std::istreambuf_iterator<char>(f),{}};}
static uint32_t hash(std::span<const uint8_t> bytes){uint32_t h=2166136261;for(auto v:bytes)h=(h^v)*16777619;return h;}
static void status(const Board& b,const char* name){std::cerr<<name<<" cycles="<<b.state.cycles<<" pc="<<std::hex<<b.state.cpu.pc<<" AB="<<unsigned(b.state.ram[0xab])<<" mode="<<unsigned(b.state.ram[0x9d])<<std::dec<<" pulse="<<b.state.tape.pulse<<" motor="<<b.motor()<<'\n';}
static void run(Board& b,uint64_t n){if(!b.run(n)){status(b,"FAULT");std::abort();}}
static void type(Board& b,const char* text){for(;*text;text++){
 unsigned col=0,row=0;switch(*text){case 'L':col=5;row=2;break;case 'O':col=4;row=6;break;case 'A':col=1;row=2;break;case 'D':col=2;row=2;break;case 'R':col=2;row=1;break;case 'U':col=3;row=6;break;case 'N':col=4;row=7;break;case '\r':col=0;row=1;break;default:std::abort();}
 b.key(col,row,true);run(b,100000);b.key(col,row,false);run(b,100000);
}}
static void image(const Board& b,const std::string& name){
 // Fixed presentation palette, not an analog PAL color simulation. Crop borders.
 const uint32_t palette[]={0x000000,0xffffff,0x813338,0x75cec8,0x8e3c97,0x56ac4d,0x2e2c9b,0xedf171,0x8e5029,0x553800,0xc46c71,0x4a4a4a,0x7b7b7b,0xa9ff9f,0x706deb,0xb2b2b2};
 std::ofstream f(name,std::ios::binary);f<<"P6\n392 272\n255\n";
 for(unsigned y=16;y<288;y++)for(unsigned x=0;x<392;x++){const auto c=palette[b.state.vic.pixels[y*504+(x+496)%504]&15];const char rgb[]={char(c>>16),char(c>>8),char(c)};f.write(rgb,3);}
}
int main(int argc,char** argv){
 assert(argc==9||argc==10);auto board=std::make_unique<Board>();auto& b=*board;
 const StateIdentity identity{};std::string error; // Test host: input hashes are pinned by check.py.
 const auto basic=file(argv[1]),kernal=file(argv[2]),chars=file(argv[3]),tape=file(argv[4]),expected=file(argv[5]),pages=file(argv[6]),graphics=file(argv[7]),graphicsMask=file(argv[8]);
 assert(basic.size()==8192&&kernal.size()==8192&&chars.size()==4096&&expected.size()==65536);
 std::copy(basic.begin(),basic.end(),b.basic.begin());std::copy(kernal.begin(),kernal.end(),b.kernal.begin());std::copy(chars.begin(),chars.end(),b.chars.begin());assert(b.loadTape(tape));b.power();
 run(b,3000000);type(b,"LOAD\r");b.play(true);
 for(unsigned slice=0;slice<10000;slice++){run(b,10000);if(b.state.tape.pulse>=48233&&!b.motor())break;}
 status(b,"basic-loaded");assert(b.state.tape.pulse>=48233&&!b.motor());run(b,200000);type(b,"RUN\r");
 unsigned stores=0,oscReads=0;uint32_t noiseHash=2166136261;bool entry=false,loading=false;
 for(uint64_t wait=0;wait<150000000;wait++){
  const auto bus=b.state.cpu.bus;
  if(b.state.cpu.stage==Stage::Fetch&&bus.address==0x8600){entry=true;break;}
  // Observe authentic Novaload writes without altering the machine.
  run(b,1);const auto& access=b.state.lastBus;
  if(access.valid&&access.write&&access.pc==0x38c){const auto wanted=uint16_t((pages.at(stores/256)<<8)|(stores&255));assert(access.address==wanted&&access.value==expected[wanted]);++stores;}
  if(!loading&&stores>=11*256){loading=true;status(b,"loading");
   const auto checkpoint=saveState(b,identity);run(b,50000);const auto ram=hash(b.state.ram),frame=hash(b.state.vic.pixels);const auto phase=b.state.sid.voice[2].phase,noise=b.state.sid.voice[2].noise;
   assert(loadState(b,checkpoint,identity,error));run(b,50000);assert(hash(b.state.ram)==ram&&hash(b.state.vic.pixels)==frame&&b.state.sid.voice[2].phase==phase&&b.state.sid.voice[2].noise==noise);assert(loadState(b,checkpoint,identity,error));
if(argc==10)image(b,std::string(argv[9])+"-loading.ppm");}
 }
 status(b,"entry");assert(entry&&stores==pages.size()*256&&b.state.ram[0xab]==0);
 for(unsigned a=0x7000;a<0xb900;a++)assert(b.state.ram[a]==expected[a]);
 run(b,3000000);status(b,"title");assert(b.state.ram[0x9d]==1);if(argc==10)image(b,std::string(argv[9])+"-title.ppm");
 b.state.joy2=16;run(b,400000);b.state.joy2=0;for(unsigned slice=0;slice<1500&&b.state.ram[0x9d]!=2;slice++)run(b,10000);assert(b.state.ram[0x9d]==2);
 assert(graphics.size()==65536&&graphicsMask.size()==65536);unsigned checkedGraphics=0;
 for(unsigned a=0;a<65536;a++)if(graphicsMask[a]){assert(b.state.ram[a]==graphics[a]);++checkedGraphics;}assert(checkedGraphics==2251);
 const auto saved=saveState(b,identity);
 auto play=[&](){oscReads=0;noiseHash=2166136261;for(unsigned i=0;i<2000000;i++){
  run(b,1);const auto& bus=b.state.lastBus;if(bus.valid&&bus.ready&&!bus.write&&bus.address==0xd41b){++oscReads;noiseHash=(noiseHash^bus.value)*16777619;}
 }};
 play();status(b,"gameplay");const auto ramHash=hash(b.state.ram),frameHash=hash(b.state.vic.pixels),sidHash=noiseHash,reads=oscReads;assert(reads>100);
 if(argc==10)image(b,std::string(argv[9])+"-gameplay.ppm");assert(loadState(b,saved,identity,error));play();assert(hash(b.state.ram)==ramHash&&hash(b.state.vic.pixels)==frameHash&&noiseHash==sidHash&&oscReads==reads);
 std::cout<<"PASS authentic Fort TAP boot: "<<stores<<" payload stores; 2251 extracted graphics bytes; gameplay and portable replay; RAM="<<ramHash<<" frame="<<frameHash<<" OSC3 reads="<<reads<<" digest="<<sidHash<<'\n';
}
