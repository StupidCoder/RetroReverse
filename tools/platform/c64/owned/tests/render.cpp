#include "../browser/core.cpp"
#include <cassert>
#include <iostream>
#include <set>
using namespace bridge;
static void contains(const char* s,const char* part){assert(std::string(s).find(part)!=std::string::npos);}
static std::vector<uint8_t> save(){const auto n=rr_state_save();assert(n);return {rr_state_data(),rr_state_data()+n};}
static void load(const std::vector<uint8_t>& b){auto p=rr_state_input(b.size());std::copy(b.begin(),b.end(),p);assert(rr_state_load(b.size()));}
static void boundaryFrame(){assert(rr_run(20000,7,0)>0);}
static void store(std::vector<uint8_t>& code,unsigned a,uint8_t v){code.insert(code.end(),{0xa9,v,0x8d,uint8_t(a),uint8_t(a>>8)});}
static void setup(){
 std::fill_n(rr_input(),20480,0);assert(rr_init(8192,8192,4096));std::fill_n(rr_input(),21,0);std::copy_n("C64-TAPE-RAW",12,rr_input());rr_input()[16]=rr_input()[20]=1;assert(rr_tape(21));
 std::fill_n(rr_input(),65536,0);std::fill_n(rr_input()+0x400,1000,1);std::fill_n(rr_input()+0x2008,8,0x80);std::fill_n(rr_input()+0x2808,8,255);std::fill_n(rr_input()+0x3000,63,255);rr_input()[0x7f8]=0xc0;
 std::vector<uint8_t> code;for(auto [a,v]:std::vector<std::pair<unsigned,uint8_t>>{{1,0x35},{0xd011,0x1b},{0xd016,8},{0xd018,0x18},{0xd020,6},{0xd021,0},{0xd000,100},{0xd001,100},{0xd015,1},{0xd01b,1},{0xd027,5}})store(code,a,v);
 const auto loop=0x800+code.size();code.insert(code.end(),{0x4c,uint8_t(loop),uint8_t(loop>>8)});std::copy(code.begin(),code.end(),rr_input()+0x800);assert(rr_prepare(0x800,0));board.state.color.fill(2);std::fill_n(rr_input(),228,0);assert(rr_state_bind());assert(rr_run(60000,0,0)>0);boundaryFrame();
}
int main(){assert(render::key(2,0)!=render::key(2,0xd000));assert(render::key(2,1)!=render::key(2,0xd001));setup();const auto initial=save();assert(rr_capture_begin());boundaryFrame();assert(rr_capture_end());contains(rr_raster_info(),"\"complete\":true");assert(render::complete());rr_raster_seek(271);unsigned sprites=0,behind=0,text=0;
 for(unsigned i=0;i<392*272;i++){const auto& p=render::captured[i];assert(p.cycle&&render::Palette[p.color]==rr_frame()[i]);assert(rr_raster_frame(2)[i]==rr_frame()[i]);if(p.hits&&!p.border){sprites++;assert(rr_raster_frame(1)[i]==render::Palette[5]);if(p.foreground){assert(p.color==2);behind++;}const auto& f=render::refs[p.sprites[0]];assert(f.address>=0x3000&&f.address<0x303f&&f.value==255);assert(f.pointer&&render::refs[f.pointer].address==0x7f8);}if(!p.border&&p.foreground){text++;const auto& f=render::refs[p.data];assert(f.address>=0x2008&&f.address<0x2010&&f.value==0x80);}}
 assert(sprites==504&&behind&&text>7000);const auto observed=save();assert(rr_replay_seek(272));assert(std::equal(rr_frame(),rr_frame()+392*272,rr_replay_frame()));
 for(int y:{0,70,100,200,271}){rr_raster_seek(y);rr_tileset_data();rr_raster_pixel(0,100,100);rr_raster_pixel(1,100,100);rr_raster_pixel(2,50,y);}assert(save()==observed);
 load(initial);boundaryFrame();assert(save()==observed); // capture does not alter any hardware byte
 // Every valid mode plus the character ROM overlay uses the owned VIC in both
 // projections. Sprite-off frozen graphics equal the actual steady frame.
 for(const auto& mode:std::vector<std::array<int,3>>{{0x1b,8,0x18},{0x1b,0x18,0x18},{0x3b,8,0x18},{0x3b,0x18,0x18},{0x5b,8,0x18},{0x1b,8,0x14}}){board.write(0xd015,0);board.write(0xd011,mode[0]);board.write(0xd016,mode[1]);board.write(0xd018,mode[2]);board.state.color.fill(10);rr_run(40000,0,0);boundaryFrame();rr_capture_begin();boundaryFrame();rr_capture_end();rr_raster_seek(271);assert(std::equal(rr_frame(),rr_frame()+392*272,rr_raster_frame(0)));}
 setup();std::vector<uint8_t> code;auto wait=[&](uint8_t line){code.insert(code.end(),{0xad,0x12,0xd0,0xc9,line,0xd0,0xf9});};wait(70);store(code,0x2008,0x40);wait(120);store(code,0xd018,0x1a);wait(250);store(code,0xd018,0x18);store(code,0x2008,0x80);code.insert(code.end(),{0x4c,0,8});std::copy(code.begin(),code.end(),rr_ram()+0x800);Cpu cpu;cpu.start(0x800);board.state.cpu=cpu.state;boundaryFrame();boundaryFrame();rr_capture_begin();boundaryFrame();rr_capture_end();
 int changed=-1,switched=-1;for(unsigned y=1;y<272;y++){if(render::lines[y].ram[0x2008]!=render::lines[y-1].ram[0x2008]&&changed<0)changed=y;if(render::lines[y].regs[0x18]==0x1a&&render::lines[y-1].regs[0x18]==0x18)switched=y;}assert(changed>0&&switched>changed);
 rr_raster_seek(changed-1);assert(rr_tileset_data()[2192+8]==0x80);rr_raster_seek(changed);assert(rr_tileset_data()[2192+8]==0x40);contains(rr_raster_seek(changed),"\"tileWrites\":1");rr_raster_seek(switched);assert(rr_tileset_data()[65]==0x28&&rr_tileset_data()[2192+8]==255);contains(rr_raster_pixel(2,0,switched+1),"not been drawn");
 bool first=false,second=false,writer=false;int chosen=-1;for(unsigned i=0;i<render::captured.size();i++){const auto& p=render::captured[i];if(p.border||!p.data)continue;const auto& f=render::refs[p.data];first|=f.address>=0x2008&&f.address<0x2010;second|=f.address>=0x2808&&f.address<0x2810;if(f.address==0x2008&&f.value==0x40){assert(f.writer.pc==0x809&&f.writer.code[0]==0x8d&&f.writer.code[1]==8&&f.writer.code[2]==0x20&&f.writer.cycle<f.cycle);writer=true;chosen=i;}}
 assert(first&&second&&writer&&chosen>=0);const std::string historical=rr_pixel(chosen%392,chosen/392);std::fill_n(rr_ram(),65536,0);assert(historical==rr_pixel(chosen%392,chosen/392));rr_raster_seek(changed);assert(rr_tileset_data()[2192+8]==0x40);
 // Same-value writes have distinct writer identities, and pixels drawn in the
 // write's own clock still use the preceding VIC-register version.
 setup();code.clear();for(auto value:{3,3,6})store(code,0xd020,value);code.insert(code.end(),{0x4c,0,8});std::copy(code.begin(),code.end(),rr_ram()+0x800);cpu.start(0x800);board.state.cpu=cpu.state;rr_capture_begin();boundaryFrame();boundaryFrame();rr_capture_end();std::set<uint64_t> versions;unsigned checked=0;
 for(const auto& p:render::captured)if(p.border&&p.control){const auto& c=render::controls[p.control];const auto& r=render::refs[c.refs[0x20]];assert(r.value==p.color);if(r.writer.cycle){assert(r.writer.cycle<p.cycle&&r.value==r.writer.value);checked++;if(r.value==3)versions.insert(r.writer.cycle);}}
 assert(checked>10000&&versions.size()>100&&!render::overflow);
 rr_capture_begin();rr_capture_end();contains(rr_raster_info(),"\"complete\":false");contains(rr_raster_seek(0),"\"error\"");assert(!rr_tileset_size());
 // A checkpoint in the last cycle of STA must retain the bytes actually
 // fetched, even if the backing instruction is subsequently overwritten.
 setup();rr_ram()[0x800]=0x8d;rr_ram()[0x801]=0x20;rr_ram()[0x802]=0xd0;cpu.start(0x800);cpu.state.a=4;board.state.cpu=cpu.state;
 assert(rr_run(3,0,0)==3);assert(board.state.cpu.bus.write);const auto partial=save();load(partial);rr_ram()[0x800]=0xea;rr_ram()[0x801]=rr_ram()[0x802]=0;
 assert(rr_run(1,0,0)==1);const auto& resumed=render::history[render::key(2,0xd020)];assert(resumed.pc==0x800&&resumed.value==4&&(resumed.code==std::array<uint8_t,3>{0x8d,0x20,0xd0}));
 std::cout<<"PASS owned rendering: actual text/bitmap/ROM/sprite fetches, priority, frozen layers, charset split, glyph/register writer versions, same-value stores, complete replay and unchanged checkpoint bytes\n";
}
