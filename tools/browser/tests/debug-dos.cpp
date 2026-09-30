#include "../../platform/dos/browser/core/api.cpp"
#include <cassert>
#include <iostream>
#include "x86-decode-vectors.h"
static void boot(){auto b=Slice<uint8_t>::make(128);b[0]='M';b[1]='Z';b[2]=128;b[4]=1;b[8]=4;b[16]=0xfe;b[17]=0xff;b[64]=0x90;b[65]=0xeb;b[66]=0xfd;rrfiles::overlay["test.exe"]=b;rrBoot("test.exe",false);realMachine->EnableIRQ=false;}
int main(){boot();auto*c=rrCPU();auto pc=rrPC();auto before=std::string(rr_proof());auto s=std::string(rr_debug_snapshot(-1));assert(s.find("real16")!=std::string::npos);assert(s.find("NOP")!=std::string::npos);assert(before==rr_proof());assert(std::string(rr_debug_snapshot(0xa0000)).find("\"bytes\":[null")!=std::string::npos);
 assert(rr_debug_begin(2,pc,0));assert(rr_debug_run(10)==4);assert(c->Steps==0);assert(rr_debug_begin(1,0,0));assert(rr_debug_run(10)==2);assert(rrPC()==pc+1&&c->Steps==1);assert(rr_debug_begin(2,pc,0));assert(rr_debug_run(10)==4);
 // SMC is decoded from current backing bytes, including operand changes.
 auto&m=realMachine->Mem;m[pc]=0xb8;m[pc+1]=0x34;m[pc+2]=0x12;assert(std::string(rr_debug_snapshot(-1)).find("$1234")!=std::string::npos);m[pc+1]=0x78;assert(std::string(rr_debug_snapshot(-1)).find("$1278")!=std::string::npos);
 // Interrupt delivery is a boundary with zero retired instructions.
 bool irq=true;int ticks=0;c->OnStep=[&](x86_CPU*c){ticks++;if(irq){irq=false;c->IP=0x100;}};m[(c->Seg[1]<<4)+0x100]=0x90;auto steps=c->Steps;assert(rr_debug_begin(1,0,0));assert(rr_debug_run(1)==3);assert(c->Steps==steps&&ticks==1&&rrDebugHookPending);assert(rr_debug_begin(1,0,0));assert(rr_debug_run(1)==2);assert(c->Steps==steps+1&&ticks==1&&!rrDebugHookPending);
 irq=true;steps=c->Steps;assert(rr_debug_begin(2,(c->Seg[1]<<4)+0x100,1));assert(rr_debug_run(1)==4&&c->Steps==steps&&rrDebugHookPending);
 // Pending device boundary survives portable state and normal execution.
 rrBind(realMachine);rrDebugHookPending=true;auto n=rr_state_save();assert(n);auto saved=stateOutput;rrDebugHookPending=false;memcpy(rr_state_input(n),saved.data(),n);assert(rr_state_load(n)&&rrDebugHookPending);c=rrCPU();assert(rr_run(1)==1&&!rrDebugHookPending);
 // Real-mode instruction fetch and near branches wrap at the segment boundary.
 c->IP=0xffff;auto base=uint32_t(c->Seg[1])<<4;realMachine->Mem[(base+0xffff)&0xfffff]=0xeb;realMachine->Mem[base]=0xfe;
 s=rr_debug_snapshot(-1);assert(s.find("\"bytes\":[235,254]")!=std::string::npos);assert(rr_debug_begin(1,0,0));assert(rr_debug_run(1)==2&&c->IP==0xffff);
 // Flat protected mode uses cached descriptor bases, never selector*16.
 c->Mode=1;c->Seg[1]=0x23;c->SegBase[1]=0x40000;c->IP=0x100;realMachine->Mem[0x40100]=0xb8;realMachine->Mem[0x40101]=0x78;realMachine->Mem[0x40102]=0x56;realMachine->Mem[0x40103]=0x34;realMachine->Mem[0x40104]=0x12;
 s=rr_debug_snapshot(-1);assert(rrPC()==0x40100);assert(s.find("protected32")!=std::string::npos&&s.find("$12345678")!=std::string::npos);assert(rr_debug_begin(1,0,0));assert(rr_debug_run(1)==2);assert(c->Regs[0]==0x12345678&&c->IP==0x105);
 assert(rr_debug_begin(2,rrPC(),0));rr_debug_bind(0x24,0x40000);assert(rr_debug_run(1)==5);
 // REP budget is explicit; unsupported huge atomic operations do not hang.
 auto a=rrPC();realMachine->Mem[a]=0xf3;realMachine->Mem[a+1]=0xa4;c->Regs[1]=0xffffffff;assert(rr_debug_begin(1,0,0));assert(rr_debug_run(1)==7);
 // Differential against the repository Go decoder, in both default modes.
 for(const auto&v:decodeVectors){c->Mode=v.wide?1:0;memcpy(realMachine->Mem.p+0x1234,v.bytes.data(),15);auto d=rrDecode(0x1234);if(d.Len!=v.length||d.Text!=v.text){std::cerr<<int(v.bytes[0])<<" "<<v.wide<<" "<<d.Text<<" != "<<v.text<<"\n";assert(false);}}
 std::cout<<"DOS safe peeks, live 16/32-bit decoding, descriptor bases, SMC, single-step, target contexts, IRQ boundary/restore and REP bound: PASS\n";
}
