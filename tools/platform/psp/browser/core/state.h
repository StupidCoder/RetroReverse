#pragma once
#include "../../../../browser/state/translated.h"
#include "state-fields.h"
inline void rebindState(psp_Machine*m,psp_Volume*volume){
 if(!m||!m->CPU||m->vram.n!=2097152||m->scratch.n!=16384||m->fbFormat>3||m->fbWidth>4096)throw std::runtime_error("Invalid PSP machine state");
 m->vol=volume;m->CPU->bus=m;m->CPU->Syscall=[m](auto...args){return psp_Machine_handleSyscall(m,args...);};
 for(auto&[id,call]:*m->syscalls.p){if(!call)throw std::runtime_error("Invalid syscall state");call->handler=psp_handlerFor(call->name);}
 rrBind(m);
}
