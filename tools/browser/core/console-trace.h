#pragma once
#include "capture.h"
// Coalesce consecutive byte stores into their original 32-bit word. Flush at
// every command boundary so writers from different commands never merge.
namespace rrconsole {
inline uint32_t pending=UINT32_MAX,value=0,event=0,pc=0;inline uint64_t clock=0;
inline void flush(){if(pending!=UINT32_MAX){rrcapture::trace.record(pending,value,4,clock,pc,event);pending=UINT32_MAX;}}
inline void byte(uint32_t at,uint8_t v,uint64_t ticks,uint32_t instruction){auto&t=rrcapture::trace;const auto word=at&~3u;if(pending!=word||event!=t.current){flush();pending=word;value=t.value(t.shadow,word,4);event=t.current;clock=ticks;pc=instruction;}const auto shift=(at&3)*8;value=(value&~(255u<<shift))|uint32_t(v)<<shift;}
inline uint32_t command(uint64_t clock,uint32_t pc,std::string detail,uint32_t resource=0){flush();return rrcapture::trace.event(clock,pc,std::move(detail),resource);}
struct EventScope {uint32_t previous;explicit EventScope(uint32_t p):previous(p){}~EventScope(){flush();rrcapture::trace.current=previous;}};
}
namespace rrconsole {inline std::string escape(const std::string&s){std::string o;for(auto c:s){if(c=='"'||c=='\\'){o+='\\';o+=c;}else if(uint8_t(c)<32)o+=' ';else o+=c;}return o;}}
