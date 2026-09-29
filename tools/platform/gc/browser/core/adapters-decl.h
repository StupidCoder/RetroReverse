#pragma once
template<class...A>void gekko_CPU_Halt(gekko_CPU*c,std::string f,A...a){if(!c->Halted){c->Halted=true;c->HaltReason=go_fmt_Sprintf(f,a...);}}
template<class...A>void gcdsp_CPU_Halt(gcdsp_CPU*c,std::string f,A...a){c->Halted=true;c->Reason=go_fmt_Sprintf(f,a...);}
template<class...A>void gc_Machine_note(gc_Machine*m,std::string f,A...a){auto s=go_fmt_Sprintf(f,a...);if(m->logSeen.size()<4096&&!get(m->logSeen,s)){m->logSeen[s]=true;m->Log=append(m->Log,s);}}
template<class...A>void gc_siLog(std::string,A...){ }
gekko_CPU*gekko_NewCPU(gc_Machine*);
uint32_t gekko_CPU_fetch(gekko_CPU*,uint32_t);
std::tuple<Slice<uint8_t>,Error>gc_Disc_Read(gc_Disc*,int64_t,int64_t);
inline std::tuple<std::string,Error>gc_Disc_MD5(gc_Disc*){return {"local-disc",{}};}
void gc_gpu_fill(gc_gpu*,gc_Machine*,gc_tevState*,Slice<gc_rasterTri>);

template<class...A>void gc_Machine_logf(gc_Machine*m,std::string f,A...a){gc_Machine_note(m,f,a...);}
template<class...A>void gekko_Inst_set(gekko_Inst*in,std::string name,std::string format,A...a){in->Mnem=name;in->Text=name+(format.empty()?"":" "+go_fmt_Sprintf(format,a...));}
inline gcdsp_CPU*gcdsp_CPU_Clone(gcdsp_CPU*c){return arenaNew(*c);}

inline void gc_gpu_dumpTex0Once(gc_gpu*,gc_Machine*,uint32_t){}
inline constexpr uint32_t rrEFBBase=24*1024*1024;
inline gc_Machine*rrCaptureMachine=nullptr;
inline void rrEFBWrite(gc_gpu*g,uint32_t i){auto&t=rrcapture::trace;rrconsole::flush();auto*m=rrCaptureMachine;if(m&&i<uint32_t(g->EFB.n))t.record(rrEFBBase+i*4,g->EFB[i],4,m->Instrs,m->CPU->PC,t.current);}
inline void rrGXWrite(gc_Machine*m,uint32_t a){if(a<uint32_t(m->RAM.n))rrconsole::byte(a,m->RAM[a],m->Instrs,m->CPU->PC);}
inline rrconsole::EventScope rrEFBClear(gc_gpu*){auto&t=rrcapture::trace;auto old=t.current;if(t.active&&rrCaptureMachine)rrconsole::command(rrCaptureMachine->Instrs,rrCaptureMachine->CPU->PC,"{\"kind\":\"EFB color and depth clear\"}");return rrconsole::EventScope(old);}
