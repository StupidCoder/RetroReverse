// Timer 1 in HBlank mode advances with the field clock, never with MMIO reads.
#include "../../platform/psx/browser/core/machine.cpp"
#include <cassert>
int main() {
 auto m = std::make_unique<Machine>();
 m->cpu.setPC(0x1000);
 m->write32(0x1f801114, 0x107);
 m->write32(0x1f801110, 0x1234);
 for(int i=0;i<20;i++) assert(m->read32(0x1f801110)==0x1234);
 m->run(250000);
 assert(m->read32(0x1f801110)==0x1234+263);
 m->write16(0x1f801114,0x107);
 assert(m->read32(0x1f801110)==0);
 m->write16(0x1f801110,0xffff);
 m->run(951);
 assert(m->read32(0x1f801110)==0);
 assert(m->fields==1);
 puts("PASS: Timer 1 HBlank cadence, stable reads, count writes, mode reset and wrap");
}
