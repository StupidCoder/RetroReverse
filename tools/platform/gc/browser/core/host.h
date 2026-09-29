#pragma once
#include "generated.cpp"
#include "../../../../browser/core/local-disc.h"
inline LocalDisc disc;
inline std::tuple<Slice<uint8_t>,Error>gc_Disc_Read(gc_Disc*d,int64_t off,int64_t n){
  try {if(off<0||n<0||n>64*1024*1024)throw std::runtime_error("Invalid disc read size");auto b=Slice<uint8_t>::make(n);disc.read(off,b.p,n);return {b,{}};}catch(const std::exception&e){return {{},{e.what()}};}
}
inline gc_Machine*rrBoot(uint64_t bytes){
  disc.mount(bytes);auto*d=arenaNew(gc_Disc{});d->Size=bytes;
  auto read=[&](uint64_t off,int n){auto[b,e]=gc_Disc_Read(d,off,n);if(e)throw std::runtime_error(e.text);return b;};
  auto h=read(0,0x440);if(be_Uint32(sub(h,0x1c,0x20))!=0xc2339f3d)throw std::runtime_error("Select an uncompressed GameCube ISO or GCM image");
  auto u=[&](int a){return be_Uint32(sub(h,a,a+4));};
  d->Header.GameID=cast<std::string>(sub(h,0,6));d->Header.Title=gc_cstr(sub(h,0x20,0x400));
  d->Header.DiscID=h[6];d->Header.Version=h[7];d->Header.Streaming=h[8]!=0;
  d->Header.DOLOffset=u(0x420);d->Header.FSTOffset=u(0x424);d->Header.FSTSize=u(0x428);d->Header.FSTMaxSize=u(0x42c);d->Header.FSTAddr=u(0x430);d->Header.UserOffset=u(0x434);d->Header.UserSize=u(0x438);
  auto a=read(0x2440,32);d->Apploader={gc_cstr(sub(a,0,16)),be_Uint32(sub(a,16,20)),be_Uint32(sub(a,20,24)),be_Uint32(sub(a,24,28))};
  auto[f,fe]=gc_ParseFST(read(d->Header.FSTOffset,d->Header.FSTSize));if(fe)throw std::runtime_error(fe.text);d->FST=f;
  auto[m,e]=gc_NewMachine(d);if(e)throw std::runtime_error(e.text);m->SingleThreaded=true;
  auto[entry,ae]=gc_Machine_RunApploader(m);if(ae)throw std::runtime_error(ae.text);
  return m;
}
inline uint32_t rrHash(const uint8_t*p,size_t n){uint32_t h=2166136261;for(size_t i=0;i<n;i++)h=(h^p[i])*16777619;return h;}
inline std::vector<uint8_t>rrFrame(gc_Machine*m){
  std::vector<uint8_t>out(640*480*4);uint32_t addr=gc_vi_XFBAddr(&m->vi);
  for(int y=0;y<480;y++)for(int x=0;x<640;x+=2){uint64_t a=uint64_t(addr)+(y*640+x)*2;if(a+3>=uint64_t(m->RAM.n))continue;
    auto[r0,g0,b0]=gc_yuv2rgb(m->RAM[a],m->RAM[a+1],m->RAM[a+3]);auto[r1,g1,b1]=gc_yuv2rgb(m->RAM[a+2],m->RAM[a+1],m->RAM[a+3]);
    auto*p=out.data()+(y*640+x)*4;p[0]=r0;p[1]=g0;p[2]=b0;p[3]=255;p[4]=r1;p[5]=g1;p[6]=b1;p[7]=255;}
  return out;
}
