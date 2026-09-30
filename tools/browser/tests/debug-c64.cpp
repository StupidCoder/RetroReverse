#include "../../platform/c64/browser/core/core.cpp"
#include <cassert>
#include <iostream>
static void setup(){
 memset(rr_input(),0,20480);assert(rr_init(8192,8192,4096));
 memcpy(rr_input(),"C64-TAPE-RAW",12);memset(rr_input()+12,0,9);rr_input()[16]=1;rr_input()[20]=1;assert(rr_tape(21));
 memset(rr_input(),0xea,65536);rr_input()[0x800]=0xee;rr_input()[0x801]=0;rr_input()[0x802]=2;rr_input()[0x803]=0x4c;rr_input()[0x804]=0;rr_input()[0x805]=8;rr_input()[0x200]=0;
 assert(rr_prepare(0x800,0));
}
static int pc(){return M6502_GET_ADDR(machine.pins);}
int main(){
 setup();auto cycle=ctx.cycles;auto pins=machine.pins;const c64_t original=machine;rr_debug_snapshot(0xdc00);assert(ctx.cycles==cycle&&machine.pins==pins&&memcmp(&original,&machine,sizeof(machine))==0);
 assert(std::string(rr_debug_snapshot(0xdc00)).find("\"bytes\":[null")!=std::string::npos);
 assert(rr_debug_begin(2,0x800,0));assert(rr_debug_run(100)==4);assert(ctx.cycles==0&&machine.ram[0x200]==0);
 assert(rr_debug_begin(1,0,0));assert(rr_debug_run(100)==2);assert(pc()==0x803&&machine.ram[0x200]==1);assert(ctx.cycles==6);
 assert(rr_debug_begin(2,0x800,0));assert(rr_debug_run(100)==4);assert(machine.ram[0x200]==1);
 assert(rr_debug_begin(2,0x800,1));assert(rr_debug_run(1)==0);assert(rr_debug_run(100)==4);assert(machine.ram[0x200]==2);
 // Taken branch, then self-modified immediate, fetched bytes reflect writes.
 setup();machine.ram[0x800]=0xd0;machine.ram[0x801]=2;M6502_SET_DATA(machine.pins,0xd0);
 assert(rr_debug_begin(1,0,0));assert(rr_debug_run(100)==2);assert(pc()==0x804);
 machine.ram[0x810]=0x55;machine.ram[0x811]=0x20;assert(std::string(rr_debug_snapshot(0x810)).find("\"bytes\":[85,32")!=std::string::npos);
 // Execute a real self-modifying sequence and decode the updated fetch.
 setup();const uint8_t modify[]={0xa9,0x55,0x8d,0x10,0x08,0x4c,0x10,0x08};memcpy(machine.ram+0x800,modify,sizeof(modify));M6502_SET_DATA(machine.pins,0xa9);
 assert(rr_debug_begin(2,0x810,0));assert(rr_debug_run(100)==4);assert(machine.ram[0x810]==0x55);assert(M6502_GET_DATA(machine.pins)==0x55);
 // JSR and RTS each retire once, including stack effects.
 setup();machine.ram[0x800]=0x20;machine.ram[0x801]=0;machine.ram[0x802]=9;machine.ram[0x900]=0x60;M6502_SET_DATA(machine.pins,0x20);
 assert(rr_debug_begin(1,0,0));assert(rr_debug_run(100)==2);assert(pc()==0x900&&machine.cpu.S==0xfb);
 assert(rr_debug_begin(1,0,0));assert(rr_debug_run(100)==2);assert(pc()==0x803&&machine.cpu.S==0xfd);
 // Partial instruction cannot be silently stepped/normalized by inspection.
 setup();rr_run(1,0,0);cycle=ctx.cycles;assert(!rr_debug_begin(1,0,0));rr_debug_snapshot(-1);assert(ctx.cycles==cycle);assert(rr_debug_begin(0,0,0));assert(rr_debug_run(100)==1);assert(pc()==0x803);
 // IRQ entry is separate from retirement and stops before first handler opcode.
 setup();ctx.synthetic=false;machine.rom_kernal[0x1ffe]=0;machine.rom_kernal[0x1fff]=9;machine.cpu.irq_pip=0x400;machine.cpu.P=0;
 assert(rr_debug_begin(1,0,0));assert(rr_debug_run(100)==3);assert(pc()==0x900&&machine.ram[0x200]==0);
 setup();ctx.synthetic=false;machine.rom_kernal[0x1ffa]=0;machine.rom_kernal[0x1ffb]=9;machine.cpu.nmi_pip=0x400;
 assert(rr_debug_begin(1,0,0));assert(rr_debug_run(100)==3);assert(pc()==0x900&&machine.ram[0x200]==0);
 // Safe peeks use the current bank, never physical RAM under visible ROM.
 machine.ram[0xa000]=0x55;machine.rom_basic[0]=0xaa;assert(std::string(rr_debug_snapshot(0xa000)).find("\"bytes\":[170")!=std::string::npos);
 _c64_cpu_port_out(0x34,&machine);assert(std::string(rr_debug_snapshot(0xa000)).find("\"bytes\":[85")!=std::string::npos);
 // A target bound to visible ROM cannot silently match RAM after bank-out.
 setup();const uint8_t bankout[]={0xa9,0x34,0x85,0x01,0x4c,0,0xa0};memcpy(machine.ram+0x800,bankout,sizeof(bankout));M6502_SET_DATA(machine.pins,0xa9);
 assert(rr_debug_begin(2,0xa000,0));assert(rr_debug_run(100)==5);assert(pc()==0xa000);
 // RDY stall cannot be counted as a completed instruction.
 setup();machine.pins|=M6502_RDY;assert(rr_debug_begin(1,0,0));assert(rr_debug_run(1)==0);assert(machine.ram[0x200]==0);assert(rr_debug_run(100)==2);assert(machine.ram[0x200]==1);
 // JAM must remain bounded rather than hanging the worker.
 setup();machine.ram[0x800]=2;M6502_SET_DATA(machine.pins,2);assert(rr_debug_begin(1,0,0));assert(rr_debug_run(100)==6);
 std::cout<<"C64 debug boundary, target, branch, interrupt, banking, stall and JAM: PASS\n";
 // Checked batches validate every expected byte and mapped address before mutation.
 setup();uint8_t edits[]={0x00,0x08,0xee,0xce,0x01,0x02,0xff,9};memcpy(rr_input(),edits,8);
 assert(!rr_debug_edit(2)&&machine.ram[0x800]==0xee);
 rr_input()[6]=machine.ram[0x201];assert(rr_debug_edit(2)&&machine.ram[0x201]==9);
 assert(rr_debug_begin(1,0,0)&&rr_debug_run(100)==2&&machine.ram[0x200]==255); // patched prefetched DEC
 setup();rr_input()[0]=0;rr_input()[1]=0xdc;rr_input()[2]=machine.ram[0xdc00];rr_input()[3]=0;assert(!rr_debug_edit(1));
 assert(!rr_debug_edit(257));rr_run(1,0,0);assert(!rr_debug_edit(1));
 // Nested calls: over returns to the caller; out stops at the matching RTS.
 setup();const uint8_t caller[]={0x20,0,9,0xea};memcpy(machine.ram+0x800,caller,sizeof(caller));
 const uint8_t callee[]={0x20,0,10,0x60};memcpy(machine.ram+0x900,callee,sizeof(callee));machine.ram[0xa00]=0x60;M6502_SET_DATA(machine.pins,0x20);
 assert(rr_debug_begin(3,0,0)&&rr_debug_run(100)==12&&pc()==0x803&&machine.cpu.S==0xfd);
 assert(rr_debug_begin(3,0,0)&&rr_debug_run(100)==2&&pc()==0x804); // non-call over is one step
 setup();memcpy(machine.ram+0x800,caller,sizeof(caller));memcpy(machine.ram+0x900,callee,sizeof(callee));machine.ram[0xa00]=0x60;M6502_SET_DATA(machine.pins,0x20);
 assert(rr_debug_begin(1,0,0)&&rr_debug_run(100)==2&&pc()==0x900);
 assert(rr_debug_begin(4,0,0)&&rr_debug_run(100)==12&&pc()==0x803);
 assert(!rr_debug_begin(4,0,0)); // no conventional return frame
 // A hardware interrupt inside a subroutine must not count its RTS as our return.
 setup();ctx.synthetic=false;memcpy(machine.ram+0x800,caller,sizeof(caller));machine.ram[0x900]=0xea;machine.ram[0x901]=0x60;M6502_SET_DATA(machine.pins,0x20);
 assert(rr_debug_begin(3,0,0));assert(rr_debug_run(6)==0&&pc()==0x900);
 machine.rom_kernal[0x1ffe]=0;machine.rom_kernal[0x1fff]=10;machine.cpu.irq_pip=0x400;machine.cpu.P=0;
 const uint8_t handler[]={0x20,0,11,0x40};memcpy(machine.ram+0xa00,handler,sizeof(handler));machine.ram[0xb00]=0x60;
 assert(rr_debug_run(200)==12&&pc()==0x803&&machine.cpu.S==0xfd);
 // RMW's dummy same-value write is observable, with exact writer attribution.
 setup();assert(!rr_debug_watch(0xdc00,0,0,-1));assert(!rr_debug_watch(0x200,2,1,-1));
 assert(rr_debug_watch(0x200,0,255,0x800)&&rr_debug_run(100)==13);
 assert(!debug::boundary()&&machine.ram[0x200]==0);
 assert(std::string(rr_debug_event()).find("\"pc\":2048")!=std::string::npos);
 assert(rr_debug_watch(0x200,1,255,0x800)&&rr_debug_run(100)==13&&machine.ram[0x200]==1);
 assert(rr_debug_begin(0,0,0)&&rr_debug_run(100)==1);
 assert(rr_debug_watch(0x200,0,0,0x900)&&rr_debug_run(100)==0); // writer mismatch
 std::cout<<"C64 nested over/out and masked actual-write watchpoints: PASS\n";

}
