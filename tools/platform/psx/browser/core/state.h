#pragma once
#include "../../../../browser/state/archive.h"
inline void stateFields(rrstate::Archive&a,CPU&v){a(v.r,v.out,v.cop0,v.hi,v.lo,v.pc,v.next,v.cur,v.ldReg,v.ldVal,v.branchAddr,v.delay,v.pendingDelay,v.halted,v.steps,v.error);}
inline void stateFields(rrstate::Archive&a,GTE&v){a(v.data,v.ctrl);}
inline void stateFields(rrstate::Archive&a,GPU&v){a(v.vram,v.fifo,v.need,v.imgX,v.imgY,v.imgW,v.imgH,v.imgCurX,v.imgCurY,v.imgPx,v.rdX,v.rdY,v.rdW,v.rdH,v.rdCurX,v.rdCurY,v.rdPx,v.drawL,v.drawT,v.drawR,v.drawB,v.offX,v.offY,v.texPageX,v.texPageY,v.texDepth,v.texWinMX,v.texWinMY,v.texWinOX,v.texWinOY,v.dispX,v.dispY,v.dispW,v.dispH,v.dispEnabled,v.gp0Read,v.statField,v.commands,v.words,v.pixels,v.commandHash);}
inline void stateFields(rrstate::Archive&a,CD::Pending&v){a(v.delay,v.cause,v.response,v.read);}
inline void stateFields(rrstate::Archive&a,CD&v){a(v.index,v.stat,v.mode,v.irqEnable,v.irqFlags,v.params,v.data,v.response);uint64_t pos=v.dataPos;a(pos);v.dataPos=pos;a(v.loc,v.readLBA,v.reading,v.queue,v.commands);}
inline void stateFields(rrstate::Archive&a,Machine::ISR&v){a(v.active,v.pc,v.r,v.hi,v.lo);}
inline void stateFields(rrstate::Archive&a,Machine&v){a(v.cpu,v.gte,v.gpu,v.cd,v.ram,v.scratch,v.io,v.irqStat,v.irqMask,v.timer,v.dmaFlags,v.vblankAcc,v.fields,v.isrChain,v.isrHandler,v.isr,v.heapPtr,v.heapEnd,v.nextEvent,v.randSeed,v.padBuf,v.padActive,v.buttons,v.tty,v.diagnostics);}
inline void validateState(Machine&m){if(m.ram.size()!=2*1024*1024||m.scratch.size()!=1024||m.gpu.vram.size()!=1024*512||m.gpu.fifo.size()>32||m.cd.queue.size()>256||m.cd.dataPos>m.cd.data.size()||m.gpu.dispW<1||m.gpu.dispW>1024||m.gpu.dispH<1||m.gpu.dispH>512)throw std::runtime_error("Invalid PS1 machine state");}
