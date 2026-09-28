#include "runtime.h"
struct r4300_TLBEntry;
struct r4300_CPU;
struct rsp_CPU;
struct n64_ai;
struct n64_BootConfig;
struct n64_isv;
struct n64_Machine;
struct n64_mi;
struct n64_pi;
struct n64_rdpImage;
struct n64_tile;
struct n64_rdp;
struct n64_rdpStop;
struct n64_combineInputs;
struct n64_rgba;
struct n64_uint32Vec3;
struct n64_combinerSelects;
struct n64_PixelEvent;
struct n64_triAttrs;
struct n64_Result;
struct n64_runState;
struct n64_si;
struct n64_Controller;
struct n64_vi;
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
struct Anon3;
struct Anon4;
struct Anon5;
struct Anon6;
struct Anon7;
struct Anon8;
struct Anon9;
struct n64_ROM {Slice<uint8_t> Data;std::string MD5;};
struct r4300_TLBEntry{
uint32_t PageMask{};
uint64_t EntryHi{};
uint64_t EntryLo0{};
uint64_t EntryLo1{};
};
struct r4300_CPU{
std::array<uint64_t,32> R{};
uint64_t HI{};
uint64_t LO{};
uint64_t PC{};
uint64_t nextPC{};
std::array<uint64_t,32> COP0{};
std::array<r4300_TLBEntry,32> TLB{};
std::array<uint64_t,32> FGR{};
uint32_t FCR31{};
uint64_t COP2Latch{};
bool LLBit{};
bool Halted{};
std::string HaltReason{};
uint64_t Steps{};
n64_Machine* bus{};
std::function<uint32_t(uint32_t)> fetch{};
uint64_t curPC{};
bool delaySlot{};
bool pendingDelay{};
uint64_t branchAddr{};
uint64_t countFrac{};
};
struct rsp_CPU{
std::array<uint32_t,32> R{};
std::array<std::array<uint16_t,8>,32> V{};
std::array<uint64_t,8> Acc{};
uint16_t VCO{};
uint16_t VCC{};
uint8_t VCE{};
uint16_t divIn{};
uint16_t divOut{};
bool divInLoaded{};
Slice<uint8_t> DMEM{};
Slice<uint8_t> IMEM{};
uint32_t PC{};
uint32_t nextPC{};
bool Halted{};
bool Broke{};
std::string HaltReason{};
uint64_t Steps{};
Map<uint32_t,int64_t> Unimplemented{};
n64_Machine* regs{};
uint32_t curPC{};
};
using n64_regFile=Map<uint32_t,uint32_t>;
struct n64_ai{
n64_regFile Regs{};
};
struct n64_BootConfig{
uint32_t TVType{};
uint32_t ResetType{};
uint32_t OSVersion{};
};
struct n64_isv{
std::array<uint8_t,512> Buf{};
Slice<std::string> Lines{};
};
struct n64_mi{
uint32_t InitMode{};
uint32_t Intr{};
uint32_t Mask{};
};
struct n64_pi{
uint32_t DramAddr{};
uint32_t CartAddr{};
uint32_t Status{};
n64_regFile Regs{};
};
struct n64_vi{
n64_regFile Regs{};
uint64_t Acc{};
uint32_t Current{};
};
struct n64_si{
n64_regFile Regs{};
uint32_t DramAddr{};
};
struct n64_rdpImage{
uint32_t Format{};
uint32_t Size{};
uint32_t Width{};
uint32_t Addr{};
};
struct Anon12{uint32_t XH{};uint32_t YH{};uint32_t XL{};uint32_t YL{};};
struct n64_tile{
uint32_t Format{};
uint32_t Size{};
uint32_t Line{};
uint32_t TMem{};
uint32_t Palette{};
uint32_t CMS{};
uint32_t CMT{};
uint32_t MaskS{};
uint32_t MaskT{};
uint32_t ShiftS{};
uint32_t ShiftT{};
uint32_t SL{};
uint32_t TL{};
uint32_t SH{};
uint32_t TH{};
};
struct n64_rdp{
n64_rdpImage Color{};
n64_rdpImage Texture{};
uint32_t Mask{};
Anon12 Scissor{};
uint64_t OtherModes{};
uint64_t Combine{};
uint32_t FillColor{};
uint32_t FogColor{};
uint32_t BlendColor{};
uint32_t PrimColor{};
uint32_t EnvColor{};
uint32_t PrimDepth{};
std::array<n64_tile,8> Tiles{};
std::array<uint8_t,4096> TMem{};
Slice<uint64_t> Pending{};
};
struct n64_Controller{
bool Present{};
uint16_t Buttons{};
int8_t StickX{};
int8_t StickY{};
};
struct n64_runState{
Map<uint32_t,bool> breakpoints{};
};
struct n64_Machine{
Slice<uint8_t> RDRAM{};
Slice<uint8_t> DMEM{};
Slice<uint8_t> IMEM{};
Slice<uint8_t> ROM{};
Slice<uint8_t> PIF{};
std::string romMD5{};
r4300_CPU* CPU{};
rsp_CPU* RSP{};
n64_mi mi{};
n64_pi pi{};
n64_vi vi{};
n64_ai ai{};
n64_si si{};
n64_isv isv{};
n64_rdp rdp{};
n64_regFile ri{};
n64_regFile rd{};
n64_regFile sp{};
uint32_t spPC{};
n64_regFile dp{};
std::array<n64_Controller,4> Controllers{};
Slice<uint8_t> EEPROM{};
n64_runState run{};
bool noSpin{};
bool rspRunning{};
uint64_t rspSteps{};
uint64_t rdpWords{};
Slice<std::string> Log{};
Map<std::string,bool> logSeen{};
uint32_t WatchLo{};
uint32_t WatchHi{};
std::function<void(uint32_t,uint32_t,uint32_t)> OnWrite{};
std::function<void(n64_Machine*,uint32_t)> OnStep{};
uint32_t RWatchLo{};
uint32_t RWatchHi{};
std::function<void(uint32_t,uint32_t,uint32_t)> OnRead{};
std::function<void(std::string,uint32_t,uint32_t,uint32_t)> OnDMA{};
std::function<void(n64_Machine*)> OnDisplay{};
std::function<void(n64_Machine*,uint32_t)> OnRSPTask{};
std::function<void(uint32_t,uint32_t,uint32_t)> OnAIBuffer{};
std::function<void(n64_Machine*,uint32_t,Slice<uint64_t>)> OnRDPCmd{};
std::function<void(uint32_t,uint32_t,n64_PixelEvent)> OnPixel{};
std::function<void(n64_Machine*,std::string)> OnPrint{};
uint64_t ContPolls{};
Map<uint8_t,uint64_t> JoybusCmds{};
uint64_t SIWrites{};
uint64_t SIReads{};
bool StopRequested{};
int64_t rdpStopAt{};
int64_t rdpCount{};
std::function<void(uint32_t,uint32_t)> dpWriteHook{};
bool hookMuted{};
};
struct n64_rdpStop{
};
struct n64_rgba{
uint32_t R{};
uint32_t G{};
uint32_t B{};
uint32_t A{};
};
struct n64_combineInputs{
n64_rgba Texel0{};
n64_rgba Texel1{};
n64_rgba Shade{};
n64_rgba Prim{};
n64_rgba Env{};
n64_rgba Key{};
n64_rgba Comb{};
uint32_t LODFrac{};
uint32_t PrimLODFrac{};
int32_t texS{};
int32_t texT{};
};
struct n64_uint32Vec3{
uint32_t R{};
uint32_t G{};
uint32_t B{};
};
struct n64_combinerSelects{
uint32_t subAC{};
uint32_t subBC{};
uint32_t mulC{};
uint32_t addC{};
uint32_t subAA{};
uint32_t subBA{};
uint32_t mulA{};
uint32_t addA{};
};
struct n64_PixelEvent{
bool Drawn{};
bool ZReject{};
bool AlphaReject{};
uint32_t R{};
uint32_t G{};
uint32_t B{};
uint32_t A{};
int64_t Z{};
uint32_t TexR{};
uint32_t TexG{};
uint32_t TexB{};
uint32_t TexA{};
int32_t TexS{};
int32_t TexT{};
};
struct n64_triAttrs{
int64_t base{};
int64_t dx{};
int64_t de{};
};
struct n64_Result{
uint64_t Steps{};
uint32_t PC{};
std::string Reason{};
};
struct Anon18{uint32_t base{};uint32_t shift{};};
constexpr int rsp_DMEMSize=4096,rsp_IMEMSize=4096,n64_IPL3End=4096;

#include "adapters-decl.h"
constexpr int64_t r4300_TLBSize=32ULL;
constexpr int64_t r4300_entryLoG=1ULL;
constexpr int64_t r4300_entryLoV=2ULL;
constexpr int64_t r4300_entryLoD=4ULL;
constexpr int64_t r4300_kuseg=0ULL;
constexpr int64_t r4300_kseg0=2147483648ULL;
constexpr int64_t r4300_kseg1=2684354560ULL;
constexpr int64_t r4300_ksseg=3221225472ULL;
constexpr int64_t r4300_kseg3=3758096384ULL;
constexpr int64_t r4300_cop0Index=0ULL;
constexpr int64_t r4300_cop0Random=1ULL;
constexpr int64_t r4300_cop0EntryLo0=2ULL;
constexpr int64_t r4300_cop0EntryLo1=3ULL;
constexpr int64_t r4300_cop0Context=4ULL;
constexpr int64_t r4300_cop0PageMask=5ULL;
constexpr int64_t r4300_cop0Wired=6ULL;
constexpr int64_t r4300_cop0BadVAddr=8ULL;
constexpr int64_t r4300_cop0Count=9ULL;
constexpr int64_t r4300_cop0EntryHi=10ULL;
constexpr int64_t r4300_cop0Compare=11ULL;
constexpr int64_t r4300_cop0Status=12ULL;
constexpr int64_t r4300_cop0Cause=13ULL;
constexpr int64_t r4300_cop0EPC=14ULL;
constexpr int64_t r4300_cop0PRId=15ULL;
constexpr int64_t r4300_cop0Config=16ULL;
constexpr int64_t r4300_cop0LLAddr=17ULL;
constexpr int64_t r4300_cop0XContext=20ULL;
constexpr int64_t r4300_cop0ErrorEPC=30ULL;
constexpr int64_t r4300_statusIE=1ULL;
constexpr int64_t r4300_statusEXL=2ULL;
constexpr int64_t r4300_statusERL=4ULL;
constexpr int64_t r4300_statusBEV=4194304ULL;
constexpr int64_t r4300_statusFR=67108864ULL;
constexpr int64_t r4300_statusCU0=268435456ULL;
constexpr int64_t r4300_statusCU1=536870912ULL;
constexpr int64_t r4300_statusCU2=1073741824ULL;
constexpr int64_t r4300_causeBD=2147483648ULL;
constexpr int64_t r4300_causeIP2=1024ULL;
constexpr int64_t r4300_causeIP7=32768ULL;
constexpr int64_t r4300_excInt=0ULL;
constexpr int64_t r4300_excMod=1ULL;
constexpr int64_t r4300_excTLBL=2ULL;
constexpr int64_t r4300_excTLBS=3ULL;
constexpr int64_t r4300_excAdEL=4ULL;
constexpr int64_t r4300_excAdES=5ULL;
constexpr int64_t r4300_excSys=8ULL;
constexpr int64_t r4300_excBp=9ULL;
constexpr int64_t r4300_excRI=10ULL;
constexpr int64_t r4300_excCpU=11ULL;
constexpr int64_t r4300_excOv=12ULL;
constexpr int64_t r4300_excTrap=13ULL;
constexpr int64_t r4300_excFPE=15ULL;
constexpr int64_t r4300_vecRAM=2147483648ULL;
constexpr int64_t r4300_vecROM=3217031680ULL;
constexpr int64_t r4300_Cop0Count=9ULL;
constexpr int64_t r4300_Cop0Compare=11ULL;
constexpr int64_t r4300_Cop0Status=12ULL;
constexpr int64_t r4300_Cop0Cause=13ULL;
constexpr int64_t r4300_Cop0EPC=14ULL;
constexpr int64_t r4300_Cop0BadVAddr=8ULL;
constexpr int64_t r4300_StatusIE=1ULL;
constexpr int64_t r4300_StatusEXL=2ULL;
constexpr int64_t r4300_StatusERL=4ULL;
constexpr int64_t r4300_StatusBEV=4194304ULL;
constexpr int64_t r4300_StatusFR=67108864ULL;
constexpr int64_t r4300_StatusCU0=268435456ULL;
constexpr int64_t r4300_StatusCU1=536870912ULL;
constexpr int64_t r4300_StatusCU2=1073741824ULL;
constexpr int64_t r4300_CauseIP2=1024ULL;
constexpr int64_t r4300_CauseIP7=32768ULL;
constexpr int64_t r4300_fcr31FS=16777216ULL;
constexpr int64_t r4300_qNaN32=2143289343ULL;
constexpr int64_t r4300_qNaN64=9221120237041090559ULL;
constexpr int64_t r4300_fcr31Cond=8388608ULL;
constexpr int64_t r4300_fcr31CauseUnimpl=131072ULL;
constexpr int64_t r4300_fcr31CauseMask=258048ULL;
constexpr int64_t r4300_roundNearest=0ULL;
constexpr int64_t r4300_roundToZero=1ULL;
constexpr int64_t r4300_roundCeil=2ULL;
constexpr int64_t r4300_roundFloor=3ULL;
constexpr int64_t r4300_fpInexact=1ULL;
constexpr int64_t r4300_fpUnderflow=2ULL;
constexpr int64_t r4300_fpOverflow=4ULL;
constexpr int64_t r4300_fpDivByZero=8ULL;
constexpr int64_t r4300_fpInvalid=16ULL;
constexpr int64_t r4300_fpUnimpl=32ULL;
constexpr int64_t r4300_fpTiny=64ULL;
constexpr int64_t r4300_fcr31FlagShift=2ULL;
constexpr int64_t r4300_fcr31EnableShift=7ULL;
constexpr int64_t r4300_fcr31CauseShift=12ULL;
std::array<uint16_t,512> rsp_rcpROM=std::array<uint16_t,512>{cast<uint16_t>(65535ULL),cast<uint16_t>(65280ULL),cast<uint16_t>(65025ULL),cast<uint16_t>(64772ULL),cast<uint16_t>(64519ULL),cast<uint16_t>(64268ULL),cast<uint16_t>(64017ULL),cast<uint16_t>(63768ULL),cast<uint16_t>(63519ULL),cast<uint16_t>(63271ULL),cast<uint16_t>(63025ULL),cast<uint16_t>(62779ULL),cast<uint16_t>(62534ULL),cast<uint16_t>(62290ULL),cast<uint16_t>(62047ULL),cast<uint16_t>(61805ULL),cast<uint16_t>(61564ULL),cast<uint16_t>(61323ULL),cast<uint16_t>(61084ULL),cast<uint16_t>(60846ULL),cast<uint16_t>(60608ULL),cast<uint16_t>(60371ULL),cast<uint16_t>(60136ULL),cast<uint16_t>(59901ULL),cast<uint16_t>(59667ULL),cast<uint16_t>(59433ULL),cast<uint16_t>(59201ULL),cast<uint16_t>(58970ULL),cast<uint16_t>(58739ULL),cast<uint16_t>(58509ULL),cast<uint16_t>(58281ULL),cast<uint16_t>(58053ULL),cast<uint16_t>(57825ULL),cast<uint16_t>(57599ULL),cast<uint16_t>(57374ULL),cast<uint16_t>(57149ULL),cast<uint16_t>(56925ULL),cast<uint16_t>(56702ULL),cast<uint16_t>(56480ULL),cast<uint16_t>(56258ULL),cast<uint16_t>(56038ULL),cast<uint16_t>(55818ULL),cast<uint16_t>(55599ULL),cast<uint16_t>(55380ULL),cast<uint16_t>(55163ULL),cast<uint16_t>(54946ULL),cast<uint16_t>(54730ULL),cast<uint16_t>(54515ULL),cast<uint16_t>(54301ULL),cast<uint16_t>(54087ULL),cast<uint16_t>(53874ULL),cast<uint16_t>(53662ULL),cast<uint16_t>(53451ULL),cast<uint16_t>(53240ULL),cast<uint16_t>(53030ULL),cast<uint16_t>(52821ULL),cast<uint16_t>(52613ULL),cast<uint16_t>(52405ULL),cast<uint16_t>(52198ULL),cast<uint16_t>(51992ULL),cast<uint16_t>(51787ULL),cast<uint16_t>(51582ULL),cast<uint16_t>(51378ULL),cast<uint16_t>(51175ULL),cast<uint16_t>(50972ULL),cast<uint16_t>(50770ULL),cast<uint16_t>(50569ULL),cast<uint16_t>(50368ULL),cast<uint16_t>(50168ULL),cast<uint16_t>(49969ULL),cast<uint16_t>(49771ULL),cast<uint16_t>(49573ULL),cast<uint16_t>(49376ULL),cast<uint16_t>(49180ULL),cast<uint16_t>(48984ULL),cast<uint16_t>(48789ULL),cast<uint16_t>(48594ULL),cast<uint16_t>(48400ULL),cast<uint16_t>(48207ULL),cast<uint16_t>(48015ULL),cast<uint16_t>(47823ULL),cast<uint16_t>(47632ULL),cast<uint16_t>(47441ULL),cast<uint16_t>(47252ULL),cast<uint16_t>(47062ULL),cast<uint16_t>(46874ULL),cast<uint16_t>(46686ULL),cast<uint16_t>(46498ULL),cast<uint16_t>(46312ULL),cast<uint16_t>(46126ULL),cast<uint16_t>(45940ULL),cast<uint16_t>(45755ULL),cast<uint16_t>(45571ULL),cast<uint16_t>(45387ULL),cast<uint16_t>(45204ULL),cast<uint16_t>(45022ULL),cast<uint16_t>(44840ULL),cast<uint16_t>(44659ULL),cast<uint16_t>(44478ULL),cast<uint16_t>(44298ULL),cast<uint16_t>(44119ULL),cast<uint16_t>(43940ULL),cast<uint16_t>(43761ULL),cast<uint16_t>(43584ULL),cast<uint16_t>(43406ULL),cast<uint16_t>(43230ULL),cast<uint16_t>(43054ULL),cast<uint16_t>(42878ULL),cast<uint16_t>(42704ULL),cast<uint16_t>(42529ULL),cast<uint16_t>(42356ULL),cast<uint16_t>(42182ULL),cast<uint16_t>(42010ULL),cast<uint16_t>(41838ULL),cast<uint16_t>(41666ULL),cast<uint16_t>(41495ULL),cast<uint16_t>(41325ULL),cast<uint16_t>(41155ULL),cast<uint16_t>(40986ULL),cast<uint16_t>(40817ULL),cast<uint16_t>(40648ULL),cast<uint16_t>(40481ULL),cast<uint16_t>(40313ULL),cast<uint16_t>(40147ULL),cast<uint16_t>(39981ULL),cast<uint16_t>(39815ULL),cast<uint16_t>(39650ULL),cast<uint16_t>(39485ULL),cast<uint16_t>(39321ULL),cast<uint16_t>(39158ULL),cast<uint16_t>(38994ULL),cast<uint16_t>(38832ULL),cast<uint16_t>(38670ULL),cast<uint16_t>(38508ULL),cast<uint16_t>(38347ULL),cast<uint16_t>(38187ULL),cast<uint16_t>(38027ULL),cast<uint16_t>(37867ULL),cast<uint16_t>(37708ULL),cast<uint16_t>(37549ULL),cast<uint16_t>(37391ULL),cast<uint16_t>(37234ULL),cast<uint16_t>(37076ULL),cast<uint16_t>(36920ULL),cast<uint16_t>(36764ULL),cast<uint16_t>(36608ULL),cast<uint16_t>(36453ULL),cast<uint16_t>(36298ULL),cast<uint16_t>(36144ULL),cast<uint16_t>(35990ULL),cast<uint16_t>(35836ULL),cast<uint16_t>(35684ULL),cast<uint16_t>(35531ULL),cast<uint16_t>(35379ULL),cast<uint16_t>(35228ULL),cast<uint16_t>(35076ULL),cast<uint16_t>(34926ULL),cast<uint16_t>(34776ULL),cast<uint16_t>(34626ULL),cast<uint16_t>(34477ULL),cast<uint16_t>(34328ULL),cast<uint16_t>(34179ULL),cast<uint16_t>(34032ULL),cast<uint16_t>(33884ULL),cast<uint16_t>(33737ULL),cast<uint16_t>(33590ULL),cast<uint16_t>(33444ULL),cast<uint16_t>(33298ULL),cast<uint16_t>(33153ULL),cast<uint16_t>(33008ULL),cast<uint16_t>(32864ULL),cast<uint16_t>(32720ULL),cast<uint16_t>(32576ULL),cast<uint16_t>(32433ULL),cast<uint16_t>(32290ULL),cast<uint16_t>(32147ULL),cast<uint16_t>(32005ULL),cast<uint16_t>(31864ULL),cast<uint16_t>(31723ULL),cast<uint16_t>(31582ULL),cast<uint16_t>(31442ULL),cast<uint16_t>(31302ULL),cast<uint16_t>(31162ULL),cast<uint16_t>(31023ULL),cast<uint16_t>(30884ULL),cast<uint16_t>(30746ULL),cast<uint16_t>(30608ULL),cast<uint16_t>(30470ULL),cast<uint16_t>(30333ULL),cast<uint16_t>(30197ULL),cast<uint16_t>(30060ULL),cast<uint16_t>(29924ULL),cast<uint16_t>(29789ULL),cast<uint16_t>(29653ULL),cast<uint16_t>(29519ULL),cast<uint16_t>(29384ULL),cast<uint16_t>(29250ULL),cast<uint16_t>(29116ULL),cast<uint16_t>(28983ULL),cast<uint16_t>(28850ULL),cast<uint16_t>(28718ULL),cast<uint16_t>(28585ULL),cast<uint16_t>(28454ULL),cast<uint16_t>(28322ULL),cast<uint16_t>(28191ULL),cast<uint16_t>(28060ULL),cast<uint16_t>(27930ULL),cast<uint16_t>(27800ULL),cast<uint16_t>(27670ULL),cast<uint16_t>(27541ULL),cast<uint16_t>(27412ULL),cast<uint16_t>(27284ULL),cast<uint16_t>(27155ULL),cast<uint16_t>(27027ULL),cast<uint16_t>(26900ULL),cast<uint16_t>(26773ULL),cast<uint16_t>(26646ULL),cast<uint16_t>(26520ULL),cast<uint16_t>(26393ULL),cast<uint16_t>(26268ULL),cast<uint16_t>(26142ULL),cast<uint16_t>(26017ULL),cast<uint16_t>(25892ULL),cast<uint16_t>(25768ULL),cast<uint16_t>(25644ULL),cast<uint16_t>(25520ULL),cast<uint16_t>(25397ULL),cast<uint16_t>(25274ULL),cast<uint16_t>(25151ULL),cast<uint16_t>(25029ULL),cast<uint16_t>(24907ULL),cast<uint16_t>(24785ULL),cast<uint16_t>(24664ULL),cast<uint16_t>(24543ULL),cast<uint16_t>(24422ULL),cast<uint16_t>(24301ULL),cast<uint16_t>(24181ULL),cast<uint16_t>(24061ULL),cast<uint16_t>(23942ULL),cast<uint16_t>(23823ULL),cast<uint16_t>(23704ULL),cast<uint16_t>(23586ULL),cast<uint16_t>(23467ULL),cast<uint16_t>(23349ULL),cast<uint16_t>(23232ULL),cast<uint16_t>(23115ULL),cast<uint16_t>(22998ULL),cast<uint16_t>(22881ULL),cast<uint16_t>(22765ULL),cast<uint16_t>(22649ULL),cast<uint16_t>(22533ULL),cast<uint16_t>(22417ULL),cast<uint16_t>(22302ULL),cast<uint16_t>(22188ULL),cast<uint16_t>(22073ULL),cast<uint16_t>(21959ULL),cast<uint16_t>(21845ULL),cast<uint16_t>(21731ULL),cast<uint16_t>(21618ULL),cast<uint16_t>(21505ULL),cast<uint16_t>(21392ULL),cast<uint16_t>(21280ULL),cast<uint16_t>(21167ULL),cast<uint16_t>(21056ULL),cast<uint16_t>(20944ULL),cast<uint16_t>(20833ULL),cast<uint16_t>(20722ULL),cast<uint16_t>(20611ULL),cast<uint16_t>(20501ULL),cast<uint16_t>(20390ULL),cast<uint16_t>(20280ULL),cast<uint16_t>(20171ULL),cast<uint16_t>(20062ULL),cast<uint16_t>(19953ULL),cast<uint16_t>(19844ULL),cast<uint16_t>(19735ULL),cast<uint16_t>(19627ULL),cast<uint16_t>(19519ULL),cast<uint16_t>(19411ULL),cast<uint16_t>(19304ULL),cast<uint16_t>(19197ULL),cast<uint16_t>(19090ULL),cast<uint16_t>(18983ULL),cast<uint16_t>(18877ULL),cast<uint16_t>(18771ULL),cast<uint16_t>(18665ULL),cast<uint16_t>(18560ULL),cast<uint16_t>(18455ULL),cast<uint16_t>(18350ULL),cast<uint16_t>(18245ULL),cast<uint16_t>(18140ULL),cast<uint16_t>(18036ULL),cast<uint16_t>(17932ULL),cast<uint16_t>(17829ULL),cast<uint16_t>(17725ULL),cast<uint16_t>(17622ULL),cast<uint16_t>(17519ULL),cast<uint16_t>(17416ULL),cast<uint16_t>(17314ULL),cast<uint16_t>(17212ULL),cast<uint16_t>(17110ULL),cast<uint16_t>(17008ULL),cast<uint16_t>(16907ULL),cast<uint16_t>(16806ULL),cast<uint16_t>(16705ULL),cast<uint16_t>(16604ULL),cast<uint16_t>(16504ULL),cast<uint16_t>(16404ULL),cast<uint16_t>(16304ULL),cast<uint16_t>(16204ULL),cast<uint16_t>(16104ULL),cast<uint16_t>(16005ULL),cast<uint16_t>(15906ULL),cast<uint16_t>(15808ULL),cast<uint16_t>(15709ULL),cast<uint16_t>(15611ULL),cast<uint16_t>(15513ULL),cast<uint16_t>(15415ULL),cast<uint16_t>(15318ULL),cast<uint16_t>(15220ULL),cast<uint16_t>(15123ULL),cast<uint16_t>(15026ULL),cast<uint16_t>(14930ULL),cast<uint16_t>(14833ULL),cast<uint16_t>(14737ULL),cast<uint16_t>(14641ULL),cast<uint16_t>(14546ULL),cast<uint16_t>(14450ULL),cast<uint16_t>(14355ULL),cast<uint16_t>(14260ULL),cast<uint16_t>(14165ULL),cast<uint16_t>(14071ULL),cast<uint16_t>(13976ULL),cast<uint16_t>(13882ULL),cast<uint16_t>(13788ULL),cast<uint16_t>(13695ULL),cast<uint16_t>(13601ULL),cast<uint16_t>(13508ULL),cast<uint16_t>(13415ULL),cast<uint16_t>(13322ULL),cast<uint16_t>(13230ULL),cast<uint16_t>(13137ULL),cast<uint16_t>(13045ULL),cast<uint16_t>(12953ULL),cast<uint16_t>(12862ULL),cast<uint16_t>(12770ULL),cast<uint16_t>(12679ULL),cast<uint16_t>(12588ULL),cast<uint16_t>(12497ULL),cast<uint16_t>(12406ULL),cast<uint16_t>(12316ULL),cast<uint16_t>(12226ULL),cast<uint16_t>(12136ULL),cast<uint16_t>(12046ULL),cast<uint16_t>(11956ULL),cast<uint16_t>(11867ULL),cast<uint16_t>(11778ULL),cast<uint16_t>(11689ULL),cast<uint16_t>(11600ULL),cast<uint16_t>(11512ULL),cast<uint16_t>(11423ULL),cast<uint16_t>(11335ULL),cast<uint16_t>(11247ULL),cast<uint16_t>(11159ULL),cast<uint16_t>(11072ULL),cast<uint16_t>(10984ULL),cast<uint16_t>(10897ULL),cast<uint16_t>(10810ULL),cast<uint16_t>(10724ULL),cast<uint16_t>(10637ULL),cast<uint16_t>(10551ULL),cast<uint16_t>(10464ULL),cast<uint16_t>(10379ULL),cast<uint16_t>(10293ULL),cast<uint16_t>(10207ULL),cast<uint16_t>(10122ULL),cast<uint16_t>(10037ULL),cast<uint16_t>(9952ULL),cast<uint16_t>(9867ULL),cast<uint16_t>(9782ULL),cast<uint16_t>(9698ULL),cast<uint16_t>(9613ULL),cast<uint16_t>(9529ULL),cast<uint16_t>(9445ULL),cast<uint16_t>(9362ULL),cast<uint16_t>(9278ULL),cast<uint16_t>(9195ULL),cast<uint16_t>(9112ULL),cast<uint16_t>(9029ULL),cast<uint16_t>(8946ULL),cast<uint16_t>(8864ULL),cast<uint16_t>(8781ULL),cast<uint16_t>(8699ULL),cast<uint16_t>(8617ULL),cast<uint16_t>(8535ULL),cast<uint16_t>(8453ULL),cast<uint16_t>(8372ULL),cast<uint16_t>(8291ULL),cast<uint16_t>(8210ULL),cast<uint16_t>(8129ULL),cast<uint16_t>(8048ULL),cast<uint16_t>(7967ULL),cast<uint16_t>(7887ULL),cast<uint16_t>(7807ULL),cast<uint16_t>(7726ULL),cast<uint16_t>(7647ULL),cast<uint16_t>(7567ULL),cast<uint16_t>(7487ULL),cast<uint16_t>(7408ULL),cast<uint16_t>(7329ULL),cast<uint16_t>(7250ULL),cast<uint16_t>(7171ULL),cast<uint16_t>(7092ULL),cast<uint16_t>(7014ULL),cast<uint16_t>(6935ULL),cast<uint16_t>(6857ULL),cast<uint16_t>(6779ULL),cast<uint16_t>(6701ULL),cast<uint16_t>(6624ULL),cast<uint16_t>(6546ULL),cast<uint16_t>(6469ULL),cast<uint16_t>(6392ULL),cast<uint16_t>(6315ULL),cast<uint16_t>(6238ULL),cast<uint16_t>(6161ULL),cast<uint16_t>(6084ULL),cast<uint16_t>(6008ULL),cast<uint16_t>(5932ULL),cast<uint16_t>(5856ULL),cast<uint16_t>(5780ULL),cast<uint16_t>(5704ULL),cast<uint16_t>(5629ULL),cast<uint16_t>(5553ULL),cast<uint16_t>(5478ULL),cast<uint16_t>(5403ULL),cast<uint16_t>(5328ULL),cast<uint16_t>(5253ULL),cast<uint16_t>(5179ULL),cast<uint16_t>(5104ULL),cast<uint16_t>(5030ULL),cast<uint16_t>(4956ULL),cast<uint16_t>(4882ULL),cast<uint16_t>(4808ULL),cast<uint16_t>(4735ULL),cast<uint16_t>(4661ULL),cast<uint16_t>(4588ULL),cast<uint16_t>(4515ULL),cast<uint16_t>(4441ULL),cast<uint16_t>(4369ULL),cast<uint16_t>(4296ULL),cast<uint16_t>(4223ULL),cast<uint16_t>(4151ULL),cast<uint16_t>(4079ULL),cast<uint16_t>(4006ULL),cast<uint16_t>(3934ULL),cast<uint16_t>(3863ULL),cast<uint16_t>(3791ULL),cast<uint16_t>(3719ULL),cast<uint16_t>(3648ULL),cast<uint16_t>(3577ULL),cast<uint16_t>(3506ULL),cast<uint16_t>(3435ULL),cast<uint16_t>(3364ULL),cast<uint16_t>(3293ULL),cast<uint16_t>(3223ULL),cast<uint16_t>(3152ULL),cast<uint16_t>(3082ULL),cast<uint16_t>(3012ULL),cast<uint16_t>(2942ULL),cast<uint16_t>(2872ULL),cast<uint16_t>(2802ULL),cast<uint16_t>(2733ULL),cast<uint16_t>(2664ULL),cast<uint16_t>(2594ULL),cast<uint16_t>(2525ULL),cast<uint16_t>(2456ULL),cast<uint16_t>(2387ULL),cast<uint16_t>(2319ULL),cast<uint16_t>(2250ULL),cast<uint16_t>(2182ULL),cast<uint16_t>(2114ULL),cast<uint16_t>(2045ULL),cast<uint16_t>(1977ULL),cast<uint16_t>(1910ULL),cast<uint16_t>(1842ULL),cast<uint16_t>(1774ULL),cast<uint16_t>(1707ULL),cast<uint16_t>(1640ULL),cast<uint16_t>(1572ULL),cast<uint16_t>(1505ULL),cast<uint16_t>(1438ULL),cast<uint16_t>(1372ULL),cast<uint16_t>(1305ULL),cast<uint16_t>(1238ULL),cast<uint16_t>(1172ULL),cast<uint16_t>(1106ULL),cast<uint16_t>(1040ULL),cast<uint16_t>(974ULL),cast<uint16_t>(908ULL),cast<uint16_t>(842ULL),cast<uint16_t>(777ULL),cast<uint16_t>(711ULL),cast<uint16_t>(646ULL),cast<uint16_t>(581ULL),cast<uint16_t>(516ULL),cast<uint16_t>(451ULL),cast<uint16_t>(386ULL),cast<uint16_t>(321ULL),cast<uint16_t>(257ULL),cast<uint16_t>(192ULL),cast<uint16_t>(128ULL),cast<uint16_t>(64ULL)};
std::array<uint16_t,512> rsp_rsqROM=std::array<uint16_t,512>{cast<uint16_t>(65535ULL),cast<uint16_t>(65280ULL),cast<uint16_t>(65026ULL),cast<uint16_t>(64774ULL),cast<uint16_t>(64523ULL),cast<uint16_t>(64274ULL),cast<uint16_t>(64026ULL),cast<uint16_t>(63779ULL),cast<uint16_t>(63534ULL),cast<uint16_t>(63291ULL),cast<uint16_t>(63048ULL),cast<uint16_t>(62807ULL),cast<uint16_t>(62567ULL),cast<uint16_t>(62329ULL),cast<uint16_t>(62092ULL),cast<uint16_t>(61856ULL),cast<uint16_t>(61622ULL),cast<uint16_t>(61389ULL),cast<uint16_t>(61157ULL),cast<uint16_t>(60927ULL),cast<uint16_t>(60697ULL),cast<uint16_t>(60469ULL),cast<uint16_t>(60242ULL),cast<uint16_t>(60017ULL),cast<uint16_t>(59792ULL),cast<uint16_t>(59569ULL),cast<uint16_t>(59347ULL),cast<uint16_t>(59126ULL),cast<uint16_t>(58907ULL),cast<uint16_t>(58688ULL),cast<uint16_t>(58471ULL),cast<uint16_t>(58254ULL),cast<uint16_t>(58039ULL),cast<uint16_t>(57825ULL),cast<uint16_t>(57613ULL),cast<uint16_t>(57401ULL),cast<uint16_t>(57190ULL),cast<uint16_t>(56980ULL),cast<uint16_t>(56772ULL),cast<uint16_t>(56564ULL),cast<uint16_t>(56358ULL),cast<uint16_t>(56153ULL),cast<uint16_t>(55948ULL),cast<uint16_t>(55745ULL),cast<uint16_t>(55543ULL),cast<uint16_t>(55341ULL),cast<uint16_t>(55141ULL),cast<uint16_t>(54942ULL),cast<uint16_t>(54743ULL),cast<uint16_t>(54546ULL),cast<uint16_t>(54350ULL),cast<uint16_t>(54154ULL),cast<uint16_t>(53960ULL),cast<uint16_t>(53766ULL),cast<uint16_t>(53574ULL),cast<uint16_t>(53382ULL),cast<uint16_t>(53191ULL),cast<uint16_t>(53002ULL),cast<uint16_t>(52813ULL),cast<uint16_t>(52625ULL),cast<uint16_t>(52438ULL),cast<uint16_t>(52251ULL),cast<uint16_t>(52066ULL),cast<uint16_t>(51881ULL),cast<uint16_t>(51698ULL),cast<uint16_t>(51515ULL),cast<uint16_t>(51333ULL),cast<uint16_t>(51152ULL),cast<uint16_t>(50972ULL),cast<uint16_t>(50793ULL),cast<uint16_t>(50614ULL),cast<uint16_t>(50436ULL),cast<uint16_t>(50259ULL),cast<uint16_t>(50083ULL),cast<uint16_t>(49908ULL),cast<uint16_t>(49733ULL),cast<uint16_t>(49560ULL),cast<uint16_t>(49387ULL),cast<uint16_t>(49215ULL),cast<uint16_t>(49043ULL),cast<uint16_t>(48873ULL),cast<uint16_t>(48703ULL),cast<uint16_t>(48534ULL),cast<uint16_t>(48365ULL),cast<uint16_t>(48198ULL),cast<uint16_t>(48031ULL),cast<uint16_t>(47864ULL),cast<uint16_t>(47699ULL),cast<uint16_t>(47534ULL),cast<uint16_t>(47370ULL),cast<uint16_t>(47207ULL),cast<uint16_t>(47045ULL),cast<uint16_t>(46883ULL),cast<uint16_t>(46721ULL),cast<uint16_t>(46561ULL),cast<uint16_t>(46401ULL),cast<uint16_t>(46242ULL),cast<uint16_t>(46084ULL),cast<uint16_t>(45926ULL),cast<uint16_t>(45769ULL),cast<uint16_t>(45612ULL),cast<uint16_t>(45457ULL),cast<uint16_t>(45301ULL),cast<uint16_t>(45147ULL),cast<uint16_t>(44993ULL),cast<uint16_t>(44840ULL),cast<uint16_t>(44687ULL),cast<uint16_t>(44535ULL),cast<uint16_t>(44384ULL),cast<uint16_t>(44233ULL),cast<uint16_t>(44083ULL),cast<uint16_t>(43934ULL),cast<uint16_t>(43785ULL),cast<uint16_t>(43637ULL),cast<uint16_t>(43489ULL),cast<uint16_t>(43342ULL),cast<uint16_t>(43196ULL),cast<uint16_t>(43050ULL),cast<uint16_t>(42905ULL),cast<uint16_t>(42760ULL),cast<uint16_t>(42616ULL),cast<uint16_t>(42472ULL),cast<uint16_t>(42329ULL),cast<uint16_t>(42187ULL),cast<uint16_t>(42045ULL),cast<uint16_t>(41904ULL),cast<uint16_t>(41763ULL),cast<uint16_t>(41623ULL),cast<uint16_t>(41483ULL),cast<uint16_t>(41344ULL),cast<uint16_t>(41206ULL),cast<uint16_t>(41068ULL),cast<uint16_t>(40930ULL),cast<uint16_t>(40793ULL),cast<uint16_t>(40657ULL),cast<uint16_t>(40521ULL),cast<uint16_t>(40386ULL),cast<uint16_t>(40251ULL),cast<uint16_t>(40116ULL),cast<uint16_t>(39983ULL),cast<uint16_t>(39849ULL),cast<uint16_t>(39717ULL),cast<uint16_t>(39584ULL),cast<uint16_t>(39452ULL),cast<uint16_t>(39321ULL),cast<uint16_t>(39190ULL),cast<uint16_t>(39060ULL),cast<uint16_t>(38930ULL),cast<uint16_t>(38801ULL),cast<uint16_t>(38672ULL),cast<uint16_t>(38543ULL),cast<uint16_t>(38415ULL),cast<uint16_t>(38288ULL),cast<uint16_t>(38161ULL),cast<uint16_t>(38034ULL),cast<uint16_t>(37908ULL),cast<uint16_t>(37783ULL),cast<uint16_t>(37658ULL),cast<uint16_t>(37533ULL),cast<uint16_t>(37409ULL),cast<uint16_t>(37285ULL),cast<uint16_t>(37161ULL),cast<uint16_t>(37039ULL),cast<uint16_t>(36916ULL),cast<uint16_t>(36794ULL),cast<uint16_t>(36672ULL),cast<uint16_t>(36551ULL),cast<uint16_t>(36431ULL),cast<uint16_t>(36310ULL),cast<uint16_t>(36190ULL),cast<uint16_t>(36071ULL),cast<uint16_t>(35952ULL),cast<uint16_t>(35833ULL),cast<uint16_t>(35715ULL),cast<uint16_t>(35597ULL),cast<uint16_t>(35480ULL),cast<uint16_t>(35363ULL),cast<uint16_t>(35246ULL),cast<uint16_t>(35130ULL),cast<uint16_t>(35014ULL),cast<uint16_t>(34899ULL),cast<uint16_t>(34784ULL),cast<uint16_t>(34669ULL),cast<uint16_t>(34555ULL),cast<uint16_t>(34441ULL),cast<uint16_t>(34328ULL),cast<uint16_t>(34215ULL),cast<uint16_t>(34102ULL),cast<uint16_t>(33990ULL),cast<uint16_t>(33878ULL),cast<uint16_t>(33767ULL),cast<uint16_t>(33655ULL),cast<uint16_t>(33545ULL),cast<uint16_t>(33434ULL),cast<uint16_t>(33324ULL),cast<uint16_t>(33215ULL),cast<uint16_t>(33105ULL),cast<uint16_t>(32996ULL),cast<uint16_t>(32888ULL),cast<uint16_t>(32780ULL),cast<uint16_t>(32672ULL),cast<uint16_t>(32564ULL),cast<uint16_t>(32457ULL),cast<uint16_t>(32350ULL),cast<uint16_t>(32244ULL),cast<uint16_t>(32138ULL),cast<uint16_t>(32032ULL),cast<uint16_t>(31926ULL),cast<uint16_t>(31821ULL),cast<uint16_t>(31717ULL),cast<uint16_t>(31612ULL),cast<uint16_t>(31508ULL),cast<uint16_t>(31404ULL),cast<uint16_t>(31301ULL),cast<uint16_t>(31198ULL),cast<uint16_t>(31095ULL),cast<uint16_t>(30993ULL),cast<uint16_t>(30891ULL),cast<uint16_t>(30789ULL),cast<uint16_t>(30687ULL),cast<uint16_t>(30586ULL),cast<uint16_t>(30485ULL),cast<uint16_t>(30385ULL),cast<uint16_t>(30285ULL),cast<uint16_t>(30185ULL),cast<uint16_t>(30085ULL),cast<uint16_t>(29986ULL),cast<uint16_t>(29887ULL),cast<uint16_t>(29789ULL),cast<uint16_t>(29690ULL),cast<uint16_t>(29592ULL),cast<uint16_t>(29495ULL),cast<uint16_t>(29397ULL),cast<uint16_t>(29300ULL),cast<uint16_t>(29203ULL),cast<uint16_t>(29107ULL),cast<uint16_t>(29010ULL),cast<uint16_t>(28914ULL),cast<uint16_t>(28819ULL),cast<uint16_t>(28723ULL),cast<uint16_t>(28628ULL),cast<uint16_t>(28534ULL),cast<uint16_t>(28439ULL),cast<uint16_t>(28345ULL),cast<uint16_t>(28251ULL),cast<uint16_t>(28157ULL),cast<uint16_t>(28064ULL),cast<uint16_t>(27971ULL),cast<uint16_t>(27878ULL),cast<uint16_t>(27786ULL),cast<uint16_t>(27693ULL),cast<uint16_t>(27601ULL),cast<uint16_t>(27510ULL),cast<uint16_t>(27418ULL),cast<uint16_t>(27327ULL),cast<uint16_t>(27236ULL),cast<uint16_t>(27145ULL),cast<uint16_t>(26965ULL),cast<uint16_t>(26785ULL),cast<uint16_t>(26607ULL),cast<uint16_t>(26430ULL),cast<uint16_t>(26253ULL),cast<uint16_t>(26078ULL),cast<uint16_t>(25904ULL),cast<uint16_t>(25730ULL),cast<uint16_t>(25558ULL),cast<uint16_t>(25387ULL),cast<uint16_t>(25216ULL),cast<uint16_t>(25047ULL),cast<uint16_t>(24878ULL),cast<uint16_t>(24711ULL),cast<uint16_t>(24544ULL),cast<uint16_t>(24378ULL),cast<uint16_t>(24213ULL),cast<uint16_t>(24049ULL),cast<uint16_t>(23886ULL),cast<uint16_t>(23724ULL),cast<uint16_t>(23563ULL),cast<uint16_t>(23403ULL),cast<uint16_t>(23243ULL),cast<uint16_t>(23084ULL),cast<uint16_t>(22927ULL),cast<uint16_t>(22770ULL),cast<uint16_t>(22613ULL),cast<uint16_t>(22458ULL),cast<uint16_t>(22304ULL),cast<uint16_t>(22150ULL),cast<uint16_t>(21997ULL),cast<uint16_t>(21845ULL),cast<uint16_t>(21694ULL),cast<uint16_t>(21543ULL),cast<uint16_t>(21393ULL),cast<uint16_t>(21244ULL),cast<uint16_t>(21096ULL),cast<uint16_t>(20949ULL),cast<uint16_t>(20802ULL),cast<uint16_t>(20656ULL),cast<uint16_t>(20511ULL),cast<uint16_t>(20366ULL),cast<uint16_t>(20222ULL),cast<uint16_t>(20079ULL),cast<uint16_t>(19937ULL),cast<uint16_t>(19795ULL),cast<uint16_t>(19654ULL),cast<uint16_t>(19514ULL),cast<uint16_t>(19375ULL),cast<uint16_t>(19236ULL),cast<uint16_t>(19098ULL),cast<uint16_t>(18960ULL),cast<uint16_t>(18823ULL),cast<uint16_t>(18687ULL),cast<uint16_t>(18552ULL),cast<uint16_t>(18417ULL),cast<uint16_t>(18283ULL),cast<uint16_t>(18149ULL),cast<uint16_t>(18016ULL),cast<uint16_t>(17884ULL),cast<uint16_t>(17752ULL),cast<uint16_t>(17621ULL),cast<uint16_t>(17491ULL),cast<uint16_t>(17361ULL),cast<uint16_t>(17231ULL),cast<uint16_t>(17103ULL),cast<uint16_t>(16975ULL),cast<uint16_t>(16847ULL),cast<uint16_t>(16721ULL),cast<uint16_t>(16594ULL),cast<uint16_t>(16469ULL),cast<uint16_t>(16344ULL),cast<uint16_t>(16219ULL),cast<uint16_t>(16095ULL),cast<uint16_t>(15972ULL),cast<uint16_t>(15849ULL),cast<uint16_t>(15726ULL),cast<uint16_t>(15605ULL),cast<uint16_t>(15484ULL),cast<uint16_t>(15363ULL),cast<uint16_t>(15243ULL),cast<uint16_t>(15123ULL),cast<uint16_t>(15004ULL),cast<uint16_t>(14886ULL),cast<uint16_t>(14768ULL),cast<uint16_t>(14650ULL),cast<uint16_t>(14533ULL),cast<uint16_t>(14417ULL),cast<uint16_t>(14301ULL),cast<uint16_t>(14185ULL),cast<uint16_t>(14070ULL),cast<uint16_t>(13956ULL),cast<uint16_t>(13842ULL),cast<uint16_t>(13728ULL),cast<uint16_t>(13615ULL),cast<uint16_t>(13503ULL),cast<uint16_t>(13391ULL),cast<uint16_t>(13279ULL),cast<uint16_t>(13168ULL),cast<uint16_t>(13058ULL),cast<uint16_t>(12947ULL),cast<uint16_t>(12838ULL),cast<uint16_t>(12729ULL),cast<uint16_t>(12620ULL),cast<uint16_t>(12511ULL),cast<uint16_t>(12404ULL),cast<uint16_t>(12296ULL),cast<uint16_t>(12189ULL),cast<uint16_t>(12083ULL),cast<uint16_t>(11976ULL),cast<uint16_t>(11871ULL),cast<uint16_t>(11766ULL),cast<uint16_t>(11661ULL),cast<uint16_t>(11556ULL),cast<uint16_t>(11452ULL),cast<uint16_t>(11349ULL),cast<uint16_t>(11246ULL),cast<uint16_t>(11143ULL),cast<uint16_t>(11041ULL),cast<uint16_t>(10939ULL),cast<uint16_t>(10837ULL),cast<uint16_t>(10736ULL),cast<uint16_t>(10635ULL),cast<uint16_t>(10535ULL),cast<uint16_t>(10435ULL),cast<uint16_t>(10336ULL),cast<uint16_t>(10237ULL),cast<uint16_t>(10138ULL),cast<uint16_t>(10040ULL),cast<uint16_t>(9942ULL),cast<uint16_t>(9844ULL),cast<uint16_t>(9747ULL),cast<uint16_t>(9650ULL),cast<uint16_t>(9554ULL),cast<uint16_t>(9458ULL),cast<uint16_t>(9362ULL),cast<uint16_t>(9266ULL),cast<uint16_t>(9171ULL),cast<uint16_t>(9077ULL),cast<uint16_t>(8983ULL),cast<uint16_t>(8889ULL),cast<uint16_t>(8795ULL),cast<uint16_t>(8702ULL),cast<uint16_t>(8609ULL),cast<uint16_t>(8517ULL),cast<uint16_t>(8424ULL),cast<uint16_t>(8333ULL),cast<uint16_t>(8241ULL),cast<uint16_t>(8150ULL),cast<uint16_t>(8059ULL),cast<uint16_t>(7969ULL),cast<uint16_t>(7879ULL),cast<uint16_t>(7789ULL),cast<uint16_t>(7699ULL),cast<uint16_t>(7610ULL),cast<uint16_t>(7521ULL),cast<uint16_t>(7433ULL),cast<uint16_t>(7345ULL),cast<uint16_t>(7257ULL),cast<uint16_t>(7169ULL),cast<uint16_t>(7082ULL),cast<uint16_t>(6995ULL),cast<uint16_t>(6908ULL),cast<uint16_t>(6822ULL),cast<uint16_t>(6736ULL),cast<uint16_t>(6650ULL),cast<uint16_t>(6565ULL),cast<uint16_t>(6480ULL),cast<uint16_t>(6395ULL),cast<uint16_t>(6311ULL),cast<uint16_t>(6227ULL),cast<uint16_t>(6143ULL),cast<uint16_t>(6059ULL),cast<uint16_t>(5976ULL),cast<uint16_t>(5893ULL),cast<uint16_t>(5810ULL),cast<uint16_t>(5728ULL),cast<uint16_t>(5645ULL),cast<uint16_t>(5564ULL),cast<uint16_t>(5482ULL),cast<uint16_t>(5401ULL),cast<uint16_t>(5320ULL),cast<uint16_t>(5239ULL),cast<uint16_t>(5158ULL),cast<uint16_t>(5078ULL),cast<uint16_t>(4998ULL),cast<uint16_t>(4919ULL),cast<uint16_t>(4839ULL),cast<uint16_t>(4760ULL),cast<uint16_t>(4681ULL),cast<uint16_t>(4603ULL),cast<uint16_t>(4524ULL),cast<uint16_t>(4446ULL),cast<uint16_t>(4369ULL),cast<uint16_t>(4291ULL),cast<uint16_t>(4214ULL),cast<uint16_t>(4137ULL),cast<uint16_t>(4060ULL),cast<uint16_t>(3983ULL),cast<uint16_t>(3907ULL),cast<uint16_t>(3831ULL),cast<uint16_t>(3755ULL),cast<uint16_t>(3680ULL),cast<uint16_t>(3605ULL),cast<uint16_t>(3530ULL),cast<uint16_t>(3455ULL),cast<uint16_t>(3380ULL),cast<uint16_t>(3306ULL),cast<uint16_t>(3232ULL),cast<uint16_t>(3158ULL),cast<uint16_t>(3084ULL),cast<uint16_t>(3011ULL),cast<uint16_t>(2938ULL),cast<uint16_t>(2865ULL),cast<uint16_t>(2792ULL),cast<uint16_t>(2720ULL),cast<uint16_t>(2648ULL),cast<uint16_t>(2576ULL),cast<uint16_t>(2504ULL),cast<uint16_t>(2433ULL),cast<uint16_t>(2361ULL),cast<uint16_t>(2290ULL),cast<uint16_t>(2219ULL),cast<uint16_t>(2149ULL),cast<uint16_t>(2078ULL),cast<uint16_t>(2008ULL),cast<uint16_t>(1938ULL),cast<uint16_t>(1869ULL),cast<uint16_t>(1799ULL),cast<uint16_t>(1730ULL),cast<uint16_t>(1661ULL),cast<uint16_t>(1592ULL),cast<uint16_t>(1523ULL),cast<uint16_t>(1455ULL),cast<uint16_t>(1386ULL),cast<uint16_t>(1318ULL),cast<uint16_t>(1250ULL),cast<uint16_t>(1183ULL),cast<uint16_t>(1115ULL),cast<uint16_t>(1048ULL),cast<uint16_t>(981ULL),cast<uint16_t>(914ULL),cast<uint16_t>(848ULL),cast<uint16_t>(781ULL),cast<uint16_t>(715ULL),cast<uint16_t>(649ULL),cast<uint16_t>(583ULL),cast<uint16_t>(518ULL),cast<uint16_t>(452ULL),cast<uint16_t>(387ULL),cast<uint16_t>(322ULL),cast<uint16_t>(257ULL),cast<uint16_t>(192ULL),cast<uint16_t>(128ULL),cast<uint16_t>(64ULL)};
constexpr int64_t rsp_accMask=281474976710655ULL;
Map<uint32_t,uint32_t> rsp_sfvStart=Map<uint32_t,uint32_t>{{cast<uint32_t>(0ULL),cast<uint32_t>(0ULL)},{cast<uint32_t>(1ULL),cast<uint32_t>(6ULL)},{cast<uint32_t>(4ULL),cast<uint32_t>(1ULL)},{cast<uint32_t>(5ULL),cast<uint32_t>(7ULL)},{cast<uint32_t>(8ULL),cast<uint32_t>(4ULL)},{cast<uint32_t>(11ULL),cast<uint32_t>(3ULL)},{cast<uint32_t>(12ULL),cast<uint32_t>(5ULL)},{cast<uint32_t>(15ULL),cast<uint32_t>(0ULL)}};
constexpr int64_t n64_aiDramAddr=0ULL;
constexpr int64_t n64_aiLength=4ULL;
constexpr int64_t n64_aiControl=8ULL;
constexpr int64_t n64_aiStatus=12ULL;
constexpr int64_t n64_aiDacRate=16ULL;
constexpr int64_t n64_aiBitRate=20ULL;
constexpr int64_t n64_aiStatusBusy=1073741824ULL;
constexpr int64_t n64_aiStatusFull=2147483648ULL;
constexpr int64_t n64_ipl3Entry=2751463488ULL;
constexpr int64_t n64_ipl2Stack=2751471600ULL;
constexpr int64_t n64_ipl2Return=2751468880ULL;
constexpr int64_t n64_osMemSizeAddr=2147484440ULL;
constexpr int64_t n64_osMemSizePhys=792ULL;
constexpr int64_t n64_riSelect=12ULL;
constexpr int64_t n64_riSelectInitialised=20ULL;
constexpr int64_t n64_TVPAL=0ULL;
constexpr int64_t n64_TVNTSC=1ULL;
constexpr int64_t n64_TVMPAL=2ULL;
constexpr int64_t n64_ResetCold=0ULL;
constexpr int64_t n64_ResetNMI=1ULL;
constexpr int64_t n64_dpStart=0ULL;
constexpr int64_t n64_dpEnd=4ULL;
constexpr int64_t n64_dpCurrent=8ULL;
constexpr int64_t n64_dpStatus=12ULL;
constexpr int64_t n64_dpClock=16ULL;
constexpr int64_t n64_dpBufBusy=20ULL;
constexpr int64_t n64_dpPipeBusy=24ULL;
constexpr int64_t n64_dpTMem=28ULL;
constexpr int64_t n64_dpStatusXBusDMEM=1ULL;
constexpr int64_t n64_dpStatusFreeze=2ULL;
constexpr int64_t n64_dpStatusFlush=4ULL;
constexpr int64_t n64_dpStatusCmdBusy=64ULL;
constexpr int64_t n64_dpStatusCbufReady=128ULL;
constexpr int64_t n64_dpStatusDMABusy=256ULL;
constexpr int64_t n64_dpStatusEndValid=512ULL;
constexpr int64_t n64_dpStatusStartValid=1024ULL;
constexpr int64_t n64_eepromBlockSize=8ULL;
constexpr int64_t n64_eeprom4KBlocks=64ULL;
constexpr int64_t n64_eeprom16KBlocks=256ULL;
constexpr uint16_t n64_devEEPROM4K=128ULL;
constexpr uint16_t n64_devEEPROM16K=192ULL;
constexpr int64_t n64_isvBase=335478784ULL;
constexpr int64_t n64_isvLen=335478804ULL;
constexpr int64_t n64_isvBuffer=335478816ULL;
constexpr int64_t n64_isvEnd=335479328ULL;
constexpr int64_t n64_rdramSize=4194304ULL;
constexpr int64_t n64_rdramEnd=8388608ULL;
constexpr int64_t n64_rdramRegBase=66060288ULL;
constexpr int64_t n64_rdramRegEnd=67108864ULL;
constexpr int64_t n64_spDMEMBase=67108864ULL;
constexpr int64_t n64_spIMEMBase=67112960ULL;
constexpr int64_t n64_spMemEnd=67117056ULL;
constexpr int64_t n64_spRegBase=67371008ULL;
constexpr int64_t n64_spRegEnd=68157440ULL;
constexpr int64_t n64_dpRegBase=68157440ULL;
constexpr int64_t n64_dpRegEnd=70254592ULL;
constexpr int64_t n64_miRegBase=70254592ULL;
constexpr int64_t n64_miRegEnd=71303168ULL;
constexpr int64_t n64_viRegBase=71303168ULL;
constexpr int64_t n64_viRegEnd=72351744ULL;
constexpr int64_t n64_aiRegBase=72351744ULL;
constexpr int64_t n64_aiRegEnd=73400320ULL;
constexpr int64_t n64_piRegBase=73400320ULL;
constexpr int64_t n64_piRegEnd=74448896ULL;
constexpr int64_t n64_riRegBase=74448896ULL;
constexpr int64_t n64_riRegEnd=75497472ULL;
constexpr int64_t n64_siRegBase=75497472ULL;
constexpr int64_t n64_siRegEnd=76546048ULL;
constexpr int64_t n64_cartBase=268435456ULL;
constexpr int64_t n64_cartEnd=532676608ULL;
constexpr int64_t n64_pifBase=532676608ULL;
constexpr int64_t n64_pifEnd=532678656ULL;
constexpr int64_t n64_spMemSize=4096ULL;
constexpr int64_t n64_miInitMode=0ULL;
constexpr int64_t n64_miVersion=4ULL;
constexpr int64_t n64_miIntr=8ULL;
constexpr int64_t n64_miIntrMask=12ULL;
constexpr int64_t n64_intrSP=1ULL;
constexpr int64_t n64_intrSI=2ULL;
constexpr int64_t n64_intrAI=4ULL;
constexpr int64_t n64_intrVI=8ULL;
constexpr int64_t n64_intrPI=16ULL;
constexpr int64_t n64_intrDP=32ULL;
constexpr int64_t n64_miVersionValue=33685762ULL;
constexpr int64_t n64_piDramAddr=0ULL;
constexpr int64_t n64_piCartAddr=4ULL;
constexpr int64_t n64_piRdLen=8ULL;
constexpr int64_t n64_piWrLen=12ULL;
constexpr int64_t n64_piStatus=16ULL;
constexpr int64_t n64_piStatusDMABusy=1ULL;
constexpr int64_t n64_piStatusIOBusy=2ULL;
constexpr int64_t n64_piStatusResetCmd=1ULL;
constexpr int64_t n64_piStatusClearIntrCmd=2ULL;
constexpr int64_t n64_cmdNoOp=0ULL;
constexpr int64_t n64_cmdTriFill=8ULL;
constexpr int64_t n64_cmdTriFillZ=9ULL;
constexpr int64_t n64_cmdTriTex=10ULL;
constexpr int64_t n64_cmdTriTexZ=11ULL;
constexpr int64_t n64_cmdTriShade=12ULL;
constexpr int64_t n64_cmdTriShadeZ=13ULL;
constexpr int64_t n64_cmdTriShadeTex=14ULL;
constexpr int64_t n64_cmdTriShadeTexZ=15ULL;
constexpr int64_t n64_cmdTexRect=36ULL;
constexpr int64_t n64_cmdTexRectFlip=37ULL;
constexpr int64_t n64_cmdSyncLoad=38ULL;
constexpr int64_t n64_cmdSyncPipe=39ULL;
constexpr int64_t n64_cmdSyncTile=40ULL;
constexpr int64_t n64_cmdSyncFull=41ULL;
constexpr int64_t n64_cmdSetKeyGB=42ULL;
constexpr int64_t n64_cmdSetKeyR=43ULL;
constexpr int64_t n64_cmdSetConvert=44ULL;
constexpr int64_t n64_cmdSetScissor=45ULL;
constexpr int64_t n64_cmdSetPrimDepth=46ULL;
constexpr int64_t n64_cmdSetOtherModes=47ULL;
constexpr int64_t n64_cmdLoadTLUT=48ULL;
constexpr int64_t n64_cmdSetTileSize=50ULL;
constexpr int64_t n64_cmdLoadBlock=51ULL;
constexpr int64_t n64_cmdLoadTile=52ULL;
constexpr int64_t n64_cmdSetTile=53ULL;
constexpr int64_t n64_cmdFillRect=54ULL;
constexpr int64_t n64_cmdSetFillColor=55ULL;
constexpr int64_t n64_cmdSetFogColor=56ULL;
constexpr int64_t n64_cmdSetBlendColor=57ULL;
constexpr int64_t n64_cmdSetPrimColor=58ULL;
constexpr int64_t n64_cmdSetEnvColor=59ULL;
constexpr int64_t n64_cmdSetCombineMode=60ULL;
constexpr int64_t n64_cmdSetTextureImage=61ULL;
constexpr int64_t n64_cmdSetMaskImage=62ULL;
constexpr int64_t n64_cmdSetColorImage=63ULL;
std::array<int64_t,64> n64_cmdWords=[](){std::array<int64_t,64> v{};v[8]=cast<int64_t>(4ULL);v[9]=cast<int64_t>(6ULL);v[10]=cast<int64_t>(12ULL);v[11]=cast<int64_t>(14ULL);v[12]=cast<int64_t>(12ULL);v[13]=cast<int64_t>(14ULL);v[14]=cast<int64_t>(20ULL);v[15]=cast<int64_t>(22ULL);v[36]=cast<int64_t>(2ULL);v[37]=cast<int64_t>(2ULL);return v;}();
Map<uint32_t,std::string> n64_cmdName=Map<uint32_t,std::string>{{n64_cmdNoOp,"No_Op"},{n64_cmdTriFill,"Triangle"},{n64_cmdTriFillZ,"Triangle_Z"},{n64_cmdTriTex,"Triangle_Tex"},{n64_cmdTriTexZ,"Triangle_Tex_Z"},{n64_cmdTriShade,"Triangle_Shade"},{n64_cmdTriShadeZ,"Triangle_Shade_Z"},{n64_cmdTriShadeTex,"Triangle_Shade_Tex"},{n64_cmdTriShadeTexZ,"Triangle_Shade_Tex_Z"},{n64_cmdTexRect,"Texture_Rectangle"},{n64_cmdTexRectFlip,"Texture_Rectangle_Flip"},{n64_cmdSyncLoad,"Sync_Load"},{n64_cmdSyncPipe,"Sync_Pipe"},{n64_cmdSyncTile,"Sync_Tile"},{n64_cmdSyncFull,"Sync_Full"},{n64_cmdSetKeyGB,"Set_Key_GB"},{n64_cmdSetKeyR,"Set_Key_R"},{n64_cmdSetConvert,"Set_Convert"},{n64_cmdSetScissor,"Set_Scissor"},{n64_cmdSetPrimDepth,"Set_Prim_Depth"},{n64_cmdSetOtherModes,"Set_Other_Modes"},{n64_cmdLoadTLUT,"Load_TLUT"},{n64_cmdSetTileSize,"Set_Tile_Size"},{n64_cmdLoadBlock,"Load_Block"},{n64_cmdLoadTile,"Load_Tile"},{n64_cmdSetTile,"Set_Tile"},{n64_cmdFillRect,"Fill_Rectangle"},{n64_cmdSetFillColor,"Set_Fill_Color"},{n64_cmdSetFogColor,"Set_Fog_Color"},{n64_cmdSetBlendColor,"Set_Blend_Color"},{n64_cmdSetPrimColor,"Set_Prim_Color"},{n64_cmdSetEnvColor,"Set_Env_Color"},{n64_cmdSetCombineMode,"Set_Combine_Mode"},{n64_cmdSetTextureImage,"Set_Texture_Image"},{n64_cmdSetMaskImage,"Set_Mask_Image"},{n64_cmdSetColorImage,"Set_Color_Image"}};
constexpr int64_t n64_cycle1=0ULL;
constexpr int64_t n64_cycle2=1ULL;
constexpr int64_t n64_cycleCopy=2ULL;
constexpr int64_t n64_cycleFill=3ULL;
constexpr int64_t n64_fmtRGBA=0ULL;
constexpr int64_t n64_fmtYUV=1ULL;
constexpr int64_t n64_fmtCI=2ULL;
constexpr int64_t n64_fmtIA=3ULL;
constexpr int64_t n64_fmtI=4ULL;
constexpr int64_t n64_size4=0ULL;
constexpr int64_t n64_size8=1ULL;
constexpr int64_t n64_size16=2ULL;
constexpr int64_t n64_size32=3ULL;
constexpr int64_t n64_omAlphaCompare=1ULL;
constexpr int64_t n64_omZSourceSel=4ULL;
constexpr int64_t n64_omAntialias=8ULL;
constexpr int64_t n64_omZCompare=16ULL;
constexpr int64_t n64_omZUpdate=32ULL;
constexpr int64_t n64_omImageRead=64ULL;
constexpr int64_t n64_omCvgTimesAlpha=4096ULL;
constexpr int64_t n64_omAlphaCvgSel=8192ULL;
constexpr int64_t n64_omForceBlend=16384ULL;
constexpr int64_t n64_tlutBase=2048ULL;
constexpr int64_t n64_omSampleType=35184372088832ULL;
std::array<Anon18,8> n64_zRanges=std::array<Anon18,8>{Anon18{cast<uint32_t>(0ULL),cast<uint32_t>(6ULL)},Anon18{cast<uint32_t>(131072ULL),cast<uint32_t>(5ULL)},Anon18{cast<uint32_t>(196608ULL),cast<uint32_t>(4ULL)},Anon18{cast<uint32_t>(229376ULL),cast<uint32_t>(3ULL)},Anon18{cast<uint32_t>(245760ULL),cast<uint32_t>(2ULL)},Anon18{cast<uint32_t>(253952ULL),cast<uint32_t>(1ULL)},Anon18{cast<uint32_t>(258048ULL),cast<uint32_t>(0ULL)},Anon18{cast<uint32_t>(260096ULL),cast<uint32_t>(0ULL)}};
constexpr int64_t n64_omPerspTex=2251799813685248ULL;
constexpr int64_t n64_siDramAddr=0ULL;
constexpr int64_t n64_siReadAddr=4ULL;
constexpr int64_t n64_siWriteAddr=16ULL;
constexpr int64_t n64_siStatus=24ULL;
constexpr int64_t n64_siStatusDMABusy=1ULL;
constexpr int64_t n64_siStatusIOBusy=2ULL;
constexpr int64_t n64_siStatusInterrupt=4096ULL;
constexpr int64_t n64_pifRAMSize=64ULL;
constexpr int64_t n64_jbInfo=0ULL;
constexpr int64_t n64_jbControllerState=1ULL;
constexpr int64_t n64_jbReadAccessory=2ULL;
constexpr int64_t n64_jbWriteAccessory=3ULL;
constexpr int64_t n64_jbEepromRead=4ULL;
constexpr int64_t n64_jbEepromWrite=5ULL;
constexpr int64_t n64_jbReset=255ULL;
constexpr uint16_t n64_devController=1280ULL;
constexpr int64_t n64_BtnA=32768ULL;
constexpr int64_t n64_BtnB=16384ULL;
constexpr int64_t n64_BtnZ=8192ULL;
constexpr int64_t n64_BtnStart=4096ULL;
constexpr int64_t n64_BtnDUp=2048ULL;
constexpr int64_t n64_BtnDDown=1024ULL;
constexpr int64_t n64_BtnDLeft=512ULL;
constexpr int64_t n64_BtnDRight=256ULL;
constexpr int64_t n64_BtnL=32ULL;
constexpr int64_t n64_BtnR=16ULL;
constexpr int64_t n64_BtnCUp=8ULL;
constexpr int64_t n64_BtnCDown=4ULL;
constexpr int64_t n64_BtnCLeft=2ULL;
constexpr int64_t n64_BtnCRight=1ULL;
constexpr int64_t n64_spMemAddr=0ULL;
constexpr int64_t n64_spDramAddr=4ULL;
constexpr int64_t n64_spRdLen=8ULL;
constexpr int64_t n64_spWrLen=12ULL;
constexpr int64_t n64_spStatus=16ULL;
constexpr int64_t n64_spDMAFull=20ULL;
constexpr int64_t n64_spDMABusy=24ULL;
constexpr int64_t n64_spSemaphore=28ULL;
constexpr int64_t n64_spRegsBase=67371008ULL;
constexpr int64_t n64_spPCBase=67633152ULL;
constexpr int64_t n64_spPCEnd=67895296ULL;
constexpr int64_t n64_spStatusHalt=1ULL;
constexpr int64_t n64_spStatusBroke=2ULL;
constexpr int64_t n64_spStatusDMABusy=4ULL;
constexpr int64_t n64_spStatusDMAFull=8ULL;
constexpr int64_t n64_spStatusIOFull=16ULL;
constexpr int64_t n64_spStatusSingleStep=32ULL;
constexpr int64_t n64_spStatusIntrBreak=64ULL;
constexpr int64_t n64_spStatusSig0=128ULL;
constexpr int64_t n64_spWClearHalt=1ULL;
constexpr int64_t n64_spWSetHalt=2ULL;
constexpr int64_t n64_spWClearBroke=4ULL;
constexpr int64_t n64_spWClearIntr=8ULL;
constexpr int64_t n64_spWSetIntr=16ULL;
constexpr int64_t n64_spWClearSStep=32ULL;
constexpr int64_t n64_spWSetSStep=64ULL;
constexpr int64_t n64_spWClearIntrBreak=128ULL;
constexpr int64_t n64_spWSetIntrBreak=256ULL;
constexpr int64_t n64_spWSignalBase=9ULL;
constexpr int64_t n64_rspBudget=20000000ULL;
constexpr int64_t n64_viStatus=0ULL;
constexpr int64_t n64_viOrigin=4ULL;
constexpr int64_t n64_viWidth=8ULL;
constexpr int64_t n64_viIntr=12ULL;
constexpr int64_t n64_viCurrent=16ULL;
constexpr int64_t n64_viBurst=20ULL;
constexpr int64_t n64_viVSync=24ULL;
constexpr int64_t n64_viHSync=28ULL;
constexpr int64_t n64_viLeap=32ULL;
constexpr int64_t n64_viHStart=36ULL;
constexpr int64_t n64_viVStart=40ULL;
constexpr int64_t n64_viVBurst=44ULL;
constexpr int64_t n64_viXScale=48ULL;
constexpr int64_t n64_viYScale=52ULL;
constexpr int64_t n64_viTypeBlank=0ULL;
constexpr int64_t n64_viTypeReserved=1ULL;
constexpr int64_t n64_viTypeRGBA16=2ULL;
constexpr int64_t n64_viTypeRGBA32=3ULL;
constexpr int64_t n64_stepsPerField=750000ULL;
constexpr int64_t n64_halflinesPerField=525ULL;
uint64_t r4300_TLBEntry_pairSize(r4300_TLBEntry* e);
uint64_t r4300_TLBEntry_pageSize(r4300_TLBEntry* e);
std::tuple<uint32_t,bool> r4300_CPU_Translate(r4300_CPU* c,uint64_t vaddr,bool store);
std::tuple<uint32_t,bool> r4300_CPU_tlbTranslate(r4300_CPU* c,uint64_t vaddr,bool store);
void r4300_CPU_tlbException(r4300_CPU* c,uint64_t vaddr,bool store,bool refill);
void r4300_CPU_setEntryHiVPN(r4300_CPU* c,uint64_t vaddr);
void r4300_CPU_setFaultAddress(r4300_CPU* c,uint64_t vaddr);
void r4300_CPU_tlbr(r4300_CPU* c);
void r4300_CPU_tlbw(r4300_CPU* c,uint64_t i);
void r4300_CPU_tlbp(r4300_CPU* c);
uint64_t r4300_CPU_random(r4300_CPU* c);
uint64_t r4300_CPU_readCop0(r4300_CPU* c,uint32_t i);
void r4300_CPU_writeCop0(r4300_CPU* c,uint32_t i,uint64_t v);
void r4300_CPU_cop0(r4300_CPU* c,uint32_t w,uint32_t rs,uint32_t rt,uint32_t rd);
std::string r4300_CPU_DumpTLB(r4300_CPU* c);
void r4300_CPU_Reset(r4300_CPU* c);
void r4300_CPU_SetPC(r4300_CPU* c,uint64_t pc);
void r4300_CPU_SetReg(r4300_CPU* c,uint32_t i,uint64_t v);
uint64_t r4300_CPU_Reg(r4300_CPU* c,uint32_t i);
uint64_t r4300_CPU_CurPC(r4300_CPU* c);
void r4300_CPU_set(r4300_CPU* c,uint32_t i,uint64_t v);
uint64_t r4300_sext32(uint32_t v);
uint32_t r4300_CPU_read8(r4300_CPU* c,uint32_t a);
void r4300_CPU_write8(r4300_CPU* c,uint32_t a,uint32_t v);
uint32_t r4300_CPU_read16(r4300_CPU* c,uint32_t a);
void r4300_CPU_write16(r4300_CPU* c,uint32_t a,uint32_t v);
uint32_t r4300_CPU_read32(r4300_CPU* c,uint32_t a);
void r4300_CPU_write32(r4300_CPU* c,uint32_t a,uint32_t v);
uint64_t r4300_CPU_read64(r4300_CPU* c,uint32_t a);
void r4300_CPU_write64(r4300_CPU* c,uint32_t a,uint64_t v);
void r4300_CPU_Exception(r4300_CPU* c,uint32_t code);
void r4300_CPU_exceptionAt(r4300_CPU* c,uint32_t code,bool tlbRefill);
uint64_t r4300_sext64(uint64_t v);
void r4300_CPU_eret(r4300_CPU* c);
void r4300_CPU_coprocessorUnusable(r4300_CPU* c,uint32_t unit);
void r4300_CPU_addrError(r4300_CPU* c,uint32_t code,uint64_t vaddr);
bool r4300_CPU_Interrupt(r4300_CPU* c,bool pending);
bool r4300_CPU_checkInterrupt(r4300_CPU* c);
void r4300_CPU_tickCount(r4300_CPU* c);
int64_t r4300_CPU_Step(r4300_CPU* c);
std::tuple<uint32_t,bool> r4300_CPU_translateFetch(r4300_CPU* c,uint64_t vaddr);
void r4300_CPU_doBranch(r4300_CPU* c,bool taken,uint64_t target);
void r4300_CPU_doBranchLikely(r4300_CPU* c,bool taken,uint64_t target);
void r4300_CPU_execute(r4300_CPU* c,uint32_t w);
void r4300_CPU_cop2(r4300_CPU* c,uint32_t w,uint32_t rs,uint32_t rt);
void r4300_CPU_special(r4300_CPU* c,uint32_t w,uint32_t rs,uint32_t rt);
void r4300_CPU_regimm(r4300_CPU* c,uint32_t w,uint32_t rs,uint32_t rt,uint64_t branchT,uint64_t simm);
void r4300_CPU_trapIf(r4300_CPU* c,bool cond);
void r4300_CPU_divSigned(r4300_CPU* c,int32_t a,int32_t b);
void r4300_CPU_divUnsigned(r4300_CPU* c,uint32_t a,uint32_t b);
void r4300_CPU_ddivSigned(r4300_CPU* c,int64_t a,int64_t b);
void r4300_CPU_ddivUnsigned(r4300_CPU* c,uint64_t a,uint64_t b);
std::tuple<uint64_t,uint64_t> r4300_mul64Signed(int64_t a,int64_t b);
void r4300_CPU_loadOp(r4300_CPU* c,uint32_t op,uint32_t rs,uint32_t rt,uint64_t simm);
void r4300_CPU_storeOp(r4300_CPU* c,uint32_t op,uint32_t rs,uint32_t rt,uint64_t simm);
std::tuple<uint32_t,bool> r4300_addOv32(uint32_t a,uint32_t b);
std::tuple<uint32_t,bool> r4300_subOv32(uint32_t a,uint32_t b);
std::tuple<uint64_t,bool> r4300_addOv64(uint64_t a,uint64_t b);
std::tuple<uint64_t,bool> r4300_subOv64(uint64_t a,uint64_t b);
uint64_t r4300_b2u(bool b);
bool r4300_CPU_snanCheckS(r4300_CPU* c,uint32_t fd,uint32_t fs);
bool r4300_CPU_snanCheckD(r4300_CPU* c,uint32_t fd,uint32_t fs);
bool r4300_CPU_snanCheckSToD(r4300_CPU* c,uint32_t fd,uint32_t fs);
bool r4300_CPU_snanCheckDToS(r4300_CPU* c,uint32_t fd,uint32_t fs);
void r4300_CPU_setSResult(r4300_CPU* c,uint32_t i,float v);
void r4300_CPU_setDResult(r4300_CPU* c,uint32_t i,double v);
void r4300_CPU_clearCause(r4300_CPU* c);
void r4300_CPU_fpUnimplemented(r4300_CPU* c);
uint32_t r4300_CPU_readFGR32(r4300_CPU* c,uint32_t i);
void r4300_CPU_writeFGR32(r4300_CPU* c,uint32_t i,uint32_t v);
uint64_t r4300_CPU_readFGR64(r4300_CPU* c,uint32_t i);
void r4300_CPU_writeFGR64(r4300_CPU* c,uint32_t i,uint64_t v);
float r4300_CPU_fs(r4300_CPU* c,uint32_t i);
double r4300_CPU_fd(r4300_CPU* c,uint32_t i);
void r4300_CPU_setS(r4300_CPU* c,uint32_t i,float v);
void r4300_CPU_setD(r4300_CPU* c,uint32_t i,double v);
double r4300_CPU_round(r4300_CPU* c,double v);
void r4300_CPU_cop1Mem(r4300_CPU* c,uint32_t op,uint32_t rs,uint32_t ft,uint64_t simm);
void r4300_CPU_cop1(r4300_CPU* c,uint32_t w,uint32_t rs,uint32_t rt,uint32_t rd,uint32_t shamt,uint64_t branchT);
void r4300_CPU_cop1Single(r4300_CPU* c,uint32_t w,uint32_t funct,uint32_t ft,uint32_t fs,uint32_t fd);
void r4300_CPU_cop1Double(r4300_CPU* c,uint32_t w,uint32_t funct,uint32_t ft,uint32_t fs,uint32_t fd);
void r4300_CPU_compare(r4300_CPU* c,double a,double b,uint32_t cond,bool snanIn);
void r4300_CPU_toInt32(r4300_CPU* c,uint32_t fd,double v,std::function<double(double)> round);
void r4300_CPU_toInt64(r4300_CPU* c,uint32_t fd,double v,std::function<double(double)> round);
bool r4300_isNaN(double v);
bool r4300_isInf(double v);
double r4300_nan();
double r4300_sign(double v);
double r4300_inf(double s);
double r4300_zero(double s);
bool r4300_isSNaN32(uint32_t bits);
bool r4300_isSNaN64(uint64_t bits);
bool r4300_CPU_fpApply(r4300_CPU* c,uint32_t cond);
bool r4300_denormal(double v,bool single);
double r4300_applyRounding(uint32_t rm,double rounded,int64_t dsign,bool single);
double r4300_next(double v,double toward,bool single);
int64_t r4300_signOf(double v);
std::tuple<double,uint32_t> r4300_classifySingle(double rounded,bool inexact);
bool r4300_isSubnormal32(float f);
bool r4300_isSubnormal64(double v);
std::tuple<double,uint32_t> r4300_classifyDouble(double r,bool inexact);
std::tuple<double,uint32_t> r4300_fpArith(uint32_t rm,uint8_t op,double a,double b,bool single,bool snanIn);
double r4300_twoSumResidual(double a,double b,double s);
double r4300_infiniteResult(uint8_t op,double a,double b);
uint32_t r4300_sqrtConditions(double r,double x);
rsp_CPU* rsp_NewCPU(Slice<uint8_t> dmem,Slice<uint8_t> imem,n64_Machine* regs);
void rsp_CPU_Reset(rsp_CPU* c);
void rsp_CPU_unimpl(rsp_CPU* c,uint32_t w);
void rsp_CPU_SetPC(rsp_CPU* c,uint32_t pc);
uint32_t rsp_CPU_CurPC(rsp_CPU* c);
void rsp_CPU_Start(rsp_CPU* c,uint32_t pc);
uint32_t rsp_CPU_rd8(rsp_CPU* c,uint32_t a);
void rsp_CPU_wr8(rsp_CPU* c,uint32_t a,uint32_t v);
uint32_t rsp_CPU_rd16(rsp_CPU* c,uint32_t a);
void rsp_CPU_wr16(rsp_CPU* c,uint32_t a,uint32_t v);
uint32_t rsp_CPU_rd32(rsp_CPU* c,uint32_t a);
void rsp_CPU_wr32(rsp_CPU* c,uint32_t a,uint32_t v);
void rsp_CPU_set(rsp_CPU* c,uint32_t i,uint32_t v);
void rsp_CPU_Step(rsp_CPU* c);
uint64_t rsp_CPU_Run(rsp_CPU* c,uint64_t maxSteps);
void rsp_CPU_doBranch(rsp_CPU* c,bool taken,uint32_t target);
void rsp_CPU_execute(rsp_CPU* c,uint32_t w);
void rsp_CPU_special(rsp_CPU* c,uint32_t w,uint32_t rs,uint32_t rt);
uint32_t rsp_b2u(bool b);
void rsp_CPU_observeVector(rsp_CPU* c,uint32_t unused0,bool unused1);
uint32_t rsp_element(uint32_t e,uint32_t i);
uint16_t rsp_vte(std::array<uint16_t,8> v,uint32_t e,uint32_t i);
int64_t rsp_CPU_acc(rsp_CPU* c,uint32_t i);
void rsp_CPU_setAcc(rsp_CPU* c,uint32_t i,int64_t v);
void rsp_CPU_setAccLo(rsp_CPU* c,uint32_t i,uint16_t v);
uint16_t rsp_clampS(int64_t v);
uint16_t rsp_clampU(int64_t v);
uint16_t rsp_clampLow(int64_t acc);
int64_t rsp_s16(uint16_t v);
int64_t rsp_u16(uint16_t v);
bool rsp_bit(uint16_t f,uint32_t i);
void rsp_setBit(uint16_t* f,uint32_t i,bool on);
void rsp_CPU_cop2(rsp_CPU* c,uint32_t w);
uint16_t rsp_CPU_ctrl(rsp_CPU* c,uint32_t i);
void rsp_CPU_setCtrl(rsp_CPU* c,uint32_t i,uint16_t v);
uint8_t rsp_CPU_vecByte(rsp_CPU* c,uint32_t r,uint32_t b);
void rsp_CPU_setVecByte(rsp_CPU* c,uint32_t r,uint32_t b,uint8_t v);
void rsp_CPU_vectorALU(rsp_CPU* c,uint32_t w,uint32_t funct,uint32_t e,uint32_t vt,uint32_t vs,uint32_t vd);
int64_t rsp_b2i(bool b);
void rsp_CPU_compare(rsp_CPU* c,uint32_t funct,uint32_t e,std::array<uint16_t,8> vsv,std::array<uint16_t,8> vtv,uint32_t vd);
void rsp_CPU_vch(rsp_CPU* c,uint32_t e,std::array<uint16_t,8> vsv,std::array<uint16_t,8> vtv,uint32_t vd);
void rsp_CPU_vcl(rsp_CPU* c,uint32_t e,std::array<uint16_t,8> vsv,std::array<uint16_t,8> vtv,uint32_t vd);
void rsp_CPU_vcr(rsp_CPU* c,uint32_t e,std::array<uint16_t,8> vsv,std::array<uint16_t,8> vtv,uint32_t vd);
void rsp_CPU_vecLoad(rsp_CPU* c,uint32_t w,uint32_t rs,uint32_t vt);
void rsp_CPU_packedLoad(rsp_CPU* c,uint32_t vt,uint32_t e,uint32_t addr,uint32_t shift,uint32_t stride);
void rsp_CPU_lfv(rsp_CPU* c,uint32_t vt,uint32_t e,uint32_t addr);
void rsp_CPU_ltv(rsp_CPU* c,uint32_t vt,uint32_t e,uint32_t addr);
void rsp_CPU_vecStore(rsp_CPU* c,uint32_t w,uint32_t rs,uint32_t vt);
void rsp_CPU_packedStore(rsp_CPU* c,uint32_t vt,uint32_t e,uint32_t addr,uint32_t shift,uint32_t altShift);
void rsp_CPU_shv(rsp_CPU* c,uint32_t vt,uint32_t e,uint32_t addr);
void rsp_CPU_sfv(rsp_CPU* c,uint32_t vt,uint32_t e,uint32_t addr);
void rsp_CPU_stv(rsp_CPU* c,uint32_t vt,uint32_t e,uint32_t addr);
void rsp_CPU_divide(rsp_CPU* c,uint32_t e,std::array<uint16_t,8> vtv,uint32_t vs,uint32_t vd,uint32_t result);
int32_t rsp_CPU_divInput(rsp_CPU* c,uint16_t lo);
uint32_t rsp_CPU_reciprocal(rsp_CPU* c,int32_t in);
uint32_t rsp_CPU_rsqrt(rsp_CPU* c,int32_t in);
void n64_ai_init(n64_ai* a);
uint32_t n64_Machine_aiRead(n64_Machine* m,uint32_t addr);
void n64_Machine_aiWrite(n64_Machine* m,uint32_t addr,uint32_t v);
n64_BootConfig n64_DefaultBoot();
uint32_t n64_Machine_OSMemSize(n64_Machine* m);
uint32_t n64_Machine_dpRead(n64_Machine* m,uint32_t addr);
void n64_Machine_dpWrite(n64_Machine* m,uint32_t addr,uint32_t v);
void n64_Machine_joybusEEPROM(n64_Machine* m,Slice<uint8_t> cmd,Slice<uint8_t> res,int64_t rxAt);
bool n64_inISViewer(uint32_t addr);
uint32_t n64_Machine_isvRead(n64_Machine* m,uint32_t addr);
void n64_Machine_isvWrite(n64_Machine* m,uint32_t addr,uint32_t v);
void n64_Machine_isvWriteByte(n64_Machine* m,uint32_t addr,uint8_t v);
void n64_Machine_isvPrint(n64_Machine* m,std::string s);
void n64_Machine_isvEmit(n64_Machine* m,std::string line);
Slice<std::string> n64_Machine_ISViewerLines(n64_Machine* m);
r4300_CPU* n64_newBareCPU(n64_Machine* m);
uint8_t n64_Machine_rdramRead(n64_Machine* m,uint32_t a);
void n64_Machine_rdramWrite(n64_Machine* m,uint32_t a,uint8_t v);
uint32_t n64_Machine_pc(n64_Machine* m);
std::tuple<Slice<uint8_t>,uint32_t> n64_Machine_backing(n64_Machine* m,uint32_t addr);
uint8_t n64_Machine_Read(n64_Machine* m,uint32_t addr);
void n64_Machine_Write(n64_Machine* m,uint32_t addr,uint8_t v);
uint32_t n64_Machine_ioRead(n64_Machine* m,uint32_t addr);
void n64_Machine_ioWrite(n64_Machine* m,uint32_t addr,uint32_t v);
void n64_Machine_writePhys32(n64_Machine* m,uint32_t addr,uint32_t v);
std::tuple<uint32_t,bool> n64_Machine_ReadVirt(n64_Machine* m,uint64_t vaddr);
uint64_t n64_Machine_RSPSteps(n64_Machine* m);
uint64_t n64_Machine_RDPWords(n64_Machine* m);
void n64_Machine_raiseIRQ(n64_Machine* m,uint32_t bit);
void n64_Machine_clearIRQ(n64_Machine* m,uint32_t bit);
bool n64_Machine_irqPending(n64_Machine* m);
uint32_t n64_Machine_miRead(n64_Machine* m,uint32_t addr);
void n64_Machine_miWrite(n64_Machine* m,uint32_t addr,uint32_t v);
void n64_pi_init(n64_pi* p);
uint32_t n64_Machine_piRead(n64_Machine* m,uint32_t addr);
void n64_Machine_piWrite(n64_Machine* m,uint32_t addr,uint32_t v);
void n64_Machine_piDMA(n64_Machine* m,uint32_t length);
int64_t n64_cmdLen(uint32_t op);
uint32_t n64_rdp_cycleType(n64_rdp* r);
Slice<uint8_t> n64_Machine_RDPTMem(n64_Machine* m);
void n64_Machine_runRDP(n64_Machine* m);
std::string n64_cmdNameOf(uint32_t op);
std::string n64_RDPName(uint32_t op);
void n64_Machine_execRDP(n64_Machine* m,uint32_t op,Slice<uint64_t> w);
n64_rdpImage n64_decodeImage(uint64_t w);
n64_tile n64_decodeTile(uint64_t w);
n64_uint32Vec3 n64_combineInputs_subA(n64_combineInputs* in,uint32_t sel);
n64_uint32Vec3 n64_combineInputs_subB(n64_combineInputs* in,uint32_t sel);
n64_uint32Vec3 n64_combineInputs_mul(n64_combineInputs* in,uint32_t sel);
n64_uint32Vec3 n64_combineInputs_add(n64_combineInputs* in,uint32_t sel);
uint32_t n64_combineInputs_alphaABD(n64_combineInputs* in,uint32_t sel);
uint32_t n64_combineInputs_alphaMul(n64_combineInputs* in,uint32_t sel);
n64_uint32Vec3 n64_vec3(n64_rgba c);
n64_uint32Vec3 n64_splat(uint32_t v);
uint32_t n64_combineChannel(uint32_t a,uint32_t b,uint32_t c,uint32_t d);
n64_combinerSelects n64_rdp_combinerSelects(n64_rdp* r,int64_t cycle);
n64_rgba n64_rdp_combine(n64_rdp* r,n64_combineInputs* in);
std::tuple<uint32_t,uint32_t,uint32_t,uint32_t> n64_rdp_blenderSelects(n64_rdp* r,int64_t cycle);
n64_rgba n64_rdp_blend(n64_rdp* r,n64_rgba cyc,n64_rgba mem,uint32_t shadeAlpha);
uint32_t n64_rdp_pixelAddr(n64_rdp* r,uint32_t x,uint32_t y);
uint16_t n64_rgba5551(uint32_t rr,uint32_t gg,uint32_t bb,uint32_t aa);
void n64_Machine_writePixel(n64_Machine* m,uint32_t x,uint32_t y,uint32_t rr,uint32_t gg,uint32_t bb,uint32_t aa);
void n64_Machine_storeRDRAM8(n64_Machine* m,uint32_t a,uint8_t v);
void n64_Machine_storeRDRAM16(n64_Machine* m,uint32_t a,uint16_t v);
void n64_Machine_storeRDRAM32(n64_Machine* m,uint32_t a,uint32_t v);
std::tuple<uint32_t,uint32_t,uint32_t,uint32_t,bool> n64_rdp_clip(n64_rdp* r,uint32_t xh,uint32_t yh,uint32_t xl,uint32_t yl);
void n64_Machine_fillRect(n64_Machine* m,uint64_t w);
void n64_Machine_texRect(n64_Machine* m,Slice<uint64_t> w,bool flip);
void n64_Machine_loadBlock(n64_Machine* m,uint64_t w);
void n64_Machine_loadTile(n64_Machine* m,uint64_t w);
void n64_Machine_loadTLUT(n64_Machine* m,uint64_t w);
uint32_t n64_texelBytes(uint32_t size);
uint32_t n64_texCoord(int32_t v,uint32_t mask,uint32_t cm,uint32_t lo,uint32_t hi);
uint8_t n64_rdp_tmem(n64_rdp* r,uint32_t off);
uint16_t n64_rdp_tmem16(n64_rdp* r,uint32_t off);
uint16_t n64_rdp_tlut(n64_rdp* r,uint32_t i);
n64_rgba n64_fromRGBA16(uint16_t v);
std::tuple<n64_rgba,bool> n64_Machine_sample(n64_Machine* m,n64_tile* t,int32_t sFix,int32_t tFix);
n64_rgba n64_tap3(n64_rgba base,n64_rgba alongS,n64_rgba alongT,uint32_t sf,uint32_t tf);
uint32_t n64_swizzle(uint32_t off,uint32_t row);
std::tuple<n64_rgba,bool> n64_Machine_texelAt(n64_Machine* m,n64_tile* t,int32_t s,int32_t tt);
n64_rgba n64_Machine_readPixel(n64_Machine* m,uint32_t x,uint32_t y);
uint32_t n64_Machine_depthAt(n64_Machine* m,uint32_t x,uint32_t y);
void n64_Machine_setDepth(n64_Machine* m,uint32_t x,uint32_t y,uint32_t z);
uint32_t n64_depthOf(int64_t z);
void n64_Machine_drawPixel(n64_Machine* m,uint32_t x,uint32_t y,n64_combineInputs* in,int64_t z,bool useZ);
bool n64_rdp_blenderReadsMemory(n64_rdp* r);
int64_t n64_triAttrs_at(n64_triAttrs a,int64_t dy,int64_t dxPix);
int64_t n64_s32(uint32_t v);
int32_t n64_s14(uint32_t v);
int64_t n64_pair(uint64_t intWord,uint64_t fracWord,uint64_t shift);
void n64_Machine_triangle(n64_Machine* m,uint32_t op,Slice<uint64_t> w);
int64_t n64_ceilQuarter(int64_t v);
uint32_t n64_clamp8(int64_t v);
n64_rgba n64_unpackColor(uint32_t v);
std::string n64_Result_String(n64_Result r);
void n64_Machine_SetSpinDetect(n64_Machine* m,bool on);
void n64_Machine_SetBreakpoint(n64_Machine* m,uint32_t vaddr);
void n64_Machine_ClearBreakpoints(n64_Machine* m);
n64_Result n64_Machine_Run(n64_Machine* m,uint64_t maxSteps);
void n64_si_init(n64_si* s);
uint32_t n64_Machine_siRead(n64_Machine* m,uint32_t addr);
void n64_Machine_siWrite(n64_Machine* m,uint32_t addr,uint32_t v);
void n64_Machine_joybus(n64_Machine* m);
void n64_Machine_joybusChannel(n64_Machine* m,int64_t ch,Slice<uint8_t> cmd,Slice<uint8_t> res,int64_t rxAt);
uint32_t n64_Machine_spRead(n64_Machine* m,uint32_t addr);
void n64_Machine_spWrite(n64_Machine* m,uint32_t addr,uint32_t v);
void n64_Machine_spStatusWrite(n64_Machine* m,uint32_t v);
void n64_Machine_runRSP(n64_Machine* m);
uint32_t n64_Machine_ReadCop0(n64_Machine* m,uint32_t reg);
void n64_Machine_WriteCop0(n64_Machine* m,uint32_t reg,uint32_t v);
void n64_Machine_spDMA(n64_Machine* m,uint32_t lenReg,bool toRDRAM);
void n64_vi_init(n64_vi* v);
uint32_t n64_Machine_viRead(n64_Machine* m,uint32_t addr);
void n64_Machine_viWrite(n64_Machine* m,uint32_t addr,uint32_t v);
void n64_Machine_tickVI(n64_Machine* m);
uint32_t n64_Machine_Origin(n64_Machine* m);
uint32_t n64_Machine_Width(n64_Machine* m);
uint32_t n64_Machine_PixelType(n64_Machine* m);

#include "adapters.h"
// tools/cpu/r4300/cop0.go:36:1
uint64_t r4300_TLBEntry_pairSize(r4300_TLBEntry* e){
{
return cast<uint64_t>((cast<uint64_t>(cast<uint32_t>((e->PageMask | cast<uint32_t>(8191ULL)))) + cast<uint64_t>(1ULL)));
}
}
// tools/cpu/r4300/cop0.go:41:1
uint64_t r4300_TLBEntry_pageSize(r4300_TLBEntry* e){
{
return divi<uint64_t>(r4300_TLBEntry_pairSize(e),cast<uint64_t>(2ULL));
}
}
// tools/cpu/r4300/cop0.go:58:1
std::tuple<uint32_t,bool> r4300_CPU_Translate(r4300_CPU* c,uint64_t vaddr,bool store){
{
uint32_t v = cast<uint32_t>(vaddr);
{
if (((v >= r4300_kseg0) && (v < r4300_kseg1))){
return {cast<uint32_t>((v - r4300_kseg0)),true};
}
else if (((v >= r4300_kseg1) && (v < r4300_ksseg))){
return {cast<uint32_t>((v - r4300_kseg1)),true};
}
}
tmp1:;
return r4300_CPU_tlbTranslate(c,vaddr,store);
}
}
// tools/cpu/r4300/cop0.go:72:1
std::tuple<uint32_t,bool> r4300_CPU_tlbTranslate(r4300_CPU* c,uint64_t vaddr,bool store){
{
uint64_t asid = cast<uint64_t>((c->COP0[r4300_cop0EntryHi] & cast<uint64_t>(255ULL)));
{auto&& tmp2 = c->TLB;
for(int64_t tmp3=0;tmp3<len(tmp2);++tmp3){
auto i=tmp3;r4300_TLBEntry* e = (&c->TLB[i]);
uint64_t ps = r4300_TLBEntry_pageSize(e);
uint64_t pairMask = cast<uint64_t>(~(cast<uint64_t>((r4300_TLBEntry_pairSize(e) - cast<uint64_t>(1ULL)))));
if (((cast<uint64_t>((cast<uint64_t>((vaddr & pairMask)) & cast<uint64_t>(4294967295ULL)))) != (cast<uint64_t>((cast<uint64_t>((e->EntryHi & pairMask)) & cast<uint64_t>(4294967295ULL)))))) {
continue;
}
bool global = (cast<uint64_t>((cast<uint64_t>((e->EntryLo0 & e->EntryLo1)) & r4300_entryLoG)) != cast<uint64_t>(0ULL));
if (((!global) && (cast<uint64_t>((e->EntryHi & cast<uint64_t>(255ULL))) != asid))) {
continue;
}
uint64_t lo = e->EntryLo0;
if ((cast<uint64_t>((vaddr & ps)) != cast<uint64_t>(0ULL))) {
lo = e->EntryLo1;
}
if ((cast<uint64_t>((lo & r4300_entryLoV)) == cast<uint64_t>(0ULL))) {
r4300_CPU_tlbException(c,vaddr,store,false);
return {cast<uint32_t>(0ULL),false};
}
if ((store && (cast<uint64_t>((lo & r4300_entryLoD)) == cast<uint64_t>(0ULL)))) {
r4300_CPU_setFaultAddress(c,vaddr);
r4300_CPU_setEntryHiVPN(c,vaddr);
r4300_CPU_Exception(c,r4300_excMod);
return {cast<uint32_t>(0ULL),false};
}
uint64_t pfn = cast<uint64_t>(((shr<uint64_t>(lo,cast<int64_t>(6ULL))) & cast<uint64_t>(16777215ULL)));
return {cast<uint32_t>(cast<uint64_t>((((shl<uint64_t>(pfn,cast<int64_t>(12ULL))) & ~((cast<uint64_t>((ps - cast<uint64_t>(1ULL)))))) | (cast<uint64_t>((vaddr & (cast<uint64_t>((ps - cast<uint64_t>(1ULL)))))))))),true};
}}
r4300_CPU_tlbException(c,vaddr,store,true);
return {cast<uint32_t>(0ULL),false};
}
}
// tools/cpu/r4300/cop0.go:120:1
void r4300_CPU_tlbException(r4300_CPU* c,uint64_t vaddr,bool store,bool refill){
{
uint32_t code = cast<uint32_t>(r4300_excTLBL);
if (store) {
code = r4300_excTLBS;
}
r4300_CPU_setFaultAddress(c,vaddr);
r4300_CPU_setEntryHiVPN(c,vaddr);
r4300_CPU_exceptionAt(c,code,refill);
}
}
// tools/cpu/r4300/cop0.go:132:1
void r4300_CPU_setEntryHiVPN(r4300_CPU* c,uint64_t vaddr){
{
c->COP0[r4300_cop0EntryHi] = cast<uint64_t>(((cast<uint64_t>((c->COP0[r4300_cop0EntryHi] & cast<uint64_t>(255ULL)))) | ((vaddr & ~(cast<uint64_t>(8191ULL))))));
}
}
// tools/cpu/r4300/cop0.go:148:1
void r4300_CPU_setFaultAddress(r4300_CPU* c,uint64_t vaddr){
{
c->COP0[r4300_cop0BadVAddr] = vaddr;
uint64_t badVPN2 = cast<uint64_t>(((shr<uint64_t>(vaddr,cast<int64_t>(13ULL))) & cast<uint64_t>(524287ULL)));
c->COP0[r4300_cop0Context] = cast<uint64_t>((((c->COP0[r4300_cop0Context] & ~(cast<uint64_t>(8388607ULL)))) | (shl<uint64_t>(badVPN2,cast<int64_t>(4ULL)))));
uint64_t badVPN2x = cast<uint64_t>(((shr<uint64_t>(vaddr,cast<int64_t>(13ULL))) & cast<uint64_t>(134217727ULL)));
uint64_t region = cast<uint64_t>(((shr<uint64_t>(vaddr,cast<int64_t>(62ULL))) & cast<uint64_t>(3ULL)));
c->COP0[r4300_cop0XContext] = cast<uint64_t>((cast<uint64_t>((((c->COP0[r4300_cop0XContext] & ~(cast<uint64_t>(8589934591ULL)))) | (shl<uint64_t>(badVPN2x,cast<int64_t>(4ULL))))) | (shl<uint64_t>(region,cast<int64_t>(31ULL)))));
}
}
// tools/cpu/r4300/cop0.go:168:1
void r4300_CPU_tlbr(r4300_CPU* c){
{
uint64_t i = cast<uint64_t>((c->COP0[r4300_cop0Index] & cast<uint64_t>(63ULL)));
if ((i >= r4300_TLBSize)) {
r4300_CPU_Halt(c,"tlbr: Index %d out of range at 0x%08X",i,cast<uint32_t>(c->curPC));
return ;
}
r4300_TLBEntry* e = (&c->TLB[i]);
c->COP0[r4300_cop0PageMask] = cast<uint64_t>(e->PageMask);
c->COP0[r4300_cop0EntryHi] = e->EntryHi;
c->COP0[r4300_cop0EntryLo0] = e->EntryLo0;
c->COP0[r4300_cop0EntryLo1] = e->EntryLo1;
}
}
// tools/cpu/r4300/cop0.go:182:1
void r4300_CPU_tlbw(r4300_CPU* c,uint64_t i){
{
if ((i >= r4300_TLBSize)) {
r4300_CPU_Halt(c,"tlbw: index %d out of range at 0x%08X",i,cast<uint32_t>(c->curPC));
return ;
}
uint32_t pm = cast<uint32_t>((cast<uint32_t>(c->COP0[r4300_cop0PageMask]) & cast<uint32_t>(33546240ULL)));
c->TLB[i] = r4300_TLBEntry{pm,(c->COP0[r4300_cop0EntryHi] & ~(cast<uint64_t>(pm))),c->COP0[r4300_cop0EntryLo0],c->COP0[r4300_cop0EntryLo1]};
}
}
// tools/cpu/r4300/cop0.go:200:1
void r4300_CPU_tlbp(r4300_CPU* c){
{
uint64_t hi = c->COP0[r4300_cop0EntryHi];
uint64_t asid = cast<uint64_t>((hi & cast<uint64_t>(255ULL)));
{auto&& tmp4 = c->TLB;
for(int64_t tmp5=0;tmp5<len(tmp4);++tmp5){
auto i=tmp5;r4300_TLBEntry* e = (&c->TLB[i]);
uint64_t pairMask = cast<uint64_t>(~(cast<uint64_t>((r4300_TLBEntry_pairSize(e) - cast<uint64_t>(1ULL)))));
if (((cast<uint64_t>((cast<uint64_t>((hi & pairMask)) & cast<uint64_t>(4294967295ULL)))) != (cast<uint64_t>((cast<uint64_t>((e->EntryHi & pairMask)) & cast<uint64_t>(4294967295ULL)))))) {
continue;
}
bool global = (cast<uint64_t>((cast<uint64_t>((e->EntryLo0 & e->EntryLo1)) & r4300_entryLoG)) != cast<uint64_t>(0ULL));
if (((!global) && (cast<uint64_t>((e->EntryHi & cast<uint64_t>(255ULL))) != asid))) {
continue;
}
c->COP0[r4300_cop0Index] = cast<uint64_t>(i);
return ;
}}
c->COP0[r4300_cop0Index] = shl<uint64_t>(cast<int64_t>(1ULL),cast<int64_t>(31ULL));
}
}
// tools/cpu/r4300/cop0.go:221:1
uint64_t r4300_CPU_random(r4300_CPU* c){
{
uint64_t wired = cast<uint64_t>((c->COP0[r4300_cop0Wired] & cast<uint64_t>(63ULL)));
if ((wired >= r4300_TLBSize)) {
return cast<uint64_t>((r4300_TLBSize - cast<int64_t>(1ULL)));
}
uint64_t r = cast<uint64_t>((c->COP0[r4300_cop0Random] & cast<uint64_t>(63ULL)));
if (((r <= wired) || (r >= r4300_TLBSize))) {
r = cast<uint64_t>((r4300_TLBSize - cast<int64_t>(1ULL)));
}
else {
r--;
}
c->COP0[r4300_cop0Random] = r;
return r;
}
}
// tools/cpu/r4300/cop0.go:240:1
uint64_t r4300_CPU_readCop0(r4300_CPU* c,uint32_t i){
{
{
switch(i){
case r4300_cop0Random:{
return c->COP0[r4300_cop0Random];
break;}
case r4300_cop0Count:case r4300_cop0Compare:case r4300_cop0Status:case r4300_cop0Cause:case r4300_cop0PRId:case r4300_cop0Config:case r4300_cop0Index:case r4300_cop0Wired:case r4300_cop0PageMask:case r4300_cop0LLAddr:{
return r4300_sext32(cast<uint32_t>(c->COP0[i]));
break;}
}}
return c->COP0[i];
}
}
// tools/cpu/r4300/cop0.go:253:1
void r4300_CPU_writeCop0(r4300_CPU* c,uint32_t i,uint64_t v){
{
{
switch(i){
case r4300_cop0Compare:{
c->COP0[r4300_cop0Compare] = cast<uint64_t>(cast<uint32_t>(v));
c->COP0[r4300_cop0Cause] &= ~(r4300_causeIP7);
return ;
break;}
case r4300_cop0Cause:{
c->COP0[r4300_cop0Cause] = cast<uint64_t>((((c->COP0[r4300_cop0Cause] & ~(cast<uint64_t>(768ULL)))) | (cast<uint64_t>((v & cast<uint64_t>(768ULL))))));
return ;
break;}
case r4300_cop0Count:{
c->COP0[r4300_cop0Count] = cast<uint64_t>(cast<uint32_t>(v));
c->countFrac = cast<uint64_t>(0ULL);
return ;
break;}
case r4300_cop0PRId:case r4300_cop0Random:{
return ;
break;}
}}
c->COP0[i] = v;
}
}
// tools/cpu/r4300/cop0.go:274:1
void r4300_CPU_cop0(r4300_CPU* c,uint32_t w,uint32_t rs,uint32_t rt,uint32_t rd){
{
if ((cast<uint32_t>((w & (shl<uint32_t>(cast<int64_t>(1ULL),cast<int64_t>(25ULL))))) != cast<uint32_t>(0ULL))) {
{
switch(cast<uint32_t>((w & cast<uint32_t>(63ULL)))){
case cast<uint32_t>(1ULL):{
r4300_CPU_tlbr(c);
break;}
case cast<uint32_t>(2ULL):{
r4300_CPU_tlbw(c,cast<uint64_t>((c->COP0[r4300_cop0Index] & cast<uint64_t>(63ULL))));
break;}
case cast<uint32_t>(6ULL):{
r4300_CPU_tlbw(c,r4300_CPU_random(c));
break;}
case cast<uint32_t>(8ULL):{
r4300_CPU_tlbp(c);
break;}
case cast<uint32_t>(24ULL):{
r4300_CPU_eret(c);
break;}
default:{
break;}
}}
return ;
}
{
switch(rs){
case cast<uint32_t>(0ULL):{
r4300_CPU_set(c,rt,r4300_sext32(cast<uint32_t>(r4300_CPU_readCop0(c,rd))));
break;}
case cast<uint32_t>(1ULL):{
r4300_CPU_set(c,rt,r4300_CPU_readCop0(c,rd));
break;}
case cast<uint32_t>(4ULL):{
r4300_CPU_writeCop0(c,rd,r4300_sext32(cast<uint32_t>(c->R[rt])));
break;}
case cast<uint32_t>(5ULL):{
r4300_CPU_writeCop0(c,rd,c->R[rt]);
break;}
default:{
r4300_CPU_Exception(c,r4300_excRI);
break;}
}}
}
}
// tools/cpu/r4300/cop0.go:312:1
std::string r4300_CPU_DumpTLB(r4300_CPU* c){
{
std::string s = "";
{auto&& tmp6 = c->TLB;
for(int64_t tmp7=0;tmp7<len(tmp6);++tmp7){
auto i=tmp7;r4300_TLBEntry* e = (&c->TLB[i]);
if (((cast<uint64_t>((e->EntryLo0 & r4300_entryLoV)) == cast<uint64_t>(0ULL)) && (cast<uint64_t>((e->EntryLo1 & r4300_entryLoV)) == cast<uint64_t>(0ULL)))) {
continue;
}
s += go_fmt_Sprintf("%2d: hi=%016X lo0=%016X lo1=%016X mask=%08X size=%d\n",i,e->EntryHi,e->EntryLo0,e->EntryLo1,e->PageMask,r4300_TLBEntry_pageSize(e));
}}
if ((s == "")) {
return "TLB: no valid entries\n";
}
return s;
}
}
// tools/cpu/r4300/cpu.go:190:1
void r4300_CPU_Reset(r4300_CPU* c){
{
c->R = std::array<uint64_t,32>{};
auto tmp8 = std::make_tuple(cast<uint64_t>(0ULL),cast<uint64_t>(0ULL));
c->HI = std::get<0>(tmp8);
c->LO = std::get<1>(tmp8);
auto tmp9 = std::make_tuple(cast<uint64_t>(18446744072631615488ULL),cast<uint64_t>(18446744072631615492ULL));
c->PC = std::get<0>(tmp9);
c->nextPC = std::get<1>(tmp9);
c->COP0 = std::array<uint64_t,32>{};
c->COP0[r4300_cop0Status] = cast<uint64_t>((r4300_statusERL | r4300_statusBEV));
c->COP0[r4300_cop0PRId] = cast<uint64_t>(2850ULL);
c->COP0[r4300_cop0Config] = cast<uint64_t>(1879499875ULL);
c->COP0[r4300_cop0Random] = cast<uint64_t>((r4300_TLBSize - cast<int64_t>(1ULL)));
c->TLB = std::array<r4300_TLBEntry,32>{};
c->FGR = std::array<uint64_t,32>{};
c->FCR31 = cast<uint32_t>(0ULL);
c->COP2Latch = cast<uint64_t>(0ULL);
c->LLBit = false;
auto tmp10 = std::make_tuple(false,false);
c->delaySlot = std::get<0>(tmp10);
c->pendingDelay = std::get<1>(tmp10);
auto tmp11 = std::make_tuple(false,"");
c->Halted = std::get<0>(tmp11);
c->HaltReason = std::get<1>(tmp11);
auto tmp12 = std::make_tuple(cast<uint64_t>(0ULL),cast<uint64_t>(0ULL));
c->Steps = std::get<0>(tmp12);
c->countFrac = std::get<1>(tmp12);
}
}
// tools/cpu/r4300/cpu.go:219:1
void r4300_CPU_SetPC(r4300_CPU* c,uint64_t pc){
{
auto tmp13 = std::make_tuple(pc,cast<uint64_t>((pc + cast<uint64_t>(4ULL))));
c->PC = std::get<0>(tmp13);
c->nextPC = std::get<1>(tmp13);
c->pendingDelay = false;
}
}
// tools/cpu/r4300/cpu.go:225:1
void r4300_CPU_SetReg(r4300_CPU* c,uint32_t i,uint64_t v){
{
if ((i != cast<uint32_t>(0ULL))) {
c->R[i] = v;
}
}
}
// tools/cpu/r4300/cpu.go:232:1
uint64_t r4300_CPU_Reg(r4300_CPU* c,uint32_t i){
{
return c->R[i];
}
}
// tools/cpu/r4300/cpu.go:236:1
uint64_t r4300_CPU_CurPC(r4300_CPU* c){
{
return c->curPC;
}
}
// tools/cpu/r4300/cpu.go:239:1
void r4300_CPU_set(r4300_CPU* c,uint32_t i,uint64_t v){
{
if ((i != cast<uint32_t>(0ULL))) {
c->R[i] = v;
}
}
}
// tools/cpu/r4300/cpu.go:248:1
uint64_t r4300_sext32(uint32_t v){
{
return cast<uint64_t>(cast<int64_t>(cast<int32_t>(v)));
}
}
// tools/cpu/r4300/cpu.go:252:1
uint32_t r4300_CPU_read8(r4300_CPU* c,uint32_t a){
{
return cast<uint32_t>(n64_Machine_Read(c->bus,a));
}
}
// tools/cpu/r4300/cpu.go:253:1
void r4300_CPU_write8(r4300_CPU* c,uint32_t a,uint32_t v){
{
n64_Machine_Write(c->bus,a,cast<uint8_t>(v));
}
}
// tools/cpu/r4300/cpu.go:255:1
uint32_t r4300_CPU_read16(r4300_CPU* c,uint32_t a){
{
return cast<uint32_t>((shl<uint32_t>(cast<uint32_t>(n64_Machine_Read(c->bus,a)),cast<int64_t>(8ULL)) | cast<uint32_t>(n64_Machine_Read(c->bus,cast<uint32_t>((a + cast<uint32_t>(1ULL)))))));
}
}
// tools/cpu/r4300/cpu.go:258:1
void r4300_CPU_write16(r4300_CPU* c,uint32_t a,uint32_t v){
{
n64_Machine_Write(c->bus,a,cast<uint8_t>(shr<uint32_t>(v,cast<int64_t>(8ULL))));
n64_Machine_Write(c->bus,cast<uint32_t>((a + cast<uint32_t>(1ULL))),cast<uint8_t>(v));
}
}
// tools/cpu/r4300/cpu.go:263:1
uint32_t r4300_CPU_read32(r4300_CPU* c,uint32_t a){
{
return n64_Machine_Read32(c->bus,a);
}
}
// tools/cpu/r4300/cpu.go:264:1
void r4300_CPU_write32(r4300_CPU* c,uint32_t a,uint32_t v){
{
n64_Machine_Write32(c->bus,a,v);
}
}
// tools/cpu/r4300/cpu.go:266:1
uint64_t r4300_CPU_read64(r4300_CPU* c,uint32_t a){
{
return cast<uint64_t>((shl<uint64_t>(cast<uint64_t>(n64_Machine_Read32(c->bus,a)),cast<int64_t>(32ULL)) | cast<uint64_t>(n64_Machine_Read32(c->bus,cast<uint32_t>((a + cast<uint32_t>(4ULL)))))));
}
}
// tools/cpu/r4300/cpu.go:269:1
void r4300_CPU_write64(r4300_CPU* c,uint32_t a,uint64_t v){
{
n64_Machine_Write32(c->bus,a,cast<uint32_t>(shr<uint64_t>(v,cast<int64_t>(32ULL))));
n64_Machine_Write32(c->bus,cast<uint32_t>((a + cast<uint32_t>(4ULL))),cast<uint32_t>(v));
}
}
// tools/cpu/r4300/cpu.go:284:1
void r4300_CPU_Exception(r4300_CPU* c,uint32_t code){
{
r4300_CPU_exceptionAt(c,code,false);
}
}
// tools/cpu/r4300/cpu.go:288:1
void r4300_CPU_exceptionAt(r4300_CPU* c,uint32_t code,bool tlbRefill){
{
uint64_t sr = c->COP0[r4300_cop0Status];
if ((cast<uint64_t>((sr & r4300_statusEXL)) == cast<uint64_t>(0ULL))) {
uint64_t epc = c->curPC;
uint64_t cause = (c->COP0[r4300_cop0Cause] & ~(cast<uint64_t>(18446744069414584320ULL)));
if (c->delaySlot) {
epc = c->branchAddr;
cause |= r4300_causeBD;
}
else {
cause &= ~(r4300_causeBD);
}
c->COP0[r4300_cop0EPC] = epc;
c->COP0[r4300_cop0Cause] = cast<uint64_t>((((cause & ~(cast<uint64_t>(124ULL)))) | cast<uint64_t>(shl<uint32_t>(code,cast<int64_t>(2ULL)))));
}
else {
c->COP0[r4300_cop0Cause] = cast<uint64_t>((((c->COP0[r4300_cop0Cause] & ~(cast<uint64_t>(124ULL)))) | cast<uint64_t>(shl<uint32_t>(code,cast<int64_t>(2ULL)))));
}
c->COP0[r4300_cop0Cause] &= ~(cast<uint64_t>(805306368ULL));
uint64_t base = cast<uint64_t>(r4300_vecRAM);
if ((cast<uint64_t>((sr & r4300_statusBEV)) != cast<uint64_t>(0ULL))) {
base = r4300_vecROM;
}
uint64_t offset = cast<uint64_t>(cast<uint64_t>(384ULL));
if ((tlbRefill && (cast<uint64_t>((sr & r4300_statusEXL)) == cast<uint64_t>(0ULL)))) {
offset = cast<uint64_t>(0ULL);
}
c->COP0[r4300_cop0Status] = cast<uint64_t>((sr | r4300_statusEXL));
c->LLBit = false;
r4300_CPU_SetPC(c,r4300_sext64(cast<uint64_t>((base + offset))));
c->delaySlot = false;
}
}
// tools/cpu/r4300/cpu.go:331:1
uint64_t r4300_sext64(uint64_t v){
{
return cast<uint64_t>(cast<int64_t>(cast<int32_t>(cast<uint32_t>(v))));
}
}
// tools/cpu/r4300/cpu.go:335:1
void r4300_CPU_eret(r4300_CPU* c){
{
uint64_t sr = c->COP0[r4300_cop0Status];
if ((cast<uint64_t>((sr & r4300_statusERL)) != cast<uint64_t>(0ULL))) {
r4300_CPU_SetPC(c,c->COP0[r4300_cop0ErrorEPC]);
c->COP0[r4300_cop0Status] = (sr & ~(r4300_statusERL));
}
else {
r4300_CPU_SetPC(c,c->COP0[r4300_cop0EPC]);
c->COP0[r4300_cop0Status] = (sr & ~(r4300_statusEXL));
}
c->LLBit = false;
c->pendingDelay = false;
}
}
// tools/cpu/r4300/cpu.go:352:1
void r4300_CPU_coprocessorUnusable(r4300_CPU* c,uint32_t unit){
{
r4300_CPU_Exception(c,r4300_excCpU);
c->COP0[r4300_cop0Cause] = cast<uint64_t>((((c->COP0[r4300_cop0Cause] & ~(cast<uint64_t>(805306368ULL)))) | shl<uint64_t>(cast<uint64_t>(unit),cast<int64_t>(28ULL))));
}
}
// tools/cpu/r4300/cpu.go:358:1
void r4300_CPU_addrError(r4300_CPU* c,uint32_t code,uint64_t vaddr){
{
r4300_CPU_setFaultAddress(c,vaddr);
r4300_CPU_Exception(c,code);
}
}
// tools/cpu/r4300/cpu.go:370:1
bool r4300_CPU_Interrupt(r4300_CPU* c,bool pending){
{
if (pending) {
c->COP0[r4300_cop0Cause] |= r4300_causeIP2;
}
else {
c->COP0[r4300_cop0Cause] &= ~(r4300_causeIP2);
}
return r4300_CPU_checkInterrupt(c);
}
}
// tools/cpu/r4300/cpu.go:380:1
bool r4300_CPU_checkInterrupt(r4300_CPU* c){
{
uint64_t sr = c->COP0[r4300_cop0Status];
if (((cast<uint64_t>((sr & r4300_statusIE)) == cast<uint64_t>(0ULL)) || (cast<uint64_t>((sr & (cast<uint64_t>((r4300_statusEXL | r4300_statusERL))))) != cast<uint64_t>(0ULL)))) {
return false;
}
if ((cast<uint64_t>((cast<uint64_t>((c->COP0[r4300_cop0Cause] & sr)) & cast<uint64_t>(65280ULL))) == cast<uint64_t>(0ULL))) {
return false;
}
if (c->pendingDelay) {
return false;
}
c->curPC = c->PC;
c->delaySlot = false;
r4300_CPU_Exception(c,r4300_excInt);
return true;
}
}
// tools/cpu/r4300/cpu.go:406:1
void r4300_CPU_tickCount(r4300_CPU* c){
{
c->countFrac++;
if ((cast<uint64_t>((c->countFrac & cast<uint64_t>(1ULL))) != cast<uint64_t>(0ULL))) {
return ;
}
uint32_t count = cast<uint32_t>((cast<uint32_t>(c->COP0[r4300_cop0Count]) + cast<uint32_t>(1ULL)));
c->COP0[r4300_cop0Count] = cast<uint64_t>(count);
if ((count == cast<uint32_t>(c->COP0[r4300_cop0Compare]))) {
c->COP0[r4300_cop0Cause] |= r4300_causeIP7;
}
}
}
// tools/cpu/r4300/exec.go:19:1
int64_t r4300_CPU_Step(r4300_CPU* c){
{
if (c->Halted) {
return cast<int64_t>(0ULL);
}
r4300_CPU_tickCount(c);
if (r4300_CPU_checkInterrupt(c)) {
return cast<int64_t>(1ULL);
}
c->curPC = c->PC;
c->delaySlot = c->pendingDelay;
c->pendingDelay = false;
auto tmp14 = r4300_CPU_translateFetch(c,c->PC);
uint32_t paddr = std::get<0>(tmp14);
bool ok = std::get<1>(tmp14);
if ((!ok)) {
return cast<int64_t>(1ULL);
}
uint32_t w = n64_Machine_Fetch32(c->bus,paddr);
c->PC = c->nextPC;
c->nextPC += cast<uint64_t>(4ULL);
r4300_CPU_execute(c,w);
c->Steps++;
return cast<int64_t>(1ULL);
}
}
// tools/cpu/r4300/exec.go:48:1
std::tuple<uint32_t,bool> r4300_CPU_translateFetch(r4300_CPU* c,uint64_t vaddr){
{
if ((cast<uint64_t>((vaddr & cast<uint64_t>(3ULL))) != cast<uint64_t>(0ULL))) {
r4300_CPU_addrError(c,r4300_excAdEL,vaddr);
return {cast<uint32_t>(0ULL),false};
}
return r4300_CPU_Translate(c,vaddr,false);
}
}
// tools/cpu/r4300/exec.go:58:1
void r4300_CPU_doBranch(r4300_CPU* c,bool taken,uint64_t target){
{
c->pendingDelay = true;
c->branchAddr = c->curPC;
if (taken) {
c->nextPC = target;
}
}
}
// tools/cpu/r4300/exec.go:69:1
void r4300_CPU_doBranchLikely(r4300_CPU* c,bool taken,uint64_t target){
{
if (taken) {
r4300_CPU_doBranch(c,true,target);
return ;
}
c->PC = c->nextPC;
c->nextPC += cast<uint64_t>(4ULL);
}
}
// tools/cpu/r4300/exec.go:78:1
void r4300_CPU_execute(r4300_CPU* c,uint32_t w){
{
uint32_t op = shr<uint32_t>(w,cast<int64_t>(26ULL));
uint32_t rs = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(21ULL))) & cast<uint32_t>(31ULL)));
uint32_t rt = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(31ULL)));
uint32_t imm = cast<uint32_t>((w & cast<uint32_t>(65535ULL)));
uint64_t simm = cast<uint64_t>(cast<int64_t>(cast<int16_t>(imm)));
uint64_t branchT = cast<uint64_t>((cast<uint64_t>((c->curPC + cast<uint64_t>(4ULL))) + shl<uint64_t>(simm,cast<int64_t>(2ULL))));
uint64_t jumpT = cast<uint64_t>((cast<uint64_t>(((cast<uint64_t>((c->curPC + cast<uint64_t>(4ULL)))) & cast<uint64_t>(18446744073441116160ULL))) | shl<uint64_t>(cast<uint64_t>(cast<uint32_t>((w & cast<uint32_t>(67108863ULL)))),cast<int64_t>(2ULL))));
{
switch(op){
case cast<uint32_t>(0ULL):{
r4300_CPU_special(c,w,rs,rt);
break;}
case cast<uint32_t>(1ULL):{
r4300_CPU_regimm(c,w,rs,rt,branchT,simm);
break;}
case cast<uint32_t>(2ULL):{
r4300_CPU_doBranch(c,true,jumpT);
break;}
case cast<uint32_t>(3ULL):{
r4300_CPU_set(c,cast<uint32_t>(31ULL),cast<uint64_t>((c->curPC + cast<uint64_t>(8ULL))));
r4300_CPU_doBranch(c,true,jumpT);
break;}
case cast<uint32_t>(4ULL):{
r4300_CPU_doBranch(c,(c->R[rs] == c->R[rt]),branchT);
break;}
case cast<uint32_t>(5ULL):{
r4300_CPU_doBranch(c,(c->R[rs] != c->R[rt]),branchT);
break;}
case cast<uint32_t>(6ULL):{
r4300_CPU_doBranch(c,(cast<int64_t>(c->R[rs]) <= cast<int64_t>(0ULL)),branchT);
break;}
case cast<uint32_t>(7ULL):{
r4300_CPU_doBranch(c,(cast<int64_t>(c->R[rs]) > cast<int64_t>(0ULL)),branchT);
break;}
case cast<uint32_t>(20ULL):{
r4300_CPU_doBranchLikely(c,(c->R[rs] == c->R[rt]),branchT);
break;}
case cast<uint32_t>(21ULL):{
r4300_CPU_doBranchLikely(c,(c->R[rs] != c->R[rt]),branchT);
break;}
case cast<uint32_t>(22ULL):{
r4300_CPU_doBranchLikely(c,(cast<int64_t>(c->R[rs]) <= cast<int64_t>(0ULL)),branchT);
break;}
case cast<uint32_t>(23ULL):{
r4300_CPU_doBranchLikely(c,(cast<int64_t>(c->R[rs]) > cast<int64_t>(0ULL)),branchT);
break;}
case cast<uint32_t>(8ULL):{
{
auto tmp15 = r4300_addOv32(cast<uint32_t>(c->R[rs]),cast<uint32_t>(simm));
uint32_t r = std::get<0>(tmp15);
bool ov = std::get<1>(tmp15);
if (ov) {
r4300_CPU_Exception(c,r4300_excOv);
}
else {
r4300_CPU_set(c,rt,r4300_sext32(r));
}
}
break;}
case cast<uint32_t>(9ULL):{
r4300_CPU_set(c,rt,r4300_sext32(cast<uint32_t>((cast<uint32_t>(c->R[rs]) + cast<uint32_t>(simm)))));
break;}
case cast<uint32_t>(10ULL):{
r4300_CPU_set(c,rt,r4300_b2u((cast<int64_t>(c->R[rs]) < cast<int64_t>(simm))));
break;}
case cast<uint32_t>(11ULL):{
r4300_CPU_set(c,rt,r4300_b2u((c->R[rs] < simm)));
break;}
case cast<uint32_t>(12ULL):{
r4300_CPU_set(c,rt,cast<uint64_t>((c->R[rs] & cast<uint64_t>(imm))));
break;}
case cast<uint32_t>(13ULL):{
r4300_CPU_set(c,rt,cast<uint64_t>((c->R[rs] | cast<uint64_t>(imm))));
break;}
case cast<uint32_t>(14ULL):{
r4300_CPU_set(c,rt,cast<uint64_t>((c->R[rs] ^ cast<uint64_t>(imm))));
break;}
case cast<uint32_t>(15ULL):{
r4300_CPU_set(c,rt,r4300_sext32(shl<uint32_t>(imm,cast<int64_t>(16ULL))));
break;}
case cast<uint32_t>(24ULL):{
{
auto tmp16 = r4300_addOv64(c->R[rs],simm);
uint64_t r = std::get<0>(tmp16);
bool ov = std::get<1>(tmp16);
if (ov) {
r4300_CPU_Exception(c,r4300_excOv);
}
else {
r4300_CPU_set(c,rt,r);
}
}
break;}
case cast<uint32_t>(25ULL):{
r4300_CPU_set(c,rt,cast<uint64_t>((c->R[rs] + simm)));
break;}
case cast<uint32_t>(16ULL):{
r4300_CPU_cop0(c,w,rs,rt,cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(11ULL))) & cast<uint32_t>(31ULL))));
break;}
case cast<uint32_t>(17ULL):{
if ((cast<uint64_t>((c->COP0[r4300_cop0Status] & r4300_statusCU1)) == cast<uint64_t>(0ULL))) {
r4300_CPU_coprocessorUnusable(c,cast<uint32_t>(1ULL));
return ;
}
r4300_CPU_cop1(c,w,rs,rt,cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(11ULL))) & cast<uint32_t>(31ULL))),cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(6ULL))) & cast<uint32_t>(31ULL))),branchT);
break;}
case cast<uint32_t>(18ULL):{
r4300_CPU_cop2(c,w,rs,rt);
break;}
case cast<uint32_t>(32ULL):case cast<uint32_t>(33ULL):case cast<uint32_t>(34ULL):case cast<uint32_t>(35ULL):case cast<uint32_t>(36ULL):case cast<uint32_t>(37ULL):case cast<uint32_t>(38ULL):case cast<uint32_t>(39ULL):case cast<uint32_t>(26ULL):case cast<uint32_t>(27ULL):case cast<uint32_t>(48ULL):case cast<uint32_t>(52ULL):case cast<uint32_t>(55ULL):{
r4300_CPU_loadOp(c,op,rs,rt,simm);
break;}
case cast<uint32_t>(40ULL):case cast<uint32_t>(41ULL):case cast<uint32_t>(42ULL):case cast<uint32_t>(43ULL):case cast<uint32_t>(44ULL):case cast<uint32_t>(45ULL):case cast<uint32_t>(46ULL):case cast<uint32_t>(56ULL):case cast<uint32_t>(60ULL):case cast<uint32_t>(63ULL):{
r4300_CPU_storeOp(c,op,rs,rt,simm);
break;}
case cast<uint32_t>(47ULL):{
break;}
case cast<uint32_t>(49ULL):case cast<uint32_t>(53ULL):case cast<uint32_t>(57ULL):case cast<uint32_t>(61ULL):{
if ((cast<uint64_t>((c->COP0[r4300_cop0Status] & r4300_statusCU1)) == cast<uint64_t>(0ULL))) {
r4300_CPU_coprocessorUnusable(c,cast<uint32_t>(1ULL));
return ;
}
r4300_CPU_cop1Mem(c,op,rs,rt,simm);
break;}
default:{
r4300_CPU_Exception(c,r4300_excRI);
break;}
}}
}
}
// tools/cpu/r4300/exec.go:193:1
void r4300_CPU_cop2(r4300_CPU* c,uint32_t w,uint32_t rs,uint32_t rt){
{
if ((cast<uint64_t>((c->COP0[r4300_cop0Status] & r4300_statusCU2)) == cast<uint64_t>(0ULL))) {
r4300_CPU_coprocessorUnusable(c,cast<uint32_t>(2ULL));
return ;
}
{
switch(rs){
case cast<uint32_t>(0ULL):{
r4300_CPU_set(c,rt,r4300_sext32(cast<uint32_t>(c->COP2Latch)));
break;}
case cast<uint32_t>(1ULL):{
r4300_CPU_set(c,rt,c->COP2Latch);
break;}
case cast<uint32_t>(2ULL):{
r4300_CPU_set(c,rt,r4300_sext32(cast<uint32_t>(c->COP2Latch)));
break;}
case cast<uint32_t>(4ULL):{
c->COP2Latch = cast<uint64_t>(((c->COP2Latch & ~(cast<uint64_t>(4294967295ULL))) | cast<uint64_t>(cast<uint32_t>(c->R[rt]))));
break;}
case cast<uint32_t>(5ULL):{
c->COP2Latch = c->R[rt];
break;}
case cast<uint32_t>(6ULL):{
c->COP2Latch = cast<uint64_t>(((c->COP2Latch & ~(cast<uint64_t>(4294967295ULL))) | cast<uint64_t>(cast<uint32_t>(c->R[rt]))));
break;}
default:{
break;}
}}
}
}
// tools/cpu/r4300/exec.go:216:1
void r4300_CPU_special(r4300_CPU* c,uint32_t w,uint32_t rs,uint32_t rt){
{
uint32_t rd = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(11ULL))) & cast<uint32_t>(31ULL)));
uint32_t shamt = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(6ULL))) & cast<uint32_t>(31ULL)));
uint32_t funct = cast<uint32_t>((w & cast<uint32_t>(63ULL)));
{
switch(funct){
case cast<uint32_t>(0ULL):{
r4300_CPU_set(c,rd,r4300_sext32(shl<uint32_t>(cast<uint32_t>(c->R[rt]),shamt)));
break;}
case cast<uint32_t>(2ULL):{
r4300_CPU_set(c,rd,r4300_sext32(shr<uint32_t>(cast<uint32_t>(c->R[rt]),shamt)));
break;}
case cast<uint32_t>(3ULL):{
r4300_CPU_set(c,rd,r4300_sext32(cast<uint32_t>(shr<int32_t>(cast<int32_t>(cast<uint32_t>(c->R[rt])),shamt))));
break;}
case cast<uint32_t>(4ULL):{
r4300_CPU_set(c,rd,r4300_sext32(shl<uint32_t>(cast<uint32_t>(c->R[rt]),(cast<uint64_t>((c->R[rs] & cast<uint64_t>(31ULL)))))));
break;}
case cast<uint32_t>(6ULL):{
r4300_CPU_set(c,rd,r4300_sext32(shr<uint32_t>(cast<uint32_t>(c->R[rt]),(cast<uint64_t>((c->R[rs] & cast<uint64_t>(31ULL)))))));
break;}
case cast<uint32_t>(7ULL):{
r4300_CPU_set(c,rd,r4300_sext32(cast<uint32_t>(shr<int32_t>(cast<int32_t>(cast<uint32_t>(c->R[rt])),(cast<uint64_t>((c->R[rs] & cast<uint64_t>(31ULL))))))));
break;}
case cast<uint32_t>(8ULL):{
r4300_CPU_doBranch(c,true,c->R[rs]);
break;}
case cast<uint32_t>(9ULL):{
r4300_CPU_set(c,rd,cast<uint64_t>((c->curPC + cast<uint64_t>(8ULL))));
r4300_CPU_doBranch(c,true,c->R[rs]);
break;}
case cast<uint32_t>(12ULL):{
r4300_CPU_Exception(c,r4300_excSys);
break;}
case cast<uint32_t>(13ULL):{
r4300_CPU_Exception(c,r4300_excBp);
break;}
case cast<uint32_t>(15ULL):{
break;}
case cast<uint32_t>(16ULL):{
r4300_CPU_set(c,rd,c->HI);
break;}
case cast<uint32_t>(17ULL):{
c->HI = c->R[rs];
break;}
case cast<uint32_t>(18ULL):{
r4300_CPU_set(c,rd,c->LO);
break;}
case cast<uint32_t>(19ULL):{
c->LO = c->R[rs];
break;}
case cast<uint32_t>(20ULL):{
r4300_CPU_set(c,rd,shl<uint64_t>(c->R[rt],(cast<uint64_t>((c->R[rs] & cast<uint64_t>(63ULL))))));
break;}
case cast<uint32_t>(22ULL):{
r4300_CPU_set(c,rd,shr<uint64_t>(c->R[rt],(cast<uint64_t>((c->R[rs] & cast<uint64_t>(63ULL))))));
break;}
case cast<uint32_t>(23ULL):{
r4300_CPU_set(c,rd,cast<uint64_t>(shr<int64_t>(cast<int64_t>(c->R[rt]),(cast<uint64_t>((c->R[rs] & cast<uint64_t>(63ULL)))))));
break;}
case cast<uint32_t>(24ULL):{
int64_t p = cast<int64_t>((cast<int64_t>(cast<int32_t>(cast<uint32_t>(c->R[rs]))) * cast<int64_t>(cast<int32_t>(cast<uint32_t>(c->R[rt])))));
auto tmp17 = std::make_tuple(r4300_sext32(cast<uint32_t>(p)),r4300_sext32(cast<uint32_t>(shr<uint64_t>(cast<uint64_t>(p),cast<int64_t>(32ULL)))));
c->LO = std::get<0>(tmp17);
c->HI = std::get<1>(tmp17);
break;}
case cast<uint32_t>(25ULL):{
uint64_t p = cast<uint64_t>((cast<uint64_t>(cast<uint32_t>(c->R[rs])) * cast<uint64_t>(cast<uint32_t>(c->R[rt]))));
auto tmp18 = std::make_tuple(r4300_sext32(cast<uint32_t>(p)),r4300_sext32(cast<uint32_t>(shr<uint64_t>(p,cast<int64_t>(32ULL)))));
c->LO = std::get<0>(tmp18);
c->HI = std::get<1>(tmp18);
break;}
case cast<uint32_t>(26ULL):{
r4300_CPU_divSigned(c,cast<int32_t>(cast<uint32_t>(c->R[rs])),cast<int32_t>(cast<uint32_t>(c->R[rt])));
break;}
case cast<uint32_t>(27ULL):{
r4300_CPU_divUnsigned(c,cast<uint32_t>(c->R[rs]),cast<uint32_t>(c->R[rt]));
break;}
case cast<uint32_t>(28ULL):{
auto tmp19 = r4300_mul64Signed(cast<int64_t>(c->R[rs]),cast<int64_t>(c->R[rt]));
uint64_t hi = std::get<0>(tmp19);
uint64_t lo = std::get<1>(tmp19);
auto tmp20 = std::make_tuple(lo,hi);
c->LO = std::get<0>(tmp20);
c->HI = std::get<1>(tmp20);
break;}
case cast<uint32_t>(29ULL):{
auto tmp21 = go_bits_Mul64(c->R[rs],c->R[rt]);
uint64_t hi = std::get<0>(tmp21);
uint64_t lo = std::get<1>(tmp21);
auto tmp22 = std::make_tuple(lo,hi);
c->LO = std::get<0>(tmp22);
c->HI = std::get<1>(tmp22);
break;}
case cast<uint32_t>(30ULL):{
r4300_CPU_ddivSigned(c,cast<int64_t>(c->R[rs]),cast<int64_t>(c->R[rt]));
break;}
case cast<uint32_t>(31ULL):{
r4300_CPU_ddivUnsigned(c,c->R[rs],c->R[rt]);
break;}
case cast<uint32_t>(32ULL):{
{
auto tmp23 = r4300_addOv32(cast<uint32_t>(c->R[rs]),cast<uint32_t>(c->R[rt]));
uint32_t r = std::get<0>(tmp23);
bool ov = std::get<1>(tmp23);
if (ov) {
r4300_CPU_Exception(c,r4300_excOv);
}
else {
r4300_CPU_set(c,rd,r4300_sext32(r));
}
}
break;}
case cast<uint32_t>(33ULL):{
r4300_CPU_set(c,rd,r4300_sext32(cast<uint32_t>((cast<uint32_t>(c->R[rs]) + cast<uint32_t>(c->R[rt])))));
break;}
case cast<uint32_t>(34ULL):{
{
auto tmp24 = r4300_subOv32(cast<uint32_t>(c->R[rs]),cast<uint32_t>(c->R[rt]));
uint32_t r = std::get<0>(tmp24);
bool ov = std::get<1>(tmp24);
if (ov) {
r4300_CPU_Exception(c,r4300_excOv);
}
else {
r4300_CPU_set(c,rd,r4300_sext32(r));
}
}
break;}
case cast<uint32_t>(35ULL):{
r4300_CPU_set(c,rd,r4300_sext32(cast<uint32_t>((cast<uint32_t>(c->R[rs]) - cast<uint32_t>(c->R[rt])))));
break;}
case cast<uint32_t>(36ULL):{
r4300_CPU_set(c,rd,cast<uint64_t>((c->R[rs] & c->R[rt])));
break;}
case cast<uint32_t>(37ULL):{
r4300_CPU_set(c,rd,cast<uint64_t>((c->R[rs] | c->R[rt])));
break;}
case cast<uint32_t>(38ULL):{
r4300_CPU_set(c,rd,cast<uint64_t>((c->R[rs] ^ c->R[rt])));
break;}
case cast<uint32_t>(39ULL):{
r4300_CPU_set(c,rd,cast<uint64_t>(~(cast<uint64_t>((c->R[rs] | c->R[rt])))));
break;}
case cast<uint32_t>(42ULL):{
r4300_CPU_set(c,rd,r4300_b2u((cast<int64_t>(c->R[rs]) < cast<int64_t>(c->R[rt]))));
break;}
case cast<uint32_t>(43ULL):{
r4300_CPU_set(c,rd,r4300_b2u((c->R[rs] < c->R[rt])));
break;}
case cast<uint32_t>(44ULL):{
{
auto tmp25 = r4300_addOv64(c->R[rs],c->R[rt]);
uint64_t r = std::get<0>(tmp25);
bool ov = std::get<1>(tmp25);
if (ov) {
r4300_CPU_Exception(c,r4300_excOv);
}
else {
r4300_CPU_set(c,rd,r);
}
}
break;}
case cast<uint32_t>(45ULL):{
r4300_CPU_set(c,rd,cast<uint64_t>((c->R[rs] + c->R[rt])));
break;}
case cast<uint32_t>(46ULL):{
{
auto tmp26 = r4300_subOv64(c->R[rs],c->R[rt]);
uint64_t r = std::get<0>(tmp26);
bool ov = std::get<1>(tmp26);
if (ov) {
r4300_CPU_Exception(c,r4300_excOv);
}
else {
r4300_CPU_set(c,rd,r);
}
}
break;}
case cast<uint32_t>(47ULL):{
r4300_CPU_set(c,rd,cast<uint64_t>((c->R[rs] - c->R[rt])));
break;}
case cast<uint32_t>(48ULL):{
r4300_CPU_trapIf(c,(cast<int64_t>(c->R[rs]) >= cast<int64_t>(c->R[rt])));
break;}
case cast<uint32_t>(49ULL):{
r4300_CPU_trapIf(c,(c->R[rs] >= c->R[rt]));
break;}
case cast<uint32_t>(50ULL):{
r4300_CPU_trapIf(c,(cast<int64_t>(c->R[rs]) < cast<int64_t>(c->R[rt])));
break;}
case cast<uint32_t>(51ULL):{
r4300_CPU_trapIf(c,(c->R[rs] < c->R[rt]));
break;}
case cast<uint32_t>(52ULL):{
r4300_CPU_trapIf(c,(c->R[rs] == c->R[rt]));
break;}
case cast<uint32_t>(54ULL):{
r4300_CPU_trapIf(c,(c->R[rs] != c->R[rt]));
break;}
case cast<uint32_t>(56ULL):{
r4300_CPU_set(c,rd,shl<uint64_t>(c->R[rt],shamt));
break;}
case cast<uint32_t>(58ULL):{
r4300_CPU_set(c,rd,shr<uint64_t>(c->R[rt],shamt));
break;}
case cast<uint32_t>(59ULL):{
r4300_CPU_set(c,rd,cast<uint64_t>(shr<int64_t>(cast<int64_t>(c->R[rt]),shamt)));
break;}
case cast<uint32_t>(60ULL):{
r4300_CPU_set(c,rd,shl<uint64_t>(c->R[rt],(cast<uint32_t>((shamt + cast<uint32_t>(32ULL))))));
break;}
case cast<uint32_t>(62ULL):{
r4300_CPU_set(c,rd,shr<uint64_t>(c->R[rt],(cast<uint32_t>((shamt + cast<uint32_t>(32ULL))))));
break;}
case cast<uint32_t>(63ULL):{
r4300_CPU_set(c,rd,cast<uint64_t>(shr<int64_t>(cast<int64_t>(c->R[rt]),(cast<uint32_t>((shamt + cast<uint32_t>(32ULL)))))));
break;}
default:{
r4300_CPU_Exception(c,r4300_excRI);
break;}
}}
}
}
// tools/cpu/r4300/exec.go:368:1
void r4300_CPU_regimm(r4300_CPU* c,uint32_t w,uint32_t rs,uint32_t rt,uint64_t branchT,uint64_t simm){
{
int64_t s = cast<int64_t>(c->R[rs]);
{
switch(rt){
case cast<uint32_t>(0ULL):{
r4300_CPU_doBranch(c,(s < cast<int64_t>(0ULL)),branchT);
break;}
case cast<uint32_t>(1ULL):{
r4300_CPU_doBranch(c,(s >= cast<int64_t>(0ULL)),branchT);
break;}
case cast<uint32_t>(2ULL):{
r4300_CPU_doBranchLikely(c,(s < cast<int64_t>(0ULL)),branchT);
break;}
case cast<uint32_t>(3ULL):{
r4300_CPU_doBranchLikely(c,(s >= cast<int64_t>(0ULL)),branchT);
break;}
case cast<uint32_t>(8ULL):{
r4300_CPU_trapIf(c,(s >= cast<int64_t>(simm)));
break;}
case cast<uint32_t>(9ULL):{
r4300_CPU_trapIf(c,(c->R[rs] >= simm));
break;}
case cast<uint32_t>(10ULL):{
r4300_CPU_trapIf(c,(s < cast<int64_t>(simm)));
break;}
case cast<uint32_t>(11ULL):{
r4300_CPU_trapIf(c,(c->R[rs] < simm));
break;}
case cast<uint32_t>(12ULL):{
r4300_CPU_trapIf(c,(c->R[rs] == simm));
break;}
case cast<uint32_t>(14ULL):{
r4300_CPU_trapIf(c,(c->R[rs] != simm));
break;}
case cast<uint32_t>(16ULL):{
r4300_CPU_set(c,cast<uint32_t>(31ULL),cast<uint64_t>((c->curPC + cast<uint64_t>(8ULL))));
r4300_CPU_doBranch(c,(s < cast<int64_t>(0ULL)),branchT);
break;}
case cast<uint32_t>(17ULL):{
r4300_CPU_set(c,cast<uint32_t>(31ULL),cast<uint64_t>((c->curPC + cast<uint64_t>(8ULL))));
r4300_CPU_doBranch(c,(s >= cast<int64_t>(0ULL)),branchT);
break;}
case cast<uint32_t>(18ULL):{
r4300_CPU_set(c,cast<uint32_t>(31ULL),cast<uint64_t>((c->curPC + cast<uint64_t>(8ULL))));
r4300_CPU_doBranchLikely(c,(s < cast<int64_t>(0ULL)),branchT);
break;}
case cast<uint32_t>(19ULL):{
r4300_CPU_set(c,cast<uint32_t>(31ULL),cast<uint64_t>((c->curPC + cast<uint64_t>(8ULL))));
r4300_CPU_doBranchLikely(c,(s >= cast<int64_t>(0ULL)),branchT);
break;}
default:{
r4300_CPU_Exception(c,r4300_excRI);
break;}
}}
}
}
// tools/cpu/r4300/exec.go:412:1
void r4300_CPU_trapIf(r4300_CPU* c,bool cond){
{
if (cond) {
r4300_CPU_Exception(c,r4300_excTrap);
}
}
}
// tools/cpu/r4300/exec.go:422:1
void r4300_CPU_divSigned(r4300_CPU* c,int32_t a,int32_t b){
{
{
if ((b == cast<int32_t>(0ULL))){
c->HI = r4300_sext32(cast<uint32_t>(a));
if ((a >= cast<int32_t>(0ULL))) {
c->LO = r4300_sext32(cast<uint32_t>(4294967295ULL));
}
else {
c->LO = cast<uint64_t>(1ULL);
}
}
else if (((a == cast<int32_t>(-cast<int64_t>(2147483648ULL))) && (b == cast<int32_t>(-cast<int64_t>(1ULL))))){
auto tmp28 = std::make_tuple(r4300_sext32(cast<uint32_t>(2147483648ULL)),cast<uint64_t>(0ULL));
c->LO = std::get<0>(tmp28);
c->HI = std::get<1>(tmp28);
}
else {
auto tmp29 = std::make_tuple(r4300_sext32(cast<uint32_t>(divi<int32_t>(a,b))),r4300_sext32(cast<uint32_t>(modi<int32_t>(a,b))));
c->LO = std::get<0>(tmp29);
c->HI = std::get<1>(tmp29);
}
}
tmp27:;
}
}
// tools/cpu/r4300/exec.go:438:1
void r4300_CPU_divUnsigned(r4300_CPU* c,uint32_t a,uint32_t b){
{
if ((b == cast<uint32_t>(0ULL))) {
auto tmp30 = std::make_tuple(r4300_sext32(cast<uint32_t>(4294967295ULL)),r4300_sext32(a));
c->LO = std::get<0>(tmp30);
c->HI = std::get<1>(tmp30);
return ;
}
auto tmp31 = std::make_tuple(r4300_sext32(divi<uint32_t>(a,b)),r4300_sext32(modi<uint32_t>(a,b)));
c->LO = std::get<0>(tmp31);
c->HI = std::get<1>(tmp31);
}
}
// tools/cpu/r4300/exec.go:446:1
void r4300_CPU_ddivSigned(r4300_CPU* c,int64_t a,int64_t b){
{
{
if ((b == cast<int64_t>(0ULL))){
c->HI = cast<uint64_t>(a);
if ((a >= cast<int64_t>(0ULL))) {
c->LO = cast<uint64_t>(~cast<uint64_t>(cast<uint64_t>(0ULL)));
}
else {
c->LO = cast<uint64_t>(1ULL);
}
}
else if (((a == cast<int64_t>(-cast<int64_t>(9223372036854775808ULL))) && (b == cast<int64_t>(-cast<int64_t>(1ULL))))){
auto tmp33 = std::make_tuple(cast<uint64_t>(a),cast<uint64_t>(0ULL));
c->LO = std::get<0>(tmp33);
c->HI = std::get<1>(tmp33);
}
else {
auto tmp34 = std::make_tuple(cast<uint64_t>(divi<int64_t>(a,b)),cast<uint64_t>(modi<int64_t>(a,b)));
c->LO = std::get<0>(tmp34);
c->HI = std::get<1>(tmp34);
}
}
tmp32:;
}
}
// tools/cpu/r4300/exec.go:462:1
void r4300_CPU_ddivUnsigned(r4300_CPU* c,uint64_t a,uint64_t b){
{
if ((b == cast<uint64_t>(0ULL))) {
auto tmp35 = std::make_tuple(cast<uint64_t>(~cast<uint64_t>(cast<uint64_t>(0ULL))),a);
c->LO = std::get<0>(tmp35);
c->HI = std::get<1>(tmp35);
return ;
}
auto tmp36 = std::make_tuple(divi<uint64_t>(a,b),modi<uint64_t>(a,b));
c->LO = std::get<0>(tmp36);
c->HI = std::get<1>(tmp36);
}
}
// tools/cpu/r4300/exec.go:472:1
std::tuple<uint64_t,uint64_t> r4300_mul64Signed(int64_t a,int64_t b){
uint64_t hi{};
uint64_t lo{};
{
auto tmp37 = go_bits_Mul64(cast<uint64_t>(a),cast<uint64_t>(b));
hi = std::get<0>(tmp37);
lo = std::get<1>(tmp37);
if ((a < cast<int64_t>(0ULL))) {
hi -= cast<uint64_t>(b);
}
if ((b < cast<int64_t>(0ULL))) {
hi -= cast<uint64_t>(a);
}
return {hi,lo};
}
}
// tools/cpu/r4300/exec.go:490:1
void r4300_CPU_loadOp(r4300_CPU* c,uint32_t op,uint32_t rs,uint32_t rt,uint64_t simm){
{
uint64_t vaddr = cast<uint64_t>((c->R[rs] + simm));
std::function<std::tuple<uint32_t,bool>(uint64_t)> align = [&](uint64_t n)->std::tuple<uint32_t,bool>{
if ((cast<uint64_t>((vaddr & (cast<uint64_t>((n - cast<uint64_t>(1ULL)))))) != cast<uint64_t>(0ULL))) {
r4300_CPU_addrError(c,r4300_excAdEL,vaddr);
return {cast<uint32_t>(0ULL),false};
}
return r4300_CPU_Translate(c,vaddr,false);
}
;
{
switch(op){
case cast<uint32_t>(32ULL):{
auto tmp38 = r4300_CPU_Translate(c,vaddr,false);
uint32_t p = std::get<0>(tmp38);
bool ok = std::get<1>(tmp38);
if ((!ok)) {
return ;
}
r4300_CPU_set(c,rt,cast<uint64_t>(cast<int64_t>(cast<int8_t>(cast<uint8_t>(r4300_CPU_read8(c,p))))));
break;}
case cast<uint32_t>(36ULL):{
auto tmp39 = r4300_CPU_Translate(c,vaddr,false);
uint32_t p = std::get<0>(tmp39);
bool ok = std::get<1>(tmp39);
if ((!ok)) {
return ;
}
r4300_CPU_set(c,rt,cast<uint64_t>(r4300_CPU_read8(c,p)));
break;}
case cast<uint32_t>(33ULL):{
auto tmp40 = align(cast<uint64_t>(2ULL));
uint32_t p = std::get<0>(tmp40);
bool ok = std::get<1>(tmp40);
if ((!ok)) {
return ;
}
r4300_CPU_set(c,rt,cast<uint64_t>(cast<int64_t>(cast<int16_t>(cast<uint16_t>(r4300_CPU_read16(c,p))))));
break;}
case cast<uint32_t>(37ULL):{
auto tmp41 = align(cast<uint64_t>(2ULL));
uint32_t p = std::get<0>(tmp41);
bool ok = std::get<1>(tmp41);
if ((!ok)) {
return ;
}
r4300_CPU_set(c,rt,cast<uint64_t>(r4300_CPU_read16(c,p)));
break;}
case cast<uint32_t>(35ULL):{
auto tmp42 = align(cast<uint64_t>(4ULL));
uint32_t p = std::get<0>(tmp42);
bool ok = std::get<1>(tmp42);
if ((!ok)) {
return ;
}
r4300_CPU_set(c,rt,r4300_sext32(r4300_CPU_read32(c,p)));
break;}
case cast<uint32_t>(39ULL):{
auto tmp43 = align(cast<uint64_t>(4ULL));
uint32_t p = std::get<0>(tmp43);
bool ok = std::get<1>(tmp43);
if ((!ok)) {
return ;
}
r4300_CPU_set(c,rt,cast<uint64_t>(r4300_CPU_read32(c,p)));
break;}
case cast<uint32_t>(55ULL):{
auto tmp44 = align(cast<uint64_t>(8ULL));
uint32_t p = std::get<0>(tmp44);
bool ok = std::get<1>(tmp44);
if ((!ok)) {
return ;
}
r4300_CPU_set(c,rt,r4300_CPU_read64(c,p));
break;}
case cast<uint32_t>(48ULL):{
auto tmp45 = align(cast<uint64_t>(4ULL));
uint32_t p = std::get<0>(tmp45);
bool ok = std::get<1>(tmp45);
if ((!ok)) {
return ;
}
r4300_CPU_set(c,rt,r4300_sext32(r4300_CPU_read32(c,p)));
c->COP0[r4300_cop0LLAddr] = cast<uint64_t>(shr<uint32_t>(p,cast<int64_t>(4ULL)));
c->LLBit = true;
break;}
case cast<uint32_t>(52ULL):{
auto tmp46 = align(cast<uint64_t>(8ULL));
uint32_t p = std::get<0>(tmp46);
bool ok = std::get<1>(tmp46);
if ((!ok)) {
return ;
}
r4300_CPU_set(c,rt,r4300_CPU_read64(c,p));
c->COP0[r4300_cop0LLAddr] = cast<uint64_t>(shr<uint32_t>(p,cast<int64_t>(4ULL)));
c->LLBit = true;
break;}
case cast<uint32_t>(34ULL):{
auto tmp47 = r4300_CPU_Translate(c,(vaddr & ~(cast<uint64_t>(3ULL))),false);
uint32_t p = std::get<0>(tmp47);
bool ok = std::get<1>(tmp47);
if ((!ok)) {
return ;
}
uint64_t shift = cast<uint64_t>(((cast<uint64_t>((vaddr & cast<uint64_t>(3ULL)))) * cast<uint64_t>(8ULL)));
uint32_t word = r4300_CPU_read32(c,p);
uint32_t cur = cast<uint32_t>(c->R[rt]);
r4300_CPU_set(c,rt,r4300_sext32(cast<uint32_t>(((cast<uint32_t>((cur & (cast<uint32_t>((shl<uint32_t>(cast<uint32_t>(cast<uint32_t>(1ULL)),shift) - cast<uint32_t>(1ULL))))))) | (shl<uint32_t>(word,shift))))));
break;}
case cast<uint32_t>(38ULL):{
auto tmp48 = r4300_CPU_Translate(c,(vaddr & ~(cast<uint64_t>(3ULL))),false);
uint32_t p = std::get<0>(tmp48);
bool ok = std::get<1>(tmp48);
if ((!ok)) {
return ;
}
uint64_t shift = cast<uint64_t>(((cast<uint64_t>((cast<uint64_t>(3ULL) - cast<uint64_t>((vaddr & cast<uint64_t>(3ULL)))))) * cast<uint64_t>(8ULL)));
uint32_t word = r4300_CPU_read32(c,p);
uint32_t cur = cast<uint32_t>(c->R[rt]);
r4300_CPU_set(c,rt,r4300_sext32(cast<uint32_t>((((cur & ~((shr<uint32_t>(cast<uint32_t>(cast<uint32_t>(4294967295ULL)),shift))))) | (shr<uint32_t>(word,shift))))));
break;}
case cast<uint32_t>(26ULL):{
auto tmp49 = r4300_CPU_Translate(c,(vaddr & ~(cast<uint64_t>(7ULL))),false);
uint32_t p = std::get<0>(tmp49);
bool ok = std::get<1>(tmp49);
if ((!ok)) {
return ;
}
uint64_t shift = cast<uint64_t>(((cast<uint64_t>((vaddr & cast<uint64_t>(7ULL)))) * cast<uint64_t>(8ULL)));
uint64_t d = r4300_CPU_read64(c,p);
r4300_CPU_set(c,rt,cast<uint64_t>(((cast<uint64_t>((c->R[rt] & (cast<uint64_t>((shl<uint64_t>(cast<uint64_t>(cast<uint64_t>(1ULL)),shift) - cast<uint64_t>(1ULL))))))) | (shl<uint64_t>(d,shift)))));
break;}
case cast<uint32_t>(27ULL):{
auto tmp50 = r4300_CPU_Translate(c,(vaddr & ~(cast<uint64_t>(7ULL))),false);
uint32_t p = std::get<0>(tmp50);
bool ok = std::get<1>(tmp50);
if ((!ok)) {
return ;
}
uint64_t shift = cast<uint64_t>(((cast<uint64_t>((cast<uint64_t>(7ULL) - cast<uint64_t>((vaddr & cast<uint64_t>(7ULL)))))) * cast<uint64_t>(8ULL)));
uint64_t d = r4300_CPU_read64(c,p);
r4300_CPU_set(c,rt,cast<uint64_t>((((c->R[rt] & ~((shr<uint64_t>(cast<uint64_t>(~cast<uint64_t>(cast<uint64_t>(0ULL))),shift))))) | (shr<uint64_t>(d,shift)))));
break;}
}}
}
}
// tools/cpu/r4300/exec.go:602:1
void r4300_CPU_storeOp(r4300_CPU* c,uint32_t op,uint32_t rs,uint32_t rt,uint64_t simm){
{
uint64_t vaddr = cast<uint64_t>((c->R[rs] + simm));
uint64_t v = c->R[rt];
std::function<std::tuple<uint32_t,bool>(uint64_t)> align = [&](uint64_t n)->std::tuple<uint32_t,bool>{
if ((cast<uint64_t>((vaddr & (cast<uint64_t>((n - cast<uint64_t>(1ULL)))))) != cast<uint64_t>(0ULL))) {
r4300_CPU_addrError(c,r4300_excAdES,vaddr);
return {cast<uint32_t>(0ULL),false};
}
return r4300_CPU_Translate(c,vaddr,true);
}
;
{
switch(op){
case cast<uint32_t>(40ULL):{
auto tmp51 = r4300_CPU_Translate(c,vaddr,true);
uint32_t p = std::get<0>(tmp51);
bool ok = std::get<1>(tmp51);
if ((!ok)) {
return ;
}
r4300_CPU_write8(c,p,cast<uint32_t>((cast<uint32_t>(v) & cast<uint32_t>(255ULL))));
break;}
case cast<uint32_t>(41ULL):{
auto tmp52 = align(cast<uint64_t>(2ULL));
uint32_t p = std::get<0>(tmp52);
bool ok = std::get<1>(tmp52);
if ((!ok)) {
return ;
}
r4300_CPU_write16(c,p,cast<uint32_t>((cast<uint32_t>(v) & cast<uint32_t>(65535ULL))));
break;}
case cast<uint32_t>(43ULL):{
auto tmp53 = align(cast<uint64_t>(4ULL));
uint32_t p = std::get<0>(tmp53);
bool ok = std::get<1>(tmp53);
if ((!ok)) {
return ;
}
r4300_CPU_write32(c,p,cast<uint32_t>(v));
break;}
case cast<uint32_t>(63ULL):{
auto tmp54 = align(cast<uint64_t>(8ULL));
uint32_t p = std::get<0>(tmp54);
bool ok = std::get<1>(tmp54);
if ((!ok)) {
return ;
}
r4300_CPU_write64(c,p,v);
break;}
case cast<uint32_t>(56ULL):{
auto tmp55 = align(cast<uint64_t>(4ULL));
uint32_t p = std::get<0>(tmp55);
bool ok = std::get<1>(tmp55);
if ((!ok)) {
return ;
}
if (c->LLBit) {
r4300_CPU_write32(c,p,cast<uint32_t>(v));
}
r4300_CPU_set(c,rt,r4300_b2u(c->LLBit));
c->LLBit = false;
break;}
case cast<uint32_t>(60ULL):{
auto tmp56 = align(cast<uint64_t>(8ULL));
uint32_t p = std::get<0>(tmp56);
bool ok = std::get<1>(tmp56);
if ((!ok)) {
return ;
}
if (c->LLBit) {
r4300_CPU_write64(c,p,v);
}
r4300_CPU_set(c,rt,r4300_b2u(c->LLBit));
c->LLBit = false;
break;}
case cast<uint32_t>(42ULL):{
auto tmp57 = r4300_CPU_Translate(c,(vaddr & ~(cast<uint64_t>(3ULL))),true);
uint32_t p = std::get<0>(tmp57);
bool ok = std::get<1>(tmp57);
if ((!ok)) {
return ;
}
uint64_t shift = cast<uint64_t>(((cast<uint64_t>((vaddr & cast<uint64_t>(3ULL)))) * cast<uint64_t>(8ULL)));
uint32_t word = r4300_CPU_read32(c,p);
r4300_CPU_write32(c,p,cast<uint32_t>((((word & ~((shr<uint32_t>(cast<uint32_t>(cast<uint32_t>(4294967295ULL)),shift))))) | (shr<uint32_t>(cast<uint32_t>(v),shift)))));
break;}
case cast<uint32_t>(46ULL):{
auto tmp58 = r4300_CPU_Translate(c,(vaddr & ~(cast<uint64_t>(3ULL))),true);
uint32_t p = std::get<0>(tmp58);
bool ok = std::get<1>(tmp58);
if ((!ok)) {
return ;
}
uint64_t shift = cast<uint64_t>(((cast<uint64_t>((cast<uint64_t>(3ULL) - cast<uint64_t>((vaddr & cast<uint64_t>(3ULL)))))) * cast<uint64_t>(8ULL)));
uint32_t word = r4300_CPU_read32(c,p);
r4300_CPU_write32(c,p,cast<uint32_t>((((word & ~((shl<uint32_t>(cast<uint32_t>(cast<uint32_t>(4294967295ULL)),shift))))) | (shl<uint32_t>(cast<uint32_t>(v),shift)))));
break;}
case cast<uint32_t>(44ULL):{
auto tmp59 = r4300_CPU_Translate(c,(vaddr & ~(cast<uint64_t>(7ULL))),true);
uint32_t p = std::get<0>(tmp59);
bool ok = std::get<1>(tmp59);
if ((!ok)) {
return ;
}
uint64_t shift = cast<uint64_t>(((cast<uint64_t>((vaddr & cast<uint64_t>(7ULL)))) * cast<uint64_t>(8ULL)));
r4300_CPU_write64(c,p,cast<uint64_t>((((r4300_CPU_read64(c,p) & ~((shr<uint64_t>(cast<uint64_t>(~cast<uint64_t>(cast<uint64_t>(0ULL))),shift))))) | (shr<uint64_t>(v,shift)))));
break;}
case cast<uint32_t>(45ULL):{
auto tmp60 = r4300_CPU_Translate(c,(vaddr & ~(cast<uint64_t>(7ULL))),true);
uint32_t p = std::get<0>(tmp60);
bool ok = std::get<1>(tmp60);
if ((!ok)) {
return ;
}
uint64_t shift = cast<uint64_t>(((cast<uint64_t>((cast<uint64_t>(7ULL) - cast<uint64_t>((vaddr & cast<uint64_t>(7ULL)))))) * cast<uint64_t>(8ULL)));
r4300_CPU_write64(c,p,cast<uint64_t>((((r4300_CPU_read64(c,p) & ~((shl<uint64_t>(cast<uint64_t>(~cast<uint64_t>(cast<uint64_t>(0ULL))),shift))))) | (shl<uint64_t>(v,shift)))));
break;}
}}
}
}
// tools/cpu/r4300/exec.go:697:1
std::tuple<uint32_t,bool> r4300_addOv32(uint32_t a,uint32_t b){
{
uint32_t r = cast<uint32_t>((a + b));
return {r,(cast<uint32_t>((cast<uint32_t>(((cast<uint32_t>((a ^ r))) & (cast<uint32_t>((b ^ r))))) & cast<uint32_t>(2147483648ULL))) != cast<uint32_t>(0ULL))};
}
}
// tools/cpu/r4300/exec.go:701:1
std::tuple<uint32_t,bool> r4300_subOv32(uint32_t a,uint32_t b){
{
uint32_t r = cast<uint32_t>((a - b));
return {r,(cast<uint32_t>((cast<uint32_t>(((cast<uint32_t>((a ^ b))) & (cast<uint32_t>((a ^ r))))) & cast<uint32_t>(2147483648ULL))) != cast<uint32_t>(0ULL))};
}
}
// tools/cpu/r4300/exec.go:705:1
std::tuple<uint64_t,bool> r4300_addOv64(uint64_t a,uint64_t b){
{
uint64_t r = cast<uint64_t>((a + b));
return {r,(cast<uint64_t>((cast<uint64_t>(((cast<uint64_t>((a ^ r))) & (cast<uint64_t>((b ^ r))))) & cast<uint64_t>(9223372036854775808ULL))) != cast<uint64_t>(0ULL))};
}
}
// tools/cpu/r4300/exec.go:709:1
std::tuple<uint64_t,bool> r4300_subOv64(uint64_t a,uint64_t b){
{
uint64_t r = cast<uint64_t>((a - b));
return {r,(cast<uint64_t>((cast<uint64_t>(((cast<uint64_t>((a ^ b))) & (cast<uint64_t>((a ^ r))))) & cast<uint64_t>(9223372036854775808ULL))) != cast<uint64_t>(0ULL))};
}
}
// tools/cpu/r4300/exec.go:714:1
uint64_t r4300_b2u(bool b){
{
if (b) {
return cast<uint64_t>(1ULL);
}
return cast<uint64_t>(0ULL);
}
}
// tools/cpu/r4300/fpu.go:38:1
bool r4300_CPU_snanCheckS(r4300_CPU* c,uint32_t fd,uint32_t fs){
{
if ((!r4300_isSNaN32(r4300_CPU_readFGR32(c,fs)))) {
return false;
}
r4300_CPU_fpApply(c,r4300_fpInvalid);
r4300_CPU_writeFGR32(c,fd,r4300_qNaN32);
return true;
}
}
// tools/cpu/r4300/fpu.go:50:1
bool r4300_CPU_snanCheckD(r4300_CPU* c,uint32_t fd,uint32_t fs){
{
if ((!r4300_isSNaN64(r4300_CPU_readFGR64(c,fs)))) {
return false;
}
r4300_CPU_fpApply(c,r4300_fpInvalid);
r4300_CPU_writeFGR64(c,fd,r4300_qNaN64);
return true;
}
}
// tools/cpu/r4300/fpu.go:65:1
bool r4300_CPU_snanCheckSToD(r4300_CPU* c,uint32_t fd,uint32_t fs){
{
if ((!r4300_isSNaN32(r4300_CPU_readFGR32(c,fs)))) {
return false;
}
r4300_CPU_fpApply(c,r4300_fpInvalid);
r4300_CPU_writeFGR64(c,fd,r4300_qNaN64);
return true;
}
}
// tools/cpu/r4300/fpu.go:74:1
bool r4300_CPU_snanCheckDToS(r4300_CPU* c,uint32_t fd,uint32_t fs){
{
if ((!r4300_isSNaN64(r4300_CPU_readFGR64(c,fs)))) {
return false;
}
r4300_CPU_fpApply(c,r4300_fpInvalid);
r4300_CPU_writeFGR32(c,fd,r4300_qNaN32);
return true;
}
}
// tools/cpu/r4300/fpu.go:83:1
void r4300_CPU_setSResult(r4300_CPU* c,uint32_t i,float v){
{
if (go_math_IsNaN(cast<double>(v))) {
r4300_CPU_writeFGR32(c,i,r4300_qNaN32);
return ;
}
r4300_CPU_setS(c,i,v);
}
}
// tools/cpu/r4300/fpu.go:91:1
void r4300_CPU_setDResult(r4300_CPU* c,uint32_t i,double v){
{
if (go_math_IsNaN(v)) {
r4300_CPU_writeFGR64(c,i,r4300_qNaN64);
return ;
}
r4300_CPU_setD(c,i,v);
}
}
// tools/cpu/r4300/fpu.go:115:1
void r4300_CPU_clearCause(r4300_CPU* c){
{
c->FCR31 &= ~(r4300_fcr31CauseMask);
}
}
// tools/cpu/r4300/fpu.go:125:1
void r4300_CPU_fpUnimplemented(r4300_CPU* c){
{
r4300_CPU_fpApply(c,r4300_fpUnimpl);
}
}
// tools/cpu/r4300/fpu.go:138:1
uint32_t r4300_CPU_readFGR32(r4300_CPU* c,uint32_t i){
{
if ((cast<uint64_t>((c->COP0[r4300_cop0Status] & r4300_statusFR)) != cast<uint64_t>(0ULL))) {
return cast<uint32_t>(c->FGR[i]);
}
if ((cast<uint32_t>((i & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL))) {
return cast<uint32_t>(shr<uint64_t>(c->FGR[(i & ~(cast<uint32_t>(1ULL)))],cast<int64_t>(32ULL)));
}
return cast<uint32_t>(c->FGR[i]);
}
}
// tools/cpu/r4300/fpu.go:148:1
void r4300_CPU_writeFGR32(r4300_CPU* c,uint32_t i,uint32_t v){
{
if ((cast<uint64_t>((c->COP0[r4300_cop0Status] & r4300_statusFR)) != cast<uint64_t>(0ULL))) {
c->FGR[i] = cast<uint64_t>(v);
return ;
}
if ((cast<uint32_t>((i & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL))) {
c->FGR[(i & ~(cast<uint32_t>(1ULL)))] = cast<uint64_t>((cast<uint64_t>((c->FGR[(i & ~(cast<uint32_t>(1ULL)))] & cast<uint64_t>(4294967295ULL))) | shl<uint64_t>(cast<uint64_t>(v),cast<int64_t>(32ULL))));
return ;
}
c->FGR[i] = cast<uint64_t>((cast<uint64_t>((c->FGR[i] & cast<uint64_t>(18446744069414584320ULL))) | cast<uint64_t>(v)));
}
}
// tools/cpu/r4300/fpu.go:161:1
uint64_t r4300_CPU_readFGR64(r4300_CPU* c,uint32_t i){
{
if ((cast<uint64_t>((c->COP0[r4300_cop0Status] & r4300_statusFR)) == cast<uint64_t>(0ULL))) {
i &= ~(cast<uint32_t>(1ULL));
}
return c->FGR[i];
}
}
// tools/cpu/r4300/fpu.go:168:1
void r4300_CPU_writeFGR64(r4300_CPU* c,uint32_t i,uint64_t v){
{
if ((cast<uint64_t>((c->COP0[r4300_cop0Status] & r4300_statusFR)) == cast<uint64_t>(0ULL))) {
i &= ~(cast<uint32_t>(1ULL));
}
c->FGR[i] = v;
}
}
// tools/cpu/r4300/fpu.go:175:1
float r4300_CPU_fs(r4300_CPU* c,uint32_t i){
{
return go_math_Float32frombits(r4300_CPU_readFGR32(c,i));
}
}
// tools/cpu/r4300/fpu.go:176:1
double r4300_CPU_fd(r4300_CPU* c,uint32_t i){
{
return go_math_Float64frombits(r4300_CPU_readFGR64(c,i));
}
}
// tools/cpu/r4300/fpu.go:177:1
void r4300_CPU_setS(r4300_CPU* c,uint32_t i,float v){
{
r4300_CPU_writeFGR32(c,i,go_math_Float32bits(v));
}
}
// tools/cpu/r4300/fpu.go:178:1
void r4300_CPU_setD(r4300_CPU* c,uint32_t i,double v){
{
r4300_CPU_writeFGR64(c,i,go_math_Float64bits(v));
}
}
// tools/cpu/r4300/fpu.go:181:1
double r4300_CPU_round(r4300_CPU* c,double v){
{
{
switch(cast<uint32_t>((c->FCR31 & cast<uint32_t>(3ULL)))){
case r4300_roundToZero:{
return go_math_Trunc(v);
break;}
case r4300_roundCeil:{
return go_math_Ceil(v);
break;}
case r4300_roundFloor:{
return go_math_Floor(v);
break;}
}}
return go_math_RoundToEven(v);
}
}
// tools/cpu/r4300/fpu.go:195:1
void r4300_CPU_cop1Mem(r4300_CPU* c,uint32_t op,uint32_t rs,uint32_t ft,uint64_t simm){
{
uint64_t vaddr = cast<uint64_t>((c->R[rs] + simm));
{
switch(op){
case cast<uint32_t>(49ULL):{
if ((cast<uint64_t>((vaddr & cast<uint64_t>(3ULL))) != cast<uint64_t>(0ULL))) {
r4300_CPU_addrError(c,r4300_excAdEL,vaddr);
return ;
}
auto tmp61 = r4300_CPU_Translate(c,vaddr,false);
uint32_t p = std::get<0>(tmp61);
bool ok = std::get<1>(tmp61);
if ((!ok)) {
return ;
}
r4300_CPU_writeFGR32(c,ft,r4300_CPU_read32(c,p));
break;}
case cast<uint32_t>(53ULL):{
if ((cast<uint64_t>((vaddr & cast<uint64_t>(7ULL))) != cast<uint64_t>(0ULL))) {
r4300_CPU_addrError(c,r4300_excAdEL,vaddr);
return ;
}
auto tmp62 = r4300_CPU_Translate(c,vaddr,false);
uint32_t p = std::get<0>(tmp62);
bool ok = std::get<1>(tmp62);
if ((!ok)) {
return ;
}
r4300_CPU_writeFGR64(c,ft,r4300_CPU_read64(c,p));
break;}
case cast<uint32_t>(57ULL):{
if ((cast<uint64_t>((vaddr & cast<uint64_t>(3ULL))) != cast<uint64_t>(0ULL))) {
r4300_CPU_addrError(c,r4300_excAdES,vaddr);
return ;
}
auto tmp63 = r4300_CPU_Translate(c,vaddr,true);
uint32_t p = std::get<0>(tmp63);
bool ok = std::get<1>(tmp63);
if ((!ok)) {
return ;
}
r4300_CPU_write32(c,p,r4300_CPU_readFGR32(c,ft));
break;}
case cast<uint32_t>(61ULL):{
if ((cast<uint64_t>((vaddr & cast<uint64_t>(7ULL))) != cast<uint64_t>(0ULL))) {
r4300_CPU_addrError(c,r4300_excAdES,vaddr);
return ;
}
auto tmp64 = r4300_CPU_Translate(c,vaddr,true);
uint32_t p = std::get<0>(tmp64);
bool ok = std::get<1>(tmp64);
if ((!ok)) {
return ;
}
r4300_CPU_write64(c,p,r4300_CPU_readFGR64(c,ft));
break;}
}}
}
}
// tools/cpu/r4300/fpu.go:243:1
void r4300_CPU_cop1(r4300_CPU* c,uint32_t w,uint32_t rs,uint32_t rt,uint32_t rd,uint32_t shamt,uint64_t branchT){
{
{
switch(rs){
case cast<uint32_t>(0ULL):{
r4300_CPU_set(c,rt,r4300_sext32(r4300_CPU_readFGR32(c,rd)));
return ;
break;}
case cast<uint32_t>(1ULL):{
r4300_CPU_set(c,rt,r4300_CPU_readFGR64(c,rd));
return ;
break;}
case cast<uint32_t>(2ULL):{
{
switch(rd){
case cast<uint32_t>(0ULL):{
r4300_CPU_set(c,rt,cast<uint64_t>(2560ULL));
break;}
case cast<uint32_t>(31ULL):{
r4300_CPU_set(c,rt,r4300_sext32(c->FCR31));
break;}
}}
return ;
break;}
case cast<uint32_t>(4ULL):{
r4300_CPU_writeFGR32(c,rd,cast<uint32_t>(c->R[rt]));
return ;
break;}
case cast<uint32_t>(5ULL):{
r4300_CPU_writeFGR64(c,rd,c->R[rt]);
return ;
break;}
case cast<uint32_t>(6ULL):{
if ((rd == cast<uint32_t>(31ULL))) {
c->FCR31 = cast<uint32_t>(c->R[rt]);
}
return ;
break;}
case cast<uint32_t>(8ULL):{
bool cond = (cast<uint32_t>((c->FCR31 & r4300_fcr31Cond)) != cast<uint32_t>(0ULL));
{
switch(cast<uint32_t>((rt & cast<uint32_t>(3ULL)))){
case cast<uint32_t>(0ULL):{
r4300_CPU_doBranch(c,(!cond),branchT);
break;}
case cast<uint32_t>(1ULL):{
r4300_CPU_doBranch(c,cond,branchT);
break;}
case cast<uint32_t>(2ULL):{
r4300_CPU_doBranchLikely(c,(!cond),branchT);
break;}
case cast<uint32_t>(3ULL):{
r4300_CPU_doBranchLikely(c,cond,branchT);
break;}
}}
return ;
break;}
}}
uint32_t funct = cast<uint32_t>((w & cast<uint32_t>(63ULL)));
auto tmp65 = std::make_tuple(rt,rd,shamt);
uint32_t ft = std::get<0>(tmp65);
uint32_t fs = std::get<1>(tmp65);
uint32_t fd = std::get<2>(tmp65);
{
switch(rs){
case cast<uint32_t>(16ULL):{
r4300_CPU_cop1Single(c,w,funct,ft,fs,fd);
break;}
case cast<uint32_t>(17ULL):{
r4300_CPU_cop1Double(c,w,funct,ft,fs,fd);
break;}
case cast<uint32_t>(20ULL):{
r4300_CPU_clearCause(c);
{
switch(funct){
case cast<uint32_t>(32ULL):{
r4300_CPU_setSResult(c,fd,cast<float>(cast<int32_t>(r4300_CPU_readFGR32(c,fs))));
break;}
case cast<uint32_t>(33ULL):{
r4300_CPU_setD(c,fd,cast<double>(cast<int32_t>(r4300_CPU_readFGR32(c,fs))));
break;}
default:{
r4300_CPU_fpUnimplemented(c);
break;}
}}
break;}
case cast<uint32_t>(21ULL):{
r4300_CPU_clearCause(c);
{
switch(funct){
case cast<uint32_t>(32ULL):{
r4300_CPU_setS(c,fd,cast<float>(cast<int64_t>(r4300_CPU_readFGR64(c,fs))));
break;}
case cast<uint32_t>(33ULL):{
r4300_CPU_setD(c,fd,cast<double>(cast<int64_t>(r4300_CPU_readFGR64(c,fs))));
break;}
default:{
r4300_CPU_fpUnimplemented(c);
break;}
}}
break;}
default:{
r4300_CPU_fpUnimplemented(c);
break;}
}}
}
}
// tools/cpu/r4300/fpu.go:319:1
void r4300_CPU_cop1Single(r4300_CPU* c,uint32_t w,uint32_t funct,uint32_t ft,uint32_t fs,uint32_t fd){
{
if ((funct != cast<uint32_t>(6ULL))) {
r4300_CPU_clearCause(c);
}
if ((cast<uint32_t>((funct & cast<uint32_t>(48ULL))) == cast<uint32_t>(48ULL))) {
r4300_CPU_compare(c,cast<double>(r4300_CPU_fs(c,fs)),cast<double>(r4300_CPU_fs(c,ft)),cast<uint32_t>((funct & cast<uint32_t>(15ULL))),(r4300_isSNaN32(r4300_CPU_readFGR32(c,fs)) || r4300_isSNaN32(r4300_CPU_readFGR32(c,ft))));
return ;
}
{
switch(funct){
case cast<uint32_t>(0ULL):case cast<uint32_t>(1ULL):case cast<uint32_t>(2ULL):case cast<uint32_t>(3ULL):{
bool snan = (r4300_isSNaN32(r4300_CPU_readFGR32(c,fs)) || r4300_isSNaN32(r4300_CPU_readFGR32(c,ft)));
auto tmp66 = r4300_fpArith(cast<uint32_t>((c->FCR31 & cast<uint32_t>(3ULL))),"+-*/"[funct],cast<double>(r4300_CPU_fs(c,fs)),cast<double>(r4300_CPU_fs(c,ft)),true,snan);
double r = std::get<0>(tmp66);
uint32_t cond = std::get<1>(tmp66);
if (r4300_CPU_fpApply(c,cond)) {
return ;
}
r4300_CPU_setSResult(c,fd,cast<float>(r));
break;}
case cast<uint32_t>(4ULL):{
double x = cast<double>(r4300_CPU_fs(c,fs));
if (((x < cast<double>(0ULL)) || r4300_isSNaN32(r4300_CPU_readFGR32(c,fs)))) {
r4300_CPU_fpApply(c,r4300_fpInvalid);
r4300_CPU_setSResult(c,fd,cast<float>(go_math_NaN()));
break;
}
double r = go_math_Sqrt(x);
if (r4300_CPU_fpApply(c,r4300_sqrtConditions(r,x))) {
break;
}
r4300_CPU_setSResult(c,fd,cast<float>(r));
break;}
case cast<uint32_t>(5ULL):{
if (r4300_CPU_snanCheckS(c,fd,fs)) {
break;
}
r4300_CPU_setSResult(c,fd,cast<float>(go_math_Abs(cast<double>(r4300_CPU_fs(c,fs)))));
break;}
case cast<uint32_t>(6ULL):{
r4300_CPU_writeFGR32(c,fd,r4300_CPU_readFGR32(c,fs));
break;}
case cast<uint32_t>(7ULL):{
if (r4300_CPU_snanCheckS(c,fd,fs)) {
break;
}
r4300_CPU_setSResult(c,fd,cast<float>(-r4300_CPU_fs(c,fs)));
break;}
case cast<uint32_t>(8ULL):{
r4300_CPU_writeFGR64(c,fd,cast<uint64_t>(cast<int64_t>(r4300_CPU_round(c,cast<double>(r4300_CPU_fs(c,fs))))));
break;}
case cast<uint32_t>(9ULL):{
r4300_CPU_writeFGR64(c,fd,cast<uint64_t>(cast<int64_t>(go_math_Trunc(cast<double>(r4300_CPU_fs(c,fs))))));
break;}
case cast<uint32_t>(10ULL):{
r4300_CPU_writeFGR64(c,fd,cast<uint64_t>(cast<int64_t>(go_math_Ceil(cast<double>(r4300_CPU_fs(c,fs))))));
break;}
case cast<uint32_t>(11ULL):{
r4300_CPU_writeFGR64(c,fd,cast<uint64_t>(cast<int64_t>(go_math_Floor(cast<double>(r4300_CPU_fs(c,fs))))));
break;}
case cast<uint32_t>(12ULL):{
r4300_CPU_toInt32(c,fd,cast<double>(r4300_CPU_fs(c,fs)),go_math_RoundToEven);
break;}
case cast<uint32_t>(13ULL):{
r4300_CPU_toInt32(c,fd,cast<double>(r4300_CPU_fs(c,fs)),go_math_Trunc);
break;}
case cast<uint32_t>(14ULL):{
r4300_CPU_toInt32(c,fd,cast<double>(r4300_CPU_fs(c,fs)),go_math_Ceil);
break;}
case cast<uint32_t>(15ULL):{
r4300_CPU_toInt32(c,fd,cast<double>(r4300_CPU_fs(c,fs)),go_math_Floor);
break;}
case cast<uint32_t>(33ULL):{
if (r4300_CPU_snanCheckSToD(c,fd,fs)) {
break;
}
r4300_CPU_setDResult(c,fd,cast<double>(r4300_CPU_fs(c,fs)));
break;}
case cast<uint32_t>(36ULL):{
r4300_CPU_toInt32(c,fd,cast<double>(r4300_CPU_fs(c,fs)),[=](auto...args){return r4300_CPU_round(c,args...);});
break;}
case cast<uint32_t>(37ULL):{
r4300_CPU_toInt64(c,fd,cast<double>(r4300_CPU_fs(c,fs)),[=](auto...args){return r4300_CPU_round(c,args...);});
break;}
default:{
r4300_CPU_fpUnimplemented(c);
break;}
}}
}
}
// tools/cpu/r4300/fpu.go:400:1
void r4300_CPU_cop1Double(r4300_CPU* c,uint32_t w,uint32_t funct,uint32_t ft,uint32_t fs,uint32_t fd){
{
if ((funct != cast<uint32_t>(6ULL))) {
r4300_CPU_clearCause(c);
}
if ((cast<uint32_t>((funct & cast<uint32_t>(48ULL))) == cast<uint32_t>(48ULL))) {
r4300_CPU_compare(c,r4300_CPU_fd(c,fs),r4300_CPU_fd(c,ft),cast<uint32_t>((funct & cast<uint32_t>(15ULL))),(r4300_isSNaN64(r4300_CPU_readFGR64(c,fs)) || r4300_isSNaN64(r4300_CPU_readFGR64(c,ft))));
return ;
}
{
switch(funct){
case cast<uint32_t>(0ULL):case cast<uint32_t>(1ULL):case cast<uint32_t>(2ULL):case cast<uint32_t>(3ULL):{
bool snan = (r4300_isSNaN64(r4300_CPU_readFGR64(c,fs)) || r4300_isSNaN64(r4300_CPU_readFGR64(c,ft)));
auto tmp67 = r4300_fpArith(cast<uint32_t>((c->FCR31 & cast<uint32_t>(3ULL))),"+-*/"[funct],r4300_CPU_fd(c,fs),r4300_CPU_fd(c,ft),false,snan);
double r = std::get<0>(tmp67);
uint32_t cond = std::get<1>(tmp67);
if (r4300_CPU_fpApply(c,cond)) {
return ;
}
r4300_CPU_setDResult(c,fd,r);
break;}
case cast<uint32_t>(4ULL):{
double x = r4300_CPU_fd(c,fs);
if (((x < cast<double>(0ULL)) || r4300_isSNaN64(r4300_CPU_readFGR64(c,fs)))) {
r4300_CPU_fpApply(c,r4300_fpInvalid);
r4300_CPU_setDResult(c,fd,go_math_NaN());
break;
}
double r = go_math_Sqrt(x);
if (r4300_CPU_fpApply(c,r4300_sqrtConditions(r,x))) {
break;
}
r4300_CPU_setDResult(c,fd,r);
break;}
case cast<uint32_t>(5ULL):{
if (r4300_CPU_snanCheckD(c,fd,fs)) {
break;
}
r4300_CPU_setDResult(c,fd,go_math_Abs(r4300_CPU_fd(c,fs)));
break;}
case cast<uint32_t>(6ULL):{
r4300_CPU_writeFGR64(c,fd,r4300_CPU_readFGR64(c,fs));
break;}
case cast<uint32_t>(7ULL):{
if (r4300_CPU_snanCheckD(c,fd,fs)) {
break;
}
r4300_CPU_setDResult(c,fd,cast<double>(-r4300_CPU_fd(c,fs)));
break;}
case cast<uint32_t>(8ULL):{
r4300_CPU_writeFGR64(c,fd,cast<uint64_t>(cast<int64_t>(r4300_CPU_round(c,r4300_CPU_fd(c,fs)))));
break;}
case cast<uint32_t>(9ULL):{
r4300_CPU_writeFGR64(c,fd,cast<uint64_t>(cast<int64_t>(go_math_Trunc(r4300_CPU_fd(c,fs)))));
break;}
case cast<uint32_t>(10ULL):{
r4300_CPU_writeFGR64(c,fd,cast<uint64_t>(cast<int64_t>(go_math_Ceil(r4300_CPU_fd(c,fs)))));
break;}
case cast<uint32_t>(11ULL):{
r4300_CPU_writeFGR64(c,fd,cast<uint64_t>(cast<int64_t>(go_math_Floor(r4300_CPU_fd(c,fs)))));
break;}
case cast<uint32_t>(12ULL):{
r4300_CPU_toInt32(c,fd,r4300_CPU_fd(c,fs),go_math_RoundToEven);
break;}
case cast<uint32_t>(13ULL):{
r4300_CPU_toInt32(c,fd,r4300_CPU_fd(c,fs),go_math_Trunc);
break;}
case cast<uint32_t>(14ULL):{
r4300_CPU_toInt32(c,fd,r4300_CPU_fd(c,fs),go_math_Ceil);
break;}
case cast<uint32_t>(15ULL):{
r4300_CPU_toInt32(c,fd,r4300_CPU_fd(c,fs),go_math_Floor);
break;}
case cast<uint32_t>(32ULL):{
if (r4300_CPU_snanCheckDToS(c,fd,fs)) {
break;
}
double v = r4300_CPU_fd(c,fs);
if ((cast<double>(cast<float>(v)) != v)) {
r4300_CPU_fpApply(c,r4300_fpInexact);
}
r4300_CPU_setSResult(c,fd,cast<float>(v));
break;}
case cast<uint32_t>(36ULL):{
r4300_CPU_toInt32(c,fd,r4300_CPU_fd(c,fs),[=](auto...args){return r4300_CPU_round(c,args...);});
break;}
case cast<uint32_t>(37ULL):{
r4300_CPU_toInt64(c,fd,r4300_CPU_fd(c,fs),[=](auto...args){return r4300_CPU_round(c,args...);});
break;}
default:{
r4300_CPU_fpUnimplemented(c);
break;}
}}
}
}
// tools/cpu/r4300/fpu.go:483:1
void r4300_CPU_compare(r4300_CPU* c,double a,double b,uint32_t cond,bool snanIn){
{
bool unordered = (go_math_IsNaN(a) || go_math_IsNaN(b));
if (unordered) {
bool signalling = ((cast<uint32_t>((cond & cast<uint32_t>(8ULL))) != cast<uint32_t>(0ULL)) || snanIn);
if ((signalling && r4300_CPU_fpApply(c,r4300_fpInvalid))) {
return ;
}
if ((!signalling)) {
r4300_CPU_clearCause(c);
}
}
bool r={};
if (unordered) {
r = (cast<uint32_t>((cond & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL));
}
else {
r = ((((cast<uint32_t>((cond & cast<uint32_t>(4ULL))) != cast<uint32_t>(0ULL)) && (a < b))) || (((cast<uint32_t>((cond & cast<uint32_t>(2ULL))) != cast<uint32_t>(0ULL)) && (a == b))));
}
if (r) {
c->FCR31 |= r4300_fcr31Cond;
}
else {
c->FCR31 &= ~(r4300_fcr31Cond);
}
}
}
// tools/cpu/r4300/fpu.go:512:1
void r4300_CPU_toInt32(r4300_CPU* c,uint32_t fd,double v,std::function<double(double)> round){
{
double r = round(v);
uint32_t cond={};
if ((r != v)) {
cond |= r4300_fpInexact;
}
if (r4300_CPU_fpApply(c,cond)) {
return ;
}
r4300_CPU_writeFGR32(c,fd,cast<uint32_t>(cast<int32_t>(r)));
}
}
// tools/cpu/r4300/fpu.go:524:1
void r4300_CPU_toInt64(r4300_CPU* c,uint32_t fd,double v,std::function<double(double)> round){
{
double r = round(v);
uint32_t cond={};
if ((r != v)) {
cond |= r4300_fpInexact;
}
if (r4300_CPU_fpApply(c,cond)) {
return ;
}
r4300_CPU_writeFGR64(c,fd,cast<uint64_t>(cast<int64_t>(r)));
}
}
// tools/cpu/r4300/fpu_flags.go:22:1
bool r4300_isNaN(double v){
{
return go_math_IsNaN(v);
}
}
// tools/cpu/r4300/fpu_flags.go:23:1
bool r4300_isInf(double v){
{
return go_math_IsInf(v,cast<int64_t>(0ULL));
}
}
// tools/cpu/r4300/fpu_flags.go:24:1
double r4300_nan(){
{
return go_math_NaN();
}
}
// tools/cpu/r4300/fpu_flags.go:26:1
double r4300_sign(double v){
{
if (go_math_Signbit(v)) {
return cast<double>(-cast<int64_t>(1ULL));
}
return cast<double>(1ULL);
}
}
// tools/cpu/r4300/fpu_flags.go:33:1
double r4300_inf(double s){
{
if ((s < cast<double>(0ULL))) {
return go_math_Inf(cast<int64_t>(-cast<int64_t>(1ULL)));
}
return go_math_Inf(cast<int64_t>(1ULL));
}
}
// tools/cpu/r4300/fpu_flags.go:40:1
double r4300_zero(double s){
{
if ((s < cast<double>(0ULL))) {
return go_math_Copysign(cast<double>(0ULL),cast<double>(-cast<int64_t>(1ULL)));
}
return cast<double>(0ULL);
}
}
// tools/cpu/r4300/fpu_flags.go:75:1
bool r4300_isSNaN32(uint32_t bits){
{
return (((cast<uint32_t>((bits & cast<uint32_t>(2139095040ULL))) == cast<uint32_t>(2139095040ULL)) && (cast<uint32_t>((bits & cast<uint32_t>(8388607ULL))) != cast<uint32_t>(0ULL))) && (cast<uint32_t>((bits & cast<uint32_t>(4194304ULL))) != cast<uint32_t>(0ULL)));
}
}
// tools/cpu/r4300/fpu_flags.go:79:1
bool r4300_isSNaN64(uint64_t bits){
{
return (((cast<uint64_t>((bits & cast<uint64_t>(9218868437227405312ULL))) == cast<uint64_t>(9218868437227405312ULL)) && (cast<uint64_t>((bits & cast<uint64_t>(4503599627370495ULL))) != cast<uint64_t>(0ULL))) && (cast<uint64_t>((bits & (shl<uint64_t>(cast<int64_t>(1ULL),cast<int64_t>(51ULL))))) != cast<uint64_t>(0ULL)));
}
}
// tools/cpu/r4300/fpu_flags.go:96:1
bool r4300_CPU_fpApply(r4300_CPU* c,uint32_t cond){
{
uint32_t enables = cast<uint32_t>(((shr<uint32_t>(c->FCR31,r4300_fcr31EnableShift)) & cast<uint32_t>(31ULL)));
if ((cast<uint32_t>((cond & r4300_fpTiny)) != cast<uint32_t>(0ULL))) {
cond &= ~(r4300_fpTiny);
if (((cast<uint32_t>((enables & (cast<uint32_t>((r4300_fpUnderflow | r4300_fpInexact))))) != cast<uint32_t>(0ULL)) || (cast<uint32_t>((c->FCR31 & r4300_fcr31FS)) == cast<uint32_t>(0ULL)))) {
cond = r4300_fpUnimpl;
}
}
c->FCR31 = cast<uint32_t>((((c->FCR31 & ~(r4300_fcr31CauseMask))) | shl<uint32_t>(cond,r4300_fcr31CauseShift)));
if ((cast<uint32_t>((cond & r4300_fpUnimpl)) != cast<uint32_t>(0ULL))) {
r4300_CPU_Exception(c,r4300_excFPE);
return true;
}
if ((cast<uint32_t>((cond & enables)) != cast<uint32_t>(0ULL))) {
r4300_CPU_Exception(c,r4300_excFPE);
return true;
}
c->FCR31 |= shl<uint32_t>(cond,r4300_fcr31FlagShift);
return false;
}
}
// tools/cpu/r4300/fpu_flags.go:144:1
bool r4300_denormal(double v,bool single){
{
if (single) {
return r4300_isSubnormal32(cast<float>(v));
}
return r4300_isSubnormal64(v);
}
}
// tools/cpu/r4300/fpu_flags.go:179:1
double r4300_applyRounding(uint32_t rm,double rounded,int64_t dsign,bool single){
{
if (((dsign == cast<int64_t>(0ULL)) || (rm == r4300_roundNearest))) {
return rounded;
}
std::function<double()> up = [&]()->double{
return r4300_next(rounded,go_math_Inf(cast<int64_t>(1ULL)),single);
}
;
std::function<double()> down = [&]()->double{
return r4300_next(rounded,go_math_Inf(cast<int64_t>(-cast<int64_t>(1ULL))),single);
}
;
{
switch(rm){
case r4300_roundToZero:{
if (((rounded > cast<double>(0ULL)) && (dsign < cast<int64_t>(0ULL)))) {
return down();
}
if (((rounded < cast<double>(0ULL)) && (dsign > cast<int64_t>(0ULL)))) {
return up();
}
break;}
case r4300_roundCeil:{
if ((dsign > cast<int64_t>(0ULL))) {
return up();
}
break;}
case r4300_roundFloor:{
if ((dsign < cast<int64_t>(0ULL))) {
return down();
}
break;}
}}
return rounded;
}
}
// tools/cpu/r4300/fpu_flags.go:208:1
double r4300_next(double v,double toward,bool single){
{
if (single) {
return cast<double>(go_math_Nextafter32(cast<float>(v),cast<float>(toward)));
}
return go_math_Nextafter(v,toward);
}
}
// tools/cpu/r4300/fpu_flags.go:215:1
int64_t r4300_signOf(double v){
{
{
if ((v > cast<double>(0ULL))){
return cast<int64_t>(1ULL);
}
else if ((v < cast<double>(0ULL))){
return cast<int64_t>(-cast<int64_t>(1ULL));
}
}
tmp68:;
return cast<int64_t>(0ULL);
}
}
// tools/cpu/r4300/fpu_flags.go:228:1
std::tuple<double,uint32_t> r4300_classifySingle(double rounded,bool inexact){
{
uint32_t cond={};
float f = cast<float>(rounded);
if ((inexact || (cast<double>(f) != rounded))) {
cond |= r4300_fpInexact;
}
{
if ((r4300_isInf(cast<double>(f)) && (!r4300_isInf(rounded)))){
cond |= cast<uint32_t>((r4300_fpOverflow | r4300_fpInexact));
}
else if ((r4300_isSubnormal32(f) || (((f == cast<float>(0ULL)) && (rounded != cast<double>(0ULL)))))){
return {r4300_zero(r4300_sign(rounded)),cast<uint32_t>((cast<int64_t>((r4300_fpUnderflow | r4300_fpInexact)) | r4300_fpTiny))};
}
}
tmp69:;
return {cast<double>(f),cond};
}
}
// tools/cpu/r4300/fpu_flags.go:244:1
bool r4300_isSubnormal32(float f){
{
return ((f != cast<float>(0ULL)) && (go_math_Abs(cast<double>(f)) < 1.1754943508222875e-38));
}
}
// tools/cpu/r4300/fpu_flags.go:248:1
bool r4300_isSubnormal64(double v){
{
return ((v != cast<double>(0ULL)) && (go_math_Abs(v) < 2.2250738585072014e-308));
}
}
// tools/cpu/r4300/fpu_flags.go:253:1
std::tuple<double,uint32_t> r4300_classifyDouble(double r,bool inexact){
{
uint32_t cond={};
if (inexact) {
cond |= r4300_fpInexact;
}
{
if (r4300_isInf(r)){
cond |= cast<uint32_t>((r4300_fpOverflow | r4300_fpInexact));
}
else if ((r4300_isSubnormal64(r) || (((r == cast<double>(0ULL)) && inexact)))){
return {r4300_zero(r4300_sign(r)),cast<uint32_t>((cast<int64_t>((r4300_fpUnderflow | r4300_fpInexact)) | r4300_fpTiny))};
}
}
tmp70:;
return {r,cond};
}
}
// tools/cpu/r4300/fpu_flags.go:270:1
std::tuple<double,uint32_t> r4300_fpArith(uint32_t rm,uint8_t op,double a,double b,bool single,bool snanIn){
{
if (snanIn) {
return {r4300_nan(),r4300_fpInvalid};
}
if ((r4300_isNaN(a) || r4300_isNaN(b))) {
return {r4300_nan(),cast<uint32_t>(0ULL)};
}
if ((r4300_denormal(a,single) || r4300_denormal(b,single))) {
return {cast<double>(0ULL),r4300_fpUnimpl};
}
{
switch(op){
case '/':{
if ((b == cast<double>(0ULL))) {
if ((a == cast<double>(0ULL))) {
return {r4300_nan(),r4300_fpInvalid};
}
return {r4300_inf((r4300_sign(a) * r4300_sign(b))),r4300_fpDivByZero};
}
if ((r4300_isInf(a) && r4300_isInf(b))) {
return {r4300_nan(),r4300_fpInvalid};
}
break;}
case '*':{
if ((((r4300_isInf(a) && (b == cast<double>(0ULL)))) || (((a == cast<double>(0ULL)) && r4300_isInf(b))))) {
return {r4300_nan(),r4300_fpInvalid};
}
break;}
case '+':{
if (((r4300_isInf(a) && r4300_isInf(b)) && (r4300_sign(a) != r4300_sign(b)))) {
return {r4300_nan(),r4300_fpInvalid};
}
break;}
case '-':{
if (((r4300_isInf(a) && r4300_isInf(b)) && (r4300_sign(a) == r4300_sign(b)))) {
return {r4300_nan(),r4300_fpInvalid};
}
break;}
}}
if ((r4300_isInf(a) || r4300_isInf(b))) {
return {r4300_infiniteResult(op,a,b),cast<uint32_t>(0ULL)};
}
if (single) {
{
switch(op){
case '+':case '-':{
if ((op == '-')) {
b = cast<double>(-b);
}
double s = (a + b);
double err = r4300_twoSumResidual(a,b,s);
auto tmp71 = r4300_classifySingle(s,(err != cast<double>(0ULL)));
double v = std::get<0>(tmp71);
uint32_t cond = std::get<1>(tmp71);
double d = (((s - v)) + err);
return {r4300_applyRounding(rm,v,r4300_signOf(d),true),cond};
break;}
case '*':{
double p = (a * b);
auto tmp72 = r4300_classifySingle(p,false);
double v = std::get<0>(tmp72);
uint32_t cond = std::get<1>(tmp72);
return {r4300_applyRounding(rm,v,r4300_signOf((p - v)),true),cond};
break;}
}}
double q = (a / b);
double res = go_math_FMA(cast<double>(cast<float>(q)),b,cast<double>(-a));
auto tmp73 = r4300_classifySingle(q,(res != cast<double>(0ULL)));
double v = std::get<0>(tmp73);
uint32_t cond = std::get<1>(tmp73);
return {r4300_applyRounding(rm,v,cast<int64_t>((cast<int64_t>(-r4300_signOf(res)) * r4300_signOf(b))),true),cond};
}
{
switch(op){
case '+':case '-':{
if ((op == '-')) {
b = cast<double>(-b);
}
double r = (a + b);
double err = r4300_twoSumResidual(a,b,r);
auto tmp74 = r4300_classifyDouble(r,(err != cast<double>(0ULL)));
double v = std::get<0>(tmp74);
uint32_t cond = std::get<1>(tmp74);
return {r4300_applyRounding(rm,v,r4300_signOf(err),false),cond};
break;}
case '*':{
double r = (a * b);
if (r4300_isInf(r)) {
return {r,cast<uint32_t>((r4300_fpOverflow | r4300_fpInexact))};
}
double err = go_math_FMA(a,b,cast<double>(-r));
auto tmp75 = r4300_classifyDouble(r,(err != cast<double>(0ULL)));
double v = std::get<0>(tmp75);
uint32_t cond = std::get<1>(tmp75);
return {r4300_applyRounding(rm,v,r4300_signOf(err),false),cond};
break;}
default:{
double r = (a / b);
double res = go_math_FMA(r,b,cast<double>(-a));
auto tmp76 = r4300_classifyDouble(r,(res != cast<double>(0ULL)));
double v = std::get<0>(tmp76);
uint32_t cond = std::get<1>(tmp76);
return {r4300_applyRounding(rm,v,cast<int64_t>((cast<int64_t>(-r4300_signOf(res)) * r4300_signOf(b))),false),cond};
break;}
}}
}
}
// tools/cpu/r4300/fpu_flags.go:363:1
double r4300_twoSumResidual(double a,double b,double s){
{
double bb = (s - a);
return (((a - ((s - bb)))) + ((b - bb)));
}
}
// tools/cpu/r4300/fpu_flags.go:370:1
double r4300_infiniteResult(uint8_t op,double a,double b){
{
{
switch(op){
case '+':{
if (r4300_isInf(a)) {
return a;
}
return b;
break;}
case '-':{
if (r4300_isInf(a)) {
return a;
}
return cast<double>(-b);
break;}
case '*':{
return r4300_inf((r4300_sign(a) * r4300_sign(b)));
break;}
default:{
if (r4300_isInf(a)) {
return r4300_inf((r4300_sign(a) * r4300_sign(b)));
}
return r4300_zero((r4300_sign(a) * r4300_sign(b)));
break;}
}}
}
}
// tools/cpu/r4300/fpu_flags.go:394:1
uint32_t r4300_sqrtConditions(double r,double x){
{
if ((((x == cast<double>(0ULL)) || r4300_isInf(x)) || r4300_isNaN(x))) {
return cast<uint32_t>(0ULL);
}
if ((go_math_FMA(r,r,cast<double>(-x)) != cast<double>(0ULL))) {
return r4300_fpInexact;
}
return cast<uint32_t>(0ULL);
}
}
// tools/cpu/rsp/cpu.go:67:1
rsp_CPU* rsp_NewCPU(Slice<uint8_t> dmem,Slice<uint8_t> imem,n64_Machine* regs){
{
rsp_CPU* c = new rsp_CPU{{},{},{},{},{},{},{},{},{},dmem,imem,{},{},{},{},{},{},{},regs,{}};
rsp_CPU_Reset(c);
return c;
}
}
// tools/cpu/rsp/cpu.go:74:1
void rsp_CPU_Reset(rsp_CPU* c){
{
c->R = std::array<uint32_t,32>{};
c->V = std::array<std::array<uint16_t,8>,32>{};
c->Acc = std::array<uint64_t,8>{};
auto tmp1 = std::make_tuple(cast<uint16_t>(0ULL),cast<uint16_t>(0ULL),cast<uint8_t>(0ULL));
c->VCO = std::get<0>(tmp1);
c->VCC = std::get<1>(tmp1);
c->VCE = std::get<2>(tmp1);
auto tmp2 = std::make_tuple(cast<uint16_t>(0ULL),cast<uint16_t>(0ULL),false);
c->divIn = std::get<0>(tmp2);
c->divOut = std::get<1>(tmp2);
c->divInLoaded = std::get<2>(tmp2);
auto tmp3 = std::make_tuple(cast<uint32_t>(0ULL),cast<uint32_t>(4ULL));
c->PC = std::get<0>(tmp3);
c->nextPC = std::get<1>(tmp3);
auto tmp4 = std::make_tuple(true,false,"");
c->Halted = std::get<0>(tmp4);
c->Broke = std::get<1>(tmp4);
c->HaltReason = std::get<2>(tmp4);
c->Steps = cast<uint64_t>(0ULL);
}
}
// tools/cpu/rsp/cpu.go:87:1
void rsp_CPU_unimpl(rsp_CPU* c,uint32_t w){
{
if ((!c->Unimplemented)) {
c->Unimplemented = Map<uint32_t,int64_t>{};
}
c->Unimplemented[w]++;
}
}
// tools/cpu/rsp/cpu.go:101:1
void rsp_CPU_SetPC(rsp_CPU* c,uint32_t pc){
{
auto tmp5 = std::make_tuple(cast<uint32_t>((pc & cast<uint32_t>(4092ULL))),cast<uint32_t>(((cast<uint32_t>((pc + cast<uint32_t>(4ULL)))) & cast<uint32_t>(4092ULL))));
c->PC = std::get<0>(tmp5);
c->nextPC = std::get<1>(tmp5);
}
}
// tools/cpu/rsp/cpu.go:106:1
uint32_t rsp_CPU_CurPC(rsp_CPU* c){
{
return c->curPC;
}
}
// tools/cpu/rsp/cpu.go:109:1
void rsp_CPU_Start(rsp_CPU* c,uint32_t pc){
{
rsp_CPU_SetPC(c,pc);
auto tmp6 = std::make_tuple(false,false,"");
c->Halted = std::get<0>(tmp6);
c->Broke = std::get<1>(tmp6);
c->HaltReason = std::get<2>(tmp6);
}
}
// tools/cpu/rsp/cpu.go:116:1
uint32_t rsp_CPU_rd8(rsp_CPU* c,uint32_t a){
{
return cast<uint32_t>(c->DMEM[cast<uint32_t>((a & cast<uint32_t>(4095ULL)))]);
}
}
// tools/cpu/rsp/cpu.go:117:1
void rsp_CPU_wr8(rsp_CPU* c,uint32_t a,uint32_t v){
{
c->DMEM[cast<uint32_t>((a & cast<uint32_t>(4095ULL)))] = cast<uint8_t>(v);
}
}
// tools/cpu/rsp/cpu.go:119:1
uint32_t rsp_CPU_rd16(rsp_CPU* c,uint32_t a){
{
return cast<uint32_t>((shl<uint32_t>(cast<uint32_t>(c->DMEM[cast<uint32_t>((a & cast<uint32_t>(4095ULL)))]),cast<int64_t>(8ULL)) | cast<uint32_t>(c->DMEM[cast<uint32_t>(((cast<uint32_t>((a + cast<uint32_t>(1ULL)))) & cast<uint32_t>(4095ULL)))])));
}
}
// tools/cpu/rsp/cpu.go:122:1
void rsp_CPU_wr16(rsp_CPU* c,uint32_t a,uint32_t v){
{
c->DMEM[cast<uint32_t>((a & cast<uint32_t>(4095ULL)))] = cast<uint8_t>(shr<uint32_t>(v,cast<int64_t>(8ULL)));
c->DMEM[cast<uint32_t>(((cast<uint32_t>((a + cast<uint32_t>(1ULL)))) & cast<uint32_t>(4095ULL)))] = cast<uint8_t>(v);
}
}
// tools/cpu/rsp/cpu.go:126:1
uint32_t rsp_CPU_rd32(rsp_CPU* c,uint32_t a){
{
return cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((shl<uint32_t>(cast<uint32_t>(c->DMEM[cast<uint32_t>((a & cast<uint32_t>(4095ULL)))]),cast<int64_t>(24ULL)) | shl<uint32_t>(cast<uint32_t>(c->DMEM[cast<uint32_t>(((cast<uint32_t>((a + cast<uint32_t>(1ULL)))) & cast<uint32_t>(4095ULL)))]),cast<int64_t>(16ULL)))) | shl<uint32_t>(cast<uint32_t>(c->DMEM[cast<uint32_t>(((cast<uint32_t>((a + cast<uint32_t>(2ULL)))) & cast<uint32_t>(4095ULL)))]),cast<int64_t>(8ULL)))) | cast<uint32_t>(c->DMEM[cast<uint32_t>(((cast<uint32_t>((a + cast<uint32_t>(3ULL)))) & cast<uint32_t>(4095ULL)))])));
}
}
// tools/cpu/rsp/cpu.go:130:1
void rsp_CPU_wr32(rsp_CPU* c,uint32_t a,uint32_t v){
{
c->DMEM[cast<uint32_t>((a & cast<uint32_t>(4095ULL)))] = cast<uint8_t>(shr<uint32_t>(v,cast<int64_t>(24ULL)));
c->DMEM[cast<uint32_t>(((cast<uint32_t>((a + cast<uint32_t>(1ULL)))) & cast<uint32_t>(4095ULL)))] = cast<uint8_t>(shr<uint32_t>(v,cast<int64_t>(16ULL)));
c->DMEM[cast<uint32_t>(((cast<uint32_t>((a + cast<uint32_t>(2ULL)))) & cast<uint32_t>(4095ULL)))] = cast<uint8_t>(shr<uint32_t>(v,cast<int64_t>(8ULL)));
c->DMEM[cast<uint32_t>(((cast<uint32_t>((a + cast<uint32_t>(3ULL)))) & cast<uint32_t>(4095ULL)))] = cast<uint8_t>(v);
}
}
// tools/cpu/rsp/cpu.go:137:1
void rsp_CPU_set(rsp_CPU* c,uint32_t i,uint32_t v){
{
if ((i != cast<uint32_t>(0ULL))) {
c->R[i] = v;
}
}
}
// tools/cpu/rsp/cpu.go:146:1
void rsp_CPU_Step(rsp_CPU* c){
{
if (c->Halted) {
return ;
}
c->curPC = c->PC;
uint32_t w = cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((shl<uint32_t>(cast<uint32_t>(c->IMEM[c->PC]),cast<int64_t>(24ULL)) | shl<uint32_t>(cast<uint32_t>(c->IMEM[cast<uint32_t>((c->PC + cast<uint32_t>(1ULL)))]),cast<int64_t>(16ULL)))) | shl<uint32_t>(cast<uint32_t>(c->IMEM[cast<uint32_t>((c->PC + cast<uint32_t>(2ULL)))]),cast<int64_t>(8ULL)))) | cast<uint32_t>(c->IMEM[cast<uint32_t>((c->PC + cast<uint32_t>(3ULL)))])));
c->PC = c->nextPC;
c->nextPC = cast<uint32_t>(((cast<uint32_t>((c->nextPC + cast<uint32_t>(4ULL)))) & cast<uint32_t>(4092ULL)));
rsp_CPU_execute(c,w);
c->Steps++;
}
}
// tools/cpu/rsp/cpu.go:164:1
uint64_t rsp_CPU_Run(rsp_CPU* c,uint64_t maxSteps){
{
uint64_t n={};
for (;((n < maxSteps) && (!c->Halted));){
rsp_CPU_Step(c);
n++;
}
return n;
}
}
// tools/cpu/rsp/cpu.go:175:1
void rsp_CPU_doBranch(rsp_CPU* c,bool taken,uint32_t target){
{
if (taken) {
c->nextPC = cast<uint32_t>((target & cast<uint32_t>(4092ULL)));
}
}
}
// tools/cpu/rsp/cpu.go:181:1
void rsp_CPU_execute(rsp_CPU* c,uint32_t w){
{
uint32_t op = shr<uint32_t>(w,cast<int64_t>(26ULL));
uint32_t rs = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(21ULL))) & cast<uint32_t>(31ULL)));
uint32_t rt = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(31ULL)));
uint32_t imm = cast<uint32_t>((w & cast<uint32_t>(65535ULL)));
uint32_t simm = cast<uint32_t>(cast<int32_t>(cast<int16_t>(imm)));
uint32_t branchT = cast<uint32_t>(((cast<uint32_t>((cast<uint32_t>((c->curPC + cast<uint32_t>(4ULL))) + shl<uint32_t>(simm,cast<int64_t>(2ULL))))) & cast<uint32_t>(4092ULL)));
uint32_t jumpT = cast<uint32_t>(((shl<uint32_t>(cast<uint32_t>((w & cast<uint32_t>(67108863ULL))),cast<int64_t>(2ULL))) & cast<uint32_t>(4092ULL)));
{
switch(op){
case cast<uint32_t>(0ULL):{
rsp_CPU_special(c,w,rs,rt);
break;}
case cast<uint32_t>(1ULL):{
int32_t s = cast<int32_t>(c->R[rs]);
{
switch(rt){
case cast<uint32_t>(0ULL):{
rsp_CPU_doBranch(c,(s < cast<int32_t>(0ULL)),branchT);
break;}
case cast<uint32_t>(1ULL):{
rsp_CPU_doBranch(c,(s >= cast<int32_t>(0ULL)),branchT);
break;}
case cast<uint32_t>(16ULL):{
rsp_CPU_set(c,cast<uint32_t>(31ULL),cast<uint32_t>(((cast<uint32_t>((c->curPC + cast<uint32_t>(8ULL)))) & cast<uint32_t>(4092ULL))));
rsp_CPU_doBranch(c,(s < cast<int32_t>(0ULL)),branchT);
break;}
case cast<uint32_t>(17ULL):{
rsp_CPU_set(c,cast<uint32_t>(31ULL),cast<uint32_t>(((cast<uint32_t>((c->curPC + cast<uint32_t>(8ULL)))) & cast<uint32_t>(4092ULL))));
rsp_CPU_doBranch(c,(s >= cast<int32_t>(0ULL)),branchT);
break;}
default:{
rsp_CPU_unimpl(c,w);
break;}
}}
break;}
case cast<uint32_t>(2ULL):{
rsp_CPU_doBranch(c,true,jumpT);
break;}
case cast<uint32_t>(3ULL):{
rsp_CPU_set(c,cast<uint32_t>(31ULL),cast<uint32_t>(((cast<uint32_t>((c->curPC + cast<uint32_t>(8ULL)))) & cast<uint32_t>(4092ULL))));
rsp_CPU_doBranch(c,true,jumpT);
break;}
case cast<uint32_t>(4ULL):{
rsp_CPU_doBranch(c,(c->R[rs] == c->R[rt]),branchT);
break;}
case cast<uint32_t>(5ULL):{
rsp_CPU_doBranch(c,(c->R[rs] != c->R[rt]),branchT);
break;}
case cast<uint32_t>(6ULL):{
rsp_CPU_doBranch(c,(cast<int32_t>(c->R[rs]) <= cast<int32_t>(0ULL)),branchT);
break;}
case cast<uint32_t>(7ULL):{
rsp_CPU_doBranch(c,(cast<int32_t>(c->R[rs]) > cast<int32_t>(0ULL)),branchT);
break;}
case cast<uint32_t>(8ULL):case cast<uint32_t>(9ULL):{
rsp_CPU_set(c,rt,cast<uint32_t>((c->R[rs] + simm)));
break;}
case cast<uint32_t>(10ULL):{
rsp_CPU_set(c,rt,rsp_b2u((cast<int32_t>(c->R[rs]) < cast<int32_t>(simm))));
break;}
case cast<uint32_t>(11ULL):{
rsp_CPU_set(c,rt,rsp_b2u((c->R[rs] < simm)));
break;}
case cast<uint32_t>(12ULL):{
rsp_CPU_set(c,rt,cast<uint32_t>((c->R[rs] & imm)));
break;}
case cast<uint32_t>(13ULL):{
rsp_CPU_set(c,rt,cast<uint32_t>((c->R[rs] | imm)));
break;}
case cast<uint32_t>(14ULL):{
rsp_CPU_set(c,rt,cast<uint32_t>((c->R[rs] ^ imm)));
break;}
case cast<uint32_t>(15ULL):{
rsp_CPU_set(c,rt,shl<uint32_t>(imm,cast<int64_t>(16ULL)));
break;}
case cast<uint32_t>(16ULL):{
uint32_t rd = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(11ULL))) & cast<uint32_t>(31ULL)));
{
switch(rs){
case cast<uint32_t>(0ULL):{
rsp_CPU_set(c,rt,n64_Machine_ReadCop0(c->regs,rd));
break;}
case cast<uint32_t>(4ULL):{
n64_Machine_WriteCop0(c->regs,rd,c->R[rt]);
break;}
default:{
rsp_CPU_unimpl(c,w);
break;}
}}
break;}
case cast<uint32_t>(18ULL):{
rsp_CPU_cop2(c,w);
break;}
case cast<uint32_t>(32ULL):{
rsp_CPU_set(c,rt,cast<uint32_t>(cast<int32_t>(cast<int8_t>(cast<uint8_t>(rsp_CPU_rd8(c,cast<uint32_t>((c->R[rs] + simm))))))));
break;}
case cast<uint32_t>(36ULL):{
rsp_CPU_set(c,rt,rsp_CPU_rd8(c,cast<uint32_t>((c->R[rs] + simm))));
break;}
case cast<uint32_t>(33ULL):{
rsp_CPU_set(c,rt,cast<uint32_t>(cast<int32_t>(cast<int16_t>(cast<uint16_t>(rsp_CPU_rd16(c,cast<uint32_t>((c->R[rs] + simm))))))));
break;}
case cast<uint32_t>(37ULL):{
rsp_CPU_set(c,rt,rsp_CPU_rd16(c,cast<uint32_t>((c->R[rs] + simm))));
break;}
case cast<uint32_t>(35ULL):case cast<uint32_t>(39ULL):{
rsp_CPU_set(c,rt,rsp_CPU_rd32(c,cast<uint32_t>((c->R[rs] + simm))));
break;}
case cast<uint32_t>(40ULL):{
rsp_CPU_wr8(c,cast<uint32_t>((c->R[rs] + simm)),c->R[rt]);
break;}
case cast<uint32_t>(41ULL):{
rsp_CPU_wr16(c,cast<uint32_t>((c->R[rs] + simm)),c->R[rt]);
break;}
case cast<uint32_t>(43ULL):{
rsp_CPU_wr32(c,cast<uint32_t>((c->R[rs] + simm)),c->R[rt]);
break;}
case cast<uint32_t>(50ULL):{
rsp_CPU_vecLoad(c,w,rs,rt);
break;}
case cast<uint32_t>(58ULL):{
rsp_CPU_vecStore(c,w,rs,rt);
break;}
default:{
rsp_CPU_unimpl(c,w);
break;}
}}
}
}
// tools/cpu/rsp/cpu.go:279:1
void rsp_CPU_special(rsp_CPU* c,uint32_t w,uint32_t rs,uint32_t rt){
{
uint32_t rd = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(11ULL))) & cast<uint32_t>(31ULL)));
uint32_t shamt = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(6ULL))) & cast<uint32_t>(31ULL)));
{
switch(cast<uint32_t>((w & cast<uint32_t>(63ULL)))){
case cast<uint32_t>(0ULL):{
rsp_CPU_set(c,rd,shl<uint32_t>(c->R[rt],shamt));
break;}
case cast<uint32_t>(2ULL):{
rsp_CPU_set(c,rd,shr<uint32_t>(c->R[rt],shamt));
break;}
case cast<uint32_t>(3ULL):{
rsp_CPU_set(c,rd,cast<uint32_t>(shr<int32_t>(cast<int32_t>(c->R[rt]),shamt)));
break;}
case cast<uint32_t>(4ULL):{
rsp_CPU_set(c,rd,shl<uint32_t>(c->R[rt],(cast<uint32_t>((c->R[rs] & cast<uint32_t>(31ULL))))));
break;}
case cast<uint32_t>(6ULL):{
rsp_CPU_set(c,rd,shr<uint32_t>(c->R[rt],(cast<uint32_t>((c->R[rs] & cast<uint32_t>(31ULL))))));
break;}
case cast<uint32_t>(7ULL):{
rsp_CPU_set(c,rd,cast<uint32_t>(shr<int32_t>(cast<int32_t>(c->R[rt]),(cast<uint32_t>((c->R[rs] & cast<uint32_t>(31ULL)))))));
break;}
case cast<uint32_t>(8ULL):{
rsp_CPU_doBranch(c,true,c->R[rs]);
break;}
case cast<uint32_t>(9ULL):{
uint32_t target = c->R[rs];
rsp_CPU_set(c,rd,cast<uint32_t>(((cast<uint32_t>((c->curPC + cast<uint32_t>(8ULL)))) & cast<uint32_t>(4092ULL))));
rsp_CPU_doBranch(c,true,target);
break;}
case cast<uint32_t>(13ULL):{
c->Broke = true;
c->Halted = true;
c->HaltReason = "break";
break;}
case cast<uint32_t>(32ULL):case cast<uint32_t>(33ULL):{
rsp_CPU_set(c,rd,cast<uint32_t>((c->R[rs] + c->R[rt])));
break;}
case cast<uint32_t>(34ULL):case cast<uint32_t>(35ULL):{
rsp_CPU_set(c,rd,cast<uint32_t>((c->R[rs] - c->R[rt])));
break;}
case cast<uint32_t>(36ULL):{
rsp_CPU_set(c,rd,cast<uint32_t>((c->R[rs] & c->R[rt])));
break;}
case cast<uint32_t>(37ULL):{
rsp_CPU_set(c,rd,cast<uint32_t>((c->R[rs] | c->R[rt])));
break;}
case cast<uint32_t>(38ULL):{
rsp_CPU_set(c,rd,cast<uint32_t>((c->R[rs] ^ c->R[rt])));
break;}
case cast<uint32_t>(39ULL):{
rsp_CPU_set(c,rd,cast<uint32_t>(~(cast<uint32_t>((c->R[rs] | c->R[rt])))));
break;}
case cast<uint32_t>(42ULL):{
rsp_CPU_set(c,rd,rsp_b2u((cast<int32_t>(c->R[rs]) < cast<int32_t>(c->R[rt]))));
break;}
case cast<uint32_t>(43ULL):{
rsp_CPU_set(c,rd,rsp_b2u((c->R[rs] < c->R[rt])));
break;}
default:{
rsp_CPU_unimpl(c,w);
break;}
}}
}
}
// tools/cpu/rsp/cpu.go:328:1
uint32_t rsp_b2u(bool b){
{
if (b) {
return cast<uint32_t>(1ULL);
}
return cast<uint32_t>(0ULL);
}
}
// tools/cpu/rsp/observe_default.go:7:1
void rsp_CPU_observeVector(rsp_CPU* c,uint32_t unused0,bool unused1){
{
}
}
// tools/cpu/rsp/vu.go:29:1
uint32_t rsp_element(uint32_t e,uint32_t i){
{
{
if ((e < cast<uint32_t>(2ULL))){
return i;
}
else if ((e < cast<uint32_t>(4ULL))){
return cast<uint32_t>(((i & ~(cast<uint32_t>(1ULL))) | cast<uint32_t>((e & cast<uint32_t>(1ULL)))));
}
else if ((e < cast<uint32_t>(8ULL))){
return cast<uint32_t>(((i & ~(cast<uint32_t>(3ULL))) | cast<uint32_t>((e & cast<uint32_t>(3ULL)))));
}
else {
return cast<uint32_t>((e & cast<uint32_t>(7ULL)));
}
}
tmp7:;
}
}
// tools/cpu/rsp/vu.go:43:1
uint16_t rsp_vte(std::array<uint16_t,8> v,uint32_t e,uint32_t i){
{
return v[rsp_element(e,i)];
}
}
// tools/cpu/rsp/vu.go:46:1
int64_t rsp_CPU_acc(rsp_CPU* c,uint32_t i){
{
return shr<int64_t>(cast<int64_t>(shl<uint64_t>(c->Acc[i],cast<int64_t>(16ULL))),cast<int64_t>(16ULL));
}
}
// tools/cpu/rsp/vu.go:48:1
void rsp_CPU_setAcc(rsp_CPU* c,uint32_t i,int64_t v){
{
c->Acc[i] = cast<uint64_t>((cast<uint64_t>(v) & rsp_accMask));
}
}
// tools/cpu/rsp/vu.go:59:1
void rsp_CPU_setAccLo(rsp_CPU* c,uint32_t i,uint16_t v){
{
c->Acc[i] = cast<uint64_t>(((c->Acc[i] & ~(cast<uint64_t>(65535ULL))) | cast<uint64_t>(v)));
}
}
// tools/cpu/rsp/vu.go:64:1
uint16_t rsp_clampS(int64_t v){
{
if ((v > cast<int64_t>(32767ULL))) {
return cast<uint16_t>(32767ULL);
}
if ((v < cast<int64_t>(-cast<int64_t>(32768ULL)))) {
return cast<uint16_t>(32768ULL);
}
return cast<uint16_t>(cast<int16_t>(v));
}
}
// tools/cpu/rsp/vu.go:79:1
uint16_t rsp_clampU(int64_t v){
{
if ((v < cast<int64_t>(0ULL))) {
return cast<uint16_t>(0ULL);
}
if ((v > cast<int64_t>(32767ULL))) {
return cast<uint16_t>(65535ULL);
}
return cast<uint16_t>(v);
}
}
// tools/cpu/rsp/vu.go:92:1
uint16_t rsp_clampLow(int64_t acc){
{
int64_t hi = shr<int64_t>(acc,cast<int64_t>(16ULL));
if ((hi < cast<int64_t>(-cast<int64_t>(32768ULL)))) {
return cast<uint16_t>(0ULL);
}
if ((hi > cast<int64_t>(32767ULL))) {
return cast<uint16_t>(65535ULL);
}
return cast<uint16_t>(acc);
}
}
// tools/cpu/rsp/vu.go:103:1
int64_t rsp_s16(uint16_t v){
{
return cast<int64_t>(cast<int16_t>(v));
}
}
// tools/cpu/rsp/vu.go:104:1
int64_t rsp_u16(uint16_t v){
{
return cast<int64_t>(v);
}
}
// tools/cpu/rsp/vu.go:107:1
bool rsp_bit(uint16_t f,uint32_t i){
{
return (cast<uint16_t>((f & (shl<uint16_t>(cast<uint16_t>(1ULL),i)))) != cast<uint16_t>(0ULL));
}
}
// tools/cpu/rsp/vu.go:109:1
void rsp_setBit(uint16_t* f,uint32_t i,bool on){
{
if (on) {
(*f) |= shl<uint16_t>(cast<uint16_t>(1ULL),i);
}
else {
(*f) &= ~(shl<uint16_t>(cast<uint16_t>(1ULL),i));
}
}
}
// tools/cpu/rsp/vu.go:118:1
void rsp_CPU_cop2(rsp_CPU* c,uint32_t w){
{
uint32_t rs = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(21ULL))) & cast<uint32_t>(31ULL)));
uint32_t rt = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(31ULL)));
uint32_t rd = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(11ULL))) & cast<uint32_t>(31ULL)));
uint32_t shamt = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(6ULL))) & cast<uint32_t>(31ULL)));
if ((cast<uint32_t>((w & (shl<uint32_t>(cast<int64_t>(1ULL),cast<int64_t>(25ULL))))) == cast<uint32_t>(0ULL))) {
uint32_t e = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(7ULL))) & cast<uint32_t>(15ULL)));
{
switch(rs){
case cast<uint32_t>(0ULL):{
uint32_t b = cast<uint32_t>((e & cast<uint32_t>(15ULL)));
uint32_t v = cast<uint32_t>((shl<uint32_t>(cast<uint32_t>(rsp_CPU_vecByte(c,rd,b)),cast<int64_t>(8ULL)) | cast<uint32_t>(rsp_CPU_vecByte(c,rd,cast<uint32_t>(((cast<uint32_t>((b + cast<uint32_t>(1ULL)))) & cast<uint32_t>(15ULL)))))));
rsp_CPU_set(c,rt,cast<uint32_t>(cast<int32_t>(cast<int16_t>(cast<uint16_t>(v)))));
break;}
case cast<uint32_t>(4ULL):{
uint32_t b = cast<uint32_t>((e & cast<uint32_t>(15ULL)));
rsp_CPU_setVecByte(c,rd,b,cast<uint8_t>(shr<uint32_t>(c->R[rt],cast<int64_t>(8ULL))));
if ((cast<uint32_t>((b + cast<uint32_t>(1ULL))) < cast<uint32_t>(16ULL))) {
rsp_CPU_setVecByte(c,rd,cast<uint32_t>((b + cast<uint32_t>(1ULL))),cast<uint8_t>(c->R[rt]));
}
break;}
case cast<uint32_t>(2ULL):{
rsp_CPU_set(c,rt,cast<uint32_t>(cast<int32_t>(cast<int16_t>(rsp_CPU_ctrl(c,cast<uint32_t>((rd & cast<uint32_t>(3ULL))))))));
break;}
case cast<uint32_t>(6ULL):{
rsp_CPU_setCtrl(c,cast<uint32_t>((rd & cast<uint32_t>(3ULL))),cast<uint16_t>(c->R[rt]));
break;}
default:{
rsp_CPU_unimpl(c,w);
break;}
}}
return ;
}
rsp_CPU_vectorALU(c,w,cast<uint32_t>((w & cast<uint32_t>(63ULL))),cast<uint32_t>((rs & cast<uint32_t>(15ULL))),rt,rd,shamt);
}
}
// tools/cpu/rsp/vu.go:152:1
uint16_t rsp_CPU_ctrl(rsp_CPU* c,uint32_t i){
{
{
switch(cast<uint32_t>((i & cast<uint32_t>(3ULL)))){
case cast<uint32_t>(0ULL):{
return c->VCO;
break;}
case cast<uint32_t>(1ULL):{
return c->VCC;
break;}
default:{
return cast<uint16_t>(c->VCE);
break;}
}}
}
}
// tools/cpu/rsp/vu.go:163:1
void rsp_CPU_setCtrl(rsp_CPU* c,uint32_t i,uint16_t v){
{
{
switch(cast<uint32_t>((i & cast<uint32_t>(3ULL)))){
case cast<uint32_t>(0ULL):{
c->VCO = v;
break;}
case cast<uint32_t>(1ULL):{
c->VCC = v;
break;}
default:{
c->VCE = cast<uint8_t>(v);
break;}
}}
}
}
// tools/cpu/rsp/vu.go:176:1
uint8_t rsp_CPU_vecByte(rsp_CPU* c,uint32_t r,uint32_t b){
{
uint32_t lane = cast<uint32_t>(((shr<uint32_t>(b,cast<int64_t>(1ULL))) & cast<uint32_t>(7ULL)));
if ((cast<uint32_t>((b & cast<uint32_t>(1ULL))) == cast<uint32_t>(0ULL))) {
return cast<uint8_t>(shr<uint16_t>(c->V[r][lane],cast<int64_t>(8ULL)));
}
return cast<uint8_t>(c->V[r][lane]);
}
}
// tools/cpu/rsp/vu.go:184:1
void rsp_CPU_setVecByte(rsp_CPU* c,uint32_t r,uint32_t b,uint8_t v){
{
uint32_t lane = cast<uint32_t>(((shr<uint32_t>(b,cast<int64_t>(1ULL))) & cast<uint32_t>(7ULL)));
if ((cast<uint32_t>((b & cast<uint32_t>(1ULL))) == cast<uint32_t>(0ULL))) {
c->V[r][lane] = cast<uint16_t>((cast<uint16_t>((c->V[r][lane] & cast<uint16_t>(255ULL))) | shl<uint16_t>(cast<uint16_t>(v),cast<int64_t>(8ULL))));
}
else {
c->V[r][lane] = cast<uint16_t>((cast<uint16_t>((c->V[r][lane] & cast<uint16_t>(65280ULL))) | cast<uint16_t>(v)));
}
}
}
// tools/cpu/rsp/vu.go:201:1
void rsp_CPU_vectorALU(rsp_CPU* c,uint32_t w,uint32_t funct,uint32_t e,uint32_t vt,uint32_t vs,uint32_t vd){
{
rsp_CPU_observeVector(c,w,false);
auto tmp8 = std::make_tuple(c->V[vs],c->V[vt]);
std::array<uint16_t,8> vsv = std::get<0>(tmp8);
std::array<uint16_t,8> vtv = std::get<1>(tmp8);
{
switch(funct){
case cast<uint32_t>(0ULL):{
for (uint32_t i = cast<uint32_t>(cast<uint32_t>(0ULL));(i < cast<uint32_t>(8ULL));i++){
int64_t p = cast<int64_t>((rsp_s16(vsv[i]) * rsp_s16(rsp_vte(vtv,e,i))));
rsp_CPU_setAcc(c,i,cast<int64_t>((shl<int64_t>(p,cast<int64_t>(1ULL)) + cast<int64_t>(32768ULL))));
c->V[vd][i] = rsp_clampS(shr<int64_t>(rsp_CPU_acc(c,i),cast<int64_t>(16ULL)));
}
break;}
case cast<uint32_t>(1ULL):{
for (uint32_t i = cast<uint32_t>(cast<uint32_t>(0ULL));(i < cast<uint32_t>(8ULL));i++){
int64_t p = cast<int64_t>((rsp_s16(vsv[i]) * rsp_s16(rsp_vte(vtv,e,i))));
rsp_CPU_setAcc(c,i,cast<int64_t>((shl<int64_t>(p,cast<int64_t>(1ULL)) + cast<int64_t>(32768ULL))));
c->V[vd][i] = rsp_clampU(shr<int64_t>(rsp_CPU_acc(c,i),cast<int64_t>(16ULL)));
}
break;}
case cast<uint32_t>(4ULL):{
for (uint32_t i = cast<uint32_t>(cast<uint32_t>(0ULL));(i < cast<uint32_t>(8ULL));i++){
rsp_CPU_setAcc(c,i,shr<int64_t>(cast<int64_t>((rsp_u16(vsv[i]) * rsp_u16(rsp_vte(vtv,e,i)))),cast<int64_t>(16ULL)));
c->V[vd][i] = cast<uint16_t>(rsp_CPU_acc(c,i));
}
break;}
case cast<uint32_t>(5ULL):{
for (uint32_t i = cast<uint32_t>(cast<uint32_t>(0ULL));(i < cast<uint32_t>(8ULL));i++){
rsp_CPU_setAcc(c,i,cast<int64_t>((rsp_s16(vsv[i]) * rsp_u16(rsp_vte(vtv,e,i)))));
c->V[vd][i] = rsp_clampS(shr<int64_t>(rsp_CPU_acc(c,i),cast<int64_t>(16ULL)));
}
break;}
case cast<uint32_t>(6ULL):{
for (uint32_t i = cast<uint32_t>(cast<uint32_t>(0ULL));(i < cast<uint32_t>(8ULL));i++){
rsp_CPU_setAcc(c,i,cast<int64_t>((rsp_u16(vsv[i]) * rsp_s16(rsp_vte(vtv,e,i)))));
c->V[vd][i] = cast<uint16_t>(rsp_CPU_acc(c,i));
}
break;}
case cast<uint32_t>(7ULL):{
for (uint32_t i = cast<uint32_t>(cast<uint32_t>(0ULL));(i < cast<uint32_t>(8ULL));i++){
int64_t p = cast<int64_t>((rsp_s16(vsv[i]) * rsp_s16(rsp_vte(vtv,e,i))));
rsp_CPU_setAcc(c,i,shl<int64_t>(p,cast<int64_t>(16ULL)));
c->V[vd][i] = rsp_clampS(p);
}
break;}
case cast<uint32_t>(3ULL):{
for (uint32_t i = cast<uint32_t>(cast<uint32_t>(0ULL));(i < cast<uint32_t>(8ULL));i++){
int64_t a = shl<int64_t>(cast<int64_t>((rsp_s16(vsv[i]) * rsp_s16(rsp_vte(vtv,e,i)))),cast<int64_t>(16ULL));
if ((a < cast<int64_t>(0ULL))) {
a += cast<int64_t>(2031616ULL);
}
rsp_CPU_setAcc(c,i,a);
c->V[vd][i] = cast<uint16_t>((rsp_clampS(shr<int64_t>(a,cast<int64_t>(17ULL))) & cast<uint16_t>(65520ULL)));
}
break;}
case cast<uint32_t>(2ULL):case cast<uint32_t>(10ULL):{
for (uint32_t i = cast<uint32_t>(cast<uint32_t>(0ULL));(i < cast<uint32_t>(8ULL));i++){
int64_t p = rsp_s16(rsp_vte(vtv,e,i));
if ((cast<uint32_t>((vs & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL))) {
p = shl<int64_t>(p,cast<int64_t>(16ULL));
}
int64_t a = rsp_CPU_acc(c,i);
if ((((funct == cast<uint32_t>(2ULL))) == ((a >= cast<int64_t>(0ULL))))) {
a += p;
}
rsp_CPU_setAcc(c,i,a);
c->V[vd][i] = rsp_clampS(shr<int64_t>(rsp_CPU_acc(c,i),cast<int64_t>(16ULL)));
}
break;}
case cast<uint32_t>(11ULL):{
for (uint32_t i = cast<uint32_t>(cast<uint32_t>(0ULL));(i < cast<uint32_t>(8ULL));i++){
int64_t a = rsp_CPU_acc(c,i);
if ((cast<int64_t>((a & cast<int64_t>(2097152ULL))) == cast<int64_t>(0ULL))) {
if ((shr<int64_t>(a,cast<int64_t>(22ULL)) < cast<int64_t>(0ULL))) {
a += cast<int64_t>(2097152ULL);
}
else if ((shr<int64_t>(a,cast<int64_t>(22ULL)) > cast<int64_t>(0ULL))) {
a -= cast<int64_t>(2097152ULL);
}
}
rsp_CPU_setAcc(c,i,a);
uint16_t r = cast<uint16_t>(shr<int64_t>(a,cast<int64_t>(17ULL)));
if (((a < cast<int64_t>(0ULL)) && (shr<int64_t>((cast<int64_t>(~a)),cast<int64_t>(32ULL)) != cast<int64_t>(0ULL)))) {
r = cast<uint16_t>(32768ULL);
}
else if (((a >= cast<int64_t>(0ULL)) && (shr<int64_t>(a,cast<int64_t>(32ULL)) != cast<int64_t>(0ULL)))) {
r = cast<uint16_t>(32767ULL);
}
c->V[vd][i] = cast<uint16_t>((r & cast<uint16_t>(65520ULL)));
}
break;}
case cast<uint32_t>(8ULL):{
for (uint32_t i = cast<uint32_t>(cast<uint32_t>(0ULL));(i < cast<uint32_t>(8ULL));i++){
int64_t p = cast<int64_t>((rsp_s16(vsv[i]) * rsp_s16(rsp_vte(vtv,e,i))));
rsp_CPU_setAcc(c,i,cast<int64_t>((rsp_CPU_acc(c,i) + shl<int64_t>(p,cast<int64_t>(1ULL)))));
c->V[vd][i] = rsp_clampS(shr<int64_t>(rsp_CPU_acc(c,i),cast<int64_t>(16ULL)));
}
break;}
case cast<uint32_t>(9ULL):{
for (uint32_t i = cast<uint32_t>(cast<uint32_t>(0ULL));(i < cast<uint32_t>(8ULL));i++){
int64_t p = cast<int64_t>((rsp_s16(vsv[i]) * rsp_s16(rsp_vte(vtv,e,i))));
rsp_CPU_setAcc(c,i,cast<int64_t>((rsp_CPU_acc(c,i) + shl<int64_t>(p,cast<int64_t>(1ULL)))));
c->V[vd][i] = rsp_clampU(shr<int64_t>(rsp_CPU_acc(c,i),cast<int64_t>(16ULL)));
}
break;}
case cast<uint32_t>(12ULL):{
for (uint32_t i = cast<uint32_t>(cast<uint32_t>(0ULL));(i < cast<uint32_t>(8ULL));i++){
rsp_CPU_setAcc(c,i,cast<int64_t>((rsp_CPU_acc(c,i) + shr<int64_t>(cast<int64_t>((rsp_u16(vsv[i]) * rsp_u16(rsp_vte(vtv,e,i)))),cast<int64_t>(16ULL)))));
c->V[vd][i] = rsp_clampLow(rsp_CPU_acc(c,i));
}
break;}
case cast<uint32_t>(13ULL):{
for (uint32_t i = cast<uint32_t>(cast<uint32_t>(0ULL));(i < cast<uint32_t>(8ULL));i++){
rsp_CPU_setAcc(c,i,cast<int64_t>((rsp_CPU_acc(c,i) + cast<int64_t>((rsp_s16(vsv[i]) * rsp_u16(rsp_vte(vtv,e,i)))))));
c->V[vd][i] = rsp_clampS(shr<int64_t>(rsp_CPU_acc(c,i),cast<int64_t>(16ULL)));
}
break;}
case cast<uint32_t>(14ULL):{
for (uint32_t i = cast<uint32_t>(cast<uint32_t>(0ULL));(i < cast<uint32_t>(8ULL));i++){
rsp_CPU_setAcc(c,i,cast<int64_t>((rsp_CPU_acc(c,i) + cast<int64_t>((rsp_u16(vsv[i]) * rsp_s16(rsp_vte(vtv,e,i)))))));
c->V[vd][i] = rsp_clampLow(rsp_CPU_acc(c,i));
}
break;}
case cast<uint32_t>(15ULL):{
for (uint32_t i = cast<uint32_t>(cast<uint32_t>(0ULL));(i < cast<uint32_t>(8ULL));i++){
int64_t p = cast<int64_t>((rsp_s16(vsv[i]) * rsp_s16(rsp_vte(vtv,e,i))));
rsp_CPU_setAcc(c,i,cast<int64_t>((rsp_CPU_acc(c,i) + shl<int64_t>(p,cast<int64_t>(16ULL)))));
c->V[vd][i] = rsp_clampS(shr<int64_t>(rsp_CPU_acc(c,i),cast<int64_t>(16ULL)));
}
break;}
case cast<uint32_t>(16ULL):{
for (uint32_t i = cast<uint32_t>(cast<uint32_t>(0ULL));(i < cast<uint32_t>(8ULL));i++){
int64_t sum = cast<int64_t>((cast<int64_t>((rsp_s16(vsv[i]) + rsp_s16(rsp_vte(vtv,e,i)))) + rsp_b2i(rsp_bit(c->VCO,i))));
rsp_CPU_setAccLo(c,i,cast<uint16_t>(sum));
c->V[vd][i] = rsp_clampS(sum);
}
c->VCO = cast<uint16_t>(0ULL);
break;}
case cast<uint32_t>(17ULL):{
for (uint32_t i = cast<uint32_t>(cast<uint32_t>(0ULL));(i < cast<uint32_t>(8ULL));i++){
int64_t d = cast<int64_t>((cast<int64_t>((rsp_s16(vsv[i]) - rsp_s16(rsp_vte(vtv,e,i)))) - rsp_b2i(rsp_bit(c->VCO,i))));
rsp_CPU_setAccLo(c,i,cast<uint16_t>(d));
c->V[vd][i] = rsp_clampS(d);
}
c->VCO = cast<uint16_t>(0ULL);
break;}
case cast<uint32_t>(19ULL):{
for (uint32_t i = cast<uint32_t>(cast<uint32_t>(0ULL));(i < cast<uint32_t>(8ULL));i++){
int64_t s = rsp_s16(vsv[i]);
uint16_t t = rsp_vte(vtv,e,i);
uint16_t r={};
uint16_t a={};
{
if ((s < cast<int64_t>(0ULL))){
a = cast<uint16_t>(cast<int16_t>(-cast<int16_t>(t)));
r = a;
if ((t == cast<uint16_t>(32768ULL))) {
r = cast<uint16_t>(32767ULL);
}
}
else if ((s > cast<int64_t>(0ULL))){
auto tmp10 = std::make_tuple(t,t);
r = std::get<0>(tmp10);
a = std::get<1>(tmp10);
}
}
tmp9:;
rsp_CPU_setAccLo(c,i,a);
c->V[vd][i] = r;
}
break;}
case cast<uint32_t>(20ULL):{
c->VCO = cast<uint16_t>(0ULL);
for (uint32_t i = cast<uint32_t>(cast<uint32_t>(0ULL));(i < cast<uint32_t>(8ULL));i++){
int64_t sum = cast<int64_t>((rsp_u16(vsv[i]) + rsp_u16(rsp_vte(vtv,e,i))));
rsp_CPU_setAccLo(c,i,cast<uint16_t>(sum));
c->V[vd][i] = cast<uint16_t>(sum);
rsp_setBit((&c->VCO),i,(sum > cast<int64_t>(65535ULL)));
}
break;}
case cast<uint32_t>(21ULL):{
c->VCO = cast<uint16_t>(0ULL);
for (uint32_t i = cast<uint32_t>(cast<uint32_t>(0ULL));(i < cast<uint32_t>(8ULL));i++){
int64_t d = cast<int64_t>((rsp_u16(vsv[i]) - rsp_u16(rsp_vte(vtv,e,i))));
rsp_CPU_setAccLo(c,i,cast<uint16_t>(d));
c->V[vd][i] = cast<uint16_t>(d);
rsp_setBit((&c->VCO),i,(d < cast<int64_t>(0ULL)));
rsp_setBit((&c->VCO),cast<uint32_t>((i + cast<uint32_t>(8ULL))),(d != cast<int64_t>(0ULL)));
}
break;}
case cast<uint32_t>(29ULL):{
for (uint32_t i = cast<uint32_t>(cast<uint32_t>(0ULL));(i < cast<uint32_t>(8ULL));i++){
{
switch(e){
case cast<uint32_t>(8ULL):{
c->V[vd][i] = cast<uint16_t>(shr<uint64_t>(c->Acc[i],cast<int64_t>(32ULL)));
break;}
case cast<uint32_t>(9ULL):{
c->V[vd][i] = cast<uint16_t>(shr<uint64_t>(c->Acc[i],cast<int64_t>(16ULL)));
break;}
case cast<uint32_t>(10ULL):{
c->V[vd][i] = cast<uint16_t>(c->Acc[i]);
break;}
default:{
c->V[vd][i] = cast<uint16_t>(0ULL);
break;}
}}
}
break;}
case cast<uint32_t>(32ULL):case cast<uint32_t>(33ULL):case cast<uint32_t>(34ULL):case cast<uint32_t>(35ULL):{
rsp_CPU_compare(c,funct,e,vsv,vtv,vd);
break;}
case cast<uint32_t>(36ULL):{
rsp_CPU_vcl(c,e,vsv,vtv,vd);
break;}
case cast<uint32_t>(37ULL):{
rsp_CPU_vch(c,e,vsv,vtv,vd);
break;}
case cast<uint32_t>(38ULL):{
rsp_CPU_vcr(c,e,vsv,vtv,vd);
break;}
case cast<uint32_t>(39ULL):{
for (uint32_t i = cast<uint32_t>(cast<uint32_t>(0ULL));(i < cast<uint32_t>(8ULL));i++){
uint16_t r = rsp_vte(vtv,e,i);
if (rsp_bit(c->VCC,i)) {
r = vsv[i];
}
rsp_CPU_setAccLo(c,i,r);
c->V[vd][i] = r;
}
c->VCO = cast<uint16_t>(0ULL);
break;}
case cast<uint32_t>(40ULL):case cast<uint32_t>(41ULL):case cast<uint32_t>(42ULL):case cast<uint32_t>(43ULL):case cast<uint32_t>(44ULL):case cast<uint32_t>(45ULL):{
for (uint32_t i = cast<uint32_t>(cast<uint32_t>(0ULL));(i < cast<uint32_t>(8ULL));i++){
auto tmp11 = std::make_tuple(vsv[i],rsp_vte(vtv,e,i));
uint16_t a = std::get<0>(tmp11);
uint16_t b = std::get<1>(tmp11);
uint16_t r={};
{
switch(funct){
case cast<uint32_t>(40ULL):{
r = cast<uint16_t>((a & b));
break;}
case cast<uint32_t>(41ULL):{
r = cast<uint16_t>(~(cast<uint16_t>((a & b))));
break;}
case cast<uint32_t>(42ULL):{
r = cast<uint16_t>((a | b));
break;}
case cast<uint32_t>(43ULL):{
r = cast<uint16_t>(~(cast<uint16_t>((a | b))));
break;}
case cast<uint32_t>(44ULL):{
r = cast<uint16_t>((a ^ b));
break;}
case cast<uint32_t>(45ULL):{
r = cast<uint16_t>(~(cast<uint16_t>((a ^ b))));
break;}
}}
rsp_CPU_setAccLo(c,i,r);
c->V[vd][i] = r;
}
break;}
case cast<uint32_t>(51ULL):{
uint32_t de = cast<uint32_t>((vs & cast<uint32_t>(7ULL)));
c->V[vd][de] = rsp_vte(vtv,e,de);
for (uint32_t i = cast<uint32_t>(cast<uint32_t>(0ULL));(i < cast<uint32_t>(8ULL));i++){
rsp_CPU_setAccLo(c,i,rsp_vte(vtv,e,i));
}
break;}
case cast<uint32_t>(55ULL):{
break;}
case cast<uint32_t>(18ULL):case cast<uint32_t>(22ULL):case cast<uint32_t>(23ULL):case cast<uint32_t>(24ULL):case cast<uint32_t>(25ULL):case cast<uint32_t>(26ULL):case cast<uint32_t>(27ULL):case cast<uint32_t>(28ULL):case cast<uint32_t>(30ULL):case cast<uint32_t>(31ULL):case cast<uint32_t>(46ULL):case cast<uint32_t>(47ULL):case cast<uint32_t>(56ULL):case cast<uint32_t>(57ULL):case cast<uint32_t>(58ULL):case cast<uint32_t>(59ULL):case cast<uint32_t>(60ULL):case cast<uint32_t>(61ULL):case cast<uint32_t>(62ULL):{
for (uint32_t i = cast<uint32_t>(cast<uint32_t>(0ULL));(i < cast<uint32_t>(8ULL));i++){
rsp_CPU_setAccLo(c,i,cast<uint16_t>((vsv[i] + rsp_vte(vtv,e,i))));
c->V[vd][i] = cast<uint16_t>(0ULL);
}
break;}
case cast<uint32_t>(63ULL):{
break;}
case cast<uint32_t>(48ULL):{
auto tmp12 = std::make_tuple(cast<uint16_t>(0ULL),false);
c->divIn = std::get<0>(tmp12);
c->divInLoaded = std::get<1>(tmp12);
rsp_CPU_divide(c,e,vtv,vs,vd,rsp_CPU_reciprocal(c,cast<int32_t>(cast<int16_t>(vtv[cast<uint32_t>((e & cast<uint32_t>(7ULL)))]))));
break;}
case cast<uint32_t>(49ULL):{
rsp_CPU_divide(c,e,vtv,vs,vd,rsp_CPU_reciprocal(c,rsp_CPU_divInput(c,vtv[cast<uint32_t>((e & cast<uint32_t>(7ULL)))])));
break;}
case cast<uint32_t>(52ULL):{
auto tmp13 = std::make_tuple(cast<uint16_t>(0ULL),false);
c->divIn = std::get<0>(tmp13);
c->divInLoaded = std::get<1>(tmp13);
rsp_CPU_divide(c,e,vtv,vs,vd,rsp_CPU_rsqrt(c,cast<int32_t>(cast<int16_t>(vtv[cast<uint32_t>((e & cast<uint32_t>(7ULL)))]))));
break;}
case cast<uint32_t>(53ULL):{
rsp_CPU_divide(c,e,vtv,vs,vd,rsp_CPU_rsqrt(c,rsp_CPU_divInput(c,vtv[cast<uint32_t>((e & cast<uint32_t>(7ULL)))])));
break;}
case cast<uint32_t>(50ULL):case cast<uint32_t>(54ULL):{
c->V[vd][cast<uint32_t>((vs & cast<uint32_t>(7ULL)))] = c->divOut;
auto tmp14 = std::make_tuple(vtv[cast<uint32_t>((e & cast<uint32_t>(7ULL)))],true);
c->divIn = std::get<0>(tmp14);
c->divInLoaded = std::get<1>(tmp14);
for (uint32_t i = cast<uint32_t>(cast<uint32_t>(0ULL));(i < cast<uint32_t>(8ULL));i++){
rsp_CPU_setAccLo(c,i,rsp_vte(vtv,e,i));
}
break;}
default:{
rsp_CPU_unimpl(c,w);
break;}
}}
rsp_CPU_observeVector(c,w,true);
}
}
// tools/cpu/rsp/vu.go:485:1
int64_t rsp_b2i(bool b){
{
if (b) {
return cast<int64_t>(1ULL);
}
return cast<int64_t>(0ULL);
}
}
// tools/cpu/rsp/vu.go:494:1
void rsp_CPU_compare(rsp_CPU* c,uint32_t funct,uint32_t e,std::array<uint16_t,8> vsv,std::array<uint16_t,8> vtv,uint32_t vd){
{
c->VCC = cast<uint16_t>(0ULL);
for (uint32_t i = cast<uint32_t>(cast<uint32_t>(0ULL));(i < cast<uint32_t>(8ULL));i++){
auto tmp15 = std::make_tuple(rsp_s16(vsv[i]),rsp_s16(rsp_vte(vtv,e,i)));
int64_t a = std::get<0>(tmp15);
int64_t b = std::get<1>(tmp15);
auto tmp16 = std::make_tuple(rsp_bit(c->VCO,i),rsp_bit(c->VCO,cast<uint32_t>((i + cast<uint32_t>(8ULL)))));
bool carry = std::get<0>(tmp16);
bool ne = std::get<1>(tmp16);
bool cond={};
{
switch(funct){
case cast<uint32_t>(32ULL):{
cond = ((a < b) || ((((a == b) && carry) && ne)));
break;}
case cast<uint32_t>(33ULL):{
cond = ((a == b) && (!ne));
break;}
case cast<uint32_t>(34ULL):{
cond = ((a != b) || ne);
break;}
case cast<uint32_t>(35ULL):{
cond = ((a > b) || (((a == b) && (!((carry && ne))))));
break;}
}}
rsp_setBit((&c->VCC),i,cond);
uint16_t r = cast<uint16_t>(b);
if (cond) {
r = cast<uint16_t>(a);
}
rsp_CPU_setAccLo(c,i,r);
c->V[vd][i] = r;
}
c->VCO = cast<uint16_t>(0ULL);
}
}
// tools/cpu/rsp/vu.go:531:1
void rsp_CPU_vch(rsp_CPU* c,uint32_t e,std::array<uint16_t,8> vsv,std::array<uint16_t,8> vtv,uint32_t vd){
{
auto tmp17 = std::make_tuple(cast<uint16_t>(0ULL),cast<uint16_t>(0ULL),cast<uint8_t>(0ULL));
c->VCO = std::get<0>(tmp17);
c->VCC = std::get<1>(tmp17);
c->VCE = std::get<2>(tmp17);
for (uint32_t i = cast<uint32_t>(cast<uint32_t>(0ULL));(i < cast<uint32_t>(8ULL));i++){
auto tmp18 = std::make_tuple(rsp_s16(vsv[i]),rsp_s16(rsp_vte(vtv,e,i)));
int64_t a = std::get<0>(tmp18);
int64_t b = std::get<1>(tmp18);
int64_t r={};
if (((cast<int64_t>((a ^ b))) < cast<int64_t>(0ULL))) {
int64_t sum = cast<int64_t>((a + b));
bool le = (sum <= cast<int64_t>(0ULL));
rsp_setBit((&c->VCC),i,le);
rsp_setBit((&c->VCC),cast<uint32_t>((i + cast<uint32_t>(8ULL))),(b < cast<int64_t>(0ULL)));
if (le) {
r = cast<int64_t>(-b);
}
else {
r = a;
}
rsp_setBit((&c->VCO),i,true);
rsp_setBit((&c->VCO),cast<uint32_t>((i + cast<uint32_t>(8ULL))),((sum != cast<int64_t>(0ULL)) && (cast<uint16_t>(a) != cast<uint16_t>(~cast<uint16_t>(b)))));
if ((sum == cast<int64_t>(-cast<int64_t>(1ULL)))) {
c->VCE |= shl<uint8_t>(cast<uint8_t>(1ULL),i);
}
}
else {
bool ge = (cast<int64_t>((a - b)) >= cast<int64_t>(0ULL));
rsp_setBit((&c->VCC),i,(b < cast<int64_t>(0ULL)));
rsp_setBit((&c->VCC),cast<uint32_t>((i + cast<uint32_t>(8ULL))),ge);
if (ge) {
r = b;
}
else {
r = a;
}
rsp_setBit((&c->VCO),cast<uint32_t>((i + cast<uint32_t>(8ULL))),(cast<int64_t>((a - b)) != cast<int64_t>(0ULL)));
}
rsp_CPU_setAccLo(c,i,cast<uint16_t>(r));
c->V[vd][i] = cast<uint16_t>(r);
}
}
}
// tools/cpu/rsp/vu.go:572:1
void rsp_CPU_vcl(rsp_CPU* c,uint32_t e,std::array<uint16_t,8> vsv,std::array<uint16_t,8> vtv,uint32_t vd){
{
for (uint32_t i = cast<uint32_t>(cast<uint32_t>(0ULL));(i < cast<uint32_t>(8ULL));i++){
auto tmp19 = std::make_tuple(vsv[i],rsp_vte(vtv,e,i));
uint16_t a = std::get<0>(tmp19);
uint16_t b = std::get<1>(tmp19);
auto tmp20 = std::make_tuple(rsp_bit(c->VCO,i),rsp_bit(c->VCO,cast<uint32_t>((i + cast<uint32_t>(8ULL)))));
bool carry = std::get<0>(tmp20);
bool ne = std::get<1>(tmp20);
bool vce = (cast<uint8_t>((c->VCE & (shl<uint8_t>(cast<uint8_t>(1ULL),i)))) != cast<uint8_t>(0ULL));
uint16_t r={};
if (carry) {
if ((!ne)) {
uint32_t sum = cast<uint32_t>((cast<uint32_t>(a) + cast<uint32_t>(b)));
bool exact = (cast<uint32_t>((sum & cast<uint32_t>(65535ULL))) == cast<uint32_t>(0ULL));
bool noCarry = (cast<uint32_t>((sum & cast<uint32_t>(65536ULL))) == cast<uint32_t>(0ULL));
bool le = (exact && noCarry);
if (vce) {
le = (exact || noCarry);
}
rsp_setBit((&c->VCC),i,le);
}
if (rsp_bit(c->VCC,i)) {
r = cast<uint16_t>(cast<int16_t>(-cast<int16_t>(b)));
}
else {
r = a;
}
}
else {
if ((!ne)) {
rsp_setBit((&c->VCC),cast<uint32_t>((i + cast<uint32_t>(8ULL))),(a >= b));
}
if (rsp_bit(c->VCC,cast<uint32_t>((i + cast<uint32_t>(8ULL))))) {
r = b;
}
else {
r = a;
}
}
rsp_CPU_setAccLo(c,i,r);
c->V[vd][i] = r;
}
auto tmp21 = std::make_tuple(cast<uint16_t>(0ULL),cast<uint8_t>(0ULL));
c->VCO = std::get<0>(tmp21);
c->VCE = std::get<1>(tmp21);
}
}
// tools/cpu/rsp/vu.go:614:1
void rsp_CPU_vcr(rsp_CPU* c,uint32_t e,std::array<uint16_t,8> vsv,std::array<uint16_t,8> vtv,uint32_t vd){
{
auto tmp22 = std::make_tuple(cast<uint16_t>(0ULL),cast<uint16_t>(0ULL),cast<uint8_t>(0ULL));
c->VCO = std::get<0>(tmp22);
c->VCC = std::get<1>(tmp22);
c->VCE = std::get<2>(tmp22);
for (uint32_t i = cast<uint32_t>(cast<uint32_t>(0ULL));(i < cast<uint32_t>(8ULL));i++){
auto tmp23 = std::make_tuple(rsp_s16(vsv[i]),rsp_s16(rsp_vte(vtv,e,i)));
int64_t a = std::get<0>(tmp23);
int64_t b = std::get<1>(tmp23);
int64_t r={};
if (((cast<int64_t>((a ^ b))) < cast<int64_t>(0ULL))) {
bool le = (cast<int64_t>((a + b)) < cast<int64_t>(0ULL));
rsp_setBit((&c->VCC),i,le);
rsp_setBit((&c->VCC),cast<uint32_t>((i + cast<uint32_t>(8ULL))),(b < cast<int64_t>(0ULL)));
if (le) {
r = cast<int64_t>(~b);
}
else {
r = a;
}
}
else {
bool ge = (cast<int64_t>((a - b)) >= cast<int64_t>(0ULL));
rsp_setBit((&c->VCC),i,(b < cast<int64_t>(0ULL)));
rsp_setBit((&c->VCC),cast<uint32_t>((i + cast<uint32_t>(8ULL))),ge);
if (ge) {
r = b;
}
else {
r = a;
}
}
rsp_CPU_setAccLo(c,i,cast<uint16_t>(r));
c->V[vd][i] = cast<uint16_t>(r);
}
}
}
// tools/cpu/rsp/vu.go:649:1
void rsp_CPU_vecLoad(rsp_CPU* c,uint32_t w,uint32_t rs,uint32_t vt){
{
uint32_t op = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(11ULL))) & cast<uint32_t>(31ULL)));
uint32_t e = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(7ULL))) & cast<uint32_t>(15ULL)));
int32_t off = shr<int32_t>(shl<int32_t>(cast<int32_t>(cast<uint32_t>((w & cast<uint32_t>(127ULL)))),cast<int64_t>(25ULL)),cast<int64_t>(25ULL));
uint32_t addr = cast<uint32_t>((c->R[rs] + cast<uint32_t>(shl<int32_t>(off,rsp_memShift[op]))));
{
switch(op){
case cast<uint32_t>(0ULL):case cast<uint32_t>(1ULL):case cast<uint32_t>(2ULL):case cast<uint32_t>(3ULL):{
uint32_t n = shl<uint32_t>(cast<uint32_t>(cast<uint32_t>(1ULL)),op);
for (uint32_t i = cast<uint32_t>(cast<uint32_t>(0ULL));((i < n) && (cast<uint32_t>((e + i)) < cast<uint32_t>(16ULL)));i++){
rsp_CPU_setVecByte(c,vt,cast<uint32_t>((e + i)),c->DMEM[cast<uint32_t>(((cast<uint32_t>((addr + i))) & cast<uint32_t>(4095ULL)))]);
}
break;}
case cast<uint32_t>(4ULL):{
uint32_t end = cast<uint32_t>((((addr & ~(cast<uint32_t>(15ULL)))) + cast<uint32_t>(16ULL)));
for (uint32_t b = e;((b < cast<uint32_t>(16ULL)) && (addr < end));[&](){auto tmp24 = std::make_tuple(cast<uint32_t>((b + cast<uint32_t>(1ULL))),cast<uint32_t>((addr + cast<uint32_t>(1ULL))));
b = std::get<0>(tmp24);
addr = std::get<1>(tmp24);}()){
rsp_CPU_setVecByte(c,vt,b,c->DMEM[cast<uint32_t>((addr & cast<uint32_t>(4095ULL)))]);
}
break;}
case cast<uint32_t>(5ULL):{
uint32_t n = cast<uint32_t>((addr & cast<uint32_t>(15ULL)));
uint32_t base = (addr & ~(cast<uint32_t>(15ULL)));
for (uint32_t i = cast<uint32_t>(cast<uint32_t>(0ULL));(cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(16ULL) - n)) + e)) + i)) < cast<uint32_t>(16ULL));i++){
rsp_CPU_setVecByte(c,vt,cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(16ULL) - n)) + e)) + i)),c->DMEM[cast<uint32_t>(((cast<uint32_t>((base + i))) & cast<uint32_t>(4095ULL)))]);
}
break;}
case cast<uint32_t>(6ULL):{
rsp_CPU_packedLoad(c,vt,e,addr,cast<uint32_t>(8ULL),cast<uint32_t>(1ULL));
break;}
case cast<uint32_t>(7ULL):{
rsp_CPU_packedLoad(c,vt,e,addr,cast<uint32_t>(7ULL),cast<uint32_t>(1ULL));
break;}
case cast<uint32_t>(8ULL):{
rsp_CPU_packedLoad(c,vt,e,addr,cast<uint32_t>(7ULL),cast<uint32_t>(2ULL));
break;}
case cast<uint32_t>(9ULL):{
rsp_CPU_lfv(c,vt,e,addr);
break;}
case cast<uint32_t>(10ULL):{
break;}
case cast<uint32_t>(11ULL):{
rsp_CPU_ltv(c,vt,e,addr);
break;}
default:{
rsp_CPU_unimpl(c,w);
break;}
}}
}
}
// tools/cpu/rsp/vu.go:706:1
void rsp_CPU_packedLoad(rsp_CPU* c,uint32_t vt,uint32_t e,uint32_t addr,uint32_t shift,uint32_t stride){
{
auto tmp25 = std::make_tuple((addr & ~(cast<uint32_t>(7ULL))),cast<uint32_t>((addr & cast<uint32_t>(7ULL))));
uint32_t base = std::get<0>(tmp25);
uint32_t mis = std::get<1>(tmp25);
for (uint32_t i = cast<uint32_t>(cast<uint32_t>(0ULL));(i < cast<uint32_t>(8ULL));i++){
uint8_t b = c->DMEM[cast<uint32_t>(((cast<uint32_t>((base + cast<uint32_t>(((cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(16ULL) - e)) + cast<uint32_t>((i * stride)))) + mis))) & cast<uint32_t>(15ULL)))))) & cast<uint32_t>(4095ULL)))];
c->V[vt][i] = shl<uint16_t>(cast<uint16_t>(b),shift);
}
}
}
// tools/cpu/rsp/vu.go:719:1
void rsp_CPU_lfv(rsp_CPU* c,uint32_t vt,uint32_t e,uint32_t addr){
{
auto tmp26 = std::make_tuple((addr & ~(cast<uint32_t>(7ULL))),cast<uint32_t>((addr & cast<uint32_t>(7ULL))));
uint32_t base = std::get<0>(tmp26);
uint32_t mis = std::get<1>(tmp26);
std::array<uint32_t,8> off = std::array<uint32_t,8>{cast<uint32_t>((mis + e)),cast<uint32_t>((cast<uint32_t>((mis + cast<uint32_t>(4ULL))) - e)),cast<uint32_t>((cast<uint32_t>((mis + cast<uint32_t>(8ULL))) - e)),cast<uint32_t>((cast<uint32_t>((mis + cast<uint32_t>(12ULL))) - e)),cast<uint32_t>((cast<uint32_t>((mis + cast<uint32_t>(8ULL))) - e)),cast<uint32_t>((cast<uint32_t>((mis + cast<uint32_t>(12ULL))) - e)),cast<uint32_t>((mis - e)),cast<uint32_t>((cast<uint32_t>((mis + cast<uint32_t>(4ULL))) - e))};
std::array<uint8_t,16> tmp={};
for (uint32_t i = cast<uint32_t>(cast<uint32_t>(0ULL));(i < cast<uint32_t>(8ULL));i++){
uint16_t v = shl<uint16_t>(cast<uint16_t>(c->DMEM[cast<uint32_t>(((cast<uint32_t>((base + cast<uint32_t>((off[i] & cast<uint32_t>(15ULL)))))) & cast<uint32_t>(4095ULL)))]),cast<int64_t>(7ULL));
tmp[cast<uint32_t>((i * cast<uint32_t>(2ULL)))] = cast<uint8_t>(shr<uint16_t>(v,cast<int64_t>(8ULL)));
tmp[cast<uint32_t>((cast<uint32_t>((i * cast<uint32_t>(2ULL))) + cast<uint32_t>(1ULL)))] = cast<uint8_t>(v);
}
for (uint32_t b = e;((b < cast<uint32_t>(16ULL)) && (b < cast<uint32_t>((e + cast<uint32_t>(8ULL)))));b++){
rsp_CPU_setVecByte(c,vt,b,tmp[b]);
}
}
}
// tools/cpu/rsp/vu.go:739:1
void rsp_CPU_ltv(rsp_CPU* c,uint32_t vt,uint32_t e,uint32_t addr){
{
uint32_t base = (addr & ~(cast<uint32_t>(7ULL)));
uint32_t rot = cast<uint32_t>((base & cast<uint32_t>(8ULL)));
for (uint32_t i = cast<uint32_t>(cast<uint32_t>(0ULL));(i < cast<uint32_t>(8ULL));i++){
uint32_t reg = cast<uint32_t>(((vt & ~(cast<uint32_t>(7ULL))) + cast<uint32_t>(((cast<uint32_t>((divi<uint32_t>(e,cast<uint32_t>(2ULL)) + i))) & cast<uint32_t>(7ULL)))));
for (uint32_t h = cast<uint32_t>(cast<uint32_t>(0ULL));(h < cast<uint32_t>(2ULL));h++){
rsp_CPU_setVecByte(c,reg,cast<uint32_t>((cast<uint32_t>((i * cast<uint32_t>(2ULL))) + h)),c->DMEM[cast<uint32_t>(((cast<uint32_t>((base + cast<uint32_t>(((cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((rot + e)) + cast<uint32_t>((i * cast<uint32_t>(2ULL))))) + h))) & cast<uint32_t>(15ULL)))))) & cast<uint32_t>(4095ULL)))]);
}
}
}
}
// tools/cpu/rsp/vu.go:751:1
void rsp_CPU_vecStore(rsp_CPU* c,uint32_t w,uint32_t rs,uint32_t vt){
{
uint32_t op = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(11ULL))) & cast<uint32_t>(31ULL)));
uint32_t e = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(7ULL))) & cast<uint32_t>(15ULL)));
int32_t off = shr<int32_t>(shl<int32_t>(cast<int32_t>(cast<uint32_t>((w & cast<uint32_t>(127ULL)))),cast<int64_t>(25ULL)),cast<int64_t>(25ULL));
uint32_t addr = cast<uint32_t>((c->R[rs] + cast<uint32_t>(shl<int32_t>(off,rsp_memShift[op]))));
{
switch(op){
case cast<uint32_t>(0ULL):case cast<uint32_t>(1ULL):case cast<uint32_t>(2ULL):case cast<uint32_t>(3ULL):{
uint32_t n = shl<uint32_t>(cast<uint32_t>(cast<uint32_t>(1ULL)),op);
for (uint32_t i = cast<uint32_t>(cast<uint32_t>(0ULL));(i < n);i++){
c->DMEM[cast<uint32_t>(((cast<uint32_t>((addr + i))) & cast<uint32_t>(4095ULL)))] = rsp_CPU_vecByte(c,vt,cast<uint32_t>(((cast<uint32_t>((e + i))) & cast<uint32_t>(15ULL))));
}
break;}
case cast<uint32_t>(4ULL):{
uint32_t end = cast<uint32_t>((((addr & ~(cast<uint32_t>(15ULL)))) + cast<uint32_t>(16ULL)));
for (uint32_t b = e;(addr < end);[&](){auto tmp27 = std::make_tuple(cast<uint32_t>((b + cast<uint32_t>(1ULL))),cast<uint32_t>((addr + cast<uint32_t>(1ULL))));
b = std::get<0>(tmp27);
addr = std::get<1>(tmp27);}()){
c->DMEM[cast<uint32_t>((addr & cast<uint32_t>(4095ULL)))] = rsp_CPU_vecByte(c,vt,cast<uint32_t>((b & cast<uint32_t>(15ULL))));
}
break;}
case cast<uint32_t>(5ULL):{
uint32_t n = cast<uint32_t>((addr & cast<uint32_t>(15ULL)));
uint32_t base = (addr & ~(cast<uint32_t>(15ULL)));
for (uint32_t i = cast<uint32_t>(cast<uint32_t>(0ULL));(i < n);i++){
c->DMEM[cast<uint32_t>(((cast<uint32_t>((base + i))) & cast<uint32_t>(4095ULL)))] = rsp_CPU_vecByte(c,vt,cast<uint32_t>(((cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((e + cast<uint32_t>(16ULL))) - n)) + i))) & cast<uint32_t>(15ULL))));
}
break;}
case cast<uint32_t>(6ULL):{
rsp_CPU_packedStore(c,vt,e,addr,cast<uint32_t>(8ULL),cast<uint32_t>(7ULL));
break;}
case cast<uint32_t>(7ULL):{
rsp_CPU_packedStore(c,vt,e,addr,cast<uint32_t>(7ULL),cast<uint32_t>(8ULL));
break;}
case cast<uint32_t>(8ULL):{
rsp_CPU_shv(c,vt,e,addr);
break;}
case cast<uint32_t>(9ULL):{
rsp_CPU_sfv(c,vt,e,addr);
break;}
case cast<uint32_t>(10ULL):{
auto tmp28 = std::make_tuple((addr & ~(cast<uint32_t>(7ULL))),cast<uint32_t>((addr & cast<uint32_t>(7ULL))));
uint32_t base = std::get<0>(tmp28);
uint32_t mis = std::get<1>(tmp28);
for (uint32_t i = cast<uint32_t>(cast<uint32_t>(0ULL));(i < cast<uint32_t>(16ULL));i++){
c->DMEM[cast<uint32_t>(((cast<uint32_t>((base + cast<uint32_t>(((cast<uint32_t>((mis + i))) & cast<uint32_t>(15ULL)))))) & cast<uint32_t>(4095ULL)))] = rsp_CPU_vecByte(c,vt,cast<uint32_t>(((cast<uint32_t>((e + i))) & cast<uint32_t>(15ULL))));
}
break;}
case cast<uint32_t>(11ULL):{
rsp_CPU_stv(c,vt,e,addr);
break;}
default:{
rsp_CPU_unimpl(c,w);
break;}
}}
}
}
// tools/cpu/rsp/vu.go:802:1
void rsp_CPU_packedStore(rsp_CPU* c,uint32_t vt,uint32_t e,uint32_t addr,uint32_t shift,uint32_t altShift){
{
for (uint32_t i = cast<uint32_t>(cast<uint32_t>(0ULL));(i < cast<uint32_t>(8ULL));i++){
uint32_t ei = cast<uint32_t>((e + i));
uint32_t s = shift;
if ((cast<uint32_t>((ei & cast<uint32_t>(8ULL))) != cast<uint32_t>(0ULL))) {
s = altShift;
}
c->DMEM[cast<uint32_t>(((cast<uint32_t>((addr + i))) & cast<uint32_t>(4095ULL)))] = cast<uint8_t>(shr<uint16_t>(c->V[vt][cast<uint32_t>((ei & cast<uint32_t>(7ULL)))],s));
}
}
}
// tools/cpu/rsp/vu.go:816:1
void rsp_CPU_shv(rsp_CPU* c,uint32_t vt,uint32_t e,uint32_t addr){
{
auto tmp29 = std::make_tuple((addr & ~(cast<uint32_t>(7ULL))),cast<uint32_t>((addr & cast<uint32_t>(7ULL))));
uint32_t base = std::get<0>(tmp29);
uint32_t mis = std::get<1>(tmp29);
for (uint32_t i = cast<uint32_t>(cast<uint32_t>(0ULL));(i < cast<uint32_t>(8ULL));i++){
uint32_t ei = cast<uint32_t>((e + cast<uint32_t>((i * cast<uint32_t>(2ULL)))));
uint16_t v = cast<uint16_t>((shl<uint16_t>(cast<uint16_t>(rsp_CPU_vecByte(c,vt,cast<uint32_t>((ei & cast<uint32_t>(15ULL))))),cast<int64_t>(8ULL)) | cast<uint16_t>(rsp_CPU_vecByte(c,vt,cast<uint32_t>(((cast<uint32_t>((ei + cast<uint32_t>(1ULL)))) & cast<uint32_t>(15ULL)))))));
c->DMEM[cast<uint32_t>(((cast<uint32_t>((base + cast<uint32_t>(((cast<uint32_t>((mis + cast<uint32_t>((i * cast<uint32_t>(2ULL)))))) & cast<uint32_t>(15ULL)))))) & cast<uint32_t>(4095ULL)))] = cast<uint8_t>(shr<uint16_t>(v,cast<int64_t>(7ULL)));
}
}
}
// tools/cpu/rsp/vu.go:831:1
void rsp_CPU_sfv(rsp_CPU* c,uint32_t vt,uint32_t e,uint32_t addr){
{
auto tmp30 = std::make_tuple((addr & ~(cast<uint32_t>(7ULL))),cast<uint32_t>((addr & cast<uint32_t>(7ULL))));
uint32_t base = std::get<0>(tmp30);
uint32_t mis = std::get<1>(tmp30);
auto tmp31 = lookup(rsp_sfvStart,e);
uint32_t start = std::get<0>(tmp31);
bool ok = std::get<1>(tmp31);
for (uint32_t i = cast<uint32_t>(cast<uint32_t>(0ULL));(i < cast<uint32_t>(4ULL));i++){
uint8_t b={};
if (ok) {
uint32_t lane = cast<uint32_t>((cast<uint32_t>((start & cast<uint32_t>(4ULL))) | cast<uint32_t>(((cast<uint32_t>((start + i))) & cast<uint32_t>(3ULL)))));
b = cast<uint8_t>(shr<uint16_t>(c->V[vt][lane],cast<int64_t>(7ULL)));
}
c->DMEM[cast<uint32_t>(((cast<uint32_t>((base + cast<uint32_t>(((cast<uint32_t>((mis + cast<uint32_t>((i * cast<uint32_t>(4ULL)))))) & cast<uint32_t>(15ULL)))))) & cast<uint32_t>(4095ULL)))] = b;
}
}
}
// tools/cpu/rsp/vu.go:848:1
void rsp_CPU_stv(rsp_CPU* c,uint32_t vt,uint32_t e,uint32_t addr){
{
uint32_t base = (addr & ~(cast<uint32_t>(7ULL)));
uint32_t rot = divi<uint32_t>((cast<uint32_t>((base & cast<uint32_t>(8ULL)))),cast<uint32_t>(2ULL));
for (uint32_t i = cast<uint32_t>(cast<uint32_t>(0ULL));(i < cast<uint32_t>(16ULL));i++){
uint32_t reg = cast<uint32_t>(((vt & ~(cast<uint32_t>(7ULL))) + cast<uint32_t>(((cast<uint32_t>((cast<uint32_t>((divi<uint32_t>(i,cast<uint32_t>(2ULL)) - rot)) + divi<uint32_t>(e,cast<uint32_t>(2ULL))))) & cast<uint32_t>(7ULL)))));
c->DMEM[cast<uint32_t>(((cast<uint32_t>((base + cast<uint32_t>(((cast<uint32_t>((addr + i))) & cast<uint32_t>(15ULL)))))) & cast<uint32_t>(4095ULL)))] = rsp_CPU_vecByte(c,reg,cast<uint32_t>(((cast<uint32_t>((i + base))) & cast<uint32_t>(15ULL))));
}
}
}
// tools/cpu/rsp/vu.go:872:1
void rsp_CPU_divide(rsp_CPU* c,uint32_t e,std::array<uint16_t,8> vtv,uint32_t vs,uint32_t vd,uint32_t result){
{
c->V[vd][cast<uint32_t>((vs & cast<uint32_t>(7ULL)))] = cast<uint16_t>(result);
for (uint32_t i = cast<uint32_t>(cast<uint32_t>(0ULL));(i < cast<uint32_t>(8ULL));i++){
rsp_CPU_setAccLo(c,i,rsp_vte(vtv,e,i));
}
}
}
// tools/cpu/rsp/vu.go:882:1
int32_t rsp_CPU_divInput(rsp_CPU* c,uint16_t lo){
{
int32_t in = cast<int32_t>(cast<int16_t>(lo));
if (c->divInLoaded) {
in = cast<int32_t>(cast<uint32_t>((shl<uint32_t>(cast<uint32_t>(c->divIn),cast<int64_t>(16ULL)) | cast<uint32_t>(lo))));
}
auto tmp32 = std::make_tuple(cast<uint16_t>(0ULL),false);
c->divIn = std::get<0>(tmp32);
c->divInLoaded = std::get<1>(tmp32);
return in;
}
}
// tools/cpu/rsp/vu.go:912:1
uint32_t rsp_CPU_reciprocal(rsp_CPU* c,int32_t in){
{
int32_t mask = shr<int32_t>(in,cast<int64_t>(31ULL));
uint32_t x = cast<uint32_t>(cast<int32_t>((in ^ mask)));
if ((in > cast<int32_t>(-cast<int64_t>(32768ULL)))) {
x = cast<uint32_t>(cast<int32_t>(((cast<int32_t>((in ^ mask))) - mask)));
}
{
if ((x == cast<uint32_t>(0ULL))){
c->divOut = cast<uint16_t>(32767ULL);
return cast<uint32_t>(2147483647ULL);
}
else if ((in == cast<int32_t>(-cast<int64_t>(32768ULL)))){
c->divOut = cast<uint16_t>(65535ULL);
return cast<uint32_t>(4294901760ULL);
}
}
tmp33:;
uint32_t scaleIn = cast<uint32_t>(cast<uint32_t>(31ULL));
for (;(cast<uint32_t>((x & (shl<uint32_t>(cast<uint32_t>(1ULL),scaleIn)))) == cast<uint32_t>(0ULL));){
scaleIn--;
}
uint32_t idx={};
if ((scaleIn >= cast<uint32_t>(9ULL))) {
idx = cast<uint32_t>((shr<uint32_t>(x,(cast<uint32_t>((scaleIn - cast<uint32_t>(9ULL))))) & cast<uint32_t>(511ULL)));
}
else {
idx = cast<uint32_t>((shl<uint32_t>(x,(cast<uint32_t>((cast<uint32_t>(9ULL) - scaleIn)))) & cast<uint32_t>(511ULL)));
}
uint64_t v = cast<uint64_t>((cast<uint64_t>(cast<uint64_t>(65536ULL)) | cast<uint64_t>(rsp_rcpROM[idx])));
uint32_t result={};
if ((scaleIn <= cast<uint32_t>(14ULL))) {
result = cast<uint32_t>(shl<uint64_t>(v,(cast<uint32_t>((cast<uint32_t>(14ULL) - scaleIn)))));
}
else {
result = cast<uint32_t>(shr<uint64_t>(v,(cast<uint32_t>((scaleIn - cast<uint32_t>(14ULL))))));
}
result ^= cast<uint32_t>(mask);
c->divOut = cast<uint16_t>(shr<uint32_t>(result,cast<int64_t>(16ULL)));
return result;
}
}
// tools/cpu/rsp/vu.go:967:1
uint32_t rsp_CPU_rsqrt(rsp_CPU* c,int32_t in){
{
int32_t mask = shr<int32_t>(in,cast<int64_t>(31ULL));
uint32_t x = cast<uint32_t>(cast<int32_t>((in ^ mask)));
if ((in > cast<int32_t>(-cast<int64_t>(32768ULL)))) {
x = cast<uint32_t>(cast<int32_t>(((cast<int32_t>((in ^ mask))) - mask)));
}
{
if ((x == cast<uint32_t>(0ULL))){
c->divOut = cast<uint16_t>(32767ULL);
return cast<uint32_t>(2147483647ULL);
}
else if ((in == cast<int32_t>(-cast<int64_t>(32768ULL)))){
c->divOut = cast<uint16_t>(65535ULL);
return cast<uint32_t>(4294901760ULL);
}
}
tmp34:;
uint32_t scaleIn = cast<uint32_t>(cast<uint32_t>(31ULL));
for (;(cast<uint32_t>((x & (shl<uint32_t>(cast<uint32_t>(1ULL),scaleIn)))) == cast<uint32_t>(0ULL));){
scaleIn--;
}
uint32_t idx={};
if ((scaleIn >= cast<uint32_t>(8ULL))) {
idx = cast<uint32_t>((shr<uint32_t>(x,(cast<uint32_t>((scaleIn - cast<uint32_t>(8ULL))))) & cast<uint32_t>(255ULL)));
}
else {
idx = cast<uint32_t>((shl<uint32_t>(x,(cast<uint32_t>((cast<uint32_t>(8ULL) - scaleIn)))) & cast<uint32_t>(255ULL)));
}
idx |= shl<uint32_t>((cast<uint32_t>((scaleIn & cast<uint32_t>(1ULL)))),cast<int64_t>(8ULL));
uint64_t v = cast<uint64_t>((cast<uint64_t>(cast<uint64_t>(65536ULL)) | cast<uint64_t>(rsp_rsqROM[idx])));
uint32_t result = cast<uint32_t>(shr<uint64_t>(shl<uint64_t>(v,cast<int64_t>(14ULL)),(shr<uint32_t>(scaleIn,cast<int64_t>(1ULL)))));
result ^= cast<uint32_t>(mask);
c->divOut = cast<uint16_t>(shr<uint32_t>(result,cast<int64_t>(16ULL)));
return result;
}
}
// tools/platform/n64/ai.go:34:1
void n64_ai_init(n64_ai* a){
{
a->Regs = n64_regFile{};
}
}
// tools/platform/n64/ai.go:36:1
uint32_t n64_Machine_aiRead(n64_Machine* m,uint32_t addr){
{
{
switch(cast<uint32_t>((addr & cast<uint32_t>(255ULL)))){
case n64_aiLength:case n64_aiStatus:{
return cast<uint32_t>(0ULL);
break;}
}}
return get(m->ai.Regs,cast<uint32_t>((addr & cast<uint32_t>(255ULL))));
}
}
// tools/platform/n64/ai.go:44:1
void n64_Machine_aiWrite(n64_Machine* m,uint32_t addr,uint32_t v){
{
{
switch(cast<uint32_t>((addr & cast<uint32_t>(255ULL)))){
case n64_aiStatus:{
n64_Machine_clearIRQ(m,n64_intrAI);
return ;
break;}
case n64_aiLength:{
uint32_t length = cast<uint32_t>((v & cast<uint32_t>(262136ULL)));
m->ai.Regs[n64_aiLength] = length;
if ((((!m->hookMuted) && bool(m->OnAIBuffer)) && (length > cast<uint32_t>(0ULL)))) {
m->OnAIBuffer(cast<uint32_t>((get(m->ai.Regs,n64_aiDramAddr) & cast<uint32_t>(16777215ULL))),length,get(m->ai.Regs,n64_aiDacRate));
}
n64_Machine_raiseIRQ(m,n64_intrAI);
return ;
break;}
}}
m->ai.Regs[cast<uint32_t>((addr & cast<uint32_t>(255ULL)))] = v;
}
}
// tools/platform/n64/boot.go:71:1
n64_BootConfig n64_DefaultBoot(){
{
return n64_BootConfig{n64_TVNTSC,n64_ResetCold,cast<uint32_t>(0ULL)};
}
}
// tools/platform/n64/boot.go:133:1
uint32_t n64_Machine_OSMemSize(n64_Machine* m){
{
auto tmp1 = n64_Machine_ReadVirt(m,n64_osMemSizeAddr);
uint32_t v = std::get<0>(tmp1);
return v;
}
}
// tools/platform/n64/dp.go:32:1
uint32_t n64_Machine_dpRead(n64_Machine* m,uint32_t addr){
{
{
switch(cast<uint32_t>((addr & cast<uint32_t>(31ULL)))){
case n64_dpCurrent:case n64_dpEnd:{
return get(m->dp,n64_dpEnd);
break;}
case n64_dpStatus:{
return cast<uint32_t>((get(m->dp,n64_dpStatus) | n64_dpStatusCbufReady));
break;}
}}
return get(m->dp,cast<uint32_t>((addr & cast<uint32_t>(31ULL))));
}
}
// tools/platform/n64/dp.go:44:1
void n64_Machine_dpWrite(n64_Machine* m,uint32_t addr,uint32_t v){
{
{
switch(cast<uint32_t>((addr & cast<uint32_t>(31ULL)))){
case n64_dpStart:{
if ((cast<uint32_t>((get(m->dp,n64_dpStatus) & n64_dpStatusStartValid)) == cast<uint32_t>(0ULL))) {
m->dp[n64_dpStart] = cast<uint32_t>((v & cast<uint32_t>(16777208ULL)));
m->dp[n64_dpStatus] |= n64_dpStatusStartValid;
}
break;}
case n64_dpEnd:{
m->dp[n64_dpEnd] = cast<uint32_t>((v & cast<uint32_t>(16777208ULL)));
if ((cast<uint32_t>((get(m->dp,n64_dpStatus) & n64_dpStatusStartValid)) != cast<uint32_t>(0ULL))) {
m->dp[n64_dpCurrent] = get(m->dp,n64_dpStart);
m->dp[n64_dpStatus] &= ~(n64_dpStatusStartValid);
}
if ((cast<uint32_t>((get(m->dp,n64_dpStatus) & n64_dpStatusFreeze)) == cast<uint32_t>(0ULL))) {
n64_Machine_runRDP(m);
}
break;}
case n64_dpStatus:{
if ((cast<uint32_t>((v & (shl<uint32_t>(cast<int64_t>(1ULL),cast<int64_t>(0ULL))))) != cast<uint32_t>(0ULL))) {
m->dp[n64_dpStatus] &= ~(n64_dpStatusXBusDMEM);
}
if ((cast<uint32_t>((v & (shl<uint32_t>(cast<int64_t>(1ULL),cast<int64_t>(1ULL))))) != cast<uint32_t>(0ULL))) {
m->dp[n64_dpStatus] |= n64_dpStatusXBusDMEM;
}
if ((cast<uint32_t>((v & (shl<uint32_t>(cast<int64_t>(1ULL),cast<int64_t>(2ULL))))) != cast<uint32_t>(0ULL))) {
m->dp[n64_dpStatus] &= ~(n64_dpStatusFreeze);
n64_Machine_runRDP(m);
}
if ((cast<uint32_t>((v & (shl<uint32_t>(cast<int64_t>(1ULL),cast<int64_t>(3ULL))))) != cast<uint32_t>(0ULL))) {
m->dp[n64_dpStatus] |= n64_dpStatusFreeze;
}
if ((cast<uint32_t>((v & (shl<uint32_t>(cast<int64_t>(1ULL),cast<int64_t>(4ULL))))) != cast<uint32_t>(0ULL))) {
m->dp[n64_dpStatus] &= ~(n64_dpStatusFlush);
}
if ((cast<uint32_t>((v & (shl<uint32_t>(cast<int64_t>(1ULL),cast<int64_t>(5ULL))))) != cast<uint32_t>(0ULL))) {
m->dp[n64_dpStatus] |= n64_dpStatusFlush;
}
break;}
default:{
m->dp[cast<uint32_t>((addr & cast<uint32_t>(31ULL)))] = v;
break;}
}}
}
}
// tools/platform/n64/eeprom.go:28:1
void n64_Machine_joybusEEPROM(n64_Machine* m,Slice<uint8_t> cmd,Slice<uint8_t> res,int64_t rxAt){
{
if ((len(m->EEPROM) == cast<int64_t>(0ULL))) {
m->PIF[rxAt] |= cast<uint8_t>(128ULL);
return ;
}
{
switch(cmd[cast<int64_t>(0ULL)]){
case n64_jbInfo:case n64_jbReset:{
if ((len(res) >= cast<int64_t>(3ULL))) {
uint16_t id = n64_devEEPROM4K;
if ((len(m->EEPROM) > cast<int64_t>((n64_eeprom4KBlocks * n64_eepromBlockSize)))) {
id = n64_devEEPROM16K;
}
res[cast<int64_t>(0ULL)] = cast<uint8_t>(shr<uint16_t>(id,cast<int64_t>(8ULL)));
res[cast<int64_t>(1ULL)] = cast<uint8_t>(id);
res[cast<int64_t>(2ULL)] = cast<uint8_t>(0ULL);
}
break;}
case n64_jbEepromRead:{
if (((len(cmd) < cast<int64_t>(2ULL)) || (len(res) < n64_eepromBlockSize))) {
m->PIF[rxAt] |= cast<uint8_t>(64ULL);
return ;
}
int64_t off = cast<int64_t>((cast<int64_t>(cmd[cast<int64_t>(1ULL)]) * n64_eepromBlockSize));
if ((cast<int64_t>((off + n64_eepromBlockSize)) > len(m->EEPROM))) {
m->PIF[rxAt] |= cast<uint8_t>(64ULL);
return ;
}
gcopy(res,sub(m->EEPROM,off,cast<int64_t>((off + n64_eepromBlockSize))));
break;}
case n64_jbEepromWrite:{
if ((len(cmd) < cast<int64_t>((cast<int64_t>(2ULL) + n64_eepromBlockSize)))) {
m->PIF[rxAt] |= cast<uint8_t>(64ULL);
return ;
}
int64_t off = cast<int64_t>((cast<int64_t>(cmd[cast<int64_t>(1ULL)]) * n64_eepromBlockSize));
if ((cast<int64_t>((off + n64_eepromBlockSize)) > len(m->EEPROM))) {
m->PIF[rxAt] |= cast<uint8_t>(64ULL);
return ;
}
gcopy(sub(m->EEPROM,off,cast<int64_t>((off + n64_eepromBlockSize))),sub(cmd,cast<int64_t>(2ULL),cast<int64_t>((cast<int64_t>(2ULL) + n64_eepromBlockSize))));
if ((len(res) >= cast<int64_t>(1ULL))) {
res[cast<int64_t>(0ULL)] = cast<uint8_t>(0ULL);
}
break;}
default:{
n64_Machine_note(m,"joybus: unmodelled EEPROM command 0x%02X",cmd[cast<int64_t>(0ULL)]);
m->PIF[rxAt] |= cast<uint8_t>(128ULL);
break;}
}}
}
}
// tools/platform/n64/isviewer.go:31:1
bool n64_inISViewer(uint32_t addr){
{
return ((addr >= n64_isvBase) && (addr < n64_isvEnd));
}
}
// tools/platform/n64/isviewer.go:35:1
uint32_t n64_Machine_isvRead(n64_Machine* m,uint32_t addr){
{
if ((addr >= n64_isvBuffer)) {
uint32_t off = cast<uint32_t>((addr - n64_isvBuffer));
return cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((shl<uint32_t>(cast<uint32_t>(m->isv.Buf[off]),cast<int64_t>(24ULL)) | shl<uint32_t>(cast<uint32_t>(m->isv.Buf[cast<uint32_t>((off + cast<uint32_t>(1ULL)))]),cast<int64_t>(16ULL)))) | shl<uint32_t>(cast<uint32_t>(m->isv.Buf[cast<uint32_t>((off + cast<uint32_t>(2ULL)))]),cast<int64_t>(8ULL)))) | cast<uint32_t>(m->isv.Buf[cast<uint32_t>((off + cast<uint32_t>(3ULL)))])));
}
return cast<uint32_t>(0ULL);
}
}
// tools/platform/n64/isviewer.go:44:1
void n64_Machine_isvWrite(n64_Machine* m,uint32_t addr,uint32_t v){
{
{
if ((addr == n64_isvLen)){
int64_t n = cast<int64_t>(v);
if ((n > len(m->isv.Buf))) {
n = len(m->isv.Buf);
}
n64_Machine_isvPrint(m,cast<std::string>(sub(m->isv.Buf,0,n)));
}
else if (((addr >= n64_isvBuffer) && (cast<uint32_t>((addr + cast<uint32_t>(3ULL))) < n64_isvEnd))){
uint32_t off = cast<uint32_t>((addr - n64_isvBuffer));
m->isv.Buf[off] = cast<uint8_t>(shr<uint32_t>(v,cast<int64_t>(24ULL)));
m->isv.Buf[cast<uint32_t>((off + cast<uint32_t>(1ULL)))] = cast<uint8_t>(shr<uint32_t>(v,cast<int64_t>(16ULL)));
m->isv.Buf[cast<uint32_t>((off + cast<uint32_t>(2ULL)))] = cast<uint8_t>(shr<uint32_t>(v,cast<int64_t>(8ULL)));
m->isv.Buf[cast<uint32_t>((off + cast<uint32_t>(3ULL)))] = cast<uint8_t>(v);
}
}
tmp2:;
}
}
// tools/platform/n64/isviewer.go:62:1
void n64_Machine_isvWriteByte(n64_Machine* m,uint32_t addr,uint8_t v){
{
if (((addr >= n64_isvBuffer) && (addr < n64_isvEnd))) {
m->isv.Buf[cast<uint32_t>((addr - n64_isvBuffer))] = v;
}
}
}
// tools/platform/n64/isviewer.go:69:1
void n64_Machine_isvPrint(n64_Machine* m,std::string s){
{
std::string line = "";
{auto&& tmp3 = s;
for(int64_t tmp4=0;tmp4<len(tmp3);++tmp4){
auto c=tmp3[tmp4];if (((c == '\n') || (c == '\r'))) {
if ((line != "")) {
n64_Machine_isvEmit(m,line);
line = "";
}
continue;
}
line += cast<std::string>(c);
}}
if ((line != "")) {
n64_Machine_isvEmit(m,line);
}
}
}
// tools/platform/n64/isviewer.go:86:1
void n64_Machine_isvEmit(n64_Machine* m,std::string line){
{
m->isv.Lines = append(m->isv.Lines,line);
if (bool(m->OnPrint)) {
m->OnPrint(m,line);
}
}
}
// tools/platform/n64/isviewer.go:94:1
Slice<std::string> n64_Machine_ISViewerLines(n64_Machine* m){
{
return m->isv.Lines;
}
}
// tools/platform/n64/machine.go:201:1
r4300_CPU* n64_newBareCPU(n64_Machine* m){
{
return r4300_NewCPU(m);
}
}
// tools/platform/n64/machine.go:217:1
uint8_t n64_Machine_rdramRead(n64_Machine* m,uint32_t a){
{
if ((cast<int64_t>(a) < len(m->RDRAM))) {
return m->RDRAM[a];
}
return cast<uint8_t>(0ULL);
}
}
// tools/platform/n64/machine.go:224:1
void n64_Machine_rdramWrite(n64_Machine* m,uint32_t a,uint8_t v){
{
if ((cast<int64_t>(a) < len(m->RDRAM))) {
m->RDRAM[a] = v;
}
}
}
// tools/platform/n64/machine.go:242:1
uint32_t n64_Machine_pc(n64_Machine* m){
{
return cast<uint32_t>(r4300_CPU_CurPC(m->CPU));
}
}
// tools/platform/n64/machine.go:248:1
std::tuple<Slice<uint8_t>,uint32_t> n64_Machine_backing(n64_Machine* m,uint32_t addr){
{
{
if ((addr < n64_rdramSize)){
return {m->RDRAM,addr};
}
else if (((addr >= n64_spDMEMBase) && (addr < n64_spIMEMBase))){
return {m->DMEM,cast<uint32_t>((addr - n64_spDMEMBase))};
}
else if (((addr >= n64_spIMEMBase) && (addr < n64_spMemEnd))){
return {m->IMEM,cast<uint32_t>((addr - n64_spIMEMBase))};
}
else if (n64_inISViewer(addr)){
return {{},cast<uint32_t>(0ULL)};
}
else if (((addr >= n64_cartBase) && (addr < n64_cartEnd))){
uint32_t off = cast<uint32_t>((addr - n64_cartBase));
if ((cast<int64_t>(off) < len(m->ROM))) {
return {m->ROM,off};
}
}
else if (((addr >= n64_pifBase) && (addr < n64_pifEnd))){
if ((addr >= cast<uint32_t>((n64_pifEnd - cast<int64_t>(64ULL))))) {
return {m->PIF,cast<uint32_t>((addr - (cast<uint32_t>((n64_pifEnd - cast<int64_t>(64ULL))))))};
}
}
}
tmp5:;
return {{},cast<uint32_t>(0ULL)};
}
}
// tools/platform/n64/machine.go:318:1
uint8_t n64_Machine_Read(n64_Machine* m,uint32_t addr){
{
{
auto tmp6 = n64_Machine_backing(m,addr);
Slice<uint8_t> b = std::get<0>(tmp6);
uint32_t off = std::get<1>(tmp6);
if (bool(b)) {
uint8_t v = b[off];
if (((((!m->hookMuted) && bool(m->OnRead)) && (addr >= m->RWatchLo)) && (addr < m->RWatchHi))) {
m->OnRead(addr,cast<uint32_t>(v),n64_Machine_pc(m));
}
return v;
}
}
uint32_t w = n64_Machine_Read32(m,(addr & ~(cast<uint32_t>(3ULL))));
return cast<uint8_t>(shr<uint32_t>(w,(cast<uint32_t>((cast<uint32_t>(8ULL) * (cast<uint32_t>((cast<uint32_t>(3ULL) - cast<uint32_t>((addr & cast<uint32_t>(3ULL)))))))))));
}
}
// tools/platform/n64/machine.go:330:1
void n64_Machine_Write(n64_Machine* m,uint32_t addr,uint8_t v){
{
if (((((!m->hookMuted) && bool(m->OnWrite)) && (addr >= m->WatchLo)) && (addr < m->WatchHi))) {
m->OnWrite(addr,cast<uint32_t>(v),n64_Machine_pc(m));
}
if (n64_inISViewer(addr)) {
n64_Machine_isvWriteByte(m,addr,v);
return ;
}
{
auto tmp7 = n64_Machine_backing(m,addr);
Slice<uint8_t> b = std::get<0>(tmp7);
uint32_t off = std::get<1>(tmp7);
if (bool(b)) {
if (((addr >= n64_cartBase) && (addr < n64_cartEnd))) {
n64_Machine_note(m,"write 0x%02X to the cartridge at 0x%08X (ignored)",v,addr);
return ;
}
b[off] = v;
return ;
}
}
uint32_t shift = cast<uint32_t>((cast<uint32_t>(8ULL) * (cast<uint32_t>((cast<uint32_t>(3ULL) - cast<uint32_t>((addr & cast<uint32_t>(3ULL))))))));
uint32_t w = cast<uint32_t>(((n64_Machine_Read32(m,(addr & ~(cast<uint32_t>(3ULL)))) & ~((shl<uint32_t>(cast<uint32_t>(255ULL),shift)))) | shl<uint32_t>(cast<uint32_t>(v),shift)));
n64_Machine_Write32(m,(addr & ~(cast<uint32_t>(3ULL))),w);
}
}
// tools/platform/n64/machine.go:352:1
uint32_t n64_Machine_ioRead(n64_Machine* m,uint32_t addr){
{
{
if ((addr < n64_rdramEnd)){
return cast<uint32_t>(0ULL);
}
else if (((addr >= n64_rdramRegBase) && (addr < n64_rdramRegEnd))){
return get(m->rd,cast<uint32_t>((addr & cast<uint32_t>(255ULL))));
}
else if (((addr >= n64_spRegBase) && (addr < n64_spRegEnd))){
return n64_Machine_spRead(m,addr);
}
else if (((addr >= n64_dpRegBase) && (addr < n64_dpRegEnd))){
return n64_Machine_dpRead(m,addr);
}
else if (((addr >= n64_miRegBase) && (addr < n64_miRegEnd))){
return n64_Machine_miRead(m,addr);
}
else if (((addr >= n64_viRegBase) && (addr < n64_viRegEnd))){
return n64_Machine_viRead(m,addr);
}
else if (((addr >= n64_aiRegBase) && (addr < n64_aiRegEnd))){
return n64_Machine_aiRead(m,addr);
}
else if (((addr >= n64_piRegBase) && (addr < n64_piRegEnd))){
return n64_Machine_piRead(m,addr);
}
else if (((addr >= n64_riRegBase) && (addr < n64_riRegEnd))){
return get(m->ri,cast<uint32_t>((addr & cast<uint32_t>(255ULL))));
}
else if (((addr >= n64_siRegBase) && (addr < n64_siRegEnd))){
return n64_Machine_siRead(m,addr);
}
else if (n64_inISViewer(addr)){
return n64_Machine_isvRead(m,addr);
}
}
tmp8:;
n64_Machine_note(m,"read from unmapped physical address 0x%08X (returning 0)",addr);
return cast<uint32_t>(0ULL);
}
}
// tools/platform/n64/machine.go:384:1
void n64_Machine_ioWrite(n64_Machine* m,uint32_t addr,uint32_t v){
{
{
if ((addr < n64_rdramEnd)){
return ;
}
else if (((addr >= n64_rdramRegBase) && (addr < n64_rdramRegEnd))){
m->rd[cast<uint32_t>((addr & cast<uint32_t>(255ULL)))] = v;
}
else if (((addr >= n64_spRegBase) && (addr < n64_spRegEnd))){
n64_Machine_spWrite(m,addr,v);
}
else if (((addr >= n64_dpRegBase) && (addr < n64_dpRegEnd))){
n64_Machine_dpWrite(m,addr,v);
}
else if (((addr >= n64_miRegBase) && (addr < n64_miRegEnd))){
n64_Machine_miWrite(m,addr,v);
}
else if (((addr >= n64_viRegBase) && (addr < n64_viRegEnd))){
n64_Machine_viWrite(m,addr,v);
}
else if (((addr >= n64_aiRegBase) && (addr < n64_aiRegEnd))){
n64_Machine_aiWrite(m,addr,v);
}
else if (((addr >= n64_piRegBase) && (addr < n64_piRegEnd))){
n64_Machine_piWrite(m,addr,v);
}
else if (((addr >= n64_riRegBase) && (addr < n64_riRegEnd))){
m->ri[cast<uint32_t>((addr & cast<uint32_t>(255ULL)))] = v;
}
else if (((addr >= n64_siRegBase) && (addr < n64_siRegEnd))){
n64_Machine_siWrite(m,addr,v);
}
else if (n64_inISViewer(addr)){
n64_Machine_isvWrite(m,addr,v);
}
else {
n64_Machine_note(m,"write 0x%08X to unmapped physical address 0x%08X (ignored)",v,addr);
}
}
tmp9:;
}
}
// tools/platform/n64/machine.go:430:1
void n64_Machine_writePhys32(n64_Machine* m,uint32_t addr,uint32_t v){
{
bool muted = m->hookMuted;
m->hookMuted = true;
n64_Machine_Write32(m,addr,v);
m->hookMuted = muted;
}
}
// tools/platform/n64/machine.go:438:1
std::tuple<uint32_t,bool> n64_Machine_ReadVirt(n64_Machine* m,uint64_t vaddr){
{
auto tmp10 = r4300_CPU_Translate(m->CPU,vaddr,false);
uint32_t p = std::get<0>(tmp10);
bool ok = std::get<1>(tmp10);
if ((!ok)) {
return {cast<uint32_t>(0ULL),false};
}
bool muted = m->hookMuted;
m->hookMuted = true;
uint32_t v = n64_Machine_Read32(m,p);
m->hookMuted = muted;
return {v,true};
}
}
// tools/platform/n64/machine.go:452:1
uint64_t n64_Machine_RSPSteps(n64_Machine* m){
{
return m->rspSteps;
}
}
// tools/platform/n64/machine.go:453:1
uint64_t n64_Machine_RDPWords(n64_Machine* m){
{
return m->rdpWords;
}
}
// tools/platform/n64/mi.go:46:1
void n64_Machine_raiseIRQ(n64_Machine* m,uint32_t bit){
{
m->mi.Intr |= bit;
}
}
// tools/platform/n64/mi.go:49:1
void n64_Machine_clearIRQ(n64_Machine* m,uint32_t bit){
{
m->mi.Intr &= ~(bit);
}
}
// tools/platform/n64/mi.go:52:1
bool n64_Machine_irqPending(n64_Machine* m){
{
return (cast<uint32_t>((m->mi.Intr & m->mi.Mask)) != cast<uint32_t>(0ULL));
}
}
// tools/platform/n64/mi.go:54:1
uint32_t n64_Machine_miRead(n64_Machine* m,uint32_t addr){
{
{
switch(cast<uint32_t>((addr & cast<uint32_t>(255ULL)))){
case n64_miInitMode:{
return m->mi.InitMode;
break;}
case n64_miVersion:{
return n64_miVersionValue;
break;}
case n64_miIntr:{
return m->mi.Intr;
break;}
case n64_miIntrMask:{
return m->mi.Mask;
break;}
}}
n64_Machine_note(m,"MI: read from undecoded register 0x%08X",addr);
return cast<uint32_t>(0ULL);
}
}
// tools/platform/n64/mi.go:72:1
void n64_Machine_miWrite(n64_Machine* m,uint32_t addr,uint32_t v){
{
{
switch(cast<uint32_t>((addr & cast<uint32_t>(255ULL)))){
case n64_miInitMode:{
m->mi.InitMode = cast<uint32_t>((((m->mi.InitMode & ~(cast<uint32_t>(127ULL)))) | (cast<uint32_t>((v & cast<uint32_t>(127ULL))))));
if ((cast<uint32_t>((v & (shl<uint32_t>(cast<int64_t>(1ULL),cast<int64_t>(11ULL))))) != cast<uint32_t>(0ULL))) {
n64_Machine_clearIRQ(m,n64_intrDP);
}
break;}
case n64_miIntrMask:{
for (uint32_t i = cast<uint32_t>(cast<uint32_t>(0ULL));(i < cast<uint32_t>(6ULL));i++){
if ((cast<uint32_t>((v & (shl<uint32_t>(cast<uint32_t>(1ULL),(cast<uint32_t>((cast<uint32_t>(2ULL) * i))))))) != cast<uint32_t>(0ULL))) {
m->mi.Mask &= ~(shl<uint32_t>(cast<uint32_t>(1ULL),i));
}
if ((cast<uint32_t>((v & (shl<uint32_t>(cast<uint32_t>(1ULL),(cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(2ULL) * i)) + cast<uint32_t>(1ULL)))))))) != cast<uint32_t>(0ULL))) {
m->mi.Mask |= shl<uint32_t>(cast<uint32_t>(1ULL),i);
}
}
break;}
case n64_miIntr:case n64_miVersion:{
break;}
default:{
n64_Machine_note(m,"MI: write 0x%08X to undecoded register 0x%08X",v,addr);
break;}
}}
}
}
// tools/platform/n64/pi.go:40:1
void n64_pi_init(n64_pi* p){
{
p->Regs = n64_regFile{};
}
}
// tools/platform/n64/pi.go:42:1
uint32_t n64_Machine_piRead(n64_Machine* m,uint32_t addr){
{
{
switch(cast<uint32_t>((addr & cast<uint32_t>(255ULL)))){
case n64_piDramAddr:{
return m->pi.DramAddr;
break;}
case n64_piCartAddr:{
return m->pi.CartAddr;
break;}
case n64_piRdLen:case n64_piWrLen:{
return get(m->pi.Regs,cast<uint32_t>((addr & cast<uint32_t>(255ULL))));
break;}
case n64_piStatus:{
return m->pi.Status;
break;}
}}
return get(m->pi.Regs,cast<uint32_t>((addr & cast<uint32_t>(255ULL))));
}
}
// tools/platform/n64/pi.go:61:1
void n64_Machine_piWrite(n64_Machine* m,uint32_t addr,uint32_t v){
{
{
switch(cast<uint32_t>((addr & cast<uint32_t>(255ULL)))){
case n64_piDramAddr:{
m->pi.DramAddr = cast<uint32_t>((cast<uint32_t>((v & cast<uint32_t>(16777215ULL))) & cast<uint32_t>(~cast<uint32_t>(cast<uint32_t>(1ULL)))));
break;}
case n64_piCartAddr:{
m->pi.CartAddr = (v & ~(cast<uint32_t>(1ULL)));
break;}
case n64_piRdLen:{
m->pi.Regs[n64_piRdLen] = v;
n64_Machine_note(m,"PI: write-to-cartridge DMA of %d bytes (ignored: the ROM is read-only)",cast<uint32_t>((v + cast<uint32_t>(1ULL))));
n64_Machine_raiseIRQ(m,n64_intrPI);
break;}
case n64_piWrLen:{
m->pi.Regs[n64_piWrLen] = v;
n64_Machine_piDMA(m,cast<uint32_t>((v + cast<uint32_t>(1ULL))));
break;}
case n64_piStatus:{
if ((cast<uint32_t>((v & n64_piStatusResetCmd)) != cast<uint32_t>(0ULL))) {
m->pi.Status = cast<uint32_t>(0ULL);
}
if ((cast<uint32_t>((v & n64_piStatusClearIntrCmd)) != cast<uint32_t>(0ULL))) {
n64_Machine_clearIRQ(m,n64_intrPI);
}
break;}
default:{
m->pi.Regs[cast<uint32_t>((addr & cast<uint32_t>(255ULL)))] = v;
break;}
}}
}
}
// tools/platform/n64/pi.go:89:1
void n64_Machine_piDMA(n64_Machine* m,uint32_t length){
{
auto tmp11 = std::make_tuple(m->pi.DramAddr,m->pi.CartAddr);
uint32_t dram = std::get<0>(tmp11);
uint32_t cart = std::get<1>(tmp11);
if (bool(m->OnDMA)) {
m->OnDMA("pi",dram,cart,length);
}
for (uint32_t i = cast<uint32_t>(cast<uint32_t>(0ULL));(i < length);i++){
uint32_t d = cast<uint32_t>((dram + i));
if ((cast<int64_t>(d) >= len(m->RDRAM))) {
n64_Machine_note(m,"PI: DMA of %d bytes to 0x%08X runs past RDRAM (truncated)",length,dram);
break;
}
uint32_t src = cast<uint32_t>((cart + i));
uint8_t b={};
{
auto tmp12 = n64_Machine_backing(m,src);
Slice<uint8_t> s = std::get<0>(tmp12);
uint32_t off = std::get<1>(tmp12);
if (bool(s)) {
b = s[off];
}
else {
n64_Machine_note(m,"PI: DMA source 0x%08X is not the cartridge (reading 0)",src);
}
}
m->RDRAM[d] = b;
}
m->pi.DramAddr = cast<uint32_t>((dram + length));
m->pi.CartAddr = cast<uint32_t>((cart + length));
n64_Machine_raiseIRQ(m,n64_intrPI);
}
}
// tools/platform/n64/rdp.go:72:1
int64_t n64_cmdLen(uint32_t op){
{
{
int64_t n = n64_cmdWords[cast<uint32_t>((op & cast<uint32_t>(63ULL)))];
if ((n != cast<int64_t>(0ULL))) {
return n;
}
}
return cast<int64_t>(1ULL);
}
}
// tools/platform/n64/rdp.go:178:1
uint32_t n64_rdp_cycleType(n64_rdp* r){
{
return cast<uint32_t>((cast<uint32_t>(shr<uint64_t>(r->OtherModes,cast<int64_t>(52ULL))) & cast<uint32_t>(3ULL)));
}
}
// tools/platform/n64/rdp.go:182:1
Slice<uint8_t> n64_Machine_RDPTMem(n64_Machine* m){
{
Slice<uint8_t> t = Slice<uint8_t>::make(len(m->rdp.TMem));
gcopy(t,sub(m->rdp.TMem,0,len(m->rdp.TMem)));
return t;
}
}
// tools/platform/n64/rdp.go:199:1
void n64_Machine_runRDP(n64_Machine* m){
{
auto tmp13 = std::make_tuple(get(m->dp,n64_dpCurrent),get(m->dp,n64_dpEnd));
uint32_t start = std::get<0>(tmp13);
uint32_t end = std::get<1>(tmp13);
if ((start < get(m->dp,n64_dpStart))) {
start = get(m->dp,n64_dpStart);
}
if ((end <= start)) {
return ;
}
bool xbus = (cast<uint32_t>((get(m->dp,n64_dpStatus) & n64_dpStatusXBusDMEM)) != cast<uint32_t>(0ULL));
std::function<uint64_t(uint32_t)> read64 = [&](uint32_t addr)->uint64_t{
if (xbus) {
uint32_t a = cast<uint32_t>((addr & cast<uint32_t>(4088ULL)));
return be_Uint64(sub(m->DMEM,a,len(m->DMEM)));
}
if ((cast<int64_t>((cast<int64_t>(addr) + cast<int64_t>(8ULL))) > len(m->RDRAM))) {
return cast<uint64_t>(0ULL);
}
return be_Uint64(sub(m->RDRAM,addr,len(m->RDRAM)));
}
;
n64_rdp* r = (&m->rdp);
for (uint32_t addr = start;(addr < end);){
r->Pending = append(r->Pending,read64(addr));
addr += cast<uint32_t>(8ULL);
m->dp[n64_dpCurrent] = addr;
m->rdpWords++;
uint32_t op = cast<uint32_t>(cast<uint64_t>((shr<uint64_t>(r->Pending[cast<int64_t>(0ULL)],cast<int64_t>(56ULL)) & cast<uint64_t>(63ULL))));
if ((len(r->Pending) < n64_cmdLen(op))) {
continue;
}
Slice<uint64_t> words = r->Pending;
r->Pending = {};
if (bool(m->OnRDPCmd)) {
m->OnRDPCmd(m,op,words);
}
n64_Machine_execRDP(m,op,words);
if (m->CPU->Halted) {
return ;
}
if ((m->rdpStopAt != cast<int64_t>(0ULL))) {
m->rdpCount++;
if ((m->rdpCount >= m->rdpStopAt)) {
throw n64_rdpStop{};
}
}
}
}
}
// tools/platform/n64/rdp.go:292:1
std::string n64_cmdNameOf(uint32_t op){
{
{
auto tmp14 = lookup(n64_cmdName,op);
std::string n = std::get<0>(tmp14);
bool ok = std::get<1>(tmp14);
if (ok) {
return n;
}
}
return "unknown";
}
}
// tools/platform/n64/rdp.go:303:1
std::string n64_RDPName(uint32_t op){
{
return n64_cmdNameOf(op);
}
}
// tools/platform/n64/rdp.go:306:1
void n64_Machine_execRDP(n64_Machine* m,uint32_t op,Slice<uint64_t> w){
{
n64_rdp* r = (&m->rdp);
{
switch(op){
case n64_cmdNoOp:{
break;}
case n64_cmdSetColorImage:{
r->Color = n64_decodeImage(w[cast<int64_t>(0ULL)]);
break;}
case n64_cmdSetTextureImage:{
r->Texture = n64_decodeImage(w[cast<int64_t>(0ULL)]);
break;}
case n64_cmdSetMaskImage:{
r->Mask = cast<uint32_t>((cast<uint32_t>(w[cast<int64_t>(0ULL)]) & cast<uint32_t>(67108863ULL)));
break;}
case n64_cmdSetScissor:{
r->Scissor.XH = cast<uint32_t>(cast<uint64_t>((shr<uint64_t>(w[cast<int64_t>(0ULL)],cast<int64_t>(44ULL)) & cast<uint64_t>(4095ULL))));
r->Scissor.YH = cast<uint32_t>(cast<uint64_t>((shr<uint64_t>(w[cast<int64_t>(0ULL)],cast<int64_t>(32ULL)) & cast<uint64_t>(4095ULL))));
r->Scissor.XL = cast<uint32_t>(cast<uint64_t>((shr<uint64_t>(w[cast<int64_t>(0ULL)],cast<int64_t>(12ULL)) & cast<uint64_t>(4095ULL))));
r->Scissor.YL = cast<uint32_t>(cast<uint64_t>((w[cast<int64_t>(0ULL)] & cast<uint64_t>(4095ULL))));
break;}
case n64_cmdSetOtherModes:{
r->OtherModes = cast<uint64_t>((w[cast<int64_t>(0ULL)] & cast<uint64_t>(72057594037927935ULL)));
break;}
case n64_cmdSetCombineMode:{
r->Combine = cast<uint64_t>((w[cast<int64_t>(0ULL)] & cast<uint64_t>(72057594037927935ULL)));
break;}
case n64_cmdSetFillColor:{
r->FillColor = cast<uint32_t>(w[cast<int64_t>(0ULL)]);
break;}
case n64_cmdSetFogColor:{
r->FogColor = cast<uint32_t>(w[cast<int64_t>(0ULL)]);
break;}
case n64_cmdSetBlendColor:{
r->BlendColor = cast<uint32_t>(w[cast<int64_t>(0ULL)]);
break;}
case n64_cmdSetPrimColor:{
r->PrimColor = cast<uint32_t>(w[cast<int64_t>(0ULL)]);
break;}
case n64_cmdSetEnvColor:{
r->EnvColor = cast<uint32_t>(w[cast<int64_t>(0ULL)]);
break;}
case n64_cmdSetPrimDepth:{
r->PrimDepth = cast<uint32_t>(cast<uint64_t>((shr<uint64_t>(w[cast<int64_t>(0ULL)],cast<int64_t>(16ULL)) & cast<uint64_t>(65535ULL))));
break;}
case n64_cmdSetConvert:case n64_cmdSetKeyGB:case n64_cmdSetKeyR:{
break;}
case n64_cmdSetTile:{
r->Tiles[cast<uint64_t>((shr<uint64_t>(w[cast<int64_t>(0ULL)],cast<int64_t>(24ULL)) & cast<uint64_t>(7ULL)))] = n64_decodeTile(w[cast<int64_t>(0ULL)]);
break;}
case n64_cmdSetTileSize:{
n64_tile* t = (&r->Tiles[cast<uint64_t>((shr<uint64_t>(w[cast<int64_t>(0ULL)],cast<int64_t>(24ULL)) & cast<uint64_t>(7ULL)))]);
t->SL = cast<uint32_t>(cast<uint64_t>((shr<uint64_t>(w[cast<int64_t>(0ULL)],cast<int64_t>(44ULL)) & cast<uint64_t>(4095ULL))));
t->TL = cast<uint32_t>(cast<uint64_t>((shr<uint64_t>(w[cast<int64_t>(0ULL)],cast<int64_t>(32ULL)) & cast<uint64_t>(4095ULL))));
t->SH = cast<uint32_t>(cast<uint64_t>((shr<uint64_t>(w[cast<int64_t>(0ULL)],cast<int64_t>(12ULL)) & cast<uint64_t>(4095ULL))));
t->TH = cast<uint32_t>(cast<uint64_t>((w[cast<int64_t>(0ULL)] & cast<uint64_t>(4095ULL))));
break;}
case n64_cmdLoadBlock:{
n64_Machine_loadBlock(m,w[cast<int64_t>(0ULL)]);
break;}
case n64_cmdLoadTile:{
n64_Machine_loadTile(m,w[cast<int64_t>(0ULL)]);
break;}
case n64_cmdLoadTLUT:{
n64_Machine_loadTLUT(m,w[cast<int64_t>(0ULL)]);
break;}
case n64_cmdSyncLoad:case n64_cmdSyncPipe:case n64_cmdSyncTile:{
break;}
case n64_cmdSyncFull:{
n64_Machine_raiseIRQ(m,n64_intrDP);
break;}
case n64_cmdFillRect:{
n64_Machine_fillRect(m,w[cast<int64_t>(0ULL)]);
break;}
case n64_cmdTexRect:case n64_cmdTexRectFlip:{
n64_Machine_texRect(m,w,(op == n64_cmdTexRectFlip));
break;}
case n64_cmdTriFill:case n64_cmdTriFillZ:case n64_cmdTriTex:case n64_cmdTriTexZ:case n64_cmdTriShade:case n64_cmdTriShadeZ:case n64_cmdTriShadeTex:case n64_cmdTriShadeTexZ:{
n64_Machine_triangle(m,op,w);
break;}
default:{
r4300_CPU_Halt(m->CPU,"unmodelled RDP command 0x%02X (%s) at DPC 0x%08X",op,n64_cmdNameOf(op),get(m->dp,n64_dpCurrent));
break;}
}}
}
}
// tools/platform/n64/rdp.go:380:1
n64_rdpImage n64_decodeImage(uint64_t w){
{
return n64_rdpImage{cast<uint32_t>(cast<uint64_t>((shr<uint64_t>(w,cast<int64_t>(53ULL)) & cast<uint64_t>(7ULL)))),cast<uint32_t>(cast<uint64_t>((shr<uint64_t>(w,cast<int64_t>(51ULL)) & cast<uint64_t>(3ULL)))),cast<uint32_t>((cast<uint32_t>(cast<uint64_t>((shr<uint64_t>(w,cast<int64_t>(32ULL)) & cast<uint64_t>(1023ULL)))) + cast<uint32_t>(1ULL))),cast<uint32_t>((cast<uint32_t>(w) & cast<uint32_t>(67108863ULL)))};
}
}
// tools/platform/n64/rdp.go:389:1
n64_tile n64_decodeTile(uint64_t w){
{
return n64_tile{cast<uint32_t>(cast<uint64_t>((shr<uint64_t>(w,cast<int64_t>(53ULL)) & cast<uint64_t>(7ULL)))),cast<uint32_t>(cast<uint64_t>((shr<uint64_t>(w,cast<int64_t>(51ULL)) & cast<uint64_t>(3ULL)))),cast<uint32_t>(cast<uint64_t>((shr<uint64_t>(w,cast<int64_t>(41ULL)) & cast<uint64_t>(511ULL)))),cast<uint32_t>(cast<uint64_t>((shr<uint64_t>(w,cast<int64_t>(32ULL)) & cast<uint64_t>(511ULL)))),cast<uint32_t>(cast<uint64_t>((shr<uint64_t>(w,cast<int64_t>(20ULL)) & cast<uint64_t>(15ULL)))),cast<uint32_t>(cast<uint64_t>((shr<uint64_t>(w,cast<int64_t>(8ULL)) & cast<uint64_t>(3ULL)))),cast<uint32_t>(cast<uint64_t>((shr<uint64_t>(w,cast<int64_t>(18ULL)) & cast<uint64_t>(3ULL)))),cast<uint32_t>(cast<uint64_t>((shr<uint64_t>(w,cast<int64_t>(4ULL)) & cast<uint64_t>(15ULL)))),cast<uint32_t>(cast<uint64_t>((shr<uint64_t>(w,cast<int64_t>(14ULL)) & cast<uint64_t>(15ULL)))),cast<uint32_t>(cast<uint64_t>((w & cast<uint64_t>(15ULL)))),cast<uint32_t>(cast<uint64_t>((shr<uint64_t>(w,cast<int64_t>(10ULL)) & cast<uint64_t>(15ULL)))),{},{},{},{}};
}
}
// tools/platform/n64/rdp_combine.go:40:1
n64_uint32Vec3 n64_combineInputs_subA(n64_combineInputs* in,uint32_t sel){
{
{
switch(sel){
case cast<uint32_t>(0ULL):{
return n64_vec3(in->Comb);
break;}
case cast<uint32_t>(1ULL):{
return n64_vec3(in->Texel0);
break;}
case cast<uint32_t>(2ULL):{
return n64_vec3(in->Texel1);
break;}
case cast<uint32_t>(3ULL):{
return n64_vec3(in->Prim);
break;}
case cast<uint32_t>(4ULL):{
return n64_vec3(in->Shade);
break;}
case cast<uint32_t>(5ULL):{
return n64_vec3(in->Env);
break;}
case cast<uint32_t>(6ULL):{
return n64_uint32Vec3{cast<uint32_t>(255ULL),cast<uint32_t>(255ULL),cast<uint32_t>(255ULL)};
break;}
case cast<uint32_t>(7ULL):{
return n64_uint32Vec3{cast<uint32_t>(0ULL),cast<uint32_t>(0ULL),cast<uint32_t>(0ULL)};
break;}
}}
return n64_uint32Vec3{};
}
}
// tools/platform/n64/rdp_combine.go:62:1
n64_uint32Vec3 n64_combineInputs_subB(n64_combineInputs* in,uint32_t sel){
{
{
switch(sel){
case cast<uint32_t>(0ULL):{
return n64_vec3(in->Comb);
break;}
case cast<uint32_t>(1ULL):{
return n64_vec3(in->Texel0);
break;}
case cast<uint32_t>(2ULL):{
return n64_vec3(in->Texel1);
break;}
case cast<uint32_t>(3ULL):{
return n64_vec3(in->Prim);
break;}
case cast<uint32_t>(4ULL):{
return n64_vec3(in->Shade);
break;}
case cast<uint32_t>(5ULL):{
return n64_vec3(in->Env);
break;}
case cast<uint32_t>(6ULL):{
return n64_vec3(in->Key);
break;}
case cast<uint32_t>(7ULL):{
return n64_uint32Vec3{};
break;}
}}
return n64_uint32Vec3{};
}
}
// tools/platform/n64/rdp_combine.go:84:1
n64_uint32Vec3 n64_combineInputs_mul(n64_combineInputs* in,uint32_t sel){
{
{
switch(sel){
case cast<uint32_t>(0ULL):{
return n64_vec3(in->Comb);
break;}
case cast<uint32_t>(1ULL):{
return n64_vec3(in->Texel0);
break;}
case cast<uint32_t>(2ULL):{
return n64_vec3(in->Texel1);
break;}
case cast<uint32_t>(3ULL):{
return n64_vec3(in->Prim);
break;}
case cast<uint32_t>(4ULL):{
return n64_vec3(in->Shade);
break;}
case cast<uint32_t>(5ULL):{
return n64_vec3(in->Env);
break;}
case cast<uint32_t>(6ULL):{
return n64_vec3(in->Key);
break;}
case cast<uint32_t>(7ULL):{
return n64_splat(in->Comb.A);
break;}
case cast<uint32_t>(8ULL):{
return n64_splat(in->Texel0.A);
break;}
case cast<uint32_t>(9ULL):{
return n64_splat(in->Texel1.A);
break;}
case cast<uint32_t>(10ULL):{
return n64_splat(in->Prim.A);
break;}
case cast<uint32_t>(11ULL):{
return n64_splat(in->Shade.A);
break;}
case cast<uint32_t>(12ULL):{
return n64_splat(in->Env.A);
break;}
case cast<uint32_t>(13ULL):{
return n64_splat(in->LODFrac);
break;}
case cast<uint32_t>(14ULL):{
return n64_splat(in->PrimLODFrac);
break;}
case cast<uint32_t>(15ULL):{
return n64_uint32Vec3{};
break;}
}}
return n64_uint32Vec3{};
}
}
// tools/platform/n64/rdp_combine.go:122:1
n64_uint32Vec3 n64_combineInputs_add(n64_combineInputs* in,uint32_t sel){
{
{
switch(sel){
case cast<uint32_t>(0ULL):{
return n64_vec3(in->Comb);
break;}
case cast<uint32_t>(1ULL):{
return n64_vec3(in->Texel0);
break;}
case cast<uint32_t>(2ULL):{
return n64_vec3(in->Texel1);
break;}
case cast<uint32_t>(3ULL):{
return n64_vec3(in->Prim);
break;}
case cast<uint32_t>(4ULL):{
return n64_vec3(in->Shade);
break;}
case cast<uint32_t>(5ULL):{
return n64_vec3(in->Env);
break;}
case cast<uint32_t>(6ULL):{
return n64_uint32Vec3{cast<uint32_t>(255ULL),cast<uint32_t>(255ULL),cast<uint32_t>(255ULL)};
break;}
}}
return n64_uint32Vec3{};
}
}
// tools/platform/n64/rdp_combine.go:145:1
uint32_t n64_combineInputs_alphaABD(n64_combineInputs* in,uint32_t sel){
{
{
switch(sel){
case cast<uint32_t>(0ULL):{
return in->Comb.A;
break;}
case cast<uint32_t>(1ULL):{
return in->Texel0.A;
break;}
case cast<uint32_t>(2ULL):{
return in->Texel1.A;
break;}
case cast<uint32_t>(3ULL):{
return in->Prim.A;
break;}
case cast<uint32_t>(4ULL):{
return in->Shade.A;
break;}
case cast<uint32_t>(5ULL):{
return in->Env.A;
break;}
case cast<uint32_t>(6ULL):{
return cast<uint32_t>(255ULL);
break;}
}}
return cast<uint32_t>(0ULL);
}
}
// tools/platform/n64/rdp_combine.go:165:1
uint32_t n64_combineInputs_alphaMul(n64_combineInputs* in,uint32_t sel){
{
{
switch(sel){
case cast<uint32_t>(0ULL):{
return in->LODFrac;
break;}
case cast<uint32_t>(1ULL):{
return in->Texel0.A;
break;}
case cast<uint32_t>(2ULL):{
return in->Texel1.A;
break;}
case cast<uint32_t>(3ULL):{
return in->Prim.A;
break;}
case cast<uint32_t>(4ULL):{
return in->Shade.A;
break;}
case cast<uint32_t>(5ULL):{
return in->Env.A;
break;}
case cast<uint32_t>(6ULL):{
return in->PrimLODFrac;
break;}
}}
return cast<uint32_t>(0ULL);
}
}
// tools/platform/n64/rdp_combine.go:187:1
n64_uint32Vec3 n64_vec3(n64_rgba c){
{
return n64_uint32Vec3{c.R,c.G,c.B};
}
}
// tools/platform/n64/rdp_combine.go:188:1
n64_uint32Vec3 n64_splat(uint32_t v){
{
return n64_uint32Vec3{v,v,v};
}
}
// tools/platform/n64/rdp_combine.go:192:1
uint32_t n64_combineChannel(uint32_t a,uint32_t b,uint32_t c,uint32_t d){
{
int32_t v = cast<int32_t>((cast<int32_t>(((cast<int32_t>((cast<int32_t>(a) - cast<int32_t>(b)))) * cast<int32_t>(c))) + cast<int32_t>(128ULL)));
v = cast<int32_t>(((shr<int32_t>(v,cast<int64_t>(8ULL))) + cast<int32_t>(d)));
if ((v < cast<int32_t>(0ULL))) {
return cast<uint32_t>(0ULL);
}
if ((v > cast<int32_t>(255ULL))) {
return cast<uint32_t>(255ULL);
}
return cast<uint32_t>(v);
}
}
// tools/platform/n64/rdp_combine.go:212:1
n64_combinerSelects n64_rdp_combinerSelects(n64_rdp* r,int64_t cycle){
{
uint64_t w = r->Combine;
if ((cycle == cast<int64_t>(0ULL))) {
return n64_combinerSelects{cast<uint32_t>(cast<uint64_t>((shr<uint64_t>(w,cast<int64_t>(52ULL)) & cast<uint64_t>(15ULL)))),cast<uint32_t>(cast<uint64_t>((shr<uint64_t>(w,cast<int64_t>(28ULL)) & cast<uint64_t>(15ULL)))),cast<uint32_t>(cast<uint64_t>((shr<uint64_t>(w,cast<int64_t>(47ULL)) & cast<uint64_t>(31ULL)))),cast<uint32_t>(cast<uint64_t>((shr<uint64_t>(w,cast<int64_t>(15ULL)) & cast<uint64_t>(7ULL)))),cast<uint32_t>(cast<uint64_t>((shr<uint64_t>(w,cast<int64_t>(44ULL)) & cast<uint64_t>(7ULL)))),cast<uint32_t>(cast<uint64_t>((shr<uint64_t>(w,cast<int64_t>(12ULL)) & cast<uint64_t>(7ULL)))),cast<uint32_t>(cast<uint64_t>((shr<uint64_t>(w,cast<int64_t>(41ULL)) & cast<uint64_t>(7ULL)))),cast<uint32_t>(cast<uint64_t>((shr<uint64_t>(w,cast<int64_t>(9ULL)) & cast<uint64_t>(7ULL))))};
}
return n64_combinerSelects{cast<uint32_t>(cast<uint64_t>((shr<uint64_t>(w,cast<int64_t>(37ULL)) & cast<uint64_t>(15ULL)))),cast<uint32_t>(cast<uint64_t>((shr<uint64_t>(w,cast<int64_t>(24ULL)) & cast<uint64_t>(15ULL)))),cast<uint32_t>(cast<uint64_t>((shr<uint64_t>(w,cast<int64_t>(32ULL)) & cast<uint64_t>(31ULL)))),cast<uint32_t>(cast<uint64_t>((shr<uint64_t>(w,cast<int64_t>(6ULL)) & cast<uint64_t>(7ULL)))),cast<uint32_t>(cast<uint64_t>((shr<uint64_t>(w,cast<int64_t>(21ULL)) & cast<uint64_t>(7ULL)))),cast<uint32_t>(cast<uint64_t>((shr<uint64_t>(w,cast<int64_t>(3ULL)) & cast<uint64_t>(7ULL)))),cast<uint32_t>(cast<uint64_t>((shr<uint64_t>(w,cast<int64_t>(18ULL)) & cast<uint64_t>(7ULL)))),cast<uint32_t>(cast<uint64_t>((w & cast<uint64_t>(7ULL))))};
}
}
// tools/platform/n64/rdp_combine.go:233:1
n64_rgba n64_rdp_combine(n64_rdp* r,n64_combineInputs* in){
{
int64_t cycles = cast<int64_t>(1ULL);
if ((n64_rdp_cycleType(r) == n64_cycle2)) {
cycles = cast<int64_t>(2ULL);
}
n64_rgba out={};
for (int64_t c = cast<int64_t>(0ULL);(c < cycles);c++){
n64_combinerSelects s = n64_rdp_combinerSelects(r,c);
auto tmp15 = std::make_tuple(n64_combineInputs_subA(in,s.subAC),n64_combineInputs_subB(in,s.subBC),n64_combineInputs_mul(in,s.mulC),n64_combineInputs_add(in,s.addC));
n64_uint32Vec3 a = std::get<0>(tmp15);
n64_uint32Vec3 b = std::get<1>(tmp15);
n64_uint32Vec3 m = std::get<2>(tmp15);
n64_uint32Vec3 d = std::get<3>(tmp15);
out = n64_rgba{n64_combineChannel(a.R,b.R,m.R,d.R),n64_combineChannel(a.G,b.G,m.G,d.G),n64_combineChannel(a.B,b.B,m.B,d.B),n64_combineChannel(n64_combineInputs_alphaABD(in,s.subAA),n64_combineInputs_alphaABD(in,s.subBA),n64_combineInputs_alphaMul(in,s.mulA),n64_combineInputs_alphaABD(in,s.addA))};
in->Comb = out;
}
return out;
}
}
// tools/platform/n64/rdp_combine.go:270:1
std::tuple<uint32_t,uint32_t,uint32_t,uint32_t> n64_rdp_blenderSelects(n64_rdp* r,int64_t cycle){
uint32_t p{};
uint32_t a{};
uint32_t m{};
uint32_t b{};
{
uint64_t w = r->OtherModes;
if ((cycle == cast<int64_t>(0ULL))) {
return {cast<uint32_t>(cast<uint64_t>((shr<uint64_t>(w,cast<int64_t>(30ULL)) & cast<uint64_t>(3ULL)))),cast<uint32_t>(cast<uint64_t>((shr<uint64_t>(w,cast<int64_t>(26ULL)) & cast<uint64_t>(3ULL)))),cast<uint32_t>(cast<uint64_t>((shr<uint64_t>(w,cast<int64_t>(22ULL)) & cast<uint64_t>(3ULL)))),cast<uint32_t>(cast<uint64_t>((shr<uint64_t>(w,cast<int64_t>(18ULL)) & cast<uint64_t>(3ULL))))};
}
return {cast<uint32_t>(cast<uint64_t>((shr<uint64_t>(w,cast<int64_t>(28ULL)) & cast<uint64_t>(3ULL)))),cast<uint32_t>(cast<uint64_t>((shr<uint64_t>(w,cast<int64_t>(24ULL)) & cast<uint64_t>(3ULL)))),cast<uint32_t>(cast<uint64_t>((shr<uint64_t>(w,cast<int64_t>(20ULL)) & cast<uint64_t>(3ULL)))),cast<uint32_t>(cast<uint64_t>((shr<uint64_t>(w,cast<int64_t>(16ULL)) & cast<uint64_t>(3ULL))))};
}
}
// tools/platform/n64/rdp_combine.go:285:1
n64_rgba n64_rdp_blend(n64_rdp* r,n64_rgba cyc,n64_rgba mem,uint32_t shadeAlpha){
{
int64_t cycle = cast<int64_t>(0ULL);
if ((n64_rdp_cycleType(r) == n64_cycle2)) {
cycle = cast<int64_t>(1ULL);
}
auto tmp16 = n64_rdp_blenderSelects(r,cycle);
uint32_t pSel = std::get<0>(tmp16);
uint32_t aSel = std::get<1>(tmp16);
uint32_t mSel = std::get<2>(tmp16);
uint32_t bSel = std::get<3>(tmp16);
std::function<n64_rgba(uint32_t)> pick = [&](uint32_t sel)->n64_rgba{
{
switch(sel){
case cast<uint32_t>(0ULL):{
return cyc;
break;}
case cast<uint32_t>(1ULL):{
return mem;
break;}
case cast<uint32_t>(2ULL):{
return n64_rgba{cast<uint32_t>((shr<uint32_t>(r->BlendColor,cast<int64_t>(24ULL)) & cast<uint32_t>(255ULL))),cast<uint32_t>((shr<uint32_t>(r->BlendColor,cast<int64_t>(16ULL)) & cast<uint32_t>(255ULL))),cast<uint32_t>((shr<uint32_t>(r->BlendColor,cast<int64_t>(8ULL)) & cast<uint32_t>(255ULL))),cast<uint32_t>((r->BlendColor & cast<uint32_t>(255ULL)))};
break;}
default:{
return n64_rgba{cast<uint32_t>((shr<uint32_t>(r->FogColor,cast<int64_t>(24ULL)) & cast<uint32_t>(255ULL))),cast<uint32_t>((shr<uint32_t>(r->FogColor,cast<int64_t>(16ULL)) & cast<uint32_t>(255ULL))),cast<uint32_t>((shr<uint32_t>(r->FogColor,cast<int64_t>(8ULL)) & cast<uint32_t>(255ULL))),cast<uint32_t>((r->FogColor & cast<uint32_t>(255ULL)))};
break;}
}}
}
;
auto tmp17 = std::make_tuple(pick(pSel),pick(mSel));
n64_rgba p = std::get<0>(tmp17);
n64_rgba m = std::get<1>(tmp17);
uint32_t a={};
{
switch(aSel){
case cast<uint32_t>(0ULL):{
a = cyc.A;
break;}
case cast<uint32_t>(1ULL):{
a = cast<uint32_t>((r->FogColor & cast<uint32_t>(255ULL)));
break;}
case cast<uint32_t>(2ULL):{
a = shadeAlpha;
break;}
default:{
a = cast<uint32_t>(0ULL);
break;}
}}
uint32_t b={};
{
switch(bSel){
case cast<uint32_t>(0ULL):{
b = cast<uint32_t>((cast<uint32_t>(255ULL) - a));
break;}
case cast<uint32_t>(1ULL):{
b = mem.A;
break;}
case cast<uint32_t>(2ULL):{
b = cast<uint32_t>(255ULL);
break;}
default:{
b = cast<uint32_t>(0ULL);
break;}
}}
std::function<uint32_t(uint32_t,uint32_t)> mix = [&](uint32_t pc,uint32_t mc)->uint32_t{
uint32_t num = cast<uint32_t>((cast<uint32_t>((pc * a)) + cast<uint32_t>((mc * b))));
uint32_t den = cast<uint32_t>((a + b));
if ((den == cast<uint32_t>(0ULL))) {
return pc;
}
if ((cast<uint64_t>((r->OtherModes & n64_omForceBlend)) != cast<uint64_t>(0ULL))) {
den = cast<uint32_t>(255ULL);
}
uint32_t v = divi<uint32_t>(num,den);
if ((v > cast<uint32_t>(255ULL))) {
return cast<uint32_t>(255ULL);
}
return v;
}
;
return n64_rgba{mix(p.R,m.R),mix(p.G,m.G),mix(p.B,m.B),cyc.A};
}
}
// tools/platform/n64/rdp_raster.go:16:1
uint32_t n64_rdp_pixelAddr(n64_rdp* r,uint32_t x,uint32_t y){
{
{
switch(r->Color.Size){
case n64_size16:{
return cast<uint32_t>((r->Color.Addr + cast<uint32_t>(((cast<uint32_t>((cast<uint32_t>((y * r->Color.Width)) + x))) * cast<uint32_t>(2ULL)))));
break;}
case n64_size32:{
return cast<uint32_t>((r->Color.Addr + cast<uint32_t>(((cast<uint32_t>((cast<uint32_t>((y * r->Color.Width)) + x))) * cast<uint32_t>(4ULL)))));
break;}
case n64_size8:{
return cast<uint32_t>((cast<uint32_t>((r->Color.Addr + cast<uint32_t>((y * r->Color.Width)))) + x));
break;}
}}
return r->Color.Addr;
}
}
// tools/platform/n64/rdp_raster.go:29:1
uint16_t n64_rgba5551(uint32_t rr,uint32_t gg,uint32_t bb,uint32_t aa){
{
return cast<uint16_t>((cast<uint16_t>((cast<uint16_t>((shl<uint16_t>(cast<uint16_t>(shr<uint32_t>(rr,cast<int64_t>(3ULL))),cast<int64_t>(11ULL)) | shl<uint16_t>(cast<uint16_t>(shr<uint32_t>(gg,cast<int64_t>(3ULL))),cast<int64_t>(6ULL)))) | shl<uint16_t>(cast<uint16_t>(shr<uint32_t>(bb,cast<int64_t>(3ULL))),cast<int64_t>(1ULL)))) | cast<uint16_t>(cast<uint32_t>((shr<uint32_t>(aa,cast<int64_t>(7ULL)) & cast<uint32_t>(1ULL))))));
}
}
// tools/platform/n64/rdp_raster.go:34:1
void n64_Machine_writePixel(n64_Machine* m,uint32_t x,uint32_t y,uint32_t rr,uint32_t gg,uint32_t bb,uint32_t aa){
{
n64_rdp* r = (&m->rdp);
uint32_t a = n64_rdp_pixelAddr(&(m->rdp),x,y);
{
switch(r->Color.Size){
case n64_size16:{
n64_Machine_storeRDRAM16(m,a,n64_rgba5551(rr,gg,bb,aa));
break;}
case n64_size32:{
n64_Machine_storeRDRAM32(m,a,cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((shl<uint32_t>(rr,cast<int64_t>(24ULL)) | shl<uint32_t>(gg,cast<int64_t>(16ULL)))) | shl<uint32_t>(bb,cast<int64_t>(8ULL)))) | aa)));
break;}
case n64_size8:{
n64_Machine_storeRDRAM8(m,a,cast<uint8_t>(rr));
break;}
default:{
r4300_CPU_Halt(m->CPU,"unmodelled colour image size %d",r->Color.Size);
break;}
}}
}
}
// tools/platform/n64/rdp_raster.go:49:1
void n64_Machine_storeRDRAM8(n64_Machine* m,uint32_t a,uint8_t v){
{
if ((cast<int64_t>(a) < len(m->RDRAM))) {
m->RDRAM[a] = v;
}
}
}
// tools/platform/n64/rdp_raster.go:55:1
void n64_Machine_storeRDRAM16(n64_Machine* m,uint32_t a,uint16_t v){
{
if ((cast<int64_t>((cast<int64_t>(a) + cast<int64_t>(1ULL))) < len(m->RDRAM))) {
m->RDRAM[a] = cast<uint8_t>(shr<uint16_t>(v,cast<int64_t>(8ULL)));
m->RDRAM[cast<uint32_t>((a + cast<uint32_t>(1ULL)))] = cast<uint8_t>(v);
}
}
}
// tools/platform/n64/rdp_raster.go:62:1
void n64_Machine_storeRDRAM32(n64_Machine* m,uint32_t a,uint32_t v){
{
if ((cast<int64_t>((cast<int64_t>(a) + cast<int64_t>(3ULL))) < len(m->RDRAM))) {
m->RDRAM[a] = cast<uint8_t>(shr<uint32_t>(v,cast<int64_t>(24ULL)));
m->RDRAM[cast<uint32_t>((a + cast<uint32_t>(1ULL)))] = cast<uint8_t>(shr<uint32_t>(v,cast<int64_t>(16ULL)));
m->RDRAM[cast<uint32_t>((a + cast<uint32_t>(2ULL)))] = cast<uint8_t>(shr<uint32_t>(v,cast<int64_t>(8ULL)));
m->RDRAM[cast<uint32_t>((a + cast<uint32_t>(3ULL)))] = cast<uint8_t>(v);
}
}
}
// tools/platform/n64/rdp_raster.go:72:1
std::tuple<uint32_t,uint32_t,uint32_t,uint32_t,bool> n64_rdp_clip(n64_rdp* r,uint32_t xh,uint32_t yh,uint32_t xl,uint32_t yl){
{
auto tmp18 = std::make_tuple(shr<uint32_t>(r->Scissor.XH,cast<int64_t>(2ULL)),shr<uint32_t>(r->Scissor.YH,cast<int64_t>(2ULL)));
uint32_t sxh = std::get<0>(tmp18);
uint32_t syh = std::get<1>(tmp18);
auto tmp19 = std::make_tuple(shr<uint32_t>(r->Scissor.XL,cast<int64_t>(2ULL)),shr<uint32_t>(r->Scissor.YL,cast<int64_t>(2ULL)));
uint32_t sxl = std::get<0>(tmp19);
uint32_t syl = std::get<1>(tmp19);
if ((xh < sxh)) {
xh = sxh;
}
if ((yh < syh)) {
yh = syh;
}
if ((xl > sxl)) {
xl = sxl;
}
if ((yl > syl)) {
yl = syl;
}
if ((xl > r->Color.Width)) {
xl = r->Color.Width;
}
return {xh,yh,xl,yl,((xh < xl) && (yh < yl))};
}
}
// tools/platform/n64/rdp_raster.go:99:1
void n64_Machine_fillRect(n64_Machine* m,uint64_t w){
{
n64_rdp* r = (&m->rdp);
{
uint32_t ct = n64_rdp_cycleType(r);
if (((ct != n64_cycleFill) && (ct != n64_cycleCopy))) {
r4300_CPU_Halt(m->CPU,"unmodelled Fill_Rectangle in cycle type %d (only FILL and COPY are modelled)",ct);
return ;
}
}
uint32_t xl = shr<uint32_t>(cast<uint32_t>(cast<uint64_t>((shr<uint64_t>(w,cast<int64_t>(44ULL)) & cast<uint64_t>(4095ULL)))),cast<int64_t>(2ULL));
uint32_t yl = shr<uint32_t>(cast<uint32_t>(cast<uint64_t>((shr<uint64_t>(w,cast<int64_t>(32ULL)) & cast<uint64_t>(4095ULL)))),cast<int64_t>(2ULL));
uint32_t xh = shr<uint32_t>(cast<uint32_t>(cast<uint64_t>((shr<uint64_t>(w,cast<int64_t>(12ULL)) & cast<uint64_t>(4095ULL)))),cast<int64_t>(2ULL));
uint32_t yh = shr<uint32_t>(cast<uint32_t>(cast<uint64_t>((w & cast<uint64_t>(4095ULL)))),cast<int64_t>(2ULL));
auto tmp20 = n64_rdp_clip(r,xh,yh,cast<uint32_t>((xl + cast<uint32_t>(1ULL))),cast<uint32_t>((yl + cast<uint32_t>(1ULL))));
xh = std::get<0>(tmp20);
yh = std::get<1>(tmp20);
xl = std::get<2>(tmp20);
yl = std::get<3>(tmp20);
bool ok = std::get<4>(tmp20);
if ((!ok)) {
return ;
}
for (uint32_t y = yh;(y < yl);y++){
for (uint32_t x = xh;(x < xl);x++){
{
switch(r->Color.Size){
case n64_size16:{
uint16_t v = cast<uint16_t>(shr<uint32_t>(r->FillColor,cast<int64_t>(16ULL)));
if ((cast<uint32_t>((x & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
v = cast<uint16_t>(r->FillColor);
}
n64_Machine_storeRDRAM16(m,n64_rdp_pixelAddr(r,x,y),v);
if (bool(m->OnPixel)) {
n64_rgba c = n64_fromRGBA16(v);
m->OnPixel(x,y,n64_PixelEvent{true,{},{},c.R,c.G,c.B,c.A,{},{},{},{},{},{},{}});
}
break;}
case n64_size32:{
n64_Machine_storeRDRAM32(m,n64_rdp_pixelAddr(r,x,y),r->FillColor);
if (bool(m->OnPixel)) {
m->OnPixel(x,y,n64_PixelEvent{true,{},{},cast<uint32_t>((shr<uint32_t>(r->FillColor,cast<int64_t>(24ULL)) & cast<uint32_t>(255ULL))),cast<uint32_t>((shr<uint32_t>(r->FillColor,cast<int64_t>(16ULL)) & cast<uint32_t>(255ULL))),cast<uint32_t>((shr<uint32_t>(r->FillColor,cast<int64_t>(8ULL)) & cast<uint32_t>(255ULL))),cast<uint32_t>((r->FillColor & cast<uint32_t>(255ULL))),{},{},{},{},{},{},{}});
}
break;}
default:{
r4300_CPU_Halt(m->CPU,"unmodelled Fill_Rectangle into a %d-bit colour image",shl<int64_t>(cast<int64_t>(4ULL),r->Color.Size));
return ;
break;}
}}
}
}
}
}
// tools/platform/n64/rdp_raster.go:148:1
void n64_Machine_texRect(n64_Machine* m,Slice<uint64_t> w,bool flip){
{
n64_rdp* r = (&m->rdp);
uint32_t ct = n64_rdp_cycleType(r);
if ((ct == n64_cycleFill)) {
r4300_CPU_Halt(m->CPU,"Texture_Rectangle in FILL mode, which the hardware does not define");
return ;
}
uint32_t xl = shr<uint32_t>(cast<uint32_t>(cast<uint64_t>((shr<uint64_t>(w[cast<int64_t>(0ULL)],cast<int64_t>(44ULL)) & cast<uint64_t>(4095ULL)))),cast<int64_t>(2ULL));
uint32_t yl = shr<uint32_t>(cast<uint32_t>(cast<uint64_t>((shr<uint64_t>(w[cast<int64_t>(0ULL)],cast<int64_t>(32ULL)) & cast<uint64_t>(4095ULL)))),cast<int64_t>(2ULL));
uint32_t tileIdx = cast<uint32_t>(cast<uint64_t>((shr<uint64_t>(w[cast<int64_t>(0ULL)],cast<int64_t>(24ULL)) & cast<uint64_t>(7ULL))));
uint32_t xh = shr<uint32_t>(cast<uint32_t>(cast<uint64_t>((shr<uint64_t>(w[cast<int64_t>(0ULL)],cast<int64_t>(12ULL)) & cast<uint64_t>(4095ULL)))),cast<int64_t>(2ULL));
uint32_t yh = shr<uint32_t>(cast<uint32_t>(cast<uint64_t>((w[cast<int64_t>(0ULL)] & cast<uint64_t>(4095ULL)))),cast<int64_t>(2ULL));
int32_t s0 = cast<int32_t>(cast<int16_t>(cast<uint16_t>(shr<uint64_t>(w[cast<int64_t>(1ULL)],cast<int64_t>(48ULL)))));
int32_t t0 = cast<int32_t>(cast<int16_t>(cast<uint16_t>(shr<uint64_t>(w[cast<int64_t>(1ULL)],cast<int64_t>(32ULL)))));
int32_t dsdx = cast<int32_t>(cast<int16_t>(cast<uint16_t>(shr<uint64_t>(w[cast<int64_t>(1ULL)],cast<int64_t>(16ULL)))));
int32_t dtdy = cast<int32_t>(cast<int16_t>(cast<uint16_t>(w[cast<int64_t>(1ULL)])));
if (flip) {
auto tmp21 = std::make_tuple(dtdy,dsdx);
dsdx = std::get<0>(tmp21);
dtdy = std::get<1>(tmp21);
}
if ((ct == n64_cycleCopy)) {
dsdx = shr<int32_t>(dsdx,cast<int64_t>(2ULL));
}
auto tmp22 = n64_rdp_clip(r,xh,yh,cast<uint32_t>((xl + cast<uint32_t>(1ULL))),cast<uint32_t>((yl + cast<uint32_t>(1ULL))));
uint32_t cxh = std::get<0>(tmp22);
uint32_t cyh = std::get<1>(tmp22);
uint32_t cxl = std::get<2>(tmp22);
uint32_t cyl = std::get<3>(tmp22);
bool ok = std::get<4>(tmp22);
if ((!ok)) {
return ;
}
n64_tile* tile = (&r->Tiles[tileIdx]);
n64_rgba prim = n64_unpackColor(r->PrimColor);
n64_rgba env = n64_unpackColor(r->EnvColor);
for (uint32_t y = cyh;(y < cyl);y++){
int32_t tv = cast<int32_t>((t0 + divi<int32_t>(cast<int32_t>((dtdy * cast<int32_t>(cast<uint32_t>((y - yh))))),cast<int32_t>(32ULL))));
for (uint32_t x = cxh;(x < cxl);x++){
int32_t sv = cast<int32_t>((s0 + divi<int32_t>(cast<int32_t>((dsdx * cast<int32_t>(cast<uint32_t>((x - xh))))),cast<int32_t>(32ULL))));
auto tmp23 = n64_Machine_sample(m,tile,sv,tv);
n64_rgba texel = std::get<0>(tmp23);
bool ok = std::get<1>(tmp23);
if ((!ok)) {
return ;
}
if ((ct == n64_cycleCopy)) {
if (((cast<uint64_t>((r->OtherModes & n64_omAlphaCompare)) != cast<uint64_t>(0ULL)) && (texel.A == cast<uint32_t>(0ULL)))) {
if (bool(m->OnPixel)) {
m->OnPixel(x,y,n64_PixelEvent{{},{},true,{},{},{},{},{},{},{},{},{},{},{}});
}
continue;
}
n64_Machine_writePixel(m,x,y,texel.R,texel.G,texel.B,texel.A);
if (bool(m->OnPixel)) {
m->OnPixel(x,y,n64_PixelEvent{true,{},{},texel.R,texel.G,texel.B,texel.A,{},texel.R,texel.G,texel.B,texel.A,{},{}});
}
continue;
}
n64_combineInputs in = n64_combineInputs{texel,texel,n64_rgba{cast<uint32_t>(255ULL),cast<uint32_t>(255ULL),cast<uint32_t>(255ULL),cast<uint32_t>(255ULL)},prim,env,{},{},{},{},sv,tv};
n64_Machine_drawPixel(m,x,y,(&in),cast<int64_t>(0ULL),false);
if (m->CPU->Halted) {
return ;
}
}
}
}
}
// tools/platform/n64/rdp_raster.go:235:1
void n64_Machine_loadBlock(n64_Machine* m,uint64_t w){
{
n64_rdp* r = (&m->rdp);
uint32_t tileIdx = cast<uint32_t>(cast<uint64_t>((shr<uint64_t>(w,cast<int64_t>(24ULL)) & cast<uint64_t>(7ULL))));
uint32_t sl = cast<uint32_t>(cast<uint64_t>((shr<uint64_t>(w,cast<int64_t>(44ULL)) & cast<uint64_t>(4095ULL))));
uint32_t tl = cast<uint32_t>(cast<uint64_t>((shr<uint64_t>(w,cast<int64_t>(32ULL)) & cast<uint64_t>(4095ULL))));
uint32_t sh = cast<uint32_t>(cast<uint64_t>((shr<uint64_t>(w,cast<int64_t>(12ULL)) & cast<uint64_t>(4095ULL))));
uint32_t dxt = cast<uint32_t>(cast<uint64_t>((w & cast<uint64_t>(4095ULL))));
n64_tile* t = (&r->Tiles[tileIdx]);
uint32_t bpt = n64_texelBytes(r->Texture.Size);
if ((bpt == cast<uint32_t>(0ULL))) {
r4300_CPU_Halt(m->CPU,"unmodelled Load_Block from a %d-bit texture image",shl<int64_t>(cast<int64_t>(4ULL),r->Texture.Size));
return ;
}
uint32_t src = cast<uint32_t>((r->Texture.Addr + cast<uint32_t>(((cast<uint32_t>((cast<uint32_t>((tl * r->Texture.Width)) + sl))) * bpt))));
uint32_t n = cast<uint32_t>(((cast<uint32_t>((cast<uint32_t>((sh - sl)) + cast<uint32_t>(1ULL)))) * bpt));
uint32_t dst = cast<uint32_t>((t->TMem * cast<uint32_t>(8ULL)));
for (uint32_t i = cast<uint32_t>(cast<uint32_t>(0ULL));((i < n) && (cast<uint32_t>((dst + i)) < cast<uint32_t>(4096ULL)));i++){
if ((cast<int64_t>(cast<uint32_t>((src + i))) >= len(m->RDRAM))) {
continue;
}
uint32_t d = cast<uint32_t>((dst + i));
if ((dxt != cast<uint32_t>(0ULL))) {
uint32_t word = divi<uint32_t>(i,cast<uint32_t>(8ULL));
if ((cast<uint32_t>((shr<uint32_t>(cast<uint32_t>((word * dxt)),cast<int64_t>(11ULL)) & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL))) {
d = cast<uint32_t>((cast<uint32_t>((dst + ((i & ~(cast<uint32_t>(7ULL)))))) + (cast<uint32_t>(((cast<uint32_t>((i & cast<uint32_t>(7ULL)))) ^ cast<uint32_t>(4ULL))))));
}
}
if ((d < cast<uint32_t>(4096ULL))) {
r->TMem[d] = m->RDRAM[cast<uint32_t>((src + i))];
}
}
}
}
// tools/platform/n64/rdp_raster.go:277:1
void n64_Machine_loadTile(n64_Machine* m,uint64_t w){
{
n64_rdp* r = (&m->rdp);
uint32_t tileIdx = cast<uint32_t>(cast<uint64_t>((shr<uint64_t>(w,cast<int64_t>(24ULL)) & cast<uint64_t>(7ULL))));
uint32_t sl = shr<uint32_t>(cast<uint32_t>(cast<uint64_t>((shr<uint64_t>(w,cast<int64_t>(44ULL)) & cast<uint64_t>(4095ULL)))),cast<int64_t>(2ULL));
uint32_t tl = shr<uint32_t>(cast<uint32_t>(cast<uint64_t>((shr<uint64_t>(w,cast<int64_t>(32ULL)) & cast<uint64_t>(4095ULL)))),cast<int64_t>(2ULL));
uint32_t sh = shr<uint32_t>(cast<uint32_t>(cast<uint64_t>((shr<uint64_t>(w,cast<int64_t>(12ULL)) & cast<uint64_t>(4095ULL)))),cast<int64_t>(2ULL));
uint32_t th = shr<uint32_t>(cast<uint32_t>(cast<uint64_t>((w & cast<uint64_t>(4095ULL)))),cast<int64_t>(2ULL));
n64_tile* t = (&r->Tiles[tileIdx]);
uint32_t bpt = n64_texelBytes(r->Texture.Size);
if ((bpt == cast<uint32_t>(0ULL))) {
r4300_CPU_Halt(m->CPU,"unmodelled Load_Tile from a %d-bit texture image",shl<int64_t>(cast<int64_t>(4ULL),r->Texture.Size));
return ;
}
for (uint32_t row = tl;(row <= th);row++){
uint32_t src = cast<uint32_t>((r->Texture.Addr + cast<uint32_t>(((cast<uint32_t>((cast<uint32_t>((row * r->Texture.Width)) + sl))) * bpt))));
uint32_t dst = cast<uint32_t>((cast<uint32_t>((t->TMem * cast<uint32_t>(8ULL))) + cast<uint32_t>((cast<uint32_t>(((cast<uint32_t>((row - tl))) * t->Line)) * cast<uint32_t>(8ULL)))));
uint32_t n = cast<uint32_t>(((cast<uint32_t>((cast<uint32_t>((sh - sl)) + cast<uint32_t>(1ULL)))) * bpt));
for (uint32_t i = cast<uint32_t>(cast<uint32_t>(0ULL));((i < n) && (cast<uint32_t>((dst + i)) < cast<uint32_t>(4096ULL)));i++){
if ((cast<int64_t>(cast<uint32_t>((src + i))) >= len(m->RDRAM))) {
continue;
}
uint32_t d = cast<uint32_t>((dst + i));
if ((cast<uint32_t>(((cast<uint32_t>((row - tl))) & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL))) {
d = cast<uint32_t>((cast<uint32_t>((dst + ((i & ~(cast<uint32_t>(7ULL)))))) + (cast<uint32_t>(((cast<uint32_t>((i & cast<uint32_t>(7ULL)))) ^ cast<uint32_t>(4ULL))))));
}
if ((d < cast<uint32_t>(4096ULL))) {
r->TMem[d] = m->RDRAM[cast<uint32_t>((src + i))];
}
}
}
auto tmp24 = std::make_tuple(shl<uint32_t>(sl,cast<int64_t>(2ULL)),shl<uint32_t>(tl,cast<int64_t>(2ULL)),shl<uint32_t>(sh,cast<int64_t>(2ULL)),shl<uint32_t>(th,cast<int64_t>(2ULL)));
t->SL = std::get<0>(tmp24);
t->TL = std::get<1>(tmp24);
t->SH = std::get<2>(tmp24);
t->TH = std::get<3>(tmp24);
}
}
// tools/platform/n64/rdp_raster.go:316:1
void n64_Machine_loadTLUT(n64_Machine* m,uint64_t w){
{
n64_rdp* r = (&m->rdp);
uint32_t tileIdx = cast<uint32_t>(cast<uint64_t>((shr<uint64_t>(w,cast<int64_t>(24ULL)) & cast<uint64_t>(7ULL))));
uint32_t sl = shr<uint32_t>(cast<uint32_t>(cast<uint64_t>((shr<uint64_t>(w,cast<int64_t>(44ULL)) & cast<uint64_t>(4095ULL)))),cast<int64_t>(2ULL));
uint32_t sh = shr<uint32_t>(cast<uint32_t>(cast<uint64_t>((shr<uint64_t>(w,cast<int64_t>(12ULL)) & cast<uint64_t>(4095ULL)))),cast<int64_t>(2ULL));
n64_tile* t = (&r->Tiles[tileIdx]);
for (uint32_t i = sl;(i <= sh);i++){
uint32_t src = cast<uint32_t>((r->Texture.Addr + cast<uint32_t>((i * cast<uint32_t>(2ULL)))));
uint32_t dst = cast<uint32_t>(((cast<uint32_t>((cast<uint32_t>((t->TMem * cast<uint32_t>(8ULL))) + cast<uint32_t>(((cast<uint32_t>((i - sl))) * cast<uint32_t>(8ULL)))))) & cast<uint32_t>(4095ULL)));
if (((cast<int64_t>((cast<int64_t>(src) + cast<int64_t>(1ULL))) < len(m->RDRAM)) && (cast<uint32_t>((dst + cast<uint32_t>(1ULL))) < cast<uint32_t>(4096ULL)))) {
r->TMem[dst] = m->RDRAM[src];
r->TMem[cast<uint32_t>((dst + cast<uint32_t>(1ULL)))] = m->RDRAM[cast<uint32_t>((src + cast<uint32_t>(1ULL)))];
}
}
}
}
// tools/platform/n64/rdp_raster.go:334:1
uint32_t n64_texelBytes(uint32_t size){
{
{
switch(size){
case n64_size8:{
return cast<uint32_t>(1ULL);
break;}
case n64_size16:{
return cast<uint32_t>(2ULL);
break;}
case n64_size32:{
return cast<uint32_t>(4ULL);
break;}
}}
return cast<uint32_t>(0ULL);
}
}
// tools/platform/n64/rdp_texture.go:39:1
uint32_t n64_texCoord(int32_t v,uint32_t mask,uint32_t cm,uint32_t lo,uint32_t hi){
{
v -= cast<int32_t>(lo);
if (((cast<uint32_t>((cm & cast<uint32_t>(2ULL))) != cast<uint32_t>(0ULL)) || (mask == cast<uint32_t>(0ULL)))) {
if ((v < cast<int32_t>(0ULL))) {
v = cast<int32_t>(0ULL);
}
if (((hi > lo) && (v > cast<int32_t>(cast<uint32_t>((hi - lo)))))) {
v = cast<int32_t>(cast<uint32_t>((hi - lo)));
}
}
if ((mask == cast<uint32_t>(0ULL))) {
return cast<uint32_t>(v);
}
int32_t m = cast<int32_t>((shl<int32_t>(cast<int32_t>(cast<int32_t>(1ULL)),mask) - cast<int32_t>(1ULL)));
if (((cast<uint32_t>((cm & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL)) && (cast<int32_t>((v & (cast<int32_t>((m + cast<int32_t>(1ULL)))))) != cast<int32_t>(0ULL)))) {
return cast<uint32_t>(cast<int32_t>((m - (cast<int32_t>((v & m))))));
}
return cast<uint32_t>(cast<int32_t>((v & m)));
}
}
// tools/platform/n64/rdp_texture.go:60:1
uint8_t n64_rdp_tmem(n64_rdp* r,uint32_t off){
{
return r->TMem[cast<uint32_t>((off & cast<uint32_t>(4095ULL)))];
}
}
// tools/platform/n64/rdp_texture.go:62:1
uint16_t n64_rdp_tmem16(n64_rdp* r,uint32_t off){
{
return cast<uint16_t>((shl<uint16_t>(cast<uint16_t>(n64_rdp_tmem(r,off)),cast<int64_t>(8ULL)) | cast<uint16_t>(n64_rdp_tmem(r,cast<uint32_t>((off + cast<uint32_t>(1ULL)))))));
}
}
// tools/platform/n64/rdp_texture.go:68:1
uint16_t n64_rdp_tlut(n64_rdp* r,uint32_t i){
{
return n64_rdp_tmem16(r,cast<uint32_t>((n64_tlutBase + cast<uint32_t>((i * cast<uint32_t>(8ULL))))));
}
}
// tools/platform/n64/rdp_texture.go:70:1
n64_rgba n64_fromRGBA16(uint16_t v){
{
return n64_rgba{shl<uint32_t>(cast<uint32_t>(cast<uint16_t>((shr<uint16_t>(v,cast<int64_t>(11ULL)) & cast<uint16_t>(31ULL)))),cast<int64_t>(3ULL)),shl<uint32_t>(cast<uint32_t>(cast<uint16_t>((shr<uint16_t>(v,cast<int64_t>(6ULL)) & cast<uint16_t>(31ULL)))),cast<int64_t>(3ULL)),shl<uint32_t>(cast<uint32_t>(cast<uint16_t>((shr<uint16_t>(v,cast<int64_t>(1ULL)) & cast<uint16_t>(31ULL)))),cast<int64_t>(3ULL)),cast<uint32_t>((cast<uint32_t>(cast<uint16_t>((v & cast<uint16_t>(1ULL)))) * cast<uint32_t>(255ULL)))};
}
}
// tools/platform/n64/rdp_texture.go:90:1
std::tuple<n64_rgba,bool> n64_Machine_sample(n64_Machine* m,n64_tile* t,int32_t sFix,int32_t tFix){
{
auto tmp25 = std::make_tuple(shr<int32_t>(sFix,cast<int64_t>(5ULL)),shr<int32_t>(tFix,cast<int64_t>(5ULL)));
int32_t s = std::get<0>(tmp25);
int32_t tt = std::get<1>(tmp25);
if ((cast<uint64_t>((m->rdp.OtherModes & n64_omSampleType)) == cast<uint64_t>(0ULL))) {
return n64_Machine_texelAt(m,t,s,tt);
}
auto tmp26 = std::make_tuple(cast<uint32_t>(cast<int32_t>((sFix & cast<int32_t>(31ULL)))),cast<uint32_t>(cast<int32_t>((tFix & cast<int32_t>(31ULL)))));
uint32_t sf = std::get<0>(tmp26);
uint32_t tf = std::get<1>(tmp26);
auto tmp27 = n64_Machine_texelAt(m,t,s,tt);
n64_rgba c00 = std::get<0>(tmp27);
bool ok = std::get<1>(tmp27);
if ((!ok)) {
return {n64_rgba{},false};
}
auto tmp28 = n64_Machine_texelAt(m,t,cast<int32_t>((s + cast<int32_t>(1ULL))),tt);
n64_rgba c10 = std::get<0>(tmp28);
ok = std::get<1>(tmp28);
if ((!ok)) {
return {n64_rgba{},false};
}
auto tmp29 = n64_Machine_texelAt(m,t,s,cast<int32_t>((tt + cast<int32_t>(1ULL))));
n64_rgba c01 = std::get<0>(tmp29);
ok = std::get<1>(tmp29);
if ((!ok)) {
return {n64_rgba{},false};
}
if ((cast<uint32_t>((sf + tf)) < cast<uint32_t>(32ULL))) {
return {n64_tap3(c00,c10,c01,sf,tf),true};
}
auto tmp30 = n64_Machine_texelAt(m,t,cast<int32_t>((s + cast<int32_t>(1ULL))),cast<int32_t>((tt + cast<int32_t>(1ULL))));
n64_rgba c11 = std::get<0>(tmp30);
ok = std::get<1>(tmp30);
if ((!ok)) {
return {n64_rgba{},false};
}
return {n64_tap3(c11,c01,c10,cast<uint32_t>((cast<uint32_t>(32ULL) - sf)),cast<uint32_t>((cast<uint32_t>(32ULL) - tf))),true};
}
}
// tools/platform/n64/rdp_texture.go:123:1
n64_rgba n64_tap3(n64_rgba base,n64_rgba alongS,n64_rgba alongT,uint32_t sf,uint32_t tf){
{
std::function<uint32_t(uint32_t,uint32_t,uint32_t)> ch = [&](uint32_t b,uint32_t s,uint32_t t)->uint32_t{
int32_t v = cast<int32_t>((cast<int32_t>((cast<int32_t>(b) + divi<int32_t>(cast<int32_t>(((cast<int32_t>((cast<int32_t>(s) - cast<int32_t>(b)))) * cast<int32_t>(sf))),cast<int32_t>(32ULL)))) + divi<int32_t>(cast<int32_t>(((cast<int32_t>((cast<int32_t>(t) - cast<int32_t>(b)))) * cast<int32_t>(tf))),cast<int32_t>(32ULL))));
if ((v < cast<int32_t>(0ULL))) {
return cast<uint32_t>(0ULL);
}
if ((v > cast<int32_t>(255ULL))) {
return cast<uint32_t>(255ULL);
}
return cast<uint32_t>(v);
}
;
return n64_rgba{ch(base.R,alongS.R,alongT.R),ch(base.G,alongS.G,alongT.G),ch(base.B,alongS.B,alongT.B),ch(base.A,alongS.A,alongT.A)};
}
}
// tools/platform/n64/rdp_texture.go:152:1
uint32_t n64_swizzle(uint32_t off,uint32_t row){
{
if ((cast<uint32_t>((row & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL))) {
return cast<uint32_t>((off ^ cast<uint32_t>(4ULL)));
}
return off;
}
}
// tools/platform/n64/rdp_texture.go:160:1
std::tuple<n64_rgba,bool> n64_Machine_texelAt(n64_Machine* m,n64_tile* t,int32_t s,int32_t tt){
{
n64_rdp* r = (&m->rdp);
uint32_t sx = n64_texCoord(s,t->MaskS,t->CMS,shr<uint32_t>(t->SL,cast<int64_t>(2ULL)),shr<uint32_t>(t->SH,cast<int64_t>(2ULL)));
uint32_t ty = n64_texCoord(tt,t->MaskT,t->CMT,shr<uint32_t>(t->TL,cast<int64_t>(2ULL)),shr<uint32_t>(t->TH,cast<int64_t>(2ULL)));
uint32_t row = cast<uint32_t>((cast<uint32_t>((t->TMem * cast<uint32_t>(8ULL))) + cast<uint32_t>((cast<uint32_t>((ty * t->Line)) * cast<uint32_t>(8ULL)))));
{
if (((t->Format == n64_fmtRGBA) && (t->Size == n64_size16))){
return {n64_fromRGBA16(n64_rdp_tmem16(r,n64_swizzle(cast<uint32_t>((row + cast<uint32_t>((sx * cast<uint32_t>(2ULL))))),ty))),true};
}
else if (((t->Format == n64_fmtRGBA) && (t->Size == n64_size32))){
uint16_t lo = n64_rdp_tmem16(r,n64_swizzle(cast<uint32_t>((row + cast<uint32_t>((sx * cast<uint32_t>(2ULL))))),ty));
uint16_t hi = n64_rdp_tmem16(r,cast<uint32_t>((n64_swizzle(cast<uint32_t>((row + cast<uint32_t>((sx * cast<uint32_t>(2ULL))))),ty) + cast<uint32_t>(2048ULL))));
return {n64_rgba{cast<uint32_t>(shr<uint16_t>(lo,cast<int64_t>(8ULL))),cast<uint32_t>(cast<uint16_t>((lo & cast<uint16_t>(255ULL)))),cast<uint32_t>(shr<uint16_t>(hi,cast<int64_t>(8ULL))),cast<uint32_t>(cast<uint16_t>((hi & cast<uint16_t>(255ULL))))},true};
}
else if (((t->Format == n64_fmtI) && (t->Size == n64_size8))){
uint32_t i = cast<uint32_t>(n64_rdp_tmem(r,n64_swizzle(cast<uint32_t>((row + sx)),ty)));
return {n64_rgba{i,i,i,i},true};
}
else if (((t->Format == n64_fmtI) && (t->Size == n64_size4))){
uint8_t v = n64_rdp_tmem(r,n64_swizzle(cast<uint32_t>((row + divi<uint32_t>(sx,cast<uint32_t>(2ULL)))),ty));
uint32_t n = cast<uint32_t>(shr<uint8_t>(v,cast<int64_t>(4ULL)));
if ((cast<uint32_t>((sx & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
n = cast<uint32_t>(cast<uint8_t>((v & cast<uint8_t>(15ULL))));
}
uint32_t i = cast<uint32_t>((n * cast<uint32_t>(17ULL)));
return {n64_rgba{i,i,i,i},true};
}
else if (((t->Format == n64_fmtIA) && (t->Size == n64_size16))){
uint16_t v = n64_rdp_tmem16(r,n64_swizzle(cast<uint32_t>((row + cast<uint32_t>((sx * cast<uint32_t>(2ULL))))),ty));
uint32_t i = cast<uint32_t>(shr<uint16_t>(v,cast<int64_t>(8ULL)));
return {n64_rgba{i,i,i,cast<uint32_t>(cast<uint16_t>((v & cast<uint16_t>(255ULL))))},true};
}
else if (((t->Format == n64_fmtIA) && (t->Size == n64_size8))){
uint8_t v = n64_rdp_tmem(r,n64_swizzle(cast<uint32_t>((row + sx)),ty));
uint32_t i = cast<uint32_t>((cast<uint32_t>(shr<uint8_t>(v,cast<int64_t>(4ULL))) * cast<uint32_t>(17ULL)));
return {n64_rgba{i,i,i,cast<uint32_t>((cast<uint32_t>(cast<uint8_t>((v & cast<uint8_t>(15ULL)))) * cast<uint32_t>(17ULL)))},true};
}
else if (((t->Format == n64_fmtIA) && (t->Size == n64_size4))){
uint8_t v = n64_rdp_tmem(r,n64_swizzle(cast<uint32_t>((row + divi<uint32_t>(sx,cast<uint32_t>(2ULL)))),ty));
uint32_t n = cast<uint32_t>(shr<uint8_t>(v,cast<int64_t>(4ULL)));
if ((cast<uint32_t>((sx & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
n = cast<uint32_t>(cast<uint8_t>((v & cast<uint8_t>(15ULL))));
}
uint32_t i = cast<uint32_t>(((shr<uint32_t>(n,cast<int64_t>(1ULL))) * cast<uint32_t>(36ULL)));
uint32_t a = cast<uint32_t>(cast<uint32_t>(0ULL));
if ((cast<uint32_t>((n & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL))) {
a = cast<uint32_t>(255ULL);
}
return {n64_rgba{i,i,i,a},true};
}
else if (((t->Format == n64_fmtCI) && (t->Size == n64_size8))){
return {n64_fromRGBA16(n64_rdp_tlut(r,cast<uint32_t>(n64_rdp_tmem(r,n64_swizzle(cast<uint32_t>((row + sx)),ty))))),true};
}
else if (((t->Format == n64_fmtCI) && (t->Size == n64_size4))){
uint8_t v = n64_rdp_tmem(r,n64_swizzle(cast<uint32_t>((row + divi<uint32_t>(sx,cast<uint32_t>(2ULL)))),ty));
uint32_t n = cast<uint32_t>(shr<uint8_t>(v,cast<int64_t>(4ULL)));
if ((cast<uint32_t>((sx & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
n = cast<uint32_t>(cast<uint8_t>((v & cast<uint8_t>(15ULL))));
}
return {n64_fromRGBA16(n64_rdp_tlut(r,cast<uint32_t>((shl<uint32_t>(t->Palette,cast<int64_t>(4ULL)) | n)))),true};
}
}
tmp31:;
r4300_CPU_Halt(m->CPU,"unmodelled texture format %d size %d (%d bits)",t->Format,t->Size,shl<int64_t>(cast<int64_t>(4ULL),t->Size));
return {n64_rgba{},false};
}
}
// tools/platform/n64/rdp_texture.go:227:1
n64_rgba n64_Machine_readPixel(n64_Machine* m,uint32_t x,uint32_t y){
{
n64_rdp* r = (&m->rdp);
uint32_t a = n64_rdp_pixelAddr(r,x,y);
{
switch(r->Color.Size){
case n64_size16:{
if ((cast<int64_t>((cast<int64_t>(a) + cast<int64_t>(1ULL))) >= len(m->RDRAM))) {
return n64_rgba{};
}
return n64_fromRGBA16(cast<uint16_t>((shl<uint16_t>(cast<uint16_t>(m->RDRAM[a]),cast<int64_t>(8ULL)) | cast<uint16_t>(m->RDRAM[cast<uint32_t>((a + cast<uint32_t>(1ULL)))]))));
break;}
case n64_size32:{
if ((cast<int64_t>((cast<int64_t>(a) + cast<int64_t>(3ULL))) >= len(m->RDRAM))) {
return n64_rgba{};
}
return n64_rgba{cast<uint32_t>(m->RDRAM[a]),cast<uint32_t>(m->RDRAM[cast<uint32_t>((a + cast<uint32_t>(1ULL)))]),cast<uint32_t>(m->RDRAM[cast<uint32_t>((a + cast<uint32_t>(2ULL)))]),cast<uint32_t>(m->RDRAM[cast<uint32_t>((a + cast<uint32_t>(3ULL)))])};
break;}
}}
return n64_rgba{};
}
}
// tools/platform/n64/rdp_texture.go:246:1
uint32_t n64_Machine_depthAt(n64_Machine* m,uint32_t x,uint32_t y){
{
n64_rdp* r = (&m->rdp);
uint32_t a = cast<uint32_t>((r->Mask + cast<uint32_t>(((cast<uint32_t>((cast<uint32_t>((y * r->Color.Width)) + x))) * cast<uint32_t>(2ULL)))));
if ((cast<int64_t>((cast<int64_t>(a) + cast<int64_t>(1ULL))) >= len(m->RDRAM))) {
return cast<uint32_t>(65535ULL);
}
return cast<uint32_t>((shl<uint32_t>(cast<uint32_t>(m->RDRAM[a]),cast<int64_t>(8ULL)) | cast<uint32_t>(m->RDRAM[cast<uint32_t>((a + cast<uint32_t>(1ULL)))])));
}
}
// tools/platform/n64/rdp_texture.go:255:1
void n64_Machine_setDepth(n64_Machine* m,uint32_t x,uint32_t y,uint32_t z){
{
n64_rdp* r = (&m->rdp);
uint32_t a = cast<uint32_t>((r->Mask + cast<uint32_t>(((cast<uint32_t>((cast<uint32_t>((y * r->Color.Width)) + x))) * cast<uint32_t>(2ULL)))));
n64_Machine_storeRDRAM16(m,a,cast<uint16_t>(z));
}
}
// tools/platform/n64/rdp_texture.go:286:1
uint32_t n64_depthOf(int64_t z){
{
if ((z < cast<int64_t>(0ULL))) {
z = cast<int64_t>(0ULL);
}
uint32_t v = cast<uint32_t>(shr<int64_t>(z,cast<int64_t>(13ULL)));
if ((v > cast<uint32_t>(262143ULL))) {
v = cast<uint32_t>(262143ULL);
}
int64_t exp = cast<int64_t>(7ULL);
for (int64_t i = cast<int64_t>(1ULL);(i < cast<int64_t>(8ULL));i++){
if ((v < n64_zRanges[i].base)) {
exp = cast<int64_t>((i - cast<int64_t>(1ULL)));
break;
}
}
Anon18 r = n64_zRanges[exp];
uint32_t mantissa = shr<uint32_t>((cast<uint32_t>((v - r.base))),r.shift);
return shl<uint32_t>((cast<uint32_t>((shl<uint32_t>(cast<uint32_t>(exp),cast<int64_t>(11ULL)) | cast<uint32_t>((mantissa & cast<uint32_t>(2047ULL)))))),cast<int64_t>(2ULL));
}
}
// tools/platform/n64/rdp_texture.go:325:1
void n64_Machine_drawPixel(n64_Machine* m,uint32_t x,uint32_t y,n64_combineInputs* in,int64_t z,bool useZ){
{
n64_rdp* r = (&m->rdp);
if (((useZ && (cast<uint64_t>((r->OtherModes & n64_omZCompare)) != cast<uint64_t>(0ULL))) && (r->Mask != cast<uint32_t>(0ULL)))) {
uint32_t zv = n64_depthOf(z);
if ((zv >= n64_Machine_depthAt(m,x,y))) {
if (bool(m->OnPixel)) {
m->OnPixel(x,y,n64_PixelEvent{{},true,{},{},{},{},{},z,{},{},{},{},{},{}});
}
return ;
}
}
n64_rgba col = n64_rdp_combine(r,in);
if (((cast<uint64_t>((r->OtherModes & n64_omAlphaCompare)) != cast<uint64_t>(0ULL)) && (col.A < cast<uint32_t>((r->BlendColor & cast<uint32_t>(255ULL)))))) {
if (bool(m->OnPixel)) {
m->OnPixel(x,y,n64_PixelEvent{{},{},true,{},{},{},{},z,{},{},{},{},{},{}});
}
return ;
}
n64_rgba out = col;
if (((cast<uint64_t>((r->OtherModes & n64_omForceBlend)) != cast<uint64_t>(0ULL)) || n64_rdp_blenderReadsMemory(r))) {
out = n64_rdp_blend(r,col,n64_Machine_readPixel(m,x,y),in->Shade.A);
}
n64_Machine_writePixel(m,x,y,out.R,out.G,out.B,out.A);
if (((useZ && (cast<uint64_t>((r->OtherModes & n64_omZUpdate)) != cast<uint64_t>(0ULL))) && (r->Mask != cast<uint32_t>(0ULL)))) {
n64_Machine_setDepth(m,x,y,n64_depthOf(z));
}
if (bool(m->OnPixel)) {
m->OnPixel(x,y,n64_PixelEvent{true,{},{},out.R,out.G,out.B,out.A,z,in->Texel0.R,in->Texel0.G,in->Texel0.B,in->Texel0.A,in->texS,in->texT});
}
}
}
// tools/platform/n64/rdp_texture.go:369:1
bool n64_rdp_blenderReadsMemory(n64_rdp* r){
{
int64_t cycle = cast<int64_t>(0ULL);
if ((n64_rdp_cycleType(r) == n64_cycle2)) {
cycle = cast<int64_t>(1ULL);
}
auto tmp32 = n64_rdp_blenderSelects(r,cycle);
uint32_t p = std::get<0>(tmp32);
uint32_t mm = std::get<2>(tmp32);
uint32_t b = std::get<3>(tmp32);
return (((p == cast<uint32_t>(1ULL)) || (mm == cast<uint32_t>(1ULL))) || (b == cast<uint32_t>(1ULL)));
}
}
// tools/platform/n64/rdp_tri.go:56:1
int64_t n64_triAttrs_at(n64_triAttrs a,int64_t dy,int64_t dxPix){
{
return cast<int64_t>((cast<int64_t>((a.base + cast<int64_t>((a.de * dy)))) + cast<int64_t>((a.dx * dxPix))));
}
}
// tools/platform/n64/rdp_tri.go:59:1
int64_t n64_s32(uint32_t v){
{
return cast<int64_t>(cast<int32_t>(v));
}
}
// tools/platform/n64/rdp_tri.go:62:1
int32_t n64_s14(uint32_t v){
{
return shr<int32_t>(cast<int32_t>(shl<uint32_t>(v,cast<int64_t>(18ULL))),cast<int64_t>(18ULL));
}
}
// tools/platform/n64/rdp_tri.go:66:1
int64_t n64_pair(uint64_t intWord,uint64_t fracWord,uint64_t shift){
{
int64_t i = cast<int64_t>(cast<int16_t>(cast<uint16_t>(shr<uint64_t>(intWord,shift))));
int64_t f = cast<int64_t>(cast<uint16_t>(shr<uint64_t>(fracWord,shift)));
return cast<int64_t>((shl<int64_t>(i,cast<int64_t>(16ULL)) | f));
}
}
// tools/platform/n64/rdp_tri.go:73:1
void n64_Machine_triangle(n64_Machine* m,uint32_t op,Slice<uint64_t> w){
{
n64_rdp* r = (&m->rdp);
{
uint32_t ct = n64_rdp_cycleType(r);
if (((ct != n64_cycle1) && (ct != n64_cycle2))) {
r4300_CPU_Halt(m->CPU,"unmodelled triangle in cycle type %d",ct);
return ;
}
}
bool hasShade = (cast<uint32_t>((op & cast<uint32_t>(4ULL))) != cast<uint32_t>(0ULL));
bool hasTex = (cast<uint32_t>((op & cast<uint32_t>(2ULL))) != cast<uint32_t>(0ULL));
bool hasZ = (cast<uint32_t>((op & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL));
bool lft = (cast<uint64_t>((shr<uint64_t>(w[cast<int64_t>(0ULL)],cast<int64_t>(55ULL)) & cast<uint64_t>(1ULL))) != cast<uint64_t>(0ULL));
uint32_t tileIdx = cast<uint32_t>(cast<uint64_t>((shr<uint64_t>(w[cast<int64_t>(0ULL)],cast<int64_t>(48ULL)) & cast<uint64_t>(7ULL))));
int32_t yl = n64_s14(cast<uint32_t>(cast<uint64_t>((shr<uint64_t>(w[cast<int64_t>(0ULL)],cast<int64_t>(32ULL)) & cast<uint64_t>(16383ULL)))));
int32_t ym = n64_s14(cast<uint32_t>(cast<uint64_t>((shr<uint64_t>(w[cast<int64_t>(0ULL)],cast<int64_t>(16ULL)) & cast<uint64_t>(16383ULL)))));
int32_t yh = n64_s14(cast<uint32_t>(cast<uint64_t>((w[cast<int64_t>(0ULL)] & cast<uint64_t>(16383ULL)))));
auto tmp33 = std::make_tuple(n64_s32(cast<uint32_t>(shr<uint64_t>(w[cast<int64_t>(1ULL)],cast<int64_t>(32ULL)))),n64_s32(cast<uint32_t>(w[cast<int64_t>(1ULL)])));
int64_t xl = std::get<0>(tmp33);
int64_t dxldy = std::get<1>(tmp33);
auto tmp34 = std::make_tuple(n64_s32(cast<uint32_t>(shr<uint64_t>(w[cast<int64_t>(2ULL)],cast<int64_t>(32ULL)))),n64_s32(cast<uint32_t>(w[cast<int64_t>(2ULL)])));
int64_t xh = std::get<0>(tmp34);
int64_t dxhdy = std::get<1>(tmp34);
auto tmp35 = std::make_tuple(n64_s32(cast<uint32_t>(shr<uint64_t>(w[cast<int64_t>(3ULL)],cast<int64_t>(32ULL)))),n64_s32(cast<uint32_t>(w[cast<int64_t>(3ULL)])));
int64_t xm = std::get<0>(tmp35);
int64_t dxmdy = std::get<1>(tmp35);
int64_t next = cast<int64_t>(4ULL);
std::array<n64_triAttrs,4> sh={};
if (hasShade) {
Slice<uint64_t> b = sub(w,next,cast<int64_t>((next + cast<int64_t>(8ULL))));
for (uint64_t i = cast<uint64_t>(cast<uint64_t>(0ULL));(i < cast<uint64_t>(4ULL));i++){
uint64_t s = cast<uint64_t>((cast<uint64_t>(48ULL) - cast<uint64_t>((cast<uint64_t>(16ULL) * i))));
sh[i] = n64_triAttrs{n64_pair(b[cast<int64_t>(0ULL)],b[cast<int64_t>(2ULL)],s),n64_pair(b[cast<int64_t>(1ULL)],b[cast<int64_t>(3ULL)],s),n64_pair(b[cast<int64_t>(4ULL)],b[cast<int64_t>(6ULL)],s)};
}
next += cast<int64_t>(8ULL);
}
std::array<n64_triAttrs,3> tex={};
if (hasTex) {
Slice<uint64_t> b = sub(w,next,cast<int64_t>((next + cast<int64_t>(8ULL))));
for (uint64_t i = cast<uint64_t>(cast<uint64_t>(0ULL));(i < cast<uint64_t>(3ULL));i++){
uint64_t s = cast<uint64_t>((cast<uint64_t>(48ULL) - cast<uint64_t>((cast<uint64_t>(16ULL) * i))));
tex[i] = n64_triAttrs{n64_pair(b[cast<int64_t>(0ULL)],b[cast<int64_t>(2ULL)],s),n64_pair(b[cast<int64_t>(1ULL)],b[cast<int64_t>(3ULL)],s),n64_pair(b[cast<int64_t>(4ULL)],b[cast<int64_t>(6ULL)],s)};
}
next += cast<int64_t>(8ULL);
}
n64_triAttrs zz={};
if (hasZ) {
Slice<uint64_t> b = sub(w,next,cast<int64_t>((next + cast<int64_t>(2ULL))));
zz = n64_triAttrs{n64_s32(cast<uint32_t>(shr<uint64_t>(b[cast<int64_t>(0ULL)],cast<int64_t>(32ULL)))),n64_s32(cast<uint32_t>(b[cast<int64_t>(0ULL)])),n64_s32(cast<uint32_t>(shr<uint64_t>(b[cast<int64_t>(1ULL)],cast<int64_t>(32ULL))))};
next += cast<int64_t>(2ULL);
}
int64_t q0 = cast<int64_t>(yh);
{
int64_t lo = cast<int64_t>(r->Scissor.YH);
if ((q0 < lo)) {
q0 = lo;
}
}
int64_t q1 = cast<int64_t>(yl);
{
int64_t hi = cast<int64_t>(r->Scissor.YL);
if ((q1 > hi)) {
q1 = hi;
}
}
int64_t yhBase = (cast<int64_t>(yh) & ~(cast<int64_t>(3ULL)));
int64_t yhScan = shr<int64_t>(cast<int64_t>(yh),cast<int64_t>(2ULL));
int64_t ymQ = cast<int64_t>(ym);
int64_t sampLo = cast<int64_t>(r->Scissor.XH);
int64_t sampHi = cast<int64_t>(r->Scissor.XL);
{
int64_t hi = cast<int64_t>((cast<int64_t>(r->Color.Width) * cast<int64_t>(4ULL)));
if ((sampHi > hi)) {
sampHi = hi;
}
}
n64_tile* tile = (&r->Tiles[tileIdx]);
n64_rgba prim = n64_unpackColor(r->PrimColor);
n64_rgba env = n64_unpackColor(r->EnvColor);
for (int64_t q = q0;(q < q1);){
int64_t y = shr<int64_t>(q,cast<int64_t>(2ULL));
int64_t qEnd = cast<int64_t>(((cast<int64_t>((y + cast<int64_t>(1ULL)))) * cast<int64_t>(4ULL)));
if ((qEnd > q1)) {
qEnd = q1;
}
std::array<std::array<int64_t,2>,4> spans={};
int64_t n = cast<int64_t>(0ULL);
for (;(q < qEnd);q++){
int64_t major = cast<int64_t>((xh + divi<int64_t>(cast<int64_t>((dxhdy * (cast<int64_t>((q - yhBase))))),cast<int64_t>(4ULL))));
int64_t minor={};
if ((q < ymQ)) {
minor = cast<int64_t>((xm + divi<int64_t>(cast<int64_t>((dxmdy * (cast<int64_t>((q - yhBase))))),cast<int64_t>(4ULL))));
}
else {
minor = cast<int64_t>((xl + divi<int64_t>(cast<int64_t>((dxldy * (cast<int64_t>((q - ymQ))))),cast<int64_t>(4ULL))));
}
auto tmp36 = std::make_tuple(major,minor);
int64_t xs = std::get<0>(tmp36);
int64_t xe = std::get<1>(tmp36);
if ((!lft)) {
auto tmp37 = std::make_tuple(minor,major);
xs = std::get<0>(tmp37);
xe = std::get<1>(tmp37);
}
if ((xe <= xs)) {
continue;
}
auto tmp38 = std::make_tuple(n64_ceilQuarter(xs),n64_ceilQuarter(xe));
int64_t c0 = std::get<0>(tmp38);
int64_t c1 = std::get<1>(tmp38);
if ((c0 < sampLo)) {
c0 = sampLo;
}
if ((c1 > sampHi)) {
c1 = sampHi;
}
if ((c0 < c1)) {
spans[n] = std::array<int64_t,2>{c0,c1};
n++;
}
}
if ((n == cast<int64_t>(0ULL))) {
continue;
}
auto tmp39 = std::make_tuple(spans[cast<int64_t>(0ULL)][cast<int64_t>(0ULL)],spans[cast<int64_t>(0ULL)][cast<int64_t>(1ULL)]);
int64_t lo = std::get<0>(tmp39);
int64_t hi = std::get<1>(tmp39);
for (int64_t i = cast<int64_t>(1ULL);(i < n);i++){
if ((spans[i][cast<int64_t>(0ULL)] < lo)) {
lo = spans[i][cast<int64_t>(0ULL)];
}
if ((spans[i][cast<int64_t>(1ULL)] > hi)) {
hi = spans[i][cast<int64_t>(1ULL)];
}
}
int64_t dy = cast<int64_t>((y - yhScan));
int64_t majorPix = shr<int64_t>((cast<int64_t>((xh + cast<int64_t>((dxhdy * dy))))),cast<int64_t>(16ULL));
for (int64_t x = shr<int64_t>(lo,cast<int64_t>(2ULL));(x <= shr<int64_t>((cast<int64_t>((hi - cast<int64_t>(1ULL)))),cast<int64_t>(2ULL)));x++){
bool covered = false;
for (int64_t i = cast<int64_t>(0ULL);(i < n);i++){
if (((spans[i][cast<int64_t>(0ULL)] < cast<int64_t>((cast<int64_t>((x * cast<int64_t>(4ULL))) + cast<int64_t>(4ULL)))) && (spans[i][cast<int64_t>(1ULL)] > cast<int64_t>((x * cast<int64_t>(4ULL)))))) {
covered = true;
break;
}
}
if ((!covered)) {
continue;
}
int64_t dxPix = cast<int64_t>((x - majorPix));
n64_combineInputs in = n64_combineInputs{{},{},n64_rgba{cast<uint32_t>(255ULL),cast<uint32_t>(255ULL),cast<uint32_t>(255ULL),cast<uint32_t>(255ULL)},prim,env,{},{},{},{},{},{}};
if (hasShade) {
in.Shade = n64_rgba{n64_clamp8(shr<int64_t>(n64_triAttrs_at(sh[cast<int64_t>(0ULL)],dy,dxPix),cast<int64_t>(16ULL))),n64_clamp8(shr<int64_t>(n64_triAttrs_at(sh[cast<int64_t>(1ULL)],dy,dxPix),cast<int64_t>(16ULL))),n64_clamp8(shr<int64_t>(n64_triAttrs_at(sh[cast<int64_t>(2ULL)],dy,dxPix),cast<int64_t>(16ULL))),n64_clamp8(shr<int64_t>(n64_triAttrs_at(sh[cast<int64_t>(3ULL)],dy,dxPix),cast<int64_t>(16ULL)))};
}
if (hasTex) {
int64_t sv = n64_triAttrs_at(tex[cast<int64_t>(0ULL)],dy,dxPix);
int64_t tv = n64_triAttrs_at(tex[cast<int64_t>(1ULL)],dy,dxPix);
int64_t wv = n64_triAttrs_at(tex[cast<int64_t>(2ULL)],dy,dxPix);
auto tmp40 = std::make_tuple(cast<int32_t>(shr<int64_t>(sv,cast<int64_t>(16ULL))),cast<int32_t>(shr<int64_t>(tv,cast<int64_t>(16ULL))));
int32_t sFix = std::get<0>(tmp40);
int32_t tFix = std::get<1>(tmp40);
if (((cast<uint64_t>((r->OtherModes & n64_omPerspTex)) != cast<uint64_t>(0ULL)) && (wv > cast<int64_t>(0ULL)))) {
sFix = cast<int32_t>(divi<int64_t>(cast<int64_t>((sv * cast<int64_t>(32768ULL))),wv));
tFix = cast<int32_t>(divi<int64_t>(cast<int64_t>((tv * cast<int64_t>(32768ULL))),wv));
}
auto tmp41 = n64_Machine_sample(m,tile,sFix,tFix);
n64_rgba texel = std::get<0>(tmp41);
bool ok = std::get<1>(tmp41);
if ((!ok)) {
return ;
}
auto tmp42 = std::make_tuple(texel,texel);
in.Texel0 = std::get<0>(tmp42);
in.Texel1 = std::get<1>(tmp42);
auto tmp43 = std::make_tuple(sFix,tFix);
in.texS = std::get<0>(tmp43);
in.texT = std::get<1>(tmp43);
}
int64_t z = cast<int64_t>(cast<int64_t>(0ULL));
if (hasZ) {
z = n64_triAttrs_at(zz,dy,dxPix);
}
n64_Machine_drawPixel(m,cast<uint32_t>(x),cast<uint32_t>(y),(&in),z,hasZ);
if (m->CPU->Halted) {
return ;
}
}
}
}
}
// tools/platform/n64/rdp_tri.go:288:1
int64_t n64_ceilQuarter(int64_t v){
{
int64_t q = shr<int64_t>(v,cast<int64_t>(14ULL));
if ((cast<int64_t>((v & cast<int64_t>(16383ULL))) != cast<int64_t>(0ULL))) {
q++;
}
return q;
}
}
// tools/platform/n64/rdp_tri.go:299:1
uint32_t n64_clamp8(int64_t v){
{
if ((v < cast<int64_t>(0ULL))) {
return cast<uint32_t>(0ULL);
}
if ((v > cast<int64_t>(255ULL))) {
return cast<uint32_t>(255ULL);
}
return cast<uint32_t>(v);
}
}
// tools/platform/n64/rdp_tri.go:309:1
n64_rgba n64_unpackColor(uint32_t v){
{
return n64_rgba{cast<uint32_t>((shr<uint32_t>(v,cast<int64_t>(24ULL)) & cast<uint32_t>(255ULL))),cast<uint32_t>((shr<uint32_t>(v,cast<int64_t>(16ULL)) & cast<uint32_t>(255ULL))),cast<uint32_t>((shr<uint32_t>(v,cast<int64_t>(8ULL)) & cast<uint32_t>(255ULL))),cast<uint32_t>((v & cast<uint32_t>(255ULL)))};
}
}
// tools/platform/n64/run.go:15:1
std::string n64_Result_String(n64_Result r){
{
return go_fmt_Sprintf("stopped at 0x%08X after %d steps: %s",r.PC,r.Steps,r.Reason);
}
}
// tools/platform/n64/run.go:31:1
void n64_Machine_SetSpinDetect(n64_Machine* m,bool on){
{
m->noSpin = (!on);
}
}
// tools/platform/n64/run.go:34:1
void n64_Machine_SetBreakpoint(n64_Machine* m,uint32_t vaddr){
{
if ((!m->run.breakpoints)) {
m->run.breakpoints = Map<uint32_t,bool>{};
}
m->run.breakpoints[vaddr] = true;
}
}
// tools/platform/n64/run.go:42:1
void n64_Machine_ClearBreakpoints(n64_Machine* m){
{
m->run.breakpoints = {};
}
}
// tools/platform/n64/run.go:48:1
n64_Result n64_Machine_Run(n64_Machine* m,uint64_t maxSteps){
{
uint64_t steps={};
bool first = true;
Map<uint32_t,bool> spin = Map<uint32_t,bool>{};
constexpr int64_t spinWindow=1048576ULL;
uint64_t sinceReset={};
for (;(steps < maxSteps);){
uint32_t pc = cast<uint32_t>(m->CPU->PC);
if (m->StopRequested) {
m->StopRequested = false;
return n64_Result{steps,pc,"stop requested"};
}
if ((get(m->run.breakpoints,pc) && (!first))) {
return n64_Result{steps,pc,go_fmt_Sprintf("breakpoint at 0x%08X",pc)};
}
first = false;
if (bool(m->OnStep)) {
m->OnStep(m,pc);
}
n64_Machine_tickVI(m);
r4300_CPU_Interrupt(m->CPU,n64_Machine_irqPending(m));
if ((!m->noSpin)) {
spin[pc] = true;
sinceReset++;
if ((sinceReset >= spinWindow)) {
if ((len(spin) < cast<int64_t>(3ULL))) {
return n64_Result{steps,pc,"spin (tight loop)"};
}
spin = Map<uint32_t,bool>{};
sinceReset = cast<uint64_t>(0ULL);
}
}
r4300_CPU_Step(m->CPU);
steps++;
if (m->CPU->Halted) {
return n64_Result{steps,cast<uint32_t>(r4300_CPU_CurPC(m->CPU)),m->CPU->HaltReason};
}
}
return n64_Result{steps,cast<uint32_t>(m->CPU->PC),"step budget exhausted"};
}
}
// tools/platform/n64/si.go:81:1
void n64_si_init(n64_si* s){
{
s->Regs = n64_regFile{};
}
}
// tools/platform/n64/si.go:91:1
uint32_t n64_Machine_siRead(n64_Machine* m,uint32_t addr){
{
{
switch(cast<uint32_t>((addr & cast<uint32_t>(255ULL)))){
case n64_siDramAddr:{
return m->si.DramAddr;
break;}
case n64_siStatus:{
uint32_t v={};
if ((cast<uint32_t>((m->mi.Intr & n64_intrSI)) != cast<uint32_t>(0ULL))) {
v |= n64_siStatusInterrupt;
}
return v;
break;}
}}
return get(m->si.Regs,cast<uint32_t>((addr & cast<uint32_t>(255ULL))));
}
}
// tools/platform/n64/si.go:107:1
void n64_Machine_siWrite(n64_Machine* m,uint32_t addr,uint32_t v){
{
{
switch(cast<uint32_t>((addr & cast<uint32_t>(255ULL)))){
case n64_siDramAddr:{
m->si.DramAddr = cast<uint32_t>((v & cast<uint32_t>(16777215ULL)));
break;}
case n64_siWriteAddr:{
m->SIWrites++;
for (int64_t i = cast<int64_t>(0ULL);(i < n64_pifRAMSize);i++){
m->PIF[i] = n64_Machine_rdramRead(m,cast<uint32_t>((m->si.DramAddr + cast<uint32_t>(i))));
}
n64_Machine_joybus(m);
n64_Machine_raiseIRQ(m,n64_intrSI);
break;}
case n64_siReadAddr:{
m->SIReads++;
n64_Machine_joybus(m);
for (int64_t i = cast<int64_t>(0ULL);(i < n64_pifRAMSize);i++){
n64_Machine_rdramWrite(m,cast<uint32_t>((m->si.DramAddr + cast<uint32_t>(i))),m->PIF[i]);
}
n64_Machine_raiseIRQ(m,n64_intrSI);
break;}
case n64_siStatus:{
n64_Machine_clearIRQ(m,n64_intrSI);
break;}
default:{
m->si.Regs[cast<uint32_t>((addr & cast<uint32_t>(255ULL)))] = v;
break;}
}}
}
}
// tools/platform/n64/si.go:152:1
void n64_Machine_joybus(n64_Machine* m){
{
if ((cast<uint8_t>((m->PIF[cast<int64_t>((n64_pifRAMSize - cast<int64_t>(1ULL)))] & cast<uint8_t>(1ULL))) == cast<uint8_t>(0ULL))) {
return ;
}
int64_t ch = cast<int64_t>(0ULL);
for (int64_t i = cast<int64_t>(0ULL);(i < cast<int64_t>((n64_pifRAMSize - cast<int64_t>(1ULL))));){
uint8_t tx = m->PIF[i];
{
switch(tx){
case cast<uint8_t>(254ULL):{
return ;
break;}
case cast<uint8_t>(255ULL):case cast<uint8_t>(0ULL):{
if ((tx == cast<uint8_t>(0ULL))) {
ch++;
}
i++;
continue;
break;}
}}
if ((cast<int64_t>((i + cast<int64_t>(1ULL))) >= cast<int64_t>((n64_pifRAMSize - cast<int64_t>(1ULL))))) {
return ;
}
uint8_t rx = cast<uint8_t>((m->PIF[cast<int64_t>((i + cast<int64_t>(1ULL)))] & cast<uint8_t>(63ULL)));
int64_t cmdAt = cast<int64_t>((i + cast<int64_t>(2ULL)));
if ((cmdAt >= cast<int64_t>((n64_pifRAMSize - cast<int64_t>(1ULL))))) {
return ;
}
int64_t resAt = cast<int64_t>((cmdAt + cast<int64_t>(cast<uint8_t>((tx & cast<uint8_t>(63ULL))))));
if ((cast<int64_t>((resAt + cast<int64_t>(rx))) > cast<int64_t>((n64_pifRAMSize - cast<int64_t>(1ULL))))) {
return ;
}
n64_Machine_joybusChannel(m,ch,sub(m->PIF,cmdAt,resAt),sub(m->PIF,resAt,cast<int64_t>((resAt + cast<int64_t>(rx)))),cast<int64_t>((i + cast<int64_t>(1ULL))));
i = cast<int64_t>((resAt + cast<int64_t>(rx)));
ch++;
}
}
}
// tools/platform/n64/si.go:191:1
void n64_Machine_joybusChannel(n64_Machine* m,int64_t ch,Slice<uint8_t> cmd,Slice<uint8_t> res,int64_t rxAt){
{
if ((len(cmd) == cast<int64_t>(0ULL))) {
return ;
}
if ((ch >= cast<int64_t>(4ULL))) {
n64_Machine_joybusEEPROM(m,cmd,res,rxAt);
return ;
}
n64_Controller pad = m->Controllers[ch];
if ((!m->JoybusCmds)) {
m->JoybusCmds = Map<uint8_t,uint64_t>{};
}
m->JoybusCmds[cmd[cast<int64_t>(0ULL)]]++;
{
switch(cmd[cast<int64_t>(0ULL)]){
case n64_jbInfo:case n64_jbReset:{
if ((!pad.Present)) {
m->PIF[rxAt] |= cast<uint8_t>(128ULL);
return ;
}
if ((len(res) >= cast<int64_t>(3ULL))) {
res[cast<int64_t>(0ULL)] = cast<uint8_t>(shr<uint16_t>(n64_devController,cast<int64_t>(8ULL)));
res[cast<int64_t>(1ULL)] = cast<uint8_t>(cast<uint16_t>((n64_devController & cast<uint16_t>(255ULL))));
res[cast<int64_t>(2ULL)] = cast<uint8_t>(0ULL);
}
break;}
case n64_jbControllerState:{
m->ContPolls++;
if ((!pad.Present)) {
m->PIF[rxAt] |= cast<uint8_t>(128ULL);
return ;
}
if ((len(res) >= cast<int64_t>(4ULL))) {
res[cast<int64_t>(0ULL)] = cast<uint8_t>(shr<uint16_t>(pad.Buttons,cast<int64_t>(8ULL)));
res[cast<int64_t>(1ULL)] = cast<uint8_t>(pad.Buttons);
res[cast<int64_t>(2ULL)] = cast<uint8_t>(pad.StickX);
res[cast<int64_t>(3ULL)] = cast<uint8_t>(pad.StickY);
}
break;}
case n64_jbReadAccessory:case n64_jbWriteAccessory:{
m->PIF[rxAt] |= cast<uint8_t>(128ULL);
break;}
default:{
n64_Machine_note(m,"joybus: unmodelled command 0x%02X on controller channel %d",cmd[cast<int64_t>(0ULL)],ch);
m->PIF[rxAt] |= cast<uint8_t>(128ULL);
break;}
}}
}
}
// tools/platform/n64/sp.go:67:1
uint32_t n64_Machine_spRead(n64_Machine* m,uint32_t addr){
{
if (((addr >= n64_spPCBase) && (addr < n64_spPCEnd))) {
return m->spPC;
}
{
switch(cast<uint32_t>((addr & cast<uint32_t>(31ULL)))){
case n64_spStatus:{
return get(m->sp,n64_spStatus);
break;}
case n64_spDMABusy:{
return cast<uint32_t>((get(m->sp,n64_spStatus) & n64_spStatusDMABusy));
break;}
case n64_spDMAFull:{
return cast<uint32_t>((get(m->sp,n64_spStatus) & n64_spStatusDMAFull));
break;}
case n64_spSemaphore:{
uint32_t v = get(m->sp,n64_spSemaphore);
m->sp[n64_spSemaphore] = cast<uint32_t>(1ULL);
return v;
break;}
}}
return get(m->sp,cast<uint32_t>((addr & cast<uint32_t>(31ULL))));
}
}
// tools/platform/n64/sp.go:87:1
void n64_Machine_spWrite(n64_Machine* m,uint32_t addr,uint32_t v){
{
if (((addr >= n64_spPCBase) && (addr < n64_spPCEnd))) {
m->spPC = cast<uint32_t>((v & cast<uint32_t>(4092ULL)));
return ;
}
{
switch(cast<uint32_t>((addr & cast<uint32_t>(31ULL)))){
case n64_spMemAddr:case n64_spDramAddr:{
m->sp[cast<uint32_t>((addr & cast<uint32_t>(31ULL)))] = v;
break;}
case n64_spRdLen:{
n64_Machine_spDMA(m,v,false);
break;}
case n64_spWrLen:{
n64_Machine_spDMA(m,v,true);
break;}
case n64_spStatus:{
n64_Machine_spStatusWrite(m,v);
break;}
case n64_spSemaphore:{
m->sp[n64_spSemaphore] = cast<uint32_t>(0ULL);
break;}
default:{
m->sp[cast<uint32_t>((addr & cast<uint32_t>(31ULL)))] = v;
break;}
}}
}
}
// tools/platform/n64/sp.go:108:1
void n64_Machine_spStatusWrite(n64_Machine* m,uint32_t v){
{
uint32_t s = get(m->sp,n64_spStatus);
if ((cast<uint32_t>((v & n64_spWSetHalt)) != cast<uint32_t>(0ULL))) {
s |= n64_spStatusHalt;
}
if ((cast<uint32_t>((v & n64_spWClearBroke)) != cast<uint32_t>(0ULL))) {
s &= ~(n64_spStatusBroke);
}
if ((cast<uint32_t>((v & n64_spWClearIntr)) != cast<uint32_t>(0ULL))) {
n64_Machine_clearIRQ(m,n64_intrSP);
}
if ((cast<uint32_t>((v & n64_spWSetIntr)) != cast<uint32_t>(0ULL))) {
n64_Machine_raiseIRQ(m,n64_intrSP);
}
if ((cast<uint32_t>((v & n64_spWSetSStep)) != cast<uint32_t>(0ULL))) {
s |= n64_spStatusSingleStep;
}
if ((cast<uint32_t>((v & n64_spWClearSStep)) != cast<uint32_t>(0ULL))) {
s &= ~(n64_spStatusSingleStep);
}
if ((cast<uint32_t>((v & n64_spWSetIntrBreak)) != cast<uint32_t>(0ULL))) {
s |= n64_spStatusIntrBreak;
}
if ((cast<uint32_t>((v & n64_spWClearIntrBreak)) != cast<uint32_t>(0ULL))) {
s &= ~(n64_spStatusIntrBreak);
}
for (uint32_t i = cast<uint32_t>(cast<uint32_t>(0ULL));(i < cast<uint32_t>(8ULL));i++){
if ((cast<uint32_t>((v & (shl<uint32_t>(cast<uint32_t>(1ULL),(cast<uint32_t>((n64_spWSignalBase + cast<uint32_t>((cast<uint32_t>(2ULL) * i))))))))) != cast<uint32_t>(0ULL))) {
s &= ~(shl<uint32_t>(n64_spStatusSig0,i));
}
if ((cast<uint32_t>((v & (shl<uint32_t>(cast<uint32_t>(1ULL),(cast<uint32_t>((cast<uint32_t>((n64_spWSignalBase + cast<uint32_t>((cast<uint32_t>(2ULL) * i)))) + cast<uint32_t>(1ULL)))))))) != cast<uint32_t>(0ULL))) {
s |= shl<uint32_t>(n64_spStatusSig0,i);
}
}
bool start = ((cast<uint32_t>((v & n64_spWClearHalt)) != cast<uint32_t>(0ULL)) && (cast<uint32_t>((s & n64_spStatusHalt)) != cast<uint32_t>(0ULL)));
if (start) {
s &= ~(n64_spStatusHalt);
}
m->sp[n64_spStatus] = s;
if (start) {
n64_Machine_runRSP(m);
}
}
}
// tools/platform/n64/sp.go:162:1
void n64_Machine_runRSP(n64_Machine* m){
{
if (m->rspRunning) {
return ;
}
m->rspRunning = true;
auto tmp44=defer([&](){[&]()->void{
m->rspRunning = false;
}
();});
if ((!m->RSP)) {
m->RSP = rsp_NewCPU(m->DMEM,m->IMEM,m);
}
if (bool(m->OnRSPTask)) {
m->OnRSPTask(m,m->spPC);
}
rsp_CPU_Start(m->RSP,m->spPC);
uint64_t n = rsp_CPU_Run(m->RSP,n64_rspBudget);
m->rspSteps += n;
if ((!m->RSP->Broke)) {
if (m->RSP->Halted) {
r4300_CPU_Halt(m->CPU,"RSP halted: %s (after %d RSP instructions)",m->RSP->HaltReason,n);
return ;
}
r4300_CPU_Halt(m->CPU,"RSP ran %d instructions without a BREAK: the microcode is looping",n);
return ;
}
m->sp[n64_spStatus] |= cast<uint32_t>((n64_spStatusHalt | n64_spStatusBroke));
m->spPC = cast<uint32_t>((m->RSP->PC & cast<uint32_t>(4092ULL)));
if ((cast<uint32_t>((get(m->sp,n64_spStatus) & n64_spStatusIntrBreak)) != cast<uint32_t>(0ULL))) {
n64_Machine_raiseIRQ(m,n64_intrSP);
}
}
}
// tools/platform/n64/sp.go:204:1
uint32_t n64_Machine_ReadCop0(n64_Machine* m,uint32_t reg){
{
if ((reg < cast<uint32_t>(8ULL))) {
return n64_Machine_spRead(m,cast<uint32_t>((n64_spRegsBase + cast<uint32_t>((reg * cast<uint32_t>(4ULL))))));
}
return n64_Machine_dpRead(m,cast<uint32_t>((n64_dpRegBase + cast<uint32_t>(((cast<uint32_t>((reg - cast<uint32_t>(8ULL)))) * cast<uint32_t>(4ULL))))));
}
}
// tools/platform/n64/sp.go:211:1
void n64_Machine_WriteCop0(n64_Machine* m,uint32_t reg,uint32_t v){
{
if ((reg < cast<uint32_t>(8ULL))) {
n64_Machine_spWrite(m,cast<uint32_t>((n64_spRegsBase + cast<uint32_t>((reg * cast<uint32_t>(4ULL))))),v);
return ;
}
n64_Machine_dpWrite(m,cast<uint32_t>((n64_dpRegBase + cast<uint32_t>(((cast<uint32_t>((reg - cast<uint32_t>(8ULL)))) * cast<uint32_t>(4ULL))))),v);
}
}
// tools/platform/n64/sp.go:224:1
void n64_Machine_spDMA(n64_Machine* m,uint32_t lenReg,bool toRDRAM){
{
uint32_t length = cast<uint32_t>(((cast<uint32_t>((lenReg & cast<uint32_t>(4095ULL)))) + cast<uint32_t>(1ULL)));
uint32_t memAddr = cast<uint32_t>((get(m->sp,n64_spMemAddr) & cast<uint32_t>(8191ULL)));
uint32_t dramAddr = cast<uint32_t>((get(m->sp,n64_spDramAddr) & cast<uint32_t>(16777215ULL)));
Slice<uint8_t> mem = m->DMEM;
if ((memAddr >= n64_spMemSize)) {
mem = m->IMEM;
memAddr -= n64_spMemSize;
}
std::string kind = "sp-read";
if (toRDRAM) {
kind = "sp-write";
}
if (bool(m->OnDMA)) {
m->OnDMA(kind,dramAddr,memAddr,length);
}
for (uint32_t i = cast<uint32_t>(cast<uint32_t>(0ULL));(i < length);i++){
uint32_t d = cast<uint32_t>((dramAddr + i));
uint32_t s = modi<uint32_t>((cast<uint32_t>((memAddr + i))),n64_spMemSize);
if (toRDRAM) {
n64_Machine_rdramWrite(m,d,mem[s]);
}
else {
mem[s] = n64_Machine_rdramRead(m,d);
}
}
m->sp[n64_spDramAddr] = cast<uint32_t>((dramAddr + length));
m->sp[n64_spMemAddr] = cast<uint32_t>((get(m->sp,n64_spMemAddr) + length));
}
}
// tools/platform/n64/vi.go:61:1
void n64_vi_init(n64_vi* v){
{
v->Regs = n64_regFile{};
}
}
// tools/platform/n64/vi.go:63:1
uint32_t n64_Machine_viRead(n64_Machine* m,uint32_t addr){
{
{
switch(cast<uint32_t>((addr & cast<uint32_t>(255ULL)))){
case n64_viCurrent:{
uint32_t lines = get(m->vi.Regs,n64_viVSync);
if ((lines == cast<uint32_t>(0ULL))) {
lines = n64_halflinesPerField;
}
return modi<uint32_t>(cast<uint32_t>(divi<uint64_t>(cast<uint64_t>((m->vi.Acc * cast<uint64_t>(lines))),n64_stepsPerField)),lines);
break;}
}}
return get(m->vi.Regs,cast<uint32_t>((addr & cast<uint32_t>(255ULL))));
}
}
// tools/platform/n64/vi.go:78:1
void n64_Machine_viWrite(n64_Machine* m,uint32_t addr,uint32_t v){
{
{
switch(cast<uint32_t>((addr & cast<uint32_t>(255ULL)))){
case n64_viCurrent:{
n64_Machine_clearIRQ(m,n64_intrVI);
return ;
break;}
case n64_viOrigin:{
m->vi.Regs[n64_viOrigin] = cast<uint32_t>((v & cast<uint32_t>(16777215ULL)));
return ;
break;}
}}
m->vi.Regs[cast<uint32_t>((addr & cast<uint32_t>(255ULL)))] = v;
}
}
// tools/platform/n64/vi.go:92:1
void n64_Machine_tickVI(n64_Machine* m){
{
m->vi.Acc++;
if ((m->vi.Acc < n64_stepsPerField)) {
return ;
}
m->vi.Acc = cast<uint64_t>(0ULL);
m->vi.Current = get(m->vi.Regs,n64_viIntr);
n64_Machine_raiseIRQ(m,n64_intrVI);
if (bool(m->OnDisplay)) {
m->OnDisplay(m);
}
}
}
// tools/platform/n64/vi.go:107:1
uint32_t n64_Machine_Origin(n64_Machine* m){
{
return get(m->vi.Regs,n64_viOrigin);
}
}
// tools/platform/n64/vi.go:108:1
uint32_t n64_Machine_Width(n64_Machine* m){
{
return get(m->vi.Regs,n64_viWidth);
}
}
// tools/platform/n64/vi.go:111:1
uint32_t n64_Machine_PixelType(n64_Machine* m){
{
return cast<uint32_t>((get(m->vi.Regs,n64_viStatus) & cast<uint32_t>(3ULL)));
}
}
