#pragma once
#include "board.h"
#include "drive.h"
namespace rr::c64 {
struct IecEvent {uint64_t time=0;uint8_t actor=0,changed=0,levels=7;};
struct IecState {IecLines lines;uint8_t hostPulls=0,drivePulls=0;uint64_t transitions=0;IecEvent last;};
struct SystemSnapshot {BoardState board;DriveState drive;IecState iec;};
// One PAL C64 + one 1 MHz 1541. Time units are 1/(985248*1000000) s.
// Each drive edge is ordered before the next C64 edge, including double ticks.
class System {
public:
 Board board;
 Drive drive;
 IecState iec;
 void power();
 bool tick();
 bool run(uint64_t cycles);
 SystemSnapshot save()const{return {board.state,drive.state,iec};}
 void restore(const SystemSnapshot& s){board.state=s.board;drive.state=s.drive;iec=s.iec;}
private:
 void resolve(uint64_t time,uint8_t actor);
};
}
