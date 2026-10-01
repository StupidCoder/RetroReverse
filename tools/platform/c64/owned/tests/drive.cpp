#include "drive-support.h"
#include <cassert>
#include <fstream>
#include <iostream>
#include <memory>
#include <string>
using namespace rr::c64;
static std::vector<uint8_t> diskImage(){
 std::vector<uint8_t> image(174848);const auto bam=Disk::offset(18,0),dir=Disk::offset(18,1);
 image[bam]=18;image[bam+1]=1;image[bam+2]=0x41;
 for(unsigned t=1;t<=35;t++){image[bam+t*4]=uint8_t(Disk::sectors(t));for(unsigned s=0;s<Disk::sectors(t);s++)image[bam+t*4+1+s/8]|=uint8_t(1<<(s&7));}
 auto use=[&](unsigned t,unsigned s){--image[bam+t*4];image[bam+t*4+1+s/8]&=uint8_t(~(1<<(s&7)));};use(18,0);use(18,1);
 std::fill(image.begin()+bam+0x90,image.begin()+bam+0xab,0xa0);const std::string title="C4 TEST";std::copy(title.begin(),title.end(),image.begin()+bam+0x90);image[bam+0xa2]='0';image[bam+0xa3]='1';image[bam+0xa5]='2';image[bam+0xa6]='A';
 image[dir+1]=255;image[dir+2]=0x82;image[dir+3]=1;image[dir+4]=0;std::fill(image.begin()+dir+5,image.begin()+dir+21,0xa0);const std::string name="TEST";std::copy(name.begin(),name.end(),image.begin()+dir+5);image[dir+30]=6;
 const unsigned tracks[]={1,17,19,25,31,35};std::vector<uint8_t> file{0,0x20};for(unsigned i=0;i<6*254-2;i++)file.push_back(uint8_t((i*73)^(i>>3)));
 for(unsigned i=0;i<6;i++){const auto off=Disk::offset(tracks[i],0);use(tracks[i],0);image[off]=i==5?0:uint8_t(tracks[i+1]);image[off+1]=i==5?255:0;std::copy(file.begin()+i*254,file.begin()+(i+1)*254,image.begin()+off+2);}
 return image;
}
static uint64_t trace(System& s,unsigned clocks){uint64_t hash=1469598103934665603ull;while(clocks--){assert(s.tick());for(uint64_t v:{uint64_t(s.board.state.cpu.pc),uint64_t(s.drive.state.cpu.pc),uint64_t(s.drive.state.position),s.iec.transitions,uint64_t(s.iec.last.levels),uint64_t(s.drive.state.bits)})hash=(hash^v)*1099511628211ull;}return hash;}
static void replay(System& s){auto before=std::make_unique<SystemSnapshot>(s.save());const auto hash=trace(s,20000);auto after=std::make_unique<SystemSnapshot>(s.save());s.restore(*before);assert(trace(s,20000)==hash);assert(s.board.state.ram==after->board.ram&&s.drive.state.ram==after->drive.ram&&s.drive.state.media.tracks==after->drive.media.tracks&&s.drive.state.media.dirty==after->drive.media.dirty);std::cout<<"PASS in-flight dual-machine replay: "<<hash<<'\n';}
static void boot(int argc,char** argv){assert(argc==5||argc==6);auto machine=std::make_unique<System>();auto& s=*machine;
 load(argv[1],s.board.basic);load(argv[2],s.board.kernal);load(argv[3],s.board.chars);load(argv[4],s.drive.rom);assert(s.drive.mount(diskImage(),false));s.power();run(s,3000000);
 type(s,"LOAD\"$\",8\r");run(s,5000000);
 type(s,"LIST\r");run(s,500000);
 type(s,"LOAD\"TEST\",8,1\r");run(s,1000000);replay(s);run(s,10000000);
 for(unsigned i=0;i<6*254-2;i++)if(s.board.state.ram[0x2000+i]!=uint8_t((i*73)^(i>>3))){std::cerr<<"payload mismatch at "<<i<<'\n';std::abort();}
 assert(screen(s).find("C4 TEST")!=std::string::npos);assert(s.drive.state.steps>90);
 std::cout<<"PASS real-ROM IEC file load\n";
 type(s,"NEW\r10 PRINT 42\rSAVE\"NEW\",8\r");
 for(unsigned i=0;i<5000000&&!s.drive.state.writing;i++)run(s,1);assert(s.drive.state.writing);replay(s);run(s,8000000);
 std::vector<uint8_t> saved;assert(s.drive.state.media.changed());assert(s.drive.state.media.exportD64(saved));
 const std::vector<uint8_t> program(s.board.state.ram.begin()+0x801,s.board.state.ram.begin()+0x80c);
 type(s,"NEW\rLOAD\"NEW\",8\r");run(s,6000000);type(s,"LIST\r");run(s,500000);assert(screen(s).find("10 PRINT 42")!=std::string::npos);assert(std::equal(program.begin(),program.end(),s.board.state.ram.begin()+0x801));
 std::cout<<"PASS real-ROM IEC save and reload\n";
 assert(s.drive.mount(diskImage(),true));type(s,"SAVE\"BLOCK\",8\r");run(s,6000000);assert(!s.drive.state.media.changed());assert(s.drive.state.media.exportD64(saved)&&saved==diskImage());
 std::cout<<"PASS real-ROM write-protected SAVE leaves media unchanged\n";
 if(argc==6){std::ifstream f(argv[5],std::ios::binary);std::vector<uint8_t> image((std::istreambuf_iterator<char>(f)),{});assert(s.drive.mount(image));
  type(s,"NEW\rLOAD\"$\",8\r");run(s,6000000);type(s,"LIST\r");run(s,1000000);
  assert(screen(s).find("GIANA-GAME")!=std::string::npos);assert(screen(s).find("LADER")!=std::string::npos);
  type(s,"LOAD\"F\",8,1\r");run(s,6000000);
  uint32_t hash=2166136261u;for(unsigned i=0;i<992;i++)hash=(hash^s.board.state.ram[0xcc00+i])*16777619u;
  assert(hash==735537004u);assert(!s.drive.state.media.changed());std::cout<<"PASS Giana G64 directory/media change and ROM-driven F payload: "<<hash<<'\n';
 }
}
int main(int argc,char** argv){if(argc>1)boot(argc,argv);}
