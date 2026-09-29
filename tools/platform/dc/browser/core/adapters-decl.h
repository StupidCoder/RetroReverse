#pragma once
template<class... A>void arm_CPU_Halt(arm_CPU*c,std::string f,A...args){c->Halted=true;c->HaltReason=go_fmt_Sprintf(f,args...);}
arm_CPU* arm_NewCPU(dc_armBus*);
uint32_t arm_CPU_read16(arm_CPU*,uint32_t);
uint32_t arm_CPU_read32aligned(arm_CPU*,uint32_t);
void arm_CPU_write16(arm_CPU*,uint32_t,uint32_t);
void arm_CPU_write32aligned(arm_CPU*,uint32_t,uint32_t);
arm_CPU*arm_NewCPU(dc_armBus);
sh4_CPU*sh4_NewCPU(dc_Machine*);
uint16_t sh4_CPU_fetchInstr(sh4_CPU*,uint32_t);
template<class...A>void sh4_CPU_Halt(sh4_CPU*c,std::string f,A...a){c->Halted=true;c->HaltReason=go_fmt_Sprintf(f,a...);}
template<class...A>void dc_Machine_logf(dc_Machine*m,std::string f,A...a){m->gaps[go_fmt_Sprintf(f,a...)]++;}
struct dc_Disc{std::string Path;Slice<dc_Track>Tracks;dc_IPBin IP;iso9660_Volume*Vol{};uint64_t size{};dc_Track data;};
std::tuple<Slice<uint8_t>,Error>dc_Disc_ReadBlock(dc_Disc*,int64_t);
inline auto dc_Disc_ReadSector(dc_Disc*d,int64_t n){return dc_Disc_ReadBlock(d,n);}
inline std::string dc_Disc_BootFilePath(dc_Disc*d){return go_strings_TrimSpace(d->IP.BootFile);}
template<class...A>void sh4_Inst_set(sh4_Inst*in,std::string n,std::string f,A...a){in->Mnem=n;in->Text=n+(f.empty()?"":" "+go_fmt_Sprintf(f,a...));}
inline void rrDCWrite(dc_Machine*m,const Slice<uint8_t>&b,uint32_t at){if(b.p==m->VRAM.p&&at<uint32_t(b.n))rrconsole::byte(at,b[at],m->Instrs,m->CPU->curPC);}
