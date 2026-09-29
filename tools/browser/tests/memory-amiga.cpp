#include "../../platform/amiga/browser/core/api.cpp"
#include <cassert>
#include <iostream>
int main(){
 std::vector<uint8_t> rom(262144),disk(901120);std::memcpy(rr_firmware(rom.size()),rom.data(),rom.size());std::memcpy(rr_input(disk.size()),disk.data(),disk.size());assert(rr_init(disk.size()));
 rr_inspect_regions();rr_activity_begin(3);
 auto&m=*machine;m.cylinder=2;m.side=1;m.motor=m.selected=true;m.prepareTrack();
 assert(rr_activity_count()==0); // Host-side track encoding must not look like a disk read.
 m.trackIndex=166+3*544+32;m.diskPtr=0x1000;m.diskRemaining=1;m.dma=0x210;m.diskSync=true;m.diskClock=222;m.diskTick();rr_activity_end();
 assert(rr_activity_count()==3);auto&e=rrmem::events[0];assert(e.region==3&&e.kind==17&&e.offset==(5*11+3)*512);
 assert(rrmem::events[1].region==0&&rrmem::events[1].kind==10&&rrmem::events[1].offset==0x1000);
 rr_activity_begin(3);m.overlay=false;m.write8(0x100,0x77);assert(m.read8(0x100)==0x77);m.chipWrite(0x200,0xabcd);assert(m.chip16(0x200)==0xabcd);rr_activity_end();assert(rr_activity_count()==5);
 std::cout<<"PASS Amiga memory: CPU/DMA accesses, actual MFM payload attribution, no host-loading activity\n";
}
