#pragma once
#include "../../../../browser/core/replay.h"
extern "C" {
int rr_capture_begin(){try{auto b=rrgba::memory(machine,false);rrcapture::trace.begin(b.data(),b.size());return 1;}catch(const std::exception&e){errorText=e.what();return 0;}}
int rr_capture_end(){try{auto b=rrgba::memory(machine,true);rrcapture::trace.end(b.data(),b.size());return 1;}catch(const std::exception&e){errorText=e.what();return 0;}}
const char*rr_capture_info(){reply=rrcapture::trace.info();return reply.c_str();}
const char*rr_pixel(int x,int y){auto a=x>=0&&x<240&&y>=0&&y<160?rrgba::frameBase+(y*240+x)*4:UINT32_MAX;reply=rrcapture::trace.pixel(a,4);return reply.c_str();}
const char*rr_source(uint32_t a,int n,uint32_t before,uint32_t expected){reply=rrcapture::trace.pixel(a,n,before,expected);return reply.c_str();}
const char*rr_resource(uint32_t id,uint32_t off){reply=rrcapture::trace.resourceJSON(id,off);return reply.c_str();}
const char*rr_replay_begin(){rrreplay::replay.begin();reply=rrreplay::replay.info();return reply.c_str();}
int rr_replay_seek(uint32_t n){return rrreplay::replay.seek(n);}
const char*rr_replay_info(){reply=rrreplay::replay.info();return reply.c_str();}
uint32_t rr_replay_for_write(uint32_t id){return rrreplay::replay.forWrite(id);}
uint8_t*rr_replay_frame(){auto&b=rrreplay::replay.memory;pixels.resize(rrgba::frameBytes);for(unsigned i=0;i<240*160;i++){auto c=rrcapture::trace.value(b,rrgba::frameBase+i*4,4);pixels[i*4]=c>>16;pixels[i*4+1]=c>>8;pixels[i*4+2]=c;pixels[i*4+3]=255;}return pixels.data();}
}
