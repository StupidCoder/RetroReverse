#pragma once
template<class...A>void allegrex_CPU_Halt(allegrex_CPU*c,std::string f,A...args){c->Halted=true;c->HaltReason=go_fmt_Sprintf(f,args...);}
template<class...A>void psp_Machine_note(psp_Machine*m,std::string f,A...args){auto s=go_fmt_Sprintf(f,args...);if(m->logSeen.size()<65536&&!get(m->logSeen,s)){m->logSeen[s]=true;if(m->Log.n<4096)m->Log=append(m->Log,s);}}
template<class...A>void psp_fmtPrintf(std::string,A...){ }
allegrex_CPU*allegrex_NewCPU(psp_Machine*);
uint32_t allegrex_CPU_read16(allegrex_CPU*,uint32_t);
uint32_t allegrex_CPU_read32(allegrex_CPU*,uint32_t);
void allegrex_CPU_write16(allegrex_CPU*,uint32_t,uint32_t);
void allegrex_CPU_write32(allegrex_CPU*,uint32_t,uint32_t);
psp_Result psp_Machine_Run(psp_Machine*,uint64_t);
template<class...A>uint32_t psp_Machine_callGuest(psp_Machine*,uint32_t,A...);
std::tuple<Slice<uint8_t>,Error>psp_cbcDecryptZero(Slice<uint8_t>,Slice<uint8_t>);
std::tuple<Slice<uint8_t>,Error>blockSource_ReadBlock(BlockSource*,int64_t);
uint32_t psp_Machine_read32(psp_Machine*,uint32_t);
void psp_Machine_write16(psp_Machine*,uint32_t,uint16_t);
void psp_Machine_write32(psp_Machine*,uint32_t,uint32_t);

inline void(*rrObserveWrite)(psp_Machine*,uint32_t,uint32_t,unsigned)=nullptr;

uint16_t psp_u16(psp_Machine*,uint32_t);
