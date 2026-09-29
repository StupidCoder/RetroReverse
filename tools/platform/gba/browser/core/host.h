#pragma once
#include "generated.cpp"
inline uint32_t rrHash(const uint8_t*p,size_t n){uint32_t h=2166136261;for(size_t i=0;i<n;i++)h=(h^p[i])*16777619;return h;}
inline std::vector<uint8_t>rrFrame(gbamachine_Machine*m){std::vector<uint8_t>b(240*160*4);for(size_t i=0;i<240*160;i++){auto c=m->screen[i];b[i*4]=c>>16;b[i*4+1]=c>>8;b[i*4+2]=c;b[i*4+3]=255;}return b;}
