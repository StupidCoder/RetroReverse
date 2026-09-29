#pragma once
#include <string>
#include <cstdint>
#define RR_OS_FILE_DEFINED
struct os_File { std::string name; int64_t offset{}; bool writable{},closed{}; };
#include "console-runtime.h"
struct RRX86Bus {
 void*ctx{}; uint8_t(*read)(void*,uint32_t){}; void(*write)(void*,uint32_t,uint8_t){};
 uint8_t*(*span)(void*,uint32_t,uint32_t,bool){};
 RRX86Bus()=default;
 template<class T>RRX86Bus(T*t):ctx(t),read([](void*p,uint32_t a){return rrRead(static_cast<T*>(p),a);}),write([](void*p,uint32_t a,uint8_t v){rrWrite(static_cast<T*>(p),a,v);}),span([](void*p,uint32_t a,uint32_t n,bool w){return rrSpan(static_cast<T*>(p),a,n,w);}){}
};
inline uint8_t rrBus_Read(RRX86Bus b,uint32_t a){return b.read(b.ctx,a);}
inline void rrBus_Write(RRX86Bus b,uint32_t a,uint8_t v){b.write(b.ctx,a,v);}
inline double go_math_Atan2(double y,double x){return std::atan2(y,x);}
inline std::tuple<double,int64_t>go_math_Frexp(double x){int e;auto f=std::frexp(x,&e);return {f,e};}
inline double go_math_Remainder(double x,double y){return std::remainder(x,y);}
inline double go_math_Tan(double x){return std::tan(x);}
inline std::string go_strings_ReplaceAll(std::string a,std::string b,std::string c){size_t pos=0;while((pos=a.find(b,pos))!=a.npos){a.replace(pos,b.size(),c);pos+=c.size();}return a;}
inline std::string go_strings_Join(Slice<std::string>s,std::string sep){std::string o;for(auto v:s){if(!o.empty())o+=sep;o+=v;}return o;}
inline int64_t go_strings_LastIndex(std::string s,std::string n){auto p=s.rfind(n);return p==s.npos?-1:p;}
inline Error go_io_EOF{"EOF"};
using fs_FileMode=uint32_t;
