import {lightingWGSL} from './lighting-3ds.js';
// Binary32 arithmetic expressed in integers preserves WASM's rounding at each
// operation. WGSL permits floating-point reassociation, fusion and approximate
// division; using ordinary shader floats would change coverage and texel choice.
// Division uses a normal-range float estimate, then corrects it with an exact
// 64-bit integer remainder before rounding. No estimate reaches guest memory.
export const rasterFloatWGSL=`
fn rjam(x:u32,n:u32)->u32{if(n==0u){return x;}if(n>=32u){return select(0u,1u,x!=0u);}return (x>>n)|select(0u,1u,(x<<(32u-n))!=0u);}
fn rnorm(x:u32)->vec2<i32>{
 let e=(x>>23u)&255u;let m=x&0x7fffffu;
 if(e!=0u){return vec2<i32>(i32(m|0x800000u),i32(e)-127);}
 let shift=countLeadingZeros(m)-8u;return vec2<i32>(i32(m<<shift),-126-i32(shift));
}
fn rround(sign:u32,exponent:i32,significand:u32)->u32{
 var e=exponent;var s=significand;if(s==0u){return sign;}
 if(s>=0x8000000u){s=rjam(s,1u);e++;}
 if(s<0x4000000u){let n=countLeadingZeros(s)-5u;s<<=n;e-=i32(n);}
 if(e< -126){s=rjam(s,u32(-126-e));e= -126;}
 var m=(s>>3u)+select(0u,1u,(s&7u)>4u||((s&7u)==4u&&(s&8u)!=0u));
 if(m>=0x1000000u){m>>=1u;e++;}if(e>127){return sign|0x7f800000u;}
 return sign|select(u32(e+127)<<23u,0u,m<0x800000u)|(m&0x7fffffu);
}
fn radd(x:u32,y:u32)->u32{
 var a=x;var b=y;if((a&0x7fffffffu)<(b&0x7fffffffu)){a=y;b=x;}
 if((a&0x7f800000u)==0x7f800000u){if((a&0x7fffffu)!=0u||((b&0x7fffffffu)==0x7f800000u&&((a^b)>>31u)!=0u)){return 0x7fc00000u;}return a;}
 // Exact identities avoid unpacking/rounding common zero-valued attributes.
 if((b&0x7fffffffu)==0u){if((a&0x7fffffffu)==0u){return (a&b)&0x80000000u;}return a;}
 let ea=max((a>>23u)&255u,1u);let eb=max((b>>23u)&255u,1u);
 let ma=((a&0x7fffffu)|select(0u,0x800000u,(a&0x7f800000u)!=0u))<<3u;
 let mb=rjam(((b&0x7fffffu)|select(0u,0x800000u,(b&0x7f800000u)!=0u))<<3u,ea-eb);
 if(((a^b)&0x80000000u)==0u){return rround(a&0x80000000u,i32(ea)-127,ma+mb);}
 if(ma==mb){return 0u;}return rround(a&0x80000000u,i32(ea)-127,ma-mb);
}
fn rsub(a:u32,b:u32)->u32{return radd(a,b^0x80000000u);}
fn rmulWide(a:u32,b:u32)->vec2<u32>{
 let a0=a&65535u;let b0=b&65535u;let a1=a>>16u;let b1=b>>16u;
 let w0=a0*b0;let t=a1*b0+(w0>>16u);let w1=(t&65535u)+a0*b1;
 return vec2<u32>((w1<<16u)|(w0&65535u),a1*b1+(t>>16u)+(w1>>16u));
}
fn rmul(a:u32,b:u32)->u32{
 let sign=(a^b)&0x80000000u;let aa=a&0x7fffffffu;let bb=b&0x7fffffffu;
 if(aa>=0x7f800000u||bb>=0x7f800000u){if(aa>0x7f800000u||bb>0x7f800000u||aa==0u||bb==0u){return 0x7fc00000u;}return sign|0x7f800000u;}
 if(aa==0u||bb==0u){return sign;}
 if(bb==0x3f800000u){return sign|aa;}if(aa==0x3f800000u){return sign|bb;}
 let na=rnorm(a);let nb=rnorm(b);
 // Exponent-only multiplication is exact, with rround retaining subnormal ties.
 if(nb.x==0x800000){return rround(sign,na.y+nb.y,u32(na.x)<<3u);}
 if(na.x==0x800000){return rround(sign,na.y+nb.y,u32(nb.x)<<3u);}
 let product=rmulWide(u32(na.x),u32(nb.x));
 let high=(product.y&0x8000u)!=0u;let shift=select(20u,21u,high);
 let s=(product.y<<(32u-shift))|(product.x>>shift)|select(0u,1u,(product.x<< (32u-shift))!=0u);
 return rround(sign,na.y+nb.y+select(0,1,high),s);
}
fn rdiv(a:u32,b:u32)->u32{
 let sign=(a^b)&0x80000000u;let aa=a&0x7fffffffu;let bb=b&0x7fffffffu;
 if(aa>=0x7f800000u||bb>=0x7f800000u){if(aa>0x7f800000u||bb>0x7f800000u||(aa==bb)){return 0x7fc00000u;}return select(sign,sign|0x7f800000u,aa==0x7f800000u);}
 if(bb==0u){return select(sign|0x7f800000u,0x7fc00000u,aa==0u);}if(aa==0u){return sign;}
 if(bb==0x3f800000u){return sign|aa;}
 let na=rnorm(a);let nb=rnorm(b);var ma=u32(na.x);let mb=u32(nb.x);var e=na.y-nb.y;
 if(mb==0x800000u){return rround(sign,e,ma<<3u);}
 if(ma<mb){ma<<=1u;e--;}
 let numerator=vec2<u32>(ma<<23u,ma>>9u);
 var q=u32((f32(ma)*8388608.0)/f32(mb));var product=rmulWide(q,mb);
 // WGSL's bounded division error makes these correction loops short. Comparisons
 // and remainder arithmetic are integer-only, including exact halfway cases.
 while(product.y>numerator.y||(product.y==numerator.y&&product.x>numerator.x)){
  q--;let borrow=select(0u,1u,product.x<mb);product=vec2<u32>(product.x-mb,product.y-borrow);
 }
 var remainder=numerator.x-product.x;
 while(remainder>=mb){q++;remainder-=mb;}
 let tail=(remainder<<3u)/mb;let sticky=select(0u,1u,((remainder<<3u)%mb)!=0u);
 return rround(sign,e,(q<<3u)|tail|sticky);
}
// Correct a normal-range sqrt estimate with exact 64-bit squares, then compare
// the exact midpoint. Every binary32 radicand is handled, including subnormals.
fn rwideLess(a:vec2<u32>,b:vec2<u32>)->bool{return a.y<b.y||(a.y==b.y&&a.x<b.x);}
fn rsqrt(a:u32)->u32{
 if((a&0x7fffffffu)==0u){return a;}if((a>>31u)!=0u||a>0x7f800000u){return 0x7fc00000u;}if(a==0x7f800000u){return a;}
 let n=rnorm(a);var e=n.y;var m=u32(n.x);if((e&1)!=0){m<<=1u;e--;}
 let value=vec2<u32>(m<<23u,m>>9u);var q=u32(sqrt(f32(m)*8388608.0));
 while(rwideLess(value,rmulWide(q,q))){q--;}
 while(!rwideLess(value,rmulWide(q+1u,q+1u))){q++;}
 let four=vec2<u32>(value.x<<2u,(value.y<<2u)|(value.x>>30u));let mid=rmulWide(q*2u+1u,q*2u+1u);
 if(rwideLess(mid,four)||(all(mid==four)&&(q&1u)!=0u)){q++;}
 if(q==0x1000000u){q>>=1u;e+=2;}
 return (u32(e/2+127)<<23u)|(q&0x7fffffu);
}
fn rnegative(x:u32)->bool{return (x&0x80000000u)!=0u&&(x&0x7fffffffu)!=0u;}
fn rbyte(x:u32)->i32{
 if((x&0x80000000u)!=0u||x<0x3f800000u){return 0;}if(x>=0x437f0000u){return 255;}
 return i32(((x&0x7fffffu)|0x800000u)>>(150u-(x>>23u)));
}
fn rfloor(x:u32)->i32{
 let a=x&0x7fffffffu;if(a==0u){return 0;}let e=i32(a>>23u)-127;
 if(e<0){return select(0,-1,(x>>31u)!=0u);}
 let m=(a&0x7fffffu)|0x800000u;
 if(e>=23){return select(i32(m<<u32(e-23)),-i32(m<<u32(e-23)),(x>>31u)!=0u);}
 let shift=u32(23-e);let n=i32(m>>shift);return select(n,-n-select(0,1,(m<< (32u-shift))!=0u),(x>>31u)!=0u);
}
`;

