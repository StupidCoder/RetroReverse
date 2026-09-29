#pragma once

// Separate the small, ordinary BAT lookup from the formatted fault path. This
// lets the compiler inline translation into instruction fetches and data loads.
// All four entries and the privilege/locked-cache rules remain live on every
// access, so guest BAT/MSR writes need no host-cache invalidation.
inline std::tuple<uint32_t,bool> gekko_CPU_Translate(gekko_CPU*c,uint32_t ea,bool store,bool insn){
  if(!(c->MSR&(insn?gekko_MSRIR:gekko_MSRDR)))return {ea,true};
  if(gekko_LockedCache_Enabled(&c->LC)&&gekko_LockedCache_Contains(&c->LC,ea))return {ea,true};
  const auto&bats=insn?c->IBAT:c->DBAT;
  const uint32_t valid=(c->MSR&gekko_MSRPR)?1:2;
  for(unsigned i=0;i<4;i++){
    uint32_t upper=bats[i][0],lower=bats[i][1];
    if(!(upper&valid))continue;
    uint32_t offset=((upper>>2)&0x7ff)<<17|0x1ffff,block=~offset;
    if((ea&block)==(upper&0xfffe0000&block))
      return {(lower&0xfffe0000&block)|(ea&offset),true};
  }
  return gekko_CPU_Translate_reference(c,ea,store,insn);
}

// An instruction fetch needs one checked, big-endian word, not a temporary Go
// slice with four separately checked byte reads. Unusual addresses retain the
// reference path's diagnostics and bounds behavior.
inline uint32_t gc_Machine_Fetch32(gc_Machine*m,uint32_t a) {
  if(a<=24u*1024*1024-4&&uint64_t(a)+4<=uint64_t(m->RAM.n)){
    uint32_t word;std::memcpy(&word,m->RAM.p+a,4);
    if constexpr(std::endian::native==std::endian::little)word=__builtin_bswap32(word);
    return word;
  }
  return gc_Machine_Fetch32_reference(m,a);
}

inline gc_tevState gc_gpu_tevstate(gc_gpu*g){
  auto t=gc_gpu_tevstate_reference(g);
  constexpr uint8_t colorBase[]{0,3,4,7,8,11,12,15,16,19,20,23,28,29,24,30};
  constexpr uint8_t alphaBase[]{3,7,11,15,19,23,27,30};
  for(int s=0;s<t.numStages;s++){
    auto&st=t.stages[s];
    for(unsigned a=0;a<4;a++){
      unsigned code=(st.cc>>(a*4))&15;
      for(unsigned k=0;k<3;k++)
        st.rrOperands.color[a][k]=colorBase[code]+((code<12&&!(code&1))||code==14?k:0);
      st.rrOperands.alpha[a]=alphaBase[(st.ac>>(4+a*3))&7];
    }
  }
  t.rrPrepared=true;
  return t;
}

// Keep the TEV's arithmetic and stage ordering unchanged, but resolve operand
// selectors once per draw. The old interpreter switches separately for every
// operand of every stage of every fragment. Indexed operands also expose the
// three independent color lanes to LLVM's vectorizer.
inline std::tuple<uint8_t,uint8_t,uint8_t,uint8_t,bool> gc_gpu_shade(
    gc_gpu*g,gc_Machine*m,gc_tevState*t,
    std::array<std::array<uint8_t,4>,2>*rasCol,std::array<gc_texCoord,8>*tc){
  if(!t->rrPrepared)return gc_gpu_shade_reference(g,m,t,rasCol,tc);
  std::array<float,32> inputs{};
  for(unsigned r=0;r<4;r++)for(unsigned k=0;k<4;k++)inputs[r*4+k]=t->seed[r][k];
  inputs[28]=255;inputs[29]=128;
  for(int s=0;s<t->numStages;s++){
    const auto&st=t->stages[s];
    auto ras=gc_swizzle(st.swapRas,gc_rasSelect(st.rasSel,rasCol));
    std::array<float,4> tex{};
    if(st.texEnable){
      auto[r,gc,b,a]=gc_gpu_sampleTexmap(g,m,&t->tex[st.texmap],(*tc)[st.texcoord].s,(*tc)[st.texcoord].t);
      tex=gc_swizzle(st.swapTex,{float(r),float(gc),float(b),float(a)});
    }
    for(unsigned k=0;k<4;k++){inputs[16+k]=tex[k];inputs[20+k]=ras[k];}
    for(unsigned k=0;k<3;k++)inputs[24+k]=st.konstC[k];
    inputs[27]=st.konstA;
    std::array<float,3> a{},b{},c{},d{},out{};
    for(unsigned k=0;k<3;k++){
      d[k]=inputs[st.rrOperands.color[0][k]];c[k]=inputs[st.rrOperands.color[1][k]];
      b[k]=inputs[st.rrOperands.color[2][k]];a[k]=inputs[st.rrOperands.color[3][k]];
    }
    unsigned bias=(st.cc>>16)&3,scale=(st.cc>>20)&3;
    bool sub=(st.cc>>18)&1,clamp=(st.cc>>19)&1;
    if(bias==3){
      unsigned sel=scale*2+sub;
      bool packed=scale!=3&&gc_tevCompare(sel,a,b,0,0,-1);
      for(unsigned k=0;k<3;k++){
        bool pass=scale==3?gc_tevCompare(sel,a,b,0,0,k):packed;
        out[k]=pass?d[k]+c[k]:d[k];
        if(clamp)out[k]=gc_clampf(out[k]);
      }
    }else{
      for(unsigned k=0;k<3;k++)out[k]=gc_tevFormula(a[k],b[k],c[k],d[k],bias,sub,scale,clamp);
    }
    float ad=inputs[st.rrOperands.alpha[0]],ac=inputs[st.rrOperands.alpha[1]],
          ab=inputs[st.rrOperands.alpha[2]],aa=inputs[st.rrOperands.alpha[3]],alpha;
    bias=(st.ac>>16)&3;scale=(st.ac>>20)&3;sub=(st.ac>>18)&1;clamp=(st.ac>>19)&1;
    if(bias==3){
      alpha=ad;
      if(gc_tevCompare(scale*2+sub,a,b,aa,ab,-1))alpha+=ac;
      if(clamp)alpha=gc_clampf(alpha);
    }else alpha=gc_tevFormula(aa,ab,ac,ad,bias,sub,scale,clamp);
    for(unsigned k=0;k<3;k++)inputs[st.cdest*4+k]=out[k];
    inputs[st.adest*4+3]=alpha;
  }
  auto alpha=gc_toU8(inputs[3]);
  if(!gc_gpu_alphaTest(g,alpha))return {0,0,0,0,false};
  return {gc_toU8(inputs[0]),gc_toU8(inputs[1]),gc_toU8(inputs[2]),alpha,true};
}
