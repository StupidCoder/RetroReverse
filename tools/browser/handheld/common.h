#pragma once
#include "../core/capture.h"
#include "../core/replay.h"
#include "../core/profile.h"
namespace rrhh {
constexpr unsigned width=160,height=144,pixels=width*height,frameBase=0x20000,memorySize=frameBase+pixels*4;
struct Video {
 std::array<uint32_t,pixels>draw{},screen{};
 uint64_t frames=0,steps=0;uint32_t phase=0,offDots=0;
 int32_t windowLine=0,lastLine=-1;bool statIRQ=false;uint16_t pc=0;
};
inline Video video;
inline void clean(){auto&t=rrcapture::trace;t.source=t.palette=t.texel=t.sourceValue=t.paletteValue=0;t.sourceBefore=t.paletteBefore=t.writes.size();t.u=t.v=0;}
inline void memoryWrite(uint32_t a,uint32_t v,int size=1){auto&t=rrcapture::trace;if(!t.active)return;clean();if(t.value(t.shadow,a,size)!=v)t.record(a,v,size,video.steps,video.pc);}
inline uint32_t event(const char*kind,int y,bool source=true,int index=-1,uint32_t descriptor=0){auto&t=rrcapture::trace;if(!t.active)return 0;clean();std::ostringstream s;s<<"{\"kind\":\""<<kind<<"\",\"scanline\":"<<y<<",\"hasSource\":"<<(source?"true":"false")<<",\"object\":"<<index<<",\"descriptor\":"<<descriptor<<"}";return t.event(video.steps,video.pc,s.str());}
inline void dot(int x,int y,uint32_t color,uint32_t source=0,uint32_t sourceValue=0,uint32_t palette=0,uint32_t paletteValue=0,unsigned texel=0,int u=0,int v=0,uint8_t flags=1){
 if(x<0||x>=160||y<0||y>=144)return;auto&t=rrcapture::trace;
 if(t.active){t.source=source;t.sourceValue=sourceValue;t.palette=palette;t.paletteValue=paletteValue;t.texel=texel;t.u=u;t.v=v;t.sourceBefore=t.paletteBefore=t.writes.size();t.record(frameBase+(y*width+x)*4,color,4,video.steps,video.pc,t.current,flags);}
 if(flags&1)video.draw[y*width+x]=color;
}
inline void clear(uint32_t color){if(rrcapture::trace.active){event("LCD frame start",0,false);for(int y=0;y<144;y++)for(int x=0;x<160;x++)dot(x,y,color);}else video.draw.fill(color);}
inline void present(){video.screen=video.draw;video.frames++;}
inline void init(uint32_t color){video={};video.draw.fill(color);video.screen.fill(color);}
inline void includeFrame(std::vector<uint8_t>&b){std::memcpy(b.data()+frameBase,video.draw.data(),pixels*4);}
inline uint32_t hash(const uint8_t*p,size_t n){uint32_t h=2166136261;for(size_t i=0;i<n;i++)h=(h^p[i])*16777619;return h;}
}
