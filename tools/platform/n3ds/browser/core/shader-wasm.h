#pragma once
#ifdef __EMSCRIPTEN__
EM_JS(void,rr_pica_prepare,(uint32_t id,const uint32_t* code,const uint32_t* desc,uint32_t entry),{
 Module.picaShaders?.prepare(id,HEAPU32.slice(code>>>2,(code>>>2)+4096),HEAPU32.slice(desc>>>2,(desc>>>2)+128),entry,wasmMemory);
});
EM_JS(int,rr_pica_ready,(uint32_t id),{return Module.picaShaders?.ready(id)?1:0;});
EM_JS(int,rr_pica_run,(uint32_t id,const void* input,void* output,uint32_t n,const void* uniforms,const void* integers,uint32_t booleans),{
 return Module.picaShaders?.run(id,input,output,n,uniforms,integers,booleans)||0;
});
#endif
namespace rrshader {
using Values=rrvertex::Attributes;
struct Program {uint64_t hash;uint32_t entry;std::array<uint32_t,4096> code;std::array<uint32_t,128> desc;};
inline std::vector<Program> programs;
inline n3ds_GPU*lastGPU=nullptr;inline uint32_t lastEpoch=0,lastEntry=0,lastID=0;
inline void invalidate(){lastGPU=nullptr;}
#ifdef RR_SHADER_TEST
inline void(*prepareTest)(uint32_t,const Program&)=nullptr;
inline bool(*readyTest)(uint32_t)=nullptr;
inline bool(*runTest)(uint32_t,const Values*,Values*,uint32_t,n3ds_GPU*)=nullptr;
#endif
inline uint32_t program(n3ds_GPU*g){
 const uint32_t entry=g->Regs[0x2ba]&4095;
 if(lastGPU==g&&lastEpoch==g->shEpoch&&lastEntry==entry)return lastID;
 const auto hash=rrperf::shaderHash(g);
 uint32_t id=0;
 for(size_t i=0;i<programs.size();i++){auto&p=programs[i];if(p.hash==hash&&p.entry==entry&&p.code==g->Code&&p.desc==g->Opdesc){id=i+1;break;}}
 if(!id&&programs.size()<128){
  programs.push_back({hash,entry,g->Code,g->Opdesc});id=programs.size();
#ifdef __EMSCRIPTEN__
  rr_pica_prepare(id,g->Code.data(),g->Opdesc.data(),entry);
#elif defined(RR_SHADER_TEST)
  if(prepareTest)prepareTest(id,programs.back());
#endif
 }
 lastGPU=g;lastEpoch=g->shEpoch;lastEntry=entry;return lastID=id;
}
inline bool ready(uint32_t id){
#ifdef __EMSCRIPTEN__
 return rr_pica_ready(id);
#elif defined(RR_SHADER_TEST)
 return readyTest&&readyTest(id);
#else
 return false;
#endif
}
inline bool execute(uint32_t id,const Values*inputs,Values*outputs,uint32_t count,n3ds_GPU*g){
#ifdef __EMSCRIPTEN__
 return rr_pica_run(id,inputs,outputs,count,g->Float.data(),g->Int.data(),g->Bool);
#elif defined(RR_SHADER_TEST)
 return runTest&&runTest(id,inputs,outputs,count,g);
#else
 return false;
#endif
}
inline std::vector<Values> inputs,outputs;
inline std::vector<uint32_t> mapping,firstOutputs;
inline bool batch(rrvertex::Draw&fast,uint32_t first,uint32_t count,uint64_t perm,int64_t maxIn,Slice<n3ds_vsOut>outs){
 if(!fast.eligible||!rrvertex::reuseEnabled||!rrvertex::fetchEnabled||!count||count>32768)return false;
 const uint32_t id=program(fast.gpu);if(!id||!ready(id))return false;
 // A failed attempt discards all scratch results and starts a fresh reuse
 // generation. Otherwise the interpreter could copy an uninitialized output.
 const uint32_t generation=fast.generation;
 auto discard=defer([&](){if(generation){if(++rrvertex::nextGeneration==0){rrvertex::generations.fill(0);++rrvertex::nextGeneration;}fast.generation=rrvertex::nextGeneration;}});
 mapping.resize(count);firstOutputs.clear();inputs.clear();
 const double began=rrperf::stride?rrprof::now():0;
 Values attrs{};
 for(uint32_t i=0;i<count;i++){
  uint32_t v=fast.indexed?fast.index(i):first+i;
  if(generation&&rrvertex::generations[v]==generation){mapping[i]=rrvertex::outputIndices[v];continue;}
  if(inputs.size()>=8192||!fast.fetch(v,&attrs))return false;
  mapping[i]=inputs.size();firstOutputs.push_back(i);inputs.emplace_back();n3ds_mapAttrsToInputs(&inputs.back(),&attrs,perm,maxIn);
  if(generation){rrvertex::generations[v]=generation;rrvertex::outputIndices[v]=mapping[i];}
 }
 outputs.resize(inputs.size());const double shade=rrperf::stride?rrprof::now():0;
 if(!execute(id,inputs.data(),outputs.data(),inputs.size(),fast.gpu))return false;
#ifdef RR_SHADER_VERIFY
 // Test build only: compare every accepted real-scene vertex before mapping.
 for(uint32_t i=0;i<inputs.size();i++){
  Values reference{};
  if(!n3ds_GPU_shaderRun(fast.gpu,&inputs[i],&reference,fast.gpu->Regs[0x2ba]&4095)||std::memcmp(&reference,&outputs[i],sizeof(Values)))throw std::runtime_error("Compiled PICA vertex differs from Reference");
 }
#endif
 // No observable result changes before the entire compiled batch succeeds.
 for(uint32_t i=0;i<inputs.size();i++)n3ds_GPU_mapOutputs(fast.gpu,&outputs[i],&outs[firstOutputs[i]]);
 for(uint32_t i=0;i<count;i++)if(i!=firstOutputs[mapping[i]])outs[i]=outs[firstOutputs[mapping[i]]];
 if(rrperf::current){auto*d=rrperf::current;d->count+=count;d->unique+=inputs.size();d->shaded+=inputs.size();d->cacheHits+=count-inputs.size();d->compiled+=inputs.size();d->batchFetch+=shade-began;d->batchShader+=rrprof::now()-shade;}
 return true;
}
}
