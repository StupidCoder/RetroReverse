#include "../../platform/c64/browser/core/core.cpp"
#include <cassert>
#include <iostream>
int main(){
 std::memset(rr_input(),0,20480);assert(rr_init(8192,8192,4096));
 // A TAP with one long pulse permits a prepared RAM-only test program.
 std::memset(rr_input(),0,24);std::memcpy(rr_input(),"C64-TAPE-RAW",12);rr_input()[12]=1;rr_input()[16]=4;rr_input()[21]=255;rr_input()[22]=255;assert(rr_tape(24));
 std::memset(rr_input(),0,65536);const uint8_t code[]={0xa9,0x42,0x8d,0x00,0x20,0xad,0x00,0x20,0x4c,0x00,0x10};std::memcpy(rr_input()+0x1000,code,sizeof(code));assert(rr_prepare(0x1000,0));
 rr_inspect_regions();rr_activity_begin(3);assert(rr_run(200,0,0)==200);rr_activity_end();
 bool read=false,write=false;for(auto&e:rrmem::events)if(e.region==0&&e.offset==0x2000){assert(e.value==0x42);read|=e.kind==1;write|=e.kind==2;}
 assert(read&&write&&rr_ram()[0x2000]==0x42);
 const auto count=rr_activity_count();const auto cycle=ctx.cycles;rr_inspect_regions();rr_inspect_data(0);assert(count==rr_activity_count()&&cycle==ctx.cycles);
 rr_activity_begin(3);observation::write(1,0xd811,0xab);assert(rrmem::events.back().value==0xab);rr_activity_end();
 rr_activity_begin(3);ctx.remaining=1;machine.cas_port&=~C64_CASPORT_MOTOR;rr_play(1);rr_run(1,0,0);rr_activity_end();bool pulse=false;for(auto&e:rrmem::events)pulse|=e.kind==17&&e.region==5&&e.offset==0;assert(pulse);
 rr_activity_begin(3);for(size_t i=0;i<rrmem::limit+3;i++)rrmem::access(i,0,0,0,1,2,0,0);assert(rr_activity_count()==rrmem::limit&&rr_activity_dropped()==3);rr_activity_end();
 std::cout<<"PASS C64 memory: actual CPU reads and same-value writes, immutable peeks, bounded overflow\n";
}
