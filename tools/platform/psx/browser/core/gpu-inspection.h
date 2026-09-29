#pragma once
#include "gpu.h"
#include "../../../../browser/core/replay.h"

// Capture-only state: no changes to guest execution or portable machine states.
// Unsupported controls are unknown until explicitly observed in this capture.
namespace rrps1 {
struct GPUState {
 int pageX=0,pageY=0,depth=0,clut=-1;
 int maskX=0,maskY=0,windowX=0,windowY=0;
 int left=0,top=0,right=1024,bottom=512,offsetX=0,offsetY=0;
 int blend=-1,dither=-1,forceMask=-1,checkMask=-1;
 unsigned opcode=0;bool textured=false,gouraud=false,raw=false,semi=false;
 std::array<std::array<int,2>,4> uv{};unsigned vertices=0;
};
inline GPUState initial,last;
inline std::vector<GPUState> states;
inline GPUState read(const GPU&g){
 GPUState s;s.pageX=g.texPageX;s.pageY=g.texPageY;s.depth=g.texDepth;
 s.maskX=g.texWinMX;s.maskY=g.texWinMY;s.windowX=g.texWinOX;s.windowY=g.texWinOY;
 s.left=g.drawL;s.top=g.drawT;s.right=g.drawR;s.bottom=g.drawB;s.offsetX=g.offX;s.offsetY=g.offY;return s;
}
inline void begin(const GPU&g){initial=last=read(g);states.clear();states.push_back(initial);}
inline void page(GPUState&s,u32 word){s.pageX=(word&15)*64;s.pageY=((word>>4)&1)*256;s.depth=(word>>7)&3;s.blend=(word>>5)&3;}
inline GPUState command(const GPU&g,const std::vector<u32>&words){
 GPUState s=read(g);s.clut=last.clut;s.blend=last.blend;s.dither=last.dither;s.forceMask=last.forceMask;s.checkMask=last.checkMask;
 if(words.empty())return s;auto w=words[0];unsigned op=w>>24;s.opcode=op;
 bool polygon=op>=0x20&&op<0x40,rect=op>=0x60&&op<0x80;
 s.textured=(polygon||rect)&&(op&4);s.gouraud=polygon&&(op&16);s.raw=s.textured&&(op&1);s.semi=(polygon||rect)&&(op&2);
 if(polygon&&s.textured){
  unsigned count=op&8?4:3,at=s.gouraud?0:1;
  for(unsigned i=0;i<count;i++){
   if(s.gouraud)at++;at++;if(at>=words.size())break;u32 uv=words[at++];
   s.uv[s.vertices++]={int(uv&255),int((uv>>8)&255)};
   if(i==0)s.clut=(uv>>16)&0x7fff;if(i==1)page(s,uv>>16);
  }
 }else if(rect&&s.textured&&words.size()>=3){
  auto uv=words[2];s.clut=(uv>>16)&0x7fff;int u=uv&255,v=(uv>>8)&255,width=1,height=1;
  if((op&24)==16)width=height=8;else if((op&24)==24)width=height=16;
  else if((op&24)==0&&words.size()>=4){width=words[3]&65535;height=words[3]>>16;}
  s.uv={{{u,v},{u+width,v},{u+width,v+height},{u,v+height}}};s.vertices=4;
 }
 // The GPU command hook runs before execution. Project only the pending packet's
 // settings into the snapshot so this command gets its own page, not its predecessor's.
 switch(op){
 case 0xe1:page(s,w);s.dither=(w>>9)&1;break;
 case 0xe2:s.maskX=w&31;s.maskY=(w>>5)&31;s.windowX=(w>>10)&31;s.windowY=(w>>15)&31;break;
 case 0xe3:s.left=w&1023;s.top=(w>>10)&1023;break;
 case 0xe4:s.right=(w&1023)+1;s.bottom=((w>>10)&1023)+1;break;
 case 0xe5:s.offsetX=GPU::sext11(w&2047);s.offsetY=GPU::sext11((w>>11)&2047);break;
 case 0xe6:s.forceMask=w&1;s.checkMask=(w>>1)&1;break;
 }
 return last=s;
}
inline void record(uint32_t event,const GPUState&s){if(!event)return;if(states.size()<=event)states.resize(event+1);states[event]=s;}
inline const GPUState&selected(){
 const auto&r=rrreplay::replay;
 if(r.cursor&&r.cursor<=r.steps.size()){auto e=r.steps[r.cursor-1].event;if(e<states.size())return states[e];}
 return initial;
}
inline std::string info(){
 if(!rrcapture::trace.valid||states.empty())return "{\"error\":\"No captured GPU state\"}";
 const auto&s=selected();std::ostringstream o;
 o<<"{\"cursor\":"<<rrreplay::replay.cursor<<",\"complete\":"<<(!rrcapture::trace.overflow?"true":"false")
  <<",\"pageX\":"<<s.pageX<<",\"pageY\":"<<s.pageY<<",\"depth\":"<<s.depth<<",\"clut\":"<<s.clut
  <<",\"window\":["<<s.maskX<<','<<s.maskY<<','<<s.windowX<<','<<s.windowY<<']'
  <<",\"area\":["<<s.left<<','<<s.top<<','<<s.right<<','<<s.bottom<<"],\"offset\":["<<s.offsetX<<','<<s.offsetY<<']'
  <<",\"blend\":"<<s.blend<<",\"dither\":"<<s.dither<<",\"forceMask\":"<<s.forceMask<<",\"checkMask\":"<<s.checkMask
  <<",\"opcode\":"<<s.opcode<<",\"textured\":"<<(s.textured?"true":"false")<<",\"gouraud\":"<<(s.gouraud?"true":"false")
  <<",\"raw\":"<<(s.raw?"true":"false")<<",\"semi\":"<<(s.semi?"true":"false")<<",\"uv\":[";
 for(unsigned i=0;i<s.vertices;i++){if(i)o<<',';o<<'['<<s.uv[i][0]<<','<<s.uv[i][1]<<']';}
 return o.str()+"]}";
}
}
