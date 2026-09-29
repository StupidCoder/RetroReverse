#pragma once
// Keep the translated arithmetic/decoder, but handle M1 fetches, repeated
// index prefixes, ignored prefixes before ED, HALT and EI delay explicitly.
inline void z80_CPU_Step(z80_CPU*c){
 if(c->Halted)return;
 auto refresh=[&](){c->R=(c->R&128)|((c->R+1)&127);};
 c->Instrs++;
 if(c->intReq&&c->IFF1&&!c->eiPending){
  refresh();c->waiting=false;c->IFF1=c->IFF2=false;z80_CPU_push16(c,c->PC);
  c->PC=c->IM==2?z80_CPU_read16(c,(uint16_t(c->I)<<8)|255):0x38;return;
 }
 if(c->waiting){refresh();return;}
 c->eiPending=false;c->idx=0;c->usesHL=c->dispSet=false;
 auto fetch=[&](){refresh();return z80_CPU_fetch(c);};
 auto op=fetch();unsigned prefixes=0;
 while(op==0xdd||op==0xfd){if(++prefixes>=65536)throw std::runtime_error("Z80 prefix stream has no instruction");c->idx=op==0xdd?1:2;op=fetch();}
 if(op==0xed){c->idx=0;z80_CPU_execED(c,fetch());}
 else if(op==0xcb){
  if(c->idx){c->disp=z80_CPU_idxReg(c)+int8_t(z80_CPU_fetch(c));c->dispSet=c->usesHL=true;z80_CPU_execCB(c,z80_CPU_fetch(c));}
  else z80_CPU_execCB(c,fetch());
 }else z80_CPU_execMain(c,op);
}

inline uint8_t gamegear_Machine_In(gamegear_Machine*m,uint16_t port){
 rrgg::syncIO(m);auto p=uint8_t(port);auto&v=m->VDP;
 if(p>=0x80&&p<0xc0){if(!(p&1)){auto result=gamegear_VDP_readData(&v);v.addr&=0x3fff;return result;}auto result=gamegear_VDP_readStatus(&v);rrgg::timing.linePending=false;rrgg::irq(m);return result;}
 if(p>=0x40&&p<0x80)return p&1?0:v.line; // External H-counter latch is not modeled.
 if(p==0)return m->Pad00;
 if(p>=0xc0)return p&1?0xff:m->PadDC;
 return 0xff;
}
inline void gamegear_Machine_Out(gamegear_Machine*m,uint16_t port,uint8_t value){
 rrgg::syncIO(m);auto p=uint8_t(port);auto&v=m->VDP;
 if(p>=0x80&&p<0xc0){
  if(p&1){gamegear_VDP_writeControl(&v,value);v.addr&=0x3fff;if(!v.latched&&v.code==2){rrhh::memoryWrite(0x14040+(value&15),v.Regs[value&15]);rrgg::irq(m);}}
  else {
   auto address=v.addr;v.latched=false;v.readBuf=value;v.addr=(address+1)&0x3fff;
   if(v.code==3){
    if(!(address&1))rrgg::timing.cramLow=value;
    else {unsigned a=address&62;v.CRAM[a]=rrgg::timing.cramLow;v.CRAM[a+1]=value&15;v.CRAMWrites++;rrhh::memoryWrite(0x14000+a,v.CRAM[a]|uint16_t(v.CRAM[a+1])<<8,2);}
   }else {address&=0x3fff;v.VRAM[address]=value;v.Writes[address>>10]++;rrhh::memoryWrite(0x10000+address,value);}
  }
 }else if(p>=0x40&&p<0x80)gamegear_PSG_Write(&m->PSG,value);
}
