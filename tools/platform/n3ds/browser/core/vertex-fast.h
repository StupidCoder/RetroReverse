#pragma once

// Browser Experimental mode only. A draw owns its cache generation; no guest
// bytes, uniforms, shader code or output mapping can change during vertex work.
// This state is disposable and deliberately absent from save states.
namespace rrvertex {
inline std::array<uint32_t,65536> generations{},outputIndices{};
inline uint32_t nextGeneration=0;
inline bool reuseEnabled=true,fetchEnabled=true; // Native differential tests.
using Attributes=std::array<std::array<float,4>,16>;

// Require the same ordinary, fully indexed RAM region on every covered page.
// This also catches a later mapping overriding a page inside a larger region.
// Partial regions and unusual boundaries retain the byte-wise Reference path.
inline uint8_t* ordinary(n3ds_Machine*m,uint32_t address,uint64_t size){
 if(!size||!m->pages||uint64_t(address)+size>0x100000000ull)return nullptr;
 auto r=m->pages[address>>12];
 if(!r||address<r->base||uint64_t(address-r->base)+size>uint64_t(r->data.n))return nullptr;
 for(uint32_t page=address>>12,last=uint32_t((uint64_t(address)+size-1)>>12);page<=last;page++)
  if(m->pages[page]!=r)return nullptr;
 return r->data.p+(address-r->base);
}
struct Component {uint16_t offset=0;uint8_t attribute=0,format=0,count=0;};
struct Buffer {
 uint32_t physical=0,stride=0;uint16_t width=0,count=0;
 std::array<Component,15> components{};
};
struct Draw {
 n3ds_GPU*gpu=nullptr;uint8_t*indices=nullptr;
 uint32_t generation=0,indexWidth=0,fixedMask=0;size_t bufferCount=0;
 std::array<Buffer,12> buffers{};
 bool eligible=false,indexed=false,cacheable=false;
 Draw(n3ds_GPU*g,bool isIndexed,uint32_t physical,uint64_t format,uint32_t fixed,
      Slice<n3ds_loaderBuf> loaders,Slice<int64_t> components,
      uint32_t indexAddress,bool index16,uint32_t count,bool trace):indexed(isIndexed){
  auto m=g->m;
  if(!rrgpu::enabled||!m->SingleThreaded||m->OnRead||m->OnWrite||m->OnPixel||m->HidTrace||m->Profile||m->Verbose||trace||g->TraceDraws||g->TraceUniforms||g->Census)return;
  if(isIndexed){indexWidth=index16?2:1;indices=ordinary(m,indexAddress,uint64_t(count)*indexWidth);if(!indices)return;}
  if(loaders.n>12)return;
  for(auto&loader:loaders){
   if(loader.n>15||loader.n<0||loader.first<0||loader.first+loader.n>components.n||uint64_t(physical)+loader.off>UINT32_MAX)return;
   auto&b=buffers[bufferCount++];b.physical=physical+loader.off;b.stride=loader.stride;
   uint32_t offset=0;
   for(int64_t j=0;j<loader.n;j++){
    const int64_t c=components[loader.first+j];if(c<0||c>15)return;
    if(c>=12){offset+=uint32_t(c-11)*4;continue;}
    const uint32_t f=(format>>(4*c))&15,n=(f>>2)+1,type=f&3;
    b.components[b.count++]={uint16_t(offset),uint8_t(c),uint8_t(type),uint8_t(n)};
    offset+=n*(type<2?1:type==2?2:4);
   }
   b.width=offset;
  }
  gpu=g;fixedMask=fixed;eligible=true;
  if(indexed&&reuseEnabled){if(++nextGeneration==0){generations.fill(0);++nextGeneration;}generation=nextGeneration;}
 }
 uint32_t index(uint32_t i)const{return indexWidth==2?uint32_t(indices[i*2])|(uint32_t(indices[i*2+1])<<8):indices[i];}
 bool reuse(uint32_t vertex,uint32_t output,Slice<n3ds_vsOut> outs){
  cacheable=false;
  if(!generation||vertex>=65536||generations[vertex]!=generation)return false;
  outs[output]=outs[outputIndices[vertex]];return true;
 }
 bool fetch(uint32_t vertex,Attributes*attrs){
  cacheable=false;if(!eligible)return false;
  std::array<const uint8_t*,12> sources{};
  for(size_t i=0;i<bufferCount;i++){
   auto&b=buffers[i];if(!b.count)continue;
   uint64_t physical=uint64_t(b.physical)+uint64_t(vertex)*b.stride;
   if(physical>UINT32_MAX)return false;
   sources[i]=ordinary(gpu->m,n3ds_Machine_gpuAddrToVirt(gpu->m,uint32_t(physical)),b.width);
   if(!sources[i])return false;
  }
  cacheable=true;
  if(!fetchEnabled)return false; // Still validates sources for isolated reuse tests.
  for(int a=0;a<16;a++)(*attrs)[a]=(fixedMask&(1u<<a))?gpu->fixedVal[a]:std::array<float,4>{0,0,0,1};
  for(size_t i=0;i<bufferCount;i++){
   auto&b=buffers[i];
   for(size_t j=0;j<b.count;j++){
    const auto&c=b.components[j];const auto*p=sources[i]+c.offset;
    std::array<float,4> value{0,0,0,1};
    for(int k=0;k<c.count;k++)switch(c.format){
     case 0:value[k]=int8_t(*p++);break;
     case 1:value[k]=*p++;break;
     case 2:value[k]=int16_t(uint16_t(p[0])|(uint16_t(p[1])<<8));p+=2;break;
     case 3:{uint32_t bits=uint32_t(p[0])|(uint32_t(p[1])<<8)|(uint32_t(p[2])<<16)|(uint32_t(p[3])<<24);std::memcpy(&value[k],&bits,4);p+=4;break;}
    }
    (*attrs)[c.attribute]=value;
   }
  }
  return true;
 }
 void remember(uint32_t vertex,uint32_t output){
  if(generation&&cacheable&&vertex<65536){generations[vertex]=generation;outputIndices[vertex]=output;}
 }
};
}
