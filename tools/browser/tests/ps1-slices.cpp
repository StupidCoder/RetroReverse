#include "../../platform/psx/browser/core/proof.h"
#include <fstream>
#include <iostream>
int main(int argc,char**argv){
 if(argc!=2)return 2;
 std::ifstream f(argv[1],std::ios::binary);std::vector<u8> data{std::istreambuf_iterator<char>(f),{}};
 auto d=std::make_shared<Disc>();d->open(data.data(),data.size());
 auto cached=std::make_shared<Disc>();cached->openFile(data.size(),[&](size_t o,u8*p,size_t n){if(o+n>data.size())throw std::runtime_error("range");memcpy(p,data.data()+o,n);});
 auto a=std::make_unique<Machine>(),b=std::make_unique<Machine>();a->boot(d,0);b->boot(cached,0);
 for(auto [target,input]:std::vector<std::pair<u64,u16>>{{10000000,65535},{380000000,0xfff7},{380380000,65535},{386000000,0xbfff},{386380000,65535},{500000000,65535}}){
  for(auto [m,budget]:std::vector<std::pair<Machine*,u32>>{{a.get(),10000},{b.get(),17777}})while(m->cpu.steps<target)m->run(std::min<u64>(budget,target-m->cpu.steps));
  if(proof(*a)!=proof(*b)){std::cerr<<"Mismatch at "<<target<<'\n';return 1;}
  std::cout<<proof(*a)<<'\n';a->buttons=b->buttons=input;
 }
}
