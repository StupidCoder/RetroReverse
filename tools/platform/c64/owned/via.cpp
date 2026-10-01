#include "via.h"
namespace rr::c64 {
uint8_t Via::outputB()const{auto v=uint8_t(orb|uint8_t(~ddrb));if(acr&128)v=(v&127)|(t1Out?128:0);return v;}
uint8_t Via::peek(uint8_t r)const{switch(r&15){
 case 0:{auto v=uint8_t((orb&ddrb)|((acr&2?latchB:pb)&uint8_t(~ddrb)));if(acr&128)v=(v&127)|(t1Out?128:0);return v;}
 case 1:case 15:return uint8_t((ora&ddra)|((acr&1?latchA:pa)&uint8_t(~ddra)));
 case 2:return ddrb;case 3:return ddra;case 4:return uint8_t(t1);case 5:return uint8_t(t1>>8);case 6:return uint8_t(t1Latch);case 7:return uint8_t(t1Latch>>8);
 case 8:return uint8_t(t2);case 9:return uint8_t(t2>>8);case 10:return sr;case 11:return acr;case 12:return pcr;case 13:return ifr|(irq()?128:0);case 14:return ier|128;default:return 255;}}
void Via::accessA(){ifr&=uint8_t(~2);const auto mode=(pcr>>1)&7;if(mode<4&&!(mode&1))ifr&=uint8_t(~1);if(mode==4||mode==5){ca2Out=false;pulseA=mode==5;}}
void Via::accessB(bool writing){ifr&=uint8_t(~16);const auto mode=(pcr>>5)&7;if(mode<4&&!(mode&1))ifr&=uint8_t(~8);if(writing&&(mode==4||mode==5)){cb2Out=false;pulseB=mode==5;}}
uint8_t Via::read(uint8_t r){const auto v=peek(r);switch(r&15){case 0:accessB(false);break;case 1:accessA();break;case 4:ifr&=uint8_t(~64);break;case 8:ifr&=uint8_t(~32);break;case 10:ifr&=uint8_t(~4);shiftBits=0;shiftActive=true;shiftTimer=t2Low+1;break;default:break;}return v;}
void Via::write(uint8_t r,uint8_t v){switch(r&15){
 case 0:orb=v;accessB(true);break;case 1:ora=v;accessA();break;case 15:ora=v;break;case 2:ddrb=v;break;case 3:ddra=v;break;
 case 4:case 6:t1Latch=(t1Latch&0xff00)|v;break;
 case 5:t1Latch=uint16_t((t1Latch&255)|(v<<8));t1=t1Latch;t1Delay=1;t1Reload=false;t1Armed=true;t1Out=false;ifr&=uint8_t(~64);break;
 case 7:t1Latch=uint16_t((t1Latch&255)|(v<<8));ifr&=uint8_t(~64);break;
 case 8:t2Low=v;break;case 9:t2=uint16_t(t2Low|(v<<8));t2Delay=1;t2Armed=true;ifr&=uint8_t(~32);break;
 case 10:sr=v;shiftBits=0;shiftActive=true;shiftClock=true;shiftTimer=t2Low+1;ifr&=uint8_t(~4);if(acr&16)cb2Out=sr&128;break;
 case 11:acr=v;break;case 12:pcr=v;pulseA=pulseB=false;ca2Out=((pcr>>1)&7)!=6;cb2Out=((pcr>>5)&7)!=6;break;
 case 13:ifr&=uint8_t(~v)&127;break;case 14:if(v&128)ier|=v&127;else ier&=uint8_t(~v)&127;break;
 }}
void Via::shift(bool input){const auto mode=(acr>>2)&7;if(!mode||!shiftActive)return;const bool out=sr&128;sr=uint8_t((sr<<1)|((mode&4)?out:input));if(mode&4)cb2Out=sr&128;if(++shiftBits==8){shiftBits=0;if(mode!=4){ifr|=4;shiftActive=false;}}}
void Via::tick(uint8_t portA,uint8_t portB,bool pinCA1,bool pinCA2,bool pinCB1,bool pinCB2){
 if(pulseA){ca2Out=true;pulseA=false;}if(pulseB){cb2Out=true;pulseB=false;}
 if(pinCA1!=ca1&&pinCA1==bool(pcr&1)){ifr|=2;latchA=portA;if(((pcr>>1)&7)==4)ca2Out=true;}
 if(pinCB1!=cb1&&pinCB1==bool(pcr&16)){ifr|=16;latchB=portB;if(((pcr>>5)&7)==4)cb2Out=true;}
 const auto amode=(pcr>>1)&7,bmode=(pcr>>5)&7;
 if(amode<4&&pinCA2!=ca2&&pinCA2==bool(amode&2))ifr|=1;
 if(bmode<4&&pinCB2!=cb2&&pinCB2==bool(bmode&2))ifr|=8;
 if(t1Delay)--t1Delay;else if(t1Reload){t1=t1Latch;t1Reload=false;}else if(t1--==0){if(t1Armed){ifr|=64;if(acr&64)t1Out=!t1Out;else{t1Out=true;t1Armed=false;}}if(acr&64)t1Reload=true;}
 if(t2Delay)--t2Delay;else if(!(acr&32)||((pb&64)&&!(portB&64))){if(t2--==0&&t2Armed){ifr|=32;t2Armed=false;}}
 const auto mode=(acr>>2)&7;if(mode&&shiftActive){
  if(mode==3||mode==7){if(!cb1&&pinCB1)shift(pinCB2);}
  else{bool clock=false;if(mode==2||mode==6)clock=true;else if(shiftTimer--==0){shiftTimer=t2Low+1;clock=true;}
   if(clock){shiftClock=!shiftClock;if(shiftClock)shift(pinCB2);}
  }
 }
 pa=portA;pb=portB;ca1=pinCA1;ca2=pinCA2;cb1=pinCB1;cb2=pinCB2;
}
}
