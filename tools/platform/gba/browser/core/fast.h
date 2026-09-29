#pragma once
// These direct reads preserve EEPROM/I/O behavior and observation hooks by
// falling back to the translated reference outside ordinary backed memory.
uint16_t gbamachine_bus_Read16(gbamachine_bus*b,uint32_t a){
 auto*m=b->m;if(m->OnRead||(a&0xffffff)==0xffffff)return gbamachine_bus_Read16_reference(b,a);
 const uint8_t*p=nullptr;uint32_t off=0,n=0;
 switch(a>>24){case 2:p=m->ewram.p;off=a&0x3ffff;n=m->ewram.n;break;case 3:p=m->iwram.p;off=a&0x7fff;n=m->iwram.n;break;case 5:p=m->pal.p;off=a&1023;n=m->pal.n;break;case 6:p=m->vram.p;off=a&0x1ffff;if(off>=0x18000)off-=0x8000;n=m->vram.n;break;case 7:p=m->oam.p;off=a&1023;n=m->oam.n;break;case 8:case 9:case 10:case 11:case 12:p=m->rom.p;off=a&0x1ffffff;n=m->rom.n;break;}
 if(p&&uint64_t(off)+2<=n){uint16_t v;std::memcpy(&v,p+off,2);return v;}
 return gbamachine_bus_Read16_reference(b,a);
}
