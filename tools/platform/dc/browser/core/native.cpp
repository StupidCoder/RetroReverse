#include "api.cpp"
#include <filesystem>
int main(int argc,char**argv){try{
 if(argc<2)throw std::runtime_error("Usage: dc-native CUE [ADDITIONAL-FIELDS] [PPM]");std::ifstream cue(argv[1]);std::string text((std::istreambuf_iterator<char>(cue)),{});auto[name,tracks,error]=dc_parseCue(text);if(error)throw std::runtime_error(error.text);
 rrDiscFile.open(std::filesystem::path(argv[1]).parent_path()/name,std::ios::binary);if(!rrDiscFile)throw std::runtime_error("Cannot open disc");rrDiscFile.seekg(0,std::ios::end);auto bytes=uint64_t(rrDiscFile.tellg());
 for(int64_t i=0;i<tracks.n;i++){auto t=tracks[i];auto length=(i+1<tracks.n?uint64_t(tracks[i+1].FileOffset):bytes)-t.FileOffset;int64_t lba=-1;if(dc_Track_IsData(t)){uint8_t h[16];if(!rrReadLocalDisc(h,t.FileOffset,16))throw std::runtime_error("Cannot read track");auto bcd=[](uint8_t b){if((b&15)>9||(b>>4)>9)throw std::runtime_error("Invalid track header");return (b>>4)*10+(b&15);};lba=(bcd(h[12])*60+bcd(h[13]))*75+bcd(h[14])-150;}if(!rr_disc_track(t.Number,lba,t.FileOffset,length,dc_Track_IsData(t)))throw std::runtime_error("Invalid track");}
 if(!rr_init_file(bytes))throw std::runtime_error(rr_error());
 if(auto file=std::getenv("RR_LOAD")){std::ifstream f(file,std::ios::binary);std::vector<uint8_t>b((std::istreambuf_iterator<char>(f)),{});std::memcpy(rr_state_input(b.size()),b.data(),b.size());if(!rr_state_load(b.size()))throw std::runtime_error("Invalid state");}
 auto end=machine->Fields+(argc>2?std::stoull(argv[2]):60);auto start=go_time_Now();
 while(machine->Fields<end){if(std::getenv("RR_AUTO"))rr_pad(machine->Fields%120<8?12:0,0,0);if(rr_run(10000)<0)throw std::runtime_error(rr_error());}
 std::cout<<rr_proof()<<"\n"<<rr_profile()<<"\n";std::cerr<<"Seconds "<<(go_time_Now()-start)/1e9<<"\n";
 if(auto file=std::getenv("RR_SAVE")){auto n=rr_state_save();if(!n)throw std::runtime_error("State save failed");std::ofstream f(file,std::ios::binary);f.write((char*)rr_state_data(),n);}
 if(argc>3){auto b=rrFrame(machine);auto[w,h]=rrDimensions(machine);std::ofstream f(argv[3],std::ios::binary);f<<"P6\n"<<w<<" "<<h<<"\n255\n";for(size_t i=0;i<b.size();i+=4)f.write((char*)b.data()+i,3);}
 }catch(const std::exception&e){std::cerr<<e.what()<<"\n";return 1;}}