export const rasterWGSL=`
${rasterFloatWGSL}
${lightingWGSL}
fn redge(a:u32,b:u32,x:u32,y:u32)->u32{
 return rsub(rmul(rsub(src[b],src[a]),rsub(y,src[a+1u])),rmul(rsub(src[b+1u],src[a+1u]),rsub(x,src[a])));
}
fn rpc(a:u32,field:u32,l:vec3<u32>,iw:u32)->u32{
 let stride=select(select(13u,14u,operation>=7u),21u,operation==8u);let b=a+stride;let c=b+stride;
 return rdiv(radd(radd(rmul(rmul(l.x,src[a+field]),src[a+2u]),rmul(rmul(l.y,src[b+field]),src[b+2u])),rmul(rmul(l.z,src[c+field]),src[c+2u])),iw);
}
fn rwrap(value:u32,n:u32,mode:u32)->i32{
 var v=rfloor(rmul(value,bitcast<u32>(f32(n))));let size=i32(n);
 if(mode==0u){return clamp(v,0,size-1);}if(mode==1u){return select(-1,v,v>=0&&v<size);}
 if(mode==2u){v%=size;if(v<0){v+=size;}return v;}
 let period=size*2;v%=period;if(v<0){v+=period;}return select(v,period-1-v,v>=size);
}
fn rdepthPass(value:u32,oldDepth:u32,fnCode:u32)->bool{
 switch(fnCode){case 0u:{return false;}case 1u:{return true;}case 2u:{return value==oldDepth;}case 3u:{return value!=oldDepth;}case 4u:{return value<oldDepth;}case 5u:{return value<=oldDepth;}case 6u:{return value>oldDepth;}default:{return value>=oldDepth;}}
}
fn rasterPixel(index:u32)->u32{
 let w=p[4];let h=p[5];let tile=index/64u;let mo=index%64u;
 let x=(tile%(w/8u))*8u+(mo&1u)+((mo>>1u)&2u)+((mo>>2u)&4u);
 let ty=(tile/(w/8u))*8u+((mo>>1u)&1u)+((mo>>2u)&2u)+((mo>>3u)&4u);let y=h-1u-ty;
 let px=bitcast<u32>(f32(x)+0.5);let py=bitcast<u32>(f32(y)+0.5);
 let bin=((y/16u)*p[40]+x/16u)*2u;let begin=src[bin];let end=begin+src[bin+1u];
 let original=old[index];var dstColor=vec4<i32>(i32(original>>24u),i32((original>>16u)&255u),i32((original>>8u)&255u),i32(original&255u));var count=0u;var killed=0u;var depthWord=0u;if(operation>=7u){depthWord=old[w*h+index];}
 for(var at=begin;at<end;at++){
  let t=src[at];if(x<src[t]||x>=src[t+1u]||y<src[t+2u]||y>=src[t+3u]){continue;}
  let a=t+5u;let stride=select(select(13u,14u,operation>=7u),21u,operation==8u);let b=a+stride;let c=b+stride;
  let weights=vec3<u32>(redge(b,c,px,py),redge(c,a,px,py),redge(a,b,px,py));
  if(rnegative(weights.x)||rnegative(weights.y)||rnegative(weights.z)){continue;}
  let area=src[t+4u];let l=vec3<u32>(rdiv(weights.x,area),rdiv(weights.y,area),rdiv(weights.z,area));
  let iw=radd(radd(rmul(l.x,src[a+2u]),rmul(l.y,src[b+2u])),rmul(l.z,src[c+2u]));
  if(iw==0u||iw>=0x7f800000u){atomicOr(&counts.drawn,0x80000000u);continue;}
  var newDepth=0u;
  if(operation>=7u){
   // Match each Reference binary32 operation, including W-buffer division and
   // float32 multiplication before truncating to the 24-bit depth integer.
   let z=radd(radd(rmul(l.x,src[a+13u]),rmul(l.y,src[b+13u])),rmul(l.z,src[c+13u]));
   var depth=radd(rmul(z,p[56]),p[57]);if((p[55]&4u)==0u){depth=rdiv(depth,iw);}
   if((depth&0x80000000u)!=0u){depth=0u;}else if(depth>0x3f800000u){depth=0x3f800000u;}
   newDepth=u32(rfloor(rmul(depth,0x4b7fffffu)));
   if(!rdepthPass(newDepth,depthWord&0xffffffu,(p[55]>>4u)&7u)){killed++;continue;}
  }
  var vertex:vec4<i32>;for(var channel=0u;channel<4u;channel++){vertex[channel]=rbyte(rmul(rpc(a,3u+channel,l,iw),0x437f0000u));}
  var tex:array<vec4<i32>,3>;
  for(var unit=0u;unit<3u;unit++){
   if((p[42]&(1u<<unit))==0u){continue;}let desc=43u+unit*4u;let tw=p[desc+1u];let th=p[desc+2u];
   if(tw==0u||th==0u){tex[unit]=vec4<i32>(255);continue;}
   let s=rpc(a,7u+unit*2u,l,iw);let v=rpc(a,8u+unit*2u,l,iw);
   // The reference converts these bounded coordinates to signed 32-bit values.
   if((s&0x7fffffffu)>0x46000000u||(v&0x7fffffffu)>0x46000000u){atomicOr(&counts.drawn,0x80000000u);continue;}
   let tx=rwrap(s,tw,(p[desc+3u]>>12u)&7u);let vy=rwrap(v,th,(p[desc+3u]>>8u)&7u);
   if(tx>=0&&vy>=0){tex[unit]=rgba(src[p[desc]+(th-1u-u32(vy))*tw+u32(tx)]);}
  }
  var color:vec4<i32>;
  if(operation==8u){
   var q:vec4<u32>;var view:vec3<u32>;var valid=true;
   for(var k=0u;k<4u;k++){q[k]=rpc(a,14u+k,l,iw);valid=valid&&((q[k]&0x7fffffffu)<=0x4b800000u);}
   for(var k=0u;k<3u;k++){view[k]=rpc(a,18u+k,l,iw);valid=valid&&((view[k]&0x7fffffffu)<=0x4b800000u);}
   if(!valid){atomicOr(&counts.drawn,0x80000000u);continue;}
   let light=lighting(q,view,tex);color=litTev(vertex,light[0],light[1],tex);
  }else{color=tev(vertex,tex);}if(alphaPass(color.a)){let value=blend(color,dstColor);for(var channel=0u;channel<4u;channel++){if((p[6]&(1u<<channel))!=0u){dstColor[channel]=value[channel];}}if(operation>=7u&&(p[55]&2u)!=0u){depthWord=(depthWord&0xff000000u)|newDepth;}count++;}
 }
 if(operation>=7u){dst[w*h+index]=depthWord;if(killed!=0u){atomicAdd(&counts.depthKilled,killed);}}
 if(count!=0u){atomicAdd(&counts.drawn,count);}
 return (u32(dstColor.r)<<24u)|(u32(dstColor.g)<<16u)|(u32(dstColor.b)<<8u)|u32(dstColor.a);
}`;
