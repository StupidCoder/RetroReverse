#include "runtime.h"
struct arm_Inst;
struct arm_CPU;
struct arm_Banks;
struct arm_vfpState;
struct gba_Header;
struct gba_ROM;
struct gbamachine_bus;
struct gbamachine_dmaChan;
struct gbamachine_eeprom;
struct gbamachine_Machine;
struct gbamachine_video;
struct gbamachine_ppu;
struct gbamachine_lineBuf;
struct gbamachine_Result;
struct gbamachine_square;
struct gbamachine_waveCh;
struct gbamachine_noiseCh;
struct gbamachine_fifo;
struct gbamachine_apu;
struct gbamachine_timer;
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
struct Anon22;
struct Anon23;
struct Anon3;
struct Anon4;
struct Anon5;
struct Anon6;
struct Anon7;
struct Anon8;
struct Anon9;
using arm_Flow=int64_t;
struct arm_Inst{
uint32_t Addr{};
int64_t Len{};
std::string Mnem{};
std::string Text{};
arm_Flow Flow{};
uint32_t Target{};
bool HasTarget{};
bool Thumb{};
bool TargetThumb{};
int64_t Cond{};
};
using arm_Variant=int64_t;
struct arm_vfpState{
std::array<uint32_t,32> S{};
uint32_t FPSCR{};
uint32_t FPEXC{};
};
struct arm_CPU{
std::array<uint32_t,16> R{};
bool N{};
bool Z{};
bool C{};
bool V{};
bool Q{};
uint32_t GE{};
bool Thumb{};
bool BigEndian{};
bool IRQDisable{};
bool FIQDisable{};
uint32_t Mode{};
arm_Variant Arch{};
arm_vfpState VFP{};
bool exclValid{};
uint32_t exclAddr{};
std::array<uint32_t,6> bankR13{};
std::array<uint32_t,6> bankR14{};
std::array<uint32_t,6> bankSPSR{};
std::array<uint32_t,5> fiqR8_12{};
std::array<uint32_t,5> usrR8_12{};
std::function<bool(arm_CPU*,uint32_t)> SWI{};
std::function<void(arm_CPU*,bool,uint32_t,uint32_t,uint32_t,uint32_t,uint32_t,uint32_t*)> Coproc{};
gbamachine_bus* bus{};
gbamachine_bus* wide{};
bool Halted{};
std::string HaltReason{};
uint64_t Instrs{};
uint32_t cur{};
bool branched{};
};
struct arm_Banks{
std::array<uint32_t,6> R13{};
std::array<uint32_t,6> R14{};
std::array<uint32_t,6> SPSR{};
std::array<uint32_t,5> FIQR8_12{};
std::array<uint32_t,5> USRR8_12{};
};
using arm_parMode=int64_t;
struct gba_Header{
uint32_t EntryInstr{};
uint32_t EntryAddr{};
std::string LogoMD5{};
std::string Title{};
std::string GameCode{};
std::string MakerCode{};
uint8_t Fixed96{};
uint8_t UnitCode{};
uint8_t DeviceType{};
uint8_t Version{};
uint8_t Checksum{};
bool ChecksumOK{};
};
struct gba_ROM{
Slice<uint8_t> Data{};
gba_Header Header{};
};
struct gbamachine_bus{
gbamachine_Machine* m{};
};
struct gbamachine_dmaChan{
uint32_t src{};
uint32_t dst{};
uint32_t latchSrc{};
uint32_t latchDst{};
uint16_t count{};
uint16_t ctrl{};
};
struct gbamachine_eeprom{
bool present{};
Slice<uint8_t> data{};
bool sized{};
Slice<uint8_t> inBits{};
Slice<uint8_t> outBits{};
bool ready{};
};
struct gbamachine_video{
int64_t line{};
bool hblank{};
uint64_t frames{};
};
struct gbamachine_ppu{
int32_t bg2x{};
int32_t bg2y{};
int32_t bg3x{};
int32_t bg3y{};
};
struct gbamachine_timer{
uint16_t counter{};
uint16_t reload{};
uint16_t ctrl{};
int64_t frac{};
};
struct gbamachine_Machine{
arm_CPU* cpu{};
Slice<uint8_t> rom{};
Slice<uint8_t> ewram{};
Slice<uint8_t> iwram{};
Slice<uint8_t> pal{};
Slice<uint8_t> vram{};
Slice<uint8_t> oam{};
Map<uint32_t,uint16_t> io{};
bool ime{};
uint16_t ie{};
uint16_t if_{};
bool waiting{};
bool waitAny{};
uint16_t waitMask{};
gbamachine_video vid{};
gbamachine_ppu ppu{};
gbamachine_apu* apu{};
std::array<gbamachine_dmaChan,4> dma{};
std::array<gbamachine_timer,4> timers{};
gbamachine_eeprom eeprom{};
uint16_t keys{};
uint64_t Steps{};
std::array<uint32_t,38400> screen{};
Slice<std::string> Log{};
Map<std::string,bool> logSeen{};
Map<uint32_t,bool> visited{};
std::function<void(uint32_t)> OnStep{};
std::function<void(bool,uint32_t,uint16_t,uint32_t)> OnIO{};
std::function<void(uint16_t,uint32_t,uint32_t)> OnIRQ{};
std::function<void()> OnFrame{};
std::function<void(uint32_t,uint8_t,uint32_t)> OnRead{};
std::function<void(uint32_t,uint8_t,uint32_t)> OnWrite{};
std::function<void(uint32_t,uint32_t,uint32_t,int64_t,uint32_t,uint32_t)> OnDecompress{};
Map<uint32_t,bool> bps{};
bool stop{};
bool stopped{};
uint32_t stoppedPC{};
};
struct gbamachine_lineBuf{
std::array<uint16_t,240> c{};
std::array<bool,240> on{};
};
struct gbamachine_Result{
uint64_t Steps{};
uint64_t Frames{};
std::string Reason{};
Map<uint32_t,uint64_t> Milestone{};
};
struct gbamachine_square{
bool enabled{};
bool hasSweep{};
double phase{};
int64_t dutySel{};
int64_t freq{};
int64_t length{};
bool lengthEn{};
int64_t vol{};
int64_t envDir{};
int64_t envPeriod{};
int64_t envTimer{};
int64_t swPeriod{};
int64_t swShift{};
int64_t swDir{};
int64_t swTimer{};
bool swOn{};
int64_t swShadow{};
};
struct gbamachine_waveCh{
bool enabled{};
bool dacOn{};
double phase{};
int64_t freq{};
int64_t length{};
bool lengthEn{};
int64_t volShift{};
bool force75{};
std::array<std::array<uint8_t,16>,2> ram{};
int64_t bank{};
bool twoBanks{};
};
struct gbamachine_noiseCh{
bool enabled{};
uint16_t lfsr{};
double timer{};
int64_t length{};
bool lengthEn{};
int64_t vol{};
int64_t envDir{};
int64_t envPeriod{};
int64_t envTimer{};
int64_t divisor{};
int64_t shift{};
bool width7{};
};
struct gbamachine_fifo{
Slice<int8_t> q{};
int8_t cur{};
};
struct gbamachine_apu{
gbamachine_square ch1{};
gbamachine_square ch2{};
gbamachine_waveCh ch3{};
gbamachine_noiseCh ch4{};
gbamachine_fifo dsA{};
gbamachine_fifo dsB{};
bool powered{};
double frameAcc{};
int64_t frameSeq{};
uint16_t mixRegL{};
uint16_t mixRegH{};
double sampleAcc{};
Slice<int16_t> PCM{};
bool Capture{};
};
struct Anon0{uint32_t Addr{};int64_t Len{};std::string Mnem{};std::string Text{};arm_Flow Flow{};uint32_t Target{};bool HasTarget{};bool Thumb{};bool TargetThumb{};int64_t Cond{};};
struct Anon1{std::array<uint32_t,16> R{};bool N{};bool Z{};bool C{};bool V{};bool Q{};uint32_t GE{};bool Thumb{};bool BigEndian{};bool IRQDisable{};bool FIQDisable{};uint32_t Mode{};arm_Variant Arch{};arm_vfpState VFP{};bool exclValid{};uint32_t exclAddr{};std::array<uint32_t,6> bankR13{};std::array<uint32_t,6> bankR14{};std::array<uint32_t,6> bankSPSR{};std::array<uint32_t,5> fiqR8_12{};std::array<uint32_t,5> usrR8_12{};std::function<bool(arm_CPU*,uint32_t)> SWI{};std::function<void(arm_CPU*,bool,uint32_t,uint32_t,uint32_t,uint32_t,uint32_t,uint32_t*)> Coproc{};gbamachine_bus* bus{};gbamachine_bus* wide{};bool Halted{};std::string HaltReason{};uint64_t Instrs{};uint32_t cur{};bool branched{};};
struct Anon10{bool present{};Slice<uint8_t> data{};bool sized{};Slice<uint8_t> inBits{};Slice<uint8_t> outBits{};bool ready{};};
struct Anon11{arm_CPU* cpu{};Slice<uint8_t> rom{};Slice<uint8_t> ewram{};Slice<uint8_t> iwram{};Slice<uint8_t> pal{};Slice<uint8_t> vram{};Slice<uint8_t> oam{};Map<uint32_t,uint16_t> io{};bool ime{};uint16_t ie{};uint16_t if_{};bool waiting{};bool waitAny{};uint16_t waitMask{};gbamachine_video vid{};gbamachine_ppu ppu{};gbamachine_apu* apu{};std::array<gbamachine_dmaChan,4> dma{};std::array<gbamachine_timer,4> timers{};gbamachine_eeprom eeprom{};uint16_t keys{};uint64_t Steps{};std::array<uint32_t,38400> screen{};Slice<std::string> Log{};Map<std::string,bool> logSeen{};Map<uint32_t,bool> visited{};std::function<void(uint32_t)> OnStep{};std::function<void(bool,uint32_t,uint16_t,uint32_t)> OnIO{};std::function<void(uint16_t,uint32_t,uint32_t)> OnIRQ{};std::function<void()> OnFrame{};std::function<void(uint32_t,uint8_t,uint32_t)> OnRead{};std::function<void(uint32_t,uint8_t,uint32_t)> OnWrite{};std::function<void(uint32_t,uint32_t,uint32_t,int64_t,uint32_t,uint32_t)> OnDecompress{};Map<uint32_t,bool> bps{};bool stop{};bool stopped{};uint32_t stoppedPC{};};
struct Anon12{int64_t line{};bool hblank{};uint64_t frames{};};
struct Anon13{int32_t bg2x{};int32_t bg2y{};int32_t bg3x{};int32_t bg3y{};};
struct Anon14{std::array<uint16_t,240> c{};std::array<bool,240> on{};};
struct Anon15{uint64_t Steps{};uint64_t Frames{};std::string Reason{};Map<uint32_t,uint64_t> Milestone{};};
struct Anon16{bool enabled{};bool hasSweep{};double phase{};int64_t dutySel{};int64_t freq{};int64_t length{};bool lengthEn{};int64_t vol{};int64_t envDir{};int64_t envPeriod{};int64_t envTimer{};int64_t swPeriod{};int64_t swShift{};int64_t swDir{};int64_t swTimer{};bool swOn{};int64_t swShadow{};};
struct Anon17{bool enabled{};bool dacOn{};double phase{};int64_t freq{};int64_t length{};bool lengthEn{};int64_t volShift{};bool force75{};std::array<std::array<uint8_t,16>,2> ram{};int64_t bank{};bool twoBanks{};};
struct Anon18{bool enabled{};uint16_t lfsr{};double timer{};int64_t length{};bool lengthEn{};int64_t vol{};int64_t envDir{};int64_t envPeriod{};int64_t envTimer{};int64_t divisor{};int64_t shift{};bool width7{};};
struct Anon19{Slice<int8_t> q{};int8_t cur{};};
struct Anon2{std::array<uint32_t,6> R13{};std::array<uint32_t,6> R14{};std::array<uint32_t,6> SPSR{};std::array<uint32_t,5> FIQR8_12{};std::array<uint32_t,5> USRR8_12{};};
struct Anon20{gbamachine_square ch1{};gbamachine_square ch2{};gbamachine_waveCh ch3{};gbamachine_noiseCh ch4{};gbamachine_fifo dsA{};gbamachine_fifo dsB{};bool powered{};double frameAcc{};int64_t frameSeq{};uint16_t mixRegL{};uint16_t mixRegH{};double sampleAcc{};Slice<int16_t> PCM{};bool Capture{};};
struct Anon21{gbamachine_fifo* f{};int64_t timerSel{};int64_t dmaCh{};};
struct Anon22{std::string name{};gbamachine_fifo* f{};int64_t timerSel{};int64_t dmaCh{};};
struct Anon23{uint16_t counter{};uint16_t reload{};uint16_t ctrl{};int64_t frac{};};
struct Anon3{uint32_t bit{};std::string name{};};
struct Anon4{std::array<uint32_t,32> S{};uint32_t FPSCR{};uint32_t FPEXC{};};
struct Anon5{uint32_t EntryInstr{};uint32_t EntryAddr{};std::string LogoMD5{};std::string Title{};std::string GameCode{};std::string MakerCode{};uint8_t Fixed96{};uint8_t UnitCode{};uint8_t DeviceType{};uint8_t Version{};uint8_t Checksum{};bool ChecksumOK{};};
struct Anon6{Slice<uint8_t> Data{};gba_Header Header{};};
struct Anon7{std::string id{};std::string desc{};};
struct Anon8{gbamachine_Machine* m{};};
struct Anon9{uint32_t src{};uint32_t dst{};uint32_t latchSrc{};uint32_t latchDst{};uint16_t count{};uint16_t ctrl{};};

#include "adapters-decl.h"
uint32_t arm_signExtend(uint32_t v,uint64_t n);
std::string arm_imm(uint32_t v);
arm_Inst arm_Decode(Slice<uint8_t> code,uint32_t addr,bool thumb);
std::tuple<uint32_t,bool> arm_word(Slice<uint8_t> code);
arm_Inst arm_DecodeARM(Slice<uint8_t> code,uint32_t addr);
std::string arm_cn(int64_t cond);
arm_Inst arm_undef(uint32_t w,arm_Inst in);
arm_Inst arm_decodeDataMisc(uint32_t w,uint32_t addr,arm_Inst in,bool immForm);
arm_Inst arm_decodeDataProc(uint32_t w,uint32_t addr,arm_Inst in,bool immForm);
void arm_classifyPCWrite(arm_Inst* in,uint32_t w,uint32_t op);
std::string arm_shiftOperand(uint32_t w);
uint32_t arm_ror32(uint32_t v,uint32_t n);
arm_Inst arm_decodeMisc(uint32_t w,arm_Inst in,bool immForm);
arm_Inst arm_decodeMSRreg(uint32_t w,arm_Inst in);
arm_Inst arm_decodeMSRimm(uint32_t w,arm_Inst in);
std::string arm_msrFields(uint32_t w);
arm_Inst arm_decodeSignedMul(uint32_t w,arm_Inst in);
arm_Inst arm_decodeExtension(uint32_t w,arm_Inst in);
arm_Inst arm_decodeMul(uint32_t w,arm_Inst in);
arm_Inst arm_decodeMulLong(uint32_t w,arm_Inst in);
arm_Inst arm_decodeHalf(uint32_t w,arm_Inst in,uint32_t sh);
std::string arm_plusMinus(std::string s);
arm_Inst arm_decodeSingle(uint32_t w,arm_Inst in,bool regOff);
std::string arm_addrForm(std::string mnem,uint32_t rd,uint32_t rn,std::string off,uint32_t p,uint32_t wb);
arm_Inst arm_decodeBlock(uint32_t w,arm_Inst in);
std::string arm_regList(uint32_t mask);
std::string arm_joinComma(Slice<std::string> parts);
arm_Inst arm_decodeBranch(uint32_t w,uint32_t addr,arm_Inst in);
arm_Inst arm_decodeUncond(uint32_t w,uint32_t addr,arm_Inst in);
arm_Inst arm_decodeCoproLS(uint32_t w,arm_Inst in);
arm_Inst arm_decodeCopro(uint32_t w,arm_Inst in);
int64_t arm_CPU_Step(arm_CPU* c);
uint32_t arm_CPU_boolToU(arm_CPU* c,bool b);
int64_t arm_CPU_stepARM(arm_CPU* c);
int64_t arm_CPU_execUncondARM(arm_CPU* c,uint32_t w);
void arm_CPU_execARM(arm_CPU* c,uint32_t w);
void arm_CPU_execDataMisc(arm_CPU* c,uint32_t w,bool immForm);
std::tuple<uint32_t,uint32_t> arm_CPU_dpOperand(arm_CPU* c,uint32_t w,bool immForm);
void arm_CPU_execDataProc(arm_CPU* c,uint32_t w,bool immForm);
void arm_CPU_branchTo(arm_CPU* c,uint32_t v);
void arm_CPU_bxTo(arm_CPU* c,uint32_t v);
void arm_CPU_execBranch(arm_CPU* c,uint32_t w);
void arm_CPU_execExtension(arm_CPU* c,uint32_t w);
void arm_CPU_execMul(arm_CPU* c,uint32_t w);
void arm_CPU_execMulLong(arm_CPU* c,uint32_t w);
void arm_CPU_execSwap(arm_CPU* c,uint32_t w);
void arm_CPU_execHalf(arm_CPU* c,uint32_t w,uint32_t sh);
void arm_CPU_execSingle(arm_CPU* c,uint32_t w,bool regOff);
void arm_CPU_execBlock(arm_CPU* c,uint32_t w);
bool arm_CPU_inUserBank(arm_CPU* c);
uint32_t arm_CPU_userReg(arm_CPU* c,uint32_t i);
void arm_CPU_setUserReg(arm_CPU* c,uint32_t i,uint32_t v);
void arm_CPU_Reset(arm_CPU* c);
uint32_t arm_CPU_PC(arm_CPU* c);
uint32_t arm_CPU_CPSR(arm_CPU* c);
void arm_CPU_SetCPSR(arm_CPU* c,uint32_t v);
int64_t arm_modeIndex(uint32_t mode);
void arm_CPU_switchMode(arm_CPU* c,uint32_t mode);
uint32_t arm_CPU_SPSR(arm_CPU* c);
void arm_CPU_SetSPSR(arm_CPU* c,uint32_t v);
uint8_t arm_CPU_read8(arm_CPU* c,uint32_t a);
void arm_CPU_write8(arm_CPU* c,uint32_t a,uint8_t v);
uint32_t arm_CPU_read32(arm_CPU* c,uint32_t a);
void arm_CPU_write32(arm_CPU* c,uint32_t a,uint32_t v);
uint32_t arm_CPU_reg(arm_CPU* c,uint32_t i);
void arm_CPU_setReg(arm_CPU* c,uint32_t i,uint32_t v);
void arm_CPU_setNZ(arm_CPU* c,uint32_t v);
uint32_t arm_CPU_add(arm_CPU* c,uint32_t a,uint32_t b,uint32_t cin);
uint32_t arm_CPU_sub_(arm_CPU* c,uint32_t a,uint32_t b,uint32_t cin);
bool arm_CPU_cond(arm_CPU* c,int64_t cc);
std::tuple<uint32_t,uint32_t> arm_CPU_shift(arm_CPU* c,uint32_t typ,uint32_t amt,uint32_t val,bool regForm,uint32_t cin);
void arm_CPU_ClearExclusive(arm_CPU* c);
arm_Banks arm_CPU_SaveBanks(arm_CPU* c);
void arm_CPU_RestoreBanks(arm_CPU* c,arm_Banks b);
std::tuple<arm_Inst,bool> arm_decodeARMv6(uint32_t w,uint32_t addr,arm_Inst in);
arm_Inst arm_decodeSync(uint32_t w,arm_Inst in);
arm_Inst arm_decodeMedia(uint32_t w,arm_Inst in);
std::string arm_parallelName(uint32_t class_,uint32_t op2);
arm_Inst arm_decodeMediaPack(uint32_t w,arm_Inst in);
arm_Inst arm_satText(uint32_t w,arm_Inst in,std::string name,uint32_t sat);
std::string arm_extendName(uint32_t op1);
arm_Inst arm_decodeMediaMul(uint32_t w,arm_Inst in);
std::tuple<arm_Inst,bool> arm_decodeV6Uncond(uint32_t w,arm_Inst in);
bool arm_CPU_execARMv6(arm_CPU* c,uint32_t w);
bool arm_CPU_execUncondARMv6(arm_CPU* c,uint32_t w);
void arm_CPU_execSync(arm_CPU* c,uint32_t w);
void arm_CPU_execUMAAL(arm_CPU* c,uint32_t w);
bool arm_CPU_execMedia(arm_CPU* c,uint32_t w);
bool arm_CPU_execParallel(arm_CPU* c,uint32_t w);
std::tuple<uint32_t,bool> arm_parLane(uint32_t av,uint32_t bv,uint64_t width,bool signed_,arm_parMode mode,bool sub_);
int32_t arm_int16v(uint32_t v);
int32_t arm_signExtendLane(uint32_t v,uint64_t width);
bool arm_CPU_execMediaPack(arm_CPU* c,uint32_t w);
uint32_t arm_CPU_doSat(arm_CPU* c,uint32_t w,bool signed_);
bool arm_CPU_execExtend(arm_CPU* c,uint32_t w);
bool arm_CPU_execMediaMul(arm_CPU* c,uint32_t w);
void arm_CPU_execUSAD(arm_CPU* c,uint32_t w);
void arm_CPU_execMisc(arm_CPU* c,uint32_t w,bool immForm);
void arm_CPU_undefV4T(arm_CPU* c,uint32_t w,std::string name);
void arm_CPU_execMSR(arm_CPU* c,uint32_t w,bool immForm);
uint32_t arm_CPU_qsat(arm_CPU* c,int64_t v);
void arm_CPU_execSaturating(arm_CPU* c,uint32_t w);
int64_t arm_half16(uint32_t v,uint32_t t);
void arm_CPU_execSignedMul(arm_CPU* c,uint32_t w);
void arm_CPU_execSWI(arm_CPU* c,uint32_t w);
void arm_CPU_exception(arm_CPU* c,uint32_t mode,uint32_t vector,uint32_t returnAddr);
void arm_CPU_Exception(arm_CPU* c,uint32_t mode,uint32_t vector,uint32_t returnAddr);
bool arm_CPU_IRQ(arm_CPU* c);
void arm_CPU_execCopro(arm_CPU* c,uint32_t w);
arm_Inst arm_DecodeThumb(Slice<uint8_t> code,uint32_t addr);
std::string arm_lo(uint32_t v);
arm_Inst arm_thumbAddSub(uint32_t h,arm_Inst in);
arm_Inst arm_thumbALU(uint32_t h,arm_Inst in);
arm_Inst arm_thumbHiReg(uint32_t h,arm_Inst in);
arm_Inst arm_thumbLoadStoreReg(uint32_t h,arm_Inst in);
arm_Inst arm_thumbLoadStoreSExt(uint32_t h,arm_Inst in);
arm_Inst arm_thumbLoadStoreImm(uint32_t h,arm_Inst in);
arm_Inst arm_thumbPushPop(uint32_t h,arm_Inst in);
arm_Inst arm_thumbCondBranch(uint32_t h,uint32_t addr,arm_Inst in);
arm_Inst arm_thumbLongBranch(Slice<uint8_t> code,uint32_t h,uint32_t addr,arm_Inst in);
int64_t arm_CPU_stepThumb(arm_CPU* c);
void arm_CPU_execThumb(arm_CPU* c,uint32_t h);
void arm_CPU_thumbAddSub(arm_CPU* c,uint32_t h);
void arm_CPU_thumbImm(arm_CPU* c,uint32_t h);
void arm_CPU_thumbALUExec(arm_CPU* c,uint32_t h);
void arm_CPU_thumbHiRegExec(arm_CPU* c,uint32_t h);
void arm_CPU_thumbLoadStoreReg(arm_CPU* c,uint32_t h);
void arm_CPU_thumbLoadStoreSExt(arm_CPU* c,uint32_t h);
void arm_CPU_thumbLoadStoreImm(arm_CPU* c,uint32_t h);
void arm_CPU_thumbPushPop(arm_CPU* c,uint32_t h);
void arm_CPU_thumbBlock(arm_CPU* c,uint32_t h);
void arm_CPU_thumbCondBranch(arm_CPU* c,uint32_t h);
std::string arm_Variant_String(arm_Variant v);
bool arm_Variant_isV6(arm_Variant v);
bool arm_Variant_v5OrLater(arm_Variant v);
arm_Inst arm_DecodeVariant(Slice<uint8_t> code,uint32_t addr,bool thumb,arm_Variant v);
arm_Inst arm_DecodeARMVariant(Slice<uint8_t> code,uint32_t addr,arm_Variant v);
arm_Inst arm_v4tFilter(arm_Inst in);
float arm_CPU_sGet(arm_CPU* c,uint32_t n);
void arm_CPU_sSet(arm_CPU* c,uint32_t n,float v);
uint32_t arm_CPU_sBits(arm_CPU* c,uint32_t n);
void arm_CPU_sSetBits(arm_CPU* c,uint32_t n,uint32_t v);
double arm_CPU_dGet(arm_CPU* c,uint32_t n);
void arm_CPU_dSet(arm_CPU* c,uint32_t n,double v);
uint32_t arm_sReg(uint32_t v,uint32_t bit);
uint32_t arm_dReg(uint32_t v,uint32_t bit);
bool arm_isVFP(uint32_t w);
arm_Inst arm_decodeVFP(uint32_t w,arm_Inst in);
arm_Inst arm_decodeVFPMove(uint32_t w,arm_Inst in);
arm_Inst arm_decodeVFPData(uint32_t w,arm_Inst in,bool single,std::string prec,std::string sfx);
arm_Inst arm_decodeVFPExt(uint32_t w,arm_Inst in,bool single,std::string prec,std::string sfx,std::string vd,std::string vm);
std::string arm_vfpRegName(bool single,uint32_t v,uint32_t bit);
std::string arm_wbMark(uint32_t wb);
bool arm_CPU_execVFP(arm_CPU* c,uint32_t w);
void arm_CPU_execVFPLoadStore(arm_CPU* c,uint32_t w);
bool arm_CPU_execVFPData(arm_CPU* c,uint32_t w);
void arm_CPU_vfpNegate(arm_CPU* c,uint32_t vd,bool single);
void arm_CPU_vfpMulAcc(arm_CPU* c,uint32_t vd,uint32_t vn,uint32_t vm,bool single,bool sub_,bool negAcc);
bool arm_CPU_execVFPExt(arm_CPU* c,uint32_t w,bool single,uint32_t vd,uint32_t vm);
void arm_CPU_vfpCopy(arm_CPU* c,uint32_t vd,uint32_t vm,bool single);
void arm_CPU_vfpUnary(arm_CPU* c,uint32_t vd,uint32_t vm,bool single,std::function<double(double)> f);
void arm_CPU_execVCMP(arm_CPU* c,uint32_t w,bool single,uint32_t vd,uint32_t vm,bool withZero);
uint32_t arm_vdOf(uint32_t w,bool single);
uint32_t arm_vmOf(uint32_t w,bool single);
void arm_CPU_execVCVTPrec(arm_CPU* c,uint32_t w,bool srcSingle);
void arm_CPU_execVCVTFromInt(arm_CPU* c,uint32_t w,bool single);
void arm_CPU_execVCVTToInt(arm_CPU* c,uint32_t w,bool single);
bool arm_CPU_execVFPMove(arm_CPU* c,uint32_t w);
bool arm_CPU_execVFPMove64(arm_CPU* c,uint32_t w);
uint64_t arm_vfpExpandImm(uint32_t w,bool single);
uint32_t arm_repeat(uint32_t bit,uint32_t n);
uint64_t arm_repeat64(uint64_t bit,uint32_t n);
uint8_t gba_ComplementCheck(Slice<uint8_t> data);
std::tuple<std::string,std::string> gba_ROM_SaveType(gba_ROM* r);
std::function<bool(arm_CPU*,uint32_t)> gbamachine_biosSWI(gbamachine_Machine* m);
void gbamachine_Machine_noteDecompress(gbamachine_Machine* m,gbamachine_bus* b,uint32_t swi,uint32_t src,uint32_t dst,uint32_t pc);
void gbamachine_Machine_intrWait(gbamachine_Machine* m,gbamachine_bus* b,bool discard,uint16_t mask);
void gbamachine_Machine_bgAffineSet(gbamachine_Machine* m,gbamachine_bus* b,uint32_t src,uint32_t dst,uint32_t num);
void gbamachine_Machine_objAffineSet(gbamachine_Machine* m,gbamachine_bus* b,uint32_t src,uint32_t dst,uint32_t num,uint32_t stride);
void gbamachine_Machine_bitUnPack(gbamachine_Machine* m,gbamachine_bus* b,uint32_t src,uint32_t dst,uint32_t params);
std::tuple<uint32_t,bool> gbamachine_Machine_lzHeader(gbamachine_Machine* m,gbamachine_bus* b,uint32_t src,uint32_t wantType,std::string name);
void gbamachine_Machine_lz77(gbamachine_Machine* m,gbamachine_bus* b,uint32_t src,uint32_t dst,bool vram);
void gbamachine_Machine_rle(gbamachine_Machine* m,gbamachine_bus* b,uint32_t src,uint32_t dst,bool vram);
void gbamachine_Machine_huffman(gbamachine_Machine* m,gbamachine_bus* b,uint32_t src,uint32_t dst);
std::tuple<Slice<uint8_t>,uint32_t,bool> gbamachine_bus_region(gbamachine_bus* b,uint32_t a);
std::tuple<uint8_t,bool> gbamachine_bus_romByte(gbamachine_bus* b,uint32_t a);
uint8_t gbamachine_bus_Read(gbamachine_bus* b,uint32_t a);
void gbamachine_bus_Write(gbamachine_bus* b,uint32_t a,uint8_t v);
uint16_t gbamachine_bus_Read16(gbamachine_bus* b,uint32_t a);
uint16_t gbamachine_bus_Read16_reference(gbamachine_bus* b,uint32_t a);
void gbamachine_bus_Write16(gbamachine_bus* b,uint32_t a,uint16_t v);
uint32_t gbamachine_bus_Read32(gbamachine_bus* b,uint32_t a);
void gbamachine_bus_Write32(gbamachine_bus* b,uint32_t a,uint32_t v);
uint32_t gbamachine_bus_r32(gbamachine_bus* b,uint32_t a);
void gbamachine_bus_w32(gbamachine_bus* b,uint32_t a,uint32_t v);
uint16_t gbamachine_bus_r16(gbamachine_bus* b,uint32_t a);
void gbamachine_bus_w16(gbamachine_bus* b,uint32_t a,uint16_t v);
void gbamachine_Machine_dmaRegWrite(gbamachine_Machine* m,uint32_t reg,uint16_t v);
void gbamachine_Machine_dmaSoundRefill(gbamachine_Machine* m,int64_t n);
void gbamachine_Machine_dmaTrigger(gbamachine_Machine* m,uint16_t timing);
void gbamachine_Machine_dmaRun(gbamachine_Machine* m,int64_t n);
void gbamachine_eeprom_init(gbamachine_eeprom* e);
void gbamachine_eeprom_write(gbamachine_eeprom* e,uint16_t v);
uint16_t gbamachine_eeprom_read(gbamachine_eeprom* e);
void gbamachine_eeprom_endFrame(gbamachine_eeprom* e,gbamachine_Machine* m);
void gbamachine_eeprom_size(gbamachine_eeprom* e,bool large);
uint16_t gbamachine_Machine_ioRead16(gbamachine_Machine* m,uint32_t a);
void gbamachine_Machine_ioWrite8(gbamachine_Machine* m,uint32_t a,uint8_t v);
void gbamachine_Machine_ioWrite16(gbamachine_Machine* m,uint32_t a,uint16_t v);
gbamachine_Machine* gbamachine_New(gba_ROM* rom);
uint32_t gbamachine_Machine_PC(gbamachine_Machine* m);
std::array<uint32_t,16> gbamachine_Machine_Regs(gbamachine_Machine* m);
bool gbamachine_Machine_ThumbState(gbamachine_Machine* m);
std::tuple<uint16_t,uint16_t,bool> gbamachine_Machine_IRQState(gbamachine_Machine* m);
bool gbamachine_Machine_IRQDisabled(gbamachine_Machine* m);
std::string gbamachine_Machine_Parked(gbamachine_Machine* m);
uint64_t gbamachine_Machine_Frame(gbamachine_Machine* m);
int64_t gbamachine_Machine_Line(gbamachine_Machine* m);
std::tuple<bool,std::string> gbamachine_Machine_Halted(gbamachine_Machine* m);
uint64_t gbamachine_Machine_Instrs(gbamachine_Machine* m);
uint16_t gbamachine_Machine_Reg(gbamachine_Machine* m,uint32_t a);
Slice<uint8_t> gbamachine_Machine_Snapshot(gbamachine_Machine* m,uint32_t addr,uint32_t n);
void gbamachine_Machine_Poke(gbamachine_Machine* m,uint32_t addr,Slice<uint8_t> data);
void gbamachine_Machine_SetKeys(gbamachine_Machine* m,uint16_t mask);
void gbamachine_Machine_AddBreakpoint(gbamachine_Machine* m,uint32_t pc);
std::string gbamachine_Machine_GameCode(gbamachine_Machine* m);
uint32_t gbamachine_ppu_objVRAMBase(gbamachine_ppu* p,gbamachine_Machine* m);
void gbamachine_ppu_reloadAffineRef(gbamachine_ppu* p,gbamachine_Machine* m,uint32_t reg);
void gbamachine_ppu_startFrame(gbamachine_ppu* p,gbamachine_Machine* m);
void gbamachine_ppu_stepAffine(gbamachine_ppu* p,gbamachine_Machine* m);
uint32_t gbamachine_rgb15(uint16_t c);
uint16_t gbamachine_Machine_pal16(gbamachine_Machine* m,int64_t bank,int64_t idx);
void gbamachine_ppu_renderLine(gbamachine_ppu* p,gbamachine_Machine* m,int64_t y);
bool gbamachine_ppu_inWindow(gbamachine_ppu* p,gbamachine_Machine* m,uint32_t hreg,uint32_t vreg,int64_t x,int64_t y);
uint16_t gbamachine_blend(uint16_t a,uint16_t b,int64_t eva,int64_t evb);
uint16_t gbamachine_brighten(uint16_t c,int64_t evy);
uint16_t gbamachine_darken(uint16_t c,int64_t evy);
void gbamachine_ppu_textLine(gbamachine_ppu* p,gbamachine_Machine* m,int64_t n,int64_t y,gbamachine_lineBuf* out);
void gbamachine_ppu_affineLine(gbamachine_ppu* p,gbamachine_Machine* m,int64_t n,gbamachine_lineBuf* out);
void gbamachine_ppu_bitmapLine(gbamachine_ppu* p,gbamachine_Machine* m,int64_t mode,gbamachine_lineBuf* out);
void gbamachine_ppu_objLine(gbamachine_ppu* p,gbamachine_Machine* m,int64_t y,uint16_t dispcnt,Slice<uint16_t> c,Slice<bool> on,Slice<uint8_t> prio,Slice<bool> semi,Slice<bool> win);
void gbamachine_ppu_objPixel(gbamachine_ppu* p,gbamachine_Machine* m,int64_t x,int64_t idx,int64_t palBank,uint8_t pr,uint16_t mode,Slice<uint16_t> c,Slice<bool> on,Slice<uint8_t> prio,Slice<bool> semi,Slice<bool> win);
gbamachine_Result gbamachine_Machine_Run(gbamachine_Machine* m,uint64_t budget,Map<uint32_t,std::string> milestones);
gbamachine_Result gbamachine_Machine_RunFrames(gbamachine_Machine* m,uint64_t n,uint64_t budget);
gbamachine_Result gbamachine_Machine_run(gbamachine_Machine* m,uint64_t budget,Map<uint32_t,std::string> milestones,uint64_t untilFrame);
void gbamachine_Machine_startLine(gbamachine_Machine* m);
void gbamachine_Machine_hblankNow(gbamachine_Machine* m);
void gbamachine_Machine_raise(gbamachine_Machine* m,uint16_t mask);
void gbamachine_Machine_deliver(gbamachine_Machine* m);
void gbamachine_Machine_biosIRQExit(gbamachine_Machine* m);
void gbamachine_Machine_runQuantum(gbamachine_Machine* m,int64_t n,Map<uint32_t,std::string> milestones,Map<uint32_t,uint64_t> hit);
uint64_t gbamachine_Machine_progressSig(gbamachine_Machine* m);
void gbamachine_fifo_push(gbamachine_fifo* f,uint8_t b);
void gbamachine_fifo_pop(gbamachine_fifo* f);
void gbamachine_fifo_reset(gbamachine_fifo* f);
gbamachine_apu* gbamachine_newAPU();
bool gbamachine_Machine_soundWrite(gbamachine_Machine* m,uint32_t reg,uint16_t v);
void gbamachine_apu_writeDutyEnv(gbamachine_apu* a,gbamachine_square* c,uint16_t v);
void gbamachine_apu_writeFreqCtl(gbamachine_apu* a,gbamachine_square* c,uint16_t v);
std::tuple<uint16_t,bool> gbamachine_Machine_soundRead(gbamachine_Machine* m,uint32_t reg);
void gbamachine_Machine_fifoTimerOverflow(gbamachine_Machine* m,int64_t n,int64_t times);
void gbamachine_apu_mixCycles(gbamachine_apu* a,int64_t cycles);
void gbamachine_apu_mixOne(gbamachine_apu* a);
void gbamachine_apu_stepFrame(gbamachine_apu* a,int64_t step);
void gbamachine_apu_clockLength(gbamachine_apu* a);
void gbamachine_apu_clockSweep(gbamachine_apu* a);
void gbamachine_apu_clockEnv(gbamachine_apu* a);
double gbamachine_square_sample(gbamachine_square* c,double dt);
double gbamachine_waveCh_sample(gbamachine_waveCh* c,double dt);
double gbamachine_noiseCh_sample(gbamachine_noiseCh* c,double dt);
void gbamachine_Machine_AudioCapture(gbamachine_Machine* m,bool on);
int64_t gbamachine_Machine_AudioSamples(gbamachine_Machine* m);
Slice<std::string> gbamachine_Machine_SoundState(gbamachine_Machine* m);
std::string gbamachine_enabledStr(bool on);
void gbamachine_Machine_timerRegWrite(gbamachine_Machine* m,uint32_t reg,uint16_t v);
void gbamachine_Machine_tickTimers(gbamachine_Machine* m,int64_t cycles);
constexpr arm_Flow arm_FlowSeq=0ULL;
constexpr arm_Flow arm_FlowBranch=1ULL;
constexpr arm_Flow arm_FlowJump=2ULL;
constexpr arm_Flow arm_FlowCall=3ULL;
constexpr arm_Flow arm_FlowReturn=4ULL;
constexpr arm_Flow arm_FlowIndJump=5ULL;
constexpr arm_Flow arm_FlowStop=6ULL;
constexpr arm_Flow arm_FlowIndCall=7ULL;
constexpr int64_t arm_condEQ=0ULL;
constexpr int64_t arm_condNE=1ULL;
constexpr int64_t arm_condCS=2ULL;
constexpr int64_t arm_condCC=3ULL;
constexpr int64_t arm_condMI=4ULL;
constexpr int64_t arm_condPL=5ULL;
constexpr int64_t arm_condVS=6ULL;
constexpr int64_t arm_condVC=7ULL;
constexpr int64_t arm_condHI=8ULL;
constexpr int64_t arm_condLS=9ULL;
constexpr int64_t arm_condGE=10ULL;
constexpr int64_t arm_condLT=11ULL;
constexpr int64_t arm_condGT=12ULL;
constexpr int64_t arm_condLE=13ULL;
constexpr int64_t arm_condAL=14ULL;
constexpr int64_t arm_condNV=15ULL;
std::array<std::string,16> arm_condName=std::array<std::string,16>{std::string("EQ",2),std::string("NE",2),std::string("CS",2),std::string("CC",2),std::string("MI",2),std::string("PL",2),std::string("VS",2),std::string("VC",2),std::string("HI",2),std::string("LS",2),std::string("GE",2),std::string("LT",2),std::string("GT",2),std::string("LE",2),std::string("",0),std::string("NV",2)};
std::array<std::string,16> arm_regName=std::array<std::string,16>{std::string("r0",2),std::string("r1",2),std::string("r2",2),std::string("r3",2),std::string("r4",2),std::string("r5",2),std::string("r6",2),std::string("r7",2),std::string("r8",2),std::string("r9",2),std::string("r10",3),std::string("r11",3),std::string("r12",3),std::string("sp",2),std::string("lr",2),std::string("pc",2)};
std::array<std::string,16> arm_dpOps=std::array<std::string,16>{std::string("AND",3),std::string("EOR",3),std::string("SUB",3),std::string("RSB",3),std::string("ADD",3),std::string("ADC",3),std::string("SBC",3),std::string("RSC",3),std::string("TST",3),std::string("TEQ",3),std::string("CMP",3),std::string("CMN",3),std::string("ORR",3),std::string("MOV",3),std::string("BIC",3),std::string("MVN",3)};
std::array<std::string,4> arm_shiftName=std::array<std::string,4>{std::string("LSL",3),std::string("LSR",3),std::string("ASR",3),std::string("ROR",3)};
constexpr int64_t arm_ModeUSR=16ULL;
constexpr int64_t arm_ModeFIQ=17ULL;
constexpr int64_t arm_ModeIRQ=18ULL;
constexpr int64_t arm_ModeSVC=19ULL;
constexpr int64_t arm_ModeABT=23ULL;
constexpr int64_t arm_ModeUND=27ULL;
constexpr int64_t arm_ModeSYS=31ULL;
constexpr arm_parMode arm_parWrap=0ULL;
constexpr arm_parMode arm_parSat=1ULL;
constexpr arm_parMode arm_parHalve=2ULL;
std::array<std::string,16> arm_thumbALUOps=std::array<std::string,16>{std::string("AND",3),std::string("EOR",3),std::string("LSL",3),std::string("LSR",3),std::string("ASR",3),std::string("ADC",3),std::string("SBC",3),std::string("ROR",3),std::string("TST",3),std::string("NEG",3),std::string("CMP",3),std::string("CMN",3),std::string("ORR",3),std::string("MUL",3),std::string("BIC",3),std::string("MVN",3)};
constexpr arm_Variant arm_V5TE=0ULL;
constexpr arm_Variant arm_V6K=1ULL;
constexpr arm_Variant arm_V4T=2ULL;
Map<std::string,bool> arm_v4tMnem=Map<std::string,bool>{{std::string("BLX",3),true},{std::string("CLZ",3),true},{std::string("BKPT",4),true},{std::string("PLD",3),true},{std::string("LDRD",4),true},{std::string("STRD",4),true},{std::string("QADD",4),true},{std::string("QSUB",4),true},{std::string("QDADD",5),true},{std::string("QDSUB",5),true},{std::string("SMLABB",6),true},{std::string("SMLABT",6),true},{std::string("SMLATB",6),true},{std::string("SMLATT",6),true},{std::string("SMLAWB",6),true},{std::string("SMLAWT",6),true},{std::string("SMULWB",6),true},{std::string("SMULWT",6),true},{std::string("SMLALBB",7),true},{std::string("SMLALBT",7),true},{std::string("SMLALTB",7),true},{std::string("SMLALTT",7),true},{std::string("SMULBB",6),true},{std::string("SMULBT",6),true},{std::string("SMULTB",6),true},{std::string("SMULTT",6),true}};
constexpr int64_t gba_HeaderLen=192ULL;
constexpr int64_t gba_ROMBase=134217728ULL;
Slice<Anon7> gba_saveIDs=Slice<Anon7>{Anon7{std::string("EEPROM_V",8),std::string("EEPROM (serial, 512 B or 8 KiB)",31)},Anon7{std::string("FLASH1M_V",9),std::string("Flash 128 KiB",13)},Anon7{std::string("FLASH512_V",10),std::string("Flash 64 KiB",12)},Anon7{std::string("FLASH_V",7),std::string("Flash 64 KiB",12)},Anon7{std::string("SRAM_V",6),std::string("SRAM 32 KiB",11)},Anon7{std::string("SRAM_F_V",8),std::string("SRAM 32 KiB",11)}};
constexpr int64_t gbamachine_biosSize=16384ULL;
constexpr int64_t gbamachine_ewramBase=33554432ULL;
constexpr int64_t gbamachine_ewramSize=262144ULL;
constexpr int64_t gbamachine_iwramBase=50331648ULL;
constexpr int64_t gbamachine_iwramSize=32768ULL;
constexpr int64_t gbamachine_ioBase=67108864ULL;
constexpr int64_t gbamachine_palBase=83886080ULL;
constexpr int64_t gbamachine_palSize=1024ULL;
constexpr int64_t gbamachine_vramBase=100663296ULL;
constexpr int64_t gbamachine_vramSize=98304ULL;
constexpr int64_t gbamachine_oamBase=117440512ULL;
constexpr int64_t gbamachine_oamSize=1024ULL;
constexpr int64_t gbamachine_romBase=134217728ULL;
constexpr int64_t gbamachine_irqHandlerSlot=50364412ULL;
constexpr int64_t gbamachine_irqCheckFlags=50364408ULL;
constexpr int64_t gbamachine_screenW=240ULL;
constexpr int64_t gbamachine_screenH=160ULL;
constexpr int64_t gbamachine_irqVBlank=1ULL;
constexpr int64_t gbamachine_irqHBlank=2ULL;
constexpr int64_t gbamachine_irqVCount=4ULL;
constexpr int64_t gbamachine_irqTimer0=8ULL;
constexpr int64_t gbamachine_irqSerial=128ULL;
constexpr int64_t gbamachine_irqDMA0=256ULL;
constexpr int64_t gbamachine_irqKeypad=4096ULL;
constexpr int64_t gbamachine_irqGamePak=8192ULL;
std::array<std::array<std::array<int64_t,2>,4>,3> gbamachine_objSizes=std::array<std::array<std::array<int64_t,2>,4>,3>{std::array<std::array<int64_t,2>,4>{std::array<int64_t,2>{cast<int64_t>(8ULL),cast<int64_t>(8ULL)},std::array<int64_t,2>{cast<int64_t>(16ULL),cast<int64_t>(16ULL)},std::array<int64_t,2>{cast<int64_t>(32ULL),cast<int64_t>(32ULL)},std::array<int64_t,2>{cast<int64_t>(64ULL),cast<int64_t>(64ULL)}},std::array<std::array<int64_t,2>,4>{std::array<int64_t,2>{cast<int64_t>(16ULL),cast<int64_t>(8ULL)},std::array<int64_t,2>{cast<int64_t>(32ULL),cast<int64_t>(8ULL)},std::array<int64_t,2>{cast<int64_t>(32ULL),cast<int64_t>(16ULL)},std::array<int64_t,2>{cast<int64_t>(64ULL),cast<int64_t>(32ULL)}},std::array<std::array<int64_t,2>,4>{std::array<int64_t,2>{cast<int64_t>(8ULL),cast<int64_t>(16ULL)},std::array<int64_t,2>{cast<int64_t>(8ULL),cast<int64_t>(32ULL)},std::array<int64_t,2>{cast<int64_t>(16ULL),cast<int64_t>(32ULL)},std::array<int64_t,2>{cast<int64_t>(32ULL),cast<int64_t>(64ULL)}}};
constexpr int64_t gbamachine_linesPerFrame=228ULL;
constexpr int64_t gbamachine_visibleLines=160ULL;
constexpr int64_t gbamachine_cyclesPerLine=1232ULL;
constexpr int64_t gbamachine_instrsPerLine=410ULL;
constexpr int64_t gbamachine_biosIRQReturn=4294905856ULL;
constexpr int64_t gbamachine_audioRate=32768ULL;
constexpr double gbamachine_cpuHz=1.67772160000000000e+07;
std::array<std::array<uint8_t,8>,4> gbamachine_dutyTable=std::array<std::array<uint8_t,8>,4>{std::array<uint8_t,8>{cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),cast<uint8_t>(1ULL)},std::array<uint8_t,8>{cast<uint8_t>(1ULL),cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),cast<uint8_t>(1ULL)},std::array<uint8_t,8>{cast<uint8_t>(1ULL),cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),cast<uint8_t>(1ULL),cast<uint8_t>(1ULL),cast<uint8_t>(1ULL)},std::array<uint8_t,8>{cast<uint8_t>(0ULL),cast<uint8_t>(1ULL),cast<uint8_t>(1ULL),cast<uint8_t>(1ULL),cast<uint8_t>(1ULL),cast<uint8_t>(1ULL),cast<uint8_t>(1ULL),cast<uint8_t>(0ULL)}};
std::array<int64_t,8> gbamachine_noiseDivisors=std::array<int64_t,8>{cast<int64_t>(8ULL),cast<int64_t>(16ULL),cast<int64_t>(32ULL),cast<int64_t>(48ULL),cast<int64_t>(64ULL),cast<int64_t>(80ULL),cast<int64_t>(96ULL),cast<int64_t>(112ULL)};
std::array<double,4> gbamachine_psgVolTable=std::array<double,4>{0.25,0.5,1.0,1.0};
std::array<int64_t,4> gbamachine_timerPrescale=std::array<int64_t,4>{cast<int64_t>(1ULL),cast<int64_t>(64ULL),cast<int64_t>(256ULL),cast<int64_t>(1024ULL)};

#include "adapters.h"
// tools/cpu/arm/arm.go:109:1
uint32_t arm_signExtend(uint32_t v,uint64_t n){
{
uint32_t m = shl<uint32_t>(cast<uint32_t>(1ULL),(cast<uint64_t>((n - cast<uint64_t>(1ULL)))));
return cast<uint32_t>(((cast<uint32_t>((v ^ m))) - m));
}
}
// tools/cpu/arm/arm.go:116:1
std::string arm_imm(uint32_t v){
{
if ((v < cast<uint32_t>(10ULL))) {
return go_fmt_Sprintf(std::string("#%d",3),v);
}
return go_fmt_Sprintf(std::string("#0x%X",5),v);
}
}
// tools/cpu/arm/arm.go:127:1
arm_Inst arm_Decode(Slice<uint8_t> code,uint32_t addr,bool thumb){
{
if (thumb) {
return arm_DecodeThumb(code,addr);
}
return arm_DecodeARM(code,addr);
}
}
// tools/cpu/arm/arm.go:136:1
std::tuple<uint32_t,bool> arm_word(Slice<uint8_t> code){
{
if ((len(code) < cast<int64_t>(4ULL))) {
return {cast<uint32_t>(0ULL),false};
}
return {cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(code[cast<int64_t>(0ULL)]) | shl<uint32_t>(cast<uint32_t>(code[cast<int64_t>(1ULL)]),cast<int64_t>(8ULL)))) | shl<uint32_t>(cast<uint32_t>(code[cast<int64_t>(2ULL)]),cast<int64_t>(16ULL)))) | shl<uint32_t>(cast<uint32_t>(code[cast<int64_t>(3ULL)]),cast<int64_t>(24ULL)))),true};
}
}
// tools/cpu/arm/arm.go:144:1
arm_Inst arm_DecodeARM(Slice<uint8_t> code,uint32_t addr){
{
auto tmp1 = arm_word(code);
uint32_t w = std::get<0>(tmp1);
bool ok = std::get<1>(tmp1);
if ((!ok)) {
return arm_Inst{addr,len(code),std::string(".word",5),std::string(".word ; truncated",17),cast<arm_Flow>(6ULL),{},{},{},{},cast<int64_t>(14ULL)};
}
int64_t cond = cast<int64_t>(shr<uint32_t>(w,cast<int64_t>(28ULL)));
arm_Inst in = arm_Inst{addr,cast<int64_t>(4ULL),{},{},cast<arm_Flow>(0ULL),{},{},{},{},cond};
if ((cond == cast<int64_t>(15ULL))) {
return arm_decodeUncond(w,addr,in);
}
{
switch(cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(25ULL))) & cast<uint32_t>(7ULL)))){
case cast<uint32_t>(0ULL):{
return arm_decodeDataMisc(w,addr,in,false);
break;}
case cast<uint32_t>(1ULL):{
return arm_decodeDataMisc(w,addr,in,true);
break;}
case cast<uint32_t>(2ULL):{
return arm_decodeSingle(w,in,false);
break;}
case cast<uint32_t>(3ULL):{
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(4ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
return arm_undef(w,in);
}
return arm_decodeSingle(w,in,true);
break;}
case cast<uint32_t>(4ULL):{
return arm_decodeBlock(w,in);
break;}
case cast<uint32_t>(5ULL):{
return arm_decodeBranch(w,addr,in);
break;}
case cast<uint32_t>(6ULL):{
return arm_decodeCoproLS(w,in);
break;}
default:{
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(24ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
in.Mnem = (std::string("SWI",3) + arm_cn(cond));
in.Text = go_fmt_Sprintf(std::string("%s #0x%X",8),in.Mnem,cast<uint32_t>((w & cast<uint32_t>(16777215ULL))));
return in;
}
return arm_decodeCopro(w,in);
break;}
}}
}
}
// tools/cpu/arm/arm.go:186:1
std::string arm_cn(int64_t cond){
{
return arm_condName[cond];
}
}
// tools/cpu/arm/arm.go:188:1
arm_Inst arm_undef(uint32_t w,arm_Inst in){
{
auto tmp2 = std::make_tuple(std::string(".word",5),cast<arm_Flow>(6ULL));
in.Mnem = std::get<0>(tmp2);
in.Flow = std::get<1>(tmp2);
in.Text = go_fmt_Sprintf(std::string(".word 0x%08X",12),w);
return in;
}
}
// tools/cpu/arm/arm.go:199:1
arm_Inst arm_decodeDataMisc(uint32_t w,uint32_t addr,arm_Inst in,bool immForm){
{
uint32_t op = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(21ULL))) & cast<uint32_t>(15ULL)));
uint32_t s = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(20ULL))) & cast<uint32_t>(1ULL)));
if ((((!immForm) && (cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(4ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) && (cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(7ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL)))) {
return arm_decodeExtension(w,in);
}
if ((((s == cast<uint32_t>(0ULL)) && (op >= cast<uint32_t>(8ULL))) && (op <= cast<uint32_t>(11ULL)))) {
return arm_decodeMisc(w,in,immForm);
}
return arm_decodeDataProc(w,addr,in,immForm);
}
}
// tools/cpu/arm/arm.go:215:1
arm_Inst arm_decodeDataProc(uint32_t w,uint32_t addr,arm_Inst in,bool immForm){
{
uint32_t op = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(21ULL))) & cast<uint32_t>(15ULL)));
uint32_t s = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(20ULL))) & cast<uint32_t>(1ULL)));
uint32_t rn = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(15ULL)));
uint32_t rd = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(12ULL))) & cast<uint32_t>(15ULL)));
std::string op2={};
if (immForm) {
uint32_t v = arm_ror32(cast<uint32_t>((w & cast<uint32_t>(255ULL))),cast<uint32_t>(((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(8ULL))) & cast<uint32_t>(15ULL)))) * cast<uint32_t>(2ULL))));
op2 = arm_imm(v);
}
else {
op2 = arm_shiftOperand(w);
}
std::string sfx = std::string("",0);
if ((s == cast<uint32_t>(1ULL))) {
sfx = std::string("S",1);
}
std::string name = arm_dpOps[op];
in.Mnem = ((name + arm_cn(in.Cond)) + sfx);
{
switch(op){
case cast<uint32_t>(8ULL):case cast<uint32_t>(9ULL):case cast<uint32_t>(10ULL):case cast<uint32_t>(11ULL):{
in.Mnem = (arm_dpOps[op] + arm_cn(in.Cond));
in.Text = go_fmt_Sprintf(std::string("%s %s, %s",9),in.Mnem,arm_regName[rn],op2);
break;}
case cast<uint32_t>(13ULL):case cast<uint32_t>(15ULL):{
in.Text = go_fmt_Sprintf(std::string("%s %s, %s",9),in.Mnem,arm_regName[rd],op2);
break;}
default:{
in.Text = go_fmt_Sprintf(std::string("%s %s, %s, %s",13),in.Mnem,arm_regName[rd],arm_regName[rn],op2);
break;}
}}
if ((((((rd == cast<uint32_t>(15ULL)) && (op != cast<uint32_t>(8ULL))) && (op != cast<uint32_t>(9ULL))) && (op != cast<uint32_t>(10ULL))) && (op != cast<uint32_t>(11ULL)))) {
arm_classifyPCWrite((&in),w,op);
}
return in;
}
}
// tools/cpu/arm/arm.go:255:1
void arm_classifyPCWrite(arm_Inst* in,uint32_t w,uint32_t op){
{
uint32_t rm = cast<uint32_t>((w & cast<uint32_t>(15ULL)));
bool isReg = ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(25ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(0ULL)) && (cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(4ULL))) & cast<uint32_t>(255ULL))) == cast<uint32_t>(0ULL)));
if ((((op == cast<uint32_t>(13ULL)) && isReg) && (rm == cast<uint32_t>(14ULL)))) {
in->Flow = cast<arm_Flow>(4ULL);
return ;
}
in->Flow = cast<arm_Flow>(5ULL);
}
}
// tools/cpu/arm/arm.go:266:1
std::string arm_shiftOperand(uint32_t w){
{
uint32_t rm = cast<uint32_t>((w & cast<uint32_t>(15ULL)));
uint32_t styp = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(5ULL))) & cast<uint32_t>(3ULL)));
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(4ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
uint32_t rs = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(8ULL))) & cast<uint32_t>(15ULL)));
return go_fmt_Sprintf(std::string("%s, %s %s",9),arm_regName[rm],arm_shiftName[styp],arm_regName[rs]);
}
uint32_t amt = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(7ULL))) & cast<uint32_t>(31ULL)));
{
switch(styp){
case cast<uint32_t>(0ULL):{
if ((amt == cast<uint32_t>(0ULL))) {
return arm_regName[rm];
}
return go_fmt_Sprintf(std::string("%s, LSL #%d",11),arm_regName[rm],amt);
break;}
case cast<uint32_t>(1ULL):{
if ((amt == cast<uint32_t>(0ULL))) {
amt = cast<uint32_t>(32ULL);
}
return go_fmt_Sprintf(std::string("%s, LSR #%d",11),arm_regName[rm],amt);
break;}
case cast<uint32_t>(2ULL):{
if ((amt == cast<uint32_t>(0ULL))) {
amt = cast<uint32_t>(32ULL);
}
return go_fmt_Sprintf(std::string("%s, ASR #%d",11),arm_regName[rm],amt);
break;}
default:{
if ((amt == cast<uint32_t>(0ULL))) {
return go_fmt_Sprintf(std::string("%s, RRX",7),arm_regName[rm]);
}
return go_fmt_Sprintf(std::string("%s, ROR #%d",11),arm_regName[rm],amt);
break;}
}}
}
}
// tools/cpu/arm/arm.go:299:1
uint32_t arm_ror32(uint32_t v,uint32_t n){
{
n &= cast<uint32_t>(31ULL);
return cast<uint32_t>((shr<uint32_t>(v,n) | shl<uint32_t>(v,(cast<uint32_t>((cast<uint32_t>(32ULL) - n))))));
}
}
// tools/cpu/arm/arm.go:307:1
arm_Inst arm_decodeMisc(uint32_t w,arm_Inst in,bool immForm){
{
uint32_t op = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(21ULL))) & cast<uint32_t>(15ULL)));
if (immForm) {
return arm_decodeMSRimm(w,in);
}
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(7ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
return arm_decodeSignedMul(w,in);
}
uint32_t op2 = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(4ULL))) & cast<uint32_t>(15ULL)));
{
switch(op2){
case cast<uint32_t>(0ULL):{
if (((op == cast<uint32_t>(8ULL)) || (op == cast<uint32_t>(10ULL)))) {
std::string psr = std::string("CPSR",4);
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(22ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
psr = std::string("SPSR",4);
}
in.Mnem = (std::string("MRS",3) + arm_cn(in.Cond));
in.Text = go_fmt_Sprintf(std::string("%s %s, %s",9),in.Mnem,arm_regName[cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(12ULL))) & cast<uint32_t>(15ULL)))],psr);
return in;
}
return arm_decodeMSRreg(w,in);
break;}
case cast<uint32_t>(1ULL):{
{
switch(op){
case cast<uint32_t>(9ULL):{
uint32_t rm = cast<uint32_t>((w & cast<uint32_t>(15ULL)));
in.Mnem = (std::string("BX",2) + arm_cn(in.Cond));
in.Text = go_fmt_Sprintf(std::string("%s %s",5),in.Mnem,arm_regName[rm]);
if ((rm == cast<uint32_t>(14ULL))) {
in.Flow = cast<arm_Flow>(4ULL);
}
else {
in.Flow = cast<arm_Flow>(5ULL);
}
return in;
break;}
case cast<uint32_t>(11ULL):{
in.Mnem = (std::string("CLZ",3) + arm_cn(in.Cond));
in.Text = go_fmt_Sprintf(std::string("%s %s, %s",9),in.Mnem,arm_regName[cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(12ULL))) & cast<uint32_t>(15ULL)))],arm_regName[cast<uint32_t>((w & cast<uint32_t>(15ULL)))]);
return in;
break;}
}}
break;}
case cast<uint32_t>(3ULL):{
if ((op == cast<uint32_t>(9ULL))) {
in.Mnem = (std::string("BLX",3) + arm_cn(in.Cond));
in.Text = go_fmt_Sprintf(std::string("%s %s",5),in.Mnem,arm_regName[cast<uint32_t>((w & cast<uint32_t>(15ULL)))]);
in.Flow = cast<arm_Flow>(7ULL);
return in;
}
break;}
case cast<uint32_t>(5ULL):{
std::string sat = std::array<std::string,4>{std::string("QADD",4),std::string("QSUB",4),std::string("QDADD",5),std::string("QDSUB",5)}[cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(21ULL))) & cast<uint32_t>(3ULL)))];
in.Mnem = (sat + arm_cn(in.Cond));
in.Text = go_fmt_Sprintf(std::string("%s %s, %s, %s",13),in.Mnem,arm_regName[cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(12ULL))) & cast<uint32_t>(15ULL)))],arm_regName[cast<uint32_t>((w & cast<uint32_t>(15ULL)))],arm_regName[cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(15ULL)))]);
return in;
break;}
case cast<uint32_t>(7ULL):{
if ((op == cast<uint32_t>(9ULL))) {
in.Mnem = std::string("BKPT",4);
uint32_t imm16 = cast<uint32_t>((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(4ULL))) & cast<uint32_t>(65520ULL))) | (cast<uint32_t>((w & cast<uint32_t>(15ULL))))));
in.Text = go_fmt_Sprintf(std::string("BKPT #0x%X",10),imm16);
in.Flow = cast<arm_Flow>(6ULL);
return in;
}
break;}
}}
return arm_undef(w,in);
}
}
// tools/cpu/arm/arm.go:378:1
arm_Inst arm_decodeMSRreg(uint32_t w,arm_Inst in){
{
std::string psr = std::string("CPSR",4);
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(22ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
psr = std::string("SPSR",4);
}
in.Mnem = (std::string("MSR",3) + arm_cn(in.Cond));
in.Text = go_fmt_Sprintf(std::string("%s %s%s, %s",11),in.Mnem,psr,arm_msrFields(w),arm_regName[cast<uint32_t>((w & cast<uint32_t>(15ULL)))]);
return in;
}
}
// tools/cpu/arm/arm.go:389:1
arm_Inst arm_decodeMSRimm(uint32_t w,arm_Inst in){
{
std::string psr = std::string("CPSR",4);
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(22ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
psr = std::string("SPSR",4);
}
uint32_t v = arm_ror32(cast<uint32_t>((w & cast<uint32_t>(255ULL))),cast<uint32_t>(((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(8ULL))) & cast<uint32_t>(15ULL)))) * cast<uint32_t>(2ULL))));
in.Mnem = (std::string("MSR",3) + arm_cn(in.Cond));
in.Text = go_fmt_Sprintf(std::string("%s %s%s, %s",11),in.Mnem,psr,arm_msrFields(w),arm_imm(v));
return in;
}
}
// tools/cpu/arm/arm.go:401:1
std::string arm_msrFields(uint32_t w){
{
uint32_t m = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(15ULL)));
if ((m == cast<uint32_t>(0ULL))) {
return std::string("",0);
}
std::string s = std::string("_",1);
{auto&& tmp3 = Slice<std::string>{std::string("c",1),std::string("x",1),std::string("s",1),std::string("f",1)};
for(int64_t tmp4=0;tmp4<len(tmp3);++tmp4){
auto i=tmp4;auto f=tmp3[tmp4];if ((cast<uint32_t>((m & (shl<uint32_t>(cast<uint32_t>(1ULL),cast<uint64_t>(i))))) != cast<uint32_t>(0ULL))) {
s += f;
}
}}
return s;
}
}
// tools/cpu/arm/arm.go:417:1
arm_Inst arm_decodeSignedMul(uint32_t w,arm_Inst in){
{
uint32_t op = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(21ULL))) & cast<uint32_t>(15ULL)));
uint32_t x = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(5ULL))) & cast<uint32_t>(1ULL)));
uint32_t y = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(6ULL))) & cast<uint32_t>(1ULL)));
auto bt = [&](uint32_t b)->uint8_t{
if ((b == cast<uint32_t>(0ULL))) {
return cast<uint8_t>(66ULL);
}
return cast<uint8_t>(84ULL);
}
;
uint32_t rd = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(15ULL)));
uint32_t rn = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(12ULL))) & cast<uint32_t>(15ULL)));
uint32_t rs = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(8ULL))) & cast<uint32_t>(15ULL)));
uint32_t rm = cast<uint32_t>((w & cast<uint32_t>(15ULL)));
{
switch(op){
case cast<uint32_t>(8ULL):{
in.Mnem = go_fmt_Sprintf(std::string("SMLA%c%c%s",10),bt(x),bt(y),arm_cn(in.Cond));
in.Text = go_fmt_Sprintf(std::string("%s %s, %s, %s, %s",17),in.Mnem,arm_regName[rd],arm_regName[rm],arm_regName[rs],arm_regName[rn]);
break;}
case cast<uint32_t>(9ULL):{
if ((x == cast<uint32_t>(0ULL))) {
in.Mnem = go_fmt_Sprintf(std::string("SMLAW%c%s",9),bt(y),arm_cn(in.Cond));
in.Text = go_fmt_Sprintf(std::string("%s %s, %s, %s, %s",17),in.Mnem,arm_regName[rd],arm_regName[rm],arm_regName[rs],arm_regName[rn]);
}
else {
in.Mnem = go_fmt_Sprintf(std::string("SMULW%c%s",9),bt(y),arm_cn(in.Cond));
in.Text = go_fmt_Sprintf(std::string("%s %s, %s, %s",13),in.Mnem,arm_regName[rd],arm_regName[rm],arm_regName[rs]);
}
break;}
case cast<uint32_t>(10ULL):{
in.Mnem = go_fmt_Sprintf(std::string("SMLAL%c%c%s",11),bt(x),bt(y),arm_cn(in.Cond));
in.Text = go_fmt_Sprintf(std::string("%s %s, %s, %s, %s",17),in.Mnem,arm_regName[rn],arm_regName[rd],arm_regName[rm],arm_regName[rs]);
break;}
default:{
in.Mnem = go_fmt_Sprintf(std::string("SMUL%c%c%s",10),bt(x),bt(y),arm_cn(in.Cond));
in.Text = go_fmt_Sprintf(std::string("%s %s, %s, %s",13),in.Mnem,arm_regName[rd],arm_regName[rm],arm_regName[rs]);
break;}
}}
return in;
}
}
// tools/cpu/arm/arm.go:456:1
arm_Inst arm_decodeExtension(uint32_t w,arm_Inst in){
{
uint32_t sh = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(5ULL))) & cast<uint32_t>(3ULL)));
if ((sh == cast<uint32_t>(0ULL))) {
{
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(24ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))){
std::string b = std::string("",0);
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(22ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
b = std::string("B",1);
}
in.Mnem = ((std::string("SWP",3) + arm_cn(in.Cond)) + b);
in.Text = go_fmt_Sprintf(std::string("%s %s, %s, [%s]",15),in.Mnem,arm_regName[cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(12ULL))) & cast<uint32_t>(15ULL)))],arm_regName[cast<uint32_t>((w & cast<uint32_t>(15ULL)))],arm_regName[cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(15ULL)))]);
return in;
}
else if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(23ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))){
return arm_decodeMulLong(w,in);
}
else {
return arm_decodeMul(w,in);
}
}
tmp5:;
}
return arm_decodeHalf(w,in,sh);
}
}
// tools/cpu/arm/arm.go:479:1
arm_Inst arm_decodeMul(uint32_t w,arm_Inst in){
{
std::string s = std::string("",0);
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(20ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
s = std::string("S",1);
}
uint32_t rd = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(15ULL)));
uint32_t rn = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(12ULL))) & cast<uint32_t>(15ULL)));
uint32_t rs = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(8ULL))) & cast<uint32_t>(15ULL)));
uint32_t rm = cast<uint32_t>((w & cast<uint32_t>(15ULL)));
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(21ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
in.Mnem = ((std::string("MLA",3) + arm_cn(in.Cond)) + s);
in.Text = go_fmt_Sprintf(std::string("%s %s, %s, %s, %s",17),in.Mnem,arm_regName[rd],arm_regName[rm],arm_regName[rs],arm_regName[rn]);
}
else {
in.Mnem = ((std::string("MUL",3) + arm_cn(in.Cond)) + s);
in.Text = go_fmt_Sprintf(std::string("%s %s, %s, %s",13),in.Mnem,arm_regName[rd],arm_regName[rm],arm_regName[rs]);
}
return in;
}
}
// tools/cpu/arm/arm.go:499:1
arm_Inst arm_decodeMulLong(uint32_t w,arm_Inst in){
{
std::array<std::string,4> names = std::array<std::string,4>{std::string("UMULL",5),std::string("UMLAL",5),std::string("SMULL",5),std::string("SMLAL",5)};
std::string s = std::string("",0);
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(20ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
s = std::string("S",1);
}
in.Mnem = ((names[cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(21ULL))) & cast<uint32_t>(3ULL)))] + arm_cn(in.Cond)) + s);
in.Text = go_fmt_Sprintf(std::string("%s %s, %s, %s, %s",17),in.Mnem,arm_regName[cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(12ULL))) & cast<uint32_t>(15ULL)))],arm_regName[cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(15ULL)))],arm_regName[cast<uint32_t>((w & cast<uint32_t>(15ULL)))],arm_regName[cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(8ULL))) & cast<uint32_t>(15ULL)))]);
return in;
}
}
// tools/cpu/arm/arm.go:513:1
arm_Inst arm_decodeHalf(uint32_t w,arm_Inst in,uint32_t sh){
{
uint32_t l = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(20ULL))) & cast<uint32_t>(1ULL)));
std::string name = std::string("",0);
{
if (((l == cast<uint32_t>(0ULL)) && (sh == cast<uint32_t>(1ULL)))){
name = std::string("STRH",4);
}
else if (((l == cast<uint32_t>(1ULL)) && (sh == cast<uint32_t>(1ULL)))){
name = std::string("LDRH",4);
}
else if (((l == cast<uint32_t>(1ULL)) && (sh == cast<uint32_t>(2ULL)))){
name = std::string("LDRSB",5);
}
else if (((l == cast<uint32_t>(1ULL)) && (sh == cast<uint32_t>(3ULL)))){
name = std::string("LDRSH",5);
}
else {
if ((sh == cast<uint32_t>(2ULL))) {
name = std::string("LDRD",4);
}
else {
name = std::string("STRD",4);
}
}
}
tmp6:;
in.Mnem = (name + arm_cn(in.Cond));
uint32_t rn = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(15ULL)));
uint32_t rd = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(12ULL))) & cast<uint32_t>(15ULL)));
uint32_t p = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(24ULL))) & cast<uint32_t>(1ULL)));
uint32_t u = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(23ULL))) & cast<uint32_t>(1ULL)));
std::string sign = std::string("+",1);
if ((u == cast<uint32_t>(0ULL))) {
sign = std::string("-",1);
}
std::string off={};
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(22ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
uint32_t v = cast<uint32_t>((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(4ULL))) & cast<uint32_t>(240ULL))) | (cast<uint32_t>((w & cast<uint32_t>(15ULL))))));
if ((v == cast<uint32_t>(0ULL))) {
off = std::string("",0);
}
else {
off = go_fmt_Sprintf(std::string(", #%s0x%X",9),arm_plusMinus(sign),v);
}
}
else {
off = go_fmt_Sprintf(std::string(", %s%s",6),arm_plusMinus(sign),arm_regName[cast<uint32_t>((w & cast<uint32_t>(15ULL)))]);
}
in.Text = arm_addrForm(in.Mnem,rd,rn,off,p,cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(21ULL))) & cast<uint32_t>(1ULL))));
return in;
}
}
// tools/cpu/arm/arm.go:557:1
std::string arm_plusMinus(std::string s){
{
if ((s == std::string("-",1))) {
return std::string("-",1);
}
return std::string("",0);
}
}
// tools/cpu/arm/arm.go:565:1
arm_Inst arm_decodeSingle(uint32_t w,arm_Inst in,bool regOff){
{
uint32_t l = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(20ULL))) & cast<uint32_t>(1ULL)));
uint32_t b = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(22ULL))) & cast<uint32_t>(1ULL)));
std::string name = std::string("STR",3);
if ((l == cast<uint32_t>(1ULL))) {
name = std::string("LDR",3);
}
if ((b == cast<uint32_t>(1ULL))) {
name += std::string("B",1);
}
in.Mnem = (name + arm_cn(in.Cond));
uint32_t rn = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(15ULL)));
uint32_t rd = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(12ULL))) & cast<uint32_t>(15ULL)));
uint32_t p = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(24ULL))) & cast<uint32_t>(1ULL)));
uint32_t u = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(23ULL))) & cast<uint32_t>(1ULL)));
std::string sign = std::string("",0);
if ((u == cast<uint32_t>(0ULL))) {
sign = std::string("-",1);
}
std::string off={};
if ((!regOff)) {
uint32_t v = cast<uint32_t>((w & cast<uint32_t>(4095ULL)));
if ((v == cast<uint32_t>(0ULL))) {
off = std::string("",0);
}
else {
off = go_fmt_Sprintf(std::string(", #%s0x%X",9),sign,v);
}
}
else {
off = ((std::string(", ",2) + sign) + arm_shiftOperand(w));
}
in.Text = arm_addrForm(in.Mnem,rd,rn,off,p,cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(21ULL))) & cast<uint32_t>(1ULL))));
if (((l == cast<uint32_t>(1ULL)) && (rd == cast<uint32_t>(15ULL)))) {
in.Flow = cast<arm_Flow>(5ULL);
}
return in;
}
}
// tools/cpu/arm/arm.go:609:1
std::string arm_addrForm(std::string mnem,uint32_t rd,uint32_t rn,std::string off,uint32_t p,uint32_t wb){
{
if ((p == cast<uint32_t>(1ULL))) {
std::string bang = std::string("",0);
if ((wb == cast<uint32_t>(1ULL))) {
bang = std::string("!",1);
}
return go_fmt_Sprintf(std::string("%s %s, [%s%s]%s",15),mnem,arm_regName[rd],arm_regName[rn],off,bang);
}
return go_fmt_Sprintf(std::string("%s %s, [%s]%s",13),mnem,arm_regName[rd],arm_regName[rn],off);
}
}
// tools/cpu/arm/arm.go:624:1
arm_Inst arm_decodeBlock(uint32_t w,arm_Inst in){
{
uint32_t l = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(20ULL))) & cast<uint32_t>(1ULL)));
uint32_t p = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(24ULL))) & cast<uint32_t>(1ULL)));
uint32_t u = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(23ULL))) & cast<uint32_t>(1ULL)));
std::string base = std::string("STM",3);
if ((l == cast<uint32_t>(1ULL))) {
base = std::string("LDM",3);
}
std::string mode = std::array<std::string,4>{std::string("DA",2),std::string("DB",2),std::string("IA",2),std::string("IB",2)}[cast<uint32_t>((shl<uint32_t>(u,cast<int64_t>(1ULL)) | p))];
in.Mnem = ((base + arm_cn(in.Cond)) + mode);
uint32_t rn = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(15ULL)));
std::string bang = std::string("",0);
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(21ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
bang = std::string("!",1);
}
std::string usr = std::string("",0);
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(22ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
usr = std::string("^",1);
}
in.Text = go_fmt_Sprintf(std::string("%s %s%s, {%s}%s",15),in.Mnem,arm_regName[rn],bang,arm_regList(cast<uint32_t>((w & cast<uint32_t>(65535ULL)))),usr);
if (((l == cast<uint32_t>(1ULL)) && (cast<uint32_t>((w & cast<uint32_t>(32768ULL))) != cast<uint32_t>(0ULL)))) {
in.Flow = cast<arm_Flow>(4ULL);
}
return in;
}
}
// tools/cpu/arm/arm.go:654:1
std::string arm_regList(uint32_t mask){
{
Slice<std::string> parts={};
int64_t i = cast<int64_t>(0ULL);
{;for (;(i < cast<int64_t>(16ULL));){
if ((cast<uint32_t>((mask & (shl<uint32_t>(cast<uint32_t>(1ULL),cast<uint64_t>(i))))) == cast<uint32_t>(0ULL))) {
i++;
continue;
}
int64_t j = i;
{;for (;((cast<int64_t>((j + cast<int64_t>(1ULL))) < cast<int64_t>(16ULL)) && (cast<uint32_t>((mask & (shl<uint32_t>(cast<uint32_t>(1ULL),cast<uint64_t>(cast<int64_t>((j + cast<int64_t>(1ULL)))))))) != cast<uint32_t>(0ULL)));){
j++;
}
}{
if ((j == i)){
parts = append(parts,Slice<std::string>{arm_regName[i]});
}
else if ((j == cast<int64_t>((i + cast<int64_t>(1ULL))))){
parts = append(parts,Slice<std::string>{arm_regName[i],arm_regName[j]});
}
else {
parts = append(parts,Slice<std::string>{((arm_regName[i] + std::string("-",1)) + arm_regName[j])});
}
}
tmp7:;
i = cast<int64_t>((j + cast<int64_t>(1ULL)));
}
}return arm_joinComma(parts);
}
}
// tools/cpu/arm/arm.go:679:1
std::string arm_joinComma(Slice<std::string> parts){
{
std::string out = std::string("",0);
{auto&& tmp8 = parts;
for(int64_t tmp9=0;tmp9<len(tmp8);++tmp9){
auto i=tmp9;auto p=tmp8[tmp9];if ((i > cast<int64_t>(0ULL))) {
out += std::string(", ",2);
}
out += p;
}}
return out;
}
}
// tools/cpu/arm/arm.go:692:1
arm_Inst arm_decodeBranch(uint32_t w,uint32_t addr,arm_Inst in){
{
uint32_t off = shl<uint32_t>(arm_signExtend(cast<uint32_t>((w & cast<uint32_t>(16777215ULL))),cast<uint64_t>(24ULL)),cast<int64_t>(2ULL));
uint32_t target = cast<uint32_t>((cast<uint32_t>((addr + cast<uint32_t>(8ULL))) + off));
auto tmp10 = std::make_tuple(target,true,false);
in.Target = std::get<0>(tmp10);
in.HasTarget = std::get<1>(tmp10);
in.TargetThumb = std::get<2>(tmp10);
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(24ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
in.Mnem = (std::string("BL",2) + arm_cn(in.Cond));
in.Text = go_fmt_Sprintf(std::string("%s 0x%08X",9),in.Mnem,target);
in.Flow = cast<arm_Flow>(3ULL);
return in;
}
in.Mnem = (std::string("B",1) + arm_cn(in.Cond));
in.Text = go_fmt_Sprintf(std::string("%s 0x%08X",9),in.Mnem,target);
if ((in.Cond == cast<int64_t>(14ULL))) {
in.Flow = cast<arm_Flow>(2ULL);
}
else {
in.Flow = cast<arm_Flow>(1ULL);
}
return in;
}
}
// tools/cpu/arm/arm.go:715:1
arm_Inst arm_decodeUncond(uint32_t w,uint32_t addr,arm_Inst in){
{
{
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(25ULL))) & cast<uint32_t>(7ULL))) == cast<uint32_t>(5ULL))){
uint32_t h = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(24ULL))) & cast<uint32_t>(1ULL)));
uint32_t off = cast<uint32_t>((shl<uint32_t>(arm_signExtend(cast<uint32_t>((w & cast<uint32_t>(16777215ULL))),cast<uint64_t>(24ULL)),cast<int64_t>(2ULL)) | shl<uint32_t>(h,cast<int64_t>(1ULL))));
uint32_t target = cast<uint32_t>((cast<uint32_t>((addr + cast<uint32_t>(8ULL))) + off));
in.Mnem = std::string("BLX",3);
in.Text = go_fmt_Sprintf(std::string("BLX 0x%08X",10),target);
auto tmp12 = std::make_tuple(cast<arm_Flow>(3ULL),target,true,true);
in.Flow = std::get<0>(tmp12);
in.Target = std::get<1>(tmp12);
in.HasTarget = std::get<2>(tmp12);
in.TargetThumb = std::get<3>(tmp12);
return in;
}
else if (((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(26ULL))) & cast<uint32_t>(3ULL))) == cast<uint32_t>(1ULL)) && (cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(20ULL))) & cast<uint32_t>(247ULL))) == cast<uint32_t>(85ULL)))){
in.Mnem = std::string("PLD",3);
in.Text = go_fmt_Sprintf(std::string("PLD [%s]",8),arm_regName[cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(15ULL)))]);
return in;
}
}
tmp11:;
return arm_undef(w,in);
}
}
// tools/cpu/arm/arm.go:740:1
arm_Inst arm_decodeCoproLS(uint32_t w,arm_Inst in){
{
uint32_t l = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(20ULL))) & cast<uint32_t>(1ULL)));
std::string name = std::string("STC",3);
if ((l == cast<uint32_t>(1ULL))) {
name = std::string("LDC",3);
}
in.Mnem = (name + arm_cn(in.Cond));
in.Text = go_fmt_Sprintf(std::string("%s p%d, c%d, [%s]",17),in.Mnem,cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(8ULL))) & cast<uint32_t>(15ULL))),cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(12ULL))) & cast<uint32_t>(15ULL))),arm_regName[cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(15ULL)))]);
return in;
}
}
// tools/cpu/arm/arm.go:754:1
arm_Inst arm_decodeCopro(uint32_t w,arm_Inst in){
{
uint32_t cp = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(8ULL))) & cast<uint32_t>(15ULL)));
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(4ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(0ULL))) {
in.Mnem = (std::string("CDP",3) + arm_cn(in.Cond));
in.Text = go_fmt_Sprintf(std::string("%s p%d, #%d, c%d, c%d, c%d",26),in.Mnem,cp,cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(20ULL))) & cast<uint32_t>(15ULL))),cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(12ULL))) & cast<uint32_t>(15ULL))),cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(15ULL))),cast<uint32_t>((w & cast<uint32_t>(15ULL))));
return in;
}
std::string name = std::string("MCR",3);
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(20ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
name = std::string("MRC",3);
}
in.Mnem = (name + arm_cn(in.Cond));
in.Text = go_fmt_Sprintf(std::string("%s p%d, #%d, %s, c%d, c%d, #%d",30),in.Mnem,cp,cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(21ULL))) & cast<uint32_t>(7ULL))),arm_regName[cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(12ULL))) & cast<uint32_t>(15ULL)))],cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(15ULL))),cast<uint32_t>((w & cast<uint32_t>(15ULL))),cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(5ULL))) & cast<uint32_t>(7ULL))));
return in;
}
}
// tools/cpu/arm/armexec.go:10:1
int64_t arm_CPU_Step(arm_CPU* c){
{
if (c->Halted) {
return cast<int64_t>(0ULL);
}
c->cur = c->R[cast<int64_t>(15ULL)];
c->branched = false;
c->Instrs++;
if (c->Thumb) {
return arm_CPU_stepThumb(c);
}
return arm_CPU_stepARM(c);
}
}
// tools/cpu/arm/armexec.go:23:1
uint32_t arm_CPU_boolToU(arm_CPU* c,bool b){
{
if (b) {
return cast<uint32_t>(1ULL);
}
return cast<uint32_t>(0ULL);
}
}
// tools/cpu/arm/armexec.go:31:1
int64_t arm_CPU_stepARM(arm_CPU* c){
{
uint32_t w = arm_CPU_read32aligned(c,c->cur);
int64_t cond = cast<int64_t>(shr<uint32_t>(w,cast<int64_t>(28ULL)));
if ((cond == cast<int64_t>(15ULL))) {
if ((arm_Variant_isV6(c->Arch) && arm_CPU_execUncondARMv6(c,w))) {
return cast<int64_t>(1ULL);
}
return arm_CPU_execUncondARM(c,w);
}
if ((!arm_CPU_cond(c,cond))) {
c->R[cast<int64_t>(15ULL)] = cast<uint32_t>((c->cur + cast<uint32_t>(4ULL)));
return cast<int64_t>(1ULL);
}
if ((arm_Variant_isV6(c->Arch) && arm_CPU_execARMv6(c,w))) {
if ((!c->branched)) {
c->R[cast<int64_t>(15ULL)] = cast<uint32_t>((c->cur + cast<uint32_t>(4ULL)));
}
return cast<int64_t>(1ULL);
}
arm_CPU_execARM(c,w);
if ((!c->branched)) {
c->R[cast<int64_t>(15ULL)] = cast<uint32_t>((c->cur + cast<uint32_t>(4ULL)));
}
return cast<int64_t>(1ULL);
}
}
// tools/cpu/arm/armexec.go:59:1
int64_t arm_CPU_execUncondARM(arm_CPU* c,uint32_t w){
{
if ((!arm_Variant_v5OrLater(c->Arch))) {
arm_CPU_undefV4T(c,w,std::string("cond=0b1111 (BLX immediate / PLD)",33));
return cast<int64_t>(1ULL);
}
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(25ULL))) & cast<uint32_t>(7ULL))) == cast<uint32_t>(5ULL))) {
uint32_t h = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(24ULL))) & cast<uint32_t>(1ULL)));
uint32_t off = cast<uint32_t>((shl<uint32_t>(arm_signExtend(cast<uint32_t>((w & cast<uint32_t>(16777215ULL))),cast<uint64_t>(24ULL)),cast<int64_t>(2ULL)) | shl<uint32_t>(h,cast<int64_t>(1ULL))));
c->R[cast<int64_t>(14ULL)] = cast<uint32_t>((c->cur + cast<uint32_t>(4ULL)));
c->Thumb = true;
c->R[cast<int64_t>(15ULL)] = ((cast<uint32_t>((cast<uint32_t>((c->cur + cast<uint32_t>(8ULL))) + off))) & ~(cast<uint32_t>(1ULL)));
c->branched = true;
return cast<int64_t>(3ULL);
}
c->R[cast<int64_t>(15ULL)] = cast<uint32_t>((c->cur + cast<uint32_t>(4ULL)));
return cast<int64_t>(1ULL);
}
}
// tools/cpu/arm/armexec.go:82:1
void arm_CPU_execARM(arm_CPU* c,uint32_t w){
{
{
switch(cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(25ULL))) & cast<uint32_t>(7ULL)))){
case cast<uint32_t>(0ULL):{
arm_CPU_execDataMisc(c,w,false);
break;}
case cast<uint32_t>(1ULL):{
arm_CPU_execDataMisc(c,w,true);
break;}
case cast<uint32_t>(2ULL):{
arm_CPU_execSingle(c,w,false);
break;}
case cast<uint32_t>(3ULL):{
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(4ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
arm_CPU_Halt(c,std::string("undefined media instruction 0x%08X at 0x%08X",44),w,c->cur);
return ;
}
arm_CPU_execSingle(c,w,true);
break;}
case cast<uint32_t>(4ULL):{
arm_CPU_execBlock(c,w);
break;}
case cast<uint32_t>(5ULL):{
arm_CPU_execBranch(c,w);
break;}
case cast<uint32_t>(6ULL):{
arm_CPU_Halt(c,std::string("unimplemented coprocessor transfer 0x%08X at 0x%08X",51),w,c->cur);
break;}
default:{
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(24ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
arm_CPU_execSWI(c,w);
}
else {
arm_CPU_execCopro(c,w);
}
break;}
}}
}
}
// tools/cpu/arm/armexec.go:111:1
void arm_CPU_execDataMisc(arm_CPU* c,uint32_t w,bool immForm){
{
uint32_t op = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(21ULL))) & cast<uint32_t>(15ULL)));
uint32_t s = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(20ULL))) & cast<uint32_t>(1ULL)));
if ((((!immForm) && (cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(4ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) && (cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(7ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL)))) {
arm_CPU_execExtension(c,w);
return ;
}
if ((((s == cast<uint32_t>(0ULL)) && (op >= cast<uint32_t>(8ULL))) && (op <= cast<uint32_t>(11ULL)))) {
arm_CPU_execMisc(c,w,immForm);
return ;
}
arm_CPU_execDataProc(c,w,immForm);
}
}
// tools/cpu/arm/armexec.go:127:1
std::tuple<uint32_t,uint32_t> arm_CPU_dpOperand(arm_CPU* c,uint32_t w,bool immForm){
{
if (immForm) {
uint32_t rot = cast<uint32_t>(((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(8ULL))) & cast<uint32_t>(15ULL)))) * cast<uint32_t>(2ULL)));
uint32_t v = arm_ror32(cast<uint32_t>((w & cast<uint32_t>(255ULL))),rot);
if ((rot == cast<uint32_t>(0ULL))) {
return {v,arm_CPU_boolToU(c,c->C)};
}
return {v,cast<uint32_t>(((shr<uint32_t>(v,cast<int64_t>(31ULL))) & cast<uint32_t>(1ULL)))};
}
uint32_t rm = arm_CPU_reg(c,cast<uint32_t>((w & cast<uint32_t>(15ULL))));
uint32_t typ = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(5ULL))) & cast<uint32_t>(3ULL)));
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(4ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
return arm_CPU_shift(c,typ,arm_CPU_reg(c,cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(8ULL))) & cast<uint32_t>(15ULL)))),rm,true,arm_CPU_boolToU(c,c->C));
}
return arm_CPU_shift(c,typ,cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(7ULL))) & cast<uint32_t>(31ULL))),rm,false,arm_CPU_boolToU(c,c->C));
}
}
// tools/cpu/arm/armexec.go:144:1
void arm_CPU_execDataProc(arm_CPU* c,uint32_t w,bool immForm){
{
uint32_t op = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(21ULL))) & cast<uint32_t>(15ULL)));
bool setFlags = (cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(20ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL));
uint32_t rn = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(15ULL)));
uint32_t rd = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(12ULL))) & cast<uint32_t>(15ULL)));
auto tmp13 = arm_CPU_dpOperand(c,w,immForm);
uint32_t op2 = std::get<0>(tmp13);
uint32_t shc = std::get<1>(tmp13);
uint32_t a = arm_CPU_reg(c,rn);
auto tmp14 = std::make_tuple(c->N,c->Z,c->C,c->V);
bool sn = std::get<0>(tmp14);
bool sz = std::get<1>(tmp14);
bool sc = std::get<2>(tmp14);
bool sv = std::get<3>(tmp14);
uint32_t res={};
bool logical = true;
{
switch(op){
case cast<uint32_t>(0ULL):{
res = cast<uint32_t>((a & op2));
break;}
case cast<uint32_t>(1ULL):{
res = cast<uint32_t>((a ^ op2));
break;}
case cast<uint32_t>(2ULL):{
auto tmp15 = std::make_tuple(arm_CPU_sub_(c,a,op2,cast<uint32_t>(1ULL)),false);
res = std::get<0>(tmp15);
logical = std::get<1>(tmp15);
break;}
case cast<uint32_t>(3ULL):{
auto tmp16 = std::make_tuple(arm_CPU_sub_(c,op2,a,cast<uint32_t>(1ULL)),false);
res = std::get<0>(tmp16);
logical = std::get<1>(tmp16);
break;}
case cast<uint32_t>(4ULL):{
auto tmp17 = std::make_tuple(arm_CPU_add(c,a,op2,cast<uint32_t>(0ULL)),false);
res = std::get<0>(tmp17);
logical = std::get<1>(tmp17);
break;}
case cast<uint32_t>(5ULL):{
auto tmp18 = std::make_tuple(arm_CPU_add(c,a,op2,arm_CPU_boolToU(c,sc)),false);
res = std::get<0>(tmp18);
logical = std::get<1>(tmp18);
break;}
case cast<uint32_t>(6ULL):{
auto tmp19 = std::make_tuple(arm_CPU_sub_(c,a,op2,arm_CPU_boolToU(c,sc)),false);
res = std::get<0>(tmp19);
logical = std::get<1>(tmp19);
break;}
case cast<uint32_t>(7ULL):{
auto tmp20 = std::make_tuple(arm_CPU_sub_(c,op2,a,arm_CPU_boolToU(c,sc)),false);
res = std::get<0>(tmp20);
logical = std::get<1>(tmp20);
break;}
case cast<uint32_t>(8ULL):{
res = cast<uint32_t>((a & op2));
break;}
case cast<uint32_t>(9ULL):{
res = cast<uint32_t>((a ^ op2));
break;}
case cast<uint32_t>(10ULL):{
arm_CPU_sub_(c,a,op2,cast<uint32_t>(1ULL));
logical = false;
break;}
case cast<uint32_t>(11ULL):{
arm_CPU_add(c,a,op2,cast<uint32_t>(0ULL));
logical = false;
break;}
case cast<uint32_t>(12ULL):{
res = cast<uint32_t>((a | op2));
break;}
case cast<uint32_t>(13ULL):{
res = op2;
break;}
case cast<uint32_t>(14ULL):{
res = (a & ~(op2));
break;}
default:{
res = cast<uint32_t>(~op2);
break;}
}}
bool isTest = ((op >= cast<uint32_t>(8ULL)) && (op <= cast<uint32_t>(11ULL)));
if (isTest) {
if (logical) {
arm_CPU_setNZ(c,res);
auto tmp21 = std::make_tuple((shc == cast<uint32_t>(1ULL)),sv);
c->C = std::get<0>(tmp21);
c->V = std::get<1>(tmp21);
}
return ;
}
if (setFlags) {
if (logical) {
arm_CPU_setNZ(c,res);
auto tmp22 = std::make_tuple((shc == cast<uint32_t>(1ULL)),sv);
c->C = std::get<0>(tmp22);
c->V = std::get<1>(tmp22);
}
}
else {
auto tmp23 = std::make_tuple(sn,sz,sc,sv);
c->N = std::get<0>(tmp23);
c->Z = std::get<1>(tmp23);
c->C = std::get<2>(tmp23);
c->V = std::get<3>(tmp23);
}
if ((rd == cast<uint32_t>(15ULL))) {
if (setFlags) {
arm_CPU_SetCPSR(c,arm_CPU_SPSR(c));
}
arm_CPU_branchTo(c,res);
return ;
}
arm_CPU_setReg(c,rd,res);
}
}
// tools/cpu/arm/armexec.go:223:1
void arm_CPU_branchTo(arm_CPU* c,uint32_t v){
{
if (c->Thumb) {
c->R[cast<int64_t>(15ULL)] = (v & ~(cast<uint32_t>(1ULL)));
}
else {
c->R[cast<int64_t>(15ULL)] = (v & ~(cast<uint32_t>(3ULL)));
}
c->branched = true;
}
}
// tools/cpu/arm/armexec.go:234:1
void arm_CPU_bxTo(arm_CPU* c,uint32_t v){
{
c->Thumb = (cast<uint32_t>((v & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL));
arm_CPU_branchTo(c,v);
}
}
// tools/cpu/arm/armexec.go:239:1
void arm_CPU_execBranch(arm_CPU* c,uint32_t w){
{
uint32_t off = shl<uint32_t>(arm_signExtend(cast<uint32_t>((w & cast<uint32_t>(16777215ULL))),cast<uint64_t>(24ULL)),cast<int64_t>(2ULL));
uint32_t target = cast<uint32_t>((cast<uint32_t>((c->cur + cast<uint32_t>(8ULL))) + off));
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(24ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
c->R[cast<int64_t>(14ULL)] = cast<uint32_t>((c->cur + cast<uint32_t>(4ULL)));
}
arm_CPU_branchTo(c,target);
}
}
// tools/cpu/arm/armexec.go:249:1
void arm_CPU_execExtension(arm_CPU* c,uint32_t w){
{
uint32_t sh = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(5ULL))) & cast<uint32_t>(3ULL)));
if ((sh == cast<uint32_t>(0ULL))) {
{
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(24ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))){
arm_CPU_execSwap(c,w);
}
else if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(23ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))){
arm_CPU_execMulLong(c,w);
}
else {
arm_CPU_execMul(c,w);
}
}
tmp24:;
return ;
}
arm_CPU_execHalf(c,w,sh);
}
}
// tools/cpu/arm/armexec.go:265:1
void arm_CPU_execMul(arm_CPU* c,uint32_t w){
{
uint32_t rd = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(15ULL)));
uint32_t rn = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(12ULL))) & cast<uint32_t>(15ULL)));
uint32_t rs = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(8ULL))) & cast<uint32_t>(15ULL)));
uint32_t rm = cast<uint32_t>((w & cast<uint32_t>(15ULL)));
uint32_t res = cast<uint32_t>((arm_CPU_reg(c,rm) * arm_CPU_reg(c,rs)));
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(21ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
res += arm_CPU_reg(c,rn);
}
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(20ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
arm_CPU_setNZ(c,res);
}
arm_CPU_setReg(c,rd,res);
}
}
// tools/cpu/arm/armexec.go:280:1
void arm_CPU_execMulLong(arm_CPU* c,uint32_t w){
{
uint32_t rdHi = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(15ULL)));
uint32_t rdLo = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(12ULL))) & cast<uint32_t>(15ULL)));
uint32_t rs = arm_CPU_reg(c,cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(8ULL))) & cast<uint32_t>(15ULL))));
uint32_t rm = arm_CPU_reg(c,cast<uint32_t>((w & cast<uint32_t>(15ULL))));
bool signed_ = (cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(22ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL));
bool accum = (cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(21ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL));
uint64_t result={};
if (signed_) {
result = cast<uint64_t>(cast<int64_t>((cast<int64_t>(cast<int32_t>(rm)) * cast<int64_t>(cast<int32_t>(rs)))));
}
else {
result = cast<uint64_t>((cast<uint64_t>(rm) * cast<uint64_t>(rs)));
}
if (accum) {
result += cast<uint64_t>((shl<uint64_t>(cast<uint64_t>(arm_CPU_reg(c,rdHi)),cast<int64_t>(32ULL)) | cast<uint64_t>(arm_CPU_reg(c,rdLo))));
}
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(20ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
c->Z = (result == cast<uint64_t>(0ULL));
c->N = (cast<uint64_t>((result & cast<uint64_t>(9223372036854775808ULL))) != cast<uint64_t>(0ULL));
}
arm_CPU_setReg(c,rdLo,cast<uint32_t>(result));
arm_CPU_setReg(c,rdHi,cast<uint32_t>(shr<uint64_t>(result,cast<int64_t>(32ULL))));
}
}
// tools/cpu/arm/armexec.go:305:1
void arm_CPU_execSwap(arm_CPU* c,uint32_t w){
{
uint32_t rn = arm_CPU_reg(c,cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(15ULL))));
uint32_t rd = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(12ULL))) & cast<uint32_t>(15ULL)));
uint32_t rm = cast<uint32_t>((w & cast<uint32_t>(15ULL)));
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(22ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
uint32_t old = cast<uint32_t>(arm_CPU_read8(c,rn));
arm_CPU_write8(c,rn,cast<uint8_t>(arm_CPU_reg(c,rm)));
arm_CPU_setReg(c,rd,old);
}
else {
uint32_t old = arm_CPU_read32(c,rn);
arm_CPU_write32aligned(c,rn,arm_CPU_reg(c,rm));
arm_CPU_setReg(c,rd,old);
}
}
}
// tools/cpu/arm/armexec.go:320:1
void arm_CPU_execHalf(arm_CPU* c,uint32_t w,uint32_t sh){
{
uint32_t l = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(20ULL))) & cast<uint32_t>(1ULL)));
uint32_t p = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(24ULL))) & cast<uint32_t>(1ULL)));
uint32_t u = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(23ULL))) & cast<uint32_t>(1ULL)));
uint32_t wbit = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(21ULL))) & cast<uint32_t>(1ULL)));
uint32_t rn = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(15ULL)));
uint32_t rd = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(12ULL))) & cast<uint32_t>(15ULL)));
uint32_t base = arm_CPU_reg(c,rn);
uint32_t off={};
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(22ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
off = cast<uint32_t>((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(4ULL))) & cast<uint32_t>(240ULL))) | (cast<uint32_t>((w & cast<uint32_t>(15ULL))))));
}
else {
off = arm_CPU_reg(c,cast<uint32_t>((w & cast<uint32_t>(15ULL))));
}
uint32_t addr = base;
if ((p == cast<uint32_t>(1ULL))) {
if ((u == cast<uint32_t>(1ULL))) {
addr += off;
}
else {
addr -= off;
}
}
auto writeback = [&]()->void{
if ((p == cast<uint32_t>(0ULL))) {
if ((u == cast<uint32_t>(1ULL))) {
base += off;
}
else {
base -= off;
}
arm_CPU_setReg(c,rn,base);
}
else if ((wbit == cast<uint32_t>(1ULL))) {
arm_CPU_setReg(c,rn,addr);
}
}
;
{
if (((l == cast<uint32_t>(1ULL)) && (sh == cast<uint32_t>(1ULL)))){
uint32_t v = arm_CPU_read16(c,addr);
writeback();
arm_CPU_setReg(c,rd,v);
}
else if (((l == cast<uint32_t>(1ULL)) && (sh == cast<uint32_t>(2ULL)))){
uint32_t v = cast<uint32_t>(cast<int32_t>(cast<int8_t>(arm_CPU_read8(c,addr))));
writeback();
arm_CPU_setReg(c,rd,v);
}
else if (((l == cast<uint32_t>(1ULL)) && (sh == cast<uint32_t>(3ULL)))){
uint32_t v = cast<uint32_t>(cast<int32_t>(cast<int16_t>(arm_CPU_read16(c,addr))));
writeback();
arm_CPU_setReg(c,rd,v);
}
else if (((l == cast<uint32_t>(0ULL)) && (sh == cast<uint32_t>(1ULL)))){
writeback();
arm_CPU_write16(c,addr,arm_CPU_reg(c,rd));
}
else if (((l == cast<uint32_t>(0ULL)) && (sh == cast<uint32_t>(2ULL)))){
if ((!arm_Variant_v5OrLater(c->Arch))) {
arm_CPU_undefV4T(c,w,std::string("LDRD",4));
return ;
}
uint32_t lo = arm_CPU_read32(c,addr);
uint32_t hi = arm_CPU_read32(c,cast<uint32_t>((addr + cast<uint32_t>(4ULL))));
writeback();
arm_CPU_setReg(c,rd,lo);
arm_CPU_setReg(c,cast<uint32_t>((rd + cast<uint32_t>(1ULL))),hi);
}
else {
if ((!arm_Variant_v5OrLater(c->Arch))) {
arm_CPU_undefV4T(c,w,std::string("STRD",4));
return ;
}
uint32_t lo = arm_CPU_reg(c,rd);
uint32_t hi = arm_CPU_reg(c,cast<uint32_t>((rd + cast<uint32_t>(1ULL))));
writeback();
arm_CPU_write32(c,addr,lo);
arm_CPU_write32(c,cast<uint32_t>((addr + cast<uint32_t>(4ULL))),hi);
}
}
tmp25:;
}
}
// tools/cpu/arm/armexec.go:399:1
void arm_CPU_execSingle(arm_CPU* c,uint32_t w,bool regOff){
{
uint32_t l = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(20ULL))) & cast<uint32_t>(1ULL)));
uint32_t b = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(22ULL))) & cast<uint32_t>(1ULL)));
uint32_t p = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(24ULL))) & cast<uint32_t>(1ULL)));
uint32_t u = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(23ULL))) & cast<uint32_t>(1ULL)));
uint32_t wbit = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(21ULL))) & cast<uint32_t>(1ULL)));
uint32_t rn = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(15ULL)));
uint32_t rd = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(12ULL))) & cast<uint32_t>(15ULL)));
uint32_t base = arm_CPU_reg(c,rn);
uint32_t off={};
if ((!regOff)) {
off = cast<uint32_t>((w & cast<uint32_t>(4095ULL)));
}
else {
auto tmp26 = arm_CPU_shift(c,cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(5ULL))) & cast<uint32_t>(3ULL))),cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(7ULL))) & cast<uint32_t>(31ULL))),arm_CPU_reg(c,cast<uint32_t>((w & cast<uint32_t>(15ULL)))),false,arm_CPU_boolToU(c,c->C));
off = std::get<0>(tmp26);
}
uint32_t addr = base;
if ((p == cast<uint32_t>(1ULL))) {
if ((u == cast<uint32_t>(1ULL))) {
addr += off;
}
else {
addr -= off;
}
}
auto writeback = [&]()->void{
if ((p == cast<uint32_t>(0ULL))) {
if ((u == cast<uint32_t>(1ULL))) {
base += off;
}
else {
base -= off;
}
arm_CPU_setReg(c,rn,base);
}
else if ((wbit == cast<uint32_t>(1ULL))) {
arm_CPU_setReg(c,rn,addr);
}
}
;
if ((l == cast<uint32_t>(1ULL))) {
uint32_t v={};
if ((b == cast<uint32_t>(1ULL))) {
v = cast<uint32_t>(arm_CPU_read8(c,addr));
}
else {
v = arm_CPU_read32(c,addr);
}
writeback();
if ((rd == cast<uint32_t>(15ULL))) {
arm_CPU_bxTo(c,v);
}
else {
arm_CPU_setReg(c,rd,v);
}
}
else {
uint32_t v = arm_CPU_reg(c,rd);
if ((rd == cast<uint32_t>(15ULL))) {
v = cast<uint32_t>((c->cur + cast<uint32_t>(12ULL)));
}
if ((b == cast<uint32_t>(1ULL))) {
arm_CPU_write8(c,addr,cast<uint8_t>(v));
}
else {
arm_CPU_write32(c,addr,v);
}
writeback();
}
}
}
// tools/cpu/arm/armexec.go:462:1
void arm_CPU_execBlock(arm_CPU* c,uint32_t w){
{
uint32_t l = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(20ULL))) & cast<uint32_t>(1ULL)));
uint32_t p = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(24ULL))) & cast<uint32_t>(1ULL)));
uint32_t u = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(23ULL))) & cast<uint32_t>(1ULL)));
uint32_t wbit = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(21ULL))) & cast<uint32_t>(1ULL)));
uint32_t s = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(22ULL))) & cast<uint32_t>(1ULL)));
uint32_t rn = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(15ULL)));
uint32_t mask = cast<uint32_t>((w & cast<uint32_t>(65535ULL)));
uint32_t base = c->R[rn];
uint32_t n = cast<uint32_t>(go_bits_OnesCount32(mask));
uint32_t start={};
uint32_t final={};
if ((u == cast<uint32_t>(1ULL))) {
final = cast<uint32_t>((base + cast<uint32_t>((n * cast<uint32_t>(4ULL)))));
if ((p == cast<uint32_t>(1ULL))) {
start = cast<uint32_t>((base + cast<uint32_t>(4ULL)));
}
else {
start = base;
}
}
else {
final = cast<uint32_t>((base - cast<uint32_t>((n * cast<uint32_t>(4ULL)))));
if ((p == cast<uint32_t>(1ULL))) {
start = cast<uint32_t>((base - cast<uint32_t>((n * cast<uint32_t>(4ULL)))));
}
else {
start = cast<uint32_t>((cast<uint32_t>((base - cast<uint32_t>((n * cast<uint32_t>(4ULL))))) + cast<uint32_t>(4ULL)));
}
}
bool userBank = ((s == cast<uint32_t>(1ULL)) && (cast<uint32_t>((mask & cast<uint32_t>(32768ULL))) == cast<uint32_t>(0ULL)));
uint32_t addr = start;
bool loadedRn = false;
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < cast<uint32_t>(16ULL));i++){
if ((cast<uint32_t>((mask & (shl<uint32_t>(cast<uint32_t>(1ULL),i)))) == cast<uint32_t>(0ULL))) {
continue;
}
if ((l == cast<uint32_t>(1ULL))) {
uint32_t v = arm_CPU_read32aligned(c,addr);
{
if ((i == cast<uint32_t>(15ULL))){
if ((s == cast<uint32_t>(1ULL))) {
arm_CPU_SetCPSR(c,arm_CPU_SPSR(c));
}
arm_CPU_bxTo(c,v);
}
else if (userBank){
arm_CPU_setUserReg(c,i,v);
}
else {
c->R[i] = v;
if ((i == rn)) {
loadedRn = true;
}
}
}
tmp27:;
}
else {
uint32_t v={};
{
if ((i == cast<uint32_t>(15ULL))){
v = cast<uint32_t>((c->cur + cast<uint32_t>(12ULL)));
}
else if (userBank){
v = arm_CPU_userReg(c,i);
}
else {
v = c->R[i];
}
}
tmp28:;
arm_CPU_write32aligned(c,addr,v);
}
addr += cast<uint32_t>(4ULL);
}
}if (((wbit == cast<uint32_t>(1ULL)) && (!(((l == cast<uint32_t>(1ULL)) && loadedRn))))) {
arm_CPU_setReg(c,rn,final);
}
}
}
// tools/cpu/arm/armexec.go:556:1
bool arm_CPU_inUserBank(arm_CPU* c){
{
return ((c->Mode == cast<uint32_t>(16ULL)) || (c->Mode == cast<uint32_t>(31ULL)));
}
}
// tools/cpu/arm/armexec.go:558:1
uint32_t arm_CPU_userReg(arm_CPU* c,uint32_t i){
{
{
if ((((i >= cast<uint32_t>(8ULL)) && (i <= cast<uint32_t>(12ULL))) && (c->Mode == cast<uint32_t>(17ULL)))){
return c->usrR8_12[cast<uint32_t>((i - cast<uint32_t>(8ULL)))];
}
else if (((i == cast<uint32_t>(13ULL)) && (!arm_CPU_inUserBank(c)))){
return c->bankR13[cast<int64_t>(0ULL)];
}
else if (((i == cast<uint32_t>(14ULL)) && (!arm_CPU_inUserBank(c)))){
return c->bankR14[cast<int64_t>(0ULL)];
}
}
tmp29:;
return c->R[i];
}
}
// tools/cpu/arm/armexec.go:570:1
void arm_CPU_setUserReg(arm_CPU* c,uint32_t i,uint32_t v){
{
{
if ((((i >= cast<uint32_t>(8ULL)) && (i <= cast<uint32_t>(12ULL))) && (c->Mode == cast<uint32_t>(17ULL)))){
c->usrR8_12[cast<uint32_t>((i - cast<uint32_t>(8ULL)))] = v;
}
else if (((i == cast<uint32_t>(13ULL)) && (!arm_CPU_inUserBank(c)))){
c->bankR13[cast<int64_t>(0ULL)] = v;
}
else if (((i == cast<uint32_t>(14ULL)) && (!arm_CPU_inUserBank(c)))){
c->bankR14[cast<int64_t>(0ULL)] = v;
}
else {
c->R[i] = v;
}
}
tmp30:;
}
}
// tools/cpu/arm/cpu.go:128:1
void arm_CPU_Reset(arm_CPU* c){
{
c->R = std::array<uint32_t,16>{};
auto tmp31 = std::make_tuple(false,false,false,false,false);
c->N = std::get<0>(tmp31);
c->Z = std::get<1>(tmp31);
c->C = std::get<2>(tmp31);
c->V = std::get<3>(tmp31);
c->Q = std::get<4>(tmp31);
c->Thumb = false;
auto tmp32 = std::make_tuple(true,true);
c->IRQDisable = std::get<0>(tmp32);
c->FIQDisable = std::get<1>(tmp32);
c->Mode = cast<uint32_t>(19ULL);
auto tmp33 = std::make_tuple(false,std::string("",0));
c->Halted = std::get<0>(tmp33);
c->HaltReason = std::get<1>(tmp33);
}
}
// tools/cpu/arm/cpu.go:146:1
uint32_t arm_CPU_PC(arm_CPU* c){
{
return c->R[cast<int64_t>(15ULL)];
}
}
// tools/cpu/arm/cpu.go:151:1
uint32_t arm_CPU_CPSR(arm_CPU* c){
{
uint32_t v={};
if (c->N) {
v |= cast<uint32_t>(2147483648ULL);
}
if (c->Z) {
v |= cast<uint32_t>(1073741824ULL);
}
if (c->C) {
v |= cast<uint32_t>(536870912ULL);
}
if (c->V) {
v |= cast<uint32_t>(268435456ULL);
}
if (c->Q) {
v |= cast<uint32_t>(134217728ULL);
}
v |= shl<uint32_t>((cast<uint32_t>((c->GE & cast<uint32_t>(15ULL)))),cast<int64_t>(16ULL));
if (c->BigEndian) {
v |= cast<uint32_t>(512ULL);
}
if (c->IRQDisable) {
v |= cast<uint32_t>(128ULL);
}
if (c->FIQDisable) {
v |= cast<uint32_t>(64ULL);
}
if (c->Thumb) {
v |= cast<uint32_t>(32ULL);
}
return cast<uint32_t>((v | (cast<uint32_t>((c->Mode & cast<uint32_t>(31ULL))))));
}
}
// tools/cpu/arm/cpu.go:185:1
void arm_CPU_SetCPSR(arm_CPU* c,uint32_t v){
{
c->N = (cast<uint32_t>((v & cast<uint32_t>(2147483648ULL))) != cast<uint32_t>(0ULL));
c->Z = (cast<uint32_t>((v & cast<uint32_t>(1073741824ULL))) != cast<uint32_t>(0ULL));
c->C = (cast<uint32_t>((v & cast<uint32_t>(536870912ULL))) != cast<uint32_t>(0ULL));
c->V = (cast<uint32_t>((v & cast<uint32_t>(268435456ULL))) != cast<uint32_t>(0ULL));
c->Q = (cast<uint32_t>((v & cast<uint32_t>(134217728ULL))) != cast<uint32_t>(0ULL));
c->GE = cast<uint32_t>(((shr<uint32_t>(v,cast<int64_t>(16ULL))) & cast<uint32_t>(15ULL)));
c->BigEndian = (cast<uint32_t>((v & cast<uint32_t>(512ULL))) != cast<uint32_t>(0ULL));
c->IRQDisable = (cast<uint32_t>((v & cast<uint32_t>(128ULL))) != cast<uint32_t>(0ULL));
c->FIQDisable = (cast<uint32_t>((v & cast<uint32_t>(64ULL))) != cast<uint32_t>(0ULL));
c->Thumb = (cast<uint32_t>((v & cast<uint32_t>(32ULL))) != cast<uint32_t>(0ULL));
arm_CPU_switchMode(c,cast<uint32_t>((v & cast<uint32_t>(31ULL))));
}
}
// tools/cpu/arm/cpu.go:200:1
int64_t arm_modeIndex(uint32_t mode){
{
{
switch(mode){
case cast<uint32_t>(17ULL):{
return cast<int64_t>(1ULL);
break;}
case cast<uint32_t>(18ULL):{
return cast<int64_t>(2ULL);
break;}
case cast<uint32_t>(19ULL):{
return cast<int64_t>(3ULL);
break;}
case cast<uint32_t>(23ULL):{
return cast<int64_t>(4ULL);
break;}
case cast<uint32_t>(27ULL):{
return cast<int64_t>(5ULL);
break;}
default:{
return cast<int64_t>(0ULL);
break;}
}}
}
}
// tools/cpu/arm/cpu.go:218:1
void arm_CPU_switchMode(arm_CPU* c,uint32_t mode){
{
mode &= cast<uint32_t>(31ULL);
if ((mode == c->Mode)) {
return ;
}
auto tmp34 = std::make_tuple(arm_modeIndex(c->Mode),arm_modeIndex(mode));
int64_t from = std::get<0>(tmp34);
int64_t to = std::get<1>(tmp34);
c->bankR13[from] = c->R[cast<int64_t>(13ULL)];
c->bankR14[from] = c->R[cast<int64_t>(14ULL)];
if ((c->Mode == cast<uint32_t>(17ULL))) {
gcopy(sub(c->fiqR8_12,0,len(c->fiqR8_12)),sub(c->R,cast<int64_t>(8ULL),cast<int64_t>(13ULL)));
}
else {
gcopy(sub(c->usrR8_12,0,len(c->usrR8_12)),sub(c->R,cast<int64_t>(8ULL),cast<int64_t>(13ULL)));
}
c->R[cast<int64_t>(13ULL)] = c->bankR13[to];
c->R[cast<int64_t>(14ULL)] = c->bankR14[to];
if ((mode == cast<uint32_t>(17ULL))) {
gcopy(sub(c->R,cast<int64_t>(8ULL),cast<int64_t>(13ULL)),sub(c->fiqR8_12,0,len(c->fiqR8_12)));
}
else {
gcopy(sub(c->R,cast<int64_t>(8ULL),cast<int64_t>(13ULL)),sub(c->usrR8_12,0,len(c->usrR8_12)));
}
c->Mode = mode;
}
}
// tools/cpu/arm/cpu.go:245:1
uint32_t arm_CPU_SPSR(arm_CPU* c){
{
return c->bankSPSR[arm_modeIndex(c->Mode)];
}
}
// tools/cpu/arm/cpu.go:248:1
void arm_CPU_SetSPSR(arm_CPU* c,uint32_t v){
{
c->bankSPSR[arm_modeIndex(c->Mode)] = v;
}
}
// tools/cpu/arm/cpu.go:252:1
uint8_t arm_CPU_read8(arm_CPU* c,uint32_t a){
{
return gbamachine_bus_Read(c->bus,a);
}
}
// tools/cpu/arm/cpu.go:253:1
void arm_CPU_write8(arm_CPU* c,uint32_t a,uint8_t v){
{
gbamachine_bus_Write(c->bus,a,v);
}
}
// tools/cpu/arm/cpu.go:269:1
uint32_t arm_CPU_read32(arm_CPU* c,uint32_t a){
{
if (arm_Variant_isV6(c->Arch)) {
if ((bool(c->wide) && (cast<uint32_t>((a & cast<uint32_t>(3ULL))) == cast<uint32_t>(0ULL)))) {
return gbamachine_bus_Read32(c->wide,a);
}
return cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(gbamachine_bus_Read(c->bus,a)) | shl<uint32_t>(cast<uint32_t>(gbamachine_bus_Read(c->bus,cast<uint32_t>((a + cast<uint32_t>(1ULL))))),cast<int64_t>(8ULL)))) | shl<uint32_t>(cast<uint32_t>(gbamachine_bus_Read(c->bus,cast<uint32_t>((a + cast<uint32_t>(2ULL))))),cast<int64_t>(16ULL)))) | shl<uint32_t>(cast<uint32_t>(gbamachine_bus_Read(c->bus,cast<uint32_t>((a + cast<uint32_t>(3ULL))))),cast<int64_t>(24ULL))));
}
uint32_t aligned = (a & ~(cast<uint32_t>(3ULL)));
uint32_t v = arm_CPU_read32aligned(c,aligned);
{
uint32_t r = cast<uint32_t>(((cast<uint32_t>((a & cast<uint32_t>(3ULL)))) * cast<uint32_t>(8ULL)));
if ((r != cast<uint32_t>(0ULL))) {
v = arm_ror32(v,r);
}
}
return v;
}
}
// tools/cpu/arm/cpu.go:331:1
void arm_CPU_write32(arm_CPU* c,uint32_t a,uint32_t v){
{
if ((!arm_Variant_isV6(c->Arch))) {
a &= ~(cast<uint32_t>(3ULL));
}
if ((bool(c->wide) && (cast<uint32_t>((a & cast<uint32_t>(3ULL))) == cast<uint32_t>(0ULL)))) {
gbamachine_bus_Write32(c->wide,a,v);
return ;
}
gbamachine_bus_Write(c->bus,a,cast<uint8_t>(v));
gbamachine_bus_Write(c->bus,cast<uint32_t>((a + cast<uint32_t>(1ULL))),cast<uint8_t>(shr<uint32_t>(v,cast<int64_t>(8ULL))));
gbamachine_bus_Write(c->bus,cast<uint32_t>((a + cast<uint32_t>(2ULL))),cast<uint8_t>(shr<uint32_t>(v,cast<int64_t>(16ULL))));
gbamachine_bus_Write(c->bus,cast<uint32_t>((a + cast<uint32_t>(3ULL))),cast<uint8_t>(shr<uint32_t>(v,cast<int64_t>(24ULL))));
}
}
// tools/cpu/arm/cpu.go:349:1
uint32_t arm_CPU_reg(arm_CPU* c,uint32_t i){
{
if ((i == cast<uint32_t>(15ULL))) {
if (c->Thumb) {
return cast<uint32_t>((c->cur + cast<uint32_t>(4ULL)));
}
return cast<uint32_t>((c->cur + cast<uint32_t>(8ULL)));
}
return c->R[i];
}
}
// tools/cpu/arm/cpu.go:360:1
void arm_CPU_setReg(arm_CPU* c,uint32_t i,uint32_t v){
{
c->R[i] = v;
if ((i == cast<uint32_t>(15ULL))) {
c->branched = true;
}
}
}
// tools/cpu/arm/cpu.go:369:1
void arm_CPU_setNZ(arm_CPU* c,uint32_t v){
{
c->Z = (v == cast<uint32_t>(0ULL));
c->N = (cast<uint32_t>((v & cast<uint32_t>(2147483648ULL))) != cast<uint32_t>(0ULL));
}
}
// tools/cpu/arm/cpu.go:375:1
uint32_t arm_CPU_add(arm_CPU* c,uint32_t a,uint32_t b,uint32_t cin){
{
uint64_t r = cast<uint64_t>((cast<uint64_t>((cast<uint64_t>(a) + cast<uint64_t>(b))) + cast<uint64_t>(cin)));
uint32_t res = cast<uint32_t>(r);
arm_CPU_setNZ(c,res);
c->C = (r > cast<uint64_t>(4294967295ULL));
c->V = (cast<uint32_t>((cast<uint32_t>(((cast<uint32_t>((a ^ res))) & (cast<uint32_t>((b ^ res))))) & cast<uint32_t>(2147483648ULL))) != cast<uint32_t>(0ULL));
return res;
}
}
// tools/cpu/arm/cpu.go:385:1
uint32_t arm_CPU_sub_(arm_CPU* c,uint32_t a,uint32_t b,uint32_t cin){
{
return arm_CPU_add(c,a,cast<uint32_t>(~b),cin);
}
}
// tools/cpu/arm/cpu.go:388:1
bool arm_CPU_cond(arm_CPU* c,int64_t cc){
{
{
switch(cc){
case cast<int64_t>(0ULL):{
return c->Z;
break;}
case cast<int64_t>(1ULL):{
return (!c->Z);
break;}
case cast<int64_t>(2ULL):{
return c->C;
break;}
case cast<int64_t>(3ULL):{
return (!c->C);
break;}
case cast<int64_t>(4ULL):{
return c->N;
break;}
case cast<int64_t>(5ULL):{
return (!c->N);
break;}
case cast<int64_t>(6ULL):{
return c->V;
break;}
case cast<int64_t>(7ULL):{
return (!c->V);
break;}
case cast<int64_t>(8ULL):{
return (c->C && (!c->Z));
break;}
case cast<int64_t>(9ULL):{
return ((!c->C) || c->Z);
break;}
case cast<int64_t>(10ULL):{
return (c->N == c->V);
break;}
case cast<int64_t>(11ULL):{
return (c->N != c->V);
break;}
case cast<int64_t>(12ULL):{
return ((!c->Z) && (c->N == c->V));
break;}
case cast<int64_t>(13ULL):{
return (c->Z || (c->N != c->V));
break;}
default:{
return true;
break;}
}}
}
}
// tools/cpu/arm/cpu.go:429:1
std::tuple<uint32_t,uint32_t> arm_CPU_shift(arm_CPU* c,uint32_t typ,uint32_t amt,uint32_t val,bool regForm,uint32_t cin){
{
if (regForm) {
amt &= cast<uint32_t>(255ULL);
if ((amt == cast<uint32_t>(0ULL))) {
return {val,cin};
}
}
{
switch(typ){
case cast<uint32_t>(0ULL):{
{
if ((amt == cast<uint32_t>(0ULL))){
return {val,cin};
}
else if ((amt < cast<uint32_t>(32ULL))){
return {shl<uint32_t>(val,amt),cast<uint32_t>(((shr<uint32_t>(val,(cast<uint32_t>((cast<uint32_t>(32ULL) - amt))))) & cast<uint32_t>(1ULL)))};
}
else if ((amt == cast<uint32_t>(32ULL))){
return {cast<uint32_t>(0ULL),cast<uint32_t>((val & cast<uint32_t>(1ULL)))};
}
else {
return {cast<uint32_t>(0ULL),cast<uint32_t>(0ULL)};
}
}
tmp35:;
break;}
case cast<uint32_t>(1ULL):{
if (((amt == cast<uint32_t>(0ULL)) && (!regForm))) {
amt = cast<uint32_t>(32ULL);
}
{
if ((amt == cast<uint32_t>(0ULL))){
return {val,cin};
}
else if ((amt < cast<uint32_t>(32ULL))){
return {shr<uint32_t>(val,amt),cast<uint32_t>(((shr<uint32_t>(val,(cast<uint32_t>((amt - cast<uint32_t>(1ULL)))))) & cast<uint32_t>(1ULL)))};
}
else if ((amt == cast<uint32_t>(32ULL))){
return {cast<uint32_t>(0ULL),cast<uint32_t>(((shr<uint32_t>(val,cast<int64_t>(31ULL))) & cast<uint32_t>(1ULL)))};
}
else {
return {cast<uint32_t>(0ULL),cast<uint32_t>(0ULL)};
}
}
tmp36:;
break;}
case cast<uint32_t>(2ULL):{
if (((amt == cast<uint32_t>(0ULL)) && (!regForm))) {
amt = cast<uint32_t>(32ULL);
}
int32_t sv = cast<int32_t>(val);
{
if ((amt == cast<uint32_t>(0ULL))){
return {val,cin};
}
else if ((amt < cast<uint32_t>(32ULL))){
return {cast<uint32_t>(shr<int32_t>(sv,amt)),cast<uint32_t>((cast<uint32_t>(shr<uint32_t>(val,(cast<uint32_t>((amt - cast<uint32_t>(1ULL)))))) & cast<uint32_t>(1ULL)))};
}
else {
if ((cast<uint32_t>((val & cast<uint32_t>(2147483648ULL))) != cast<uint32_t>(0ULL))) {
return {cast<uint32_t>(4294967295ULL),cast<uint32_t>(1ULL)};
}
return {cast<uint32_t>(0ULL),cast<uint32_t>(0ULL)};
}
}
tmp37:;
break;}
default:{
if (((amt == cast<uint32_t>(0ULL)) && (!regForm))) {
return {cast<uint32_t>((shl<uint32_t>(cin,cast<int64_t>(31ULL)) | shr<uint32_t>(val,cast<int64_t>(1ULL)))),cast<uint32_t>((val & cast<uint32_t>(1ULL)))};
}
if ((amt == cast<uint32_t>(0ULL))) {
return {val,cin};
}
amt &= cast<uint32_t>(31ULL);
if ((amt == cast<uint32_t>(0ULL))) {
return {val,cast<uint32_t>(((shr<uint32_t>(val,cast<int64_t>(31ULL))) & cast<uint32_t>(1ULL)))};
}
return {arm_ror32(val,amt),cast<uint32_t>(((shr<uint32_t>(val,(cast<uint32_t>((amt - cast<uint32_t>(1ULL)))))) & cast<uint32_t>(1ULL)))};
break;}
}}
}
}
// tools/cpu/arm/cpu.go:497:1
void arm_CPU_ClearExclusive(arm_CPU* c){
{
c->exclValid = false;
}
}
// tools/cpu/arm/cpu.go:519:1
arm_Banks arm_CPU_SaveBanks(arm_CPU* c){
{
arm_Banks b = arm_Banks{c->bankR13,c->bankR14,c->bankSPSR,c->fiqR8_12,c->usrR8_12};
int64_t i = arm_modeIndex(c->Mode);
auto tmp38 = std::make_tuple(c->R[cast<int64_t>(13ULL)],c->R[cast<int64_t>(14ULL)]);
b.R13[i] = std::get<0>(tmp38);
b.R14[i] = std::get<1>(tmp38);
if ((c->Mode == cast<uint32_t>(17ULL))) {
gcopy(sub(b.FIQR8_12,0,len(b.FIQR8_12)),sub(c->R,cast<int64_t>(8ULL),cast<int64_t>(13ULL)));
}
else {
gcopy(sub(b.USRR8_12,0,len(b.USRR8_12)),sub(c->R,cast<int64_t>(8ULL),cast<int64_t>(13ULL)));
}
return b;
}
}
// tools/cpu/arm/cpu.go:534:1
void arm_CPU_RestoreBanks(arm_CPU* c,arm_Banks b){
{
auto tmp39 = std::make_tuple(b.R13,b.R14,b.SPSR);
c->bankR13 = std::get<0>(tmp39);
c->bankR14 = std::get<1>(tmp39);
c->bankSPSR = std::get<2>(tmp39);
auto tmp40 = std::make_tuple(b.FIQR8_12,b.USRR8_12);
c->fiqR8_12 = std::get<0>(tmp40);
c->usrR8_12 = std::get<1>(tmp40);
}
}
// tools/cpu/arm/decode_v6.go:10:1
std::tuple<arm_Inst,bool> arm_decodeARMv6(uint32_t w,uint32_t addr,arm_Inst in){
{
int64_t cond = cast<int64_t>(shr<uint32_t>(w,cast<int64_t>(28ULL)));
if ((arm_isVFP(w) && (((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(25ULL))) & cast<uint32_t>(7ULL))) == cast<uint32_t>(6ULL)) || (cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(24ULL))) & cast<uint32_t>(15ULL))) == cast<uint32_t>(14ULL)))))) {
return {arm_decodeVFP(w,in),true};
}
if ((cond == cast<int64_t>(15ULL))) {
{
auto tmp41 = arm_decodeV6Uncond(w,in);
arm_Inst out = std::get<0>(tmp41);
bool ok = std::get<1>(tmp41);
if (ok) {
return {out,true};
}
}
return {in,false};
}
if ((((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(24ULL))) & cast<uint32_t>(15ULL))) == cast<uint32_t>(1ULL)) && (cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(4ULL))) & cast<uint32_t>(15ULL))) == cast<uint32_t>(9ULL))) && (cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(23ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL)))) {
return {arm_decodeSync(w,in),true};
}
if ((((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(24ULL))) & cast<uint32_t>(15ULL))) == cast<uint32_t>(0ULL)) && (cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(4ULL))) & cast<uint32_t>(15ULL))) == cast<uint32_t>(9ULL))) && (cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(20ULL))) & cast<uint32_t>(15ULL))) == cast<uint32_t>(4ULL)))) {
in.Mnem = (std::string("UMAAL",5) + arm_cn(cond));
in.Text = go_fmt_Sprintf(std::string("%s %s, %s, %s, %s",17),in.Mnem,arm_regName[cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(12ULL))) & cast<uint32_t>(15ULL)))],arm_regName[cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(15ULL)))],arm_regName[cast<uint32_t>((w & cast<uint32_t>(15ULL)))],arm_regName[cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(8ULL))) & cast<uint32_t>(15ULL)))]);
return {in,true};
}
if (((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(25ULL))) & cast<uint32_t>(7ULL))) == cast<uint32_t>(3ULL)) && (cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(4ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL)))) {
return {arm_decodeMedia(w,in),true};
}
return {in,false};
}
}
// tools/cpu/arm/decode_v6.go:52:1
arm_Inst arm_decodeSync(uint32_t w,arm_Inst in){
{
bool load = (cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(20ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL));
uint32_t sz = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(21ULL))) & cast<uint32_t>(3ULL)));
std::string suffix = std::array<std::string,4>{std::string("",0),std::string("D",1),std::string("B",1),std::string("H",1)}[sz];
std::string rn = arm_regName[cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(15ULL)))];
if (load) {
in.Mnem = ((std::string("LDREX",5) + suffix) + arm_cn(in.Cond));
uint32_t rt = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(12ULL))) & cast<uint32_t>(15ULL)));
if ((sz == cast<uint32_t>(1ULL))) {
in.Text = go_fmt_Sprintf(std::string("%s %s, %s, [%s]",15),in.Mnem,arm_regName[rt],arm_regName[cast<uint32_t>(((cast<uint32_t>((rt + cast<uint32_t>(1ULL)))) & cast<uint32_t>(15ULL)))],rn);
}
else {
in.Text = go_fmt_Sprintf(std::string("%s %s, [%s]",11),in.Mnem,arm_regName[rt],rn);
}
}
else {
in.Mnem = ((std::string("STREX",5) + suffix) + arm_cn(in.Cond));
uint32_t rt = cast<uint32_t>((w & cast<uint32_t>(15ULL)));
if ((sz == cast<uint32_t>(1ULL))) {
in.Text = go_fmt_Sprintf(std::string("%s %s, %s, %s, [%s]",19),in.Mnem,arm_regName[cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(12ULL))) & cast<uint32_t>(15ULL)))],arm_regName[rt],arm_regName[cast<uint32_t>(((cast<uint32_t>((rt + cast<uint32_t>(1ULL)))) & cast<uint32_t>(15ULL)))],rn);
}
else {
in.Text = go_fmt_Sprintf(std::string("%s %s, %s, [%s]",15),in.Mnem,arm_regName[cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(12ULL))) & cast<uint32_t>(15ULL)))],arm_regName[rt],rn);
}
}
return in;
}
}
// tools/cpu/arm/decode_v6.go:82:1
arm_Inst arm_decodeMedia(uint32_t w,arm_Inst in){
{
uint32_t op1 = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(20ULL))) & cast<uint32_t>(31ULL)));
uint32_t op2 = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(5ULL))) & cast<uint32_t>(7ULL)));
std::string rd = arm_regName[cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(12ULL))) & cast<uint32_t>(15ULL)))];
std::string rn = arm_regName[cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(15ULL)))];
std::string rm = arm_regName[cast<uint32_t>((w & cast<uint32_t>(15ULL)))];
{
if (((shr<uint32_t>(op1,cast<int64_t>(3ULL)) == cast<uint32_t>(0ULL)) && (cast<uint32_t>((op1 & cast<uint32_t>(7ULL))) != cast<uint32_t>(0ULL)))){
{
std::string name = arm_parallelName(cast<uint32_t>((op1 & cast<uint32_t>(7ULL))),op2);
if ((name != std::string("",0))) {
in.Mnem = (name + arm_cn(in.Cond));
in.Text = go_fmt_Sprintf(std::string("%s %s, %s, %s",13),in.Mnem,rd,rn,rm);
return in;
}
}
}
else if ((shr<uint32_t>(op1,cast<int64_t>(3ULL)) == cast<uint32_t>(1ULL))){
return arm_decodeMediaPack(w,in);
}
else if ((shr<uint32_t>(op1,cast<int64_t>(3ULL)) == cast<uint32_t>(2ULL))){
return arm_decodeMediaMul(w,in);
}
else if ((op1 == cast<uint32_t>(24ULL))){
uint32_t ra = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(12ULL))) & cast<uint32_t>(15ULL)));
if ((ra == cast<uint32_t>(15ULL))) {
in.Mnem = (std::string("USAD8",5) + arm_cn(in.Cond));
in.Text = go_fmt_Sprintf(std::string("%s %s, %s, %s",13),in.Mnem,rn,rm,arm_regName[cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(8ULL))) & cast<uint32_t>(15ULL)))]);
}
else {
in.Mnem = (std::string("USADA8",6) + arm_cn(in.Cond));
in.Text = go_fmt_Sprintf(std::string("%s %s, %s, %s, %s",17),in.Mnem,rn,rm,arm_regName[cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(8ULL))) & cast<uint32_t>(15ULL)))],arm_regName[ra]);
}
return in;
}
}
tmp42:;
return arm_undef(w,in);
}
}
// tools/cpu/arm/decode_v6.go:116:1
std::string arm_parallelName(uint32_t class_,uint32_t op2){
{
std::string prefix = get(Map<uint32_t,std::string>{{cast<uint32_t>(1ULL),std::string("S",1)},{cast<uint32_t>(2ULL),std::string("Q",1)},{cast<uint32_t>(3ULL),std::string("SH",2)},{cast<uint32_t>(5ULL),std::string("U",1)},{cast<uint32_t>(6ULL),std::string("UQ",2)},{cast<uint32_t>(7ULL),std::string("UH",2)}},class_);
std::string op = get(Map<uint32_t,std::string>{{cast<uint32_t>(0ULL),std::string("ADD16",5)},{cast<uint32_t>(1ULL),std::string("ASX",3)},{cast<uint32_t>(2ULL),std::string("SAX",3)},{cast<uint32_t>(3ULL),std::string("SUB16",5)},{cast<uint32_t>(4ULL),std::string("ADD8",4)},{cast<uint32_t>(7ULL),std::string("SUB8",4)}},op2);
if (((prefix == std::string("",0)) || (op == std::string("",0)))) {
return std::string("",0);
}
return (prefix + op);
}
}
// tools/cpu/arm/decode_v6.go:127:1
arm_Inst arm_decodeMediaPack(uint32_t w,arm_Inst in){
{
uint32_t op1 = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(20ULL))) & cast<uint32_t>(31ULL)));
uint32_t op2 = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(5ULL))) & cast<uint32_t>(7ULL)));
std::string rd = arm_regName[cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(12ULL))) & cast<uint32_t>(15ULL)))];
std::string rn = arm_regName[cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(15ULL)))];
std::string rm = arm_regName[cast<uint32_t>((w & cast<uint32_t>(15ULL)))];
{
if (((op1 == cast<uint32_t>(8ULL)) && (cast<uint32_t>((op2 & cast<uint32_t>(1ULL))) == cast<uint32_t>(0ULL)))){
uint32_t imm = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(7ULL))) & cast<uint32_t>(31ULL)));
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(6ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(0ULL))) {
in.Mnem = (std::string("PKHBT",5) + arm_cn(in.Cond));
in.Text = go_fmt_Sprintf(std::string("%s %s, %s, %s, LSL #%d",22),in.Mnem,rd,rn,rm,imm);
}
else {
in.Mnem = (std::string("PKHTB",5) + arm_cn(in.Cond));
if ((imm == cast<uint32_t>(0ULL))) {
imm = cast<uint32_t>(32ULL);
}
in.Text = go_fmt_Sprintf(std::string("%s %s, %s, %s, ASR #%d",22),in.Mnem,rd,rn,rm,imm);
}
return in;
}
else if (((((op1 == cast<uint32_t>(10ULL)) || (op1 == cast<uint32_t>(11ULL)))) && (cast<uint32_t>((op2 & cast<uint32_t>(1ULL))) == cast<uint32_t>(0ULL)))){
uint32_t sat = cast<uint32_t>((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(31ULL))) + cast<uint32_t>(1ULL)));
return arm_satText(w,in,std::string("SSAT",4),sat);
}
else if (((((op1 == cast<uint32_t>(14ULL)) || (op1 == cast<uint32_t>(15ULL)))) && (cast<uint32_t>((op2 & cast<uint32_t>(1ULL))) == cast<uint32_t>(0ULL)))){
uint32_t sat = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(31ULL)));
return arm_satText(w,in,std::string("USAT",4),sat);
}
else if (((op1 == cast<uint32_t>(10ULL)) && (op2 == cast<uint32_t>(1ULL)))){
in.Mnem = (std::string("SSAT16",6) + arm_cn(in.Cond));
in.Text = go_fmt_Sprintf(std::string("%s %s, #%d, %s",14),in.Mnem,rd,cast<uint32_t>((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(15ULL))) + cast<uint32_t>(1ULL))),rm);
return in;
}
else if (((op1 == cast<uint32_t>(14ULL)) && (op2 == cast<uint32_t>(1ULL)))){
in.Mnem = (std::string("USAT16",6) + arm_cn(in.Cond));
in.Text = go_fmt_Sprintf(std::string("%s %s, #%d, %s",14),in.Mnem,rd,cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(15ULL))),rm);
return in;
}
else if ((op2 == cast<uint32_t>(3ULL))){
{
std::string name = arm_extendName(op1);
if ((name != std::string("",0))) {
uint32_t rot = cast<uint32_t>((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(10ULL))) & cast<uint32_t>(3ULL))) * cast<uint32_t>(8ULL)));
bool acc = (cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(15ULL))) != cast<uint32_t>(15ULL));
in.Mnem = (name + arm_cn(in.Cond));
if (acc) {
in.Text = go_fmt_Sprintf(std::string("%s %s, %s, %s",13),in.Mnem,rd,rn,rm);
}
else {
in.Text = go_fmt_Sprintf(std::string("%s %s, %s",9),in.Mnem,rd,rm);
}
if ((rot != cast<uint32_t>(0ULL))) {
in.Text += go_fmt_Sprintf(std::string(", ROR #%d",9),rot);
}
return in;
}
}
}
else if (((op1 == cast<uint32_t>(11ULL)) && (op2 == cast<uint32_t>(1ULL)))){
in.Mnem = (std::string("REV",3) + arm_cn(in.Cond));
in.Text = go_fmt_Sprintf(std::string("%s %s, %s",9),in.Mnem,rd,rm);
return in;
}
else if (((op1 == cast<uint32_t>(11ULL)) && (op2 == cast<uint32_t>(5ULL)))){
in.Mnem = (std::string("REV16",5) + arm_cn(in.Cond));
in.Text = go_fmt_Sprintf(std::string("%s %s, %s",9),in.Mnem,rd,rm);
return in;
}
else if (((op1 == cast<uint32_t>(15ULL)) && (op2 == cast<uint32_t>(5ULL)))){
in.Mnem = (std::string("REVSH",5) + arm_cn(in.Cond));
in.Text = go_fmt_Sprintf(std::string("%s %s, %s",9),in.Mnem,rd,rm);
return in;
}
else if (((op1 == cast<uint32_t>(8ULL)) && (op2 == cast<uint32_t>(5ULL)))){
in.Mnem = (std::string("SEL",3) + arm_cn(in.Cond));
in.Text = go_fmt_Sprintf(std::string("%s %s, %s, %s",13),in.Mnem,rd,rn,rm);
return in;
}
}
tmp43:;
return arm_undef(w,in);
}
}
// tools/cpu/arm/decode_v6.go:199:1
arm_Inst arm_satText(uint32_t w,arm_Inst in,std::string name,uint32_t sat){
{
std::string rd = arm_regName[cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(12ULL))) & cast<uint32_t>(15ULL)))];
std::string rm = arm_regName[cast<uint32_t>((w & cast<uint32_t>(15ULL)))];
uint32_t amt = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(7ULL))) & cast<uint32_t>(31ULL)));
std::string sh = std::string("LSL",3);
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(6ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
sh = std::string("ASR",3);
if ((amt == cast<uint32_t>(0ULL))) {
amt = cast<uint32_t>(32ULL);
}
}
in.Mnem = (name + arm_cn(in.Cond));
if ((amt == cast<uint32_t>(0ULL))) {
in.Text = go_fmt_Sprintf(std::string("%s %s, #%d, %s",14),in.Mnem,rd,sat,rm);
}
else {
in.Text = go_fmt_Sprintf(std::string("%s %s, #%d, %s, %s #%d",22),in.Mnem,rd,sat,rm,sh,amt);
}
return in;
}
}
// tools/cpu/arm/decode_v6.go:221:1
std::string arm_extendName(uint32_t op1){
{
{
switch(op1){
case cast<uint32_t>(8ULL):{
return std::string("SXTB16",6);
break;}
case cast<uint32_t>(10ULL):{
return std::string("SXTB",4);
break;}
case cast<uint32_t>(11ULL):{
return std::string("SXTH",4);
break;}
case cast<uint32_t>(12ULL):{
return std::string("UXTB16",6);
break;}
case cast<uint32_t>(14ULL):{
return std::string("UXTB",4);
break;}
case cast<uint32_t>(15ULL):{
return std::string("UXTH",4);
break;}
}}
return std::string("",0);
}
}
// tools/cpu/arm/decode_v6.go:241:1
arm_Inst arm_decodeMediaMul(uint32_t w,arm_Inst in){
{
uint32_t op1 = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(20ULL))) & cast<uint32_t>(31ULL)));
uint32_t op2 = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(5ULL))) & cast<uint32_t>(7ULL)));
std::string rd = arm_regName[cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(15ULL)))];
uint32_t ra = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(12ULL))) & cast<uint32_t>(15ULL)));
std::string rm = arm_regName[cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(8ULL))) & cast<uint32_t>(15ULL)))];
std::string rn = arm_regName[cast<uint32_t>((w & cast<uint32_t>(15ULL)))];
std::string x = std::string("X",1);
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(5ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(0ULL))) {
x = std::string("",0);
}
{
switch(op1){
case cast<uint32_t>(16ULL):{
std::string base={};
{
switch(shr<uint32_t>(op2,cast<int64_t>(1ULL))){
case cast<uint32_t>(0ULL):{
base = std::string("SMLAD",5);
break;}
case cast<uint32_t>(1ULL):{
base = std::string("SMLSD",5);
break;}
default:{
return arm_undef(w,in);
break;}
}}
if ((ra == cast<uint32_t>(15ULL))) {
base = get(Map<std::string,std::string>{{std::string("SMLAD",5),std::string("SMUAD",5)},{std::string("SMLSD",5),std::string("SMUSD",5)}},base);
in.Mnem = ((base + x) + arm_cn(in.Cond));
in.Text = go_fmt_Sprintf(std::string("%s %s, %s, %s",13),in.Mnem,rd,rn,rm);
}
else {
in.Mnem = ((base + x) + arm_cn(in.Cond));
in.Text = go_fmt_Sprintf(std::string("%s %s, %s, %s, %s",17),in.Mnem,rd,rn,rm,arm_regName[ra]);
}
return in;
break;}
case cast<uint32_t>(20ULL):{
std::string base = std::string("SMLALD",6);
if ((shr<uint32_t>(op2,cast<int64_t>(1ULL)) == cast<uint32_t>(1ULL))) {
base = std::string("SMLSLD",6);
}
in.Mnem = ((base + x) + arm_cn(in.Cond));
in.Text = go_fmt_Sprintf(std::string("%s %s, %s, %s, %s",17),in.Mnem,arm_regName[ra],rd,rn,rm);
return in;
break;}
case cast<uint32_t>(21ULL):{
std::string r = std::string("",0);
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(5ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
r = std::string("R",1);
}
{
switch(shr<uint32_t>(op2,cast<int64_t>(1ULL))){
case cast<uint32_t>(0ULL):{
if ((ra == cast<uint32_t>(15ULL))) {
in.Mnem = ((std::string("SMMUL",5) + r) + arm_cn(in.Cond));
in.Text = go_fmt_Sprintf(std::string("%s %s, %s, %s",13),in.Mnem,rd,rn,rm);
}
else {
in.Mnem = ((std::string("SMMLA",5) + r) + arm_cn(in.Cond));
in.Text = go_fmt_Sprintf(std::string("%s %s, %s, %s, %s",17),in.Mnem,rd,rn,rm,arm_regName[ra]);
}
return in;
break;}
case cast<uint32_t>(3ULL):{
in.Mnem = ((std::string("SMMLS",5) + r) + arm_cn(in.Cond));
in.Text = go_fmt_Sprintf(std::string("%s %s, %s, %s, %s",17),in.Mnem,rd,rn,rm,arm_regName[ra]);
return in;
break;}
}}
break;}
}}
return arm_undef(w,in);
}
}
// tools/cpu/arm/decode_v6.go:308:1
std::tuple<arm_Inst,bool> arm_decodeV6Uncond(uint32_t w,arm_Inst in){
{
{
if (((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(20ULL))) & cast<uint32_t>(255ULL))) == cast<uint32_t>(16ULL)) && (cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(0ULL)))){
uint32_t imod = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(18ULL))) & cast<uint32_t>(3ULL)));
in.Mnem = std::string("CPS",3);
{
switch(imod){
case cast<uint32_t>(2ULL):{
in.Mnem = std::string("CPSIE",5);
break;}
case cast<uint32_t>(3ULL):{
in.Mnem = std::string("CPSID",5);
break;}
}}
std::string flags = std::string("",0);
{auto&& tmp45 = Slice<Anon3>{Anon3{cast<uint32_t>(8ULL),std::string("a",1)},Anon3{cast<uint32_t>(7ULL),std::string("i",1)},Anon3{cast<uint32_t>(6ULL),std::string("f",1)}};
for(int64_t tmp46=0;tmp46<len(tmp45);++tmp46){
auto ch=tmp45[tmp46];if ((cast<uint32_t>(((shr<uint32_t>(w,ch.bit)) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
flags += ch.name;
}
}}
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(17ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
in.Text = go_fmt_Sprintf(std::string("%s %s, #0x%X",12),in.Mnem,flags,cast<uint32_t>((w & cast<uint32_t>(31ULL))));
}
else {
in.Text = go_fmt_Sprintf(std::string("%s %s",5),in.Mnem,flags);
}
return {in,true};
}
else if (((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(65535ULL))) == cast<uint32_t>(61697ULL)) && (cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(4ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(0ULL)))){
std::string e = std::string("LE",2);
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(9ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
e = std::string("BE",2);
}
in.Mnem = std::string("SETEND",6);
in.Text = (std::string("SETEND ",7) + e);
return {in,true};
}
else if (((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(20ULL))) & cast<uint32_t>(255ULL))) == cast<uint32_t>(87ULL)) && (cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(4ULL))) & cast<uint32_t>(15ULL))) == cast<uint32_t>(1ULL)))){
in.Mnem = std::string("CLREX",5);
in.Text = std::string("CLREX",5);
return {in,true};
}
}
tmp44:;
return {in,false};
}
}
// tools/cpu/arm/exec_v6.go:10:1
bool arm_CPU_execARMv6(arm_CPU* c,uint32_t w){
{
if ((arm_isVFP(w) && (((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(25ULL))) & cast<uint32_t>(7ULL))) == cast<uint32_t>(6ULL)) || (cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(24ULL))) & cast<uint32_t>(15ULL))) == cast<uint32_t>(14ULL)))))) {
return arm_CPU_execVFP(c,w);
}
if ((((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(24ULL))) & cast<uint32_t>(15ULL))) == cast<uint32_t>(1ULL)) && (cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(4ULL))) & cast<uint32_t>(15ULL))) == cast<uint32_t>(9ULL))) && (cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(23ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL)))) {
arm_CPU_execSync(c,w);
return true;
}
if ((((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(24ULL))) & cast<uint32_t>(15ULL))) == cast<uint32_t>(0ULL)) && (cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(4ULL))) & cast<uint32_t>(15ULL))) == cast<uint32_t>(9ULL))) && (cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(20ULL))) & cast<uint32_t>(15ULL))) == cast<uint32_t>(4ULL)))) {
arm_CPU_execUMAAL(c,w);
return true;
}
if (((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(25ULL))) & cast<uint32_t>(7ULL))) == cast<uint32_t>(3ULL)) && (cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(4ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL)))) {
return arm_CPU_execMedia(c,w);
}
return false;
}
}
// tools/cpu/arm/exec_v6.go:34:1
bool arm_CPU_execUncondARMv6(arm_CPU* c,uint32_t w){
{
{
if (((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(20ULL))) & cast<uint32_t>(255ULL))) == cast<uint32_t>(16ULL)) && (cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(0ULL)))){
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(17ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
arm_CPU_switchMode(c,cast<uint32_t>((w & cast<uint32_t>(31ULL))));
}
uint32_t imod = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(18ULL))) & cast<uint32_t>(3ULL)));
if (((imod == cast<uint32_t>(2ULL)) || (imod == cast<uint32_t>(3ULL)))) {
bool disable = (imod == cast<uint32_t>(3ULL));
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(7ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
c->IRQDisable = disable;
}
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(6ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
c->FIQDisable = disable;
}
}
c->R[cast<int64_t>(15ULL)] = cast<uint32_t>((c->cur + cast<uint32_t>(4ULL)));
return true;
}
else if (((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(65535ULL))) == cast<uint32_t>(61697ULL)) && (cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(4ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(0ULL)))){
c->BigEndian = (cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(9ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL));
c->R[cast<int64_t>(15ULL)] = cast<uint32_t>((c->cur + cast<uint32_t>(4ULL)));
return true;
}
else if (((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(20ULL))) & cast<uint32_t>(255ULL))) == cast<uint32_t>(87ULL)) && (cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(4ULL))) & cast<uint32_t>(15ULL))) == cast<uint32_t>(1ULL)))){
c->exclValid = false;
c->R[cast<int64_t>(15ULL)] = cast<uint32_t>((c->cur + cast<uint32_t>(4ULL)));
return true;
}
}
tmp47:;
return false;
}
}
// tools/cpu/arm/exec_v6.go:66:1
void arm_CPU_execSync(arm_CPU* c,uint32_t w){
{
bool load = (cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(20ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL));
uint32_t sz = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(21ULL))) & cast<uint32_t>(3ULL)));
uint32_t addr = arm_CPU_reg(c,cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(15ULL))));
if (load) {
uint32_t rt = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(12ULL))) & cast<uint32_t>(15ULL)));
{
switch(sz){
case cast<uint32_t>(0ULL):{
arm_CPU_setReg(c,rt,arm_CPU_read32(c,addr));
break;}
case cast<uint32_t>(1ULL):{
arm_CPU_setReg(c,rt,arm_CPU_read32(c,addr));
arm_CPU_setReg(c,cast<uint32_t>((rt + cast<uint32_t>(1ULL))),arm_CPU_read32(c,cast<uint32_t>((addr + cast<uint32_t>(4ULL)))));
break;}
case cast<uint32_t>(2ULL):{
arm_CPU_setReg(c,rt,cast<uint32_t>(arm_CPU_read8(c,addr)));
break;}
case cast<uint32_t>(3ULL):{
arm_CPU_setReg(c,rt,arm_CPU_read16(c,addr));
break;}
}}
c->exclValid = true;
c->exclAddr = addr;
return ;
}
uint32_t rd = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(12ULL))) & cast<uint32_t>(15ULL)));
uint32_t rt = arm_CPU_reg(c,cast<uint32_t>((w & cast<uint32_t>(15ULL))));
if (((!c->exclValid) || (c->exclAddr != addr))) {
arm_CPU_setReg(c,rd,cast<uint32_t>(1ULL));
return ;
}
{
switch(sz){
case cast<uint32_t>(0ULL):{
arm_CPU_write32(c,addr,rt);
break;}
case cast<uint32_t>(1ULL):{
arm_CPU_write32(c,addr,rt);
arm_CPU_write32(c,cast<uint32_t>((addr + cast<uint32_t>(4ULL))),arm_CPU_reg(c,cast<uint32_t>(((cast<uint32_t>((w & cast<uint32_t>(15ULL)))) + cast<uint32_t>(1ULL)))));
break;}
case cast<uint32_t>(2ULL):{
arm_CPU_write8(c,addr,cast<uint8_t>(rt));
break;}
case cast<uint32_t>(3ULL):{
arm_CPU_write16(c,addr,rt);
break;}
}}
arm_CPU_setReg(c,rd,cast<uint32_t>(0ULL));
c->exclValid = false;
}
}
// tools/cpu/arm/exec_v6.go:116:1
void arm_CPU_execUMAAL(arm_CPU* c,uint32_t w){
{
uint32_t rdHi = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(15ULL)));
uint32_t rdLo = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(12ULL))) & cast<uint32_t>(15ULL)));
uint32_t rm = arm_CPU_reg(c,cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(8ULL))) & cast<uint32_t>(15ULL))));
uint32_t rn = arm_CPU_reg(c,cast<uint32_t>((w & cast<uint32_t>(15ULL))));
uint64_t res = cast<uint64_t>((cast<uint64_t>((cast<uint64_t>((cast<uint64_t>(rn) * cast<uint64_t>(rm))) + cast<uint64_t>(arm_CPU_reg(c,rdHi)))) + cast<uint64_t>(arm_CPU_reg(c,rdLo))));
arm_CPU_setReg(c,rdLo,cast<uint32_t>(res));
arm_CPU_setReg(c,rdHi,cast<uint32_t>(shr<uint64_t>(res,cast<int64_t>(32ULL))));
}
}
// tools/cpu/arm/exec_v6.go:129:1
bool arm_CPU_execMedia(arm_CPU* c,uint32_t w){
{
uint32_t op1 = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(20ULL))) & cast<uint32_t>(31ULL)));
{
if (((shr<uint32_t>(op1,cast<int64_t>(3ULL)) == cast<uint32_t>(0ULL)) && (cast<uint32_t>((op1 & cast<uint32_t>(7ULL))) != cast<uint32_t>(0ULL)))){
return arm_CPU_execParallel(c,w);
}
else if ((shr<uint32_t>(op1,cast<int64_t>(3ULL)) == cast<uint32_t>(1ULL))){
return arm_CPU_execMediaPack(c,w);
}
else if ((shr<uint32_t>(op1,cast<int64_t>(3ULL)) == cast<uint32_t>(2ULL))){
return arm_CPU_execMediaMul(c,w);
}
else if ((op1 == cast<uint32_t>(24ULL))){
arm_CPU_execUSAD(c,w);
return true;
}
}
tmp48:;
arm_CPU_Halt(c,std::string("unimplemented media instruction 0x%08X at 0x%08X",48),w,c->cur);
return true;
}
}
// tools/cpu/arm/exec_v6.go:159:1
bool arm_CPU_execParallel(arm_CPU* c,uint32_t w){
{
uint32_t class_ = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(20ULL))) & cast<uint32_t>(7ULL)));
uint32_t op2 = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(5ULL))) & cast<uint32_t>(7ULL)));
uint32_t rn = arm_CPU_reg(c,cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(15ULL))));
uint32_t rm = arm_CPU_reg(c,cast<uint32_t>((w & cast<uint32_t>(15ULL))));
uint32_t rd = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(12ULL))) & cast<uint32_t>(15ULL)));
bool signed_={};
arm_parMode mode={};
{
switch(class_){
case cast<uint32_t>(1ULL):{
auto tmp49 = std::make_tuple(true,cast<arm_parMode>(0ULL));
signed_ = std::get<0>(tmp49);
mode = std::get<1>(tmp49);
break;}
case cast<uint32_t>(2ULL):{
auto tmp50 = std::make_tuple(true,cast<arm_parMode>(1ULL));
signed_ = std::get<0>(tmp50);
mode = std::get<1>(tmp50);
break;}
case cast<uint32_t>(3ULL):{
auto tmp51 = std::make_tuple(true,cast<arm_parMode>(2ULL));
signed_ = std::get<0>(tmp51);
mode = std::get<1>(tmp51);
break;}
case cast<uint32_t>(5ULL):{
auto tmp52 = std::make_tuple(false,cast<arm_parMode>(0ULL));
signed_ = std::get<0>(tmp52);
mode = std::get<1>(tmp52);
break;}
case cast<uint32_t>(6ULL):{
auto tmp53 = std::make_tuple(false,cast<arm_parMode>(1ULL));
signed_ = std::get<0>(tmp53);
mode = std::get<1>(tmp53);
break;}
case cast<uint32_t>(7ULL):{
auto tmp54 = std::make_tuple(false,cast<arm_parMode>(2ULL));
signed_ = std::get<0>(tmp54);
mode = std::get<1>(tmp54);
break;}
default:{
arm_CPU_Halt(c,std::string("unimplemented parallel arithmetic 0x%08X (class %d) at 0x%08X",61),w,class_,c->cur);
return true;
break;}
}}
auto half = [&](uint32_t v,uint64_t h)->uint32_t{
return cast<uint32_t>(((shr<uint32_t>(v,(cast<uint64_t>((cast<uint64_t>(16ULL) * h))))) & cast<uint32_t>(65535ULL)));
}
;
uint32_t res={};
uint32_t ge={};
{
switch(op2){
case cast<uint32_t>(0ULL):case cast<uint32_t>(3ULL):{
bool sub_ = (op2 == cast<uint32_t>(3ULL));
{uint64_t h = cast<uint64_t>(0ULL);for (;(h < cast<uint64_t>(2ULL));h++){
auto tmp55 = arm_parLane(half(rn,h),half(rm,h),cast<uint64_t>(16ULL),signed_,mode,sub_);
uint32_t v = std::get<0>(tmp55);
bool g = std::get<1>(tmp55);
res |= shl<uint32_t>(v,(cast<uint64_t>((cast<uint64_t>(16ULL) * h))));
if (g) {
ge |= shl<uint32_t>(cast<uint32_t>(3ULL),(cast<uint64_t>((cast<uint64_t>(2ULL) * h))));
}
}
}break;}
case cast<uint32_t>(1ULL):{
auto tmp56 = arm_parLane(half(rn,cast<uint64_t>(0ULL)),half(rm,cast<uint64_t>(1ULL)),cast<uint64_t>(16ULL),signed_,mode,true);
uint32_t lo = std::get<0>(tmp56);
bool glo = std::get<1>(tmp56);
auto tmp57 = arm_parLane(half(rn,cast<uint64_t>(1ULL)),half(rm,cast<uint64_t>(0ULL)),cast<uint64_t>(16ULL),signed_,mode,false);
uint32_t hi = std::get<0>(tmp57);
bool ghi = std::get<1>(tmp57);
res = cast<uint32_t>((lo | shl<uint32_t>(hi,cast<int64_t>(16ULL))));
if (glo) {
ge |= cast<uint32_t>(3ULL);
}
if (ghi) {
ge |= cast<uint32_t>(12ULL);
}
break;}
case cast<uint32_t>(2ULL):{
auto tmp58 = arm_parLane(half(rn,cast<uint64_t>(0ULL)),half(rm,cast<uint64_t>(1ULL)),cast<uint64_t>(16ULL),signed_,mode,false);
uint32_t lo = std::get<0>(tmp58);
bool glo = std::get<1>(tmp58);
auto tmp59 = arm_parLane(half(rn,cast<uint64_t>(1ULL)),half(rm,cast<uint64_t>(0ULL)),cast<uint64_t>(16ULL),signed_,mode,true);
uint32_t hi = std::get<0>(tmp59);
bool ghi = std::get<1>(tmp59);
res = cast<uint32_t>((lo | shl<uint32_t>(hi,cast<int64_t>(16ULL))));
if (glo) {
ge |= cast<uint32_t>(3ULL);
}
if (ghi) {
ge |= cast<uint32_t>(12ULL);
}
break;}
case cast<uint32_t>(4ULL):case cast<uint32_t>(7ULL):{
bool sub_ = (op2 == cast<uint32_t>(7ULL));
{uint64_t b = cast<uint64_t>(0ULL);for (;(b < cast<uint64_t>(4ULL));b++){
uint32_t av = cast<uint32_t>(((shr<uint32_t>(rn,(cast<uint64_t>((cast<uint64_t>(8ULL) * b))))) & cast<uint32_t>(255ULL)));
uint32_t bv = cast<uint32_t>(((shr<uint32_t>(rm,(cast<uint64_t>((cast<uint64_t>(8ULL) * b))))) & cast<uint32_t>(255ULL)));
auto tmp60 = arm_parLane(av,bv,cast<uint64_t>(8ULL),signed_,mode,sub_);
uint32_t v = std::get<0>(tmp60);
bool g = std::get<1>(tmp60);
res |= shl<uint32_t>(v,(cast<uint64_t>((cast<uint64_t>(8ULL) * b))));
if (g) {
ge |= shl<uint32_t>(cast<uint32_t>(1ULL),b);
}
}
}break;}
default:{
arm_CPU_Halt(c,std::string("unimplemented parallel op2=%d at 0x%08X",39),op2,c->cur);
return true;
break;}
}}
arm_CPU_setReg(c,rd,res);
if ((mode == cast<arm_parMode>(0ULL))) {
c->GE = ge;
}
return true;
}
}
// tools/cpu/arm/exec_v6.go:260:1
std::tuple<uint32_t,bool> arm_parLane(uint32_t av,uint32_t bv,uint64_t width,bool signed_,arm_parMode mode,bool sub_){
{
uint32_t mask = cast<uint32_t>((shl<uint32_t>(cast<uint32_t>(1ULL),width) - cast<uint32_t>(1ULL)));
int32_t val={};
if (signed_) {
auto tmp61 = std::make_tuple(arm_signExtendLane(av,width),arm_signExtendLane(bv,width));
int32_t a = std::get<0>(tmp61);
int32_t b = std::get<1>(tmp61);
if (sub_) {
val = cast<int32_t>((a - b));
}
else {
val = cast<int32_t>((a + b));
}
}
else if (sub_) {
val = cast<int32_t>((cast<int32_t>(av) - cast<int32_t>(bv)));
}
else {
val = cast<int32_t>((cast<int32_t>(av) + cast<int32_t>(bv)));
}
bool ge={};
{
if (signed_){
ge = (val >= cast<int32_t>(0ULL));
}
else if (sub_){
ge = (av >= bv);
}
else {
ge = (val > cast<int32_t>(mask));
}
}
tmp62:;
{
switch(mode){
case cast<arm_parMode>(1ULL):{
if (signed_) {
int32_t hi = cast<int32_t>((shl<int32_t>(cast<int32_t>(1ULL),(cast<uint64_t>((width - cast<uint64_t>(1ULL))))) - cast<int32_t>(1ULL)));
int32_t lo = cast<int32_t>(-(shl<int32_t>(cast<int32_t>(1ULL),(cast<uint64_t>((width - cast<uint64_t>(1ULL)))))));
if ((val > hi)) {
val = hi;
}
else if ((val < lo)) {
val = lo;
}
}
else {
if ((val < cast<int32_t>(0ULL))) {
val = cast<int32_t>(0ULL);
}
else if ((val > cast<int32_t>(mask))) {
val = cast<int32_t>(mask);
}
}
break;}
case cast<arm_parMode>(2ULL):{
val = shr<int32_t>(val,cast<int64_t>(1ULL));
break;}
}}
return {cast<uint32_t>((cast<uint32_t>(val) & mask)),ge};
}
}
// tools/cpu/arm/exec_v6.go:312:1
int32_t arm_int16v(uint32_t v){
{
return cast<int32_t>(cast<int16_t>(cast<uint16_t>(v)));
}
}
// tools/cpu/arm/exec_v6.go:315:1
int32_t arm_signExtendLane(uint32_t v,uint64_t width){
{
uint64_t sh = cast<uint64_t>((cast<uint64_t>(32ULL) - width));
return shr<int32_t>(cast<int32_t>(shl<uint32_t>(v,sh)),sh);
}
}
// tools/cpu/arm/exec_v6.go:320:1
bool arm_CPU_execMediaPack(arm_CPU* c,uint32_t w){
{
uint32_t op1 = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(20ULL))) & cast<uint32_t>(31ULL)));
uint32_t op2 = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(5ULL))) & cast<uint32_t>(7ULL)));
uint32_t rd = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(12ULL))) & cast<uint32_t>(15ULL)));
uint32_t rnv = arm_CPU_reg(c,cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(15ULL))));
uint32_t rmv = arm_CPU_reg(c,cast<uint32_t>((w & cast<uint32_t>(15ULL))));
{
if (((op1 == cast<uint32_t>(8ULL)) && (cast<uint32_t>((op2 & cast<uint32_t>(1ULL))) == cast<uint32_t>(0ULL)))){
uint32_t imm = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(7ULL))) & cast<uint32_t>(31ULL)));
uint32_t res={};
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(6ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(0ULL))) {
res = cast<uint32_t>(((cast<uint32_t>((rnv & cast<uint32_t>(65535ULL)))) | (cast<uint32_t>(((shl<uint32_t>(rmv,imm)) & cast<uint32_t>(4294901760ULL))))));
}
else {
if ((imm == cast<uint32_t>(0ULL))) {
imm = cast<uint32_t>(32ULL);
}
res = cast<uint32_t>(((cast<uint32_t>((rnv & cast<uint32_t>(4294901760ULL)))) | (cast<uint32_t>((cast<uint32_t>(shr<int32_t>(cast<int32_t>(rmv),imm)) & cast<uint32_t>(65535ULL))))));
}
arm_CPU_setReg(c,rd,res);
return true;
}
else if (((((op1 == cast<uint32_t>(10ULL)) || (op1 == cast<uint32_t>(11ULL)))) && (cast<uint32_t>((op2 & cast<uint32_t>(1ULL))) == cast<uint32_t>(0ULL)))){
arm_CPU_setReg(c,rd,arm_CPU_doSat(c,w,true));
return true;
}
else if (((((op1 == cast<uint32_t>(14ULL)) || (op1 == cast<uint32_t>(15ULL)))) && (cast<uint32_t>((op2 & cast<uint32_t>(1ULL))) == cast<uint32_t>(0ULL)))){
arm_CPU_setReg(c,rd,arm_CPU_doSat(c,w,false));
return true;
}
else if ((op2 == cast<uint32_t>(3ULL))){
if (arm_CPU_execExtend(c,w)) {
return true;
}
}
else if (((op1 == cast<uint32_t>(11ULL)) && (op2 == cast<uint32_t>(1ULL)))){
arm_CPU_setReg(c,rd,go_bits_ReverseBytes32(rmv));
return true;
}
else if (((op1 == cast<uint32_t>(11ULL)) && (op2 == cast<uint32_t>(5ULL)))){
arm_CPU_setReg(c,rd,cast<uint32_t>((shl<uint32_t>((cast<uint32_t>((rmv & cast<uint32_t>(16711935ULL)))),cast<int64_t>(8ULL)) | shr<uint32_t>((cast<uint32_t>((rmv & cast<uint32_t>(4278255360ULL)))),cast<int64_t>(8ULL)))));
return true;
}
else if (((op1 == cast<uint32_t>(15ULL)) && (op2 == cast<uint32_t>(5ULL)))){
uint32_t v = cast<uint32_t>((shl<uint32_t>((cast<uint32_t>((rmv & cast<uint32_t>(255ULL)))),cast<int64_t>(8ULL)) | shr<uint32_t>((cast<uint32_t>((rmv & cast<uint32_t>(65280ULL)))),cast<int64_t>(8ULL))));
arm_CPU_setReg(c,rd,cast<uint32_t>(cast<int32_t>(cast<int16_t>(cast<uint16_t>(v)))));
return true;
}
else if (((op1 == cast<uint32_t>(8ULL)) && (op2 == cast<uint32_t>(5ULL)))){
uint32_t res={};
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(4ULL));i++){
uint64_t sh = cast<uint64_t>(cast<int64_t>((i * cast<int64_t>(8ULL))));
if ((cast<uint32_t>((c->GE & (shl<uint32_t>(cast<uint32_t>(1ULL),cast<uint64_t>(i))))) != cast<uint32_t>(0ULL))) {
res |= cast<uint32_t>((rnv & (shl<uint32_t>(cast<uint32_t>(255ULL),sh))));
}
else {
res |= cast<uint32_t>((rmv & (shl<uint32_t>(cast<uint32_t>(255ULL),sh))));
}
}
}arm_CPU_setReg(c,rd,res);
return true;
}
}
tmp63:;
arm_CPU_Halt(c,std::string("unimplemented media pack 0x%08X at 0x%08X",41),w,c->cur);
return true;
}
}
// tools/cpu/arm/exec_v6.go:381:1
uint32_t arm_CPU_doSat(arm_CPU* c,uint32_t w,bool signed_){
{
uint32_t rmv = arm_CPU_reg(c,cast<uint32_t>((w & cast<uint32_t>(15ULL))));
uint32_t amt = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(7ULL))) & cast<uint32_t>(31ULL)));
int32_t val = cast<int32_t>(rmv);
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(6ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
if ((amt == cast<uint32_t>(0ULL))) {
amt = cast<uint32_t>(32ULL);
}
val = shr<int32_t>(val,amt);
}
else if ((amt != cast<uint32_t>(0ULL))) {
val = cast<int32_t>(shl<uint32_t>(rmv,amt));
}
if (signed_) {
uint32_t n = cast<uint32_t>((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(31ULL))) + cast<uint32_t>(1ULL)));
int32_t hi = cast<int32_t>((shl<int32_t>(cast<int32_t>(1ULL),(cast<uint32_t>((n - cast<uint32_t>(1ULL))))) - cast<int32_t>(1ULL)));
int32_t lo = shl<int32_t>(cast<int32_t>(-1ULL),(cast<uint32_t>((n - cast<uint32_t>(1ULL)))));
if ((val > hi)) {
c->Q = true;
return cast<uint32_t>(hi);
}
if ((val < lo)) {
c->Q = true;
return cast<uint32_t>(lo);
}
return cast<uint32_t>(val);
}
uint32_t n = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(31ULL)));
int32_t hi = cast<int32_t>(cast<uint32_t>(((shl<uint32_t>(cast<uint32_t>(1ULL),n)) - cast<uint32_t>(1ULL))));
if ((val > hi)) {
c->Q = true;
return cast<uint32_t>(hi);
}
if ((val < cast<int32_t>(0ULL))) {
c->Q = true;
return cast<uint32_t>(0ULL);
}
return cast<uint32_t>(val);
}
}
// tools/cpu/arm/exec_v6.go:422:1
bool arm_CPU_execExtend(arm_CPU* c,uint32_t w){
{
uint32_t op1 = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(20ULL))) & cast<uint32_t>(31ULL)));
uint32_t rd = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(12ULL))) & cast<uint32_t>(15ULL)));
uint32_t rn = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(15ULL)));
uint32_t rmv = arm_CPU_reg(c,cast<uint32_t>((w & cast<uint32_t>(15ULL))));
uint32_t rot = cast<uint32_t>((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(10ULL))) & cast<uint32_t>(3ULL))) * cast<uint32_t>(8ULL)));
uint32_t rotated = arm_ror32(rmv,rot);
bool acc = (rn != cast<uint32_t>(15ULL));
uint32_t rnv = arm_CPU_reg(c,rn);
uint32_t res={};
{
switch(op1){
case cast<uint32_t>(10ULL):{
res = cast<uint32_t>(cast<int32_t>(cast<int8_t>(cast<uint8_t>(rotated))));
if (acc) {
res += rnv;
}
break;}
case cast<uint32_t>(11ULL):{
res = cast<uint32_t>(cast<int32_t>(cast<int16_t>(cast<uint16_t>(rotated))));
if (acc) {
res += rnv;
}
break;}
case cast<uint32_t>(14ULL):{
res = cast<uint32_t>((rotated & cast<uint32_t>(255ULL)));
if (acc) {
res += rnv;
}
break;}
case cast<uint32_t>(15ULL):{
res = cast<uint32_t>((rotated & cast<uint32_t>(65535ULL)));
if (acc) {
res += rnv;
}
break;}
case cast<uint32_t>(8ULL):{
uint32_t lo = cast<uint32_t>((cast<uint32_t>(cast<int32_t>(cast<int8_t>(cast<uint8_t>(rotated)))) & cast<uint32_t>(65535ULL)));
uint32_t hi = cast<uint32_t>((cast<uint32_t>(cast<int32_t>(cast<int8_t>(cast<uint8_t>(shr<uint32_t>(rotated,cast<int64_t>(16ULL)))))) & cast<uint32_t>(65535ULL)));
if (acc) {
lo = cast<uint32_t>(((cast<uint32_t>((lo + rnv))) & cast<uint32_t>(65535ULL)));
hi = cast<uint32_t>(((cast<uint32_t>((hi + (shr<uint32_t>(rnv,cast<int64_t>(16ULL)))))) & cast<uint32_t>(65535ULL)));
}
res = cast<uint32_t>((lo | shl<uint32_t>(hi,cast<int64_t>(16ULL))));
break;}
case cast<uint32_t>(12ULL):{
uint32_t lo = cast<uint32_t>((rotated & cast<uint32_t>(255ULL)));
uint32_t hi = cast<uint32_t>(((shr<uint32_t>(rotated,cast<int64_t>(16ULL))) & cast<uint32_t>(255ULL)));
if (acc) {
lo = cast<uint32_t>(((cast<uint32_t>((lo + rnv))) & cast<uint32_t>(65535ULL)));
hi = cast<uint32_t>(((cast<uint32_t>((hi + (shr<uint32_t>(rnv,cast<int64_t>(16ULL)))))) & cast<uint32_t>(65535ULL)));
}
res = cast<uint32_t>((lo | shl<uint32_t>(hi,cast<int64_t>(16ULL))));
break;}
default:{
return false;
break;}
}}
arm_CPU_setReg(c,rd,res);
return true;
}
}
// tools/cpu/arm/exec_v6.go:479:1
bool arm_CPU_execMediaMul(arm_CPU* c,uint32_t w){
{
uint32_t op1 = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(20ULL))) & cast<uint32_t>(31ULL)));
uint32_t op2 = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(5ULL))) & cast<uint32_t>(7ULL)));
uint32_t rd = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(15ULL)));
uint32_t ra = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(12ULL))) & cast<uint32_t>(15ULL)));
uint32_t rmv = arm_CPU_reg(c,cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(8ULL))) & cast<uint32_t>(15ULL))));
uint32_t rnv = arm_CPU_reg(c,cast<uint32_t>((w & cast<uint32_t>(15ULL))));
bool swap = (cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(5ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL));
uint32_t m = rmv;
if (swap) {
m = cast<uint32_t>(((shr<uint32_t>(rmv,cast<int64_t>(16ULL))) | (shl<uint32_t>(rmv,cast<int64_t>(16ULL)))));
}
{
switch(op1){
case cast<uint32_t>(16ULL):{
int32_t p0 = cast<int32_t>((arm_int16v(rnv) * arm_int16v(m)));
int32_t p1 = cast<int32_t>((arm_int16v(shr<uint32_t>(rnv,cast<int64_t>(16ULL))) * arm_int16v(shr<uint32_t>(m,cast<int64_t>(16ULL)))));
int32_t dual={};
if ((shr<uint32_t>(op2,cast<int64_t>(1ULL)) == cast<uint32_t>(0ULL))) {
dual = cast<int32_t>((p0 + p1));
}
else {
dual = cast<int32_t>((p0 - p1));
}
if ((ra != cast<uint32_t>(15ULL))) {
dual += cast<int32_t>(arm_CPU_reg(c,ra));
}
arm_CPU_setReg(c,rd,cast<uint32_t>(dual));
return true;
break;}
case cast<uint32_t>(21ULL):{
int64_t prod = cast<int64_t>((cast<int64_t>(cast<int32_t>(rnv)) * cast<int64_t>(cast<int32_t>(m))));
bool round = (cast<uint32_t>((op2 & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL));
if (round) {
prod += cast<int64_t>(2147483648ULL);
}
int32_t res={};
{
switch(shr<uint32_t>(op2,cast<int64_t>(1ULL))){
case cast<uint32_t>(0ULL):{
res = cast<int32_t>(shr<int64_t>(prod,cast<int64_t>(32ULL)));
if ((ra != cast<uint32_t>(15ULL))) {
res += cast<int32_t>(arm_CPU_reg(c,ra));
}
break;}
case cast<uint32_t>(3ULL):{
res = cast<int32_t>((cast<int32_t>(arm_CPU_reg(c,ra)) - cast<int32_t>(shr<int64_t>(prod,cast<int64_t>(32ULL)))));
break;}
default:{
return false;
break;}
}}
arm_CPU_setReg(c,rd,cast<uint32_t>(res));
return true;
break;}
}}
arm_CPU_Halt(c,std::string("unimplemented media multiply 0x%08X at 0x%08X",45),w,c->cur);
return true;
}
}
// tools/cpu/arm/exec_v6.go:534:1
void arm_CPU_execUSAD(arm_CPU* c,uint32_t w){
{
uint32_t rd = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(15ULL)));
uint32_t ra = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(12ULL))) & cast<uint32_t>(15ULL)));
uint32_t rnv = arm_CPU_reg(c,cast<uint32_t>((w & cast<uint32_t>(15ULL))));
uint32_t rmv = arm_CPU_reg(c,cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(8ULL))) & cast<uint32_t>(15ULL))));
uint32_t sum={};
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(4ULL));i++){
uint64_t sh = cast<uint64_t>(cast<int64_t>((i * cast<int64_t>(8ULL))));
int32_t d = cast<int32_t>((cast<int32_t>(cast<uint32_t>(((shr<uint32_t>(rnv,sh)) & cast<uint32_t>(255ULL)))) - cast<int32_t>(cast<uint32_t>(((shr<uint32_t>(rmv,sh)) & cast<uint32_t>(255ULL))))));
if ((d < cast<int32_t>(0ULL))) {
d = cast<int32_t>(-d);
}
sum += cast<uint32_t>(d);
}
}if ((ra != cast<uint32_t>(15ULL))) {
sum += arm_CPU_reg(c,ra);
}
arm_CPU_setReg(c,rd,sum);
}
}
// tools/cpu/arm/miscexec.go:7:1
void arm_CPU_execMisc(arm_CPU* c,uint32_t w,bool immForm){
{
uint32_t op = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(21ULL))) & cast<uint32_t>(15ULL)));
if (immForm) {
arm_CPU_execMSR(c,w,true);
return ;
}
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(7ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
if ((!arm_Variant_v5OrLater(c->Arch))) {
arm_CPU_undefV4T(c,w,std::string("SMLAxy",6));
return ;
}
arm_CPU_execSignedMul(c,w);
return ;
}
{
switch(cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(4ULL))) & cast<uint32_t>(15ULL)))){
case cast<uint32_t>(0ULL):{
if (((op == cast<uint32_t>(8ULL)) || (op == cast<uint32_t>(10ULL)))) {
uint32_t v = arm_CPU_CPSR(c);
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(22ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
v = arm_CPU_SPSR(c);
}
arm_CPU_setReg(c,cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(12ULL))) & cast<uint32_t>(15ULL))),v);
}
else {
arm_CPU_execMSR(c,w,false);
}
break;}
case cast<uint32_t>(1ULL):{
{
switch(op){
case cast<uint32_t>(9ULL):{
arm_CPU_bxTo(c,arm_CPU_reg(c,cast<uint32_t>((w & cast<uint32_t>(15ULL)))));
break;}
case cast<uint32_t>(11ULL):{
if ((!arm_Variant_v5OrLater(c->Arch))) {
arm_CPU_undefV4T(c,w,std::string("CLZ",3));
return ;
}
arm_CPU_setReg(c,cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(12ULL))) & cast<uint32_t>(15ULL))),cast<uint32_t>(go_bits_LeadingZeros32(arm_CPU_reg(c,cast<uint32_t>((w & cast<uint32_t>(15ULL)))))));
break;}
default:{
arm_CPU_Halt(c,std::string("unimplemented misc 0x%08X at 0x%08X",35),w,c->cur);
break;}
}}
break;}
case cast<uint32_t>(3ULL):{
if ((op == cast<uint32_t>(9ULL))) {
if ((!arm_Variant_v5OrLater(c->Arch))) {
arm_CPU_undefV4T(c,w,std::string("BLX (register)",14));
return ;
}
c->R[cast<int64_t>(14ULL)] = cast<uint32_t>((c->cur + cast<uint32_t>(4ULL)));
arm_CPU_bxTo(c,arm_CPU_reg(c,cast<uint32_t>((w & cast<uint32_t>(15ULL)))));
}
else {
arm_CPU_Halt(c,std::string("unimplemented misc 0x%08X at 0x%08X",35),w,c->cur);
}
break;}
case cast<uint32_t>(5ULL):{
if ((!arm_Variant_v5OrLater(c->Arch))) {
arm_CPU_undefV4T(c,w,std::string("QADD/QSUB",9));
return ;
}
arm_CPU_execSaturating(c,w);
break;}
case cast<uint32_t>(7ULL):{
if ((!arm_Variant_v5OrLater(c->Arch))) {
arm_CPU_undefV4T(c,w,std::string("BKPT",4));
return ;
}
arm_CPU_Halt(c,std::string("BKPT #0x%X at 0x%08X",20),cast<uint32_t>((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(4ULL))) & cast<uint32_t>(65520ULL))) | (cast<uint32_t>((w & cast<uint32_t>(15ULL)))))),c->cur);
break;}
default:{
arm_CPU_Halt(c,std::string("unimplemented misc 0x%08X at 0x%08X",35),w,c->cur);
break;}
}}
}
}
// tools/cpu/arm/miscexec.go:78:1
void arm_CPU_undefV4T(arm_CPU* c,uint32_t w,std::string name){
{
arm_CPU_Halt(c,std::string("%s is undefined on ARMv4T: 0x%08X at 0x%08X",43),name,w,c->cur);
}
}
// tools/cpu/arm/miscexec.go:82:1
void arm_CPU_execMSR(arm_CPU* c,uint32_t w,bool immForm){
{
uint32_t val={};
if (immForm) {
val = arm_ror32(cast<uint32_t>((w & cast<uint32_t>(255ULL))),cast<uint32_t>(((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(8ULL))) & cast<uint32_t>(15ULL)))) * cast<uint32_t>(2ULL))));
}
else {
val = arm_CPU_reg(c,cast<uint32_t>((w & cast<uint32_t>(15ULL))));
}
uint32_t fields = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(15ULL)));
uint32_t mask={};
if ((cast<uint32_t>((fields & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL))) {
mask |= cast<uint32_t>(255ULL);
}
if ((cast<uint32_t>((fields & cast<uint32_t>(2ULL))) != cast<uint32_t>(0ULL))) {
mask |= cast<uint32_t>(65280ULL);
}
if ((cast<uint32_t>((fields & cast<uint32_t>(4ULL))) != cast<uint32_t>(0ULL))) {
mask |= cast<uint32_t>(16711680ULL);
}
if ((cast<uint32_t>((fields & cast<uint32_t>(8ULL))) != cast<uint32_t>(0ULL))) {
mask |= cast<uint32_t>(4278190080ULL);
}
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(22ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
arm_CPU_SetSPSR(c,cast<uint32_t>(((arm_CPU_SPSR(c) & ~(mask)) | cast<uint32_t>((val & mask)))));
return ;
}
if ((c->Mode == cast<uint32_t>(16ULL))) {
mask &= cast<uint32_t>(4278190080ULL);
}
arm_CPU_SetCPSR(c,cast<uint32_t>(((arm_CPU_CPSR(c) & ~(mask)) | cast<uint32_t>((val & mask)))));
}
}
// tools/cpu/arm/miscexec.go:117:1
uint32_t arm_CPU_qsat(arm_CPU* c,int64_t v){
{
if ((v > cast<int64_t>(2147483647ULL))) {
c->Q = true;
return cast<uint32_t>(2147483647ULL);
}
if ((v < cast<int64_t>(-2147483648ULL))) {
c->Q = true;
return cast<uint32_t>(2147483648ULL);
}
return cast<uint32_t>(cast<int32_t>(v));
}
}
// tools/cpu/arm/miscexec.go:129:1
void arm_CPU_execSaturating(arm_CPU* c,uint32_t w){
{
int64_t rn = cast<int64_t>(cast<int32_t>(arm_CPU_reg(c,cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(15ULL))))));
int64_t rm = cast<int64_t>(cast<int32_t>(arm_CPU_reg(c,cast<uint32_t>((w & cast<uint32_t>(15ULL))))));
uint32_t rd = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(12ULL))) & cast<uint32_t>(15ULL)));
{
switch(cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(21ULL))) & cast<uint32_t>(3ULL)))){
case cast<uint32_t>(0ULL):{
arm_CPU_setReg(c,rd,arm_CPU_qsat(c,cast<int64_t>((rm + rn))));
break;}
case cast<uint32_t>(1ULL):{
arm_CPU_setReg(c,rd,arm_CPU_qsat(c,cast<int64_t>((rm - rn))));
break;}
case cast<uint32_t>(2ULL):{
uint32_t dbl = arm_CPU_qsat(c,cast<int64_t>((cast<int64_t>(2ULL) * rn)));
arm_CPU_setReg(c,rd,arm_CPU_qsat(c,cast<int64_t>((rm + cast<int64_t>(cast<int32_t>(dbl))))));
break;}
default:{
uint32_t dbl = arm_CPU_qsat(c,cast<int64_t>((cast<int64_t>(2ULL) * rn)));
arm_CPU_setReg(c,rd,arm_CPU_qsat(c,cast<int64_t>((rm - cast<int64_t>(cast<int32_t>(dbl))))));
break;}
}}
}
}
// tools/cpu/arm/miscexec.go:148:1
int64_t arm_half16(uint32_t v,uint32_t t){
{
if ((t == cast<uint32_t>(1ULL))) {
return cast<int64_t>(cast<int16_t>(shr<uint32_t>(v,cast<int64_t>(16ULL))));
}
return cast<int64_t>(cast<int16_t>(v));
}
}
// tools/cpu/arm/miscexec.go:155:1
void arm_CPU_execSignedMul(arm_CPU* c,uint32_t w){
{
uint32_t op = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(21ULL))) & cast<uint32_t>(15ULL)));
uint32_t x = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(5ULL))) & cast<uint32_t>(1ULL)));
uint32_t y = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(6ULL))) & cast<uint32_t>(1ULL)));
uint32_t rd = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(15ULL)));
uint32_t rn = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(12ULL))) & cast<uint32_t>(15ULL)));
uint32_t rs = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(8ULL))) & cast<uint32_t>(15ULL)));
uint32_t rm = cast<uint32_t>((w & cast<uint32_t>(15ULL)));
{
switch(op){
case cast<uint32_t>(8ULL):{
int64_t prod = cast<int64_t>((arm_half16(arm_CPU_reg(c,rm),x) * arm_half16(arm_CPU_reg(c,rs),y)));
int64_t acc = cast<int64_t>((prod + cast<int64_t>(cast<int32_t>(arm_CPU_reg(c,rn)))));
arm_CPU_setReg(c,rd,cast<uint32_t>(acc));
if ((acc != cast<int64_t>(cast<int32_t>(acc)))) {
c->Q = true;
}
break;}
case cast<uint32_t>(9ULL):{
int64_t wide = cast<int64_t>((cast<int64_t>(cast<int32_t>(arm_CPU_reg(c,rm))) * arm_half16(arm_CPU_reg(c,rs),y)));
int32_t res = cast<int32_t>(shr<int64_t>(wide,cast<int64_t>(16ULL)));
if ((x == cast<uint32_t>(0ULL))) {
int64_t acc = cast<int64_t>((cast<int64_t>(res) + cast<int64_t>(cast<int32_t>(arm_CPU_reg(c,rn)))));
arm_CPU_setReg(c,rd,cast<uint32_t>(acc));
if ((acc != cast<int64_t>(cast<int32_t>(acc)))) {
c->Q = true;
}
}
else {
arm_CPU_setReg(c,rd,cast<uint32_t>(res));
}
break;}
case cast<uint32_t>(10ULL):{
int64_t prod = cast<int64_t>((arm_half16(arm_CPU_reg(c,rm),x) * arm_half16(arm_CPU_reg(c,rs),y)));
int64_t acc = cast<int64_t>((cast<int64_t>(cast<uint64_t>((shl<uint64_t>(cast<uint64_t>(arm_CPU_reg(c,rd)),cast<int64_t>(32ULL)) | cast<uint64_t>(arm_CPU_reg(c,rn))))) + prod));
arm_CPU_setReg(c,rn,cast<uint32_t>(acc));
arm_CPU_setReg(c,rd,cast<uint32_t>(shr<uint64_t>(cast<uint64_t>(acc),cast<int64_t>(32ULL))));
break;}
default:{
arm_CPU_setReg(c,rd,cast<uint32_t>(cast<int64_t>((arm_half16(arm_CPU_reg(c,rm),x) * arm_half16(arm_CPU_reg(c,rs),y)))));
break;}
}}
}
}
// tools/cpu/arm/miscexec.go:193:1
void arm_CPU_execSWI(arm_CPU* c,uint32_t w){
{
if ((bool(c->SWI) && c->SWI(c,cast<uint32_t>((w & cast<uint32_t>(16777215ULL)))))) {
c->R[cast<int64_t>(15ULL)] = cast<uint32_t>((c->cur + cast<uint32_t>(4ULL)));
c->branched = true;
return ;
}
arm_CPU_exception(c,cast<uint32_t>(19ULL),cast<uint32_t>(8ULL),cast<uint32_t>((c->cur + cast<uint32_t>(4ULL))));
}
}
// tools/cpu/arm/miscexec.go:204:1
void arm_CPU_exception(arm_CPU* c,uint32_t mode,uint32_t vector,uint32_t returnAddr){
{
uint32_t saved = arm_CPU_CPSR(c);
arm_CPU_switchMode(c,mode);
arm_CPU_SetSPSR(c,saved);
c->R[cast<int64_t>(14ULL)] = returnAddr;
c->IRQDisable = true;
c->Thumb = false;
c->R[cast<int64_t>(15ULL)] = vector;
c->branched = true;
}
}
// tools/cpu/arm/miscexec.go:221:1
void arm_CPU_Exception(arm_CPU* c,uint32_t mode,uint32_t vector,uint32_t returnAddr){
{
arm_CPU_exception(c,mode,vector,returnAddr);
}
}
// tools/cpu/arm/miscexec.go:228:1
bool arm_CPU_IRQ(arm_CPU* c){
{
if (c->IRQDisable) {
return false;
}
uint32_t ret = c->R[cast<int64_t>(15ULL)];
if (c->Thumb) {
ret += cast<uint32_t>(2ULL);
}
arm_CPU_exception(c,cast<uint32_t>(18ULL),cast<uint32_t>(24ULL),cast<uint32_t>((cast<uint32_t>((ret + cast<uint32_t>(4ULL))) - cast<uint32_t>((arm_CPU_boolToU(c,c->Thumb) * cast<uint32_t>(2ULL))))));
return true;
}
}
// tools/cpu/arm/miscexec.go:244:1
void arm_CPU_execCopro(arm_CPU* c,uint32_t w){
{
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(4ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(0ULL))) {
return ;
}
bool load = (cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(20ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL));
uint32_t cp = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(8ULL))) & cast<uint32_t>(15ULL)));
uint32_t op1 = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(21ULL))) & cast<uint32_t>(7ULL)));
uint32_t crn = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(15ULL)));
uint32_t crm = cast<uint32_t>((w & cast<uint32_t>(15ULL)));
uint32_t op2 = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(5ULL))) & cast<uint32_t>(7ULL)));
uint32_t rd = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(12ULL))) & cast<uint32_t>(15ULL)));
if (bool(c->Coproc)) {
uint32_t v = arm_CPU_reg(c,rd);
c->Coproc(c,load,cp,op1,crn,crm,op2,(&v));
if ((load && (rd != cast<uint32_t>(15ULL)))) {
arm_CPU_setReg(c,rd,v);
}
return ;
}
if ((load && (rd != cast<uint32_t>(15ULL)))) {
arm_CPU_setReg(c,rd,cast<uint32_t>(0ULL));
}
}
}
// tools/cpu/arm/thumb.go:12:1
arm_Inst arm_DecodeThumb(Slice<uint8_t> code,uint32_t addr){
{
if ((len(code) < cast<int64_t>(2ULL))) {
return arm_Inst{addr,len(code),std::string(".hword",6),std::string(".hword ; truncated",18),cast<arm_Flow>(6ULL),{},{},true,{},cast<int64_t>(14ULL)};
}
uint32_t h = cast<uint32_t>((cast<uint32_t>(code[cast<int64_t>(0ULL)]) | shl<uint32_t>(cast<uint32_t>(code[cast<int64_t>(1ULL)]),cast<int64_t>(8ULL))));
arm_Inst in = arm_Inst{addr,cast<int64_t>(2ULL),{},{},cast<arm_Flow>(0ULL),{},{},true,true,cast<int64_t>(14ULL)};
{
if (((shr<uint32_t>(h,cast<int64_t>(13ULL)) == cast<uint32_t>(0ULL)) && (cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(11ULL))) & cast<uint32_t>(3ULL))) != cast<uint32_t>(3ULL)))){
std::string op = arm_shiftName[cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(11ULL))) & cast<uint32_t>(3ULL)))];
in.Mnem = op;
in.Text = go_fmt_Sprintf(std::string("%s %s, %s, #%d",14),op,arm_lo(h),arm_lo(shr<uint32_t>(h,cast<int64_t>(3ULL))),cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(6ULL))) & cast<uint32_t>(31ULL))));
}
else if ((shr<uint32_t>(h,cast<int64_t>(11ULL)) == cast<uint32_t>(3ULL))){
return arm_thumbAddSub(h,in);
}
else if ((shr<uint32_t>(h,cast<int64_t>(13ULL)) == cast<uint32_t>(1ULL))){
std::string op = std::array<std::string,4>{std::string("MOV",3),std::string("CMP",3),std::string("ADD",3),std::string("SUB",3)}[cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(11ULL))) & cast<uint32_t>(3ULL)))];
in.Mnem = op;
in.Text = go_fmt_Sprintf(std::string("%s %s, #0x%X",12),op,arm_lo(shr<uint32_t>(h,cast<int64_t>(8ULL))),cast<uint32_t>((h & cast<uint32_t>(255ULL))));
}
else if ((shr<uint32_t>(h,cast<int64_t>(10ULL)) == cast<uint32_t>(16ULL))){
return arm_thumbALU(h,in);
}
else if ((shr<uint32_t>(h,cast<int64_t>(10ULL)) == cast<uint32_t>(17ULL))){
return arm_thumbHiReg(h,in);
}
else if ((shr<uint32_t>(h,cast<int64_t>(11ULL)) == cast<uint32_t>(9ULL))){
in.Mnem = std::string("LDR",3);
in.Text = go_fmt_Sprintf(std::string("LDR %s, [pc, #0x%X]",19),arm_lo(shr<uint32_t>(h,cast<int64_t>(8ULL))),cast<uint32_t>(((cast<uint32_t>((h & cast<uint32_t>(255ULL)))) * cast<uint32_t>(4ULL))));
}
else if (((shr<uint32_t>(h,cast<int64_t>(12ULL)) == cast<uint32_t>(5ULL)) && (cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(9ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(0ULL)))){
return arm_thumbLoadStoreReg(h,in);
}
else if ((shr<uint32_t>(h,cast<int64_t>(12ULL)) == cast<uint32_t>(5ULL))){
return arm_thumbLoadStoreSExt(h,in);
}
else if ((shr<uint32_t>(h,cast<int64_t>(13ULL)) == cast<uint32_t>(3ULL))){
return arm_thumbLoadStoreImm(h,in);
}
else if ((shr<uint32_t>(h,cast<int64_t>(12ULL)) == cast<uint32_t>(8ULL))){
uint32_t l = cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(11ULL))) & cast<uint32_t>(1ULL)));
std::string name = std::string("STRH",4);
if ((l == cast<uint32_t>(1ULL))) {
name = std::string("LDRH",4);
}
in.Mnem = name;
in.Text = go_fmt_Sprintf(std::string("%s %s, [%s, #0x%X]",18),name,arm_lo(h),arm_lo(shr<uint32_t>(h,cast<int64_t>(3ULL))),cast<uint32_t>(((cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(6ULL))) & cast<uint32_t>(31ULL)))) * cast<uint32_t>(2ULL))));
}
else if ((shr<uint32_t>(h,cast<int64_t>(12ULL)) == cast<uint32_t>(9ULL))){
uint32_t l = cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(11ULL))) & cast<uint32_t>(1ULL)));
std::string name = std::string("STR",3);
if ((l == cast<uint32_t>(1ULL))) {
name = std::string("LDR",3);
}
in.Mnem = name;
in.Text = go_fmt_Sprintf(std::string("%s %s, [sp, #0x%X]",18),name,arm_lo(shr<uint32_t>(h,cast<int64_t>(8ULL))),cast<uint32_t>(((cast<uint32_t>((h & cast<uint32_t>(255ULL)))) * cast<uint32_t>(4ULL))));
}
else if ((shr<uint32_t>(h,cast<int64_t>(12ULL)) == cast<uint32_t>(10ULL))){
std::string base = std::string("pc",2);
if ((cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(11ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
base = std::string("sp",2);
}
in.Mnem = std::string("ADD",3);
in.Text = go_fmt_Sprintf(std::string("ADD %s, %s, #0x%X",17),arm_lo(shr<uint32_t>(h,cast<int64_t>(8ULL))),base,cast<uint32_t>(((cast<uint32_t>((h & cast<uint32_t>(255ULL)))) * cast<uint32_t>(4ULL))));
}
else if ((shr<uint32_t>(h,cast<int64_t>(8ULL)) == cast<uint32_t>(176ULL))){
std::string sign = std::string("",0);
if ((cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(7ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
sign = std::string("-",1);
}
in.Mnem = std::string("ADD",3);
in.Text = go_fmt_Sprintf(std::string("ADD sp, #%s0x%X",15),sign,cast<uint32_t>(((cast<uint32_t>((h & cast<uint32_t>(127ULL)))) * cast<uint32_t>(4ULL))));
}
else if (((shr<uint32_t>(h,cast<int64_t>(12ULL)) == cast<uint32_t>(11ULL)) && (cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(9ULL))) & cast<uint32_t>(3ULL))) == cast<uint32_t>(2ULL)))){
return arm_thumbPushPop(h,in);
}
else if ((shr<uint32_t>(h,cast<int64_t>(12ULL)) == cast<uint32_t>(12ULL))){
uint32_t l = cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(11ULL))) & cast<uint32_t>(1ULL)));
std::string name = std::string("STMIA",5);
if ((l == cast<uint32_t>(1ULL))) {
name = std::string("LDMIA",5);
}
in.Mnem = name;
in.Text = go_fmt_Sprintf(std::string("%s %s!, {%s}",12),name,arm_lo(shr<uint32_t>(h,cast<int64_t>(8ULL))),arm_regList(cast<uint32_t>((h & cast<uint32_t>(255ULL)))));
}
else if ((shr<uint32_t>(h,cast<int64_t>(12ULL)) == cast<uint32_t>(13ULL))){
return arm_thumbCondBranch(h,addr,in);
}
else if ((shr<uint32_t>(h,cast<int64_t>(11ULL)) == cast<uint32_t>(28ULL))){
uint32_t off = shl<uint32_t>(arm_signExtend(cast<uint32_t>((h & cast<uint32_t>(2047ULL))),cast<uint64_t>(11ULL)),cast<int64_t>(1ULL));
uint32_t target = cast<uint32_t>((cast<uint32_t>((addr + cast<uint32_t>(4ULL))) + off));
in.Mnem = std::string("B",1);
in.Text = go_fmt_Sprintf(std::string("B 0x%08X",8),target);
auto tmp65 = std::make_tuple(cast<arm_Flow>(2ULL),target,true);
in.Flow = std::get<0>(tmp65);
in.Target = std::get<1>(tmp65);
in.HasTarget = std::get<2>(tmp65);
}
else if ((((shr<uint32_t>(h,cast<int64_t>(11ULL)) == cast<uint32_t>(30ULL)) || (shr<uint32_t>(h,cast<int64_t>(11ULL)) == cast<uint32_t>(31ULL))) || (shr<uint32_t>(h,cast<int64_t>(11ULL)) == cast<uint32_t>(29ULL)))){
return arm_thumbLongBranch(code,h,addr,in);
}
else {
auto tmp66 = std::make_tuple(std::string(".hword",6),cast<arm_Flow>(6ULL));
in.Mnem = std::get<0>(tmp66);
in.Flow = std::get<1>(tmp66);
in.Text = go_fmt_Sprintf(std::string(".hword 0x%04X",13),h);
}
}
tmp64:;
return in;
}
}
// tools/cpu/arm/thumb.go:101:1
std::string arm_lo(uint32_t v){
{
return arm_regName[cast<uint32_t>((v & cast<uint32_t>(7ULL)))];
}
}
// tools/cpu/arm/thumb.go:103:1
arm_Inst arm_thumbAddSub(uint32_t h,arm_Inst in){
{
uint32_t i = cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(10ULL))) & cast<uint32_t>(1ULL)));
uint32_t op = cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(9ULL))) & cast<uint32_t>(1ULL)));
uint32_t rnOff = cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(6ULL))) & cast<uint32_t>(7ULL)));
std::string name = std::string("ADD",3);
if ((op == cast<uint32_t>(1ULL))) {
name = std::string("SUB",3);
}
in.Mnem = name;
if ((i == cast<uint32_t>(1ULL))) {
in.Text = go_fmt_Sprintf(std::string("%s %s, %s, #%d",14),name,arm_lo(h),arm_lo(shr<uint32_t>(h,cast<int64_t>(3ULL))),rnOff);
}
else {
in.Text = go_fmt_Sprintf(std::string("%s %s, %s, %s",13),name,arm_lo(h),arm_lo(shr<uint32_t>(h,cast<int64_t>(3ULL))),arm_regName[rnOff]);
}
return in;
}
}
// tools/cpu/arm/thumb.go:125:1
arm_Inst arm_thumbALU(uint32_t h,arm_Inst in){
{
std::string op = arm_thumbALUOps[cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(6ULL))) & cast<uint32_t>(15ULL)))];
in.Mnem = op;
in.Text = go_fmt_Sprintf(std::string("%s %s, %s",9),op,arm_lo(h),arm_lo(shr<uint32_t>(h,cast<int64_t>(3ULL))));
return in;
}
}
// tools/cpu/arm/thumb.go:135:1
arm_Inst arm_thumbHiReg(uint32_t h,arm_Inst in){
{
uint32_t op = cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(8ULL))) & cast<uint32_t>(3ULL)));
uint32_t rd = cast<uint32_t>(((cast<uint32_t>((h & cast<uint32_t>(7ULL)))) | cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(4ULL))) & cast<uint32_t>(8ULL)))));
uint32_t rm = cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(3ULL))) & cast<uint32_t>(15ULL)));
{
switch(op){
case cast<uint32_t>(0ULL):{
in.Mnem = std::string("ADD",3);
in.Text = go_fmt_Sprintf(std::string("ADD %s, %s",10),arm_regName[rd],arm_regName[rm]);
if ((rd == cast<uint32_t>(15ULL))) {
in.Flow = cast<arm_Flow>(5ULL);
}
break;}
case cast<uint32_t>(1ULL):{
in.Mnem = std::string("CMP",3);
in.Text = go_fmt_Sprintf(std::string("CMP %s, %s",10),arm_regName[rd],arm_regName[rm]);
break;}
case cast<uint32_t>(2ULL):{
in.Mnem = std::string("MOV",3);
in.Text = go_fmt_Sprintf(std::string("MOV %s, %s",10),arm_regName[rd],arm_regName[rm]);
if ((rd == cast<uint32_t>(15ULL))) {
if ((rm == cast<uint32_t>(14ULL))) {
in.Flow = cast<arm_Flow>(4ULL);
}
else {
in.Flow = cast<arm_Flow>(5ULL);
}
}
break;}
default:{
if ((cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(7ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(0ULL))) {
in.Mnem = std::string("BX",2);
in.Text = go_fmt_Sprintf(std::string("BX %s",5),arm_regName[rm]);
if ((rm == cast<uint32_t>(14ULL))) {
in.Flow = cast<arm_Flow>(4ULL);
}
else {
in.Flow = cast<arm_Flow>(5ULL);
}
}
else {
in.Mnem = std::string("BLX",3);
in.Text = go_fmt_Sprintf(std::string("BLX %s",6),arm_regName[rm]);
in.Flow = cast<arm_Flow>(7ULL);
}
break;}
}}
return in;
}
}
// tools/cpu/arm/thumb.go:177:1
arm_Inst arm_thumbLoadStoreReg(uint32_t h,arm_Inst in){
{
uint32_t l = cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(11ULL))) & cast<uint32_t>(1ULL)));
uint32_t b = cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(10ULL))) & cast<uint32_t>(1ULL)));
std::string name = std::string("STR",3);
if ((l == cast<uint32_t>(1ULL))) {
name = std::string("LDR",3);
}
if ((b == cast<uint32_t>(1ULL))) {
name += std::string("B",1);
}
in.Mnem = name;
in.Text = go_fmt_Sprintf(std::string("%s %s, [%s, %s]",15),name,arm_lo(h),arm_lo(shr<uint32_t>(h,cast<int64_t>(3ULL))),arm_lo(shr<uint32_t>(h,cast<int64_t>(6ULL))));
return in;
}
}
// tools/cpu/arm/thumb.go:192:1
arm_Inst arm_thumbLoadStoreSExt(uint32_t h,arm_Inst in){
{
std::string name = std::array<std::string,4>{std::string("STRH",4),std::string("LDSB",4),std::string("LDRH",4),std::string("LDSH",4)}[cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(10ULL))) & cast<uint32_t>(3ULL)))];
in.Mnem = name;
in.Text = go_fmt_Sprintf(std::string("%s %s, [%s, %s]",15),name,arm_lo(h),arm_lo(shr<uint32_t>(h,cast<int64_t>(3ULL))),arm_lo(shr<uint32_t>(h,cast<int64_t>(6ULL))));
return in;
}
}
// tools/cpu/arm/thumb.go:199:1
arm_Inst arm_thumbLoadStoreImm(uint32_t h,arm_Inst in){
{
uint32_t b = cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(12ULL))) & cast<uint32_t>(1ULL)));
uint32_t l = cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(11ULL))) & cast<uint32_t>(1ULL)));
std::string name = std::string("STR",3);
if ((l == cast<uint32_t>(1ULL))) {
name = std::string("LDR",3);
}
uint32_t off = cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(6ULL))) & cast<uint32_t>(31ULL)));
if ((b == cast<uint32_t>(1ULL))) {
name += std::string("B",1);
}
else {
off *= cast<uint32_t>(4ULL);
}
in.Mnem = name;
in.Text = go_fmt_Sprintf(std::string("%s %s, [%s, #0x%X]",18),name,arm_lo(h),arm_lo(shr<uint32_t>(h,cast<int64_t>(3ULL))),off);
return in;
}
}
// tools/cpu/arm/thumb.go:217:1
arm_Inst arm_thumbPushPop(uint32_t h,arm_Inst in){
{
uint32_t l = cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(11ULL))) & cast<uint32_t>(1ULL)));
uint32_t r = cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(8ULL))) & cast<uint32_t>(1ULL)));
uint32_t mask = cast<uint32_t>((h & cast<uint32_t>(255ULL)));
if ((l == cast<uint32_t>(1ULL))) {
uint32_t extra = cast<uint32_t>(0ULL);
if ((r == cast<uint32_t>(1ULL))) {
extra = cast<uint32_t>(32768ULL);
}
in.Mnem = std::string("POP",3);
in.Text = go_fmt_Sprintf(std::string("POP {%s}",8),arm_regList(cast<uint32_t>((mask | extra))));
if ((r == cast<uint32_t>(1ULL))) {
in.Flow = cast<arm_Flow>(4ULL);
}
}
else {
uint32_t extra = cast<uint32_t>(0ULL);
if ((r == cast<uint32_t>(1ULL))) {
extra = cast<uint32_t>(16384ULL);
}
in.Mnem = std::string("PUSH",4);
in.Text = go_fmt_Sprintf(std::string("PUSH {%s}",9),arm_regList(cast<uint32_t>((mask | extra))));
}
return in;
}
}
// tools/cpu/arm/thumb.go:242:1
arm_Inst arm_thumbCondBranch(uint32_t h,uint32_t addr,arm_Inst in){
{
int64_t cond = cast<int64_t>(cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(8ULL))) & cast<uint32_t>(15ULL))));
if ((cond == cast<int64_t>(15ULL))) {
in.Mnem = std::string("SWI",3);
in.Text = go_fmt_Sprintf(std::string("SWI #0x%X",9),cast<uint32_t>((h & cast<uint32_t>(255ULL))));
in.Cond = cast<int64_t>(14ULL);
return in;
}
if ((cond == cast<int64_t>(14ULL))) {
auto tmp67 = std::make_tuple(std::string(".hword",6),cast<arm_Flow>(6ULL));
in.Mnem = std::get<0>(tmp67);
in.Flow = std::get<1>(tmp67);
in.Text = go_fmt_Sprintf(std::string(".hword 0x%04X",13),h);
return in;
}
uint32_t off = shl<uint32_t>(arm_signExtend(cast<uint32_t>((h & cast<uint32_t>(255ULL))),cast<uint64_t>(8ULL)),cast<int64_t>(1ULL));
uint32_t target = cast<uint32_t>((cast<uint32_t>((addr + cast<uint32_t>(4ULL))) + off));
in.Cond = cond;
in.Mnem = (std::string("B",1) + arm_condName[cond]);
in.Text = go_fmt_Sprintf(std::string("%s 0x%08X",9),in.Mnem,target);
auto tmp68 = std::make_tuple(cast<arm_Flow>(1ULL),target,true);
in.Flow = std::get<0>(tmp68);
in.Target = std::get<1>(tmp68);
in.HasTarget = std::get<2>(tmp68);
return in;
}
}
// tools/cpu/arm/thumb.go:268:1
arm_Inst arm_thumbLongBranch(Slice<uint8_t> code,uint32_t h,uint32_t addr,arm_Inst in){
{
uint32_t top = shr<uint32_t>(h,cast<int64_t>(11ULL));
if ((top != cast<uint32_t>(30ULL))) {
in.Mnem = std::string("BL(suffix)",10);
in.Text = go_fmt_Sprintf(std::string("; BL/BLX suffix (no prefix) 0x%04X",34),h);
return in;
}
if ((len(code) < cast<int64_t>(4ULL))) {
in.Mnem = std::string("BL(hi)",6);
in.Text = go_fmt_Sprintf(std::string("BL (hi) #0x%X",13),cast<uint32_t>((h & cast<uint32_t>(2047ULL))));
return in;
}
uint32_t h2 = cast<uint32_t>((cast<uint32_t>(code[cast<int64_t>(2ULL)]) | shl<uint32_t>(cast<uint32_t>(code[cast<int64_t>(3ULL)]),cast<int64_t>(8ULL))));
uint32_t suf = shr<uint32_t>(h2,cast<int64_t>(11ULL));
if (((suf != cast<uint32_t>(31ULL)) && (suf != cast<uint32_t>(29ULL)))) {
in.Mnem = std::string("BL(hi)",6);
in.Text = go_fmt_Sprintf(std::string("BL (hi) #0x%X",13),cast<uint32_t>((h & cast<uint32_t>(2047ULL))));
return in;
}
in.Len = cast<int64_t>(4ULL);
uint32_t hiOff = shl<uint32_t>(arm_signExtend(cast<uint32_t>((h & cast<uint32_t>(2047ULL))),cast<uint64_t>(11ULL)),cast<int64_t>(12ULL));
uint32_t loOff = shl<uint32_t>((cast<uint32_t>((h2 & cast<uint32_t>(2047ULL)))),cast<int64_t>(1ULL));
uint32_t target = cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((addr + cast<uint32_t>(4ULL))) + hiOff)) + loOff));
if ((suf == cast<uint32_t>(29ULL))) {
target &= ~(cast<uint32_t>(3ULL));
in.Mnem = std::string("BLX",3);
in.Text = go_fmt_Sprintf(std::string("BLX 0x%08X",10),target);
auto tmp69 = std::make_tuple(cast<arm_Flow>(3ULL),target,true,false);
in.Flow = std::get<0>(tmp69);
in.Target = std::get<1>(tmp69);
in.HasTarget = std::get<2>(tmp69);
in.TargetThumb = std::get<3>(tmp69);
}
else {
in.Mnem = std::string("BL",2);
in.Text = go_fmt_Sprintf(std::string("BL 0x%08X",9),target);
auto tmp70 = std::make_tuple(cast<arm_Flow>(3ULL),target,true,true);
in.Flow = std::get<0>(tmp70);
in.Target = std::get<1>(tmp70);
in.HasTarget = std::get<2>(tmp70);
in.TargetThumb = std::get<3>(tmp70);
}
return in;
}
}
// tools/cpu/arm/thumbexec.go:9:1
int64_t arm_CPU_stepThumb(arm_CPU* c){
{
uint32_t h = arm_CPU_read16(c,c->cur);
arm_CPU_execThumb(c,h);
if ((!c->branched)) {
c->R[cast<int64_t>(15ULL)] = cast<uint32_t>((c->cur + cast<uint32_t>(2ULL)));
}
return cast<int64_t>(1ULL);
}
}
// tools/cpu/arm/thumbexec.go:18:1
void arm_CPU_execThumb(arm_CPU* c,uint32_t h){
{
{
if (((shr<uint32_t>(h,cast<int64_t>(13ULL)) == cast<uint32_t>(0ULL)) && (cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(11ULL))) & cast<uint32_t>(3ULL))) != cast<uint32_t>(3ULL)))){
uint32_t typ = cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(11ULL))) & cast<uint32_t>(3ULL)));
auto tmp72 = arm_CPU_shift(c,typ,cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(6ULL))) & cast<uint32_t>(31ULL))),c->R[cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(3ULL))) & cast<uint32_t>(7ULL)))],false,arm_CPU_boolToU(c,c->C));
uint32_t res = std::get<0>(tmp72);
uint32_t carry = std::get<1>(tmp72);
c->R[cast<uint32_t>((h & cast<uint32_t>(7ULL)))] = res;
arm_CPU_setNZ(c,res);
c->C = (carry == cast<uint32_t>(1ULL));
}
else if ((shr<uint32_t>(h,cast<int64_t>(11ULL)) == cast<uint32_t>(3ULL))){
arm_CPU_thumbAddSub(c,h);
}
else if ((shr<uint32_t>(h,cast<int64_t>(13ULL)) == cast<uint32_t>(1ULL))){
arm_CPU_thumbImm(c,h);
}
else if ((shr<uint32_t>(h,cast<int64_t>(10ULL)) == cast<uint32_t>(16ULL))){
arm_CPU_thumbALUExec(c,h);
}
else if ((shr<uint32_t>(h,cast<int64_t>(10ULL)) == cast<uint32_t>(17ULL))){
arm_CPU_thumbHiRegExec(c,h);
}
else if ((shr<uint32_t>(h,cast<int64_t>(11ULL)) == cast<uint32_t>(9ULL))){
uint32_t addr = cast<uint32_t>(((((cast<uint32_t>((c->cur + cast<uint32_t>(4ULL)))) & ~(cast<uint32_t>(3ULL)))) + cast<uint32_t>(((cast<uint32_t>((h & cast<uint32_t>(255ULL)))) * cast<uint32_t>(4ULL)))));
c->R[cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(8ULL))) & cast<uint32_t>(7ULL)))] = arm_CPU_read32(c,addr);
}
else if (((shr<uint32_t>(h,cast<int64_t>(12ULL)) == cast<uint32_t>(5ULL)) && (cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(9ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(0ULL)))){
arm_CPU_thumbLoadStoreReg(c,h);
}
else if ((shr<uint32_t>(h,cast<int64_t>(12ULL)) == cast<uint32_t>(5ULL))){
arm_CPU_thumbLoadStoreSExt(c,h);
}
else if ((shr<uint32_t>(h,cast<int64_t>(13ULL)) == cast<uint32_t>(3ULL))){
arm_CPU_thumbLoadStoreImm(c,h);
}
else if ((shr<uint32_t>(h,cast<int64_t>(12ULL)) == cast<uint32_t>(8ULL))){
uint32_t addr = cast<uint32_t>((c->R[cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(3ULL))) & cast<uint32_t>(7ULL)))] + cast<uint32_t>(((cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(6ULL))) & cast<uint32_t>(31ULL)))) * cast<uint32_t>(2ULL)))));
if ((cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(11ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
c->R[cast<uint32_t>((h & cast<uint32_t>(7ULL)))] = arm_CPU_read16(c,addr);
}
else {
arm_CPU_write16(c,addr,c->R[cast<uint32_t>((h & cast<uint32_t>(7ULL)))]);
}
}
else if ((shr<uint32_t>(h,cast<int64_t>(12ULL)) == cast<uint32_t>(9ULL))){
uint32_t addr = cast<uint32_t>((c->R[cast<int64_t>(13ULL)] + cast<uint32_t>(((cast<uint32_t>((h & cast<uint32_t>(255ULL)))) * cast<uint32_t>(4ULL)))));
if ((cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(11ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
c->R[cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(8ULL))) & cast<uint32_t>(7ULL)))] = arm_CPU_read32(c,addr);
}
else {
arm_CPU_write32(c,addr,c->R[cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(8ULL))) & cast<uint32_t>(7ULL)))]);
}
}
else if ((shr<uint32_t>(h,cast<int64_t>(12ULL)) == cast<uint32_t>(10ULL))){
uint32_t off = cast<uint32_t>(((cast<uint32_t>((h & cast<uint32_t>(255ULL)))) * cast<uint32_t>(4ULL)));
if ((cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(11ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
c->R[cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(8ULL))) & cast<uint32_t>(7ULL)))] = cast<uint32_t>((c->R[cast<int64_t>(13ULL)] + off));
}
else {
c->R[cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(8ULL))) & cast<uint32_t>(7ULL)))] = cast<uint32_t>(((((cast<uint32_t>((c->cur + cast<uint32_t>(4ULL)))) & ~(cast<uint32_t>(3ULL)))) + off));
}
}
else if ((shr<uint32_t>(h,cast<int64_t>(8ULL)) == cast<uint32_t>(176ULL))){
uint32_t off = cast<uint32_t>(((cast<uint32_t>((h & cast<uint32_t>(127ULL)))) * cast<uint32_t>(4ULL)));
if ((cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(7ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
c->R[cast<int64_t>(13ULL)] -= off;
}
else {
c->R[cast<int64_t>(13ULL)] += off;
}
}
else if (((shr<uint32_t>(h,cast<int64_t>(12ULL)) == cast<uint32_t>(11ULL)) && (cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(9ULL))) & cast<uint32_t>(3ULL))) == cast<uint32_t>(2ULL)))){
arm_CPU_thumbPushPop(c,h);
}
else if ((shr<uint32_t>(h,cast<int64_t>(12ULL)) == cast<uint32_t>(12ULL))){
arm_CPU_thumbBlock(c,h);
}
else if ((shr<uint32_t>(h,cast<int64_t>(12ULL)) == cast<uint32_t>(13ULL))){
arm_CPU_thumbCondBranch(c,h);
}
else if ((shr<uint32_t>(h,cast<int64_t>(11ULL)) == cast<uint32_t>(28ULL))){
arm_CPU_branchTo(c,cast<uint32_t>((cast<uint32_t>((c->cur + cast<uint32_t>(4ULL))) + shl<uint32_t>(arm_signExtend(cast<uint32_t>((h & cast<uint32_t>(2047ULL))),cast<uint64_t>(11ULL)),cast<int64_t>(1ULL)))));
}
else if ((shr<uint32_t>(h,cast<int64_t>(11ULL)) == cast<uint32_t>(30ULL))){
c->R[cast<int64_t>(14ULL)] = cast<uint32_t>((cast<uint32_t>((c->cur + cast<uint32_t>(4ULL))) + shl<uint32_t>(arm_signExtend(cast<uint32_t>((h & cast<uint32_t>(2047ULL))),cast<uint64_t>(11ULL)),cast<int64_t>(12ULL))));
}
else if ((shr<uint32_t>(h,cast<int64_t>(11ULL)) == cast<uint32_t>(31ULL))){
uint32_t target = cast<uint32_t>((c->R[cast<int64_t>(14ULL)] + shl<uint32_t>((cast<uint32_t>((h & cast<uint32_t>(2047ULL)))),cast<int64_t>(1ULL))));
c->R[cast<int64_t>(14ULL)] = cast<uint32_t>(((cast<uint32_t>((c->cur + cast<uint32_t>(2ULL)))) | cast<uint32_t>(1ULL)));
arm_CPU_branchTo(c,target);
}
else if ((shr<uint32_t>(h,cast<int64_t>(11ULL)) == cast<uint32_t>(29ULL))){
if ((!arm_Variant_v5OrLater(c->Arch))) {
arm_CPU_Halt(c,std::string("Thumb BLX (immediate) is undefined on ARMv4T: 0x%04X at 0x%08X",62),h,c->cur);
return ;
}
uint32_t target = ((cast<uint32_t>((c->R[cast<int64_t>(14ULL)] + shl<uint32_t>((cast<uint32_t>((h & cast<uint32_t>(2047ULL)))),cast<int64_t>(1ULL))))) & ~(cast<uint32_t>(3ULL)));
c->R[cast<int64_t>(14ULL)] = cast<uint32_t>(((cast<uint32_t>((c->cur + cast<uint32_t>(2ULL)))) | cast<uint32_t>(1ULL)));
c->Thumb = false;
c->R[cast<int64_t>(15ULL)] = target;
c->branched = true;
}
else {
arm_CPU_Halt(c,std::string("undefined Thumb 0x%04X at 0x%08X",32),h,c->cur);
}
}
tmp71:;
}
}
// tools/cpu/arm/thumbexec.go:100:1
void arm_CPU_thumbAddSub(arm_CPU* c,uint32_t h){
{
auto tmp73 = std::make_tuple(cast<uint32_t>((h & cast<uint32_t>(7ULL))),cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(3ULL))) & cast<uint32_t>(7ULL))));
uint32_t rd = std::get<0>(tmp73);
uint32_t rs = std::get<1>(tmp73);
uint32_t operand={};
if ((cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(10ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
operand = cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(6ULL))) & cast<uint32_t>(7ULL)));
}
else {
operand = c->R[cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(6ULL))) & cast<uint32_t>(7ULL)))];
}
if ((cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(9ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
c->R[rd] = arm_CPU_sub_(c,c->R[rs],operand,cast<uint32_t>(1ULL));
}
else {
c->R[rd] = arm_CPU_add(c,c->R[rs],operand,cast<uint32_t>(0ULL));
}
}
}
// tools/cpu/arm/thumbexec.go:115:1
void arm_CPU_thumbImm(arm_CPU* c,uint32_t h){
{
uint32_t rd = cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(8ULL))) & cast<uint32_t>(7ULL)));
uint32_t imm = cast<uint32_t>((h & cast<uint32_t>(255ULL)));
{
switch(cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(11ULL))) & cast<uint32_t>(3ULL)))){
case cast<uint32_t>(0ULL):{
c->R[rd] = imm;
arm_CPU_setNZ(c,imm);
break;}
case cast<uint32_t>(1ULL):{
arm_CPU_sub_(c,c->R[rd],imm,cast<uint32_t>(1ULL));
break;}
case cast<uint32_t>(2ULL):{
c->R[rd] = arm_CPU_add(c,c->R[rd],imm,cast<uint32_t>(0ULL));
break;}
default:{
c->R[rd] = arm_CPU_sub_(c,c->R[rd],imm,cast<uint32_t>(1ULL));
break;}
}}
}
}
// tools/cpu/arm/thumbexec.go:131:1
void arm_CPU_thumbALUExec(arm_CPU* c,uint32_t h){
{
auto tmp74 = std::make_tuple(cast<uint32_t>((h & cast<uint32_t>(7ULL))),cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(3ULL))) & cast<uint32_t>(7ULL))));
uint32_t rd = std::get<0>(tmp74);
uint32_t rs = std::get<1>(tmp74);
auto tmp75 = std::make_tuple(c->R[rd],c->R[rs]);
uint32_t a = std::get<0>(tmp75);
uint32_t b = std::get<1>(tmp75);
{
switch(cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(6ULL))) & cast<uint32_t>(15ULL)))){
case cast<uint32_t>(0ULL):{
c->R[rd] = cast<uint32_t>((a & b));
arm_CPU_setNZ(c,c->R[rd]);
break;}
case cast<uint32_t>(1ULL):{
c->R[rd] = cast<uint32_t>((a ^ b));
arm_CPU_setNZ(c,c->R[rd]);
break;}
case cast<uint32_t>(2ULL):{
auto tmp76 = arm_CPU_shift(c,cast<uint32_t>(0ULL),b,a,true,arm_CPU_boolToU(c,c->C));
uint32_t res = std::get<0>(tmp76);
uint32_t carry = std::get<1>(tmp76);
c->R[rd] = res;
arm_CPU_setNZ(c,res);
c->C = (carry == cast<uint32_t>(1ULL));
break;}
case cast<uint32_t>(3ULL):{
auto tmp77 = arm_CPU_shift(c,cast<uint32_t>(1ULL),b,a,true,arm_CPU_boolToU(c,c->C));
uint32_t res = std::get<0>(tmp77);
uint32_t carry = std::get<1>(tmp77);
c->R[rd] = res;
arm_CPU_setNZ(c,res);
c->C = (carry == cast<uint32_t>(1ULL));
break;}
case cast<uint32_t>(4ULL):{
auto tmp78 = arm_CPU_shift(c,cast<uint32_t>(2ULL),b,a,true,arm_CPU_boolToU(c,c->C));
uint32_t res = std::get<0>(tmp78);
uint32_t carry = std::get<1>(tmp78);
c->R[rd] = res;
arm_CPU_setNZ(c,res);
c->C = (carry == cast<uint32_t>(1ULL));
break;}
case cast<uint32_t>(5ULL):{
c->R[rd] = arm_CPU_add(c,a,b,arm_CPU_boolToU(c,c->C));
break;}
case cast<uint32_t>(6ULL):{
c->R[rd] = arm_CPU_sub_(c,a,b,arm_CPU_boolToU(c,c->C));
break;}
case cast<uint32_t>(7ULL):{
auto tmp79 = arm_CPU_shift(c,cast<uint32_t>(3ULL),b,a,true,arm_CPU_boolToU(c,c->C));
uint32_t res = std::get<0>(tmp79);
uint32_t carry = std::get<1>(tmp79);
c->R[rd] = res;
arm_CPU_setNZ(c,res);
c->C = (carry == cast<uint32_t>(1ULL));
break;}
case cast<uint32_t>(8ULL):{
arm_CPU_setNZ(c,cast<uint32_t>((a & b)));
break;}
case cast<uint32_t>(9ULL):{
c->R[rd] = arm_CPU_sub_(c,cast<uint32_t>(0ULL),b,cast<uint32_t>(1ULL));
break;}
case cast<uint32_t>(10ULL):{
arm_CPU_sub_(c,a,b,cast<uint32_t>(1ULL));
break;}
case cast<uint32_t>(11ULL):{
arm_CPU_add(c,a,b,cast<uint32_t>(0ULL));
break;}
case cast<uint32_t>(12ULL):{
c->R[rd] = cast<uint32_t>((a | b));
arm_CPU_setNZ(c,c->R[rd]);
break;}
case cast<uint32_t>(13ULL):{
c->R[rd] = cast<uint32_t>((a * b));
arm_CPU_setNZ(c,c->R[rd]);
break;}
case cast<uint32_t>(14ULL):{
c->R[rd] = (a & ~(b));
arm_CPU_setNZ(c,c->R[rd]);
break;}
default:{
c->R[rd] = cast<uint32_t>(~b);
arm_CPU_setNZ(c,c->R[rd]);
break;}
}}
}
}
// tools/cpu/arm/thumbexec.go:188:1
void arm_CPU_thumbHiRegExec(arm_CPU* c,uint32_t h){
{
uint32_t rd = cast<uint32_t>(((cast<uint32_t>((h & cast<uint32_t>(7ULL)))) | cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(4ULL))) & cast<uint32_t>(8ULL)))));
uint32_t rm = cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(3ULL))) & cast<uint32_t>(15ULL)));
{
switch(cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(8ULL))) & cast<uint32_t>(3ULL)))){
case cast<uint32_t>(0ULL):{
uint32_t v = cast<uint32_t>((arm_CPU_reg(c,rd) + arm_CPU_reg(c,rm)));
if ((rd == cast<uint32_t>(15ULL))) {
arm_CPU_branchTo(c,v);
}
else {
c->R[rd] = v;
}
break;}
case cast<uint32_t>(1ULL):{
arm_CPU_sub_(c,arm_CPU_reg(c,rd),arm_CPU_reg(c,rm),cast<uint32_t>(1ULL));
break;}
case cast<uint32_t>(2ULL):{
uint32_t v = arm_CPU_reg(c,rm);
if ((rd == cast<uint32_t>(15ULL))) {
arm_CPU_branchTo(c,v);
}
else {
c->R[rd] = v;
}
break;}
default:{
if ((cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(7ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
if ((!arm_Variant_v5OrLater(c->Arch))) {
arm_CPU_Halt(c,std::string("Thumb BLX (register) is undefined on ARMv4T: 0x%04X at 0x%08X",61),h,c->cur);
return ;
}
c->R[cast<int64_t>(14ULL)] = cast<uint32_t>(((cast<uint32_t>((c->cur + cast<uint32_t>(2ULL)))) | cast<uint32_t>(1ULL)));
}
arm_CPU_bxTo(c,arm_CPU_reg(c,rm));
break;}
}}
}
}
// tools/cpu/arm/thumbexec.go:220:1
void arm_CPU_thumbLoadStoreReg(arm_CPU* c,uint32_t h){
{
uint32_t addr = cast<uint32_t>((c->R[cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(3ULL))) & cast<uint32_t>(7ULL)))] + c->R[cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(6ULL))) & cast<uint32_t>(7ULL)))]));
uint32_t rd = cast<uint32_t>((h & cast<uint32_t>(7ULL)));
auto tmp80 = std::make_tuple(cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(11ULL))) & cast<uint32_t>(1ULL))),cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(10ULL))) & cast<uint32_t>(1ULL))));
uint32_t l = std::get<0>(tmp80);
uint32_t b = std::get<1>(tmp80);
{
if (((l == cast<uint32_t>(1ULL)) && (b == cast<uint32_t>(1ULL)))){
c->R[rd] = cast<uint32_t>(arm_CPU_read8(c,addr));
}
else if ((l == cast<uint32_t>(1ULL))){
c->R[rd] = arm_CPU_read32(c,addr);
}
else if ((b == cast<uint32_t>(1ULL))){
arm_CPU_write8(c,addr,cast<uint8_t>(c->R[rd]));
}
else {
arm_CPU_write32(c,addr,c->R[rd]);
}
}
tmp81:;
}
}
// tools/cpu/arm/thumbexec.go:236:1
void arm_CPU_thumbLoadStoreSExt(arm_CPU* c,uint32_t h){
{
uint32_t addr = cast<uint32_t>((c->R[cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(3ULL))) & cast<uint32_t>(7ULL)))] + c->R[cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(6ULL))) & cast<uint32_t>(7ULL)))]));
uint32_t rd = cast<uint32_t>((h & cast<uint32_t>(7ULL)));
{
switch(cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(10ULL))) & cast<uint32_t>(3ULL)))){
case cast<uint32_t>(0ULL):{
arm_CPU_write16(c,addr,c->R[rd]);
break;}
case cast<uint32_t>(1ULL):{
c->R[rd] = cast<uint32_t>(cast<int32_t>(cast<int8_t>(arm_CPU_read8(c,addr))));
break;}
case cast<uint32_t>(2ULL):{
c->R[rd] = arm_CPU_read16(c,addr);
break;}
default:{
c->R[rd] = cast<uint32_t>(cast<int32_t>(cast<int16_t>(arm_CPU_read16(c,addr))));
break;}
}}
}
}
// tools/cpu/arm/thumbexec.go:251:1
void arm_CPU_thumbLoadStoreImm(arm_CPU* c,uint32_t h){
{
auto tmp82 = std::make_tuple(cast<uint32_t>((h & cast<uint32_t>(7ULL))),cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(3ULL))) & cast<uint32_t>(7ULL))));
uint32_t rd = std::get<0>(tmp82);
uint32_t rb = std::get<1>(tmp82);
auto tmp83 = std::make_tuple(cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(11ULL))) & cast<uint32_t>(1ULL))),cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(12ULL))) & cast<uint32_t>(1ULL))));
uint32_t l = std::get<0>(tmp83);
uint32_t b = std::get<1>(tmp83);
uint32_t off = cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(6ULL))) & cast<uint32_t>(31ULL)));
if ((b == cast<uint32_t>(0ULL))) {
off *= cast<uint32_t>(4ULL);
}
uint32_t addr = cast<uint32_t>((c->R[rb] + off));
{
if (((l == cast<uint32_t>(1ULL)) && (b == cast<uint32_t>(1ULL)))){
c->R[rd] = cast<uint32_t>(arm_CPU_read8(c,addr));
}
else if ((l == cast<uint32_t>(1ULL))){
c->R[rd] = arm_CPU_read32(c,addr);
}
else if ((b == cast<uint32_t>(1ULL))){
arm_CPU_write8(c,addr,cast<uint8_t>(c->R[rd]));
}
else {
arm_CPU_write32(c,addr,c->R[rd]);
}
}
tmp84:;
}
}
// tools/cpu/arm/thumbexec.go:271:1
void arm_CPU_thumbPushPop(arm_CPU* c,uint32_t h){
{
uint32_t mask = cast<uint32_t>((h & cast<uint32_t>(255ULL)));
if ((cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(11ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
if ((cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(8ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
mask |= cast<uint32_t>(32768ULL);
}
uint32_t addr = c->R[cast<int64_t>(13ULL)];
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < cast<uint32_t>(16ULL));i++){
if ((cast<uint32_t>((mask & (shl<uint32_t>(cast<uint32_t>(1ULL),i)))) == cast<uint32_t>(0ULL))) {
continue;
}
uint32_t v = arm_CPU_read32aligned(c,addr);
if ((i == cast<uint32_t>(15ULL))) {
arm_CPU_bxTo(c,v);
}
else {
c->R[i] = v;
}
addr += cast<uint32_t>(4ULL);
}
}c->R[cast<int64_t>(13ULL)] = addr;
}
else {
if ((cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(8ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
mask |= cast<uint32_t>(16384ULL);
}
uint32_t addr = cast<uint32_t>((c->R[cast<int64_t>(13ULL)] - cast<uint32_t>((cast<uint32_t>(4ULL) * cast<uint32_t>(go_bits_OnesCount32(mask))))));
c->R[cast<int64_t>(13ULL)] = addr;
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < cast<uint32_t>(16ULL));i++){
if ((cast<uint32_t>((mask & (shl<uint32_t>(cast<uint32_t>(1ULL),i)))) == cast<uint32_t>(0ULL))) {
continue;
}
arm_CPU_write32aligned(c,addr,c->R[i]);
addr += cast<uint32_t>(4ULL);
}
}}
}
}
// tools/cpu/arm/thumbexec.go:307:1
void arm_CPU_thumbBlock(arm_CPU* c,uint32_t h){
{
uint32_t rb = cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(8ULL))) & cast<uint32_t>(7ULL)));
uint32_t mask = cast<uint32_t>((h & cast<uint32_t>(255ULL)));
uint32_t addr = c->R[rb];
bool load = (cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(11ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL));
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < cast<uint32_t>(8ULL));i++){
if ((cast<uint32_t>((mask & (shl<uint32_t>(cast<uint32_t>(1ULL),i)))) == cast<uint32_t>(0ULL))) {
continue;
}
if (load) {
c->R[i] = arm_CPU_read32aligned(c,addr);
}
else {
arm_CPU_write32aligned(c,addr,c->R[i]);
}
addr += cast<uint32_t>(4ULL);
}
}if ((!((load && (cast<uint32_t>((mask & (shl<uint32_t>(cast<uint32_t>(1ULL),rb)))) != cast<uint32_t>(0ULL)))))) {
c->R[rb] = addr;
}
}
}
// tools/cpu/arm/thumbexec.go:329:1
void arm_CPU_thumbCondBranch(arm_CPU* c,uint32_t h){
{
int64_t cond = cast<int64_t>(cast<uint32_t>(((shr<uint32_t>(h,cast<int64_t>(8ULL))) & cast<uint32_t>(15ULL))));
if ((cond == cast<int64_t>(15ULL))) {
if ((bool(c->SWI) && c->SWI(c,cast<uint32_t>((h & cast<uint32_t>(255ULL)))))) {
return ;
}
arm_CPU_exception(c,cast<uint32_t>(19ULL),cast<uint32_t>(8ULL),cast<uint32_t>((c->cur + cast<uint32_t>(2ULL))));
return ;
}
if ((cond == cast<int64_t>(14ULL))) {
arm_CPU_Halt(c,std::string("undefined Thumb 0x%04X at 0x%08X",32),h,c->cur);
return ;
}
if (arm_CPU_cond(c,cond)) {
arm_CPU_branchTo(c,cast<uint32_t>((cast<uint32_t>((c->cur + cast<uint32_t>(4ULL))) + shl<uint32_t>(arm_signExtend(cast<uint32_t>((h & cast<uint32_t>(255ULL))),cast<uint64_t>(8ULL)),cast<int64_t>(1ULL)))));
}
}
}
// tools/cpu/arm/variant.go:52:1
std::string arm_Variant_String(arm_Variant v){
{
{
switch(v){
case cast<arm_Variant>(0ULL):{
return std::string("ARMv5TE",7);
break;}
case cast<arm_Variant>(1ULL):{
return std::string("ARMv6K",6);
break;}
case cast<arm_Variant>(2ULL):{
return std::string("ARMv4T",6);
break;}
}}
return std::string("ARM?",4);
}
}
// tools/cpu/arm/variant.go:66:1
bool arm_Variant_isV6(arm_Variant v){
{
return (v == cast<arm_Variant>(1ULL));
}
}
// tools/cpu/arm/variant.go:70:1
bool arm_Variant_v5OrLater(arm_Variant v){
{
return (v != cast<arm_Variant>(2ULL));
}
}
// tools/cpu/arm/variant.go:73:1
arm_Inst arm_DecodeVariant(Slice<uint8_t> code,uint32_t addr,bool thumb,arm_Variant v){
{
if (thumb) {
return arm_DecodeThumb(code,addr);
}
return arm_DecodeARMVariant(code,addr,v);
}
}
// tools/cpu/arm/variant.go:88:1
arm_Inst arm_DecodeARMVariant(Slice<uint8_t> code,uint32_t addr,arm_Variant v){
{
if ((v == cast<arm_Variant>(2ULL))) {
return arm_v4tFilter(arm_DecodeARM(code,addr));
}
if ((!arm_Variant_isV6(v))) {
return arm_DecodeARM(code,addr);
}
auto tmp85 = arm_word(code);
uint32_t w = std::get<0>(tmp85);
bool ok = std::get<1>(tmp85);
if ((!ok)) {
return arm_Inst{addr,len(code),std::string(".word",5),std::string(".word ; truncated",17),cast<arm_Flow>(6ULL),{},{},{},{},cast<int64_t>(14ULL)};
}
arm_Inst in = arm_Inst{addr,cast<int64_t>(4ULL),{},{},cast<arm_Flow>(0ULL),{},{},{},{},cast<int64_t>(shr<uint32_t>(w,cast<int64_t>(28ULL)))};
{
auto tmp86 = arm_decodeARMv6(w,addr,in);
arm_Inst out = std::get<0>(tmp86);
bool handled = std::get<1>(tmp86);
if (handled) {
return out;
}
}
return arm_DecodeARM(code,addr);
}
}
// tools/cpu/arm/variant.go:126:1
arm_Inst arm_v4tFilter(arm_Inst in){
{
std::string mnem = in.Mnem;
{auto&& tmp87 = arm_condName;
for(int64_t tmp88=0;tmp88<len(tmp87);++tmp88){
auto c=tmp87[tmp88];if ((((c != std::string("",0)) && (len(mnem) > len(c))) && (sub(mnem,cast<int64_t>((len(mnem) - len(c))),len(mnem)) == c))) {
mnem = sub(mnem,0,cast<int64_t>((len(mnem) - len(c))));
break;
}
}}
if ((!get(arm_v4tMnem,mnem))) {
return in;
}
return arm_Inst{in.Addr,in.Len,std::string(".word",5),((std::string(".word ; ",8) + in.Text) + std::string(" \342\200\224 undefined on ARMv4T",24)),cast<arm_Flow>(6ULL),{},{},{},{},in.Cond};
}
}
// tools/cpu/arm/vfp.go:40:1
float arm_CPU_sGet(arm_CPU* c,uint32_t n){
{
return go_math_Float32frombits(c->VFP.S[n]);
}
}
// tools/cpu/arm/vfp.go:41:1
void arm_CPU_sSet(arm_CPU* c,uint32_t n,float v){
{
c->VFP.S[n] = go_math_Float32bits(v);
}
}
// tools/cpu/arm/vfp.go:42:1
uint32_t arm_CPU_sBits(arm_CPU* c,uint32_t n){
{
return c->VFP.S[n];
}
}
// tools/cpu/arm/vfp.go:43:1
void arm_CPU_sSetBits(arm_CPU* c,uint32_t n,uint32_t v){
{
c->VFP.S[n] = v;
}
}
// tools/cpu/arm/vfp.go:45:1
double arm_CPU_dGet(arm_CPU* c,uint32_t n){
{
return go_math_Float64frombits(cast<uint64_t>((cast<uint64_t>(c->VFP.S[cast<uint32_t>((cast<uint32_t>(2ULL) * n))]) | shl<uint64_t>(cast<uint64_t>(c->VFP.S[cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(2ULL) * n)) + cast<uint32_t>(1ULL)))]),cast<int64_t>(32ULL)))));
}
}
// tools/cpu/arm/vfp.go:48:1
void arm_CPU_dSet(arm_CPU* c,uint32_t n,double v){
{
uint64_t b = go_math_Float64bits(v);
c->VFP.S[cast<uint32_t>((cast<uint32_t>(2ULL) * n))] = cast<uint32_t>(b);
c->VFP.S[cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(2ULL) * n)) + cast<uint32_t>(1ULL)))] = cast<uint32_t>(shr<uint64_t>(b,cast<int64_t>(32ULL)));
}
}
// tools/cpu/arm/vfp.go:57:1
uint32_t arm_sReg(uint32_t v,uint32_t bit){
{
return cast<uint32_t>((shl<uint32_t>(v,cast<int64_t>(1ULL)) | bit));
}
}
// tools/cpu/arm/vfp.go:58:1
uint32_t arm_dReg(uint32_t v,uint32_t bit){
{
return cast<uint32_t>((shl<uint32_t>(bit,cast<int64_t>(4ULL)) | v));
}
}
// tools/cpu/arm/vfp.go:61:1
bool arm_isVFP(uint32_t w){
{
uint32_t cp = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(8ULL))) & cast<uint32_t>(15ULL)));
return ((cp == cast<uint32_t>(10ULL)) || (cp == cast<uint32_t>(11ULL)));
}
}
// tools/cpu/arm/vfp.go:72:1
arm_Inst arm_decodeVFP(uint32_t w,arm_Inst in){
{
bool single = (cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(8ULL))) & cast<uint32_t>(15ULL))) == cast<uint32_t>(10ULL));
std::string prec = std::string("F64",3);
if (single) {
prec = std::string("F32",3);
}
std::string sfx = arm_cn(in.Cond);
{
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(25ULL))) & cast<uint32_t>(7ULL))) == cast<uint32_t>(6ULL))){
if (((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(21ULL))) & cast<uint32_t>(127ULL))) == cast<uint32_t>(98ULL)) || (((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(23ULL))) & cast<uint32_t>(31ULL))) == cast<uint32_t>(24ULL)) && (cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(21ULL))) & cast<uint32_t>(3ULL))) == cast<uint32_t>(2ULL)))))) {
in.Mnem = (std::string("VMOV",4) + sfx);
in.Text = go_fmt_Sprintf(std::string("%s (64-bit move) ; 0x%08X",25),in.Mnem,w);
return in;
}
auto tmp90 = std::make_tuple(cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(24ULL))) & cast<uint32_t>(1ULL))),cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(21ULL))) & cast<uint32_t>(1ULL))),cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(20ULL))) & cast<uint32_t>(1ULL))));
uint32_t p = std::get<0>(tmp90);
uint32_t wb = std::get<1>(tmp90);
uint32_t l = std::get<2>(tmp90);
std::string rn = arm_regName[cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(15ULL)))];
std::string vd = arm_vfpRegName(single,cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(12ULL))) & cast<uint32_t>(15ULL))),cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(22ULL))) & cast<uint32_t>(1ULL))));
if (((p == cast<uint32_t>(1ULL)) && (wb == cast<uint32_t>(0ULL)))) {
std::string name = std::string("VSTR",4);
if ((l == cast<uint32_t>(1ULL))) {
name = std::string("VLDR",4);
}
uint32_t off = shl<uint32_t>((cast<uint32_t>((w & cast<uint32_t>(255ULL)))),cast<int64_t>(2ULL));
std::string sign = std::string("+",1);
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(23ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(0ULL))) {
sign = std::string("-",1);
}
in.Mnem = (name + sfx);
in.Text = go_fmt_Sprintf(std::string("%s %s, [%s, #%s0x%X]",20),in.Mnem,vd,rn,sign,off);
return in;
}
std::string name = std::string("VSTM",4);
if ((l == cast<uint32_t>(1ULL))) {
name = std::string("VLDM",4);
}
in.Mnem = (name + sfx);
in.Text = go_fmt_Sprintf(std::string("%s %s%s, %s{%d regs}",20),in.Mnem,rn,arm_wbMark(wb),std::string("",0),cast<uint32_t>((w & cast<uint32_t>(255ULL))));
return in;
}
else if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(24ULL))) & cast<uint32_t>(15ULL))) == cast<uint32_t>(14ULL))){
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(4ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
return arm_decodeVFPMove(w,in);
}
return arm_decodeVFPData(w,in,single,prec,sfx);
}
}
tmp89:;
in.Mnem = std::string("VFP?",4);
in.Text = go_fmt_Sprintf(std::string("VFP? 0x%08X",11),w);
return in;
}
}
// tools/cpu/arm/vfp.go:123:1
arm_Inst arm_decodeVFPMove(uint32_t w,arm_Inst in){
{
uint32_t l = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(20ULL))) & cast<uint32_t>(1ULL)));
std::string rt = arm_regName[cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(12ULL))) & cast<uint32_t>(15ULL)))];
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(21ULL))) & cast<uint32_t>(7ULL))) == cast<uint32_t>(7ULL))) {
if ((l == cast<uint32_t>(1ULL))) {
in.Mnem = (std::string("VMRS",4) + arm_cn(in.Cond));
std::string dst = rt;
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(12ULL))) & cast<uint32_t>(15ULL))) == cast<uint32_t>(15ULL))) {
dst = std::string("APSR_nzcv",9);
}
in.Text = go_fmt_Sprintf(std::string("%s %s, FPSCR",12),in.Mnem,dst);
}
else {
in.Mnem = (std::string("VMSR",4) + arm_cn(in.Cond));
in.Text = go_fmt_Sprintf(std::string("%s FPSCR, %s",12),in.Mnem,rt);
}
return in;
}
std::string sn = arm_vfpRegName(true,cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(15ULL))),cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(7ULL))) & cast<uint32_t>(1ULL))));
in.Mnem = (std::string("VMOV",4) + arm_cn(in.Cond));
if ((l == cast<uint32_t>(1ULL))) {
in.Text = go_fmt_Sprintf(std::string("%s %s, %s",9),in.Mnem,rt,sn);
}
else {
in.Text = go_fmt_Sprintf(std::string("%s %s, %s",9),in.Mnem,sn,rt);
}
return in;
}
}
// tools/cpu/arm/vfp.go:150:1
arm_Inst arm_decodeVFPData(uint32_t w,arm_Inst in,bool single,std::string prec,std::string sfx){
{
uint32_t pqr = cast<uint32_t>((cast<uint32_t>((shl<uint32_t>(cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(23ULL))) & cast<uint32_t>(1ULL))),cast<int64_t>(2ULL)) | shl<uint32_t>(cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(21ULL))) & cast<uint32_t>(1ULL))),cast<int64_t>(1ULL)))) | cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(20ULL))) & cast<uint32_t>(1ULL)))));
uint32_t op = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(6ULL))) & cast<uint32_t>(1ULL)));
std::string vd = arm_vfpRegName(single,cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(12ULL))) & cast<uint32_t>(15ULL))),cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(22ULL))) & cast<uint32_t>(1ULL))));
std::string vn = arm_vfpRegName(single,cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(15ULL))),cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(7ULL))) & cast<uint32_t>(1ULL))));
std::string vm = arm_vfpRegName(single,cast<uint32_t>((w & cast<uint32_t>(15ULL))),cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(5ULL))) & cast<uint32_t>(1ULL))));
std::string name = std::string("",0);
{
switch(pqr){
case cast<uint32_t>(3ULL):{
if ((op == cast<uint32_t>(0ULL))) {
name = std::string("VADD",4);
}
else {
name = std::string("VSUB",4);
}
break;}
case cast<uint32_t>(2ULL):{
if ((op == cast<uint32_t>(0ULL))) {
name = std::string("VMUL",4);
}
else {
name = std::string("VNMUL",5);
}
break;}
case cast<uint32_t>(4ULL):{
name = std::string("VDIV",4);
break;}
case cast<uint32_t>(0ULL):{
if ((op == cast<uint32_t>(0ULL))) {
name = std::string("VMLA",4);
}
else {
name = std::string("VMLS",4);
}
break;}
case cast<uint32_t>(1ULL):{
if ((op == cast<uint32_t>(0ULL))) {
name = std::string("VNMLS",5);
}
else {
name = std::string("VNMLA",5);
}
break;}
case cast<uint32_t>(7ULL):{
return arm_decodeVFPExt(w,in,single,prec,sfx,vd,vm);
break;}
}}
if ((name == std::string("",0))) {
auto tmp91 = std::make_tuple(std::string("VFP?",4),go_fmt_Sprintf(std::string("VFP? 0x%08X",11),w));
in.Mnem = std::get<0>(tmp91);
in.Text = std::get<1>(tmp91);
return in;
}
in.Mnem = (((name + sfx) + std::string(".",1)) + prec);
in.Text = go_fmt_Sprintf(std::string("%s %s, %s, %s",13),in.Mnem,vd,vn,vm);
return in;
}
}
// tools/cpu/arm/vfp.go:197:1
arm_Inst arm_decodeVFPExt(uint32_t w,arm_Inst in,bool single,std::string prec,std::string sfx,std::string vd,std::string vm){
{
uint32_t opc2 = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(15ULL)));
uint32_t opc3 = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(6ULL))) & cast<uint32_t>(3ULL)));
if ((cast<uint32_t>((opc3 & cast<uint32_t>(1ULL))) == cast<uint32_t>(0ULL))) {
in.Mnem = (((std::string("VMOV",4) + sfx) + std::string(".",1)) + prec);
in.Text = go_fmt_Sprintf(std::string("%s %s, #imm",11),in.Mnem,vd);
return in;
}
std::string name = std::string("",0);
{
switch(opc2){
case cast<uint32_t>(0ULL):{
if ((opc3 == cast<uint32_t>(1ULL))) {
name = std::string("VMOV",4);
}
else {
name = std::string("VABS",4);
}
break;}
case cast<uint32_t>(1ULL):{
if ((opc3 == cast<uint32_t>(1ULL))) {
name = std::string("VNEG",4);
}
else {
name = std::string("VSQRT",5);
}
break;}
case cast<uint32_t>(4ULL):case cast<uint32_t>(5ULL):{
name = std::string("VCMP",4);
break;}
case cast<uint32_t>(7ULL):{
name = std::string("VCVT",4);
break;}
case cast<uint32_t>(8ULL):case cast<uint32_t>(12ULL):case cast<uint32_t>(13ULL):{
name = std::string("VCVT",4);
break;}
}}
if ((name == std::string("",0))) {
auto tmp92 = std::make_tuple(std::string("VFP?",4),go_fmt_Sprintf(std::string("VFP? 0x%08X",11),w));
in.Mnem = std::get<0>(tmp92);
in.Text = std::get<1>(tmp92);
return in;
}
in.Mnem = (((name + sfx) + std::string(".",1)) + prec);
in.Text = go_fmt_Sprintf(std::string("%s %s, %s",9),in.Mnem,vd,vm);
return in;
}
}
// tools/cpu/arm/vfp.go:235:1
std::string arm_vfpRegName(bool single,uint32_t v,uint32_t bit){
{
if (single) {
return go_fmt_Sprintf(std::string("s%d",3),arm_sReg(v,bit));
}
return go_fmt_Sprintf(std::string("d%d",3),arm_dReg(v,bit));
}
}
// tools/cpu/arm/vfp.go:242:1
std::string arm_wbMark(uint32_t wb){
{
if ((wb == cast<uint32_t>(1ULL))) {
return std::string("!",1);
}
return std::string("",0);
}
}
// tools/cpu/arm/vfp.go:253:1
bool arm_CPU_execVFP(arm_CPU* c,uint32_t w){
{
if ((!arm_isVFP(w))) {
return false;
}
{
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(25ULL))) & cast<uint32_t>(7ULL))) == cast<uint32_t>(6ULL))){
if (((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(21ULL))) & cast<uint32_t>(127ULL))) == cast<uint32_t>(98ULL)) || ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(23ULL))) & cast<uint32_t>(31ULL))) == cast<uint32_t>(24ULL)) && (cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(21ULL))) & cast<uint32_t>(3ULL))) == cast<uint32_t>(2ULL))))) {
return arm_CPU_execVFPMove64(c,w);
}
arm_CPU_execVFPLoadStore(c,w);
return true;
}
else if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(24ULL))) & cast<uint32_t>(15ULL))) == cast<uint32_t>(14ULL))){
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(4ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
return arm_CPU_execVFPMove(c,w);
}
return arm_CPU_execVFPData(c,w);
}
}
tmp93:;
return false;
}
}
// tools/cpu/arm/vfp.go:275:1
void arm_CPU_execVFPLoadStore(arm_CPU* c,uint32_t w){
{
uint32_t p = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(24ULL))) & cast<uint32_t>(1ULL)));
uint32_t u = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(23ULL))) & cast<uint32_t>(1ULL)));
uint32_t d = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(22ULL))) & cast<uint32_t>(1ULL)));
uint32_t wb = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(21ULL))) & cast<uint32_t>(1ULL)));
uint32_t l = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(20ULL))) & cast<uint32_t>(1ULL)));
uint32_t rn = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(15ULL)));
uint32_t vd = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(12ULL))) & cast<uint32_t>(15ULL)));
bool single = (cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(8ULL))) & cast<uint32_t>(15ULL))) == cast<uint32_t>(10ULL));
uint32_t imm8 = cast<uint32_t>((w & cast<uint32_t>(255ULL)));
uint32_t base = arm_CPU_reg(c,rn);
if (((p == cast<uint32_t>(1ULL)) && (wb == cast<uint32_t>(0ULL)))) {
uint32_t offset = shl<uint32_t>(imm8,cast<int64_t>(2ULL));
uint32_t addr = cast<uint32_t>((base + offset));
if ((u == cast<uint32_t>(0ULL))) {
addr = cast<uint32_t>((base - offset));
}
if (single) {
uint32_t sd = arm_sReg(vd,d);
if ((l == cast<uint32_t>(1ULL))) {
arm_CPU_sSetBits(c,sd,arm_CPU_read32(c,addr));
}
else {
arm_CPU_write32(c,addr,arm_CPU_sBits(c,sd));
}
}
else {
uint32_t dd = arm_dReg(vd,d);
if ((l == cast<uint32_t>(1ULL))) {
c->VFP.S[cast<uint32_t>((cast<uint32_t>(2ULL) * dd))] = arm_CPU_read32(c,addr);
c->VFP.S[cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(2ULL) * dd)) + cast<uint32_t>(1ULL)))] = arm_CPU_read32(c,cast<uint32_t>((addr + cast<uint32_t>(4ULL))));
}
else {
arm_CPU_write32(c,addr,c->VFP.S[cast<uint32_t>((cast<uint32_t>(2ULL) * dd))]);
arm_CPU_write32(c,cast<uint32_t>((addr + cast<uint32_t>(4ULL))),c->VFP.S[cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(2ULL) * dd)) + cast<uint32_t>(1ULL)))]);
}
}
return ;
}
uint32_t count = imm8;
uint32_t regs = count;
if ((!single)) {
regs = divi<uint32_t>(count,cast<uint32_t>(2ULL));
}
uint32_t addr = base;
if ((u == cast<uint32_t>(0ULL))) {
addr = cast<uint32_t>((base - cast<uint32_t>((count * cast<uint32_t>(4ULL)))));
}
uint32_t start = addr;
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < regs);i++){
if (single) {
uint32_t sd = cast<uint32_t>((arm_sReg(vd,d) + i));
if ((l == cast<uint32_t>(1ULL))) {
arm_CPU_sSetBits(c,sd,arm_CPU_read32(c,addr));
}
else {
arm_CPU_write32(c,addr,arm_CPU_sBits(c,sd));
}
addr += cast<uint32_t>(4ULL);
}
else {
uint32_t dd = cast<uint32_t>((arm_dReg(vd,d) + i));
if ((l == cast<uint32_t>(1ULL))) {
c->VFP.S[cast<uint32_t>((cast<uint32_t>(2ULL) * dd))] = arm_CPU_read32(c,addr);
c->VFP.S[cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(2ULL) * dd)) + cast<uint32_t>(1ULL)))] = arm_CPU_read32(c,cast<uint32_t>((addr + cast<uint32_t>(4ULL))));
}
else {
arm_CPU_write32(c,addr,c->VFP.S[cast<uint32_t>((cast<uint32_t>(2ULL) * dd))]);
arm_CPU_write32(c,cast<uint32_t>((addr + cast<uint32_t>(4ULL))),c->VFP.S[cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(2ULL) * dd)) + cast<uint32_t>(1ULL)))]);
}
addr += cast<uint32_t>(8ULL);
}
}
}if ((wb == cast<uint32_t>(1ULL))) {
if ((u == cast<uint32_t>(1ULL))) {
arm_CPU_setReg(c,rn,cast<uint32_t>((base + cast<uint32_t>((count * cast<uint32_t>(4ULL))))));
}
else {
arm_CPU_setReg(c,rn,start);
}
}
}
}
// tools/cpu/arm/vfp.go:363:1
bool arm_CPU_execVFPData(arm_CPU* c,uint32_t w){
{
uint32_t pqr = cast<uint32_t>((cast<uint32_t>((shl<uint32_t>(cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(23ULL))) & cast<uint32_t>(1ULL))),cast<int64_t>(2ULL)) | shl<uint32_t>(cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(21ULL))) & cast<uint32_t>(1ULL))),cast<int64_t>(1ULL)))) | cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(20ULL))) & cast<uint32_t>(1ULL)))));
bool single = (cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(8ULL))) & cast<uint32_t>(15ULL))) == cast<uint32_t>(10ULL));
uint32_t op = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(6ULL))) & cast<uint32_t>(1ULL)));
uint32_t dn = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(15ULL)));
uint32_t dd = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(12ULL))) & cast<uint32_t>(15ULL)));
uint32_t dm = cast<uint32_t>((w & cast<uint32_t>(15ULL)));
uint32_t nBit = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(7ULL))) & cast<uint32_t>(1ULL)));
uint32_t mBit = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(5ULL))) & cast<uint32_t>(1ULL)));
uint32_t dBit = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(22ULL))) & cast<uint32_t>(1ULL)));
uint32_t vd={};
uint32_t vn={};
uint32_t vm={};
if (single) {
auto tmp94 = std::make_tuple(arm_sReg(dd,dBit),arm_sReg(dn,nBit),arm_sReg(dm,mBit));
vd = std::get<0>(tmp94);
vn = std::get<1>(tmp94);
vm = std::get<2>(tmp94);
}
else {
auto tmp95 = std::make_tuple(arm_dReg(dd,dBit),arm_dReg(dn,nBit),arm_dReg(dm,mBit));
vd = std::get<0>(tmp95);
vn = std::get<1>(tmp95);
vm = std::get<2>(tmp95);
}
auto binop = [&](std::function<float(float,float)> f32,std::function<double(double,double)> f64)->void{
if (single) {
arm_CPU_sSet(c,vd,f32(arm_CPU_sGet(c,vn),arm_CPU_sGet(c,vm)));
}
else {
arm_CPU_dSet(c,vd,f64(arm_CPU_dGet(c,vn),arm_CPU_dGet(c,vm)));
}
}
;
{
switch(pqr){
case cast<uint32_t>(3ULL):{
if ((op == cast<uint32_t>(0ULL))) {
binop([&](float a,float b)->float{
return (a + b);
}
,[&](double a,double b)->double{
return (a + b);
}
);
}
else {
binop([&](float a,float b)->float{
return (a - b);
}
,[&](double a,double b)->double{
return (a - b);
}
);
}
return true;
break;}
case cast<uint32_t>(2ULL):{
binop([&](float a,float b)->float{
return (a * b);
}
,[&](double a,double b)->double{
return (a * b);
}
);
if ((op == cast<uint32_t>(1ULL))) {
arm_CPU_vfpNegate(c,vd,single);
}
return true;
break;}
case cast<uint32_t>(4ULL):{
binop([&](float a,float b)->float{
return (a / b);
}
,[&](double a,double b)->double{
return (a / b);
}
);
return true;
break;}
case cast<uint32_t>(0ULL):{
arm_CPU_vfpMulAcc(c,vd,vn,vm,single,(op == cast<uint32_t>(1ULL)),false);
return true;
break;}
case cast<uint32_t>(1ULL):{
arm_CPU_vfpMulAcc(c,vd,vn,vm,single,(op == cast<uint32_t>(1ULL)),true);
return true;
break;}
case cast<uint32_t>(7ULL):{
return arm_CPU_execVFPExt(c,w,single,vd,vm);
break;}
}}
arm_CPU_Halt(c,std::string("unimplemented VFP data op %03b (0x%08X) at 0x%08X",49),pqr,w,c->cur);
return true;
}
}
// tools/cpu/arm/vfp.go:426:1
void arm_CPU_vfpNegate(arm_CPU* c,uint32_t vd,bool single){
{
if (single) {
arm_CPU_sSet(c,vd,cast<float>(-arm_CPU_sGet(c,vd)));
}
else {
arm_CPU_dSet(c,vd,cast<double>(-arm_CPU_dGet(c,vd)));
}
}
}
// tools/cpu/arm/vfp.go:436:1
void arm_CPU_vfpMulAcc(arm_CPU* c,uint32_t vd,uint32_t vn,uint32_t vm,bool single,bool sub_,bool negAcc){
{
if (single) {
float acc = arm_CPU_sGet(c,vd);
if (negAcc) {
acc = cast<float>(-acc);
}
float p = (arm_CPU_sGet(c,vn) * arm_CPU_sGet(c,vm));
if (sub_) {
p = cast<float>(-p);
}
arm_CPU_sSet(c,vd,(acc + p));
}
else {
double acc = arm_CPU_dGet(c,vd);
if (negAcc) {
acc = cast<double>(-acc);
}
double p = (arm_CPU_dGet(c,vn) * arm_CPU_dGet(c,vm));
if (sub_) {
p = cast<double>(-p);
}
arm_CPU_dSet(c,vd,(acc + p));
}
}
}
// tools/cpu/arm/vfp.go:462:1
bool arm_CPU_execVFPExt(arm_CPU* c,uint32_t w,bool single,uint32_t vd,uint32_t vm){
{
uint32_t opc2 = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(15ULL)));
uint32_t opc3 = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(6ULL))) & cast<uint32_t>(3ULL)));
if ((cast<uint32_t>((opc3 & cast<uint32_t>(1ULL))) == cast<uint32_t>(0ULL))) {
uint64_t imm = arm_vfpExpandImm(w,single);
if (single) {
arm_CPU_sSetBits(c,vd,cast<uint32_t>(imm));
}
else {
c->VFP.S[cast<uint32_t>((cast<uint32_t>(2ULL) * vd))] = cast<uint32_t>(imm);
c->VFP.S[cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(2ULL) * vd)) + cast<uint32_t>(1ULL)))] = cast<uint32_t>(shr<uint64_t>(imm,cast<int64_t>(32ULL)));
}
return true;
}
{
switch(opc2){
case cast<uint32_t>(0ULL):{
if ((opc3 == cast<uint32_t>(1ULL))) {
arm_CPU_vfpCopy(c,vd,vm,single);
}
else {
arm_CPU_vfpUnary(c,vd,vm,single,go_math_Abs);
}
return true;
break;}
case cast<uint32_t>(1ULL):{
if ((opc3 == cast<uint32_t>(1ULL))) {
arm_CPU_vfpUnary(c,vd,vm,single,[&](double f)->double{
return cast<double>(-f);
}
);
}
else {
arm_CPU_vfpUnary(c,vd,vm,single,go_math_Sqrt);
}
return true;
break;}
case cast<uint32_t>(4ULL):case cast<uint32_t>(5ULL):{
arm_CPU_execVCMP(c,w,single,vd,vm,(opc2 == cast<uint32_t>(5ULL)));
return true;
break;}
case cast<uint32_t>(7ULL):{
arm_CPU_execVCVTPrec(c,w,single);
return true;
break;}
case cast<uint32_t>(8ULL):{
arm_CPU_execVCVTFromInt(c,w,single);
return true;
break;}
case cast<uint32_t>(12ULL):case cast<uint32_t>(13ULL):{
arm_CPU_execVCVTToInt(c,w,single);
return true;
break;}
}}
arm_CPU_Halt(c,std::string("unimplemented VFP ext op2=0x%X (0x%08X) at 0x%08X",49),opc2,w,c->cur);
return true;
}
}
// tools/cpu/arm/vfp.go:510:1
void arm_CPU_vfpCopy(arm_CPU* c,uint32_t vd,uint32_t vm,bool single){
{
if (single) {
arm_CPU_sSetBits(c,vd,arm_CPU_sBits(c,vm));
}
else {
c->VFP.S[cast<uint32_t>((cast<uint32_t>(2ULL) * vd))] = c->VFP.S[cast<uint32_t>((cast<uint32_t>(2ULL) * vm))];
c->VFP.S[cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(2ULL) * vd)) + cast<uint32_t>(1ULL)))] = c->VFP.S[cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(2ULL) * vm)) + cast<uint32_t>(1ULL)))];
}
}
}
// tools/cpu/arm/vfp.go:519:1
void arm_CPU_vfpUnary(arm_CPU* c,uint32_t vd,uint32_t vm,bool single,std::function<double(double)> f){
{
if (single) {
arm_CPU_sSet(c,vd,cast<float>(f(cast<double>(arm_CPU_sGet(c,vm)))));
}
else {
arm_CPU_dSet(c,vd,f(arm_CPU_dGet(c,vm)));
}
}
}
// tools/cpu/arm/vfp.go:530:1
void arm_CPU_execVCMP(arm_CPU* c,uint32_t w,bool single,uint32_t vd,uint32_t vm,bool withZero){
{
double a={};
double b={};
if (single) {
a = cast<double>(arm_CPU_sGet(c,vd));
if ((!withZero)) {
b = cast<double>(arm_CPU_sGet(c,vm));
}
}
else {
a = arm_CPU_dGet(c,vd);
if ((!withZero)) {
b = arm_CPU_dGet(c,vm);
}
}
bool n={};
bool z={};
bool cc={};
bool v={};
{
if ((go_math_IsNaN(a) || go_math_IsNaN(b))){
auto tmp97 = std::make_tuple(true,true);
cc = std::get<0>(tmp97);
v = std::get<1>(tmp97);
}
else if ((a == b)){
auto tmp98 = std::make_tuple(true,true);
z = std::get<0>(tmp98);
cc = std::get<1>(tmp98);
}
else if ((a < b)){
n = true;
}
else {
cc = true;
}
}
tmp96:;
uint32_t f = (c->VFP.FPSCR & ~(cast<uint32_t>(4026531840ULL)));
if (n) {
f |= cast<uint32_t>(2147483648ULL);
}
if (z) {
f |= cast<uint32_t>(1073741824ULL);
}
if (cc) {
f |= cast<uint32_t>(536870912ULL);
}
if (v) {
f |= cast<uint32_t>(268435456ULL);
}
c->VFP.FPSCR = f;
}
}
// tools/cpu/arm/vfp.go:575:1
uint32_t arm_vdOf(uint32_t w,bool single){
{
if (single) {
return arm_sReg(cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(12ULL))) & cast<uint32_t>(15ULL))),cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(22ULL))) & cast<uint32_t>(1ULL))));
}
return arm_dReg(cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(12ULL))) & cast<uint32_t>(15ULL))),cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(22ULL))) & cast<uint32_t>(1ULL))));
}
}
// tools/cpu/arm/vfp.go:581:1
uint32_t arm_vmOf(uint32_t w,bool single){
{
if (single) {
return arm_sReg(cast<uint32_t>((w & cast<uint32_t>(15ULL))),cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(5ULL))) & cast<uint32_t>(1ULL))));
}
return arm_dReg(cast<uint32_t>((w & cast<uint32_t>(15ULL))),cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(5ULL))) & cast<uint32_t>(1ULL))));
}
}
// tools/cpu/arm/vfp.go:591:1
void arm_CPU_execVCVTPrec(arm_CPU* c,uint32_t w,bool srcSingle){
{
if (srcSingle) {
arm_CPU_dSet(c,arm_vdOf(w,false),cast<double>(arm_CPU_sGet(c,arm_vmOf(w,true))));
}
else {
arm_CPU_sSet(c,arm_vdOf(w,true),cast<float>(arm_CPU_dGet(c,arm_vmOf(w,false))));
}
}
}
// tools/cpu/arm/vfp.go:601:1
void arm_CPU_execVCVTFromInt(arm_CPU* c,uint32_t w,bool single){
{
bool signed_ = (cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(7ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL));
uint32_t bits = arm_CPU_sBits(c,arm_vmOf(w,true));
double f={};
if (signed_) {
f = cast<double>(cast<int32_t>(bits));
}
else {
f = cast<double>(bits);
}
if (single) {
arm_CPU_sSet(c,arm_vdOf(w,true),cast<float>(f));
}
else {
arm_CPU_dSet(c,arm_vdOf(w,false),f);
}
}
}
// tools/cpu/arm/vfp.go:620:1
void arm_CPU_execVCVTToInt(arm_CPU* c,uint32_t w,bool single){
{
bool signed_ = (cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL));
double f={};
if (single) {
f = cast<double>(arm_CPU_sGet(c,arm_vmOf(w,true)));
}
else {
f = arm_CPU_dGet(c,arm_vmOf(w,false));
}
uint32_t out={};
if (signed_) {
out = cast<uint32_t>(cast<int32_t>(f));
}
else {
if ((f < cast<double>(0ULL))) {
f = cast<double>(0ULL);
}
out = cast<uint32_t>(f);
}
arm_CPU_sSetBits(c,arm_vdOf(w,true),out);
}
}
// tools/cpu/arm/vfp.go:641:1
bool arm_CPU_execVFPMove(arm_CPU* c,uint32_t w){
{
uint32_t l = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(20ULL))) & cast<uint32_t>(1ULL)));
uint32_t rt = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(12ULL))) & cast<uint32_t>(15ULL)));
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(21ULL))) & cast<uint32_t>(7ULL))) == cast<uint32_t>(7ULL))) {
uint32_t sysreg = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(15ULL)));
if ((l == cast<uint32_t>(1ULL))) {
if ((sysreg == cast<uint32_t>(1ULL))) {
if ((rt == cast<uint32_t>(15ULL))) {
c->N = (cast<uint32_t>((c->VFP.FPSCR & cast<uint32_t>(2147483648ULL))) != cast<uint32_t>(0ULL));
c->Z = (cast<uint32_t>((c->VFP.FPSCR & cast<uint32_t>(1073741824ULL))) != cast<uint32_t>(0ULL));
c->C = (cast<uint32_t>((c->VFP.FPSCR & cast<uint32_t>(536870912ULL))) != cast<uint32_t>(0ULL));
c->V = (cast<uint32_t>((c->VFP.FPSCR & cast<uint32_t>(268435456ULL))) != cast<uint32_t>(0ULL));
}
else {
arm_CPU_setReg(c,rt,c->VFP.FPSCR);
}
}
else if ((sysreg == cast<uint32_t>(0ULL))) {
arm_CPU_setReg(c,rt,cast<uint32_t>(1090592948ULL));
}
else {
arm_CPU_setReg(c,rt,cast<uint32_t>(0ULL));
}
}
else {
if ((sysreg == cast<uint32_t>(1ULL))) {
c->VFP.FPSCR = arm_CPU_reg(c,rt);
}
}
return true;
}
uint32_t vn = arm_sReg(cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(15ULL))),cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(7ULL))) & cast<uint32_t>(1ULL))));
if ((l == cast<uint32_t>(1ULL))) {
arm_CPU_setReg(c,rt,arm_CPU_sBits(c,vn));
}
else {
arm_CPU_sSetBits(c,vn,arm_CPU_reg(c,rt));
}
return true;
}
}
// tools/cpu/arm/vfp.go:683:1
bool arm_CPU_execVFPMove64(arm_CPU* c,uint32_t w){
{
uint32_t l = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(20ULL))) & cast<uint32_t>(1ULL)));
uint32_t rt = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(12ULL))) & cast<uint32_t>(15ULL)));
uint32_t rt2 = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(15ULL)));
bool single = (cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(8ULL))) & cast<uint32_t>(15ULL))) == cast<uint32_t>(10ULL));
uint32_t mBit = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(5ULL))) & cast<uint32_t>(1ULL)));
uint32_t vm = cast<uint32_t>((w & cast<uint32_t>(15ULL)));
if (single) {
uint32_t sm = arm_sReg(vm,mBit);
if ((l == cast<uint32_t>(1ULL))) {
arm_CPU_setReg(c,rt,arm_CPU_sBits(c,sm));
arm_CPU_setReg(c,rt2,arm_CPU_sBits(c,cast<uint32_t>((sm + cast<uint32_t>(1ULL)))));
}
else {
arm_CPU_sSetBits(c,sm,arm_CPU_reg(c,rt));
arm_CPU_sSetBits(c,cast<uint32_t>((sm + cast<uint32_t>(1ULL))),arm_CPU_reg(c,rt2));
}
}
else {
uint32_t dm = arm_dReg(vm,mBit);
if ((l == cast<uint32_t>(1ULL))) {
arm_CPU_setReg(c,rt,c->VFP.S[cast<uint32_t>((cast<uint32_t>(2ULL) * dm))]);
arm_CPU_setReg(c,rt2,c->VFP.S[cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(2ULL) * dm)) + cast<uint32_t>(1ULL)))]);
}
else {
c->VFP.S[cast<uint32_t>((cast<uint32_t>(2ULL) * dm))] = arm_CPU_reg(c,rt);
c->VFP.S[cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(2ULL) * dm)) + cast<uint32_t>(1ULL)))] = arm_CPU_reg(c,rt2);
}
}
return true;
}
}
// tools/cpu/arm/vfp.go:723:1
uint64_t arm_vfpExpandImm(uint32_t w,bool single){
{
uint32_t imm8 = cast<uint32_t>(((shl<uint32_t>((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(15ULL)))),cast<int64_t>(4ULL))) | (cast<uint32_t>((w & cast<uint32_t>(15ULL))))));
uint32_t sign = cast<uint32_t>(((shr<uint32_t>(imm8,cast<int64_t>(7ULL))) & cast<uint32_t>(1ULL)));
uint32_t b = cast<uint32_t>(((shr<uint32_t>(imm8,cast<int64_t>(6ULL))) & cast<uint32_t>(1ULL)));
if (single) {
uint32_t exp = cast<uint32_t>((cast<uint32_t>(((shl<uint32_t>((cast<uint32_t>((cast<uint32_t>(~b) & cast<uint32_t>(1ULL)))),cast<int64_t>(7ULL))) | (shl<uint32_t>(arm_repeat(b,cast<uint32_t>(5ULL)),cast<int64_t>(2ULL))))) | (cast<uint32_t>(((shr<uint32_t>(imm8,cast<int64_t>(4ULL))) & cast<uint32_t>(3ULL))))));
uint32_t frac = cast<uint32_t>((imm8 & cast<uint32_t>(15ULL)));
return cast<uint64_t>(cast<uint32_t>((cast<uint32_t>((shl<uint32_t>(sign,cast<int64_t>(31ULL)) | shl<uint32_t>(exp,cast<int64_t>(23ULL)))) | shl<uint32_t>(frac,cast<int64_t>(19ULL)))));
}
uint64_t exp = (cast<uint64_t>((cast<uint64_t>(((shl<uint64_t>(cast<uint64_t>(cast<uint32_t>((cast<uint32_t>(~b) & cast<uint32_t>(1ULL)))),cast<int64_t>(10ULL))) | (shl<uint64_t>(arm_repeat64(cast<uint64_t>(b),cast<uint32_t>(8ULL)),cast<int64_t>(2ULL))))) | cast<uint64_t>(cast<uint32_t>(((shr<uint32_t>(imm8,cast<int64_t>(4ULL))) & cast<uint32_t>(3ULL)))))));
uint64_t frac = cast<uint64_t>(cast<uint32_t>((imm8 & cast<uint32_t>(15ULL))));
return cast<uint64_t>((cast<uint64_t>((shl<uint64_t>(cast<uint64_t>(sign),cast<int64_t>(63ULL)) | shl<uint64_t>(exp,cast<int64_t>(52ULL)))) | shl<uint64_t>(frac,cast<int64_t>(48ULL))));
}
}
// tools/cpu/arm/vfp.go:738:1
uint32_t arm_repeat(uint32_t bit,uint32_t n){
{
uint32_t v={};
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < n);i++){
v = cast<uint32_t>((shl<uint32_t>(v,cast<int64_t>(1ULL)) | (cast<uint32_t>((bit & cast<uint32_t>(1ULL))))));
}
}return v;
}
}
// tools/cpu/arm/vfp.go:745:1
uint64_t arm_repeat64(uint64_t bit,uint32_t n){
{
uint64_t v={};
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < n);i++){
v = cast<uint64_t>((shl<uint64_t>(v,cast<int64_t>(1ULL)) | (cast<uint64_t>((bit & cast<uint64_t>(1ULL))))));
}
}return v;
}
}
// tools/platform/gba/rom.go:94:1
uint8_t gba_ComplementCheck(Slice<uint8_t> data){
{
uint8_t sum={};
{auto&& tmp1 = sub(data,cast<int64_t>(160ULL),cast<int64_t>(189ULL));
for(int64_t tmp2=0;tmp2<len(tmp1);++tmp2){
auto b=tmp1[tmp2];sum += b;
}}
return cast<uint8_t>(-(cast<uint8_t>((sum + cast<uint8_t>(25ULL)))));
}
}
// tools/platform/gba/rom.go:116:1
std::tuple<std::string,std::string> gba_ROM_SaveType(gba_ROM* r){
std::string id{};
std::string desc{};
{
std::string s = cast<std::string>(r->Data);
{auto&& tmp3 = gba_saveIDs;
for(int64_t tmp4=0;tmp4<len(tmp3);++tmp4){
auto c=tmp3[tmp4];int64_t i = go_strings_Index(s,c.id);
if ((i < cast<int64_t>(0ULL))) {
continue;
}
int64_t j = cast<int64_t>((i + len(c.id)));
{;for (;(((j < len(s)) && (cast<uint8_t>(s[j]) >= cast<uint8_t>(48ULL))) && (cast<uint8_t>(s[j]) <= cast<uint8_t>(57ULL)));){
j++;
}
}return {sub(s,i,j),c.desc};
}}
return {std::string("",0),std::string("",0)};
}
}
// tools/platform/gba/gbamachine/bios.go:26:1
std::function<bool(arm_CPU*,uint32_t)> gbamachine_biosSWI(gbamachine_Machine* m){
{
return [m](arm_CPU* c,uint32_t comment)->bool{gbamachine_bus busStorage{m};auto*b=&busStorage;
uint32_t n = cast<uint32_t>((comment & cast<uint32_t>(255ULL)));
if ((!c->Thumb)) {
n = cast<uint32_t>(((shr<uint32_t>(comment,cast<int64_t>(16ULL))) & cast<uint32_t>(255ULL)));
}
{
switch(n){
case cast<uint32_t>(1ULL):{
uint32_t f = c->R[cast<int64_t>(0ULL)];
if ((cast<uint32_t>((f & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL))) {
clear(m->ewram);
}
if ((cast<uint32_t>((f & cast<uint32_t>(2ULL))) != cast<uint32_t>(0ULL))) {
clear(sub(m->iwram,0,cast<int64_t>(32256ULL)));
}
if ((cast<uint32_t>((f & cast<uint32_t>(4ULL))) != cast<uint32_t>(0ULL))) {
clear(m->pal);if(rrcapture::trace.active)for(uint32_t i=0;i<m->pal.n;i++)rrGBAMemWrite(m,m->pal,i);
}
if ((cast<uint32_t>((f & cast<uint32_t>(8ULL))) != cast<uint32_t>(0ULL))) {
clear(m->vram);if(rrcapture::trace.active)for(uint32_t i=0;i<m->vram.n;i++)rrGBAMemWrite(m,m->vram,i);
}
if ((cast<uint32_t>((f & cast<uint32_t>(16ULL))) != cast<uint32_t>(0ULL))) {
clear(m->oam);if(rrcapture::trace.active)for(uint32_t i=0;i<m->oam.n;i++)rrGBAMemWrite(m,m->oam,i);
}
break;}
case cast<uint32_t>(2ULL):{
auto tmp1 = std::make_tuple(true,true,cast<uint16_t>(0ULL));
m->waiting = std::get<0>(tmp1);
m->waitAny = std::get<1>(tmp1);
m->waitMask = std::get<2>(tmp1);
break;}
case cast<uint32_t>(4ULL):{
gbamachine_Machine_intrWait(m,b,(c->R[cast<int64_t>(0ULL)] != cast<uint32_t>(0ULL)),cast<uint16_t>(c->R[cast<int64_t>(1ULL)]));
break;}
case cast<uint32_t>(5ULL):{
gbamachine_Machine_intrWait(m,b,true,cast<uint16_t>(1ULL));
break;}
case cast<uint32_t>(6ULL):{
auto tmp2 = std::make_tuple(cast<int32_t>(c->R[cast<int64_t>(0ULL)]),cast<int32_t>(c->R[cast<int64_t>(1ULL)]));
int32_t num = std::get<0>(tmp2);
int32_t den = std::get<1>(tmp2);
if ((den == cast<int32_t>(0ULL))) {
gbamachine_Machine_note(m,std::string("BIOS Div by zero at PC 0x%08X",29),c->R[cast<int64_t>(15ULL)]);
int32_t q = cast<int32_t>(1ULL);
if ((num < cast<int32_t>(0ULL))) {
q = cast<int32_t>(-1ULL);
}
auto tmp3 = std::make_tuple(cast<uint32_t>(q),cast<uint32_t>(num));
c->R[cast<int64_t>(0ULL)] = std::get<0>(tmp3);
c->R[cast<int64_t>(1ULL)] = std::get<1>(tmp3);
c->R[cast<int64_t>(3ULL)] = cast<uint32_t>(1ULL);
break;
}
int32_t q = divi<int32_t>(num,den);
auto tmp4 = std::make_tuple(cast<uint32_t>(q),cast<uint32_t>(modi<int32_t>(num,den)));
c->R[cast<int64_t>(0ULL)] = std::get<0>(tmp4);
c->R[cast<int64_t>(1ULL)] = std::get<1>(tmp4);
if ((q < cast<int32_t>(0ULL))) {
q = cast<int32_t>(-q);
}
c->R[cast<int64_t>(3ULL)] = cast<uint32_t>(q);
break;}
case cast<uint32_t>(7ULL):{
auto tmp5 = std::make_tuple(c->R[cast<int64_t>(1ULL)],c->R[cast<int64_t>(0ULL)]);
c->R[cast<int64_t>(0ULL)] = std::get<0>(tmp5);
c->R[cast<int64_t>(1ULL)] = std::get<1>(tmp5);
return gbamachine_biosSWI(m)(c,cast<uint32_t>(393216ULL));
break;}
case cast<uint32_t>(8ULL):{
c->R[cast<int64_t>(0ULL)] = cast<uint32_t>(go_math_Sqrt(cast<double>(c->R[cast<int64_t>(0ULL)])));
break;}
case cast<uint32_t>(9ULL):{
c->R[cast<int64_t>(0ULL)] = cast<uint32_t>(cast<uint16_t>(cast<int16_t>(go_math_Round(((go_math_Atan((cast<double>(cast<int16_t>(c->R[cast<int64_t>(0ULL)])) / cast<double>(16384ULL))) * cast<double>(32768ULL)) / go_math_Pi)))));
break;}
case cast<uint32_t>(10ULL):{
double a = go_math_Atan2(cast<double>(cast<int16_t>(c->R[cast<int64_t>(1ULL)])),cast<double>(cast<int16_t>(c->R[cast<int64_t>(0ULL)])));
c->R[cast<int64_t>(0ULL)] = cast<uint32_t>(cast<uint16_t>(go_math_Round(((a * cast<double>(32768ULL)) / go_math_Pi))));
break;}
case cast<uint32_t>(11ULL):{
uint32_t cnt = cast<uint32_t>((c->R[cast<int64_t>(2ULL)] & cast<uint32_t>(2097151ULL)));
bool fill = (cast<uint32_t>((c->R[cast<int64_t>(2ULL)] & cast<uint32_t>(16777216ULL))) != cast<uint32_t>(0ULL));
if ((cast<uint32_t>((c->R[cast<int64_t>(2ULL)] & cast<uint32_t>(67108864ULL))) != cast<uint32_t>(0ULL))) {
auto tmp6 = std::make_tuple((c->R[cast<int64_t>(0ULL)] & ~(cast<uint32_t>(3ULL))),(c->R[cast<int64_t>(1ULL)] & ~(cast<uint32_t>(3ULL))));
uint32_t src = std::get<0>(tmp6);
uint32_t dst = std::get<1>(tmp6);
uint32_t v = gbamachine_bus_Read32(b,src);
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < cnt);i++){
if ((!fill)) {
v = gbamachine_bus_Read32(b,cast<uint32_t>((src + cast<uint32_t>((i * cast<uint32_t>(4ULL))))));
}
gbamachine_bus_Write32(b,cast<uint32_t>((dst + cast<uint32_t>((i * cast<uint32_t>(4ULL))))),v);
}
}}
else {
auto tmp7 = std::make_tuple((c->R[cast<int64_t>(0ULL)] & ~(cast<uint32_t>(1ULL))),(c->R[cast<int64_t>(1ULL)] & ~(cast<uint32_t>(1ULL))));
uint32_t src = std::get<0>(tmp7);
uint32_t dst = std::get<1>(tmp7);
uint16_t v = gbamachine_bus_Read16(b,src);
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < cnt);i++){
if ((!fill)) {
v = gbamachine_bus_Read16(b,cast<uint32_t>((src + cast<uint32_t>((i * cast<uint32_t>(2ULL))))));
}
gbamachine_bus_Write16(b,cast<uint32_t>((dst + cast<uint32_t>((i * cast<uint32_t>(2ULL))))),v);
}
}}
break;}
case cast<uint32_t>(12ULL):{
uint32_t cnt = ((cast<uint32_t>((cast<uint32_t>((c->R[cast<int64_t>(2ULL)] & cast<uint32_t>(2097151ULL))) + cast<uint32_t>(7ULL)))) & ~(cast<uint32_t>(7ULL)));
bool fill = (cast<uint32_t>((c->R[cast<int64_t>(2ULL)] & cast<uint32_t>(16777216ULL))) != cast<uint32_t>(0ULL));
auto tmp8 = std::make_tuple((c->R[cast<int64_t>(0ULL)] & ~(cast<uint32_t>(3ULL))),(c->R[cast<int64_t>(1ULL)] & ~(cast<uint32_t>(3ULL))));
uint32_t src = std::get<0>(tmp8);
uint32_t dst = std::get<1>(tmp8);
uint32_t v = gbamachine_bus_Read32(b,src);
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < cnt);i++){
if ((!fill)) {
v = gbamachine_bus_Read32(b,cast<uint32_t>((src + cast<uint32_t>((i * cast<uint32_t>(4ULL))))));
}
gbamachine_bus_Write32(b,cast<uint32_t>((dst + cast<uint32_t>((i * cast<uint32_t>(4ULL))))),v);
}
}break;}
case cast<uint32_t>(13ULL):{
c->R[cast<int64_t>(0ULL)] = cast<uint32_t>(3131971711ULL);
break;}
case cast<uint32_t>(14ULL):{
gbamachine_Machine_bgAffineSet(m,b,c->R[cast<int64_t>(0ULL)],c->R[cast<int64_t>(1ULL)],c->R[cast<int64_t>(2ULL)]);
break;}
case cast<uint32_t>(15ULL):{
gbamachine_Machine_objAffineSet(m,b,c->R[cast<int64_t>(0ULL)],c->R[cast<int64_t>(1ULL)],c->R[cast<int64_t>(2ULL)],c->R[cast<int64_t>(3ULL)]);
break;}
case cast<uint32_t>(16ULL):{
gbamachine_Machine_bitUnPack(m,b,c->R[cast<int64_t>(0ULL)],c->R[cast<int64_t>(1ULL)],c->R[cast<int64_t>(2ULL)]);
break;}
case cast<uint32_t>(17ULL):case cast<uint32_t>(18ULL):{
gbamachine_Machine_noteDecompress(m,b,n,c->R[cast<int64_t>(0ULL)],c->R[cast<int64_t>(1ULL)],c->R[cast<int64_t>(15ULL)]);
gbamachine_Machine_lz77(m,b,c->R[cast<int64_t>(0ULL)],c->R[cast<int64_t>(1ULL)],(n == cast<uint32_t>(18ULL)));
break;}
case cast<uint32_t>(19ULL):{
gbamachine_Machine_noteDecompress(m,b,n,c->R[cast<int64_t>(0ULL)],c->R[cast<int64_t>(1ULL)],c->R[cast<int64_t>(15ULL)]);
gbamachine_Machine_huffman(m,b,c->R[cast<int64_t>(0ULL)],c->R[cast<int64_t>(1ULL)]);
break;}
case cast<uint32_t>(20ULL):case cast<uint32_t>(21ULL):{
gbamachine_Machine_noteDecompress(m,b,n,c->R[cast<int64_t>(0ULL)],c->R[cast<int64_t>(1ULL)],c->R[cast<int64_t>(15ULL)]);
gbamachine_Machine_rle(m,b,c->R[cast<int64_t>(0ULL)],c->R[cast<int64_t>(1ULL)],(n == cast<uint32_t>(21ULL)));
break;}
case cast<uint32_t>(25ULL):{
m->io[cast<uint32_t>(136ULL)] = cast<uint16_t>((shl<uint16_t>(cast<uint16_t>(cast<uint32_t>((c->R[cast<int64_t>(0ULL)] & cast<uint32_t>(1ULL)))),cast<int64_t>(9ULL)) | cast<uint16_t>(256ULL)));
break;}
case cast<uint32_t>(31ULL):{
uint32_t base = gbamachine_bus_Read32(b,cast<uint32_t>((c->R[cast<int64_t>(0ULL)] + cast<uint32_t>(4ULL))));
c->R[cast<int64_t>(0ULL)] = cast<uint32_t>((cast<double>(base) / go_math_Exp2(((cast<double>(cast<uint32_t>((cast<uint32_t>(180ULL) - c->R[cast<int64_t>(1ULL)]))) / cast<double>(12ULL)) - (cast<double>(c->R[cast<int64_t>(2ULL)]) / cast<double>(3072ULL))))));
break;}
case cast<uint32_t>(40ULL):{
break;}
case cast<uint32_t>(41ULL):{
break;}
default:{
arm_CPU_Halt(c,std::string("unimplemented BIOS SWI 0x%02X (r0=0x%08X r1=0x%08X r2=0x%08X) at 0x%08X",71),n,c->R[cast<int64_t>(0ULL)],c->R[cast<int64_t>(1ULL)],c->R[cast<int64_t>(2ULL)],c->R[cast<int64_t>(15ULL)]);
break;}
}}
return true;
}
;
}
}
// tools/platform/gba/gbamachine/bios.go:179:1
void gbamachine_Machine_noteDecompress(gbamachine_Machine* m,gbamachine_bus* b,uint32_t swi,uint32_t src,uint32_t dst,uint32_t pc){
{
if ((!m->OnDecompress)) {
return ;
}
m->OnDecompress(swi,src,dst,cast<int64_t>(shr<uint32_t>(gbamachine_bus_Read32(b,src),cast<int64_t>(8ULL))),pc,m->cpu->R[cast<int64_t>(14ULL)]);
}
}
// tools/platform/gba/gbamachine/bios.go:188:1
void gbamachine_Machine_intrWait(gbamachine_Machine* m,gbamachine_bus* b,bool discard,uint16_t mask){
{
m->ime = true;
uint16_t flags = cast<uint16_t>(gbamachine_bus_r32(b,cast<uint32_t>(50364408ULL)));
if (((!discard) && (cast<uint16_t>((flags & mask)) != cast<uint16_t>(0ULL)))) {
gbamachine_bus_w32(b,cast<uint32_t>(50364408ULL),cast<uint32_t>((flags & ~(mask))));
return ;
}
if (discard) {
gbamachine_bus_w32(b,cast<uint32_t>(50364408ULL),cast<uint32_t>((flags & ~(mask))));
}
auto tmp9 = std::make_tuple(true,false,mask);
m->waiting = std::get<0>(tmp9);
m->waitAny = std::get<1>(tmp9);
m->waitMask = std::get<2>(tmp9);
}
}
// tools/platform/gba/gbamachine/bios.go:205:1
void gbamachine_Machine_bgAffineSet(gbamachine_Machine* m,gbamachine_bus* b,uint32_t src,uint32_t dst,uint32_t num){
{
{;for (;(num > cast<uint32_t>(0ULL));num--){
double ox = (cast<double>(cast<int32_t>(gbamachine_bus_Read32(b,src))) / cast<double>(256ULL));
double oy = (cast<double>(cast<int32_t>(gbamachine_bus_Read32(b,cast<uint32_t>((src + cast<uint32_t>(4ULL)))))) / cast<double>(256ULL));
double dx = cast<double>(cast<int16_t>(gbamachine_bus_Read16(b,cast<uint32_t>((src + cast<uint32_t>(8ULL))))));
double dy = cast<double>(cast<int16_t>(gbamachine_bus_Read16(b,cast<uint32_t>((src + cast<uint32_t>(10ULL))))));
double sx = (cast<double>(cast<int16_t>(gbamachine_bus_Read16(b,cast<uint32_t>((src + cast<uint32_t>(12ULL)))))) / cast<double>(256ULL));
double sy = (cast<double>(cast<int16_t>(gbamachine_bus_Read16(b,cast<uint32_t>((src + cast<uint32_t>(14ULL)))))) / cast<double>(256ULL));
double theta = ((cast<double>(shr<uint16_t>(gbamachine_bus_Read16(b,cast<uint32_t>((src + cast<uint32_t>(16ULL)))),cast<int64_t>(8ULL))) / cast<double>(128ULL)) * go_math_Pi);
src += cast<uint32_t>(20ULL);
auto tmp10 = std::make_tuple(go_math_Sin(theta),go_math_Cos(theta));
double sin = std::get<0>(tmp10);
double cos = std::get<1>(tmp10);
auto tmp11 = std::make_tuple((sx * cos),(cast<double>(-sx) * sin));
double pa = std::get<0>(tmp11);
double pb = std::get<1>(tmp11);
auto tmp12 = std::make_tuple((sy * sin),(sy * cos));
double pc = std::get<0>(tmp12);
double pd = std::get<1>(tmp12);
gbamachine_bus_w16(b,dst,cast<uint16_t>(cast<int16_t>(go_math_Round((pa * cast<double>(256ULL))))));
gbamachine_bus_w16(b,cast<uint32_t>((dst + cast<uint32_t>(2ULL))),cast<uint16_t>(cast<int16_t>(go_math_Round((pb * cast<double>(256ULL))))));
gbamachine_bus_w16(b,cast<uint32_t>((dst + cast<uint32_t>(4ULL))),cast<uint16_t>(cast<int16_t>(go_math_Round((pc * cast<double>(256ULL))))));
gbamachine_bus_w16(b,cast<uint32_t>((dst + cast<uint32_t>(6ULL))),cast<uint16_t>(cast<int16_t>(go_math_Round((pd * cast<double>(256ULL))))));
double x0 = (ox - (((pa * dx) + (pb * dy))));
double y0 = (oy - (((pc * dx) + (pd * dy))));
gbamachine_bus_w32(b,cast<uint32_t>((dst + cast<uint32_t>(8ULL))),cast<uint32_t>(cast<int32_t>(go_math_Round((x0 * cast<double>(256ULL))))));
gbamachine_bus_w32(b,cast<uint32_t>((dst + cast<uint32_t>(12ULL))),cast<uint32_t>(cast<int32_t>(go_math_Round((y0 * cast<double>(256ULL))))));
dst += cast<uint32_t>(16ULL);
}
}}
}
// tools/platform/gba/gbamachine/bios.go:232:1
void gbamachine_Machine_objAffineSet(gbamachine_Machine* m,gbamachine_bus* b,uint32_t src,uint32_t dst,uint32_t num,uint32_t stride){
{
{;for (;(num > cast<uint32_t>(0ULL));num--){
double sx = (cast<double>(cast<int16_t>(gbamachine_bus_Read16(b,src))) / cast<double>(256ULL));
double sy = (cast<double>(cast<int16_t>(gbamachine_bus_Read16(b,cast<uint32_t>((src + cast<uint32_t>(2ULL)))))) / cast<double>(256ULL));
double theta = ((cast<double>(shr<uint16_t>(gbamachine_bus_Read16(b,cast<uint32_t>((src + cast<uint32_t>(4ULL)))),cast<int64_t>(8ULL))) / cast<double>(128ULL)) * go_math_Pi);
src += cast<uint32_t>(8ULL);
auto tmp13 = std::make_tuple(go_math_Sin(theta),go_math_Cos(theta));
double sin = std::get<0>(tmp13);
double cos = std::get<1>(tmp13);
auto put = [&](uint32_t off,double v)->void{
gbamachine_bus_w16(b,cast<uint32_t>((dst + cast<uint32_t>((off * stride)))),cast<uint16_t>(cast<int16_t>(go_math_Round((v * cast<double>(256ULL))))));
}
;
put(cast<uint32_t>(0ULL),(sx * cos));
put(cast<uint32_t>(1ULL),(cast<double>(-sx) * sin));
put(cast<uint32_t>(2ULL),(sy * sin));
put(cast<uint32_t>(3ULL),(sy * cos));
dst += cast<uint32_t>((cast<uint32_t>(4ULL) * stride));
}
}}
}
// tools/platform/gba/gbamachine/bios.go:254:1
void gbamachine_Machine_bitUnPack(gbamachine_Machine* m,gbamachine_bus* b,uint32_t src,uint32_t dst,uint32_t params){
{
uint32_t srcLen = cast<uint32_t>(gbamachine_bus_Read16(b,params));
int64_t srcW = cast<int64_t>(gbamachine_bus_Read(b,cast<uint32_t>((params + cast<uint32_t>(2ULL)))));
int64_t dstW = cast<int64_t>(gbamachine_bus_Read(b,cast<uint32_t>((params + cast<uint32_t>(3ULL)))));
uint32_t ofs = gbamachine_bus_Read32(b,cast<uint32_t>((params + cast<uint32_t>(4ULL))));
uint32_t bias = cast<uint32_t>((ofs & cast<uint32_t>(2147483647ULL)));
bool zeroToo = (cast<uint32_t>((ofs & cast<uint32_t>(2147483648ULL))) != cast<uint32_t>(0ULL));
uint32_t out={};
uint32_t outBits={};
auto flush = [&](int64_t bits)->void{
outBits += cast<uint32_t>(bits);
if ((outBits == cast<uint32_t>(32ULL))) {
gbamachine_bus_Write32(b,dst,out);
dst += cast<uint32_t>(4ULL);
auto tmp14 = std::make_tuple(cast<uint32_t>(0ULL),cast<uint32_t>(0ULL));
out = std::get<0>(tmp14);
outBits = std::get<1>(tmp14);
}
}
;
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < srcLen);i++){
uint32_t v = cast<uint32_t>(gbamachine_bus_Read(b,cast<uint32_t>((src + i))));
{int64_t bit = cast<int64_t>(0ULL);for (;(bit < cast<int64_t>(8ULL));bit += srcW){
uint32_t unit = cast<uint32_t>((shr<uint32_t>(v,cast<uint64_t>(bit)) & (cast<uint32_t>((shl<uint32_t>(cast<uint32_t>(1ULL),cast<uint64_t>(srcW)) - cast<uint32_t>(1ULL))))));
if (((unit != cast<uint32_t>(0ULL)) || zeroToo)) {
unit += bias;
}
out |= shl<uint32_t>(unit,outBits);
flush(dstW);
}
}}
}if ((outBits > cast<uint32_t>(0ULL))) {
gbamachine_bus_Write32(b,dst,out);
}
}
}
// tools/platform/gba/gbamachine/bios.go:288:1
std::tuple<uint32_t,bool> gbamachine_Machine_lzHeader(gbamachine_Machine* m,gbamachine_bus* b,uint32_t src,uint32_t wantType,std::string name){
uint32_t size{};
bool ok{};
{
uint32_t h = gbamachine_bus_Read32(b,src);
if ((cast<uint32_t>((shr<uint32_t>(h,cast<int64_t>(4ULL)) & cast<uint32_t>(15ULL))) != wantType)) {
arm_CPU_Halt(m->cpu,std::string("BIOS %s: header 0x%08X at 0x%08X is not type %d",47),name,h,src,wantType);
return {cast<uint32_t>(0ULL),false};
}
return {shr<uint32_t>(h,cast<int64_t>(8ULL)),true};
}
}
// tools/platform/gba/gbamachine/bios.go:299:1
void gbamachine_Machine_lz77(gbamachine_Machine* m,gbamachine_bus* b,uint32_t src,uint32_t dst,bool vram){
{
auto tmp15 = gbamachine_Machine_lzHeader(m,b,src,cast<uint32_t>(1ULL),std::string("LZ77UnComp",10));
uint32_t size = std::get<0>(tmp15);
bool ok = std::get<1>(tmp15);
if ((!ok)) {
return ;
}
src += cast<uint32_t>(4ULL);
uint16_t pend={};
uint32_t pendBytes={};
auto put = [&](uint8_t v)->void{
if ((!vram)) {
gbamachine_bus_Write(b,dst,v);
dst++;
return ;
}
pend |= shl<uint16_t>(cast<uint16_t>(v),(cast<uint32_t>((cast<uint32_t>(8ULL) * pendBytes))));
{
pendBytes++;
if ((pendBytes == cast<uint32_t>(2ULL))) {
gbamachine_bus_Write16(b,dst,pend);
dst += cast<uint32_t>(2ULL);
auto tmp16 = std::make_tuple(cast<uint16_t>(0ULL),cast<uint32_t>(0ULL));
pend = std::get<0>(tmp16);
pendBytes = std::get<1>(tmp16);
}
}
}
;
Slice<uint8_t> written = Slice<uint8_t>::make(cast<int64_t>(0ULL),size);
auto emit = [&](uint8_t v)->void{
written = append(written,Slice<uint8_t>{v});
put(v);
}
;
{;for (;(cast<uint32_t>(len(written)) < size);){
uint8_t flags = gbamachine_bus_Read(b,src);
src++;
{int64_t bit = cast<int64_t>(7ULL);for (;((bit >= cast<int64_t>(0ULL)) && (cast<uint32_t>(len(written)) < size));bit--){
if ((cast<uint8_t>((shr<uint8_t>(flags,cast<uint64_t>(bit)) & cast<uint8_t>(1ULL))) == cast<uint8_t>(0ULL))) {
emit(gbamachine_bus_Read(b,src));
src++;
continue;
}
auto tmp17 = std::make_tuple(gbamachine_bus_Read(b,src),gbamachine_bus_Read(b,cast<uint32_t>((src + cast<uint32_t>(1ULL)))));
uint8_t b1 = std::get<0>(tmp17);
uint8_t b2 = std::get<1>(tmp17);
src += cast<uint32_t>(2ULL);
int64_t length = cast<int64_t>((cast<int64_t>(shr<uint8_t>(b1,cast<int64_t>(4ULL))) + cast<int64_t>(3ULL)));
int64_t disp = cast<int64_t>((cast<int64_t>((shl<int64_t>(cast<int64_t>(cast<uint8_t>((b1 & cast<uint8_t>(15ULL)))),cast<int64_t>(8ULL)) | cast<int64_t>(b2))) + cast<int64_t>(1ULL)));
{int64_t i = cast<int64_t>(0ULL);for (;((i < length) && (cast<uint32_t>(len(written)) < size));i++){
emit(written[cast<int64_t>((len(written) - disp))]);
}
}}
}}
}if ((vram && (pendBytes == cast<uint32_t>(1ULL)))) {
gbamachine_bus_Write16(b,dst,pend);
}
}
}
// tools/platform/gba/gbamachine/bios.go:351:1
void gbamachine_Machine_rle(gbamachine_Machine* m,gbamachine_bus* b,uint32_t src,uint32_t dst,bool vram){
{
auto tmp18 = gbamachine_Machine_lzHeader(m,b,src,cast<uint32_t>(3ULL),std::string("RLUnComp",8));
uint32_t size = std::get<0>(tmp18);
bool ok = std::get<1>(tmp18);
if ((!ok)) {
return ;
}
src += cast<uint32_t>(4ULL);
uint16_t pend={};
uint32_t pendBytes={};
uint32_t done={};
auto put = [&](uint8_t v)->void{
done++;
if ((!vram)) {
gbamachine_bus_Write(b,dst,v);
dst++;
return ;
}
pend |= shl<uint16_t>(cast<uint16_t>(v),(cast<uint32_t>((cast<uint32_t>(8ULL) * pendBytes))));
{
pendBytes++;
if ((pendBytes == cast<uint32_t>(2ULL))) {
gbamachine_bus_Write16(b,dst,pend);
dst += cast<uint32_t>(2ULL);
auto tmp19 = std::make_tuple(cast<uint16_t>(0ULL),cast<uint32_t>(0ULL));
pend = std::get<0>(tmp19);
pendBytes = std::get<1>(tmp19);
}
}
}
;
{;for (;(done < size);){
uint8_t f = gbamachine_bus_Read(b,src);
src++;
if ((cast<uint8_t>((f & cast<uint8_t>(128ULL))) != cast<uint8_t>(0ULL))) {
uint8_t v = gbamachine_bus_Read(b,src);
src++;
{int64_t i = cast<int64_t>(0ULL);for (;((i < cast<int64_t>((cast<int64_t>(cast<uint8_t>((f & cast<uint8_t>(127ULL)))) + cast<int64_t>(3ULL)))) && (done < size));i++){
put(v);
}
}}
else {
{int64_t i = cast<int64_t>(0ULL);for (;((i < cast<int64_t>((cast<int64_t>(cast<uint8_t>((f & cast<uint8_t>(127ULL)))) + cast<int64_t>(1ULL)))) && (done < size));i++){
put(gbamachine_bus_Read(b,src));
src++;
}
}}
}
}if ((vram && (pendBytes == cast<uint32_t>(1ULL)))) {
gbamachine_bus_Write16(b,dst,pend);
}
}
}
// tools/platform/gba/gbamachine/bios.go:396:1
void gbamachine_Machine_huffman(gbamachine_Machine* m,gbamachine_bus* b,uint32_t src,uint32_t dst){
{
uint32_t h = gbamachine_bus_Read32(b,src);
int64_t dataBits = cast<int64_t>(cast<uint32_t>((h & cast<uint32_t>(15ULL))));
uint32_t size = shr<uint32_t>(h,cast<int64_t>(8ULL));
uint32_t treeSize = cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(gbamachine_bus_Read(b,cast<uint32_t>((src + cast<uint32_t>(4ULL))))) * cast<uint32_t>(2ULL))) + cast<uint32_t>(1ULL)));
uint32_t treeRoot = cast<uint32_t>((src + cast<uint32_t>(5ULL)));
uint32_t bits = cast<uint32_t>((treeRoot + treeSize));
uint32_t out={};
uint32_t outBits={};
uint32_t done={};
uint32_t node = treeRoot;
uint8_t nodeVal = gbamachine_bus_Read(b,node);
{;for (;(done < size);){
uint32_t w = gbamachine_bus_Read32(b,bits);
bits += cast<uint32_t>(4ULL);
{int64_t i = cast<int64_t>(31ULL);for (;(i >= cast<int64_t>(0ULL));i--){
uint32_t bit = cast<uint32_t>((shr<uint32_t>(w,cast<uint64_t>(i)) & cast<uint32_t>(1ULL)));
uint32_t child = cast<uint32_t>(((cast<uint32_t>((cast<uint32_t>(((node & ~(cast<uint32_t>(1ULL))) + cast<uint32_t>((cast<uint32_t>(cast<uint8_t>((nodeVal & cast<uint8_t>(63ULL)))) * cast<uint32_t>(2ULL))))) + cast<uint32_t>(2ULL)))) + bit));
bool leaf = (cast<uint8_t>((nodeVal & (shr<uint8_t>(cast<uint8_t>(128ULL),bit)))) != cast<uint8_t>(0ULL));
nodeVal = gbamachine_bus_Read(b,child);
node = child;
if (leaf) {
out |= shl<uint32_t>(cast<uint32_t>(nodeVal),outBits);
outBits += cast<uint32_t>(dataBits);
node = treeRoot;
nodeVal = gbamachine_bus_Read(b,node);
if ((outBits == cast<uint32_t>(32ULL))) {
gbamachine_bus_Write32(b,dst,out);
dst += cast<uint32_t>(4ULL);
auto tmp20 = std::make_tuple(cast<uint32_t>(0ULL),cast<uint32_t>(0ULL));
out = std::get<0>(tmp20);
outBits = std::get<1>(tmp20);
{
done += cast<uint32_t>(4ULL);
if ((done >= size)) {
break;
}
}
}
}
}
}}
}}
}
// tools/platform/gba/gbamachine/bus.go:22:1
std::tuple<Slice<uint8_t>,uint32_t,bool> gbamachine_bus_region(gbamachine_bus* b,uint32_t a){
Slice<uint8_t> mem{};
uint32_t off{};
bool ok{};
{
gbamachine_Machine* m = b->m;
{
switch(shr<uint32_t>(a,cast<int64_t>(24ULL))){
case cast<uint32_t>(2ULL):{
return {m->ewram,cast<uint32_t>((a & cast<uint32_t>(262143ULL))),true};
break;}
case cast<uint32_t>(3ULL):{
return {m->iwram,cast<uint32_t>((a & cast<uint32_t>(32767ULL))),true};
break;}
case cast<uint32_t>(5ULL):{
return {m->pal,cast<uint32_t>((a & cast<uint32_t>(1023ULL))),true};
break;}
case cast<uint32_t>(6ULL):{
off = cast<uint32_t>((a & cast<uint32_t>(131071ULL)));
if ((off >= cast<uint32_t>(98304ULL))) {
off -= cast<uint32_t>(32768ULL);
}
return {m->vram,off,true};
break;}
case cast<uint32_t>(7ULL):{
return {m->oam,cast<uint32_t>((a & cast<uint32_t>(1023ULL))),true};
break;}
}}
return {{},cast<uint32_t>(0ULL),false};
}
}
// tools/platform/gba/gbamachine/bus.go:45:1
std::tuple<uint8_t,bool> gbamachine_bus_romByte(gbamachine_bus* b,uint32_t a){
{
uint32_t off = cast<uint32_t>((a & cast<uint32_t>(33554431ULL)));
if ((cast<int64_t>(off) < len(b->m->rom))) {
return {b->m->rom[off],true};
}
return {cast<uint8_t>(0ULL),false};
}
}
// tools/platform/gba/gbamachine/bus.go:54:1
uint8_t gbamachine_bus_Read(gbamachine_bus* b,uint32_t a){
{
gbamachine_Machine* m = b->m;
if (bool(m->OnRead)) {
auto tmp21=defer([&](){[&]()->void{
m->OnRead(a,cast<uint8_t>(0ULL),m->cpu->R[cast<int64_t>(15ULL)]);
}
();});
}
{
switch(shr<uint32_t>(a,cast<int64_t>(24ULL))){
case cast<uint32_t>(0ULL):case cast<uint32_t>(1ULL):{
if ((a < cast<uint32_t>(16384ULL))) {
gbamachine_Machine_note(m,std::string("read of BIOS region 0x%08X (no BIOS image; returns 0)",53),a);
return cast<uint8_t>(0ULL);
}
break;}
case cast<uint32_t>(4ULL):{
uint16_t v = gbamachine_Machine_ioRead16(m,(a & ~(cast<uint32_t>(1ULL))));
return cast<uint8_t>(shr<uint16_t>(v,(cast<uint32_t>((cast<uint32_t>(8ULL) * (cast<uint32_t>((a & cast<uint32_t>(1ULL)))))))));
break;}
case cast<uint32_t>(8ULL):case cast<uint32_t>(9ULL):case cast<uint32_t>(10ULL):case cast<uint32_t>(11ULL):case cast<uint32_t>(12ULL):{
{
auto tmp22 = gbamachine_bus_romByte(b,a);
uint8_t v = std::get<0>(tmp22);
bool ok = std::get<1>(tmp22);
if (ok) {
return v;
}
}
return cast<uint8_t>(shr<uint32_t>((shr<uint32_t>(a,cast<int64_t>(1ULL))),(cast<uint32_t>((cast<uint32_t>(8ULL) * (cast<uint32_t>((a & cast<uint32_t>(1ULL)))))))));
break;}
case cast<uint32_t>(13ULL):{
if (m->eeprom.present) {
return cast<uint8_t>(gbamachine_eeprom_read(&(m->eeprom)));
}
{
auto tmp23 = gbamachine_bus_romByte(b,a);
uint8_t v = std::get<0>(tmp23);
bool ok = std::get<1>(tmp23);
if (ok) {
return v;
}
}
return cast<uint8_t>(shr<uint32_t>((shr<uint32_t>(a,cast<int64_t>(1ULL))),(cast<uint32_t>((cast<uint32_t>(8ULL) * (cast<uint32_t>((a & cast<uint32_t>(1ULL)))))))));
break;}
case cast<uint32_t>(14ULL):case cast<uint32_t>(15ULL):{
gbamachine_Machine_note(m,std::string("read of SRAM/Flash region 0x%08X (this cart is EEPROM; returns 0)",65),a);
return cast<uint8_t>(0ULL);
break;}
}}
{
auto tmp24 = gbamachine_bus_region(b,a);
Slice<uint8_t> mem = std::get<0>(tmp24);
uint32_t off = std::get<1>(tmp24);
bool ok = std::get<2>(tmp24);
if (ok) {
return mem[off];
}
}
gbamachine_Machine_note(m,std::string("read of unmapped address 0x%08X",31),a);
return cast<uint8_t>(0ULL);
}
}
// tools/platform/gba/gbamachine/bus.go:98:1
void gbamachine_bus_Write(gbamachine_bus* b,uint32_t a,uint8_t v){
{
gbamachine_Machine* m = b->m;
if (bool(m->OnWrite)) {
m->OnWrite(a,v,m->cpu->R[cast<int64_t>(15ULL)]);
}
{
switch(shr<uint32_t>(a,cast<int64_t>(24ULL))){
case cast<uint32_t>(0ULL):case cast<uint32_t>(1ULL):{
gbamachine_Machine_note(m,std::string("write to BIOS region 0x%08X (ignored)",37),a);
return ;
break;}
case cast<uint32_t>(4ULL):{
gbamachine_Machine_ioWrite8(m,a,v);
return ;
break;}
case cast<uint32_t>(5ULL):{
uint32_t off = (cast<uint32_t>((a & cast<uint32_t>(1023ULL))) & ~(cast<uint32_t>(1ULL)));
auto tmp25 = std::make_tuple(v,v);
m->pal[off] = std::get<0>(tmp25);
m->pal[cast<uint32_t>((off + cast<uint32_t>(1ULL)))] = std::get<1>(tmp25);
if(rrcapture::trace.active)rrGBAMemWrite(b->m,m->pal,off);
if(rrcapture::trace.active)rrGBAMemWrite(b->m,m->pal,cast<uint32_t>((off + cast<uint32_t>(1ULL))));
return ;
break;}
case cast<uint32_t>(6ULL):{
uint32_t off = cast<uint32_t>((a & cast<uint32_t>(131071ULL)));
if ((off >= cast<uint32_t>(98304ULL))) {
off -= cast<uint32_t>(32768ULL);
}
if ((off >= gbamachine_ppu_objVRAMBase(&(m->ppu),m))) {
return ;
}
off &= ~(cast<uint32_t>(1ULL));
auto tmp26 = std::make_tuple(v,v);
m->vram[off] = std::get<0>(tmp26);
m->vram[cast<uint32_t>((off + cast<uint32_t>(1ULL)))] = std::get<1>(tmp26);
if(rrcapture::trace.active)rrGBAMemWrite(b->m,m->vram,off);
if(rrcapture::trace.active)rrGBAMemWrite(b->m,m->vram,cast<uint32_t>((off + cast<uint32_t>(1ULL))));
return ;
break;}
case cast<uint32_t>(7ULL):{
return ;
break;}
case cast<uint32_t>(8ULL):case cast<uint32_t>(9ULL):case cast<uint32_t>(10ULL):case cast<uint32_t>(11ULL):case cast<uint32_t>(12ULL):{
gbamachine_Machine_note(m,std::string("byte write to ROM region 0x%08X (ignored)",41),a);
return ;
break;}
case cast<uint32_t>(13ULL):{
if (m->eeprom.present) {
gbamachine_eeprom_write(&(m->eeprom),cast<uint16_t>(v));
return ;
}
gbamachine_Machine_note(m,std::string("byte write to 0x0D region 0x%08X with no EEPROM (ignored)",57),a);
return ;
break;}
case cast<uint32_t>(14ULL):case cast<uint32_t>(15ULL):{
gbamachine_Machine_note(m,std::string("write to SRAM/Flash region 0x%08X (this cart is EEPROM; ignored)",64),a);
return ;
break;}
}}
{
auto tmp27 = gbamachine_bus_region(b,a);
Slice<uint8_t> mem = std::get<0>(tmp27);
uint32_t off = std::get<1>(tmp27);
bool ok = std::get<2>(tmp27);
if (ok) {
mem[off] = v;
if(rrcapture::trace.active)rrGBAMemWrite(b->m,mem,off);
return ;
}
}
gbamachine_Machine_note(m,std::string("write to unmapped address 0x%08X",32),a);
}
}
// tools/platform/gba/gbamachine/bus.go:150:1
uint16_t gbamachine_bus_Read16_reference(gbamachine_bus* b,uint32_t a){
{
gbamachine_Machine* m = b->m;
if (bool(m->OnRead)) {
m->OnRead(a,cast<uint8_t>(0ULL),m->cpu->R[cast<int64_t>(15ULL)]);
m->OnRead(cast<uint32_t>((a + cast<uint32_t>(1ULL))),cast<uint8_t>(0ULL),m->cpu->R[cast<int64_t>(15ULL)]);
}
{
switch(shr<uint32_t>(a,cast<int64_t>(24ULL))){
case cast<uint32_t>(4ULL):{
return gbamachine_Machine_ioRead16(m,a);
break;}
case cast<uint32_t>(13ULL):{
if (m->eeprom.present) {
return gbamachine_eeprom_read(&(m->eeprom));
}
break;}
}}
{
auto tmp28 = gbamachine_bus_region(b,a);
Slice<uint8_t> mem = std::get<0>(tmp28);
uint32_t off = std::get<1>(tmp28);
bool ok = std::get<2>(tmp28);
if (ok) {
return cast<uint16_t>((cast<uint16_t>(mem[off]) | shl<uint16_t>(cast<uint16_t>(mem[cast<uint32_t>((off + cast<uint32_t>(1ULL)))]),cast<int64_t>(8ULL))));
}
}
return cast<uint16_t>((cast<uint16_t>(gbamachine_bus_Read(b,a)) | shl<uint16_t>(cast<uint16_t>(gbamachine_bus_Read(b,cast<uint32_t>((a + cast<uint32_t>(1ULL))))),cast<int64_t>(8ULL))));
}
}
// tools/platform/gba/gbamachine/bus.go:170:1
void gbamachine_bus_Write16(gbamachine_bus* b,uint32_t a,uint16_t v){
{
gbamachine_Machine* m = b->m;
if (bool(m->OnWrite)) {
m->OnWrite(a,cast<uint8_t>(v),m->cpu->R[cast<int64_t>(15ULL)]);
m->OnWrite(cast<uint32_t>((a + cast<uint32_t>(1ULL))),cast<uint8_t>(shr<uint16_t>(v,cast<int64_t>(8ULL))),m->cpu->R[cast<int64_t>(15ULL)]);
}
{
switch(shr<uint32_t>(a,cast<int64_t>(24ULL))){
case cast<uint32_t>(4ULL):{
gbamachine_Machine_ioWrite16(m,a,v);
return ;
break;}
case cast<uint32_t>(13ULL):{
if (m->eeprom.present) {
gbamachine_eeprom_write(&(m->eeprom),v);
return ;
}
break;}
}}
{
auto tmp29 = gbamachine_bus_region(b,a);
Slice<uint8_t> mem = std::get<0>(tmp29);
uint32_t off = std::get<1>(tmp29);
bool ok = std::get<2>(tmp29);
if (ok) {
mem[off] = cast<uint8_t>(v);
if(rrcapture::trace.active)rrGBAMemWrite(b->m,mem,off);
mem[cast<uint32_t>((off + cast<uint32_t>(1ULL)))] = cast<uint8_t>(shr<uint16_t>(v,cast<int64_t>(8ULL)));
if(rrcapture::trace.active)rrGBAMemWrite(b->m,mem,cast<uint32_t>((off + cast<uint32_t>(1ULL))));
return ;
}
}
gbamachine_bus_Write(b,a,cast<uint8_t>(v));
gbamachine_bus_Write(b,cast<uint32_t>((a + cast<uint32_t>(1ULL))),cast<uint8_t>(shr<uint16_t>(v,cast<int64_t>(8ULL))));
}
}
// tools/platform/gba/gbamachine/bus.go:195:1
uint32_t gbamachine_bus_Read32(gbamachine_bus* b,uint32_t a){
{
return cast<uint32_t>((cast<uint32_t>(gbamachine_bus_Read16(b,a)) | shl<uint32_t>(cast<uint32_t>(gbamachine_bus_Read16(b,cast<uint32_t>((a + cast<uint32_t>(2ULL))))),cast<int64_t>(16ULL))));
}
}
// tools/platform/gba/gbamachine/bus.go:199:1
void gbamachine_bus_Write32(gbamachine_bus* b,uint32_t a,uint32_t v){
{
gbamachine_bus_Write16(b,a,cast<uint16_t>(v));
gbamachine_bus_Write16(b,cast<uint32_t>((a + cast<uint32_t>(2ULL))),cast<uint16_t>(shr<uint32_t>(v,cast<int64_t>(16ULL))));
}
}
// tools/platform/gba/gbamachine/bus.go:205:1
uint32_t gbamachine_bus_r32(gbamachine_bus* b,uint32_t a){
{
return gbamachine_bus_Read32(b,a);
}
}
// tools/platform/gba/gbamachine/bus.go:206:1
void gbamachine_bus_w32(gbamachine_bus* b,uint32_t a,uint32_t v){
{
gbamachine_bus_Write32(b,a,v);
}
}
// tools/platform/gba/gbamachine/bus.go:207:1
uint16_t gbamachine_bus_r16(gbamachine_bus* b,uint32_t a){
{
return gbamachine_bus_Read16(b,a);
}
}
// tools/platform/gba/gbamachine/bus.go:208:1
void gbamachine_bus_w16(gbamachine_bus* b,uint32_t a,uint16_t v){
{
gbamachine_bus_Write16(b,a,v);
}
}
// tools/platform/gba/gbamachine/dma.go:18:1
void gbamachine_Machine_dmaRegWrite(gbamachine_Machine* m,uint32_t reg,uint16_t v){
{
int64_t n = divi<int64_t>(cast<int64_t>(cast<uint32_t>((reg - cast<uint32_t>(176ULL)))),cast<int64_t>(12ULL));
gbamachine_dmaChan* d = (&m->dma[n]);
{
switch(modi<uint32_t>((cast<uint32_t>((reg - cast<uint32_t>(176ULL)))),cast<uint32_t>(12ULL))){
case cast<uint32_t>(0ULL):{
d->src = cast<uint32_t>((cast<uint32_t>((d->src & cast<uint32_t>(4294901760ULL))) | cast<uint32_t>(v)));
break;}
case cast<uint32_t>(2ULL):{
d->src = cast<uint32_t>((cast<uint32_t>((d->src & cast<uint32_t>(65535ULL))) | shl<uint32_t>(cast<uint32_t>(v),cast<int64_t>(16ULL))));
break;}
case cast<uint32_t>(4ULL):{
d->dst = cast<uint32_t>((cast<uint32_t>((d->dst & cast<uint32_t>(4294901760ULL))) | cast<uint32_t>(v)));
break;}
case cast<uint32_t>(6ULL):{
d->dst = cast<uint32_t>((cast<uint32_t>((d->dst & cast<uint32_t>(65535ULL))) | shl<uint32_t>(cast<uint32_t>(v),cast<int64_t>(16ULL))));
break;}
case cast<uint32_t>(8ULL):{
d->count = v;
break;}
case cast<uint32_t>(10ULL):{
uint16_t was = d->ctrl;
d->ctrl = v;
if (((cast<uint16_t>((v & cast<uint16_t>(32768ULL))) != cast<uint16_t>(0ULL)) && (cast<uint16_t>((was & cast<uint16_t>(32768ULL))) == cast<uint16_t>(0ULL)))) {
auto tmp30 = std::make_tuple(d->src,d->dst);
d->latchSrc = std::get<0>(tmp30);
d->latchDst = std::get<1>(tmp30);
if ((cast<uint16_t>((shr<uint16_t>(v,cast<int64_t>(12ULL)) & cast<uint16_t>(3ULL))) == cast<uint16_t>(0ULL))) {
gbamachine_Machine_dmaRun(m,n);
}
}
break;}
}}
}
}
// tools/platform/gba/gbamachine/dma.go:52:1
void gbamachine_Machine_dmaSoundRefill(gbamachine_Machine* m,int64_t n){
{
gbamachine_dmaChan* d = (&m->dma[n]);
if (((cast<uint16_t>((d->ctrl & cast<uint16_t>(32768ULL))) == cast<uint16_t>(0ULL)) || (cast<uint16_t>((shr<uint16_t>(d->ctrl,cast<int64_t>(12ULL)) & cast<uint16_t>(3ULL))) != cast<uint16_t>(3ULL)))) {
return ;
}
gbamachine_bus busStorage{m}; auto*b=&busStorage;
uint32_t dst = d->dst;
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < cast<uint32_t>(4ULL));i++){
gbamachine_bus_Write32(b,dst,gbamachine_bus_Read32(b,(d->latchSrc & ~(cast<uint32_t>(3ULL)))));
d->latchSrc += cast<uint32_t>(4ULL);
}
}if ((cast<uint16_t>((d->ctrl & cast<uint16_t>(16384ULL))) != cast<uint16_t>(0ULL))) {
gbamachine_Machine_raise(m,shl<uint16_t>(cast<uint16_t>(256ULL),cast<uint64_t>(n)));
}
}
}
// tools/platform/gba/gbamachine/dma.go:71:1
void gbamachine_Machine_dmaTrigger(gbamachine_Machine* m,uint16_t timing){
{
{auto&& tmp31 = m->dma;
for(int64_t tmp32=0;tmp32<len(tmp31);++tmp32){
auto n=tmp32;gbamachine_dmaChan* d = (&m->dma[n]);
if (((cast<uint16_t>((d->ctrl & cast<uint16_t>(32768ULL))) != cast<uint16_t>(0ULL)) && (cast<uint16_t>((shr<uint16_t>(d->ctrl,cast<int64_t>(12ULL)) & cast<uint16_t>(3ULL))) == timing))) {
gbamachine_Machine_dmaRun(m,n);
}
}}
}
}
// tools/platform/gba/gbamachine/dma.go:81:1
void gbamachine_Machine_dmaRun(gbamachine_Machine* m,int64_t n){
{rrprof::Scope timing(2,"DMA");
{
gbamachine_dmaChan* d = (&m->dma[n]);
gbamachine_bus busStorage{m}; auto*b=&busStorage;
uint32_t count = cast<uint32_t>(d->count);
uint32_t max = cast<uint32_t>(16384ULL);
if ((n == cast<int64_t>(3ULL))) {
max = cast<uint32_t>(65536ULL);
}
if ((count == cast<uint32_t>(0ULL))) {
count = max;
}
bool word = (cast<uint16_t>((d->ctrl & cast<uint16_t>(1024ULL))) != cast<uint16_t>(0ULL));
uint32_t unit = cast<uint32_t>(2ULL);
if (word) {
unit = cast<uint32_t>(4ULL);
}
auto step = [&](uint16_t mode)->int32_t{
{
switch(mode){
case cast<uint16_t>(1ULL):{
return cast<int32_t>(-cast<int32_t>(unit));
break;}
case cast<uint16_t>(2ULL):{
return cast<int32_t>(0ULL);
break;}
default:{
return cast<int32_t>(unit);
break;}
}}
}
;
int32_t sstep = step(cast<uint16_t>((shr<uint16_t>(d->ctrl,cast<int64_t>(7ULL)) & cast<uint16_t>(3ULL))));
uint16_t dmode = cast<uint16_t>((shr<uint16_t>(d->ctrl,cast<int64_t>(5ULL)) & cast<uint16_t>(3ULL)));
int32_t dstep = step(dmode);
auto tmp33 = std::make_tuple(d->latchSrc,d->latchDst);
uint32_t src = std::get<0>(tmp33);
uint32_t dst = std::get<1>(tmp33);
uint32_t startDst = dst;
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < count);i++){
if (word) {
gbamachine_bus_Write32(b,(dst & ~(cast<uint32_t>(3ULL))),gbamachine_bus_Read32(b,(src & ~(cast<uint32_t>(3ULL)))));
}
else {
gbamachine_bus_Write16(b,(dst & ~(cast<uint32_t>(1ULL))),gbamachine_bus_Read16(b,(src & ~(cast<uint32_t>(1ULL)))));
}
src = cast<uint32_t>(cast<int32_t>((cast<int32_t>(src) + sstep)));
dst = cast<uint32_t>(cast<int32_t>((cast<int32_t>(dst) + dstep)));
}
}d->latchSrc = src;
if ((dmode == cast<uint16_t>(3ULL))) {
d->latchDst = d->dst;
}
else {
d->latchDst = dst;
}
if ((m->eeprom.present && (shr<uint32_t>(startDst,cast<int64_t>(24ULL)) == cast<uint32_t>(13ULL)))) {
gbamachine_eeprom_endFrame(&(m->eeprom),m);
}
if ((cast<uint16_t>((d->ctrl & cast<uint16_t>(16384ULL))) != cast<uint16_t>(0ULL))) {
gbamachine_Machine_raise(m,shl<uint16_t>(cast<uint16_t>(256ULL),cast<uint64_t>(n)));
}
if (((cast<uint16_t>((d->ctrl & cast<uint16_t>(512ULL))) == cast<uint16_t>(0ULL)) || (cast<uint16_t>((shr<uint16_t>(d->ctrl,cast<int64_t>(12ULL)) & cast<uint16_t>(3ULL))) == cast<uint16_t>(0ULL)))) {
d->ctrl &= ~(cast<uint16_t>(32768ULL));
}
}
}
}
// tools/platform/gba/gbamachine/eeprom.go:33:1
void gbamachine_eeprom_init(gbamachine_eeprom* e){
{
e->present = true;
e->data = Slice<uint8_t>::make(cast<int64_t>(8192ULL));
{auto&& tmp34 = e->data;
for(int64_t tmp35=0;tmp35<len(tmp34);++tmp35){
auto i=tmp35;e->data[i] = cast<uint8_t>(255ULL);
}}
e->ready = true;
}
}
// tools/platform/gba/gbamachine/eeprom.go:43:1
void gbamachine_eeprom_write(gbamachine_eeprom* e,uint16_t v){
{
e->inBits = append(e->inBits,Slice<uint8_t>{cast<uint8_t>(cast<uint16_t>((v & cast<uint16_t>(1ULL))))});
}
}
// tools/platform/gba/gbamachine/eeprom.go:46:1
uint16_t gbamachine_eeprom_read(gbamachine_eeprom* e){
{
if ((len(e->outBits) > cast<int64_t>(0ULL))) {
uint16_t v = cast<uint16_t>(e->outBits[cast<int64_t>(0ULL)]);
e->outBits = sub(e->outBits,cast<int64_t>(1ULL),len(e->outBits));
return v;
}
if (e->ready) {
return cast<uint16_t>(1ULL);
}
return cast<uint16_t>(0ULL);
}
}
// tools/platform/gba/gbamachine/eeprom.go:60:1
void gbamachine_eeprom_endFrame(gbamachine_eeprom* e,gbamachine_Machine* m){
{
Slice<uint8_t> bits = e->inBits;
e->inBits = sub(e->inBits,0,cast<int64_t>(0ULL));
if ((len(bits) < cast<int64_t>(3ULL))) {
return ;
}
int64_t addrBits={};
{
switch(len(bits)){
case cast<int64_t>(9ULL):case cast<int64_t>(73ULL):{
addrBits = cast<int64_t>(6ULL);
break;}
case cast<int64_t>(17ULL):case cast<int64_t>(81ULL):{
addrBits = cast<int64_t>(14ULL);
break;}
default:{
gbamachine_Machine_note(m,std::string("EEPROM request of %d bits is not a known frame length (ignored)",63),len(bits));
return ;
break;}
}}
gbamachine_eeprom_size(e,(addrBits == cast<int64_t>(14ULL)));
int64_t addr = cast<int64_t>(0ULL);
{int64_t i = cast<int64_t>(0ULL);for (;(i < addrBits);i++){
addr = cast<int64_t>((shl<int64_t>(addr,cast<int64_t>(1ULL)) | cast<int64_t>(bits[cast<int64_t>((cast<int64_t>(2ULL) + i))])));
}
}addr &= cast<int64_t>((divi<int64_t>(len(e->data),cast<int64_t>(8ULL)) - cast<int64_t>(1ULL)));
{
if (((bits[cast<int64_t>(0ULL)] == cast<uint8_t>(1ULL)) && (bits[cast<int64_t>(1ULL)] == cast<uint8_t>(1ULL)))){
e->outBits = sub(e->outBits,0,cast<int64_t>(0ULL));
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(4ULL));i++){
e->outBits = append(e->outBits,Slice<uint8_t>{cast<uint8_t>(0ULL)});
}
}{auto&& tmp37 = sub(e->data,cast<int64_t>((addr * cast<int64_t>(8ULL))),cast<int64_t>((cast<int64_t>((addr * cast<int64_t>(8ULL))) + cast<int64_t>(8ULL))));
for(int64_t tmp38=0;tmp38<len(tmp37);++tmp38){
auto b=tmp37[tmp38];{int64_t bit = cast<int64_t>(7ULL);for (;(bit >= cast<int64_t>(0ULL));bit--){
e->outBits = append(e->outBits,Slice<uint8_t>{cast<uint8_t>((shr<uint8_t>(b,cast<uint64_t>(bit)) & cast<uint8_t>(1ULL)))});
}
}}}
}
else if (((bits[cast<int64_t>(0ULL)] == cast<uint8_t>(1ULL)) && (bits[cast<int64_t>(1ULL)] == cast<uint8_t>(0ULL)))){
Slice<uint8_t> blk = sub(e->data,cast<int64_t>((addr * cast<int64_t>(8ULL))),cast<int64_t>((cast<int64_t>((addr * cast<int64_t>(8ULL))) + cast<int64_t>(8ULL))));
{auto&& tmp39 = blk;
for(int64_t tmp40=0;tmp40<len(tmp39);++tmp40){
auto i=tmp40;uint8_t b={};
{int64_t bit = cast<int64_t>(0ULL);for (;(bit < cast<int64_t>(8ULL));bit++){
b = cast<uint8_t>((shl<uint8_t>(b,cast<int64_t>(1ULL)) | bits[cast<int64_t>((cast<int64_t>((cast<int64_t>((cast<int64_t>(2ULL) + addrBits)) + cast<int64_t>((i * cast<int64_t>(8ULL))))) + bit))]));
}
}blk[i] = b;
}}
e->ready = true;
}
else {
gbamachine_Machine_note(m,std::string("EEPROM request opens with %d%d (neither a read nor a write)",59),bits[cast<int64_t>(0ULL)],bits[cast<int64_t>(1ULL)]);
}
}
tmp36:;
}
}
// tools/platform/gba/gbamachine/eeprom.go:112:1
void gbamachine_eeprom_size(gbamachine_eeprom* e,bool large){
{
if (e->sized) {
return ;
}
e->sized = true;
if ((!large)) {
e->data = sub(e->data,0,cast<int64_t>(512ULL));
}
}
}
// tools/platform/gba/gbamachine/io.go:10:1
uint16_t gbamachine_Machine_ioRead16(gbamachine_Machine* m,uint32_t a){
{
uint32_t reg = (cast<uint32_t>((a & cast<uint32_t>(16777215ULL))) & ~(cast<uint32_t>(1ULL)));
uint16_t v={};
{
switch(reg){
case cast<uint32_t>(0ULL):{
v = get(m->io,reg);
break;}
case cast<uint32_t>(4ULL):{
v = cast<uint16_t>((get(m->io,reg) & cast<uint16_t>(65528ULL)));
if (((m->vid.line >= cast<int64_t>(160ULL)) && (m->vid.line <= cast<int64_t>(226ULL)))) {
v |= cast<uint16_t>(1ULL);
}
if (m->vid.hblank) {
v |= cast<uint16_t>(2ULL);
}
if ((m->vid.line == cast<int64_t>(shr<uint16_t>(get(m->io,cast<uint32_t>(4ULL)),cast<int64_t>(8ULL))))) {
v |= cast<uint16_t>(4ULL);
}
break;}
case cast<uint32_t>(6ULL):{
v = cast<uint16_t>(m->vid.line);
break;}
case cast<uint32_t>(136ULL):{
v = get(m->io,reg);
break;}
case cast<uint32_t>(132ULL):case cast<uint32_t>(144ULL):case cast<uint32_t>(146ULL):case cast<uint32_t>(148ULL):case cast<uint32_t>(150ULL):case cast<uint32_t>(152ULL):case cast<uint32_t>(154ULL):case cast<uint32_t>(156ULL):case cast<uint32_t>(158ULL):{
auto tmp41 = gbamachine_Machine_soundRead(m,reg);
v = std::get<0>(tmp41);
break;}
case cast<uint32_t>(184ULL):case cast<uint32_t>(196ULL):case cast<uint32_t>(208ULL):case cast<uint32_t>(220ULL):{
v = cast<uint16_t>(0ULL);
break;}
case cast<uint32_t>(186ULL):case cast<uint32_t>(198ULL):case cast<uint32_t>(210ULL):case cast<uint32_t>(222ULL):{
v = m->dma[divi<uint32_t>((cast<uint32_t>((reg - cast<uint32_t>(186ULL)))),cast<uint32_t>(12ULL))].ctrl;
break;}
case cast<uint32_t>(256ULL):case cast<uint32_t>(260ULL):case cast<uint32_t>(264ULL):case cast<uint32_t>(268ULL):{
v = m->timers[divi<uint32_t>((cast<uint32_t>((reg - cast<uint32_t>(256ULL)))),cast<uint32_t>(4ULL))].counter;
break;}
case cast<uint32_t>(258ULL):case cast<uint32_t>(262ULL):case cast<uint32_t>(266ULL):case cast<uint32_t>(270ULL):{
v = m->timers[divi<uint32_t>((cast<uint32_t>((reg - cast<uint32_t>(258ULL)))),cast<uint32_t>(4ULL))].ctrl;
break;}
case cast<uint32_t>(304ULL):{
v = cast<uint16_t>((cast<uint16_t>(~m->keys) & cast<uint16_t>(1023ULL)));
break;}
case cast<uint32_t>(306ULL):{
v = get(m->io,reg);
break;}
case cast<uint32_t>(512ULL):{
v = m->ie;
break;}
case cast<uint32_t>(514ULL):{
v = m->if_;
break;}
case cast<uint32_t>(516ULL):{
v = get(m->io,reg);
break;}
case cast<uint32_t>(520ULL):{
if (m->ime) {
v = cast<uint16_t>(1ULL);
}
break;}
case cast<uint32_t>(768ULL):{
v = cast<uint16_t>(1ULL);
break;}
default:{
{
if ((reg <= cast<uint32_t>(84ULL))){
v = get(m->io,reg);
}
else if (((reg >= cast<uint32_t>(96ULL)) && (reg <= cast<uint32_t>(166ULL)))){
v = get(m->io,reg);
}
else if (((reg >= cast<uint32_t>(288ULL)) && (reg <= cast<uint32_t>(346ULL)))){
gbamachine_Machine_note(m,std::string("serial register 0x%03X read (link port not modelled; reads idle)",64),reg);
v = cast<uint16_t>(0ULL);
}
else {
gbamachine_Machine_note(m,std::string("unmodelled I/O register 0x%03X read",35),reg);
v = cast<uint16_t>(0ULL);
}
}
tmp42:;
break;}
}}
if (bool(m->OnIO)) {
m->OnIO(false,reg,v,m->cpu->R[cast<int64_t>(15ULL)]);
}
return v;
}
}
// tools/platform/gba/gbamachine/io.go:80:1
void gbamachine_Machine_ioWrite8(gbamachine_Machine* m,uint32_t a,uint8_t v){
{
uint32_t reg = (a & ~(cast<uint32_t>(1ULL)));
uint16_t old = get(m->io,cast<uint32_t>((reg & cast<uint32_t>(16777215ULL))));
if ((cast<uint32_t>((a & cast<uint32_t>(1ULL))) == cast<uint32_t>(0ULL))) {
gbamachine_Machine_ioWrite16(m,reg,cast<uint16_t>((cast<uint16_t>((old & cast<uint16_t>(65280ULL))) | cast<uint16_t>(v))));
}
else {
gbamachine_Machine_ioWrite16(m,reg,cast<uint16_t>((cast<uint16_t>((old & cast<uint16_t>(255ULL))) | shl<uint16_t>(cast<uint16_t>(v),cast<int64_t>(8ULL)))));
}
}
}
// tools/platform/gba/gbamachine/io.go:91:1
void gbamachine_Machine_ioWrite16(gbamachine_Machine* m,uint32_t a,uint16_t v){
{
uint32_t reg = (cast<uint32_t>((a & cast<uint32_t>(16777215ULL))) & ~(cast<uint32_t>(1ULL)));
if (bool(m->OnIO)) {
m->OnIO(true,reg,v,m->cpu->R[cast<int64_t>(15ULL)]);
}
m->io[reg] = v;
rrGBARegisterWrite(m,reg,v);
{
if (((((((((reg == cast<uint32_t>(40ULL)) || (reg == cast<uint32_t>(42ULL))) || (reg == cast<uint32_t>(44ULL))) || (reg == cast<uint32_t>(46ULL))) || (reg == cast<uint32_t>(56ULL))) || (reg == cast<uint32_t>(58ULL))) || (reg == cast<uint32_t>(60ULL))) || (reg == cast<uint32_t>(62ULL)))){
gbamachine_ppu_reloadAffineRef(&(m->ppu),m,reg);
}
else if ((reg == cast<uint32_t>(514ULL))){
m->if_ &= ~(v);
}
else if ((reg == cast<uint32_t>(512ULL))){
m->ie = v;
}
else if ((reg == cast<uint32_t>(520ULL))){
m->ime = (cast<uint16_t>((v & cast<uint16_t>(1ULL))) != cast<uint16_t>(0ULL));
}
else if (((reg >= cast<uint32_t>(176ULL)) && (reg <= cast<uint32_t>(222ULL)))){
gbamachine_Machine_dmaRegWrite(m,reg,v);
}
else if (((reg >= cast<uint32_t>(256ULL)) && (reg <= cast<uint32_t>(270ULL)))){
gbamachine_Machine_timerRegWrite(m,reg,v);
}
else if ((reg == cast<uint32_t>(769ULL))){
auto tmp44 = std::make_tuple(true,true,cast<uint16_t>(0ULL));
m->waiting = std::get<0>(tmp44);
m->waitAny = std::get<1>(tmp44);
m->waitMask = std::get<2>(tmp44);
}
else if (((reg >= cast<uint32_t>(96ULL)) && (reg <= cast<uint32_t>(166ULL)))){
if ((!gbamachine_Machine_soundWrite(m,reg,v))) {
gbamachine_Machine_note(m,std::string("sound register 0x%03X written but is not decoded (value 0x%04X)",63),reg,v);
}
}
else if (((reg > cast<uint32_t>(84ULL)) && (reg < cast<uint32_t>(96ULL))) || ((reg >= cast<uint32_t>(168ULL)) && (reg < cast<uint32_t>(176ULL))) || ((reg >= cast<uint32_t>(272ULL)) && (reg < cast<uint32_t>(288ULL))) || (((reg >= cast<uint32_t>(308ULL)) && (reg < cast<uint32_t>(512ULL))) && (reg != cast<uint32_t>(306ULL)))){
gbamachine_Machine_note(m,std::string("unmodelled I/O register 0x%03X written (value 0x%04X)",53),reg,v);
}
}
tmp43:;
}
}
// tools/platform/gba/gbamachine/machine.go:152:1
gbamachine_Machine* gbamachine_New(gba_ROM* rom){
{
gbamachine_Machine* m = arenaNew(gbamachine_Machine{{},rom->Data,Slice<uint8_t>::make(cast<int64_t>(262144ULL)),Slice<uint8_t>::make(cast<int64_t>(32768ULL)),Slice<uint8_t>::make(cast<int64_t>(1024ULL)),Slice<uint8_t>::make(cast<int64_t>(98304ULL)),Slice<uint8_t>::make(cast<int64_t>(1024ULL)),Map<uint32_t,uint16_t>{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},Map<std::string,bool>{},{},{},{},{},{},{},{},{},{},{},{},{}});
gbamachine_eeprom_init(&(m->eeprom));
m->apu = gbamachine_newAPU();
m->cpu = arm_NewCPU(arenaNew(gbamachine_bus{m}));
m->cpu->Arch = cast<arm_Variant>(2ULL);
m->cpu->SWI = gbamachine_biosSWI(m);
arm_CPU* c = m->cpu;
arm_CPU_SetCPSR(c,cast<uint32_t>((((arm_CPU_CPSR(c) & ~(cast<uint32_t>(31ULL)))) | cast<uint32_t>(18ULL))));
c->R[cast<int64_t>(13ULL)] = cast<uint32_t>(50364320ULL);
arm_CPU_SetCPSR(c,cast<uint32_t>((((arm_CPU_CPSR(c) & ~(cast<uint32_t>(31ULL)))) | cast<uint32_t>(19ULL))));
c->R[cast<int64_t>(13ULL)] = cast<uint32_t>(50364384ULL);
arm_CPU_SetCPSR(c,cast<uint32_t>((((arm_CPU_CPSR(c) & ~(cast<uint32_t>(31ULL)))) | cast<uint32_t>(31ULL))));
c->R[cast<int64_t>(13ULL)] = cast<uint32_t>(50364160ULL);
auto tmp45 = std::make_tuple(false,true);
c->IRQDisable = std::get<0>(tmp45);
c->FIQDisable = std::get<1>(tmp45);
c->R[cast<int64_t>(15ULL)] = cast<uint32_t>(134217728ULL);
m->io[cast<uint32_t>(136ULL)] = cast<uint16_t>(512ULL);
return m;
}
}
// tools/platform/gba/gbamachine/machine.go:202:1
uint32_t gbamachine_Machine_PC(gbamachine_Machine* m){
{
return m->cpu->R[cast<int64_t>(15ULL)];
}
}
// tools/platform/gba/gbamachine/machine.go:205:1
std::array<uint32_t,16> gbamachine_Machine_Regs(gbamachine_Machine* m){
{
return m->cpu->R;
}
}
// tools/platform/gba/gbamachine/machine.go:208:1
bool gbamachine_Machine_ThumbState(gbamachine_Machine* m){
{
return m->cpu->Thumb;
}
}
// tools/platform/gba/gbamachine/machine.go:212:1
std::tuple<uint16_t,uint16_t,bool> gbamachine_Machine_IRQState(gbamachine_Machine* m){
uint16_t ie{};
uint16_t if_{};
bool ime{};
{
return {m->ie,m->if_,m->ime};
}
}
// tools/platform/gba/gbamachine/machine.go:215:1
bool gbamachine_Machine_IRQDisabled(gbamachine_Machine* m){
{
return m->cpu->IRQDisable;
}
}
// tools/platform/gba/gbamachine/machine.go:218:1
std::string gbamachine_Machine_Parked(gbamachine_Machine* m){
{
{
if ((!m->waiting)){
return std::string("running",7);
}
else if (m->waitAny){
return std::string("halted for IRQ",14);
}
else {
return go_fmt_Sprintf(std::string("IntrWait 0x%X",13),m->waitMask);
}
}
tmp46:;
}
}
// tools/platform/gba/gbamachine/machine.go:230:1
uint64_t gbamachine_Machine_Frame(gbamachine_Machine* m){
{
return m->vid.frames;
}
}
// tools/platform/gba/gbamachine/machine.go:231:1
int64_t gbamachine_Machine_Line(gbamachine_Machine* m){
{
return m->vid.line;
}
}
// tools/platform/gba/gbamachine/machine.go:234:1
std::tuple<bool,std::string> gbamachine_Machine_Halted(gbamachine_Machine* m){
{
return {m->cpu->Halted,m->cpu->HaltReason};
}
}
// tools/platform/gba/gbamachine/machine.go:237:1
uint64_t gbamachine_Machine_Instrs(gbamachine_Machine* m){
{
return m->cpu->Instrs;
}
}
// tools/platform/gba/gbamachine/machine.go:241:1
uint16_t gbamachine_Machine_Reg(gbamachine_Machine* m,uint32_t a){
{
return get(m->io,(cast<uint32_t>((a & cast<uint32_t>(65535ULL))) & ~(cast<uint32_t>(1ULL))));
}
}
// tools/platform/gba/gbamachine/machine.go:244:1
Slice<uint8_t> gbamachine_Machine_Snapshot(gbamachine_Machine* m,uint32_t addr,uint32_t n){
{
gbamachine_bus busStorage{m}; auto*b=&busStorage;
Slice<uint8_t> out = Slice<uint8_t>::make(n);
{auto&& tmp47 = out;
for(int64_t tmp48=0;tmp48<len(tmp47);++tmp48){
auto i=tmp48;out[i] = gbamachine_bus_Read(b,cast<uint32_t>((addr + cast<uint32_t>(i))));
}}
return out;
}
}
// tools/platform/gba/gbamachine/machine.go:254:1
void gbamachine_Machine_Poke(gbamachine_Machine* m,uint32_t addr,Slice<uint8_t> data){
{
gbamachine_bus busStorage{m}; auto*b=&busStorage;
{auto&& tmp49 = data;
for(int64_t tmp50=0;tmp50<len(tmp49);++tmp50){
auto i=tmp50;auto v=tmp49[tmp50];gbamachine_bus_Write(b,cast<uint32_t>((addr + cast<uint32_t>(i))),v);
}}
}
}
// tools/platform/gba/gbamachine/machine.go:263:1
void gbamachine_Machine_SetKeys(gbamachine_Machine* m,uint16_t mask){
{
m->keys = cast<uint16_t>((mask & cast<uint16_t>(1023ULL)));
}
}
// tools/platform/gba/gbamachine/machine.go:266:1
void gbamachine_Machine_AddBreakpoint(gbamachine_Machine* m,uint32_t pc){
{
if ((!m->bps)) {
m->bps = Map<uint32_t,bool>{};
}
m->bps[pc] = true;
}
}
// tools/platform/gba/gbamachine/machine.go:274:1
std::string gbamachine_Machine_GameCode(gbamachine_Machine* m){
{
if ((len(m->rom) < cast<int64_t>(176ULL))) {
return std::string("",0);
}
return cast<std::string>(sub(m->rom,cast<int64_t>(172ULL),cast<int64_t>(176ULL)));
}
}
// tools/platform/gba/gbamachine/ppu.go:26:1
uint32_t gbamachine_ppu_objVRAMBase(gbamachine_ppu* p,gbamachine_Machine* m){
{
if ((cast<uint16_t>((get(m->io,cast<uint32_t>(0ULL)) & cast<uint16_t>(7ULL))) >= cast<uint16_t>(3ULL))) {
return cast<uint32_t>(81920ULL);
}
return cast<uint32_t>(65536ULL);
}
}
// tools/platform/gba/gbamachine/ppu.go:35:1
void gbamachine_ppu_reloadAffineRef(gbamachine_ppu* p,gbamachine_Machine* m,uint32_t reg){
{
auto ref = [&](uint32_t lo)->int32_t{
uint32_t v = cast<uint32_t>((cast<uint32_t>(get(m->io,lo)) | shl<uint32_t>(cast<uint32_t>(get(m->io,cast<uint32_t>((lo + cast<uint32_t>(2ULL))))),cast<int64_t>(16ULL))));
return shr<int32_t>(cast<int32_t>(shl<uint32_t>(v,cast<int64_t>(4ULL))),cast<int64_t>(4ULL));
}
;
{
switch((reg & ~(cast<uint32_t>(2ULL)))){
case cast<uint32_t>(40ULL):{
p->bg2x = ref(cast<uint32_t>(40ULL));
break;}
case cast<uint32_t>(44ULL):{
p->bg2y = ref(cast<uint32_t>(44ULL));
break;}
case cast<uint32_t>(56ULL):{
p->bg3x = ref(cast<uint32_t>(56ULL));
break;}
case cast<uint32_t>(60ULL):{
p->bg3y = ref(cast<uint32_t>(60ULL));
break;}
}}
}
}
// tools/platform/gba/gbamachine/ppu.go:52:1
void gbamachine_ppu_startFrame(gbamachine_ppu* p,gbamachine_Machine* m){
{
gbamachine_ppu_reloadAffineRef(p,m,cast<uint32_t>(40ULL));
gbamachine_ppu_reloadAffineRef(p,m,cast<uint32_t>(44ULL));
gbamachine_ppu_reloadAffineRef(p,m,cast<uint32_t>(56ULL));
gbamachine_ppu_reloadAffineRef(p,m,cast<uint32_t>(60ULL));
}
}
// tools/platform/gba/gbamachine/ppu.go:60:1
void gbamachine_ppu_stepAffine(gbamachine_ppu* p,gbamachine_Machine* m){
{
p->bg2x += cast<int32_t>(cast<int16_t>(get(m->io,cast<uint32_t>(34ULL))));
p->bg2y += cast<int32_t>(cast<int16_t>(get(m->io,cast<uint32_t>(38ULL))));
p->bg3x += cast<int32_t>(cast<int16_t>(get(m->io,cast<uint32_t>(50ULL))));
p->bg3y += cast<int32_t>(cast<int16_t>(get(m->io,cast<uint32_t>(54ULL))));
}
}
// tools/platform/gba/gbamachine/ppu.go:68:1
uint32_t gbamachine_rgb15(uint16_t c){
{
uint32_t r = cast<uint32_t>((cast<uint32_t>(c) & cast<uint32_t>(31ULL)));
uint32_t g = cast<uint32_t>((cast<uint32_t>(shr<uint16_t>(c,cast<int64_t>(5ULL))) & cast<uint32_t>(31ULL)));
uint32_t b = cast<uint32_t>((cast<uint32_t>(shr<uint16_t>(c,cast<int64_t>(10ULL))) & cast<uint32_t>(31ULL)));
auto exp = [&](uint32_t v)->uint32_t{
return cast<uint32_t>((shl<uint32_t>(v,cast<int64_t>(3ULL)) | shr<uint32_t>(v,cast<int64_t>(2ULL))));
}
;
return cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(4278190080ULL) | shl<uint32_t>(exp(r),cast<int64_t>(16ULL)))) | shl<uint32_t>(exp(g),cast<int64_t>(8ULL)))) | exp(b)));
}
}
// tools/platform/gba/gbamachine/ppu.go:76:1
uint16_t gbamachine_Machine_pal16(gbamachine_Machine* m,int64_t bank,int64_t idx){
{
int64_t o = cast<int64_t>((cast<int64_t>((bank * cast<int64_t>(32ULL))) + cast<int64_t>((idx * cast<int64_t>(2ULL)))));
return cast<uint16_t>((cast<uint16_t>(m->pal[o]) | shl<uint16_t>(cast<uint16_t>(m->pal[cast<int64_t>((o + cast<int64_t>(1ULL)))]),cast<int64_t>(8ULL))));
}
}
// tools/platform/gba/gbamachine/ppu.go:88:1
void gbamachine_ppu_renderLine(gbamachine_ppu* p,gbamachine_Machine* m,int64_t y){
{rrprof::Scope timing(1,"PPU scanline composition");
{if(rrcapture::trace.active)rrgba::beginLine(m,y);
{std::array<uint16_t,4>rrPriority{};for(unsigned i=0;i<4;i++)rrPriority[i]=get(m->io,8u+2u*i);
{
uint16_t dispcnt = get(m->io,cast<uint32_t>(0ULL));
Slice<uint32_t> out = sub(m->screen,cast<int64_t>((y * cast<int64_t>(240ULL))),cast<int64_t>(((cast<int64_t>((y + cast<int64_t>(1ULL)))) * cast<int64_t>(240ULL))));
if ((cast<uint16_t>((dispcnt & cast<uint16_t>(128ULL))) != cast<uint16_t>(0ULL))) {
{auto&& tmp51 = out;
for(int64_t tmp52=0;tmp52<len(tmp51);++tmp52){
auto i=tmp52;out[i] = cast<uint32_t>(4294967295ULL);
if(rrcapture::trace.active){rrgba::clean();rrcapture::trace.record(rrgba::frameBase+(y*240+i)*4,0xffffffff,4,m->Steps,m->cpu->R[15],rrcapture::trace.current);}
}}
return ;
}
int64_t mode = cast<int64_t>(cast<uint16_t>((dispcnt & cast<uint16_t>(7ULL))));
std::array<gbamachine_lineBuf,4> bg={};
auto bgEnabled = [&](int64_t n)->bool{
return (cast<uint16_t>((dispcnt & (shl<uint16_t>(cast<uint16_t>(1ULL),(cast<int64_t>((cast<int64_t>(8ULL) + n))))))) != cast<uint16_t>(0ULL));
}
;
{
switch(mode){
case cast<int64_t>(0ULL):{
{int64_t n = cast<int64_t>(0ULL);for (;(n < cast<int64_t>(4ULL));n++){
if (bgEnabled(n)) {
gbamachine_ppu_textLine(p,m,n,y,(&bg[n]));
}
}
}break;}
case cast<int64_t>(1ULL):{
if (bgEnabled(cast<int64_t>(0ULL))) {
gbamachine_ppu_textLine(p,m,cast<int64_t>(0ULL),y,(&bg[cast<int64_t>(0ULL)]));
}
if (bgEnabled(cast<int64_t>(1ULL))) {
gbamachine_ppu_textLine(p,m,cast<int64_t>(1ULL),y,(&bg[cast<int64_t>(1ULL)]));
}
if (bgEnabled(cast<int64_t>(2ULL))) {
gbamachine_ppu_affineLine(p,m,cast<int64_t>(2ULL),(&bg[cast<int64_t>(2ULL)]));
}
break;}
case cast<int64_t>(2ULL):{
if (bgEnabled(cast<int64_t>(2ULL))) {
gbamachine_ppu_affineLine(p,m,cast<int64_t>(2ULL),(&bg[cast<int64_t>(2ULL)]));
}
if (bgEnabled(cast<int64_t>(3ULL))) {
gbamachine_ppu_affineLine(p,m,cast<int64_t>(3ULL),(&bg[cast<int64_t>(3ULL)]));
}
break;}
case cast<int64_t>(3ULL):case cast<int64_t>(4ULL):case cast<int64_t>(5ULL):{
if (bgEnabled(cast<int64_t>(2ULL))) {
gbamachine_ppu_bitmapLine(p,m,mode,(&bg[cast<int64_t>(2ULL)]));
}
break;}
default:{
gbamachine_Machine_note(m,std::string("display mode %d selected (undefined on hardware)",48),mode);
break;}
}}
std::array<uint16_t,240> objC={};
std::array<bool,240> objOn={};
std::array<uint8_t,240> objPrio={};
std::array<bool,240> objSemi={};
std::array<bool,240> objWin={};
if ((cast<uint16_t>((dispcnt & cast<uint16_t>(4096ULL))) != cast<uint16_t>(0ULL))) {
gbamachine_ppu_objLine(p,m,y,dispcnt,sub(objC,0,len(objC)),sub(objOn,0,len(objOn)),sub(objPrio,0,len(objPrio)),sub(objSemi,0,len(objSemi)),sub(objWin,0,len(objWin)));
}
bool winEnabled = (cast<uint16_t>((dispcnt & cast<uint16_t>(57344ULL))) != cast<uint16_t>(0ULL));
auto tmp53 = std::make_tuple(get(m->io,cast<uint32_t>(72ULL)),get(m->io,cast<uint32_t>(74ULL)));
uint16_t winIn = std::get<0>(tmp53);
uint16_t winOut = std::get<1>(tmp53);
auto winCtl = [&](int64_t x)->uint16_t{
if ((!winEnabled)) {
return cast<uint16_t>(63ULL);
}
if (((cast<uint16_t>((dispcnt & cast<uint16_t>(8192ULL))) != cast<uint16_t>(0ULL)) && gbamachine_ppu_inWindow(p,m,cast<uint32_t>(64ULL),cast<uint32_t>(68ULL),x,y))) {
return cast<uint16_t>((winIn & cast<uint16_t>(63ULL)));
}
if (((cast<uint16_t>((dispcnt & cast<uint16_t>(16384ULL))) != cast<uint16_t>(0ULL)) && gbamachine_ppu_inWindow(p,m,cast<uint32_t>(66ULL),cast<uint32_t>(70ULL),x,y))) {
return cast<uint16_t>((shr<uint16_t>(winIn,cast<int64_t>(8ULL)) & cast<uint16_t>(63ULL)));
}
if (((cast<uint16_t>((dispcnt & cast<uint16_t>(32768ULL))) != cast<uint16_t>(0ULL)) && objWin[x])) {
return cast<uint16_t>((shr<uint16_t>(winOut,cast<int64_t>(8ULL)) & cast<uint16_t>(63ULL)));
}
return cast<uint16_t>((winOut & cast<uint16_t>(63ULL)));
}
;
uint16_t bldcnt = get(m->io,cast<uint32_t>(80ULL));
uint16_t bldMode = cast<uint16_t>((shr<uint16_t>(bldcnt,cast<int64_t>(6ULL)) & cast<uint16_t>(3ULL)));
int64_t eva = cast<int64_t>((cast<int64_t>(get(m->io,cast<uint32_t>(82ULL))) & cast<int64_t>(31ULL)));
int64_t evb = cast<int64_t>((cast<int64_t>(shr<uint16_t>(get(m->io,cast<uint32_t>(82ULL)),cast<int64_t>(8ULL))) & cast<int64_t>(31ULL)));
int64_t evy = cast<int64_t>((cast<int64_t>(get(m->io,cast<uint32_t>(84ULL))) & cast<int64_t>(31ULL)));
if ((eva > cast<int64_t>(16ULL))) {
eva = cast<int64_t>(16ULL);
}
if ((evb > cast<int64_t>(16ULL))) {
evb = cast<int64_t>(16ULL);
}
if ((evy > cast<int64_t>(16ULL))) {
evy = cast<int64_t>(16ULL);
}
uint16_t backdrop = gbamachine_Machine_pal16(m,cast<int64_t>(0ULL),cast<int64_t>(0ULL));
{int64_t x = cast<int64_t>(0ULL);for (;(x < cast<int64_t>(240ULL));x++){
uint16_t ctl = winCtl(x);
auto tmp54 = std::make_tuple(cast<int64_t>(5ULL),cast<int64_t>(5ULL));
int64_t top = std::get<0>(tmp54);
int64_t second = std::get<1>(tmp54);
auto tmp55 = std::make_tuple(backdrop,backdrop);
uint16_t topC = std::get<0>(tmp55);
uint16_t secondC = std::get<1>(tmp55);
bool topSemi = false;
int64_t found = cast<int64_t>(0ULL);
{int64_t prio = cast<int64_t>(0ULL);for (;((prio < cast<int64_t>(4ULL)) && (found < cast<int64_t>(2ULL)));prio++){
if (((objOn[x] && (cast<int64_t>(objPrio[x]) == prio)) && (cast<uint16_t>((ctl & cast<uint16_t>(16ULL))) != cast<uint16_t>(0ULL)))) {
if ((found == cast<int64_t>(0ULL))) {
auto tmp56 = std::make_tuple(cast<int64_t>(4ULL),objC[x],objSemi[x]);
top = std::get<0>(tmp56);
topC = std::get<1>(tmp56);
topSemi = std::get<2>(tmp56);
}
else {
auto tmp57 = std::make_tuple(cast<int64_t>(4ULL),objC[x]);
second = std::get<0>(tmp57);
secondC = std::get<1>(tmp57);
}
found++;
if ((found == cast<int64_t>(2ULL))) {
break;
}
}
{int64_t n = cast<int64_t>(0ULL);for (;((n < cast<int64_t>(4ULL)) && (found < cast<int64_t>(2ULL)));n++){
if (((bg[n].on[x] && (cast<int64_t>(cast<uint16_t>((rrPriority[n] & cast<uint16_t>(3ULL)))) == prio)) && (cast<uint16_t>((ctl & (shl<uint16_t>(cast<uint16_t>(1ULL),cast<uint64_t>(n))))) != cast<uint16_t>(0ULL)))) {
if ((found == cast<int64_t>(0ULL))) {
auto tmp58 = std::make_tuple(n,bg[n].c[x]);
top = std::get<0>(tmp58);
topC = std::get<1>(tmp58);
}
else {
auto tmp59 = std::make_tuple(n,bg[n].c[x]);
second = std::get<0>(tmp59);
secondC = std::get<1>(tmp59);
}
found++;
}
}
}}
}uint16_t c = topC;
bool effects = (cast<uint16_t>((ctl & cast<uint16_t>(32ULL))) != cast<uint16_t>(0ULL));
auto firstMask = [&](int64_t layer)->bool{
return (cast<uint16_t>((bldcnt & (shl<uint16_t>(cast<uint16_t>(1ULL),cast<uint64_t>(layer))))) != cast<uint16_t>(0ULL));
}
;
auto secondMask = [&](int64_t layer)->bool{
return (cast<uint16_t>((bldcnt & (shl<uint16_t>(cast<uint16_t>(1ULL),cast<uint64_t>(cast<int64_t>((cast<int64_t>(8ULL) + layer))))))) != cast<uint16_t>(0ULL));
}
;
{
if (((topSemi && effects) && secondMask(second))){
c = gbamachine_blend(topC,secondC,eva,evb);
}
else if (((((bldMode == cast<uint16_t>(1ULL)) && effects) && firstMask(top)) && secondMask(second))){
c = gbamachine_blend(topC,secondC,eva,evb);
}
else if ((((bldMode == cast<uint16_t>(2ULL)) && effects) && firstMask(top))){
c = gbamachine_brighten(topC,evy);
}
else if ((((bldMode == cast<uint16_t>(3ULL)) && effects) && firstMask(top))){
c = gbamachine_darken(topC,evy);
}
}
tmp60:;
out[x] = gbamachine_rgb15(c);
if(rrcapture::trace.active)rrgba::compose(m,x,y,top,second,topC,secondC,c,ctl);
}
}}
}
}
}
}
// tools/platform/gba/gbamachine/ppu.go:235:1
bool gbamachine_ppu_inWindow(gbamachine_ppu* p,gbamachine_Machine* m,uint32_t hreg,uint32_t vreg,int64_t x,int64_t y){
{
auto tmp61 = std::make_tuple(get(m->io,hreg),get(m->io,vreg));
uint16_t h = std::get<0>(tmp61);
uint16_t v = std::get<1>(tmp61);
auto tmp62 = std::make_tuple(cast<int64_t>(shr<uint16_t>(h,cast<int64_t>(8ULL))),cast<int64_t>(cast<uint16_t>((h & cast<uint16_t>(255ULL)))));
int64_t x1 = std::get<0>(tmp62);
int64_t x2 = std::get<1>(tmp62);
auto tmp63 = std::make_tuple(cast<int64_t>(shr<uint16_t>(v,cast<int64_t>(8ULL))),cast<int64_t>(cast<uint16_t>((v & cast<uint16_t>(255ULL)))));
int64_t y1 = std::get<0>(tmp63);
int64_t y2 = std::get<1>(tmp63);
bool inX = ((x >= x1) && (x < x2));
if ((x1 > x2)) {
inX = ((x >= x1) || (x < x2));
}
bool inY = ((y >= y1) && (y < y2));
if ((y1 > y2)) {
inY = ((y >= y1) || (y < y2));
}
return (inX && inY);
}
}
// tools/platform/gba/gbamachine/ppu.go:250:1
uint16_t gbamachine_blend(uint16_t a,uint16_t b,int64_t eva,int64_t evb){
{
auto mix = [&](uint16_t ca,uint16_t cb,uint64_t sh)->uint16_t{
int64_t v = shr<int64_t>((cast<int64_t>((cast<int64_t>((cast<int64_t>(cast<uint16_t>((shr<uint16_t>(ca,sh) & cast<uint16_t>(31ULL)))) * eva)) + cast<int64_t>((cast<int64_t>(cast<uint16_t>((shr<uint16_t>(cb,sh) & cast<uint16_t>(31ULL)))) * evb))))),cast<int64_t>(4ULL));
if ((v > cast<int64_t>(31ULL))) {
v = cast<int64_t>(31ULL);
}
return shl<uint16_t>(cast<uint16_t>(v),sh);
}
;
return cast<uint16_t>((cast<uint16_t>((mix(a,b,cast<uint64_t>(0ULL)) | mix(a,b,cast<uint64_t>(5ULL)))) | mix(a,b,cast<uint64_t>(10ULL))));
}
}
// tools/platform/gba/gbamachine/ppu.go:261:1
uint16_t gbamachine_brighten(uint16_t c,int64_t evy){
{
auto f = [&](uint64_t sh)->uint16_t{
int64_t v = cast<int64_t>(cast<uint16_t>((shr<uint16_t>(c,sh) & cast<uint16_t>(31ULL))));
v += shr<int64_t>(cast<int64_t>(((cast<int64_t>((cast<int64_t>(31ULL) - v))) * evy)),cast<int64_t>(4ULL));
return shl<uint16_t>(cast<uint16_t>(v),sh);
}
;
return cast<uint16_t>((cast<uint16_t>((f(cast<uint64_t>(0ULL)) | f(cast<uint64_t>(5ULL)))) | f(cast<uint64_t>(10ULL))));
}
}
// tools/platform/gba/gbamachine/ppu.go:270:1
uint16_t gbamachine_darken(uint16_t c,int64_t evy){
{
auto f = [&](uint64_t sh)->uint16_t{
int64_t v = cast<int64_t>(cast<uint16_t>((shr<uint16_t>(c,sh) & cast<uint16_t>(31ULL))));
v -= shr<int64_t>(cast<int64_t>((v * evy)),cast<int64_t>(4ULL));
return shl<uint16_t>(cast<uint16_t>(v),sh);
}
;
return cast<uint16_t>((cast<uint16_t>((f(cast<uint64_t>(0ULL)) | f(cast<uint64_t>(5ULL)))) | f(cast<uint64_t>(10ULL))));
}
}
// tools/platform/gba/gbamachine/ppu.go:281:1
void gbamachine_ppu_textLine(gbamachine_ppu* p,gbamachine_Machine* m,int64_t n,int64_t y,gbamachine_lineBuf* out){
{if(rrcapture::trace.active)rrgba::event(m,"PPU text background",n,y);
{
uint16_t cnt = get(m->io,cast<uint32_t>((cast<uint32_t>(8ULL) + cast<uint32_t>((cast<uint32_t>(2ULL) * cast<uint32_t>(n))))));
if ((cast<uint16_t>((cnt & cast<uint16_t>(64ULL))) != cast<uint16_t>(0ULL))) {
gbamachine_Machine_note(m,std::string("BG%d uses mosaic (not modelled)",31),n);
}
uint32_t charBase = cast<uint32_t>((cast<uint32_t>(cast<uint16_t>((shr<uint16_t>(cnt,cast<int64_t>(2ULL)) & cast<uint16_t>(3ULL)))) * cast<uint32_t>(16384ULL)));
uint32_t scrBase = cast<uint32_t>((cast<uint32_t>(cast<uint16_t>((shr<uint16_t>(cnt,cast<int64_t>(8ULL)) & cast<uint16_t>(31ULL)))) * cast<uint32_t>(2048ULL)));
bool eightBpp = (cast<uint16_t>((cnt & cast<uint16_t>(128ULL))) != cast<uint16_t>(0ULL));
uint16_t size = cast<uint16_t>((shr<uint16_t>(cnt,cast<int64_t>(14ULL)) & cast<uint16_t>(3ULL)));
auto tmp64 = std::make_tuple(cast<int64_t>(256ULL),cast<int64_t>(256ULL));
int64_t w = std::get<0>(tmp64);
int64_t h = std::get<1>(tmp64);
if (((size == cast<uint16_t>(1ULL)) || (size == cast<uint16_t>(3ULL)))) {
w = cast<int64_t>(512ULL);
}
if (((size == cast<uint16_t>(2ULL)) || (size == cast<uint16_t>(3ULL)))) {
h = cast<int64_t>(512ULL);
}
int64_t hofs = cast<int64_t>(cast<uint16_t>((get(m->io,cast<uint32_t>((cast<uint32_t>(16ULL) + cast<uint32_t>((cast<uint32_t>(4ULL) * cast<uint32_t>(n)))))) & cast<uint16_t>(511ULL))));
int64_t vofs = cast<int64_t>(cast<uint16_t>((get(m->io,cast<uint32_t>((cast<uint32_t>(18ULL) + cast<uint32_t>((cast<uint32_t>(4ULL) * cast<uint32_t>(n)))))) & cast<uint16_t>(511ULL))));
int64_t sy = cast<int64_t>(((cast<int64_t>((y + vofs))) & (cast<int64_t>((h - cast<int64_t>(1ULL))))));
{int64_t x = cast<int64_t>(0ULL);for (;(x < cast<int64_t>(240ULL));x++){
int64_t sx = cast<int64_t>(((cast<int64_t>((x + hofs))) & (cast<int64_t>((w - cast<int64_t>(1ULL))))));
int64_t quad = cast<int64_t>(0ULL);
if (((w == cast<int64_t>(512ULL)) && (sx >= cast<int64_t>(256ULL)))) {
quad++;
}
if (((h == cast<int64_t>(512ULL)) && (sy >= cast<int64_t>(256ULL)))) {
quad += divi<int64_t>(w,cast<int64_t>(256ULL));
}
uint32_t entryOff = cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((scrBase + cast<uint32_t>((cast<uint32_t>(quad) * cast<uint32_t>(2048ULL))))) + cast<uint32_t>((cast<uint32_t>(shr<int64_t>((cast<int64_t>((sy & cast<int64_t>(255ULL)))),cast<int64_t>(3ULL))) * cast<uint32_t>(64ULL))))) + cast<uint32_t>((cast<uint32_t>(shr<int64_t>((cast<int64_t>((sx & cast<int64_t>(255ULL)))),cast<int64_t>(3ULL))) * cast<uint32_t>(2ULL)))));
uint16_t entry = cast<uint16_t>((cast<uint16_t>(m->vram[entryOff]) | shl<uint16_t>(cast<uint16_t>(m->vram[cast<uint32_t>((entryOff + cast<uint32_t>(1ULL)))]),cast<int64_t>(8ULL))));
uint32_t tile = cast<uint32_t>(cast<uint16_t>((entry & cast<uint16_t>(1023ULL))));
auto tmp65 = std::make_tuple(cast<int64_t>((sx & cast<int64_t>(7ULL))),cast<int64_t>((sy & cast<int64_t>(7ULL))));
int64_t tx = std::get<0>(tmp65);
int64_t ty = std::get<1>(tmp65);
if ((cast<uint16_t>((entry & cast<uint16_t>(1024ULL))) != cast<uint16_t>(0ULL))) {
tx = cast<int64_t>((cast<int64_t>(7ULL) - tx));
}
if ((cast<uint16_t>((entry & cast<uint16_t>(2048ULL))) != cast<uint16_t>(0ULL))) {
ty = cast<int64_t>((cast<int64_t>(7ULL) - ty));
}
int64_t idx={};
int64_t bank={};
if (eightBpp) {
uint32_t a = cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((charBase + cast<uint32_t>((tile * cast<uint32_t>(64ULL))))) + cast<uint32_t>((cast<uint32_t>(ty) * cast<uint32_t>(8ULL))))) + cast<uint32_t>(tx)));
if ((a >= cast<uint32_t>(65536ULL))) {
continue;
}
auto tmp66 = std::make_tuple(cast<int64_t>(m->vram[a]),cast<int64_t>(0ULL));
idx = std::get<0>(tmp66);
bank = std::get<1>(tmp66);
if(rrcapture::trace.active)rrgba::source(m,n,x,a,idx,bank,tx,ty);
}
else {
uint32_t a = cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((charBase + cast<uint32_t>((tile * cast<uint32_t>(32ULL))))) + cast<uint32_t>((cast<uint32_t>(ty) * cast<uint32_t>(4ULL))))) + cast<uint32_t>(shr<int64_t>(tx,cast<int64_t>(1ULL)))));
if ((a >= cast<uint32_t>(65536ULL))) {
continue;
}
uint8_t b = m->vram[a];
if ((cast<int64_t>((tx & cast<int64_t>(1ULL))) == cast<int64_t>(1ULL))) {
idx = cast<int64_t>(shr<uint8_t>(b,cast<int64_t>(4ULL)));
}
else {
idx = cast<int64_t>(cast<uint8_t>((b & cast<uint8_t>(15ULL))));
}
bank = cast<int64_t>(shr<uint16_t>(entry,cast<int64_t>(12ULL)));
if(rrcapture::trace.active)rrgba::source(m,n,x,a,idx,bank,tx,ty);
}
if ((idx != cast<int64_t>(0ULL))) {
out->c[x] = gbamachine_Machine_pal16(m,bank,idx);
out->on[x] = true;
if(rrcapture::trace.active)rrgba::dot(m,n,x,out->c[x]);
}
}
}}
}
}
// tools/platform/gba/gbamachine/ppu.go:350:1
void gbamachine_ppu_affineLine(gbamachine_ppu* p,gbamachine_Machine* m,int64_t n,gbamachine_lineBuf* out){
{if(rrcapture::trace.active)rrgba::event(m,"PPU affine background",n,rrgba::currentY);
{
uint16_t cnt = get(m->io,cast<uint32_t>((cast<uint32_t>(8ULL) + cast<uint32_t>((cast<uint32_t>(2ULL) * cast<uint32_t>(n))))));
uint32_t charBase = cast<uint32_t>((cast<uint32_t>(cast<uint16_t>((shr<uint16_t>(cnt,cast<int64_t>(2ULL)) & cast<uint16_t>(3ULL)))) * cast<uint32_t>(16384ULL)));
uint32_t scrBase = cast<uint32_t>((cast<uint32_t>(cast<uint16_t>((shr<uint16_t>(cnt,cast<int64_t>(8ULL)) & cast<uint16_t>(31ULL)))) * cast<uint32_t>(2048ULL)));
bool wrap = (cast<uint16_t>((cnt & cast<uint16_t>(8192ULL))) != cast<uint16_t>(0ULL));
std::array<int64_t,4> sizes = std::array<int64_t,4>{cast<int64_t>(128ULL),cast<int64_t>(256ULL),cast<int64_t>(512ULL),cast<int64_t>(1024ULL)};
int64_t size = sizes[cast<uint16_t>((shr<uint16_t>(cnt,cast<int64_t>(14ULL)) & cast<uint16_t>(3ULL)))];
int32_t cx={};
int32_t cy={};
int32_t pa={};
int32_t pc={};
if ((n == cast<int64_t>(2ULL))) {
auto tmp67 = std::make_tuple(p->bg2x,p->bg2y);
cx = std::get<0>(tmp67);
cy = std::get<1>(tmp67);
auto tmp68 = std::make_tuple(cast<int32_t>(cast<int16_t>(get(m->io,cast<uint32_t>(32ULL)))),cast<int32_t>(cast<int16_t>(get(m->io,cast<uint32_t>(36ULL)))));
pa = std::get<0>(tmp68);
pc = std::get<1>(tmp68);
}
else {
auto tmp69 = std::make_tuple(p->bg3x,p->bg3y);
cx = std::get<0>(tmp69);
cy = std::get<1>(tmp69);
auto tmp70 = std::make_tuple(cast<int32_t>(cast<int16_t>(get(m->io,cast<uint32_t>(48ULL)))),cast<int32_t>(cast<int16_t>(get(m->io,cast<uint32_t>(52ULL)))));
pa = std::get<0>(tmp70);
pc = std::get<1>(tmp70);
}
{int64_t x = cast<int64_t>(0ULL);for (;(x < cast<int64_t>(240ULL));x++){
auto tmp71 = std::make_tuple(cast<int64_t>(shr<int32_t>(cx,cast<int64_t>(8ULL))),cast<int64_t>(shr<int32_t>(cy,cast<int64_t>(8ULL))));
int64_t tx = std::get<0>(tmp71);
int64_t ty = std::get<1>(tmp71);
cx += pa;
cy += pc;
if (wrap) {
tx &= cast<int64_t>((size - cast<int64_t>(1ULL)));
ty &= cast<int64_t>((size - cast<int64_t>(1ULL)));
}
else if (((((tx < cast<int64_t>(0ULL)) || (ty < cast<int64_t>(0ULL))) || (tx >= size)) || (ty >= size))) {
continue;
}
int64_t tilesPerRow = divi<int64_t>(size,cast<int64_t>(8ULL));
uint32_t mapOff = cast<uint32_t>((cast<uint32_t>((scrBase + cast<uint32_t>((cast<uint32_t>(shr<int64_t>(ty,cast<int64_t>(3ULL))) * cast<uint32_t>(tilesPerRow))))) + cast<uint32_t>(shr<int64_t>(tx,cast<int64_t>(3ULL)))));
uint32_t tile = cast<uint32_t>(m->vram[mapOff]);
uint32_t a = cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((charBase + cast<uint32_t>((tile * cast<uint32_t>(64ULL))))) + cast<uint32_t>((cast<uint32_t>(cast<int64_t>((ty & cast<int64_t>(7ULL)))) * cast<uint32_t>(8ULL))))) + cast<uint32_t>(cast<int64_t>((tx & cast<int64_t>(7ULL))))));
if ((a >= cast<uint32_t>(65536ULL))) {
continue;
}
{
int64_t idx = cast<int64_t>(m->vram[a]);
if ((idx != cast<int64_t>(0ULL))) {
out->c[x] = gbamachine_Machine_pal16(m,cast<int64_t>(0ULL),idx);
out->on[x] = true;
if(rrcapture::trace.active){rrgba::source(m,n,x,a,idx,0,tx,ty);rrgba::dot(m,n,x,out->c[x]);}
}
}
}
}}
}
}
// tools/platform/gba/gbamachine/ppu.go:394:1
void gbamachine_ppu_bitmapLine(gbamachine_ppu* p,gbamachine_Machine* m,int64_t mode,gbamachine_lineBuf* out){
{if(rrcapture::trace.active)rrgba::event(m,"PPU bitmap background",2,rrgba::currentY);
{
auto tmp72 = std::make_tuple(cast<int32_t>(cast<int16_t>(get(m->io,cast<uint32_t>(32ULL)))),cast<int32_t>(cast<int16_t>(get(m->io,cast<uint32_t>(36ULL)))));
int32_t pa = std::get<0>(tmp72);
int32_t pc = std::get<1>(tmp72);
auto tmp73 = std::make_tuple(p->bg2x,p->bg2y);
int32_t cx = std::get<0>(tmp73);
int32_t cy = std::get<1>(tmp73);
uint32_t page = cast<uint32_t>(0ULL);
if (((cast<uint16_t>((get(m->io,cast<uint32_t>(0ULL)) & cast<uint16_t>(16ULL))) != cast<uint16_t>(0ULL)) && (mode != cast<int64_t>(3ULL)))) {
page = cast<uint32_t>(40960ULL);
}
auto tmp74 = std::make_tuple(cast<int64_t>(240ULL),cast<int64_t>(160ULL));
int64_t w = std::get<0>(tmp74);
int64_t h = std::get<1>(tmp74);
if ((mode == cast<int64_t>(5ULL))) {
auto tmp75 = std::make_tuple(cast<int64_t>(160ULL),cast<int64_t>(128ULL));
w = std::get<0>(tmp75);
h = std::get<1>(tmp75);
}
{int64_t x = cast<int64_t>(0ULL);for (;(x < cast<int64_t>(240ULL));x++){
auto tmp76 = std::make_tuple(cast<int64_t>(shr<int32_t>(cx,cast<int64_t>(8ULL))),cast<int64_t>(shr<int32_t>(cy,cast<int64_t>(8ULL))));
int64_t tx = std::get<0>(tmp76);
int64_t ty = std::get<1>(tmp76);
cx += pa;
cy += pc;
if (((((tx < cast<int64_t>(0ULL)) || (ty < cast<int64_t>(0ULL))) || (tx >= w)) || (ty >= h))) {
continue;
}
{
switch(mode){
case cast<int64_t>(3ULL):{
uint32_t o = cast<uint32_t>(((cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(ty) * cast<uint32_t>(w))) + cast<uint32_t>(tx)))) * cast<uint32_t>(2ULL)));
out->c[x] = cast<uint16_t>((cast<uint16_t>(m->vram[o]) | shl<uint16_t>(cast<uint16_t>(m->vram[cast<uint32_t>((o + cast<uint32_t>(1ULL)))]),cast<int64_t>(8ULL))));
out->on[x] = true;
if(rrcapture::trace.active){auto a=page+(ty*w+tx)*(mode==4?1:2);rrgba::source(m,2,x,a,mode==4?m->vram[a]:0,mode==4?0:-1,tx,ty);rrgba::dot(m,2,x,out->c[x]);}
break;}
case cast<int64_t>(4ULL):{
{
int64_t idx = cast<int64_t>(m->vram[cast<uint32_t>((cast<uint32_t>((page + cast<uint32_t>((cast<uint32_t>(ty) * cast<uint32_t>(w))))) + cast<uint32_t>(tx)))]);
if ((idx != cast<int64_t>(0ULL))) {
out->c[x] = gbamachine_Machine_pal16(m,cast<int64_t>(0ULL),idx);
out->on[x] = true;
if(rrcapture::trace.active){auto a=page+(ty*w+tx)*(mode==4?1:2);rrgba::source(m,2,x,a,mode==4?m->vram[a]:0,mode==4?0:-1,tx,ty);rrgba::dot(m,2,x,out->c[x]);}
}
}
break;}
case cast<int64_t>(5ULL):{
uint32_t o = cast<uint32_t>((page + cast<uint32_t>(((cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(ty) * cast<uint32_t>(w))) + cast<uint32_t>(tx)))) * cast<uint32_t>(2ULL)))));
out->c[x] = cast<uint16_t>((cast<uint16_t>(m->vram[o]) | shl<uint16_t>(cast<uint16_t>(m->vram[cast<uint32_t>((o + cast<uint32_t>(1ULL)))]),cast<int64_t>(8ULL))));
out->on[x] = true;
if(rrcapture::trace.active){auto a=page+(ty*w+tx)*(mode==4?1:2);rrgba::source(m,2,x,a,mode==4?m->vram[a]:0,mode==4?0:-1,tx,ty);rrgba::dot(m,2,x,out->c[x]);}
break;}
}}
}
}}
}
}
// tools/platform/gba/gbamachine/ppu.go:441:1
void gbamachine_ppu_objLine(gbamachine_ppu* p,gbamachine_Machine* m,int64_t y,uint16_t dispcnt,Slice<uint16_t> c,Slice<bool> on,Slice<uint8_t> prio,Slice<bool> semi,Slice<bool> win){
{
bool oneDim = (cast<uint16_t>((dispcnt & cast<uint16_t>(64ULL))) != cast<uint16_t>(0ULL));
uint32_t minTile = cast<uint32_t>(0ULL);
if ((cast<uint16_t>((dispcnt & cast<uint16_t>(7ULL))) >= cast<uint16_t>(3ULL))) {
minTile = cast<uint32_t>(512ULL);
}
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(128ULL));i++){
int64_t o = cast<int64_t>((i * cast<int64_t>(8ULL)));
uint16_t attr0 = cast<uint16_t>((cast<uint16_t>(m->oam[o]) | shl<uint16_t>(cast<uint16_t>(m->oam[cast<int64_t>((o + cast<int64_t>(1ULL)))]),cast<int64_t>(8ULL))));
uint16_t attr1 = cast<uint16_t>((cast<uint16_t>(m->oam[cast<int64_t>((o + cast<int64_t>(2ULL)))]) | shl<uint16_t>(cast<uint16_t>(m->oam[cast<int64_t>((o + cast<int64_t>(3ULL)))]),cast<int64_t>(8ULL))));
uint16_t attr2 = cast<uint16_t>((cast<uint16_t>(m->oam[cast<int64_t>((o + cast<int64_t>(4ULL)))]) | shl<uint16_t>(cast<uint16_t>(m->oam[cast<int64_t>((o + cast<int64_t>(5ULL)))]),cast<int64_t>(8ULL))));
bool affine = (cast<uint16_t>((attr0 & cast<uint16_t>(256ULL))) != cast<uint16_t>(0ULL));
if (((!affine) && (cast<uint16_t>((attr0 & cast<uint16_t>(512ULL))) != cast<uint16_t>(0ULL)))) {
continue;
}
uint16_t mode = cast<uint16_t>((shr<uint16_t>(attr0,cast<int64_t>(10ULL)) & cast<uint16_t>(3ULL)));
if ((mode == cast<uint16_t>(3ULL))) {
continue;
}
if ((cast<uint16_t>((attr0 & cast<uint16_t>(4096ULL))) != cast<uint16_t>(0ULL))) {
gbamachine_Machine_note(m,std::string("sprite uses mosaic (not modelled)",33));
}
int64_t shape = cast<int64_t>(shr<uint16_t>(attr0,cast<int64_t>(14ULL)));
if ((shape == cast<int64_t>(3ULL))) {
continue;
}
int64_t w = gbamachine_objSizes[shape][shr<uint16_t>(attr1,cast<int64_t>(14ULL))][cast<int64_t>(0ULL)];
int64_t h = gbamachine_objSizes[shape][shr<uint16_t>(attr1,cast<int64_t>(14ULL))][cast<int64_t>(1ULL)];
auto tmp77 = std::make_tuple(w,h);
int64_t bw = std::get<0>(tmp77);
int64_t bh = std::get<1>(tmp77);
if ((affine && (cast<uint16_t>((attr0 & cast<uint16_t>(512ULL))) != cast<uint16_t>(0ULL)))) {
auto tmp78 = std::make_tuple(cast<int64_t>((cast<int64_t>(2ULL) * w)),cast<int64_t>((cast<int64_t>(2ULL) * h)));
bw = std::get<0>(tmp78);
bh = std::get<1>(tmp78);
}
int64_t oy = cast<int64_t>(cast<uint16_t>((attr0 & cast<uint16_t>(255ULL))));
if ((cast<int64_t>((oy + bh)) > cast<int64_t>(256ULL))) {
oy -= cast<int64_t>(256ULL);
}
int64_t ox = cast<int64_t>(cast<uint16_t>((attr1 & cast<uint16_t>(511ULL))));
if ((ox >= cast<int64_t>(256ULL))) {
ox -= cast<int64_t>(512ULL);
}
int64_t row = cast<int64_t>((y - oy));
if (((row < cast<int64_t>(0ULL)) || (row >= bh))) {
continue;
}
bool eightBpp = (cast<uint16_t>((attr0 & cast<uint16_t>(8192ULL))) != cast<uint16_t>(0ULL));
uint32_t tile = cast<uint32_t>(cast<uint16_t>((attr2 & cast<uint16_t>(1023ULL))));
if ((tile < minTile)) {
continue;
}
uint8_t pr = cast<uint8_t>(cast<uint16_t>((shr<uint16_t>(attr2,cast<int64_t>(10ULL)) & cast<uint16_t>(3ULL))));
int64_t bank = cast<int64_t>(shr<uint16_t>(attr2,cast<int64_t>(12ULL)));
int32_t paf=cast<int32_t>(256ULL);
int32_t pbf=cast<int32_t>(0ULL);
int32_t pcf=cast<int32_t>(0ULL);
int32_t pdf=cast<int32_t>(256ULL);
if (affine) {
int64_t g = cast<int64_t>(cast<uint16_t>((shr<uint16_t>(attr1,cast<int64_t>(9ULL)) & cast<uint16_t>(31ULL))));
auto rd = [&](int64_t k)->int32_t{
int64_t a = cast<int64_t>((cast<int64_t>((cast<int64_t>((g * cast<int64_t>(32ULL))) + cast<int64_t>((k * cast<int64_t>(8ULL))))) + cast<int64_t>(6ULL)));
return cast<int32_t>(cast<int16_t>(cast<uint16_t>((cast<uint16_t>(m->oam[a]) | shl<uint16_t>(cast<uint16_t>(m->oam[cast<int64_t>((a + cast<int64_t>(1ULL)))]),cast<int64_t>(8ULL))))));
}
;
auto tmp79 = std::make_tuple(rd(cast<int64_t>(0ULL)),rd(cast<int64_t>(1ULL)),rd(cast<int64_t>(2ULL)),rd(cast<int64_t>(3ULL)));
paf = std::get<0>(tmp79);
pbf = std::get<1>(tmp79);
pcf = std::get<2>(tmp79);
pdf = std::get<3>(tmp79);
}
bool hflip = ((!affine) && (cast<uint16_t>((attr1 & cast<uint16_t>(4096ULL))) != cast<uint16_t>(0ULL)));
bool vflip = ((!affine) && (cast<uint16_t>((attr1 & cast<uint16_t>(8192ULL))) != cast<uint16_t>(0ULL)));
{int64_t bx = cast<int64_t>(0ULL);for (;(bx < bw);bx++){
int64_t x = cast<int64_t>((ox + bx));
if (((x < cast<int64_t>(0ULL)) || (x >= cast<int64_t>(240ULL)))) {
continue;
}
int64_t sx={};
int64_t sy={};
if (affine) {
auto tmp80 = std::make_tuple(cast<int32_t>(cast<int64_t>((bx - divi<int64_t>(bw,cast<int64_t>(2ULL))))),cast<int32_t>(cast<int64_t>((row - divi<int64_t>(bh,cast<int64_t>(2ULL))))));
int32_t dx = std::get<0>(tmp80);
int32_t dy = std::get<1>(tmp80);
int32_t fx = cast<int32_t>((cast<int32_t>((cast<int32_t>((paf * dx)) + cast<int32_t>((pbf * dy)))) + shl<int32_t>(cast<int32_t>(divi<int64_t>(w,cast<int64_t>(2ULL))),cast<int64_t>(8ULL))));
int32_t fy = cast<int32_t>((cast<int32_t>((cast<int32_t>((pcf * dx)) + cast<int32_t>((pdf * dy)))) + shl<int32_t>(cast<int32_t>(divi<int64_t>(h,cast<int64_t>(2ULL))),cast<int64_t>(8ULL))));
auto tmp81 = std::make_tuple(cast<int64_t>(shr<int32_t>(fx,cast<int64_t>(8ULL))),cast<int64_t>(shr<int32_t>(fy,cast<int64_t>(8ULL))));
sx = std::get<0>(tmp81);
sy = std::get<1>(tmp81);
if (((((sx < cast<int64_t>(0ULL)) || (sy < cast<int64_t>(0ULL))) || (sx >= w)) || (sy >= h))) {
continue;
}
}
else {
auto tmp82 = std::make_tuple(bx,row);
sx = std::get<0>(tmp82);
sy = std::get<1>(tmp82);
if (hflip) {
sx = cast<int64_t>((cast<int64_t>((w - cast<int64_t>(1ULL))) - sx));
}
if (vflip) {
sy = cast<int64_t>((cast<int64_t>((h - cast<int64_t>(1ULL))) - sy));
}
}
auto tmp83 = std::make_tuple(cast<uint32_t>(shr<int64_t>(sx,cast<int64_t>(3ULL))),cast<uint32_t>(shr<int64_t>(sy,cast<int64_t>(3ULL))));
uint32_t tileX = std::get<0>(tmp83);
uint32_t tileY = std::get<1>(tmp83);
uint32_t t={};
if (eightBpp) {
if (oneDim) {
t = cast<uint32_t>((cast<uint32_t>((tile + cast<uint32_t>((cast<uint32_t>((tileY * cast<uint32_t>(shr<int64_t>(w,cast<int64_t>(3ULL))))) * cast<uint32_t>(2ULL))))) + cast<uint32_t>((tileX * cast<uint32_t>(2ULL)))));
}
else {
t = cast<uint32_t>((cast<uint32_t>((((tile & ~(cast<uint32_t>(1ULL)))) + cast<uint32_t>((tileY * cast<uint32_t>(32ULL))))) + cast<uint32_t>((tileX * cast<uint32_t>(2ULL)))));
}
uint32_t a = cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(65536ULL) + cast<uint32_t>((t * cast<uint32_t>(32ULL))))) + cast<uint32_t>((cast<uint32_t>(cast<int64_t>((sy & cast<int64_t>(7ULL)))) * cast<uint32_t>(8ULL))))) + cast<uint32_t>(cast<int64_t>((sx & cast<int64_t>(7ULL))))));
if ((a >= cast<uint32_t>(98304ULL))) {
continue;
}
if(rrcapture::trace.active){rrgba::currentObject=i;rrgba::objectSource=a;rrgba::objectU=sx;rrgba::objectV=sy;}
gbamachine_ppu_objPixel(p,m,x,cast<int64_t>(m->vram[a]),cast<int64_t>(16ULL),pr,mode,c,on,prio,semi,win);
}
else {
if (oneDim) {
t = cast<uint32_t>((cast<uint32_t>((tile + cast<uint32_t>((tileY * cast<uint32_t>(shr<int64_t>(w,cast<int64_t>(3ULL))))))) + tileX));
}
else {
t = cast<uint32_t>((cast<uint32_t>((tile + cast<uint32_t>((tileY * cast<uint32_t>(32ULL))))) + tileX));
}
uint32_t a = cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(65536ULL) + cast<uint32_t>((t * cast<uint32_t>(32ULL))))) + cast<uint32_t>((cast<uint32_t>(cast<int64_t>((sy & cast<int64_t>(7ULL)))) * cast<uint32_t>(4ULL))))) + cast<uint32_t>(cast<int64_t>((shr<int64_t>(sx,cast<int64_t>(1ULL)) & cast<int64_t>(3ULL))))));
if ((a >= cast<uint32_t>(98304ULL))) {
continue;
}
uint8_t v = m->vram[a];
int64_t idx = cast<int64_t>(cast<uint8_t>((v & cast<uint8_t>(15ULL))));
if ((cast<int64_t>((sx & cast<int64_t>(1ULL))) == cast<int64_t>(1ULL))) {
idx = cast<int64_t>(shr<uint8_t>(v,cast<int64_t>(4ULL)));
}
if(rrcapture::trace.active){rrgba::currentObject=i;rrgba::objectSource=a;rrgba::objectU=sx;rrgba::objectV=sy;}
gbamachine_ppu_objPixel(p,m,x,idx,cast<int64_t>((cast<int64_t>(16ULL) + bank)),pr,mode,c,on,prio,semi,win);
}
}
}}
}}
}
// tools/platform/gba/gbamachine/ppu.go:570:1
void gbamachine_ppu_objPixel(gbamachine_ppu* p,gbamachine_Machine* m,int64_t x,int64_t idx,int64_t palBank,uint8_t pr,uint16_t mode,Slice<uint16_t> c,Slice<bool> on,Slice<uint8_t> prio,Slice<bool> semi,Slice<bool> win){
{
if ((idx == cast<int64_t>(0ULL))) {
return ;
}
if ((mode == cast<uint16_t>(2ULL))) {
win[x] = true;
return ;
}
if (on[x]) {
return ;
}
on[x] = true;
c[x] = gbamachine_Machine_pal16(m,palBank,idx);
if(rrcapture::trace.active){rrgba::event(m,"PPU object texel",4,rrgba::currentY);rrgba::source(m,4,x,rrgba::objectSource,idx,palBank,rrgba::objectU,rrgba::objectV);rrgba::dot(m,4,x,c[x]);}
prio[x] = pr;
semi[x] = (mode == cast<uint16_t>(1ULL));
}
}
// tools/platform/gba/gbamachine/run.go:49:1
gbamachine_Result gbamachine_Machine_Run(gbamachine_Machine* m,uint64_t budget,Map<uint32_t,std::string> milestones){
{
return gbamachine_Machine_run(m,budget,milestones,cast<uint64_t>(0ULL));
}
}
// tools/platform/gba/gbamachine/run.go:54:1
gbamachine_Result gbamachine_Machine_RunFrames(gbamachine_Machine* m,uint64_t n,uint64_t budget){
{
return gbamachine_Machine_run(m,budget,{},cast<uint64_t>((m->vid.frames + n)));
}
}
// tools/platform/gba/gbamachine/run.go:58:1
gbamachine_Result gbamachine_Machine_run(gbamachine_Machine* m,uint64_t budget,Map<uint32_t,std::string> milestones,uint64_t untilFrame){
{rrprof::Scope timing(0,"ARM7 and scheduler");
{
gbamachine_Result res = gbamachine_Result{{},{},{},Map<uint32_t,uint64_t>{}};
if ((!m->visited)) {
m->visited = Map<uint32_t,bool>{};
}
uint64_t lastProgress = m->Steps;
uint64_t prevSig = gbamachine_Machine_progressSig(m);
int64_t prevPages = len(m->visited);
auto tmp84 = std::make_tuple(false,false,cast<uint32_t>(0ULL));
m->stop = std::get<0>(tmp84);
m->stopped = std::get<1>(tmp84);
m->stoppedPC = std::get<2>(tmp84);
uint64_t end = cast<uint64_t>((m->Steps + budget));
{;for (;(m->Steps < end);){
if (((untilFrame != cast<uint64_t>(0ULL)) && (m->vid.frames >= untilFrame))) {
res.Reason = go_fmt_Sprintf(std::string("reached frame %d",16),m->vid.frames);
break;
}
gbamachine_Machine_startLine(m);
if (m->stop) {
res.Reason = std::string("stopped",7);
break;
}
constexpr int64_t quantum=32ULL;
bool hb = false;
{int64_t spent = cast<int64_t>(0ULL);for (;(spent < cast<int64_t>(410ULL));spent += cast<int64_t>(32ULL)){
if (((!hb) && (divi<int64_t>(cast<int64_t>((spent * cast<int64_t>(308ULL))),cast<int64_t>(410ULL)) >= cast<int64_t>(240ULL)))) {
gbamachine_Machine_hblankNow(m);
hb = true;
}
gbamachine_Machine_deliver(m);
gbamachine_Machine_runQuantum(m,cast<int64_t>(32ULL),milestones,res.Milestone);
m->Steps += cast<uint64_t>(32ULL);
if ((m->cpu->Halted || m->stop)) {
break;
}
}
}if ((!hb)) {
gbamachine_Machine_hblankNow(m);
}
gbamachine_Machine_tickTimers(m,cast<int64_t>(1232ULL));
gbamachine_apu_mixCycles(m->apu,cast<int64_t>(1232ULL));
if (m->stop) {
res.Reason = std::string("stopped",7);
if (m->stopped) {
res.Reason = go_fmt_Sprintf(std::string("breakpoint at 0x%08X",20),m->stoppedPC);
}
break;
}
if (m->cpu->Halted) {
res.Reason = (std::string("halted: ",8) + m->cpu->HaltReason);
break;
}
uint64_t sig = gbamachine_Machine_progressSig(m);
if (((sig != prevSig) || (len(m->visited) != prevPages))) {
auto tmp85 = std::make_tuple(sig,len(m->visited));
prevSig = std::get<0>(tmp85);
prevPages = std::get<1>(tmp85);
lastProgress = m->Steps;
}
else if ((cast<uint64_t>((m->Steps - lastProgress)) > cast<uint64_t>(24000000ULL))) {
res.Reason = go_fmt_Sprintf(std::string("settled \342\200\224 CPU at 0x%08X (%s); no new code or interrupt traffic",64),m->cpu->R[cast<int64_t>(15ULL)],gbamachine_Machine_Parked(m));
break;
}
}
}if ((res.Reason == std::string("",0))) {
res.Reason = go_fmt_Sprintf(std::string("step budget (%d) reached",24),budget);
}
auto tmp86 = std::make_tuple(m->Steps,m->vid.frames);
res.Steps = std::get<0>(tmp86);
res.Frames = std::get<1>(tmp86);
return res;
}
}
}
// tools/platform/gba/gbamachine/run.go:144:1
void gbamachine_Machine_startLine(gbamachine_Machine* m){
{
m->vid.hblank = false;
m->vid.line++;
if ((m->vid.line >= cast<int64_t>(228ULL))) {
m->vid.line = cast<int64_t>(0ULL);
}
int64_t line = m->vid.line;
uint16_t dispstat = get(m->io,cast<uint32_t>(4ULL));
{
if ((line == cast<int64_t>(0ULL))){
gbamachine_ppu_startFrame(&(m->ppu),m);
}
else if ((line == cast<int64_t>(160ULL))){
m->vid.frames++;
if ((cast<uint16_t>((dispstat & cast<uint16_t>(8ULL))) != cast<uint16_t>(0ULL))) {
gbamachine_Machine_raise(m,cast<uint16_t>(1ULL));
}
gbamachine_Machine_dmaTrigger(m,cast<uint16_t>(1ULL));
if (bool(m->OnFrame)) {
m->OnFrame();
}
}
}
tmp87:;
if ((line < cast<int64_t>(160ULL))) {
gbamachine_ppu_renderLine(&(m->ppu),m,line);
gbamachine_ppu_stepAffine(&(m->ppu),m);
}
if (((line == cast<int64_t>(shr<uint16_t>(dispstat,cast<int64_t>(8ULL)))) && (cast<uint16_t>((dispstat & cast<uint16_t>(32ULL))) != cast<uint16_t>(0ULL)))) {
gbamachine_Machine_raise(m,cast<uint16_t>(4ULL));
}
}
}
// tools/platform/gba/gbamachine/run.go:177:1
void gbamachine_Machine_hblankNow(gbamachine_Machine* m){
{
m->vid.hblank = true;
if ((m->vid.line < cast<int64_t>(160ULL))) {
if ((cast<uint16_t>((get(m->io,cast<uint32_t>(4ULL)) & cast<uint16_t>(16ULL))) != cast<uint16_t>(0ULL))) {
gbamachine_Machine_raise(m,cast<uint16_t>(2ULL));
}
gbamachine_Machine_dmaTrigger(m,cast<uint16_t>(2ULL));
}
}
}
// tools/platform/gba/gbamachine/run.go:188:1
void gbamachine_Machine_raise(gbamachine_Machine* m,uint16_t mask){
{
m->if_ |= mask;
}
}
// tools/platform/gba/gbamachine/run.go:192:1
void gbamachine_Machine_deliver(gbamachine_Machine* m){
{
uint16_t pending = cast<uint16_t>((m->ie & m->if_));
if ((pending == cast<uint16_t>(0ULL))) {
return ;
}
if (m->waiting) {
m->waiting = false;
m->waitAny = false;
}
else if (m->cpu->IRQDisable) {
return ;
}
if ((!m->ime)) {
return ;
}
gbamachine_bus busStorage{m}; auto*b=&busStorage;
uint32_t handler = gbamachine_bus_r32(b,cast<uint32_t>(50364412ULL));
if ((handler == cast<uint32_t>(0ULL))) {
return ;
}
uint32_t ret = m->cpu->R[cast<int64_t>(15ULL)];
if (bool(m->OnIRQ)) {
m->OnIRQ(pending,handler,ret);
}
arm_CPU_Exception(m->cpu,cast<uint32_t>(18ULL),handler,cast<uint32_t>((ret + cast<uint32_t>(4ULL))));
uint32_t sp = cast<uint32_t>((m->cpu->R[cast<int64_t>(13ULL)] - cast<uint32_t>(24ULL)));
m->cpu->R[cast<int64_t>(13ULL)] = sp;
gbamachine_bus_w32(b,cast<uint32_t>((sp + cast<uint32_t>(0ULL))),m->cpu->R[cast<int64_t>(0ULL)]);
gbamachine_bus_w32(b,cast<uint32_t>((sp + cast<uint32_t>(4ULL))),m->cpu->R[cast<int64_t>(1ULL)]);
gbamachine_bus_w32(b,cast<uint32_t>((sp + cast<uint32_t>(8ULL))),m->cpu->R[cast<int64_t>(2ULL)]);
gbamachine_bus_w32(b,cast<uint32_t>((sp + cast<uint32_t>(12ULL))),m->cpu->R[cast<int64_t>(3ULL)]);
gbamachine_bus_w32(b,cast<uint32_t>((sp + cast<uint32_t>(16ULL))),m->cpu->R[cast<int64_t>(12ULL)]);
gbamachine_bus_w32(b,cast<uint32_t>((sp + cast<uint32_t>(20ULL))),m->cpu->R[cast<int64_t>(14ULL)]);
m->cpu->R[cast<int64_t>(0ULL)] = cast<uint32_t>(67108864ULL);
m->cpu->R[cast<int64_t>(14ULL)] = cast<uint32_t>(4294905856ULL);
}
}
// tools/platform/gba/gbamachine/run.go:238:1
void gbamachine_Machine_biosIRQExit(gbamachine_Machine* m){
{
gbamachine_bus busStorage{m}; auto*b=&busStorage;
uint32_t sp = m->cpu->R[cast<int64_t>(13ULL)];
m->cpu->R[cast<int64_t>(0ULL)] = gbamachine_bus_r32(b,cast<uint32_t>((sp + cast<uint32_t>(0ULL))));
m->cpu->R[cast<int64_t>(1ULL)] = gbamachine_bus_r32(b,cast<uint32_t>((sp + cast<uint32_t>(4ULL))));
m->cpu->R[cast<int64_t>(2ULL)] = gbamachine_bus_r32(b,cast<uint32_t>((sp + cast<uint32_t>(8ULL))));
m->cpu->R[cast<int64_t>(3ULL)] = gbamachine_bus_r32(b,cast<uint32_t>((sp + cast<uint32_t>(12ULL))));
m->cpu->R[cast<int64_t>(12ULL)] = gbamachine_bus_r32(b,cast<uint32_t>((sp + cast<uint32_t>(16ULL))));
uint32_t lr = gbamachine_bus_r32(b,cast<uint32_t>((sp + cast<uint32_t>(20ULL))));
m->cpu->R[cast<int64_t>(13ULL)] = cast<uint32_t>((sp + cast<uint32_t>(24ULL)));
uint32_t spsr = arm_CPU_SPSR(m->cpu);
arm_CPU_SetCPSR(m->cpu,spsr);
m->cpu->R[cast<int64_t>(15ULL)] = cast<uint32_t>((lr - cast<uint32_t>(4ULL)));
if ((m->waitMask != cast<uint16_t>(0ULL))) {
uint16_t flags = cast<uint16_t>(gbamachine_bus_r32(b,cast<uint32_t>(50364408ULL)));
if ((cast<uint16_t>((flags & m->waitMask)) != cast<uint16_t>(0ULL))) {
gbamachine_bus_w32(b,cast<uint32_t>(50364408ULL),cast<uint32_t>((flags & ~(m->waitMask))));
m->waitMask = cast<uint16_t>(0ULL);
}
else {
m->waiting = true;
}
}
}
}
// tools/platform/gba/gbamachine/run.go:268:1
void gbamachine_Machine_runQuantum(gbamachine_Machine* m,int64_t n,Map<uint32_t,std::string> milestones,Map<uint32_t,uint64_t> hit){
{
if (m->waiting) {
return ;
}
{int64_t i = cast<int64_t>(0ULL);for (;(i < n);i++){
if (((m->waiting || m->cpu->Halted) || m->stop)) {
return ;
}
if ((m->cpu->R[cast<int64_t>(15ULL)] == cast<uint32_t>(4294905856ULL))) {
gbamachine_Machine_biosIRQExit(m);
continue;
}
uint32_t pc = m->cpu->R[cast<int64_t>(15ULL)];
if (bool(m->OnStep)) {
m->OnStep(pc);
}
if (get(m->bps,pc)) {
auto tmp88 = std::make_tuple(true,true,pc);
m->stop = std::get<0>(tmp88);
m->stopped = std::get<1>(tmp88);
m->stoppedPC = std::get<2>(tmp88);
return ;
}
m->visited[shr<uint32_t>(pc,cast<int64_t>(8ULL))] = true;
if (bool(milestones)) {
{
auto tmp89 = lookup(milestones,pc);
bool ok = std::get<1>(tmp89);
if (ok) {
{
auto tmp90 = lookup(hit,pc);
bool seen = std::get<1>(tmp90);
if ((!seen)) {
hit[pc] = cast<uint64_t>((m->Steps + cast<uint64_t>(i)));
}
}
}
}
}
arm_CPU_Step(m->cpu);
}
}}
}
// tools/platform/gba/gbamachine/run.go:303:1
uint64_t gbamachine_Machine_progressSig(gbamachine_Machine* m){
{
return cast<uint64_t>((cast<uint64_t>((cast<uint64_t>((cast<uint64_t>(m->if_) ^ shl<uint64_t>(cast<uint64_t>(m->ie),cast<int64_t>(16ULL)))) ^ shl<uint64_t>(cast<uint64_t>(m->keys),cast<int64_t>(32ULL)))) ^ shl<uint64_t>(cast<uint64_t>(len(m->eeprom.inBits)),cast<int64_t>(48ULL))));
}
}
// tools/platform/gba/gbamachine/sound.go:116:1
void gbamachine_fifo_push(gbamachine_fifo* f,uint8_t b){
{
if ((len(f->q) < cast<int64_t>(32ULL))) {
f->q = append(f->q,Slice<int8_t>{cast<int8_t>(b)});
}
}
}
// tools/platform/gba/gbamachine/sound.go:125:1
void gbamachine_fifo_pop(gbamachine_fifo* f){
{
if ((len(f->q) > cast<int64_t>(0ULL))) {
f->cur = f->q[cast<int64_t>(0ULL)];
f->q = sub(f->q,cast<int64_t>(1ULL),len(f->q));
}
}
}
// tools/platform/gba/gbamachine/sound.go:132:1
void gbamachine_fifo_reset(gbamachine_fifo* f){
{
auto tmp91 = std::make_tuple(sub(f->q,0,cast<int64_t>(0ULL)),cast<int8_t>(0ULL));
f->q = std::get<0>(tmp91);
f->cur = std::get<1>(tmp91);
}
}
// tools/platform/gba/gbamachine/sound.go:159:1
gbamachine_apu* gbamachine_newAPU(){
{
gbamachine_apu* a = arenaNew(gbamachine_apu{});
a->ch1.hasSweep = true;
a->ch4.lfsr = cast<uint16_t>(32767ULL);
return a;
}
}
// tools/platform/gba/gbamachine/sound.go:170:1
bool gbamachine_Machine_soundWrite(gbamachine_Machine* m,uint32_t reg,uint16_t v){
{
gbamachine_apu* a = m->apu;
{
if ((reg == cast<uint32_t>(96ULL))){
a->ch1.swPeriod = cast<int64_t>(cast<uint16_t>((shr<uint16_t>(v,cast<int64_t>(4ULL)) & cast<uint16_t>(7ULL))));
a->ch1.swShift = cast<int64_t>(cast<uint16_t>((v & cast<uint16_t>(7ULL))));
a->ch1.swDir = cast<int64_t>(1ULL);
if ((cast<uint16_t>((v & cast<uint16_t>(8ULL))) != cast<uint16_t>(0ULL))) {
a->ch1.swDir = cast<int64_t>(-1ULL);
}
}
else if ((reg == cast<uint32_t>(98ULL))){
gbamachine_apu_writeDutyEnv(a,(&a->ch1),v);
}
else if ((reg == cast<uint32_t>(100ULL))){
gbamachine_apu_writeFreqCtl(a,(&a->ch1),v);
}
else if ((reg == cast<uint32_t>(104ULL))){
gbamachine_apu_writeDutyEnv(a,(&a->ch2),v);
}
else if ((reg == cast<uint32_t>(108ULL))){
gbamachine_apu_writeFreqCtl(a,(&a->ch2),v);
}
else if ((reg == cast<uint32_t>(112ULL))){
a->ch3.twoBanks = (cast<uint16_t>((v & cast<uint16_t>(32ULL))) != cast<uint16_t>(0ULL));
a->ch3.bank = cast<int64_t>(cast<uint16_t>((shr<uint16_t>(v,cast<int64_t>(6ULL)) & cast<uint16_t>(1ULL))));
a->ch3.dacOn = (cast<uint16_t>((v & cast<uint16_t>(128ULL))) != cast<uint16_t>(0ULL));
if ((!a->ch3.dacOn)) {
a->ch3.enabled = false;
}
}
else if ((reg == cast<uint32_t>(114ULL))){
a->ch3.length = cast<int64_t>((cast<int64_t>(256ULL) - cast<int64_t>(cast<uint16_t>((v & cast<uint16_t>(255ULL))))));
a->ch3.volShift = cast<int64_t>(cast<uint16_t>((shr<uint16_t>(v,cast<int64_t>(13ULL)) & cast<uint16_t>(3ULL))));
a->ch3.force75 = (cast<uint16_t>((v & cast<uint16_t>(32768ULL))) != cast<uint16_t>(0ULL));
}
else if ((reg == cast<uint32_t>(116ULL))){
a->ch3.freq = cast<int64_t>(cast<uint16_t>((v & cast<uint16_t>(2047ULL))));
a->ch3.lengthEn = (cast<uint16_t>((v & cast<uint16_t>(16384ULL))) != cast<uint16_t>(0ULL));
if (((cast<uint16_t>((v & cast<uint16_t>(32768ULL))) != cast<uint16_t>(0ULL)) && a->ch3.dacOn)) {
a->ch3.enabled = true;
a->ch3.phase = cast<double>(0ULL);
if ((a->ch3.length == cast<int64_t>(0ULL))) {
a->ch3.length = cast<int64_t>(256ULL);
}
}
}
else if ((reg == cast<uint32_t>(120ULL))){
a->ch4.length = cast<int64_t>((cast<int64_t>(64ULL) - cast<int64_t>(cast<uint16_t>((v & cast<uint16_t>(63ULL))))));
a->ch4.envPeriod = cast<int64_t>(cast<uint16_t>((shr<uint16_t>(v,cast<int64_t>(8ULL)) & cast<uint16_t>(7ULL))));
a->ch4.vol = cast<int64_t>(cast<uint16_t>((shr<uint16_t>(v,cast<int64_t>(12ULL)) & cast<uint16_t>(15ULL))));
a->ch4.envDir = cast<int64_t>(-1ULL);
if ((cast<uint16_t>((v & cast<uint16_t>(2048ULL))) != cast<uint16_t>(0ULL))) {
a->ch4.envDir = cast<int64_t>(1ULL);
}
if ((cast<uint16_t>((v & cast<uint16_t>(63488ULL))) == cast<uint16_t>(0ULL))) {
a->ch4.enabled = false;
}
}
else if ((reg == cast<uint32_t>(124ULL))){
a->ch4.divisor = gbamachine_noiseDivisors[cast<uint16_t>((v & cast<uint16_t>(7ULL)))];
a->ch4.width7 = (cast<uint16_t>((v & cast<uint16_t>(8ULL))) != cast<uint16_t>(0ULL));
a->ch4.shift = cast<int64_t>(cast<uint16_t>((shr<uint16_t>(v,cast<int64_t>(4ULL)) & cast<uint16_t>(15ULL))));
a->ch4.lengthEn = (cast<uint16_t>((v & cast<uint16_t>(16384ULL))) != cast<uint16_t>(0ULL));
if ((cast<uint16_t>((v & cast<uint16_t>(32768ULL))) != cast<uint16_t>(0ULL))) {
a->ch4.enabled = true;
a->ch4.lfsr = cast<uint16_t>(32767ULL);
a->ch4.envTimer = a->ch4.envPeriod;
if ((a->ch4.length == cast<int64_t>(0ULL))) {
a->ch4.length = cast<int64_t>(64ULL);
}
}
}
else if ((reg == cast<uint32_t>(128ULL))){
a->mixRegL = v;
}
else if ((reg == cast<uint32_t>(130ULL))){
a->mixRegH = v;
if ((cast<uint16_t>((v & cast<uint16_t>(2048ULL))) != cast<uint16_t>(0ULL))) {
gbamachine_fifo_reset(&(a->dsA));
}
if ((cast<uint16_t>((v & cast<uint16_t>(32768ULL))) != cast<uint16_t>(0ULL))) {
gbamachine_fifo_reset(&(a->dsB));
}
}
else if ((reg == cast<uint32_t>(132ULL))){
a->powered = (cast<uint16_t>((v & cast<uint16_t>(128ULL))) != cast<uint16_t>(0ULL));
if ((!a->powered)) {
auto tmp93 = std::make_tuple(false,false);
a->ch1.enabled = std::get<0>(tmp93);
a->ch2.enabled = std::get<1>(tmp93);
auto tmp94 = std::make_tuple(false,false);
a->ch3.enabled = std::get<0>(tmp94);
a->ch4.enabled = std::get<1>(tmp94);
}
}
else if ((reg == cast<uint32_t>(136ULL))){
}
else if (((reg >= cast<uint32_t>(144ULL)) && (reg <= cast<uint32_t>(158ULL)))){
uint32_t i = cast<uint32_t>(((cast<uint32_t>((reg - cast<uint32_t>(144ULL)))) & cast<uint32_t>(15ULL)));
int64_t w = cast<int64_t>((cast<int64_t>(1ULL) - a->ch3.bank));
a->ch3.ram[w][i] = cast<uint8_t>(v);
a->ch3.ram[w][cast<uint32_t>((i + cast<uint32_t>(1ULL)))] = cast<uint8_t>(shr<uint16_t>(v,cast<int64_t>(8ULL)));
}
else if (((reg == cast<uint32_t>(160ULL)) || (reg == cast<uint32_t>(162ULL)))){
gbamachine_fifo_push(&(a->dsA),cast<uint8_t>(v));
gbamachine_fifo_push(&(a->dsA),cast<uint8_t>(shr<uint16_t>(v,cast<int64_t>(8ULL))));
}
else if (((reg == cast<uint32_t>(164ULL)) || (reg == cast<uint32_t>(166ULL)))){
gbamachine_fifo_push(&(a->dsB),cast<uint8_t>(v));
gbamachine_fifo_push(&(a->dsB),cast<uint8_t>(shr<uint16_t>(v,cast<int64_t>(8ULL))));
}
else {
return false;
}
}
tmp92:;
return true;
}
}
// tools/platform/gba/gbamachine/sound.go:273:1
void gbamachine_apu_writeDutyEnv(gbamachine_apu* a,gbamachine_square* c,uint16_t v){
{
c->length = cast<int64_t>((cast<int64_t>(64ULL) - cast<int64_t>(cast<uint16_t>((v & cast<uint16_t>(63ULL))))));
c->dutySel = cast<int64_t>(cast<uint16_t>((shr<uint16_t>(v,cast<int64_t>(6ULL)) & cast<uint16_t>(3ULL))));
c->envPeriod = cast<int64_t>(cast<uint16_t>((shr<uint16_t>(v,cast<int64_t>(8ULL)) & cast<uint16_t>(7ULL))));
c->vol = cast<int64_t>(cast<uint16_t>((shr<uint16_t>(v,cast<int64_t>(12ULL)) & cast<uint16_t>(15ULL))));
c->envDir = cast<int64_t>(-1ULL);
if ((cast<uint16_t>((v & cast<uint16_t>(2048ULL))) != cast<uint16_t>(0ULL))) {
c->envDir = cast<int64_t>(1ULL);
}
if ((cast<uint16_t>((v & cast<uint16_t>(63488ULL))) == cast<uint16_t>(0ULL))) {
c->enabled = false;
}
}
}
// tools/platform/gba/gbamachine/sound.go:287:1
void gbamachine_apu_writeFreqCtl(gbamachine_apu* a,gbamachine_square* c,uint16_t v){
{
c->freq = cast<int64_t>(cast<uint16_t>((v & cast<uint16_t>(2047ULL))));
c->lengthEn = (cast<uint16_t>((v & cast<uint16_t>(16384ULL))) != cast<uint16_t>(0ULL));
if ((cast<uint16_t>((v & cast<uint16_t>(32768ULL))) != cast<uint16_t>(0ULL))) {
c->enabled = true;
c->envTimer = c->envPeriod;
if ((c->length == cast<int64_t>(0ULL))) {
c->length = cast<int64_t>(64ULL);
}
if (c->hasSweep) {
c->swShadow = c->freq;
c->swTimer = c->swPeriod;
c->swOn = ((c->swPeriod > cast<int64_t>(0ULL)) || (c->swShift > cast<int64_t>(0ULL)));
}
}
}
}
// tools/platform/gba/gbamachine/sound.go:306:1
std::tuple<uint16_t,bool> gbamachine_Machine_soundRead(gbamachine_Machine* m,uint32_t reg){
{
gbamachine_apu* a = m->apu;
{
if ((reg == cast<uint32_t>(132ULL))){
uint16_t v={};
if (a->powered) {
v |= cast<uint16_t>(128ULL);
}
{auto&& tmp96 = Slice<bool>{a->ch1.enabled,a->ch2.enabled,a->ch3.enabled,a->ch4.enabled};
for(int64_t tmp97=0;tmp97<len(tmp96);++tmp97){
auto i=tmp97;auto on=tmp96[tmp97];if (on) {
v |= shl<uint16_t>(cast<uint16_t>(1ULL),cast<uint64_t>(i));
}
}}
return {v,true};
}
else if (((reg >= cast<uint32_t>(144ULL)) && (reg <= cast<uint32_t>(158ULL)))){
uint32_t i = cast<uint32_t>(((cast<uint32_t>((reg - cast<uint32_t>(144ULL)))) & cast<uint32_t>(15ULL)));
int64_t w = cast<int64_t>((cast<int64_t>(1ULL) - a->ch3.bank));
return {cast<uint16_t>((cast<uint16_t>(a->ch3.ram[w][i]) | shl<uint16_t>(cast<uint16_t>(a->ch3.ram[w][cast<uint32_t>((i + cast<uint32_t>(1ULL)))]),cast<int64_t>(8ULL)))),true};
}
}
tmp95:;
return {cast<uint16_t>(0ULL),false};
}
}
// tools/platform/gba/gbamachine/sound.go:333:1
void gbamachine_Machine_fifoTimerOverflow(gbamachine_Machine* m,int64_t n,int64_t times){
{
uint16_t h = get(m->io,cast<uint32_t>(130ULL));
{auto&& tmp98 = Slice<Anon21>{Anon21{(&m->apu->dsA),cast<int64_t>(cast<uint16_t>((shr<uint16_t>(h,cast<int64_t>(10ULL)) & cast<uint16_t>(1ULL)))),cast<int64_t>(1ULL)},Anon21{(&m->apu->dsB),cast<int64_t>(cast<uint16_t>((shr<uint16_t>(h,cast<int64_t>(14ULL)) & cast<uint16_t>(1ULL)))),cast<int64_t>(2ULL)}};
for(int64_t tmp99=0;tmp99<len(tmp98);++tmp99){
auto ch=tmp98[tmp99];if ((ch.timerSel != n)) {
continue;
}
{int64_t i = cast<int64_t>(0ULL);for (;(i < times);i++){
gbamachine_fifo_pop(ch.f);
}
}if ((len(ch.f->q) <= cast<int64_t>(16ULL))) {
gbamachine_Machine_dmaSoundRefill(m,ch.dmaCh);
}
}}
}
}
// tools/platform/gba/gbamachine/sound.go:360:1
void gbamachine_apu_mixCycles(gbamachine_apu* a,int64_t cycles){
{rrprof::Scope timing(3,"Audio synthesis");
{
if ((!a->Capture)) {
return ;
}
a->sampleAcc += ((cast<double>(cycles) * gbamachine_audioRate) / gbamachine_cpuHz);
int64_t n = cast<int64_t>(a->sampleAcc);
a->sampleAcc -= cast<double>(n);
{int64_t i = cast<int64_t>(0ULL);for (;(i < n);i++){
gbamachine_apu_mixOne(a);
}
}}
}
}
// tools/platform/gba/gbamachine/sound.go:376:1
void gbamachine_apu_mixOne(gbamachine_apu* a){
{
constexpr double dt=3.05175781250000000e-05;
a->frameAcc += (gbamachine_cpuHz / cast<double>(32768ULL));
{;for (;(a->frameAcc >= (gbamachine_cpuHz / cast<double>(512ULL)));){
a->frameAcc -= (gbamachine_cpuHz / cast<double>(512ULL));
gbamachine_apu_stepFrame(a,a->frameSeq);
a->frameSeq = cast<int64_t>(((cast<int64_t>((a->frameSeq + cast<int64_t>(1ULL)))) & cast<int64_t>(7ULL)));
}
}double s1 = gbamachine_square_sample(&(a->ch1),dt);
double s2 = gbamachine_square_sample(&(a->ch2),dt);
double s3 = gbamachine_waveCh_sample(&(a->ch3),dt);
double s4 = gbamachine_noiseCh_sample(&(a->ch4),dt);
auto tmp100 = std::make_tuple(a->mixRegL,a->mixRegH);
uint16_t cntL = std::get<0>(tmp100);
uint16_t cntH = std::get<1>(tmp100);
double l={};
double r={};
auto pan = [&](double s,int64_t bit)->void{
if ((cast<uint16_t>((cntL & (shl<uint16_t>(cast<uint16_t>(1ULL),cast<uint64_t>(cast<int64_t>((cast<int64_t>(8ULL) + bit))))))) != cast<uint16_t>(0ULL))) {
r += s;
}
if ((cast<uint16_t>((cntL & (shl<uint16_t>(cast<uint16_t>(1ULL),cast<uint64_t>(cast<int64_t>((cast<int64_t>(12ULL) + bit))))))) != cast<uint16_t>(0ULL))) {
l += s;
}
}
;
pan(s1,cast<int64_t>(0ULL));
pan(s2,cast<int64_t>(1ULL));
pan(s3,cast<int64_t>(2ULL));
pan(s4,cast<int64_t>(3ULL));
r *= (cast<double>(cast<uint16_t>((cast<uint16_t>((cntL & cast<uint16_t>(7ULL))) + cast<uint16_t>(1ULL)))) / cast<double>(8ULL));
l *= (cast<double>(cast<uint16_t>((cast<uint16_t>((shr<uint16_t>(cntL,cast<int64_t>(4ULL)) & cast<uint16_t>(7ULL))) + cast<uint16_t>(1ULL)))) / cast<double>(8ULL));
double psg = gbamachine_psgVolTable[cast<uint16_t>((cntH & cast<uint16_t>(3ULL)))];
l *= psg;
r *= psg;
auto dsMix = [&](gbamachine_fifo* f,int64_t volBit,int64_t rBit,int64_t lBit)->void{
double v = (cast<double>(f->cur) / cast<double>(128ULL));
if ((cast<uint16_t>((cntH & (shl<uint16_t>(cast<uint16_t>(1ULL),cast<uint64_t>(volBit))))) == cast<uint16_t>(0ULL))) {
v *= 0.5;
}
if ((cast<uint16_t>((cntH & (shl<uint16_t>(cast<uint16_t>(1ULL),cast<uint64_t>(rBit))))) != cast<uint16_t>(0ULL))) {
r += v;
}
if ((cast<uint16_t>((cntH & (shl<uint16_t>(cast<uint16_t>(1ULL),cast<uint64_t>(lBit))))) != cast<uint16_t>(0ULL))) {
l += v;
}
}
;
dsMix((&a->dsA),cast<int64_t>(2ULL),cast<int64_t>(8ULL),cast<int64_t>(9ULL));
dsMix((&a->dsB),cast<int64_t>(3ULL),cast<int64_t>(12ULL),cast<int64_t>(13ULL));
auto clamp = [&](double v)->int16_t{
v *= 0.35;
if ((v > cast<double>(1ULL))) {
v = cast<double>(1ULL);
}
else if ((v < cast<double>(-cast<int64_t>(1ULL)))) {
v = cast<double>(-cast<int64_t>(1ULL));
}
return cast<int16_t>((v * cast<double>(32000ULL)));
}
;
a->PCM = append(a->PCM,Slice<int16_t>{clamp(l),clamp(r)});
}
}
// tools/platform/gba/gbamachine/sound.go:442:1
void gbamachine_apu_stepFrame(gbamachine_apu* a,int64_t step){
{
if ((modi<int64_t>(step,cast<int64_t>(2ULL)) == cast<int64_t>(0ULL))) {
gbamachine_apu_clockLength(a);
}
if (((step == cast<int64_t>(2ULL)) || (step == cast<int64_t>(6ULL)))) {
gbamachine_apu_clockSweep(a);
}
if ((step == cast<int64_t>(7ULL))) {
gbamachine_apu_clockEnv(a);
}
}
}
// tools/platform/gba/gbamachine/sound.go:454:1
void gbamachine_apu_clockLength(gbamachine_apu* a){
{
auto dec = [&](bool* en,int64_t* length,bool lengthEn)->void{
if ((lengthEn && ((*length) > cast<int64_t>(0ULL)))) {
(*length)--;
if (((*length) == cast<int64_t>(0ULL))) {
(*en) = false;
}
}
}
;
dec((&a->ch1.enabled),(&a->ch1.length),a->ch1.lengthEn);
dec((&a->ch2.enabled),(&a->ch2.length),a->ch2.lengthEn);
dec((&a->ch3.enabled),(&a->ch3.length),a->ch3.lengthEn);
dec((&a->ch4.enabled),(&a->ch4.length),a->ch4.lengthEn);
}
}
// tools/platform/gba/gbamachine/sound.go:469:1
void gbamachine_apu_clockSweep(gbamachine_apu* a){
{
gbamachine_square* c = (&a->ch1);
if (((!c->swOn) || (c->swPeriod == cast<int64_t>(0ULL)))) {
return ;
}
{
c->swTimer--;
if ((c->swTimer > cast<int64_t>(0ULL))) {
return ;
}
}
c->swTimer = c->swPeriod;
int64_t next = cast<int64_t>((c->swShadow + cast<int64_t>((c->swDir * (shr<int64_t>(c->swShadow,cast<uint64_t>(c->swShift)))))));
if ((next > cast<int64_t>(2047ULL))) {
c->enabled = false;
return ;
}
if ((next < cast<int64_t>(0ULL))) {
return ;
}
auto tmp101 = std::make_tuple(next,next);
c->swShadow = std::get<0>(tmp101);
c->freq = std::get<1>(tmp101);
}
}
// tools/platform/gba/gbamachine/sound.go:489:1
void gbamachine_apu_clockEnv(gbamachine_apu* a){
{
auto step = [&](int64_t* vol,int64_t dir,int64_t period,int64_t* timer)->void{
if ((period == cast<int64_t>(0ULL))) {
return ;
}
{
(*timer)--;
if (((*timer) > cast<int64_t>(0ULL))) {
return ;
}
}
(*timer) = period;
int64_t v = cast<int64_t>(((*vol) + dir));
if (((v >= cast<int64_t>(0ULL)) && (v <= cast<int64_t>(15ULL)))) {
(*vol) = v;
}
}
;
step((&a->ch1.vol),a->ch1.envDir,a->ch1.envPeriod,(&a->ch1.envTimer));
step((&a->ch2.vol),a->ch2.envDir,a->ch2.envPeriod,(&a->ch2.envTimer));
step((&a->ch4.vol),a->ch4.envDir,a->ch4.envPeriod,(&a->ch4.envTimer));
}
}
// tools/platform/gba/gbamachine/sound.go:510:1
double gbamachine_square_sample(gbamachine_square* c,double dt){
{
if (((!c->enabled) || (c->freq >= cast<int64_t>(2048ULL)))) {
return cast<double>(0ULL);
}
double hz = (131072.0 / cast<double>(cast<int64_t>((cast<int64_t>(2048ULL) - c->freq))));
c->phase += (hz * dt);
c->phase -= cast<double>(cast<int64_t>(c->phase));
int64_t pos = cast<int64_t>((c->phase * cast<double>(8ULL)));
double v = (cast<double>(c->vol) / cast<double>(15ULL));
if ((gbamachine_dutyTable[c->dutySel][cast<int64_t>((pos & cast<int64_t>(7ULL)))] == cast<uint8_t>(0ULL))) {
return cast<double>(-v);
}
return v;
}
}
// tools/platform/gba/gbamachine/sound.go:525:1
double gbamachine_waveCh_sample(gbamachine_waveCh* c,double dt){
{
if ((((!c->enabled) || (!c->dacOn)) || (c->freq >= cast<int64_t>(2048ULL)))) {
return cast<double>(0ULL);
}
double hz = (2097152.0 / cast<double>(cast<int64_t>((cast<int64_t>(2048ULL) - c->freq))));
int64_t n = cast<int64_t>(32ULL);
if (c->twoBanks) {
n = cast<int64_t>(64ULL);
}
c->phase += ((hz / cast<double>(n)) * dt);
c->phase -= cast<double>(cast<int64_t>(c->phase));
int64_t pos = cast<int64_t>((c->phase * cast<double>(n)));
int64_t bank = c->bank;
if ((c->twoBanks && (pos >= cast<int64_t>(32ULL)))) {
auto tmp102 = std::make_tuple(cast<int64_t>((cast<int64_t>(1ULL) - bank)),cast<int64_t>((pos - cast<int64_t>(32ULL))));
bank = std::get<0>(tmp102);
pos = std::get<1>(tmp102);
}
uint8_t b = c->ram[cast<int64_t>((bank & cast<int64_t>(1ULL)))][cast<int64_t>(((shr<int64_t>(pos,cast<int64_t>(1ULL))) & cast<int64_t>(15ULL)))];
uint8_t nib = cast<uint8_t>((b & cast<uint8_t>(15ULL)));
if ((cast<int64_t>((pos & cast<int64_t>(1ULL))) == cast<int64_t>(0ULL))) {
nib = shr<uint8_t>(b,cast<int64_t>(4ULL));
}
double v={};
{
if (c->force75){
v = (cast<double>(nib) * 0.75);
}
else if ((c->volShift == cast<int64_t>(0ULL))){
return cast<double>(0ULL);
}
else {
v = cast<double>(shr<uint8_t>(nib,cast<uint64_t>(cast<int64_t>((c->volShift - cast<int64_t>(1ULL))))));
}
}
tmp103:;
return ((v / 7.5) - cast<double>(1ULL));
}
}
// tools/platform/gba/gbamachine/sound.go:558:1
double gbamachine_noiseCh_sample(gbamachine_noiseCh* c,double dt){
{
if (((!c->enabled) || (c->shift >= cast<int64_t>(14ULL)))) {
return cast<double>(0ULL);
}
double hz = ((524288.0 / cast<double>(c->divisor)) / cast<double>(shl<uint64_t>(cast<uint64_t>(1ULL),cast<uint64_t>(c->shift))));
c->timer += (hz * dt);
{;for (;(c->timer >= cast<double>(1ULL));){
c->timer--;
uint16_t bit = cast<uint16_t>(((cast<uint16_t>((c->lfsr ^ (shr<uint16_t>(c->lfsr,cast<int64_t>(1ULL)))))) & cast<uint16_t>(1ULL)));
c->lfsr = shr<uint16_t>(c->lfsr,cast<int64_t>(1ULL));
c->lfsr |= shl<uint16_t>(bit,cast<int64_t>(14ULL));
if (c->width7) {
c->lfsr = cast<uint16_t>(((c->lfsr & ~(cast<uint16_t>(64ULL))) | shl<uint16_t>(bit,cast<int64_t>(6ULL))));
}
}
}double v = (cast<double>(c->vol) / cast<double>(15ULL));
if ((cast<uint16_t>((c->lfsr & cast<uint16_t>(1ULL))) == cast<uint16_t>(0ULL))) {
return v;
}
return cast<double>(-v);
}
}
// tools/platform/gba/gbamachine/sound.go:584:1
void gbamachine_Machine_AudioCapture(gbamachine_Machine* m,bool on){
{
m->apu->Capture = on;
}
}
// tools/platform/gba/gbamachine/sound.go:587:1
int64_t gbamachine_Machine_AudioSamples(gbamachine_Machine* m){
{
return divi<int64_t>(len(m->apu->PCM),cast<int64_t>(2ULL));
}
}
// tools/platform/gba/gbamachine/sound.go:632:1
Slice<std::string> gbamachine_Machine_SoundState(gbamachine_Machine* m){
{
gbamachine_apu* a = m->apu;
Slice<std::string> out = Slice<std::string>{go_fmt_Sprintf(std::string("SOUNDCNT_L=%04X SOUNDCNT_H=%04X SOUNDCNT_X=%04X (powered=%v)",60),a->mixRegL,a->mixRegH,get(m->io,cast<uint32_t>(132ULL)),a->powered),go_fmt_Sprintf(std::string("PSG ch1 on=%v vol=%2d freq=%4d | ch2 on=%v vol=%2d freq=%4d | ch3 on=%v vol>>%d freq=%4d | ch4 on=%v vol=%2d",108),a->ch1.enabled,a->ch1.vol,a->ch1.freq,a->ch2.enabled,a->ch2.vol,a->ch2.freq,a->ch3.enabled,a->ch3.volShift,a->ch3.freq,a->ch4.enabled,a->ch4.vol)};
{auto&& tmp104 = Slice<Anon22>{Anon22{std::string("A",1),(&a->dsA),cast<int64_t>(cast<uint16_t>((shr<uint16_t>(a->mixRegH,cast<int64_t>(10ULL)) & cast<uint16_t>(1ULL)))),cast<int64_t>(1ULL)},Anon22{std::string("B",1),(&a->dsB),cast<int64_t>(cast<uint16_t>((shr<uint16_t>(a->mixRegH,cast<int64_t>(14ULL)) & cast<uint16_t>(1ULL)))),cast<int64_t>(2ULL)}};
for(int64_t tmp105=0;tmp105<len(tmp104);++tmp105){
auto i=tmp105;auto ds=tmp104[tmp105];gbamachine_timer* t = (&m->timers[ds.timerSel]);
gbamachine_dmaChan* d = (&m->dma[ds.dmaCh]);
double rate = 0.0;
if ((cast<uint16_t>((t->ctrl & cast<uint16_t>(128ULL))) != cast<uint16_t>(0ULL))) {
rate = ((gbamachine_cpuHz / cast<double>(gbamachine_timerPrescale[cast<uint16_t>((t->ctrl & cast<uint16_t>(3ULL)))])) / cast<double>(cast<int64_t>((cast<int64_t>(65536ULL) - cast<int64_t>(t->reload)))));
}
out = append(out,Slice<std::string>{go_fmt_Sprintf(std::string("DirectSound %s: fifo %2d/32 cur=%4d | timer%d ctrl=%04X reload=%04X -> %.0f Hz | DMA%d ctrl=%04X (timing %d, %s) src=%08X dst=%08X",130),ds.name,len(ds.f->q),ds.f->cur,ds.timerSel,t->ctrl,t->reload,rate,ds.dmaCh,d->ctrl,cast<uint16_t>((shr<uint16_t>(d->ctrl,cast<int64_t>(12ULL)) & cast<uint16_t>(3ULL))),gbamachine_enabledStr((cast<uint16_t>((d->ctrl & cast<uint16_t>(32768ULL))) != cast<uint16_t>(0ULL))),d->latchSrc,d->dst)});
(void)(i);
}}
return out;
}
}
// tools/platform/gba/gbamachine/sound.go:669:1
std::string gbamachine_enabledStr(bool on){
{
if (on) {
return std::string("enabled",7);
}
return std::string("DISABLED",8);
}
}
// tools/platform/gba/gbamachine/timer.go:20:1
void gbamachine_Machine_timerRegWrite(gbamachine_Machine* m,uint32_t reg,uint16_t v){
{
int64_t n = divi<int64_t>(cast<int64_t>(cast<uint32_t>((reg - cast<uint32_t>(256ULL)))),cast<int64_t>(4ULL));
gbamachine_timer* t = (&m->timers[n]);
if ((cast<uint32_t>((reg & cast<uint32_t>(2ULL))) == cast<uint32_t>(0ULL))) {
t->reload = v;
return ;
}
uint16_t was = t->ctrl;
t->ctrl = v;
if (((cast<uint16_t>((v & cast<uint16_t>(128ULL))) != cast<uint16_t>(0ULL)) && (cast<uint16_t>((was & cast<uint16_t>(128ULL))) == cast<uint16_t>(0ULL)))) {
t->counter = t->reload;
t->frac = cast<int64_t>(0ULL);
}
}
}
// tools/platform/gba/gbamachine/timer.go:36:1
void gbamachine_Machine_tickTimers(gbamachine_Machine* m,int64_t cycles){
{
bool overflowBelow = false;
{auto&& tmp106 = m->timers;
for(int64_t tmp107=0;tmp107<len(tmp106);++tmp107){
auto n=tmp107;gbamachine_timer* t = (&m->timers[n]);
if ((cast<uint16_t>((t->ctrl & cast<uint16_t>(128ULL))) == cast<uint16_t>(0ULL))) {
overflowBelow = false;
continue;
}
int64_t ticks = cast<int64_t>(0ULL);
if (((n > cast<int64_t>(0ULL)) && (cast<uint16_t>((t->ctrl & cast<uint16_t>(4ULL))) != cast<uint16_t>(0ULL)))) {
if (overflowBelow) {
ticks = cast<int64_t>(1ULL);
}
}
else {
t->frac += cycles;
int64_t p = gbamachine_timerPrescale[cast<uint16_t>((t->ctrl & cast<uint16_t>(3ULL)))];
ticks = divi<int64_t>(t->frac,p);
t->frac %= p;
}
overflowBelow = false;
int64_t overflows = cast<int64_t>(0ULL);
{;for (;(ticks > cast<int64_t>(0ULL));ticks--){
t->counter++;
if ((t->counter == cast<uint16_t>(0ULL))) {
t->counter = t->reload;
overflowBelow = true;
overflows++;
if ((cast<uint16_t>((t->ctrl & cast<uint16_t>(64ULL))) != cast<uint16_t>(0ULL))) {
gbamachine_Machine_raise(m,shl<uint16_t>(cast<uint16_t>(8ULL),cast<uint64_t>(n)));
}
}
}
}if (((overflows > cast<int64_t>(0ULL)) && (n < cast<int64_t>(2ULL)))) {
gbamachine_Machine_fifoTimerOverflow(m,n,overflows);
}
}}
}
}

#include "fast.h"
