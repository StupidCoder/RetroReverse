// Integer PICA fragment tail. Float32 rasterization and sampling stay in WASM.
// Parameters include every TEV/blend input; a single cached pipeline cannot
// accidentally reuse a shader compiled for another guest program or uniform.
export const fragmentWGSL=`
@group(0) @binding(4) var<storage,read_write> drawn:atomic<u32>;
fn rgba(v:u32)->vec4<i32>{return vec4<i32>(i32(v&255u),i32((v>>8u)&255u),i32((v>>16u)&255u),i32(v>>24u));}
fn fetchColor(sel:u32,vertex:vec4<i32>,tex:array<vec4<i32>,3>,buf:vec4<i32>,prev:vec4<i32>,konst:vec4<i32>)->vec4<i32>{
 switch(sel){case 0u,1u:{return vertex;}case 2u:{return vec4<i32>(0);}case 3u,4u,5u:{return tex[sel-3u];}case 13u:{return buf;}case 14u:{return konst;}default:{return prev;}}
}
fn colorOperand(s:vec4<i32>,op:u32)->vec3<i32>{
 switch(op){case 1u:{return vec3<i32>(255)-s.rgb;}case 2u:{return vec3<i32>(s.a);}case 3u:{return vec3<i32>(255-s.a);}case 4u:{return vec3<i32>(s.r);}case 5u:{return vec3<i32>(255-s.r);}case 8u:{return vec3<i32>(s.g);}case 9u:{return vec3<i32>(255-s.g);}case 12u:{return vec3<i32>(s.b);}case 13u:{return vec3<i32>(255-s.b);}default:{return s.rgb;}}
}
fn alphaOperand(s:vec4<i32>,op:u32)->i32{switch(op){case 0u:{return s.a;}case 1u:{return 255-s.a;}case 2u:{return s.r;}case 3u:{return 255-s.r;}case 4u:{return s.g;}case 5u:{return 255-s.g;}case 6u:{return s.b;}default:{return 255-s.b;}}}
fn combine(op:u32,a:i32,b:i32,c:i32)->i32{
 switch(op){case 0u:{return a;}case 1u:{return (a*b)/255;}case 2u:{return a+b;}case 3u:{return a+b-128;}case 4u:{return (a*c+b*(255-c))/255;}case 5u:{return a-b;}case 8u:{return (a*b)/255+c;}default:{return (clamp(a+b,0,255)*c)/255;}}
}
fn tev(vertex:vec4<i32>,tex:array<vec4<i32>,3>)->vec4<i32>{
 var prev=vertex;var buf=vec4<i32>(0);var next=rgba(p[11]);
 for(var stage=0u;stage<6u;stage++){
  let b=14u+stage*4u;let cs=p[b];let alphas=p[b+1u];let ops=p[b+2u];let konst=rgba(p[b+3u]);
  var ci:array<vec3<i32>,3>;var ai:array<i32,3>;
  for(var j=0u;j<3u;j++){let c=(cs>>(j*8u))&255u;let a=(alphas>>(j*8u))&255u;ci[j]=colorOperand(fetchColor(c&15u,vertex,tex,buf,prev,konst),c>>4u);ai[j]=alphaOperand(fetchColor(a&15u,vertex,tex,buf,prev,konst),a>>4u);}
  var value=vec4<i32>(combine(ops&255u,ci[0].r,ci[1].r,ci[2].r),combine(ops&255u,ci[0].g,ci[1].g,ci[2].g),combine(ops&255u,ci[0].b,ci[1].b,ci[2].b),combine((ops>>8u)&255u,ai[0],ai[1],ai[2]));
  value=clamp(value<<vec4<u32>(vec3<u32>((ops>>16u)&3u),(ops>>20u)&3u),vec4<i32>(0),vec4<i32>(255));
  buf=next;if((ops&0x1000000u)!=0u){next=vec4<i32>(value.rgb,next.a);}if((ops&0x2000000u)!=0u){next.a=value.a;}prev=value;
 }return prev;
}
fn alphaPass(a:i32)->bool{let cfg=p[12];if((cfg&1u)==0u){return true;}let r=i32((cfg>>8u)&255u);switch((cfg>>4u)&7u){case 0u:{return false;}case 1u:{return true;}case 2u:{return a==r;}case 3u:{return a!=r;}case 4u:{return a<r;}case 5u:{return a<=r;}case 6u:{return a>r;}default:{return a>=r;}}}
fn logic(op:u32,s:i32,d:i32)->i32{
 var r=0;switch(op){case 0u:{r=0;}case 1u:{r=s&d;}case 2u:{r=s&(~d);}case 3u:{r=s;}case 4u:{r=255;}case 5u:{r=~s;}case 6u:{r=d;}case 7u:{r=~d;}case 8u:{r=~(s&d);}case 9u:{r=s|d;}case 10u:{r=~(s|d);}case 11u:{r=s^d;}case 12u:{r=~(s^d);}case 13u:{r=d&(~s);}case 14u:{r=s|(~d);}default:{r=(~s)|d;}}return r&255;
}
fn factor(code:u32,s:i32,d:i32,sa:i32,da:i32,channel:u32)->i32{
 let cc=rgba(p[10]);let constant=cc[channel];
 switch(code){case 0u:{return 0;}case 1u:{return 255;}case 2u:{return s;}case 3u:{return 255-s;}case 4u:{return d;}case 5u:{return 255-d;}case 6u:{return sa;}case 7u:{return 255-sa;}case 8u:{return da;}case 9u:{return 255-da;}case 10u:{return constant;}case 11u:{return 255-constant;}case 12u:{return cc.a;}case 13u:{return 255-cc.a;}default:{return min(sa,255-da);}}
}
fn equation(op:u32,s:i32,d:i32,fs:i32,fd:i32)->i32{
 switch(op){case 0u:{return clamp((s*fs+d*fd)/255,0,255);}case 1u:{return clamp((s*fs-d*fd)/255,0,255);}case 2u:{return clamp((d*fd-s*fs)/255,0,255);}case 3u:{return min(s,d);}default:{return max(s,d);}}
}
fn blend(s:vec4<i32>,d:vec4<i32>)->vec4<i32>{
 var out:vec4<i32>;
 for(var c=0u;c<4u;c++){
  if((p[7]&256u)==0u){out[c]=logic(p[9]&15u,s[c],d[c]);continue;}
  let cfg=p[8];let alpha=c==3u;let eq=select(cfg&7u,(cfg>>8u)&7u,alpha);let sc=select((cfg>>16u)&15u,(cfg>>24u)&15u,alpha);let dc=select((cfg>>20u)&15u,(cfg>>28u)&15u,alpha);
  out[c]=equation(eq,s[c],d[c],factor(sc,s[c],d[c],s.a,d.a,c),factor(dc,s[c],d[c],s.a,d.a,c));
 }return out;
}
fn fragmentPixel(index:u32)->u32{
 let original=old[index];var dstColor=vec4<i32>(i32(original>>24u),i32((original>>16u)&255u),i32((original>>8u)&255u),i32(original&255u));var at=src[index];var count=0u;
 // The CPU creates acyclic lists; the cap also protects isolated malformed input.
 for(var i=0u;i<p[13]&&at!=0xffffffffu;i++){
  let vertex=rgba(src[at+1u]);let tex=array<vec4<i32>,3>(rgba(src[at+2u]),rgba(src[at+3u]),rgba(src[at+4u]));let color=tev(vertex,tex);
  if(alphaPass(color.a)){let value=blend(color,dstColor);for(var c=0u;c<4u;c++){if((p[6]&(1u<<c))!=0u){dstColor[c]=value[c];}}count++;}at=src[at];
 }
 if(count!=0u){atomicAdd(&drawn,count);}
 return (u32(dstColor.r)<<24u)|(u32(dstColor.g)<<16u)|(u32(dstColor.b)<<8u)|u32(dstColor.a);
}`;
