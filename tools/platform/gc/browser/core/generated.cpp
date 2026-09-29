#include "runtime.h"
struct gekko_LockedCache;
struct gekko_FPR;
struct gekko_CPU;
struct gekko_Inst;
struct gekko_State;
struct gekko_Case;
struct gcdsp_CPU;
struct gcdsp_LoopFrame;
struct gcdsp_extPending;
struct gc_mi;
struct gc_ai;
struct gc_cp;
struct gc_pe;
struct gc_wgPipe;
struct gc_di;
struct gc_Header;
struct gc_Apploader;
struct gc_Disc;
struct gc_Segment;
struct gc_DOL;
struct gc_dsp;
struct gc_dspBus;
struct gc_exiChannel;
struct gc_exi;
struct gc_File;
struct gc_FST;
struct gc_gpu;
struct gc_clipPlane;
struct gc_attrLayout;
struct gc_texAttr;
struct gc_rasterTri;
struct gc_rstats;
struct gc_tevStage;
struct gc_tevState;
struct gc_texCoord;
struct gc_texState;
struct gc_screenVertex;
struct gc_clipVertex;
struct gc_idleSnap;
struct gc_idleState;
struct gc_Machine;
struct gc_PixelEvent;
struct gc_pi;
struct gc_ProfileBucket;
struct gc_ProfileCounter;
struct gc_FrameProfile;
struct gc_profState;
struct gc_profCounters;
struct gc_Result;
struct gc_runState;
struct gc_si;
struct gc_padPort;
struct gc_StackSample;
struct gc_stackProf;
struct gc_vi;
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
struct Anon24;
struct Anon25;
struct Anon26;
struct Anon27;
struct Anon28;
struct Anon29;
struct Anon3;
struct Anon30;
struct Anon31;
struct Anon32;
struct Anon33;
struct Anon34;
struct Anon35;
struct Anon36;
struct Anon37;
struct Anon38;
struct Anon39;
struct Anon4;
struct Anon40;
struct Anon41;
struct Anon42;
struct Anon43;
struct Anon44;
struct Anon45;
struct Anon46;
struct Anon47;
struct Anon48;
struct Anon49;
struct Anon5;
struct Anon50;
struct Anon51;
struct Anon52;
struct Anon53;
struct Anon54;
struct Anon6;
struct Anon7;
struct Anon8;
struct Anon9;
struct gekko_LockedCache{
std::array<uint8_t,16384> Data{};
uint32_t Base{};
bool Enabled_{};
};
struct gekko_FPR{
double PS0{};
double PS1{};
};
struct gekko_CPU{
std::array<uint32_t,32> GPR{};
std::array<gekko_FPR,32> FPR{};
uint32_t PC{};
uint32_t LR{};
uint32_t CTR{};
uint32_t CR{};
uint32_t XER{};
uint32_t MSR{};
uint32_t FPSCR{};
std::array<uint32_t,8> GQR{};
uint32_t HID0{};
uint32_t HID1{};
uint32_t HID2{};
uint32_t HID4{};
uint32_t WPAR{};
uint32_t DMAU{};
uint32_t DMAL{};
uint32_t L2CR{};
uint32_t SRR0{};
uint32_t SRR1{};
std::array<uint32_t,4> SPRG{};
uint32_t DSISR{};
uint32_t DAR{};
uint32_t SDR1{};
std::array<uint32_t,16> SR{};
std::array<std::array<uint32_t,2>,4> IBAT{};
std::array<std::array<uint32_t,2>,4> DBAT{};
uint32_t PVR{};
uint64_t TB{};
uint32_t DEC{};
uint32_t clockFrac{};
bool decArmed{};
bool Reserved{};
uint32_t ReserveAddr{};
gekko_LockedCache LC{};
bool Halted{};
std::string HaltReason{};
uint64_t Steps{};
bool ExtInt{};
gc_Machine* bus{};
gc_Machine* fetcher{};
std::function<bool(gekko_CPU*)> SC{};
};
using gekko_Flow=int64_t;
struct gekko_Inst{
uint32_t Addr{};
uint32_t Word{};
int64_t Len{};
std::string Mnem{};
std::string Text{};
gekko_Flow Flow{};
uint32_t Target{};
bool HasTarget{};
};
struct gekko_State{
std::array<uint32_t,32> GPR{};
std::array<gekko_FPR,32> FPR{};
uint32_t PC{};
uint32_t LR{};
uint32_t CTR{};
uint32_t CR{};
uint32_t XER{};
uint32_t MSR{};
uint32_t FPSCR{};
std::array<uint32_t,8> GQR{};
uint32_t HID0{};
uint32_t HID1{};
uint32_t HID2{};
uint32_t HID4{};
uint32_t WPAR{};
uint32_t DMAU{};
uint32_t DMAL{};
uint32_t L2CR{};
uint32_t SRR0{};
uint32_t SRR1{};
std::array<uint32_t,4> SPRG{};
uint32_t DSISR{};
uint32_t DAR{};
uint32_t SDR1{};
uint32_t PVR{};
std::array<uint32_t,16> SR{};
std::array<std::array<uint32_t,2>,4> IBAT{};
std::array<std::array<uint32_t,2>,4> DBAT{};
uint64_t TB{};
uint32_t DEC{};
uint32_t ClockFrac{};
bool DecArmed{};
bool Reserved{};
uint32_t ReserveAddr{};
Slice<uint8_t> LCData{};
uint32_t LCBase{};
bool LCEnabled{};
bool ExtInt{};
uint64_t Steps{};
bool Halted{};
std::string HaltReason{};
};
struct gekko_Case{
uint32_t Op{};
Map<std::string,uint32_t> GPR{};
Map<std::string,std::array<uint64_t,2>> FPR{};
Map<std::string,uint32_t> GQR{};
Map<std::string,uint32_t> Mem{};
uint32_t XER{};
uint32_t CR{};
uint32_t HID2{};
Map<std::string,uint32_t> OutGPR{};
Map<std::string,std::array<uint64_t,2>> OutFPR{};
Map<std::string,uint32_t> OutMem{};
uint32_t OutXER{};
uint32_t OutCR{};
bool CheckXER{};
bool CheckCR{};
Slice<std::string> DontCare{};
std::string Note{};
};
using gekko_Suite=Map<std::string,Slice<gekko_Case>>;
struct gcdsp_LoopFrame{
uint16_t Start{};
uint16_t End{};
uint16_t Count{};
};
struct gcdsp_CPU{
uint16_t PC{};
std::array<uint16_t,32> Reg{};
std::array<uint16_t,4096> IRAM{};
std::array<uint16_t,4096> DRAM{};
Slice<uint16_t> IROM{};
Slice<uint16_t> DROM{};
gc_dspBus bus{};
bool Halted{};
std::string Reason{};
bool Branched{};
bool InInterrupt{};
std::array<Slice<uint16_t>,4> Stacks{};
Slice<gcdsp_LoopFrame> Loops{};
};
struct gcdsp_extPending{
int64_t n{};
std::array<uint16_t,2> reg{};
std::array<uint16_t,2> val{};
};
struct gc_mi{
std::array<uint16_t,64> Reg{};
};
struct gc_ai{
uint32_t Control{};
uint32_t Volume{};
uint32_t SCnt{};
uint32_t ITCnt{};
};
struct gc_cp{
uint16_t Status{};
uint16_t Control{};
uint16_t Clear{};
std::array<uint16_t,64> Reg{};
};
struct gc_pe{
std::array<uint16_t,32> Reg{};
uint16_t Token{};
bool TokenEnable{};
bool FinishEnable{};
};
struct gc_wgPipe{
uint64_t Bytes{};
Slice<uint8_t> Buf{};
};
struct gc_di{
uint32_t SR{};
uint32_t Cover{};
std::array<uint32_t,3> Cmd{};
uint32_t MAR{};
uint32_t Length{};
uint32_t CR{};
uint32_t ImmBuf{};
uint32_t Cfg{};
int64_t BusyInstr{};
int64_t PendOff{};
uint32_t PendLen{};
uint32_t PendMAR{};
int64_t LastEnd{};
};
struct gc_Header{
std::string GameID{};
uint8_t DiscID{};
uint8_t Version{};
bool Streaming{};
std::string Title{};
uint32_t DOLOffset{};
uint32_t FSTOffset{};
uint32_t FSTSize{};
uint32_t FSTMaxSize{};
uint32_t FSTAddr{};
uint32_t UserOffset{};
uint32_t UserSize{};
};
struct gc_Apploader{
std::string Date{};
uint32_t Entry{};
uint32_t Size{};
uint32_t TrailerSize{};
};
struct gc_Disc{
std::string Path{};
int64_t Size{};
gc_Header Header{};
gc_Apploader Apploader{};
gc_FST* FST{};
os_File* f{};
std::string md5{};
};
struct gc_Segment{
int64_t Index{};
bool Text{};
uint32_t Offset{};
uint32_t Addr{};
uint32_t Size{};
Slice<uint8_t> Data{};
};
struct gc_DOL{
uint32_t Entry{};
uint32_t BSSAddr{};
uint32_t BSSSize{};
Slice<gc_Segment> Segments{};
};
struct gc_dsp{
uint32_t ToDSP{};
uint32_t FromDSP{};
uint32_t CSR{};
int64_t BootStep{};
bool UcodeRunning{};
bool AwaitValue{};
bool AwaitStartArg{};
uint32_t LoadCmd{};
uint32_t UcodeSrc{};
uint32_t UcodeLen{};
uint32_t UcodeDst{};
uint32_t UcodeEntry{};
gcdsp_CPU* Core{};
bool CoreHalt{};
bool CoreBlocked{};
bool corePolledEmpty{};
uint32_t ARMMAddr{};
uint32_t ARARAddr{};
uint32_t ARCtrl{};
uint32_t ARSize{};
uint32_t AIDStart{};
uint16_t AIDControl{};
uint16_t AIDRemaining{};
uint32_t AIDCur{};
uint64_t AIDAccum{};
uint32_t DSMAAddr{};
uint16_t DSPAddr{};
uint16_t DSCtrl{};
std::array<uint16_t,16> Coef{};
uint16_t AccFormat{};
uint32_t AccStart{};
uint32_t AccEnd{};
uint32_t AccCur{};
uint16_t AccPred{};
uint16_t AccYn1{};
uint16_t AccYn2{};
uint16_t AccGain{};
bool AccStopped{};
};
struct gc_exiChannel{
uint32_t CSR{};
uint32_t Data{};
uint32_t CR{};
uint32_t DMAAddr{};
uint32_t DMALen{};
int64_t Dev{};
int64_t Phase{};
uint32_t Command{};
};
struct gc_exi{
std::array<gc_exiChannel,3> Ch{};
std::array<uint8_t,64> SRAM{};
uint32_t RTC{};
};
struct gc_File{
std::string Path{};
std::string Name{};
bool Dir{};
int64_t Offset{};
int64_t Size{};
};
struct gc_FST{
Slice<gc_File> Entries{};
Slice<gc_File> files{};
};
struct gc_texCoord{
float s{};
float t{};
float q{};
};
struct gc_screenVertex{
float x{};
float y{};
float z{};
std::array<std::array<uint8_t,4>,2> col{};
std::array<gc_texCoord,8> tc{};
int64_t ntc{};
float invW{};
};
struct gc_rasterTri{
gc_screenVertex v0{};
gc_screenVertex v1{};
gc_screenVertex v2{};
float area{};
int64_t minX{};
int64_t maxX{};
int64_t minY{};
int64_t maxY{};
bool zEnable{};
bool zWrite{};
int64_t zFunc{};
};
struct gc_gpu{
Slice<uint8_t> Buf{};
std::array<uint64_t,256> Census{};
int64_t pixWritten{};
int64_t pixZRej{};
int64_t pixARej{};
int64_t profDraws{};
int64_t profCulled{};
int64_t profParFills{};
int64_t profSerFills{};
bool inDisplayList{};
std::array<uint32_t,256> CPReg{};
std::array<uint32_t,256> BP{};
std::array<uint32_t,4192> XFMem{};
std::array<std::array<uint32_t,2>,4> TevColorReg{};
std::array<std::array<uint32_t,2>,4> TevKonstReg{};
Slice<uint32_t> EFB{};
Slice<uint32_t> ZBuf{};
Slice<uint8_t> Tlut{};
Slice<gc_rasterTri> tris{};
gc_workPool* workers{};
};
struct gc_clipPlane{
float a{};
float b{};
float c{};
float d{};
};
struct gc_texAttr{
int64_t off{};
uint32_t desc{};
uint32_t fmt{};
uint32_t elem{};
uint32_t frac{};
};
struct gc_attrLayout{
int64_t stride{};
bool hasMatIdx{};
int64_t matIdxOff{};
int64_t posOff{};
uint32_t posDesc{};
uint32_t posFmt{};
int64_t posComps{};
uint32_t posFrac{};
int64_t nrmOff{};
uint32_t nrmDesc{};
uint32_t nrmFmt{};
int64_t nrmComps{};
int64_t nrmIdxComps{};
int64_t col0Off{};
uint32_t col0Desc{};
uint32_t col0Comp{};
std::array<int64_t,8> texMtxIdxOff{};
std::array<gc_texAttr,8> tex{};
};
struct gc_rstats{
int64_t written{};
int64_t zRej{};
int64_t aRej{};
};
struct gc_tevStage{
uint32_t cc{};
uint32_t ac{};
int64_t texmap{};
int64_t texcoord{};
bool texEnable{};
uint32_t rasSel{};
std::array<int64_t,4> swapRas{};
std::array<int64_t,4> swapTex{};
std::array<float,3> konstC{};
float konstA{};
uint32_t cdest{};
uint32_t adest{};
RRTevOperands rrOperands{};
};
struct gc_texState{
int64_t format{};
int64_t width{};
int64_t height{};
uint32_t base{};
int64_t wrapS{};
int64_t wrapT{};
int64_t tlutOff{};
int64_t tlutFmt{};
uint32_t tlutRAM{};
bool tlutFromRAM{};
};
struct gc_tevState{
int64_t numStages{};
std::array<gc_tevStage,16> stages{};
std::array<std::array<float,4>,4> seed{};
std::array<gc_texState,8> tex{};
std::array<bool,8> texValid{};
bool canHalt{};
bool rrPrepared{};
};
struct gc_clipVertex{
float cx{};
float cy{};
float cz{};
float cw{};
std::array<std::array<uint8_t,4>,2> col{};
std::array<gc_texCoord,8> tc{};
int64_t ntc{};
};
struct gc_idleSnap{
uint32_t PC{};
uint32_t LR{};
uint32_t CTR{};
uint32_t CR{};
uint32_t XER{};
uint32_t MSR{};
std::array<uint32_t,32> GPR{};
std::array<std::array<uint64_t,2>,32> FPR{};
bool operator==(const gc_idleSnap&)const=default;
};
struct gc_idleState{
bool armed{};
gc_idleSnap snap{};
uint64_t stores{};
int64_t insns{};
int64_t period{};
uint64_t next{};
uint64_t Skipped{};
uint64_t Hits{};
};
struct gc_pi{
uint32_t Cause{};
uint32_t Mask{};
uint32_t FIFOBase{};
uint32_t FIFOEnd{};
uint32_t FIFOWrite{};
uint32_t ResetCode{};
};
struct gc_vi{
std::array<uint32_t,4> DI{};
uint32_t TFBL{};
uint32_t BFBL{};
uint32_t DCR{};
uint64_t Field{};
uint32_t Line{};
uint32_t Counter{};
};
struct gc_padPort{
bool Connected{};
uint16_t Buttons{};
uint8_t StickX{};
uint8_t StickY{};
uint8_t SubX{};
uint8_t SubY{};
uint8_t TriggerL{};
uint8_t TriggerR{};
};
struct gc_si{
std::array<std::array<uint32_t,3>,4> Chan{};
uint32_t Poll{};
uint32_t ComCSR{};
uint32_t Status{};
uint32_t ExiLk{};
std::array<uint8_t,128> IOBuf{};
std::array<gc_padPort,4> Pad{};
};
struct gc_stackProf{
uint64_t every{};
int64_t depth{};
uint64_t next{};
Map<std::string,gc_StackSample*> stacks{};
int64_t samples{};
};
struct gc_profCounters{
int64_t cmds{};
int64_t draws{};
int64_t culled{};
int64_t frags{};
int64_t zRejected{};
int64_t aRejected{};
int64_t fifoBytes{};
};
struct gc_ProfileBucket{
std::string Name{};
double Millis{};
int64_t Count{};
};
struct gc_ProfileCounter{
std::string Name{};
int64_t Value{};
};
struct gc_FrameProfile{
double TotalMs{};
Slice<gc_ProfileBucket> Buckets{};
Slice<gc_ProfileCounter> Counters{};
bool Drew{};
};
struct gc_profState{
std::array<int64_t,5> ns{};
std::array<int64_t,5> count{};
time_Time runStart{};
bool inRun{};
int64_t frameNs{};
time_Time fifoStart{};
bool inFIFO{};
gc_profCounters base{};
uint64_t baseInstr{};
gc_FrameProfile last{};
bool has{};
};
struct gc_runState{
Map<uint32_t,bool> breakpoints{};
};
struct gc_Machine{
Slice<uint8_t> RAM{};
Slice<uint8_t> ARAM{};
gekko_CPU* CPU{};
gc_Disc* disc{};
std::string discMD5{};
gc_pi pi{};
gc_mi mi{};
gc_vi vi{};
gc_di di{};
gc_si si{};
gc_exi exi{};
gc_ai ai{};
gc_dsp dsp{};
gc_cp cp{};
gc_pe pe{};
gc_gpu gpu{};
gc_wgPipe wgFIFO{};
Slice<std::string> Log{};
Map<std::string,bool> logSeen{};
std::function<void(gc_Machine*,uint32_t)> OnStep{};
uint32_t WatchLo{};
uint32_t WatchHi{};
std::function<void(uint32_t,uint32_t,uint32_t)> OnWrite{};
uint32_t RWatchLo{};
uint32_t RWatchHi{};
std::function<void(uint32_t,uint32_t,uint32_t)> OnRead{};
std::function<void(int64_t,uint32_t,uint32_t)> OnDVDRead{};
std::function<void(gc_Machine*)> OnDisplay{};
std::function<void(Slice<uint8_t>)> OnFIFO{};
std::function<void(Slice<uint8_t>)> AIDTap{};
std::function<void(gc_Machine*,uint8_t,Slice<uint32_t>)> OnGXCmd{};
std::function<void(int64_t,int64_t,gc_PixelEvent)> OnPixel{};
std::function<void(gc_Machine*)> OnFlip{};
bool StopRequested{};
bool SingleThreaded{};
uint64_t stores{};
gc_idleState idle{};
bool noIdle{};
gc_stackProf stack{};
uint64_t Instrs{};
bool Profile{};
gc_profState prof{};
int64_t gxCmdCount{};
int64_t gxStopAfter{};
bool gxStopped{};
int64_t gxTotalCmds{};
gc_runState run{};
bool noSpin{};
};
struct gc_PixelEvent{
uint8_t R{};
uint8_t G{};
uint8_t B{};
uint8_t A{};
bool Drawn{};
};
struct gc_Result{
uint64_t Steps{};
uint32_t PC{};
std::string Reason{};
};
struct gc_StackSample{
Slice<uint32_t> Stack{};
int64_t Count{};
};
struct Anon0{std::array<uint8_t,16384> Data{};uint32_t Base{};bool Enabled_{};};
struct Anon1{double PS0{};double PS1{};};
struct Anon10{uint32_t Control{};uint32_t Volume{};uint32_t SCnt{};uint32_t ITCnt{};};
struct Anon11{uint16_t Status{};uint16_t Control{};uint16_t Clear{};std::array<uint16_t,64> Reg{};};
struct Anon12{std::array<uint16_t,32> Reg{};uint16_t Token{};bool TokenEnable{};bool FinishEnable{};};
struct Anon13{uint64_t Bytes{};Slice<uint8_t> Buf{};};
struct Anon14{uint32_t SR{};uint32_t Cover{};std::array<uint32_t,3> Cmd{};uint32_t MAR{};uint32_t Length{};uint32_t CR{};uint32_t ImmBuf{};uint32_t Cfg{};int64_t BusyInstr{};int64_t PendOff{};uint32_t PendLen{};uint32_t PendMAR{};int64_t LastEnd{};};
struct Anon15{std::string GameID{};uint8_t DiscID{};uint8_t Version{};bool Streaming{};std::string Title{};uint32_t DOLOffset{};uint32_t FSTOffset{};uint32_t FSTSize{};uint32_t FSTMaxSize{};uint32_t FSTAddr{};uint32_t UserOffset{};uint32_t UserSize{};};
struct Anon16{std::string Date{};uint32_t Entry{};uint32_t Size{};uint32_t TrailerSize{};};
struct Anon17{std::string Path{};int64_t Size{};gc_Header Header{};gc_Apploader Apploader{};gc_FST* FST{};os_File* f{};std::string md5{};};
struct Anon18{int64_t Index{};bool Text{};uint32_t Offset{};uint32_t Addr{};uint32_t Size{};Slice<uint8_t> Data{};};
struct Anon19{uint32_t Entry{};uint32_t BSSAddr{};uint32_t BSSSize{};Slice<gc_Segment> Segments{};};
struct Anon2{std::array<uint32_t,32> GPR{};std::array<gekko_FPR,32> FPR{};uint32_t PC{};uint32_t LR{};uint32_t CTR{};uint32_t CR{};uint32_t XER{};uint32_t MSR{};uint32_t FPSCR{};std::array<uint32_t,8> GQR{};uint32_t HID0{};uint32_t HID1{};uint32_t HID2{};uint32_t HID4{};uint32_t WPAR{};uint32_t DMAU{};uint32_t DMAL{};uint32_t L2CR{};uint32_t SRR0{};uint32_t SRR1{};std::array<uint32_t,4> SPRG{};uint32_t DSISR{};uint32_t DAR{};uint32_t SDR1{};std::array<uint32_t,16> SR{};std::array<std::array<uint32_t,2>,4> IBAT{};std::array<std::array<uint32_t,2>,4> DBAT{};uint32_t PVR{};uint64_t TB{};uint32_t DEC{};uint32_t clockFrac{};bool decArmed{};bool Reserved{};uint32_t ReserveAddr{};gekko_LockedCache LC{};bool Halted{};std::string HaltReason{};uint64_t Steps{};bool ExtInt{};gc_Machine* bus{};gc_Machine* fetcher{};std::function<bool(gekko_CPU*)> SC{};};
struct Anon20{uint32_t ToDSP{};uint32_t FromDSP{};uint32_t CSR{};int64_t BootStep{};bool UcodeRunning{};bool AwaitValue{};bool AwaitStartArg{};uint32_t LoadCmd{};uint32_t UcodeSrc{};uint32_t UcodeLen{};uint32_t UcodeDst{};uint32_t UcodeEntry{};gcdsp_CPU* Core{};bool CoreHalt{};bool CoreBlocked{};bool corePolledEmpty{};uint32_t ARMMAddr{};uint32_t ARARAddr{};uint32_t ARCtrl{};uint32_t ARSize{};uint32_t AIDStart{};uint16_t AIDControl{};uint16_t AIDRemaining{};uint32_t AIDCur{};uint64_t AIDAccum{};uint32_t DSMAAddr{};uint16_t DSPAddr{};uint16_t DSCtrl{};std::array<uint16_t,16> Coef{};uint16_t AccFormat{};uint32_t AccStart{};uint32_t AccEnd{};uint32_t AccCur{};uint16_t AccPred{};uint16_t AccYn1{};uint16_t AccYn2{};uint16_t AccGain{};bool AccStopped{};};
struct Anon21{gc_Machine* m{};};
struct Anon22{uint32_t CSR{};uint32_t Data{};uint32_t CR{};uint32_t DMAAddr{};uint32_t DMALen{};int64_t Dev{};int64_t Phase{};uint32_t Command{};};
struct Anon23{std::array<gc_exiChannel,3> Ch{};std::array<uint8_t,64> SRAM{};uint32_t RTC{};};
struct Anon24{std::string Path{};std::string Name{};bool Dir{};int64_t Offset{};int64_t Size{};};
struct Anon25{Slice<gc_File> Entries{};Slice<gc_File> files{};};
struct Anon26{Slice<uint8_t> Buf{};std::array<uint64_t,256> Census{};int64_t pixWritten{};int64_t pixZRej{};int64_t pixARej{};int64_t profDraws{};int64_t profCulled{};int64_t profParFills{};int64_t profSerFills{};bool inDisplayList{};std::array<uint32_t,256> CPReg{};std::array<uint32_t,256> BP{};std::array<uint32_t,4192> XFMem{};std::array<std::array<uint32_t,2>,4> TevColorReg{};std::array<std::array<uint32_t,2>,4> TevKonstReg{};Slice<uint32_t> EFB{};Slice<uint32_t> ZBuf{};Slice<uint8_t> Tlut{};Slice<gc_rasterTri> tris{};gc_workPool* workers{};};
struct Anon27{float a{};float b{};float c{};float d{};};
struct Anon28{int64_t stride{};bool hasMatIdx{};int64_t matIdxOff{};int64_t posOff{};uint32_t posDesc{};uint32_t posFmt{};int64_t posComps{};uint32_t posFrac{};int64_t nrmOff{};uint32_t nrmDesc{};uint32_t nrmFmt{};int64_t nrmComps{};int64_t nrmIdxComps{};int64_t col0Off{};uint32_t col0Desc{};uint32_t col0Comp{};std::array<int64_t,8> texMtxIdxOff{};std::array<gc_texAttr,8> tex{};};
struct Anon29{int64_t off{};uint32_t desc{};uint32_t fmt{};uint32_t elem{};uint32_t frac{};};
struct Anon3{uint32_t Addr{};uint32_t Word{};int64_t Len{};std::string Mnem{};std::string Text{};gekko_Flow Flow{};uint32_t Target{};bool HasTarget{};};
struct Anon30{gc_screenVertex v0{};gc_screenVertex v1{};gc_screenVertex v2{};float area{};int64_t minX{};int64_t maxX{};int64_t minY{};int64_t maxY{};bool zEnable{};bool zWrite{};int64_t zFunc{};};
struct Anon31{int64_t written{};int64_t zRej{};int64_t aRej{};};
struct Anon32{uint32_t cc{};uint32_t ac{};int64_t texmap{};int64_t texcoord{};bool texEnable{};uint32_t rasSel{};std::array<int64_t,4> swapRas{};std::array<int64_t,4> swapTex{};std::array<float,3> konstC{};float konstA{};uint32_t cdest{};uint32_t adest{};};
struct Anon33{int64_t numStages{};std::array<gc_tevStage,16> stages{};std::array<std::array<float,4>,4> seed{};std::array<gc_texState,8> tex{};std::array<bool,8> texValid{};bool canHalt{};};
struct Anon34{float s{};float t{};float q{};};
struct Anon35{int64_t format{};int64_t width{};int64_t height{};uint32_t base{};int64_t wrapS{};int64_t wrapT{};int64_t tlutOff{};int64_t tlutFmt{};uint32_t tlutRAM{};bool tlutFromRAM{};};
struct Anon36{float x{};float y{};float z{};std::array<std::array<uint8_t,4>,2> col{};std::array<gc_texCoord,8> tc{};int64_t ntc{};float invW{};};
struct Anon37{float cx{};float cy{};float cz{};float cw{};std::array<std::array<uint8_t,4>,2> col{};std::array<gc_texCoord,8> tc{};int64_t ntc{};};
struct Anon38{uint32_t PC{};uint32_t LR{};uint32_t CTR{};uint32_t CR{};uint32_t XER{};uint32_t MSR{};std::array<uint32_t,32> GPR{};std::array<std::array<uint64_t,2>,32> FPR{};};
struct Anon39{bool armed{};gc_idleSnap snap{};uint64_t stores{};int64_t insns{};int64_t period{};uint64_t next{};uint64_t Skipped{};uint64_t Hits{};};
struct Anon4{std::array<uint32_t,32> GPR{};std::array<gekko_FPR,32> FPR{};uint32_t PC{};uint32_t LR{};uint32_t CTR{};uint32_t CR{};uint32_t XER{};uint32_t MSR{};uint32_t FPSCR{};std::array<uint32_t,8> GQR{};uint32_t HID0{};uint32_t HID1{};uint32_t HID2{};uint32_t HID4{};uint32_t WPAR{};uint32_t DMAU{};uint32_t DMAL{};uint32_t L2CR{};uint32_t SRR0{};uint32_t SRR1{};std::array<uint32_t,4> SPRG{};uint32_t DSISR{};uint32_t DAR{};uint32_t SDR1{};uint32_t PVR{};std::array<uint32_t,16> SR{};std::array<std::array<uint32_t,2>,4> IBAT{};std::array<std::array<uint32_t,2>,4> DBAT{};uint64_t TB{};uint32_t DEC{};uint32_t ClockFrac{};bool DecArmed{};bool Reserved{};uint32_t ReserveAddr{};Slice<uint8_t> LCData{};uint32_t LCBase{};bool LCEnabled{};bool ExtInt{};uint64_t Steps{};bool Halted{};std::string HaltReason{};};
struct Anon40{Slice<uint8_t> RAM{};Slice<uint8_t> ARAM{};gekko_CPU* CPU{};gc_Disc* disc{};std::string discMD5{};gc_pi pi{};gc_mi mi{};gc_vi vi{};gc_di di{};gc_si si{};gc_exi exi{};gc_ai ai{};gc_dsp dsp{};gc_cp cp{};gc_pe pe{};gc_gpu gpu{};gc_wgPipe wgFIFO{};Slice<std::string> Log{};Map<std::string,bool> logSeen{};std::function<void(gc_Machine*,uint32_t)> OnStep{};uint32_t WatchLo{};uint32_t WatchHi{};std::function<void(uint32_t,uint32_t,uint32_t)> OnWrite{};uint32_t RWatchLo{};uint32_t RWatchHi{};std::function<void(uint32_t,uint32_t,uint32_t)> OnRead{};std::function<void(int64_t,uint32_t,uint32_t)> OnDVDRead{};std::function<void(gc_Machine*)> OnDisplay{};std::function<void(Slice<uint8_t>)> OnFIFO{};std::function<void(Slice<uint8_t>)> AIDTap{};std::function<void(gc_Machine*,uint8_t,Slice<uint32_t>)> OnGXCmd{};std::function<void(int64_t,int64_t,gc_PixelEvent)> OnPixel{};std::function<void(gc_Machine*)> OnFlip{};bool StopRequested{};bool SingleThreaded{};uint64_t stores{};gc_idleState idle{};bool noIdle{};gc_stackProf stack{};uint64_t Instrs{};bool Profile{};gc_profState prof{};int64_t gxCmdCount{};int64_t gxStopAfter{};bool gxStopped{};int64_t gxTotalCmds{};gc_runState run{};bool noSpin{};};
struct Anon41{uint8_t R{};uint8_t G{};uint8_t B{};uint8_t A{};bool Drawn{};};
struct Anon42{uint32_t Cause{};uint32_t Mask{};uint32_t FIFOBase{};uint32_t FIFOEnd{};uint32_t FIFOWrite{};uint32_t ResetCode{};};
struct Anon43{std::string Name{};double Millis{};int64_t Count{};};
struct Anon44{std::string Name{};int64_t Value{};};
struct Anon45{double TotalMs{};Slice<gc_ProfileBucket> Buckets{};Slice<gc_ProfileCounter> Counters{};bool Drew{};};
struct Anon46{std::array<int64_t,5> ns{};std::array<int64_t,5> count{};time_Time runStart{};bool inRun{};int64_t frameNs{};time_Time fifoStart{};bool inFIFO{};gc_profCounters base{};uint64_t baseInstr{};gc_FrameProfile last{};bool has{};};
struct Anon47{int64_t cmds{};int64_t draws{};int64_t culled{};int64_t frags{};int64_t zRejected{};int64_t aRejected{};int64_t fifoBytes{};};
struct Anon48{uint64_t Steps{};uint32_t PC{};std::string Reason{};};
struct Anon49{Map<uint32_t,bool> breakpoints{};};
struct Anon5{uint32_t Op{};Map<std::string,uint32_t> GPR{};Map<std::string,std::array<uint64_t,2>> FPR{};Map<std::string,uint32_t> GQR{};Map<std::string,uint32_t> Mem{};uint32_t XER{};uint32_t CR{};uint32_t HID2{};Map<std::string,uint32_t> OutGPR{};Map<std::string,std::array<uint64_t,2>> OutFPR{};Map<std::string,uint32_t> OutMem{};uint32_t OutXER{};uint32_t OutCR{};bool CheckXER{};bool CheckCR{};Slice<std::string> DontCare{};std::string Note{};};
struct Anon50{std::array<std::array<uint32_t,3>,4> Chan{};uint32_t Poll{};uint32_t ComCSR{};uint32_t Status{};uint32_t ExiLk{};std::array<uint8_t,128> IOBuf{};std::array<gc_padPort,4> Pad{};};
struct Anon51{bool Connected{};uint16_t Buttons{};uint8_t StickX{};uint8_t StickY{};uint8_t SubX{};uint8_t SubY{};uint8_t TriggerL{};uint8_t TriggerR{};};
struct Anon52{Slice<uint32_t> Stack{};int64_t Count{};};
struct Anon53{uint64_t every{};int64_t depth{};uint64_t next{};Map<std::string,gc_StackSample*> stacks{};int64_t samples{};};
struct Anon54{std::array<uint32_t,4> DI{};uint32_t TFBL{};uint32_t BFBL{};uint32_t DCR{};uint64_t Field{};uint32_t Line{};uint32_t Counter{};};
struct Anon6{uint16_t PC{};std::array<uint16_t,32> Reg{};std::array<uint16_t,4096> IRAM{};std::array<uint16_t,4096> DRAM{};Slice<uint16_t> IROM{};Slice<uint16_t> DROM{};gc_dspBus bus{};bool Halted{};std::string Reason{};bool Branched{};bool InInterrupt{};std::array<Slice<uint16_t>,4> Stacks{};Slice<gcdsp_LoopFrame> Loops{};};
struct Anon7{uint16_t Start{};uint16_t End{};uint16_t Count{};};
struct Anon8{int64_t n{};std::array<uint16_t,2> reg{};std::array<uint16_t,2> val{};};
struct Anon9{std::array<uint16_t,64> Reg{};};

#include "adapters-decl.h"
void gekko_LockedCache_Reset(gekko_LockedCache* l);
bool gekko_LockedCache_Enabled(gekko_LockedCache* l);
bool gekko_LockedCache_Contains(gekko_LockedCache* l,uint32_t a);
uint32_t gekko_LockedCache_off(gekko_LockedCache* l,uint32_t a);
uint8_t gekko_LockedCache_Read8(gekko_LockedCache* l,uint32_t a);
uint16_t gekko_LockedCache_Read16(gekko_LockedCache* l,uint32_t a);
uint32_t gekko_LockedCache_Read32(gekko_LockedCache* l,uint32_t a);
void gekko_LockedCache_Write8(gekko_LockedCache* l,uint32_t a,uint8_t v);
void gekko_LockedCache_Write16(gekko_LockedCache* l,uint32_t a,uint16_t v);
void gekko_LockedCache_Write32(gekko_LockedCache* l,uint32_t a,uint32_t v);
void gekko_CPU_setHID2(gekko_CPU* c,uint32_t v);
void gekko_CPU_dcbz(gekko_CPU* c,uint32_t ea);
void gekko_CPU_dcbzL(gekko_CPU* c,uint32_t ea);
void gekko_CPU_runDMA(gekko_CPU* c);
uint32_t gekko_CPU_memPhys(gekko_CPU* c,uint32_t ea);
void gekko_CPU_Reset(gekko_CPU* c);
uint64_t gekko_CPU_CurPC(gekko_CPU* c);
void gekko_CPU_SetPC(gekko_CPU* c,uint32_t pc);
uint32_t gekko_CPU_Reg(gekko_CPU* c,uint32_t i);
void gekko_CPU_SetReg(gekko_CPU* c,uint32_t i,uint32_t v);
uint32_t gekko_CPU_CRField(gekko_CPU* c,uint32_t n);
void gekko_CPU_SetCRField(gekko_CPU* c,uint32_t n,uint32_t v);
void gekko_CPU_setCR0(gekko_CPU* c,uint32_t v);
void gekko_CPU_setCR1(gekko_CPU* c);
void gekko_CPU_setCA(gekko_CPU* c,bool on);
uint32_t gekko_CPU_ca(gekko_CPU* c);
void gekko_CPU_setOV(gekko_CPU* c,bool on);
uint8_t gekko_CPU_read8(gekko_CPU* c,uint32_t ea);
uint16_t gekko_CPU_read16(gekko_CPU* c,uint32_t ea);
uint32_t gekko_CPU_read32(gekko_CPU* c,uint32_t ea);
uint64_t gekko_CPU_read64(gekko_CPU* c,uint32_t ea);
void gekko_CPU_write8(gekko_CPU* c,uint32_t ea,uint8_t v);
void gekko_CPU_write16(gekko_CPU* c,uint32_t ea,uint16_t v);
void gekko_CPU_write32(gekko_CPU* c,uint32_t ea,uint32_t v);
void gekko_CPU_write64(gekko_CPU* c,uint32_t ea,uint64_t v);
void gekko_CPU_clearReservation(gekko_CPU* c,uint32_t pa);
uint8_t gekko_CPU_ReadMem(gekko_CPU* c,uint32_t ea);
double gekko_f32(double v);
bool gekko_CPU_psEnabled(gekko_CPU* c);
uint64_t gekko_f64bits(double f);
double gekko_f64from(uint64_t b);
uint64_t gekko_bits64(double f);
uint32_t gekko_float32bitsOf(double v);
gekko_Inst gekko_Decode(Slice<uint8_t> code,uint32_t addr);
gekko_Inst gekko_DecodeWord(uint32_t w,uint32_t addr);
std::string gekko_r(uint32_t n);
std::string gekko_fr(uint32_t n);
std::string gekko_dot(std::string mnem,uint32_t w);
std::string gekko_oeDot(std::string mnem,uint32_t w);
void gekko_decode(gekko_Inst* in,uint32_t w,uint32_t addr);
void gekko_mem(gekko_Inst* in,std::string mnem,std::string reg,uint32_t w);
void gekko_memx(gekko_Inst* in,std::string mnem,std::string reg,uint32_t w);
void gekko_decodeRlwinm(gekko_Inst* in,uint32_t w);
bool gekko_boAlways(uint32_t bo);
std::tuple<std::string,bool> gekko_branchMnem(uint32_t bo,uint32_t bi,std::string suffix,bool link);
void gekko_decodeBC(gekko_Inst* in,uint32_t w,uint32_t addr);
void gekko_decodeB(gekko_Inst* in,uint32_t w,uint32_t addr);
void gekko_decode19(gekko_Inst* in,uint32_t w);
void gekko_crLogic(gekko_Inst* in,uint32_t w);
void gekko_decode31(gekko_Inst* in,uint32_t w);
void gekko_decode59(gekko_Inst* in,uint32_t w);
void gekko_decode63(gekko_Inst* in,uint32_t w);
void gekko_decodePS(gekko_Inst* in,uint32_t w);
void gekko_psqX(gekko_Inst* in,std::string mnem,uint32_t w);
void gekko_decodePSQ(gekko_Inst* in,uint32_t w);
int32_t gekko_psqDisp(uint32_t w);
Slice<std::string> gekko_Disassemble(Slice<uint8_t> code,uint32_t base);
uint32_t gekko_CPU_vectorBase(gekko_CPU* c);
void gekko_CPU_Exception(gekko_CPU* c,uint32_t vec,uint32_t resume,uint32_t srr1Extra);
void gekko_CPU_programException(gekko_CPU* c,uint32_t kind);
bool gekko_needsFPU(uint32_t w);
bool gekko_CPU_fpUnavailable(gekko_CPU* c,uint32_t w,uint32_t pc);
bool gekko_CPU_checkInterrupt(gekko_CPU* c);
void gekko_CPU_Interrupt(gekko_CPU* c,bool pending);
int64_t gekko_CPU_Step(gekko_CPU* c);
void gekko_CPU_execute(gekko_CPU* c,uint32_t w,uint32_t pc);
uint32_t gekko_CPU_raOrZero(gekko_CPU* c,uint32_t w);
uint32_t gekko_CPU_ea(gekko_CPU* c,uint32_t w);
uint32_t gekko_CPU_eaU(gekko_CPU* c,uint32_t w);
uint32_t gekko_CPU_eax(gekko_CPU* c,uint32_t w);
uint32_t gekko_CPU_eaxU(gekko_CPU* c,uint32_t w);
void gekko_CPU_rc(gekko_CPU* c,uint32_t w,uint32_t result);
void gekko_CPU_loadFS(gekko_CPU* c,uint32_t d,uint32_t ea);
uint32_t gekko_rotl32(uint32_t v,uint32_t n);
uint32_t gekko_mask32(uint32_t mb,uint32_t me);
bool gekko_carryAdd(uint32_t a,uint32_t b,uint32_t carryIn);
bool gekko_carrySub(uint32_t a,uint32_t b);
bool gekko_overflowAdd(uint32_t a,uint32_t b,uint32_t r);
void gekko_CPU_compareArith(gekko_CPU* c,uint32_t crf,uint32_t a,uint32_t b);
void gekko_CPU_compareLogical(gekko_CPU* c,uint32_t crf,uint32_t a,uint32_t b);
bool gekko_CPU_trapCond(gekko_CPU* c,uint32_t to,uint32_t a,uint32_t b);
bool gekko_CPU_branchTaken(gekko_CPU* c,uint32_t bo,uint32_t bi);
void gekko_CPU_execBC(gekko_CPU* c,uint32_t w,uint32_t pc);
void gekko_CPU_execB(gekko_CPU* c,uint32_t w,uint32_t pc);
void gekko_CPU_exec19(gekko_CPU* c,uint32_t w,uint32_t pc);
void gekko_CPU_execCRLogic(gekko_CPU* c,uint32_t w);
void gekko_CPU_exec31(gekko_CPU* c,uint32_t w,uint32_t pc);
void gekko_CPU_srawi(gekko_CPU* c,uint32_t w,uint32_t sh);
void gekko_CPU_oeRc(gekko_CPU* c,uint32_t w,uint32_t result,bool overflow);
void gekko_CPU_loadString(gekko_CPU* c,uint32_t d,uint32_t ea,uint32_t n);
void gekko_CPU_storeString(gekko_CPU* c,uint32_t d,uint32_t ea,uint32_t n);
void gekko_CPU_setFPRF(gekko_CPU* c,double v);
void gekko_CPU_setFPSCRBit(gekko_CPU* c,uint32_t bit);
bool gekko_isSNaN(double v);
void gekko_CPU_fpResult(gekko_CPU* c,uint32_t w,uint32_t d,double v);
void gekko_CPU_fpResultS(gekko_CPU* c,uint32_t w,uint32_t d,double v);
double gekko_CPU_fadd(gekko_CPU* c,double a,double b);
double gekko_CPU_fsub(gekko_CPU* c,double a,double b);
double gekko_CPU_fmul(gekko_CPU* c,double a,double b);
double gekko_CPU_fdiv(gekko_CPU* c,double a,double b);
int64_t gekko_sign(double v);
double gekko_CPU_fmaddRaw(gekko_CPU* c,double a,double cc,double b);
double gekko_CPU_fsqrt(gekko_CPU* c,double v);
double gekko_CPU_fres(gekko_CPU* c,double v);
double gekko_CPU_frsqrte(gekko_CPU* c,double v);
void gekko_CPU_exec59(gekko_CPU* c,uint32_t w);
void gekko_CPU_exec63(gekko_CPU* c,uint32_t w);
void gekko_CPU_rcF(gekko_CPU* c,uint32_t w);
void gekko_CPU_fcmp(gekko_CPU* c,uint32_t crf,double a,double b,bool ordered);
void gekko_CPU_fctiw(gekko_CPU* c,uint32_t w,uint32_t d,double v,bool truncate);
double gekko_CPU_roundToNearest(gekko_CPU* c,double v);
std::string gekko_Flow_String(gekko_Flow f);
std::string gekko_Inst_String(gekko_Inst in);
uint32_t gekko_opcd(uint32_t w);
uint32_t gekko_rs(uint32_t w);
uint32_t gekko_ra(uint32_t w);
uint32_t gekko_rb(uint32_t w);
uint32_t gekko_rc(uint32_t w);
uint32_t gekko_xo10(uint32_t w);
uint32_t gekko_xo5(uint32_t w);
uint32_t gekko_xo6(uint32_t w);
bool gekko_rcbit(uint32_t w);
bool gekko_oe(uint32_t w);
bool gekko_lk(uint32_t w);
bool gekko_aa(uint32_t w);
int32_t gekko_simm(uint32_t w);
uint32_t gekko_uimm(uint32_t w);
uint32_t gekko_crfD(uint32_t w);
uint32_t gekko_crfS(uint32_t w);
uint32_t gekko_shOf(uint32_t w);
uint32_t gekko_mbOf(uint32_t w);
uint32_t gekko_meOf(uint32_t w);
uint32_t gekko_sprOf(uint32_t w);
std::string gekko_sprStr(uint32_t n);
std::tuple<uint32_t,bool> gekko_CPU_Translate(gekko_CPU* c,uint32_t ea,bool store,bool insn);
std::tuple<uint32_t,bool> gekko_CPU_Translate_reference(gekko_CPU* c,uint32_t ea,bool store,bool insn);
std::tuple<uint32_t,bool> gekko_batMatch(uint32_t upper,uint32_t lower,uint32_t ea,bool user);
std::string gekko_CPU_BATString(gekko_CPU* c);
std::string gekko_batLine(std::string kind,int64_t i,uint32_t upper,uint32_t lower);
std::tuple<uint32_t,int32_t> gekko_gqrLoad(uint32_t g);
std::tuple<uint32_t,int32_t> gekko_gqrStore(uint32_t g);
double gekko_CPU_dequantize(gekko_CPU* c,uint32_t raw,uint32_t typ,int32_t scale);
uint32_t gekko_CPU_quantize(gekko_CPU* c,double v,uint32_t typ,int32_t scale);
double gekko_clamp(double v,double lo,double hi);
uint32_t gekko_qsize(uint32_t typ);
uint32_t gekko_CPU_readQ(gekko_CPU* c,uint32_t ea,uint32_t typ);
void gekko_CPU_writeQ(gekko_CPU* c,uint32_t ea,uint32_t typ,uint32_t v);
void gekko_CPU_psqLoad(gekko_CPU* c,uint32_t d,uint32_t ea,uint32_t gqr,uint32_t wBit);
void gekko_CPU_psqStore(gekko_CPU* c,uint32_t s,uint32_t ea,uint32_t gqr,uint32_t wBit);
void gekko_CPU_execPSQ(gekko_CPU* c,uint32_t w);
void gekko_CPU_execPS(gekko_CPU* c,uint32_t w,uint32_t pc);
double gekko_negf(double v);
void gekko_CPU_psResult(gekko_CPU* c,uint32_t w,uint32_t d,double p0,double p1);
uint32_t gekko_CPU_readSPR(gekko_CPU* c,uint32_t n,uint32_t pc);
void gekko_CPU_writeSPR(gekko_CPU* c,uint32_t n,uint32_t v,uint32_t pc);
gekko_State gekko_CPU_Snapshot(gekko_CPU* c);
void gekko_CPU_Restore(gekko_CPU* c,gekko_State s);
void gekko_CPU_tick(gekko_CPU* c,int64_t cycles);
void gekko_CPU_setDEC(gekko_CPU* c,uint32_t v);
uint64_t gekko_CPU_InstrsToDecUnderflow(gekko_CPU* c);
void gekko_CPU_SkipInstructions(gekko_CPU* c,uint64_t n);
gcdsp_CPU* gcdsp_New(gc_dspBus bus);
void gcdsp_CPU_SetBus(gcdsp_CPU* c,gc_dspBus bus);
int64_t gcdsp_CPU_ac(gcdsp_CPU* c,int64_t n);
void gcdsp_CPU_setAc(gcdsp_CPU* c,int64_t n,int64_t v);
int64_t gcdsp_CPU_ax(gcdsp_CPU* c,int64_t n);
int64_t gcdsp_CPU_prod(gcdsp_CPU* c);
int64_t gcdsp_CPU_prodRounded(gcdsp_CPU* c);
uint16_t gcdsp_CPU_sr(gcdsp_CPU* c);
void gcdsp_CPU_setSR(gcdsp_CPU* c,uint16_t v);
void gcdsp_CPU_setFlag(gcdsp_CPU* c,uint16_t bit,bool on);
std::string gcdsp_regName(uint16_t r);
std::string gcdsp_lowHigh(uint16_t bit);
std::string gcdsp_condName(uint16_t cc);
std::tuple<std::string,uint16_t> gcdsp_Disasm(std::function<uint16_t(uint16_t)> read,uint16_t pc);
std::string gcdsp_arithMnemonic(uint16_t op);
std::string gcdsp_extMnemonic(uint16_t ext);
std::tuple<uint16_t,bool> gcdsp_branchTarget(std::function<uint16_t(uint16_t)> read,uint16_t pc);
std::tuple<Slice<uint16_t>,int64_t,bool> gcdsp_DisasmValidate(std::function<uint16_t(uint16_t)> read,uint16_t nWords);
std::string gcdsp_DisasmRange(std::function<uint16_t(uint16_t)> read,uint16_t pc,uint16_t nWords);
bool gcdsp_CPU_Step(gcdsp_CPU* c);
uint16_t gcdsp_CPU_execute(gcdsp_CPU* c,uint16_t pc,uint16_t op);
uint16_t gcdsp_CPU_execArith(gcdsp_CPU* c,uint16_t pc,uint16_t op);
std::string gcdsp_mustText(std::function<uint16_t(uint16_t)> read,uint16_t pc);
void gcdsp_extPending_add(gcdsp_extPending* p,uint16_t reg,uint16_t val);
void gcdsp_extPending_commit(gcdsp_extPending p,gcdsp_CPU* c);
void gcdsp_CPU_extStep(gcdsp_CPU* c,int64_t n,bool byIndex);
gcdsp_extPending gcdsp_CPU_extBegin(gcdsp_CPU* c,uint16_t ext);
int64_t gcdsp_CPU_mul16(gcdsp_CPU* c,uint16_t a,uint16_t b);
int64_t gcdsp_CPU_mulx16(gcdsp_CPU* c,uint16_t a,uint16_t b,bool aHigh,bool bHigh);
int64_t gcdsp_CPU_mulxProd(gcdsp_CPU* c,uint16_t op);
void gcdsp_CPU_setProd(gcdsp_CPU* c,int64_t v);
uint16_t gcdsp_CPU_imem(gcdsp_CPU* c,uint16_t a);
uint16_t gcdsp_CPU_dataRead(gcdsp_CPU* c,uint16_t a);
void gcdsp_CPU_dataWrite(gcdsp_CPU* c,uint16_t a,uint16_t v);
uint16_t gcdsp_CPU_getReg(gcdsp_CPU* c,uint16_t r);
void gcdsp_CPU_setReg(gcdsp_CPU* c,uint16_t r,uint16_t v);
void gcdsp_CPU_setRegExtend(gcdsp_CPU* c,uint16_t r,uint16_t v);
void gcdsp_CPU_arInc(gcdsp_CPU* c,int64_t n);
void gcdsp_CPU_arDec(gcdsp_CPU* c,int64_t n);
void gcdsp_CPU_arAdd(gcdsp_CPU* c,int64_t n,int16_t ix);
void gcdsp_CPU_arSub(gcdsp_CPU* c,int64_t n,int16_t ix);
void gcdsp_CPU_push(gcdsp_CPU* c,uint16_t reg,uint16_t v);
uint16_t gcdsp_CPU_pop(gcdsp_CPU* c,uint16_t reg);
void gcdsp_CPU_startLoop(gcdsp_CPU* c,uint16_t start,uint16_t end,uint16_t count);
void gcdsp_CPU_serviceLoops(gcdsp_CPU* c,uint16_t execAddr,bool branched);
void gcdsp_CPU_shiftAcc(gcdsp_CPU* c,int64_t r,bool arith,int64_t amt);
int64_t gcdsp_shiftAmount7(uint16_t v);
void gcdsp_CPU_setArithFlags(gcdsp_CPU* c,int64_t n);
void gcdsp_CPU_aluAddSub(gcdsp_CPU* c,int64_t d,int64_t b,bool sub_);
void gcdsp_CPU_setTestFlags(gcdsp_CPU* c,int64_t n);
void gcdsp_CPU_setLogicFlags(gcdsp_CPU* c,int64_t n);
void gcdsp_CPU_subFlags(gcdsp_CPU* c,int64_t a,int64_t b);
bool gcdsp_CPU_cond(gcdsp_CPU* c,uint16_t cc);
std::tuple<uint32_t,Error> gc_Machine_LoadDOL(gc_Machine* m);
void gc_Machine_Poke(gc_Machine* m,uint32_t addr,uint32_t val);
void gc_Machine_PoisonLowMem(gc_Machine* m);
std::string gc_Machine_String(gc_Machine* m);
uint8_t gc_Machine_ReadVirt8(gc_Machine* m,uint32_t addr);
uint32_t gc_Machine_ReadVirt32(gc_Machine* m,uint32_t addr);
uint64_t gc_Machine_VIField(gc_Machine* m);
Slice<uint32_t> gc_Machine_Backtrace(gc_Machine* m);
Slice<uint32_t> gc_Machine_backtraceFrom(gc_Machine* m,uint32_t sp);
bool gc_validSP(uint32_t a);
uint32_t gc_Machine_stackWord(gc_Machine* m,uint32_t ea);
std::string gc_Machine_BacktraceString(gc_Machine* m);
std::string gc_Machine_ThreadsString(gc_Machine* m);
std::tuple<image_RGBA*,Error> gc_Machine_RenderXFB(gc_Machine* m);
std::tuple<int64_t,int64_t> gc_Machine_xfbSize(gc_Machine* m);
std::tuple<image_RGBA*,Error> gc_Machine_RenderEFB(gc_Machine* m);
void gc_setPix(image_RGBA* img,int64_t x,int64_t y,uint8_t r,uint8_t g,uint8_t b);
std::tuple<uint8_t,uint8_t,uint8_t> gc_yuv2rgb(uint8_t y,uint8_t cb,uint8_t cr);
uint8_t gc_clamp8(double v);
std::string gc_Machine_IntrState(gc_Machine* m);
std::string gc_Machine_RegString(gc_Machine* m);
uint32_t gc_mi_read(gc_mi* d,gc_Machine* m,uint32_t off,int64_t size);
void gc_mi_write(gc_mi* d,gc_Machine* m,uint32_t off,uint32_t v,int64_t size);
uint32_t gc_ai_read(gc_ai* d,gc_Machine* m,uint32_t off,int64_t size);
void gc_ai_write(gc_ai* d,gc_Machine* m,uint32_t off,uint32_t v,int64_t size);
uint32_t gc_cp_read(gc_cp* d,gc_Machine* m,uint32_t off,int64_t size);
void gc_cp_write(gc_cp* d,gc_Machine* m,uint32_t off,uint32_t v,int64_t size);
uint32_t gc_pe_read(gc_pe* d,gc_Machine* m,uint32_t off,int64_t size);
void gc_pe_write(gc_pe* d,gc_Machine* m,uint32_t off,uint32_t v,int64_t size);
void gc_pe_setFinish(gc_pe* d,gc_Machine* m);
void gc_pe_setToken(gc_pe* d,gc_Machine* m,uint16_t tok,bool raise);
void gc_wgPipe_push(gc_wgPipe* w,gc_Machine* m,Slice<uint8_t> b);
void gc_wgPipe_write8(gc_wgPipe* w,gc_Machine* m,uint8_t v);
void gc_wgPipe_write16(gc_wgPipe* w,gc_Machine* m,uint16_t v);
void gc_wgPipe_write32(gc_wgPipe* w,gc_Machine* m,uint32_t v);
void gc_di_init(gc_di* d);
uint32_t gc_di_read(gc_di* d,gc_Machine* m,uint32_t off,int64_t size);
void gc_di_write(gc_di* d,gc_Machine* m,uint32_t off,uint32_t v,int64_t size);
void gc_di_exec(gc_di* d,gc_Machine* m);
void gc_Machine_tickDI(gc_Machine* m);
void gc_di_complete(gc_di* d,gc_Machine* m);
void gc_di_raiseError(gc_di* d,gc_Machine* m);
void gc_Machine_diRefreshIRQ(gc_Machine* m);
int64_t gc_Apploader_Body(gc_Apploader a);
std::tuple<Slice<uint8_t>,Error> gc_Disc_ApploaderCode(gc_Disc* d);
std::tuple<gc_DOL*,Error> gc_Disc_DOL(gc_Disc* d);
uint32_t gc_be32(Slice<uint8_t> b);
std::string gc_cstr(Slice<uint8_t> b);
std::string gc_Segment_Name(gc_Segment s);
int64_t gc_dolLength(Slice<uint8_t> hdr);
std::tuple<gc_DOL*,Error> gc_ParseDOL(Slice<uint8_t> b);
void gc_DOL_Load(gc_DOL* d,std::function<void(uint32_t,Slice<uint8_t>)> write);
std::tuple<uint32_t,Slice<uint8_t>> gc_DOL_Flat(gc_DOL* d);
bool gc_DOL_Text(gc_DOL* d,uint32_t addr);
void gc_dsp_init(gc_dsp* d);
uint32_t gc_dsp_read(gc_dsp* d,gc_Machine* m,uint32_t off,int64_t size);
void gc_dsp_write(gc_dsp* d,gc_Machine* m,uint32_t off,uint32_t v,int64_t size);
void gc_dsp_consumeMail(gc_dsp* d,gc_Machine* m);
void gc_dsp_advanceBoot(gc_dsp* d,gc_Machine* m);
void gc_dsp_post(gc_dsp* d,gc_Machine* m,uint32_t v);
void gc_dsp_runARAMDMA(gc_dsp* d,gc_Machine* m);
void gc_Machine_tickAID(gc_Machine* m);
void gc_Machine_dspRefreshIRQ(gc_Machine* m);
void gc_dsp_startCore(gc_dsp* d,gc_Machine* m);
uint16_t gc_dspBus_HWRead(gc_dspBus b,uint16_t a);
void gc_dspBus_HWWrite(gc_dspBus b,uint16_t a,uint16_t v);
uint16_t gc_dsp_hwRead(gc_dsp* d,gc_Machine* m,uint16_t a);
void gc_dsp_hwWrite(gc_dsp* d,gc_Machine* m,uint16_t a,uint16_t v);
uint8_t gc_dsp_aramByte(gc_dsp* d,gc_Machine* m,uint32_t addr);
uint16_t gc_dsp_accCurrentSample(gc_dsp* d,gc_Machine* m);
uint16_t gc_dsp_accReadRaw(gc_dsp* d,gc_Machine* m);
void gc_dsp_accWriteRaw(gc_dsp* d,gc_Machine* m,uint16_t v);
uint16_t gc_dsp_accReadSample(gc_dsp* d,gc_Machine* m);
void gc_dsp_runMemDMA(gc_dsp* d,gc_Machine* m,uint16_t lenBytes);
void gc_Machine_tickDSP(gc_Machine* m);
void gc_exi_init(gc_exi* e);
uint32_t gc_exi_read(gc_exi* e,gc_Machine* m,uint32_t off,int64_t size);
void gc_exi_write(gc_exi* e,gc_Machine* m,uint32_t off,uint32_t v,int64_t size);
void gc_exi_refreshIRQ(gc_exi* e,gc_Machine* m);
void gc_exi_transfer(gc_exi* e,gc_Machine* m,uint32_t ch);
void gc_exi_dmaTransfer(gc_exi* e,gc_Machine* m,uint32_t ch,bool asWrite);
uint32_t gc_rtcDevAddr(uint32_t cmd);
void gc_exi_rtcTransfer(gc_exi* e,gc_Machine* m,gc_exiChannel* c);
uint16_t gc_beU16(Slice<uint8_t> b);
uint32_t gc_beU32(Slice<uint8_t> b);
void gc_writeBE16(Slice<uint8_t> b,uint16_t v);
void gc_writeBE32(Slice<uint8_t> b,uint32_t v);
std::tuple<gc_FST*,Error> gc_ParseFST(Slice<uint8_t> b);
Slice<gc_File> gc_FST_Files(gc_FST* f);
std::tuple<gc_File,bool> gc_FST_ByPath(gc_FST* f,std::string path);
std::tuple<gc_File,int64_t,bool> gc_FST_ByOffset(gc_FST* f,int64_t off);
Error gc_FST_Validate(gc_FST* f,int64_t discSize);
void gc_gpu_xfStore(gc_gpu* g,int64_t addr,uint32_t val);
void gc_gpu_feed(gc_gpu* g,gc_Machine* m,Slice<uint8_t> b);
bool gc_gpu_gxCmd(gc_gpu* g,gc_Machine* m,uint8_t op,Slice<uint32_t> words);
bool gc_gpu_step(gc_gpu* g,gc_Machine* m);
std::string gc_GXName(uint8_t op);
std::string gc_primName(uint8_t prim);
void gc_gpu_callDisplayList(gc_gpu* g,gc_Machine* m,uint32_t addr,uint32_t size);
void gc_gpu_loadBP(gc_gpu* g,gc_Machine* m,uint8_t reg,uint32_t data);
uint16_t gc_be16(Slice<uint8_t> b);
float gc_clipPlane_dist(gc_clipPlane p,gc_clipVertex v);
uint8_t gc_outcode(gc_clipVertex v);
Slice<gc_clipVertex> gc_clipByPlane(Slice<gc_clipVertex> out,Slice<gc_clipVertex> in,gc_clipPlane p);
std::tuple<Slice<gc_clipVertex>,Slice<gc_clipVertex>> gc_clipTriangle(Slice<gc_clipVertex> dst,Slice<gc_clipVertex> scratch,std::array<gc_clipVertex,3> tri);
std::tuple<Slice<gc_clipVertex>,Slice<gc_clipVertex>> gc_gpu_clipAndSetup(gc_gpu* g,gc_Machine* m,Slice<gc_clipVertex> dst,Slice<gc_clipVertex> scratch,gc_clipVertex v0,gc_clipVertex v1,gc_clipVertex v2);
void gc_gpu_ensureEFB(gc_gpu* g);
std::tuple<uint8_t,uint8_t,uint8_t,uint8_t> gc_gpu_clearColor(gc_gpu* g);
void gc_gpu_clearEFB(gc_gpu* g);
uint32_t gc_gpu_efbAt(gc_gpu* g,int64_t x,int64_t y);
uint32_t gc_gpu_zAt(gc_gpu* g,int64_t x,int64_t y);
void gc_gpu_copyDisplay(gc_gpu* g,gc_Machine* m,uint32_t params);
void gc_gpu_copyTexture(gc_gpu* g,gc_Machine* m,uint32_t params);
std::tuple<uint8_t,uint8_t,uint8_t,uint8_t> gc_gpu_efbBox2(gc_gpu* g,int64_t x,int64_t y);
std::tuple<int64_t,int64_t> gc_copyTexelOffset(int64_t stride,int64_t bw,int64_t bh,int64_t tileBytes,int64_t x,int64_t y);
bool gc_gpu_encodeCopyTexel(gc_gpu* g,gc_Machine* m,uint32_t dst,int64_t stride,int64_t format,int64_t x,int64_t y,uint8_t r,uint8_t gg,uint8_t b,uint8_t a);
void gc_Machine_texWriteByte(gc_Machine* m,uint32_t addr,uint8_t v);
uint32_t gc_packRGBA(uint8_t r,uint8_t g,uint8_t b,uint8_t a);
std::tuple<uint8_t,uint8_t,uint8_t> gc_unpackRGB(uint32_t px);
uint8_t gc_luma(uint8_t r,uint8_t g,uint8_t b);
uint8_t gc_chromaB(uint8_t r,uint8_t g,uint8_t b);
uint8_t gc_chromaR(uint8_t r,uint8_t g,uint8_t b);
gc_attrLayout gc_gpu_layout(gc_gpu* g,int64_t vat);
Slice<uint8_t> gc_gpu_attrData(gc_gpu* g,gc_Machine* m,Slice<uint8_t> v,int64_t off,uint32_t desc,int64_t attr,int64_t size);
void gc_gpu_drawPrimitive(gc_gpu* g,gc_Machine* m,uint32_t prim,int64_t vat,int64_t vsize,Slice<uint8_t> data);
std::tuple<float,float> gc_readTexCoord(Slice<uint8_t> b,gc_texAttr* ta);
std::tuple<float,float,float> gc_readNormal(Slice<uint8_t> b,uint32_t format);
std::tuple<float,float,float> gc_gpu_readPos(gc_gpu* g,Slice<uint8_t> b,gc_attrLayout lay);
float gc_readComponent(Slice<uint8_t> b,uint32_t format);
std::tuple<uint8_t,uint8_t,uint8_t,uint8_t> gc_readColor(Slice<uint8_t> b,uint32_t comp);
uint8_t gc_expand4(uint16_t v);
uint8_t gc_expand5(uint16_t v);
uint8_t gc_expand6(uint16_t v);
void gc_gpu_rasterPrimitive(gc_gpu* g,gc_Machine* m,uint32_t prim,Slice<gc_clipVertex> v);
std::array<float,4> gc_gpu_xfColor(gc_gpu* g,int64_t addr);
uint32_t gc_lightMask(uint32_t ctl);
std::array<std::array<uint8_t,4>,2> gc_gpu_rasChannel(gc_gpu* g,gc_Machine* m,uint8_t vr,uint8_t vg,uint8_t vb,uint8_t va,float ex,float ey,float ez,float nx,float ny,float nz);
uint8_t gc_chanU8(float v);
std::array<float,4> gc_gpu_evalChannel(gc_gpu* g,gc_Machine* m,uint32_t ctl,bool isAlpha,std::array<float,4> vtx,std::array<float,4> amb,std::array<float,4> mat,float ex,float ey,float ez,float nx,float ny,float nz);
std::tuple<float,float,float> gc_normalize3f(float x,float y,float z);
std::tuple<float,float,float> gc_gpu_normalToEye(gc_gpu* g,int64_t mtxIdx,float nx,float ny,float nz);
void gc_gpu_ensureRaster(gc_gpu* g);
void gc_perspTexCoords(float b0,float b1,float b2,const gc_screenVertex& v0,const gc_screenVertex& v1,const gc_screenVertex& v2,std::array<gc_texCoord,8>* out);
void gc_gpu_init(gc_gpu* g);
std::tuple<int64_t,int64_t,int64_t,int64_t> gc_gpu_scissorBox(gc_gpu* g);
std::tuple<gc_rasterTri,bool> gc_gpu_setupTri(gc_gpu* g,gc_screenVertex v0,gc_screenVertex v1,gc_screenVertex v2);
void gc_gpu_fillTri(gc_gpu* g,gc_Machine* m,gc_tevState* tev,gc_rasterTri* t,int64_t yLo,int64_t yHi,gc_rstats* st);
void gc_gpu_mergeStats(gc_gpu* g,gc_rstats* st);
bool gc_gpu_cullTest(gc_gpu* g,float area);
bool gc_gpu_cullTestExperiment(gc_gpu* g,float area);
bool gc_depthCompare(uint32_t z,uint32_t buf,int64_t comp);
float gc_edge(float ax,float ay,float bx,float by,float cx,float cy);
float gc_min3(float a,float b,float c);
float gc_max3(float a,float b,float c);
float gc_minf(float a,float b);
float gc_maxf(float a,float b);
std::tuple<float,float,float,float> gc_tevColorReg(std::array<uint32_t,2> w);
int32_t gc_sext11(uint32_t v);
std::tuple<uint8_t,uint8_t,uint8_t,uint8_t,bool> gc_gpu_shade(gc_gpu* g,gc_Machine* m,gc_tevState* t,std::array<std::array<uint8_t,4>,2>* rasCol,std::array<gc_texCoord,8>* tc);
std::tuple<uint8_t,uint8_t,uint8_t,uint8_t,bool> gc_gpu_shade_reference(gc_gpu* g,gc_Machine* m,gc_tevState* t,std::array<std::array<uint8_t,4>,2>* rasCol,std::array<gc_texCoord,8>* tc);
uint8_t gc_toU8(float f);
std::array<float,4> gc_rasSelect(uint32_t sel,std::array<std::array<uint8_t,4>,2>* col);
std::tuple<std::array<float,3>,std::array<float,3>,std::array<float,3>> gc_combineColor(uint32_t cc,std::array<std::array<float,4>,4>* reg,std::array<float,4> tex,std::array<float,4> ras,std::array<float,3> konst);
float gc_combineAlpha(uint32_t ac,std::array<std::array<float,4>,4>* reg,std::array<float,4> tex,std::array<float,4> ras,float konst,std::array<float,3> ca,std::array<float,3> cb);
int64_t gc_b2i(bool b);
bool gc_tevCompare(int64_t sel,std::array<float,3> ca,std::array<float,3> cb,float aa,float ab,int64_t ch);
float gc_tevFormula(float a,float b,float c,float d,int64_t bias,bool sub_,int64_t scale,bool clamp);
std::array<float,3> gc_colorArg(int64_t code,std::array<std::array<float,4>,4>* reg,std::array<float,4> tex,std::array<float,4> ras,std::array<float,3> konst);
float gc_alphaArg(int64_t code,std::array<std::array<float,4>,4>* reg,std::array<float,4> tex,std::array<float,4> ras,float konst);
std::array<float,3> gc_gpu_konstColor(gc_gpu* g,int64_t stage);
float gc_gpu_konstAlpha(gc_gpu* g,int64_t stage);
std::array<int64_t,4> gc_gpu_swapTable(gc_gpu* g,int64_t t);
std::array<float,4> gc_swizzle(std::array<int64_t,4> t,std::array<float,4> v);
int64_t gc_gpu_kSel(gc_gpu* g,int64_t stage,bool alpha);
bool gc_gpu_alphaTest(gc_gpu* g,uint8_t a);
bool gc_alphaCompare(uint8_t a,uint8_t ref,int64_t comp);
uint32_t gc_gpu_blend(gc_gpu* g,uint32_t dst,uint8_t sr,uint8_t sg,uint8_t sb,uint8_t sa);
uint32_t gc_gpu_applyUpdates(gc_gpu* g,uint32_t cm,uint32_t dst,uint8_t r,uint8_t gg,uint8_t b,uint8_t a);
float gc_blendFactor(int64_t code,float chanOther,float sa,float da);
float gc_clampf(float v);
bool gc_texCanHalt(gc_texState* tx);
gc_tevState gc_gpu_tevstate(gc_gpu* g);
gc_tevState gc_gpu_tevstate_reference(gc_gpu* g);
int64_t gc_gpu_texGenCount(gc_gpu* g);
int64_t gc_gpu_texMtxRow(gc_gpu* g,int64_t i);
gc_texCoord gc_gpu_genTexCoord(gc_gpu* g,gc_Machine* m,int64_t i,int64_t mtxRow,float mx,float my,float mz,float nx,float ny,float nz,std::array<gc_texCoord,8>* vtc,std::array<std::array<uint8_t,4>,2>* col);
std::array<float,3> gc_normalize3(std::array<float,3> v);
gc_texState gc_gpu_texSetup(gc_gpu* g,int64_t i);
void gc_gpu_loadTlut(gc_gpu* g,gc_Machine* m,uint32_t data);
std::tuple<uint8_t,uint8_t,uint8_t,uint8_t> gc_gpu_tlutColor(gc_gpu* g,gc_Machine* m,const gc_texState& tx,int64_t idx);
std::tuple<uint8_t,uint8_t,uint8_t,uint8_t> gc_gpu_sampleTexmap(gc_gpu* g,gc_Machine* m,gc_texState* tx,float s,float t);
int64_t gc_wrapCoord(int64_t v,int64_t size,int64_t mode);
std::tuple<uint8_t,uint8_t,uint8_t,uint8_t> gc_gpu_decodeTexel(gc_gpu* g,gc_Machine* m,const gc_texState& tx,int64_t x,int64_t y);
std::tuple<int64_t,int64_t> gc_tileByteOffset(int64_t width,int64_t bw,int64_t bh,int64_t x,int64_t y);
uint16_t gc_gpu_tileHalf(gc_gpu* g,gc_Machine* m,uint32_t base,int64_t width,int64_t bw,int64_t bh,int64_t x,int64_t y);
uint8_t gc_gpu_tileByte(gc_gpu* g,gc_Machine* m,uint32_t base,int64_t width,int64_t bw,int64_t bh,int64_t x,int64_t y);
uint8_t gc_gpu_tileNibble(gc_gpu* g,gc_Machine* m,uint32_t base,int64_t width,int64_t bw,int64_t bh,int64_t x,int64_t y);
std::tuple<uint8_t,uint8_t,uint8_t,uint8_t> gc_gpu_decodeRGBA8(gc_gpu* g,gc_Machine* m,uint32_t base,int64_t width,int64_t x,int64_t y);
std::tuple<uint8_t,uint8_t,uint8_t,uint8_t> gc_gpu_decodeCMPR(gc_gpu* g,gc_Machine* m,uint32_t base,int64_t width,int64_t x,int64_t y);
std::tuple<uint8_t,uint8_t,uint8_t,uint8_t> gc_decodeRGB565(uint16_t v);
std::tuple<uint8_t,uint8_t,uint8_t,uint8_t> gc_decodeRGB5A3(uint16_t v);
uint8_t gc_expand3(uint16_t v);
uint8_t gc_lerp2(uint8_t a,uint8_t b);
uint8_t gc_avg2(uint8_t a,uint8_t b);
uint8_t gc_Machine_texByte(gc_Machine* m,uint32_t addr);
bool gc_Machine_TextureBound(gc_Machine* m,int64_t i);
std::tuple<image_RGBA*,Error> gc_Machine_DumpTexture(gc_Machine* m,int64_t i);
int64_t gc_componentBytes(uint32_t format);
int64_t gc_colorBytes(uint32_t comp);
int64_t gc_normalIndexCount(uint32_t g0);
int64_t gc_gpu_vertexSize(gc_gpu* g,int64_t vat);
float gc_gpu_xfFloat(gc_gpu* g,int64_t addr);
gc_clipVertex gc_lerpClip(gc_clipVertex a,gc_clipVertex b,float t);
std::tuple<float,float,float> gc_gpu_transform(gc_gpu* g,int64_t mtxIdx,float mx,float my,float mz);
std::tuple<float,float,float> gc_gpu_eyePos(gc_gpu* g,int64_t mtxIdx,float mx,float my,float mz);
std::tuple<float,float,float> gc_gpu_project(gc_gpu* g,float ex,float ey,float ez);
std::tuple<float,float,float,float> gc_gpu_clipPos(gc_gpu* g,float ex,float ey,float ez);
std::tuple<float,float,float> gc_gpu_toScreen(gc_gpu* g,float cx,float cy,float cz,float cw);
gc_idleSnap gc_Machine_snapshotCPU(gc_Machine* m);
bool gc_Machine_idleStep(gc_Machine* m,uint32_t pc);
uint64_t gc_Machine_idleDeadline(gc_Machine* m);
void gc_Machine_idleSkip(gc_Machine* m,uint64_t n);
std::tuple<uint64_t,uint64_t> gc_Machine_IdleStats(gc_Machine* m);
void gc_Machine_SetIdleSkip(gc_Machine* m,bool on);
void gc_Machine_setupLowMem(gc_Machine* m);
void gc_Machine_setupState(gc_Machine* m);
bool gc_Machine_handleSyscall(gc_Machine* m,gekko_CPU* c);
std::tuple<uint32_t,Error> gc_Machine_RunApploader(gc_Machine* m);
Error gc_Machine_call(gc_Machine* m,uint32_t fn);
void gc_Machine_serviceReport(gc_Machine* m);
std::string gc_Machine_readCString(gc_Machine* m,uint32_t addr);
std::string gc_trimReport(std::string s);
std::tuple<gc_Machine*,Error> gc_NewMachine(gc_Disc* disc);
gc_Disc* gc_Machine_Disc(gc_Machine* m);
uint8_t gc_Machine_Read8(gc_Machine* m,uint32_t a);
uint16_t gc_Machine_Read16(gc_Machine* m,uint32_t a);
uint32_t gc_Machine_Read32(gc_Machine* m,uint32_t a);
void gc_Machine_Write8(gc_Machine* m,uint32_t a,uint8_t v);
void gc_Machine_Write16(gc_Machine* m,uint32_t a,uint16_t v);
void gc_Machine_Write32(gc_Machine* m,uint32_t a,uint32_t v);
uint32_t gc_Machine_Fetch32(gc_Machine* m,uint32_t a);
uint32_t gc_Machine_Fetch32_reference(gc_Machine* m,uint32_t a);
uint32_t gc_Machine_regRead(gc_Machine* m,uint32_t a,int64_t size);
void gc_Machine_regWrite(gc_Machine* m,uint32_t a,uint32_t v,int64_t size);
void gc_Machine_dmaToRAM(gc_Machine* m,uint32_t addr,Slice<uint8_t> data);
uint32_t gc_phys(uint32_t a);
uint32_t gc_Machine_ram32(gc_Machine* m,uint32_t a);
void gc_Machine_setRAM32(gc_Machine* m,uint32_t a,uint32_t v);
void gc_Machine_readWatch(gc_Machine* m,uint32_t a,uint32_t v);
void gc_Machine_writeWatch(gc_Machine* m,uint32_t a,uint32_t v);
Slice<std::string> gc_Machine_Census(gc_Machine* m);
std::array<uint64_t,256> gc_Machine_GPUCensus(gc_Machine* m);
uint32_t gc_pi_read(gc_pi* p,gc_Machine* m,uint32_t off,int64_t size);
void gc_pi_write(gc_pi* p,gc_Machine* m,uint32_t off,uint32_t v,int64_t size);
void gc_Machine_raiseInt(gc_Machine* m,int64_t cause);
void gc_Machine_clearInt(gc_Machine* m,int64_t cause);
void gc_Machine_updateIRQ(gc_Machine* m);
gc_profCounters gc_Machine_profCounters(gc_Machine* m);
time_Time gc_Machine_profStart(gc_Machine* m);
void gc_Machine_profEnd(gc_Machine* m,int64_t bucket,time_Time t);
void gc_Machine_profFIFOEnter(gc_Machine* m);
void gc_Machine_profFIFOExit(gc_Machine* m);
void gc_Machine_profRunEnter(gc_Machine* m);
void gc_Machine_profRunExit(gc_Machine* m);
void gc_Machine_profFrame(gc_Machine* m);
gc_FrameProfile gc_Machine_FrameProfile(gc_Machine* m);
void gc_Machine_SetProfile(gc_Machine* m,bool on);
std::string gc_Result_String(gc_Result r);
void gc_Machine_SetSpinDetect(gc_Machine* m,bool on);
void gc_Machine_SetBreakpoint(gc_Machine* m,uint32_t vaddr);
void gc_Machine_ClearBreakpoints(gc_Machine* m);
gc_Result gc_Machine_RunStopAfterGXCommand(gc_Machine* m,int64_t n,uint64_t maxSteps);
int64_t gc_Machine_GXCommandCount(gc_Machine* m);
gc_Result gc_Machine_RunFields(gc_Machine* m,int64_t n,uint64_t budget);
gc_Result gc_Machine_Run(gc_Machine* m,uint64_t maxSteps);
void gc_si_connectPad(gc_si* d,int64_t port);
uint32_t gc_si_read(gc_si* d,gc_Machine* m,uint32_t off,int64_t size);
uint32_t gc_si_readReg(gc_si* d,uint32_t r,int64_t size);
void gc_si_write(gc_si* d,gc_Machine* m,uint32_t off,uint32_t v,int64_t size);
void gc_si_writeComCSR(gc_si* d,gc_Machine* m,uint32_t v);
void gc_si_startTransfer(gc_si* d,gc_Machine* m);
Slice<uint8_t> gc_padPort_respond(gc_padPort* p,uint8_t cmd,int64_t inlen);
std::tuple<uint32_t,uint32_t> gc_padPort_status(gc_padPort* p);
void gc_Machine_tickSI(gc_Machine* m);
void gc_Machine_siRefreshIRQ(gc_Machine* m);
void gc_Machine_SetController(gc_Machine* m,int64_t port,bool connected);
void gc_Machine_SetPadButtons(gc_Machine* m,int64_t port,uint16_t buttons);
uint16_t gc_Machine_PadButtons(gc_Machine* m,int64_t port);
void gc_Machine_SetPadStick(gc_Machine* m,int64_t port,uint8_t x,uint8_t y);
std::tuple<uint8_t,uint8_t> gc_Machine_PadStick(gc_Machine* m,int64_t port);
std::tuple<uint16_t,bool> gc_PadButton(std::string name);
Slice<std::string> gc_PadButtonNames();
uint32_t gc_si_iobufRead(gc_si* d,uint32_t o,int64_t size);
void gc_si_iobufWrite(gc_si* d,uint32_t o,uint32_t v,int64_t size);
void gc_putWord(Slice<uint8_t> b,int64_t o,uint32_t v);
void gc_Machine_SetStackProfile(gc_Machine* m,uint64_t every,int64_t depth);
void gc_Machine_StopStackProfile(gc_Machine* m);
void gc_Machine_sampleStack(gc_Machine* m,uint32_t pc);
std::tuple<Slice<gc_StackSample>,int64_t> gc_Machine_StackProfile(gc_Machine* m);
std::tuple<Slice<gc_StackSample>,int64_t> gc_Machine_StackProfileByCaller(gc_Machine* m,int64_t depth);
std::string gc_Machine_StackProfileString(gc_Machine* m,int64_t n);
void gc_vi_init(gc_vi* v);
uint32_t gc_vi_read(gc_vi* v,gc_Machine* m,uint32_t off,int64_t size);
void gc_vi_write(gc_vi* v,gc_Machine* m,uint32_t off,uint32_t val,int64_t size);
uint32_t gc_composeHalfword(uint32_t reg,uint32_t val,uint32_t r,int64_t size);
uint32_t gc_halfword(uint32_t reg,uint32_t r,int64_t size);
uint32_t gc_vi_XFBAddr(gc_vi* v);
void gc_Machine_tickVI(gc_Machine* m);
void gc_Machine_viRefreshIRQ(gc_Machine* m);
constexpr int64_t gekko_CacheLine=32ULL;
constexpr int64_t gekko_LockedCacheBase=3758096384ULL;
constexpr int64_t gekko_LockedCacheSize=16384ULL;
constexpr int64_t gekko_crLT=8ULL;
constexpr int64_t gekko_crGT=4ULL;
constexpr int64_t gekko_crEQ=2ULL;
constexpr int64_t gekko_crSO=1ULL;
constexpr int64_t gekko_boNoCond=16ULL;
constexpr int64_t gekko_boCondSet=8ULL;
constexpr int64_t gekko_boNoDec=4ULL;
constexpr int64_t gekko_boCTRZero=2ULL;
std::array<std::string,4> gekko_crBitName=std::array<std::string,4>{std::string("lt",2),std::string("gt",2),std::string("eq",2),std::string("so",2)};
std::array<std::string,4> gekko_brTrue=std::array<std::string,4>{std::string("blt",3),std::string("bgt",3),std::string("beq",3),std::string("bso",3)};
std::array<std::string,4> gekko_brFalse=std::array<std::string,4>{std::string("bge",3),std::string("ble",3),std::string("bne",3),std::string("bns",3)};
Map<uint32_t,std::string> gekko_crLogicName=Map<uint32_t,std::string>{{cast<uint32_t>(33ULL),std::string("crnor",5)},{cast<uint32_t>(129ULL),std::string("crandc",6)},{cast<uint32_t>(193ULL),std::string("crxor",5)},{cast<uint32_t>(225ULL),std::string("crnand",6)},{cast<uint32_t>(257ULL),std::string("crand",5)},{cast<uint32_t>(289ULL),std::string("creqv",5)},{cast<uint32_t>(417ULL),std::string("crorc",5)},{cast<uint32_t>(449ULL),std::string("cror",4)}};
Map<uint32_t,bool> gekko_oeForm=Map<uint32_t,bool>{{cast<uint32_t>(8ULL),true},{cast<uint32_t>(10ULL),true},{cast<uint32_t>(40ULL),true},{cast<uint32_t>(104ULL),true},{cast<uint32_t>(136ULL),true},{cast<uint32_t>(138ULL),true},{cast<uint32_t>(200ULL),true},{cast<uint32_t>(202ULL),true},{cast<uint32_t>(232ULL),true},{cast<uint32_t>(234ULL),true},{cast<uint32_t>(235ULL),true},{cast<uint32_t>(266ULL),true},{cast<uint32_t>(459ULL),true},{cast<uint32_t>(491ULL),true}};
Map<uint32_t,std::string> gekko_psAForm=Map<uint32_t,std::string>{{cast<uint32_t>(10ULL),std::string("ps_sum0",7)},{cast<uint32_t>(11ULL),std::string("ps_sum1",7)},{cast<uint32_t>(12ULL),std::string("ps_muls0",8)},{cast<uint32_t>(13ULL),std::string("ps_muls1",8)},{cast<uint32_t>(14ULL),std::string("ps_madds0",9)},{cast<uint32_t>(15ULL),std::string("ps_madds1",9)},{cast<uint32_t>(18ULL),std::string("ps_div",6)},{cast<uint32_t>(20ULL),std::string("ps_sub",6)},{cast<uint32_t>(21ULL),std::string("ps_add",6)},{cast<uint32_t>(23ULL),std::string("ps_sel",6)},{cast<uint32_t>(24ULL),std::string("ps_res",6)},{cast<uint32_t>(25ULL),std::string("ps_mul",6)},{cast<uint32_t>(26ULL),std::string("ps_rsqrte",9)},{cast<uint32_t>(28ULL),std::string("ps_msub",7)},{cast<uint32_t>(29ULL),std::string("ps_madd",7)},{cast<uint32_t>(30ULL),std::string("ps_nmsub",8)},{cast<uint32_t>(31ULL),std::string("ps_nmadd",8)}};
constexpr int64_t gekko_shapeAB=0ULL;
constexpr int64_t gekko_shapeAC=1ULL;
constexpr int64_t gekko_shapeACB=2ULL;
constexpr int64_t gekko_shapeB=3ULL;
Map<uint32_t,int64_t> gekko_psShape=Map<uint32_t,int64_t>{{cast<uint32_t>(10ULL),cast<int64_t>(2ULL)},{cast<uint32_t>(11ULL),cast<int64_t>(2ULL)},{cast<uint32_t>(12ULL),cast<int64_t>(1ULL)},{cast<uint32_t>(13ULL),cast<int64_t>(1ULL)},{cast<uint32_t>(14ULL),cast<int64_t>(2ULL)},{cast<uint32_t>(15ULL),cast<int64_t>(2ULL)},{cast<uint32_t>(18ULL),cast<int64_t>(0ULL)},{cast<uint32_t>(20ULL),cast<int64_t>(0ULL)},{cast<uint32_t>(21ULL),cast<int64_t>(0ULL)},{cast<uint32_t>(23ULL),cast<int64_t>(2ULL)},{cast<uint32_t>(24ULL),cast<int64_t>(3ULL)},{cast<uint32_t>(26ULL),cast<int64_t>(3ULL)},{cast<uint32_t>(25ULL),cast<int64_t>(1ULL)},{cast<uint32_t>(28ULL),cast<int64_t>(2ULL)},{cast<uint32_t>(29ULL),cast<int64_t>(2ULL)},{cast<uint32_t>(30ULL),cast<int64_t>(2ULL)},{cast<uint32_t>(31ULL),cast<int64_t>(2ULL)}};
Map<uint32_t,std::string> gekko_psX=Map<uint32_t,std::string>{{cast<uint32_t>(0ULL),std::string("ps_cmpu0",8)},{cast<uint32_t>(32ULL),std::string("ps_cmpo0",8)},{cast<uint32_t>(64ULL),std::string("ps_cmpu1",8)},{cast<uint32_t>(96ULL),std::string("ps_cmpo1",8)},{cast<uint32_t>(40ULL),std::string("ps_neg",6)},{cast<uint32_t>(72ULL),std::string("ps_mr",5)},{cast<uint32_t>(136ULL),std::string("ps_nabs",7)},{cast<uint32_t>(264ULL),std::string("ps_abs",6)},{cast<uint32_t>(528ULL),std::string("ps_merge00",10)},{cast<uint32_t>(560ULL),std::string("ps_merge01",10)},{cast<uint32_t>(592ULL),std::string("ps_merge10",10)},{cast<uint32_t>(624ULL),std::string("ps_merge11",10)},{cast<uint32_t>(1014ULL),std::string("dcbz_l",6)}};
constexpr int64_t gekko_VecReset=256ULL;
constexpr int64_t gekko_VecMachineCheck=512ULL;
constexpr int64_t gekko_VecDSI=768ULL;
constexpr int64_t gekko_VecISI=1024ULL;
constexpr int64_t gekko_VecExternal=1280ULL;
constexpr int64_t gekko_VecAlignment=1536ULL;
constexpr int64_t gekko_VecProgram=1792ULL;
constexpr int64_t gekko_VecFPUnavail=2048ULL;
constexpr int64_t gekko_VecDecrementer=2304ULL;
constexpr int64_t gekko_VecSyscall=3072ULL;
constexpr int64_t gekko_VecTrace=3328ULL;
constexpr int64_t gekko_VecPerfMon=3840ULL;
constexpr uint32_t gekko_SRR1FPEnabled=1048576ULL;
constexpr uint32_t gekko_SRR1IllegalOp=524288ULL;
constexpr uint32_t gekko_SRR1Privileged=262144ULL;
constexpr uint32_t gekko_SRR1Trap=131072ULL;
constexpr uint32_t gekko_SRR1NotNextPC=65536ULL;
constexpr uint32_t gekko_FPSCRFX=2147483648ULL;
constexpr uint32_t gekko_FPSCRFEX=1073741824ULL;
constexpr uint32_t gekko_FPSCRVX=536870912ULL;
constexpr uint32_t gekko_FPSCROX=268435456ULL;
constexpr uint32_t gekko_FPSCRUX=134217728ULL;
constexpr uint32_t gekko_FPSCRZX=67108864ULL;
constexpr uint32_t gekko_FPSCRXX=33554432ULL;
constexpr uint32_t gekko_FPSCRVXSNAN=16777216ULL;
constexpr uint32_t gekko_FPSCRVXISI=8388608ULL;
constexpr uint32_t gekko_FPSCRVXIDI=4194304ULL;
constexpr uint32_t gekko_FPSCRVXZDZ=2097152ULL;
constexpr uint32_t gekko_FPSCRVXIMZ=1048576ULL;
constexpr uint32_t gekko_FPSCRVXVC=524288ULL;
constexpr uint32_t gekko_FPSCRFR=262144ULL;
constexpr uint32_t gekko_FPSCRFI=131072ULL;
constexpr uint32_t gekko_FPSCRVXSOFT=1024ULL;
constexpr uint32_t gekko_FPSCRVXSQRT=512ULL;
constexpr uint32_t gekko_FPSCRVXCVI=256ULL;
constexpr int64_t gekko_fprfShift=12ULL;
constexpr uint32_t gekko_fprfMask=126976ULL;
constexpr int64_t gekko_fprfQNaN=17ULL;
constexpr int64_t gekko_fprfNegInf=9ULL;
constexpr int64_t gekko_fprfNegNorm=8ULL;
constexpr int64_t gekko_fprfNegDenom=24ULL;
constexpr int64_t gekko_fprfNegZero=18ULL;
constexpr int64_t gekko_fprfPosZero=2ULL;
constexpr int64_t gekko_fprfPosDenom=20ULL;
constexpr int64_t gekko_fprfPosNorm=4ULL;
constexpr int64_t gekko_fprfPosInf=5ULL;
double gekko_qnan=go_math_Float64frombits(cast<uint64_t>(9221120237041090560ULL));
constexpr gekko_Flow gekko_FlowSeq=0ULL;
constexpr gekko_Flow gekko_FlowBranch=1ULL;
constexpr gekko_Flow gekko_FlowJump=2ULL;
constexpr gekko_Flow gekko_FlowCall=3ULL;
constexpr gekko_Flow gekko_FlowReturn=4ULL;
constexpr gekko_Flow gekko_FlowIndJump=5ULL;
constexpr gekko_Flow gekko_FlowIndCall=6ULL;
constexpr gekko_Flow gekko_FlowStop=7ULL;
constexpr int64_t gekko_SPRXER=1ULL;
constexpr int64_t gekko_SPRLR=8ULL;
constexpr int64_t gekko_SPRCTR=9ULL;
constexpr int64_t gekko_SPRDSISR=18ULL;
constexpr int64_t gekko_SPRDAR=19ULL;
constexpr int64_t gekko_SPRDEC=22ULL;
constexpr int64_t gekko_SPRSDR1=25ULL;
constexpr int64_t gekko_SPRSRR0=26ULL;
constexpr int64_t gekko_SPRSRR1=27ULL;
constexpr int64_t gekko_SPRSPRG0=272ULL;
constexpr int64_t gekko_SPRSPRG1=273ULL;
constexpr int64_t gekko_SPRSPRG2=274ULL;
constexpr int64_t gekko_SPRSPRG3=275ULL;
constexpr int64_t gekko_SPREAR=282ULL;
constexpr int64_t gekko_SPRTBL=284ULL;
constexpr int64_t gekko_SPRTBU=285ULL;
constexpr int64_t gekko_SPRPVR=287ULL;
constexpr int64_t gekko_SPRIBAT0U=528ULL;
constexpr int64_t gekko_SPRIBAT0L=529ULL;
constexpr int64_t gekko_SPRIBAT1U=530ULL;
constexpr int64_t gekko_SPRIBAT1L=531ULL;
constexpr int64_t gekko_SPRIBAT2U=532ULL;
constexpr int64_t gekko_SPRIBAT2L=533ULL;
constexpr int64_t gekko_SPRIBAT3U=534ULL;
constexpr int64_t gekko_SPRIBAT3L=535ULL;
constexpr int64_t gekko_SPRDBAT0U=536ULL;
constexpr int64_t gekko_SPRDBAT0L=537ULL;
constexpr int64_t gekko_SPRDBAT1U=538ULL;
constexpr int64_t gekko_SPRDBAT1L=539ULL;
constexpr int64_t gekko_SPRDBAT2U=540ULL;
constexpr int64_t gekko_SPRDBAT2L=541ULL;
constexpr int64_t gekko_SPRDBAT3U=542ULL;
constexpr int64_t gekko_SPRDBAT3L=543ULL;
constexpr int64_t gekko_SPRGQR0=912ULL;
constexpr int64_t gekko_SPRGQR7=919ULL;
constexpr int64_t gekko_SPRHID2=920ULL;
constexpr int64_t gekko_SPRWPAR=921ULL;
constexpr int64_t gekko_SPRDMAU=922ULL;
constexpr int64_t gekko_SPRDMAL=923ULL;
constexpr int64_t gekko_SPRHID0=1008ULL;
constexpr int64_t gekko_SPRHID1=1009ULL;
constexpr int64_t gekko_SPRIABR=1010ULL;
constexpr int64_t gekko_SPRHID4=1011ULL;
constexpr int64_t gekko_SPRDABR=1013ULL;
constexpr int64_t gekko_SPRL2CR=1017ULL;
constexpr int64_t gekko_SPRICTC=1019ULL;
constexpr int64_t gekko_SPRTHRM1=1020ULL;
constexpr int64_t gekko_SPRTHRM2=1021ULL;
constexpr int64_t gekko_SPRTHRM3=1022ULL;
Map<uint32_t,std::string> gekko_sprName=Map<uint32_t,std::string>{{cast<uint32_t>(1ULL),std::string("XER",3)},{cast<uint32_t>(8ULL),std::string("LR",2)},{cast<uint32_t>(9ULL),std::string("CTR",3)},{cast<uint32_t>(18ULL),std::string("DSISR",5)},{cast<uint32_t>(19ULL),std::string("DAR",3)},{cast<uint32_t>(22ULL),std::string("DEC",3)},{cast<uint32_t>(25ULL),std::string("SDR1",4)},{cast<uint32_t>(26ULL),std::string("SRR0",4)},{cast<uint32_t>(27ULL),std::string("SRR1",4)},{cast<uint32_t>(272ULL),std::string("SPRG0",5)},{cast<uint32_t>(273ULL),std::string("SPRG1",5)},{cast<uint32_t>(274ULL),std::string("SPRG2",5)},{cast<uint32_t>(275ULL),std::string("SPRG3",5)},{cast<uint32_t>(282ULL),std::string("EAR",3)},{cast<uint32_t>(284ULL),std::string("TBL",3)},{cast<uint32_t>(285ULL),std::string("TBU",3)},{cast<uint32_t>(287ULL),std::string("PVR",3)},{cast<uint32_t>(528ULL),std::string("IBAT0U",6)},{cast<uint32_t>(529ULL),std::string("IBAT0L",6)},{cast<uint32_t>(530ULL),std::string("IBAT1U",6)},{cast<uint32_t>(531ULL),std::string("IBAT1L",6)},{cast<uint32_t>(532ULL),std::string("IBAT2U",6)},{cast<uint32_t>(533ULL),std::string("IBAT2L",6)},{cast<uint32_t>(534ULL),std::string("IBAT3U",6)},{cast<uint32_t>(535ULL),std::string("IBAT3L",6)},{cast<uint32_t>(536ULL),std::string("DBAT0U",6)},{cast<uint32_t>(537ULL),std::string("DBAT0L",6)},{cast<uint32_t>(538ULL),std::string("DBAT1U",6)},{cast<uint32_t>(539ULL),std::string("DBAT1L",6)},{cast<uint32_t>(540ULL),std::string("DBAT2U",6)},{cast<uint32_t>(541ULL),std::string("DBAT2L",6)},{cast<uint32_t>(542ULL),std::string("DBAT3U",6)},{cast<uint32_t>(543ULL),std::string("DBAT3L",6)},{cast<uint32_t>(912ULL),std::string("GQR0",4)},{cast<uint32_t>(913ULL),std::string("GQR1",4)},{cast<uint32_t>(914ULL),std::string("GQR2",4)},{cast<uint32_t>(915ULL),std::string("GQR3",4)},{cast<uint32_t>(916ULL),std::string("GQR4",4)},{cast<uint32_t>(917ULL),std::string("GQR5",4)},{cast<uint32_t>(918ULL),std::string("GQR6",4)},{cast<uint32_t>(919ULL),std::string("GQR7",4)},{cast<uint32_t>(920ULL),std::string("HID2",4)},{cast<uint32_t>(921ULL),std::string("WPAR",4)},{cast<uint32_t>(922ULL),std::string("DMAU",4)},{cast<uint32_t>(923ULL),std::string("DMAL",4)},{cast<uint32_t>(1008ULL),std::string("HID0",4)},{cast<uint32_t>(1009ULL),std::string("HID1",4)},{cast<uint32_t>(1010ULL),std::string("IABR",4)},{cast<uint32_t>(1011ULL),std::string("HID4",4)},{cast<uint32_t>(1013ULL),std::string("DABR",4)},{cast<uint32_t>(1017ULL),std::string("L2CR",4)},{cast<uint32_t>(1019ULL),std::string("ICTC",4)},{cast<uint32_t>(1020ULL),std::string("THRM1",5)},{cast<uint32_t>(1021ULL),std::string("THRM2",5)},{cast<uint32_t>(1022ULL),std::string("THRM3",5)}};
constexpr uint32_t gekko_MSRLE=1ULL;
constexpr uint32_t gekko_MSRRI=2ULL;
constexpr uint32_t gekko_MSRDR=16ULL;
constexpr uint32_t gekko_MSRIR=32ULL;
constexpr uint32_t gekko_MSRIP=64ULL;
constexpr uint32_t gekko_MSRFE1=256ULL;
constexpr uint32_t gekko_MSRBE=512ULL;
constexpr uint32_t gekko_MSRSE=1024ULL;
constexpr uint32_t gekko_MSRFE0=2048ULL;
constexpr uint32_t gekko_MSRME=4096ULL;
constexpr uint32_t gekko_MSRFP=8192ULL;
constexpr uint32_t gekko_MSRPR=16384ULL;
constexpr uint32_t gekko_MSREE=32768ULL;
constexpr uint32_t gekko_MSRPOW=262144ULL;
constexpr uint32_t gekko_XERSO=2147483648ULL;
constexpr uint32_t gekko_XEROV=1073741824ULL;
constexpr uint32_t gekko_XERCA=536870912ULL;
constexpr uint32_t gekko_HID2WPE=1073741824ULL;
constexpr uint32_t gekko_HID2PSE=536870912ULL;
constexpr uint32_t gekko_HID2LCE=268435456ULL;
constexpr int64_t gekko_qFloat=0ULL;
constexpr int64_t gekko_qU8=4ULL;
constexpr int64_t gekko_qU16=5ULL;
constexpr int64_t gekko_qS8=6ULL;
constexpr int64_t gekko_qS16=7ULL;
constexpr int64_t gekko_CoreClock=486000000ULL;
constexpr int64_t gekko_BusClock=162000000ULL;
constexpr int64_t gekko_TimerClock=40500000ULL;
constexpr int64_t gekko_ClocksPerTick=12ULL;
constexpr int64_t gcdsp_regAR0=0ULL;
constexpr int64_t gcdsp_regIX0=4ULL;
constexpr int64_t gcdsp_regWR0=8ULL;
constexpr int64_t gcdsp_regST0=12ULL;
constexpr int64_t gcdsp_regST1=13ULL;
constexpr int64_t gcdsp_regST2=14ULL;
constexpr int64_t gcdsp_regST3=15ULL;
constexpr int64_t gcdsp_regAC0H=16ULL;
constexpr int64_t gcdsp_regAC1H=17ULL;
constexpr int64_t gcdsp_regCONFIG=18ULL;
constexpr int64_t gcdsp_regSR=19ULL;
constexpr int64_t gcdsp_regPRODL=20ULL;
constexpr int64_t gcdsp_regPRODM1=21ULL;
constexpr int64_t gcdsp_regPRODH=22ULL;
constexpr int64_t gcdsp_regPRODM2=23ULL;
constexpr int64_t gcdsp_regAX0L=24ULL;
constexpr int64_t gcdsp_regAX1L=25ULL;
constexpr int64_t gcdsp_regAX0H=26ULL;
constexpr int64_t gcdsp_regAX1H=27ULL;
constexpr int64_t gcdsp_regAC0L=28ULL;
constexpr int64_t gcdsp_regAC1L=29ULL;
constexpr int64_t gcdsp_regAC0M=30ULL;
constexpr int64_t gcdsp_regAC1M=31ULL;
constexpr int64_t gcdsp_srCarry=1ULL;
constexpr int64_t gcdsp_srOverflow=2ULL;
constexpr int64_t gcdsp_srZero=4ULL;
constexpr int64_t gcdsp_srSign=8ULL;
constexpr int64_t gcdsp_srAboveS32=16ULL;
constexpr int64_t gcdsp_srTopTwo=32ULL;
constexpr int64_t gcdsp_srLogicZero=64ULL;
constexpr int64_t gcdsp_srOverSticky=128ULL;
constexpr int64_t gcdsp_srMulNoDouble=8192ULL;
constexpr int64_t gcdsp_srMode40=16384ULL;
constexpr int64_t gcdsp_srMulUnsigned=32768ULL;
constexpr int64_t gc_stackWindow=2147483648ULL;
bool gc_diTrace=(go_os_Getenv(std::string("RR_GC_DITRACE",13)) != std::string("",0));
constexpr int64_t gc_diBreakInt=64ULL;
constexpr int64_t gc_diBreakMask=32ULL;
constexpr int64_t gc_diTCInt=16ULL;
constexpr int64_t gc_diTCMask=8ULL;
constexpr int64_t gc_diErrInt=4ULL;
constexpr int64_t gc_diErrMask=2ULL;
constexpr int64_t gc_diBreak=1ULL;
constexpr int64_t gc_diInstrPerSec=486000000ULL;
constexpr int64_t gc_diInstrPerByte=194ULL;
constexpr int64_t gc_diCmdInstr=486000ULL;
constexpr int64_t gc_diSeekInstr=14580000ULL;
bool gc_diInstant=(go_os_Getenv(std::string("RR_GC_DIINSTANT",15)) != std::string("",0));
constexpr int64_t gc_bootBinOff=0ULL;
constexpr int64_t gc_bootBinSize=1088ULL;
constexpr int64_t gc_bi2Off=1088ULL;
constexpr int64_t gc_bi2Size=8192ULL;
constexpr int64_t gc_apploaderOff=9280ULL;
constexpr int64_t gc_GameMagic=3258163005ULL;
constexpr int64_t gc_dolHeaderSize=256ULL;
constexpr int64_t gc_dolTextSegs=7ULL;
constexpr int64_t gc_dolDataSegs=11ULL;
constexpr int64_t gc_dolSegs=18ULL;
constexpr int64_t gc_dspCSRReset=1ULL;
constexpr int64_t gc_dspCSRPIInt=2ULL;
constexpr int64_t gc_dspCSRHalt=4ULL;
constexpr int64_t gc_dspCSRAIInt=8ULL;
constexpr int64_t gc_dspCSRAIMask=16ULL;
constexpr int64_t gc_dspCSRARInt=32ULL;
constexpr int64_t gc_dspCSRARMask=64ULL;
constexpr int64_t gc_dspCSRDSPInt=128ULL;
constexpr int64_t gc_dspCSRDSPMask=256ULL;
constexpr int64_t gc_dspCSRIntAck=170ULL;
constexpr int64_t gc_aidSamplesPerBlock=8ULL;
constexpr int64_t gc_aidInstrPerSample=10125ULL;
constexpr int64_t gc_aidInstrPerBlock=81000ULL;
bool gc_dspPCTrace=(go_os_Getenv(std::string("RR_GC_DSPPC",11)) != std::string("",0));
bool gc_dspTrace=(go_os_Getenv(std::string("RR_GC_DSPTRACE",14)) != std::string("",0));
bool gc_exiTrace=(go_os_Getenv(std::string("RR_GC_EXITRACE",14)) != std::string("",0));
constexpr int64_t gc_exiCSRTCIntMask=4ULL;
constexpr int64_t gc_exiCSRTCInt=8ULL;
constexpr int64_t gc_exiCSRStatus=2058ULL;
constexpr int64_t gc_exiDevMemCard=0ULL;
constexpr int64_t gc_exiDevRTC=1ULL;
constexpr int64_t gc_exiDevNone=-1ULL;
constexpr int64_t gc_fstEntrySize=12ULL;
constexpr int64_t gc_guardBand=2ULL;
gc_clipPlane gc_planeNear=gc_clipPlane{cast<float>(0ULL),cast<float>(0ULL),cast<float>(1ULL),cast<float>(1ULL)};
std::array<gc_clipPlane,5> gc_clipPlanes=std::array<gc_clipPlane,5>{gc_planeNear,gc_clipPlane{cast<float>(1ULL),cast<float>(0ULL),cast<float>(0ULL),gc_guardBand},gc_clipPlane{cast<float>(-cast<int64_t>(1ULL)),cast<float>(0ULL),cast<float>(0ULL),gc_guardBand},gc_clipPlane{cast<float>(0ULL),cast<float>(1ULL),cast<float>(0ULL),gc_guardBand},gc_clipPlane{cast<float>(0ULL),cast<float>(-cast<int64_t>(1ULL)),cast<float>(0ULL),gc_guardBand}};
constexpr int64_t gc_efbWidth=640ULL;
constexpr int64_t gc_efbHeight=528ULL;
bool gc_drawTrace=(go_os_Getenv(std::string("RR_GC_DRAWTRACE",15)) != std::string("",0));
std::string gc_texDump0=go_os_Getenv(std::string("RR_GC_TEXDUMP0",14));
auto [gc_pixDbgX,gc_pixDbgY]=[]()->std::tuple<int64_t,int64_t>{
std::string s = go_os_Getenv(std::string("RR_GC_PIXDBG",12));
if ((s == std::string("",0))) {
return {cast<int64_t>(-1ULL),cast<int64_t>(-1ULL)};
}
int64_t x={};
int64_t y={};
{
auto tmp110 = go_fmt_Sscanf(s,std::string("%d,%d",5),(&x),(&y));
Error err = std::get<1>(tmp110);
if (bool(err)) {
return {cast<int64_t>(-1ULL),cast<int64_t>(-1ULL)};
}
}
return {x,y};
}
();
constexpr int64_t gc_bandRows=4ULL;
constexpr int64_t gc_cullNone=0ULL;
constexpr int64_t gc_cullNegArea=1ULL;
constexpr int64_t gc_cullPosArea=2ULL;
constexpr int64_t gc_cullAll=3ULL;
std::string gc_cullExperiment=go_os_Getenv(std::string("RR_GC_CULLMODE",14));
constexpr int64_t gc_tevBiasCompare=3ULL;
constexpr int64_t gc_maxTexCoord=8ULL;
constexpr int64_t gc_texGenRegular=0ULL;
constexpr int64_t gc_texGenEmboss=1ULL;
constexpr int64_t gc_texGenColor0=2ULL;
constexpr int64_t gc_texGenColor1=3ULL;
constexpr int64_t gc_texSrcGeom=0ULL;
constexpr int64_t gc_texSrcNormal=1ULL;
constexpr int64_t gc_texSrcTex0=5ULL;
constexpr int64_t gc_texI4=0ULL;
constexpr int64_t gc_texI8=1ULL;
constexpr int64_t gc_texIA4=2ULL;
constexpr int64_t gc_texIA8=3ULL;
constexpr int64_t gc_texRGB565=4ULL;
constexpr int64_t gc_texRGB5A3=5ULL;
constexpr int64_t gc_texRGBA8=6ULL;
constexpr int64_t gc_texC4=8ULL;
constexpr int64_t gc_texC8=9ULL;
constexpr int64_t gc_texC14X2=10ULL;
constexpr int64_t gc_texCMPR=14ULL;
constexpr int64_t gc_descNone=0ULL;
constexpr int64_t gc_descDirect=1ULL;
constexpr int64_t gc_descIndex8=2ULL;
constexpr int64_t gc_descIndex16=3ULL;
constexpr int64_t gc_idleArmEvery=4096ULL;
constexpr int64_t gc_idleWatch=64ULL;
constexpr int64_t gc_apploaderAddr=2166358016ULL;
constexpr int64_t gc_reportTramp=2167603200ULL;
constexpr int64_t gc_returnSentinel=2167668736ULL;
constexpr int64_t gc_bootMagic=219540062ULL;
constexpr int64_t gc_BusClock=162000000ULL;
constexpr int64_t gc_CoreClock=486000000ULL;
bool gc_iplTrace=(go_os_Getenv(std::string("RR_GC_IPLTRACE",14)) != std::string("",0));
constexpr int64_t gc_RAMSize=25165824ULL;
constexpr int64_t gc_ARAMSize=16777216ULL;
constexpr int64_t gc_hwBase=201326592ULL;
constexpr int64_t gc_hwEnd=201392128ULL;
constexpr int64_t gc_regCP=201326592ULL;
constexpr int64_t gc_regPE=201330688ULL;
constexpr int64_t gc_regVI=201334784ULL;
constexpr int64_t gc_regPI=201338880ULL;
constexpr int64_t gc_regMI=201342976ULL;
constexpr int64_t gc_regDSP=201347072ULL;
constexpr int64_t gc_regDI=201351168ULL;
constexpr int64_t gc_regSI=201352192ULL;
constexpr int64_t gc_regEXI=201353216ULL;
constexpr int64_t gc_regAIS=201354240ULL;
constexpr int64_t gc_regWGPipe=201359360ULL;
constexpr int64_t gc_IntError=0ULL;
constexpr int64_t gc_IntReset=1ULL;
constexpr int64_t gc_IntDI=2ULL;
constexpr int64_t gc_IntSI=3ULL;
constexpr int64_t gc_IntEXI=4ULL;
constexpr int64_t gc_IntAI=5ULL;
constexpr int64_t gc_IntDSP=6ULL;
constexpr int64_t gc_IntMEM=7ULL;
constexpr int64_t gc_IntVI=8ULL;
constexpr int64_t gc_IntPEToken=9ULL;
constexpr int64_t gc_IntPEFinish=10ULL;
constexpr int64_t gc_IntCP=11ULL;
constexpr int64_t gc_IntDebug=12ULL;
constexpr int64_t gc_IntHSP=13ULL;
constexpr int64_t gc_bucketFIFO=0ULL;
constexpr int64_t gc_bucketVertex=1ULL;
constexpr int64_t gc_bucketRaster=2ULL;
constexpr int64_t gc_bucketCopy=3ULL;
constexpr int64_t gc_bucketDSP=4ULL;
constexpr int64_t gc_numBuckets=5ULL;
constexpr int64_t gc_spinWindow=16200000ULL;
bool gc_siTrace=(go_os_Getenv(std::string("RR_GC_SITRACE",13)) != std::string("",0));
constexpr int64_t gc_siTStart=1ULL;
constexpr int64_t gc_siTCInt=2147483648ULL;
constexpr int64_t gc_siTCIntMsk=1073741824ULL;
constexpr int64_t gc_siRDST=32ULL;
constexpr int64_t gc_siNoResponse=8ULL;
constexpr int64_t gc_PadStickCentre=128ULL;
constexpr int64_t gc_PadStickFull=96ULL;
Map<std::string,uint16_t> gc_padButtons=Map<std::string,uint16_t>{{std::string("a",1),cast<uint16_t>(256ULL)},{std::string("b",1),cast<uint16_t>(512ULL)},{std::string("x",1),cast<uint16_t>(1024ULL)},{std::string("y",1),cast<uint16_t>(2048ULL)},{std::string("start",5),cast<uint16_t>(4096ULL)},{std::string("z",1),cast<uint16_t>(16ULL)},{std::string("r",1),cast<uint16_t>(32ULL)},{std::string("l",1),cast<uint16_t>(64ULL)},{std::string("up",2),cast<uint16_t>(8ULL)},{std::string("down",4),cast<uint16_t>(4ULL)},{std::string("right",5),cast<uint16_t>(2ULL)},{std::string("left",4),cast<uint16_t>(1ULL)}};
constexpr int64_t gc_stackProfDefaultEvery=199ULL;
constexpr int64_t gc_fieldInstructions=8100000ULL;

#include "adapters.h"
// tools/cpu/gekko/cache.go:43:1
void gekko_LockedCache_Reset(gekko_LockedCache* l){
{
l->Data = std::array<uint8_t,16384>{};
l->Base = cast<uint32_t>(3758096384ULL);
l->Enabled_ = false;
}
}
// tools/cpu/gekko/cache.go:49:1
bool gekko_LockedCache_Enabled(gekko_LockedCache* l){
{
return l->Enabled_;
}
}
// tools/cpu/gekko/cache.go:54:1
bool gekko_LockedCache_Contains(gekko_LockedCache* l,uint32_t a){
{
return ((l->Enabled_ && (a >= l->Base)) && (a < cast<uint32_t>((l->Base + cast<uint32_t>(16384ULL)))));
}
}
// tools/cpu/gekko/cache.go:58:1
uint32_t gekko_LockedCache_off(gekko_LockedCache* l,uint32_t a){
{
return cast<uint32_t>(((cast<uint32_t>((a - l->Base))) & cast<uint32_t>(16383ULL)));
}
}
// tools/cpu/gekko/cache.go:60:1
uint8_t gekko_LockedCache_Read8(gekko_LockedCache* l,uint32_t a){
{
return l->Data[gekko_LockedCache_off(l,a)];
}
}
// tools/cpu/gekko/cache.go:62:1
uint16_t gekko_LockedCache_Read16(gekko_LockedCache* l,uint32_t a){
{
uint32_t o = gekko_LockedCache_off(l,a);
return cast<uint16_t>((shl<uint16_t>(cast<uint16_t>(l->Data[o]),cast<int64_t>(8ULL)) | cast<uint16_t>(l->Data[cast<uint32_t>((o + cast<uint32_t>(1ULL)))])));
}
}
// tools/cpu/gekko/cache.go:67:1
uint32_t gekko_LockedCache_Read32(gekko_LockedCache* l,uint32_t a){
{
uint32_t o = gekko_LockedCache_off(l,a);
return cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((shl<uint32_t>(cast<uint32_t>(l->Data[o]),cast<int64_t>(24ULL)) | shl<uint32_t>(cast<uint32_t>(l->Data[cast<uint32_t>((o + cast<uint32_t>(1ULL)))]),cast<int64_t>(16ULL)))) | shl<uint32_t>(cast<uint32_t>(l->Data[cast<uint32_t>((o + cast<uint32_t>(2ULL)))]),cast<int64_t>(8ULL)))) | cast<uint32_t>(l->Data[cast<uint32_t>((o + cast<uint32_t>(3ULL)))])));
}
}
// tools/cpu/gekko/cache.go:72:1
void gekko_LockedCache_Write8(gekko_LockedCache* l,uint32_t a,uint8_t v){
{
l->Data[gekko_LockedCache_off(l,a)] = v;
}
}
// tools/cpu/gekko/cache.go:74:1
void gekko_LockedCache_Write16(gekko_LockedCache* l,uint32_t a,uint16_t v){
{
uint32_t o = gekko_LockedCache_off(l,a);
auto tmp1 = std::make_tuple(cast<uint8_t>(shr<uint16_t>(v,cast<int64_t>(8ULL))),cast<uint8_t>(v));
l->Data[o] = std::get<0>(tmp1);
l->Data[cast<uint32_t>((o + cast<uint32_t>(1ULL)))] = std::get<1>(tmp1);
}
}
// tools/cpu/gekko/cache.go:79:1
void gekko_LockedCache_Write32(gekko_LockedCache* l,uint32_t a,uint32_t v){
{
uint32_t o = gekko_LockedCache_off(l,a);
auto tmp2 = std::make_tuple(cast<uint8_t>(shr<uint32_t>(v,cast<int64_t>(24ULL))),cast<uint8_t>(shr<uint32_t>(v,cast<int64_t>(16ULL))));
l->Data[o] = std::get<0>(tmp2);
l->Data[cast<uint32_t>((o + cast<uint32_t>(1ULL)))] = std::get<1>(tmp2);
auto tmp3 = std::make_tuple(cast<uint8_t>(shr<uint32_t>(v,cast<int64_t>(8ULL))),cast<uint8_t>(v));
l->Data[cast<uint32_t>((o + cast<uint32_t>(2ULL)))] = std::get<0>(tmp3);
l->Data[cast<uint32_t>((o + cast<uint32_t>(3ULL)))] = std::get<1>(tmp3);
}
}
// tools/cpu/gekko/cache.go:86:1
void gekko_CPU_setHID2(gekko_CPU* c,uint32_t v){
{
c->HID2 = v;
c->LC.Enabled_ = (cast<uint32_t>((v & cast<uint32_t>(268435456ULL))) != cast<uint32_t>(0ULL));
}
}
// tools/cpu/gekko/cache.go:93:1
void gekko_CPU_dcbz(gekko_CPU* c,uint32_t ea){
{
uint32_t base = (ea & ~(cast<uint32_t>(31ULL)));
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < cast<uint32_t>(32ULL));i += cast<uint32_t>(4ULL)){
gekko_CPU_write32(c,cast<uint32_t>((base + i)),cast<uint32_t>(0ULL));
}
}}
}
// tools/cpu/gekko/cache.go:103:1
void gekko_CPU_dcbzL(gekko_CPU* c,uint32_t ea){
{
if ((!gekko_LockedCache_Enabled(&(c->LC)))) {
gekko_CPU_Halt(c,std::string("gekko: dcbz_l at PC 0x%08X but the locked cache is not enabled (HID2 = 0x%08X)",78),c->PC,c->HID2);
return ;
}
if ((!gekko_LockedCache_Contains(&(c->LC),ea))) {
gekko_CPU_Halt(c,std::string("gekko: dcbz_l on 0x%08X, which is outside the locked cache at 0x%08X",68),ea,c->LC.Base);
return ;
}
uint32_t base = (ea & ~(cast<uint32_t>(31ULL)));
uint32_t o = gekko_LockedCache_off(&(c->LC),base);
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < cast<uint32_t>(32ULL));i++){
c->LC.Data[cast<uint32_t>((o + i))] = cast<uint8_t>(0ULL);
}
}}
}
// tools/cpu/gekko/cache.go:125:1
void gekko_CPU_runDMA(gekko_CPU* c){
{
if ((cast<uint32_t>((c->DMAL & cast<uint32_t>(2ULL))) == cast<uint32_t>(0ULL))) {
return ;
}
uint32_t memAddr = cast<uint32_t>((c->DMAU & cast<uint32_t>(4294967264ULL)));
uint32_t lcAddr = cast<uint32_t>((c->DMAL & cast<uint32_t>(4294967264ULL)));
uint32_t length = cast<uint32_t>(((shl<uint32_t>((cast<uint32_t>((c->DMAU & cast<uint32_t>(31ULL)))),cast<int64_t>(2ULL))) | (cast<uint32_t>(((shr<uint32_t>(c->DMAL,cast<int64_t>(2ULL))) & cast<uint32_t>(3ULL))))));
if ((length == cast<uint32_t>(0ULL))) {
length = cast<uint32_t>(128ULL);
}
bool load = (cast<uint32_t>((c->DMAL & cast<uint32_t>(16ULL))) != cast<uint32_t>(0ULL));
if ((!gekko_LockedCache_Enabled(&(c->LC)))) {
gekko_CPU_Halt(c,std::string("gekko: a locked-cache DMA was triggered but the cache is not enabled (HID2 = 0x%08X)",84),c->HID2);
return ;
}
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < cast<uint32_t>((length * cast<uint32_t>(32ULL))));i++){
if (load) {
c->LC.Data[gekko_LockedCache_off(&(c->LC),cast<uint32_t>((lcAddr + i)))] = gc_Machine_Read8(c->bus,gekko_CPU_memPhys(c,cast<uint32_t>((memAddr + i))));
}
else {
gc_Machine_Write8(c->bus,gekko_CPU_memPhys(c,cast<uint32_t>((memAddr + i))),c->LC.Data[gekko_LockedCache_off(&(c->LC),cast<uint32_t>((lcAddr + i)))]);
}
}
}c->DMAL &= ~(cast<uint32_t>(2ULL));
}
}
// tools/cpu/gekko/cache.go:154:1
uint32_t gekko_CPU_memPhys(gekko_CPU* c,uint32_t ea){
{
auto tmp4 = gekko_CPU_Translate(c,ea,false,false);
uint32_t pa = std::get<0>(tmp4);
bool ok = std::get<1>(tmp4);
if ((!ok)) {
return cast<uint32_t>(0ULL);
}
return pa;
}
}
// tools/cpu/gekko/cpu.go:133:1
void gekko_CPU_Reset(gekko_CPU* c){
{
(*c) = gekko_CPU{{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},c->bus,c->fetcher,c->SC};
c->MSR = cast<uint32_t>(4160ULL);
c->PC = cast<uint32_t>(4293918976ULL);
c->PVR = cast<uint32_t>(537108ULL);
c->FPSCR = cast<uint32_t>(0ULL);
gekko_LockedCache_Reset(&(c->LC));
}
}
// tools/cpu/gekko/cpu.go:156:1
uint64_t gekko_CPU_CurPC(gekko_CPU* c){
{
return cast<uint64_t>(c->PC);
}
}
// tools/cpu/gekko/cpu.go:159:1
void gekko_CPU_SetPC(gekko_CPU* c,uint32_t pc){
{
c->PC = pc;
}
}
// tools/cpu/gekko/cpu.go:162:1
uint32_t gekko_CPU_Reg(gekko_CPU* c,uint32_t i){
{
return c->GPR[cast<uint32_t>((i & cast<uint32_t>(31ULL)))];
}
}
// tools/cpu/gekko/cpu.go:165:1
void gekko_CPU_SetReg(gekko_CPU* c,uint32_t i,uint32_t v){
{
c->GPR[cast<uint32_t>((i & cast<uint32_t>(31ULL)))] = v;
}
}
// tools/cpu/gekko/cpu.go:171:1
uint32_t gekko_CPU_CRField(gekko_CPU* c,uint32_t n){
{
return cast<uint32_t>(((shr<uint32_t>(c->CR,(cast<uint32_t>((cast<uint32_t>(28ULL) - cast<uint32_t>((cast<uint32_t>(4ULL) * (cast<uint32_t>((n & cast<uint32_t>(7ULL))))))))))) & cast<uint32_t>(15ULL)));
}
}
// tools/cpu/gekko/cpu.go:176:1
void gekko_CPU_SetCRField(gekko_CPU* c,uint32_t n,uint32_t v){
{
uint32_t sh = cast<uint32_t>((cast<uint32_t>(28ULL) - cast<uint32_t>((cast<uint32_t>(4ULL) * (cast<uint32_t>((n & cast<uint32_t>(7ULL))))))));
c->CR = cast<uint32_t>((((c->CR & ~((shl<uint32_t>(cast<uint32_t>(15ULL),sh))))) | (shl<uint32_t>((cast<uint32_t>((v & cast<uint32_t>(15ULL)))),sh))));
}
}
// tools/cpu/gekko/cpu.go:192:1
void gekko_CPU_setCR0(gekko_CPU* c,uint32_t v){
{
uint32_t f = cast<uint32_t>(0ULL);
{
if ((cast<int32_t>(v) < cast<int32_t>(0ULL))){
f = cast<uint32_t>(8ULL);
}
else if ((cast<int32_t>(v) > cast<int32_t>(0ULL))){
f = cast<uint32_t>(4ULL);
}
else {
f = cast<uint32_t>(2ULL);
}
}
tmp5:;
if ((cast<uint32_t>((c->XER & cast<uint32_t>(2147483648ULL))) != cast<uint32_t>(0ULL))) {
f |= cast<uint32_t>(1ULL);
}
gekko_CPU_SetCRField(c,cast<uint32_t>(0ULL),f);
}
}
// tools/cpu/gekko/cpu.go:210:1
void gekko_CPU_setCR1(gekko_CPU* c){
{
gekko_CPU_SetCRField(c,cast<uint32_t>(1ULL),shr<uint32_t>(c->FPSCR,cast<int64_t>(28ULL)));
}
}
// tools/cpu/gekko/cpu.go:216:1
void gekko_CPU_setCA(gekko_CPU* c,bool on){
{
if (on) {
c->XER |= cast<uint32_t>(536870912ULL);
}
else {
c->XER &= ~(cast<uint32_t>(536870912ULL));
}
}
}
// tools/cpu/gekko/cpu.go:224:1
uint32_t gekko_CPU_ca(gekko_CPU* c){
{
if ((cast<uint32_t>((c->XER & cast<uint32_t>(536870912ULL))) != cast<uint32_t>(0ULL))) {
return cast<uint32_t>(1ULL);
}
return cast<uint32_t>(0ULL);
}
}
// tools/cpu/gekko/cpu.go:233:1
void gekko_CPU_setOV(gekko_CPU* c,bool on){
{
if (on) {
c->XER |= cast<uint32_t>(3221225472ULL);
}
else {
c->XER &= ~(cast<uint32_t>(1073741824ULL));
}
}
}
// tools/cpu/gekko/cpu.go:248:1
uint8_t gekko_CPU_read8(gekko_CPU* c,uint32_t ea){
{
auto tmp6 = gekko_CPU_Translate(c,ea,false,false);
uint32_t pa = std::get<0>(tmp6);
bool ok = std::get<1>(tmp6);
if ((!ok)) {
return cast<uint8_t>(0ULL);
}
if (gekko_LockedCache_Contains(&(c->LC),pa)) {
return gekko_LockedCache_Read8(&(c->LC),pa);
}
return gc_Machine_Read8(c->bus,pa);
}
}
// tools/cpu/gekko/cpu.go:259:1
uint16_t gekko_CPU_read16(gekko_CPU* c,uint32_t ea){
{
auto tmp7 = gekko_CPU_Translate(c,ea,false,false);
uint32_t pa = std::get<0>(tmp7);
bool ok = std::get<1>(tmp7);
if ((!ok)) {
return cast<uint16_t>(0ULL);
}
if (gekko_LockedCache_Contains(&(c->LC),pa)) {
return gekko_LockedCache_Read16(&(c->LC),pa);
}
return gc_Machine_Read16(c->bus,pa);
}
}
// tools/cpu/gekko/cpu.go:270:1
uint32_t gekko_CPU_read32(gekko_CPU* c,uint32_t ea){
{
auto tmp8 = gekko_CPU_Translate(c,ea,false,false);
uint32_t pa = std::get<0>(tmp8);
bool ok = std::get<1>(tmp8);
if ((!ok)) {
return cast<uint32_t>(0ULL);
}
if (gekko_LockedCache_Contains(&(c->LC),pa)) {
return gekko_LockedCache_Read32(&(c->LC),pa);
}
return gc_Machine_Read32(c->bus,pa);
}
}
// tools/cpu/gekko/cpu.go:281:1
uint64_t gekko_CPU_read64(gekko_CPU* c,uint32_t ea){
{
return cast<uint64_t>((shl<uint64_t>(cast<uint64_t>(gekko_CPU_read32(c,ea)),cast<int64_t>(32ULL)) | cast<uint64_t>(gekko_CPU_read32(c,cast<uint32_t>((ea + cast<uint32_t>(4ULL)))))));
}
}
// tools/cpu/gekko/cpu.go:285:1
void gekko_CPU_write8(gekko_CPU* c,uint32_t ea,uint8_t v){
{
auto tmp9 = gekko_CPU_Translate(c,ea,true,false);
uint32_t pa = std::get<0>(tmp9);
bool ok = std::get<1>(tmp9);
if ((!ok)) {
return ;
}
if (gekko_LockedCache_Contains(&(c->LC),pa)) {
gekko_LockedCache_Write8(&(c->LC),pa,v);
return ;
}
gekko_CPU_clearReservation(c,pa);
gc_Machine_Write8(c->bus,pa,v);
}
}
// tools/cpu/gekko/cpu.go:298:1
void gekko_CPU_write16(gekko_CPU* c,uint32_t ea,uint16_t v){
{
auto tmp10 = gekko_CPU_Translate(c,ea,true,false);
uint32_t pa = std::get<0>(tmp10);
bool ok = std::get<1>(tmp10);
if ((!ok)) {
return ;
}
if (gekko_LockedCache_Contains(&(c->LC),pa)) {
gekko_LockedCache_Write16(&(c->LC),pa,v);
return ;
}
gekko_CPU_clearReservation(c,pa);
gc_Machine_Write16(c->bus,pa,v);
}
}
// tools/cpu/gekko/cpu.go:311:1
void gekko_CPU_write32(gekko_CPU* c,uint32_t ea,uint32_t v){
{
auto tmp11 = gekko_CPU_Translate(c,ea,true,false);
uint32_t pa = std::get<0>(tmp11);
bool ok = std::get<1>(tmp11);
if ((!ok)) {
return ;
}
if (gekko_LockedCache_Contains(&(c->LC),pa)) {
gekko_LockedCache_Write32(&(c->LC),pa,v);
return ;
}
gekko_CPU_clearReservation(c,pa);
gc_Machine_Write32(c->bus,pa,v);
}
}
// tools/cpu/gekko/cpu.go:324:1
void gekko_CPU_write64(gekko_CPU* c,uint32_t ea,uint64_t v){
{
gekko_CPU_write32(c,ea,cast<uint32_t>(shr<uint64_t>(v,cast<int64_t>(32ULL))));
gekko_CPU_write32(c,cast<uint32_t>((ea + cast<uint32_t>(4ULL))),cast<uint32_t>(v));
}
}
// tools/cpu/gekko/cpu.go:331:1
void gekko_CPU_clearReservation(gekko_CPU* c,uint32_t pa){
{
if ((c->Reserved && ((pa & ~(cast<uint32_t>(31ULL))) == (c->ReserveAddr & ~(cast<uint32_t>(31ULL)))))) {
c->Reserved = false;
}
}
}
// tools/cpu/gekko/cpu.go:339:1
uint8_t gekko_CPU_ReadMem(gekko_CPU* c,uint32_t ea){
{
return gekko_CPU_read8(c,ea);
}
}
// tools/cpu/gekko/cpu.go:359:1
double gekko_f32(double v){
{
return cast<double>(cast<float>(v));
}
}
// tools/cpu/gekko/cpu.go:364:1
bool gekko_CPU_psEnabled(gekko_CPU* c){
{
return (cast<uint32_t>((c->HID2 & cast<uint32_t>(536870912ULL))) != cast<uint32_t>(0ULL));
}
}
// tools/cpu/gekko/cpu.go:369:1
uint64_t gekko_f64bits(double f){
{
return go_math_Float64bits(f);
}
}
// tools/cpu/gekko/cpu.go:370:1
double gekko_f64from(uint64_t b){
{
return go_math_Float64frombits(b);
}
}
// tools/cpu/gekko/cpu.go:371:1
uint64_t gekko_bits64(double f){
{
return go_math_Float64bits(f);
}
}
// tools/cpu/gekko/cpu.go:374:1
uint32_t gekko_float32bitsOf(double v){
{
return go_math_Float32bits(cast<float>(v));
}
}
// tools/cpu/gekko/decode.go:32:1
gekko_Inst gekko_Decode(Slice<uint8_t> code,uint32_t addr){
{
if ((len(code) < cast<int64_t>(4ULL))) {
return gekko_Inst{addr,{},cast<int64_t>(0ULL),std::string("?",1),std::string("(truncated)",11),cast<gekko_Flow>(7ULL),{},{}};
}
return gekko_DecodeWord(be_Uint32(code),addr);
}
}
// tools/cpu/gekko/decode.go:40:1
gekko_Inst gekko_DecodeWord(uint32_t w,uint32_t addr){
{
gekko_Inst in = gekko_Inst{addr,w,cast<int64_t>(4ULL),{},{},cast<gekko_Flow>(0ULL),{},{}};
gekko_decode((&in),w,addr);
if ((in.Mnem == std::string("",0))) {
in.Mnem = std::string(".word",5);
in.Text = go_fmt_Sprintf(std::string(".word 0x%08X",12),w);
in.Flow = cast<gekko_Flow>(7ULL);
}
return in;
}
}
// tools/cpu/gekko/decode.go:61:1
std::string gekko_r(uint32_t n){
{
return go_fmt_Sprintf(std::string("r%d",3),n);
}
}
// tools/cpu/gekko/decode.go:62:1
std::string gekko_fr(uint32_t n){
{
return go_fmt_Sprintf(std::string("f%d",3),n);
}
}
// tools/cpu/gekko/decode.go:66:1
std::string gekko_dot(std::string mnem,uint32_t w){
{
if (gekko_rcbit(w)) {
return (mnem + std::string(".",1));
}
return mnem;
}
}
// tools/cpu/gekko/decode.go:74:1
std::string gekko_oeDot(std::string mnem,uint32_t w){
{
if (gekko_oe(w)) {
mnem += std::string("o",1);
}
return gekko_dot(mnem,w);
}
}
// tools/cpu/gekko/decode.go:81:1
void gekko_decode(gekko_Inst* in,uint32_t w,uint32_t addr){
{
{
switch(gekko_opcd(w)){
case cast<uint32_t>(3ULL):{
gekko_Inst_set(in,std::string("twi",3),std::string("%d,%s,%d",8),gekko_rs(w),gekko_r(gekko_ra(w)),gekko_simm(w));
in->Flow = cast<gekko_Flow>(7ULL);
break;}
case cast<uint32_t>(4ULL):{
gekko_decodePS(in,w);
break;}
case cast<uint32_t>(7ULL):{
gekko_Inst_set(in,std::string("mulli",5),std::string("%s,%s,%d",8),gekko_r(gekko_rs(w)),gekko_r(gekko_ra(w)),gekko_simm(w));
break;}
case cast<uint32_t>(8ULL):{
gekko_Inst_set(in,std::string("subfic",6),std::string("%s,%s,%d",8),gekko_r(gekko_rs(w)),gekko_r(gekko_ra(w)),gekko_simm(w));
break;}
case cast<uint32_t>(10ULL):{
gekko_Inst_set(in,std::string("cmplwi",6),std::string("cr%d,%s,%d",10),gekko_crfD(w),gekko_r(gekko_ra(w)),gekko_uimm(w));
break;}
case cast<uint32_t>(11ULL):{
gekko_Inst_set(in,std::string("cmpwi",5),std::string("cr%d,%s,%d",10),gekko_crfD(w),gekko_r(gekko_ra(w)),gekko_simm(w));
break;}
case cast<uint32_t>(12ULL):{
gekko_Inst_set(in,std::string("addic",5),std::string("%s,%s,%d",8),gekko_r(gekko_rs(w)),gekko_r(gekko_ra(w)),gekko_simm(w));
break;}
case cast<uint32_t>(13ULL):{
gekko_Inst_set(in,std::string("addic.",6),std::string("%s,%s,%d",8),gekko_r(gekko_rs(w)),gekko_r(gekko_ra(w)),gekko_simm(w));
break;}
case cast<uint32_t>(14ULL):{
if ((gekko_ra(w) == cast<uint32_t>(0ULL))) {
gekko_Inst_set(in,std::string("li",2),std::string("%s,%d",5),gekko_r(gekko_rs(w)),gekko_simm(w));
}
else {
gekko_Inst_set(in,std::string("addi",4),std::string("%s,%s,%d",8),gekko_r(gekko_rs(w)),gekko_r(gekko_ra(w)),gekko_simm(w));
}
break;}
case cast<uint32_t>(15ULL):{
if ((gekko_ra(w) == cast<uint32_t>(0ULL))) {
gekko_Inst_set(in,std::string("lis",3),std::string("%s,0x%X",7),gekko_r(gekko_rs(w)),gekko_uimm(w));
}
else {
gekko_Inst_set(in,std::string("addis",5),std::string("%s,%s,%d",8),gekko_r(gekko_rs(w)),gekko_r(gekko_ra(w)),gekko_simm(w));
}
break;}
case cast<uint32_t>(16ULL):{
gekko_decodeBC(in,w,addr);
break;}
case cast<uint32_t>(17ULL):{
gekko_Inst_set(in,std::string("sc",2),std::string("",0));
in->Flow = cast<gekko_Flow>(7ULL);
break;}
case cast<uint32_t>(18ULL):{
gekko_decodeB(in,w,addr);
break;}
case cast<uint32_t>(19ULL):{
gekko_decode19(in,w);
break;}
case cast<uint32_t>(20ULL):{
gekko_Inst_set(in,gekko_dot(std::string("rlwimi",6),w),std::string("%s,%s,%d,%d,%d",14),gekko_r(gekko_ra(w)),gekko_r(gekko_rs(w)),gekko_shOf(w),gekko_mbOf(w),gekko_meOf(w));
break;}
case cast<uint32_t>(21ULL):{
gekko_decodeRlwinm(in,w);
break;}
case cast<uint32_t>(23ULL):{
gekko_Inst_set(in,gekko_dot(std::string("rlwnm",5),w),std::string("%s,%s,%s,%d,%d",14),gekko_r(gekko_ra(w)),gekko_r(gekko_rs(w)),gekko_r(gekko_rb(w)),gekko_mbOf(w),gekko_meOf(w));
break;}
case cast<uint32_t>(24ULL):{
if ((((gekko_rs(w) == cast<uint32_t>(0ULL)) && (gekko_ra(w) == cast<uint32_t>(0ULL))) && (gekko_uimm(w) == cast<uint32_t>(0ULL)))) {
gekko_Inst_set(in,std::string("nop",3),std::string("",0));
}
else {
gekko_Inst_set(in,std::string("ori",3),std::string("%s,%s,0x%X",10),gekko_r(gekko_ra(w)),gekko_r(gekko_rs(w)),gekko_uimm(w));
}
break;}
case cast<uint32_t>(25ULL):{
gekko_Inst_set(in,std::string("oris",4),std::string("%s,%s,0x%X",10),gekko_r(gekko_ra(w)),gekko_r(gekko_rs(w)),gekko_uimm(w));
break;}
case cast<uint32_t>(26ULL):{
gekko_Inst_set(in,std::string("xori",4),std::string("%s,%s,0x%X",10),gekko_r(gekko_ra(w)),gekko_r(gekko_rs(w)),gekko_uimm(w));
break;}
case cast<uint32_t>(27ULL):{
gekko_Inst_set(in,std::string("xoris",5),std::string("%s,%s,0x%X",10),gekko_r(gekko_ra(w)),gekko_r(gekko_rs(w)),gekko_uimm(w));
break;}
case cast<uint32_t>(28ULL):{
gekko_Inst_set(in,std::string("andi.",5),std::string("%s,%s,0x%X",10),gekko_r(gekko_ra(w)),gekko_r(gekko_rs(w)),gekko_uimm(w));
break;}
case cast<uint32_t>(29ULL):{
gekko_Inst_set(in,std::string("andis.",6),std::string("%s,%s,0x%X",10),gekko_r(gekko_ra(w)),gekko_r(gekko_rs(w)),gekko_uimm(w));
break;}
case cast<uint32_t>(31ULL):{
gekko_decode31(in,w);
break;}
case cast<uint32_t>(32ULL):{
gekko_mem(in,std::string("lwz",3),gekko_r(gekko_rs(w)),w);
break;}
case cast<uint32_t>(33ULL):{
gekko_mem(in,std::string("lwzu",4),gekko_r(gekko_rs(w)),w);
break;}
case cast<uint32_t>(34ULL):{
gekko_mem(in,std::string("lbz",3),gekko_r(gekko_rs(w)),w);
break;}
case cast<uint32_t>(35ULL):{
gekko_mem(in,std::string("lbzu",4),gekko_r(gekko_rs(w)),w);
break;}
case cast<uint32_t>(36ULL):{
gekko_mem(in,std::string("stw",3),gekko_r(gekko_rs(w)),w);
break;}
case cast<uint32_t>(37ULL):{
gekko_mem(in,std::string("stwu",4),gekko_r(gekko_rs(w)),w);
break;}
case cast<uint32_t>(38ULL):{
gekko_mem(in,std::string("stb",3),gekko_r(gekko_rs(w)),w);
break;}
case cast<uint32_t>(39ULL):{
gekko_mem(in,std::string("stbu",4),gekko_r(gekko_rs(w)),w);
break;}
case cast<uint32_t>(40ULL):{
gekko_mem(in,std::string("lhz",3),gekko_r(gekko_rs(w)),w);
break;}
case cast<uint32_t>(41ULL):{
gekko_mem(in,std::string("lhzu",4),gekko_r(gekko_rs(w)),w);
break;}
case cast<uint32_t>(42ULL):{
gekko_mem(in,std::string("lha",3),gekko_r(gekko_rs(w)),w);
break;}
case cast<uint32_t>(43ULL):{
gekko_mem(in,std::string("lhau",4),gekko_r(gekko_rs(w)),w);
break;}
case cast<uint32_t>(44ULL):{
gekko_mem(in,std::string("sth",3),gekko_r(gekko_rs(w)),w);
break;}
case cast<uint32_t>(45ULL):{
gekko_mem(in,std::string("sthu",4),gekko_r(gekko_rs(w)),w);
break;}
case cast<uint32_t>(46ULL):{
gekko_mem(in,std::string("lmw",3),gekko_r(gekko_rs(w)),w);
break;}
case cast<uint32_t>(47ULL):{
gekko_mem(in,std::string("stmw",4),gekko_r(gekko_rs(w)),w);
break;}
case cast<uint32_t>(48ULL):{
gekko_mem(in,std::string("lfs",3),gekko_fr(gekko_rs(w)),w);
break;}
case cast<uint32_t>(49ULL):{
gekko_mem(in,std::string("lfsu",4),gekko_fr(gekko_rs(w)),w);
break;}
case cast<uint32_t>(50ULL):{
gekko_mem(in,std::string("lfd",3),gekko_fr(gekko_rs(w)),w);
break;}
case cast<uint32_t>(51ULL):{
gekko_mem(in,std::string("lfdu",4),gekko_fr(gekko_rs(w)),w);
break;}
case cast<uint32_t>(52ULL):{
gekko_mem(in,std::string("stfs",4),gekko_fr(gekko_rs(w)),w);
break;}
case cast<uint32_t>(53ULL):{
gekko_mem(in,std::string("stfsu",5),gekko_fr(gekko_rs(w)),w);
break;}
case cast<uint32_t>(54ULL):{
gekko_mem(in,std::string("stfd",4),gekko_fr(gekko_rs(w)),w);
break;}
case cast<uint32_t>(55ULL):{
gekko_mem(in,std::string("stfdu",5),gekko_fr(gekko_rs(w)),w);
break;}
case cast<uint32_t>(56ULL):case cast<uint32_t>(57ULL):case cast<uint32_t>(60ULL):case cast<uint32_t>(61ULL):{
gekko_decodePSQ(in,w);
break;}
case cast<uint32_t>(59ULL):{
gekko_decode59(in,w);
break;}
case cast<uint32_t>(63ULL):{
gekko_decode63(in,w);
break;}
}}
}
}
// tools/cpu/gekko/decode.go:210:1
void gekko_mem(gekko_Inst* in,std::string mnem,std::string reg,uint32_t w){
{
int32_t d = gekko_simm(w);
if ((gekko_ra(w) == cast<uint32_t>(0ULL))) {
gekko_Inst_set(in,mnem,std::string("%s,%d(0)",8),reg,d);
return ;
}
gekko_Inst_set(in,mnem,std::string("%s,%d(%s)",9),reg,d,gekko_r(gekko_ra(w)));
}
}
// tools/cpu/gekko/decode.go:222:1
void gekko_memx(gekko_Inst* in,std::string mnem,std::string reg,uint32_t w){
{
std::string base = gekko_r(gekko_ra(w));
if ((gekko_ra(w) == cast<uint32_t>(0ULL))) {
base = std::string("0",1);
}
gekko_Inst_set(in,mnem,std::string("%s,%s,%s",8),reg,base,gekko_r(gekko_rb(w)));
}
}
// tools/cpu/gekko/decode.go:234:1
void gekko_decodeRlwinm(gekko_Inst* in,uint32_t w){
{
auto tmp12 = std::make_tuple(gekko_shOf(w),gekko_mbOf(w),gekko_meOf(w));
uint32_t sh = std::get<0>(tmp12);
uint32_t mb = std::get<1>(tmp12);
uint32_t me = std::get<2>(tmp12);
std::string name = gekko_dot(std::string("rlwinm",6),w);
{
if (((mb == cast<uint32_t>(0ULL)) && (me == cast<uint32_t>((cast<uint32_t>(31ULL) - sh))))){
gekko_Inst_set(in,gekko_dot(std::string("slwi",4),w),std::string("%s,%s,%d",8),gekko_r(gekko_ra(w)),gekko_r(gekko_rs(w)),sh);
}
else if ((((me == cast<uint32_t>(31ULL)) && (mb == cast<uint32_t>((cast<uint32_t>(32ULL) - sh)))) && (sh != cast<uint32_t>(0ULL)))){
gekko_Inst_set(in,gekko_dot(std::string("srwi",4),w),std::string("%s,%s,%d",8),gekko_r(gekko_ra(w)),gekko_r(gekko_rs(w)),cast<uint32_t>((cast<uint32_t>(32ULL) - sh)));
}
else if (((sh == cast<uint32_t>(0ULL)) && (mb == cast<uint32_t>(0ULL)))){
gekko_Inst_set(in,gekko_dot(std::string("clrrwi",6),w),std::string("%s,%s,%d",8),gekko_r(gekko_ra(w)),gekko_r(gekko_rs(w)),cast<uint32_t>((cast<uint32_t>(31ULL) - me)));
}
else if (((sh == cast<uint32_t>(0ULL)) && (me == cast<uint32_t>(31ULL)))){
gekko_Inst_set(in,gekko_dot(std::string("clrlwi",6),w),std::string("%s,%s,%d",8),gekko_r(gekko_ra(w)),gekko_r(gekko_rs(w)),mb);
}
else {
gekko_Inst_set(in,name,std::string("%s,%s,%d,%d,%d",14),gekko_r(gekko_ra(w)),gekko_r(gekko_rs(w)),sh,mb,me);
}
}
tmp13:;
}
}
// tools/cpu/gekko/decode.go:265:1
bool gekko_boAlways(uint32_t bo){
{
return ((cast<uint32_t>((bo & cast<uint32_t>(16ULL))) != cast<uint32_t>(0ULL)) && (cast<uint32_t>((bo & cast<uint32_t>(4ULL))) != cast<uint32_t>(0ULL)));
}
}
// tools/cpu/gekko/decode.go:276:1
std::tuple<std::string,bool> gekko_branchMnem(uint32_t bo,uint32_t bi,std::string suffix,bool link){
{
std::string l = std::string("",0);
if (link) {
l = std::string("l",1);
}
bool dec = (cast<uint32_t>((bo & cast<uint32_t>(4ULL))) == cast<uint32_t>(0ULL));
bool cond = (cast<uint32_t>((bo & cast<uint32_t>(16ULL))) == cast<uint32_t>(0ULL));
{
if ((dec && (!cond))){
std::string name = std::string("bdnz",4);
if ((cast<uint32_t>((bo & cast<uint32_t>(2ULL))) != cast<uint32_t>(0ULL))) {
name = std::string("bdz",3);
}
return {((name + suffix) + l),true};
}
else if ((dec && cond)){
std::string name = std::string("bdnz",4);
if ((cast<uint32_t>((bo & cast<uint32_t>(2ULL))) != cast<uint32_t>(0ULL))) {
name = std::string("bdz",3);
}
if ((cast<uint32_t>((bo & cast<uint32_t>(8ULL))) != cast<uint32_t>(0ULL))) {
name += std::string("t",1);
}
else {
name += std::string("f",1);
}
return {((name + suffix) + l),false};
}
else if (((!dec) && cond)){
std::array<std::string,4> tbl = gekko_brFalse;
if ((cast<uint32_t>((bo & cast<uint32_t>(8ULL))) != cast<uint32_t>(0ULL))) {
tbl = gekko_brTrue;
}
return {((tbl[cast<uint32_t>((bi & cast<uint32_t>(3ULL)))] + suffix) + l),true};
}
}
tmp14:;
return {((std::string("b",1) + suffix) + l),true};
}
}
// tools/cpu/gekko/decode.go:315:1
void gekko_decodeBC(gekko_Inst* in,uint32_t w,uint32_t addr){
{
auto tmp15 = std::make_tuple(gekko_rs(w),gekko_ra(w));
uint32_t bo = std::get<0>(tmp15);
uint32_t bi = std::get<1>(tmp15);
int32_t d = cast<int32_t>(cast<int16_t>(cast<uint32_t>((w & cast<uint32_t>(65532ULL)))));
uint32_t target = cast<uint32_t>(cast<int32_t>((cast<int32_t>(addr) + d)));
if (gekko_aa(w)) {
target = cast<uint32_t>(d);
}
auto tmp16 = std::make_tuple(target,true);
in->Target = std::get<0>(tmp16);
in->HasTarget = std::get<1>(tmp16);
auto tmp17 = gekko_branchMnem(bo,bi,std::string("",0),gekko_lk(w));
std::string name = std::get<0>(tmp17);
bool complete = std::get<1>(tmp17);
uint32_t crf = shr<uint32_t>(bi,cast<int64_t>(2ULL));
{
if ((complete && gekko_boAlways(bo))){
gekko_Inst_set(in,name,std::string("0x%08X",6),target);
}
else if (complete){
if ((crf == cast<uint32_t>(0ULL))) {
gekko_Inst_set(in,name,std::string("0x%08X",6),target);
}
else {
gekko_Inst_set(in,name,std::string("cr%d,0x%08X",11),crf,target);
}
}
else {
gekko_Inst_set(in,name,std::string("cr%d[%s],0x%08X",15),crf,gekko_crBitName[cast<uint32_t>((bi & cast<uint32_t>(3ULL)))],target);
}
}
tmp18:;
{
if (gekko_lk(w)){
in->Flow = cast<gekko_Flow>(3ULL);
}
else if (gekko_boAlways(bo)){
in->Flow = cast<gekko_Flow>(2ULL);
}
else {
in->Flow = cast<gekko_Flow>(1ULL);
}
}
tmp19:;
}
}
// tools/cpu/gekko/decode.go:350:1
void gekko_decodeB(gekko_Inst* in,uint32_t w,uint32_t addr){
{
int32_t d = shr<int32_t>(shl<int32_t>(cast<int32_t>(cast<uint32_t>((w & cast<uint32_t>(67108860ULL)))),cast<int64_t>(6ULL)),cast<int64_t>(6ULL));
uint32_t target = cast<uint32_t>(cast<int32_t>((cast<int32_t>(addr) + d)));
if (gekko_aa(w)) {
target = cast<uint32_t>(d);
}
auto tmp20 = std::make_tuple(target,true);
in->Target = std::get<0>(tmp20);
in->HasTarget = std::get<1>(tmp20);
std::string name = std::string("b",1);
if (gekko_aa(w)) {
name += std::string("a",1);
}
if (gekko_lk(w)) {
name += std::string("l",1);
}
gekko_Inst_set(in,name,std::string("0x%08X",6),target);
{
if ((!gekko_lk(w))){
in->Flow = cast<gekko_Flow>(2ULL);
}
else if ((target == cast<uint32_t>((addr + cast<uint32_t>(4ULL))))){
in->Flow = cast<gekko_Flow>(0ULL);
}
else {
in->Flow = cast<gekko_Flow>(3ULL);
}
}
tmp21:;
}
}
// tools/cpu/gekko/decode.go:383:1
void gekko_decode19(gekko_Inst* in,uint32_t w){
{
auto tmp22 = std::make_tuple(gekko_rs(w),gekko_ra(w));
uint32_t bo = std::get<0>(tmp22);
uint32_t bi = std::get<1>(tmp22);
{
switch(gekko_xo10(w)){
case cast<uint32_t>(0ULL):{
gekko_Inst_set(in,std::string("mcrf",4),std::string("cr%d,cr%d",9),gekko_crfD(w),gekko_crfS(w));
break;}
case cast<uint32_t>(16ULL):case cast<uint32_t>(528ULL):{
bool toCTR = (gekko_xo10(w) == cast<uint32_t>(528ULL));
std::string suffix = std::string("lr",2);
if (toCTR) {
suffix = std::string("ctr",3);
}
auto tmp23 = gekko_branchMnem(bo,bi,suffix,gekko_lk(w));
std::string name = std::get<0>(tmp23);
bool complete = std::get<1>(tmp23);
uint32_t crf = shr<uint32_t>(bi,cast<int64_t>(2ULL));
{
if ((complete && gekko_boAlways(bo))){
gekko_Inst_set(in,name,std::string("",0));
}
else if (complete){
gekko_Inst_set(in,name,std::string("cr%d",4),crf);
}
else {
gekko_Inst_set(in,name,std::string("cr%d[%s]",8),crf,gekko_crBitName[cast<uint32_t>((bi & cast<uint32_t>(3ULL)))]);
}
}
tmp24:;
{
if (gekko_lk(w)){
in->Flow = cast<gekko_Flow>(6ULL);
}
else if ((!gekko_boAlways(bo))){
in->Flow = cast<gekko_Flow>(1ULL);
}
else if (toCTR){
in->Flow = cast<gekko_Flow>(5ULL);
}
else {
in->Flow = cast<gekko_Flow>(4ULL);
}
}
tmp25:;
break;}
case cast<uint32_t>(50ULL):{
gekko_Inst_set(in,std::string("rfi",3),std::string("",0));
in->Flow = cast<gekko_Flow>(7ULL);
break;}
case cast<uint32_t>(150ULL):{
gekko_Inst_set(in,std::string("isync",5),std::string("",0));
break;}
case cast<uint32_t>(33ULL):case cast<uint32_t>(129ULL):case cast<uint32_t>(193ULL):case cast<uint32_t>(225ULL):case cast<uint32_t>(257ULL):case cast<uint32_t>(289ULL):case cast<uint32_t>(417ULL):case cast<uint32_t>(449ULL):{
gekko_crLogic(in,w);
break;}
}}
}
}
// tools/cpu/gekko/decode.go:433:1
void gekko_crLogic(gekko_Inst* in,uint32_t w){
{
std::string name = get(gekko_crLogicName,gekko_xo10(w));
auto tmp26 = std::make_tuple(gekko_rs(w),gekko_ra(w),gekko_rb(w));
uint32_t d = std::get<0>(tmp26);
uint32_t a = std::get<1>(tmp26);
uint32_t b = std::get<2>(tmp26);
{
if ((((name == std::string("crxor",5)) && (d == a)) && (a == b))){
gekko_Inst_set(in,std::string("crclr",5),std::string("cr%d[%s]",8),shr<uint32_t>(d,cast<int64_t>(2ULL)),gekko_crBitName[cast<uint32_t>((d & cast<uint32_t>(3ULL)))]);
}
else if ((((name == std::string("creqv",5)) && (d == a)) && (a == b))){
gekko_Inst_set(in,std::string("crset",5),std::string("cr%d[%s]",8),shr<uint32_t>(d,cast<int64_t>(2ULL)),gekko_crBitName[cast<uint32_t>((d & cast<uint32_t>(3ULL)))]);
}
else if (((name == std::string("cror",4)) && (a == b))){
gekko_Inst_set(in,std::string("crmove",6),std::string("cr%d[%s],cr%d[%s]",17),shr<uint32_t>(d,cast<int64_t>(2ULL)),gekko_crBitName[cast<uint32_t>((d & cast<uint32_t>(3ULL)))],shr<uint32_t>(a,cast<int64_t>(2ULL)),gekko_crBitName[cast<uint32_t>((a & cast<uint32_t>(3ULL)))]);
}
else {
gekko_Inst_set(in,name,std::string("cr%d[%s],cr%d[%s],cr%d[%s]",26),shr<uint32_t>(d,cast<int64_t>(2ULL)),gekko_crBitName[cast<uint32_t>((d & cast<uint32_t>(3ULL)))],shr<uint32_t>(a,cast<int64_t>(2ULL)),gekko_crBitName[cast<uint32_t>((a & cast<uint32_t>(3ULL)))],shr<uint32_t>(b,cast<int64_t>(2ULL)),gekko_crBitName[cast<uint32_t>((b & cast<uint32_t>(3ULL)))]);
}
}
tmp27:;
}
}
// tools/cpu/gekko/decode.go:466:1
void gekko_decode31(gekko_Inst* in,uint32_t w){
{
auto tmp28 = std::make_tuple(gekko_rs(w),gekko_ra(w),gekko_rb(w));
uint32_t d = std::get<0>(tmp28);
uint32_t a = std::get<1>(tmp28);
uint32_t b = std::get<2>(tmp28);
uint32_t x = gekko_xo10(w);
if (((x >= cast<uint32_t>(512ULL)) && get(gekko_oeForm,cast<uint32_t>((x - cast<uint32_t>(512ULL)))))) {
x -= cast<uint32_t>(512ULL);
}
{
switch(x){
case cast<uint32_t>(0ULL):{
gekko_Inst_set(in,std::string("cmpw",4),std::string("cr%d,%s,%s",10),gekko_crfD(w),gekko_r(a),gekko_r(b));
break;}
case cast<uint32_t>(32ULL):{
gekko_Inst_set(in,std::string("cmplw",5),std::string("cr%d,%s,%s",10),gekko_crfD(w),gekko_r(a),gekko_r(b));
break;}
case cast<uint32_t>(4ULL):{
gekko_Inst_set(in,std::string("tw",2),std::string("%d,%s,%s",8),d,gekko_r(a),gekko_r(b));
in->Flow = cast<gekko_Flow>(7ULL);
break;}
case cast<uint32_t>(8ULL):{
gekko_Inst_set(in,gekko_oeDot(std::string("subfc",5),w),std::string("%s,%s,%s",8),gekko_r(d),gekko_r(a),gekko_r(b));
break;}
case cast<uint32_t>(10ULL):{
gekko_Inst_set(in,gekko_oeDot(std::string("addc",4),w),std::string("%s,%s,%s",8),gekko_r(d),gekko_r(a),gekko_r(b));
break;}
case cast<uint32_t>(11ULL):{
gekko_Inst_set(in,gekko_dot(std::string("mulhwu",6),w),std::string("%s,%s,%s",8),gekko_r(d),gekko_r(a),gekko_r(b));
break;}
case cast<uint32_t>(40ULL):{
gekko_Inst_set(in,gekko_oeDot(std::string("subf",4),w),std::string("%s,%s,%s",8),gekko_r(d),gekko_r(a),gekko_r(b));
break;}
case cast<uint32_t>(75ULL):{
gekko_Inst_set(in,gekko_dot(std::string("mulhw",5),w),std::string("%s,%s,%s",8),gekko_r(d),gekko_r(a),gekko_r(b));
break;}
case cast<uint32_t>(104ULL):{
gekko_Inst_set(in,gekko_oeDot(std::string("neg",3),w),std::string("%s,%s",5),gekko_r(d),gekko_r(a));
break;}
case cast<uint32_t>(136ULL):{
gekko_Inst_set(in,gekko_oeDot(std::string("subfe",5),w),std::string("%s,%s,%s",8),gekko_r(d),gekko_r(a),gekko_r(b));
break;}
case cast<uint32_t>(138ULL):{
gekko_Inst_set(in,gekko_oeDot(std::string("adde",4),w),std::string("%s,%s,%s",8),gekko_r(d),gekko_r(a),gekko_r(b));
break;}
case cast<uint32_t>(200ULL):{
gekko_Inst_set(in,gekko_oeDot(std::string("subfze",6),w),std::string("%s,%s",5),gekko_r(d),gekko_r(a));
break;}
case cast<uint32_t>(202ULL):{
gekko_Inst_set(in,gekko_oeDot(std::string("addze",5),w),std::string("%s,%s",5),gekko_r(d),gekko_r(a));
break;}
case cast<uint32_t>(232ULL):{
gekko_Inst_set(in,gekko_oeDot(std::string("subfme",6),w),std::string("%s,%s",5),gekko_r(d),gekko_r(a));
break;}
case cast<uint32_t>(234ULL):{
gekko_Inst_set(in,gekko_oeDot(std::string("addme",5),w),std::string("%s,%s",5),gekko_r(d),gekko_r(a));
break;}
case cast<uint32_t>(235ULL):{
gekko_Inst_set(in,gekko_oeDot(std::string("mullw",5),w),std::string("%s,%s,%s",8),gekko_r(d),gekko_r(a),gekko_r(b));
break;}
case cast<uint32_t>(266ULL):{
gekko_Inst_set(in,gekko_oeDot(std::string("add",3),w),std::string("%s,%s,%s",8),gekko_r(d),gekko_r(a),gekko_r(b));
break;}
case cast<uint32_t>(459ULL):{
gekko_Inst_set(in,gekko_oeDot(std::string("divwu",5),w),std::string("%s,%s,%s",8),gekko_r(d),gekko_r(a),gekko_r(b));
break;}
case cast<uint32_t>(491ULL):{
gekko_Inst_set(in,gekko_oeDot(std::string("divw",4),w),std::string("%s,%s,%s",8),gekko_r(d),gekko_r(a),gekko_r(b));
break;}
case cast<uint32_t>(28ULL):{
gekko_Inst_set(in,gekko_dot(std::string("and",3),w),std::string("%s,%s,%s",8),gekko_r(a),gekko_r(d),gekko_r(b));
break;}
case cast<uint32_t>(60ULL):{
gekko_Inst_set(in,gekko_dot(std::string("andc",4),w),std::string("%s,%s,%s",8),gekko_r(a),gekko_r(d),gekko_r(b));
break;}
case cast<uint32_t>(124ULL):{
if ((d == b)) {
gekko_Inst_set(in,gekko_dot(std::string("not",3),w),std::string("%s,%s",5),gekko_r(a),gekko_r(d));
}
else {
gekko_Inst_set(in,gekko_dot(std::string("nor",3),w),std::string("%s,%s,%s",8),gekko_r(a),gekko_r(d),gekko_r(b));
}
break;}
case cast<uint32_t>(284ULL):{
gekko_Inst_set(in,gekko_dot(std::string("eqv",3),w),std::string("%s,%s,%s",8),gekko_r(a),gekko_r(d),gekko_r(b));
break;}
case cast<uint32_t>(316ULL):{
gekko_Inst_set(in,gekko_dot(std::string("xor",3),w),std::string("%s,%s,%s",8),gekko_r(a),gekko_r(d),gekko_r(b));
break;}
case cast<uint32_t>(412ULL):{
gekko_Inst_set(in,gekko_dot(std::string("orc",3),w),std::string("%s,%s,%s",8),gekko_r(a),gekko_r(d),gekko_r(b));
break;}
case cast<uint32_t>(444ULL):{
if ((d == b)) {
gekko_Inst_set(in,gekko_dot(std::string("mr",2),w),std::string("%s,%s",5),gekko_r(a),gekko_r(d));
}
else {
gekko_Inst_set(in,gekko_dot(std::string("or",2),w),std::string("%s,%s,%s",8),gekko_r(a),gekko_r(d),gekko_r(b));
}
break;}
case cast<uint32_t>(476ULL):{
gekko_Inst_set(in,gekko_dot(std::string("nand",4),w),std::string("%s,%s,%s",8),gekko_r(a),gekko_r(d),gekko_r(b));
break;}
case cast<uint32_t>(24ULL):{
gekko_Inst_set(in,gekko_dot(std::string("slw",3),w),std::string("%s,%s,%s",8),gekko_r(a),gekko_r(d),gekko_r(b));
break;}
case cast<uint32_t>(536ULL):{
gekko_Inst_set(in,gekko_dot(std::string("srw",3),w),std::string("%s,%s,%s",8),gekko_r(a),gekko_r(d),gekko_r(b));
break;}
case cast<uint32_t>(792ULL):{
gekko_Inst_set(in,gekko_dot(std::string("sraw",4),w),std::string("%s,%s,%s",8),gekko_r(a),gekko_r(d),gekko_r(b));
break;}
case cast<uint32_t>(824ULL):{
gekko_Inst_set(in,gekko_dot(std::string("srawi",5),w),std::string("%s,%s,%d",8),gekko_r(a),gekko_r(d),gekko_shOf(w));
break;}
case cast<uint32_t>(26ULL):{
gekko_Inst_set(in,gekko_dot(std::string("cntlzw",6),w),std::string("%s,%s",5),gekko_r(a),gekko_r(d));
break;}
case cast<uint32_t>(922ULL):{
gekko_Inst_set(in,gekko_dot(std::string("extsh",5),w),std::string("%s,%s",5),gekko_r(a),gekko_r(d));
break;}
case cast<uint32_t>(954ULL):{
gekko_Inst_set(in,gekko_dot(std::string("extsb",5),w),std::string("%s,%s",5),gekko_r(a),gekko_r(d));
break;}
case cast<uint32_t>(19ULL):{
gekko_Inst_set(in,std::string("mfcr",4),std::string("%s",2),gekko_r(d));
break;}
case cast<uint32_t>(144ULL):{
uint32_t mask = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(12ULL))) & cast<uint32_t>(255ULL)));
if ((mask == cast<uint32_t>(255ULL))) {
gekko_Inst_set(in,std::string("mtcr",4),std::string("%s",2),gekko_r(d));
}
else {
gekko_Inst_set(in,std::string("mtcrf",5),std::string("0x%02X,%s",9),mask,gekko_r(d));
}
break;}
case cast<uint32_t>(339ULL):{
gekko_Inst_set(in,std::string("mfspr",5),std::string("%s,%s",5),gekko_r(d),gekko_sprStr(gekko_sprOf(w)));
break;}
case cast<uint32_t>(467ULL):{
gekko_Inst_set(in,std::string("mtspr",5),std::string("%s,%s",5),gekko_sprStr(gekko_sprOf(w)),gekko_r(d));
break;}
case cast<uint32_t>(371ULL):{
gekko_Inst_set(in,std::string("mftb",4),std::string("%s,%s",5),gekko_r(d),gekko_sprStr(gekko_sprOf(w)));
break;}
case cast<uint32_t>(83ULL):{
gekko_Inst_set(in,std::string("mfmsr",5),std::string("%s",2),gekko_r(d));
break;}
case cast<uint32_t>(146ULL):{
gekko_Inst_set(in,std::string("mtmsr",5),std::string("%s",2),gekko_r(d));
break;}
case cast<uint32_t>(210ULL):{
gekko_Inst_set(in,std::string("mtsr",4),std::string("%d,%s",5),cast<uint32_t>((a & cast<uint32_t>(15ULL))),gekko_r(d));
break;}
case cast<uint32_t>(242ULL):{
gekko_Inst_set(in,std::string("mtsrin",6),std::string("%s,%s",5),gekko_r(d),gekko_r(b));
break;}
case cast<uint32_t>(595ULL):{
gekko_Inst_set(in,std::string("mfsr",4),std::string("%s,%d",5),gekko_r(d),cast<uint32_t>((a & cast<uint32_t>(15ULL))));
break;}
case cast<uint32_t>(659ULL):{
gekko_Inst_set(in,std::string("mfsrin",6),std::string("%s,%s",5),gekko_r(d),gekko_r(b));
break;}
case cast<uint32_t>(512ULL):{
gekko_Inst_set(in,std::string("mcrxr",5),std::string("cr%d",4),gekko_crfD(w));
break;}
case cast<uint32_t>(23ULL):{
gekko_memx(in,std::string("lwzx",4),gekko_r(d),w);
break;}
case cast<uint32_t>(55ULL):{
gekko_memx(in,std::string("lwzux",5),gekko_r(d),w);
break;}
case cast<uint32_t>(87ULL):{
gekko_memx(in,std::string("lbzx",4),gekko_r(d),w);
break;}
case cast<uint32_t>(119ULL):{
gekko_memx(in,std::string("lbzux",5),gekko_r(d),w);
break;}
case cast<uint32_t>(279ULL):{
gekko_memx(in,std::string("lhzx",4),gekko_r(d),w);
break;}
case cast<uint32_t>(311ULL):{
gekko_memx(in,std::string("lhzux",5),gekko_r(d),w);
break;}
case cast<uint32_t>(343ULL):{
gekko_memx(in,std::string("lhax",4),gekko_r(d),w);
break;}
case cast<uint32_t>(375ULL):{
gekko_memx(in,std::string("lhaux",5),gekko_r(d),w);
break;}
case cast<uint32_t>(151ULL):{
gekko_memx(in,std::string("stwx",4),gekko_r(d),w);
break;}
case cast<uint32_t>(183ULL):{
gekko_memx(in,std::string("stwux",5),gekko_r(d),w);
break;}
case cast<uint32_t>(215ULL):{
gekko_memx(in,std::string("stbx",4),gekko_r(d),w);
break;}
case cast<uint32_t>(247ULL):{
gekko_memx(in,std::string("stbux",5),gekko_r(d),w);
break;}
case cast<uint32_t>(407ULL):{
gekko_memx(in,std::string("sthx",4),gekko_r(d),w);
break;}
case cast<uint32_t>(439ULL):{
gekko_memx(in,std::string("sthux",5),gekko_r(d),w);
break;}
case cast<uint32_t>(534ULL):{
gekko_memx(in,std::string("lwbrx",5),gekko_r(d),w);
break;}
case cast<uint32_t>(790ULL):{
gekko_memx(in,std::string("lhbrx",5),gekko_r(d),w);
break;}
case cast<uint32_t>(662ULL):{
gekko_memx(in,std::string("stwbrx",6),gekko_r(d),w);
break;}
case cast<uint32_t>(918ULL):{
gekko_memx(in,std::string("sthbrx",6),gekko_r(d),w);
break;}
case cast<uint32_t>(533ULL):{
gekko_memx(in,std::string("lswx",4),gekko_r(d),w);
break;}
case cast<uint32_t>(597ULL):{
gekko_Inst_set(in,std::string("lswi",4),std::string("%s,%s,%d",8),gekko_r(d),gekko_r(a),gekko_rb(w));
break;}
case cast<uint32_t>(661ULL):{
gekko_memx(in,std::string("stswx",5),gekko_r(d),w);
break;}
case cast<uint32_t>(725ULL):{
gekko_Inst_set(in,std::string("stswi",5),std::string("%s,%s,%d",8),gekko_r(d),gekko_r(a),gekko_rb(w));
break;}
case cast<uint32_t>(20ULL):{
gekko_memx(in,std::string("lwarx",5),gekko_r(d),w);
break;}
case cast<uint32_t>(150ULL):{
gekko_memx(in,std::string("stwcx.",6),gekko_r(d),w);
break;}
case cast<uint32_t>(535ULL):{
gekko_memx(in,std::string("lfsx",4),gekko_fr(d),w);
break;}
case cast<uint32_t>(567ULL):{
gekko_memx(in,std::string("lfsux",5),gekko_fr(d),w);
break;}
case cast<uint32_t>(599ULL):{
gekko_memx(in,std::string("lfdx",4),gekko_fr(d),w);
break;}
case cast<uint32_t>(631ULL):{
gekko_memx(in,std::string("lfdux",5),gekko_fr(d),w);
break;}
case cast<uint32_t>(663ULL):{
gekko_memx(in,std::string("stfsx",5),gekko_fr(d),w);
break;}
case cast<uint32_t>(695ULL):{
gekko_memx(in,std::string("stfsux",6),gekko_fr(d),w);
break;}
case cast<uint32_t>(727ULL):{
gekko_memx(in,std::string("stfdx",5),gekko_fr(d),w);
break;}
case cast<uint32_t>(759ULL):{
gekko_memx(in,std::string("stfdux",6),gekko_fr(d),w);
break;}
case cast<uint32_t>(983ULL):{
gekko_memx(in,std::string("stfiwx",6),gekko_fr(d),w);
break;}
case cast<uint32_t>(54ULL):{
gekko_Inst_set(in,std::string("dcbst",5),std::string("%s,%s",5),gekko_r(a),gekko_r(b));
break;}
case cast<uint32_t>(86ULL):{
gekko_Inst_set(in,std::string("dcbf",4),std::string("%s,%s",5),gekko_r(a),gekko_r(b));
break;}
case cast<uint32_t>(246ULL):{
gekko_Inst_set(in,std::string("dcbtst",6),std::string("%s,%s",5),gekko_r(a),gekko_r(b));
break;}
case cast<uint32_t>(278ULL):{
gekko_Inst_set(in,std::string("dcbt",4),std::string("%s,%s",5),gekko_r(a),gekko_r(b));
break;}
case cast<uint32_t>(470ULL):{
gekko_Inst_set(in,std::string("dcbi",4),std::string("%s,%s",5),gekko_r(a),gekko_r(b));
break;}
case cast<uint32_t>(1014ULL):{
gekko_Inst_set(in,std::string("dcbz",4),std::string("%s,%s",5),gekko_r(a),gekko_r(b));
break;}
case cast<uint32_t>(982ULL):{
gekko_Inst_set(in,std::string("icbi",4),std::string("%s,%s",5),gekko_r(a),gekko_r(b));
break;}
case cast<uint32_t>(598ULL):{
gekko_Inst_set(in,std::string("sync",4),std::string("",0));
break;}
case cast<uint32_t>(854ULL):{
gekko_Inst_set(in,std::string("eieio",5),std::string("",0));
break;}
case cast<uint32_t>(306ULL):{
gekko_Inst_set(in,std::string("tlbie",5),std::string("%s",2),gekko_r(b));
break;}
case cast<uint32_t>(370ULL):{
gekko_Inst_set(in,std::string("tlbia",5),std::string("",0));
break;}
case cast<uint32_t>(566ULL):{
gekko_Inst_set(in,std::string("tlbsync",7),std::string("",0));
break;}
case cast<uint32_t>(310ULL):{
gekko_memx(in,std::string("eciwx",5),gekko_r(d),w);
break;}
case cast<uint32_t>(438ULL):{
gekko_memx(in,std::string("ecowx",5),gekko_r(d),w);
break;}
}}
}
}
// tools/cpu/gekko/decode.go:711:1
void gekko_decode59(gekko_Inst* in,uint32_t w){
{
auto tmp29 = std::make_tuple(gekko_rs(w),gekko_ra(w),gekko_rb(w),gekko_rc(w));
uint32_t d = std::get<0>(tmp29);
uint32_t a = std::get<1>(tmp29);
uint32_t b = std::get<2>(tmp29);
uint32_t c = std::get<3>(tmp29);
{
switch(gekko_xo5(w)){
case cast<uint32_t>(18ULL):{
gekko_Inst_set(in,gekko_dot(std::string("fdivs",5),w),std::string("%s,%s,%s",8),gekko_fr(d),gekko_fr(a),gekko_fr(b));
break;}
case cast<uint32_t>(20ULL):{
gekko_Inst_set(in,gekko_dot(std::string("fsubs",5),w),std::string("%s,%s,%s",8),gekko_fr(d),gekko_fr(a),gekko_fr(b));
break;}
case cast<uint32_t>(21ULL):{
gekko_Inst_set(in,gekko_dot(std::string("fadds",5),w),std::string("%s,%s,%s",8),gekko_fr(d),gekko_fr(a),gekko_fr(b));
break;}
case cast<uint32_t>(22ULL):{
gekko_Inst_set(in,gekko_dot(std::string("fsqrts",6),w),std::string("%s,%s",5),gekko_fr(d),gekko_fr(b));
break;}
case cast<uint32_t>(24ULL):{
gekko_Inst_set(in,gekko_dot(std::string("fres",4),w),std::string("%s,%s",5),gekko_fr(d),gekko_fr(b));
break;}
case cast<uint32_t>(25ULL):{
gekko_Inst_set(in,gekko_dot(std::string("fmuls",5),w),std::string("%s,%s,%s",8),gekko_fr(d),gekko_fr(a),gekko_fr(c));
break;}
case cast<uint32_t>(28ULL):{
gekko_Inst_set(in,gekko_dot(std::string("fmsubs",6),w),std::string("%s,%s,%s,%s",11),gekko_fr(d),gekko_fr(a),gekko_fr(c),gekko_fr(b));
break;}
case cast<uint32_t>(29ULL):{
gekko_Inst_set(in,gekko_dot(std::string("fmadds",6),w),std::string("%s,%s,%s,%s",11),gekko_fr(d),gekko_fr(a),gekko_fr(c),gekko_fr(b));
break;}
case cast<uint32_t>(30ULL):{
gekko_Inst_set(in,gekko_dot(std::string("fnmsubs",7),w),std::string("%s,%s,%s,%s",11),gekko_fr(d),gekko_fr(a),gekko_fr(c),gekko_fr(b));
break;}
case cast<uint32_t>(31ULL):{
gekko_Inst_set(in,gekko_dot(std::string("fnmadds",7),w),std::string("%s,%s,%s,%s",11),gekko_fr(d),gekko_fr(a),gekko_fr(c),gekko_fr(b));
break;}
}}
}
}
// tools/cpu/gekko/decode.go:740:1
void gekko_decode63(gekko_Inst* in,uint32_t w){
{
auto tmp30 = std::make_tuple(gekko_rs(w),gekko_ra(w),gekko_rb(w),gekko_rc(w));
uint32_t d = std::get<0>(tmp30);
uint32_t a = std::get<1>(tmp30);
uint32_t b = std::get<2>(tmp30);
uint32_t c = std::get<3>(tmp30);
{
switch(gekko_xo5(w)){
case cast<uint32_t>(18ULL):{
gekko_Inst_set(in,gekko_dot(std::string("fdiv",4),w),std::string("%s,%s,%s",8),gekko_fr(d),gekko_fr(a),gekko_fr(b));
return ;
break;}
case cast<uint32_t>(20ULL):{
gekko_Inst_set(in,gekko_dot(std::string("fsub",4),w),std::string("%s,%s,%s",8),gekko_fr(d),gekko_fr(a),gekko_fr(b));
return ;
break;}
case cast<uint32_t>(21ULL):{
gekko_Inst_set(in,gekko_dot(std::string("fadd",4),w),std::string("%s,%s,%s",8),gekko_fr(d),gekko_fr(a),gekko_fr(b));
return ;
break;}
case cast<uint32_t>(22ULL):{
gekko_Inst_set(in,gekko_dot(std::string("fsqrt",5),w),std::string("%s,%s",5),gekko_fr(d),gekko_fr(b));
return ;
break;}
case cast<uint32_t>(23ULL):{
gekko_Inst_set(in,gekko_dot(std::string("fsel",4),w),std::string("%s,%s,%s,%s",11),gekko_fr(d),gekko_fr(a),gekko_fr(c),gekko_fr(b));
return ;
break;}
case cast<uint32_t>(25ULL):{
gekko_Inst_set(in,gekko_dot(std::string("fmul",4),w),std::string("%s,%s,%s",8),gekko_fr(d),gekko_fr(a),gekko_fr(c));
return ;
break;}
case cast<uint32_t>(26ULL):{
gekko_Inst_set(in,gekko_dot(std::string("frsqrte",7),w),std::string("%s,%s",5),gekko_fr(d),gekko_fr(b));
return ;
break;}
case cast<uint32_t>(28ULL):{
gekko_Inst_set(in,gekko_dot(std::string("fmsub",5),w),std::string("%s,%s,%s,%s",11),gekko_fr(d),gekko_fr(a),gekko_fr(c),gekko_fr(b));
return ;
break;}
case cast<uint32_t>(29ULL):{
gekko_Inst_set(in,gekko_dot(std::string("fmadd",5),w),std::string("%s,%s,%s,%s",11),gekko_fr(d),gekko_fr(a),gekko_fr(c),gekko_fr(b));
return ;
break;}
case cast<uint32_t>(30ULL):{
gekko_Inst_set(in,gekko_dot(std::string("fnmsub",6),w),std::string("%s,%s,%s,%s",11),gekko_fr(d),gekko_fr(a),gekko_fr(c),gekko_fr(b));
return ;
break;}
case cast<uint32_t>(31ULL):{
gekko_Inst_set(in,gekko_dot(std::string("fnmadd",6),w),std::string("%s,%s,%s,%s",11),gekko_fr(d),gekko_fr(a),gekko_fr(c),gekko_fr(b));
return ;
break;}
}}
{
switch(gekko_xo10(w)){
case cast<uint32_t>(0ULL):{
gekko_Inst_set(in,std::string("fcmpu",5),std::string("cr%d,%s,%s",10),gekko_crfD(w),gekko_fr(a),gekko_fr(b));
break;}
case cast<uint32_t>(32ULL):{
gekko_Inst_set(in,std::string("fcmpo",5),std::string("cr%d,%s,%s",10),gekko_crfD(w),gekko_fr(a),gekko_fr(b));
break;}
case cast<uint32_t>(12ULL):{
gekko_Inst_set(in,gekko_dot(std::string("frsp",4),w),std::string("%s,%s",5),gekko_fr(d),gekko_fr(b));
break;}
case cast<uint32_t>(14ULL):{
gekko_Inst_set(in,gekko_dot(std::string("fctiw",5),w),std::string("%s,%s",5),gekko_fr(d),gekko_fr(b));
break;}
case cast<uint32_t>(15ULL):{
gekko_Inst_set(in,gekko_dot(std::string("fctiwz",6),w),std::string("%s,%s",5),gekko_fr(d),gekko_fr(b));
break;}
case cast<uint32_t>(40ULL):{
gekko_Inst_set(in,gekko_dot(std::string("fneg",4),w),std::string("%s,%s",5),gekko_fr(d),gekko_fr(b));
break;}
case cast<uint32_t>(72ULL):{
gekko_Inst_set(in,gekko_dot(std::string("fmr",3),w),std::string("%s,%s",5),gekko_fr(d),gekko_fr(b));
break;}
case cast<uint32_t>(136ULL):{
gekko_Inst_set(in,gekko_dot(std::string("fnabs",5),w),std::string("%s,%s",5),gekko_fr(d),gekko_fr(b));
break;}
case cast<uint32_t>(264ULL):{
gekko_Inst_set(in,gekko_dot(std::string("fabs",4),w),std::string("%s,%s",5),gekko_fr(d),gekko_fr(b));
break;}
case cast<uint32_t>(64ULL):{
gekko_Inst_set(in,std::string("mcrfs",5),std::string("cr%d,cr%d",9),gekko_crfD(w),gekko_crfS(w));
break;}
case cast<uint32_t>(38ULL):{
gekko_Inst_set(in,gekko_dot(std::string("mtfsb1",6),w),std::string("%d",2),d);
break;}
case cast<uint32_t>(70ULL):{
gekko_Inst_set(in,gekko_dot(std::string("mtfsb0",6),w),std::string("%d",2),d);
break;}
case cast<uint32_t>(134ULL):{
gekko_Inst_set(in,gekko_dot(std::string("mtfsfi",6),w),std::string("cr%d,%d",7),gekko_crfD(w),cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(12ULL))) & cast<uint32_t>(15ULL))));
break;}
case cast<uint32_t>(583ULL):{
gekko_Inst_set(in,gekko_dot(std::string("mffs",4),w),std::string("%s",2),gekko_fr(d));
break;}
case cast<uint32_t>(711ULL):{
gekko_Inst_set(in,gekko_dot(std::string("mtfsf",5),w),std::string("0x%02X,%s",9),cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(17ULL))) & cast<uint32_t>(255ULL))),gekko_fr(b));
break;}
}}
}
}
// tools/cpu/gekko/decode_ps.go:67:1
void gekko_decodePS(gekko_Inst* in,uint32_t w){
{
auto tmp31 = std::make_tuple(gekko_rs(w),gekko_ra(w),gekko_rb(w),gekko_rc(w));
uint32_t d = std::get<0>(tmp31);
uint32_t a = std::get<1>(tmp31);
uint32_t b = std::get<2>(tmp31);
uint32_t c = std::get<3>(tmp31);
{
auto tmp32 = lookup(gekko_psAForm,gekko_xo5(w));
std::string name = std::get<0>(tmp32);
bool ok = std::get<1>(tmp32);
if (ok) {
{
switch(get(gekko_psShape,gekko_xo5(w))){
case cast<int64_t>(0ULL):{
gekko_Inst_set(in,gekko_dot(name,w),std::string("%s,%s,%s",8),gekko_fr(d),gekko_fr(a),gekko_fr(b));
break;}
case cast<int64_t>(1ULL):{
gekko_Inst_set(in,gekko_dot(name,w),std::string("%s,%s,%s",8),gekko_fr(d),gekko_fr(a),gekko_fr(c));
break;}
case cast<int64_t>(2ULL):{
gekko_Inst_set(in,gekko_dot(name,w),std::string("%s,%s,%s,%s",11),gekko_fr(d),gekko_fr(a),gekko_fr(c),gekko_fr(b));
break;}
case cast<int64_t>(3ULL):{
gekko_Inst_set(in,gekko_dot(name,w),std::string("%s,%s",5),gekko_fr(d),gekko_fr(b));
break;}
}}
return ;
}
}
{
switch(gekko_xo6(w)){
case cast<uint32_t>(6ULL):{
gekko_psqX(in,std::string("psq_lx",6),w);
return ;
break;}
case cast<uint32_t>(7ULL):{
gekko_psqX(in,std::string("psq_stx",7),w);
return ;
break;}
case cast<uint32_t>(38ULL):{
gekko_psqX(in,std::string("psq_lux",7),w);
return ;
break;}
case cast<uint32_t>(39ULL):{
gekko_psqX(in,std::string("psq_stux",8),w);
return ;
break;}
}}
auto tmp33 = lookup(gekko_psX,gekko_xo10(w));
std::string name = std::get<0>(tmp33);
bool ok = std::get<1>(tmp33);
if ((!ok)) {
return ;
}
{
auto tmp35=name;
if (tmp35==(std::string("ps_cmpu0",8)) || tmp35==(std::string("ps_cmpo0",8)) || tmp35==(std::string("ps_cmpu1",8)) || tmp35==(std::string("ps_cmpo1",8))){
gekko_Inst_set(in,name,std::string("cr%d,%s,%s",10),gekko_crfD(w),gekko_fr(a),gekko_fr(b));
}
else if (tmp35==(std::string("dcbz_l",6))){
std::string base = gekko_r(a);
if ((a == cast<uint32_t>(0ULL))) {
base = std::string("0",1);
}
gekko_Inst_set(in,name,std::string("%s,%s",5),base,gekko_r(b));
}
else if (tmp35==(std::string("ps_merge00",10)) || tmp35==(std::string("ps_merge01",10)) || tmp35==(std::string("ps_merge10",10)) || tmp35==(std::string("ps_merge11",10))){
gekko_Inst_set(in,gekko_dot(name,w),std::string("%s,%s,%s",8),gekko_fr(d),gekko_fr(a),gekko_fr(b));
}
else {
gekko_Inst_set(in,gekko_dot(name,w),std::string("%s,%s",5),gekko_fr(d),gekko_fr(b));
}
}
tmp34:;
}
}
// tools/cpu/gekko/decode_ps.go:127:1
void gekko_psqX(gekko_Inst* in,std::string mnem,uint32_t w){
{
std::string base = gekko_r(gekko_ra(w));
if ((gekko_ra(w) == cast<uint32_t>(0ULL))) {
base = std::string("0",1);
}
uint32_t wBit = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(15ULL))) & cast<uint32_t>(1ULL)));
uint32_t gqr = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(12ULL))) & cast<uint32_t>(7ULL)));
gekko_Inst_set(in,mnem,std::string("%s,%s,%s,%d,gqr%d",17),gekko_fr(gekko_rs(w)),base,gekko_r(gekko_rb(w)),wBit,gqr);
}
}
// tools/cpu/gekko/decode_ps.go:140:1
void gekko_decodePSQ(gekko_Inst* in,uint32_t w){
{
std::string mnem={};
{
switch(gekko_opcd(w)){
case cast<uint32_t>(56ULL):{
mnem = std::string("psq_l",5);
break;}
case cast<uint32_t>(57ULL):{
mnem = std::string("psq_lu",6);
break;}
case cast<uint32_t>(60ULL):{
mnem = std::string("psq_st",6);
break;}
case cast<uint32_t>(61ULL):{
mnem = std::string("psq_stu",7);
break;}
}}
int32_t d = gekko_psqDisp(w);
uint32_t wBit = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(15ULL))) & cast<uint32_t>(1ULL)));
uint32_t gqr = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(12ULL))) & cast<uint32_t>(7ULL)));
std::string base = gekko_r(gekko_ra(w));
if ((gekko_ra(w) == cast<uint32_t>(0ULL))) {
base = std::string("0",1);
}
gekko_Inst_set(in,mnem,std::string("%s,%d(%s),%d,gqr%d",18),gekko_fr(gekko_rs(w)),d,base,wBit,gqr);
}
}
// tools/cpu/gekko/decode_ps.go:164:1
int32_t gekko_psqDisp(uint32_t w){
{
return shr<int32_t>(shl<int32_t>(cast<int32_t>(cast<uint32_t>((w & cast<uint32_t>(4095ULL)))),cast<int64_t>(20ULL)),cast<int64_t>(20ULL));
}
}
// tools/cpu/gekko/disasm.go:10:1
Slice<std::string> gekko_Disassemble(Slice<uint8_t> code,uint32_t base){
{
Slice<std::string> out={};
{int64_t i = cast<int64_t>(0ULL);for (;(cast<int64_t>((i + cast<int64_t>(4ULL))) <= len(code));i += cast<int64_t>(4ULL)){
gekko_Inst in = gekko_Decode(sub(code,i,len(code)),cast<uint32_t>((base + cast<uint32_t>(i))));
out = append(out,Slice<std::string>{go_fmt_Sprintf(std::string("%08X  %08X  %s",14),in.Addr,in.Word,in.Text)});
}
}return out;
}
}
// tools/cpu/gekko/exception.go:44:1
uint32_t gekko_CPU_vectorBase(gekko_CPU* c){
{
if ((cast<uint32_t>((c->MSR & cast<uint32_t>(64ULL))) != cast<uint32_t>(0ULL))) {
return cast<uint32_t>(4293918720ULL);
}
return cast<uint32_t>(0ULL);
}
}
// tools/cpu/gekko/exception.go:59:1
void gekko_CPU_Exception(gekko_CPU* c,uint32_t vec,uint32_t resume,uint32_t srr1Extra){
{
c->SRR0 = resume;
c->SRR1 = cast<uint32_t>(((cast<uint32_t>((c->MSR & cast<uint32_t>(2277572467ULL)))) | srr1Extra));
c->MSR &= ~(cast<uint32_t>(61234ULL));
c->PC = cast<uint32_t>((gekko_CPU_vectorBase(c) + vec));
}
}
// tools/cpu/gekko/exception.go:70:1
void gekko_CPU_programException(gekko_CPU* c,uint32_t kind){
{
gekko_CPU_Exception(c,cast<uint32_t>(1792ULL),c->PC,cast<uint32_t>((kind | cast<uint32_t>(65536ULL))));
}
}
// tools/cpu/gekko/exception.go:81:1
bool gekko_needsFPU(uint32_t w){
{
{
switch(gekko_opcd(w)){
case cast<uint32_t>(4ULL):case cast<uint32_t>(48ULL):case cast<uint32_t>(49ULL):case cast<uint32_t>(50ULL):case cast<uint32_t>(51ULL):case cast<uint32_t>(52ULL):case cast<uint32_t>(53ULL):case cast<uint32_t>(54ULL):case cast<uint32_t>(55ULL):case cast<uint32_t>(56ULL):case cast<uint32_t>(57ULL):case cast<uint32_t>(59ULL):case cast<uint32_t>(60ULL):case cast<uint32_t>(61ULL):case cast<uint32_t>(63ULL):{
return true;
break;}
case cast<uint32_t>(31ULL):{
{
switch(gekko_xo10(w)){
case cast<uint32_t>(535ULL):case cast<uint32_t>(567ULL):case cast<uint32_t>(599ULL):case cast<uint32_t>(631ULL):case cast<uint32_t>(663ULL):case cast<uint32_t>(695ULL):case cast<uint32_t>(727ULL):case cast<uint32_t>(759ULL):case cast<uint32_t>(983ULL):{
return true;
break;}
}}
break;}
}}
return false;
}
}
// tools/cpu/gekko/exception.go:114:1
bool gekko_CPU_fpUnavailable(gekko_CPU* c,uint32_t w,uint32_t pc){
{
if (((cast<uint32_t>((c->MSR & cast<uint32_t>(8192ULL))) != cast<uint32_t>(0ULL)) || (!gekko_needsFPU(w)))) {
return false;
}
gekko_CPU_Exception(c,cast<uint32_t>(2048ULL),pc,cast<uint32_t>(0ULL));
return true;
}
}
// tools/cpu/gekko/exception.go:126:1
bool gekko_CPU_checkInterrupt(gekko_CPU* c){
{
if ((cast<uint32_t>((c->MSR & cast<uint32_t>(32768ULL))) == cast<uint32_t>(0ULL))) {
return false;
}
if (c->ExtInt) {
gekko_CPU_Exception(c,cast<uint32_t>(1280ULL),c->PC,cast<uint32_t>(0ULL));
return true;
}
if (((cast<uint32_t>((c->DEC & cast<uint32_t>(2147483648ULL))) != cast<uint32_t>(0ULL)) && c->decArmed)) {
c->decArmed = false;
gekko_CPU_Exception(c,cast<uint32_t>(2304ULL),c->PC,cast<uint32_t>(0ULL));
return true;
}
return false;
}
}
// tools/cpu/gekko/exception.go:147:1
void gekko_CPU_Interrupt(gekko_CPU* c,bool pending){
{
c->ExtInt = pending;
}
}
// tools/cpu/gekko/exec.go:21:1
int64_t gekko_CPU_Step(gekko_CPU* c){
{
if (c->Halted) {
return cast<int64_t>(0ULL);
}
if (gekko_CPU_checkInterrupt(c)) {
c->Steps++;
gekko_CPU_tick(c,cast<int64_t>(1ULL));
return cast<int64_t>(1ULL);
}
uint32_t w = gekko_CPU_fetch(c,c->PC);
if (c->Halted) {
return cast<int64_t>(0ULL);
}
uint32_t pc = c->PC;
c->PC = cast<uint32_t>((pc + cast<uint32_t>(4ULL)));
gekko_CPU_execute(c,w,pc);
c->Steps++;
gekko_CPU_tick(c,cast<int64_t>(1ULL));
return cast<int64_t>(1ULL);
}
}
// tools/cpu/gekko/exec.go:45:1
void gekko_CPU_execute(gekko_CPU* c,uint32_t w,uint32_t pc){
{
if (gekko_CPU_fpUnavailable(c,w,pc)) {
return ;
}
{
switch(gekko_opcd(w)){
case cast<uint32_t>(3ULL):{
if (gekko_CPU_trapCond(c,gekko_rs(w),c->GPR[gekko_ra(w)],cast<uint32_t>(gekko_simm(w)))) {
gekko_CPU_programException(c,cast<uint32_t>(131072ULL));
}
break;}
case cast<uint32_t>(4ULL):{
gekko_CPU_execPS(c,w,pc);
break;}
case cast<uint32_t>(7ULL):{
c->GPR[gekko_rs(w)] = cast<uint32_t>(cast<int32_t>((cast<int32_t>(c->GPR[gekko_ra(w)]) * gekko_simm(w))));
break;}
case cast<uint32_t>(8ULL):{
uint32_t a = c->GPR[gekko_ra(w)];
uint32_t imm = cast<uint32_t>(gekko_simm(w));
uint32_t r = cast<uint32_t>((imm - a));
gekko_CPU_setCA(c,gekko_carrySub(imm,a));
c->GPR[gekko_rs(w)] = r;
break;}
case cast<uint32_t>(10ULL):{
gekko_CPU_compareLogical(c,gekko_crfD(w),c->GPR[gekko_ra(w)],gekko_uimm(w));
break;}
case cast<uint32_t>(11ULL):{
gekko_CPU_compareArith(c,gekko_crfD(w),c->GPR[gekko_ra(w)],cast<uint32_t>(gekko_simm(w)));
break;}
case cast<uint32_t>(12ULL):{
uint32_t a = c->GPR[gekko_ra(w)];
uint32_t imm = cast<uint32_t>(gekko_simm(w));
uint32_t r = cast<uint32_t>((a + imm));
gekko_CPU_setCA(c,gekko_carryAdd(a,imm,cast<uint32_t>(0ULL)));
c->GPR[gekko_rs(w)] = r;
break;}
case cast<uint32_t>(13ULL):{
uint32_t a = c->GPR[gekko_ra(w)];
uint32_t imm = cast<uint32_t>(gekko_simm(w));
uint32_t r = cast<uint32_t>((a + imm));
gekko_CPU_setCA(c,gekko_carryAdd(a,imm,cast<uint32_t>(0ULL)));
c->GPR[gekko_rs(w)] = r;
gekko_CPU_setCR0(c,r);
break;}
case cast<uint32_t>(14ULL):{
c->GPR[gekko_rs(w)] = cast<uint32_t>((gekko_CPU_raOrZero(c,w) + cast<uint32_t>(gekko_simm(w))));
break;}
case cast<uint32_t>(15ULL):{
c->GPR[gekko_rs(w)] = cast<uint32_t>((gekko_CPU_raOrZero(c,w) + shl<uint32_t>(gekko_uimm(w),cast<int64_t>(16ULL))));
break;}
case cast<uint32_t>(16ULL):{
gekko_CPU_execBC(c,w,pc);
break;}
case cast<uint32_t>(17ULL):{
if ((bool(c->SC) && c->SC(c))) {
return ;
}
gekko_CPU_Exception(c,cast<uint32_t>(3072ULL),c->PC,cast<uint32_t>(0ULL));
break;}
case cast<uint32_t>(18ULL):{
gekko_CPU_execB(c,w,pc);
break;}
case cast<uint32_t>(19ULL):{
gekko_CPU_exec19(c,w,pc);
break;}
case cast<uint32_t>(20ULL):{
uint32_t n = gekko_shOf(w);
uint32_t m = gekko_mask32(gekko_mbOf(w),gekko_meOf(w));
uint32_t r = cast<uint32_t>(((cast<uint32_t>((gekko_rotl32(c->GPR[gekko_rs(w)],n) & m))) | ((c->GPR[gekko_ra(w)] & ~(m)))));
c->GPR[gekko_ra(w)] = r;
gekko_CPU_rc(c,w,r);
break;}
case cast<uint32_t>(21ULL):{
uint32_t r = cast<uint32_t>((gekko_rotl32(c->GPR[gekko_rs(w)],gekko_shOf(w)) & gekko_mask32(gekko_mbOf(w),gekko_meOf(w))));
c->GPR[gekko_ra(w)] = r;
gekko_CPU_rc(c,w,r);
break;}
case cast<uint32_t>(23ULL):{
uint32_t r = cast<uint32_t>((gekko_rotl32(c->GPR[gekko_rs(w)],cast<uint32_t>((c->GPR[gekko_rb(w)] & cast<uint32_t>(31ULL)))) & gekko_mask32(gekko_mbOf(w),gekko_meOf(w))));
c->GPR[gekko_ra(w)] = r;
gekko_CPU_rc(c,w,r);
break;}
case cast<uint32_t>(24ULL):{
c->GPR[gekko_ra(w)] = cast<uint32_t>((c->GPR[gekko_rs(w)] | gekko_uimm(w)));
break;}
case cast<uint32_t>(25ULL):{
c->GPR[gekko_ra(w)] = cast<uint32_t>((c->GPR[gekko_rs(w)] | shl<uint32_t>(gekko_uimm(w),cast<int64_t>(16ULL))));
break;}
case cast<uint32_t>(26ULL):{
c->GPR[gekko_ra(w)] = cast<uint32_t>((c->GPR[gekko_rs(w)] ^ gekko_uimm(w)));
break;}
case cast<uint32_t>(27ULL):{
c->GPR[gekko_ra(w)] = cast<uint32_t>((c->GPR[gekko_rs(w)] ^ shl<uint32_t>(gekko_uimm(w),cast<int64_t>(16ULL))));
break;}
case cast<uint32_t>(28ULL):{
uint32_t r = cast<uint32_t>((c->GPR[gekko_rs(w)] & gekko_uimm(w)));
c->GPR[gekko_ra(w)] = r;
gekko_CPU_setCR0(c,r);
break;}
case cast<uint32_t>(29ULL):{
uint32_t r = cast<uint32_t>((c->GPR[gekko_rs(w)] & (shl<uint32_t>(gekko_uimm(w),cast<int64_t>(16ULL)))));
c->GPR[gekko_ra(w)] = r;
gekko_CPU_setCR0(c,r);
break;}
case cast<uint32_t>(31ULL):{
gekko_CPU_exec31(c,w,pc);
break;}
case cast<uint32_t>(32ULL):{
c->GPR[gekko_rs(w)] = gekko_CPU_read32(c,gekko_CPU_ea(c,w));
break;}
case cast<uint32_t>(33ULL):{
uint32_t ea = gekko_CPU_eaU(c,w);
c->GPR[gekko_rs(w)] = gekko_CPU_read32(c,ea);
c->GPR[gekko_ra(w)] = ea;
break;}
case cast<uint32_t>(34ULL):{
c->GPR[gekko_rs(w)] = cast<uint32_t>(gekko_CPU_read8(c,gekko_CPU_ea(c,w)));
break;}
case cast<uint32_t>(35ULL):{
uint32_t ea = gekko_CPU_eaU(c,w);
c->GPR[gekko_rs(w)] = cast<uint32_t>(gekko_CPU_read8(c,ea));
c->GPR[gekko_ra(w)] = ea;
break;}
case cast<uint32_t>(36ULL):{
gekko_CPU_write32(c,gekko_CPU_ea(c,w),c->GPR[gekko_rs(w)]);
break;}
case cast<uint32_t>(37ULL):{
uint32_t ea = gekko_CPU_eaU(c,w);
gekko_CPU_write32(c,ea,c->GPR[gekko_rs(w)]);
c->GPR[gekko_ra(w)] = ea;
break;}
case cast<uint32_t>(38ULL):{
gekko_CPU_write8(c,gekko_CPU_ea(c,w),cast<uint8_t>(c->GPR[gekko_rs(w)]));
break;}
case cast<uint32_t>(39ULL):{
uint32_t ea = gekko_CPU_eaU(c,w);
gekko_CPU_write8(c,ea,cast<uint8_t>(c->GPR[gekko_rs(w)]));
c->GPR[gekko_ra(w)] = ea;
break;}
case cast<uint32_t>(40ULL):{
c->GPR[gekko_rs(w)] = cast<uint32_t>(gekko_CPU_read16(c,gekko_CPU_ea(c,w)));
break;}
case cast<uint32_t>(41ULL):{
uint32_t ea = gekko_CPU_eaU(c,w);
c->GPR[gekko_rs(w)] = cast<uint32_t>(gekko_CPU_read16(c,ea));
c->GPR[gekko_ra(w)] = ea;
break;}
case cast<uint32_t>(42ULL):{
c->GPR[gekko_rs(w)] = cast<uint32_t>(cast<int32_t>(cast<int16_t>(gekko_CPU_read16(c,gekko_CPU_ea(c,w)))));
break;}
case cast<uint32_t>(43ULL):{
uint32_t ea = gekko_CPU_eaU(c,w);
c->GPR[gekko_rs(w)] = cast<uint32_t>(cast<int32_t>(cast<int16_t>(gekko_CPU_read16(c,ea))));
c->GPR[gekko_ra(w)] = ea;
break;}
case cast<uint32_t>(44ULL):{
gekko_CPU_write16(c,gekko_CPU_ea(c,w),cast<uint16_t>(c->GPR[gekko_rs(w)]));
break;}
case cast<uint32_t>(45ULL):{
uint32_t ea = gekko_CPU_eaU(c,w);
gekko_CPU_write16(c,ea,cast<uint16_t>(c->GPR[gekko_rs(w)]));
c->GPR[gekko_ra(w)] = ea;
break;}
case cast<uint32_t>(46ULL):{
uint32_t ea = gekko_CPU_ea(c,w);
{uint32_t i = gekko_rs(w);for (;(i < cast<uint32_t>(32ULL));i++){
c->GPR[i] = gekko_CPU_read32(c,ea);
ea += cast<uint32_t>(4ULL);
}
}break;}
case cast<uint32_t>(47ULL):{
uint32_t ea = gekko_CPU_ea(c,w);
{uint32_t i = gekko_rs(w);for (;(i < cast<uint32_t>(32ULL));i++){
gekko_CPU_write32(c,ea,c->GPR[i]);
ea += cast<uint32_t>(4ULL);
}
}break;}
case cast<uint32_t>(48ULL):{
gekko_CPU_loadFS(c,gekko_rs(w),gekko_CPU_ea(c,w));
break;}
case cast<uint32_t>(49ULL):{
uint32_t ea = gekko_CPU_eaU(c,w);
gekko_CPU_loadFS(c,gekko_rs(w),ea);
c->GPR[gekko_ra(w)] = ea;
break;}
case cast<uint32_t>(50ULL):{
c->FPR[gekko_rs(w)].PS0 = gekko_f64from(gekko_CPU_read64(c,gekko_CPU_ea(c,w)));
break;}
case cast<uint32_t>(51ULL):{
uint32_t ea = gekko_CPU_eaU(c,w);
c->FPR[gekko_rs(w)].PS0 = gekko_f64from(gekko_CPU_read64(c,ea));
c->GPR[gekko_ra(w)] = ea;
break;}
case cast<uint32_t>(52ULL):{
gekko_CPU_write32(c,gekko_CPU_ea(c,w),go_math_Float32bits(cast<float>(c->FPR[gekko_rs(w)].PS0)));
break;}
case cast<uint32_t>(53ULL):{
uint32_t ea = gekko_CPU_eaU(c,w);
gekko_CPU_write32(c,ea,go_math_Float32bits(cast<float>(c->FPR[gekko_rs(w)].PS0)));
c->GPR[gekko_ra(w)] = ea;
break;}
case cast<uint32_t>(54ULL):{
gekko_CPU_write64(c,gekko_CPU_ea(c,w),gekko_f64bits(c->FPR[gekko_rs(w)].PS0));
break;}
case cast<uint32_t>(55ULL):{
uint32_t ea = gekko_CPU_eaU(c,w);
gekko_CPU_write64(c,ea,gekko_f64bits(c->FPR[gekko_rs(w)].PS0));
c->GPR[gekko_ra(w)] = ea;
break;}
case cast<uint32_t>(56ULL):case cast<uint32_t>(57ULL):case cast<uint32_t>(60ULL):case cast<uint32_t>(61ULL):{
gekko_CPU_execPSQ(c,w);
break;}
case cast<uint32_t>(59ULL):{
gekko_CPU_exec59(c,w);
break;}
case cast<uint32_t>(63ULL):{
gekko_CPU_exec63(c,w);
break;}
default:{
gekko_CPU_Halt(c,std::string("gekko: unimplemented primary opcode %d (word 0x%08X) at 0x%08X",62),gekko_opcd(w),w,pc);
break;}
}}
}
}
// tools/cpu/gekko/exec.go:235:1
uint32_t gekko_CPU_raOrZero(gekko_CPU* c,uint32_t w){
{
if ((gekko_ra(w) == cast<uint32_t>(0ULL))) {
return cast<uint32_t>(0ULL);
}
return c->GPR[gekko_ra(w)];
}
}
// tools/cpu/gekko/exec.go:243:1
uint32_t gekko_CPU_ea(gekko_CPU* c,uint32_t w){
{
return cast<uint32_t>((gekko_CPU_raOrZero(c,w) + cast<uint32_t>(gekko_simm(w))));
}
}
// tools/cpu/gekko/exec.go:247:1
uint32_t gekko_CPU_eaU(gekko_CPU* c,uint32_t w){
{
return cast<uint32_t>((c->GPR[gekko_ra(w)] + cast<uint32_t>(gekko_simm(w))));
}
}
// tools/cpu/gekko/exec.go:250:1
uint32_t gekko_CPU_eax(gekko_CPU* c,uint32_t w){
{
return cast<uint32_t>((gekko_CPU_raOrZero(c,w) + c->GPR[gekko_rb(w)]));
}
}
// tools/cpu/gekko/exec.go:251:1
uint32_t gekko_CPU_eaxU(gekko_CPU* c,uint32_t w){
{
return cast<uint32_t>((c->GPR[gekko_ra(w)] + c->GPR[gekko_rb(w)]));
}
}
// tools/cpu/gekko/exec.go:254:1
void gekko_CPU_rc(gekko_CPU* c,uint32_t w,uint32_t result){
{
if (gekko_rcbit(w)) {
gekko_CPU_setCR0(c,result);
}
}
}
// tools/cpu/gekko/exec.go:264:1
void gekko_CPU_loadFS(gekko_CPU* c,uint32_t d,uint32_t ea){
{
double v = cast<double>(go_math_Float32frombits(gekko_CPU_read32(c,ea)));
c->FPR[d].PS0 = v;
c->FPR[d].PS1 = v;
}
}
// tools/cpu/gekko/exec.go:270:1
uint32_t gekko_rotl32(uint32_t v,uint32_t n){
{
return cast<uint32_t>((shl<uint32_t>(v,n) | (shr<uint32_t>(v,(cast<uint32_t>((cast<uint32_t>(32ULL) - n)))) & ~((shl<uint32_t>(cast<uint32_t>(4294967295ULL),n))))));
}
}
// tools/cpu/gekko/exec.go:275:1
uint32_t gekko_mask32(uint32_t mb,uint32_t me){
{
if ((mb <= me)) {
return cast<uint32_t>(((shr<uint32_t>(cast<uint32_t>(4294967295ULL),mb)) & (shl<uint32_t>(cast<uint32_t>(4294967295ULL),(cast<uint32_t>((cast<uint32_t>(31ULL) - me)))))));
}
return cast<uint32_t>(((shr<uint32_t>(cast<uint32_t>(4294967295ULL),mb)) | (shl<uint32_t>(cast<uint32_t>(4294967295ULL),(cast<uint32_t>((cast<uint32_t>(31ULL) - me)))))));
}
}
// tools/cpu/gekko/exec.go:287:1
bool gekko_carryAdd(uint32_t a,uint32_t b,uint32_t carryIn){
{
return (cast<uint64_t>((cast<uint64_t>((cast<uint64_t>(a) + cast<uint64_t>(b))) + cast<uint64_t>(carryIn))) > cast<uint64_t>(4294967295ULL));
}
}
// tools/cpu/gekko/exec.go:294:1
bool gekko_carrySub(uint32_t a,uint32_t b){
{
return gekko_carryAdd(a,cast<uint32_t>(~b),cast<uint32_t>(1ULL));
}
}
// tools/cpu/gekko/exec.go:300:1
bool gekko_overflowAdd(uint32_t a,uint32_t b,uint32_t r){
{
return (cast<uint32_t>((cast<uint32_t>(((cast<uint32_t>((a ^ r))) & (cast<uint32_t>((b ^ r))))) & cast<uint32_t>(2147483648ULL))) != cast<uint32_t>(0ULL));
}
}
// tools/cpu/gekko/exec.go:306:1
void gekko_CPU_compareArith(gekko_CPU* c,uint32_t crf,uint32_t a,uint32_t b){
{
uint32_t f = cast<uint32_t>(0ULL);
{
if ((cast<int32_t>(a) < cast<int32_t>(b))){
f = cast<uint32_t>(8ULL);
}
else if ((cast<int32_t>(a) > cast<int32_t>(b))){
f = cast<uint32_t>(4ULL);
}
else {
f = cast<uint32_t>(2ULL);
}
}
tmp36:;
if ((cast<uint32_t>((c->XER & cast<uint32_t>(2147483648ULL))) != cast<uint32_t>(0ULL))) {
f |= cast<uint32_t>(1ULL);
}
gekko_CPU_SetCRField(c,crf,f);
}
}
// tools/cpu/gekko/exec.go:322:1
void gekko_CPU_compareLogical(gekko_CPU* c,uint32_t crf,uint32_t a,uint32_t b){
{
uint32_t f = cast<uint32_t>(0ULL);
{
if ((a < b)){
f = cast<uint32_t>(8ULL);
}
else if ((a > b)){
f = cast<uint32_t>(4ULL);
}
else {
f = cast<uint32_t>(2ULL);
}
}
tmp37:;
if ((cast<uint32_t>((c->XER & cast<uint32_t>(2147483648ULL))) != cast<uint32_t>(0ULL))) {
f |= cast<uint32_t>(1ULL);
}
gekko_CPU_SetCRField(c,crf,f);
}
}
// tools/cpu/gekko/exec.go:339:1
bool gekko_CPU_trapCond(gekko_CPU* c,uint32_t to,uint32_t a,uint32_t b){
{
auto tmp38 = std::make_tuple(cast<int32_t>(a),cast<int32_t>(b));
int32_t sa = std::get<0>(tmp38);
int32_t sb = std::get<1>(tmp38);
{
if (((cast<uint32_t>((to & cast<uint32_t>(16ULL))) != cast<uint32_t>(0ULL)) && (sa < sb))){
return true;
}
else if (((cast<uint32_t>((to & cast<uint32_t>(8ULL))) != cast<uint32_t>(0ULL)) && (sa > sb))){
return true;
}
else if (((cast<uint32_t>((to & cast<uint32_t>(4ULL))) != cast<uint32_t>(0ULL)) && (a == b))){
return true;
}
else if (((cast<uint32_t>((to & cast<uint32_t>(2ULL))) != cast<uint32_t>(0ULL)) && (a < b))){
return true;
}
else if (((cast<uint32_t>((to & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL)) && (a > b))){
return true;
}
}
tmp39:;
return false;
}
}
// tools/cpu/gekko/exec.go:361:1
bool gekko_CPU_branchTaken(gekko_CPU* c,uint32_t bo,uint32_t bi){
{
bool ctrOK = true;
if ((cast<uint32_t>((bo & cast<uint32_t>(4ULL))) == cast<uint32_t>(0ULL))) {
c->CTR--;
if ((cast<uint32_t>((bo & cast<uint32_t>(2ULL))) != cast<uint32_t>(0ULL))) {
ctrOK = (c->CTR == cast<uint32_t>(0ULL));
}
else {
ctrOK = (c->CTR != cast<uint32_t>(0ULL));
}
}
bool condOK = true;
if ((cast<uint32_t>((bo & cast<uint32_t>(16ULL))) == cast<uint32_t>(0ULL))) {
uint32_t bit = cast<uint32_t>(((shr<uint32_t>(c->CR,(cast<uint32_t>((cast<uint32_t>(31ULL) - bi))))) & cast<uint32_t>(1ULL)));
uint32_t want = cast<uint32_t>(0ULL);
if ((cast<uint32_t>((bo & cast<uint32_t>(8ULL))) != cast<uint32_t>(0ULL))) {
want = cast<uint32_t>(1ULL);
}
condOK = (bit == want);
}
return (ctrOK && condOK);
}
}
// tools/cpu/gekko/exec.go:383:1
void gekko_CPU_execBC(gekko_CPU* c,uint32_t w,uint32_t pc){
{
if (gekko_lk(w)) {
c->LR = cast<uint32_t>((pc + cast<uint32_t>(4ULL)));
}
if ((!gekko_CPU_branchTaken(c,gekko_rs(w),gekko_ra(w)))) {
return ;
}
int32_t d = cast<int32_t>(cast<int16_t>(cast<uint32_t>((w & cast<uint32_t>(65532ULL)))));
if (gekko_aa(w)) {
c->PC = cast<uint32_t>(d);
}
else {
c->PC = cast<uint32_t>(cast<int32_t>((cast<int32_t>(pc) + d)));
}
}
}
// tools/cpu/gekko/exec.go:398:1
void gekko_CPU_execB(gekko_CPU* c,uint32_t w,uint32_t pc){
{
int32_t d = shr<int32_t>(shl<int32_t>(cast<int32_t>(cast<uint32_t>((w & cast<uint32_t>(67108860ULL)))),cast<int64_t>(6ULL)),cast<int64_t>(6ULL));
if (gekko_lk(w)) {
c->LR = cast<uint32_t>((pc + cast<uint32_t>(4ULL)));
}
if (gekko_aa(w)) {
c->PC = cast<uint32_t>(d);
}
else {
c->PC = cast<uint32_t>(cast<int32_t>((cast<int32_t>(pc) + d)));
}
}
}
// tools/cpu/gekko/exec.go:410:1
void gekko_CPU_exec19(gekko_CPU* c,uint32_t w,uint32_t pc){
{
{
switch(gekko_xo10(w)){
case cast<uint32_t>(0ULL):{
gekko_CPU_SetCRField(c,gekko_crfD(w),gekko_CPU_CRField(c,gekko_crfS(w)));
break;}
case cast<uint32_t>(16ULL):{
uint32_t target = (c->LR & ~(cast<uint32_t>(3ULL)));
bool taken = gekko_CPU_branchTaken(c,gekko_rs(w),gekko_ra(w));
if (gekko_lk(w)) {
c->LR = cast<uint32_t>((pc + cast<uint32_t>(4ULL)));
}
if (taken) {
c->PC = target;
}
break;}
case cast<uint32_t>(528ULL):{
uint32_t target = (c->CTR & ~(cast<uint32_t>(3ULL)));
uint32_t bo = gekko_rs(w);
bool condOK = true;
if ((cast<uint32_t>((bo & cast<uint32_t>(16ULL))) == cast<uint32_t>(0ULL))) {
uint32_t bit = cast<uint32_t>(((shr<uint32_t>(c->CR,(cast<uint32_t>((cast<uint32_t>(31ULL) - gekko_ra(w)))))) & cast<uint32_t>(1ULL)));
uint32_t want = cast<uint32_t>(0ULL);
if ((cast<uint32_t>((bo & cast<uint32_t>(8ULL))) != cast<uint32_t>(0ULL))) {
want = cast<uint32_t>(1ULL);
}
condOK = (bit == want);
}
if (gekko_lk(w)) {
c->LR = cast<uint32_t>((pc + cast<uint32_t>(4ULL)));
}
if (condOK) {
c->PC = target;
}
break;}
case cast<uint32_t>(50ULL):{
c->MSR = cast<uint32_t>((((c->MSR & ~(cast<uint32_t>(2277572467ULL)))) | (cast<uint32_t>((c->SRR1 & cast<uint32_t>(2277572467ULL))))));
c->MSR &= ~(cast<uint32_t>(262144ULL));
c->PC = (c->SRR0 & ~(cast<uint32_t>(3ULL)));
c->LC.Enabled_ = (cast<uint32_t>((c->HID2 & cast<uint32_t>(268435456ULL))) != cast<uint32_t>(0ULL));
break;}
case cast<uint32_t>(150ULL):{
break;}
case cast<uint32_t>(33ULL):case cast<uint32_t>(129ULL):case cast<uint32_t>(193ULL):case cast<uint32_t>(225ULL):case cast<uint32_t>(257ULL):case cast<uint32_t>(289ULL):case cast<uint32_t>(417ULL):case cast<uint32_t>(449ULL):{
gekko_CPU_execCRLogic(c,w);
break;}
default:{
gekko_CPU_Halt(c,std::string("gekko: unimplemented opcode 19 extended %d (word 0x%08X) at 0x%08X",66),gekko_xo10(w),w,pc);
break;}
}}
}
}
// tools/cpu/gekko/exec.go:457:1
void gekko_CPU_execCRLogic(gekko_CPU* c,uint32_t w){
{
auto tmp40 = std::make_tuple(gekko_rs(w),gekko_ra(w),gekko_rb(w));
uint32_t d = std::get<0>(tmp40);
uint32_t a = std::get<1>(tmp40);
uint32_t b = std::get<2>(tmp40);
uint32_t ba = cast<uint32_t>(((shr<uint32_t>(c->CR,(cast<uint32_t>((cast<uint32_t>(31ULL) - a))))) & cast<uint32_t>(1ULL)));
uint32_t bb = cast<uint32_t>(((shr<uint32_t>(c->CR,(cast<uint32_t>((cast<uint32_t>(31ULL) - b))))) & cast<uint32_t>(1ULL)));
uint32_t v={};
{
switch(gekko_xo10(w)){
case cast<uint32_t>(33ULL):{
v = cast<uint32_t>((cast<uint32_t>(~(cast<uint32_t>((ba | bb)))) & cast<uint32_t>(1ULL)));
break;}
case cast<uint32_t>(129ULL):{
v = cast<uint32_t>(((ba & ~(bb)) & cast<uint32_t>(1ULL)));
break;}
case cast<uint32_t>(193ULL):{
v = cast<uint32_t>((ba ^ bb));
break;}
case cast<uint32_t>(225ULL):{
v = cast<uint32_t>((cast<uint32_t>(~(cast<uint32_t>((ba & bb)))) & cast<uint32_t>(1ULL)));
break;}
case cast<uint32_t>(257ULL):{
v = cast<uint32_t>((ba & bb));
break;}
case cast<uint32_t>(289ULL):{
v = cast<uint32_t>((cast<uint32_t>(~(cast<uint32_t>((ba ^ bb)))) & cast<uint32_t>(1ULL)));
break;}
case cast<uint32_t>(417ULL):{
v = cast<uint32_t>(((cast<uint32_t>((ba | (cast<uint32_t>((cast<uint32_t>(~bb) & cast<uint32_t>(1ULL))))))) & cast<uint32_t>(1ULL)));
break;}
case cast<uint32_t>(449ULL):{
v = cast<uint32_t>((ba | bb));
break;}
}}
uint32_t sh = cast<uint32_t>((cast<uint32_t>(31ULL) - d));
c->CR = cast<uint32_t>((((c->CR & ~((shl<uint32_t>(cast<uint32_t>(1ULL),sh))))) | (shl<uint32_t>(v,sh))));
}
}
// tools/cpu/gekko/exec31.go:21:1
void gekko_CPU_exec31(gekko_CPU* c,uint32_t w,uint32_t pc){
{
auto tmp41 = std::make_tuple(gekko_rs(w),gekko_ra(w),gekko_rb(w));
uint32_t d = std::get<0>(tmp41);
uint32_t a = std::get<1>(tmp41);
uint32_t b = std::get<2>(tmp41);
uint32_t x = gekko_xo10(w);
if (((x >= cast<uint32_t>(512ULL)) && get(gekko_oeForm,cast<uint32_t>((x - cast<uint32_t>(512ULL)))))) {
x -= cast<uint32_t>(512ULL);
}
{
switch(x){
case cast<uint32_t>(0ULL):{
gekko_CPU_compareArith(c,gekko_crfD(w),c->GPR[a],c->GPR[b]);
break;}
case cast<uint32_t>(32ULL):{
gekko_CPU_compareLogical(c,gekko_crfD(w),c->GPR[a],c->GPR[b]);
break;}
case cast<uint32_t>(4ULL):{
if (gekko_CPU_trapCond(c,d,c->GPR[a],c->GPR[b])) {
gekko_CPU_programException(c,cast<uint32_t>(131072ULL));
}
break;}
case cast<uint32_t>(266ULL):{
auto tmp42 = std::make_tuple(c->GPR[a],c->GPR[b]);
uint32_t va = std::get<0>(tmp42);
uint32_t vb = std::get<1>(tmp42);
uint32_t r = cast<uint32_t>((va + vb));
c->GPR[d] = r;
gekko_CPU_oeRc(c,w,r,gekko_overflowAdd(va,vb,r));
break;}
case cast<uint32_t>(10ULL):{
auto tmp43 = std::make_tuple(c->GPR[a],c->GPR[b]);
uint32_t va = std::get<0>(tmp43);
uint32_t vb = std::get<1>(tmp43);
uint32_t r = cast<uint32_t>((va + vb));
gekko_CPU_setCA(c,gekko_carryAdd(va,vb,cast<uint32_t>(0ULL)));
c->GPR[d] = r;
gekko_CPU_oeRc(c,w,r,gekko_overflowAdd(va,vb,r));
break;}
case cast<uint32_t>(138ULL):{
auto tmp44 = std::make_tuple(c->GPR[a],c->GPR[b],gekko_CPU_ca(c));
uint32_t va = std::get<0>(tmp44);
uint32_t vb = std::get<1>(tmp44);
uint32_t ci = std::get<2>(tmp44);
uint32_t r = cast<uint32_t>((cast<uint32_t>((va + vb)) + ci));
gekko_CPU_setCA(c,gekko_carryAdd(va,vb,ci));
c->GPR[d] = r;
gekko_CPU_oeRc(c,w,r,gekko_overflowAdd(va,vb,r));
break;}
case cast<uint32_t>(234ULL):{
auto tmp45 = std::make_tuple(c->GPR[a],gekko_CPU_ca(c));
uint32_t va = std::get<0>(tmp45);
uint32_t ci = std::get<1>(tmp45);
uint32_t r = cast<uint32_t>((cast<uint32_t>((va + cast<uint32_t>(4294967295ULL))) + ci));
gekko_CPU_setCA(c,gekko_carryAdd(va,cast<uint32_t>(4294967295ULL),ci));
c->GPR[d] = r;
gekko_CPU_oeRc(c,w,r,gekko_overflowAdd(va,cast<uint32_t>(4294967295ULL),r));
break;}
case cast<uint32_t>(202ULL):{
auto tmp46 = std::make_tuple(c->GPR[a],gekko_CPU_ca(c));
uint32_t va = std::get<0>(tmp46);
uint32_t ci = std::get<1>(tmp46);
uint32_t r = cast<uint32_t>((va + ci));
gekko_CPU_setCA(c,gekko_carryAdd(va,cast<uint32_t>(0ULL),ci));
c->GPR[d] = r;
gekko_CPU_oeRc(c,w,r,gekko_overflowAdd(va,cast<uint32_t>(0ULL),r));
break;}
case cast<uint32_t>(40ULL):{
auto tmp47 = std::make_tuple(c->GPR[a],c->GPR[b]);
uint32_t va = std::get<0>(tmp47);
uint32_t vb = std::get<1>(tmp47);
uint32_t r = cast<uint32_t>((vb - va));
c->GPR[d] = r;
gekko_CPU_oeRc(c,w,r,gekko_overflowAdd(cast<uint32_t>(~va),vb,r));
break;}
case cast<uint32_t>(8ULL):{
auto tmp48 = std::make_tuple(c->GPR[a],c->GPR[b]);
uint32_t va = std::get<0>(tmp48);
uint32_t vb = std::get<1>(tmp48);
uint32_t r = cast<uint32_t>((vb - va));
gekko_CPU_setCA(c,gekko_carrySub(vb,va));
c->GPR[d] = r;
gekko_CPU_oeRc(c,w,r,gekko_overflowAdd(cast<uint32_t>(~va),vb,r));
break;}
case cast<uint32_t>(136ULL):{
auto tmp49 = std::make_tuple(c->GPR[a],c->GPR[b],gekko_CPU_ca(c));
uint32_t va = std::get<0>(tmp49);
uint32_t vb = std::get<1>(tmp49);
uint32_t ci = std::get<2>(tmp49);
uint32_t r = cast<uint32_t>((cast<uint32_t>((vb + cast<uint32_t>(~va))) + ci));
gekko_CPU_setCA(c,gekko_carryAdd(vb,cast<uint32_t>(~va),ci));
c->GPR[d] = r;
gekko_CPU_oeRc(c,w,r,gekko_overflowAdd(cast<uint32_t>(~va),vb,r));
break;}
case cast<uint32_t>(232ULL):{
auto tmp50 = std::make_tuple(c->GPR[a],gekko_CPU_ca(c));
uint32_t va = std::get<0>(tmp50);
uint32_t ci = std::get<1>(tmp50);
uint32_t r = cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(~va) + cast<uint32_t>(4294967295ULL))) + ci));
gekko_CPU_setCA(c,gekko_carryAdd(cast<uint32_t>(~va),cast<uint32_t>(4294967295ULL),ci));
c->GPR[d] = r;
gekko_CPU_oeRc(c,w,r,gekko_overflowAdd(cast<uint32_t>(~va),cast<uint32_t>(4294967295ULL),r));
break;}
case cast<uint32_t>(200ULL):{
auto tmp51 = std::make_tuple(c->GPR[a],gekko_CPU_ca(c));
uint32_t va = std::get<0>(tmp51);
uint32_t ci = std::get<1>(tmp51);
uint32_t r = cast<uint32_t>((cast<uint32_t>(~va) + ci));
gekko_CPU_setCA(c,gekko_carryAdd(cast<uint32_t>(~va),cast<uint32_t>(0ULL),ci));
c->GPR[d] = r;
gekko_CPU_oeRc(c,w,r,gekko_overflowAdd(cast<uint32_t>(~va),cast<uint32_t>(0ULL),r));
break;}
case cast<uint32_t>(104ULL):{
uint32_t va = c->GPR[a];
uint32_t r = cast<uint32_t>(-va);
c->GPR[d] = r;
gekko_CPU_oeRc(c,w,r,(va == cast<uint32_t>(2147483648ULL)));
break;}
case cast<uint32_t>(235ULL):{
auto tmp52 = std::make_tuple(cast<int32_t>(c->GPR[a]),cast<int32_t>(c->GPR[b]));
int32_t va = std::get<0>(tmp52);
int32_t vb = std::get<1>(tmp52);
int64_t full = cast<int64_t>((cast<int64_t>(va) * cast<int64_t>(vb)));
uint32_t r = cast<uint32_t>(full);
c->GPR[d] = r;
gekko_CPU_oeRc(c,w,r,(full != cast<int64_t>(cast<int32_t>(full))));
break;}
case cast<uint32_t>(75ULL):{
int64_t full = cast<int64_t>((cast<int64_t>(cast<int32_t>(c->GPR[a])) * cast<int64_t>(cast<int32_t>(c->GPR[b]))));
uint32_t r = cast<uint32_t>(shr<int64_t>(full,cast<int64_t>(32ULL)));
c->GPR[d] = r;
gekko_CPU_rc(c,w,r);
break;}
case cast<uint32_t>(11ULL):{
uint64_t full = cast<uint64_t>((cast<uint64_t>(c->GPR[a]) * cast<uint64_t>(c->GPR[b])));
uint32_t r = cast<uint32_t>(shr<uint64_t>(full,cast<int64_t>(32ULL)));
c->GPR[d] = r;
gekko_CPU_rc(c,w,r);
break;}
case cast<uint32_t>(491ULL):{
auto tmp53 = std::make_tuple(cast<int32_t>(c->GPR[a]),cast<int32_t>(c->GPR[b]));
int32_t va = std::get<0>(tmp53);
int32_t vb = std::get<1>(tmp53);
if (((vb == cast<int32_t>(0ULL)) || (((va == cast<int32_t>(-2147483648ULL)) && (vb == cast<int32_t>(-1ULL)))))) {
c->GPR[d] = cast<uint32_t>(0ULL);
gekko_CPU_oeRc(c,w,cast<uint32_t>(0ULL),true);
return ;
}
uint32_t r = cast<uint32_t>(divi<int32_t>(va,vb));
c->GPR[d] = r;
gekko_CPU_oeRc(c,w,r,false);
break;}
case cast<uint32_t>(459ULL):{
auto tmp54 = std::make_tuple(c->GPR[a],c->GPR[b]);
uint32_t va = std::get<0>(tmp54);
uint32_t vb = std::get<1>(tmp54);
if ((vb == cast<uint32_t>(0ULL))) {
c->GPR[d] = cast<uint32_t>(0ULL);
gekko_CPU_oeRc(c,w,cast<uint32_t>(0ULL),true);
return ;
}
uint32_t r = divi<uint32_t>(va,vb);
c->GPR[d] = r;
gekko_CPU_oeRc(c,w,r,false);
break;}
case cast<uint32_t>(28ULL):{
uint32_t r = cast<uint32_t>((c->GPR[d] & c->GPR[b]));
c->GPR[a] = r;
gekko_CPU_rc(c,w,r);
break;}
case cast<uint32_t>(60ULL):{
uint32_t r = (c->GPR[d] & ~(c->GPR[b]));
c->GPR[a] = r;
gekko_CPU_rc(c,w,r);
break;}
case cast<uint32_t>(444ULL):{
uint32_t r = cast<uint32_t>((c->GPR[d] | c->GPR[b]));
c->GPR[a] = r;
gekko_CPU_rc(c,w,r);
break;}
case cast<uint32_t>(412ULL):{
uint32_t r = cast<uint32_t>((c->GPR[d] | cast<uint32_t>(~c->GPR[b])));
c->GPR[a] = r;
gekko_CPU_rc(c,w,r);
break;}
case cast<uint32_t>(316ULL):{
uint32_t r = cast<uint32_t>((c->GPR[d] ^ c->GPR[b]));
c->GPR[a] = r;
gekko_CPU_rc(c,w,r);
break;}
case cast<uint32_t>(124ULL):{
uint32_t r = cast<uint32_t>(~(cast<uint32_t>((c->GPR[d] | c->GPR[b]))));
c->GPR[a] = r;
gekko_CPU_rc(c,w,r);
break;}
case cast<uint32_t>(476ULL):{
uint32_t r = cast<uint32_t>(~(cast<uint32_t>((c->GPR[d] & c->GPR[b]))));
c->GPR[a] = r;
gekko_CPU_rc(c,w,r);
break;}
case cast<uint32_t>(284ULL):{
uint32_t r = cast<uint32_t>(~(cast<uint32_t>((c->GPR[d] ^ c->GPR[b]))));
c->GPR[a] = r;
gekko_CPU_rc(c,w,r);
break;}
case cast<uint32_t>(26ULL):{
uint32_t r = cast<uint32_t>(go_bits_LeadingZeros32(c->GPR[d]));
c->GPR[a] = r;
gekko_CPU_rc(c,w,r);
break;}
case cast<uint32_t>(922ULL):{
uint32_t r = cast<uint32_t>(cast<int32_t>(cast<int16_t>(c->GPR[d])));
c->GPR[a] = r;
gekko_CPU_rc(c,w,r);
break;}
case cast<uint32_t>(954ULL):{
uint32_t r = cast<uint32_t>(cast<int32_t>(cast<int8_t>(c->GPR[d])));
c->GPR[a] = r;
gekko_CPU_rc(c,w,r);
break;}
case cast<uint32_t>(24ULL):{
uint32_t sh = cast<uint32_t>((c->GPR[b] & cast<uint32_t>(63ULL)));
uint32_t r = cast<uint32_t>(0ULL);
if ((sh < cast<uint32_t>(32ULL))) {
r = shl<uint32_t>(c->GPR[d],sh);
}
c->GPR[a] = r;
gekko_CPU_rc(c,w,r);
break;}
case cast<uint32_t>(536ULL):{
uint32_t sh = cast<uint32_t>((c->GPR[b] & cast<uint32_t>(63ULL)));
uint32_t r = cast<uint32_t>(0ULL);
if ((sh < cast<uint32_t>(32ULL))) {
r = shr<uint32_t>(c->GPR[d],sh);
}
c->GPR[a] = r;
gekko_CPU_rc(c,w,r);
break;}
case cast<uint32_t>(792ULL):{
gekko_CPU_srawi(c,w,cast<uint32_t>((c->GPR[b] & cast<uint32_t>(63ULL))));
break;}
case cast<uint32_t>(824ULL):{
gekko_CPU_srawi(c,w,gekko_shOf(w));
break;}
case cast<uint32_t>(19ULL):{
c->GPR[d] = c->CR;
break;}
case cast<uint32_t>(144ULL):{
uint32_t mask = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(12ULL))) & cast<uint32_t>(255ULL)));
uint32_t m = cast<uint32_t>(0ULL);
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(8ULL));i++){
if ((cast<uint32_t>((mask & (shr<uint32_t>(cast<uint32_t>(128ULL),i)))) != cast<uint32_t>(0ULL))) {
m |= shl<uint32_t>(cast<uint32_t>(15ULL),(cast<int64_t>((cast<int64_t>(28ULL) - cast<int64_t>((cast<int64_t>(4ULL) * i))))));
}
}
}c->CR = cast<uint32_t>((((c->CR & ~(m))) | (cast<uint32_t>((c->GPR[d] & m)))));
break;}
case cast<uint32_t>(512ULL):{
gekko_CPU_SetCRField(c,gekko_crfD(w),shr<uint32_t>(c->XER,cast<int64_t>(28ULL)));
c->XER &= ~(cast<uint32_t>(3758096384ULL));
break;}
case cast<uint32_t>(339ULL):{
c->GPR[d] = gekko_CPU_readSPR(c,gekko_sprOf(w),pc);
break;}
case cast<uint32_t>(467ULL):{
gekko_CPU_writeSPR(c,gekko_sprOf(w),c->GPR[d],pc);
break;}
case cast<uint32_t>(371ULL):{
{
switch(gekko_sprOf(w)){
case cast<uint32_t>(268ULL):{
c->GPR[d] = cast<uint32_t>(c->TB);
break;}
case cast<uint32_t>(269ULL):{
c->GPR[d] = cast<uint32_t>(shr<uint64_t>(c->TB,cast<int64_t>(32ULL)));
break;}
default:{
gekko_CPU_Halt(c,std::string("gekko: mftb of SPR %d at 0x%08X",31),gekko_sprOf(w),pc);
break;}
}}
break;}
case cast<uint32_t>(83ULL):{
c->GPR[d] = c->MSR;
break;}
case cast<uint32_t>(146ULL):{
c->MSR = c->GPR[d];
break;}
case cast<uint32_t>(210ULL):{
c->SR[cast<uint32_t>((a & cast<uint32_t>(15ULL)))] = c->GPR[d];
break;}
case cast<uint32_t>(242ULL):{
c->SR[cast<uint32_t>(((shr<uint32_t>(c->GPR[b],cast<int64_t>(28ULL))) & cast<uint32_t>(15ULL)))] = c->GPR[d];
break;}
case cast<uint32_t>(595ULL):{
c->GPR[d] = c->SR[cast<uint32_t>((a & cast<uint32_t>(15ULL)))];
break;}
case cast<uint32_t>(659ULL):{
c->GPR[d] = c->SR[cast<uint32_t>(((shr<uint32_t>(c->GPR[b],cast<int64_t>(28ULL))) & cast<uint32_t>(15ULL)))];
break;}
case cast<uint32_t>(23ULL):{
c->GPR[d] = gekko_CPU_read32(c,gekko_CPU_eax(c,w));
break;}
case cast<uint32_t>(55ULL):{
uint32_t ea = gekko_CPU_eaxU(c,w);
c->GPR[d] = gekko_CPU_read32(c,ea);
c->GPR[a] = ea;
break;}
case cast<uint32_t>(87ULL):{
c->GPR[d] = cast<uint32_t>(gekko_CPU_read8(c,gekko_CPU_eax(c,w)));
break;}
case cast<uint32_t>(119ULL):{
uint32_t ea = gekko_CPU_eaxU(c,w);
c->GPR[d] = cast<uint32_t>(gekko_CPU_read8(c,ea));
c->GPR[a] = ea;
break;}
case cast<uint32_t>(279ULL):{
c->GPR[d] = cast<uint32_t>(gekko_CPU_read16(c,gekko_CPU_eax(c,w)));
break;}
case cast<uint32_t>(311ULL):{
uint32_t ea = gekko_CPU_eaxU(c,w);
c->GPR[d] = cast<uint32_t>(gekko_CPU_read16(c,ea));
c->GPR[a] = ea;
break;}
case cast<uint32_t>(343ULL):{
c->GPR[d] = cast<uint32_t>(cast<int32_t>(cast<int16_t>(gekko_CPU_read16(c,gekko_CPU_eax(c,w)))));
break;}
case cast<uint32_t>(375ULL):{
uint32_t ea = gekko_CPU_eaxU(c,w);
c->GPR[d] = cast<uint32_t>(cast<int32_t>(cast<int16_t>(gekko_CPU_read16(c,ea))));
c->GPR[a] = ea;
break;}
case cast<uint32_t>(151ULL):{
gekko_CPU_write32(c,gekko_CPU_eax(c,w),c->GPR[d]);
break;}
case cast<uint32_t>(183ULL):{
uint32_t ea = gekko_CPU_eaxU(c,w);
gekko_CPU_write32(c,ea,c->GPR[d]);
c->GPR[a] = ea;
break;}
case cast<uint32_t>(215ULL):{
gekko_CPU_write8(c,gekko_CPU_eax(c,w),cast<uint8_t>(c->GPR[d]));
break;}
case cast<uint32_t>(247ULL):{
uint32_t ea = gekko_CPU_eaxU(c,w);
gekko_CPU_write8(c,ea,cast<uint8_t>(c->GPR[d]));
c->GPR[a] = ea;
break;}
case cast<uint32_t>(407ULL):{
gekko_CPU_write16(c,gekko_CPU_eax(c,w),cast<uint16_t>(c->GPR[d]));
break;}
case cast<uint32_t>(439ULL):{
uint32_t ea = gekko_CPU_eaxU(c,w);
gekko_CPU_write16(c,ea,cast<uint16_t>(c->GPR[d]));
c->GPR[a] = ea;
break;}
case cast<uint32_t>(534ULL):{
uint32_t v = gekko_CPU_read32(c,gekko_CPU_eax(c,w));
c->GPR[d] = go_bits_ReverseBytes32(v);
break;}
case cast<uint32_t>(790ULL):{
uint16_t v = gekko_CPU_read16(c,gekko_CPU_eax(c,w));
c->GPR[d] = cast<uint32_t>(go_bits_ReverseBytes16(v));
break;}
case cast<uint32_t>(662ULL):{
gekko_CPU_write32(c,gekko_CPU_eax(c,w),go_bits_ReverseBytes32(c->GPR[d]));
break;}
case cast<uint32_t>(918ULL):{
gekko_CPU_write16(c,gekko_CPU_eax(c,w),go_bits_ReverseBytes16(cast<uint16_t>(c->GPR[d])));
break;}
case cast<uint32_t>(597ULL):{
uint32_t n = gekko_rb(w);
if ((n == cast<uint32_t>(0ULL))) {
n = cast<uint32_t>(32ULL);
}
gekko_CPU_loadString(c,d,gekko_CPU_raOrZero(c,w),n);
break;}
case cast<uint32_t>(725ULL):{
uint32_t n = gekko_rb(w);
if ((n == cast<uint32_t>(0ULL))) {
n = cast<uint32_t>(32ULL);
}
gekko_CPU_storeString(c,d,gekko_CPU_raOrZero(c,w),n);
break;}
case cast<uint32_t>(533ULL):{
gekko_CPU_loadString(c,d,gekko_CPU_eax(c,w),cast<uint32_t>((c->XER & cast<uint32_t>(127ULL))));
break;}
case cast<uint32_t>(661ULL):{
gekko_CPU_storeString(c,d,gekko_CPU_eax(c,w),cast<uint32_t>((c->XER & cast<uint32_t>(127ULL))));
break;}
case cast<uint32_t>(20ULL):{
uint32_t ea = gekko_CPU_eax(c,w);
c->GPR[d] = gekko_CPU_read32(c,ea);
c->Reserved = true;
c->ReserveAddr = ea;
break;}
case cast<uint32_t>(150ULL):{
uint32_t ea = gekko_CPU_eax(c,w);
uint32_t f = cast<uint32_t>(0ULL);
if ((cast<uint32_t>((c->XER & cast<uint32_t>(2147483648ULL))) != cast<uint32_t>(0ULL))) {
f |= cast<uint32_t>(1ULL);
}
if ((c->Reserved && ((c->ReserveAddr & ~(cast<uint32_t>(31ULL))) == (ea & ~(cast<uint32_t>(31ULL)))))) {
gekko_CPU_write32(c,ea,c->GPR[d]);
f |= cast<uint32_t>(2ULL);
}
c->Reserved = false;
gekko_CPU_SetCRField(c,cast<uint32_t>(0ULL),f);
break;}
case cast<uint32_t>(535ULL):{
gekko_CPU_loadFS(c,d,gekko_CPU_eax(c,w));
break;}
case cast<uint32_t>(567ULL):{
uint32_t ea = gekko_CPU_eaxU(c,w);
gekko_CPU_loadFS(c,d,ea);
c->GPR[a] = ea;
break;}
case cast<uint32_t>(599ULL):{
c->FPR[d].PS0 = gekko_f64from(gekko_CPU_read64(c,gekko_CPU_eax(c,w)));
break;}
case cast<uint32_t>(631ULL):{
uint32_t ea = gekko_CPU_eaxU(c,w);
c->FPR[d].PS0 = gekko_f64from(gekko_CPU_read64(c,ea));
c->GPR[a] = ea;
break;}
case cast<uint32_t>(663ULL):{
gekko_CPU_write32(c,gekko_CPU_eax(c,w),gekko_float32bitsOf(c->FPR[d].PS0));
break;}
case cast<uint32_t>(695ULL):{
uint32_t ea = gekko_CPU_eaxU(c,w);
gekko_CPU_write32(c,ea,gekko_float32bitsOf(c->FPR[d].PS0));
c->GPR[a] = ea;
break;}
case cast<uint32_t>(727ULL):{
gekko_CPU_write64(c,gekko_CPU_eax(c,w),gekko_bits64(c->FPR[d].PS0));
break;}
case cast<uint32_t>(759ULL):{
uint32_t ea = gekko_CPU_eaxU(c,w);
gekko_CPU_write64(c,ea,gekko_bits64(c->FPR[d].PS0));
c->GPR[a] = ea;
break;}
case cast<uint32_t>(983ULL):{
gekko_CPU_write32(c,gekko_CPU_eax(c,w),cast<uint32_t>(gekko_bits64(c->FPR[d].PS0)));
break;}
case cast<uint32_t>(1014ULL):{
gekko_CPU_dcbz(c,gekko_CPU_eax(c,w));
break;}
case cast<uint32_t>(54ULL):case cast<uint32_t>(86ULL):case cast<uint32_t>(246ULL):case cast<uint32_t>(278ULL):case cast<uint32_t>(470ULL):case cast<uint32_t>(982ULL):{
break;}
case cast<uint32_t>(598ULL):case cast<uint32_t>(854ULL):case cast<uint32_t>(566ULL):{
break;}
case cast<uint32_t>(306ULL):case cast<uint32_t>(370ULL):{
break;}
case cast<uint32_t>(310ULL):case cast<uint32_t>(438ULL):{
gekko_CPU_Halt(c,std::string("gekko: external control (eciwx/ecowx) at 0x%08X; this machine has no such device",80),pc);
break;}
default:{
gekko_CPU_Halt(c,std::string("gekko: unimplemented opcode 31 extended %d (word 0x%08X) at 0x%08X",66),gekko_xo10(w),w,pc);
break;}
}}
}
}
// tools/cpu/gekko/exec31.go:409:1
void gekko_CPU_srawi(gekko_CPU* c,uint32_t w,uint32_t sh){
{
int32_t v = cast<int32_t>(c->GPR[gekko_rs(w)]);
int32_t r={};
if ((sh >= cast<uint32_t>(32ULL))) {
r = shr<int32_t>(v,cast<int64_t>(31ULL));
gekko_CPU_setCA(c,(v < cast<int32_t>(0ULL)));
}
else {
r = shr<int32_t>(v,sh);
uint32_t lost = cast<uint32_t>((cast<uint32_t>(v) & (cast<uint32_t>(((shl<uint32_t>(cast<uint32_t>(1ULL),sh)) - cast<uint32_t>(1ULL))))));
gekko_CPU_setCA(c,((v < cast<int32_t>(0ULL)) && (lost != cast<uint32_t>(0ULL))));
}
c->GPR[gekko_ra(w)] = cast<uint32_t>(r);
gekko_CPU_rc(c,w,cast<uint32_t>(r));
}
}
// tools/cpu/gekko/exec31.go:429:1
void gekko_CPU_oeRc(gekko_CPU* c,uint32_t w,uint32_t result,bool overflow){
{
if (gekko_oe(w)) {
gekko_CPU_setOV(c,overflow);
}
gekko_CPU_rc(c,w,result);
}
}
// tools/cpu/gekko/exec31.go:438:1
void gekko_CPU_loadString(gekko_CPU* c,uint32_t d,uint32_t ea,uint32_t n){
{
uint32_t reg = d;
c->GPR[reg] = cast<uint32_t>(0ULL);
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < n);i++){
uint32_t shift = cast<uint32_t>((cast<uint32_t>(24ULL) - cast<uint32_t>((cast<uint32_t>(8ULL) * (modi<uint32_t>(i,cast<uint32_t>(4ULL)))))));
c->GPR[reg] |= shl<uint32_t>(cast<uint32_t>(gekko_CPU_read8(c,cast<uint32_t>((ea + i)))),shift);
if ((modi<uint32_t>(i,cast<uint32_t>(4ULL)) == cast<uint32_t>(3ULL))) {
reg = cast<uint32_t>(((cast<uint32_t>((reg + cast<uint32_t>(1ULL)))) & cast<uint32_t>(31ULL)));
c->GPR[reg] = cast<uint32_t>(0ULL);
}
}
}}
}
// tools/cpu/gekko/exec31.go:451:1
void gekko_CPU_storeString(gekko_CPU* c,uint32_t d,uint32_t ea,uint32_t n){
{
uint32_t reg = d;
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < n);i++){
uint32_t shift = cast<uint32_t>((cast<uint32_t>(24ULL) - cast<uint32_t>((cast<uint32_t>(8ULL) * (modi<uint32_t>(i,cast<uint32_t>(4ULL)))))));
gekko_CPU_write8(c,cast<uint32_t>((ea + i)),cast<uint8_t>(shr<uint32_t>(c->GPR[reg],shift)));
if ((modi<uint32_t>(i,cast<uint32_t>(4ULL)) == cast<uint32_t>(3ULL))) {
reg = cast<uint32_t>(((cast<uint32_t>((reg + cast<uint32_t>(1ULL)))) & cast<uint32_t>(31ULL)));
}
}
}}
}
// tools/cpu/gekko/fpu.go:69:1
void gekko_CPU_setFPRF(gekko_CPU* c,double v){
{
uint32_t class_={};
{
if (go_math_IsNaN(v)){
class_ = cast<uint32_t>(17ULL);
}
else if (go_math_IsInf(v,cast<int64_t>(1ULL))){
class_ = cast<uint32_t>(5ULL);
}
else if (go_math_IsInf(v,cast<int64_t>(-1ULL))){
class_ = cast<uint32_t>(9ULL);
}
else if ((v == cast<double>(0ULL))){
if (go_math_Signbit(v)) {
class_ = cast<uint32_t>(18ULL);
}
else {
class_ = cast<uint32_t>(2ULL);
}
}
else if ((go_math_Abs(v) < 2.2250738585072014e-308)){
if ((v < cast<double>(0ULL))) {
class_ = cast<uint32_t>(24ULL);
}
else {
class_ = cast<uint32_t>(20ULL);
}
}
else if ((v < cast<double>(0ULL))){
class_ = cast<uint32_t>(8ULL);
}
else {
class_ = cast<uint32_t>(4ULL);
}
}
tmp55:;
c->FPSCR = cast<uint32_t>((((c->FPSCR & ~(cast<uint32_t>(126976ULL)))) | (shl<uint32_t>(class_,cast<int64_t>(12ULL)))));
}
}
// tools/cpu/gekko/fpu.go:101:1
void gekko_CPU_setFPSCRBit(gekko_CPU* c,uint32_t bit){
{
if ((cast<uint32_t>((c->FPSCR & bit)) == cast<uint32_t>(0ULL))) {
c->FPSCR |= cast<uint32_t>(2147483648ULL);
}
c->FPSCR |= bit;
if ((cast<uint32_t>((bit & cast<uint32_t>(33031936ULL))) != cast<uint32_t>(0ULL))) {
c->FPSCR |= cast<uint32_t>(536870912ULL);
}
}
}
// tools/cpu/gekko/fpu.go:112:1
bool gekko_isSNaN(double v){
{
uint64_t b = go_math_Float64bits(v);
return (go_math_IsNaN(v) && (cast<uint64_t>((b & cast<uint64_t>(2251799813685248ULL))) == cast<uint64_t>(0ULL)));
}
}
// tools/cpu/gekko/fpu.go:121:1
void gekko_CPU_fpResult(gekko_CPU* c,uint32_t w,uint32_t d,double v){
{
c->FPR[d].PS0 = v;
gekko_CPU_setFPRF(c,v);
if (gekko_rcbit(w)) {
gekko_CPU_setCR1(c);
}
}
}
// tools/cpu/gekko/fpu.go:132:1
void gekko_CPU_fpResultS(gekko_CPU* c,uint32_t w,uint32_t d,double v){
{
double r = gekko_f32(v);
c->FPR[d].PS0 = r;
c->FPR[d].PS1 = r;
gekko_CPU_setFPRF(c,r);
if (gekko_rcbit(w)) {
gekko_CPU_setCR1(c);
}
}
}
// tools/cpu/gekko/fpu.go:146:1
double gekko_CPU_fadd(gekko_CPU* c,double a,double b){
{
if ((gekko_isSNaN(a) || gekko_isSNaN(b))) {
gekko_CPU_setFPSCRBit(c,cast<uint32_t>(16777216ULL));
}
if (((go_math_IsInf(a,cast<int64_t>(0ULL)) && go_math_IsInf(b,cast<int64_t>(0ULL))) && (go_math_Signbit(a) != go_math_Signbit(b)))) {
gekko_CPU_setFPSCRBit(c,cast<uint32_t>(8388608ULL));
return gekko_qnan;
}
return (a + b);
}
}
// tools/cpu/gekko/fpu.go:159:1
double gekko_CPU_fsub(gekko_CPU* c,double a,double b){
{
if ((gekko_isSNaN(a) || gekko_isSNaN(b))) {
gekko_CPU_setFPSCRBit(c,cast<uint32_t>(16777216ULL));
}
if (((go_math_IsInf(a,cast<int64_t>(0ULL)) && go_math_IsInf(b,cast<int64_t>(0ULL))) && (go_math_Signbit(a) == go_math_Signbit(b)))) {
gekko_CPU_setFPSCRBit(c,cast<uint32_t>(8388608ULL));
return gekko_qnan;
}
return (a - b);
}
}
// tools/cpu/gekko/fpu.go:170:1
double gekko_CPU_fmul(gekko_CPU* c,double a,double b){
{
if ((gekko_isSNaN(a) || gekko_isSNaN(b))) {
gekko_CPU_setFPSCRBit(c,cast<uint32_t>(16777216ULL));
}
if ((((go_math_IsInf(a,cast<int64_t>(0ULL)) && (b == cast<double>(0ULL)))) || (((a == cast<double>(0ULL)) && go_math_IsInf(b,cast<int64_t>(0ULL)))))) {
gekko_CPU_setFPSCRBit(c,cast<uint32_t>(1048576ULL));
return gekko_qnan;
}
return (a * b);
}
}
// tools/cpu/gekko/fpu.go:181:1
double gekko_CPU_fdiv(gekko_CPU* c,double a,double b){
{
if ((gekko_isSNaN(a) || gekko_isSNaN(b))) {
gekko_CPU_setFPSCRBit(c,cast<uint32_t>(16777216ULL));
}
{
if ((go_math_IsInf(a,cast<int64_t>(0ULL)) && go_math_IsInf(b,cast<int64_t>(0ULL)))){
gekko_CPU_setFPSCRBit(c,cast<uint32_t>(4194304ULL));
return gekko_qnan;
}
else if (((a == cast<double>(0ULL)) && (b == cast<double>(0ULL)))){
gekko_CPU_setFPSCRBit(c,cast<uint32_t>(2097152ULL));
return gekko_qnan;
}
else if (((b == cast<double>(0ULL)) && (!go_math_IsNaN(a)))){
gekko_CPU_setFPSCRBit(c,cast<uint32_t>(67108864ULL));
return go_math_Inf(cast<int64_t>((gekko_sign(a) * gekko_sign(b))));
}
}
tmp56:;
return (a / b);
}
}
// tools/cpu/gekko/fpu.go:201:1
int64_t gekko_sign(double v){
{
if (go_math_Signbit(v)) {
return cast<int64_t>(-1ULL);
}
return cast<int64_t>(1ULL);
}
}
// tools/cpu/gekko/fpu.go:222:1
double gekko_CPU_fmaddRaw(gekko_CPU* c,double a,double cc,double b){
{
if (((gekko_isSNaN(a) || gekko_isSNaN(b)) || gekko_isSNaN(cc))) {
gekko_CPU_setFPSCRBit(c,cast<uint32_t>(16777216ULL));
}
if ((((go_math_IsInf(a,cast<int64_t>(0ULL)) && (cc == cast<double>(0ULL)))) || (((a == cast<double>(0ULL)) && go_math_IsInf(cc,cast<int64_t>(0ULL)))))) {
gekko_CPU_setFPSCRBit(c,cast<uint32_t>(1048576ULL));
return gekko_qnan;
}
double prod = (a * cc);
if (((go_math_IsInf(prod,cast<int64_t>(0ULL)) && go_math_IsInf(b,cast<int64_t>(0ULL))) && (go_math_Signbit(prod) != go_math_Signbit(b)))) {
gekko_CPU_setFPSCRBit(c,cast<uint32_t>(8388608ULL));
return gekko_qnan;
}
return go_math_FMA(a,cc,b);
}
}
// tools/cpu/gekko/fpu.go:238:1
double gekko_CPU_fsqrt(gekko_CPU* c,double v){
{
if (gekko_isSNaN(v)) {
gekko_CPU_setFPSCRBit(c,cast<uint32_t>(16777216ULL));
}
if (((v < cast<double>(0ULL)) && (!go_math_IsNaN(v)))) {
gekko_CPU_setFPSCRBit(c,cast<uint32_t>(512ULL));
return gekko_qnan;
}
return go_math_Sqrt(v);
}
}
// tools/cpu/gekko/fpu.go:259:1
double gekko_CPU_fres(gekko_CPU* c,double v){
{
if (gekko_isSNaN(v)) {
gekko_CPU_setFPSCRBit(c,cast<uint32_t>(16777216ULL));
}
if ((v == cast<double>(0ULL))) {
gekko_CPU_setFPSCRBit(c,cast<uint32_t>(67108864ULL));
return go_math_Inf(gekko_sign(v));
}
return gekko_f32((cast<double>(1ULL) / v));
}
}
// tools/cpu/gekko/fpu.go:270:1
double gekko_CPU_frsqrte(gekko_CPU* c,double v){
{
if (gekko_isSNaN(v)) {
gekko_CPU_setFPSCRBit(c,cast<uint32_t>(16777216ULL));
}
if ((v == cast<double>(0ULL))) {
gekko_CPU_setFPSCRBit(c,cast<uint32_t>(67108864ULL));
return go_math_Inf(gekko_sign(v));
}
if ((v < cast<double>(0ULL))) {
gekko_CPU_setFPSCRBit(c,cast<uint32_t>(512ULL));
return gekko_qnan;
}
return (cast<double>(1ULL) / go_math_Sqrt(v));
}
}
// tools/cpu/gekko/fpu.go:286:1
void gekko_CPU_exec59(gekko_CPU* c,uint32_t w){
{
auto tmp57 = std::make_tuple(gekko_rs(w),gekko_ra(w),gekko_rb(w),gekko_rc(w));
uint32_t d = std::get<0>(tmp57);
uint32_t a = std::get<1>(tmp57);
uint32_t b = std::get<2>(tmp57);
uint32_t cc = std::get<3>(tmp57);
auto tmp58 = std::make_tuple(c->FPR[a].PS0,c->FPR[b].PS0,c->FPR[cc].PS0);
double fa = std::get<0>(tmp58);
double fb = std::get<1>(tmp58);
double fc = std::get<2>(tmp58);
{
switch(gekko_xo5(w)){
case cast<uint32_t>(18ULL):{
gekko_CPU_fpResultS(c,w,d,gekko_CPU_fdiv(c,fa,fb));
break;}
case cast<uint32_t>(20ULL):{
gekko_CPU_fpResultS(c,w,d,gekko_CPU_fsub(c,fa,fb));
break;}
case cast<uint32_t>(21ULL):{
gekko_CPU_fpResultS(c,w,d,gekko_CPU_fadd(c,fa,fb));
break;}
case cast<uint32_t>(22ULL):{
gekko_CPU_fpResultS(c,w,d,gekko_CPU_fsqrt(c,fb));
break;}
case cast<uint32_t>(24ULL):{
gekko_CPU_fpResultS(c,w,d,gekko_CPU_fres(c,fb));
break;}
case cast<uint32_t>(25ULL):{
gekko_CPU_fpResultS(c,w,d,gekko_CPU_fmul(c,fa,fc));
break;}
case cast<uint32_t>(28ULL):{
gekko_CPU_fpResultS(c,w,d,gekko_CPU_fmaddRaw(c,fa,fc,cast<double>(-fb)));
break;}
case cast<uint32_t>(29ULL):{
gekko_CPU_fpResultS(c,w,d,gekko_CPU_fmaddRaw(c,fa,fc,fb));
break;}
case cast<uint32_t>(30ULL):{
gekko_CPU_fpResultS(c,w,d,cast<double>(-gekko_CPU_fmaddRaw(c,fa,fc,cast<double>(-fb))));
break;}
case cast<uint32_t>(31ULL):{
gekko_CPU_fpResultS(c,w,d,cast<double>(-gekko_CPU_fmaddRaw(c,fa,fc,fb)));
break;}
default:{
gekko_CPU_Halt(c,std::string("gekko: unimplemented opcode 59 extended %d (word 0x%08X) at 0x%08X",66),gekko_xo5(w),w,cast<uint32_t>((c->PC - cast<uint32_t>(4ULL))));
break;}
}}
}
}
// tools/cpu/gekko/fpu.go:317:1
void gekko_CPU_exec63(gekko_CPU* c,uint32_t w){
{
auto tmp59 = std::make_tuple(gekko_rs(w),gekko_ra(w),gekko_rb(w),gekko_rc(w));
uint32_t d = std::get<0>(tmp59);
uint32_t a = std::get<1>(tmp59);
uint32_t b = std::get<2>(tmp59);
uint32_t cc = std::get<3>(tmp59);
auto tmp60 = std::make_tuple(c->FPR[a].PS0,c->FPR[b].PS0,c->FPR[cc].PS0);
double fa = std::get<0>(tmp60);
double fb = std::get<1>(tmp60);
double fc = std::get<2>(tmp60);
{
switch(gekko_xo5(w)){
case cast<uint32_t>(18ULL):{
gekko_CPU_fpResult(c,w,d,gekko_CPU_fdiv(c,fa,fb));
return ;
break;}
case cast<uint32_t>(20ULL):{
gekko_CPU_fpResult(c,w,d,gekko_CPU_fsub(c,fa,fb));
return ;
break;}
case cast<uint32_t>(21ULL):{
gekko_CPU_fpResult(c,w,d,gekko_CPU_fadd(c,fa,fb));
return ;
break;}
case cast<uint32_t>(22ULL):{
gekko_CPU_fpResult(c,w,d,gekko_CPU_fsqrt(c,fb));
return ;
break;}
case cast<uint32_t>(23ULL):{
double v = fb;
if (((fa >= cast<double>(0ULL)) && (!go_math_IsNaN(fa)))) {
v = fc;
}
c->FPR[d].PS0 = v;
if (gekko_rcbit(w)) {
gekko_CPU_setCR1(c);
}
return ;
break;}
case cast<uint32_t>(25ULL):{
gekko_CPU_fpResult(c,w,d,gekko_CPU_fmul(c,fa,fc));
return ;
break;}
case cast<uint32_t>(26ULL):{
gekko_CPU_fpResult(c,w,d,gekko_CPU_frsqrte(c,fb));
return ;
break;}
case cast<uint32_t>(28ULL):{
gekko_CPU_fpResult(c,w,d,gekko_CPU_fmaddRaw(c,fa,fc,cast<double>(-fb)));
return ;
break;}
case cast<uint32_t>(29ULL):{
gekko_CPU_fpResult(c,w,d,gekko_CPU_fmaddRaw(c,fa,fc,fb));
return ;
break;}
case cast<uint32_t>(30ULL):{
gekko_CPU_fpResult(c,w,d,cast<double>(-gekko_CPU_fmaddRaw(c,fa,fc,cast<double>(-fb))));
return ;
break;}
case cast<uint32_t>(31ULL):{
gekko_CPU_fpResult(c,w,d,cast<double>(-gekko_CPU_fmaddRaw(c,fa,fc,fb)));
return ;
break;}
}}
{
switch(gekko_xo10(w)){
case cast<uint32_t>(0ULL):{
gekko_CPU_fcmp(c,gekko_crfD(w),fa,fb,false);
break;}
case cast<uint32_t>(32ULL):{
gekko_CPU_fcmp(c,gekko_crfD(w),fa,fb,true);
break;}
case cast<uint32_t>(12ULL):{
gekko_CPU_fpResult(c,w,d,gekko_f32(fb));
break;}
case cast<uint32_t>(14ULL):{
gekko_CPU_fctiw(c,w,d,fb,false);
break;}
case cast<uint32_t>(15ULL):{
gekko_CPU_fctiw(c,w,d,fb,true);
break;}
case cast<uint32_t>(40ULL):{
c->FPR[d].PS0 = go_math_Float64frombits(cast<uint64_t>((go_math_Float64bits(fb) ^ cast<uint64_t>(9223372036854775808ULL))));
gekko_CPU_rcF(c,w);
break;}
case cast<uint32_t>(72ULL):{
c->FPR[d].PS0 = fb;
gekko_CPU_rcF(c,w);
break;}
case cast<uint32_t>(136ULL):{
c->FPR[d].PS0 = go_math_Float64frombits(cast<uint64_t>((go_math_Float64bits(fb) | cast<uint64_t>(9223372036854775808ULL))));
gekko_CPU_rcF(c,w);
break;}
case cast<uint32_t>(264ULL):{
c->FPR[d].PS0 = go_math_Float64frombits((go_math_Float64bits(fb) & ~(cast<uint64_t>(9223372036854775808ULL))));
gekko_CPU_rcF(c,w);
break;}
case cast<uint32_t>(64ULL):{
uint32_t f = cast<uint32_t>(((shr<uint32_t>(c->FPSCR,(cast<uint32_t>((cast<uint32_t>(28ULL) - cast<uint32_t>((cast<uint32_t>(4ULL) * gekko_crfS(w)))))))) & cast<uint32_t>(15ULL)));
gekko_CPU_SetCRField(c,gekko_crfD(w),f);
c->FPSCR &= ~(shl<uint32_t>(cast<uint32_t>(15ULL),(cast<uint32_t>((cast<uint32_t>(28ULL) - cast<uint32_t>((cast<uint32_t>(4ULL) * gekko_crfS(w))))))));
break;}
case cast<uint32_t>(38ULL):{
gekko_CPU_setFPSCRBit(c,shl<uint32_t>(cast<uint32_t>(1ULL),(cast<uint32_t>((cast<uint32_t>(31ULL) - d)))));
gekko_CPU_rcF(c,w);
break;}
case cast<uint32_t>(70ULL):{
c->FPSCR &= ~(shl<uint32_t>(cast<uint32_t>(1ULL),(cast<uint32_t>((cast<uint32_t>(31ULL) - d)))));
gekko_CPU_rcF(c,w);
break;}
case cast<uint32_t>(134ULL):{
uint32_t sh = cast<uint32_t>((cast<uint32_t>(28ULL) - cast<uint32_t>((cast<uint32_t>(4ULL) * gekko_crfD(w)))));
c->FPSCR = cast<uint32_t>((((c->FPSCR & ~((shl<uint32_t>(cast<uint32_t>(15ULL),sh))))) | (shl<uint32_t>((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(12ULL))) & cast<uint32_t>(15ULL)))),sh))));
gekko_CPU_rcF(c,w);
break;}
case cast<uint32_t>(583ULL):{
c->FPR[d].PS0 = go_math_Float64frombits(cast<uint64_t>(c->FPSCR));
gekko_CPU_rcF(c,w);
break;}
case cast<uint32_t>(711ULL):{
uint32_t mask = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(17ULL))) & cast<uint32_t>(255ULL)));
uint32_t m = cast<uint32_t>(0ULL);
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(8ULL));i++){
if ((cast<uint32_t>((mask & (shr<uint32_t>(cast<uint32_t>(128ULL),i)))) != cast<uint32_t>(0ULL))) {
m |= shl<uint32_t>(cast<uint32_t>(15ULL),(cast<int64_t>((cast<int64_t>(28ULL) - cast<int64_t>((cast<int64_t>(4ULL) * i))))));
}
}
}c->FPSCR = cast<uint32_t>((((c->FPSCR & ~(m))) | (cast<uint32_t>((cast<uint32_t>(go_math_Float64bits(fb)) & m)))));
gekko_CPU_rcF(c,w);
break;}
default:{
gekko_CPU_Halt(c,std::string("gekko: unimplemented opcode 63 extended %d (word 0x%08X) at 0x%08X",66),gekko_xo10(w),w,cast<uint32_t>((c->PC - cast<uint32_t>(4ULL))));
break;}
}}
}
}
// tools/cpu/gekko/fpu.go:420:1
void gekko_CPU_rcF(gekko_CPU* c,uint32_t w){
{
if (gekko_rcbit(w)) {
gekko_CPU_setCR1(c);
}
}
}
// tools/cpu/gekko/fpu.go:429:1
void gekko_CPU_fcmp(gekko_CPU* c,uint32_t crf,double a,double b,bool ordered){
{
uint32_t f={};
{
if ((go_math_IsNaN(a) || go_math_IsNaN(b))){
f = cast<uint32_t>(1ULL);
if ((gekko_isSNaN(a) || gekko_isSNaN(b))) {
gekko_CPU_setFPSCRBit(c,cast<uint32_t>(16777216ULL));
if (ordered) {
gekko_CPU_setFPSCRBit(c,cast<uint32_t>(524288ULL));
}
}
else if (ordered) {
gekko_CPU_setFPSCRBit(c,cast<uint32_t>(524288ULL));
}
}
else if ((a < b)){
f = cast<uint32_t>(8ULL);
}
else if ((a > b)){
f = cast<uint32_t>(4ULL);
}
else {
f = cast<uint32_t>(2ULL);
}
}
tmp61:;
gekko_CPU_SetCRField(c,crf,f);
c->FPSCR = cast<uint32_t>((((c->FPSCR & ~(cast<uint32_t>(126976ULL)))) | (shl<uint32_t>(f,cast<int64_t>(12ULL)))));
}
}
// tools/cpu/gekko/fpu.go:458:1
void gekko_CPU_fctiw(gekko_CPU* c,uint32_t w,uint32_t d,double v,bool truncate){
{
int32_t i={};
{
if (go_math_IsNaN(v)){
gekko_CPU_setFPSCRBit(c,cast<uint32_t>(256ULL));
i = cast<int32_t>(-2147483648ULL);
}
else if ((v > 2147483647.0)){
gekko_CPU_setFPSCRBit(c,cast<uint32_t>(256ULL));
i = cast<int32_t>(2147483647ULL);
}
else if ((v < cast<double>(-2147483648.0))){
gekko_CPU_setFPSCRBit(c,cast<uint32_t>(256ULL));
i = cast<int32_t>(-2147483648ULL);
}
else {
if (truncate) {
i = cast<int32_t>(go_math_Trunc(v));
}
else {
i = cast<int32_t>(gekko_CPU_roundToNearest(c,v));
}
}
}
tmp62:;
c->FPR[d].PS0 = go_math_Float64frombits(cast<uint64_t>((cast<uint64_t>(18444492273895866368ULL) | cast<uint64_t>(cast<uint32_t>(i)))));
gekko_CPU_rcF(c,w);
}
}
// tools/cpu/gekko/fpu.go:486:1
double gekko_CPU_roundToNearest(gekko_CPU* c,double v){
{
return go_math_RoundToEven(v);
}
}
// tools/cpu/gekko/gekko.go:56:1
std::string gekko_Flow_String(gekko_Flow f){
{
{
switch(f){
case cast<gekko_Flow>(0ULL):{
return std::string("seq",3);
break;}
case cast<gekko_Flow>(1ULL):{
return std::string("branch",6);
break;}
case cast<gekko_Flow>(2ULL):{
return std::string("jump",4);
break;}
case cast<gekko_Flow>(3ULL):{
return std::string("call",4);
break;}
case cast<gekko_Flow>(4ULL):{
return std::string("return",6);
break;}
case cast<gekko_Flow>(5ULL):{
return std::string("indjump",7);
break;}
case cast<gekko_Flow>(6ULL):{
return std::string("indcall",7);
break;}
case cast<gekko_Flow>(7ULL):{
return std::string("stop",4);
break;}
}}
return std::string("?",1);
}
}
// tools/cpu/gekko/gekko.go:94:1
std::string gekko_Inst_String(gekko_Inst in){
{
return go_fmt_Sprintf(std::string("$%08X: %s",9),in.Addr,in.Text);
}
}
// tools/cpu/gekko/gekko.go:101:1
uint32_t gekko_opcd(uint32_t w){
{
return shr<uint32_t>(w,cast<int64_t>(26ULL));
}
}
// tools/cpu/gekko/gekko.go:102:1
uint32_t gekko_rs(uint32_t w){
{
return cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(21ULL))) & cast<uint32_t>(31ULL)));
}
}
// tools/cpu/gekko/gekko.go:103:1
uint32_t gekko_ra(uint32_t w){
{
return cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(31ULL)));
}
}
// tools/cpu/gekko/gekko.go:104:1
uint32_t gekko_rb(uint32_t w){
{
return cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(11ULL))) & cast<uint32_t>(31ULL)));
}
}
// tools/cpu/gekko/gekko.go:105:1
uint32_t gekko_rc(uint32_t w){
{
return cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(6ULL))) & cast<uint32_t>(31ULL)));
}
}
// tools/cpu/gekko/gekko.go:106:1
uint32_t gekko_xo10(uint32_t w){
{
return cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(1ULL))) & cast<uint32_t>(1023ULL)));
}
}
// tools/cpu/gekko/gekko.go:107:1
uint32_t gekko_xo5(uint32_t w){
{
return cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(1ULL))) & cast<uint32_t>(31ULL)));
}
}
// tools/cpu/gekko/gekko.go:108:1
uint32_t gekko_xo6(uint32_t w){
{
return cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(1ULL))) & cast<uint32_t>(63ULL)));
}
}
// tools/cpu/gekko/gekko.go:109:1
bool gekko_rcbit(uint32_t w){
{
return (cast<uint32_t>((w & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL));
}
}
// tools/cpu/gekko/gekko.go:110:1
bool gekko_oe(uint32_t w){
{
return (cast<uint32_t>((w & cast<uint32_t>(1024ULL))) != cast<uint32_t>(0ULL));
}
}
// tools/cpu/gekko/gekko.go:111:1
bool gekko_lk(uint32_t w){
{
return (cast<uint32_t>((w & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL));
}
}
// tools/cpu/gekko/gekko.go:112:1
bool gekko_aa(uint32_t w){
{
return (cast<uint32_t>((w & cast<uint32_t>(2ULL))) != cast<uint32_t>(0ULL));
}
}
// tools/cpu/gekko/gekko.go:113:1
int32_t gekko_simm(uint32_t w){
{
return cast<int32_t>(cast<int16_t>(w));
}
}
// tools/cpu/gekko/gekko.go:114:1
uint32_t gekko_uimm(uint32_t w){
{
return cast<uint32_t>((w & cast<uint32_t>(65535ULL)));
}
}
// tools/cpu/gekko/gekko.go:115:1
uint32_t gekko_crfD(uint32_t w){
{
return cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(23ULL))) & cast<uint32_t>(7ULL)));
}
}
// tools/cpu/gekko/gekko.go:116:1
uint32_t gekko_crfS(uint32_t w){
{
return cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(18ULL))) & cast<uint32_t>(7ULL)));
}
}
// tools/cpu/gekko/gekko.go:117:1
uint32_t gekko_shOf(uint32_t w){
{
return cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(11ULL))) & cast<uint32_t>(31ULL)));
}
}
// tools/cpu/gekko/gekko.go:118:1
uint32_t gekko_mbOf(uint32_t w){
{
return cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(6ULL))) & cast<uint32_t>(31ULL)));
}
}
// tools/cpu/gekko/gekko.go:119:1
uint32_t gekko_meOf(uint32_t w){
{
return cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(1ULL))) & cast<uint32_t>(31ULL)));
}
}
// tools/cpu/gekko/gekko.go:124:1
uint32_t gekko_sprOf(uint32_t w){
{
uint32_t raw = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(11ULL))) & cast<uint32_t>(1023ULL)));
return cast<uint32_t>(((shl<uint32_t>((cast<uint32_t>((raw & cast<uint32_t>(31ULL)))),cast<int64_t>(5ULL))) | (shr<uint32_t>(raw,cast<int64_t>(5ULL)))));
}
}
// tools/cpu/gekko/gekko.go:197:1
std::string gekko_sprStr(uint32_t n){
{
{
auto tmp63 = lookup(gekko_sprName,n);
std::string s = std::get<0>(tmp63);
bool ok = std::get<1>(tmp63);
if (ok) {
return s;
}
}
return go_fmt_Sprintf(std::string("%d",2),n);
}
}
// tools/cpu/gekko/mmu.go:37:1
std::tuple<uint32_t,bool> gekko_CPU_Translate_reference(gekko_CPU* c,uint32_t ea,bool store,bool insn){
{
bool on = (cast<uint32_t>((c->MSR & cast<uint32_t>(16ULL))) != cast<uint32_t>(0ULL));
if (insn) {
on = (cast<uint32_t>((c->MSR & cast<uint32_t>(32ULL))) != cast<uint32_t>(0ULL));
}
if ((!on)) {
return {ea,true};
}
if ((gekko_LockedCache_Enabled(&(c->LC)) && gekko_LockedCache_Contains(&(c->LC),ea))) {
return {ea,true};
}
std::array<std::array<uint32_t,2>,4>* bats = (&c->DBAT);
if (insn) {
bats = (&c->IBAT);
}
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(4ULL));i++){
auto tmp64 = std::make_tuple((*bats)[i][cast<int64_t>(0ULL)],(*bats)[i][cast<int64_t>(1ULL)]);
uint32_t upper = std::get<0>(tmp64);
uint32_t lower = std::get<1>(tmp64);
{
auto tmp65 = gekko_batMatch(upper,lower,ea,(cast<uint32_t>((c->MSR & cast<uint32_t>(16384ULL))) != cast<uint32_t>(0ULL)));
uint32_t pa = std::get<0>(tmp65);
bool ok = std::get<1>(tmp65);
if (ok) {
return {pa,true};
}
}
}
}std::string kind = std::string("data",4);
if (insn) {
kind = std::string("instruction",11);
}
gekko_CPU_Halt(c,std::string("gekko: no BAT maps the %s address 0x%08X (PC 0x%08X); this machine does not use a page table",92),kind,ea,c->PC);
return {cast<uint32_t>(0ULL),false};
}
}
// tools/cpu/gekko/mmu.go:83:1
std::tuple<uint32_t,bool> gekko_batMatch(uint32_t upper,uint32_t lower,uint32_t ea,bool user){
{
bool vs = (cast<uint32_t>((upper & cast<uint32_t>(2ULL))) != cast<uint32_t>(0ULL));
bool vp = (cast<uint32_t>((upper & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL));
if ((user && (!vp))) {
return {cast<uint32_t>(0ULL),false};
}
if (((!user) && (!vs))) {
return {cast<uint32_t>(0ULL),false};
}
uint32_t bepi = cast<uint32_t>((upper & cast<uint32_t>(4294836224ULL)));
uint32_t bl = cast<uint32_t>(((shr<uint32_t>(upper,cast<int64_t>(2ULL))) & cast<uint32_t>(2047ULL)));
uint32_t offsetMask = cast<uint32_t>(((shl<uint32_t>(bl,cast<int64_t>(17ULL))) | cast<uint32_t>(131071ULL)));
uint32_t blockMask = cast<uint32_t>(~offsetMask);
if ((cast<uint32_t>((ea & blockMask)) != cast<uint32_t>((bepi & blockMask)))) {
return {cast<uint32_t>(0ULL),false};
}
uint32_t brpn = cast<uint32_t>((lower & cast<uint32_t>(4294836224ULL)));
return {cast<uint32_t>(((cast<uint32_t>((brpn & blockMask))) | (cast<uint32_t>((ea & offsetMask))))),true};
}
}
// tools/cpu/gekko/mmu.go:116:1
std::string gekko_CPU_BATString(gekko_CPU* c){
{
std::string s={};
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(4ULL));i++){
s += gekko_batLine(std::string("IBAT",4),i,c->IBAT[i][cast<int64_t>(0ULL)],c->IBAT[i][cast<int64_t>(1ULL)]);
}
}{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(4ULL));i++){
s += gekko_batLine(std::string("DBAT",4),i,c->DBAT[i][cast<int64_t>(0ULL)],c->DBAT[i][cast<int64_t>(1ULL)]);
}
}return s;
}
}
// tools/cpu/gekko/mmu.go:127:1
std::string gekko_batLine(std::string kind,int64_t i,uint32_t upper,uint32_t lower){
{
if ((cast<uint32_t>((upper & cast<uint32_t>(3ULL))) == cast<uint32_t>(0ULL))) {
return std::string("",0);
}
uint32_t bl = cast<uint32_t>(((shr<uint32_t>(upper,cast<int64_t>(2ULL))) & cast<uint32_t>(2047ULL)));
uint64_t size = cast<uint64_t>((cast<uint64_t>(((cast<uint64_t>((cast<uint64_t>(bl) + cast<uint64_t>(1ULL)))) * cast<uint64_t>(128ULL))) * cast<uint64_t>(1024ULL)));
return go_fmt_Sprintf(std::string("%s%d  0x%08X..0x%08X -> 0x%08X  (%d KiB)\012",41),kind,i,cast<uint32_t>((upper & cast<uint32_t>(4294836224ULL))),cast<uint64_t>((cast<uint64_t>(cast<uint32_t>((upper & cast<uint32_t>(4294836224ULL)))) + size)),cast<uint32_t>((lower & cast<uint32_t>(4294836224ULL))),divi<uint64_t>(size,cast<uint64_t>(1024ULL)));
}
}
// tools/cpu/gekko/ps.go:56:1
std::tuple<uint32_t,int32_t> gekko_gqrLoad(uint32_t g){
uint32_t typ{};
int32_t scale{};
{
typ = cast<uint32_t>(((shr<uint32_t>(g,cast<int64_t>(16ULL))) & cast<uint32_t>(7ULL)));
scale = shr<int32_t>(cast<int32_t>(shl<uint32_t>(g,cast<int64_t>(2ULL))),cast<int64_t>(26ULL));
return {typ,scale};
}
}
// tools/cpu/gekko/ps.go:63:1
std::tuple<uint32_t,int32_t> gekko_gqrStore(uint32_t g){
uint32_t typ{};
int32_t scale{};
{
typ = cast<uint32_t>((g & cast<uint32_t>(7ULL)));
scale = shr<int32_t>(cast<int32_t>(shl<uint32_t>(g,cast<int64_t>(18ULL))),cast<int64_t>(26ULL));
return {typ,scale};
}
}
// tools/cpu/gekko/ps.go:71:1
double gekko_CPU_dequantize(gekko_CPU* c,uint32_t raw,uint32_t typ,int32_t scale){
{
double v={};
{
switch(typ){
case cast<uint32_t>(0ULL):{
return cast<double>(go_math_Float32frombits(raw));
break;}
case cast<uint32_t>(4ULL):{
v = cast<double>(cast<uint8_t>(raw));
break;}
case cast<uint32_t>(5ULL):{
v = cast<double>(cast<uint16_t>(raw));
break;}
case cast<uint32_t>(6ULL):{
v = cast<double>(cast<int8_t>(raw));
break;}
case cast<uint32_t>(7ULL):{
v = cast<double>(cast<int16_t>(raw));
break;}
default:{
gekko_CPU_Halt(c,std::string("gekko: a quantised load used the undefined GQR type %d at 0x%08X",64),typ,cast<uint32_t>((c->PC - cast<uint32_t>(4ULL))));
return cast<double>(0ULL);
break;}
}}
return (v * go_math_Ldexp(cast<double>(1ULL),cast<int64_t>(cast<int32_t>(-scale))));
}
}
// tools/cpu/gekko/ps.go:97:1
uint32_t gekko_CPU_quantize(gekko_CPU* c,double v,uint32_t typ,int32_t scale){
{
if ((typ == cast<uint32_t>(0ULL))) {
return go_math_Float32bits(cast<float>(v));
}
v *= go_math_Ldexp(cast<double>(1ULL),cast<int64_t>(scale));
if (go_math_IsNaN(v)) {
v = cast<double>(0ULL);
}
{
switch(typ){
case cast<uint32_t>(4ULL):{
return cast<uint32_t>(cast<uint8_t>(gekko_clamp(v,cast<double>(0ULL),cast<double>(255ULL))));
break;}
case cast<uint32_t>(5ULL):{
return cast<uint32_t>(cast<uint16_t>(gekko_clamp(v,cast<double>(0ULL),cast<double>(65535ULL))));
break;}
case cast<uint32_t>(6ULL):{
return cast<uint32_t>(cast<uint8_t>(cast<int8_t>(gekko_clamp(v,cast<double>(-cast<int64_t>(128ULL)),cast<double>(127ULL)))));
break;}
case cast<uint32_t>(7ULL):{
return cast<uint32_t>(cast<uint16_t>(cast<int16_t>(gekko_clamp(v,cast<double>(-cast<int64_t>(32768ULL)),cast<double>(32767ULL)))));
break;}
}}
gekko_CPU_Halt(c,std::string("gekko: a quantised store used the undefined GQR type %d at 0x%08X",65),typ,cast<uint32_t>((c->PC - cast<uint32_t>(4ULL))));
return cast<uint32_t>(0ULL);
}
}
// tools/cpu/gekko/ps.go:123:1
double gekko_clamp(double v,double lo,double hi){
{
v = go_math_Trunc(v);
if ((v < lo)) {
return lo;
}
if ((v > hi)) {
return hi;
}
return v;
}
}
// tools/cpu/gekko/ps.go:135:1
uint32_t gekko_qsize(uint32_t typ){
{
{
switch(typ){
case cast<uint32_t>(0ULL):{
return cast<uint32_t>(4ULL);
break;}
case cast<uint32_t>(5ULL):case cast<uint32_t>(7ULL):{
return cast<uint32_t>(2ULL);
break;}
case cast<uint32_t>(4ULL):case cast<uint32_t>(6ULL):{
return cast<uint32_t>(1ULL);
break;}
}}
return cast<uint32_t>(4ULL);
}
}
// tools/cpu/gekko/ps.go:148:1
uint32_t gekko_CPU_readQ(gekko_CPU* c,uint32_t ea,uint32_t typ){
{
{
switch(gekko_qsize(typ)){
case cast<uint32_t>(1ULL):{
return cast<uint32_t>(gekko_CPU_read8(c,ea));
break;}
case cast<uint32_t>(2ULL):{
return cast<uint32_t>(gekko_CPU_read16(c,ea));
break;}
}}
return gekko_CPU_read32(c,ea);
}
}
// tools/cpu/gekko/ps.go:158:1
void gekko_CPU_writeQ(gekko_CPU* c,uint32_t ea,uint32_t typ,uint32_t v){
{
{
switch(gekko_qsize(typ)){
case cast<uint32_t>(1ULL):{
gekko_CPU_write8(c,ea,cast<uint8_t>(v));
break;}
case cast<uint32_t>(2ULL):{
gekko_CPU_write16(c,ea,cast<uint16_t>(v));
break;}
default:{
gekko_CPU_write32(c,ea,v);
break;}
}}
}
}
// tools/cpu/gekko/ps.go:173:1
void gekko_CPU_psqLoad(gekko_CPU* c,uint32_t d,uint32_t ea,uint32_t gqr,uint32_t wBit){
{
auto tmp66 = gekko_gqrLoad(c->GQR[cast<uint32_t>((gqr & cast<uint32_t>(7ULL)))]);
uint32_t typ = std::get<0>(tmp66);
int32_t scale = std::get<1>(tmp66);
if ((wBit != cast<uint32_t>(0ULL))) {
c->FPR[d].PS0 = gekko_CPU_dequantize(c,gekko_CPU_readQ(c,ea,typ),typ,scale);
c->FPR[d].PS1 = 1.0;
return ;
}
uint32_t sz = gekko_qsize(typ);
c->FPR[d].PS0 = gekko_CPU_dequantize(c,gekko_CPU_readQ(c,ea,typ),typ,scale);
c->FPR[d].PS1 = gekko_CPU_dequantize(c,gekko_CPU_readQ(c,cast<uint32_t>((ea + sz)),typ),typ,scale);
}
}
// tools/cpu/gekko/ps.go:185:1
void gekko_CPU_psqStore(gekko_CPU* c,uint32_t s,uint32_t ea,uint32_t gqr,uint32_t wBit){
{
auto tmp67 = gekko_gqrStore(c->GQR[cast<uint32_t>((gqr & cast<uint32_t>(7ULL)))]);
uint32_t typ = std::get<0>(tmp67);
int32_t scale = std::get<1>(tmp67);
gekko_CPU_writeQ(c,ea,typ,gekko_CPU_quantize(c,c->FPR[s].PS0,typ,scale));
if ((wBit != cast<uint32_t>(0ULL))) {
return ;
}
gekko_CPU_writeQ(c,cast<uint32_t>((ea + gekko_qsize(typ))),typ,gekko_CPU_quantize(c,c->FPR[s].PS1,typ,scale));
}
}
// tools/cpu/gekko/ps.go:195:1
void gekko_CPU_execPSQ(gekko_CPU* c,uint32_t w){
{
if ((!gekko_CPU_psEnabled(c))) {
gekko_CPU_Halt(c,std::string("gekko: a quantised load/store ran with the paired-single unit off (HID2 = 0x%08X) at 0x%08X",91),c->HID2,cast<uint32_t>((c->PC - cast<uint32_t>(4ULL))));
return ;
}
uint32_t d = gekko_rs(w);
uint32_t gqr = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(12ULL))) & cast<uint32_t>(7ULL)));
uint32_t wBit = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(15ULL))) & cast<uint32_t>(1ULL)));
uint32_t disp = cast<uint32_t>(gekko_psqDisp(w));
{
switch(gekko_opcd(w)){
case cast<uint32_t>(56ULL):{
gekko_CPU_psqLoad(c,d,cast<uint32_t>((gekko_CPU_raOrZero(c,w) + disp)),gqr,wBit);
break;}
case cast<uint32_t>(57ULL):{
uint32_t ea = cast<uint32_t>((c->GPR[gekko_ra(w)] + disp));
gekko_CPU_psqLoad(c,d,ea,gqr,wBit);
c->GPR[gekko_ra(w)] = ea;
break;}
case cast<uint32_t>(60ULL):{
gekko_CPU_psqStore(c,d,cast<uint32_t>((gekko_CPU_raOrZero(c,w) + disp)),gqr,wBit);
break;}
case cast<uint32_t>(61ULL):{
uint32_t ea = cast<uint32_t>((c->GPR[gekko_ra(w)] + disp));
gekko_CPU_psqStore(c,d,ea,gqr,wBit);
c->GPR[gekko_ra(w)] = ea;
break;}
}}
}
}
// tools/cpu/gekko/ps.go:223:1
void gekko_CPU_execPS(gekko_CPU* c,uint32_t w,uint32_t pc){
{
{
switch(gekko_xo6(w)){
case cast<uint32_t>(6ULL):case cast<uint32_t>(7ULL):case cast<uint32_t>(38ULL):case cast<uint32_t>(39ULL):{
if ((!gekko_CPU_psEnabled(c))) {
gekko_CPU_Halt(c,std::string("gekko: an indexed quantised load/store ran with the paired-single unit off at 0x%08X",84),pc);
return ;
}
uint32_t d = gekko_rs(w);
uint32_t gqr = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(12ULL))) & cast<uint32_t>(7ULL)));
uint32_t wBit = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(15ULL))) & cast<uint32_t>(1ULL)));
{
switch(gekko_xo6(w)){
case cast<uint32_t>(6ULL):{
gekko_CPU_psqLoad(c,d,gekko_CPU_eax(c,w),gqr,wBit);
break;}
case cast<uint32_t>(7ULL):{
gekko_CPU_psqStore(c,d,gekko_CPU_eax(c,w),gqr,wBit);
break;}
case cast<uint32_t>(38ULL):{
uint32_t ea = gekko_CPU_eaxU(c,w);
gekko_CPU_psqLoad(c,d,ea,gqr,wBit);
c->GPR[gekko_ra(w)] = ea;
break;}
case cast<uint32_t>(39ULL):{
uint32_t ea = gekko_CPU_eaxU(c,w);
gekko_CPU_psqStore(c,d,ea,gqr,wBit);
c->GPR[gekko_ra(w)] = ea;
break;}
}}
return ;
break;}
}}
if ((gekko_xo10(w) == cast<uint32_t>(1014ULL))) {
gekko_CPU_dcbzL(c,gekko_CPU_eax(c,w));
return ;
}
if ((!gekko_CPU_psEnabled(c))) {
gekko_CPU_Halt(c,std::string("gekko: a paired-single instruction ran with the unit off (HID2 = 0x%08X) at 0x%08X",82),c->HID2,pc);
return ;
}
auto tmp68 = std::make_tuple(gekko_rs(w),gekko_ra(w),gekko_rb(w),gekko_rc(w));
uint32_t d = std::get<0>(tmp68);
uint32_t a = std::get<1>(tmp68);
uint32_t b = std::get<2>(tmp68);
uint32_t cc = std::get<3>(tmp68);
auto tmp69 = std::make_tuple(c->FPR[a],c->FPR[b],c->FPR[cc]);
gekko_FPR A = std::get<0>(tmp69);
gekko_FPR B = std::get<1>(tmp69);
gekko_FPR C = std::get<2>(tmp69);
{
switch(gekko_xo5(w)){
case cast<uint32_t>(21ULL):{
gekko_CPU_psResult(c,w,d,gekko_CPU_fadd(c,A.PS0,B.PS0),gekko_CPU_fadd(c,A.PS1,B.PS1));
return ;
break;}
case cast<uint32_t>(20ULL):{
gekko_CPU_psResult(c,w,d,gekko_CPU_fsub(c,A.PS0,B.PS0),gekko_CPU_fsub(c,A.PS1,B.PS1));
return ;
break;}
case cast<uint32_t>(25ULL):{
gekko_CPU_psResult(c,w,d,gekko_CPU_fmul(c,A.PS0,C.PS0),gekko_CPU_fmul(c,A.PS1,C.PS1));
return ;
break;}
case cast<uint32_t>(18ULL):{
gekko_CPU_psResult(c,w,d,gekko_CPU_fdiv(c,A.PS0,B.PS0),gekko_CPU_fdiv(c,A.PS1,B.PS1));
return ;
break;}
case cast<uint32_t>(12ULL):{
gekko_CPU_psResult(c,w,d,gekko_CPU_fmul(c,A.PS0,C.PS0),gekko_CPU_fmul(c,A.PS1,C.PS0));
return ;
break;}
case cast<uint32_t>(13ULL):{
gekko_CPU_psResult(c,w,d,gekko_CPU_fmul(c,A.PS0,C.PS1),gekko_CPU_fmul(c,A.PS1,C.PS1));
return ;
break;}
case cast<uint32_t>(29ULL):{
gekko_CPU_psResult(c,w,d,gekko_CPU_fmaddRaw(c,A.PS0,C.PS0,B.PS0),gekko_CPU_fmaddRaw(c,A.PS1,C.PS1,B.PS1));
return ;
break;}
case cast<uint32_t>(28ULL):{
gekko_CPU_psResult(c,w,d,gekko_CPU_fmaddRaw(c,A.PS0,C.PS0,cast<double>(-B.PS0)),gekko_CPU_fmaddRaw(c,A.PS1,C.PS1,cast<double>(-B.PS1)));
return ;
break;}
case cast<uint32_t>(31ULL):{
gekko_CPU_psResult(c,w,d,cast<double>(-gekko_CPU_fmaddRaw(c,A.PS0,C.PS0,B.PS0)),cast<double>(-gekko_CPU_fmaddRaw(c,A.PS1,C.PS1,B.PS1)));
return ;
break;}
case cast<uint32_t>(30ULL):{
gekko_CPU_psResult(c,w,d,cast<double>(-gekko_CPU_fmaddRaw(c,A.PS0,C.PS0,cast<double>(-B.PS0))),cast<double>(-gekko_CPU_fmaddRaw(c,A.PS1,C.PS1,cast<double>(-B.PS1))));
return ;
break;}
case cast<uint32_t>(14ULL):{
gekko_CPU_psResult(c,w,d,gekko_CPU_fmaddRaw(c,A.PS0,C.PS0,B.PS0),gekko_CPU_fmaddRaw(c,A.PS1,C.PS0,B.PS1));
return ;
break;}
case cast<uint32_t>(15ULL):{
gekko_CPU_psResult(c,w,d,gekko_CPU_fmaddRaw(c,A.PS0,C.PS1,B.PS0),gekko_CPU_fmaddRaw(c,A.PS1,C.PS1,B.PS1));
return ;
break;}
case cast<uint32_t>(10ULL):{
gekko_CPU_psResult(c,w,d,gekko_CPU_fadd(c,A.PS0,B.PS1),C.PS1);
return ;
break;}
case cast<uint32_t>(11ULL):{
gekko_CPU_psResult(c,w,d,C.PS0,gekko_CPU_fadd(c,A.PS0,B.PS1));
return ;
break;}
case cast<uint32_t>(23ULL):{
auto tmp70 = std::make_tuple(B.PS0,B.PS1);
double p0 = std::get<0>(tmp70);
double p1 = std::get<1>(tmp70);
if (((A.PS0 >= cast<double>(0ULL)) && (!go_math_IsNaN(A.PS0)))) {
p0 = C.PS0;
}
if (((A.PS1 >= cast<double>(0ULL)) && (!go_math_IsNaN(A.PS1)))) {
p1 = C.PS1;
}
auto tmp71 = std::make_tuple(p0,p1);
c->FPR[d].PS0 = std::get<0>(tmp71);
c->FPR[d].PS1 = std::get<1>(tmp71);
gekko_CPU_rcF(c,w);
return ;
break;}
case cast<uint32_t>(24ULL):{
gekko_CPU_psResult(c,w,d,gekko_CPU_fres(c,B.PS0),gekko_CPU_fres(c,B.PS1));
return ;
break;}
case cast<uint32_t>(26ULL):{
gekko_CPU_psResult(c,w,d,gekko_CPU_frsqrte(c,B.PS0),gekko_CPU_frsqrte(c,B.PS1));
return ;
break;}
}}
{
switch(gekko_xo10(w)){
case cast<uint32_t>(528ULL):{
c->FPR[d] = gekko_FPR{A.PS0,B.PS0};
gekko_CPU_rcF(c,w);
break;}
case cast<uint32_t>(560ULL):{
c->FPR[d] = gekko_FPR{A.PS0,B.PS1};
gekko_CPU_rcF(c,w);
break;}
case cast<uint32_t>(592ULL):{
c->FPR[d] = gekko_FPR{A.PS1,B.PS0};
gekko_CPU_rcF(c,w);
break;}
case cast<uint32_t>(624ULL):{
c->FPR[d] = gekko_FPR{A.PS1,B.PS1};
gekko_CPU_rcF(c,w);
break;}
case cast<uint32_t>(40ULL):{
c->FPR[d] = gekko_FPR{gekko_negf(B.PS0),gekko_negf(B.PS1)};
gekko_CPU_rcF(c,w);
break;}
case cast<uint32_t>(72ULL):{
c->FPR[d] = B;
gekko_CPU_rcF(c,w);
break;}
case cast<uint32_t>(136ULL):{
c->FPR[d] = gekko_FPR{cast<double>(-go_math_Abs(B.PS0)),cast<double>(-go_math_Abs(B.PS1))};
gekko_CPU_rcF(c,w);
break;}
case cast<uint32_t>(264ULL):{
c->FPR[d] = gekko_FPR{go_math_Abs(B.PS0),go_math_Abs(B.PS1)};
gekko_CPU_rcF(c,w);
break;}
case cast<uint32_t>(0ULL):{
gekko_CPU_fcmp(c,gekko_crfD(w),A.PS0,B.PS0,false);
break;}
case cast<uint32_t>(32ULL):{
gekko_CPU_fcmp(c,gekko_crfD(w),A.PS0,B.PS0,true);
break;}
case cast<uint32_t>(64ULL):{
gekko_CPU_fcmp(c,gekko_crfD(w),A.PS1,B.PS1,false);
break;}
case cast<uint32_t>(96ULL):{
gekko_CPU_fcmp(c,gekko_crfD(w),A.PS1,B.PS1,true);
break;}
default:{
gekko_CPU_Halt(c,std::string("gekko: unimplemented opcode 4 extended %d (word 0x%08X) at 0x%08X",65),gekko_xo10(w),w,pc);
break;}
}}
}
}
// tools/cpu/gekko/ps.go:368:1
double gekko_negf(double v){
{
return go_math_Float64frombits(cast<uint64_t>((go_math_Float64bits(v) ^ cast<uint64_t>(9223372036854775808ULL))));
}
}
// tools/cpu/gekko/ps.go:375:1
void gekko_CPU_psResult(gekko_CPU* c,uint32_t w,uint32_t d,double p0,double p1){
{
c->FPR[d].PS0 = gekko_f32(p0);
c->FPR[d].PS1 = gekko_f32(p1);
gekko_CPU_setFPRF(c,c->FPR[d].PS0);
gekko_CPU_rcF(c,w);
}
}
// tools/cpu/gekko/spr.go:18:1
uint32_t gekko_CPU_readSPR(gekko_CPU* c,uint32_t n,uint32_t pc){
{
{
if ((n == cast<uint32_t>(1ULL))){
return c->XER;
}
else if ((n == cast<uint32_t>(8ULL))){
return c->LR;
}
else if ((n == cast<uint32_t>(9ULL))){
return c->CTR;
}
else if ((n == cast<uint32_t>(18ULL))){
return c->DSISR;
}
else if ((n == cast<uint32_t>(19ULL))){
return c->DAR;
}
else if ((n == cast<uint32_t>(22ULL))){
return c->DEC;
}
else if ((n == cast<uint32_t>(25ULL))){
return c->SDR1;
}
else if ((n == cast<uint32_t>(26ULL))){
return c->SRR0;
}
else if ((n == cast<uint32_t>(27ULL))){
return c->SRR1;
}
else if (((n >= cast<uint32_t>(272ULL)) && (n <= cast<uint32_t>(275ULL)))){
return c->SPRG[cast<uint32_t>((n - cast<uint32_t>(272ULL)))];
}
else if ((n == cast<uint32_t>(287ULL))){
return c->PVR;
}
else if ((n == cast<uint32_t>(284ULL))){
return cast<uint32_t>(c->TB);
}
else if ((n == cast<uint32_t>(285ULL))){
return cast<uint32_t>(shr<uint64_t>(c->TB,cast<int64_t>(32ULL)));
}
else if (((n >= cast<uint32_t>(528ULL)) && (n <= cast<uint32_t>(535ULL)))){
uint32_t i = divi<uint32_t>((cast<uint32_t>((n - cast<uint32_t>(528ULL)))),cast<uint32_t>(2ULL));
return c->IBAT[i][modi<uint32_t>((cast<uint32_t>((n - cast<uint32_t>(528ULL)))),cast<uint32_t>(2ULL))];
}
else if (((n >= cast<uint32_t>(536ULL)) && (n <= cast<uint32_t>(543ULL)))){
uint32_t i = divi<uint32_t>((cast<uint32_t>((n - cast<uint32_t>(536ULL)))),cast<uint32_t>(2ULL));
return c->DBAT[i][modi<uint32_t>((cast<uint32_t>((n - cast<uint32_t>(536ULL)))),cast<uint32_t>(2ULL))];
}
else if (((n >= cast<uint32_t>(912ULL)) && (n <= cast<uint32_t>(919ULL)))){
return c->GQR[cast<uint32_t>((n - cast<uint32_t>(912ULL)))];
}
else if ((n == cast<uint32_t>(1008ULL))){
return c->HID0;
}
else if ((n == cast<uint32_t>(1009ULL))){
return c->HID1;
}
else if ((n == cast<uint32_t>(920ULL))){
return c->HID2;
}
else if ((n == cast<uint32_t>(1011ULL))){
return c->HID4;
}
else if ((n == cast<uint32_t>(921ULL))){
return c->WPAR;
}
else if ((n == cast<uint32_t>(922ULL))){
return c->DMAU;
}
else if ((n == cast<uint32_t>(923ULL))){
return c->DMAL;
}
else if ((n == cast<uint32_t>(1017ULL))){
return c->L2CR;
}
else if ((n == cast<uint32_t>(282ULL))){
return cast<uint32_t>(0ULL);
}
else if ((((n == cast<uint32_t>(1013ULL)) || (n == cast<uint32_t>(1010ULL))) || (n == cast<uint32_t>(1019ULL)))){
return cast<uint32_t>(0ULL);
}
else if (((n >= cast<uint32_t>(1020ULL)) && (n <= cast<uint32_t>(1022ULL)))){
return cast<uint32_t>(0ULL);
}
}
tmp72:;
gekko_CPU_Halt(c,std::string("gekko: mfspr from the unknown SPR %d at 0x%08X",46),n,pc);
return cast<uint32_t>(0ULL);
}
}
// tools/cpu/gekko/spr.go:83:1
void gekko_CPU_writeSPR(gekko_CPU* c,uint32_t n,uint32_t v,uint32_t pc){
{
{
if ((n == cast<uint32_t>(1ULL))){
c->XER = v;
}
else if ((n == cast<uint32_t>(8ULL))){
c->LR = v;
}
else if ((n == cast<uint32_t>(9ULL))){
c->CTR = v;
}
else if ((n == cast<uint32_t>(18ULL))){
c->DSISR = v;
}
else if ((n == cast<uint32_t>(19ULL))){
c->DAR = v;
}
else if ((n == cast<uint32_t>(22ULL))){
gekko_CPU_setDEC(c,v);
}
else if ((n == cast<uint32_t>(25ULL))){
c->SDR1 = v;
}
else if ((n == cast<uint32_t>(26ULL))){
c->SRR0 = v;
}
else if ((n == cast<uint32_t>(27ULL))){
c->SRR1 = v;
}
else if (((n >= cast<uint32_t>(272ULL)) && (n <= cast<uint32_t>(275ULL)))){
c->SPRG[cast<uint32_t>((n - cast<uint32_t>(272ULL)))] = v;
}
else if ((n == cast<uint32_t>(284ULL))){
c->TB = cast<uint64_t>((cast<uint64_t>((c->TB & cast<uint64_t>(18446744069414584320ULL))) | cast<uint64_t>(v)));
}
else if ((n == cast<uint32_t>(285ULL))){
c->TB = cast<uint64_t>((cast<uint64_t>((c->TB & cast<uint64_t>(4294967295ULL))) | shl<uint64_t>(cast<uint64_t>(v),cast<int64_t>(32ULL))));
}
else if (((n >= cast<uint32_t>(528ULL)) && (n <= cast<uint32_t>(535ULL)))){
uint32_t i = divi<uint32_t>((cast<uint32_t>((n - cast<uint32_t>(528ULL)))),cast<uint32_t>(2ULL));
c->IBAT[i][modi<uint32_t>((cast<uint32_t>((n - cast<uint32_t>(528ULL)))),cast<uint32_t>(2ULL))] = v;
}
else if (((n >= cast<uint32_t>(536ULL)) && (n <= cast<uint32_t>(543ULL)))){
uint32_t i = divi<uint32_t>((cast<uint32_t>((n - cast<uint32_t>(536ULL)))),cast<uint32_t>(2ULL));
c->DBAT[i][modi<uint32_t>((cast<uint32_t>((n - cast<uint32_t>(536ULL)))),cast<uint32_t>(2ULL))] = v;
}
else if (((n >= cast<uint32_t>(912ULL)) && (n <= cast<uint32_t>(919ULL)))){
c->GQR[cast<uint32_t>((n - cast<uint32_t>(912ULL)))] = v;
}
else if ((n == cast<uint32_t>(1008ULL))){
c->HID0 = v;
}
else if ((n == cast<uint32_t>(1009ULL))){
c->HID1 = v;
}
else if ((n == cast<uint32_t>(920ULL))){
gekko_CPU_setHID2(c,v);
}
else if ((n == cast<uint32_t>(1011ULL))){
c->HID4 = v;
}
else if ((n == cast<uint32_t>(921ULL))){
c->WPAR = v;
}
else if ((n == cast<uint32_t>(922ULL))){
c->DMAU = v;
}
else if ((n == cast<uint32_t>(923ULL))){
c->DMAL = v;
gekko_CPU_runDMA(c);
}
else if ((n == cast<uint32_t>(1017ULL))){
c->L2CR = v;
}
else if ((n == cast<uint32_t>(282ULL)) || (n == cast<uint32_t>(1013ULL)) || (n == cast<uint32_t>(1010ULL)) || (n == cast<uint32_t>(1019ULL)) || (n == cast<uint32_t>(287ULL))){
}
else if (((n >= cast<uint32_t>(1020ULL)) && (n <= cast<uint32_t>(1022ULL)))){
}
else {
gekko_CPU_Halt(c,std::string("gekko: mtspr to the unknown SPR %d (value 0x%08X) at 0x%08X",59),n,v,pc);
}
}
tmp73:;
}
}
// tools/cpu/gekko/state.go:54:1
gekko_State gekko_CPU_Snapshot(gekko_CPU* c){
{
gekko_State s = gekko_State{c->GPR,c->FPR,c->PC,c->LR,c->CTR,c->CR,c->XER,c->MSR,c->FPSCR,c->GQR,c->HID0,c->HID1,c->HID2,c->HID4,c->WPAR,c->DMAU,c->DMAL,c->L2CR,c->SRR0,c->SRR1,c->SPRG,c->DSISR,c->DAR,c->SDR1,c->PVR,c->SR,c->IBAT,c->DBAT,c->TB,c->DEC,c->clockFrac,c->decArmed,c->Reserved,c->ReserveAddr,{},c->LC.Base,c->LC.Enabled_,c->ExtInt,c->Steps,c->Halted,c->HaltReason};
s.LCData = Slice<uint8_t>::make(cast<int64_t>(16384ULL));
gcopy(s.LCData,sub(c->LC.Data,0,len(c->LC.Data)));
return s;
}
}
// tools/cpu/gekko/state.go:80:1
void gekko_CPU_Restore(gekko_CPU* c,gekko_State s){
{
auto tmp74 = std::make_tuple(s.GPR,s.FPR);
c->GPR = std::get<0>(tmp74);
c->FPR = std::get<1>(tmp74);
auto tmp75 = std::make_tuple(s.PC,s.LR,s.CTR);
c->PC = std::get<0>(tmp75);
c->LR = std::get<1>(tmp75);
c->CTR = std::get<2>(tmp75);
auto tmp76 = std::make_tuple(s.CR,s.XER,s.MSR,s.FPSCR);
c->CR = std::get<0>(tmp76);
c->XER = std::get<1>(tmp76);
c->MSR = std::get<2>(tmp76);
c->FPSCR = std::get<3>(tmp76);
auto tmp77 = std::make_tuple(s.GQR,s.HID0,s.HID1,s.HID2,s.HID4);
c->GQR = std::get<0>(tmp77);
c->HID0 = std::get<1>(tmp77);
c->HID1 = std::get<2>(tmp77);
c->HID2 = std::get<3>(tmp77);
c->HID4 = std::get<4>(tmp77);
auto tmp78 = std::make_tuple(s.WPAR,s.DMAU,s.DMAL,s.L2CR);
c->WPAR = std::get<0>(tmp78);
c->DMAU = std::get<1>(tmp78);
c->DMAL = std::get<2>(tmp78);
c->L2CR = std::get<3>(tmp78);
auto tmp79 = std::make_tuple(s.SRR0,s.SRR1,s.SPRG);
c->SRR0 = std::get<0>(tmp79);
c->SRR1 = std::get<1>(tmp79);
c->SPRG = std::get<2>(tmp79);
auto tmp80 = std::make_tuple(s.DSISR,s.DAR,s.SDR1,s.PVR);
c->DSISR = std::get<0>(tmp80);
c->DAR = std::get<1>(tmp80);
c->SDR1 = std::get<2>(tmp80);
c->PVR = std::get<3>(tmp80);
auto tmp81 = std::make_tuple(s.SR,s.IBAT,s.DBAT);
c->SR = std::get<0>(tmp81);
c->IBAT = std::get<1>(tmp81);
c->DBAT = std::get<2>(tmp81);
auto tmp82 = std::make_tuple(s.TB,s.DEC,s.ClockFrac,s.DecArmed);
c->TB = std::get<0>(tmp82);
c->DEC = std::get<1>(tmp82);
c->clockFrac = std::get<2>(tmp82);
c->decArmed = std::get<3>(tmp82);
auto tmp83 = std::make_tuple(s.Reserved,s.ReserveAddr);
c->Reserved = std::get<0>(tmp83);
c->ReserveAddr = std::get<1>(tmp83);
auto tmp84 = std::make_tuple(s.LCBase,s.LCEnabled);
c->LC.Base = std::get<0>(tmp84);
c->LC.Enabled_ = std::get<1>(tmp84);
gcopy(sub(c->LC.Data,0,len(c->LC.Data)),s.LCData);
auto tmp85 = std::make_tuple(s.ExtInt,s.Steps);
c->ExtInt = std::get<0>(tmp85);
c->Steps = std::get<1>(tmp85);
auto tmp86 = std::make_tuple(s.Halted,s.HaltReason);
c->Halted = std::get<0>(tmp86);
c->HaltReason = std::get<1>(tmp86);
}
}
// tools/cpu/gekko/timer.go:33:1
void gekko_CPU_tick(gekko_CPU* c,int64_t cycles){
{
c->clockFrac += cast<uint32_t>(cycles);
{;for (;(c->clockFrac >= cast<uint32_t>(12ULL));){
c->clockFrac -= cast<uint32_t>(12ULL);
c->TB++;
uint32_t prev = c->DEC;
c->DEC--;
if (((cast<uint32_t>((prev & cast<uint32_t>(2147483648ULL))) == cast<uint32_t>(0ULL)) && (cast<uint32_t>((c->DEC & cast<uint32_t>(2147483648ULL))) != cast<uint32_t>(0ULL)))) {
c->decArmed = true;
}
}
}}
}
// tools/cpu/gekko/timer.go:51:1
void gekko_CPU_setDEC(gekko_CPU* c,uint32_t v){
{
c->DEC = v;
c->decArmed = false;
}
}
// tools/cpu/gekko/timer.go:67:1
uint64_t gekko_CPU_InstrsToDecUnderflow(gekko_CPU* c){
{
if ((c->decArmed || (cast<uint32_t>((c->DEC & cast<uint32_t>(2147483648ULL))) != cast<uint32_t>(0ULL)))) {
return cast<uint64_t>(18446744073709551615ULL);
}
uint64_t ticks = cast<uint64_t>((cast<uint64_t>(c->DEC) + cast<uint64_t>(1ULL)));
return cast<uint64_t>((cast<uint64_t>((ticks * cast<uint64_t>(12ULL))) - cast<uint64_t>(c->clockFrac)));
}
}
// tools/cpu/gekko/timer.go:85:1
void gekko_CPU_SkipInstructions(gekko_CPU* c,uint64_t n){
{
c->Steps += n;
uint64_t total = cast<uint64_t>((cast<uint64_t>(c->clockFrac) + n));
uint64_t ticks = divi<uint64_t>(total,cast<uint64_t>(12ULL));
c->clockFrac = cast<uint32_t>(modi<uint64_t>(total,cast<uint64_t>(12ULL)));
c->TB += ticks;
c->DEC -= cast<uint32_t>(ticks);
}
}
// tools/cpu/gcdsp/cpu.go:119:1
gcdsp_CPU* gcdsp_New(gc_dspBus bus){
{
return arenaNew(gcdsp_CPU{{},{},{},{},{},{},bus,{},{},{},{},{},{}});
}
}
// tools/cpu/gcdsp/cpu.go:124:1
void gcdsp_CPU_SetBus(gcdsp_CPU* c,gc_dspBus bus){
{
c->bus = bus;
}
}
// tools/cpu/gcdsp/cpu.go:157:1
int64_t gcdsp_CPU_ac(gcdsp_CPU* c,int64_t n){
{
int64_t h = cast<int64_t>(cast<int16_t>(c->Reg[cast<int64_t>((cast<int64_t>(16ULL) + n))]));
int64_t m = cast<int64_t>(c->Reg[cast<int64_t>((cast<int64_t>(30ULL) + n))]);
int64_t l = cast<int64_t>(c->Reg[cast<int64_t>((cast<int64_t>(28ULL) + n))]);
return cast<int64_t>((cast<int64_t>(((shl<int64_t>(h,cast<int64_t>(32ULL))) | (shl<int64_t>(m,cast<int64_t>(16ULL))))) | l));
}
}
// tools/cpu/gcdsp/cpu.go:166:1
void gcdsp_CPU_setAc(gcdsp_CPU* c,int64_t n,int64_t v){
{
c->Reg[cast<int64_t>((cast<int64_t>(28ULL) + n))] = cast<uint16_t>(v);
c->Reg[cast<int64_t>((cast<int64_t>(30ULL) + n))] = cast<uint16_t>(shr<int64_t>(v,cast<int64_t>(16ULL)));
c->Reg[cast<int64_t>((cast<int64_t>(16ULL) + n))] = cast<uint16_t>(shr<int64_t>(v,cast<int64_t>(32ULL)));
if ((cast<int64_t>((v & cast<int64_t>(549755813888ULL))) != cast<int64_t>(0ULL))) {
c->Reg[cast<int64_t>((cast<int64_t>(16ULL) + n))] |= cast<uint16_t>(65280ULL);
}
else {
c->Reg[cast<int64_t>((cast<int64_t>(16ULL) + n))] &= cast<uint16_t>(255ULL);
}
}
}
// tools/cpu/gcdsp/cpu.go:179:1
int64_t gcdsp_CPU_ax(gcdsp_CPU* c,int64_t n){
{
int64_t h = cast<int64_t>(cast<int16_t>(c->Reg[cast<int64_t>((cast<int64_t>(26ULL) + n))]));
int64_t l = cast<int64_t>(c->Reg[cast<int64_t>((cast<int64_t>(24ULL) + n))]);
return cast<int64_t>(((shl<int64_t>(h,cast<int64_t>(16ULL))) | l));
}
}
// tools/cpu/gcdsp/cpu.go:188:1
int64_t gcdsp_CPU_prod(gcdsp_CPU* c){
{
int64_t p = shl<int64_t>(cast<int64_t>(cast<int8_t>(c->Reg[cast<int64_t>(22ULL)])),cast<int64_t>(32ULL));
p += shl<int64_t>((cast<int64_t>((cast<int64_t>(c->Reg[cast<int64_t>(21ULL)]) + cast<int64_t>(c->Reg[cast<int64_t>(23ULL)])))),cast<int64_t>(16ULL));
p += cast<int64_t>(c->Reg[cast<int64_t>(20ULL)]);
return p;
}
}
// tools/cpu/gcdsp/cpu.go:197:1
int64_t gcdsp_CPU_prodRounded(gcdsp_CPU* c){
{
int64_t p = gcdsp_CPU_prod(c);
if ((cast<int64_t>((p & cast<int64_t>(65536ULL))) != cast<int64_t>(0ULL))) {
return ((cast<int64_t>((p + cast<int64_t>(32768ULL)))) & ~(cast<int64_t>(65535ULL)));
}
return ((cast<int64_t>((p + cast<int64_t>(32767ULL)))) & ~(cast<int64_t>(65535ULL)));
}
}
// tools/cpu/gcdsp/cpu.go:206:1
uint16_t gcdsp_CPU_sr(gcdsp_CPU* c){
{
return c->Reg[cast<int64_t>(19ULL)];
}
}
// tools/cpu/gcdsp/cpu.go:207:1
void gcdsp_CPU_setSR(gcdsp_CPU* c,uint16_t v){
{
c->Reg[cast<int64_t>(19ULL)] = v;
}
}
// tools/cpu/gcdsp/cpu.go:208:1
void gcdsp_CPU_setFlag(gcdsp_CPU* c,uint16_t bit,bool on){
{
if (on) {
c->Reg[cast<int64_t>(19ULL)] |= bit;
}
else {
c->Reg[cast<int64_t>(19ULL)] &= ~(bit);
}
}
}
// tools/cpu/gcdsp/disasm.go:21:1
std::string gcdsp_regName(uint16_t r){
{
{
if ((r < cast<uint16_t>(4ULL))){
return go_fmt_Sprintf(std::string("ar%d",4),r);
}
else if ((r < cast<uint16_t>(8ULL))){
return go_fmt_Sprintf(std::string("ix%d",4),cast<uint16_t>((r - cast<uint16_t>(4ULL))));
}
else if ((r < cast<uint16_t>(12ULL))){
return go_fmt_Sprintf(std::string("wr%d",4),cast<uint16_t>((r - cast<uint16_t>(8ULL))));
}
else if ((r < cast<uint16_t>(16ULL))){
return go_fmt_Sprintf(std::string("st%d",4),cast<uint16_t>((r - cast<uint16_t>(12ULL))));
}
}
tmp1:;
{
switch(r){
case cast<uint16_t>(16ULL):{
return std::string("ac0.h",5);
break;}
case cast<uint16_t>(17ULL):{
return std::string("ac1.h",5);
break;}
case cast<uint16_t>(18ULL):{
return std::string("config",6);
break;}
case cast<uint16_t>(19ULL):{
return std::string("sr",2);
break;}
case cast<uint16_t>(20ULL):{
return std::string("prod.l",6);
break;}
case cast<uint16_t>(21ULL):{
return std::string("prod.m1",7);
break;}
case cast<uint16_t>(22ULL):{
return std::string("prod.h",6);
break;}
case cast<uint16_t>(23ULL):{
return std::string("prod.m2",7);
break;}
case cast<uint16_t>(24ULL):{
return std::string("ax0.l",5);
break;}
case cast<uint16_t>(25ULL):{
return std::string("ax1.l",5);
break;}
case cast<uint16_t>(26ULL):{
return std::string("ax0.h",5);
break;}
case cast<uint16_t>(27ULL):{
return std::string("ax1.h",5);
break;}
case cast<uint16_t>(28ULL):{
return std::string("ac0.l",5);
break;}
case cast<uint16_t>(29ULL):{
return std::string("ac1.l",5);
break;}
case cast<uint16_t>(30ULL):{
return std::string("ac0.m",5);
break;}
case cast<uint16_t>(31ULL):{
return std::string("ac1.m",5);
break;}
}}
return go_fmt_Sprintf(std::string("r%d",3),r);
}
}
// tools/cpu/gcdsp/disasm.go:70:1
std::string gcdsp_lowHigh(uint16_t bit){
{
if ((bit != cast<uint16_t>(0ULL))) {
return std::string("h",1);
}
return std::string("l",1);
}
}
// tools/cpu/gcdsp/disasm.go:79:1
std::string gcdsp_condName(uint16_t cc){
{
{
switch(cc){
case cast<uint16_t>(0ULL):{
return std::string("ge",2);
break;}
case cast<uint16_t>(1ULL):{
return std::string("l",1);
break;}
case cast<uint16_t>(2ULL):{
return std::string("g",1);
break;}
case cast<uint16_t>(3ULL):{
return std::string("le",2);
break;}
case cast<uint16_t>(4ULL):{
return std::string("nz",2);
break;}
case cast<uint16_t>(5ULL):{
return std::string("z",1);
break;}
case cast<uint16_t>(6ULL):{
return std::string("nc",2);
break;}
case cast<uint16_t>(7ULL):{
return std::string("c",1);
break;}
case cast<uint16_t>(12ULL):{
return std::string("lnz",3);
break;}
case cast<uint16_t>(13ULL):{
return std::string("lz",2);
break;}
case cast<uint16_t>(14ULL):{
return std::string("o",1);
break;}
case cast<uint16_t>(15ULL):{
return std::string("",0);
break;}
}}
return go_fmt_Sprintf(std::string("?%X",3),cc);
}
}
// tools/cpu/gcdsp/disasm.go:112:1
std::tuple<std::string,uint16_t> gcdsp_Disasm(std::function<uint16_t(uint16_t)> read,uint16_t pc){
std::string text{};
uint16_t words{};
{
uint16_t op = read(pc);
auto next = [&]()->uint16_t{
return read(cast<uint16_t>((pc + cast<uint16_t>(1ULL))));
}
;
auto cond = [&](std::string base,uint16_t cc)->std::string{
return (base + gcdsp_condName(cc));
}
;
{
if ((op == cast<uint16_t>(0ULL))){
return {std::string("nop",3),cast<uint16_t>(1ULL)};
}
else if ((op == cast<uint16_t>(33ULL))){
return {std::string("halt",4),cast<uint16_t>(1ULL)};
}
else if ((cast<uint16_t>((op & cast<uint16_t>(65532ULL))) == cast<uint16_t>(4ULL))){
return {go_fmt_Sprintf(std::string("dar    %s",9),gcdsp_regName(cast<uint16_t>((op & cast<uint16_t>(3ULL))))),cast<uint16_t>(1ULL)};
}
else if ((cast<uint16_t>((op & cast<uint16_t>(65532ULL))) == cast<uint16_t>(8ULL))){
return {go_fmt_Sprintf(std::string("iar    %s",9),gcdsp_regName(cast<uint16_t>((op & cast<uint16_t>(3ULL))))),cast<uint16_t>(1ULL)};
}
else if ((cast<uint16_t>((op & cast<uint16_t>(65532ULL))) == cast<uint16_t>(12ULL))){
return {go_fmt_Sprintf(std::string("subarn %s",9),gcdsp_regName(cast<uint16_t>((op & cast<uint16_t>(3ULL))))),cast<uint16_t>(1ULL)};
}
else if ((cast<uint16_t>((op & cast<uint16_t>(65520ULL))) == cast<uint16_t>(16ULL))){
return {go_fmt_Sprintf(std::string("addarn %s, %s",13),gcdsp_regName(cast<uint16_t>((op & cast<uint16_t>(3ULL)))),gcdsp_regName(cast<uint16_t>((cast<uint16_t>(4ULL) + (cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(2ULL))) & cast<uint16_t>(3ULL)))))))),cast<uint16_t>(1ULL)};
}
else if ((cast<uint16_t>((op & cast<uint16_t>(65504ULL))) == cast<uint16_t>(64ULL))){
return {go_fmt_Sprintf(std::string("loop   %s",9),gcdsp_regName(cast<uint16_t>((op & cast<uint16_t>(31ULL))))),cast<uint16_t>(1ULL)};
}
else if ((cast<uint16_t>((op & cast<uint16_t>(65504ULL))) == cast<uint16_t>(96ULL))){
return {go_fmt_Sprintf(std::string("bloop  %s, 0x%04X",17),gcdsp_regName(cast<uint16_t>((op & cast<uint16_t>(31ULL)))),next()),cast<uint16_t>(2ULL)};
}
else if ((cast<uint16_t>((op & cast<uint16_t>(65280ULL))) == cast<uint16_t>(4096ULL))){
return {go_fmt_Sprintf(std::string("loopi  #0x%02X",14),cast<uint16_t>((op & cast<uint16_t>(255ULL)))),cast<uint16_t>(1ULL)};
}
else if ((cast<uint16_t>((op & cast<uint16_t>(65280ULL))) == cast<uint16_t>(4352ULL))){
return {go_fmt_Sprintf(std::string("bloopi #0x%02X, 0x%04X",22),cast<uint16_t>((op & cast<uint16_t>(255ULL))),next()),cast<uint16_t>(2ULL)};
}
else if ((cast<uint16_t>((op & cast<uint16_t>(65280ULL))) == cast<uint16_t>(4608ULL))){
return {go_fmt_Sprintf(std::string("sbclr  #%d",10),cast<uint16_t>((op & cast<uint16_t>(255ULL)))),cast<uint16_t>(1ULL)};
}
else if ((cast<uint16_t>((op & cast<uint16_t>(65280ULL))) == cast<uint16_t>(4864ULL))){
return {go_fmt_Sprintf(std::string("sbset  #%d",10),cast<uint16_t>((op & cast<uint16_t>(255ULL)))),cast<uint16_t>(1ULL)};
}
else if ((cast<uint16_t>((op & cast<uint16_t>(65504ULL))) == cast<uint16_t>(128ULL))){
return {go_fmt_Sprintf(std::string("lri    %s, #0x%04X",18),gcdsp_regName(cast<uint16_t>((op & cast<uint16_t>(31ULL)))),next()),cast<uint16_t>(2ULL)};
}
else if ((cast<uint16_t>((op & cast<uint16_t>(65504ULL))) == cast<uint16_t>(160ULL))){
return {go_fmt_Sprintf(std::string("lrr?   %s <- (0x%04X)",21),gcdsp_regName(cast<uint16_t>((op & cast<uint16_t>(31ULL)))),next()),cast<uint16_t>(2ULL)};
}
else if ((cast<uint16_t>((op & cast<uint16_t>(65504ULL))) == cast<uint16_t>(192ULL))){
return {go_fmt_Sprintf(std::string("lr     %s, @0x%04X",18),gcdsp_regName(cast<uint16_t>((op & cast<uint16_t>(31ULL)))),next()),cast<uint16_t>(2ULL)};
}
else if ((cast<uint16_t>((op & cast<uint16_t>(65504ULL))) == cast<uint16_t>(224ULL))){
return {go_fmt_Sprintf(std::string("sr     @0x%04X, %s",18),next(),gcdsp_regName(cast<uint16_t>((op & cast<uint16_t>(31ULL))))),cast<uint16_t>(2ULL)};
}
else if ((cast<uint16_t>((op & cast<uint16_t>(65280ULL))) == cast<uint16_t>(5632ULL))){
uint16_t si = cast<uint16_t>((op & cast<uint16_t>(255ULL)));
if ((si >= cast<uint16_t>(128ULL))) {
si |= cast<uint16_t>(65280ULL);
}
return {go_fmt_Sprintf(std::string("si     @0x%04X, #0x%04X",23),si,next()),cast<uint16_t>(2ULL)};
}
else if ((cast<uint16_t>((op & cast<uint16_t>(64512ULL))) == cast<uint16_t>(6144ULL))){
std::string ar = gcdsp_regName(cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(5ULL))) & cast<uint16_t>(3ULL))));
std::string reg = gcdsp_regName(cast<uint16_t>((op & cast<uint16_t>(31ULL))));
bool store = (cast<uint16_t>((op & cast<uint16_t>(512ULL))) != cast<uint16_t>(0ULL));
std::string suffix={};
{
switch(cast<uint16_t>((op & cast<uint16_t>(384ULL)))){
case cast<uint16_t>(128ULL):{
suffix = std::string("d",1);
break;}
case cast<uint16_t>(256ULL):{
suffix = std::string("i",1);
break;}
case cast<uint16_t>(384ULL):{
suffix = std::string("n",1);
break;}
}}
if (store) {
return {go_fmt_Sprintf(std::string("%-6s @%s, %s",12),(std::string("srr",3) + suffix),ar,reg),cast<uint16_t>(1ULL)};
}
return {go_fmt_Sprintf(std::string("%-6s %s, @%s",12),(std::string("lrr",3) + suffix),reg,ar),cast<uint16_t>(1ULL)};
}
else if ((cast<uint16_t>((op & cast<uint16_t>(64512ULL))) == cast<uint16_t>(7168ULL))){
return {go_fmt_Sprintf(std::string("mrr    %s, %s",13),gcdsp_regName(cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(5ULL))) & cast<uint16_t>(31ULL)))),gcdsp_regName(cast<uint16_t>((op & cast<uint16_t>(31ULL))))),cast<uint16_t>(1ULL)};
}
else if ((cast<uint16_t>((op & cast<uint16_t>(65024ULL))) == cast<uint16_t>(5120ULL))){
uint16_t r = cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(8ULL))) & cast<uint16_t>(1ULL)));
bool arith = (cast<uint16_t>((op & cast<uint16_t>(128ULL))) != cast<uint16_t>(0ULL));
int64_t amt = cast<int64_t>(cast<uint16_t>((op & cast<uint16_t>(127ULL))));
if ((cast<int64_t>((amt & cast<int64_t>(64ULL))) != cast<int64_t>(0ULL))) {
amt -= cast<int64_t>(128ULL);
}
std::string mn={};
{
if ((arith && (amt >= cast<int64_t>(0ULL)))){
mn = std::string("asl",3);
}
else if (arith){
mn = std::string("asr",3);
amt = cast<int64_t>(-amt);
}
else if ((amt >= cast<int64_t>(0ULL))){
mn = std::string("lsl",3);
}
else {
mn = std::string("lsr",3);
amt = cast<int64_t>(-amt);
}
}
tmp3:;
return {go_fmt_Sprintf(std::string("%-6s ac%d, #%d",14),mn,r,amt),cast<uint16_t>(1ULL)};
}
else if ((cast<uint16_t>((op & cast<uint16_t>(63488ULL))) == cast<uint16_t>(2048ULL))){
return {go_fmt_Sprintf(std::string("lris   %s, #0x%02X",18),gcdsp_regName(cast<uint16_t>((cast<uint16_t>(24ULL) + (cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(8ULL))) & cast<uint16_t>(7ULL))))))),cast<uint16_t>((op & cast<uint16_t>(255ULL)))),cast<uint16_t>(1ULL)};
}
else if ((cast<uint16_t>((op & cast<uint16_t>(63488ULL))) == cast<uint16_t>(8192ULL))){
return {go_fmt_Sprintf(std::string("lrs    %s, @CR:0x%02X",21),gcdsp_regName(cast<uint16_t>((cast<uint16_t>(24ULL) + (cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(8ULL))) & cast<uint16_t>(7ULL))))))),cast<uint16_t>((op & cast<uint16_t>(255ULL)))),cast<uint16_t>(1ULL)};
}
else if ((cast<uint16_t>((op & cast<uint16_t>(63488ULL))) == cast<uint16_t>(10240ULL))){
return {go_fmt_Sprintf(std::string("srs    @CR:0x%02X, %s",21),cast<uint16_t>((op & cast<uint16_t>(255ULL))),gcdsp_regName(cast<uint16_t>((cast<uint16_t>(24ULL) + (cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(8ULL))) & cast<uint16_t>(7ULL)))))))),cast<uint16_t>(1ULL)};
}
else if ((cast<uint16_t>((op & cast<uint16_t>(65279ULL))) == cast<uint16_t>(512ULL))){
return {go_fmt_Sprintf(std::string("addi   ac%d, #0x%04X",20),cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(8ULL))) & cast<uint16_t>(1ULL))),next()),cast<uint16_t>(2ULL)};
}
else if ((cast<uint16_t>((op & cast<uint16_t>(65279ULL))) == cast<uint16_t>(544ULL))){
return {go_fmt_Sprintf(std::string("xori   ac%d.m, #0x%04X",22),cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(8ULL))) & cast<uint16_t>(1ULL))),next()),cast<uint16_t>(2ULL)};
}
else if ((cast<uint16_t>((op & cast<uint16_t>(65279ULL))) == cast<uint16_t>(576ULL))){
return {go_fmt_Sprintf(std::string("andi   ac%d.m, #0x%04X",22),cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(8ULL))) & cast<uint16_t>(1ULL))),next()),cast<uint16_t>(2ULL)};
}
else if ((cast<uint16_t>((op & cast<uint16_t>(65279ULL))) == cast<uint16_t>(608ULL))){
return {go_fmt_Sprintf(std::string("ori    ac%d.m, #0x%04X",22),cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(8ULL))) & cast<uint16_t>(1ULL))),next()),cast<uint16_t>(2ULL)};
}
else if ((cast<uint16_t>((op & cast<uint16_t>(65279ULL))) == cast<uint16_t>(640ULL))){
return {go_fmt_Sprintf(std::string("cmpi   ac%d, #0x%04X",20),cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(8ULL))) & cast<uint16_t>(1ULL))),next()),cast<uint16_t>(2ULL)};
}
else if ((cast<uint16_t>((op & cast<uint16_t>(65279ULL))) == cast<uint16_t>(672ULL))){
return {go_fmt_Sprintf(std::string("andf   ac%d.m, #0x%04X",22),cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(8ULL))) & cast<uint16_t>(1ULL))),next()),cast<uint16_t>(2ULL)};
}
else if ((cast<uint16_t>((op & cast<uint16_t>(65279ULL))) == cast<uint16_t>(704ULL))){
return {go_fmt_Sprintf(std::string("andcf  ac%d.m, #0x%04X",22),cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(8ULL))) & cast<uint16_t>(1ULL))),next()),cast<uint16_t>(2ULL)};
}
else if ((op == cast<uint16_t>(714ULL))){
return {std::string("lsrn   ac0, ac1.m",17),cast<uint16_t>(1ULL)};
}
else if ((op == cast<uint16_t>(715ULL))){
return {std::string("asrn   ac0, ac1.m",17),cast<uint16_t>(1ULL)};
}
else if ((cast<uint16_t>((op & cast<uint16_t>(65520ULL))) == cast<uint16_t>(624ULL))){
return {cond(std::string("if",2),cast<uint16_t>((op & cast<uint16_t>(15ULL)))),cast<uint16_t>(1ULL)};
}
else if ((cast<uint16_t>((op & cast<uint16_t>(65520ULL))) == cast<uint16_t>(656ULL))){
return {go_fmt_Sprintf(std::string("%-6s 0x%04X",11),cond(std::string("jmp",3),cast<uint16_t>((op & cast<uint16_t>(15ULL)))),next()),cast<uint16_t>(2ULL)};
}
else if ((cast<uint16_t>((op & cast<uint16_t>(65520ULL))) == cast<uint16_t>(688ULL))){
return {go_fmt_Sprintf(std::string("%-6s 0x%04X",11),cond(std::string("call",4),cast<uint16_t>((op & cast<uint16_t>(15ULL)))),next()),cast<uint16_t>(2ULL)};
}
else if ((cast<uint16_t>((op & cast<uint16_t>(65520ULL))) == cast<uint16_t>(720ULL))){
return {cond(std::string("ret",3),cast<uint16_t>((op & cast<uint16_t>(15ULL)))),cast<uint16_t>(1ULL)};
}
else if ((cast<uint16_t>((op & cast<uint16_t>(65520ULL))) == cast<uint16_t>(752ULL))){
return {cond(std::string("rti",3),cast<uint16_t>((op & cast<uint16_t>(15ULL)))),cast<uint16_t>(1ULL)};
}
else if ((cast<uint16_t>((op & cast<uint16_t>(65280ULL))) == cast<uint16_t>(5888ULL))){
std::string mn = std::string("jmpr",4);
if ((cast<uint16_t>((op & cast<uint16_t>(16ULL))) != cast<uint16_t>(0ULL))) {
mn = std::string("callr",5);
}
return {go_fmt_Sprintf(std::string("%-6s %s",7),cond(mn,cast<uint16_t>((op & cast<uint16_t>(15ULL)))),gcdsp_regName(cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(5ULL))) & cast<uint16_t>(7ULL))))),cast<uint16_t>(1ULL)};
}
else if ((op == cast<uint16_t>(4609ULL))){
}
else if ((cast<uint16_t>((op & cast<uint16_t>(65024ULL))) == cast<uint16_t>(1024ULL))){
return {go_fmt_Sprintf(std::string("addis  ac%d, #0x%02X",20),cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(8ULL))) & cast<uint16_t>(1ULL))),cast<uint16_t>((op & cast<uint16_t>(255ULL)))),cast<uint16_t>(1ULL)};
}
else if ((cast<uint16_t>((op & cast<uint16_t>(65024ULL))) == cast<uint16_t>(1536ULL))){
return {go_fmt_Sprintf(std::string("cmpis  ac%d, #0x%02X",20),cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(8ULL))) & cast<uint16_t>(1ULL))),cast<uint16_t>((op & cast<uint16_t>(255ULL)))),cast<uint16_t>(1ULL)};
}
}
tmp2:;
if ((op >= cast<uint16_t>(12288ULL))) {
std::string main = gcdsp_arithMnemonic(op);
uint16_t ext = cast<uint16_t>((op & cast<uint16_t>(255ULL)));
if ((cast<uint16_t>((op & cast<uint16_t>(61440ULL))) == cast<uint16_t>(12288ULL))) {
ext = cast<uint16_t>((op & cast<uint16_t>(127ULL)));
}
if ((ext != cast<uint16_t>(0ULL))) {
return {go_fmt_Sprintf(std::string("%-18s : %s",10),main,gcdsp_extMnemonic(ext)),cast<uint16_t>(1ULL)};
}
return {main,cast<uint16_t>(1ULL)};
}
return {go_fmt_Sprintf(std::string(".word  0x%04X",13),op),cast<uint16_t>(1ULL)};
}
}
// tools/cpu/gcdsp/disasm.go:298:1
std::string gcdsp_arithMnemonic(uint16_t op){
{
uint16_t d = cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(8ULL))) & cast<uint16_t>(1ULL)));
{
if ((cast<uint16_t>((op & cast<uint16_t>(65280ULL))) == cast<uint16_t>(35328ULL))){
return std::string("m2",2);
}
else if ((cast<uint16_t>((op & cast<uint16_t>(65280ULL))) == cast<uint16_t>(35584ULL))){
return std::string("m0",2);
}
else if ((cast<uint16_t>((op & cast<uint16_t>(65280ULL))) == cast<uint16_t>(35840ULL))){
return std::string("clr15",5);
}
else if ((cast<uint16_t>((op & cast<uint16_t>(65280ULL))) == cast<uint16_t>(36096ULL))){
return std::string("set15",5);
}
else if ((cast<uint16_t>((op & cast<uint16_t>(65280ULL))) == cast<uint16_t>(36352ULL))){
return std::string("set16",5);
}
else if ((cast<uint16_t>((op & cast<uint16_t>(65280ULL))) == cast<uint16_t>(36608ULL))){
return std::string("set40",5);
}
else if ((cast<uint16_t>((op & cast<uint16_t>(65280ULL))) == cast<uint16_t>(32768ULL))){
return std::string("nx",2);
}
else if ((cast<uint16_t>((op & cast<uint16_t>(65280ULL))) == cast<uint16_t>(33792ULL))){
return std::string("clrp",4);
}
else if ((cast<uint16_t>((op & cast<uint16_t>(65152ULL))) == cast<uint16_t>(12928ULL))){
return go_fmt_Sprintf(std::string("not    ac%d.m",13),d);
}
else if ((cast<uint16_t>((op & cast<uint16_t>(64640ULL))) == cast<uint16_t>(12288ULL))){
return go_fmt_Sprintf(std::string("xorr   ac%d.m, ax%d.h",21),d,cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(9ULL))) & cast<uint16_t>(1ULL))));
}
else if ((cast<uint16_t>((op & cast<uint16_t>(64640ULL))) == cast<uint16_t>(13312ULL))){
return go_fmt_Sprintf(std::string("andr   ac%d.m, ax%d.h",21),d,cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(9ULL))) & cast<uint16_t>(1ULL))));
}
else if ((cast<uint16_t>((op & cast<uint16_t>(64640ULL))) == cast<uint16_t>(14336ULL))){
return go_fmt_Sprintf(std::string("orr    ac%d.m, ax%d.h",21),d,cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(9ULL))) & cast<uint16_t>(1ULL))));
}
else if ((cast<uint16_t>((op & cast<uint16_t>(65152ULL))) == cast<uint16_t>(12416ULL))){
return go_fmt_Sprintf(std::string("xorc   ac%d.m",13),d);
}
else if ((cast<uint16_t>((op & cast<uint16_t>(64640ULL))) == cast<uint16_t>(13440ULL))){
return go_fmt_Sprintf(std::string("lsrnrx ac%d, ax%d.h",19),d,cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(9ULL))) & cast<uint16_t>(1ULL))));
}
else if ((cast<uint16_t>((op & cast<uint16_t>(64640ULL))) == cast<uint16_t>(14464ULL))){
return go_fmt_Sprintf(std::string("asrnrx ac%d, ax%d.h",19),d,cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(9ULL))) & cast<uint16_t>(1ULL))));
}
else if ((cast<uint16_t>((op & cast<uint16_t>(65152ULL))) == cast<uint16_t>(15488ULL))){
return go_fmt_Sprintf(std::string("lsrnr  ac%d",11),d);
}
else if ((cast<uint16_t>((op & cast<uint16_t>(65152ULL))) == cast<uint16_t>(16000ULL))){
return go_fmt_Sprintf(std::string("asrnr  ac%d",11),d);
}
else if ((cast<uint16_t>((op & cast<uint16_t>(63232ULL))) == cast<uint16_t>(33024ULL))){
return go_fmt_Sprintf(std::string("clr    ac%d",11),cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(11ULL))) & cast<uint16_t>(1ULL))));
}
else if ((cast<uint16_t>((op & cast<uint16_t>(65280ULL))) == cast<uint16_t>(33280ULL))){
return std::string("cmp",3);
}
else if ((cast<uint16_t>((op & cast<uint16_t>(65024ULL))) == cast<uint16_t>(34304ULL))){
return go_fmt_Sprintf(std::string("tstaxh ax%d.h",13),d);
}
else if ((cast<uint16_t>((op & cast<uint16_t>(63232ULL))) == cast<uint16_t>(45312ULL))){
return go_fmt_Sprintf(std::string("tst    ac%d",11),cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(11ULL))) & cast<uint16_t>(1ULL))));
}
else if ((cast<uint16_t>((op & cast<uint16_t>(59136ULL))) == cast<uint16_t>(49408ULL))){
return go_fmt_Sprintf(std::string("cmpaxh ac%d, ax%d.h",19),cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(11ULL))) & cast<uint16_t>(1ULL))),cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(12ULL))) & cast<uint16_t>(1ULL))));
}
else if ((cast<uint16_t>((op & cast<uint16_t>(63232ULL))) == cast<uint16_t>(37120ULL))){
return go_fmt_Sprintf(std::string("asr16  ac%d",11),cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(11ULL))) & cast<uint16_t>(1ULL))));
}
else if ((cast<uint16_t>((op & cast<uint16_t>(63232ULL))) == cast<uint16_t>(36864ULL))){
return go_fmt_Sprintf(std::string("mul    ax%d",11),cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(11ULL))) & cast<uint16_t>(1ULL))));
}
else if ((cast<uint16_t>((op & cast<uint16_t>(62976ULL))) == cast<uint16_t>(37376ULL))){
return go_fmt_Sprintf(std::string("mulmvz ac%d",11),d);
}
else if ((cast<uint16_t>((op & cast<uint16_t>(62976ULL))) == cast<uint16_t>(37888ULL))){
return go_fmt_Sprintf(std::string("mulac  ac%d",11),d);
}
else if ((cast<uint16_t>((op & cast<uint16_t>(62976ULL))) == cast<uint16_t>(38400ULL))){
return go_fmt_Sprintf(std::string("mulmv  ac%d",11),d);
}
else if ((cast<uint16_t>((op & cast<uint16_t>(59136ULL))) == cast<uint16_t>(40960ULL))){
return std::string("mulx",4);
}
else if ((cast<uint16_t>((op & cast<uint16_t>(58880ULL))) == cast<uint16_t>(41472ULL))){
return go_fmt_Sprintf(std::string("mulxmvz ac%d",12),d);
}
else if ((cast<uint16_t>((op & cast<uint16_t>(58880ULL))) == cast<uint16_t>(41984ULL))){
return go_fmt_Sprintf(std::string("mulxac ac%d",11),d);
}
else if ((cast<uint16_t>((op & cast<uint16_t>(58880ULL))) == cast<uint16_t>(42496ULL))){
return go_fmt_Sprintf(std::string("mulxmv ac%d",11),d);
}
else if ((cast<uint16_t>((op & cast<uint16_t>(59136ULL))) == cast<uint16_t>(49152ULL))){
return go_fmt_Sprintf(std::string("mulc   ac%d.m, ax%d.h",21),cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(12ULL))) & cast<uint16_t>(1ULL))),cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(11ULL))) & cast<uint16_t>(1ULL))));
}
else if ((cast<uint16_t>((op & cast<uint16_t>(58880ULL))) == cast<uint16_t>(49664ULL))){
return go_fmt_Sprintf(std::string("mulcmvz ac%d.m, ax%d.h, ac%d",28),cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(12ULL))) & cast<uint16_t>(1ULL))),cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(11ULL))) & cast<uint16_t>(1ULL))),d);
}
else if ((cast<uint16_t>((op & cast<uint16_t>(58880ULL))) == cast<uint16_t>(50176ULL))){
return go_fmt_Sprintf(std::string("mulcac ac%d.m, ax%d.h, ac%d",27),cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(12ULL))) & cast<uint16_t>(1ULL))),cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(11ULL))) & cast<uint16_t>(1ULL))),d);
}
else if ((cast<uint16_t>((op & cast<uint16_t>(58880ULL))) == cast<uint16_t>(50688ULL))){
return go_fmt_Sprintf(std::string("mulcmv ac%d.m, ax%d.h, ac%d",27),cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(12ULL))) & cast<uint16_t>(1ULL))),cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(11ULL))) & cast<uint16_t>(1ULL))),d);
}
else if ((cast<uint16_t>((op & cast<uint16_t>(64512ULL))) == cast<uint16_t>(57344ULL))){
return go_fmt_Sprintf(std::string("maddx  ax0.%s, ax1.%s",21),gcdsp_lowHigh(cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(9ULL))) & cast<uint16_t>(1ULL)))),gcdsp_lowHigh(cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(8ULL))) & cast<uint16_t>(1ULL)))));
}
else if ((cast<uint16_t>((op & cast<uint16_t>(64512ULL))) == cast<uint16_t>(58368ULL))){
return go_fmt_Sprintf(std::string("msubx  ax0.%s, ax1.%s",21),gcdsp_lowHigh(cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(9ULL))) & cast<uint16_t>(1ULL)))),gcdsp_lowHigh(cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(8ULL))) & cast<uint16_t>(1ULL)))));
}
else if ((cast<uint16_t>((op & cast<uint16_t>(64512ULL))) == cast<uint16_t>(59392ULL))){
return go_fmt_Sprintf(std::string("maddc  ac%d.m, ax%d.h",21),cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(9ULL))) & cast<uint16_t>(1ULL))),cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(8ULL))) & cast<uint16_t>(1ULL))));
}
else if ((cast<uint16_t>((op & cast<uint16_t>(64512ULL))) == cast<uint16_t>(60416ULL))){
return go_fmt_Sprintf(std::string("msubc  ac%d.m, ax%d.h",21),cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(9ULL))) & cast<uint16_t>(1ULL))),cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(8ULL))) & cast<uint16_t>(1ULL))));
}
else if ((cast<uint16_t>((op & cast<uint16_t>(65024ULL))) == cast<uint16_t>(61952ULL))){
return go_fmt_Sprintf(std::string("madd   ax%d",11),d);
}
else if ((cast<uint16_t>((op & cast<uint16_t>(65024ULL))) == cast<uint16_t>(62976ULL))){
return go_fmt_Sprintf(std::string("msub   ax%d",11),d);
}
else if ((cast<uint16_t>((op & cast<uint16_t>(63488ULL))) == cast<uint16_t>(16384ULL))){
return go_fmt_Sprintf(std::string("addr   ac%d",11),d);
}
else if ((cast<uint16_t>((op & cast<uint16_t>(64512ULL))) == cast<uint16_t>(18432ULL))){
return go_fmt_Sprintf(std::string("addax  ac%d, ax%d",17),d,cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(9ULL))) & cast<uint16_t>(1ULL))));
}
else if ((cast<uint16_t>((op & cast<uint16_t>(65024ULL))) == cast<uint16_t>(19456ULL))){
return go_fmt_Sprintf(std::string("add    ac%d",11),d);
}
else if ((cast<uint16_t>((op & cast<uint16_t>(65024ULL))) == cast<uint16_t>(19968ULL))){
return go_fmt_Sprintf(std::string("addp   ac%d",11),d);
}
else if ((cast<uint16_t>((op & cast<uint16_t>(63488ULL))) == cast<uint16_t>(20480ULL))){
return go_fmt_Sprintf(std::string("subr   ac%d",11),d);
}
else if ((cast<uint16_t>((op & cast<uint16_t>(64512ULL))) == cast<uint16_t>(22528ULL))){
return go_fmt_Sprintf(std::string("subax  ac%d, ax%d",17),d,cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(9ULL))) & cast<uint16_t>(1ULL))));
}
else if ((cast<uint16_t>((op & cast<uint16_t>(65024ULL))) == cast<uint16_t>(23552ULL))){
return go_fmt_Sprintf(std::string("sub    ac%d",11),d);
}
else if ((cast<uint16_t>((op & cast<uint16_t>(63488ULL))) == cast<uint16_t>(24576ULL))){
return go_fmt_Sprintf(std::string("movr   ac%d, r%d",16),d,cast<uint16_t>((cast<uint16_t>(24ULL) + (cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(9ULL))) & cast<uint16_t>(3ULL)))))));
}
else if ((cast<uint16_t>((op & cast<uint16_t>(64512ULL))) == cast<uint16_t>(26624ULL))){
return go_fmt_Sprintf(std::string("movax  ac%d, ax%d",17),d,cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(9ULL))) & cast<uint16_t>(1ULL))));
}
else if ((cast<uint16_t>((op & cast<uint16_t>(65024ULL))) == cast<uint16_t>(27648ULL))){
return go_fmt_Sprintf(std::string("mov    ac%d",11),d);
}
else if ((cast<uint16_t>((op & cast<uint16_t>(65024ULL))) == cast<uint16_t>(28160ULL))){
return go_fmt_Sprintf(std::string("movp   ac%d",11),d);
}
else if ((cast<uint16_t>((op & cast<uint16_t>(65024ULL))) == cast<uint16_t>(32256ULL))){
return go_fmt_Sprintf(std::string("movnp  ac%d",11),d);
}
else if ((cast<uint16_t>((op & cast<uint16_t>(65024ULL))) == cast<uint16_t>(65024ULL))){
return go_fmt_Sprintf(std::string("movpz  ac%d",11),d);
}
else if ((cast<uint16_t>((op & cast<uint16_t>(65024ULL))) == cast<uint16_t>(29696ULL))){
return go_fmt_Sprintf(std::string("incm   ac%d.m",13),d);
}
else if ((cast<uint16_t>((op & cast<uint16_t>(65024ULL))) == cast<uint16_t>(30208ULL))){
return go_fmt_Sprintf(std::string("inc    ac%d",11),d);
}
else if ((cast<uint16_t>((op & cast<uint16_t>(65024ULL))) == cast<uint16_t>(30720ULL))){
return go_fmt_Sprintf(std::string("decm   ac%d.m",13),d);
}
else if ((cast<uint16_t>((op & cast<uint16_t>(65024ULL))) == cast<uint16_t>(31232ULL))){
return go_fmt_Sprintf(std::string("dec    ac%d",11),d);
}
else if ((cast<uint16_t>((op & cast<uint16_t>(65024ULL))) == cast<uint16_t>(31744ULL))){
return go_fmt_Sprintf(std::string("neg    ac%d",11),d);
}
else if ((cast<uint16_t>((op & cast<uint16_t>(64512ULL))) == cast<uint16_t>(28672ULL))){
return go_fmt_Sprintf(std::string("addaxl ac%d, ax%d",17),d,cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(9ULL))) & cast<uint16_t>(1ULL))));
}
else if ((cast<uint16_t>((op & cast<uint16_t>(64512ULL))) == cast<uint16_t>(63488ULL))){
return go_fmt_Sprintf(std::string("addpaxz ac%d, ax%d",18),d,cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(9ULL))) & cast<uint16_t>(1ULL))));
}
else if ((cast<uint16_t>((op & cast<uint16_t>(65024ULL))) == cast<uint16_t>(61440ULL))){
return go_fmt_Sprintf(std::string("lsl16  ac%d",11),d);
}
else if ((cast<uint16_t>((op & cast<uint16_t>(65024ULL))) == cast<uint16_t>(62464ULL))){
return go_fmt_Sprintf(std::string("lsr16  ac%d",11),d);
}
}
tmp4:;
return go_fmt_Sprintf(std::string("op8_%02X",8),shr<uint16_t>(op,cast<int64_t>(8ULL)));
}
}
// tools/cpu/gcdsp/disasm.go:444:1
std::string gcdsp_extMnemonic(uint16_t ext){
{
{
if ((ext == cast<uint16_t>(0ULL))){
return std::string("",0);
}
else if ((cast<uint16_t>((ext & cast<uint16_t>(192ULL))) == cast<uint16_t>(128ULL))){
std::string name = std::string("ls",2);
if ((cast<uint16_t>((ext & cast<uint16_t>(2ULL))) != cast<uint16_t>(0ULL))) {
name = std::string("sl",2);
}
if ((cast<uint16_t>((ext & cast<uint16_t>(4ULL))) != cast<uint16_t>(0ULL))) {
name += std::string("n",1);
}
if ((cast<uint16_t>((ext & cast<uint16_t>(8ULL))) != cast<uint16_t>(0ULL))) {
name += std::string("m",1);
}
return go_fmt_Sprintf(std::string("%s r%d, ac%d.m",14),name,cast<uint16_t>((cast<uint16_t>(24ULL) + (cast<uint16_t>(((shr<uint16_t>(ext,cast<int64_t>(4ULL))) & cast<uint16_t>(3ULL)))))),cast<uint16_t>((ext & cast<uint16_t>(1ULL))));
}
else if ((cast<uint16_t>((ext & cast<uint16_t>(192ULL))) == cast<uint16_t>(192ULL))){
std::string suffix = std::string("",0);
if ((cast<uint16_t>((ext & cast<uint16_t>(4ULL))) != cast<uint16_t>(0ULL))) {
suffix += std::string("n",1);
}
if ((cast<uint16_t>((ext & cast<uint16_t>(8ULL))) != cast<uint16_t>(0ULL))) {
suffix += std::string("m",1);
}
auto half = [&](uint16_t bit)->std::string{
if ((bit != cast<uint16_t>(0ULL))) {
return std::string("h",1);
}
return std::string("l",1);
}
;
if ((cast<uint16_t>((ext & cast<uint16_t>(3ULL))) == cast<uint16_t>(3ULL))) {
return go_fmt_Sprintf(std::string("ld2%s ax%d, @ar%d",17),suffix,cast<uint16_t>(((shr<uint16_t>(ext,cast<int64_t>(4ULL))) & cast<uint16_t>(1ULL))),cast<uint16_t>(((shr<uint16_t>(ext,cast<int64_t>(5ULL))) & cast<uint16_t>(1ULL))));
}
return go_fmt_Sprintf(std::string("ld%s ax0.%s, ax1.%s, @ar%d",26),suffix,half(cast<uint16_t>((ext & cast<uint16_t>(32ULL)))),half(cast<uint16_t>((ext & cast<uint16_t>(16ULL)))),cast<uint16_t>((ext & cast<uint16_t>(3ULL))));
}
else if ((cast<uint16_t>((ext & cast<uint16_t>(192ULL))) == cast<uint16_t>(64ULL))){
std::string name = std::string("l",1);
if ((cast<uint16_t>((ext & cast<uint16_t>(4ULL))) != cast<uint16_t>(0ULL))) {
name = std::string("ln",2);
}
return go_fmt_Sprintf(std::string("%s r%d, @ar%d",13),name,cast<uint16_t>((cast<uint16_t>(24ULL) + (cast<uint16_t>(((shr<uint16_t>(ext,cast<int64_t>(3ULL))) & cast<uint16_t>(7ULL)))))),cast<uint16_t>((ext & cast<uint16_t>(3ULL))));
}
else if ((cast<uint16_t>((ext & cast<uint16_t>(224ULL))) == cast<uint16_t>(32ULL))){
std::string name = std::string("s",1);
if ((cast<uint16_t>((ext & cast<uint16_t>(4ULL))) != cast<uint16_t>(0ULL))) {
name = std::string("sn",2);
}
return go_fmt_Sprintf(std::string("%s @ar%d, r%d",13),name,cast<uint16_t>((ext & cast<uint16_t>(3ULL))),cast<uint16_t>((cast<uint16_t>(28ULL) + (cast<uint16_t>(((shr<uint16_t>(ext,cast<int64_t>(3ULL))) & cast<uint16_t>(3ULL)))))));
}
else if ((cast<uint16_t>((ext & cast<uint16_t>(240ULL))) == cast<uint16_t>(16ULL))){
return go_fmt_Sprintf(std::string("mv r%d, r%d",11),cast<uint16_t>((cast<uint16_t>(24ULL) + (cast<uint16_t>(((shr<uint16_t>(ext,cast<int64_t>(2ULL))) & cast<uint16_t>(3ULL)))))),cast<uint16_t>((cast<uint16_t>(28ULL) + (cast<uint16_t>((ext & cast<uint16_t>(3ULL)))))));
}
else if ((cast<uint16_t>((ext & cast<uint16_t>(252ULL))) == cast<uint16_t>(4ULL))){
return go_fmt_Sprintf(std::string("dr ar%d",7),cast<uint16_t>((ext & cast<uint16_t>(3ULL))));
}
else if ((cast<uint16_t>((ext & cast<uint16_t>(252ULL))) == cast<uint16_t>(8ULL))){
return go_fmt_Sprintf(std::string("ir ar%d",7),cast<uint16_t>((ext & cast<uint16_t>(3ULL))));
}
else if ((cast<uint16_t>((ext & cast<uint16_t>(252ULL))) == cast<uint16_t>(12ULL))){
return go_fmt_Sprintf(std::string("nr ar%d",7),cast<uint16_t>((ext & cast<uint16_t>(3ULL))));
}
}
tmp5:;
return go_fmt_Sprintf(std::string("ext 0x%02X",10),ext);
}
}
// tools/cpu/gcdsp/disasm.go:505:1
std::tuple<uint16_t,bool> gcdsp_branchTarget(std::function<uint16_t(uint16_t)> read,uint16_t pc){
uint16_t target{};
bool isBranch{};
{
uint16_t op = read(pc);
{
if ((cast<uint16_t>((op & cast<uint16_t>(65520ULL))) == cast<uint16_t>(656ULL)) || (cast<uint16_t>((op & cast<uint16_t>(65520ULL))) == cast<uint16_t>(688ULL)) || (cast<uint16_t>((op & cast<uint16_t>(65504ULL))) == cast<uint16_t>(96ULL)) || (cast<uint16_t>((op & cast<uint16_t>(65280ULL))) == cast<uint16_t>(4352ULL))){
return {read(cast<uint16_t>((pc + cast<uint16_t>(1ULL)))),true};
}
}
tmp6:;
return {cast<uint16_t>(0ULL),false};
}
}
// tools/cpu/gcdsp/disasm.go:521:1
std::tuple<Slice<uint16_t>,int64_t,bool> gcdsp_DisasmValidate(std::function<uint16_t(uint16_t)> read,uint16_t nWords){
Slice<uint16_t> misaligned{};
int64_t boundaries{};
bool overrun{};
{
Map<uint16_t,bool> starts = Map<uint16_t,bool>{};
{uint16_t w = cast<uint16_t>(0ULL);for (;(w < nWords);){
starts[w] = true;
auto tmp7 = gcdsp_Disasm(read,w);
uint16_t span = std::get<1>(tmp7);
if ((cast<uint16_t>((w + span)) > nWords)) {
overrun = true;
}
w += span;
}
}{uint16_t w = cast<uint16_t>(0ULL);for (;(w < nWords);){
{
auto tmp8 = gcdsp_branchTarget(read,w);
uint16_t t = std::get<0>(tmp8);
bool ok = std::get<1>(tmp8);
if (((ok && (t < nWords)) && (!get(starts,t)))) {
misaligned = append(misaligned,Slice<uint16_t>{w});
}
}
auto tmp9 = gcdsp_Disasm(read,w);
uint16_t span = std::get<1>(tmp9);
w += span;
}
}return {misaligned,len(starts),overrun};
}
}
// tools/cpu/gcdsp/disasm.go:543:1
std::string gcdsp_DisasmRange(std::function<uint16_t(uint16_t)> read,uint16_t pc,uint16_t nWords){
{
strings_Builder b={};
{uint16_t w = cast<uint16_t>(0ULL);for (;(w < nWords);){
uint16_t addr = cast<uint16_t>((pc + w));
auto tmp10 = gcdsp_Disasm(read,addr);
std::string text = std::get<0>(tmp10);
uint16_t span = std::get<1>(tmp10);
uint16_t raw = read(addr);
if ((span == cast<uint16_t>(2ULL))) {
go_fmt_Fprintf((&b),std::string("%04X  %04X %04X  %s\012",20),addr,raw,read(cast<uint16_t>((addr + cast<uint16_t>(1ULL)))),text);
}
else {
go_fmt_Fprintf((&b),std::string("%04X  %04X       %s\012",20),addr,raw,text);
}
w += span;
}
}return strings_Builder_String(&(b));
}
}
// tools/cpu/gcdsp/exec.go:15:1
bool gcdsp_CPU_Step(gcdsp_CPU* c){
{
if (c->Halted) {
return false;
}
uint16_t pc = c->PC;
uint16_t op = gcdsp_CPU_imem(c,pc);
uint16_t span = gcdsp_CPU_execute(c,pc,op);
if (c->Halted) {
return false;
}
bool branched = c->Branched;
c->Branched = false;
if ((!branched)) {
c->PC = cast<uint16_t>((pc + span));
}
gcdsp_CPU_serviceLoops(c,pc,branched);
return (!c->Halted);
}
}
// tools/cpu/gcdsp/exec.go:40:1
uint16_t gcdsp_CPU_execute(gcdsp_CPU* c,uint16_t pc,uint16_t op){
uint16_t span{};
{
{
if ((op == cast<uint16_t>(0ULL))){
return cast<uint16_t>(1ULL);
}
else if ((op == cast<uint16_t>(33ULL))){
gcdsp_CPU_Halt(c,std::string("ucode executed HALT at 0x%04X",29),pc);
return cast<uint16_t>(1ULL);
}
else if ((cast<uint16_t>((op & cast<uint16_t>(65532ULL))) == cast<uint16_t>(4ULL))){
gcdsp_CPU_arDec(c,cast<int64_t>(cast<uint16_t>((op & cast<uint16_t>(3ULL)))));
return cast<uint16_t>(1ULL);
}
else if ((cast<uint16_t>((op & cast<uint16_t>(65532ULL))) == cast<uint16_t>(8ULL))){
gcdsp_CPU_arInc(c,cast<int64_t>(cast<uint16_t>((op & cast<uint16_t>(3ULL)))));
return cast<uint16_t>(1ULL);
}
else if ((cast<uint16_t>((op & cast<uint16_t>(65532ULL))) == cast<uint16_t>(12ULL))){
int64_t n = cast<int64_t>(cast<uint16_t>((op & cast<uint16_t>(3ULL))));
gcdsp_CPU_arSub(c,n,cast<int16_t>(c->Reg[cast<int64_t>((cast<int64_t>(4ULL) + n))]));
return cast<uint16_t>(1ULL);
}
else if ((cast<uint16_t>((op & cast<uint16_t>(65520ULL))) == cast<uint16_t>(16ULL))){
int64_t d = cast<int64_t>(cast<uint16_t>((op & cast<uint16_t>(3ULL))));
int64_t s = cast<int64_t>(cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(2ULL))) & cast<uint16_t>(3ULL))));
gcdsp_CPU_arAdd(c,d,cast<int16_t>(c->Reg[cast<int64_t>((cast<int64_t>(4ULL) + s))]));
return cast<uint16_t>(1ULL);
}
else if ((cast<uint16_t>((op & cast<uint16_t>(65504ULL))) == cast<uint16_t>(64ULL))){
gcdsp_CPU_startLoop(c,cast<uint16_t>((pc + cast<uint16_t>(1ULL))),cast<uint16_t>((pc + cast<uint16_t>(1ULL))),c->Reg[cast<uint16_t>((op & cast<uint16_t>(31ULL)))]);
return cast<uint16_t>(1ULL);
}
else if ((cast<uint16_t>((op & cast<uint16_t>(65504ULL))) == cast<uint16_t>(96ULL))){
uint16_t end = gcdsp_CPU_imem(c,cast<uint16_t>((pc + cast<uint16_t>(1ULL))));
gcdsp_CPU_startLoop(c,cast<uint16_t>((pc + cast<uint16_t>(2ULL))),end,c->Reg[cast<uint16_t>((op & cast<uint16_t>(31ULL)))]);
return cast<uint16_t>(2ULL);
}
else if ((cast<uint16_t>((op & cast<uint16_t>(65280ULL))) == cast<uint16_t>(4096ULL))){
gcdsp_CPU_startLoop(c,cast<uint16_t>((pc + cast<uint16_t>(1ULL))),cast<uint16_t>((pc + cast<uint16_t>(1ULL))),cast<uint16_t>((op & cast<uint16_t>(255ULL))));
return cast<uint16_t>(1ULL);
}
else if ((cast<uint16_t>((op & cast<uint16_t>(65280ULL))) == cast<uint16_t>(4352ULL))){
uint16_t end = gcdsp_CPU_imem(c,cast<uint16_t>((pc + cast<uint16_t>(1ULL))));
gcdsp_CPU_startLoop(c,cast<uint16_t>((pc + cast<uint16_t>(2ULL))),end,cast<uint16_t>((op & cast<uint16_t>(255ULL))));
return cast<uint16_t>(2ULL);
}
else if ((cast<uint16_t>((op & cast<uint16_t>(65280ULL))) == cast<uint16_t>(4608ULL))){
c->Reg[cast<int64_t>(19ULL)] &= ~(shl<uint16_t>(cast<uint16_t>(1ULL),(cast<uint16_t>((cast<uint16_t>(6ULL) + (cast<uint16_t>((op & cast<uint16_t>(255ULL)))))))));
return cast<uint16_t>(1ULL);
}
else if ((cast<uint16_t>((op & cast<uint16_t>(65280ULL))) == cast<uint16_t>(4864ULL))){
c->Reg[cast<int64_t>(19ULL)] |= shl<uint16_t>(cast<uint16_t>(1ULL),(cast<uint16_t>((cast<uint16_t>(6ULL) + (cast<uint16_t>((op & cast<uint16_t>(255ULL))))))));
return cast<uint16_t>(1ULL);
}
else if ((cast<uint16_t>((op & cast<uint16_t>(65504ULL))) == cast<uint16_t>(128ULL))){
gcdsp_CPU_setRegExtend(c,cast<uint16_t>((op & cast<uint16_t>(31ULL))),gcdsp_CPU_imem(c,cast<uint16_t>((pc + cast<uint16_t>(1ULL)))));
return cast<uint16_t>(2ULL);
}
else if ((cast<uint16_t>((op & cast<uint16_t>(65504ULL))) == cast<uint16_t>(192ULL))){
gcdsp_CPU_setRegExtend(c,cast<uint16_t>((op & cast<uint16_t>(31ULL))),gcdsp_CPU_dataRead(c,gcdsp_CPU_imem(c,cast<uint16_t>((pc + cast<uint16_t>(1ULL))))));
return cast<uint16_t>(2ULL);
}
else if ((cast<uint16_t>((op & cast<uint16_t>(65504ULL))) == cast<uint16_t>(224ULL))){
gcdsp_CPU_dataWrite(c,gcdsp_CPU_imem(c,cast<uint16_t>((pc + cast<uint16_t>(1ULL)))),gcdsp_CPU_getReg(c,cast<uint16_t>((op & cast<uint16_t>(31ULL)))));
return cast<uint16_t>(2ULL);
}
else if ((cast<uint16_t>((op & cast<uint16_t>(65280ULL))) == cast<uint16_t>(5632ULL))){
uint16_t addr = cast<uint16_t>((op & cast<uint16_t>(255ULL)));
if ((addr >= cast<uint16_t>(128ULL))) {
addr |= cast<uint16_t>(65280ULL);
}
gcdsp_CPU_dataWrite(c,addr,gcdsp_CPU_imem(c,cast<uint16_t>((pc + cast<uint16_t>(1ULL)))));
return cast<uint16_t>(2ULL);
}
else if ((cast<uint16_t>((op & cast<uint16_t>(63488ULL))) == cast<uint16_t>(8192ULL))){
gcdsp_CPU_setRegExtend(c,cast<uint16_t>((cast<uint16_t>(24ULL) + (cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(8ULL))) & cast<uint16_t>(7ULL)))))),gcdsp_CPU_dataRead(c,cast<uint16_t>((shl<uint16_t>((cast<uint16_t>((c->Reg[cast<int64_t>(18ULL)] & cast<uint16_t>(255ULL)))),cast<int64_t>(8ULL)) | (cast<uint16_t>((op & cast<uint16_t>(255ULL))))))));
return cast<uint16_t>(1ULL);
}
else if ((cast<uint16_t>((op & cast<uint16_t>(63488ULL))) == cast<uint16_t>(10240ULL))){
gcdsp_CPU_dataWrite(c,cast<uint16_t>((shl<uint16_t>((cast<uint16_t>((c->Reg[cast<int64_t>(18ULL)] & cast<uint16_t>(255ULL)))),cast<int64_t>(8ULL)) | (cast<uint16_t>((op & cast<uint16_t>(255ULL)))))),gcdsp_CPU_getReg(c,cast<uint16_t>((cast<uint16_t>(24ULL) + (cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(8ULL))) & cast<uint16_t>(7ULL))))))));
return cast<uint16_t>(1ULL);
}
else if ((cast<uint16_t>((op & cast<uint16_t>(64512ULL))) == cast<uint16_t>(6144ULL))){
int64_t s = cast<int64_t>(cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(5ULL))) & cast<uint16_t>(3ULL))));
uint16_t reg = cast<uint16_t>((op & cast<uint16_t>(31ULL)));
uint16_t addr = c->Reg[cast<int64_t>((cast<int64_t>(0ULL) + s))];
if ((cast<uint16_t>((op & cast<uint16_t>(512ULL))) != cast<uint16_t>(0ULL))) {
gcdsp_CPU_dataWrite(c,addr,gcdsp_CPU_getReg(c,reg));
}
else {
gcdsp_CPU_setRegExtend(c,reg,gcdsp_CPU_dataRead(c,addr));
}
{
switch(cast<uint16_t>((op & cast<uint16_t>(384ULL)))){
case cast<uint16_t>(128ULL):{
gcdsp_CPU_arDec(c,s);
break;}
case cast<uint16_t>(256ULL):{
gcdsp_CPU_arInc(c,s);
break;}
case cast<uint16_t>(384ULL):{
gcdsp_CPU_arAdd(c,s,cast<int16_t>(c->Reg[cast<int64_t>((cast<int64_t>(4ULL) + s))]));
break;}
}}
return cast<uint16_t>(1ULL);
}
else if ((cast<uint16_t>((op & cast<uint16_t>(64512ULL))) == cast<uint16_t>(7168ULL))){
gcdsp_CPU_setRegExtend(c,cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(5ULL))) & cast<uint16_t>(31ULL))),gcdsp_CPU_getReg(c,cast<uint16_t>((op & cast<uint16_t>(31ULL)))));
return cast<uint16_t>(1ULL);
}
else if ((cast<uint16_t>((op & cast<uint16_t>(65024ULL))) == cast<uint16_t>(5120ULL))){
int64_t r = cast<int64_t>(cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(8ULL))) & cast<uint16_t>(1ULL))));
bool arith = (cast<uint16_t>((op & cast<uint16_t>(128ULL))) != cast<uint16_t>(0ULL));
int64_t amt = cast<int64_t>(cast<uint16_t>((op & cast<uint16_t>(127ULL))));
if ((cast<int64_t>((amt & cast<int64_t>(64ULL))) != cast<int64_t>(0ULL))) {
amt -= cast<int64_t>(128ULL);
}
gcdsp_CPU_shiftAcc(c,r,arith,amt);
return cast<uint16_t>(1ULL);
}
else if ((cast<uint16_t>((op & cast<uint16_t>(63488ULL))) == cast<uint16_t>(2048ULL))){
gcdsp_CPU_setRegExtend(c,cast<uint16_t>((cast<uint16_t>(24ULL) + (cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(8ULL))) & cast<uint16_t>(7ULL)))))),cast<uint16_t>(cast<int16_t>(cast<int8_t>(cast<uint16_t>((op & cast<uint16_t>(255ULL)))))));
return cast<uint16_t>(1ULL);
}
else if ((cast<uint16_t>((op & cast<uint16_t>(65279ULL))) == cast<uint16_t>(512ULL))){
int64_t d = cast<int64_t>(cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(8ULL))) & cast<uint16_t>(1ULL))));
gcdsp_CPU_setAc(c,d,cast<int64_t>((gcdsp_CPU_ac(c,d) + shl<int64_t>(cast<int64_t>(cast<int16_t>(gcdsp_CPU_imem(c,cast<uint16_t>((pc + cast<uint16_t>(1ULL)))))),cast<int64_t>(16ULL)))));
gcdsp_CPU_setArithFlags(c,d);
return cast<uint16_t>(2ULL);
}
else if ((cast<uint16_t>((op & cast<uint16_t>(65279ULL))) == cast<uint16_t>(544ULL))){
int64_t d = cast<int64_t>(cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(8ULL))) & cast<uint16_t>(1ULL))));
c->Reg[cast<int64_t>((cast<int64_t>(30ULL) + d))] ^= gcdsp_CPU_imem(c,cast<uint16_t>((pc + cast<uint16_t>(1ULL))));
gcdsp_CPU_setLogicFlags(c,d);
return cast<uint16_t>(2ULL);
}
else if ((cast<uint16_t>((op & cast<uint16_t>(65279ULL))) == cast<uint16_t>(576ULL))){
int64_t d = cast<int64_t>(cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(8ULL))) & cast<uint16_t>(1ULL))));
c->Reg[cast<int64_t>((cast<int64_t>(30ULL) + d))] &= gcdsp_CPU_imem(c,cast<uint16_t>((pc + cast<uint16_t>(1ULL))));
gcdsp_CPU_setLogicFlags(c,d);
return cast<uint16_t>(2ULL);
}
else if ((cast<uint16_t>((op & cast<uint16_t>(65279ULL))) == cast<uint16_t>(608ULL))){
int64_t d = cast<int64_t>(cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(8ULL))) & cast<uint16_t>(1ULL))));
c->Reg[cast<int64_t>((cast<int64_t>(30ULL) + d))] |= gcdsp_CPU_imem(c,cast<uint16_t>((pc + cast<uint16_t>(1ULL))));
gcdsp_CPU_setLogicFlags(c,d);
return cast<uint16_t>(2ULL);
}
else if ((cast<uint16_t>((op & cast<uint16_t>(65279ULL))) == cast<uint16_t>(640ULL))){
int64_t d = cast<int64_t>(cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(8ULL))) & cast<uint16_t>(1ULL))));
gcdsp_CPU_subFlags(c,gcdsp_CPU_ac(c,d),shl<int64_t>(cast<int64_t>(cast<int16_t>(gcdsp_CPU_imem(c,cast<uint16_t>((pc + cast<uint16_t>(1ULL)))))),cast<int64_t>(16ULL)));
return cast<uint16_t>(2ULL);
}
else if ((cast<uint16_t>((op & cast<uint16_t>(65279ULL))) == cast<uint16_t>(672ULL))){
int64_t d = cast<int64_t>(cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(8ULL))) & cast<uint16_t>(1ULL))));
gcdsp_CPU_setFlag(c,cast<uint16_t>(64ULL),(cast<uint16_t>((c->Reg[cast<int64_t>((cast<int64_t>(30ULL) + d))] & gcdsp_CPU_imem(c,cast<uint16_t>((pc + cast<uint16_t>(1ULL)))))) == cast<uint16_t>(0ULL)));
return cast<uint16_t>(2ULL);
}
else if ((cast<uint16_t>((op & cast<uint16_t>(65279ULL))) == cast<uint16_t>(704ULL))){
int64_t d = cast<int64_t>(cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(8ULL))) & cast<uint16_t>(1ULL))));
uint16_t imm = gcdsp_CPU_imem(c,cast<uint16_t>((pc + cast<uint16_t>(1ULL))));
gcdsp_CPU_setFlag(c,cast<uint16_t>(64ULL),(cast<uint16_t>((c->Reg[cast<int64_t>((cast<int64_t>(30ULL) + d))] & imm)) == imm));
return cast<uint16_t>(2ULL);
}
else if ((cast<uint16_t>((op & cast<uint16_t>(65024ULL))) == cast<uint16_t>(1024ULL))){
int64_t d = cast<int64_t>(cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(8ULL))) & cast<uint16_t>(1ULL))));
gcdsp_CPU_aluAddSub(c,d,shl<int64_t>(cast<int64_t>(cast<int8_t>(op)),cast<int64_t>(16ULL)),false);
return cast<uint16_t>(1ULL);
}
else if ((cast<uint16_t>((op & cast<uint16_t>(65024ULL))) == cast<uint16_t>(1536ULL))){
int64_t d = cast<int64_t>(cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(8ULL))) & cast<uint16_t>(1ULL))));
gcdsp_CPU_subFlags(c,gcdsp_CPU_ac(c,d),shl<int64_t>(cast<int64_t>(cast<int8_t>(op)),cast<int64_t>(16ULL)));
return cast<uint16_t>(1ULL);
}
else if ((cast<uint16_t>((op & cast<uint16_t>(65520ULL))) == cast<uint16_t>(624ULL))){
if ((!gcdsp_CPU_cond(c,cast<uint16_t>((op & cast<uint16_t>(15ULL)))))) {
auto tmp12 = gcdsp_Disasm([&](uint16_t a)->uint16_t{
return gcdsp_CPU_imem(c,a);
}
,cast<uint16_t>((pc + cast<uint16_t>(1ULL))));
uint16_t span = std::get<1>(tmp12);
return cast<uint16_t>((cast<uint16_t>(1ULL) + span));
}
return cast<uint16_t>(1ULL);
}
else if ((op == cast<uint16_t>(714ULL))){
gcdsp_CPU_shiftAcc(c,cast<int64_t>(0ULL),false,cast<int64_t>(-gcdsp_shiftAmount7(c->Reg[cast<int64_t>(31ULL)])));
return cast<uint16_t>(1ULL);
}
else if ((op == cast<uint16_t>(715ULL))){
gcdsp_CPU_shiftAcc(c,cast<int64_t>(0ULL),true,cast<int64_t>(-gcdsp_shiftAmount7(c->Reg[cast<int64_t>(31ULL)])));
return cast<uint16_t>(1ULL);
}
else if ((cast<uint16_t>((op & cast<uint16_t>(65520ULL))) == cast<uint16_t>(656ULL))){
uint16_t dst = gcdsp_CPU_imem(c,cast<uint16_t>((pc + cast<uint16_t>(1ULL))));
if (gcdsp_CPU_cond(c,cast<uint16_t>((op & cast<uint16_t>(15ULL))))) {
c->PC = dst;
c->Branched = true;
}
return cast<uint16_t>(2ULL);
}
else if ((cast<uint16_t>((op & cast<uint16_t>(65520ULL))) == cast<uint16_t>(688ULL))){
uint16_t dst = gcdsp_CPU_imem(c,cast<uint16_t>((pc + cast<uint16_t>(1ULL))));
if (gcdsp_CPU_cond(c,cast<uint16_t>((op & cast<uint16_t>(15ULL))))) {
gcdsp_CPU_push(c,cast<uint16_t>(12ULL),cast<uint16_t>((pc + cast<uint16_t>(2ULL))));
c->PC = dst;
c->Branched = true;
}
return cast<uint16_t>(2ULL);
}
else if ((cast<uint16_t>((op & cast<uint16_t>(65520ULL))) == cast<uint16_t>(720ULL))){
if (gcdsp_CPU_cond(c,cast<uint16_t>((op & cast<uint16_t>(15ULL))))) {
c->PC = gcdsp_CPU_pop(c,cast<uint16_t>(12ULL));
c->Branched = true;
}
return cast<uint16_t>(1ULL);
}
else if ((cast<uint16_t>((op & cast<uint16_t>(65520ULL))) == cast<uint16_t>(752ULL))){
if (gcdsp_CPU_cond(c,cast<uint16_t>((op & cast<uint16_t>(15ULL))))) {
c->PC = gcdsp_CPU_pop(c,cast<uint16_t>(12ULL));
c->Branched = true;
c->InInterrupt = false;
}
return cast<uint16_t>(1ULL);
}
else if ((cast<uint16_t>((op & cast<uint16_t>(65280ULL))) == cast<uint16_t>(5888ULL))){
if (gcdsp_CPU_cond(c,cast<uint16_t>((op & cast<uint16_t>(15ULL))))) {
if ((cast<uint16_t>((op & cast<uint16_t>(16ULL))) != cast<uint16_t>(0ULL))) {
gcdsp_CPU_push(c,cast<uint16_t>(12ULL),cast<uint16_t>((pc + cast<uint16_t>(1ULL))));
}
c->PC = c->Reg[cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(5ULL))) & cast<uint16_t>(7ULL)))];
c->Branched = true;
}
return cast<uint16_t>(1ULL);
}
}
tmp11:;
if ((op >= cast<uint16_t>(12288ULL))) {
return gcdsp_CPU_execArith(c,pc,op);
}
gcdsp_CPU_Halt(c,std::string("unmodelled DSP instruction 0x%04X at 0x%04X (%s)",48),op,pc,gcdsp_mustText([=](auto...args){return gcdsp_CPU_imem(c,args...);},pc));
return cast<uint16_t>(1ULL);
}
}
// tools/cpu/gcdsp/exec.go:306:1
uint16_t gcdsp_CPU_execArith(gcdsp_CPU* c,uint16_t pc,uint16_t op){
{
{
if ((cast<uint16_t>((op & cast<uint16_t>(65280ULL))) == cast<uint16_t>(35328ULL))){
gcdsp_CPU_setFlag(c,cast<uint16_t>(8192ULL),false);
return cast<uint16_t>(1ULL);
}
else if ((cast<uint16_t>((op & cast<uint16_t>(65280ULL))) == cast<uint16_t>(35584ULL))){
gcdsp_CPU_setFlag(c,cast<uint16_t>(8192ULL),true);
return cast<uint16_t>(1ULL);
}
else if ((cast<uint16_t>((op & cast<uint16_t>(65280ULL))) == cast<uint16_t>(35840ULL))){
gcdsp_CPU_setFlag(c,cast<uint16_t>(32768ULL),false);
return cast<uint16_t>(1ULL);
}
else if ((cast<uint16_t>((op & cast<uint16_t>(65280ULL))) == cast<uint16_t>(36096ULL))){
gcdsp_CPU_setFlag(c,cast<uint16_t>(32768ULL),true);
return cast<uint16_t>(1ULL);
}
else if ((cast<uint16_t>((op & cast<uint16_t>(65280ULL))) == cast<uint16_t>(36352ULL))){
gcdsp_CPU_setFlag(c,cast<uint16_t>(16384ULL),false);
return cast<uint16_t>(1ULL);
}
else if ((cast<uint16_t>((op & cast<uint16_t>(65280ULL))) == cast<uint16_t>(36608ULL))){
gcdsp_CPU_setFlag(c,cast<uint16_t>(16384ULL),true);
return cast<uint16_t>(1ULL);
}
}
tmp13:;
uint16_t extBits = cast<uint16_t>((op & cast<uint16_t>(255ULL)));
if ((cast<uint16_t>((op & cast<uint16_t>(61440ULL))) == cast<uint16_t>(12288ULL))) {
extBits = cast<uint16_t>((op & cast<uint16_t>(127ULL)));
}
gcdsp_extPending p = gcdsp_CPU_extBegin(c,extBits);
{
if ((cast<uint16_t>((op & cast<uint16_t>(65280ULL))) == cast<uint16_t>(32768ULL))){
}
else if ((cast<uint16_t>((op & cast<uint16_t>(65280ULL))) == cast<uint16_t>(33792ULL))){
c->Reg[cast<int64_t>(20ULL)] = cast<uint16_t>(0ULL);
c->Reg[cast<int64_t>(21ULL)] = cast<uint16_t>(65520ULL);
c->Reg[cast<int64_t>(22ULL)] = cast<uint16_t>(255ULL);
c->Reg[cast<int64_t>(23ULL)] = cast<uint16_t>(16ULL);
}
else if ((cast<uint16_t>((op & cast<uint16_t>(65152ULL))) == cast<uint16_t>(12928ULL))){
int64_t d = cast<int64_t>(cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(8ULL))) & cast<uint16_t>(1ULL))));
gcdsp_CPU_setReg(c,cast<uint16_t>(cast<int64_t>((cast<int64_t>(30ULL) + d))),cast<uint16_t>(~c->Reg[cast<int64_t>((cast<int64_t>(30ULL) + d))]));
gcdsp_CPU_setLogicFlags(c,d);
}
else if ((cast<uint16_t>((op & cast<uint16_t>(64640ULL))) == cast<uint16_t>(12288ULL))){
int64_t d = cast<int64_t>(cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(8ULL))) & cast<uint16_t>(1ULL))));
gcdsp_CPU_setReg(c,cast<uint16_t>(cast<int64_t>((cast<int64_t>(30ULL) + d))),cast<uint16_t>((c->Reg[cast<int64_t>((cast<int64_t>(30ULL) + d))] ^ c->Reg[cast<uint16_t>((cast<uint16_t>(26ULL) + (cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(9ULL))) & cast<uint16_t>(1ULL))))))])));
gcdsp_CPU_setLogicFlags(c,d);
}
else if ((cast<uint16_t>((op & cast<uint16_t>(64640ULL))) == cast<uint16_t>(13312ULL))){
int64_t d = cast<int64_t>(cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(8ULL))) & cast<uint16_t>(1ULL))));
gcdsp_CPU_setReg(c,cast<uint16_t>(cast<int64_t>((cast<int64_t>(30ULL) + d))),cast<uint16_t>((c->Reg[cast<int64_t>((cast<int64_t>(30ULL) + d))] & c->Reg[cast<uint16_t>((cast<uint16_t>(26ULL) + (cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(9ULL))) & cast<uint16_t>(1ULL))))))])));
gcdsp_CPU_setLogicFlags(c,d);
}
else if ((cast<uint16_t>((op & cast<uint16_t>(64640ULL))) == cast<uint16_t>(14336ULL))){
int64_t d = cast<int64_t>(cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(8ULL))) & cast<uint16_t>(1ULL))));
gcdsp_CPU_setReg(c,cast<uint16_t>(cast<int64_t>((cast<int64_t>(30ULL) + d))),cast<uint16_t>((c->Reg[cast<int64_t>((cast<int64_t>(30ULL) + d))] | c->Reg[cast<uint16_t>((cast<uint16_t>(26ULL) + (cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(9ULL))) & cast<uint16_t>(1ULL))))))])));
gcdsp_CPU_setLogicFlags(c,d);
}
else if ((cast<uint16_t>((op & cast<uint16_t>(65152ULL))) == cast<uint16_t>(12416ULL))){
int64_t d = cast<int64_t>(cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(8ULL))) & cast<uint16_t>(1ULL))));
gcdsp_CPU_setReg(c,cast<uint16_t>(cast<int64_t>((cast<int64_t>(30ULL) + d))),cast<uint16_t>((c->Reg[cast<int64_t>((cast<int64_t>(30ULL) + d))] ^ c->Reg[cast<int64_t>((cast<int64_t>(30ULL) + (cast<int64_t>((cast<int64_t>(1ULL) - d)))))])));
gcdsp_CPU_setLogicFlags(c,d);
}
else if ((cast<uint16_t>((op & cast<uint16_t>(64640ULL))) == cast<uint16_t>(13440ULL))){
gcdsp_CPU_shiftAcc(c,cast<int64_t>(cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(8ULL))) & cast<uint16_t>(1ULL)))),false,gcdsp_shiftAmount7(c->Reg[cast<uint16_t>((cast<uint16_t>(26ULL) + (cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(9ULL))) & cast<uint16_t>(1ULL))))))]));
}
else if ((cast<uint16_t>((op & cast<uint16_t>(64640ULL))) == cast<uint16_t>(14464ULL))){
gcdsp_CPU_shiftAcc(c,cast<int64_t>(cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(8ULL))) & cast<uint16_t>(1ULL)))),true,gcdsp_shiftAmount7(c->Reg[cast<uint16_t>((cast<uint16_t>(26ULL) + (cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(9ULL))) & cast<uint16_t>(1ULL))))))]));
}
else if ((cast<uint16_t>((op & cast<uint16_t>(65152ULL))) == cast<uint16_t>(15488ULL))){
int64_t d = cast<int64_t>(cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(8ULL))) & cast<uint16_t>(1ULL))));
gcdsp_CPU_shiftAcc(c,d,false,gcdsp_shiftAmount7(c->Reg[cast<int64_t>((cast<int64_t>(30ULL) + (cast<int64_t>((cast<int64_t>(1ULL) - d)))))]));
}
else if ((cast<uint16_t>((op & cast<uint16_t>(65152ULL))) == cast<uint16_t>(16000ULL))){
int64_t d = cast<int64_t>(cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(8ULL))) & cast<uint16_t>(1ULL))));
gcdsp_CPU_shiftAcc(c,d,true,gcdsp_shiftAmount7(c->Reg[cast<int64_t>((cast<int64_t>(30ULL) + (cast<int64_t>((cast<int64_t>(1ULL) - d)))))]));
}
else if ((cast<uint16_t>((op & cast<uint16_t>(63232ULL))) == cast<uint16_t>(33024ULL))){
int64_t r = cast<int64_t>(cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(11ULL))) & cast<uint16_t>(1ULL))));
gcdsp_CPU_setAc(c,r,cast<int64_t>(0ULL));
gcdsp_CPU_setArithFlags(c,r);
}
else if ((cast<uint16_t>((op & cast<uint16_t>(65024ULL))) == cast<uint16_t>(34304ULL))){
int16_t v = cast<int16_t>(c->Reg[cast<uint16_t>((cast<uint16_t>(26ULL) + (cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(8ULL))) & cast<uint16_t>(1ULL))))))]);
gcdsp_CPU_setFlag(c,cast<uint16_t>(4ULL),(v == cast<int16_t>(0ULL)));
gcdsp_CPU_setFlag(c,cast<uint16_t>(8ULL),(v < cast<int16_t>(0ULL)));
gcdsp_CPU_setFlag(c,cast<uint16_t>(1ULL),false);
gcdsp_CPU_setFlag(c,cast<uint16_t>(2ULL),false);
}
else if ((cast<uint16_t>((op & cast<uint16_t>(63232ULL))) == cast<uint16_t>(45312ULL))){
gcdsp_CPU_setTestFlags(c,cast<int64_t>(cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(11ULL))) & cast<uint16_t>(1ULL)))));
}
else if ((cast<uint16_t>((op & cast<uint16_t>(65280ULL))) == cast<uint16_t>(33280ULL))){
gcdsp_CPU_subFlags(c,gcdsp_CPU_ac(c,cast<int64_t>(0ULL)),gcdsp_CPU_ac(c,cast<int64_t>(1ULL)));
}
else if ((cast<uint16_t>((op & cast<uint16_t>(59136ULL))) == cast<uint16_t>(49408ULL))){
gcdsp_CPU_subFlags(c,gcdsp_CPU_ac(c,cast<int64_t>(cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(11ULL))) & cast<uint16_t>(1ULL))))),shl<int64_t>(cast<int64_t>(cast<int16_t>(c->Reg[cast<uint16_t>((cast<uint16_t>(26ULL) + (cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(12ULL))) & cast<uint16_t>(1ULL))))))])),cast<int64_t>(16ULL)));
}
else if ((cast<uint16_t>((op & cast<uint16_t>(63488ULL))) == cast<uint16_t>(16384ULL))){
int64_t d = cast<int64_t>(cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(8ULL))) & cast<uint16_t>(1ULL))));
uint16_t reg = cast<uint16_t>((cast<uint16_t>(24ULL) + (cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(9ULL))) & cast<uint16_t>(3ULL))))));
gcdsp_CPU_aluAddSub(c,d,shl<int64_t>(cast<int64_t>(cast<int16_t>(gcdsp_CPU_getReg(c,reg))),cast<int64_t>(16ULL)),false);
}
else if ((cast<uint16_t>((op & cast<uint16_t>(64512ULL))) == cast<uint16_t>(18432ULL))){
gcdsp_CPU_aluAddSub(c,cast<int64_t>(cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(8ULL))) & cast<uint16_t>(1ULL)))),gcdsp_CPU_ax(c,cast<int64_t>(cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(9ULL))) & cast<uint16_t>(1ULL))))),false);
}
else if ((cast<uint16_t>((op & cast<uint16_t>(65024ULL))) == cast<uint16_t>(19456ULL))){
int64_t d = cast<int64_t>(cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(8ULL))) & cast<uint16_t>(1ULL))));
gcdsp_CPU_aluAddSub(c,d,gcdsp_CPU_ac(c,cast<int64_t>((cast<int64_t>(1ULL) - d))),false);
}
else if ((cast<uint16_t>((op & cast<uint16_t>(63488ULL))) == cast<uint16_t>(20480ULL))){
int64_t d = cast<int64_t>(cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(8ULL))) & cast<uint16_t>(1ULL))));
uint16_t reg = cast<uint16_t>((cast<uint16_t>(24ULL) + (cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(9ULL))) & cast<uint16_t>(3ULL))))));
gcdsp_CPU_aluAddSub(c,d,shl<int64_t>(cast<int64_t>(cast<int16_t>(gcdsp_CPU_getReg(c,reg))),cast<int64_t>(16ULL)),true);
}
else if ((cast<uint16_t>((op & cast<uint16_t>(64512ULL))) == cast<uint16_t>(22528ULL))){
gcdsp_CPU_aluAddSub(c,cast<int64_t>(cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(8ULL))) & cast<uint16_t>(1ULL)))),gcdsp_CPU_ax(c,cast<int64_t>(cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(9ULL))) & cast<uint16_t>(1ULL))))),true);
}
else if ((cast<uint16_t>((op & cast<uint16_t>(65024ULL))) == cast<uint16_t>(23552ULL))){
int64_t d = cast<int64_t>(cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(8ULL))) & cast<uint16_t>(1ULL))));
gcdsp_CPU_aluAddSub(c,d,gcdsp_CPU_ac(c,cast<int64_t>((cast<int64_t>(1ULL) - d))),true);
}
else if ((cast<uint16_t>((op & cast<uint16_t>(64512ULL))) == cast<uint16_t>(28672ULL))){
int64_t d = cast<int64_t>(cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(8ULL))) & cast<uint16_t>(1ULL))));
int64_t s = cast<int64_t>(cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(9ULL))) & cast<uint16_t>(1ULL))));
gcdsp_CPU_aluAddSub(c,d,cast<int64_t>(c->Reg[cast<int64_t>((cast<int64_t>(24ULL) + s))]),false);
}
else if ((cast<uint16_t>((op & cast<uint16_t>(65024ULL))) == cast<uint16_t>(29696ULL))){
gcdsp_CPU_aluAddSub(c,cast<int64_t>(cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(8ULL))) & cast<uint16_t>(1ULL)))),cast<int64_t>(65536ULL),false);
}
else if ((cast<uint16_t>((op & cast<uint16_t>(65024ULL))) == cast<uint16_t>(30208ULL))){
gcdsp_CPU_aluAddSub(c,cast<int64_t>(cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(8ULL))) & cast<uint16_t>(1ULL)))),cast<int64_t>(1ULL),false);
}
else if ((cast<uint16_t>((op & cast<uint16_t>(65024ULL))) == cast<uint16_t>(30720ULL))){
gcdsp_CPU_aluAddSub(c,cast<int64_t>(cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(8ULL))) & cast<uint16_t>(1ULL)))),cast<int64_t>(65536ULL),true);
}
else if ((cast<uint16_t>((op & cast<uint16_t>(65024ULL))) == cast<uint16_t>(31232ULL))){
gcdsp_CPU_aluAddSub(c,cast<int64_t>(cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(8ULL))) & cast<uint16_t>(1ULL)))),cast<int64_t>(1ULL),true);
}
else if ((cast<uint16_t>((op & cast<uint16_t>(65024ULL))) == cast<uint16_t>(31744ULL))){
int64_t d = cast<int64_t>(cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(8ULL))) & cast<uint16_t>(1ULL))));
gcdsp_CPU_setAc(c,d,cast<int64_t>(-gcdsp_CPU_ac(c,d)));
gcdsp_CPU_setArithFlags(c,d);
}
else if ((cast<uint16_t>((op & cast<uint16_t>(65024ULL))) == cast<uint16_t>(61440ULL))){
gcdsp_CPU_shiftAcc(c,cast<int64_t>(cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(8ULL))) & cast<uint16_t>(1ULL)))),false,cast<int64_t>(16ULL));
}
else if ((cast<uint16_t>((op & cast<uint16_t>(65024ULL))) == cast<uint16_t>(62464ULL))){
gcdsp_CPU_shiftAcc(c,cast<int64_t>(cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(8ULL))) & cast<uint16_t>(1ULL)))),false,cast<int64_t>(-16ULL));
}
else if ((cast<uint16_t>((op & cast<uint16_t>(63232ULL))) == cast<uint16_t>(37120ULL))){
gcdsp_CPU_shiftAcc(c,cast<int64_t>(cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(11ULL))) & cast<uint16_t>(1ULL)))),true,cast<int64_t>(-16ULL));
}
else if ((cast<uint16_t>((op & cast<uint16_t>(63488ULL))) == cast<uint16_t>(24576ULL))){
int64_t d = cast<int64_t>(cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(8ULL))) & cast<uint16_t>(1ULL))));
uint16_t reg = cast<uint16_t>((cast<uint16_t>(24ULL) + (cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(9ULL))) & cast<uint16_t>(3ULL))))));
gcdsp_CPU_setAc(c,d,shl<int64_t>(cast<int64_t>(cast<int16_t>(c->Reg[reg])),cast<int64_t>(16ULL)));
gcdsp_CPU_setArithFlags(c,d);
}
else if ((cast<uint16_t>((op & cast<uint16_t>(64512ULL))) == cast<uint16_t>(26624ULL))){
int64_t d = cast<int64_t>(cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(8ULL))) & cast<uint16_t>(1ULL))));
gcdsp_CPU_setAc(c,d,gcdsp_CPU_ax(c,cast<int64_t>(cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(9ULL))) & cast<uint16_t>(1ULL))))));
gcdsp_CPU_setArithFlags(c,d);
}
else if ((cast<uint16_t>((op & cast<uint16_t>(65024ULL))) == cast<uint16_t>(27648ULL))){
int64_t d = cast<int64_t>(cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(8ULL))) & cast<uint16_t>(1ULL))));
gcdsp_CPU_setAc(c,d,gcdsp_CPU_ac(c,cast<int64_t>((cast<int64_t>(1ULL) - d))));
gcdsp_CPU_setArithFlags(c,d);
}
else if ((cast<uint16_t>((op & cast<uint16_t>(63232ULL))) == cast<uint16_t>(36864ULL))){
int64_t s = cast<int64_t>(cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(11ULL))) & cast<uint16_t>(1ULL))));
gcdsp_CPU_setProd(c,gcdsp_CPU_mul16(c,c->Reg[cast<int64_t>((cast<int64_t>(24ULL) + s))],c->Reg[cast<int64_t>((cast<int64_t>(26ULL) + s))]));
}
else if ((cast<uint16_t>((op & cast<uint16_t>(62976ULL))) == cast<uint16_t>(37888ULL))){
int64_t s = cast<int64_t>(cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(11ULL))) & cast<uint16_t>(1ULL))));
int64_t d = cast<int64_t>(cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(8ULL))) & cast<uint16_t>(1ULL))));
gcdsp_CPU_setAc(c,d,cast<int64_t>((gcdsp_CPU_ac(c,d) + gcdsp_CPU_prod(c))));
gcdsp_CPU_setArithFlags(c,d);
gcdsp_CPU_setProd(c,gcdsp_CPU_mul16(c,c->Reg[cast<int64_t>((cast<int64_t>(24ULL) + s))],c->Reg[cast<int64_t>((cast<int64_t>(26ULL) + s))]));
}
else if ((cast<uint16_t>((op & cast<uint16_t>(62976ULL))) == cast<uint16_t>(38400ULL))){
int64_t s = cast<int64_t>(cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(11ULL))) & cast<uint16_t>(1ULL))));
int64_t d = cast<int64_t>(cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(8ULL))) & cast<uint16_t>(1ULL))));
gcdsp_CPU_setAc(c,d,gcdsp_CPU_prod(c));
gcdsp_CPU_setArithFlags(c,d);
gcdsp_CPU_setProd(c,gcdsp_CPU_mul16(c,c->Reg[cast<int64_t>((cast<int64_t>(24ULL) + s))],c->Reg[cast<int64_t>((cast<int64_t>(26ULL) + s))]));
}
else if ((cast<uint16_t>((op & cast<uint16_t>(62976ULL))) == cast<uint16_t>(37376ULL))){
int64_t s = cast<int64_t>(cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(11ULL))) & cast<uint16_t>(1ULL))));
int64_t d = cast<int64_t>(cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(8ULL))) & cast<uint16_t>(1ULL))));
gcdsp_CPU_setAc(c,d,(gcdsp_CPU_prod(c) & ~(cast<int64_t>(65535ULL))));
gcdsp_CPU_setArithFlags(c,d);
gcdsp_CPU_setProd(c,gcdsp_CPU_mul16(c,c->Reg[cast<int64_t>((cast<int64_t>(24ULL) + s))],c->Reg[cast<int64_t>((cast<int64_t>(26ULL) + s))]));
}
else if ((cast<uint16_t>((op & cast<uint16_t>(59136ULL))) == cast<uint16_t>(49152ULL))){
int64_t s1 = cast<int64_t>(cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(12ULL))) & cast<uint16_t>(1ULL))));
int64_t s2 = cast<int64_t>(cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(11ULL))) & cast<uint16_t>(1ULL))));
gcdsp_CPU_setProd(c,gcdsp_CPU_mul16(c,c->Reg[cast<int64_t>((cast<int64_t>(30ULL) + s1))],c->Reg[cast<int64_t>((cast<int64_t>(26ULL) + s2))]));
}
else if ((cast<uint16_t>((op & cast<uint16_t>(58880ULL))) == cast<uint16_t>(49664ULL))){
int64_t s1 = cast<int64_t>(cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(12ULL))) & cast<uint16_t>(1ULL))));
int64_t s2 = cast<int64_t>(cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(11ULL))) & cast<uint16_t>(1ULL))));
int64_t r = cast<int64_t>(cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(8ULL))) & cast<uint16_t>(1ULL))));
gcdsp_CPU_setAc(c,r,(gcdsp_CPU_prod(c) & ~(cast<int64_t>(65535ULL))));
gcdsp_CPU_setArithFlags(c,r);
gcdsp_CPU_setProd(c,gcdsp_CPU_mul16(c,c->Reg[cast<int64_t>((cast<int64_t>(30ULL) + s1))],c->Reg[cast<int64_t>((cast<int64_t>(26ULL) + s2))]));
}
else if ((cast<uint16_t>((op & cast<uint16_t>(58880ULL))) == cast<uint16_t>(50176ULL))){
int64_t s1 = cast<int64_t>(cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(12ULL))) & cast<uint16_t>(1ULL))));
int64_t s2 = cast<int64_t>(cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(11ULL))) & cast<uint16_t>(1ULL))));
int64_t r = cast<int64_t>(cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(8ULL))) & cast<uint16_t>(1ULL))));
gcdsp_CPU_setAc(c,r,cast<int64_t>((gcdsp_CPU_ac(c,r) + gcdsp_CPU_prod(c))));
gcdsp_CPU_setArithFlags(c,r);
gcdsp_CPU_setProd(c,gcdsp_CPU_mul16(c,c->Reg[cast<int64_t>((cast<int64_t>(30ULL) + s1))],c->Reg[cast<int64_t>((cast<int64_t>(26ULL) + s2))]));
}
else if ((cast<uint16_t>((op & cast<uint16_t>(58880ULL))) == cast<uint16_t>(50688ULL))){
int64_t s1 = cast<int64_t>(cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(12ULL))) & cast<uint16_t>(1ULL))));
int64_t s2 = cast<int64_t>(cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(11ULL))) & cast<uint16_t>(1ULL))));
int64_t r = cast<int64_t>(cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(8ULL))) & cast<uint16_t>(1ULL))));
gcdsp_CPU_setAc(c,r,gcdsp_CPU_prod(c));
gcdsp_CPU_setArithFlags(c,r);
gcdsp_CPU_setProd(c,gcdsp_CPU_mul16(c,c->Reg[cast<int64_t>((cast<int64_t>(30ULL) + s1))],c->Reg[cast<int64_t>((cast<int64_t>(26ULL) + s2))]));
}
else if ((cast<uint16_t>((op & cast<uint16_t>(64512ULL))) == cast<uint16_t>(57344ULL))){
gcdsp_CPU_setProd(c,cast<int64_t>((gcdsp_CPU_prod(c) + gcdsp_CPU_mul16(c,c->Reg[cast<uint16_t>((cast<uint16_t>(24ULL) + cast<uint16_t>(((cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(9ULL))) & cast<uint16_t>(1ULL)))) * cast<uint16_t>(2ULL)))))],c->Reg[cast<uint16_t>((cast<uint16_t>(25ULL) + cast<uint16_t>(((cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(8ULL))) & cast<uint16_t>(1ULL)))) * cast<uint16_t>(2ULL)))))]))));
}
else if ((cast<uint16_t>((op & cast<uint16_t>(64512ULL))) == cast<uint16_t>(58368ULL))){
gcdsp_CPU_setProd(c,cast<int64_t>((gcdsp_CPU_prod(c) - gcdsp_CPU_mul16(c,c->Reg[cast<uint16_t>((cast<uint16_t>(24ULL) + cast<uint16_t>(((cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(9ULL))) & cast<uint16_t>(1ULL)))) * cast<uint16_t>(2ULL)))))],c->Reg[cast<uint16_t>((cast<uint16_t>(25ULL) + cast<uint16_t>(((cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(8ULL))) & cast<uint16_t>(1ULL)))) * cast<uint16_t>(2ULL)))))]))));
}
else if ((cast<uint16_t>((op & cast<uint16_t>(64512ULL))) == cast<uint16_t>(59392ULL))){
gcdsp_CPU_setProd(c,cast<int64_t>((gcdsp_CPU_prod(c) + gcdsp_CPU_mul16(c,c->Reg[cast<uint16_t>((cast<uint16_t>(30ULL) + (cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(9ULL))) & cast<uint16_t>(1ULL))))))],c->Reg[cast<uint16_t>((cast<uint16_t>(26ULL) + (cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(8ULL))) & cast<uint16_t>(1ULL))))))]))));
}
else if ((cast<uint16_t>((op & cast<uint16_t>(64512ULL))) == cast<uint16_t>(60416ULL))){
gcdsp_CPU_setProd(c,cast<int64_t>((gcdsp_CPU_prod(c) - gcdsp_CPU_mul16(c,c->Reg[cast<uint16_t>((cast<uint16_t>(30ULL) + (cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(9ULL))) & cast<uint16_t>(1ULL))))))],c->Reg[cast<uint16_t>((cast<uint16_t>(26ULL) + (cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(8ULL))) & cast<uint16_t>(1ULL))))))]))));
}
else if ((cast<uint16_t>((op & cast<uint16_t>(59136ULL))) == cast<uint16_t>(40960ULL))){
gcdsp_CPU_setProd(c,gcdsp_CPU_mulxProd(c,op));
}
else if ((cast<uint16_t>((op & cast<uint16_t>(58880ULL))) == cast<uint16_t>(41984ULL))){
int64_t d = cast<int64_t>(cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(8ULL))) & cast<uint16_t>(1ULL))));
int64_t p = gcdsp_CPU_mulxProd(c,op);
gcdsp_CPU_setAc(c,d,cast<int64_t>((gcdsp_CPU_ac(c,d) + gcdsp_CPU_prod(c))));
gcdsp_CPU_setArithFlags(c,d);
gcdsp_CPU_setProd(c,p);
}
else if ((cast<uint16_t>((op & cast<uint16_t>(58880ULL))) == cast<uint16_t>(42496ULL))){
int64_t d = cast<int64_t>(cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(8ULL))) & cast<uint16_t>(1ULL))));
int64_t p = gcdsp_CPU_mulxProd(c,op);
gcdsp_CPU_setAc(c,d,gcdsp_CPU_prod(c));
gcdsp_CPU_setArithFlags(c,d);
gcdsp_CPU_setProd(c,p);
}
else if ((cast<uint16_t>((op & cast<uint16_t>(58880ULL))) == cast<uint16_t>(41472ULL))){
int64_t d = cast<int64_t>(cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(8ULL))) & cast<uint16_t>(1ULL))));
int64_t p = gcdsp_CPU_mulxProd(c,op);
gcdsp_CPU_setAc(c,d,(gcdsp_CPU_prod(c) & ~(cast<int64_t>(65535ULL))));
gcdsp_CPU_setArithFlags(c,d);
gcdsp_CPU_setProd(c,p);
}
else if ((cast<uint16_t>((op & cast<uint16_t>(65024ULL))) == cast<uint16_t>(61952ULL))){
int64_t s = cast<int64_t>(cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(8ULL))) & cast<uint16_t>(1ULL))));
gcdsp_CPU_setProd(c,cast<int64_t>((gcdsp_CPU_prod(c) + gcdsp_CPU_mul16(c,c->Reg[cast<int64_t>((cast<int64_t>(24ULL) + s))],c->Reg[cast<int64_t>((cast<int64_t>(26ULL) + s))]))));
}
else if ((cast<uint16_t>((op & cast<uint16_t>(65024ULL))) == cast<uint16_t>(62976ULL))){
int64_t s = cast<int64_t>(cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(8ULL))) & cast<uint16_t>(1ULL))));
gcdsp_CPU_setProd(c,cast<int64_t>((gcdsp_CPU_prod(c) - gcdsp_CPU_mul16(c,c->Reg[cast<int64_t>((cast<int64_t>(24ULL) + s))],c->Reg[cast<int64_t>((cast<int64_t>(26ULL) + s))]))));
}
else if ((cast<uint16_t>((op & cast<uint16_t>(65024ULL))) == cast<uint16_t>(19968ULL))){
int64_t d = cast<int64_t>(cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(8ULL))) & cast<uint16_t>(1ULL))));
gcdsp_CPU_setAc(c,d,cast<int64_t>((gcdsp_CPU_ac(c,d) + gcdsp_CPU_prod(c))));
gcdsp_CPU_setArithFlags(c,d);
}
else if ((cast<uint16_t>((op & cast<uint16_t>(65024ULL))) == cast<uint16_t>(28160ULL))){
int64_t d = cast<int64_t>(cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(8ULL))) & cast<uint16_t>(1ULL))));
gcdsp_CPU_setAc(c,d,gcdsp_CPU_prod(c));
gcdsp_CPU_setArithFlags(c,d);
}
else if ((cast<uint16_t>((op & cast<uint16_t>(65024ULL))) == cast<uint16_t>(32256ULL))){
int64_t d = cast<int64_t>(cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(8ULL))) & cast<uint16_t>(1ULL))));
gcdsp_CPU_setAc(c,d,cast<int64_t>(-gcdsp_CPU_prod(c)));
gcdsp_CPU_setArithFlags(c,d);
}
else if ((cast<uint16_t>((op & cast<uint16_t>(65024ULL))) == cast<uint16_t>(65024ULL))){
int64_t d = cast<int64_t>(cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(8ULL))) & cast<uint16_t>(1ULL))));
gcdsp_CPU_setAc(c,d,gcdsp_CPU_prodRounded(c));
gcdsp_CPU_setArithFlags(c,d);
}
else if ((cast<uint16_t>((op & cast<uint16_t>(64512ULL))) == cast<uint16_t>(63488ULL))){
int64_t d = cast<int64_t>(cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(8ULL))) & cast<uint16_t>(1ULL))));
int64_t s = cast<int64_t>(cast<uint16_t>(((shr<uint16_t>(op,cast<int64_t>(9ULL))) & cast<uint16_t>(1ULL))));
int64_t oldProd = gcdsp_CPU_prod(c);
int64_t v = cast<int64_t>((gcdsp_CPU_prodRounded(c) + ((gcdsp_CPU_ax(c,s) & ~(cast<int64_t>(65535ULL))))));
gcdsp_CPU_setAc(c,d,v);
int64_t res = gcdsp_CPU_ac(c,d);
gcdsp_CPU_setFlag(c,cast<uint16_t>(4ULL),(res == cast<int64_t>(0ULL)));
gcdsp_CPU_setFlag(c,cast<uint16_t>(8ULL),(res < cast<int64_t>(0ULL)));
gcdsp_CPU_setFlag(c,cast<uint16_t>(1ULL),(cast<uint64_t>(oldProd) > cast<uint64_t>(res)));
gcdsp_CPU_setFlag(c,cast<uint16_t>(2ULL),false);
}
else {
gcdsp_CPU_Halt(c,std::string("unmodelled DSP arithmetic op 0x%04X at 0x%04X (%s)",50),op,pc,gcdsp_mustText([=](auto...args){return gcdsp_CPU_imem(c,args...);},pc));
return cast<uint16_t>(1ULL);
}
}
tmp14:;
gcdsp_extPending_commit(p,c);
return cast<uint16_t>(1ULL);
}
}
// tools/cpu/gcdsp/exec.go:605:1
std::string gcdsp_mustText(std::function<uint16_t(uint16_t)> read,uint16_t pc){
{
auto tmp15 = gcdsp_Disasm(read,pc);
std::string t = std::get<0>(tmp15);
return t;
}
}
// tools/cpu/gcdsp/exec_ext.go:33:1
void gcdsp_extPending_add(gcdsp_extPending* p,uint16_t reg,uint16_t val){
{
if ((p->n < cast<int64_t>(2ULL))) {
p->reg[p->n] = reg;
p->val[p->n] = val;
p->n++;
}
}
}
// tools/cpu/gcdsp/exec_ext.go:41:1
void gcdsp_extPending_commit(gcdsp_extPending p,gcdsp_CPU* c){
{
{int64_t i = cast<int64_t>(0ULL);for (;(i < p.n);i++){
gcdsp_CPU_setReg(c,p.reg[i],p.val[i]);
}
}}
}
// tools/cpu/gcdsp/exec_ext.go:49:1
void gcdsp_CPU_extStep(gcdsp_CPU* c,int64_t n,bool byIndex){
{
if (byIndex) {
gcdsp_CPU_arAdd(c,n,cast<int16_t>(c->Reg[cast<int64_t>((cast<int64_t>(4ULL) + n))]));
}
else {
gcdsp_CPU_arInc(c,n);
}
}
}
// tools/cpu/gcdsp/exec_ext.go:61:1
gcdsp_extPending gcdsp_CPU_extBegin(gcdsp_CPU* c,uint16_t ext){
{
gcdsp_extPending p={};
{
if ((ext == cast<uint16_t>(0ULL))){
}
else if ((cast<uint16_t>((ext & cast<uint16_t>(192ULL))) == cast<uint16_t>(128ULL))){
uint16_t reg = cast<uint16_t>((cast<uint16_t>(24ULL) + (cast<uint16_t>(((shr<uint16_t>(ext,cast<int64_t>(4ULL))) & cast<uint16_t>(3ULL))))));
uint16_t accm = cast<uint16_t>((cast<uint16_t>(30ULL) + (cast<uint16_t>((ext & cast<uint16_t>(1ULL))))));
gcdsp_CPU_dataWrite(c,c->Reg[cast<int64_t>(3ULL)],gcdsp_CPU_getReg(c,accm));
gcdsp_extPending_add(&(p),reg,gcdsp_CPU_dataRead(c,c->Reg[cast<int64_t>(0ULL)]));
gcdsp_CPU_extStep(c,cast<int64_t>(0ULL),(cast<uint16_t>((ext & cast<uint16_t>(4ULL))) != cast<uint16_t>(0ULL)));
gcdsp_CPU_extStep(c,cast<int64_t>(3ULL),(cast<uint16_t>((ext & cast<uint16_t>(8ULL))) != cast<uint16_t>(0ULL)));
}
else if ((cast<uint16_t>((ext & cast<uint16_t>(192ULL))) == cast<uint16_t>(64ULL))){
uint16_t reg = cast<uint16_t>((cast<uint16_t>(24ULL) + (cast<uint16_t>(((shr<uint16_t>(ext,cast<int64_t>(3ULL))) & cast<uint16_t>(7ULL))))));
int64_t prg = cast<int64_t>(cast<uint16_t>((ext & cast<uint16_t>(3ULL))));
gcdsp_extPending_add(&(p),reg,gcdsp_CPU_dataRead(c,c->Reg[cast<int64_t>((cast<int64_t>(0ULL) + prg))]));
gcdsp_CPU_extStep(c,prg,(cast<uint16_t>((ext & cast<uint16_t>(4ULL))) != cast<uint16_t>(0ULL)));
}
else if ((cast<uint16_t>((ext & cast<uint16_t>(224ULL))) == cast<uint16_t>(32ULL))){
int64_t prg = cast<int64_t>(cast<uint16_t>((ext & cast<uint16_t>(3ULL))));
uint16_t src = cast<uint16_t>((cast<uint16_t>(28ULL) + (cast<uint16_t>(((shr<uint16_t>(ext,cast<int64_t>(3ULL))) & cast<uint16_t>(3ULL))))));
gcdsp_CPU_dataWrite(c,c->Reg[cast<int64_t>((cast<int64_t>(0ULL) + prg))],gcdsp_CPU_getReg(c,src));
gcdsp_CPU_extStep(c,prg,(cast<uint16_t>((ext & cast<uint16_t>(4ULL))) != cast<uint16_t>(0ULL)));
}
else if ((cast<uint16_t>((ext & cast<uint16_t>(240ULL))) == cast<uint16_t>(16ULL))){
uint16_t dst = cast<uint16_t>((cast<uint16_t>(24ULL) + (cast<uint16_t>(((shr<uint16_t>(ext,cast<int64_t>(2ULL))) & cast<uint16_t>(3ULL))))));
uint16_t src = cast<uint16_t>((cast<uint16_t>(28ULL) + (cast<uint16_t>((ext & cast<uint16_t>(3ULL))))));
gcdsp_extPending_add(&(p),dst,gcdsp_CPU_getReg(c,src));
}
else if ((cast<uint16_t>((ext & cast<uint16_t>(192ULL))) == cast<uint16_t>(192ULL))){
if ((cast<uint16_t>((ext & cast<uint16_t>(3ULL))) == cast<uint16_t>(3ULL))) {
int64_t s = cast<int64_t>(cast<uint16_t>(((shr<uint16_t>(ext,cast<int64_t>(5ULL))) & cast<uint16_t>(1ULL))));
uint16_t r = cast<uint16_t>(((shr<uint16_t>(ext,cast<int64_t>(4ULL))) & cast<uint16_t>(1ULL)));
gcdsp_extPending_add(&(p),cast<uint16_t>((cast<uint16_t>(26ULL) + r)),gcdsp_CPU_dataRead(c,c->Reg[cast<int64_t>((cast<int64_t>(0ULL) + s))]));
gcdsp_extPending_add(&(p),cast<uint16_t>((cast<uint16_t>(24ULL) + r)),gcdsp_CPU_dataRead(c,c->Reg[cast<int64_t>(3ULL)]));
gcdsp_CPU_extStep(c,s,(cast<uint16_t>((ext & cast<uint16_t>(4ULL))) != cast<uint16_t>(0ULL)));
gcdsp_CPU_extStep(c,cast<int64_t>(3ULL),(cast<uint16_t>((ext & cast<uint16_t>(8ULL))) != cast<uint16_t>(0ULL)));
}
else {
int64_t s = cast<int64_t>(cast<uint16_t>((ext & cast<uint16_t>(3ULL))));
gcdsp_extPending_add(&(p),cast<uint16_t>((cast<uint16_t>(24ULL) + cast<uint16_t>(((cast<uint16_t>(((shr<uint16_t>(ext,cast<int64_t>(5ULL))) & cast<uint16_t>(1ULL)))) * cast<uint16_t>(2ULL))))),gcdsp_CPU_dataRead(c,c->Reg[cast<int64_t>((cast<int64_t>(0ULL) + s))]));
gcdsp_extPending_add(&(p),cast<uint16_t>((cast<uint16_t>(25ULL) + cast<uint16_t>(((cast<uint16_t>(((shr<uint16_t>(ext,cast<int64_t>(4ULL))) & cast<uint16_t>(1ULL)))) * cast<uint16_t>(2ULL))))),gcdsp_CPU_dataRead(c,c->Reg[cast<int64_t>(3ULL)]));
gcdsp_CPU_extStep(c,s,(cast<uint16_t>((ext & cast<uint16_t>(4ULL))) != cast<uint16_t>(0ULL)));
gcdsp_CPU_extStep(c,cast<int64_t>(3ULL),(cast<uint16_t>((ext & cast<uint16_t>(8ULL))) != cast<uint16_t>(0ULL)));
}
}
else if ((cast<uint16_t>((ext & cast<uint16_t>(252ULL))) == cast<uint16_t>(4ULL))){
gcdsp_CPU_arDec(c,cast<int64_t>(cast<uint16_t>((ext & cast<uint16_t>(3ULL)))));
}
else if ((cast<uint16_t>((ext & cast<uint16_t>(252ULL))) == cast<uint16_t>(8ULL))){
gcdsp_CPU_arInc(c,cast<int64_t>(cast<uint16_t>((ext & cast<uint16_t>(3ULL)))));
}
else if ((cast<uint16_t>((ext & cast<uint16_t>(252ULL))) == cast<uint16_t>(12ULL))){
int64_t n = cast<int64_t>(cast<uint16_t>((ext & cast<uint16_t>(3ULL))));
gcdsp_CPU_arAdd(c,n,cast<int16_t>(c->Reg[cast<int64_t>((cast<int64_t>(4ULL) + n))]));
}
else {
gcdsp_CPU_Halt(c,std::string("unmodelled parallel extension 0x%02X at 0x%04X",46),ext,c->PC);
}
}
tmp16:;
return p;
}
}
// tools/cpu/gcdsp/exec_ext.go:157:1
int64_t gcdsp_CPU_mul16(gcdsp_CPU* c,uint16_t a,uint16_t b){
{
int64_t p = cast<int64_t>((cast<int64_t>(cast<int16_t>(a)) * cast<int64_t>(cast<int16_t>(b))));
if ((cast<uint16_t>((gcdsp_CPU_sr(c) & cast<uint16_t>(8192ULL))) == cast<uint16_t>(0ULL))) {
p = shl<int64_t>(p,cast<int64_t>(1ULL));
}
return p;
}
}
// tools/cpu/gcdsp/exec_ext.go:167:1
int64_t gcdsp_CPU_mulx16(gcdsp_CPU* c,uint16_t a,uint16_t b,bool aHigh,bool bHigh){
{
if ((cast<uint16_t>((gcdsp_CPU_sr(c) & cast<uint16_t>(32768ULL))) == cast<uint16_t>(0ULL))) {
return gcdsp_CPU_mul16(c,a,b);
}
int64_t p={};
{
if (((!aHigh) && (!bHigh))){
p = cast<int64_t>((cast<int64_t>(cast<uint32_t>(a)) * cast<int64_t>(cast<uint32_t>(b))));
}
else if (((!aHigh) && bHigh)){
p = cast<int64_t>((cast<int64_t>(cast<uint32_t>(a)) * cast<int64_t>(cast<int16_t>(b))));
}
else if ((aHigh && (!bHigh))){
p = cast<int64_t>((cast<int64_t>(cast<uint32_t>(b)) * cast<int64_t>(cast<int16_t>(a))));
}
else {
p = cast<int64_t>((cast<int64_t>(cast<int16_t>(a)) * cast<int64_t>(cast<int16_t>(b))));
}
}
tmp17:;
if ((cast<uint16_t>((gcdsp_CPU_sr(c) & cast<uint16_t>(8192ULL))) == cast<uint16_t>(0ULL))) {
p = shl<int64_t>(p,cast<int64_t>(1ULL));
}
return p;
}
}
// tools/cpu/gcdsp/exec_ext.go:190:1
int64_t gcdsp_CPU_mulxProd(gcdsp_CPU* c,uint16_t op){
{
bool aHigh = (cast<uint16_t>((op & cast<uint16_t>(4096ULL))) != cast<uint16_t>(0ULL));
bool bHigh = (cast<uint16_t>((op & cast<uint16_t>(2048ULL))) != cast<uint16_t>(0ULL));
uint16_t a = c->Reg[cast<uint16_t>((cast<uint16_t>(24ULL) + (shr<uint16_t>((cast<uint16_t>((op & cast<uint16_t>(4096ULL)))),cast<int64_t>(11ULL)))))];
uint16_t b = c->Reg[cast<uint16_t>((cast<uint16_t>(25ULL) + (shr<uint16_t>((cast<uint16_t>((op & cast<uint16_t>(2048ULL)))),cast<int64_t>(10ULL)))))];
return gcdsp_CPU_mulx16(c,a,b,aHigh,bHigh);
}
}
// tools/cpu/gcdsp/exec_ext.go:201:1
void gcdsp_CPU_setProd(gcdsp_CPU* c,int64_t v){
{
v &= cast<int64_t>(1099511627775ULL);
c->Reg[cast<int64_t>(20ULL)] = cast<uint16_t>(v);
c->Reg[cast<int64_t>(21ULL)] = cast<uint16_t>(shr<int64_t>(v,cast<int64_t>(16ULL)));
c->Reg[cast<int64_t>(22ULL)] = cast<uint16_t>(shr<int64_t>(v,cast<int64_t>(32ULL)));
c->Reg[cast<int64_t>(23ULL)] = cast<uint16_t>(0ULL);
}
}
// tools/cpu/gcdsp/exec_support.go:12:1
uint16_t gcdsp_CPU_imem(gcdsp_CPU* c,uint16_t a){
{
{
if ((a < cast<uint16_t>(4096ULL))){
return c->IRAM[a];
}
else if ((a >= cast<uint16_t>(32768ULL))){
if ((bool(c->IROM) && (cast<int64_t>(cast<uint16_t>((a - cast<uint16_t>(32768ULL)))) < len(c->IROM)))) {
return c->IROM[cast<uint16_t>((a - cast<uint16_t>(32768ULL)))];
}
gcdsp_CPU_Halt(c,std::string("instruction fetch from boot IROM @0x%04X \342\200\224 IROM not present",61),a);
}
else {
gcdsp_CPU_Halt(c,std::string("instruction fetch @0x%04X \342\200\224 unmapped",38),a);
}
}
tmp18:;
return cast<uint16_t>(0ULL);
}
}
// tools/cpu/gcdsp/exec_support.go:29:1
uint16_t gcdsp_CPU_dataRead(gcdsp_CPU* c,uint16_t a){
{
{
if ((a < cast<uint16_t>(4096ULL))){
return c->DRAM[a];
}
else if ((a < cast<uint16_t>(8192ULL))){
if ((bool(c->DROM) && (cast<int64_t>(cast<uint16_t>((a - cast<uint16_t>(4096ULL)))) < len(c->DROM)))) {
return c->DROM[cast<uint16_t>((a - cast<uint16_t>(4096ULL)))];
}
gcdsp_CPU_Halt(c,std::string("DSP read of coefficient ROM @0x%04X \342\200\224 DROM not present (a resampling table the ucode wants)",93),a);
}
else if ((a >= cast<uint16_t>(65280ULL))){
if (bool(c->bus)) {
return gc_dspBus_HWRead(c->bus,a);
}
gcdsp_CPU_Halt(c,std::string("DSP hardware read @0x%04X \342\200\224 no bus attached",45),a);
}
else {
gcdsp_CPU_Halt(c,std::string("DSP data read @0x%04X \342\200\224 unmapped",34),a);
}
}
tmp19:;
return cast<uint16_t>(0ULL);
}
}
// tools/cpu/gcdsp/exec_support.go:50:1
void gcdsp_CPU_dataWrite(gcdsp_CPU* c,uint16_t a,uint16_t v){
{
{
if ((a < cast<uint16_t>(4096ULL))){
c->DRAM[a] = v;
}
else if ((a >= cast<uint16_t>(65280ULL))){
if (bool(c->bus)) {
gc_dspBus_HWWrite(c->bus,a,v);
return ;
}
gcdsp_CPU_Halt(c,std::string("DSP hardware write @0x%04X \342\200\224 no bus attached",46),a);
}
else {
gcdsp_CPU_Halt(c,std::string("DSP data write @0x%04X = 0x%04X \342\200\224 unmapped",44),a,v);
}
}
tmp20:;
}
}
// tools/cpu/gcdsp/exec_support.go:71:1
uint16_t gcdsp_CPU_getReg(gcdsp_CPU* c,uint16_t r){
{
{
if (((r >= cast<uint16_t>(12ULL)) && (r <= cast<uint16_t>(15ULL)))){
return gcdsp_CPU_pop(c,r);
}
else if (((((r == cast<uint16_t>(30ULL)) || (r == cast<uint16_t>(31ULL)))) && (cast<uint16_t>((gcdsp_CPU_sr(c) & cast<uint16_t>(16384ULL))) != cast<uint16_t>(0ULL)))){
int64_t n = cast<int64_t>(cast<uint16_t>((r - cast<uint16_t>(30ULL))));
int64_t v = gcdsp_CPU_ac(c,n);
if ((v != cast<int64_t>(cast<int32_t>(v)))) {
if ((v > cast<int64_t>(0ULL))) {
return cast<uint16_t>(32767ULL);
}
return cast<uint16_t>(32768ULL);
}
return c->Reg[r];
}
else {
return c->Reg[r];
}
}
tmp21:;
}
}
// tools/cpu/gcdsp/exec_support.go:96:1
void gcdsp_CPU_setReg(gcdsp_CPU* c,uint16_t r,uint16_t v){
{
{
if (((r >= cast<uint16_t>(12ULL)) && (r <= cast<uint16_t>(15ULL)))){
gcdsp_CPU_push(c,r,v);
}
else if (((r == cast<uint16_t>(16ULL)) || (r == cast<uint16_t>(17ULL)))){
c->Reg[r] = cast<uint16_t>(cast<int16_t>(cast<int8_t>(v)));
}
else if ((r == cast<uint16_t>(22ULL))){
c->Reg[r] = cast<uint16_t>((v & cast<uint16_t>(255ULL)));
}
else if ((r == cast<uint16_t>(18ULL))){
c->Reg[r] = cast<uint16_t>((v & cast<uint16_t>(255ULL)));
}
else if ((r == cast<uint16_t>(19ULL))){
c->Reg[r] = (v & ~(cast<uint16_t>(256ULL)));
}
else {
c->Reg[r] = v;
}
}
tmp22:;
}
}
// tools/cpu/gcdsp/exec_support.go:118:1
void gcdsp_CPU_setRegExtend(gcdsp_CPU* c,uint16_t r,uint16_t v){
{
gcdsp_CPU_setReg(c,r,v);
if (((((r == cast<uint16_t>(30ULL)) || (r == cast<uint16_t>(31ULL)))) && (cast<uint16_t>((gcdsp_CPU_sr(c) & cast<uint16_t>(16384ULL))) != cast<uint16_t>(0ULL)))) {
int64_t n = cast<int64_t>(cast<uint16_t>((r - cast<uint16_t>(30ULL))));
if ((cast<uint16_t>((v & cast<uint16_t>(32768ULL))) != cast<uint16_t>(0ULL))) {
c->Reg[cast<int64_t>((cast<int64_t>(16ULL) + n))] = cast<uint16_t>(65535ULL);
}
else {
c->Reg[cast<int64_t>((cast<int64_t>(16ULL) + n))] = cast<uint16_t>(0ULL);
}
c->Reg[cast<int64_t>((cast<int64_t>(28ULL) + n))] = cast<uint16_t>(0ULL);
}
}
}
// tools/cpu/gcdsp/exec_support.go:146:1
void gcdsp_CPU_arInc(gcdsp_CPU* c,int64_t n){
{
auto tmp23 = std::make_tuple(cast<uint32_t>(c->Reg[cast<int64_t>((cast<int64_t>(0ULL) + n))]),cast<uint32_t>(c->Reg[cast<int64_t>((cast<int64_t>(8ULL) + n))]));
uint32_t ar = std::get<0>(tmp23);
uint32_t wr = std::get<1>(tmp23);
uint32_t nar = cast<uint32_t>((ar + cast<uint32_t>(1ULL)));
if (((cast<uint32_t>((nar ^ ar))) > shl<uint32_t>((cast<uint32_t>((wr | cast<uint32_t>(1ULL)))),cast<int64_t>(1ULL)))) {
nar -= cast<uint32_t>((wr + cast<uint32_t>(1ULL)));
}
c->Reg[cast<int64_t>((cast<int64_t>(0ULL) + n))] = cast<uint16_t>(nar);
}
}
// tools/cpu/gcdsp/exec_support.go:156:1
void gcdsp_CPU_arDec(gcdsp_CPU* c,int64_t n){
{
auto tmp24 = std::make_tuple(cast<uint32_t>(c->Reg[cast<int64_t>((cast<int64_t>(0ULL) + n))]),cast<uint32_t>(c->Reg[cast<int64_t>((cast<int64_t>(8ULL) + n))]));
uint32_t ar = std::get<0>(tmp24);
uint32_t wr = std::get<1>(tmp24);
uint32_t nar = cast<uint32_t>((ar + wr));
if ((cast<uint32_t>(((cast<uint32_t>((nar ^ ar))) & (shl<uint32_t>((cast<uint32_t>((wr | cast<uint32_t>(1ULL)))),cast<int64_t>(1ULL))))) > wr)) {
nar -= cast<uint32_t>((wr + cast<uint32_t>(1ULL)));
}
c->Reg[cast<int64_t>((cast<int64_t>(0ULL) + n))] = cast<uint16_t>(nar);
}
}
// tools/cpu/gcdsp/exec_support.go:167:1
void gcdsp_CPU_arAdd(gcdsp_CPU* c,int64_t n,int16_t ix){
{
auto tmp25 = std::make_tuple(cast<uint32_t>(c->Reg[cast<int64_t>((cast<int64_t>(0ULL) + n))]),cast<uint32_t>(c->Reg[cast<int64_t>((cast<int64_t>(8ULL) + n))]));
uint32_t ar = std::get<0>(tmp25);
uint32_t wr = std::get<1>(tmp25);
uint32_t mx = shl<uint32_t>((cast<uint32_t>((wr | cast<uint32_t>(1ULL)))),cast<int64_t>(1ULL));
uint32_t nar = cast<uint32_t>((ar + cast<uint32_t>(cast<int32_t>(ix))));
uint32_t dar = cast<uint32_t>(((cast<uint32_t>((cast<uint32_t>((nar ^ ar)) ^ cast<uint32_t>(cast<int32_t>(ix))))) & mx));
if ((ix >= cast<int16_t>(0ULL))) {
if ((dar > wr)) {
nar -= cast<uint32_t>((wr + cast<uint32_t>(1ULL)));
}
}
else {
if ((cast<uint32_t>(((cast<uint32_t>(((cast<uint32_t>((cast<uint32_t>((nar + wr)) + cast<uint32_t>(1ULL)))) ^ nar))) & dar)) <= wr)) {
nar += cast<uint32_t>((wr + cast<uint32_t>(1ULL)));
}
}
c->Reg[cast<int64_t>((cast<int64_t>(0ULL) + n))] = cast<uint16_t>(nar);
}
}
// tools/cpu/gcdsp/exec_support.go:185:1
void gcdsp_CPU_arSub(gcdsp_CPU* c,int64_t n,int16_t ix){
{
auto tmp26 = std::make_tuple(cast<uint32_t>(c->Reg[cast<int64_t>((cast<int64_t>(0ULL) + n))]),cast<uint32_t>(c->Reg[cast<int64_t>((cast<int64_t>(8ULL) + n))]));
uint32_t ar = std::get<0>(tmp26);
uint32_t wr = std::get<1>(tmp26);
uint32_t mx = shl<uint32_t>((cast<uint32_t>((wr | cast<uint32_t>(1ULL)))),cast<int64_t>(1ULL));
uint32_t nar = cast<uint32_t>((ar - cast<uint32_t>(cast<int32_t>(ix))));
uint32_t dar = cast<uint32_t>(((cast<uint32_t>((cast<uint32_t>((nar ^ ar)) ^ cast<uint32_t>(~cast<uint32_t>(cast<int32_t>(ix)))))) & mx));
if (((ix < cast<int16_t>(0ULL)) && (ix != cast<int16_t>(-32768ULL)))) {
if ((dar > wr)) {
nar -= cast<uint32_t>((wr + cast<uint32_t>(1ULL)));
}
}
else {
if ((cast<uint32_t>(((cast<uint32_t>(((cast<uint32_t>((cast<uint32_t>((nar + wr)) + cast<uint32_t>(1ULL)))) ^ nar))) & dar)) <= wr)) {
nar += cast<uint32_t>((wr + cast<uint32_t>(1ULL)));
}
}
c->Reg[cast<int64_t>((cast<int64_t>(0ULL) + n))] = cast<uint16_t>(nar);
}
}
// tools/cpu/gcdsp/exec_support.go:204:1
void gcdsp_CPU_push(gcdsp_CPU* c,uint16_t reg,uint16_t v){
{
uint16_t i = cast<uint16_t>((reg - cast<uint16_t>(12ULL)));
c->Stacks[i] = append(c->Stacks[i],Slice<uint16_t>{v});
}
}
// tools/cpu/gcdsp/exec_support.go:209:1
uint16_t gcdsp_CPU_pop(gcdsp_CPU* c,uint16_t reg){
{
uint16_t i = cast<uint16_t>((reg - cast<uint16_t>(12ULL)));
if ((len(c->Stacks[i]) == cast<int64_t>(0ULL))) {
gcdsp_CPU_Halt(c,std::string("DSP stack ST%d underflow at 0x%04X",34),i,c->PC);
return cast<uint16_t>(0ULL);
}
uint16_t v = c->Stacks[i][cast<int64_t>((len(c->Stacks[i]) - cast<int64_t>(1ULL)))];
c->Stacks[i] = sub(c->Stacks[i],0,cast<int64_t>((len(c->Stacks[i]) - cast<int64_t>(1ULL))));
return v;
}
}
// tools/cpu/gcdsp/exec_support.go:223:1
void gcdsp_CPU_startLoop(gcdsp_CPU* c,uint16_t start,uint16_t end,uint16_t count){
{
if ((count == cast<uint16_t>(0ULL))) {
auto tmp27 = gcdsp_Disasm([&](uint16_t a)->uint16_t{
return gcdsp_CPU_imem(c,a);
}
,end);
uint16_t words = std::get<1>(tmp27);
c->PC = cast<uint16_t>((end + words));
c->Branched = true;
return ;
}
c->Loops = append(c->Loops,Slice<gcdsp_LoopFrame>{gcdsp_LoopFrame{start,end,count}});
}
}
// tools/cpu/gcdsp/exec_support.go:240:1
void gcdsp_CPU_serviceLoops(gcdsp_CPU* c,uint16_t execAddr,bool branched){
{
if (((len(c->Loops) == cast<int64_t>(0ULL)) || branched)) {
return ;
}
gcdsp_LoopFrame* top = (&c->Loops[cast<int64_t>((len(c->Loops) - cast<int64_t>(1ULL)))]);
if ((execAddr != top->End)) {
return ;
}
top->Count--;
if ((top->Count > cast<uint16_t>(0ULL))) {
c->PC = top->Start;
}
else {
c->Loops = sub(c->Loops,0,cast<int64_t>((len(c->Loops) - cast<int64_t>(1ULL))));
}
}
}
// tools/cpu/gcdsp/exec_support.go:263:1
void gcdsp_CPU_shiftAcc(gcdsp_CPU* c,int64_t r,bool arith,int64_t amt){
{
{
if ((amt >= cast<int64_t>(0ULL))){
uint64_t v = cast<uint64_t>(((shl<uint64_t>(cast<uint64_t>(gcdsp_CPU_ac(c,r)),cast<uint64_t>(amt))) & cast<uint64_t>(1099511627775ULL)));
gcdsp_CPU_setAc(c,r,cast<int64_t>(v));
}
else if (arith){
gcdsp_CPU_setAc(c,r,shr<int64_t>(gcdsp_CPU_ac(c,r),cast<uint64_t>(cast<int64_t>(-amt))));
}
else {
uint64_t v = shr<uint64_t>((cast<uint64_t>((cast<uint64_t>(gcdsp_CPU_ac(c,r)) & cast<uint64_t>(1099511627775ULL)))),cast<uint64_t>(cast<int64_t>(-amt)));
gcdsp_CPU_setAc(c,r,cast<int64_t>(v));
}
}
tmp28:;
gcdsp_CPU_setArithFlags(c,r);
}
}
// tools/cpu/gcdsp/exec_support.go:282:1
int64_t gcdsp_shiftAmount7(uint16_t v){
{
{
if ((cast<uint16_t>((v & cast<uint16_t>(63ULL))) == cast<uint16_t>(0ULL))){
return cast<int64_t>(0ULL);
}
else if ((cast<uint16_t>((v & cast<uint16_t>(64ULL))) != cast<uint16_t>(0ULL))){
return cast<int64_t>((cast<int64_t>(-64ULL) + cast<int64_t>(cast<uint16_t>((v & cast<uint16_t>(63ULL))))));
}
else {
return cast<int64_t>(cast<uint16_t>((v & cast<uint16_t>(63ULL))));
}
}
tmp29:;
}
}
// tools/cpu/gcdsp/exec_support.go:299:1
void gcdsp_CPU_setArithFlags(gcdsp_CPU* c,int64_t n){
{
int64_t v = gcdsp_CPU_ac(c,n);
gcdsp_CPU_setFlag(c,cast<uint16_t>(4ULL),(v == cast<int64_t>(0ULL)));
gcdsp_CPU_setFlag(c,cast<uint16_t>(8ULL),(v < cast<int64_t>(0ULL)));
gcdsp_CPU_setFlag(c,cast<uint16_t>(1ULL),false);
gcdsp_CPU_setFlag(c,cast<uint16_t>(2ULL),false);
}
}
// tools/cpu/gcdsp/exec_support.go:313:1
void gcdsp_CPU_aluAddSub(gcdsp_CPU* c,int64_t d,int64_t b,bool sub_){
{
int64_t a = gcdsp_CPU_ac(c,d);
if (sub_) {
b = cast<int64_t>(-b);
}
gcdsp_CPU_setAc(c,d,cast<int64_t>((a + b)));
int64_t v = gcdsp_CPU_ac(c,d);
gcdsp_CPU_setFlag(c,cast<uint16_t>(4ULL),(v == cast<int64_t>(0ULL)));
gcdsp_CPU_setFlag(c,cast<uint16_t>(8ULL),(v < cast<int64_t>(0ULL)));
if (sub_) {
gcdsp_CPU_setFlag(c,cast<uint16_t>(1ULL),(cast<uint64_t>(a) >= cast<uint64_t>(v)));
}
else {
gcdsp_CPU_setFlag(c,cast<uint16_t>(1ULL),(cast<uint64_t>(a) > cast<uint64_t>(v)));
}
gcdsp_CPU_setFlag(c,cast<uint16_t>(2ULL),(cast<int64_t>(((cast<int64_t>((a ^ v))) & (cast<int64_t>((b ^ v))))) < cast<int64_t>(0ULL)));
if ((cast<uint16_t>((c->Reg[cast<int64_t>(19ULL)] & cast<uint16_t>(2ULL))) != cast<uint16_t>(0ULL))) {
c->Reg[cast<int64_t>(19ULL)] |= cast<uint16_t>(128ULL);
}
}
}
// tools/cpu/gcdsp/exec_support.go:338:1
void gcdsp_CPU_setTestFlags(gcdsp_CPU* c,int64_t n){
{
int64_t v = gcdsp_CPU_ac(c,n);
gcdsp_CPU_setFlag(c,cast<uint16_t>(4ULL),(v == cast<int64_t>(0ULL)));
gcdsp_CPU_setFlag(c,cast<uint16_t>(8ULL),(v < cast<int64_t>(0ULL)));
gcdsp_CPU_setFlag(c,cast<uint16_t>(1ULL),false);
gcdsp_CPU_setFlag(c,cast<uint16_t>(2ULL),false);
}
}
// tools/cpu/gcdsp/exec_support.go:350:1
void gcdsp_CPU_setLogicFlags(gcdsp_CPU* c,int64_t n){
{
int16_t v = cast<int16_t>(c->Reg[cast<int64_t>((cast<int64_t>(30ULL) + n))]);
gcdsp_CPU_setFlag(c,cast<uint16_t>(4ULL),(v == cast<int16_t>(0ULL)));
gcdsp_CPU_setFlag(c,cast<uint16_t>(8ULL),(v < cast<int16_t>(0ULL)));
gcdsp_CPU_setFlag(c,cast<uint16_t>(1ULL),false);
gcdsp_CPU_setFlag(c,cast<uint16_t>(2ULL),false);
}
}
// tools/cpu/gcdsp/exec_support.go:362:1
void gcdsp_CPU_subFlags(gcdsp_CPU* c,int64_t a,int64_t b){
{
int64_t res = shr<int64_t>((shl<int64_t>((cast<int64_t>((a - b))),cast<int64_t>(24ULL))),cast<int64_t>(24ULL));
gcdsp_CPU_setFlag(c,cast<uint16_t>(4ULL),(res == cast<int64_t>(0ULL)));
gcdsp_CPU_setFlag(c,cast<uint16_t>(8ULL),(res < cast<int64_t>(0ULL)));
gcdsp_CPU_setFlag(c,cast<uint16_t>(1ULL),(cast<uint64_t>(a) >= cast<uint64_t>(res)));
gcdsp_CPU_setFlag(c,cast<uint16_t>(2ULL),(cast<int64_t>(((cast<int64_t>((a ^ res))) & (cast<int64_t>((cast<int64_t>(-b) ^ res))))) < cast<int64_t>(0ULL)));
if ((cast<uint16_t>((c->Reg[cast<int64_t>(19ULL)] & cast<uint16_t>(2ULL))) != cast<uint16_t>(0ULL))) {
c->Reg[cast<int64_t>(19ULL)] |= cast<uint16_t>(128ULL);
}
}
}
// tools/cpu/gcdsp/exec_support.go:378:1
bool gcdsp_CPU_cond(gcdsp_CPU* c,uint16_t cc){
{
uint16_t sr = c->Reg[cast<int64_t>(19ULL)];
bool z = (cast<uint16_t>((sr & cast<uint16_t>(4ULL))) != cast<uint16_t>(0ULL));
bool s = (cast<uint16_t>((sr & cast<uint16_t>(8ULL))) != cast<uint16_t>(0ULL));
bool o = (cast<uint16_t>((sr & cast<uint16_t>(2ULL))) != cast<uint16_t>(0ULL));
bool cf = (cast<uint16_t>((sr & cast<uint16_t>(1ULL))) != cast<uint16_t>(0ULL));
bool lz = (cast<uint16_t>((sr & cast<uint16_t>(64ULL))) != cast<uint16_t>(0ULL));
{
switch(cc){
case cast<uint16_t>(0ULL):{
return (s == o);
break;}
case cast<uint16_t>(1ULL):{
return (s != o);
break;}
case cast<uint16_t>(2ULL):{
return ((s == o) && (!z));
break;}
case cast<uint16_t>(3ULL):{
return ((s != o) || z);
break;}
case cast<uint16_t>(4ULL):{
return (!z);
break;}
case cast<uint16_t>(5ULL):{
return z;
break;}
case cast<uint16_t>(6ULL):{
return (!cf);
break;}
case cast<uint16_t>(7ULL):{
return cf;
break;}
case cast<uint16_t>(12ULL):{
return (!lz);
break;}
case cast<uint16_t>(13ULL):{
return lz;
break;}
case cast<uint16_t>(15ULL):{
return true;
break;}
}}
gcdsp_CPU_Halt(c,std::string("unmodelled branch condition 0x%X at 0x%04X",42),cc,c->PC);
return false;
}
}
// tools/platform/gc/boot.go:17:1
std::tuple<uint32_t,Error> gc_Machine_LoadDOL(gc_Machine* m){
uint32_t entry{};
Error err{};
{
auto tmp1 = gc_Disc_DOL(m->disc);
gc_DOL* dol = std::get<0>(tmp1);
err = std::get<1>(tmp1);
if (bool(err)) {
return {cast<uint32_t>(0ULL),err};
}
gc_Machine_setupLowMem(m);
gc_Machine_setupState(m);
gc_DOL_Load(dol,[&](uint32_t addr,Slice<uint8_t> b)->void{
gc_Machine_dmaToRAM(m,cast<uint32_t>((addr & cast<uint32_t>(67108863ULL))),b);
}
);
m->CPU->PC = dol->Entry;
return {dol->Entry,{}};
}
}
// tools/platform/gc/boot.go:34:1
void gc_Machine_Poke(gc_Machine* m,uint32_t addr,uint32_t val){
{
gc_Machine_setRAM32(m,cast<uint32_t>((addr & cast<uint32_t>(67108863ULL))),val);
}
}
// tools/platform/gc/boot.go:43:1
void gc_Machine_PoisonLowMem(gc_Machine* m){
{
{uint32_t a = cast<uint32_t>(0ULL);for (;(a < cast<uint32_t>(12288ULL));a += cast<uint32_t>(4ULL)){
gc_Machine_setRAM32(m,a,cast<uint32_t>((cast<uint32_t>(4027383808ULL) | a)));
}
}}
}
// tools/platform/gc/boot.go:50:1
std::string gc_Machine_String(gc_Machine* m){
{
return go_fmt_Sprintf(std::string("Gekko PC 0x%08X, %d steps",25),m->CPU->PC,m->CPU->Steps);
}
}
// tools/platform/gc/debug.go:22:1
uint8_t gc_Machine_ReadVirt8(gc_Machine* m,uint32_t addr){
{
return gekko_CPU_ReadMem(m->CPU,addr);
}
}
// tools/platform/gc/debug.go:24:1
uint32_t gc_Machine_ReadVirt32(gc_Machine* m,uint32_t addr){
{
return cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((shl<uint32_t>(cast<uint32_t>(gekko_CPU_ReadMem(m->CPU,addr)),cast<int64_t>(24ULL)) | shl<uint32_t>(cast<uint32_t>(gekko_CPU_ReadMem(m->CPU,cast<uint32_t>((addr + cast<uint32_t>(1ULL))))),cast<int64_t>(16ULL)))) | shl<uint32_t>(cast<uint32_t>(gekko_CPU_ReadMem(m->CPU,cast<uint32_t>((addr + cast<uint32_t>(2ULL))))),cast<int64_t>(8ULL)))) | cast<uint32_t>(gekko_CPU_ReadMem(m->CPU,cast<uint32_t>((addr + cast<uint32_t>(3ULL)))))));
}
}
// tools/platform/gc/debug.go:31:1
uint64_t gc_Machine_VIField(gc_Machine* m){
{
return m->vi.Field;
}
}
// tools/platform/gc/debug.go:36:1
Slice<uint32_t> gc_Machine_Backtrace(gc_Machine* m){
{
return gc_Machine_backtraceFrom(m,m->CPU->GPR[cast<int64_t>(1ULL)]);
}
}
// tools/platform/gc/debug.go:55:1
Slice<uint32_t> gc_Machine_backtraceFrom(gc_Machine* m,uint32_t sp){
{
Slice<uint32_t> out={};
{int64_t depth = cast<int64_t>(0ULL);for (;(depth < cast<int64_t>(64ULL));depth++){
if ((!gc_validSP(sp))) {
break;
}
uint32_t next = gc_Machine_stackWord(m,sp);
if (((!gc_validSP(next)) || (next <= sp))) {
break;
}
{
uint32_t lr = gc_Machine_stackWord(m,cast<uint32_t>((next + cast<uint32_t>(4ULL))));
if ((lr != cast<uint32_t>(0ULL))) {
out = append(out,Slice<uint32_t>{lr});
}
}
sp = next;
}
}return out;
}
}
// tools/platform/gc/debug.go:77:1
bool gc_validSP(uint32_t a){
{
return ((((a != cast<uint32_t>(0ULL)) && (cast<uint32_t>((a & cast<uint32_t>(3ULL))) == cast<uint32_t>(0ULL))) && (a >= cast<uint32_t>(2147483648ULL))) && (a < cast<uint32_t>(2172649472ULL)));
}
}
// tools/platform/gc/debug.go:101:1
uint32_t gc_Machine_stackWord(gc_Machine* m,uint32_t ea){
{
uint32_t pa = cast<uint32_t>((ea - cast<uint32_t>(2147483648ULL)));
if ((cast<int64_t>((cast<int64_t>(pa) + cast<int64_t>(4ULL))) > len(m->RAM))) {
return cast<uint32_t>(0ULL);
}
return be_Uint32(rrBorrow(m->RAM,pa,cast<uint32_t>((pa + cast<uint32_t>(4ULL)))));
}
}
// tools/platform/gc/debug.go:110:1
std::string gc_Machine_BacktraceString(gc_Machine* m){
{
Slice<uint32_t> bt = gc_Machine_Backtrace(m);
if ((len(bt) == cast<int64_t>(0ULL))) {
return std::string("  (no frames)",13);
}
std::string s = std::string("",0);
{auto&& tmp2 = bt;
for(int64_t tmp3=0;tmp3<len(tmp2);++tmp3){
auto i=tmp3;auto a=tmp2[tmp3];s += go_fmt_Sprintf(std::string("  #%d  0x%08X\012",14),i,a);
}}
return s;
}
}
// tools/platform/gc/debug.go:128:1
std::string gc_Machine_ThreadsString(gc_Machine* m){
{
strings_Builder b={};
Map<uint32_t,bool> seen = Map<uint32_t,bool>{};
{uint32_t t = gc_Machine_ReadVirt32(m,cast<uint32_t>(2147483872ULL));for (;(((t >= cast<uint32_t>(2147483648ULL)) && (t < cast<uint32_t>(2172649472ULL))) && (!get(seen,t)));t = gc_Machine_ReadVirt32(m,cast<uint32_t>((t + cast<uint32_t>(768ULL))))){
seen[t] = true;
uint32_t state = shr<uint32_t>(gc_Machine_ReadVirt32(m,cast<uint32_t>((t + cast<uint32_t>(712ULL)))),cast<int64_t>(16ULL));
uint32_t queue = gc_Machine_ReadVirt32(m,cast<uint32_t>((t + cast<uint32_t>(732ULL))));
uint32_t srr0 = gc_Machine_ReadVirt32(m,cast<uint32_t>((t + cast<uint32_t>(408ULL))));
uint32_t lr = gc_Machine_ReadVirt32(m,cast<uint32_t>((t + cast<uint32_t>(132ULL))));
uint32_t sp = gc_Machine_ReadVirt32(m,cast<uint32_t>((t + cast<uint32_t>(4ULL))));
go_fmt_Fprintf((&b),std::string("thread 0x%08X state=0x%04X queue=0x%08X SRR0=0x%08X LR=0x%08X\012",62),t,state,queue,srr0,lr);
{auto&& tmp4 = gc_Machine_backtraceFrom(m,sp);
for(int64_t tmp5=0;tmp5<len(tmp4);++tmp5){
auto i=tmp5;auto a=tmp4[tmp5];go_fmt_Fprintf((&b),std::string("    #%d  0x%08X\012",16),i,a);
if ((i >= cast<int64_t>(9ULL))) {
break;
}
}}
}
}if ((strings_Builder_Len(&(b)) == cast<int64_t>(0ULL))) {
return std::string("  (no threads found at 0x800000E0)\012",35);
}
return strings_Builder_String(&(b));
}
}
// tools/platform/gc/debug.go:160:1
std::tuple<image_RGBA*,Error> gc_Machine_RenderXFB(gc_Machine* m){
{
uint32_t addr = gc_vi_XFBAddr(&(m->vi));
if ((addr == cast<uint32_t>(0ULL))) {
return {{},go_fmt_Errorf(std::string("VI has no framebuffer address yet (the game has not set TFBL)",61))};
}
auto tmp6 = gc_Machine_xfbSize(m);
int64_t w = std::get<0>(tmp6);
int64_t h = std::get<1>(tmp6);
image_RGBA* img = go_image_NewRGBA(go_image_Rect(cast<int64_t>(0ULL),cast<int64_t>(0ULL),w,h));
int64_t stride = cast<int64_t>((w * cast<int64_t>(2ULL)));
{int64_t y = cast<int64_t>(0ULL);for (;(y < h);y++){
uint32_t row = cast<uint32_t>((addr + cast<uint32_t>(cast<int64_t>((y * stride)))));
{int64_t x = cast<int64_t>(0ULL);for (;(x < w);x += cast<int64_t>(2ULL)){
uint32_t o = cast<uint32_t>((row + cast<uint32_t>(cast<int64_t>((x * cast<int64_t>(2ULL))))));
if ((cast<int64_t>((cast<int64_t>(o) + cast<int64_t>(3ULL))) >= len(m->RAM))) {
break;
}
uint8_t y0 = m->RAM[o];
uint8_t cb = m->RAM[cast<uint32_t>((o + cast<uint32_t>(1ULL)))];
uint8_t y1 = m->RAM[cast<uint32_t>((o + cast<uint32_t>(2ULL)))];
uint8_t cr = m->RAM[cast<uint32_t>((o + cast<uint32_t>(3ULL)))];
auto tmp7 = gc_yuv2rgb(y0,cb,cr);
uint8_t r0 = std::get<0>(tmp7);
uint8_t g0 = std::get<1>(tmp7);
uint8_t b0 = std::get<2>(tmp7);
auto tmp8 = gc_yuv2rgb(y1,cb,cr);
uint8_t r1 = std::get<0>(tmp8);
uint8_t g1 = std::get<1>(tmp8);
uint8_t b1 = std::get<2>(tmp8);
gc_setPix(img,x,y,r0,g0,b0);
gc_setPix(img,cast<int64_t>((x + cast<int64_t>(1ULL))),y,r1,g1,b1);
}
}}
}return {img,{}};
}
}
// tools/platform/gc/debug.go:191:1
std::tuple<int64_t,int64_t> gc_Machine_xfbSize(gc_Machine* m){
{
return {cast<int64_t>(640ULL),cast<int64_t>(480ULL)};
}
}
// tools/platform/gc/debug.go:199:1
std::tuple<image_RGBA*,Error> gc_Machine_RenderEFB(gc_Machine* m){
{
if ((!m->gpu.EFB)) {
return {{},go_fmt_Errorf(std::string("the EFB has not been drawn to yet",33))};
}
image_RGBA* img = go_image_NewRGBA(go_image_Rect(cast<int64_t>(0ULL),cast<int64_t>(0ULL),cast<int64_t>(640ULL),cast<int64_t>(528ULL)));
{int64_t y = cast<int64_t>(0ULL);for (;(y < cast<int64_t>(528ULL));y++){
{int64_t x = cast<int64_t>(0ULL);for (;(x < cast<int64_t>(640ULL));x++){
uint32_t px = m->gpu.EFB[cast<int64_t>((cast<int64_t>((y * cast<int64_t>(640ULL))) + x))];
gc_setPix(img,x,y,cast<uint8_t>(shr<uint32_t>(px,cast<int64_t>(24ULL))),cast<uint8_t>(shr<uint32_t>(px,cast<int64_t>(16ULL))),cast<uint8_t>(shr<uint32_t>(px,cast<int64_t>(8ULL))));
}
}}
}return {img,{}};
}
}
// tools/platform/gc/debug.go:213:1
void gc_setPix(image_RGBA* img,int64_t x,int64_t y,uint8_t r,uint8_t g,uint8_t b){
{
if (((((x < cast<int64_t>(0ULL)) || (y < cast<int64_t>(0ULL))) || (x >= image_Rectangle_Dx(image_RGBA_Bounds(img)))) || (y >= image_Rectangle_Dy(image_RGBA_Bounds(img))))) {
return ;
}
int64_t i = image_RGBA_PixOffset(img,x,y);
auto tmp9 = std::make_tuple(r,g,b,cast<uint8_t>(255ULL));
img->Pix[i] = std::get<0>(tmp9);
img->Pix[cast<int64_t>((i + cast<int64_t>(1ULL)))] = std::get<1>(tmp9);
img->Pix[cast<int64_t>((i + cast<int64_t>(2ULL)))] = std::get<2>(tmp9);
img->Pix[cast<int64_t>((i + cast<int64_t>(3ULL)))] = std::get<3>(tmp9);
}
}
// tools/platform/gc/debug.go:223:1
std::tuple<uint8_t,uint8_t,uint8_t> gc_yuv2rgb(uint8_t y,uint8_t cb,uint8_t cr){
{
double yf = cast<double>(y);
double u = (cast<double>(cb) - cast<double>(128ULL));
double v = (cast<double>(cr) - cast<double>(128ULL));
double r = (yf + (1.371 * v));
double g = ((yf - (0.336 * u)) - (0.698 * v));
double b = (yf + (1.732 * u));
return {gc_clamp8(r),gc_clamp8(g),gc_clamp8(b)};
}
}
// tools/platform/gc/debug.go:233:1
uint8_t gc_clamp8(double v){
{
if ((v < cast<double>(0ULL))) {
return cast<uint8_t>(0ULL);
}
if ((v > cast<double>(255ULL))) {
return cast<uint8_t>(255ULL);
}
return cast<uint8_t>(v);
}
}
// tools/platform/gc/debug.go:247:1
std::string gc_Machine_IntrState(gc_Machine* m){
{
bool viArmed = false;
{auto&& tmp10 = m->vi.DI;
for(int64_t tmp11=0;tmp11<len(tmp10);++tmp11){
auto d=tmp10[tmp11];if ((cast<uint32_t>((d & cast<uint32_t>(268435456ULL))) != cast<uint32_t>(0ULL))) {
viArmed = true;
}
}}
std::string dspState = std::string("none",4);
if (bool(m->dsp.Core)) {
dspState = go_fmt_Sprintf(std::string("PC=0x%04X blocked=%v halt=%v",28),m->dsp.Core->PC,m->dsp.CoreBlocked,m->dsp.Core->Halted);
}
return go_fmt_Sprintf(std::string("PI cause=0x%08X mask=0x%08X | CPU ExtInt=%v MSR[EE]=%v | VI armed=%v field=%d | PE reg0=0x%04X | TFBL=0x%08X XFB=0x%08X | DSP core %s toDSP=0x%08X fromDSP=0x%08X csr=0x%04X | AID start=0x%08X ctrl=0x%04X rem=%d",210),m->pi.Cause,m->pi.Mask,m->CPU->ExtInt,(cast<uint32_t>((m->CPU->MSR & cast<uint32_t>(32768ULL))) != cast<uint32_t>(0ULL)),viArmed,m->vi.Field,m->pe.Reg[cast<int64_t>(0ULL)],m->vi.TFBL,gc_vi_XFBAddr(&(m->vi)),dspState,m->dsp.ToDSP,m->dsp.FromDSP,m->dsp.CSR,m->dsp.AIDStart,m->dsp.AIDControl,m->dsp.AIDRemaining);
}
}
// tools/platform/gc/debug.go:266:1
std::string gc_Machine_RegString(gc_Machine* m){
{
strings_Builder b={};
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(32ULL));i += cast<int64_t>(4ULL)){
go_fmt_Fprintf((&b),std::string("  r%-2d %08X  r%-2d %08X  r%-2d %08X  r%-2d %08X\012",49),i,m->CPU->GPR[i],cast<int64_t>((i + cast<int64_t>(1ULL))),m->CPU->GPR[cast<int64_t>((i + cast<int64_t>(1ULL)))],cast<int64_t>((i + cast<int64_t>(2ULL))),m->CPU->GPR[cast<int64_t>((i + cast<int64_t>(2ULL)))],cast<int64_t>((i + cast<int64_t>(3ULL))),m->CPU->GPR[cast<int64_t>((i + cast<int64_t>(3ULL)))]);
}
}go_fmt_Fprintf((&b),std::string("  PC %08X  LR %08X  CTR %08X  SRR0 %08X  SRR1 %08X\012",51),m->CPU->PC,m->CPU->LR,m->CPU->CTR,m->CPU->SRR0,m->CPU->SRR1);
return strings_Builder_String(&(b));
}
}
// tools/platform/gc/devices.go:19:1
uint32_t gc_mi_read(gc_mi* d,gc_Machine* m,uint32_t off,int64_t size){
{
uint32_t i = divi<uint32_t>((cast<uint32_t>((off & cast<uint32_t>(4095ULL)))),cast<uint32_t>(2ULL));
if ((cast<int64_t>(i) < cast<int64_t>(64ULL))) {
return cast<uint32_t>(d->Reg[i]);
}
gc_Machine_logf(m,std::string("MI read unmodelled 0x%03X",25),cast<uint32_t>((off & cast<uint32_t>(4095ULL))));
return cast<uint32_t>(0ULL);
}
}
// tools/platform/gc/devices.go:28:1
void gc_mi_write(gc_mi* d,gc_Machine* m,uint32_t off,uint32_t v,int64_t size){
{
uint32_t i = divi<uint32_t>((cast<uint32_t>((off & cast<uint32_t>(4095ULL)))),cast<uint32_t>(2ULL));
if ((cast<int64_t>(i) < cast<int64_t>(64ULL))) {
d->Reg[i] = cast<uint16_t>(v);
return ;
}
gc_Machine_logf(m,std::string("MI write unmodelled 0x%03X = 0x%08X",35),cast<uint32_t>((off & cast<uint32_t>(4095ULL))),v);
}
}
// tools/platform/gc/devices.go:52:1
uint32_t gc_ai_read(gc_ai* d,gc_Machine* m,uint32_t off,int64_t size){
{
{
switch(cast<uint32_t>((off & cast<uint32_t>(255ULL)))){
case cast<uint32_t>(0ULL):{
return d->Control;
break;}
case cast<uint32_t>(4ULL):{
return d->Volume;
break;}
case cast<uint32_t>(8ULL):{
d->SCnt++;
return d->SCnt;
break;}
case cast<uint32_t>(12ULL):{
return d->ITCnt;
break;}
}}
gc_Machine_logf(m,std::string("AI read unmodelled 0x%02X",25),cast<uint32_t>((off & cast<uint32_t>(255ULL))));
return cast<uint32_t>(0ULL);
}
}
// tools/platform/gc/devices.go:68:1
void gc_ai_write(gc_ai* d,gc_Machine* m,uint32_t off,uint32_t v,int64_t size){
{
{
switch(cast<uint32_t>((off & cast<uint32_t>(255ULL)))){
case cast<uint32_t>(0ULL):{
d->Control = v;
break;}
case cast<uint32_t>(4ULL):{
d->Volume = v;
break;}
case cast<uint32_t>(8ULL):{
d->SCnt = v;
break;}
case cast<uint32_t>(12ULL):{
d->ITCnt = v;
break;}
default:{
gc_Machine_logf(m,std::string("AI write unmodelled 0x%02X = 0x%08X",35),cast<uint32_t>((off & cast<uint32_t>(255ULL))),v);
break;}
}}
}
}
// tools/platform/gc/devices.go:97:1
uint32_t gc_cp_read(gc_cp* d,gc_Machine* m,uint32_t off,int64_t size){
{
{
switch(cast<uint32_t>((off & cast<uint32_t>(4095ULL)))){
case cast<uint32_t>(0ULL):{
return cast<uint32_t>(6ULL);
break;}
case cast<uint32_t>(2ULL):{
return cast<uint32_t>(d->Control);
break;}
}}
uint32_t i = divi<uint32_t>((cast<uint32_t>((off & cast<uint32_t>(4095ULL)))),cast<uint32_t>(2ULL));
if ((cast<int64_t>(i) < cast<int64_t>(64ULL))) {
return cast<uint32_t>(d->Reg[i]);
}
gc_Machine_logf(m,std::string("CP read unmodelled 0x%03X",25),cast<uint32_t>((off & cast<uint32_t>(4095ULL))));
return cast<uint32_t>(0ULL);
}
}
// tools/platform/gc/devices.go:114:1
void gc_cp_write(gc_cp* d,gc_Machine* m,uint32_t off,uint32_t v,int64_t size){
{
uint32_t i = divi<uint32_t>((cast<uint32_t>((off & cast<uint32_t>(4095ULL)))),cast<uint32_t>(2ULL));
if ((cast<int64_t>(i) < cast<int64_t>(64ULL))) {
d->Reg[i] = cast<uint16_t>(v);
return ;
}
gc_Machine_logf(m,std::string("CP write unmodelled 0x%03X = 0x%08X",35),cast<uint32_t>((off & cast<uint32_t>(4095ULL))),v);
}
}
// tools/platform/gc/devices.go:144:1
uint32_t gc_pe_read(gc_pe* d,gc_Machine* m,uint32_t off,int64_t size){
{
{
switch(cast<uint32_t>((off & cast<uint32_t>(4095ULL)))){
case cast<uint32_t>(10ULL):{
uint16_t v={};
if (d->TokenEnable) {
v |= cast<uint16_t>(1ULL);
}
if (d->FinishEnable) {
v |= cast<uint16_t>(2ULL);
}
if ((cast<uint32_t>((m->pi.Cause & cast<uint32_t>(512ULL))) != cast<uint32_t>(0ULL))) {
v |= cast<uint16_t>(4ULL);
}
if ((cast<uint32_t>((m->pi.Cause & cast<uint32_t>(1024ULL))) != cast<uint32_t>(0ULL))) {
v |= cast<uint16_t>(8ULL);
}
return cast<uint32_t>(v);
break;}
case cast<uint32_t>(14ULL):{
return cast<uint32_t>(d->Token);
break;}
}}
uint32_t i = divi<uint32_t>((cast<uint32_t>((off & cast<uint32_t>(4095ULL)))),cast<uint32_t>(2ULL));
if ((cast<int64_t>(i) < cast<int64_t>(32ULL))) {
return cast<uint32_t>(d->Reg[i]);
}
gc_Machine_logf(m,std::string("PE read unmodelled 0x%03X",25),cast<uint32_t>((off & cast<uint32_t>(4095ULL))));
return cast<uint32_t>(0ULL);
}
}
// tools/platform/gc/devices.go:174:1
void gc_pe_write(gc_pe* d,gc_Machine* m,uint32_t off,uint32_t v,int64_t size){
{
{
switch(cast<uint32_t>((off & cast<uint32_t>(4095ULL)))){
case cast<uint32_t>(10ULL):{
d->TokenEnable = (cast<uint32_t>((v & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL));
d->FinishEnable = (cast<uint32_t>((v & cast<uint32_t>(2ULL))) != cast<uint32_t>(0ULL));
if ((cast<uint32_t>((v & cast<uint32_t>(4ULL))) != cast<uint32_t>(0ULL))) {
gc_Machine_clearInt(m,cast<int64_t>(9ULL));
}
if ((cast<uint32_t>((v & cast<uint32_t>(8ULL))) != cast<uint32_t>(0ULL))) {
gc_Machine_clearInt(m,cast<int64_t>(10ULL));
}
return ;
break;}
}}
uint32_t i = divi<uint32_t>((cast<uint32_t>((off & cast<uint32_t>(4095ULL)))),cast<uint32_t>(2ULL));
if ((cast<int64_t>(i) < cast<int64_t>(32ULL))) {
d->Reg[i] = cast<uint16_t>(v);
return ;
}
gc_Machine_logf(m,std::string("PE write unmodelled 0x%03X = 0x%08X",35),cast<uint32_t>((off & cast<uint32_t>(4095ULL))),v);
}
}
// tools/platform/gc/devices.go:198:1
void gc_pe_setFinish(gc_pe* d,gc_Machine* m){
{
if (d->FinishEnable) {
gc_Machine_raiseInt(m,cast<int64_t>(10ULL));
}
}
}
// tools/platform/gc/devices.go:206:1
void gc_pe_setToken(gc_pe* d,gc_Machine* m,uint16_t tok,bool raise){
{
d->Token = tok;
if ((raise && d->TokenEnable)) {
gc_Machine_raiseInt(m,cast<int64_t>(9ULL));
}
}
}
// tools/platform/gc/devices.go:229:1
void gc_wgPipe_push(gc_wgPipe* w,gc_Machine* m,Slice<uint8_t> b){
{
w->Bytes += cast<uint64_t>(len(b));
gc_gpu_feed(&(m->gpu),m,b);
if ((!m->OnFIFO)) {
return ;
}
w->Buf = append(w->Buf,b);
{;for (;(len(w->Buf) >= cast<int64_t>(32ULL));){
m->OnFIFO(sub(w->Buf,0,cast<int64_t>(32ULL)));
w->Buf = sub(w->Buf,cast<int64_t>(32ULL),len(w->Buf));
}
}}
}
// tools/platform/gc/devices.go:244:1
void gc_wgPipe_write8(gc_wgPipe* w,gc_Machine* m,uint8_t v){
{
gc_wgPipe_push(w,m,Slice<uint8_t>{v});
}
}
// tools/platform/gc/devices.go:245:1
void gc_wgPipe_write16(gc_wgPipe* w,gc_Machine* m,uint16_t v){
{
gc_wgPipe_push(w,m,Slice<uint8_t>{cast<uint8_t>(shr<uint16_t>(v,cast<int64_t>(8ULL))),cast<uint8_t>(v)});
}
}
// tools/platform/gc/devices.go:246:1
void gc_wgPipe_write32(gc_wgPipe* w,gc_Machine* m,uint32_t v){
{
gc_wgPipe_push(w,m,Slice<uint8_t>{cast<uint8_t>(shr<uint32_t>(v,cast<int64_t>(24ULL))),cast<uint8_t>(shr<uint32_t>(v,cast<int64_t>(16ULL))),cast<uint8_t>(shr<uint32_t>(v,cast<int64_t>(8ULL))),cast<uint8_t>(v)});
}
}
// tools/platform/gc/di.go:60:1
void gc_di_init(gc_di* d){
{
d->Cover = cast<uint32_t>(0ULL);
d->SR = cast<uint32_t>(42ULL);
}
}
// tools/platform/gc/di.go:68:1
uint32_t gc_di_read(gc_di* d,gc_Machine* m,uint32_t off,int64_t size){
{
{
switch(cast<uint32_t>((off & cast<uint32_t>(255ULL)))){
case cast<uint32_t>(0ULL):{
return d->SR;
break;}
case cast<uint32_t>(4ULL):{
return d->Cover;
break;}
case cast<uint32_t>(8ULL):case cast<uint32_t>(12ULL):case cast<uint32_t>(16ULL):{
return d->Cmd[divi<uint32_t>((cast<uint32_t>((cast<uint32_t>((off & cast<uint32_t>(255ULL))) - cast<uint32_t>(8ULL)))),cast<uint32_t>(4ULL))];
break;}
case cast<uint32_t>(20ULL):{
return d->MAR;
break;}
case cast<uint32_t>(24ULL):{
return d->Length;
break;}
case cast<uint32_t>(28ULL):{
return d->CR;
break;}
case cast<uint32_t>(32ULL):{
return d->ImmBuf;
break;}
case cast<uint32_t>(36ULL):{
return d->Cfg;
break;}
}}
gc_Machine_logf(m,std::string("DI read unmodelled 0x%02X",25),cast<uint32_t>((off & cast<uint32_t>(255ULL))));
return cast<uint32_t>(0ULL);
}
}
// tools/platform/gc/di.go:91:1
void gc_di_write(gc_di* d,gc_Machine* m,uint32_t off,uint32_t v,int64_t size){
{
{
switch(cast<uint32_t>((off & cast<uint32_t>(255ULL)))){
case cast<uint32_t>(0ULL):{
d->SR &= ~(cast<uint32_t>((v & cast<uint32_t>(84ULL))));
d->SR = cast<uint32_t>((((d->SR & ~(cast<uint32_t>(42ULL)))) | (cast<uint32_t>((v & cast<uint32_t>(42ULL))))));
if ((((cast<uint32_t>((v & cast<uint32_t>(64ULL))) != cast<uint32_t>(0ULL)) || (cast<uint32_t>((v & cast<uint32_t>(16ULL))) != cast<uint32_t>(0ULL))) || (cast<uint32_t>((v & cast<uint32_t>(4ULL))) != cast<uint32_t>(0ULL)))) {
gc_Machine_diRefreshIRQ(m);
}
break;}
case cast<uint32_t>(4ULL):{
d->Cover &= ~(cast<uint32_t>((v & cast<uint32_t>(4ULL))));
break;}
case cast<uint32_t>(8ULL):case cast<uint32_t>(12ULL):case cast<uint32_t>(16ULL):{
d->Cmd[divi<uint32_t>((cast<uint32_t>((cast<uint32_t>((off & cast<uint32_t>(255ULL))) - cast<uint32_t>(8ULL)))),cast<uint32_t>(4ULL))] = v;
if ((gc_diTrace && ((cast<uint32_t>((off & cast<uint32_t>(255ULL)))) == cast<uint32_t>(16ULL)))) {
go_fmt_Fprintf(go_os_Stderr,std::string("DI Cmd2(offset word)=0x%08X -> byte 0x%X (pc 0x%08X)\012%s",55),v,shl<int64_t>(cast<int64_t>(v),cast<int64_t>(2ULL)),m->CPU->PC,gc_Machine_BacktraceString(m));
}
break;}
case cast<uint32_t>(20ULL):{
d->MAR = cast<uint32_t>((v & cast<uint32_t>(67108832ULL)));
break;}
case cast<uint32_t>(24ULL):{
d->Length = cast<uint32_t>((v & cast<uint32_t>(67108832ULL)));
break;}
case cast<uint32_t>(28ULL):{
d->CR = v;
if ((cast<uint32_t>((v & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL))) {
gc_di_exec(d,m);
}
break;}
case cast<uint32_t>(36ULL):{
d->Cfg = v;
break;}
default:{
gc_Machine_logf(m,std::string("DI write unmodelled 0x%02X = 0x%08X",35),cast<uint32_t>((off & cast<uint32_t>(255ULL))),v);
break;}
}}
}
}
// tools/platform/gc/di.go:127:1
void gc_di_exec(gc_di* d,gc_Machine* m){
{
uint32_t opcode = shr<uint32_t>(d->Cmd[cast<int64_t>(0ULL)],cast<int64_t>(24ULL));
{
switch(opcode){
case cast<uint32_t>(168ULL):{
int64_t discOff = shl<int64_t>(cast<int64_t>(d->Cmd[cast<int64_t>(1ULL)]),cast<int64_t>(2ULL));
uint32_t length = d->Length;
if (bool(m->OnDVDRead)) {
m->OnDVDRead(discOff,length,d->MAR);
}
if ((d->BusyInstr != cast<int64_t>(0ULL))) {
gc_Machine_logf(m,std::string("DI read issued while a transfer is in flight (offset 0x%X)",58),discOff);
}
auto tmp12 = std::make_tuple(discOff,length,d->MAR);
d->PendOff = std::get<0>(tmp12);
d->PendLen = std::get<1>(tmp12);
d->PendMAR = std::get<2>(tmp12);
d->BusyInstr = cast<int64_t>((cast<int64_t>(486000ULL) + cast<int64_t>((cast<int64_t>(length) * cast<int64_t>(194ULL)))));
if ((discOff != d->LastEnd)) {
d->BusyInstr += cast<int64_t>(14580000ULL);
}
if (gc_diInstant) {
d->BusyInstr = cast<int64_t>(1ULL);
}
break;}
case cast<uint32_t>(224ULL):{
d->ImmBuf = cast<uint32_t>(0ULL);
gc_di_complete(d,m);
break;}
case cast<uint32_t>(18ULL):{
gc_Machine_dmaToRAM(m,d->MAR,Slice<uint8_t>{cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),cast<uint8_t>(32ULL),cast<uint8_t>(1ULL),cast<uint8_t>(6ULL),cast<uint8_t>(3ULL),cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),cast<uint8_t>(0ULL)});
gc_di_complete(d,m);
break;}
default:{
gc_Machine_logf(m,std::string("DI unimplemented command 0x%02X (0x%08X 0x%08X 0x%08X)",54),opcode,d->Cmd[cast<int64_t>(0ULL)],d->Cmd[cast<int64_t>(1ULL)],d->Cmd[cast<int64_t>(2ULL)]);
gc_di_complete(d,m);
break;}
}}
}
}
// tools/platform/gc/di.go:187:1
void gc_Machine_tickDI(gc_Machine* m){
{
gc_di* d = (&m->di);
if ((d->BusyInstr == cast<int64_t>(0ULL))) {
return ;
}
d->BusyInstr--;
if ((d->BusyInstr != cast<int64_t>(0ULL))) {
return ;
}
d->LastEnd = cast<int64_t>((d->PendOff + cast<int64_t>(d->PendLen)));
if (bool(m->disc)) {
auto tmp13 = gc_Disc_Read(m->disc,d->PendOff,cast<int64_t>(d->PendLen));
Slice<uint8_t> data = std::get<0>(tmp13);
Error err = std::get<1>(tmp13);
if (bool(err)) {
gc_Machine_logf(m,std::string("DI read past the disc: offset 0x%X length %d: %v",48),d->PendOff,d->PendLen,err);
gc_di_raiseError(d,m);
return ;
}
gc_Machine_dmaToRAM(m,d->PendMAR,data);
}
gc_di_complete(d,m);
}
}
// tools/platform/gc/di.go:211:1
void gc_di_complete(gc_di* d,gc_Machine* m){
{
d->CR &= ~(cast<uint32_t>(1ULL));
d->Length = cast<uint32_t>(0ULL);
d->SR |= cast<uint32_t>(16ULL);
gc_Machine_diRefreshIRQ(m);
}
}
// tools/platform/gc/di.go:218:1
void gc_di_raiseError(gc_di* d,gc_Machine* m){
{
d->CR &= ~(cast<uint32_t>(1ULL));
d->SR |= cast<uint32_t>(4ULL);
gc_Machine_diRefreshIRQ(m);
}
}
// tools/platform/gc/di.go:225:1
void gc_Machine_diRefreshIRQ(gc_Machine* m){
{
uint32_t s = m->di.SR;
bool pending = (((((cast<uint32_t>((s & cast<uint32_t>(16ULL))) != cast<uint32_t>(0ULL)) && (cast<uint32_t>((s & cast<uint32_t>(8ULL))) != cast<uint32_t>(0ULL)))) || (((cast<uint32_t>((s & cast<uint32_t>(4ULL))) != cast<uint32_t>(0ULL)) && (cast<uint32_t>((s & cast<uint32_t>(2ULL))) != cast<uint32_t>(0ULL))))) || (((cast<uint32_t>((s & cast<uint32_t>(64ULL))) != cast<uint32_t>(0ULL)) && (cast<uint32_t>((s & cast<uint32_t>(32ULL))) != cast<uint32_t>(0ULL)))));
if (pending) {
gc_Machine_raiseInt(m,cast<int64_t>(2ULL));
}
else {
gc_Machine_clearInt(m,cast<int64_t>(2ULL));
}
}
}
// tools/platform/gc/disc.go:75:1
int64_t gc_Apploader_Body(gc_Apploader a){
{
return cast<int64_t>((cast<int64_t>(a.Size) + cast<int64_t>(a.TrailerSize)));
}
}
// tools/platform/gc/disc.go:171:1
std::tuple<Slice<uint8_t>,Error> gc_Disc_ApploaderCode(gc_Disc* d){
{
return gc_Disc_Read(d,cast<int64_t>(9312ULL),gc_Apploader_Body(d->Apploader));
}
}
// tools/platform/gc/disc.go:176:1
std::tuple<gc_DOL*,Error> gc_Disc_DOL(gc_Disc* d){
{
auto tmp14 = gc_Disc_Read(d,cast<int64_t>(d->Header.DOLOffset),cast<int64_t>(256ULL));
Slice<uint8_t> hdr = std::get<0>(tmp14);
Error err = std::get<1>(tmp14);
if (bool(err)) {
return {{},err};
}
int64_t n = gc_dolLength(hdr);
auto tmp15 = gc_Disc_Read(d,cast<int64_t>(d->Header.DOLOffset),n);
Slice<uint8_t> b = std::get<0>(tmp15);
err = std::get<1>(tmp15);
if (bool(err)) {
return {{},err};
}
return gc_ParseDOL(b);
}
}
// tools/platform/gc/disc.go:207:1
uint32_t gc_be32(Slice<uint8_t> b){
{
return be_Uint32(b);
}
}
// tools/platform/gc/disc.go:210:1
std::string gc_cstr(Slice<uint8_t> b){
{
{auto&& tmp16 = b;
for(int64_t tmp17=0;tmp17<len(tmp16);++tmp17){
auto i=tmp17;auto c=tmp16[tmp17];if ((c == cast<uint8_t>(0ULL))) {
return cast<std::string>(sub(b,0,i));
}
}}
return cast<std::string>(b);
}
}
// tools/platform/gc/dol.go:39:1
std::string gc_Segment_Name(gc_Segment s){
{
if (s.Text) {
return go_fmt_Sprintf(std::string("text%d",6),s.Index);
}
return go_fmt_Sprintf(std::string("data%d",6),cast<int64_t>((s.Index - cast<int64_t>(7ULL))));
}
}
// tools/platform/gc/dol.go:57:1
int64_t gc_dolLength(Slice<uint8_t> hdr){
{
uint32_t n = cast<uint32_t>(256ULL);
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(18ULL));i++){
uint32_t off = gc_be32(rrBorrow(hdr,cast<int64_t>((i * cast<int64_t>(4ULL))),len(hdr)));
uint32_t size = gc_be32(rrBorrow(hdr,cast<int64_t>((cast<int64_t>(144ULL) + cast<int64_t>((i * cast<int64_t>(4ULL))))),len(hdr)));
if ((size == cast<uint32_t>(0ULL))) {
continue;
}
{
uint32_t end = cast<uint32_t>((off + size));
if ((end > n)) {
n = end;
}
}
}
}return cast<int64_t>(n);
}
}
// tools/platform/gc/dol.go:73:1
std::tuple<gc_DOL*,Error> gc_ParseDOL(Slice<uint8_t> b){
{
if ((len(b) < cast<int64_t>(256ULL))) {
return {{},go_errors_New(std::string("gc: the executable is shorter than its own header",49))};
}
gc_DOL* d = arenaNew(gc_DOL{gc_be32(rrBorrow(b,cast<int64_t>(224ULL),len(b))),gc_be32(rrBorrow(b,cast<int64_t>(216ULL),len(b))),gc_be32(rrBorrow(b,cast<int64_t>(220ULL),len(b))),{}});
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(18ULL));i++){
uint32_t off = gc_be32(rrBorrow(b,cast<int64_t>((i * cast<int64_t>(4ULL))),len(b)));
uint32_t addr = gc_be32(rrBorrow(b,cast<int64_t>((cast<int64_t>(72ULL) + cast<int64_t>((i * cast<int64_t>(4ULL))))),len(b)));
uint32_t size = gc_be32(rrBorrow(b,cast<int64_t>((cast<int64_t>(144ULL) + cast<int64_t>((i * cast<int64_t>(4ULL))))),len(b)));
if ((size == cast<uint32_t>(0ULL))) {
continue;
}
if ((cast<int64_t>((cast<int64_t>(off) + cast<int64_t>(size))) > len(b))) {
return {{},go_fmt_Errorf(std::string("gc: segment %d spans %#x+%#x, past the %d-byte executable",57),i,off,size,len(b))};
}
d->Segments = append(d->Segments,Slice<gc_Segment>{gc_Segment{i,(i < cast<int64_t>(7ULL)),off,addr,size,sub(b,off,cast<uint32_t>((off + size)))}});
}
}if ((len(d->Segments) == cast<int64_t>(0ULL))) {
return {{},go_errors_New(std::string("gc: the executable has no segments",34))};
}
return {d,{}};
}
}
// tools/platform/gc/dol.go:110:1
void gc_DOL_Load(gc_DOL* d,std::function<void(uint32_t,Slice<uint8_t>)> write){
{
{auto&& tmp18 = d->Segments;
for(int64_t tmp19=0;tmp19<len(tmp18);++tmp19){
auto s=tmp18[tmp19];write(s.Addr,s.Data);
}}
if ((d->BSSSize > cast<uint32_t>(0ULL))) {
write(d->BSSAddr,Slice<uint8_t>::make(d->BSSSize));
}
}
}
// tools/platform/gc/dol.go:126:1
std::tuple<uint32_t,Slice<uint8_t>> gc_DOL_Flat(gc_DOL* d){
uint32_t base{};
Slice<uint8_t> mem{};
{
auto tmp20 = std::make_tuple(d->Segments[cast<int64_t>(0ULL)].Addr,cast<uint32_t>(0ULL));
uint32_t lo = std::get<0>(tmp20);
uint32_t hi = std::get<1>(tmp20);
{auto&& tmp21 = d->Segments;
for(int64_t tmp22=0;tmp22<len(tmp21);++tmp22){
auto s=tmp21[tmp22];if ((s.Addr < lo)) {
lo = s.Addr;
}
{
uint32_t end = cast<uint32_t>((s.Addr + s.Size));
if ((end > hi)) {
hi = end;
}
}
}}
mem = Slice<uint8_t>::make(cast<uint32_t>((hi - lo)));
{auto&& tmp23 = d->Segments;
for(int64_t tmp24=0;tmp24<len(tmp23);++tmp24){
auto s=tmp23[tmp24];gcopy(sub(mem,cast<uint32_t>((s.Addr - lo)),len(mem)),s.Data);
}}
return {lo,mem};
}
}
// tools/platform/gc/dol.go:145:1
bool gc_DOL_Text(gc_DOL* d,uint32_t addr){
{
{auto&& tmp25 = d->Segments;
for(int64_t tmp26=0;tmp26<len(tmp25);++tmp26){
auto s=tmp25[tmp26];if (((s.Text && (addr >= s.Addr)) && (addr < cast<uint32_t>((s.Addr + s.Size))))) {
return true;
}
}}
return false;
}
}
// tools/platform/gc/dsp.go:138:1
void gc_dsp_init(gc_dsp* d){
{
d->ARSize = cast<uint32_t>(16777216ULL);
d->FromDSP = cast<uint32_t>(2154954477ULL);
d->BootStep = cast<int64_t>(1ULL);
}
}
// tools/platform/gc/dsp.go:147:1
uint32_t gc_dsp_read(gc_dsp* d,gc_Machine* m,uint32_t off,int64_t size){
{
if (gc_dspTrace) {
auto tmp27=defer([&](){[&]()->void{
uint32_t r = cast<uint32_t>((off & cast<uint32_t>(4095ULL)));
if ((r <= cast<uint32_t>(10ULL))) {
go_fmt_Fprintf(go_os_Stderr,std::string("  DSP rd 0x%03X (pc 0x%08X)  toDSP=0x%08X fromDSP=0x%08X csr=0x%04X\012",68),r,m->CPU->PC,d->ToDSP,d->FromDSP,d->CSR);
}
}
();});
}
{
switch(cast<uint32_t>((off & cast<uint32_t>(4095ULL)))){
case cast<uint32_t>(0ULL):{
return shr<uint32_t>(d->ToDSP,cast<int64_t>(16ULL));
break;}
case cast<uint32_t>(2ULL):{
return cast<uint32_t>((d->ToDSP & cast<uint32_t>(65535ULL)));
break;}
case cast<uint32_t>(4ULL):{
return shr<uint32_t>(d->FromDSP,cast<int64_t>(16ULL));
break;}
case cast<uint32_t>(6ULL):{
uint32_t v = cast<uint32_t>((d->FromDSP & cast<uint32_t>(65535ULL)));
d->FromDSP &= ~(cast<uint32_t>(2147483648ULL));
gc_dsp_advanceBoot(d,m);
return v;
break;}
case cast<uint32_t>(10ULL):{
return d->CSR;
break;}
case cast<uint32_t>(18ULL):{
return d->ARSize;
break;}
case cast<uint32_t>(32ULL):{
return shr<uint32_t>(d->ARMMAddr,cast<int64_t>(16ULL));
break;}
case cast<uint32_t>(34ULL):{
return cast<uint32_t>((d->ARMMAddr & cast<uint32_t>(65535ULL)));
break;}
case cast<uint32_t>(36ULL):{
return shr<uint32_t>(d->ARARAddr,cast<int64_t>(16ULL));
break;}
case cast<uint32_t>(38ULL):{
return cast<uint32_t>((d->ARARAddr & cast<uint32_t>(65535ULL)));
break;}
case cast<uint32_t>(40ULL):{
return shr<uint32_t>(d->ARCtrl,cast<int64_t>(16ULL));
break;}
case cast<uint32_t>(42ULL):{
return cast<uint32_t>((d->ARCtrl & cast<uint32_t>(65535ULL)));
break;}
case cast<uint32_t>(22ULL):{
return cast<uint32_t>(1ULL);
break;}
case cast<uint32_t>(26ULL):{
return d->ARCtrl;
break;}
case cast<uint32_t>(48ULL):{
return shr<uint32_t>(d->AIDStart,cast<int64_t>(16ULL));
break;}
case cast<uint32_t>(50ULL):{
return cast<uint32_t>((d->AIDStart & cast<uint32_t>(65535ULL)));
break;}
case cast<uint32_t>(54ULL):{
return cast<uint32_t>(d->AIDControl);
break;}
case cast<uint32_t>(58ULL):{
return cast<uint32_t>(d->AIDRemaining);
break;}
}}
gc_Machine_logf(m,std::string("DSP read unmodelled 0x%03X",26),cast<uint32_t>((off & cast<uint32_t>(4095ULL))));
return cast<uint32_t>(0ULL);
}
}
// tools/platform/gc/dsp.go:207:1
void gc_dsp_write(gc_dsp* d,gc_Machine* m,uint32_t off,uint32_t v,int64_t size){
{
if ((gc_dspTrace && ((cast<uint32_t>((off & cast<uint32_t>(4095ULL)))) <= cast<uint32_t>(10ULL)))) {
go_fmt_Fprintf(go_os_Stderr,std::string("  DSP wr 0x%03X = 0x%08X (pc 0x%08X)\012",37),cast<uint32_t>((off & cast<uint32_t>(4095ULL))),v,m->CPU->PC);
}
{
switch(cast<uint32_t>((off & cast<uint32_t>(4095ULL)))){
case cast<uint32_t>(0ULL):{
d->ToDSP = cast<uint32_t>(((cast<uint32_t>((d->ToDSP & cast<uint32_t>(65535ULL)))) | (shl<uint32_t>(v,cast<int64_t>(16ULL)))));
break;}
case cast<uint32_t>(2ULL):{
d->ToDSP = cast<uint32_t>((cast<uint32_t>(((cast<uint32_t>((d->ToDSP & cast<uint32_t>(4294901760ULL)))) | (cast<uint32_t>((v & cast<uint32_t>(65535ULL)))))) | cast<uint32_t>(2147483648ULL)));
d->CoreBlocked = false;
gc_dsp_consumeMail(d,m);
break;}
case cast<uint32_t>(10ULL):{
uint32_t prevHalt = cast<uint32_t>((d->CSR & cast<uint32_t>(4ULL)));
uint32_t ack = cast<uint32_t>((v & cast<uint32_t>(170ULL)));
uint32_t keep = (v & ~(cast<uint32_t>(171ULL)));
d->CSR = cast<uint32_t>(((((d->CSR & ~(cast<uint32_t>(340ULL))) & ~(ack))) | keep));
if ((cast<uint32_t>((v & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL))) {
d->BootStep = cast<int64_t>(1ULL);
d->FromDSP = cast<uint32_t>(2154954477ULL);
d->UcodeRunning = false;
d->AwaitValue = false;
d->AwaitStartArg = false;
d->Core = {};
d->CoreBlocked = false;
}
if (((((!d->UcodeRunning) && (prevHalt != cast<uint32_t>(0ULL))) && (cast<uint32_t>((d->CSR & cast<uint32_t>(4ULL))) == cast<uint32_t>(0ULL))) && (cast<uint32_t>((d->FromDSP & cast<uint32_t>(2147483648ULL))) == cast<uint32_t>(0ULL)))) {
d->FromDSP = cast<uint32_t>(2154954477ULL);
}
gc_Machine_dspRefreshIRQ(m);
break;}
case cast<uint32_t>(18ULL):{
d->ARSize = v;
break;}
case cast<uint32_t>(32ULL):{
if ((size == cast<int64_t>(4ULL))) {
d->ARMMAddr = v;
}
else {
d->ARMMAddr = cast<uint32_t>(((cast<uint32_t>((d->ARMMAddr & cast<uint32_t>(65535ULL)))) | (shl<uint32_t>(v,cast<int64_t>(16ULL)))));
}
break;}
case cast<uint32_t>(34ULL):{
d->ARMMAddr = cast<uint32_t>(((cast<uint32_t>((d->ARMMAddr & cast<uint32_t>(4294901760ULL)))) | (cast<uint32_t>((v & cast<uint32_t>(65535ULL))))));
break;}
case cast<uint32_t>(36ULL):{
if ((size == cast<int64_t>(4ULL))) {
d->ARARAddr = v;
}
else {
d->ARARAddr = cast<uint32_t>(((cast<uint32_t>((d->ARARAddr & cast<uint32_t>(65535ULL)))) | (shl<uint32_t>(v,cast<int64_t>(16ULL)))));
}
break;}
case cast<uint32_t>(38ULL):{
d->ARARAddr = cast<uint32_t>(((cast<uint32_t>((d->ARARAddr & cast<uint32_t>(4294901760ULL)))) | (cast<uint32_t>((v & cast<uint32_t>(65535ULL))))));
break;}
case cast<uint32_t>(40ULL):{
if ((size == cast<int64_t>(4ULL))) {
d->ARCtrl = v;
}
else {
d->ARCtrl = cast<uint32_t>(((cast<uint32_t>((d->ARCtrl & cast<uint32_t>(65535ULL)))) | (shl<uint32_t>(v,cast<int64_t>(16ULL)))));
}
gc_dsp_runARAMDMA(d,m);
break;}
case cast<uint32_t>(42ULL):{
d->ARCtrl = cast<uint32_t>(((cast<uint32_t>((d->ARCtrl & cast<uint32_t>(4294901760ULL)))) | (cast<uint32_t>((v & cast<uint32_t>(65535ULL))))));
gc_dsp_runARAMDMA(d,m);
break;}
case cast<uint32_t>(26ULL):{
d->ARCtrl = v;
break;}
case cast<uint32_t>(48ULL):{
d->AIDStart = cast<uint32_t>(((cast<uint32_t>((d->AIDStart & cast<uint32_t>(65535ULL)))) | (shl<uint32_t>(v,cast<int64_t>(16ULL)))));
break;}
case cast<uint32_t>(50ULL):{
d->AIDStart = cast<uint32_t>(((cast<uint32_t>((d->AIDStart & cast<uint32_t>(4294901760ULL)))) | (cast<uint32_t>((v & cast<uint32_t>(65535ULL))))));
break;}
case cast<uint32_t>(54ULL):{
d->AIDControl = cast<uint16_t>(v);
d->AIDRemaining = cast<uint16_t>((d->AIDControl & cast<uint16_t>(32767ULL)));
d->AIDCur = d->AIDStart;
d->AIDAccum = cast<uint64_t>(0ULL);
break;}
default:{
gc_Machine_logf(m,std::string("DSP write unmodelled 0x%03X = 0x%08X",36),cast<uint32_t>((off & cast<uint32_t>(4095ULL))),v);
break;}
}}
}
}
// tools/platform/gc/dsp.go:316:1
void gc_dsp_consumeMail(gc_dsp* d,gc_Machine* m){
{
if (bool(d->Core)) {
return ;
}
uint32_t mail = (d->ToDSP & ~(cast<uint32_t>(2147483648ULL)));
d->ToDSP &= ~(cast<uint32_t>(2147483648ULL));
if (d->UcodeRunning) {
d->FromDSP = cast<uint32_t>(2147483648ULL);
return ;
}
if (d->AwaitValue) {
d->AwaitValue = false;
if (gc_dspTrace) {
go_fmt_Fprintf(go_os_Stderr,std::string("  DSP ucode-load param for cmd 0x%08X = 0x%08X\012",47),d->LoadCmd,mail);
}
{
switch(d->LoadCmd){
case cast<uint32_t>(15966209ULL):{
d->UcodeSrc = mail;
break;}
case cast<uint32_t>(15966210ULL):{
d->UcodeLen = mail;
break;}
case cast<uint32_t>(15974402ULL):{
d->UcodeDst = mail;
break;}
}}
if (d->AwaitStartArg) {
d->AwaitStartArg = false;
d->UcodeEntry = mail;
gc_dsp_startCore(d,m);
}
return ;
}
{
switch(mail){
case cast<uint32_t>(15966209ULL):case cast<uint32_t>(15966210ULL):case cast<uint32_t>(15974402ULL):case cast<uint32_t>(15970306ULL):{
d->AwaitValue = true;
d->LoadCmd = mail;
break;}
case cast<uint32_t>(15978497ULL):{
d->AwaitValue = true;
d->AwaitStartArg = true;
d->LoadCmd = mail;
break;}
default:{
gc_Machine_logf(m,std::string("DSP mail from CPU: 0x%08X (boot step %d) \342\200\224 acknowledged; the exact protocol is a work item",92),mail,d->BootStep);
break;}
}}
}
}
// tools/platform/gc/dsp.go:385:1
void gc_dsp_advanceBoot(gc_dsp* d,gc_Machine* m){
{
}
}
// tools/platform/gc/dsp.go:391:1
void gc_dsp_post(gc_dsp* d,gc_Machine* m,uint32_t v){
{
d->FromDSP = cast<uint32_t>((v | cast<uint32_t>(2147483648ULL)));
if ((cast<uint32_t>((d->CSR & cast<uint32_t>(256ULL))) != cast<uint32_t>(0ULL))) {
if (gc_dspTrace) {
go_fmt_Fprintf(go_os_Stderr,std::string("  DSP IRQ raise (post 0x%08X) csr=0x%04X\012",41),d->FromDSP,d->CSR);
}
d->CSR |= cast<uint32_t>(128ULL);
gc_Machine_dspRefreshIRQ(m);
}
}
}
// tools/platform/gc/dsp.go:404:1
void gc_dsp_runARAMDMA(gc_dsp* d,gc_Machine* m){
{
uint32_t length = cast<uint32_t>((d->ARCtrl & cast<uint32_t>(67108832ULL)));
if ((length == cast<uint32_t>(0ULL))) {
return ;
}
bool toARAM = (cast<uint32_t>((d->ARCtrl & cast<uint32_t>(2147483648ULL))) == cast<uint32_t>(0ULL));
uint32_t mm = cast<uint32_t>((d->ARMMAddr & cast<uint32_t>(67108832ULL)));
uint32_t ar = cast<uint32_t>((d->ARARAddr & cast<uint32_t>(67108863ULL)));
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < length);i++){
if (((cast<int64_t>(cast<uint32_t>((mm + i))) >= len(m->RAM)) || (cast<int64_t>(cast<uint32_t>((ar + i))) >= len(m->ARAM)))) {
break;
}
if (toARAM) {
m->ARAM[cast<uint32_t>((ar + i))] = m->RAM[cast<uint32_t>((mm + i))];
}
else {
m->RAM[cast<uint32_t>((mm + i))] = m->ARAM[cast<uint32_t>((ar + i))];
}
}
}d->ARCtrl = cast<uint32_t>(0ULL);
d->CSR |= cast<uint32_t>(32ULL);
if ((cast<uint32_t>((d->CSR & cast<uint32_t>(64ULL))) != cast<uint32_t>(0ULL))) {
gc_Machine_dspRefreshIRQ(m);
}
}
}
// tools/platform/gc/dsp.go:449:1
void gc_Machine_tickAID(gc_Machine* m){
{
gc_dsp* d = (&m->dsp);
if (((cast<uint16_t>((d->AIDControl & cast<uint16_t>(32768ULL))) == cast<uint16_t>(0ULL)) || (cast<uint16_t>((d->AIDControl & cast<uint16_t>(32767ULL))) == cast<uint16_t>(0ULL)))) {
return ;
}
d->AIDAccum++;
if ((d->AIDAccum < cast<uint64_t>(81000ULL))) {
return ;
}
d->AIDAccum = cast<uint64_t>(0ULL);
if ((d->AIDRemaining > cast<uint16_t>(0ULL))) {
d->AIDRemaining--;
if (bool(m->AIDTap)) {
uint32_t a = gc_phys(d->AIDCur);
if ((cast<int64_t>((cast<int64_t>(a) + cast<int64_t>(32ULL))) <= len(m->RAM))) {
m->AIDTap(sub(m->RAM,a,cast<uint32_t>((a + cast<uint32_t>(32ULL)))));
}
}
d->AIDCur += cast<uint32_t>(32ULL);
}
if ((d->AIDRemaining == cast<uint16_t>(0ULL))) {
d->AIDCur = d->AIDStart;
d->AIDRemaining = cast<uint16_t>((d->AIDControl & cast<uint16_t>(32767ULL)));
d->CSR |= cast<uint32_t>(8ULL);
gc_Machine_dspRefreshIRQ(m);
}
}
}
// tools/platform/gc/dsp.go:481:1
void gc_Machine_dspRefreshIRQ(gc_Machine* m){
{
uint32_t c = m->dsp.CSR;
bool pending = (((((cast<uint32_t>((c & cast<uint32_t>(128ULL))) != cast<uint32_t>(0ULL)) && (cast<uint32_t>((c & cast<uint32_t>(256ULL))) != cast<uint32_t>(0ULL)))) || (((cast<uint32_t>((c & cast<uint32_t>(32ULL))) != cast<uint32_t>(0ULL)) && (cast<uint32_t>((c & cast<uint32_t>(64ULL))) != cast<uint32_t>(0ULL))))) || (((cast<uint32_t>((c & cast<uint32_t>(8ULL))) != cast<uint32_t>(0ULL)) && (cast<uint32_t>((c & cast<uint32_t>(16ULL))) != cast<uint32_t>(0ULL)))));
if (pending) {
gc_Machine_raiseInt(m,cast<int64_t>(6ULL));
}
else {
gc_Machine_clearInt(m,cast<int64_t>(6ULL));
}
}
}
// tools/platform/gc/dsp.go:496:1
void gc_dsp_startCore(gc_dsp* d,gc_Machine* m){
{
uint32_t src = cast<uint32_t>((d->UcodeSrc & cast<uint32_t>(67108863ULL)));
gcdsp_CPU* core = gcdsp_New(gc_dspBus{m});
uint16_t dst = cast<uint16_t>(d->UcodeDst);
{uint32_t i = cast<uint32_t>(0ULL);for (;(cast<uint32_t>((i + cast<uint32_t>(1ULL))) < d->UcodeLen);i += cast<uint32_t>(2ULL)){
if ((cast<int64_t>((cast<int64_t>(cast<uint32_t>((src + i))) + cast<int64_t>(1ULL))) >= len(m->RAM))) {
break;
}
uint16_t w = cast<uint16_t>((shl<uint16_t>(cast<uint16_t>(m->RAM[cast<uint32_t>((src + i))]),cast<int64_t>(8ULL)) | cast<uint16_t>(m->RAM[cast<uint32_t>((cast<uint32_t>((src + i)) + cast<uint32_t>(1ULL)))])));
{
uint16_t wi = cast<uint16_t>((dst + cast<uint16_t>(divi<uint32_t>(i,cast<uint32_t>(2ULL)))));
if ((cast<int64_t>(wi) < cast<int64_t>(4096ULL))) {
core->IRAM[wi] = w;
}
}
}
}core->PC = cast<uint16_t>(d->UcodeEntry);
d->Core = core;
d->UcodeRunning = true;
gc_Machine_logf(m,std::string("DSP: real core started \342\200\224 %d-word ucode from 0x%08X, entry 0x%04X",66),divi<uint32_t>(d->UcodeLen,cast<uint32_t>(2ULL)),src,d->UcodeEntry);
}
}
// tools/platform/gc/dsp.go:521:1
uint16_t gc_dspBus_HWRead(gc_dspBus b,uint16_t a){
{
return gc_dsp_hwRead(&(b.m->dsp),b.m,a);
}
}
// tools/platform/gc/dsp.go:522:1
void gc_dspBus_HWWrite(gc_dspBus b,uint16_t a,uint16_t v){
{
gc_dsp_hwWrite(&(b.m->dsp),b.m,a,v);
}
}
// tools/platform/gc/dsp.go:527:1
uint16_t gc_dsp_hwRead(gc_dsp* d,gc_Machine* m,uint16_t a){
{
{
switch(a){
case cast<uint16_t>(65532ULL):{
return cast<uint16_t>(shr<uint32_t>(d->FromDSP,cast<int64_t>(16ULL)));
break;}
case cast<uint16_t>(65533ULL):{
return cast<uint16_t>(d->FromDSP);
break;}
case cast<uint16_t>(65534ULL):{
if ((cast<uint32_t>((d->ToDSP & cast<uint32_t>(2147483648ULL))) == cast<uint32_t>(0ULL))) {
d->corePolledEmpty = true;
}
return cast<uint16_t>(shr<uint32_t>(d->ToDSP,cast<int64_t>(16ULL)));
break;}
case cast<uint16_t>(65535ULL):{
uint16_t v = cast<uint16_t>(d->ToDSP);
d->ToDSP &= ~(cast<uint32_t>(2147483648ULL));
return v;
break;}
case cast<uint16_t>(65481ULL):{
return d->DSCtrl;
break;}
}}
if (((a >= cast<uint16_t>(65440ULL)) && (a <= cast<uint16_t>(65455ULL)))) {
return d->Coef[cast<uint16_t>((a - cast<uint16_t>(65440ULL)))];
}
{
switch(a){
case cast<uint16_t>(65489ULL):{
return d->AccFormat;
break;}
case cast<uint16_t>(65491ULL):{
return gc_dsp_accReadRaw(d,m);
break;}
case cast<uint16_t>(65492ULL):{
return cast<uint16_t>(shr<uint32_t>(d->AccStart,cast<int64_t>(16ULL)));
break;}
case cast<uint16_t>(65493ULL):{
return cast<uint16_t>(d->AccStart);
break;}
case cast<uint16_t>(65494ULL):{
return cast<uint16_t>(shr<uint32_t>(d->AccEnd,cast<int64_t>(16ULL)));
break;}
case cast<uint16_t>(65495ULL):{
return cast<uint16_t>(d->AccEnd);
break;}
case cast<uint16_t>(65496ULL):{
return cast<uint16_t>(shr<uint32_t>(d->AccCur,cast<int64_t>(16ULL)));
break;}
case cast<uint16_t>(65497ULL):{
return cast<uint16_t>(d->AccCur);
break;}
case cast<uint16_t>(65498ULL):{
return d->AccPred;
break;}
case cast<uint16_t>(65499ULL):{
return d->AccYn1;
break;}
case cast<uint16_t>(65500ULL):{
return d->AccYn2;
break;}
case cast<uint16_t>(65501ULL):{
return gc_dsp_accReadSample(d,m);
break;}
case cast<uint16_t>(65502ULL):{
return d->AccGain;
break;}
}}
gcdsp_CPU_Halt(d->Core,std::string("DSP read of unmodelled hardware register 0x%04X at ucode 0x%04X",63),a,d->Core->PC);
return cast<uint16_t>(0ULL);
}
}
// tools/platform/gc/dsp.go:590:1
void gc_dsp_hwWrite(gc_dsp* d,gc_Machine* m,uint16_t a,uint16_t v){
{
{
switch(a){
case cast<uint16_t>(65531ULL):{
if ((cast<uint16_t>((v & cast<uint16_t>(1ULL))) != cast<uint16_t>(0ULL))) {
if (gc_dspTrace) {
go_fmt_Fprintf(go_os_Stderr,std::string("  DSP IRQ raise (DIRQ, ucode pc 0x%04X) csr=0x%04X\012",51),d->Core->PC,d->CSR);
}
d->CSR |= cast<uint32_t>(128ULL);
gc_Machine_dspRefreshIRQ(m);
}
return ;
break;}
case cast<uint16_t>(65532ULL):{
d->FromDSP = cast<uint32_t>(((cast<uint32_t>((d->FromDSP & cast<uint32_t>(65535ULL)))) | (shl<uint32_t>(cast<uint32_t>(v),cast<int64_t>(16ULL)))));
return ;
break;}
case cast<uint16_t>(65533ULL):{
d->FromDSP = cast<uint32_t>((((d->FromDSP & ~(cast<uint32_t>(65535ULL)))) | cast<uint32_t>(v)));
if (gc_dspTrace) {
go_fmt_Fprintf(go_os_Stderr,std::string("  DSP mail out 0x%08X (ucode pc 0x%04X)\012",40),d->FromDSP,d->Core->PC);
}
return ;
break;}
case cast<uint16_t>(65481ULL):{
d->DSCtrl = v;
return ;
break;}
case cast<uint16_t>(65485ULL):{
d->DSPAddr = v;
return ;
break;}
case cast<uint16_t>(65486ULL):{
d->DSMAAddr = cast<uint32_t>(((cast<uint32_t>((d->DSMAAddr & cast<uint32_t>(65535ULL)))) | (shl<uint32_t>(cast<uint32_t>(v),cast<int64_t>(16ULL)))));
return ;
break;}
case cast<uint16_t>(65487ULL):{
d->DSMAAddr = cast<uint32_t>(((cast<uint32_t>((d->DSMAAddr & cast<uint32_t>(4294901760ULL)))) | cast<uint32_t>(v)));
return ;
break;}
case cast<uint16_t>(65483ULL):{
gc_dsp_runMemDMA(d,m,v);
return ;
break;}
}}
if (((a >= cast<uint16_t>(65440ULL)) && (a <= cast<uint16_t>(65455ULL)))) {
d->Coef[cast<uint16_t>((a - cast<uint16_t>(65440ULL)))] = v;
return ;
}
{
switch(a){
case cast<uint16_t>(65489ULL):{
d->AccFormat = v;
return ;
break;}
case cast<uint16_t>(65491ULL):{
gc_dsp_accWriteRaw(d,m,v);
return ;
break;}
case cast<uint16_t>(65492ULL):{
d->AccStart = cast<uint32_t>(((cast<uint32_t>((cast<uint32_t>((d->AccStart & cast<uint32_t>(65535ULL))) | shl<uint32_t>(cast<uint32_t>(v),cast<int64_t>(16ULL))))) & cast<uint32_t>(1073741823ULL)));
return ;
break;}
case cast<uint16_t>(65493ULL):{
d->AccStart = cast<uint32_t>(((cast<uint32_t>((cast<uint32_t>((d->AccStart & cast<uint32_t>(4294901760ULL))) | cast<uint32_t>(v)))) & cast<uint32_t>(1073741823ULL)));
return ;
break;}
case cast<uint16_t>(65494ULL):{
d->AccEnd = cast<uint32_t>(((cast<uint32_t>((cast<uint32_t>((d->AccEnd & cast<uint32_t>(65535ULL))) | shl<uint32_t>(cast<uint32_t>(v),cast<int64_t>(16ULL))))) & cast<uint32_t>(1073741823ULL)));
return ;
break;}
case cast<uint16_t>(65495ULL):{
d->AccEnd = cast<uint32_t>(((cast<uint32_t>((cast<uint32_t>((d->AccEnd & cast<uint32_t>(4294901760ULL))) | cast<uint32_t>(v)))) & cast<uint32_t>(1073741823ULL)));
return ;
break;}
case cast<uint16_t>(65496ULL):{
d->AccCur = cast<uint32_t>(((cast<uint32_t>((cast<uint32_t>((d->AccCur & cast<uint32_t>(65535ULL))) | shl<uint32_t>(cast<uint32_t>(v),cast<int64_t>(16ULL))))) & cast<uint32_t>(3221225471ULL)));
return ;
break;}
case cast<uint16_t>(65497ULL):{
d->AccCur = cast<uint32_t>(((cast<uint32_t>((cast<uint32_t>((d->AccCur & cast<uint32_t>(4294901760ULL))) | cast<uint32_t>(v)))) & cast<uint32_t>(3221225471ULL)));
return ;
break;}
case cast<uint16_t>(65498ULL):{
d->AccPred = cast<uint16_t>((v & cast<uint16_t>(127ULL)));
return ;
break;}
case cast<uint16_t>(65499ULL):{
d->AccYn1 = v;
return ;
break;}
case cast<uint16_t>(65500ULL):{
d->AccYn2 = v;
d->AccStopped = false;
return ;
break;}
case cast<uint16_t>(65502ULL):{
d->AccGain = v;
return ;
break;}
}}
gcdsp_CPU_Halt(d->Core,std::string("DSP write of unmodelled hardware register 0x%04X = 0x%04X at ucode 0x%04X",73),a,v,d->Core->PC);
}
}
// tools/platform/gc/dsp.go:682:1
uint8_t gc_dsp_aramByte(gc_dsp* d,gc_Machine* m,uint32_t addr){
{
if ((cast<int64_t>(addr) >= len(m->ARAM))) {
return cast<uint8_t>(0ULL);
}
return m->ARAM[addr];
}
}
// tools/platform/gc/dsp.go:693:1
uint16_t gc_dsp_accCurrentSample(gc_dsp* d,gc_Machine* m){
{
{
switch(cast<uint16_t>((d->AccFormat & cast<uint16_t>(3ULL)))){
case cast<uint16_t>(0ULL):{
uint8_t v = gc_dsp_aramByte(d,m,shr<uint32_t>(d->AccCur,cast<int64_t>(1ULL)));
if ((cast<uint32_t>((d->AccCur & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL))) {
return cast<uint16_t>(cast<uint8_t>((v & cast<uint8_t>(15ULL))));
}
return cast<uint16_t>(shr<uint8_t>(v,cast<int64_t>(4ULL)));
break;}
case cast<uint16_t>(1ULL):{
return cast<uint16_t>(gc_dsp_aramByte(d,m,d->AccCur));
break;}
case cast<uint16_t>(2ULL):{
return cast<uint16_t>((shl<uint16_t>(cast<uint16_t>(gc_dsp_aramByte(d,m,cast<uint32_t>((d->AccCur * cast<uint32_t>(2ULL))))),cast<int64_t>(8ULL)) | cast<uint16_t>(gc_dsp_aramByte(d,m,cast<uint32_t>((cast<uint32_t>((d->AccCur * cast<uint32_t>(2ULL))) + cast<uint32_t>(1ULL)))))));
break;}
default:{
gcdsp_CPU_Halt(d->Core,std::string("DSP accelerator: invalid sample width in format 0x%04X",54),d->AccFormat);
return cast<uint16_t>(0ULL);
break;}
}}
}
}
// tools/platform/gc/dsp.go:715:1
uint16_t gc_dsp_accReadRaw(gc_dsp* d,gc_Machine* m){
{
uint16_t v = gc_dsp_accCurrentSample(d,m);
if (d->Core->Halted) {
return cast<uint16_t>(0ULL);
}
d->AccCur++;
if ((cast<uint32_t>((d->AccCur - cast<uint32_t>(1ULL))) == d->AccEnd)) {
d->AccCur = d->AccStart;
}
d->AccCur &= cast<uint32_t>(3221225471ULL);
return v;
}
}
// tools/platform/gc/dsp.go:735:1
void gc_dsp_accWriteRaw(gc_dsp* d,gc_Machine* m,uint16_t v){
{
if ((cast<uint32_t>((d->AccCur & cast<uint32_t>(2147483648ULL))) == cast<uint32_t>(0ULL))) {
gc_Machine_logf(m,std::string("DSP accelerator: raw write without the address write flag (cur=0x%08X)",70),d->AccCur);
return ;
}
uint32_t byteAddr = cast<uint32_t>((d->AccCur * cast<uint32_t>(2ULL)));
if ((cast<int64_t>((cast<int64_t>(byteAddr) + cast<int64_t>(1ULL))) < len(m->ARAM))) {
m->ARAM[byteAddr] = cast<uint8_t>(shr<uint16_t>(v,cast<int64_t>(8ULL)));
m->ARAM[cast<uint32_t>((byteAddr + cast<uint32_t>(1ULL)))] = cast<uint8_t>(v);
}
d->AccCur++;
}
}
// tools/platform/gc/dsp.go:752:1
uint16_t gc_dsp_accReadSample(gc_dsp* d,gc_Machine* m){
{
if (d->AccStopped) {
return cast<uint16_t>(0ULL);
}
uint16_t decode = cast<uint16_t>(((shr<uint16_t>(d->AccFormat,cast<int64_t>(2ULL))) & cast<uint16_t>(3ULL)));
if (((decode == cast<uint16_t>(1ULL)) || (decode == cast<uint16_t>(3ULL)))) {
gcdsp_CPU_Halt(d->Core,std::string("DSP accelerator: MMIO-input PCM decode (format 0x%04X) not yet implemented",74),d->AccFormat);
return cast<uint16_t>(0ULL);
}
int32_t raw = cast<int32_t>(gc_dsp_accCurrentSample(d,m));
if (d->Core->Halted) {
return cast<uint16_t>(0ULL);
}
uint16_t coefIdx = cast<uint16_t>(((shr<uint16_t>(d->AccPred,cast<int64_t>(4ULL))) & cast<uint16_t>(7ULL)));
int32_t c1 = cast<int32_t>(cast<int16_t>(d->Coef[cast<uint16_t>((coefIdx * cast<uint16_t>(2ULL)))]));
int32_t c2 = cast<int32_t>(cast<int16_t>(d->Coef[cast<uint16_t>((cast<uint16_t>((coefIdx * cast<uint16_t>(2ULL))) + cast<uint16_t>(1ULL)))]));
int32_t yn1 = cast<int32_t>(cast<int16_t>(d->AccYn1));
int32_t yn2 = cast<int32_t>(cast<int16_t>(d->AccYn2));
uint16_t val={};
uint32_t step = cast<uint32_t>(2ULL);
if ((decode == cast<uint16_t>(0ULL))) {
raw &= cast<int32_t>(15ULL);
if ((raw >= cast<int32_t>(8ULL))) {
raw -= cast<int32_t>(16ULL);
}
int32_t scale = shl<int32_t>(cast<int32_t>(1ULL),(cast<uint16_t>((d->AccPred & cast<uint16_t>(15ULL)))));
int32_t v32 = cast<int32_t>((cast<int32_t>((scale * raw)) + (shr<int32_t>((cast<int32_t>((cast<int32_t>((cast<int32_t>(1024ULL) + cast<int32_t>((c1 * yn1)))) + cast<int32_t>((c2 * yn2))))),cast<int64_t>(11ULL)))));
if ((v32 > cast<int32_t>(32767ULL))) {
v32 = cast<int32_t>(32767ULL);
}
else if ((v32 < cast<int32_t>(-32767ULL))) {
v32 = cast<int32_t>(-32767ULL);
}
val = cast<uint16_t>(cast<int16_t>(v32));
auto tmp28 = std::make_tuple(d->AccYn1,val);
d->AccYn2 = std::get<0>(tmp28);
d->AccYn1 = std::get<1>(tmp28);
d->AccCur++;
{
if (((cast<uint32_t>((d->AccEnd & cast<uint32_t>(15ULL))) == cast<uint32_t>(0ULL)) && (d->AccCur == d->AccEnd))){
d->AccCur = cast<uint32_t>((d->AccStart + cast<uint32_t>(1ULL)));
}
else if (((cast<uint32_t>((d->AccEnd & cast<uint32_t>(15ULL))) == cast<uint32_t>(1ULL)) && (d->AccCur == cast<uint32_t>((d->AccEnd - cast<uint32_t>(1ULL)))))){
d->AccCur = d->AccStart;
}
else if ((cast<uint32_t>((d->AccCur & cast<uint32_t>(15ULL))) == cast<uint32_t>(0ULL))){
d->AccPred = cast<uint16_t>((cast<uint16_t>(gc_dsp_aramByte(d,m,shr<uint32_t>(((d->AccCur & ~(cast<uint32_t>(15ULL)))),cast<int64_t>(1ULL)))) & cast<uint16_t>(127ULL)));
d->AccCur += cast<uint32_t>(2ULL);
step += cast<uint32_t>(2ULL);
}
}
tmp29:;
}
else {
uint64_t gainShift={};
{
switch(cast<uint16_t>(((shr<uint16_t>(d->AccFormat,cast<int64_t>(4ULL))) & cast<uint16_t>(3ULL)))){
case cast<uint16_t>(0ULL):{
gainShift = cast<uint64_t>(11ULL);
break;}
case cast<uint16_t>(1ULL):{
gainShift = cast<uint64_t>(0ULL);
break;}
case cast<uint16_t>(2ULL):{
gainShift = cast<uint64_t>(16ULL);
break;}
default:{
gcdsp_CPU_Halt(d->Core,std::string("DSP accelerator: invalid gain scale in format 0x%04X",52),d->AccFormat);
return cast<uint16_t>(0ULL);
break;}
}}
int32_t v32 = cast<int32_t>((cast<int32_t>((shr<int32_t>((cast<int32_t>((cast<int32_t>(cast<int16_t>(d->AccGain)) * cast<int32_t>(cast<int16_t>(raw))))),gainShift) + shr<int32_t>((cast<int32_t>((c1 * yn1))),gainShift))) + shr<int32_t>((cast<int32_t>((c2 * yn2))),gainShift)));
val = cast<uint16_t>(cast<int16_t>(v32));
auto tmp30 = std::make_tuple(d->AccYn1,val);
d->AccYn2 = std::get<0>(tmp30);
d->AccYn1 = std::get<1>(tmp30);
d->AccCur++;
}
if ((d->AccCur == cast<uint32_t>((cast<uint32_t>((d->AccEnd + step)) - cast<uint32_t>(1ULL))))) {
d->AccCur = d->AccStart;
d->AccStopped = true;
}
d->AccCur &= cast<uint32_t>(3221225471ULL);
return val;
}
}
// tools/platform/gc/dsp.go:841:1
void gc_dsp_runMemDMA(gc_dsp* d,gc_Machine* m,uint16_t lenBytes){
{
if ((lenBytes == cast<uint16_t>(0ULL))) {
return ;
}
if ((cast<uint16_t>((d->DSCtrl & cast<uint16_t>(2ULL))) != cast<uint16_t>(0ULL))) {
gcdsp_CPU_Halt(d->Core,std::string("DSP memory DMA into instruction memory (DSCR=0x%04X) not modelled",65),d->DSCtrl);
return ;
}
bool toDSP = (cast<uint16_t>((d->DSCtrl & cast<uint16_t>(1ULL))) == cast<uint16_t>(0ULL));
uint32_t main = cast<uint32_t>((d->DSMAAddr & cast<uint32_t>(67108863ULL)));
uint32_t words = divi<uint32_t>(cast<uint32_t>(lenBytes),cast<uint32_t>(2ULL));
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < words);i++){
uint32_t mb = cast<uint32_t>((main + cast<uint32_t>((i * cast<uint32_t>(2ULL)))));
uint16_t dw = cast<uint16_t>((d->DSPAddr + cast<uint16_t>(i)));
if (((cast<int64_t>((cast<int64_t>(mb) + cast<int64_t>(1ULL))) >= len(m->RAM)) || (cast<int64_t>(dw) >= cast<int64_t>(4096ULL)))) {
break;
}
if (toDSP) {
d->Core->DRAM[dw] = cast<uint16_t>((shl<uint16_t>(cast<uint16_t>(m->RAM[mb]),cast<int64_t>(8ULL)) | cast<uint16_t>(m->RAM[cast<uint32_t>((mb + cast<uint32_t>(1ULL)))])));
}
else {
uint16_t w = d->Core->DRAM[dw];
m->RAM[mb] = cast<uint8_t>(shr<uint16_t>(w,cast<int64_t>(8ULL)));
m->RAM[cast<uint32_t>((mb + cast<uint32_t>(1ULL)))] = cast<uint8_t>(w);
}
}
}}
}
// tools/platform/gc/dsp.go:873:1
void gc_Machine_tickDSP(gc_Machine* m){
{
gc_dsp* d = (&m->dsp);
if (((((!d->Core) || d->CoreHalt) || d->CoreBlocked) || d->Core->Halted)) {
return ;
}
rrprof::Scope timing(5,"DSP core");
time_Time t = gc_Machine_profStart(m);
auto tmp31=defer([&](){gc_Machine_profEnd(m,cast<int64_t>(4ULL),t);});
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(64ULL));i++){
d->corePolledEmpty = false;
if (gc_dspPCTrace) {
go_fmt_Fprintf(go_os_Stderr,std::string("  ucode pc 0x%04X\012",18),d->Core->PC);
}
if ((!gcdsp_CPU_Step(d->Core))) {
gekko_CPU_Halt(m->CPU,std::string("DSP core halted: %s",19),d->Core->Reason);
return ;
}
if (d->corePolledEmpty) {
d->CoreBlocked = true;
return ;
}
}
}}
}
// tools/platform/gc/exi.go:57:1
void gc_exi_init(gc_exi* e){
{
Slice<uint8_t> s = sub(e->SRAM,0,len(e->SRAM));
s[cast<int64_t>(12ULL)] = cast<uint8_t>(0ULL);
gc_writeBE16(rrBorrow(s,cast<int64_t>(8ULL),len(s)),cast<uint16_t>(0ULL));
uint16_t sum={};
{int64_t i = cast<int64_t>(4ULL);for (;(i < cast<int64_t>(20ULL));i += cast<int64_t>(2ULL)){
sum += gc_beU16(sub(s,i,len(s)));
}
}gc_writeBE16(rrBorrow(s,cast<int64_t>(0ULL),len(s)),sum);
gc_writeBE16(rrBorrow(s,cast<int64_t>(2ULL),len(s)),cast<uint16_t>(~sum));
e->RTC = cast<uint32_t>(805306368ULL);
}
}
// tools/platform/gc/exi.go:87:1
uint32_t gc_exi_read(gc_exi* e,gc_Machine* m,uint32_t off,int64_t size){
{
uint32_t ch = divi<uint32_t>((cast<uint32_t>((off & cast<uint32_t>(255ULL)))),cast<uint32_t>(20ULL));
if ((ch >= cast<uint32_t>(3ULL))) {
gc_Machine_logf(m,std::string("EXI read unmodelled 0x%02X",26),cast<uint32_t>((off & cast<uint32_t>(255ULL))));
return cast<uint32_t>(0ULL);
}
gc_exiChannel* c = (&e->Ch[ch]);
{
switch(modi<uint32_t>((cast<uint32_t>((off & cast<uint32_t>(255ULL)))),cast<uint32_t>(20ULL))){
case cast<uint32_t>(0ULL):{
return c->CSR;
break;}
case cast<uint32_t>(4ULL):{
return c->DMAAddr;
break;}
case cast<uint32_t>(8ULL):{
return c->DMALen;
break;}
case cast<uint32_t>(12ULL):{
return c->CR;
break;}
case cast<uint32_t>(16ULL):{
return c->Data;
break;}
}}
gc_Machine_logf(m,std::string("EXI read unmodelled channel %d 0x%02X",37),ch,modi<uint32_t>((cast<uint32_t>((off & cast<uint32_t>(255ULL)))),cast<uint32_t>(20ULL)));
return cast<uint32_t>(0ULL);
}
}
// tools/platform/gc/exi.go:110:1
void gc_exi_write(gc_exi* e,gc_Machine* m,uint32_t off,uint32_t v,int64_t size){
{
uint32_t ch = divi<uint32_t>((cast<uint32_t>((off & cast<uint32_t>(255ULL)))),cast<uint32_t>(20ULL));
if ((ch >= cast<uint32_t>(3ULL))) {
gc_Machine_logf(m,std::string("EXI write unmodelled 0x%02X = 0x%08X",36),cast<uint32_t>((off & cast<uint32_t>(255ULL))),v);
return ;
}
if (gc_exiTrace) {
go_fmt_Fprintf(go_os_Stderr,std::string("  EXI wr ch%d 0x%02X = 0x%08X (pc 0x%08X)\012",42),ch,modi<uint32_t>((cast<uint32_t>((off & cast<uint32_t>(255ULL)))),cast<uint32_t>(20ULL)),v,m->CPU->PC);
}
gc_exiChannel* c = (&e->Ch[ch]);
{
switch(modi<uint32_t>((cast<uint32_t>((off & cast<uint32_t>(255ULL)))),cast<uint32_t>(20ULL))){
case cast<uint32_t>(0ULL):{
uint32_t ack = cast<uint32_t>((v & cast<uint32_t>(2058ULL)));
c->CSR = cast<uint32_t>((((v & ~(cast<uint32_t>(2058ULL)))) | ((cast<uint32_t>((c->CSR & cast<uint32_t>(2058ULL))) & ~(ack)))));
if ((cast<uint32_t>((v & cast<uint32_t>(896ULL))) == cast<uint32_t>(0ULL))) {
c->Dev = cast<int64_t>(-1ULL);
c->Phase = cast<int64_t>(0ULL);
}
else {
{
switch(cast<uint32_t>(((shr<uint32_t>(v,cast<int64_t>(7ULL))) & cast<uint32_t>(7ULL)))){
case cast<uint32_t>(1ULL):{
c->Dev = cast<int64_t>(0ULL);
break;}
case cast<uint32_t>(2ULL):{
c->Dev = cast<int64_t>(1ULL);
break;}
default:{
c->Dev = cast<int64_t>(-1ULL);
break;}
}}
c->Phase = cast<int64_t>(0ULL);
}
gc_exi_refreshIRQ(e,m);
break;}
case cast<uint32_t>(4ULL):{
c->DMAAddr = cast<uint32_t>((v & cast<uint32_t>(67108832ULL)));
break;}
case cast<uint32_t>(8ULL):{
c->DMALen = cast<uint32_t>((v & cast<uint32_t>(67108832ULL)));
break;}
case cast<uint32_t>(12ULL):{
c->CR = v;
if ((cast<uint32_t>((v & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL))) {
if ((cast<uint32_t>((v & cast<uint32_t>(2ULL))) != cast<uint32_t>(0ULL))) {
gc_exi_dmaTransfer(e,m,ch,(cast<uint32_t>(((shr<uint32_t>(v,cast<int64_t>(2ULL))) & cast<uint32_t>(3ULL))) == cast<uint32_t>(1ULL)));
}
else {
gc_exi_transfer(e,m,ch);
}
c->CR &= ~(cast<uint32_t>(1ULL));
c->CSR |= cast<uint32_t>(8ULL);
gc_exi_refreshIRQ(e,m);
}
break;}
case cast<uint32_t>(16ULL):{
c->Data = v;
break;}
default:{
gc_Machine_logf(m,std::string("EXI write unmodelled channel %d 0x%02X = 0x%08X",47),ch,modi<uint32_t>((cast<uint32_t>((off & cast<uint32_t>(255ULL)))),cast<uint32_t>(20ULL)),v);
break;}
}}
}
}
// tools/platform/gc/exi.go:172:1
void gc_exi_refreshIRQ(gc_exi* e,gc_Machine* m){
{
bool pending = false;
{auto&& tmp32 = e->Ch;
for(int64_t tmp33=0;tmp33<len(tmp32);++tmp33){
auto i=tmp33;uint32_t csr = e->Ch[i].CSR;
if ((cast<uint32_t>((cast<uint32_t>((csr & (shl<uint32_t>(csr,cast<int64_t>(1ULL))))) & cast<uint32_t>(2058ULL))) != cast<uint32_t>(0ULL))) {
pending = true;
}
}}
if (pending) {
gc_Machine_raiseInt(m,cast<int64_t>(4ULL));
}
else {
gc_Machine_clearInt(m,cast<int64_t>(4ULL));
}
}
}
// tools/platform/gc/exi.go:190:1
void gc_exi_transfer(gc_exi* e,gc_Machine* m,uint32_t ch){
{
gc_exiChannel* c = (&e->Ch[ch]);
{
switch(c->Dev){
case cast<int64_t>(1ULL):{
gc_exi_rtcTransfer(e,m,c);
break;}
case cast<int64_t>(0ULL):{
c->Data = cast<uint32_t>(4294967295ULL);
break;}
default:{
c->Data = cast<uint32_t>(0ULL);
break;}
}}
}
}
// tools/platform/gc/exi.go:207:1
void gc_exi_dmaTransfer(gc_exi* e,gc_Machine* m,uint32_t ch,bool asWrite){
{
gc_exiChannel* c = (&e->Ch[ch]);
auto tmp34 = std::make_tuple(c->DMAAddr,c->DMALen);
uint32_t addr = std::get<0>(tmp34);
uint32_t n = std::get<1>(tmp34);
if ((cast<int64_t>((cast<int64_t>(addr) + cast<int64_t>(n))) > len(m->RAM))) {
gc_Machine_logf(m,std::string("EXI DMA out of range 0x%08X+0x%X",32),addr,n);
return ;
}
{
switch(c->Dev){
case cast<int64_t>(1ULL):{
uint32_t dev = gc_rtcDevAddr(c->Command);
if ((dev < cast<uint32_t>(4ULL))) {
gc_Machine_logf(m,std::string("EXI DMA to RTC device address %d (command 0x%08X) unmodelled",60),dev,c->Command);
return ;
}
uint32_t off = cast<uint32_t>((dev - cast<uint32_t>(4ULL)));
{uint32_t i = cast<uint32_t>(0ULL);for (;((i < n) && (cast<uint32_t>((off + i)) < cast<uint32_t>(64ULL)));i++){
if (asWrite) {
e->SRAM[cast<uint32_t>((off + i))] = m->RAM[cast<uint32_t>((addr + i))];
}
else {
m->RAM[cast<uint32_t>((addr + i))] = e->SRAM[cast<uint32_t>((off + i))];
}
}
}break;}
case cast<int64_t>(0ULL):{
if ((!asWrite)) {
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < n);i++){
m->RAM[cast<uint32_t>((addr + i))] = cast<uint8_t>(255ULL);
}
}}
break;}
default:{
gc_Machine_logf(m,std::string("EXI DMA with no device selected (channel %d)",44),ch);
break;}
}}
}
}
// tools/platform/gc/exi.go:245:1
uint32_t gc_rtcDevAddr(uint32_t cmd){
{
return cast<uint32_t>(((shr<uint32_t>(cmd,cast<int64_t>(6ULL))) & cast<uint32_t>(32767ULL)));
}
}
// tools/platform/gc/exi.go:250:1
void gc_exi_rtcTransfer(gc_exi* e,gc_Machine* m,gc_exiChannel* c){
{
if ((c->Phase == cast<int64_t>(0ULL))) {
c->Command = c->Data;
c->Phase = cast<int64_t>(1ULL);
return ;
}
bool write = (cast<uint32_t>((c->Command & cast<uint32_t>(2147483648ULL))) != cast<uint32_t>(0ULL));
uint32_t dev = gc_rtcDevAddr(c->Command);
uint32_t word = cast<uint32_t>((cast<uint32_t>(cast<int64_t>((c->Phase - cast<int64_t>(1ULL)))) * cast<uint32_t>(4ULL)));
c->Phase++;
{
if ((dev == cast<uint32_t>(0ULL))){
if ((!write)) {
c->Data = e->RTC;
}
}
else if (((dev >= cast<uint32_t>(4ULL)) && (cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((dev - cast<uint32_t>(4ULL))) + word)) + cast<uint32_t>(3ULL))) < cast<uint32_t>(64ULL)))){
uint32_t idx = cast<uint32_t>((cast<uint32_t>((dev - cast<uint32_t>(4ULL))) + word));
if (write) {
gc_writeBE32(rrBorrow(e->SRAM,idx,len(e->SRAM)),c->Data);
}
else {
c->Data = gc_beU32(sub(e->SRAM,idx,len(e->SRAM)));
}
}
else {
gc_Machine_logf(m,std::string("EXI RTC-device access at device address %d (command 0x%08X) unmodelled",70),dev,c->Command);
if ((!write)) {
c->Data = cast<uint32_t>(0ULL);
}
}
}
tmp35:;
}
}
// tools/platform/gc/exi.go:283:1
uint16_t gc_beU16(Slice<uint8_t> b){
{
return cast<uint16_t>((shl<uint16_t>(cast<uint16_t>(b[cast<int64_t>(0ULL)]),cast<int64_t>(8ULL)) | cast<uint16_t>(b[cast<int64_t>(1ULL)])));
}
}
// tools/platform/gc/exi.go:284:1
uint32_t gc_beU32(Slice<uint8_t> b){
{
return cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((shl<uint32_t>(cast<uint32_t>(b[cast<int64_t>(0ULL)]),cast<int64_t>(24ULL)) | shl<uint32_t>(cast<uint32_t>(b[cast<int64_t>(1ULL)]),cast<int64_t>(16ULL)))) | shl<uint32_t>(cast<uint32_t>(b[cast<int64_t>(2ULL)]),cast<int64_t>(8ULL)))) | cast<uint32_t>(b[cast<int64_t>(3ULL)])));
}
}
// tools/platform/gc/exi.go:287:1
void gc_writeBE16(Slice<uint8_t> b,uint16_t v){
{
auto tmp36 = std::make_tuple(cast<uint8_t>(shr<uint16_t>(v,cast<int64_t>(8ULL))),cast<uint8_t>(v));
b[cast<int64_t>(0ULL)] = std::get<0>(tmp36);
b[cast<int64_t>(1ULL)] = std::get<1>(tmp36);
}
}
// tools/platform/gc/exi.go:290:1
void gc_writeBE32(Slice<uint8_t> b,uint32_t v){
{
auto tmp37 = std::make_tuple(cast<uint8_t>(shr<uint32_t>(v,cast<int64_t>(24ULL))),cast<uint8_t>(shr<uint32_t>(v,cast<int64_t>(16ULL))),cast<uint8_t>(shr<uint32_t>(v,cast<int64_t>(8ULL))),cast<uint8_t>(v));
b[cast<int64_t>(0ULL)] = std::get<0>(tmp37);
b[cast<int64_t>(1ULL)] = std::get<1>(tmp37);
b[cast<int64_t>(2ULL)] = std::get<2>(tmp37);
b[cast<int64_t>(3ULL)] = std::get<3>(tmp37);
}
}
// tools/platform/gc/fst.go:48:1
std::tuple<gc_FST*,Error> gc_ParseFST(Slice<uint8_t> b){
{
if ((len(b) < cast<int64_t>(12ULL))) {
return {{},go_errors_New(std::string("gc: the FST is too short to hold even its root",46))};
}
int64_t count = cast<int64_t>(gc_be32(rrBorrow(b,cast<int64_t>(8ULL),len(b))));
if ((count < cast<int64_t>(1ULL))) {
return {{},go_fmt_Errorf(std::string("gc: the FST claims %d entries",29),count)};
}
int64_t strTab = cast<int64_t>((count * cast<int64_t>(12ULL)));
if ((strTab > len(b))) {
return {{},go_fmt_Errorf(std::string("gc: the FST claims %d entries (%d bytes) but is only %d bytes",61),count,strTab,len(b))};
}
Slice<uint8_t> names = sub(b,strTab,len(b));
gc_FST* f = arenaNew(gc_FST{Slice<gc_File>::make(count),{}});
f->Entries[cast<int64_t>(0ULL)] = gc_File{std::string("/",1),std::string("/",1),true,{},cast<int64_t>(count)};
Slice<int64_t> dirEnd = Slice<int64_t>{count};
Slice<std::string> dirPath = Slice<std::string>{std::string("",0)};
{int64_t i = cast<int64_t>(1ULL);for (;(i < count);i++){
Slice<uint8_t> e = sub(b,cast<int64_t>((i * cast<int64_t>(12ULL))),len(b));
bool isDir = (e[cast<int64_t>(0ULL)] != cast<uint8_t>(0ULL));
int64_t nameOff = cast<int64_t>((cast<int64_t>((shl<int64_t>(cast<int64_t>(e[cast<int64_t>(1ULL)]),cast<int64_t>(16ULL)) | shl<int64_t>(cast<int64_t>(e[cast<int64_t>(2ULL)]),cast<int64_t>(8ULL)))) | cast<int64_t>(e[cast<int64_t>(3ULL)])));
uint32_t a = gc_be32(rrBorrow(e,cast<int64_t>(4ULL),len(e)));
uint32_t c = gc_be32(rrBorrow(e,cast<int64_t>(8ULL),len(e)));
if ((nameOff >= len(names))) {
return {{},go_fmt_Errorf(std::string("gc: entry %d names a string at %#x, past the %d-byte string table",65),i,nameOff,len(names))};
}
std::string name = gc_cstr(sub(names,nameOff,len(names)));
if ((name == std::string("",0))) {
return {{},go_fmt_Errorf(std::string("gc: entry %d has an empty name",30),i)};
}
{;for (;((len(dirEnd) > cast<int64_t>(1ULL)) && (i >= dirEnd[cast<int64_t>((len(dirEnd) - cast<int64_t>(1ULL)))]));){
dirEnd = sub(dirEnd,0,cast<int64_t>((len(dirEnd) - cast<int64_t>(1ULL))));
dirPath = sub(dirPath,0,cast<int64_t>((len(dirPath) - cast<int64_t>(1ULL))));
}
}std::string path = ((dirPath[cast<int64_t>((len(dirPath) - cast<int64_t>(1ULL)))] + std::string("/",1)) + name);
if (isDir) {
int64_t next = cast<int64_t>(c);
if (((next <= i) || (next > count))) {
return {{},go_fmt_Errorf(std::string("gc: directory %q at entry %d ends at %d, which is not after it and inside %d",76),path,i,next,count)};
}
f->Entries[i] = gc_File{path,name,true,{},cast<int64_t>(cast<int64_t>((cast<int64_t>((next - i)) - cast<int64_t>(1ULL))))};
dirEnd = append(dirEnd,Slice<int64_t>{next});
dirPath = append(dirPath,Slice<std::string>{path});
continue;
}
f->Entries[i] = gc_File{path,name,{},cast<int64_t>(a),cast<int64_t>(c)};
}
}{auto&& tmp38 = f->Entries;
for(int64_t tmp39=0;tmp39<len(tmp38);++tmp39){
auto e=tmp38[tmp39];if ((!e.Dir)) {
f->files = append(f->files,Slice<gc_File>{e});
}
}}
go_sort_Slice(f->files,[&](int64_t i,int64_t j)->bool{
return (f->files[i].Offset < f->files[j].Offset);
}
);
return {f,{}};
}
}
// tools/platform/gc/fst.go:117:1
Slice<gc_File> gc_FST_Files(gc_FST* f){
{
Slice<gc_File> out={};
{auto&& tmp40 = f->Entries;
for(int64_t tmp41=0;tmp41<len(tmp40);++tmp41){
auto e=tmp40[tmp41];if ((!e.Dir)) {
out = append(out,Slice<gc_File>{e});
}
}}
return out;
}
}
// tools/platform/gc/fst.go:130:1
std::tuple<gc_File,bool> gc_FST_ByPath(gc_FST* f,std::string path){
{
{auto&& tmp42 = f->Entries;
for(int64_t tmp43=0;tmp43<len(tmp42);++tmp43){
auto e=tmp42[tmp43];if (go_strings_EqualFold(e.Path,path)) {
return {e,true};
}
}}
return {gc_File{},false};
}
}
// tools/platform/gc/fst.go:142:1
std::tuple<gc_File,int64_t,bool> gc_FST_ByOffset(gc_FST* f,int64_t off){
gc_File file{};
int64_t within{};
bool ok{};
{
int64_t i = go_sort_Search(len(f->files),[&](int64_t i)->bool{
return (f->files[i].Offset > off);
}
);
if ((i == cast<int64_t>(0ULL))) {
return {gc_File{},cast<int64_t>(0ULL),false};
}
gc_File e = f->files[cast<int64_t>((i - cast<int64_t>(1ULL)))];
if ((off >= cast<int64_t>((e.Offset + e.Size)))) {
return {gc_File{},cast<int64_t>(0ULL),false};
}
return {e,cast<int64_t>((off - e.Offset)),true};
}
}
// tools/platform/gc/fst.go:158:1
Error gc_FST_Validate(gc_FST* f,int64_t discSize){
{
{auto&& tmp44 = f->files;
for(int64_t tmp45=0;tmp45<len(tmp44);++tmp45){
auto e=tmp44[tmp45];if ((((e.Offset < cast<int64_t>(0ULL)) || (e.Size < cast<int64_t>(0ULL))) || (cast<int64_t>((e.Offset + e.Size)) > discSize))) {
return go_fmt_Errorf(std::string("%s: extent %#x+%#x runs past the %d-byte image",46),e.Path,e.Offset,e.Size,discSize);
}
}}
{int64_t i = cast<int64_t>(1ULL);for (;(i < len(f->files));i++){
auto tmp46 = std::make_tuple(f->files[cast<int64_t>((i - cast<int64_t>(1ULL)))],f->files[i]);
gc_File prev = std::get<0>(tmp46);
gc_File cur = std::get<1>(tmp46);
if ((cast<int64_t>((prev.Offset + prev.Size)) > cur.Offset)) {
return go_fmt_Errorf(std::string("%s (%#x+%#x) overlaps %s (%#x)",30),prev.Path,prev.Offset,prev.Size,cur.Path,cur.Offset);
}
}
}return {};
}
}
// tools/platform/gc/gpu.go:78:1
void gc_gpu_xfStore(gc_gpu* g,int64_t addr,uint32_t val){
{
if (((addr >= cast<int64_t>(0ULL)) && (addr < cast<int64_t>(4192ULL)))) {
g->XFMem[addr] = val;
}
}
}
// tools/platform/gc/gpu.go:88:1
void gc_gpu_feed(gc_gpu* g,gc_Machine* m,Slice<uint8_t> b){
{rrprof::Scope timing(1,"GX command processor");
{
gc_Machine_profFIFOEnter(m);
g->Buf = append(g->Buf,b);
{;for (;gc_gpu_step(g,m);){
}
}gc_Machine_profFIFOExit(m);
}
}
}
// tools/platform/gc/gpu.go:103:1
bool gc_gpu_gxCmd(gc_gpu* g,gc_Machine* m,uint8_t op,Slice<uint32_t> words){
{
if (((m->gxStopAfter > cast<int64_t>(0ULL)) && (m->gxCmdCount >= m->gxStopAfter))) {
m->gxStopped = true;
m->StopRequested = true;
return false;
}
m->gxCmdCount++;
m->gxTotalCmds++;
if (bool(m->OnGXCmd)) {
m->OnGXCmd(m,op,words);
}
return true;
}
}
// tools/platform/gc/gpu.go:126:1
bool gc_gpu_step(gc_gpu* g,gc_Machine* m){
{
if (((m->CPU->Halted || m->gxStopped) || (len(g->Buf) == cast<int64_t>(0ULL)))) {
return false;
}
uint8_t op = g->Buf[cast<int64_t>(0ULL)];
{
if ((op == cast<uint8_t>(0ULL))){
if ((!gc_gpu_gxCmd(g,m,op,{}))) {
return false;
}
g->Census[cast<int64_t>(0ULL)]++;
g->Buf = sub(g->Buf,cast<int64_t>(1ULL),len(g->Buf));
}
else if ((op == cast<uint8_t>(8ULL))){
if ((len(g->Buf) < cast<int64_t>(6ULL))) {
return false;
}
if ((!gc_gpu_gxCmd(g,m,op,Slice<uint32_t>{cast<uint32_t>(g->Buf[cast<int64_t>(1ULL)]),gc_be32(rrBorrow(g->Buf,cast<int64_t>(2ULL),len(g->Buf)))}))) {
return false;
}
g->CPReg[g->Buf[cast<int64_t>(1ULL)]] = gc_be32(rrBorrow(g->Buf,cast<int64_t>(2ULL),len(g->Buf)));
g->Census[cast<int64_t>(8ULL)]++;
g->Buf = sub(g->Buf,cast<int64_t>(6ULL),len(g->Buf));
}
else if ((op == cast<uint8_t>(16ULL))){
if ((len(g->Buf) < cast<int64_t>(5ULL))) {
return false;
}
uint32_t cmd = gc_be32(rrBorrow(g->Buf,cast<int64_t>(1ULL),len(g->Buf)));
int64_t cnt = cast<int64_t>((cast<int64_t>(cast<uint32_t>(((shr<uint32_t>(cmd,cast<int64_t>(16ULL))) & cast<uint32_t>(15ULL)))) + cast<int64_t>(1ULL)));
int64_t total = cast<int64_t>((cast<int64_t>(5ULL) + cast<int64_t>((cast<int64_t>(4ULL) * cnt))));
if ((len(g->Buf) < total)) {
return false;
}
int64_t addr = cast<int64_t>(cast<uint32_t>((cmd & cast<uint32_t>(65535ULL))));
Slice<uint32_t> words = Slice<uint32_t>::make(cast<int64_t>(0ULL),cast<int64_t>((cnt + cast<int64_t>(1ULL))));
words = append(words,Slice<uint32_t>{cmd});
{int64_t k = cast<int64_t>(0ULL);for (;(k < cnt);k++){
words = append(words,Slice<uint32_t>{gc_be32(rrBorrow(g->Buf,cast<int64_t>((cast<int64_t>(5ULL) + cast<int64_t>((cast<int64_t>(4ULL) * k)))),len(g->Buf)))});
}
}if ((!gc_gpu_gxCmd(g,m,op,words))) {
return false;
}
{int64_t k = cast<int64_t>(0ULL);for (;(k < cnt);k++){
gc_gpu_xfStore(g,cast<int64_t>((addr + k)),gc_be32(rrBorrow(g->Buf,cast<int64_t>((cast<int64_t>(5ULL) + cast<int64_t>((cast<int64_t>(4ULL) * k)))),len(g->Buf))));
}
}g->Census[cast<int64_t>(16ULL)]++;
g->Buf = sub(g->Buf,total,len(g->Buf));
}
else if ((((op >= cast<uint8_t>(32ULL)) && (op <= cast<uint8_t>(56ULL))) && (cast<uint8_t>((op & cast<uint8_t>(7ULL))) == cast<uint8_t>(0ULL)))){
if ((len(g->Buf) < cast<int64_t>(5ULL))) {
return false;
}
if ((!gc_gpu_gxCmd(g,m,op,Slice<uint32_t>{gc_be32(rrBorrow(g->Buf,cast<int64_t>(1ULL),len(g->Buf)))}))) {
return false;
}
g->Census[op]++;
g->Buf = sub(g->Buf,cast<int64_t>(5ULL),len(g->Buf));
}
else if ((op == cast<uint8_t>(64ULL))){
if ((len(g->Buf) < cast<int64_t>(9ULL))) {
return false;
}
uint32_t addr = gc_be32(rrBorrow(g->Buf,cast<int64_t>(1ULL),len(g->Buf)));
uint32_t size = gc_be32(rrBorrow(g->Buf,cast<int64_t>(5ULL),len(g->Buf)));
if ((!gc_gpu_gxCmd(g,m,op,Slice<uint32_t>{addr,size}))) {
return false;
}
g->Census[cast<int64_t>(64ULL)]++;
g->Buf = sub(g->Buf,cast<int64_t>(9ULL),len(g->Buf));
gc_gpu_callDisplayList(g,m,addr,size);
}
else if ((op == cast<uint8_t>(72ULL))){
if ((!gc_gpu_gxCmd(g,m,op,{}))) {
return false;
}
g->Census[cast<int64_t>(72ULL)]++;
g->Buf = sub(g->Buf,cast<int64_t>(1ULL),len(g->Buf));
}
else if ((op == cast<uint8_t>(97ULL))){
if ((len(g->Buf) < cast<int64_t>(5ULL))) {
return false;
}
uint32_t v = gc_be32(rrBorrow(g->Buf,cast<int64_t>(1ULL),len(g->Buf)));
if ((!gc_gpu_gxCmd(g,m,op,Slice<uint32_t>{v}))) {
return false;
}
g->Census[cast<int64_t>(97ULL)]++;
g->Buf = sub(g->Buf,cast<int64_t>(5ULL),len(g->Buf));
gc_gpu_loadBP(g,m,cast<uint8_t>(shr<uint32_t>(v,cast<int64_t>(24ULL))),cast<uint32_t>((v & cast<uint32_t>(16777215ULL))));
}
else if ((op >= cast<uint8_t>(128ULL))){
if ((len(g->Buf) < cast<int64_t>(3ULL))) {
return false;
}
int64_t vat = cast<int64_t>(cast<uint8_t>((op & cast<uint8_t>(7ULL))));
int64_t vsize = gc_gpu_vertexSize(g,vat);
if ((vsize == cast<int64_t>(0ULL))) {
gekko_CPU_Halt(m->CPU,std::string("CP: draw 0x%02X has zero-size vertices \342\200\224 VCD/VAT not decoded (CP regs 0x50/0x60/0x70)",87),op);
return false;
}
int64_t count = cast<int64_t>(gc_be16(rrBorrow(g->Buf,cast<int64_t>(1ULL),len(g->Buf))));
int64_t total = cast<int64_t>((cast<int64_t>(3ULL) + cast<int64_t>((count * vsize))));
if ((len(g->Buf) < total)) {
return false;
}
if ((!gc_gpu_gxCmd(g,m,op,Slice<uint32_t>{cast<uint32_t>(vat),cast<uint32_t>(count)}))) {
return false;
}
g->Census[op]++;
gc_gpu_drawPrimitive(g,m,cast<uint32_t>(cast<uint8_t>((op & cast<uint8_t>(248ULL)))),vat,vsize,sub(g->Buf,cast<int64_t>(3ULL),total));
g->Buf = sub(g->Buf,total,len(g->Buf));
}
else {
gekko_CPU_Halt(m->CPU,std::string("CP: unknown FIFO opcode 0x%02X",30),op);
return false;
}
}
tmp47:;
return true;
}
}
// tools/platform/gc/gpu.go:250:1
std::string gc_GXName(uint8_t op){
{
{
if ((op == cast<uint8_t>(0ULL))){
return std::string("NOP",3);
}
else if ((op == cast<uint8_t>(8ULL))){
return std::string("CP_LOADREG",10);
}
else if ((op == cast<uint8_t>(16ULL))){
return std::string("XF_LOADREGS",11);
}
else if ((((op >= cast<uint8_t>(32ULL)) && (op <= cast<uint8_t>(56ULL))) && (cast<uint8_t>((op & cast<uint8_t>(7ULL))) == cast<uint8_t>(0ULL)))){
return std::string("XF_LOADINDEX",12);
}
else if ((op == cast<uint8_t>(64ULL))){
return std::string("CALL_DISPLAYLIST",16);
}
else if ((op == cast<uint8_t>(72ULL))){
return std::string("INVAL_VTXCACHE",14);
}
else if ((op == cast<uint8_t>(97ULL))){
return std::string("BP_LOADREG",10);
}
else if ((op >= cast<uint8_t>(128ULL))){
return (std::string("DRAW_",5) + gc_primName(cast<uint8_t>((op & cast<uint8_t>(248ULL)))));
}
}
tmp48:;
return std::string("UNKNOWN",7);
}
}
// tools/platform/gc/gpu.go:273:1
std::string gc_primName(uint8_t prim){
{
{
switch(prim){
case cast<uint8_t>(128ULL):{
return std::string("QUADS",5);
break;}
case cast<uint8_t>(136ULL):{
return std::string("QUADS2",6);
break;}
case cast<uint8_t>(144ULL):{
return std::string("TRIANGLES",9);
break;}
case cast<uint8_t>(152ULL):{
return std::string("TRIANGLESTRIP",13);
break;}
case cast<uint8_t>(160ULL):{
return std::string("TRIANGLEFAN",11);
break;}
case cast<uint8_t>(168ULL):{
return std::string("LINES",5);
break;}
case cast<uint8_t>(176ULL):{
return std::string("LINESTRIP",9);
break;}
case cast<uint8_t>(184ULL):{
return std::string("POINTS",6);
break;}
}}
return std::string("PRIM",4);
}
}
// tools/platform/gc/gpu.go:302:1
void gc_gpu_callDisplayList(gc_gpu* g,gc_Machine* m,uint32_t addr,uint32_t size){
{
if (g->inDisplayList) {
gekko_CPU_Halt(m->CPU,std::string("CP: display list at 0x%08X calls another display list",53),addr);
return ;
}
uint32_t a = gc_phys(addr);
if ((cast<int64_t>((cast<int64_t>(a) + cast<int64_t>(size))) > len(m->RAM))) {
gekko_CPU_Halt(m->CPU,std::string("CP: display list 0x%08X+0x%X runs past the end of RAM",53),addr,size);
return ;
}
Slice<uint8_t> saved = g->Buf;
g->Buf = append(Slice<uint8_t>{},sub(m->RAM,a,cast<uint32_t>((a + size))));
g->inDisplayList = true;
{;for (;gc_gpu_step(g,m);){
}
}int64_t leftover = len(g->Buf);
g->inDisplayList = false;
g->Buf = saved;
if ((((leftover != cast<int64_t>(0ULL)) && (!m->CPU->Halted)) && (!m->gxStopped))) {
gekko_CPU_Halt(m->CPU,std::string("CP: display list at 0x%08X ended mid-command (%d bytes left of %d)",66),addr,leftover,size);
}
}
}
// tools/platform/gc/gpu.go:334:1
void gc_gpu_loadBP(gc_gpu* g,gc_Machine* m,uint8_t reg,uint32_t data){
{
g->BP[reg] = data;
{
switch(reg){
case cast<uint8_t>(69ULL):{
if ((cast<uint32_t>((data & cast<uint32_t>(255ULL))) == cast<uint32_t>(2ULL))) {
gc_pe_setFinish(&(m->pe),m);
}
break;}
case cast<uint8_t>(71ULL):{
gc_pe_setToken(&(m->pe),m,cast<uint16_t>(data),false);
break;}
case cast<uint8_t>(72ULL):{
gc_pe_setToken(&(m->pe),m,cast<uint16_t>(data),true);
break;}
case cast<uint8_t>(82ULL):{
gc_gpu_copyDisplay(g,m,data);
break;}
case cast<uint8_t>(101ULL):{
gc_gpu_loadTlut(g,m,data);
break;}
}}
if (((reg >= cast<uint8_t>(224ULL)) && (reg <= cast<uint8_t>(231ULL)))) {
uint8_t pair = divi<uint8_t>((cast<uint8_t>((reg - cast<uint8_t>(224ULL)))),cast<uint8_t>(2ULL));
uint8_t word = cast<uint8_t>(((cast<uint8_t>((reg - cast<uint8_t>(224ULL)))) & cast<uint8_t>(1ULL)));
if ((cast<uint32_t>((data & cast<uint32_t>(8388608ULL))) != cast<uint32_t>(0ULL))) {
g->TevKonstReg[pair][word] = data;
}
else {
g->TevColorReg[pair][word] = data;
}
}
}
}
// tools/platform/gc/gpu.go:371:1
uint16_t gc_be16(Slice<uint8_t> b){
{
return cast<uint16_t>((shl<uint16_t>(cast<uint16_t>(b[cast<int64_t>(0ULL)]),cast<int64_t>(8ULL)) | cast<uint16_t>(b[cast<int64_t>(1ULL)])));
}
}
// tools/platform/gc/gpu_clip.go:48:1
float gc_clipPlane_dist(gc_clipPlane p,gc_clipVertex v){
{
return ((((p.a * v.cx) + (p.b * v.cy)) + (p.c * v.cz)) + (p.d * v.cw));
}
}
// tools/platform/gc/gpu_clip.go:71:1
uint8_t gc_outcode(gc_clipVertex v){
{
uint8_t c={};
{auto&& tmp49 = gc_clipPlanes;
for(int64_t tmp50=0;tmp50<len(tmp49);++tmp50){
auto i=tmp50;auto p=tmp49[tmp50];if ((gc_clipPlane_dist(p,v) < cast<float>(0ULL))) {
c |= shl<uint8_t>(cast<uint8_t>(1ULL),cast<uint64_t>(i));
}
}}
return c;
}
}
// tools/platform/gc/gpu_clip.go:87:1
Slice<gc_clipVertex> gc_clipByPlane(Slice<gc_clipVertex> out,Slice<gc_clipVertex> in,gc_clipPlane p){
{
out = sub(out,0,cast<int64_t>(0ULL));
if ((len(in) == cast<int64_t>(0ULL))) {
return out;
}
gc_clipVertex prev = in[cast<int64_t>((len(in) - cast<int64_t>(1ULL)))];
float dp = gc_clipPlane_dist(p,prev);
{auto&& tmp51 = in;
for(int64_t tmp52=0;tmp52<len(tmp51);++tmp52){
auto cur=tmp51[tmp52];float dc = gc_clipPlane_dist(p,cur);
if ((((dp >= cast<float>(0ULL))) != ((dc >= cast<float>(0ULL))))) {
out = append(out,Slice<gc_clipVertex>{gc_lerpClip(prev,cur,(dp / ((dp - dc))))});
}
if ((dc >= cast<float>(0ULL))) {
out = append(out,Slice<gc_clipVertex>{cur});
}
auto tmp53 = std::make_tuple(cur,dc);
prev = std::get<0>(tmp53);
dp = std::get<1>(tmp53);
}}
return out;
}
}
// tools/platform/gc/gpu_clip.go:114:1
std::tuple<Slice<gc_clipVertex>,Slice<gc_clipVertex>> gc_clipTriangle(Slice<gc_clipVertex> dst,Slice<gc_clipVertex> scratch,std::array<gc_clipVertex,3> tri){
{
auto tmp54 = std::make_tuple(gc_outcode(tri[cast<int64_t>(0ULL)]),gc_outcode(tri[cast<int64_t>(1ULL)]),gc_outcode(tri[cast<int64_t>(2ULL)]));
uint8_t c0 = std::get<0>(tmp54);
uint8_t c1 = std::get<1>(tmp54);
uint8_t c2 = std::get<2>(tmp54);
if ((cast<uint8_t>((cast<uint8_t>((c0 | c1)) | c2)) == cast<uint8_t>(0ULL))) {
return {append(sub(dst,0,cast<int64_t>(0ULL)),Slice<gc_clipVertex>{tri[cast<int64_t>(0ULL)],tri[cast<int64_t>(1ULL)],tri[cast<int64_t>(2ULL)]}),scratch};
}
if ((cast<uint8_t>((cast<uint8_t>((c0 & c1)) & c2)) != cast<uint8_t>(0ULL))) {
return {sub(dst,0,cast<int64_t>(0ULL)),scratch};
}
Slice<gc_clipVertex> poly = append(sub(dst,0,cast<int64_t>(0ULL)),Slice<gc_clipVertex>{tri[cast<int64_t>(0ULL)],tri[cast<int64_t>(1ULL)],tri[cast<int64_t>(2ULL)]});
Slice<gc_clipVertex> other = scratch;
{auto&& tmp55 = gc_clipPlanes;
for(int64_t tmp56=0;tmp56<len(tmp55);++tmp56){
auto i=tmp56;auto p=tmp55[tmp56];if ((cast<uint8_t>(((cast<uint8_t>((cast<uint8_t>((c0 | c1)) | c2))) & (shl<uint8_t>(cast<uint8_t>(1ULL),cast<uint64_t>(i))))) == cast<uint8_t>(0ULL))) {
continue;
}
other = gc_clipByPlane(other,poly,p);
auto tmp57 = std::make_tuple(other,poly);
poly = std::get<0>(tmp57);
other = std::get<1>(tmp57);
if ((len(poly) < cast<int64_t>(3ULL))) {
return {sub(poly,0,cast<int64_t>(0ULL)),other};
}
}}
return {poly,other};
}
}
// tools/platform/gc/gpu_clip.go:146:1
std::tuple<Slice<gc_clipVertex>,Slice<gc_clipVertex>> gc_gpu_clipAndSetup(gc_gpu* g,gc_Machine* m,Slice<gc_clipVertex> dst,Slice<gc_clipVertex> scratch,gc_clipVertex v0,gc_clipVertex v1,gc_clipVertex v2){
{
auto tmp58 = gc_clipTriangle(dst,scratch,std::array<gc_clipVertex,3>{v0,v1,v2});
Slice<gc_clipVertex> poly = std::get<0>(tmp58);
scratch = std::get<1>(tmp58);
if ((len(poly) < cast<int64_t>(3ULL))) {
return {poly,scratch};
}
auto toScreen = [&](gc_clipVertex c)->gc_screenVertex{
auto tmp59 = gc_gpu_toScreen(g,c.cx,c.cy,c.cz,c.cw);
float sx = std::get<0>(tmp59);
float sy = std::get<1>(tmp59);
float sz = std::get<2>(tmp59);
return gc_screenVertex{cast<float>(sx),cast<float>(sy),cast<float>(sz),c.col,c.tc,c.ntc,cast<float>((cast<float>(1ULL) / c.cw))};
}
;
gc_screenVertex a = toScreen(poly[cast<int64_t>(0ULL)]);
{int64_t i = cast<int64_t>(1ULL);for (;(cast<int64_t>((i + cast<int64_t>(1ULL))) < len(poly));i++){
{
auto tmp60 = gc_gpu_setupTri(g,a,toScreen(poly[i]),toScreen(poly[cast<int64_t>((i + cast<int64_t>(1ULL)))]));
gc_rasterTri t = std::get<0>(tmp60);
bool ok = std::get<1>(tmp60);
if (ok) {
g->tris = append(g->tris,Slice<gc_rasterTri>{t});
}
}
}
}return {poly,scratch};
}
}
// tools/platform/gc/gpu_efb.go:33:1
void gc_gpu_ensureEFB(gc_gpu* g){
{
if ((!g->EFB)) {
g->EFB = Slice<uint32_t>::make(cast<int64_t>(337920ULL));
gc_gpu_clearEFB(g);
}
}
}
// tools/platform/gc/gpu_efb.go:43:1
std::tuple<uint8_t,uint8_t,uint8_t,uint8_t> gc_gpu_clearColor(gc_gpu* g){
uint8_t r{};
uint8_t gg{};
uint8_t b{};
uint8_t a{};
{
uint32_t ar = g->BP[cast<int64_t>(79ULL)];
uint32_t gb = g->BP[cast<int64_t>(80ULL)];
return {cast<uint8_t>(cast<uint32_t>((ar & cast<uint32_t>(255ULL)))),cast<uint8_t>(shr<uint32_t>(gb,cast<int64_t>(8ULL))),cast<uint8_t>(cast<uint32_t>((gb & cast<uint32_t>(255ULL)))),cast<uint8_t>(shr<uint32_t>(ar,cast<int64_t>(8ULL)))};
}
}
// tools/platform/gc/gpu_efb.go:54:1
void gc_gpu_clearEFB(gc_gpu* g){
{auto restore=rrEFBClear(g);
{
auto tmp61 = gc_gpu_clearColor(g);
uint8_t r = std::get<0>(tmp61);
uint8_t gg = std::get<1>(tmp61);
uint8_t b = std::get<2>(tmp61);
uint8_t a = std::get<3>(tmp61);
uint32_t px = gc_packRGBA(r,gg,b,a);
{auto&& tmp62 = g->EFB;
for(int64_t tmp63=0;tmp63<len(tmp62);++tmp63){
auto i=tmp63;g->EFB[i] = px;
if(rrcapture::trace.active)rrEFBWrite(g,i);
}}
if (bool(g->ZBuf)) {
uint32_t z = cast<uint32_t>((g->BP[cast<int64_t>(81ULL)] & cast<uint32_t>(16777215ULL)));
{auto&& tmp64 = g->ZBuf;
for(int64_t tmp65=0;tmp65<len(tmp64);++tmp65){
auto i=tmp65;g->ZBuf[i] = z;
}}
}
}
}
}
// tools/platform/gc/gpu_efb.go:70:1
uint32_t gc_gpu_efbAt(gc_gpu* g,int64_t x,int64_t y){
{
if ((x < cast<int64_t>(0ULL))) {
x = cast<int64_t>(0ULL);
}
if ((y < cast<int64_t>(0ULL))) {
y = cast<int64_t>(0ULL);
}
if ((x >= cast<int64_t>(640ULL))) {
x = cast<int64_t>(639ULL);
}
if ((y >= cast<int64_t>(528ULL))) {
y = cast<int64_t>(527ULL);
}
return g->EFB[cast<int64_t>((cast<int64_t>((y * cast<int64_t>(640ULL))) + x))];
}
}
// tools/platform/gc/gpu_efb.go:88:1
uint32_t gc_gpu_zAt(gc_gpu* g,int64_t x,int64_t y){
{
if ((!g->ZBuf)) {
return cast<uint32_t>((g->BP[cast<int64_t>(81ULL)] & cast<uint32_t>(16777215ULL)));
}
if ((x < cast<int64_t>(0ULL))) {
x = cast<int64_t>(0ULL);
}
if ((y < cast<int64_t>(0ULL))) {
y = cast<int64_t>(0ULL);
}
if ((x >= cast<int64_t>(640ULL))) {
x = cast<int64_t>(639ULL);
}
if ((y >= cast<int64_t>(528ULL))) {
y = cast<int64_t>(527ULL);
}
return g->ZBuf[cast<int64_t>((cast<int64_t>((y * cast<int64_t>(640ULL))) + x))];
}
}
// tools/platform/gc/gpu_efb.go:111:1
void gc_gpu_copyDisplay(gc_gpu* g,gc_Machine* m,uint32_t params){
{rrprof::Scope timing(4,"Pixel engine copies");
{
gc_gpu_ensureEFB(g);
bool toXFB = (cast<uint32_t>((params & cast<uint32_t>(16384ULL))) != cast<uint32_t>(0ULL));
bool clear = (cast<uint32_t>((params & cast<uint32_t>(2048ULL))) != cast<uint32_t>(0ULL));
if ((!toXFB)) {
time_Time t = gc_Machine_profStart(m);
gc_gpu_copyTexture(g,m,params);
if ((clear && (!m->CPU->Halted))) {
gc_gpu_clearEFB(g);
}
gc_Machine_profEnd(m,cast<int64_t>(3ULL),t);
return ;
}
uint32_t tl = g->BP[cast<int64_t>(73ULL)];
uint32_t wh = g->BP[cast<int64_t>(74ULL)];
int64_t sx = cast<int64_t>(cast<uint32_t>((tl & cast<uint32_t>(1023ULL))));
int64_t sy = cast<int64_t>(cast<uint32_t>(((shr<uint32_t>(tl,cast<int64_t>(10ULL))) & cast<uint32_t>(1023ULL))));
int64_t w = cast<int64_t>((cast<int64_t>(cast<uint32_t>((wh & cast<uint32_t>(1023ULL)))) + cast<int64_t>(1ULL)));
int64_t h = cast<int64_t>((cast<int64_t>(cast<uint32_t>(((shr<uint32_t>(wh,cast<int64_t>(10ULL))) & cast<uint32_t>(1023ULL)))) + cast<int64_t>(1ULL)));
uint32_t dst = gc_phys(shl<uint32_t>((cast<uint32_t>((g->BP[cast<int64_t>(75ULL)] & cast<uint32_t>(16777215ULL)))),cast<int64_t>(5ULL)));
uint32_t stride = cast<uint32_t>(cast<int64_t>((w * cast<int64_t>(2ULL))));
auto tmp66 = gc_gpu_clearColor(g);
uint8_t r = std::get<0>(tmp66);
uint8_t gg = std::get<1>(tmp66);
uint8_t b = std::get<2>(tmp66);
uint8_t a = std::get<3>(tmp66);
gc_Machine_logf(m,std::string("PE copy-to-XFB: EFB (%d,%d) %dx%d -> 0x%08X clear=%v (clear colour R%d G%d B%d A%d)",83),sx,sy,w,h,dst,clear,r,gg,b,a);
if (gc_drawTrace) {
go_fmt_Fprintf(go_os_Stderr,std::string("COPY-XFB (%d,%d) %dx%d -> 0x%08X clear=%v\012",42),sx,sy,w,h,dst,clear);
}
time_Time tConv = gc_Machine_profStart(m);
{int64_t y = cast<int64_t>(0ULL);for (;(y < h);y++){
uint32_t base = cast<uint32_t>((dst + cast<uint32_t>((cast<uint32_t>(y) * stride))));
{int64_t x = cast<int64_t>(0ULL);for (;(x < w);x += cast<int64_t>(2ULL)){
auto tmp67 = gc_unpackRGB(gc_gpu_efbAt(g,cast<int64_t>((sx + x)),cast<int64_t>((sy + y))));
uint8_t r0 = std::get<0>(tmp67);
uint8_t g0 = std::get<1>(tmp67);
uint8_t b0 = std::get<2>(tmp67);
auto tmp68 = gc_unpackRGB(gc_gpu_efbAt(g,cast<int64_t>((cast<int64_t>((sx + x)) + cast<int64_t>(1ULL))),cast<int64_t>((sy + y))));
uint8_t r1 = std::get<0>(tmp68);
uint8_t g1 = std::get<1>(tmp68);
uint8_t b1 = std::get<2>(tmp68);
uint8_t y0 = gc_luma(r0,g0,b0);
uint8_t y1 = gc_luma(r1,g1,b1);
uint8_t ar = cast<uint8_t>(divi<int64_t>((cast<int64_t>((cast<int64_t>(r0) + cast<int64_t>(r1)))),cast<int64_t>(2ULL)));
uint8_t ag = cast<uint8_t>(divi<int64_t>((cast<int64_t>((cast<int64_t>(g0) + cast<int64_t>(g1)))),cast<int64_t>(2ULL)));
uint8_t ab = cast<uint8_t>(divi<int64_t>((cast<int64_t>((cast<int64_t>(b0) + cast<int64_t>(b1)))),cast<int64_t>(2ULL)));
uint8_t cb = gc_chromaB(ar,ag,ab);
uint8_t cr = gc_chromaR(ar,ag,ab);
uint32_t o = cast<uint32_t>((base + cast<uint32_t>((cast<uint32_t>(x) * cast<uint32_t>(2ULL)))));
if ((cast<int64_t>((cast<int64_t>(o) + cast<int64_t>(3ULL))) >= len(m->RAM))) {
break;
}
auto tmp69 = std::make_tuple(y0,cb,y1,cr);
m->RAM[o] = std::get<0>(tmp69);
m->RAM[cast<uint32_t>((o + cast<uint32_t>(1ULL)))] = std::get<1>(tmp69);
m->RAM[cast<uint32_t>((o + cast<uint32_t>(2ULL)))] = std::get<2>(tmp69);
m->RAM[cast<uint32_t>((o + cast<uint32_t>(3ULL)))] = std::get<3>(tmp69);
if(rrcapture::trace.active)rrGXWrite(m,o);
if(rrcapture::trace.active)rrGXWrite(m,cast<uint32_t>((o + cast<uint32_t>(1ULL))));
if(rrcapture::trace.active)rrGXWrite(m,cast<uint32_t>((o + cast<uint32_t>(2ULL))));
if(rrcapture::trace.active)rrGXWrite(m,cast<uint32_t>((o + cast<uint32_t>(3ULL))));
}
}}
}gc_Machine_profEnd(m,cast<int64_t>(3ULL),tConv);
gc_Machine_profFrame(m);
if (bool(m->OnFlip)) {
m->OnFlip(m);
}
time_Time tClear = gc_Machine_profStart(m);
if (clear) {
gc_gpu_clearEFB(g);
}
gc_Machine_profEnd(m,cast<int64_t>(3ULL),tClear);
}
}
}
// tools/platform/gc/gpu_efb.go:209:1
void gc_gpu_copyTexture(gc_gpu* g,gc_Machine* m,uint32_t params){
{
uint32_t tl = g->BP[cast<int64_t>(73ULL)];
uint32_t wh = g->BP[cast<int64_t>(74ULL)];
int64_t sx = cast<int64_t>(cast<uint32_t>((tl & cast<uint32_t>(1023ULL))));
int64_t sy = cast<int64_t>(cast<uint32_t>(((shr<uint32_t>(tl,cast<int64_t>(10ULL))) & cast<uint32_t>(1023ULL))));
int64_t w = cast<int64_t>((cast<int64_t>(cast<uint32_t>((wh & cast<uint32_t>(1023ULL)))) + cast<int64_t>(1ULL)));
int64_t h = cast<int64_t>((cast<int64_t>(cast<uint32_t>(((shr<uint32_t>(wh,cast<int64_t>(10ULL))) & cast<uint32_t>(1023ULL)))) + cast<int64_t>(1ULL)));
uint32_t dst = shl<uint32_t>((cast<uint32_t>((g->BP[cast<int64_t>(75ULL)] & cast<uint32_t>(16777215ULL)))),cast<int64_t>(5ULL));
int64_t stride = shl<int64_t>(cast<int64_t>(cast<uint32_t>((g->BP[cast<int64_t>(77ULL)] & cast<uint32_t>(1023ULL)))),cast<int64_t>(5ULL));
int64_t format = cast<int64_t>(cast<uint32_t>((cast<uint32_t>(((shr<uint32_t>(params,cast<int64_t>(4ULL))) & cast<uint32_t>(7ULL))) | shl<uint32_t>((cast<uint32_t>(((shr<uint32_t>(params,cast<int64_t>(3ULL))) & cast<uint32_t>(1ULL)))),cast<int64_t>(3ULL)))));
bool intensity = (cast<uint32_t>((params & cast<uint32_t>(32768ULL))) != cast<uint32_t>(0ULL));
bool halfScale = (cast<uint32_t>((params & cast<uint32_t>(512ULL))) != cast<uint32_t>(0ULL));
if (intensity) {
gekko_CPU_Halt(m->CPU,std::string("PE copy-to-texture: intensity format not yet implemented (params 0x%06X)",72),params);
return ;
}
if ((halfScale && (cast<uint32_t>((g->BP[cast<int64_t>(67ULL)] & cast<uint32_t>(7ULL))) == cast<uint32_t>(3ULL)))) {
gekko_CPU_Halt(m->CPU,std::string("PE copy-to-texture: half-scale copy from the Z buffer not yet implemented (params 0x%06X)",89),params);
return ;
}
if ((cast<uint32_t>((g->BP[cast<int64_t>(67ULL)] & cast<uint32_t>(7ULL))) == cast<uint32_t>(3ULL))) {
{
switch(format){
case cast<int64_t>(6ULL):{
gc_Machine_logf(m,std::string("PE copy-to-texture (Z24X8): EFB (%d,%d) %dx%d -> 0x%08X stride %d",65),sx,sy,w,h,dst,stride);
if (gc_drawTrace) {
go_fmt_Fprintf(go_os_Stderr,std::string("COPY-TEX-Z24X8 (%d,%d) %dx%d -> 0x%08X\012",39),sx,sy,w,h,dst);
}
{int64_t y = cast<int64_t>(0ULL);for (;(y < h);y++){
{int64_t x = cast<int64_t>(0ULL);for (;(x < w);x++){
uint32_t z = gc_gpu_zAt(g,cast<int64_t>((sx + x)),cast<int64_t>((sy + y)));
auto tmp70 = gc_copyTexelOffset(stride,cast<int64_t>(4ULL),cast<int64_t>(4ULL),cast<int64_t>(64ULL),x,y);
int64_t tb = std::get<0>(tmp70);
int64_t in = std::get<1>(tmp70);
gc_Machine_texWriteByte(m,cast<uint32_t>((dst + cast<uint32_t>(cast<int64_t>((tb + cast<int64_t>((in * cast<int64_t>(2ULL)))))))),cast<uint8_t>(255ULL));
gc_Machine_texWriteByte(m,cast<uint32_t>((cast<uint32_t>((dst + cast<uint32_t>(cast<int64_t>((tb + cast<int64_t>((in * cast<int64_t>(2ULL)))))))) + cast<uint32_t>(1ULL))),cast<uint8_t>(shr<uint32_t>(z,cast<int64_t>(16ULL))));
gc_Machine_texWriteByte(m,cast<uint32_t>((dst + cast<uint32_t>(cast<int64_t>((cast<int64_t>((tb + cast<int64_t>(32ULL))) + cast<int64_t>((in * cast<int64_t>(2ULL)))))))),cast<uint8_t>(shr<uint32_t>(z,cast<int64_t>(8ULL))));
gc_Machine_texWriteByte(m,cast<uint32_t>((cast<uint32_t>((dst + cast<uint32_t>(cast<int64_t>((cast<int64_t>((tb + cast<int64_t>(32ULL))) + cast<int64_t>((in * cast<int64_t>(2ULL)))))))) + cast<uint32_t>(1ULL))),cast<uint8_t>(z));
}
}}
}break;}
case cast<int64_t>(11ULL):{
gc_Machine_logf(m,std::string("PE copy-to-texture (Z16, fmt 0xB): EFB (%d,%d) %dx%d -> 0x%08X stride %d",72),sx,sy,w,h,dst,stride);
if (gc_drawTrace) {
go_fmt_Fprintf(go_os_Stderr,std::string("COPY-TEX-Z16 (%d,%d) %dx%d -> 0x%08X\012",37),sx,sy,w,h,dst);
}
{int64_t y = cast<int64_t>(0ULL);for (;(y < h);y++){
{int64_t x = cast<int64_t>(0ULL);for (;(x < w);x++){
uint32_t z = gc_gpu_zAt(g,cast<int64_t>((sx + x)),cast<int64_t>((sy + y)));
auto tmp71 = gc_copyTexelOffset(stride,cast<int64_t>(4ULL),cast<int64_t>(4ULL),cast<int64_t>(32ULL),x,y);
int64_t tb = std::get<0>(tmp71);
int64_t in = std::get<1>(tmp71);
uint32_t off = cast<uint32_t>((dst + cast<uint32_t>(cast<int64_t>((tb + cast<int64_t>((in * cast<int64_t>(2ULL))))))));
gc_Machine_texWriteByte(m,off,cast<uint8_t>(shr<uint32_t>(z,cast<int64_t>(8ULL))));
gc_Machine_texWriteByte(m,cast<uint32_t>((off + cast<uint32_t>(1ULL))),cast<uint8_t>(shr<uint32_t>(z,cast<int64_t>(16ULL))));
}
}}
}break;}
default:{
gekko_CPU_Halt(m->CPU,std::string("PE copy-to-texture: Z-source copy into format 0x%X not yet implemented (params 0x%06X)",86),format,params);
break;}
}}
return ;
}
auto tmp72 = std::make_tuple(w,h);
int64_t dw = std::get<0>(tmp72);
int64_t dh = std::get<1>(tmp72);
if (halfScale) {
auto tmp73 = std::make_tuple(divi<int64_t>(w,cast<int64_t>(2ULL)),divi<int64_t>(h,cast<int64_t>(2ULL)));
dw = std::get<0>(tmp73);
dh = std::get<1>(tmp73);
}
gc_Machine_logf(m,std::string("PE copy-to-texture: EFB (%d,%d) %dx%d -> 0x%08X format 0x%X stride %d halfScale=%v",82),sx,sy,w,h,dst,format,stride,halfScale);
if (gc_drawTrace) {
go_fmt_Fprintf(go_os_Stderr,std::string("COPY-TEX (%d,%d) %dx%d -> 0x%08X format 0x%X half=%v\012",53),sx,sy,w,h,dst,format,halfScale);
}
{int64_t y = cast<int64_t>(0ULL);for (;(y < dh);y++){
{int64_t x = cast<int64_t>(0ULL);for (;(x < dw);x++){
uint8_t r={};
uint8_t gg={};
uint8_t b={};
uint8_t a={};
if (halfScale) {
auto tmp74 = gc_gpu_efbBox2(g,cast<int64_t>((sx + cast<int64_t>((x * cast<int64_t>(2ULL))))),cast<int64_t>((sy + cast<int64_t>((y * cast<int64_t>(2ULL))))));
r = std::get<0>(tmp74);
gg = std::get<1>(tmp74);
b = std::get<2>(tmp74);
a = std::get<3>(tmp74);
}
else {
uint32_t px = gc_gpu_efbAt(g,cast<int64_t>((sx + x)),cast<int64_t>((sy + y)));
auto tmp75 = gc_unpackRGB(px);
r = std::get<0>(tmp75);
gg = std::get<1>(tmp75);
b = std::get<2>(tmp75);
a = cast<uint8_t>(px);
}
if ((!gc_gpu_encodeCopyTexel(g,m,dst,stride,format,x,y,r,gg,b,a))) {
gekko_CPU_Halt(m->CPU,std::string("PE copy-to-texture: destination format 0x%X not yet implemented (params 0x%06X)",79),format,params);
return ;
}
}
}}
}}
}
// tools/platform/gc/gpu_efb.go:337:1
std::tuple<uint8_t,uint8_t,uint8_t,uint8_t> gc_gpu_efbBox2(gc_gpu* g,int64_t x,int64_t y){
uint8_t r{};
uint8_t gg{};
uint8_t b{};
uint8_t a{};
{
uint32_t sr={};
uint32_t sg={};
uint32_t sb={};
uint32_t sa={};
{int64_t dy = cast<int64_t>(0ULL);for (;(dy < cast<int64_t>(2ULL));dy++){
{int64_t dx = cast<int64_t>(0ULL);for (;(dx < cast<int64_t>(2ULL));dx++){
uint32_t px = gc_gpu_efbAt(g,cast<int64_t>((x + dx)),cast<int64_t>((y + dy)));
auto tmp76 = gc_unpackRGB(px);
uint8_t pr = std::get<0>(tmp76);
uint8_t pg = std::get<1>(tmp76);
uint8_t pb = std::get<2>(tmp76);
sr += cast<uint32_t>(pr);
sg += cast<uint32_t>(pg);
sb += cast<uint32_t>(pb);
sa += cast<uint32_t>(cast<uint8_t>(px));
}
}}
}return {cast<uint8_t>(divi<uint32_t>((cast<uint32_t>((sr + cast<uint32_t>(2ULL)))),cast<uint32_t>(4ULL))),cast<uint8_t>(divi<uint32_t>((cast<uint32_t>((sg + cast<uint32_t>(2ULL)))),cast<uint32_t>(4ULL))),cast<uint8_t>(divi<uint32_t>((cast<uint32_t>((sb + cast<uint32_t>(2ULL)))),cast<uint32_t>(4ULL))),cast<uint8_t>(divi<uint32_t>((cast<uint32_t>((sa + cast<uint32_t>(2ULL)))),cast<uint32_t>(4ULL)))};
}
}
// tools/platform/gc/gpu_efb.go:355:1
std::tuple<int64_t,int64_t> gc_copyTexelOffset(int64_t stride,int64_t bw,int64_t bh,int64_t tileBytes,int64_t x,int64_t y){
int64_t tileBase{};
int64_t inTile{};
{
return {cast<int64_t>((cast<int64_t>(((divi<int64_t>(y,bh)) * stride)) + cast<int64_t>(((divi<int64_t>(x,bw)) * tileBytes)))),cast<int64_t>((cast<int64_t>(((modi<int64_t>(y,bh)) * bw)) + modi<int64_t>(x,bw)))};
}
}
// tools/platform/gc/gpu_efb.go:363:1
bool gc_gpu_encodeCopyTexel(gc_gpu* g,gc_Machine* m,uint32_t dst,int64_t stride,int64_t format,int64_t x,int64_t y,uint8_t r,uint8_t gg,uint8_t b,uint8_t a){
{
{
switch(format){
case cast<int64_t>(0ULL):{
auto tmp77 = gc_copyTexelOffset(stride,cast<int64_t>(8ULL),cast<int64_t>(8ULL),cast<int64_t>(32ULL),x,y);
int64_t tb = std::get<0>(tmp77);
int64_t in = std::get<1>(tmp77);
uint32_t off = cast<uint32_t>((dst + cast<uint32_t>(cast<int64_t>((tb + divi<int64_t>(in,cast<int64_t>(2ULL)))))));
uint8_t old = gc_Machine_texByte(m,off);
if ((cast<int64_t>((in & cast<int64_t>(1ULL))) == cast<int64_t>(0ULL))) {
gc_Machine_texWriteByte(m,off,cast<uint8_t>((cast<uint8_t>((old & cast<uint8_t>(15ULL))) | (cast<uint8_t>((r & cast<uint8_t>(240ULL)))))));
}
else {
gc_Machine_texWriteByte(m,off,cast<uint8_t>((cast<uint8_t>((old & cast<uint8_t>(240ULL))) | (shr<uint8_t>(r,cast<int64_t>(4ULL))))));
}
break;}
case cast<int64_t>(2ULL):{
auto tmp78 = gc_copyTexelOffset(stride,cast<int64_t>(8ULL),cast<int64_t>(4ULL),cast<int64_t>(32ULL),x,y);
int64_t tb = std::get<0>(tmp78);
int64_t in = std::get<1>(tmp78);
gc_Machine_texWriteByte(m,cast<uint32_t>((dst + cast<uint32_t>(cast<int64_t>((tb + in))))),cast<uint8_t>((cast<uint8_t>((a & cast<uint8_t>(240ULL))) | (shr<uint8_t>(r,cast<int64_t>(4ULL))))));
break;}
case cast<int64_t>(3ULL):{
auto tmp79 = gc_copyTexelOffset(stride,cast<int64_t>(4ULL),cast<int64_t>(4ULL),cast<int64_t>(32ULL),x,y);
int64_t tb = std::get<0>(tmp79);
int64_t in = std::get<1>(tmp79);
uint32_t off = cast<uint32_t>((dst + cast<uint32_t>(cast<int64_t>((tb + cast<int64_t>((in * cast<int64_t>(2ULL))))))));
gc_Machine_texWriteByte(m,off,a);
gc_Machine_texWriteByte(m,cast<uint32_t>((off + cast<uint32_t>(1ULL))),r);
break;}
case cast<int64_t>(4ULL):{
auto tmp80 = gc_copyTexelOffset(stride,cast<int64_t>(4ULL),cast<int64_t>(4ULL),cast<int64_t>(32ULL),x,y);
int64_t tb = std::get<0>(tmp80);
int64_t in = std::get<1>(tmp80);
uint16_t v = cast<uint16_t>((cast<uint16_t>((shl<uint16_t>(cast<uint16_t>(shr<uint8_t>(r,cast<int64_t>(3ULL))),cast<int64_t>(11ULL)) | shl<uint16_t>(cast<uint16_t>(shr<uint8_t>(gg,cast<int64_t>(2ULL))),cast<int64_t>(5ULL)))) | cast<uint16_t>(shr<uint8_t>(b,cast<int64_t>(3ULL)))));
uint32_t off = cast<uint32_t>((dst + cast<uint32_t>(cast<int64_t>((tb + cast<int64_t>((in * cast<int64_t>(2ULL))))))));
gc_Machine_texWriteByte(m,off,cast<uint8_t>(shr<uint16_t>(v,cast<int64_t>(8ULL))));
gc_Machine_texWriteByte(m,cast<uint32_t>((off + cast<uint32_t>(1ULL))),cast<uint8_t>(v));
break;}
case cast<int64_t>(5ULL):{
auto tmp81 = gc_copyTexelOffset(stride,cast<int64_t>(4ULL),cast<int64_t>(4ULL),cast<int64_t>(32ULL),x,y);
int64_t tb = std::get<0>(tmp81);
int64_t in = std::get<1>(tmp81);
uint16_t v={};
if ((a >= cast<uint8_t>(224ULL))) {
v = cast<uint16_t>((cast<uint16_t>((cast<uint16_t>((cast<uint16_t>(32768ULL) | shl<uint16_t>(cast<uint16_t>(shr<uint8_t>(r,cast<int64_t>(3ULL))),cast<int64_t>(10ULL)))) | shl<uint16_t>(cast<uint16_t>(shr<uint8_t>(gg,cast<int64_t>(3ULL))),cast<int64_t>(5ULL)))) | cast<uint16_t>(shr<uint8_t>(b,cast<int64_t>(3ULL)))));
}
else {
v = cast<uint16_t>((cast<uint16_t>((cast<uint16_t>((shl<uint16_t>(cast<uint16_t>(shr<uint8_t>(a,cast<int64_t>(5ULL))),cast<int64_t>(12ULL)) | shl<uint16_t>(cast<uint16_t>(shr<uint8_t>(r,cast<int64_t>(4ULL))),cast<int64_t>(8ULL)))) | shl<uint16_t>(cast<uint16_t>(shr<uint8_t>(gg,cast<int64_t>(4ULL))),cast<int64_t>(4ULL)))) | cast<uint16_t>(shr<uint8_t>(b,cast<int64_t>(4ULL)))));
}
uint32_t off = cast<uint32_t>((dst + cast<uint32_t>(cast<int64_t>((tb + cast<int64_t>((in * cast<int64_t>(2ULL))))))));
gc_Machine_texWriteByte(m,off,cast<uint8_t>(shr<uint16_t>(v,cast<int64_t>(8ULL))));
gc_Machine_texWriteByte(m,cast<uint32_t>((off + cast<uint32_t>(1ULL))),cast<uint8_t>(v));
break;}
case cast<int64_t>(6ULL):{
auto tmp82 = gc_copyTexelOffset(stride,cast<int64_t>(4ULL),cast<int64_t>(4ULL),cast<int64_t>(64ULL),x,y);
int64_t tb = std::get<0>(tmp82);
int64_t in = std::get<1>(tmp82);
gc_Machine_texWriteByte(m,cast<uint32_t>((dst + cast<uint32_t>(cast<int64_t>((tb + cast<int64_t>((in * cast<int64_t>(2ULL)))))))),a);
gc_Machine_texWriteByte(m,cast<uint32_t>((cast<uint32_t>((dst + cast<uint32_t>(cast<int64_t>((tb + cast<int64_t>((in * cast<int64_t>(2ULL)))))))) + cast<uint32_t>(1ULL))),r);
gc_Machine_texWriteByte(m,cast<uint32_t>((dst + cast<uint32_t>(cast<int64_t>((cast<int64_t>((tb + cast<int64_t>(32ULL))) + cast<int64_t>((in * cast<int64_t>(2ULL)))))))),gg);
gc_Machine_texWriteByte(m,cast<uint32_t>((cast<uint32_t>((dst + cast<uint32_t>(cast<int64_t>((cast<int64_t>((tb + cast<int64_t>(32ULL))) + cast<int64_t>((in * cast<int64_t>(2ULL)))))))) + cast<uint32_t>(1ULL))),b);
break;}
case cast<int64_t>(7ULL):case cast<int64_t>(8ULL):case cast<int64_t>(9ULL):case cast<int64_t>(10ULL):{
auto tmp83 = gc_copyTexelOffset(stride,cast<int64_t>(8ULL),cast<int64_t>(4ULL),cast<int64_t>(32ULL),x,y);
int64_t tb = std::get<0>(tmp83);
int64_t in = std::get<1>(tmp83);
uint8_t v={};
{
switch(format){
case cast<int64_t>(7ULL):{
v = a;
break;}
case cast<int64_t>(8ULL):{
v = r;
break;}
case cast<int64_t>(9ULL):{
v = gg;
break;}
default:{
v = b;
break;}
}}
gc_Machine_texWriteByte(m,cast<uint32_t>((dst + cast<uint32_t>(cast<int64_t>((tb + in))))),v);
break;}
default:{
return false;
break;}
}}
return true;
}
}
// tools/platform/gc/gpu_efb.go:427:1
void gc_Machine_texWriteByte(gc_Machine* m,uint32_t addr,uint8_t v){
{
uint32_t a = gc_phys(addr);
if ((cast<int64_t>(a) >= len(m->RAM))) {
return ;
}
m->RAM[a] = v;
if(rrcapture::trace.active)rrGXWrite(m,a);
}
}
// tools/platform/gc/gpu_efb.go:437:1
uint32_t gc_packRGBA(uint8_t r,uint8_t g,uint8_t b,uint8_t a){
{
return cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((shl<uint32_t>(cast<uint32_t>(r),cast<int64_t>(24ULL)) | shl<uint32_t>(cast<uint32_t>(g),cast<int64_t>(16ULL)))) | shl<uint32_t>(cast<uint32_t>(b),cast<int64_t>(8ULL)))) | cast<uint32_t>(a)));
}
}
// tools/platform/gc/gpu_efb.go:441:1
std::tuple<uint8_t,uint8_t,uint8_t> gc_unpackRGB(uint32_t px){
uint8_t r{};
uint8_t g{};
uint8_t b{};
{
return {cast<uint8_t>(shr<uint32_t>(px,cast<int64_t>(24ULL))),cast<uint8_t>(shr<uint32_t>(px,cast<int64_t>(16ULL))),cast<uint8_t>(shr<uint32_t>(px,cast<int64_t>(8ULL)))};
}
}
// tools/platform/gc/gpu_efb.go:449:1
uint8_t gc_luma(uint8_t r,uint8_t g,uint8_t b){
{
return gc_clamp8((((0.299 * cast<double>(r)) + (0.587 * cast<double>(g))) + (0.114 * cast<double>(b))));
}
}
// tools/platform/gc/gpu_efb.go:453:1
uint8_t gc_chromaB(uint8_t r,uint8_t g,uint8_t b){
{
double y = (((0.299 * cast<double>(r)) + (0.587 * cast<double>(g))) + (0.114 * cast<double>(b)));
return gc_clamp8((cast<double>(128ULL) + (((cast<double>(b) - y)) / 1.732)));
}
}
// tools/platform/gc/gpu_efb.go:458:1
uint8_t gc_chromaR(uint8_t r,uint8_t g,uint8_t b){
{
double y = (((0.299 * cast<double>(r)) + (0.587 * cast<double>(g))) + (0.114 * cast<double>(b)));
return gc_clamp8((cast<double>(128ULL) + (((cast<double>(r) - y)) / 1.371)));
}
}
// tools/platform/gc/gpu_fetch.go:107:1
gc_attrLayout gc_gpu_layout(gc_gpu* g,int64_t vat){
{
uint32_t lo = g->CPReg[cast<int64_t>(80ULL)];
uint32_t hi = g->CPReg[cast<int64_t>(96ULL)];
uint32_t g0 = g->CPReg[cast<uint32_t>((cast<uint32_t>(112ULL) + cast<uint32_t>(vat)))];
uint32_t g1 = g->CPReg[cast<uint32_t>((cast<uint32_t>(128ULL) + cast<uint32_t>(vat)))];
uint32_t g2 = g->CPReg[cast<uint32_t>((cast<uint32_t>(144ULL) + cast<uint32_t>(vat)))];
gc_attrLayout a = gc_attrLayout{{},{},cast<int64_t>(-1ULL),cast<int64_t>(-1ULL),{},{},{},{},cast<int64_t>(-1ULL),{},{},{},{},cast<int64_t>(-1ULL),{},{},{},{}};
{auto&& tmp84 = a.texMtxIdxOff;
for(int64_t tmp85=0;tmp85<len(tmp84);++tmp85){
auto i=tmp85;a.texMtxIdxOff[i] = cast<int64_t>(-1ULL);
a.tex[i].off = cast<int64_t>(-1ULL);
}}
int64_t off = cast<int64_t>(0ULL);
if ((cast<uint32_t>((lo & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL))) {
a.hasMatIdx = true;
a.matIdxOff = off;
off++;
}
{uint32_t b = cast<uint32_t>(1ULL);for (;(b < cast<uint32_t>(9ULL));b++){
if ((cast<uint32_t>((lo & (shl<uint32_t>(cast<uint32_t>(1ULL),b)))) != cast<uint32_t>(0ULL))) {
a.texMtxIdxOff[cast<uint32_t>((b - cast<uint32_t>(1ULL)))] = off;
off++;
}
}
}auto sizeOfN = [&](uint32_t desc,int64_t direct,int64_t nIdx)->int64_t{
{
switch(desc){
case cast<uint32_t>(2ULL):{
return nIdx;
break;}
case cast<uint32_t>(3ULL):{
return cast<int64_t>((cast<int64_t>(2ULL) * nIdx));
break;}
case cast<uint32_t>(1ULL):{
return direct;
break;}
}}
return cast<int64_t>(0ULL);
}
;
auto sizeOf = [&](uint32_t desc,int64_t direct)->int64_t{
return sizeOfN(desc,direct,cast<int64_t>(1ULL));
}
;
a.posDesc = cast<uint32_t>(((shr<uint32_t>(lo,cast<int64_t>(9ULL))) & cast<uint32_t>(3ULL)));
a.posFmt = cast<uint32_t>(((shr<uint32_t>(g0,cast<int64_t>(1ULL))) & cast<uint32_t>(7ULL)));
a.posComps = cast<int64_t>(2ULL);
if ((cast<uint32_t>((g0 & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL))) {
a.posComps = cast<int64_t>(3ULL);
}
a.posFrac = cast<uint32_t>(((shr<uint32_t>(g0,cast<int64_t>(4ULL))) & cast<uint32_t>(31ULL)));
if ((a.posDesc != cast<uint32_t>(0ULL))) {
a.posOff = off;
off += sizeOf(a.posDesc,cast<int64_t>((a.posComps * gc_componentBytes(a.posFmt))));
}
a.nrmDesc = cast<uint32_t>(((shr<uint32_t>(lo,cast<int64_t>(11ULL))) & cast<uint32_t>(3ULL)));
a.nrmFmt = cast<uint32_t>(((shr<uint32_t>(g0,cast<int64_t>(10ULL))) & cast<uint32_t>(7ULL)));
a.nrmComps = cast<int64_t>(3ULL);
if ((cast<uint32_t>(((shr<uint32_t>(g0,cast<int64_t>(9ULL))) & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL))) {
a.nrmComps = cast<int64_t>(9ULL);
}
a.nrmIdxComps = a.nrmComps;
if ((gc_normalIndexCount(g0) == cast<int64_t>(3ULL))) {
a.nrmIdxComps = cast<int64_t>(3ULL);
}
if ((a.nrmDesc != cast<uint32_t>(0ULL))) {
a.nrmOff = off;
off += sizeOfN(a.nrmDesc,cast<int64_t>((a.nrmComps * gc_componentBytes(a.nrmFmt))),gc_normalIndexCount(g0));
}
a.col0Desc = cast<uint32_t>(((shr<uint32_t>(lo,cast<int64_t>(13ULL))) & cast<uint32_t>(3ULL)));
a.col0Comp = cast<uint32_t>(((shr<uint32_t>(g0,cast<int64_t>(14ULL))) & cast<uint32_t>(7ULL)));
if ((a.col0Desc != cast<uint32_t>(0ULL))) {
a.col0Off = off;
off += sizeOf(a.col0Desc,gc_colorBytes(a.col0Comp));
}
off += sizeOf(cast<uint32_t>(((shr<uint32_t>(lo,cast<int64_t>(15ULL))) & cast<uint32_t>(3ULL))),gc_colorBytes(cast<uint32_t>(((shr<uint32_t>(g0,cast<int64_t>(18ULL))) & cast<uint32_t>(7ULL)))));
auto texBytes = [&](uint32_t elem,uint32_t format)->int64_t{
int64_t comps = cast<int64_t>(1ULL);
if ((elem != cast<uint32_t>(0ULL))) {
comps = cast<int64_t>(2ULL);
}
return cast<int64_t>((comps * gc_componentBytes(format)));
}
;
auto tex = [&](int64_t k)->uint32_t{
return cast<uint32_t>(((shr<uint32_t>(hi,(cast<uint32_t>((cast<uint32_t>(2ULL) * cast<uint32_t>(k)))))) & cast<uint32_t>(3ULL)));
}
;
std::array<std::array<uint32_t,3>,8> texFields = std::array<std::array<uint32_t,3>,8>{std::array<uint32_t,3>{cast<uint32_t>(((shr<uint32_t>(g0,cast<int64_t>(21ULL))) & cast<uint32_t>(1ULL))),cast<uint32_t>(((shr<uint32_t>(g0,cast<int64_t>(22ULL))) & cast<uint32_t>(7ULL))),cast<uint32_t>(((shr<uint32_t>(g0,cast<int64_t>(25ULL))) & cast<uint32_t>(31ULL)))},std::array<uint32_t,3>{cast<uint32_t>(((shr<uint32_t>(g1,cast<int64_t>(0ULL))) & cast<uint32_t>(1ULL))),cast<uint32_t>(((shr<uint32_t>(g1,cast<int64_t>(1ULL))) & cast<uint32_t>(7ULL))),cast<uint32_t>(((shr<uint32_t>(g1,cast<int64_t>(4ULL))) & cast<uint32_t>(31ULL)))},std::array<uint32_t,3>{cast<uint32_t>(((shr<uint32_t>(g1,cast<int64_t>(9ULL))) & cast<uint32_t>(1ULL))),cast<uint32_t>(((shr<uint32_t>(g1,cast<int64_t>(10ULL))) & cast<uint32_t>(7ULL))),cast<uint32_t>(((shr<uint32_t>(g1,cast<int64_t>(13ULL))) & cast<uint32_t>(31ULL)))},std::array<uint32_t,3>{cast<uint32_t>(((shr<uint32_t>(g1,cast<int64_t>(18ULL))) & cast<uint32_t>(1ULL))),cast<uint32_t>(((shr<uint32_t>(g1,cast<int64_t>(19ULL))) & cast<uint32_t>(7ULL))),cast<uint32_t>(((shr<uint32_t>(g1,cast<int64_t>(22ULL))) & cast<uint32_t>(31ULL)))},std::array<uint32_t,3>{cast<uint32_t>(((shr<uint32_t>(g1,cast<int64_t>(27ULL))) & cast<uint32_t>(1ULL))),cast<uint32_t>(((shr<uint32_t>(g1,cast<int64_t>(28ULL))) & cast<uint32_t>(7ULL))),cast<uint32_t>(((shr<uint32_t>(g2,cast<int64_t>(0ULL))) & cast<uint32_t>(31ULL)))},std::array<uint32_t,3>{cast<uint32_t>(((shr<uint32_t>(g2,cast<int64_t>(5ULL))) & cast<uint32_t>(1ULL))),cast<uint32_t>(((shr<uint32_t>(g2,cast<int64_t>(6ULL))) & cast<uint32_t>(7ULL))),cast<uint32_t>(((shr<uint32_t>(g2,cast<int64_t>(9ULL))) & cast<uint32_t>(31ULL)))},std::array<uint32_t,3>{cast<uint32_t>(((shr<uint32_t>(g2,cast<int64_t>(14ULL))) & cast<uint32_t>(1ULL))),cast<uint32_t>(((shr<uint32_t>(g2,cast<int64_t>(15ULL))) & cast<uint32_t>(7ULL))),cast<uint32_t>(((shr<uint32_t>(g2,cast<int64_t>(18ULL))) & cast<uint32_t>(31ULL)))},std::array<uint32_t,3>{cast<uint32_t>(((shr<uint32_t>(g2,cast<int64_t>(23ULL))) & cast<uint32_t>(1ULL))),cast<uint32_t>(((shr<uint32_t>(g2,cast<int64_t>(24ULL))) & cast<uint32_t>(7ULL))),cast<uint32_t>(((shr<uint32_t>(g2,cast<int64_t>(27ULL))) & cast<uint32_t>(31ULL)))}};
{int64_t k = cast<int64_t>(0ULL);for (;(k < cast<int64_t>(8ULL));k++){
gc_texAttr* t = (&a.tex[k]);
t->desc = tex(k);
auto tmp86 = std::make_tuple(texFields[k][cast<int64_t>(0ULL)],texFields[k][cast<int64_t>(1ULL)],texFields[k][cast<int64_t>(2ULL)]);
t->elem = std::get<0>(tmp86);
t->fmt = std::get<1>(tmp86);
t->frac = std::get<2>(tmp86);
if ((t->desc != cast<uint32_t>(0ULL))) {
t->off = off;
off += sizeOf(t->desc,texBytes(t->elem,t->fmt));
}
}
}a.stride = off;
return a;
}
}
// tools/platform/gc/gpu_fetch.go:235:1
Slice<uint8_t> gc_gpu_attrData(gc_gpu* g,gc_Machine* m,Slice<uint8_t> v,int64_t off,uint32_t desc,int64_t attr,int64_t size){
{
{
switch(desc){
case cast<uint32_t>(1ULL):{
return sub(v,off,len(v));
break;}
case cast<uint32_t>(2ULL):case cast<uint32_t>(3ULL):{
int64_t idx = cast<int64_t>(v[off]);
if ((desc == cast<uint32_t>(3ULL))) {
idx = cast<int64_t>(gc_be16(rrBorrow(v,off,len(v))));
}
uint32_t base = g->CPReg[cast<uint32_t>((cast<uint32_t>(160ULL) + cast<uint32_t>(attr)))];
uint32_t stride = cast<uint32_t>((g->CPReg[cast<uint32_t>((cast<uint32_t>(176ULL) + cast<uint32_t>(attr)))] & cast<uint32_t>(255ULL)));
uint32_t a = gc_phys(cast<uint32_t>((base + cast<uint32_t>((cast<uint32_t>(idx) * stride)))));
if ((cast<int64_t>((cast<int64_t>(a) + size)) > len(m->RAM))) {
return {};
}
return sub(m->RAM,a,cast<int64_t>((cast<int64_t>(a) + size)));
break;}
}}
return {};
}
}
// tools/platform/gc/gpu_fetch.go:256:1
void gc_gpu_drawPrimitive(gc_gpu* g,gc_Machine* m,uint32_t prim,int64_t vat,int64_t vsize,Slice<uint8_t> data){
{rrprof::Scope timing(2,"Vertices and transforms");
{
gc_attrLayout lay = gc_gpu_layout(g,vat);
if ((lay.posDesc == cast<uint32_t>(0ULL))) {
gc_Machine_logf(m,std::string("CP: draw with no position attribute (primitive 0x%02X)",54),prim);
return ;
}
time_Time tv = gc_Machine_profStart(m);
g->profDraws++;
int64_t count = divi<int64_t>(len(data),vsize);
Slice<gc_clipVertex> verts = Slice<gc_clipVertex>::make(count);
int64_t nTexGen = gc_gpu_texGenCount(g);
{int64_t i = cast<int64_t>(0ULL);for (;(i < count);i++){
Slice<uint8_t> v = sub(data,cast<int64_t>((i * vsize)),len(data));
int64_t mtxIdx = cast<int64_t>(cast<uint32_t>((g->CPReg[cast<int64_t>(48ULL)] & cast<uint32_t>(63ULL))));
if (lay.hasMatIdx) {
mtxIdx = cast<int64_t>(v[lay.matIdxOff]);
}
Slice<uint8_t> pos = gc_gpu_attrData(g,m,v,lay.posOff,lay.posDesc,cast<int64_t>(0ULL),cast<int64_t>((lay.posComps * gc_componentBytes(lay.posFmt))));
if ((!pos)) {
gc_Machine_logf(m,std::string("CP: draw's position index reads past RAM (primitive 0x%02X)",59),prim);
return ;
}
auto tmp87 = gc_gpu_readPos(g,pos,lay);
float mx = std::get<0>(tmp87);
float my = std::get<1>(tmp87);
float mz = std::get<2>(tmp87);
auto tmp88 = gc_gpu_eyePos(g,mtxIdx,mx,my,mz);
float ex = std::get<0>(tmp88);
float ey = std::get<1>(tmp88);
float ez = std::get<2>(tmp88);
auto tmp89 = gc_gpu_clipPos(g,ex,ey,ez);
float cx = std::get<0>(tmp89);
float cy = std::get<1>(tmp89);
float cz = std::get<2>(tmp89);
float cw = std::get<3>(tmp89);
auto tmp90 = std::make_tuple(cast<uint8_t>(255ULL),cast<uint8_t>(255ULL),cast<uint8_t>(255ULL),cast<uint8_t>(255ULL));
uint8_t r = std::get<0>(tmp90);
uint8_t gg = std::get<1>(tmp90);
uint8_t b = std::get<2>(tmp90);
uint8_t a = std::get<3>(tmp90);
if ((lay.col0Off >= cast<int64_t>(0ULL))) {
{
Slice<uint8_t> col = gc_gpu_attrData(g,m,v,lay.col0Off,lay.col0Desc,cast<int64_t>(2ULL),gc_colorBytes(lay.col0Comp));
if (bool(col)) {
auto tmp91 = gc_readColor(col,lay.col0Comp);
r = std::get<0>(tmp91);
gg = std::get<1>(tmp91);
b = std::get<2>(tmp91);
a = std::get<3>(tmp91);
}
}
}
float nex={};
float ney={};
float nez={};
float mnx={};
float mny={};
float mnz={};
if ((lay.nrmOff >= cast<int64_t>(0ULL))) {
{
Slice<uint8_t> nb = gc_gpu_attrData(g,m,v,lay.nrmOff,lay.nrmDesc,cast<int64_t>(1ULL),cast<int64_t>((lay.nrmIdxComps * gc_componentBytes(lay.nrmFmt))));
if (bool(nb)) {
auto tmp92 = gc_readNormal(nb,lay.nrmFmt);
mnx = std::get<0>(tmp92);
mny = std::get<1>(tmp92);
mnz = std::get<2>(tmp92);
auto tmp93 = gc_gpu_normalToEye(g,mtxIdx,mnx,mny,mnz);
nex = std::get<0>(tmp93);
ney = std::get<1>(tmp93);
nez = std::get<2>(tmp93);
}
}
}
std::array<std::array<uint8_t,4>,2> col = gc_gpu_rasChannel(g,m,r,gg,b,a,ex,ey,ez,nex,ney,nez);
std::array<gc_texCoord,8> vtc={};
{int64_t k = cast<int64_t>(0ULL);for (;(k < cast<int64_t>(8ULL));k++){
gc_texAttr* ta = (&lay.tex[k]);
if ((ta->off < cast<int64_t>(0ULL))) {
continue;
}
int64_t comps = cast<int64_t>(1ULL);
if ((ta->elem != cast<uint32_t>(0ULL))) {
comps = cast<int64_t>(2ULL);
}
{
Slice<uint8_t> tc = gc_gpu_attrData(g,m,v,ta->off,ta->desc,cast<int64_t>((cast<int64_t>(4ULL) + k)),cast<int64_t>((comps * gc_componentBytes(ta->fmt))));
if (bool(tc)) {
auto tmp94 = gc_readTexCoord(tc,ta);
float u = std::get<0>(tmp94);
float vv = std::get<1>(tmp94);
vtc[k] = gc_texCoord{cast<float>(u),cast<float>(vv),cast<float>(cast<float>(1ULL))};
}
}
}
}gc_clipVertex cv = gc_clipVertex{cast<float>(cx),cast<float>(cy),cast<float>(cz),cast<float>(cw),col,{},nTexGen};
{int64_t k = cast<int64_t>(0ULL);for (;(k < nTexGen);k++){
int64_t row = gc_gpu_texMtxRow(g,k);
if ((lay.texMtxIdxOff[k] >= cast<int64_t>(0ULL))) {
row = cast<int64_t>(v[lay.texMtxIdxOff[k]]);
}
cv.tc[k] = gc_gpu_genTexCoord(g,m,k,row,mx,my,mz,mnx,mny,mnz,(&vtc),(&col));
}
}verts[i] = cv;
}
}gc_Machine_profEnd(m,cast<int64_t>(1ULL),tv);
if ((gc_drawTrace && (count > cast<int64_t>(0ULL)))) {
auto tmp95 = std::make_tuple(g->pixWritten,g->pixZRej,g->pixARej);
int64_t w0 = std::get<0>(tmp95);
int64_t z0 = std::get<1>(tmp95);
int64_t a0 = std::get<2>(tmp95);
time_Time tr = gc_Machine_profStart(m);
gc_gpu_rasterPrimitive(g,m,prim,verts);
gc_Machine_profEnd(m,cast<int64_t>(2ULL),tr);
gc_clipVertex c0 = verts[cast<int64_t>(0ULL)];
auto tmp96 = gc_gpu_toScreen(g,c0.cx,c0.cy,c0.cz,c0.cw);
float v0x = std::get<0>(tmp96);
float v0y = std::get<1>(tmp96);
float v0z = std::get<2>(tmp96);
gc_screenVertex v0 = gc_screenVertex{cast<float>(v0x),cast<float>(v0y),cast<float>(v0z),c0.col,c0.tc,c0.ntc,{}};
std::string clipped = std::string("",0);
if (((c0.cz + c0.cw) < cast<float>(0ULL))) {
clipped = go_fmt_Sprintf(std::string(" CLIPPED(w=%.2f)",16),c0.cw);
}
auto tmp97 = gc_tevColorReg(g->TevColorReg[cast<int64_t>(1ULL)]);
float c0a = std::get<3>(tmp97);
auto tmp98 = gc_tevColorReg(g->TevKonstReg[cast<int64_t>(0ULL)]);
float k0a = std::get<3>(tmp98);
gc_texState t0 = gc_gpu_texSetup(g,cast<int64_t>(0ULL));
gc_gpu_dumpTex0Once(g,m,t0.base);
go_fmt_Fprintf(go_os_Stderr,std::string("DRAW prim 0x%02X vat %d n %d  v0 (%.1f,%.1f,z%.0f)%s rgba %d,%d,%d,%d ntc=%d uv (%.2f,%.2f)  tex0 0x%06X fmt%X %dx%d  px w=%d zrej=%d arej=%d  stages=%d cull=%d bp41=%06X af=%06X zm=%02X a0=%.0f ka0=%.0f vcd=%03X mat0=%08X amb0=%08X cc0=%08X ca0=%08X tev0=%06X/%06X\012",266),prim,vat,count,v0.x,v0.y,v0.z,clipped,v0.col[cast<int64_t>(0ULL)][cast<int64_t>(0ULL)],v0.col[cast<int64_t>(0ULL)][cast<int64_t>(1ULL)],v0.col[cast<int64_t>(0ULL)][cast<int64_t>(2ULL)],v0.col[cast<int64_t>(0ULL)][cast<int64_t>(3ULL)],v0.ntc,v0.tc[cast<int64_t>(0ULL)].s,v0.tc[cast<int64_t>(0ULL)].t,t0.base,t0.format,t0.width,t0.height,cast<int64_t>((g->pixWritten - w0)),cast<int64_t>((g->pixZRej - z0)),cast<int64_t>((g->pixARej - a0)),cast<int64_t>((cast<int64_t>(cast<uint32_t>(((shr<uint32_t>(g->BP[cast<int64_t>(0ULL)],cast<int64_t>(10ULL))) & cast<uint32_t>(15ULL)))) + cast<int64_t>(1ULL))),cast<uint32_t>(((shr<uint32_t>(g->BP[cast<int64_t>(0ULL)],cast<int64_t>(14ULL))) & cast<uint32_t>(3ULL))),g->BP[cast<int64_t>(65ULL)],g->BP[cast<int64_t>(243ULL)],g->BP[cast<int64_t>(64ULL)],c0a,k0a,cast<uint32_t>((g->CPReg[cast<int64_t>(80ULL)] & cast<uint32_t>(65535ULL))),g->XFMem[cast<int64_t>(4108ULL)],g->XFMem[cast<int64_t>(4106ULL)],g->XFMem[cast<int64_t>(4110ULL)],g->XFMem[cast<int64_t>(4112ULL)],g->BP[cast<int64_t>(192ULL)],g->BP[cast<int64_t>(193ULL)]);
int64_t scisY0 = cast<int64_t>((cast<int64_t>(cast<uint32_t>((g->BP[cast<int64_t>(32ULL)] & cast<uint32_t>(2047ULL)))) - cast<int64_t>(342ULL)));
int64_t scisX0 = cast<int64_t>((cast<int64_t>(cast<uint32_t>(((shr<uint32_t>(g->BP[cast<int64_t>(32ULL)],cast<int64_t>(12ULL))) & cast<uint32_t>(2047ULL)))) - cast<int64_t>(342ULL)));
int64_t scisY1 = cast<int64_t>((cast<int64_t>(cast<uint32_t>((g->BP[cast<int64_t>(33ULL)] & cast<uint32_t>(2047ULL)))) - cast<int64_t>(342ULL)));
int64_t scisX1 = cast<int64_t>((cast<int64_t>(cast<uint32_t>(((shr<uint32_t>(g->BP[cast<int64_t>(33ULL)],cast<int64_t>(12ULL))) & cast<uint32_t>(2047ULL)))) - cast<int64_t>(342ULL)));
go_fmt_Fprintf(go_os_Stderr,std::string("  SCIS box (%d,%d)-(%d,%d) off=(%d,%d) raw20=%06X raw21=%06X raw59=%06X\012",72),scisX0,scisY0,scisX1,scisY1,cast<int64_t>((cast<int64_t>(((cast<int64_t>(cast<uint32_t>((g->BP[cast<int64_t>(89ULL)] & cast<uint32_t>(1023ULL))))) * cast<int64_t>(2ULL))) - cast<int64_t>(342ULL))),cast<int64_t>((cast<int64_t>(((cast<int64_t>(cast<uint32_t>(((shr<uint32_t>(g->BP[cast<int64_t>(89ULL)],cast<int64_t>(10ULL))) & cast<uint32_t>(1023ULL))))) * cast<int64_t>(2ULL))) - cast<int64_t>(342ULL))),g->BP[cast<int64_t>(32ULL)],g->BP[cast<int64_t>(33ULL)],g->BP[cast<int64_t>(89ULL)]);
go_fmt_Fprintf(go_os_Stderr,std::string("  XFDBG nchan=%d cc1=%08X ca1=%08X mat1=%08X amb1=%08X ntg=%d tg=[%08X %08X %08X %08X %08X %08X %08X %08X] dual=%X ord=[%06X %06X %06X %06X]\012",141),g->XFMem[cast<int64_t>(4105ULL)],g->XFMem[cast<int64_t>(4111ULL)],g->XFMem[cast<int64_t>(4113ULL)],g->XFMem[cast<int64_t>(4109ULL)],g->XFMem[cast<int64_t>(4107ULL)],gc_gpu_texGenCount(g),g->XFMem[cast<int64_t>(4160ULL)],g->XFMem[cast<int64_t>(4161ULL)],g->XFMem[cast<int64_t>(4162ULL)],g->XFMem[cast<int64_t>(4163ULL)],g->XFMem[cast<int64_t>(4164ULL)],g->XFMem[cast<int64_t>(4165ULL)],g->XFMem[cast<int64_t>(4166ULL)],g->XFMem[cast<int64_t>(4167ULL)],g->XFMem[cast<int64_t>(4114ULL)],g->BP[cast<int64_t>(40ULL)],g->BP[cast<int64_t>(41ULL)],g->BP[cast<int64_t>(42ULL)],g->BP[cast<int64_t>(43ULL)]);
{int64_t li = cast<int64_t>(0ULL);for (;(li < cast<int64_t>(8ULL));li++){
if (((cast<uint32_t>((gc_lightMask(g->XFMem[cast<int64_t>(4110ULL)]) & (shl<uint32_t>(cast<uint32_t>(1ULL),li)))) == cast<uint32_t>(0ULL)) && (cast<uint32_t>((gc_lightMask(g->XFMem[cast<int64_t>(4111ULL)]) & (shl<uint32_t>(cast<uint32_t>(1ULL),li)))) == cast<uint32_t>(0ULL)))) {
continue;
}
int64_t b = cast<int64_t>((cast<int64_t>(1536ULL) + cast<int64_t>((li * cast<int64_t>(16ULL)))));
go_fmt_Fprintf(go_os_Stderr,std::string("  LIGHT%d col=%08X a=(%g,%g,%g) k=(%g,%g,%g) pos=(%.2f,%.2f,%.2f) dir=(%.4f,%.4f,%.4f)\012",87),li,g->XFMem[cast<int64_t>((b + cast<int64_t>(3ULL)))],gc_gpu_xfFloat(g,cast<int64_t>((b + cast<int64_t>(4ULL)))),gc_gpu_xfFloat(g,cast<int64_t>((b + cast<int64_t>(5ULL)))),gc_gpu_xfFloat(g,cast<int64_t>((b + cast<int64_t>(6ULL)))),gc_gpu_xfFloat(g,cast<int64_t>((b + cast<int64_t>(7ULL)))),gc_gpu_xfFloat(g,cast<int64_t>((b + cast<int64_t>(8ULL)))),gc_gpu_xfFloat(g,cast<int64_t>((b + cast<int64_t>(9ULL)))),gc_gpu_xfFloat(g,cast<int64_t>((b + cast<int64_t>(10ULL)))),gc_gpu_xfFloat(g,cast<int64_t>((b + cast<int64_t>(11ULL)))),gc_gpu_xfFloat(g,cast<int64_t>((b + cast<int64_t>(12ULL)))),gc_gpu_xfFloat(g,cast<int64_t>((b + cast<int64_t>(13ULL)))),gc_gpu_xfFloat(g,cast<int64_t>((b + cast<int64_t>(14ULL)))),gc_gpu_xfFloat(g,cast<int64_t>((b + cast<int64_t>(15ULL)))));
}
}return ;
}
time_Time tr = gc_Machine_profStart(m);
gc_gpu_rasterPrimitive(g,m,prim,verts);
gc_Machine_profEnd(m,cast<int64_t>(2ULL),tr);
}
}
}
// tools/platform/gc/gpu_fetch.go:411:1
std::tuple<float,float> gc_readTexCoord(Slice<uint8_t> b,gc_texAttr* ta){
float u{};
float v{};
{
float scale = (cast<float>(cast<float>(1ULL)) / cast<float>(shl<uint32_t>(cast<uint32_t>(1ULL),ta->frac)));
int64_t sz = gc_componentBytes(ta->fmt);
u = (gc_readComponent(b,ta->fmt) * scale);
if ((ta->elem != cast<uint32_t>(0ULL))) {
v = (gc_readComponent(rrBorrow(b,sz,len(b)),ta->fmt) * scale);
}
return {u,v};
}
}
// tools/platform/gc/gpu_fetch.go:424:1
std::tuple<float,float,float> gc_readNormal(Slice<uint8_t> b,uint32_t format){
float x{};
float y{};
float z{};
{
float scale = cast<float>(cast<float>(1ULL));
{
switch(format){
case cast<uint32_t>(1ULL):{
scale = (1.0 / cast<double>(64ULL));
break;}
case cast<uint32_t>(3ULL):{
scale = (1.0 / cast<double>(16384ULL));
break;}
}}
int64_t sz = gc_componentBytes(format);
x = (gc_readComponent(b,format) * scale);
y = (gc_readComponent(rrBorrow(b,sz,len(b)),format) * scale);
z = (gc_readComponent(rrBorrow(b,cast<int64_t>((cast<int64_t>(2ULL) * sz)),len(b)),format) * scale);
return {x,y,z};
}
}
// tools/platform/gc/gpu_fetch.go:440:1
std::tuple<float,float,float> gc_gpu_readPos(gc_gpu* g,Slice<uint8_t> b,gc_attrLayout lay){
float x{};
float y{};
float z{};
{
std::array<float,3> c={};
float scale = (cast<float>(cast<float>(1ULL)) / cast<float>(shl<uint32_t>(cast<uint32_t>(1ULL),lay.posFrac)));
int64_t sz = gc_componentBytes(lay.posFmt);
{int64_t k = cast<int64_t>(0ULL);for (;(k < lay.posComps);k++){
c[k] = (gc_readComponent(rrBorrow(b,cast<int64_t>((k * sz)),len(b)),lay.posFmt) * scale);
}
}return {c[cast<int64_t>(0ULL)],c[cast<int64_t>(1ULL)],c[cast<int64_t>(2ULL)]};
}
}
// tools/platform/gc/gpu_fetch.go:451:1
float gc_readComponent(Slice<uint8_t> b,uint32_t format){
{
{
switch(format){
case cast<uint32_t>(0ULL):{
return cast<float>(b[cast<int64_t>(0ULL)]);
break;}
case cast<uint32_t>(1ULL):{
return cast<float>(cast<int8_t>(b[cast<int64_t>(0ULL)]));
break;}
case cast<uint32_t>(2ULL):{
return cast<float>(cast<uint16_t>((shl<uint16_t>(cast<uint16_t>(b[cast<int64_t>(0ULL)]),cast<int64_t>(8ULL)) | cast<uint16_t>(b[cast<int64_t>(1ULL)]))));
break;}
case cast<uint32_t>(3ULL):{
return cast<float>(cast<int16_t>(cast<uint16_t>((shl<uint16_t>(cast<uint16_t>(b[cast<int64_t>(0ULL)]),cast<int64_t>(8ULL)) | cast<uint16_t>(b[cast<int64_t>(1ULL)])))));
break;}
case cast<uint32_t>(4ULL):{
return go_math_Float32frombits(cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((shl<uint32_t>(cast<uint32_t>(b[cast<int64_t>(0ULL)]),cast<int64_t>(24ULL)) | shl<uint32_t>(cast<uint32_t>(b[cast<int64_t>(1ULL)]),cast<int64_t>(16ULL)))) | shl<uint32_t>(cast<uint32_t>(b[cast<int64_t>(2ULL)]),cast<int64_t>(8ULL)))) | cast<uint32_t>(b[cast<int64_t>(3ULL)]))));
break;}
}}
return cast<float>(0ULL);
}
}
// tools/platform/gc/gpu_fetch.go:468:1
std::tuple<uint8_t,uint8_t,uint8_t,uint8_t> gc_readColor(Slice<uint8_t> b,uint32_t comp){
uint8_t r{};
uint8_t g{};
uint8_t bl{};
uint8_t a{};
{
{
switch(comp){
case cast<uint32_t>(0ULL):{
uint16_t v = cast<uint16_t>((shl<uint16_t>(cast<uint16_t>(b[cast<int64_t>(0ULL)]),cast<int64_t>(8ULL)) | cast<uint16_t>(b[cast<int64_t>(1ULL)])));
return {gc_expand5(shr<uint16_t>(v,cast<int64_t>(11ULL))),gc_expand6(cast<uint16_t>(((shr<uint16_t>(v,cast<int64_t>(5ULL))) & cast<uint16_t>(63ULL)))),gc_expand5(cast<uint16_t>((v & cast<uint16_t>(31ULL)))),cast<uint8_t>(255ULL)};
break;}
case cast<uint32_t>(1ULL):{
return {b[cast<int64_t>(0ULL)],b[cast<int64_t>(1ULL)],b[cast<int64_t>(2ULL)],cast<uint8_t>(255ULL)};
break;}
case cast<uint32_t>(2ULL):{
return {b[cast<int64_t>(0ULL)],b[cast<int64_t>(1ULL)],b[cast<int64_t>(2ULL)],cast<uint8_t>(255ULL)};
break;}
case cast<uint32_t>(3ULL):{
uint16_t v = cast<uint16_t>((shl<uint16_t>(cast<uint16_t>(b[cast<int64_t>(0ULL)]),cast<int64_t>(8ULL)) | cast<uint16_t>(b[cast<int64_t>(1ULL)])));
return {gc_expand4(shr<uint16_t>(v,cast<int64_t>(12ULL))),gc_expand4(cast<uint16_t>(((shr<uint16_t>(v,cast<int64_t>(8ULL))) & cast<uint16_t>(15ULL)))),gc_expand4(cast<uint16_t>(((shr<uint16_t>(v,cast<int64_t>(4ULL))) & cast<uint16_t>(15ULL)))),gc_expand4(cast<uint16_t>((v & cast<uint16_t>(15ULL))))};
break;}
case cast<uint32_t>(5ULL):{
return {b[cast<int64_t>(0ULL)],b[cast<int64_t>(1ULL)],b[cast<int64_t>(2ULL)],b[cast<int64_t>(3ULL)]};
break;}
}}
return {cast<uint8_t>(255ULL),cast<uint8_t>(255ULL),cast<uint8_t>(255ULL),cast<uint8_t>(255ULL)};
}
}
// tools/platform/gc/gpu_fetch.go:486:1
uint8_t gc_expand4(uint16_t v){
{
return cast<uint8_t>(cast<uint16_t>((shl<uint16_t>(v,cast<int64_t>(4ULL)) | v)));
}
}
// tools/platform/gc/gpu_fetch.go:487:1
uint8_t gc_expand5(uint16_t v){
{
return cast<uint8_t>(cast<uint16_t>((shl<uint16_t>(v,cast<int64_t>(3ULL)) | shr<uint16_t>(v,cast<int64_t>(2ULL)))));
}
}
// tools/platform/gc/gpu_fetch.go:488:1
uint8_t gc_expand6(uint16_t v){
{
return cast<uint8_t>(cast<uint16_t>((shl<uint16_t>(v,cast<int64_t>(2ULL)) | shr<uint16_t>(v,cast<int64_t>(4ULL)))));
}
}
// tools/platform/gc/gpu_fetch.go:495:1
void gc_gpu_rasterPrimitive(gc_gpu* g,gc_Machine* m,uint32_t prim,Slice<gc_clipVertex> v){
{
Slice<gc_clipVertex> buf = Slice<gc_clipVertex>::make(cast<int64_t>(0ULL),cast<int64_t>(8ULL));
Slice<gc_clipVertex> scratch = Slice<gc_clipVertex>::make(cast<int64_t>(0ULL),cast<int64_t>(8ULL));
gc_tevState tev = gc_gpu_tevstate(g);
g->tris = sub(g->tris,0,cast<int64_t>(0ULL));
auto draw = [&](gc_clipVertex a,gc_clipVertex b,gc_clipVertex c)->void{
auto tmp99 = gc_gpu_clipAndSetup(g,m,buf,scratch,a,b,c);
buf = std::get<0>(tmp99);
scratch = std::get<1>(tmp99);
}
;
{
switch(prim){
case cast<uint32_t>(128ULL):case cast<uint32_t>(136ULL):{
{int64_t i = cast<int64_t>(0ULL);for (;(cast<int64_t>((i + cast<int64_t>(4ULL))) <= len(v));i += cast<int64_t>(4ULL)){
draw(v[i],v[cast<int64_t>((i + cast<int64_t>(1ULL)))],v[cast<int64_t>((i + cast<int64_t>(2ULL)))]);
draw(v[i],v[cast<int64_t>((i + cast<int64_t>(2ULL)))],v[cast<int64_t>((i + cast<int64_t>(3ULL)))]);
}
}break;}
case cast<uint32_t>(144ULL):{
{int64_t i = cast<int64_t>(0ULL);for (;(cast<int64_t>((i + cast<int64_t>(2ULL))) < len(v));i += cast<int64_t>(3ULL)){
draw(v[i],v[cast<int64_t>((i + cast<int64_t>(1ULL)))],v[cast<int64_t>((i + cast<int64_t>(2ULL)))]);
}
}break;}
case cast<uint32_t>(152ULL):{
{int64_t i = cast<int64_t>(0ULL);for (;(cast<int64_t>((i + cast<int64_t>(2ULL))) < len(v));i++){
if ((cast<int64_t>((i & cast<int64_t>(1ULL))) == cast<int64_t>(0ULL))) {
draw(v[i],v[cast<int64_t>((i + cast<int64_t>(1ULL)))],v[cast<int64_t>((i + cast<int64_t>(2ULL)))]);
}
else {
draw(v[cast<int64_t>((i + cast<int64_t>(1ULL)))],v[i],v[cast<int64_t>((i + cast<int64_t>(2ULL)))]);
}
}
}break;}
case cast<uint32_t>(160ULL):{
{int64_t i = cast<int64_t>(1ULL);for (;(cast<int64_t>((i + cast<int64_t>(1ULL))) < len(v));i++){
draw(v[cast<int64_t>(0ULL)],v[i],v[cast<int64_t>((i + cast<int64_t>(1ULL)))]);
}
}break;}
}}
gc_gpu_fill(g,m,(&tev),g->tris);
}
}
// tools/platform/gc/gpu_light.go:33:1
std::array<float,4> gc_gpu_xfColor(gc_gpu* g,int64_t addr){
{
uint32_t w = cast<uint32_t>(0ULL);
if (((addr >= cast<int64_t>(0ULL)) && (addr < cast<int64_t>(4192ULL)))) {
w = g->XFMem[addr];
}
return std::array<float,4>{(cast<float>(shr<uint32_t>(w,cast<int64_t>(24ULL))) / cast<float>(255ULL)),(cast<float>(cast<uint32_t>((shr<uint32_t>(w,cast<int64_t>(16ULL)) & cast<uint32_t>(255ULL)))) / cast<float>(255ULL)),(cast<float>(cast<uint32_t>((shr<uint32_t>(w,cast<int64_t>(8ULL)) & cast<uint32_t>(255ULL)))) / cast<float>(255ULL)),(cast<float>(cast<uint32_t>((w & cast<uint32_t>(255ULL)))) / cast<float>(255ULL))};
}
}
// tools/platform/gc/gpu_light.go:48:1
uint32_t gc_lightMask(uint32_t ctl){
{
return cast<uint32_t>((cast<uint32_t>(((shr<uint32_t>(ctl,cast<int64_t>(2ULL))) & cast<uint32_t>(15ULL))) | cast<uint32_t>(((shr<uint32_t>(ctl,cast<int64_t>(7ULL))) & cast<uint32_t>(240ULL)))));
}
}
// tools/platform/gc/gpu_light.go:66:1
std::array<std::array<uint8_t,4>,2> gc_gpu_rasChannel(gc_gpu* g,gc_Machine* m,uint8_t vr,uint8_t vg,uint8_t vb,uint8_t va,float ex,float ey,float ez,float nx,float ny,float nz){
{
std::array<std::array<uint8_t,4>,2> out={};
int64_t nchan = cast<int64_t>(g->XFMem[cast<int64_t>(4105ULL)]);
std::array<float,4> vtx = std::array<float,4>{(cast<float>(vr) / cast<float>(255ULL)),(cast<float>(vg) / cast<float>(255ULL)),(cast<float>(vb) / cast<float>(255ULL)),(cast<float>(va) / cast<float>(255ULL))};
{int64_t c = cast<int64_t>(0ULL);for (;((c < cast<int64_t>(2ULL)) && (c < nchan));c++){
std::array<float,4> amb = gc_gpu_xfColor(g,cast<int64_t>((cast<int64_t>(4106ULL) + c)));
std::array<float,4> mat = gc_gpu_xfColor(g,cast<int64_t>((cast<int64_t>(4108ULL) + c)));
std::array<float,4> rgb = gc_gpu_evalChannel(g,m,g->XFMem[cast<uint32_t>((cast<uint32_t>(4110ULL) + cast<uint32_t>(c)))],false,vtx,amb,mat,ex,ey,ez,nx,ny,nz);
std::array<float,4> alpha = gc_gpu_evalChannel(g,m,g->XFMem[cast<uint32_t>((cast<uint32_t>(4112ULL) + cast<uint32_t>(c)))],true,vtx,amb,mat,ex,ey,ez,nx,ny,nz);
out[c] = std::array<uint8_t,4>{gc_chanU8(rgb[cast<int64_t>(0ULL)]),gc_chanU8(rgb[cast<int64_t>(1ULL)]),gc_chanU8(rgb[cast<int64_t>(2ULL)]),gc_chanU8(alpha[cast<int64_t>(3ULL)])};
}
}return out;
}
}
// tools/platform/gc/gpu_light.go:81:1
uint8_t gc_chanU8(float v){
{
if ((v <= cast<float>(0ULL))) {
return cast<uint8_t>(0ULL);
}
if ((v >= cast<float>(1ULL))) {
return cast<uint8_t>(255ULL);
}
return cast<uint8_t>(((v * cast<float>(255ULL)) + 0.5));
}
}
// tools/platform/gc/gpu_light.go:94:1
std::array<float,4> gc_gpu_evalChannel(gc_gpu* g,gc_Machine* m,uint32_t ctl,bool isAlpha,std::array<float,4> vtx,std::array<float,4> amb,std::array<float,4> mat,float ex,float ey,float ez,float nx,float ny,float nz){
{
if ((cast<uint32_t>((ctl & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL))) {
mat = vtx;
}
if ((cast<uint32_t>((ctl & cast<uint32_t>(2ULL))) == cast<uint32_t>(0ULL))) {
return mat;
}
if ((cast<uint32_t>((ctl & cast<uint32_t>(64ULL))) != cast<uint32_t>(0ULL))) {
amb = vtx;
}
std::array<float,4> illum = amb;
uint32_t mask = gc_lightMask(ctl);
uint32_t diffFn = cast<uint32_t>(((shr<uint32_t>(ctl,cast<int64_t>(7ULL))) & cast<uint32_t>(3ULL)));
bool attnNotNone = (cast<uint32_t>((ctl & cast<uint32_t>(512ULL))) != cast<uint32_t>(0ULL));
bool attnNotSpec = (cast<uint32_t>((ctl & cast<uint32_t>(1024ULL))) != cast<uint32_t>(0ULL));
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(8ULL));i++){
if ((cast<uint32_t>((mask & (shl<uint32_t>(cast<uint32_t>(1ULL),i)))) == cast<uint32_t>(0ULL))) {
continue;
}
int64_t base = cast<int64_t>((cast<int64_t>(1536ULL) + cast<int64_t>((i * cast<int64_t>(16ULL)))));
std::array<float,4> lcol = gc_gpu_xfColor(g,cast<int64_t>((base + cast<int64_t>(3ULL))));
float lpx = gc_gpu_xfFloat(g,cast<int64_t>((base + cast<int64_t>(10ULL))));
float lpy = gc_gpu_xfFloat(g,cast<int64_t>((base + cast<int64_t>(11ULL))));
float lpz = gc_gpu_xfFloat(g,cast<int64_t>((base + cast<int64_t>(12ULL))));
float ldx = gc_gpu_xfFloat(g,cast<int64_t>((base + cast<int64_t>(13ULL))));
float ldy = gc_gpu_xfFloat(g,cast<int64_t>((base + cast<int64_t>(14ULL))));
float ldz = gc_gpu_xfFloat(g,cast<int64_t>((base + cast<int64_t>(15ULL))));
auto tmp100 = std::make_tuple(gc_gpu_xfFloat(g,cast<int64_t>((base + cast<int64_t>(4ULL)))),gc_gpu_xfFloat(g,cast<int64_t>((base + cast<int64_t>(5ULL)))),gc_gpu_xfFloat(g,cast<int64_t>((base + cast<int64_t>(6ULL)))));
float a0 = std::get<0>(tmp100);
float a1 = std::get<1>(tmp100);
float a2 = std::get<2>(tmp100);
auto tmp101 = std::make_tuple(gc_gpu_xfFloat(g,cast<int64_t>((base + cast<int64_t>(7ULL)))),gc_gpu_xfFloat(g,cast<int64_t>((base + cast<int64_t>(8ULL)))),gc_gpu_xfFloat(g,cast<int64_t>((base + cast<int64_t>(9ULL)))));
float k0 = std::get<0>(tmp101);
float k1 = std::get<1>(tmp101);
float k2 = std::get<2>(tmp101);
auto tmp102 = std::make_tuple((lpx - ex),(lpy - ey),(lpz - ez));
float dx = std::get<0>(tmp102);
float dy = std::get<1>(tmp102);
float dz = std::get<2>(tmp102);
float dist = cast<float>(go_math_Sqrt(cast<double>((((dx * dx) + (dy * dy)) + (dz * dz)))));
if ((dist > cast<float>(0ULL))) {
auto tmp103 = std::make_tuple((dx / dist),(dy / dist),(dz / dist));
dx = std::get<0>(tmp103);
dy = std::get<1>(tmp103);
dz = std::get<2>(tmp103);
}
float attn={};
if ((!attnNotNone)) {
attn = cast<float>(1ULL);
}
else if (attnNotSpec) {
float cs = (((dx * ldx) + (dy * ldy)) + (dz * ldz));
if ((cs < cast<float>(0ULL))) {
cs = cast<float>(0ULL);
}
float num = ((a0 + (a1 * cs)) + ((a2 * cs) * cs));
if ((num < cast<float>(0ULL))) {
num = cast<float>(0ULL);
}
float den = ((k0 + (k1 * dist)) + ((k2 * dist) * dist));
if ((den != cast<float>(0ULL))) {
attn = (num / den);
}
}
else {
auto tmp104 = gc_normalize3f(lpx,lpy,lpz);
dx = std::get<0>(tmp104);
dy = std::get<1>(tmp104);
dz = std::get<2>(tmp104);
float cs = cast<float>(cast<float>(0ULL));
if (((((nx * dx) + (ny * dy)) + (nz * dz)) >= cast<float>(0ULL))) {
cs = (((nx * ldx) + (ny * ldy)) + (nz * ldz));
if ((cs < cast<float>(0ULL))) {
cs = cast<float>(0ULL);
}
}
float num = ((a0 + (a1 * cs)) + ((a2 * cs) * cs));
if ((num < cast<float>(0ULL))) {
num = cast<float>(0ULL);
}
auto tmp105 = std::make_tuple(k0,k1,k2);
float kx = std::get<0>(tmp105);
float ky = std::get<1>(tmp105);
float kz = std::get<2>(tmp105);
if ((diffFn != cast<uint32_t>(0ULL))) {
auto tmp106 = gc_normalize3f(k0,k1,k2);
kx = std::get<0>(tmp106);
ky = std::get<1>(tmp106);
kz = std::get<2>(tmp106);
}
float den = ((kx + (ky * cs)) + ((kz * cs) * cs));
if ((den != cast<float>(0ULL))) {
attn = (num / den);
}
}
float diff = cast<float>(cast<float>(1ULL));
if ((diffFn != cast<uint32_t>(0ULL))) {
diff = (((nx * dx) + (ny * dy)) + (nz * dz));
if (((diffFn == cast<uint32_t>(2ULL)) && (diff < cast<float>(0ULL)))) {
diff = cast<float>(0ULL);
}
}
if (isAlpha) {
illum[cast<int64_t>(3ULL)] += ((attn * diff) * lcol[cast<int64_t>(3ULL)]);
}
else {
illum[cast<int64_t>(0ULL)] += ((attn * diff) * lcol[cast<int64_t>(0ULL)]);
illum[cast<int64_t>(1ULL)] += ((attn * diff) * lcol[cast<int64_t>(1ULL)]);
illum[cast<int64_t>(2ULL)] += ((attn * diff) * lcol[cast<int64_t>(2ULL)]);
}
}
}{auto&& tmp107 = illum;
for(int64_t tmp108=0;tmp108<len(tmp107);++tmp108){
auto k=tmp108;if ((illum[k] < cast<float>(0ULL))) {
illum[k] = cast<float>(0ULL);
}
if ((illum[k] > cast<float>(1ULL))) {
illum[k] = cast<float>(1ULL);
}
}}
return std::array<float,4>{(mat[cast<int64_t>(0ULL)] * illum[cast<int64_t>(0ULL)]),(mat[cast<int64_t>(1ULL)] * illum[cast<int64_t>(1ULL)]),(mat[cast<int64_t>(2ULL)] * illum[cast<int64_t>(2ULL)]),(mat[cast<int64_t>(3ULL)] * illum[cast<int64_t>(3ULL)])};
}
}
// tools/platform/gc/gpu_light.go:232:1
std::tuple<float,float,float> gc_normalize3f(float x,float y,float z){
{
float l = cast<float>(go_math_Sqrt(cast<double>((((x * x) + (y * y)) + (z * z)))));
if ((l == cast<float>(0ULL))) {
return {x,y,z};
}
return {(x / l),(y / l),(z / l)};
}
}
// tools/platform/gc/gpu_light.go:242:1
std::tuple<float,float,float> gc_gpu_normalToEye(gc_gpu* g,int64_t mtxIdx,float nx,float ny,float nz){
float ox{};
float oy{};
float oz{};
{
int64_t base = cast<int64_t>((cast<int64_t>(1024ULL) + cast<int64_t>((mtxIdx * cast<int64_t>(3ULL)))));
ox = (((gc_gpu_xfFloat(g,cast<int64_t>((base + cast<int64_t>(0ULL)))) * nx) + (gc_gpu_xfFloat(g,cast<int64_t>((base + cast<int64_t>(1ULL)))) * ny)) + (gc_gpu_xfFloat(g,cast<int64_t>((base + cast<int64_t>(2ULL)))) * nz));
oy = (((gc_gpu_xfFloat(g,cast<int64_t>((base + cast<int64_t>(3ULL)))) * nx) + (gc_gpu_xfFloat(g,cast<int64_t>((base + cast<int64_t>(4ULL)))) * ny)) + (gc_gpu_xfFloat(g,cast<int64_t>((base + cast<int64_t>(5ULL)))) * nz));
oz = (((gc_gpu_xfFloat(g,cast<int64_t>((base + cast<int64_t>(6ULL)))) * nx) + (gc_gpu_xfFloat(g,cast<int64_t>((base + cast<int64_t>(7ULL)))) * ny)) + (gc_gpu_xfFloat(g,cast<int64_t>((base + cast<int64_t>(8ULL)))) * nz));
float l = cast<float>(go_math_Sqrt(cast<double>((((ox * ox) + (oy * oy)) + (oz * oz)))));
if ((l > cast<float>(0ULL))) {
auto tmp109 = std::make_tuple((ox / l),(oy / l),(oz / l));
ox = std::get<0>(tmp109);
oy = std::get<1>(tmp109);
oz = std::get<2>(tmp109);
}
return {ox,oy,oz};
}
}
// tools/platform/gc/gpu_raster.go:41:1
void gc_gpu_ensureRaster(gc_gpu* g){
{
gc_gpu_ensureEFB(g);
if ((!g->ZBuf)) {
g->ZBuf = Slice<uint32_t>::make(cast<int64_t>(337920ULL));
{auto&& tmp111 = g->ZBuf;
for(int64_t tmp112=0;tmp112<len(tmp111);++tmp112){
auto i=tmp112;g->ZBuf[i] = cast<uint32_t>(16777215ULL);
}}
}
}
}
// tools/platform/gc/gpu_raster.go:72:1
void gc_perspTexCoords(float b0,float b1,float b2,const gc_screenVertex& v0,const gc_screenVertex& v1,const gc_screenVertex& v2,std::array<gc_texCoord,8>* out){
{
float iw = (((b0 * v0.invW) + (b1 * v1.invW)) + (b2 * v2.invW));
{int64_t i = cast<int64_t>(0ULL);for (;(i < v0.ntc);i++){
float s = ((((((b0 * v0.tc[i].s) * v0.invW) + ((b1 * v1.tc[i].s) * v1.invW)) + ((b2 * v2.tc[i].s) * v2.invW))) / iw);
float t = ((((((b0 * v0.tc[i].t) * v0.invW) + ((b1 * v1.tc[i].t) * v1.invW)) + ((b2 * v2.tc[i].t) * v2.invW))) / iw);
float q = ((((((b0 * v0.tc[i].q) * v0.invW) + ((b1 * v1.tc[i].q) * v1.invW)) + ((b2 * v2.tc[i].q) * v2.invW))) / iw);
if ((q != cast<float>(0ULL))) {
s /= q;
t /= q;
}
(*out)[i] = gc_texCoord{cast<float>(s),cast<float>(t),cast<float>(q)};
}
}}
}
// tools/platform/gc/gpu_raster.go:120:1
void gc_gpu_init(gc_gpu* g){
{
constexpr int64_t bias=342ULL;
g->BP[cast<int64_t>(32ULL)] = cast<uint32_t>(1401174ULL);
g->BP[cast<int64_t>(33ULL)] = cast<uint32_t>(4019045ULL);
g->BP[cast<int64_t>(89ULL)] = cast<uint32_t>(175275ULL);
}
}
// tools/platform/gc/gpu_raster.go:144:1
std::tuple<int64_t,int64_t,int64_t,int64_t> gc_gpu_scissorBox(gc_gpu* g){
int64_t x0{};
int64_t y0{};
int64_t x1{};
int64_t y1{};
{
int64_t offX = cast<int64_t>((cast<int64_t>((cast<int64_t>(cast<uint32_t>((g->BP[cast<int64_t>(89ULL)] & cast<uint32_t>(1023ULL)))) * cast<int64_t>(2ULL))) - cast<int64_t>(342ULL)));
int64_t offY = cast<int64_t>((cast<int64_t>((cast<int64_t>(cast<uint32_t>(((shr<uint32_t>(g->BP[cast<int64_t>(89ULL)],cast<int64_t>(10ULL))) & cast<uint32_t>(1023ULL)))) * cast<int64_t>(2ULL))) - cast<int64_t>(342ULL)));
x0 = cast<int64_t>((cast<int64_t>((cast<int64_t>(cast<uint32_t>(((shr<uint32_t>(g->BP[cast<int64_t>(32ULL)],cast<int64_t>(12ULL))) & cast<uint32_t>(2047ULL)))) - cast<int64_t>(342ULL))) - offX));
y0 = cast<int64_t>((cast<int64_t>((cast<int64_t>(cast<uint32_t>((g->BP[cast<int64_t>(32ULL)] & cast<uint32_t>(2047ULL)))) - cast<int64_t>(342ULL))) - offY));
x1 = cast<int64_t>((cast<int64_t>((cast<int64_t>((cast<int64_t>(cast<uint32_t>(((shr<uint32_t>(g->BP[cast<int64_t>(33ULL)],cast<int64_t>(12ULL))) & cast<uint32_t>(2047ULL)))) - cast<int64_t>(342ULL))) - offX)) + cast<int64_t>(1ULL)));
y1 = cast<int64_t>((cast<int64_t>((cast<int64_t>((cast<int64_t>(cast<uint32_t>((g->BP[cast<int64_t>(33ULL)] & cast<uint32_t>(2047ULL)))) - cast<int64_t>(342ULL))) - offY)) + cast<int64_t>(1ULL)));
if ((x0 < cast<int64_t>(0ULL))) {
x0 = cast<int64_t>(0ULL);
}
if ((y0 < cast<int64_t>(0ULL))) {
y0 = cast<int64_t>(0ULL);
}
if ((x1 > cast<int64_t>(640ULL))) {
x1 = cast<int64_t>(640ULL);
}
if ((y1 > cast<int64_t>(528ULL))) {
y1 = cast<int64_t>(528ULL);
}
return {x0,y0,x1,y1};
}
}
// tools/platform/gc/gpu_raster.go:173:1
std::tuple<gc_rasterTri,bool> gc_gpu_setupTri(gc_gpu* g,gc_screenVertex v0,gc_screenVertex v1,gc_screenVertex v2){
{
int64_t minX = cast<int64_t>(gc_min3(v0.x,v1.x,v2.x));
int64_t maxX = cast<int64_t>((cast<int64_t>(gc_max3(v0.x,v1.x,v2.x)) + cast<int64_t>(1ULL)));
int64_t minY = cast<int64_t>(gc_min3(v0.y,v1.y,v2.y));
int64_t maxY = cast<int64_t>((cast<int64_t>(gc_max3(v0.y,v1.y,v2.y)) + cast<int64_t>(1ULL)));
auto tmp113 = gc_gpu_scissorBox(g);
int64_t sx0 = std::get<0>(tmp113);
int64_t sy0 = std::get<1>(tmp113);
int64_t sx1 = std::get<2>(tmp113);
int64_t sy1 = std::get<3>(tmp113);
if ((minX < sx0)) {
minX = sx0;
}
if ((minY < sy0)) {
minY = sy0;
}
if ((maxX > sx1)) {
maxX = sx1;
}
if ((maxY > sy1)) {
maxY = sy1;
}
if (((minX >= maxX) || (minY >= maxY))) {
return {gc_rasterTri{},false};
}
float area = gc_edge(v0.x,v0.y,v1.x,v1.y,v2.x,v2.y);
if ((area == cast<float>(0ULL))) {
return {gc_rasterTri{},false};
}
if (gc_gpu_cullTest(g,area)) {
g->profCulled++;
return {gc_rasterTri{},false};
}
uint32_t zm = g->BP[cast<int64_t>(64ULL)];
bool zEnable = (cast<uint32_t>((zm & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL));
return {gc_rasterTri{v0,v1,v2,cast<float>(area),minX,maxX,minY,maxY,zEnable,(zEnable && (cast<uint32_t>((zm & cast<uint32_t>(16ULL))) != cast<uint32_t>(0ULL))),cast<int64_t>(cast<uint32_t>(((shr<uint32_t>(zm,cast<int64_t>(1ULL))) & cast<uint32_t>(7ULL))))},true};
}
}
// tools/platform/gc/gpu_raster.go:234:1
void gc_gpu_fillTri(gc_gpu* g,gc_Machine* m,gc_tevState* tev,gc_rasterTri* t,int64_t yLo,int64_t yHi,gc_rstats* st){
{
auto tmp114 = std::make_tuple(t->v0,t->v1,t->v2);
gc_screenVertex v0 = std::get<0>(tmp114);
gc_screenVertex v1 = std::get<1>(tmp114);
gc_screenVertex v2 = std::get<2>(tmp114);
float area = t->area;
auto tmp115 = std::make_tuple(t->zEnable,t->zFunc,t->zWrite);
bool zEnable = std::get<0>(tmp115);
int64_t zFunc = std::get<1>(tmp115);
bool zWrite = std::get<2>(tmp115);
auto tmp116 = std::make_tuple(t->minX,t->maxX);
int64_t minX = std::get<0>(tmp116);
int64_t maxX = std::get<1>(tmp116);
{int64_t y = yLo;for (;(y < yHi);y++){
float py = (cast<float>(y) + 0.5);
{int64_t x = minX;for (;(x < maxX);x++){
float px = (cast<float>(x) + 0.5);
float w0 = gc_edge(v1.x,v1.y,v2.x,v2.y,px,py);
float w1 = gc_edge(v2.x,v2.y,v0.x,v0.y,px,py);
float w2 = gc_edge(v0.x,v0.y,v1.x,v1.y,px,py);
if ((((((w0 < cast<float>(0ULL)) || (w1 < cast<float>(0ULL))) || (w2 < cast<float>(0ULL)))) && ((((w0 > cast<float>(0ULL)) || (w1 > cast<float>(0ULL))) || (w2 > cast<float>(0ULL)))))) {
continue;
}
float b0 = (w0 / area);
float b1 = (w1 / area);
float b2 = (w2 / area);
uint32_t z = cast<uint32_t>((((b0 * v0.z) + (b1 * v1.z)) + (b2 * v2.z)));
int64_t idx = cast<int64_t>((cast<int64_t>((y * cast<int64_t>(640ULL))) + x));
if ((zEnable && (!gc_depthCompare(z,g->ZBuf[idx],zFunc)))) {
st->zRej++;
if (bool(m->OnPixel)) {
m->OnPixel(x,y,gc_PixelEvent{});
}
continue;
}
std::array<std::array<uint8_t,4>,2> ras={};
{int64_t c = cast<int64_t>(0ULL);for (;(c < cast<int64_t>(2ULL));c++){
{int64_t k = cast<int64_t>(0ULL);for (;(k < cast<int64_t>(4ULL));k++){
ras[c][k] = cast<uint8_t>((((b0 * cast<float>(v0.col[c][k])) + (b1 * cast<float>(v1.col[c][k]))) + (b2 * cast<float>(v2.col[c][k]))));
}
}}
}std::array<gc_texCoord,8> tc={};
gc_perspTexCoords(b0,b1,b2,v0,v1,v2,(&tc));
auto tmp117 = gc_gpu_shade(g,m,tev,(&ras),(&tc));
uint8_t fr = std::get<0>(tmp117);
uint8_t fg = std::get<1>(tmp117);
uint8_t fb = std::get<2>(tmp117);
uint8_t fa = std::get<3>(tmp117);
bool pass = std::get<4>(tmp117);
if (((x == gc_pixDbgX) && (y == gc_pixDbgY))) {
gc_texState t0 = gc_gpu_texSetup(g,cast<int64_t>(0ULL));
auto tmp118 = std::make_tuple(tc[cast<int64_t>(0ULL)].s,tc[cast<int64_t>(0ULL)].t);
float u = std::get<0>(tmp118);
float v = std::get<1>(tmp118);
auto tmp119 = gc_gpu_sampleTexmap(g,m,(&t0),u,v);
uint8_t tr = std::get<0>(tmp119);
uint8_t tg = std::get<1>(tmp119);
uint8_t tb = std::get<2>(tmp119);
uint8_t ta = std::get<3>(tmp119);
go_fmt_Fprintf(go_os_Stderr,std::string("PIXDBG (%d,%d): ras0 %d,%d,%d,%d ras1 %d,%d,%d,%d uv (%.4f,%.4f) tex0 0x%06X fmt%X %dx%d texel (%d,%d)=%d,%d,%d,%d -> out %d,%d,%d,%d pass=%v dst %08X\012",151),x,y,ras[cast<int64_t>(0ULL)][cast<int64_t>(0ULL)],ras[cast<int64_t>(0ULL)][cast<int64_t>(1ULL)],ras[cast<int64_t>(0ULL)][cast<int64_t>(2ULL)],ras[cast<int64_t>(0ULL)][cast<int64_t>(3ULL)],ras[cast<int64_t>(1ULL)][cast<int64_t>(0ULL)],ras[cast<int64_t>(1ULL)][cast<int64_t>(1ULL)],ras[cast<int64_t>(1ULL)][cast<int64_t>(2ULL)],ras[cast<int64_t>(1ULL)][cast<int64_t>(3ULL)],u,v,t0.base,t0.format,t0.width,t0.height,gc_wrapCoord(cast<int64_t>((u * cast<float>(t0.width))),t0.width,t0.wrapS),gc_wrapCoord(cast<int64_t>((v * cast<float>(t0.height))),t0.height,t0.wrapT),tr,tg,tb,ta,fr,fg,fb,fa,pass,g->EFB[idx]);
}
if ((!pass)) {
st->aRej++;
if (bool(m->OnPixel)) {
m->OnPixel(x,y,gc_PixelEvent{fr,fg,fb,fa,{}});
}
continue;
}
g->EFB[idx] = gc_gpu_blend(g,g->EFB[idx],fr,fg,fb,fa);
if(rrcapture::trace.active)rrEFBWrite(g,idx);
if (zWrite) {
g->ZBuf[idx] = z;
}
st->written++;
if (bool(m->OnPixel)) {
m->OnPixel(x,y,gc_PixelEvent{fr,fg,fb,fa,true});
}
}
}}
}}
}
// tools/platform/gc/gpu_raster.go:387:1
void gc_gpu_mergeStats(gc_gpu* g,gc_rstats* st){
{
g->pixWritten += st->written;
g->pixZRej += st->zRej;
g->pixARej += st->aRej;
}
}
// tools/platform/gc/gpu_raster.go:492:1
bool gc_gpu_cullTest(gc_gpu* g,float area){
{
if ((gc_cullExperiment != std::string("",0))) {
return gc_gpu_cullTestExperiment(g,area);
}
{
switch(cast<uint32_t>(((shr<uint32_t>(g->BP[cast<int64_t>(0ULL)],cast<int64_t>(14ULL))) & cast<uint32_t>(3ULL)))){
case cast<uint32_t>(3ULL):{
return true;
break;}
case cast<uint32_t>(2ULL):{
return (area > cast<float>(0ULL));
break;}
case cast<uint32_t>(1ULL):{
return (area < cast<float>(0ULL));
break;}
}}
return false;
}
}
// tools/platform/gc/gpu_raster.go:519:1
bool gc_gpu_cullTestExperiment(gc_gpu* g,float area){
{
{
auto tmp121=gc_cullExperiment;
if (tmp121==(std::string("off",3))){
return false;
}
else if (tmp121==(std::string("flip",4))){
{
switch(cast<uint32_t>(((shr<uint32_t>(g->BP[cast<int64_t>(0ULL)],cast<int64_t>(14ULL))) & cast<uint32_t>(3ULL)))){
case cast<uint32_t>(3ULL):{
return true;
break;}
case cast<uint32_t>(2ULL):{
return (area < cast<float>(0ULL));
break;}
case cast<uint32_t>(1ULL):{
return (area > cast<float>(0ULL));
break;}
}}
}
}
tmp120:;
return false;
}
}
// tools/platform/gc/gpu_raster.go:538:1
bool gc_depthCompare(uint32_t z,uint32_t buf,int64_t comp){
{
{
switch(comp){
case cast<int64_t>(0ULL):{
return false;
break;}
case cast<int64_t>(1ULL):{
return (z < buf);
break;}
case cast<int64_t>(2ULL):{
return (z == buf);
break;}
case cast<int64_t>(3ULL):{
return (z <= buf);
break;}
case cast<int64_t>(4ULL):{
return (z > buf);
break;}
case cast<int64_t>(5ULL):{
return (z != buf);
break;}
case cast<int64_t>(6ULL):{
return (z >= buf);
break;}
default:{
return true;
break;}
}}
}
}
// tools/platform/gc/gpu_raster.go:561:1
float gc_edge(float ax,float ay,float bx,float by,float cx,float cy){
{
return ((((bx - ax)) * ((cy - ay))) - (((by - ay)) * ((cx - ax))));
}
}
// tools/platform/gc/gpu_raster.go:565:1
float gc_min3(float a,float b,float c){
{
return gc_minf(gc_minf(a,b),c);
}
}
// tools/platform/gc/gpu_raster.go:566:1
float gc_max3(float a,float b,float c){
{
return gc_maxf(gc_maxf(a,b),c);
}
}
// tools/platform/gc/gpu_raster.go:568:1
float gc_minf(float a,float b){
{
if ((a < b)) {
return a;
}
return b;
}
}
// tools/platform/gc/gpu_raster.go:575:1
float gc_maxf(float a,float b){
{
if ((a > b)) {
return a;
}
return b;
}
}
// tools/platform/gc/gpu_tev.go:36:1
std::tuple<float,float,float,float> gc_tevColorReg(std::array<uint32_t,2> w){
float r{};
float g{};
float b{};
float a{};
{
r = cast<float>(gc_sext11(cast<uint32_t>((w[cast<int64_t>(0ULL)] & cast<uint32_t>(2047ULL)))));
a = cast<float>(gc_sext11(cast<uint32_t>(((shr<uint32_t>(w[cast<int64_t>(0ULL)],cast<int64_t>(12ULL))) & cast<uint32_t>(2047ULL)))));
b = cast<float>(gc_sext11(cast<uint32_t>((w[cast<int64_t>(1ULL)] & cast<uint32_t>(2047ULL)))));
g = cast<float>(gc_sext11(cast<uint32_t>(((shr<uint32_t>(w[cast<int64_t>(1ULL)],cast<int64_t>(12ULL))) & cast<uint32_t>(2047ULL)))));
return {r,g,b,a};
}
}
// tools/platform/gc/gpu_tev.go:45:1
int32_t gc_sext11(uint32_t v){
{
if ((cast<uint32_t>((v & cast<uint32_t>(1024ULL))) != cast<uint32_t>(0ULL))) {
return cast<int32_t>((cast<int32_t>(v) - cast<int32_t>(2048ULL)));
}
return cast<int32_t>(v);
}
}
// tools/platform/gc/gpu_tev.go:57:1
std::tuple<uint8_t,uint8_t,uint8_t,uint8_t,bool> gc_gpu_shade_reference(gc_gpu* g,gc_Machine* m,gc_tevState* t,std::array<std::array<uint8_t,4>,2>* rasCol,std::array<gc_texCoord,8>* tc){
uint8_t fr{};
uint8_t fg{};
uint8_t fb{};
uint8_t fa{};
bool pass{};
{
std::array<std::array<float,4>,4> reg = t->seed;
{int64_t s = cast<int64_t>(0ULL);for (;(s < t->numStages);s++){
gc_tevStage* st = (&t->stages[s]);
std::array<float,4> ras = gc_rasSelect(st->rasSel,rasCol);
std::array<float,4> sras = gc_swizzle(st->swapRas,ras);
std::array<float,4> tex={};
if (st->texEnable) {
auto tmp122 = gc_gpu_sampleTexmap(g,m,(&t->tex[st->texmap]),(*tc)[st->texcoord].s,(*tc)[st->texcoord].t);
uint8_t tr = std::get<0>(tmp122);
uint8_t tg = std::get<1>(tmp122);
uint8_t tb = std::get<2>(tmp122);
uint8_t ta = std::get<3>(tmp122);
tex = gc_swizzle(st->swapTex,std::array<float,4>{cast<float>(tr),cast<float>(tg),cast<float>(tb),cast<float>(ta)});
}
auto tmp123 = gc_combineColor(st->cc,(&reg),tex,sras,st->konstC);
std::array<float,3> crgb = std::get<0>(tmp123);
std::array<float,3> ca = std::get<1>(tmp123);
std::array<float,3> cb = std::get<2>(tmp123);
float aout = gc_combineAlpha(st->ac,(&reg),tex,sras,st->konstA,ca,cb);
auto tmp124 = std::make_tuple(crgb[cast<int64_t>(0ULL)],crgb[cast<int64_t>(1ULL)],crgb[cast<int64_t>(2ULL)]);
reg[st->cdest][cast<int64_t>(0ULL)] = std::get<0>(tmp124);
reg[st->cdest][cast<int64_t>(1ULL)] = std::get<1>(tmp124);
reg[st->cdest][cast<int64_t>(2ULL)] = std::get<2>(tmp124);
reg[st->adest][cast<int64_t>(3ULL)] = aout;
}
}std::array<float,4> out = reg[cast<int64_t>(0ULL)];
uint8_t a8 = gc_toU8(out[cast<int64_t>(3ULL)]);
if ((!gc_gpu_alphaTest(g,a8))) {
return {cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),false};
}
return {gc_toU8(out[cast<int64_t>(0ULL)]),gc_toU8(out[cast<int64_t>(1ULL)]),gc_toU8(out[cast<int64_t>(2ULL)]),a8,true};
}
}
// tools/platform/gc/gpu_tev.go:108:1
uint8_t gc_toU8(float f){
{
return cast<uint8_t>((gc_clampf(f) + 0.5));
}
}
// tools/platform/gc/gpu_tev.go:123:1
std::array<float,4> gc_rasSelect(uint32_t sel,std::array<std::array<uint8_t,4>,2>* col){
{
{
switch(sel){
case cast<uint32_t>(0ULL):case cast<uint32_t>(1ULL):{
std::array<uint8_t,4> c = (*col)[sel];
return std::array<float,4>{cast<float>(c[cast<int64_t>(0ULL)]),cast<float>(c[cast<int64_t>(1ULL)]),cast<float>(c[cast<int64_t>(2ULL)]),cast<float>(c[cast<int64_t>(3ULL)])};
break;}
case cast<uint32_t>(7ULL):{
return std::array<float,4>{};
break;}
}}
return std::array<float,4>{};
}
}
// tools/platform/gc/gpu_tev.go:146:1
std::tuple<std::array<float,3>,std::array<float,3>,std::array<float,3>> gc_combineColor(uint32_t cc,std::array<std::array<float,4>,4>* reg,std::array<float,4> tex,std::array<float,4> ras,std::array<float,3> konst){
std::array<float,3> out{};
std::array<float,3> ca{};
std::array<float,3> cb{};
{
std::array<float,3> d = gc_colorArg(cast<int64_t>(cast<uint32_t>((cc & cast<uint32_t>(15ULL)))),reg,tex,ras,konst);
std::array<float,3> c = gc_colorArg(cast<int64_t>(cast<uint32_t>(((shr<uint32_t>(cc,cast<int64_t>(4ULL))) & cast<uint32_t>(15ULL)))),reg,tex,ras,konst);
std::array<float,3> b = gc_colorArg(cast<int64_t>(cast<uint32_t>(((shr<uint32_t>(cc,cast<int64_t>(8ULL))) & cast<uint32_t>(15ULL)))),reg,tex,ras,konst);
std::array<float,3> a = gc_colorArg(cast<int64_t>(cast<uint32_t>(((shr<uint32_t>(cc,cast<int64_t>(12ULL))) & cast<uint32_t>(15ULL)))),reg,tex,ras,konst);
int64_t bias = cast<int64_t>(cast<uint32_t>(((shr<uint32_t>(cc,cast<int64_t>(16ULL))) & cast<uint32_t>(3ULL))));
bool op = (cast<uint32_t>(((shr<uint32_t>(cc,cast<int64_t>(18ULL))) & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL));
bool clamp = (cast<uint32_t>(((shr<uint32_t>(cc,cast<int64_t>(19ULL))) & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL));
int64_t scale = cast<int64_t>(cast<uint32_t>(((shr<uint32_t>(cc,cast<int64_t>(20ULL))) & cast<uint32_t>(3ULL))));
if ((bias == cast<int64_t>(3ULL))) {
int64_t sel = cast<int64_t>((shl<int64_t>(scale,cast<int64_t>(1ULL)) | gc_b2i(op)));
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(3ULL));i++){
int64_t ch = i;
if ((shr<int64_t>(sel,cast<int64_t>(1ULL)) != cast<int64_t>(3ULL))) {
ch = cast<int64_t>(-1ULL);
}
if (gc_tevCompare(sel,a,b,cast<float>(0ULL),cast<float>(0ULL),ch)) {
out[i] = (d[i] + c[i]);
}
else {
out[i] = d[i];
}
if (clamp) {
out[i] = gc_clampf(out[i]);
}
}
}return {out,a,b};
}
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(3ULL));i++){
out[i] = gc_tevFormula(a[i],b[i],c[i],d[i],bias,op,scale,clamp);
}
}return {out,a,b};
}
}
// tools/platform/gc/gpu_tev.go:193:1
float gc_combineAlpha(uint32_t ac,std::array<std::array<float,4>,4>* reg,std::array<float,4> tex,std::array<float,4> ras,float konst,std::array<float,3> ca,std::array<float,3> cb){
{
float d = gc_alphaArg(cast<int64_t>(cast<uint32_t>(((shr<uint32_t>(ac,cast<int64_t>(4ULL))) & cast<uint32_t>(7ULL)))),reg,tex,ras,konst);
float c = gc_alphaArg(cast<int64_t>(cast<uint32_t>(((shr<uint32_t>(ac,cast<int64_t>(7ULL))) & cast<uint32_t>(7ULL)))),reg,tex,ras,konst);
float b = gc_alphaArg(cast<int64_t>(cast<uint32_t>(((shr<uint32_t>(ac,cast<int64_t>(10ULL))) & cast<uint32_t>(7ULL)))),reg,tex,ras,konst);
float a = gc_alphaArg(cast<int64_t>(cast<uint32_t>(((shr<uint32_t>(ac,cast<int64_t>(13ULL))) & cast<uint32_t>(7ULL)))),reg,tex,ras,konst);
int64_t bias = cast<int64_t>(cast<uint32_t>(((shr<uint32_t>(ac,cast<int64_t>(16ULL))) & cast<uint32_t>(3ULL))));
bool op = (cast<uint32_t>(((shr<uint32_t>(ac,cast<int64_t>(18ULL))) & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL));
bool clamp = (cast<uint32_t>(((shr<uint32_t>(ac,cast<int64_t>(19ULL))) & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL));
int64_t scale = cast<int64_t>(cast<uint32_t>(((shr<uint32_t>(ac,cast<int64_t>(20ULL))) & cast<uint32_t>(3ULL))));
if ((bias == cast<int64_t>(3ULL))) {
int64_t sel = cast<int64_t>((shl<int64_t>(scale,cast<int64_t>(1ULL)) | gc_b2i(op)));
float out = d;
if (gc_tevCompare(sel,ca,cb,a,b,cast<int64_t>(-1ULL))) {
out += c;
}
if (clamp) {
out = gc_clampf(out);
}
return out;
}
return gc_tevFormula(a,b,c,d,bias,op,scale,clamp);
}
}
// tools/platform/gc/gpu_tev.go:220:1
int64_t gc_b2i(bool b){
{
if (b) {
return cast<int64_t>(1ULL);
}
return cast<int64_t>(0ULL);
}
}
// tools/platform/gc/gpu_tev.go:240:1
bool gc_tevCompare(int64_t sel,std::array<float,3> ca,std::array<float,3> cb,float aa,float ab,int64_t ch){
{
auto u8 = [&](float f)->uint32_t{
return cast<uint32_t>((gc_clampf(f) + 0.5));
}
;
auto pack = [&](std::array<float,3> v,int64_t n)->uint32_t{
uint32_t r={};
{int64_t i = cast<int64_t>((n - cast<int64_t>(1ULL)));for (;(i >= cast<int64_t>(0ULL));i--){
r = cast<uint32_t>((shl<uint32_t>(r,cast<int64_t>(8ULL)) | u8(v[i])));
}
}return r;
}
;
uint32_t x={};
uint32_t y={};
{
switch(shr<int64_t>(sel,cast<int64_t>(1ULL))){
case cast<int64_t>(0ULL):{
auto tmp125 = std::make_tuple(pack(ca,cast<int64_t>(1ULL)),pack(cb,cast<int64_t>(1ULL)));
x = std::get<0>(tmp125);
y = std::get<1>(tmp125);
break;}
case cast<int64_t>(1ULL):{
auto tmp126 = std::make_tuple(pack(ca,cast<int64_t>(2ULL)),pack(cb,cast<int64_t>(2ULL)));
x = std::get<0>(tmp126);
y = std::get<1>(tmp126);
break;}
case cast<int64_t>(2ULL):{
auto tmp127 = std::make_tuple(pack(ca,cast<int64_t>(3ULL)),pack(cb,cast<int64_t>(3ULL)));
x = std::get<0>(tmp127);
y = std::get<1>(tmp127);
break;}
default:{
if ((ch < cast<int64_t>(0ULL))) {
auto tmp128 = std::make_tuple(u8(aa),u8(ab));
x = std::get<0>(tmp128);
y = std::get<1>(tmp128);
}
else {
auto tmp129 = std::make_tuple(u8(ca[ch]),u8(cb[ch]));
x = std::get<0>(tmp129);
y = std::get<1>(tmp129);
}
break;}
}}
if ((cast<int64_t>((sel & cast<int64_t>(1ULL))) == cast<int64_t>(0ULL))) {
return (x > y);
}
return (x == y);
}
}
// tools/platform/gc/gpu_tev.go:272:1
float gc_tevFormula(float a,float b,float c,float d,int64_t bias,bool sub_,int64_t scale,bool clamp){
{
float cc = (c / cast<float>(255ULL));
float lerp = ((a * ((cast<float>(1ULL) - cc))) + (b * cc));
float r={};
if (sub_) {
r = (d - lerp);
}
else {
r = (d + lerp);
}
{
switch(bias){
case cast<int64_t>(1ULL):{
r += 127.5;
break;}
case cast<int64_t>(2ULL):{
r -= 127.5;
break;}
}}
{
switch(scale){
case cast<int64_t>(1ULL):{
r *= cast<float>(2ULL);
break;}
case cast<int64_t>(2ULL):{
r *= cast<float>(4ULL);
break;}
case cast<int64_t>(3ULL):{
r *= 0.5;
break;}
}}
if (clamp) {
if ((r < cast<float>(0ULL))) {
r = cast<float>(0ULL);
}
if ((r > cast<float>(255ULL))) {
r = cast<float>(255ULL);
}
}
return r;
}
}
// tools/platform/gc/gpu_tev.go:308:1
std::array<float,3> gc_colorArg(int64_t code,std::array<std::array<float,4>,4>* reg,std::array<float,4> tex,std::array<float,4> ras,std::array<float,3> konst){
{
auto rep = [&](float v)->std::array<float,3>{
return std::array<float,3>{v,v,v};
}
;
{
switch(code){
case cast<int64_t>(0ULL):{
return std::array<float,3>{(*reg)[cast<int64_t>(0ULL)][cast<int64_t>(0ULL)],(*reg)[cast<int64_t>(0ULL)][cast<int64_t>(1ULL)],(*reg)[cast<int64_t>(0ULL)][cast<int64_t>(2ULL)]};
break;}
case cast<int64_t>(1ULL):{
return rep((*reg)[cast<int64_t>(0ULL)][cast<int64_t>(3ULL)]);
break;}
case cast<int64_t>(2ULL):{
return std::array<float,3>{(*reg)[cast<int64_t>(1ULL)][cast<int64_t>(0ULL)],(*reg)[cast<int64_t>(1ULL)][cast<int64_t>(1ULL)],(*reg)[cast<int64_t>(1ULL)][cast<int64_t>(2ULL)]};
break;}
case cast<int64_t>(3ULL):{
return rep((*reg)[cast<int64_t>(1ULL)][cast<int64_t>(3ULL)]);
break;}
case cast<int64_t>(4ULL):{
return std::array<float,3>{(*reg)[cast<int64_t>(2ULL)][cast<int64_t>(0ULL)],(*reg)[cast<int64_t>(2ULL)][cast<int64_t>(1ULL)],(*reg)[cast<int64_t>(2ULL)][cast<int64_t>(2ULL)]};
break;}
case cast<int64_t>(5ULL):{
return rep((*reg)[cast<int64_t>(2ULL)][cast<int64_t>(3ULL)]);
break;}
case cast<int64_t>(6ULL):{
return std::array<float,3>{(*reg)[cast<int64_t>(3ULL)][cast<int64_t>(0ULL)],(*reg)[cast<int64_t>(3ULL)][cast<int64_t>(1ULL)],(*reg)[cast<int64_t>(3ULL)][cast<int64_t>(2ULL)]};
break;}
case cast<int64_t>(7ULL):{
return rep((*reg)[cast<int64_t>(3ULL)][cast<int64_t>(3ULL)]);
break;}
case cast<int64_t>(8ULL):{
return std::array<float,3>{tex[cast<int64_t>(0ULL)],tex[cast<int64_t>(1ULL)],tex[cast<int64_t>(2ULL)]};
break;}
case cast<int64_t>(9ULL):{
return rep(tex[cast<int64_t>(3ULL)]);
break;}
case cast<int64_t>(10ULL):{
return std::array<float,3>{ras[cast<int64_t>(0ULL)],ras[cast<int64_t>(1ULL)],ras[cast<int64_t>(2ULL)]};
break;}
case cast<int64_t>(11ULL):{
return rep(ras[cast<int64_t>(3ULL)]);
break;}
case cast<int64_t>(12ULL):{
return rep(cast<float>(255ULL));
break;}
case cast<int64_t>(13ULL):{
return rep(cast<float>(128ULL));
break;}
case cast<int64_t>(14ULL):{
return konst;
break;}
default:{
return std::array<float,3>{};
break;}
}}
}
}
// tools/platform/gc/gpu_tev.go:347:1
float gc_alphaArg(int64_t code,std::array<std::array<float,4>,4>* reg,std::array<float,4> tex,std::array<float,4> ras,float konst){
{
{
switch(code){
case cast<int64_t>(0ULL):{
return (*reg)[cast<int64_t>(0ULL)][cast<int64_t>(3ULL)];
break;}
case cast<int64_t>(1ULL):{
return (*reg)[cast<int64_t>(1ULL)][cast<int64_t>(3ULL)];
break;}
case cast<int64_t>(2ULL):{
return (*reg)[cast<int64_t>(2ULL)][cast<int64_t>(3ULL)];
break;}
case cast<int64_t>(3ULL):{
return (*reg)[cast<int64_t>(3ULL)][cast<int64_t>(3ULL)];
break;}
case cast<int64_t>(4ULL):{
return tex[cast<int64_t>(3ULL)];
break;}
case cast<int64_t>(5ULL):{
return ras[cast<int64_t>(3ULL)];
break;}
case cast<int64_t>(6ULL):{
return konst;
break;}
default:{
return cast<float>(0ULL);
break;}
}}
}
}
// tools/platform/gc/gpu_tev.go:371:1
std::array<float,3> gc_gpu_konstColor(gc_gpu* g,int64_t stage){
{
int64_t sel = gc_gpu_kSel(g,stage,false);
if ((sel < cast<int64_t>(8ULL))) {
float v = ((cast<float>(cast<int64_t>((cast<int64_t>(8ULL) - sel))) / cast<float>(8ULL)) * cast<float>(255ULL));
return std::array<float,3>{v,v,v};
}
if ((sel >= cast<int64_t>(12ULL))) {
auto tmp130 = gc_tevColorReg(g->TevKonstReg[cast<int64_t>(((cast<int64_t>((sel - cast<int64_t>(12ULL)))) & cast<int64_t>(3ULL)))]);
float r = std::get<0>(tmp130);
float gg = std::get<1>(tmp130);
float b = std::get<2>(tmp130);
float a = std::get<3>(tmp130);
{
switch(divi<int64_t>((cast<int64_t>((sel - cast<int64_t>(12ULL)))),cast<int64_t>(4ULL))){
case cast<int64_t>(0ULL):{
return std::array<float,3>{r,gg,b};
break;}
case cast<int64_t>(1ULL):{
return std::array<float,3>{r,r,r};
break;}
case cast<int64_t>(2ULL):{
return std::array<float,3>{gg,gg,gg};
break;}
case cast<int64_t>(3ULL):{
return std::array<float,3>{b,b,b};
break;}
default:{
return std::array<float,3>{a,a,a};
break;}
}}
}
return std::array<float,3>{};
}
}
// tools/platform/gc/gpu_tev.go:395:1
float gc_gpu_konstAlpha(gc_gpu* g,int64_t stage){
{
int64_t sel = gc_gpu_kSel(g,stage,true);
if ((sel < cast<int64_t>(8ULL))) {
return ((cast<float>(cast<int64_t>((cast<int64_t>(8ULL) - sel))) / cast<float>(8ULL)) * cast<float>(255ULL));
}
if ((sel >= cast<int64_t>(16ULL))) {
auto tmp131 = gc_tevColorReg(g->TevKonstReg[cast<int64_t>(((cast<int64_t>((sel - cast<int64_t>(16ULL)))) & cast<int64_t>(3ULL)))]);
float r = std::get<0>(tmp131);
float gg = std::get<1>(tmp131);
float b = std::get<2>(tmp131);
float a = std::get<3>(tmp131);
{
switch(divi<int64_t>((cast<int64_t>((sel - cast<int64_t>(16ULL)))),cast<int64_t>(4ULL))){
case cast<int64_t>(0ULL):{
return r;
break;}
case cast<int64_t>(1ULL):{
return gg;
break;}
case cast<int64_t>(2ULL):{
return b;
break;}
default:{
return a;
break;}
}}
}
return cast<float>(0ULL);
}
}
// tools/platform/gc/gpu_tev.go:423:1
std::array<int64_t,4> gc_gpu_swapTable(gc_gpu* g,int64_t t){
{
uint32_t lo = g->BP[cast<uint32_t>((cast<uint32_t>(246ULL) + cast<uint32_t>((cast<uint32_t>(t) * cast<uint32_t>(2ULL)))))];
uint32_t hi = g->BP[cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(246ULL) + cast<uint32_t>((cast<uint32_t>(t) * cast<uint32_t>(2ULL))))) + cast<uint32_t>(1ULL)))];
return std::array<int64_t,4>{cast<int64_t>(cast<uint32_t>((lo & cast<uint32_t>(3ULL)))),cast<int64_t>(cast<uint32_t>(((shr<uint32_t>(lo,cast<int64_t>(2ULL))) & cast<uint32_t>(3ULL)))),cast<int64_t>(cast<uint32_t>((hi & cast<uint32_t>(3ULL)))),cast<int64_t>(cast<uint32_t>(((shr<uint32_t>(hi,cast<int64_t>(2ULL))) & cast<uint32_t>(3ULL))))};
}
}
// tools/platform/gc/gpu_tev.go:431:1
std::array<float,4> gc_swizzle(std::array<int64_t,4> t,std::array<float,4> v){
{
return std::array<float,4>{v[t[cast<int64_t>(0ULL)]],v[t[cast<int64_t>(1ULL)]],v[t[cast<int64_t>(2ULL)]],v[t[cast<int64_t>(3ULL)]]};
}
}
// tools/platform/gc/gpu_tev.go:437:1
int64_t gc_gpu_kSel(gc_gpu* g,int64_t stage,bool alpha){
{
uint32_t reg = g->BP[cast<uint32_t>((cast<uint32_t>(246ULL) + cast<uint32_t>(divi<int64_t>(stage,cast<int64_t>(2ULL)))))];
uint64_t shift = cast<uint64_t>(4ULL);
if ((cast<int64_t>((stage & cast<int64_t>(1ULL))) == cast<int64_t>(1ULL))) {
shift = cast<uint64_t>(14ULL);
}
if (alpha) {
shift += cast<uint64_t>(5ULL);
}
return cast<int64_t>(cast<uint32_t>(((shr<uint32_t>(reg,shift)) & cast<uint32_t>(31ULL))));
}
}
// tools/platform/gc/gpu_tev.go:451:1
bool gc_gpu_alphaTest(gc_gpu* g,uint8_t a){
{
uint32_t f = g->BP[cast<int64_t>(243ULL)];
uint8_t ref0 = cast<uint8_t>(cast<uint32_t>((f & cast<uint32_t>(255ULL))));
uint8_t ref1 = cast<uint8_t>(cast<uint32_t>(((shr<uint32_t>(f,cast<int64_t>(8ULL))) & cast<uint32_t>(255ULL))));
bool r0 = gc_alphaCompare(a,ref0,cast<int64_t>(cast<uint32_t>(((shr<uint32_t>(f,cast<int64_t>(16ULL))) & cast<uint32_t>(7ULL)))));
bool r1 = gc_alphaCompare(a,ref1,cast<int64_t>(cast<uint32_t>(((shr<uint32_t>(f,cast<int64_t>(19ULL))) & cast<uint32_t>(7ULL)))));
{
switch(cast<uint32_t>(((shr<uint32_t>(f,cast<int64_t>(22ULL))) & cast<uint32_t>(3ULL)))){
case cast<uint32_t>(0ULL):{
return (r0 && r1);
break;}
case cast<uint32_t>(1ULL):{
return (r0 || r1);
break;}
case cast<uint32_t>(2ULL):{
return (r0 != r1);
break;}
default:{
return (r0 == r1);
break;}
}}
}
}
// tools/platform/gc/gpu_tev.go:469:1
bool gc_alphaCompare(uint8_t a,uint8_t ref,int64_t comp){
{
{
switch(comp){
case cast<int64_t>(0ULL):{
return false;
break;}
case cast<int64_t>(1ULL):{
return (a < ref);
break;}
case cast<int64_t>(2ULL):{
return (a == ref);
break;}
case cast<int64_t>(3ULL):{
return (a <= ref);
break;}
case cast<int64_t>(4ULL):{
return (a > ref);
break;}
case cast<int64_t>(5ULL):{
return (a != ref);
break;}
case cast<int64_t>(6ULL):{
return (a >= ref);
break;}
default:{
return true;
break;}
}}
}
}
// tools/platform/gc/gpu_tev.go:494:1
uint32_t gc_gpu_blend(gc_gpu* g,uint32_t dst,uint8_t sr,uint8_t sg,uint8_t sb,uint8_t sa){
{
uint32_t cm = g->BP[cast<int64_t>(65ULL)];
auto tmp132 = gc_unpackRGB(dst);
uint8_t dr = std::get<0>(tmp132);
uint8_t dgc = std::get<1>(tmp132);
uint8_t db = std::get<2>(tmp132);
uint8_t da = cast<uint8_t>(dst);
if ((cast<uint32_t>((cm & cast<uint32_t>(1ULL))) == cast<uint32_t>(0ULL))) {
return gc_gpu_applyUpdates(g,cm,dst,sr,sg,sb,sa);
}
float saf = (cast<float>(sa) / cast<float>(255ULL));
float daf = (cast<float>(da) / cast<float>(255ULL));
auto srcF = [&](float chanOther)->float{
return gc_blendFactor(cast<int64_t>(cast<uint32_t>(((shr<uint32_t>(cm,cast<int64_t>(8ULL))) & cast<uint32_t>(7ULL)))),(chanOther / cast<float>(255ULL)),saf,daf);
}
;
auto dstF = [&](float chanSelf)->float{
return gc_blendFactor(cast<int64_t>(cast<uint32_t>(((shr<uint32_t>(cm,cast<int64_t>(5ULL))) & cast<uint32_t>(7ULL)))),(chanSelf / cast<float>(255ULL)),saf,daf);
}
;
bool sub_ = (cast<uint32_t>(((shr<uint32_t>(cm,cast<int64_t>(11ULL))) & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL));
auto comb = [&](uint8_t s,uint8_t d)->uint8_t{
float sf = srcF(cast<float>(d));
float df = dstF(cast<float>(s));
float v={};
if (sub_) {
v = ((cast<float>(d) * df) - (cast<float>(s) * sf));
}
else {
v = ((cast<float>(s) * sf) + (cast<float>(d) * df));
}
return gc_toU8(v);
}
;
uint8_t nr = comb(sr,dr);
uint8_t ng = comb(sg,dgc);
uint8_t nb = comb(sb,db);
uint8_t na = comb(sa,da);
return gc_gpu_applyUpdates(g,cm,dst,nr,ng,nb,na);
}
}
// tools/platform/gc/gpu_tev.go:529:1
uint32_t gc_gpu_applyUpdates(gc_gpu* g,uint32_t cm,uint32_t dst,uint8_t r,uint8_t gg,uint8_t b,uint8_t a){
{
auto tmp133 = gc_unpackRGB(dst);
uint8_t or_ = std::get<0>(tmp133);
uint8_t og = std::get<1>(tmp133);
uint8_t ob = std::get<2>(tmp133);
uint8_t oa = cast<uint8_t>(dst);
if ((cast<uint32_t>((cm & cast<uint32_t>(8ULL))) != cast<uint32_t>(0ULL))) {
auto tmp134 = std::make_tuple(r,gg,b);
or_ = std::get<0>(tmp134);
og = std::get<1>(tmp134);
ob = std::get<2>(tmp134);
}
if ((cast<uint32_t>((cm & cast<uint32_t>(16ULL))) != cast<uint32_t>(0ULL))) {
oa = a;
}
return gc_packRGBA(or_,og,ob,oa);
}
}
// tools/platform/gc/gpu_tev.go:544:1
float gc_blendFactor(int64_t code,float chanOther,float sa,float da){
{
{
switch(code){
case cast<int64_t>(0ULL):{
return cast<float>(0ULL);
break;}
case cast<int64_t>(1ULL):{
return cast<float>(1ULL);
break;}
case cast<int64_t>(2ULL):{
return chanOther;
break;}
case cast<int64_t>(3ULL):{
return (cast<float>(1ULL) - chanOther);
break;}
case cast<int64_t>(4ULL):{
return sa;
break;}
case cast<int64_t>(5ULL):{
return (cast<float>(1ULL) - sa);
break;}
case cast<int64_t>(6ULL):{
return da;
break;}
default:{
return (cast<float>(1ULL) - da);
break;}
}}
}
}
// tools/platform/gc/gpu_tev.go:566:1
float gc_clampf(float v){
{
if ((v < cast<float>(0ULL))) {
return cast<float>(0ULL);
}
if ((v > cast<float>(255ULL))) {
return cast<float>(255ULL);
}
return v;
}
}
// tools/platform/gc/gpu_tev_state.go:81:1
bool gc_texCanHalt(gc_texState* tx){
{
{
switch(tx->format){
case cast<int64_t>(0ULL):case cast<int64_t>(1ULL):case cast<int64_t>(2ULL):case cast<int64_t>(3ULL):case cast<int64_t>(4ULL):case cast<int64_t>(5ULL):case cast<int64_t>(6ULL):case cast<int64_t>(14ULL):{
return false;
break;}
case cast<int64_t>(8ULL):case cast<int64_t>(9ULL):case cast<int64_t>(10ULL):{
return (tx->tlutFmt > cast<int64_t>(2ULL));
break;}
default:{
return true;
break;}
}}
}
}
// tools/platform/gc/gpu_tev_state.go:94:1
gc_tevState gc_gpu_tevstate_reference(gc_gpu* g){
{
gc_tevState t={};
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(4ULL));i++){
auto tmp135 = gc_tevColorReg(g->TevColorReg[i]);
float r = std::get<0>(tmp135);
float gg = std::get<1>(tmp135);
float b = std::get<2>(tmp135);
float a = std::get<3>(tmp135);
t.seed[i] = std::array<float,4>{r,gg,b,a};
}
}t.numStages = cast<int64_t>((cast<int64_t>(cast<uint32_t>(((shr<uint32_t>(g->BP[cast<int64_t>(0ULL)],cast<int64_t>(10ULL))) & cast<uint32_t>(15ULL)))) + cast<int64_t>(1ULL)));
{int64_t s = cast<int64_t>(0ULL);for (;(s < t.numStages);s++){
gc_tevStage* st = (&t.stages[s]);
uint32_t ord = g->BP[cast<uint32_t>((cast<uint32_t>(40ULL) + cast<uint32_t>(divi<int64_t>(s,cast<int64_t>(2ULL)))))];
if ((cast<int64_t>((s & cast<int64_t>(1ULL))) == cast<int64_t>(1ULL))) {
ord = shr<uint32_t>(ord,cast<int64_t>(12ULL));
}
st->texmap = cast<int64_t>(cast<uint32_t>((ord & cast<uint32_t>(7ULL))));
st->texcoord = cast<int64_t>(cast<uint32_t>(((shr<uint32_t>(ord,cast<int64_t>(3ULL))) & cast<uint32_t>(7ULL))));
st->texEnable = (cast<uint32_t>(((shr<uint32_t>(ord,cast<int64_t>(6ULL))) & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL));
st->rasSel = cast<uint32_t>(((shr<uint32_t>(ord,cast<int64_t>(7ULL))) & cast<uint32_t>(7ULL)));
st->cc = g->BP[cast<uint32_t>((cast<uint32_t>(192ULL) + cast<uint32_t>((cast<uint32_t>(s) * cast<uint32_t>(2ULL)))))];
st->ac = g->BP[cast<uint32_t>((cast<uint32_t>(193ULL) + cast<uint32_t>((cast<uint32_t>(s) * cast<uint32_t>(2ULL)))))];
st->swapRas = gc_gpu_swapTable(g,cast<int64_t>(cast<uint32_t>((st->ac & cast<uint32_t>(3ULL)))));
st->swapTex = gc_gpu_swapTable(g,cast<int64_t>(cast<uint32_t>(((shr<uint32_t>(st->ac,cast<int64_t>(2ULL))) & cast<uint32_t>(3ULL)))));
st->konstC = gc_gpu_konstColor(g,s);
st->konstA = gc_gpu_konstAlpha(g,s);
st->cdest = cast<uint32_t>(((shr<uint32_t>(st->cc,cast<int64_t>(22ULL))) & cast<uint32_t>(3ULL)));
st->adest = cast<uint32_t>(((shr<uint32_t>(st->ac,cast<int64_t>(22ULL))) & cast<uint32_t>(3ULL)));
if ((st->texEnable && (!t.texValid[st->texmap]))) {
t.tex[st->texmap] = gc_gpu_texSetup(g,st->texmap);
t.texValid[st->texmap] = true;
if (gc_texCanHalt((&t.tex[st->texmap]))) {
t.canHalt = true;
}
}
}
}return t;
}
}
// tools/platform/gc/gpu_texgen.go:72:1
int64_t gc_gpu_texGenCount(gc_gpu* g){
{
int64_t n = cast<int64_t>(cast<uint32_t>((g->XFMem[cast<int64_t>(4159ULL)] & cast<uint32_t>(15ULL))));
if ((n > cast<int64_t>(8ULL))) {
n = cast<int64_t>(8ULL);
}
return n;
}
}
// tools/platform/gc/gpu_texgen.go:83:1
int64_t gc_gpu_texMtxRow(gc_gpu* g,int64_t i){
{
if ((i < cast<int64_t>(4ULL))) {
return cast<int64_t>(cast<uint32_t>(((shr<uint32_t>(g->CPReg[cast<int64_t>(48ULL)],(cast<uint32_t>((cast<uint32_t>(6ULL) + cast<uint32_t>((cast<uint32_t>(6ULL) * cast<uint32_t>(i)))))))) & cast<uint32_t>(63ULL))));
}
return cast<int64_t>(cast<uint32_t>(((shr<uint32_t>(g->CPReg[cast<int64_t>(49ULL)],(cast<uint32_t>((cast<uint32_t>(6ULL) * cast<uint32_t>(cast<int64_t>((i - cast<int64_t>(4ULL))))))))) & cast<uint32_t>(63ULL))));
}
}
// tools/platform/gc/gpu_texgen.go:95:1
gc_texCoord gc_gpu_genTexCoord(gc_gpu* g,gc_Machine* m,int64_t i,int64_t mtxRow,float mx,float my,float mz,float nx,float ny,float nz,std::array<gc_texCoord,8>* vtc,std::array<std::array<uint8_t,4>,2>* col){
{
uint32_t info = g->XFMem[cast<uint32_t>((cast<uint32_t>(4160ULL) + cast<uint32_t>(i)))];
uint32_t proj = cast<uint32_t>(((shr<uint32_t>(info,cast<int64_t>(1ULL))) & cast<uint32_t>(1ULL)));
uint32_t inputForm = cast<uint32_t>(((shr<uint32_t>(info,cast<int64_t>(2ULL))) & cast<uint32_t>(3ULL)));
uint32_t genType = cast<uint32_t>(((shr<uint32_t>(info,cast<int64_t>(4ULL))) & cast<uint32_t>(7ULL)));
uint32_t srcRow = cast<uint32_t>(((shr<uint32_t>(info,cast<int64_t>(7ULL))) & cast<uint32_t>(31ULL)));
{
switch(genType){
case cast<uint32_t>(0ULL):{
break;}
case cast<uint32_t>(2ULL):case cast<uint32_t>(3ULL):{
std::array<uint8_t,4> c = (*col)[cast<uint32_t>((genType - cast<uint32_t>(2ULL)))];
return gc_texCoord{cast<float>((cast<float>(c[cast<int64_t>(0ULL)]) / cast<float>(255ULL))),cast<float>((cast<float>(c[cast<int64_t>(1ULL)]) / cast<float>(255ULL))),cast<float>(cast<float>(1ULL))};
break;}
default:{
gc_Machine_logf(m,std::string("XF: texgen %d uses generation type %d, only the regular and colour types are implemented",88),i,genType);
return gc_texCoord{{},{},cast<float>(cast<float>(1ULL))};
break;}
}}
std::array<float,4> in={};
{
if ((srcRow == cast<uint32_t>(0ULL))){
in = std::array<float,4>{mx,my,mz,cast<float>(1ULL)};
}
else if ((srcRow == cast<uint32_t>(1ULL))){
in = std::array<float,4>{nx,ny,nz,cast<float>(1ULL)};
}
else if (((srcRow >= cast<uint32_t>(5ULL)) && (srcRow < cast<uint32_t>(13ULL)))){
gc_texCoord v = (*vtc)[cast<uint32_t>((srcRow - cast<uint32_t>(5ULL)))];
in = std::array<float,4>{v.s,v.t,cast<float>(1ULL),cast<float>(1ULL)};
}
else {
gc_Machine_logf(m,std::string("XF: texgen %d sources row %d, which the vertex fetch does not capture",69),i,srcRow);
return gc_texCoord{{},{},cast<float>(cast<float>(1ULL))};
}
}
tmp136:;
if ((inputForm == cast<uint32_t>(0ULL))) {
in[cast<int64_t>(2ULL)] = cast<float>(1ULL);
}
int64_t base = cast<int64_t>((mtxRow * cast<int64_t>(4ULL)));
auto dot = [&](int64_t r)->float{
return ((((gc_gpu_xfFloat(g,cast<int64_t>((base + cast<int64_t>((r * cast<int64_t>(4ULL)))))) * in[cast<int64_t>(0ULL)]) + (gc_gpu_xfFloat(g,cast<int64_t>((cast<int64_t>((base + cast<int64_t>((r * cast<int64_t>(4ULL))))) + cast<int64_t>(1ULL)))) * in[cast<int64_t>(1ULL)])) + (gc_gpu_xfFloat(g,cast<int64_t>((cast<int64_t>((base + cast<int64_t>((r * cast<int64_t>(4ULL))))) + cast<int64_t>(2ULL)))) * in[cast<int64_t>(2ULL)])) + (gc_gpu_xfFloat(g,cast<int64_t>((cast<int64_t>((base + cast<int64_t>((r * cast<int64_t>(4ULL))))) + cast<int64_t>(3ULL)))) * in[cast<int64_t>(3ULL)]));
}
;
gc_texCoord out = gc_texCoord{cast<float>(dot(cast<int64_t>(0ULL))),cast<float>(dot(cast<int64_t>(1ULL))),cast<float>(cast<float>(1ULL))};
if ((proj == cast<uint32_t>(1ULL))) {
out.q = dot(cast<int64_t>(2ULL));
}
if ((cast<uint32_t>((g->XFMem[cast<int64_t>(4114ULL)] & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL))) {
uint32_t pi = g->XFMem[cast<uint32_t>((cast<uint32_t>(4176ULL) + cast<uint32_t>(i)))];
int64_t pbase = cast<int64_t>((cast<int64_t>(1280ULL) + cast<int64_t>((cast<int64_t>(cast<uint32_t>((pi & cast<uint32_t>(63ULL)))) * cast<int64_t>(4ULL)))));
std::array<float,3> v = std::array<float,3>{out.s,out.t,out.q};
if ((cast<uint32_t>(((shr<uint32_t>(pi,cast<int64_t>(8ULL))) & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL))) {
v = gc_normalize3(v);
}
auto pdot = [&](int64_t r)->float{
return ((((gc_gpu_xfFloat(g,cast<int64_t>((pbase + cast<int64_t>((r * cast<int64_t>(4ULL)))))) * v[cast<int64_t>(0ULL)]) + (gc_gpu_xfFloat(g,cast<int64_t>((cast<int64_t>((pbase + cast<int64_t>((r * cast<int64_t>(4ULL))))) + cast<int64_t>(1ULL)))) * v[cast<int64_t>(1ULL)])) + (gc_gpu_xfFloat(g,cast<int64_t>((cast<int64_t>((pbase + cast<int64_t>((r * cast<int64_t>(4ULL))))) + cast<int64_t>(2ULL)))) * v[cast<int64_t>(2ULL)])) + gc_gpu_xfFloat(g,cast<int64_t>((cast<int64_t>((pbase + cast<int64_t>((r * cast<int64_t>(4ULL))))) + cast<int64_t>(3ULL)))));
}
;
out = gc_texCoord{cast<float>(pdot(cast<int64_t>(0ULL))),cast<float>(pdot(cast<int64_t>(1ULL))),cast<float>(pdot(cast<int64_t>(2ULL)))};
}
if ((proj != cast<uint32_t>(1ULL))) {
out.q = cast<float>(1ULL);
}
return out;
}
}
// tools/platform/gc/gpu_texgen.go:180:1
std::array<float,3> gc_normalize3(std::array<float,3> v){
{
float l = cast<float>(go_math_Sqrt(cast<double>((((v[cast<int64_t>(0ULL)] * v[cast<int64_t>(0ULL)]) + (v[cast<int64_t>(1ULL)] * v[cast<int64_t>(1ULL)])) + (v[cast<int64_t>(2ULL)] * v[cast<int64_t>(2ULL)])))));
if ((l == cast<float>(0ULL))) {
return v;
}
return std::array<float,3>{(v[cast<int64_t>(0ULL)] / l),(v[cast<int64_t>(1ULL)] / l),(v[cast<int64_t>(2ULL)] / l)};
}
}
// tools/platform/gc/gpu_texture.go:73:1
gc_texState gc_gpu_texSetup(gc_gpu* g,int64_t i){
{
uint32_t mode0={};
uint32_t image0={};
uint32_t image3={};
uint32_t settlut={};
if ((i < cast<int64_t>(4ULL))) {
mode0 = g->BP[cast<uint32_t>((cast<uint32_t>(128ULL) + cast<uint32_t>(i)))];
image0 = g->BP[cast<uint32_t>((cast<uint32_t>(136ULL) + cast<uint32_t>(i)))];
image3 = g->BP[cast<uint32_t>((cast<uint32_t>(148ULL) + cast<uint32_t>(i)))];
settlut = g->BP[cast<uint32_t>((cast<uint32_t>(152ULL) + cast<uint32_t>(i)))];
}
else {
mode0 = g->BP[cast<uint32_t>((cast<uint32_t>(160ULL) + cast<uint32_t>(cast<int64_t>((i - cast<int64_t>(4ULL))))))];
image0 = g->BP[cast<uint32_t>((cast<uint32_t>(168ULL) + cast<uint32_t>(cast<int64_t>((i - cast<int64_t>(4ULL))))))];
image3 = g->BP[cast<uint32_t>((cast<uint32_t>(180ULL) + cast<uint32_t>(cast<int64_t>((i - cast<int64_t>(4ULL))))))];
settlut = g->BP[cast<uint32_t>((cast<uint32_t>(184ULL) + cast<uint32_t>(cast<int64_t>((i - cast<int64_t>(4ULL))))))];
}
return gc_texState{cast<int64_t>(cast<uint32_t>(((shr<uint32_t>(image0,cast<int64_t>(20ULL))) & cast<uint32_t>(15ULL)))),cast<int64_t>((cast<int64_t>(cast<uint32_t>((image0 & cast<uint32_t>(1023ULL)))) + cast<int64_t>(1ULL))),cast<int64_t>((cast<int64_t>(cast<uint32_t>(((shr<uint32_t>(image0,cast<int64_t>(10ULL))) & cast<uint32_t>(1023ULL)))) + cast<int64_t>(1ULL))),shl<uint32_t>((cast<uint32_t>((image3 & cast<uint32_t>(16777215ULL)))),cast<int64_t>(5ULL)),cast<int64_t>(cast<uint32_t>((mode0 & cast<uint32_t>(3ULL)))),cast<int64_t>(cast<uint32_t>(((shr<uint32_t>(mode0,cast<int64_t>(2ULL))) & cast<uint32_t>(3ULL)))),shl<int64_t>(cast<int64_t>(cast<uint32_t>((settlut & cast<uint32_t>(1023ULL)))),cast<int64_t>(9ULL)),cast<int64_t>(cast<uint32_t>(((shr<uint32_t>(settlut,cast<int64_t>(10ULL))) & cast<uint32_t>(3ULL)))),{},{}};
}
}
// tools/platform/gc/gpu_texture.go:100:1
void gc_gpu_loadTlut(gc_gpu* g,gc_Machine* m,uint32_t data){
{
uint32_t src = gc_phys(shl<uint32_t>((cast<uint32_t>((g->BP[cast<int64_t>(100ULL)] & cast<uint32_t>(16777215ULL)))),cast<int64_t>(5ULL)));
int64_t off = shl<int64_t>(cast<int64_t>(cast<uint32_t>((data & cast<uint32_t>(1023ULL)))),cast<int64_t>(9ULL));
int64_t n = cast<int64_t>((cast<int64_t>(cast<uint32_t>(((shr<uint32_t>(data,cast<int64_t>(10ULL))) & cast<uint32_t>(2047ULL)))) * cast<int64_t>(32ULL)));
if ((!g->Tlut)) {
g->Tlut = Slice<uint8_t>::make(cast<int64_t>(524288ULL));
}
if (((cast<int64_t>((off + n)) > len(g->Tlut)) || (cast<int64_t>((cast<int64_t>(src) + n)) > len(m->RAM)))) {
gekko_CPU_Halt(m->CPU,std::string("TLUT load out of range: src 0x%08X, tmem offset 0x%X, %d bytes",62),src,off,n);
return ;
}
gcopy(sub(g->Tlut,off,cast<int64_t>((off + n))),sub(m->RAM,src,cast<int64_t>((cast<int64_t>(src) + n))));
if (gc_drawTrace) {
go_fmt_Fprintf(go_os_Stderr,std::string("TLUT load 0x%08X -> tmem+0x%05X %d bytes\012",41),src,off,n);
}
}
}
// tools/platform/gc/gpu_texture.go:119:1
std::tuple<uint8_t,uint8_t,uint8_t,uint8_t> gc_gpu_tlutColor(gc_gpu* g,gc_Machine* m,const gc_texState& tx,int64_t idx){
uint8_t r{};
uint8_t gg{};
uint8_t b{};
uint8_t a{};
{
uint16_t v={};
if (tx.tlutFromRAM) {
uint32_t o = gc_phys(cast<uint32_t>((cast<uint32_t>((tx.tlutRAM + cast<uint32_t>(tx.tlutOff))) + cast<uint32_t>((cast<uint32_t>(idx) * cast<uint32_t>(2ULL))))));
if ((cast<int64_t>((cast<int64_t>(o) + cast<int64_t>(1ULL))) >= len(m->RAM))) {
return {cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),cast<uint8_t>(0ULL)};
}
v = cast<uint16_t>((shl<uint16_t>(cast<uint16_t>(m->RAM[o]),cast<int64_t>(8ULL)) | cast<uint16_t>(m->RAM[cast<uint32_t>((o + cast<uint32_t>(1ULL)))])));
}
else {
int64_t o = cast<int64_t>((tx.tlutOff + cast<int64_t>((idx * cast<int64_t>(2ULL)))));
if (((!g->Tlut) || (cast<int64_t>((o + cast<int64_t>(1ULL))) >= len(g->Tlut)))) {
return {cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),cast<uint8_t>(0ULL)};
}
v = cast<uint16_t>((shl<uint16_t>(cast<uint16_t>(g->Tlut[o]),cast<int64_t>(8ULL)) | cast<uint16_t>(g->Tlut[cast<int64_t>((o + cast<int64_t>(1ULL)))])));
}
{
switch(tx.tlutFmt){
case cast<int64_t>(0ULL):{
uint8_t i = cast<uint8_t>(cast<uint16_t>((v & cast<uint16_t>(255ULL))));
return {i,i,i,cast<uint8_t>(shr<uint16_t>(v,cast<int64_t>(8ULL)))};
break;}
case cast<int64_t>(1ULL):{
return gc_decodeRGB565(v);
break;}
case cast<int64_t>(2ULL):{
return gc_decodeRGB5A3(v);
break;}
default:{
gekko_CPU_Halt(m->CPU,std::string("TLUT: entry format %d is not a hardware format",46),tx.tlutFmt);
return {cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),cast<uint8_t>(0ULL)};
break;}
}}
}
}
// tools/platform/gc/gpu_texture.go:156:1
std::tuple<uint8_t,uint8_t,uint8_t,uint8_t> gc_gpu_sampleTexmap(gc_gpu* g,gc_Machine* m,gc_texState* tx,float s,float t){
uint8_t r{};
uint8_t gg{};
uint8_t b{};
uint8_t a{};
{
int64_t x = gc_wrapCoord(cast<int64_t>((s * cast<float>(tx->width))),tx->width,tx->wrapS);
int64_t y = gc_wrapCoord(cast<int64_t>((t * cast<float>(tx->height))),tx->height,tx->wrapT);
return gc_gpu_decodeTexel(g,m,(*tx),x,y);
}
}
// tools/platform/gc/gpu_texture.go:164:1
int64_t gc_wrapCoord(int64_t v,int64_t size,int64_t mode){
{
if ((size <= cast<int64_t>(0ULL))) {
return cast<int64_t>(0ULL);
}
{
switch(mode){
case cast<int64_t>(1ULL):{
v %= size;
if ((v < cast<int64_t>(0ULL))) {
v += size;
}
return v;
break;}
case cast<int64_t>(2ULL):{
int64_t p = cast<int64_t>((size * cast<int64_t>(2ULL)));
v %= p;
if ((v < cast<int64_t>(0ULL))) {
v += p;
}
if ((v >= size)) {
v = cast<int64_t>((cast<int64_t>((p - cast<int64_t>(1ULL))) - v));
}
return v;
break;}
default:{
if ((v < cast<int64_t>(0ULL))) {
return cast<int64_t>(0ULL);
}
if ((v >= size)) {
return cast<int64_t>((size - cast<int64_t>(1ULL)));
}
return v;
break;}
}}
}
}
// tools/platform/gc/gpu_texture.go:198:1
std::tuple<uint8_t,uint8_t,uint8_t,uint8_t> gc_gpu_decodeTexel(gc_gpu* g,gc_Machine* m,const gc_texState& tx,int64_t x,int64_t y){
uint8_t r{};
uint8_t gg{};
uint8_t b{};
uint8_t a{};
{
{
switch(tx.format){
case cast<int64_t>(0ULL):{
uint8_t v = gc_gpu_tileNibble(g,m,tx.base,tx.width,cast<int64_t>(8ULL),cast<int64_t>(8ULL),x,y);
uint8_t i = gc_expand4(cast<uint16_t>(v));
return {i,i,i,i};
break;}
case cast<int64_t>(1ULL):{
uint8_t v = gc_gpu_tileByte(g,m,tx.base,tx.width,cast<int64_t>(8ULL),cast<int64_t>(4ULL),x,y);
return {v,v,v,v};
break;}
case cast<int64_t>(2ULL):{
uint8_t v = gc_gpu_tileByte(g,m,tx.base,tx.width,cast<int64_t>(8ULL),cast<int64_t>(4ULL),x,y);
uint8_t i = gc_expand4(cast<uint16_t>(cast<uint8_t>((v & cast<uint8_t>(15ULL)))));
uint8_t al = gc_expand4(cast<uint16_t>(shr<uint8_t>(v,cast<int64_t>(4ULL))));
return {i,i,i,al};
break;}
case cast<int64_t>(3ULL):{
uint16_t v = gc_gpu_tileHalf(g,m,tx.base,tx.width,cast<int64_t>(4ULL),cast<int64_t>(4ULL),x,y);
uint8_t i = cast<uint8_t>(cast<uint16_t>((v & cast<uint16_t>(255ULL))));
uint8_t al = cast<uint8_t>(shr<uint16_t>(v,cast<int64_t>(8ULL)));
return {i,i,i,al};
break;}
case cast<int64_t>(4ULL):{
uint16_t v = gc_gpu_tileHalf(g,m,tx.base,tx.width,cast<int64_t>(4ULL),cast<int64_t>(4ULL),x,y);
return gc_decodeRGB565(v);
break;}
case cast<int64_t>(5ULL):{
uint16_t v = gc_gpu_tileHalf(g,m,tx.base,tx.width,cast<int64_t>(4ULL),cast<int64_t>(4ULL),x,y);
return gc_decodeRGB5A3(v);
break;}
case cast<int64_t>(6ULL):{
return gc_gpu_decodeRGBA8(g,m,tx.base,tx.width,x,y);
break;}
case cast<int64_t>(14ULL):{
return gc_gpu_decodeCMPR(g,m,tx.base,tx.width,x,y);
break;}
case cast<int64_t>(8ULL):{
uint8_t v = gc_gpu_tileNibble(g,m,tx.base,tx.width,cast<int64_t>(8ULL),cast<int64_t>(8ULL),x,y);
return gc_gpu_tlutColor(g,m,tx,cast<int64_t>(v));
break;}
case cast<int64_t>(9ULL):{
uint8_t v = gc_gpu_tileByte(g,m,tx.base,tx.width,cast<int64_t>(8ULL),cast<int64_t>(4ULL),x,y);
return gc_gpu_tlutColor(g,m,tx,cast<int64_t>(v));
break;}
case cast<int64_t>(10ULL):{
uint16_t v = gc_gpu_tileHalf(g,m,tx.base,tx.width,cast<int64_t>(4ULL),cast<int64_t>(4ULL),x,y);
return gc_gpu_tlutColor(g,m,tx,cast<int64_t>(cast<uint16_t>((v & cast<uint16_t>(16383ULL)))));
break;}
default:{
gekko_CPU_Halt(m->CPU,std::string("TX: unknown texture format 0x%X",31),tx.format);
return {cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),cast<uint8_t>(0ULL)};
break;}
}}
}
}
// tools/platform/gc/gpu_texture.go:244:1
std::tuple<int64_t,int64_t> gc_tileByteOffset(int64_t width,int64_t bw,int64_t bh,int64_t x,int64_t y){
int64_t tileBytesBase{};
int64_t inTile{};
{
int64_t tilesPerRow = divi<int64_t>((cast<int64_t>((cast<int64_t>((width + bw)) - cast<int64_t>(1ULL)))),bw);
auto tmp137 = std::make_tuple(divi<int64_t>(x,bw),divi<int64_t>(y,bh));
int64_t bx = std::get<0>(tmp137);
int64_t by = std::get<1>(tmp137);
auto tmp138 = std::make_tuple(modi<int64_t>(x,bw),modi<int64_t>(y,bh));
int64_t ix = std::get<0>(tmp138);
int64_t iy = std::get<1>(tmp138);
int64_t tileIdx = cast<int64_t>((cast<int64_t>((by * tilesPerRow)) + bx));
int64_t texelsPerTile = cast<int64_t>((bw * bh));
return {cast<int64_t>((tileIdx * texelsPerTile)),cast<int64_t>((cast<int64_t>((iy * bw)) + ix))};
}
}
// tools/platform/gc/gpu_texture.go:255:1
uint16_t gc_gpu_tileHalf(gc_gpu* g,gc_Machine* m,uint32_t base,int64_t width,int64_t bw,int64_t bh,int64_t x,int64_t y){
{
auto tmp139 = gc_tileByteOffset(width,bw,bh,x,y);
int64_t tb = std::get<0>(tmp139);
int64_t in = std::get<1>(tmp139);
uint32_t off = cast<uint32_t>((base + cast<uint32_t>(cast<int64_t>(((cast<int64_t>((tb + in))) * cast<int64_t>(2ULL))))));
return cast<uint16_t>((shl<uint16_t>(cast<uint16_t>(gc_Machine_texByte(m,off)),cast<int64_t>(8ULL)) | cast<uint16_t>(gc_Machine_texByte(m,cast<uint32_t>((off + cast<uint32_t>(1ULL)))))));
}
}
// tools/platform/gc/gpu_texture.go:261:1
uint8_t gc_gpu_tileByte(gc_gpu* g,gc_Machine* m,uint32_t base,int64_t width,int64_t bw,int64_t bh,int64_t x,int64_t y){
{
auto tmp140 = gc_tileByteOffset(width,bw,bh,x,y);
int64_t tb = std::get<0>(tmp140);
int64_t in = std::get<1>(tmp140);
return gc_Machine_texByte(m,cast<uint32_t>((base + cast<uint32_t>(cast<int64_t>((tb + in))))));
}
}
// tools/platform/gc/gpu_texture.go:266:1
uint8_t gc_gpu_tileNibble(gc_gpu* g,gc_Machine* m,uint32_t base,int64_t width,int64_t bw,int64_t bh,int64_t x,int64_t y){
{
auto tmp141 = gc_tileByteOffset(width,bw,bh,x,y);
int64_t tb = std::get<0>(tmp141);
int64_t in = std::get<1>(tmp141);
uint8_t b = gc_Machine_texByte(m,cast<uint32_t>((base + cast<uint32_t>(divi<int64_t>((cast<int64_t>((tb + in))),cast<int64_t>(2ULL))))));
if ((cast<int64_t>((in & cast<int64_t>(1ULL))) == cast<int64_t>(0ULL))) {
return shr<uint8_t>(b,cast<int64_t>(4ULL));
}
return cast<uint8_t>((b & cast<uint8_t>(15ULL)));
}
}
// tools/platform/gc/gpu_texture.go:278:1
std::tuple<uint8_t,uint8_t,uint8_t,uint8_t> gc_gpu_decodeRGBA8(gc_gpu* g,gc_Machine* m,uint32_t base,int64_t width,int64_t x,int64_t y){
uint8_t r{};
uint8_t gg{};
uint8_t b{};
uint8_t a{};
{
auto tmp142 = gc_tileByteOffset(width,cast<int64_t>(4ULL),cast<int64_t>(4ULL),x,y);
int64_t tb = std::get<0>(tmp142);
int64_t in = std::get<1>(tmp142);
uint32_t tileBase = cast<uint32_t>((base + cast<uint32_t>(cast<int64_t>((tb * cast<int64_t>(4ULL))))));
a = gc_Machine_texByte(m,cast<uint32_t>((cast<uint32_t>((tileBase + cast<uint32_t>(cast<int64_t>((in * cast<int64_t>(2ULL)))))) + cast<uint32_t>(0ULL))));
r = gc_Machine_texByte(m,cast<uint32_t>((cast<uint32_t>((tileBase + cast<uint32_t>(cast<int64_t>((in * cast<int64_t>(2ULL)))))) + cast<uint32_t>(1ULL))));
gg = gc_Machine_texByte(m,cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((tileBase + cast<uint32_t>(32ULL))) + cast<uint32_t>(cast<int64_t>((in * cast<int64_t>(2ULL)))))) + cast<uint32_t>(0ULL))));
b = gc_Machine_texByte(m,cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((tileBase + cast<uint32_t>(32ULL))) + cast<uint32_t>(cast<int64_t>((in * cast<int64_t>(2ULL)))))) + cast<uint32_t>(1ULL))));
return {r,gg,b,a};
}
}
// tools/platform/gc/gpu_texture.go:294:1
std::tuple<uint8_t,uint8_t,uint8_t,uint8_t> gc_gpu_decodeCMPR(gc_gpu* g,gc_Machine* m,uint32_t base,int64_t width,int64_t x,int64_t y){
uint8_t r{};
uint8_t gg{};
uint8_t b{};
uint8_t a{};
{
int64_t tilesPerRow = divi<int64_t>((cast<int64_t>((width + cast<int64_t>(7ULL)))),cast<int64_t>(8ULL));
auto tmp143 = std::make_tuple(divi<int64_t>(x,cast<int64_t>(8ULL)),divi<int64_t>(y,cast<int64_t>(8ULL)));
int64_t bx = std::get<0>(tmp143);
int64_t by = std::get<1>(tmp143);
int64_t tileIdx = cast<int64_t>((cast<int64_t>((by * tilesPerRow)) + bx));
auto tmp144 = std::make_tuple(divi<int64_t>((modi<int64_t>(x,cast<int64_t>(8ULL))),cast<int64_t>(4ULL)),divi<int64_t>((modi<int64_t>(y,cast<int64_t>(8ULL))),cast<int64_t>(4ULL)));
int64_t subX = std::get<0>(tmp144);
int64_t subY = std::get<1>(tmp144);
int64_t subIdx = cast<int64_t>((cast<int64_t>((subY * cast<int64_t>(2ULL))) + subX));
uint32_t blockBase = cast<uint32_t>((base + cast<uint32_t>(cast<int64_t>((cast<int64_t>((tileIdx * cast<int64_t>(32ULL))) + cast<int64_t>((subIdx * cast<int64_t>(8ULL))))))));
uint16_t c0 = cast<uint16_t>((shl<uint16_t>(cast<uint16_t>(gc_Machine_texByte(m,blockBase)),cast<int64_t>(8ULL)) | cast<uint16_t>(gc_Machine_texByte(m,cast<uint32_t>((blockBase + cast<uint32_t>(1ULL)))))));
uint16_t c1 = cast<uint16_t>((shl<uint16_t>(cast<uint16_t>(gc_Machine_texByte(m,cast<uint32_t>((blockBase + cast<uint32_t>(2ULL))))),cast<int64_t>(8ULL)) | cast<uint16_t>(gc_Machine_texByte(m,cast<uint32_t>((blockBase + cast<uint32_t>(3ULL)))))));
auto tmp145 = std::make_tuple(modi<int64_t>(x,cast<int64_t>(4ULL)),modi<int64_t>(y,cast<int64_t>(4ULL)));
int64_t ix = std::get<0>(tmp145);
int64_t iy = std::get<1>(tmp145);
uint8_t bits = gc_Machine_texByte(m,cast<uint32_t>((cast<uint32_t>((blockBase + cast<uint32_t>(4ULL))) + cast<uint32_t>(iy))));
uint8_t idx = cast<uint8_t>(((shr<uint8_t>(bits,(cast<uint64_t>((cast<uint64_t>(6ULL) - cast<uint64_t>((cast<uint64_t>(ix) * cast<uint64_t>(2ULL)))))))) & cast<uint8_t>(3ULL)));
auto tmp146 = gc_decodeRGB565(c0);
uint8_t r0 = std::get<0>(tmp146);
uint8_t g0 = std::get<1>(tmp146);
uint8_t b0 = std::get<2>(tmp146);
auto tmp147 = gc_decodeRGB565(c1);
uint8_t r1 = std::get<0>(tmp147);
uint8_t g1 = std::get<1>(tmp147);
uint8_t b1 = std::get<2>(tmp147);
{
switch(idx){
case cast<uint8_t>(0ULL):{
return {r0,g0,b0,cast<uint8_t>(255ULL)};
break;}
case cast<uint8_t>(1ULL):{
return {r1,g1,b1,cast<uint8_t>(255ULL)};
break;}
case cast<uint8_t>(2ULL):{
if ((c0 > c1)) {
return {gc_lerp2(r0,r1),gc_lerp2(g0,g1),gc_lerp2(b0,b1),cast<uint8_t>(255ULL)};
}
return {gc_avg2(r0,r1),gc_avg2(g0,g1),gc_avg2(b0,b1),cast<uint8_t>(255ULL)};
break;}
default:{
if ((c0 > c1)) {
return {gc_lerp2(r1,r0),gc_lerp2(g1,g0),gc_lerp2(b1,b0),cast<uint8_t>(255ULL)};
}
return {cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),cast<uint8_t>(0ULL)};
break;}
}}
}
}
// tools/platform/gc/gpu_texture.go:332:1
std::tuple<uint8_t,uint8_t,uint8_t,uint8_t> gc_decodeRGB565(uint16_t v){
uint8_t r{};
uint8_t g{};
uint8_t b{};
uint8_t a{};
{
return {gc_expand5(shr<uint16_t>(v,cast<int64_t>(11ULL))),gc_expand6(cast<uint16_t>(((shr<uint16_t>(v,cast<int64_t>(5ULL))) & cast<uint16_t>(63ULL)))),gc_expand5(cast<uint16_t>((v & cast<uint16_t>(31ULL)))),cast<uint8_t>(255ULL)};
}
}
// tools/platform/gc/gpu_texture.go:336:1
std::tuple<uint8_t,uint8_t,uint8_t,uint8_t> gc_decodeRGB5A3(uint16_t v){
uint8_t r{};
uint8_t g{};
uint8_t b{};
uint8_t a{};
{
if ((cast<uint16_t>((v & cast<uint16_t>(32768ULL))) != cast<uint16_t>(0ULL))) {
return {gc_expand5(cast<uint16_t>(((shr<uint16_t>(v,cast<int64_t>(10ULL))) & cast<uint16_t>(31ULL)))),gc_expand5(cast<uint16_t>(((shr<uint16_t>(v,cast<int64_t>(5ULL))) & cast<uint16_t>(31ULL)))),gc_expand5(cast<uint16_t>((v & cast<uint16_t>(31ULL)))),cast<uint8_t>(255ULL)};
}
return {gc_expand4(cast<uint16_t>(((shr<uint16_t>(v,cast<int64_t>(8ULL))) & cast<uint16_t>(15ULL)))),gc_expand4(cast<uint16_t>(((shr<uint16_t>(v,cast<int64_t>(4ULL))) & cast<uint16_t>(15ULL)))),gc_expand4(cast<uint16_t>((v & cast<uint16_t>(15ULL)))),gc_expand3(cast<uint16_t>(((shr<uint16_t>(v,cast<int64_t>(12ULL))) & cast<uint16_t>(7ULL))))};
}
}
// tools/platform/gc/gpu_texture.go:343:1
uint8_t gc_expand3(uint16_t v){
{
return cast<uint8_t>(cast<uint16_t>((cast<uint16_t>((shl<uint16_t>(v,cast<int64_t>(5ULL)) | shl<uint16_t>(v,cast<int64_t>(2ULL)))) | shr<uint16_t>(v,cast<int64_t>(1ULL)))));
}
}
// tools/platform/gc/gpu_texture.go:347:1
uint8_t gc_lerp2(uint8_t a,uint8_t b){
{
return cast<uint8_t>(divi<uint16_t>((cast<uint16_t>((cast<uint16_t>((cast<uint16_t>(a) * cast<uint16_t>(2ULL))) + cast<uint16_t>(b)))),cast<uint16_t>(3ULL)));
}
}
// tools/platform/gc/gpu_texture.go:348:1
uint8_t gc_avg2(uint8_t a,uint8_t b){
{
return cast<uint8_t>(divi<uint16_t>((cast<uint16_t>((cast<uint16_t>(a) + cast<uint16_t>(b)))),cast<uint16_t>(2ULL)));
}
}
// tools/platform/gc/gpu_texture.go:353:1
uint8_t gc_Machine_texByte(gc_Machine* m,uint32_t addr){
{
uint32_t a = gc_phys(addr);
if ((cast<int64_t>(a) >= len(m->RAM))) {
return cast<uint8_t>(0ULL);
}
return m->RAM[a];
}
}
// tools/platform/gc/gpu_texture.go:368:1
bool gc_Machine_TextureBound(gc_Machine* m,int64_t i){
{
return (gc_gpu_texSetup(&(m->gpu),i).base != cast<uint32_t>(0ULL));
}
}
// tools/platform/gc/gpu_texture.go:376:1
std::tuple<image_RGBA*,Error> gc_Machine_DumpTexture(gc_Machine* m,int64_t i){
{
gc_texState tx = gc_gpu_texSetup(&(m->gpu),i);
if (((tx.width <= cast<int64_t>(0ULL)) || (tx.height <= cast<int64_t>(0ULL)))) {
return {{},go_fmt_Errorf(std::string("texture %d has no size (the game has not bound it)",50),i)};
}
image_RGBA* img = go_image_NewRGBA(go_image_Rect(cast<int64_t>(0ULL),cast<int64_t>(0ULL),tx.width,tx.height));
{int64_t y = cast<int64_t>(0ULL);for (;(y < tx.height);y++){
{int64_t x = cast<int64_t>(0ULL);for (;(x < tx.width);x++){
auto tmp148 = gc_gpu_decodeTexel(&(m->gpu),m,tx,x,y);
uint8_t r = std::get<0>(tmp148);
uint8_t gg = std::get<1>(tmp148);
uint8_t b = std::get<2>(tmp148);
uint8_t a = std::get<3>(tmp148);
int64_t o = image_RGBA_PixOffset(img,x,y);
auto tmp149 = std::make_tuple(r,gg,b,a);
img->Pix[o] = std::get<0>(tmp149);
img->Pix[cast<int64_t>((o + cast<int64_t>(1ULL)))] = std::get<1>(tmp149);
img->Pix[cast<int64_t>((o + cast<int64_t>(2ULL)))] = std::get<2>(tmp149);
img->Pix[cast<int64_t>((o + cast<int64_t>(3ULL)))] = std::get<3>(tmp149);
}
}}
}return {img,{}};
}
}
// tools/platform/gc/gpu_vtx.go:25:1
int64_t gc_componentBytes(uint32_t format){
{
{
switch(format){
case cast<uint32_t>(0ULL):case cast<uint32_t>(1ULL):{
return cast<int64_t>(1ULL);
break;}
case cast<uint32_t>(2ULL):case cast<uint32_t>(3ULL):{
return cast<int64_t>(2ULL);
break;}
case cast<uint32_t>(4ULL):{
return cast<int64_t>(4ULL);
break;}
}}
return cast<int64_t>(0ULL);
}
}
// tools/platform/gc/gpu_vtx.go:38:1
int64_t gc_colorBytes(uint32_t comp){
{
{
switch(comp){
case cast<uint32_t>(0ULL):{
return cast<int64_t>(2ULL);
break;}
case cast<uint32_t>(1ULL):{
return cast<int64_t>(3ULL);
break;}
case cast<uint32_t>(2ULL):{
return cast<int64_t>(4ULL);
break;}
case cast<uint32_t>(3ULL):{
return cast<int64_t>(2ULL);
break;}
case cast<uint32_t>(4ULL):{
return cast<int64_t>(3ULL);
break;}
case cast<uint32_t>(5ULL):{
return cast<int64_t>(4ULL);
break;}
}}
return cast<int64_t>(0ULL);
}
}
// tools/platform/gc/gpu_vtx.go:68:1
int64_t gc_normalIndexCount(uint32_t g0){
{
if (((cast<uint32_t>(((shr<uint32_t>(g0,cast<int64_t>(9ULL))) & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL)) && (cast<uint32_t>(((shr<uint32_t>(g0,cast<int64_t>(31ULL))) & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL)))) {
return cast<int64_t>(3ULL);
}
return cast<int64_t>(1ULL);
}
}
// tools/platform/gc/gpu_vtx.go:78:1
int64_t gc_gpu_vertexSize(gc_gpu* g,int64_t vat){
{
uint32_t lo = g->CPReg[cast<int64_t>(80ULL)];
uint32_t hi = g->CPReg[cast<int64_t>(96ULL)];
uint32_t g0 = g->CPReg[cast<uint32_t>((cast<uint32_t>(112ULL) + cast<uint32_t>(vat)))];
uint32_t g1 = g->CPReg[cast<uint32_t>((cast<uint32_t>(128ULL) + cast<uint32_t>(vat)))];
uint32_t g2 = g->CPReg[cast<uint32_t>((cast<uint32_t>(144ULL) + cast<uint32_t>(vat)))];
int64_t size = cast<int64_t>(0ULL);
{uint32_t b = cast<uint32_t>(0ULL);for (;(b < cast<uint32_t>(9ULL));b++){
if ((cast<uint32_t>((lo & (shl<uint32_t>(cast<uint32_t>(1ULL),b)))) != cast<uint32_t>(0ULL))) {
size++;
}
}
}auto addN = [&](uint32_t desc,int64_t direct,int64_t nIdx)->void{
{
switch(desc){
case cast<uint32_t>(2ULL):{
size += nIdx;
break;}
case cast<uint32_t>(3ULL):{
size += cast<int64_t>((cast<int64_t>(2ULL) * nIdx));
break;}
case cast<uint32_t>(1ULL):{
size += direct;
break;}
}}
}
;
auto add = [&](uint32_t desc,int64_t direct)->void{
addN(desc,direct,cast<int64_t>(1ULL));
}
;
int64_t posComps = cast<int64_t>(2ULL);
if ((cast<uint32_t>((g0 & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL))) {
posComps = cast<int64_t>(3ULL);
}
add(cast<uint32_t>(((shr<uint32_t>(lo,cast<int64_t>(9ULL))) & cast<uint32_t>(3ULL))),cast<int64_t>((posComps * gc_componentBytes(cast<uint32_t>(((shr<uint32_t>(g0,cast<int64_t>(1ULL))) & cast<uint32_t>(7ULL)))))));
int64_t nrmComps = cast<int64_t>(3ULL);
if ((cast<uint32_t>(((shr<uint32_t>(g0,cast<int64_t>(9ULL))) & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL))) {
nrmComps = cast<int64_t>(9ULL);
}
addN(cast<uint32_t>(((shr<uint32_t>(lo,cast<int64_t>(11ULL))) & cast<uint32_t>(3ULL))),cast<int64_t>((nrmComps * gc_componentBytes(cast<uint32_t>(((shr<uint32_t>(g0,cast<int64_t>(10ULL))) & cast<uint32_t>(7ULL)))))),gc_normalIndexCount(g0));
add(cast<uint32_t>(((shr<uint32_t>(lo,cast<int64_t>(13ULL))) & cast<uint32_t>(3ULL))),gc_colorBytes(cast<uint32_t>(((shr<uint32_t>(g0,cast<int64_t>(14ULL))) & cast<uint32_t>(7ULL)))));
add(cast<uint32_t>(((shr<uint32_t>(lo,cast<int64_t>(15ULL))) & cast<uint32_t>(3ULL))),gc_colorBytes(cast<uint32_t>(((shr<uint32_t>(g0,cast<int64_t>(18ULL))) & cast<uint32_t>(7ULL)))));
auto texBytes = [&](uint32_t elem,uint32_t format)->int64_t{
int64_t comps = cast<int64_t>(1ULL);
if ((elem != cast<uint32_t>(0ULL))) {
comps = cast<int64_t>(2ULL);
}
return cast<int64_t>((comps * gc_componentBytes(format)));
}
;
auto texDesc = [&](int64_t k)->uint32_t{
return cast<uint32_t>(((shr<uint32_t>(hi,(cast<uint32_t>((cast<uint32_t>(2ULL) * cast<uint32_t>(k)))))) & cast<uint32_t>(3ULL)));
}
;
add(texDesc(cast<int64_t>(0ULL)),texBytes(cast<uint32_t>(((shr<uint32_t>(g0,cast<int64_t>(21ULL))) & cast<uint32_t>(1ULL))),cast<uint32_t>(((shr<uint32_t>(g0,cast<int64_t>(22ULL))) & cast<uint32_t>(7ULL)))));
add(texDesc(cast<int64_t>(1ULL)),texBytes(cast<uint32_t>(((shr<uint32_t>(g1,cast<int64_t>(0ULL))) & cast<uint32_t>(1ULL))),cast<uint32_t>(((shr<uint32_t>(g1,cast<int64_t>(1ULL))) & cast<uint32_t>(7ULL)))));
add(texDesc(cast<int64_t>(2ULL)),texBytes(cast<uint32_t>(((shr<uint32_t>(g1,cast<int64_t>(9ULL))) & cast<uint32_t>(1ULL))),cast<uint32_t>(((shr<uint32_t>(g1,cast<int64_t>(10ULL))) & cast<uint32_t>(7ULL)))));
add(texDesc(cast<int64_t>(3ULL)),texBytes(cast<uint32_t>(((shr<uint32_t>(g1,cast<int64_t>(18ULL))) & cast<uint32_t>(1ULL))),cast<uint32_t>(((shr<uint32_t>(g1,cast<int64_t>(19ULL))) & cast<uint32_t>(7ULL)))));
add(texDesc(cast<int64_t>(4ULL)),texBytes(cast<uint32_t>(((shr<uint32_t>(g1,cast<int64_t>(27ULL))) & cast<uint32_t>(1ULL))),cast<uint32_t>(((shr<uint32_t>(g1,cast<int64_t>(28ULL))) & cast<uint32_t>(7ULL)))));
add(texDesc(cast<int64_t>(5ULL)),texBytes(cast<uint32_t>(((shr<uint32_t>(g2,cast<int64_t>(5ULL))) & cast<uint32_t>(1ULL))),cast<uint32_t>(((shr<uint32_t>(g2,cast<int64_t>(6ULL))) & cast<uint32_t>(7ULL)))));
add(texDesc(cast<int64_t>(6ULL)),texBytes(cast<uint32_t>(((shr<uint32_t>(g2,cast<int64_t>(14ULL))) & cast<uint32_t>(1ULL))),cast<uint32_t>(((shr<uint32_t>(g2,cast<int64_t>(15ULL))) & cast<uint32_t>(7ULL)))));
add(texDesc(cast<int64_t>(7ULL)),texBytes(cast<uint32_t>(((shr<uint32_t>(g2,cast<int64_t>(23ULL))) & cast<uint32_t>(1ULL))),cast<uint32_t>(((shr<uint32_t>(g2,cast<int64_t>(24ULL))) & cast<uint32_t>(7ULL)))));
return size;
}
}
// tools/platform/gc/gpu_xf.go:19:1
float gc_gpu_xfFloat(gc_gpu* g,int64_t addr){
{
if (((addr < cast<int64_t>(0ULL)) || (addr >= cast<int64_t>(4192ULL)))) {
return cast<float>(0ULL);
}
return go_math_Float32frombits(g->XFMem[addr]);
}
}
// tools/platform/gc/gpu_xf.go:68:1
gc_clipVertex gc_lerpClip(gc_clipVertex a,gc_clipVertex b,float t){
{
auto li = [&](uint8_t x,uint8_t y)->uint8_t{
return cast<uint8_t>(((cast<float>(x) + (((cast<float>(y) - cast<float>(x))) * t)) + 0.5));
}
;
auto lf = [&](float x,float y)->float{
return (x + (((y - x)) * t));
}
;
gc_clipVertex out = gc_clipVertex{cast<float>(lf(a.cx,b.cx)),cast<float>(lf(a.cy,b.cy)),cast<float>(lf(a.cz,b.cz)),cast<float>(lf(a.cw,b.cw)),{},{},a.ntc};
{int64_t c = cast<int64_t>(0ULL);for (;(c < cast<int64_t>(2ULL));c++){
{int64_t k = cast<int64_t>(0ULL);for (;(k < cast<int64_t>(4ULL));k++){
out.col[c][k] = li(a.col[c][k],b.col[c][k]);
}
}}
}{int64_t i = cast<int64_t>(0ULL);for (;(i < a.ntc);i++){
out.tc[i] = gc_texCoord{cast<float>(lf(a.tc[i].s,b.tc[i].s)),cast<float>(lf(a.tc[i].t,b.tc[i].t)),cast<float>(lf(a.tc[i].q,b.tc[i].q))};
}
}return out;
}
}
// tools/platform/gc/gpu_xf.go:94:1
std::tuple<float,float,float> gc_gpu_transform(gc_gpu* g,int64_t mtxIdx,float mx,float my,float mz){
float sx{};
float sy{};
float sz{};
{
auto tmp150 = gc_gpu_eyePos(g,mtxIdx,mx,my,mz);
float ex = std::get<0>(tmp150);
float ey = std::get<1>(tmp150);
float ez = std::get<2>(tmp150);
return gc_gpu_project(g,ex,ey,ez);
}
}
// tools/platform/gc/gpu_xf.go:102:1
std::tuple<float,float,float> gc_gpu_eyePos(gc_gpu* g,int64_t mtxIdx,float mx,float my,float mz){
float ex{};
float ey{};
float ez{};
{
int64_t base = cast<int64_t>((mtxIdx * cast<int64_t>(4ULL)));
ex = ((((gc_gpu_xfFloat(g,cast<int64_t>((base + cast<int64_t>(0ULL)))) * mx) + (gc_gpu_xfFloat(g,cast<int64_t>((base + cast<int64_t>(1ULL)))) * my)) + (gc_gpu_xfFloat(g,cast<int64_t>((base + cast<int64_t>(2ULL)))) * mz)) + gc_gpu_xfFloat(g,cast<int64_t>((base + cast<int64_t>(3ULL)))));
ey = ((((gc_gpu_xfFloat(g,cast<int64_t>((base + cast<int64_t>(4ULL)))) * mx) + (gc_gpu_xfFloat(g,cast<int64_t>((base + cast<int64_t>(5ULL)))) * my)) + (gc_gpu_xfFloat(g,cast<int64_t>((base + cast<int64_t>(6ULL)))) * mz)) + gc_gpu_xfFloat(g,cast<int64_t>((base + cast<int64_t>(7ULL)))));
ez = ((((gc_gpu_xfFloat(g,cast<int64_t>((base + cast<int64_t>(8ULL)))) * mx) + (gc_gpu_xfFloat(g,cast<int64_t>((base + cast<int64_t>(9ULL)))) * my)) + (gc_gpu_xfFloat(g,cast<int64_t>((base + cast<int64_t>(10ULL)))) * mz)) + gc_gpu_xfFloat(g,cast<int64_t>((base + cast<int64_t>(11ULL)))));
return {ex,ey,ez};
}
}
// tools/platform/gc/gpu_xf.go:113:1
std::tuple<float,float,float> gc_gpu_project(gc_gpu* g,float ex,float ey,float ez){
float sx{};
float sy{};
float sz{};
{
auto tmp151 = gc_gpu_clipPos(g,ex,ey,ez);
float cx = std::get<0>(tmp151);
float cy = std::get<1>(tmp151);
float cz = std::get<2>(tmp151);
float cw = std::get<3>(tmp151);
return gc_gpu_toScreen(g,cx,cy,cz,cw);
}
}
// tools/platform/gc/gpu_xf.go:128:1
std::tuple<float,float,float,float> gc_gpu_clipPos(gc_gpu* g,float ex,float ey,float ez){
float cx{};
float cy{};
float cz{};
float cw{};
{
float p0 = gc_gpu_xfFloat(g,cast<int64_t>(4128ULL));
float p1 = gc_gpu_xfFloat(g,cast<int64_t>(4129ULL));
float p2 = gc_gpu_xfFloat(g,cast<int64_t>(4130ULL));
float p3 = gc_gpu_xfFloat(g,cast<int64_t>(4131ULL));
float p4 = gc_gpu_xfFloat(g,cast<int64_t>(4132ULL));
float p5 = gc_gpu_xfFloat(g,cast<int64_t>(4133ULL));
if ((g->XFMem[cast<int64_t>(4134ULL)] == cast<uint32_t>(0ULL))) {
cx = ((p0 * ex) + (p1 * ez));
cy = ((p2 * ey) + (p3 * ez));
cz = ((p4 * ez) + p5);
cw = cast<float>(-ez);
}
else {
cx = ((p0 * ex) + p1);
cy = ((p2 * ey) + p3);
cz = ((p4 * ez) + p5);
cw = cast<float>(1ULL);
}
return {cx,cy,cz,cw};
}
}
// tools/platform/gc/gpu_xf.go:152:1
std::tuple<float,float,float> gc_gpu_toScreen(gc_gpu* g,float cx,float cy,float cz,float cw){
float sx{};
float sy{};
float sz{};
{
if ((cw == cast<float>(0ULL))) {
cw = cast<float>(1ULL);
}
float nx = (cx / cw);
float ny = (cy / cw);
float nz = (cz / cw);
float vsx = gc_gpu_xfFloat(g,cast<int64_t>(4122ULL));
float vsy = gc_gpu_xfFloat(g,cast<int64_t>(4123ULL));
float vsz = gc_gpu_xfFloat(g,cast<int64_t>(4124ULL));
float vox = gc_gpu_xfFloat(g,cast<int64_t>(4125ULL));
float voy = gc_gpu_xfFloat(g,cast<int64_t>(4126ULL));
float voz = gc_gpu_xfFloat(g,cast<int64_t>(4127ULL));
sx = ((nx * vsx) + ((vox - cast<float>(342ULL))));
sy = ((ny * vsy) + ((voy - cast<float>(342ULL))));
sz = ((nz * vsz) + voz);
return {sx,sy,sz};
}
}
// tools/platform/gc/idle.go:87:1
gc_idleSnap gc_Machine_snapshotCPU(gc_Machine* m){
{
gekko_CPU* c = m->CPU;
gc_idleSnap s = gc_idleSnap{c->PC,c->LR,c->CTR,c->CR,c->XER,c->MSR,c->GPR,{}};
{auto&& tmp152 = c->FPR;
for(int64_t tmp153=0;tmp153<len(tmp152);++tmp153){
auto i=tmp153;s.FPR[i][cast<int64_t>(0ULL)] = go_math_Float64bits(c->FPR[i].PS0);
s.FPR[i][cast<int64_t>(1ULL)] = go_math_Float64bits(c->FPR[i].PS1);
}}
return s;
}
}
// tools/platform/gc/idle.go:104:1
bool gc_Machine_idleStep(gc_Machine* m,uint32_t pc){
{
if ((!m->idle.armed)) {
if ((m->Instrs < m->idle.next)) {
return false;
}
m->idle.armed = true;
m->idle.insns = cast<int64_t>(0ULL);
m->idle.stores = m->stores;
m->idle.snap = gc_Machine_snapshotCPU(m);
return false;
}
m->idle.insns++;
if ((m->idle.insns > cast<int64_t>(64ULL))) {
m->idle.armed = false;
m->idle.next = cast<uint64_t>((m->Instrs + cast<uint64_t>(4096ULL)));
return false;
}
if (((pc != m->idle.snap.PC) || (m->stores != m->idle.stores))) {
return false;
}
if ((gc_Machine_snapshotCPU(m) != m->idle.snap)) {
return false;
}
m->idle.period = m->idle.insns;
m->idle.armed = false;
m->idle.next = cast<uint64_t>((m->Instrs + cast<uint64_t>(4096ULL)));
return true;
}
}
// tools/platform/gc/idle.go:140:1
uint64_t gc_Machine_idleDeadline(gc_Machine* m){
{
gc_dsp* d = (&m->dsp);
if ((((bool(d->Core) && (!d->CoreHalt)) && (!d->CoreBlocked)) && (!d->Core->Halted))) {
return cast<uint64_t>(0ULL);
}
uint64_t n = cast<uint64_t>(cast<uint32_t>((cast<uint32_t>(8100000ULL) - m->vi.Counter)));
if (((cast<uint16_t>((d->AIDControl & cast<uint16_t>(32768ULL))) != cast<uint16_t>(0ULL)) && (cast<uint16_t>((d->AIDControl & cast<uint16_t>(32767ULL))) != cast<uint16_t>(0ULL)))) {
{
uint64_t left = cast<uint64_t>((cast<uint64_t>(81000ULL) - d->AIDAccum));
if ((left < n)) {
n = left;
}
}
}
if (((m->di.BusyInstr > cast<int64_t>(0ULL)) && (cast<uint64_t>(m->di.BusyInstr) < n))) {
n = cast<uint64_t>(m->di.BusyInstr);
}
{
uint64_t dec = gekko_CPU_InstrsToDecUnderflow(m->CPU);
if ((dec < n)) {
n = dec;
}
}
if ((n == cast<uint64_t>(0ULL))) {
return cast<uint64_t>(0ULL);
}
return cast<uint64_t>((n - cast<uint64_t>(1ULL)));
}
}
// tools/platform/gc/idle.go:187:1
void gc_Machine_idleSkip(gc_Machine* m,uint64_t n){
{
if ((m->idle.period <= cast<int64_t>(0ULL))) {
return ;
}
n -= modi<uint64_t>(n,cast<uint64_t>(m->idle.period));
if ((n == cast<uint64_t>(0ULL))) {
return ;
}
m->vi.Counter += cast<uint32_t>(n);
{
gc_dsp* d = (&m->dsp);
if (((cast<uint16_t>((d->AIDControl & cast<uint16_t>(32768ULL))) != cast<uint16_t>(0ULL)) && (cast<uint16_t>((d->AIDControl & cast<uint16_t>(32767ULL))) != cast<uint16_t>(0ULL)))) {
d->AIDAccum += n;
}
}
if ((m->di.BusyInstr > cast<int64_t>(0ULL))) {
m->di.BusyInstr -= cast<int64_t>(n);
}
gekko_CPU_SkipInstructions(m->CPU,n);
m->Instrs += n;
m->idle.Skipped += n;
m->idle.Hits++;
}
}
// tools/platform/gc/idle.go:210:1
std::tuple<uint64_t,uint64_t> gc_Machine_IdleStats(gc_Machine* m){
uint64_t skipped{};
uint64_t hits{};
{
return {m->idle.Skipped,m->idle.Hits};
}
}
// tools/platform/gc/idle.go:214:1
void gc_Machine_SetIdleSkip(gc_Machine* m,bool on){
{
m->noIdle = (!on);
}
}
// tools/platform/gc/ipl.go:54:1
void gc_Machine_setupLowMem(gc_Machine* m){
{
auto tmp154 = gc_Disc_Read(m->disc,cast<int64_t>(0ULL),cast<int64_t>(32ULL));
Slice<uint8_t> boot = std::get<0>(tmp154);
gc_Machine_dmaToRAM(m,cast<uint32_t>(0ULL),boot);
auto set = [&](uint32_t addr,uint32_t v)->void{
gc_Machine_setRAM32(m,addr,v);
}
;
set(cast<uint32_t>(32ULL),cast<uint32_t>(219540062ULL));
set(cast<uint32_t>(36ULL),cast<uint32_t>(1ULL));
set(cast<uint32_t>(40ULL),cast<uint32_t>(25165824ULL));
set(cast<uint32_t>(44ULL),cast<uint32_t>(3ULL));
set(cast<uint32_t>(48ULL),cast<uint32_t>(0ULL));
set(cast<uint32_t>(52ULL),cast<uint32_t>(2172643520ULL));
set(cast<uint32_t>(56ULL),m->disc->Header.FSTAddr);
set(cast<uint32_t>(60ULL),m->disc->Header.FSTMaxSize);
set(cast<uint32_t>(204ULL),cast<uint32_t>(0ULL));
set(cast<uint32_t>(240ULL),cast<uint32_t>(25165824ULL));
set(cast<uint32_t>(248ULL),cast<uint32_t>(162000000ULL));
set(cast<uint32_t>(252ULL),cast<uint32_t>(486000000ULL));
set(cast<uint32_t>(244ULL),cast<uint32_t>(0ULL));
}
}
// tools/platform/gc/ipl.go:92:1
void gc_Machine_setupState(gc_Machine* m){
{
gekko_CPU* c = m->CPU;
constexpr int64_t bl256=8188ULL;
c->DBAT[cast<int64_t>(0ULL)] = std::array<uint32_t,2>{cast<uint32_t>(2147491838ULL),cast<uint32_t>(2ULL)};
c->DBAT[cast<int64_t>(1ULL)] = std::array<uint32_t,2>{cast<uint32_t>(3221233662ULL),cast<uint32_t>(18ULL)};
c->IBAT[cast<int64_t>(0ULL)] = std::array<uint32_t,2>{cast<uint32_t>(2147491838ULL),cast<uint32_t>(2ULL)};
c->MSR = cast<uint32_t>(12336ULL);
c->GPR[cast<int64_t>(1ULL)] = cast<uint32_t>(2171600640ULL);
c->GPR[cast<int64_t>(2ULL)] = cast<uint32_t>(2170552320ULL);
c->GPR[cast<int64_t>(13ULL)] = cast<uint32_t>(2169503744ULL);
c->SC = [=](auto...args){return gc_Machine_handleSyscall(m,args...);};
}
}
// tools/platform/gc/ipl.go:128:1
bool gc_Machine_handleSyscall(gc_Machine* m,gekko_CPU* c){
{
uint32_t vec = cast<uint32_t>(3072ULL);
if ((cast<uint32_t>((c->MSR & cast<uint32_t>(64ULL))) != cast<uint32_t>(0ULL))) {
vec = cast<uint32_t>(4293921792ULL);
}
if ((gc_Machine_ram32(m,vec) == cast<uint32_t>(0ULL))) {
return true;
}
return false;
}
}
// tools/platform/gc/ipl.go:142:1
std::tuple<uint32_t,Error> gc_Machine_RunApploader(gc_Machine* m){
uint32_t entry{};
Error err{};
{
gc_Machine_setupLowMem(m);
gc_Machine_setupState(m);
auto tmp155 = gc_Disc_ApploaderCode(m->disc);
Slice<uint8_t> code = std::get<0>(tmp155);
err = std::get<1>(tmp155);
if (bool(err)) {
return {cast<uint32_t>(0ULL),err};
}
gc_Machine_dmaToRAM(m,cast<uint32_t>(18874368ULL),code);
constexpr int64_t pInit=2167406592ULL;
constexpr int64_t pMain=2167406596ULL;
constexpr int64_t pClose=2167406600ULL;
gekko_CPU* c = m->CPU;
auto tmp156 = std::make_tuple(cast<uint32_t>(2167406592ULL),cast<uint32_t>(2167406596ULL),cast<uint32_t>(2167406600ULL));
c->GPR[cast<int64_t>(3ULL)] = std::get<0>(tmp156);
c->GPR[cast<int64_t>(4ULL)] = std::get<1>(tmp156);
c->GPR[cast<int64_t>(5ULL)] = std::get<2>(tmp156);
{
Error err = gc_Machine_call(m,m->disc->Apploader.Entry);
if (bool(err)) {
return {cast<uint32_t>(0ULL),go_fmt_Errorf(std::string("apploader entry: %w",19),err)};
}
}
uint32_t initFn = gc_Machine_ram32(m,cast<uint32_t>(2167406592ULL));
uint32_t mainFn = gc_Machine_ram32(m,cast<uint32_t>(2167406596ULL));
uint32_t closeFn = gc_Machine_ram32(m,cast<uint32_t>(2167406600ULL));
if ((((initFn == cast<uint32_t>(0ULL)) || (mainFn == cast<uint32_t>(0ULL))) || (closeFn == cast<uint32_t>(0ULL)))) {
return {cast<uint32_t>(0ULL),go_fmt_Errorf(std::string("the apploader did not report its functions (init=0x%08X main=0x%08X close=0x%08X)",81),initFn,mainFn,closeFn)};
}
c->GPR[cast<int64_t>(3ULL)] = cast<uint32_t>(2167603200ULL);
{
Error err = gc_Machine_call(m,initFn);
if (bool(err)) {
return {cast<uint32_t>(0ULL),go_fmt_Errorf(std::string("apploader init: %w",18),err)};
}
}
constexpr int64_t pDst=2167406604ULL;
constexpr int64_t pSize=2167406608ULL;
constexpr int64_t pOff=2167406612ULL;
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(1000ULL));i++){
auto tmp157 = std::make_tuple(cast<uint32_t>(2167406604ULL),cast<uint32_t>(2167406608ULL),cast<uint32_t>(2167406612ULL));
c->GPR[cast<int64_t>(3ULL)] = std::get<0>(tmp157);
c->GPR[cast<int64_t>(4ULL)] = std::get<1>(tmp157);
c->GPR[cast<int64_t>(5ULL)] = std::get<2>(tmp157);
{
Error err = gc_Machine_call(m,mainFn);
if (bool(err)) {
return {cast<uint32_t>(0ULL),go_fmt_Errorf(std::string("apploader main (iteration %d): %w",33),i,err)};
}
}
if ((c->GPR[cast<int64_t>(3ULL)] == cast<uint32_t>(0ULL))) {
break;
}
uint32_t dst = gc_Machine_ram32(m,cast<uint32_t>(2167406604ULL));
uint32_t size = gc_Machine_ram32(m,cast<uint32_t>(2167406608ULL));
uint32_t off = gc_Machine_ram32(m,cast<uint32_t>(2167406612ULL));
if (bool(m->OnDVDRead)) {
m->OnDVDRead(cast<int64_t>(off),size,dst);
}
auto tmp158 = gc_Disc_Read(m->disc,cast<int64_t>(off),cast<int64_t>(size));
Slice<uint8_t> data = std::get<0>(tmp158);
Error rerr = std::get<1>(tmp158);
if (bool(rerr)) {
return {cast<uint32_t>(0ULL),go_fmt_Errorf(std::string("apploader read (offset 0x%X size %d): %w",40),off,size,rerr)};
}
gc_Machine_dmaToRAM(m,cast<uint32_t>((dst & cast<uint32_t>(67108863ULL))),data);
}
}{
Error err = gc_Machine_call(m,closeFn);
if (bool(err)) {
return {cast<uint32_t>(0ULL),go_fmt_Errorf(std::string("apploader close: %w",19),err)};
}
}
entry = c->GPR[cast<int64_t>(3ULL)];
c->PC = entry;
return {entry,{}};
}
}
// tools/platform/gc/ipl.go:221:1
Error gc_Machine_call(gc_Machine* m,uint32_t fn){
{
gekko_CPU* c = m->CPU;
c->LR = cast<uint32_t>(2167668736ULL);
c->PC = fn;
constexpr int64_t budget=200000000ULL;
bool trace = gc_iplTrace;
std::array<uint32_t,24> ring={};
int64_t ri = cast<int64_t>(0ULL);
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(200000000ULL));i++){
{
switch(c->PC){
case cast<uint32_t>(2167668736ULL):{
return {};
break;}
case cast<uint32_t>(2167603200ULL):{
gc_Machine_serviceReport(m);
continue;
break;}
}}
if (trace) {
ring[modi<int64_t>(ri,cast<int64_t>(24ULL))] = c->PC;
ri++;
if ((c->PC < cast<uint32_t>(2147483648ULL))) {
go_fmt_Fprintf(go_os_Stderr,std::string("ipl: control left the apploader for 0x%08X. Last PCs:\012",54),c->PC);
{int64_t k = cast<int64_t>(0ULL);for (;(k < cast<int64_t>(24ULL));k++){
go_fmt_Fprintf(go_os_Stderr,std::string("    0x%08X\012",11),ring[modi<int64_t>((cast<int64_t>((ri + k))),cast<int64_t>(24ULL))]);
}
}return go_fmt_Errorf(std::string("apploader jumped to 0x%08X (SRR0=0x%08X SRR1=0x%08X)",52),c->PC,c->SRR0,c->SRR1);
}
}
gc_Machine_tickVI(m);
gekko_CPU_Step(c);
if (c->Halted) {
return go_fmt_Errorf(std::string("halted at 0x%08X: %s",20),c->PC,c->HaltReason);
}
}
}return go_fmt_Errorf(std::string("did not return within %d instructions (PC 0x%08X)",49),cast<int64_t>(200000000ULL),c->PC);
}
}
// tools/platform/gc/ipl.go:263:1
void gc_Machine_serviceReport(gc_Machine* m){
{
gekko_CPU* c = m->CPU;
std::string msg = gc_Machine_readCString(m,c->GPR[cast<int64_t>(3ULL)]);
if ((msg != std::string("",0))) {
gc_Machine_logf(m,std::string("apploader: %s",13),gc_trimReport(msg));
}
c->PC = c->LR;
}
}
// tools/platform/gc/ipl.go:274:1
std::string gc_Machine_readCString(gc_Machine* m,uint32_t addr){
{
Slice<uint8_t> b={};
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < cast<uint32_t>(256ULL));i++){
uint8_t ch = gekko_CPU_ReadMem(m->CPU,cast<uint32_t>((addr + i)));
if ((ch == cast<uint8_t>(0ULL))) {
break;
}
b = append(b,Slice<uint8_t>{ch});
}
}return cast<std::string>(b);
}
}
// tools/platform/gc/ipl.go:286:1
std::string gc_trimReport(std::string s){
{
{;for (;((len(s) > cast<int64_t>(0ULL)) && (((cast<uint8_t>(s[cast<int64_t>((len(s) - cast<int64_t>(1ULL)))]) == cast<uint8_t>(10ULL)) || (cast<uint8_t>(s[cast<int64_t>((len(s) - cast<int64_t>(1ULL)))]) == cast<uint8_t>(13ULL)))));){
s = sub(s,0,cast<int64_t>((len(s) - cast<int64_t>(1ULL))));
}
}return s;
}
}
// tools/platform/gc/machine.go:165:1
std::tuple<gc_Machine*,Error> gc_NewMachine(gc_Disc* disc){
{
gc_Machine* m = arenaNew(gc_Machine{Slice<uint8_t>::make(cast<int64_t>(25165824ULL)),Slice<uint8_t>::make(cast<int64_t>(16777216ULL)),{},disc,{},{},{},{},{},{},{},{},{},{},{},{},{},{},Map<std::string,bool>{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{}});
if (bool(disc)) {
auto tmp159 = gc_Disc_MD5(disc);
std::string md5 = std::get<0>(tmp159);
Error err = std::get<1>(tmp159);
if (bool(err)) {
return {{},err};
}
m->discMD5 = md5;
}
m->CPU = gekko_NewCPU(m);
gc_di_init(&(m->di));
gc_dsp_init(&(m->dsp));
gc_exi_init(&(m->exi));
gc_vi_init(&(m->vi));
gc_gpu_init(&(m->gpu));
gc_si_connectPad(&(m->si),cast<int64_t>(0ULL));
return {m,{}};
}
}
// tools/platform/gc/machine.go:192:1
gc_Disc* gc_Machine_Disc(gc_Machine* m){
{
return m->disc;
}
}
// tools/platform/gc/machine.go:209:1
uint8_t gc_Machine_Read8(gc_Machine* m,uint32_t a){
{
if ((a < cast<uint32_t>(25165824ULL))) {
uint8_t v = m->RAM[a];
gc_Machine_readWatch(m,a,cast<uint32_t>(v));
return v;
}
if (((a >= cast<uint32_t>(201326592ULL)) && (a < cast<uint32_t>(201392128ULL)))) {
return cast<uint8_t>(gc_Machine_regRead(m,a,cast<int64_t>(1ULL)));
}
gc_Machine_logf(m,std::string("read8 unmapped 0x%08X",21),a);
return cast<uint8_t>(0ULL);
}
}
// tools/platform/gc/machine.go:222:1
uint16_t gc_Machine_Read16(gc_Machine* m,uint32_t a){
{
if ((cast<uint32_t>((a + cast<uint32_t>(1ULL))) < cast<uint32_t>(25165824ULL))) {
uint16_t v = be_Uint16(sub(m->RAM,a,cast<uint32_t>((a + cast<uint32_t>(2ULL)))));
gc_Machine_readWatch(m,a,cast<uint32_t>(v));
return v;
}
if (((a >= cast<uint32_t>(201326592ULL)) && (a < cast<uint32_t>(201392128ULL)))) {
return cast<uint16_t>(gc_Machine_regRead(m,a,cast<int64_t>(2ULL)));
}
gc_Machine_logf(m,std::string("read16 unmapped 0x%08X",22),a);
return cast<uint16_t>(0ULL);
}
}
// tools/platform/gc/machine.go:235:1
uint32_t gc_Machine_Read32(gc_Machine* m,uint32_t a){
{
if ((cast<uint32_t>((a + cast<uint32_t>(3ULL))) < cast<uint32_t>(25165824ULL))) {
uint32_t v = be_Uint32(rrBorrow(m->RAM,a,cast<uint32_t>((a + cast<uint32_t>(4ULL)))));
gc_Machine_readWatch(m,a,v);
return v;
}
if (((a >= cast<uint32_t>(201326592ULL)) && (a < cast<uint32_t>(201392128ULL)))) {
return gc_Machine_regRead(m,a,cast<int64_t>(4ULL));
}
gc_Machine_logf(m,std::string("read32 unmapped 0x%08X (PC 0x%08X)",34),a,m->CPU->PC);
return cast<uint32_t>(0ULL);
}
}
// tools/platform/gc/machine.go:248:1
void gc_Machine_Write8(gc_Machine* m,uint32_t a,uint8_t v){
{
m->stores++;
if ((a < cast<uint32_t>(25165824ULL))) {
gc_Machine_writeWatch(m,a,cast<uint32_t>(v));
m->RAM[a] = v;
return ;
}
if (((a >= cast<uint32_t>(201359360ULL)) && (a < cast<uint32_t>(201359392ULL)))) {
gc_wgPipe_write8(&(m->wgFIFO),m,v);
return ;
}
if (((a >= cast<uint32_t>(201326592ULL)) && (a < cast<uint32_t>(201392128ULL)))) {
gc_Machine_regWrite(m,a,cast<uint32_t>(v),cast<int64_t>(1ULL));
return ;
}
gc_Machine_logf(m,std::string("write8 unmapped 0x%08X = 0x%02X",31),a,v);
}
}
// tools/platform/gc/machine.go:266:1
void gc_Machine_Write16(gc_Machine* m,uint32_t a,uint16_t v){
{
m->stores++;
if ((cast<uint32_t>((a + cast<uint32_t>(1ULL))) < cast<uint32_t>(25165824ULL))) {
gc_Machine_writeWatch(m,a,cast<uint32_t>(v));
be_PutUint16(sub(m->RAM,a,cast<uint32_t>((a + cast<uint32_t>(2ULL)))),v);
return ;
}
if (((a >= cast<uint32_t>(201359360ULL)) && (a < cast<uint32_t>(201359392ULL)))) {
gc_wgPipe_write16(&(m->wgFIFO),m,v);
return ;
}
if (((a >= cast<uint32_t>(201326592ULL)) && (a < cast<uint32_t>(201392128ULL)))) {
gc_Machine_regWrite(m,a,cast<uint32_t>(v),cast<int64_t>(2ULL));
return ;
}
gc_Machine_logf(m,std::string("write16 unmapped 0x%08X = 0x%04X",32),a,v);
}
}
// tools/platform/gc/machine.go:284:1
void gc_Machine_Write32(gc_Machine* m,uint32_t a,uint32_t v){
{
m->stores++;
if ((cast<uint32_t>((a + cast<uint32_t>(3ULL))) < cast<uint32_t>(25165824ULL))) {
gc_Machine_writeWatch(m,a,v);
be_PutUint32(rrBorrow(m->RAM,a,cast<uint32_t>((a + cast<uint32_t>(4ULL)))),v);
return ;
}
if (((a >= cast<uint32_t>(201359360ULL)) && (a < cast<uint32_t>(201359392ULL)))) {
gc_wgPipe_write32(&(m->wgFIFO),m,v);
return ;
}
if (((a >= cast<uint32_t>(201326592ULL)) && (a < cast<uint32_t>(201392128ULL)))) {
gc_Machine_regWrite(m,a,v,cast<int64_t>(4ULL));
return ;
}
gc_Machine_logf(m,std::string("write32 unmapped 0x%08X = 0x%08X (PC 0x%08X)",44),a,v,m->CPU->PC);
}
}
// tools/platform/gc/machine.go:304:1
uint32_t gc_Machine_Fetch32_reference(gc_Machine* m,uint32_t a){
{
if ((cast<uint32_t>((a + cast<uint32_t>(3ULL))) < cast<uint32_t>(25165824ULL))) {
return be_Uint32(rrBorrow(m->RAM,a,cast<uint32_t>((a + cast<uint32_t>(4ULL)))));
}
gc_Machine_logf(m,std::string("fetch from unmapped 0x%08X",26),a);
return cast<uint32_t>(0ULL);
}
}
// tools/platform/gc/machine.go:314:1
uint32_t gc_Machine_regRead(gc_Machine* m,uint32_t a,int64_t size){
{
uint32_t off = cast<uint32_t>((a & cast<uint32_t>(65535ULL)));
{
if (((off >= cast<uint32_t>(0ULL)) && (off < cast<uint32_t>(4096ULL)))){
return gc_cp_read(&(m->cp),m,off,size);
}
else if (((off >= cast<uint32_t>(4096ULL)) && (off < cast<uint32_t>(8192ULL)))){
return gc_pe_read(&(m->pe),m,off,size);
}
else if (((off >= cast<uint32_t>(8192ULL)) && (off < cast<uint32_t>(12288ULL)))){
return gc_vi_read(&(m->vi),m,off,size);
}
else if (((off >= cast<uint32_t>(12288ULL)) && (off < cast<uint32_t>(16384ULL)))){
return gc_pi_read(&(m->pi),m,off,size);
}
else if (((off >= cast<uint32_t>(16384ULL)) && (off < cast<uint32_t>(20480ULL)))){
return gc_mi_read(&(m->mi),m,off,size);
}
else if (((off >= cast<uint32_t>(20480ULL)) && (off < cast<uint32_t>(24576ULL)))){
return gc_dsp_read(&(m->dsp),m,off,size);
}
else if (((off >= cast<uint32_t>(24576ULL)) && (off < cast<uint32_t>(25600ULL)))){
return gc_di_read(&(m->di),m,off,size);
}
else if (((off >= cast<uint32_t>(25600ULL)) && (off < cast<uint32_t>(26624ULL)))){
return gc_si_read(&(m->si),m,off,size);
}
else if (((off >= cast<uint32_t>(26624ULL)) && (off < cast<uint32_t>(27648ULL)))){
return gc_exi_read(&(m->exi),m,off,size);
}
else if (((off >= cast<uint32_t>(27648ULL)) && (off < cast<uint32_t>(28672ULL)))){
return gc_ai_read(&(m->ai),m,off,size);
}
}
tmp160:;
gc_Machine_logf(m,std::string("read%d unmodelled register 0x%08X (PC 0x%08X)",45),cast<int64_t>((size * cast<int64_t>(8ULL))),a,m->CPU->PC);
return cast<uint32_t>(0ULL);
}
}
// tools/platform/gc/machine.go:342:1
void gc_Machine_regWrite(gc_Machine* m,uint32_t a,uint32_t v,int64_t size){
{
uint32_t off = cast<uint32_t>((a & cast<uint32_t>(65535ULL)));
{
if (((off >= cast<uint32_t>(0ULL)) && (off < cast<uint32_t>(4096ULL)))){
gc_cp_write(&(m->cp),m,off,v,size);
}
else if (((off >= cast<uint32_t>(4096ULL)) && (off < cast<uint32_t>(8192ULL)))){
gc_pe_write(&(m->pe),m,off,v,size);
}
else if (((off >= cast<uint32_t>(8192ULL)) && (off < cast<uint32_t>(12288ULL)))){
gc_vi_write(&(m->vi),m,off,v,size);
}
else if (((off >= cast<uint32_t>(12288ULL)) && (off < cast<uint32_t>(16384ULL)))){
gc_pi_write(&(m->pi),m,off,v,size);
}
else if (((off >= cast<uint32_t>(16384ULL)) && (off < cast<uint32_t>(20480ULL)))){
gc_mi_write(&(m->mi),m,off,v,size);
}
else if (((off >= cast<uint32_t>(20480ULL)) && (off < cast<uint32_t>(24576ULL)))){
gc_dsp_write(&(m->dsp),m,off,v,size);
}
else if (((off >= cast<uint32_t>(24576ULL)) && (off < cast<uint32_t>(25600ULL)))){
gc_di_write(&(m->di),m,off,v,size);
}
else if (((off >= cast<uint32_t>(25600ULL)) && (off < cast<uint32_t>(26624ULL)))){
gc_si_write(&(m->si),m,off,v,size);
}
else if (((off >= cast<uint32_t>(26624ULL)) && (off < cast<uint32_t>(27648ULL)))){
gc_exi_write(&(m->exi),m,off,v,size);
}
else if (((off >= cast<uint32_t>(27648ULL)) && (off < cast<uint32_t>(28672ULL)))){
gc_ai_write(&(m->ai),m,off,v,size);
}
else {
gc_Machine_logf(m,std::string("write%d unmodelled register 0x%08X = 0x%08X (PC 0x%08X)",55),cast<int64_t>((size * cast<int64_t>(8ULL))),a,v,m->CPU->PC);
}
}
tmp161:;
}
}
// tools/platform/gc/machine.go:376:1
void gc_Machine_dmaToRAM(gc_Machine* m,uint32_t addr,Slice<uint8_t> data){
{
addr = gc_phys(addr);
{auto&& tmp162 = data;
for(int64_t tmp163=0;tmp163<len(tmp162);++tmp163){
auto i=tmp163;auto b=tmp162[tmp163];if ((cast<int64_t>((cast<int64_t>(addr) + i)) < len(m->RAM))) {
m->RAM[cast<uint32_t>((addr + cast<uint32_t>(i)))] = b;
}
}}
}
}
// tools/platform/gc/machine.go:392:1
uint32_t gc_phys(uint32_t a){
{
return cast<uint32_t>((a & cast<uint32_t>(67108863ULL)));
}
}
// tools/platform/gc/machine.go:394:1
uint32_t gc_Machine_ram32(gc_Machine* m,uint32_t a){
{
a = gc_phys(a);
if ((cast<uint32_t>((a + cast<uint32_t>(3ULL))) >= cast<uint32_t>(25165824ULL))) {
return cast<uint32_t>(0ULL);
}
return be_Uint32(rrBorrow(m->RAM,a,cast<uint32_t>((a + cast<uint32_t>(4ULL)))));
}
}
// tools/platform/gc/machine.go:402:1
void gc_Machine_setRAM32(gc_Machine* m,uint32_t a,uint32_t v){
{
a = gc_phys(a);
if ((cast<uint32_t>((a + cast<uint32_t>(3ULL))) >= cast<uint32_t>(25165824ULL))) {
return ;
}
be_PutUint32(rrBorrow(m->RAM,a,cast<uint32_t>((a + cast<uint32_t>(4ULL)))),v);
}
}
// tools/platform/gc/machine.go:412:1
void gc_Machine_readWatch(gc_Machine* m,uint32_t a,uint32_t v){
{
if (((bool(m->OnRead) && (a >= m->RWatchLo)) && (a < m->RWatchHi))) {
m->OnRead(a,v,m->CPU->PC);
}
}
}
// tools/platform/gc/machine.go:418:1
void gc_Machine_writeWatch(gc_Machine* m,uint32_t a,uint32_t v){
{
if (((bool(m->OnWrite) && (a >= m->WatchLo)) && (a < m->WatchHi))) {
m->OnWrite(a,v,m->CPU->PC);
}
}
}
// tools/platform/gc/machine.go:438:1
Slice<std::string> gc_Machine_Census(gc_Machine* m){
{
return m->Log;
}
}
// tools/platform/gc/machine.go:442:1
std::array<uint64_t,256> gc_Machine_GPUCensus(gc_Machine* m){
{
return m->gpu.Census;
}
}
// tools/platform/gc/pi.go:40:1
uint32_t gc_pi_read(gc_pi* p,gc_Machine* m,uint32_t off,int64_t size){
{
{
switch(cast<uint32_t>((off & cast<uint32_t>(4095ULL)))){
case cast<uint32_t>(0ULL):{
return p->Cause;
break;}
case cast<uint32_t>(4ULL):{
return p->Mask;
break;}
case cast<uint32_t>(12ULL):{
return p->FIFOBase;
break;}
case cast<uint32_t>(16ULL):{
return p->FIFOEnd;
break;}
case cast<uint32_t>(20ULL):{
return p->FIFOWrite;
break;}
case cast<uint32_t>(36ULL):{
return cast<uint32_t>(538968070ULL);
break;}
case cast<uint32_t>(44ULL):{
return p->ResetCode;
break;}
}}
gc_Machine_logf(m,std::string("PI read unmodelled 0x%03X",25),cast<uint32_t>((off & cast<uint32_t>(4095ULL))));
return cast<uint32_t>(0ULL);
}
}
// tools/platform/gc/pi.go:63:1
void gc_pi_write(gc_pi* p,gc_Machine* m,uint32_t off,uint32_t v,int64_t size){
{
{
switch(cast<uint32_t>((off & cast<uint32_t>(4095ULL)))){
case cast<uint32_t>(0ULL):{
p->Cause &= ~(v);
gc_Machine_updateIRQ(m);
break;}
case cast<uint32_t>(4ULL):{
p->Mask = v;
gc_Machine_updateIRQ(m);
break;}
case cast<uint32_t>(12ULL):{
p->FIFOBase = cast<uint32_t>((v & cast<uint32_t>(67108832ULL)));
break;}
case cast<uint32_t>(16ULL):{
p->FIFOEnd = cast<uint32_t>((v & cast<uint32_t>(67108832ULL)));
break;}
case cast<uint32_t>(20ULL):{
p->FIFOWrite = cast<uint32_t>((v & cast<uint32_t>(67108832ULL)));
break;}
case cast<uint32_t>(36ULL):{
break;}
case cast<uint32_t>(44ULL):{
p->ResetCode = v;
break;}
default:{
gc_Machine_logf(m,std::string("PI write unmodelled 0x%03X = 0x%08X",35),cast<uint32_t>((off & cast<uint32_t>(4095ULL))),v);
break;}
}}
}
}
// tools/platform/gc/pi.go:89:1
void gc_Machine_raiseInt(gc_Machine* m,int64_t cause){
{
m->pi.Cause |= shl<uint32_t>(cast<uint32_t>(1ULL),cause);
gc_Machine_updateIRQ(m);
}
}
// tools/platform/gc/pi.go:95:1
void gc_Machine_clearInt(gc_Machine* m,int64_t cause){
{
m->pi.Cause &= ~(shl<uint32_t>(cast<uint32_t>(1ULL),cause));
gc_Machine_updateIRQ(m);
}
}
// tools/platform/gc/pi.go:103:1
void gc_Machine_updateIRQ(gc_Machine* m){
{
gekko_CPU_Interrupt(m->CPU,(cast<uint32_t>((m->pi.Cause & m->pi.Mask)) != cast<uint32_t>(0ULL)));
}
}
// tools/platform/gc/profile.go:116:1
gc_profCounters gc_Machine_profCounters(gc_Machine* m){
{
gc_gpu* g = (&m->gpu);
return gc_profCounters{m->gxTotalCmds,g->profDraws,g->profCulled,g->pixWritten,g->pixZRej,g->pixARej,cast<int64_t>(m->wgFIFO.Bytes)};
}
}
// tools/platform/gc/profile.go:126:1
time_Time gc_Machine_profStart(gc_Machine* m){
{
if ((!m->Profile)) {
return time_Time{};
}
return go_time_Now();
}
}
// tools/platform/gc/profile.go:133:1
void gc_Machine_profEnd(gc_Machine* m,int64_t bucket,time_Time t){
{
if (time_Time_IsZero(t)) {
return ;
}
m->prof.ns[bucket] += cast<int64_t>(go_time_Since(t));
m->prof.count[bucket]++;
}
}
// tools/platform/gc/profile.go:144:1
void gc_Machine_profFIFOEnter(gc_Machine* m){
{
if (((!m->Profile) || m->prof.inFIFO)) {
return ;
}
auto tmp164 = std::make_tuple(go_time_Now(),true);
m->prof.fifoStart = std::get<0>(tmp164);
m->prof.inFIFO = std::get<1>(tmp164);
}
}
// tools/platform/gc/profile.go:151:1
void gc_Machine_profFIFOExit(gc_Machine* m){
{
if (((!m->Profile) || (!m->prof.inFIFO))) {
return ;
}
m->prof.ns[cast<int64_t>(0ULL)] += cast<int64_t>(go_time_Since(m->prof.fifoStart));
m->prof.count[cast<int64_t>(0ULL)]++;
m->prof.inFIFO = false;
}
}
// tools/platform/gc/profile.go:162:1
void gc_Machine_profRunEnter(gc_Machine* m){
{
if ((!m->Profile)) {
return ;
}
auto tmp165 = std::make_tuple(go_time_Now(),true);
m->prof.runStart = std::get<0>(tmp165);
m->prof.inRun = std::get<1>(tmp165);
}
}
// tools/platform/gc/profile.go:169:1
void gc_Machine_profRunExit(gc_Machine* m){
{
if (((!m->Profile) || (!m->prof.inRun))) {
return ;
}
m->prof.frameNs += cast<int64_t>(go_time_Since(m->prof.runStart));
m->prof.inRun = false;
}
}
// tools/platform/gc/profile.go:194:1
void gc_Machine_profFrame(gc_Machine* m){
{
if ((!m->Profile)) {
return ;
}
gc_profState* p = (&m->prof);
int64_t total = p->frameNs;
if (p->inRun) {
total += cast<int64_t>(go_time_Since(p->runStart));
p->runStart = go_time_Now();
}
p->frameNs = cast<int64_t>(0ULL);
if (p->inFIFO) {
p->ns[cast<int64_t>(0ULL)] += cast<int64_t>(go_time_Since(p->fifoStart));
p->count[cast<int64_t>(0ULL)]++;
p->fifoStart = go_time_Now();
}
auto ms = [&](int64_t ns)->double{
return (cast<double>(ns) / 1e6);
}
;
int64_t nested = cast<int64_t>((cast<int64_t>((p->ns[cast<int64_t>(1ULL)] + p->ns[cast<int64_t>(2ULL)])) + p->ns[cast<int64_t>(3ULL)]));
int64_t decode = cast<int64_t>((p->ns[cast<int64_t>(0ULL)] - nested));
if ((decode < cast<int64_t>(0ULL))) {
decode = cast<int64_t>(0ULL);
}
Slice<gc_ProfileBucket> buckets = Slice<gc_ProfileBucket>{gc_ProfileBucket{std::string("command decode (derived)",24),ms(decode),cast<int64_t>(p->count[cast<int64_t>(0ULL)])},gc_ProfileBucket{std::string("vertex + xf",11),ms(p->ns[cast<int64_t>(1ULL)]),cast<int64_t>(p->count[cast<int64_t>(1ULL)])},gc_ProfileBucket{std::string("rasterise",9),ms(p->ns[cast<int64_t>(2ULL)]),cast<int64_t>(p->count[cast<int64_t>(2ULL)])},gc_ProfileBucket{std::string("pe copy",7),ms(p->ns[cast<int64_t>(3ULL)]),cast<int64_t>(p->count[cast<int64_t>(3ULL)])},gc_ProfileBucket{std::string("dsp",3),ms(p->ns[cast<int64_t>(4ULL)]),cast<int64_t>(p->count[cast<int64_t>(4ULL)])}};
int64_t summed = cast<int64_t>((cast<int64_t>((decode + nested)) + p->ns[cast<int64_t>(4ULL)]));
int64_t other = cast<int64_t>((total - summed));
if ((other < cast<int64_t>(0ULL))) {
other = cast<int64_t>(0ULL);
}
buckets = append(buckets,Slice<gc_ProfileBucket>{gc_ProfileBucket{std::string("gekko + rest (derived)",22),ms(other),cast<int64_t>(0ULL)}});
gc_profCounters now = gc_Machine_profCounters(m);
gc_profCounters d = gc_profCounters{cast<int64_t>((now.cmds - p->base.cmds)),cast<int64_t>((now.draws - p->base.draws)),cast<int64_t>((now.culled - p->base.culled)),cast<int64_t>((now.frags - p->base.frags)),cast<int64_t>((now.zRejected - p->base.zRejected)),cast<int64_t>((now.aRejected - p->base.aRejected)),cast<int64_t>((now.fifoBytes - p->base.fifoBytes))};
int64_t instrs = cast<int64_t>(cast<uint64_t>((m->Instrs - p->baseInstr)));
p->last = gc_FrameProfile{ms(total),buckets,Slice<gc_ProfileCounter>{gc_ProfileCounter{std::string("gx commands",11),d.cmds},gc_ProfileCounter{std::string("draws",5),d.draws},gc_ProfileCounter{std::string("tris culled",11),d.culled},gc_ProfileCounter{std::string("fragments drawn",15),d.frags},gc_ProfileCounter{std::string("depth-rejected",14),d.zRejected},gc_ProfileCounter{std::string("alpha-rejected",14),d.aRejected},gc_ProfileCounter{std::string("fifo bytes",10),d.fifoBytes},gc_ProfileCounter{std::string("gekko instructions",18),instrs}},(d.draws > cast<int64_t>(0ULL))};
p->has = true;
auto tmp166 = std::make_tuple(std::array<int64_t,5>{},std::array<int64_t,5>{});
p->ns = std::get<0>(tmp166);
p->count = std::get<1>(tmp166);
auto tmp167 = std::make_tuple(now,m->Instrs);
p->base = std::get<0>(tmp167);
p->baseInstr = std::get<1>(tmp167);
}
}
// tools/platform/gc/profile.go:271:1
gc_FrameProfile gc_Machine_FrameProfile(gc_Machine* m){
{
if ((!m->prof.has)) {
return gc_FrameProfile{};
}
return m->prof.last;
}
}
// tools/platform/gc/profile.go:280:1
void gc_Machine_SetProfile(gc_Machine* m,bool on){
{
m->Profile = on;
m->prof = gc_profState{};
if (on) {
auto tmp168 = std::make_tuple(gc_Machine_profCounters(m),m->Instrs);
m->prof.base = std::get<0>(tmp168);
m->prof.baseInstr = std::get<1>(tmp168);
}
}
}
// tools/platform/gc/run.go:18:1
std::string gc_Result_String(gc_Result r){
{
return go_fmt_Sprintf(std::string("stopped at 0x%08X after %d steps: %s",36),r.PC,r.Steps,r.Reason);
}
}
// tools/platform/gc/run.go:34:1
void gc_Machine_SetSpinDetect(gc_Machine* m,bool on){
{
m->noSpin = (!on);
}
}
// tools/platform/gc/run.go:37:1
void gc_Machine_SetBreakpoint(gc_Machine* m,uint32_t vaddr){
{
if ((!m->run.breakpoints)) {
m->run.breakpoints = Map<uint32_t,bool>{};
}
m->run.breakpoints[vaddr] = true;
}
}
// tools/platform/gc/run.go:45:1
void gc_Machine_ClearBreakpoints(gc_Machine* m){
{
m->run.breakpoints = {};
}
}
// tools/platform/gc/run.go:62:1
gc_Result gc_Machine_RunStopAfterGXCommand(gc_Machine* m,int64_t n,uint64_t maxSteps){
{
auto tmp169 = std::make_tuple(cast<int64_t>(0ULL),n,false);
m->gxCmdCount = std::get<0>(tmp169);
m->gxStopAfter = std::get<1>(tmp169);
m->gxStopped = std::get<2>(tmp169);
gc_Result res = gc_Machine_Run(m,maxSteps);
auto tmp170 = std::make_tuple(cast<int64_t>(0ULL),false);
m->gxStopAfter = std::get<0>(tmp170);
m->gxStopped = std::get<1>(tmp170);
return res;
}
}
// tools/platform/gc/run.go:71:1
int64_t gc_Machine_GXCommandCount(gc_Machine* m){
{
return m->gxCmdCount;
}
}
// tools/platform/gc/run.go:88:1
gc_Result gc_Machine_RunFields(gc_Machine* m,int64_t n,uint64_t budget){
{
if ((n <= cast<int64_t>(0ULL))) {
return gc_Result{{},m->CPU->PC,std::string("no fields requested",19)};
}
uint64_t target = cast<uint64_t>((m->vi.Field + cast<uint64_t>(n)));
std::function<void(gc_Machine*)> prev = m->OnDisplay;
m->OnDisplay = [&](gc_Machine* mm)->void{
if (bool(prev)) {
prev(mm);
}
if ((mm->vi.Field >= target)) {
mm->StopRequested = true;
}
}
;
auto tmp171=defer([&](){[&]()->void{
m->OnDisplay = prev;
}
();});
gc_Result res = gc_Machine_Run(m,budget);
if (((res.Reason == std::string("stop requested",14)) && (m->vi.Field >= target))) {
res.Reason = go_fmt_Sprintf(std::string("%d fields",9),n);
}
return res;
}
}
// tools/platform/gc/run.go:114:1
gc_Result gc_Machine_Run(gc_Machine* m,uint64_t maxSteps){
{rrprof::Scope timing(0,"Gekko and devices");
{
gc_Machine_profRunEnter(m);
auto tmp172=defer([&](){gc_Machine_profRunExit(m);});
uint64_t start = m->Instrs;
auto steps = [&]()->uint64_t{
return cast<uint64_t>((m->Instrs - start));
}
;
bool first = true;
std::array<uint32_t,4> spinPCs={};
int64_t spinN={};
uint64_t spinStart = m->Instrs;
{;for (;(steps() < maxSteps);){
uint32_t pc = m->CPU->PC;
if (m->StopRequested) {
m->StopRequested = false;
return gc_Result{steps(),pc,std::string("stop requested",14)};
}
if ((get(m->run.breakpoints,pc) && (!first))) {
return gc_Result{steps(),pc,go_fmt_Sprintf(std::string("breakpoint at 0x%08X",20),pc)};
}
first = false;
if (bool(m->OnStep)) {
m->OnStep(m,pc);
}
if (((m->stack.every != cast<uint64_t>(0ULL)) && (m->Instrs >= m->stack.next))) {
gc_Machine_sampleStack(m,pc);
}
if (((!m->noIdle) && gc_Machine_idleStep(m,pc))) {
{
uint64_t n = gc_Machine_idleDeadline(m);
if ((n > cast<uint64_t>(0ULL))) {
{
uint64_t rem = cast<uint64_t>((maxSteps - steps()));
if ((n > rem)) {
n = rem;
}
}
gc_Machine_idleSkip(m,n);
}
}
}
gc_Machine_tickVI(m);
gc_Machine_tickDSP(m);
gc_Machine_tickAID(m);
gc_Machine_tickDI(m);
if ((!m->noSpin)) {
if ((spinN < cast<int64_t>(4ULL))) {
bool seen = false;
{int64_t i = cast<int64_t>(0ULL);for (;(i < spinN);i++){
if ((spinPCs[i] == pc)) {
seen = true;
break;
}
}
}if ((!seen)) {
spinPCs[spinN] = pc;
spinN++;
}
}
if ((cast<uint64_t>((m->Instrs - spinStart)) >= cast<uint64_t>(16200000ULL))) {
if ((spinN < cast<int64_t>(4ULL))) {
return gc_Result{steps(),pc,std::string("spin (tight loop)",17)};
}
spinN = cast<int64_t>(0ULL);
spinStart = m->Instrs;
}
}
gekko_CPU_Step(m->CPU);
m->Instrs++;
if (m->CPU->Halted) {
return gc_Result{steps(),m->CPU->PC,m->CPU->HaltReason};
}
}
}return gc_Result{steps(),m->CPU->PC,std::string("step budget exhausted",21)};
}
}
}
// tools/platform/gc/si.go:93:1
void gc_si_connectPad(gc_si* d,int64_t port){
{
d->Pad[port] = gc_padPort{true,{},cast<uint8_t>(128ULL),cast<uint8_t>(128ULL),cast<uint8_t>(128ULL),cast<uint8_t>(128ULL),{},{}};
}
}
// tools/platform/gc/si.go:97:1
uint32_t gc_si_read(gc_si* d,gc_Machine* m,uint32_t off,int64_t size){
{
uint32_t r = cast<uint32_t>((off & cast<uint32_t>(255ULL)));
uint32_t v = gc_si_readReg(d,r,size);
if (gc_siTrace) {
gc_siLog(std::string("SI rd 0x%02X -> 0x%08X (pc 0x%08X)\012",35),r,v,m->CPU->PC);
}
return v;
}
}
// tools/platform/gc/si.go:106:1
uint32_t gc_si_readReg(gc_si* d,uint32_t r,int64_t size){
{
{
if ((r < cast<uint32_t>(48ULL))){
auto tmp174 = std::make_tuple(divi<uint32_t>(r,cast<uint32_t>(12ULL)),divi<uint32_t>((modi<uint32_t>(r,cast<uint32_t>(12ULL))),cast<uint32_t>(4ULL)));
uint32_t c = std::get<0>(tmp174);
uint32_t w = std::get<1>(tmp174);
return d->Chan[c][w];
}
else if ((r == cast<uint32_t>(48ULL))){
return d->Poll;
}
else if ((r == cast<uint32_t>(52ULL))){
return d->ComCSR;
}
else if ((r == cast<uint32_t>(56ULL))){
return d->Status;
}
else if ((r == cast<uint32_t>(60ULL))){
return d->ExiLk;
}
else if (((r >= cast<uint32_t>(128ULL)) && (r < cast<uint32_t>(256ULL)))){
return gc_si_iobufRead(d,cast<uint32_t>((r - cast<uint32_t>(128ULL))),size);
}
}
tmp173:;
return cast<uint32_t>(0ULL);
}
}
// tools/platform/gc/si.go:125:1
void gc_si_write(gc_si* d,gc_Machine* m,uint32_t off,uint32_t v,int64_t size){
{
uint32_t r = cast<uint32_t>((off & cast<uint32_t>(255ULL)));
if (gc_siTrace) {
gc_siLog(std::string("SI wr 0x%02X = 0x%08X (pc 0x%08X)\012",34),r,v,m->CPU->PC);
}
{
if ((r < cast<uint32_t>(48ULL))){
auto tmp176 = std::make_tuple(divi<uint32_t>(r,cast<uint32_t>(12ULL)),divi<uint32_t>((modi<uint32_t>(r,cast<uint32_t>(12ULL))),cast<uint32_t>(4ULL)));
uint32_t c = std::get<0>(tmp176);
uint32_t w = std::get<1>(tmp176);
d->Chan[c][w] = v;
}
else if ((r == cast<uint32_t>(48ULL))){
d->Poll = v;
}
else if ((r == cast<uint32_t>(52ULL))){
gc_si_writeComCSR(d,m,v);
}
else if ((r == cast<uint32_t>(56ULL))){
d->Status = v;
}
else if ((r == cast<uint32_t>(60ULL))){
d->ExiLk = v;
}
else if (((r >= cast<uint32_t>(128ULL)) && (r < cast<uint32_t>(256ULL)))){
gc_si_iobufWrite(d,cast<uint32_t>((r - cast<uint32_t>(128ULL))),v,size);
}
else {
gc_Machine_logf(m,std::string("SI write unmodelled 0x%02X = 0x%08X",35),r,v);
}
}
tmp175:;
}
}
// tools/platform/gc/si.go:154:1
void gc_si_writeComCSR(gc_si* d,gc_Machine* m,uint32_t v){
{
uint32_t tcint = cast<uint32_t>((d->ComCSR & cast<uint32_t>(2147483648ULL)));
if ((cast<uint32_t>((v & cast<uint32_t>(2147483648ULL))) != cast<uint32_t>(0ULL))) {
tcint = cast<uint32_t>(0ULL);
}
d->ComCSR = cast<uint32_t>((tcint | ((v & ~(cast<uint32_t>(2147483648ULL))))));
if ((cast<uint32_t>((v & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL))) {
gc_si_startTransfer(d,m);
return ;
}
gc_Machine_siRefreshIRQ(m);
}
}
// tools/platform/gc/si.go:172:1
void gc_si_startTransfer(gc_si* d,gc_Machine* m){
{
uint32_t c = cast<uint32_t>(((shr<uint32_t>(d->ComCSR,cast<int64_t>(1ULL))) & cast<uint32_t>(3ULL)));
uint32_t inlen = cast<uint32_t>(((shr<uint32_t>(d->ComCSR,cast<int64_t>(8ULL))) & cast<uint32_t>(127ULL)));
if ((inlen == cast<uint32_t>(0ULL))) {
inlen = cast<uint32_t>(128ULL);
}
uint8_t cmd = d->IOBuf[cast<int64_t>(0ULL)];
uint32_t shift = cast<uint32_t>(((cast<uint32_t>((cast<uint32_t>(3ULL) - c))) * cast<uint32_t>(8ULL)));
d->Status &= ~(shl<uint32_t>(cast<uint32_t>(255ULL),shift));
if (d->Pad[c].Connected) {
Slice<uint8_t> resp = gc_padPort_respond(&(d->Pad[c]),cmd,cast<int64_t>(inlen));
gcopy(sub(d->IOBuf,0,len(d->IOBuf)),resp);
d->Status |= shl<uint32_t>(cast<uint32_t>(32ULL),shift);
}
else {
d->Status |= shl<uint32_t>(cast<uint32_t>(8ULL),shift);
}
d->ComCSR = cast<uint32_t>((((d->ComCSR & ~(cast<uint32_t>(1ULL)))) | cast<uint32_t>(2147483648ULL)));
gc_Machine_siRefreshIRQ(m);
}
}
// tools/platform/gc/si.go:203:1
Slice<uint8_t> gc_padPort_respond(gc_padPort* p,uint8_t cmd,int64_t inlen){
{
Slice<uint8_t> buf = Slice<uint8_t>::make(inlen);
{
switch(cmd){
case cast<uint8_t>(0ULL):case cast<uint8_t>(255ULL):{
if ((inlen >= cast<int64_t>(3ULL))) {
auto tmp177 = std::make_tuple(cast<uint8_t>(9ULL),cast<uint8_t>(0ULL),cast<uint8_t>(0ULL));
buf[cast<int64_t>(0ULL)] = std::get<0>(tmp177);
buf[cast<int64_t>(1ULL)] = std::get<1>(tmp177);
buf[cast<int64_t>(2ULL)] = std::get<2>(tmp177);
}
break;}
default:{
auto tmp178 = gc_padPort_status(p);
uint32_t hi = std::get<0>(tmp178);
uint32_t lo = std::get<1>(tmp178);
gc_putWord(buf,cast<int64_t>(0ULL),hi);
gc_putWord(buf,cast<int64_t>(4ULL),lo);
break;}
}}
return buf;
}
}
// tools/platform/gc/si.go:222:1
std::tuple<uint32_t,uint32_t> gc_padPort_status(gc_padPort* p){
uint32_t hi{};
uint32_t lo{};
{
hi = cast<uint32_t>((cast<uint32_t>((shl<uint32_t>(cast<uint32_t>(p->Buttons),cast<int64_t>(16ULL)) | shl<uint32_t>(cast<uint32_t>(p->StickX),cast<int64_t>(8ULL)))) | cast<uint32_t>(p->StickY)));
lo = cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((shl<uint32_t>(cast<uint32_t>(p->SubX),cast<int64_t>(24ULL)) | shl<uint32_t>(cast<uint32_t>(p->SubY),cast<int64_t>(16ULL)))) | shl<uint32_t>(cast<uint32_t>(p->TriggerL),cast<int64_t>(8ULL)))) | cast<uint32_t>(p->TriggerR)));
return {hi,lo};
}
}
// tools/platform/gc/si.go:231:1
void gc_Machine_tickSI(gc_Machine* m){
{
if ((m->si.Poll == cast<uint32_t>(0ULL))) {
return ;
}
{int64_t c = cast<int64_t>(0ULL);for (;(c < cast<int64_t>(4ULL));c++){
gc_padPort* p = (&m->si.Pad[c]);
if ((!p->Connected)) {
continue;
}
auto tmp179 = gc_padPort_status(p);
uint32_t hi = std::get<0>(tmp179);
uint32_t lo = std::get<1>(tmp179);
m->si.Chan[c][cast<int64_t>(1ULL)] = hi;
m->si.Chan[c][cast<int64_t>(2ULL)] = lo;
uint64_t shift = cast<uint64_t>((cast<uint64_t>(cast<int64_t>((cast<int64_t>(3ULL) - c))) * cast<uint64_t>(8ULL)));
m->si.Status = cast<uint32_t>((((m->si.Status & ~((shl<uint32_t>(cast<uint32_t>(255ULL),shift))))) | (shl<uint32_t>(cast<uint32_t>(32ULL),shift))));
}
}}
}
// tools/platform/gc/si.go:251:1
void gc_Machine_siRefreshIRQ(gc_Machine* m){
{
if (((cast<uint32_t>((m->si.ComCSR & cast<uint32_t>(2147483648ULL))) != cast<uint32_t>(0ULL)) && (cast<uint32_t>((m->si.ComCSR & cast<uint32_t>(1073741824ULL))) != cast<uint32_t>(0ULL)))) {
gc_Machine_raiseInt(m,cast<int64_t>(3ULL));
}
else {
gc_Machine_clearInt(m,cast<int64_t>(3ULL));
}
}
}
// tools/platform/gc/si.go:262:1
void gc_Machine_SetController(gc_Machine* m,int64_t port,bool connected){
{
if (((port < cast<int64_t>(0ULL)) || (port > cast<int64_t>(3ULL)))) {
return ;
}
if (connected) {
gc_si_connectPad(&(m->si),port);
}
else {
m->si.Pad[port] = gc_padPort{};
}
}
}
// tools/platform/gc/si.go:275:1
void gc_Machine_SetPadButtons(gc_Machine* m,int64_t port,uint16_t buttons){
{
if (((port < cast<int64_t>(0ULL)) || (port > cast<int64_t>(3ULL)))) {
return ;
}
m->si.Pad[port].Buttons = buttons;
}
}
// tools/platform/gc/si.go:284:1
uint16_t gc_Machine_PadButtons(gc_Machine* m,int64_t port){
{
if (((port < cast<int64_t>(0ULL)) || (port > cast<int64_t>(3ULL)))) {
return cast<uint16_t>(0ULL);
}
return m->si.Pad[port].Buttons;
}
}
// tools/platform/gc/si.go:315:1
void gc_Machine_SetPadStick(gc_Machine* m,int64_t port,uint8_t x,uint8_t y){
{
if (((port < cast<int64_t>(0ULL)) || (port > cast<int64_t>(3ULL)))) {
return ;
}
auto tmp180 = std::make_tuple(x,y);
m->si.Pad[port].StickX = std::get<0>(tmp180);
m->si.Pad[port].StickY = std::get<1>(tmp180);
}
}
// tools/platform/gc/si.go:323:1
std::tuple<uint8_t,uint8_t> gc_Machine_PadStick(gc_Machine* m,int64_t port){
uint8_t x{};
uint8_t y{};
{
if (((port < cast<int64_t>(0ULL)) || (port > cast<int64_t>(3ULL)))) {
return {cast<uint8_t>(128ULL),cast<uint8_t>(128ULL)};
}
return {m->si.Pad[port].StickX,m->si.Pad[port].StickY};
}
}
// tools/platform/gc/si.go:341:1
std::tuple<uint16_t,bool> gc_PadButton(std::string name){
{
auto tmp181 = lookup(gc_padButtons,name);
uint16_t b = std::get<0>(tmp181);
bool ok = std::get<1>(tmp181);
return {b,ok};
}
}
// tools/platform/gc/si.go:348:1
Slice<std::string> gc_PadButtonNames(){
{
Slice<std::string> out = Slice<std::string>::make(cast<int64_t>(0ULL),len(gc_padButtons));
{auto&& tmp182 = gc_padButtons;
for(auto [tmp183,tmp184]:tmp182){
auto n=tmp183;out = append(out,Slice<std::string>{n});
}}
go_sort_Strings(out);
return out;
}
}
// tools/platform/gc/si.go:359:1
uint32_t gc_si_iobufRead(gc_si* d,uint32_t o,int64_t size){
{
{
switch(size){
case cast<int64_t>(1ULL):{
return cast<uint32_t>(d->IOBuf[o]);
break;}
case cast<int64_t>(2ULL):{
return cast<uint32_t>((shl<uint32_t>(cast<uint32_t>(d->IOBuf[o]),cast<int64_t>(8ULL)) | cast<uint32_t>(d->IOBuf[cast<uint32_t>((o + cast<uint32_t>(1ULL)))])));
break;}
default:{
return cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((shl<uint32_t>(cast<uint32_t>(d->IOBuf[o]),cast<int64_t>(24ULL)) | shl<uint32_t>(cast<uint32_t>(d->IOBuf[cast<uint32_t>((o + cast<uint32_t>(1ULL)))]),cast<int64_t>(16ULL)))) | shl<uint32_t>(cast<uint32_t>(d->IOBuf[cast<uint32_t>((o + cast<uint32_t>(2ULL)))]),cast<int64_t>(8ULL)))) | cast<uint32_t>(d->IOBuf[cast<uint32_t>((o + cast<uint32_t>(3ULL)))])));
break;}
}}
}
}
// tools/platform/gc/si.go:370:1
void gc_si_iobufWrite(gc_si* d,uint32_t o,uint32_t v,int64_t size){
{
{
switch(size){
case cast<int64_t>(1ULL):{
d->IOBuf[o] = cast<uint8_t>(v);
break;}
case cast<int64_t>(2ULL):{
auto tmp185 = std::make_tuple(cast<uint8_t>(shr<uint32_t>(v,cast<int64_t>(8ULL))),cast<uint8_t>(v));
d->IOBuf[o] = std::get<0>(tmp185);
d->IOBuf[cast<uint32_t>((o + cast<uint32_t>(1ULL)))] = std::get<1>(tmp185);
break;}
default:{
auto tmp186 = std::make_tuple(cast<uint8_t>(shr<uint32_t>(v,cast<int64_t>(24ULL))),cast<uint8_t>(shr<uint32_t>(v,cast<int64_t>(16ULL))),cast<uint8_t>(shr<uint32_t>(v,cast<int64_t>(8ULL))),cast<uint8_t>(v));
d->IOBuf[o] = std::get<0>(tmp186);
d->IOBuf[cast<uint32_t>((o + cast<uint32_t>(1ULL)))] = std::get<1>(tmp186);
d->IOBuf[cast<uint32_t>((o + cast<uint32_t>(2ULL)))] = std::get<2>(tmp186);
d->IOBuf[cast<uint32_t>((o + cast<uint32_t>(3ULL)))] = std::get<3>(tmp186);
break;}
}}
}
}
// tools/platform/gc/si.go:381:1
void gc_putWord(Slice<uint8_t> b,int64_t o,uint32_t v){
{
if ((cast<int64_t>((o + cast<int64_t>(3ULL))) < len(b))) {
auto tmp187 = std::make_tuple(cast<uint8_t>(shr<uint32_t>(v,cast<int64_t>(24ULL))),cast<uint8_t>(shr<uint32_t>(v,cast<int64_t>(16ULL))),cast<uint8_t>(shr<uint32_t>(v,cast<int64_t>(8ULL))),cast<uint8_t>(v));
b[o] = std::get<0>(tmp187);
b[cast<int64_t>((o + cast<int64_t>(1ULL)))] = std::get<1>(tmp187);
b[cast<int64_t>((o + cast<int64_t>(2ULL)))] = std::get<2>(tmp187);
b[cast<int64_t>((o + cast<int64_t>(3ULL)))] = std::get<3>(tmp187);
}
}
}
// tools/platform/gc/stackprof.go:61:1
void gc_Machine_SetStackProfile(gc_Machine* m,uint64_t every,int64_t depth){
{
if ((depth <= cast<int64_t>(0ULL))) {
depth = cast<int64_t>(8ULL);
}
if ((every == cast<uint64_t>(0ULL))) {
every = cast<uint64_t>(199ULL);
}
m->stack = gc_stackProf{every,depth,cast<uint64_t>((m->Instrs + every)),Map<std::string,gc_StackSample*>{},{}};
}
}
// tools/platform/gc/stackprof.go:77:1
void gc_Machine_StopStackProfile(gc_Machine* m){
{
m->stack.every = cast<uint64_t>(0ULL);
}
}
// tools/platform/gc/stackprof.go:87:1
void gc_Machine_sampleStack(gc_Machine* m,uint32_t pc){
{
m->stack.next = cast<uint64_t>((m->Instrs + m->stack.every));
m->stack.samples++;
Slice<uint32_t> frames = Slice<uint32_t>::make(cast<int64_t>(0ULL),cast<int64_t>((m->stack.depth + cast<int64_t>(1ULL))));
frames = append(frames,Slice<uint32_t>{pc});
Slice<uint32_t> bt = gc_Machine_Backtrace(m);
{int64_t i = cast<int64_t>(0ULL);for (;((i < len(bt)) && (i < m->stack.depth));i++){
frames = append(frames,Slice<uint32_t>{bt[i]});
}
}strings_Builder b={};
{auto&& tmp188 = frames;
for(int64_t tmp189=0;tmp189<len(tmp188);++tmp189){
auto i=tmp189;auto f=tmp188[tmp189];if ((i > cast<int64_t>(0ULL))) {
strings_Builder_WriteByte(&(b),cast<uint8_t>(32ULL));
}
go_fmt_Fprintf((&b),std::string("%08X",4),f);
}}
std::string k = strings_Builder_String(&(b));
{
gc_StackSample* s = get(m->stack.stacks,k);
if (bool(s)) {
s->Count++;
return ;
}
}
m->stack.stacks[k] = arenaNew(gc_StackSample{frames,cast<int64_t>(1ULL)});
}
}
// tools/platform/gc/stackprof.go:115:1
std::tuple<Slice<gc_StackSample>,int64_t> gc_Machine_StackProfile(gc_Machine* m){
Slice<gc_StackSample> samples{};
int64_t total{};
{
{auto&& tmp190 = m->stack.stacks;
for(auto [tmp191,tmp192]:tmp190){
auto s=tmp192;samples = append(samples,Slice<gc_StackSample>{(*s)});
}}
go_sort_Slice(samples,[&](int64_t i,int64_t j)->bool{
if ((samples[i].Count != samples[j].Count)) {
return (samples[i].Count > samples[j].Count);
}
return (samples[i].Stack[cast<int64_t>(0ULL)] < samples[j].Stack[cast<int64_t>(0ULL)]);
}
);
return {samples,m->stack.samples};
}
}
// tools/platform/gc/stackprof.go:134:1
std::tuple<Slice<gc_StackSample>,int64_t> gc_Machine_StackProfileByCaller(gc_Machine* m,int64_t depth){
Slice<gc_StackSample> samples{};
int64_t total{};
{
Map<uint32_t,int64_t> by = Map<uint32_t,int64_t>{};
{auto&& tmp193 = m->stack.stacks;
for(auto [tmp194,tmp195]:tmp193){
auto s=tmp195;if ((depth >= len(s->Stack))) {
continue;
}
by[s->Stack[depth]] += s->Count;
}}
{auto&& tmp196 = by;
for(auto [tmp197,tmp198]:tmp196){
auto a=tmp197;auto n=tmp198;samples = append(samples,Slice<gc_StackSample>{gc_StackSample{Slice<uint32_t>{a},n}});
}}
go_sort_Slice(samples,[&](int64_t i,int64_t j)->bool{
if ((samples[i].Count != samples[j].Count)) {
return (samples[i].Count > samples[j].Count);
}
return (samples[i].Stack[cast<int64_t>(0ULL)] < samples[j].Stack[cast<int64_t>(0ULL)]);
}
);
return {samples,m->stack.samples};
}
}
// tools/platform/gc/stackprof.go:155:1
std::string gc_Machine_StackProfileString(gc_Machine* m,int64_t n){
{
auto tmp199 = gc_Machine_StackProfile(m);
Slice<gc_StackSample> samples = std::get<0>(tmp199);
int64_t total = std::get<1>(tmp199);
if ((total == cast<int64_t>(0ULL))) {
return std::string("  (no stack samples; is the profiler on, and did the run cover any instructions?)\012",82);
}
strings_Builder b={};
go_fmt_Fprintf((&b),std::string("%d stack samples, every %d instructions \342\200\224 %d distinct stacks\012",63),total,m->stack.every,len(samples));
{auto&& tmp200 = samples;
for(int64_t tmp201=0;tmp201<len(tmp200);++tmp201){
auto i=tmp201;auto s=tmp200[tmp201];if ((i >= n)) {
break;
}
go_fmt_Fprintf((&b),std::string("%6.2f%%  ",9),((cast<double>(s.Count) / cast<double>(total)) * cast<double>(100ULL)));
{auto&& tmp202 = s.Stack;
for(int64_t tmp203=0;tmp203<len(tmp202);++tmp203){
auto j=tmp203;auto f=tmp202[tmp203];if ((j > cast<int64_t>(0ULL))) {
strings_Builder_WriteString(&(b),std::string(" < ",3));
}
go_fmt_Fprintf((&b),std::string("0x%08X",6),f);
}}
strings_Builder_WriteByte(&(b),cast<uint8_t>(10ULL));
}}
return strings_Builder_String(&(b));
}
}
// tools/platform/gc/vi.go:28:1
void gc_vi_init(gc_vi* v){
{
}
}
// tools/platform/gc/vi.go:33:1
uint32_t gc_vi_read(gc_vi* v,gc_Machine* m,uint32_t off,int64_t size){
{
uint32_t r = cast<uint32_t>((off & cast<uint32_t>(4095ULL)));
{
if ((r == cast<uint32_t>(2ULL))){
return v->DCR;
}
else if (((r == cast<uint32_t>(28ULL)) || (r == cast<uint32_t>(30ULL)))){
return gc_halfword(v->TFBL,r,size);
}
else if (((r == cast<uint32_t>(36ULL)) || (r == cast<uint32_t>(38ULL)))){
return gc_halfword(v->BFBL,r,size);
}
else if ((r == cast<uint32_t>(44ULL))){
return v->Line;
}
else if (((r >= cast<uint32_t>(48ULL)) && (r < cast<uint32_t>(64ULL)))){
uint32_t full = v->DI[divi<uint32_t>((cast<uint32_t>((r - cast<uint32_t>(48ULL)))),cast<uint32_t>(4ULL))];
if ((size == cast<int64_t>(2ULL))) {
if ((cast<uint32_t>((r & cast<uint32_t>(2ULL))) == cast<uint32_t>(0ULL))) {
return shr<uint32_t>(full,cast<int64_t>(16ULL));
}
return cast<uint32_t>((full & cast<uint32_t>(65535ULL)));
}
return full;
}
}
tmp204:;
gc_Machine_logf(m,std::string("VI read unmodelled 0x%03X",25),r);
return cast<uint32_t>(0ULL);
}
}
// tools/platform/gc/vi.go:65:1
void gc_vi_write(gc_vi* v,gc_Machine* m,uint32_t off,uint32_t val,int64_t size){
{
uint32_t r = cast<uint32_t>((off & cast<uint32_t>(4095ULL)));
{
if ((r == cast<uint32_t>(2ULL))){
v->DCR = val;
}
else if (((r == cast<uint32_t>(28ULL)) || (r == cast<uint32_t>(30ULL)))){
v->TFBL = gc_composeHalfword(v->TFBL,val,r,size);
}
else if (((r == cast<uint32_t>(36ULL)) || (r == cast<uint32_t>(38ULL)))){
v->BFBL = gc_composeHalfword(v->BFBL,val,r,size);
}
else if (((r >= cast<uint32_t>(48ULL)) && (r < cast<uint32_t>(64ULL)))){
uint32_t i = divi<uint32_t>((cast<uint32_t>((r - cast<uint32_t>(48ULL)))),cast<uint32_t>(4ULL));
{
if (((size == cast<int64_t>(2ULL)) && (cast<uint32_t>((r & cast<uint32_t>(2ULL))) == cast<uint32_t>(0ULL)))){
v->DI[i] = cast<uint32_t>(((cast<uint32_t>((v->DI[i] & cast<uint32_t>(65535ULL)))) | shl<uint32_t>((cast<uint32_t>((val & cast<uint32_t>(65535ULL)))),cast<int64_t>(16ULL))));
}
else if ((size == cast<int64_t>(2ULL))){
v->DI[i] = cast<uint32_t>(((cast<uint32_t>((v->DI[i] & cast<uint32_t>(4294901760ULL)))) | (cast<uint32_t>((val & cast<uint32_t>(65535ULL))))));
}
else {
v->DI[i] = val;
}
}
tmp206:;
if ((cast<uint32_t>((v->DI[i] & cast<uint32_t>(2147483648ULL))) == cast<uint32_t>(0ULL))) {
gc_Machine_viRefreshIRQ(m);
}
}
else {
}
}
tmp205:;
}
}
// tools/platform/gc/vi.go:109:1
uint32_t gc_composeHalfword(uint32_t reg,uint32_t val,uint32_t r,int64_t size){
{
if ((size != cast<int64_t>(2ULL))) {
return val;
}
if ((cast<uint32_t>((r & cast<uint32_t>(2ULL))) == cast<uint32_t>(0ULL))) {
return cast<uint32_t>(((cast<uint32_t>((reg & cast<uint32_t>(65535ULL)))) | shl<uint32_t>((cast<uint32_t>((val & cast<uint32_t>(65535ULL)))),cast<int64_t>(16ULL))));
}
return cast<uint32_t>(((cast<uint32_t>((reg & cast<uint32_t>(4294901760ULL)))) | (cast<uint32_t>((val & cast<uint32_t>(65535ULL))))));
}
}
// tools/platform/gc/vi.go:120:1
uint32_t gc_halfword(uint32_t reg,uint32_t r,int64_t size){
{
if ((size != cast<int64_t>(2ULL))) {
return reg;
}
if ((cast<uint32_t>((r & cast<uint32_t>(2ULL))) == cast<uint32_t>(0ULL))) {
return shr<uint32_t>(reg,cast<int64_t>(16ULL));
}
return cast<uint32_t>((reg & cast<uint32_t>(65535ULL)));
}
}
// tools/platform/gc/vi.go:133:1
uint32_t gc_vi_XFBAddr(gc_vi* v){
{
uint32_t a = cast<uint32_t>((v->TFBL & cast<uint32_t>(16777215ULL)));
if ((cast<uint32_t>((v->TFBL & cast<uint32_t>(268435456ULL))) != cast<uint32_t>(0ULL))) {
a = shl<uint32_t>(a,cast<int64_t>(5ULL));
}
return a;
}
}
// tools/platform/gc/vi.go:149:1
void gc_Machine_tickVI(gc_Machine* m){
{
m->vi.Counter++;
if ((m->vi.Counter < cast<uint32_t>(8100000ULL))) {
return ;
}
m->vi.Counter = cast<uint32_t>(0ULL);
m->vi.Field++;
m->vi.Line = cast<uint32_t>(0ULL);
bool fired = false;
{auto&& tmp207 = m->vi.DI;
for(int64_t tmp208=0;tmp208<len(tmp207);++tmp208){
auto i=tmp208;if ((cast<uint32_t>((m->vi.DI[i] & cast<uint32_t>(268435456ULL))) != cast<uint32_t>(0ULL))) {
m->vi.DI[i] |= cast<uint32_t>(2147483648ULL);
fired = true;
}
}}
if (fired) {
gc_Machine_raiseInt(m,cast<int64_t>(8ULL));
}
gc_Machine_tickSI(m);
if (bool(m->OnDisplay)) {
m->OnDisplay(m);
}
}
}
// tools/platform/gc/vi.go:180:1
void gc_Machine_viRefreshIRQ(gc_Machine* m){
{
{auto&& tmp209 = m->vi.DI;
for(int64_t tmp210=0;tmp210<len(tmp209);++tmp210){
auto d=tmp209[tmp210];if (((cast<uint32_t>((d & cast<uint32_t>(268435456ULL))) != cast<uint32_t>(0ULL)) && (cast<uint32_t>((d & cast<uint32_t>(2147483648ULL))) != cast<uint32_t>(0ULL)))) {
return ;
}
}}
gc_Machine_clearInt(m,cast<int64_t>(8ULL));
}
}

#include "fast.h"
