#include "../../platform/n64/browser/core/host.h"
int main(int argc,char**argv){
 if(argc<2)return 2;
 std::ifstream f(argv[1],std::ios::binary);std::vector<uint8_t>d{std::istreambuf_iterator<char>(f),{}};
 auto rom=Slice<uint8_t>::make(d.size());std::copy(d.begin(),d.end(),rom.p);auto a=boot(rom),b=boot(rom);
 // Normalize all three byte orders to the same initial machine state.
 for(int order:{2,4}){
  auto changed=Slice<uint8_t>::make(rom.n);gcopy(changed,rom);
  for(int64_t i=0;i<changed.n;i+=order)std::reverse(changed.p+i,changed.p+i+order);
  auto c=boot(changed);if(proof(c,0)!=proof(a,0))return 3;destroy(c);
 }
 auto bad=Slice<uint8_t>::make(rom.n);gcopy(bad,rom);bad[0x80]^=1;
 bool rejected=false;try{auto c=boot(bad);destroy(c);}catch(const std::exception&e){rejected=std::string(e.what()).find("Unsupported IPL3")!=std::string::npos;}
 if(!rejected)return 4;
 if(argc>2){destroy(a);destroy(b);return 0;}
 uint64_t steps=0;for(uint64_t target:{1000000ULL,10000000ULL,50000000ULL,150000000ULL,300000000ULL}){
  for(auto [m,budget]:std::vector<std::pair<n64_Machine*,uint64_t>>{{a,10000},{b,17777}}){uint64_t n=steps;while(n<target){auto r=n64_Machine_Run(m,std::min(budget,target-n));if(m->CPU->Halted)throw std::runtime_error(m->CPU->HaltReason);n+=r.Steps;}}
  steps=target;if(proof(a,steps)!=proof(b,steps)){std::cerr<<"Mismatch at "<<steps<<std::endl;return 1;}std::cout<<proof(a,steps)<<std::endl;
 }
 destroy(a);destroy(b);
}
