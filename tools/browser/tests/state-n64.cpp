#include "../../platform/n64/browser/core/host.h"
#include "../../platform/n64/browser/core/state.h"
#include <fstream>
int main(int argc,char**argv){if(argc!=2)return 2;std::ifstream f(argv[1],std::ios::binary);std::vector<uint8_t>b{std::istreambuf_iterator<char>(f),{}};auto rom=Slice<uint8_t>::make(b.size());memcpy(rom.p,b.data(),b.size());auto m=boot(rom);
 for(uint64_t target:{10000ull,10000000ull,150000000ull,300000000ull}){
  while(m->CPU->Steps<target)n64_Machine_Run(m,10000);
  rrstate::Archive a;a.header(3,1);uint64_t clock=0;a(m,clock);if(auto p=std::getenv("NATIVE_STATE")){std::ofstream out(p,std::ios::binary);out.write((char*)a.bytes.data(),a.bytes.size());}rrstate::Archive r(a.bytes.data(),a.bytes.size());n64_Machine*c=nullptr;r.header(3,1);r(c,clock);r.finish();rebindState(c,m->ROM);
  for(int i=0;i<37;i++){n64_Machine_Run(m,17777);n64_Machine_Run(c,17777);}if(proof(m,0)!=proof(c,0))throw std::runtime_error("N64 continuation mismatch");rrstate::Archive x,y;x(m);y(c);if(x.bytes!=y.bytes)throw std::runtime_error("N64 full state mismatch");std::cout<<target<<" "<<a.bytes.size()<<" bytes OK\n";
 }
}
