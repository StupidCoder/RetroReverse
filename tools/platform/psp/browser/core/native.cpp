#include "api.cpp"
// Private-media harness. Same API and execution boundaries as the web worker.
int main(int argc,char**argv){try{
 if(argc<2)throw std::runtime_error("Usage: psp-native IMAGE [STEPS] [PPM]");
 rrDiscFile.open(argv[1],std::ios::binary);if(!rrDiscFile)throw std::runtime_error("Cannot open image");rrDiscFile.seekg(0,std::ios::end);if(!rr_init_file(rrDiscFile.tellg()))throw std::runtime_error(rr_error());
 if(auto file=std::getenv("PSP_LOAD")){std::ifstream f(file,std::ios::binary);std::vector<uint8_t>b((std::istreambuf_iterator<char>(f)),{});std::memcpy(rr_state_input(b.size()),b.data(),b.size());if(!rr_state_load(b.size()))throw std::runtime_error("State load failed");}
 const auto end=rrSteps+(argc>2?std::stoull(argv[2]):100000000ull);auto start=go_time_Now();uint64_t last=rrFrames;bool automatic=std::getenv("PSP_CROSS")!=nullptr;uint64_t maxFrames=std::getenv("PSP_FRAMES")?std::stoull(std::getenv("PSP_FRAMES")):UINT64_MAX;
 if(std::getenv("PSP_CAPTURE"))rr_capture_begin();
 while(rrSteps<end&&rrFrames<maxFrames){if(automatic)rr_pad(machine->vblanks>180&&machine->vblanks%120<20?16384:0,0,0);if(rr_run(std::min<uint64_t>(10000,end-rrSteps))<0)throw std::runtime_error(rr_error());if(rrFrames!=last){last=rrFrames;if(std::getenv("PSP_TRACE"))std::cout<<rr_proof()<<std::endl;}}
 if(std::getenv("PSP_CAPTURE")){rr_capture_end();std::cout<<rr_capture_info()<<std::endl;}
 std::cout<<rr_proof()<<"\n"<<rr_profile()<<"\n";std::cerr<<"Seconds "<<(go_time_Now()-start)/1e9<<" frames "<<rrFrames<<" steps "<<rrSteps<<" reads "<<disc.reads<<"\n";
 if(auto file=std::getenv("PSP_SAVE")){auto n=rr_state_save();if(!n)throw std::runtime_error("State save failed");std::ofstream f(file,std::ios::binary);f.write((char*)rr_state_data(),n);}
 if(argc>3){auto b=rrFrame(machine);std::ofstream f(argv[3],std::ios::binary);f<<"P6\n480 272\n255\n";for(size_t i=0;i<b.size();i+=4)f.write((char*)b.data()+i,3);}if(std::getenv("PSP_LOG"))for(auto s:machine->Log)std::cerr<<s<<"\n";
 }catch(const std::exception&e){std::cerr<<e.what()<<"\n";return 1;}}
