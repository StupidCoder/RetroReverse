#pragma once
#include <array>
#include <cstdint>
namespace rr::c64 {
struct CiaTimer {
 uint16_t latch=0xffff,counter=0xffff;
 uint8_t control=0,queue=0,loadDelay=0;
 bool inhibit=false,previousOneShot=false,toggle=false,pulse=false;
 bool tick(bool source);
 void controlWrite(uint8_t value);
};
// phi2 tick precedes the CPU's register transaction. FLAG/CNT are pin levels.
// All pipelines and edge history are value state, including timer reload delays.
struct Cia {
 CiaTimer a,b;
 uint8_t pra=0,prb=0,ddra=0,ddrb=0,flags=0,mask=0,sdr=0,shift=0,bits=0;
 bool irq=false,irqDelay=false,flagLine=true,cntLine=true,serialPending=false,serialActive=false,serialClock=true,serialData=true;
 std::array<uint8_t,4> tod{0,0,0,1},alarm{},todLatch{};
 uint8_t todDivider=0;
 bool todStopped=true,todLatched=false;
 uint8_t outputA()const{return pra|uint8_t(~ddra);}
 uint8_t outputB()const;
 uint8_t peek(uint8_t reg,uint8_t pa=255,uint8_t pb=255)const;
 uint8_t read(uint8_t reg,uint8_t pa=255,uint8_t pb=255);
 void write(uint8_t reg,uint8_t value);
 void tick(bool flag=true,bool cnt=true,bool sp=true);
 void todEdge();
 void event(uint8_t bits);
};
}
