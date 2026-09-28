#include "core.h"
#include <cstdio>
#include <cstdlib>
#include <cstring>
#include <fstream>
#include <vector>
#include <string>
#include <chrono>
static std::vector<uint8_t> read(const std::string& path){std::ifstream f(path,std::ios::binary);if(!f){fprintf(stderr,"Missing %s\n",path.c_str());exit(1);}return {std::istreambuf_iterator<char>(f),{}};}
static void check(int ok){if(!ok){fprintf(stderr,"%s\n",rr_error());exit(1);}}
int main(int argc,char**argv){
 if(argc<4){fprintf(stderr,"usage: c64-native TAPE FIXTURES OUTPUT [ROMDIR]\n");return 2;}
 auto tape=read(argv[1]);if(tape.size()>2*1024*1024){fprintf(stderr,"Tape too large\n");return 1;}std::string fixture=argv[2],output=argv[3];
 if(argc>4){size_t pos=0;for(auto name:{"basic.rom","kernal.rom","chargen.rom"}){auto b=read(std::string(argv[4])+"/"+name);size_t want=pos==16384?4096:8192;if(b.size()!=want){fprintf(stderr,"Wrong ROM size: %s\n",name);return 1;}memcpy(rr_input()+pos,b.data(),b.size());pos+=b.size();}}
 else memset(rr_input(),0xea,20480);
 check(rr_init(8192,8192,4096));memcpy(rr_input(),tape.data(),tape.size());check(rr_tape(tape.size()));
 if(argc==4){auto ram=read(fixture+"/prepared.bin");memcpy(rr_input(),ram.data(),ram.size());check(rr_prepare(0x83e,48233));rr_play(1);
 // Reach immediately before the first byte, then record its actual bus evidence.
 for(unsigned attempts=0;;attempts++){if(attempts>10000){fprintf(stderr,"Prepared pulse target timed out\n");return 1;}if(rr_run(100000,1,0)<0){fprintf(stderr,"%s\n",rr_error());return 1;}unsigned pulse=0;const char* p=strstr(rr_status(),"\"pulse\":");sscanf(p,"\"pulse\":%u",&pulse);if(pulse>=50395)break;if(strstr(rr_status(),"\"cycle\":0"))return 1;}
 rr_run(250,2,0x351);rr_run(250,2,0x3e5);check(rr_checkpoint(0));rr_trace(1,0x351,0x3e5);rr_run(20000,3,0xe000);
 std::ofstream(output+".json")<<"{\"state\":"<<rr_status()<<",\"events\":"<<rr_events()<<"}\n";
 printf("%s\n",rr_status());printf("E000=%02X\n",rr_ram()[0xe000]);if(rr_ram()[0xe000]!=0x4c)return 1;
 std::string s=rr_status(),e=rr_events();check(rr_restore(0));rr_trace(1,0x351,0x3e5);rr_run(20000,3,0xe000);if(s!=rr_status()||e!=rr_events()){fprintf(stderr,"Replay mismatch\n");return 1;}
 // Validate every fastloader store at the moment it happens, before runtime overwrites.
 check(rr_restore(0));rr_trace(0,0,0);auto expected=read(fixture+"/expected.bin"),pages=read(fixture+"/pages.bin");
 unsigned stores=0;
 for(uint8_t page:pages)for(unsigned offset=0;offset<256;offset++){
  if(rr_run(1000000,4,0x38a)<0){fprintf(stderr,"%s\n",rr_error());return 1;}
  uint32_t bus=rr_bus();unsigned address=(page<<8)|offset;
  if((bus&65535)!=address||((bus>>16)&255)!=expected[address]){fprintf(stderr,"Store %u differs: wanted %04X=%02X, bus=%06X %s\n",stores,address,expected[address],bus,rr_status());return 1;}stores++;
 }
 printf("Verified %u live fastloader writes across %zu pages.\n",stores,pages.size());
 std::ofstream(output+"-fastload.json")<<"{\"stores\":"<<stores<<",\"pages\":"<<pages.size()<<",\"state\":"<<rr_status()<<"}\n";
 // Finish checksum/end-marker handling and the music-controlled epilogue.
 bool epilogue=false;for(int i=0;i<1000;i++){if(rr_run(10000,2,0x3f2)<0)break;if((rr_bus()&65535)==0x3f2){epilogue=true;break;}}
 if(!epilogue||rr_ram()[0xab]!=0||rr_ram()[0x3f5]!=0x4c||rr_ram()[0x3f7]!=0x86){fprintf(stderr,"Loader epilogue did not complete, status=%02X: %s\n",rr_ram()[0xab],rr_status());return 1;}
 std::ofstream(output+"-epilogue.json")<<"{\"statusAB\":0,\"patchedEntry\":34304,\"state\":"<<rr_status()<<"}\n";
 puts("Guest verified stream, completed loading-screen program, and patched JMP $8600; stopped before real ROM IOINIT.");
 // Corruption branches from the same byte boundary and must trip guest checksum handling.
 check(rr_restore(0));check(rr_corrupt(50396,672));rr_trace(0,0,0);
 for(int i=0;i<200&&rr_ram()[0xab]!=0x80;i++){if(rr_run(10000,0,0)<0)break;}
 if(rr_ram()[0xab]!=0x80){fprintf(stderr,"Corrupt pulse did not trip checksum: %s\n",rr_status());return 1;}
 std::ofstream(output+"-corrupt.json")<<"{\"statusAB\":128,\"state\":"<<rr_status()<<"}\n";puts("Corrupted pulse rejected by guest checksum routine ($AB=$80).");
 }else{

 // Actual keyboard/tape boot. Record only inputs, cycle budgets and observations;
 // the browser replays this schedule through the same generic hardware API.
 std::ofstream script(output+"-schedule.json");script<<"[";bool comma=false;uint64_t pending=0;
 auto emit=[&](const std::string& item){if(comma)script<<",";script<<item;comma=true;};
 auto flush=[&](){if(pending){emit("{\"run\":"+std::to_string(pending)+"}");pending=0;}};
 auto run=[&](int cycles,int kind=0,int target=0){int n=rr_run(cycles,kind,target);check(n>=0);pending+=n;return n;};
 auto advance=[&](int cycles){while(cycles){int n=cycles>10000?10000:cycles;run(n);cycles-=n;}};
 auto key=[&](int code,int down){flush();emit("{\"key\":["+std::to_string(code)+","+std::to_string(down)+"]}");rr_key(code,down);};
 auto type=[&](const char* text){for(;*text;text++){key(*text,1);advance(100000);key(*text,0);advance(100000);}};
 auto frame=[&](const std::string& name){std::ofstream f(output+"-"+name+".ppm",std::ios::binary);f<<"P6\n392 272\n255\n";auto pixels=rr_frame();for(int i=0;i<392*272;i++){char rgb[3]={char(pixels[i]),char(pixels[i]>>8),char(pixels[i]>>16)};f.write(rgb,3);}};
 auto observe=[&](const std::string& name){flush();emit("{\"observe\":\""+name+"\",\"state\":"+rr_status()+"}");std::ofstream(output+"-"+name+".json")<<rr_status()<<"\n";frame(name);fprintf(stderr,"%s: %s\n",name.c_str(),rr_status());};
 auto started=std::chrono::steady_clock::now();
 advance(3000000);type("LOAD\r");flush();emit("{\"play\":1}");rr_play(1);
 bool loaded=false;for(unsigned slice=0;slice<25000;slice++){
  run(10000,2,0xa474);unsigned pulse=0;sscanf(strstr(rr_status(),"\"pulse\":"),"\"pulse\":%u",&pulse);
  if(pulse>=48233&&strstr(rr_status(),"\"motor\":false")){loaded=true;break;}
 }
 if(!loaded){fprintf(stderr,"KERNAL LOAD timed out: %s\n",rr_status());return 1;}
 advance(200000);observe("basic-loaded");
 auto prepared=read(fixture+"/prepared.bin");
 if(memcmp(rr_ram()+0x801,prepared.data()+0x801,175)||memcmp(rr_ram()+0x33c,prepared.data()+0x33c,192)){fprintf(stderr,"KERNAL relocation/header mismatch\n");return 1;}
 type("RUN\r");auto expected=read(fixture+"/expected.bin"),pages=read(fixture+"/pages.bin");unsigned stores=0;
 for(uint8_t page:pages){for(unsigned offset=0;offset<256;offset++){
  for(int wait=0;wait<10;wait++){if(run(1000000,4,0x38a)<1000000)break;}uint32_t bus=rr_bus();unsigned address=page*256+offset;
  if((bus&65535)!=address||((bus>>16)&255)!=expected[address]){fprintf(stderr,"Authentic store %u differs at %04X: %s\n",stores,address,rr_status());return 1;}stores++;
 }if(stores==11*256){
 observe("loading");emit("{\"checkpoint\":0}");check(rr_checkpoint(0));advance(50000);std::string loadedReplay=rr_status();observe("loading-forward");emit("{\"restore\":0}");check(rr_restore(0));advance(50000);if(loadedReplay!=rr_status()){fprintf(stderr,"Loading replay mismatch\n");return 1;}observe("loading-replay");emit("{\"restore\":0}");check(rr_restore(0));
 }}
 observe("payload-loaded");bool entry=false;
 for(unsigned slice=0;slice<15000;slice++){run(10000,2,0x8600);if((rr_bus()&65535)==0x8600){entry=true;break;}}
 if(!entry||rr_ram()[0xab]!=0){fprintf(stderr,"Game-entry/checksum target failed: %s\n",rr_status());return 1;}
 for(unsigned a=0x7000;a<0xb900;a++)if(rr_ram()[a]!=expected[a]){fprintf(stderr,"Game image differs before entry at %04X\n",a);return 1;}
 observe("entry");advance(3000000);observe("title");
 flush();emit("{\"joystick\":[2,16]}");rr_joystick(2,16);advance(400000);flush();emit("{\"joystick\":[2,0]}");rr_joystick(2,0);
 for(int i=0;i<1500&&rr_ram()[0x9d]!=2;i++)advance(10000);
 if(rr_ram()[0x9d]!=2){fprintf(stderr,"Scripted fire did not reach gameplay: mode=%u %s\n",rr_ram()[0x9d],rr_status());return 1;}
 std::ofstream(output+"-graphics-ram.bin",std::ios::binary).write(reinterpret_cast<char*>(rr_ram()),65536);
 auto graphics=read(fixture+"/graphics.bin"),mask=read(fixture+"/graphics-mask.bin");unsigned graphicsBytes=0;
 for(unsigned a=0;a<65536;a++)if(mask[a]){if(rr_ram()[a]!=graphics[a]){fprintf(stderr,"Extracted graphics differ at %04X: %02X != %02X\n",a,rr_ram()[a],graphics[a]);return 1;}graphicsBytes++;}
 flush();emit("{\"checkpoint\":1}");check(rr_checkpoint(1));advance(2000000);std::string played=rr_status();observe("gameplay");
 double wallMs=std::chrono::duration<double,std::milli>(std::chrono::steady_clock::now()-started).count();
 flush();emit("{\"restore\":1}");check(rr_restore(1));advance(2000000);if(played!=rr_status()){fprintf(stderr,"Gameplay replay mismatch\n");return 1;}observe("replay");
 flush();script<<"]\n";
 std::ofstream(output+"-acceptance.json")<<"{\"authenticReset\":true,\"kernalBytes\":367,\"payloadStores\":"<<stores<<",\"graphicsBytes\":"<<graphicsBytes<<",\"gameplayReplay\":true,\"wallToGameplayMs\":"<<wallMs<<"}\n";
 puts("Reset-to-gameplay path, extracted graphics, and checkpoint replay passed.");

 }
 return 0;
}
