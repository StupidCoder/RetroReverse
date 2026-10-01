static void rrMemoryRegions(){
 using namespace rrmem;regions.clear();supported=true;const auto p=board.port();
 std::string maps="[{\"cpu\":2,\"offset\":2,\"size\":40958},{\"cpu\":49152,\"offset\":49152,\"size\":4096}";
 auto map=[&](unsigned a,unsigned n){maps+=",{\"cpu\":"+std::to_string(a)+",\"offset\":"+std::to_string(a)+",\"size\":"+std::to_string(n)+"}";};
 if((p&3)!=3)map(0xa000,8192);if(!(p&3))map(0xd000,4096);if(!(p&2))map(0xe000,8192);maps+="]";
 add("ram","Underlying RAM","ram",board.state.ram.data(),65536,"[]",0,maps);
 add("basic","BASIC ROM","rom",board.basic.data(),8192,(p&3)==3?"[40960]":"[]",0xa000);
 add("kernal","KERNAL ROM","rom",board.kernal.data(),8192,p&2?"[57344]":"[]",0xe000);
 add("chars","Character ROM","rom",board.chars.data(),4096,(p&3)&&!(p&4)?"[53248]":"[]",0xd000);
 add("color","Color RAM","ram",board.state.color.data(),1024,(p&3)&&(p&4)?"[55296]":"[]",0xd800);
 // Index 5 remains reserved for the worker's decoded tape-pulse region.
}
#include "../../../../browser/core/memory-api.inc"
