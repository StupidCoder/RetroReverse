#include "../../platform/threedo/browser/core/host.h"
#include "../../platform/threedo/browser/core/state.h"
int main(int argc,char**argv){if(argc!=2)return 2;nativeDisc(argv[1]);auto m=boot();
 for(uint64_t target:{0ull,1ull,60ull,300ull,900ull}){
  while(m->frame<target)runSlice(m,10000);
  rrstate::Archive a;a.header(4,1);a(m,totalSteps,runContext);if(auto p=std::getenv("NATIVE_STATE")){std::ofstream out(p,std::ios::binary);out.write((char*)a.bytes.data(),a.bytes.size());}rrstate::Archive r(a.bytes.data(),a.bytes.size());threedo_Machine*c=nullptr;RunContext context;r.header(4,1);uint64_t clock=0;r(c,clock,context);r.finish();rebindState(c);
  for(int i=0;i<37;i++){threedo_Machine_RunSlice(m,17777,runContext);threedo_Machine_RunSlice(c,17777,context);}
  rrstate::Archive x,y;x(m,runContext);y(c,context);if(x.bytes!=y.bytes){std::cerr<<proof(m)<<"\n"<<proof(c)<<"\n";throw std::runtime_error("3DO full state mismatch");}std::cout<<target<<" "<<a.bytes.size()<<" bytes OK"<<std::endl;
 }
}
