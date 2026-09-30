#pragma once
template<class...A>void x86_CPU_Halt(x86_CPU*c,std::string f,A...a){c->Halted=true;c->HaltReason=go_fmt_Sprintf(f,a...);}
template<class...A>void dos_Machine_logf(dos_Machine*m,std::string f,A...a){if(m->Log.n<1024)m->Log=append(m->Log,go_fmt_Sprintf(f,a...));}
template<class...A>void dos_PM_logf(dos_PM*m,std::string f,A...a){if(m->Log.n<1024)m->Log=append(m->Log,go_fmt_Sprintf(f,a...));}
inline std::tuple<uint16_t,uint16_t>dos_dosDateTime(os_File*){return {0x21,0};}
uint8_t rrRead(dos_Machine*,uint32_t);uint8_t rrRead(dos_PM*,uint32_t);
void rrWrite(dos_Machine*,uint32_t,uint8_t);void rrWrite(dos_PM*,uint32_t,uint8_t);

template<class...A>void dos_Machine_logInject(dos_Machine*m,std::string f,A...a){m->keyHits++;if(m->keyHits<100)dos_Machine_logf(m,f,a...);}
inline dos_Machine*rrDOSReal=nullptr;inline dos_PM*rrDOSProtected=nullptr;
inline void rrDOSWrite(uint8_t*p){
 uint32_t a=rrdos::NONE; x86_CPU*c=nullptr; bool literal=true;
 auto map=[&](uint8_t*b,size_t n,uint32_t base){auto at=reinterpret_cast<uintptr_t>(p),start=reinterpret_cast<uintptr_t>(b);if(at>=start&&at-start<n){a=base+at-start;return true;}return false;};
 if(rrDOSReal){auto*m=rrDOSReal;c=m->CPU;
  if(map(m->Mem.p,m->Mem.n,0)){}
  // The existing VGA model treats mode 3 as mode 0; track its modeled literal stores too.
  else if(map(m->vga->planes[0].data(),262144,rrdos::VIDEO|0xa0000)){auto*v=m->vga;literal=(v->seq[4]&8)||(((v->gc[5]&3)==0||(v->gc[5]&3)==3)&&v->gc[8]==255&&v->gc[1]==0);}
  else if(map(m->io->Pal.data(),768,rrdos::VIDEO|0xe0000)||map(m->vga->crtc.data(),32,rrdos::VIDEO|0xe0300)||map(m->vga->seq.data(),8,rrdos::VIDEO|0xe0320))literal=false;
 }
 if(rrDOSProtected){auto*m=rrDOSProtected;c=m->CPU;if(map(m->Mem.p,m->Mem.n,0)){if(a>=0xa0000&&a<0xb0000){a|=rrdos::VIDEO;}}else if(map(m->Pal.data(),768,rrdos::VIDEO|0xe0000))literal=false;}
 if(a!=rrdos::NONE&&c&&rrmem::active&&!(a&rrdos::VIDEO))rrmem::access(c->Steps,0,a,*p,1,2,(c->Mode?c->SegBase[1]:uint32_t(c->Seg[1])<<4)+c->instrIP,a);
 if(a!=rrdos::NONE&&c&&rrmem::active&&(a&rrdos::VIDEO)){auto off=a&~rrdos::VIDEO;if(rrDOSProtected){if(off>=0xa0000&&off<0xb0000)rrmem::access(c->Steps,0,off,*p,1,2,c->SegBase[1]+c->instrIP,off);else if(off>=0xe0000&&off<0xe0300)rrmem::access(c->Steps,1,off-0xe0000,*p,1,2,c->SegBase[1]+c->instrIP,off);}else if(off>=0xa0000&&off<0xe0000)rrmem::access(c->Steps,1+(off-0xa0000)/65536,(off-0xa0000)%65536,*p,1,2,(c->Mode?c->SegBase[1]:uint32_t(c->Seg[1])<<4)+c->instrIP,off);else if(off>=0xe0000&&off<0xe0300)rrmem::access(c->Steps,5,off-0xe0000,*p,1,2,(c->Mode?c->SegBase[1]:uint32_t(c->Seg[1])<<4)+c->instrIP,off);}
 if(a!=rrdos::NONE&&c&&rrcapture::trace.active)rrdos::record(a,*p,c->Steps,(c->Mode?c->SegBase[1]:uint32_t(c->Seg[1])<<4)+c->instrIP,literal);
}
void rrDOSMovs(x86_CPU*,uint32_t,uint32_t,int64_t,uint32_t,uint32_t);
inline uint8_t*rrSpan(dos_PM*m,uint32_t a,uint32_t n,bool w){if(a>uint64_t(m->Mem.n)||n>uint64_t(m->Mem.n)-a||(w?bool(m->onW):bool(m->onR))||(w&&(rrcapture::trace.active||rrmem::active)))return nullptr;return m->Mem.p+a;}
inline uint8_t*rrSpan(dos_Machine*m,uint32_t a,uint32_t n,bool w){if(a>1048576||n>1048576-a||(a<0xb0000&&uint64_t(a)+n>0xa0000)||(w?(rrcapture::trace.active||rrmem::active||m->WatchLen||m->VGAProfileAt):m->RdProfileAt))return nullptr;return m->Mem.p+a;}
