#pragma once
// Resolve host memory once per primitive. Texels stay live: feedback rendering
// and writes through mirrored addresses cannot make a decoded cache stale.
struct rrTextureView {uint8_t*p=nullptr;uint32_t bytes=0,stride=0;};
inline psp_geState*rrTextureState=nullptr;
inline std::array<rrTextureView,8>rrTextureViews{};
struct rrTextureScope {
 psp_geState*previous;std::array<rrTextureView,8>old;
 rrTextureScope(psp_Machine*m,psp_geState*s):previous(rrTextureState),old(rrTextureViews){rrTextureState=s;
 for(unsigned i=0;i<8;i++){uint32_t a=s->texAddr,stride=s->texStride;if(i&&s->texAddrN[i]){a=s->texAddrN[i];stride=s->texStrideN[i];}if(!stride)stride=s->texW>>i;
 auto&v=rrTextureViews[i];v={};if(!a||!stride||m->OnRead)continue;auto p=a&0x1fffffff;uint32_t bytes=p>=0x08000000&&p<0x0a000000?0x0a000000-p:p>=0x04000000&&p<0x04200000?0x04200000-p:p>=65536&&p<81920?81920-p:0;if(bytes)v={rrMemory(m,a,1),bytes,stride};}}
 ~rrTextureScope(){rrTextureState=previous;rrTextureViews=old;}
};
__attribute__((always_inline)) inline std::tuple<uint8_t,uint8_t,uint8_t,uint8_t>psp_Machine_sampleTexLvl(psp_Machine*m,psp_geState*s,uint32_t x,uint32_t y,uint32_t level){
 if(s!=rrTextureState||level>=8||!rrTextureViews[level].p)return psp_Machine_sampleTexLvl_Reference(m,s,x,y,level);
 auto&v=rrTextureViews[level];uint32_t fmt=s->texFmt,xb,row,n;if(fmt==3){xb=x*4;row=v.stride*4;n=4;}else if(fmt<3){xb=x*2;row=v.stride*2;n=2;}else if(fmt==4){xb=x/2;row=v.stride/2;n=1;}else if(fmt==5){xb=x;row=v.stride;n=1;}else return psp_Machine_sampleTexLvl_Reference(m,s,x,y,level);
 uint32_t off=s->texSwizzle&&row>=16?((y/8)*(row/16)+xb/16)*128+(y%8)*16+xb%16:y*row+xb;
 if(off>v.bytes||n>v.bytes-off)return psp_Machine_sampleTexLvl_Reference(m,s,x,y,level);
 uint32_t c=0;if(n==4)std::memcpy(&c,v.p+off,4);else if(n==2){uint16_t h;std::memcpy(&h,v.p+off,2);c=h;}else c=v.p[off];if(fmt==3)return {uint8_t(c),uint8_t(c>>8),uint8_t(c>>16),uint8_t(c>>24)};
 if(fmt<3)return psp_decode16a(c,fmt);if(fmt==4)c=(c>>(4*(x&1)))&15;return psp_geState_clutLookup(s,c);
}
