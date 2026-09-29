#pragma once
// Side-effect-free physical storage descriptors and a bounded access recorder.
// Pointers are private to WASM; only copies cross the worker boundary.
#include <cstdint>
#include <vector>
#include <string>
#include <sstream>
namespace rrmem {
struct Region { std::string id,name,kind; const uint8_t* data; uint32_t size; std::string aliases; uint32_t base=0; std::string mappings; };
inline std::vector<Region> regions;
inline std::string reply;
inline bool supported=false,active=false;
inline uint32_t dropped=0,mask=3;
// Little-endian values in address order. kind: read=1, write=2, fetch=4,
// device read=9, device write=10, media consumption=17.
struct Event {uint32_t lo,hi,region,offset,value,size,kind,pc,address;};
static_assert(sizeof(Event)==36);
inline std::vector<Event> events;
constexpr size_t limit=524288;
inline void access(uint64_t clock,uint32_t region,uint32_t offset,uint32_t value,uint32_t size,uint32_t kind,uint32_t pc,uint32_t address){
 if(!active||!(mask&(kind&7)))return;
 if(events.size()>=limit){dropped++;return;}
 events.push_back({uint32_t(clock),uint32_t(clock>>32),region,offset,value,size,kind,pc,address});
}
inline void add(std::string id,std::string name,std::string kind,const uint8_t*p,uint32_t n,std::string aliases="[]",uint32_t base=0,std::string mappings=""){regions.push_back({id,name,kind,p,n,aliases,base,mappings});}
inline const char*describe(){
 std::ostringstream s;s<<"{\"schema\":1,\"activity\":"<<(supported?"true":"false")<<",\"regions\":[";
 for(size_t i=0;i<regions.size();i++){auto&r=regions[i];if(i)s<<',';s<<"{\"index\":"<<i<<",\"id\":\""<<r.id<<"\",\"name\":\""<<r.name<<"\",\"kind\":\""<<r.kind<<"\",\"size\":"<<r.size<<",\"base\":"<<r.base<<",\"aliases\":"<<r.aliases;if(!r.mappings.empty())s<<",\"mappings\":"<<r.mappings;s<<"}";}
 s<<"]}";reply=s.str();return reply.c_str();
}
}
