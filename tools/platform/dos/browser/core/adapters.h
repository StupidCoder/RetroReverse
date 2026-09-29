#pragma once
uint8_t rrRead(dos_Machine*m,uint32_t a){return dos_Machine_Read(m,a);}uint8_t rrRead(dos_PM*m,uint32_t a){return dos_PM_Read(m,a);}
inline uint32_t rrDOSCopyByte(uint32_t address,uint8_t value){for(int i=0;i<rrdos::copySize;i++)if(rrdos::copyDestinations[i]==address&&uint8_t(rrdos::copyValue>>(i*8))==value)return rrdos::copySources[i];return rrdos::NONE;}
void rrWrite(dos_Machine*m,uint32_t a,uint8_t v){rrdos::copySource=rrDOSCopyByte(a,v);dos_Machine_Write(m,a,v);rrdos::copySource=rrdos::NONE;}
void rrWrite(dos_PM*m,uint32_t a,uint8_t v){rrdos::copySource=rrDOSCopyByte(a,v);dos_PM_Write(m,a,v);rrdos::copySource=rrdos::NONE;}
void rrDOSMovs(x86_CPU*c,uint32_t dstBase,uint32_t dst,int64_t n,uint32_t srcBase,uint32_t src){
 if(rrdos::collecting)rrdos::copyReadBefore=rrdos::raw.size();
 auto value=x86_CPU_memRead(c,srcBase,src,n);
 if(rrdos::collecting){rrdos::copyValue=value;rrdos::copySize=n;for(int i=0;i<n;i++){rrdos::copySources[i]=x86_CPU_linear(c,srcBase,src+i);rrdos::copyDestinations[i]=x86_CPU_linear(c,dstBase,dst+i);auto a=rrdos::copySources[i];if(a>=0xa0000&&a<0xb0000)rrdos::copySources[i]=rrdos::NONE;}}
 x86_CPU_memWrite(c,dstBase,dst,n,value);rrdos::copySize=0;
}
