#pragma once
template<class...A>void sm83_CPU_Halt(sm83_CPU*c,std::string f,A...args){c->Halted=true;c->HaltReason=go_fmt_Sprintf(f,args...);}
void rrAfterWrite(gameboy_Machine*,uint16_t,uint8_t);
