#pragma once
#include "../../../../browser/core/replay.h"
static RRSurface capturedSurface;
extern "C" {
int rr_capture_begin(){try{rrconsole::pending=UINT32_MAX;rrcapture::trace.begin(machine->RAM.p,machine->RAM.n);return 1;}catch(const std::exception&e){errorText=e.what();return 0;}}
int rr_capture_end(){try{rrconsole::flush();capturedSurface=rrSurface(machine);rrcapture::trace.end(machine->RAM.p,machine->RAM.n);return 1;}catch(const std::exception&e){errorText=e.what();return 0;}}
const char*rr_capture_info(){reply=rrcapture::trace.info();return reply.c_str();}
const char*rr_pixel(int x,int y){auto&s=capturedSurface;reply=rrcapture::trace.pixel(rrAddress(s,x,y),4);reply.pop_back();reply+=",\"displayFormat\":\"NV2A BGRA8888; first stored sample\",\"samplesPerPixel\":"+std::to_string(s.ax*s.ay)+",\"sampleAddresses\":[";bool comma=false;for(int sy=0;sy<s.ay;sy++)for(int sx=0;sx<s.ax;sx++){if(comma)reply+=',';comma=true;reply+=std::to_string(rrAddress(s,x,y,sx,sy));}reply+="]}";return reply.c_str();}
const char*rr_source(uint32_t a,int n,uint32_t before,uint32_t expected){reply=rrcapture::trace.pixel(a,n,before,expected);return reply.c_str();}
const char*rr_resource(uint32_t id,uint32_t off){reply=rrcapture::trace.resourceJSON(id,off);return reply.c_str();}
const char*rr_replay_begin(){rrreplay::replay.begin();reply=rrreplay::replay.info();return reply.c_str();}
int rr_replay_seek(uint32_t n){return rrreplay::replay.seek(n);}
const char*rr_replay_info(){reply=rrreplay::replay.info();return reply.c_str();}
uint32_t rr_replay_for_write(uint32_t id){return rrreplay::replay.forWrite(id);}
uint8_t*rr_replay_frame(){auto&b=rrreplay::replay.memory;pixels=rrFrame(b.data(),b.size(),capturedSurface);return pixels.data();}
}
