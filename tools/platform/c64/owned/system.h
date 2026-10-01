#pragma once
#include "board.h"
#include "drive.h"
namespace rr::c64 {
enum class Processor : uint8_t { C64, Drive };
struct IecEvent {
 uint64_t time=0;
 uint8_t actor=0,changed=0,levels=7,hostPulls=0,drivePulls=0;
 uint16_t hostPC=0,drivePC=0; // PCs after the causative clock edge.
 uint8_t halfTrack=0;
 uint32_t bit=0;
};
struct IecState {IecLines lines,driveInputs;uint8_t hostPulls=0,drivePulls=0;uint64_t transitions=0,hostSampleCycle=0,driveSampleCycle=0;IecEvent last;};
struct SystemSnapshot {BoardState board;DriveState drive;IecState iec;};
// One PAL C64 + one 1 MHz 1541. Time units are 1/(985248*1000000) s.
// Each drive edge is ordered before the next C64 edge, including double ticks.
class System {
public:
 Board board;
 Drive drive;
 IecState iec;
 void power();
 Processor nextProcessor()const;
 bool tickEdge(); // Advance exactly the next clock edge on the shared timeline.
 bool tick();
 bool run(uint64_t cycles);
 SystemSnapshot save()const{return {board.state,drive.state,iec};}
 void restore(const SystemSnapshot& s){board.state=s.board;drive.state=s.drive;iec=s.iec;}
private:
 void resolve(uint64_t time,uint8_t actor);
};
}
