#pragma once
#include <cstdint>

namespace rr::c64 {
// Independently authored NMOS CPU engine. The host services bus(), then calls
// tick() once per CPU clock. No device memory access or timing is hidden here.
enum Flag : uint8_t { C=1,Z=2,I=4,D=8,B=16,U=32,V=64,N=128 };
enum class Op : uint8_t { Unknown,JAM,SLO,RLA,SRE,RRA,SAX,LAX,DCP,ISC,ANC,ALR,ARR,AXS,LAS,ADC,AND,ASL,BCC,BCS,BEQ,BIT,BMI,BNE,BPL,BRK,BVC,BVS,CLC,CLD,CLI,CLV,CMP,CPX,CPY,DEC,DEX,DEY,EOR,INC,INX,INY,JMP,JSR,LDA,LDX,LDY,LSR,NOP,ORA,PHA,PHP,PLA,PLP,ROL,ROR,RTI,RTS,SBC,SEC,SED,SEI,STA,STX,STY,TAX,TAY,TSX,TXA,TXS,TYA };
enum class Mode : uint8_t { imp,acc,imm,zp,zpx,zpy,izx,izy,rel,abs,abx,aby,ind };
struct Instruction { Op op;Mode mode; };
extern const Instruction instructions[256];
struct Bus { uint16_t address=0;uint8_t data=0;bool write=false,sync=false; };
enum class Stage : uint8_t {
 Fetch,Implied,Immediate,Zero,ZeroIndex,AbsoluteLow,AbsoluteHigh,Indexed,
 PointerLow,PointerHigh,Read,Write,ModifyOld,ModifyNew,BranchOffset,BranchDummy,BranchCross,
 JumpLow,JumpHigh,SubLow,SubDummy,SubHighPush,SubLowPush,SubHigh,
 ReturnDummy,ReturnStack,ReturnLow,ReturnHigh,ReturnEnd,
 RtiDummy,RtiStack,RtiP,RtiLow,RtiHigh,PushDummy,Push,PullDummy,PullStack,Pull,
 BrkPad,InterruptDummy,InterruptHigh,InterruptLow,InterruptP,VectorLow,VectorHigh,
 ResetDummy,ResetStack1,ResetStack2,ResetStack3,ResetLow,ResetHigh,Jam,Fault
};
struct CpuState {
 uint16_t pc=0,address=0,target=0,vector=0xfffe;
 uint8_t a=0,x=0,y=0,s=0,p=U|I,opcode=0,low=0,pointer=0,value=0;
 Op op=Op::Unknown;Mode mode=Mode::imp;Stage stage=Stage::Fetch;
 Bus bus{};
 uint64_t clocks=0,retired=0;
 bool soLine=false,nmiLine=false,nmiPending=false,interruptPending=false,softwareInterrupt=false,irq=false,pollIrq=false,pollNmi=false;
};
class Cpu {
public:
 CpuState state{};
 void start(uint16_t pc); // Test/debug start at a known instruction boundary.
 void reset();           // Seven read cycles; preserves registers except I and SP effects.
 const Bus& bus()const{return state.bus;}
 bool boundary()const{return state.stage==Stage::Fetch;}
 bool faulted()const{return state.stage==Stage::Fault||state.stage==Stage::Jam;}
 // irq/nmi are asserted levels, rdy=false stalls reads but never writes.
 void tick(uint8_t data,bool irq=false,bool nmi=false,bool rdy=true,bool so=false);
private:
 void read(Stage stage,uint16_t address,bool sync=false);
 void write(Stage stage,uint16_t address,uint8_t data);
 void poll(bool masked);
 void finish(bool masked,bool sample=true);
 void access();
 void operand(uint8_t value);
 uint8_t modify(uint8_t value);
 void implied();
 void nz(uint8_t value);
 void flag(uint8_t bit,bool value);
 void compare(uint8_t left,uint8_t right);
 void adc(uint8_t value);
 void sbc(uint8_t value);
 bool store()const;
 bool modifying()const;
 bool branch()const;
};
} // namespace rr::c64
