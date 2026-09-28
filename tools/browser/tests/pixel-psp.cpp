#include "../../platform/psp/browser/core/api.cpp"
#include <cassert>
#include <filesystem>
static Slice<uint8_t>hex(const char*s){auto out=Slice<uint8_t>::make(std::strlen(s)/2);for(int i=0;i<out.n;i++)out[i]=std::stoul(std::string(s+i*2,2),nullptr,16);return out;}
static void cryptoTests(){
 for(auto [message,want]:std::array<std::pair<const char*,const char*>,3>{{{"","da39a3ee5e6b4b0d3255bfef95601890afd80709"},{"abc","a9993e364706816aba3e25717850c26c9cd0d89d"},{"abcdbcdecdefdefgefghfghighijhijkijkljklmklmnlmnomnopnopq","84983e441c3bd26ebaae4aa1f95129e5e54670f1"}}}){auto digest=go_sha1_Sum(cast<Slice<uint8_t>>(std::string(message)));auto expected=hex(want);assert(std::equal(digest.begin(),digest.end(),expected.begin()));}
 auto key=hex("000102030405060708090a0b0c0d0e0f"),encrypted=hex("69c4e0d86a7b0430d8cdb78070b4c55aaa"),plain=hex("00112233445566778899aabbccddeeffaa");auto[result,e]=rrcrypto::cbc(key,encrypted);assert(!e&&std::equal(result.begin(),result.end(),plain.begin()));assert(std::get<1>(rrcrypto::cbc({},encrypted)));
}
static void discTests(){
 // Construct a public two-block CSO: one raw sector and one DEFLATE sector.
 // Validate repeated, reverse-order random reads and malformed index rejection.
 std::vector<uint8_t>image(24+3*4+2048),raw(2048);for(unsigned i=0;i<2048;i++)raw[i]=i*31;
 std::memcpy(image.data(),"CISO",4);auto put=[&](unsigned at,uint32_t v){std::memcpy(image.data()+at,&v,4);};put(4,24);put(8,4096);put(16,2048);image[20]=1;put(24,36|0x80000000);put(28,36+2048);std::memcpy(image.data()+36,raw.data(),2048);
 z_stream z{};assert(deflateInit2(&z,6,Z_DEFLATED,-15,8,Z_DEFAULT_STRATEGY)==Z_OK);std::array<uint8_t,4096>compressed{};z.next_in=raw.data();z.avail_in=2048;z.next_out=compressed.data();z.avail_out=compressed.size();assert(deflate(&z,Z_FINISH)==Z_STREAM_END);deflateEnd(&z);image.insert(image.end(),compressed.begin(),compressed.begin()+z.total_out);put(32,image.size());
 // CSO geometry itself permits small images; a UMD volume is checked separately.
 auto path=std::filesystem::temp_directory_path()/"rr-psp-public.cso";{std::ofstream f(path,std::ios::binary);f.write((char*)image.data(),image.size());}
 rrDiscFile.open(path,std::ios::binary);BlockSource reader;reader.mount(image.size());for(auto n:{1,0,1,0}){auto block=reader.block(n);assert(std::equal(block.begin(),block.end(),raw.begin()));}rrDiscFile.close();
 image[28]=0;image[29]=0;{std::ofstream f(path,std::ios::binary);f.write((char*)image.data(),image.size());}rrDiscFile.open(path,std::ios::binary);bool rejected=false;try{BlockSource invalid;invalid.mount(image.size());}catch(...){rejected=true;}assert(rejected);rrDiscFile.close();std::filesystem::remove(path);
}
int main(){cryptoTests();discTests();machine=psp_NewMachine();machine->vol=arenaNew(psp_Volume{});rrBind(machine);auto m=machine;
 psp_geState state{};m->geSt=&state;state.fbLow=0x04000000;state.fbStride=512;state.fbFmt=3;state.scX1=479;state.scY1=271;m->fbAddr=0x04000000;m->fbWidth=512;m->fbFormat=3;
 // All supported texture formats, mip levels, swizzles and memory aliases use
 // the unchanged scalar sampler as an independent oracle for the fast view.
 for(int i=0;i<65536;i++)m->ram[i]=uint8_t(i*31+i/256);for(int i=0;i<256;i++)state.clut[i]=i*0x01020407;state.clutFmt=0xff00;state.texAddr=0x88000000;state.texStride=64;state.texW=state.texH=64;state.texAddrN[1]=0x08004000;state.texStrideN[1]=32;
 for(unsigned fmt=0;fmt<=5;fmt++)for(bool swizzle:{false,true}){state.texFmt=fmt;state.texSwizzle=swizzle;rrTextureScope scope(m,&state);for(unsigned level=0;level<2;level++)for(unsigned y=0;y<32;y++)for(unsigned x=0;x<32;x++)assert(psp_Machine_sampleTexLvl(m,&state,x,y,level)==psp_Machine_sampleTexLvl_Reference(m,&state,x,y,level));m->ram[0]^=255;assert(psp_Machine_sampleTexLvl(m,&state,0,0,0)==psp_Machine_sampleTexLvl_Reference(m,&state,0,0,0));}
 assert(rr_capture_begin());{rrGETraceScope scope;m->OnGeCmd(0x04060002);psp_Machine_storePixel(m,&state,0x04000000,0,255,0,0,255);m->OnGeCmd(0x04060002);psp_Machine_storePixel(m,&state,0x04000000,0,0,255,0,255);m->OnPixel(0,0,psp_PixelEvent{false,false,false,false,true,false,0,0,255,255});}
 rrWrite(m,0x88000010,0x12345678,4);assert(rr_capture_end());assert(rrcapture::trace.shadow==rrcapture::trace.final);assert(!rrcapture::trace.overflow);auto pixel=std::string(rr_pixel(0,0));assert(pixel.find("displayFormat\":3")!=std::string::npos);assert(pixel.find("depthRejected\":true")!=std::string::npos);rr_replay_begin();assert(rrreplay::replay.steps.size()>=3);while(!rr_replay_seek(1)){}assert(rr_replay_frame()[0]==255);while(!rr_replay_seek(rrreplay::replay.steps.size())){}auto final=rrFrame(m);assert(std::equal(final.begin(),final.end(),rr_replay_frame()));while(!rr_replay_seek(0)){}assert(rr_replay_frame()[0]==0);
 // Preserve the persistent display clock across arbitrary execution slices.
 m->geSt=arenaNew(state);allegrex_CPU_SetPC(m->CPU,0x08010000);rrWrite(m,0x08010000,0x0a004000,4);rrWrite(m,0x08010004,0,4);rrUntilVBlank=100;rrAnalogX=91;rrAnalogY=220;
 auto n=rr_state_save();assert(n>33554432);auto saved=stateOutput;for(int i=0;i<10;i++)assert(rr_run(13)>=0);auto proof=std::string(rr_proof());auto clock=rrUntilVBlank;
 std::memcpy(rr_state_input(saved.size()),saved.data(),saved.size());assert(rr_state_load(saved.size()));assert(rrAnalogX==91&&rrAnalogY==220);assert(rr_run(130)>=0);assert(std::string(rr_proof())==proof&&rrUntilVBlank==clock);
 // A failed import is transactional.
 auto before=std::string(rr_proof());rr_state_input(12);assert(!rr_state_load(12));assert(std::string(rr_proof())==before);
}
