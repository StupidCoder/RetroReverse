// Targeted boundaries exercised by tape loading and Fort's raster splits.
#include "../../platform/c64/browser/core/core.cpp"
#include <cassert>
#include <string>
#include <vector>
#include <cstdio>
#include <set>
static uint64_t cia_write(m6526_t& c,int a,int v){uint64_t p=M6526_CS;M6502_SET_ADDR(p,a);M6526_SET_DATA(p,v);return m6526_tick(&c,p);}
static void cia_ticks(m6526_t& c,int n){while(n--)m6526_tick(&c,M6526_RW);}
int main(){
 m6526_t c; m6526_init(&c);cia_write(c,13,0x90);cia_ticks(c,3);m6526_tick(&c,M6526_RW|M6526_FLAG);cia_ticks(c,4);assert(c.intr.icr&16);assert(c.intr.irq);
 uint64_t p=M6526_RW|M6526_CS;M6502_SET_ADDR(p,13);p=m6526_tick(&c,p);assert(M6526_GET_DATA(p)&0x90);cia_ticks(c,4);assert(!c.intr.irq);
 cia_write(c,4,0xf4);cia_write(c,5,3);cia_write(c,14,0x11);cia_ticks(c,300);assert(c.ta.counter<0x300&&c.ta.counter>0x200);cia_write(c,14,0x11);cia_ticks(c,675);assert(c.ta.counter<0x200);
 memset(rr_input(),0xea,20480);assert(rr_init(8192,8192,4096));machine.ram[0xa000]=0x12;machine.ram[0xe000]=0x34;
 _c64_cpu_port_out(0x37,&machine);assert(mem_rd(&machine.mem_cpu,0xa000)==0xea);assert(mem_rd(&machine.mem_cpu,0xe000)==0xea);
 _c64_cpu_port_out(0x35,&machine);assert(mem_rd(&machine.mem_cpu,0xa000)==0x12);assert(mem_rd(&machine.mem_cpu,0xe000)==0x34);assert(machine.io_mapped);
 _c64_cpu_port_out(0x30,&machine);assert(!machine.io_mapped);
 // A guest program configures a raster interrupt, then waits in RAM.
 const uint8_t program[]={0xa9,0x35,0x85,1,0xa9,0x1b,0x8d,0x11,0xd0,0xa9,0x76,0x8d,0x12,0xd0,0xa9,1,0x8d,0x1a,0xd0,0x4c,0x13,0x08};
 memset(input,0,65536);memcpy(input+0x800,program,sizeof(program));durations[0]=100000;pulse_count=1;assert(rr_prepare(0x800,0));rr_trace(0,0,0);
 int irqline=-1;bool stalled=false;
 for(int i=0;i<19656;i++){rr_run(1,0,0);if(machine.vic.rs.v_count==0x76&&(machine.pins&M6502_IRQ))irqline=machine.vic.rs.v_count;if(machine.pins&M6502_RDY)stalled=true;}
 assert(irqline==0x76);assert(stalled);assert(ctx.frames==1);
 // Snapshot includes timer pipelines, raster, SID sample phase and CPU microstate.
 assert(rr_checkpoint(0));rr_run(9850,0,0);std::string state1=rr_status();std::vector<float> audio1(audio,audio+audio_count);assert(rr_restore(0));rr_run(9850,0,0);assert(state1==rr_status());assert(audio1==std::vector<float>(audio,audio+audio_count));
 // Known synthetic character and sprite patterns rendered by live VIC fetches.
 memset(rr_input(),0xea,20480);assert(rr_init(8192,8192,4096));
 memset(input,0,65536);memset(input+0x400,1,1000);memset(input+0x2008,0x80,8);memset(input+0x3000,0xff,63);input[0x7f8]=0xc0;
 std::vector<uint8_t> code;
 auto store=[&](unsigned a,uint8_t v){code.insert(code.end(),{0xa9,v,0x8d,uint8_t(a),uint8_t(a>>8)});};
 store(1,0x35);store(0xd011,0x1b);store(0xd016,0x08);store(0xd018,0x18);store(0xd020,6);store(0xd021,0);
 store(0xd000,100);store(0xd001,100);store(0xd015,1);store(0xd027,5);
 store(0xd400,0);store(0xd401,0x20);store(0xd404,0x21);store(0xd405,0);store(0xd406,0xf0);store(0xd418,15);
 unsigned loop=0x800+code.size();code.insert(code.end(),{0x4c,uint8_t(loop),uint8_t(loop>>8)});memcpy(input+0x800,code.data(),code.size());
 pulse_count=1;durations[0]=100000;assert(rr_prepare(0x800,0));memset(machine.color_ram,2,1000);assert(rr_run(60000,0,0)>0);
 unsigned green=0,red=0;for(auto v:machine.fb){green+=v==5;red+=v==2;}
 assert(green==24*21);assert(red>7000&&red<8001);assert(std::any_of(audio,audio+audio_count,[](float f){return f!=0;}));
 // Capture actual fetches, then overwrite RAM: selected evidence must stay historical.
 assert(rr_checkpoint(6));std::string beforeCapture=rr_status();rr_capture_begin();rr_run(19656*2,0,0);rr_capture_end();
 bool textFound=false,spriteFound=false;
 for(unsigned i=0;i<392*272;i++){
  const auto& p=observation::captured[i];
  if(!p.border&&p.color==2&&p.graphics){const auto& g=observation::graphs[p.graphics];const auto& data=observation::refs[g.data];const auto& screen=observation::refs[g.screen];assert(data.address>=0x2008&&data.address<0x2010&&data.value==0x80);assert(screen.address>=0x400&&screen.address<0x7e8&&screen.value==1);textFound=true;}
  if(!p.border&&p.color==5&&p.spriteMask==1){const auto& data=observation::refs[p.sprites[0]];assert(data.address>=0x3000&&data.address<0x303f&&data.value==255);spriteFound=true;}
 }
 assert(textFound&&spriteFound&&observation::overflow==0);
 std::string savedPixel;int chosen=-1;for(unsigned i=0;i<392*272;i++)if(observation::captured[i].spriteMask==1){chosen=i;break;}
 assert(chosen>=0);savedPixel=rr_pixel(chosen%392,chosen/392);machine.ram[0x3000]=0;assert(savedPixel==rr_pixel(chosen%392,chosen/392));
 assert(rr_restore(6));assert(beforeCapture==rr_status());
 // Register writes retain their fetched instruction bytes and separate I/O from RAM.
 const auto colorWrite=observation::history.last[observation::key(2,0xd027)];assert(colorWrite.value==5&&colorWrite.code[0]==0x8d&&colorWrite.code[1]==0x27&&colorWrite.code[2]==0xd0);
 machine.ram[colorWrite.pc]=0xea;assert(observation::history.last[observation::key(2,0xd027)].code[0]==0x8d);
 assert(observation::history.last[observation::key(0,0xd027)].cycle==0);
 assert(rr_previous()>=0&&rr_previous()<rr_cycle());
 // The VIC renders before committing a CPU register write in the same tick.
 // Alternating and repeated writes must retain the value AND exact writer version.
 memset(input,0,65536);code.clear();store(1,0x35);store(0xd011,0x1b);store(0xd016,8);
 loop=0x800+code.size();store(0xd020,3);store(0xd020,3);store(0xd020,6);
 code.insert(code.end(),{0x4c,uint8_t(loop),uint8_t(loop>>8)});memcpy(input+0x800,code.data(),code.size());
 assert(rr_prepare(0x800,0));rr_capture_begin();rr_run(39312,0,0);rr_capture_end();
 unsigned checked=0;std::set<uint64_t> repeatedVersions;
 for(const auto& px:observation::captured)if(px.control&&px.border==2){
  const auto& r=observation::refs[observation::controls[px.control].refs[0x20]];
  if(r.writer.cycle){assert(r.value==(r.writer.value&15));assert(px.color==r.value);assert(r.writer.cycle<px.cycle);checked++;if(r.value==3)repeatedVersions.insert(r.writer.cycle);}
 }
 assert(checked>10000&&repeatedVersions.size()>100&&observation::overflow==0);
 _c64_cpu_port_out(0x37,&machine);assert(std::string(rr_memory(3,0xe000)).find("\"writer\":null")!=std::string::npos);
 _c64_cpu_port_out(0x35,&machine);assert(std::string(rr_memory(3,0xe000)).find("error")!=std::string::npos);
 // Invalid TAP does not replace a valid tape. Extended values must be bounded.
 memcpy(input,"C64-TAPE-RAW",12);input[12]=1;input[13]=input[14]=input[15]=0;input[16]=1;input[17]=input[18]=input[19]=0;input[20]=0;assert(!rr_tape(21));
 puts("PASS: CIA FLAG/ICR, timer discrimination/reload, banking, raster IRQ, badline stalls, device/audio checkpoint, live charset/sprite fetches, SID output, malformed TAP");
}
