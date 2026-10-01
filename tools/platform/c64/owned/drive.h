#pragma once
#include "cpu.h"
#include "via.h"
#include "disk.h"
#include <array>
namespace rr::c64 {
struct IecLines {bool atn=true,clock=true,data=true;}; // True is released/high.
struct DriveState {
 CpuState cpu;
 Bus lastBus;
 Via serial,disk;
 std::array<uint8_t,2048> ram{};
 Disk media;
 uint64_t clocks=0,bytes=0,steps=0,writtenBits=0;
 uint32_t rotation=0,position=0,mediaChange=0;
 uint8_t halfTrack=36,phase=0,mediaPhase=0,lastHalf=36,bitPhase=0,bits=0,ones=0,readShift=0,writeShift=0,bus=255,device=8;
 bool sync=false,byteReady=false,writing=false,motor=false,led=false;
};
class Drive {
public:
 DriveState state;
 std::array<uint8_t,16384> rom{};
 void power();
 bool mount(std::span<const uint8_t> image,bool writeProtected=true);
 void eject();
 uint8_t peek(uint16_t address)const;
 uint8_t read(uint16_t address);
 void write(uint16_t address,uint8_t value);
 bool tick(IecLines lines);
 bool dataPull(bool atn)const;
 bool clockPull()const;
private:
 void mechanics();
};
}
