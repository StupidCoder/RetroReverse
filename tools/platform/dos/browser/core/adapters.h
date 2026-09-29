#pragma once
uint8_t rrRead(dos_Machine*m,uint32_t a){return dos_Machine_Read(m,a);}uint8_t rrRead(dos_PM*m,uint32_t a){return dos_PM_Read(m,a);}
void rrWrite(dos_Machine*m,uint32_t a,uint8_t v){dos_Machine_Write(m,a,v);}void rrWrite(dos_PM*m,uint32_t a,uint8_t v){dos_PM_Write(m,a,v);}
