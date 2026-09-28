#pragma once
// One setup per GE draw, live buffer contents per fragment. The scalar Go port
// remains the oracle and handles observers / unusual buffer geometry.
struct rrFragmentContext {
 psp_Machine*m;psp_geState*s;uint8_t*fb=nullptr,*zb=nullptr;
 uint32_t fbBase=0,zBase=0,stride=0,zStride=0;
 bool depthTest=false,depthWrite=false;
 void(*emit)(rrFragmentContext&,int,int,float,uint8_t,uint8_t,uint8_t,uint8_t)=nullptr;
};
inline rrFragmentContext*rrFragments=nullptr;
template<unsigned F>inline uint32_t rrLoadColor(const uint8_t*p){
 uint32_t v=0;std::memcpy(&v,p,F==3?4:2);
 if constexpr(F==3)return v;
 else {auto[r,g,b,a]=psp_decode16a(v,F);return uint32_t(r)|(uint32_t(g)<<8)|(uint32_t(b)<<16)|(uint32_t(a)<<24);}
}
template<unsigned F>inline void rrStoreColor(rrFragmentContext&c,uint32_t off,uint32_t rgba){
 uint32_t v=rgba;
 if constexpr(F==0)v=((rgba&255)>>3)|(((rgba>>8&255)>>2)<<5)|(((rgba>>16&255)>>3)<<11);
 if constexpr(F==1)v=((rgba&255)>>3)|(((rgba>>8&255)>>3)<<5)|(((rgba>>16&255)>>3)<<10)|((rgba>>31)<<15);
 if constexpr(F==2)v=((rgba&255)>>4)|(((rgba>>8&255)>>4)<<4)|(((rgba>>16&255)>>4)<<8)|((rgba>>28)<<12);
 constexpr unsigned n=F==3?4:2;
 if(rrObserveWrite)rrObserveWrite(c.m,c.fbBase+off*n,v,n);
 std::memcpy(c.fb+off*n,&v,n);
}
template<unsigned F,bool Stencil,unsigned Blend>void rrEmit(rrFragmentContext&c,int x,int y,float z,uint8_t r,uint8_t g,uint8_t b,uint8_t a){
 auto m=c.m;auto s=c.s;
 auto event=[&](psp_PixelEvent e){psp_Machine_pixelEvent(m,x,y,e);};
 if(!s->clearOn&&s->scX1>0&&(x<s->scX0||x>s->scX1||y<s->scY0||y>s->scY1)){event({false,true,false,false,false,false,r,g,b,a});return;}
 if(!psp_geState_alphaPass(s,a)){event({false,false,true,false,false,false,r,g,b,a});return;}
 uint32_t off=uint32_t(y)*c.stride+x;uint32_t dst=0;uint8_t dstA=0;
 if constexpr(Stencil){
  dst=rrLoadColor<F>(c.fb+off*(F==3?4:2));dstA=dst>>24;
  if(!psp_geState_stencilTest(s,dstA)){
   rrStoreColor<F>(c,off,(dst&0xffffff)|(uint32_t(psp_stencilOp(s->stSFail,dstA,uint8_t(s->stRef)))<<24));
   event({false,false,false,true,false,false,r,g,b,a});return;
  }
 }
 uint32_t zo=(uint32_t(y)*c.zStride+x)*2;uint16_t zi=0;
 if(c.depthTest||c.depthWrite)zi=cast<uint16_t>(psp_clampF(z,0,65535));
 if(c.depthTest){uint16_t old;std::memcpy(&old,c.zb+zo,2);bool pass;
  switch(s->zFunc){case 0:pass=false;break;case 1:pass=true;break;case 2:pass=zi==old;break;case 3:pass=zi!=old;break;case 4:pass=zi<old;break;case 5:pass=zi<=old;break;case 6:pass=zi>old;break;default:pass=zi>=old;}
  if(!pass){if constexpr(Stencil)rrStoreColor<F>(c,off,(dst&0xffffff)|(uint32_t(psp_stencilOp(s->stZFail,dstA,uint8_t(s->stRef)))<<24));event({false,false,false,false,true,false,r,g,b,a});return;}
 }
 if(c.depthWrite){if(rrObserveWrite)rrObserveWrite(m,c.zBase+zo,zi,2);std::memcpy(c.zb+zo,&zi,2);}
 uint8_t outA=a;if constexpr(Stencil)outA=psp_stencilOp(s->stZPass,dstA,uint8_t(s->stRef));
 // Read after depth writes: framebuffer/depth-buffer aliasing is legal. Reuse
 // this value for blend and masks, between which no memory write occurs.
 if constexpr(Blend!=0){dst=rrLoadColor<F>(c.fb+off*(F==3?4:2));
  auto ch=[&](uint8_t sc,uint8_t dc,int i)->uint8_t{
   if constexpr(Blend==1)return std::min(255u,unsigned(sc)*a/255+unsigned(dc)*(255-a)/255);
   else {unsigned sf=psp_blendFactor(s->blendSrc,sc,a,dst>>24,s->blendFixA,i),df=psp_blendFactor(s->blendDst,sc,a,dst>>24,s->blendFixB,i);int sv=unsigned(sc)*sf/255,dv=unsigned(dc)*df/255;
    switch(s->blendEq){case 1:return psp_clamp255(sv-dv);case 2:return psp_clamp255(dv-sv);case 3:return std::min(sc,dc);case 4:return std::max(sc,dc);case 5:return std::abs(int(sc)-dc);default:return psp_clamp255(sv+dv);}
   }
  };r=ch(r,dst,0);g=ch(g,dst>>8,1);b=ch(b,dst>>16,2);
 }
 if(s->maskRGB||s->maskA){
  if(s->maskRGB==0xffffff&&s->maskA==255){event({false,false,false,false,false,true,r,g,b,a});return;}
  if constexpr(Blend==0)dst=rrLoadColor<F>(c.fb+off*(F==3?4:2));
  r=(r&~uint8_t(s->maskRGB))|(uint8_t(dst)&uint8_t(s->maskRGB));
  g=(g&~uint8_t(s->maskRGB>>8))|(uint8_t(dst>>8)&uint8_t(s->maskRGB>>8));
  b=(b&~uint8_t(s->maskRGB>>16))|(uint8_t(dst>>16)&uint8_t(s->maskRGB>>16));
  outA=(outA&~uint8_t(s->maskA))|(uint8_t(dst>>24)&uint8_t(s->maskA));
 }
 rrStoreColor<F>(c,off,uint32_t(r)|(uint32_t(g)<<8)|(uint32_t(b)<<16)|(uint32_t(outA)<<24));event({true,false,false,false,false,false,r,g,b,outA});
}
template<unsigned F,bool Stencil>inline void rrSelectBlend(rrFragmentContext&c){auto s=c.s;
 if(!s->blendOn||s->clearOn)c.emit=rrEmit<F,Stencil,0>;
 else if(s->blendEq==0&&s->blendSrc==2&&s->blendDst==3)c.emit=rrEmit<F,Stencil,1>;
 else c.emit=rrEmit<F,Stencil,2>;
}
template<unsigned F>inline void rrSelectFragment(rrFragmentContext&c){if(c.s->stencilOn&&!c.s->clearOn)rrSelectBlend<F,true>(c);else rrSelectBlend<F,false>(c);}
struct rrFragmentScope {
 rrFragmentContext context;rrFragmentContext*previous;
 rrFragmentScope(psp_Machine*m,psp_geState*s):context{m,s},previous(rrFragments){
  rrFragments=&context;auto&c=context;
  if(m->OnRead||(m->OnWrite&&!rrObserveWrite)||psp_geProbeX>=0)return;
  c.fbBase=psp_geState_fbAddress(s);c.stride=s->fbStride?s->fbStride:480;
  uint64_t bytes=(uint64_t(271)*c.stride+480)*(s->fbFmt==3?4:2);
  if(bytes>UINT32_MAX||!(c.fb=rrMemory(m,c.fbBase,bytes)))return;
  c.depthTest=s->zLow&&s->zStride&&!s->clearOn&&s->zTestOn&&!psp_geNoZ;
  c.depthWrite=s->zLow&&s->zStride&&(s->clearOn?s->clearDepth:s->zTestOn&&!s->zNoWrite);
  if(c.depthTest||c.depthWrite){c.zBase=psp_geState_zAddress(s);c.zStride=s->zStride;uint64_t zb=(uint64_t(271)*c.zStride+480)*2;if(zb>UINT32_MAX||!(c.zb=rrMemory(m,c.zBase,zb)))return;}
  switch(s->fbFmt){case 1:rrSelectFragment<1>(c);break;case 2:rrSelectFragment<2>(c);break;case 3:rrSelectFragment<3>(c);break;default:rrSelectFragment<0>(c);}
 }
 ~rrFragmentScope(){rrFragments=previous;}
};
inline void psp_Machine_putPixel(psp_Machine*m,psp_geState*s,int64_t x,int64_t y,float z,uint8_t r,uint8_t g,uint8_t b,uint8_t a){
 if(rrFragments&&rrFragments->m==m&&rrFragments->s==s&&rrFragments->emit&&x>=0&&x<480&&y>=0&&y<272)rrFragments->emit(*rrFragments,x,y,z,r,g,b,a);
 else psp_Machine_putPixel_Reference(m,s,x,y,z,r,g,b,a);
}
