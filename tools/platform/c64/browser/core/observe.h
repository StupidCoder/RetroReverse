#pragma once
#include <cstdint>
// Hooks do not alter machine state. Capture hooks are dormant during ordinary execution.
namespace observation {
extern bool capturing;
void write(int space,uint16_t address,uint8_t value);
void fetch(int kind,int slot,uint16_t address,uint16_t value);
void pixel(void* vic,uint8_t* dst,uint16_t graphics,uint16_t sprites,bool border);
void load(bool enabled);
void idle();
void spriteOutput(int index,bool pair);
void spriteShift(int index);
}
#define RR_VIC_WRITE(addr,value) observation::write(2,addr,value)
#define RR_CPU_WRITE(space,address,value) observation::write(space,address,value)
#define RR_VIC_FETCH(kind,slot,addr,value) do { if(observation::capturing) observation::fetch(kind,slot,addr,value); } while(0)
#define RR_VIC_IDLE() do { if(observation::capturing) observation::idle(); } while(0)
#define RR_VIC_LOAD(enabled) do { if(observation::capturing) observation::load(enabled); } while(0)
#define RR_VIC_PIXEL(vic,dst,bmc,sc,border) do { if(observation::capturing) observation::pixel(vic,dst,bmc,sc,border); } while(0)
#define RR_SPRITE_OUTPUT(i,pair) do { if(observation::capturing) observation::spriteOutput(i,pair); } while(0)
#define RR_SPRITE_SHIFT(i) do { if(observation::capturing) observation::spriteShift(i); } while(0)
