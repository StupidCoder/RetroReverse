#pragma once
#include "../../../../browser/core/x86-runtime.h"
#include "../../../../browser/core/local-files.h"
#include "../../../../browser/core/local-disc.h"
struct xbox_xidDevice;
inline bool go_errors_Is(Error a,Error b){return a.text==b.text;}
inline std::string go_strings_TrimSuffix(std::string s,std::string p){return s.ends_with(p)?s.substr(0,s.size()-p.size()):s;}
inline std::tuple<RRFileInfo,Error>os_File_Stat(os_File*f){return go_os_Stat(f->name);}
inline Slice<int32_t>go_utf16_Decode(Slice<uint16_t>s){Slice<int32_t>r;for(int64_t i=0;i<s.n;i++){int32_t v=s[i];if(v>=0xd800&&v<=0xdbff&&i+1<s.n&&s[i+1]>=0xdc00&&s[i+1]<=0xdfff)v=0x10000+((v-0xd800)<<10)+(s[++i]-0xdc00);else if(v>=0xd800&&v<=0xdfff)v=0xfffd;r=append(r,v);}return r;}
struct SHA1;
#include "sha1.h"
inline void be_PutUint64(Slice<uint8_t>b,uint64_t x){for(int i=0;i<8;i++)b[i]=x>>(56-i*8);}

template<>inline std::string cast<std::string,uint8_t>(uint8_t c){if(c<128)return std::string(1,char(c));return std::string{char(0xc0|(c>>6)),char(0x80|(c&63))};}

template<>inline std::string cast<std::string,Slice<int32_t>>(Slice<int32_t>s){std::string out;for(auto c:s){if(c<0||c>0x10ffff||(c>=0xd800&&c<=0xdfff))c=0xfffd;if(c<128)out+=char(c);else if(c<2048){out+=char(0xc0|(c>>6));out+=char(0x80|(c&63));}else if(c<65536){out+=char(0xe0|(c>>12));out+=char(0x80|((c>>6)&63));out+=char(0x80|(c&63));}else{out+=char(0xf0|(c>>18));out+=char(0x80|((c>>12)&63));out+=char(0x80|((c>>6)&63));out+=char(0x80|(c&63));}}return out;}
#define RR_CAPTURE_WRITE_CAP 8388608
#define RR_CAPTURE_EVENT_CAP 131072
#include "../../../../browser/core/console-trace.h"
