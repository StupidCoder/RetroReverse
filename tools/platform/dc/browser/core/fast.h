#pragma once
inline const uint8_t*rrDCBacking(dc_Machine*m,uint32_t a,uint32_t size){
 const uint8_t*p=nullptr;uint32_t off=0,n=0;
 if(a>>26==3){p=m->RAM.p;off=a&0xffffff;n=m->RAM.n;}
 else if(a>=0x05000000&&a<0x05800000){p=m->VRAM.p;off=dc_vram32to64(a-0x05000000);n=m->VRAM.n;}
 else if(a>=0x04000000&&a<0x04800000){p=m->VRAM.p;off=a-0x04000000;n=m->VRAM.n;}
 else if(a>=0x00800000&&a<0x00a00000){p=m->AICARAM.p;off=a-0x00800000;n=m->AICARAM.n;}
 return p&&uint64_t(off)+size<=n?p+off:nullptr;
}
uint16_t dc_Machine_Fetch16(dc_Machine*m,uint32_t a){if(auto*p=rrDCBacking(m,a,2)){uint16_t v;std::memcpy(&v,p,2);return v;}return dc_Machine_Fetch16_reference(m,a);}
uint16_t dc_Machine_read16i(dc_Machine*m,uint32_t a){if(auto*p=rrDCBacking(m,a,2)){uint16_t v;std::memcpy(&v,p,2);return v;}return dc_Machine_read16i_reference(m,a);}
uint32_t dc_Machine_read32i(dc_Machine*m,uint32_t a){if(auto*p=rrDCBacking(m,a,4)){uint32_t v;std::memcpy(&v,p,4);return v;}return dc_Machine_read32i_reference(m,a);}
uint32_t dc_armBus_Read32(dc_armBus b,uint32_t addr){uint32_t a=addr&0xffffff,off=a&0x1fffff;if(a<0x800000&&uint64_t(off)+4<=uint64_t(b.m->AICARAM.n)){uint32_t v;std::memcpy(&v,b.m->AICARAM.p+off,4);return v;}return dc_armBus_Read32_reference(b,addr);}
inline uint32_t rrSpread16(uint32_t v){v&=65535;v=(v|v<<8)&0x00ff00ff;v=(v|v<<4)&0x0f0f0f0f;v=(v|v<<2)&0x33333333;return (v|v<<1)&0x55555555;}
uint32_t dc_twiddle(uint32_t x,uint32_t y){return rrSpread16(y)|(rrSpread16(x)<<1);}
void dc_Machine_tickField(dc_Machine*m){
 ++m->instrInField;const uint32_t total=dc_Machine_spgTotalLines(m);
 if(total!=m->rrLastTotal){m->rrLastTotal=total;m->rrLinePeriod=3333333u/total;}
 if(m->instrInField<m->rrLinePeriod)return;
 m->instrInField=0;if(++m->CurLine>=total){m->CurLine=0;m->FieldNum^=1;}
 auto v=m->PVRRegs[0xcc/4];
 if(m->CurLine==(v&1023)){++m->Fields;dc_Machine_raiseNRM(m,dc_istVBlankIn);if(m->OnDisplay)m->OnDisplay(m->Fields);}
 else if(m->CurLine==(v>>16&1023))dc_Machine_raiseNRM(m,dc_istVBlankOut);
}
