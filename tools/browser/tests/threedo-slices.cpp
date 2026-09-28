#include "../../platform/threedo/browser/core/host.h"
// Compare stdout with native.cpp checkpoints. Optional frame/input script has
// the same format as native.cpp. Run with several budgets to exercise yields.
int main(int argc,char**argv){
 if(argc<4)return 2;
 nativeDisc(argv[1]);auto m=boot();auto frames=std::stoull(argv[2]),slice=std::stoull(argv[3]);
 std::unordered_map<uint64_t,uint32_t> inputs;if(argc>4){std::ifstream f(argv[4]);uint64_t n;uint32_t b;while(f>>n>>b)inputs[n]=b;}
 for(uint64_t i=0;i<frames;i++){
  if(inputs.count(i))threedo_Machine_SendPadEvent(m,inputs[i]);
  auto before=m->frame;do{runSlice(m,slice);}while(m->frame==before||m->StopRequested);
  if(i==0||i%30==29)std::cout<<proof(m)<<std::endl;
 }
}
