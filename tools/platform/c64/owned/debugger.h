#pragma once
#include "system.h"
#include <deque>
#include <optional>
#include <string>
#include <vector>
namespace rr::c64 {
struct Disassembly {
 uint16_t address=0;
 uint8_t length=1;
 std::array<uint8_t,3> bytes{};
 Op op=Op::Unknown;
 Mode mode=Mode::imp;
 std::string text;
};
struct Breakpoint {Processor processor;uint16_t address;};
enum class StopReason { Paused, Budget, Breakpoint, Step, Fault };
struct DebugStop {StopReason reason=StopReason::Paused;Processor processor=Processor::C64;uint16_t address=0;uint64_t edges=0;};
// Single-threaded execution controller. The host yields between bounded runs;
// while paused, no CPU, device, tape or rotating disk clock advances.
class Debugger {
public:
 explicit Debugger(System& system,size_t traceCapacity=4096):machine(system),capacity(traceCapacity){}
 void pause(){paused=true;}
 void resume();
 bool isPaused()const{return paused;}
 void breakpoint(Processor processor,uint16_t address,bool enabled=true);
 const std::vector<Breakpoint>& breakpoints()const{return points;}
 DebugStop run(uint64_t maxEdges);
 DebugStop step(Processor processor,uint64_t maxEdges=1000000);
 const CpuState& cpu(Processor processor)const;
 uint8_t peek(Processor processor,uint16_t address)const;
 Disassembly disassemble(Processor processor,uint16_t address)const;
 const DriveState& drive()const{return machine.drive.state;}
 const IecState& iec()const{return machine.iec;}
 const std::deque<IecEvent>& events()const{return history;}
 uint64_t droppedEvents()const{return dropped;}
 void clearEvents(){history.clear();dropped=0;}
 // Restores emulated state, retains configured breakpoints, pauses, and starts
 // a fresh trace. It must not retain a stale breakpoint bypass from the future.
 void restore(const SystemSnapshot& snapshot);
private:
 System& machine;
 bool paused=true;
 size_t capacity;
 uint64_t dropped=0;
 std::deque<IecEvent> history;
 std::vector<Breakpoint> points;
 struct Bypass {Processor processor;uint16_t address;uint64_t retired;};
 std::optional<Bypass> bypass;
 DebugStop last;
 void ignoreCurrent(Processor processor);
 bool hit(Processor processor);
 DebugStop execute(uint64_t maxEdges,std::optional<Processor> stepping);
};
}
