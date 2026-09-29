#include "../../platform/dos/browser/core/api.cpp"
#include <cassert>
#include <iostream>
int main(){auto b=Slice<uint8_t>::make(128);b[0]='M';b[1]='Z';b[2]=128;b[4]=1;b[8]=4;b[16]=0xfe;b[17]=0xff;b[64]=0xeb;b[65]=0xfe;rrfiles::overlay["test.exe"]=b;rrBoot("test.exe",false);auto*m=realMachine;auto*c=m->CPU;x86_CPU_SetReg16(c,0,0x13);assert(dos_Machine_handleInt(m,c,0x10));
 assert(rr_capture_begin());dos_Machine_portOut(m,0x3c8,1,1);for(int val:{63,0,0})dos_Machine_portOut(m,0x3c9,1,val);for(int i=0;i<100;i++)dos_Machine_Write(m,0xa0000+i,1);assert(rr_capture_end());auto f=rrFrame();assert(f[0]==252&&f[1]==0);assert(rrcapture::trace.shadow==rrcapture::trace.final);assert(std::string(rr_pixel(0,0)).find("complete\":true")!=std::string::npos);rr_replay_begin();while(!rr_replay_seek(rrreplay::replay.steps.size())){}assert(!memcmp(f.data(),rr_replay_frame(),f.size()));while(!rr_replay_seek(0)){}assert(memcmp(f.data(),rr_replay_frame(),f.size()));
 auto[handle,error]=go_os_Create("saves/TEST.SAV");assert(!error);auto payload=Slice<uint8_t>{1,2,3,4};auto[n,err]=os_File_Write(handle,payload);assert(!err&&n==4);m->files[7]=handle;
 auto proof=std::string(rr_proof());auto size=rr_state_save();assert(size);auto saved=stateOutput;rrfiles::overlay["saves/test.sav"][0]=99;c->Regs[0]=123;memcpy(rr_state_input(size),saved.data(),size);assert(rr_state_load(size));assert(proof==rr_proof());assert(rrfiles::overlay["saves/test.sav"][0]==1);assert(realMachine->files[7]->offset==4);memcpy(rr_state_input(size-1),saved.data(),size-1);assert(!rr_state_load(size-1));assert(proof==rr_proof());
 rr_key(0x48,1);auto count=realMachine->keyEvents.n;rr_pad(1);rr_pad(0);assert(realMachine->keyEvents.n==count);rr_key(0x48,0);assert(realMachine->keyEvents.n==count+1);
 std::cout<<"DOS MZ boot, VGA/palette provenance, replay, file-overlay restore and combined inputs pass\n";
}
