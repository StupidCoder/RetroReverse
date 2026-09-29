#pragma once
#include "state-fields.h"
namespace rrhh {inline void stateFields(rrstate::Archive&a,Video&v){a(v.draw,v.screen,v.frames,v.steps,v.phase,v.offDots,v.windowLine,v.lastLine,v.statIRQ,v.pc);}}
namespace rrgg {inline void stateFields(rrstate::Archive&a,Timing&v){a(v.cycles,v.scrollX,v.scrollY,v.lineCounter,v.cramLow,v.linePending,v.started);}}
static std::vector<std::shared_ptr<void>>stateOwners;
static void stateWrite(rrstate::Archive&a){a.header(9,2);a(machine,rrhh::video,rrgg::timing);}
static void stateRead(rrstate::Archive&a){a.header(9,2);Machine*next=nullptr;rrhh::Video video;rrgg::Timing timing;a(next,video,timing);a.finish();
 if(!next||!next->CPU||next->nbanks!=machine->nbanks||video.phase>=rrgg::frameCycles||video.phase!=timing.cycles%rrgg::frameCycles||next->CPU->IM>2||next->VDP.addr>=0x4000||next->VDP.code>3)throw std::runtime_error("Invalid Game Gear timing/state");
 for(auto bank:next->slot)if(bank<0||bank>=next->nbanks)throw std::runtime_error("Invalid Sega mapper state");
 next->rom=machine->rom;next->CPU->bus=next;machine=next;rrhh::video=video;rrgg::timing=timing;rrgg::instructionCycles=rrgg::ioCycles=rrgg::elapsedCycles=0;rrgg::begin();stateOwners=std::move(a.owned);
}
