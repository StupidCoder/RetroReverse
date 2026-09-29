#pragma once
#include "../../../../browser/core/x86-fast.h"

#ifdef __wasm_simd128__
#include <wasm_simd128.h>
inline v128_t rrClamp4(v128_t v,float lo,float hi){auto l=wasm_f32x4_splat(lo),h=wasm_f32x4_splat(hi);v=wasm_v128_bitselect(h,v,wasm_f32x4_gt(v,h));return wasm_v128_bitselect(l,v,wasm_f32x4_lt(v,l));}
#endif
void xbox_combMap(std::array<float,4>*v,uint32_t op){
#if defined(__wasm_simd128__) && !defined(RR_GPU_REFERENCE)
 auto x=wasm_v128_load(v->data()),one=wasm_f32x4_splat(1),half=wasm_f32x4_splat(.5f),two=wasm_f32x4_splat(2);
 switch(op){case 0:x=rrClamp4(x,0,1);break;case 1:x=wasm_f32x4_sub(one,rrClamp4(x,0,1));break;case 2:case 3:x=wasm_f32x4_sub(wasm_f32x4_mul(two,rrClamp4(x,0,1)),one);if(op==3)x=wasm_f32x4_neg(x);break;case 4:case 5:x=wasm_f32x4_sub(rrClamp4(x,0,1),half);if(op==5)x=wasm_f32x4_neg(x);break;case 7:x=wasm_f32x4_neg(x);break;}
 wasm_v128_store(v->data(),x);
#else
 xbox_combMap_reference(v,op);
#endif
}
void xbox_combOp(std::array<float,4>*v,uint32_t op){
#if defined(__wasm_simd128__) && !defined(RR_GPU_REFERENCE)
 auto x=wasm_v128_load(v->data());switch(op){case 1:x=wasm_f32x4_sub(x,wasm_f32x4_splat(.5f));break;case 2:x=wasm_f32x4_mul(x,wasm_f32x4_splat(2));break;case 3:x=wasm_f32x4_mul(wasm_f32x4_sub(x,wasm_f32x4_splat(.5f)),wasm_f32x4_splat(2));break;case 4:x=wasm_f32x4_mul(x,wasm_f32x4_splat(4));break;case 6:x=wasm_f32x4_mul(x,wasm_f32x4_splat(.5f));break;}wasm_v128_store(v->data(),rrClamp4(x,-1,1));
#else
 xbox_combOp_reference(v,op);
#endif
}
