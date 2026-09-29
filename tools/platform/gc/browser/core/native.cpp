#include "api.cpp"
int main(int argc,char**argv){try{
 if(argc<2)throw std::runtime_error("Usage: gc-native IMAGE [FIELDS] [PPM]");
 rrDiscFile.open(argv[1],std::ios::binary);if(!rrDiscFile)throw std::runtime_error("Cannot open image");rrDiscFile.seekg(0,std::ios::end);if(!rr_init_file(rrDiscFile.tellg()))throw std::runtime_error(rr_error());
 if(auto file=std::getenv("GC_LOAD")){std::ifstream f(file,std::ios::binary);std::vector<uint8_t>b((std::istreambuf_iterator<char>(f)),{});std::memcpy(rr_state_input(b.size()),b.data(),b.size());if(!rr_state_load(b.size()))throw std::runtime_error("Invalid state");}
 auto end=argc>2?std::stoull(argv[2]):120;auto start=go_time_Now();uint64_t last=machine->vi.Field;
 while(machine->vi.Field<end){if(std::getenv("GC_AUTO"))rr_pad(machine->vi.Field%120<10?0x1100:0,0,0);if(rr_run(10000)<0)throw std::runtime_error(rr_error());if(last!=machine->vi.Field){last=machine->vi.Field;if(std::getenv("GC_TRACE"))std::cerr<<rr_proof()<<std::endl;}}
 std::cout<<rr_proof()<<"\n"<<rr_profile()<<"\n";std::cerr<<"Seconds "<<(go_time_Now()-start)/1e9<<" fields "<<machine->vi.Field<<"\n";
 if(auto file=std::getenv("GC_SAVE")){auto n=rr_state_save();if(!n)throw std::runtime_error("State save failed");std::ofstream f(file,std::ios::binary);f.write((char*)rr_state_data(),n);}
 if(argc>3){auto b=rrFrame(machine);std::ofstream f(argv[3],std::ios::binary);f<<"P6\n640 480\n255\n";for(size_t i=0;i<b.size();i+=4)f.write((char*)b.data()+i,3);}
 }catch(const std::exception&e){std::cerr<<e.what()<<"\n";if(machine)for(auto s:machine->Log)std::cerr<<s<<"\n";return 1;}}
