#pragma once
#include "generated.cpp"
#include "elf.h"
#include "../../../../browser/core/local-disc.h"
inline LocalDisc disc;
struct LocalSource{iso9660_Geometry geometry;};
inline LocalSource source;
inline std::tuple<Slice<uint8_t>,Error>localSource_ReadBlock(LocalSource*s,int64_t n){try{if(n<0)throw std::runtime_error("Negative disc block");auto b=Slice<uint8_t>::make(2048);disc.read(uint64_t(n)*s->geometry.SectorSize+s->geometry.DataOffset,b.p,2048);return {b,{}};}catch(const std::exception&e){return {{},{e.what()}};}}
inline std::tuple<iso9660_Geometry,bool>iso9660_Volume_Geometry(iso9660_Volume*v){return {v->src->geometry,true};}
inline ps2_Machine*rrBoot(uint64_t bytes){
 disc.mount(bytes);bool found=false;
 for(auto g:iso9660_geometries){source.geometry=g;auto[b,e]=localSource_ReadBlock(&source,16);if(!e&&b[0]==1&&std::memcmp(b.p+1,"CD001",5)==0&&b[6]==1&&le_Uint16(sub(b,128,130))==2048&&uint64_t(le_Uint32(sub(b,80,84)))*g.SectorSize<=bytes){found=true;break;}}
 if(!found)throw std::runtime_error("Select a PS2 ISO or supported raw BIN disc image");
 auto[v,ve]=iso9660_OpenVolume(&source);if(ve)throw std::runtime_error(ve.text);
 auto[cnf,ce]=iso9660_Volume_ReadFile(v,"SYSTEM.CNF");if(ce)throw std::runtime_error(ce.text);
 std::string exe;std::istringstream text(cast<std::string>(cnf));std::string line;
 while(std::getline(text,line)){auto p=line.find('=');if(p==line.npos||go_strings_TrimSpace(line.substr(0,p))!="BOOT2")continue;exe=go_strings_TrimSpace(line.substr(p+1));auto colon=exe.find(':');if(colon!=exe.npos)exe=exe.substr(colon+1);}
 if(exe.empty())throw std::runtime_error("SYSTEM.CNF has no BOOT2 executable");
 auto[b,e]=iso9660_Volume_ReadFile(v,exe);if(e)throw std::runtime_error(e.text);auto[elf,ee]=ps2_LoadELF(b);if(ee)throw std::runtime_error(ee.text);
 ps2_init_iopintr();ps2_init_iopkernel();auto*m=ps2_NewMachine();m->SingleThreaded=true;ps2_Machine_SetVolume(m,v);ps2_Machine_LoadExecutable(m,elf);return m;
}
inline uint32_t rrHash(const uint8_t*p,size_t n){uint32_t h=2166136261;for(size_t i=0;i<n;i++)h=(h^p[i])*16777619;return h;}
