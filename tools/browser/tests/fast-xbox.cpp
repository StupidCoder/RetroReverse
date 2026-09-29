#include "../../platform/xbox/browser/core/host.h"
#include <cassert>
#include <iostream>
int main(){uint32_t r=12345;auto rand=[&](){r^=r<<13;r^=r>>17;r^=r<<5;return r;};for(int n=0;n<10000;n++)for(uint32_t op=0;op<8;op++){std::array<float,4>a{};for(auto&v:a)v=std::bit_cast<float>(rand());auto b=a;xbox_combMap(&a,op);xbox_combMap_reference(&b,op);for(int i=0;i<4;i++)assert((std::isnan(a[i])&&std::isnan(b[i]))||std::bit_cast<uint32_t>(a[i])==std::bit_cast<uint32_t>(b[i]));a=b;auto c=b;xbox_combOp(&a,op);xbox_combOp_reference(&c,op);for(int i=0;i<4;i++)assert((std::isnan(a[i])&&std::isnan(c[i]))||std::bit_cast<uint32_t>(a[i])==std::bit_cast<uint32_t>(c[i]));}
 std::cout<<"Xbox SIMD combiner maps/ops match scalar reference for random bit patterns, NaNs and infinities\n";
}
