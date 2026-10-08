#pragma once
// First lighting packet: one directional light, normal/tangent maps, distribution,
// reflection and Fresnel LUTs. Unsupported combinations stay in Reference before
// any cache or framebuffer mutation. The normalized direction is draw-constant.
namespace rrgpu {
inline bool lightingSupported(bool paired,n3ds_lightState*ls){
 if(!ls->enabled)return true;
 if(!paired||ls->count!=1||ls->shadow||ls->env>7)return false;
 const auto&l=ls->lights[0];if(!l.directional||l.distAtten||l.spotAtten||l.geo0||l.geo1)return false;
 for(auto a:{ls->ambient,l.specular0,l.specular1,l.diffuse,l.ambient,l.pos,l.spotDir})for(float f:a)if(!std::isfinite(f)||std::abs(f)>65536)return false;
 return true;
}
inline bool lightingParams(n3ds_GPU*g,n3ds_lightState*ls,std::vector<uint32_t>&p,std::vector<uint32_t>&data){
 const auto&l=ls->lights[0];uint32_t tables=0;
 for(auto [t,on]:{std::pair{0,!ls->noD0},std::pair{1,!ls->noD1},std::pair{3,!ls->noFR},std::pair{4,!ls->noRB},std::pair{5,!ls->noRG},std::pair{6,!ls->noRR}})if(on&&n3ds_lutSupported(ls->env,t))tables|=1u<<t;
 // Snapshot only the seven shared tables (including reserved table 2). Never
 // evaluate dormant tables: saved states can contain arbitrary unused values.
 const auto base=uint32_t(data.size());
 for(int t=0;t<7;t++)for(int i=0;i<256;i++)for(float f:{g->LUT[t][i],g->LUTDiff[t][i]}){
  if((tables&(1u<<t))&&(!std::isfinite(f)||std::abs(f)>1))return false;
  data.push_back((tables&(1u<<t))?floatWord(f):0);
 }
 const uint32_t config=uint32_t(ls->primaryAlpha)<<2|uint32_t(ls->secondAlpha)<<3|uint32_t(ls->clampHighlight)<<27|ls->bumpMode<<28|uint32_t(std::clamp<int64_t>(ls->bumpSel,0,2))<<22|uint32_t(ls->noBumpRenorm)<<30;
 p.insert(p.end(),{config,tables,ls->lutIn,ls->lutAbs,ls->lutScale,base,uint32_t(l.twoSided)});
 for(auto a:{ls->ambient,l.specular0,l.specular1,l.diffuse,l.ambient,n3ds_normalize3(l.pos),l.spotDir})for(float f:a)p.push_back(floatWord(f));
 return data.size()*4<=16u*1024*1024;
}
}
