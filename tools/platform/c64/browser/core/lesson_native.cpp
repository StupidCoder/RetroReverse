// Fort lesson acceptance and checkpoint authoring; never linked into the machine core.
#include "core.h"
#include <fstream>
#include <vector>
#include <string>
#include <cstdio>
#include <cstdlib>
#include <cstring>
static std::vector<uint8_t> read(const std::string& p){std::ifstream f(p,std::ios::binary);if(!f){fprintf(stderr,"Missing %s\n",p.c_str());exit(1);}return {std::istreambuf_iterator<char>(f),{}};}
static void require(bool v,const char* why){if(!v){fprintf(stderr,"%s: %s %s\n",why,rr_error(),rr_status());exit(1);}}
struct Recorder{
 std::ofstream out;bool comma=false;uint64_t pending=0;
 Recorder(const std::string& p):out(p){out<<"[";}
 void emit(const std::string& s){if(comma)out<<",";out<<s;comma=true;}
 void flush(){if(pending){emit("{\"run\":"+std::to_string(pending)+"}");pending=0;}}
 int run(int n,int kind=0,int target=0){int used=rr_run(n,kind,target);require(used>=0,"execute");pending+=used;return used;}
 void advance(int n){while(n){int step=n>10000?10000:n;run(step);n-=step;}}
 void until(int kind,int target,int budget=150000000){while(budget>0){int n=budget>10000?10000:budget;int used=run(n,kind,target);budget-=used;if(rr_stop_reason())return;}require(false,"lesson predicate timed out");}
 void input(const char* name,int a,int b=-1){flush();emit("{\""+std::string(name)+"\":"+(b<0?std::to_string(a):"["+std::to_string(a)+","+std::to_string(b)+"]")+"}");if(std::string(name)=="key")rr_key(a,b);else if(std::string(name)=="joystick")rr_joystick(a,b);else rr_play(a);}
 void type(const char* text){for(;*text;text++){input("key",*text,1);advance(100000);input("key",*text,0);advance(100000);}}
 void mark(int slot,const char* id){flush();emit("{\"chapter\":{\"id\":\""+std::string(id)+"\",\"slot\":"+std::to_string(slot)+"},\"state\":"+rr_status()+"}");require(rr_checkpoint(slot),"checkpoint");fprintf(stderr,"chapter %s at %.0f\n",id,rr_cycle());}
 void observe(const char* name){flush();emit("{\"observe\":\""+std::string(name)+"\",\"state\":"+rr_status()+"}");}
 void finish(){flush();out<<"]\n";}
};
int main(int argc,char** argv){
 if(argc!=5){fprintf(stderr,"usage: lesson-native TAPE FIXTURES ROMDIR OUTPUT_PREFIX\n");return 2;}
 std::string fixtures=argv[2],roms=argv[3],prefix=argv[4];int offset=0;
 for(auto name:{"basic.rom","kernal.rom","chargen.rom"}){auto bytes=read(roms+"/"+name);require(bytes.size()==(offset==16384?4096u:8192u),"ROM size");memcpy(rr_input()+offset,bytes.data(),bytes.size());offset+=bytes.size();}
 require(rr_init(8192,8192,4096),"reset");auto tape=read(argv[1]);memcpy(rr_input(),tape.data(),tape.size());require(rr_tape(tape.size()),"tape");
 Recorder r(prefix+"-schedule.json");r.advance(3000000);r.type("LOAD\r");r.input("play",1);
 r.until(5,27315);r.mark(0,"signal"); // first header payload, after nine 20-pulse countdown bytes
 rr_trace(1,0xf900,0xfbff);rr_trace_reads(0);r.until(3,0x033c,100000);
 require(rr_ram()[0x033c]==1,"KERNAL header type");
 std::ofstream(prefix+"-kernal.json")<<"{\"state\":"<<rr_status()<<",\"events\":"<<rr_events()<<",\"writer\":"<<rr_memory(0,0x33c)<<"}\n";
 rr_trace(0,0,0);rr_trace_reads(0);r.observe("kernal-byte");
 // Complete original KERNAL loader and settle in BASIC before typing RUN.
 r.until(5,48236);r.advance(250000);r.mark(1,"header");
 auto prepared=read(fixtures+"/prepared.bin");require(!memcmp(rr_ram()+0x33c,prepared.data()+0x33c,192)&&!memcmp(rr_ram()+0x801,prepared.data()+0x801,175),"KERNAL records");
 r.type("RUN\r");r.until(5,50395);r.until(2,0x351,1000);r.until(2,0x3e5,1000);r.mark(2,"takeover");
 rr_trace(1,0x351,0x3e5);r.until(3,0xe000,20000);require(rr_ram()[0xe000]==0x4c,"NOVALOAD byte");std::ofstream(prefix+"-byte.json")<<"{\"state\":"<<rr_status()<<",\"events\":"<<rr_events()<<"}\n";rr_trace(0,0,0);
 r.until(2,0xe000,20000000);r.mark(3,"loader-life");
 r.until(3,0x3f7,5000000);require(rr_ram()[0x3f5]==0x4c&&rr_ram()[0x3f6]==0&&rr_ram()[0x3f7]==0x86,"music entry patch");
 std::ofstream(prefix+"-patch.json")<<"["<<rr_memory(0,0x3f5)<<","<<rr_memory(0,0x3f6)<<","<<rr_memory(0,0x3f7)<<"]\n";r.observe("music-patch");
 r.until(2,0x8600);auto expected=read(fixtures+"/expected.bin");require(!memcmp(rr_ram()+0x7000,expected.data()+0x7000,0x4900),"loaded game image");r.observe("game-entry");r.emit("{\"checkpoint\":6}");require(rr_checkpoint(6),"initialization checkpoint");r.advance(3000000);
 r.input("joystick",2,16);
 // Retain exact input duration but stop before the terrain decompressor.
 r.until(2,0x8cdb,10000000);
 for(int tries=0;rr_ram()[0x21]!=0&&tries<10;tries++){r.advance(1);r.until(2,0x8cdb,5000000);}
 require(rr_ram()[0x21]==0&&rr_ram()[0x1a]==0&&rr_ram()[0x1b]==0x70,"terrain entry");
 r.input("joystick",2,0);r.mark(4,"unpacking");
 rr_trace(1,0x8cdb,0x8d2a);rr_trace_reads(1);r.until(2,0x8cdb,20000);
 std::ofstream(prefix+"-rle.json")<<"{\"state\":"<<rr_status()<<",\"events\":"<<rr_events()<<"}\n";rr_trace(0,0,0);rr_trace_reads(0);r.observe("first-run");
 r.until(2,0x8d1f,2000000);auto map=read(fixtures+"/raw-map.bin");require(map.size()==10240&&!memcmp(rr_ram()+0x503,map.data(),map.size()),"independent full terrain decompression");r.observe("terrain-expanded");
 for(int i=0;i<2000&&rr_ram()[0x9d]!=2;i++)r.advance(10000);require(rr_ram()[0x9d]==2,"gameplay");r.advance(2000000);r.mark(5,"world");
 // Capture from actual raster execution and emit generic pixel/writer evidence.
 r.flush();r.emit("{\"checkpoint\":10}");require(rr_checkpoint(10),"capture origin");r.emit("{\"captureBegin\":true}");rr_capture_begin();r.until(7,0,19656);r.until(7,0,19656);rr_capture_end();r.flush();r.emit("{\"captureEnd\":true}");r.observe("captured-world");
 std::ofstream proof(prefix+"-pixels.json");proof<<"[";bool first=true;
 for(auto xy:std::vector<std::pair<int,int>>{{150,220},{174,164},{90,70},{30,30},{200,90},{200,200}}){if(!first)proof<<",";first=false;proof<<rr_pixel(xy.first,xy.second);}proof<<"]\n";
 r.flush();r.emit("{\"restore\":10}");require(rr_restore(10),"restore capture origin");r.finish();
 std::ofstream(prefix+"-acceptance.json")<<"{\"kernalByte\":1,\"novaloadByte\":76,\"musicEntry\":34304,\"terrainBytes\":10240,\"chapters\":6,\"authentic\":true}\n";
 puts("PASS: six authentic chapter checkpoints, KERNAL/NOVALOAD bytes, music patch, full terrain output and pixel evidence");
}
