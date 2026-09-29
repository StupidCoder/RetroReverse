#include "../../platform/gamegear/browser/core/api.cpp"
#include <cassert>
#include <iostream>
int main(){
 auto*p=rr_input(65536);for(unsigned i=0;i<65536;i++)p[i]=i/16384;
 assert(rr_init(65536));rr_inspect_regions();assert(rrmem::regions.size()==7);
 rr_activity_begin(3);
 // The timing estimator and inspector are not CPU bus transactions.
 gamegear_Machine_Read(machine,0x4000);rr_inspect_regions();assert(rr_activity_count()==0);
 assert(z80_CPU_read(machine->CPU,0x4000)==1);
 gamegear_Machine_Write(machine,0xfffe,3);
 assert(z80_CPU_read(machine->CPU,0x4000)==3);
 assert(rrmem::events[0].region==4&&rrmem::events[0].offset==0);
 assert(rrmem::events[1].region==0&&rrmem::events[1].kind==2);
 assert(rrmem::events[2].region==6&&rrmem::events[2].offset==0);
 gamegear_Machine_Write(machine,0xfffd,2);
 assert(z80_CPU_read(machine->CPU,0)==0); // First 1 KiB is fixed, even with slot 0 switched.
 assert(z80_CPU_read(machine->CPU,0x400)==2);
 rr_activity_end();const auto count=rr_activity_count();rr_inspect_regions();assert(count==rr_activity_count());
 assert(rrmem::regions[3].mappings.find("1024")!=std::string::npos);
 std::cout<<"PASS GG memory: physical bank attribution, fixed window, mapper writes, no timing-probe reads\n";
}
