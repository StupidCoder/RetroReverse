#pragma once
#include "generated.cpp"
inline uint32_t rrHash(const uint8_t*p,size_t n){uint32_t h=2166136261;for(size_t i=0;i<n;i++)h=(h^p[i])*16777619;return h;}
inline dos_Machine*realMachine=nullptr;inline dos_PM*protectedMachine=nullptr;
inline x86_CPU*rrCPU(){return realMachine?realMachine->CPU:protectedMachine?protectedMachine->CPU:nullptr;}
inline uint64_t rrFramePeriod(){return protectedMachine?545454:142857;}
inline uint64_t rrFrames(){return rrCPU()?rrCPU()->Steps/rrFramePeriod():0;}
inline void rrBoot(std::string exe,bool compatibility){auto data=rrfiles::all(exe);auto[coff,ce]=dos_ParseGo32COFF(data);if(!ce){auto[p,e]=dos_LoadGo32Bytes(data,go_filepath_Dir(exe));if(e)throw std::runtime_error(e.text);protectedMachine=p;p->rrQuakeBase=compatibility&&data.n==409600&&rrHash(data.p,data.n)==0xb155c323;p->CPU->OnStep=[p](x86_CPU*c){dos_PM_PumpInput(p,c);};}else{auto[m,e]=dos_LoadEXE(exe,go_filepath_Dir(exe));if(e)throw std::runtime_error(e.text);realMachine=m;m->EnableIRQ=true;}rrDOSReal=realMachine;rrDOSProtected=protectedMachine;}
inline std::vector<uint8_t>rrVideo(){std::vector<uint8_t>b(0xe0340);if(realMachine){auto*m=realMachine;for(int i=0;i<4;i++)memcpy(b.data()+0xa0000+i*65536,m->vga->planes[i].data(),65536);memcpy(b.data()+0xe0000,m->io->Pal.data(),768);memcpy(b.data()+0xe0300,m->vga->crtc.data(),32);memcpy(b.data()+0xe0320,m->vga->seq.data(),8);}else{memcpy(b.data()+0xa0000,protectedMachine->Mem.p+0xa0000,65536);memcpy(b.data()+0xe0000,protectedMachine->Pal.data(),768);}return b;}
inline uint32_t rrAddress(const std::vector<uint8_t>&b,int x,int y){if(x<0||y<0||x>=320||y>=200)return UINT32_MAX;if(protectedMachine)return 0xa0000+y*320+x;auto start=int(b[0xe030c])*256+b[0xe030d];auto pitch=int(b[0xe0313])*2;if(!pitch)pitch=80;auto off=b[0xe0324]&8?(y*80+x/4):start+y*pitch+x/4;return 0xa0000+(x&3)*65536+(off&65535);}
inline std::vector<uint8_t>rrFrame(const std::vector<uint8_t>&b){std::vector<uint8_t>out(320*200*4);for(int y=0;y<200;y++)for(int x=0;x<320;x++){auto pi=b[rrAddress(b,x,y)];int i=(y*320+x)*4;for(int j=0;j<3;j++)out[i+j]=b[0xe0000+pi*3+j]<<2;out[i+3]=255;}return out;}
inline auto rrFrame(){return rrFrame(rrVideo());}
#ifdef __EMSCRIPTEN__
#include <emscripten.h>
EM_JS(int,rrReadGameFile,(int id, uint8_t* out, double at, int n),{try{const b=new Uint8Array(new FileReaderSync().readAsArrayBuffer(Module.gameFiles[id].slice(at,at+n)));HEAPU8.set(b,out);return b.length;}catch(e){Module.discError=String(e);return -1;}});
#endif
