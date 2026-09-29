static void rrMemoryRegions(){
 using namespace rrmem;regions.clear();supported=true;
 const auto b=machine.cpu_port;
 std::string maps="[{\"cpu\":2,\"offset\":2,\"size\":40958},{\"cpu\":49152,\"offset\":49152,\"size\":4096}";
 auto map=[&](unsigned a,unsigned n){maps+=",{\"cpu\":"+std::to_string(a)+",\"offset\":"+std::to_string(a)+",\"size\":"+std::to_string(n)+"}";};
 if((b&3)!=3)map(0xa000,8192);if(!(b&3))map(0xd000,4096);if(!(b&2))map(0xe000,8192);maps+="]";
 add("ram","Underlying RAM","ram",machine.ram,65536,"[]",0,maps);
 add("basic","BASIC ROM","rom",machine.rom_basic,8192,(b&3)==3?"[40960]":"[]",0xa000);
 add("kernal","KERNAL ROM","rom",machine.rom_kernal,8192,b&2?"[57344]":"[]",0xe000);
 add("chars","Character ROM","rom",machine.rom_char,4096,(b&3)&&!(b&4)?"[53248]":"[]",0xd000);
 add("color","Color RAM","ram",machine.color_ram,1024,machine.io_mapped?"[55296]":"[]",0xd800);
}
#include "../../../../browser/core/memory-api.inc"
