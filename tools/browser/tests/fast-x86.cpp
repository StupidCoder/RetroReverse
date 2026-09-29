#include "../../platform/dos/browser/core/api.cpp"
#include <cassert>
#include <iostream>
int main(){dos_Machine m{};m.Mem=Slice<uint8_t>::make(1<<20);dos_vgaState v{};dos_ioState io{};m.vga=&v;m.io=&io;auto*c=x86_NewCPU(&m);m.CPU=c;uint32_t r=17;auto rand=[&](){r^=r<<13;r^=r>>17;r^=r<<5;return r;};for(auto&b:m.Mem)b=rand();
 for(int i=0;i<20000;i++){auto base=(rand()&0xffff)<<4,off=rand()&0xffff;for(auto n:{1,2,4})assert(x86_CPU_memRead(c,base,off,n)==x86_CPU_memRead_reference(c,base,off,n));}
 for(auto base:{0u,0xa0000u,0xffff0u})for(auto off:{0u,0xfffdu,0xfffeu,0xffffu})for(auto n:{1,2,4}){assert(x86_CPU_memRead(c,base,off,n)==x86_CPU_memRead_reference(c,base,off,n));auto value=rand();x86_CPU_memWrite(c,base,off,n,value);assert(x86_CPU_memRead(c,base,off,n)==x86_CPU_memRead_reference(c,base,off,n));}
 for(auto ip:{0u,0xfffdu,0xfffeu,0xffffu}){c->Seg[1]=0xffff;c->IP=ip;auto got=x86_CPU_fetch32(c),end=c->IP;c->IP=ip;assert(got==x86_CPU_fetch32_reference(c)&&end==c->IP);}
 dos_PM pm{};pm.Mem=Slice<uint8_t>::make(65*1024*1024);auto*p=x86_NewCPU(&pm);pm.CPU=p;p->Mode=1;int writes=0;pm.wWLo=0;pm.wWHi=16;pm.onW=[&](uint32_t,uint32_t,uint32_t){writes++;};x86_CPU_memWrite(p,0,1,4,0x12345678);assert(writes==4);assert(x86_CPU_memRead(p,0,1,4)==0x12345678);assert(!rrSpan(&pm,pm.Mem.n-2,4,false));
 std::cout<<"Shared x86 wide reads/fetches: real-mode wrapping, unaligned RAM, VGA and watch hooks pass\n";
}
