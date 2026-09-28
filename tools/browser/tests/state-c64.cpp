#include "../../platform/c64/browser/core/core.h"
#include <fstream>
#include <vector>
#include <iostream>
#include <cstring>
extern "C" {uint32_t rr_state_save();uint8_t*rr_state_data();uint8_t*rr_state_input(uint32_t);int rr_state_load(uint32_t);}
std::vector<uint8_t> read(const char*p){std::ifstream f(p,std::ios::binary);if(!f)throw std::runtime_error("Missing test input");return {std::istreambuf_iterator<char>(f),{}};}
std::vector<uint8_t> save(){auto n=rr_state_save();if(!n)throw 1;auto p=rr_state_data();return {p,p+n};}
int main(int argc,char**argv){if(argc!=5)return 2;size_t at=0;for(int i=1;i<4;i++){auto b=read(argv[i]);memcpy(rr_input()+at,b.data(),b.size());at+=b.size();}if(!rr_init(8192,8192,4096))throw 2;auto tape=read(argv[4]);memcpy(rr_input(),tape.data(),tape.size());if(!rr_tape(tape.size()))throw 3;
 for(int point=0;point<4;point++){for(int i=0;i<100;i++)rr_run(10000,0,0);if(point==1){rr_key('L',1);rr_play(1);}auto a=save();if(auto p=std::getenv("NATIVE_STATE")){std::ofstream out(p,std::ios::binary);out.write((char*)a.data(),a.size());}for(int i=0;i<53;i++)rr_run(17777,0,0);auto b=save();memcpy(rr_state_input(a.size()),a.data(),a.size());if(!rr_state_load(a.size()))throw 4;for(int i=0;i<53;i++)rr_run(17777,0,0);auto c=save();if(b!=c)throw std::runtime_error("C64 continuation mismatch");std::cout<<point<<" "<<a.size()<<" bytes OK\n";}
}
