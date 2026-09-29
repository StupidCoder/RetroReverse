#pragma once
#include "../../../../browser/core/memory.h"
static void rrMemoryRegions(){
 using namespace rrmem;regions.clear();supported=true;auto*m=machine.get();
 add("chip","Chip RAM","ram",m->ram.data(),m->ram.size(),m->overlay?"[]":"[0]");
 add("slow","Slow RAM","ram",m->slow.data(),m->slow.size(),"[12582912]",0xc00000);
 add("kickstart","Kickstart ROM","rom",m->rom.data(),m->rom.size(),m->overlay?"[0,16252928]":"[16252928]",0xf80000);
 // One physical disk split in the UI by track/side. Region index stays stable
 // across all tracks, so DMA attribution does not depend on the current head.
 add("disk","ADF disk sectors","disk",m->disk.data(),m->disk.size());
}
#include "../../../../browser/core/memory-api.inc"
