#pragma once
#define RR_CAPTURE_WRITE_CAP 8388608
#define RR_CAPTURE_CHUNKED_WRITES
#define RR_CAPTURE_EVENT_CAP 262144
#define RR_CAPTURE_META_CAP (32 * 1024 * 1024)
#include <array>
#include <ostream>
template<class T,size_t N>std::ostream&operator<<(std::ostream&o,const std::array<T,N>&a){o<<'[';for(auto v:a)o<<v<<' ';return o<<']';}
#include "../../../nds/browser/core/runtime.h"
struct sync_WaitGroup{};
// Only used for non-escaping render-target views while the draw owns the memory.
template<class T>Slice<T>borrowSub(const Slice<T>&s,int64_t lo,int64_t hi){
 if(lo<0||hi<lo||hi>s.c)throw std::runtime_error("slice bounds");
 Slice<T>v;v.p=s.p+lo;v.n=hi-lo;v.c=s.c-lo;return v;
}

struct n3ds_workPool{};
using image_NRGBA=image_RGBA;using color_NRGBA=color_RGBA;
inline image_NRGBA*go_image_NewNRGBA(image_Rectangle r){return go_image_NewRGBA(r);}
inline int64_t image_NRGBA_PixOffset(image_NRGBA*r,int64_t x,int64_t y){return image_RGBA_PixOffset(r,x,y);}
inline void image_NRGBA_Set(image_NRGBA*r,int64_t x,int64_t y,color_NRGBA c){image_RGBA_Set(r,x,y,c);}
inline void image_NRGBA_SetNRGBA(image_NRGBA*r,int64_t x,int64_t y,color_NRGBA c){image_RGBA_Set(r,x,y,c);}

template<class...A>void go_fmt_Printf(std::string f,A...args){std::cerr<<go_fmt_Sprintf(f,args...);}
template<class...A>void go_fmt_Fprintf(std::ostream*o,std::string f,A...args){if(o)*o<<go_fmt_Sprintf(f,args...);}

template<>inline std::string cast<std::string,Slice<int32_t>>(Slice<int32_t>v){std::string s;for(auto c:v){if(c<0||c>0x10ffff||(c>=0xd800&&c<=0xdfff))c=0xfffd;if(c<128)s+=char(c);else if(c<2048){s+=char(0xc0|(c>>6));s+=char(0x80|(c&63));}else if(c<65536){s+=char(0xe0|(c>>12));s+=char(0x80|((c>>6)&63));s+=char(0x80|(c&63));}else{s+=char(0xf0|(c>>18));s+=char(0x80|((c>>12)&63));s+=char(0x80|((c>>6)&63));s+=char(0x80|(c&63));}}return s;}
inline double go_math_Ldexp(double x,int64_t e){return std::ldexp(x,e);}
template<class...A>void go_fmt_Fprintln(std::ostream*o,A...a){if(o){((*o<<a<<' '),...);*o<<'\n';}}
template<class...A>void go_fmt_Println(A...a){go_fmt_Fprintln(&std::cerr,a...);}
inline int32_t go_atomic_AddInt32(int32_t*p,int32_t n){return *p+=n;}
inline Slice<int32_t>go_utf16_Decode(Slice<uint16_t>v){Slice<int32_t>s;for(int64_t i=0;i<v.n;i++){uint32_t c=v[i];if(c>=0xd800&&c<=0xdbff&&i+1<v.n&&v[i+1]>=0xdc00&&v[i+1]<=0xdfff){c=0x10000+((c-0xd800)<<10)+(v[++i]-0xdc00);}else if(c>=0xd800&&c<=0xdfff)c=0xfffd;s=append(s,int32_t(c));}return s;}
inline bool go_strings_ContainsAny(std::string s,std::string chars){auto a=cast<Slice<int32_t>>(s),b=cast<Slice<int32_t>>(chars);for(auto c:a)for(auto d:b)if(c==d)return true;return false;}
template<class F>int64_t go_sort_Search(int64_t n,F f){int64_t a=0,b=n;while(a<b){auto m=a+(b-a)/2;if(f(m))b=m;else a=m+1;}return a;}
