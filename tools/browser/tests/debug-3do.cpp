#include "../../platform/threedo/browser/core/api.cpp"
#include <cassert>
int main(){machine=threedo_NewMachine();machine->PaceFields=true;machine->vol=arenaNew(threedo_Volume{});machine->vol->stride=2048;machine->vol->nsect=1;auto*c=machine->CPU;c->R[15]=0x1000;threedo_Machine_write32(machine,0x1000,0xe2800001);threedo_Machine_write32(machine,0x1004,0xeafffffd);auto old=proof(machine);auto s=std::string(rr_debug_snapshot(-1));assert(s.find("ADD r0, r0, #1")!=std::string::npos&&proof(machine)==old);assert(std::string(rr_debug_snapshot(0x300000)).find("null")!=std::string::npos);
 assert(rr_debug_begin(2,0x1000,0)&&rr_debug_run(10)==4&&c->Instrs==0&&runContext.prepared);assert(rr_debug_begin(1,0,0)&&rr_debug_run(1)==2&&c->R[0]==1&&c->Instrs==1);assert(rr_debug_begin(2,0x1000,0)&&rr_debug_run(10)==4);assert(rr_debug_begin(2,0x1000,1)&&rr_debug_run(10)==4&&c->R[0]==2);
 // Exact state resumes after the same scheduler prelude, not a duplicate tick.
 auto n=rr_state_save();assert(n);auto saved=stateOutput;assert(runContext.prepared);assert(rr_debug_begin(1,0,0)&&rr_debug_run(1)==2);auto after=proof(machine);memcpy(rr_state_input(n),saved.data(),n);assert(rr_state_load(n)&&runContext.prepared);assert(rr_debug_begin(1,0,0)&&rr_debug_run(1)==2&&proof(machine)==after);
 // Write fast paths cannot bypass selected CPU bus activity; physical VRAM stays region 1.
 c=machine->CPU;rr_inspect_regions();rr_activity_clear_selection();assert(rr_activity_select(0x200100,4));rr_activity_begin(2);arm60_CPU_write32(c,0x200100,0x12345678);arm60_CPU_write32(c,0x100,9);rr_activity_end();assert(rr_activity_count()==4);assert(rrmem::events[0].region==1&&rrmem::events[0].offset==0x100&&rrmem::events[0].value==0x12);rr_activity_clear_selection();assert(rr_ram()[0x200100]==0x12);
 // Conditional ARM words are decoded live; unsupported 26-bit/Thumb modes are not advertised.
 machine->dram[0x1000]=0x12;s=rr_debug_snapshot(0x1000);assert(s.find("ADDNE")!=std::string::npos);assert(!rr_debug_begin(2,0x1001,0));
 // Virtual Portfolio services consume a scheduler unit without a guest opcode.
 c->R[15]=threedo_hleBase;c->R[14]=0x1000;auto retired=c->Instrs;
 assert(std::string(rr_debug_snapshot(-1)).find("Portfolio HLE entry")!=std::string::npos);
 assert(rr_debug_begin(1,c->R[15],0)&&rr_debug_run(1)==8&&c->Instrs==retired&&c->R[15]==0x1000);
 // SWI is one retired ARM instruction with an atomic HLE service.
 threedo_Machine_write32(machine,0x1000,0xef000000);c->SWI=[](arm60_CPU*,uint32_t){return true;};
 assert(rr_debug_begin(1,0,0)&&rr_debug_run(1)==10&&c->Instrs==retired+1);
 c->R[15]=0x1000;assert(rr_debug_begin(2,0x1000,0));rr_debug_bind(rrTask()+1);assert(rr_debug_run(1)!=4);
 rr_activity_end();rr_activity_begin(2);rr_activity_begin(2);rr_activity_end();rr_activity_end();assert(!machine->OnWrite);
 std::cout<<"PASS ARM60 live decoding, safe backing snapshots, instruction/target boundaries, pending-prelude state restore and selected bus writes\n";}
