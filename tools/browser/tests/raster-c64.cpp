#include "../../platform/c64/browser/core/core.cpp"
#include <cassert>
#include <iostream>
static void contains(const char*s,const char*part){assert(std::string(s).find(part)!=std::string::npos);}
static void boundary(){assert(rr_run(20000,7,0)>0);}
int main(){
 memset(rr_input(),0xea,20480);assert(rr_init(8192,8192,4096));memset(input,0,65536);
 memset(input+0x400,1,1000);memset(input+0x2008,0x80,8);memset(input+0x2808,0xff,8);
 memset(input+0x3000,0xff,63);input[0x7f8]=0xc0;
 std::vector<uint8_t> code;
 auto store=[&](unsigned a,uint8_t v){code.insert(code.end(),{0xa9,v,0x8d,uint8_t(a),uint8_t(a>>8)});};
 store(1,0x35);store(0xd011,0x1b);store(0xd016,8);store(0xd018,0x18);store(0xd020,6);store(0xd021,0);
 store(0xd000,100);store(0xd001,100);store(0xd015,1);store(0xd01b,1);store(0xd027,5);
 unsigned loop=0x800+code.size();code.insert(code.end(),{0x4c,uint8_t(loop),uint8_t(loop>>8)});
 memcpy(input+0x800,code.data(),code.size());pulse_count=1;durations[0]=100000;assert(rr_prepare(0x800,0));memset(machine.color_ram,2,1000);
 rr_run(60000,0,0);boundary();rr_capture_begin();boundary();rr_capture_end();
 contains(rr_raster_info(),"\"complete\":true");rr_raster_seek(271);
 unsigned covered=0,behind=0;
 for(int i=0;i<392*272;i++){
  auto&p=observation::captured[i];assert(p.cycle);
  assert(raster::pictures[2][i]==m6569_color(p.color&15));
  if(!p.spriteMask||p.border)assert(raster::pictures[0][i]==m6569_color(p.color&15));
  if(p.spriteMask&&!p.border){assert(raster::pictures[1][i]==m6569_color(5));covered++;if(p.foreground){assert(p.color==2);behind++;}}
 }
 assert(covered==24*21&&behind>0);
 // All supported text/bitmap modes and the VIC's character-ROM overlay use
 // the same renderer in the frozen view. No CPU or display-register shortcuts.
 auto reg=[](int a,int value){uint64_t pins=a;M6569_SET_DATA(pins,value);_m6569_write(&machine.vic,pins);};
 reg(0xd015,0);memset(machine.color_ram,10,1000);
 for(auto mode:std::vector<std::array<int,3>>{{0x1b,8,0x18},{0x1b,0x18,0x18},{0x3b,8,0x18},{0x3b,0x18,0x18},{0x5b,8,0x18},{0x1b,8,0x14}}){
  reg(0xd011,mode[0]);reg(0xd016,mode[1]);reg(0xd018,mode[2]);rr_run(40000,0,0);boundary();rr_capture_begin();boundary();rr_capture_end();rr_raster_seek(271);
  assert(!memcmp(rr_raster_frame(0),rr_frame(),392*272*4));
 }
 reg(0xd011,0x1b);reg(0xd016,8);reg(0xd018,0x18);reg(0xd015,1);memset(machine.color_ram,2,1000);
 // Run real guest writes: redefine a character, then switch charset mid-frame.
 code.clear();auto wait=[&](uint8_t line){code.insert(code.end(),{0xad,0x12,0xd0,0xc9,line,0xd0,0xf9});};
 wait(70);store(0x2008,0x40);wait(120);store(0xd018,0x1a);wait(250);store(0xd018,0x18);store(0x2008,0x80);
 code.insert(code.end(),{0x4c,0x00,0x08});memcpy(machine.ram+0x800,code.data(),code.size());
 machine.cpu.PC=0x800;machine.pins=M6502_SYNC|M6502_RW;M6502_SET_ADDR(machine.pins,0x800);M6502_SET_DATA(machine.pins,machine.ram[0x800]);ctx.pc=0x800;
 boundary();boundary();rr_capture_begin();boundary();rr_capture_end();
 int changed=-1,switched=-1;for(int y=1;y<272;y++){
  if(raster::lines[y].ram[0x2008]!=raster::lines[y-1].ram[0x2008]&&changed<0)changed=y;
  if(raster::lines[y].regs[0x18]==0x1a&&raster::lines[y-1].regs[0x18]==0x18)switched=y;
 }
 assert(changed>0&&switched>changed);
 contains(rr_raster_seek(changed),"\"tileWrites\":1");contains(rr_raster_seek(switched),"\"charsetBase\":10240");
 auto newBackground=raster::pictures[0];rr_raster_seek(switched-1);assert(newBackground!=raster::pictures[0]);
 assert(raster::pictures[2][switched*392]==0);contains(rr_raster_pixel(2,0,switched),"not been drawn");
 // Actual pixels use cycle-level fetched bytes, not the frozen preview's state.
 bool first=false,second=false,writer=false;
 for(int y=0;y<272;y++)for(int x=0;x<392;x++){
  auto&p=observation::captured[y*392+x];if(p.border||!p.graphics)continue;auto&r=observation::refs[observation::graphs[p.graphics].data];
  first|=r.address>=0x2008&&r.address<0x2010;second|=r.address>=0x2808&&r.address<0x2810;
  if(r.address==0x2008&&r.value==0x40){assert(r.writer.pc==0x809);writer=true;}
 }
 assert(first&&second&&writer);
 auto size=rr_state_save();std::vector<uint8_t> saved(rr_state_data(),rr_state_data()+size);auto status=std::string(rr_status());
 for(int y:{0,70,100,200,271,0,271}){contains(rr_raster_seek(y),"\"complete\":true");rr_raster_pixel(0,100,100);rr_raster_pixel(1,100,100);rr_raster_pixel(2,50,y);}
 assert(std::string(rr_status())==status);assert(rr_state_save()==size);assert(!memcmp(rr_state_data(),saved.data(),size));
 assert(!memcmp(rr_raster_frame(2),rr_frame(),392*272*4));
 auto frozen=raster::pictures[0];memset(machine.ram,0,65536);memset(machine.color_ram,0,1024);rr_raster_seek(271);assert(raster::pictures[0]==frozen);
 rr_capture_begin();rr_capture_end();contains(rr_raster_info(),"\"complete\":false");contains(rr_raster_seek(0),"\"error\"");
 std::cout<<"C64 raster: frozen VIC layers, priority, charset split, glyph writers, accumulated output and unchanged state pass\n";
}
