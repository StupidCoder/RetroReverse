#pragma once
#define RR_CAPTURE_WRITE_CAP 4194304
#include "../../../../browser/core/capture.h"
#include "../../../nds/browser/core/runtime.h"
struct psp_Machine; struct BlockSource;
#include <numbers>
#include <cfloat>
constexpr float go_math_MaxFloat32=FLT_MAX;
constexpr double go_math_Pi=std::numbers::pi,go_math_E=std::numbers::e,go_math_Sqrt2=std::numbers::sqrt2,go_math_Log2E=std::numbers::log2e,go_math_Log10E=std::numbers::log10e,go_math_Ln2=std::numbers::ln2;
inline double go_math_Log(double x){return std::log(x);}inline double go_math_Log10(double x){return std::log10(x);}
inline double go_math_Log2(double x){return std::log2(x);}inline double go_math_Exp2(double x){return std::exp2(x);}
inline double go_math_Sin(double x){return std::sin(x);}inline double go_math_Cos(double x){return std::cos(x);}inline double go_math_Asin(double x){return std::asin(x);}
inline double go_math_Pow(double x,double y){return std::pow(x,y);}inline double go_math_Hypot(double x,double y){return std::hypot(x,y);}
inline double go_math_Min(double x,double y){return std::fmin(x,y);}inline double go_math_Max(double x,double y){return std::fmax(x,y);}
inline std::string go_os_Getenv(std::string){return {};}
inline std::tuple<uint64_t,Error>go_strconv_ParseUint(std::string s,int base,int bits){try{size_t n=0;auto v=std::stoull(s,&n,base);if(n!=s.size()||(bits<64&&v>>bits))throw 1;return {v,{}};}catch(...){return {0,{"invalid integer"}};}}
inline std::tuple<int64_t,Error>go_strconv_Atoi(std::string s){try{return {std::stoll(s),{}};}catch(...){return {0,{"invalid integer"}};}}
inline int64_t go_strings_Index(std::string s,std::string n){auto p=s.find(n);return p==s.npos?-1:p;}
inline int64_t go_strings_IndexByte(std::string s,uint8_t b){auto p=s.find(char(b));return p==s.npos?-1:p;}
inline std::string go_strings_ToUpper(std::string s){for(auto&c:s)c=std::toupper(uint8_t(c));return s;}
template<class T,size_t N>Slice<T>sub(std::array<T,N>*p,int64_t a,int64_t b){return sub(*p,a,b);}
template<class T,size_t N>int64_t len(std::array<T,N>*){return N;}
inline Slice<uint8_t>append(Slice<uint8_t>s,std::string v){return append(s,cast<Slice<uint8_t>>(v));}
inline double go_math_Mod(double a,double b){return std::fmod(a,b);}
inline bool go_strings_HasSuffix(std::string s,std::string n){return s.ends_with(n);}
inline std::tuple<int64_t,Error>go_strconv_ParseInt(std::string s,int b,int){try{return {std::stoll(s,nullptr,b),{}};}catch(...){return {0,{"invalid integer"}};}}
template<class...A>void go_fmt_Printf(std::string,A...){ }
template<class...A>void go_fmt_Println(A...){ }
#include "crypto.h"

#include "../../../../browser/core/profile.h"
inline uint8_t rrAnalogX=128,rrAnalogY=128;

struct rrGETraceScope {bool previous;rrGETraceScope():previous(rrcapture::trace.rendering){rrcapture::trace.rendering=true;}~rrGETraceScope(){rrcapture::trace.rendering=previous;rrcapture::trace.current=0;}};
