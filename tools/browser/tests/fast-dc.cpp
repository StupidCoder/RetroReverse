#include "../../platform/dc/browser/core/api.cpp"
#include <cassert>
#include <iostream>
int main(){auto*m=dc_NewMachine(nullptr);uint32_t random=12345;auto next=[&](){random^=random<<13;random^=random>>17;random^=random<<5;return random;};for(auto*s:{&m->RAM,&m->VRAM,&m->AICARAM})for(auto&v:*s)v=next();for(int i=0;i<20000;i++){auto x=next(),y=next();assert(dc_twiddle(x,y)==dc_twiddle_reference(x,y));uint32_t a=std::array<uint32_t,4>{0x0c000000,0x04000000,0x05000000,0x00800000}[i%4]+(next()&0x1ffffc);assert(dc_Machine_Fetch16(m,a)==dc_Machine_Fetch16_reference(m,a));assert(dc_Machine_read16i(m,a)==dc_Machine_read16i_reference(m,a));assert(dc_Machine_read32i(m,a)==dc_Machine_read32i_reference(m,a));assert(dc_armBus_Read32({m},a)==dc_armBus_Read32_reference({m},a));}
 // Register changes and wrap/interrupt boundaries must take effect immediately.
 auto*n=dc_NewMachine(nullptr);for(int i=0;i<100000;i++){if(i%13==0){uint32_t load=next();m->PVRRegs[0xf8/4]=n->PVRRegs[0xf8/4]=load;m->PVRRegs[0xcc/4]=n->PVRRegs[0xcc/4]=next();m->instrInField=n->instrInField=next()%20000;}dc_Machine_tickField(m);dc_Machine_tickField_reference(n);assert(m->Fields==n->Fields&&m->CurLine==n->CurLine&&m->instrInField==n->instrInField&&m->FieldNum==n->FieldNum&&m->Holly.ISTNRM==n->Holly.ISTNRM);}
 std::cout<<"Dreamcast direct reads, texture addressing and field timing match reference\n";
}
