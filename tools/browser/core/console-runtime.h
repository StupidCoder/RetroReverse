#pragma once
#include "../../platform/nds/browser/core/runtime.h"
#include "profile.h"
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

#ifndef RR_OS_FILE_DEFINED
struct os_File{};
#endif
inline os_File*go_os_Stderr=nullptr;
template<class...A>void go_fmt_Fprintf(os_File*,std::string,A...){ }
template<class...A>void go_fmt_Fprintln(os_File*,A...){ }
inline double go_math_Ldexp(double x,int64_t n){return std::ldexp(x,n);}
inline uint32_t go_bits_RotateLeft32(uint32_t x,int64_t k){return std::rotl(x,int(k));}
inline int64_t go_bits_LeadingZeros64(uint64_t x){return std::countl_zero(x);}
inline int64_t time_Time_Sub(time_Time a,time_Time b){return a-b;}
inline double go_math_Logb(double x){return std::logb(x);}



inline uint16_t go_bits_ReverseBytes16(uint16_t x){return __builtin_bswap16(x);}
inline std::tuple<int64_t,Error>go_fmt_Sscanf(std::string s,std::string f,int64_t*x,int64_t*y){long long a=0,b=0;int n=sscanf(s.c_str(),"%lld,%lld",&a,&b);*x=a;*y=b;return {n,{}};}

using strings_Builder=std::string;
inline std::tuple<int64_t,Error>strings_Builder_WriteString(std::string*s,std::string v){*s+=v;return {int64_t(v.size()),{}};}
inline Error strings_Builder_WriteByte(std::string*s,uint8_t v){*s+=char(v);return {};}
inline std::string strings_Builder_String(std::string*s){return *s;}
inline image_Rectangle image_RGBA_Bounds(image_RGBA*i){return i->Rect;}
inline int64_t image_Rectangle_Dx(image_Rectangle r){return r.Max.X-r.Min.X;}
inline int64_t image_Rectangle_Dy(image_Rectangle r){return r.Max.Y-r.Min.Y;}
template<class F>int64_t go_sort_Search(int64_t n,F pred){int64_t lo=0;while(lo<n){auto mid=lo+(n-lo)/2;if(pred(mid))n=mid;else lo=mid+1;}return lo;}
template<class...A>void go_fmt_Fprintf(std::string*s,std::string f,A...a){*s+=go_fmt_Sprintf(f,a...);}

inline int64_t strings_Builder_Len(std::string*s){return s->size();}
// Borrowed endian-helper span: its consumer neither stores nor returns it.
// Retaining shared ownership for each CPU fetch and texel read is unnecessary.
template<class T>inline Slice<T>rrBorrow(const Slice<T>&s,int64_t lo,int64_t hi){if(lo<0||hi<lo||hi>s.c)throw std::runtime_error("slice bounds");Slice<T>v;v.p=s.p+lo;v.n=hi-lo;v.c=s.c-lo;return v;}
template<class T,size_t N>inline Slice<T>rrBorrow(std::array<T,N>&s,int64_t lo,int64_t hi){return sub(s,lo,hi);}
