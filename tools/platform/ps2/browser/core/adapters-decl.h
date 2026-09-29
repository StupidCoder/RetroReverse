#pragma once
r5900_CPU*r5900_NewCPU(ps2_Machine*);
mips_CPU*mips_NewCPU(ps2_IOP*);
std::tuple<Slice<uint8_t>,Error>localSource_ReadBlock(LocalSource*,int64_t);
std::tuple<iso9660_Geometry,bool>iso9660_Volume_Geometry(iso9660_Volume*);
std::tuple<ps2_Executable*,Error>ps2_LoadELF(Slice<uint8_t>);
std::tuple<ps2_IRX*,Error>ps2_ReadIRX(Slice<uint8_t>);
template<class...A>std::string ps2_sprintf(std::string f,A...a){return go_fmt_Sprintf(f,a...);}
template<class...A>void ps2_Machine_note(ps2_Machine*m,std::string f,A...a){auto s=go_fmt_Sprintf(f,a...);if(m->Log.n<4096)m->Log=append(m->Log,s);}
template<class...A>void ps2_Machine_Halt(ps2_Machine*m,std::string f,A...a){m->Halted=true;m->HaltReason=go_fmt_Sprintf(f,a...);}
template<class...A>void r5900_CPU_Halt(r5900_CPU*c,std::string f,A...a){c->Halted=true;c->HaltReason=go_fmt_Sprintf(f,a...);}
template<class...A>void mips_CPU_Halt(mips_CPU*c,std::string f,A...a){c->Halted=true;c->HaltReason=go_fmt_Sprintf(f,a...);}
template<class...A>void r5900_Inst_set(r5900_Inst*in,std::string name,std::string format,A...a){in->Mnem=name;in->Text=name+(format.empty()?"":" "+go_fmt_Sprintf(format,a...));}
template<class...A>void mips_Inst_set(mips_Inst*in,std::string name,std::string format,A...a){in->Mnem=name;in->Text=name+(format.empty()?"":" "+go_fmt_Sprintf(format,a...));}

template<class...A>void ps2_IOP_halt(ps2_IOP*p,std::string f,A...a){p->Halted=true;p->HaltReason=go_fmt_Sprintf(f,a...);}
template<class...A>void ps2_sprintfLog(std::string,A...){ }
void ps2_GS_rasterFill(ps2_GS*,ps2_gsTarget*,ps2_gsSampler*,int32_t,int32_t,int32_t,int32_t,std::function<void(int32_t,int32_t,ps2_gsStats*)>);

inline void ps2_GS_mergeStats(ps2_GS*g,ps2_gsStats*s){g->plotted+=s->plotted;g->rejScissor+=s->rejScissor;g->rejZ+=s->rejZ;g->rejAlpha+=s->rejAlpha;g->rejDate+=s->rejDate;g->texBlack+=s->texBlack;g->texColor+=s->texColor;for(int i=0;i<8;i++)g->plotNonBlack[i]+=s->plotNonBlack[i];for(int i=0;i<64;i++){g->texBlackPSM[i]+=s->texBlackPSM[i];g->texColorPSM[i]+=s->texColorPSM[i];}}
inline void ps2_GS_rasterFill(ps2_GS*g,ps2_gsTarget*,ps2_gsSampler*,int32_t,int32_t,int32_t y0,int32_t y1,std::function<void(int32_t,int32_t,ps2_gsStats*)>body){g->serFills++;ps2_gsStats s{};body(y0,y1,&s);ps2_GS_mergeStats(g,&s);}
inline void rrGSWrite(ps2_GS*g,uint32_t a){if(a<uint32_t(g->vram.n))rrconsole::byte(a,g->vram[a],g->m?g->m->steps:0,g->m?g->m->CPU->PC:0);}
inline rrconsole::EventScope rrGSTransfer(ps2_GS*g,const char*kind){auto&t=rrcapture::trace;auto old=t.current;if(t.active){std::ostringstream s;s<<"{\"kind\":\"GS transfer\",\"operation\":\""<<kind<<"\",\"bitbltbuf\":\"0x"<<std::hex<<g->reg[0x50]<<"\",\"trxpos\":\"0x"<<g->reg[0x51]<<"\",\"trxreg\":\"0x"<<g->reg[0x52]<<"\"}";rrconsole::command(g->m?g->m->steps:0,g->m?g->m->CPU->PC:0,s.str());}return rrconsole::EventScope(old);}
