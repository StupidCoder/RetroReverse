#pragma once
#include <array>
#include <cstdint>
namespace rr::c64 {
enum class Envelope : uint8_t { Attack,Decay,Release };
struct SidVoice {
 uint32_t phase=0,noise=0x7ffff8;
 uint16_t rate=0,testCycles=0;
 uint8_t envelope=0,exponential=0,divider=1,noiseDelay=0;
 Envelope stage=Envelope::Release;
 bool zero=true;
};
// Digital register-visible SID model. No audio/filter synthesis. Combined
// waveforms and analog TEST discharge use an explicit approximation (see ledger).
struct Sid {
 std::array<uint8_t,32> regs{};
 std::array<SidVoice,3> voice{};
 uint8_t bus=0;
 uint32_t busAge=0;
 bool combinedWaveformUsed=false;
 void tick();
 void write(uint8_t reg,uint8_t value);
 uint8_t peek(uint8_t reg)const;
 uint8_t read(uint8_t reg);
 uint16_t wave(unsigned index)const;
};
}
