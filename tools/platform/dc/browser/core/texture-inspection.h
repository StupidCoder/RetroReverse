#pragma once
#include "../../../../browser/core/replay.h"

// Capture-only TA state. The command hook precedes execution, so include the
// pending header/vertex. These snapshots never touch the running renderer.
namespace rrdc::textures {
struct State {
 uint32_t pcw=0,isp=0,tsp=0,tcw=0,resource=0;
 int type=-1;bool known=false,sprite=false,draw=false;
 std::array<std::array<float,2>,4> uv{};unsigned count=0;
};
inline State last;
inline std::vector<State> states;
inline std::array<std::array<float,2>,3> strip{};
inline unsigned stripCount=0;
inline void reset(){last={};states.clear();stripCount=0;}
inline void record(uint32_t event,Slice<uint8_t> p,uint32_t resource){
 if(!event)return;
 State s=last;s.resource=resource;s.draw=false;s.count=0;s.type=-1;
 auto word=[&](int at){return le_Uint32(rrBorrow(p,at,at+4));};
 auto uv=[&](int at){auto [u,v]=dc_unpackUV16(word(at));return std::array<float,2>{u,v};};
 if(p.n<4){s={};stripCount=0;}else{
  uint32_t pcw=word(0);s.type=pcw>>29;
  if((s.type==4||s.type==5)&&p.n>=16){
   s.pcw=pcw;s.isp=word(4);s.tsp=word(8);s.tcw=word(12);s.known=true;s.sprite=s.type==5;stripCount=0;
  }else if(s.type==7&&s.known&&(s.pcw&8)){
   if(s.sprite&&p.n>=64){
    s.uv[0]=uv(52);s.uv[1]=uv(56);s.uv[2]=uv(60);
    for(int j=0;j<2;j++)s.uv[3][j]=s.uv[0][j]+s.uv[2][j]-s.uv[1][j];
    s.count=4;s.draw=true;
   }else if(!s.sprite&&p.n>=24){
    auto v=(s.pcw&1)?uv(16):std::array<float,2>{std::bit_cast<float>(word(16)),std::bit_cast<float>(word(20))};
    if(stripCount==3){strip[0]=strip[1];strip[1]=strip[2];strip[2]=v;}else strip[stripCount++]=v;
    s.count=stripCount;std::copy_n(strip.begin(),stripCount,s.uv.begin());s.draw=stripCount==3;
    if(pcw&(1u<<28))stripCount=0;
   }
  }
 }
 last=s;if(states.size()<=event)states.resize(event+1);states[event]=s;
}
inline State selected(){auto&r=rrreplay::replay;if(r.cursor&&r.cursor<=r.steps.size()){auto e=r.steps[r.cursor-1].event;if(e&&e<states.size())return states[e];}return {};}
inline std::string info(){
 auto&t=rrcapture::trace;auto&r=rrreplay::replay;
 if(!t.valid||r.memory.size()!=8388608)return "{\"error\":\"No captured texture state\"}";
 auto s=selected();std::ostringstream o;
 auto textured=[&](size_t i){auto e=r.steps[i].event;return e&&e<states.size()&&states[e].draw&&(states[e].pcw&8);};
 int previous=-1,next=-1;for(int i=int(r.cursor)-2;i>=0;i--)if(textured(i)){previous=i+1;break;}
 for(size_t i=r.cursor;i<r.steps.size();i++)if(textured(i)){next=i+1;break;}
 o<<"{\"previousTextured\":"<<previous<<",\"nextTextured\":"<<next<<',';
 o<<"\"cursor\":"<<r.cursor<<",\"complete\":"<<(!t.overflow?"true":"false")<<",\"known\":"<<(s.known?"true":"false")
  <<",\"type\":"<<s.type<<",\"sprite\":"<<(s.sprite?"true":"false")<<",\"draw\":"<<(s.draw?"true":"false")
  <<",\"pcw\":"<<s.pcw<<",\"isp\":"<<s.isp<<",\"tsp\":"<<s.tsp<<",\"tcw\":"<<s.tcw<<",\"uv\":[";
 for(unsigned i=0;i<s.count;i++){if(i)o<<',';o<<'[';for(int j=0;j<2;j++){if(j)o<<',';if(std::isfinite(s.uv[i][j]))o<<s.uv[i][j];else o<<"null";}o<<']';}o<<']';
 // Palette and global controls come from this event's historical register file.
 const std::vector<uint8_t>*regs=s.resource&&s.resource<=t.resources.size()?&t.resources[s.resource-1]:nullptr;
 auto reg=[&](unsigned a)->uint32_t{if(!regs||a+4>regs->size())return 0;return uint32_t((*regs)[a])|uint32_t((*regs)[a+1])<<8|uint32_t((*regs)[a+2])<<16|uint32_t((*regs)[a+3])<<24;};
 o<<",\"registersKnown\":"<<(regs&&regs->size()==8192?"true":"false")<<",\"paletteFormat\":"<<(reg(0x108)&3)<<",\"stride\":"<<((reg(0xe4)&31)*32)<<",\"target\":"<<(reg(0x60)&0xffffff)<<",\"palette\":[";
 auto fmt=s.tcw>>27&7;unsigned count=fmt==5?16:fmt==6?256:0,base=fmt==5?((s.tcw>>21)&63)*16:((s.tcw>>25)&3)*256;
 for(unsigned i=0;i<count;i++){if(i)o<<',';o<<reg(0x1000+(base+i)*4);}o<<"]}";return o.str();
}
}
