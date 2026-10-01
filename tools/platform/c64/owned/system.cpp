#include "system.h"
namespace rr::c64 {
void System::resolve(uint64_t time,uint8_t actor){
 const auto port=board.state.cia2.pra&board.state.cia2.ddra;
 const IecLines next{!(port&8),!(port&16)&&!drive.clockPull(),!(port&32)&&!drive.dataPull(!(port&8))};
 const uint8_t before=(iec.lines.atn?1:0)|(iec.lines.clock?2:0)|(iec.lines.data?4:0),after=(next.atn?1:0)|(next.clock?2:0)|(next.data?4:0);
 iec.hostPulls=uint8_t((port>>3)&7);iec.drivePulls=uint8_t((drive.clockPull()?2:0)|(drive.dataPull(next.atn)?4:0));iec.lines=next;
 if(before!=after){++iec.transitions;iec.last={time,actor,uint8_t(before^after),after,iec.hostPulls,iec.drivePulls,board.state.cpu.pc,drive.state.cpu.pc,drive.state.halfTrack,drive.state.position};}
}
void System::power(){board.power();drive.power();iec={};resolve(0,0);}
Processor System::nextProcessor()const{return (drive.state.clocks+1)*985248<=(board.state.cycles+1)*1000000?Processor::Drive:Processor::C64;}
bool System::tickEdge(){
 // CIA/VIA peripheral inputs are sampled at their PHI2 rising read phase,
 // half a clock before the CPU consumes the read. Preserve each sample
 // across intervening peer edges (MOS 6526 read diagram; MOS 6522 Fig. 21).
 // This digital phase model does not simulate analog propagation delays.
 const auto edge=nextProcessor()==Processor::Drive?(drive.state.clocks+1)*985248:(board.state.cycles+1)*1000000;
 if(iec.hostSampleCycle!=board.state.cycles+1&&board.state.cycles*1000000+500000<=edge){
  iec.hostSampleCycle=board.state.cycles+1;
  board.state.iecInputs=uint8_t(0x3f|(iec.lines.clock?64:0)|(iec.lines.data?128:0));
 }
 if(iec.driveSampleCycle!=drive.state.clocks+1&&drive.state.clocks*985248+492624<=edge){iec.driveSampleCycle=drive.state.clocks+1;iec.driveInputs=iec.lines;}
 if(nextProcessor()==Processor::Drive){if(!drive.tick(iec.driveInputs))return false;resolve(drive.state.clocks*985248,1);}
 else{if(!board.tick())return false;resolve(board.state.cycles*1000000,0);}return true;
}
bool System::tick(){const auto until=board.state.cycles+1;while(board.state.cycles<until)if(!tickEdge())return false;return true;}
bool System::run(uint64_t cycles){while(cycles--)if(!tick())return false;return true;}
}
