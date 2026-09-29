#include "api.cpp"
#include <iostream>
int main(int argc,char**argv){try{if(argc<3)throw std::runtime_error("dos-native FOLDER EXE [FRAMES] [PPM]");rrfiles::mountNative(argv[1]);auto name=rrfiles::path(argv[2]);memcpy(rr_input(name.size()),name.data(),name.size());if(!rr_init(name.size(),1))throw std::runtime_error(rr_error());
if(auto file=std::getenv("RR_LOAD")){std::ifstream f(file,std::ios::binary);std::vector<uint8_t>s((std::istreambuf_iterator<char>(f)),{});memcpy(rr_state_input(s.size()),s.data(),s.size());if(!rr_state_load(s.size()))throw std::runtime_error("State load failed");}
auto end=rrFrames()+(argc>3?std::stoull(argv[3]):700);auto start=go_time_Now();while(rrFrames()<end){if(std::getenv("RR_AUTO"))rr_key(0x1c,rrFrames()%140<5);if(rr_run(10000)<0)throw std::runtime_error(rr_error());}std::cout<<rr_proof()<<"\n"<<rr_profile()<<"\n";std::cerr<<"Seconds "<<(go_time_Now()-start)/1e9<<"\n";
if(auto file=std::getenv("RR_SAVE")){auto n=rr_state_save();if(!n)throw std::runtime_error("State save failed");std::ofstream f(file,std::ios::binary);f.write((char*)rr_state_data(),n);}
if(argc>4){auto*p=rr_frame();std::ofstream f(argv[4],std::ios::binary);f<<"P6\n320 200\n255\n";for(int i=0;i<64000;i++)f.write((char*)p+i*4,3);}return 0;}catch(const std::exception&e){std::cerr<<e.what()<<"\n";return 1;}}
