#pragma once
#include <cstdint>
// Hooks do not alter machine state. Capture hooks are dormant during ordinary execution.
namespace observation {
struct Write;
extern bool capturing;
void write(int space,uint16_t address,uint8_t value);
void fetch(int kind,int slot,uint16_t address,uint16_t value);
void pixel(void* vic,uint8_t* dst,uint16_t graphics,uint16_t sprites,bool border);
void load(bool enabled);
void idle();
void spriteOutput(int index,bool pair);
void spriteShift(int index);
}
namespace raster {
extern bool previewing;
void reset();void begin();void write(const observation::Write&);
void recordLine(unsigned y,void* vic);
void used(int kind,uint16_t address);
void fetch(int kind,int slot,uint16_t address,uint16_t value);
void pixel(void* vic,uint8_t* dst,uint16_t graphics,uint16_t sprites,bool border);
void load(bool enabled);void idle();void spriteOutput(int index,bool pair);void spriteShift(int index);
}
#define RR_VIC_WRITE(addr,value) do { if(!raster::previewing) observation::write(2,addr,value); } while(0)
#define RR_CPU_WRITE(space,address,value) observation::write(space,address,value)
#define RR_VIC_FETCH(kind,slot,addr,value) do { if(observation::capturing) observation::fetch(kind,slot,addr,value); else if(raster::previewing) raster::fetch(kind,slot,addr,value); } while(0)
#define RR_VIC_IDLE() do { if(observation::capturing) observation::idle(); else if(raster::previewing) raster::idle(); } while(0)
#define RR_VIC_LOAD(enabled) do { if(observation::capturing) observation::load(enabled); else if(raster::previewing) raster::load(enabled); } while(0)
#define RR_VIC_PIXEL(vic,dst,bmc,sc,border) do { if(observation::capturing) observation::pixel(vic,dst,bmc,sc,border); else if(raster::previewing) raster::pixel(vic,dst,bmc,sc,border); } while(0)
#define RR_SPRITE_OUTPUT(i,pair) do { if(observation::capturing) observation::spriteOutput(i,pair); else if(raster::previewing) raster::spriteOutput(i,pair); } while(0)
#define RR_SPRITE_SHIFT(i) do { if(observation::capturing) observation::spriteShift(i); else if(raster::previewing) raster::spriteShift(i); } while(0)
