#pragma once
#include <cstdint>
namespace rr::c64 {
// Independent 6522 model. tick samples peripheral pins before the CPU access.
// VIA timer/control semantics are deliberately separate from the CIA model.
struct Via {
 uint8_t ora=0,orb=0,ddra=0,ddrb=0,acr=0,pcr=0,ifr=0,ier=0,sr=0;
 uint8_t pa=255,pb=255,latchA=255,latchB=255,shiftBits=0;
 uint16_t t1=65535,t1Latch=65535,t2=65535;
 uint8_t t2Low=255,t1Delay=0,t2Delay=0;
 uint16_t shiftTimer=0;
 bool t1Armed=false,t2Armed=false,t1Reload=false,t1Out=true;
 bool ca1=true,ca2=true,cb1=true,cb2=true,ca2Out=true,cb2Out=true;
 bool pulseA=false,pulseB=false,shiftActive=false,shiftClock=true;
 void tick(uint8_t portA=255,uint8_t portB=255,bool pinCA1=true,bool pinCA2=true,bool pinCB1=true,bool pinCB2=true);
 uint8_t peek(uint8_t reg)const;
 uint8_t read(uint8_t reg);
 void write(uint8_t reg,uint8_t value);
 uint8_t outputA()const{return ora|uint8_t(~ddra);}
 uint8_t outputB()const;
 bool irq()const{return (ifr&ier&127)!=0;}
private:
 void accessA();
 void accessB(bool writing);
 void shift(bool input);
};
}
