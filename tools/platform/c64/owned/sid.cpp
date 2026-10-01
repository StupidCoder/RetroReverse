#include "sid.h"
namespace rr::c64 {
static constexpr uint16_t rates[]={9,32,63,95,149,220,267,313,392,977,1954,3126,3907,11720,19532,31251};
void Sid::write(uint8_t reg,uint8_t value){
 reg&=31;bus=value;busAge=0;if(reg>=25)return;
 const auto old=regs[reg];regs[reg]=value;
 if(reg<21&&reg%7==4){auto& v=voice[reg/7];
  if((old^value)&1){v.stage=value&1?Envelope::Attack:Envelope::Release;if(value&1)v.zero=false;}
  if((value&8)&&!(old&8)){v.phase=0;v.noiseDelay=0;v.testCycles=0;}
  if((old&8)&&!(value&8)){v.noise=((v.noise<<1)|((~v.noise>>17)&1))&0x7fffff;}
  const auto wave=value>>4;if(wave&&(wave&(wave-1)))combinedWaveformUsed=true;
 }
}
uint16_t Sid::wave(unsigned i)const{
 const auto& v=voice[i];const auto base=i*7;const uint8_t control=regs[base+4];
 uint16_t output=4095;bool selected=false;
 if(control&16){uint32_t sign=v.phase;if(control&4)sign^=voice[(i+2)%3].phase;auto triangle=uint16_t((v.phase>>11)&4095);if(sign&0x800000)triangle^=4095;output&=triangle;selected=true;}
 if(control&32){output&=uint16_t(v.phase>>12);selected=true;}
 if(control&64){const auto width=uint16_t(regs[base+2]|((regs[base+3]&15)<<8));output&=(control&8)||(v.phase>>12)>=width?4095:0;selected=true;}
 if(control&128){const unsigned taps[]={20,18,14,11,9,5,2,0};uint16_t noise=0;for(auto bit:taps)noise=uint16_t((noise<<1)|((v.noise>>bit)&1));output&=noise<<4;selected=true;}
 return selected?output:0;
}
uint8_t Sid::peek(uint8_t reg)const{switch(reg&31){case 25:case 26:return 255;case 27:return uint8_t(wave(2)>>4);case 28:return voice[2].envelope;default:return bus;}}
uint8_t Sid::read(uint8_t reg){const auto value=peek(reg);if((reg&31)>=25&&(reg&31)<=28){bus=value;busAge=0;}return value;}
void Sid::tick(){
 if(busAge<0x2000&&++busAge==0x2000)bus=0;
 std::array<bool,3> rising{};
 for(unsigned i=0;i<3;i++){
  auto& v=voice[i];const auto base=i*7;const auto control=regs[base+4];
  if(control&8){v.phase=0;if(v.testCycles<0x8000&&++v.testCycles==0x8000)v.noise=0x7fffff;}
  else{
   if(v.noiseDelay&&!--v.noiseDelay)v.noise=((v.noise<<1)|(((v.noise>>22)^(v.noise>>17))&1))&0x7fffff;
   const auto old=v.phase;v.phase=(v.phase+regs[base]+(uint32_t(regs[base+1])<<8))&0xffffff;
   rising[i]=!(old&0x800000)&&(v.phase&0x800000);if(!(old&0x80000)&&(v.phase&0x80000))v.noiseDelay=2;
  }
  // A rate change doesn't reset the counter: equality after wrapping reproduces
  // the ADSR delay mechanism. Gate transitions don't restart the rate phase.
  if(++v.rate==0x8000)v.rate=1;
  const auto rate=v.stage==Envelope::Attack?regs[base+5]>>4:v.stage==Envelope::Decay?regs[base+5]&15:regs[base+6]&15;
  if(v.rate!=rates[rate])continue;v.rate=0;
  if(v.stage==Envelope::Attack){v.exponential=0;if(++v.envelope==255)v.stage=Envelope::Decay;}
  else if(++v.exponential==v.divider){v.exponential=0;if(!v.zero&&(v.stage!=Envelope::Decay||v.envelope!=uint8_t((regs[base+6]>>4)*17)))--v.envelope;}
  switch(v.envelope){case 255:v.divider=1;break;case 0x5d:v.divider=2;break;case 0x36:v.divider=4;break;case 0x1a:v.divider=8;break;case 0x0e:v.divider=16;break;case 6:v.divider=30;break;case 0:v.divider=1;v.zero=true;break;default:break;}
 }
 // Evaluate all three sources before any reset (no voice-order dependence).
 for(unsigned i=0;i<3;i++){const auto source=(i+2)%3;if((regs[i*7+4]&2)&&rising[source]&&!((regs[source*7+4]&2)&&rising[(source+2)%3]))voice[i].phase=0;}
}
}
