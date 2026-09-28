#pragma once
#include "lcd.h"

inline uint8_t gameboy_Machine_Read(gameboy_Machine*m,uint16_t a){
 if(a<0x4000&&m->rom[0x147]!=0&&m->mode==1){auto bank=(m->ramBank<<5)%m->nbanks;return m->rom[bank*16384+a];}
 return gameboy_Machine_Read_Reference(m,a);
}
inline void gameboy_Machine_mbcWrite(gameboy_Machine*m,uint16_t a,uint8_t v){if(m->rom[0x147])gameboy_Machine_mbcWrite_Reference(m,a,v);}
