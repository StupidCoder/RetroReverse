#include "cpu.h"
namespace rr::c64 {
void Cpu::read(Stage stage,uint16_t address,bool sync){state.stage=stage;state.bus={address,0,false,sync};}
void Cpu::write(Stage stage,uint16_t address,uint8_t data){state.stage=stage;state.bus={address,data,true,false};}
void Cpu::start(uint16_t pc){state={};state.pc=pc;read(Stage::Fetch,pc,true);}
void Cpu::reset(){state.interruptPending=state.nmiPending=false;state.softwareInterrupt=false;state.vector=0xfffc;state.value=0;read(Stage::ResetDummy,state.pc);}
void Cpu::flag(uint8_t bit,bool value){state.p=value?state.p|bit:state.p&~bit;}
void Cpu::nz(uint8_t v){flag(Z,v==0);flag(N,v&128);}
void Cpu::compare(uint8_t l,uint8_t r){flag(C,l>=r);nz(uint8_t(l-r));}
void Cpu::adc(uint8_t v){
 const unsigned a=state.a,carry=(state.p&C)?1:0,sum=a+v+carry;
 if(state.p&D){
  unsigned low=(a&15)+(v&15)+carry;if(low>=10)low=((low+6)&15)+16;
  unsigned intermediate=(a&0xf0)+(v&0xf0)+low;
  flag(Z,uint8_t(sum)==0);flag(N,intermediate&128);flag(V,(~(a^v)&(a^intermediate)&128)!=0);
  if(intermediate>=0xa0)intermediate+=0x60;
  flag(C,intermediate>255);state.a=uint8_t(intermediate);
 }else{flag(V,(~(a^v)&(a^sum)&128)!=0);flag(C,sum>255);state.a=uint8_t(sum);nz(state.a);}
}
void Cpu::sbc(uint8_t v){
 const int a=state.a,borrow=(state.p&C)?0:1,result=a-v-borrow;
 flag(V,((a^v)&(a^result)&128)!=0);flag(C,result>=0);nz(uint8_t(result));
 if(state.p&D){int low=(a&15)-(v&15)-borrow,high=(a>>4)-(v>>4);if(low<0){low-=6;high--;}if(high<0)high-=6;state.a=uint8_t((high*16)+(low&15));}
 else state.a=uint8_t(result);
}
bool Cpu::store()const{return state.op==Op::STA||state.op==Op::STX||state.op==Op::STY;}
bool Cpu::modifying()const{switch(state.op){case Op::ASL:case Op::LSR:case Op::ROL:case Op::ROR:case Op::INC:case Op::DEC:return true;default:return false;}}
void Cpu::operand(uint8_t v){switch(state.op){
 case Op::LDA:state.a=v;nz(v);break;case Op::LDX:state.x=v;nz(v);break;case Op::LDY:state.y=v;nz(v);break;
 case Op::ORA:state.a|=v;nz(state.a);break;case Op::AND:state.a&=v;nz(state.a);break;case Op::EOR:state.a^=v;nz(state.a);break;
 case Op::ADC:adc(v);break;case Op::SBC:sbc(v);break;
 case Op::CMP:compare(state.a,v);break;case Op::CPX:compare(state.x,v);break;case Op::CPY:compare(state.y,v);break;
 case Op::BIT:flag(Z,(state.a&v)==0);flag(N,v&128);flag(V,v&64);break;default:break;
}}
uint8_t Cpu::modify(uint8_t v){const uint8_t carry=(state.p&C)?1:0;switch(state.op){
 case Op::ASL:flag(C,v&128);v<<=1;break;case Op::LSR:flag(C,v&1);v>>=1;break;
 case Op::ROL:flag(C,v&128);v=uint8_t((v<<1)|carry);break;case Op::ROR:flag(C,v&1);v=uint8_t((v>>1)|(carry<<7));break;
 case Op::INC:v++;break;case Op::DEC:v--;break;default:break;
 }nz(v);return v;}
void Cpu::implied(){switch(state.op){
 case Op::CLC:flag(C,false);break;case Op::SEC:flag(C,true);break;case Op::CLI:flag(I,false);break;case Op::SEI:flag(I,true);break;
 case Op::CLD:flag(D,false);break;case Op::SED:flag(D,true);break;case Op::CLV:flag(V,false);break;
 case Op::DEX:--state.x;nz(state.x);break;case Op::DEY:--state.y;nz(state.y);break;case Op::INX:++state.x;nz(state.x);break;case Op::INY:++state.y;nz(state.y);break;
 case Op::TAX:state.x=state.a;nz(state.x);break;case Op::TAY:state.y=state.a;nz(state.y);break;
 case Op::TXA:state.a=state.x;nz(state.a);break;case Op::TYA:state.a=state.y;nz(state.a);break;
 case Op::TSX:state.x=state.s;nz(state.x);break;case Op::TXS:state.s=state.x;break;default:break;
}}
bool Cpu::branch()const{switch(state.op){
 case Op::BCC:return !(state.p&C);case Op::BCS:return state.p&C;case Op::BEQ:return state.p&Z;case Op::BNE:return !(state.p&Z);
 case Op::BMI:return state.p&N;case Op::BPL:return !(state.p&N);case Op::BVS:return state.p&V;case Op::BVC:return !(state.p&V);default:return false;
}}
void Cpu::finish(bool masked){++state.retired;state.interruptPending=state.nmiPending||(state.irq&&!masked);read(Stage::Fetch,state.pc,true);}
void Cpu::access(){if(store())write(Stage::Write,state.address,state.op==Op::STA?state.a:state.op==Op::STX?state.x:state.y);else read(Stage::Read,state.address);}
void Cpu::tick(uint8_t data,bool irq,bool nmi,bool rdy){
 ++state.clocks;state.irq=irq;if(nmi&&!state.nmiLine)state.nmiPending=true;state.nmiLine=nmi;
 if(!rdy&&!state.bus.write)return;
 const bool masked=state.p&I;
 const auto nextByte=[&](Stage s){read(s,state.pc++);};
 const auto stack=[&](Stage s){read(s,uint16_t(0x100|state.s));};
 switch(state.stage){
 case Stage::Fetch:{
  if(state.interruptPending){state.interruptPending=false;state.softwareInterrupt=false;state.vector=state.nmiPending?0xfffa:0xfffe;state.nmiPending=false;read(Stage::InterruptDummy,state.pc);break;}
  state.opcode=data;state.op=instructions[data].op;state.mode=instructions[data].mode;state.pc++;
  switch(state.op){
   case Op::Unknown:read(Stage::Fault,state.pc);return;
   case Op::BRK:state.softwareInterrupt=true;state.vector=0xfffe;nextByte(Stage::BrkPad);return;
   case Op::JSR:nextByte(Stage::SubLow);return;
   case Op::RTS:read(Stage::ReturnDummy,state.pc);return;
   case Op::RTI:read(Stage::RtiDummy,state.pc);return;
   case Op::PHA:case Op::PHP:read(Stage::PushDummy,state.pc);return;
   case Op::PLA:case Op::PLP:read(Stage::PullDummy,state.pc);return;
   default:break;
  }
  switch(state.mode){
   case Mode::imp:case Mode::acc:read(Stage::Implied,state.pc);break;
   case Mode::imm:nextByte(Stage::Immediate);break;case Mode::rel:nextByte(Stage::BranchOffset);break;
   case Mode::zp:case Mode::zpx:case Mode::zpy:case Mode::izx:case Mode::izy:nextByte(Stage::Zero);break;
   default:nextByte(Stage::AbsoluteLow);break;
  }break;
 }
 case Stage::Implied:if(state.mode==Mode::acc)state.a=modify(state.a);else implied();finish(masked);break;
 case Stage::Immediate:operand(data);finish(masked);break;
 case Stage::Zero:
  state.pointer=data;state.address=data;
  if(state.mode==Mode::zpx||state.mode==Mode::zpy||state.mode==Mode::izx)read(Stage::ZeroIndex,data);
  else if(state.mode==Mode::izy)read(Stage::PointerLow,data);else access();break;
 case Stage::ZeroIndex:
  state.pointer=uint8_t(state.pointer+(state.mode==Mode::zpy?state.y:state.x));state.address=state.pointer;
  if(state.mode==Mode::izx)read(Stage::PointerLow,state.pointer);else access();break;
 case Stage::PointerLow:state.low=data;read(Stage::PointerHigh,uint8_t(state.pointer+1));break;
 case Stage::PointerHigh:
  state.address=uint16_t(state.low|(data<<8));
  if(state.mode==Mode::izy){state.target=uint16_t(state.address+state.y);read(Stage::Indexed,uint16_t((state.address&0xff00)|(state.target&255)));}
  else access();break;
 case Stage::AbsoluteLow:state.low=data;nextByte(Stage::AbsoluteHigh);break;
 case Stage::AbsoluteHigh:
  state.address=uint16_t(state.low|(data<<8));
  if(state.op==Op::JMP){if(state.mode==Mode::ind)read(Stage::JumpLow,state.address);else{state.pc=state.address;finish(masked);}}
  else if(state.mode==Mode::abx||state.mode==Mode::aby){state.target=uint16_t(state.address+(state.mode==Mode::abx?state.x:state.y));read(Stage::Indexed,uint16_t((state.address&0xff00)|(state.target&255)));}
  else access();break;
 case Stage::Indexed:{bool crossing=(state.address&0xff00)!=(state.target&0xff00);state.address=state.target;
  if(crossing||store()||modifying())access();else{operand(data);finish(masked);}break;}
 case Stage::Read:if(modifying()){state.value=data;write(Stage::ModifyOld,state.address,data);}else{operand(data);finish(masked);}break;
 case Stage::Write:finish(masked);break;
 case Stage::ModifyOld:state.value=modify(state.value);write(Stage::ModifyNew,state.address,state.value);break;
 case Stage::ModifyNew:finish(masked);break;
 case Stage::BranchOffset:
  if(!branch()){finish(masked);break;}state.target=uint16_t(state.pc+int8_t(data));read(Stage::BranchDummy,state.pc);break;
 case Stage::BranchDummy:
  if((state.pc&0xff00)!=(state.target&0xff00))read(Stage::BranchCross,uint16_t((state.pc&0xff00)|(state.target&255)));
  else{state.pc=state.target;finish(masked);}break;
 case Stage::BranchCross:state.pc=state.target;finish(masked);break;
 case Stage::JumpLow:state.low=data;read(Stage::JumpHigh,uint16_t((state.address&0xff00)|uint8_t(state.address+1)));break;
 case Stage::JumpHigh:state.pc=uint16_t(state.low|(data<<8));finish(masked);break;
 case Stage::SubLow:state.low=data;stack(Stage::SubDummy);break;
 case Stage::SubDummy:write(Stage::SubHighPush,uint16_t(0x100|state.s),uint8_t(state.pc>>8));break;
 case Stage::SubHighPush:--state.s;write(Stage::SubLowPush,uint16_t(0x100|state.s),uint8_t(state.pc));break;
 case Stage::SubLowPush:--state.s;read(Stage::SubHigh,state.pc);break;
 case Stage::SubHigh:state.pc=uint16_t(state.low|(data<<8));finish(masked);break;
 case Stage::ReturnDummy:stack(Stage::ReturnStack);break;
 case Stage::ReturnStack:++state.s;stack(Stage::ReturnLow);break;
 case Stage::ReturnLow:state.low=data;++state.s;stack(Stage::ReturnHigh);break;
 case Stage::ReturnHigh:state.pc=uint16_t(state.low|(data<<8));read(Stage::ReturnEnd,state.pc);break;
 case Stage::ReturnEnd:++state.pc;finish(masked);break;
 case Stage::RtiDummy:stack(Stage::RtiStack);break;
 case Stage::RtiStack:++state.s;stack(Stage::RtiP);break;
 case Stage::RtiP:state.p=(data&~B)|U;++state.s;stack(Stage::RtiLow);break;
 case Stage::RtiLow:state.low=data;++state.s;stack(Stage::RtiHigh);break;
 case Stage::RtiHigh:state.pc=uint16_t(state.low|(data<<8));finish(masked);break;
 case Stage::PushDummy:write(Stage::Push,uint16_t(0x100|state.s),state.op==Op::PHA?state.a:state.p|B|U);break;
 case Stage::Push:--state.s;finish(masked);break;
 case Stage::PullDummy:stack(Stage::PullStack);break;
 case Stage::PullStack:++state.s;stack(Stage::Pull);break;
 case Stage::Pull:if(state.op==Op::PLA){state.a=data;nz(data);}else state.p=(data&~B)|U;finish(masked);break;
 case Stage::BrkPad:case Stage::InterruptDummy:write(Stage::InterruptHigh,uint16_t(0x100|state.s),uint8_t(state.pc>>8));break;
 case Stage::InterruptHigh:--state.s;write(Stage::InterruptLow,uint16_t(0x100|state.s),uint8_t(state.pc));break;
 case Stage::InterruptLow:--state.s;write(Stage::InterruptP,uint16_t(0x100|state.s),uint8_t((state.p|U)&~B)|(state.softwareInterrupt?B:0));break;
 case Stage::InterruptP:--state.s;flag(I,true);if(state.nmiPending){state.vector=0xfffa;state.nmiPending=false;}read(Stage::VectorLow,state.vector);break;
 case Stage::VectorLow:state.low=data;read(Stage::VectorHigh,uint16_t(state.vector+1));break;
 case Stage::VectorHigh:state.pc=uint16_t(state.low|(data<<8));if(state.softwareInterrupt)++state.retired;state.softwareInterrupt=false;read(Stage::Fetch,state.pc,true);break;
 case Stage::ResetDummy:if(state.value++==0)read(Stage::ResetDummy,state.pc);else stack(Stage::ResetStack1);break;
 case Stage::ResetStack1:--state.s;stack(Stage::ResetStack2);break;
 case Stage::ResetStack2:--state.s;stack(Stage::ResetStack3);break;
 case Stage::ResetStack3:--state.s;flag(I,true);read(Stage::ResetLow,0xfffc);break;
 case Stage::ResetLow:state.low=data;read(Stage::ResetHigh,0xfffd);break;
 case Stage::ResetHigh:state.pc=uint16_t(state.low|(data<<8));read(Stage::Fetch,state.pc,true);break;
 case Stage::Fault:break;
 }
}
} // namespace rr::c64
