#pragma once
#include <zlib.h>
#ifdef __EMSCRIPTEN__
#include <emscripten.h>
EM_JS(int,rrReadFile,(uint8_t * dst, double offset, int size),{
 try {var f=Module.discFile;if(!f)throw new Error('Select a local PSP image first');
 var b=new Uint8Array(new FileReaderSync().readAsArrayBuffer(f.slice(offset,offset+size)));
 if(b.length!==size)return 0;HEAPU8.set(b,dst);return 1;}catch(e){Module.discError=String(e);return 0;}
});
#else
inline std::ifstream rrDiscFile;
inline int rrReadFile(uint8_t*p,double off,int n){rrDiscFile.clear();rrDiscFile.seekg(uint64_t(off));rrDiscFile.read((char*)p,n);return rrDiscFile.gcount()==n;}
#endif
struct BlockSource {
 uint64_t bytes=0,total=0,windowAt=UINT64_MAX,reads=0,readBytes=0;unsigned align=0;
 bool compressed=false;std::vector<uint32_t>index;std::vector<uint8_t>window;
 struct Cache{uint32_t block=UINT32_MAX;Slice<uint8_t>data;};std::array<Cache,512>cache;
 void read(uint64_t off,uint8_t*p,size_t n){
  if(off>bytes||n>bytes-off)throw std::runtime_error("UMD read outside selected file");
  while(n){if(windowAt==UINT64_MAX||off<windowAt||off>=windowAt+window.size()){
   windowAt=off&~uint64_t(65535);size_t size=std::min<uint64_t>(262144,bytes-windowAt);window.resize(size);
   rrprof::Scope timing(4,"UMD file reads");if(!rrReadFile(window.data(),double(windowAt),size))throw std::runtime_error("Unable to read selected UMD file");reads++;readBytes+=size;
  }size_t count=std::min<uint64_t>(n,windowAt+window.size()-off);std::memcpy(p,window.data()+off-windowAt,count);off+=count;p+=count;n-=count;}
 }
 static uint32_t u32(const uint8_t*p){uint32_t v;std::memcpy(&v,p,4);return v;}
 void mount(uint64_t size){if(size<24||size>4ull*1024*1024*1024)throw std::runtime_error("Unsupported PSP image size");bytes=total=size;uint8_t h[24];read(0,h,24);compressed=std::memcmp(h,"CISO",4)==0;if(!compressed)return;
  total=uint64_t(u32(h+8))|uint64_t(u32(h+12))<<32;align=h[21];if(h[20]>1||u32(h+16)!=2048||align>16||total<2048||total>4ull*1024*1024*1024)throw std::runtime_error("Unsupported or invalid CSO header");
  auto blocks=(total+2047)/2048;index.resize(blocks+1);uint64_t start=std::max<uint32_t>(24,u32(h+4));read(start,(uint8_t*)index.data(),index.size()*4);
  uint64_t previous=start+index.size()*4;for(auto e:index){uint64_t at=uint64_t(e&0x7fffffff)<<align;if(at<previous||at>bytes)throw std::runtime_error("Invalid CSO block offsets");previous=at;}
 }
 Slice<uint8_t>block(uint32_t n){if(uint64_t(n)*2048>=total)throw std::runtime_error("UMD block outside image");auto&c=cache[n%cache.size()];if(c.block==n)return c.data;
  auto out=Slice<uint8_t>::make(2048);if(!compressed)read(uint64_t(n)*2048,out.p,std::min<uint64_t>(2048,total-uint64_t(n)*2048));else{
   uint64_t off=uint64_t(index[n]&0x7fffffff)<<align,end=uint64_t(index[n+1]&0x7fffffff)<<align;
   if(index[n]&0x80000000){if(end-off<2048)throw std::runtime_error("Truncated raw CSO block");read(off,out.p,2048);}else{
    if(end<=off||end-off>131072)throw std::runtime_error("Invalid compressed CSO block");std::vector<uint8_t>in(end-off);read(off,in.data(),in.size());z_stream z{};z.next_in=in.data();z.avail_in=in.size();z.next_out=out.p;z.avail_out=2048;
    if(inflateInit2(&z,-15)!=Z_OK)throw std::runtime_error("Cannot initialize CSO decoder");int result=inflate(&z,Z_FINISH);inflateEnd(&z);if((result!=Z_STREAM_END&&result!=Z_BUF_ERROR)||z.total_out!=2048)throw std::runtime_error("Invalid DEFLATE data in CSO block");
   }
  }c.block=n;c.data=out;return out;
 }
};
inline std::tuple<Slice<uint8_t>,Error>blockSource_ReadBlock(BlockSource*s,int64_t n){try{if(!s||n<0)throw std::runtime_error("No UMD image mounted");return {s->block(n),{}};}catch(const std::exception&e){return {{},{e.what()}};}}
