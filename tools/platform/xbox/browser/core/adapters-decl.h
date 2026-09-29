#pragma once
#include <unordered_set>
template<class...A>void x86_CPU_Halt(x86_CPU*c,std::string f,A...a){c->Halted=true;c->HaltReason=go_fmt_Sprintf(f,a...);}
template<class...A>void xbox_Machine_logf(xbox_Machine*m,std::string f,A...a){if(m->Log.n<1024)m->Log=append(m->Log,go_fmt_Sprintf(f,a...));}
uint8_t rrRead(xbox_Machine*,uint32_t);void rrWrite(xbox_Machine*,uint32_t,uint8_t);
inline int64_t xbox_rasterWorkers(){return 1;}inline int64_t xbox_vshWorkers(){return 1;}
inline void xbox_parallelChunks(int64_t lo,int64_t hi,int64_t,std::function<void(int64_t,int64_t)>f){f(lo,hi);}
void xbox_pgraph_rasterParallel(xbox_pgraph*,Slice<xbox_kelvinVtx>,Slice<std::array<int64_t,3>>);
SHA1*xbox_Machine_shaLoad(xbox_Machine*,uint32_t);void xbox_Machine_shaStore(xbox_Machine*,uint32_t,SHA1*);

inline std::string xbox_Machine_disasmAt(xbox_Machine*,uint32_t a){return go_fmt_Sprintf("PC %08X",a);}
inline Slice<std::string>xbox_Machine_DisasmForward(xbox_Machine*,uint32_t,int64_t){return {};}
inline void xbox_pgraph_dumpTexShaderConfig(xbox_pgraph*,xbox_rasterState*,xbox_kelvinVtx*,xbox_kelvinVtx*,xbox_kelvinVtx*,float,float,float,float,float,float,float,int64_t,int64_t){}

inline void xbox_pgraph_dumpReceiverState(xbox_pgraph*){}

inline void xbox_sortU32(Slice<uint32_t>s){std::sort(s.begin(),s.end());}
template<class...A>Error xbox_Machine_usbUnsupported(xbox_Machine*m,std::string f,A...a){auto message=go_fmt_Sprintf("usb: "+f,a...);x86_CPU_Halt(m->CPU,message);return {message};}
inline void rrXboxWrite(xbox_Machine*m,const Slice<uint8_t>&b,uint32_t a){if(b.p==m->RAM.p&&a<uint32_t(b.n))rrconsole::byte(a,b[a],m->CPU?m->CPU->Steps:0,m->CPU?m->CPU->SegBase[1]+m->CPU->instrIP:0);}
inline void rrXboxDraw(xbox_pgraph*g,const char*kind){auto&t=rrcapture::trace;std::vector<uint8_t>b(8192);for(int i=0;i<2048;i++)for(int j=0;j<4;j++)b[i*4+j]=g->Regs[i]>>(j*8);auto r=t.resource(b.data(),b.size());std::ostringstream s;s<<"{\"kind\":\""<<kind<<"\",\"primitive\":"<<g->prim<<",\"draw\":"<<g->Draws<<",\"colorBase\":"<<g->Regs[0x210/4]<<",\"registerSnapshot\":\"2048 Kelvin registers, little-endian 32-bit\"}";rrconsole::command(g->m->CPU->Steps,g->m->CPU->SegBase[1]+g->m->CPU->instrIP,s.str(),r);}
inline uint8_t*rrSpan(xbox_Machine*m,uint32_t a,uint32_t n,bool w){if((w&&(rrcapture::trace.active||m->onW))||(!w&&m->onR))return nullptr;uint32_t p=a&0x0fffffffu;auto window=a>>28;if(window!=0&&window!=8&&window!=11&&window!=13&&window!=15)return nullptr;if(p>uint64_t(m->RAM.n)||n>uint64_t(m->RAM.n)-p)return nullptr;return m->RAM.p+p;}

// Go's garbage collector released replaced texture entries. Give their C++
// counterparts an explicit lifetime instead of the machine-lifetime arena.
inline std::vector<std::unique_ptr<xbox_texImage>>rrTextures;
inline std::vector<std::unique_ptr<xbox_texEntry>>rrTextureEntries;
inline xbox_texImage*rrTextureNew(xbox_texImage v){auto p=std::make_unique<xbox_texImage>(std::move(v));auto*r=p.get();rrTextures.push_back(std::move(p));return r;}
inline xbox_texEntry*rrTextureEntryNew(xbox_texEntry v){auto p=std::make_unique<xbox_texEntry>(std::move(v));auto*r=p.get();rrTextureEntries.push_back(std::move(p));return r;}
inline void rrCollectTextures(xbox_pgraph*g){
 size_t bytes=0;for(auto&[k,e]:g->texCache)bytes+=e->img->pix.n+e->img->depth.n*4;
 while(len(g->texCache)>256||bytes>96*1024*1024){auto oldest=g->texCache.begin();for(auto it=g->texCache.begin();it!=g->texCache.end();++it)if(it->second->validated<oldest->second->validated)oldest=it;auto*e=oldest->second;bytes-=e->img->pix.n+e->img->depth.n*4;g->texCache.p->erase(oldest);}
 std::unordered_set<xbox_texEntry*>entries;std::unordered_set<xbox_texImage*>images;for(auto&[k,e]:g->texCache){entries.insert(e);images.insert(e->img);}for(auto*i:g->rast.texImg)if(i)images.insert(i);
 std::erase_if(rrTextureEntries,[&](auto&p){return !entries.contains(p.get());});std::erase_if(rrTextures,[&](auto&p){return !images.contains(p.get());});
}
