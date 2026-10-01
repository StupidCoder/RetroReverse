#include "cia.h"
namespace rr::c64 {
bool CiaTimer::tick(bool source){
 const bool count=queue&2,next=queue&1;queue=uint8_t(((queue<<1)|((control&1)&&source))&3);
 const bool skip=inhibit;inhibit=false;pulse=false;bool underflow=false;
 if(loadDelay&&!--loadDelay){counter=latch;inhibit=true;}
 else if(!skip){
  if(count&&counter)counter--;
  if(!counter&&next){counter=latch;inhibit=true;underflow=true;pulse=true;toggle=!toggle;
   if((control&8)||previousOneShot){control&=~1;queue=0;}
  }
 }
 previousOneShot=control&8;return underflow;
}
void CiaTimer::controlWrite(uint8_t value){if(!(control&1)&&(value&1))toggle=true;if(value&16)loadDelay=2;control=value&~16;}
uint8_t Cia::outputB()const{
 auto value=uint8_t(prb|uint8_t(~ddrb));if(a.control&2)value=(value&~64)|((a.control&4?a.toggle:a.pulse)?64:0);
 if(b.control&2)value=(value&~128)|((b.control&4?b.toggle:b.pulse)?128:0);return value;
}
uint8_t Cia::peek(uint8_t reg,uint8_t pa,uint8_t pb)const{
 switch(reg&15){case 0:return outputA()&pa;case 1:return outputB()&pb;case 2:return ddra;case 3:return ddrb;
 case 4:return uint8_t(a.counter);case 5:return uint8_t(a.counter>>8);case 6:return uint8_t(b.counter);case 7:return uint8_t(b.counter>>8);
 case 8:case 9:case 10:case 11:return (todLatched?todLatch:tod)[(reg&15)-8];case 12:return sdr;case 13:return flags|(irq?128:0);case 14:return a.control;case 15:return b.control;default:return 0xff;}
}
uint8_t Cia::read(uint8_t reg,uint8_t pa,uint8_t pb){
 reg&=15;if(reg==11&&!todLatched){todLatch=tod;todLatched=true;}
 const auto value=peek(reg,pa,pb);if(reg==8)todLatched=false;if(reg==13){flags=0;irq=irqDelay=false;}return value;
}
void Cia::event(uint8_t bits){flags|=bits&31;if(flags&mask)irqDelay=true;}
void Cia::write(uint8_t reg,uint8_t value){
 reg&=15;switch(reg){case 0:pra=value;break;case 1:prb=value;break;case 2:ddra=value;break;case 3:ddrb=value;break;
 case 4:a.latch=(a.latch&0xff00)|value;break;case 5:a.latch=uint16_t((a.latch&255)|(value<<8));if(!(a.control&1))a.counter=a.latch;break;
 case 6:b.latch=(b.latch&0xff00)|value;break;case 7:b.latch=uint16_t((b.latch&255)|(value<<8));if(!(b.control&1))b.counter=b.latch;break;
 case 8:case 9:case 10:case 11:{const auto index=reg-8;const auto v=uint8_t(value&(index==0?15:index==3?0x9f:0x7f));
  if(b.control&128)alarm[index]=v;else{tod[index]=v;if(index==3)todStopped=true;if(index==0){todStopped=false;todDivider=0;}}break;}
 case 12:sdr=value;serialPending=true;break;
 case 13:if(value&128)mask|=value&31;else mask&=uint8_t(~value)&31;if(flags&mask)irqDelay=true;break;
 case 14:if((a.control^value)&64){bits=0;serialActive=false;serialClock=true;}a.controlWrite(value);break;
 case 15:b.controlWrite(value);break;
 }
}
void Cia::tick(bool flag,bool cnt,bool sp){
 if(irqDelay){irq=true;irqDelay=false;}
 const bool rising=cnt&&!cntLine;if(!flag&&flagLine)event(16);flagLine=flag;cntLine=cnt;
 const bool underA=a.tick((a.control&32)?rising:true);
 const auto mode=(b.control>>5)&3;
 const bool underB=b.tick(mode==0?true:mode==1?rising:mode==2?underA:underA&&cnt);
 if(underA)event(1);if(underB)event(2);
 if(!(a.control&64)){
  if(rising){shift=uint8_t((shift<<1)|(sp?1:0));if(++bits==8){sdr=shift;bits=0;event(8);}}
 }else if(underA){
  if(!serialActive&&serialPending){shift=sdr;bits=0;serialActive=true;serialPending=false;}
  if(serialActive){serialClock=!serialClock;if(!serialClock)serialData=shift&128;else{shift<<=1;if(++bits==8){serialActive=false;event(8);}}}
 }
}
static uint8_t increment(uint8_t v){return uint8_t((v&15)==9?(v+7):(v+1));}
void Cia::todEdge(){
 if(todStopped)return;if(++todDivider<((a.control&128)?5:6))return;todDivider=0;
 if(++tod[0]>=10){tod[0]=0;tod[1]=increment(tod[1]);if(tod[1]>=0x60){tod[1]=0;tod[2]=increment(tod[2]);if(tod[2]>=0x60){tod[2]=0;auto h=tod[3]&31;auto pm=tod[3]&128;if(h==0x11){h=0x12;pm^=128;}else if(h==0x12)h=1;else h=increment(h);tod[3]=h|pm;}}}
 if(tod==alarm)event(4);
}
}
