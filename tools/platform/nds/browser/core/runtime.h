#pragma once
#include "../../../threedo/browser/core/runtime.h"
using time_Duration=int64_t;
inline int64_t time_Duration_Nanoseconds(time_Duration t){return t;}
inline double time_Duration_Seconds(time_Duration t){return t/1e9;}
struct binary_littleEndian{};inline binary_littleEndian go_binary_LittleEndian;
inline uint16_t le_Uint16(const Slice<uint8_t>&s){return uint16_t(s[0])|uint16_t(s[1])<<8;}
inline uint32_t le_Uint32(const Slice<uint8_t>&s){return uint32_t(s[0])|uint32_t(s[1])<<8|uint32_t(s[2])<<16|uint32_t(s[3])<<24;}
inline uint64_t le_Uint64(const Slice<uint8_t>&s){uint64_t v=0;for(int i=0;i<8;i++)v|=uint64_t(s[i])<<(i*8);return v;}
inline void le_PutUint16(const Slice<uint8_t>&s,uint16_t v){s[0]=v;s[1]=v>>8;}
inline void le_PutUint32(const Slice<uint8_t>&s,uint32_t v){for(int i=0;i<4;i++)s[i]=v>>(i*8);}
inline void le_PutUint64(const Slice<uint8_t>&s,uint64_t v){for(int i=0;i<8;i++)s[i]=v>>(i*8);}
inline uint32_t go_bits_ReverseBytes32(uint32_t n){return __builtin_bswap32(n);}
inline int64_t go_bits_LeadingZeros32(uint32_t n){return std::countl_zero(n);}
inline Error go_errors_New(std::string s){return {s};}
inline std::string go_strings_TrimPrefix(std::string s,const std::string&p){return s.starts_with(p)?s.substr(p.size()):s;}
inline void image_RGBA_Set(image_RGBA*p,int64_t x,int64_t y,color_RGBA c){auto n=image_RGBA_PixOffset(p,x,y);p->Pix[n]=c.R;p->Pix[n+1]=c.G;p->Pix[n+2]=c.B;p->Pix[n+3]=c.A;}
inline constexpr int64_t go_math_MinInt64=INT64_MIN;
template<>inline Slice<int32_t> cast<Slice<int32_t>,std::string>(std::string s){Slice<int32_t> out;for(size_t i=0;i<s.size();){uint32_t v=uint8_t(s[i++]);int more=v<128?0:v<224?1:v<240?2:3;v&=more?((1u<<(6-more))-1):255;while(more--&&i<s.size())v=(v<<6)|(uint8_t(s[i++])&63);out=append(out,int32_t(v));}return out;}
inline Slice<uint16_t>go_utf16_Encode(Slice<int32_t>s){Slice<uint16_t>out;for(auto c:s){if(c<0||c>0x10ffff||(c>=0xd800&&c<=0xdfff))c=0xfffd;if(c<0x10000)out=append(out,uint16_t(c));else{c-=0x10000;out=append(out,uint16_t(0xd800+(c>>10)));out=append(out,uint16_t(0xdc00+(c&1023)));}}return out;}

template<class T>Slice<T>fullSub(Slice<T>s,int64_t lo,int64_t hi,int64_t cap){if(lo<0||hi<lo||cap<hi||cap>s.c)throw std::runtime_error("slice capacity bounds");s=sub(s,lo,hi);s.c=cap-lo;return s;}
