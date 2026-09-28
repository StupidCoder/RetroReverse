#pragma once
// Portable SHA-1 for PSP NIDs and PRX header checks. Not an authentication API.
inline std::array<uint8_t,20> go_sha1_Sum(Slice<uint8_t>s){
 uint64_t bytes=s.n;std::vector<uint8_t>b(s.begin(),s.end());b.push_back(0x80);while(b.size()%64!=56)b.push_back(0);for(int i=7;i>=0;i--)b.push_back((bytes*8)>>(i*8));
 uint32_t h[5]={0x67452301,0xefcdab89,0x98badcfe,0x10325476,0xc3d2e1f0};
 for(size_t off=0;off<b.size();off+=64){uint32_t w[80];for(int i=0;i<16;i++){w[i]=0;for(int j=0;j<4;j++)w[i]=(w[i]<<8)|b[off+i*4+j];}for(int i=16;i<80;i++)w[i]=std::rotl(w[i-3]^w[i-8]^w[i-14]^w[i-16],1);
 uint32_t a=h[0],c=h[2],d=h[3],e=h[4],v=h[1];for(int i=0;i<80;i++){uint32_t f,k;if(i<20){f=(v&c)|(~v&d);k=0x5a827999;}else if(i<40){f=v^c^d;k=0x6ed9eba1;}else if(i<60){f=(v&c)|(v&d)|(c&d);k=0x8f1bbcdc;}else{f=v^c^d;k=0xca62c1d6;}uint32_t t=std::rotl(a,5)+f+e+k+w[i];e=d;d=c;c=std::rotl(v,30);v=a;a=t;}h[0]+=a;h[1]+=v;h[2]+=c;h[3]+=d;h[4]+=e;}
 std::array<uint8_t,20>o;for(int i=0;i<20;i++)o[i]=h[i/4]>>(24-8*(i&3));return o;
}
struct SHA1 {Slice<uint8_t> bytes;};
inline SHA1*go_sha1_New(){return arenaNew(SHA1{});}
inline void hash_Write(SHA1*h,Slice<uint8_t>s){h->bytes=append(h->bytes,s);}
inline Slice<uint8_t>hash_Sum(SHA1*h,Slice<uint8_t>s){auto b=go_sha1_Sum(h->bytes);return append(s,sub(b,0,20));}
// AES-128 inverse cipher, using computed S-box tables and byte-oriented rounds.
// KIRK CBC uses a zero IV and leaves incomplete trailing blocks unchanged.
namespace rrcrypto {
inline uint8_t mul(uint8_t a,uint8_t b){uint8_t r=0;for(int i=0;i<8;i++){if(b&1)r^=a;bool high=a&128;a<<=1;if(high)a^=0x1b;b>>=1;}return r;}
struct Tables{std::array<uint8_t,256>s{},inv{};Tables(){for(int i=0;i<256;i++){uint8_t x=1;if(i==0)x=0;else for(int j=0;j<254;j++)x=mul(x,i);uint8_t v=x^std::rotl(x,1)^std::rotl(x,2)^std::rotl(x,3)^std::rotl(x,4)^0x63;s[i]=v;inv[v]=i;}}};
inline const Tables tables;
inline std::tuple<Slice<uint8_t>,Error>cbc(Slice<uint8_t>key,Slice<uint8_t>src){if(key.n!=16)return {{},{"AES-128 key must be 16 bytes"}};auto out=Slice<uint8_t>::make(src.n);gcopy(out,src);uint8_t rk[176];std::memcpy(rk,key.p,16);uint8_t rc=1;
 for(int i=16;i<176;i+=4){uint8_t t[4];std::memcpy(t,rk+i-4,4);if(i%16==0){auto v=t[0];t[0]=tables.s[t[1]]^rc;t[1]=tables.s[t[2]];t[2]=tables.s[t[3]];t[3]=tables.s[v];rc=mul(rc,2);}for(int j=0;j<4;j++)rk[i+j]=rk[i+j-16]^t[j];}
 uint8_t prev[16]={};for(int64_t off=0;off+16<=src.n;off+=16){uint8_t s[16];for(int i=0;i<16;i++)s[i]=src[off+i]^rk[160+i];for(int round=9;round>=0;round--){uint8_t t[16];for(int c=0;c<4;c++)for(int r=0;r<4;r++)t[c*4+r]=tables.inv[s[((c-r+4)&3)*4+r]];for(int i=0;i<16;i++)s[i]=t[i]^rk[round*16+i];if(round)for(int c=0;c<4;c++){auto p=s+c*4;uint8_t a=p[0],b=p[1],d=p[2],e=p[3];p[0]=mul(a,14)^mul(b,11)^mul(d,13)^mul(e,9);p[1]=mul(a,9)^mul(b,14)^mul(d,11)^mul(e,13);p[2]=mul(a,13)^mul(b,9)^mul(d,14)^mul(e,11);p[3]=mul(a,11)^mul(b,13)^mul(d,9)^mul(e,14);}}
 for(int i=0;i<16;i++){out[off+i]=s[i]^prev[i];prev[i]=src[off+i];}}
 return {out,{}};
}
}
