#pragma once
template<class...A>void x86_CPU_Halt(x86_CPU*c,std::string f,A...a){c->Halted=true;c->HaltReason=go_fmt_Sprintf(f,a...);}
template<class...A>void dos_Machine_logf(dos_Machine*m,std::string f,A...a){if(m->Log.n<1024)m->Log=append(m->Log,go_fmt_Sprintf(f,a...));}
template<class...A>void dos_PM_logf(dos_PM*m,std::string f,A...a){if(m->Log.n<1024)m->Log=append(m->Log,go_fmt_Sprintf(f,a...));}
inline std::tuple<uint16_t,uint16_t>dos_dosDateTime(os_File*){return {0x21,0};}
uint8_t rrRead(dos_Machine*,uint32_t);uint8_t rrRead(dos_PM*,uint32_t);
void rrWrite(dos_Machine*,uint32_t,uint8_t);void rrWrite(dos_PM*,uint32_t,uint8_t);

template<class...A>void dos_Machine_logInject(dos_Machine*m,std::string f,A...a){m->keyHits++;if(m->keyHits<100)dos_Machine_logf(m,f,a...);}
inline dos_Machine*rrDOSReal=nullptr;inline dos_PM*rrDOSProtected=nullptr;
inline void rrDOSWrite(uint8_t*p){uint32_t a=UINT32_MAX;x86_CPU*c=nullptr;
 auto map=[&](uint8_t*b,size_t n,uint32_t base){auto at=reinterpret_cast<uintptr_t>(p),start=reinterpret_cast<uintptr_t>(b);if(at>=start&&at-start<n)a=base+at-start;};
 if(rrDOSReal){auto*m=rrDOSReal;c=m->CPU;map(m->vga->planes[0].data(),262144,0xa0000);map(m->io->Pal.data(),768,0xe0000);map(m->vga->crtc.data(),32,0xe0300);map(m->vga->seq.data(),8,0xe0320);}
 if(rrDOSProtected){auto*m=rrDOSProtected;c=m->CPU;map(m->Mem.p+0xa0000,65536,0xa0000);map(m->Pal.data(),768,0xe0000);}
 if(a!=UINT32_MAX&&c){auto&t=rrcapture::trace;auto pc=c->SegBase[1]+c->instrIP;if(!t.current||t.events[t.current-1].pc!=pc)t.event(c->Steps,pc,"{\"kind\":\"x86 VGA write\"}");t.record(a,*p,1,c->Steps,pc,t.current);}
}
inline uint8_t*rrSpan(dos_PM*m,uint32_t a,uint32_t n,bool w){if(a>uint64_t(m->Mem.n)||n>uint64_t(m->Mem.n)-a||(w?bool(m->onW):bool(m->onR))||(w&&rrcapture::trace.active&&a<0xb0000&&uint64_t(a)+n>0xa0000))return nullptr;return m->Mem.p+a;}
inline uint8_t*rrSpan(dos_Machine*m,uint32_t a,uint32_t n,bool w){if(a>1048576||n>1048576-a||(a<0xb0000&&uint64_t(a)+n>0xa0000)||(w?(m->WatchLen||m->VGAProfileAt):m->RdProfileAt))return nullptr;return m->Mem.p+a;}
