#include "../browser/core.cpp"
#include <cassert>
#include <iostream>
static void setup(){
 std::fill_n(rr_input(),20480,0);assert(rr_init(8192,8192,4096));
 std::fill_n(rr_input(),24,0);std::copy_n("C64-TAPE-RAW",12,rr_input());rr_input()[12]=1;rr_input()[16]=4;rr_input()[21]=255;rr_input()[22]=255;assert(rr_tape(24));
 std::fill_n(rr_input(),65536,0xea);const uint8_t loop[]={0xee,0,2,0x4c,0,8};std::copy(std::begin(loop),std::end(loop),rr_input()+0x800);rr_input()[0x200]=0;assert(rr_prepare(0x800,0));
}
static void bind(){for(unsigned i=0;i<228;i++)rr_input()[i]=uint8_t(i+1);assert(rr_state_bind());}
static std::vector<uint8_t> saved(){const auto n=rr_state_save();assert(n);return {rr_state_data(),rr_state_data()+n};}
static bool load(const std::vector<uint8_t>& bytes){auto* p=rr_state_input(bytes.size());assert(p);std::copy(bytes.begin(),bytes.end(),p);return rr_state_load(bytes.size());}
static unsigned pc(){return board.state.cpu.pc;}
int main(){
 setup();const auto original=saveState(board,identity);const auto cycle=rr_cycle();
 assert(std::string(rr_debug_snapshot(0xdc00)).find("\"bytes\":[null")!=std::string::npos);assert(rr_cycle()==cycle&&saveState(board,identity)==original);
 assert(rr_debug_begin(2,0x800,0)&&rr_debug_run(100)==4&&rr_cycle()==0);
 assert(rr_debug_begin(1,0,0)&&rr_debug_run(100)==2&&pc()==0x803&&rr_ram()[0x200]==1&&rr_cycle()==6);
 assert(std::string(rr_debug_snapshot(-1)).find("\"lastExecutedPC\":2048")!=std::string::npos);
 assert(rr_debug_begin(2,0x800,0)&&rr_debug_run(100)==4);
 assert(rr_debug_begin(2,0x800,1)&&rr_debug_run(1)==0&&rr_debug_run(100)==4&&rr_ram()[0x200]==2);
 setup();assert(rr_run(1,0,0)==1);assert(!rr_debug_begin(1,0,0));assert(rr_debug_begin(0,0,0)&&rr_debug_run(100)==1&&pc()==0x803);
 // IRQ/NMI entry is a distinct stop, with no instruction counted as retired.
 for(bool nmi:{false,true}){
  setup();ctx.synthetic=false;board.kernal[nmi?0x1ffa:0x1ffe]=0;board.kernal[nmi?0x1ffb:0x1fff]=9;
  board.state.cpu.interruptPending=true;board.state.cpu.nmiPending=nmi;const auto retired=board.state.cpu.retired;
  assert(rr_debug_begin(1,0,0)&&rr_debug_run(100)==3&&pc()==0x900&&board.state.cpu.retired==retired&&rr_ram()[0x200]==0);
 }
 // JSR/RTS, nesting, an interrupt with its own subroutine, and bad return SP.
 const uint8_t caller[]={0x20,0,9,0xea},callee[]={0x20,0,10,0x60};
 setup();std::copy(std::begin(caller),std::end(caller),rr_ram()+0x800);std::copy(std::begin(callee),std::end(callee),rr_ram()+0x900);rr_ram()[0xa00]=0x60;
 assert(rr_debug_begin(3,0,0)&&rr_debug_run(100)==12&&pc()==0x803&&board.state.cpu.s==0xfd);
 assert(rr_debug_begin(3,0,0)&&rr_debug_run(100)==2&&pc()==0x804);
 setup();std::copy(std::begin(caller),std::end(caller),rr_ram()+0x800);rr_ram()[0x900]=0x60;
 assert(rr_debug_begin(1,0,0)&&rr_debug_run(100)==2&&pc()==0x900);
 assert(rr_debug_begin(4,0,0)&&rr_debug_run(100)==12&&pc()==0x803);assert(!rr_debug_begin(4,0,0));
 setup();ctx.synthetic=false;std::copy(std::begin(caller),std::end(caller),rr_ram()+0x800);rr_ram()[0x900]=0xea;rr_ram()[0x901]=0x60;
 assert(rr_debug_begin(3,0,0)&&rr_debug_run(6)==0&&pc()==0x900);board.state.cpu.interruptPending=true;board.kernal[0x1ffe]=0;board.kernal[0x1fff]=10;
 const uint8_t handler[]={0x20,0,11,0x40};std::copy(std::begin(handler),std::end(handler),rr_ram()+0xa00);rr_ram()[0xb00]=0x60;
 assert(rr_debug_run(200)==12&&pc()==0x803&&board.state.cpu.s==0xfd);
 setup();std::copy(std::begin(caller),std::end(caller),rr_ram()+0x800);rr_ram()[0x900]=0x60;assert(rr_debug_begin(3,0,0)&&rr_debug_run(6)==0);rr_ram()[0x1fc]=0x10;assert(rr_debug_run(100)==14);
 // Actual RMW dummy and final stores, writer filtering and ROM-under-RAM writes.
 setup();assert(!rr_debug_watch(0xdc00,0,0,-1)&&!rr_debug_watch(0x200,2,1,-1));
 assert(rr_debug_watch(0x200,0,255,0x800)&&rr_debug_run(100)==13&&!boundary());assert(rr_ram()[0x200]==0);
 assert(std::string(rr_debug_event()).find("\"pc\":2048")!=std::string::npos);
 assert(rr_debug_watch(0x200,1,255,0x800)&&rr_debug_run(100)==13&&rr_ram()[0x200]==1);
 assert(rr_debug_begin(0,0,0)&&rr_debug_run(100)==1);assert(rr_debug_watch(0x200,0,0,0x900)&&rr_debug_run(100)==0);
 // Mapping-bound targets and safe live disassembly after self-modification.
 setup();const uint8_t bankout[]={0xa9,0x34,0x85,1,0x4c,0,0xa0};std::copy(std::begin(bankout),std::end(bankout),rr_ram()+0x800);
 assert(rr_debug_begin(2,0xa000,0)&&rr_debug_run(100)==5&&pc()==0xa000);
 setup();rr_ram()[0xa000]=0x55;board.basic[0]=0xaa;assert(std::string(rr_debug_snapshot(0xa000)).find("\"bytes\":[170")!=std::string::npos);board.state.data=0x34;assert(std::string(rr_debug_snapshot(0xa000)).find("\"bytes\":[85")!=std::string::npos);
 setup();const uint8_t modify[]={0xa9,0x55,0x8d,0x10,8,0x4c,0x10,8};std::copy(std::begin(modify),std::end(modify),rr_ram()+0x800);assert(rr_debug_begin(2,0x810,0)&&rr_debug_run(100)==4&&((rr_bus()>>16)&255)==0x55);
 // RDY-held opcode fetch is not mistaken for an instruction or next match.
 setup();board.state.vic.raster=51;board.state.vic.cycle=11;board.state.vic.den=true;board.state.vic.regs[0x11]=0x13;
 assert(rr_debug_begin(1,0,0)&&rr_debug_run(43)==0&&rr_ram()[0x200]==0);assert(rr_debug_run(100)==2&&rr_ram()[0x200]==1);
 setup();rr_ram()[0x800]=2;assert(rr_debug_begin(1,0,0)&&rr_debug_run(100)==6);
 // Whole-batch edit validation, including the next opcode and duplicate targets.
 setup();const uint8_t edits[]={0,8,0xee,0xce,1,2,0xff,9};std::copy(std::begin(edits),std::end(edits),rr_input());assert(!rr_debug_edit(2)&&rr_ram()[0x800]==0xee);
 rr_input()[6]=rr_ram()[0x201];assert(rr_debug_edit(2));assert(rr_debug_begin(1,0,0)&&rr_debug_run(100)==2&&rr_ram()[0x200]==255);
 setup();rr_input()[0]=0;rr_input()[1]=0xdc;assert(!rr_debug_edit(1));assert(!rr_debug_edit(257));rr_run(1,0,0);assert(!rr_debug_edit(1));
 // Exported memory protocol observes real reads/writes, including writes under
 // visible ROM. Inspection itself never advances the machine.
 setup();rr_inspect_regions();rr_activity_begin(7);assert(rr_run(30,0,0)==30);rr_activity_end();bool dummy=false,write=false,fetch=false;
 for(const auto& e:rrmem::events){if(e.offset==0x200&&e.kind==2){dummy|=e.value==0;write|=e.value==1;}fetch|=e.kind==4;}assert(dummy&&write&&fetch);
 const auto count=rr_activity_count();const auto before=rr_cycle();rr_inspect_regions();rr_inspect_data(0);assert(rr_activity_count()==count&&rr_cycle()==before);
 setup();board.pulses={1};board.state.tape.pulse=0;board.state.tape.remaining=1;board.state.data=0x17;rr_play(1);rr_activity_begin(1);assert(rr_run(1,0,0)==1);assert(board.state.tape.pulse==1);assert(rrmem::events.back().region==5&&rrmem::events.back().offset==0&&rrmem::events.back().value==1);rr_activity_end(); // EOF advances the cursor; activity identifies the last consumed pulse.
 setup();const uint8_t under[]={0xa9,0x42,0x8d,0,0xa0};std::copy(std::begin(under),std::end(under),rr_ram()+0x800);rr_inspect_regions();rr_activity_begin(3);assert(rr_run(6,0,0)==6);assert(rrmem::events.back().region==0&&rrmem::events.back().offset==0xa000&&rrmem::events.back().value==0x42);rr_activity_end();
 // Keyboard mapping is switch-based, including simultaneous shifted keys.
 setup();rr_key('"',1);rr_key('!',1);assert(board.state.keys[1]&128);rr_key('"',0);assert(board.state.keys[1]&128);rr_key('!',0);assert(!(board.state.keys[1]&128));rr_key(255,1);assert(board.state.restore);rr_key(255,0);assert(!board.state.restore);
 // Slots and portable transport preserve partial execution, inputs and drive
 // mode; media replacement invalidates bindings, and failed loads are atomic.
 setup();assert(!rr_state_save());bind();rr_key('L',1);rr_run(2,0,0);const auto partial=saved();assert(rr_checkpoint(0));rr_run(50,0,0);const auto end=saved();assert(load(partial));rr_run(50,0,0);assert(saved()==end);
 assert(!rr_restore(0));assert(load(partial));assert(rr_checkpoint(0));rr_run(20,0,0);assert(rr_restore(0)&&saved()==partial);rr_key('L',0);assert(!(board.state.keys[5]&4));
 auto broken=partial;broken[30]^=1;const auto live=saved();assert(!load(broken)&&saved()==live);assert(!rr_state_input(0xffffffff));
 bind();rr_input()[0]^=1;assert(rr_state_bind());const auto changed=saved();assert(!load(partial)&&saved()==changed);
 // Observer IDs and trace settings cannot change deterministic machine bytes.
 setup();bind();const auto unobserved=saved();rr_trace(1,0,65535);rr_run(100,0,0);assert(!events.empty());const auto observed=saved();assert(load(unobserved)&&!tracing);rr_run(100,0,0);assert(saved()==observed);
 setup();std::fill_n(rr_input(),16384,0xea);rr_input()[0]=0x4c;rr_input()[1]=0;rr_input()[2]=0xc0;rr_input()[0x3ffc]=0;rr_input()[0x3ffd]=0xc0;assert(rr_drive_rom(16384));bind();assert(rr_run(100,0,0)==100&&machine.drive.state.clocks>board.state.cycles);
 const auto driveBefore=saveState(machine,identity);rr_drive_debug_snapshot(0x1800);assert(saveState(machine,identity)==driveBefore);
 assert(rr_drive_debug_begin(0,0,0));int dr=0;while(!dr)dr=rr_drive_debug_run(100);assert(dr==1);
 const auto retired=machine.drive.state.cpu.retired;assert(rr_drive_debug_begin(1,0,0));dr=0;while(!dr)dr=rr_drive_debug_run(100);assert(dr==2&&machine.drive.state.cpu.retired==retired+1);
 assert(rr_drive_debug_begin(2,0xc000,1));dr=0;while(!dr)dr=rr_drive_debug_run(100);assert(dr==4&&machine.drive.state.cpu.pc==0xc000);
 const auto dual=saved();rr_run(100,0,0);const auto dualEnd=saved();assert(load(dual));assert(rr_drive_debug_run(10)==-1);rr_run(100,0,0);assert(saved()==dualEnd);assert(!rr_drive_rom(16384));
 setup();bind();const auto noDrive=saved();assert(!load(dual)&&saved()==noDrive);
 std::cout<<"PASS owned browser ABI: execution, IRQ/partial/RDY stepping, over/out, actual-write watches, atomic edits, banking, memory activity, input, slots and bound portable state\n";
}
