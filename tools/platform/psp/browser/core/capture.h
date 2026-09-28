#pragma once
#include "../../../../browser/core/replay.h"
namespace rrpsp {
// Compact capture address map: VRAM at 0, main RAM at 2 MiB, scratch at 34 MiB.
constexpr uint32_t ram=2097152,scratch=ram+33554432,total=scratch+16384;
inline uint32_t fb=0,stride=512,format=3;
inline uint32_t address(uint32_t a){a&=0x1fffffff;if(a>=0x04000000&&a<0x04200000)return a-0x04000000;if(a>=0x08000000&&a<0x0a000000)return ram+a-0x08000000;if(a>=65536&&a<81920)return scratch+a-65536;return UINT32_MAX;}
inline std::vector<uint8_t>memory(psp_Machine*m){std::vector<uint8_t>b(total);std::memcpy(b.data(),m->vram.p,ram);std::memcpy(b.data()+ram,m->ram.p,33554432);std::memcpy(b.data()+scratch,m->scratch.p,16384);return b;}
inline void clean(){auto&t=rrcapture::trace;t.source=t.palette=t.texel=t.sourceValue=t.paletteValue=0;t.u=t.v=0;}
inline void write(psp_Machine*m,uint32_t a,uint32_t v,unsigned n){auto at=address(a);auto&t=rrcapture::trace;if(!t.rendering&&t.value(t.shadow,at,n)==v)return;t.record(at,v,n,rrSteps,m->CPU->PC,t.rendering?t.current:0);}
inline uint32_t pixelAddress(int x,int y){if(x<0||x>=480||y<0||y>=272)return UINT32_MAX;return address(fb+(y*stride+x)*(format==3?4:2));}
inline std::vector<uint8_t>display(const std::vector<uint8_t>&b){std::vector<uint8_t>out(480*272*4);auto&t=rrcapture::trace;for(int y=0;y<272;y++)for(int x=0;x<480;x++){auto p=t.value(b,pixelAddress(x,y),format==3?4:2);uint8_t r=p,g=p>>8,blue=p>>16;if(format!=3){auto[r0,g0,b0,a0]=psp_decode16a(p,format);r=r0;g=g0;blue=b0;}auto off=(y*480+x)*4;out[off]=r;out[off+1]=g;out[off+2]=blue;out[off+3]=255;}return out;}
}
extern "C"{
int rr_capture_begin(){try{auto b=rrpsp::memory(machine);auto&t=rrcapture::trace;t.begin(b.data(),b.size());machine->WatchLo=0;machine->WatchHi=UINT32_MAX;
 rrObserveWrite=rrpsp::write;machine->OnWrite=[](uint32_t a,uint32_t v,uint32_t){rrpsp::write(machine,a,v,1);};
 machine->OnGeCmd=[](uint32_t word){rrpsp::clean();auto m=machine;auto s=m->geSt;std::ostringstream o;o<<"{\"kind\":\"GE command\",\"opcode\":"<<(word>>24)<<",\"value\":"<<(word&0xffffff);if(s)o<<",\"framebuffer\":"<<psp_geState_fbAddress(s)<<",\"stride\":"<<s->fbStride<<",\"format\":"<<s->fbFmt<<",\"texture\":"<<s->texAddr<<",\"textureFormat\":"<<s->texFmt<<",\"blend\":"<<(s->blendOn?"true":"false")<<",\"depthTest\":"<<(s->zTestOn?"true":"false");o<<"}";rrcapture::trace.event(rrSteps,m->CPU->PC,o.str());};
 machine->OnPixel=[](uint32_t x,uint32_t y,psp_PixelEvent e){if(e.Drawn||!machine->geSt)return;auto s=machine->geSt;if(x>=480||y>=272)return;auto n=s->fbFmt==3?4:2;auto a=rrpsp::address(psp_geState_fbAddress(s)+(y*s->fbStride+x)*n);auto v=uint32_t(e.R)|uint32_t(e.G)<<8|uint32_t(e.B)<<16|uint32_t(e.A)<<24;if(n==2){v=s->fbFmt==0?(e.R>>3)|uint32_t(e.G>>2)<<5|uint32_t(e.B>>3)<<11:s->fbFmt==1?(e.R>>3)|uint32_t(e.G>>3)<<5|uint32_t(e.B>>3)<<10|uint32_t(e.A>=128)<<15:(e.R>>4)|uint32_t(e.G>>4)<<4|uint32_t(e.B>>4)<<8|uint32_t(e.A>>4)<<12;}rrcapture::trace.record(a,v,n,rrSteps,machine->CPU->PC,rrcapture::trace.current,(e.ZReject?2:0)|(e.AlphaReject?4:0)|(e.StencilReject?16:0)|(e.ScissorReject?32:0)|(e.MaskReject?64:0));};return 1;}catch(const std::exception&e){errorText=e.what();return 0;}}
int rr_capture_end(){try{rrObserveWrite=nullptr;machine->OnWrite={};machine->OnGeCmd={};machine->OnPixel={};auto[a,s,f]=psp_Machine_Scanout(machine);rrpsp::fb=a;rrpsp::stride=s;rrpsp::format=f;auto b=rrpsp::memory(machine);rrcapture::trace.end(b.data(),b.size());return 1;}catch(const std::exception&e){errorText=e.what();return 0;}}
const char*rr_capture_info(){reply=rrcapture::trace.info();return reply.c_str();}
const char*rr_pixel(int x,int y){reply=rrcapture::trace.pixel(rrpsp::pixelAddress(x,y),rrpsp::format==3?4:2);reply.pop_back();reply+=",\"displayFormat\":"+std::to_string(rrpsp::format)+"}";return reply.c_str();}
const char*rr_source(uint32_t a,int size,uint32_t before,uint32_t expected){reply=rrcapture::trace.pixel(a,size,before,expected);return reply.c_str();}
const char*rr_resource(uint32_t id,uint32_t offset){reply=rrcapture::trace.resourceJSON(id,offset);return reply.c_str();}
const char*rr_replay_begin(){rrreplay::replay.begin();reply=rrreplay::replay.info();return reply.c_str();}
int rr_replay_seek(uint32_t n){return rrreplay::replay.seek(n);}
const char*rr_replay_info(){reply=rrreplay::replay.info();return reply.c_str();}
uint32_t rr_replay_for_write(uint32_t id){return rrreplay::replay.forWrite(id);}
uint8_t*rr_replay_frame(){pixels=rrpsp::display(rrreplay::replay.memory);return pixels.data();}
}
