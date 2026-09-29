#include "api.cpp"
int main(int argc,char**argv){try{
 if(argc<2)throw std::runtime_error("Usage: ps2-native IMAGE [FIELDS] [PPM]");
 rrDiscFile.open(argv[1],std::ios::binary);if(!rrDiscFile)throw std::runtime_error("Cannot open image");rrDiscFile.seekg(0,std::ios::end);if(!rr_init_file(rrDiscFile.tellg()))throw std::runtime_error(rr_error());
 if(auto p=std::getenv("PS2_BIOS")){std::ifstream f(p,std::ios::binary);std::vector<uint8_t>b((std::istreambuf_iterator<char>(f)),{});std::memcpy(rr_input(b.size()),b.data(),b.size());if(!rr_bios(b.size()))throw std::runtime_error("Firmware load failed");}
 if(!rr_boot())throw std::runtime_error(rr_error());
 if(auto file=std::getenv("PS2_LOAD")){std::ifstream f(file,std::ios::binary);std::vector<uint8_t>b((std::istreambuf_iterator<char>(f)),{});std::memcpy(rr_state_input(b.size()),b.data(),b.size());if(!rr_state_load(b.size()))throw std::runtime_error("Invalid state");}
 auto end=argc>2?std::stoull(argv[2]):120;auto start=go_time_Now();uint64_t last=machine->vblanks;
 while(machine->vblanks<end){if(std::getenv("PS2_AUTO"))rr_pad(machine->vblanks%120<10?0x4008:0,0,0);if(rr_run(10000)<0)throw std::runtime_error(rr_error());if(last!=machine->vblanks){last=machine->vblanks;if(std::getenv("PS2_TRACE"))std::cerr<<rr_proof()<<std::endl;}}
 std::cout<<rr_proof()<<"\n"<<rr_profile()<<"\n";std::cerr<<"Seconds "<<(go_time_Now()-start)/1e9<<" fields "<<machine->vblanks<<"\n";
 if(auto file=std::getenv("PS2_SAVE")){auto n=rr_state_save();if(!n)throw std::runtime_error("State save failed");std::ofstream f(file,std::ios::binary);f.write((char*)rr_state_data(),n);}
 if(argc>3){auto*b=rr_frame();std::ofstream f(argv[3],std::ios::binary);f<<"P6\n"<<width<<" "<<height<<"\n255\n";for(int64_t i=0;i<pixels.n;i+=4)f.write((char*)b+i,3);}
 }catch(const std::exception&e){std::cerr<<e.what()<<"\n";if(machine)for(auto s:machine->Log)std::cerr<<s<<"\n";return 1;}}
