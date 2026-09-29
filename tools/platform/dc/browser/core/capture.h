#pragma once
#include "../../../../browser/core/replay.h"
namespace rrdc {
inline std::array<uint32_t,2048>regs{};
inline std::pair<int,int> dimensions(){if(!(regs[0x44/4]&1))return {640,480};auto s=regs[0x5c/4],d=regs[0x44/4]>>2&3;return {int((s&1023)+1)*4/(d<2?2:d==2?3:4),int((s>>10&1023)+1)};}
inline uint32_t linearAddress(int x,int y){if(!(regs[0x44/4]&1))return UINT32_MAX;auto[w,h]=dimensions();if(x<0||y<0||x>=w||y>=h)return UINT32_MAX;auto size=regs[0x5c/4];int bpp=(regs[0x44/4]>>2&3)<2?2:(regs[0x44/4]>>2&3)==2?3:4;int64_t a=(regs[0x50/4]&0xffffff)+int64_t(y)*(int64_t((size&1023)+1)*4+(int64_t(size>>20&1023)-1)*4)+x*bpp;if(a<0||a+bpp>8388608)return UINT32_MAX;return a;}
inline std::vector<uint8_t> display(const std::vector<uint8_t>&b){auto[w,h]=dimensions();std::vector<uint8_t>out(w*h*4);auto d=regs[0x44/4]>>2&3;for(int y=0;y<h;y++)for(int x=0;x<w;x++){auto a=linearAddress(x,y);auto*p=out.data()+(y*w+x)*4;if(a==UINT32_MAX)continue;uint32_t c=0;for(int i=0;i<(d<2?2:3);i++)c|=rrcapture::trace.value(b,dc_vram32to64(a+i),1)<<(i*8);if(d<2){p[0]=(c>>(d?11:10)&31)<<3;p[1]=(c>>5&(d?63:31))<<(d?2:3);p[2]=(c&31)<<3;}else{p[0]=c>>16;p[1]=c>>8;p[2]=c;}p[3]=255;}return out;}
inline void command(dc_Machine*m,const char*kind,Slice<uint8_t>param={}){
 auto&t=rrcapture::trace;std::vector<uint8_t>snapshot(2048*4);std::memcpy(snapshot.data(),m->PVRRegs.data(),2048*4);
 std::ostringstream s;s<<"{\"kind\":\""<<kind<<"\",\"target\":"<<(m->PVRRegs[0x60/4]&0xffffff)<<",\"registerSnapshot\":\"PowerVR register file, little-endian 32-bit\",\"parameterBytes\":"<<param.n;if(param.n>=4)s<<",\"parameterControl\":"<<le_Uint32(param);s<<",\"parameterWords\":[";for(int i=0;i+4<=param.n;i+=4){if(i)s<<',';s<<le_Uint32(rrBorrow(param,i,i+4));}s<<"]}";rrconsole::command(m->Instrs,m->CPU->PC,s.str(),t.resource(snapshot.data(),snapshot.size()));
}
}
extern "C" {
int rr_capture_begin(){try{auto m=machine;rrconsole::pending=UINT32_MAX;rrcapture::trace.begin(m->VRAM.p,m->VRAM.n);m->OnPVRClear=[](int64_t,int64_t){rrdc::command(machine,"PowerVR framebuffer clear");};m->OnPVRCmd=[](Slice<uint8_t>p){rrdc::command(machine,"PowerVR TA parameter",p);};return 1;}catch(const std::exception&e){errorText=e.what();return 0;}}
int rr_capture_end(){try{rrconsole::flush();machine->OnPVRClear={};machine->OnPVRCmd={};rrdc::regs=machine->PVRRegs;rrcapture::trace.end(machine->VRAM.p,machine->VRAM.n);return 1;}catch(const std::exception&e){errorText=e.what();return 0;}}
const char*rr_capture_info(){reply=rrcapture::trace.info();return reply.c_str();}
const char*rr_pixel(int x,int y){if(!(rrdc::regs[0x44/4]&1)){reply="{\"blank\":true,\"contributors\":[]}";return reply.c_str();}auto a=rrdc::linearAddress(x,y);if(a==UINT32_MAX){reply="{\"error\":\"Outside scanout memory\"}";return reply.c_str();}uint32_t addresses[3];for(int i=0;i<3;i++)addresses[i]=dc_vram32to64(a+i);reply=rrcapture::trace.pixel(addresses[0],(rrdc::regs[0x44/4]>>2&3)<2?2:3,UINT32_MAX,0,addresses);reply.pop_back();reply+=",\"displayFormat\":"+std::to_string(rrdc::regs[0x44/4]>>2&3)+"}";return reply.c_str();}
const char*rr_source(uint32_t a,int n,uint32_t before,uint32_t expected){reply=rrcapture::trace.pixel(a,n,before,expected);return reply.c_str();}
const char*rr_resource(uint32_t id,uint32_t off){reply=rrcapture::trace.resourceJSON(id,off);return reply.c_str();}
const char*rr_replay_begin(){rrreplay::replay.begin();reply=rrreplay::replay.info();return reply.c_str();}
int rr_replay_seek(uint32_t n){return rrreplay::replay.seek(n);}
const char*rr_replay_info(){reply=rrreplay::replay.info();return reply.c_str();}
uint32_t rr_replay_for_write(uint32_t id){return rrreplay::replay.forWrite(id);}
uint8_t*rr_replay_frame(){pixels=rrdc::display(rrreplay::replay.memory);return pixels.data();}
}
