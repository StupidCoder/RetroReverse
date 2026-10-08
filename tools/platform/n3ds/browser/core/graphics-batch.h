#pragma once
// Batches exist only inside a PICA command-list call while the guest CPU is
// suspended. CPU-visible RAM is materialized before reads, fallbacks and return.
#ifdef __EMSCRIPTEN__
EM_ASYNC_JS(int,rr_gpu_batch,(const uint32_t* descriptors,uint32_t count,uint8_t* target,uint32_t size,uint8_t* depth,uint32_t* counters),{
 if(!Module.graphicsBatch)return 0;
 try{
  let before;
  const snapshot=()=>{if(!before){before=new Uint8Array(size*(depth?2:1));before.set(HEAPU8.subarray(target,target+size));if(depth)before.set(HEAPU8.subarray(depth,depth+size),size);}return before;};
  const packets=[];
  for(let i=0;i<count;i++){
   const at=(descriptors>>>2)+i*5,[kind,params,np,input,ni]=HEAPU32.subarray(at,at+5);
   packets.push({kind,params:Array.from(HEAPU32.subarray(params>>>2,(params>>>2)+np)),get input(){return HEAPU8.subarray(input,input+ni);},get before(){return snapshot();}});
  }
  const result=await Module.graphicsBatch(packets);
  if(!result?.supported||!(result.bytes instanceof Uint8Array)||result.bytes.length!==size*(depth?2:1)||!Array.isArray(result.counts)||result.counts.length!==count)return 0;
  for(let i=0;i<count;i++){
   const {drawn,depthKilled}=result.counts[i];
   if(!Number.isInteger(drawn)||!Number.isInteger(depthKilled)||drawn<0||depthKilled<0||drawn+depthKilled>packets[i].params[9]||(!depth&&depthKilled))return 0;
  }
  // Validate ALL draws before committing either surface or any guest counter.
  HEAPU8.set(result.bytes.subarray(0,size),target);if(depth)HEAPU8.set(result.bytes.subarray(size),depth);
  for(let i=0;i<count;i++){HEAPU32[(counters>>>2)+i*2]=result.counts[i].drawn;HEAPU32[(counters>>>2)+i*2+1]=result.counts[i].depthKilled;}
  return 1;
 }catch(e){Module.graphicsError=String(e);return 0;}
});
#endif
namespace rrbatch {
inline bool enabled=true,replaying=false;
inline n3ds_Machine* owner=nullptr;
struct Entry {
 rrgpu::Operation operation;
 n3ds_fbState fb; n3ds_lightState light; n3ds_tevState tev;
 std::vector<n3ds_rasterTri> triangles;
 std::unique_ptr<n3ds_GPU> reference;
 Entry(rrgpu::Operation&&o,n3ds_fbState*f,n3ds_lightState*l,n3ds_tevState*t,Slice<n3ds_rasterTri>tris,std::unique_ptr<n3ds_GPU>r)
  :operation(std::move(o)),fb(*f),light(*l),tev(*t),triangles(tris.begin(),tris.end()),reference(std::move(r)){}
};
inline std::vector<Entry> pending;
inline size_t inputBytes=0;
// Native fault/dependency fixtures use the identical queue and replay path.
inline std::function<bool(const std::vector<Entry>&,uint32_t*)> testSubmit;
inline void flush(){
 if(pending.empty()||replaying)return;
 auto entries=std::move(pending);pending.clear();inputBytes=0;
 auto&first=entries.front().operation;std::array<uint32_t,16> counts{};bool committed=false;
 {
  rrprof::Scope roundTrip(5,"WebGPU upload / wait / readback");
#ifdef __EMSCRIPTEN__
  std::vector<uint32_t> descriptors;descriptors.reserve(entries.size()*5);
  for(auto&e:entries){auto&o=e.operation;descriptors.insert(descriptors.end(),{o.kind,uint32_t(uintptr_t(o.params.data())),uint32_t(o.params.size()),uint32_t(uintptr_t(o.source)),uint32_t(o.inputSize)});}
  committed=rr_gpu_batch(descriptors.data(),entries.size(),first.target,first.size,first.depthTarget,counts.data())!=0;
#else
  if(testSubmit)committed=testSubmit(entries,counts.data());
#endif
 }
 if(committed){
  for(size_t i=0;i<entries.size();i++){auto&o=entries[i].operation;o.statsOwner->PixelsDrawn+=counts[i*2];o.statsOwner->DepthKilled+=counts[i*2+1];rrgpu::accelerated[4]++;}
  rrgpu::committedBytes+=first.size*(first.depthTarget?2:1);return;
 }
 // RAM still contains the pre-batch version. Replay immutable draw inputs in
 // order, using their original registers/LUTs/decoded textures. Never retry a
 // command list or roll back unrelated vertex, register or cache changes.
 replaying=true;const bool gpuEnabled=rrgpu::enabled;rrgpu::enabled=false;
 auto restore=defer([&](){rrgpu::enabled=gpuEnabled;replaying=false;});
 for(auto&e:entries){
  Slice<n3ds_rasterTri> tris;tris.p=e.triangles.data();tris.n=tris.c=e.triangles.size();
  n3ds_GPU_fill(e.reference.get(),&e.fb,&e.light,&e.tev,tris);
  e.operation.statsOwner->PixelsDrawn+=e.reference->PixelsDrawn;
  e.operation.statsOwner->DepthKilled+=e.reference->DepthKilled;
 }
}
inline void beforeRead(n3ds_Machine*m,uint32_t address,uint64_t size){
 if(pending.empty()||replaying||m!=owner||!size)return;
 if(uint64_t(address)+size>0x100000000ull){flush();return;}
 auto&o=pending.front().operation;
 // Check every page: a later mapping can override a page inside a larger
 // region. Unknown/partial ranges flush; pointer overlap catches aliases.
 while(size){
  const uint64_t n=std::min<uint64_t>(size,4096-(address&4095));auto p=rrgpu::range(m,address,n);
  if(!p||rrgpu::overlap(p,n,o.target,o.size)||(o.depthTarget&&rrgpu::overlap(p,n,o.depthTarget,o.size))){flush();return;}
  address+=n;size-=n;
 }
}
inline uint8_t read(n3ds_Machine*m,uint32_t a){beforeRead(m,a,1);return n3ds_Machine_Read(m,a);}
inline uint32_t readWord(n3ds_Machine*m,uint32_t a){beforeRead(m,a,4);return n3ds_Machine_ReadWord(m,a);}
inline auto readRange(n3ds_Machine*m,uint32_t a,uint32_t size){beforeRead(m,a,size);return n3ds_Machine_directRange(m,a,size);}
inline bool enqueue(rrgpu::Operation&o,n3ds_GPU*g,n3ds_fbState*fb,n3ds_lightState*ls,n3ds_tevState*tv,Slice<n3ds_rasterTri>tris){
 if(!owner||owner!=g->m||replaying||!o.live||o.record||!o.target||o.kind<6||o.params.back()||o.size<4096||o.inputSize>8u*1024*1024)return false;
 if(!pending.empty()){
  auto&first=pending.front().operation;
  if(first.target!=o.target||first.depthTarget!=o.depthTarget||first.size!=o.size||first.params[0]!=o.params[0]||first.params[1]!=o.params[1]||pending.size()>=8||inputBytes+o.inputSize>8u*1024*1024)flush();
 }
 auto reference=std::make_unique<n3ds_GPU>();reference->m=g->m;reference->Regs=g->Regs;
 if(ls->enabled){reference->LUT=g->LUT;reference->LUTDiff=g->LUTDiff;reference->lutSet=g->lutSet;}
 // Only already decoded textures are needed by an eligible packet. Retain
 // their owners until this synchronous command-list scope ends.
 for(int u=0;u<3;u++)if(tv->texEnable&(1u<<u)){
  auto[dim,param,addr,fmt]=n3ds_texUnitRegs(u);uint32_t w=(g->Regs[dim]>>16)&2047,h=g->Regs[dim]&2047;if(!w||!h)continue;
  n3ds_texKey key{n3ds_Machine_gpuAddrToVirt(g->m,g->Regs[addr]<<3),g->Regs[fmt]&15,w,h};
  auto[image,found]=lookup(g->texCache,key);if(!found)return false;reference->texCache[key]=image;
 }
 inputBytes+=o.inputSize;pending.emplace_back(std::move(o),fb,ls,tv,tris,std::move(reference));return true;
}
struct Scope {
 bool acquired=false;
 explicit Scope(n3ds_GPU*g){
  auto m=g->m;
  if(!enabled||!rrgpu::enabled||rrgpu::recording||owner||!m->SingleThreaded||m->OnRead||m->OnWrite||m->OnPixel||m->OnPICACmd||m->HidTrace||m->Profile||m->Verbose||m->GXCapture||m->picaLimit||g->TraceDraws||g->TraceUniforms||g->Census)return;
#ifndef __EMSCRIPTEN__
  if(!testSubmit)return;
#endif
  owner=m;acquired=true;
 }
 ~Scope(){if(acquired){flush();owner=nullptr;}}
};
}
