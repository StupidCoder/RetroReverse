#pragma once
// Four independent colour lanes, with precisely the scalar multiply/add order.
// WASM SIMD and native NEON process these together; no relaxed SIMD or fast-math.
using rrFloat4=float __attribute__((ext_vector_type(4)));
inline rrFloat4 rrTexel4(psp_Machine*m,psp_geState*s,uint32_t x,uint32_t y,uint32_t level){auto[r,g,b,a]=psp_Machine_sampleTexLvl(m,s,x,y,level);return {float(r),float(g),float(b),float(a)};}
inline std::tuple<uint8_t,uint8_t,uint8_t,uint8_t>psp_modTex(psp_Machine*m,psp_geState*s,float u,float v,float rho,uint8_t r,uint8_t g,uint8_t b,uint8_t a){
 if(!s->texW||!s->texH||!s->texAddr)return {r,g,b,a};
 auto level=psp_geState_texLevel(s,rho);uint32_t tw=s->texW>>level,th=s->texH>>level;
 if(level&&s->texWN[level]){tw=s->texWN[level];th=s->texHN[level];}if(!tw||!th){level=0;tw=s->texW;th=s->texH;}
 uint8_t tr,tg,tb,ta;
 if(s->texLinear){
  float fu=u*float(tw)-0.5f,fv=v*float(th)-0.5f;
  float floorU=std::floor(fu),floorV=std::floor(fv),fx=fu-floorU,fy=fv-floorV;
  auto iu=cast<int64_t>(floorU),iv=cast<int64_t>(floorV);
  auto u0=psp_wrapTexel(iu,tw,s->texWrapU),v0=psp_wrapTexel(iv,th,s->texWrapV),u1=psp_wrapTexel(iu+1,tw,s->texWrapU),v1=psp_wrapTexel(iv+1,th,s->texWrapV);
  auto c00=rrTexel4(m,s,u0,v0,level),c10=rrTexel4(m,s,u1,v0,level),c01=rrTexel4(m,s,u0,v1,level),c11=rrTexel4(m,s,u1,v1,level);
  rrFloat4 top=c00*(1-fx)+c10*fx,bot=c01*(1-fx)+c11*fx,filtered=top*(1-fy)+bot*fy;
  tr=cast<uint8_t>(filtered[0]);tg=cast<uint8_t>(filtered[1]);tb=cast<uint8_t>(filtered[2]);ta=cast<uint8_t>(filtered[3]);
 }else{
  auto x=psp_wrapTexel(cast<int64_t>(u*float(tw)),tw,s->texWrapU),y=psp_wrapTexel(cast<int64_t>(v*float(th)),th,s->texWrapV);
  std::tie(tr,tg,tb,ta)=psp_Machine_sampleTexLvl(m,s,x,y,0);
 }
 uint8_t outA=a;
 switch(s->texFunc){
  case 1:if(s->texUseA){r=(unsigned(r)*(255-ta)+unsigned(tr)*ta)/255;g=(unsigned(g)*(255-ta)+unsigned(tg)*ta)/255;b=(unsigned(b)*(255-ta)+unsigned(tb)*ta)/255;}else{r=tr;g=tg;b=tb;}break;
  case 2:{auto mix=[](uint8_t c,uint8_t e,uint8_t t)->uint8_t{return (unsigned(c)*(255-t)+unsigned(e)*t)/255;};r=mix(r,s->texEnvCol,tr);g=mix(g,s->texEnvCol>>8,tg);b=mix(b,s->texEnvCol>>16,tb);if(s->texUseA)outA=unsigned(ta)*a/255;break;}
  case 3:r=tr;g=tg;b=tb;if(s->texUseA)outA=ta;break;
  case 4:r=std::min(255u,unsigned(r)+tr);g=std::min(255u,unsigned(g)+tg);b=std::min(255u,unsigned(b)+tb);if(s->texUseA)outA=unsigned(ta)*a/255;break;
  default:r=unsigned(tr)*r/255;g=unsigned(tg)*g/255;b=unsigned(tb)*b/255;if(s->texUseA)outA=unsigned(ta)*a/255;
 }
 if(s->texDouble){r=std::min(255,2*int(r));g=std::min(255,2*int(g));b=std::min(255,2*int(b));}
 return {r,g,b,outA};
}
