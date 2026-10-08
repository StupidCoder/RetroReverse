// Exact single-directional-light subset. Float values are binary32 bit patterns;
// arithmetic preserves Reference rounding, LUT indexing and quaternion order.
// Packet offsets below include the transport's four-word prefix.
export const lightingWGSL=`
fn lclamp(x:u32)->u32{if(rnegative(x)){return 0u;}return select(x,0x3f800000u,(x&0x7fffffffu)>0x3f800000u);}
fn labs(x:u32)->u32{return select(x,x^0x80000000u,rnegative(x));}
fn ldot(a:vec3<u32>,b:vec3<u32>)->u32{return radd(radd(rmul(a.x,b.x),rmul(a.y,b.y)),rmul(a.z,b.z));}
fn lnormalize(v:vec3<u32>)->vec3<u32>{
 let square=ldot(v,v);if((square&0x7fffffffu)==0u||rnegative(square)){return vec3<u32>(0u,0u,0x3f800000u);}
 let len=rsqrt(square);return vec3<u32>(rdiv(v.x,len),rdiv(v.y,len),rdiv(v.z,len));
}
fn lrotate(q:vec4<u32>,v:vec3<u32>)->vec3<u32>{
 let len=rsqrt(radd(ldot(q.xyz,q.xyz),rmul(q.w,q.w)));if((len&0x7fffffffu)==0u){return v;}
 let x=rdiv(q.x,len);let y=rdiv(q.y,len);let z=rdiv(q.z,len);let w=rdiv(q.w,len);
 let tx=rmul(0x40000000u,rsub(rmul(y,v.z),rmul(z,v.y)));
 let ty=rmul(0x40000000u,rsub(rmul(z,v.x),rmul(x,v.z)));
 let tz=rmul(0x40000000u,rsub(rmul(x,v.y),rmul(y,v.x)));
 return vec3<u32>(radd(radd(v.x,rmul(w,tx)),rsub(rmul(y,tz),rmul(z,ty))),radd(radd(v.y,rmul(w,ty)),rsub(rmul(z,tx),rmul(x,tz))),radd(radd(v.z,rmul(w,tz)),rsub(rmul(x,ty),rmul(y,tx))));
}
fn lscale(code:u32)->u32{switch(code&7u){case 0u:{return 0x3f800000u;}case 1u:{return 0x40000000u;}case 2u:{return 0x40800000u;}case 3u:{return 0x41000000u;}case 6u:{return 0x3e800000u;}case 7u:{return 0x3f000000u;}default:{return 0u;}}}
fn llut(field:u32,table:u32,angles:array<u32,5>)->u32{
 let input=(p[62]>>field)&7u;var c=0u;if(input<5u){c=angles[input];}
 var idx=0;var delta=0u;
 if(((p[63]>>(field+1u))&1u)==0u){
  if(p[66]!=0u){c=labs(c);}else if(rnegative(c)){c=0u;}
  c=lclamp(c);let scaled=rmul(c,0x43800000u);idx=clamp(rfloor(scaled),0,255);delta=rsub(scaled,bitcast<u32>(f32(idx)));
 }else{let scaled=rmul(c,0x43000000u);idx=clamp(rfloor(scaled),-128,127);delta=rsub(scaled,bitcast<u32>(f32(idx)));}
 let at=p[65]+table*512u+(u32(idx)&255u)*2u;
 return rmul(lscale(p[64]>>field),lclamp(radd(src[at],rmul(src[at+1u],delta))));
}
fn lighting(q:vec4<u32>,view:vec3<u32>,tex:array<vec4<i32>,3>)->array<vec4<i32>,2>{
 let cfg=p[60];let tables=p[61];var normal=vec3<u32>(0u,0u,0x3f800000u);
 if(((cfg>>28u)&3u)==1u){
  let t=tex[(cfg>>22u)&3u];for(var i=0u;i<3u;i++){normal[i]=rsub(rdiv(bitcast<u32>(f32(t[i])),0x42ff0000u),0x3f800000u);}
  if((cfg&0x40000000u)==0u){var z=rsub(0x3f800000u,radd(rmul(normal.x,normal.x),rmul(normal.y,normal.y)));if(rnegative(z)){z=0u;}normal.z=rsqrt(z);}
 }
 normal=lrotate(q,normal);
 let nv=lnormalize(view);let dir=vec3<u32>(p[82],p[83],p[84]);let spot=vec3<u32>(p[85],p[86],p[87]);
 let halfVector=lnormalize(vec3<u32>(radd(nv.x,dir.x),radd(nv.y,dir.y),radd(nv.z,dir.z)));
 let ndlRaw=ldot(dir,normal);
 let angles=array<u32,5>(ldot(normal,halfVector),ldot(nv,halfVector),ldot(normal,nv),ndlRaw,ldot(dir,spot));
 var d0=0x3f800000u;var d1=d0;var fr=d0;var refl=vec3<u32>(d0);
 if((tables&1u)!=0u){d0=llut(0u,0u,angles);}if((tables&2u)!=0u){d1=llut(4u,1u,angles);}
 if((tables&64u)!=0u){refl.x=llut(24u,6u,angles);}refl.y=refl.x;refl.z=refl.x;
 if((tables&32u)!=0u){refl.y=llut(20u,5u,angles);}if((tables&16u)!=0u){refl.z=llut(16u,4u,angles);}
 if((tables&8u)!=0u){fr=llut(12u,3u,angles);}
 var ndl=ndlRaw;if(p[66]!=0u){ndl=labs(ndl);}else if(rnegative(ndl)){ndl=0u;}
 let highlight=select(0x3f800000u,0u,(cfg&0x08000000u)!=0u&&(ndl&0x7fffffffu)==0u);
 var primary:vec4<i32>;var secondary:vec4<i32>;
 for(var i=0u;i<3u;i++){
  // Reference starts each accumulator at +0; preserve its signed-zero rule.
  let pri=lclamp(radd(radd(0u,radd(rmul(p[76u+i],ndl),p[79u+i])),p[67u+i]));
  let spec0=rmul(p[70u+i],d0);let spec1=rmul(rmul(d1,refl[i]),p[73u+i]);
  let sec=lclamp(radd(0u,rmul(radd(spec0,spec1),highlight)));
  primary[i]=rbyte(rmul(pri,0x437f0000u));secondary[i]=rbyte(rmul(sec,0x437f0000u));
 }
 primary.a=rbyte(rmul(lclamp(select(0x3f800000u,fr,(cfg&4u)!=0u)),0x437f0000u));
 secondary.a=rbyte(rmul(lclamp(select(0x3f800000u,fr,(cfg&8u)!=0u)),0x437f0000u));
 return array<vec4<i32>,2>(primary,secondary);
}`;
