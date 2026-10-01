#pragma once
#include "../state.h"
#include "../../../../browser/state/archive.h"

// This field order IS format version 1. Additions/reordering require a version
// change; no compiler layout, padding, size_t, pointers or host endianness leak.
namespace rr::c64 {
using rrstate::Archive;
inline void stateFields(Archive& a,StateIdentity& s){a(s.core,s.basic,s.kernal,s.characters,s.driveRom,s.tape,s.disk,s.configuration);}
inline void stateFields(Archive& a,Bus& s){a(s.address,s.data,s.write,s.sync);}
inline void stateFields(Archive& a,CpuState& s){
 a(s.pc,s.address,s.target,s.vector,s.a,s.x,s.y,s.s,s.p,s.opcode,s.low,s.pointer,s.value,s.op,s.mode,s.stage,s.bus,s.clocks,s.retired,
   s.soLine,s.nmiLine,s.nmiPending,s.interruptPending,s.softwareInterrupt,s.irq,s.pollIrq,s.pollNmi);
}
inline void stateFields(Archive& a,CiaTimer& s){a(s.latch,s.counter,s.control,s.queue,s.loadDelay,s.inhibit,s.previousOneShot,s.toggle,s.pulse);}
inline void stateFields(Archive& a,Cia& s){
 a(s.a,s.b,s.pra,s.prb,s.ddra,s.ddrb,s.flags,s.mask,s.sdr,s.shift,s.bits,s.irq,s.irqDelay,s.flagLine,s.cntLine,
   s.serialPending,s.serialActive,s.serialClock,s.serialData,s.tod,s.alarm,s.todLatch,s.todDivider,s.todStopped,s.todLatched);
}
inline void stateFields(Archive& a,VicFetch& s){a(s.cycle,s.address,s.value,s.color,s.slot,s.phase,s.kind,s.rom);}
inline void stateFields(Archive& a,VicCell& s){a(s.matrix,s.graphics,s.code,s.color);}
inline void stateFields(Archive& a,VicSprite& s){a(s.bytes,s.shift,s.pointer,s.mc,s.base,s.pixel,s.repeat,s.dma,s.display,s.advance,s.active);}
inline void stateFields(Archive& a,Vic& s){
 a(s.regs,s.pixels,s.cells,s.sprites,s.fetches,s.clocks,s.frames,s.raster,s.compare,s.vc,s.base,s.cycle,s.row,s.index,s.flags,s.mask,
   s.collisionSprites,s.collisionGraphics,s.refresh,s.fetchCount,s.den,s.bad,s.display,s.verticalBorder,s.border,s.ba,s.aec,s.irqMatch);
}
inline void stateFields(Archive& a,SidVoice& s){a(s.phase,s.noise,s.rate,s.testCycles,s.envelope,s.exponential,s.divider,s.noiseDelay,s.stage,s.zero);}
inline void stateFields(Archive& a,Sid& s){a(s.regs,s.voice,s.bus,s.busAge,s.combinedWaveformUsed);}
inline void stateFields(Archive& a,TapeState& s){a(s.pulse,s.remaining,s.play,s.flag);}
inline void stateFields(Archive& a,CpuBusSample& s){a(s.cycle,s.address,s.pc,s.value,s.valid,s.write,s.ready,s.sync);}
inline void stateFields(Archive& a,BoardState& s){
 a(s.lastBus,s.cpu,s.cia1,s.cia2,s.ram,s.color,s.vic,s.sid,s.keys,s.ddr,s.data,s.bus,s.joy1,s.joy2,s.iecInputs,s.todPhase,s.cycles,s.tape,s.restore);
}
inline void stateFields(Archive& a,Via& s){
 a(s.ora,s.orb,s.ddra,s.ddrb,s.acr,s.pcr,s.ifr,s.ier,s.sr,s.pa,s.pb,s.latchA,s.latchB,s.shiftBits,
   s.t1,s.t1Latch,s.t2,s.t2Low,s.t1Delay,s.t2Delay,s.shiftTimer,s.t1Armed,s.t2Armed,s.t1Reload,s.t1Out,
   s.ca1,s.ca2,s.cb1,s.cb2,s.ca2Out,s.cb2Out,s.pulseA,s.pulseB,s.shiftActive,s.shiftClock);
}
inline void stateFields(Archive& a,Disk& s){
 // Check bounds BEFORE allocation. G64 v0 track lengths are 16-bit; the
 // recovered per-byte speed map must have exactly the same length.
 for(unsigned i=0;i<84;i++)for(auto* v:{&s.tracks[i],&s.speeds[i]}){
  const auto n=a.count(v->size());
  if(n>65535||(v==&s.speeds[i]&&n!=s.tracks[i].size()))throw std::runtime_error("Invalid disk track size");
  if(a.reading){if(n>a.bytes.size()-a.pos)throw std::runtime_error("Truncated disk track");a.charge(n);v->resize(n);}
  if(n)a.raw(v->data(),n);
 }
 uint32_t tracks=s.trackCount;a(s.dirty,tracks,s.g64,s.writeProtected);s.trackCount=tracks;
}
inline void stateFields(Archive& a,DriveState& s){
 a(s.cpu,s.lastBus,s.serial,s.disk,s.ram,s.media,s.clocks,s.bytes,s.steps,s.writtenBits,s.rotation,s.position,s.mediaChange,
   s.halfTrack,s.phase,s.mediaPhase,s.lastHalf,s.bitPhase,s.bits,s.ones,s.readShift,s.writeShift,s.bus,s.device,
   s.sync,s.byteReady,s.writing,s.motor,s.led);
}
inline void stateFields(Archive& a,IecLines& s){a(s.atn,s.clock,s.data);}
inline void stateFields(Archive& a,IecEvent& s){a(s.time,s.actor,s.changed,s.levels,s.hostPulls,s.drivePulls,s.hostPC,s.drivePC,s.halfTrack,s.bit);}
inline void stateFields(Archive& a,IecState& s){a(s.lines,s.driveInputs,s.hostPulls,s.drivePulls,s.transitions,s.hostSampleCycle,s.driveSampleCycle,s.last);}
inline void stateFields(Archive& a,SystemSnapshot& s){a(s.board,s.drive,s.iec);}
}
