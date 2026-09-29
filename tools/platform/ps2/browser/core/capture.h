#pragma once
#include "../../../../browser/core/replay.h"
namespace rrps2 {
inline uint32_t fbp=0,fbw=0,dbx=0,dby=0;inline int w=640,h=448;
inline void scanout(){auto m=machine;uint32_t reg=(get(m->io,uint32_t(ps2_gsPMODE))&3)==1?ps2_gsDISPFB1:ps2_gsDISPFB2;uint64_t d=uint64_t(get(m->io,reg))|uint64_t(get(m->io,reg+4))<<32;fbp=(d&511)*32;fbw=(d>>9)&63;dbx=(d>>32)&2047;dby=(d>>43)&2047;w=width;h=height;}
inline uint32_t address(int x,int y){if(x<0||y<0||x>=w||y>=h||!fbw)return UINT32_MAX;return ps2_addrPSMCT32(fbp,fbw,dbx+x,dby+y);}
inline Slice<uint8_t>display(const std::vector<uint8_t>&b){auto out=Slice<uint8_t>::make(w*h*4);for(int y=0;y<h;y++)for(int x=0;x<w;x++){auto a=address(x,y);if(uint64_t(a)+4>b.size())continue;auto o=(y*w+x)*4;out[o]=b[a];out[o+1]=b[a+1];out[o+2]=b[a+2];out[o+3]=255;}return out;}
}
extern "C" {
int rr_capture_begin(){try{if(!machine)throw std::runtime_error("No machine");ps2_Machine_ensureGS(machine);auto g=machine->gs;rrconsole::pending=UINT32_MAX;rrcapture::trace.begin(g->vram.p,g->vram.n);
 machine->OnGSPrim=[](int64_t typ,std::string producer){auto g=machine->gs;auto p=ps2_GS_prim(g);auto t=ps2_GS_target(g,p);auto&trace=rrcapture::trace;std::vector<uint8_t>snapshot(128*8);for(size_t i=0;i<g->reg.size();i++)for(int j=0;j<8;j++)snapshot[i*8+j]=g->reg[i]>>(j*8);const auto resource=trace.resource(snapshot.data(),snapshot.size());std::ostringstream s;s<<"{\"kind\":\"GS "<<ps2_primNames[typ&7]<<"\",\"producer\":\""<<rrconsole::escape(producer)<<"\",\"framebuffer\":"<<t.fbp*256<<",\"stride\":"<<t.fbw*64<<",\"format\":"<<t.psm<<",\"depthFormat\":"<<t.zpsm<<",\"blend\":"<<(t.abe?"true":"false")<<",\"registerSnapshot\":\"128 GS registers, little-endian 64-bit\",\"vertices\":[";for(int i=0;i<g->vqN;i++){if(i)s<<',';auto&v=g->vq[i];s<<"{\"x\":"<<v.x<<",\"y\":"<<v.y<<",\"z\":"<<v.z<<",\"rgba\":"<<v.rgba<<"}";}s<<"]}";return int64_t(rrconsole::command(machine->steps,machine->CPU->PC,s.str(),resource));};return 1;}catch(const std::exception&e){errorText=e.what();return 0;}}
int rr_capture_end(){try{rrconsole::flush();machine->OnGSPrim={};rr_frame();rrps2::scanout();auto g=machine->gs;rrcapture::trace.end(g->vram.p,g->vram.n);return 1;}catch(const std::exception&e){errorText=e.what();return 0;}}
const char*rr_capture_info(){reply=rrcapture::trace.info();return reply.c_str();}
const char*rr_pixel(int x,int y){reply=rrcapture::trace.pixel(rrps2::address(x,y),4);return reply.c_str();}
const char*rr_source(uint32_t a,int n,uint32_t before,uint32_t expected){reply=rrcapture::trace.pixel(a,n,before,expected);return reply.c_str();}
const char*rr_resource(uint32_t id,uint32_t off){reply=rrcapture::trace.resourceJSON(id,off);return reply.c_str();}
const char*rr_replay_begin(){rrreplay::replay.begin();reply=rrreplay::replay.info();return reply.c_str();}
int rr_replay_seek(uint32_t n){return rrreplay::replay.seek(n);}
const char*rr_replay_info(){reply=rrreplay::replay.info();reply.pop_back();reply+=",\"surface\":\"Final display buffer in GS VRAM\"}";return reply.c_str();}
uint32_t rr_replay_for_write(uint32_t id){return rrreplay::replay.forWrite(id);}
uint8_t*rr_replay_frame(){pixels=rrps2::display(rrreplay::replay.memory);return pixels.p;}
}
