#include "../../platform/psx/browser/core/proof.h"
#include "../../platform/psx/browser/core/state.h"
#include <fstream>
#include <iostream>
int main(int argc,char**argv){if(argc!=2)return 2;std::ifstream f(argv[1],std::ios::binary);std::vector<u8>b{std::istreambuf_iterator<char>(f),{}};auto d=std::make_shared<Disc>();d->open(b.data(),b.size());auto m=std::make_unique<Machine>();m->boot(d,0);
 for(uint64_t target:{10000ull,10000000ull,400000000ull,500000000ull}){
  while(m->cpu.steps<target)m->run(std::min<uint64_t>(10000,target-m->cpu.steps));
  rrstate::Archive a;a.header(2,1);a(*m);if(auto p=std::getenv("NATIVE_STATE")){std::ofstream out(p,std::ios::binary);out.write((char*)a.bytes.data(),a.bytes.size());}rrstate::Archive r(a.bytes.data(),a.bytes.size());auto clone=std::make_unique<Machine>();r.header(2,1);r(*clone);r.finish();validateState(*clone);clone->disc=d;
  for(int i=0;i<37;i++){m->run(17777);clone->run(17777);}if(proof(*m)!=proof(*clone))throw std::runtime_error("PS1 continuation mismatch");
  rrstate::Archive x,y;x(*m);y(*clone);if(x.bytes!=y.bytes)throw std::runtime_error("PS1 full state mismatch");std::cout<<target<<" "<<a.bytes.size()<<" bytes OK\n";
 }
}
