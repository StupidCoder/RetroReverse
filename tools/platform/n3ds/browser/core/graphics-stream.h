#pragma once
#ifdef __EMSCRIPTEN__
#include <emscripten.h>
// Asyncify preserves the actual C++ continuation: no command is retried.
// Guest bytes remain authoritative until a complete isolated result is ready.
EM_ASYNC_JS(int,rr_gpu_submit,(uint32_t kind,const uint32_t* params,uint32_t np,const uint8_t* input,uint32_t ni,uint8_t* target,uint32_t n),{
 if(!Module.graphicsTransfer)return 0;
 try{
  const packet={kind,params:Array.from(HEAPU32.subarray(params>>>2,(params>>>2)+np)),input:HEAPU8.slice(input,input+ni),before:HEAPU8.slice(target,target+n)};
  const result=await Module.graphicsTransfer(packet);
  if(!result?.supported||!(result.bytes instanceof Uint8Array)||result.bytes.length!==n)return 0;
  HEAPU8.set(result.bytes,target);return 1;
 }catch(e){Module.graphicsError=String(e);return 0;}
});
#endif
// Browser-only immutable inputs. This is deliberately separate from Render's
// output-write provenance. Recording never supplies results to the machine.
namespace rrgpu {
constexpr uint32_t magic=0x50475252,version=1;
constexpr size_t limit=256u*1024*1024;
inline bool recording=false;
inline bool enabled=false;
inline std::array<uint32_t,5> operations{},accelerated{};
inline uint64_t committedBytes=0;
inline bool active(n3ds_Machine*m,uint32_t kind){
 if(enabled)operations[kind]++;
 return recording||(enabled&&!m->OnRead&&!m->OnWrite&&!m->OnPixel&&!m->HidTrace&&!m->Profile);
}
inline uint32_t dropped=0,draws=0;
inline std::vector<uint8_t> stream;
inline void word(std::vector<uint8_t>&v,uint32_t x){for(int j=0;j<4;j++)v.push_back(x>>(j*8));}
inline void bytes(std::vector<uint8_t>&v,const std::vector<uint8_t>&b){v.insert(v.end(),b.begin(),b.end());while(v.size()%4)v.push_back(0);}
inline uint8_t* range(n3ds_Machine*m,uint32_t a,uint64_t n){
 if(n>16u*1024*1024||uint64_t(a)+n>UINT32_MAX)return nullptr;
 auto r=n3ds_Machine_regionOf(m,a);if(!r||uint64_t(a-r->base)+n>uint64_t(r->data.n))return nullptr;
 return r->data.p+(a-r->base);
}
inline bool overlap(uint8_t*a,size_t an,uint8_t*b,size_t bn){return uintptr_t(a)<uintptr_t(b)+bn&&uintptr_t(b)<uintptr_t(a)+an;}
struct Operation {
 uint32_t kind=0,src=0,dst=0;std::vector<uint32_t> params;
 uint8_t*target=nullptr,*source=nullptr;size_t size=0,inputSize=0;bool record=false,live=false;std::vector<uint8_t> input,before;
 void start(n3ds_Machine*m,uint32_t k,uint32_t s,uint64_t sn,uint32_t d,uint64_t dn,std::vector<uint32_t>p){
  if(!recording&&!enabled)return;
  auto sp=sn?range(m,s,sn):nullptr;target=range(m,d,dn);
  if((sn&&!sp)||!target||!dn||(recording&&stream.size()+sn+dn*2+p.size()*4+256>limit)){if(recording)dropped++;target=nullptr;return;}
  kind=k;src=s;dst=d;size=dn;params=std::move(p);params.push_back(sn&&overlap(sp,sn,target,dn));
  source=sp;inputSize=sn;record=recording;live=enabled&&!m->OnRead&&!m->OnWrite&&!m->OnPixel&&!m->HidTrace&&!m->Profile;if(record){if(sn)input.assign(sp,sp+sn);before.assign(target,target+dn);}
 }
 static Operation fill(n3ds_Machine*m,uint32_t s,uint32_t value,uint32_t e,uint32_t ctl){
  Operation o;if(!active(m,1))return o;s=n3ds_Machine_gpuAddrToVirt(m,s);e=n3ds_Machine_gpuAddrToVirt(m,e);
  uint32_t width=(ctl>>8)&3;if(width==3||e<s){dropped++;return o;}uint32_t unit=width+2;
  o.start(m,1,0,0,s,e-s,{unit,value,uint32_t((e-s)/unit*unit)});return o;
 }
 static Operation copy(n3ds_Machine*m,uint32_t s,uint32_t d,uint32_t n,uint32_t in,uint32_t out){
  Operation o;if(!active(m,2))return o;s=n3ds_Machine_gpuAddrToVirt(m,s);d=n3ds_Machine_gpuAddrToVirt(m,d);
  uint32_t iw=(in&65535)*2,ig=(in>>16)*2,ow=(out&65535)*2,og=(out>>16)*2;
  if(!iw&&!ig)iw=n;if(!ow&&!og)ow=n;if(!iw||!ow||!n||n%iw){dropped++;return o;}
  uint64_t sn=uint64_t(n)+uint64_t((n-1)/iw)*ig,dn=uint64_t(n)+uint64_t((n-1)/ow)*og;
  o.start(m,2,s,sn,d,dn,{n,iw,ig,ow,og});return o;
 }
 static Operation display(n3ds_Machine*m,uint32_t s,uint32_t d,uint32_t in,uint32_t out,uint32_t flags){
  Operation o;if(!active(m,3))return o;s=n3ds_Machine_gpuAddrToVirt(m,s);d=n3ds_Machine_gpuAddrToVirt(m,d);
  uint32_t sw=in&65535,sh=in>>16,dw=out&65535,dh=out>>16,fmt=(flags>>12)&7,bpp=fmt==0?4:fmt==1?3:2;
  if(!sw||!sh||!dw||!dh||fmt>4||(flags&0x03000702)){dropped++;return o;}
  o.start(m,3,s,uint64_t((sw+7)&~7u)*((sh+7)&~7u)*4,d,uint64_t(dw)*dh*bpp,{sw,sh,dw,dh,flags});return o;
 }
 static Operation stencil(n3ds_GPU*g,n3ds_fbState*fb,n3ds_lightState*ls,n3ds_tevState*tv,Slice<n3ds_rasterTri>tris){
  Operation o;if(!active(g->m,4))return o;if(recording)draws++;
  // Exact first draw subset: stencil NEVER rejects every covered fragment.
  // Integer half-pixel coordinates keep every reference edge product exact.
  auto cfg=g->Regs[0x105];
  if(!(cfg&1)||((cfg>>4)&7)!=0||(g->Regs[0x116]&3)!=3||fb->shadowMode||ls->enabled||tv->texEnable||tv->alphaTest||!tris.n||tris.n>1024||!fb->width||!fb->height||fb->width>1024||fb->height>1024||(fb->width%8)||(fb->height%8))return o;
  if(!std::get<1>(n3ds_tevState_run(tv,{},{},{},{})))return o;
  std::vector<uint32_t> p={fb->width,fb->height,cfg,g->Regs[0x106]&7,g->Regs[0x115],uint32_t(tris.n)};
  for(auto&t:tris){
   if(t.minX<0||t.minY<0||t.maxX>fb->width||t.maxY>fb->height||!(t.area>0))return o;
   p.insert(p.end(),{uint32_t(t.minX),uint32_t(t.maxX),uint32_t(t.minY),uint32_t(t.maxY)});
   for(auto*v:{&t.v0,&t.v1,&t.v2}){
    if(!std::isfinite(v->x)||!std::isfinite(v->y)||v->x<0||v->y<0||v->x>1024||v->y>1024||v->x*2!=std::floor(v->x*2)||v->y*2!=std::floor(v->y*2)||!(v->iw>=0x1p-20f&&v->iw<=0x1p20f))return o;
    for(auto c:v->col)if(!std::isfinite(c)||std::abs(c)>1e6f)return o;
    p.push_back(uint32_t(v->x*2));p.push_back(uint32_t(v->y*2));
   }
  }
  o.start(g->m,4,0,0,fb->depthAddr,uint64_t(fb->width)*fb->height*4,std::move(p));return o;
 }
 bool execute(){
  if(!live||!enabled||!target||params.back()||size<4096)return false;
#ifdef __EMSCRIPTEN__
  if(rr_gpu_submit(kind,params.data(),params.size(),source,inputSize,target,size)){accelerated[kind]++;committedBytes+=size;return true;}
#endif
  return false;
 }
 ~Operation(){
  if(!target||!record)return;
  auto begin=stream.size();word(stream,0);word(stream,kind);word(stream,params.size());word(stream,input.size());word(stream,before.size());word(stream,size);word(stream,src);word(stream,dst);
  for(auto p:params)word(stream,p);bytes(stream,input);bytes(stream,before);bytes(stream,std::vector<uint8_t>(target,target+size));
  uint32_t n=stream.size()-begin;for(int j=0;j<4;j++)stream[begin+j]=n>>(j*8);
 }
 Operation()=default;Operation(const Operation&)=delete;Operation&operator=(const Operation&)=delete;
 Operation(Operation&&o):kind(o.kind),src(o.src),dst(o.dst),params(std::move(o.params)),target(o.target),source(o.source),size(o.size),inputSize(o.inputSize),record(o.record),live(o.live),input(std::move(o.input)),before(std::move(o.before)){o.target=nullptr;}
};
inline void begin(){recording=true;dropped=draws=0;stream.clear();word(stream,magic);word(stream,version);}
}
