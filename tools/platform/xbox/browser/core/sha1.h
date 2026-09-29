#pragma once
// Streaming SHA-1. The state encoding matches Go crypto/sha1 for oracle snapshots.
struct SHA1{std::array<uint32_t,5>h{0x67452301,0xefcdab89,0x98badcfe,0x10325476,0xc3d2e1f0};std::array<uint8_t,64>buffer{};uint64_t length{};
 void block(const uint8_t*b){uint32_t w[80];for(int i=0;i<16;i++)w[i]=uint32_t(b[i*4])<<24|uint32_t(b[i*4+1])<<16|uint32_t(b[i*4+2])<<8|b[i*4+3];for(int i=16;i<80;i++)w[i]=std::rotl(w[i-3]^w[i-8]^w[i-14]^w[i-16],1);uint32_t a=h[0],v=h[1],c=h[2],d=h[3],e=h[4];for(int i=0;i<80;i++){uint32_t f=i<20?(v&c)|(~v&d):i<40?v^c^d:i<60?(v&c)|(v&d)|(c&d):v^c^d;uint32_t k=i<20?0x5a827999:i<40?0x6ed9eba1:i<60?0x8f1bbcdc:0xca62c1d6;uint32_t t=std::rotl(a,5)+f+e+k+w[i];e=d;d=c;c=std::rotl(v,30);v=a;a=t;}h[0]+=a;h[1]+=v;h[2]+=c;h[3]+=d;h[4]+=e;}
 void update(const uint8_t*b,size_t n){while(n){auto used=length%64;auto take=std::min(n,size_t(64-used));memcpy(buffer.data()+used,b,take);length+=take;b+=take;n-=take;if(length%64==0)block(buffer.data());}}
 std::array<uint8_t,20>sum()const{auto c=*this;uint64_t bits=length*8;uint8_t pad[128]{};pad[0]=128;c.update(pad,((length%64)<56?56:120)-length%64);for(int i=0;i<8;i++)pad[i]=bits>>(56-8*i);c.update(pad,8);std::array<uint8_t,20>out{};for(int i=0;i<20;i++)out[i]=c.h[i/4]>>(24-8*(i%4));return out;}
};
inline thread_local SHA1 rrSHA;
inline SHA1*go_sha1_New(){rrSHA=SHA1{};return &rrSHA;}
inline void hash_Write(SHA1*h,Slice<uint8_t>b){h->update(b.p,b.n);}
inline Slice<uint8_t>hash_Sum(SHA1*h,Slice<uint8_t>s){auto b=h->sum();return append(s,sub(b,0,20));}
inline auto go_sha1_Sum(Slice<uint8_t>b){SHA1 h;h.update(b.p,b.n);return h.sum();}
