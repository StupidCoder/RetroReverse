#include "../../platform/gba/browser/core/api.cpp"
#include <cassert>
#include <iostream>
int main(){rr_input(32*1024*1024);assert(rr_init(32*1024*1024));auto m=machine;gbamachine_bus b{m};uint32_t random=0xabc123;auto next=[&](){random^=random<<13;random^=random>>17;random^=random<<5;return random;};for(auto*s:{&m->rom,&m->ewram,&m->iwram,&m->vram,&m->pal,&m->oam})for(auto&v:*s)v=next();
 for(int i=0;i<20000;i++){uint32_t a=(2+next()%12)*0x1000000u+(next()&0xffffff);if(i%8==0)a=(a&0xff000000u)|0xffffff;uint16_t x=0,y=0;bool ex=false,ey=false;try{x=gbamachine_bus_Read16(&b,a);}catch(...){ex=true;}try{y=gbamachine_bus_Read16_reference(&b,a);}catch(...){ey=true;}assert(ex==ey);if(!ex)assert(x==y);}
 std::cout<<"GBA direct reads match reference\n";
}
