#pragma once
#include "generated.cpp"
#include "disc.h"
#include "../../../../browser/core/profile.h"
inline uint64_t rrSteps=0,rrFrames=0,rrUntilVBlank=1000000;
inline psp_Result psp_Machine_Run(psp_Machine*m,uint64_t budget){
 uint64_t steps=0;rrprof::Scope measured(0,"Allegrex / VFPU and scheduler");
 for(;steps<budget;){
  if(m->StopRequested){m->StopRequested=false;break;}
  if(!--rrUntilVBlank){rrUntilVBlank=1000000;psp_Machine_deliverVBlank(m);}
  if(m->CPU->Halted||m->Halted)return {steps,m->CPU->PC,m->CPU->Halted?m->CPU->HaltReason:m->HaltReason};
  if(m->CPU->PC==psp_threadExitAddr){psp_Machine_onThreadExit(m);continue;}
  if(!m->CPU->PC)return {steps,0,"returned to null address"};
  if(!psp_Machine_mapped(m,psp_phys(m->CPU->PC)))return {steps,m->CPU->PC,"PC left mapped memory (HLE wall)"};
  if(m->OnStep)m->OnStep(m,m->CPU->PC);
  allegrex_CPU_Step(m->CPU);steps++;rrSteps++;
 }
 return {steps,m->CPU->PC,"budget reached"};
}
inline void rrBind(psp_Machine*m){m->OnDisplay=[](psp_Machine*m){rrFrames++;m->GeLists.n=0;m->StopRequested=true;};}
inline uint32_t rrHash(const uint8_t*p,size_t n){uint32_t h=2166136261;for(size_t i=0;i<n;i++)h=(h^p[i])*16777619;return h;}
inline std::vector<uint8_t>rrFrameReference(psp_Machine*m){std::vector<uint8_t>out(480*272*4);for(uint32_t y=0;y<272;y++)for(uint32_t x=0;x<480;x++){auto c=psp_Machine_readPixelFmt(m,m->fbAddr?m->fbAddr:0x04000000,m->fbWidth?m->fbWidth:480,m->fbFormat,x,y);auto i=(y*480+x)*4;out[i]=c.R;out[i+1]=c.G;out[i+2]=c.B;out[i+3]=255;}return out;}
template<unsigned F>inline void rrScanout(const uint8_t*src,uint32_t stride,uint8_t*out){
 for(unsigned y=0;y<272;y++)for(unsigned x=0;x<480;x++){uint32_t c=rrLoadColor<F>(src+(uint64_t(y)*stride+x)*(F==3?4:2))|0xff000000;std::memcpy(out+(y*480+x)*4,&c,4);}
}
inline std::vector<uint8_t>rrFrame(psp_Machine*m){
 uint32_t base=m->fbAddr?m->fbAddr:0x04000000,stride=m->fbWidth?m->fbWidth:480;uint64_t bytes=(uint64_t(271)*stride+480)*(m->fbFormat==3?4:2);
 const uint8_t*p=!m->OnRead&&bytes<=UINT32_MAX?rrMemory(m,base,bytes):nullptr;if(!p)return rrFrameReference(m);
 std::vector<uint8_t>out(480*272*4);switch(m->fbFormat){case 1:rrScanout<1>(p,stride,out.data());break;case 2:rrScanout<2>(p,stride,out.data());break;case 3:rrScanout<3>(p,stride,out.data());break;default:rrScanout<0>(p,stride,out.data());}return out;
}
inline psp_Machine*rrBoot(BlockSource*disc){auto[v,e]=psp_OpenVolume(disc);if(e)throw std::runtime_error(e.text);auto[raw,re]=psp_Volume_ReadFile(v,"PSP_GAME/SYSDIR/EBOOT.BIN");if(re)throw std::runtime_error(re.text);auto[mod,me]=psp_LoadModuleImage(raw);if(me)throw std::runtime_error(me.text);auto*m=psp_NewMachine();psp_Machine_SetVolume(m,v);auto le=psp_Machine_LoadModule(m,mod);if(le)throw std::runtime_error(le.text);rrSteps=rrFrames=0;rrUntilVBlank=1000000;rrBind(m);return m;}
