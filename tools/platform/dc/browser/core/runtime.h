#pragma once
#include "../../../../browser/core/console-runtime.h"
struct dc_Disc;struct LocalSource;
inline constexpr int64_t go_math_MaxInt32=INT32_MAX,go_math_MinInt32=INT32_MIN;
inline std::tuple<double,int64_t>go_math_Frexp(double x){int e=0;double f=std::frexp(x,&e);return {f,e};}
inline double go_math_Atan(double x){return std::atan(x);}inline double go_math_Atan2(double y,double x){return std::atan2(y,x);}inline double go_math_Exp(double x){return std::exp(x);}
inline bool go_strings_ContainsAny(std::string a,std::string b){return a.find_first_of(b)!=a.npos;}
inline std::string go_strings_ReplaceAll(std::string a,std::string b,std::string c){size_t pos=0;while((pos=a.find(b,pos))!=a.npos){a.replace(pos,b.size(),c);pos+=c.size();}return a;}
inline Error go_io_EOF{"EOF"};

inline void print(std::string){}
using fs_FileMode=uint32_t;
inline Error go_os_WriteFile(std::string,Slice<uint8_t>,fs_FileMode){return {"Diagnostic file export unavailable"};}
inline std::string go_strings_Join(Slice<std::string>s,std::string sep){std::string o;for(auto v:s){if(!o.empty())o+=sep;o+=v;}return o;}
inline int64_t go_strings_LastIndexAny(std::string s,std::string a){auto p=s.find_last_of(a);return p==s.npos?-1:p;}

inline std::string go_strings_TrimSuffix(std::string s,std::string p){return s.ends_with(p)?s.substr(0,s.size()-p.size()):s;}

inline int64_t go_bytes_IndexByte(Slice<uint8_t>b,uint8_t c){for(int64_t i=0;i<b.n;i++)if(b[i]==c)return i;return -1;}
inline Slice<std::string>go_strings_Fields(std::string s){Slice<std::string>o;std::istringstream in(s);std::string w;while(in>>w)o=append(o,w);return o;}

inline std::tuple<double,double>go_math_Sincos(double x){return {std::sin(x),std::cos(x)};}

#define RR_CAPTURE_WRITE_CAP 8388608
#define RR_CAPTURE_EVENT_CAP 131072
#include "../../../../browser/core/console-trace.h"
