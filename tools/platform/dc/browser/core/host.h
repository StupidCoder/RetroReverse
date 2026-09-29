#pragma once
#include "generated.cpp"
#include "../../../../browser/core/local-disc.h"
inline LocalDisc disc;
inline Slice<dc_Track>discTracks;
inline std::tuple<Slice<uint8_t>,Error>dc_Disc_ReadBlock(dc_Disc*d,int64_t lba){try{for(auto t:d->Tracks)if(dc_Track_IsData(t)&&lba>=t.StartLBA&&lba<t.StartLBA+t.Length/2352){auto b=Slice<uint8_t>::make(2048);disc.read(t.FileOffset+(lba-t.StartLBA)*2352+16,b.p,2048);return {b,{}};}throw std::runtime_error("LBA outside selected Dreamcast data tracks");}catch(const std::exception&e){return {{},{e.what()}};}}
inline dc_Machine*rrBoot(uint64_t bytes){disc.mount(bytes);if(discTracks.n==0)throw std::runtime_error("Select the Dreamcast CUE and its BIN companion");auto*d=arenaNew(dc_Disc{});d->Tracks=discTracks;d->size=bytes;for(auto&t:d->Tracks){if(t.FileOffset<0||t.Length<=0||uint64_t(t.FileOffset)>bytes||uint64_t(t.Length)>bytes-t.FileOffset)throw std::runtime_error("Invalid track bounds");if(dc_Track_IsData(t))d->data=t;}if(d->data.StartLBA<0||d->data.Mode.empty())throw std::runtime_error("No Dreamcast data track");auto[v,e]=iso9660_OpenVolumeAt(d,d->data.StartLBA+16);if(e)throw std::runtime_error(e.text);d->Vol=v;auto[b,be]=dc_Disc_ReadBlock(d,d->data.StartLBA);if(be)throw std::runtime_error(be.text);auto[ip,ie]=dc_parseIPBin(b);if(ie)throw std::runtime_error(ie.text);d->IP=ip;auto*m=dc_NewMachine(d);auto error=dc_Machine_Boot(m);if(error)throw std::runtime_error(error.text);return m;}
inline uint32_t rrHash(const uint8_t*p,size_t n){uint32_t h=2166136261;for(size_t i=0;i<n;i++)h=(h^p[i])*16777619;return h;}
inline std::pair<int,int>rrDimensions(dc_Machine*m){if(!(m->PVRRegs[0x44/4]&1))return {640,480};auto s=m->PVRRegs[0x5c/4],d=(m->PVRRegs[0x44/4]>>2)&3;int bpp=d<2?2:d==2?3:4;return {int((s&1023)+1)*4/bpp,int((s>>10&1023)+1)};}
inline std::vector<uint8_t>rrFrame(dc_Machine*m){
 auto[w,h]=rrDimensions(m);std::vector<uint8_t>out(w*h*4);if(!(m->PVRRegs[0x44/4]&1))return out;
 auto depth=(m->PVRRegs[0x44/4]>>2)&3;int bpp=depth<2?2:depth==2?3:4;auto size=m->PVRRegs[0x5c/4];int64_t row=m->PVRRegs[0x50/4]&0xffffff;
 for(int y=0;y<h;y++,row+=int64_t((size&1023)+1)*4+(int64_t(size>>20&1023)-1)*4)for(int x=0;x<w;x++){
  int64_t a=row+x*bpp;if(a<0||a+bpp>m->VRAM.n)throw std::runtime_error("Invalid Dreamcast scanout address");auto read=[&](int n){return m->VRAM[dc_vram32to64(a+n)];};
  auto*p=out.data()+(y*w+x)*4;uint16_t c=read(0)|uint16_t(read(1))<<8;
  if(depth<2){p[0]=(c>>(depth?11:10)&31)<<3;p[1]=(c>>5&(depth?63:31))<<(depth?2:3);p[2]=(c&31)<<3;}else{p[0]=read(2);p[1]=read(1);p[2]=read(0);}p[3]=255;
 }return out;
}
