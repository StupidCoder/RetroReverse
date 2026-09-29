#include "runtime.h"
struct x86_CPU;
struct x86_ea;
struct x86_FPUState;
struct dos_COFFSection;
struct dos_COFF;
struct dos_dos_432_kv;
struct dos_findState;
struct dos_Machine;
struct dos_VGAWriter;
struct dos_emsHandle;
struct dos_emsState;
struct dos_ioState;
struct dos_injEvent;
struct dos_mouseState;
struct dos_vgaState;
struct dos_PM;
struct dos_MemRegion;
struct dos_pmVector;
struct dos_pitState;
struct dos_rmcs;
struct dos_MZ;
struct dos_Reloc;
struct Anon0;
struct Anon1;
struct Anon10;
struct Anon11;
struct Anon12;
struct Anon13;
struct Anon14;
struct Anon15;
struct Anon16;
struct Anon17;
struct Anon18;
struct Anon19;
struct Anon2;
struct Anon20;
struct Anon21;
struct Anon3;
struct Anon4;
struct Anon5;
struct Anon6;
struct Anon7;
struct Anon8;
struct Anon9;
struct x86_FPUState{
std::array<double,8> St{};
std::array<uint8_t,8> Tag{};
int64_t Top{};
uint16_t Ctrl{};
uint16_t Stat{};
};
struct x86_CPU{
std::array<uint32_t,8> Regs{};
std::array<uint16_t,8> Seg{};
uint32_t IP{};
uint32_t instrIP{};
bool CF{};
bool PF{};
bool AF{};
bool ZF{};
bool SF{};
bool TF{};
bool IF{};
bool DF{};
bool OF{};
bool Halted{};
std::string HaltReason{};
uint64_t Steps{};
uint64_t Ext386{};
int64_t Mode{};
std::array<uint32_t,8> SegBase{};
std::function<uint32_t(uint16_t)> SegResolve{};
std::function<bool(x86_CPU*,uint8_t)> IntHook{};
std::function<void(x86_CPU*)> OnStep{};
std::function<uint32_t(uint16_t,int64_t)> PortIn{};
std::function<void(uint16_t,int64_t,uint32_t)> PortOut{};
RRX86Bus bus{};
bool ssShadow{};
x86_FPUState FPU{};
std::array<std::array<uint8_t,16>,8> XMM{};
std::array<std::array<uint8_t,8>,8> MMX{};
uint32_t MXCSR{};
uint64_t TSCMul{};
std::function<uint64_t()> TSCFunc{};
int64_t dSeg{};
int64_t dOpsize{};
int64_t dAddrsize{};
};
struct x86_ea{
bool isReg{};
uint8_t reg{};
uint32_t base{};
uint32_t off{};
};
using x86_sseKind=int64_t;
struct dos_COFFSection{
std::string Name{};
uint32_t VAddr{};
uint32_t Size{};
uint32_t FileOff{};
uint32_t Flags{};
Slice<uint8_t> Data{};
};
struct dos_COFF{
int64_t StubEnd{};
int64_t NSections{};
uint16_t Flags{};
uint32_t Entry{};
uint32_t TextStart{};
uint32_t DataStart{};
uint32_t TextSize{};
uint32_t DataSize{};
uint32_t BSSSize{};
Slice<dos_COFFSection> Sections{};
};
struct dos_dos_432_kv{
uint8_t ah{};
int64_t n{};
};
struct dos_findState{
Slice<std::string> matches{};
int64_t idx{};
};
using dos_injKind=uint8_t;
struct dos_injEvent{
dos_injKind kind{};
uint8_t code{};
int64_t x{};
int64_t y{};
int64_t delay{};
};
struct dos_Machine{
Slice<uint8_t> Mem{};
x86_CPU* CPU{};
std::string gameDir{};
Map<uint16_t,os_File*> files{};
uint16_t pspSeg{};
uint16_t envSeg{};
uint16_t loadSeg{};
uint16_t memTop{};
uint16_t firstMCB{};
uint16_t dtaSeg{};
uint16_t dtaOff{};
Map<uint32_t,dos_findState*> finds{};
std::string scratchDir{};
Slice<std::string> Log{};
Map<uint8_t,int64_t> IntCounts{};
Map<uint8_t,int64_t> otherInts{};
int64_t OverlayCalls{};
dos_emsState* ems{};
dos_ioState* io{};
Map<uint16_t,int64_t> video{};
dos_vgaState* vga{};
dos_mouseState* ms{};
Slice<dos_injEvent> keyEvents{};
int64_t keyWait{};
int64_t keyHits{};
bool keyRetry{};
bool EnableIRQ{};
uint32_t WatchAddr{};
uint32_t WatchLen{};
int64_t watchHits{};
uint64_t VGAProfileAt{};
uint32_t ProfLo{};
uint32_t ProfHi{};
Map<uint32_t,int64_t> vgaProfile{};
uint64_t RdProfileAt{};
uint32_t RdLo{};
uint32_t RdHi{};
Map<uint32_t,int64_t> rdProfile{};
bool Terminated{};
uint8_t ExitCode{};
Map<uint16_t,int64_t> Int33Hist{};
};
struct dos_VGAWriter{
uint16_t Seg{};
uint16_t Off{};
int64_t Count{};
};
struct dos_emsHandle{
int64_t base{};
int64_t count{};
};
struct dos_emsState{
Slice<uint8_t> backing{};
int64_t nextPage{};
Map<uint16_t,dos_emsHandle> handles{};
uint16_t nextH{};
std::array<int64_t,4> slot{};
Map<uint16_t,std::array<int64_t,4>> saved{};
};
struct dos_ioState{
uint8_t kbdOut{};
bool kbdOutFull{};
bool expectData{};
bool retrace{};
uint16_t pit{};
uint8_t oplReg{};
bool oplTimer{};
uint32_t tick{};
uint8_t mixReg{};
std::array<uint8_t,256> mixRegs{};
Slice<uint8_t> dspQueue{};
uint8_t dspReset{};
int64_t dacIndex{};
std::array<uint8_t,768> Pal{};
Map<uint16_t,bool> seen{};
};
struct dos_mouseState{
int64_t x{};
int64_t y{};
int64_t accX{};
int64_t accY{};
uint8_t buttons{};
int64_t xMin{};
int64_t xMax{};
int64_t yMin{};
int64_t yMax{};
bool hidden{};
int64_t pressL{};
int64_t pressR{};
};
struct dos_vgaState{
std::array<std::array<uint8_t,65536>,4> planes{};
std::array<uint8_t,4> latch{};
uint8_t seqIdx{};
std::array<uint8_t,8> seq{};
uint8_t gcIdx{};
std::array<uint8_t,16> gc{};
uint8_t crtcIdx{};
std::array<uint8_t,32> crtc{};
};
struct dos_pitState{
uint16_t reload{};
bool writeHi{};
uint16_t latched{};
bool haveLatch{};
bool readHi{};
};
struct dos_pmVector{
uint16_t sel{};
uint32_t off{};
bool set{};
};
struct dos_PM{
Slice<uint8_t> Mem{};
x86_CPU* CPU{};
std::string gameDir{};
Map<uint16_t,os_File*> files{};
uint32_t convBase{};
uint32_t convNext{};
uint32_t convTop{};
uint32_t heapBase{};
uint32_t heapNext{};
uint32_t stackFloor{};
uint32_t infoBase{};
uint32_t imgEnd{};
uint32_t wWLo{};
uint32_t wWHi{};
std::function<void(uint32_t,uint32_t,uint32_t)> onW{};
uint32_t wRLo{};
uint32_t wRHi{};
std::function<void(uint32_t,uint32_t,uint32_t)> onR{};
uint16_t nextSel{};
uint16_t nextCallback{};
Map<uint16_t,uint32_t> sels{};
bool virtIF{};
uint16_t lolSeg{};
uint16_t lolOff{};
uint16_t dtaSeg{};
uint16_t dtaOff{};
dos_pitState pit{};
bool retrace{};
std::array<dos_pmVector,256> pmVectors{};
dos_pmVector defIntVec{};
Slice<dos_injEvent> keyEvents{};
Slice<uint8_t> injKeys{};
int64_t keyWait{};
bool keyRetry{};
int64_t injTick{};
int64_t keyHits{};
uint8_t kbdData{};
bool kbdFull{};
int64_t dacIndex{};
std::array<uint8_t,768> Pal{};
Slice<std::string> Log{};
Map<uint16_t,int64_t> DPMICounts{};
Map<uint8_t,int64_t> IntCounts{};
Map<uint8_t,int64_t> DOSCounts{};
bool Terminated{};
uint8_t ExitCode{};
Slice<uint8_t> Console{};
bool rrQuakeBase{};
};
struct dos_MemRegion{
std::string Name{};
uint32_t Lo{};
uint32_t Hi{};
};
struct dos_rmcs{
uint32_t addr{};
uint32_t edi{};
uint32_t esi{};
uint32_t ebp{};
uint32_t ebx{};
uint32_t edx{};
uint32_t ecx{};
uint32_t eax{};
uint16_t flags{};
uint16_t es{};
uint16_t ds{};
uint16_t fs{};
uint16_t gs{};
uint16_t ip{};
uint16_t cs{};
uint16_t sp{};
uint16_t ss{};
};
struct dos_Reloc{
uint16_t Segment{};
uint16_t Offset{};
};
struct dos_MZ{
uint16_t LastPageBytes{};
uint16_t Pages{};
uint16_t Relocations{};
uint16_t HeaderParas{};
uint16_t MinAlloc{};
uint16_t MaxAlloc{};
uint16_t InitSS{};
uint16_t InitSP{};
uint16_t Checksum{};
uint16_t InitIP{};
uint16_t InitCS{};
uint16_t RelocOffset{};
uint16_t OverlayNumber{};
int64_t FileSize{};
int64_t LoadImageEnd{};
int64_t LoadModuleOffset{};
int64_t LoadModuleSize{};
int64_t AppendedSize{};
Slice<dos_Reloc> Relocs{};
};
struct Anon0{std::array<uint32_t,8> Regs{};std::array<uint16_t,8> Seg{};uint32_t IP{};uint32_t instrIP{};bool CF{};bool PF{};bool AF{};bool ZF{};bool SF{};bool TF{};bool IF{};bool DF{};bool OF{};bool Halted{};std::string HaltReason{};uint64_t Steps{};uint64_t Ext386{};int64_t Mode{};std::array<uint32_t,8> SegBase{};std::function<uint32_t(uint16_t)> SegResolve{};std::function<bool(x86_CPU*,uint8_t)> IntHook{};std::function<void(x86_CPU*)> OnStep{};std::function<uint32_t(uint16_t,int64_t)> PortIn{};std::function<void(uint16_t,int64_t,uint32_t)> PortOut{};RRX86Bus bus{};bool ssShadow{};x86_FPUState FPU{};std::array<std::array<uint8_t,16>,8> XMM{};std::array<std::array<uint8_t,8>,8> MMX{};uint32_t MXCSR{};uint64_t TSCMul{};std::function<uint64_t()> TSCFunc{};int64_t dSeg{};int64_t dOpsize{};int64_t dAddrsize{};};
struct Anon1{bool isReg{};uint8_t reg{};uint32_t base{};uint32_t off{};};
struct Anon10{Slice<uint8_t> backing{};int64_t nextPage{};Map<uint16_t,dos_emsHandle> handles{};uint16_t nextH{};std::array<int64_t,4> slot{};Map<uint16_t,std::array<int64_t,4>> saved{};};
struct Anon11{uint8_t kbdOut{};bool kbdOutFull{};bool expectData{};bool retrace{};uint16_t pit{};uint8_t oplReg{};bool oplTimer{};uint32_t tick{};uint8_t mixReg{};std::array<uint8_t,256> mixRegs{};Slice<uint8_t> dspQueue{};uint8_t dspReset{};int64_t dacIndex{};std::array<uint8_t,768> Pal{};Map<uint16_t,bool> seen{};};
struct Anon12{dos_injKind kind{};uint8_t code{};int64_t x{};int64_t y{};int64_t delay{};};
struct Anon13{int64_t x{};int64_t y{};int64_t accX{};int64_t accY{};uint8_t buttons{};int64_t xMin{};int64_t xMax{};int64_t yMin{};int64_t yMax{};bool hidden{};int64_t pressL{};int64_t pressR{};};
struct Anon14{std::array<std::array<uint8_t,65536>,4> planes{};std::array<uint8_t,4> latch{};uint8_t seqIdx{};std::array<uint8_t,8> seq{};uint8_t gcIdx{};std::array<uint8_t,16> gc{};uint8_t crtcIdx{};std::array<uint8_t,32> crtc{};};
struct Anon15{Slice<uint8_t> Mem{};x86_CPU* CPU{};std::string gameDir{};Map<uint16_t,os_File*> files{};uint32_t convBase{};uint32_t convNext{};uint32_t convTop{};uint32_t heapBase{};uint32_t heapNext{};uint32_t stackFloor{};uint32_t infoBase{};uint32_t imgEnd{};uint32_t wWLo{};uint32_t wWHi{};std::function<void(uint32_t,uint32_t,uint32_t)> onW{};uint32_t wRLo{};uint32_t wRHi{};std::function<void(uint32_t,uint32_t,uint32_t)> onR{};uint16_t nextSel{};uint16_t nextCallback{};Map<uint16_t,uint32_t> sels{};bool virtIF{};uint16_t lolSeg{};uint16_t lolOff{};uint16_t dtaSeg{};uint16_t dtaOff{};dos_pitState pit{};bool retrace{};std::array<dos_pmVector,256> pmVectors{};dos_pmVector defIntVec{};Slice<dos_injEvent> keyEvents{};Slice<uint8_t> injKeys{};int64_t keyWait{};bool keyRetry{};int64_t injTick{};int64_t keyHits{};uint8_t kbdData{};bool kbdFull{};int64_t dacIndex{};std::array<uint8_t,768> Pal{};Slice<std::string> Log{};Map<uint16_t,int64_t> DPMICounts{};Map<uint8_t,int64_t> IntCounts{};Map<uint8_t,int64_t> DOSCounts{};bool Terminated{};uint8_t ExitCode{};Slice<uint8_t> Console{};};
struct Anon16{std::string Name{};uint32_t Lo{};uint32_t Hi{};};
struct Anon17{uint16_t sel{};uint32_t off{};bool set{};};
struct Anon18{uint16_t reload{};bool writeHi{};uint16_t latched{};bool haveLatch{};bool readHi{};};
struct Anon19{uint32_t addr{};uint32_t edi{};uint32_t esi{};uint32_t ebp{};uint32_t ebx{};uint32_t edx{};uint32_t ecx{};uint32_t eax{};uint16_t flags{};uint16_t es{};uint16_t ds{};uint16_t fs{};uint16_t gs{};uint16_t ip{};uint16_t cs{};uint16_t sp{};uint16_t ss{};};
struct Anon2{std::array<double,8> St{};std::array<uint8_t,8> Tag{};int64_t Top{};uint16_t Ctrl{};uint16_t Stat{};};
struct Anon20{uint16_t LastPageBytes{};uint16_t Pages{};uint16_t Relocations{};uint16_t HeaderParas{};uint16_t MinAlloc{};uint16_t MaxAlloc{};uint16_t InitSS{};uint16_t InitSP{};uint16_t Checksum{};uint16_t InitIP{};uint16_t InitCS{};uint16_t RelocOffset{};uint16_t OverlayNumber{};int64_t FileSize{};int64_t LoadImageEnd{};int64_t LoadModuleOffset{};int64_t LoadModuleSize{};int64_t AppendedSize{};Slice<dos_Reloc> Relocs{};};
struct Anon21{uint16_t Segment{};uint16_t Offset{};};
struct Anon3{std::string Name{};uint32_t VAddr{};uint32_t Size{};uint32_t FileOff{};uint32_t Flags{};Slice<uint8_t> Data{};};
struct Anon4{int64_t StubEnd{};int64_t NSections{};uint16_t Flags{};uint32_t Entry{};uint32_t TextStart{};uint32_t DataStart{};uint32_t TextSize{};uint32_t DataSize{};uint32_t BSSSize{};Slice<dos_COFFSection> Sections{};};
struct Anon5{Slice<std::string> matches{};int64_t idx{};};
struct Anon6{Slice<uint8_t> Mem{};x86_CPU* CPU{};std::string gameDir{};Map<uint16_t,os_File*> files{};uint16_t pspSeg{};uint16_t envSeg{};uint16_t loadSeg{};uint16_t memTop{};uint16_t firstMCB{};uint16_t dtaSeg{};uint16_t dtaOff{};Map<uint32_t,dos_findState*> finds{};std::string scratchDir{};Slice<std::string> Log{};Map<uint8_t,int64_t> IntCounts{};Map<uint8_t,int64_t> otherInts{};int64_t OverlayCalls{};dos_emsState* ems{};dos_ioState* io{};Map<uint16_t,int64_t> video{};dos_vgaState* vga{};dos_mouseState* ms{};Slice<dos_injEvent> keyEvents{};int64_t keyWait{};int64_t keyHits{};bool keyRetry{};bool EnableIRQ{};uint32_t WatchAddr{};uint32_t WatchLen{};int64_t watchHits{};uint64_t VGAProfileAt{};uint32_t ProfLo{};uint32_t ProfHi{};Map<uint32_t,int64_t> vgaProfile{};uint64_t RdProfileAt{};uint32_t RdLo{};uint32_t RdHi{};Map<uint32_t,int64_t> rdProfile{};bool Terminated{};uint8_t ExitCode{};Map<uint16_t,int64_t> Int33Hist{};};
struct Anon7{uint16_t Seg{};uint16_t Off{};int64_t Count{};};
struct Anon8{uint8_t ah{};int64_t n{};};
struct Anon9{int64_t base{};int64_t count{};};

#include "adapters-decl.h"
x86_CPU* x86_NewCPU(RRX86Bus bus);
uint32_t x86_CPU_LinearPC(x86_CPU* c);
std::string x86_CPU_at(x86_CPU* c);
uint32_t x86_CPU_segBase(x86_CPU* c,int64_t idx);
uint32_t x86_CPU_linear(x86_CPU* c,uint32_t base,uint32_t off);
uint32_t x86_CPU_ipMask(x86_CPU* c,uint32_t ip);
bool x86_CPU_stack32(x86_CPU* c);
uint32_t x86_CPU_rd8(x86_CPU* c,uint32_t a);
uint32_t x86_CPU_rd16(x86_CPU* c,uint32_t a);
uint32_t x86_CPU_rd32(x86_CPU* c,uint32_t a);
void x86_CPU_wr8(x86_CPU* c,uint32_t a,uint32_t v);
void x86_CPU_wr16(x86_CPU* c,uint32_t a,uint32_t v);
void x86_CPU_wr32(x86_CPU* c,uint32_t a,uint32_t v);
uint32_t x86_CPU_memRead(x86_CPU* c,uint32_t base,uint32_t off,int64_t b);
uint32_t x86_CPU_memRead_reference(x86_CPU* c,uint32_t base,uint32_t off,int64_t b);
void x86_CPU_memWrite(x86_CPU* c,uint32_t base,uint32_t off,int64_t b,uint32_t v);
void x86_CPU_memWrite_reference(x86_CPU* c,uint32_t base,uint32_t off,int64_t b,uint32_t v);
uint32_t x86_CPU_fetch8(x86_CPU* c);
uint32_t x86_CPU_fetch16(x86_CPU* c);
uint32_t x86_CPU_fetch16_reference(x86_CPU* c);
uint32_t x86_CPU_fetch32(x86_CPU* c);
uint32_t x86_CPU_fetch32_reference(x86_CPU* c);
uint32_t x86_CPU_fetchImm(x86_CPU* c);
void x86_CPU_loadSeg(x86_CPU* c,int64_t idx,uint16_t sel);
uint32_t x86_CPU_g8(x86_CPU* c,uint8_t i);
void x86_CPU_s8(x86_CPU* c,uint8_t i,uint32_t v);
uint32_t x86_CPU_g16(x86_CPU* c,uint8_t i);
void x86_CPU_s16(x86_CPU* c,uint8_t i,uint32_t v);
uint32_t x86_CPU_getReg(x86_CPU* c,uint8_t i,int64_t b);
void x86_CPU_setReg(x86_CPU* c,uint8_t i,int64_t b,uint32_t v);
uint32_t x86_CPU_gw(x86_CPU* c,int64_t i);
void x86_CPU_sw(x86_CPU* c,int64_t i,uint32_t v);
uint8_t x86_CPU_Reg8(x86_CPU* c,int64_t i);
void x86_CPU_SetReg8(x86_CPU* c,int64_t i,uint8_t v);
uint16_t x86_CPU_Reg16(x86_CPU* c,int64_t i);
void x86_CPU_SetReg16(x86_CPU* c,int64_t i,uint16_t v);
void x86_CPU_push(x86_CPU* c,int64_t b,uint32_t v);
uint32_t x86_CPU_pop(x86_CPU* c,int64_t b);
void x86_CPU_push16(x86_CPU* c,uint32_t v);
uint32_t x86_CPU_pop16(x86_CPU* c);
void x86_CPU_spAdd(x86_CPU* c,uint32_t n);
uint16_t x86_CPU_EFlags(x86_CPU* c);
void x86_CPU_SetEFlags(x86_CPU* c,uint16_t f);
uint32_t x86_widthMask(int64_t b);
uint32_t x86_signMask(int64_t b);
uint32_t x86_signExtByte(uint32_t v);
uint32_t x86_signExtWord(uint32_t v);
void x86_CPU_setSZP(x86_CPU* c,uint32_t res,int64_t b);
bool x86_parity(uint8_t v);
uint32_t x86_CPU_flagsAdd(x86_CPU* c,uint32_t a,uint32_t b,uint32_t cin,int64_t w);
uint32_t x86_CPU_flagsSub(x86_CPU* c,uint32_t a,uint32_t b,uint32_t cin,int64_t w);
uint32_t x86_CPU_flagsLogic(x86_CPU* c,uint32_t res,int64_t w);
bool x86_CPU_cond(x86_CPU* c,uint8_t cc);
uint32_t x86_CPU_rEA(x86_CPU* c,x86_ea o,int64_t b);
void x86_CPU_wEA(x86_CPU* c,x86_ea o,int64_t b,uint32_t v);
int64_t x86_CPU_osz(x86_CPU* c);
int64_t x86_CPU_asz(x86_CPU* c);
uint32_t x86_CPU_segBaseFor(x86_CPU* c,int64_t def);
uint32_t x86_b2u(bool b);
uint32_t x86_CPU_inPort(x86_CPU* c,uint16_t port,int64_t size);
void x86_CPU_outPort(x86_CPU* c,uint16_t port,int64_t size,uint32_t v);
int32_t x86_signExtToInt(uint32_t v,int64_t w);
std::tuple<uint8_t,x86_ea> x86_CPU_modrmE(x86_CPU* c);
x86_ea x86_CPU_ea16(x86_CPU* c,uint8_t mod,uint8_t rm);
x86_ea x86_CPU_ea32(x86_CPU* c,uint8_t mod,uint8_t rm);
void x86_CPU_Step(x86_CPU* c);
uint64_t x86_CPU_Run(x86_CPU* c,uint64_t maxSteps);
std::tuple<uint32_t,bool> x86_CPU_alu(x86_CPU* c,int64_t idx,uint32_t a,uint32_t b,int64_t w);
void x86_CPU_exec(x86_CPU* c,uint8_t op,uint8_t rep);
uint32_t x86_CPU_moffs(x86_CPU* c);
void x86_CPU_aluGrid(x86_CPU* c,int64_t idx,uint8_t z);
uint32_t x86_CPU_incDec(x86_CPU* c,uint32_t v,int64_t w,bool inc);
void x86_CPU_doInt(x86_CPU* c,uint8_t n);
void x86_CPU_divErr(x86_CPU* c);
void x86_CPU_dispatchIVT(x86_CPU* c,uint8_t n);
bool x86_CPU_Interrupt(x86_CPU* c,uint8_t n);
bool x86_CPU_InterruptPM(x86_CPU* c,uint16_t cs,uint32_t eip);
void x86_CPU_grp1(x86_CPU* c,uint8_t op);
void x86_CPU_grp2(x86_CPU* c,int64_t w,std::function<uint32_t()> count);
uint32_t x86_CPU_shiftOp(x86_CPU* c,int64_t idx,uint32_t val,uint32_t cnt,int64_t w);
void x86_CPU_grp3(x86_CPU* c,int64_t w);
void x86_CPU_grp5(x86_CPU* c);
void x86_CPU_mulOp(x86_CPU* c,x86_ea o,int64_t w,bool signed_);
void x86_CPU_divOp(x86_CPU* c,x86_ea o,int64_t w,bool signed_);
uint32_t x86_CPU_imulTrunc(x86_CPU* c,uint32_t a,uint32_t b,int64_t w);
void x86_CPU_stringOp(x86_CPU* c,uint8_t op,uint8_t rep);
void x86_CPU_stringPortOp(x86_CPU* c,uint8_t op,uint8_t rep);
void x86_CPU_advSI(x86_CPU* c,int64_t w,int64_t aw);
void x86_CPU_advDI(x86_CPU* c,int64_t w,int64_t aw);
void x86_CPU_daa(x86_CPU* c,bool sub_);
void x86_CPU_aaa(x86_CPU* c,bool sub_);
void x86_CPU_aam(x86_CPU* c,uint8_t base);
void x86_CPU_aad(x86_CPU* c,uint8_t base);
void x86_CPU_exec0F(x86_CPU* c,uint8_t op,uint8_t rep);
void x86_CPU_bitOp(x86_CPU* c,x86_ea o,uint32_t idx,int64_t w,int64_t mode,bool memBitString);
uint32_t x86_applyBit(uint32_t v,uint32_t b,int64_t mode);
void x86_CPU_bitScan(x86_CPU* c,uint8_t reg,uint32_t src,int64_t w,bool reverse);
void x86_CPU_doubleShift(x86_CPU* c,bool left,x86_ea o,uint8_t reg,uint32_t count,int64_t w);
void x86_FPUState_finit(x86_FPUState* f);
uint8_t x86_classify(double v);
int64_t x86_FPUState_phys(x86_FPUState* f,int64_t i);
double x86_FPUState_st(x86_FPUState* f,int64_t i);
void x86_FPUState_setst(x86_FPUState* f,int64_t i,double v);
void x86_FPUState_push(x86_FPUState* f,double v);
void x86_FPUState_pop(x86_FPUState* f);
uint16_t x86_FPUState_statusWord(x86_FPUState* f);
void x86_FPUState_setCC(x86_FPUState* f,bool c3,bool c2,bool c0);
double x86_FPUState_round(x86_FPUState* f,double v);
double x86_CPU_fLoad32(x86_CPU* c,x86_ea o);
void x86_CPU_fStore32(x86_CPU* c,x86_ea o,double v);
double x86_CPU_fLoad64(x86_CPU* c,x86_ea o);
void x86_CPU_fStore64(x86_CPU* c,x86_ea o,double v);
double x86_CPU_fLoad80(x86_CPU* c,x86_ea o);
void x86_CPU_fStore80(x86_CPU* c,x86_ea o,double v);
double x86_f80ToF64(uint64_t mant,uint16_t se);
std::tuple<uint64_t,uint16_t> x86_f64ToF80(double v);
void x86_CPU_fpuExec(x86_CPU* c,uint8_t op);
std::tuple<double,bool> x86_arithST(int64_t sub_,double a,double b);
void x86_FPUState_fcom(x86_FPUState* f,double a,double b);
void x86_CPU_fcomi(x86_CPU* c,double a,double b);
void x86_CPU_fpuMemExec(x86_CPU* c,uint8_t op,uint8_t reg,x86_ea o);
void x86_CPU_fistStore(x86_CPU* c,x86_ea o,int64_t n,double fv);
void x86_CPU_fpuRegExec(x86_CPU* c,uint8_t op,uint8_t mb,uint8_t reg,uint8_t rm);
double x86_dstArith(int64_t sub_,double dst,double src);
void x86_CPU_fpuD9(x86_CPU* c,uint8_t mb,int64_t i);
void x86_CPU_fxam(x86_CPU* c);
void x86_CPU_fxtract(x86_CPU* c);
void x86_CPU_fprem(x86_CPU* c,bool ieee);
void x86_CPU_fpuDA(x86_CPU* c,uint8_t mb,int64_t i);
void x86_CPU_fpuDB(x86_CPU* c,uint8_t mb,int64_t i);
void x86_CPU_fpuDD(x86_CPU* c,uint8_t mb,int64_t i);
void x86_CPU_fpuDF(x86_CPU* c,uint8_t mb,int64_t i);
bool x86_CPU_envIs32(x86_CPU* c);
void x86_CPU_fnstenv(x86_CPU* c,x86_ea o);
void x86_CPU_fldenv(x86_CPU* c,x86_ea o);
void x86_CPU_fnsave(x86_CPU* c,x86_ea o);
void x86_CPU_frstor(x86_CPU* c,x86_ea o);
void x86_CPU_envWriteHeader(x86_CPU* c,x86_ea o,x86_FPUState* f);
void x86_CPU_envReadHeader(x86_CPU* c,x86_ea o,x86_FPUState* f);
uint16_t x86_FPUState_tagWord(x86_FPUState* f);
void x86_FPUState_setTagWord(x86_FPUState* f,uint16_t w);
std::tuple<uint8_t,std::array<uint8_t,16>,x86_ea,bool> x86_CPU_intOperands(x86_CPU* c,x86_sseKind k);
std::array<uint8_t,16> x86_CPU_intReg(x86_CPU* c,uint8_t reg,bool wide);
void x86_CPU_setIntReg(x86_CPU* c,uint8_t reg,bool wide,std::array<uint8_t,16> v);
int64_t x86_intWidth(bool wide);
uint8_t x86_satI8(int32_t v);
uint8_t x86_satU8(int32_t v);
uint16_t x86_satI16(int32_t v);
uint16_t x86_satU16(int32_t v);
uint16_t x86_w16(Slice<uint8_t> b);
void x86_putw16(Slice<uint8_t> b,uint16_t v);
bool x86_CPU_execMMXInt(x86_CPU* c,uint8_t op,uint8_t rep);
uint16_t x86_satI16Clamp(int32_t v);
uint64_t x86_shiftLeft(uint64_t v,uint32_t bits,uint32_t cnt);
uint64_t x86_shiftRightLogical(uint64_t v,uint32_t bits,uint32_t cnt);
uint64_t x86_shiftRightArith(uint64_t v,uint32_t bits,uint32_t cnt);
void x86_shiftLanes(std::array<uint8_t,16>* v,int64_t n,int64_t elem,uint32_t cnt,std::function<uint64_t(uint64_t,uint32_t,uint32_t)> fn);
void x86_byteShift(std::array<uint8_t,16>* v,int64_t n,bool left);
x86_sseKind x86_CPU_sseKindOf(x86_CPU* c,uint8_t rep);
std::array<uint8_t,16> x86_CPU_sseRM(x86_CPU* c,x86_ea o,int64_t n);
std::array<uint8_t,8> x86_CPU_mmxRM(x86_CPU* c,x86_ea o);
void x86_CPU_mmxStoreRM(x86_CPU* c,x86_ea o,std::array<uint8_t,8> v);
void x86_CPU_sseStoreRM(x86_CPU* c,x86_ea o,std::array<uint8_t,16> v,int64_t n);
float x86_f32Lane(std::array<uint8_t,16> b,int64_t i);
double x86_f64Lane(std::array<uint8_t,16> b,int64_t i);
void x86_setF32Lane(std::array<uint8_t,16>* b,int64_t i,float v);
void x86_setF64Lane(std::array<uint8_t,16>* b,int64_t i,double v);
uint32_t x86_le32b(Slice<uint8_t> b);
uint64_t x86_le64b(Slice<uint8_t> b);
void x86_putle32(Slice<uint8_t> b,uint32_t v);
void x86_putle64(Slice<uint8_t> b,uint64_t v);
int64_t x86_sseWidth(x86_sseKind k);
bool x86_CPU_execSSE(x86_CPU* c,uint8_t op,uint8_t rep);
void x86_CPU_sseArith(x86_CPU* c,uint8_t op,x86_sseKind k,uint8_t reg,x86_ea o);
void x86_CPU_sseUnary(x86_CPU* c,x86_sseKind k,uint8_t reg,x86_ea o,std::function<double(double)> fn);
void x86_CPU_sseCompareFlags(x86_CPU* c,double a,double b);
bool dos_COFFSection_IsText(dos_COFFSection s);
bool dos_COFFSection_IsBSS(dos_COFFSection s);
std::tuple<dos_COFF*,Error> dos_ParseGo32COFF(Slice<uint8_t> data);
dos_COFFSection* dos_COFF_TextSection(dos_COFF* c);
dos_COFFSection* dos_COFF_SectionAt(dos_COFF* c,uint32_t vaddr);
std::tuple<dos_Machine*,Error> dos_LoadEXE(std::string exePath,std::string gameDir);
uint8_t dos_Machine_Read(dos_Machine* m,uint32_t a);
Slice<dos_VGAWriter> dos_Machine_ReadProfile(dos_Machine* m);
void dos_Machine_Write(dos_Machine* m,uint32_t a,uint8_t v);
Slice<dos_VGAWriter> dos_Machine_VGAProfile(dos_Machine* m);
uint16_t dos_Machine_r16(dos_Machine* m,uint32_t lin);
void dos_Machine_w16(dos_Machine* m,uint32_t lin,uint16_t v);
uint32_t dos_lin(uint16_t seg,uint16_t off);
void dos_Machine_setupBIOS(dos_Machine* m);
void dos_Machine_setupPSP(dos_Machine* m);
void dos_Machine_setupEnv(dos_Machine* m,std::string exePath);
std::tuple<std::string,bool> dos_Machine_resolveFile(dos_Machine* m,std::string dosPath);
std::string dos_Machine_scratchPath(dos_Machine* m,std::string dosPath);
Error dos_Machine_SeedSaveFile(dos_Machine* m,std::string dosPath,Slice<uint8_t> data);
std::tuple<std::string,bool> dos_Machine_walkPath(dos_Machine* m,std::string dosPath);
std::tuple<std::string,bool> dos_walkRoot(std::string root,std::string dosPath);
void dos_Machine_SeedDir(dos_Machine* m,std::string rel);
uint16_t dos_Machine_allocFH(dos_Machine* m);
std::string dos_Machine_asciiz(dos_Machine* m,uint16_t seg,uint16_t off);
Slice<std::string> dos_Machine_IntSummary(dos_Machine* m);
void dos_Machine_setupEMS(dos_Machine* m);
uint32_t dos_emsState_frameLin(dos_emsState* e,int64_t slot);
Slice<uint8_t> dos_Machine_EmsBacking(dos_Machine* m);
int64_t dos_Machine_EmsHandleBase(dos_Machine* m,uint16_t h);
void dos_Machine_emsFlush(dos_Machine* m,int64_t slot);
void dos_Machine_emsLoad(dos_Machine* m,int64_t slot,int64_t p);
bool dos_Machine_int67(dos_Machine* m,x86_CPU* c);
void dos_Machine_emsPageMap(dos_Machine* m,x86_CPU* c);
bool dos_Machine_handleInt(dos_Machine* m,x86_CPU* c,uint8_t n);
bool dos_Machine_int21(dos_Machine* m,x86_CPU* c);
bool dos_Machine_overlayInt(dos_Machine* m,x86_CPU* c);
bool dos_Machine_findFirst(dos_Machine* m,x86_CPU* c);
bool dos_Machine_findNext(dos_Machine* m,x86_CPU* c);
bool dos_dosMatch(std::string pattern,std::string name);
bool dos_globSegments(std::string p,std::string n);
bool dos_Machine_dosErr(dos_Machine* m,x86_CPU* c,uint16_t code);
std::string dos_Machine_dollarStr(dos_Machine* m,uint16_t seg,uint16_t off);
uint32_t dos_Machine_portIn(dos_Machine* m,uint16_t port,int64_t size);
void dos_Machine_portOut(dos_Machine* m,uint16_t port,int64_t size,uint32_t v);
uint32_t dos_widthMask8(int64_t size);
void dos_Machine_onStep(dos_Machine* m,x86_CPU* c);
std::tuple<uint8_t,bool> dos_Scancode(std::string token);
std::tuple<Slice<dos_injEvent>,Error> dos_ParseKeys(std::string spec);
Slice<std::string> dos_splitTokens(std::string spec);
void dos_Machine_SetKeys(dos_Machine* m,Slice<dos_injEvent> events);
bool dos_Machine_KeysPending(dos_Machine* m);
void dos_Machine_pumpKeys(dos_Machine* m);
void dos_Machine_setupMCB(dos_Machine* m);
void dos_Machine_writeMCB(dos_Machine* m,uint16_t mcbSeg,uint8_t mark,uint16_t owner,uint16_t size);
std::tuple<uint8_t,uint16_t,uint16_t> dos_Machine_readMCB(dos_Machine* m,uint16_t mcbSeg);
std::tuple<uint16_t,uint16_t,bool> dos_Machine_allocBlock(dos_Machine* m,uint16_t want);
void dos_Machine_carve(dos_Machine* m,uint16_t mcb,uint8_t mark,uint16_t size,uint16_t want,uint16_t owner);
bool dos_Machine_freeBlock(dos_Machine* m,uint16_t blockSeg);
void dos_Machine_coalesce(dos_Machine* m);
std::tuple<uint16_t,bool> dos_Machine_resizeBlock(dos_Machine* m,uint16_t blockSeg,uint16_t want);
dos_mouseState* dos_Machine_mouse(dos_Machine* m);
bool dos_Machine_int33(dos_Machine* m,x86_CPU* c);
void dos_mouseState_clamp(dos_mouseState* ms);
void dos_Machine_HomeMouse(dos_Machine* m);
void dos_Machine_MoveMouseTo(dos_Machine* m,int64_t x,int64_t y);
void dos_Machine_FeedMickeys(dos_Machine* m,int64_t dx,int64_t dy);
void dos_Machine_SetMouseButtons(dos_Machine* m,uint8_t mask);
void dos_Machine_vgaInit13h(dos_Machine* m,bool clear);
bool dos_vgaState_chained(dos_vgaState* v);
void dos_Machine_vgaWrite(dos_Machine* m,uint32_t a,uint8_t val);
uint8_t dos_Machine_vgaRead(dos_Machine* m,uint32_t a);
void dos_Machine_vgaRegOut(dos_Machine* m,uint16_t port,uint8_t b);
bool dos_Machine_int10(dos_Machine* m,x86_CPU* c);
std::tuple<dos_PM*,Error> dos_LoadGo32(std::string exePath,std::string gameDir);
std::tuple<dos_PM*,Error> dos_LoadGo32Bytes(Slice<uint8_t> data,std::string gameDir);
uint32_t dos_PM_resolveSel(dos_PM* p,uint16_t sel);
void dos_PM_mapSel(dos_PM* p,uint16_t sel,uint32_t base);
void dos_PM_setupInfoBlock(dos_PM* p,uint32_t xferLinear);
void dos_PM_setupDefaultIntVec(dos_PM* p);
bool dos_PM_writeDOSStructures(dos_PM* p);
void dos_PM_enforceBaseAddress(dos_PM* p);
uint8_t dos_PM_Read(dos_PM* p,uint32_t a);
void dos_PM_Write(dos_PM* p,uint32_t a,uint8_t v);
uint32_t dos_PM_watchPC(dos_PM* p);
void dos_PM_SetWriteWatch(dos_PM* p,uint32_t lo,uint32_t hi,std::function<void(uint32_t,uint32_t,uint32_t)> cb);
void dos_PM_SetReadWatch(dos_PM* p,uint32_t lo,uint32_t hi,std::function<void(uint32_t,uint32_t,uint32_t)> cb);
Slice<dos_MemRegion> dos_PM_MemRegions(dos_PM* p);
void dos_PM_fault(dos_PM* p,std::string kind,uint32_t a);
uint32_t dos_align(uint32_t v,uint32_t a);
uint32_t dos_maxu32(uint32_t a,uint32_t b);
std::string dos_pcHex(x86_CPU* c);
uint32_t dos_PM_r32(dos_PM* p,uint32_t a);
void dos_PM_w32(dos_PM* p,uint32_t a,uint32_t v);
void dos_PM_w16(dos_PM* p,uint32_t a,uint16_t v);
std::string dos_PM_asciiz(dos_PM* p,uint32_t a);
uint32_t dos_PM_rmLinear(dos_PM* p,uint16_t seg,uint32_t off);
bool dos_PM_dosErr(dos_PM* p,dos_rmcs* r,uint16_t code);
void dos_setAX(dos_rmcs* r,uint16_t v);
void dos_setBX(dos_rmcs* r,uint16_t v);
void dos_setCX(dos_rmcs* r,uint16_t v);
void dos_setDX(dos_rmcs* r,uint16_t v);
uint16_t dos_PM_allocHandle(dos_PM* p);
std::string dos_PM_resolveHostPath(dos_PM* p,std::string name);
bool dos_PM_dosFile(dos_PM* p,dos_rmcs* r);
bool dos_PM_dosTruename(dos_PM* p,dos_rmcs* r);
std::string dos_canonicalizeDOSPath(std::string s);
bool dos_PM_dosOpen(dos_PM* p,dos_rmcs* r);
bool dos_PM_dosCreate(dos_PM* p,dos_rmcs* r);
bool dos_PM_dosClose(dos_PM* p,dos_rmcs* r);
bool dos_PM_dosRead(dos_PM* p,dos_rmcs* r);
bool dos_PM_dosWrite(dos_PM* p,dos_rmcs* r);
bool dos_PM_dosSeek(dos_PM* p,dos_rmcs* r);
bool dos_PM_dosFileTime(dos_PM* p,dos_rmcs* r);
bool dos_PM_dosIoctl(dos_PM* p,dos_rmcs* r);
bool dos_PM_handleInt(dos_PM* p,x86_CPU* c,uint8_t n);
void dos_PM_countInt(dos_PM* p,uint8_t n);
bool dos_PM_dpmi(dos_PM* p,x86_CPU* c);
void dos_PM_writeDescriptor(dos_PM* p,uint32_t a,uint32_t base);
uint32_t dos_PM_readDescriptorBase(dos_PM* p,uint32_t a);
std::tuple<uint16_t,uint16_t> dos_PM_allocCallback(dos_PM* p);
uint8_t dos_b2u8(bool v);
uint16_t dos_PM_allocSel(dos_PM* p,uint16_t n);
std::tuple<uint16_t,uint16_t,bool> dos_PM_allocConv(dos_PM* p,uint16_t paras);
std::tuple<uint32_t,bool> dos_PM_allocHeap(dos_PM* p,uint32_t size);
void dos_PM_freeMemInfo(dos_PM* p,uint32_t a);
bool dos_PM_int21(dos_PM* p,x86_CPU* c);
std::string dos_PM_dollarStr(dos_PM* p,uint32_t a);
void dos_PM_setPMVector(dos_PM* p,uint8_t bl,uint16_t sel,uint32_t off);
void dos_PM_SetKeys(dos_PM* p,Slice<dos_injEvent> events);
bool dos_PM_KeysPending(dos_PM* p);
void dos_PM_PumpInput(dos_PM* p,x86_CPU* c);
void dos_PM_EnqueueScancode(dos_PM* p,uint8_t sc);
bool dos_PM_InteractiveKeysPending(dos_PM* p);
void dos_PM_pumpInteractiveKeys(dos_PM* p,x86_CPU* c);
bool dos_PM_deliverKey(dos_PM* p,x86_CPU* c);
bool dos_PM_deliverScancode(dos_PM* p,x86_CPU* c,uint8_t sc);
void dos_PM_popKeyEvent(dos_PM* p);
uint64_t dos_PM_pitTotalTicks(dos_PM* p);
uint16_t dos_PM_pitCounter(dos_PM* p);
void dos_PM_pitSyncBiosTick(dos_PM* p);
uint32_t dos_PM_portIn(dos_PM* p,uint16_t port,int64_t size);
void dos_PM_portOut(dos_PM* p,uint16_t port,int64_t size,uint32_t v);
dos_rmcs dos_PM_readRMCS(dos_PM* p,uint32_t a);
void dos_PM_writeBack(dos_PM* p,dos_rmcs* r);
bool dos_PM_simulateRealInt(dos_PM* p,x86_CPU* c);
bool dos_PM_rmDOS(dos_PM* p,dos_rmcs* r);
int64_t dos_Reloc_FarAddr(dos_Reloc r);
std::tuple<dos_MZ*,Error> dos_ParseMZ(Slice<uint8_t> data);
int64_t dos_MZ_EntryLinear(dos_MZ* m);
int64_t dos_MZ_StackLinear(dos_MZ* m);
constexpr int64_t x86_AX=0ULL;
constexpr int64_t x86_CX=1ULL;
constexpr int64_t x86_DX=2ULL;
constexpr int64_t x86_BX=3ULL;
constexpr int64_t x86_SP=4ULL;
constexpr int64_t x86_BP=5ULL;
constexpr int64_t x86_SI=6ULL;
constexpr int64_t x86_DI=7ULL;
constexpr int64_t x86_AL=0ULL;
constexpr int64_t x86_CL=1ULL;
constexpr int64_t x86_DL=2ULL;
constexpr int64_t x86_BL=3ULL;
constexpr int64_t x86_AH=4ULL;
constexpr int64_t x86_CH=5ULL;
constexpr int64_t x86_DH=6ULL;
constexpr int64_t x86_BH=7ULL;
constexpr int64_t x86_ES=0ULL;
constexpr int64_t x86_CS=1ULL;
constexpr int64_t x86_SS=2ULL;
constexpr int64_t x86_DS=3ULL;
constexpr int64_t x86_FS=4ULL;
constexpr int64_t x86_GS=5ULL;
constexpr int64_t x86_ModeReal=0ULL;
constexpr int64_t x86_ModeProt=1ULL;
constexpr int64_t x86_tagValid=0ULL;
constexpr int64_t x86_tagZero=1ULL;
constexpr int64_t x86_tagSpecial=2ULL;
constexpr int64_t x86_tagEmpty=3ULL;
constexpr x86_sseKind x86_ssePS=0ULL;
constexpr x86_sseKind x86_ssePD=1ULL;
constexpr x86_sseKind x86_sseSS=2ULL;
constexpr x86_sseKind x86_sseSD=3ULL;
constexpr int64_t dos_coffI386Magic=332ULL;
constexpr int64_t dos_coffZMagic=267ULL;
constexpr int64_t dos_coffStypText=32ULL;
constexpr int64_t dos_coffStypData=64ULL;
constexpr int64_t dos_coffStypBss=128ULL;
constexpr int64_t dos_emsPageSize=16384ULL;
constexpr int64_t dos_emsFrameSeg=57344ULL;
constexpr int64_t dos_emsSigSeg=49152ULL;
constexpr int64_t dos_emsTotalPage=512ULL;
constexpr int64_t dos_EmsPageSize=16384ULL;
constexpr int64_t dos_ticksEveryInstrs=800ULL;
constexpr dos_injKind dos_injWait=0ULL;
constexpr dos_injKind dos_injKey=1ULL;
constexpr dos_injKind dos_injHome=2ULL;
constexpr dos_injKind dos_injMove=3ULL;
constexpr dos_injKind dos_injButton=4ULL;
constexpr dos_injKind dos_injDelta=5ULL;
Map<std::string,uint8_t> dos_makeScancodes=Map<std::string,uint8_t>{{std::string("esc",3),cast<uint8_t>(1ULL)},{std::string("1",1),cast<uint8_t>(2ULL)},{std::string("2",1),cast<uint8_t>(3ULL)},{std::string("3",1),cast<uint8_t>(4ULL)},{std::string("4",1),cast<uint8_t>(5ULL)},{std::string("5",1),cast<uint8_t>(6ULL)},{std::string("6",1),cast<uint8_t>(7ULL)},{std::string("7",1),cast<uint8_t>(8ULL)},{std::string("8",1),cast<uint8_t>(9ULL)},{std::string("9",1),cast<uint8_t>(10ULL)},{std::string("0",1),cast<uint8_t>(11ULL)},{std::string("-",1),cast<uint8_t>(12ULL)},{std::string("=",1),cast<uint8_t>(13ULL)},{std::string("bs",2),cast<uint8_t>(14ULL)},{std::string("backspace",9),cast<uint8_t>(14ULL)},{std::string("tab",3),cast<uint8_t>(15ULL)},{std::string("q",1),cast<uint8_t>(16ULL)},{std::string("w",1),cast<uint8_t>(17ULL)},{std::string("e",1),cast<uint8_t>(18ULL)},{std::string("r",1),cast<uint8_t>(19ULL)},{std::string("t",1),cast<uint8_t>(20ULL)},{std::string("y",1),cast<uint8_t>(21ULL)},{std::string("u",1),cast<uint8_t>(22ULL)},{std::string("i",1),cast<uint8_t>(23ULL)},{std::string("o",1),cast<uint8_t>(24ULL)},{std::string("p",1),cast<uint8_t>(25ULL)},{std::string("enter",5),cast<uint8_t>(28ULL)},{std::string("return",6),cast<uint8_t>(28ULL)},{std::string("a",1),cast<uint8_t>(30ULL)},{std::string("s",1),cast<uint8_t>(31ULL)},{std::string("d",1),cast<uint8_t>(32ULL)},{std::string("f",1),cast<uint8_t>(33ULL)},{std::string("g",1),cast<uint8_t>(34ULL)},{std::string("h",1),cast<uint8_t>(35ULL)},{std::string("j",1),cast<uint8_t>(36ULL)},{std::string("k",1),cast<uint8_t>(37ULL)},{std::string("l",1),cast<uint8_t>(38ULL)},{std::string(";",1),cast<uint8_t>(39ULL)},{std::string("z",1),cast<uint8_t>(44ULL)},{std::string("x",1),cast<uint8_t>(45ULL)},{std::string("c",1),cast<uint8_t>(46ULL)},{std::string("v",1),cast<uint8_t>(47ULL)},{std::string("b",1),cast<uint8_t>(48ULL)},{std::string("n",1),cast<uint8_t>(49ULL)},{std::string("m",1),cast<uint8_t>(50ULL)},{std::string(",",1),cast<uint8_t>(51ULL)},{std::string(".",1),cast<uint8_t>(52ULL)},{std::string("/",1),cast<uint8_t>(53ULL)},{std::string("space",5),cast<uint8_t>(57ULL)},{std::string("sp",2),cast<uint8_t>(57ULL)},{std::string("lshift",6),cast<uint8_t>(42ULL)},{std::string("shift",5),cast<uint8_t>(42ULL)},{std::string("rshift",6),cast<uint8_t>(54ULL)},{std::string("ctrl",4),cast<uint8_t>(29ULL)},{std::string("alt",3),cast<uint8_t>(56ULL)},{std::string("up",2),cast<uint8_t>(72ULL)},{std::string("left",4),cast<uint8_t>(75ULL)},{std::string("right",5),cast<uint8_t>(77ULL)},{std::string("down",4),cast<uint8_t>(80ULL)},{std::string("home",4),cast<uint8_t>(71ULL)},{std::string("end",3),cast<uint8_t>(79ULL)},{std::string("pgup",4),cast<uint8_t>(73ULL)},{std::string("pgdn",4),cast<uint8_t>(81ULL)},{std::string("ins",3),cast<uint8_t>(82ULL)},{std::string("del",3),cast<uint8_t>(83ULL)},{std::string("f1",2),cast<uint8_t>(59ULL)},{std::string("f2",2),cast<uint8_t>(60ULL)},{std::string("f3",2),cast<uint8_t>(61ULL)},{std::string("f4",2),cast<uint8_t>(62ULL)},{std::string("f5",2),cast<uint8_t>(63ULL)},{std::string("f6",2),cast<uint8_t>(64ULL)},{std::string("f7",2),cast<uint8_t>(65ULL)},{std::string("f8",2),cast<uint8_t>(66ULL)},{std::string("f9",2),cast<uint8_t>(67ULL)},{std::string("f10",3),cast<uint8_t>(68ULL)}};
constexpr int64_t dos_mcbNormal=77ULL;
constexpr int64_t dos_mcbLast=90ULL;
constexpr int64_t dos_mcbFree=0ULL;
constexpr int64_t dos_mickeysPerPixel=2ULL;
constexpr int64_t dos_go32ImgBase=1048576ULL;
constexpr int64_t dos_go32VASize=67108864ULL;
constexpr int64_t dos_go32MemSize=68157440ULL;
constexpr int64_t dos_go32StackBytes=8388608ULL;
constexpr int64_t dos_go32PageSize=4096ULL;
constexpr int64_t dos_go32InfoBytes=16384ULL;
constexpr int64_t dos_go32XferBytes=8192ULL;
constexpr int64_t dos_go32InfoSel=64ULL;
constexpr int64_t dos_go32DosDS=24ULL;
constexpr int64_t dos_go32MinStack=8388608ULL;
constexpr int64_t dos_go32ConvBase=65536ULL;
constexpr int64_t dos_go32BaseAddrVA=448204ULL;
constexpr int64_t dos_dosJFTSize=64ULL;
constexpr int64_t dos_dosSFTEntrySize=59ULL;
constexpr int64_t dos_go32MaxFreeReport=16777216ULL;
constexpr int64_t dos_go32InjectPeriod=40000ULL;
constexpr int64_t dos_pitBiosTickAddr=1132ULL;
constexpr int64_t dos_pitInstrsPerTick=32ULL;

#include "adapters.h"
// tools/cpu/x86/cpu.go:154:1
x86_CPU* x86_NewCPU(RRX86Bus bus){
{
x86_CPU* c = arenaNew(x86_CPU{{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},bus,{},{},{},{},{},{},{},{},{},{}});
x86_FPUState_finit(&(c->FPU));
c->MXCSR = cast<uint32_t>(8064ULL);
c->TSCMul = cast<uint64_t>(1ULL);
return c;
}
}
// tools/cpu/x86/cpu.go:179:1
uint32_t x86_CPU_LinearPC(x86_CPU* c){
{
return cast<uint32_t>((c->SegBase[cast<int64_t>(1ULL)] + c->instrIP));
}
}
// tools/cpu/x86/cpu.go:185:1
std::string x86_CPU_at(x86_CPU* c){
{
if ((c->Mode == cast<int64_t>(1ULL))) {
return go_fmt_Sprintf(std::string("%08X",4),c->instrIP);
}
return go_fmt_Sprintf(std::string("%04X:%04X",9),c->Seg[cast<int64_t>(1ULL)],cast<uint32_t>((c->instrIP & cast<uint32_t>(65535ULL))));
}
}
// tools/cpu/x86/cpu.go:196:1
uint32_t x86_CPU_segBase(x86_CPU* c,int64_t idx){
{
if ((c->Mode == cast<int64_t>(0ULL))) {
return shl<uint32_t>(cast<uint32_t>(c->Seg[idx]),cast<int64_t>(4ULL));
}
return c->SegBase[idx];
}
}
// tools/cpu/x86/cpu.go:206:1
uint32_t x86_CPU_linear(x86_CPU* c,uint32_t base,uint32_t off){
{
if ((c->Mode == cast<int64_t>(0ULL))) {
return cast<uint32_t>(((cast<uint32_t>((base + (cast<uint32_t>((off & cast<uint32_t>(65535ULL))))))) & cast<uint32_t>(1048575ULL)));
}
return cast<uint32_t>((base + off));
}
}
// tools/cpu/x86/cpu.go:214:1
uint32_t x86_CPU_ipMask(x86_CPU* c,uint32_t ip){
{
if ((c->Mode == cast<int64_t>(0ULL))) {
return cast<uint32_t>((ip & cast<uint32_t>(65535ULL)));
}
return ip;
}
}
// tools/cpu/x86/cpu.go:222:1
bool x86_CPU_stack32(x86_CPU* c){
{
return (c->Mode == cast<int64_t>(1ULL));
}
}
// tools/cpu/x86/cpu.go:224:1
uint32_t x86_CPU_rd8(x86_CPU* c,uint32_t a){
{
return cast<uint32_t>(rrBus_Read(c->bus,a));
}
}
// tools/cpu/x86/cpu.go:225:1
uint32_t x86_CPU_rd16(x86_CPU* c,uint32_t a){
{
return cast<uint32_t>((x86_CPU_rd8(c,a) | shl<uint32_t>(x86_CPU_rd8(c,cast<uint32_t>((a + cast<uint32_t>(1ULL)))),cast<int64_t>(8ULL))));
}
}
// tools/cpu/x86/cpu.go:228:1
uint32_t x86_CPU_rd32(x86_CPU* c,uint32_t a){
{
return cast<uint32_t>((x86_CPU_rd16(c,a) | shl<uint32_t>(x86_CPU_rd16(c,cast<uint32_t>((a + cast<uint32_t>(2ULL)))),cast<int64_t>(16ULL))));
}
}
// tools/cpu/x86/cpu.go:231:1
void x86_CPU_wr8(x86_CPU* c,uint32_t a,uint32_t v){
{
rrBus_Write(c->bus,a,cast<uint8_t>(v));
}
}
// tools/cpu/x86/cpu.go:232:1
void x86_CPU_wr16(x86_CPU* c,uint32_t a,uint32_t v){
{
x86_CPU_wr8(c,a,v);
x86_CPU_wr8(c,cast<uint32_t>((a + cast<uint32_t>(1ULL))),shr<uint32_t>(v,cast<int64_t>(8ULL)));
}
}
// tools/cpu/x86/cpu.go:236:1
void x86_CPU_wr32(x86_CPU* c,uint32_t a,uint32_t v){
{
x86_CPU_wr16(c,a,v);
x86_CPU_wr16(c,cast<uint32_t>((a + cast<uint32_t>(2ULL))),shr<uint32_t>(v,cast<int64_t>(16ULL)));
}
}
// tools/cpu/x86/cpu.go:246:1
uint32_t x86_CPU_memRead_reference(x86_CPU* c,uint32_t base,uint32_t off,int64_t b){
{
uint32_t v = x86_CPU_rd8(c,x86_CPU_linear(c,base,off));
if ((b >= cast<int64_t>(2ULL))) {
v |= shl<uint32_t>(x86_CPU_rd8(c,x86_CPU_linear(c,base,cast<uint32_t>((off + cast<uint32_t>(1ULL))))),cast<int64_t>(8ULL));
}
if ((b == cast<int64_t>(4ULL))) {
v |= cast<uint32_t>((shl<uint32_t>(x86_CPU_rd8(c,x86_CPU_linear(c,base,cast<uint32_t>((off + cast<uint32_t>(2ULL))))),cast<int64_t>(16ULL)) | shl<uint32_t>(x86_CPU_rd8(c,x86_CPU_linear(c,base,cast<uint32_t>((off + cast<uint32_t>(3ULL))))),cast<int64_t>(24ULL))));
}
return v;
}
}
// tools/cpu/x86/cpu.go:256:1
void x86_CPU_memWrite_reference(x86_CPU* c,uint32_t base,uint32_t off,int64_t b,uint32_t v){
{
x86_CPU_wr8(c,x86_CPU_linear(c,base,off),v);
if ((b >= cast<int64_t>(2ULL))) {
x86_CPU_wr8(c,x86_CPU_linear(c,base,cast<uint32_t>((off + cast<uint32_t>(1ULL)))),shr<uint32_t>(v,cast<int64_t>(8ULL)));
}
if ((b == cast<int64_t>(4ULL))) {
x86_CPU_wr8(c,x86_CPU_linear(c,base,cast<uint32_t>((off + cast<uint32_t>(2ULL)))),shr<uint32_t>(v,cast<int64_t>(16ULL)));
x86_CPU_wr8(c,x86_CPU_linear(c,base,cast<uint32_t>((off + cast<uint32_t>(3ULL)))),shr<uint32_t>(v,cast<int64_t>(24ULL)));
}
}
}
// tools/cpu/x86/cpu.go:269:1
uint32_t x86_CPU_fetch8(x86_CPU* c){
{
uint32_t v = x86_CPU_rd8(c,x86_CPU_linear(c,x86_CPU_segBase(c,cast<int64_t>(1ULL)),c->IP));
c->IP = x86_CPU_ipMask(c,cast<uint32_t>((c->IP + cast<uint32_t>(1ULL))));
return v;
}
}
// tools/cpu/x86/cpu.go:274:1
uint32_t x86_CPU_fetch16_reference(x86_CPU* c){
{
uint32_t lo = x86_CPU_fetch8(c);
uint32_t hi = x86_CPU_fetch8(c);
return cast<uint32_t>((lo | shl<uint32_t>(hi,cast<int64_t>(8ULL))));
}
}
// tools/cpu/x86/cpu.go:279:1
uint32_t x86_CPU_fetch32_reference(x86_CPU* c){
{
uint32_t lo = x86_CPU_fetch16(c);
uint32_t hi = x86_CPU_fetch16(c);
return cast<uint32_t>((lo | shl<uint32_t>(hi,cast<int64_t>(16ULL))));
}
}
// tools/cpu/x86/cpu.go:286:1
uint32_t x86_CPU_fetchImm(x86_CPU* c){
{
if ((c->dOpsize == cast<int64_t>(32ULL))) {
return x86_CPU_fetch32(c);
}
return x86_CPU_fetch16(c);
}
}
// tools/cpu/x86/cpu.go:300:1
void x86_CPU_loadSeg(x86_CPU* c,int64_t idx,uint16_t sel){
{
c->Seg[idx] = sel;
if (((c->Mode == cast<int64_t>(1ULL)) && bool(c->SegResolve))) {
c->SegBase[idx] = c->SegResolve(sel);
}
}
}
// tools/cpu/x86/cpu.go:307:1
uint32_t x86_CPU_g8(x86_CPU* c,uint8_t i){
{
if ((i < cast<uint8_t>(4ULL))) {
return cast<uint32_t>((c->Regs[i] & cast<uint32_t>(255ULL)));
}
return cast<uint32_t>(((shr<uint32_t>(c->Regs[cast<uint8_t>((i - cast<uint8_t>(4ULL)))],cast<int64_t>(8ULL))) & cast<uint32_t>(255ULL)));
}
}
// tools/cpu/x86/cpu.go:313:1
void x86_CPU_s8(x86_CPU* c,uint8_t i,uint32_t v){
{
if ((i < cast<uint8_t>(4ULL))) {
c->Regs[i] = cast<uint32_t>((((c->Regs[i] & ~(cast<uint32_t>(255ULL)))) | (cast<uint32_t>((v & cast<uint32_t>(255ULL))))));
}
else {
c->Regs[cast<uint8_t>((i - cast<uint8_t>(4ULL)))] = cast<uint32_t>((((c->Regs[cast<uint8_t>((i - cast<uint8_t>(4ULL)))] & ~(cast<uint32_t>(65280ULL)))) | (shl<uint32_t>((cast<uint32_t>((v & cast<uint32_t>(255ULL)))),cast<int64_t>(8ULL)))));
}
}
}
// tools/cpu/x86/cpu.go:320:1
uint32_t x86_CPU_g16(x86_CPU* c,uint8_t i){
{
return cast<uint32_t>((c->Regs[i] & cast<uint32_t>(65535ULL)));
}
}
// tools/cpu/x86/cpu.go:321:1
void x86_CPU_s16(x86_CPU* c,uint8_t i,uint32_t v){
{
c->Regs[i] = cast<uint32_t>((((c->Regs[i] & ~(cast<uint32_t>(65535ULL)))) | (cast<uint32_t>((v & cast<uint32_t>(65535ULL))))));
}
}
// tools/cpu/x86/cpu.go:324:1
uint32_t x86_CPU_getReg(x86_CPU* c,uint8_t i,int64_t b){
{
{
switch(b){
case cast<int64_t>(1ULL):{
return x86_CPU_g8(c,i);
break;}
case cast<int64_t>(2ULL):{
return x86_CPU_g16(c,i);
break;}
default:{
return c->Regs[i];
break;}
}}
}
}
// tools/cpu/x86/cpu.go:334:1
void x86_CPU_setReg(x86_CPU* c,uint8_t i,int64_t b,uint32_t v){
{
{
switch(b){
case cast<int64_t>(1ULL):{
x86_CPU_s8(c,i,v);
break;}
case cast<int64_t>(2ULL):{
x86_CPU_s16(c,i,v);
break;}
default:{
c->Regs[i] = v;
break;}
}}
}
}
// tools/cpu/x86/cpu.go:346:1
uint32_t x86_CPU_gw(x86_CPU* c,int64_t i){
{
return cast<uint32_t>((c->Regs[i] & cast<uint32_t>(65535ULL)));
}
}
// tools/cpu/x86/cpu.go:347:1
void x86_CPU_sw(x86_CPU* c,int64_t i,uint32_t v){
{
c->Regs[i] = cast<uint32_t>((((c->Regs[i] & ~(cast<uint32_t>(65535ULL)))) | (cast<uint32_t>((v & cast<uint32_t>(65535ULL))))));
}
}
// tools/cpu/x86/cpu.go:352:1
uint8_t x86_CPU_Reg8(x86_CPU* c,int64_t i){
{
return cast<uint8_t>(x86_CPU_g8(c,cast<uint8_t>(i)));
}
}
// tools/cpu/x86/cpu.go:355:1
void x86_CPU_SetReg8(x86_CPU* c,int64_t i,uint8_t v){
{
x86_CPU_s8(c,cast<uint8_t>(i),cast<uint32_t>(v));
}
}
// tools/cpu/x86/cpu.go:358:1
uint16_t x86_CPU_Reg16(x86_CPU* c,int64_t i){
{
return cast<uint16_t>(x86_CPU_gw(c,i));
}
}
// tools/cpu/x86/cpu.go:361:1
void x86_CPU_SetReg16(x86_CPU* c,int64_t i,uint16_t v){
{
x86_CPU_sw(c,i,cast<uint32_t>(v));
}
}
// tools/cpu/x86/cpu.go:368:1
void x86_CPU_push(x86_CPU* c,int64_t b,uint32_t v){
{
if (x86_CPU_stack32(c)) {
c->Regs[cast<int64_t>(4ULL)] -= cast<uint32_t>(b);
x86_CPU_memWrite(c,x86_CPU_segBase(c,cast<int64_t>(2ULL)),c->Regs[cast<int64_t>(4ULL)],b,v);
return ;
}
x86_CPU_sw(c,cast<int64_t>(4ULL),cast<uint32_t>((x86_CPU_gw(c,cast<int64_t>(4ULL)) - cast<uint32_t>(b))));
x86_CPU_memWrite(c,x86_CPU_segBase(c,cast<int64_t>(2ULL)),x86_CPU_gw(c,cast<int64_t>(4ULL)),b,v);
}
}
// tools/cpu/x86/cpu.go:377:1
uint32_t x86_CPU_pop(x86_CPU* c,int64_t b){
{
if (x86_CPU_stack32(c)) {
uint32_t v = x86_CPU_memRead(c,x86_CPU_segBase(c,cast<int64_t>(2ULL)),c->Regs[cast<int64_t>(4ULL)],b);
c->Regs[cast<int64_t>(4ULL)] += cast<uint32_t>(b);
return v;
}
uint32_t v = x86_CPU_memRead(c,x86_CPU_segBase(c,cast<int64_t>(2ULL)),x86_CPU_gw(c,cast<int64_t>(4ULL)),b);
x86_CPU_sw(c,cast<int64_t>(4ULL),cast<uint32_t>((x86_CPU_gw(c,cast<int64_t>(4ULL)) + cast<uint32_t>(b))));
return v;
}
}
// tools/cpu/x86/cpu.go:387:1
void x86_CPU_push16(x86_CPU* c,uint32_t v){
{
x86_CPU_push(c,cast<int64_t>(2ULL),v);
}
}
// tools/cpu/x86/cpu.go:388:1
uint32_t x86_CPU_pop16(x86_CPU* c){
{
return x86_CPU_pop(c,cast<int64_t>(2ULL));
}
}
// tools/cpu/x86/cpu.go:392:1
void x86_CPU_spAdd(x86_CPU* c,uint32_t n){
{
if (x86_CPU_stack32(c)) {
c->Regs[cast<int64_t>(4ULL)] += n;
}
else {
x86_CPU_sw(c,cast<int64_t>(4ULL),cast<uint32_t>((x86_CPU_gw(c,cast<int64_t>(4ULL)) + n)));
}
}
}
// tools/cpu/x86/cpu.go:403:1
uint16_t x86_CPU_EFlags(x86_CPU* c){
{
uint16_t f = cast<uint16_t>(2ULL);
auto set = [&](bool cond,uint64_t bit)->void{
if (cond) {
f |= shl<uint16_t>(cast<uint16_t>(1ULL),bit);
}
}
;
set(c->CF,cast<uint64_t>(0ULL));
set(c->PF,cast<uint64_t>(2ULL));
set(c->AF,cast<uint64_t>(4ULL));
set(c->ZF,cast<uint64_t>(6ULL));
set(c->SF,cast<uint64_t>(7ULL));
set(c->TF,cast<uint64_t>(8ULL));
set(c->IF,cast<uint64_t>(9ULL));
set(c->DF,cast<uint64_t>(10ULL));
set(c->OF,cast<uint64_t>(11ULL));
return f;
}
}
// tools/cpu/x86/cpu.go:423:1
void x86_CPU_SetEFlags(x86_CPU* c,uint16_t f){
{
c->CF = (cast<uint16_t>((f & cast<uint16_t>(1ULL))) != cast<uint16_t>(0ULL));
c->PF = (cast<uint16_t>((f & cast<uint16_t>(4ULL))) != cast<uint16_t>(0ULL));
c->AF = (cast<uint16_t>((f & cast<uint16_t>(16ULL))) != cast<uint16_t>(0ULL));
c->ZF = (cast<uint16_t>((f & cast<uint16_t>(64ULL))) != cast<uint16_t>(0ULL));
c->SF = (cast<uint16_t>((f & cast<uint16_t>(128ULL))) != cast<uint16_t>(0ULL));
c->TF = (cast<uint16_t>((f & cast<uint16_t>(256ULL))) != cast<uint16_t>(0ULL));
c->IF = (cast<uint16_t>((f & cast<uint16_t>(512ULL))) != cast<uint16_t>(0ULL));
c->DF = (cast<uint16_t>((f & cast<uint16_t>(1024ULL))) != cast<uint16_t>(0ULL));
c->OF = (cast<uint16_t>((f & cast<uint16_t>(2048ULL))) != cast<uint16_t>(0ULL));
}
}
// tools/cpu/x86/cpu.go:437:1
uint32_t x86_widthMask(int64_t b){
{
{
switch(b){
case cast<int64_t>(1ULL):{
return cast<uint32_t>(255ULL);
break;}
case cast<int64_t>(2ULL):{
return cast<uint32_t>(65535ULL);
break;}
default:{
return cast<uint32_t>(4294967295ULL);
break;}
}}
}
}
// tools/cpu/x86/cpu.go:447:1
uint32_t x86_signMask(int64_t b){
{
{
switch(b){
case cast<int64_t>(1ULL):{
return cast<uint32_t>(128ULL);
break;}
case cast<int64_t>(2ULL):{
return cast<uint32_t>(32768ULL);
break;}
default:{
return cast<uint32_t>(2147483648ULL);
break;}
}}
}
}
// tools/cpu/x86/cpu.go:457:1
uint32_t x86_signExtByte(uint32_t v){
{
return cast<uint32_t>(cast<int32_t>(cast<int8_t>(cast<uint8_t>(v))));
}
}
// tools/cpu/x86/cpu.go:458:1
uint32_t x86_signExtWord(uint32_t v){
{
return cast<uint32_t>(cast<int32_t>(cast<int16_t>(cast<uint16_t>(v))));
}
}
// tools/cpu/x86/cpu.go:461:1
void x86_CPU_setSZP(x86_CPU* c,uint32_t res,int64_t b){
{
uint32_t m = x86_widthMask(b);
c->ZF = (cast<uint32_t>((res & m)) == cast<uint32_t>(0ULL));
c->SF = (cast<uint32_t>((res & x86_signMask(b))) != cast<uint32_t>(0ULL));
c->PF = x86_parity(cast<uint8_t>(res));
}
}
// tools/cpu/x86/cpu.go:468:1
bool x86_parity(uint8_t v){
{
v ^= shr<uint8_t>(v,cast<int64_t>(4ULL));
v ^= shr<uint8_t>(v,cast<int64_t>(2ULL));
v ^= shr<uint8_t>(v,cast<int64_t>(1ULL));
return (cast<uint8_t>((v & cast<uint8_t>(1ULL))) == cast<uint8_t>(0ULL));
}
}
// tools/cpu/x86/cpu.go:476:1
uint32_t x86_CPU_flagsAdd(x86_CPU* c,uint32_t a,uint32_t b,uint32_t cin,int64_t w){
{
uint32_t m = x86_widthMask(w);
uint64_t full = cast<uint64_t>((cast<uint64_t>((cast<uint64_t>(cast<uint32_t>((a & m))) + cast<uint64_t>(cast<uint32_t>((b & m))))) + cast<uint64_t>(cin)));
uint32_t res = cast<uint32_t>((cast<uint32_t>(full) & m));
c->CF = (cast<uint64_t>((full & (cast<uint64_t>((cast<uint64_t>(m) + cast<uint64_t>(1ULL)))))) != cast<uint64_t>(0ULL));
c->AF = (cast<uint32_t>(((cast<uint32_t>((cast<uint32_t>((a ^ b)) ^ res))) & cast<uint32_t>(16ULL))) != cast<uint32_t>(0ULL));
c->OF = (cast<uint32_t>(((cast<uint32_t>((cast<uint32_t>(~(cast<uint32_t>((a ^ b)))) & (cast<uint32_t>((a ^ res)))))) & x86_signMask(w))) != cast<uint32_t>(0ULL));
x86_CPU_setSZP(c,res,w);
return res;
}
}
// tools/cpu/x86/cpu.go:488:1
uint32_t x86_CPU_flagsSub(x86_CPU* c,uint32_t a,uint32_t b,uint32_t cin,int64_t w){
{
uint32_t m = x86_widthMask(w);
uint32_t res = cast<uint32_t>(((cast<uint32_t>((cast<uint32_t>((a - b)) - cin))) & m));
c->CF = (cast<uint64_t>((cast<uint64_t>(cast<uint32_t>((b & m))) + cast<uint64_t>(cin))) > cast<uint64_t>(cast<uint32_t>((a & m))));
c->AF = (cast<uint32_t>(((cast<uint32_t>((cast<uint32_t>((a ^ b)) ^ res))) & cast<uint32_t>(16ULL))) != cast<uint32_t>(0ULL));
c->OF = (cast<uint32_t>(((cast<uint32_t>(((cast<uint32_t>((a ^ b))) & (cast<uint32_t>((a ^ res)))))) & x86_signMask(w))) != cast<uint32_t>(0ULL));
x86_CPU_setSZP(c,res,w);
return res;
}
}
// tools/cpu/x86/cpu.go:499:1
uint32_t x86_CPU_flagsLogic(x86_CPU* c,uint32_t res,int64_t w){
{
res &= x86_widthMask(w);
auto tmp1 = std::make_tuple(false,false,false);
c->CF = std::get<0>(tmp1);
c->OF = std::get<1>(tmp1);
c->AF = std::get<2>(tmp1);
x86_CPU_setSZP(c,res,w);
return res;
}
}
// tools/cpu/x86/cpu.go:507:1
bool x86_CPU_cond(x86_CPU* c,uint8_t cc){
{
{
switch(cc){
case cast<uint8_t>(0ULL):{
return c->OF;
break;}
case cast<uint8_t>(1ULL):{
return (!c->OF);
break;}
case cast<uint8_t>(2ULL):{
return c->CF;
break;}
case cast<uint8_t>(3ULL):{
return (!c->CF);
break;}
case cast<uint8_t>(4ULL):{
return c->ZF;
break;}
case cast<uint8_t>(5ULL):{
return (!c->ZF);
break;}
case cast<uint8_t>(6ULL):{
return (c->CF || c->ZF);
break;}
case cast<uint8_t>(7ULL):{
return ((!c->CF) && (!c->ZF));
break;}
case cast<uint8_t>(8ULL):{
return c->SF;
break;}
case cast<uint8_t>(9ULL):{
return (!c->SF);
break;}
case cast<uint8_t>(10ULL):{
return c->PF;
break;}
case cast<uint8_t>(11ULL):{
return (!c->PF);
break;}
case cast<uint8_t>(12ULL):{
return (c->SF != c->OF);
break;}
case cast<uint8_t>(13ULL):{
return (c->SF == c->OF);
break;}
case cast<uint8_t>(14ULL):{
return (c->ZF || ((c->SF != c->OF)));
break;}
default:{
return ((!c->ZF) && ((c->SF == c->OF)));
break;}
}}
}
}
// tools/cpu/x86/exec.go:17:1
uint32_t x86_CPU_rEA(x86_CPU* c,x86_ea o,int64_t b){
{
if (o.isReg) {
return x86_CPU_getReg(c,o.reg,b);
}
return x86_CPU_memRead(c,o.base,o.off,b);
}
}
// tools/cpu/x86/exec.go:23:1
void x86_CPU_wEA(x86_CPU* c,x86_ea o,int64_t b,uint32_t v){
{
if (o.isReg) {
x86_CPU_setReg(c,o.reg,b,v);
return ;
}
x86_CPU_memWrite(c,o.base,o.off,b,v);
}
}
// tools/cpu/x86/exec.go:32:1
int64_t x86_CPU_osz(x86_CPU* c){
{
if ((c->dOpsize == cast<int64_t>(32ULL))) {
return cast<int64_t>(4ULL);
}
return cast<int64_t>(2ULL);
}
}
// tools/cpu/x86/exec.go:41:1
int64_t x86_CPU_asz(x86_CPU* c){
{
if ((c->dAddrsize == cast<int64_t>(32ULL))) {
return cast<int64_t>(4ULL);
}
return cast<int64_t>(2ULL);
}
}
// tools/cpu/x86/exec.go:50:1
uint32_t x86_CPU_segBaseFor(x86_CPU* c,int64_t def){
{
if ((c->dSeg >= cast<int64_t>(0ULL))) {
return x86_CPU_segBase(c,c->dSeg);
}
return x86_CPU_segBase(c,def);
}
}
// tools/cpu/x86/exec.go:57:1
uint32_t x86_b2u(bool b){
{
if (b) {
return cast<uint32_t>(1ULL);
}
return cast<uint32_t>(0ULL);
}
}
// tools/cpu/x86/exec.go:66:1
uint32_t x86_CPU_inPort(x86_CPU* c,uint16_t port,int64_t size){
{
if (bool(c->PortIn)) {
return cast<uint32_t>((c->PortIn(port,size) & x86_widthMask(size)));
}
return x86_widthMask(size);
}
}
// tools/cpu/x86/exec.go:72:1
void x86_CPU_outPort(x86_CPU* c,uint16_t port,int64_t size,uint32_t v){
{
if (bool(c->PortOut)) {
c->PortOut(port,size,v);
}
}
}
// tools/cpu/x86/exec.go:78:1
int32_t x86_signExtToInt(uint32_t v,int64_t w){
{
{
switch(w){
case cast<int64_t>(1ULL):{
return cast<int32_t>(cast<int8_t>(cast<uint8_t>(v)));
break;}
case cast<int64_t>(2ULL):{
return cast<int32_t>(cast<int16_t>(cast<uint16_t>(v)));
break;}
default:{
return cast<int32_t>(v);
break;}
}}
}
}
// tools/cpu/x86/exec.go:90:1
std::tuple<uint8_t,x86_ea> x86_CPU_modrmE(x86_CPU* c){
{
uint32_t mb = x86_CPU_fetch8(c);
auto tmp2 = std::make_tuple(cast<uint8_t>(shr<uint32_t>(mb,cast<int64_t>(6ULL))),cast<uint8_t>(cast<uint32_t>(((shr<uint32_t>(mb,cast<int64_t>(3ULL))) & cast<uint32_t>(7ULL)))),cast<uint8_t>(cast<uint32_t>((mb & cast<uint32_t>(7ULL)))));
uint8_t mod = std::get<0>(tmp2);
uint8_t reg = std::get<1>(tmp2);
uint8_t rm = std::get<2>(tmp2);
if ((mod == cast<uint8_t>(3ULL))) {
return {reg,x86_ea{true,rm,{},{}}};
}
if ((c->dAddrsize == cast<int64_t>(32ULL))) {
return {reg,x86_CPU_ea32(c,mod,rm)};
}
return {reg,x86_CPU_ea16(c,mod,rm)};
}
}
// tools/cpu/x86/exec.go:102:1
x86_ea x86_CPU_ea16(x86_CPU* c,uint8_t mod,uint8_t rm){
{
uint32_t off={};
int64_t defSeg = cast<int64_t>(3ULL);
{
switch(rm){
case cast<uint8_t>(0ULL):{
off = cast<uint32_t>((x86_CPU_gw(c,cast<int64_t>(3ULL)) + x86_CPU_gw(c,cast<int64_t>(6ULL))));
break;}
case cast<uint8_t>(1ULL):{
off = cast<uint32_t>((x86_CPU_gw(c,cast<int64_t>(3ULL)) + x86_CPU_gw(c,cast<int64_t>(7ULL))));
break;}
case cast<uint8_t>(2ULL):{
off = cast<uint32_t>((x86_CPU_gw(c,cast<int64_t>(5ULL)) + x86_CPU_gw(c,cast<int64_t>(6ULL))));
defSeg = cast<int64_t>(2ULL);
break;}
case cast<uint8_t>(3ULL):{
off = cast<uint32_t>((x86_CPU_gw(c,cast<int64_t>(5ULL)) + x86_CPU_gw(c,cast<int64_t>(7ULL))));
defSeg = cast<int64_t>(2ULL);
break;}
case cast<uint8_t>(4ULL):{
off = x86_CPU_gw(c,cast<int64_t>(6ULL));
break;}
case cast<uint8_t>(5ULL):{
off = x86_CPU_gw(c,cast<int64_t>(7ULL));
break;}
case cast<uint8_t>(6ULL):{
if ((mod == cast<uint8_t>(0ULL))) {
off = x86_CPU_fetch16(c);
}
else {
off = x86_CPU_gw(c,cast<int64_t>(5ULL));
defSeg = cast<int64_t>(2ULL);
}
break;}
case cast<uint8_t>(7ULL):{
off = x86_CPU_gw(c,cast<int64_t>(3ULL));
break;}
}}
{
switch(mod){
case cast<uint8_t>(1ULL):{
off += x86_signExtByte(x86_CPU_fetch8(c));
break;}
case cast<uint8_t>(2ULL):{
off += x86_CPU_fetch16(c);
break;}
}}
int64_t seg = defSeg;
if ((c->dSeg >= cast<int64_t>(0ULL))) {
seg = c->dSeg;
}
return x86_ea{{},{},x86_CPU_segBase(c,seg),cast<uint32_t>((off & cast<uint32_t>(65535ULL)))};
}
}
// tools/cpu/x86/exec.go:143:1
x86_ea x86_CPU_ea32(x86_CPU* c,uint8_t mod,uint8_t rm){
{
uint32_t off={};
int64_t defSeg = cast<int64_t>(3ULL);
{
if ((rm == cast<uint8_t>(4ULL))){
uint32_t sib = x86_CPU_fetch8(c);
auto tmp4 = std::make_tuple(cast<uint8_t>(shr<uint32_t>(sib,cast<int64_t>(6ULL))),cast<uint8_t>(cast<uint32_t>(((shr<uint32_t>(sib,cast<int64_t>(3ULL))) & cast<uint32_t>(7ULL)))),cast<uint8_t>(cast<uint32_t>((sib & cast<uint32_t>(7ULL)))));
uint8_t scale = std::get<0>(tmp4);
uint8_t idx = std::get<1>(tmp4);
uint8_t base = std::get<2>(tmp4);
if ((idx != cast<uint8_t>(4ULL))) {
off += shl<uint32_t>(c->Regs[idx],scale);
}
if (((base == cast<uint8_t>(5ULL)) && (mod == cast<uint8_t>(0ULL)))) {
off += x86_CPU_fetch32(c);
}
else {
off += c->Regs[base];
if (((base == cast<uint8_t>(4ULL)) || (base == cast<uint8_t>(5ULL)))) {
defSeg = cast<int64_t>(2ULL);
}
}
}
else if (((rm == cast<uint8_t>(5ULL)) && (mod == cast<uint8_t>(0ULL)))){
off = x86_CPU_fetch32(c);
}
else {
off = c->Regs[rm];
if (((rm == cast<uint8_t>(4ULL)) || (rm == cast<uint8_t>(5ULL)))) {
defSeg = cast<int64_t>(2ULL);
}
}
}
tmp3:;
{
switch(mod){
case cast<uint8_t>(1ULL):{
off += x86_signExtByte(x86_CPU_fetch8(c));
break;}
case cast<uint8_t>(2ULL):{
off += x86_CPU_fetch32(c);
break;}
}}
int64_t seg = defSeg;
if ((c->dSeg >= cast<int64_t>(0ULL))) {
seg = c->dSeg;
}
return x86_ea{{},{},x86_CPU_segBase(c,seg),off};
}
}
// tools/cpu/x86/exec.go:183:1
void x86_CPU_Step(x86_CPU* c){
{
if (c->Halted) {
return ;
}
if (bool(c->OnStep)) {
c->OnStep(c);
if (c->Halted) {
return ;
}
}
c->ssShadow = false;
c->Steps++;
c->instrIP = c->IP;
c->dSeg = cast<int64_t>(-1ULL);
int64_t alt = cast<int64_t>(32ULL);
if ((c->Mode == cast<int64_t>(1ULL))) {
auto tmp5 = std::make_tuple(cast<int64_t>(32ULL),cast<int64_t>(32ULL));
c->dOpsize = std::get<0>(tmp5);
c->dAddrsize = std::get<1>(tmp5);
alt = cast<int64_t>(16ULL);
}
else {
auto tmp6 = std::make_tuple(cast<int64_t>(16ULL),cast<int64_t>(16ULL));
c->dOpsize = std::get<0>(tmp6);
c->dAddrsize = std::get<1>(tmp6);
}
uint8_t rep={};
uint8_t op={};
{;for (;;){
op = cast<uint8_t>(x86_CPU_fetch8(c));
bool prefix = true;
{
switch(op){
case cast<uint8_t>(38ULL):{
c->dSeg = cast<int64_t>(0ULL);
break;}
case cast<uint8_t>(46ULL):{
c->dSeg = cast<int64_t>(1ULL);
break;}
case cast<uint8_t>(54ULL):{
c->dSeg = cast<int64_t>(2ULL);
break;}
case cast<uint8_t>(62ULL):{
c->dSeg = cast<int64_t>(3ULL);
break;}
case cast<uint8_t>(100ULL):{
c->dSeg = cast<int64_t>(4ULL);
break;}
case cast<uint8_t>(101ULL):{
c->dSeg = cast<int64_t>(5ULL);
break;}
case cast<uint8_t>(102ULL):{
c->dOpsize = alt;
c->Ext386++;
break;}
case cast<uint8_t>(103ULL):{
c->dAddrsize = alt;
c->Ext386++;
break;}
case cast<uint8_t>(240ULL):{
break;}
case cast<uint8_t>(242ULL):{
rep = cast<uint8_t>(242ULL);
break;}
case cast<uint8_t>(243ULL):{
rep = cast<uint8_t>(243ULL);
break;}
default:{
prefix = false;
break;}
}}
if ((!prefix)) {
break;
}
}
}x86_CPU_exec(c,op,rep);
}
}
// tools/cpu/x86/exec.go:248:1
uint64_t x86_CPU_Run(x86_CPU* c,uint64_t maxSteps){
{
uint64_t n={};
{;for (;((!c->Halted) && (n < maxSteps));){
x86_CPU_Step(c);
n++;
}
}return n;
}
}
// tools/cpu/x86/exec.go:259:1
std::tuple<uint32_t,bool> x86_CPU_alu(x86_CPU* c,int64_t idx,uint32_t a,uint32_t b,int64_t w){
{
{
switch(idx){
case cast<int64_t>(0ULL):{
return {x86_CPU_flagsAdd(c,a,b,cast<uint32_t>(0ULL),w),true};
break;}
case cast<int64_t>(1ULL):{
return {x86_CPU_flagsLogic(c,cast<uint32_t>((a | b)),w),true};
break;}
case cast<int64_t>(2ULL):{
return {x86_CPU_flagsAdd(c,a,b,x86_b2u(c->CF),w),true};
break;}
case cast<int64_t>(3ULL):{
return {x86_CPU_flagsSub(c,a,b,x86_b2u(c->CF),w),true};
break;}
case cast<int64_t>(4ULL):{
return {x86_CPU_flagsLogic(c,cast<uint32_t>((a & b)),w),true};
break;}
case cast<int64_t>(5ULL):{
return {x86_CPU_flagsSub(c,a,b,cast<uint32_t>(0ULL),w),true};
break;}
case cast<int64_t>(6ULL):{
return {x86_CPU_flagsLogic(c,cast<uint32_t>((a ^ b)),w),true};
break;}
default:{
return {x86_CPU_flagsSub(c,a,b,cast<uint32_t>(0ULL),w),false};
break;}
}}
}
}
// tools/cpu/x86/exec.go:281:1
void x86_CPU_exec(x86_CPU* c,uint8_t op,uint8_t rep){
{
if (((op < cast<uint8_t>(64ULL)) && (cast<uint8_t>((op & cast<uint8_t>(7ULL))) < cast<uint8_t>(6ULL)))) {
x86_CPU_aluGrid(c,cast<int64_t>(shr<uint8_t>(op,cast<int64_t>(3ULL))),cast<uint8_t>((op & cast<uint8_t>(7ULL))));
return ;
}
{
if (((op >= cast<uint8_t>(64ULL)) && (op <= cast<uint8_t>(71ULL)))){
x86_CPU_setReg(c,cast<uint8_t>((op & cast<uint8_t>(7ULL))),x86_CPU_osz(c),x86_CPU_incDec(c,x86_CPU_getReg(c,cast<uint8_t>((op & cast<uint8_t>(7ULL))),x86_CPU_osz(c)),x86_CPU_osz(c),true));
return ;
}
else if (((op >= cast<uint8_t>(72ULL)) && (op <= cast<uint8_t>(79ULL)))){
x86_CPU_setReg(c,cast<uint8_t>((op & cast<uint8_t>(7ULL))),x86_CPU_osz(c),x86_CPU_incDec(c,x86_CPU_getReg(c,cast<uint8_t>((op & cast<uint8_t>(7ULL))),x86_CPU_osz(c)),x86_CPU_osz(c),false));
return ;
}
else if (((op >= cast<uint8_t>(80ULL)) && (op <= cast<uint8_t>(87ULL)))){
x86_CPU_push(c,x86_CPU_osz(c),x86_CPU_getReg(c,cast<uint8_t>((op & cast<uint8_t>(7ULL))),x86_CPU_osz(c)));
return ;
}
else if (((op >= cast<uint8_t>(88ULL)) && (op <= cast<uint8_t>(95ULL)))){
x86_CPU_setReg(c,cast<uint8_t>((op & cast<uint8_t>(7ULL))),x86_CPU_osz(c),x86_CPU_pop(c,x86_CPU_osz(c)));
return ;
}
else if (((op >= cast<uint8_t>(112ULL)) && (op <= cast<uint8_t>(127ULL)))){
uint32_t rel = x86_signExtByte(x86_CPU_fetch8(c));
if (x86_CPU_cond(c,cast<uint8_t>((op & cast<uint8_t>(15ULL))))) {
c->IP = x86_CPU_ipMask(c,cast<uint32_t>((c->IP + rel)));
}
return ;
}
else if (((op >= cast<uint8_t>(145ULL)) && (op <= cast<uint8_t>(151ULL)))){
uint8_t i = cast<uint8_t>((op & cast<uint8_t>(7ULL)));
int64_t w = x86_CPU_osz(c);
uint32_t t = x86_CPU_getReg(c,cast<uint8_t>(0ULL),w);
x86_CPU_setReg(c,cast<uint8_t>(0ULL),w,x86_CPU_getReg(c,i,w));
x86_CPU_setReg(c,i,w,t);
return ;
}
else if (((op >= cast<uint8_t>(176ULL)) && (op <= cast<uint8_t>(183ULL)))){
x86_CPU_s8(c,cast<uint8_t>((op & cast<uint8_t>(7ULL))),x86_CPU_fetch8(c));
return ;
}
else if (((op >= cast<uint8_t>(184ULL)) && (op <= cast<uint8_t>(191ULL)))){
x86_CPU_setReg(c,cast<uint8_t>((op & cast<uint8_t>(7ULL))),x86_CPU_osz(c),x86_CPU_fetchImm(c));
return ;
}
}
tmp7:;
{
switch(op){
case cast<uint8_t>(6ULL):{
x86_CPU_push(c,x86_CPU_osz(c),cast<uint32_t>(c->Seg[cast<int64_t>(0ULL)]));
break;}
case cast<uint8_t>(7ULL):{
x86_CPU_loadSeg(c,cast<int64_t>(0ULL),cast<uint16_t>(x86_CPU_pop(c,x86_CPU_osz(c))));
break;}
case cast<uint8_t>(14ULL):{
x86_CPU_push(c,x86_CPU_osz(c),cast<uint32_t>(c->Seg[cast<int64_t>(1ULL)]));
break;}
case cast<uint8_t>(22ULL):{
x86_CPU_push(c,x86_CPU_osz(c),cast<uint32_t>(c->Seg[cast<int64_t>(2ULL)]));
break;}
case cast<uint8_t>(23ULL):{
x86_CPU_loadSeg(c,cast<int64_t>(2ULL),cast<uint16_t>(x86_CPU_pop(c,x86_CPU_osz(c))));
c->ssShadow = true;
break;}
case cast<uint8_t>(30ULL):{
x86_CPU_push(c,x86_CPU_osz(c),cast<uint32_t>(c->Seg[cast<int64_t>(3ULL)]));
break;}
case cast<uint8_t>(31ULL):{
x86_CPU_loadSeg(c,cast<int64_t>(3ULL),cast<uint16_t>(x86_CPU_pop(c,x86_CPU_osz(c))));
break;}
case cast<uint8_t>(39ULL):{
x86_CPU_daa(c,false);
break;}
case cast<uint8_t>(47ULL):{
x86_CPU_daa(c,true);
break;}
case cast<uint8_t>(55ULL):{
x86_CPU_aaa(c,false);
break;}
case cast<uint8_t>(63ULL):{
x86_CPU_aaa(c,true);
break;}
case cast<uint8_t>(212ULL):{
x86_CPU_aam(c,cast<uint8_t>(x86_CPU_fetch8(c)));
break;}
case cast<uint8_t>(213ULL):{
x86_CPU_aad(c,cast<uint8_t>(x86_CPU_fetch8(c)));
break;}
case cast<uint8_t>(15ULL):{
c->Ext386++;
x86_CPU_exec0F(c,cast<uint8_t>(x86_CPU_fetch8(c)),rep);
break;}
case cast<uint8_t>(216ULL):case cast<uint8_t>(217ULL):case cast<uint8_t>(218ULL):case cast<uint8_t>(219ULL):case cast<uint8_t>(220ULL):case cast<uint8_t>(221ULL):case cast<uint8_t>(222ULL):case cast<uint8_t>(223ULL):{
x86_CPU_fpuExec(c,op);
break;}
case cast<uint8_t>(96ULL):{
int64_t w = x86_CPU_osz(c);
uint32_t sp = x86_CPU_getReg(c,cast<uint8_t>(4ULL),w);
{auto&& tmp8 = Slice<uint8_t>{cast<uint8_t>(0ULL),cast<uint8_t>(1ULL),cast<uint8_t>(2ULL),cast<uint8_t>(3ULL)};
for(int64_t tmp9=0;tmp9<len(tmp8);++tmp9){
auto r=tmp8[tmp9];x86_CPU_push(c,w,x86_CPU_getReg(c,r,w));
}}
x86_CPU_push(c,w,sp);
{auto&& tmp10 = Slice<uint8_t>{cast<uint8_t>(5ULL),cast<uint8_t>(6ULL),cast<uint8_t>(7ULL)};
for(int64_t tmp11=0;tmp11<len(tmp10);++tmp11){
auto r=tmp10[tmp11];x86_CPU_push(c,w,x86_CPU_getReg(c,r,w));
}}
break;}
case cast<uint8_t>(97ULL):{
int64_t w = x86_CPU_osz(c);
{auto&& tmp12 = Slice<uint8_t>{cast<uint8_t>(7ULL),cast<uint8_t>(6ULL),cast<uint8_t>(5ULL)};
for(int64_t tmp13=0;tmp13<len(tmp12);++tmp13){
auto r=tmp12[tmp13];x86_CPU_setReg(c,r,w,x86_CPU_pop(c,w));
}}
x86_CPU_pop(c,w);
{auto&& tmp14 = Slice<uint8_t>{cast<uint8_t>(3ULL),cast<uint8_t>(2ULL),cast<uint8_t>(1ULL),cast<uint8_t>(0ULL)};
for(int64_t tmp15=0;tmp15<len(tmp14);++tmp15){
auto r=tmp14[tmp15];x86_CPU_setReg(c,r,w,x86_CPU_pop(c,w));
}}
break;}
case cast<uint8_t>(98ULL):{
x86_CPU_modrmE(c);
break;}
case cast<uint8_t>(99ULL):{
x86_CPU_modrmE(c);
break;}
case cast<uint8_t>(108ULL):case cast<uint8_t>(109ULL):case cast<uint8_t>(110ULL):case cast<uint8_t>(111ULL):{
x86_CPU_stringPortOp(c,op,rep);
break;}
case cast<uint8_t>(155ULL):{
break;}
case cast<uint8_t>(241ULL):{
x86_CPU_doInt(c,cast<uint8_t>(1ULL));
break;}
case cast<uint8_t>(104ULL):{
x86_CPU_push(c,x86_CPU_osz(c),x86_CPU_fetchImm(c));
break;}
case cast<uint8_t>(105ULL):{
auto tmp16 = x86_CPU_modrmE(c);
uint8_t reg = std::get<0>(tmp16);
x86_ea o = std::get<1>(tmp16);
int64_t w = x86_CPU_osz(c);
uint32_t src = x86_CPU_rEA(c,o,w);
uint32_t imm = x86_CPU_fetchImm(c);
x86_CPU_setReg(c,reg,w,x86_CPU_imulTrunc(c,src,imm,w));
break;}
case cast<uint8_t>(106ULL):{
x86_CPU_push(c,x86_CPU_osz(c),x86_signExtByte(x86_CPU_fetch8(c)));
break;}
case cast<uint8_t>(107ULL):{
auto tmp17 = x86_CPU_modrmE(c);
uint8_t reg = std::get<0>(tmp17);
x86_ea o = std::get<1>(tmp17);
int64_t w = x86_CPU_osz(c);
uint32_t src = x86_CPU_rEA(c,o,w);
uint32_t imm = x86_signExtByte(x86_CPU_fetch8(c));
x86_CPU_setReg(c,reg,w,x86_CPU_imulTrunc(c,src,imm,w));
break;}
case cast<uint8_t>(128ULL):case cast<uint8_t>(129ULL):case cast<uint8_t>(130ULL):case cast<uint8_t>(131ULL):{
x86_CPU_grp1(c,op);
break;}
case cast<uint8_t>(132ULL):{
auto tmp18 = x86_CPU_modrmE(c);
uint8_t reg = std::get<0>(tmp18);
x86_ea o = std::get<1>(tmp18);
x86_CPU_flagsLogic(c,cast<uint32_t>((x86_CPU_rEA(c,o,cast<int64_t>(1ULL)) & x86_CPU_g8(c,reg))),cast<int64_t>(1ULL));
break;}
case cast<uint8_t>(133ULL):{
auto tmp19 = x86_CPU_modrmE(c);
uint8_t reg = std::get<0>(tmp19);
x86_ea o = std::get<1>(tmp19);
int64_t w = x86_CPU_osz(c);
x86_CPU_flagsLogic(c,cast<uint32_t>((x86_CPU_rEA(c,o,w) & x86_CPU_getReg(c,reg,w))),w);
break;}
case cast<uint8_t>(134ULL):{
auto tmp20 = x86_CPU_modrmE(c);
uint8_t reg = std::get<0>(tmp20);
x86_ea o = std::get<1>(tmp20);
uint32_t t = x86_CPU_rEA(c,o,cast<int64_t>(1ULL));
x86_CPU_wEA(c,o,cast<int64_t>(1ULL),x86_CPU_g8(c,reg));
x86_CPU_s8(c,reg,t);
break;}
case cast<uint8_t>(135ULL):{
auto tmp21 = x86_CPU_modrmE(c);
uint8_t reg = std::get<0>(tmp21);
x86_ea o = std::get<1>(tmp21);
int64_t w = x86_CPU_osz(c);
uint32_t t = x86_CPU_rEA(c,o,w);
x86_CPU_wEA(c,o,w,x86_CPU_getReg(c,reg,w));
x86_CPU_setReg(c,reg,w,t);
break;}
case cast<uint8_t>(136ULL):{
auto tmp22 = x86_CPU_modrmE(c);
uint8_t reg = std::get<0>(tmp22);
x86_ea o = std::get<1>(tmp22);
x86_CPU_wEA(c,o,cast<int64_t>(1ULL),x86_CPU_g8(c,reg));
break;}
case cast<uint8_t>(137ULL):{
auto tmp23 = x86_CPU_modrmE(c);
uint8_t reg = std::get<0>(tmp23);
x86_ea o = std::get<1>(tmp23);
int64_t w = x86_CPU_osz(c);
x86_CPU_wEA(c,o,w,x86_CPU_getReg(c,reg,w));
break;}
case cast<uint8_t>(138ULL):{
auto tmp24 = x86_CPU_modrmE(c);
uint8_t reg = std::get<0>(tmp24);
x86_ea o = std::get<1>(tmp24);
x86_CPU_s8(c,reg,x86_CPU_rEA(c,o,cast<int64_t>(1ULL)));
break;}
case cast<uint8_t>(139ULL):{
auto tmp25 = x86_CPU_modrmE(c);
uint8_t reg = std::get<0>(tmp25);
x86_ea o = std::get<1>(tmp25);
int64_t w = x86_CPU_osz(c);
x86_CPU_setReg(c,reg,w,x86_CPU_rEA(c,o,w));
break;}
case cast<uint8_t>(140ULL):{
auto tmp26 = x86_CPU_modrmE(c);
uint8_t reg = std::get<0>(tmp26);
x86_ea o = std::get<1>(tmp26);
x86_CPU_wEA(c,o,cast<int64_t>(2ULL),cast<uint32_t>(c->Seg[cast<uint8_t>((reg & cast<uint8_t>(7ULL)))]));
break;}
case cast<uint8_t>(141ULL):{
auto tmp27 = x86_CPU_modrmE(c);
uint8_t reg = std::get<0>(tmp27);
x86_ea o = std::get<1>(tmp27);
x86_CPU_setReg(c,reg,x86_CPU_osz(c),o.off);
break;}
case cast<uint8_t>(142ULL):{
auto tmp28 = x86_CPU_modrmE(c);
uint8_t reg = std::get<0>(tmp28);
x86_ea o = std::get<1>(tmp28);
x86_CPU_loadSeg(c,cast<int64_t>(cast<uint8_t>((reg & cast<uint8_t>(7ULL)))),cast<uint16_t>(x86_CPU_rEA(c,o,cast<int64_t>(2ULL))));
if ((cast<uint8_t>((reg & cast<uint8_t>(7ULL))) == cast<uint8_t>(2ULL))) {
c->ssShadow = true;
}
break;}
case cast<uint8_t>(143ULL):{
auto tmp29 = x86_CPU_modrmE(c);
x86_ea o = std::get<1>(tmp29);
x86_CPU_wEA(c,o,x86_CPU_osz(c),x86_CPU_pop(c,x86_CPU_osz(c)));
break;}
case cast<uint8_t>(144ULL):{
break;}
case cast<uint8_t>(152ULL):{
if ((c->dOpsize == cast<int64_t>(32ULL))) {
c->Regs[cast<int64_t>(0ULL)] = cast<uint32_t>(cast<int32_t>(cast<int16_t>(cast<uint16_t>(x86_CPU_gw(c,cast<int64_t>(0ULL))))));
}
else {
x86_CPU_s16(c,cast<uint8_t>(0ULL),x86_signExtByte(x86_CPU_g8(c,cast<uint8_t>(0ULL))));
}
break;}
case cast<uint8_t>(153ULL):{
if ((c->dOpsize == cast<int64_t>(32ULL))) {
if ((cast<uint32_t>((c->Regs[cast<int64_t>(0ULL)] & cast<uint32_t>(2147483648ULL))) != cast<uint32_t>(0ULL))) {
c->Regs[cast<int64_t>(2ULL)] = cast<uint32_t>(4294967295ULL);
}
else {
c->Regs[cast<int64_t>(2ULL)] = cast<uint32_t>(0ULL);
}
}
else if ((cast<uint32_t>((x86_CPU_gw(c,cast<int64_t>(0ULL)) & cast<uint32_t>(32768ULL))) != cast<uint32_t>(0ULL))) {
x86_CPU_s16(c,cast<uint8_t>(2ULL),cast<uint32_t>(65535ULL));
}
else {
x86_CPU_s16(c,cast<uint8_t>(2ULL),cast<uint32_t>(0ULL));
}
break;}
case cast<uint8_t>(154ULL):{
uint32_t off = x86_CPU_fetchImm(c);
uint32_t seg = x86_CPU_fetch16(c);
x86_CPU_push(c,x86_CPU_osz(c),cast<uint32_t>(c->Seg[cast<int64_t>(1ULL)]));
x86_CPU_push(c,x86_CPU_osz(c),c->IP);
x86_CPU_loadSeg(c,cast<int64_t>(1ULL),cast<uint16_t>(seg));
c->IP = x86_CPU_ipMask(c,off);
break;}
case cast<uint8_t>(156ULL):{
x86_CPU_push(c,x86_CPU_osz(c),cast<uint32_t>(x86_CPU_EFlags(c)));
break;}
case cast<uint8_t>(157ULL):{
x86_CPU_SetEFlags(c,cast<uint16_t>(x86_CPU_pop(c,x86_CPU_osz(c))));
break;}
case cast<uint8_t>(158ULL):{
x86_CPU_SetEFlags(c,cast<uint16_t>((((x86_CPU_EFlags(c) & ~(cast<uint16_t>(255ULL)))) | cast<uint16_t>(x86_CPU_g8(c,cast<uint8_t>(4ULL))))));
break;}
case cast<uint8_t>(159ULL):{
x86_CPU_s8(c,cast<uint8_t>(4ULL),cast<uint32_t>(cast<uint8_t>(x86_CPU_EFlags(c))));
break;}
case cast<uint8_t>(160ULL):{
x86_CPU_s8(c,cast<uint8_t>(0ULL),x86_CPU_memRead(c,x86_CPU_segBaseFor(c,cast<int64_t>(3ULL)),x86_CPU_moffs(c),cast<int64_t>(1ULL)));
break;}
case cast<uint8_t>(161ULL):{
int64_t w = x86_CPU_osz(c);
x86_CPU_setReg(c,cast<uint8_t>(0ULL),w,x86_CPU_memRead(c,x86_CPU_segBaseFor(c,cast<int64_t>(3ULL)),x86_CPU_moffs(c),w));
break;}
case cast<uint8_t>(162ULL):{
x86_CPU_memWrite(c,x86_CPU_segBaseFor(c,cast<int64_t>(3ULL)),x86_CPU_moffs(c),cast<int64_t>(1ULL),x86_CPU_g8(c,cast<uint8_t>(0ULL)));
break;}
case cast<uint8_t>(163ULL):{
int64_t w = x86_CPU_osz(c);
x86_CPU_memWrite(c,x86_CPU_segBaseFor(c,cast<int64_t>(3ULL)),x86_CPU_moffs(c),w,x86_CPU_getReg(c,cast<uint8_t>(0ULL),w));
break;}
case cast<uint8_t>(164ULL):case cast<uint8_t>(165ULL):case cast<uint8_t>(166ULL):case cast<uint8_t>(167ULL):case cast<uint8_t>(170ULL):case cast<uint8_t>(171ULL):case cast<uint8_t>(172ULL):case cast<uint8_t>(173ULL):case cast<uint8_t>(174ULL):case cast<uint8_t>(175ULL):{
x86_CPU_stringOp(c,op,rep);
break;}
case cast<uint8_t>(168ULL):{
x86_CPU_flagsLogic(c,cast<uint32_t>((x86_CPU_g8(c,cast<uint8_t>(0ULL)) & x86_CPU_fetch8(c))),cast<int64_t>(1ULL));
break;}
case cast<uint8_t>(169ULL):{
int64_t w = x86_CPU_osz(c);
x86_CPU_flagsLogic(c,cast<uint32_t>((x86_CPU_getReg(c,cast<uint8_t>(0ULL),w) & x86_CPU_fetchImm(c))),w);
break;}
case cast<uint8_t>(192ULL):{
x86_CPU_grp2(c,cast<int64_t>(1ULL),[&]()->uint32_t{
return x86_CPU_fetch8(c);
}
);
break;}
case cast<uint8_t>(193ULL):{
x86_CPU_grp2(c,x86_CPU_osz(c),[&]()->uint32_t{
return x86_CPU_fetch8(c);
}
);
break;}
case cast<uint8_t>(194ULL):{
uint32_t n = x86_CPU_fetch16(c);
c->IP = x86_CPU_ipMask(c,x86_CPU_pop(c,x86_CPU_osz(c)));
x86_CPU_spAdd(c,n);
break;}
case cast<uint8_t>(195ULL):{
c->IP = x86_CPU_ipMask(c,x86_CPU_pop(c,x86_CPU_osz(c)));
break;}
case cast<uint8_t>(196ULL):{
auto tmp30 = x86_CPU_modrmE(c);
uint8_t reg = std::get<0>(tmp30);
x86_ea o = std::get<1>(tmp30);
x86_CPU_setReg(c,reg,x86_CPU_osz(c),x86_CPU_rEA(c,o,cast<int64_t>(2ULL)));
x86_CPU_loadSeg(c,cast<int64_t>(0ULL),cast<uint16_t>(x86_CPU_memRead(c,o.base,cast<uint32_t>((o.off + cast<uint32_t>(2ULL))),cast<int64_t>(2ULL))));
break;}
case cast<uint8_t>(197ULL):{
auto tmp31 = x86_CPU_modrmE(c);
uint8_t reg = std::get<0>(tmp31);
x86_ea o = std::get<1>(tmp31);
x86_CPU_setReg(c,reg,x86_CPU_osz(c),x86_CPU_rEA(c,o,cast<int64_t>(2ULL)));
x86_CPU_loadSeg(c,cast<int64_t>(3ULL),cast<uint16_t>(x86_CPU_memRead(c,o.base,cast<uint32_t>((o.off + cast<uint32_t>(2ULL))),cast<int64_t>(2ULL))));
break;}
case cast<uint8_t>(198ULL):{
auto tmp32 = x86_CPU_modrmE(c);
x86_ea o = std::get<1>(tmp32);
x86_CPU_wEA(c,o,cast<int64_t>(1ULL),x86_CPU_fetch8(c));
break;}
case cast<uint8_t>(199ULL):{
auto tmp33 = x86_CPU_modrmE(c);
x86_ea o = std::get<1>(tmp33);
x86_CPU_wEA(c,o,x86_CPU_osz(c),x86_CPU_fetchImm(c));
break;}
case cast<uint8_t>(200ULL):{
uint32_t sz = x86_CPU_fetch16(c);
uint32_t lvl = cast<uint32_t>(cast<uint32_t>((x86_CPU_fetch8(c) & cast<uint32_t>(31ULL))));
int64_t w = x86_CPU_osz(c);
x86_CPU_push(c,w,x86_CPU_getReg(c,cast<uint8_t>(5ULL),w));
uint32_t frame = x86_CPU_getReg(c,cast<uint8_t>(4ULL),w);
{uint32_t i = cast<uint32_t>(1ULL);for (;(i < lvl);i++){
x86_CPU_setReg(c,cast<uint8_t>(5ULL),w,cast<uint32_t>((x86_CPU_getReg(c,cast<uint8_t>(5ULL),w) - cast<uint32_t>(w))));
x86_CPU_push(c,w,x86_CPU_getReg(c,cast<uint8_t>(5ULL),w));
}
}if ((lvl > cast<uint32_t>(0ULL))) {
x86_CPU_push(c,w,frame);
}
x86_CPU_setReg(c,cast<uint8_t>(5ULL),w,frame);
x86_CPU_setReg(c,cast<uint8_t>(4ULL),w,cast<uint32_t>((x86_CPU_getReg(c,cast<uint8_t>(4ULL),w) - sz)));
break;}
case cast<uint8_t>(201ULL):{
int64_t w = x86_CPU_osz(c);
x86_CPU_setReg(c,cast<uint8_t>(4ULL),w,x86_CPU_getReg(c,cast<uint8_t>(5ULL),w));
x86_CPU_setReg(c,cast<uint8_t>(5ULL),w,x86_CPU_pop(c,w));
break;}
case cast<uint8_t>(202ULL):{
uint32_t n = x86_CPU_fetch16(c);
c->IP = x86_CPU_ipMask(c,x86_CPU_pop(c,x86_CPU_osz(c)));
x86_CPU_loadSeg(c,cast<int64_t>(1ULL),cast<uint16_t>(x86_CPU_pop(c,x86_CPU_osz(c))));
x86_CPU_spAdd(c,n);
break;}
case cast<uint8_t>(203ULL):{
c->IP = x86_CPU_ipMask(c,x86_CPU_pop(c,x86_CPU_osz(c)));
x86_CPU_loadSeg(c,cast<int64_t>(1ULL),cast<uint16_t>(x86_CPU_pop(c,x86_CPU_osz(c))));
break;}
case cast<uint8_t>(204ULL):{
x86_CPU_doInt(c,cast<uint8_t>(3ULL));
break;}
case cast<uint8_t>(205ULL):{
x86_CPU_doInt(c,cast<uint8_t>(x86_CPU_fetch8(c)));
break;}
case cast<uint8_t>(206ULL):{
if (c->OF) {
x86_CPU_doInt(c,cast<uint8_t>(4ULL));
}
break;}
case cast<uint8_t>(207ULL):{
c->IP = x86_CPU_ipMask(c,x86_CPU_pop(c,x86_CPU_osz(c)));
x86_CPU_loadSeg(c,cast<int64_t>(1ULL),cast<uint16_t>(x86_CPU_pop(c,x86_CPU_osz(c))));
x86_CPU_SetEFlags(c,cast<uint16_t>(x86_CPU_pop(c,x86_CPU_osz(c))));
break;}
case cast<uint8_t>(208ULL):{
x86_CPU_grp2(c,cast<int64_t>(1ULL),[&]()->uint32_t{
return cast<uint32_t>(1ULL);
}
);
break;}
case cast<uint8_t>(209ULL):{
x86_CPU_grp2(c,x86_CPU_osz(c),[&]()->uint32_t{
return cast<uint32_t>(1ULL);
}
);
break;}
case cast<uint8_t>(210ULL):{
x86_CPU_grp2(c,cast<int64_t>(1ULL),[&]()->uint32_t{
return x86_CPU_g8(c,cast<uint8_t>(1ULL));
}
);
break;}
case cast<uint8_t>(211ULL):{
x86_CPU_grp2(c,x86_CPU_osz(c),[&]()->uint32_t{
return x86_CPU_g8(c,cast<uint8_t>(1ULL));
}
);
break;}
case cast<uint8_t>(215ULL):{
int64_t aw = x86_CPU_asz(c);
x86_CPU_s8(c,cast<uint8_t>(0ULL),x86_CPU_memRead(c,x86_CPU_segBaseFor(c,cast<int64_t>(3ULL)),cast<uint32_t>((x86_CPU_getReg(c,cast<uint8_t>(3ULL),aw) + x86_CPU_g8(c,cast<uint8_t>(0ULL)))),cast<int64_t>(1ULL)));
break;}
case cast<uint8_t>(224ULL):case cast<uint8_t>(225ULL):case cast<uint8_t>(226ULL):{
uint32_t rel = x86_signExtByte(x86_CPU_fetch8(c));
int64_t aw = x86_CPU_asz(c);
x86_CPU_setReg(c,cast<uint8_t>(1ULL),aw,cast<uint32_t>((x86_CPU_getReg(c,cast<uint8_t>(1ULL),aw) - cast<uint32_t>(1ULL))));
bool take = (x86_CPU_getReg(c,cast<uint8_t>(1ULL),aw) != cast<uint32_t>(0ULL));
if ((op == cast<uint8_t>(225ULL))) {
take = (take && c->ZF);
}
else if ((op == cast<uint8_t>(224ULL))) {
take = (take && (!c->ZF));
}
if (take) {
c->IP = x86_CPU_ipMask(c,cast<uint32_t>((c->IP + rel)));
}
break;}
case cast<uint8_t>(227ULL):{
uint32_t rel = x86_signExtByte(x86_CPU_fetch8(c));
if ((x86_CPU_getReg(c,cast<uint8_t>(1ULL),x86_CPU_asz(c)) == cast<uint32_t>(0ULL))) {
c->IP = x86_CPU_ipMask(c,cast<uint32_t>((c->IP + rel)));
}
break;}
case cast<uint8_t>(232ULL):{
uint32_t rel = x86_CPU_fetchImm(c);
if ((c->dOpsize != cast<int64_t>(32ULL))) {
rel = x86_signExtWord(rel);
}
x86_CPU_push(c,x86_CPU_osz(c),c->IP);
c->IP = x86_CPU_ipMask(c,cast<uint32_t>((c->IP + rel)));
break;}
case cast<uint8_t>(233ULL):{
uint32_t rel = x86_CPU_fetchImm(c);
if ((c->dOpsize != cast<int64_t>(32ULL))) {
rel = x86_signExtWord(rel);
}
c->IP = x86_CPU_ipMask(c,cast<uint32_t>((c->IP + rel)));
break;}
case cast<uint8_t>(234ULL):{
uint32_t off = x86_CPU_fetchImm(c);
uint32_t seg = x86_CPU_fetch16(c);
x86_CPU_loadSeg(c,cast<int64_t>(1ULL),cast<uint16_t>(seg));
c->IP = x86_CPU_ipMask(c,off);
break;}
case cast<uint8_t>(235ULL):{
uint32_t rel = x86_signExtByte(x86_CPU_fetch8(c));
c->IP = x86_CPU_ipMask(c,cast<uint32_t>((c->IP + rel)));
break;}
case cast<uint8_t>(228ULL):{
x86_CPU_s8(c,cast<uint8_t>(0ULL),x86_CPU_inPort(c,cast<uint16_t>(x86_CPU_fetch8(c)),cast<int64_t>(1ULL)));
break;}
case cast<uint8_t>(229ULL):{
x86_CPU_setReg(c,cast<uint8_t>(0ULL),x86_CPU_osz(c),x86_CPU_inPort(c,cast<uint16_t>(x86_CPU_fetch8(c)),x86_CPU_osz(c)));
break;}
case cast<uint8_t>(230ULL):{
x86_CPU_outPort(c,cast<uint16_t>(x86_CPU_fetch8(c)),cast<int64_t>(1ULL),x86_CPU_g8(c,cast<uint8_t>(0ULL)));
break;}
case cast<uint8_t>(231ULL):{
x86_CPU_outPort(c,cast<uint16_t>(x86_CPU_fetch8(c)),x86_CPU_osz(c),x86_CPU_getReg(c,cast<uint8_t>(0ULL),x86_CPU_osz(c)));
break;}
case cast<uint8_t>(236ULL):{
x86_CPU_s8(c,cast<uint8_t>(0ULL),x86_CPU_inPort(c,cast<uint16_t>(x86_CPU_gw(c,cast<int64_t>(2ULL))),cast<int64_t>(1ULL)));
break;}
case cast<uint8_t>(237ULL):{
x86_CPU_setReg(c,cast<uint8_t>(0ULL),x86_CPU_osz(c),x86_CPU_inPort(c,cast<uint16_t>(x86_CPU_gw(c,cast<int64_t>(2ULL))),x86_CPU_osz(c)));
break;}
case cast<uint8_t>(238ULL):{
x86_CPU_outPort(c,cast<uint16_t>(x86_CPU_gw(c,cast<int64_t>(2ULL))),cast<int64_t>(1ULL),x86_CPU_g8(c,cast<uint8_t>(0ULL)));
break;}
case cast<uint8_t>(239ULL):{
x86_CPU_outPort(c,cast<uint16_t>(x86_CPU_gw(c,cast<int64_t>(2ULL))),x86_CPU_osz(c),x86_CPU_getReg(c,cast<uint8_t>(0ULL),x86_CPU_osz(c)));
break;}
case cast<uint8_t>(244ULL):{
x86_CPU_Halt(c,std::string("HLT at %s",9),x86_CPU_at(c));
break;}
case cast<uint8_t>(245ULL):{
c->CF = (!c->CF);
break;}
case cast<uint8_t>(246ULL):{
x86_CPU_grp3(c,cast<int64_t>(1ULL));
break;}
case cast<uint8_t>(247ULL):{
x86_CPU_grp3(c,x86_CPU_osz(c));
break;}
case cast<uint8_t>(248ULL):{
c->CF = false;
break;}
case cast<uint8_t>(249ULL):{
c->CF = true;
break;}
case cast<uint8_t>(250ULL):{
c->IF = false;
break;}
case cast<uint8_t>(251ULL):{
c->IF = true;
break;}
case cast<uint8_t>(252ULL):{
c->DF = false;
break;}
case cast<uint8_t>(253ULL):{
c->DF = true;
break;}
case cast<uint8_t>(254ULL):{
auto tmp34 = x86_CPU_modrmE(c);
uint8_t reg = std::get<0>(tmp34);
x86_ea o = std::get<1>(tmp34);
uint32_t v = x86_CPU_rEA(c,o,cast<int64_t>(1ULL));
x86_CPU_wEA(c,o,cast<int64_t>(1ULL),x86_CPU_incDec(c,v,cast<int64_t>(1ULL),(reg == cast<uint8_t>(0ULL))));
break;}
case cast<uint8_t>(255ULL):{
x86_CPU_grp5(c);
break;}
default:{
x86_CPU_Halt(c,std::string("unimplemented opcode $%02X at %s",32),op,x86_CPU_at(c));
break;}
}}
}
}
// tools/cpu/x86/exec.go:675:1
uint32_t x86_CPU_moffs(x86_CPU* c){
{
if ((c->dAddrsize == cast<int64_t>(32ULL))) {
return x86_CPU_fetch32(c);
}
return x86_CPU_fetch16(c);
}
}
// tools/cpu/x86/exec.go:683:1
void x86_CPU_aluGrid(x86_CPU* c,int64_t idx,uint8_t z){
{
{
switch(z){
case cast<uint8_t>(0ULL):{
auto tmp35 = x86_CPU_modrmE(c);
uint8_t reg = std::get<0>(tmp35);
x86_ea o = std::get<1>(tmp35);
auto tmp36 = x86_CPU_alu(c,idx,x86_CPU_rEA(c,o,cast<int64_t>(1ULL)),x86_CPU_g8(c,reg),cast<int64_t>(1ULL));
uint32_t res = std::get<0>(tmp36);
bool wr = std::get<1>(tmp36);
if (wr) {
x86_CPU_wEA(c,o,cast<int64_t>(1ULL),res);
}
break;}
case cast<uint8_t>(1ULL):{
auto tmp37 = x86_CPU_modrmE(c);
uint8_t reg = std::get<0>(tmp37);
x86_ea o = std::get<1>(tmp37);
int64_t w = x86_CPU_osz(c);
auto tmp38 = x86_CPU_alu(c,idx,x86_CPU_rEA(c,o,w),x86_CPU_getReg(c,reg,w),w);
uint32_t res = std::get<0>(tmp38);
bool wr = std::get<1>(tmp38);
if (wr) {
x86_CPU_wEA(c,o,w,res);
}
break;}
case cast<uint8_t>(2ULL):{
auto tmp39 = x86_CPU_modrmE(c);
uint8_t reg = std::get<0>(tmp39);
x86_ea o = std::get<1>(tmp39);
auto tmp40 = x86_CPU_alu(c,idx,x86_CPU_g8(c,reg),x86_CPU_rEA(c,o,cast<int64_t>(1ULL)),cast<int64_t>(1ULL));
uint32_t res = std::get<0>(tmp40);
bool wr = std::get<1>(tmp40);
if (wr) {
x86_CPU_s8(c,reg,res);
}
break;}
case cast<uint8_t>(3ULL):{
auto tmp41 = x86_CPU_modrmE(c);
uint8_t reg = std::get<0>(tmp41);
x86_ea o = std::get<1>(tmp41);
int64_t w = x86_CPU_osz(c);
auto tmp42 = x86_CPU_alu(c,idx,x86_CPU_getReg(c,reg,w),x86_CPU_rEA(c,o,w),w);
uint32_t res = std::get<0>(tmp42);
bool wr = std::get<1>(tmp42);
if (wr) {
x86_CPU_setReg(c,reg,w,res);
}
break;}
case cast<uint8_t>(4ULL):{
auto tmp43 = x86_CPU_alu(c,idx,x86_CPU_g8(c,cast<uint8_t>(0ULL)),x86_CPU_fetch8(c),cast<int64_t>(1ULL));
uint32_t res = std::get<0>(tmp43);
bool wr = std::get<1>(tmp43);
if (wr) {
x86_CPU_s8(c,cast<uint8_t>(0ULL),res);
}
break;}
default:{
int64_t w = x86_CPU_osz(c);
auto tmp44 = x86_CPU_alu(c,idx,x86_CPU_getReg(c,cast<uint8_t>(0ULL),w),x86_CPU_fetchImm(c),w);
uint32_t res = std::get<0>(tmp44);
bool wr = std::get<1>(tmp44);
if (wr) {
x86_CPU_setReg(c,cast<uint8_t>(0ULL),w,res);
}
break;}
}}
}
}
// tools/cpu/x86/exec.go:726:1
uint32_t x86_CPU_incDec(x86_CPU* c,uint32_t v,int64_t w,bool inc){
{
bool saved = c->CF;
uint32_t res={};
if (inc) {
res = x86_CPU_flagsAdd(c,v,cast<uint32_t>(1ULL),cast<uint32_t>(0ULL),w);
}
else {
res = x86_CPU_flagsSub(c,v,cast<uint32_t>(1ULL),cast<uint32_t>(0ULL),w);
}
c->CF = saved;
return res;
}
}
// tools/cpu/x86/exec.go:742:1
void x86_CPU_doInt(x86_CPU* c,uint8_t n){
{
if ((bool(c->IntHook) && c->IntHook(c,n))) {
return ;
}
if ((c->Mode == cast<int64_t>(1ULL))) {
x86_CPU_Halt(c,std::string("unhandled INT $%02X in protected mode at %08X",45),n,c->instrIP);
return ;
}
x86_CPU_dispatchIVT(c,n);
}
}
// tools/cpu/x86/exec.go:761:1
void x86_CPU_divErr(x86_CPU* c){
{
c->IP = c->instrIP;
x86_CPU_dispatchIVT(c,cast<uint8_t>(0ULL));
}
}
// tools/cpu/x86/exec.go:768:1
void x86_CPU_dispatchIVT(x86_CPU* c,uint8_t n){
{
x86_CPU_push16(c,cast<uint32_t>(x86_CPU_EFlags(c)));
auto tmp45 = std::make_tuple(false,false);
c->IF = std::get<0>(tmp45);
c->TF = std::get<1>(tmp45);
x86_CPU_push16(c,cast<uint32_t>(c->Seg[cast<int64_t>(1ULL)]));
x86_CPU_push16(c,c->IP);
c->IP = x86_CPU_rd16(c,cast<uint32_t>((cast<uint32_t>(n) * cast<uint32_t>(4ULL))));
c->Seg[cast<int64_t>(1ULL)] = cast<uint16_t>(x86_CPU_rd16(c,cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(n) * cast<uint32_t>(4ULL))) + cast<uint32_t>(2ULL)))));
}
}
// tools/cpu/x86/exec.go:781:1
bool x86_CPU_Interrupt(x86_CPU* c,uint8_t n){
{
if ((((!c->IF) || c->Halted) || c->ssShadow)) {
return false;
}
x86_CPU_dispatchIVT(c,n);
return true;
}
}
// tools/cpu/x86/exec.go:795:1
bool x86_CPU_InterruptPM(x86_CPU* c,uint16_t cs,uint32_t eip){
{
if ((((!c->IF) || c->Halted) || c->ssShadow)) {
return false;
}
x86_CPU_push(c,cast<int64_t>(4ULL),cast<uint32_t>(x86_CPU_EFlags(c)));
auto tmp46 = std::make_tuple(false,false);
c->IF = std::get<0>(tmp46);
c->TF = std::get<1>(tmp46);
x86_CPU_push(c,cast<int64_t>(4ULL),cast<uint32_t>(c->Seg[cast<int64_t>(1ULL)]));
x86_CPU_push(c,cast<int64_t>(4ULL),c->IP);
x86_CPU_loadSeg(c,cast<int64_t>(1ULL),cs);
c->IP = eip;
return true;
}
}
// tools/cpu/x86/exec2.go:8:1
void x86_CPU_grp1(x86_CPU* c,uint8_t op){
{
auto tmp47 = x86_CPU_modrmE(c);
uint8_t reg = std::get<0>(tmp47);
x86_ea o = std::get<1>(tmp47);
int64_t w={};
uint32_t imm={};
{
switch(op){
case cast<uint8_t>(128ULL):case cast<uint8_t>(130ULL):{
w = cast<int64_t>(1ULL);
imm = x86_CPU_fetch8(c);
break;}
case cast<uint8_t>(129ULL):{
w = x86_CPU_osz(c);
imm = x86_CPU_fetchImm(c);
break;}
default:{
w = x86_CPU_osz(c);
imm = cast<uint32_t>((x86_signExtByte(x86_CPU_fetch8(c)) & x86_widthMask(w)));
break;}
}}
auto tmp48 = x86_CPU_alu(c,cast<int64_t>(reg),x86_CPU_rEA(c,o,w),imm,w);
uint32_t res = std::get<0>(tmp48);
bool wr = std::get<1>(tmp48);
if (wr) {
x86_CPU_wEA(c,o,w,res);
}
}
}
// tools/cpu/x86/exec2.go:31:1
void x86_CPU_grp2(x86_CPU* c,int64_t w,std::function<uint32_t()> count){
{
auto tmp49 = x86_CPU_modrmE(c);
uint8_t reg = std::get<0>(tmp49);
x86_ea o = std::get<1>(tmp49);
uint32_t cnt = count();
x86_CPU_wEA(c,o,w,x86_CPU_shiftOp(c,cast<int64_t>(reg),x86_CPU_rEA(c,o,w),cnt,w));
}
}
// tools/cpu/x86/exec2.go:39:1
uint32_t x86_CPU_shiftOp(x86_CPU* c,int64_t idx,uint32_t val,uint32_t cnt,int64_t w){
{
auto tmp50 = std::make_tuple(x86_widthMask(w),x86_signMask(w),cast<uint32_t>(cast<int64_t>((cast<int64_t>(8ULL) * w))));
uint32_t m = std::get<0>(tmp50);
uint32_t sb = std::get<1>(tmp50);
uint32_t bits = std::get<2>(tmp50);
cnt &= cast<uint32_t>(31ULL);
val &= m;
if ((cnt == cast<uint32_t>(0ULL))) {
return val;
}
{
switch(idx){
case cast<int64_t>(4ULL):case cast<int64_t>(6ULL):{
uint32_t res = cast<uint32_t>(((shl<uint32_t>(val,cnt)) & m));
if ((cnt <= bits)) {
c->CF = (cast<uint32_t>(((shr<uint32_t>(val,(cast<uint32_t>((bits - cnt))))) & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL));
}
else {
c->CF = false;
}
if ((cnt == cast<uint32_t>(1ULL))) {
c->OF = (((cast<uint32_t>((res & sb)) != cast<uint32_t>(0ULL))) != c->CF);
}
x86_CPU_setSZP(c,res,w);
return res;
break;}
case cast<int64_t>(5ULL):{
c->CF = (cast<uint32_t>(((shr<uint32_t>(val,(cast<uint32_t>((cnt - cast<uint32_t>(1ULL)))))) & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL));
uint32_t res = cast<uint32_t>(((shr<uint32_t>(val,cnt)) & m));
if ((cnt == cast<uint32_t>(1ULL))) {
c->OF = (cast<uint32_t>((val & sb)) != cast<uint32_t>(0ULL));
}
x86_CPU_setSZP(c,res,w);
return res;
break;}
case cast<int64_t>(7ULL):{
int32_t sv = x86_signExtToInt(val,w);
c->CF = (cast<uint32_t>(((shr<uint32_t>(cast<uint32_t>(sv),(cast<uint32_t>((cnt - cast<uint32_t>(1ULL)))))) & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL));
uint32_t res = cast<uint32_t>((cast<uint32_t>(shr<int32_t>(sv,cnt)) & m));
if ((cnt == cast<uint32_t>(1ULL))) {
c->OF = false;
}
x86_CPU_setSZP(c,res,w);
return res;
break;}
case cast<int64_t>(0ULL):{
uint32_t n = modi<uint32_t>(cnt,bits);
uint32_t res = val;
if ((n != cast<uint32_t>(0ULL))) {
res = cast<uint32_t>(((cast<uint32_t>(((shl<uint32_t>(val,n)) | (shr<uint32_t>(val,(cast<uint32_t>((bits - n)))))))) & m));
}
c->CF = (cast<uint32_t>((res & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL));
if ((cnt == cast<uint32_t>(1ULL))) {
c->OF = (((cast<uint32_t>((res & sb)) != cast<uint32_t>(0ULL))) != c->CF);
}
return res;
break;}
case cast<int64_t>(1ULL):{
uint32_t n = modi<uint32_t>(cnt,bits);
uint32_t res = val;
if ((n != cast<uint32_t>(0ULL))) {
res = cast<uint32_t>(((cast<uint32_t>(((shr<uint32_t>(val,n)) | (shl<uint32_t>(val,(cast<uint32_t>((bits - n)))))))) & m));
}
c->CF = (cast<uint32_t>((res & sb)) != cast<uint32_t>(0ULL));
if ((cnt == cast<uint32_t>(1ULL))) {
c->OF = (((cast<uint32_t>((res & sb)) != cast<uint32_t>(0ULL))) != ((cast<uint32_t>((res & (shr<uint32_t>(sb,cast<int64_t>(1ULL))))) != cast<uint32_t>(0ULL))));
}
return res;
break;}
case cast<int64_t>(2ULL):{
auto tmp51 = std::make_tuple(val,x86_b2u(c->CF));
uint32_t res = std::get<0>(tmp51);
uint32_t cf = std::get<1>(tmp51);
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < cnt);i++){
uint32_t nc = cast<uint32_t>(((shr<uint32_t>(res,(cast<uint32_t>((bits - cast<uint32_t>(1ULL)))))) & cast<uint32_t>(1ULL)));
res = cast<uint32_t>(((cast<uint32_t>(((shl<uint32_t>(res,cast<int64_t>(1ULL))) | cf))) & m));
cf = nc;
}
}c->CF = (cf != cast<uint32_t>(0ULL));
if ((cnt == cast<uint32_t>(1ULL))) {
c->OF = (((cast<uint32_t>((res & sb)) != cast<uint32_t>(0ULL))) != c->CF);
}
return res;
break;}
default:{
auto tmp52 = std::make_tuple(val,x86_b2u(c->CF));
uint32_t res = std::get<0>(tmp52);
uint32_t cf = std::get<1>(tmp52);
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < cnt);i++){
uint32_t nc = cast<uint32_t>((res & cast<uint32_t>(1ULL)));
res = cast<uint32_t>(((shr<uint32_t>(res,cast<int64_t>(1ULL))) | (shl<uint32_t>(cf,(cast<uint32_t>((bits - cast<uint32_t>(1ULL))))))));
cf = nc;
}
}res &= m;
c->CF = (cf != cast<uint32_t>(0ULL));
if ((cnt == cast<uint32_t>(1ULL))) {
c->OF = (((cast<uint32_t>((res & sb)) != cast<uint32_t>(0ULL))) != ((cast<uint32_t>((res & (shr<uint32_t>(sb,cast<int64_t>(1ULL))))) != cast<uint32_t>(0ULL))));
}
return res;
break;}
}}
}
}
// tools/cpu/x86/exec2.go:127:1
void x86_CPU_grp3(x86_CPU* c,int64_t w){
{
auto tmp53 = x86_CPU_modrmE(c);
uint8_t reg = std::get<0>(tmp53);
x86_ea o = std::get<1>(tmp53);
{
switch(reg){
case cast<uint8_t>(0ULL):case cast<uint8_t>(1ULL):{
uint32_t imm={};
if ((w == cast<int64_t>(1ULL))) {
imm = x86_CPU_fetch8(c);
}
else {
imm = x86_CPU_fetchImm(c);
}
x86_CPU_flagsLogic(c,cast<uint32_t>((x86_CPU_rEA(c,o,w) & imm)),w);
break;}
case cast<uint8_t>(2ULL):{
x86_CPU_wEA(c,o,w,cast<uint32_t>((cast<uint32_t>(~x86_CPU_rEA(c,o,w)) & x86_widthMask(w))));
break;}
case cast<uint8_t>(3ULL):{
uint32_t v = x86_CPU_rEA(c,o,w);
uint32_t res = x86_CPU_flagsSub(c,cast<uint32_t>(0ULL),v,cast<uint32_t>(0ULL),w);
c->CF = (cast<uint32_t>((v & x86_widthMask(w))) != cast<uint32_t>(0ULL));
x86_CPU_wEA(c,o,w,res);
break;}
case cast<uint8_t>(4ULL):{
x86_CPU_mulOp(c,o,w,false);
break;}
case cast<uint8_t>(5ULL):{
x86_CPU_mulOp(c,o,w,true);
break;}
case cast<uint8_t>(6ULL):{
x86_CPU_divOp(c,o,w,false);
break;}
case cast<uint8_t>(7ULL):{
x86_CPU_divOp(c,o,w,true);
break;}
}}
}
}
// tools/cpu/x86/exec2.go:159:1
void x86_CPU_grp5(x86_CPU* c){
{
auto tmp54 = x86_CPU_modrmE(c);
uint8_t reg = std::get<0>(tmp54);
x86_ea o = std::get<1>(tmp54);
int64_t w = x86_CPU_osz(c);
{
switch(reg){
case cast<uint8_t>(0ULL):{
x86_CPU_wEA(c,o,w,x86_CPU_incDec(c,x86_CPU_rEA(c,o,w),w,true));
break;}
case cast<uint8_t>(1ULL):{
x86_CPU_wEA(c,o,w,x86_CPU_incDec(c,x86_CPU_rEA(c,o,w),w,false));
break;}
case cast<uint8_t>(2ULL):{
uint32_t t = x86_CPU_rEA(c,o,w);
x86_CPU_push(c,x86_CPU_osz(c),c->IP);
c->IP = x86_CPU_ipMask(c,t);
break;}
case cast<uint8_t>(3ULL):{
uint32_t off = x86_CPU_rEA(c,o,w);
uint32_t seg = x86_CPU_memRead(c,o.base,cast<uint32_t>((o.off + cast<uint32_t>(w))),cast<int64_t>(2ULL));
x86_CPU_push(c,x86_CPU_osz(c),cast<uint32_t>(c->Seg[cast<int64_t>(1ULL)]));
x86_CPU_push(c,x86_CPU_osz(c),c->IP);
x86_CPU_loadSeg(c,cast<int64_t>(1ULL),cast<uint16_t>(seg));
c->IP = x86_CPU_ipMask(c,off);
break;}
case cast<uint8_t>(4ULL):{
c->IP = x86_CPU_ipMask(c,x86_CPU_rEA(c,o,w));
break;}
case cast<uint8_t>(5ULL):{
uint32_t off = x86_CPU_rEA(c,o,w);
uint32_t seg = x86_CPU_memRead(c,o.base,cast<uint32_t>((o.off + cast<uint32_t>(w))),cast<int64_t>(2ULL));
x86_CPU_loadSeg(c,cast<int64_t>(1ULL),cast<uint16_t>(seg));
c->IP = x86_CPU_ipMask(c,off);
break;}
case cast<uint8_t>(6ULL):{
x86_CPU_push(c,w,x86_CPU_rEA(c,o,w));
break;}
default:{
x86_CPU_Halt(c,std::string("grp5 /7 (invalid) at %s",23),x86_CPU_at(c));
break;}
}}
}
}
// tools/cpu/x86/exec2.go:193:1
void x86_CPU_mulOp(x86_CPU* c,x86_ea o,int64_t w,bool signed_){
{
uint32_t src = x86_CPU_rEA(c,o,w);
{
switch(w){
case cast<int64_t>(1ULL):{
uint32_t p={};
if (signed_) {
p = cast<uint32_t>(cast<int32_t>((cast<int32_t>(cast<int8_t>(cast<uint8_t>(x86_CPU_g8(c,cast<uint8_t>(0ULL))))) * cast<int32_t>(cast<int8_t>(cast<uint8_t>(src))))));
}
else {
p = cast<uint32_t>(((cast<uint32_t>((x86_CPU_g8(c,cast<uint8_t>(0ULL)) & cast<uint32_t>(255ULL)))) * (cast<uint32_t>((src & cast<uint32_t>(255ULL))))));
}
x86_CPU_s16(c,cast<uint8_t>(0ULL),cast<uint32_t>((p & cast<uint32_t>(65535ULL))));
if (signed_) {
c->CF = (cast<int32_t>(cast<int16_t>(cast<uint16_t>(p))) != cast<int32_t>(cast<int8_t>(cast<uint8_t>(p))));
}
else {
c->CF = (cast<uint32_t>(((shr<uint32_t>(p,cast<int64_t>(8ULL))) & cast<uint32_t>(255ULL))) != cast<uint32_t>(0ULL));
}
c->OF = c->CF;
break;}
case cast<int64_t>(2ULL):{
uint32_t p={};
if (signed_) {
p = cast<uint32_t>(cast<int32_t>((cast<int32_t>(cast<int16_t>(cast<uint16_t>(x86_CPU_gw(c,cast<int64_t>(0ULL))))) * cast<int32_t>(cast<int16_t>(cast<uint16_t>(src))))));
}
else {
p = cast<uint32_t>((x86_CPU_gw(c,cast<int64_t>(0ULL)) * (cast<uint32_t>((src & cast<uint32_t>(65535ULL))))));
}
x86_CPU_s16(c,cast<uint8_t>(0ULL),cast<uint32_t>((p & cast<uint32_t>(65535ULL))));
x86_CPU_s16(c,cast<uint8_t>(2ULL),cast<uint32_t>(((shr<uint32_t>(p,cast<int64_t>(16ULL))) & cast<uint32_t>(65535ULL))));
if (signed_) {
c->CF = (cast<int32_t>(p) != cast<int32_t>(cast<int16_t>(cast<uint16_t>(p))));
}
else {
c->CF = (shr<uint32_t>(p,cast<int64_t>(16ULL)) != cast<uint32_t>(0ULL));
}
c->OF = c->CF;
break;}
default:{
uint64_t p={};
if (signed_) {
p = cast<uint64_t>(cast<int64_t>((cast<int64_t>(cast<int32_t>(c->Regs[cast<int64_t>(0ULL)])) * cast<int64_t>(cast<int32_t>(src)))));
}
else {
p = cast<uint64_t>((cast<uint64_t>(c->Regs[cast<int64_t>(0ULL)]) * cast<uint64_t>(src)));
}
c->Regs[cast<int64_t>(0ULL)] = cast<uint32_t>(p);
c->Regs[cast<int64_t>(2ULL)] = cast<uint32_t>(shr<uint64_t>(p,cast<int64_t>(32ULL)));
if (signed_) {
c->CF = (cast<int64_t>(p) != cast<int64_t>(cast<int32_t>(cast<uint32_t>(p))));
}
else {
c->CF = (shr<uint64_t>(p,cast<int64_t>(32ULL)) != cast<uint64_t>(0ULL));
}
c->OF = c->CF;
break;}
}}
}
}
// tools/cpu/x86/exec2.go:249:1
void x86_CPU_divOp(x86_CPU* c,x86_ea o,int64_t w,bool signed_){
{
uint32_t src = x86_CPU_rEA(c,o,w);
{
switch(w){
case cast<int64_t>(1ULL):{
if ((cast<uint32_t>((src & cast<uint32_t>(255ULL))) == cast<uint32_t>(0ULL))) {
x86_CPU_divErr(c);
return ;
}
if (signed_) {
int16_t num = cast<int16_t>(cast<uint16_t>(x86_CPU_gw(c,cast<int64_t>(0ULL))));
int16_t d = cast<int16_t>(cast<int8_t>(cast<uint8_t>(src)));
auto tmp55 = std::make_tuple(divi<int16_t>(num,d),modi<int16_t>(num,d));
int16_t q = std::get<0>(tmp55);
int16_t r = std::get<1>(tmp55);
if (((q < cast<int16_t>(-128ULL)) || (q > cast<int16_t>(127ULL)))) {
x86_CPU_divErr(c);
return ;
}
x86_CPU_s8(c,cast<uint8_t>(0ULL),cast<uint32_t>(cast<uint8_t>(cast<int8_t>(q))));
x86_CPU_s8(c,cast<uint8_t>(4ULL),cast<uint32_t>(cast<uint8_t>(cast<int8_t>(r))));
}
else {
uint32_t num = x86_CPU_gw(c,cast<int64_t>(0ULL));
uint32_t d = cast<uint32_t>((src & cast<uint32_t>(255ULL)));
auto tmp56 = std::make_tuple(divi<uint32_t>(num,d),modi<uint32_t>(num,d));
uint32_t q = std::get<0>(tmp56);
uint32_t r = std::get<1>(tmp56);
if ((q > cast<uint32_t>(255ULL))) {
x86_CPU_divErr(c);
return ;
}
x86_CPU_s8(c,cast<uint8_t>(0ULL),q);
x86_CPU_s8(c,cast<uint8_t>(4ULL),r);
}
break;}
case cast<int64_t>(2ULL):{
if ((cast<uint32_t>((src & cast<uint32_t>(65535ULL))) == cast<uint32_t>(0ULL))) {
x86_CPU_divErr(c);
return ;
}
uint32_t num = cast<uint32_t>((shl<uint32_t>(x86_CPU_gw(c,cast<int64_t>(2ULL)),cast<int64_t>(16ULL)) | x86_CPU_gw(c,cast<int64_t>(0ULL))));
if (signed_) {
int32_t n = cast<int32_t>(num);
int32_t d = cast<int32_t>(cast<int16_t>(cast<uint16_t>(src)));
auto tmp57 = std::make_tuple(divi<int32_t>(n,d),modi<int32_t>(n,d));
int32_t q = std::get<0>(tmp57);
int32_t r = std::get<1>(tmp57);
if (((q < cast<int32_t>(-32768ULL)) || (q > cast<int32_t>(32767ULL)))) {
x86_CPU_divErr(c);
return ;
}
x86_CPU_s16(c,cast<uint8_t>(0ULL),cast<uint32_t>(cast<uint16_t>(cast<int16_t>(q))));
x86_CPU_s16(c,cast<uint8_t>(2ULL),cast<uint32_t>(cast<uint16_t>(cast<int16_t>(r))));
}
else {
uint32_t d = cast<uint32_t>((src & cast<uint32_t>(65535ULL)));
auto tmp58 = std::make_tuple(divi<uint32_t>(num,d),modi<uint32_t>(num,d));
uint32_t q = std::get<0>(tmp58);
uint32_t r = std::get<1>(tmp58);
if ((q > cast<uint32_t>(65535ULL))) {
x86_CPU_divErr(c);
return ;
}
x86_CPU_s16(c,cast<uint8_t>(0ULL),q);
x86_CPU_s16(c,cast<uint8_t>(2ULL),r);
}
break;}
default:{
if ((src == cast<uint32_t>(0ULL))) {
x86_CPU_divErr(c);
return ;
}
uint64_t num = cast<uint64_t>((shl<uint64_t>(cast<uint64_t>(c->Regs[cast<int64_t>(2ULL)]),cast<int64_t>(32ULL)) | cast<uint64_t>(c->Regs[cast<int64_t>(0ULL)])));
if (signed_) {
int64_t n = cast<int64_t>(num);
int64_t d = cast<int64_t>(cast<int32_t>(src));
c->Regs[cast<int64_t>(0ULL)] = cast<uint32_t>(divi<int64_t>(n,d));
c->Regs[cast<int64_t>(2ULL)] = cast<uint32_t>(modi<int64_t>(n,d));
}
else {
uint64_t d = cast<uint64_t>(src);
c->Regs[cast<int64_t>(0ULL)] = cast<uint32_t>(divi<uint64_t>(num,d));
c->Regs[cast<int64_t>(2ULL)] = cast<uint32_t>(modi<uint64_t>(num,d));
}
break;}
}}
}
}
// tools/cpu/x86/exec2.go:325:1
uint32_t x86_CPU_imulTrunc(x86_CPU* c,uint32_t a,uint32_t b,int64_t w){
{
int64_t full={};
if ((w == cast<int64_t>(2ULL))) {
full = cast<int64_t>((cast<int64_t>(cast<int16_t>(cast<uint16_t>(a))) * cast<int64_t>(cast<int16_t>(cast<uint16_t>(b)))));
}
else {
full = cast<int64_t>((cast<int64_t>(cast<int32_t>(a)) * cast<int64_t>(cast<int32_t>(b))));
}
uint32_t res = cast<uint32_t>((cast<uint32_t>(full) & x86_widthMask(w)));
c->CF = (full != cast<int64_t>(x86_signExtToInt(res,w)));
c->OF = c->CF;
return res;
}
}
// tools/cpu/x86/exec2.go:342:1
void x86_CPU_stringOp(x86_CPU* c,uint8_t op,uint8_t rep){
{
int64_t w = cast<int64_t>(1ULL);
{
switch(op){
case cast<uint8_t>(165ULL):case cast<uint8_t>(167ULL):case cast<uint8_t>(171ULL):case cast<uint8_t>(173ULL):case cast<uint8_t>(175ULL):{
w = x86_CPU_osz(c);
break;}
}}
int64_t aw = x86_CPU_asz(c);
uint32_t dsBase = x86_CPU_segBaseFor(c,cast<int64_t>(3ULL));
uint32_t esBase = x86_CPU_segBase(c,cast<int64_t>(0ULL));
auto step = [&]()->void{
{
switch(op){
case cast<uint8_t>(164ULL):case cast<uint8_t>(165ULL):{
x86_CPU_memWrite(c,esBase,x86_CPU_getReg(c,cast<uint8_t>(7ULL),aw),w,x86_CPU_memRead(c,dsBase,x86_CPU_getReg(c,cast<uint8_t>(6ULL),aw),w));
x86_CPU_advSI(c,w,aw);
x86_CPU_advDI(c,w,aw);
break;}
case cast<uint8_t>(166ULL):case cast<uint8_t>(167ULL):{
x86_CPU_flagsSub(c,x86_CPU_memRead(c,dsBase,x86_CPU_getReg(c,cast<uint8_t>(6ULL),aw),w),x86_CPU_memRead(c,esBase,x86_CPU_getReg(c,cast<uint8_t>(7ULL),aw),w),cast<uint32_t>(0ULL),w);
x86_CPU_advSI(c,w,aw);
x86_CPU_advDI(c,w,aw);
break;}
case cast<uint8_t>(170ULL):case cast<uint8_t>(171ULL):{
x86_CPU_memWrite(c,esBase,x86_CPU_getReg(c,cast<uint8_t>(7ULL),aw),w,x86_CPU_getReg(c,cast<uint8_t>(0ULL),w));
x86_CPU_advDI(c,w,aw);
break;}
case cast<uint8_t>(172ULL):case cast<uint8_t>(173ULL):{
x86_CPU_setReg(c,cast<uint8_t>(0ULL),w,x86_CPU_memRead(c,dsBase,x86_CPU_getReg(c,cast<uint8_t>(6ULL),aw),w));
x86_CPU_advSI(c,w,aw);
break;}
case cast<uint8_t>(174ULL):case cast<uint8_t>(175ULL):{
x86_CPU_flagsSub(c,x86_CPU_getReg(c,cast<uint8_t>(0ULL),w),x86_CPU_memRead(c,esBase,x86_CPU_getReg(c,cast<uint8_t>(7ULL),aw),w),cast<uint32_t>(0ULL),w);
x86_CPU_advDI(c,w,aw);
break;}
}}
}
;
if ((rep == cast<uint8_t>(0ULL))) {
step();
return ;
}
bool isCmp = ((((op == cast<uint8_t>(166ULL)) || (op == cast<uint8_t>(167ULL))) || (op == cast<uint8_t>(174ULL))) || (op == cast<uint8_t>(175ULL)));
{;for (;(x86_CPU_getReg(c,cast<uint8_t>(1ULL),aw) != cast<uint32_t>(0ULL));){
step();
x86_CPU_setReg(c,cast<uint8_t>(1ULL),aw,cast<uint32_t>((x86_CPU_getReg(c,cast<uint8_t>(1ULL),aw) - cast<uint32_t>(1ULL))));
if (isCmp) {
if (((rep == cast<uint8_t>(243ULL)) && (!c->ZF))) {
break;
}
if (((rep == cast<uint8_t>(242ULL)) && c->ZF)) {
break;
}
}
}
}}
}
// tools/cpu/x86/exec2.go:392:1
void x86_CPU_stringPortOp(x86_CPU* c,uint8_t op,uint8_t rep){
{
int64_t w = cast<int64_t>(1ULL);
if (((op == cast<uint8_t>(109ULL)) || (op == cast<uint8_t>(111ULL)))) {
w = x86_CPU_osz(c);
}
int64_t aw = x86_CPU_asz(c);
uint32_t dsBase = x86_CPU_segBaseFor(c,cast<int64_t>(3ULL));
uint32_t esBase = x86_CPU_segBase(c,cast<int64_t>(0ULL));
uint16_t port = x86_CPU_Reg16(c,cast<int64_t>(2ULL));
auto step = [&]()->void{
{
switch(op){
case cast<uint8_t>(108ULL):case cast<uint8_t>(109ULL):{
x86_CPU_memWrite(c,esBase,x86_CPU_getReg(c,cast<uint8_t>(7ULL),aw),w,x86_CPU_inPort(c,port,w));
x86_CPU_advDI(c,w,aw);
break;}
default:{
x86_CPU_outPort(c,port,w,x86_CPU_memRead(c,dsBase,x86_CPU_getReg(c,cast<uint8_t>(6ULL),aw),w));
x86_CPU_advSI(c,w,aw);
break;}
}}
}
;
if ((rep == cast<uint8_t>(0ULL))) {
step();
return ;
}
{;for (;(x86_CPU_getReg(c,cast<uint8_t>(1ULL),aw) != cast<uint32_t>(0ULL));){
step();
x86_CPU_setReg(c,cast<uint8_t>(1ULL),aw,cast<uint32_t>((x86_CPU_getReg(c,cast<uint8_t>(1ULL),aw) - cast<uint32_t>(1ULL))));
}
}}
}
// tools/cpu/x86/exec2.go:423:1
void x86_CPU_advSI(x86_CPU* c,int64_t w,int64_t aw){
{
if (c->DF) {
x86_CPU_setReg(c,cast<uint8_t>(6ULL),aw,cast<uint32_t>((x86_CPU_getReg(c,cast<uint8_t>(6ULL),aw) - cast<uint32_t>(w))));
}
else {
x86_CPU_setReg(c,cast<uint8_t>(6ULL),aw,cast<uint32_t>((x86_CPU_getReg(c,cast<uint8_t>(6ULL),aw) + cast<uint32_t>(w))));
}
}
}
// tools/cpu/x86/exec2.go:430:1
void x86_CPU_advDI(x86_CPU* c,int64_t w,int64_t aw){
{
if (c->DF) {
x86_CPU_setReg(c,cast<uint8_t>(7ULL),aw,cast<uint32_t>((x86_CPU_getReg(c,cast<uint8_t>(7ULL),aw) - cast<uint32_t>(w))));
}
else {
x86_CPU_setReg(c,cast<uint8_t>(7ULL),aw,cast<uint32_t>((x86_CPU_getReg(c,cast<uint8_t>(7ULL),aw) + cast<uint32_t>(w))));
}
}
}
// tools/cpu/x86/exec2.go:439:1
void x86_CPU_daa(x86_CPU* c,bool sub_){
{
uint32_t al = x86_CPU_g8(c,cast<uint8_t>(0ULL));
auto tmp59 = std::make_tuple(al,c->CF);
uint32_t oldAL = std::get<0>(tmp59);
bool oldCF = std::get<1>(tmp59);
c->CF = false;
if (((cast<uint32_t>((al & cast<uint32_t>(15ULL))) > cast<uint32_t>(9ULL)) || c->AF)) {
if (sub_) {
al -= cast<uint32_t>(6ULL);
}
else {
al += cast<uint32_t>(6ULL);
}
c->AF = true;
}
else {
c->AF = false;
}
if (((oldAL > cast<uint32_t>(153ULL)) || oldCF)) {
if (sub_) {
al -= cast<uint32_t>(96ULL);
}
else {
al += cast<uint32_t>(96ULL);
}
c->CF = true;
}
al &= cast<uint32_t>(255ULL);
x86_CPU_s8(c,cast<uint8_t>(0ULL),al);
x86_CPU_setSZP(c,al,cast<int64_t>(1ULL));
}
}
// tools/cpu/x86/exec2.go:467:1
void x86_CPU_aaa(x86_CPU* c,bool sub_){
{
auto tmp60 = std::make_tuple(x86_CPU_g8(c,cast<uint8_t>(0ULL)),x86_CPU_g8(c,cast<uint8_t>(4ULL)));
uint32_t al = std::get<0>(tmp60);
uint32_t ah = std::get<1>(tmp60);
if (((cast<uint32_t>((al & cast<uint32_t>(15ULL))) > cast<uint32_t>(9ULL)) || c->AF)) {
if (sub_) {
al -= cast<uint32_t>(6ULL);
ah--;
}
else {
al += cast<uint32_t>(6ULL);
ah++;
}
auto tmp61 = std::make_tuple(true,true);
c->AF = std::get<0>(tmp61);
c->CF = std::get<1>(tmp61);
}
else {
auto tmp62 = std::make_tuple(false,false);
c->AF = std::get<0>(tmp62);
c->CF = std::get<1>(tmp62);
}
x86_CPU_s8(c,cast<uint8_t>(0ULL),cast<uint32_t>((al & cast<uint32_t>(15ULL))));
x86_CPU_s8(c,cast<uint8_t>(4ULL),cast<uint32_t>((ah & cast<uint32_t>(255ULL))));
}
}
// tools/cpu/x86/exec2.go:486:1
void x86_CPU_aam(x86_CPU* c,uint8_t base){
{
if ((base == cast<uint8_t>(0ULL))) {
x86_CPU_Halt(c,std::string("AAM by zero at %s",17),x86_CPU_at(c));
return ;
}
uint32_t al = x86_CPU_g8(c,cast<uint8_t>(0ULL));
x86_CPU_s8(c,cast<uint8_t>(4ULL),divi<uint32_t>(al,cast<uint32_t>(base)));
x86_CPU_s8(c,cast<uint8_t>(0ULL),modi<uint32_t>(al,cast<uint32_t>(base)));
x86_CPU_setSZP(c,x86_CPU_g8(c,cast<uint8_t>(0ULL)),cast<int64_t>(1ULL));
}
}
// tools/cpu/x86/exec2.go:496:1
void x86_CPU_aad(x86_CPU* c,uint8_t base){
{
uint32_t res = cast<uint32_t>(((cast<uint32_t>((x86_CPU_g8(c,cast<uint8_t>(0ULL)) + cast<uint32_t>((x86_CPU_g8(c,cast<uint8_t>(4ULL)) * cast<uint32_t>(base)))))) & cast<uint32_t>(255ULL)));
x86_CPU_s8(c,cast<uint8_t>(0ULL),res);
x86_CPU_s8(c,cast<uint8_t>(4ULL),cast<uint32_t>(0ULL));
x86_CPU_setSZP(c,res,cast<int64_t>(1ULL));
}
}
// tools/cpu/x86/exec2.go:506:1
void x86_CPU_exec0F(x86_CPU* c,uint8_t op,uint8_t rep){
{
{
if (((op >= cast<uint8_t>(128ULL)) && (op <= cast<uint8_t>(143ULL)))){
uint32_t rel = x86_CPU_fetchImm(c);
if ((c->dOpsize != cast<int64_t>(32ULL))) {
rel = x86_signExtWord(rel);
}
if (x86_CPU_cond(c,cast<uint8_t>((op & cast<uint8_t>(15ULL))))) {
c->IP = x86_CPU_ipMask(c,cast<uint32_t>((c->IP + rel)));
}
return ;
}
else if (((op >= cast<uint8_t>(144ULL)) && (op <= cast<uint8_t>(159ULL)))){
auto tmp64 = x86_CPU_modrmE(c);
x86_ea o = std::get<1>(tmp64);
x86_CPU_wEA(c,o,cast<int64_t>(1ULL),x86_b2u(x86_CPU_cond(c,cast<uint8_t>((op & cast<uint8_t>(15ULL))))));
return ;
}
else if (((op >= cast<uint8_t>(64ULL)) && (op <= cast<uint8_t>(79ULL)))){
auto tmp65 = x86_CPU_modrmE(c);
uint8_t reg = std::get<0>(tmp65);
x86_ea o = std::get<1>(tmp65);
int64_t w = x86_CPU_osz(c);
uint32_t v = x86_CPU_rEA(c,o,w);
if (x86_CPU_cond(c,cast<uint8_t>((op & cast<uint8_t>(15ULL))))) {
x86_CPU_setReg(c,reg,w,v);
}
return ;
}
}
tmp63:;
{
switch(op){
case cast<uint8_t>(8ULL):case cast<uint8_t>(9ULL):{
break;}
case cast<uint8_t>(49ULL):{
uint64_t tsc={};
if (bool(c->TSCFunc)) {
tsc = c->TSCFunc();
}
else {
uint64_t mul = c->TSCMul;
if ((mul == cast<uint64_t>(0ULL))) {
mul = cast<uint64_t>(1ULL);
}
tsc = cast<uint64_t>((c->Steps * mul));
}
c->Regs[cast<int64_t>(0ULL)] = cast<uint32_t>(tsc);
c->Regs[cast<int64_t>(2ULL)] = cast<uint32_t>(shr<uint64_t>(tsc,cast<int64_t>(32ULL)));
break;}
case cast<uint8_t>(176ULL):case cast<uint8_t>(177ULL):{
int64_t w = cast<int64_t>(1ULL);
if ((op == cast<uint8_t>(177ULL))) {
w = x86_CPU_osz(c);
}
auto tmp66 = x86_CPU_modrmE(c);
uint8_t reg = std::get<0>(tmp66);
x86_ea o = std::get<1>(tmp66);
uint32_t dst = x86_CPU_rEA(c,o,w);
x86_CPU_flagsSub(c,x86_CPU_getReg(c,cast<uint8_t>(0ULL),w),dst,cast<uint32_t>(0ULL),w);
if (c->ZF) {
x86_CPU_wEA(c,o,w,x86_CPU_getReg(c,reg,w));
}
else {
x86_CPU_setReg(c,cast<uint8_t>(0ULL),w,dst);
}
break;}
case cast<uint8_t>(192ULL):case cast<uint8_t>(193ULL):{
int64_t w = cast<int64_t>(1ULL);
if ((op == cast<uint8_t>(193ULL))) {
w = x86_CPU_osz(c);
}
auto tmp67 = x86_CPU_modrmE(c);
uint8_t reg = std::get<0>(tmp67);
x86_ea o = std::get<1>(tmp67);
uint32_t dst = x86_CPU_rEA(c,o,w);
uint32_t sum = x86_CPU_flagsAdd(c,dst,x86_CPU_getReg(c,reg,w),cast<uint32_t>(0ULL),w);
x86_CPU_setReg(c,reg,w,dst);
x86_CPU_wEA(c,o,w,sum);
break;}
case cast<uint8_t>(24ULL):case cast<uint8_t>(25ULL):case cast<uint8_t>(26ULL):case cast<uint8_t>(27ULL):case cast<uint8_t>(28ULL):case cast<uint8_t>(29ULL):case cast<uint8_t>(30ULL):case cast<uint8_t>(31ULL):{
x86_CPU_modrmE(c);
break;}
case cast<uint8_t>(174ULL):{
auto tmp68 = x86_CPU_modrmE(c);
uint8_t reg = std::get<0>(tmp68);
x86_ea o = std::get<1>(tmp68);
{
if (o.isReg){
}
else if ((reg == cast<uint8_t>(2ULL))){
c->MXCSR = x86_CPU_rEA(c,o,cast<int64_t>(4ULL));
}
else if ((reg == cast<uint8_t>(3ULL))){
x86_CPU_wEA(c,o,cast<int64_t>(4ULL),c->MXCSR);
}
else if ((reg == cast<uint8_t>(7ULL))){
}
else {
x86_CPU_Halt(c,std::string("unimplemented 0F AE /%d (FXSAVE/FXRSTOR) at %s",46),reg,x86_CPU_at(c));
}
}
tmp69:;
break;}
case cast<uint8_t>(160ULL):{
x86_CPU_push(c,x86_CPU_osz(c),cast<uint32_t>(c->Seg[cast<int64_t>(4ULL)]));
break;}
case cast<uint8_t>(161ULL):{
x86_CPU_loadSeg(c,cast<int64_t>(4ULL),cast<uint16_t>(x86_CPU_pop(c,x86_CPU_osz(c))));
break;}
case cast<uint8_t>(168ULL):{
x86_CPU_push(c,x86_CPU_osz(c),cast<uint32_t>(c->Seg[cast<int64_t>(5ULL)]));
break;}
case cast<uint8_t>(169ULL):{
x86_CPU_loadSeg(c,cast<int64_t>(5ULL),cast<uint16_t>(x86_CPU_pop(c,x86_CPU_osz(c))));
break;}
case cast<uint8_t>(164ULL):{
auto tmp70 = x86_CPU_modrmE(c);
uint8_t reg = std::get<0>(tmp70);
x86_ea o = std::get<1>(tmp70);
x86_CPU_doubleShift(c,true,o,reg,x86_CPU_fetch8(c),x86_CPU_osz(c));
break;}
case cast<uint8_t>(165ULL):{
auto tmp71 = x86_CPU_modrmE(c);
uint8_t reg = std::get<0>(tmp71);
x86_ea o = std::get<1>(tmp71);
x86_CPU_doubleShift(c,true,o,reg,x86_CPU_g8(c,cast<uint8_t>(1ULL)),x86_CPU_osz(c));
break;}
case cast<uint8_t>(172ULL):{
auto tmp72 = x86_CPU_modrmE(c);
uint8_t reg = std::get<0>(tmp72);
x86_ea o = std::get<1>(tmp72);
x86_CPU_doubleShift(c,false,o,reg,x86_CPU_fetch8(c),x86_CPU_osz(c));
break;}
case cast<uint8_t>(173ULL):{
auto tmp73 = x86_CPU_modrmE(c);
uint8_t reg = std::get<0>(tmp73);
x86_ea o = std::get<1>(tmp73);
x86_CPU_doubleShift(c,false,o,reg,x86_CPU_g8(c,cast<uint8_t>(1ULL)),x86_CPU_osz(c));
break;}
case cast<uint8_t>(163ULL):{
auto tmp74 = x86_CPU_modrmE(c);
uint8_t reg = std::get<0>(tmp74);
x86_ea o = std::get<1>(tmp74);
x86_CPU_bitOp(c,o,x86_CPU_getReg(c,reg,x86_CPU_osz(c)),x86_CPU_osz(c),cast<int64_t>(0ULL),true);
break;}
case cast<uint8_t>(171ULL):{
auto tmp75 = x86_CPU_modrmE(c);
uint8_t reg = std::get<0>(tmp75);
x86_ea o = std::get<1>(tmp75);
x86_CPU_bitOp(c,o,x86_CPU_getReg(c,reg,x86_CPU_osz(c)),x86_CPU_osz(c),cast<int64_t>(1ULL),true);
break;}
case cast<uint8_t>(179ULL):{
auto tmp76 = x86_CPU_modrmE(c);
uint8_t reg = std::get<0>(tmp76);
x86_ea o = std::get<1>(tmp76);
x86_CPU_bitOp(c,o,x86_CPU_getReg(c,reg,x86_CPU_osz(c)),x86_CPU_osz(c),cast<int64_t>(2ULL),true);
break;}
case cast<uint8_t>(187ULL):{
auto tmp77 = x86_CPU_modrmE(c);
uint8_t reg = std::get<0>(tmp77);
x86_ea o = std::get<1>(tmp77);
x86_CPU_bitOp(c,o,x86_CPU_getReg(c,reg,x86_CPU_osz(c)),x86_CPU_osz(c),cast<int64_t>(3ULL),true);
break;}
case cast<uint8_t>(186ULL):{
auto tmp78 = x86_CPU_modrmE(c);
uint8_t sub_ = std::get<0>(tmp78);
x86_ea o = std::get<1>(tmp78);
uint32_t imm = x86_CPU_fetch8(c);
x86_CPU_bitOp(c,o,imm,x86_CPU_osz(c),cast<int64_t>(cast<uint8_t>((sub_ & cast<uint8_t>(3ULL)))),false);
break;}
case cast<uint8_t>(188ULL):{
auto tmp79 = x86_CPU_modrmE(c);
uint8_t reg = std::get<0>(tmp79);
x86_ea o = std::get<1>(tmp79);
x86_CPU_bitScan(c,reg,x86_CPU_rEA(c,o,x86_CPU_osz(c)),x86_CPU_osz(c),false);
break;}
case cast<uint8_t>(189ULL):{
auto tmp80 = x86_CPU_modrmE(c);
uint8_t reg = std::get<0>(tmp80);
x86_ea o = std::get<1>(tmp80);
x86_CPU_bitScan(c,reg,x86_CPU_rEA(c,o,x86_CPU_osz(c)),x86_CPU_osz(c),true);
break;}
case cast<uint8_t>(175ULL):{
auto tmp81 = x86_CPU_modrmE(c);
uint8_t reg = std::get<0>(tmp81);
x86_ea o = std::get<1>(tmp81);
int64_t w = x86_CPU_osz(c);
x86_CPU_setReg(c,reg,w,x86_CPU_imulTrunc(c,x86_CPU_getReg(c,reg,w),x86_CPU_rEA(c,o,w),w));
break;}
case cast<uint8_t>(182ULL):{
auto tmp82 = x86_CPU_modrmE(c);
uint8_t reg = std::get<0>(tmp82);
x86_ea o = std::get<1>(tmp82);
x86_CPU_setReg(c,reg,x86_CPU_osz(c),cast<uint32_t>((x86_CPU_rEA(c,o,cast<int64_t>(1ULL)) & cast<uint32_t>(255ULL))));
break;}
case cast<uint8_t>(183ULL):{
auto tmp83 = x86_CPU_modrmE(c);
uint8_t reg = std::get<0>(tmp83);
x86_ea o = std::get<1>(tmp83);
x86_CPU_setReg(c,reg,x86_CPU_osz(c),cast<uint32_t>((x86_CPU_rEA(c,o,cast<int64_t>(2ULL)) & cast<uint32_t>(65535ULL))));
break;}
case cast<uint8_t>(190ULL):{
auto tmp84 = x86_CPU_modrmE(c);
uint8_t reg = std::get<0>(tmp84);
x86_ea o = std::get<1>(tmp84);
x86_CPU_setReg(c,reg,x86_CPU_osz(c),x86_signExtByte(x86_CPU_rEA(c,o,cast<int64_t>(1ULL))));
break;}
case cast<uint8_t>(191ULL):{
auto tmp85 = x86_CPU_modrmE(c);
uint8_t reg = std::get<0>(tmp85);
x86_ea o = std::get<1>(tmp85);
x86_CPU_setReg(c,reg,x86_CPU_osz(c),x86_signExtWord(x86_CPU_rEA(c,o,cast<int64_t>(2ULL))));
break;}
default:{
if ((x86_CPU_execSSE(c,op,rep) || x86_CPU_execMMXInt(c,op,rep))) {
return ;
}
x86_CPU_Halt(c,std::string("unimplemented 0F opcode $%02X at %s",35),op,x86_CPU_at(c));
break;}
}}
}
}
// tools/cpu/x86/exec2.go:673:1
void x86_CPU_bitOp(x86_CPU* c,x86_ea o,uint32_t idx,int64_t w,int64_t mode,bool memBitString){
{
if (o.isReg) {
uint32_t b = modi<uint32_t>(idx,cast<uint32_t>(cast<int64_t>((w * cast<int64_t>(8ULL)))));
uint32_t v = x86_CPU_getReg(c,o.reg,w);
c->CF = (cast<uint32_t>(((shr<uint32_t>(v,b)) & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL));
if ((mode != cast<int64_t>(0ULL))) {
x86_CPU_setReg(c,o.reg,w,x86_applyBit(v,b,mode));
}
return ;
}
auto tmp86 = std::make_tuple(o.base,o.off,idx);
uint32_t base = std::get<0>(tmp86);
uint32_t off = std::get<1>(tmp86);
uint32_t b = std::get<2>(tmp86);
int64_t bw = w;
if (memBitString) {
off += cast<uint32_t>(shr<int32_t>(cast<int32_t>(idx),cast<int64_t>(3ULL)));
b = cast<uint32_t>((idx & cast<uint32_t>(7ULL)));
bw = cast<int64_t>(1ULL);
}
else {
b = modi<uint32_t>(idx,cast<uint32_t>(cast<int64_t>((w * cast<int64_t>(8ULL)))));
}
uint32_t v = x86_CPU_memRead(c,base,off,bw);
c->CF = (cast<uint32_t>(((shr<uint32_t>(v,b)) & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL));
if ((mode != cast<int64_t>(0ULL))) {
x86_CPU_memWrite(c,base,off,bw,x86_applyBit(v,b,mode));
}
}
}
// tools/cpu/x86/exec2.go:699:1
uint32_t x86_applyBit(uint32_t v,uint32_t b,int64_t mode){
{
{
switch(mode){
case cast<int64_t>(1ULL):{
return cast<uint32_t>((v | (shl<uint32_t>(cast<uint32_t>(1ULL),b))));
break;}
case cast<int64_t>(2ULL):{
return (v & ~((shl<uint32_t>(cast<uint32_t>(1ULL),b))));
break;}
case cast<int64_t>(3ULL):{
return cast<uint32_t>((v ^ (shl<uint32_t>(cast<uint32_t>(1ULL),b))));
break;}
}}
return v;
}
}
// tools/cpu/x86/exec2.go:715:1
void x86_CPU_bitScan(x86_CPU* c,uint8_t reg,uint32_t src,int64_t w,bool reverse){
{
src &= x86_widthMask(w);
if ((src == cast<uint32_t>(0ULL))) {
c->ZF = true;
return ;
}
c->ZF = false;
uint32_t idx={};
if (reverse) {
idx = cast<uint32_t>((cast<uint32_t>(cast<int64_t>((w * cast<int64_t>(8ULL)))) - cast<uint32_t>(1ULL)));
{;for (;(cast<uint32_t>((src & (shl<uint32_t>(cast<uint32_t>(1ULL),idx)))) == cast<uint32_t>(0ULL));){
idx--;
}
}}
else {
{;for (;(cast<uint32_t>((src & (shl<uint32_t>(cast<uint32_t>(1ULL),idx)))) == cast<uint32_t>(0ULL));){
idx++;
}
}}
x86_CPU_setReg(c,reg,w,idx);
}
}
// tools/cpu/x86/exec2.go:742:1
void x86_CPU_doubleShift(x86_CPU* c,bool left,x86_ea o,uint8_t reg,uint32_t count,int64_t w){
{
count &= cast<uint32_t>(31ULL);
uint32_t dst = x86_CPU_rEA(c,o,w);
if ((count == cast<uint32_t>(0ULL))) {
return ;
}
uint64_t bits = cast<uint64_t>(cast<int64_t>((w * cast<int64_t>(8ULL))));
if ((count > cast<uint32_t>(bits))) {
count = cast<uint32_t>(bits);
}
uint64_t n = cast<uint64_t>(count);
uint32_t src = x86_CPU_getReg(c,reg,w);
uint32_t res={};
if (left) {
res = shl<uint32_t>(dst,n);
if ((n < bits)) {
res |= shr<uint32_t>(src,(cast<uint64_t>((bits - n))));
}
c->CF = (cast<uint32_t>(((shr<uint32_t>(dst,(cast<uint64_t>((bits - n))))) & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL));
}
else {
res = shr<uint32_t>(dst,n);
if ((n < bits)) {
res |= shl<uint32_t>(src,(cast<uint64_t>((bits - n))));
}
c->CF = (cast<uint32_t>(((shr<uint32_t>(dst,(cast<uint64_t>((n - cast<uint64_t>(1ULL)))))) & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL));
}
res &= x86_widthMask(w);
if ((count == cast<uint32_t>(1ULL))) {
c->OF = (cast<uint32_t>(((cast<uint32_t>((res ^ dst))) & x86_signMask(w))) != cast<uint32_t>(0ULL));
}
x86_CPU_setSZP(c,res,w);
x86_CPU_wEA(c,o,w,res);
}
}
// tools/cpu/x86/fpuexec.go:41:1
void x86_FPUState_finit(x86_FPUState* f){
{
f->Ctrl = cast<uint16_t>(895ULL);
f->Stat = cast<uint16_t>(0ULL);
f->Top = cast<int64_t>(0ULL);
{auto&& tmp87 = f->Tag;
for(int64_t tmp88=0;tmp88<len(tmp87);++tmp88){
auto i=tmp88;f->Tag[i] = cast<uint8_t>(3ULL);
f->St[i] = cast<double>(0.00000000000000000e+00);
}}
}
}
// tools/cpu/x86/fpuexec.go:51:1
uint8_t x86_classify(double v){
{
{
if ((v == cast<double>(0.00000000000000000e+00))){
return cast<uint8_t>(1ULL);
}
else if ((go_math_IsNaN(v) || go_math_IsInf(v,cast<int64_t>(0ULL)))){
return cast<uint8_t>(2ULL);
}
else {
return cast<uint8_t>(0ULL);
}
}
tmp89:;
}
}
// tools/cpu/x86/fpuexec.go:63:1
int64_t x86_FPUState_phys(x86_FPUState* f,int64_t i){
{
return cast<int64_t>(((cast<int64_t>((f->Top + i))) & cast<int64_t>(7ULL)));
}
}
// tools/cpu/x86/fpuexec.go:66:1
double x86_FPUState_st(x86_FPUState* f,int64_t i){
{
return f->St[x86_FPUState_phys(f,i)];
}
}
// tools/cpu/x86/fpuexec.go:67:1
void x86_FPUState_setst(x86_FPUState* f,int64_t i,double v){
{
int64_t p = x86_FPUState_phys(f,i);
f->St[p] = v;
f->Tag[p] = x86_classify(v);
}
}
// tools/cpu/x86/fpuexec.go:77:1
void x86_FPUState_push(x86_FPUState* f,double v){
{
f->Top = cast<int64_t>(((cast<int64_t>((f->Top - cast<int64_t>(1ULL)))) & cast<int64_t>(7ULL)));
f->St[f->Top] = v;
f->Tag[f->Top] = x86_classify(v);
}
}
// tools/cpu/x86/fpuexec.go:82:1
void x86_FPUState_pop(x86_FPUState* f){
{
f->Tag[f->Top] = cast<uint8_t>(3ULL);
f->Top = cast<int64_t>(((cast<int64_t>((f->Top + cast<int64_t>(1ULL)))) & cast<int64_t>(7ULL)));
}
}
// tools/cpu/x86/fpuexec.go:89:1
uint16_t x86_FPUState_statusWord(x86_FPUState* f){
{
return cast<uint16_t>((((f->Stat & ~(cast<uint16_t>(14336ULL)))) | shl<uint16_t>(cast<uint16_t>(cast<int64_t>((f->Top & cast<int64_t>(7ULL)))),cast<int64_t>(11ULL))));
}
}
// tools/cpu/x86/fpuexec.go:94:1
void x86_FPUState_setCC(x86_FPUState* f,bool c3,bool c2,bool c0){
{
f->Stat &= ~(cast<uint16_t>(17664ULL));
if (c0) {
f->Stat |= cast<uint16_t>(256ULL);
}
if (c2) {
f->Stat |= cast<uint16_t>(1024ULL);
}
if (c3) {
f->Stat |= cast<uint16_t>(16384ULL);
}
}
}
// tools/cpu/x86/fpuexec.go:109:1
double x86_FPUState_round(x86_FPUState* f,double v){
{
{
switch(cast<uint16_t>(((shr<uint16_t>(f->Ctrl,cast<int64_t>(10ULL))) & cast<uint16_t>(3ULL)))){
case cast<uint16_t>(0ULL):{
return go_math_RoundToEven(v);
break;}
case cast<uint16_t>(1ULL):{
return go_math_Floor(v);
break;}
case cast<uint16_t>(2ULL):{
return go_math_Ceil(v);
break;}
default:{
return go_math_Trunc(v);
break;}
}}
}
}
// tools/cpu/x86/fpuexec.go:124:1
double x86_CPU_fLoad32(x86_CPU* c,x86_ea o){
{
return cast<double>(go_math_Float32frombits(x86_CPU_memRead(c,o.base,o.off,cast<int64_t>(4ULL))));
}
}
// tools/cpu/x86/fpuexec.go:125:1
void x86_CPU_fStore32(x86_CPU* c,x86_ea o,double v){
{
x86_CPU_memWrite(c,o.base,o.off,cast<int64_t>(4ULL),go_math_Float32bits(cast<float>(v)));
}
}
// tools/cpu/x86/fpuexec.go:128:1
double x86_CPU_fLoad64(x86_CPU* c,x86_ea o){
{
uint64_t lo = cast<uint64_t>(x86_CPU_memRead(c,o.base,o.off,cast<int64_t>(4ULL)));
uint64_t hi = cast<uint64_t>(x86_CPU_memRead(c,o.base,cast<uint32_t>((o.off + cast<uint32_t>(4ULL))),cast<int64_t>(4ULL)));
return go_math_Float64frombits(cast<uint64_t>((lo | shl<uint64_t>(hi,cast<int64_t>(32ULL)))));
}
}
// tools/cpu/x86/fpuexec.go:133:1
void x86_CPU_fStore64(x86_CPU* c,x86_ea o,double v){
{
uint64_t b = go_math_Float64bits(v);
x86_CPU_memWrite(c,o.base,o.off,cast<int64_t>(4ULL),cast<uint32_t>(b));
x86_CPU_memWrite(c,o.base,cast<uint32_t>((o.off + cast<uint32_t>(4ULL))),cast<int64_t>(4ULL),cast<uint32_t>(shr<uint64_t>(b,cast<int64_t>(32ULL))));
}
}
// tools/cpu/x86/fpuexec.go:138:1
double x86_CPU_fLoad80(x86_CPU* c,x86_ea o){
{
uint64_t mant = cast<uint64_t>((cast<uint64_t>(x86_CPU_memRead(c,o.base,o.off,cast<int64_t>(4ULL))) | shl<uint64_t>(cast<uint64_t>(x86_CPU_memRead(c,o.base,cast<uint32_t>((o.off + cast<uint32_t>(4ULL))),cast<int64_t>(4ULL))),cast<int64_t>(32ULL))));
uint16_t se = cast<uint16_t>(x86_CPU_memRead(c,o.base,cast<uint32_t>((o.off + cast<uint32_t>(8ULL))),cast<int64_t>(2ULL)));
return x86_f80ToF64(mant,se);
}
}
// tools/cpu/x86/fpuexec.go:143:1
void x86_CPU_fStore80(x86_CPU* c,x86_ea o,double v){
{
auto tmp90 = x86_f64ToF80(v);
uint64_t mant = std::get<0>(tmp90);
uint16_t se = std::get<1>(tmp90);
x86_CPU_memWrite(c,o.base,o.off,cast<int64_t>(4ULL),cast<uint32_t>(mant));
x86_CPU_memWrite(c,o.base,cast<uint32_t>((o.off + cast<uint32_t>(4ULL))),cast<int64_t>(4ULL),cast<uint32_t>(shr<uint64_t>(mant,cast<int64_t>(32ULL))));
x86_CPU_memWrite(c,o.base,cast<uint32_t>((o.off + cast<uint32_t>(8ULL))),cast<int64_t>(2ULL),cast<uint32_t>(se));
}
}
// tools/cpu/x86/fpuexec.go:152:1
double x86_f80ToF64(uint64_t mant,uint16_t se){
{
uint16_t sign = shr<uint16_t>(se,cast<int64_t>(15ULL));
int64_t exp = cast<int64_t>(cast<uint16_t>((se & cast<uint16_t>(32767ULL))));
auto neg = [&](double x)->double{
if ((sign == cast<uint16_t>(1ULL))) {
return cast<double>(-x);
}
return x;
}
;
{
if (((exp == cast<int64_t>(0ULL)) && (mant == cast<uint64_t>(0ULL)))){
return neg(cast<double>(0.00000000000000000e+00));
}
else if ((exp == cast<int64_t>(32767ULL))){
if ((shl<uint64_t>(mant,cast<int64_t>(1ULL)) == cast<uint64_t>(0ULL))) {
return neg(go_math_Inf(cast<int64_t>(1ULL)));
}
return go_math_NaN();
}
}
tmp91:;
return neg(go_math_Ldexp(cast<double>(mant),cast<int64_t>((exp - cast<int64_t>(16446ULL)))));
}
}
// tools/cpu/x86/fpuexec.go:174:1
std::tuple<uint64_t,uint16_t> x86_f64ToF80(double v){
{
uint64_t bits = go_math_Float64bits(v);
uint16_t sign = cast<uint16_t>((cast<uint16_t>(shr<uint64_t>(bits,cast<int64_t>(63ULL))) & cast<uint16_t>(1ULL)));
{
if ((v == cast<double>(0.00000000000000000e+00))){
return {cast<uint64_t>(0ULL),shl<uint16_t>(sign,cast<int64_t>(15ULL))};
}
else if (go_math_IsInf(v,cast<int64_t>(0ULL))){
return {cast<uint64_t>(9223372036854775808ULL),cast<uint16_t>((shl<uint16_t>(sign,cast<int64_t>(15ULL)) | cast<uint16_t>(32767ULL)))};
}
else if (go_math_IsNaN(v)){
return {cast<uint64_t>(13835058055282163712ULL),cast<uint16_t>((shl<uint16_t>(sign,cast<int64_t>(15ULL)) | cast<uint16_t>(32767ULL)))};
}
}
tmp92:;
auto tmp93 = go_math_Frexp(go_math_Abs(v));
double m = std::get<0>(tmp93);
int64_t e = std::get<1>(tmp93);
uint64_t mant = cast<uint64_t>((m * go_math_Exp2(cast<double>(6.40000000000000000e+01))));
int64_t e80 = cast<int64_t>((e + cast<int64_t>(16382ULL)));
return {mant,cast<uint16_t>((shl<uint16_t>(sign,cast<int64_t>(15ULL)) | cast<uint16_t>(e80)))};
}
}
// tools/cpu/x86/fpuexec.go:198:1
void x86_CPU_fpuExec(x86_CPU* c,uint8_t op){
{
uint8_t mb = cast<uint8_t>(x86_CPU_fetch8(c));
auto tmp94 = std::make_tuple(shr<uint8_t>(mb,cast<int64_t>(6ULL)),cast<uint8_t>(((shr<uint8_t>(mb,cast<int64_t>(3ULL))) & cast<uint8_t>(7ULL))),cast<uint8_t>((mb & cast<uint8_t>(7ULL))));
uint8_t mod = std::get<0>(tmp94);
uint8_t reg = std::get<1>(tmp94);
uint8_t rm = std::get<2>(tmp94);
if ((mod != cast<uint8_t>(3ULL))) {
x86_ea o={};
if ((c->dAddrsize == cast<int64_t>(32ULL))) {
o = x86_CPU_ea32(c,mod,rm);
}
else {
o = x86_CPU_ea16(c,mod,rm);
}
x86_CPU_fpuMemExec(c,op,reg,o);
return ;
}
x86_CPU_fpuRegExec(c,op,mb,reg,rm);
}
}
// tools/cpu/x86/fpuexec.go:216:1
std::tuple<double,bool> x86_arithST(int64_t sub_,double a,double b){
{
{
switch(sub_){
case cast<int64_t>(0ULL):{
return {(a + b),false};
break;}
case cast<int64_t>(1ULL):{
return {(a * b),false};
break;}
case cast<int64_t>(2ULL):case cast<int64_t>(3ULL):{
return {cast<double>(0.00000000000000000e+00),true};
break;}
case cast<int64_t>(4ULL):{
return {(a - b),false};
break;}
case cast<int64_t>(5ULL):{
return {(b - a),false};
break;}
case cast<int64_t>(6ULL):{
return {(a / b),false};
break;}
default:{
return {(b / a),false};
break;}
}}
}
}
// tools/cpu/x86/fpuexec.go:236:1
void x86_FPUState_fcom(x86_FPUState* f,double a,double b){
{
{
if ((go_math_IsNaN(a) || go_math_IsNaN(b))){
x86_FPUState_setCC(f,true,true,true);
}
else if ((a > b)){
x86_FPUState_setCC(f,false,false,false);
}
else if ((a < b)){
x86_FPUState_setCC(f,false,false,true);
}
else {
x86_FPUState_setCC(f,true,false,false);
}
}
tmp95:;
}
}
// tools/cpu/x86/fpuexec.go:251:1
void x86_CPU_fcomi(x86_CPU* c,double a,double b){
{
auto tmp96 = std::make_tuple(false,false,false);
c->OF = std::get<0>(tmp96);
c->SF = std::get<1>(tmp96);
c->AF = std::get<2>(tmp96);
{
if ((go_math_IsNaN(a) || go_math_IsNaN(b))){
auto tmp98 = std::make_tuple(true,true,true);
c->ZF = std::get<0>(tmp98);
c->PF = std::get<1>(tmp98);
c->CF = std::get<2>(tmp98);
}
else if ((a > b)){
auto tmp99 = std::make_tuple(false,false,false);
c->ZF = std::get<0>(tmp99);
c->PF = std::get<1>(tmp99);
c->CF = std::get<2>(tmp99);
}
else if ((a < b)){
auto tmp100 = std::make_tuple(false,false,true);
c->ZF = std::get<0>(tmp100);
c->PF = std::get<1>(tmp100);
c->CF = std::get<2>(tmp100);
}
else {
auto tmp101 = std::make_tuple(true,false,false);
c->ZF = std::get<0>(tmp101);
c->PF = std::get<1>(tmp101);
c->CF = std::get<2>(tmp101);
}
}
tmp97:;
}
}
// tools/cpu/x86/fpuexec.go:266:1
void x86_CPU_fpuMemExec(x86_CPU* c,uint8_t op,uint8_t reg,x86_ea o){
{
x86_FPUState* f = (&c->FPU);
{
switch(op){
case cast<uint8_t>(216ULL):{
double v = x86_CPU_fLoad32(c,o);
{
auto tmp102 = x86_arithST(cast<int64_t>(reg),x86_FPUState_st(f,cast<int64_t>(0ULL)),v);
double res = std::get<0>(tmp102);
bool cmp = std::get<1>(tmp102);
if (cmp) {
x86_FPUState_fcom(f,x86_FPUState_st(f,cast<int64_t>(0ULL)),v);
if ((reg == cast<uint8_t>(3ULL))) {
x86_FPUState_pop(f);
}
}
else {
x86_FPUState_setst(f,cast<int64_t>(0ULL),res);
}
}
break;}
case cast<uint8_t>(220ULL):{
double v = x86_CPU_fLoad64(c,o);
{
auto tmp103 = x86_arithST(cast<int64_t>(reg),x86_FPUState_st(f,cast<int64_t>(0ULL)),v);
double res = std::get<0>(tmp103);
bool cmp = std::get<1>(tmp103);
if (cmp) {
x86_FPUState_fcom(f,x86_FPUState_st(f,cast<int64_t>(0ULL)),v);
if ((reg == cast<uint8_t>(3ULL))) {
x86_FPUState_pop(f);
}
}
else {
x86_FPUState_setst(f,cast<int64_t>(0ULL),res);
}
}
break;}
case cast<uint8_t>(218ULL):{
double v = cast<double>(cast<int32_t>(x86_CPU_memRead(c,o.base,o.off,cast<int64_t>(4ULL))));
{
auto tmp104 = x86_arithST(cast<int64_t>(reg),x86_FPUState_st(f,cast<int64_t>(0ULL)),v);
double res = std::get<0>(tmp104);
bool cmp = std::get<1>(tmp104);
if (cmp) {
x86_FPUState_fcom(f,x86_FPUState_st(f,cast<int64_t>(0ULL)),v);
if ((reg == cast<uint8_t>(3ULL))) {
x86_FPUState_pop(f);
}
}
else {
x86_FPUState_setst(f,cast<int64_t>(0ULL),res);
}
}
break;}
case cast<uint8_t>(222ULL):{
double v = cast<double>(cast<int16_t>(cast<uint16_t>(x86_CPU_memRead(c,o.base,o.off,cast<int64_t>(2ULL)))));
{
auto tmp105 = x86_arithST(cast<int64_t>(reg),x86_FPUState_st(f,cast<int64_t>(0ULL)),v);
double res = std::get<0>(tmp105);
bool cmp = std::get<1>(tmp105);
if (cmp) {
x86_FPUState_fcom(f,x86_FPUState_st(f,cast<int64_t>(0ULL)),v);
if ((reg == cast<uint8_t>(3ULL))) {
x86_FPUState_pop(f);
}
}
else {
x86_FPUState_setst(f,cast<int64_t>(0ULL),res);
}
}
break;}
case cast<uint8_t>(217ULL):{
{
switch(reg){
case cast<uint8_t>(0ULL):{
x86_FPUState_push(f,x86_CPU_fLoad32(c,o));
break;}
case cast<uint8_t>(2ULL):{
x86_CPU_fStore32(c,o,x86_FPUState_st(f,cast<int64_t>(0ULL)));
break;}
case cast<uint8_t>(3ULL):{
x86_CPU_fStore32(c,o,x86_FPUState_st(f,cast<int64_t>(0ULL)));
x86_FPUState_pop(f);
break;}
case cast<uint8_t>(4ULL):{
x86_CPU_fldenv(c,o);
break;}
case cast<uint8_t>(5ULL):{
f->Ctrl = cast<uint16_t>(x86_CPU_memRead(c,o.base,o.off,cast<int64_t>(2ULL)));
break;}
case cast<uint8_t>(6ULL):{
x86_CPU_fnstenv(c,o);
break;}
case cast<uint8_t>(7ULL):{
x86_CPU_memWrite(c,o.base,o.off,cast<int64_t>(2ULL),cast<uint32_t>(f->Ctrl));
break;}
}}
break;}
case cast<uint8_t>(219ULL):{
{
switch(reg){
case cast<uint8_t>(0ULL):{
x86_FPUState_push(f,cast<double>(cast<int32_t>(x86_CPU_memRead(c,o.base,o.off,cast<int64_t>(4ULL)))));
break;}
case cast<uint8_t>(1ULL):{
x86_CPU_fistStore(c,o,cast<int64_t>(4ULL),go_math_Trunc(x86_FPUState_st(f,cast<int64_t>(0ULL))));
x86_FPUState_pop(f);
break;}
case cast<uint8_t>(2ULL):{
x86_CPU_fistStore(c,o,cast<int64_t>(4ULL),x86_FPUState_round(f,x86_FPUState_st(f,cast<int64_t>(0ULL))));
break;}
case cast<uint8_t>(3ULL):{
x86_CPU_fistStore(c,o,cast<int64_t>(4ULL),x86_FPUState_round(f,x86_FPUState_st(f,cast<int64_t>(0ULL))));
x86_FPUState_pop(f);
break;}
case cast<uint8_t>(5ULL):{
x86_FPUState_push(f,x86_CPU_fLoad80(c,o));
break;}
case cast<uint8_t>(7ULL):{
x86_CPU_fStore80(c,o,x86_FPUState_st(f,cast<int64_t>(0ULL)));
x86_FPUState_pop(f);
break;}
}}
break;}
case cast<uint8_t>(221ULL):{
{
switch(reg){
case cast<uint8_t>(0ULL):{
x86_FPUState_push(f,x86_CPU_fLoad64(c,o));
break;}
case cast<uint8_t>(1ULL):{
x86_CPU_fistStore(c,o,cast<int64_t>(8ULL),go_math_Trunc(x86_FPUState_st(f,cast<int64_t>(0ULL))));
x86_FPUState_pop(f);
break;}
case cast<uint8_t>(2ULL):{
x86_CPU_fStore64(c,o,x86_FPUState_st(f,cast<int64_t>(0ULL)));
break;}
case cast<uint8_t>(3ULL):{
x86_CPU_fStore64(c,o,x86_FPUState_st(f,cast<int64_t>(0ULL)));
x86_FPUState_pop(f);
break;}
case cast<uint8_t>(4ULL):{
x86_CPU_frstor(c,o);
break;}
case cast<uint8_t>(6ULL):{
x86_CPU_fnsave(c,o);
break;}
case cast<uint8_t>(7ULL):{
x86_CPU_memWrite(c,o.base,o.off,cast<int64_t>(2ULL),cast<uint32_t>(x86_FPUState_statusWord(f)));
break;}
}}
break;}
case cast<uint8_t>(223ULL):{
{
switch(reg){
case cast<uint8_t>(0ULL):{
x86_FPUState_push(f,cast<double>(cast<int16_t>(cast<uint16_t>(x86_CPU_memRead(c,o.base,o.off,cast<int64_t>(2ULL))))));
break;}
case cast<uint8_t>(1ULL):{
x86_CPU_fistStore(c,o,cast<int64_t>(2ULL),go_math_Trunc(x86_FPUState_st(f,cast<int64_t>(0ULL))));
x86_FPUState_pop(f);
break;}
case cast<uint8_t>(2ULL):{
x86_CPU_fistStore(c,o,cast<int64_t>(2ULL),x86_FPUState_round(f,x86_FPUState_st(f,cast<int64_t>(0ULL))));
break;}
case cast<uint8_t>(3ULL):{
x86_CPU_fistStore(c,o,cast<int64_t>(2ULL),x86_FPUState_round(f,x86_FPUState_st(f,cast<int64_t>(0ULL))));
x86_FPUState_pop(f);
break;}
case cast<uint8_t>(4ULL):{
x86_FPUState_push(f,cast<double>(0.00000000000000000e+00));
break;}
case cast<uint8_t>(5ULL):{
uint64_t lo = cast<uint64_t>((cast<uint64_t>(x86_CPU_memRead(c,o.base,o.off,cast<int64_t>(4ULL))) | shl<uint64_t>(cast<uint64_t>(x86_CPU_memRead(c,o.base,cast<uint32_t>((o.off + cast<uint32_t>(4ULL))),cast<int64_t>(4ULL))),cast<int64_t>(32ULL))));
x86_FPUState_push(f,cast<double>(cast<int64_t>(lo)));
break;}
case cast<uint8_t>(6ULL):{
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < cast<uint32_t>(10ULL));i++){
x86_CPU_wr8(c,x86_CPU_linear(c,o.base,cast<uint32_t>((o.off + i))),cast<uint32_t>(0ULL));
}
}x86_FPUState_pop(f);
break;}
case cast<uint8_t>(7ULL):{
x86_CPU_fistStore(c,o,cast<int64_t>(8ULL),x86_FPUState_round(f,x86_FPUState_st(f,cast<int64_t>(0ULL))));
x86_FPUState_pop(f);
break;}
}}
break;}
}}
}
}
// tools/cpu/x86/fpuexec.go:398:1
void x86_CPU_fistStore(x86_CPU* c,x86_ea o,int64_t n,double fv){
{
int64_t iv = cast<int64_t>(fv);
{
switch(n){
case cast<int64_t>(2ULL):{
if (((fv < cast<double>(-3.27680000000000000e+04)) || (fv > cast<double>(3.27670000000000000e+04)))) {
iv = cast<int64_t>(-32768ULL);
}
x86_CPU_memWrite(c,o.base,o.off,cast<int64_t>(2ULL),cast<uint32_t>(cast<uint16_t>(cast<int16_t>(iv))));
break;}
case cast<int64_t>(4ULL):{
if (((fv < cast<double>(-2.14748364800000000e+09)) || (fv > cast<double>(2.14748364700000000e+09)))) {
iv = cast<int64_t>(-2147483648ULL);
}
x86_CPU_memWrite(c,o.base,o.off,cast<int64_t>(4ULL),cast<uint32_t>(cast<int32_t>(iv)));
break;}
default:{
x86_CPU_memWrite(c,o.base,o.off,cast<int64_t>(4ULL),cast<uint32_t>(iv));
x86_CPU_memWrite(c,o.base,cast<uint32_t>((o.off + cast<uint32_t>(4ULL))),cast<int64_t>(4ULL),cast<uint32_t>(shr<uint64_t>(cast<uint64_t>(iv),cast<int64_t>(32ULL))));
break;}
}}
}
}
// tools/cpu/x86/fpuexec.go:420:1
void x86_CPU_fpuRegExec(x86_CPU* c,uint8_t op,uint8_t mb,uint8_t reg,uint8_t rm){
{
x86_FPUState* f = (&c->FPU);
int64_t i = cast<int64_t>(rm);
{
switch(op){
case cast<uint8_t>(216ULL):{
double v = x86_FPUState_st(f,i);
{
auto tmp106 = x86_arithST(cast<int64_t>(reg),x86_FPUState_st(f,cast<int64_t>(0ULL)),v);
double res = std::get<0>(tmp106);
bool cmp = std::get<1>(tmp106);
if (cmp) {
x86_FPUState_fcom(f,x86_FPUState_st(f,cast<int64_t>(0ULL)),v);
if ((reg == cast<uint8_t>(3ULL))) {
x86_FPUState_pop(f);
}
}
else {
x86_FPUState_setst(f,cast<int64_t>(0ULL),res);
}
}
return ;
break;}
case cast<uint8_t>(220ULL):{
if (((reg == cast<uint8_t>(2ULL)) || (reg == cast<uint8_t>(3ULL)))) {
x86_FPUState_fcom(f,x86_FPUState_st(f,cast<int64_t>(0ULL)),x86_FPUState_st(f,i));
if ((reg == cast<uint8_t>(3ULL))) {
x86_FPUState_pop(f);
}
return ;
}
x86_FPUState_setst(f,i,x86_dstArith(cast<int64_t>(reg),x86_FPUState_st(f,i),x86_FPUState_st(f,cast<int64_t>(0ULL))));
return ;
break;}
case cast<uint8_t>(222ULL):{
if ((mb == cast<uint8_t>(217ULL))) {
x86_FPUState_fcom(f,x86_FPUState_st(f,cast<int64_t>(0ULL)),x86_FPUState_st(f,cast<int64_t>(1ULL)));
x86_FPUState_pop(f);
x86_FPUState_pop(f);
return ;
}
if (((reg == cast<uint8_t>(2ULL)) || (reg == cast<uint8_t>(3ULL)))) {
x86_FPUState_fcom(f,x86_FPUState_st(f,cast<int64_t>(0ULL)),x86_FPUState_st(f,i));
x86_FPUState_pop(f);
if ((reg == cast<uint8_t>(3ULL))) {
x86_FPUState_pop(f);
}
return ;
}
x86_FPUState_setst(f,i,x86_dstArith(cast<int64_t>(reg),x86_FPUState_st(f,i),x86_FPUState_st(f,cast<int64_t>(0ULL))));
x86_FPUState_pop(f);
return ;
break;}
}}
{
switch(op){
case cast<uint8_t>(217ULL):{
x86_CPU_fpuD9(c,mb,i);
break;}
case cast<uint8_t>(218ULL):{
x86_CPU_fpuDA(c,mb,i);
break;}
case cast<uint8_t>(219ULL):{
x86_CPU_fpuDB(c,mb,i);
break;}
case cast<uint8_t>(221ULL):{
x86_CPU_fpuDD(c,mb,i);
break;}
case cast<uint8_t>(223ULL):{
x86_CPU_fpuDF(c,mb,i);
break;}
}}
}
}
// tools/cpu/x86/fpuexec.go:482:1
double x86_dstArith(int64_t sub_,double dst,double src){
{
{
switch(sub_){
case cast<int64_t>(0ULL):{
return (dst + src);
break;}
case cast<int64_t>(1ULL):{
return (dst * src);
break;}
case cast<int64_t>(4ULL):{
return (src - dst);
break;}
case cast<int64_t>(5ULL):{
return (dst - src);
break;}
case cast<int64_t>(6ULL):{
return (src / dst);
break;}
default:{
return (dst / src);
break;}
}}
}
}
// tools/cpu/x86/fpuexec.go:501:1
void x86_CPU_fpuD9(x86_CPU* c,uint8_t mb,int64_t i){
{
x86_FPUState* f = (&c->FPU);
{
if (((mb >= cast<uint8_t>(192ULL)) && (mb <= cast<uint8_t>(199ULL)))){
x86_FPUState_push(f,x86_FPUState_st(f,i));
return ;
}
else if (((mb >= cast<uint8_t>(200ULL)) && (mb <= cast<uint8_t>(207ULL)))){
auto tmp108 = std::make_tuple(x86_FPUState_phys(f,cast<int64_t>(0ULL)),x86_FPUState_phys(f,i));
int64_t p0 = std::get<0>(tmp108);
int64_t pi = std::get<1>(tmp108);
auto tmp109 = std::make_tuple(f->St[pi],f->St[p0]);
f->St[p0] = std::get<0>(tmp109);
f->St[pi] = std::get<1>(tmp109);
auto tmp110 = std::make_tuple(f->Tag[pi],f->Tag[p0]);
f->Tag[p0] = std::get<0>(tmp110);
f->Tag[pi] = std::get<1>(tmp110);
return ;
}
}
tmp107:;
{
switch(mb){
case cast<uint8_t>(208ULL):{
break;}
case cast<uint8_t>(224ULL):{
x86_FPUState_setst(f,cast<int64_t>(0ULL),cast<double>(-x86_FPUState_st(f,cast<int64_t>(0ULL))));
break;}
case cast<uint8_t>(225ULL):{
x86_FPUState_setst(f,cast<int64_t>(0ULL),go_math_Abs(x86_FPUState_st(f,cast<int64_t>(0ULL))));
break;}
case cast<uint8_t>(228ULL):{
x86_FPUState_fcom(f,x86_FPUState_st(f,cast<int64_t>(0ULL)),cast<double>(0.00000000000000000e+00));
break;}
case cast<uint8_t>(229ULL):{
x86_CPU_fxam(c);
break;}
case cast<uint8_t>(232ULL):{
x86_FPUState_push(f,cast<double>(1.00000000000000000e+00));
break;}
case cast<uint8_t>(233ULL):{
x86_FPUState_push(f,go_math_Log2(cast<double>(1.00000000000000000e+01)));
break;}
case cast<uint8_t>(234ULL):{
x86_FPUState_push(f,go_math_Log2(cast<double>(2.71828182845904509e+00)));
break;}
case cast<uint8_t>(235ULL):{
x86_FPUState_push(f,cast<double>(3.14159265358979312e+00));
break;}
case cast<uint8_t>(236ULL):{
x86_FPUState_push(f,go_math_Log10(cast<double>(2.00000000000000000e+00)));
break;}
case cast<uint8_t>(237ULL):{
x86_FPUState_push(f,cast<double>(6.93147180559945286e-01));
break;}
case cast<uint8_t>(238ULL):{
x86_FPUState_push(f,cast<double>(0.00000000000000000e+00));
break;}
case cast<uint8_t>(240ULL):{
x86_FPUState_setst(f,cast<int64_t>(0ULL),(go_math_Exp2(x86_FPUState_st(f,cast<int64_t>(0ULL))) - cast<double>(1.00000000000000000e+00)));
break;}
case cast<uint8_t>(241ULL):{
x86_FPUState_setst(f,cast<int64_t>(1ULL),(x86_FPUState_st(f,cast<int64_t>(1ULL)) * go_math_Log2(x86_FPUState_st(f,cast<int64_t>(0ULL)))));
x86_FPUState_pop(f);
break;}
case cast<uint8_t>(242ULL):{
x86_FPUState_setst(f,cast<int64_t>(0ULL),go_math_Tan(x86_FPUState_st(f,cast<int64_t>(0ULL))));
x86_FPUState_push(f,cast<double>(1.00000000000000000e+00));
f->Stat &= ~(cast<uint16_t>(1024ULL));
break;}
case cast<uint8_t>(243ULL):{
x86_FPUState_setst(f,cast<int64_t>(1ULL),go_math_Atan2(x86_FPUState_st(f,cast<int64_t>(1ULL)),x86_FPUState_st(f,cast<int64_t>(0ULL))));
x86_FPUState_pop(f);
break;}
case cast<uint8_t>(244ULL):{
x86_CPU_fxtract(c);
break;}
case cast<uint8_t>(245ULL):{
x86_CPU_fprem(c,true);
break;}
case cast<uint8_t>(246ULL):{
f->Top = cast<int64_t>(((cast<int64_t>((f->Top - cast<int64_t>(1ULL)))) & cast<int64_t>(7ULL)));
break;}
case cast<uint8_t>(247ULL):{
f->Top = cast<int64_t>(((cast<int64_t>((f->Top + cast<int64_t>(1ULL)))) & cast<int64_t>(7ULL)));
break;}
case cast<uint8_t>(248ULL):{
x86_CPU_fprem(c,false);
break;}
case cast<uint8_t>(249ULL):{
x86_FPUState_setst(f,cast<int64_t>(1ULL),(x86_FPUState_st(f,cast<int64_t>(1ULL)) * go_math_Log2((x86_FPUState_st(f,cast<int64_t>(0ULL)) + cast<double>(1.00000000000000000e+00)))));
x86_FPUState_pop(f);
break;}
case cast<uint8_t>(250ULL):{
x86_FPUState_setst(f,cast<int64_t>(0ULL),go_math_Sqrt(x86_FPUState_st(f,cast<int64_t>(0ULL))));
break;}
case cast<uint8_t>(251ULL):{
double x = x86_FPUState_st(f,cast<int64_t>(0ULL));
x86_FPUState_setst(f,cast<int64_t>(0ULL),go_math_Sin(x));
x86_FPUState_push(f,go_math_Cos(x));
f->Stat &= ~(cast<uint16_t>(1024ULL));
break;}
case cast<uint8_t>(252ULL):{
x86_FPUState_setst(f,cast<int64_t>(0ULL),x86_FPUState_round(f,x86_FPUState_st(f,cast<int64_t>(0ULL))));
break;}
case cast<uint8_t>(253ULL):{
x86_FPUState_setst(f,cast<int64_t>(0ULL),go_math_Ldexp(x86_FPUState_st(f,cast<int64_t>(0ULL)),cast<int64_t>(go_math_Trunc(x86_FPUState_st(f,cast<int64_t>(1ULL))))));
break;}
case cast<uint8_t>(254ULL):{
x86_FPUState_setst(f,cast<int64_t>(0ULL),go_math_Sin(x86_FPUState_st(f,cast<int64_t>(0ULL))));
f->Stat &= ~(cast<uint16_t>(1024ULL));
break;}
case cast<uint8_t>(255ULL):{
x86_FPUState_setst(f,cast<int64_t>(0ULL),go_math_Cos(x86_FPUState_st(f,cast<int64_t>(0ULL))));
f->Stat &= ~(cast<uint16_t>(1024ULL));
break;}
}}
}
}
// tools/cpu/x86/fpuexec.go:583:1
void x86_CPU_fxam(x86_CPU* c){
{
x86_FPUState* f = (&c->FPU);
double v = x86_FPUState_st(f,cast<int64_t>(0ULL));
bool sign = go_math_Signbit(v);
bool c3={};
bool c2={};
bool c0={};
{
if ((f->Tag[x86_FPUState_phys(f,cast<int64_t>(0ULL))] == cast<uint8_t>(3ULL))){
auto tmp112 = std::make_tuple(true,false,true);
c3 = std::get<0>(tmp112);
c2 = std::get<1>(tmp112);
c0 = std::get<2>(tmp112);
}
else if (go_math_IsNaN(v)){
auto tmp113 = std::make_tuple(false,false,true);
c3 = std::get<0>(tmp113);
c2 = std::get<1>(tmp113);
c0 = std::get<2>(tmp113);
}
else if (go_math_IsInf(v,cast<int64_t>(0ULL))){
auto tmp114 = std::make_tuple(false,true,true);
c3 = std::get<0>(tmp114);
c2 = std::get<1>(tmp114);
c0 = std::get<2>(tmp114);
}
else if ((v == cast<double>(0.00000000000000000e+00))){
auto tmp115 = std::make_tuple(true,false,false);
c3 = std::get<0>(tmp115);
c2 = std::get<1>(tmp115);
c0 = std::get<2>(tmp115);
}
else {
auto tmp116 = std::make_tuple(false,true,false);
c3 = std::get<0>(tmp116);
c2 = std::get<1>(tmp116);
c0 = std::get<2>(tmp116);
}
}
tmp111:;
x86_FPUState_setCC(f,c3,c2,c0);
f->Stat &= ~(cast<uint16_t>(512ULL));
if (sign) {
f->Stat |= cast<uint16_t>(512ULL);
}
}
}
// tools/cpu/x86/fpuexec.go:608:1
void x86_CPU_fxtract(x86_CPU* c){
{
x86_FPUState* f = (&c->FPU);
double x = x86_FPUState_st(f,cast<int64_t>(0ULL));
if ((((x == cast<double>(0.00000000000000000e+00)) || go_math_IsInf(x,cast<int64_t>(0ULL))) || go_math_IsNaN(x))) {
x86_FPUState_push(f,x);
return ;
}
double exp = go_math_Floor(go_math_Log2(go_math_Abs(x)));
double sig = (x / go_math_Exp2(exp));
x86_FPUState_setst(f,cast<int64_t>(0ULL),exp);
x86_FPUState_push(f,sig);
}
}
// tools/cpu/x86/fpuexec.go:624:1
void x86_CPU_fprem(x86_CPU* c,bool ieee){
{
x86_FPUState* f = (&c->FPU);
auto tmp117 = std::make_tuple(x86_FPUState_st(f,cast<int64_t>(0ULL)),x86_FPUState_st(f,cast<int64_t>(1ULL)));
double a = std::get<0>(tmp117);
double b = std::get<1>(tmp117);
double r={};
double q={};
if (ieee) {
r = go_math_Remainder(a,b);
q = go_math_RoundToEven((a / b));
}
else {
r = go_math_Mod(a,b);
q = go_math_Trunc((a / b));
}
x86_FPUState_setst(f,cast<int64_t>(0ULL),r);
int64_t qi = cast<int64_t>(go_math_Abs(q));
f->Stat &= ~(cast<uint16_t>(18176ULL));
if ((cast<int64_t>((qi & cast<int64_t>(1ULL))) != cast<int64_t>(0ULL))) {
f->Stat |= cast<uint16_t>(512ULL);
}
if ((cast<int64_t>((qi & cast<int64_t>(2ULL))) != cast<int64_t>(0ULL))) {
f->Stat |= cast<uint16_t>(16384ULL);
}
if ((cast<int64_t>((qi & cast<int64_t>(4ULL))) != cast<int64_t>(0ULL))) {
f->Stat |= cast<uint16_t>(256ULL);
}
}
}
// tools/cpu/x86/fpuexec.go:650:1
void x86_CPU_fpuDA(x86_CPU* c,uint8_t mb,int64_t i){
{
x86_FPUState* f = (&c->FPU);
if ((mb == cast<uint8_t>(233ULL))) {
x86_FPUState_fcom(f,x86_FPUState_st(f,cast<int64_t>(0ULL)),x86_FPUState_st(f,cast<int64_t>(1ULL)));
x86_FPUState_pop(f);
x86_FPUState_pop(f);
return ;
}
bool move={};
{
switch(cast<uint8_t>((mb & cast<uint8_t>(248ULL)))){
case cast<uint8_t>(192ULL):{
move = c->CF;
break;}
case cast<uint8_t>(200ULL):{
move = c->ZF;
break;}
case cast<uint8_t>(208ULL):{
move = (c->CF || c->ZF);
break;}
case cast<uint8_t>(216ULL):{
move = c->PF;
break;}
}}
if (move) {
x86_FPUState_setst(f,cast<int64_t>(0ULL),x86_FPUState_st(f,i));
}
}
}
// tools/cpu/x86/fpuexec.go:675:1
void x86_CPU_fpuDB(x86_CPU* c,uint8_t mb,int64_t i){
{
x86_FPUState* f = (&c->FPU);
{
switch(mb){
case cast<uint8_t>(226ULL):{
f->Stat &= ~(cast<uint16_t>(33023ULL));
return ;
break;}
case cast<uint8_t>(227ULL):{
x86_FPUState_finit(f);
return ;
break;}
}}
{
switch(cast<uint8_t>((mb & cast<uint8_t>(248ULL)))){
case cast<uint8_t>(192ULL):case cast<uint8_t>(200ULL):case cast<uint8_t>(208ULL):case cast<uint8_t>(216ULL):{
bool move={};
{
switch(cast<uint8_t>((mb & cast<uint8_t>(248ULL)))){
case cast<uint8_t>(192ULL):{
move = (!c->CF);
break;}
case cast<uint8_t>(200ULL):{
move = (!c->ZF);
break;}
case cast<uint8_t>(208ULL):{
move = (!((c->CF || c->ZF)));
break;}
case cast<uint8_t>(216ULL):{
move = (!c->PF);
break;}
}}
if (move) {
x86_FPUState_setst(f,cast<int64_t>(0ULL),x86_FPUState_st(f,i));
}
break;}
case cast<uint8_t>(232ULL):{
x86_CPU_fcomi(c,x86_FPUState_st(f,cast<int64_t>(0ULL)),x86_FPUState_st(f,i));
break;}
case cast<uint8_t>(240ULL):{
x86_CPU_fcomi(c,x86_FPUState_st(f,cast<int64_t>(0ULL)),x86_FPUState_st(f,i));
break;}
}}
}
}
// tools/cpu/x86/fpuexec.go:709:1
void x86_CPU_fpuDD(x86_CPU* c,uint8_t mb,int64_t i){
{
x86_FPUState* f = (&c->FPU);
{
switch(cast<uint8_t>((mb & cast<uint8_t>(248ULL)))){
case cast<uint8_t>(192ULL):{
f->Tag[x86_FPUState_phys(f,i)] = cast<uint8_t>(3ULL);
break;}
case cast<uint8_t>(208ULL):{
x86_FPUState_setst(f,i,x86_FPUState_st(f,cast<int64_t>(0ULL)));
break;}
case cast<uint8_t>(216ULL):{
x86_FPUState_setst(f,i,x86_FPUState_st(f,cast<int64_t>(0ULL)));
x86_FPUState_pop(f);
break;}
case cast<uint8_t>(224ULL):{
x86_FPUState_fcom(f,x86_FPUState_st(f,cast<int64_t>(0ULL)),x86_FPUState_st(f,i));
break;}
case cast<uint8_t>(232ULL):{
x86_FPUState_fcom(f,x86_FPUState_st(f,cast<int64_t>(0ULL)),x86_FPUState_st(f,i));
x86_FPUState_pop(f);
break;}
}}
}
}
// tools/cpu/x86/fpuexec.go:728:1
void x86_CPU_fpuDF(x86_CPU* c,uint8_t mb,int64_t i){
{
x86_FPUState* f = (&c->FPU);
if ((mb == cast<uint8_t>(224ULL))) {
x86_CPU_s16(c,cast<uint8_t>(0ULL),cast<uint32_t>(x86_FPUState_statusWord(f)));
return ;
}
{
switch(cast<uint8_t>((mb & cast<uint8_t>(248ULL)))){
case cast<uint8_t>(192ULL):{
f->Tag[x86_FPUState_phys(f,i)] = cast<uint8_t>(3ULL);
x86_FPUState_pop(f);
break;}
case cast<uint8_t>(232ULL):{
x86_CPU_fcomi(c,x86_FPUState_st(f,cast<int64_t>(0ULL)),x86_FPUState_st(f,i));
x86_FPUState_pop(f);
break;}
case cast<uint8_t>(240ULL):{
x86_CPU_fcomi(c,x86_FPUState_st(f,cast<int64_t>(0ULL)),x86_FPUState_st(f,i));
x86_FPUState_pop(f);
break;}
}}
}
}
// tools/cpu/x86/fpuexec.go:755:1
bool x86_CPU_envIs32(x86_CPU* c){
{
return (c->dOpsize == cast<int64_t>(32ULL));
}
}
// tools/cpu/x86/fpuexec.go:757:1
void x86_CPU_fnstenv(x86_CPU* c,x86_ea o){
{
x86_FPUState* f = (&c->FPU);
x86_CPU_envWriteHeader(c,o,f);
}
}
// tools/cpu/x86/fpuexec.go:761:1
void x86_CPU_fldenv(x86_CPU* c,x86_ea o){
{
x86_FPUState* f = (&c->FPU);
x86_CPU_envReadHeader(c,o,f);
}
}
// tools/cpu/x86/fpuexec.go:765:1
void x86_CPU_fnsave(x86_CPU* c,x86_ea o){
{
x86_FPUState* f = (&c->FPU);
x86_CPU_envWriteHeader(c,o,f);
uint32_t base = cast<uint32_t>(14ULL);
if (x86_CPU_envIs32(c)) {
base = cast<uint32_t>(28ULL);
}
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(8ULL));i++){
auto tmp118 = x86_f64ToF80(f->St[i]);
uint64_t mant = std::get<0>(tmp118);
uint16_t se = std::get<1>(tmp118);
auto tmp119 = cast<uint32_t>((cast<uint32_t>((o.off + base)) + cast<uint32_t>((cast<uint32_t>(i) * cast<uint32_t>(10ULL)))));
uint32_t off = tmp119;
x86_CPU_memWrite(c,o.base,off,cast<int64_t>(4ULL),cast<uint32_t>(mant));
x86_CPU_memWrite(c,o.base,cast<uint32_t>((off + cast<uint32_t>(4ULL))),cast<int64_t>(4ULL),cast<uint32_t>(shr<uint64_t>(mant,cast<int64_t>(32ULL))));
x86_CPU_memWrite(c,o.base,cast<uint32_t>((off + cast<uint32_t>(8ULL))),cast<int64_t>(2ULL),cast<uint32_t>(se));
}
}x86_FPUState_finit(f);
}
}
// tools/cpu/x86/fpuexec.go:781:1
void x86_CPU_frstor(x86_CPU* c,x86_ea o){
{
x86_FPUState* f = (&c->FPU);
x86_CPU_envReadHeader(c,o,f);
uint32_t base = cast<uint32_t>(14ULL);
if (x86_CPU_envIs32(c)) {
base = cast<uint32_t>(28ULL);
}
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(8ULL));i++){
auto tmp120 = cast<uint32_t>((cast<uint32_t>((o.off + base)) + cast<uint32_t>((cast<uint32_t>(i) * cast<uint32_t>(10ULL)))));
uint32_t off = tmp120;
uint64_t mant = cast<uint64_t>((cast<uint64_t>(x86_CPU_memRead(c,o.base,off,cast<int64_t>(4ULL))) | shl<uint64_t>(cast<uint64_t>(x86_CPU_memRead(c,o.base,cast<uint32_t>((off + cast<uint32_t>(4ULL))),cast<int64_t>(4ULL))),cast<int64_t>(32ULL))));
uint16_t se = cast<uint16_t>(x86_CPU_memRead(c,o.base,cast<uint32_t>((off + cast<uint32_t>(8ULL))),cast<int64_t>(2ULL)));
f->St[i] = x86_f80ToF64(mant,se);
}
}}
}
// tools/cpu/x86/fpuexec.go:796:1
void x86_CPU_envWriteHeader(x86_CPU* c,x86_ea o,x86_FPUState* f){
{
if (x86_CPU_envIs32(c)) {
x86_CPU_memWrite(c,o.base,o.off,cast<int64_t>(4ULL),cast<uint32_t>(f->Ctrl));
x86_CPU_memWrite(c,o.base,cast<uint32_t>((o.off + cast<uint32_t>(4ULL))),cast<int64_t>(4ULL),cast<uint32_t>(x86_FPUState_statusWord(f)));
x86_CPU_memWrite(c,o.base,cast<uint32_t>((o.off + cast<uint32_t>(8ULL))),cast<int64_t>(4ULL),cast<uint32_t>(x86_FPUState_tagWord(f)));
{uint32_t i = cast<uint32_t>(12ULL);for (;(i < cast<uint32_t>(28ULL));i += cast<uint32_t>(4ULL)){
x86_CPU_memWrite(c,o.base,cast<uint32_t>((o.off + i)),cast<int64_t>(4ULL),cast<uint32_t>(0ULL));
}
}}
else {
x86_CPU_memWrite(c,o.base,o.off,cast<int64_t>(2ULL),cast<uint32_t>(f->Ctrl));
x86_CPU_memWrite(c,o.base,cast<uint32_t>((o.off + cast<uint32_t>(2ULL))),cast<int64_t>(2ULL),cast<uint32_t>(x86_FPUState_statusWord(f)));
x86_CPU_memWrite(c,o.base,cast<uint32_t>((o.off + cast<uint32_t>(4ULL))),cast<int64_t>(2ULL),cast<uint32_t>(x86_FPUState_tagWord(f)));
{uint32_t i = cast<uint32_t>(6ULL);for (;(i < cast<uint32_t>(14ULL));i += cast<uint32_t>(2ULL)){
x86_CPU_memWrite(c,o.base,cast<uint32_t>((o.off + i)),cast<int64_t>(2ULL),cast<uint32_t>(0ULL));
}
}}
}
}
// tools/cpu/x86/fpuexec.go:814:1
void x86_CPU_envReadHeader(x86_CPU* c,x86_ea o,x86_FPUState* f){
{
if (x86_CPU_envIs32(c)) {
f->Ctrl = cast<uint16_t>(x86_CPU_memRead(c,o.base,o.off,cast<int64_t>(4ULL)));
f->Stat = cast<uint16_t>(x86_CPU_memRead(c,o.base,cast<uint32_t>((o.off + cast<uint32_t>(4ULL))),cast<int64_t>(4ULL)));
x86_FPUState_setTagWord(f,cast<uint16_t>(x86_CPU_memRead(c,o.base,cast<uint32_t>((o.off + cast<uint32_t>(8ULL))),cast<int64_t>(4ULL))));
}
else {
f->Ctrl = cast<uint16_t>(x86_CPU_memRead(c,o.base,o.off,cast<int64_t>(2ULL)));
f->Stat = cast<uint16_t>(x86_CPU_memRead(c,o.base,cast<uint32_t>((o.off + cast<uint32_t>(2ULL))),cast<int64_t>(2ULL)));
x86_FPUState_setTagWord(f,cast<uint16_t>(x86_CPU_memRead(c,o.base,cast<uint32_t>((o.off + cast<uint32_t>(4ULL))),cast<int64_t>(2ULL))));
}
f->Top = cast<int64_t>((cast<int64_t>(shr<uint16_t>(f->Stat,cast<int64_t>(11ULL))) & cast<int64_t>(7ULL)));
f->Stat &= ~(cast<uint16_t>(14336ULL));
}
}
// tools/cpu/x86/fpuexec.go:829:1
uint16_t x86_FPUState_tagWord(x86_FPUState* f){
{
uint16_t w={};
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(8ULL));i++){
w |= shl<uint16_t>(cast<uint16_t>(cast<uint8_t>((f->Tag[i] & cast<uint8_t>(3ULL)))),(cast<uint64_t>((cast<uint64_t>(i) * cast<uint64_t>(2ULL)))));
}
}return w;
}
}
// tools/cpu/x86/fpuexec.go:836:1
void x86_FPUState_setTagWord(x86_FPUState* f,uint16_t w){
{
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(8ULL));i++){
f->Tag[i] = cast<uint8_t>((cast<uint8_t>(shr<uint16_t>(w,(cast<uint64_t>((cast<uint64_t>(i) * cast<uint64_t>(2ULL)))))) & cast<uint8_t>(3ULL)));
}
}}
}
// tools/cpu/x86/mmxint.go:14:1
std::tuple<uint8_t,std::array<uint8_t,16>,x86_ea,bool> x86_CPU_intOperands(x86_CPU* c,x86_sseKind k){
uint8_t reg{};
std::array<uint8_t,16> src{};
x86_ea o{};
bool wide{};
{
auto tmp121 = x86_CPU_modrmE(c);
reg = std::get<0>(tmp121);
o = std::get<1>(tmp121);
wide = (k == cast<x86_sseKind>(1ULL));
if (wide) {
src = x86_CPU_sseRM(c,o,cast<int64_t>(16ULL));
}
else {
std::array<uint8_t,8> s8 = x86_CPU_mmxRM(c,o);
gcopy(sub(src,0,cast<int64_t>(8ULL)),sub(s8,0,len(s8)));
}
return {reg,src,o,wide};
}
}
// tools/cpu/x86/mmxint.go:26:1
std::array<uint8_t,16> x86_CPU_intReg(x86_CPU* c,uint8_t reg,bool wide){
{
if (wide) {
return c->XMM[reg];
}
std::array<uint8_t,16> v={};
gcopy(sub(v,0,cast<int64_t>(8ULL)),sub(c->MMX[cast<uint8_t>((reg & cast<uint8_t>(7ULL)))],0,len(c->MMX[cast<uint8_t>((reg & cast<uint8_t>(7ULL)))])));
return v;
}
}
// tools/cpu/x86/mmxint.go:35:1
void x86_CPU_setIntReg(x86_CPU* c,uint8_t reg,bool wide,std::array<uint8_t,16> v){
{
if (wide) {
c->XMM[reg] = v;
return ;
}
std::array<uint8_t,8> m={};
gcopy(sub(m,0,len(m)),sub(v,0,cast<int64_t>(8ULL)));
c->MMX[cast<uint8_t>((reg & cast<uint8_t>(7ULL)))] = m;
}
}
// tools/cpu/x86/mmxint.go:46:1
int64_t x86_intWidth(bool wide){
{
if (wide) {
return cast<int64_t>(16ULL);
}
return cast<int64_t>(8ULL);
}
}
// tools/cpu/x86/mmxint.go:55:1
uint8_t x86_satI8(int32_t v){
{
if ((v < cast<int32_t>(-128ULL))) {
v = cast<int32_t>(-128ULL);
}
if ((v > cast<int32_t>(127ULL))) {
v = cast<int32_t>(127ULL);
}
return cast<uint8_t>(v);
}
}
// tools/cpu/x86/mmxint.go:65:1
uint8_t x86_satU8(int32_t v){
{
if ((v < cast<int32_t>(0ULL))) {
v = cast<int32_t>(0ULL);
}
if ((v > cast<int32_t>(255ULL))) {
v = cast<int32_t>(255ULL);
}
return cast<uint8_t>(v);
}
}
// tools/cpu/x86/mmxint.go:75:1
uint16_t x86_satI16(int32_t v){
{
if ((v < cast<int32_t>(-32768ULL))) {
v = cast<int32_t>(-32768ULL);
}
if ((v > cast<int32_t>(32767ULL))) {
v = cast<int32_t>(32767ULL);
}
return cast<uint16_t>(v);
}
}
// tools/cpu/x86/mmxint.go:85:1
uint16_t x86_satU16(int32_t v){
{
if ((v < cast<int32_t>(0ULL))) {
v = cast<int32_t>(0ULL);
}
if ((v > cast<int32_t>(65535ULL))) {
v = cast<int32_t>(65535ULL);
}
return cast<uint16_t>(v);
}
}
// tools/cpu/x86/mmxint.go:96:1
uint16_t x86_w16(Slice<uint8_t> b){
{
return cast<uint16_t>((cast<uint16_t>(b[cast<int64_t>(0ULL)]) | shl<uint16_t>(cast<uint16_t>(b[cast<int64_t>(1ULL)]),cast<int64_t>(8ULL))));
}
}
// tools/cpu/x86/mmxint.go:97:1
void x86_putw16(Slice<uint8_t> b,uint16_t v){
{
b[cast<int64_t>(0ULL)] = cast<uint8_t>(v);
b[cast<int64_t>(1ULL)] = cast<uint8_t>(shr<uint16_t>(v,cast<int64_t>(8ULL)));
}
}
// tools/cpu/x86/mmxint.go:101:1
bool x86_CPU_execMMXInt(x86_CPU* c,uint8_t op,uint8_t rep){
{
x86_sseKind k = x86_CPU_sseKindOf(c,rep);
{
switch(op){
case cast<uint8_t>(96ULL):case cast<uint8_t>(97ULL):case cast<uint8_t>(98ULL):case cast<uint8_t>(104ULL):case cast<uint8_t>(105ULL):case cast<uint8_t>(106ULL):case cast<uint8_t>(108ULL):case cast<uint8_t>(109ULL):{
if (((((op == cast<uint8_t>(108ULL)) || (op == cast<uint8_t>(109ULL)))) && (k != cast<x86_sseKind>(1ULL)))) {
return false;
}
auto tmp122 = x86_CPU_intOperands(c,k);
uint8_t reg = std::get<0>(tmp122);
std::array<uint8_t,16> src = std::get<1>(tmp122);
bool wide = std::get<3>(tmp122);
std::array<uint8_t,16> dst = x86_CPU_intReg(c,reg,wide);
int64_t n = x86_intWidth(wide);
int64_t elem={};
{
switch(cast<uint8_t>((op & cast<uint8_t>(15ULL)))){
case cast<uint8_t>(0ULL):case cast<uint8_t>(8ULL):{
elem = cast<int64_t>(1ULL);
break;}
case cast<uint8_t>(1ULL):case cast<uint8_t>(9ULL):{
elem = cast<int64_t>(2ULL);
break;}
case cast<uint8_t>(2ULL):case cast<uint8_t>(10ULL):{
elem = cast<int64_t>(4ULL);
break;}
default:{
elem = cast<int64_t>(8ULL);
break;}
}}
bool high = ((((op == cast<uint8_t>(104ULL)) || (op == cast<uint8_t>(105ULL))) || (op == cast<uint8_t>(106ULL))) || (op == cast<uint8_t>(109ULL)));
int64_t half = divi<int64_t>(n,cast<int64_t>(2ULL));
int64_t off = cast<int64_t>(0ULL);
if (high) {
off = half;
}
std::array<uint8_t,16> r={};
int64_t pairs = divi<int64_t>(half,elem);
{int64_t i = cast<int64_t>(0ULL);for (;(i < pairs);i++){
gcopy(sub(r,cast<int64_t>((cast<int64_t>((i * cast<int64_t>(2ULL))) * elem)),len(r)),sub(dst,cast<int64_t>((off + cast<int64_t>((i * elem)))),cast<int64_t>((cast<int64_t>((off + cast<int64_t>((i * elem)))) + elem))));
gcopy(sub(r,cast<int64_t>((cast<int64_t>((cast<int64_t>((i * cast<int64_t>(2ULL))) * elem)) + elem)),len(r)),sub(src,cast<int64_t>((off + cast<int64_t>((i * elem)))),cast<int64_t>((cast<int64_t>((off + cast<int64_t>((i * elem)))) + elem))));
}
}x86_CPU_setIntReg(c,reg,wide,r);
return true;
break;}
case cast<uint8_t>(99ULL):{
auto tmp123 = x86_CPU_intOperands(c,k);
uint8_t reg = std::get<0>(tmp123);
std::array<uint8_t,16> src = std::get<1>(tmp123);
bool wide = std::get<3>(tmp123);
std::array<uint8_t,16> dst = x86_CPU_intReg(c,reg,wide);
int64_t n = x86_intWidth(wide);
std::array<uint8_t,16> r={};
{int64_t i = cast<int64_t>(0ULL);for (;(i < divi<int64_t>(n,cast<int64_t>(2ULL)));i++){
r[i] = x86_satI8(cast<int32_t>(cast<int16_t>(x86_w16(sub(dst,cast<int64_t>((i * cast<int64_t>(2ULL))),len(dst))))));
}
}{int64_t i = cast<int64_t>(0ULL);for (;(i < divi<int64_t>(n,cast<int64_t>(2ULL)));i++){
r[cast<int64_t>((divi<int64_t>(n,cast<int64_t>(2ULL)) + i))] = x86_satI8(cast<int32_t>(cast<int16_t>(x86_w16(sub(src,cast<int64_t>((i * cast<int64_t>(2ULL))),len(src))))));
}
}x86_CPU_setIntReg(c,reg,wide,r);
return true;
break;}
case cast<uint8_t>(103ULL):{
auto tmp124 = x86_CPU_intOperands(c,k);
uint8_t reg = std::get<0>(tmp124);
std::array<uint8_t,16> src = std::get<1>(tmp124);
bool wide = std::get<3>(tmp124);
std::array<uint8_t,16> dst = x86_CPU_intReg(c,reg,wide);
int64_t n = x86_intWidth(wide);
std::array<uint8_t,16> r={};
{int64_t i = cast<int64_t>(0ULL);for (;(i < divi<int64_t>(n,cast<int64_t>(2ULL)));i++){
r[i] = x86_satU8(cast<int32_t>(cast<int16_t>(x86_w16(sub(dst,cast<int64_t>((i * cast<int64_t>(2ULL))),len(dst))))));
}
}{int64_t i = cast<int64_t>(0ULL);for (;(i < divi<int64_t>(n,cast<int64_t>(2ULL)));i++){
r[cast<int64_t>((divi<int64_t>(n,cast<int64_t>(2ULL)) + i))] = x86_satU8(cast<int32_t>(cast<int16_t>(x86_w16(sub(src,cast<int64_t>((i * cast<int64_t>(2ULL))),len(src))))));
}
}x86_CPU_setIntReg(c,reg,wide,r);
return true;
break;}
case cast<uint8_t>(107ULL):{
auto tmp125 = x86_CPU_intOperands(c,k);
uint8_t reg = std::get<0>(tmp125);
std::array<uint8_t,16> src = std::get<1>(tmp125);
bool wide = std::get<3>(tmp125);
std::array<uint8_t,16> dst = x86_CPU_intReg(c,reg,wide);
int64_t n = x86_intWidth(wide);
std::array<uint8_t,16> r={};
{int64_t i = cast<int64_t>(0ULL);for (;(i < divi<int64_t>(n,cast<int64_t>(4ULL)));i++){
int32_t v = cast<int32_t>(x86_le32b(sub(dst,cast<int64_t>((i * cast<int64_t>(4ULL))),len(dst))));
x86_putw16(sub(r,cast<int64_t>((i * cast<int64_t>(2ULL))),len(r)),x86_satI16Clamp(v));
}
}{int64_t i = cast<int64_t>(0ULL);for (;(i < divi<int64_t>(n,cast<int64_t>(4ULL)));i++){
int32_t v = cast<int32_t>(x86_le32b(sub(src,cast<int64_t>((i * cast<int64_t>(4ULL))),len(src))));
x86_putw16(sub(r,cast<int64_t>((divi<int64_t>(n,cast<int64_t>(2ULL)) + cast<int64_t>((i * cast<int64_t>(2ULL))))),len(r)),x86_satI16Clamp(v));
}
}x86_CPU_setIntReg(c,reg,wide,r);
return true;
break;}
case cast<uint8_t>(100ULL):case cast<uint8_t>(101ULL):case cast<uint8_t>(102ULL):case cast<uint8_t>(116ULL):case cast<uint8_t>(117ULL):case cast<uint8_t>(118ULL):{
auto tmp126 = x86_CPU_intOperands(c,k);
uint8_t reg = std::get<0>(tmp126);
std::array<uint8_t,16> src = std::get<1>(tmp126);
bool wide = std::get<3>(tmp126);
std::array<uint8_t,16> dst = x86_CPU_intReg(c,reg,wide);
int64_t n = x86_intWidth(wide);
bool eq = (op >= cast<uint8_t>(116ULL));
std::array<uint8_t,16> r={};
{
switch(cast<uint8_t>((op & cast<uint8_t>(3ULL)))){
case cast<uint8_t>(0ULL):{
{int64_t i = cast<int64_t>(0ULL);for (;(i < n);i++){
bool hit = (((eq && (dst[i] == src[i]))) || (((!eq) && (cast<int8_t>(dst[i]) > cast<int8_t>(src[i])))));
if (hit) {
r[i] = cast<uint8_t>(255ULL);
}
}
}break;}
case cast<uint8_t>(1ULL):{
{int64_t i = cast<int64_t>(0ULL);for (;(i < divi<int64_t>(n,cast<int64_t>(2ULL)));i++){
auto tmp127 = std::make_tuple(cast<int16_t>(x86_w16(sub(dst,cast<int64_t>((i * cast<int64_t>(2ULL))),len(dst)))),cast<int16_t>(x86_w16(sub(src,cast<int64_t>((i * cast<int64_t>(2ULL))),len(src)))));
int16_t a = std::get<0>(tmp127);
int16_t b = std::get<1>(tmp127);
if ((((eq && (a == b))) || (((!eq) && (a > b))))) {
x86_putw16(sub(r,cast<int64_t>((i * cast<int64_t>(2ULL))),len(r)),cast<uint16_t>(65535ULL));
}
}
}break;}
case cast<uint8_t>(2ULL):{
{int64_t i = cast<int64_t>(0ULL);for (;(i < divi<int64_t>(n,cast<int64_t>(4ULL)));i++){
auto tmp128 = std::make_tuple(cast<int32_t>(x86_le32b(sub(dst,cast<int64_t>((i * cast<int64_t>(4ULL))),len(dst)))),cast<int32_t>(x86_le32b(sub(src,cast<int64_t>((i * cast<int64_t>(4ULL))),len(src)))));
int32_t a = std::get<0>(tmp128);
int32_t b = std::get<1>(tmp128);
if ((((eq && (a == b))) || (((!eq) && (a > b))))) {
x86_putle32(sub(r,cast<int64_t>((i * cast<int64_t>(4ULL))),len(r)),cast<uint32_t>(4294967295ULL));
}
}
}break;}
}}
x86_CPU_setIntReg(c,reg,wide,r);
return true;
break;}
case cast<uint8_t>(112ULL):{
auto tmp129 = x86_CPU_modrmE(c);
uint8_t reg = std::get<0>(tmp129);
x86_ea o = std::get<1>(tmp129);
uint8_t imm = cast<uint8_t>(x86_CPU_fetch8(c));
{
switch(k){
case cast<x86_sseKind>(0ULL):{
std::array<uint8_t,8> src = x86_CPU_mmxRM(c,o);
std::array<uint8_t,8> r={};
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(4ULL));i++){
uint8_t s = cast<uint8_t>(((shr<uint8_t>(imm,(cast<int64_t>((i * cast<int64_t>(2ULL)))))) & cast<uint8_t>(3ULL)));
gcopy(sub(r,cast<int64_t>((i * cast<int64_t>(2ULL))),len(r)),sub(src,cast<uint8_t>((s * cast<uint8_t>(2ULL))),cast<uint8_t>((cast<uint8_t>((s * cast<uint8_t>(2ULL))) + cast<uint8_t>(2ULL)))));
}
}c->MMX[cast<uint8_t>((reg & cast<uint8_t>(7ULL)))] = r;
break;}
case cast<x86_sseKind>(1ULL):{
std::array<uint8_t,16> src = x86_CPU_sseRM(c,o,cast<int64_t>(16ULL));
std::array<uint8_t,16> r={};
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(4ULL));i++){
uint8_t s = cast<uint8_t>(((shr<uint8_t>(imm,(cast<int64_t>((i * cast<int64_t>(2ULL)))))) & cast<uint8_t>(3ULL)));
gcopy(sub(r,cast<int64_t>((i * cast<int64_t>(4ULL))),len(r)),sub(src,cast<uint8_t>((s * cast<uint8_t>(4ULL))),cast<uint8_t>((cast<uint8_t>((s * cast<uint8_t>(4ULL))) + cast<uint8_t>(4ULL)))));
}
}c->XMM[reg] = r;
break;}
case cast<x86_sseKind>(2ULL):{
std::array<uint8_t,16> src = x86_CPU_sseRM(c,o,cast<int64_t>(16ULL));
std::array<uint8_t,16> r = src;
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(4ULL));i++){
uint8_t s = cast<uint8_t>(((shr<uint8_t>(imm,(cast<int64_t>((i * cast<int64_t>(2ULL)))))) & cast<uint8_t>(3ULL)));
gcopy(sub(r,cast<int64_t>((cast<int64_t>(8ULL) + cast<int64_t>((i * cast<int64_t>(2ULL))))),len(r)),sub(src,cast<uint8_t>((cast<uint8_t>(8ULL) + cast<uint8_t>((s * cast<uint8_t>(2ULL))))),cast<uint8_t>((cast<uint8_t>((cast<uint8_t>(8ULL) + cast<uint8_t>((s * cast<uint8_t>(2ULL))))) + cast<uint8_t>(2ULL)))));
}
}c->XMM[reg] = r;
break;}
case cast<x86_sseKind>(3ULL):{
std::array<uint8_t,16> src = x86_CPU_sseRM(c,o,cast<int64_t>(16ULL));
std::array<uint8_t,16> r = src;
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(4ULL));i++){
uint8_t s = cast<uint8_t>(((shr<uint8_t>(imm,(cast<int64_t>((i * cast<int64_t>(2ULL)))))) & cast<uint8_t>(3ULL)));
gcopy(sub(r,cast<int64_t>((i * cast<int64_t>(2ULL))),len(r)),sub(src,cast<uint8_t>((s * cast<uint8_t>(2ULL))),cast<uint8_t>((cast<uint8_t>((s * cast<uint8_t>(2ULL))) + cast<uint8_t>(2ULL)))));
}
}c->XMM[reg] = r;
break;}
}}
return true;
break;}
case cast<uint8_t>(113ULL):case cast<uint8_t>(114ULL):case cast<uint8_t>(115ULL):{
auto tmp130 = x86_CPU_modrmE(c);
uint8_t sub_ = std::get<0>(tmp130);
x86_ea o = std::get<1>(tmp130);
uint32_t imm = cast<uint32_t>(x86_CPU_fetch8(c));
if ((!o.isReg)) {
return false;
}
bool wide = (k == cast<x86_sseKind>(1ULL));
uint8_t tgt = o.reg;
std::array<uint8_t,16> v = x86_CPU_intReg(c,tgt,wide);
int64_t n = x86_intWidth(wide);
int64_t elem={};
{
switch(op){
case cast<uint8_t>(113ULL):{
elem = cast<int64_t>(2ULL);
break;}
case cast<uint8_t>(114ULL):{
elem = cast<int64_t>(4ULL);
break;}
default:{
elem = cast<int64_t>(8ULL);
break;}
}}
{
switch(sub_){
case cast<uint8_t>(2ULL):{
x86_shiftLanes((&v),n,elem,imm,x86_shiftRightLogical);
break;}
case cast<uint8_t>(3ULL):{
if (((op != cast<uint8_t>(115ULL)) || (!wide))) {
return false;
}
x86_byteShift((&v),cast<int64_t>(imm),false);
break;}
case cast<uint8_t>(4ULL):{
if ((op == cast<uint8_t>(115ULL))) {
return false;
}
x86_shiftLanes((&v),n,elem,imm,x86_shiftRightArith);
break;}
case cast<uint8_t>(6ULL):{
x86_shiftLanes((&v),n,elem,imm,x86_shiftLeft);
break;}
case cast<uint8_t>(7ULL):{
if (((op != cast<uint8_t>(115ULL)) || (!wide))) {
return false;
}
x86_byteShift((&v),cast<int64_t>(imm),true);
break;}
default:{
return false;
break;}
}}
x86_CPU_setIntReg(c,tgt,wide,v);
return true;
break;}
case cast<uint8_t>(209ULL):case cast<uint8_t>(210ULL):case cast<uint8_t>(211ULL):case cast<uint8_t>(225ULL):case cast<uint8_t>(226ULL):case cast<uint8_t>(241ULL):case cast<uint8_t>(242ULL):case cast<uint8_t>(243ULL):{
auto tmp131 = x86_CPU_intOperands(c,k);
uint8_t reg = std::get<0>(tmp131);
std::array<uint8_t,16> src = std::get<1>(tmp131);
bool wide = std::get<3>(tmp131);
std::array<uint8_t,16> v = x86_CPU_intReg(c,reg,wide);
int64_t n = x86_intWidth(wide);
uint32_t cnt = x86_le32b(sub(src,cast<int64_t>(0ULL),cast<int64_t>(4ULL)));
if ((x86_le32b(sub(src,cast<int64_t>(4ULL),cast<int64_t>(8ULL))) != cast<uint32_t>(0ULL))) {
cnt = cast<uint32_t>(4294967295ULL);
}
int64_t elem={};
{
switch(op){
case cast<uint8_t>(209ULL):case cast<uint8_t>(225ULL):case cast<uint8_t>(241ULL):{
elem = cast<int64_t>(2ULL);
break;}
case cast<uint8_t>(210ULL):case cast<uint8_t>(226ULL):case cast<uint8_t>(242ULL):{
elem = cast<int64_t>(4ULL);
break;}
default:{
elem = cast<int64_t>(8ULL);
break;}
}}
{
if ((op >= cast<uint8_t>(241ULL))){
x86_shiftLanes((&v),n,elem,cnt,x86_shiftLeft);
}
else if ((op >= cast<uint8_t>(225ULL))){
x86_shiftLanes((&v),n,elem,cnt,x86_shiftRightArith);
}
else {
x86_shiftLanes((&v),n,elem,cnt,x86_shiftRightLogical);
}
}
tmp132:;
x86_CPU_setIntReg(c,reg,wide,v);
return true;
break;}
case cast<uint8_t>(219ULL):case cast<uint8_t>(223ULL):case cast<uint8_t>(235ULL):case cast<uint8_t>(239ULL):{
if (((op == cast<uint8_t>(239ULL)) && (k == cast<x86_sseKind>(1ULL)))) {
return false;
}
auto tmp133 = x86_CPU_intOperands(c,k);
uint8_t reg = std::get<0>(tmp133);
std::array<uint8_t,16> src = std::get<1>(tmp133);
bool wide = std::get<3>(tmp133);
std::array<uint8_t,16> dst = x86_CPU_intReg(c,reg,wide);
int64_t n = x86_intWidth(wide);
std::array<uint8_t,16> r={};
{int64_t i = cast<int64_t>(0ULL);for (;(i < n);i++){
{
switch(op){
case cast<uint8_t>(219ULL):{
r[i] = cast<uint8_t>((dst[i] & src[i]));
break;}
case cast<uint8_t>(223ULL):{
r[i] = cast<uint8_t>((cast<uint8_t>(~dst[i]) & src[i]));
break;}
case cast<uint8_t>(235ULL):{
r[i] = cast<uint8_t>((dst[i] | src[i]));
break;}
case cast<uint8_t>(239ULL):{
r[i] = cast<uint8_t>((dst[i] ^ src[i]));
break;}
}}
}
}x86_CPU_setIntReg(c,reg,wide,r);
return true;
break;}
case cast<uint8_t>(252ULL):case cast<uint8_t>(253ULL):case cast<uint8_t>(254ULL):case cast<uint8_t>(212ULL):case cast<uint8_t>(248ULL):case cast<uint8_t>(249ULL):case cast<uint8_t>(250ULL):case cast<uint8_t>(251ULL):{
auto tmp134 = x86_CPU_intOperands(c,k);
uint8_t reg = std::get<0>(tmp134);
std::array<uint8_t,16> src = std::get<1>(tmp134);
bool wide = std::get<3>(tmp134);
std::array<uint8_t,16> dst = x86_CPU_intReg(c,reg,wide);
int64_t n = x86_intWidth(wide);
bool sub_ = ((op >= cast<uint8_t>(248ULL)) && (op <= cast<uint8_t>(251ULL)));
int64_t elem={};
{
switch(op){
case cast<uint8_t>(252ULL):case cast<uint8_t>(248ULL):{
elem = cast<int64_t>(1ULL);
break;}
case cast<uint8_t>(253ULL):case cast<uint8_t>(249ULL):{
elem = cast<int64_t>(2ULL);
break;}
case cast<uint8_t>(254ULL):case cast<uint8_t>(250ULL):{
elem = cast<int64_t>(4ULL);
break;}
default:{
elem = cast<int64_t>(8ULL);
break;}
}}
std::array<uint8_t,16> r={};
{int64_t i = cast<int64_t>(0ULL);for (;(i < n);i += elem){
uint64_t a={};
uint64_t b={};
{int64_t j = cast<int64_t>(0ULL);for (;(j < elem);j++){
a |= shl<uint64_t>(cast<uint64_t>(dst[cast<int64_t>((i + j))]),(cast<int64_t>((cast<int64_t>(8ULL) * j))));
b |= shl<uint64_t>(cast<uint64_t>(src[cast<int64_t>((i + j))]),(cast<int64_t>((cast<int64_t>(8ULL) * j))));
}
}uint64_t v={};
if (sub_) {
v = cast<uint64_t>((a - b));
}
else {
v = cast<uint64_t>((a + b));
}
{int64_t j = cast<int64_t>(0ULL);for (;(j < elem);j++){
r[cast<int64_t>((i + j))] = cast<uint8_t>(shr<uint64_t>(v,(cast<int64_t>((cast<int64_t>(8ULL) * j)))));
}
}}
}x86_CPU_setIntReg(c,reg,wide,r);
return true;
break;}
case cast<uint8_t>(236ULL):case cast<uint8_t>(237ULL):case cast<uint8_t>(220ULL):case cast<uint8_t>(221ULL):case cast<uint8_t>(232ULL):case cast<uint8_t>(233ULL):case cast<uint8_t>(216ULL):case cast<uint8_t>(217ULL):{
auto tmp135 = x86_CPU_intOperands(c,k);
uint8_t reg = std::get<0>(tmp135);
std::array<uint8_t,16> src = std::get<1>(tmp135);
bool wide = std::get<3>(tmp135);
std::array<uint8_t,16> dst = x86_CPU_intReg(c,reg,wide);
int64_t n = x86_intWidth(wide);
std::array<uint8_t,16> r={};
bool signed_ = ((((op == cast<uint8_t>(236ULL)) || (op == cast<uint8_t>(237ULL))) || (op == cast<uint8_t>(232ULL))) || (op == cast<uint8_t>(233ULL)));
bool sub_ = ((((op == cast<uint8_t>(232ULL)) || (op == cast<uint8_t>(233ULL))) || (op == cast<uint8_t>(216ULL))) || (op == cast<uint8_t>(217ULL)));
bool words = ((((op == cast<uint8_t>(237ULL)) || (op == cast<uint8_t>(221ULL))) || (op == cast<uint8_t>(233ULL))) || (op == cast<uint8_t>(217ULL)));
if (words) {
{int64_t i = cast<int64_t>(0ULL);for (;(i < divi<int64_t>(n,cast<int64_t>(2ULL)));i++){
int32_t a={};
int32_t b={};
if (signed_) {
auto tmp136 = std::make_tuple(cast<int32_t>(cast<int16_t>(x86_w16(sub(dst,cast<int64_t>((i * cast<int64_t>(2ULL))),len(dst))))),cast<int32_t>(cast<int16_t>(x86_w16(sub(src,cast<int64_t>((i * cast<int64_t>(2ULL))),len(src))))));
a = std::get<0>(tmp136);
b = std::get<1>(tmp136);
}
else {
auto tmp137 = std::make_tuple(cast<int32_t>(x86_w16(sub(dst,cast<int64_t>((i * cast<int64_t>(2ULL))),len(dst)))),cast<int32_t>(x86_w16(sub(src,cast<int64_t>((i * cast<int64_t>(2ULL))),len(src)))));
a = std::get<0>(tmp137);
b = std::get<1>(tmp137);
}
int32_t v = cast<int32_t>((a + b));
if (sub_) {
v = cast<int32_t>((a - b));
}
if (signed_) {
x86_putw16(sub(r,cast<int64_t>((i * cast<int64_t>(2ULL))),len(r)),x86_satI16(v));
}
else {
x86_putw16(sub(r,cast<int64_t>((i * cast<int64_t>(2ULL))),len(r)),x86_satU16(v));
}
}
}}
else {
{int64_t i = cast<int64_t>(0ULL);for (;(i < n);i++){
int32_t a={};
int32_t b={};
if (signed_) {
auto tmp138 = std::make_tuple(cast<int32_t>(cast<int8_t>(dst[i])),cast<int32_t>(cast<int8_t>(src[i])));
a = std::get<0>(tmp138);
b = std::get<1>(tmp138);
}
else {
auto tmp139 = std::make_tuple(cast<int32_t>(dst[i]),cast<int32_t>(src[i]));
a = std::get<0>(tmp139);
b = std::get<1>(tmp139);
}
int32_t v = cast<int32_t>((a + b));
if (sub_) {
v = cast<int32_t>((a - b));
}
if (signed_) {
r[i] = x86_satI8(v);
}
else {
r[i] = x86_satU8(v);
}
}
}}
x86_CPU_setIntReg(c,reg,wide,r);
return true;
break;}
case cast<uint8_t>(213ULL):case cast<uint8_t>(229ULL):case cast<uint8_t>(228ULL):{
auto tmp140 = x86_CPU_intOperands(c,k);
uint8_t reg = std::get<0>(tmp140);
std::array<uint8_t,16> src = std::get<1>(tmp140);
bool wide = std::get<3>(tmp140);
std::array<uint8_t,16> dst = x86_CPU_intReg(c,reg,wide);
int64_t n = x86_intWidth(wide);
std::array<uint8_t,16> r={};
{int64_t i = cast<int64_t>(0ULL);for (;(i < divi<int64_t>(n,cast<int64_t>(2ULL)));i++){
{
switch(op){
case cast<uint8_t>(213ULL):{
int32_t v = cast<int32_t>((cast<int32_t>(cast<int16_t>(x86_w16(sub(dst,cast<int64_t>((i * cast<int64_t>(2ULL))),len(dst))))) * cast<int32_t>(cast<int16_t>(x86_w16(sub(src,cast<int64_t>((i * cast<int64_t>(2ULL))),len(src)))))));
x86_putw16(sub(r,cast<int64_t>((i * cast<int64_t>(2ULL))),len(r)),cast<uint16_t>(v));
break;}
case cast<uint8_t>(229ULL):{
int32_t v = cast<int32_t>((cast<int32_t>(cast<int16_t>(x86_w16(sub(dst,cast<int64_t>((i * cast<int64_t>(2ULL))),len(dst))))) * cast<int32_t>(cast<int16_t>(x86_w16(sub(src,cast<int64_t>((i * cast<int64_t>(2ULL))),len(src)))))));
x86_putw16(sub(r,cast<int64_t>((i * cast<int64_t>(2ULL))),len(r)),cast<uint16_t>(shr<uint32_t>(cast<uint32_t>(v),cast<int64_t>(16ULL))));
break;}
case cast<uint8_t>(228ULL):{
uint32_t v = cast<uint32_t>((cast<uint32_t>(x86_w16(sub(dst,cast<int64_t>((i * cast<int64_t>(2ULL))),len(dst)))) * cast<uint32_t>(x86_w16(sub(src,cast<int64_t>((i * cast<int64_t>(2ULL))),len(src))))));
x86_putw16(sub(r,cast<int64_t>((i * cast<int64_t>(2ULL))),len(r)),cast<uint16_t>(shr<uint32_t>(v,cast<int64_t>(16ULL))));
break;}
}}
}
}x86_CPU_setIntReg(c,reg,wide,r);
return true;
break;}
case cast<uint8_t>(245ULL):{
auto tmp141 = x86_CPU_intOperands(c,k);
uint8_t reg = std::get<0>(tmp141);
std::array<uint8_t,16> src = std::get<1>(tmp141);
bool wide = std::get<3>(tmp141);
std::array<uint8_t,16> dst = x86_CPU_intReg(c,reg,wide);
int64_t n = x86_intWidth(wide);
std::array<uint8_t,16> r={};
{int64_t i = cast<int64_t>(0ULL);for (;(i < divi<int64_t>(n,cast<int64_t>(4ULL)));i++){
int32_t a0 = cast<int32_t>(cast<int16_t>(x86_w16(sub(dst,cast<int64_t>((i * cast<int64_t>(4ULL))),len(dst)))));
int32_t b0 = cast<int32_t>(cast<int16_t>(x86_w16(sub(src,cast<int64_t>((i * cast<int64_t>(4ULL))),len(src)))));
int32_t a1 = cast<int32_t>(cast<int16_t>(x86_w16(sub(dst,cast<int64_t>((cast<int64_t>((i * cast<int64_t>(4ULL))) + cast<int64_t>(2ULL))),len(dst)))));
int32_t b1 = cast<int32_t>(cast<int16_t>(x86_w16(sub(src,cast<int64_t>((cast<int64_t>((i * cast<int64_t>(4ULL))) + cast<int64_t>(2ULL))),len(src)))));
x86_putle32(sub(r,cast<int64_t>((i * cast<int64_t>(4ULL))),len(r)),cast<uint32_t>(cast<int32_t>((cast<int32_t>((a0 * b0)) + cast<int32_t>((a1 * b1))))));
}
}x86_CPU_setIntReg(c,reg,wide,r);
return true;
break;}
case cast<uint8_t>(244ULL):{
auto tmp142 = x86_CPU_intOperands(c,k);
uint8_t reg = std::get<0>(tmp142);
std::array<uint8_t,16> src = std::get<1>(tmp142);
bool wide = std::get<3>(tmp142);
std::array<uint8_t,16> dst = x86_CPU_intReg(c,reg,wide);
int64_t n = x86_intWidth(wide);
std::array<uint8_t,16> r={};
{int64_t i = cast<int64_t>(0ULL);for (;(i < divi<int64_t>(n,cast<int64_t>(8ULL)));i++){
uint64_t v = cast<uint64_t>((cast<uint64_t>(x86_le32b(sub(dst,cast<int64_t>((i * cast<int64_t>(8ULL))),len(dst)))) * cast<uint64_t>(x86_le32b(sub(src,cast<int64_t>((i * cast<int64_t>(8ULL))),len(src))))));
{int64_t j = cast<int64_t>(0ULL);for (;(j < cast<int64_t>(8ULL));j++){
r[cast<int64_t>((cast<int64_t>((i * cast<int64_t>(8ULL))) + j))] = cast<uint8_t>(shr<uint64_t>(v,(cast<int64_t>((cast<int64_t>(8ULL) * j)))));
}
}}
}x86_CPU_setIntReg(c,reg,wide,r);
return true;
break;}
case cast<uint8_t>(246ULL):{
auto tmp143 = x86_CPU_intOperands(c,k);
uint8_t reg = std::get<0>(tmp143);
std::array<uint8_t,16> src = std::get<1>(tmp143);
bool wide = std::get<3>(tmp143);
std::array<uint8_t,16> dst = x86_CPU_intReg(c,reg,wide);
int64_t n = x86_intWidth(wide);
std::array<uint8_t,16> r={};
{int64_t g = cast<int64_t>(0ULL);for (;(g < divi<int64_t>(n,cast<int64_t>(8ULL)));g++){
uint32_t sum={};
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(8ULL));i++){
int32_t d = cast<int32_t>((cast<int32_t>(dst[cast<int64_t>((cast<int64_t>((g * cast<int64_t>(8ULL))) + i))]) - cast<int32_t>(src[cast<int64_t>((cast<int64_t>((g * cast<int64_t>(8ULL))) + i))])));
if ((d < cast<int32_t>(0ULL))) {
d = cast<int32_t>(-d);
}
sum += cast<uint32_t>(d);
}
}x86_putw16(sub(r,cast<int64_t>((g * cast<int64_t>(8ULL))),len(r)),cast<uint16_t>(sum));
}
}x86_CPU_setIntReg(c,reg,wide,r);
return true;
break;}
case cast<uint8_t>(218ULL):case cast<uint8_t>(222ULL):{
auto tmp144 = x86_CPU_intOperands(c,k);
uint8_t reg = std::get<0>(tmp144);
std::array<uint8_t,16> src = std::get<1>(tmp144);
bool wide = std::get<3>(tmp144);
std::array<uint8_t,16> dst = x86_CPU_intReg(c,reg,wide);
int64_t n = x86_intWidth(wide);
std::array<uint8_t,16> r={};
{int64_t i = cast<int64_t>(0ULL);for (;(i < n);i++){
if ((((op == cast<uint8_t>(218ULL))) == ((src[i] < dst[i])))) {
r[i] = src[i];
}
else {
r[i] = dst[i];
}
}
}x86_CPU_setIntReg(c,reg,wide,r);
return true;
break;}
case cast<uint8_t>(234ULL):case cast<uint8_t>(238ULL):{
auto tmp145 = x86_CPU_intOperands(c,k);
uint8_t reg = std::get<0>(tmp145);
std::array<uint8_t,16> src = std::get<1>(tmp145);
bool wide = std::get<3>(tmp145);
std::array<uint8_t,16> dst = x86_CPU_intReg(c,reg,wide);
int64_t n = x86_intWidth(wide);
std::array<uint8_t,16> r={};
{int64_t i = cast<int64_t>(0ULL);for (;(i < divi<int64_t>(n,cast<int64_t>(2ULL)));i++){
auto tmp146 = std::make_tuple(cast<int16_t>(x86_w16(sub(dst,cast<int64_t>((i * cast<int64_t>(2ULL))),len(dst)))),cast<int16_t>(x86_w16(sub(src,cast<int64_t>((i * cast<int64_t>(2ULL))),len(src)))));
int16_t a = std::get<0>(tmp146);
int16_t b = std::get<1>(tmp146);
int16_t v = a;
if ((((op == cast<uint8_t>(234ULL))) == ((b < a)))) {
v = b;
}
x86_putw16(sub(r,cast<int64_t>((i * cast<int64_t>(2ULL))),len(r)),cast<uint16_t>(v));
}
}x86_CPU_setIntReg(c,reg,wide,r);
return true;
break;}
case cast<uint8_t>(224ULL):case cast<uint8_t>(227ULL):{
auto tmp147 = x86_CPU_intOperands(c,k);
uint8_t reg = std::get<0>(tmp147);
std::array<uint8_t,16> src = std::get<1>(tmp147);
bool wide = std::get<3>(tmp147);
std::array<uint8_t,16> dst = x86_CPU_intReg(c,reg,wide);
int64_t n = x86_intWidth(wide);
std::array<uint8_t,16> r={};
if ((op == cast<uint8_t>(224ULL))) {
{int64_t i = cast<int64_t>(0ULL);for (;(i < n);i++){
r[i] = cast<uint8_t>(shr<uint32_t>((cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(dst[i]) + cast<uint32_t>(src[i]))) + cast<uint32_t>(1ULL)))),cast<int64_t>(1ULL)));
}
}}
else {
{int64_t i = cast<int64_t>(0ULL);for (;(i < divi<int64_t>(n,cast<int64_t>(2ULL)));i++){
x86_putw16(sub(r,cast<int64_t>((i * cast<int64_t>(2ULL))),len(r)),cast<uint16_t>(shr<uint32_t>((cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(x86_w16(sub(dst,cast<int64_t>((i * cast<int64_t>(2ULL))),len(dst)))) + cast<uint32_t>(x86_w16(sub(src,cast<int64_t>((i * cast<int64_t>(2ULL))),len(src)))))) + cast<uint32_t>(1ULL)))),cast<int64_t>(1ULL))));
}
}}
x86_CPU_setIntReg(c,reg,wide,r);
return true;
break;}
case cast<uint8_t>(196ULL):{
auto tmp148 = x86_CPU_modrmE(c);
uint8_t reg = std::get<0>(tmp148);
x86_ea o = std::get<1>(tmp148);
int64_t imm = cast<int64_t>(x86_CPU_fetch8(c));
uint16_t w={};
if (o.isReg) {
w = cast<uint16_t>(c->Regs[o.reg]);
}
else {
w = cast<uint16_t>(x86_CPU_memRead(c,o.base,o.off,cast<int64_t>(2ULL)));
}
if ((k == cast<x86_sseKind>(1ULL))) {
std::array<uint8_t,16> v = c->XMM[reg];
x86_putw16(sub(v,cast<int64_t>(((cast<int64_t>((imm & cast<int64_t>(7ULL)))) * cast<int64_t>(2ULL))),len(v)),w);
c->XMM[reg] = v;
}
else {
std::array<uint8_t,8> v = c->MMX[cast<uint8_t>((reg & cast<uint8_t>(7ULL)))];
x86_putw16(sub(v,cast<int64_t>(((cast<int64_t>((imm & cast<int64_t>(3ULL)))) * cast<int64_t>(2ULL))),len(v)),w);
c->MMX[cast<uint8_t>((reg & cast<uint8_t>(7ULL)))] = v;
}
return true;
break;}
case cast<uint8_t>(197ULL):{
auto tmp149 = x86_CPU_modrmE(c);
uint8_t reg = std::get<0>(tmp149);
x86_ea o = std::get<1>(tmp149);
int64_t imm = cast<int64_t>(x86_CPU_fetch8(c));
if ((!o.isReg)) {
return false;
}
if ((k == cast<x86_sseKind>(1ULL))) {
c->Regs[reg] = cast<uint32_t>(x86_w16(sub(c->XMM[o.reg],cast<int64_t>(((cast<int64_t>((imm & cast<int64_t>(7ULL)))) * cast<int64_t>(2ULL))),len(c->XMM[o.reg]))));
}
else {
c->Regs[reg] = cast<uint32_t>(x86_w16(sub(c->MMX[cast<uint8_t>((o.reg & cast<uint8_t>(7ULL)))],cast<int64_t>(((cast<int64_t>((imm & cast<int64_t>(3ULL)))) * cast<int64_t>(2ULL))),len(c->MMX[cast<uint8_t>((o.reg & cast<uint8_t>(7ULL)))]))));
}
return true;
break;}
case cast<uint8_t>(215ULL):{
auto tmp150 = x86_CPU_modrmE(c);
uint8_t reg = std::get<0>(tmp150);
x86_ea o = std::get<1>(tmp150);
if ((!o.isReg)) {
return false;
}
uint32_t mask={};
uint32_t n={};
if ((k == cast<x86_sseKind>(1ULL))) {
n = cast<uint32_t>(16ULL);
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < n);i++){
if ((cast<uint8_t>((c->XMM[o.reg][i] & cast<uint8_t>(128ULL))) != cast<uint8_t>(0ULL))) {
mask |= shl<uint32_t>(cast<uint32_t>(1ULL),i);
}
}
}}
else {
n = cast<uint32_t>(8ULL);
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < n);i++){
if ((cast<uint8_t>((c->MMX[cast<uint8_t>((o.reg & cast<uint8_t>(7ULL)))][i] & cast<uint8_t>(128ULL))) != cast<uint8_t>(0ULL))) {
mask |= shl<uint32_t>(cast<uint32_t>(1ULL),i);
}
}
}}
c->Regs[reg] = mask;
return true;
break;}
case cast<uint8_t>(231ULL):{
auto tmp151 = x86_CPU_modrmE(c);
uint8_t reg = std::get<0>(tmp151);
x86_ea o = std::get<1>(tmp151);
if (o.isReg) {
return false;
}
if ((k == cast<x86_sseKind>(1ULL))) {
x86_CPU_sseStoreRM(c,o,c->XMM[reg],cast<int64_t>(16ULL));
}
else {
x86_CPU_mmxStoreRM(c,o,c->MMX[cast<uint8_t>((reg & cast<uint8_t>(7ULL)))]);
}
return true;
break;}
}}
return false;
}
}
// tools/cpu/x86/mmxint.go:635:1
uint16_t x86_satI16Clamp(int32_t v){
{
if ((v < cast<int32_t>(-32768ULL))) {
return cast<uint16_t>(32768ULL);
}
if ((v > cast<int32_t>(32767ULL))) {
return cast<uint16_t>(32767ULL);
}
return cast<uint16_t>(v);
}
}
// tools/cpu/x86/mmxint.go:646:1
uint64_t x86_shiftLeft(uint64_t v,uint32_t bits,uint32_t cnt){
{
if ((cnt >= bits)) {
return cast<uint64_t>(0ULL);
}
uint64_t mask = cast<uint64_t>((shl<uint64_t>(cast<uint64_t>(1ULL),bits) - cast<uint64_t>(1ULL)));
if ((bits == cast<uint32_t>(64ULL))) {
mask = cast<uint64_t>(18446744073709551615ULL);
}
return cast<uint64_t>(((shl<uint64_t>(v,cnt)) & mask));
}
}
// tools/cpu/x86/mmxint.go:657:1
uint64_t x86_shiftRightLogical(uint64_t v,uint32_t bits,uint32_t cnt){
{
if ((cnt >= bits)) {
return cast<uint64_t>(0ULL);
}
return shr<uint64_t>(v,cnt);
}
}
// tools/cpu/x86/mmxint.go:664:1
uint64_t x86_shiftRightArith(uint64_t v,uint32_t bits,uint32_t cnt){
{
if ((cnt >= bits)) {
cnt = cast<uint32_t>((bits - cast<uint32_t>(1ULL)));
}
uint32_t sh = cast<uint32_t>((cast<uint32_t>(64ULL) - bits));
int64_t s = shr<int64_t>(cast<int64_t>(shl<uint64_t>(v,sh)),sh);
uint64_t r = cast<uint64_t>(shr<int64_t>(s,cnt));
uint64_t mask = cast<uint64_t>((shl<uint64_t>(cast<uint64_t>(1ULL),bits) - cast<uint64_t>(1ULL)));
if ((bits == cast<uint32_t>(64ULL))) {
mask = cast<uint64_t>(18446744073709551615ULL);
}
return cast<uint64_t>((r & mask));
}
}
// tools/cpu/x86/mmxint.go:680:1
void x86_shiftLanes(std::array<uint8_t,16>* v,int64_t n,int64_t elem,uint32_t cnt,std::function<uint64_t(uint64_t,uint32_t,uint32_t)> fn){
{
uint32_t bits = cast<uint32_t>(cast<int64_t>((elem * cast<int64_t>(8ULL))));
{int64_t i = cast<int64_t>(0ULL);for (;(i < n);i += elem){
uint64_t lane={};
{int64_t j = cast<int64_t>(0ULL);for (;(j < elem);j++){
lane |= shl<uint64_t>(cast<uint64_t>((*v)[cast<int64_t>((i + j))]),(cast<int64_t>((cast<int64_t>(8ULL) * j))));
}
}lane = fn(lane,bits,cnt);
{int64_t j = cast<int64_t>(0ULL);for (;(j < elem);j++){
(*v)[cast<int64_t>((i + j))] = cast<uint8_t>(shr<uint64_t>(lane,(cast<int64_t>((cast<int64_t>(8ULL) * j)))));
}
}}
}}
}
// tools/cpu/x86/mmxint.go:695:1
void x86_byteShift(std::array<uint8_t,16>* v,int64_t n,bool left){
{
if ((n > cast<int64_t>(16ULL))) {
n = cast<int64_t>(16ULL);
}
std::array<uint8_t,16> r={};
if (left) {
{int64_t i = cast<int64_t>(15ULL);for (;(i >= n);i--){
r[i] = (*v)[cast<int64_t>((i - n))];
}
}}
else {
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>((cast<int64_t>(16ULL) - n)));i++){
r[i] = (*v)[cast<int64_t>((i + n))];
}
}}
(*v) = r;
}
}
// tools/cpu/x86/sse.go:32:1
x86_sseKind x86_CPU_sseKindOf(x86_CPU* c,uint8_t rep){
{
{
if ((rep == cast<uint8_t>(243ULL))){
return cast<x86_sseKind>(2ULL);
}
else if ((rep == cast<uint8_t>(242ULL))){
return cast<x86_sseKind>(3ULL);
}
else if ((c->dOpsize == cast<int64_t>(16ULL))){
return cast<x86_sseKind>(1ULL);
}
else {
return cast<x86_sseKind>(0ULL);
}
}
tmp152:;
}
}
// tools/cpu/x86/sse.go:49:1
std::array<uint8_t,16> x86_CPU_sseRM(x86_CPU* c,x86_ea o,int64_t n){
{
std::array<uint8_t,16> b={};
if (o.isReg) {
return c->XMM[o.reg];
}
{int64_t i = cast<int64_t>(0ULL);for (;((i < n) && (i < cast<int64_t>(16ULL)));i++){
b[i] = cast<uint8_t>(x86_CPU_memRead(c,o.base,cast<uint32_t>((o.off + cast<uint32_t>(i))),cast<int64_t>(1ULL)));
}
}return b;
}
}
// tools/cpu/x86/sse.go:61:1
std::array<uint8_t,8> x86_CPU_mmxRM(x86_CPU* c,x86_ea o){
{
std::array<uint8_t,8> b={};
if (o.isReg) {
return c->MMX[cast<uint8_t>((o.reg & cast<uint8_t>(7ULL)))];
}
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(8ULL));i++){
b[i] = cast<uint8_t>(x86_CPU_memRead(c,o.base,cast<uint32_t>((o.off + cast<uint32_t>(i))),cast<int64_t>(1ULL)));
}
}return b;
}
}
// tools/cpu/x86/sse.go:73:1
void x86_CPU_mmxStoreRM(x86_CPU* c,x86_ea o,std::array<uint8_t,8> v){
{
if (o.isReg) {
c->MMX[cast<uint8_t>((o.reg & cast<uint8_t>(7ULL)))] = v;
return ;
}
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(8ULL));i++){
x86_CPU_memWrite(c,o.base,cast<uint32_t>((o.off + cast<uint32_t>(i))),cast<int64_t>(1ULL),cast<uint32_t>(v[i]));
}
}}
}
// tools/cpu/x86/sse.go:84:1
void x86_CPU_sseStoreRM(x86_CPU* c,x86_ea o,std::array<uint8_t,16> v,int64_t n){
{
if (o.isReg) {
c->XMM[o.reg] = v;
return ;
}
{int64_t i = cast<int64_t>(0ULL);for (;(i < n);i++){
x86_CPU_memWrite(c,o.base,cast<uint32_t>((o.off + cast<uint32_t>(i))),cast<int64_t>(1ULL),cast<uint32_t>(v[i]));
}
}}
}
// tools/cpu/x86/sse.go:95:1
float x86_f32Lane(std::array<uint8_t,16> b,int64_t i){
{
return go_math_Float32frombits(x86_le32b(sub(b,cast<int64_t>((i * cast<int64_t>(4ULL))),len(b))));
}
}
// tools/cpu/x86/sse.go:96:1
double x86_f64Lane(std::array<uint8_t,16> b,int64_t i){
{
return go_math_Float64frombits(x86_le64b(sub(b,cast<int64_t>((i * cast<int64_t>(8ULL))),len(b))));
}
}
// tools/cpu/x86/sse.go:97:1
void x86_setF32Lane(std::array<uint8_t,16>* b,int64_t i,float v){
{
x86_putle32(sub(b,cast<int64_t>((i * cast<int64_t>(4ULL))),len(b)),go_math_Float32bits(v));
}
}
// tools/cpu/x86/sse.go:98:1
void x86_setF64Lane(std::array<uint8_t,16>* b,int64_t i,double v){
{
x86_putle64(sub(b,cast<int64_t>((i * cast<int64_t>(8ULL))),len(b)),go_math_Float64bits(v));
}
}
// tools/cpu/x86/sse.go:100:1
uint32_t x86_le32b(Slice<uint8_t> b){
{
return cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(b[cast<int64_t>(0ULL)]) | shl<uint32_t>(cast<uint32_t>(b[cast<int64_t>(1ULL)]),cast<int64_t>(8ULL)))) | shl<uint32_t>(cast<uint32_t>(b[cast<int64_t>(2ULL)]),cast<int64_t>(16ULL)))) | shl<uint32_t>(cast<uint32_t>(b[cast<int64_t>(3ULL)]),cast<int64_t>(24ULL))));
}
}
// tools/cpu/x86/sse.go:103:1
uint64_t x86_le64b(Slice<uint8_t> b){
{
return cast<uint64_t>((cast<uint64_t>(x86_le32b(b)) | shl<uint64_t>(cast<uint64_t>(x86_le32b(sub(b,cast<int64_t>(4ULL),len(b)))),cast<int64_t>(32ULL))));
}
}
// tools/cpu/x86/sse.go:104:1
void x86_putle32(Slice<uint8_t> b,uint32_t v){
{
auto tmp153 = std::make_tuple(cast<uint8_t>(v),cast<uint8_t>(shr<uint32_t>(v,cast<int64_t>(8ULL))),cast<uint8_t>(shr<uint32_t>(v,cast<int64_t>(16ULL))),cast<uint8_t>(shr<uint32_t>(v,cast<int64_t>(24ULL))));
b[cast<int64_t>(0ULL)] = std::get<0>(tmp153);
b[cast<int64_t>(1ULL)] = std::get<1>(tmp153);
b[cast<int64_t>(2ULL)] = std::get<2>(tmp153);
b[cast<int64_t>(3ULL)] = std::get<3>(tmp153);
}
}
// tools/cpu/x86/sse.go:107:1
void x86_putle64(Slice<uint8_t> b,uint64_t v){
{
x86_putle32(b,cast<uint32_t>(v));
x86_putle32(sub(b,cast<int64_t>(4ULL),len(b)),cast<uint32_t>(shr<uint64_t>(v,cast<int64_t>(32ULL))));
}
}
// tools/cpu/x86/sse.go:113:1
int64_t x86_sseWidth(x86_sseKind k){
{
{
switch(k){
case cast<x86_sseKind>(2ULL):{
return cast<int64_t>(4ULL);
break;}
case cast<x86_sseKind>(3ULL):{
return cast<int64_t>(8ULL);
break;}
default:{
return cast<int64_t>(16ULL);
break;}
}}
}
}
// tools/cpu/x86/sse.go:126:1
bool x86_CPU_execSSE(x86_CPU* c,uint8_t op,uint8_t rep){
{
x86_sseKind k = x86_CPU_sseKindOf(c,rep);
{
switch(op){
case cast<uint8_t>(16ULL):{
auto tmp154 = x86_CPU_modrmE(c);
uint8_t reg = std::get<0>(tmp154);
x86_ea o = std::get<1>(tmp154);
std::array<uint8_t,16> src = x86_CPU_sseRM(c,o,x86_sseWidth(k));
std::array<uint8_t,16> dst = c->XMM[reg];
{
switch(k){
case cast<x86_sseKind>(2ULL):{
gcopy(sub(dst,cast<int64_t>(0ULL),cast<int64_t>(4ULL)),sub(src,cast<int64_t>(0ULL),cast<int64_t>(4ULL)));
if ((!o.isReg)) {
{int64_t i = cast<int64_t>(4ULL);for (;(i < cast<int64_t>(16ULL));i++){
dst[i] = cast<uint8_t>(0ULL);
}
}}
break;}
case cast<x86_sseKind>(3ULL):{
gcopy(sub(dst,cast<int64_t>(0ULL),cast<int64_t>(8ULL)),sub(src,cast<int64_t>(0ULL),cast<int64_t>(8ULL)));
if ((!o.isReg)) {
{int64_t i = cast<int64_t>(8ULL);for (;(i < cast<int64_t>(16ULL));i++){
dst[i] = cast<uint8_t>(0ULL);
}
}}
break;}
default:{
dst = src;
break;}
}}
c->XMM[reg] = dst;
return true;
break;}
case cast<uint8_t>(17ULL):{
auto tmp155 = x86_CPU_modrmE(c);
uint8_t reg = std::get<0>(tmp155);
x86_ea o = std::get<1>(tmp155);
std::array<uint8_t,16> src = c->XMM[reg];
x86_CPU_sseStoreRM(c,o,src,x86_sseWidth(k));
return true;
break;}
case cast<uint8_t>(40ULL):case cast<uint8_t>(41ULL):{
auto tmp156 = x86_CPU_modrmE(c);
uint8_t reg = std::get<0>(tmp156);
x86_ea o = std::get<1>(tmp156);
if ((op == cast<uint8_t>(40ULL))) {
c->XMM[reg] = x86_CPU_sseRM(c,o,cast<int64_t>(16ULL));
}
else {
x86_CPU_sseStoreRM(c,o,c->XMM[reg],cast<int64_t>(16ULL));
}
return true;
break;}
case cast<uint8_t>(43ULL):{
auto tmp157 = x86_CPU_modrmE(c);
uint8_t reg = std::get<0>(tmp157);
x86_ea o = std::get<1>(tmp157);
x86_CPU_sseStoreRM(c,o,c->XMM[reg],cast<int64_t>(16ULL));
return true;
break;}
case cast<uint8_t>(18ULL):case cast<uint8_t>(22ULL):{
auto tmp158 = x86_CPU_modrmE(c);
uint8_t reg = std::get<0>(tmp158);
x86_ea o = std::get<1>(tmp158);
std::array<uint8_t,16> src = x86_CPU_sseRM(c,o,cast<int64_t>(8ULL));
std::array<uint8_t,16> dst = c->XMM[reg];
int64_t half = cast<int64_t>(0ULL);
if ((op == cast<uint8_t>(22ULL))) {
half = cast<int64_t>(8ULL);
}
gcopy(sub(dst,half,cast<int64_t>((half + cast<int64_t>(8ULL)))),sub(src,cast<int64_t>(0ULL),cast<int64_t>(8ULL)));
c->XMM[reg] = dst;
return true;
break;}
case cast<uint8_t>(19ULL):case cast<uint8_t>(23ULL):{
auto tmp159 = x86_CPU_modrmE(c);
uint8_t reg = std::get<0>(tmp159);
x86_ea o = std::get<1>(tmp159);
std::array<uint8_t,16> src = c->XMM[reg];
int64_t half = cast<int64_t>(0ULL);
if ((op == cast<uint8_t>(23ULL))) {
half = cast<int64_t>(8ULL);
}
std::array<uint8_t,16> v={};
gcopy(sub(v,cast<int64_t>(0ULL),cast<int64_t>(8ULL)),sub(src,half,cast<int64_t>((half + cast<int64_t>(8ULL)))));
x86_CPU_sseStoreRM(c,o,v,cast<int64_t>(8ULL));
return true;
break;}
case cast<uint8_t>(110ULL):{
auto tmp160 = x86_CPU_modrmE(c);
uint8_t reg = std::get<0>(tmp160);
x86_ea o = std::get<1>(tmp160);
uint32_t v = x86_CPU_rEA(c,o,cast<int64_t>(4ULL));
if ((k == cast<x86_sseKind>(1ULL))) {
std::array<uint8_t,16> b={};
x86_putle32(sub(b,cast<int64_t>(0ULL),len(b)),v);
c->XMM[reg] = b;
}
else {
std::array<uint8_t,8> b={};
x86_putle32(sub(b,cast<int64_t>(0ULL),len(b)),v);
c->MMX[cast<uint8_t>((reg & cast<uint8_t>(7ULL)))] = b;
}
return true;
break;}
case cast<uint8_t>(126ULL):{
auto tmp161 = x86_CPU_modrmE(c);
uint8_t reg = std::get<0>(tmp161);
x86_ea o = std::get<1>(tmp161);
if ((rep == cast<uint8_t>(243ULL))) {
std::array<uint8_t,16> src = x86_CPU_sseRM(c,o,cast<int64_t>(8ULL));
std::array<uint8_t,16> b={};
gcopy(sub(b,cast<int64_t>(0ULL),cast<int64_t>(8ULL)),sub(src,cast<int64_t>(0ULL),cast<int64_t>(8ULL)));
c->XMM[reg] = b;
return true;
}
if ((k == cast<x86_sseKind>(1ULL))) {
x86_CPU_wEA(c,o,cast<int64_t>(4ULL),x86_le32b(sub(c->XMM[reg],cast<int64_t>(0ULL),len(c->XMM[reg]))));
return true;
}
x86_CPU_wEA(c,o,cast<int64_t>(4ULL),x86_le32b(sub(c->MMX[cast<uint8_t>((reg & cast<uint8_t>(7ULL)))],cast<int64_t>(0ULL),len(c->MMX[cast<uint8_t>((reg & cast<uint8_t>(7ULL)))]))));
return true;
break;}
case cast<uint8_t>(119ULL):{
return true;
break;}
case cast<uint8_t>(111ULL):{
auto tmp162 = x86_CPU_modrmE(c);
uint8_t reg = std::get<0>(tmp162);
x86_ea o = std::get<1>(tmp162);
if (((k == cast<x86_sseKind>(1ULL)) || (rep == cast<uint8_t>(243ULL)))) {
c->XMM[reg] = x86_CPU_sseRM(c,o,cast<int64_t>(16ULL));
}
else {
c->MMX[cast<uint8_t>((reg & cast<uint8_t>(7ULL)))] = x86_CPU_mmxRM(c,o);
}
return true;
break;}
case cast<uint8_t>(127ULL):{
auto tmp163 = x86_CPU_modrmE(c);
uint8_t reg = std::get<0>(tmp163);
x86_ea o = std::get<1>(tmp163);
if (((k == cast<x86_sseKind>(1ULL)) || (rep == cast<uint8_t>(243ULL)))) {
x86_CPU_sseStoreRM(c,o,c->XMM[reg],cast<int64_t>(16ULL));
}
else {
x86_CPU_mmxStoreRM(c,o,c->MMX[cast<uint8_t>((reg & cast<uint8_t>(7ULL)))]);
}
return true;
break;}
case cast<uint8_t>(214ULL):{
if ((k != cast<x86_sseKind>(1ULL))) {
return false;
}
auto tmp164 = x86_CPU_modrmE(c);
uint8_t reg = std::get<0>(tmp164);
x86_ea o = std::get<1>(tmp164);
std::array<uint8_t,16> v={};
gcopy(sub(v,cast<int64_t>(0ULL),cast<int64_t>(8ULL)),sub(c->XMM[reg],cast<int64_t>(0ULL),cast<int64_t>(8ULL)));
x86_CPU_sseStoreRM(c,o,v,cast<int64_t>(8ULL));
return true;
break;}
case cast<uint8_t>(198ULL):{
auto tmp165 = x86_CPU_modrmE(c);
uint8_t reg = std::get<0>(tmp165);
x86_ea o = std::get<1>(tmp165);
auto tmp166 = std::make_tuple(c->XMM[reg],x86_CPU_sseRM(c,o,cast<int64_t>(16ULL)));
std::array<uint8_t,16> a = std::get<0>(tmp166);
std::array<uint8_t,16> b = std::get<1>(tmp166);
uint8_t imm = cast<uint8_t>(x86_CPU_fetch8(c));
std::array<uint8_t,16> r={};
if ((k == cast<x86_sseKind>(1ULL))) {
gcopy(sub(r,cast<int64_t>(0ULL),cast<int64_t>(8ULL)),sub(a,cast<uint8_t>(((cast<uint8_t>((imm & cast<uint8_t>(1ULL)))) * cast<uint8_t>(8ULL))),cast<uint8_t>((cast<uint8_t>(((cast<uint8_t>((imm & cast<uint8_t>(1ULL)))) * cast<uint8_t>(8ULL))) + cast<uint8_t>(8ULL)))));
gcopy(sub(r,cast<int64_t>(8ULL),cast<int64_t>(16ULL)),sub(b,cast<uint8_t>(((cast<uint8_t>(((shr<uint8_t>(imm,cast<int64_t>(1ULL))) & cast<uint8_t>(1ULL)))) * cast<uint8_t>(8ULL))),cast<uint8_t>((cast<uint8_t>(((cast<uint8_t>(((shr<uint8_t>(imm,cast<int64_t>(1ULL))) & cast<uint8_t>(1ULL)))) * cast<uint8_t>(8ULL))) + cast<uint8_t>(8ULL)))));
}
else {
uint8_t l0 = cast<uint8_t>(((shr<uint8_t>(imm,cast<int64_t>(0ULL))) & cast<uint8_t>(3ULL)));
uint8_t l1 = cast<uint8_t>(((shr<uint8_t>(imm,cast<int64_t>(2ULL))) & cast<uint8_t>(3ULL)));
uint8_t l2 = cast<uint8_t>(((shr<uint8_t>(imm,cast<int64_t>(4ULL))) & cast<uint8_t>(3ULL)));
uint8_t l3 = cast<uint8_t>(((shr<uint8_t>(imm,cast<int64_t>(6ULL))) & cast<uint8_t>(3ULL)));
gcopy(sub(r,cast<int64_t>(0ULL),cast<int64_t>(4ULL)),sub(a,cast<uint8_t>((l0 * cast<uint8_t>(4ULL))),cast<uint8_t>((cast<uint8_t>((l0 * cast<uint8_t>(4ULL))) + cast<uint8_t>(4ULL)))));
gcopy(sub(r,cast<int64_t>(4ULL),cast<int64_t>(8ULL)),sub(a,cast<uint8_t>((l1 * cast<uint8_t>(4ULL))),cast<uint8_t>((cast<uint8_t>((l1 * cast<uint8_t>(4ULL))) + cast<uint8_t>(4ULL)))));
gcopy(sub(r,cast<int64_t>(8ULL),cast<int64_t>(12ULL)),sub(b,cast<uint8_t>((l2 * cast<uint8_t>(4ULL))),cast<uint8_t>((cast<uint8_t>((l2 * cast<uint8_t>(4ULL))) + cast<uint8_t>(4ULL)))));
gcopy(sub(r,cast<int64_t>(12ULL),cast<int64_t>(16ULL)),sub(b,cast<uint8_t>((l3 * cast<uint8_t>(4ULL))),cast<uint8_t>((cast<uint8_t>((l3 * cast<uint8_t>(4ULL))) + cast<uint8_t>(4ULL)))));
}
c->XMM[reg] = r;
return true;
break;}
case cast<uint8_t>(20ULL):case cast<uint8_t>(21ULL):{
auto tmp167 = x86_CPU_modrmE(c);
uint8_t reg = std::get<0>(tmp167);
x86_ea o = std::get<1>(tmp167);
auto tmp168 = std::make_tuple(c->XMM[reg],x86_CPU_sseRM(c,o,cast<int64_t>(16ULL)));
std::array<uint8_t,16> a = std::get<0>(tmp168);
std::array<uint8_t,16> b = std::get<1>(tmp168);
int64_t h = cast<int64_t>(0ULL);
if ((op == cast<uint8_t>(21ULL))) {
h = cast<int64_t>(8ULL);
}
std::array<uint8_t,16> r={};
if ((k == cast<x86_sseKind>(1ULL))) {
gcopy(sub(r,cast<int64_t>(0ULL),cast<int64_t>(8ULL)),sub(a,h,cast<int64_t>((h + cast<int64_t>(8ULL)))));
gcopy(sub(r,cast<int64_t>(8ULL),cast<int64_t>(16ULL)),sub(b,h,cast<int64_t>((h + cast<int64_t>(8ULL)))));
}
else {
gcopy(sub(r,cast<int64_t>(0ULL),cast<int64_t>(4ULL)),sub(a,h,cast<int64_t>((h + cast<int64_t>(4ULL)))));
gcopy(sub(r,cast<int64_t>(4ULL),cast<int64_t>(8ULL)),sub(b,h,cast<int64_t>((h + cast<int64_t>(4ULL)))));
gcopy(sub(r,cast<int64_t>(8ULL),cast<int64_t>(12ULL)),sub(a,cast<int64_t>((h + cast<int64_t>(4ULL))),cast<int64_t>((h + cast<int64_t>(8ULL)))));
gcopy(sub(r,cast<int64_t>(12ULL),cast<int64_t>(16ULL)),sub(b,cast<int64_t>((h + cast<int64_t>(4ULL))),cast<int64_t>((h + cast<int64_t>(8ULL)))));
}
c->XMM[reg] = r;
return true;
break;}
case cast<uint8_t>(84ULL):case cast<uint8_t>(85ULL):case cast<uint8_t>(86ULL):case cast<uint8_t>(87ULL):{
auto tmp169 = x86_CPU_modrmE(c);
uint8_t reg = std::get<0>(tmp169);
x86_ea o = std::get<1>(tmp169);
auto tmp170 = std::make_tuple(c->XMM[reg],x86_CPU_sseRM(c,o,cast<int64_t>(16ULL)));
std::array<uint8_t,16> a = std::get<0>(tmp170);
std::array<uint8_t,16> b = std::get<1>(tmp170);
std::array<uint8_t,16> r={};
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(16ULL));i++){
{
switch(op){
case cast<uint8_t>(84ULL):{
r[i] = cast<uint8_t>((a[i] & b[i]));
break;}
case cast<uint8_t>(85ULL):{
r[i] = cast<uint8_t>(((cast<uint8_t>(~a[i])) & b[i]));
break;}
case cast<uint8_t>(86ULL):{
r[i] = cast<uint8_t>((a[i] | b[i]));
break;}
case cast<uint8_t>(87ULL):{
r[i] = cast<uint8_t>((a[i] ^ b[i]));
break;}
}}
}
}c->XMM[reg] = r;
return true;
break;}
case cast<uint8_t>(88ULL):case cast<uint8_t>(89ULL):case cast<uint8_t>(92ULL):case cast<uint8_t>(93ULL):case cast<uint8_t>(94ULL):case cast<uint8_t>(95ULL):{
auto tmp171 = x86_CPU_modrmE(c);
uint8_t reg = std::get<0>(tmp171);
x86_ea o = std::get<1>(tmp171);
x86_CPU_sseArith(c,op,k,reg,o);
return true;
break;}
case cast<uint8_t>(81ULL):{
auto tmp172 = x86_CPU_modrmE(c);
uint8_t reg = std::get<0>(tmp172);
x86_ea o = std::get<1>(tmp172);
x86_CPU_sseUnary(c,k,reg,o,[&](double x)->double{
return go_math_Sqrt(x);
}
);
return true;
break;}
case cast<uint8_t>(83ULL):{
auto tmp173 = x86_CPU_modrmE(c);
uint8_t reg = std::get<0>(tmp173);
x86_ea o = std::get<1>(tmp173);
x86_CPU_sseUnary(c,k,reg,o,[&](double x)->double{
return (cast<double>(1.00000000000000000e+00) / x);
}
);
return true;
break;}
case cast<uint8_t>(82ULL):{
auto tmp174 = x86_CPU_modrmE(c);
uint8_t reg = std::get<0>(tmp174);
x86_ea o = std::get<1>(tmp174);
x86_CPU_sseUnary(c,k,reg,o,[&](double x)->double{
return (cast<double>(1.00000000000000000e+00) / go_math_Sqrt(x));
}
);
return true;
break;}
case cast<uint8_t>(42ULL):{
auto tmp175 = x86_CPU_modrmE(c);
uint8_t reg = std::get<0>(tmp175);
x86_ea o = std::get<1>(tmp175);
std::array<uint8_t,16> dst = c->XMM[reg];
{
switch(k){
case cast<x86_sseKind>(2ULL):case cast<x86_sseKind>(3ULL):{
int32_t iv = cast<int32_t>(x86_CPU_rEA(c,o,cast<int64_t>(4ULL)));
if ((k == cast<x86_sseKind>(3ULL))) {
x86_setF64Lane((&dst),cast<int64_t>(0ULL),cast<double>(iv));
}
else {
x86_setF32Lane((&dst),cast<int64_t>(0ULL),cast<float>(iv));
}
break;}
default:{
std::array<uint8_t,8> src = x86_CPU_mmxRM(c,o);
auto tmp176 = std::make_tuple(cast<int32_t>(x86_le32b(sub(src,cast<int64_t>(0ULL),len(src)))),cast<int32_t>(x86_le32b(sub(src,cast<int64_t>(4ULL),len(src)))));
int32_t a = std::get<0>(tmp176);
int32_t b = std::get<1>(tmp176);
if ((k == cast<x86_sseKind>(1ULL))) {
x86_setF64Lane((&dst),cast<int64_t>(0ULL),cast<double>(a));
x86_setF64Lane((&dst),cast<int64_t>(1ULL),cast<double>(b));
}
else {
x86_setF32Lane((&dst),cast<int64_t>(0ULL),cast<float>(a));
x86_setF32Lane((&dst),cast<int64_t>(1ULL),cast<float>(b));
}
break;}
}}
c->XMM[reg] = dst;
return true;
break;}
case cast<uint8_t>(44ULL):case cast<uint8_t>(45ULL):{
auto tmp177 = x86_CPU_modrmE(c);
uint8_t reg = std::get<0>(tmp177);
x86_ea o = std::get<1>(tmp177);
bool trunc = (op == cast<uint8_t>(44ULL));
auto cvt = [&](double f)->uint32_t{
if (trunc) {
return cast<uint32_t>(cast<int32_t>(f));
}
return cast<uint32_t>(cast<int32_t>(go_math_RoundToEven(f)));
}
;
{
switch(k){
case cast<x86_sseKind>(2ULL):case cast<x86_sseKind>(3ULL):{
std::array<uint8_t,16> src = x86_CPU_sseRM(c,o,x86_sseWidth(k));
double f={};
if ((k == cast<x86_sseKind>(3ULL))) {
f = x86_f64Lane(src,cast<int64_t>(0ULL));
}
else {
f = cast<double>(x86_f32Lane(src,cast<int64_t>(0ULL)));
}
x86_CPU_setReg(c,reg,cast<int64_t>(4ULL),cvt(f));
break;}
default:{
int64_t n = cast<int64_t>(8ULL);
if ((k == cast<x86_sseKind>(1ULL))) {
n = cast<int64_t>(16ULL);
}
std::array<uint8_t,16> src = x86_CPU_sseRM(c,o,n);
std::array<uint8_t,8> m={};
if ((k == cast<x86_sseKind>(1ULL))) {
x86_putle32(sub(m,cast<int64_t>(0ULL),len(m)),cvt(x86_f64Lane(src,cast<int64_t>(0ULL))));
x86_putle32(sub(m,cast<int64_t>(4ULL),len(m)),cvt(x86_f64Lane(src,cast<int64_t>(1ULL))));
}
else {
x86_putle32(sub(m,cast<int64_t>(0ULL),len(m)),cvt(cast<double>(x86_f32Lane(src,cast<int64_t>(0ULL)))));
x86_putle32(sub(m,cast<int64_t>(4ULL),len(m)),cvt(cast<double>(x86_f32Lane(src,cast<int64_t>(1ULL)))));
}
c->MMX[cast<uint8_t>((reg & cast<uint8_t>(7ULL)))] = m;
break;}
}}
return true;
break;}
case cast<uint8_t>(90ULL):{
auto tmp178 = x86_CPU_modrmE(c);
uint8_t reg = std::get<0>(tmp178);
x86_ea o = std::get<1>(tmp178);
std::array<uint8_t,16> src = x86_CPU_sseRM(c,o,x86_sseWidth(k));
std::array<uint8_t,16> dst = c->XMM[reg];
{
switch(k){
case cast<x86_sseKind>(2ULL):{
x86_setF64Lane((&dst),cast<int64_t>(0ULL),cast<double>(x86_f32Lane(src,cast<int64_t>(0ULL))));
break;}
case cast<x86_sseKind>(3ULL):{
x86_setF32Lane((&dst),cast<int64_t>(0ULL),cast<float>(x86_f64Lane(src,cast<int64_t>(0ULL))));
break;}
case cast<x86_sseKind>(0ULL):{
x86_setF64Lane((&dst),cast<int64_t>(0ULL),cast<double>(x86_f32Lane(src,cast<int64_t>(0ULL))));
x86_setF64Lane((&dst),cast<int64_t>(1ULL),cast<double>(x86_f32Lane(src,cast<int64_t>(1ULL))));
break;}
default:{
x86_setF32Lane((&dst),cast<int64_t>(0ULL),cast<float>(x86_f64Lane(src,cast<int64_t>(0ULL))));
x86_setF32Lane((&dst),cast<int64_t>(1ULL),cast<float>(x86_f64Lane(src,cast<int64_t>(1ULL))));
break;}
}}
c->XMM[reg] = dst;
return true;
break;}
case cast<uint8_t>(91ULL):{
auto tmp179 = x86_CPU_modrmE(c);
uint8_t reg = std::get<0>(tmp179);
x86_ea o = std::get<1>(tmp179);
std::array<uint8_t,16> src = x86_CPU_sseRM(c,o,cast<int64_t>(16ULL));
std::array<uint8_t,16> dst = c->XMM[reg];
if (((rep == cast<uint8_t>(243ULL)) || (c->dOpsize == cast<int64_t>(16ULL)))) {
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(4ULL));i++){
x86_putle32(sub(dst,cast<int64_t>((i * cast<int64_t>(4ULL))),len(dst)),cast<uint32_t>(cast<int32_t>(x86_f32Lane(src,i))));
}
}}
else {
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(4ULL));i++){
x86_setF32Lane((&dst),i,cast<float>(cast<int32_t>(x86_le32b(sub(src,cast<int64_t>((i * cast<int64_t>(4ULL))),len(src))))));
}
}}
c->XMM[reg] = dst;
return true;
break;}
case cast<uint8_t>(46ULL):case cast<uint8_t>(47ULL):{
auto tmp180 = x86_CPU_modrmE(c);
uint8_t reg = std::get<0>(tmp180);
x86_ea o = std::get<1>(tmp180);
std::array<uint8_t,16> src = x86_CPU_sseRM(c,o,x86_sseWidth(k));
double a={};
double b={};
if (((k == cast<x86_sseKind>(3ULL)) || ((k == cast<x86_sseKind>(1ULL))))) {
auto tmp181 = std::make_tuple(x86_f64Lane(c->XMM[reg],cast<int64_t>(0ULL)),x86_f64Lane(src,cast<int64_t>(0ULL)));
a = std::get<0>(tmp181);
b = std::get<1>(tmp181);
}
else {
auto tmp182 = std::make_tuple(cast<double>(x86_f32Lane(c->XMM[reg],cast<int64_t>(0ULL))),cast<double>(x86_f32Lane(src,cast<int64_t>(0ULL))));
a = std::get<0>(tmp182);
b = std::get<1>(tmp182);
}
x86_CPU_sseCompareFlags(c,a,b);
return true;
break;}
case cast<uint8_t>(239ULL):{
if ((k != cast<x86_sseKind>(1ULL))) {
return false;
}
auto tmp183 = x86_CPU_modrmE(c);
uint8_t reg = std::get<0>(tmp183);
x86_ea o = std::get<1>(tmp183);
auto tmp184 = std::make_tuple(c->XMM[reg],x86_CPU_sseRM(c,o,cast<int64_t>(16ULL)));
std::array<uint8_t,16> a = std::get<0>(tmp184);
std::array<uint8_t,16> b = std::get<1>(tmp184);
std::array<uint8_t,16> r={};
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(16ULL));i++){
r[i] = cast<uint8_t>((a[i] ^ b[i]));
}
}c->XMM[reg] = r;
return true;
break;}
}}
return false;
}
}
// tools/cpu/x86/sse.go:462:1
void x86_CPU_sseArith(x86_CPU* c,uint8_t op,x86_sseKind k,uint8_t reg,x86_ea o){
{
std::array<uint8_t,16> src = x86_CPU_sseRM(c,o,x86_sseWidth(k));
std::array<uint8_t,16> dst = c->XMM[reg];
auto f32 = [&](float a,float b)->float{
{
switch(op){
case cast<uint8_t>(88ULL):{
return (a + b);
break;}
case cast<uint8_t>(89ULL):{
return (a * b);
break;}
case cast<uint8_t>(92ULL):{
return (a - b);
break;}
case cast<uint8_t>(93ULL):{
return cast<float>(go_math_Min(cast<double>(a),cast<double>(b)));
break;}
case cast<uint8_t>(94ULL):{
return (a / b);
break;}
default:{
return cast<float>(go_math_Max(cast<double>(a),cast<double>(b)));
break;}
}}
}
;
auto f64 = [&](double a,double b)->double{
{
switch(op){
case cast<uint8_t>(88ULL):{
return (a + b);
break;}
case cast<uint8_t>(89ULL):{
return (a * b);
break;}
case cast<uint8_t>(92ULL):{
return (a - b);
break;}
case cast<uint8_t>(93ULL):{
return go_math_Min(a,b);
break;}
case cast<uint8_t>(94ULL):{
return (a / b);
break;}
default:{
return go_math_Max(a,b);
break;}
}}
}
;
{
switch(k){
case cast<x86_sseKind>(2ULL):{
x86_setF32Lane((&dst),cast<int64_t>(0ULL),f32(x86_f32Lane(dst,cast<int64_t>(0ULL)),x86_f32Lane(src,cast<int64_t>(0ULL))));
break;}
case cast<x86_sseKind>(3ULL):{
x86_setF64Lane((&dst),cast<int64_t>(0ULL),f64(x86_f64Lane(dst,cast<int64_t>(0ULL)),x86_f64Lane(src,cast<int64_t>(0ULL))));
break;}
case cast<x86_sseKind>(0ULL):{
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(4ULL));i++){
x86_setF32Lane((&dst),i,f32(x86_f32Lane(dst,i),x86_f32Lane(src,i)));
}
}break;}
default:{
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(2ULL));i++){
x86_setF64Lane((&dst),i,f64(x86_f64Lane(dst,i),x86_f64Lane(src,i)));
}
}break;}
}}
c->XMM[reg] = dst;
}
}
// tools/cpu/x86/sse.go:515:1
void x86_CPU_sseUnary(x86_CPU* c,x86_sseKind k,uint8_t reg,x86_ea o,std::function<double(double)> fn){
{
std::array<uint8_t,16> src = x86_CPU_sseRM(c,o,x86_sseWidth(k));
std::array<uint8_t,16> dst = c->XMM[reg];
{
switch(k){
case cast<x86_sseKind>(2ULL):{
x86_setF32Lane((&dst),cast<int64_t>(0ULL),cast<float>(fn(cast<double>(x86_f32Lane(src,cast<int64_t>(0ULL))))));
break;}
case cast<x86_sseKind>(3ULL):{
x86_setF64Lane((&dst),cast<int64_t>(0ULL),fn(x86_f64Lane(src,cast<int64_t>(0ULL))));
break;}
case cast<x86_sseKind>(0ULL):{
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(4ULL));i++){
x86_setF32Lane((&dst),i,cast<float>(fn(cast<double>(x86_f32Lane(src,i)))));
}
}break;}
default:{
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(2ULL));i++){
x86_setF64Lane((&dst),i,fn(x86_f64Lane(src,i)));
}
}break;}
}}
c->XMM[reg] = dst;
}
}
// tools/cpu/x86/sse.go:538:1
void x86_CPU_sseCompareFlags(x86_CPU* c,double a,double b){
{
auto tmp185 = std::make_tuple(false,false,false);
c->OF = std::get<0>(tmp185);
c->SF = std::get<1>(tmp185);
c->AF = std::get<2>(tmp185);
{
if ((go_math_IsNaN(a) || go_math_IsNaN(b))){
auto tmp187 = std::make_tuple(true,true,true);
c->ZF = std::get<0>(tmp187);
c->PF = std::get<1>(tmp187);
c->CF = std::get<2>(tmp187);
}
else if ((a > b)){
auto tmp188 = std::make_tuple(false,false,false);
c->ZF = std::get<0>(tmp188);
c->PF = std::get<1>(tmp188);
c->CF = std::get<2>(tmp188);
}
else if ((a < b)){
auto tmp189 = std::make_tuple(false,false,true);
c->ZF = std::get<0>(tmp189);
c->PF = std::get<1>(tmp189);
c->CF = std::get<2>(tmp189);
}
else {
auto tmp190 = std::make_tuple(true,false,false);
c->ZF = std::get<0>(tmp190);
c->PF = std::get<1>(tmp190);
c->CF = std::get<2>(tmp190);
}
}
tmp186:;
}
}
// tools/platform/dos/coff.go:45:1
bool dos_COFFSection_IsText(dos_COFFSection s){
{
return (cast<uint32_t>((s.Flags & cast<uint32_t>(32ULL))) != cast<uint32_t>(0ULL));
}
}
// tools/platform/dos/coff.go:48:1
bool dos_COFFSection_IsBSS(dos_COFFSection s){
{
return (cast<uint32_t>((s.Flags & cast<uint32_t>(128ULL))) != cast<uint32_t>(0ULL));
}
}
// tools/platform/dos/coff.go:67:1
std::tuple<dos_COFF*,Error> dos_ParseGo32COFF(Slice<uint8_t> data){
{
auto tmp1 = dos_ParseMZ(data);
dos_MZ* mz = std::get<0>(tmp1);
Error err = std::get<1>(tmp1);
if (bool(err)) {
return {{},go_fmt_Errorf(std::string("go32: parsing the MZ stub: %w",29),err)};
}
int64_t off = mz->LoadImageEnd;
if (((off <= cast<int64_t>(0ULL)) || (cast<int64_t>((off + cast<int64_t>(20ULL))) > len(data)))) {
return {{},go_fmt_Errorf(std::string("go32: stub image end %#x leaves no room for a COFF header in a %d-byte file",75),off,len(data))};
}
dos_COFF* c = arenaNew(dos_COFF{off,{},{},{},{},{},{},{},{},{}});
auto u16 = [&](int64_t o)->uint16_t{
return le_Uint16(sub(data,o,len(data)));
}
;
auto u32 = [&](int64_t o)->uint32_t{
return le_Uint32(rrBorrow(data,o,len(data)));
}
;
{
uint16_t magic = u16(off);
if ((magic != cast<uint16_t>(332ULL))) {
return {{},go_fmt_Errorf(std::string("go32: at file offset %#x: COFF magic %#04x, want %#04x (i386)",61),off,magic,cast<int64_t>(332ULL))};
}
}
c->NSections = cast<int64_t>(u16(cast<int64_t>((off + cast<int64_t>(2ULL)))));
int64_t optSize = cast<int64_t>(u16(cast<int64_t>((off + cast<int64_t>(16ULL)))));
c->Flags = u16(cast<int64_t>((off + cast<int64_t>(18ULL))));
int64_t optOff = cast<int64_t>((off + cast<int64_t>(20ULL)));
if ((optSize >= cast<int64_t>(28ULL))) {
{
uint16_t m = u16(optOff);
if ((m != cast<uint16_t>(267ULL))) {
return {{},go_fmt_Errorf(std::string("go32: aouthdr magic %#04x, want %#04x (ZMAGIC)",46),m,cast<int64_t>(267ULL))};
}
}
c->TextSize = u32(cast<int64_t>((optOff + cast<int64_t>(4ULL))));
c->DataSize = u32(cast<int64_t>((optOff + cast<int64_t>(8ULL))));
c->BSSSize = u32(cast<int64_t>((optOff + cast<int64_t>(12ULL))));
c->Entry = u32(cast<int64_t>((optOff + cast<int64_t>(16ULL))));
c->TextStart = u32(cast<int64_t>((optOff + cast<int64_t>(20ULL))));
c->DataStart = u32(cast<int64_t>((optOff + cast<int64_t>(24ULL))));
}
int64_t shOff = cast<int64_t>((optOff + optSize));
{int64_t i = cast<int64_t>(0ULL);for (;(i < c->NSections);i++){
int64_t base = cast<int64_t>((shOff + cast<int64_t>((i * cast<int64_t>(40ULL)))));
if ((cast<int64_t>((base + cast<int64_t>(40ULL))) > len(data))) {
return {{},go_fmt_Errorf(std::string("go32: section header %d at %#x runs past end of file",52),i,base)};
}
std::string name = go_strings_TrimRight(cast<std::string>(sub(data,base,cast<int64_t>((base + cast<int64_t>(8ULL))))),std::string("\000",1));
dos_COFFSection s = dos_COFFSection{name,u32(cast<int64_t>((base + cast<int64_t>(12ULL)))),u32(cast<int64_t>((base + cast<int64_t>(16ULL)))),u32(cast<int64_t>((base + cast<int64_t>(20ULL)))),u32(cast<int64_t>((base + cast<int64_t>(36ULL)))),{}};
if ((((!dos_COFFSection_IsBSS(s)) && (s.FileOff != cast<uint32_t>(0ULL))) && (s.Size != cast<uint32_t>(0ULL)))) {
int64_t start = cast<int64_t>((c->StubEnd + cast<int64_t>(s.FileOff)));
int64_t end = cast<int64_t>((start + cast<int64_t>(s.Size)));
if (((start < cast<int64_t>(0ULL)) || (end > len(data)))) {
return {{},go_fmt_Errorf(std::string("go32: section %q bytes %#x..%#x outside file (%d bytes)",55),name,start,end,len(data))};
}
s.Data = sub(data,start,end);
}
c->Sections = append(c->Sections,Slice<dos_COFFSection>{s});
}
}if (((c->Entry == cast<uint32_t>(0ULL)) && (len(c->Sections) == cast<int64_t>(0ULL)))) {
return {{},go_errors_New(std::string("go32: no optional header and no sections \342\200\224 not a go32 COFF image",66))};
}
return {c,{}};
}
}
// tools/platform/dos/coff.go:138:1
dos_COFFSection* dos_COFF_TextSection(dos_COFF* c){
{
{auto&& tmp2 = c->Sections;
for(int64_t tmp3=0;tmp3<len(tmp2);++tmp3){
auto i=tmp3;if (dos_COFFSection_IsText(c->Sections[i])) {
return (&c->Sections[i]);
}
}}
return {};
}
}
// tools/platform/dos/coff.go:148:1
dos_COFFSection* dos_COFF_SectionAt(dos_COFF* c,uint32_t vaddr){
{
{auto&& tmp4 = c->Sections;
for(int64_t tmp5=0;tmp5<len(tmp4);++tmp5){
auto i=tmp5;dos_COFFSection* s = (&c->Sections[i]);
if (((vaddr >= s->VAddr) && (vaddr < cast<uint32_t>((s->VAddr + s->Size))))) {
return s;
}
}}
return {};
}
}
// tools/platform/dos/dos.go:87:1
std::tuple<dos_Machine*,Error> dos_LoadEXE(std::string exePath,std::string gameDir){
{
auto tmp6 = go_os_ReadFile(exePath);
Slice<uint8_t> data = std::get<0>(tmp6);
Error err = std::get<1>(tmp6);
if (bool(err)) {
return {{},err};
}
auto tmp7 = dos_ParseMZ(data);
dos_MZ* mz = std::get<0>(tmp7);
err = std::get<1>(tmp7);
if (bool(err)) {
return {{},err};
}
if (((mz->LoadImageEnd > len(data)) || (mz->LoadModuleOffset > mz->LoadImageEnd))) {
return {{},go_fmt_Errorf(std::string("mz: load image %#x..%#x outside file (%d bytes)",47),mz->LoadModuleOffset,mz->LoadImageEnd,len(data))};
}
dos_Machine* m = arenaNew(dos_Machine{Slice<uint8_t>::make(cast<int64_t>(1048576ULL)),{},gameDir,Map<uint16_t,os_File*>{},cast<uint16_t>(256ULL),cast<uint16_t>(128ULL),{},cast<uint16_t>(40960ULL),{},{},{},Map<uint32_t,dos_findState*>{},{},{},Map<uint8_t,int64_t>{},Map<uint8_t,int64_t>{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{}});
m->loadSeg = cast<uint16_t>((m->pspSeg + cast<uint16_t>(16ULL)));
Slice<uint8_t> module = sub(data,mz->LoadModuleOffset,mz->LoadImageEnd);
gcopy(sub(m->Mem,shl<uint32_t>(cast<uint32_t>(m->loadSeg),cast<int64_t>(4ULL)),len(m->Mem)),module);
{auto&& tmp8 = mz->Relocs;
for(int64_t tmp9=0;tmp9<len(tmp8);++tmp9){
auto r=tmp8[tmp9];uint32_t lin = cast<uint32_t>((shl<uint32_t>((cast<uint32_t>((cast<uint32_t>(m->loadSeg) + cast<uint32_t>(r.Segment)))),cast<int64_t>(4ULL)) + cast<uint32_t>(r.Offset)));
uint16_t w = cast<uint16_t>((le_Uint16(sub(m->Mem,cast<uint32_t>((lin & cast<uint32_t>(1048575ULL))),len(m->Mem))) + m->loadSeg));
le_PutUint16(sub(m->Mem,cast<uint32_t>((lin & cast<uint32_t>(1048575ULL))),len(m->Mem)),w);
}}
dos_Machine_setupBIOS(m);
dos_Machine_setupPSP(m);
dos_Machine_setupEnv(m,exePath);
dos_Machine_setupMCB(m);
dos_Machine_setupEMS(m);
auto tmp10 = go_os_MkdirTemp(std::string("",0),std::string("dosrun-scratch-",15));
m->scratchDir = std::get<0>(tmp10);
x86_CPU* c = x86_NewCPU(m);
c->Seg[cast<int64_t>(1ULL)] = cast<uint16_t>((m->loadSeg + mz->InitCS));
c->IP = cast<uint32_t>(mz->InitIP);
c->Seg[cast<int64_t>(2ULL)] = cast<uint16_t>((m->loadSeg + mz->InitSS));
x86_CPU_SetReg16(c,cast<int64_t>(4ULL),mz->InitSP);
c->Seg[cast<int64_t>(3ULL)] = m->pspSeg;
c->Seg[cast<int64_t>(0ULL)] = m->pspSeg;
auto tmp11 = std::make_tuple(m->pspSeg,cast<uint16_t>(128ULL));
m->dtaSeg = std::get<0>(tmp11);
m->dtaOff = std::get<1>(tmp11);
c->IF = true;
c->IntHook = [=](auto...args){return dos_Machine_handleInt(m,args...);};
m->io = arenaNew(dos_ioState{});
m->vga = arenaNew(dos_vgaState{});
c->PortIn = [=](auto...args){return dos_Machine_portIn(m,args...);};
c->PortOut = [=](auto...args){return dos_Machine_portOut(m,args...);};
c->OnStep = [=](auto...args){return dos_Machine_onStep(m,args...);};
m->CPU = c;
return {m,{}};
}
}
// tools/platform/dos/dos.go:154:1
uint8_t dos_Machine_Read(dos_Machine* m,uint32_t a){
{
a &= cast<uint32_t>(1048575ULL);
if ((((((m->RdProfileAt > cast<uint64_t>(0ULL)) && bool(m->CPU)) && (m->CPU->Steps >= m->RdProfileAt)) && (a >= m->RdLo)) && (a < m->RdHi))) {
if ((!m->rdProfile)) {
m->rdProfile = Map<uint32_t,int64_t>{};
}
m->rdProfile[cast<uint32_t>((shl<uint32_t>(cast<uint32_t>(m->CPU->Seg[cast<int64_t>(1ULL)]),cast<int64_t>(16ULL)) | (cast<uint32_t>((m->CPU->IP & cast<uint32_t>(65535ULL))))))]++;
}
if (((a >= cast<uint32_t>(655360ULL)) && (a < cast<uint32_t>(720896ULL)))) {
return dos_Machine_vgaRead(m,a);
}
return m->Mem[a];
}
}
// tools/platform/dos/dos.go:172:1
Slice<dos_VGAWriter> dos_Machine_ReadProfile(dos_Machine* m){
{
Slice<dos_VGAWriter> out = Slice<dos_VGAWriter>::make(cast<int64_t>(0ULL),len(m->rdProfile));
{auto&& tmp12 = m->rdProfile;
for(auto [tmp13,tmp14]:tmp12){
auto pc=tmp13;auto n=tmp14;out = append(out,Slice<dos_VGAWriter>{dos_VGAWriter{cast<uint16_t>(shr<uint32_t>(pc,cast<int64_t>(16ULL))),cast<uint16_t>(pc),n}});
}}
go_sort_Slice(out,[&](int64_t i,int64_t j)->bool{
return (out[i].Count > out[j].Count);
}
);
return out;
}
}
// tools/platform/dos/dos.go:180:1
void dos_Machine_Write(dos_Machine* m,uint32_t a,uint8_t v){
{
a &= cast<uint32_t>(1048575ULL);
if (((((m->WatchLen > cast<uint32_t>(0ULL)) && (a >= m->WatchAddr)) && (a < cast<uint32_t>((m->WatchAddr + m->WatchLen)))) && bool(m->CPU))) {
m->watchHits++;
if ((m->watchHits <= cast<int64_t>(60ULL))) {
dos_Machine_logf(m,std::string("WATCH: write $%02X to %05X at %04X:%04X",39),v,a,m->CPU->Seg[cast<int64_t>(1ULL)],m->CPU->IP);
}
}
if ((((m->VGAProfileAt > cast<uint64_t>(0ULL)) && bool(m->CPU)) && (m->CPU->Steps >= m->VGAProfileAt))) {
auto tmp15 = std::make_tuple(m->ProfLo,m->ProfHi);
uint32_t lo = std::get<0>(tmp15);
uint32_t hi = std::get<1>(tmp15);
if ((hi == cast<uint32_t>(0ULL))) {
auto tmp16 = std::make_tuple(cast<uint32_t>(655360ULL),cast<uint32_t>(720896ULL));
lo = std::get<0>(tmp16);
hi = std::get<1>(tmp16);
}
if (((a >= lo) && (a < hi))) {
if ((!m->vgaProfile)) {
m->vgaProfile = Map<uint32_t,int64_t>{};
}
m->vgaProfile[cast<uint32_t>((shl<uint32_t>(cast<uint32_t>(m->CPU->Seg[cast<int64_t>(1ULL)]),cast<int64_t>(16ULL)) | (cast<uint32_t>((m->CPU->IP & cast<uint32_t>(65535ULL))))))]++;
}
}
if (((a >= cast<uint32_t>(655360ULL)) && (a < cast<uint32_t>(720896ULL)))) {
dos_Machine_vgaWrite(m,a,v);
return ;
}
m->Mem[a] = v;
}
}
// tools/platform/dos/dos.go:211:1
Slice<dos_VGAWriter> dos_Machine_VGAProfile(dos_Machine* m){
{
Slice<dos_VGAWriter> out = Slice<dos_VGAWriter>::make(cast<int64_t>(0ULL),len(m->vgaProfile));
{auto&& tmp17 = m->vgaProfile;
for(auto [tmp18,tmp19]:tmp17){
auto pc=tmp18;auto n=tmp19;out = append(out,Slice<dos_VGAWriter>{dos_VGAWriter{cast<uint16_t>(shr<uint32_t>(pc,cast<int64_t>(16ULL))),cast<uint16_t>(pc),n}});
}}
go_sort_Slice(out,[&](int64_t i,int64_t j)->bool{
return (out[i].Count > out[j].Count);
}
);
return out;
}
}
// tools/platform/dos/dos.go:226:1
uint16_t dos_Machine_r16(dos_Machine* m,uint32_t lin){
{
return le_Uint16(sub(m->Mem,cast<uint32_t>((lin & cast<uint32_t>(1048575ULL))),len(m->Mem)));
}
}
// tools/platform/dos/dos.go:227:1
void dos_Machine_w16(dos_Machine* m,uint32_t lin,uint16_t v){
{
le_PutUint16(sub(m->Mem,cast<uint32_t>((lin & cast<uint32_t>(1048575ULL))),len(m->Mem)),v);
}
}
// tools/platform/dos/dos.go:230:1
uint32_t dos_lin(uint16_t seg,uint16_t off){
{
return cast<uint32_t>(((cast<uint32_t>(((shl<uint32_t>(cast<uint32_t>(seg),cast<int64_t>(4ULL))) + cast<uint32_t>(off)))) & cast<uint32_t>(1048575ULL)));
}
}
// tools/platform/dos/dos.go:238:1
void dos_Machine_setupBIOS(dos_Machine* m){
{
constexpr int64_t biosSeg=61440ULL;
constexpr int64_t biosOff=65363ULL;
m->Mem[cast<uint32_t>(1048403ULL)] = cast<uint8_t>(207ULL);
{int64_t n = cast<int64_t>(0ULL);for (;(n < cast<int64_t>(256ULL));n++){
dos_Machine_w16(m,cast<uint32_t>((cast<uint32_t>(n) * cast<uint32_t>(4ULL))),cast<uint16_t>(65363ULL));
dos_Machine_w16(m,cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(n) * cast<uint32_t>(4ULL))) + cast<uint32_t>(2ULL))),cast<uint16_t>(61440ULL));
}
}uint32_t bda = cast<uint32_t>(1024ULL);
dos_Machine_w16(m,cast<uint32_t>((bda + cast<uint32_t>(16ULL))),cast<uint16_t>(33ULL));
dos_Machine_w16(m,cast<uint32_t>((bda + cast<uint32_t>(19ULL))),cast<uint16_t>(640ULL));
m->Mem[cast<uint32_t>((bda + cast<uint32_t>(73ULL)))] = cast<uint8_t>(3ULL);
dos_Machine_w16(m,cast<uint32_t>((bda + cast<uint32_t>(74ULL))),cast<uint16_t>(80ULL));
dos_Machine_w16(m,cast<uint32_t>((bda + cast<uint32_t>(99ULL))),cast<uint16_t>(980ULL));
m->Mem[cast<uint32_t>((bda + cast<uint32_t>(132ULL)))] = cast<uint8_t>(24ULL);
m->Mem[cast<uint32_t>((bda + cast<uint32_t>(133ULL)))] = cast<uint8_t>(16ULL);
}
}
// tools/platform/dos/dos.go:258:1
void dos_Machine_setupPSP(dos_Machine* m){
{
uint32_t base = shl<uint32_t>(cast<uint32_t>(m->pspSeg),cast<int64_t>(4ULL));
auto tmp20 = std::make_tuple(cast<uint8_t>(205ULL),cast<uint8_t>(32ULL));
m->Mem[cast<uint32_t>((base + cast<uint32_t>(0ULL)))] = std::get<0>(tmp20);
m->Mem[cast<uint32_t>((base + cast<uint32_t>(1ULL)))] = std::get<1>(tmp20);
dos_Machine_w16(m,cast<uint32_t>((base + cast<uint32_t>(2ULL))),m->memTop);
dos_Machine_w16(m,cast<uint32_t>((base + cast<uint32_t>(44ULL))),m->envSeg);
m->Mem[cast<uint32_t>((base + cast<uint32_t>(128ULL)))] = cast<uint8_t>(0ULL);
m->Mem[cast<uint32_t>((base + cast<uint32_t>(129ULL)))] = cast<uint8_t>(13ULL);
}
}
// tools/platform/dos/dos.go:272:1
void dos_Machine_setupEnv(dos_Machine* m,std::string exePath){
{
uint32_t base = shl<uint32_t>(cast<uint32_t>(m->envSeg),cast<int64_t>(4ULL));
Slice<uint8_t> b={};
b = append(b,cast<Slice<uint8_t>>(std::string("PATH=C:\\\000",9)));
b = append(b,cast<Slice<uint8_t>>(std::string("COMSPEC=C:\\COMMAND.COM\000",23)));
b = append(b,Slice<uint8_t>{cast<uint8_t>(0ULL)});
b = append(b,Slice<uint8_t>{cast<uint8_t>(1ULL),cast<uint8_t>(0ULL)});
b = append(b,cast<Slice<uint8_t>>(((std::string("C:\\",3) + go_strings_ToUpper(go_filepath_Base(exePath))) + std::string("\000",1))));
gcopy(sub(m->Mem,base,len(m->Mem)),b);
}
}
// tools/platform/dos/dos.go:288:1
std::tuple<std::string,bool> dos_Machine_resolveFile(dos_Machine* m,std::string dosPath){
{
{
auto tmp21 = dos_Machine_walkPath(m,dosPath);
std::string host = std::get<0>(tmp21);
bool ok = std::get<1>(tmp21);
if (ok) {
return {host,true};
}
}
std::string norm = go_strings_ReplaceAll(dosPath,std::string("\\",1),std::string("/",1));
std::string base = sub(norm,cast<int64_t>((go_strings_LastIndex(norm,std::string("/",1)) + cast<int64_t>(1ULL))),len(norm));
if (((base != std::string("",0)) && (base != dosPath))) {
{
auto tmp22 = dos_Machine_walkPath(m,base);
std::string host = std::get<0>(tmp22);
bool ok = std::get<1>(tmp22);
if (ok) {
return {host,true};
}
}
}
{
std::string sp = dos_Machine_scratchPath(m,dosPath);
if ((sp != std::string("",0))) {
{
auto tmp23 = go_os_Stat(sp);
Error err = std::get<1>(tmp23);
if ((!err)) {
return {sp,true};
}
}
}
}
return {std::string("",0),false};
}
}
// tools/platform/dos/dos.go:310:1
std::string dos_Machine_scratchPath(dos_Machine* m,std::string dosPath){
{
if ((m->scratchDir == std::string("",0))) {
return std::string("",0);
}
std::string p = go_strings_ReplaceAll(dosPath,std::string("\\",1),std::string("/",1));
p = go_strings_TrimPrefix(p,std::string("./",2));
if (((len(p) >= cast<int64_t>(2ULL)) && (cast<uint8_t>(p[cast<int64_t>(1ULL)]) == cast<uint8_t>(58ULL)))) {
p = sub(p,cast<int64_t>(2ULL),len(p));
}
p = go_strings_TrimPrefix(p,std::string("/",1));
if ((p == std::string("",0))) {
return std::string("",0);
}
return go_filepath_Join(Slice<std::string>{m->scratchDir,go_strings_ToUpper(p)});
}
}
// tools/platform/dos/dos.go:330:1
Error dos_Machine_SeedSaveFile(dos_Machine* m,std::string dosPath,Slice<uint8_t> data){
{
std::string sp = dos_Machine_scratchPath(m,dosPath);
if ((sp == std::string("",0))) {
return go_fmt_Errorf(std::string("dos: no scratch path for %q",27),dosPath);
}
{
Error err = go_os_MkdirAll(go_filepath_Dir(sp),cast<fs_FileMode>(493ULL));
if (bool(err)) {
return err;
}
}
return go_os_WriteFile(sp,data,cast<fs_FileMode>(420ULL));
}
}
// tools/platform/dos/dos.go:344:1
std::tuple<std::string,bool> dos_Machine_walkPath(dos_Machine* m,std::string dosPath){
{
if ((m->scratchDir != std::string("",0))) {
{
auto tmp24 = dos_walkRoot(m->scratchDir,dosPath);
std::string host = std::get<0>(tmp24);
bool ok = std::get<1>(tmp24);
if (ok) {
return {host,true};
}
}
}
return dos_walkRoot(m->gameDir,dosPath);
}
}
// tools/platform/dos/dos.go:355:1
std::tuple<std::string,bool> dos_walkRoot(std::string root,std::string dosPath){
{
std::string p = go_strings_ReplaceAll(dosPath,std::string("\\",1),std::string("/",1));
p = go_strings_TrimPrefix(p,std::string("./",2));
if (((len(p) >= cast<int64_t>(2ULL)) && (cast<uint8_t>(p[cast<int64_t>(1ULL)]) == cast<uint8_t>(58ULL)))) {
p = sub(p,cast<int64_t>(2ULL),len(p));
}
p = go_strings_TrimPrefix(p,std::string("/",1));
std::string cur = root;
{auto&& tmp25 = go_strings_Split(p,std::string("/",1));
for(int64_t tmp26=0;tmp26<len(tmp25);++tmp26){
auto comp=tmp25[tmp26];if ((comp == std::string("",0))) {
continue;
}
auto tmp27 = go_os_ReadDir(cur);
Slice<RRFileInfo> entries = std::get<0>(tmp27);
Error err = std::get<1>(tmp27);
if (bool(err)) {
return {std::string("",0),false};
}
std::string match = std::string("",0);
{auto&& tmp28 = entries;
for(int64_t tmp29=0;tmp29<len(tmp28);++tmp29){
auto e=tmp28[tmp29];if (go_strings_EqualFold(rrFileInfo_Name(e),comp)) {
match = rrFileInfo_Name(e);
break;
}
}}
if ((match == std::string("",0))) {
return {std::string("",0),false};
}
cur = go_filepath_Join(Slice<std::string>{cur,match});
}}
return {cur,true};
}
}
// tools/platform/dos/dos.go:391:1
void dos_Machine_SeedDir(dos_Machine* m,std::string rel){
{
{
std::string sp = dos_Machine_scratchPath(m,rel);
if ((sp != std::string("",0))) {
go_os_MkdirAll(sp,cast<fs_FileMode>(493ULL));
}
}
}
}
// tools/platform/dos/dos.go:403:1
uint16_t dos_Machine_allocFH(dos_Machine* m){
{
{uint16_t h = cast<uint16_t>(5ULL);for (;;h++){
{
auto tmp30 = lookup(m->files,h);
bool inUse = std::get<1>(tmp30);
if ((!inUse)) {
return h;
}
}
}
}}
}
// tools/platform/dos/dos.go:417:1
std::string dos_Machine_asciiz(dos_Machine* m,uint16_t seg,uint16_t off){
{
strings_Builder sb={};
uint32_t a = dos_lin(seg,off);
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(128ULL));i++){
uint8_t c = m->Mem[cast<uint32_t>(((cast<uint32_t>((a + cast<uint32_t>(i)))) & cast<uint32_t>(1048575ULL)))];
if ((c == cast<uint8_t>(0ULL))) {
break;
}
strings_Builder_WriteByte(&(sb),c);
}
}return strings_Builder_String(&(sb));
}
}
// tools/platform/dos/dos.go:431:1
Slice<std::string> dos_Machine_IntSummary(dos_Machine* m){
{
Slice<dos_dos_432_kv> xs={};
{auto&& tmp31 = m->IntCounts;
for(auto [tmp32,tmp33]:tmp31){
auto ah=tmp32;auto n=tmp33;xs = append(xs,Slice<dos_dos_432_kv>{dos_dos_432_kv{ah,n}});
}}
go_sort_Slice(xs,[&](int64_t i,int64_t j)->bool{
return (xs[i].n > xs[j].n);
}
);
Slice<std::string> out={};
{auto&& tmp34 = xs;
for(int64_t tmp35=0;tmp35<len(tmp34);++tmp35){
auto x=tmp34[tmp35];out = append(out,Slice<std::string>{go_fmt_Sprintf(std::string("AH=%02X x%d",11),x.ah,x.n)});
}}
return out;
}
}
// tools/platform/dos/dos_ems.go:38:1
void dos_Machine_setupEMS(dos_Machine* m){
{
m->ems = arenaNew(dos_emsState{Slice<uint8_t>::make(cast<int64_t>(8388608ULL)),cast<int64_t>(0ULL),Map<uint16_t,dos_emsHandle>{},cast<uint16_t>(1ULL),std::array<int64_t,4>{cast<int64_t>(-1ULL),cast<int64_t>(-1ULL),cast<int64_t>(-1ULL),cast<int64_t>(-1ULL)},Map<uint16_t,std::array<int64_t,4>>{}});
uint32_t base = cast<uint32_t>(786432ULL);
gcopy(sub(m->Mem,cast<uint32_t>((base + cast<uint32_t>(10ULL))),len(m->Mem)),cast<Slice<uint8_t>>(std::string("EMMXXXX0",8)));
dos_Machine_w16(m,cast<uint32_t>(412ULL),cast<uint16_t>(0ULL));
dos_Machine_w16(m,cast<uint32_t>(414ULL),cast<uint16_t>(49152ULL));
}
}
// tools/platform/dos/dos_ems.go:55:1
uint32_t dos_emsState_frameLin(dos_emsState* e,int64_t slot){
{
return cast<uint32_t>((cast<uint32_t>(917504ULL) + cast<uint32_t>((cast<uint32_t>(slot) * cast<uint32_t>(16384ULL)))));
}
}
// tools/platform/dos/dos_ems.go:59:1
Slice<uint8_t> dos_Machine_EmsBacking(dos_Machine* m){
{
if ((!m->ems)) {
return {};
}
return m->ems->backing;
}
}
// tools/platform/dos/dos_ems.go:70:1
int64_t dos_Machine_EmsHandleBase(dos_Machine* m,uint16_t h){
{
if ((!m->ems)) {
return cast<int64_t>(-1ULL);
}
{
auto tmp36 = lookup(m->ems->handles,h);
dos_emsHandle hd = std::get<0>(tmp36);
bool ok = std::get<1>(tmp36);
if (ok) {
return hd.base;
}
}
return cast<int64_t>(-1ULL);
}
}
// tools/platform/dos/dos_ems.go:84:1
void dos_Machine_emsFlush(dos_Machine* m,int64_t slot){
{
int64_t p = m->ems->slot[slot];
if ((p < cast<int64_t>(0ULL))) {
return ;
}
gcopy(sub(m->ems->backing,cast<int64_t>((p * cast<int64_t>(16384ULL))),len(m->ems->backing)),sub(m->Mem,dos_emsState_frameLin(m->ems,slot),cast<uint32_t>((dos_emsState_frameLin(m->ems,slot) + cast<uint32_t>(16384ULL)))));
}
}
// tools/platform/dos/dos_ems.go:93:1
void dos_Machine_emsLoad(dos_Machine* m,int64_t slot,int64_t p){
{
gcopy(sub(m->Mem,dos_emsState_frameLin(m->ems,slot),cast<uint32_t>((dos_emsState_frameLin(m->ems,slot) + cast<uint32_t>(16384ULL)))),sub(m->ems->backing,cast<int64_t>((p * cast<int64_t>(16384ULL))),cast<int64_t>(((cast<int64_t>((p + cast<int64_t>(1ULL)))) * cast<int64_t>(16384ULL)))));
m->ems->slot[slot] = p;
}
}
// tools/platform/dos/dos_ems.go:99:1
bool dos_Machine_int67(dos_Machine* m,x86_CPU* c){
{
dos_emsState* e = m->ems;
uint8_t ah = x86_CPU_Reg8(c,cast<int64_t>(4ULL));
auto ok = [&]()->void{
x86_CPU_SetReg8(c,cast<int64_t>(4ULL),cast<uint8_t>(0ULL));
}
;
auto fail = [&](uint8_t code)->void{
x86_CPU_SetReg8(c,cast<int64_t>(4ULL),code);
}
;
{
switch(ah){
case cast<uint8_t>(64ULL):{
ok();
break;}
case cast<uint8_t>(65ULL):{
x86_CPU_SetReg16(c,cast<int64_t>(3ULL),cast<uint16_t>(57344ULL));
ok();
break;}
case cast<uint8_t>(66ULL):{
x86_CPU_SetReg16(c,cast<int64_t>(3ULL),cast<uint16_t>(cast<int64_t>((cast<int64_t>(512ULL) - e->nextPage))));
x86_CPU_SetReg16(c,cast<int64_t>(2ULL),cast<uint16_t>(512ULL));
ok();
break;}
case cast<uint8_t>(67ULL):{
int64_t want = cast<int64_t>(x86_CPU_Reg16(c,cast<int64_t>(3ULL)));
if ((want == cast<int64_t>(0ULL))) {
want = cast<int64_t>(0ULL);
}
if ((cast<int64_t>((e->nextPage + want)) > cast<int64_t>(512ULL))) {
fail(cast<uint8_t>(136ULL));
return true;
}
uint16_t h = e->nextH;
e->nextH++;
e->handles[h] = dos_emsHandle{e->nextPage,want};
e->nextPage += want;
x86_CPU_SetReg16(c,cast<int64_t>(2ULL),h);
dos_Machine_logf(m,std::string("EMS alloc handle %d: %d pages (%d KiB)",38),h,want,cast<int64_t>((want * cast<int64_t>(16ULL))));
ok();
break;}
case cast<uint8_t>(68ULL):{
int64_t phys = cast<int64_t>(x86_CPU_Reg8(c,cast<int64_t>(0ULL)));
uint16_t logical = x86_CPU_Reg16(c,cast<int64_t>(3ULL));
auto tmp37 = lookup(e->handles,x86_CPU_Reg16(c,cast<int64_t>(2ULL)));
dos_emsHandle h = std::get<0>(tmp37);
bool exists = std::get<1>(tmp37);
if ((!exists)) {
fail(cast<uint8_t>(131ULL));
return true;
}
if ((phys > cast<int64_t>(3ULL))) {
fail(cast<uint8_t>(139ULL));
return true;
}
if ((logical == cast<uint16_t>(65535ULL))) {
dos_Machine_emsFlush(m,phys);
e->slot[phys] = cast<int64_t>(-1ULL);
ok();
return true;
}
if ((cast<int64_t>(logical) >= h.count)) {
fail(cast<uint8_t>(138ULL));
return true;
}
dos_Machine_emsFlush(m,phys);
dos_Machine_emsLoad(m,phys,cast<int64_t>((h.base + cast<int64_t>(logical))));
ok();
break;}
case cast<uint8_t>(69ULL):{
removeKey(e->handles,x86_CPU_Reg16(c,cast<int64_t>(2ULL)));
ok();
break;}
case cast<uint8_t>(70ULL):{
x86_CPU_SetReg8(c,cast<int64_t>(0ULL),cast<uint8_t>(64ULL));
ok();
break;}
case cast<uint8_t>(71ULL):{
e->saved[x86_CPU_Reg16(c,cast<int64_t>(2ULL))] = e->slot;
ok();
break;}
case cast<uint8_t>(72ULL):{
{
auto tmp38 = lookup(e->saved,x86_CPU_Reg16(c,cast<int64_t>(2ULL)));
std::array<int64_t,4> s = std::get<0>(tmp38);
bool exsts = std::get<1>(tmp38);
if (exsts) {
{int64_t slot = cast<int64_t>(0ULL);for (;(slot < cast<int64_t>(4ULL));slot++){
dos_Machine_emsFlush(m,slot);
}
}{int64_t slot = cast<int64_t>(0ULL);for (;(slot < cast<int64_t>(4ULL));slot++){
if ((s[slot] >= cast<int64_t>(0ULL))) {
dos_Machine_emsLoad(m,slot,s[slot]);
}
else {
e->slot[slot] = cast<int64_t>(-1ULL);
}
}
}}
}
ok();
break;}
case cast<uint8_t>(75ULL):{
x86_CPU_SetReg16(c,cast<int64_t>(3ULL),cast<uint16_t>(len(e->handles)));
ok();
break;}
case cast<uint8_t>(76ULL):{
{
auto tmp39 = lookup(e->handles,x86_CPU_Reg16(c,cast<int64_t>(2ULL)));
dos_emsHandle h = std::get<0>(tmp39);
bool exsts = std::get<1>(tmp39);
if (exsts) {
x86_CPU_SetReg16(c,cast<int64_t>(3ULL),cast<uint16_t>(h.count));
ok();
}
else {
fail(cast<uint8_t>(131ULL));
}
}
break;}
case cast<uint8_t>(77ULL):{
uint16_t di = x86_CPU_Reg16(c,cast<int64_t>(7ULL));
uint32_t dst = dos_lin(c->Seg[cast<int64_t>(0ULL)],di);
int64_t n = cast<int64_t>(0ULL);
{auto&& tmp40 = e->handles;
for(auto [tmp41,tmp42]:tmp40){
auto h=tmp41;auto hd=tmp42;dos_Machine_w16(m,cast<uint32_t>((dst + cast<uint32_t>((cast<uint32_t>(n) * cast<uint32_t>(4ULL))))),h);
dos_Machine_w16(m,cast<uint32_t>((cast<uint32_t>((dst + cast<uint32_t>((cast<uint32_t>(n) * cast<uint32_t>(4ULL))))) + cast<uint32_t>(2ULL))),cast<uint16_t>(hd.count));
n++;
}}
x86_CPU_SetReg16(c,cast<int64_t>(3ULL),cast<uint16_t>(n));
ok();
break;}
case cast<uint8_t>(78ULL):{
dos_Machine_emsPageMap(m,c);
break;}
case cast<uint8_t>(80ULL):{
auto tmp43 = lookup(e->handles,x86_CPU_Reg16(c,cast<int64_t>(2ULL)));
dos_emsHandle h = std::get<0>(tmp43);
bool exists = std::get<1>(tmp43);
if ((!exists)) {
fail(cast<uint8_t>(131ULL));
return true;
}
uint32_t src = dos_lin(c->Seg[cast<int64_t>(3ULL)],x86_CPU_Reg16(c,cast<int64_t>(6ULL)));
int64_t n = cast<int64_t>(x86_CPU_Reg16(c,cast<int64_t>(1ULL)));
{int64_t i = cast<int64_t>(0ULL);for (;(i < n);i++){
uint16_t logical = dos_Machine_r16(m,cast<uint32_t>((src + cast<uint32_t>((cast<uint32_t>(i) * cast<uint32_t>(4ULL))))));
uint16_t physRaw = dos_Machine_r16(m,cast<uint32_t>((cast<uint32_t>((src + cast<uint32_t>((cast<uint32_t>(i) * cast<uint32_t>(4ULL))))) + cast<uint32_t>(2ULL))));
int64_t phys = cast<int64_t>(physRaw);
if ((x86_CPU_Reg8(c,cast<int64_t>(0ULL)) == cast<uint8_t>(1ULL))) {
phys = divi<int64_t>(cast<int64_t>((cast<int64_t>(cast<uint16_t>((physRaw - cast<uint16_t>(57344ULL)))) * cast<int64_t>(16ULL))),cast<int64_t>(16384ULL));
}
if (((phys < cast<int64_t>(0ULL)) || (phys > cast<int64_t>(3ULL)))) {
fail(cast<uint8_t>(139ULL));
return true;
}
if ((logical == cast<uint16_t>(65535ULL))) {
dos_Machine_emsFlush(m,phys);
e->slot[phys] = cast<int64_t>(-1ULL);
continue;
}
if ((cast<int64_t>(logical) >= h.count)) {
fail(cast<uint8_t>(138ULL));
return true;
}
dos_Machine_emsFlush(m,phys);
dos_Machine_emsLoad(m,phys,cast<int64_t>((h.base + cast<int64_t>(logical))));
}
}ok();
break;}
case cast<uint8_t>(81ULL):{
uint16_t h = x86_CPU_Reg16(c,cast<int64_t>(2ULL));
int64_t want = cast<int64_t>(x86_CPU_Reg16(c,cast<int64_t>(3ULL)));
{
auto tmp44 = lookup(e->handles,h);
dos_emsHandle hd = std::get<0>(tmp44);
bool exsts = std::get<1>(tmp44);
if (exsts) {
hd.count = want;
e->handles[h] = hd;
x86_CPU_SetReg16(c,cast<int64_t>(3ULL),cast<uint16_t>(want));
ok();
}
else {
fail(cast<uint8_t>(131ULL));
}
}
break;}
case cast<uint8_t>(83ULL):{
ok();
break;}
case cast<uint8_t>(88ULL):{
if ((x86_CPU_Reg8(c,cast<int64_t>(0ULL)) == cast<uint8_t>(1ULL))) {
x86_CPU_SetReg16(c,cast<int64_t>(1ULL),cast<uint16_t>(4ULL));
}
ok();
break;}
case cast<uint8_t>(89ULL):{
x86_CPU_SetReg16(c,cast<int64_t>(3ULL),cast<uint16_t>(cast<int64_t>((cast<int64_t>(512ULL) - e->nextPage))));
x86_CPU_SetReg16(c,cast<int64_t>(2ULL),cast<uint16_t>(512ULL));
ok();
break;}
default:{
dos_Machine_logf(m,std::string("EMS: unhandled INT 67h AH=%02X AL=%02X BX=%04X DX=%04X",54),ah,x86_CPU_Reg8(c,cast<int64_t>(0ULL)),x86_CPU_Reg16(c,cast<int64_t>(3ULL)),x86_CPU_Reg16(c,cast<int64_t>(2ULL)));
ok();
break;}
}}
return true;
}
}
// tools/platform/dos/dos_ems.go:268:1
void dos_Machine_emsPageMap(dos_Machine* m,x86_CPU* c){
{
dos_emsState* e = m->ems;
{
switch(x86_CPU_Reg8(c,cast<int64_t>(0ULL))){
case cast<uint8_t>(0ULL):case cast<uint8_t>(2ULL):{
uint32_t dst = dos_lin(c->Seg[cast<int64_t>(0ULL)],x86_CPU_Reg16(c,cast<int64_t>(7ULL)));
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(4ULL));i++){
dos_Machine_emsFlush(m,i);
dos_Machine_w16(m,cast<uint32_t>((dst + cast<uint32_t>((cast<uint32_t>(i) * cast<uint32_t>(2ULL))))),cast<uint16_t>(cast<int16_t>(e->slot[i])));
}
}if ((x86_CPU_Reg8(c,cast<int64_t>(0ULL)) == cast<uint8_t>(0ULL))) {
x86_CPU_SetReg8(c,cast<int64_t>(4ULL),cast<uint8_t>(0ULL));
return ;
}
}
case cast<uint8_t>(1ULL):{
uint32_t src = dos_lin(c->Seg[cast<int64_t>(3ULL)],x86_CPU_Reg16(c,cast<int64_t>(6ULL)));
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(4ULL));i++){
dos_Machine_emsFlush(m,i);
}
}{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(4ULL));i++){
int64_t p = cast<int64_t>(cast<int16_t>(dos_Machine_r16(m,cast<uint32_t>((src + cast<uint32_t>((cast<uint32_t>(i) * cast<uint32_t>(2ULL))))))));
if ((p >= cast<int64_t>(0ULL))) {
dos_Machine_emsLoad(m,i,p);
}
else {
e->slot[i] = cast<int64_t>(-1ULL);
}
}
}x86_CPU_SetReg8(c,cast<int64_t>(4ULL),cast<uint8_t>(0ULL));
break;}
case cast<uint8_t>(3ULL):{
x86_CPU_SetReg8(c,cast<int64_t>(0ULL),cast<uint8_t>(8ULL));
x86_CPU_SetReg8(c,cast<int64_t>(4ULL),cast<uint8_t>(0ULL));
break;}
}}
}
}
// tools/platform/dos/dos_int.go:13:1
bool dos_Machine_handleInt(dos_Machine* m,x86_CPU* c,uint8_t n){
{rrprof::Scope timing(1,"DOS and BIOS services");
{
{
switch(n){
case cast<uint8_t>(32ULL):{
m->Terminated = true;
x86_CPU_Halt(c,std::string("INT 20h \342\200\224 program terminate",29));
return true;
break;}
case cast<uint8_t>(33ULL):{
return dos_Machine_int21(m,c);
break;}
case cast<uint8_t>(63ULL):{
return dos_Machine_overlayInt(m,c);
break;}
case cast<uint8_t>(103ULL):{
return dos_Machine_int67(m,c);
break;}
case cast<uint8_t>(47ULL):{
x86_CPU_SetReg8(c,cast<int64_t>(0ULL),cast<uint8_t>(0ULL));
c->CF = false;
return true;
break;}
case cast<uint8_t>(16ULL):{
return dos_Machine_int10(m,c);
break;}
case cast<uint8_t>(51ULL):{
return dos_Machine_int33(m,c);
break;}
default:{
m->otherInts[n]++;
if ((get(m->otherInts,n) == cast<int64_t>(1ULL))) {
dos_Machine_logf(m,std::string("INT %02Xh (ignored) AX=%04X at %04X:%04X",40),n,x86_CPU_Reg16(c,cast<int64_t>(0ULL)),c->Seg[cast<int64_t>(1ULL)],c->IP);
}
c->CF = false;
return true;
break;}
}}
}
}
}
// tools/platform/dos/dos_int.go:46:1
bool dos_Machine_int21(dos_Machine* m,x86_CPU* c){
{
uint8_t ah = x86_CPU_Reg8(c,cast<int64_t>(4ULL));
m->IntCounts[ah]++;
c->CF = false;
{
switch(ah){
case cast<uint8_t>(0ULL):case cast<uint8_t>(76ULL):{
m->Terminated = true;
if ((ah == cast<uint8_t>(76ULL))) {
m->ExitCode = x86_CPU_Reg8(c,cast<int64_t>(0ULL));
}
x86_CPU_Halt(c,std::string("INT 21h/%02X \342\200\224 program exit (code %d)",39),ah,m->ExitCode);
return true;
break;}
case cast<uint8_t>(2ULL):{
return true;
break;}
case cast<uint8_t>(6ULL):case cast<uint8_t>(7ULL):case cast<uint8_t>(8ULL):{
x86_CPU_SetReg8(c,cast<int64_t>(0ULL),cast<uint8_t>(0ULL));
return true;
break;}
case cast<uint8_t>(9ULL):{
dos_Machine_logf(m,std::string("DOS print: %q",13),dos_Machine_dollarStr(m,c->Seg[cast<int64_t>(3ULL)],x86_CPU_Reg16(c,cast<int64_t>(2ULL))));
return true;
break;}
case cast<uint8_t>(25ULL):{
x86_CPU_SetReg8(c,cast<int64_t>(0ULL),cast<uint8_t>(2ULL));
return true;
break;}
case cast<uint8_t>(26ULL):{
auto tmp45 = std::make_tuple(c->Seg[cast<int64_t>(3ULL)],x86_CPU_Reg16(c,cast<int64_t>(2ULL)));
m->dtaSeg = std::get<0>(tmp45);
m->dtaOff = std::get<1>(tmp45);
return true;
break;}
case cast<uint8_t>(47ULL):{
c->Seg[cast<int64_t>(0ULL)] = m->dtaSeg;
x86_CPU_SetReg16(c,cast<int64_t>(3ULL),m->dtaOff);
return true;
break;}
case cast<uint8_t>(78ULL):{
return dos_Machine_findFirst(m,c);
break;}
case cast<uint8_t>(79ULL):{
return dos_Machine_findNext(m,c);
break;}
case cast<uint8_t>(37ULL):{
uint32_t v = cast<uint32_t>((cast<uint32_t>(x86_CPU_Reg8(c,cast<int64_t>(0ULL))) * cast<uint32_t>(4ULL)));
dos_Machine_w16(m,v,x86_CPU_Reg16(c,cast<int64_t>(2ULL)));
dos_Machine_w16(m,cast<uint32_t>((v + cast<uint32_t>(2ULL))),c->Seg[cast<int64_t>(3ULL)]);
return true;
break;}
case cast<uint8_t>(42ULL):{
x86_CPU_SetReg16(c,cast<int64_t>(1ULL),cast<uint16_t>(1992ULL));
x86_CPU_SetReg8(c,cast<int64_t>(6ULL),cast<uint8_t>(3ULL));
x86_CPU_SetReg8(c,cast<int64_t>(2ULL),cast<uint8_t>(11ULL));
x86_CPU_SetReg8(c,cast<int64_t>(0ULL),cast<uint8_t>(3ULL));
return true;
break;}
case cast<uint8_t>(44ULL):{
x86_CPU_SetReg16(c,cast<int64_t>(1ULL),cast<uint16_t>(0ULL));
x86_CPU_SetReg16(c,cast<int64_t>(2ULL),cast<uint16_t>(0ULL));
return true;
break;}
case cast<uint8_t>(48ULL):{
x86_CPU_SetReg8(c,cast<int64_t>(0ULL),cast<uint8_t>(5ULL));
x86_CPU_SetReg8(c,cast<int64_t>(4ULL),cast<uint8_t>(0ULL));
x86_CPU_SetReg16(c,cast<int64_t>(3ULL),cast<uint16_t>(0ULL));
x86_CPU_SetReg16(c,cast<int64_t>(1ULL),cast<uint16_t>(0ULL));
return true;
break;}
case cast<uint8_t>(51ULL):{
x86_CPU_SetReg8(c,cast<int64_t>(2ULL),cast<uint8_t>(0ULL));
return true;
break;}
case cast<uint8_t>(43ULL):{
x86_CPU_SetReg8(c,cast<int64_t>(0ULL),cast<uint8_t>(0ULL));
return true;
break;}
case cast<uint8_t>(45ULL):{
x86_CPU_SetReg8(c,cast<int64_t>(0ULL),cast<uint8_t>(0ULL));
return true;
break;}
case cast<uint8_t>(54ULL):{
x86_CPU_SetReg16(c,cast<int64_t>(0ULL),cast<uint16_t>(8ULL));
x86_CPU_SetReg16(c,cast<int64_t>(3ULL),cast<uint16_t>(32768ULL));
x86_CPU_SetReg16(c,cast<int64_t>(1ULL),cast<uint16_t>(512ULL));
x86_CPU_SetReg16(c,cast<int64_t>(2ULL),cast<uint16_t>(65535ULL));
return true;
break;}
case cast<uint8_t>(59ULL):{
return true;
break;}
case cast<uint8_t>(65ULL):{
std::string name = dos_Machine_asciiz(m,c->Seg[cast<int64_t>(3ULL)],x86_CPU_Reg16(c,cast<int64_t>(2ULL)));
{
std::string sp = dos_Machine_scratchPath(m,name);
if ((sp != std::string("",0))) {
go_os_Remove(sp);
}
}
return true;
break;}
case cast<uint8_t>(71ULL):{
m->Mem[dos_lin(c->Seg[cast<int64_t>(3ULL)],x86_CPU_Reg16(c,cast<int64_t>(6ULL)))] = cast<uint8_t>(0ULL);
return true;
break;}
case cast<uint8_t>(53ULL):{
uint32_t v = cast<uint32_t>((cast<uint32_t>(x86_CPU_Reg8(c,cast<int64_t>(0ULL))) * cast<uint32_t>(4ULL)));
x86_CPU_SetReg16(c,cast<int64_t>(3ULL),dos_Machine_r16(m,v));
c->Seg[cast<int64_t>(0ULL)] = dos_Machine_r16(m,cast<uint32_t>((v + cast<uint32_t>(2ULL))));
return true;
break;}
case cast<uint8_t>(82ULL):{
dos_Machine_w16(m,cast<uint32_t>(1310ULL),m->firstMCB);
c->Seg[cast<int64_t>(0ULL)] = cast<uint16_t>(80ULL);
x86_CPU_SetReg16(c,cast<int64_t>(3ULL),cast<uint16_t>(32ULL));
return true;
break;}
case cast<uint8_t>(80ULL):case cast<uint8_t>(81ULL):case cast<uint8_t>(98ULL):{
if ((ah == cast<uint8_t>(80ULL))) {
m->pspSeg = x86_CPU_Reg16(c,cast<int64_t>(3ULL));
}
else {
x86_CPU_SetReg16(c,cast<int64_t>(3ULL),m->pspSeg);
}
return true;
break;}
case cast<uint8_t>(60ULL):{
std::string name = dos_Machine_asciiz(m,c->Seg[cast<int64_t>(3ULL)],x86_CPU_Reg16(c,cast<int64_t>(2ULL)));
std::string sp = dos_Machine_scratchPath(m,name);
if ((sp == std::string("",0))) {
return dos_Machine_dosErr(m,c,cast<uint16_t>(3ULL));
}
go_os_MkdirAll(go_filepath_Dir(sp),cast<fs_FileMode>(493ULL));
auto tmp46 = go_os_Create(sp);
os_File* f = std::get<0>(tmp46);
Error err = std::get<1>(tmp46);
if (bool(err)) {
return dos_Machine_dosErr(m,c,cast<uint16_t>(3ULL));
}
uint16_t h = dos_Machine_allocFH(m);
m->files[h] = f;
dos_Machine_logf(m,std::string("create %q -> scratch %s (handle %d)",35),name,go_filepath_Base(sp),h);
x86_CPU_SetReg16(c,cast<int64_t>(0ULL),h);
return true;
break;}
case cast<uint8_t>(67ULL):{
std::string name = dos_Machine_asciiz(m,c->Seg[cast<int64_t>(3ULL)],x86_CPU_Reg16(c,cast<int64_t>(2ULL)));
if ((x86_CPU_Reg8(c,cast<int64_t>(0ULL)) == cast<uint8_t>(1ULL))) {
return true;
}
{
auto tmp47 = dos_Machine_resolveFile(m,name);
bool ok = std::get<1>(tmp47);
if (ok) {
x86_CPU_SetReg16(c,cast<int64_t>(1ULL),cast<uint16_t>(32ULL));
return true;
}
}
return dos_Machine_dosErr(m,c,cast<uint16_t>(2ULL));
break;}
case cast<uint8_t>(61ULL):{
std::string name = dos_Machine_asciiz(m,c->Seg[cast<int64_t>(3ULL)],x86_CPU_Reg16(c,cast<int64_t>(2ULL)));
auto tmp48 = dos_Machine_resolveFile(m,name);
std::string host = std::get<0>(tmp48);
bool ok = std::get<1>(tmp48);
if ((!ok)) {
dos_Machine_logf(m,std::string("open FAILED %q",14),name);
return dos_Machine_dosErr(m,c,cast<uint16_t>(2ULL));
}
auto tmp49 = go_os_Open(host);
os_File* f = std::get<0>(tmp49);
Error err = std::get<1>(tmp49);
if (bool(err)) {
return dos_Machine_dosErr(m,c,cast<uint16_t>(2ULL));
}
uint16_t h = dos_Machine_allocFH(m);
m->files[h] = f;
dos_Machine_logf(m,std::string("open %q -> handle %d",20),name,h);
x86_CPU_SetReg16(c,cast<int64_t>(0ULL),h);
return true;
break;}
case cast<uint8_t>(62ULL):{
uint16_t h = x86_CPU_Reg16(c,cast<int64_t>(3ULL));
{
os_File* f = get(m->files,h);
if (bool(f)) {
os_File_Close(f);
removeKey(m->files,h);
}
}
return true;
break;}
case cast<uint8_t>(63ULL):{
uint16_t h = x86_CPU_Reg16(c,cast<int64_t>(3ULL));
os_File* f = get(m->files,h);
if ((!f)) {
return dos_Machine_dosErr(m,c,cast<uint16_t>(6ULL));
}
int64_t n = cast<int64_t>(x86_CPU_Reg16(c,cast<int64_t>(1ULL)));
auto tmp50 = os_File_Seek(f,cast<int64_t>(0ULL),cast<int64_t>(1ULL));
int64_t pos = std::get<0>(tmp50);
Slice<uint8_t> buf = Slice<uint8_t>::make(n);
auto tmp51 = os_File_Read(f,buf);
int64_t got = std::get<0>(tmp51);
uint32_t dst = dos_lin(c->Seg[cast<int64_t>(3ULL)],x86_CPU_Reg16(c,cast<int64_t>(2ULL)));
{int64_t i = cast<int64_t>(0ULL);for (;(i < got);i++){
m->Mem[cast<uint32_t>(((cast<uint32_t>((dst + cast<uint32_t>(i)))) & cast<uint32_t>(1048575ULL)))] = buf[i];
}
}dos_Machine_logf(m,std::string("read handle %d: %d/%d bytes from file $%X -> %04X:%04X (lin $%X)",64),h,got,n,pos,c->Seg[cast<int64_t>(3ULL)],x86_CPU_Reg16(c,cast<int64_t>(2ULL)),dst);
x86_CPU_SetReg16(c,cast<int64_t>(0ULL),cast<uint16_t>(got));
return true;
break;}
case cast<uint8_t>(64ULL):{
uint16_t h = x86_CPU_Reg16(c,cast<int64_t>(3ULL));
int64_t n = cast<int64_t>(x86_CPU_Reg16(c,cast<int64_t>(1ULL)));
uint32_t src = dos_lin(c->Seg[cast<int64_t>(3ULL)],x86_CPU_Reg16(c,cast<int64_t>(2ULL)));
Slice<uint8_t> buf = Slice<uint8_t>::make(n);
{int64_t i = cast<int64_t>(0ULL);for (;(i < n);i++){
buf[i] = m->Mem[cast<uint32_t>(((cast<uint32_t>((src + cast<uint32_t>(i)))) & cast<uint32_t>(1048575ULL)))];
}
}if (((h == cast<uint16_t>(1ULL)) || (h == cast<uint16_t>(2ULL)))) {
dos_Machine_logf(m,std::string("DOS write(fd%d): %q",19),h,cast<std::string>(buf));
}
else {
os_File* f = get(m->files,h);
if (bool(f)) {
os_File_Write(f,buf);
}
}
x86_CPU_SetReg16(c,cast<int64_t>(0ULL),cast<uint16_t>(n));
return true;
break;}
case cast<uint8_t>(66ULL):{
uint16_t h = x86_CPU_Reg16(c,cast<int64_t>(3ULL));
os_File* f = get(m->files,h);
if ((!f)) {
return dos_Machine_dosErr(m,c,cast<uint16_t>(6ULL));
}
int64_t off = cast<int64_t>(cast<int32_t>(cast<uint32_t>((shl<uint32_t>(cast<uint32_t>(x86_CPU_Reg16(c,cast<int64_t>(1ULL))),cast<int64_t>(16ULL)) | cast<uint32_t>(x86_CPU_Reg16(c,cast<int64_t>(2ULL)))))));
int64_t whence = cast<int64_t>(x86_CPU_Reg8(c,cast<int64_t>(0ULL)));
auto tmp52 = os_File_Seek(f,off,whence);
int64_t pos = std::get<0>(tmp52);
Error err = std::get<1>(tmp52);
if (bool(err)) {
return dos_Machine_dosErr(m,c,cast<uint16_t>(25ULL));
}
dos_Machine_logf(m,std::string("seek handle %d: whence %d off %d -> $%X",39),h,whence,off,pos);
x86_CPU_SetReg16(c,cast<int64_t>(0ULL),cast<uint16_t>(pos));
x86_CPU_SetReg16(c,cast<int64_t>(2ULL),cast<uint16_t>(shr<int64_t>(pos,cast<int64_t>(16ULL))));
return true;
break;}
case cast<uint8_t>(68ULL):{
x86_CPU_SetReg16(c,cast<int64_t>(2ULL),cast<uint16_t>(0ULL));
return true;
break;}
case cast<uint8_t>(72ULL):{
auto tmp53 = dos_Machine_allocBlock(m,x86_CPU_Reg16(c,cast<int64_t>(3ULL)));
uint16_t seg = std::get<0>(tmp53);
uint16_t largest = std::get<1>(tmp53);
bool ok = std::get<2>(tmp53);
if ((!ok)) {
x86_CPU_SetReg16(c,cast<int64_t>(3ULL),largest);
return dos_Machine_dosErr(m,c,cast<uint16_t>(8ULL));
}
x86_CPU_SetReg16(c,cast<int64_t>(0ULL),seg);
return true;
break;}
case cast<uint8_t>(73ULL):{
if ((!dos_Machine_freeBlock(m,c->Seg[cast<int64_t>(0ULL)]))) {
return dos_Machine_dosErr(m,c,cast<uint16_t>(9ULL));
}
return true;
break;}
case cast<uint8_t>(74ULL):{
auto tmp54 = dos_Machine_resizeBlock(m,c->Seg[cast<int64_t>(0ULL)],x86_CPU_Reg16(c,cast<int64_t>(3ULL)));
uint16_t max = std::get<0>(tmp54);
bool ok = std::get<1>(tmp54);
if ((!ok)) {
x86_CPU_SetReg16(c,cast<int64_t>(3ULL),max);
return dos_Machine_dosErr(m,c,cast<uint16_t>(8ULL));
}
return true;
break;}
default:{
dos_Machine_logf(m,std::string("UNHANDLED INT 21h AH=%02X AL=%02X BX=%04X CX=%04X DX=%04X at %04X:%04X",70),ah,x86_CPU_Reg8(c,cast<int64_t>(0ULL)),x86_CPU_Reg16(c,cast<int64_t>(3ULL)),x86_CPU_Reg16(c,cast<int64_t>(1ULL)),x86_CPU_Reg16(c,cast<int64_t>(2ULL)),c->Seg[cast<int64_t>(1ULL)],c->IP);
return true;
break;}
}}
}
}
// tools/platform/dos/dos_int.go:287:1
bool dos_Machine_overlayInt(dos_Machine* m,x86_CPU* c){
{
if (((m->OverlayCalls < cast<int64_t>(32ULL)) || (modi<int64_t>(m->OverlayCalls,cast<int64_t>(500ULL)) == cast<int64_t>(0ULL)))) {
auto tmp55 = std::make_tuple(c->Seg[cast<int64_t>(1ULL)],cast<uint16_t>(c->IP));
uint16_t cs = std::get<0>(tmp55);
uint16_t ip = std::get<1>(tmp55);
uint32_t a = dos_lin(cs,ip);
dos_Machine_logf(m,std::string("INT 3Fh #%d: overlay %d entry $%04X (thunk %04X:%04X)",53),m->OverlayCalls,m->Mem[cast<uint32_t>(((cast<uint32_t>((a + cast<uint32_t>(2ULL)))) & cast<uint32_t>(1048575ULL)))],dos_Machine_r16(m,a),cs,ip);
}
m->OverlayCalls++;
return false;
}
}
// tools/platform/dos/dos_int.go:300:1
bool dos_Machine_findFirst(dos_Machine* m,x86_CPU* c){
{
std::string spec = dos_Machine_asciiz(m,c->Seg[cast<int64_t>(3ULL)],x86_CPU_Reg16(c,cast<int64_t>(2ULL)));
std::string norm = go_strings_ReplaceAll(spec,std::string("\\",1),std::string("/",1));
auto tmp56 = std::make_tuple(m->gameDir,norm);
std::string dir = std::get<0>(tmp56);
std::string pat = std::get<1>(tmp56);
{
int64_t i = go_strings_LastIndex(norm,std::string("/",1));
if ((i >= cast<int64_t>(0ULL))) {
pat = sub(norm,cast<int64_t>((i + cast<int64_t>(1ULL))),len(norm));
{
std::string dirPart = sub(norm,0,i);
if (((dirPart != std::string("",0)) && (dirPart != std::string(".",1)))) {
auto tmp57 = dos_Machine_walkPath(m,dirPart);
std::string host = std::get<0>(tmp57);
bool ok = std::get<1>(tmp57);
if ((!ok)) {
dos_Machine_logf(m,std::string("FindFirst %q -> PATH NOT FOUND (dir %q)",39),spec,dirPart);
return dos_Machine_dosErr(m,c,cast<uint16_t>(3ULL));
}
dir = host;
}
}
}
}
auto tmp58 = go_os_ReadDir(dir);
Slice<RRFileInfo> entries = std::get<0>(tmp58);
Error err = std::get<1>(tmp58);
if (bool(err)) {
return dos_Machine_dosErr(m,c,cast<uint16_t>(3ULL));
}
uint16_t attrMask = x86_CPU_Reg16(c,cast<int64_t>(1ULL));
dos_findState* st = arenaNew(dos_findState{});
{auto&& tmp59 = entries;
for(int64_t tmp60=0;tmp60<len(tmp59);++tmp60){
auto e=tmp59[tmp60];if (go_strings_HasPrefix(rrFileInfo_Name(e),std::string(".",1))) {
continue;
}
if ((rrFileInfo_IsDir(e) && (cast<uint16_t>((attrMask & cast<uint16_t>(16ULL))) == cast<uint16_t>(0ULL)))) {
continue;
}
if (dos_dosMatch(pat,rrFileInfo_Name(e))) {
st->matches = append(st->matches,Slice<std::string>{go_filepath_Join(Slice<std::string>{dir,rrFileInfo_Name(e)})});
}
}}
m->finds[dos_lin(m->dtaSeg,m->dtaOff)] = st;
dos_Machine_logf(m,std::string("FindFirst %q -> %d match(es)",28),spec,len(st->matches));
return dos_Machine_findNext(m,c);
}
}
// tools/platform/dos/dos_int.go:342:1
bool dos_Machine_findNext(dos_Machine* m,x86_CPU* c){
{
dos_findState* st = get(m->finds,dos_lin(m->dtaSeg,m->dtaOff));
if (((!st) || (st->idx >= len(st->matches)))) {
if ((!st)) {
dos_Machine_logf(m,std::string("FindNext at DTA %04X:%04X -> NO STATE",37),m->dtaSeg,m->dtaOff);
}
else {
dos_Machine_logf(m,std::string("FindNext at DTA %04X:%04X -> exhausted (%d/%d)",46),m->dtaSeg,m->dtaOff,st->idx,len(st->matches));
}
return dos_Machine_dosErr(m,c,cast<uint16_t>(18ULL));
}
std::string host = st->matches[st->idx];
st->idx++;
auto tmp61 = go_os_Stat(host);
RRFileInfo info = std::get<0>(tmp61);
Error err = std::get<1>(tmp61);
if (bool(err)) {
return dos_Machine_dosErr(m,c,cast<uint16_t>(18ULL));
}
uint32_t dta = dos_lin(m->dtaSeg,m->dtaOff);
uint8_t attr = cast<uint8_t>(32ULL);
if (rrFileInfo_IsDir(info)) {
attr = cast<uint8_t>(16ULL);
}
m->Mem[cast<uint32_t>((dta + cast<uint32_t>(21ULL)))] = attr;
dos_Machine_w16(m,cast<uint32_t>((dta + cast<uint32_t>(22ULL))),cast<uint16_t>(0ULL));
dos_Machine_w16(m,cast<uint32_t>((dta + cast<uint32_t>(24ULL))),cast<uint16_t>(33ULL));
uint32_t sz = cast<uint32_t>(rrFileInfo_Size(info));
dos_Machine_w16(m,cast<uint32_t>((dta + cast<uint32_t>(26ULL))),cast<uint16_t>(sz));
dos_Machine_w16(m,cast<uint32_t>((dta + cast<uint32_t>(28ULL))),cast<uint16_t>(shr<uint32_t>(sz,cast<int64_t>(16ULL))));
std::string name = go_strings_ToUpper(go_filepath_Base(host));
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(13ULL));i++){
if ((i < len(name))) {
m->Mem[cast<uint32_t>((cast<uint32_t>((dta + cast<uint32_t>(30ULL))) + cast<uint32_t>(i)))] = cast<uint8_t>(name[i]);
}
else {
m->Mem[cast<uint32_t>((cast<uint32_t>((dta + cast<uint32_t>(30ULL))) + cast<uint32_t>(i)))] = cast<uint8_t>(0ULL);
}
}
}c->CF = false;
return true;
}
}
// tools/platform/dos/dos_int.go:385:1
bool dos_dosMatch(std::string pattern,std::string name){
{
std::string p = go_strings_ToUpper(pattern);
std::string n = go_strings_ToUpper(name);
if ((((p == std::string("*.*",3)) || (p == std::string("*",1))) || (p == std::string("",0)))) {
return true;
}
return dos_globSegments(p,n);
}
}
// tools/platform/dos/dos_int.go:395:1
bool dos_globSegments(std::string p,std::string n){
{
auto tmp62 = std::make_tuple(cast<int64_t>(0ULL),cast<int64_t>(0ULL));
int64_t pi = std::get<0>(tmp62);
int64_t ni = std::get<1>(tmp62);
auto tmp63 = std::make_tuple(cast<int64_t>(-1ULL),cast<int64_t>(0ULL));
int64_t star = std::get<0>(tmp63);
int64_t mark = std::get<1>(tmp63);
{;for (;(ni < len(n));){
if (((pi < len(p)) && (((cast<uint8_t>(p[pi]) == cast<uint8_t>(63ULL)) || (cast<uint8_t>(p[pi]) == cast<uint8_t>(n[ni])))))) {
pi++;
ni++;
}
else if (((pi < len(p)) && (cast<uint8_t>(p[pi]) == cast<uint8_t>(42ULL)))) {
star = pi;
mark = ni;
pi++;
}
else if ((star != cast<int64_t>(-1ULL))) {
pi = cast<int64_t>((star + cast<int64_t>(1ULL)));
mark++;
ni = mark;
}
else {
return false;
}
}
}{;for (;((pi < len(p)) && (cast<uint8_t>(p[pi]) == cast<uint8_t>(42ULL)));){
pi++;
}
}return (pi == len(p));
}
}
// tools/platform/dos/dos_int.go:421:1
bool dos_Machine_dosErr(dos_Machine* m,x86_CPU* c,uint16_t code){
{
c->CF = true;
x86_CPU_SetReg16(c,cast<int64_t>(0ULL),code);
return true;
}
}
// tools/platform/dos/dos_int.go:428:1
std::string dos_Machine_dollarStr(dos_Machine* m,uint16_t seg,uint16_t off){
{
Slice<uint8_t> b={};
uint32_t a = dos_lin(seg,off);
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(256ULL));i++){
uint8_t ch = m->Mem[cast<uint32_t>(((cast<uint32_t>((a + cast<uint32_t>(i)))) & cast<uint32_t>(1048575ULL)))];
if ((ch == cast<uint8_t>(36ULL))) {
break;
}
b = append(b,Slice<uint8_t>{ch});
}
}return cast<std::string>(b);
}
}
// tools/platform/dos/dos_io.go:42:1
uint32_t dos_Machine_portIn(dos_Machine* m,uint16_t port,int64_t size){
{
auto tmp64 = m->io;
dos_ioState* io = tmp64;
{
switch(port){
case cast<uint16_t>(96ULL):{
io->kbdOutFull = false;
return cast<uint32_t>(io->kbdOut);
break;}
case cast<uint16_t>(100ULL):{
uint32_t s = cast<uint32_t>(0ULL);
if (io->kbdOutFull) {
s |= cast<uint32_t>(1ULL);
}
return s;
break;}
case cast<uint16_t>(986ULL):case cast<uint16_t>(954ULL):{
io->retrace = (!io->retrace);
if (io->retrace) {
return cast<uint32_t>(9ULL);
}
return cast<uint32_t>(0ULL);
break;}
case cast<uint16_t>(64ULL):case cast<uint16_t>(65ULL):case cast<uint16_t>(66ULL):{
io->pit -= cast<uint16_t>(311ULL);
return cast<uint32_t>((cast<uint32_t>(shr<uint16_t>(io->pit,cast<int64_t>(8ULL))) & cast<uint32_t>(255ULL)));
break;}
case cast<uint16_t>(552ULL):case cast<uint16_t>(904ULL):{
if (io->oplTimer) {
return cast<uint32_t>(192ULL);
}
return cast<uint32_t>(0ULL);
break;}
case cast<uint16_t>(513ULL):{
return cast<uint32_t>(240ULL);
break;}
case cast<uint16_t>(549ULL):{
return cast<uint32_t>(io->mixRegs[io->mixReg]);
break;}
case cast<uint16_t>(554ULL):{
if ((len(io->dspQueue) > cast<int64_t>(0ULL))) {
uint8_t v = io->dspQueue[cast<int64_t>(0ULL)];
io->dspQueue = sub(io->dspQueue,cast<int64_t>(1ULL),len(io->dspQueue));
return cast<uint32_t>(v);
}
return cast<uint32_t>(0ULL);
break;}
case cast<uint16_t>(556ULL):{
return cast<uint32_t>(0ULL);
break;}
case cast<uint16_t>(558ULL):{
if ((len(io->dspQueue) > cast<int64_t>(0ULL))) {
return cast<uint32_t>(128ULL);
}
return cast<uint32_t>(0ULL);
break;}
case cast<uint16_t>(32ULL):case cast<uint16_t>(33ULL):case cast<uint16_t>(160ULL):case cast<uint16_t>(161ULL):{
return cast<uint32_t>(0ULL);
break;}
default:{
if ((!io->seen)) {
io->seen = Map<uint16_t,bool>{};
}
if ((!get(io->seen,port))) {
io->seen[port] = true;
dos_Machine_logf(m,std::string("IN port $%03X (unmodeled) at %04X:%04X",38),port,m->CPU->Seg[cast<int64_t>(1ULL)],m->CPU->IP);
}
return dos_widthMask8(size);
break;}
}}
}
}
// tools/platform/dos/dos_io.go:100:1
void dos_Machine_portOut(dos_Machine* m,uint16_t port,int64_t size,uint32_t v){
{
if (((size == cast<int64_t>(2ULL)) && ((((port == cast<uint16_t>(964ULL)) || (port == cast<uint16_t>(974ULL))) || (port == cast<uint16_t>(980ULL)))))) {
dos_Machine_portOut(m,port,cast<int64_t>(1ULL),cast<uint32_t>((v & cast<uint32_t>(255ULL))));
dos_Machine_portOut(m,cast<uint16_t>((port + cast<uint16_t>(1ULL))),cast<int64_t>(1ULL),cast<uint32_t>(((shr<uint32_t>(v,cast<int64_t>(8ULL))) & cast<uint32_t>(255ULL))));
return ;
}
auto tmp65 = m->io;
dos_ioState* io = tmp65;
uint8_t b = cast<uint8_t>(v);
{
switch(port){
case cast<uint16_t>(964ULL):case cast<uint16_t>(965ULL):case cast<uint16_t>(974ULL):case cast<uint16_t>(975ULL):case cast<uint16_t>(980ULL):case cast<uint16_t>(981ULL):{
dos_Machine_vgaRegOut(m,port,b);
break;}
case cast<uint16_t>(96ULL):{
if (io->expectData) {
io->expectData = false;
return ;
}
auto tmp66 = std::make_tuple(cast<uint8_t>(250ULL),true);
io->kbdOut = std::get<0>(tmp66);
io->kbdOutFull = std::get<1>(tmp66);
break;}
case cast<uint16_t>(100ULL):{
{
switch(b){
case cast<uint8_t>(170ULL):{
auto tmp67 = std::make_tuple(cast<uint8_t>(85ULL),true);
io->kbdOut = std::get<0>(tmp67);
io->kbdOutFull = std::get<1>(tmp67);
break;}
case cast<uint8_t>(171ULL):{
auto tmp68 = std::make_tuple(cast<uint8_t>(0ULL),true);
io->kbdOut = std::get<0>(tmp68);
io->kbdOutFull = std::get<1>(tmp68);
break;}
case cast<uint8_t>(238ULL):{
auto tmp69 = std::make_tuple(cast<uint8_t>(238ULL),true);
io->kbdOut = std::get<0>(tmp69);
io->kbdOutFull = std::get<1>(tmp69);
break;}
case cast<uint8_t>(96ULL):case cast<uint8_t>(209ULL):case cast<uint8_t>(210ULL):case cast<uint8_t>(211ULL):case cast<uint8_t>(212ULL):{
io->expectData = true;
break;}
}}
break;}
case cast<uint16_t>(552ULL):case cast<uint16_t>(904ULL):{
io->oplReg = b;
break;}
case cast<uint16_t>(553ULL):case cast<uint16_t>(905ULL):{
if ((io->oplReg == cast<uint8_t>(4ULL))) {
if ((cast<uint8_t>((b & cast<uint8_t>(128ULL))) != cast<uint8_t>(0ULL))) {
io->oplTimer = false;
}
else if ((cast<uint8_t>((b & cast<uint8_t>(1ULL))) != cast<uint8_t>(0ULL))) {
io->oplTimer = true;
}
}
break;}
case cast<uint16_t>(968ULL):{
io->dacIndex = cast<int64_t>((cast<int64_t>(b) * cast<int64_t>(3ULL)));
break;}
case cast<uint16_t>(969ULL):{
io->Pal[modi<int64_t>(io->dacIndex,cast<int64_t>(768ULL))] = b;
if(rrcapture::trace.active)rrDOSWrite(&(io->Pal[modi<int64_t>(io->dacIndex,cast<int64_t>(768ULL))]));
io->dacIndex++;
break;}
case cast<uint16_t>(548ULL):{
io->mixReg = b;
break;}
case cast<uint16_t>(549ULL):{
io->mixRegs[io->mixReg] = b;
break;}
case cast<uint16_t>(550ULL):{
if (((io->dspReset == cast<uint8_t>(1ULL)) && (b == cast<uint8_t>(0ULL)))) {
io->dspQueue = append(io->dspQueue,Slice<uint8_t>{cast<uint8_t>(170ULL)});
}
io->dspReset = b;
break;}
case cast<uint16_t>(556ULL):{
{
switch(b){
case cast<uint8_t>(225ULL):{
io->dspQueue = append(io->dspQueue,Slice<uint8_t>{cast<uint8_t>(3ULL),cast<uint8_t>(1ULL)});
break;}
}}
break;}
}}
}
}
// tools/platform/dos/dos_io.go:163:1
uint32_t dos_widthMask8(int64_t size){
{
if ((size == cast<int64_t>(2ULL))) {
return cast<uint32_t>(65535ULL);
}
return cast<uint32_t>(255ULL);
}
}
// tools/platform/dos/dos_io.go:176:1
void dos_Machine_onStep(dos_Machine* m,x86_CPU* c){
{
if ((!m->EnableIRQ)) {
return ;
}
auto tmp70 = m->io;
dos_ioState* io = tmp70;
if (((m->keyRetry && (len(m->keyEvents) > cast<int64_t>(0ULL))) && (m->keyEvents[cast<int64_t>(0ULL)].kind == cast<dos_injKind>(1ULL)))) {
dos_injEvent ev = m->keyEvents[cast<int64_t>(0ULL)];
auto tmp71 = std::make_tuple(ev.code,true);
m->io->kbdOut = std::get<0>(tmp71);
m->io->kbdOutFull = std::get<1>(tmp71);
if (x86_CPU_Interrupt(c,cast<uint8_t>(9ULL))) {
m->keyRetry = false;
dos_Machine_logInject(m,std::string("KEY scancode %02X (retry)",25),ev.code);
m->keyEvents = sub(m->keyEvents,cast<int64_t>(1ULL),len(m->keyEvents));
m->keyWait = ev.delay;
}
else {
m->io->kbdOutFull = false;
}
}
io->tick++;
if ((io->tick == cast<uint32_t>(400ULL))) {
dos_Machine_pumpKeys(m);
return ;
}
if ((io->tick < cast<uint32_t>(800ULL))) {
return ;
}
io->tick = cast<uint32_t>(0ULL);
uint32_t t = cast<uint32_t>((cast<uint32_t>(dos_Machine_r16(m,cast<uint32_t>(1132ULL))) | shl<uint32_t>(cast<uint32_t>(dos_Machine_r16(m,cast<uint32_t>(1134ULL))),cast<int64_t>(16ULL))));
t++;
dos_Machine_w16(m,cast<uint32_t>(1132ULL),cast<uint16_t>(t));
dos_Machine_w16(m,cast<uint32_t>(1134ULL),cast<uint16_t>(shr<uint32_t>(t,cast<int64_t>(16ULL))));
if ((dos_Machine_r16(m,cast<uint32_t>(34ULL)) != cast<uint16_t>(0ULL))) {
x86_CPU_Interrupt(c,cast<uint8_t>(8ULL));
}
}
}
// tools/platform/dos/dos_keyboard.go:70:1
std::tuple<uint8_t,bool> dos_Scancode(std::string token){
{
auto tmp72 = lookup(dos_makeScancodes,token);
uint8_t sc = std::get<0>(tmp72);
bool ok = std::get<1>(tmp72);
return {sc,ok};
}
}
// tools/platform/dos/dos_keyboard.go:92:1
std::tuple<Slice<dos_injEvent>,Error> dos_ParseKeys(std::string spec){
{
constexpr int64_t makeBreakGap=2ULL;
constexpr int64_t interKeyGap=10ULL;
constexpr int64_t clickGap=4ULL;
Slice<std::string> toks = dos_splitTokens(spec);
Slice<dos_injEvent> out={};
{int64_t i = cast<int64_t>(0ULL);for (;(i < len(toks));i++){
std::string tok = go_strings_ToLower(toks[i]);
{
if ((tok == std::string("",0))){
continue;
}
else if (go_strings_HasPrefix(tok,std::string("wait:",5))){
auto tmp74 = go_strconv_Atoi(sub(tok,cast<int64_t>(5ULL),len(tok)));
int64_t n = std::get<0>(tmp74);
Error err = std::get<1>(tmp74);
if (bool(err)) {
return {{},go_fmt_Errorf(std::string("keys: bad wait %q",17),tok)};
}
out = append(out,Slice<dos_injEvent>{dos_injEvent{cast<dos_injKind>(0ULL),{},{},{},n}});
}
else if (go_strings_HasPrefix(tok,std::string("at:",3))){
if ((cast<int64_t>((i + cast<int64_t>(1ULL))) >= len(toks))) {
return {{},go_fmt_Errorf(std::string("keys: %q needs X,Y",18),tok)};
}
auto tmp75 = go_strconv_Atoi(sub(tok,cast<int64_t>(3ULL),len(tok)));
int64_t x = std::get<0>(tmp75);
Error err1 = std::get<1>(tmp75);
auto tmp76 = go_strconv_Atoi(go_strings_TrimSpace(toks[cast<int64_t>((i + cast<int64_t>(1ULL)))]));
int64_t y = std::get<0>(tmp76);
Error err2 = std::get<1>(tmp76);
if ((bool(err1) || bool(err2))) {
return {{},go_fmt_Errorf(std::string("keys: bad coords %q,%q",22),tok,toks[cast<int64_t>((i + cast<int64_t>(1ULL)))])};
}
i++;
out = append(out,Slice<dos_injEvent>{dos_injEvent{cast<dos_injKind>(2ULL),{},{},{},cast<int64_t>(3ULL)},dos_injEvent{cast<dos_injKind>(3ULL),{},x,y,cast<int64_t>(2ULL)}});
}
else if (go_strings_HasPrefix(tok,std::string("dxy:",4))){
if ((cast<int64_t>((i + cast<int64_t>(1ULL))) >= len(toks))) {
return {{},go_fmt_Errorf(std::string("keys: %q needs DX,DY",20),tok)};
}
auto tmp77 = go_strconv_Atoi(sub(tok,cast<int64_t>(4ULL),len(tok)));
int64_t dx = std::get<0>(tmp77);
Error err1 = std::get<1>(tmp77);
auto tmp78 = go_strconv_Atoi(go_strings_TrimSpace(toks[cast<int64_t>((i + cast<int64_t>(1ULL)))]));
int64_t dy = std::get<0>(tmp78);
Error err2 = std::get<1>(tmp78);
if ((bool(err1) || bool(err2))) {
return {{},go_fmt_Errorf(std::string("keys: bad delta %q,%q",21),tok,toks[cast<int64_t>((i + cast<int64_t>(1ULL)))])};
}
i++;
out = append(out,Slice<dos_injEvent>{dos_injEvent{cast<dos_injKind>(5ULL),{},dx,dy,cast<int64_t>(2ULL)}});
}
else if ((tok == std::string("lclick",6)) || (tok == std::string("rclick",6))){
uint8_t mask = cast<uint8_t>(1ULL);
if ((tok == std::string("rclick",6))) {
mask = cast<uint8_t>(2ULL);
}
out = append(out,Slice<dos_injEvent>{dos_injEvent{cast<dos_injKind>(4ULL),mask,{},{},cast<int64_t>(4ULL)},dos_injEvent{cast<dos_injKind>(4ULL),cast<uint8_t>(0ULL),{},{},cast<int64_t>(10ULL)}});
}
else if ((tok == std::string("ldown",5))){
out = append(out,Slice<dos_injEvent>{dos_injEvent{cast<dos_injKind>(4ULL),cast<uint8_t>(1ULL),{},{},cast<int64_t>(2ULL)}});
}
else if ((tok == std::string("rdown",5))){
out = append(out,Slice<dos_injEvent>{dos_injEvent{cast<dos_injKind>(4ULL),cast<uint8_t>(2ULL),{},{},cast<int64_t>(2ULL)}});
}
else if ((tok == std::string("lup",3)) || (tok == std::string("rup",3))){
out = append(out,Slice<dos_injEvent>{dos_injEvent{cast<dos_injKind>(4ULL),cast<uint8_t>(0ULL),{},{},cast<int64_t>(10ULL)}});
}
else if (go_strings_HasPrefix(tok,std::string("kdown:",6)) || go_strings_HasPrefix(tok,std::string("kup:",4))){
std::string name = sub(tok,cast<int64_t>((go_strings_IndexByte(tok,cast<uint8_t>(58ULL)) + cast<int64_t>(1ULL))),len(tok));
auto tmp79 = lookup(dos_makeScancodes,name);
uint8_t sc = std::get<0>(tmp79);
bool ok = std::get<1>(tmp79);
if ((!ok)) {
return {{},go_fmt_Errorf(std::string("keys: unknown key %q in %q",26),name,tok)};
}
if (go_strings_HasPrefix(tok,std::string("kup:",4))) {
sc |= cast<uint8_t>(128ULL);
}
out = append(out,Slice<dos_injEvent>{dos_injEvent{cast<dos_injKind>(1ULL),sc,{},{},cast<int64_t>(2ULL)}});
}
else {
auto tmp80 = lookup(dos_makeScancodes,tok);
uint8_t sc = std::get<0>(tmp80);
bool ok = std::get<1>(tmp80);
if ((!ok)) {
return {{},go_fmt_Errorf(std::string("keys: unknown token %q",22),tok)};
}
out = append(out,Slice<dos_injEvent>{dos_injEvent{cast<dos_injKind>(1ULL),sc,{},{},cast<int64_t>(2ULL)},dos_injEvent{cast<dos_injKind>(1ULL),cast<uint8_t>((sc | cast<uint8_t>(128ULL))),{},{},cast<int64_t>(10ULL)}});
}
}
tmp73:;
}
}return {out,{}};
}
}
// tools/platform/dos/dos_keyboard.go:182:1
Slice<std::string> dos_splitTokens(std::string spec){
{
Slice<std::string> parts = go_strings_Split(spec,std::string(",",1));
{auto&& tmp81 = parts;
for(int64_t tmp82=0;tmp82<len(tmp81);++tmp82){
auto i=tmp82;parts[i] = go_strings_TrimSpace(parts[i]);
}}
return parts;
}
}
// tools/platform/dos/dos_keyboard.go:191:1
void dos_Machine_SetKeys(dos_Machine* m,Slice<dos_injEvent> events){
{
m->keyEvents = events;
}
}
// tools/platform/dos/dos_keyboard.go:194:1
bool dos_Machine_KeysPending(dos_Machine* m){
{
return (len(m->keyEvents) > cast<int64_t>(0ULL));
}
}
// tools/platform/dos/dos_keyboard.go:201:1
void dos_Machine_pumpKeys(dos_Machine* m){
{
if ((len(m->keyEvents) == cast<int64_t>(0ULL))) {
return ;
}
if ((m->keyWait > cast<int64_t>(0ULL))) {
m->keyWait--;
return ;
}
dos_injEvent ev = m->keyEvents[cast<int64_t>(0ULL)];
{
switch(ev.kind){
case cast<dos_injKind>(1ULL):{
auto tmp83 = std::make_tuple(ev.code,true);
m->io->kbdOut = std::get<0>(tmp83);
m->io->kbdOutFull = std::get<1>(tmp83);
if ((!x86_CPU_Interrupt(m->CPU,cast<uint8_t>(9ULL)))) {
m->io->kbdOutFull = false;
m->keyRetry = true;
return ;
}
m->keyRetry = false;
dos_Machine_logInject(m,std::string("KEY scancode %02X (IVT9=%04X:%04X)",34),ev.code,dos_Machine_r16(m,cast<uint32_t>(38ULL)),dos_Machine_r16(m,cast<uint32_t>(36ULL)));
break;}
case cast<dos_injKind>(2ULL):{
dos_Machine_HomeMouse(m);
dos_Machine_logInject(m,std::string("MOUSE home -> 0,0",17));
break;}
case cast<dos_injKind>(3ULL):{
dos_Machine_MoveMouseTo(m,ev.x,ev.y);
dos_Machine_logInject(m,std::string("MOUSE move -> %d,%d",19),ev.x,ev.y);
break;}
case cast<dos_injKind>(4ULL):{
dos_Machine_SetMouseButtons(m,ev.code);
dos_Machine_logInject(m,std::string("MOUSE buttons=%02X at %d,%d",27),ev.code,m->ms->x,m->ms->y);
break;}
case cast<dos_injKind>(5ULL):{
dos_Machine_FeedMickeys(m,ev.x,ev.y);
dos_Machine_logInject(m,std::string("MOUSE mickeys += %d,%d",22),ev.x,ev.y);
break;}
}}
m->keyEvents = sub(m->keyEvents,cast<int64_t>(1ULL),len(m->keyEvents));
m->keyWait = ev.delay;
}
}
// tools/platform/dos/dos_mem.go:29:1
void dos_Machine_setupMCB(dos_Machine* m){
{
uint16_t envMCB = cast<uint16_t>((m->envSeg - cast<uint16_t>(1ULL)));
uint16_t pspMCB = cast<uint16_t>((m->pspSeg - cast<uint16_t>(1ULL)));
dos_Machine_writeMCB(m,envMCB,cast<uint8_t>(77ULL),m->pspSeg,cast<uint16_t>((pspMCB - m->envSeg)));
dos_Machine_writeMCB(m,pspMCB,cast<uint8_t>(90ULL),m->pspSeg,cast<uint16_t>((m->memTop - m->pspSeg)));
m->firstMCB = envMCB;
}
}
// tools/platform/dos/dos_mem.go:37:1
void dos_Machine_writeMCB(dos_Machine* m,uint16_t mcbSeg,uint8_t mark,uint16_t owner,uint16_t size){
{
uint32_t a = shl<uint32_t>(cast<uint32_t>(mcbSeg),cast<int64_t>(4ULL));
m->Mem[cast<uint32_t>((a & cast<uint32_t>(1048575ULL)))] = mark;
dos_Machine_w16(m,cast<uint32_t>((a + cast<uint32_t>(1ULL))),owner);
dos_Machine_w16(m,cast<uint32_t>((a + cast<uint32_t>(3ULL))),size);
}
}
// tools/platform/dos/dos_mem.go:44:1
std::tuple<uint8_t,uint16_t,uint16_t> dos_Machine_readMCB(dos_Machine* m,uint16_t mcbSeg){
uint8_t mark{};
uint16_t owner{};
uint16_t size{};
{
uint32_t a = shl<uint32_t>(cast<uint32_t>(mcbSeg),cast<int64_t>(4ULL));
return {m->Mem[cast<uint32_t>((a & cast<uint32_t>(1048575ULL)))],dos_Machine_r16(m,cast<uint32_t>((a + cast<uint32_t>(1ULL)))),dos_Machine_r16(m,cast<uint32_t>((a + cast<uint32_t>(3ULL))))};
}
}
// tools/platform/dos/dos_mem.go:52:1
std::tuple<uint16_t,uint16_t,bool> dos_Machine_allocBlock(dos_Machine* m,uint16_t want){
uint16_t seg{};
uint16_t largest{};
bool ok{};
{
uint16_t mcb = m->firstMCB;
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(4096ULL));i++){
auto tmp84 = dos_Machine_readMCB(m,mcb);
uint8_t mark = std::get<0>(tmp84);
uint16_t owner = std::get<1>(tmp84);
uint16_t size = std::get<2>(tmp84);
if ((owner == cast<uint16_t>(0ULL))) {
if ((size >= want)) {
dos_Machine_carve(m,mcb,mark,size,want,m->pspSeg);
return {cast<uint16_t>((mcb + cast<uint16_t>(1ULL))),cast<uint16_t>(0ULL),true};
}
if ((size > largest)) {
largest = size;
}
}
if ((mark == cast<uint8_t>(90ULL))) {
break;
}
mcb += cast<uint16_t>((cast<uint16_t>(1ULL) + size));
}
}return {cast<uint16_t>(0ULL),largest,false};
}
}
// tools/platform/dos/dos_mem.go:75:1
void dos_Machine_carve(dos_Machine* m,uint16_t mcb,uint8_t mark,uint16_t size,uint16_t want,uint16_t owner){
{
if ((size >= cast<uint16_t>((want + cast<uint16_t>(1ULL))))) {
uint16_t rem = cast<uint16_t>((cast<uint16_t>((mcb + cast<uint16_t>(1ULL))) + want));
dos_Machine_writeMCB(m,rem,mark,cast<uint16_t>(0ULL),cast<uint16_t>((cast<uint16_t>((size - want)) - cast<uint16_t>(1ULL))));
dos_Machine_writeMCB(m,mcb,cast<uint8_t>(77ULL),owner,want);
}
else {
dos_Machine_writeMCB(m,mcb,mark,owner,size);
}
}
}
// tools/platform/dos/dos_mem.go:87:1
bool dos_Machine_freeBlock(dos_Machine* m,uint16_t blockSeg){
{
if (((blockSeg == cast<uint16_t>(0ULL)) || (blockSeg <= m->firstMCB))) {
return false;
}
uint16_t mcb = cast<uint16_t>((blockSeg - cast<uint16_t>(1ULL)));
auto tmp85 = dos_Machine_readMCB(m,mcb);
uint8_t mark = std::get<0>(tmp85);
uint16_t size = std::get<2>(tmp85);
if (((mark != cast<uint8_t>(77ULL)) && (mark != cast<uint8_t>(90ULL)))) {
return false;
}
dos_Machine_writeMCB(m,mcb,mark,cast<uint16_t>(0ULL),size);
dos_Machine_coalesce(m);
return true;
}
}
// tools/platform/dos/dos_mem.go:102:1
void dos_Machine_coalesce(dos_Machine* m){
{
uint16_t mcb = m->firstMCB;
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(4096ULL));i++){
auto tmp86 = dos_Machine_readMCB(m,mcb);
uint8_t mark = std::get<0>(tmp86);
uint16_t owner = std::get<1>(tmp86);
uint16_t size = std::get<2>(tmp86);
if ((mark == cast<uint8_t>(90ULL))) {
break;
}
uint16_t next = cast<uint16_t>((cast<uint16_t>((mcb + cast<uint16_t>(1ULL))) + size));
auto tmp87 = dos_Machine_readMCB(m,next);
uint8_t nmark = std::get<0>(tmp87);
uint16_t nowner = std::get<1>(tmp87);
uint16_t nsize = std::get<2>(tmp87);
if (((owner == cast<uint16_t>(0ULL)) && (nowner == cast<uint16_t>(0ULL)))) {
dos_Machine_writeMCB(m,mcb,nmark,cast<uint16_t>(0ULL),cast<uint16_t>((cast<uint16_t>((size + nsize)) + cast<uint16_t>(1ULL))));
continue;
}
mcb = next;
}
}}
}
// tools/platform/dos/dos_mem.go:122:1
std::tuple<uint16_t,bool> dos_Machine_resizeBlock(dos_Machine* m,uint16_t blockSeg,uint16_t want){
{
if (((blockSeg == cast<uint16_t>(0ULL)) || (blockSeg <= m->firstMCB))) {
return {cast<uint16_t>(0ULL),false};
}
uint16_t mcb = cast<uint16_t>((blockSeg - cast<uint16_t>(1ULL)));
auto tmp88 = dos_Machine_readMCB(m,mcb);
uint8_t mark = std::get<0>(tmp88);
uint16_t owner = std::get<1>(tmp88);
uint16_t size = std::get<2>(tmp88);
if ((want <= size)) {
if ((want < size)) {
uint16_t rem = cast<uint16_t>((cast<uint16_t>((mcb + cast<uint16_t>(1ULL))) + want));
dos_Machine_writeMCB(m,rem,mark,cast<uint16_t>(0ULL),cast<uint16_t>((cast<uint16_t>((size - want)) - cast<uint16_t>(1ULL))));
dos_Machine_writeMCB(m,mcb,cast<uint8_t>(77ULL),owner,want);
dos_Machine_coalesce(m);
}
return {want,true};
}
uint16_t avail = size;
uint8_t lastMark = mark;
uint16_t next = cast<uint16_t>((cast<uint16_t>((mcb + cast<uint16_t>(1ULL))) + size));
{int64_t i = cast<int64_t>(0ULL);for (;((i < cast<int64_t>(4096ULL)) && (lastMark != cast<uint8_t>(90ULL)));i++){
auto tmp89 = dos_Machine_readMCB(m,next);
uint8_t nmark = std::get<0>(tmp89);
uint16_t nowner = std::get<1>(tmp89);
uint16_t nsize = std::get<2>(tmp89);
if ((nowner != cast<uint16_t>(0ULL))) {
break;
}
avail += cast<uint16_t>((nsize + cast<uint16_t>(1ULL)));
lastMark = nmark;
next += cast<uint16_t>((cast<uint16_t>(1ULL) + nsize));
}
}if ((want > avail)) {
return {avail,false};
}
if ((avail > want)) {
uint16_t rem = cast<uint16_t>((cast<uint16_t>((mcb + cast<uint16_t>(1ULL))) + want));
dos_Machine_writeMCB(m,rem,lastMark,cast<uint16_t>(0ULL),cast<uint16_t>((cast<uint16_t>((avail - want)) - cast<uint16_t>(1ULL))));
dos_Machine_writeMCB(m,mcb,cast<uint8_t>(77ULL),owner,want);
}
else {
dos_Machine_writeMCB(m,mcb,lastMark,owner,want);
}
return {want,true};
}
}
// tools/platform/dos/dos_mouse.go:35:1
dos_mouseState* dos_Machine_mouse(dos_Machine* m){
{
if ((!m->ms)) {
m->ms = arenaNew(dos_mouseState{cast<int64_t>(160ULL),cast<int64_t>(100ULL),{},{},{},cast<int64_t>(0ULL),cast<int64_t>(319ULL),cast<int64_t>(0ULL),cast<int64_t>(199ULL),{},{},{}});
}
return m->ms;
}
}
// tools/platform/dos/dos_mouse.go:42:1
bool dos_Machine_int33(dos_Machine* m,x86_CPU* c){
{
dos_mouseState* ms = dos_Machine_mouse(m);
if ((!m->Int33Hist)) {
m->Int33Hist = Map<uint16_t,int64_t>{};
}
m->Int33Hist[x86_CPU_Reg16(c,cast<int64_t>(0ULL))]++;
{
switch(x86_CPU_Reg16(c,cast<int64_t>(0ULL))){
case cast<uint16_t>(0ULL):{
auto tmp90 = std::make_tuple(cast<int64_t>(160ULL),cast<int64_t>(100ULL));
ms->x = std::get<0>(tmp90);
ms->y = std::get<1>(tmp90);
auto tmp91 = std::make_tuple(cast<int64_t>(0ULL),cast<int64_t>(0ULL),cast<uint8_t>(0ULL));
ms->accX = std::get<0>(tmp91);
ms->accY = std::get<1>(tmp91);
ms->buttons = std::get<2>(tmp91);
ms->hidden = true;
x86_CPU_SetReg16(c,cast<int64_t>(0ULL),cast<uint16_t>(65535ULL));
x86_CPU_SetReg16(c,cast<int64_t>(3ULL),cast<uint16_t>(2ULL));
break;}
case cast<uint16_t>(1ULL):{
ms->hidden = false;
break;}
case cast<uint16_t>(2ULL):{
ms->hidden = true;
break;}
case cast<uint16_t>(3ULL):{
x86_CPU_SetReg16(c,cast<int64_t>(1ULL),cast<uint16_t>(ms->x));
x86_CPU_SetReg16(c,cast<int64_t>(2ULL),cast<uint16_t>(ms->y));
x86_CPU_SetReg16(c,cast<int64_t>(3ULL),cast<uint16_t>(ms->buttons));
break;}
case cast<uint16_t>(4ULL):{
auto tmp92 = std::make_tuple(cast<int64_t>(cast<int16_t>(x86_CPU_Reg16(c,cast<int64_t>(1ULL)))),cast<int64_t>(cast<int16_t>(x86_CPU_Reg16(c,cast<int64_t>(2ULL)))));
ms->x = std::get<0>(tmp92);
ms->y = std::get<1>(tmp92);
break;}
case cast<uint16_t>(5ULL):{
int64_t cnt = ms->pressL;
if ((x86_CPU_Reg16(c,cast<int64_t>(3ULL)) == cast<uint16_t>(1ULL))) {
cnt = ms->pressR;
ms->pressR = cast<int64_t>(0ULL);
}
else {
ms->pressL = cast<int64_t>(0ULL);
}
x86_CPU_SetReg16(c,cast<int64_t>(0ULL),cast<uint16_t>(ms->buttons));
x86_CPU_SetReg16(c,cast<int64_t>(3ULL),cast<uint16_t>(cnt));
x86_CPU_SetReg16(c,cast<int64_t>(1ULL),cast<uint16_t>(ms->x));
x86_CPU_SetReg16(c,cast<int64_t>(2ULL),cast<uint16_t>(ms->y));
break;}
case cast<uint16_t>(6ULL):{
x86_CPU_SetReg16(c,cast<int64_t>(0ULL),cast<uint16_t>(ms->buttons));
x86_CPU_SetReg16(c,cast<int64_t>(3ULL),cast<uint16_t>(0ULL));
x86_CPU_SetReg16(c,cast<int64_t>(1ULL),cast<uint16_t>(ms->x));
x86_CPU_SetReg16(c,cast<int64_t>(2ULL),cast<uint16_t>(ms->y));
break;}
case cast<uint16_t>(7ULL):{
auto tmp93 = std::make_tuple(cast<int64_t>(x86_CPU_Reg16(c,cast<int64_t>(1ULL))),cast<int64_t>(x86_CPU_Reg16(c,cast<int64_t>(2ULL))));
ms->xMin = std::get<0>(tmp93);
ms->xMax = std::get<1>(tmp93);
break;}
case cast<uint16_t>(8ULL):{
auto tmp94 = std::make_tuple(cast<int64_t>(x86_CPU_Reg16(c,cast<int64_t>(1ULL))),cast<int64_t>(x86_CPU_Reg16(c,cast<int64_t>(2ULL))));
ms->yMin = std::get<0>(tmp94);
ms->yMax = std::get<1>(tmp94);
break;}
case cast<uint16_t>(11ULL):{
x86_CPU_SetReg16(c,cast<int64_t>(1ULL),cast<uint16_t>(cast<int16_t>(ms->accX)));
x86_CPU_SetReg16(c,cast<int64_t>(2ULL),cast<uint16_t>(cast<int16_t>(ms->accY)));
auto tmp95 = std::make_tuple(cast<int64_t>(0ULL),cast<int64_t>(0ULL));
ms->accX = std::get<0>(tmp95);
ms->accY = std::get<1>(tmp95);
break;}
case cast<uint16_t>(12ULL):case cast<uint16_t>(20ULL):{
break;}
case cast<uint16_t>(15ULL):case cast<uint16_t>(26ULL):case cast<uint16_t>(28ULL):{
break;}
default:{
break;}
}}
return true;
}
}
// tools/platform/dos/dos_mouse.go:104:1
void dos_mouseState_clamp(dos_mouseState* ms){
{
if ((ms->x < ms->xMin)) {
ms->x = ms->xMin;
}
if ((ms->x > ms->xMax)) {
ms->x = ms->xMax;
}
if ((ms->y < ms->yMin)) {
ms->y = ms->yMin;
}
if ((ms->y > ms->yMax)) {
ms->y = ms->yMax;
}
}
}
// tools/platform/dos/dos_mouse.go:122:1
void dos_Machine_HomeMouse(dos_Machine* m){
{
dos_mouseState* ms = dos_Machine_mouse(m);
ms->accX -= cast<int64_t>(4000ULL);
ms->accY -= cast<int64_t>(4000ULL);
auto tmp96 = std::make_tuple(cast<int64_t>(0ULL),cast<int64_t>(0ULL));
ms->x = std::get<0>(tmp96);
ms->y = std::get<1>(tmp96);
}
}
// tools/platform/dos/dos_mouse.go:133:1
void dos_Machine_MoveMouseTo(dos_Machine* m,int64_t x,int64_t y){
{
dos_mouseState* ms = dos_Machine_mouse(m);
auto tmp97 = std::make_tuple(cast<int64_t>((x - ms->x)),cast<int64_t>((y - ms->y)));
int64_t dx = std::get<0>(tmp97);
int64_t dy = std::get<1>(tmp97);
auto tmp98 = std::make_tuple(x,y);
ms->x = std::get<0>(tmp98);
ms->y = std::get<1>(tmp98);
dos_mouseState_clamp(ms);
ms->accX += cast<int64_t>((dx * cast<int64_t>(2ULL)));
ms->accY += cast<int64_t>((dy * cast<int64_t>(2ULL)));
}
}
// tools/platform/dos/dos_mouse.go:146:1
void dos_Machine_FeedMickeys(dos_Machine* m,int64_t dx,int64_t dy){
{
dos_mouseState* ms = dos_Machine_mouse(m);
ms->accX += dx;
ms->accY += dy;
}
}
// tools/platform/dos/dos_mouse.go:154:1
void dos_Machine_SetMouseButtons(dos_Machine* m,uint8_t mask){
{
dos_mouseState* ms = dos_Machine_mouse(m);
if (((cast<uint8_t>((mask & cast<uint8_t>(1ULL))) != cast<uint8_t>(0ULL)) && (cast<uint8_t>((ms->buttons & cast<uint8_t>(1ULL))) == cast<uint8_t>(0ULL)))) {
ms->pressL++;
}
if (((cast<uint8_t>((mask & cast<uint8_t>(2ULL))) != cast<uint8_t>(0ULL)) && (cast<uint8_t>((ms->buttons & cast<uint8_t>(2ULL))) == cast<uint8_t>(0ULL)))) {
ms->pressR++;
}
ms->buttons = mask;
}
}
// tools/platform/dos/dos_vga.go:30:1
void dos_Machine_vgaInit13h(dos_Machine* m,bool clear){
{
dos_vgaState* v = m->vga;
v->seq[cast<int64_t>(2ULL)] = cast<uint8_t>(15ULL);
if(rrcapture::trace.active)rrDOSWrite(&(v->seq[cast<int64_t>(2ULL)]));
v->seq[cast<int64_t>(4ULL)] = cast<uint8_t>(14ULL);
if(rrcapture::trace.active)rrDOSWrite(&(v->seq[cast<int64_t>(4ULL)]));
v->gc[cast<int64_t>(4ULL)] = cast<uint8_t>(0ULL);
v->gc[cast<int64_t>(5ULL)] = cast<uint8_t>(64ULL);
v->gc[cast<int64_t>(8ULL)] = cast<uint8_t>(255ULL);
auto tmp99 = std::make_tuple(cast<uint8_t>(0ULL),cast<uint8_t>(0ULL));
v->crtc[cast<int64_t>(12ULL)] = std::get<0>(tmp99);
v->crtc[cast<int64_t>(13ULL)] = std::get<1>(tmp99);
if(rrcapture::trace.active)rrDOSWrite(&(v->crtc[cast<int64_t>(12ULL)]));
if(rrcapture::trace.active)rrDOSWrite(&(v->crtc[cast<int64_t>(13ULL)]));
v->crtc[cast<int64_t>(19ULL)] = cast<uint8_t>(40ULL);
if(rrcapture::trace.active)rrDOSWrite(&(v->crtc[cast<int64_t>(19ULL)]));
if (clear) {
{auto&& tmp100 = v->planes;
for(int64_t tmp101=0;tmp101<len(tmp100);++tmp101){
auto p=tmp101;{auto&& tmp102 = v->planes[p];
for(int64_t tmp103=0;tmp103<len(tmp102);++tmp103){
auto i=tmp103;v->planes[p][i] = cast<uint8_t>(0ULL);
if(rrcapture::trace.active)rrDOSWrite(&(v->planes[p][i]));
}}
}}
}
}
}
// tools/platform/dos/dos_vga.go:48:1
bool dos_vgaState_chained(dos_vgaState* v){
{
return (cast<uint8_t>((v->seq[cast<int64_t>(4ULL)] & cast<uint8_t>(8ULL))) != cast<uint8_t>(0ULL));
}
}
// tools/platform/dos/dos_vga.go:51:1
void dos_Machine_vgaWrite(dos_Machine* m,uint32_t a,uint8_t val){
{
dos_vgaState* v = m->vga;
uint32_t off = cast<uint32_t>((a - cast<uint32_t>(655360ULL)));
if (dos_vgaState_chained(v)) {
v->planes[cast<uint32_t>((off & cast<uint32_t>(3ULL)))][cast<uint32_t>(((shr<uint32_t>(off,cast<int64_t>(2ULL))) & cast<uint32_t>(65535ULL)))] = val;
if(rrcapture::trace.active)rrDOSWrite(&(v->planes[cast<uint32_t>((off & cast<uint32_t>(3ULL)))][cast<uint32_t>(((shr<uint32_t>(off,cast<int64_t>(2ULL))) & cast<uint32_t>(65535ULL)))]));
return ;
}
off &= cast<uint32_t>(65535ULL);
uint8_t mask = cast<uint8_t>((v->seq[cast<int64_t>(2ULL)] & cast<uint8_t>(15ULL)));
{
switch(cast<uint8_t>((v->gc[cast<int64_t>(5ULL)] & cast<uint8_t>(3ULL)))){
case cast<uint8_t>(1ULL):{
{int64_t p = cast<int64_t>(0ULL);for (;(p < cast<int64_t>(4ULL));p++){
if ((cast<uint8_t>((mask & (shl<uint8_t>(cast<uint8_t>(1ULL),p)))) != cast<uint8_t>(0ULL))) {
v->planes[p][off] = v->latch[p];
if(rrcapture::trace.active)rrDOSWrite(&(v->planes[p][off]));
}
}
}break;}
case cast<uint8_t>(2ULL):{
uint8_t bm = v->gc[cast<int64_t>(8ULL)];
{int64_t p = cast<int64_t>(0ULL);for (;(p < cast<int64_t>(4ULL));p++){
if ((cast<uint8_t>((mask & (shl<uint8_t>(cast<uint8_t>(1ULL),p)))) == cast<uint8_t>(0ULL))) {
continue;
}
uint8_t d={};
if ((cast<uint8_t>((val & (shl<uint8_t>(cast<uint8_t>(1ULL),p)))) != cast<uint8_t>(0ULL))) {
d = cast<uint8_t>(255ULL);
}
v->planes[p][off] = cast<uint8_t>(((cast<uint8_t>((d & bm))) | ((v->latch[p] & ~(bm)))));
if(rrcapture::trace.active)rrDOSWrite(&(v->planes[p][off]));
}
}break;}
default:{
uint8_t bm = v->gc[cast<int64_t>(8ULL)];
auto tmp104 = std::make_tuple(v->gc[cast<int64_t>(0ULL)],v->gc[cast<int64_t>(1ULL)]);
uint8_t sr = std::get<0>(tmp104);
uint8_t esr = std::get<1>(tmp104);
{int64_t p = cast<int64_t>(0ULL);for (;(p < cast<int64_t>(4ULL));p++){
if ((cast<uint8_t>((mask & (shl<uint8_t>(cast<uint8_t>(1ULL),p)))) == cast<uint8_t>(0ULL))) {
continue;
}
uint8_t d = val;
if ((cast<uint8_t>((esr & (shl<uint8_t>(cast<uint8_t>(1ULL),p)))) != cast<uint8_t>(0ULL))) {
if ((cast<uint8_t>((sr & (shl<uint8_t>(cast<uint8_t>(1ULL),p)))) != cast<uint8_t>(0ULL))) {
d = cast<uint8_t>(255ULL);
}
else {
d = cast<uint8_t>(0ULL);
}
}
v->planes[p][off] = cast<uint8_t>(((cast<uint8_t>((d & bm))) | ((v->latch[p] & ~(bm)))));
if(rrcapture::trace.active)rrDOSWrite(&(v->planes[p][off]));
}
}break;}
}}
}
}
// tools/platform/dos/dos_vga.go:101:1
uint8_t dos_Machine_vgaRead(dos_Machine* m,uint32_t a){
{
dos_vgaState* v = m->vga;
uint32_t off = cast<uint32_t>((a - cast<uint32_t>(655360ULL)));
if (dos_vgaState_chained(v)) {
return v->planes[cast<uint32_t>((off & cast<uint32_t>(3ULL)))][cast<uint32_t>(((shr<uint32_t>(off,cast<int64_t>(2ULL))) & cast<uint32_t>(65535ULL)))];
}
off &= cast<uint32_t>(65535ULL);
{int64_t p = cast<int64_t>(0ULL);for (;(p < cast<int64_t>(4ULL));p++){
v->latch[p] = v->planes[p][off];
}
}if ((cast<uint8_t>((v->gc[cast<int64_t>(5ULL)] & cast<uint8_t>(8ULL))) != cast<uint8_t>(0ULL))) {
auto tmp105 = std::make_tuple(v->gc[cast<int64_t>(2ULL)],v->gc[cast<int64_t>(7ULL)]);
uint8_t cmp = std::get<0>(tmp105);
uint8_t dc = std::get<1>(tmp105);
uint8_t out=cast<uint8_t>(255ULL);
{int64_t p = cast<int64_t>(0ULL);for (;(p < cast<int64_t>(4ULL));p++){
if ((cast<uint8_t>((dc & (shl<uint8_t>(cast<uint8_t>(1ULL),p)))) == cast<uint8_t>(0ULL))) {
continue;
}
uint8_t want={};
if ((cast<uint8_t>((cmp & (shl<uint8_t>(cast<uint8_t>(1ULL),p)))) != cast<uint8_t>(0ULL))) {
want = cast<uint8_t>(255ULL);
}
out &= cast<uint8_t>(~(cast<uint8_t>((v->latch[p] ^ want))));
}
}return out;
}
return v->latch[cast<uint8_t>((v->gc[cast<int64_t>(4ULL)] & cast<uint8_t>(3ULL)))];
}
}
// tools/platform/dos/dos_vga.go:130:1
void dos_Machine_vgaRegOut(dos_Machine* m,uint16_t port,uint8_t b){
{
dos_vgaState* v = m->vga;
{
switch(port){
case cast<uint16_t>(964ULL):{
v->seqIdx = b;
break;}
case cast<uint16_t>(965ULL):{
if ((cast<uint8_t>((v->seqIdx & cast<uint8_t>(7ULL))) == cast<uint8_t>(4ULL))) {
bool was = dos_vgaState_chained(v);
v->seq[cast<int64_t>(4ULL)] = b;
if(rrcapture::trace.active)rrDOSWrite(&(v->seq[cast<int64_t>(4ULL)]));
if ((was != dos_vgaState_chained(v))) {
dos_Machine_logf(m,std::string("VGA: chain-4 %v -> %v (seq[4]=%02X) at %04X:%04X",48),was,dos_vgaState_chained(v),b,m->CPU->Seg[cast<int64_t>(1ULL)],m->CPU->IP);
}
return ;
}
v->seq[cast<uint8_t>((v->seqIdx & cast<uint8_t>(7ULL)))] = b;
if(rrcapture::trace.active)rrDOSWrite(&(v->seq[cast<uint8_t>((v->seqIdx & cast<uint8_t>(7ULL)))]));
break;}
case cast<uint16_t>(974ULL):{
v->gcIdx = b;
break;}
case cast<uint16_t>(975ULL):{
v->gc[cast<uint8_t>((v->gcIdx & cast<uint8_t>(15ULL)))] = b;
break;}
case cast<uint16_t>(980ULL):{
v->crtcIdx = b;
break;}
case cast<uint16_t>(981ULL):{
v->crtc[cast<uint8_t>((v->crtcIdx & cast<uint8_t>(31ULL)))] = b;
if(rrcapture::trace.active)rrDOSWrite(&(v->crtc[cast<uint8_t>((v->crtcIdx & cast<uint8_t>(31ULL)))]));
break;}
}}
}
}
// tools/platform/dos/dos_video.go:21:1
bool dos_Machine_int10(dos_Machine* m,x86_CPU* c){
{
uint8_t ah = x86_CPU_Reg8(c,cast<int64_t>(4ULL));
uint32_t bda = cast<uint32_t>(1024ULL);
uint16_t key = cast<uint16_t>((cast<uint16_t>(4096ULL) | cast<uint16_t>(ah)));
if ((!m->video)) {
m->video = Map<uint16_t,int64_t>{};
}
m->video[key]++;
if ((get(m->video,key) == cast<int64_t>(1ULL))) {
dos_Machine_logf(m,std::string("INT 10h AH=%02X AL=%02X BX=%04X at %04X:%04X",44),ah,x86_CPU_Reg8(c,cast<int64_t>(0ULL)),x86_CPU_Reg16(c,cast<int64_t>(3ULL)),c->Seg[cast<int64_t>(1ULL)],c->IP);
}
{
switch(ah){
case cast<uint8_t>(0ULL):{
uint8_t mode = cast<uint8_t>((x86_CPU_Reg8(c,cast<int64_t>(0ULL)) & cast<uint8_t>(127ULL)));
m->Mem[cast<uint32_t>((bda + cast<uint32_t>(73ULL)))] = mode;
dos_Machine_vgaInit13h(m,(cast<uint8_t>((x86_CPU_Reg8(c,cast<int64_t>(0ULL)) & cast<uint8_t>(128ULL))) == cast<uint8_t>(0ULL)));
uint8_t cols = cast<uint8_t>(80ULL);
if ((((((mode <= cast<uint8_t>(1ULL)) || (mode == cast<uint8_t>(4ULL))) || (mode == cast<uint8_t>(5ULL))) || (mode == cast<uint8_t>(13ULL))) || (mode == cast<uint8_t>(19ULL)))) {
cols = cast<uint8_t>(40ULL);
}
if ((mode == cast<uint8_t>(19ULL))) {
cols = cast<uint8_t>(40ULL);
}
dos_Machine_w16(m,cast<uint32_t>((bda + cast<uint32_t>(74ULL))),cast<uint16_t>(cols));
return true;
break;}
case cast<uint8_t>(1ULL):case cast<uint8_t>(2ULL):case cast<uint8_t>(5ULL):case cast<uint8_t>(6ULL):case cast<uint8_t>(7ULL):case cast<uint8_t>(9ULL):case cast<uint8_t>(10ULL):case cast<uint8_t>(11ULL):case cast<uint8_t>(12ULL):case cast<uint8_t>(14ULL):case cast<uint8_t>(19ULL):{
return true;
break;}
case cast<uint8_t>(3ULL):{
x86_CPU_SetReg16(c,cast<int64_t>(2ULL),cast<uint16_t>(0ULL));
x86_CPU_SetReg16(c,cast<int64_t>(1ULL),cast<uint16_t>(1543ULL));
return true;
break;}
case cast<uint8_t>(8ULL):{
x86_CPU_SetReg16(c,cast<int64_t>(0ULL),cast<uint16_t>(1824ULL));
return true;
break;}
case cast<uint8_t>(13ULL):{
x86_CPU_SetReg8(c,cast<int64_t>(0ULL),cast<uint8_t>(0ULL));
return true;
break;}
case cast<uint8_t>(15ULL):{
x86_CPU_SetReg8(c,cast<int64_t>(0ULL),m->Mem[cast<uint32_t>((bda + cast<uint32_t>(73ULL)))]);
x86_CPU_SetReg8(c,cast<int64_t>(4ULL),cast<uint8_t>(dos_Machine_r16(m,cast<uint32_t>((bda + cast<uint32_t>(74ULL))))));
x86_CPU_SetReg8(c,cast<int64_t>(7ULL),cast<uint8_t>(0ULL));
return true;
break;}
case cast<uint8_t>(16ULL):{
{
switch(x86_CPU_Reg8(c,cast<int64_t>(0ULL))){
case cast<uint8_t>(16ULL):{
int64_t i = cast<int64_t>((modi<int64_t>(cast<int64_t>(x86_CPU_Reg16(c,cast<int64_t>(3ULL))),cast<int64_t>(256ULL)) * cast<int64_t>(3ULL)));
m->io->Pal[i] = x86_CPU_Reg8(c,cast<int64_t>(6ULL));
if(rrcapture::trace.active)rrDOSWrite(&(m->io->Pal[i]));
m->io->Pal[cast<int64_t>((i + cast<int64_t>(1ULL)))] = x86_CPU_Reg8(c,cast<int64_t>(5ULL));
if(rrcapture::trace.active)rrDOSWrite(&(m->io->Pal[cast<int64_t>((i + cast<int64_t>(1ULL)))]));
m->io->Pal[cast<int64_t>((i + cast<int64_t>(2ULL)))] = x86_CPU_Reg8(c,cast<int64_t>(1ULL));
if(rrcapture::trace.active)rrDOSWrite(&(m->io->Pal[cast<int64_t>((i + cast<int64_t>(2ULL)))]));
break;}
case cast<uint8_t>(18ULL):{
uint32_t src = dos_lin(c->Seg[cast<int64_t>(0ULL)],x86_CPU_Reg16(c,cast<int64_t>(2ULL)));
int64_t start = cast<int64_t>((cast<int64_t>(x86_CPU_Reg16(c,cast<int64_t>(3ULL))) * cast<int64_t>(3ULL)));
int64_t n = cast<int64_t>((cast<int64_t>(x86_CPU_Reg16(c,cast<int64_t>(1ULL))) * cast<int64_t>(3ULL)));
{int64_t i = cast<int64_t>(0ULL);for (;((i < n) && (cast<int64_t>((start + i)) < cast<int64_t>(768ULL)));i++){
m->io->Pal[cast<int64_t>((start + i))] = m->Mem[cast<uint32_t>(((cast<uint32_t>((src + cast<uint32_t>(i)))) & cast<uint32_t>(1048575ULL)))];
if(rrcapture::trace.active)rrDOSWrite(&(m->io->Pal[cast<int64_t>((start + i))]));
}
}break;}
case cast<uint8_t>(21ULL):{
x86_CPU_SetReg16(c,cast<int64_t>(1ULL),cast<uint16_t>(0ULL));
x86_CPU_SetReg8(c,cast<int64_t>(6ULL),cast<uint8_t>(0ULL));
break;}
case cast<uint8_t>(23ULL):{
int64_t n = cast<int64_t>((cast<int64_t>(x86_CPU_Reg16(c,cast<int64_t>(1ULL))) * cast<int64_t>(3ULL)));
uint32_t dst = dos_lin(c->Seg[cast<int64_t>(0ULL)],x86_CPU_Reg16(c,cast<int64_t>(2ULL)));
{int64_t i = cast<int64_t>(0ULL);for (;(i < n);i++){
m->Mem[cast<uint32_t>(((cast<uint32_t>((dst + cast<uint32_t>(i)))) & cast<uint32_t>(1048575ULL)))] = cast<uint8_t>(0ULL);
}
}break;}
}}
return true;
break;}
case cast<uint8_t>(17ULL):{
if ((x86_CPU_Reg8(c,cast<int64_t>(0ULL)) == cast<uint8_t>(48ULL))) {
c->Seg[cast<int64_t>(0ULL)] = cast<uint16_t>(61440ULL);
x86_CPU_SetReg16(c,cast<int64_t>(5ULL),cast<uint16_t>(64110ULL));
x86_CPU_SetReg16(c,cast<int64_t>(1ULL),cast<uint16_t>(16ULL));
x86_CPU_SetReg8(c,cast<int64_t>(2ULL),cast<uint8_t>(24ULL));
}
return true;
break;}
case cast<uint8_t>(18ULL):{
if ((x86_CPU_Reg8(c,cast<int64_t>(3ULL)) == cast<uint8_t>(16ULL))) {
x86_CPU_SetReg8(c,cast<int64_t>(7ULL),cast<uint8_t>(0ULL));
x86_CPU_SetReg8(c,cast<int64_t>(3ULL),cast<uint8_t>(3ULL));
x86_CPU_SetReg16(c,cast<int64_t>(1ULL),cast<uint16_t>(9ULL));
}
else {
x86_CPU_SetReg8(c,cast<int64_t>(0ULL),cast<uint8_t>(18ULL));
}
return true;
break;}
case cast<uint8_t>(26ULL):{
if ((x86_CPU_Reg8(c,cast<int64_t>(0ULL)) == cast<uint8_t>(0ULL))) {
x86_CPU_SetReg8(c,cast<int64_t>(0ULL),cast<uint8_t>(26ULL));
x86_CPU_SetReg8(c,cast<int64_t>(3ULL),cast<uint8_t>(8ULL));
x86_CPU_SetReg8(c,cast<int64_t>(7ULL),cast<uint8_t>(0ULL));
}
return true;
break;}
case cast<uint8_t>(27ULL):{
x86_CPU_SetReg8(c,cast<int64_t>(0ULL),cast<uint8_t>(0ULL));
return true;
break;}
}}
return true;
}
}
// tools/platform/dos/go32.go:161:1
std::tuple<dos_PM*,Error> dos_LoadGo32(std::string exePath,std::string gameDir){
{
auto tmp106 = go_os_ReadFile(exePath);
Slice<uint8_t> data = std::get<0>(tmp106);
Error err = std::get<1>(tmp106);
if (bool(err)) {
return {{},err};
}
return dos_LoadGo32Bytes(data,gameDir);
}
}
// tools/platform/dos/go32.go:171:1
std::tuple<dos_PM*,Error> dos_LoadGo32Bytes(Slice<uint8_t> data,std::string gameDir){
{
auto tmp107 = dos_ParseGo32COFF(data);
dos_COFF* coff = std::get<0>(tmp107);
Error err = std::get<1>(tmp107);
if (bool(err)) {
return {{},err};
}
dos_PM* p = arenaNew(dos_PM{Slice<uint8_t>::make(cast<int64_t>(68157440ULL)),{},gameDir,Map<uint16_t,os_File*>{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},cast<uint16_t>(256ULL),{},Map<uint16_t,uint32_t>{{cast<uint16_t>(8ULL),cast<uint32_t>(1048576ULL)},{cast<uint16_t>(16ULL),cast<uint32_t>(1048576ULL)},{cast<uint16_t>(24ULL),cast<uint32_t>(0ULL)}},true,{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},Map<uint16_t,int64_t>{},Map<uint8_t,int64_t>{},Map<uint8_t,int64_t>{},{},{},{}});
uint32_t imgEnd={};
{auto&& tmp108 = coff->Sections;
for(int64_t tmp109=0;tmp109<len(tmp108);++tmp109){
auto s=tmp108[tmp109];uint32_t end = cast<uint32_t>((s.VAddr + s.Size));
if ((end > imgEnd)) {
imgEnd = end;
}
if ((dos_COFFSection_IsBSS(s) || (!s.Data))) {
continue;
}
if ((cast<int64_t>((cast<int64_t>((cast<int64_t>(1048576ULL) + cast<int64_t>(s.VAddr))) + len(s.Data))) > len(p->Mem))) {
return {{},go_fmt_Errorf(std::string("go32: section %q at VA %#x..%#x exceeds backing memory",54),s.Name,s.VAddr,cast<uint32_t>((s.VAddr + s.Size)))};
}
gcopy(sub(p->Mem,cast<uint32_t>((cast<uint32_t>(1048576ULL) + s.VAddr)),len(p->Mem)),s.Data);
}}
p->convBase = cast<uint32_t>(65536ULL);
p->convTop = cast<uint32_t>(655360ULL);
uint32_t xferLinear = p->convBase;
p->convNext = cast<uint32_t>((p->convBase + cast<uint32_t>(8192ULL)));
p->infoBase = cast<uint32_t>((cast<uint32_t>(len(p->Mem)) - cast<uint32_t>(16384ULL)));
uint32_t stackTopVA = cast<uint32_t>(67092480ULL);
p->stackFloor = cast<uint32_t>((stackTopVA - cast<uint32_t>(8388608ULL)));
p->imgEnd = imgEnd;
p->heapBase = dos_align(dos_maxu32(imgEnd,cast<uint32_t>(1048576ULL)),cast<uint32_t>(4096ULL));
p->heapNext = p->heapBase;
dos_PM_setupInfoBlock(p,xferLinear);
dos_PM_setupDefaultIntVec(p);
x86_CPU* c = x86_NewCPU(p);
c->Mode = cast<int64_t>(1ULL);
c->Seg[cast<int64_t>(1ULL)] = cast<uint16_t>(8ULL);
c->Seg[cast<int64_t>(3ULL)] = cast<uint16_t>(16ULL);
c->Seg[cast<int64_t>(0ULL)] = cast<uint16_t>(16ULL);
c->Seg[cast<int64_t>(2ULL)] = cast<uint16_t>(16ULL);
c->Seg[cast<int64_t>(5ULL)] = cast<uint16_t>(16ULL);
{auto&& tmp110 = c->SegBase;
for(int64_t tmp111=0;tmp111<len(tmp110);++tmp111){
auto i=tmp111;c->SegBase[i] = cast<uint32_t>(1048576ULL);
}}
c->Seg[cast<int64_t>(4ULL)] = cast<uint16_t>(64ULL);
c->SegBase[cast<int64_t>(4ULL)] = p->infoBase;
c->IP = coff->Entry;
c->Regs[cast<int64_t>(4ULL)] = stackTopVA;
c->IF = true;
c->IntHook = [=](auto...args){return dos_PM_handleInt(p,args...);};
c->SegResolve = [=](auto...args){return dos_PM_resolveSel(p,args...);};
c->PortIn = [=](auto...args){return dos_PM_portIn(p,args...);};
c->PortOut = [=](auto...args){return dos_PM_portOut(p,args...);};
p->CPU = c;
return {p,{}};
}
}
// tools/platform/dos/go32.go:265:1
uint32_t dos_PM_resolveSel(dos_PM* p,uint16_t sel){
{
return get(p->sels,sel);
}
}
// tools/platform/dos/go32.go:268:1
void dos_PM_mapSel(dos_PM* p,uint16_t sel,uint32_t base){
{
p->sels[sel] = base;
}
}
// tools/platform/dos/go32.go:299:1
void dos_PM_setupInfoBlock(dos_PM* p,uint32_t xferLinear){
{
dos_PM_mapSel(p,cast<uint16_t>(64ULL),p->infoBase);
uint32_t b = p->infoBase;
dos_PM_w32(p,cast<uint32_t>((b + cast<uint32_t>(0ULL))),cast<uint32_t>(48ULL));
dos_PM_w32(p,cast<uint32_t>((b + cast<uint32_t>(4ULL))),cast<uint32_t>(753664ULL));
dos_PM_w32(p,cast<uint32_t>((b + cast<uint32_t>(8ULL))),cast<uint32_t>(720896ULL));
dos_PM_w32(p,cast<uint32_t>((b + cast<uint32_t>(12ULL))),xferLinear);
dos_PM_w32(p,cast<uint32_t>((b + cast<uint32_t>(16ULL))),cast<uint32_t>(8192ULL));
dos_PM_w32(p,cast<uint32_t>((b + cast<uint32_t>(20ULL))),cast<uint32_t>(8388608ULL));
dos_PM_Write(p,cast<uint32_t>((b + cast<uint32_t>(24ULL))),cast<uint8_t>(8ULL));
dos_PM_Write(p,cast<uint32_t>((b + cast<uint32_t>(25ULL))),cast<uint8_t>(112ULL));
dos_PM_Write(p,cast<uint32_t>((b + cast<uint32_t>(26ULL))),cast<uint8_t>(24ULL));
dos_PM_w32(p,cast<uint32_t>((b + cast<uint32_t>(28ULL))),p->stackFloor);
dos_PM_w16(p,cast<uint32_t>((b + cast<uint32_t>(32ULL))),cast<uint16_t>(8192ULL));
dos_PM_w16(p,cast<uint32_t>((b + cast<uint32_t>(36ULL))),cast<uint16_t>(shr<uint32_t>(xferLinear,cast<int64_t>(4ULL))));
}
}
// tools/platform/dos/go32.go:327:1
void dos_PM_setupDefaultIntVec(dos_PM* p){
{
uint32_t stubLinear = cast<uint32_t>((p->infoBase + cast<uint32_t>(512ULL)));
Slice<uint8_t> stub = Slice<uint8_t>{cast<uint8_t>(228ULL),cast<uint8_t>(96ULL),cast<uint8_t>(176ULL),cast<uint8_t>(32ULL),cast<uint8_t>(230ULL),cast<uint8_t>(160ULL),cast<uint8_t>(230ULL),cast<uint8_t>(32ULL),cast<uint8_t>(207ULL)};
{auto&& tmp112 = stub;
for(int64_t tmp113=0;tmp113<len(tmp112);++tmp113){
auto i=tmp113;auto b=tmp112[tmp113];dos_PM_Write(p,cast<uint32_t>((stubLinear + cast<uint32_t>(i))),b);
}}
p->defIntVec = dos_pmVector{cast<uint16_t>(8ULL),cast<uint32_t>((stubLinear - cast<uint32_t>(1048576ULL))),true};
}
}
// tools/platform/dos/go32.go:374:1
bool dos_PM_writeDOSStructures(dos_PM* p){
{
if ((p->convTop == cast<uint32_t>(0ULL))) {
return false;
}
uint32_t psp = cast<uint32_t>((p->convBase - cast<uint32_t>(256ULL)));
uint32_t jft = p->convBase;
uint32_t lol = cast<uint32_t>((p->convBase + cast<uint32_t>(64ULL)));
uint32_t sft = cast<uint32_t>((p->convBase + cast<uint32_t>(128ULL)));
uint32_t sftCount = cast<uint32_t>(64ULL);
if ((cast<uint32_t>((cast<uint32_t>((sft + cast<uint32_t>(6ULL))) + cast<uint32_t>((sftCount * cast<uint32_t>(59ULL))))) > cast<uint32_t>((p->convBase + cast<uint32_t>(8192ULL))))) {
return false;
}
dos_PM_w16(p,cast<uint32_t>((psp + cast<uint32_t>(50ULL))),cast<uint16_t>(64ULL));
dos_PM_w16(p,cast<uint32_t>((psp + cast<uint32_t>(52ULL))),cast<uint16_t>(cast<uint32_t>((jft & cast<uint32_t>(15ULL)))));
dos_PM_w16(p,cast<uint32_t>((psp + cast<uint32_t>(54ULL))),cast<uint16_t>(shr<uint32_t>(jft,cast<int64_t>(4ULL))));
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < cast<uint32_t>(64ULL));i++){
dos_PM_Write(p,cast<uint32_t>((jft + i)),cast<uint8_t>(i));
}
}dos_PM_w16(p,cast<uint32_t>((lol + cast<uint32_t>(4ULL))),cast<uint16_t>(cast<uint32_t>((sft & cast<uint32_t>(15ULL)))));
dos_PM_w16(p,cast<uint32_t>((lol + cast<uint32_t>(6ULL))),cast<uint16_t>(shr<uint32_t>(sft,cast<int64_t>(4ULL))));
dos_PM_w16(p,cast<uint32_t>((sft + cast<uint32_t>(0ULL))),cast<uint16_t>(65535ULL));
dos_PM_w16(p,cast<uint32_t>((sft + cast<uint32_t>(2ULL))),cast<uint16_t>(65535ULL));
dos_PM_w16(p,cast<uint32_t>((sft + cast<uint32_t>(4ULL))),cast<uint16_t>(sftCount));
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < sftCount);i++){
dos_PM_w16(p,cast<uint32_t>((cast<uint32_t>((sft + cast<uint32_t>(6ULL))) + cast<uint32_t>((i * cast<uint32_t>(59ULL))))),cast<uint16_t>(1ULL));
}
}auto tmp114 = std::make_tuple(cast<uint16_t>(shr<uint32_t>(lol,cast<int64_t>(4ULL))),cast<uint16_t>(cast<uint32_t>((lol & cast<uint32_t>(15ULL)))));
p->lolSeg = std::get<0>(tmp114);
p->lolOff = std::get<1>(tmp114);
return true;
}
}
// tools/platform/dos/go32.go:423:1
void dos_PM_enforceBaseAddress(dos_PM* p){
{if(p->rrQuakeBase){{
dos_PM_w32(p,cast<uint32_t>(1496780ULL),cast<uint32_t>(1048576ULL));
}
}}}
// tools/platform/dos/go32.go:429:1
uint8_t dos_PM_Read(dos_PM* p,uint32_t a){
{
if ((cast<int64_t>(a) < len(p->Mem))) {
uint8_t v = p->Mem[a];
if (((bool(p->onR) && (a >= p->wRLo)) && (a < p->wRHi))) {
p->onR(a,cast<uint32_t>(v),dos_PM_watchPC(p));
}
return v;
}
dos_PM_fault(p,std::string("read",4),a);
return cast<uint8_t>(255ULL);
}
}
// tools/platform/dos/go32.go:441:1
void dos_PM_Write(dos_PM* p,uint32_t a,uint8_t v){
{
if ((cast<int64_t>(a) < len(p->Mem))) {
p->Mem[a] = v;
if(rrcapture::trace.active)rrDOSWrite(&(p->Mem[a]));
if (((bool(p->onW) && (a >= p->wWLo)) && (a < p->wWHi))) {
p->onW(a,cast<uint32_t>(v),dos_PM_watchPC(p));
}
return ;
}
dos_PM_fault(p,std::string("write",5),a);
}
}
// tools/platform/dos/go32.go:454:1
uint32_t dos_PM_watchPC(dos_PM* p){
{
if ((!p->CPU)) {
return cast<uint32_t>(0ULL);
}
return x86_CPU_LinearPC(p->CPU);
}
}
// tools/platform/dos/go32.go:464:1
void dos_PM_SetWriteWatch(dos_PM* p,uint32_t lo,uint32_t hi,std::function<void(uint32_t,uint32_t,uint32_t)> cb){
{
auto tmp115 = std::make_tuple(lo,hi,cb);
p->wWLo = std::get<0>(tmp115);
p->wWHi = std::get<1>(tmp115);
p->onW = std::get<2>(tmp115);
}
}
// tools/platform/dos/go32.go:467:1
void dos_PM_SetReadWatch(dos_PM* p,uint32_t lo,uint32_t hi,std::function<void(uint32_t,uint32_t,uint32_t)> cb){
{
auto tmp116 = std::make_tuple(lo,hi,cb);
p->wRLo = std::get<0>(tmp116);
p->wRHi = std::get<1>(tmp116);
p->onR = std::get<2>(tmp116);
}
}
// tools/platform/dos/go32.go:483:1
Slice<dos_MemRegion> dos_PM_MemRegions(dos_PM* p){
{
uint32_t b = cast<uint32_t>(1048576ULL);
return Slice<dos_MemRegion>{dos_MemRegion{std::string("BIOS data / low",15),cast<uint32_t>(0ULL),p->convBase},dos_MemRegion{std::string("conventional (DOS mem, __tb)",28),p->convBase,p->convTop},dos_MemRegion{std::string("VGA framebuffer",15),cast<uint32_t>(655360ULL),cast<uint32_t>(786432ULL)},dos_MemRegion{std::string("program image",13),b,cast<uint32_t>((b + p->imgEnd))},dos_MemRegion{std::string("heap / extended memory",22),cast<uint32_t>((b + p->heapBase)),cast<uint32_t>((b + p->stackFloor))},dos_MemRegion{std::string("PM stack",8),cast<uint32_t>((b + p->stackFloor)),p->infoBase},dos_MemRegion{std::string("go32 info block",15),p->infoBase,cast<uint32_t>(len(p->Mem))}};
}
}
// tools/platform/dos/go32.go:501:1
void dos_PM_fault(dos_PM* p,std::string kind,uint32_t a){
{
if ((bool(p->CPU) && (!p->CPU->Halted))) {
x86_CPU_Halt(p->CPU,std::string("out-of-range %s at linear %08X (PC %s, %d MiB backing)",54),kind,a,dos_pcHex(p->CPU),cast<int64_t>(65ULL));
}
}
}
// tools/platform/dos/go32.go:510:1
uint32_t dos_align(uint32_t v,uint32_t a){
{
return ((cast<uint32_t>((cast<uint32_t>((v + a)) - cast<uint32_t>(1ULL)))) & ~((cast<uint32_t>((a - cast<uint32_t>(1ULL))))));
}
}
// tools/platform/dos/go32.go:511:1
uint32_t dos_maxu32(uint32_t a,uint32_t b){
{
if ((a > b)) {
return a;
}
return b;
}
}
// tools/platform/dos/go32.go:517:1
std::string dos_pcHex(x86_CPU* c){
{
return go_fmt_Sprintf(std::string("%08X",4),c->IP);
}
}
// tools/platform/dos/go32.go:521:1
uint32_t dos_PM_r32(dos_PM* p,uint32_t a){
{
return cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(dos_PM_Read(p,a)) | shl<uint32_t>(cast<uint32_t>(dos_PM_Read(p,cast<uint32_t>((a + cast<uint32_t>(1ULL))))),cast<int64_t>(8ULL)))) | shl<uint32_t>(cast<uint32_t>(dos_PM_Read(p,cast<uint32_t>((a + cast<uint32_t>(2ULL))))),cast<int64_t>(16ULL)))) | shl<uint32_t>(cast<uint32_t>(dos_PM_Read(p,cast<uint32_t>((a + cast<uint32_t>(3ULL))))),cast<int64_t>(24ULL))));
}
}
// tools/platform/dos/go32.go:524:1
void dos_PM_w32(dos_PM* p,uint32_t a,uint32_t v){
{
dos_PM_Write(p,a,cast<uint8_t>(v));
dos_PM_Write(p,cast<uint32_t>((a + cast<uint32_t>(1ULL))),cast<uint8_t>(shr<uint32_t>(v,cast<int64_t>(8ULL))));
dos_PM_Write(p,cast<uint32_t>((a + cast<uint32_t>(2ULL))),cast<uint8_t>(shr<uint32_t>(v,cast<int64_t>(16ULL))));
dos_PM_Write(p,cast<uint32_t>((a + cast<uint32_t>(3ULL))),cast<uint8_t>(shr<uint32_t>(v,cast<int64_t>(24ULL))));
}
}
// tools/platform/dos/go32.go:530:1
void dos_PM_w16(dos_PM* p,uint32_t a,uint16_t v){
{
dos_PM_Write(p,a,cast<uint8_t>(v));
dos_PM_Write(p,cast<uint32_t>((a + cast<uint32_t>(1ULL))),cast<uint8_t>(shr<uint16_t>(v,cast<int64_t>(8ULL))));
}
}
// tools/platform/dos/go32.go:536:1
std::string dos_PM_asciiz(dos_PM* p,uint32_t a){
{
Slice<uint8_t> b={};
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < cast<uint32_t>(4096ULL));i++){
uint8_t ch = dos_PM_Read(p,cast<uint32_t>((a + i)));
if ((ch == cast<uint8_t>(0ULL))) {
break;
}
b = append(b,Slice<uint8_t>{ch});
}
}return cast<std::string>(b);
}
}
// tools/platform/dos/go32_dosio.go:21:1
uint32_t dos_PM_rmLinear(dos_PM* p,uint16_t seg,uint32_t off){
{
return cast<uint32_t>((shl<uint32_t>(cast<uint32_t>(seg),cast<int64_t>(4ULL)) + (cast<uint32_t>((off & cast<uint32_t>(65535ULL))))));
}
}
// tools/platform/dos/go32_dosio.go:28:1
bool dos_PM_dosErr(dos_PM* p,dos_rmcs* r,uint16_t code){
{
r->flags |= cast<uint16_t>(1ULL);
r->eax = cast<uint32_t>(((cast<uint32_t>((r->eax & cast<uint32_t>(4294901760ULL)))) | cast<uint32_t>(code)));
return true;
}
}
// tools/platform/dos/go32_dosio.go:36:1
void dos_setAX(dos_rmcs* r,uint16_t v){
{
r->eax = cast<uint32_t>(((cast<uint32_t>((r->eax & cast<uint32_t>(4294901760ULL)))) | cast<uint32_t>(v)));
}
}
// tools/platform/dos/go32_dosio.go:37:1
void dos_setBX(dos_rmcs* r,uint16_t v){
{
r->ebx = cast<uint32_t>(((cast<uint32_t>((r->ebx & cast<uint32_t>(4294901760ULL)))) | cast<uint32_t>(v)));
}
}
// tools/platform/dos/go32_dosio.go:38:1
void dos_setCX(dos_rmcs* r,uint16_t v){
{
r->ecx = cast<uint32_t>(((cast<uint32_t>((r->ecx & cast<uint32_t>(4294901760ULL)))) | cast<uint32_t>(v)));
}
}
// tools/platform/dos/go32_dosio.go:39:1
void dos_setDX(dos_rmcs* r,uint16_t v){
{
r->edx = cast<uint32_t>(((cast<uint32_t>((r->edx & cast<uint32_t>(4294901760ULL)))) | cast<uint32_t>(v)));
}
}
// tools/platform/dos/go32_dosio.go:44:1
uint16_t dos_PM_allocHandle(dos_PM* p){
{
{uint16_t h = cast<uint16_t>(5ULL);for (;(h < cast<uint16_t>(255ULL));h++){
{
auto tmp117 = lookup(p->files,h);
bool ok = std::get<1>(tmp117);
if ((!ok)) {
return h;
}
}
}
}return cast<uint16_t>(65535ULL);
}
}
// tools/platform/dos/go32_dosio.go:57:1
std::string dos_PM_resolveHostPath(dos_PM* p,std::string name){
{
std::string base = p->gameDir;
if ((base == std::string("",0))) {
base = std::string(".",1);
}
name = go_strings_ReplaceAll(name,std::string("\\",1),std::string("/",1));
if (((len(name) >= cast<int64_t>(2ULL)) && (cast<uint8_t>(name[cast<int64_t>(1ULL)]) == cast<uint8_t>(58ULL)))) {
name = sub(name,cast<int64_t>(2ULL),len(name));
}
name = go_strings_TrimLeft(name,std::string("/",1));
std::string naive = go_filepath_Join(Slice<std::string>{base,go_filepath_FromSlash(name)});
{
auto tmp118 = go_os_Stat(naive);
Error err = std::get<1>(tmp118);
if ((!err)) {
return naive;
}
}
std::string cur = base;
{auto&& tmp119 = go_strings_Split(name,std::string("/",1));
for(int64_t tmp120=0;tmp120<len(tmp119);++tmp120){
auto comp=tmp119[tmp120];if (((comp == std::string("",0)) || (comp == std::string(".",1)))) {
continue;
}
std::string next = go_filepath_Join(Slice<std::string>{cur,comp});
{
auto tmp121 = go_os_Stat(next);
Error err = std::get<1>(tmp121);
if ((!err)) {
cur = next;
continue;
}
}
std::string found = std::string("",0);
{
auto tmp122 = go_os_ReadDir(cur);
Slice<RRFileInfo> ents = std::get<0>(tmp122);
Error err = std::get<1>(tmp122);
if ((!err)) {
{auto&& tmp123 = ents;
for(int64_t tmp124=0;tmp124<len(tmp123);++tmp124){
auto e=tmp123[tmp124];if (go_strings_EqualFold(rrFileInfo_Name(e),comp)) {
found = rrFileInfo_Name(e);
break;
}
}}
}
}
if ((found == std::string("",0))) {
return naive;
}
cur = go_filepath_Join(Slice<std::string>{cur,found});
}}
return cur;
}
}
// tools/platform/dos/go32_dosio.go:100:1
bool dos_PM_dosFile(dos_PM* p,dos_rmcs* r){
{
{
switch(cast<uint8_t>(shr<uint32_t>(r->eax,cast<int64_t>(8ULL)))){
case cast<uint8_t>(61ULL):{
return dos_PM_dosOpen(p,r);
break;}
case cast<uint8_t>(60ULL):{
return dos_PM_dosCreate(p,r);
break;}
case cast<uint8_t>(62ULL):{
return dos_PM_dosClose(p,r);
break;}
case cast<uint8_t>(63ULL):{
return dos_PM_dosRead(p,r);
break;}
case cast<uint8_t>(64ULL):{
return dos_PM_dosWrite(p,r);
break;}
case cast<uint8_t>(66ULL):{
return dos_PM_dosSeek(p,r);
break;}
case cast<uint8_t>(68ULL):{
return dos_PM_dosIoctl(p,r);
break;}
case cast<uint8_t>(87ULL):{
return dos_PM_dosFileTime(p,r);
break;}
case cast<uint8_t>(65ULL):{
go_os_Remove(dos_PM_resolveHostPath(p,dos_PM_asciiz(p,dos_PM_rmLinear(p,r->ds,r->edx))));
return true;
break;}
case cast<uint8_t>(67ULL):{
if ((cast<uint8_t>(r->eax) == cast<uint8_t>(0ULL))) {
dos_setAX(r,cast<uint16_t>(32ULL));
}
return true;
break;}
case cast<uint8_t>(71ULL):{
dos_PM_Write(p,dos_PM_rmLinear(p,r->ds,r->esi),cast<uint8_t>(0ULL));
dos_setAX(r,cast<uint16_t>(256ULL));
return true;
break;}
case cast<uint8_t>(78ULL):case cast<uint8_t>(79ULL):{
return dos_PM_dosErr(p,r,cast<uint16_t>(18ULL));
break;}
case cast<uint8_t>(96ULL):{
return dos_PM_dosTruename(p,r);
break;}
default:{
return false;
break;}
}}
}
}
// tools/platform/dos/go32_dosio.go:145:1
bool dos_PM_dosTruename(dos_PM* p,dos_rmcs* r){
{
std::string src = dos_PM_asciiz(p,dos_PM_rmLinear(p,r->ds,r->esi));
std::string canon = dos_canonicalizeDOSPath(src);
uint32_t dst = dos_PM_rmLinear(p,r->es,r->edi);
{int64_t i = cast<int64_t>(0ULL);for (;(i < len(canon));i++){
dos_PM_Write(p,cast<uint32_t>((dst + cast<uint32_t>(i))),cast<uint8_t>(canon[i]));
}
}dos_PM_Write(p,cast<uint32_t>((dst + cast<uint32_t>(len(canon)))),cast<uint8_t>(0ULL));
dos_PM_logf(p,std::string("DOS truename %q -> %q",21),src,canon);
return true;
}
}
// tools/platform/dos/go32_dosio.go:161:1
std::string dos_canonicalizeDOSPath(std::string s){
{
s = go_strings_ReplaceAll(s,std::string("/",1),std::string("\\",1));
std::string drive = std::string("C",1);
std::string rest = s;
if (((len(s) >= cast<int64_t>(2ULL)) && (cast<uint8_t>(s[cast<int64_t>(1ULL)]) == cast<uint8_t>(58ULL)))) {
drive = go_strings_ToUpper(sub(s,0,cast<int64_t>(1ULL)));
rest = sub(s,cast<int64_t>(2ULL),len(s));
}
Slice<std::string> stack={};
{auto&& tmp125 = go_strings_Split(rest,std::string("\\",1));
for(int64_t tmp126=0;tmp126<len(tmp125);++tmp126){
auto comp=tmp125[tmp126];{
auto tmp128=comp;
if (tmp128==(std::string("",0)) || tmp128==(std::string(".",1))){
}
else if (tmp128==(std::string("..",2))){
if ((len(stack) > cast<int64_t>(0ULL))) {
stack = sub(stack,0,cast<int64_t>((len(stack) - cast<int64_t>(1ULL))));
}
}
else {
stack = append(stack,Slice<std::string>{comp});
}
}
tmp127:;
}}
return go_strings_ToUpper(((drive + std::string(":\\",2)) + go_strings_Join(stack,std::string("\\",1))));
}
}
// tools/platform/dos/go32_dosio.go:185:1
bool dos_PM_dosOpen(dos_PM* p,dos_rmcs* r){
{
std::string name = dos_PM_asciiz(p,dos_PM_rmLinear(p,r->ds,r->edx));
std::string host = dos_PM_resolveHostPath(p,name);
os_File* f={};
Error err={};
{
switch(cast<uint8_t>((cast<uint8_t>(r->eax) & cast<uint8_t>(3ULL)))){
case cast<uint8_t>(1ULL):{
auto tmp129 = go_os_OpenFile(host,1,cast<fs_FileMode>(420ULL));
f = std::get<0>(tmp129);
err = std::get<1>(tmp129);
break;}
case cast<uint8_t>(2ULL):{
auto tmp130 = go_os_OpenFile(host,2,cast<fs_FileMode>(420ULL));
f = std::get<0>(tmp130);
err = std::get<1>(tmp130);
break;}
default:{
auto tmp131 = go_os_Open(host);
f = std::get<0>(tmp131);
err = std::get<1>(tmp131);
break;}
}}
if (bool(err)) {
dos_PM_logf(p,std::string("DOS open %q FAILED (%v)",23),name,err);
return dos_PM_dosErr(p,r,cast<uint16_t>(2ULL));
}
uint16_t h = dos_PM_allocHandle(p);
p->files[h] = f;
dos_setAX(r,h);
dos_PM_logf(p,std::string("DOS open %q -> handle %d",24),name,h);
return true;
}
}
// tools/platform/dos/go32_dosio.go:209:1
bool dos_PM_dosCreate(dos_PM* p,dos_rmcs* r){
{
std::string name = dos_PM_asciiz(p,dos_PM_rmLinear(p,r->ds,r->edx));
std::string host = dos_PM_resolveHostPath(p,name);
auto tmp132 = go_os_OpenFile(host,578,cast<fs_FileMode>(420ULL));
os_File* f = std::get<0>(tmp132);
Error err = std::get<1>(tmp132);
if (bool(err)) {
dos_PM_logf(p,std::string("DOS create %q FAILED (%v)",25),name,err);
return dos_PM_dosErr(p,r,cast<uint16_t>(5ULL));
}
uint16_t h = dos_PM_allocHandle(p);
p->files[h] = f;
dos_setAX(r,h);
dos_PM_logf(p,std::string("DOS create %q -> handle %d",26),name,h);
return true;
}
}
// tools/platform/dos/go32_dosio.go:224:1
bool dos_PM_dosClose(dos_PM* p,dos_rmcs* r){
{
uint16_t h = cast<uint16_t>(r->ebx);
{
auto tmp133 = lookup(p->files,h);
os_File* f = std::get<0>(tmp133);
bool ok = std::get<1>(tmp133);
if (ok) {
os_File_Close(f);
removeKey(p->files,h);
return true;
}
}
if ((h <= cast<uint16_t>(4ULL))) {
return true;
}
return dos_PM_dosErr(p,r,cast<uint16_t>(6ULL));
}
}
// tools/platform/dos/go32_dosio.go:237:1
bool dos_PM_dosRead(dos_PM* p,dos_rmcs* r){
{
uint16_t h = cast<uint16_t>(r->ebx);
int64_t n = cast<int64_t>(cast<uint32_t>((r->ecx & cast<uint32_t>(65535ULL))));
uint32_t dst = dos_PM_rmLinear(p,r->ds,r->edx);
auto tmp134 = lookup(p->files,h);
os_File* f = std::get<0>(tmp134);
bool ok = std::get<1>(tmp134);
if ((!ok)) {
if ((h == cast<uint16_t>(0ULL))) {
dos_setAX(r,cast<uint16_t>(0ULL));
return true;
}
return dos_PM_dosErr(p,r,cast<uint16_t>(6ULL));
}
Slice<uint8_t> buf = Slice<uint8_t>::make(n);
int64_t got = cast<int64_t>(0ULL);
{;for (;(got < n);){
auto tmp135 = os_File_Read(f,sub(buf,got,len(buf)));
int64_t m = std::get<0>(tmp135);
Error err = std::get<1>(tmp135);
got += m;
if (bool(err)) {
break;
}
}
}{int64_t i = cast<int64_t>(0ULL);for (;(i < got);i++){
dos_PM_Write(p,cast<uint32_t>((dst + cast<uint32_t>(i))),buf[i]);
}
}dos_setAX(r,cast<uint16_t>(got));
return true;
}
}
// tools/platform/dos/go32_dosio.go:265:1
bool dos_PM_dosWrite(dos_PM* p,dos_rmcs* r){
{
uint16_t h = cast<uint16_t>(r->ebx);
int64_t n = cast<int64_t>(cast<uint32_t>((r->ecx & cast<uint32_t>(65535ULL))));
uint32_t src = dos_PM_rmLinear(p,r->ds,r->edx);
Slice<uint8_t> data = Slice<uint8_t>::make(n);
{int64_t i = cast<int64_t>(0ULL);for (;(i < n);i++){
data[i] = dos_PM_Read(p,cast<uint32_t>((src + cast<uint32_t>(i))));
}
}{
switch(h){
case cast<uint16_t>(1ULL):case cast<uint16_t>(2ULL):{
p->Console = append(p->Console,data);
dos_setAX(r,cast<uint16_t>(n));
return true;
break;}
}}
auto tmp136 = lookup(p->files,h);
os_File* f = std::get<0>(tmp136);
bool ok = std::get<1>(tmp136);
if ((!ok)) {
return dos_PM_dosErr(p,r,cast<uint16_t>(6ULL));
}
auto tmp137 = os_File_Write(f,data);
int64_t m = std::get<0>(tmp137);
dos_setAX(r,cast<uint16_t>(m));
return true;
}
}
// tools/platform/dos/go32_dosio.go:288:1
bool dos_PM_dosSeek(dos_PM* p,dos_rmcs* r){
{
uint16_t h = cast<uint16_t>(r->ebx);
auto tmp138 = lookup(p->files,h);
os_File* f = std::get<0>(tmp138);
bool ok = std::get<1>(tmp138);
if ((!ok)) {
return dos_PM_dosErr(p,r,cast<uint16_t>(6ULL));
}
int64_t off = cast<int64_t>(cast<int32_t>(cast<uint32_t>((shl<uint32_t>(cast<uint32_t>(cast<uint32_t>((r->ecx & cast<uint32_t>(65535ULL)))),cast<int64_t>(16ULL)) | cast<uint32_t>(cast<uint32_t>((r->edx & cast<uint32_t>(65535ULL))))))));
auto tmp139 = os_File_Seek(f,off,cast<int64_t>(cast<uint8_t>(r->eax)));
int64_t pos = std::get<0>(tmp139);
Error err = std::get<1>(tmp139);
if (bool(err)) {
return dos_PM_dosErr(p,r,cast<uint16_t>(25ULL));
}
dos_setAX(r,cast<uint16_t>(pos));
dos_setDX(r,cast<uint16_t>(shr<int64_t>(pos,cast<int64_t>(16ULL))));
return true;
}
}
// tools/platform/dos/go32_dosio.go:308:1
bool dos_PM_dosFileTime(dos_PM* p,dos_rmcs* r){
{
uint16_t h = cast<uint16_t>(r->ebx);
auto tmp140 = lookup(p->files,h);
os_File* f = std::get<0>(tmp140);
bool ok = std::get<1>(tmp140);
if ((!ok)) {
return dos_PM_dosErr(p,r,cast<uint16_t>(6ULL));
}
{
switch(cast<uint8_t>(r->eax)){
case cast<uint8_t>(0ULL):{
auto tmp141 = dos_dosDateTime(f);
uint16_t date = std::get<0>(tmp141);
uint16_t tm = std::get<1>(tmp141);
dos_setDX(r,date);
dos_setCX(r,tm);
return true;
break;}
default:{
return true;
break;}
}}
}
}
// tools/platform/dos/go32_dosio.go:344:1
bool dos_PM_dosIoctl(dos_PM* p,dos_rmcs* r){
{
uint16_t h = cast<uint16_t>(r->ebx);
{
switch(cast<uint8_t>(r->eax)){
case cast<uint8_t>(0ULL):{
if ((h <= cast<uint16_t>(2ULL))) {
dos_setDX(r,cast<uint16_t>(32979ULL));
}
else {
dos_setDX(r,cast<uint16_t>(2ULL));
}
return true;
break;}
case cast<uint8_t>(1ULL):{
return true;
break;}
default:{
return true;
break;}
}}
}
}
// tools/platform/dos/go32_dpmi.go:17:1
bool dos_PM_handleInt(dos_PM* p,x86_CPU* c,uint8_t n){
{rrprof::Scope timing(1,"DOS and DPMI services");
{
{
switch(n){
case cast<uint8_t>(49ULL):{
return dos_PM_dpmi(p,c);
break;}
case cast<uint8_t>(33ULL):{
return dos_PM_int21(p,c);
break;}
case cast<uint8_t>(47ULL):{
x86_CPU_SetReg8(c,cast<int64_t>(0ULL),cast<uint8_t>(0ULL));
c->CF = false;
dos_PM_countInt(p,n);
return true;
break;}
case cast<uint8_t>(16ULL):{
c->CF = false;
dos_PM_countInt(p,n);
return true;
break;}
default:{
return false;
break;}
}}
}
}
}
// tools/platform/dos/go32_dpmi.go:37:1
void dos_PM_countInt(dos_PM* p,uint8_t n){
{
p->IntCounts[n]++;
if ((get(p->IntCounts,n) == cast<int64_t>(1ULL))) {
dos_PM_logf(p,std::string("INT %02Xh (stubbed) AX=%04X at %08X",35),n,x86_CPU_Reg16(p->CPU,cast<int64_t>(0ULL)),p->CPU->IP);
}
}
}
// tools/platform/dos/go32_dpmi.go:46:1
bool dos_PM_dpmi(dos_PM* p,x86_CPU* c){
{
uint16_t fn = x86_CPU_Reg16(c,cast<int64_t>(0ULL));
p->DPMICounts[fn]++;
c->CF = false;
auto get32 = [&](int64_t i)->uint32_t{
return c->Regs[i];
}
;
auto set16 = [&](int64_t i,uint16_t v)->void{
x86_CPU_SetReg16(c,i,v);
}
;
auto setPair = [&](int64_t hi,int64_t lo,uint32_t v)->void{
x86_CPU_SetReg16(c,hi,cast<uint16_t>(shr<uint32_t>(v,cast<int64_t>(16ULL))));
x86_CPU_SetReg16(c,lo,cast<uint16_t>(v));
}
;
auto fail = [&](uint16_t errCode)->void{
c->CF = true;
x86_CPU_SetReg16(c,cast<int64_t>(0ULL),errCode);
}
;
{
switch(fn){
case cast<uint16_t>(0ULL):{
set16(cast<int64_t>(0ULL),dos_PM_allocSel(p,x86_CPU_Reg16(c,cast<int64_t>(1ULL))));
break;}
case cast<uint16_t>(1ULL):{
break;}
case cast<uint16_t>(3ULL):{
set16(cast<int64_t>(0ULL),cast<uint16_t>(8ULL));
break;}
case cast<uint16_t>(6ULL):{
setPair(cast<int64_t>(1ULL),cast<int64_t>(2ULL),dos_PM_resolveSel(p,x86_CPU_Reg16(c,cast<int64_t>(3ULL))));
break;}
case cast<uint16_t>(7ULL):{
dos_PM_mapSel(p,x86_CPU_Reg16(c,cast<int64_t>(3ULL)),cast<uint32_t>((shl<uint32_t>(cast<uint32_t>(x86_CPU_Reg16(c,cast<int64_t>(1ULL))),cast<int64_t>(16ULL)) | cast<uint32_t>(x86_CPU_Reg16(c,cast<int64_t>(2ULL))))));
break;}
case cast<uint16_t>(8ULL):{
break;}
case cast<uint16_t>(9ULL):{
break;}
case cast<uint16_t>(10ULL):{
uint16_t alias = dos_PM_allocSel(p,cast<uint16_t>(1ULL));
dos_PM_mapSel(p,alias,dos_PM_resolveSel(p,x86_CPU_Reg16(c,cast<int64_t>(3ULL))));
set16(cast<int64_t>(0ULL),alias);
break;}
case cast<uint16_t>(11ULL):{
dos_PM_writeDescriptor(p,get32(cast<int64_t>(7ULL)),dos_PM_resolveSel(p,x86_CPU_Reg16(c,cast<int64_t>(3ULL))));
break;}
case cast<uint16_t>(12ULL):{
dos_PM_mapSel(p,x86_CPU_Reg16(c,cast<int64_t>(3ULL)),dos_PM_readDescriptorBase(p,get32(cast<int64_t>(7ULL))));
break;}
case cast<uint16_t>(512ULL):{
setPair(cast<int64_t>(1ULL),cast<int64_t>(2ULL),cast<uint32_t>(0ULL));
break;}
case cast<uint16_t>(513ULL):{
break;}
case cast<uint16_t>(516ULL):{
dos_pmVector v = p->pmVectors[x86_CPU_Reg8(c,cast<int64_t>(3ULL))];
if ((!v.set)) {
v = p->defIntVec;
}
set16(cast<int64_t>(1ULL),v.sel);
c->Regs[cast<int64_t>(2ULL)] = v.off;
break;}
case cast<uint16_t>(517ULL):{
dos_PM_setPMVector(p,x86_CPU_Reg8(c,cast<int64_t>(3ULL)),x86_CPU_Reg16(c,cast<int64_t>(1ULL)),get32(cast<int64_t>(2ULL)));
break;}
case cast<uint16_t>(514ULL):case cast<uint16_t>(515ULL):{
break;}
case cast<uint16_t>(256ULL):{
auto tmp142 = dos_PM_allocConv(p,x86_CPU_Reg16(c,cast<int64_t>(3ULL)));
uint16_t seg = std::get<0>(tmp142);
uint16_t sel = std::get<1>(tmp142);
bool ok = std::get<2>(tmp142);
if ((!ok)) {
fail(cast<uint16_t>(8ULL));
set16(cast<int64_t>(3ULL),cast<uint16_t>(shr<uint32_t>((cast<uint32_t>((p->convTop - p->convNext))),cast<int64_t>(4ULL))));
break;
}
set16(cast<int64_t>(0ULL),seg);
set16(cast<int64_t>(2ULL),sel);
break;}
case cast<uint16_t>(257ULL):{
break;}
case cast<uint16_t>(1024ULL):{
set16(cast<int64_t>(0ULL),cast<uint16_t>(256ULL));
set16(cast<int64_t>(3ULL),cast<uint16_t>(1ULL));
x86_CPU_SetReg8(c,cast<int64_t>(1ULL),cast<uint8_t>(6ULL));
x86_CPU_SetReg8(c,cast<int64_t>(6ULL),cast<uint8_t>(8ULL));
x86_CPU_SetReg8(c,cast<int64_t>(2ULL),cast<uint8_t>(112ULL));
break;}
case cast<uint16_t>(1280ULL):{
dos_PM_freeMemInfo(p,get32(cast<int64_t>(7ULL)));
break;}
case cast<uint16_t>(1281ULL):{
uint32_t size = cast<uint32_t>((shl<uint32_t>(cast<uint32_t>(x86_CPU_Reg16(c,cast<int64_t>(3ULL))),cast<int64_t>(16ULL)) | cast<uint32_t>(x86_CPU_Reg16(c,cast<int64_t>(1ULL)))));
auto tmp143 = dos_PM_allocHeap(p,size);
uint32_t base = std::get<0>(tmp143);
bool ok = std::get<1>(tmp143);
if ((!ok)) {
fail(cast<uint16_t>(8ULL));
break;
}
setPair(cast<int64_t>(3ULL),cast<int64_t>(1ULL),base);
setPair(cast<int64_t>(6ULL),cast<int64_t>(7ULL),base);
break;}
case cast<uint16_t>(1282ULL):{
break;}
case cast<uint16_t>(1283ULL):{
uint32_t size = cast<uint32_t>((shl<uint32_t>(cast<uint32_t>(x86_CPU_Reg16(c,cast<int64_t>(3ULL))),cast<int64_t>(16ULL)) | cast<uint32_t>(x86_CPU_Reg16(c,cast<int64_t>(1ULL)))));
auto tmp144 = dos_PM_allocHeap(p,size);
uint32_t base = std::get<0>(tmp144);
bool ok = std::get<1>(tmp144);
if ((!ok)) {
fail(cast<uint16_t>(8ULL));
break;
}
setPair(cast<int64_t>(3ULL),cast<int64_t>(1ULL),base);
setPair(cast<int64_t>(6ULL),cast<int64_t>(7ULL),base);
break;}
case cast<uint16_t>(3584ULL):{
set16(cast<int64_t>(0ULL),cast<uint16_t>(69ULL));
break;}
case cast<uint16_t>(3585ULL):{
break;}
case cast<uint16_t>(1287ULL):{
break;}
case cast<uint16_t>(1536ULL):case cast<uint16_t>(1537ULL):{
break;}
case cast<uint16_t>(1540ULL):{
setPair(cast<int64_t>(3ULL),cast<int64_t>(1ULL),cast<uint32_t>(4096ULL));
break;}
case cast<uint16_t>(2304ULL):{
x86_CPU_SetReg8(c,cast<int64_t>(0ULL),dos_b2u8(p->virtIF));
p->virtIF = false;
break;}
case cast<uint16_t>(2305ULL):{
x86_CPU_SetReg8(c,cast<int64_t>(0ULL),dos_b2u8(p->virtIF));
p->virtIF = true;
break;}
case cast<uint16_t>(2306ULL):{
x86_CPU_SetReg8(c,cast<int64_t>(0ULL),dos_b2u8(p->virtIF));
break;}
case cast<uint16_t>(768ULL):{
return dos_PM_simulateRealInt(p,c);
break;}
case cast<uint16_t>(771ULL):{
auto tmp145 = dos_PM_allocCallback(p);
uint16_t seg = std::get<0>(tmp145);
uint16_t off = std::get<1>(tmp145);
set16(cast<int64_t>(1ULL),seg);
set16(cast<int64_t>(2ULL),off);
break;}
case cast<uint16_t>(772ULL):{
break;}
default:{
x86_CPU_Halt(c,std::string("unimplemented DPMI function AX=%04X (BX=%04X CX=%04X) at %08X",61),fn,x86_CPU_Reg16(c,cast<int64_t>(3ULL)),x86_CPU_Reg16(c,cast<int64_t>(1ULL)),c->IP);
return true;
break;}
}}
return true;
}
}
// tools/platform/dos/go32_dpmi.go:184:1
void dos_PM_writeDescriptor(dos_PM* p,uint32_t a,uint32_t base){
{
uint32_t limit=cast<uint32_t>(1048575ULL);
dos_PM_Write(p,cast<uint32_t>((a + cast<uint32_t>(0ULL))),cast<uint8_t>(limit));
dos_PM_Write(p,cast<uint32_t>((a + cast<uint32_t>(1ULL))),cast<uint8_t>(shr<uint32_t>(limit,cast<int64_t>(8ULL))));
dos_PM_Write(p,cast<uint32_t>((a + cast<uint32_t>(2ULL))),cast<uint8_t>(base));
dos_PM_Write(p,cast<uint32_t>((a + cast<uint32_t>(3ULL))),cast<uint8_t>(shr<uint32_t>(base,cast<int64_t>(8ULL))));
dos_PM_Write(p,cast<uint32_t>((a + cast<uint32_t>(4ULL))),cast<uint8_t>(shr<uint32_t>(base,cast<int64_t>(16ULL))));
dos_PM_Write(p,cast<uint32_t>((a + cast<uint32_t>(5ULL))),cast<uint8_t>(243ULL));
dos_PM_Write(p,cast<uint32_t>((a + cast<uint32_t>(6ULL))),cast<uint8_t>((cast<uint8_t>((cast<uint8_t>(shr<uint32_t>(limit,cast<int64_t>(16ULL))) & cast<uint8_t>(15ULL))) | cast<uint8_t>(192ULL))));
dos_PM_Write(p,cast<uint32_t>((a + cast<uint32_t>(7ULL))),cast<uint8_t>(shr<uint32_t>(base,cast<int64_t>(24ULL))));
}
}
// tools/platform/dos/go32_dpmi.go:198:1
uint32_t dos_PM_readDescriptorBase(dos_PM* p,uint32_t a){
{
return cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(dos_PM_Read(p,cast<uint32_t>((a + cast<uint32_t>(2ULL))))) | shl<uint32_t>(cast<uint32_t>(dos_PM_Read(p,cast<uint32_t>((a + cast<uint32_t>(3ULL))))),cast<int64_t>(8ULL)))) | shl<uint32_t>(cast<uint32_t>(dos_PM_Read(p,cast<uint32_t>((a + cast<uint32_t>(4ULL))))),cast<int64_t>(16ULL)))) | shl<uint32_t>(cast<uint32_t>(dos_PM_Read(p,cast<uint32_t>((a + cast<uint32_t>(7ULL))))),cast<int64_t>(24ULL))));
}
}
// tools/platform/dos/go32_dpmi.go:206:1
std::tuple<uint16_t,uint16_t> dos_PM_allocCallback(dos_PM* p){
uint16_t seg{};
uint16_t off{};
{
off = p->nextCallback;
p->nextCallback += cast<uint16_t>(4ULL);
return {cast<uint16_t>(49152ULL),off};
}
}
// tools/platform/dos/go32_dpmi.go:212:1
uint8_t dos_b2u8(bool v){
{
if (v) {
return cast<uint8_t>(1ULL);
}
return cast<uint8_t>(0ULL);
}
}
// tools/platform/dos/go32_dpmi.go:222:1
uint16_t dos_PM_allocSel(dos_PM* p,uint16_t n){
{
if ((n == cast<uint16_t>(0ULL))) {
n = cast<uint16_t>(1ULL);
}
uint16_t sel = p->nextSel;
p->nextSel += cast<uint16_t>((n * cast<uint16_t>(8ULL)));
return sel;
}
}
// tools/platform/dos/go32_dpmi.go:235:1
std::tuple<uint16_t,uint16_t,bool> dos_PM_allocConv(dos_PM* p,uint16_t paras){
uint16_t seg{};
uint16_t sel{};
bool ok{};
{
uint32_t bytes = shl<uint32_t>(cast<uint32_t>(paras),cast<int64_t>(4ULL));
if (((p->convTop == cast<uint32_t>(0ULL)) || (cast<uint32_t>((p->convNext + bytes)) > p->convTop))) {
return {cast<uint16_t>(0ULL),cast<uint16_t>(0ULL),false};
}
uint32_t base = p->convNext;
p->convNext = dos_align(cast<uint32_t>((p->convNext + bytes)),cast<uint32_t>(16ULL));
sel = cast<uint16_t>(shr<uint32_t>(base,cast<int64_t>(4ULL)));
dos_PM_mapSel(p,sel,base);
dos_PM_logf(p,std::string("DPMI 0100h: %d paras conv mem -> seg %04X sel %04X (linear %05X)",64),paras,cast<uint16_t>(shr<uint32_t>(base,cast<int64_t>(4ULL))),sel,base);
return {cast<uint16_t>(shr<uint32_t>(base,cast<int64_t>(4ULL))),sel,true};
}
}
// tools/platform/dos/go32_dpmi.go:250:1
std::tuple<uint32_t,bool> dos_PM_allocHeap(dos_PM* p,uint32_t size){
uint32_t base{};
bool ok{};
{
base = p->heapNext;
uint32_t next = dos_align(cast<uint32_t>((base + size)),cast<uint32_t>(16ULL));
if ((next > p->stackFloor)) {
return {cast<uint32_t>(0ULL),false};
}
p->heapNext = next;
dos_PM_logf(p,std::string("DPMI 0501h: %d bytes heap -> linear %08X",40),size,base);
return {base,true};
}
}
// tools/platform/dos/go32_dpmi.go:274:1
void dos_PM_freeMemInfo(dos_PM* p,uint32_t a){
{
uint32_t free = cast<uint32_t>((p->stackFloor - p->heapNext));
if ((free > cast<uint32_t>(16777216ULL))) {
free = cast<uint32_t>(16777216ULL);
}
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < cast<uint32_t>(48ULL));i += cast<uint32_t>(4ULL)){
dos_PM_w32(p,cast<uint32_t>((a + i)),cast<uint32_t>(4294967295ULL));
}
}dos_PM_w32(p,cast<uint32_t>((a + cast<uint32_t>(0ULL))),free);
dos_PM_w32(p,cast<uint32_t>((a + cast<uint32_t>(4ULL))),divi<uint32_t>(free,cast<uint32_t>(4096ULL)));
dos_PM_w32(p,cast<uint32_t>((a + cast<uint32_t>(8ULL))),divi<uint32_t>(free,cast<uint32_t>(4096ULL)));
dos_PM_w32(p,cast<uint32_t>((a + cast<uint32_t>(16ULL))),divi<uint32_t>(free,cast<uint32_t>(4096ULL)));
dos_PM_w32(p,cast<uint32_t>((a + cast<uint32_t>(20ULL))),divi<uint32_t>(free,cast<uint32_t>(4096ULL)));
}
}
// tools/platform/dos/go32_dpmi.go:291:1
bool dos_PM_int21(dos_PM* p,x86_CPU* c){
{
uint8_t ah = x86_CPU_Reg8(c,cast<int64_t>(4ULL));
c->CF = false;
{
switch(ah){
case cast<uint8_t>(0ULL):case cast<uint8_t>(76ULL):{
p->Terminated = true;
if ((ah == cast<uint8_t>(76ULL))) {
p->ExitCode = x86_CPU_Reg8(c,cast<int64_t>(0ULL));
}
x86_CPU_Halt(c,std::string("INT 21h/%02X \342\200\224 program exit (code %d) at %08X",47),ah,p->ExitCode,c->IP);
break;}
case cast<uint8_t>(9ULL):{
dos_PM_logf(p,std::string("DOS print: %q",13),dos_PM_dollarStr(p,c->Regs[cast<int64_t>(2ULL)]));
break;}
case cast<uint8_t>(37ULL):{
break;}
case cast<uint8_t>(48ULL):{
x86_CPU_SetReg8(c,cast<int64_t>(0ULL),cast<uint8_t>(7ULL));
x86_CPU_SetReg8(c,cast<int64_t>(4ULL),cast<uint8_t>(0ULL));
break;}
case cast<uint8_t>(53ULL):{
c->Seg[cast<int64_t>(0ULL)] = cast<uint16_t>(0ULL);
c->Regs[cast<int64_t>(3ULL)] = cast<uint32_t>(0ULL);
break;}
default:{
x86_CPU_Halt(c,std::string("unimplemented INT 21h AH=%02X in protected mode at %08X",55),ah,c->IP);
break;}
}}
return true;
}
}
// tools/platform/dos/go32_dpmi.go:317:1
std::string dos_PM_dollarStr(dos_PM* p,uint32_t a){
{
Slice<uint8_t> b={};
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < cast<uint32_t>(4096ULL));i++){
uint8_t ch = dos_PM_Read(p,cast<uint32_t>((a + i)));
if ((ch == cast<uint8_t>(36ULL))) {
break;
}
b = append(b,Slice<uint8_t>{ch});
}
}return cast<std::string>(b);
}
}
// tools/platform/dos/go32_input.go:40:1
void dos_PM_setPMVector(dos_PM* p,uint8_t bl,uint16_t sel,uint32_t off){
{
if ((!p->pmVectors[bl].set)) {
dos_PM_logf(p,std::string("DPMI 0205h: PM INT %02Xh handler -> %04X:%08X",45),bl,sel,off);
}
p->pmVectors[bl] = dos_pmVector{sel,off,true};
}
}
// tools/platform/dos/go32_input.go:56:1
void dos_PM_SetKeys(dos_PM* p,Slice<dos_injEvent> events){
{
p->keyEvents = events;
}
}
// tools/platform/dos/go32_input.go:59:1
bool dos_PM_KeysPending(dos_PM* p){
{
return (len(p->keyEvents) > cast<int64_t>(0ULL));
}
}
// tools/platform/dos/go32_input.go:65:1
void dos_PM_PumpInput(dos_PM* p,x86_CPU* c){
{
dos_PM_enforceBaseAddress(p);
dos_PM_pumpInteractiveKeys(p,c);
if ((len(p->keyEvents) == cast<int64_t>(0ULL))) {
return ;
}
if (p->keyRetry) {
if (dos_PM_deliverKey(p,c)) {
p->keyRetry = false;
dos_PM_popKeyEvent(p);
}
return ;
}
{
p->injTick++;
if ((p->injTick < cast<int64_t>(40000ULL))) {
return ;
}
}
p->injTick = cast<int64_t>(0ULL);
if ((p->keyWait > cast<int64_t>(0ULL))) {
p->keyWait--;
return ;
}
if ((p->keyEvents[cast<int64_t>(0ULL)].kind == cast<dos_injKind>(1ULL))) {
if ((!dos_PM_deliverKey(p,c))) {
p->keyRetry = true;
return ;
}
}
dos_PM_popKeyEvent(p);
}
}
// tools/platform/dos/go32_input.go:108:1
void dos_PM_EnqueueScancode(dos_PM* p,uint8_t sc){
{
p->injKeys = append(p->injKeys,Slice<uint8_t>{sc});
}
}
// tools/platform/dos/go32_input.go:111:1
bool dos_PM_InteractiveKeysPending(dos_PM* p){
{
return (len(p->injKeys) > cast<int64_t>(0ULL));
}
}
// tools/platform/dos/go32_input.go:122:1
void dos_PM_pumpInteractiveKeys(dos_PM* p,x86_CPU* c){
{
if (((len(p->injKeys) == cast<int64_t>(0ULL)) || p->kbdFull)) {
return ;
}
if (dos_PM_deliverScancode(p,c,p->injKeys[cast<int64_t>(0ULL)])) {
p->injKeys = sub(p->injKeys,cast<int64_t>(1ULL),len(p->injKeys));
}
}
}
// tools/platform/dos/go32_input.go:133:1
bool dos_PM_deliverKey(dos_PM* p,x86_CPU* c){
{
return dos_PM_deliverScancode(p,c,p->keyEvents[cast<int64_t>(0ULL)].code);
}
}
// tools/platform/dos/go32_input.go:148:1
bool dos_PM_deliverScancode(dos_PM* p,x86_CPU* c,uint8_t sc){
{
dos_pmVector v = p->pmVectors[cast<int64_t>(9ULL)];
if (((!v.set) || (v.off < p->heapBase))) {
return false;
}
auto tmp146 = std::make_tuple(sc,true);
p->kbdData = std::get<0>(tmp146);
p->kbdFull = std::get<1>(tmp146);
if ((!x86_CPU_InterruptPM(c,v.sel,v.off))) {
p->kbdFull = false;
return false;
}
p->keyHits++;
if ((p->keyHits <= cast<int64_t>(40ULL))) {
dos_PM_logf(p,std::string("INJECT KEY scancode %02X via INT9 %04X:%08X",43),sc,v.sel,v.off);
}
return true;
}
}
// tools/platform/dos/go32_input.go:166:1
void dos_PM_popKeyEvent(dos_PM* p){
{
p->keyWait = p->keyEvents[cast<int64_t>(0ULL)].delay;
p->keyEvents = sub(p->keyEvents,cast<int64_t>(1ULL),len(p->keyEvents));
}
}
// tools/platform/dos/go32_ports.go:39:1
uint64_t dos_PM_pitTotalTicks(dos_PM* p){
{
return divi<uint64_t>(p->CPU->Steps,cast<uint64_t>(32ULL));
}
}
// tools/platform/dos/go32_ports.go:43:1
uint16_t dos_PM_pitCounter(dos_PM* p){
{
auto tmp147 = cast<uint32_t>(p->pit.reload);
uint32_t reload = tmp147;
if ((reload == cast<uint32_t>(0ULL))) {
reload = cast<uint32_t>(65536ULL);
}
uint32_t e = cast<uint32_t>(modi<uint64_t>(dos_PM_pitTotalTicks(p),cast<uint64_t>(reload)));
return cast<uint16_t>(cast<uint32_t>(((cast<uint32_t>((reload - e))) & cast<uint32_t>(65535ULL))));
}
}
// tools/platform/dos/go32_ports.go:54:1
void dos_PM_pitSyncBiosTick(dos_PM* p){
{
dos_PM_w32(p,cast<uint32_t>(1132ULL),cast<uint32_t>(divi<uint64_t>(dos_PM_pitTotalTicks(p),cast<uint64_t>(65536ULL))));
}
}
// tools/platform/dos/go32_ports.go:56:1
uint32_t dos_PM_portIn(dos_PM* p,uint16_t port,int64_t size){
{
{
switch(port){
case cast<uint16_t>(64ULL):{
dos_PM_pitSyncBiosTick(p);
uint16_t v = p->pit.latched;
if ((!p->pit.haveLatch)) {
v = dos_PM_pitCounter(p);
}
uint8_t b={};
if (p->pit.readHi) {
b = cast<uint8_t>(shr<uint16_t>(v,cast<int64_t>(8ULL)));
p->pit.readHi = false;
p->pit.haveLatch = false;
}
else {
b = cast<uint8_t>(v);
p->pit.readHi = true;
}
return cast<uint32_t>(b);
break;}
case cast<uint16_t>(986ULL):case cast<uint16_t>(954ULL):{
p->retrace = (!p->retrace);
if (p->retrace) {
return cast<uint32_t>(9ULL);
}
return cast<uint32_t>(0ULL);
break;}
case cast<uint16_t>(96ULL):{
p->kbdFull = false;
return cast<uint32_t>(p->kbdData);
break;}
case cast<uint16_t>(100ULL):{
if (p->kbdFull) {
return cast<uint32_t>(1ULL);
}
return cast<uint32_t>(0ULL);
break;}
case cast<uint16_t>(32ULL):case cast<uint16_t>(33ULL):case cast<uint16_t>(160ULL):case cast<uint16_t>(161ULL):{
return cast<uint32_t>(0ULL);
break;}
default:{
return dos_widthMask8(size);
break;}
}}
}
}
// tools/platform/dos/go32_ports.go:95:1
void dos_PM_portOut(dos_PM* p,uint16_t port,int64_t size,uint32_t v){
{
uint8_t b = cast<uint8_t>(v);
{
switch(port){
case cast<uint16_t>(67ULL):{
if ((shr<uint8_t>(b,cast<int64_t>(6ULL)) == cast<uint8_t>(0ULL))) {
if ((cast<uint8_t>(((shr<uint8_t>(b,cast<int64_t>(4ULL))) & cast<uint8_t>(3ULL))) == cast<uint8_t>(0ULL))) {
p->pit.latched = dos_PM_pitCounter(p);
p->pit.haveLatch = true;
p->pit.readHi = false;
}
else {
p->pit.writeHi = false;
p->pit.readHi = false;
}
}
dos_PM_pitSyncBiosTick(p);
break;}
case cast<uint16_t>(64ULL):{
if (p->pit.writeHi) {
p->pit.reload = cast<uint16_t>(((cast<uint16_t>((p->pit.reload & cast<uint16_t>(255ULL)))) | shl<uint16_t>(cast<uint16_t>(b),cast<int64_t>(8ULL))));
p->pit.writeHi = false;
}
else {
p->pit.reload = cast<uint16_t>(((cast<uint16_t>((p->pit.reload & cast<uint16_t>(65280ULL)))) | cast<uint16_t>(b)));
p->pit.writeHi = true;
}
break;}
case cast<uint16_t>(968ULL):{
p->dacIndex = cast<int64_t>((cast<int64_t>(b) * cast<int64_t>(3ULL)));
break;}
case cast<uint16_t>(969ULL):{
p->Pal[modi<int64_t>(p->dacIndex,cast<int64_t>(768ULL))] = b;
if(rrcapture::trace.active)rrDOSWrite(&(p->Pal[modi<int64_t>(p->dacIndex,cast<int64_t>(768ULL))]));
p->dacIndex++;
break;}
}}
}
}
// tools/platform/dos/go32_rmint.go:24:1
dos_rmcs dos_PM_readRMCS(dos_PM* p,uint32_t a){
{
dos_rmcs r = dos_rmcs{a,{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{}};
r.edi = dos_PM_r32(p,cast<uint32_t>((a + cast<uint32_t>(0ULL))));
r.esi = dos_PM_r32(p,cast<uint32_t>((a + cast<uint32_t>(4ULL))));
r.ebp = dos_PM_r32(p,cast<uint32_t>((a + cast<uint32_t>(8ULL))));
r.ebx = dos_PM_r32(p,cast<uint32_t>((a + cast<uint32_t>(16ULL))));
r.edx = dos_PM_r32(p,cast<uint32_t>((a + cast<uint32_t>(20ULL))));
r.ecx = dos_PM_r32(p,cast<uint32_t>((a + cast<uint32_t>(24ULL))));
r.eax = dos_PM_r32(p,cast<uint32_t>((a + cast<uint32_t>(28ULL))));
r.flags = cast<uint16_t>(dos_PM_r32(p,cast<uint32_t>((a + cast<uint32_t>(32ULL)))));
r.es = cast<uint16_t>(dos_PM_r32(p,cast<uint32_t>((a + cast<uint32_t>(34ULL)))));
r.ds = cast<uint16_t>(dos_PM_r32(p,cast<uint32_t>((a + cast<uint32_t>(36ULL)))));
r.fs = cast<uint16_t>(dos_PM_r32(p,cast<uint32_t>((a + cast<uint32_t>(38ULL)))));
r.gs = cast<uint16_t>(dos_PM_r32(p,cast<uint32_t>((a + cast<uint32_t>(40ULL)))));
r.ip = cast<uint16_t>(dos_PM_r32(p,cast<uint32_t>((a + cast<uint32_t>(42ULL)))));
r.cs = cast<uint16_t>(dos_PM_r32(p,cast<uint32_t>((a + cast<uint32_t>(44ULL)))));
r.sp = cast<uint16_t>(dos_PM_r32(p,cast<uint32_t>((a + cast<uint32_t>(46ULL)))));
r.ss = cast<uint16_t>(dos_PM_r32(p,cast<uint32_t>((a + cast<uint32_t>(48ULL)))));
return r;
}
}
// tools/platform/dos/go32_rmint.go:50:1
void dos_PM_writeBack(dos_PM* p,dos_rmcs* r){
{
dos_PM_w32(p,cast<uint32_t>((r->addr + cast<uint32_t>(0ULL))),r->edi);
dos_PM_w32(p,cast<uint32_t>((r->addr + cast<uint32_t>(4ULL))),r->esi);
dos_PM_w32(p,cast<uint32_t>((r->addr + cast<uint32_t>(8ULL))),r->ebp);
dos_PM_w32(p,cast<uint32_t>((r->addr + cast<uint32_t>(16ULL))),r->ebx);
dos_PM_w32(p,cast<uint32_t>((r->addr + cast<uint32_t>(20ULL))),r->edx);
dos_PM_w32(p,cast<uint32_t>((r->addr + cast<uint32_t>(24ULL))),r->ecx);
dos_PM_w32(p,cast<uint32_t>((r->addr + cast<uint32_t>(28ULL))),r->eax);
dos_PM_Write(p,cast<uint32_t>((r->addr + cast<uint32_t>(32ULL))),cast<uint8_t>(r->flags));
dos_PM_Write(p,cast<uint32_t>((r->addr + cast<uint32_t>(33ULL))),cast<uint8_t>(shr<uint16_t>(r->flags,cast<int64_t>(8ULL))));
dos_PM_w16(p,cast<uint32_t>((r->addr + cast<uint32_t>(34ULL))),r->es);
dos_PM_w16(p,cast<uint32_t>((r->addr + cast<uint32_t>(36ULL))),r->ds);
dos_PM_w16(p,cast<uint32_t>((r->addr + cast<uint32_t>(38ULL))),r->fs);
dos_PM_w16(p,cast<uint32_t>((r->addr + cast<uint32_t>(40ULL))),r->gs);
}
}
// tools/platform/dos/go32_rmint.go:68:1
bool dos_PM_simulateRealInt(dos_PM* p,x86_CPU* c){
{
uint8_t intno = x86_CPU_Reg8(c,cast<int64_t>(3ULL));
uint32_t rmAddr = cast<uint32_t>((c->SegBase[cast<int64_t>(0ULL)] + c->Regs[cast<int64_t>(7ULL)]));
dos_rmcs r = dos_PM_readRMCS(p,rmAddr);
r.flags &= ~(cast<uint16_t>(1ULL));
{
switch(intno){
case cast<uint8_t>(47ULL):{
r.eax &= cast<uint32_t>(4294967040ULL);
break;}
case cast<uint8_t>(33ULL):{
if ((!dos_PM_rmDOS(p,(&r)))) {
x86_CPU_Halt(c,std::string("DPMI 0300h INT 21h AH=%02X not yet modelled at %08X",51),cast<uint8_t>(shr<uint32_t>(r.eax,cast<int64_t>(8ULL))),c->IP);
return true;
}
break;}
case cast<uint8_t>(16ULL):case cast<uint8_t>(22ULL):case cast<uint8_t>(51ULL):{
break;}
default:{
x86_CPU_Halt(c,std::string("DPMI 0300h simulate-real-mode-INT %02Xh (AX=%04X) not yet modelled at %08X",74),intno,cast<uint16_t>(r.eax),c->IP);
return true;
break;}
}}
dos_PM_writeBack(p,(&r));
c->CF = false;
dos_PM_logf(p,std::string("DPMI 0300h: INT %02Xh serviced (AX=%04X)",40),intno,cast<uint16_t>(r.eax));
return true;
}
}
// tools/platform/dos/go32_rmint.go:98:1
bool dos_PM_rmDOS(dos_PM* p,dos_rmcs* r){
{
uint8_t ah = cast<uint8_t>(shr<uint32_t>(r->eax,cast<int64_t>(8ULL)));
p->DOSCounts[ah]++;
{
switch(ah){
case cast<uint8_t>(48ULL):{
r->eax = cast<uint32_t>(((cast<uint32_t>((r->eax & cast<uint32_t>(4294901760ULL)))) | cast<uint32_t>(7ULL)));
break;}
case cast<uint8_t>(26ULL):{
auto tmp148 = std::make_tuple(r->ds,cast<uint16_t>(r->edx));
p->dtaSeg = std::get<0>(tmp148);
p->dtaOff = std::get<1>(tmp148);
break;}
case cast<uint8_t>(37ULL):{
break;}
case cast<uint8_t>(53ULL):{
auto tmp149 = std::make_tuple(cast<uint16_t>(0ULL),cast<uint32_t>(0ULL));
r->es = std::get<0>(tmp149);
r->ebx = std::get<1>(tmp149);
break;}
case cast<uint8_t>(25ULL):{
r->eax = cast<uint32_t>(((cast<uint32_t>((r->eax & cast<uint32_t>(4294967040ULL)))) | cast<uint32_t>(2ULL)));
break;}
case cast<uint8_t>(82ULL):{
if (dos_PM_writeDOSStructures(p)) {
r->es = p->lolSeg;
r->ebx = cast<uint32_t>(((cast<uint32_t>((r->ebx & cast<uint32_t>(4294901760ULL)))) | cast<uint32_t>(p->lolOff)));
}
else {
auto tmp150 = std::make_tuple(cast<uint16_t>(0ULL),cast<uint32_t>(0ULL));
r->es = std::get<0>(tmp150);
r->ebx = std::get<1>(tmp150);
}
break;}
case cast<uint8_t>(51ULL):{
{
switch(cast<uint8_t>(r->eax)){
case cast<uint8_t>(0ULL):{
r->edx = cast<uint32_t>((r->edx & cast<uint32_t>(4294967040ULL)));
break;}
case cast<uint8_t>(5ULL):{
r->edx = cast<uint32_t>(((cast<uint32_t>((r->edx & cast<uint32_t>(4294967040ULL)))) | cast<uint32_t>(2ULL)));
break;}
}}
break;}
case cast<uint8_t>(60ULL):case cast<uint8_t>(61ULL):case cast<uint8_t>(62ULL):case cast<uint8_t>(63ULL):case cast<uint8_t>(64ULL):case cast<uint8_t>(65ULL):case cast<uint8_t>(66ULL):case cast<uint8_t>(67ULL):case cast<uint8_t>(68ULL):case cast<uint8_t>(71ULL):case cast<uint8_t>(78ULL):case cast<uint8_t>(79ULL):case cast<uint8_t>(87ULL):case cast<uint8_t>(96ULL):{
return dos_PM_dosFile(p,r);
break;}
default:{
return false;
break;}
}}
return true;
}
}
// tools/platform/dos/mz.go:50:1
int64_t dos_Reloc_FarAddr(dos_Reloc r){
{
return cast<int64_t>((cast<int64_t>((cast<int64_t>(r.Segment) * cast<int64_t>(16ULL))) + cast<int64_t>(r.Offset)));
}
}
// tools/platform/dos/mz.go:53:1
std::tuple<dos_MZ*,Error> dos_ParseMZ(Slice<uint8_t> data){
{
if ((len(data) < cast<int64_t>(28ULL))) {
return {{},go_errors_New(std::string("mz: file too small for a DOS header",35))};
}
if (((!(((data[cast<int64_t>(0ULL)] == cast<uint8_t>(77ULL)) && (data[cast<int64_t>(1ULL)] == cast<uint8_t>(90ULL))))) && (!(((data[cast<int64_t>(0ULL)] == cast<uint8_t>(90ULL)) && (data[cast<int64_t>(1ULL)] == cast<uint8_t>(77ULL))))))) {
return {{},go_fmt_Errorf(std::string("mz: bad signature %q, want \"MZ\"",31),cast<std::string>(sub(data,cast<int64_t>(0ULL),cast<int64_t>(2ULL))))};
}
auto u16 = [&](int64_t off)->uint16_t{
return le_Uint16(sub(data,off,len(data)));
}
;
dos_MZ* m = arenaNew(dos_MZ{u16(cast<int64_t>(2ULL)),u16(cast<int64_t>(4ULL)),u16(cast<int64_t>(6ULL)),u16(cast<int64_t>(8ULL)),u16(cast<int64_t>(10ULL)),u16(cast<int64_t>(12ULL)),u16(cast<int64_t>(14ULL)),u16(cast<int64_t>(16ULL)),u16(cast<int64_t>(18ULL)),u16(cast<int64_t>(20ULL)),u16(cast<int64_t>(22ULL)),u16(cast<int64_t>(24ULL)),u16(cast<int64_t>(26ULL)),len(data),{},{},{},{},{}});
if ((m->Pages > cast<uint16_t>(0ULL))) {
if ((m->LastPageBytes == cast<uint16_t>(0ULL))) {
m->LoadImageEnd = cast<int64_t>((cast<int64_t>(m->Pages) * cast<int64_t>(512ULL)));
}
else {
m->LoadImageEnd = cast<int64_t>((cast<int64_t>(((cast<int64_t>((cast<int64_t>(m->Pages) - cast<int64_t>(1ULL)))) * cast<int64_t>(512ULL))) + cast<int64_t>(m->LastPageBytes)));
}
}
m->LoadModuleOffset = cast<int64_t>((cast<int64_t>(m->HeaderParas) * cast<int64_t>(16ULL)));
m->LoadModuleSize = cast<int64_t>((m->LoadImageEnd - m->LoadModuleOffset));
m->AppendedSize = cast<int64_t>((m->FileSize - m->LoadImageEnd));
int64_t base = cast<int64_t>(m->RelocOffset);
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(m->Relocations));i++){
int64_t off = cast<int64_t>((base + cast<int64_t>((i * cast<int64_t>(4ULL)))));
if ((cast<int64_t>((off + cast<int64_t>(4ULL))) > len(data))) {
return {{},go_fmt_Errorf(std::string("mz: relocation %d at %#x runs past end of file",46),i,off)};
}
m->Relocs = append(m->Relocs,Slice<dos_Reloc>{dos_Reloc{le_Uint16(sub(data,cast<int64_t>((off + cast<int64_t>(2ULL))),len(data))),le_Uint16(sub(data,off,len(data)))}});
}
}return {m,{}};
}
}
// tools/platform/dos/mz.go:109:1
int64_t dos_MZ_EntryLinear(dos_MZ* m){
{
return cast<int64_t>((cast<int64_t>((cast<int64_t>(m->InitCS) * cast<int64_t>(16ULL))) + cast<int64_t>(m->InitIP)));
}
}
// tools/platform/dos/mz.go:112:1
int64_t dos_MZ_StackLinear(dos_MZ* m){
{
return cast<int64_t>((cast<int64_t>((cast<int64_t>(m->InitSS) * cast<int64_t>(16ULL))) + cast<int64_t>(m->InitSP)));
}
}

#include "fast.h"
