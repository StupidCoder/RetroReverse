#pragma once
template<class...A>void z80_CPU_Halt(z80_CPU*c,std::string f,A...args){c->Halted=true;c->HaltReason=go_fmt_Sprintf(f,args...);}
void rrAfterWrite(gamegear_Machine*,uint16_t,uint8_t);
void rrAfterPort(gamegear_Machine*,uint16_t,uint8_t);

void rrAfterRead(gamegear_Machine*,uint16_t,uint8_t);
