#pragma once
#include "generated.cpp"
#include "../../../../browser/core/profile.h"
inline uint32_t hashBytes(const uint8_t*p,size_t n){uint32_t h=2166136261;for(size_t i=0;i<n;i++)h=(h^p[i])*16777619;return h;}
inline std::vector<uint8_t> frame(dsmachine_Machine*m){auto[top,bottom]=dsmachine_gpu2d_screens(m->gpu2d);std::vector<uint8_t>o(256*384*4);for(int j=0;j<2;j++){auto s=j?bottom:top;for(int i=0;i<256*192;i++){uint32_t v=(i<s.n?s[i]:0)|255;for(int c=0;c<4;c++)o[(j*256*192+i)*4+c]=v>>(24-8*c);}}return o;}
inline std::string proof(dsmachine_Machine*m){auto p=frame(m);std::ostringstream s;s<<"{\"steps\":"<<m->Steps<<",\"frames\":"<<m->vid.frames<<",\"arm9\":"<<hashBytes((uint8_t*)m->ARM9->cpu->R.data(),64)<<",\"arm7\":"<<hashBytes((uint8_t*)m->ARM7->cpu->R.data(),64)<<",\"ram\":"<<hashBytes(m->ram.p,m->ram.n)<<",\"rgba\":"<<hashBytes(p.data(),p.size())<<"}";return s.str();}
inline void bindFrameBoundary(dsmachine_Machine*m){m->OnFrame={};}
