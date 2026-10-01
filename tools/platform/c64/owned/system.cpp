#include "system.h"
namespace rr::c64 {
void System::resolve(uint64_t time,uint8_t actor){
 const auto port=board.state.cia2.pra&board.state.cia2.ddra;
 const IecLines next{!(port&8),!(port&16)&&!drive.clockPull(),!(port&32)&&!drive.dataPull(!(port&8))};
 const uint8_t before=(iec.lines.atn?1:0)|(iec.lines.clock?2:0)|(iec.lines.data?4:0),after=(next.atn?1:0)|(next.clock?2:0)|(next.data?4:0);
 iec.hostPulls=uint8_t((port>>3)&7);iec.drivePulls=uint8_t((drive.clockPull()?2:0)|(drive.dataPull(next.atn)?4:0));iec.lines=next;
 board.state.iecInputs=uint8_t(0x3f|(next.clock?64:0)|(next.data?128:0));
 if(before!=after){++iec.transitions;iec.last={time,actor,uint8_t(before^after),after};}
}
void System::power(){board.power();drive.power();iec={};resolve(0,0);}
bool System::tick(){const uint64_t until=(board.state.cycles+1)*1000000;
 while((drive.state.clocks+1)*985248<=until){if(!drive.tick(iec.lines))return false;resolve(drive.state.clocks*985248,1);}
 if(!board.tick())return false;resolve(until,0);return true;
}
bool System::run(uint64_t cycles){while(cycles--)if(!tick())return false;return true;}
}
