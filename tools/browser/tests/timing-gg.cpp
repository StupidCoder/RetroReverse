#include "../../platform/gamegear/browser/core/api.cpp"
#include <cassert>
#include <iostream>
static void reset(){
 auto*c=machine->CPU;*c={};c->bus=machine;c->SP=0xd000;c->PC=0x100;
 machine->VDP={};machine->ram.fill(0);rrhh::init(0xff000000);rrgg::timing={};
}
static void program(std::initializer_list<uint8_t> bytes){unsigned a=0x100;for(auto v:bytes)machine->rom[a++]=v;}
static void cycles(unsigned n){auto before=rrgg::timing.cycles;assert(rr_run(1)==1);assert(rrgg::timing.cycles-before==n);}
static void reg(unsigned n,uint8_t value){gamegear_Machine_Out(machine,0xbf,value);gamegear_Machine_Out(machine,0xbf,0x80|n);}
int main(){
 std::vector<uint8_t>rom(32768);memcpy(rr_input(rom.size()),rom.data(),rom.size());assert(rr_init(rom.size()));auto*c=machine->CPU;
 reset();program({0,0x3e,7,0x32,0,0xc0,0x18,0xfe});cycles(4);cycles(7);cycles(13);cycles(12);assert(machine->ram[0]==7);
 reset();program({0x10,0xfe});c->B=2;cycles(13);cycles(8);assert(c->PC==0x102);
 reset();program({0x20,0});c->F=0x40;cycles(7);c->PC=0x100;c->F=0;cycles(12);
 reset();program({0xc4,0,2});c->F=0x40;cycles(10);c->PC=0x100;c->F=0;cycles(17);assert(c->PC==0x200);
 reset();program({0xc0});c->F=0x40;cycles(5);c->PC=0x100;c->F=0;cycles(11);
 reset();program({0xdd,0x21,0,0xc0,0xdd,0x36,1,0x5a,0xdd,0x34,1,0xdd,0xcb,1,0x46,0xdd,0xcb,1,0xc6});cycles(14);cycles(19);cycles(23);cycles(20);cycles(23);assert(machine->ram[1]==0x5b&&c->R==10);
 reset();program({0xcb,0x46,0xcb,0xc6});c->H=0xc0;cycles(12);cycles(15);
 reset();program({0xdd,0xfd,0x21,0x34,0x12});cycles(18);assert(c->IY==0x1234&&c->IX==0&&c->R==3);
 reset();program({0xdd,0xed,0x4b,0,0xc0});machine->ram[0]=0x78;machine->ram[1]=0x56;cycles(24);assert(z80_CPU_bc(c)==0x5678);
 reset();program({0xed,0xb0});c->H=0xc0;c->D=0xc1;c->C=2;machine->ram[0]=17;machine->ram[1]=18;cycles(21);cycles(16);assert(machine->ram[256]==17&&machine->ram[257]==18);
 reset();program({0xed,0xb1});c->H=0xc0;c->C=2;c->A=18;machine->ram[0]=17;machine->ram[1]=18;cycles(21);cycles(16);
 reset();program({0xed,0xb2});c->H=0xc0;c->B=2;c->C=0xdc;cycles(21);cycles(16);
 reset();program({0xed,0xb3});c->H=0xc0;c->B=2;c->C=0xbe;cycles(21);cycles(16);
 reset();program({0xfb,0xfb,0,0x76});c->intReq=true;rrgg::timing.started=true;cycles(4);cycles(4);cycles(4);cycles(13);assert(c->PC==0x38&&!c->IFF1);
 reset();program({0x76});cycles(4);cycles(4);assert(c->waiting&&c->R==2);c->intReq=c->IFF1=true;cycles(13);assert(!c->waiting);
 reset();c->IM=2;c->I=0xc0;c->IFF1=c->intReq=true;rrgg::timing.started=true;machine->ram[255]=0x34;machine->ram[256]=0x12;cycles(19);assert(c->PC==0x1234);

 reset();auto&v=machine->VDP;v.Regs[10]=3;rrgg::timing.lineCounter=2;rrgg::tick(machine,0);assert(rrgg::timing.lineCounter==1);
 rrgg::tick(machine,228);assert(rrgg::timing.lineCounter==0&&!rrgg::timing.linePending);
 rrgg::tick(machine,228);assert(rrgg::timing.lineCounter==3&&rrgg::timing.linePending&&!c->intReq);
 reg(0,0x14);assert(c->intReq);reg(10,7);assert(rrgg::timing.lineCounter==3);reg(0,4);assert(!c->intReq&&rrgg::timing.linePending);reg(0,0x14);assert(c->intReq);
 assert(!(gamegear_Machine_In(machine,0xbf)&128));assert(!rrgg::timing.linePending&&!c->intReq);
 rrgg::tick(machine,190*228);assert(v.line==192&&(v.status&128));reg(1,0x20);assert(c->intReq);
 assert(gamegear_Machine_In(machine,0xbf)&128);assert(!(v.status&128)&&!c->intReq);
 rrgg::tick(machine,20*228);assert(!(v.status&128)); // No repeated vblank flag.
 rrgg::tick(machine,6*228);assert(gamegear_Machine_In(machine,0x7e)==0xda);rrgg::tick(machine,228);assert(gamegear_Machine_In(machine,0x7e)==0xd5);
 reg(9,37);assert(rrgg::timing.scrollY==0);reg(8,19);assert(rrgg::timing.scrollX==0);rrgg::tick(machine,228);assert(rrgg::timing.scrollX==19&&rrgg::timing.scrollY==0);
 rrgg::tick(machine,41*228);assert(v.line==255&&rrgg::timing.scrollY==37);rrgg::tick(machine,228);assert(v.line==0);
 reg(9,81);rrgg::tick(machine,100*228);assert(rrgg::timing.scrollY==37);

 // Palette data is held until the odd write commits a full 12-bit color.
 gamegear_Machine_Out(machine,0xbf,0);gamegear_Machine_Out(machine,0xbf,0xc0);gamegear_Machine_Out(machine,0xbe,0xab);assert(v.CRAM[0]==0);gamegear_Machine_Out(machine,0xbe,0xfe);assert(v.CRAM[0]==0xab&&v.CRAM[1]==14);
 gamegear_Machine_Out(machine,0xbf,5);gamegear_Machine_Out(machine,0xbf,0xc0);gamegear_Machine_Out(machine,0xbe,3);assert(v.CRAM[4]==0xab&&v.CRAM[5]==3);
 // Mirrored VDP ports, 14-bit address wrapping, read-ahead and latch reset.
 gamegear_Machine_Out(machine,0x81,255);gamegear_Machine_Out(machine,0x81,0x7f);gamegear_Machine_Out(machine,0x80,0x5a);assert(v.addr==0&&v.VRAM[16383]==0x5a);gamegear_Machine_Out(machine,0x80,0x6b);
 gamegear_Machine_Out(machine,0xbf,255);gamegear_Machine_Out(machine,0xbf,0x3f);assert(gamegear_Machine_In(machine,0xbe)==0x5a);assert(gamegear_Machine_In(machine,0xbe)==0x6b);
 gamegear_Machine_Out(machine,0xbf,0x44);assert(v.latched);gamegear_Machine_In(machine,0xbf);assert(!v.latched);
 // Save pending IRQs, scroll latches, cycle phase and the incomplete CRAM word.
 reg(0,0x14);rrgg::timing.linePending=true;rrgg::irq(machine);assert(rr_state_save());auto saved=stateOutput;auto old=rrgg::timing;
 rrgg::tick(machine,43);rrgg::timing.cramLow=9;memcpy(rr_state_input(saved.size()),saved.data(),saved.size());assert(rr_state_load(saved.size()));assert(rr_state_save()&&stateOutput==saved);assert(rrgg::timing.cycles==old.cycles&&machine->CPU->intReq);
 std::cout<<"Game Gear Z80 cycles/prefixes/HALT/EI, VDP interrupts/counters, scroll latches, atomic CRAM, mirrored ports and timing states pass\n";
}
