#pragma once
#include "state-fields.h"
namespace rrhh {inline void stateFields(rrstate::Archive&a,Video&v){a(v.draw,v.screen,v.frames,v.steps,v.phase,v.offDots,v.windowLine,v.lastLine,v.statIRQ,v.pc);}}
static std::vector<std::shared_ptr<void>>stateOwners;
static void stateWrite(rrstate::Archive&a){a.header(9,1);a(machine,rrhh::video);}
static void stateRead(rrstate::Archive&a){a.header(9,1);Machine*next=nullptr;rrhh::Video video;a(next,video);a.finish();
 if(!next||!next->CPU||next->nbanks!=machine->nbanks||video.windowLine<0||video.windowLine>144||video.offDots>=70224||video.phase>=30000)throw std::runtime_error("Invalid handheld state");
 for(auto bank:next->slot)if(bank<0||bank>=next->nbanks)throw std::runtime_error("Invalid Sega mapper state");
 next->rom=machine->rom;next->CPU->bus=next;machine=next;rrhh::video=video;stateOwners=std::move(a.owned);
}
