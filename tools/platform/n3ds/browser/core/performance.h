#pragma once
#include <array>
#include <map>
#include "../../../../browser/core/profile.h"

// Optional browser diagnostics. Nothing here is serialized or enables machine
// observation hooks. Vertex clocks are sampled; draw/counter totals are exact.
// The browser core runs its work pool serially, so one active draw is sufficient.
namespace rrperf {
inline uint32_t stride=0;
inline const char* rejection="unclassified";
struct Vertices {
 uint64_t draws=0,vertices=0,unique=0,samples=0,shaded=0,cacheHits=0;
 double fetchMs=0,shaderMs=0,clipMs=0;
};
inline std::map<uint64_t,Vertices> programs;
inline uint64_t programOverflow=0;
inline n3ds_GPU* cachedGPU=nullptr;
inline uint32_t cachedEpoch=0,cachedEntry=0;
inline uint64_t cachedHash=0;
inline uint64_t shaderHash(n3ds_GPU*g){
 uint32_t entry=g->Regs[0x2ba]&4095;
 if(cachedGPU==g&&cachedEpoch==g->shEpoch&&cachedEntry==entry)return cachedHash;
 uint64_t h=14695981039346656037ull;
 auto word=[&](uint32_t v){for(int i=0;i<4;i++){h=(h^uint8_t(v))*1099511628211ull;v>>=8;}};
 for(auto w:g->Code)word(w);for(auto w:g->Opdesc)word(w);word(entry);
 cachedGPU=g;cachedEpoch=g->shEpoch;cachedEntry=entry;return cachedHash=h;
}
struct Draw;
inline Draw* current=nullptr;
struct Draw {
 Draw* parent=nullptr;Vertices* result=nullptr;
 std::array<uint64_t,1024> seen;bool indexed=false;
 uint64_t count=0,unique=0,samples=0,shaded=0,cacheHits=0;
 double fetch=0,shader=0,clip=0;
 Draw(n3ds_GPU*g,bool indices):indexed(indices){
  if(!stride)return;
  const auto key=shaderHash(g);
  if(!programs.contains(key)&&programs.size()>=128){programOverflow++;return;}
  seen.fill(0);result=&programs[key];parent=current;current=this;
 }
 ~Draw(){if(!result)return;current=parent;result->draws++;result->vertices+=count;result->unique+=unique;result->samples+=samples;result->shaded+=shaded;result->cacheHits+=cacheHits;
  const double scale=samples?double(shaded)/samples:0;
  result->fetchMs+=fetch*scale;result->shaderMs+=shader*scale;result->clipMs+=clip;
 }
 void vertex(uint32_t index){count++;if(!indexed){unique++;return;}if(index<65536){auto&bits=seen[index/64];uint64_t mask=uint64_t(1)<<(index%64);if(!(bits&mask)){bits|=mask;unique++;}}}
};
struct Vertex {
 Draw* draw=nullptr;double start=0,shading=0;bool sampled=false,finished=false;
 Vertex(){if(!stride||!current)return;draw=current;sampled=draw->shaded%stride==0;if(sampled)start=rrprof::now();}
 void shade(uint32_t index){if(!draw)return;draw->vertex(index);draw->shaded++;if(sampled){shading=rrprof::now();draw->fetch+=shading-start;draw->samples++;}}
 void reuse(uint32_t index){if(draw){draw->vertex(index);draw->cacheHits++;}finished=true;}
 void finish(){if(finished)return;finished=true;if(draw&&sampled&&shading)draw->shader+=rrprof::now()-shading;}
 ~Vertex(){finish();}
};
struct Clip {
 Draw* draw=nullptr;double start=0;
 Clip(){if(stride&&current){draw=current;start=rrprof::now();}}
 void finish(){if(draw){draw->clip+=rrprof::now()-start;draw=nullptr;}}
 ~Clip(){finish();}
};
inline void refuse(const char* why){if(stride)rejection=why;}
struct Raster {
 uint64_t draws=0,triangles=0,work=0,pixels=0,depthKilled=0,shadowSamples=0;
 double ms=0;
};
inline std::map<std::string,Raster> fallbacks;
inline uint64_t fallbackOverflow=0;
struct Fallback {
 n3ds_GPU* gpu=nullptr;Raster* result=nullptr;
 double start=0;int64_t pixels=0,killed=0,shadows=0;
 Fallback(n3ds_GPU*g,n3ds_fbState*fb,n3ds_lightState*ls,n3ds_tevState*tv,Slice<n3ds_rasterTri>tris,bool submitted){
  if(!stride||!tris.n)return;
  std::ostringstream key;key<<(submitted?"gpu-refused":rejection)<<";lit="<<ls->enabled<<";depth="<<fb->depthTest<<"/"<<fb->depthFunc<<";write="<<fb->depthWr<<";stencil="<<g->Regs[0x105]<<";shadow="<<fb->shadowMode<<";tex0="<<((g->Regs[0x83]>>28)&7)<<";textures="<<tv->texEnable;
  const auto k=key.str();if(!fallbacks.contains(k)&&fallbacks.size()>=256){fallbackOverflow++;return;}
  gpu=g;result=&fallbacks[k];result->draws++;result->triangles+=tris.n;
  for(auto&t:tris)result->work+=uint64_t(std::max<int64_t>(0,t.maxX-t.minX))*std::max<int64_t>(0,t.maxY-t.minY);
  pixels=g->PixelsDrawn;killed=g->DepthKilled;shadows=g->ShadowSamples;start=rrprof::now();
 }
 ~Fallback(){if(!result)return;result->ms+=rrprof::now()-start;result->pixels+=gpu->PixelsDrawn-pixels;result->depthKilled+=gpu->DepthKilled-killed;result->shadowSamples+=gpu->ShadowSamples-shadows;}
};
inline bool enable(uint32_t sampleStride){
 if(current||sampleStride>1024)return false;
 stride=sampleStride;programs.clear();fallbacks.clear();programOverflow=fallbackOverflow=0;cachedGPU=nullptr;rejection="unclassified";return true;
}
inline std::string json(){
 std::ostringstream s;s.precision(12);s<<"{\"sampleStride\":"<<stride<<",\"vertexTimesEstimated\":true,\"programOverflow\":"<<programOverflow<<",\"fallbackOverflow\":"<<fallbackOverflow<<",\"programs\":[";bool comma=false;
 for(auto&[key,v]:programs){if(comma)s<<',';comma=true;s<<"{\"hash\":\""<<std::hex<<key<<std::dec<<"\",\"draws\":"<<v.draws<<",\"vertices\":"<<v.vertices<<",\"unique\":"<<v.unique<<",\"samples\":"<<v.samples<<",\"shaderInvocations\":"<<v.shaded<<",\"cacheHits\":"<<v.cacheHits<<",\"fetchMs\":"<<v.fetchMs<<",\"shaderMs\":"<<v.shaderMs<<",\"clipMs\":"<<v.clipMs<<'}';}
 s<<"],\"fallbacks\":[";comma=false;
 for(auto&[key,v]:fallbacks){if(comma)s<<',';comma=true;s<<"{\"reasonAndFeatures\":\""<<key<<"\",\"draws\":"<<v.draws<<",\"triangles\":"<<v.triangles<<",\"boundingBoxWork\":"<<v.work<<",\"pixels\":"<<v.pixels<<",\"depthKilled\":"<<v.depthKilled<<",\"shadowSamples\":"<<v.shadowSamples<<",\"ms\":"<<v.ms<<'}';}
 s<<"]}";return s.str();
}
}
