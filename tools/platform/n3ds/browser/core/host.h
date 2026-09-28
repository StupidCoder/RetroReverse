#pragma once
#include "generated.cpp"
#include "../../../../browser/core/profile.h"
struct ImageScope{size_t start=arena.size();~ImageScope(){while(arena.size()>start){arena.back()();arena.pop_back();}}};
inline uint32_t hashBytes(const uint8_t*p,size_t n){uint32_t h=2166136261;for(size_t i=0;i<n;i++)h=(h^p[i])*16777619;return h;}
inline std::vector<uint8_t>frame(n3ds_Machine*m){ImageScope scope;std::vector<uint8_t>out(400*480*4);for(size_t i=3;i<out.size();i+=4)out[i]=255;for(int screen=0;screen<2;screen++){auto im=n3ds_Machine_Framebuffer(m,screen?"bottom":"top");if(!im)continue;auto w=im->Rect.Max.X,h=im->Rect.Max.Y;int left=screen?40:0;for(int64_t y=0;y<std::min<int64_t>(h,240);y++)for(int64_t x=0;x<std::min<int64_t>(w,screen?320:400);x++)for(int c=0;c<4;c++)out[((screen*240+y)*400+left+x)*4+c]=im->Pix[y*im->Stride+x*4+c];}return out;}
inline std::string proof(n3ds_Machine*m){auto p=frame(m);uint32_t mem=2166136261;for(auto r:m->regions)for(auto b:r->data)mem=(mem^b)*16777619;std::ostringstream s;s<<"{\"steps\":"<<m->instrs<<",\"frames\":"<<m->vblankCount<<",\"cpu\":"<<hashBytes((uint8_t*)m->CPU->R.data(),64)<<",\"cpsr\":"<<arm_CPU_CPSR(m->CPU)<<",\"ram\":"<<mem<<",\"rgba\":"<<hashBytes(p.data(),p.size())<<"}";return s.str();}
