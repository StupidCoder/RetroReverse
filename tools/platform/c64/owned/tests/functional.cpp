#include "../cpu.h"
#include <array>
#include <cassert>
#include <fstream>
#include <iostream>
using namespace rr::c64;
// Optional external Klaus Dormann NMOS functional binary; see COMPATIBILITY.md.
int main(int argc,char** argv){
 assert(argc==2);std::array<uint8_t,65536> ram{};
 std::ifstream file(argv[1],std::ios::binary);assert(file);
 file.read(reinterpret_cast<char*>(ram.data()),ram.size());assert(file.gcount()==65536&&file.peek()==EOF);
 Cpu cpu;cpu.start(0x0400);
 while(cpu.state.clocks<200000000){
  if(cpu.boundary()&&cpu.state.pc==0x3469){std::cout<<"PASS Klaus Dormann functional test; cycles="<<cpu.state.clocks<<" instructions="<<cpu.state.retired<<'\n';return 0;}
  const auto bus=cpu.bus();if(bus.write)ram[bus.address]=bus.data;cpu.tick(ram[bus.address]);assert(!cpu.faulted());
 }
 std::cerr<<"Functional test exceeded budget at PC="<<std::hex<<cpu.state.pc<<'\n';return 1;
}
