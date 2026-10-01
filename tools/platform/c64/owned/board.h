#pragma once
#include "cpu.h"
#include "cia.h"
#include <array>
#include <span>
#include <vector>
namespace rr::c64 {
struct TapeState {uint32_t pulse=0,remaining=0;bool play=false,flag=true;};
struct BoardState {
 CpuState cpu;
 Cia cia1,cia2;
 std::array<uint8_t,65536> ram{};
 std::array<uint8_t,1024> color{};
 std::array<uint8_t,64> vic{};
 std::array<uint8_t,32> sid{};
 std::array<uint8_t,8> keys{}; // Each entry is one PA column; bits are PB rows.
 uint8_t ddr=0,data=0,bus=0xff,joy1=0,joy2=0,vicFlags=0,vicMask=0;
 uint16_t raster=0,rasterCompare=0;
 uint8_t rasterCycle=0;
 uint32_t todPhase=0;
 uint64_t cycles=0;
 TapeState tape;
 bool restore=false,unimplementedSidRead=false;
};
class Board {
public:
 BoardState state;
 std::array<uint8_t,8192> basic{},kernal{};
 std::array<uint8_t,4096> chars{};
 std::vector<uint32_t> pulses;
 void power();
 void resetCpu();
 bool loadTape(std::span<const uint8_t> bytes);
 uint8_t port()const;
 uint8_t peek(uint16_t address)const;
 uint8_t read(uint16_t address);
 void write(uint16_t address,uint8_t value);
 bool tick(bool cpuReady=true);
 bool run(uint64_t cycles);
 void key(unsigned column,unsigned row,bool down);
 void play(bool value){state.tape.play=value;}
 bool motor()const{return !(port()&32);}
private:
 std::array<uint8_t,2> keyboard()const;
 uint8_t io(uint16_t address,bool sideEffects);
 bool ioVisible()const;
 void rasterTick();
};
}
