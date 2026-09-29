#pragma once
// One checked translation per operand, instead of four virtual byte accesses.
// Slow paths retain real-mode segment/20-bit wrapping, MMIO and watch hooks.
inline uint8_t*rrOperand(x86_CPU*c,uint32_t base,uint32_t off,uint32_t n,bool write){if(c->Mode==0&&((off&65535)+n>65536))return nullptr;auto a=x86_CPU_linear(c,base,off);return c->bus.span?c->bus.span(c->bus.ctx,a,n,write):nullptr;}
inline uint32_t rrLoadLE(const uint8_t*p,int n){if(n==1)return *p;if(n==2){uint16_t v;memcpy(&v,p,2);return v;}uint32_t v;memcpy(&v,p,4);return v;}
uint32_t x86_CPU_memRead(x86_CPU*c,uint32_t base,uint32_t off,int64_t n){
#ifndef RR_X86_REFERENCE
 if((n==1||n==2||n==4))if(auto*p=rrOperand(c,base,off,n,false))return rrLoadLE(p,n);
#endif
 return x86_CPU_memRead_reference(c,base,off,n);
}
void x86_CPU_memWrite(x86_CPU*c,uint32_t base,uint32_t off,int64_t n,uint32_t v){
#ifndef RR_X86_REFERENCE
 if(n==1||n==2||n==4)if(auto*p=rrOperand(c,base,off,n,true)){memcpy(p,&v,n);return;}
#endif
 x86_CPU_memWrite_reference(c,base,off,n,v);
}
uint32_t x86_CPU_fetch16(x86_CPU*c){
#ifndef RR_X86_REFERENCE
 if(auto*p=rrOperand(c,x86_CPU_segBase(c,1),c->IP,2,false)){auto v=rrLoadLE(p,2);c->IP=x86_CPU_ipMask(c,c->IP+2);return v;}
#endif
 return x86_CPU_fetch16_reference(c);
}
uint32_t x86_CPU_fetch32(x86_CPU*c){
#ifndef RR_X86_REFERENCE
 if(auto*p=rrOperand(c,x86_CPU_segBase(c,1),c->IP,4,false)){auto v=rrLoadLE(p,4);c->IP=x86_CPU_ipMask(c,c->IP+4);return v;}
#endif
 return x86_CPU_fetch32_reference(c);
}
