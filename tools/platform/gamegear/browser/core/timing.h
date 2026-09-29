#pragma once
namespace rrgg {
// Sega's 342-dot line is 228 Z80 T-states; phase zero is H counter F4.
constexpr unsigned lineCycles=228,frameCycles=228*262;
struct Timing {
 uint64_t cycles=0;
 uint8_t scrollX=0,scrollY=0,lineCounter=1,cramLow=0;
 bool linePending=false,started=false;
};
inline Timing timing;
inline unsigned instructionCycles=0,ioCycles=0,elapsedCycles=0;
inline void tick(gamegear_Machine*,unsigned);
inline void syncIO(gamegear_Machine*m){if(ioCycles>elapsedCycles){tick(m,ioCycles-elapsedCycles);elapsedCycles=ioCycles;}}
inline void irq(gamegear_Machine*m){z80_CPU_RequestIRQ(m->CPU,((m->VDP.status&128)&&(m->VDP.Regs[1]&32))||(timing.linePending&&(m->VDP.Regs[0]&16)));}
inline uint8_t counter(unsigned line){return line<=218?line:line-6;}

// Durations include prefixes, taken branches and each repeating block iteration.
// I/O is synchronized at the end of its bus cycle, before trailing block work.
inline unsigned duration(gamegear_Machine*m){
 auto*c=m->CPU;ioCycles=0;
 if(c->intReq&&c->IFF1&&!c->eiPending)return c->IM==2?19:13;
 if(c->waiting)return 4;
 uint16_t pc=c->PC;auto read=[&](){auto v=gamegear_Machine_Read(m,pc);pc++;return v;};
 unsigned prefixes=0;uint8_t op=read();
 while(op==0xdd||op==0xfd){if(++prefixes>=65536)throw std::runtime_error("Z80 prefix stream has no instruction");op=read();}
 unsigned extra=prefixes*4;
 if(op==0xcb){if(prefixes){read();op=read();return extra+((op>>6)==1?16:19);}op=read();return (op&7)==6?((op>>6)==1?12:15):8;}
 if(op==0xed){
  op=read();unsigned x=op>>6,y=(op>>3)&7,z=op&7;
  if(x==2&&y>=4&&z<=3){
   bool repeat=y>=6;
   if(z==0)repeat=repeat&&z80_CPU_bc(c)!=1;
   else if(z==1)repeat=repeat&&z80_CPU_bc(c)!=1&&c->A!=gamegear_Machine_Read(m,z80_CPU_hl(c));
   else {repeat=repeat&&c->B!=1;ioCycles=extra+(z==2?12:16);}
   return extra+(repeat?21:16);
  }
  if(x!=1)return extra+8;
  if(z<=1){ioCycles=extra+12;return ioCycles;}
  if(z==2)return extra+15;if(z==3)return extra+20;
  if(z==5)return extra+14;if(z==7)return extra+(y<=3?9:y<=5?18:8);
  return extra+8;
 }
 unsigned x=op>>6,y=(op>>3)&7,z=op&7,p=y>>1,q=y&1,n=4;
 if(x==0){switch(z){
  case 0:n=y<2?4:y==2?(c->B!=1?13:8):y==3?12:(z80_CPU_cond(c,y-4)?12:7);break;
  case 1:n=q?11:10;break;case 2:n=p<2?7:p==2?16:13;break;case 3:n=6;break;
  case 4:case 5:n=y==6?(prefixes?19:11):4;break;
  case 6:n=y==6?(prefixes?15:10):7;break;
 }}else if(x==1)n=op==0x76?4:(y==6||z==6)?(prefixes?15:7):4;
 else if(x==2)n=z==6?(prefixes?15:7):4;
 else switch(z){
  case 0:n=z80_CPU_cond(c,y)?11:5;break;
  case 1:n=!q?10:p==0?10:p==3?6:4;break;case 2:n=10;break;
  case 3:n=y==0?10:(y==2||y==3)?11:y==4?19:4;if(y==2||y==3)ioCycles=extra+n;break;
  case 4:n=z80_CPU_cond(c,y)?17:10;break;case 5:n=!q?11:p==0?17:4;break;
  case 6:n=7;break;case 7:n=11;break;
 }
 return extra+n;
}
}
