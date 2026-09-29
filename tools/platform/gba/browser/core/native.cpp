#include "api.cpp"
#include <fstream>
#include <iostream>
int main(int argc,char**argv){try{
 if(argc<2)throw std::runtime_error("Usage: gba-native ROM [ADDITIONAL-FRAMES] [PPM]");std::ifstream f(argv[1],std::ios::binary);std::vector<uint8_t>b((std::istreambuf_iterator<char>(f)),{});auto*p=rr_input(b.size());if(!p)throw std::runtime_error(rr_error());std::memcpy(p,b.data(),b.size());if(!rr_init(b.size()))throw std::runtime_error(rr_error());
 if(auto file=std::getenv("RR_LOAD")){std::ifstream f(file,std::ios::binary);std::vector<uint8_t>s((std::istreambuf_iterator<char>(f)),{});std::memcpy(rr_state_input(s.size()),s.data(),s.size());if(!rr_state_load(s.size()))throw std::runtime_error("State load failed");}
 auto end=machine->vid.frames+(argc>2?std::stoull(argv[2]):300);auto start=go_time_Now();
 while(machine->vid.frames<end){if(std::getenv("RR_AUTO"))rr_pad(machine->vid.frames%120<8?9:0);if(rr_run(10000)<0)throw std::runtime_error(rr_error());}
 std::cout<<rr_proof()<<"\n"<<rr_profile()<<"\n";std::cerr<<"Seconds "<<(go_time_Now()-start)/1e9<<"\n";
 if(auto file=std::getenv("RR_SAVE")){auto n=rr_state_save();if(!n)throw std::runtime_error("State save failed");std::ofstream f(file,std::ios::binary);f.write((char*)rr_state_data(),n);}
 if(argc>3){std::ofstream f(argv[3],std::ios::binary);f<<"P6\n240 160\n255\n";auto*p=rr_frame();for(unsigned i=0;i<240*160;i++)f.write((char*)p+i*4,3);}
 }catch(const std::exception&e){std::cerr<<e.what()<<"\n";return 1;}}
