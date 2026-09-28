#include "../../platform/c64/browser/core/core.cpp"
#include <fstream>
#include <iostream>
std::vector<uint8_t> readFile(const char*p){std::ifstream f(p,std::ios::binary);if(!f)throw std::runtime_error("Missing input");return {std::istreambuf_iterator<char>(f),{}};}
void runCycles(int n){while(n){int slice=std::min(n,1000000);if(rr_run(slice,0,0)<0)throw std::runtime_error(rr_error());n-=slice;}}
void type(const char*s){for(;*s;s++){rr_key(*s,1);runCycles(100000);rr_key(*s,0);runCycles(100000);}}
int main(int argc,char**argv){if(argc!=6)return 2;size_t at=0;for(int i=1;i<4;i++){auto b=readFile(argv[i]);memcpy(rr_input()+at,b.data(),b.size());at+=b.size();}if(!rr_init(8192,8192,4096))return 3;auto b=readFile(argv[4]);memcpy(rr_input(),b.data(),b.size());if(!rr_tape(b.size()))return 4;runCycles(5000000);type("LOAD\r");rr_play(1);for(int i=0;i<250&&ctx.pulse<48236;i++)runCycles(1000000);runCycles(2000000);type("RUN\r");for(int i=0;i<250&&ctx.pulse<pulse_count;i++)runCycles(1000000);runCycles(5000000);rr_joystick(2,16);runCycles(400000);rr_joystick(2,0);runCycles(2000000);auto size=rr_state_save();std::ofstream out(argv[5],std::ios::binary);out.write((char*)rr_state_data(),size);std::cout<<rr_status()<<'\n';}
