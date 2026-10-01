#include "../debugger.h"
#include <cassert>
#include <iostream>
#include <memory>
using namespace rr::c64;
static void setup(System& s){
 s.board.kernal[0x1ffc]=0;s.board.kernal[0x1ffd]=2;s.drive.rom[0x3ffc]=0;s.drive.rom[0x3ffd]=2;s.power();
 // Both CPUs run INC $20; JMP $0200.
 for(auto* ram:{s.board.state.ram.data(),s.drive.state.ram.data()}){ram[0x200]=0xe6;ram[0x201]=0x20;ram[0x202]=0x4c;ram[0x203]=0;ram[0x204]=2;}
}
int main(){
 auto machine=std::make_unique<System>();auto& s=*machine;setup(s);Debugger d(s,8);
 assert(d.run(1000).reason==StopReason::Paused&&s.board.state.cycles==0&&s.drive.state.clocks==0);
 d.breakpoint(Processor::Drive,0x200);d.breakpoint(Processor::Drive,0x200);assert(d.breakpoints().size()==1);d.resume();auto stop=d.run(1000);assert(stop.reason==StopReason::Breakpoint&&stop.processor==Processor::Drive&&s.drive.state.ram[0x20]==0&&d.isPaused());
 const auto saved=std::make_unique<SystemSnapshot>(s.save());const auto clocks=s.drive.state.clocks,host=s.board.state.cycles;assert(d.run(1000).edges==0&&s.drive.state.clocks==clocks&&s.board.state.cycles==host);
 const auto text=d.disassemble(Processor::Drive,0x200);assert(text.text=="INC $20"&&text.length==2);
 stop=d.step(Processor::Drive);assert(stop.reason==StopReason::Step&&s.drive.state.ram[0x20]==1&&s.drive.state.cpu.bus.address==0x202&&s.board.state.cycles>host&&d.isPaused());
 d.resume();stop=d.run(1000);assert(stop.reason==StopReason::Breakpoint&&s.drive.state.ram[0x20]==1);d.resume();stop=d.run(1000);assert(stop.reason==StopReason::Breakpoint&&s.drive.state.ram[0x20]==2); // Resume bypass lasts one visit.
 d.restore(*saved);assert(d.isPaused()&&d.events().empty()&&d.breakpoints().size()==1);d.resume();assert(d.run(100).reason==StopReason::Breakpoint);d.breakpoint(Processor::Drive,0x200,false);
 // Peer breakpoints interrupt a step instead of silently executing past them.
 Cpu hostReady;hostReady.start(0x300);s.board.state.cpu=hostReady.state;s.board.state.ram[0x300]=0xea;d.breakpoint(Processor::C64,0x300);stop=d.step(Processor::Drive);assert(stop.reason==StopReason::Breakpoint&&stop.processor==Processor::C64);d.breakpoint(Processor::C64,stop.address,false);stop=d.step(Processor::Drive);assert(stop.reason==StopReason::Step);
 // A breakpoint at the second drive edge between C64 edges cannot be skipped.
 setup(s);d.restore(s.save());s.board.state.cycles=66;s.drive.state.clocks=66;Cpu c;c.start(0x210);c.tick(0xea);s.drive.state.cpu=c.state;s.drive.state.ram[0x210]=0xea;
 d.breakpoint(Processor::Drive,0x211);assert(s.nextProcessor()==Processor::Drive); // Two consecutive drive clocks.
 d.resume();stop=d.run(100);assert(stop.reason==StopReason::Breakpoint&&s.drive.state.cpu.bus.address==0x211&&s.board.state.cycles==66&&s.drive.state.clocks==67);d.breakpoint(Processor::Drive,0x211,false);
 // Live disassembly re-reads self-modified operands; peeks never acknowledge VIA.
 s.drive.state.ram[0x200]=0xad;s.drive.state.ram[0x201]=1;s.drive.state.ram[0x202]=0x1c;assert(d.disassemble(Processor::Drive,0x200).text=="LDA $1C01");s.drive.state.ram[0x201]=0;assert(d.disassemble(Processor::Drive,0x200).text=="LDA $1C00");
 s.drive.state.disk.ifr=0x42;s.drive.state.disk.ier=0x42;d.disassemble(Processor::Drive,0x1c04);d.peek(Processor::Drive,0x1c01);assert(s.drive.state.disk.ifr==0x42);
 s.drive.rom[0x3fff]=0xd0;s.drive.state.ram[0]=0xfd;assert(d.disassemble(Processor::Drive,0xffff).text=="BNE $FFFE");
 // A VIA output change after the CIA read phase is visible on the wire
 // immediately but cannot change the already sampled C64 port value.
 setup(s);d.restore(s.save());s.drive.state.serial.ddrb=0x1a;
 s.drive.state.cpu.stage=Stage::Write;s.drive.state.cpu.bus={0x1800,8,true,false};
 assert(s.tickEdge()&&!s.iec.lines.clock&&(s.board.state.iecInputs&64));
 assert(s.tickEdge()&&(s.board.state.iecInputs&64));assert(s.tickEdge()&&!(s.board.state.iecInputs&64));
 // Conversely the VIA must retain an ATN sample across a later CIA write.
 setup(s);d.restore(s.save());s.board.state.cycles=33;s.drive.state.clocks=33;s.board.state.cia2.ddra=0x38;
 s.board.state.cpu.stage=Stage::Write;s.board.state.cpu.bus={0xdd00,8,true,false};
 assert(s.nextProcessor()==Processor::Drive);assert(s.tickEdge());assert(s.nextProcessor()==Processor::C64);assert(s.tickEdge()&&!s.iec.lines.atn&&s.iec.driveInputs.atn);
 assert(s.tickEdge()&&!(s.drive.state.serial.pb&128));assert(s.tick());assert(s.tickEdge()&&(s.drive.state.serial.pb&128));
 // Bounded history keeps every edge while space permits and reports overflow.
 setup(s);d.restore(s.save());d.resume();for(unsigned i=0;i<20;i++){s.board.state.cia2.ddra=0x38;s.board.state.cia2.pra=i%2?0x38:0;assert(d.run(4).reason==StopReason::Budget);}
 assert(d.events().size()==8&&d.droppedEvents()>0);uint64_t time=0;for(const auto& e:d.events()){assert(e.time>time&&e.changed);time=e.time;}
 d.pause();const auto stopped=s.board.state.cycles;assert(d.run(100).reason==StopReason::Paused&&s.board.state.cycles==stopped);d.clearEvents();assert(d.events().empty()&&!d.droppedEvents());
 // Host stepping uses the same timeline, including VIC RDY/AEC stalls.
 setup(s);d.restore(s.save());Cpu hostStep;hostStep.start(0x300);s.board.state.cpu=hostStep.state;s.board.state.ram[0x300]=0xea;
 s.board.state.vic.raster=0x33;s.board.state.vic.cycle=11;s.board.state.vic.den=true;s.board.state.vic.regs[0x11]=0x13;
 const auto driveBefore=s.drive.state.clocks;assert(d.step(Processor::C64,0).reason==StopReason::Budget&&d.isPaused());
 stop=d.step(Processor::C64);assert(stop.reason==StopReason::Step&&s.board.state.cpu.retired==1&&s.drive.state.clocks>driveBefore+40&&d.isPaused());
 // A small budget can stop mid-instruction; step finishes it, not an extra one.
 setup(s);d.restore(s.save());d.resume();assert(d.run(16).reason==StopReason::Budget);const auto retired=s.drive.state.cpu.retired;stop=d.step(Processor::Drive);assert(stop.reason==StopReason::Step&&s.drive.state.cpu.retired==retired+1);
 s.drive.state.ram[s.drive.state.cpu.pc]=2;d.resume();assert(d.run(100).reason==StopReason::Fault&&d.isPaused());
 std::cout<<"PASS debugger: global pause, dual-clock instruction steps, peer/second-edge breakpoints, resume/restore, live safe disassembly, bounded IEC trace and faults\n";
}
