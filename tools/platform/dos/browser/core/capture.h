#pragma once
#include "../../../../browser/core/replay.h"
static uint64_t captureStartStep{};
extern "C" {
int rr_capture_begin(){try{captureStartStep=rrCPU()->Steps;auto b=rrVideo();rrcapture::trace.begin(b.data(),b.size());return 1;}catch(const std::exception&e){errorText=e.what();return 0;}}
int rr_capture_end(){try{auto b=rrVideo();rrcapture::trace.end(b.data(),b.size());return 1;}catch(const std::exception&e){errorText=e.what();return 0;}}
const char*rr_capture_info(){auto&t=rrcapture::trace;reply=t.info();reply.pop_back();bool idle=!t.writes.empty()&&rrCPU()->Steps-t.writes.back().clock>=20000;bool timeout=rrCPU()->Steps-captureStartStep>=rrFramePeriod()*140;reply+=",\"ready\":"+std::string(idle||timeout||t.overflow?"true":"false")+",\"vgaWriteWindow\":true}";return reply.c_str();}
const char*rr_pixel(int x,int y){auto&t=rrcapture::trace;auto a=rrAddress(t.final,x,y);reply=t.pixel(a,1);if(a<t.final.size()){auto palette=0xe0000+t.final[a]*3;reply.pop_back();reply+=",\"displayFormat\":\"VGA palette index\",\"paletteAddress\":"+std::to_string(palette)+",\"displayRGBA\": ["+std::to_string(t.final[palette]<<2)+","+std::to_string(t.final[palette+1]<<2)+","+std::to_string(t.final[palette+2]<<2)+",255]}";}return reply.c_str();}
const char*rr_source(uint32_t a,int n,uint32_t before,uint32_t expected){reply=rrcapture::trace.pixel(a,n,before,expected);return reply.c_str();}
const char*rr_resource(uint32_t id,uint32_t off){reply=rrcapture::trace.resourceJSON(id,off);return reply.c_str();}
const char*rr_replay_begin(){rrreplay::replay.begin();reply=rrreplay::replay.info();return reply.c_str();}
int rr_replay_seek(uint32_t n){return rrreplay::replay.seek(n);}
const char*rr_replay_info(){reply=rrreplay::replay.info();return reply.c_str();}
uint32_t rr_replay_for_write(uint32_t id){return rrreplay::replay.forWrite(id);}
uint8_t*rr_replay_frame(){pixels=rrFrame(rrreplay::replay.memory);return pixels.data();}
}
