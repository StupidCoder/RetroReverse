#include "../../../../browser/core/capture.h"
// RR_CAPTURE_INSTRUMENTED
#include "../../../../browser/core/profile.h"
#include "runtime.h"
struct arm60_Inst;
struct arm60_CPU;
struct arm60_Context;
struct arm60_CPUState;
struct threedo_AIF;
struct threedo_audioEvent;
struct threedo_Cel;
struct threedo_bitReader;
struct threedo_CvidMovie;
struct threedo_cvidVec;
struct threedo_CvidDecoder;
struct threedo_CelDraw;
struct threedo_PixelEvent;
struct threedo_MemRegion;
struct threedo_diskStream;
struct threedo_dirScan;
struct threedo_span;
struct threedo_heap;
struct threedo_gfxBitmap;
struct threedo_item;
struct threedo_KernelCall;
struct threedo_Machine;
struct threedo_timerWait;
struct threedo_armedMovie;
struct threedo_Volume;
struct threedo_Entry;
struct threedo_ProfileBucket;
struct threedo_ProfileCounter;
struct threedo_FrameProfile;
struct threedo_celCounters;
struct threedo_profState;
struct threedo_Result;
struct threedo_PadStep;
struct threedo_task;
struct threedo_Resource;
struct threedo_WrapNode;
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
struct Anon4;
struct Anon5;
struct Anon6;
struct Anon7;
struct Anon8;
struct Anon9;
using arm60_Flow=int64_t;
struct arm60_Inst{
uint32_t Addr{};
int64_t Len{};
std::string Mnem{};
std::string Text{};
arm60_Flow Flow{};
uint32_t Target{};
bool HasTarget{};
int64_t Cond{};
};
struct arm60_CPU{
std::array<uint32_t,16> R{};
bool N{};
bool Z{};
bool C{};
bool V{};
bool IRQDisable{};
bool FIQDisable{};
uint32_t Mode{};
std::array<uint32_t,6> bankR13{};
std::array<uint32_t,6> bankR14{};
std::array<uint32_t,6> bankSPSR{};
std::array<uint32_t,5> fiqR8_12{};
std::array<uint32_t,5> usrR8_12{};
std::function<bool(arm60_CPU*,uint32_t)> SWI{};
threedo_Machine* bus{};
bool Halted{};
std::string HaltReason{};
uint64_t Instrs{};
uint32_t cur{};
bool branched{};
};
struct arm60_Context{
std::array<uint32_t,16> R{};
bool N{};
bool Z{};
bool C{};
bool V{};
bool IRQDisable{};
bool FIQDisable{};
uint32_t Mode{};
std::array<uint32_t,6> BankR13{};
std::array<uint32_t,6> BankR14{};
std::array<uint32_t,6> BankSPSR{};
std::array<uint32_t,5> FIQR8_12{};
std::array<uint32_t,5> USRR8_12{};
};
struct arm60_CPUState{
arm60_Context Ctx{};
bool Halted{};
std::string HaltReason{};
uint64_t Instrs{};
uint32_t Cur{};
};
struct threedo_AIF{
uint32_t Decompress{};
bool SelfReloc{};
uint32_t EntryTarget{};
uint32_t ExitSWI{};
uint32_t ROSize{};
uint32_t RWSize{};
uint32_t ZeroSize{};
uint32_t ImageBase{};
uint32_t AddrMode{};
Slice<uint8_t> Image{};
};
struct threedo_audioEvent{
int32_t cue{};
uint32_t time{};
};
struct threedo_Cel{
uint32_t Flags{};
uint32_t PRE0{};
uint32_t PRE1{};
uint32_t PIXC{};
int64_t Width{};
int64_t Height{};
int64_t BPP{};
bool Packed{};
bool Coded{};
Slice<uint16_t> PLUT{};
Slice<uint8_t> PDAT{};
};
struct threedo_bitReader{
Slice<uint8_t> data{};
int64_t pos{};
};
struct threedo_CvidMovie{
int64_t Width{};
int64_t Height{};
std::string Codec{};
int64_t FPS{};
int64_t HeaderRate{};
Slice<Slice<uint8_t>> Frames{};
Slice<uint32_t> Durations{};
Slice<uint32_t> Times{};
};
struct threedo_cvidVec{
std::array<uint8_t,4> y{};
int8_t u{};
int8_t v{};
};
struct threedo_CvidDecoder{
int64_t W{};
int64_t H{};
image_RGBA* img{};
Slice<Slice<threedo_cvidVec>> v1{};
Slice<Slice<threedo_cvidVec>> v4{};
};
struct threedo_CelDraw{
int64_t Index{};
uint32_t CCB{};
uint32_t Flags{};
uint32_t Src{};
uint32_t PLUT{};
int32_t XPos{};
int32_t YPos{};
int32_t HDX{};
int32_t HDY{};
int32_t VDX{};
int32_t VDY{};
int32_t HDDX{};
int32_t HDDY{};
uint32_t PIXC{};
uint32_t PRE0{};
uint32_t PRE1{};
int64_t BPP{};
int64_t Width{};
int64_t Height{};
bool Packed{};
bool Coded{};
bool LRForm{};
uint32_t Bitmap{};
int64_t BitmapW{};
int64_t BitmapH{};
};
struct threedo_PixelEvent{
bool Drawn{};
uint8_t R{};
uint8_t G{};
uint8_t B{};
uint8_t A{};
};
struct threedo_MemRegion{
std::string Name{};
uint32_t Base{};
uint32_t Size{};
};
struct threedo_diskStream{
std::string name{};
Slice<uint8_t> data{};
int64_t pos{};
};
struct threedo_Entry{
std::string Name{};
std::string Path{};
bool IsDir{};
std::string Type{};
int64_t Size{};
int64_t Blocks{};
int64_t Block{};
Slice<int64_t> Copies{};
};
struct threedo_dirScan{
Slice<threedo_Entry> entries{};
int64_t pos{};
};
struct threedo_span{
uint32_t addr{};
uint32_t size{};
};
struct threedo_heap{
uint32_t base{};
uint32_t total{};
Slice<threedo_span> free{};
Map<uint32_t,uint32_t> live{};
};
struct threedo_gfxBitmap{
uint32_t buf{};
int64_t w{};
int64_t h{};
};
struct threedo_item{
int32_t num{};
uint32_t typ{};
uint32_t addr{};
uint32_t tags{};
std::string name{};
int32_t owner{};
uint32_t signal{};
Slice<int32_t> msgs{};
int32_t device{};
int32_t replyPort{};
};
struct threedo_KernelCall{
uint32_t Offset{};
uint32_t From{};
std::array<uint32_t,4> Args{};
};
struct threedo_PadStep{
uint64_t AtStep{};
uint32_t Buttons{};
};
struct threedo_timerWait{
int32_t ioReq{};
int32_t submitter{};
uint32_t field{};
};
struct threedo_celCounters{
int64_t cels{};
int64_t pixels{};
int64_t chains{};
int64_t clears{};
int64_t folios{};
int64_t swis{};
};
struct threedo_ProfileBucket{
std::string Name{};
double Millis{};
int64_t Count{};
};
struct threedo_ProfileCounter{
std::string Name{};
int64_t Value{};
};
struct threedo_FrameProfile{
double TotalMs{};
Slice<threedo_ProfileBucket> Buckets{};
Slice<threedo_ProfileCounter> Counters{};
};
struct threedo_profState{
std::array<int64_t,4> ns{};
std::array<int64_t,4> count{};
time_Time runStart{};
bool inRun{};
int64_t frameNs{};
threedo_celCounters base{};
uint64_t baseInstr{};
int64_t gen{};
threedo_FrameProfile last{};
bool has{};
};
struct threedo_Machine{
Slice<uint8_t> dram{};
Slice<uint8_t> vram{};
threedo_heap* dheap{};
threedo_heap* vheap{};
Slice<uint8_t> imem{};
threedo_heap* iheap{};
arm60_CPU* CPU{};
threedo_Volume* vol{};
Map<uint32_t,threedo_diskStream*> streams{};
Map<uint32_t,threedo_dirScan*> dirs{};
Map<std::string,Slice<uint8_t>> nvram{};
Map<int32_t,threedo_gfxBitmap> bitmaps{};
Map<int32_t,int32_t> screenBM{};
uint32_t displayBuf{};
uint32_t audioTime{};
Slice<threedo_audioEvent> audioEvents{};
int32_t audioClockOwner{};
Map<int32_t,int32_t> attachCue{};
Slice<int32_t> ebListeners{};
Slice<threedo_PadStep> PadScript{};
bool CelDebug{};
Slice<std::string> CelDebugLog{};
Slice<std::string> CelFrameLog{};
bool SportDebug{};
bool PerspTint{};
uint32_t ProbeX{};
uint32_t ProbeY{};
std::function<void(threedo_Machine*,uint64_t,uint32_t)> OnDisplay{};
uint64_t frame{};
uint32_t WatchLo{};
uint32_t WatchHi{};
std::function<void(uint32_t,uint32_t,uint32_t)> OnWrite{};
uint32_t RWatchLo{};
uint32_t RWatchHi{};
std::function<void(uint32_t,uint32_t,uint32_t)> OnRead{};
std::function<void(threedo_Machine*,uint32_t)> OnStep{};
std::function<void(threedo_Machine*,uint32_t,uint32_t)> OnSWI{};
std::function<void(threedo_Machine*,int32_t,int32_t,std::string)> OnMsgQueue{};
Map<int32_t,threedo_item*> items{};
Map<uint32_t,threedo_item*> itemByType{};
int32_t nextItem{};
Slice<threedo_task*> tasks{};
int64_t cur{};
int64_t switches{};
bool needSchedule{};
Slice<threedo_KernelCall> KernelCalls{};
Slice<threedo_KernelCall> SWICalls{};
bool SpinBreak{};
int64_t SpinBreaks{};
int64_t StallTolerance{};
bool NoStreams{};
bool MovieHLE{};
Slice<threedo_armedMovie*> movieQueue{};
int64_t moviePos{};
int64_t movieFrameIdx{};
threedo_CvidDecoder* movieDec{};
uint32_t movieBase{};
uint64_t simTime{};
uint32_t vblank{};
bool PaceFields{};
Slice<threedo_timerWait> fieldWaits{};
uint32_t vblMirror{};
Slice<uint8_t> tty{};
Slice<std::string> Log{};
Map<std::string,bool> logSeen{};
bool Halted{};
std::string HaltReason{};
std::string imageHash{};
std::function<void(threedo_CelDraw)> OnCel{};
std::function<void(uint32_t,uint32_t,threedo_PixelEvent)> OnPixel{};
int64_t celLimit{};
int64_t celCount{};
bool StopRequested{};
bool Profile{};
threedo_profState prof{};
threedo_celCounters celCnt{};
};
struct threedo_armedMovie{
std::string name{};
threedo_CvidMovie* mv{};
};
struct threedo_Volume{
Slice<uint8_t> img{};
int64_t stride{};
int64_t dataOff{};
int64_t nsect{};
std::string Comment{};
std::string Label{};
uint32_t ID{};
int64_t blockSize{};
int64_t rootBlock{};
int64_t rootBlocks{};
};
struct threedo_Result{
uint64_t Steps{};
uint32_t PC{};
std::string Reason{};
};
using threedo_taskState=int64_t;
struct threedo_task{
int32_t num{};
std::string name{};
arm60_Context ctx{};
threedo_taskState state{};
uint32_t sig{};
uint32_t wait{};
uint32_t allocSigs{};
bool folioWait{};
};
struct threedo_Resource{
int64_t Offset{};
std::string Kind{};
int64_t Depth{};
};
struct threedo_WrapNode{
int64_t Offset{};
std::string Kind{};
Slice<threedo_WrapNode*> Children{};
};

#include "adapters-decl.h"
constexpr arm60_Flow arm60_FlowSeq=0ULL;
constexpr arm60_Flow arm60_FlowBranch=1ULL;
constexpr arm60_Flow arm60_FlowJump=2ULL;
constexpr arm60_Flow arm60_FlowCall=3ULL;
constexpr arm60_Flow arm60_FlowReturn=4ULL;
constexpr arm60_Flow arm60_FlowIndJump=5ULL;
constexpr arm60_Flow arm60_FlowStop=6ULL;
constexpr int64_t arm60_condEQ=0ULL;
constexpr int64_t arm60_condNE=1ULL;
constexpr int64_t arm60_condCS=2ULL;
constexpr int64_t arm60_condCC=3ULL;
constexpr int64_t arm60_condMI=4ULL;
constexpr int64_t arm60_condPL=5ULL;
constexpr int64_t arm60_condVS=6ULL;
constexpr int64_t arm60_condVC=7ULL;
constexpr int64_t arm60_condHI=8ULL;
constexpr int64_t arm60_condLS=9ULL;
constexpr int64_t arm60_condGE=10ULL;
constexpr int64_t arm60_condLT=11ULL;
constexpr int64_t arm60_condGT=12ULL;
constexpr int64_t arm60_condLE=13ULL;
constexpr int64_t arm60_condAL=14ULL;
constexpr int64_t arm60_condNV=15ULL;
std::array<std::string,16> arm60_condName=std::array<std::string,16>{std::string("EQ",2),std::string("NE",2),std::string("CS",2),std::string("CC",2),std::string("MI",2),std::string("PL",2),std::string("VS",2),std::string("VC",2),std::string("HI",2),std::string("LS",2),std::string("GE",2),std::string("LT",2),std::string("GT",2),std::string("LE",2),std::string("",0),std::string("NV",2)};
std::array<std::string,16> arm60_regName=std::array<std::string,16>{std::string("r0",2),std::string("r1",2),std::string("r2",2),std::string("r3",2),std::string("r4",2),std::string("r5",2),std::string("r6",2),std::string("r7",2),std::string("r8",2),std::string("r9",2),std::string("r10",3),std::string("r11",3),std::string("r12",3),std::string("sp",2),std::string("lr",2),std::string("pc",2)};
std::array<std::string,16> arm60_dpOps=std::array<std::string,16>{std::string("AND",3),std::string("EOR",3),std::string("SUB",3),std::string("RSB",3),std::string("ADD",3),std::string("ADC",3),std::string("SBC",3),std::string("RSC",3),std::string("TST",3),std::string("TEQ",3),std::string("CMP",3),std::string("CMN",3),std::string("ORR",3),std::string("MOV",3),std::string("BIC",3),std::string("MVN",3)};
std::array<std::string,4> arm60_shiftName=std::array<std::string,4>{std::string("LSL",3),std::string("LSR",3),std::string("ASR",3),std::string("ROR",3)};
constexpr int64_t arm60_ModeUSR=16ULL;
constexpr int64_t arm60_ModeFIQ=17ULL;
constexpr int64_t arm60_ModeIRQ=18ULL;
constexpr int64_t arm60_ModeSVC=19ULL;
constexpr int64_t arm60_ModeABT=23ULL;
constexpr int64_t arm60_ModeUND=27ULL;
constexpr int64_t arm60_ModeSYS=31ULL;
constexpr int64_t threedo_aifNOP=3785359360ULL;
constexpr int64_t threedo_swiStartInstrument=262145ULL;
constexpr int64_t threedo_swiStopInstrument=262147ULL;
constexpr int64_t threedo_swiTestHack=262151ULL;
constexpr int64_t threedo_swiConnectInstr=262152ULL;
constexpr int64_t threedo_swiSignalAtTime=262157ULL;
constexpr int64_t threedo_swiSetAudioRate=262159ULL;
constexpr int64_t threedo_swiSetAudioDuration=262160ULL;
constexpr int64_t threedo_swiTweakRawKnob=262161ULL;
constexpr int64_t threedo_swiStartAttachment=262162ULL;
constexpr int64_t threedo_swiStopAttachment=262164ULL;
constexpr int64_t threedo_swiLinkAttachments=262165ULL;
constexpr int64_t threedo_swiMonitorAttach=262166ULL;
constexpr int64_t threedo_swiSetAudioItemInfo=262171ULL;
constexpr int64_t threedo_swiAbortTimerCue=262177ULL;
constexpr int64_t threedo_typeInsTemplate=1025ULL;
constexpr int64_t threedo_typeInstrument=1026ULL;
constexpr int64_t threedo_typeKnob=1027ULL;
constexpr int64_t threedo_typeSample=1028ULL;
constexpr int64_t threedo_typeAudioCue=1029ULL;
constexpr int64_t threedo_typeAttachment=1031ULL;
constexpr int64_t threedo_audioTicksPerField=4ULL;
constexpr int64_t threedo_audioSampleRate=44100ULL;
Map<uint32_t,int64_t> threedo_bppFromPRE0=Map<uint32_t,int64_t>{{cast<uint32_t>(1ULL),cast<int64_t>(1ULL)},{cast<uint32_t>(2ULL),cast<int64_t>(2ULL)},{cast<uint32_t>(3ULL),cast<int64_t>(4ULL)},{cast<uint32_t>(4ULL),cast<int64_t>(6ULL)},{cast<uint32_t>(5ULL),cast<int64_t>(8ULL)},{cast<uint32_t>(6ULL),cast<int64_t>(16ULL)}};
constexpr int64_t threedo_ccbPacked=512ULL;
constexpr int64_t threedo_packEOL=0ULL;
constexpr int64_t threedo_packLiteral=1ULL;
constexpr int64_t threedo_packTransparent=2ULL;
constexpr int64_t threedo_packRepeat=3ULL;
constexpr int64_t threedo_seekSet=1ULL;
constexpr int64_t threedo_seekCur=2ULL;
constexpr int64_t threedo_seekEnd=3ULL;
constexpr int64_t threedo_swiOpenDiskFile=196608ULL;
constexpr int64_t threedo_swiCloseDiskFile=196609ULL;
constexpr int64_t threedo_swiChangeDir=196615ULL;
constexpr int64_t threedo_swiGetDir=196616ULL;
constexpr int64_t threedo_swiCreateFile=196617ULL;
constexpr int64_t threedo_swiDeleteFile=196618ULL;
constexpr int64_t threedo_fileErrNotFound=4294966373ULL;
constexpr int64_t threedo_simTick=100000ULL;
constexpr int64_t threedo_memtypeVRAM=65536ULL;
constexpr int64_t threedo_memtypeDRAM=524288ULL;
constexpr int64_t threedo_scrBitmapCount=52ULL;
constexpr int64_t threedo_scrTempBitmap=120ULL;
constexpr int64_t threedo_bmBuffer=36ULL;
constexpr int64_t threedo_bmWidth=40ULL;
constexpr int64_t threedo_bmHeight=44ULL;
constexpr int64_t threedo_bmClipWidth=56ULL;
constexpr int64_t threedo_bmClipHeight=60ULL;
constexpr int64_t threedo_csgDisplayHeight=1ULL;
constexpr int64_t threedo_csgScreenCount=2ULL;
constexpr int64_t threedo_csgScreenHeight=3ULL;
constexpr int64_t threedo_csgBitmapCount=4ULL;
constexpr int64_t threedo_csgBitmapWidths=5ULL;
constexpr int64_t threedo_csgBitmapHeights=6ULL;
constexpr int64_t threedo_csgBitmapBufs=7ULL;
constexpr int64_t threedo_ccbSkip=2147483648ULL;
constexpr int64_t threedo_ccbLast=1073741824ULL;
constexpr int64_t threedo_ccbNPAbs=536870912ULL;
constexpr int64_t threedo_ccbSPAbs=268435456ULL;
constexpr int64_t threedo_ccbPPAbs=134217728ULL;
constexpr int64_t threedo_ccbLDSize=67108864ULL;
constexpr int64_t threedo_ccbLDPLUT=8388608ULL;
constexpr int64_t threedo_ccbCCBPre=4194304ULL;
constexpr int64_t threedo_ccbPOVER=384ULL;
constexpr int64_t threedo_ccbBGND=32ULL;
constexpr int64_t threedo_pre1LRForm=2048ULL;
constexpr int64_t threedo_ioInfoOff=52ULL;
constexpr int64_t threedo_ioActualOff=84ULL;
constexpr int64_t threedo_ioFlagsOff=88ULL;
constexpr int64_t threedo_ioErrorOff=92ULL;
constexpr int64_t threedo_ioMsgItemOff=104ULL;
constexpr int64_t threedo_ioiCommand=0ULL;
constexpr int64_t threedo_ioiOffset=12ULL;
constexpr int64_t threedo_ioiSendBuf=16ULL;
constexpr int64_t threedo_ioiSendLen=20ULL;
constexpr int64_t threedo_ioiRecvBuf=24ULL;
constexpr int64_t threedo_ioiRecvLen=28ULL;
constexpr int64_t threedo_ioInfoBytes=32ULL;
constexpr int64_t threedo_ioDone=1ULL;
constexpr int64_t threedo_ioQuick=2ULL;
constexpr int64_t threedo_cmdWrite=0ULL;
constexpr int64_t threedo_cmdRead=1ULL;
constexpr int64_t threedo_cmdStatus=2ULL;
constexpr int64_t threedo_fileCmdReadDir=3ULL;
constexpr int64_t threedo_fileCmdAllocBlocks=6ULL;
constexpr int64_t threedo_fileCmdSetEOF=7ULL;
constexpr int64_t threedo_timerCmdWaitField=3ULL;
constexpr int64_t threedo_sportFlashWrite=6ULL;
constexpr int64_t threedo_msgReplyPort=36ULL;
constexpr int64_t threedo_msgResult=40ULL;
constexpr int64_t threedo_msgDataPtr=44ULL;
constexpr int64_t threedo_msgDataSize=48ULL;
constexpr int64_t threedo_msgMsgPort=52ULL;
constexpr int64_t threedo_ebConfigure=1ULL;
constexpr int64_t threedo_ebEventRecord=3ULL;
constexpr int64_t threedo_ebDescribePods=27ULL;
constexpr int64_t threedo_ebDescribePodsReply=28ULL;
constexpr int64_t threedo_podIsControlPad=2147483648ULL;
constexpr int64_t threedo_eventNumButtonUpdate=3ULL;
constexpr int64_t threedo_PadDown=2147483648ULL;
constexpr int64_t threedo_PadUp=1073741824ULL;
constexpr int64_t threedo_PadRight=536870912ULL;
constexpr int64_t threedo_PadLeft=268435456ULL;
constexpr int64_t threedo_PadA=134217728ULL;
constexpr int64_t threedo_PadB=67108864ULL;
constexpr int64_t threedo_PadC=33554432ULL;
constexpr int64_t threedo_PadStart=16777216ULL;
constexpr int64_t threedo_PadX=8388608ULL;
constexpr int64_t threedo_PadRightShift=4194304ULL;
constexpr int64_t threedo_PadLeftShift=2097152ULL;
Map<std::string,uint32_t> threedo_padButtonBits=Map<std::string,uint32_t>{{std::string("a",1),threedo_PadA},{std::string("b",1),threedo_PadB},{std::string("c",1),threedo_PadC},{std::string("x",1),threedo_PadX},{std::string("start",5),threedo_PadStart},{std::string("p",1),threedo_PadStart},{std::string("up",2),threedo_PadUp},{std::string("down",4),threedo_PadDown},{std::string("left",4),threedo_PadLeft},{std::string("right",5),threedo_PadRight},{std::string("ls",2),threedo_PadLeftShift},{std::string("rs",2),threedo_PadRightShift}};
constexpr int64_t threedo_swiCreateSizedItem=65536ULL;
constexpr int64_t threedo_swiWaitSignal=65537ULL;
constexpr int64_t threedo_swiSendSignal=65538ULL;
constexpr int64_t threedo_swiDeleteItem=65539ULL;
constexpr int64_t threedo_swiFindItem=65540ULL;
constexpr int64_t threedo_swiOpenItem=65541ULL;
constexpr int64_t threedo_swiUnlockItem=65542ULL;
constexpr int64_t threedo_swiLockItem=65543ULL;
constexpr int64_t threedo_swiCloseItem=65544ULL;
constexpr int64_t threedo_swiSetItemPri=65546ULL;
constexpr int64_t threedo_swiAllocSignal=65557ULL;
constexpr int64_t threedo_swiFreeSignal=65558ULL;
constexpr int64_t threedo_swiPrintf=65550ULL;
constexpr int64_t threedo_swiGetThisMsg=65551ULL;
constexpr int64_t threedo_swiPutMsg=65552ULL;
constexpr int64_t threedo_swiReplyMsg=65554ULL;
constexpr int64_t threedo_swiGetMsg=65555ULL;
constexpr int64_t threedo_swiSendIO=65560ULL;
constexpr int64_t threedo_swiAbortIO=65561ULL;
constexpr int64_t threedo_swiCompleteIO=65570ULL;
constexpr int64_t threedo_swiDoIO=65573ULL;
constexpr int64_t threedo_swiWaitIO=65577ULL;
constexpr int64_t threedo_typeSemaphore=7ULL;
constexpr int64_t threedo_typeMsg=9ULL;
constexpr int64_t threedo_typeMsgPort=10ULL;
constexpr int64_t threedo_typeIOReq=14ULL;
constexpr int64_t threedo_typeDevice=15ULL;
constexpr int64_t threedo_tagItemName=1ULL;
constexpr int64_t threedo_tagPortSignal=10ULL;
constexpr int64_t threedo_tagIOReqReplyPort=10ULL;
constexpr int64_t threedo_tagIOReqDevice=11ULL;
constexpr int64_t threedo_tagMsgReplyPort=10ULL;
constexpr int64_t threedo_sigfIODONE=8ULL;
constexpr int64_t threedo_dramSize=2097152ULL;
constexpr int64_t threedo_vramBase=2097152ULL;
constexpr int64_t threedo_vramSize=1048576ULL;
constexpr int64_t threedo_madamBase=53477376ULL;
constexpr int64_t threedo_madamEnd=54525952ULL;
constexpr int64_t threedo_clioBase=54525952ULL;
constexpr int64_t threedo_clioEnd=55574528ULL;
constexpr int64_t threedo_kernelBase=1572864ULL;
constexpr int64_t threedo_hleBase=266338304ULL;
constexpr int64_t threedo_hleSize=65536ULL;
constexpr int64_t threedo_bootTaskNum=1ULL;
constexpr int64_t threedo_fileFolioBase=1568768ULL;
constexpr int64_t threedo_hleFileTag=32768ULL;
constexpr int64_t threedo_otherFolioBase=1566720ULL;
constexpr int64_t threedo_hleOtherTag=40960ULL;
constexpr int64_t threedo_gfxFolioBase=1564672ULL;
constexpr int64_t threedo_hleGfxTag=49152ULL;
constexpr int64_t threedo_audioFolioBase=1562624ULL;
constexpr int64_t threedo_hleAudioTag=57344ULL;
constexpr int64_t threedo_mathFolioBase=1565696ULL;
constexpr int64_t threedo_hleMathTag=24576ULL;
constexpr int64_t threedo_osCtxBase=1560576ULL;
constexpr int64_t threedo_osCtxPri=10ULL;
constexpr int64_t threedo_osCtxItem=24ULL;
constexpr int64_t threedo_osCtxMemLst=168ULL;
constexpr int64_t threedo_dheapBase=524288ULL;
constexpr int64_t threedo_dheapTop=1560576ULL;
constexpr int64_t threedo_imemBase=4194304ULL;
constexpr int64_t threedo_imemSize=4194304ULL;
constexpr int64_t threedo_vheapBase=2097152ULL;
constexpr int64_t threedo_vramReserve=16384ULL;
constexpr int64_t threedo_vheapTop=3129344ULL;
constexpr int64_t threedo_screenW=320ULL;
constexpr int64_t threedo_screenH=240ULL;
constexpr int64_t threedo_rawSectorSize=2352ULL;
constexpr int64_t threedo_userSize=2048ULL;
constexpr int64_t threedo_labelBlock=0ULL;
constexpr int64_t threedo_entryTypeMask=255ULL;
constexpr int64_t threedo_typeDir=7ULL;
constexpr int64_t threedo_swiMulVec3Mat33=327680ULL;
constexpr int64_t threedo_swiMulMat33Mat33=327681ULL;
constexpr int64_t threedo_swiMulManyVec3Mat33=327682ULL;
constexpr int64_t threedo_swiDot3=327692ULL;
constexpr int64_t threedo_swiCross3=327694ULL;
constexpr int64_t threedo_bucketCel=0ULL;
constexpr int64_t threedo_bucketClear=1ULL;
constexpr int64_t threedo_bucketFolio=2ULL;
constexpr int64_t threedo_bucketSWI=3ULL;
constexpr int64_t threedo_numBuckets=4ULL;
std::array<std::string,4> threedo_bucketNames=[](){std::array<std::string,4> v{};v[0]=std::string("cel draw",8);v[1]=std::string("flash clear",11);v[2]=std::string("folio HLE",9);v[3]=std::string("swi",3);return v;}();
std::array<int64_t,2> threedo_gfxBuckets=std::array<int64_t,2>{threedo_bucketCel,threedo_bucketClear};
constexpr int64_t threedo_vblankPeriod=20000ULL;
constexpr int64_t threedo_taskExitTramp=267386880ULL;
constexpr threedo_taskState threedo_stReady=0ULL;
constexpr threedo_taskState threedo_stRunning=1ULL;
constexpr threedo_taskState threedo_stWaiting=2ULL;
constexpr threedo_taskState threedo_stDone=3ULL;
uint32_t arm60_signExtend(uint32_t v,uint64_t n);
uint32_t arm60_ror32(uint32_t v,uint32_t n);
std::string arm60_cn(int64_t cond);
std::tuple<uint32_t,bool> arm60_word(Slice<uint8_t> code);
arm60_Inst arm60_Decode(Slice<uint8_t> code,uint32_t addr);
arm60_Inst arm60_undef(uint32_t w,arm60_Inst in);
arm60_Inst arm60_decodeDataMisc(uint32_t w,arm60_Inst in,bool immForm);
arm60_Inst arm60_decodeDataProc(uint32_t w,arm60_Inst in,bool immForm);
std::string arm60_imm(uint32_t v);
std::string arm60_shiftOperand(uint32_t w);
arm60_Inst arm60_decodeMisc(uint32_t w,arm60_Inst in,bool immForm);
arm60_Inst arm60_decodeMSRreg(uint32_t w,arm60_Inst in);
arm60_Inst arm60_decodeMSRimm(uint32_t w,arm60_Inst in);
std::string arm60_msrFields(uint32_t w);
arm60_Inst arm60_decodeExtension(uint32_t w,arm60_Inst in);
arm60_Inst arm60_decodeSingle(uint32_t w,arm60_Inst in,bool regOff);
std::string arm60_addrForm(std::string mnem,uint32_t rd,uint32_t rn,std::string off,uint32_t p,uint32_t wb);
arm60_Inst arm60_decodeBlock(uint32_t w,arm60_Inst in);
std::string arm60_regList(uint32_t mask);
arm60_Inst arm60_decodeBranch(uint32_t w,uint32_t addr,arm60_Inst in);
arm60_Inst arm60_decodeCoproLS(uint32_t w,arm60_Inst in);
arm60_Inst arm60_decodeCopro(uint32_t w,arm60_Inst in);
Slice<std::string> arm60_Disassemble(Slice<uint8_t> code,uint32_t base);
arm60_CPU* arm60_NewCPU(threedo_Machine* bus);
void arm60_CPU_Reset(arm60_CPU* c);
uint32_t arm60_CPU_CurPC(arm60_CPU* c);
arm60_Context arm60_CPU_SaveContext(arm60_CPU* c);
void arm60_CPU_RestoreContext(arm60_CPU* c,arm60_Context x);
arm60_CPUState arm60_CPU_SaveState(arm60_CPU* c);
void arm60_CPU_LoadState(arm60_CPU* c,arm60_CPUState s);
void arm60_CPU_SetPC(arm60_CPU* c,uint32_t v);
void arm60_CPU_SetReg(arm60_CPU* c,uint32_t i,uint32_t v);
uint32_t arm60_CPU_Reg(arm60_CPU* c,uint32_t i);
uint32_t arm60_CPU_CPSR(arm60_CPU* c);
void arm60_CPU_SetCPSR(arm60_CPU* c,uint32_t v);
int64_t arm60_modeIndex(uint32_t mode);
void arm60_CPU_switchMode(arm60_CPU* c,uint32_t mode);
uint32_t arm60_CPU_SPSR(arm60_CPU* c);
void arm60_CPU_SetSPSR(arm60_CPU* c,uint32_t v);
uint8_t arm60_CPU_read8(arm60_CPU* c,uint32_t a);
void arm60_CPU_write8(arm60_CPU* c,uint32_t a,uint8_t v);
uint32_t arm60_CPU_read32(arm60_CPU* c,uint32_t a);
uint32_t arm60_CPU_reg(arm60_CPU* c,uint32_t i);
void arm60_CPU_setReg(arm60_CPU* c,uint32_t i,uint32_t v);
void arm60_CPU_setNZ(arm60_CPU* c,uint32_t v);
uint32_t arm60_CPU_add(arm60_CPU* c,uint32_t a,uint32_t b,uint32_t cin);
uint32_t arm60_CPU_sub(arm60_CPU* c,uint32_t a,uint32_t b,uint32_t cin);
uint32_t arm60_CPU_boolToU(arm60_CPU* c,bool b);
bool arm60_CPU_cond(arm60_CPU* c,int64_t cc);
std::tuple<uint32_t,uint32_t> arm60_CPU_shift(arm60_CPU* c,uint32_t typ,uint32_t amt,uint32_t val,bool regForm,uint32_t cin);
int64_t arm60_CPU_Step(arm60_CPU* c);
void arm60_CPU_execARM(arm60_CPU* c,uint32_t w);
void arm60_CPU_execDataMisc(arm60_CPU* c,uint32_t w,bool immForm);
std::tuple<uint32_t,uint32_t> arm60_CPU_dpOperand(arm60_CPU* c,uint32_t w,bool immForm);
void arm60_CPU_execDataProc(arm60_CPU* c,uint32_t w,bool immForm);
void arm60_CPU_execBranch(arm60_CPU* c,uint32_t w);
void arm60_CPU_execExtension(arm60_CPU* c,uint32_t w);
void arm60_CPU_execMul(arm60_CPU* c,uint32_t w);
void arm60_CPU_execMulLong(arm60_CPU* c,uint32_t w);
void arm60_CPU_execSwap(arm60_CPU* c,uint32_t w);
void arm60_CPU_execSingle(arm60_CPU* c,uint32_t w,bool regOff);
void arm60_CPU_execBlock(arm60_CPU* c,uint32_t w);
void arm60_CPU_execMisc(arm60_CPU* c,uint32_t w,bool immForm);
void arm60_CPU_execMSR(arm60_CPU* c,uint32_t w,bool immForm);
void arm60_CPU_execSWI(arm60_CPU* c,uint32_t w);
void arm60_CPU_exception(arm60_CPU* c,uint32_t mode,uint32_t vector,uint32_t returnAddr);
void arm60_CPU_Exception(arm60_CPU* c,uint32_t mode,uint32_t vector,uint32_t returnAddr);
bool arm60_CPU_IRQ(arm60_CPU* c);
void arm60_CPU_execCopro(arm60_CPU* c,uint32_t w);
uint32_t threedo_be32at(Slice<uint8_t> b,int64_t o);
std::tuple<threedo_AIF*,Error> threedo_ParseAIF(Slice<uint8_t> data);
std::string threedo_AIF_Describe(threedo_AIF* a);
bool threedo_Machine_audioFolioSWI(threedo_Machine* m,arm60_CPU* c,uint32_t swi);
void threedo_Machine_completeAttachment(threedo_Machine* m,int32_t att);
void threedo_Machine_serviceAudioFolio(threedo_Machine* m,uint32_t foff);
void threedo_Machine_advanceAudioClock(threedo_Machine* m,uint32_t fields);
std::tuple<threedo_Cel*,Error> threedo_ParseCel(Slice<uint8_t> data);
std::tuple<threedo_Cel*,Error> threedo_Cel_finish(threedo_Cel* c);
color_RGBA threedo_RGB555(uint16_t v);
color_RGBA threedo_rgb555(uint16_t v);
uint32_t threedo_bitReader_read(threedo_bitReader* br,int64_t n);
std::tuple<image_RGBA*,Error> threedo_Cel_Image(threedo_Cel* c);
std::tuple<uint16_t,uint8_t> threedo_Cel_ppmp(threedo_Cel* c,uint16_t pix,uint32_t amv);
void threedo_Cel_decodePacked(threedo_Cel* c,std::function<void(int64_t,int64_t,uint32_t)> set);
void threedo_Cel_decodeUnpacked(threedo_Cel* c,std::function<void(int64_t,int64_t,uint32_t)> set);
std::tuple<threedo_CvidMovie*,Error> threedo_DemuxStream(Slice<uint8_t> data);
threedo_CvidDecoder* threedo_NewCvidDecoder(int64_t w,int64_t h);
image_RGBA* threedo_CvidDecoder_Frame(threedo_CvidDecoder* d);
int64_t threedo_tdiv2(int64_t x);
uint8_t threedo_clamp8(int64_t v);
void threedo_CvidDecoder_setPix(threedo_CvidDecoder* d,int64_t x,int64_t y,uint8_t ly,int8_t u,int8_t v);
void threedo_CvidDecoder_paintV4(threedo_CvidDecoder* d,int64_t px,int64_t py,threedo_cvidVec a,threedo_cvidVec b,threedo_cvidVec c,threedo_cvidVec e);
void threedo_CvidDecoder_paintV1(threedo_CvidDecoder* d,int64_t px,int64_t py,threedo_cvidVec vec);
void threedo_decodeCodebook(Slice<threedo_cvidVec> book,uint16_t cid,Slice<uint8_t> data);
void threedo_CvidDecoder_decodeVectors(threedo_CvidDecoder* d,uint16_t cid,Slice<uint8_t> data,int64_t x0,int64_t y0,int64_t x1,int64_t y1,Slice<threedo_cvidVec> v1,Slice<threedo_cvidVec> v4);
void threedo_CvidDecoder_DecodeFrame(threedo_CvidDecoder* d,Slice<uint8_t> fr);
std::string threedo_CelDraw_Name(threedo_CelDraw c);
std::string threedo_CelDraw_Decoded(threedo_CelDraw c);
Slice<uint64_t> threedo_CelDraw_Words(threedo_CelDraw c);
void threedo_Machine_celDraw(threedo_Machine* m,threedo_gfxBitmap bm,uint32_t ccb,uint32_t flags,uint32_t src,uint32_t plutPtr);
void threedo_Machine_celPixel(threedo_Machine* m,threedo_gfxBitmap bm,int64_t x,int64_t y,uint16_t pix);
threedo_Result threedo_Machine_RunStopAfterCel(threedo_Machine* m,int64_t k,uint64_t budget);
int64_t threedo_Machine_Cels(threedo_Machine* m);
std::tuple<uint32_t,int64_t,int64_t,bool> threedo_Machine_DrawTarget(threedo_Machine* m);
Slice<threedo_MemRegion> threedo_Machine_Bitmaps(threedo_Machine* m);
Slice<threedo_MemRegion> threedo_Machine_MemRegions(threedo_Machine* m);
std::tuple<image_RGBA*,Error> threedo_Machine_RenderBitmap(threedo_Machine* m,uint32_t buf,int64_t w,int64_t h);
uint64_t threedo_Machine_Frames(threedo_Machine* m);
threedo_Volume* threedo_Machine_Volume(threedo_Machine* m);
void threedo_Machine_serviceFileFolio(threedo_Machine* m,uint32_t foff);
void threedo_Machine_serviceOtherFolio(threedo_Machine* m,uint32_t foff);
uint32_t threedo_Machine_openDiskStream(threedo_Machine* m,std::string name);
int32_t threedo_Machine_readDiskStream(threedo_Machine* m,uint32_t handle,uint32_t buf,int32_t n);
int32_t threedo_Machine_seekDiskStream(threedo_Machine* m,uint32_t handle,int32_t offset,uint32_t whence);
uint32_t threedo_Machine_openDirectoryPath(threedo_Machine* m,std::string path);
int32_t threedo_Machine_readDirectory(threedo_Machine* m,uint32_t dir,uint32_t buf);
void threedo_Machine_closeDirectory(threedo_Machine* m,uint32_t dir);
void threedo_Machine_closeDiskStream(threedo_Machine* m,uint32_t handle);
std::tuple<std::string,bool> threedo_nvramPath(std::string path);
bool threedo_Machine_fileFolioSWI(threedo_Machine* m,arm60_CPU* c,uint32_t swi);
uint32_t threedo_Machine_openDiskFile(threedo_Machine* m,std::string path);
std::tuple<Slice<uint8_t>,uint32_t,std::string,bool> threedo_Machine_fileData(threedo_Machine* m,std::string name);
std::tuple<Slice<uint8_t>,std::string,bool> threedo_Machine_loadDiscFile(threedo_Machine* m,std::string name);
threedo_heap* threedo_newHeap(uint32_t base,uint32_t total);
uint32_t threedo_heap_alloc(threedo_heap* h,uint32_t size);
void threedo_heap_freeBlock(threedo_heap* h,uint32_t addr);
bool threedo_Machine_serviceFolio(threedo_Machine* m,uint32_t off);
uint32_t threedo_Machine_lookupItem(threedo_Machine* m,int32_t num);
threedo_heap* threedo_Machine_poolFor(threedo_Machine* m,uint32_t flags);
threedo_heap* threedo_Machine_poolOf(threedo_Machine* m,uint32_t ptr);
image_RGBA* threedo_Machine_CaptureVRAM(threedo_Machine* m,uint32_t base,int64_t w,int64_t h);
int64_t threedo_Machine_VRAMNonZero(threedo_Machine* m,int64_t n);
void threedo_Machine_flashClearRange(threedo_Machine* m,uint32_t dest,uint32_t bytes,uint16_t val);
std::string threedo_gfxFuncName(uint32_t foff);
void threedo_Machine_serviceGraphicsFolio(threedo_Machine* m,uint32_t foff);
uint32_t threedo_Machine_createScreenGroup(threedo_Machine* m,uint32_t itemArray,uint32_t tags);
uint32_t threedo_Machine_pixelAddress(threedo_Machine* m,int32_t itemNum,uint32_t x,uint32_t y);
uint32_t threedo_Machine_drawCels(threedo_Machine* m,int32_t bitmapItem,uint32_t ccb);
bool threedo_Machine_drawOneCel(threedo_Machine* m,threedo_gfxBitmap bm,uint32_t ccb,uint32_t flags,uint32_t src,uint32_t plutPtr);
void threedo_Machine_decodeLRForm16(threedo_Machine* m,threedo_Cel* c,uint32_t src,std::function<void(int64_t,int64_t,uint32_t)> set);
std::tuple<uint16_t,uint32_t,bool> threedo_Machine_decodePixel(threedo_Machine* m,threedo_Cel* cel,uint32_t v,uint32_t flags,bool bgnd);
void threedo_Machine_blendPixel(threedo_Machine* m,threedo_gfxBitmap bm,int64_t x,int64_t y,uint16_t pix,uint32_t amv,uint32_t pixc,uint32_t flags);
uint32_t threedo_pdv(uint32_t n);
std::tuple<uint32_t,uint32_t,uint32_t> threedo_chan5(uint16_t p);
uint32_t threedo_Machine_read32(threedo_Machine* m,uint32_t a);
void threedo_Machine_write32(threedo_Machine* m,uint32_t a,uint32_t v);
std::string threedo_Machine_readCStr(threedo_Machine* m,uint32_t a);
uint32_t threedo_Machine_tagArg(threedo_Machine* m,uint32_t p,uint32_t want);
std::string threedo_Machine_tagString(threedo_Machine* m,uint32_t p,uint32_t want);
void threedo_Machine_serviceIO(threedo_Machine* m,arm60_CPU* c,bool async);
void threedo_Machine_completeIO(threedo_Machine* m,threedo_item* it);
std::string threedo_Machine_deviceName(threedo_Machine* m,threedo_item* it);
void threedo_Machine_completeFieldWait(threedo_Machine* m,threedo_item* it,int32_t submitter);
uint32_t threedo_Machine_ioError(threedo_Machine* m,int32_t ioNum);
std::tuple<int32_t,int32_t> threedo_Machine_performIO(threedo_Machine* m,threedo_item* it,uint32_t cmd,uint32_t offset,uint32_t sendBuf,uint32_t sendLen,uint32_t buf,uint32_t length);
std::tuple<int32_t,int32_t> threedo_Machine_fileDeviceIO(threedo_Machine* m,std::string name,uint32_t cmd,uint32_t offset,uint32_t sendBuf,uint32_t sendLen,uint32_t buf,uint32_t length);
void threedo_Machine_serviceMsg(threedo_Machine* m,arm60_CPU* c,uint32_t swi);
void threedo_Machine_replyMsg(threedo_Machine* m,threedo_item* msg,uint32_t result);
void threedo_Machine_waitPort(threedo_Machine* m);
std::tuple<uint32_t,bool> threedo_PadButton(std::string name);
Slice<std::string> threedo_PadButtonNames();
void threedo_Machine_eventBrokerRequest(threedo_Machine* m,threedo_item* port,threedo_item* msg,uint32_t dataPtr,uint32_t dataSize);
void threedo_Machine_SendPadEvent(threedo_Machine* m,uint32_t buttons);
uint32_t threedo_Machine_brokerPortNum(threedo_Machine* m);
uint32_t threedo_Machine_kprintf(threedo_Machine* m);
bool threedo_Machine_kernelSWI(threedo_Machine* m,arm60_CPU* c,uint32_t swi);
threedo_item* threedo_Machine_createItem(threedo_Machine* m,uint32_t typ,uint32_t tags,uint32_t size);
void threedo_Machine_initItemFromTags(threedo_Machine* m,threedo_item* it);
uint32_t threedo_Machine_allocSignalFor(threedo_Machine* m,int32_t taskNum);
Slice<std::string> threedo_Machine_ItemsSummary(threedo_Machine* m);
threedo_item* threedo_Machine_findItem(threedo_Machine* m,uint32_t typ,uint32_t tags);
threedo_Machine* threedo_NewMachine();
void threedo_Machine_LoadAIF(threedo_Machine* m,threedo_AIF* a);
void threedo_Machine_SetVolume(threedo_Machine* m,threedo_Volume* v);
uint32_t threedo_Machine_DisplayBuffer(threedo_Machine* m);
void threedo_Machine_SetVBLMirror(threedo_Machine* m,uint32_t addr);
void threedo_Machine_advanceVBlank(threedo_Machine* m,uint32_t n);
void threedo_Machine_startFieldWait(threedo_Machine* m,threedo_item* it,uint32_t fields);
bool threedo_Machine_fieldTick(threedo_Machine* m);
bool threedo_Machine_wakeByFieldTick(threedo_Machine* m);
void threedo_Machine_writeWord(threedo_Machine* m,uint32_t a,uint32_t v);
void threedo_Machine_note(threedo_Machine* m,std::string s);
std::string threedo_Machine_TTY(threedo_Machine* m);
uint8_t threedo_Machine_Read(threedo_Machine* m,uint32_t addr);
uint8_t threedo_Machine_rawRead(threedo_Machine* m,uint32_t addr);
void threedo_Machine_Write(threedo_Machine* m,uint32_t addr,uint8_t v);
bool threedo_Machine_isFlagSpin(threedo_Machine* m,Slice<uint32_t> pcs);
Slice<uint32_t> threedo_Machine_breakSpin(threedo_Machine* m,Slice<uint32_t> pcs);
bool threedo_Machine_swi(threedo_Machine* m,arm60_CPU* c,uint32_t comment);
void threedo_Machine_armMovie(threedo_Machine* m,std::string name);
int64_t threedo_Machine_MoviesPending(threedo_Machine* m);
std::tuple<uint32_t,std::string,int64_t,int64_t,bool> threedo_Machine_StepMovieFrame(threedo_Machine* m);
Slice<std::string> threedo_Machine_MovieNames(threedo_Machine* m);
void threedo_Machine_PlayMovies(threedo_Machine* m,std::function<void(std::string,int64_t,int64_t)> onFrame);
uint32_t threedo_Machine_movieTarget(threedo_Machine* m);
void threedo_Machine_blitRGBAToVRAM(threedo_Machine* m,image_RGBA* img,uint32_t base,int64_t w,int64_t h);
std::string threedo_baseName(std::string p);
std::string threedo_Entry_String(threedo_Entry e);
uint32_t threedo_be32(Slice<uint8_t> b);
std::string threedo_trimName(Slice<uint8_t> b);
int64_t threedo_indexByte(Slice<uint8_t> b,uint8_t c);
std::tuple<Slice<threedo_Entry>,Error> threedo_Volume_dirEntries(threedo_Volume* v,int64_t startBlock,std::string dirPath);
std::tuple<threedo_Entry,Error> threedo_Volume_resolve(threedo_Volume* v,std::string path);
Slice<std::string> threedo_splitPath(std::string p);
std::tuple<Slice<threedo_Entry>,Error> threedo_Volume_ReadDir(threedo_Volume* v,std::string path);
Error threedo_Volume_Walk(threedo_Volume* v,std::function<Error(threedo_Entry)> fn);
Error threedo_Volume_walk(threedo_Volume* v,int64_t block,std::string dirPath,std::function<Error(threedo_Entry)> fn);
std::tuple<Slice<uint8_t>,Error> threedo_Volume_ReadFile(threedo_Volume* v,std::string path);
bool threedo_Machine_mathFolioSWI(threedo_Machine* m,arm60_CPU* c,uint32_t swi);
int32_t threedo_fmul16(int32_t a,int32_t b);
std::array<int32_t,3> threedo_Machine_readVec3(threedo_Machine* m,uint32_t p);
void threedo_Machine_mulVec3Mat33(threedo_Machine* m,uint32_t dest,uint32_t vec,uint32_t mat);
void threedo_Machine_mulMat33Mat33(threedo_Machine* m,uint32_t dest,uint32_t src1,uint32_t src2);
void threedo_Machine_serviceMathFolio(threedo_Machine* m,uint32_t foff);
std::tuple<uint32_t,uint32_t> threedo_divUF16(uint32_t d1,uint32_t d2);
std::tuple<int32_t,uint32_t> threedo_divSF16(int32_t d1,int32_t d2);
void threedo_Machine_cross3(threedo_Machine* m,uint32_t dest,uint32_t v1p,uint32_t v2p);
time_Time threedo_Machine_profStart(threedo_Machine* m);
void threedo_Machine_profEnd(threedo_Machine* m,int64_t bucket,time_Time t);
int64_t threedo_Machine_profGfxNs(threedo_Machine* m);
void threedo_Machine_profEndFolio(threedo_Machine* m,time_Time t,int64_t gfxBefore,int64_t gen);
void threedo_Machine_profRunEnter(threedo_Machine* m);
void threedo_Machine_profRunExit(threedo_Machine* m);
void threedo_Machine_profFrame(threedo_Machine* m);
threedo_FrameProfile threedo_Machine_FrameProfile(threedo_Machine* m);
void threedo_Machine_SetProfile(threedo_Machine* m,bool on);
threedo_Result threedo_Machine_Run(threedo_Machine* m,uint64_t maxSteps);
void threedo_Machine_serviceKernelCall(threedo_Machine* m,uint32_t pc);
void threedo_Machine_SetResultAndReturn(threedo_Machine* m,uint32_t result);
std::string threedo_Machine_DisasmAt(threedo_Machine* m,uint32_t addr);
void threedo_Machine_initTasks(threedo_Machine* m);
int32_t threedo_Machine_spawnTask(threedo_Machine* m,uint32_t tagList);
Slice<std::string> threedo_Machine_TaskSummary(threedo_Machine* m);
std::tuple<uint32_t,uint32_t,uint32_t,uint32_t> threedo_Machine_parseTaskTags(threedo_Machine* m,uint32_t p);
uint32_t threedo_Machine_readWord(threedo_Machine* m,uint32_t a);
threedo_task* threedo_Machine_curTask(threedo_Machine* m);
int32_t threedo_Machine_CurrentTaskNum(threedo_Machine* m);
threedo_task* threedo_Machine_taskByNum(threedo_Machine* m,int32_t num);
bool threedo_Machine_switchTask(threedo_Machine* m);
bool threedo_Machine_sendSignal(threedo_Machine* m,int32_t num,uint32_t sigs);
void threedo_Machine_yieldTo(threedo_Machine* m,bool woke,int32_t num);
std::tuple<Slice<threedo_Resource>,Error> threedo_ParseWrap(Slice<uint8_t> data);
std::string threedo_kindOf(Slice<uint8_t> data,int64_t off);
Map<std::string,int64_t> threedo_Inventory(Slice<threedo_Resource> res);
std::tuple<threedo_WrapNode*,Error> threedo_ParseWrapTree(Slice<uint8_t> data);

#include "adapters.h"
// tools/cpu/arm60/arm60.go:84:1
uint32_t arm60_signExtend(uint32_t v,uint64_t n){
{
uint32_t m = shl<uint32_t>(cast<uint32_t>(cast<uint32_t>(1ULL)),(cast<uint64_t>((n - cast<uint64_t>(1ULL)))));
return cast<uint32_t>(((cast<uint32_t>((v ^ m))) - m));
}
}
// tools/cpu/arm60/arm60.go:89:1
uint32_t arm60_ror32(uint32_t v,uint32_t n){
{
n &= cast<uint32_t>(31ULL);
return cast<uint32_t>((shr<uint32_t>(v,n) | shl<uint32_t>(v,(cast<uint32_t>((cast<uint32_t>(32ULL) - n))))));
}
}
// tools/cpu/arm60/arm60.go:94:1
std::string arm60_cn(int64_t cond){
{
return arm60_condName[cond];
}
}
// tools/cpu/arm60/arm60.go:97:1
std::tuple<uint32_t,bool> arm60_word(Slice<uint8_t> code){
{
if ((len(code) < cast<int64_t>(4ULL))) {
return {cast<uint32_t>(0ULL),false};
}
return {cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((shl<uint32_t>(cast<uint32_t>(code[cast<int64_t>(0ULL)]),cast<int64_t>(24ULL)) | shl<uint32_t>(cast<uint32_t>(code[cast<int64_t>(1ULL)]),cast<int64_t>(16ULL)))) | shl<uint32_t>(cast<uint32_t>(code[cast<int64_t>(2ULL)]),cast<int64_t>(8ULL)))) | cast<uint32_t>(code[cast<int64_t>(3ULL)]))),true};
}
}
// tools/cpu/arm60/arm60.go:105:1
arm60_Inst arm60_Decode(Slice<uint8_t> code,uint32_t addr){
{
auto tmp1 = arm60_word(code);
uint32_t w = std::get<0>(tmp1);
bool ok = std::get<1>(tmp1);
if ((!ok)) {
return arm60_Inst{addr,len(code),std::string(".word",5),std::string(".word ; truncated",17),arm60_FlowStop,{},{},arm60_condAL};
}
int64_t cond = cast<int64_t>(shr<uint32_t>(w,cast<int64_t>(28ULL)));
arm60_Inst in = arm60_Inst{addr,cast<int64_t>(4ULL),{},{},arm60_FlowSeq,{},{},cond};
{
switch(cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(25ULL))) & cast<uint32_t>(7ULL)))){
case cast<uint32_t>(0ULL):{
in = arm60_decodeDataMisc(w,in,false);
break;}
case cast<uint32_t>(1ULL):{
in = arm60_decodeDataMisc(w,in,true);
break;}
case cast<uint32_t>(2ULL):{
in = arm60_decodeSingle(w,in,false);
break;}
case cast<uint32_t>(3ULL):{
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(4ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
in = arm60_undef(w,in);
}
else {
in = arm60_decodeSingle(w,in,true);
}
break;}
case cast<uint32_t>(4ULL):{
in = arm60_decodeBlock(w,in);
break;}
case cast<uint32_t>(5ULL):{
in = arm60_decodeBranch(w,addr,in);
break;}
case cast<uint32_t>(6ULL):{
in = arm60_decodeCoproLS(w,in);
break;}
default:{
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(24ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
in.Mnem = (std::string("SWI",3) + arm60_cn(cond));
in.Text = go_fmt_Sprintf(std::string("%s #0x%X",8),in.Mnem,cast<uint32_t>((w & cast<uint32_t>(16777215ULL))));
}
else {
in = arm60_decodeCopro(w,in);
}
break;}
}}
if (((cond == arm60_condNV) && (in.Flow != arm60_FlowStop))) {
auto tmp2 = std::make_tuple(arm60_FlowSeq,false);
in.Flow = std::get<0>(tmp2);
in.HasTarget = std::get<1>(tmp2);
}
return in;
}
}
// tools/cpu/arm60/arm60.go:149:1
arm60_Inst arm60_undef(uint32_t w,arm60_Inst in){
{
auto tmp3 = std::make_tuple(std::string(".word",5),arm60_FlowStop);
in.Mnem = std::get<0>(tmp3);
in.Flow = std::get<1>(tmp3);
in.Text = go_fmt_Sprintf(std::string(".word 0x%08X",12),w);
return in;
}
}
// tools/cpu/arm60/arm60.go:158:1
arm60_Inst arm60_decodeDataMisc(uint32_t w,arm60_Inst in,bool immForm){
{
uint32_t op = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(21ULL))) & cast<uint32_t>(15ULL)));
uint32_t s = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(20ULL))) & cast<uint32_t>(1ULL)));
if ((((!immForm) && (cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(4ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) && (cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(7ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL)))) {
return arm60_decodeExtension(w,in);
}
if ((((s == cast<uint32_t>(0ULL)) && (op >= cast<uint32_t>(8ULL))) && (op <= cast<uint32_t>(11ULL)))) {
return arm60_decodeMisc(w,in,immForm);
}
return arm60_decodeDataProc(w,in,immForm);
}
}
// tools/cpu/arm60/arm60.go:170:1
arm60_Inst arm60_decodeDataProc(uint32_t w,arm60_Inst in,bool immForm){
{
uint32_t op = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(21ULL))) & cast<uint32_t>(15ULL)));
uint32_t s = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(20ULL))) & cast<uint32_t>(1ULL)));
uint32_t rn = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(15ULL)));
uint32_t rd = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(12ULL))) & cast<uint32_t>(15ULL)));
std::string op2={};
if (immForm) {
op2 = arm60_imm(arm60_ror32(cast<uint32_t>((w & cast<uint32_t>(255ULL))),cast<uint32_t>(((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(8ULL))) & cast<uint32_t>(15ULL)))) * cast<uint32_t>(2ULL)))));
}
else {
op2 = arm60_shiftOperand(w);
}
std::string sfx = std::string("",0);
if ((s == cast<uint32_t>(1ULL))) {
sfx = std::string("S",1);
}
in.Mnem = ((arm60_dpOps[op] + arm60_cn(in.Cond)) + sfx);
{
switch(op){
case cast<uint32_t>(8ULL):case cast<uint32_t>(9ULL):case cast<uint32_t>(10ULL):case cast<uint32_t>(11ULL):{
in.Mnem = (arm60_dpOps[op] + arm60_cn(in.Cond));
in.Text = go_fmt_Sprintf(std::string("%s %s, %s",9),in.Mnem,arm60_regName[rn],op2);
break;}
case cast<uint32_t>(13ULL):case cast<uint32_t>(15ULL):{
in.Text = go_fmt_Sprintf(std::string("%s %s, %s",9),in.Mnem,arm60_regName[rd],op2);
break;}
default:{
in.Text = go_fmt_Sprintf(std::string("%s %s, %s, %s",13),in.Mnem,arm60_regName[rd],arm60_regName[rn],op2);
break;}
}}
if (((rd == cast<uint32_t>(15ULL)) && (!(((op >= cast<uint32_t>(8ULL)) && (op <= cast<uint32_t>(11ULL))))))) {
uint32_t rm = cast<uint32_t>((w & cast<uint32_t>(15ULL)));
bool isReg = ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(25ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(0ULL)) && (cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(4ULL))) & cast<uint32_t>(255ULL))) == cast<uint32_t>(0ULL)));
if ((((op == cast<uint32_t>(13ULL)) && isReg) && (rm == cast<uint32_t>(14ULL)))) {
in.Flow = arm60_FlowReturn;
}
else {
in.Flow = arm60_FlowIndJump;
}
}
return in;
}
}
// tools/cpu/arm60/arm60.go:210:1
std::string arm60_imm(uint32_t v){
{
if ((v < cast<uint32_t>(10ULL))) {
return go_fmt_Sprintf(std::string("#%d",3),v);
}
return go_fmt_Sprintf(std::string("#0x%X",5),v);
}
}
// tools/cpu/arm60/arm60.go:217:1
std::string arm60_shiftOperand(uint32_t w){
{
uint32_t rm = cast<uint32_t>((w & cast<uint32_t>(15ULL)));
uint32_t styp = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(5ULL))) & cast<uint32_t>(3ULL)));
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(4ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
uint32_t rs = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(8ULL))) & cast<uint32_t>(15ULL)));
return go_fmt_Sprintf(std::string("%s, %s %s",9),arm60_regName[rm],arm60_shiftName[styp],arm60_regName[rs]);
}
uint32_t amt = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(7ULL))) & cast<uint32_t>(31ULL)));
{
switch(styp){
case cast<uint32_t>(0ULL):{
if ((amt == cast<uint32_t>(0ULL))) {
return arm60_regName[rm];
}
return go_fmt_Sprintf(std::string("%s, LSL #%d",11),arm60_regName[rm],amt);
break;}
case cast<uint32_t>(1ULL):{
if ((amt == cast<uint32_t>(0ULL))) {
amt = cast<uint32_t>(32ULL);
}
return go_fmt_Sprintf(std::string("%s, LSR #%d",11),arm60_regName[rm],amt);
break;}
case cast<uint32_t>(2ULL):{
if ((amt == cast<uint32_t>(0ULL))) {
amt = cast<uint32_t>(32ULL);
}
return go_fmt_Sprintf(std::string("%s, ASR #%d",11),arm60_regName[rm],amt);
break;}
default:{
if ((amt == cast<uint32_t>(0ULL))) {
return go_fmt_Sprintf(std::string("%s, RRX",7),arm60_regName[rm]);
}
return go_fmt_Sprintf(std::string("%s, ROR #%d",11),arm60_regName[rm],amt);
break;}
}}
}
}
// tools/cpu/arm60/arm60.go:251:1
arm60_Inst arm60_decodeMisc(uint32_t w,arm60_Inst in,bool immForm){
{
uint32_t op = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(21ULL))) & cast<uint32_t>(15ULL)));
if (immForm) {
return arm60_decodeMSRimm(w,in);
}
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(4ULL))) & cast<uint32_t>(15ULL))) == cast<uint32_t>(0ULL))) {
if (((op == cast<uint32_t>(8ULL)) || (op == cast<uint32_t>(10ULL)))) {
std::string psr = std::string("CPSR",4);
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(22ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
psr = std::string("SPSR",4);
}
in.Mnem = (std::string("MRS",3) + arm60_cn(in.Cond));
in.Text = go_fmt_Sprintf(std::string("%s %s, %s",9),in.Mnem,arm60_regName[cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(12ULL))) & cast<uint32_t>(15ULL)))],psr);
return in;
}
return arm60_decodeMSRreg(w,in);
}
return arm60_undef(w,in);
}
}
// tools/cpu/arm60/arm60.go:271:1
arm60_Inst arm60_decodeMSRreg(uint32_t w,arm60_Inst in){
{
std::string psr = std::string("CPSR",4);
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(22ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
psr = std::string("SPSR",4);
}
in.Mnem = (std::string("MSR",3) + arm60_cn(in.Cond));
in.Text = go_fmt_Sprintf(std::string("%s %s%s, %s",11),in.Mnem,psr,arm60_msrFields(w),arm60_regName[cast<uint32_t>((w & cast<uint32_t>(15ULL)))]);
return in;
}
}
// tools/cpu/arm60/arm60.go:281:1
arm60_Inst arm60_decodeMSRimm(uint32_t w,arm60_Inst in){
{
std::string psr = std::string("CPSR",4);
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(22ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
psr = std::string("SPSR",4);
}
in.Mnem = (std::string("MSR",3) + arm60_cn(in.Cond));
in.Text = go_fmt_Sprintf(std::string("%s %s%s, %s",11),in.Mnem,psr,arm60_msrFields(w),arm60_imm(arm60_ror32(cast<uint32_t>((w & cast<uint32_t>(255ULL))),cast<uint32_t>(((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(8ULL))) & cast<uint32_t>(15ULL)))) * cast<uint32_t>(2ULL))))));
return in;
}
}
// tools/cpu/arm60/arm60.go:291:1
std::string arm60_msrFields(uint32_t w){
{
uint32_t m = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(15ULL)));
if ((m == cast<uint32_t>(0ULL))) {
return std::string("",0);
}
std::string s = std::string("_",1);
{auto&& tmp4 = Slice<std::string>{std::string("c",1),std::string("x",1),std::string("s",1),std::string("f",1)};
for(int64_t tmp5=0;tmp5<len(tmp4);++tmp5){
auto i=tmp5;auto f=tmp4[tmp5];if ((cast<uint32_t>((m & (shl<uint32_t>(cast<uint32_t>(1ULL),cast<uint64_t>(i))))) != cast<uint32_t>(0ULL))) {
s += f;
}
}}
return s;
}
}
// tools/cpu/arm60/arm60.go:307:1
arm60_Inst arm60_decodeExtension(uint32_t w,arm60_Inst in){
{
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(5ULL))) & cast<uint32_t>(3ULL))) != cast<uint32_t>(0ULL))) {
return arm60_undef(w,in);
}
{
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(24ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))){
std::string b = std::string("",0);
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(22ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
b = std::string("B",1);
}
in.Mnem = ((std::string("SWP",3) + arm60_cn(in.Cond)) + b);
in.Text = go_fmt_Sprintf(std::string("%s %s, %s, [%s]",15),in.Mnem,arm60_regName[cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(12ULL))) & cast<uint32_t>(15ULL)))],arm60_regName[cast<uint32_t>((w & cast<uint32_t>(15ULL)))],arm60_regName[cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(15ULL)))]);
return in;
}
else if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(23ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))){
std::array<std::string,4> names = std::array<std::string,4>{std::string("UMULL",5),std::string("UMLAL",5),std::string("SMULL",5),std::string("SMLAL",5)};
std::string s = std::string("",0);
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(20ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
s = std::string("S",1);
}
in.Mnem = ((names[cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(21ULL))) & cast<uint32_t>(3ULL)))] + arm60_cn(in.Cond)) + s);
in.Text = go_fmt_Sprintf(std::string("%s %s, %s, %s, %s",17),in.Mnem,arm60_regName[cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(12ULL))) & cast<uint32_t>(15ULL)))],arm60_regName[cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(15ULL)))],arm60_regName[cast<uint32_t>((w & cast<uint32_t>(15ULL)))],arm60_regName[cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(8ULL))) & cast<uint32_t>(15ULL)))]);
return in;
}
else {
std::string s = std::string("",0);
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(20ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
s = std::string("S",1);
}
uint32_t rd = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(15ULL)));
uint32_t rn = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(12ULL))) & cast<uint32_t>(15ULL)));
uint32_t rs = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(8ULL))) & cast<uint32_t>(15ULL)));
uint32_t rm = cast<uint32_t>((w & cast<uint32_t>(15ULL)));
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(21ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
in.Mnem = ((std::string("MLA",3) + arm60_cn(in.Cond)) + s);
in.Text = go_fmt_Sprintf(std::string("%s %s, %s, %s, %s",17),in.Mnem,arm60_regName[rd],arm60_regName[rm],arm60_regName[rs],arm60_regName[rn]);
}
else {
in.Mnem = ((std::string("MUL",3) + arm60_cn(in.Cond)) + s);
in.Text = go_fmt_Sprintf(std::string("%s %s, %s, %s",13),in.Mnem,arm60_regName[rd],arm60_regName[rm],arm60_regName[rs]);
}
return in;
}
}
tmp6:;
}
}
// tools/cpu/arm60/arm60.go:350:1
arm60_Inst arm60_decodeSingle(uint32_t w,arm60_Inst in,bool regOff){
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
in.Mnem = (name + arm60_cn(in.Cond));
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
{
uint32_t v = cast<uint32_t>((w & cast<uint32_t>(4095ULL)));
if ((v != cast<uint32_t>(0ULL))) {
off = go_fmt_Sprintf(std::string(", #%s0x%X",9),sign,v);
}
}
}
else {
off = ((std::string(", ",2) + sign) + arm60_shiftOperand(w));
}
in.Text = arm60_addrForm(in.Mnem,rd,rn,off,p,cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(21ULL))) & cast<uint32_t>(1ULL))));
if (((l == cast<uint32_t>(1ULL)) && (rd == cast<uint32_t>(15ULL)))) {
in.Flow = arm60_FlowIndJump;
}
return in;
}
}
// tools/cpu/arm60/arm60.go:385:1
std::string arm60_addrForm(std::string mnem,uint32_t rd,uint32_t rn,std::string off,uint32_t p,uint32_t wb){
{
if ((p == cast<uint32_t>(1ULL))) {
std::string bang = std::string("",0);
if ((wb == cast<uint32_t>(1ULL))) {
bang = std::string("!",1);
}
return go_fmt_Sprintf(std::string("%s %s, [%s%s]%s",15),mnem,arm60_regName[rd],arm60_regName[rn],off,bang);
}
return go_fmt_Sprintf(std::string("%s %s, [%s]%s",13),mnem,arm60_regName[rd],arm60_regName[rn],off);
}
}
// tools/cpu/arm60/arm60.go:396:1
arm60_Inst arm60_decodeBlock(uint32_t w,arm60_Inst in){
{
uint32_t l = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(20ULL))) & cast<uint32_t>(1ULL)));
uint32_t p = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(24ULL))) & cast<uint32_t>(1ULL)));
uint32_t u = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(23ULL))) & cast<uint32_t>(1ULL)));
std::string base = std::string("STM",3);
if ((l == cast<uint32_t>(1ULL))) {
base = std::string("LDM",3);
}
std::string mode = std::array<std::string,4>{std::string("DA",2),std::string("DB",2),std::string("IA",2),std::string("IB",2)}[cast<uint32_t>((shl<uint32_t>(u,cast<int64_t>(1ULL)) | p))];
in.Mnem = ((base + arm60_cn(in.Cond)) + mode);
uint32_t rn = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(15ULL)));
std::string bang = std::string("",0);
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(21ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
bang = std::string("!",1);
}
std::string usr = std::string("",0);
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(22ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
usr = std::string("^",1);
}
in.Text = go_fmt_Sprintf(std::string("%s %s%s, {%s}%s",15),in.Mnem,arm60_regName[rn],bang,arm60_regList(cast<uint32_t>((w & cast<uint32_t>(65535ULL)))),usr);
if (((l == cast<uint32_t>(1ULL)) && (cast<uint32_t>((w & (shl<uint32_t>(cast<int64_t>(1ULL),cast<int64_t>(15ULL))))) != cast<uint32_t>(0ULL)))) {
in.Flow = arm60_FlowReturn;
}
return in;
}
}
// tools/cpu/arm60/arm60.go:423:1
std::string arm60_regList(uint32_t mask){
{
Slice<std::string> parts={};
for (int64_t i = cast<int64_t>(0ULL);(i < cast<int64_t>(16ULL));){
if ((cast<uint32_t>((mask & (shl<uint32_t>(cast<uint32_t>(1ULL),cast<uint64_t>(i))))) == cast<uint32_t>(0ULL))) {
i++;
continue;
}
int64_t j = i;
for (;((cast<int64_t>((j + cast<int64_t>(1ULL))) < cast<int64_t>(16ULL)) && (cast<uint32_t>((mask & (shl<uint32_t>(cast<uint32_t>(1ULL),cast<uint64_t>(cast<int64_t>((j + cast<int64_t>(1ULL)))))))) != cast<uint32_t>(0ULL)));){
j++;
}
{
if ((j == i)){
parts = append(parts,arm60_regName[i]);
}
else if ((j == cast<int64_t>((i + cast<int64_t>(1ULL))))){
parts = append(parts,arm60_regName[i],arm60_regName[j]);
}
else {
parts = append(parts,((arm60_regName[i] + std::string("-",1)) + arm60_regName[j]));
}
}
tmp7:;
i = cast<int64_t>((j + cast<int64_t>(1ULL)));
}
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
// tools/cpu/arm60/arm60.go:454:1
arm60_Inst arm60_decodeBranch(uint32_t w,uint32_t addr,arm60_Inst in){
{
uint32_t off = shl<uint32_t>(arm60_signExtend(cast<uint32_t>((w & cast<uint32_t>(16777215ULL))),cast<uint64_t>(24ULL)),cast<int64_t>(2ULL));
uint32_t target = cast<uint32_t>((cast<uint32_t>((addr + cast<uint32_t>(8ULL))) + off));
auto tmp10 = std::make_tuple(target,true);
in.Target = std::get<0>(tmp10);
in.HasTarget = std::get<1>(tmp10);
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(24ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
in.Mnem = (std::string("BL",2) + arm60_cn(in.Cond));
in.Text = go_fmt_Sprintf(std::string("%s 0x%08X",9),in.Mnem,target);
in.Flow = arm60_FlowCall;
return in;
}
in.Mnem = (std::string("B",1) + arm60_cn(in.Cond));
in.Text = go_fmt_Sprintf(std::string("%s 0x%08X",9),in.Mnem,target);
if (((in.Cond == arm60_condAL) || (in.Cond == arm60_condNV))) {
in.Flow = arm60_FlowJump;
}
else {
in.Flow = arm60_FlowBranch;
}
return in;
}
}
// tools/cpu/arm60/arm60.go:474:1
arm60_Inst arm60_decodeCoproLS(uint32_t w,arm60_Inst in){
{
uint32_t l = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(20ULL))) & cast<uint32_t>(1ULL)));
std::string name = std::string("STC",3);
if ((l == cast<uint32_t>(1ULL))) {
name = std::string("LDC",3);
}
in.Mnem = (name + arm60_cn(in.Cond));
in.Text = go_fmt_Sprintf(std::string("%s p%d, c%d, [%s]",17),in.Mnem,cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(8ULL))) & cast<uint32_t>(15ULL))),cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(12ULL))) & cast<uint32_t>(15ULL))),arm60_regName[cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(15ULL)))]);
return in;
}
}
// tools/cpu/arm60/arm60.go:485:1
arm60_Inst arm60_decodeCopro(uint32_t w,arm60_Inst in){
{
uint32_t cp = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(8ULL))) & cast<uint32_t>(15ULL)));
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(4ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(0ULL))) {
in.Mnem = (std::string("CDP",3) + arm60_cn(in.Cond));
in.Text = go_fmt_Sprintf(std::string("%s p%d, #%d, c%d, c%d, c%d",26),in.Mnem,cp,cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(20ULL))) & cast<uint32_t>(15ULL))),cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(12ULL))) & cast<uint32_t>(15ULL))),cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(15ULL))),cast<uint32_t>((w & cast<uint32_t>(15ULL))));
return in;
}
std::string name = std::string("MCR",3);
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(20ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
name = std::string("MRC",3);
}
in.Mnem = (name + arm60_cn(in.Cond));
in.Text = go_fmt_Sprintf(std::string("%s p%d, #%d, %s, c%d, c%d, #%d",30),in.Mnem,cp,cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(21ULL))) & cast<uint32_t>(7ULL))),arm60_regName[cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(12ULL))) & cast<uint32_t>(15ULL)))],cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(15ULL))),cast<uint32_t>((w & cast<uint32_t>(15ULL))),cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(5ULL))) & cast<uint32_t>(7ULL))));
return in;
}
}
// tools/cpu/arm60/arm60.go:502:1
Slice<std::string> arm60_Disassemble(Slice<uint8_t> code,uint32_t base){
{
Slice<std::string> out={};
for (int64_t off = cast<int64_t>(0ULL);(cast<int64_t>((off + cast<int64_t>(4ULL))) <= len(code));off += cast<int64_t>(4ULL)){
arm60_Inst in = arm60_Decode(sub(code,off,len(code)),cast<uint32_t>((base + cast<uint32_t>(off))));
out = append(out,go_fmt_Sprintf(std::string("%08X  %s",8),in.Addr,in.Text));
}
return out;
}
}
// tools/cpu/arm60/cpu.go:52:1
arm60_CPU* arm60_NewCPU(threedo_Machine* bus){
{
arm60_CPU* c = arenaNew(arm60_CPU{{},{},{},{},{},{},{},{},{},{},{},{},{},{},bus,{},{},{},{},{}});
arm60_CPU_Reset(c);
return c;
}
}
// tools/cpu/arm60/cpu.go:59:1
void arm60_CPU_Reset(arm60_CPU* c){
{
c->R = std::array<uint32_t,16>{};
auto tmp11 = std::make_tuple(false,false,false,false);
c->N = std::get<0>(tmp11);
c->Z = std::get<1>(tmp11);
c->C = std::get<2>(tmp11);
c->V = std::get<3>(tmp11);
auto tmp12 = std::make_tuple(true,true);
c->IRQDisable = std::get<0>(tmp12);
c->FIQDisable = std::get<1>(tmp12);
c->Mode = arm60_ModeSVC;
auto tmp13 = std::make_tuple(false,std::string("",0));
c->Halted = std::get<0>(tmp13);
c->HaltReason = std::get<1>(tmp13);
}
}
// tools/cpu/arm60/cpu.go:74:1
uint32_t arm60_CPU_CurPC(arm60_CPU* c){
{
return c->cur;
}
}
// tools/cpu/arm60/cpu.go:95:1
arm60_Context arm60_CPU_SaveContext(arm60_CPU* c){
{
return arm60_Context{c->R,c->N,c->Z,c->C,c->V,c->IRQDisable,c->FIQDisable,c->Mode,c->bankR13,c->bankR14,c->bankSPSR,c->fiqR8_12,c->usrR8_12};
}
}
// tools/cpu/arm60/cpu.go:105:1
void arm60_CPU_RestoreContext(arm60_CPU* c,arm60_Context x){
{
auto tmp14 = std::make_tuple(x.R,x.N,x.Z,x.C,x.V);
c->R = std::get<0>(tmp14);
c->N = std::get<1>(tmp14);
c->Z = std::get<2>(tmp14);
c->C = std::get<3>(tmp14);
c->V = std::get<4>(tmp14);
auto tmp15 = std::make_tuple(x.IRQDisable,x.FIQDisable,x.Mode);
c->IRQDisable = std::get<0>(tmp15);
c->FIQDisable = std::get<1>(tmp15);
c->Mode = std::get<2>(tmp15);
auto tmp16 = std::make_tuple(x.BankR13,x.BankR14,x.BankSPSR);
c->bankR13 = std::get<0>(tmp16);
c->bankR14 = std::get<1>(tmp16);
c->bankSPSR = std::get<2>(tmp16);
auto tmp17 = std::make_tuple(x.FIQR8_12,x.USRR8_12);
c->fiqR8_12 = std::get<0>(tmp17);
c->usrR8_12 = std::get<1>(tmp17);
c->branched = false;
}
}
// tools/cpu/arm60/cpu.go:129:1
arm60_CPUState arm60_CPU_SaveState(arm60_CPU* c){
{
return arm60_CPUState{arm60_CPU_SaveContext(c),c->Halted,c->HaltReason,c->Instrs,c->cur};
}
}
// tools/cpu/arm60/cpu.go:137:1
void arm60_CPU_LoadState(arm60_CPU* c,arm60_CPUState s){
{
arm60_CPU_RestoreContext(c,s.Ctx);
auto tmp18 = std::make_tuple(s.Halted,s.HaltReason);
c->Halted = std::get<0>(tmp18);
c->HaltReason = std::get<1>(tmp18);
auto tmp19 = std::make_tuple(s.Instrs,s.Cur);
c->Instrs = std::get<0>(tmp19);
c->cur = std::get<1>(tmp19);
}
}
// tools/cpu/arm60/cpu.go:144:1
void arm60_CPU_SetPC(arm60_CPU* c,uint32_t v){
{
c->R[cast<int64_t>(15ULL)] = v;
}
}
// tools/cpu/arm60/cpu.go:145:1
void arm60_CPU_SetReg(arm60_CPU* c,uint32_t i,uint32_t v){
{
c->R[cast<uint32_t>((i & cast<uint32_t>(15ULL)))] = v;
}
}
// tools/cpu/arm60/cpu.go:146:1
uint32_t arm60_CPU_Reg(arm60_CPU* c,uint32_t i){
{
return c->R[cast<uint32_t>((i & cast<uint32_t>(15ULL)))];
}
}
// tools/cpu/arm60/cpu.go:150:1
uint32_t arm60_CPU_CPSR(arm60_CPU* c){
{
uint32_t v={};
if (c->N) {
v |= shl<uint32_t>(cast<int64_t>(1ULL),cast<int64_t>(31ULL));
}
if (c->Z) {
v |= shl<uint32_t>(cast<int64_t>(1ULL),cast<int64_t>(30ULL));
}
if (c->C) {
v |= shl<uint32_t>(cast<int64_t>(1ULL),cast<int64_t>(29ULL));
}
if (c->V) {
v |= shl<uint32_t>(cast<int64_t>(1ULL),cast<int64_t>(28ULL));
}
if (c->IRQDisable) {
v |= shl<uint32_t>(cast<int64_t>(1ULL),cast<int64_t>(7ULL));
}
if (c->FIQDisable) {
v |= shl<uint32_t>(cast<int64_t>(1ULL),cast<int64_t>(6ULL));
}
return cast<uint32_t>((v | (cast<uint32_t>((c->Mode & cast<uint32_t>(31ULL))))));
}
}
// tools/cpu/arm60/cpu.go:173:1
void arm60_CPU_SetCPSR(arm60_CPU* c,uint32_t v){
{
c->N = (cast<uint32_t>((v & (shl<uint32_t>(cast<int64_t>(1ULL),cast<int64_t>(31ULL))))) != cast<uint32_t>(0ULL));
c->Z = (cast<uint32_t>((v & (shl<uint32_t>(cast<int64_t>(1ULL),cast<int64_t>(30ULL))))) != cast<uint32_t>(0ULL));
c->C = (cast<uint32_t>((v & (shl<uint32_t>(cast<int64_t>(1ULL),cast<int64_t>(29ULL))))) != cast<uint32_t>(0ULL));
c->V = (cast<uint32_t>((v & (shl<uint32_t>(cast<int64_t>(1ULL),cast<int64_t>(28ULL))))) != cast<uint32_t>(0ULL));
c->IRQDisable = (cast<uint32_t>((v & (shl<uint32_t>(cast<int64_t>(1ULL),cast<int64_t>(7ULL))))) != cast<uint32_t>(0ULL));
c->FIQDisable = (cast<uint32_t>((v & (shl<uint32_t>(cast<int64_t>(1ULL),cast<int64_t>(6ULL))))) != cast<uint32_t>(0ULL));
arm60_CPU_switchMode(c,cast<uint32_t>((v & cast<uint32_t>(31ULL))));
}
}
// tools/cpu/arm60/cpu.go:183:1
int64_t arm60_modeIndex(uint32_t mode){
{
{
switch(mode){
case arm60_ModeFIQ:{
return cast<int64_t>(1ULL);
break;}
case arm60_ModeIRQ:{
return cast<int64_t>(2ULL);
break;}
case arm60_ModeSVC:{
return cast<int64_t>(3ULL);
break;}
case arm60_ModeABT:{
return cast<int64_t>(4ULL);
break;}
case arm60_ModeUND:{
return cast<int64_t>(5ULL);
break;}
default:{
return cast<int64_t>(0ULL);
break;}
}}
}
}
// tools/cpu/arm60/cpu.go:200:1
void arm60_CPU_switchMode(arm60_CPU* c,uint32_t mode){
{
mode &= cast<uint32_t>(31ULL);
if ((mode == c->Mode)) {
return ;
}
auto tmp20 = std::make_tuple(arm60_modeIndex(c->Mode),arm60_modeIndex(mode));
int64_t from = std::get<0>(tmp20);
int64_t to = std::get<1>(tmp20);
c->bankR13[from] = c->R[cast<int64_t>(13ULL)];
c->bankR14[from] = c->R[cast<int64_t>(14ULL)];
if ((c->Mode == arm60_ModeFIQ)) {
gcopy(sub(c->fiqR8_12,0,len(c->fiqR8_12)),sub(c->R,cast<int64_t>(8ULL),cast<int64_t>(13ULL)));
}
else {
gcopy(sub(c->usrR8_12,0,len(c->usrR8_12)),sub(c->R,cast<int64_t>(8ULL),cast<int64_t>(13ULL)));
}
c->R[cast<int64_t>(13ULL)] = c->bankR13[to];
c->R[cast<int64_t>(14ULL)] = c->bankR14[to];
if ((mode == arm60_ModeFIQ)) {
gcopy(sub(c->R,cast<int64_t>(8ULL),cast<int64_t>(13ULL)),sub(c->fiqR8_12,0,len(c->fiqR8_12)));
}
else {
gcopy(sub(c->R,cast<int64_t>(8ULL),cast<int64_t>(13ULL)),sub(c->usrR8_12,0,len(c->usrR8_12)));
}
c->Mode = mode;
}
}
// tools/cpu/arm60/cpu.go:223:1
uint32_t arm60_CPU_SPSR(arm60_CPU* c){
{
return c->bankSPSR[arm60_modeIndex(c->Mode)];
}
}
// tools/cpu/arm60/cpu.go:224:1
void arm60_CPU_SetSPSR(arm60_CPU* c,uint32_t v){
{
c->bankSPSR[arm60_modeIndex(c->Mode)] = v;
}
}
// tools/cpu/arm60/cpu.go:228:1
uint8_t arm60_CPU_read8(arm60_CPU* c,uint32_t a){
{
return threedo_Machine_Read(c->bus,a);
}
}
// tools/cpu/arm60/cpu.go:229:1
void arm60_CPU_write8(arm60_CPU* c,uint32_t a,uint8_t v){
{
threedo_Machine_Write(c->bus,a,v);
}
}
// tools/cpu/arm60/cpu.go:238:1
uint32_t arm60_CPU_read32(arm60_CPU* c,uint32_t a){
{
uint32_t v = arm60_CPU_read32aligned(c,a);
{
uint32_t r = cast<uint32_t>(((cast<uint32_t>((a & cast<uint32_t>(3ULL)))) * cast<uint32_t>(8ULL)));
if ((r != cast<uint32_t>(0ULL))) {
v = arm60_ror32(v,r);
}
}
return v;
}
}
// tools/cpu/arm60/cpu.go:258:1
uint32_t arm60_CPU_reg(arm60_CPU* c,uint32_t i){
{
if ((i == cast<uint32_t>(15ULL))) {
return cast<uint32_t>((c->cur + cast<uint32_t>(8ULL)));
}
return c->R[i];
}
}
// tools/cpu/arm60/cpu.go:265:1
void arm60_CPU_setReg(arm60_CPU* c,uint32_t i,uint32_t v){
{
c->R[i] = v;
if ((i == cast<uint32_t>(15ULL))) {
c->branched = true;
}
}
}
// tools/cpu/arm60/cpu.go:274:1
void arm60_CPU_setNZ(arm60_CPU* c,uint32_t v){
{
c->Z = (v == cast<uint32_t>(0ULL));
c->N = (cast<uint32_t>((v & (shl<uint32_t>(cast<int64_t>(1ULL),cast<int64_t>(31ULL))))) != cast<uint32_t>(0ULL));
}
}
// tools/cpu/arm60/cpu.go:279:1
uint32_t arm60_CPU_add(arm60_CPU* c,uint32_t a,uint32_t b,uint32_t cin){
{
uint64_t r = cast<uint64_t>((cast<uint64_t>((cast<uint64_t>(a) + cast<uint64_t>(b))) + cast<uint64_t>(cin)));
uint32_t res = cast<uint32_t>(r);
arm60_CPU_setNZ(c,res);
c->C = (r > cast<uint64_t>(4294967295ULL));
c->V = (cast<uint32_t>((cast<uint32_t>(((cast<uint32_t>((a ^ res))) & (cast<uint32_t>((b ^ res))))) & cast<uint32_t>(2147483648ULL))) != cast<uint32_t>(0ULL));
return res;
}
}
// tools/cpu/arm60/cpu.go:288:1
uint32_t arm60_CPU_sub(arm60_CPU* c,uint32_t a,uint32_t b,uint32_t cin){
{
return arm60_CPU_add(c,a,cast<uint32_t>(~b),cin);
}
}
// tools/cpu/arm60/cpu.go:290:1
uint32_t arm60_CPU_boolToU(arm60_CPU* c,bool b){
{
if (b) {
return cast<uint32_t>(1ULL);
}
return cast<uint32_t>(0ULL);
}
}
// tools/cpu/arm60/cpu.go:297:1
bool arm60_CPU_cond(arm60_CPU* c,int64_t cc){
{
{
switch(cc){
case arm60_condEQ:{
return c->Z;
break;}
case arm60_condNE:{
return (!c->Z);
break;}
case arm60_condCS:{
return c->C;
break;}
case arm60_condCC:{
return (!c->C);
break;}
case arm60_condMI:{
return c->N;
break;}
case arm60_condPL:{
return (!c->N);
break;}
case arm60_condVS:{
return c->V;
break;}
case arm60_condVC:{
return (!c->V);
break;}
case arm60_condHI:{
return (c->C && (!c->Z));
break;}
case arm60_condLS:{
return ((!c->C) || c->Z);
break;}
case arm60_condGE:{
return (c->N == c->V);
break;}
case arm60_condLT:{
return (c->N != c->V);
break;}
case arm60_condGT:{
return ((!c->Z) && (c->N == c->V));
break;}
case arm60_condLE:{
return (c->Z || (c->N != c->V));
break;}
case arm60_condNV:{
return false;
break;}
default:{
return true;
break;}
}}
}
}
// tools/cpu/arm60/cpu.go:336:1
std::tuple<uint32_t,uint32_t> arm60_CPU_shift(arm60_CPU* c,uint32_t typ,uint32_t amt,uint32_t val,bool regForm,uint32_t cin){
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
tmp21:;
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
tmp22:;
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
if ((cast<uint32_t>((val & (shl<uint32_t>(cast<int64_t>(1ULL),cast<int64_t>(31ULL))))) != cast<uint32_t>(0ULL))) {
return {cast<uint32_t>(4294967295ULL),cast<uint32_t>(1ULL)};
}
return {cast<uint32_t>(0ULL),cast<uint32_t>(0ULL)};
}
}
tmp23:;
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
return {arm60_ror32(val,amt),cast<uint32_t>(((shr<uint32_t>(val,(cast<uint32_t>((amt - cast<uint32_t>(1ULL)))))) & cast<uint32_t>(1ULL)))};
break;}
}}
}
}
// tools/cpu/arm60/exec.go:8:1
int64_t arm60_CPU_Step(arm60_CPU* c){
{
if (c->Halted) {
return cast<int64_t>(0ULL);
}
c->cur = c->R[cast<int64_t>(15ULL)];
c->branched = false;
c->Instrs++;
uint32_t w = arm60_CPU_read32aligned(c,c->cur);
int64_t cond = cast<int64_t>(shr<uint32_t>(w,cast<int64_t>(28ULL)));
if ((!arm60_CPU_cond(c,cond))) {
c->R[cast<int64_t>(15ULL)] = cast<uint32_t>((c->cur + cast<uint32_t>(4ULL)));
return cast<int64_t>(1ULL);
}
arm60_CPU_execARM(c,w);
if ((!c->branched)) {
c->R[cast<int64_t>(15ULL)] = cast<uint32_t>((c->cur + cast<uint32_t>(4ULL)));
}
return cast<int64_t>(1ULL);
}
}
// tools/cpu/arm60/exec.go:29:1
void arm60_CPU_execARM(arm60_CPU* c,uint32_t w){
{
{
switch(cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(25ULL))) & cast<uint32_t>(7ULL)))){
case cast<uint32_t>(0ULL):{
arm60_CPU_execDataMisc(c,w,false);
break;}
case cast<uint32_t>(1ULL):{
arm60_CPU_execDataMisc(c,w,true);
break;}
case cast<uint32_t>(2ULL):{
arm60_CPU_execSingle(c,w,false);
break;}
case cast<uint32_t>(3ULL):{
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(4ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
arm60_CPU_Halt(c,std::string("undefined instruction 0x%08X at 0x%08X",38),w,c->cur);
return ;
}
arm60_CPU_execSingle(c,w,true);
break;}
case cast<uint32_t>(4ULL):{
arm60_CPU_execBlock(c,w);
break;}
case cast<uint32_t>(5ULL):{
arm60_CPU_execBranch(c,w);
break;}
case cast<uint32_t>(6ULL):{
arm60_CPU_Halt(c,std::string("unimplemented coprocessor transfer 0x%08X at 0x%08X",51),w,c->cur);
break;}
default:{
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(24ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
arm60_CPU_execSWI(c,w);
}
else {
arm60_CPU_execCopro(c,w);
}
break;}
}}
}
}
// tools/cpu/arm60/exec.go:58:1
void arm60_CPU_execDataMisc(arm60_CPU* c,uint32_t w,bool immForm){
{
uint32_t op = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(21ULL))) & cast<uint32_t>(15ULL)));
uint32_t s = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(20ULL))) & cast<uint32_t>(1ULL)));
if ((((!immForm) && (cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(4ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) && (cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(7ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL)))) {
arm60_CPU_execExtension(c,w);
return ;
}
if ((((s == cast<uint32_t>(0ULL)) && (op >= cast<uint32_t>(8ULL))) && (op <= cast<uint32_t>(11ULL)))) {
arm60_CPU_execMisc(c,w,immForm);
return ;
}
arm60_CPU_execDataProc(c,w,immForm);
}
}
// tools/cpu/arm60/exec.go:72:1
std::tuple<uint32_t,uint32_t> arm60_CPU_dpOperand(arm60_CPU* c,uint32_t w,bool immForm){
{
if (immForm) {
uint32_t rot = cast<uint32_t>(((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(8ULL))) & cast<uint32_t>(15ULL)))) * cast<uint32_t>(2ULL)));
uint32_t v = arm60_ror32(cast<uint32_t>((w & cast<uint32_t>(255ULL))),rot);
if ((rot == cast<uint32_t>(0ULL))) {
return {v,arm60_CPU_boolToU(c,c->C)};
}
return {v,cast<uint32_t>(((shr<uint32_t>(v,cast<int64_t>(31ULL))) & cast<uint32_t>(1ULL)))};
}
uint32_t rm = arm60_CPU_reg(c,cast<uint32_t>((w & cast<uint32_t>(15ULL))));
uint32_t typ = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(5ULL))) & cast<uint32_t>(3ULL)));
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(4ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
return arm60_CPU_shift(c,typ,arm60_CPU_reg(c,cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(8ULL))) & cast<uint32_t>(15ULL)))),rm,true,arm60_CPU_boolToU(c,c->C));
}
return arm60_CPU_shift(c,typ,cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(7ULL))) & cast<uint32_t>(31ULL))),rm,false,arm60_CPU_boolToU(c,c->C));
}
}
// tools/cpu/arm60/exec.go:89:1
void arm60_CPU_execDataProc(arm60_CPU* c,uint32_t w,bool immForm){
{
uint32_t op = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(21ULL))) & cast<uint32_t>(15ULL)));
bool setFlags = (cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(20ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL));
uint32_t rn = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(15ULL)));
uint32_t rd = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(12ULL))) & cast<uint32_t>(15ULL)));
auto tmp24 = arm60_CPU_dpOperand(c,w,immForm);
uint32_t op2 = std::get<0>(tmp24);
uint32_t shc = std::get<1>(tmp24);
uint32_t a = arm60_CPU_reg(c,rn);
auto tmp25 = std::make_tuple(c->N,c->Z,c->C,c->V);
bool sn = std::get<0>(tmp25);
bool sz = std::get<1>(tmp25);
bool sc = std::get<2>(tmp25);
bool sv = std::get<3>(tmp25);
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
auto tmp26 = std::make_tuple(arm60_CPU_sub(c,a,op2,cast<uint32_t>(1ULL)),false);
res = std::get<0>(tmp26);
logical = std::get<1>(tmp26);
break;}
case cast<uint32_t>(3ULL):{
auto tmp27 = std::make_tuple(arm60_CPU_sub(c,op2,a,cast<uint32_t>(1ULL)),false);
res = std::get<0>(tmp27);
logical = std::get<1>(tmp27);
break;}
case cast<uint32_t>(4ULL):{
auto tmp28 = std::make_tuple(arm60_CPU_add(c,a,op2,cast<uint32_t>(0ULL)),false);
res = std::get<0>(tmp28);
logical = std::get<1>(tmp28);
break;}
case cast<uint32_t>(5ULL):{
auto tmp29 = std::make_tuple(arm60_CPU_add(c,a,op2,arm60_CPU_boolToU(c,sc)),false);
res = std::get<0>(tmp29);
logical = std::get<1>(tmp29);
break;}
case cast<uint32_t>(6ULL):{
auto tmp30 = std::make_tuple(arm60_CPU_sub(c,a,op2,arm60_CPU_boolToU(c,sc)),false);
res = std::get<0>(tmp30);
logical = std::get<1>(tmp30);
break;}
case cast<uint32_t>(7ULL):{
auto tmp31 = std::make_tuple(arm60_CPU_sub(c,op2,a,arm60_CPU_boolToU(c,sc)),false);
res = std::get<0>(tmp31);
logical = std::get<1>(tmp31);
break;}
case cast<uint32_t>(8ULL):{
res = cast<uint32_t>((a & op2));
break;}
case cast<uint32_t>(9ULL):{
res = cast<uint32_t>((a ^ op2));
break;}
case cast<uint32_t>(10ULL):{
arm60_CPU_sub(c,a,op2,cast<uint32_t>(1ULL));
logical = false;
break;}
case cast<uint32_t>(11ULL):{
arm60_CPU_add(c,a,op2,cast<uint32_t>(0ULL));
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
arm60_CPU_setNZ(c,res);
auto tmp32 = std::make_tuple((shc == cast<uint32_t>(1ULL)),sv);
c->C = std::get<0>(tmp32);
c->V = std::get<1>(tmp32);
}
return ;
}
if (setFlags) {
if (logical) {
arm60_CPU_setNZ(c,res);
auto tmp33 = std::make_tuple((shc == cast<uint32_t>(1ULL)),sv);
c->C = std::get<0>(tmp33);
c->V = std::get<1>(tmp33);
}
}
else {
auto tmp34 = std::make_tuple(sn,sz,sc,sv);
c->N = std::get<0>(tmp34);
c->Z = std::get<1>(tmp34);
c->C = std::get<2>(tmp34);
c->V = std::get<3>(tmp34);
}
if ((rd == cast<uint32_t>(15ULL))) {
if (setFlags) {
arm60_CPU_SetCPSR(c,arm60_CPU_SPSR(c));
}
c->R[cast<int64_t>(15ULL)] = (res & ~(cast<uint32_t>(3ULL)));
c->branched = true;
return ;
}
arm60_CPU_setReg(c,rd,res);
}
}
// tools/cpu/arm60/exec.go:165:1
void arm60_CPU_execBranch(arm60_CPU* c,uint32_t w){
{
uint32_t off = shl<uint32_t>(arm60_signExtend(cast<uint32_t>((w & cast<uint32_t>(16777215ULL))),cast<uint64_t>(24ULL)),cast<int64_t>(2ULL));
uint32_t target = cast<uint32_t>((cast<uint32_t>((c->cur + cast<uint32_t>(8ULL))) + off));
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(24ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
c->R[cast<int64_t>(14ULL)] = cast<uint32_t>((c->cur + cast<uint32_t>(4ULL)));
}
c->R[cast<int64_t>(15ULL)] = (target & ~(cast<uint32_t>(3ULL)));
c->branched = true;
}
}
// tools/cpu/arm60/exec.go:175:1
void arm60_CPU_execExtension(arm60_CPU* c,uint32_t w){
{
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(5ULL))) & cast<uint32_t>(3ULL))) != cast<uint32_t>(0ULL))) {
arm60_CPU_Halt(c,std::string("undefined halfword/signed transfer 0x%08X at 0x%08X",51),w,c->cur);
return ;
}
{
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(24ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))){
arm60_CPU_execSwap(c,w);
}
else if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(23ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))){
arm60_CPU_execMulLong(c,w);
}
else {
arm60_CPU_execMul(c,w);
}
}
tmp35:;
}
}
// tools/cpu/arm60/exec.go:190:1
void arm60_CPU_execMul(arm60_CPU* c,uint32_t w){
{
uint32_t rd = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(15ULL)));
uint32_t rn = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(12ULL))) & cast<uint32_t>(15ULL)));
uint32_t rs = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(8ULL))) & cast<uint32_t>(15ULL)));
uint32_t rm = cast<uint32_t>((w & cast<uint32_t>(15ULL)));
uint32_t res = cast<uint32_t>((arm60_CPU_reg(c,rm) * arm60_CPU_reg(c,rs)));
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(21ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
res += arm60_CPU_reg(c,rn);
}
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(20ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
arm60_CPU_setNZ(c,res);
}
arm60_CPU_setReg(c,rd,res);
}
}
// tools/cpu/arm60/exec.go:205:1
void arm60_CPU_execMulLong(arm60_CPU* c,uint32_t w){
{
uint32_t rdHi = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(15ULL)));
uint32_t rdLo = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(12ULL))) & cast<uint32_t>(15ULL)));
uint32_t rs = arm60_CPU_reg(c,cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(8ULL))) & cast<uint32_t>(15ULL))));
uint32_t rm = arm60_CPU_reg(c,cast<uint32_t>((w & cast<uint32_t>(15ULL))));
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
result += cast<uint64_t>((shl<uint64_t>(cast<uint64_t>(arm60_CPU_reg(c,rdHi)),cast<int64_t>(32ULL)) | cast<uint64_t>(arm60_CPU_reg(c,rdLo))));
}
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(20ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
c->Z = (result == cast<uint64_t>(0ULL));
c->N = (cast<uint64_t>((result & (shl<uint64_t>(cast<int64_t>(1ULL),cast<int64_t>(63ULL))))) != cast<uint64_t>(0ULL));
}
arm60_CPU_setReg(c,rdLo,cast<uint32_t>(result));
arm60_CPU_setReg(c,rdHi,cast<uint32_t>(shr<uint64_t>(result,cast<int64_t>(32ULL))));
}
}
// tools/cpu/arm60/exec.go:230:1
void arm60_CPU_execSwap(arm60_CPU* c,uint32_t w){
{
uint32_t rn = arm60_CPU_reg(c,cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(15ULL))));
uint32_t rd = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(12ULL))) & cast<uint32_t>(15ULL)));
uint32_t rm = cast<uint32_t>((w & cast<uint32_t>(15ULL)));
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(22ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
uint32_t old = cast<uint32_t>(arm60_CPU_read8(c,rn));
arm60_CPU_write8(c,rn,cast<uint8_t>(arm60_CPU_reg(c,rm)));
arm60_CPU_setReg(c,rd,old);
}
else {
uint32_t old = arm60_CPU_read32(c,rn);
arm60_CPU_write32(c,rn,arm60_CPU_reg(c,rm));
arm60_CPU_setReg(c,rd,old);
}
}
}
// tools/cpu/arm60/exec.go:245:1
void arm60_CPU_execSingle(arm60_CPU* c,uint32_t w,bool regOff){
{
uint32_t l = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(20ULL))) & cast<uint32_t>(1ULL)));
uint32_t b = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(22ULL))) & cast<uint32_t>(1ULL)));
uint32_t p = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(24ULL))) & cast<uint32_t>(1ULL)));
uint32_t u = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(23ULL))) & cast<uint32_t>(1ULL)));
uint32_t wbit = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(21ULL))) & cast<uint32_t>(1ULL)));
uint32_t rn = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(15ULL)));
uint32_t rd = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(12ULL))) & cast<uint32_t>(15ULL)));
uint32_t base = arm60_CPU_reg(c,rn);
uint32_t off={};
if ((!regOff)) {
off = cast<uint32_t>((w & cast<uint32_t>(4095ULL)));
}
else {
auto tmp36 = arm60_CPU_shift(c,cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(5ULL))) & cast<uint32_t>(3ULL))),cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(7ULL))) & cast<uint32_t>(31ULL))),arm60_CPU_reg(c,cast<uint32_t>((w & cast<uint32_t>(15ULL)))),false,arm60_CPU_boolToU(c,c->C));
off = std::get<0>(tmp36);
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
std::function<void()> writeback = [&]()->void{
if ((p == cast<uint32_t>(0ULL))) {
if ((u == cast<uint32_t>(1ULL))) {
base += off;
}
else {
base -= off;
}
arm60_CPU_setReg(c,rn,base);
}
else if ((wbit == cast<uint32_t>(1ULL))) {
arm60_CPU_setReg(c,rn,addr);
}
}
;
if ((l == cast<uint32_t>(1ULL))) {
uint32_t v={};
if ((b == cast<uint32_t>(1ULL))) {
v = cast<uint32_t>(arm60_CPU_read8(c,addr));
}
else {
v = arm60_CPU_read32(c,addr);
}
writeback();
if ((rd == cast<uint32_t>(15ULL))) {
c->R[cast<int64_t>(15ULL)] = (v & ~(cast<uint32_t>(3ULL)));
c->branched = true;
}
else {
arm60_CPU_setReg(c,rd,v);
}
}
else {
uint32_t v = arm60_CPU_reg(c,rd);
if ((rd == cast<uint32_t>(15ULL))) {
v = cast<uint32_t>((c->cur + cast<uint32_t>(12ULL)));
}
if ((b == cast<uint32_t>(1ULL))) {
arm60_CPU_write8(c,addr,cast<uint8_t>(v));
}
else {
arm60_CPU_write32(c,addr,v);
}
writeback();
}
}
}
// tools/cpu/arm60/exec.go:309:1
void arm60_CPU_execBlock(arm60_CPU* c,uint32_t w){
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
uint32_t addr = start;
bool loadedRn = false;
for (uint32_t i = cast<uint32_t>(cast<uint32_t>(0ULL));(i < cast<uint32_t>(16ULL));i++){
if ((cast<uint32_t>((mask & (shl<uint32_t>(cast<uint32_t>(1ULL),i)))) == cast<uint32_t>(0ULL))) {
continue;
}
if ((l == cast<uint32_t>(1ULL))) {
uint32_t v = arm60_CPU_read32aligned(c,addr);
if ((i == cast<uint32_t>(15ULL))) {
if ((s == cast<uint32_t>(1ULL))) {
arm60_CPU_SetCPSR(c,arm60_CPU_SPSR(c));
}
c->R[cast<int64_t>(15ULL)] = (v & ~(cast<uint32_t>(3ULL)));
c->branched = true;
}
else {
c->R[i] = v;
if ((i == rn)) {
loadedRn = true;
}
}
}
else {
uint32_t v = c->R[i];
if ((i == cast<uint32_t>(15ULL))) {
v = cast<uint32_t>((c->cur + cast<uint32_t>(12ULL)));
}
arm60_CPU_write32(c,addr,v);
}
addr += cast<uint32_t>(4ULL);
}
if (((wbit == cast<uint32_t>(1ULL)) && (!(((l == cast<uint32_t>(1ULL)) && loadedRn))))) {
arm60_CPU_setReg(c,rn,final);
}
}
}
// tools/cpu/arm60/exec.go:371:1
void arm60_CPU_execMisc(arm60_CPU* c,uint32_t w,bool immForm){
{
uint32_t op = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(21ULL))) & cast<uint32_t>(15ULL)));
if (immForm) {
arm60_CPU_execMSR(c,w,true);
return ;
}
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(4ULL))) & cast<uint32_t>(15ULL))) == cast<uint32_t>(0ULL))) {
if (((op == cast<uint32_t>(8ULL)) || (op == cast<uint32_t>(10ULL)))) {
uint32_t v = arm60_CPU_CPSR(c);
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(22ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
v = arm60_CPU_SPSR(c);
}
arm60_CPU_setReg(c,cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(12ULL))) & cast<uint32_t>(15ULL))),v);
}
else {
arm60_CPU_execMSR(c,w,false);
}
return ;
}
arm60_CPU_Halt(c,std::string("undefined instruction 0x%08X at 0x%08X",38),w,c->cur);
}
}
// tools/cpu/arm60/exec.go:392:1
void arm60_CPU_execMSR(arm60_CPU* c,uint32_t w,bool immForm){
{
uint32_t val={};
if (immForm) {
val = arm60_ror32(cast<uint32_t>((w & cast<uint32_t>(255ULL))),cast<uint32_t>(((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(8ULL))) & cast<uint32_t>(15ULL)))) * cast<uint32_t>(2ULL))));
}
else {
val = arm60_CPU_reg(c,cast<uint32_t>((w & cast<uint32_t>(15ULL))));
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
arm60_CPU_SetSPSR(c,cast<uint32_t>(((arm60_CPU_SPSR(c) & ~(mask)) | cast<uint32_t>((val & mask)))));
return ;
}
if ((c->Mode == arm60_ModeUSR)) {
mask &= cast<uint32_t>(4278190080ULL);
}
arm60_CPU_SetCPSR(c,cast<uint32_t>(((arm60_CPU_CPSR(c) & ~(mask)) | cast<uint32_t>((val & mask)))));
}
}
// tools/cpu/arm60/exec.go:423:1
void arm60_CPU_execSWI(arm60_CPU* c,uint32_t w){
{
if ((bool(c->SWI) && c->SWI(c,cast<uint32_t>((w & cast<uint32_t>(16777215ULL)))))) {
c->R[cast<int64_t>(15ULL)] = cast<uint32_t>((c->cur + cast<uint32_t>(4ULL)));
c->branched = true;
return ;
}
arm60_CPU_exception(c,arm60_ModeSVC,cast<uint32_t>(8ULL),cast<uint32_t>((c->cur + cast<uint32_t>(4ULL))));
}
}
// tools/cpu/arm60/exec.go:434:1
void arm60_CPU_exception(arm60_CPU* c,uint32_t mode,uint32_t vector,uint32_t returnAddr){
{
uint32_t saved = arm60_CPU_CPSR(c);
arm60_CPU_switchMode(c,mode);
arm60_CPU_SetSPSR(c,saved);
c->R[cast<int64_t>(14ULL)] = returnAddr;
c->IRQDisable = true;
c->R[cast<int64_t>(15ULL)] = vector;
c->branched = true;
}
}
// tools/cpu/arm60/exec.go:446:1
void arm60_CPU_Exception(arm60_CPU* c,uint32_t mode,uint32_t vector,uint32_t returnAddr){
{
arm60_CPU_exception(c,mode,vector,returnAddr);
}
}
// tools/cpu/arm60/exec.go:451:1
bool arm60_CPU_IRQ(arm60_CPU* c){
{
if (c->IRQDisable) {
return false;
}
arm60_CPU_exception(c,arm60_ModeIRQ,cast<uint32_t>(24ULL),cast<uint32_t>((c->R[cast<int64_t>(15ULL)] + cast<uint32_t>(4ULL))));
return true;
}
}
// tools/cpu/arm60/exec.go:462:1
void arm60_CPU_execCopro(arm60_CPU* c,uint32_t w){
{
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(4ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(0ULL))) {
return ;
}
bool load = (cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(20ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL));
uint32_t rd = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(12ULL))) & cast<uint32_t>(15ULL)));
if ((load && (rd != cast<uint32_t>(15ULL)))) {
arm60_CPU_setReg(c,rd,cast<uint32_t>(0ULL));
}
}
}
// tools/platform/threedo/aif.go:45:1
uint32_t threedo_be32at(Slice<uint8_t> b,int64_t o){
{
return cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((shl<uint32_t>(cast<uint32_t>(b[o]),cast<int64_t>(24ULL)) | shl<uint32_t>(cast<uint32_t>(b[cast<int64_t>((o + cast<int64_t>(1ULL)))]),cast<int64_t>(16ULL)))) | shl<uint32_t>(cast<uint32_t>(b[cast<int64_t>((o + cast<int64_t>(2ULL)))]),cast<int64_t>(8ULL)))) | cast<uint32_t>(b[cast<int64_t>((o + cast<int64_t>(3ULL)))])));
}
}
// tools/platform/threedo/aif.go:51:1
std::tuple<threedo_AIF*,Error> threedo_ParseAIF(Slice<uint8_t> data){
{
if ((len(data) < cast<int64_t>(128ULL))) {
return {{},go_fmt_Errorf(std::string("threedo: AIF too small (%d bytes)",33),len(data))};
}
threedo_AIF* a = arenaNew(threedo_AIF{threedo_be32at(data,cast<int64_t>(0ULL)),{},{},threedo_be32at(data,cast<int64_t>(16ULL)),threedo_be32at(data,cast<int64_t>(20ULL)),threedo_be32at(data,cast<int64_t>(24ULL)),threedo_be32at(data,cast<int64_t>(32ULL)),threedo_be32at(data,cast<int64_t>(40ULL)),threedo_be32at(data,cast<int64_t>(48ULL)),data});
std::function<bool(uint32_t)> isBranchOrNOP = [&](uint32_t w)->bool{
return ((w == threedo_aifNOP) || (shr<uint32_t>(w,cast<int64_t>(24ULL)) == cast<uint32_t>(235ULL)));
}
;
if (((!isBranchOrNOP(a->Decompress)) || (!isBranchOrNOP(threedo_be32at(data,cast<int64_t>(4ULL)))))) {
return {{},go_fmt_Errorf(std::string("threedo: not an AIF (leading words 0x%08X 0x%08X)",49),a->Decompress,threedo_be32at(data,cast<int64_t>(4ULL)))};
}
if (((a->AddrMode != cast<uint32_t>(26ULL)) && (a->AddrMode != cast<uint32_t>(32ULL)))) {
return {{},go_fmt_Errorf(std::string("threedo: AIF address mode %d (want 26 or 32)",44),a->AddrMode)};
}
a->SelfReloc = (threedo_be32at(data,cast<int64_t>(4ULL)) != threedo_aifNOP);
uint32_t imm = cast<uint32_t>((threedo_be32at(data,cast<int64_t>(12ULL)) & cast<uint32_t>(16777215ULL)));
uint32_t off = shl<uint32_t>(imm,cast<int64_t>(2ULL));
if ((cast<uint32_t>((imm & cast<uint32_t>(8388608ULL))) != cast<uint32_t>(0ULL))) {
off |= cast<uint32_t>(4227858432ULL);
}
a->EntryTarget = cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((a->ImageBase + cast<uint32_t>(12ULL))) + cast<uint32_t>(8ULL))) + off));
return {a,{}};
}
}
// tools/platform/threedo/aif.go:85:1
std::string threedo_AIF_Describe(threedo_AIF* a){
{
std::string reloc = std::string("no",2);
if (a->SelfReloc) {
reloc = std::string("yes",3);
}
return go_fmt_Sprintf((((((std::string("AIF executable\012",15) + std::string("  image base = 0x%08X   address mode = %d-bit\012",46)) + std::string("  RO size    = 0x%X (%d, incl. 128-byte header)\012",48)) + std::string("  RW size    = 0x%X (%d)\012",25)) + std::string("  zero-init  = 0x%X (%d)\012",25)) + std::string("  self-reloc = %s   entry (BL) = 0x%08X\012",40)),a->ImageBase,a->AddrMode,a->ROSize,a->ROSize,a->RWSize,a->RWSize,a->ZeroSize,a->ZeroSize,reloc,a->EntryTarget);
}
}
// tools/platform/threedo/audiofolio.go:63:1
bool threedo_Machine_audioFolioSWI(threedo_Machine* m,arm60_CPU* c,uint32_t swi){
{
{
switch(swi){
case threedo_swiSignalAtTime:{
threedo_item* cue = get(m->items,cast<int32_t>(arm60_CPU_Reg(c,cast<uint32_t>(0ULL))));
if (((!cue) || (cue->typ != threedo_typeAudioCue))) {
arm60_CPU_SetReg(c,cast<uint32_t>(0ULL),cast<uint32_t>(~cast<uint32_t>(cast<uint32_t>(0ULL))));
return true;
}
uint32_t t = arm60_CPU_Reg(c,cast<uint32_t>(1ULL));
if ((cast<int32_t>(cast<uint32_t>((t - m->audioTime))) <= cast<int32_t>(0ULL))) {
threedo_Machine_sendSignal(m,cue->owner,cue->signal);
}
else {
m->audioEvents = append(m->audioEvents,threedo_audioEvent{cue->num,t});
}
arm60_CPU_SetReg(c,cast<uint32_t>(0ULL),cast<uint32_t>(0ULL));
break;}
case threedo_swiAbortTimerCue:{
int32_t cueNum = cast<int32_t>(arm60_CPU_Reg(c,cast<uint32_t>(0ULL)));
Slice<threedo_audioEvent> kept = sub(m->audioEvents,0,cast<int64_t>(0ULL));
{auto&& tmp1 = m->audioEvents;
for(int64_t tmp2=0;tmp2<len(tmp1);++tmp2){
auto e=tmp1[tmp2];if ((e.cue != cueNum)) {
kept = append(kept,e);
}
}}
m->audioEvents = kept;
arm60_CPU_SetReg(c,cast<uint32_t>(0ULL),cast<uint32_t>(0ULL));
break;}
case threedo_swiSetAudioRate:case threedo_swiSetAudioDuration:{
threedo_Machine_note(m,go_fmt_Sprintf(std::string("audio SetAudioRate/Duration(0x%X, 0x%X) ignored (clock stays 240 Hz)",68),arm60_CPU_Reg(c,cast<uint32_t>(0ULL)),arm60_CPU_Reg(c,cast<uint32_t>(1ULL))));
arm60_CPU_SetReg(c,cast<uint32_t>(0ULL),cast<uint32_t>(0ULL));
break;}
case threedo_swiMonitorAttach:{
{
int32_t att = cast<int32_t>(arm60_CPU_Reg(c,cast<uint32_t>(0ULL)));
if ((att != cast<int32_t>(0ULL))) {
m->attachCue[att] = cast<int32_t>(arm60_CPU_Reg(c,cast<uint32_t>(1ULL)));
}
}
arm60_CPU_SetReg(c,cast<uint32_t>(0ULL),cast<uint32_t>(0ULL));
break;}
case threedo_swiStartAttachment:{
threedo_Machine_completeAttachment(m,cast<int32_t>(arm60_CPU_Reg(c,cast<uint32_t>(0ULL))));
arm60_CPU_SetReg(c,cast<uint32_t>(0ULL),cast<uint32_t>(0ULL));
break;}
case threedo_swiLinkAttachments:{
threedo_Machine_completeAttachment(m,cast<int32_t>(arm60_CPU_Reg(c,cast<uint32_t>(1ULL))));
arm60_CPU_SetReg(c,cast<uint32_t>(0ULL),cast<uint32_t>(0ULL));
break;}
case threedo_swiStopAttachment:{
{
auto tmp3 = lookup(m->attachCue,cast<int32_t>(arm60_CPU_Reg(c,cast<uint32_t>(0ULL))));
int32_t cueNum = std::get<0>(tmp3);
bool ok = std::get<1>(tmp3);
if (ok) {
Slice<threedo_audioEvent> kept = sub(m->audioEvents,0,cast<int64_t>(0ULL));
{auto&& tmp4 = m->audioEvents;
for(int64_t tmp5=0;tmp5<len(tmp4);++tmp5){
auto e=tmp4[tmp5];if ((e.cue != cueNum)) {
kept = append(kept,e);
}
}}
m->audioEvents = kept;
}
}
arm60_CPU_SetReg(c,cast<uint32_t>(0ULL),cast<uint32_t>(0ULL));
break;}
case threedo_swiStartInstrument:case threedo_swiStopInstrument:case threedo_swiTestHack:case threedo_swiConnectInstr:case threedo_swiTweakRawKnob:case threedo_swiSetAudioItemInfo:{
arm60_CPU_SetReg(c,cast<uint32_t>(0ULL),cast<uint32_t>(0ULL));
break;}
default:{
return false;
break;}
}}
return true;
}
}
// tools/platform/threedo/audiofolio.go:149:1
void threedo_Machine_completeAttachment(threedo_Machine* m,int32_t att){
{
auto tmp6 = lookup(m->attachCue,att);
int32_t cueNum = std::get<0>(tmp6);
bool ok = std::get<1>(tmp6);
if ((!ok)) {
return ;
}
threedo_item* cue = get(m->items,cueNum);
if (((!cue) || (cue->typ != threedo_typeAudioCue))) {
return ;
}
m->audioEvents = append(m->audioEvents,threedo_audioEvent{cue->num,cast<uint32_t>((m->audioTime + cast<uint32_t>(1ULL)))});
}
}
// tools/platform/threedo/audiofolio.go:163:1
void threedo_Machine_serviceAudioFolio(threedo_Machine* m,uint32_t foff){
{
arm60_CPU* c = m->CPU;
{
switch(foff){
case cast<uint32_t>(4ULL):{
threedo_Machine_SetResultAndReturn(m,cast<uint32_t>(threedo_Machine_createItem(m,threedo_typeInsTemplate,cast<uint32_t>(0ULL),cast<uint32_t>(0ULL))->num));
break;}
case cast<uint32_t>(8ULL):{
threedo_Machine_SetResultAndReturn(m,cast<uint32_t>(threedo_Machine_createItem(m,threedo_typeInstrument,cast<uint32_t>(0ULL),cast<uint32_t>(0ULL))->num));
break;}
case cast<uint32_t>(16ULL):{
threedo_Machine_SetResultAndReturn(m,cast<uint32_t>(threedo_Machine_createItem(m,threedo_typeKnob,cast<uint32_t>(0ULL),cast<uint32_t>(0ULL))->num));
break;}
case cast<uint32_t>(56ULL):{
threedo_Machine_SetResultAndReturn(m,cast<uint32_t>(threedo_Machine_createItem(m,threedo_typeSample,cast<uint32_t>(0ULL),cast<uint32_t>(0ULL))->num));
break;}
case cast<uint32_t>(144ULL):{
threedo_Machine_SetResultAndReturn(m,cast<uint32_t>(threedo_Machine_createItem(m,threedo_typeAttachment,cast<uint32_t>(0ULL),cast<uint32_t>(0ULL))->num));
break;}
case cast<uint32_t>(20ULL):{
threedo_Machine_SetResultAndReturn(m,cast<uint32_t>(0ULL));
break;}
case cast<uint32_t>(168ULL):{
threedo_Machine_SetResultAndReturn(m,m->audioTime);
break;}
case cast<uint32_t>(72ULL):{
{
threedo_item* it = get(m->items,cast<int32_t>(arm60_CPU_Reg(c,cast<uint32_t>(0ULL))));
if (bool(it)) {
threedo_Machine_SetResultAndReturn(m,it->signal);
}
else {
threedo_Machine_SetResultAndReturn(m,cast<uint32_t>(0ULL));
}
}
break;}
case cast<uint32_t>(76ULL):{
if ((m->audioClockOwner == cast<int32_t>(0ULL))) {
m->audioClockOwner = threedo_Machine_createItem(m,cast<uint32_t>(1024ULL),cast<uint32_t>(0ULL),cast<uint32_t>(0ULL))->num;
}
threedo_Machine_SetResultAndReturn(m,cast<uint32_t>(m->audioClockOwner));
break;}
case cast<uint32_t>(80ULL):{
threedo_Machine_SetResultAndReturn(m,cast<uint32_t>(0ULL));
break;}
case cast<uint32_t>(60ULL):{
threedo_Machine_SetResultAndReturn(m,shl<uint32_t>(cast<int64_t>(240ULL),cast<int64_t>(16ULL)));
break;}
case cast<uint32_t>(64ULL):{
threedo_Machine_SetResultAndReturn(m,divi<uint32_t>(threedo_audioSampleRate,cast<int64_t>(240ULL)));
break;}
case cast<uint32_t>(68ULL):{
threedo_item* cue = get(m->items,cast<int32_t>(arm60_CPU_Reg(c,cast<uint32_t>(0ULL))));
uint32_t t = arm60_CPU_Reg(c,cast<uint32_t>(1ULL));
threedo_Machine_SetResultAndReturn(m,cast<uint32_t>(0ULL));
if ((((!cue) || (cue->typ != threedo_typeAudioCue)) || (cast<int32_t>(cast<uint32_t>((t - m->audioTime))) <= cast<int32_t>(0ULL)))) {
return ;
}
m->audioEvents = append(m->audioEvents,threedo_audioEvent{cue->num,t});
threedo_task* task = threedo_Machine_curTask(m);
task->wait = cue->signal;
task->state = threedo_stWaiting;
m->needSchedule = true;
break;}
default:{
threedo_Machine_note(m,go_fmt_Sprintf(std::string("AudioFolio[-0x%X] stub (r0=0x%08X r1=0x%08X r2=0x%08X)",54),foff,arm60_CPU_Reg(c,cast<uint32_t>(0ULL)),arm60_CPU_Reg(c,cast<uint32_t>(1ULL)),arm60_CPU_Reg(c,cast<uint32_t>(2ULL))));
threedo_Machine_SetResultAndReturn(m,cast<uint32_t>(0ULL));
break;}
}}
}
}
// tools/platform/threedo/audiofolio.go:224:1
void threedo_Machine_advanceAudioClock(threedo_Machine* m,uint32_t fields){
{
m->audioTime += cast<uint32_t>((fields * threedo_audioTicksPerField));
if ((len(m->audioEvents) == cast<int64_t>(0ULL))) {
return ;
}
Slice<threedo_audioEvent> kept = sub(m->audioEvents,0,cast<int64_t>(0ULL));
{auto&& tmp7 = m->audioEvents;
for(int64_t tmp8=0;tmp8<len(tmp7);++tmp8){
auto e=tmp7[tmp8];if ((cast<int32_t>(cast<uint32_t>((m->audioTime - e.time))) >= cast<int32_t>(0ULL))) {
{
threedo_item* cue = get(m->items,e.cue);
if (bool(cue)) {
threedo_Machine_sendSignal(m,cue->owner,cue->signal);
}
}
}
else {
kept = append(kept,e);
}
}}
m->audioEvents = kept;
}
}
// tools/platform/threedo/cel.go:46:1
std::tuple<threedo_Cel*,Error> threedo_ParseCel(Slice<uint8_t> data){
{
threedo_Cel* c = arenaNew(threedo_Cel{});
bool haveCCB = false;
for (int64_t p = cast<int64_t>(0ULL);(cast<int64_t>((p + cast<int64_t>(8ULL))) <= len(data));){
std::string tag = cast<std::string>(sub(data,p,cast<int64_t>((p + cast<int64_t>(4ULL)))));
int64_t size = cast<int64_t>(threedo_be32(sub(data,cast<int64_t>((p + cast<int64_t>(4ULL))),len(data))));
if (((size < cast<int64_t>(8ULL)) || (cast<int64_t>((p + size)) > len(data)))) {
break;
}
Slice<uint8_t> body = sub(data,cast<int64_t>((p + cast<int64_t>(8ULL))),cast<int64_t>((p + size)));
{
auto tmp10=tag;
if (tmp10==(std::string("CCB ",4))){
if (haveCCB) {
return threedo_Cel_finish(c);
}
if ((len(body) < cast<int64_t>(72ULL))) {
return {{},go_fmt_Errorf(std::string("threedo: CCB chunk too small (%d)",33),len(body))};
}
c->Flags = threedo_be32(sub(body,cast<int64_t>(4ULL),len(body)));
c->PIXC = threedo_be32(sub(body,cast<int64_t>(52ULL),len(body)));
c->PRE0 = threedo_be32(sub(body,cast<int64_t>(56ULL),len(body)));
c->PRE1 = threedo_be32(sub(body,cast<int64_t>(60ULL),len(body)));
c->Width = cast<int64_t>(threedo_be32(sub(body,cast<int64_t>(64ULL),len(body))));
c->Height = cast<int64_t>(threedo_be32(sub(body,cast<int64_t>(68ULL),len(body))));
c->Packed = (cast<uint32_t>((c->Flags & threedo_ccbPacked)) != cast<uint32_t>(0ULL));
haveCCB = true;
}
else if (tmp10==(std::string("PLUT",4))){
if ((len(body) < cast<int64_t>(4ULL))) {
goto tmp9;
}
int64_t n = cast<int64_t>(threedo_be32(sub(body,cast<int64_t>(0ULL),len(body))));
for (int64_t i = cast<int64_t>(0ULL);((i < n) && (cast<int64_t>((cast<int64_t>(4ULL) + cast<int64_t>(((cast<int64_t>((i + cast<int64_t>(1ULL)))) * cast<int64_t>(2ULL))))) <= len(body)));i++){
c->PLUT = append(c->PLUT,cast<uint16_t>((shl<uint16_t>(cast<uint16_t>(body[cast<int64_t>((cast<int64_t>(4ULL) + cast<int64_t>((i * cast<int64_t>(2ULL)))))]),cast<int64_t>(8ULL)) | cast<uint16_t>(body[cast<int64_t>((cast<int64_t>((cast<int64_t>(4ULL) + cast<int64_t>((i * cast<int64_t>(2ULL))))) + cast<int64_t>(1ULL)))]))));
}
}
else if (tmp10==(std::string("PDAT",4))){
c->PDAT = body;
}
}
tmp9:;
p += size;
}
if ((!haveCCB)) {
return {{},go_fmt_Errorf(std::string("threedo: no CCB chunk (not a cel?)",34))};
}
return threedo_Cel_finish(c);
}
}
// tools/platform/threedo/cel.go:97:1
std::tuple<threedo_Cel*,Error> threedo_Cel_finish(threedo_Cel* c){
{
if (((cast<uint32_t>((c->Flags & threedo_ccbCCBPre)) == cast<uint32_t>(0ULL)) && (len(c->PDAT) >= cast<int64_t>(4ULL)))) {
c->PRE0 = threedo_be32(c->PDAT);
c->PDAT = sub(c->PDAT,cast<int64_t>(4ULL),len(c->PDAT));
if (((cast<uint32_t>((c->Flags & threedo_ccbPacked)) == cast<uint32_t>(0ULL)) && (len(c->PDAT) >= cast<int64_t>(4ULL)))) {
c->PRE1 = threedo_be32(c->PDAT);
c->PDAT = sub(c->PDAT,cast<int64_t>(4ULL),len(c->PDAT));
}
}
c->BPP = get(threedo_bppFromPRE0,cast<uint32_t>((c->PRE0 & cast<uint32_t>(7ULL))));
if ((c->BPP == cast<int64_t>(0ULL))) {
return {{},go_fmt_Errorf(std::string("threedo: unknown bpp code %d",28),cast<uint32_t>((c->PRE0 & cast<uint32_t>(7ULL))))};
}
{
int64_t w = cast<int64_t>((cast<int64_t>(cast<uint32_t>((c->PRE1 & cast<uint32_t>(2047ULL)))) + cast<int64_t>(1ULL)));
if (((w > cast<int64_t>(1ULL)) || (c->Width == cast<int64_t>(0ULL)))) {
c->Width = w;
}
}
{
int64_t h = cast<int64_t>((cast<int64_t>(cast<uint32_t>(((shr<uint32_t>(c->PRE0,cast<int64_t>(6ULL))) & cast<uint32_t>(1023ULL)))) + cast<int64_t>(1ULL)));
if (((h > cast<int64_t>(1ULL)) || (c->Height == cast<int64_t>(0ULL)))) {
c->Height = h;
}
}
c->Coded = (c->BPP <= cast<int64_t>(8ULL));
return {c,{}};
}
}
// tools/platform/threedo/cel.go:135:1
color_RGBA threedo_RGB555(uint16_t v){
{
return threedo_rgb555(v);
}
}
// tools/platform/threedo/cel.go:139:1
color_RGBA threedo_rgb555(uint16_t v){
{
uint8_t r = cast<uint8_t>(cast<uint16_t>(((shr<uint16_t>(v,cast<int64_t>(10ULL))) & cast<uint16_t>(31ULL))));
uint8_t g = cast<uint8_t>(cast<uint16_t>(((shr<uint16_t>(v,cast<int64_t>(5ULL))) & cast<uint16_t>(31ULL))));
uint8_t b = cast<uint8_t>(cast<uint16_t>((v & cast<uint16_t>(31ULL))));
return color_RGBA{cast<uint8_t>((shl<uint8_t>(r,cast<int64_t>(3ULL)) | shr<uint8_t>(r,cast<int64_t>(2ULL)))),cast<uint8_t>((shl<uint8_t>(g,cast<int64_t>(3ULL)) | shr<uint8_t>(g,cast<int64_t>(2ULL)))),cast<uint8_t>((shl<uint8_t>(b,cast<int64_t>(3ULL)) | shr<uint8_t>(b,cast<int64_t>(2ULL)))),cast<uint8_t>(255ULL)};
}
}
// tools/platform/threedo/cel.go:152:1
uint32_t threedo_bitReader_read(threedo_bitReader* br,int64_t n){
{
uint32_t v={};
for (int64_t i = cast<int64_t>(0ULL);(i < n);i++){
int64_t bytIdx = shr<int64_t>(br->pos,cast<int64_t>(3ULL));
if ((bytIdx >= len(br->data))) {
br->pos++;
v = shl<uint32_t>(v,cast<int64_t>(1ULL));
continue;
}
int64_t bit = cast<int64_t>((cast<int64_t>(7ULL) - (cast<int64_t>((br->pos & cast<int64_t>(7ULL))))));
v = cast<uint32_t>((shl<uint32_t>(v,cast<int64_t>(1ULL)) | cast<uint32_t>(cast<uint8_t>(((shr<uint8_t>(br->data[bytIdx],cast<uint64_t>(bit))) & cast<uint8_t>(1ULL))))));
br->pos++;
}
return v;
}
}
// tools/platform/threedo/cel.go:190:1
std::tuple<image_RGBA*,Error> threedo_Cel_Image(threedo_Cel* c){
{
if (((((c->Width <= cast<int64_t>(0ULL)) || (c->Height <= cast<int64_t>(0ULL))) || (c->Width > cast<int64_t>(4096ULL))) || (c->Height > cast<int64_t>(4096ULL)))) {
return {{},go_fmt_Errorf(std::string("threedo: implausible cel size %dx%d",35),c->Width,c->Height)};
}
image_RGBA* img = go_image_NewRGBA(go_image_Rect(cast<int64_t>(0ULL),cast<int64_t>(0ULL),c->Width,c->Height));
bool bgnd = (cast<uint32_t>((c->Flags & threedo_ccbBGND)) != cast<uint32_t>(0ULL));
std::function<void(int64_t,int64_t,uint32_t)> set = [&](int64_t x,int64_t y,uint32_t v)->void{
if (((x < cast<int64_t>(0ULL)) || (x >= c->Width))) {
return ;
}
uint32_t amv = cast<uint32_t>(cast<uint32_t>(73ULL));
uint16_t raw={};
if (c->Coded) {
uint32_t idx={};
uint16_t pw={};
{
switch(c->BPP){
case cast<int64_t>(1ULL):{
idx = cast<uint32_t>((cast<uint32_t>(((cast<uint32_t>((c->Flags & cast<uint32_t>(15ULL)))) * cast<uint32_t>(2ULL))) + (cast<uint32_t>((v & cast<uint32_t>(1ULL))))));
break;}
case cast<int64_t>(2ULL):{
idx = cast<uint32_t>((cast<uint32_t>(((cast<uint32_t>((c->Flags & cast<uint32_t>(14ULL)))) * cast<uint32_t>(2ULL))) + (cast<uint32_t>((v & cast<uint32_t>(3ULL))))));
break;}
case cast<int64_t>(4ULL):{
idx = cast<uint32_t>((cast<uint32_t>(((cast<uint32_t>((c->Flags & cast<uint32_t>(8ULL)))) * cast<uint32_t>(2ULL))) + (cast<uint32_t>((v & cast<uint32_t>(15ULL))))));
break;}
case cast<int64_t>(6ULL):{
idx = cast<uint32_t>((v & cast<uint32_t>(31ULL)));
pw = shl<uint16_t>(cast<uint16_t>(cast<uint32_t>(((shr<uint32_t>(v,cast<int64_t>(5ULL))) & cast<uint32_t>(1ULL)))),cast<int64_t>(15ULL));
break;}
case cast<int64_t>(8ULL):{
idx = cast<uint32_t>((v & cast<uint32_t>(31ULL)));
amv = cast<uint32_t>(((cast<uint32_t>((shl<uint32_t>((cast<uint32_t>(((shr<uint32_t>(v,cast<int64_t>(6ULL))) & cast<uint32_t>(3ULL)))),cast<int64_t>(1ULL)) | cast<uint32_t>(((shr<uint32_t>(v,cast<int64_t>(5ULL))) & cast<uint32_t>(1ULL)))))) * cast<uint32_t>(73ULL)));
break;}
default:{
idx = cast<uint32_t>((v & cast<uint32_t>(31ULL)));
break;}
}}
if ((len(c->PLUT) == cast<int64_t>(0ULL))) {
uint8_t g = cast<uint8_t>(divi<uint32_t>(cast<uint32_t>((v * cast<uint32_t>(255ULL))),cast<uint32_t>(cast<uint32_t>(((shl<uint32_t>(cast<uint32_t>(1ULL),c->BPP)) - cast<uint32_t>(1ULL))))));
image_RGBA_SetRGBA(img,x,y,color_RGBA{g,g,g,cast<uint8_t>(255ULL)});
return ;
}
if ((cast<int64_t>(idx) >= len(c->PLUT))) {
return ;
}
raw = c->PLUT[idx];
if ((c->BPP == cast<int64_t>(6ULL))) {
raw = cast<uint16_t>((cast<uint16_t>((raw & cast<uint16_t>(32767ULL))) | pw));
}
}
else {
raw = cast<uint16_t>(v);
}
if (((cast<uint16_t>((raw & cast<uint16_t>(32767ULL))) == cast<uint16_t>(0ULL)) && (!bgnd))) {
return ;
}
auto tmp11 = threedo_Cel_ppmp(c,raw,amv);
uint16_t out = std::get<0>(tmp11);
uint8_t alpha = std::get<1>(tmp11);
color_RGBA col = threedo_rgb555(out);
col.A = alpha;
image_RGBA_SetRGBA(img,x,y,col);
}
;
if (c->Packed) {
threedo_Cel_decodePacked(c,set);
}
else {
threedo_Cel_decodeUnpacked(c,set);
}
return {img,{}};
}
}
// tools/platform/threedo/cel.go:276:1
std::tuple<uint16_t,uint8_t> threedo_Cel_ppmp(threedo_Cel* c,uint16_t pix,uint32_t amv){
{
uint32_t word = cast<uint32_t>((c->PIXC & cast<uint32_t>(65535ULL)));
{
switch(cast<uint32_t>((c->Flags & threedo_ccbPOVER))){
case cast<uint32_t>(256ULL):{
break;}
case cast<uint32_t>(384ULL):{
word = shr<uint32_t>(c->PIXC,cast<int64_t>(16ULL));
break;}
default:{
if ((cast<uint16_t>((pix & cast<uint16_t>(32768ULL))) != cast<uint16_t>(0ULL))) {
word = shr<uint32_t>(c->PIXC,cast<int64_t>(16ULL));
}
break;}
}}
if ((word == cast<uint32_t>(0ULL))) {
return {pix,cast<uint8_t>(255ULL)};
}
bool s1 = (cast<uint32_t>((word & cast<uint32_t>(32768ULL))) != cast<uint32_t>(0ULL));
uint32_t ms = cast<uint32_t>(((shr<uint32_t>(word,cast<int64_t>(13ULL))) & cast<uint32_t>(3ULL)));
uint32_t mxf = cast<uint32_t>(((shr<uint32_t>(word,cast<int64_t>(10ULL))) & cast<uint32_t>(7ULL)));
uint32_t dv1 = cast<uint32_t>(((shr<uint32_t>(word,cast<int64_t>(8ULL))) & cast<uint32_t>(3ULL)));
uint32_t s2 = cast<uint32_t>(((shr<uint32_t>(word,cast<int64_t>(6ULL))) & cast<uint32_t>(3ULL)));
uint32_t avf = cast<uint32_t>(((shr<uint32_t>(word,cast<int64_t>(1ULL))) & cast<uint32_t>(31ULL)));
uint32_t dv2 = cast<uint32_t>((word & cast<uint32_t>(1ULL)));
uint32_t dv3={};
bool clip = true;
if ((cast<uint32_t>((c->Flags & cast<uint32_t>(1024ULL))) != cast<uint32_t>(0ULL))) {
dv3 = cast<uint32_t>(((shr<uint32_t>(avf,cast<int64_t>(3ULL))) & cast<uint32_t>(3ULL)));
clip = (cast<uint32_t>((avf & cast<uint32_t>(4ULL))) == cast<uint32_t>(0ULL));
avf &= ~(cast<uint32_t>(28ULL));
}
uint32_t sh1 = cast<uint32_t>((cast<uint32_t>(((cast<uint32_t>((dv1 - cast<uint32_t>(1ULL)))) & cast<uint32_t>(3ULL))) + cast<uint32_t>(1ULL)));
if (s1) {
double frac={};
{
switch(ms){
case cast<uint32_t>(1ULL):{
frac = (cast<double>(cast<uint32_t>(((cast<uint32_t>((amv & cast<uint32_t>(7ULL)))) + cast<uint32_t>(1ULL)))) / cast<double>(shl<uint32_t>(cast<uint32_t>(cast<uint32_t>(1ULL)),sh1)));
break;}
case cast<uint32_t>(2ULL):case cast<uint32_t>(3ULL):{
frac = (cast<double>((shr<int64_t>(cast<int64_t>(16ULL),cast<int64_t>(2ULL)) + cast<int64_t>(1ULL))) / cast<double>(shl<uint32_t>(cast<uint32_t>(cast<uint32_t>(1ULL)),sh1)));
break;}
default:{
frac = (cast<double>(cast<uint32_t>((mxf + cast<uint32_t>(1ULL)))) / cast<double>(shl<uint32_t>(cast<uint32_t>(cast<uint32_t>(1ULL)),sh1)));
break;}
}}
uint32_t gray={};
{
switch(s2){
case cast<uint32_t>(1ULL):{
gray = shr<uint32_t>(avf,dv3);
break;}
case cast<uint32_t>(2ULL):{
frac += (cast<double>(1ULL) / cast<double>(shl<uint32_t>(cast<uint32_t>(cast<uint32_t>(1ULL)),dv3)));
break;}
}}
frac /= cast<double>(shl<uint32_t>(cast<uint32_t>(cast<uint32_t>(1ULL)),dv2));
gray = shr<uint32_t>(gray,dv2);
double alpha = (cast<double>(1ULL) - frac);
if ((alpha < cast<double>(0ULL))) {
alpha = cast<double>(0ULL);
}
if ((alpha > cast<double>(1ULL))) {
alpha = cast<double>(1ULL);
}
uint16_t g = cast<uint16_t>(gray);
if ((g > cast<uint16_t>(31ULL))) {
g = cast<uint16_t>(31ULL);
}
return {cast<uint16_t>((cast<uint16_t>((shl<uint16_t>(g,cast<int64_t>(10ULL)) | shl<uint16_t>(g,cast<int64_t>(5ULL)))) | g)),cast<uint8_t>(((alpha * cast<double>(255ULL)) + 0.5))};
}
if ((s2 >= cast<uint32_t>(2ULL))) {
return {pix,cast<uint8_t>(255ULL)};
}
uint32_t second={};
if ((s2 == cast<uint32_t>(1ULL))) {
second = shr<uint32_t>(avf,dv3);
}
std::function<uint16_t(uint32_t)> scale = [&](uint32_t ch)->uint16_t{
uint32_t first={};
{
switch(ms){
case cast<uint32_t>(1ULL):{
first = shr<uint32_t>((cast<uint32_t>((ch * (cast<uint32_t>(((cast<uint32_t>((amv & cast<uint32_t>(7ULL)))) + cast<uint32_t>(1ULL))))))),sh1);
break;}
case cast<uint32_t>(2ULL):{
first = shr<uint32_t>((cast<uint32_t>((ch * (cast<uint32_t>(((shr<uint32_t>(ch,cast<int64_t>(2ULL))) + cast<uint32_t>(1ULL))))))),(cast<uint32_t>((cast<uint32_t>(((cast<uint32_t>((cast<uint32_t>((ch & cast<uint32_t>(3ULL))) - cast<uint32_t>(1ULL)))) & cast<uint32_t>(3ULL))) + cast<uint32_t>(1ULL)))));
break;}
case cast<uint32_t>(3ULL):{
first = shr<uint32_t>((cast<uint32_t>((ch * (cast<uint32_t>(((shr<uint32_t>(ch,cast<int64_t>(2ULL))) + cast<uint32_t>(1ULL))))))),sh1);
break;}
default:{
first = shr<uint32_t>((cast<uint32_t>((ch * (cast<uint32_t>((mxf + cast<uint32_t>(1ULL))))))),sh1);
break;}
}}
uint32_t v = shr<uint32_t>((cast<uint32_t>((first + second))),dv2);
if ((!clip)) {
return cast<uint16_t>(cast<uint32_t>((v & cast<uint32_t>(31ULL))));
}
if ((v > cast<uint32_t>(31ULL))) {
v = cast<uint32_t>(31ULL);
}
return cast<uint16_t>(v);
}
;
uint16_t r = scale(cast<uint32_t>(((shr<uint32_t>(cast<uint32_t>(pix),cast<int64_t>(10ULL))) & cast<uint32_t>(31ULL))));
uint16_t g = scale(cast<uint32_t>(((shr<uint32_t>(cast<uint32_t>(pix),cast<int64_t>(5ULL))) & cast<uint32_t>(31ULL))));
uint16_t b = scale(cast<uint32_t>((cast<uint32_t>(pix) & cast<uint32_t>(31ULL))));
return {cast<uint16_t>((cast<uint16_t>((shl<uint16_t>(r,cast<int64_t>(10ULL)) | shl<uint16_t>(g,cast<int64_t>(5ULL)))) | b)),cast<uint8_t>(255ULL)};
}
}
// tools/platform/threedo/cel.go:379:1
void threedo_Cel_decodePacked(threedo_Cel* c,std::function<void(int64_t,int64_t,uint32_t)> set){
{
int64_t offBytes = cast<int64_t>(1ULL);
if ((c->BPP >= cast<int64_t>(8ULL))) {
offBytes = cast<int64_t>(2ULL);
}
int64_t pos = cast<int64_t>(0ULL);
for (int64_t y = cast<int64_t>(0ULL);(y < c->Height);y++){
if ((cast<int64_t>((pos + offBytes)) > len(c->PDAT))) {
break;
}
int64_t lineStart = pos;
int64_t words={};
if ((offBytes == cast<int64_t>(1ULL))) {
words = cast<int64_t>((cast<int64_t>(c->PDAT[lineStart]) + cast<int64_t>(2ULL)));
}
else {
words = cast<int64_t>(((cast<int64_t>((shl<int64_t>(cast<int64_t>(c->PDAT[lineStart]),cast<int64_t>(8ULL)) | cast<int64_t>(c->PDAT[cast<int64_t>((lineStart + cast<int64_t>(1ULL)))])))) + cast<int64_t>(2ULL)));
}
threedo_bitReader brStorage = threedo_bitReader{c->PDAT,cast<int64_t>(((cast<int64_t>((lineStart + offBytes))) * cast<int64_t>(8ULL)))};
threedo_bitReader* br = &brStorage;
for (int64_t x = cast<int64_t>(0ULL);(x < c->Width);){
uint32_t typ = threedo_bitReader_read(br,cast<int64_t>(2ULL));
if ((typ == threedo_packEOL)) {
break;
}
int64_t count = cast<int64_t>((cast<int64_t>(threedo_bitReader_read(br,cast<int64_t>(6ULL))) + cast<int64_t>(1ULL)));
{
switch(typ){
case threedo_packLiteral:{
for (int64_t i = cast<int64_t>(0ULL);((i < count) && (x < c->Width));i++){
set(x,y,threedo_bitReader_read(br,c->BPP));
x++;
}
break;}
case threedo_packTransparent:{
x += count;
break;}
case threedo_packRepeat:{
uint32_t v = threedo_bitReader_read(br,c->BPP);
for (int64_t i = cast<int64_t>(0ULL);((i < count) && (x < c->Width));i++){
set(x,y,v);
x++;
}
break;}
}}
}
int64_t next = cast<int64_t>((lineStart + cast<int64_t>((words * cast<int64_t>(4ULL)))));
if ((next <= lineStart)) {
break;
}
pos = next;
}
}
}
// tools/platform/threedo/cel.go:435:1
void threedo_Cel_decodeUnpacked(threedo_Cel* c,std::function<void(int64_t,int64_t,uint32_t)> set){
{
int64_t woffset={};
if ((c->BPP >= cast<int64_t>(8ULL))) {
woffset = cast<int64_t>(cast<uint32_t>(((shr<uint32_t>(c->PRE1,cast<int64_t>(16ULL))) & cast<uint32_t>(1023ULL))));
}
else {
woffset = cast<int64_t>(cast<uint32_t>(((shr<uint32_t>(c->PRE1,cast<int64_t>(24ULL))) & cast<uint32_t>(255ULL))));
}
int64_t strideBits = cast<int64_t>(((cast<int64_t>((woffset + cast<int64_t>(2ULL)))) * cast<int64_t>(32ULL)));
{
int64_t minBits = cast<int64_t>(((divi<int64_t>((cast<int64_t>((cast<int64_t>((c->Width * c->BPP)) + cast<int64_t>(31ULL)))),cast<int64_t>(32ULL))) * cast<int64_t>(32ULL)));
if ((strideBits < minBits)) {
strideBits = minBits;
}
}
for (int64_t y = cast<int64_t>(0ULL);(y < c->Height);y++){
threedo_bitReader brStorage = threedo_bitReader{c->PDAT,cast<int64_t>((y * strideBits))};
threedo_bitReader* br = &brStorage;
for (int64_t x = cast<int64_t>(0ULL);(x < c->Width);x++){
set(x,y,threedo_bitReader_read(br,c->BPP));
}
}
}
}
// tools/platform/threedo/cvid.go:46:1
std::tuple<threedo_CvidMovie*,Error> threedo_DemuxStream(Slice<uint8_t> data){
{
threedo_CvidMovie* m = arenaNew(threedo_CvidMovie{cast<int64_t>(320ULL),cast<int64_t>(240ULL),{},{},{},{},{},{}});
bool haveHdr = false;
for (int64_t off = cast<int64_t>(0ULL);(cast<int64_t>((off + cast<int64_t>(8ULL))) <= len(data));){
std::string tag = cast<std::string>(sub(data,off,cast<int64_t>((off + cast<int64_t>(4ULL)))));
int64_t size = cast<int64_t>(be_Uint32(sub(data,cast<int64_t>((off + cast<int64_t>(4ULL))),cast<int64_t>((off + cast<int64_t>(8ULL))))));
if (((size < cast<int64_t>(8ULL)) || (cast<int64_t>((off + size)) > len(data)))) {
break;
}
if (((tag == std::string("FILM",4)) && (cast<int64_t>((off + cast<int64_t>(20ULL))) <= len(data)))) {
{
auto tmp13=cast<std::string>(sub(data,cast<int64_t>((off + cast<int64_t>(16ULL))),cast<int64_t>((off + cast<int64_t>(20ULL)))));
if (tmp13==(std::string("FHDR",4))){
if ((cast<int64_t>((off + cast<int64_t>(40ULL))) <= len(data))) {
m->Codec = cast<std::string>(sub(data,cast<int64_t>((off + cast<int64_t>(24ULL))),cast<int64_t>((off + cast<int64_t>(28ULL)))));
m->Height = cast<int64_t>(be_Uint32(sub(data,cast<int64_t>((off + cast<int64_t>(28ULL))),cast<int64_t>((off + cast<int64_t>(32ULL))))));
m->Width = cast<int64_t>(be_Uint32(sub(data,cast<int64_t>((off + cast<int64_t>(32ULL))),cast<int64_t>((off + cast<int64_t>(36ULL))))));
m->HeaderRate = cast<int64_t>(be_Uint32(sub(data,cast<int64_t>((off + cast<int64_t>(36ULL))),cast<int64_t>((off + cast<int64_t>(40ULL))))));
haveHdr = true;
}
}
else if (tmp13==(std::string("FRME",4))){
if (((cast<int64_t>((off + cast<int64_t>(28ULL))) <= cast<int64_t>((off + size))) && (cast<int64_t>((off + cast<int64_t>(28ULL))) <= len(data)))) {
uint32_t t = be_Uint32(sub(data,cast<int64_t>((off + cast<int64_t>(8ULL))),cast<int64_t>((off + cast<int64_t>(12ULL)))));
uint32_t dur = be_Uint32(sub(data,cast<int64_t>((off + cast<int64_t>(20ULL))),cast<int64_t>((off + cast<int64_t>(24ULL)))));
int64_t fsz = cast<int64_t>(be_Uint32(sub(data,cast<int64_t>((off + cast<int64_t>(24ULL))),cast<int64_t>((off + cast<int64_t>(28ULL))))));
int64_t end = cast<int64_t>((cast<int64_t>((off + cast<int64_t>(28ULL))) + fsz));
if (((end > cast<int64_t>((off + size))) || (end > len(data)))) {
end = cast<int64_t>((off + size));
}
m->Frames = append(m->Frames,sub(data,cast<int64_t>((off + cast<int64_t>(28ULL))),end));
m->Durations = append(m->Durations,dur);
m->Times = append(m->Times,t);
}
}
}
tmp12:;
}
off += size;
}
if (((!haveHdr) && (len(m->Frames) == cast<int64_t>(0ULL)))) {
return {{},go_fmt_Errorf(std::string("no FILM video track found in stream",35))};
}
{
int64_t n = len(m->Times);
if (((n >= cast<int64_t>(2ULL)) && (m->Times[cast<int64_t>((n - cast<int64_t>(1ULL)))] > m->Times[cast<int64_t>(0ULL)]))) {
double span = cast<double>(cast<uint32_t>((m->Times[cast<int64_t>((n - cast<int64_t>(1ULL)))] - m->Times[cast<int64_t>(0ULL)])));
m->FPS = cast<int64_t>(((cast<double>(240ULL) / ((span / cast<double>(cast<int64_t>((n - cast<int64_t>(1ULL))))))) + 0.5));
}
}
if ((m->FPS <= cast<int64_t>(0ULL))) {
m->FPS = m->HeaderRate;
}
if ((m->FPS <= cast<int64_t>(0ULL))) {
m->FPS = cast<int64_t>(15ULL);
}
return {m,{}};
}
}
// tools/platform/threedo/cvid.go:120:1
threedo_CvidDecoder* threedo_NewCvidDecoder(int64_t w,int64_t h){
{
return arenaNew(threedo_CvidDecoder{w,h,go_image_NewRGBA(go_image_Rect(cast<int64_t>(0ULL),cast<int64_t>(0ULL),w,h)),{},{}});
}
}
// tools/platform/threedo/cvid.go:126:1
image_RGBA* threedo_CvidDecoder_Frame(threedo_CvidDecoder* d){
{
return d->img;
}
}
// tools/platform/threedo/cvid.go:129:1
int64_t threedo_tdiv2(int64_t x){
{
if ((x < cast<int64_t>(0ULL))) {
return cast<int64_t>(-(divi<int64_t>((cast<int64_t>(-x)),cast<int64_t>(2ULL))));
}
return divi<int64_t>(x,cast<int64_t>(2ULL));
}
}
// tools/platform/threedo/cvid.go:136:1
uint8_t threedo_clamp8(int64_t v){
{
if ((v < cast<int64_t>(0ULL))) {
return cast<uint8_t>(0ULL);
}
if ((v > cast<int64_t>(255ULL))) {
return cast<uint8_t>(255ULL);
}
return cast<uint8_t>(v);
}
}
// tools/platform/threedo/cvid.go:148:1
void threedo_CvidDecoder_setPix(threedo_CvidDecoder* d,int64_t x,int64_t y,uint8_t ly,int8_t u,int8_t v){
{
if (((((x < cast<int64_t>(0ULL)) || (y < cast<int64_t>(0ULL))) || (x >= d->W)) || (y >= d->H))) {
return ;
}
int64_t Y = cast<int64_t>(ly);
int64_t o = image_RGBA_PixOffset(d->img,x,y);
d->img->Pix[o] = threedo_clamp8(cast<int64_t>((Y + cast<int64_t>((cast<int64_t>(2ULL) * cast<int64_t>(v))))));
d->img->Pix[cast<int64_t>((o + cast<int64_t>(1ULL)))] = threedo_clamp8(cast<int64_t>((cast<int64_t>((Y - threedo_tdiv2(cast<int64_t>(u)))) - cast<int64_t>(v))));
d->img->Pix[cast<int64_t>((o + cast<int64_t>(2ULL)))] = threedo_clamp8(cast<int64_t>((Y + cast<int64_t>((cast<int64_t>(2ULL) * cast<int64_t>(u))))));
d->img->Pix[cast<int64_t>((o + cast<int64_t>(3ULL)))] = cast<uint8_t>(255ULL);
}
}
// tools/platform/threedo/cvid.go:161:1
void threedo_CvidDecoder_paintV4(threedo_CvidDecoder* d,int64_t px,int64_t py,threedo_cvidVec a,threedo_cvidVec b,threedo_cvidVec c,threedo_cvidVec e){
{
{auto&& tmp14 = std::array<threedo_cvidVec,4>{a,b,c,e};
for(int64_t tmp15=0;tmp15<len(tmp14);++tmp15){
auto j=tmp15;auto vec=tmp14[tmp15];auto tmp16 = std::make_tuple(cast<int64_t>((px + cast<int64_t>(((cast<int64_t>((j & cast<int64_t>(1ULL)))) * cast<int64_t>(2ULL))))),cast<int64_t>((py + cast<int64_t>(((shr<int64_t>(j,cast<int64_t>(1ULL))) * cast<int64_t>(2ULL))))));
int64_t ox = std::get<0>(tmp16);
int64_t oy = std::get<1>(tmp16);
threedo_CvidDecoder_setPix(d,ox,oy,vec.y[cast<int64_t>(0ULL)],vec.u,vec.v);
threedo_CvidDecoder_setPix(d,cast<int64_t>((ox + cast<int64_t>(1ULL))),oy,vec.y[cast<int64_t>(1ULL)],vec.u,vec.v);
threedo_CvidDecoder_setPix(d,ox,cast<int64_t>((oy + cast<int64_t>(1ULL))),vec.y[cast<int64_t>(2ULL)],vec.u,vec.v);
threedo_CvidDecoder_setPix(d,cast<int64_t>((ox + cast<int64_t>(1ULL))),cast<int64_t>((oy + cast<int64_t>(1ULL))),vec.y[cast<int64_t>(3ULL)],vec.u,vec.v);
}}
}
}
// tools/platform/threedo/cvid.go:173:1
void threedo_CvidDecoder_paintV1(threedo_CvidDecoder* d,int64_t px,int64_t py,threedo_cvidVec vec){
{
for (int64_t j = cast<int64_t>(0ULL);(j < cast<int64_t>(4ULL));j++){
auto tmp17 = std::make_tuple(cast<int64_t>((px + cast<int64_t>(((cast<int64_t>((j & cast<int64_t>(1ULL)))) * cast<int64_t>(2ULL))))),cast<int64_t>((py + cast<int64_t>(((shr<int64_t>(j,cast<int64_t>(1ULL))) * cast<int64_t>(2ULL))))));
int64_t ox = std::get<0>(tmp17);
int64_t oy = std::get<1>(tmp17);
uint8_t c = vec.y[j];
threedo_CvidDecoder_setPix(d,ox,oy,c,vec.u,vec.v);
threedo_CvidDecoder_setPix(d,cast<int64_t>((ox + cast<int64_t>(1ULL))),oy,c,vec.u,vec.v);
threedo_CvidDecoder_setPix(d,ox,cast<int64_t>((oy + cast<int64_t>(1ULL))),c,vec.u,vec.v);
threedo_CvidDecoder_setPix(d,cast<int64_t>((ox + cast<int64_t>(1ULL))),cast<int64_t>((oy + cast<int64_t>(1ULL))),c,vec.u,vec.v);
}
}
}
// tools/platform/threedo/cvid.go:187:1
void threedo_decodeCodebook(Slice<threedo_cvidVec> book,uint16_t cid,Slice<uint8_t> data){
{
int64_t nbytes = cast<int64_t>(6ULL);
if ((cast<uint16_t>((cid & cast<uint16_t>(1024ULL))) != cast<uint16_t>(0ULL))) {
nbytes = cast<int64_t>(4ULL);
}
bool flagged = (cast<uint16_t>((cid & cast<uint16_t>(256ULL))) != cast<uint16_t>(0ULL));
int64_t d = cast<int64_t>(0ULL);
std::function<bool(int64_t)> read = [&](int64_t i)->bool{
if ((cast<int64_t>((d + nbytes)) > len(data))) {
return false;
}
threedo_cvidVec e={};
auto tmp18 = std::make_tuple(data[d],data[cast<int64_t>((d + cast<int64_t>(1ULL)))],data[cast<int64_t>((d + cast<int64_t>(2ULL)))],data[cast<int64_t>((d + cast<int64_t>(3ULL)))]);
e.y[cast<int64_t>(0ULL)] = std::get<0>(tmp18);
e.y[cast<int64_t>(1ULL)] = std::get<1>(tmp18);
e.y[cast<int64_t>(2ULL)] = std::get<2>(tmp18);
e.y[cast<int64_t>(3ULL)] = std::get<3>(tmp18);
if ((nbytes == cast<int64_t>(6ULL))) {
auto tmp19 = std::make_tuple(cast<int8_t>(data[cast<int64_t>((d + cast<int64_t>(4ULL)))]),cast<int8_t>(data[cast<int64_t>((d + cast<int64_t>(5ULL)))]));
e.u = std::get<0>(tmp19);
e.v = std::get<1>(tmp19);
}
book[i] = e;
d += nbytes;
return true;
}
;
if ((!flagged)) {
for (int64_t i = cast<int64_t>(0ULL);((i < cast<int64_t>(256ULL)) && read(i));i++){
}
return ;
}
for (int64_t i = cast<int64_t>(0ULL);((i < cast<int64_t>(256ULL)) && (d < len(data)));){
if ((cast<int64_t>((d + cast<int64_t>(4ULL))) > len(data))) {
return ;
}
uint32_t flag = be_Uint32(sub(data,d,len(data)));
d += cast<int64_t>(4ULL);
for (int64_t b = cast<int64_t>(0ULL);((b < cast<int64_t>(32ULL)) && (i < cast<int64_t>(256ULL)));b++){
if ((cast<uint32_t>((flag & cast<uint32_t>(2147483648ULL))) != cast<uint32_t>(0ULL))) {
if ((!read(i))) {
return ;
}
}
flag = shl<uint32_t>(flag,cast<int64_t>(1ULL));
i++;
}
}
}
}
// tools/platform/threedo/cvid.go:234:1
void threedo_CvidDecoder_decodeVectors(threedo_CvidDecoder* d,uint16_t cid,Slice<uint8_t> data,int64_t x0,int64_t y0,int64_t x1,int64_t y1,Slice<threedo_cvidVec> v1,Slice<threedo_cvidVec> v4){
{
uint32_t flag={};
uint32_t mask={};
int64_t pos = cast<int64_t>(0ULL);
std::function<bool()> need = [&]()->bool{
mask = shr<uint32_t>(mask,cast<int64_t>(1ULL));
if ((mask == cast<uint32_t>(0ULL))) {
if ((cast<int64_t>((pos + cast<int64_t>(4ULL))) > len(data))) {
return false;
}
flag = be_Uint32(sub(data,pos,len(data)));
pos += cast<int64_t>(4ULL);
mask = cast<uint32_t>(2147483648ULL);
}
return true;
}
;
int64_t mbx = divi<int64_t>((cast<int64_t>((x1 - x0))),cast<int64_t>(4ULL));
int64_t mby = divi<int64_t>((cast<int64_t>((y1 - y0))),cast<int64_t>(4ULL));
bool inter = (cast<uint16_t>((cid & cast<uint16_t>(256ULL))) != cast<uint16_t>(0ULL));
bool v1only = (cast<uint16_t>((cid & cast<uint16_t>(512ULL))) != cast<uint16_t>(0ULL));
for (int64_t by = cast<int64_t>(0ULL);(by < mby);by++){
for (int64_t bx = cast<int64_t>(0ULL);(bx < mbx);bx++){
auto tmp20 = std::make_tuple(cast<int64_t>((x0 + cast<int64_t>((bx * cast<int64_t>(4ULL))))),cast<int64_t>((y0 + cast<int64_t>((by * cast<int64_t>(4ULL))))));
int64_t px = std::get<0>(tmp20);
int64_t py = std::get<1>(tmp20);
if (inter) {
if ((!need())) {
return ;
}
if ((cast<uint32_t>((flag & mask)) == cast<uint32_t>(0ULL))) {
continue;
}
}
bool useV4 = false;
if ((!v1only)) {
if ((!need())) {
return ;
}
useV4 = (cast<uint32_t>((flag & mask)) != cast<uint32_t>(0ULL));
}
if (useV4) {
if ((cast<int64_t>((pos + cast<int64_t>(4ULL))) > len(data))) {
return ;
}
threedo_CvidDecoder_paintV4(d,px,py,v4[data[pos]],v4[data[cast<int64_t>((pos + cast<int64_t>(1ULL)))]],v4[data[cast<int64_t>((pos + cast<int64_t>(2ULL)))]],v4[data[cast<int64_t>((pos + cast<int64_t>(3ULL)))]]);
pos += cast<int64_t>(4ULL);
}
else {
if ((cast<int64_t>((pos + cast<int64_t>(1ULL))) > len(data))) {
return ;
}
threedo_CvidDecoder_paintV1(d,px,py,v1[data[pos]]);
pos++;
}
}
}
}
}
// tools/platform/threedo/cvid.go:290:1
void threedo_CvidDecoder_DecodeFrame(threedo_CvidDecoder* d,Slice<uint8_t> fr){
{
if ((len(fr) < cast<int64_t>(10ULL))) {
return ;
}
uint8_t flags = fr[cast<int64_t>(0ULL)];
int64_t ln = cast<int64_t>((cast<int64_t>((shl<int64_t>(cast<int64_t>(fr[cast<int64_t>(1ULL)]),cast<int64_t>(16ULL)) | shl<int64_t>(cast<int64_t>(fr[cast<int64_t>(2ULL)]),cast<int64_t>(8ULL)))) | cast<int64_t>(fr[cast<int64_t>(3ULL)])));
int64_t p = cast<int64_t>(10ULL);
int64_t ystart = cast<int64_t>(0ULL);
int64_t sidx = cast<int64_t>(0ULL);
for (;((cast<int64_t>((p + cast<int64_t>(4ULL))) <= len(fr)) && (p < cast<int64_t>((cast<int64_t>(4ULL) + ln))));){
uint16_t sid = be_Uint16(sub(fr,p,len(fr)));
int64_t ssize = cast<int64_t>(be_Uint16(sub(fr,cast<int64_t>((p + cast<int64_t>(2ULL))),len(fr))));
if (((((sid != cast<uint16_t>(4096ULL)) && (sid != cast<uint16_t>(4352ULL)))) || (ssize < cast<int64_t>(12ULL)))) {
if ((ssize < cast<int64_t>(4ULL))) {
ssize = cast<int64_t>(4ULL);
}
p += ssize;
continue;
}
int64_t ytop = cast<int64_t>(be_Uint16(sub(fr,cast<int64_t>((p + cast<int64_t>(4ULL))),len(fr))));
int64_t ybot = cast<int64_t>(be_Uint16(sub(fr,cast<int64_t>((p + cast<int64_t>(8ULL))),len(fr))));
int64_t y0 = ystart;
int64_t y1 = cast<int64_t>((ystart + (cast<int64_t>((ybot - ytop)))));
for (;(len(d->v1) <= sidx);){
d->v1 = append(d->v1,Slice<threedo_cvidVec>::make(cast<int64_t>(256ULL)));
}
for (;(len(d->v4) <= sidx);){
d->v4 = append(d->v4,Slice<threedo_cvidVec>::make(cast<int64_t>(256ULL)));
}
if (((sidx > cast<int64_t>(0ULL)) && (cast<uint8_t>((flags & cast<uint8_t>(1ULL))) == cast<uint8_t>(0ULL)))) {
gcopy(d->v1[sidx],d->v1[cast<int64_t>((sidx - cast<int64_t>(1ULL)))]);
gcopy(d->v4[sidx],d->v4[cast<int64_t>((sidx - cast<int64_t>(1ULL)))]);
}
auto tmp21 = std::make_tuple(d->v1[sidx],d->v4[sidx]);
Slice<threedo_cvidVec> v1 = std::get<0>(tmp21);
Slice<threedo_cvidVec> v4 = std::get<1>(tmp21);
int64_t end = cast<int64_t>((p + ssize));
int64_t q = cast<int64_t>((p + cast<int64_t>(12ULL)));
for (;(cast<int64_t>((q + cast<int64_t>(4ULL))) <= end);){
uint16_t cid = be_Uint16(sub(fr,q,len(fr)));
int64_t csize = cast<int64_t>(be_Uint16(sub(fr,cast<int64_t>((q + cast<int64_t>(2ULL))),len(fr))));
if (((csize < cast<int64_t>(4ULL)) || (cast<int64_t>((q + csize)) > end))) {
break;
}
Slice<uint8_t> payload = sub(fr,cast<int64_t>((q + cast<int64_t>(4ULL))),cast<int64_t>((q + csize)));
uint16_t hi = shr<uint16_t>(cid,cast<int64_t>(8ULL));
{
if (((hi >= cast<uint16_t>(32ULL)) && (hi <= cast<uint16_t>(39ULL)))){
Slice<threedo_cvidVec> book = v4;
if (((((hi == cast<uint16_t>(34ULL)) || (hi == cast<uint16_t>(35ULL))) || (hi == cast<uint16_t>(38ULL))) || (hi == cast<uint16_t>(39ULL)))) {
book = v1;
}
threedo_decodeCodebook(book,cid,payload);
}
else if ((((hi == cast<uint16_t>(48ULL)) || (hi == cast<uint16_t>(49ULL))) || (hi == cast<uint16_t>(50ULL)))){
threedo_CvidDecoder_decodeVectors(d,cid,payload,cast<int64_t>(0ULL),y0,d->W,y1,v1,v4);
}
}
tmp22:;
q += csize;
}
p = end;
ystart = y1;
sidx++;
}
}
}
// tools/platform/threedo/debug.go:59:1
std::string threedo_CelDraw_Name(threedo_CelDraw c){
{
{
if (c.LRForm){
return std::string("CEL LRFORM16",12);
}
else if (c.Packed){
return std::string("CEL packed",10);
}
else {
return std::string("CEL unpacked",12);
}
}
tmp23:;
}
}
// tools/platform/threedo/debug.go:71:1
std::string threedo_CelDraw_Decoded(threedo_CelDraw c){
{
std::string persp = std::string("",0);
if (((((c.HDDX != cast<int32_t>(0ULL)) || (c.HDDY != cast<int32_t>(0ULL))) || (c.VDX != cast<int32_t>(0ULL))) || (c.HDY != cast<int32_t>(0ULL)))) {
persp = std::string(" persp",6);
}
return go_fmt_Sprintf(std::string("%dbpp %dx%d at (%d,%d)%s pixc=%08X flags=%08X",45),c.BPP,c.Width,c.Height,shr<int32_t>(c.XPos,cast<int64_t>(16ULL)),shr<int32_t>(c.YPos,cast<int64_t>(16ULL)),persp,c.PIXC,c.Flags);
}
}
// tools/platform/threedo/debug.go:82:1
Slice<uint64_t> threedo_CelDraw_Words(threedo_CelDraw c){
{
return Slice<uint64_t>{cast<uint64_t>(c.Flags),cast<uint64_t>(c.Src),cast<uint64_t>(c.PLUT),cast<uint64_t>(cast<uint32_t>(c.XPos)),cast<uint64_t>(cast<uint32_t>(c.YPos)),cast<uint64_t>(cast<uint32_t>(c.HDX)),cast<uint64_t>(cast<uint32_t>(c.HDY)),cast<uint64_t>(cast<uint32_t>(c.VDX)),cast<uint64_t>(cast<uint32_t>(c.VDY)),cast<uint64_t>(cast<uint32_t>(c.HDDX)),cast<uint64_t>(cast<uint32_t>(c.HDDY)),cast<uint64_t>(c.PIXC),cast<uint64_t>(c.PRE0),cast<uint64_t>(c.PRE1)};
}
}
// tools/platform/threedo/debug.go:104:1
void threedo_Machine_celDraw(threedo_Machine* m,threedo_gfxBitmap bm,uint32_t ccb,uint32_t flags,uint32_t src,uint32_t plutPtr){
{
m->celCnt.cels++;
if ((!m->OnCel)) {
return ;
}
threedo_CelDraw c = threedo_CelDraw{cast<int64_t>((m->celCount - cast<int64_t>(1ULL))),ccb,flags,src,plutPtr,cast<int32_t>(threedo_Machine_read32(m,cast<uint32_t>((ccb + cast<uint32_t>(16ULL))))),cast<int32_t>(threedo_Machine_read32(m,cast<uint32_t>((ccb + cast<uint32_t>(20ULL))))),cast<int32_t>(threedo_Machine_read32(m,cast<uint32_t>((ccb + cast<uint32_t>(24ULL))))),cast<int32_t>(threedo_Machine_read32(m,cast<uint32_t>((ccb + cast<uint32_t>(28ULL))))),cast<int32_t>(threedo_Machine_read32(m,cast<uint32_t>((ccb + cast<uint32_t>(32ULL))))),cast<int32_t>(threedo_Machine_read32(m,cast<uint32_t>((ccb + cast<uint32_t>(36ULL))))),cast<int32_t>(threedo_Machine_read32(m,cast<uint32_t>((ccb + cast<uint32_t>(40ULL))))),cast<int32_t>(threedo_Machine_read32(m,cast<uint32_t>((ccb + cast<uint32_t>(44ULL))))),threedo_Machine_read32(m,cast<uint32_t>((ccb + cast<uint32_t>(48ULL)))),{},{},{},{},{},{},{},{},bm.buf,bm.w,bm.h};
uint32_t s = src;
if ((cast<uint32_t>((flags & threedo_ccbCCBPre)) != cast<uint32_t>(0ULL))) {
auto tmp24 = std::make_tuple(threedo_Machine_read32(m,cast<uint32_t>((ccb + cast<uint32_t>(52ULL)))),threedo_Machine_read32(m,cast<uint32_t>((ccb + cast<uint32_t>(56ULL)))));
c.PRE0 = std::get<0>(tmp24);
c.PRE1 = std::get<1>(tmp24);
}
else {
c.PRE0 = threedo_Machine_read32(m,s);
s += cast<uint32_t>(4ULL);
if ((cast<uint32_t>((flags & threedo_ccbPacked)) == cast<uint32_t>(0ULL))) {
c.PRE1 = threedo_Machine_read32(m,s);
}
}
c.Packed = (cast<uint32_t>((flags & threedo_ccbPacked)) != cast<uint32_t>(0ULL));
c.BPP = get(threedo_bppFromPRE0,cast<uint32_t>((c.PRE0 & cast<uint32_t>(7ULL))));
c.Coded = (c.BPP <= cast<int64_t>(8ULL));
c.LRForm = (cast<uint32_t>((c.PRE1 & threedo_pre1LRForm)) != cast<uint32_t>(0ULL));
c.Width = cast<int64_t>((cast<int64_t>(cast<uint32_t>((c.PRE1 & cast<uint32_t>(2047ULL)))) + cast<int64_t>(1ULL)));
c.Height = cast<int64_t>((cast<int64_t>(cast<uint32_t>(((shr<uint32_t>(c.PRE0,cast<int64_t>(6ULL))) & cast<uint32_t>(1023ULL)))) + cast<int64_t>(1ULL)));
if (c.Packed) {
c.Width = bm.w;
}
m->OnCel(c);
}
}
// tools/platform/threedo/debug.go:144:1
void threedo_Machine_celPixel(threedo_Machine* m,threedo_gfxBitmap bm,int64_t x,int64_t y,uint16_t pix){
{
m->celCnt.pixels++;
if ((((!m->OnPixel) || (x < cast<int64_t>(0ULL))) || (y < cast<int64_t>(0ULL)))) {
return ;
}
auto tmp25 = threedo_chan5(pix);
uint32_t r = std::get<0>(tmp25);
uint32_t g = std::get<1>(tmp25);
uint32_t b = std::get<2>(tmp25);
m->OnPixel(cast<uint32_t>(x),cast<uint32_t>(y),threedo_PixelEvent{true,cast<uint8_t>(shl<uint32_t>(r,cast<int64_t>(3ULL))),cast<uint8_t>(shl<uint32_t>(g,cast<int64_t>(3ULL))),cast<uint8_t>(shl<uint32_t>(b,cast<int64_t>(3ULL))),cast<uint8_t>(255ULL)});
}
}
// tools/platform/threedo/debug.go:165:1
threedo_Result threedo_Machine_RunStopAfterCel(threedo_Machine* m,int64_t k,uint64_t budget){
{
auto tmp26 = std::make_tuple(k,cast<int64_t>(0ULL));
m->celLimit = std::get<0>(tmp26);
m->celCount = std::get<1>(tmp26);
threedo_Result r = threedo_Machine_Run(m,budget);
auto tmp27 = std::make_tuple(cast<int64_t>(0ULL),cast<int64_t>(0ULL));
m->celLimit = std::get<0>(tmp27);
m->celCount = std::get<1>(tmp27);
return r;
}
}
// tools/platform/threedo/debug.go:173:1
int64_t threedo_Machine_Cels(threedo_Machine* m){
{
return m->celCount;
}
}
// tools/platform/threedo/debug.go:181:1
std::tuple<uint32_t,int64_t,int64_t,bool> threedo_Machine_DrawTarget(threedo_Machine* m){
uint32_t buf{};
int64_t w{};
int64_t h{};
bool ok{};
{
{auto&& tmp28 = m->bitmaps;
for(auto [tmp29,tmp30]:tmp28){
auto bm=tmp30;if (((bm.buf != cast<uint32_t>(0ULL)) && (bm.buf != m->displayBuf))) {
return {bm.buf,bm.w,bm.h,true};
}
}}
{auto&& tmp31 = m->bitmaps;
for(auto [tmp32,tmp33]:tmp31){
auto bm=tmp33;if ((bm.buf != cast<uint32_t>(0ULL))) {
return {bm.buf,bm.w,bm.h,true};
}
}}
return {cast<uint32_t>(0ULL),cast<int64_t>(0ULL),cast<int64_t>(0ULL),false};
}
}
// tools/platform/threedo/debug.go:201:1
Slice<threedo_MemRegion> threedo_Machine_Bitmaps(threedo_Machine* m){
{
Slice<threedo_MemRegion> out={};
{auto&& tmp34 = m->bitmaps;
for(auto [tmp35,tmp36]:tmp34){
auto bm=tmp36;if ((bm.buf != cast<uint32_t>(0ULL))) {
out = append(out,threedo_MemRegion{go_fmt_Sprintf(std::string("bitmap %08X",11),bm.buf),bm.buf,cast<uint32_t>(cast<int64_t>((cast<int64_t>((bm.w * bm.h)) * cast<int64_t>(2ULL))))});
}
}}
return out;
}
}
// tools/platform/threedo/debug.go:219:1
Slice<threedo_MemRegion> threedo_Machine_MemRegions(threedo_Machine* m){
{
return Slice<threedo_MemRegion>{threedo_MemRegion{std::string("dram",4),cast<uint32_t>(0ULL),threedo_dramSize},threedo_MemRegion{std::string("vram",4),threedo_vramBase,threedo_vramSize},threedo_MemRegion{std::string("madam",5),threedo_madamBase,cast<uint32_t>((threedo_madamEnd - threedo_madamBase))},threedo_MemRegion{std::string("clio",4),threedo_clioBase,cast<uint32_t>((threedo_clioEnd - threedo_clioBase))}};
}
}
// tools/platform/threedo/debug.go:231:1
std::tuple<image_RGBA*,Error> threedo_Machine_RenderBitmap(threedo_Machine* m,uint32_t buf,int64_t w,int64_t h){
{
if (((((w <= cast<int64_t>(0ULL)) || (h <= cast<int64_t>(0ULL))) || (w > cast<int64_t>(2048ULL))) || (h > cast<int64_t>(2048ULL)))) {
return {{},go_fmt_Errorf(std::string("threedo: bitmap size %dx%d out of range",39),w,h)};
}
return {threedo_Machine_CaptureVRAM(m,buf,w,h),{}};
}
}
// tools/platform/threedo/debug.go:239:1
uint64_t threedo_Machine_Frames(threedo_Machine* m){
{
return m->frame;
}
}
// tools/platform/threedo/debug.go:242:1
threedo_Volume* threedo_Machine_Volume(threedo_Machine* m){
{
return m->vol;
}
}
// tools/platform/threedo/filefolio.go:54:1
void threedo_Machine_serviceFileFolio(threedo_Machine* m,uint32_t foff){
{
arm60_CPU* c = m->CPU;
{
switch(foff){
case cast<uint32_t>(4ULL):{
threedo_Machine_SetResultAndReturn(m,threedo_Machine_openDiskStream(m,threedo_Machine_readCStr(m,arm60_CPU_Reg(c,cast<uint32_t>(0ULL)))));
break;}
case cast<uint32_t>(8ULL):{
threedo_Machine_SetResultAndReturn(m,cast<uint32_t>(threedo_Machine_readDiskStream(m,arm60_CPU_Reg(c,cast<uint32_t>(0ULL)),arm60_CPU_Reg(c,cast<uint32_t>(1ULL)),cast<int32_t>(arm60_CPU_Reg(c,cast<uint32_t>(2ULL))))));
break;}
case cast<uint32_t>(12ULL):{
threedo_Machine_SetResultAndReturn(m,cast<uint32_t>(threedo_Machine_seekDiskStream(m,arm60_CPU_Reg(c,cast<uint32_t>(0ULL)),cast<int32_t>(arm60_CPU_Reg(c,cast<uint32_t>(1ULL))),arm60_CPU_Reg(c,cast<uint32_t>(2ULL)))));
break;}
case cast<uint32_t>(16ULL):{
threedo_Machine_closeDiskStream(m,arm60_CPU_Reg(c,cast<uint32_t>(0ULL)));
threedo_Machine_SetResultAndReturn(m,cast<uint32_t>(0ULL));
break;}
case cast<uint32_t>(32ULL):{
threedo_Machine_SetResultAndReturn(m,threedo_Machine_openDirectoryPath(m,threedo_Machine_readCStr(m,arm60_CPU_Reg(c,cast<uint32_t>(0ULL)))));
break;}
case cast<uint32_t>(36ULL):{
threedo_Machine_SetResultAndReturn(m,cast<uint32_t>(threedo_Machine_readDirectory(m,arm60_CPU_Reg(c,cast<uint32_t>(0ULL)),arm60_CPU_Reg(c,cast<uint32_t>(1ULL)))));
break;}
case cast<uint32_t>(40ULL):{
threedo_Machine_closeDirectory(m,arm60_CPU_Reg(c,cast<uint32_t>(0ULL)));
threedo_Machine_SetResultAndReturn(m,cast<uint32_t>(0ULL));
break;}
default:{
threedo_Machine_note(m,go_fmt_Sprintf(std::string("FileFolio[-0x%X] stub (r0=0x%08X %q r1=0x%08X r2=0x%08X)",56),foff,arm60_CPU_Reg(c,cast<uint32_t>(0ULL)),threedo_Machine_readCStr(m,arm60_CPU_Reg(c,cast<uint32_t>(0ULL))),arm60_CPU_Reg(c,cast<uint32_t>(1ULL)),arm60_CPU_Reg(c,cast<uint32_t>(2ULL))));
threedo_Machine_SetResultAndReturn(m,cast<uint32_t>(0ULL));
break;}
}}
}
}
// tools/platform/threedo/filefolio.go:83:1
void threedo_Machine_serviceOtherFolio(threedo_Machine* m,uint32_t foff){
{
arm60_CPU* c = m->CPU;
threedo_Machine_note(m,go_fmt_Sprintf(std::string("otherFolio[-0x%X] from 0x%08X (r0=0x%08X r1=0x%08X r2=0x%08X r3=0x%08X)",71),foff,cast<uint32_t>((arm60_CPU_Reg(c,cast<uint32_t>(14ULL)) - cast<uint32_t>(8ULL))),arm60_CPU_Reg(c,cast<uint32_t>(0ULL)),arm60_CPU_Reg(c,cast<uint32_t>(1ULL)),arm60_CPU_Reg(c,cast<uint32_t>(2ULL)),arm60_CPU_Reg(c,cast<uint32_t>(3ULL))));
threedo_Machine_SetResultAndReturn(m,cast<uint32_t>(0ULL));
}
}
// tools/platform/threedo/filefolio.go:93:1
uint32_t threedo_Machine_openDiskStream(threedo_Machine* m,std::string name){
{
auto tmp37 = threedo_Machine_loadDiscFile(m,name);
Slice<uint8_t> data = std::get<0>(tmp37);
std::string path = std::get<1>(tmp37);
bool ok = std::get<2>(tmp37);
if ((!ok)) {
threedo_Machine_note(m,go_fmt_Sprintf(std::string("OpenDiskStream(%q) -> NOT FOUND",31),name));
return cast<uint32_t>(0ULL);
}
uint32_t handle = threedo_heap_alloc(m->dheap,cast<uint32_t>(32ULL));
if ((handle == cast<uint32_t>(0ULL))) {
return cast<uint32_t>(0ULL);
}
m->streams[handle] = arenaNew(threedo_diskStream{path,data,{}});
threedo_Machine_note(m,go_fmt_Sprintf(std::string("OpenDiskStream(%q) -> handle 0x%08X (%s, %d bytes)",50),name,handle,path,len(data)));
return handle;
}
}
// tools/platform/threedo/filefolio.go:111:1
int32_t threedo_Machine_readDiskStream(threedo_Machine* m,uint32_t handle,uint32_t buf,int32_t n){
{
threedo_diskStream* s = get(m->streams,handle);
if ((!s)) {
return cast<int32_t>(-cast<int64_t>(1ULL));
}
if ((n < cast<int32_t>(0ULL))) {
n = cast<int32_t>(0ULL);
}
int64_t avail = cast<int64_t>((len(s->data) - s->pos));
if ((avail <= cast<int64_t>(0ULL))) {
return cast<int32_t>(0ULL);
}
if ((cast<int64_t>(n) > avail)) {
n = cast<int32_t>(avail);
}
for (int32_t i = cast<int32_t>(cast<int32_t>(0ULL));(i < n);i++){
threedo_Machine_Write(m,cast<uint32_t>((buf + cast<uint32_t>(i))),s->data[cast<int64_t>((s->pos + cast<int64_t>(i)))]);
}
s->pos += cast<int64_t>(n);
return n;
}
}
// tools/platform/threedo/filefolio.go:137:1
int32_t threedo_Machine_seekDiskStream(threedo_Machine* m,uint32_t handle,int32_t offset,uint32_t whence){
{
threedo_diskStream* s = get(m->streams,handle);
if ((!s)) {
return cast<int32_t>(-cast<int64_t>(1ULL));
}
int64_t pos={};
{
switch(whence){
case threedo_seekSet:{
pos = cast<int64_t>(offset);
break;}
case threedo_seekCur:{
pos = cast<int64_t>((s->pos + cast<int64_t>(offset)));
break;}
case threedo_seekEnd:{
pos = cast<int64_t>((len(s->data) - cast<int64_t>(offset)));
break;}
default:{
return cast<int32_t>(-cast<int64_t>(1ULL));
break;}
}}
if (((pos < cast<int64_t>(0ULL)) || (pos > len(s->data)))) {
return cast<int32_t>(-cast<int64_t>(1ULL));
}
s->pos = pos;
return cast<int32_t>(pos);
}
}
// tools/platform/threedo/filefolio.go:172:1
uint32_t threedo_Machine_openDirectoryPath(threedo_Machine* m,std::string path){
{
threedo_dirScan* scan = arenaNew(threedo_dirScan{});
std::string trimmed = go_strings_TrimLeft(path,std::string("/",1));
if ((!go_strings_HasPrefix(go_strings_ToLower(trimmed),std::string("nvram",5)))) {
if ((!m->vol)) {
return cast<uint32_t>(0ULL);
}
auto tmp38 = threedo_Volume_ReadDir(m->vol,trimmed);
Slice<threedo_Entry> entries = std::get<0>(tmp38);
Error err = std::get<1>(tmp38);
if (bool(err)) {
threedo_Machine_note(m,go_fmt_Sprintf(std::string("OpenDirectoryPath(%q) -> NOT FOUND",34),path));
return cast<uint32_t>(0ULL);
}
scan->entries = entries;
}
uint32_t handle = threedo_heap_alloc(m->dheap,cast<uint32_t>(32ULL));
if ((handle == cast<uint32_t>(0ULL))) {
return cast<uint32_t>(0ULL);
}
m->dirs[handle] = scan;
threedo_Machine_note(m,go_fmt_Sprintf(std::string("OpenDirectoryPath(%q) -> dir 0x%08X (%d entries)",48),path,handle,len(scan->entries)));
return handle;
}
}
// tools/platform/threedo/filefolio.go:199:1
int32_t threedo_Machine_readDirectory(threedo_Machine* m,uint32_t dir,uint32_t buf){
{
threedo_dirScan* s = get(m->dirs,dir);
if (((!s) || (s->pos >= len(s->entries)))) {
return cast<int32_t>(-cast<int64_t>(1ULL));
}
threedo_Entry e = s->entries[s->pos];
s->pos++;
uint32_t flags={};
uint32_t typ={};
if (e.IsDir) {
flags = cast<uint32_t>(1ULL);
typ = cast<uint32_t>(711223666ULL);
}
else {
typ = cast<uint32_t>(707406378ULL);
}
uint32_t blocks = cast<uint32_t>(divi<int64_t>((cast<int64_t>((e.Size + cast<int64_t>(2047ULL)))),cast<int64_t>(2048ULL)));
{auto&& tmp39 = Slice<uint32_t>{flags,cast<uint32_t>(s->pos),typ,cast<uint32_t>(2048ULL),cast<uint32_t>(e.Size),blocks,cast<uint32_t>(0ULL),cast<uint32_t>(0ULL),cast<uint32_t>((cast<uint32_t>(len(e.Copies)) + cast<uint32_t>(1ULL)))};
for(int64_t tmp40=0;tmp40<len(tmp39);++tmp40){
auto i=tmp40;auto w=tmp39[tmp40];threedo_Machine_writeWord(m,cast<uint32_t>((buf + cast<uint32_t>((cast<uint32_t>(i) * cast<uint32_t>(4ULL))))),w);
}}
std::string name = e.Name;
if ((len(name) > cast<int64_t>(31ULL))) {
name = sub(name,0,cast<int64_t>(31ULL));
}
for (int64_t i = cast<int64_t>(0ULL);(i < cast<int64_t>(32ULL));i++){
uint8_t b = cast<uint8_t>(cast<uint8_t>(0ULL));
if ((i < len(name))) {
b = name[i];
}
threedo_Machine_Write(m,cast<uint32_t>((cast<uint32_t>((buf + cast<uint32_t>(36ULL))) + cast<uint32_t>(i))),b);
}
threedo_Machine_writeWord(m,cast<uint32_t>((buf + cast<uint32_t>(68ULL))),cast<uint32_t>(e.Block));
return cast<int32_t>(0ULL);
}
}
// tools/platform/threedo/filefolio.go:243:1
void threedo_Machine_closeDirectory(threedo_Machine* m,uint32_t dir){
{
{
auto tmp41 = lookup(m->dirs,dir);
bool ok = std::get<1>(tmp41);
if (ok) {
removeKey(m->dirs,dir);
threedo_heap_freeBlock(m->dheap,dir);
}
}
}
}
// tools/platform/threedo/filefolio.go:251:1
void threedo_Machine_closeDiskStream(threedo_Machine* m,uint32_t handle){
{
{
auto tmp42 = lookup(m->streams,handle);
bool ok = std::get<1>(tmp42);
if (ok) {
// Release closed stream buffers; only the tiny arena token survives reset.
if(auto stream=get(m->streams,handle))stream->data={};
removeKey(m->streams,handle);
threedo_heap_freeBlock(m->dheap,handle);
}
}
}
}
// tools/platform/threedo/filefolio.go:279:1
std::tuple<std::string,bool> threedo_nvramPath(std::string path){
{
std::string trimmed = go_strings_TrimLeft(path,std::string("/",1));
std::string low = go_strings_ToLower(trimmed);
if ((!go_strings_HasPrefix(low,std::string("nvram",5)))) {
return {std::string("",0),false};
}
std::string rest = go_strings_TrimLeft(sub(trimmed,len(std::string("nvram",5)),len(trimmed)),std::string("/",1));
return {go_strings_ToLower(rest),true};
}
}
// tools/platform/threedo/filefolio.go:291:1
bool threedo_Machine_fileFolioSWI(threedo_Machine* m,arm60_CPU* c,uint32_t swi){
{
{
switch(swi){
case threedo_swiOpenDiskFile:{
arm60_CPU_SetReg(c,cast<uint32_t>(0ULL),threedo_Machine_openDiskFile(m,threedo_Machine_readCStr(m,arm60_CPU_Reg(c,cast<uint32_t>(0ULL)))));
break;}
case threedo_swiCloseDiskFile:{
{
threedo_item* it = get(m->items,cast<int32_t>(arm60_CPU_Reg(c,cast<uint32_t>(0ULL))));
if (bool(it)) {
removeKey(m->items,it->num);
}
}
arm60_CPU_SetReg(c,cast<uint32_t>(0ULL),cast<uint32_t>(0ULL));
break;}
case threedo_swiCreateFile:{
std::string path = threedo_Machine_readCStr(m,arm60_CPU_Reg(c,cast<uint32_t>(0ULL)));
{
auto tmp43 = threedo_nvramPath(path);
std::string key = std::get<0>(tmp43);
bool ok = std::get<1>(tmp43);
if (ok) {
{
auto tmp44 = lookup(m->nvram,key);
bool exists = std::get<1>(tmp44);
if ((!exists)) {
m->nvram[key] = {};
}
}
threedo_Machine_note(m,go_fmt_Sprintf(std::string("CreateFile(%q)",14),path));
arm60_CPU_SetReg(c,cast<uint32_t>(0ULL),cast<uint32_t>(0ULL));
}
else {
arm60_CPU_SetReg(c,cast<uint32_t>(0ULL),threedo_fileErrNotFound);
}
}
break;}
case threedo_swiDeleteFile:{
std::string path = threedo_Machine_readCStr(m,arm60_CPU_Reg(c,cast<uint32_t>(0ULL)));
{
auto tmp45 = threedo_nvramPath(path);
std::string key = std::get<0>(tmp45);
bool ok = std::get<1>(tmp45);
if (ok) {
removeKey(m->nvram,key);
threedo_Machine_note(m,go_fmt_Sprintf(std::string("DeleteFile(%q)",14),path));
}
}
arm60_CPU_SetReg(c,cast<uint32_t>(0ULL),cast<uint32_t>(0ULL));
break;}
case threedo_swiChangeDir:case threedo_swiGetDir:{
arm60_CPU_SetReg(c,cast<uint32_t>(0ULL),cast<uint32_t>(0ULL));
break;}
default:{
return false;
break;}
}}
return true;
}
}
// tools/platform/threedo/filefolio.go:329:1
uint32_t threedo_Machine_openDiskFile(threedo_Machine* m,std::string path){
{
{
auto tmp46 = threedo_nvramPath(path);
std::string key = std::get<0>(tmp46);
bool ok = std::get<1>(tmp46);
if (ok) {
{
auto tmp47 = lookup(m->nvram,key);
bool exists = std::get<1>(tmp47);
if ((!exists)) {
threedo_Machine_note(m,go_fmt_Sprintf(std::string("OpenDiskFile(%q) -> no such NVRAM file",38),path));
return threedo_fileErrNotFound;
}
}
threedo_item* it = threedo_Machine_createItem(m,cast<uint32_t>(781ULL),cast<uint32_t>(0ULL),cast<uint32_t>(0ULL));
it->name = (std::string("/nvram/",7) + key);
threedo_Machine_note(m,go_fmt_Sprintf(std::string("OpenDiskFile(%q) -> item %d (nvram)",35),path,it->num));
return cast<uint32_t>(it->num);
}
}
auto tmp48 = threedo_Machine_loadDiscFile(m,path);
std::string resolved = std::get<1>(tmp48);
bool ok = std::get<2>(tmp48);
if ((!ok)) {
threedo_Machine_note(m,go_fmt_Sprintf(std::string("OpenDiskFile(%q) -> NOT FOUND",29),path));
return threedo_fileErrNotFound;
}
threedo_item* it = threedo_Machine_createItem(m,cast<uint32_t>(781ULL),cast<uint32_t>(0ULL),cast<uint32_t>(0ULL));
it->name = resolved;
threedo_Machine_note(m,go_fmt_Sprintf(std::string("OpenDiskFile(%q) -> item %d (%s)",32),path,it->num,resolved));
return cast<uint32_t>(it->num);
}
}
// tools/platform/threedo/filefolio.go:354:1
std::tuple<Slice<uint8_t>,uint32_t,std::string,bool> threedo_Machine_fileData(threedo_Machine* m,std::string name){
Slice<uint8_t> data{};
uint32_t blockSize{};
std::string nvramKey{};
bool ok{};
{
{
auto tmp49 = threedo_nvramPath(name);
std::string key = std::get<0>(tmp49);
bool isNV = std::get<1>(tmp49);
if (isNV) {
auto tmp50 = lookup(m->nvram,key);
Slice<uint8_t> d = std::get<0>(tmp50);
bool exists = std::get<1>(tmp50);
return {d,cast<uint32_t>(1ULL),key,exists};
}
}
auto tmp51 = threedo_Machine_loadDiscFile(m,name);
Slice<uint8_t> d = std::get<0>(tmp51);
bool found = std::get<2>(tmp51);
return {d,cast<uint32_t>(2048ULL),std::string("",0),found};
}
}
// tools/platform/threedo/filefolio.go:367:1
std::tuple<Slice<uint8_t>,std::string,bool> threedo_Machine_loadDiscFile(threedo_Machine* m,std::string name){
{
if (((!m->vol) || (name == std::string("",0)))) {
return {{},std::string("",0),false};
}
if ((((m->NoStreams || m->MovieHLE)) && go_strings_Contains(go_strings_ToLower(name),std::string(".stream",7)))) {
if (m->MovieHLE) {
threedo_Machine_armMovie(m,name);
}
return {{},std::string("",0),false};
}
name = go_strings_TrimLeft(name,std::string("/",1));
{
auto tmp52 = threedo_Volume_ReadFile(m->vol,name);
Slice<uint8_t> data = std::get<0>(tmp52);
Error err = std::get<1>(tmp52);
if ((!err)) {
return {data,name,true};
}
}
std::string base = name;
{
int64_t i = go_strings_LastIndexByte(base,'/');
if ((i >= cast<int64_t>(0ULL))) {
base = sub(base,cast<int64_t>((i + cast<int64_t>(1ULL))),len(base));
}
}
Slice<uint8_t> data={};
std::string path={};
bool found = false;
threedo_Volume_Walk(m->vol,[&](threedo_Entry e)->Error{
if ((((!found) && (!e.IsDir)) && go_strings_EqualFold(e.Name,base))) {
{
auto tmp53 = threedo_Volume_ReadFile(m->vol,e.Path);
Slice<uint8_t> d = std::get<0>(tmp53);
Error err = std::get<1>(tmp53);
if ((!err)) {
auto tmp54 = std::make_tuple(d,e.Path,true);
data = std::get<0>(tmp54);
path = std::get<1>(tmp54);
found = std::get<2>(tmp54);
}
}
}
return {};
}
);
return {data,path,found};
}
}
// tools/platform/threedo/folio.go:34:1
threedo_heap* threedo_newHeap(uint32_t base,uint32_t total){
{
return arenaNew(threedo_heap{base,total,Slice<threedo_span>{threedo_span{base,total}},Map<uint32_t,uint32_t>{}});
}
}
// tools/platform/threedo/folio.go:39:1
uint32_t threedo_heap_alloc(threedo_heap* h,uint32_t size){
{
if ((size == cast<uint32_t>(0ULL))) {
return cast<uint32_t>(0ULL);
}
size = ((cast<uint32_t>((size + cast<uint32_t>(15ULL)))) & ~(cast<uint32_t>(15ULL)));
{auto&& tmp55 = h->free;
for(int64_t tmp56=0;tmp56<len(tmp55);++tmp56){
auto i=tmp56;auto s=tmp55[tmp56];if ((s.size >= size)) {
uint32_t addr = s.addr;
if ((s.size == size)) {
h->free = append(sub(h->free,0,i),sub(h->free,cast<int64_t>((i + cast<int64_t>(1ULL))),len(h->free)));
}
else {
h->free[i] = threedo_span{cast<uint32_t>((s.addr + size)),cast<uint32_t>((s.size - size))};
}
h->live[addr] = size;
return addr;
}
}}
return cast<uint32_t>(0ULL);
}
}
// tools/platform/threedo/folio.go:60:1
void threedo_heap_freeBlock(threedo_heap* h,uint32_t addr){
{
auto tmp57 = lookup(h->live,addr);
uint32_t size = std::get<0>(tmp57);
bool ok = std::get<1>(tmp57);
if ((!ok)) {
return ;
}
removeKey(h->live,addr);
h->free = append(h->free,threedo_span{addr,size});
for (int64_t i = cast<int64_t>(0ULL);(i < len(h->free));i++){
for (int64_t j = cast<int64_t>((i + cast<int64_t>(1ULL)));(j < len(h->free));j++){
if ((h->free[j].addr < h->free[i].addr)) {
auto tmp58 = std::make_tuple(h->free[j],h->free[i]);
h->free[i] = std::get<0>(tmp58);
h->free[j] = std::get<1>(tmp58);
}
}
}
Slice<threedo_span> merged = sub(h->free,0,cast<int64_t>(1ULL));
{auto&& tmp59 = sub(h->free,cast<int64_t>(1ULL),len(h->free));
for(int64_t tmp60=0;tmp60<len(tmp59);++tmp60){
auto s=tmp59[tmp60];threedo_span* last = (&merged[cast<int64_t>((len(merged) - cast<int64_t>(1ULL)))]);
if ((cast<uint32_t>((last->addr + last->size)) == s.addr)) {
last->size += s.size;
}
else {
merged = append(merged,s);
}
}}
h->free = merged;
}
}
// tools/platform/threedo/folio.go:90:1
bool threedo_Machine_serviceFolio(threedo_Machine* m,uint32_t off){
{
{
switch(off){
case cast<uint32_t>(28ULL):{
uint32_t size = arm60_CPU_Reg(m->CPU,cast<uint32_t>(1ULL));
uint32_t flags = arm60_CPU_Reg(m->CPU,cast<uint32_t>(2ULL));
uint32_t ptr = threedo_heap_alloc(threedo_Machine_poolFor(m,flags),size);
threedo_Machine_note(m,go_fmt_Sprintf(std::string("AllocMem(size=0x%X, flags=0x%X) -> 0x%08X",41),size,flags,ptr));
threedo_Machine_SetResultAndReturn(m,ptr);
return true;
break;}
case cast<uint32_t>(32ULL):{
uint32_t ptr = arm60_CPU_Reg(m->CPU,cast<uint32_t>(1ULL));
threedo_heap_freeBlock(threedo_Machine_poolOf(m,ptr),ptr);
threedo_Machine_SetResultAndReturn(m,cast<uint32_t>(0ULL));
return true;
break;}
case cast<uint32_t>(48ULL):{
threedo_Machine_SetResultAndReturn(m,threedo_Machine_lookupItem(m,cast<int32_t>(arm60_CPU_Reg(m->CPU,cast<uint32_t>(0ULL)))));
return true;
break;}
case cast<uint32_t>(56ULL):{
auto tmp61 = std::make_tuple(arm60_CPU_Reg(m->CPU,cast<uint32_t>(0ULL)),arm60_CPU_Reg(m->CPU,cast<uint32_t>(1ULL)),arm60_CPU_Reg(m->CPU,cast<uint32_t>(2ULL)));
uint32_t dst = std::get<0>(tmp61);
uint32_t src = std::get<1>(tmp61);
uint32_t n = std::get<2>(tmp61);
if ((n > cast<uint32_t>(4194304ULL))) {
n = cast<uint32_t>(4194304ULL);
}
for (uint32_t i = cast<uint32_t>(cast<uint32_t>(0ULL));(i < n);i++){
threedo_Machine_Write(m,cast<uint32_t>((dst + i)),threedo_Machine_Read(m,cast<uint32_t>((src + i))));
}
threedo_Machine_SetResultAndReturn(m,dst);
return true;
break;}
case cast<uint32_t>(60ULL):{
if ((cast<uint32_t>((arm60_CPU_Reg(m->CPU,cast<uint32_t>(0ULL)) & cast<uint32_t>(65536ULL))) != cast<uint32_t>(0ULL))) {
threedo_Machine_SetResultAndReturn(m,cast<uint32_t>(2048ULL));
}
else {
threedo_Machine_SetResultAndReturn(m,cast<uint32_t>(4096ULL));
}
return true;
break;}
case cast<uint32_t>(96ULL):{
threedo_Machine_waitPort(m);
return true;
break;}
case cast<uint32_t>(52ULL):{
m->simTime += threedo_simTick;
uint32_t buf = arm60_CPU_Reg(m->CPU,cast<uint32_t>(0ULL));
threedo_Machine_writeWord(m,buf,cast<uint32_t>(divi<uint64_t>(m->simTime,cast<uint64_t>(1000000ULL))));
threedo_Machine_writeWord(m,cast<uint32_t>((buf + cast<uint32_t>(4ULL))),cast<uint32_t>(modi<uint64_t>(m->simTime,cast<uint64_t>(1000000ULL))));
threedo_Machine_SetResultAndReturn(m,cast<uint32_t>(0ULL));
return true;
break;}
}}
return false;
}
}
// tools/platform/threedo/folio.go:163:1
uint32_t threedo_Machine_lookupItem(threedo_Machine* m,int32_t num){
{
threedo_item* it = get(m->items,num);
if ((!it)) {
return cast<uint32_t>(0ULL);
}
if ((cast<uint32_t>((it->typ & cast<uint32_t>(255ULL))) != cast<uint32_t>(4ULL))) {
return it->addr;
}
{
if (go_strings_EqualFold(it->name,std::string("File",4))){
return threedo_fileFolioBase;
}
else if (go_strings_EqualFold(it->name,std::string("Graphics",8))){
return threedo_gfxFolioBase;
}
else if (go_strings_EqualFold(it->name,std::string("audio",5))){
return threedo_audioFolioBase;
}
else if (go_strings_EqualFold(it->name,std::string("Operamath",9))){
return threedo_mathFolioBase;
}
else if (go_strings_EqualFold(it->name,std::string("kernel",6))){
return threedo_kernelBase;
}
else {
threedo_Machine_note(m,go_fmt_Sprintf(std::string("LookupItem: folio %q -> generic stub window",43),it->name));
return threedo_otherFolioBase;
}
}
tmp62:;
}
}
// tools/platform/threedo/folio.go:201:1
threedo_heap* threedo_Machine_poolFor(threedo_Machine* m,uint32_t flags){
{
if ((cast<uint32_t>((flags & threedo_memtypeVRAM)) != cast<uint32_t>(0ULL))) {
return m->vheap;
}
return m->dheap;
}
}
// tools/platform/threedo/folio.go:209:1
threedo_heap* threedo_Machine_poolOf(threedo_Machine* m,uint32_t ptr){
{
if (((ptr >= threedo_vheapBase) && (ptr < threedo_vheapTop))) {
return m->vheap;
}
return m->dheap;
}
}
// tools/platform/threedo/framebuffer.go:11:1
image_RGBA* threedo_Machine_CaptureVRAM(threedo_Machine* m,uint32_t base,int64_t w,int64_t h){
{
image_RGBA* img = go_image_NewRGBA(go_image_Rect(cast<int64_t>(0ULL),cast<int64_t>(0ULL),w,h));
for (int64_t y = cast<int64_t>(0ULL);(y < h);y++){
for (int64_t x = cast<int64_t>(0ULL);(x < w);x++){
uint32_t o = cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((base - threedo_vramBase)) + cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(shr<int64_t>(y,cast<int64_t>(1ULL))) * cast<uint32_t>(w))) * cast<uint32_t>(4ULL))))) + cast<uint32_t>((cast<uint32_t>(x) * cast<uint32_t>(4ULL))))) + cast<uint32_t>((cast<uint32_t>(cast<int64_t>((y & cast<int64_t>(1ULL)))) * cast<uint32_t>(2ULL)))));
if ((cast<int64_t>((cast<int64_t>(o) + cast<int64_t>(2ULL))) > len(m->vram))) {
continue;
}
uint16_t c = cast<uint16_t>((shl<uint16_t>(cast<uint16_t>(m->vram[o]),cast<int64_t>(8ULL)) | cast<uint16_t>(m->vram[cast<uint32_t>((o + cast<uint32_t>(1ULL)))])));
image_RGBA_SetRGBA(img,x,y,threedo_rgb555(c));
}
}
return img;
}
}
// tools/platform/threedo/framebuffer.go:28:1
int64_t threedo_Machine_VRAMNonZero(threedo_Machine* m,int64_t n){
{
if ((n > len(m->vram))) {
n = len(m->vram);
}
int64_t c = cast<int64_t>(0ULL);
for (int64_t i = cast<int64_t>(0ULL);(i < n);i++){
if ((m->vram[i] != cast<uint8_t>(0ULL))) {
c++;
}
}
return c;
}
}
// tools/platform/threedo/graphicsfolio.go:76:1
void threedo_Machine_flashClearRange(threedo_Machine* m,uint32_t dest,uint32_t bytes,uint16_t val){
rrprof::Scope rrclock(3,"Flash clear");
{
m->celCnt.clears++;
auto tmp63=defer([&](){threedo_Machine_profEnd(m,threedo_bucketClear,threedo_Machine_profStart(m));});
auto tmp64 = std::make_tuple(cast<uint8_t>(shr<uint16_t>(val,cast<int64_t>(8ULL))),cast<uint8_t>(val));
uint8_t hi = std::get<0>(tmp64);
uint8_t lo = std::get<1>(tmp64);
uint32_t end = cast<uint32_t>((dest + bytes));
for (uint32_t a = dest;(cast<uint32_t>((a + cast<uint32_t>(1ULL))) < end);a += cast<uint32_t>(2ULL)){
threedo_Machine_Write(m,a,hi);
threedo_Machine_Write(m,cast<uint32_t>((a + cast<uint32_t>(1ULL))),lo);
}
if (m->CelDebug) {
m->CelDebugLog = append(m->CelDebugLog,go_fmt_Sprintf(std::string("FLASHCLEAR val=%04X -> [0x%08X,+0x%X)",37),val,dest,bytes));
}
}
}
// tools/platform/threedo/graphicsfolio.go:97:1
std::string threedo_gfxFuncName(uint32_t foff){
{
Map<uint32_t,std::string> names = Map<uint32_t,std::string>{{cast<uint32_t>(4ULL),std::string("MapSprite",9)},{cast<uint32_t>(8ULL),std::string("ReadPixel",9)},{cast<uint32_t>(12ULL),std::string("ReadVDLColor",12)},{cast<uint32_t>(16ULL),std::string("GetPixelAddress",15)},{cast<uint32_t>(20ULL),std::string("WritePixel",10)},{cast<uint32_t>(36ULL),std::string("DrawText16",10)},{cast<uint32_t>(44ULL),std::string("ResetFont",9)},{cast<uint32_t>(48ULL),std::string("CreateScreenGroup",17)},{cast<uint32_t>(52ULL),std::string("SetReadAddress",14)},{cast<uint32_t>(56ULL),std::string("ResetReadAddress",16)},{cast<uint32_t>(60ULL),std::string("SetClipOrigin",13)},{cast<uint32_t>(80ULL),std::string("SetScreenColor",14)},{cast<uint32_t>(88ULL),std::string("SetScreenColors",15)},{cast<uint32_t>(92ULL),std::string("DeleteScreenGroup",17)},{cast<uint32_t>(96ULL),std::string("SetFGPen",8)},{cast<uint32_t>(100ULL),std::string("SetBGPen",8)},{cast<uint32_t>(104ULL),std::string("AddScreenGroup",14)},{cast<uint32_t>(108ULL),std::string("RemoveScreenGroup",17)},{cast<uint32_t>(112ULL),std::string("SetClipWidth",12)},{cast<uint32_t>(116ULL),std::string("SetClipHeight",13)},{cast<uint32_t>(120ULL),std::string("MoveTo",6)},{cast<uint32_t>(128ULL),std::string("DrawChar",8)},{cast<uint32_t>(132ULL),std::string("DrawTo",6)},{cast<uint32_t>(136ULL),std::string("FillRect",8)},{cast<uint32_t>(144ULL),std::string("GetCurrentFont",14)},{cast<uint32_t>(148ULL),std::string("DrawText8",9)},{cast<uint32_t>(152ULL),std::string("SetCEControl",12)},{cast<uint32_t>(160ULL),std::string("DisplayScreen",13)},{cast<uint32_t>(164ULL),std::string("SetVDL",6)},{cast<uint32_t>(168ULL),std::string("SubmitVDL",9)},{cast<uint32_t>(172ULL),std::string("DrawCels",8)},{cast<uint32_t>(176ULL),std::string("DrawScreenCels",14)},{cast<uint32_t>(180ULL),std::string("SetCEWatchDog",13)},{cast<uint32_t>(184ULL),std::string("GetFirstDisplayInfo",19)},{cast<uint32_t>(188ULL),std::string("ModifyVDL",9)},{cast<uint32_t>(192ULL),std::string("RegisterVBLCounter",18)}};
{
auto tmp65 = lookup(names,foff);
std::string n = std::get<0>(tmp65);
bool ok = std::get<1>(tmp65);
if (ok) {
return n;
}
}
return std::string("?",1);
}
}
// tools/platform/threedo/graphicsfolio.go:122:1
void threedo_Machine_serviceGraphicsFolio(threedo_Machine* m,uint32_t foff){
{
arm60_CPU* c = m->CPU;
auto tmp66 = std::make_tuple(arm60_CPU_Reg(c,cast<uint32_t>(0ULL)),arm60_CPU_Reg(c,cast<uint32_t>(1ULL)),arm60_CPU_Reg(c,cast<uint32_t>(2ULL)));
uint32_t r0 = std::get<0>(tmp66);
uint32_t r1 = std::get<1>(tmp66);
uint32_t r2 = std::get<2>(tmp66);
threedo_Machine_note(m,go_fmt_Sprintf(std::string("Graphics[-0x%X] %s from 0x%08X (r0=0x%08X r1=0x%08X r2=0x%08X)",62),foff,threedo_gfxFuncName(foff),cast<uint32_t>((arm60_CPU_Reg(c,cast<uint32_t>(14ULL)) - cast<uint32_t>(8ULL))),r0,r1,r2));
{
switch(foff){
case cast<uint32_t>(48ULL):{
threedo_Machine_SetResultAndReturn(m,threedo_Machine_createScreenGroup(m,r0,r1));
break;}
case cast<uint32_t>(16ULL):{
threedo_Machine_SetResultAndReturn(m,threedo_Machine_pixelAddress(m,cast<int32_t>(r0),r1,r2));
break;}
case cast<uint32_t>(168ULL):{
threedo_item* it = threedo_Machine_createItem(m,cast<uint32_t>(518ULL),cast<uint32_t>(0ULL),cast<uint32_t>(0ULL));
threedo_Machine_SetResultAndReturn(m,cast<uint32_t>(it->num));
break;}
case cast<uint32_t>(164ULL):{
threedo_Machine_SetResultAndReturn(m,cast<uint32_t>(0ULL));
break;}
case cast<uint32_t>(160ULL):{
{
auto tmp67 = lookup(m->screenBM,cast<int32_t>(r0));
int32_t bmItem = std::get<0>(tmp67);
bool ok = std::get<1>(tmp67);
if (ok) {
{
auto tmp68 = lookup(m->bitmaps,bmItem);
threedo_gfxBitmap bm = std::get<0>(tmp68);
bool ok = std::get<1>(tmp68);
if (ok) {
m->displayBuf = bm.buf;
}
}
}
}
if (m->CelDebug) {
if ((len(m->CelDebugLog) > cast<int64_t>(0ULL))) {
m->CelFrameLog = m->CelDebugLog;
m->CelDebugLog = {};
}
}
m->frame++;
threedo_Machine_profFrame(m);
if (bool(m->OnDisplay)) {
m->OnDisplay(m,m->frame,m->displayBuf);
}
threedo_Machine_SetResultAndReturn(m,cast<uint32_t>(0ULL));
break;}
case cast<uint32_t>(172ULL):{
threedo_Machine_SetResultAndReturn(m,threedo_Machine_drawCels(m,cast<int32_t>(r0),r1));
break;}
case cast<uint32_t>(176ULL):{
{
auto tmp69 = lookup(m->screenBM,cast<int32_t>(r0));
int32_t bmItem = std::get<0>(tmp69);
bool ok = std::get<1>(tmp69);
if (ok) {
threedo_Machine_SetResultAndReturn(m,threedo_Machine_drawCels(m,bmItem,r1));
return ;
}
}
threedo_Machine_SetResultAndReturn(m,cast<uint32_t>(0ULL));
break;}
case cast<uint32_t>(192ULL):{
if ((r1 != cast<uint32_t>(0ULL))) {
threedo_Machine_SetVBLMirror(m,r1);
}
threedo_Machine_SetResultAndReturn(m,cast<uint32_t>(0ULL));
break;}
default:{
threedo_Machine_SetResultAndReturn(m,cast<uint32_t>(0ULL));
break;}
}}
}
}
// tools/platform/threedo/graphicsfolio.go:186:1
uint32_t threedo_Machine_createScreenGroup(threedo_Machine* m,uint32_t itemArray,uint32_t tags){
{
auto tmp70 = std::make_tuple(cast<uint32_t>(cast<uint32_t>(1ULL)),cast<uint32_t>(cast<uint32_t>(1ULL)));
uint32_t screenCount = std::get<0>(tmp70);
uint32_t bitmapCount = std::get<1>(tmp70);
auto tmp71 = std::make_tuple(cast<uint32_t>(cast<uint32_t>(320ULL)),cast<uint32_t>(cast<uint32_t>(240ULL)));
uint32_t width = std::get<0>(tmp71);
uint32_t height = std::get<1>(tmp71);
uint32_t bufArray={};
uint32_t widthArray={};
uint32_t heightArray={};
for (uint32_t p = tags;(p != cast<uint32_t>(0ULL));p += cast<uint32_t>(8ULL)){
auto tmp72 = std::make_tuple(threedo_Machine_read32(m,p),threedo_Machine_read32(m,cast<uint32_t>((p + cast<uint32_t>(4ULL)))));
uint32_t tag = std::get<0>(tmp72);
uint32_t val = std::get<1>(tmp72);
if ((tag == cast<uint32_t>(0ULL))) {
break;
}
{
switch(cast<uint32_t>((tag & cast<uint32_t>(65535ULL)))){
case threedo_csgScreenCount:{
screenCount = val;
break;}
case threedo_csgBitmapCount:{
bitmapCount = val;
break;}
case threedo_csgScreenHeight:case threedo_csgDisplayHeight:{
height = val;
break;}
case threedo_csgBitmapWidths:{
widthArray = val;
break;}
case threedo_csgBitmapHeights:{
heightArray = val;
break;}
case threedo_csgBitmapBufs:{
bufArray = val;
break;}
}}
}
if (((((screenCount == cast<uint32_t>(0ULL)) || (screenCount > cast<uint32_t>(8ULL))) || (bitmapCount == cast<uint32_t>(0ULL))) || (bitmapCount > cast<uint32_t>(8ULL)))) {
return cast<uint32_t>(~cast<uint32_t>(cast<uint32_t>(0ULL)));
}
threedo_item* group = threedo_Machine_createItem(m,cast<uint32_t>(513ULL),cast<uint32_t>(0ULL),cast<uint32_t>(0ULL));
for (uint32_t s = cast<uint32_t>(cast<uint32_t>(0ULL));(s < screenCount);s++){
threedo_item* firstBM={};
for (uint32_t b = cast<uint32_t>(cast<uint32_t>(0ULL));(b < bitmapCount);b++){
auto tmp73 = std::make_tuple(width,height);
uint32_t w = std::get<0>(tmp73);
uint32_t h = std::get<1>(tmp73);
if ((widthArray != cast<uint32_t>(0ULL))) {
w = threedo_Machine_read32(m,cast<uint32_t>((widthArray + cast<uint32_t>((b * cast<uint32_t>(4ULL))))));
}
if ((heightArray != cast<uint32_t>(0ULL))) {
h = threedo_Machine_read32(m,cast<uint32_t>((heightArray + cast<uint32_t>((b * cast<uint32_t>(4ULL))))));
}
uint32_t buf = cast<uint32_t>(cast<uint32_t>(0ULL));
if ((bufArray != cast<uint32_t>(0ULL))) {
buf = threedo_Machine_read32(m,cast<uint32_t>((bufArray + cast<uint32_t>(((cast<uint32_t>((cast<uint32_t>((s * bitmapCount)) + b))) * cast<uint32_t>(4ULL))))));
}
if ((buf == cast<uint32_t>(0ULL))) {
buf = threedo_heap_alloc(m->vheap,cast<uint32_t>((cast<uint32_t>((w * h)) * cast<uint32_t>(2ULL))));
}
threedo_item* bm = threedo_Machine_createItem(m,cast<uint32_t>(515ULL),cast<uint32_t>(0ULL),cast<uint32_t>(96ULL));
threedo_Machine_writeWord(m,cast<uint32_t>((bm->addr + threedo_bmBuffer)),buf);
threedo_Machine_writeWord(m,cast<uint32_t>((bm->addr + threedo_bmWidth)),w);
threedo_Machine_writeWord(m,cast<uint32_t>((bm->addr + threedo_bmHeight)),h);
threedo_Machine_writeWord(m,cast<uint32_t>((bm->addr + threedo_bmClipWidth)),w);
threedo_Machine_writeWord(m,cast<uint32_t>((bm->addr + threedo_bmClipHeight)),h);
m->bitmaps[bm->num] = threedo_gfxBitmap{buf,cast<int64_t>(w),cast<int64_t>(h)};
if ((!firstBM)) {
firstBM = bm;
}
}
threedo_item* scr = threedo_Machine_createItem(m,cast<uint32_t>(514ULL),cast<uint32_t>(0ULL),cast<uint32_t>(144ULL));
threedo_Machine_writeWord(m,cast<uint32_t>((scr->addr + threedo_scrBitmapCount)),bitmapCount);
threedo_Machine_writeWord(m,cast<uint32_t>((scr->addr + threedo_scrTempBitmap)),firstBM->addr);
m->screenBM[scr->num] = firstBM->num;
if ((m->displayBuf == cast<uint32_t>(0ULL))) {
m->displayBuf = get(m->bitmaps,firstBM->num).buf;
}
if ((itemArray != cast<uint32_t>(0ULL))) {
threedo_Machine_writeWord(m,cast<uint32_t>((itemArray + cast<uint32_t>((s * cast<uint32_t>(4ULL))))),cast<uint32_t>(scr->num));
}
threedo_Machine_note(m,go_fmt_Sprintf(std::string("  screen %d: item %d bitmap %d buf=0x%08X %dx%d",47),s,scr->num,firstBM->num,get(m->bitmaps,firstBM->num).buf,width,height));
}
return cast<uint32_t>(group->num);
}
}
// tools/platform/threedo/graphicsfolio.go:264:1
uint32_t threedo_Machine_pixelAddress(threedo_Machine* m,int32_t itemNum,uint32_t x,uint32_t y){
{
int32_t bmItem = itemNum;
{
auto tmp74 = lookup(m->screenBM,itemNum);
int32_t bi = std::get<0>(tmp74);
bool ok = std::get<1>(tmp74);
if (ok) {
bmItem = bi;
}
}
auto tmp75 = lookup(m->bitmaps,bmItem);
threedo_gfxBitmap bm = std::get<0>(tmp75);
bool ok = std::get<1>(tmp75);
if ((!ok)) {
return cast<uint32_t>(0ULL);
}
return cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((bm.buf + cast<uint32_t>((cast<uint32_t>(((shr<uint32_t>(y,cast<int64_t>(1ULL))) * cast<uint32_t>(bm.w))) * cast<uint32_t>(4ULL))))) + cast<uint32_t>((x * cast<uint32_t>(4ULL))))) + cast<uint32_t>(((cast<uint32_t>((y & cast<uint32_t>(1ULL)))) * cast<uint32_t>(2ULL)))));
}
}
// tools/platform/threedo/graphicsfolio.go:281:1
uint32_t threedo_Machine_drawCels(threedo_Machine* m,int32_t bitmapItem,uint32_t ccb){
{
auto tmp76 = lookup(m->bitmaps,bitmapItem);
threedo_gfxBitmap bm = std::get<0>(tmp76);
bool ok = std::get<1>(tmp76);
if (((!ok) || (bm.buf == cast<uint32_t>(0ULL)))) {
threedo_Machine_note(m,go_fmt_Sprintf(std::string("DrawCels: unknown bitmap item %d",32),bitmapItem));
return cast<uint32_t>(0ULL);
}
m->celCnt.chains++;
auto tmp77 = std::make_tuple(cast<int64_t>(0ULL),cast<int64_t>(0ULL),cast<int64_t>(0ULL));
int64_t drawn = std::get<0>(tmp77);
int64_t skipped = std::get<1>(tmp77);
int64_t total = std::get<2>(tmp77);
for (int64_t n = cast<int64_t>(0ULL);((ccb != cast<uint32_t>(0ULL)) && (n < cast<int64_t>(4096ULL)));n++){
total++;
uint32_t flags = threedo_Machine_read32(m,ccb);
uint32_t next = threedo_Machine_read32(m,cast<uint32_t>((ccb + cast<uint32_t>(4ULL))));
uint32_t src = threedo_Machine_read32(m,cast<uint32_t>((ccb + cast<uint32_t>(8ULL))));
uint32_t plut = threedo_Machine_read32(m,cast<uint32_t>((ccb + cast<uint32_t>(12ULL))));
if (((cast<uint32_t>((flags & threedo_ccbNPAbs)) == cast<uint32_t>(0ULL)) && (next != cast<uint32_t>(0ULL)))) {
next = cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((ccb + cast<uint32_t>(4ULL))) + next)) + cast<uint32_t>(4ULL)));
}
if (((cast<uint32_t>((flags & threedo_ccbSPAbs)) == cast<uint32_t>(0ULL)) && (src != cast<uint32_t>(0ULL)))) {
src = cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((ccb + cast<uint32_t>(8ULL))) + src)) + cast<uint32_t>(4ULL)));
}
if (((cast<uint32_t>((flags & threedo_ccbPPAbs)) == cast<uint32_t>(0ULL)) && (plut != cast<uint32_t>(0ULL)))) {
plut = cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((ccb + cast<uint32_t>(12ULL))) + plut)) + cast<uint32_t>(4ULL)));
}
if ((cast<uint32_t>((flags & threedo_ccbSkip)) == cast<uint32_t>(0ULL))) {
if (((m->celLimit > cast<int64_t>(0ULL)) && (m->celCount >= m->celLimit))) {
m->StopRequested = true;
return cast<uint32_t>(0ULL);
}
m->celCount++;
threedo_Machine_celDraw(m,bm,ccb,flags,src,plut);
time_Time tc = threedo_Machine_profStart(m);
if (threedo_Machine_drawOneCel(m,bm,ccb,flags,src,plut)) {
drawn++;
}
threedo_Machine_profEnd(m,threedo_bucketCel,tc);
}
else {
skipped++;
}
if ((cast<uint32_t>((flags & threedo_ccbLast)) != cast<uint32_t>(0ULL))) {
break;
}
ccb = next;
}
if ((drawn > cast<int64_t>(0ULL))) {
threedo_Machine_note(m,go_fmt_Sprintf(std::string("DrawCels: %d cel(s) -> buf 0x%08X",33),drawn,bm.buf));
}
if (m->CelDebug) {
m->CelDebugLog = append(m->CelDebugLog,go_fmt_Sprintf(std::string("== DrawCels chain: %d total, %d drawn, %d skipped -> buf 0x%08X ==",66),total,drawn,skipped,bm.buf));
}
return cast<uint32_t>(0ULL);
}
}
// tools/platform/threedo/graphicsfolio.go:351:1
bool threedo_Machine_drawOneCel(threedo_Machine* m,threedo_gfxBitmap bm,uint32_t ccb,uint32_t flags,uint32_t src,uint32_t plutPtr){
rrprof::Scope rrclock(2,"Cel software rasterizer");
rrcapture::RenderScope traceRendering;
{
if ((src == cast<uint32_t>(0ULL))) {
return false;
}
int64_t xPos = cast<int64_t>(cast<int32_t>(threedo_Machine_read32(m,cast<uint32_t>((ccb + cast<uint32_t>(16ULL))))));
int64_t yPos = cast<int64_t>(cast<int32_t>(threedo_Machine_read32(m,cast<uint32_t>((ccb + cast<uint32_t>(20ULL))))));
int64_t hdx = cast<int64_t>(cast<int32_t>(threedo_Machine_read32(m,cast<uint32_t>((ccb + cast<uint32_t>(24ULL))))));
int64_t hdy = cast<int64_t>(cast<int32_t>(threedo_Machine_read32(m,cast<uint32_t>((ccb + cast<uint32_t>(28ULL))))));
int64_t vdx = cast<int64_t>(cast<int32_t>(threedo_Machine_read32(m,cast<uint32_t>((ccb + cast<uint32_t>(32ULL))))));
int64_t vdy = cast<int64_t>(cast<int32_t>(threedo_Machine_read32(m,cast<uint32_t>((ccb + cast<uint32_t>(36ULL))))));
int64_t hddx = cast<int64_t>(cast<int32_t>(threedo_Machine_read32(m,cast<uint32_t>((ccb + cast<uint32_t>(40ULL))))));
int64_t hddy = cast<int64_t>(cast<int32_t>(threedo_Machine_read32(m,cast<uint32_t>((ccb + cast<uint32_t>(44ULL))))));
uint32_t pixc = threedo_Machine_read32(m,cast<uint32_t>((ccb + cast<uint32_t>(48ULL))));
uint32_t pre0={};
uint32_t pre1={};
if ((cast<uint32_t>((flags & threedo_ccbCCBPre)) != cast<uint32_t>(0ULL))) {
pre0 = threedo_Machine_read32(m,cast<uint32_t>((ccb + cast<uint32_t>(52ULL))));
pre1 = threedo_Machine_read32(m,cast<uint32_t>((ccb + cast<uint32_t>(56ULL))));
}
else {
pre0 = threedo_Machine_read32(m,src);
src += cast<uint32_t>(4ULL);
if ((cast<uint32_t>((flags & threedo_ccbPacked)) == cast<uint32_t>(0ULL))) {
pre1 = threedo_Machine_read32(m,src);
src += cast<uint32_t>(4ULL);
}
}
threedo_Cel celStorage = threedo_Cel{flags,pre0,pre1,{},cast<int64_t>((cast<int64_t>(cast<uint32_t>((pre1 & cast<uint32_t>(2047ULL)))) + cast<int64_t>(1ULL))),cast<int64_t>((cast<int64_t>(cast<uint32_t>(((shr<uint32_t>(pre0,cast<int64_t>(6ULL))) & cast<uint32_t>(1023ULL)))) + cast<int64_t>(1ULL))),get(threedo_bppFromPRE0,cast<uint32_t>((pre0 & cast<uint32_t>(7ULL)))),(cast<uint32_t>((flags & threedo_ccbPacked)) != cast<uint32_t>(0ULL)),{},{},{}};
threedo_Cel* cel = &celStorage;
if ((cel->BPP == cast<int64_t>(0ULL))) {
return false;
}
cel->Coded = (cel->BPP <= cast<int64_t>(8ULL));
bool lrform = (cast<uint32_t>((pre1 & threedo_pre1LRForm)) != cast<uint32_t>(0ULL));
if (cel->Packed) {
cel->Width = bm.w;
}
if (((((cel->Width <= cast<int64_t>(0ULL)) || (cel->Height <= cast<int64_t>(0ULL))) || (cel->Width > cast<int64_t>(2048ULL))) || (cel->Height > cast<int64_t>(2048ULL)))) {
return false;
}
int64_t max = cast<int64_t>((cast<int64_t>((cast<int64_t>((cast<int64_t>((cel->Height * (divi<int64_t>((cast<int64_t>((cast<int64_t>((cel->Width * cel->BPP)) + cast<int64_t>(31ULL)))),cast<int64_t>(32ULL))))) * cast<int64_t>(4ULL))) + cast<int64_t>((cel->Height * cast<int64_t>(4ULL))))) + cast<int64_t>(64ULL)));
if ((max > cast<int64_t>(1048576ULL))) {
max = cast<int64_t>(1048576ULL);
}
Slice<uint8_t> data = Slice<uint8_t>::make(max);
{auto&& tmp78 = data;
for(int64_t tmp79=0;tmp79<len(tmp78);++tmp79){
auto i=tmp79;data[i] = threedo_Machine_Read(m,cast<uint32_t>((src + cast<uint32_t>(i))));
}}
cel->PDAT = data;
if(rrcapture::trace.active&&rrcapture::trace.current)rrcapture::trace.events[rrcapture::trace.current-1].resource=rrcapture::trace.resource(data.p,data.n);
if ((cel->Coded && (plutPtr != cast<uint32_t>(0ULL)))) {
for (int64_t i = cast<int64_t>(0ULL);(i < cast<int64_t>(32ULL));i++){
cel->PLUT = append(cel->PLUT,cast<uint16_t>((shl<uint16_t>(cast<uint16_t>(threedo_Machine_Read(m,cast<uint32_t>((plutPtr + cast<uint32_t>(cast<int64_t>((i * cast<int64_t>(2ULL)))))))),cast<int64_t>(8ULL)) | cast<uint16_t>(threedo_Machine_Read(m,cast<uint32_t>((cast<uint32_t>((plutPtr + cast<uint32_t>(cast<int64_t>((i * cast<int64_t>(2ULL)))))) + cast<uint32_t>(1ULL))))))));
}
}
std::function<std::tuple<int64_t,int64_t>(int64_t,int64_t)> mapCorner = [&](int64_t c,int64_t r)->std::tuple<int64_t,int64_t>{
int64_t hdxr = cast<int64_t>((hdx + cast<int64_t>((r * hddx))));
int64_t hdyr = cast<int64_t>((hdy + cast<int64_t>((r * hddy))));
return {cast<int64_t>((cast<int64_t>((xPos + cast<int64_t>((r * vdx)))) + cast<int64_t>((c * (shr<int64_t>(hdxr,cast<int64_t>(4ULL))))))),cast<int64_t>((cast<int64_t>((yPos + cast<int64_t>((r * vdy)))) + cast<int64_t>((c * (shr<int64_t>(hdyr,cast<int64_t>(4ULL)))))))};
}
;
bool bgnd = (cast<uint32_t>((flags & threedo_ccbBGND)) != cast<uint32_t>(0ULL));
bool persp = ((((hddx != cast<int64_t>(0ULL)) || (hddy != cast<int64_t>(0ULL))) || (vdx != cast<int64_t>(0ULL))) || (hdy != cast<int64_t>(0ULL)));
int64_t written = cast<int64_t>(0ULL);
int64_t calls={};
int64_t clearN={};
int64_t offN={};
std::function<void(int64_t,int64_t,uint32_t)> put = [&](int64_t sx,int64_t sy,uint32_t v)->void{
calls++;
if(rrcapture::trace.active){auto&t=rrcapture::trace;t.u=sx;t.v=sy;t.texel=v;t.palette=plutPtr;
 t.source=cel->Packed?0:lrform?src+uint32_t((sy/2)*(cel->Width/2)*4+sx*4+(sy&1)*2):src+uint32_t(sy*((cel->Width*cel->BPP+31)/32)*4+(sx*cel->BPP)/8);}
auto tmp80 = threedo_Machine_decodePixel(m,cel,v,flags,bgnd);
uint16_t pix = std::get<0>(tmp80);
uint32_t amv = std::get<1>(tmp80);
bool transparent = std::get<2>(tmp80);
if (transparent) {
clearN++;
if(!rrcapture::trace.active)return ;
}
if ((m->PerspTint && persp)) {
auto tmp81 = std::make_tuple(cast<uint16_t>(31775ULL),cast<uint32_t>(73ULL));
pix = std::get<0>(tmp81);
amv = std::get<1>(tmp81);
}
auto tmp82 = std::make_tuple(cast<int64_t>(sx),cast<int64_t>(sy));
int64_t c = std::get<0>(tmp82);
int64_t r = std::get<1>(tmp82);
auto tmp83 = mapCorner(c,r);
int64_t x0 = std::get<0>(tmp83);
int64_t y0 = std::get<1>(tmp83);
auto tmp84 = mapCorner(cast<int64_t>((c + cast<int64_t>(1ULL))),r);
int64_t x1 = std::get<0>(tmp84);
int64_t y1 = std::get<1>(tmp84);
auto tmp85 = mapCorner(c,cast<int64_t>((r + cast<int64_t>(1ULL))));
int64_t xv = std::get<0>(tmp85);
int64_t yv = std::get<1>(tmp85);
std::function<int64_t(int64_t,int64_t)> edgeSteps = [&](int64_t dx,int64_t dy)->int64_t{
int64_t span = dx;
if ((span < cast<int64_t>(0ULL))) {
span = cast<int64_t>(-span);
}
if ((dy > span)) {
span = dy;
}
else if ((cast<int64_t>(-dy) > span)) {
span = cast<int64_t>(-dy);
}
int64_t steps = shr<int64_t>((cast<int64_t>((span + cast<int64_t>(65535ULL)))),cast<int64_t>(16ULL));
if (((steps < cast<int64_t>(1ULL)) || (steps > cast<int64_t>(12ULL)))) {
steps = cast<int64_t>(1ULL);
}
return steps;
}
;
int64_t stepsH = edgeSteps(cast<int64_t>((x1 - x0)),cast<int64_t>((y1 - y0)));
int64_t stepsV = edgeSteps(cast<int64_t>((xv - x0)),cast<int64_t>((yv - y0)));
for (int64_t t = cast<int64_t>(cast<int64_t>(0ULL));(t < stepsV);t++){
int64_t bx = cast<int64_t>((x0 + divi<int64_t>(cast<int64_t>(((cast<int64_t>((xv - x0))) * t)),stepsV)));
int64_t by = cast<int64_t>((y0 + divi<int64_t>(cast<int64_t>(((cast<int64_t>((yv - y0))) * t)),stepsV)));
for (int64_t s = cast<int64_t>(cast<int64_t>(0ULL));(s < stepsH);s++){
int64_t dx = cast<int64_t>((bx + divi<int64_t>(cast<int64_t>(((cast<int64_t>((x1 - x0))) * s)),stepsH)));
int64_t dy = cast<int64_t>((by + divi<int64_t>(cast<int64_t>(((cast<int64_t>((y1 - y0))) * s)),stepsH)));
auto tmp86 = std::make_tuple(cast<int64_t>(shr<int64_t>(dx,cast<int64_t>(16ULL))),cast<int64_t>(shr<int64_t>(dy,cast<int64_t>(16ULL))));
int64_t x = std::get<0>(tmp86);
int64_t y = std::get<1>(tmp86);
if (((((x < cast<int64_t>(0ULL)) || (y < cast<int64_t>(0ULL))) || (x >= bm.w)) || (y >= bm.h))) {
offN++;
continue;
}
if (((m->CelDebug && (x == cast<int64_t>(m->ProbeX))) && (y == cast<int64_t>(m->ProbeY)))) {
m->CelDebugLog = append(m->CelDebugLog,go_fmt_Sprintf(std::string("PROBE (%d,%d) hit by cel src=%08X %dbpp %dx%d pos=(%d,%d) HD=(%X,%X) VD=(%X,%X) lrform=%v flags=%08X",100),x,y,src,cel->BPP,cel->Width,cel->Height,cast<int64_t>(shr<int64_t>(xPos,cast<int64_t>(16ULL))),cast<int64_t>(shr<int64_t>(yPos,cast<int64_t>(16ULL))),hdx,hdy,vdx,vdy,lrform,flags));
}
if(transparent){auto&t=rrcapture::trace;uint32_t a=bm.buf+uint32_t((y/2)*bm.w*4+x*4+(y&1)*2);t.record(a,0,2,m->CPU->Instrs,m->CPU->cur,t.current,4);continue;}
if(rrcapture::trace.active)rrcapture::trace.texel=pix;
threedo_Machine_blendPixel(m,bm,x,y,pix,amv,pixc,flags);
written++;
}
}
}
;
if (((lrform && (!cel->Packed)) && (cel->BPP == cast<int64_t>(16ULL)))) {
threedo_Machine_decodeLRForm16(m,cel,src,put);
}
else if (cel->Packed) {
threedo_Cel_decodePacked(cel,put);
}
else {
threedo_Cel_decodeUnpacked(cel,put);
}
if ((m->CelDebug && (len(m->CelDebugLog) < cast<int64_t>(4000ULL)))) {
std::string kind = std::string("unpacked",8);
if (cel->Packed) {
kind = std::string("packed",6);
}
std::string coded = std::string("coded",5);
if ((!cel->Coded)) {
coded = std::string("16bpp",5);
}
std::string lr = std::string("",0);
if (lrform) {
lr = std::string(" LRFORM",7);
}
std::string persp = std::string("",0);
if (((((hddx != cast<int64_t>(0ULL)) || (hddy != cast<int64_t>(0ULL))) || (vdx != cast<int64_t>(0ULL))) || (hdy != cast<int64_t>(0ULL)))) {
persp = std::string(" PERSP",6);
}
m->CelDebugLog = append(m->CelDebugLog,go_fmt_Sprintf(std::string("cel ccb=%08X src=%08X %dbpp %s %s%s%s %dx%d pos=(%d,%d) HD=(%X,%X) VD=(%X,%X) HDD=(%X,%X) pixc=%08X flags=%08X calls=%d clear=%d off=%d wrote=%d",144),ccb,src,cel->BPP,kind,coded,lr,persp,cel->Width,cel->Height,cast<int64_t>(shr<int64_t>(xPos,cast<int64_t>(16ULL))),cast<int64_t>(shr<int64_t>(yPos,cast<int64_t>(16ULL))),hdx,hdy,vdx,vdy,hddx,hddy,pixc,flags,calls,clearN,offN,written));
}
return true;
}
}
// tools/platform/threedo/graphicsfolio.go:546:1
void threedo_Machine_decodeLRForm16(threedo_Machine* m,threedo_Cel* c,uint32_t src,std::function<void(int64_t,int64_t,uint32_t)> set){
{
int64_t cols = divi<int64_t>(c->Width,cast<int64_t>(2ULL));
for (int64_t lp = cast<int64_t>(0ULL);(lp < c->Height);lp++){
for (int64_t x = cast<int64_t>(0ULL);(x < cols);x++){
uint32_t w = cast<uint32_t>((src + cast<uint32_t>((cast<uint32_t>(cast<int64_t>((cast<int64_t>((lp * cols)) + x))) * cast<uint32_t>(4ULL)))));
set(x,cast<int64_t>((lp * cast<int64_t>(2ULL))),cast<uint32_t>((shl<uint32_t>(cast<uint32_t>(threedo_Machine_Read(m,w)),cast<int64_t>(8ULL)) | cast<uint32_t>(threedo_Machine_Read(m,cast<uint32_t>((w + cast<uint32_t>(1ULL))))))));
set(x,cast<int64_t>((cast<int64_t>((lp * cast<int64_t>(2ULL))) + cast<int64_t>(1ULL))),cast<uint32_t>((shl<uint32_t>(cast<uint32_t>(threedo_Machine_Read(m,cast<uint32_t>((w + cast<uint32_t>(2ULL))))),cast<int64_t>(8ULL)) | cast<uint32_t>(threedo_Machine_Read(m,cast<uint32_t>((w + cast<uint32_t>(3ULL))))))));
}
}
}
}
// tools/platform/threedo/graphicsfolio.go:566:1
std::tuple<uint16_t,uint32_t,bool> threedo_Machine_decodePixel(threedo_Machine* m,threedo_Cel* cel,uint32_t v,uint32_t flags,bool bgnd){
{
uint32_t amv = cast<uint32_t>(cast<uint32_t>(73ULL));
if ((!cel->Coded)) {
if (((cast<uint32_t>((v & cast<uint32_t>(32767ULL))) == cast<uint32_t>(0ULL)) && (!bgnd))) {
return {cast<uint16_t>(0ULL),amv,true};
}
return {cast<uint16_t>(v),amv,false};
}
uint32_t idx={};
uint16_t pw={};
{
switch(cel->BPP){
case cast<int64_t>(1ULL):{
idx = cast<uint32_t>((cast<uint32_t>(((cast<uint32_t>((flags & cast<uint32_t>(15ULL)))) * cast<uint32_t>(2ULL))) + (cast<uint32_t>((v & cast<uint32_t>(1ULL))))));
break;}
case cast<int64_t>(2ULL):{
idx = cast<uint32_t>((cast<uint32_t>(((cast<uint32_t>((flags & cast<uint32_t>(14ULL)))) * cast<uint32_t>(2ULL))) + (cast<uint32_t>((v & cast<uint32_t>(3ULL))))));
break;}
case cast<int64_t>(4ULL):{
idx = cast<uint32_t>((cast<uint32_t>(((cast<uint32_t>((flags & cast<uint32_t>(8ULL)))) * cast<uint32_t>(2ULL))) + (cast<uint32_t>((v & cast<uint32_t>(15ULL))))));
break;}
case cast<int64_t>(6ULL):{
idx = cast<uint32_t>((v & cast<uint32_t>(31ULL)));
pw = shl<uint16_t>(cast<uint16_t>(cast<uint32_t>(((shr<uint32_t>(v,cast<int64_t>(5ULL))) & cast<uint32_t>(1ULL)))),cast<int64_t>(15ULL));
break;}
case cast<int64_t>(8ULL):{
idx = cast<uint32_t>((v & cast<uint32_t>(31ULL)));
amv = cast<uint32_t>(((cast<uint32_t>((shl<uint32_t>((cast<uint32_t>(((shr<uint32_t>(v,cast<int64_t>(6ULL))) & cast<uint32_t>(3ULL)))),cast<int64_t>(1ULL)) | cast<uint32_t>(((shr<uint32_t>(v,cast<int64_t>(5ULL))) & cast<uint32_t>(1ULL)))))) * cast<uint32_t>(73ULL)));
break;}
default:{
idx = cast<uint32_t>((v & cast<uint32_t>(31ULL)));
break;}
}}
if ((cast<int64_t>(idx) >= len(cel->PLUT))) {
return {cast<uint16_t>(0ULL),amv,true};
}
if(rrcapture::trace.active)rrcapture::trace.palette+=idx*2;
uint16_t raw = cel->PLUT[idx];
uint16_t color = cast<uint16_t>((raw & cast<uint16_t>(32767ULL)));
if (((color == cast<uint16_t>(0ULL)) && (!bgnd))) {
return {cast<uint16_t>(0ULL),amv,true};
}
if ((cel->BPP != cast<int64_t>(6ULL))) {
pw = cast<uint16_t>((raw & cast<uint16_t>(32768ULL)));
}
return {cast<uint16_t>((color | pw)),amv,false};
}
}
// tools/platform/threedo/graphicsfolio.go:621:1
void threedo_Machine_blendPixel(threedo_Machine* m,threedo_gfxBitmap bm,int64_t x,int64_t y,uint16_t pix,uint32_t amv,uint32_t pixc,uint32_t flags){
{
uint32_t a = cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((bm.buf + cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(shr<int64_t>(y,cast<int64_t>(1ULL))) * cast<uint32_t>(bm.w))) * cast<uint32_t>(4ULL))))) + cast<uint32_t>((cast<uint32_t>(x) * cast<uint32_t>(4ULL))))) + cast<uint32_t>((cast<uint32_t>(cast<int64_t>((y & cast<int64_t>(1ULL)))) * cast<uint32_t>(2ULL)))));
uint32_t word = cast<uint32_t>((pixc & cast<uint32_t>(65535ULL)));
{
switch(cast<uint32_t>((flags & threedo_ccbPOVER))){
case cast<uint32_t>(256ULL):{
break;}
case cast<uint32_t>(384ULL):{
word = shr<uint32_t>(pixc,cast<int64_t>(16ULL));
break;}
default:{
if ((cast<uint16_t>((pix & cast<uint16_t>(32768ULL))) != cast<uint16_t>(0ULL))) {
word = shr<uint32_t>(pixc,cast<int64_t>(16ULL));
}
break;}
}}
bool s1 = (cast<uint32_t>((word & cast<uint32_t>(32768ULL))) != cast<uint32_t>(0ULL));
uint32_t ms = cast<uint32_t>(((shr<uint32_t>(word,cast<int64_t>(13ULL))) & cast<uint32_t>(3ULL)));
uint32_t mxf = cast<uint32_t>(((shr<uint32_t>(word,cast<int64_t>(10ULL))) & cast<uint32_t>(7ULL)));
uint32_t dv1 = cast<uint32_t>(((shr<uint32_t>(word,cast<int64_t>(8ULL))) & cast<uint32_t>(3ULL)));
uint32_t s2 = cast<uint32_t>(((shr<uint32_t>(word,cast<int64_t>(6ULL))) & cast<uint32_t>(3ULL)));
uint32_t avf = cast<uint32_t>(((shr<uint32_t>(word,cast<int64_t>(1ULL))) & cast<uint32_t>(31ULL)));
uint32_t dv2 = cast<uint32_t>((word & cast<uint32_t>(1ULL)));
if (((((((!s1) && (ms == cast<uint32_t>(0ULL))) && (s2 == cast<uint32_t>(0ULL))) && (dv2 == cast<uint32_t>(0ULL))) && (mxf == cast<uint32_t>(7ULL))) && (dv1 == cast<uint32_t>(3ULL)))) {
threedo_Machine_Write(m,a,cast<uint8_t>(shr<uint16_t>(pix,cast<int64_t>(8ULL))));
threedo_Machine_Write(m,cast<uint32_t>((a + cast<uint32_t>(1ULL))),cast<uint8_t>(pix));
threedo_Machine_celPixel(m,bm,x,y,pix);
return ;
}
uint32_t dv3={};
bool clip = true;
if ((cast<uint32_t>((flags & cast<uint32_t>(1024ULL))) != cast<uint32_t>(0ULL))) {
dv3 = cast<uint32_t>(((shr<uint32_t>(avf,cast<int64_t>(3ULL))) & cast<uint32_t>(3ULL)));
clip = (cast<uint32_t>((avf & cast<uint32_t>(4ULL))) == cast<uint32_t>(0ULL));
}
uint16_t dst = cast<uint16_t>((shl<uint16_t>(cast<uint16_t>(threedo_Machine_Read(m,a)),cast<int64_t>(8ULL)) | cast<uint16_t>(threedo_Machine_Read(m,cast<uint32_t>((a + cast<uint32_t>(1ULL)))))));
auto tmp87 = threedo_chan5(pix);
uint32_t ir = std::get<0>(tmp87);
uint32_t ig = std::get<1>(tmp87);
uint32_t ib = std::get<2>(tmp87);
if (s1) {
auto tmp88 = threedo_chan5(dst);
ir = std::get<0>(tmp88);
ig = std::get<1>(tmp88);
ib = std::get<2>(tmp88);
}
uint32_t sh1 = threedo_pdv(dv1);
uint32_t c1r={};
uint32_t c1g={};
uint32_t c1b={};
{
switch(ms){
case cast<uint32_t>(1ULL):{
c1r = shr<uint32_t>((cast<uint32_t>((ir * (cast<uint32_t>(((cast<uint32_t>(((shr<uint32_t>(amv,cast<int64_t>(6ULL))) & cast<uint32_t>(7ULL)))) + cast<uint32_t>(1ULL))))))),sh1);
c1g = shr<uint32_t>((cast<uint32_t>((ig * (cast<uint32_t>(((cast<uint32_t>(((shr<uint32_t>(amv,cast<int64_t>(3ULL))) & cast<uint32_t>(7ULL)))) + cast<uint32_t>(1ULL))))))),sh1);
c1b = shr<uint32_t>((cast<uint32_t>((ib * (cast<uint32_t>(((cast<uint32_t>((amv & cast<uint32_t>(7ULL)))) + cast<uint32_t>(1ULL))))))),sh1);
break;}
case cast<uint32_t>(2ULL):{
c1r = shr<uint32_t>((cast<uint32_t>((ir * (cast<uint32_t>(((shr<uint32_t>(ir,cast<int64_t>(2ULL))) + cast<uint32_t>(1ULL))))))),threedo_pdv(cast<uint32_t>((ir & cast<uint32_t>(3ULL)))));
c1g = shr<uint32_t>((cast<uint32_t>((ig * (cast<uint32_t>(((shr<uint32_t>(ig,cast<int64_t>(2ULL))) + cast<uint32_t>(1ULL))))))),threedo_pdv(cast<uint32_t>((ig & cast<uint32_t>(3ULL)))));
c1b = shr<uint32_t>((cast<uint32_t>((ib * (cast<uint32_t>(((shr<uint32_t>(ib,cast<int64_t>(2ULL))) + cast<uint32_t>(1ULL))))))),threedo_pdv(cast<uint32_t>((ib & cast<uint32_t>(3ULL)))));
break;}
case cast<uint32_t>(3ULL):{
c1r = shr<uint32_t>((cast<uint32_t>((ir * (cast<uint32_t>(((shr<uint32_t>(ir,cast<int64_t>(2ULL))) + cast<uint32_t>(1ULL))))))),sh1);
c1g = shr<uint32_t>((cast<uint32_t>((ig * (cast<uint32_t>(((shr<uint32_t>(ig,cast<int64_t>(2ULL))) + cast<uint32_t>(1ULL))))))),sh1);
c1b = shr<uint32_t>((cast<uint32_t>((ib * (cast<uint32_t>(((shr<uint32_t>(ib,cast<int64_t>(2ULL))) + cast<uint32_t>(1ULL))))))),sh1);
break;}
default:{
c1r = shr<uint32_t>((cast<uint32_t>((ir * (cast<uint32_t>((mxf + cast<uint32_t>(1ULL))))))),sh1);
c1g = shr<uint32_t>((cast<uint32_t>((ig * (cast<uint32_t>((mxf + cast<uint32_t>(1ULL))))))),sh1);
c1b = shr<uint32_t>((cast<uint32_t>((ib * (cast<uint32_t>((mxf + cast<uint32_t>(1ULL))))))),sh1);
break;}
}}
uint32_t c2r={};
uint32_t c2g={};
uint32_t c2b={};
{
switch(s2){
case cast<uint32_t>(1ULL):{
auto tmp89 = std::make_tuple(shr<uint32_t>(avf,dv3),shr<uint32_t>(avf,dv3),shr<uint32_t>(avf,dv3));
c2r = std::get<0>(tmp89);
c2g = std::get<1>(tmp89);
c2b = std::get<2>(tmp89);
break;}
case cast<uint32_t>(2ULL):{
auto tmp90 = threedo_chan5(dst);
uint32_t dr = std::get<0>(tmp90);
uint32_t dg = std::get<1>(tmp90);
uint32_t db = std::get<2>(tmp90);
auto tmp91 = std::make_tuple(shr<uint32_t>(dr,dv3),shr<uint32_t>(dg,dv3),shr<uint32_t>(db,dv3));
c2r = std::get<0>(tmp91);
c2g = std::get<1>(tmp91);
c2b = std::get<2>(tmp91);
break;}
case cast<uint32_t>(3ULL):{
auto tmp92 = threedo_chan5(pix);
uint32_t sr = std::get<0>(tmp92);
uint32_t sg = std::get<1>(tmp92);
uint32_t sb = std::get<2>(tmp92);
auto tmp93 = std::make_tuple(shr<uint32_t>(sr,dv3),shr<uint32_t>(sg,dv3),shr<uint32_t>(sb,dv3));
c2r = std::get<0>(tmp93);
c2g = std::get<1>(tmp93);
c2b = std::get<2>(tmp93);
break;}
}}
std::function<uint32_t(uint32_t)> clampOr = [&](uint32_t v)->uint32_t{
if ((!clip)) {
return cast<uint32_t>((v & cast<uint32_t>(31ULL)));
}
if ((v > cast<uint32_t>(31ULL))) {
return cast<uint32_t>(31ULL);
}
return v;
}
;
uint32_t out = cast<uint32_t>((cast<uint32_t>((shl<uint32_t>(clampOr(shr<uint32_t>((cast<uint32_t>((c1r + c2r))),dv2)),cast<int64_t>(10ULL)) | shl<uint32_t>(clampOr(shr<uint32_t>((cast<uint32_t>((c1g + c2g))),dv2)),cast<int64_t>(5ULL)))) | clampOr(shr<uint32_t>((cast<uint32_t>((c1b + c2b))),dv2))));
threedo_Machine_Write(m,a,cast<uint8_t>(shr<uint32_t>(out,cast<int64_t>(8ULL))));
threedo_Machine_Write(m,cast<uint32_t>((a + cast<uint32_t>(1ULL))),cast<uint8_t>(out));
threedo_Machine_celPixel(m,bm,x,y,cast<uint16_t>(out));
}
}
// tools/platform/threedo/graphicsfolio.go:716:1
uint32_t threedo_pdv(uint32_t n){
{
return (cast<uint32_t>((cast<uint32_t>(((cast<uint32_t>((n - cast<uint32_t>(1ULL)))) & cast<uint32_t>(3ULL))) + cast<uint32_t>(1ULL))));
}
}
// tools/platform/threedo/graphicsfolio.go:719:1
std::tuple<uint32_t,uint32_t,uint32_t> threedo_chan5(uint16_t p){
uint32_t r{};
uint32_t g{};
uint32_t b{};
{
return {cast<uint32_t>((cast<uint32_t>(shr<uint16_t>(p,cast<int64_t>(10ULL))) & cast<uint32_t>(31ULL))),cast<uint32_t>((cast<uint32_t>(shr<uint16_t>(p,cast<int64_t>(5ULL))) & cast<uint32_t>(31ULL))),cast<uint32_t>((cast<uint32_t>(p) & cast<uint32_t>(31ULL)))};
}
}
// tools/platform/threedo/io.go:78:1
uint32_t threedo_Machine_read32(threedo_Machine* m,uint32_t a){
{
return cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((shl<uint32_t>(cast<uint32_t>(threedo_Machine_Read(m,a)),cast<int64_t>(24ULL)) | shl<uint32_t>(cast<uint32_t>(threedo_Machine_Read(m,cast<uint32_t>((a + cast<uint32_t>(1ULL))))),cast<int64_t>(16ULL)))) | shl<uint32_t>(cast<uint32_t>(threedo_Machine_Read(m,cast<uint32_t>((a + cast<uint32_t>(2ULL))))),cast<int64_t>(8ULL)))) | cast<uint32_t>(threedo_Machine_Read(m,cast<uint32_t>((a + cast<uint32_t>(3ULL)))))));
}
}
// tools/platform/threedo/io.go:82:1
void threedo_Machine_write32(threedo_Machine* m,uint32_t a,uint32_t v){
{
threedo_Machine_Write(m,a,cast<uint8_t>(shr<uint32_t>(v,cast<int64_t>(24ULL))));
threedo_Machine_Write(m,cast<uint32_t>((a + cast<uint32_t>(1ULL))),cast<uint8_t>(shr<uint32_t>(v,cast<int64_t>(16ULL))));
threedo_Machine_Write(m,cast<uint32_t>((a + cast<uint32_t>(2ULL))),cast<uint8_t>(shr<uint32_t>(v,cast<int64_t>(8ULL))));
threedo_Machine_Write(m,cast<uint32_t>((a + cast<uint32_t>(3ULL))),cast<uint8_t>(v));
}
}
// tools/platform/threedo/io.go:90:1
std::string threedo_Machine_readCStr(threedo_Machine* m,uint32_t a){
{
if ((a == cast<uint32_t>(0ULL))) {
return std::string("",0);
}
Slice<uint8_t> b={};
for (int64_t i = cast<int64_t>(0ULL);(i < cast<int64_t>(256ULL));i++){
uint8_t c = threedo_Machine_Read(m,cast<uint32_t>((a + cast<uint32_t>(i))));
if ((c == cast<uint8_t>(0ULL))) {
break;
}
b = append(b,c);
}
return cast<std::string>(b);
}
}
// tools/platform/threedo/io.go:107:1
uint32_t threedo_Machine_tagArg(threedo_Machine* m,uint32_t p,uint32_t want){
{
for (int64_t i = cast<int64_t>(0ULL);((i < cast<int64_t>(64ULL)) && (p != cast<uint32_t>(0ULL)));i++){
uint32_t tag = threedo_Machine_read32(m,p);
uint32_t val = threedo_Machine_read32(m,cast<uint32_t>((p + cast<uint32_t>(4ULL))));
if ((tag == cast<uint32_t>(0ULL))) {
break;
}
if ((cast<uint32_t>((tag & cast<uint32_t>(65535ULL))) == want)) {
return val;
}
p += cast<uint32_t>(8ULL);
}
return cast<uint32_t>(0ULL);
}
}
// tools/platform/threedo/io.go:123:1
std::string threedo_Machine_tagString(threedo_Machine* m,uint32_t p,uint32_t want){
{
return threedo_Machine_readCStr(m,threedo_Machine_tagArg(m,p,want));
}
}
// tools/platform/threedo/io.go:136:1
void threedo_Machine_serviceIO(threedo_Machine* m,arm60_CPU* c,bool async){
{
int32_t ioNum = cast<int32_t>(arm60_CPU_Reg(c,cast<uint32_t>(0ULL)));
uint32_t info = arm60_CPU_Reg(c,cast<uint32_t>(1ULL));
threedo_item* it = get(m->items,ioNum);
uint32_t cmd = cast<uint32_t>(threedo_Machine_Read(m,cast<uint32_t>((info + threedo_ioiCommand))));
uint32_t offset = threedo_Machine_read32(m,cast<uint32_t>((info + threedo_ioiOffset)));
uint32_t sendBuf = threedo_Machine_read32(m,cast<uint32_t>((info + threedo_ioiSendBuf)));
uint32_t sendLen = threedo_Machine_read32(m,cast<uint32_t>((info + threedo_ioiSendLen)));
uint32_t recvBuf = threedo_Machine_read32(m,cast<uint32_t>((info + threedo_ioiRecvBuf)));
uint32_t recvLen = threedo_Machine_read32(m,cast<uint32_t>((info + threedo_ioiRecvLen)));
if ((m->SportDebug && (cmd == threedo_sportFlashWrite))) {
threedo_Machine_note(m,go_fmt_Sprintf(std::string("FLASHWRITE dest=0x%08X bytes=0x%X val=0x%08X (send=%08X,%08X)",61),recvBuf,recvLen,offset,sendBuf,sendLen));
}
if ((bool(it) && (it->addr != cast<uint32_t>(0ULL)))) {
for (uint32_t i = cast<uint32_t>(cast<uint32_t>(0ULL));(i < threedo_ioInfoBytes);i++){
threedo_Machine_Write(m,cast<uint32_t>((cast<uint32_t>((it->addr + threedo_ioInfoOff)) + i)),threedo_Machine_Read(m,cast<uint32_t>((info + i))));
}
threedo_Machine_write32(m,cast<uint32_t>((it->addr + threedo_ioFlagsOff)),(threedo_Machine_read32(m,cast<uint32_t>((it->addr + threedo_ioFlagsOff))) & ~(cast<uint32_t>(cast<uint32_t>((threedo_ioDone | threedo_ioQuick))))));
}
if ((((async && m->PaceFields) && (threedo_Machine_deviceName(m,it) == std::string("timer",5))) && (cmd == threedo_timerCmdWaitField))) {
threedo_Machine_startFieldWait(m,it,offset);
arm60_CPU_SetReg(c,cast<uint32_t>(0ULL),cast<uint32_t>(0ULL));
return ;
}
auto tmp94 = threedo_Machine_performIO(m,it,cmd,offset,sendBuf,sendLen,recvBuf,recvLen);
int32_t actual = std::get<0>(tmp94);
int32_t ioErr = std::get<1>(tmp94);
if ((bool(it) && (it->addr != cast<uint32_t>(0ULL)))) {
threedo_Machine_write32(m,cast<uint32_t>((it->addr + threedo_ioActualOff)),cast<uint32_t>(actual));
threedo_Machine_write32(m,cast<uint32_t>((it->addr + threedo_ioFlagsOff)),cast<uint32_t>((cast<uint32_t>((threedo_Machine_read32(m,cast<uint32_t>((it->addr + threedo_ioFlagsOff))) | threedo_ioDone)) | threedo_ioQuick)));
threedo_Machine_write32(m,cast<uint32_t>((it->addr + threedo_ioErrorOff)),cast<uint32_t>(ioErr));
}
threedo_Machine_completeIO(m,it);
arm60_CPU_SetReg(c,cast<uint32_t>(0ULL),cast<uint32_t>(ioErr));
}
}
// tools/platform/threedo/io.go:205:1
void threedo_Machine_completeIO(threedo_Machine* m,threedo_item* it){
{
if ((!it)) {
return ;
}
if ((it->replyPort != cast<int32_t>(0ULL))) {
{
threedo_item* p = get(m->items,it->replyPort);
if (bool(p)) {
if ((it->addr != cast<uint32_t>(0ULL))) {
threedo_Machine_write32(m,cast<uint32_t>((it->addr + threedo_msgDataPtr)),cast<uint32_t>(it->num));
}
p->msgs = append(p->msgs,it->num);
if (bool(m->OnMsgQueue)) {
m->OnMsgQueue(m,p->num,it->num,std::string("CompleteIO",10));
}
threedo_Machine_yieldTo(m,threedo_Machine_sendSignal(m,p->owner,p->signal),p->owner);
return ;
}
}
}
threedo_Machine_sendSignal(m,it->owner,threedo_sigfIODONE);
}
}
// tools/platform/threedo/io.go:230:1
std::string threedo_Machine_deviceName(threedo_Machine* m,threedo_item* it){
{
if ((bool(it) && (it->device != cast<int32_t>(0ULL)))) {
{
threedo_item* d = get(m->items,it->device);
if (bool(d)) {
return d->name;
}
}
}
return std::string("",0);
}
}
// tools/platform/threedo/io.go:246:1
void threedo_Machine_completeFieldWait(threedo_Machine* m,threedo_item* it,int32_t submitter){
{
if ((it->replyPort != cast<int32_t>(0ULL))) {
{
threedo_item* p = get(m->items,it->replyPort);
if (bool(p)) {
threedo_Machine_sendSignal(m,p->owner,p->signal);
return ;
}
}
}
threedo_Machine_sendSignal(m,submitter,threedo_sigfIODONE);
}
}
// tools/platform/threedo/io.go:257:1
uint32_t threedo_Machine_ioError(threedo_Machine* m,int32_t ioNum){
{
{
threedo_item* it = get(m->items,ioNum);
if ((bool(it) && (it->addr != cast<uint32_t>(0ULL)))) {
return threedo_Machine_read32(m,cast<uint32_t>((it->addr + threedo_ioErrorOff)));
}
}
return cast<uint32_t>(0ULL);
}
}
// tools/platform/threedo/io.go:269:1
std::tuple<int32_t,int32_t> threedo_Machine_performIO(threedo_Machine* m,threedo_item* it,uint32_t cmd,uint32_t offset,uint32_t sendBuf,uint32_t sendLen,uint32_t buf,uint32_t length){
{
std::string dev = threedo_Machine_deviceName(m,it);
threedo_Machine_note(m,go_fmt_Sprintf(std::string("IO dev=%q cmd=%d offset=0x%X buf=0x%08X len=0x%X",48),dev,cmd,offset,buf,length));
if (((dev == std::string("timer",5)) && (cmd == threedo_timerCmdWaitField))) {
threedo_Machine_advanceVBlank(m,offset);
if ((!m->PaceFields)) {
threedo_Machine_curTask(m)->state = threedo_stReady;
m->needSchedule = true;
}
return {cast<int32_t>(0ULL),cast<int32_t>(0ULL)};
}
if (((((cmd == threedo_sportFlashWrite) && (buf >= threedo_vramBase)) && (buf < cast<uint32_t>((threedo_vramBase + threedo_vramSize)))) && (length > cast<uint32_t>(0ULL)))) {
threedo_Machine_flashClearRange(m,buf,length,cast<uint16_t>(offset));
return {cast<int32_t>(0ULL),cast<int32_t>(0ULL)};
}
if ((dev == std::string("SPORT",5))) {
return {cast<int32_t>(0ULL),cast<int32_t>(0ULL)};
}
if (((dev != std::string("",0)) && (dev != std::string("timer",5)))) {
return threedo_Machine_fileDeviceIO(m,dev,cmd,offset,sendBuf,sendLen,buf,length);
}
return {cast<int32_t>(0ULL),cast<int32_t>(0ULL)};
}
}
// tools/platform/threedo/io.go:315:1
std::tuple<int32_t,int32_t> threedo_Machine_fileDeviceIOLegacy(threedo_Machine* m,std::string name,uint32_t cmd,uint32_t offset,uint32_t sendBuf,uint32_t sendLen,uint32_t buf,uint32_t length){
{
auto tmp95 = threedo_Machine_fileData(m,name);
Slice<uint8_t> data = std::get<0>(tmp95);
uint32_t blockSize = std::get<1>(tmp95);
std::string nvKey = std::get<2>(tmp95);
bool ok = std::get<3>(tmp95);
if ((!ok)) {
return {cast<int32_t>(0ULL),cast<int32_t>(-cast<int64_t>(1ULL))};
}
{
switch(cmd){
case threedo_cmdStatus:{
uint32_t blocks = cast<uint32_t>(len(data));
if ((blockSize > cast<uint32_t>(1ULL))) {
blocks = divi<uint32_t>((cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(len(data)) + blockSize)) - cast<uint32_t>(1ULL)))),blockSize);
}
{auto&& tmp96 = Slice<uint32_t>{cast<uint32_t>(0ULL),cast<uint32_t>(40ULL),blockSize,blocks,cast<uint32_t>(0ULL),cast<uint32_t>(0ULL),cast<uint32_t>(0ULL),cast<uint32_t>(0ULL),cast<uint32_t>(0ULL),cast<uint32_t>(len(data))};
for(int64_t tmp97=0;tmp97<len(tmp96);++tmp97){
auto i=tmp97;auto w=tmp96[tmp97];if ((cast<uint32_t>((cast<uint32_t>(i) * cast<uint32_t>(4ULL))) < length)) {
threedo_Machine_write32(m,cast<uint32_t>((buf + cast<uint32_t>((cast<uint32_t>(i) * cast<uint32_t>(4ULL))))),w);
}
}}
int32_t n = cast<int32_t>(cast<int32_t>(40ULL));
if ((cast<uint32_t>(n) > length)) {
n = cast<int32_t>(length);
}
return {n,cast<int32_t>(0ULL)};
break;}
case threedo_cmdRead:{
uint32_t byteOff = cast<uint32_t>((offset * blockSize));
if (((buf == cast<uint32_t>(0ULL)) || (cast<int64_t>(byteOff) >= len(data)))) {
return {cast<int32_t>(0ULL),cast<int32_t>(0ULL)};
}
int64_t n = cast<int64_t>((len(data) - cast<int64_t>(byteOff)));
if ((cast<uint32_t>(n) > length)) {
n = cast<int64_t>(length);
}
for (int64_t i = cast<int64_t>(0ULL);(i < n);i++){
threedo_Machine_Write(m,cast<uint32_t>((buf + cast<uint32_t>(i))),data[cast<int64_t>((cast<int64_t>(byteOff) + i))]);
}
return {cast<int32_t>(n),cast<int32_t>(0ULL)};
break;}
case threedo_cmdWrite:{
if ((nvKey == std::string("",0))) {
return {cast<int32_t>(0ULL),cast<int32_t>(-cast<int64_t>(1ULL))};
}
uint32_t byteOff = cast<uint32_t>((offset * blockSize));
Slice<uint8_t> stored = get(m->nvram,nvKey);
{
int64_t need = cast<int64_t>((cast<int64_t>(byteOff) + cast<int64_t>(sendLen)));
if ((need > len(stored))) {
Slice<uint8_t> grown = Slice<uint8_t>::make(need);
gcopy(grown,stored);
stored = grown;
}
}
for (uint32_t i = cast<uint32_t>(cast<uint32_t>(0ULL));(i < sendLen);i++){
stored[cast<uint32_t>((byteOff + i))] = threedo_Machine_Read(m,cast<uint32_t>((sendBuf + i)));
}
m->nvram[nvKey] = stored;
return {cast<int32_t>(sendLen),cast<int32_t>(0ULL)};
break;}
case threedo_fileCmdAllocBlocks:{
if ((nvKey == std::string("",0))) {
return {cast<int32_t>(0ULL),cast<int32_t>(-cast<int64_t>(1ULL))};
}
{
int64_t grow = cast<int64_t>((cast<int64_t>(offset) * cast<int64_t>(blockSize)));
if ((grow > cast<int64_t>(0ULL))) {
m->nvram[nvKey] = append(get(m->nvram,nvKey),Slice<uint8_t>::make(grow));
}
}
return {cast<int32_t>(0ULL),cast<int32_t>(0ULL)};
break;}
case threedo_fileCmdSetEOF:{
if ((nvKey == std::string("",0))) {
return {cast<int32_t>(0ULL),cast<int32_t>(-cast<int64_t>(1ULL))};
}
Slice<uint8_t> stored = get(m->nvram,nvKey);
if ((cast<int64_t>(offset) <= len(stored))) {
m->nvram[nvKey] = sub(stored,0,offset);
}
else {
m->nvram[nvKey] = append(stored,Slice<uint8_t>::make(cast<int64_t>((cast<int64_t>(offset) - len(stored)))));
}
return {cast<int32_t>(0ULL),cast<int32_t>(0ULL)};
break;}
default:{
return {cast<int32_t>(0ULL),cast<int32_t>(0ULL)};
break;}
}}
}
}
// tools/platform/threedo/io.go:410:1
void threedo_Machine_serviceMsg(threedo_Machine* m,arm60_CPU* c,uint32_t swi){
{
{
switch(swi){
case threedo_swiPutMsg:{
auto tmp98 = std::make_tuple(get(m->items,cast<int32_t>(arm60_CPU_Reg(c,cast<uint32_t>(0ULL)))),get(m->items,cast<int32_t>(arm60_CPU_Reg(c,cast<uint32_t>(1ULL)))));
threedo_item* port = std::get<0>(tmp98);
threedo_item* msg = std::get<1>(tmp98);
if (((!port) || (!msg))) {
arm60_CPU_SetReg(c,cast<uint32_t>(0ULL),cast<uint32_t>(~cast<uint32_t>(cast<uint32_t>(0ULL))));
return ;
}
if ((msg->addr != cast<uint32_t>(0ULL))) {
threedo_Machine_write32(m,cast<uint32_t>((msg->addr + threedo_msgDataPtr)),arm60_CPU_Reg(c,cast<uint32_t>(2ULL)));
threedo_Machine_write32(m,cast<uint32_t>((msg->addr + threedo_msgDataSize)),arm60_CPU_Reg(c,cast<uint32_t>(3ULL)));
threedo_Machine_write32(m,cast<uint32_t>((msg->addr + threedo_msgMsgPort)),cast<uint32_t>(port->num));
}
if ((port->name == std::string("eventbroker",11))) {
threedo_Machine_eventBrokerRequest(m,port,msg,arm60_CPU_Reg(c,cast<uint32_t>(2ULL)),arm60_CPU_Reg(c,cast<uint32_t>(3ULL)));
arm60_CPU_SetReg(c,cast<uint32_t>(0ULL),cast<uint32_t>(0ULL));
return ;
}
port->msgs = append(port->msgs,msg->num);
if (bool(m->OnMsgQueue)) {
m->OnMsgQueue(m,port->num,msg->num,std::string("PutMsg",6));
}
threedo_Machine_yieldTo(m,threedo_Machine_sendSignal(m,port->owner,port->signal),port->owner);
arm60_CPU_SetReg(c,cast<uint32_t>(0ULL),cast<uint32_t>(0ULL));
break;}
case threedo_swiGetMsg:{
threedo_item* port = get(m->items,cast<int32_t>(arm60_CPU_Reg(c,cast<uint32_t>(0ULL))));
if (((!port) || (len(port->msgs) == cast<int64_t>(0ULL)))) {
arm60_CPU_SetReg(c,cast<uint32_t>(0ULL),cast<uint32_t>(0ULL));
return ;
}
int32_t num = port->msgs[cast<int64_t>(0ULL)];
port->msgs = sub(port->msgs,cast<int64_t>(1ULL),len(port->msgs));
arm60_CPU_SetReg(c,cast<uint32_t>(0ULL),cast<uint32_t>(num));
break;}
case threedo_swiGetThisMsg:{
threedo_item* msg = get(m->items,cast<int32_t>(arm60_CPU_Reg(c,cast<uint32_t>(0ULL))));
if ((!msg)) {
arm60_CPU_SetReg(c,cast<uint32_t>(0ULL),cast<uint32_t>(0ULL));
return ;
}
{auto&& tmp99 = m->items;
for(auto [tmp100,tmp101]:tmp99){
auto p=tmp101;{auto&& tmp102 = p->msgs;
for(int64_t tmp103=0;tmp103<len(tmp102);++tmp103){
auto i=tmp103;auto qn=tmp102[tmp103];if ((qn == msg->num)) {
p->msgs = append(sub(p->msgs,0,i),sub(p->msgs,cast<int64_t>((i + cast<int64_t>(1ULL))),len(p->msgs)));
arm60_CPU_SetReg(c,cast<uint32_t>(0ULL),cast<uint32_t>(msg->num));
return ;
}
}}
}}
arm60_CPU_SetReg(c,cast<uint32_t>(0ULL),cast<uint32_t>(msg->num));
break;}
case threedo_swiReplyMsg:{
threedo_item* msg = get(m->items,cast<int32_t>(arm60_CPU_Reg(c,cast<uint32_t>(0ULL))));
if ((!msg)) {
arm60_CPU_SetReg(c,cast<uint32_t>(0ULL),cast<uint32_t>(~cast<uint32_t>(cast<uint32_t>(0ULL))));
return ;
}
if ((msg->addr != cast<uint32_t>(0ULL))) {
threedo_Machine_write32(m,cast<uint32_t>((msg->addr + threedo_msgDataPtr)),arm60_CPU_Reg(c,cast<uint32_t>(2ULL)));
threedo_Machine_write32(m,cast<uint32_t>((msg->addr + threedo_msgDataSize)),arm60_CPU_Reg(c,cast<uint32_t>(3ULL)));
}
threedo_Machine_replyMsg(m,msg,arm60_CPU_Reg(c,cast<uint32_t>(1ULL)));
arm60_CPU_SetReg(c,cast<uint32_t>(0ULL),cast<uint32_t>(0ULL));
break;}
}}
}
}
// tools/platform/threedo/io.go:482:1
void threedo_Machine_replyMsg(threedo_Machine* m,threedo_item* msg,uint32_t result){
{
if (((!msg) || (msg->addr == cast<uint32_t>(0ULL)))) {
return ;
}
threedo_Machine_write32(m,cast<uint32_t>((msg->addr + threedo_msgResult)),result);
threedo_item* rp = get(m->items,cast<int32_t>(threedo_Machine_read32(m,cast<uint32_t>((msg->addr + threedo_msgReplyPort)))));
if ((!rp)) {
threedo_Machine_note(m,go_fmt_Sprintf(std::string("replyMsg: msg %d has no live reply port (field=0x%X)",52),msg->num,threedo_Machine_read32(m,cast<uint32_t>((msg->addr + threedo_msgReplyPort)))));
return ;
}
if ((rp->name == std::string("eventbroker",11))) {
return ;
}
rp->msgs = append(rp->msgs,msg->num);
if (bool(m->OnMsgQueue)) {
m->OnMsgQueue(m,rp->num,msg->num,std::string("ReplyMsg",8));
}
threedo_Machine_yieldTo(m,threedo_Machine_sendSignal(m,rp->owner,rp->signal),rp->owner);
}
}
// tools/platform/threedo/io.go:511:1
void threedo_Machine_waitPort(threedo_Machine* m){
{
arm60_CPU* c = m->CPU;
threedo_task* t = threedo_Machine_curTask(m);
threedo_item* port = get(m->items,cast<int32_t>(arm60_CPU_Reg(c,cast<uint32_t>(0ULL))));
if ((!port)) {
t->folioWait = false;
threedo_Machine_SetResultAndReturn(m,cast<uint32_t>(~cast<uint32_t>(cast<uint32_t>(0ULL))));
return ;
}
int32_t wantMsg = cast<int32_t>(arm60_CPU_Reg(c,cast<uint32_t>(1ULL)));
{auto&& tmp104 = port->msgs;
for(int64_t tmp105=0;tmp105<len(tmp104);++tmp105){
auto i=tmp105;auto mn=tmp104[tmp105];if (((wantMsg == cast<int32_t>(0ULL)) || (mn == wantMsg))) {
port->msgs = append(sub(port->msgs,0,i),sub(port->msgs,cast<int64_t>((i + cast<int64_t>(1ULL))),len(port->msgs)));
t->folioWait = false;
threedo_Machine_SetResultAndReturn(m,cast<uint32_t>(mn));
return ;
}
}}
t->wait = port->signal;
t->state = threedo_stWaiting;
t->folioWait = true;
m->needSchedule = true;
}
}
// tools/platform/threedo/io.go:584:1
std::tuple<uint32_t,bool> threedo_PadButton(std::string name){
{
auto tmp106 = lookup(threedo_padButtonBits,go_strings_ToLower(go_strings_TrimSpace(name)));
uint32_t b = std::get<0>(tmp106);
bool ok = std::get<1>(tmp106);
return {b,ok};
}
}
// tools/platform/threedo/io.go:591:1
Slice<std::string> threedo_PadButtonNames(){
{
Map<uint32_t,std::string> seen = Map<uint32_t,std::string>{};
{auto&& tmp107 = threedo_padButtonBits;
for(auto [tmp108,tmp109]:tmp107){
auto n=tmp108;auto b=tmp109;if ((n == std::string("p",1))) {
continue;
}
{
auto tmp110 = lookup(seen,b);
bool ok = std::get<1>(tmp110);
if ((!ok)) {
seen[b] = n;
}
}
}}
Slice<std::string> names = Slice<std::string>::make(cast<int64_t>(0ULL),len(seen));
{auto&& tmp111 = seen;
for(auto [tmp112,tmp113]:tmp111){
auto n=tmp113;names = append(names,n);
}}
go_sort_Strings(names);
return names;
}
}
// tools/platform/threedo/io.go:613:1
void threedo_Machine_eventBrokerRequest(threedo_Machine* m,threedo_item* port,threedo_item* msg,uint32_t dataPtr,uint32_t dataSize){
{
uint32_t flavor = threedo_Machine_read32(m,dataPtr);
{
switch(flavor){
case threedo_ebConfigure:{
int32_t listener = cast<int32_t>(threedo_Machine_read32(m,cast<uint32_t>((msg->addr + threedo_msgReplyPort))));
uint32_t category = threedo_Machine_read32(m,cast<uint32_t>((dataPtr + cast<uint32_t>(4ULL))));
if (bool(get(m->items,listener))) {
m->ebListeners = append(m->ebListeners,listener);
}
threedo_Machine_note(m,go_fmt_Sprintf(std::string("eventbroker: task #%d configured (category %d) -> listener port %d",66),msg->owner,category,listener));
break;}
case threedo_ebDescribePods:{
if (((dataSize == cast<uint32_t>(0ULL)) || (dataSize >= cast<uint32_t>(48ULL)))) {
threedo_Machine_writeWord(m,cast<uint32_t>((dataPtr + cast<uint32_t>(0ULL))),threedo_ebDescribePodsReply);
threedo_Machine_writeWord(m,cast<uint32_t>((dataPtr + cast<uint32_t>(4ULL))),cast<uint32_t>(1ULL));
threedo_Machine_Write(m,cast<uint32_t>((dataPtr + cast<uint32_t>(8ULL))),cast<uint8_t>(1ULL));
threedo_Machine_Write(m,cast<uint32_t>((dataPtr + cast<uint32_t>(9ULL))),cast<uint8_t>(1ULL));
threedo_Machine_Write(m,cast<uint32_t>((dataPtr + cast<uint32_t>(10ULL))),cast<uint8_t>(0ULL));
threedo_Machine_Write(m,cast<uint32_t>((dataPtr + cast<uint32_t>(11ULL))),cast<uint8_t>(0ULL));
threedo_Machine_writeWord(m,cast<uint32_t>((dataPtr + cast<uint32_t>(12ULL))),cast<uint32_t>(0ULL));
threedo_Machine_writeWord(m,cast<uint32_t>((dataPtr + cast<uint32_t>(16ULL))),cast<uint32_t>(0ULL));
threedo_Machine_writeWord(m,cast<uint32_t>((dataPtr + cast<uint32_t>(20ULL))),cast<uint32_t>(0ULL));
threedo_Machine_writeWord(m,cast<uint32_t>((dataPtr + cast<uint32_t>(24ULL))),threedo_podIsControlPad);
for (uint32_t i = cast<uint32_t>(cast<uint32_t>(0ULL));(i < cast<uint32_t>(16ULL));i++){
threedo_Machine_Write(m,cast<uint32_t>((cast<uint32_t>((dataPtr + cast<uint32_t>(28ULL))) + i)),cast<uint8_t>(0ULL));
}
threedo_Machine_Write(m,cast<uint32_t>((dataPtr + cast<uint32_t>(28ULL))),cast<uint8_t>(1ULL));
threedo_Machine_writeWord(m,cast<uint32_t>((dataPtr + cast<uint32_t>(44ULL))),cast<uint32_t>(0ULL));
}
threedo_Machine_note(m,go_fmt_Sprintf(std::string("eventbroker: described 1 control pad to task #%d (bufsize 0x%X)",63),msg->owner,dataSize));
break;}
default:{
threedo_Machine_note(m,go_fmt_Sprintf(std::string("eventbroker: msg %d flavor %d from task #%d acknowledged",56),msg->num,flavor,msg->owner));
break;}
}}
threedo_Machine_replyMsg(m,msg,cast<uint32_t>(0ULL));
}
}
// tools/platform/threedo/io.go:659:1
void threedo_Machine_SendPadEvent(threedo_Machine* m,uint32_t buttons){
{
{auto&& tmp114 = m->ebListeners;
for(int64_t tmp115=0;tmp115<len(tmp114);++tmp115){
auto lp=tmp114[tmp115];threedo_item* port = get(m->items,lp);
if ((!port)) {
continue;
}
uint32_t rec = threedo_heap_alloc(m->dheap,cast<uint32_t>(64ULL));
if ((rec == cast<uint32_t>(0ULL))) {
return ;
}
threedo_Machine_writeWord(m,rec,threedo_ebEventRecord);
uint32_t f = cast<uint32_t>((rec + cast<uint32_t>(4ULL)));
threedo_Machine_writeWord(m,cast<uint32_t>((f + cast<uint32_t>(0ULL))),cast<uint32_t>(32ULL));
threedo_Machine_writeWord(m,cast<uint32_t>((f + cast<uint32_t>(4ULL))),cast<uint32_t>(0ULL));
threedo_Machine_writeWord(m,cast<uint32_t>((f + cast<uint32_t>(8ULL))),m->vblank);
threedo_Machine_writeWord(m,cast<uint32_t>((f + cast<uint32_t>(12ULL))),cast<uint32_t>(0ULL));
threedo_Machine_Write(m,cast<uint32_t>((f + cast<uint32_t>(16ULL))),threedo_eventNumButtonUpdate);
threedo_Machine_Write(m,cast<uint32_t>((f + cast<uint32_t>(17ULL))),cast<uint8_t>(1ULL));
threedo_Machine_Write(m,cast<uint32_t>((f + cast<uint32_t>(18ULL))),cast<uint8_t>(1ULL));
threedo_Machine_Write(m,cast<uint32_t>((f + cast<uint32_t>(19ULL))),cast<uint8_t>(1ULL));
threedo_Machine_Write(m,cast<uint32_t>((f + cast<uint32_t>(20ULL))),cast<uint8_t>(0ULL));
threedo_Machine_writeWord(m,cast<uint32_t>((f + cast<uint32_t>(24ULL))),cast<uint32_t>(0ULL));
threedo_Machine_writeWord(m,cast<uint32_t>((f + cast<uint32_t>(28ULL))),buttons);
threedo_Machine_writeWord(m,cast<uint32_t>((f + cast<uint32_t>(32ULL))),cast<uint32_t>(0ULL));
threedo_item* msg = threedo_Machine_createItem(m,cast<uint32_t>((cast<int64_t>(256ULL) | threedo_typeMsg)),cast<uint32_t>(0ULL),cast<uint32_t>(0ULL));
if ((msg->addr != cast<uint32_t>(0ULL))) {
threedo_Machine_writeWord(m,cast<uint32_t>((msg->addr + threedo_msgReplyPort)),threedo_Machine_brokerPortNum(m));
threedo_Machine_write32(m,cast<uint32_t>((msg->addr + threedo_msgDataPtr)),rec);
threedo_Machine_write32(m,cast<uint32_t>((msg->addr + threedo_msgDataSize)),cast<uint32_t>(40ULL));
threedo_Machine_write32(m,cast<uint32_t>((msg->addr + threedo_msgMsgPort)),cast<uint32_t>(port->num));
}
port->msgs = append(port->msgs,msg->num);
if (bool(m->OnMsgQueue)) {
m->OnMsgQueue(m,port->num,msg->num,std::string("PadEvent",8));
}
threedo_Machine_sendSignal(m,port->owner,port->signal);
}}
threedo_Machine_note(m,go_fmt_Sprintf(std::string("pad event 0x%08X -> %d listener(s)",34),buttons,len(m->ebListeners)));
}
}
// tools/platform/threedo/io.go:701:1
uint32_t threedo_Machine_brokerPortNum(threedo_Machine* m){
{
{auto&& tmp116 = m->items;
for(auto [tmp117,tmp118]:tmp116){
auto it=tmp118;if ((it->name == std::string("eventbroker",11))) {
return cast<uint32_t>(it->num);
}
}}
return cast<uint32_t>(0ULL);
}
}
// tools/platform/threedo/io.go:716:1
uint32_t threedo_Machine_kprintf(threedo_Machine* m){
{
arm60_CPU* c = m->CPU;
std::string format = threedo_Machine_readCStr(m,arm60_CPU_Reg(c,cast<uint32_t>(0ULL)));
int64_t argi = cast<int64_t>(0ULL);
std::function<uint32_t()> nextArg = [&]()->uint32_t{
uint32_t v={};
{
switch(argi){
case cast<int64_t>(0ULL):{
v = arm60_CPU_Reg(c,cast<uint32_t>(1ULL));
break;}
case cast<int64_t>(1ULL):{
v = arm60_CPU_Reg(c,cast<uint32_t>(2ULL));
break;}
case cast<int64_t>(2ULL):{
v = arm60_CPU_Reg(c,cast<uint32_t>(3ULL));
break;}
default:{
v = threedo_Machine_read32(m,cast<uint32_t>((arm60_CPU_Reg(c,cast<uint32_t>(13ULL)) + cast<uint32_t>((cast<uint32_t>(cast<int64_t>((argi - cast<int64_t>(3ULL)))) * cast<uint32_t>(4ULL))))));
break;}
}}
argi++;
return v;
}
;
Slice<uint8_t> out={};
for (int64_t i = cast<int64_t>(0ULL);(i < len(format));i++){
uint8_t ch = format[i];
if (((ch != '%') || (cast<int64_t>((i + cast<int64_t>(1ULL))) >= len(format)))) {
out = append(out,ch);
continue;
}
i++;
for (;((i < len(format)) && (((((((((format[i] == '-') || (format[i] == '+')) || (format[i] == ' ')) || (format[i] == '#')) || (format[i] == '0')) || (format[i] == '.')) || (format[i] == 'l')) || (((format[i] >= '1') && (format[i] <= '9'))))));){
i++;
}
if ((i >= len(format))) {
break;
}
{
switch(format[i]){
case 'd':case 'i':{
out = append(out,cast<Slice<uint8_t>>(go_fmt_Sprintf(std::string("%d",2),cast<int32_t>(nextArg()))));
break;}
case 'u':{
out = append(out,cast<Slice<uint8_t>>(go_fmt_Sprintf(std::string("%d",2),nextArg())));
break;}
case 'x':{
out = append(out,cast<Slice<uint8_t>>(go_fmt_Sprintf(std::string("%x",2),nextArg())));
break;}
case 'X':{
out = append(out,cast<Slice<uint8_t>>(go_fmt_Sprintf(std::string("%X",2),nextArg())));
break;}
case 'p':{
out = append(out,cast<Slice<uint8_t>>(go_fmt_Sprintf(std::string("0x%08X",6),nextArg())));
break;}
case 'c':{
out = append(out,cast<uint8_t>(nextArg()));
break;}
case 's':{
out = append(out,cast<Slice<uint8_t>>(threedo_Machine_readCStr(m,nextArg())));
break;}
case '%':{
out = append(out,'%');
break;}
default:{
out = append(out,'%',format[i]);
break;}
}}
}
m->tty = append(m->tty,out);
return cast<uint32_t>(0ULL);
}
}
// tools/platform/threedo/kernel.go:93:1
bool threedo_Machine_kernelSWI(threedo_Machine* m,arm60_CPU* c,uint32_t swi){
{
{
switch(swi){
case threedo_swiCreateSizedItem:{
if ((arm60_CPU_Reg(c,cast<uint32_t>(0ULL)) == cast<uint32_t>(261ULL))) {
arm60_CPU_SetReg(c,cast<uint32_t>(0ULL),cast<uint32_t>(threedo_Machine_spawnTask(m,arm60_CPU_Reg(c,cast<uint32_t>(1ULL)))));
break;
}
threedo_item* it = threedo_Machine_createItem(m,arm60_CPU_Reg(c,cast<uint32_t>(0ULL)),arm60_CPU_Reg(c,cast<uint32_t>(1ULL)),arm60_CPU_Reg(c,cast<uint32_t>(2ULL)));
threedo_Machine_initItemFromTags(m,it);
arm60_CPU_SetReg(c,cast<uint32_t>(0ULL),cast<uint32_t>(it->num));
break;}
case threedo_swiFindItem:{
arm60_CPU_SetReg(c,cast<uint32_t>(0ULL),cast<uint32_t>(threedo_Machine_findItem(m,arm60_CPU_Reg(c,cast<uint32_t>(0ULL)),arm60_CPU_Reg(c,cast<uint32_t>(1ULL)))->num));
break;}
case threedo_swiPrintf:{
arm60_CPU_SetReg(c,cast<uint32_t>(0ULL),threedo_Machine_kprintf(m));
break;}
case threedo_swiSendIO:{
threedo_Machine_serviceIO(m,c,true);
break;}
case threedo_swiDoIO:{
threedo_Machine_serviceIO(m,c,false);
break;}
case threedo_swiWaitIO:{
arm60_CPU_SetReg(c,cast<uint32_t>(0ULL),threedo_Machine_ioError(m,cast<int32_t>(arm60_CPU_Reg(c,cast<uint32_t>(0ULL)))));
break;}
case threedo_swiAbortIO:case threedo_swiCompleteIO:{
arm60_CPU_SetReg(c,cast<uint32_t>(0ULL),cast<uint32_t>(0ULL));
break;}
case threedo_swiPutMsg:case threedo_swiReplyMsg:case threedo_swiGetMsg:case threedo_swiGetThisMsg:{
threedo_Machine_serviceMsg(m,c,swi);
break;}
case threedo_swiOpenItem:{
arm60_CPU_SetReg(c,cast<uint32_t>(0ULL),arm60_CPU_Reg(c,cast<uint32_t>(0ULL)));
break;}
case threedo_swiLockItem:{
arm60_CPU_SetReg(c,cast<uint32_t>(0ULL),cast<uint32_t>(1ULL));
break;}
case threedo_swiDeleteItem:{
{
threedo_item* it = get(m->items,cast<int32_t>(arm60_CPU_Reg(c,cast<uint32_t>(0ULL))));
if ((bool(it) && (it->typ == threedo_typeAudioCue))) {
{
threedo_task* t = threedo_Machine_taskByNum(m,it->owner);
if (bool(t)) {
t->allocSigs &= ~(it->signal);
}
}
removeKey(m->items,it->num);
}
}
arm60_CPU_SetReg(c,cast<uint32_t>(0ULL),cast<uint32_t>(0ULL));
break;}
case threedo_swiUnlockItem:case threedo_swiCloseItem:case threedo_swiSetItemPri:{
arm60_CPU_SetReg(c,cast<uint32_t>(0ULL),cast<uint32_t>(0ULL));
break;}
case threedo_swiWaitSignal:{
threedo_task* t = threedo_Machine_curTask(m);
uint32_t mask = arm60_CPU_Reg(c,cast<uint32_t>(0ULL));
{
uint32_t got = cast<uint32_t>((t->sig & mask));
if (((got != cast<uint32_t>(0ULL)) || (mask == cast<uint32_t>(0ULL)))) {
t->sig &= ~(got);
arm60_CPU_SetReg(c,cast<uint32_t>(0ULL),got);
}
else {
arm60_CPU_SetReg(c,cast<uint32_t>(0ULL),mask);
t->wait = mask;
t->state = threedo_stWaiting;
m->needSchedule = true;
}
}
break;}
case threedo_swiSendSignal:{
threedo_Machine_sendSignal(m,cast<int32_t>(arm60_CPU_Reg(c,cast<uint32_t>(0ULL))),arm60_CPU_Reg(c,cast<uint32_t>(1ULL)));
arm60_CPU_SetReg(c,cast<uint32_t>(0ULL),cast<uint32_t>(0ULL));
break;}
case threedo_swiAllocSignal:{
arm60_CPU_SetReg(c,cast<uint32_t>(0ULL),threedo_Machine_allocSignalFor(m,threedo_Machine_curTask(m)->num));
break;}
case threedo_swiFreeSignal:{
threedo_Machine_curTask(m)->allocSigs &= ~(arm60_CPU_Reg(c,cast<uint32_t>(0ULL)));
arm60_CPU_SetReg(c,cast<uint32_t>(0ULL),cast<uint32_t>(0ULL));
break;}
default:{
return false;
break;}
}}
return true;
}
}
// tools/platform/threedo/kernel.go:173:1
threedo_item* threedo_Machine_createItem(threedo_Machine* m,uint32_t typ,uint32_t tags,uint32_t size){
{
m->nextItem++;
threedo_item* it = arenaNew(threedo_item{m->nextItem,typ,{},tags,{},threedo_Machine_curTask(m)->num,{},{},{},{}});
uint32_t structSize = cast<uint32_t>(cast<uint32_t>(256ULL));
if ((size > structSize)) {
structSize = size;
}
it->addr = threedo_heap_alloc(m->iheap,structSize);
if ((it->addr != cast<uint32_t>(0ULL))) {
threedo_Machine_Write(m,cast<uint32_t>((it->addr + cast<uint32_t>(8ULL))),cast<uint8_t>(shr<uint32_t>(typ,cast<int64_t>(8ULL))));
threedo_Machine_Write(m,cast<uint32_t>((it->addr + cast<uint32_t>(9ULL))),cast<uint8_t>(typ));
threedo_Machine_writeWord(m,cast<uint32_t>((it->addr + cast<uint32_t>(24ULL))),cast<uint32_t>(it->num));
}
m->items[it->num] = it;
return it;
}
}
// tools/platform/threedo/kernel.go:203:1
void threedo_Machine_initItemFromTags(threedo_Machine* m,threedo_item* it){
{
if ((it->typ == threedo_typeAudioCue)) {
it->signal = threedo_Machine_allocSignalFor(m,it->owner);
return ;
}
{
switch(cast<uint32_t>((it->typ & cast<uint32_t>(255ULL)))){
case threedo_typeMsgPort:{
it->signal = threedo_Machine_tagArg(m,it->tags,threedo_tagPortSignal);
if ((it->signal == cast<uint32_t>(0ULL))) {
it->signal = threedo_Machine_allocSignalFor(m,it->owner);
}
if ((it->addr != cast<uint32_t>(0ULL))) {
threedo_Machine_writeWord(m,cast<uint32_t>((it->addr + cast<uint32_t>(36ULL))),it->signal);
}
break;}
case threedo_typeIOReq:{
it->replyPort = cast<int32_t>(threedo_Machine_tagArg(m,it->tags,threedo_tagIOReqReplyPort));
it->device = cast<int32_t>(threedo_Machine_tagArg(m,it->tags,threedo_tagIOReqDevice));
if ((it->addr != cast<uint32_t>(0ULL))) {
threedo_Machine_write32(m,cast<uint32_t>((it->addr + threedo_ioFlagsOff)),cast<uint32_t>((threedo_ioDone | threedo_ioQuick)));
if ((it->replyPort == cast<int32_t>(0ULL))) {
threedo_Machine_write32(m,cast<uint32_t>((it->addr + threedo_ioMsgItemOff)),cast<uint32_t>(it->owner));
}
}
break;}
case threedo_typeMsg:{
if ((it->addr != cast<uint32_t>(0ULL))) {
threedo_Machine_writeWord(m,cast<uint32_t>((it->addr + threedo_msgReplyPort)),threedo_Machine_tagArg(m,it->tags,threedo_tagMsgReplyPort));
}
break;}
}}
}
}
// tools/platform/threedo/kernel.go:250:1
uint32_t threedo_Machine_allocSignalFor(threedo_Machine* m,int32_t taskNum){
{
threedo_task* t = threedo_Machine_taskByNum(m,taskNum);
if ((!t)) {
t = threedo_Machine_curTask(m);
}
uint32_t bit = cast<uint32_t>(cast<uint32_t>(256ULL));
for (;((bit != cast<uint32_t>(0ULL)) && (cast<uint32_t>((t->allocSigs & bit)) != cast<uint32_t>(0ULL)));){
bit = shl<uint32_t>(bit,cast<int64_t>(1ULL));
}
t->allocSigs |= bit;
return bit;
}
}
// tools/platform/threedo/kernel.go:264:1
Slice<std::string> threedo_Machine_ItemsSummary(threedo_Machine* m){
{
Slice<int32_t> nums={};
{auto&& tmp119 = m->items;
for(auto [tmp120,tmp121]:tmp119){
auto n=tmp120;nums = append(nums,n);
}}
go_sort_Slice(nums,[&](int64_t i,int64_t j)->bool{
return (nums[i] < nums[j]);
}
);
Slice<std::string> out={};
{auto&& tmp122 = nums;
for(int64_t tmp123=0;tmp123<len(tmp122);++tmp123){
auto n=tmp122[tmp123];threedo_item* it = get(m->items,n);
std::string s = go_fmt_Sprintf(std::string("item %d type=0x%X owner=#%d",27),it->num,it->typ,it->owner);
if ((it->name != std::string("",0))) {
s += go_fmt_Sprintf(std::string(" name=%q",8),it->name);
}
if ((it->signal != cast<uint32_t>(0ULL))) {
s += go_fmt_Sprintf(std::string(" signal=0x%X",12),it->signal);
}
if ((len(it->msgs) > cast<int64_t>(0ULL))) {
s += go_fmt_Sprintf(std::string(" queued=%v",10),it->msgs);
}
if ((it->device != cast<int32_t>(0ULL))) {
s += go_fmt_Sprintf(std::string(" device=%d",10),it->device);
}
if ((it->replyPort != cast<int32_t>(0ULL))) {
s += go_fmt_Sprintf(std::string(" replyPort=%d",13),it->replyPort);
}
if (((cast<uint32_t>((it->typ & cast<uint32_t>(255ULL))) == threedo_typeMsg) && (it->addr != cast<uint32_t>(0ULL)))) {
s += go_fmt_Sprintf(std::string(" msgReplyPort=%d result=0x%X",28),threedo_Machine_read32(m,cast<uint32_t>((it->addr + threedo_msgReplyPort))),threedo_Machine_read32(m,cast<uint32_t>((it->addr + threedo_msgResult))));
}
out = append(out,s);
}}
return out;
}
}
// tools/platform/threedo/kernel.go:299:1
threedo_item* threedo_Machine_findItem(threedo_Machine* m,uint32_t typ,uint32_t tags){
{
std::string name = threedo_Machine_tagString(m,tags,threedo_tagItemName);
uint32_t key = typ;
{
auto tmp124 = lookup(m->itemByType,key);
threedo_item* it = std::get<0>(tmp124);
bool ok = std::get<1>(tmp124);
if ((ok && (it->name == name))) {
return it;
}
}
threedo_item* it = threedo_Machine_createItem(m,typ,tags,cast<uint32_t>(0ULL));
it->name = name;
m->itemByType[key] = it;
return it;
}
}
// tools/platform/threedo/machine.go:281:1
threedo_Machine* threedo_NewMachine(){
{
threedo_Machine* m = arenaNew(threedo_Machine{Slice<uint8_t>::make(threedo_dramSize),Slice<uint8_t>::make(threedo_vramSize),threedo_newHeap(threedo_dheapBase,cast<uint32_t>((threedo_dheapTop - threedo_dheapBase))),threedo_newHeap(threedo_vheapBase,cast<uint32_t>((threedo_vheapTop - threedo_vheapBase))),Slice<uint8_t>::make(threedo_imemSize),threedo_newHeap(threedo_imemBase,threedo_imemSize),{},{},Map<uint32_t,threedo_diskStream*>{},Map<uint32_t,threedo_dirScan*>{},Map<std::string,Slice<uint8_t>>{},Map<int32_t,threedo_gfxBitmap>{},Map<int32_t,int32_t>{},{},{},{},{},Map<int32_t,int32_t>{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},Map<int32_t,threedo_item*>{},Map<uint32_t,threedo_item*>{},cast<int32_t>(4096ULL),{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},Map<std::string,bool>{},{},{},{},{},{},{},{},{},{},{},{}});
m->CPU = arm60_NewCPU(m);
m->CPU->SWI = [=](auto...args){return threedo_Machine_swi(m,args...);};
threedo_Machine_initTasks(m);
return m;
}
}
// tools/platform/threedo/machine.go:310:1
void threedo_Machine_LoadAIF(threedo_Machine* m,threedo_AIF* a){
{
uint32_t base = cast<uint32_t>((a->ImageBase & (cast<uint32_t>((threedo_dramSize - cast<int64_t>(1ULL))))));
gcopy(sub(m->dram,base,len(m->dram)),a->Image);
for (uint32_t off = cast<uint32_t>(cast<uint32_t>(4ULL));(off <= cast<uint32_t>(4096ULL));off += cast<uint32_t>(4ULL)){
threedo_Machine_writeWord(m,cast<uint32_t>((threedo_kernelBase - off)),cast<uint32_t>((threedo_hleBase + off)));
}
for (uint32_t off = cast<uint32_t>(cast<uint32_t>(4ULL));(off <= cast<uint32_t>(256ULL));off += cast<uint32_t>(4ULL)){
threedo_Machine_writeWord(m,cast<uint32_t>((threedo_fileFolioBase - off)),cast<uint32_t>((cast<uint32_t>((threedo_hleBase + threedo_hleFileTag)) + off)));
threedo_Machine_writeWord(m,cast<uint32_t>((threedo_otherFolioBase - off)),cast<uint32_t>((cast<uint32_t>((threedo_hleBase + threedo_hleOtherTag)) + off)));
threedo_Machine_writeWord(m,cast<uint32_t>((threedo_gfxFolioBase - off)),cast<uint32_t>((cast<uint32_t>((threedo_hleBase + threedo_hleGfxTag)) + off)));
threedo_Machine_writeWord(m,cast<uint32_t>((threedo_audioFolioBase - off)),cast<uint32_t>((cast<uint32_t>((threedo_hleBase + threedo_hleAudioTag)) + off)));
threedo_Machine_writeWord(m,cast<uint32_t>((threedo_mathFolioBase - off)),cast<uint32_t>((cast<uint32_t>((threedo_hleBase + threedo_hleMathTag)) + off)));
}
threedo_Machine_writeWord(m,cast<uint32_t>((threedo_kernelBase + cast<int64_t>(152ULL))),threedo_osCtxBase);
m->dram[cast<int64_t>((threedo_osCtxBase + threedo_osCtxPri))] = cast<uint8_t>(100ULL);
threedo_Machine_writeWord(m,cast<uint32_t>((threedo_osCtxBase + threedo_osCtxItem)),cast<uint32_t>(threedo_Machine_curTask(m)->num));
threedo_Machine_writeWord(m,cast<uint32_t>((threedo_osCtxBase + threedo_osCtxMemLst)),cast<uint32_t>((threedo_osCtxBase + cast<int64_t>(1024ULL))));
arm60_CPU_SetReg(m->CPU,cast<uint32_t>(5ULL),cast<uint32_t>(0ULL));
arm60_CPU_SetReg(m->CPU,cast<uint32_t>(6ULL),cast<uint32_t>(0ULL));
arm60_CPU_SetReg(m->CPU,cast<uint32_t>(7ULL),threedo_kernelBase);
arm60_CPU_SetReg(m->CPU,cast<uint32_t>(13ULL),cast<uint32_t>((threedo_dramSize - cast<int64_t>(4096ULL))));
arm60_CPU_SetPC(m->CPU,a->ImageBase);
}
}
// tools/platform/threedo/machine.go:347:1
void threedo_Machine_SetVolume(threedo_Machine* m,threedo_Volume* v){
{
m->vol = v;
}
}
// tools/platform/threedo/machine.go:351:1
uint32_t threedo_Machine_DisplayBuffer(threedo_Machine* m){
{
return m->displayBuf;
}
}
// tools/platform/threedo/machine.go:357:1
void threedo_Machine_SetVBLMirror(threedo_Machine* m,uint32_t addr){
{
m->vblMirror = addr;
}
}
// tools/platform/threedo/machine.go:372:1
void threedo_Machine_advanceVBlank(threedo_Machine* m,uint32_t n){
{
if ((n == cast<uint32_t>(0ULL))) {
n = cast<uint32_t>(1ULL);
}
m->vblank += n;
if ((m->vblMirror != cast<uint32_t>(0ULL))) {
threedo_Machine_writeWord(m,m->vblMirror,cast<uint32_t>((n * cast<uint32_t>(100ULL))));
}
threedo_Machine_writeWord(m,cast<uint32_t>((threedo_gfxFolioBase + cast<int64_t>(116ULL))),m->vblank);
threedo_Machine_advanceAudioClock(m,n);
}
}
// tools/platform/threedo/machine.go:406:1
void threedo_Machine_startFieldWait(threedo_Machine* m,threedo_item* it,uint32_t fields){
{
if ((!it)) {
return ;
}
if ((fields == cast<uint32_t>(0ULL))) {
fields = cast<uint32_t>(1ULL);
}
if ((it->addr != cast<uint32_t>(0ULL))) {
threedo_Machine_write32(m,cast<uint32_t>((it->addr + threedo_ioFlagsOff)),cast<uint32_t>(0ULL));
}
m->fieldWaits = append(m->fieldWaits,threedo_timerWait{it->num,threedo_Machine_curTask(m)->num,cast<uint32_t>((m->vblank + fields))});
}
}
// tools/platform/threedo/machine.go:425:1
bool threedo_Machine_fieldTick(threedo_Machine* m){
{
threedo_Machine_advanceVBlank(m,cast<uint32_t>(1ULL));
if ((len(m->fieldWaits) == cast<int64_t>(0ULL))) {
return false;
}
bool woke = false;
Slice<threedo_timerWait> kept = sub(m->fieldWaits,0,cast<int64_t>(0ULL));
{auto&& tmp125 = m->fieldWaits;
for(int64_t tmp126=0;tmp126<len(tmp125);++tmp126){
auto w=tmp125[tmp126];if ((cast<int32_t>(cast<uint32_t>((m->vblank - w.field))) >= cast<int32_t>(0ULL))) {
{
threedo_item* it = get(m->items,w.ioReq);
if (bool(it)) {
if ((it->addr != cast<uint32_t>(0ULL))) {
threedo_Machine_write32(m,cast<uint32_t>((it->addr + threedo_ioActualOff)),cast<uint32_t>(0ULL));
threedo_Machine_write32(m,cast<uint32_t>((it->addr + threedo_ioFlagsOff)),cast<uint32_t>((threedo_Machine_read32(m,cast<uint32_t>((it->addr + threedo_ioFlagsOff))) | threedo_ioDone)));
threedo_Machine_write32(m,cast<uint32_t>((it->addr + threedo_ioErrorOff)),cast<uint32_t>(0ULL));
}
threedo_Machine_completeFieldWait(m,it,w.submitter);
woke = true;
}
}
}
else {
kept = append(kept,w);
}
}}
m->fieldWaits = kept;
return woke;
}
}
// tools/platform/threedo/machine.go:458:1
bool threedo_Machine_wakeByFieldTick(threedo_Machine* m){
{
if ((len(m->fieldWaits) == cast<int64_t>(0ULL))) {
return false;
}
for (int64_t i = cast<int64_t>(0ULL);(i < cast<int64_t>(1000000ULL));i++){
threedo_Machine_fieldTick(m);
if (threedo_Machine_switchTask(m)) {
return true;
}
if ((len(m->fieldWaits) == cast<int64_t>(0ULL))) {
return false;
}
}
return false;
}
}
// tools/platform/threedo/machine.go:475:1
void threedo_Machine_writeWord(threedo_Machine* m,uint32_t a,uint32_t v){
{
if ((cast<int64_t>((cast<int64_t>(a) + cast<int64_t>(4ULL))) <= len(m->dram))) {
m->dram[a] = cast<uint8_t>(shr<uint32_t>(v,cast<int64_t>(24ULL)));
m->dram[cast<uint32_t>((a + cast<uint32_t>(1ULL)))] = cast<uint8_t>(shr<uint32_t>(v,cast<int64_t>(16ULL)));
m->dram[cast<uint32_t>((a + cast<uint32_t>(2ULL)))] = cast<uint8_t>(shr<uint32_t>(v,cast<int64_t>(8ULL)));
m->dram[cast<uint32_t>((a + cast<uint32_t>(3ULL)))] = cast<uint8_t>(v);
return ;
}
threedo_Machine_Write(m,a,cast<uint8_t>(shr<uint32_t>(v,cast<int64_t>(24ULL))));
threedo_Machine_Write(m,cast<uint32_t>((a + cast<uint32_t>(1ULL))),cast<uint8_t>(shr<uint32_t>(v,cast<int64_t>(16ULL))));
threedo_Machine_Write(m,cast<uint32_t>((a + cast<uint32_t>(2ULL))),cast<uint8_t>(shr<uint32_t>(v,cast<int64_t>(8ULL))));
threedo_Machine_Write(m,cast<uint32_t>((a + cast<uint32_t>(3ULL))),cast<uint8_t>(v));
}
}
// tools/platform/threedo/machine.go:491:1
void threedo_Machine_note(threedo_Machine* m,std::string s){
{
if ((!get(m->logSeen,s))) {
m->logSeen[s] = true;
m->Log = append(m->Log,s);
}
}
}
// tools/platform/threedo/machine.go:499:1
std::string threedo_Machine_TTY(threedo_Machine* m){
{
return cast<std::string>(m->tty);
}
}
// tools/platform/threedo/machine.go:503:1
uint8_t threedo_Machine_Read(threedo_Machine* m,uint32_t addr){
{
if (((bool(m->OnRead) && (addr >= m->RWatchLo)) && (addr < m->RWatchHi))) {
m->OnRead(addr,cast<uint32_t>(threedo_Machine_rawRead(m,addr)),arm60_CPU_CurPC(m->CPU));
}
return threedo_Machine_rawRead(m,addr);
}
}
// tools/platform/threedo/machine.go:510:1
uint8_t threedo_Machine_rawRead(threedo_Machine* m,uint32_t addr){
{
{
if ((addr < threedo_dramSize)){
return m->dram[addr];
}
else if (((addr >= threedo_vramBase) && (addr < cast<uint32_t>((threedo_vramBase + threedo_vramSize))))){
return m->vram[cast<uint32_t>((addr - threedo_vramBase))];
}
else if (((addr >= threedo_imemBase) && (addr < cast<uint32_t>((threedo_imemBase + threedo_imemSize))))){
return m->imem[cast<uint32_t>((addr - threedo_imemBase))];
}
else if (((addr >= threedo_madamBase) && (addr < threedo_clioEnd))){
return cast<uint8_t>(0ULL);
}
else if ((addr >= cast<uint32_t>(4294963200ULL))){
uint32_t wbase = (addr & ~(cast<uint32_t>(3ULL)));
uint32_t w = cast<uint32_t>((threedo_hleBase - wbase));
return cast<uint8_t>(shr<uint32_t>(w,cast<uint64_t>(cast<uint32_t>(((cast<uint32_t>((cast<uint32_t>(3ULL) - (cast<uint32_t>((addr & cast<uint32_t>(3ULL))))))) * cast<uint32_t>(8ULL))))));
}
else {
threedo_Machine_note(m,go_fmt_Sprintf(std::string("read unmapped 0x%08X from PC 0x%08X",35),addr,arm60_CPU_CurPC(m->CPU)));
return cast<uint8_t>(0ULL);
}
}
tmp127:;
}
}
// tools/platform/threedo/machine.go:535:1
void threedo_Machine_Write(threedo_Machine* m,uint32_t addr,uint8_t v){
{
if (((bool(m->OnWrite) && (addr >= m->WatchLo)) && (addr < m->WatchHi))) {
m->OnWrite(addr,cast<uint32_t>(v),arm60_CPU_CurPC(m->CPU));
}
{
if ((addr < threedo_dramSize)){
m->dram[addr] = v;
}
else if (((addr >= threedo_vramBase) && (addr < cast<uint32_t>((threedo_vramBase + threedo_vramSize))))){
m->vram[cast<uint32_t>((addr - threedo_vramBase))] = v;
}
else if (((addr >= threedo_imemBase) && (addr < cast<uint32_t>((threedo_imemBase + threedo_imemSize))))){
m->imem[cast<uint32_t>((addr - threedo_imemBase))] = v;
}
else if (((addr >= threedo_madamBase) && (addr < threedo_clioEnd))){
}
else {
threedo_Machine_note(m,go_fmt_Sprintf(std::string("write unmapped 0x%08X from PC 0x%08X",36),addr,arm60_CPU_CurPC(m->CPU)));
}
}
tmp128:;
}
}
// tools/platform/threedo/machine.go:563:1
bool threedo_Machine_isFlagSpin(threedo_Machine* m,Slice<uint32_t> pcs){
{
{auto&& tmp129 = pcs;
for(int64_t tmp130=0;tmp130<len(tmp129);++tmp130){
auto pc=tmp129[tmp130];if ((cast<uint32_t>((pc + cast<uint32_t>(4ULL))) > threedo_dramSize)) {
continue;
}
uint32_t w = threedo_be32(sub(m->dram,pc,len(m->dram)));
uint32_t op = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(21ULL))) & cast<uint32_t>(15ULL)));
if ((((((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(26ULL))) & cast<uint32_t>(3ULL))) == cast<uint32_t>(0ULL)) && (cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(20ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) && ((((op == cast<uint32_t>(8ULL)) || (op == cast<uint32_t>(9ULL))) || (op == cast<uint32_t>(10ULL))))) && (cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(25ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) && (cast<uint32_t>((w & cast<uint32_t>(4095ULL))) == cast<uint32_t>(0ULL)))) {
return true;
}
}}
return false;
}
}
// tools/platform/threedo/machine.go:577:1
Slice<uint32_t> threedo_Machine_breakSpin(threedo_Machine* m,Slice<uint32_t> pcs){
{
if ((!threedo_Machine_isFlagSpin(m,pcs))) {
return {};
}
Slice<uint32_t> poked={};
Map<uint32_t,bool> done = Map<uint32_t,bool>{};
{auto&& tmp131 = pcs;
for(int64_t tmp132=0;tmp132<len(tmp131);++tmp132){
auto pc=tmp131[tmp132];if ((get(done,pc) || (cast<uint32_t>((pc + cast<uint32_t>(4ULL))) > threedo_dramSize))) {
continue;
}
done[pc] = true;
uint32_t w = threedo_be32(sub(m->dram,pc,len(m->dram)));
if ((((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(26ULL))) & cast<uint32_t>(3ULL))) != cast<uint32_t>(1ULL)) || (cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(25ULL))) & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL))) || (cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(20ULL))) & cast<uint32_t>(1ULL))) != cast<uint32_t>(1ULL)))) {
continue;
}
uint32_t base = arm60_CPU_Reg(m->CPU,cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(15ULL))));
uint32_t imm = cast<uint32_t>((w & cast<uint32_t>(4095ULL)));
uint32_t addr = base;
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(24ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(23ULL))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
addr = cast<uint32_t>((base + imm));
}
else {
addr = cast<uint32_t>((base - imm));
}
}
{
if ((addr < threedo_dramSize)){
m->dram[addr] = cast<uint8_t>(1ULL);
poked = append(poked,addr);
}
else if (((addr >= threedo_vramBase) && (addr < cast<uint32_t>((threedo_vramBase + threedo_vramSize))))){
m->vram[cast<uint32_t>((addr - threedo_vramBase))] = cast<uint8_t>(1ULL);
poked = append(poked,addr);
}
}
tmp133:;
}}
return poked;
}
}
// tools/platform/threedo/machine.go:617:1
bool threedo_Machine_swi(threedo_Machine* m,arm60_CPU* c,uint32_t comment){
rrprof::Scope rrclock(5,"Portfolio SWI",true,comment);
{
m->celCnt.swis++;
auto tmp134=defer([&](){threedo_Machine_profEnd(m,threedo_bucketSWI,threedo_Machine_profStart(m));});
if ((comment == cast<uint32_t>(17ULL))) {
auto tmp135 = std::make_tuple(true,std::string("program exit (SWI #0x11)",24));
m->Halted = std::get<0>(tmp135);
m->HaltReason = std::get<1>(tmp135);
return true;
}
m->SWICalls = append(m->SWICalls,threedo_KernelCall{comment,arm60_CPU_CurPC(c),std::array<uint32_t,4>{arm60_CPU_Reg(c,cast<uint32_t>(0ULL)),arm60_CPU_Reg(c,cast<uint32_t>(1ULL)),arm60_CPU_Reg(c,cast<uint32_t>(2ULL)),arm60_CPU_Reg(c,cast<uint32_t>(3ULL))}});
if (bool(m->OnSWI)) {
m->OnSWI(m,arm60_CPU_CurPC(c),comment);
}
if (((((!threedo_Machine_kernelSWI(m,c,comment)) && (!threedo_Machine_fileFolioSWI(m,c,comment))) && (!threedo_Machine_mathFolioSWI(m,c,comment))) && (!threedo_Machine_audioFolioSWI(m,c,comment)))) {
threedo_Machine_note(m,go_fmt_Sprintf(std::string("SWI #0x%X (stub)",16),comment));
}
return true;
}
}
// tools/platform/threedo/moviehle.go:27:1
void threedo_Machine_armMovie(threedo_Machine* m,std::string name){
{
std::string trimmed = name;
for (;((len(trimmed) > cast<int64_t>(0ULL)) && (trimmed[cast<int64_t>(0ULL)] == '/'));){
trimmed = sub(trimmed,cast<int64_t>(1ULL),len(trimmed));
}
auto tmp136 = threedo_Volume_ReadFile(m->vol,trimmed);
Slice<uint8_t> data = std::get<0>(tmp136);
Error err = std::get<1>(tmp136);
if (bool(err)) {
{
std::string base = threedo_baseName(trimmed);
if ((base != trimmed)) {
{
auto tmp137 = threedo_Volume_resolve(m->vol,base);
threedo_Entry e = std::get<0>(tmp137);
Error rerr = std::get<1>(tmp137);
if ((!rerr)) {
auto tmp138 = threedo_Volume_ReadFile(m->vol,e.Path);
data = std::get<0>(tmp138);
err = std::get<1>(tmp138);
}
}
}
}
}
if (bool(err)) {
return ;
}
auto tmp139 = threedo_DemuxStream(data);
threedo_CvidMovie* mv = std::get<0>(tmp139);
err = std::get<1>(tmp139);
if (((bool(err) || (((mv->Codec != std::string("cvid",4)) && (mv->Codec != std::string("",0))))) || (len(mv->Frames) == cast<int64_t>(0ULL)))) {
return ;
}
m->movieQueue = append(m->movieQueue,arenaNew(threedo_armedMovie{trimmed,mv}));
threedo_Machine_note(m,go_fmt_Sprintf(std::string("MovieHLE armed %s (pc=0x%X lr=0x%X)",35),trimmed,arm60_CPU_Reg(m->CPU,cast<uint32_t>(15ULL)),arm60_CPU_Reg(m->CPU,cast<uint32_t>(14ULL))));
}
}
// tools/platform/threedo/moviehle.go:55:1
int64_t threedo_Machine_MoviesPending(threedo_Machine* m){
{
return cast<int64_t>((len(m->movieQueue) - m->moviePos));
}
}
// tools/platform/threedo/moviehle.go:63:1
std::tuple<uint32_t,std::string,int64_t,int64_t,bool> threedo_Machine_StepMovieFrame(threedo_Machine* m){
uint32_t buf{};
std::string name{};
int64_t frame{};
int64_t total{};
bool ok{};
{
for (;(m->moviePos < len(m->movieQueue));){
threedo_armedMovie* am = m->movieQueue[m->moviePos];
if ((!m->movieDec)) {
m->movieDec = threedo_NewCvidDecoder(am->mv->Width,am->mv->Height);
m->movieBase = threedo_Machine_movieTarget(m);
m->movieFrameIdx = cast<int64_t>(0ULL);
}
if ((m->movieFrameIdx >= len(am->mv->Frames))) {
m->moviePos++;
m->movieDec = {};
continue;
}
threedo_CvidDecoder_DecodeFrame(m->movieDec,am->mv->Frames[m->movieFrameIdx]);
threedo_Machine_blitRGBAToVRAM(m,threedo_CvidDecoder_Frame(m->movieDec),m->movieBase,am->mv->Width,am->mv->Height);
m->displayBuf = m->movieBase;
int64_t idx = m->movieFrameIdx;
m->movieFrameIdx++;
return {m->movieBase,am->name,idx,len(am->mv->Frames),true};
}
return {cast<uint32_t>(0ULL),std::string("",0),cast<int64_t>(0ULL),cast<int64_t>(0ULL),false};
}
}
// tools/platform/threedo/moviehle.go:87:1
Slice<std::string> threedo_Machine_MovieNames(threedo_Machine* m){
{
Slice<std::string> names = Slice<std::string>::make(len(m->movieQueue));
{auto&& tmp140 = m->movieQueue;
for(int64_t tmp141=0;tmp141<len(tmp140);++tmp141){
auto i=tmp141;auto am=tmp140[tmp141];names[i] = am->name;
}}
return names;
}
}
// tools/platform/threedo/moviehle.go:101:1
void threedo_Machine_PlayMovies(threedo_Machine* m,std::function<void(std::string,int64_t,int64_t)> onFrame){
{
for (;;){
auto tmp142 = threedo_Machine_StepMovieFrame(m);
uint32_t buf = std::get<0>(tmp142);
std::string name = std::get<1>(tmp142);
int64_t frame = std::get<2>(tmp142);
int64_t total = std::get<3>(tmp142);
bool ok = std::get<4>(tmp142);
if ((!ok)) {
break;
}
if (bool(onFrame)) {
onFrame(name,frame,total);
}
m->frame++;
if (bool(m->OnDisplay)) {
m->OnDisplay(m,m->frame,buf);
}
}
}
}
// tools/platform/threedo/moviehle.go:120:1
uint32_t threedo_Machine_movieTarget(threedo_Machine* m){
{
if ((m->displayBuf != cast<uint32_t>(0ULL))) {
return m->displayBuf;
}
{
auto tmp143 = threedo_Machine_DrawTarget(m);
uint32_t buf = std::get<0>(tmp143);
bool ok = std::get<3>(tmp143);
if (ok) {
m->displayBuf = buf;
return buf;
}
}
uint32_t buf = threedo_heap_alloc(m->vheap,cast<uint32_t>(cast<uint32_t>((cast<int64_t>((threedo_screenW * threedo_screenH)) * cast<int64_t>(2ULL)))));
m->displayBuf = buf;
return buf;
}
}
// tools/platform/threedo/moviehle.go:136:1
void threedo_Machine_blitRGBAToVRAM(threedo_Machine* m,image_RGBA* img,uint32_t base,int64_t w,int64_t h){
{
int64_t yoff = divi<int64_t>((cast<int64_t>((threedo_screenH - h))),cast<int64_t>(2ULL));
if ((yoff < cast<int64_t>(0ULL))) {
yoff = cast<int64_t>(0ULL);
}
for (int64_t sy = cast<int64_t>(0ULL);(sy < threedo_screenH);sy++){
int64_t src = cast<int64_t>((sy - yoff));
for (int64_t sx = cast<int64_t>(0ULL);(sx < threedo_screenW);sx++){
uint32_t o = cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((base - threedo_vramBase)) + cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(shr<int64_t>(sy,cast<int64_t>(1ULL))) * cast<uint32_t>(threedo_screenW))) * cast<uint32_t>(4ULL))))) + cast<uint32_t>((cast<uint32_t>(sx) * cast<uint32_t>(4ULL))))) + cast<uint32_t>((cast<uint32_t>(cast<int64_t>((sy & cast<int64_t>(1ULL)))) * cast<uint32_t>(2ULL)))));
if ((cast<int64_t>((cast<int64_t>(o) + cast<int64_t>(2ULL))) > len(m->vram))) {
continue;
}
uint16_t c={};
if ((((src >= cast<int64_t>(0ULL)) && (src < h)) && (sx < w))) {
int64_t p = image_RGBA_PixOffset(img,sx,src);
c = cast<uint16_t>((cast<uint16_t>((shl<uint16_t>(cast<uint16_t>(shr<uint8_t>(img->Pix[p],cast<int64_t>(3ULL))),cast<int64_t>(10ULL)) | shl<uint16_t>(cast<uint16_t>(shr<uint8_t>(img->Pix[cast<int64_t>((p + cast<int64_t>(1ULL)))],cast<int64_t>(3ULL))),cast<int64_t>(5ULL)))) | cast<uint16_t>(shr<uint8_t>(img->Pix[cast<int64_t>((p + cast<int64_t>(2ULL)))],cast<int64_t>(3ULL)))));
}
m->vram[o] = cast<uint8_t>(shr<uint16_t>(c,cast<int64_t>(8ULL)));
m->vram[cast<uint32_t>((o + cast<uint32_t>(1ULL)))] = cast<uint8_t>(c);
}
}
}
}
// tools/platform/threedo/moviehle.go:166:1
std::string threedo_baseName(std::string p){
{
for (int64_t i = cast<int64_t>((len(p) - cast<int64_t>(1ULL)));(i >= cast<int64_t>(0ULL));i--){
if ((p[i] == '/')) {
return sub(p,cast<int64_t>((i + cast<int64_t>(1ULL))),len(p));
}
}
return p;
}
}
// tools/platform/threedo/operafs.go:70:1
std::string threedo_Entry_String(threedo_Entry e){
{
if (e.IsDir) {
return go_fmt_Sprintf(std::string("%-40s  <dir>   (blk %d, %d copies)",34),e.Path,e.Block,len(e.Copies));
}
return go_fmt_Sprintf(std::string("%-40s  %9d  (blk %d, %d copies)",31),e.Path,e.Size,e.Block,len(e.Copies));
}
}
// tools/platform/threedo/operafs.go:78:1
uint32_t threedo_be32(Slice<uint8_t> b){
{
return be_Uint32(sub(b,0,cast<int64_t>(4ULL)));
}
}
// tools/platform/threedo/operafs.go:139:1
std::string threedo_trimName(Slice<uint8_t> b){
{
{
int64_t i = threedo_indexByte(b,cast<uint8_t>(0ULL));
if ((i >= cast<int64_t>(0ULL))) {
b = sub(b,0,i);
}
}
return go_strings_TrimRight(cast<std::string>(b),std::string(" ",1));
}
}
// tools/platform/threedo/operafs.go:146:1
int64_t threedo_indexByte(Slice<uint8_t> b,uint8_t c){
{
{auto&& tmp144 = b;
for(int64_t tmp145=0;tmp145<len(tmp144);++tmp145){
auto i=tmp145;auto x=tmp144[tmp145];if ((x == c)) {
return i;
}
}}
return cast<int64_t>(-cast<int64_t>(1ULL));
}
}
// tools/platform/threedo/operafs.go:161:1
std::tuple<Slice<threedo_Entry>,Error> threedo_Volume_dirEntries(threedo_Volume* v,int64_t startBlock,std::string dirPath){
{
Slice<threedo_Entry> out={};
Map<int64_t,bool> visited = Map<int64_t,bool>{};
for (int64_t rel = cast<int64_t>(0ULL);(rel >= cast<int64_t>(0ULL));){
if (get(visited,rel)) {
return {{},go_fmt_Errorf(std::string("threedo: directory block cycle at rel %d",40),rel)};
}
visited[rel] = true;
auto tmp146 = threedo_Volume_block(v,cast<int64_t>((startBlock + rel)));
Slice<uint8_t> b = std::get<0>(tmp146);
Error err = std::get<1>(tmp146);
if (bool(err)) {
return {{},err};
}
int64_t next = cast<int64_t>(cast<int32_t>(threedo_be32(sub(b,cast<int64_t>(0ULL),len(b)))));
int64_t endOffset = cast<int64_t>(threedo_be32(sub(b,cast<int64_t>(12ULL),len(b))));
int64_t firstEntry = cast<int64_t>(threedo_be32(sub(b,cast<int64_t>(16ULL),len(b))));
if ((endOffset > threedo_userSize)) {
endOffset = threedo_userSize;
}
for (int64_t p = firstEntry;(cast<int64_t>((p + cast<int64_t>(68ULL))) <= endOffset);){
uint32_t flags = threedo_be32(sub(b,p,len(b)));
int64_t ncopies = cast<int64_t>((cast<int64_t>(threedo_be32(sub(b,cast<int64_t>((p + cast<int64_t>(64ULL))),len(b)))) + cast<int64_t>(1ULL)));
threedo_Entry e = threedo_Entry{threedo_trimName(sub(b,cast<int64_t>((p + cast<int64_t>(32ULL))),cast<int64_t>((p + cast<int64_t>(64ULL))))),{},(cast<uint32_t>((flags & threedo_entryTypeMask)) == threedo_typeDir),threedo_trimName(sub(b,cast<int64_t>((p + cast<int64_t>(8ULL))),cast<int64_t>((p + cast<int64_t>(12ULL))))),cast<int64_t>(threedo_be32(sub(b,cast<int64_t>((p + cast<int64_t>(16ULL))),len(b)))),cast<int64_t>(threedo_be32(sub(b,cast<int64_t>((p + cast<int64_t>(20ULL))),len(b)))),{},{}};
for (int64_t i = cast<int64_t>(0ULL);((i < ncopies) && (cast<int64_t>((cast<int64_t>((p + cast<int64_t>(68ULL))) + cast<int64_t>(((cast<int64_t>((i + cast<int64_t>(1ULL)))) * cast<int64_t>(4ULL))))) <= endOffset));i++){
e.Copies = append(e.Copies,cast<int64_t>(threedo_be32(sub(b,cast<int64_t>((cast<int64_t>((p + cast<int64_t>(68ULL))) + cast<int64_t>((i * cast<int64_t>(4ULL))))),len(b)))));
}
if ((len(e.Copies) > cast<int64_t>(0ULL))) {
e.Block = e.Copies[cast<int64_t>(0ULL)];
}
if ((dirPath == std::string("",0))) {
e.Path = e.Name;
}
else {
e.Path = ((dirPath + std::string("/",1)) + e.Name);
}
out = append(out,e);
p += cast<int64_t>((cast<int64_t>(68ULL) + cast<int64_t>((ncopies * cast<int64_t>(4ULL)))));
}
rel = next;
}
return {out,{}};
}
}
// tools/platform/threedo/operafs.go:213:1
std::tuple<threedo_Entry,Error> threedo_Volume_resolve(threedo_Volume* v,std::string path){
{
threedo_Entry cur = threedo_Entry{v->Label,{},true,{},{},v->rootBlocks,v->rootBlock,Slice<int64_t>{v->rootBlock}};
{auto&& tmp147 = threedo_splitPath(path);
for(int64_t tmp148=0;tmp148<len(tmp147);++tmp148){
auto want=tmp147[tmp148];if ((!cur.IsDir)) {
return {threedo_Entry{},go_fmt_Errorf(std::string("threedo: %q is not a directory",30),cur.Path)};
}
auto tmp149 = threedo_Volume_dirEntries(v,cur.Block,cur.Path);
Slice<threedo_Entry> entries = std::get<0>(tmp149);
Error err = std::get<1>(tmp149);
if (bool(err)) {
return {threedo_Entry{},err};
}
bool found = false;
{auto&& tmp150 = entries;
for(int64_t tmp151=0;tmp151<len(tmp150);++tmp151){
auto e=tmp150[tmp151];if (go_strings_EqualFold(e.Name,want)) {
auto tmp152 = std::make_tuple(e,true);
cur = std::get<0>(tmp152);
found = std::get<1>(tmp152);
break;
}
}}
if ((!found)) {
return {threedo_Entry{},go_fmt_Errorf(std::string("threedo: %q not found",21),path)};
}
}}
return {cur,{}};
}
}
// tools/platform/threedo/operafs.go:237:1
Slice<std::string> threedo_splitPath(std::string p){
{
Slice<std::string> parts={};
{auto&& tmp153 = go_strings_Split(go_strings_Trim(p,std::string("/",1)),std::string("/",1));
for(int64_t tmp154=0;tmp154<len(tmp153);++tmp154){
auto c=tmp153[tmp154];if ((c != std::string("",0))) {
parts = append(parts,c);
}
}}
return parts;
}
}
// tools/platform/threedo/operafs.go:248:1
std::tuple<Slice<threedo_Entry>,Error> threedo_Volume_ReadDir(threedo_Volume* v,std::string path){
{
auto tmp155 = threedo_Volume_resolve(v,path);
threedo_Entry e = std::get<0>(tmp155);
Error err = std::get<1>(tmp155);
if (bool(err)) {
return {{},err};
}
if ((!e.IsDir)) {
return {{},go_fmt_Errorf(std::string("threedo: %q is not a directory",30),path)};
}
return threedo_Volume_dirEntries(v,e.Block,e.Path);
}
}
// tools/platform/threedo/operafs.go:260:1
Error threedo_Volume_Walk(threedo_Volume* v,std::function<Error(threedo_Entry)> fn){
{
return threedo_Volume_walk(v,v->rootBlock,std::string("",0),fn);
}
}
// tools/platform/threedo/operafs.go:264:1
Error threedo_Volume_walk(threedo_Volume* v,int64_t block,std::string dirPath,std::function<Error(threedo_Entry)> fn){
{
auto tmp156 = threedo_Volume_dirEntries(v,block,dirPath);
Slice<threedo_Entry> entries = std::get<0>(tmp156);
Error err = std::get<1>(tmp156);
if (bool(err)) {
return err;
}
{auto&& tmp157 = entries;
for(int64_t tmp158=0;tmp158<len(tmp157);++tmp158){
auto e=tmp157[tmp158];{
Error err = fn(e);
if (bool(err)) {
return err;
}
}
if (e.IsDir) {
{
Error err = threedo_Volume_walk(v,e.Block,e.Path,fn);
if (bool(err)) {
return err;
}
}
}
}}
return {};
}
}
// tools/platform/threedo/operafs.go:285:1
std::tuple<Slice<uint8_t>,Error> threedo_Volume_ReadFile(threedo_Volume* v,std::string path){
{
auto tmp159 = threedo_Volume_resolve(v,path);
threedo_Entry e = std::get<0>(tmp159);
Error err = std::get<1>(tmp159);
if (bool(err)) {
return {{},err};
}
if (e.IsDir) {
return {{},go_fmt_Errorf(std::string("threedo: %q is a directory",26),path)};
}
Slice<uint8_t> out = Slice<uint8_t>::make(cast<int64_t>(0ULL),e.Size);
for (int64_t got = cast<int64_t>(0ULL);(got < e.Size);got += threedo_userSize){
auto tmp160 = threedo_Volume_block(v,cast<int64_t>((e.Block + divi<int64_t>(got,threedo_userSize))));
Slice<uint8_t> b = std::get<0>(tmp160);
Error err = std::get<1>(tmp160);
if (bool(err)) {
return {{},err};
}
int64_t n = cast<int64_t>((e.Size - got));
if ((n > threedo_userSize)) {
n = threedo_userSize;
}
out = append(out,sub(b,0,n));
}
return {out,{}};
}
}
// tools/platform/threedo/operamath.go:32:1
bool threedo_Machine_mathFolioSWI(threedo_Machine* m,arm60_CPU* c,uint32_t swi){
{
{
switch(swi){
case threedo_swiMulVec3Mat33:{
threedo_Machine_mulVec3Mat33(m,arm60_CPU_Reg(c,cast<uint32_t>(0ULL)),arm60_CPU_Reg(c,cast<uint32_t>(1ULL)),arm60_CPU_Reg(c,cast<uint32_t>(2ULL)));
break;}
case threedo_swiMulManyVec3Mat33:{
auto tmp161 = std::make_tuple(arm60_CPU_Reg(c,cast<uint32_t>(0ULL)),arm60_CPU_Reg(c,cast<uint32_t>(1ULL)),arm60_CPU_Reg(c,cast<uint32_t>(2ULL)),arm60_CPU_Reg(c,cast<uint32_t>(3ULL)));
uint32_t dest = std::get<0>(tmp161);
uint32_t src = std::get<1>(tmp161);
uint32_t mat = std::get<2>(tmp161);
uint32_t count = std::get<3>(tmp161);
for (uint32_t i = cast<uint32_t>(cast<uint32_t>(0ULL));(i < count);i++){
threedo_Machine_mulVec3Mat33(m,cast<uint32_t>((dest + cast<uint32_t>((i * cast<uint32_t>(12ULL))))),cast<uint32_t>((src + cast<uint32_t>((i * cast<uint32_t>(12ULL))))),mat);
}
break;}
case threedo_swiMulMat33Mat33:{
threedo_Machine_mulMat33Mat33(m,arm60_CPU_Reg(c,cast<uint32_t>(0ULL)),arm60_CPU_Reg(c,cast<uint32_t>(1ULL)),arm60_CPU_Reg(c,cast<uint32_t>(2ULL)));
break;}
case threedo_swiDot3:{
auto tmp162 = std::make_tuple(arm60_CPU_Reg(c,cast<uint32_t>(0ULL)),arm60_CPU_Reg(c,cast<uint32_t>(1ULL)));
uint32_t v1 = std::get<0>(tmp162);
uint32_t v2 = std::get<1>(tmp162);
int64_t acc={};
for (uint32_t i = cast<uint32_t>(cast<uint32_t>(0ULL));(i < cast<uint32_t>(3ULL));i++){
acc += cast<int64_t>((cast<int64_t>(cast<int32_t>(threedo_Machine_read32(m,cast<uint32_t>((v1 + cast<uint32_t>((i * cast<uint32_t>(4ULL)))))))) * cast<int64_t>(cast<int32_t>(threedo_Machine_read32(m,cast<uint32_t>((v2 + cast<uint32_t>((i * cast<uint32_t>(4ULL))))))))));
}
arm60_CPU_SetReg(c,cast<uint32_t>(0ULL),cast<uint32_t>(shr<int64_t>(acc,cast<int64_t>(16ULL))));
break;}
case threedo_swiCross3:{
threedo_Machine_cross3(m,arm60_CPU_Reg(c,cast<uint32_t>(0ULL)),arm60_CPU_Reg(c,cast<uint32_t>(1ULL)),arm60_CPU_Reg(c,cast<uint32_t>(2ULL)));
break;}
default:{
return false;
break;}
}}
return true;
}
}
// tools/platform/threedo/operamath.go:59:1
int32_t threedo_fmul16(int32_t a,int32_t b){
{
return cast<int32_t>(shr<int64_t>((cast<int64_t>((cast<int64_t>(a) * cast<int64_t>(b)))),cast<int64_t>(16ULL)));
}
}
// tools/platform/threedo/operamath.go:62:1
std::array<int32_t,3> threedo_Machine_readVec3(threedo_Machine* m,uint32_t p){
{
return std::array<int32_t,3>{cast<int32_t>(threedo_Machine_read32(m,p)),cast<int32_t>(threedo_Machine_read32(m,cast<uint32_t>((p + cast<uint32_t>(4ULL))))),cast<int32_t>(threedo_Machine_read32(m,cast<uint32_t>((p + cast<uint32_t>(8ULL)))))};
}
}
// tools/platform/threedo/operamath.go:69:1
void threedo_Machine_mulVec3Mat33(threedo_Machine* m,uint32_t dest,uint32_t vec,uint32_t mat){
{
std::array<int32_t,3> v = threedo_Machine_readVec3(m,vec);
std::array<int64_t,3> out={};
for (uint32_t i = cast<uint32_t>(cast<uint32_t>(0ULL));(i < cast<uint32_t>(3ULL));i++){
for (uint32_t j = cast<uint32_t>(cast<uint32_t>(0ULL));(j < cast<uint32_t>(3ULL));j++){
out[j] += cast<int64_t>((cast<int64_t>(v[i]) * cast<int64_t>(cast<int32_t>(threedo_Machine_read32(m,cast<uint32_t>((mat + cast<uint32_t>(((cast<uint32_t>((cast<uint32_t>((i * cast<uint32_t>(3ULL))) + j))) * cast<uint32_t>(4ULL))))))))));
}
}
for (uint32_t j = cast<uint32_t>(cast<uint32_t>(0ULL));(j < cast<uint32_t>(3ULL));j++){
threedo_Machine_write32(m,cast<uint32_t>((dest + cast<uint32_t>((j * cast<uint32_t>(4ULL))))),cast<uint32_t>(shr<int64_t>(out[j],cast<int64_t>(16ULL))));
}
}
}
// tools/platform/threedo/operamath.go:84:1
void threedo_Machine_mulMat33Mat33(threedo_Machine* m,uint32_t dest,uint32_t src1,uint32_t src2){
{
std::array<std::array<int32_t,3>,3> a={};
std::array<std::array<int32_t,3>,3> b={};
for (uint32_t i = cast<uint32_t>(cast<uint32_t>(0ULL));(i < cast<uint32_t>(3ULL));i++){
for (uint32_t j = cast<uint32_t>(cast<uint32_t>(0ULL));(j < cast<uint32_t>(3ULL));j++){
a[i][j] = cast<int32_t>(threedo_Machine_read32(m,cast<uint32_t>((src1 + cast<uint32_t>(((cast<uint32_t>((cast<uint32_t>((i * cast<uint32_t>(3ULL))) + j))) * cast<uint32_t>(4ULL)))))));
b[i][j] = cast<int32_t>(threedo_Machine_read32(m,cast<uint32_t>((src2 + cast<uint32_t>(((cast<uint32_t>((cast<uint32_t>((i * cast<uint32_t>(3ULL))) + j))) * cast<uint32_t>(4ULL)))))));
}
}
std::array<std::array<int32_t,3>,3> out={};
for (int64_t i = cast<int64_t>(0ULL);(i < cast<int64_t>(3ULL));i++){
for (int64_t j = cast<int64_t>(0ULL);(j < cast<int64_t>(3ULL));j++){
int64_t acc={};
for (int64_t k = cast<int64_t>(0ULL);(k < cast<int64_t>(3ULL));k++){
acc += cast<int64_t>((cast<int64_t>(a[i][k]) * cast<int64_t>(b[k][j])));
}
out[i][j] = cast<int32_t>(shr<int64_t>(acc,cast<int64_t>(16ULL)));
}
}
for (uint32_t i = cast<uint32_t>(cast<uint32_t>(0ULL));(i < cast<uint32_t>(3ULL));i++){
for (uint32_t j = cast<uint32_t>(cast<uint32_t>(0ULL));(j < cast<uint32_t>(3ULL));j++){
threedo_Machine_write32(m,cast<uint32_t>((dest + cast<uint32_t>(((cast<uint32_t>((cast<uint32_t>((i * cast<uint32_t>(3ULL))) + j))) * cast<uint32_t>(4ULL))))),cast<uint32_t>(out[i][j]));
}
}
}
}
// tools/platform/threedo/operamath.go:124:1
void threedo_Machine_serviceMathFolio(threedo_Machine* m,uint32_t foff){
{
arm60_CPU* c = m->CPU;
auto tmp163 = std::make_tuple(arm60_CPU_Reg(c,cast<uint32_t>(0ULL)),arm60_CPU_Reg(c,cast<uint32_t>(1ULL)),arm60_CPU_Reg(c,cast<uint32_t>(2ULL)));
uint32_t r0 = std::get<0>(tmp163);
uint32_t r1 = std::get<1>(tmp163);
uint32_t r2 = std::get<2>(tmp163);
{
switch(foff){
case cast<uint32_t>(4ULL):{
threedo_Machine_SetResultAndReturn(m,cast<uint32_t>(shr<uint64_t>((cast<uint64_t>((cast<uint64_t>(r0) * cast<uint64_t>(r1)))),cast<int64_t>(16ULL))));
break;}
case cast<uint32_t>(8ULL):{
threedo_Machine_SetResultAndReturn(m,cast<uint32_t>(shr<int64_t>((cast<int64_t>((cast<int64_t>(cast<int32_t>(r0)) * cast<int64_t>(cast<int32_t>(r1))))),cast<int64_t>(16ULL))));
break;}
case cast<uint32_t>(12ULL):{
auto tmp164 = threedo_divUF16(r0,r1);
uint32_t q = std::get<0>(tmp164);
threedo_Machine_SetResultAndReturn(m,q);
break;}
case cast<uint32_t>(16ULL):{
auto tmp165 = threedo_divUF16(r1,r2);
uint32_t q = std::get<0>(tmp165);
uint32_t rem = std::get<1>(tmp165);
threedo_Machine_write32(m,r0,rem);
threedo_Machine_SetResultAndReturn(m,q);
break;}
case cast<uint32_t>(20ULL):{
auto tmp166 = threedo_divSF16(cast<int32_t>(r0),cast<int32_t>(r1));
int32_t q = std::get<0>(tmp166);
threedo_Machine_SetResultAndReturn(m,cast<uint32_t>(q));
break;}
case cast<uint32_t>(24ULL):{
auto tmp167 = threedo_divSF16(cast<int32_t>(r1),cast<int32_t>(r2));
int32_t q = std::get<0>(tmp167);
uint32_t rem = std::get<1>(tmp167);
threedo_Machine_write32(m,r0,rem);
threedo_Machine_SetResultAndReturn(m,cast<uint32_t>(q));
break;}
case cast<uint32_t>(28ULL):{
auto tmp168 = threedo_divUF16(shl<uint32_t>(cast<int64_t>(1ULL),cast<int64_t>(16ULL)),r0);
uint32_t q = std::get<0>(tmp168);
threedo_Machine_SetResultAndReturn(m,q);
break;}
case cast<uint32_t>(32ULL):{
auto tmp169 = threedo_divSF16(shl<int32_t>(cast<int64_t>(1ULL),cast<int64_t>(16ULL)),cast<int32_t>(r0));
int32_t q = std::get<0>(tmp169);
threedo_Machine_SetResultAndReturn(m,cast<uint32_t>(q));
break;}
default:{
threedo_Machine_note(m,go_fmt_Sprintf(std::string("MathFolio[-0x%X] stub (r0=0x%08X r1=0x%08X r2=0x%08X)",53),foff,r0,r1,r2));
threedo_Machine_SetResultAndReturn(m,cast<uint32_t>(0ULL));
break;}
}}
}
}
// tools/platform/threedo/operamath.go:160:1
std::tuple<uint32_t,uint32_t> threedo_divUF16(uint32_t d1,uint32_t d2){
uint32_t q{};
uint32_t rem{};
{
if ((d2 == cast<uint32_t>(0ULL))) {
return {cast<uint32_t>(4294967295ULL),cast<uint32_t>(4294967295ULL)};
}
uint64_t num = shl<uint64_t>(cast<uint64_t>(d1),cast<int64_t>(16ULL));
uint64_t quot = divi<uint64_t>(num,cast<uint64_t>(d2));
if ((quot > cast<uint64_t>(4294967295ULL))) {
return {cast<uint32_t>(4294967295ULL),cast<uint32_t>(4294967295ULL)};
}
return {cast<uint32_t>(quot),cast<uint32_t>(divi<uint64_t>((shl<uint64_t>((modi<uint64_t>(num,cast<uint64_t>(d2))),cast<int64_t>(16ULL))),cast<uint64_t>(d2)))};
}
}
// tools/platform/threedo/operamath.go:173:1
std::tuple<int32_t,uint32_t> threedo_divSF16(int32_t d1,int32_t d2){
int32_t q{};
uint32_t rem{};
{
if ((d2 == cast<int32_t>(0ULL))) {
return {cast<int32_t>(2147483647ULL),cast<uint32_t>(4294967295ULL)};
}
int64_t num = shl<int64_t>(cast<int64_t>(d1),cast<int64_t>(16ULL));
int64_t quot = divi<int64_t>(num,cast<int64_t>(d2));
if (((quot > cast<int64_t>(2147483647ULL)) || (quot < cast<int64_t>(-cast<int64_t>(2147483648ULL))))) {
return {cast<int32_t>(2147483647ULL),cast<uint32_t>(4294967295ULL)};
}
int64_t r = modi<int64_t>(num,cast<int64_t>(d2));
if ((r < cast<int64_t>(0ULL))) {
r = cast<int64_t>(-r);
}
int64_t d = cast<int64_t>(d2);
if ((d < cast<int64_t>(0ULL))) {
d = cast<int64_t>(-d);
}
return {cast<int32_t>(quot),cast<uint32_t>(divi<int64_t>((shl<int64_t>(r,cast<int64_t>(16ULL))),d))};
}
}
// tools/platform/threedo/operamath.go:194:1
void threedo_Machine_cross3(threedo_Machine* m,uint32_t dest,uint32_t v1p,uint32_t v2p){
{
auto tmp170 = std::make_tuple(threedo_Machine_readVec3(m,v1p),threedo_Machine_readVec3(m,v2p));
std::array<int32_t,3> a = std::get<0>(tmp170);
std::array<int32_t,3> b = std::get<1>(tmp170);
std::array<int32_t,3> out = std::array<int32_t,3>{cast<int32_t>((threedo_fmul16(a[cast<int64_t>(1ULL)],b[cast<int64_t>(2ULL)]) - threedo_fmul16(a[cast<int64_t>(2ULL)],b[cast<int64_t>(1ULL)]))),cast<int32_t>((threedo_fmul16(a[cast<int64_t>(2ULL)],b[cast<int64_t>(0ULL)]) - threedo_fmul16(a[cast<int64_t>(0ULL)],b[cast<int64_t>(2ULL)]))),cast<int32_t>((threedo_fmul16(a[cast<int64_t>(0ULL)],b[cast<int64_t>(1ULL)]) - threedo_fmul16(a[cast<int64_t>(1ULL)],b[cast<int64_t>(0ULL)])))};
for (uint32_t j = cast<uint32_t>(cast<uint32_t>(0ULL));(j < cast<uint32_t>(3ULL));j++){
threedo_Machine_write32(m,cast<uint32_t>((dest + cast<uint32_t>((j * cast<uint32_t>(4ULL))))),cast<uint32_t>(out[j]));
}
}
}
// tools/platform/threedo/profile.go:96:1
time_Time threedo_Machine_profStart(threedo_Machine* m){
{
if ((!m->Profile)) {
return time_Time{};
}
return go_time_Now();
}
}
// tools/platform/threedo/profile.go:103:1
void threedo_Machine_profEnd(threedo_Machine* m,int64_t bucket,time_Time t){
{
if (time_Time_IsZero(t)) {
return ;
}
m->prof.ns[bucket] += cast<int64_t>(go_time_Since(t));
m->prof.count[bucket]++;
}
}
// tools/platform/threedo/profile.go:112:1
int64_t threedo_Machine_profGfxNs(threedo_Machine* m){
{
int64_t n={};
{auto&& tmp171 = threedo_gfxBuckets;
for(int64_t tmp172=0;tmp172<len(tmp171);++tmp172){
auto b=tmp171[tmp172];n += m->prof.ns[b];
}}
return n;
}
}
// tools/platform/threedo/profile.go:128:1
void threedo_Machine_profEndFolio(threedo_Machine* m,time_Time t,int64_t gfxBefore,int64_t gen){
{
if ((time_Time_IsZero(t) || (gen != m->prof.gen))) {
return ;
}
int64_t ns = cast<int64_t>((cast<int64_t>(go_time_Since(t)) - (cast<int64_t>((threedo_Machine_profGfxNs(m) - gfxBefore)))));
if ((ns < cast<int64_t>(0ULL))) {
ns = cast<int64_t>(0ULL);
}
m->prof.ns[threedo_bucketFolio] += ns;
m->prof.count[threedo_bucketFolio]++;
}
}
// tools/platform/threedo/profile.go:140:1
void threedo_Machine_profRunEnter(threedo_Machine* m){
{
if ((!m->Profile)) {
return ;
}
auto tmp173 = std::make_tuple(go_time_Now(),true);
m->prof.runStart = std::get<0>(tmp173);
m->prof.inRun = std::get<1>(tmp173);
}
}
// tools/platform/threedo/profile.go:147:1
void threedo_Machine_profRunExit(threedo_Machine* m){
{
if (((!m->Profile) || (!m->prof.inRun))) {
return ;
}
m->prof.frameNs += cast<int64_t>(go_time_Since(m->prof.runStart));
m->prof.inRun = false;
}
}
// tools/platform/threedo/profile.go:157:1
void threedo_Machine_profFrame(threedo_Machine* m){
{
if ((!m->Profile)) {
return ;
}
threedo_profState* p = (&m->prof);
int64_t total = p->frameNs;
if (p->inRun) {
total += cast<int64_t>(go_time_Since(p->runStart));
p->runStart = go_time_Now();
}
p->frameNs = cast<int64_t>(0ULL);
std::function<double(int64_t)> ms = [&](int64_t ns)->double{
return (cast<double>(ns) / 1e6);
}
;
Slice<threedo_ProfileBucket> buckets = Slice<threedo_ProfileBucket>::make(cast<int64_t>(0ULL),cast<int64_t>((threedo_numBuckets + cast<int64_t>(1ULL))));
int64_t summed={};
for (int64_t b = cast<int64_t>(0ULL);(b < threedo_numBuckets);b++){
summed += p->ns[b];
buckets = append(buckets,threedo_ProfileBucket{threedo_bucketNames[b],ms(p->ns[b]),cast<int64_t>(p->count[b])});
}
int64_t other = cast<int64_t>((total - summed));
if ((other < cast<int64_t>(0ULL))) {
other = cast<int64_t>(0ULL);
}
buckets = append(buckets,threedo_ProfileBucket{std::string("arm60 + rest (derived)",22),ms(other),cast<int64_t>(0ULL)});
threedo_celCounters now = m->celCnt;
threedo_celCounters d = threedo_celCounters{cast<int64_t>((now.cels - p->base.cels)),cast<int64_t>((now.pixels - p->base.pixels)),cast<int64_t>((now.chains - p->base.chains)),cast<int64_t>((now.clears - p->base.clears)),cast<int64_t>((now.folios - p->base.folios)),cast<int64_t>((now.swis - p->base.swis))};
p->last = threedo_FrameProfile{ms(total),buckets,Slice<threedo_ProfileCounter>{threedo_ProfileCounter{std::string("cels drawn",10),d.cels},threedo_ProfileCounter{std::string("pixels written",14),d.pixels},threedo_ProfileCounter{std::string("DrawCels chains",15),d.chains},threedo_ProfileCounter{std::string("flash clears",12),d.clears},threedo_ProfileCounter{std::string("folio calls",11),d.folios},threedo_ProfileCounter{std::string("SWIs",4),d.swis},threedo_ProfileCounter{std::string("arm60 instructions",18),cast<int64_t>(cast<uint64_t>((m->CPU->Instrs - p->baseInstr)))}}};
p->has = true;
auto tmp174 = std::make_tuple(std::array<int64_t,4>{},std::array<int64_t,4>{});
p->ns = std::get<0>(tmp174);
p->count = std::get<1>(tmp174);
auto tmp175 = std::make_tuple(now,m->CPU->Instrs);
p->base = std::get<0>(tmp175);
p->baseInstr = std::get<1>(tmp175);
p->gen++;
}
}
// tools/platform/threedo/profile.go:215:1
threedo_FrameProfile threedo_Machine_FrameProfile(threedo_Machine* m){
{
if ((!m->prof.has)) {
return threedo_FrameProfile{};
}
return m->prof.last;
}
}
// tools/platform/threedo/profile.go:224:1
void threedo_Machine_SetProfile(threedo_Machine* m,bool on){
{
m->Profile = on;
m->prof = threedo_profState{};
if (on) {
auto tmp176 = std::make_tuple(m->celCnt,m->CPU->Instrs);
m->prof.base = std::get<0>(tmp176);
m->prof.baseInstr = std::get<1>(tmp176);
}
}
}
// tools/platform/threedo/run.go:34:1
threedo_Result threedo_Machine_Run(threedo_Machine* m,uint64_t maxSteps){
{
threedo_Machine_profRunEnter(m);
auto tmp177=defer([&](){threedo_Machine_profRunExit(m);});
uint64_t steps={};
Map<uint32_t,bool> seen = Map<uint32_t,bool>{};
uint64_t sinceNew={};
constexpr int64_t noProgress=1000000ULL;
constexpr int64_t switchAt=20000ULL;
std::array<uint32_t,64> ring={};
int64_t ri={};
int64_t unstickTries={};
int64_t stallSwitches={};
for (;(steps < maxSteps);){
if (m->StopRequested) {
m->StopRequested = false;
return threedo_Result{steps,arm60_CPU_Reg(m->CPU,cast<uint32_t>(15ULL)),std::string("stop requested",14)};
}
if (((!m->PaceFields) && (modi<uint64_t>(steps,threedo_vblankPeriod) == cast<uint64_t>(0ULL)))) {
threedo_Machine_advanceVBlank(m,cast<uint32_t>(1ULL));
}
for (;((len(m->PadScript) > cast<int64_t>(0ULL)) && (steps >= m->PadScript[cast<int64_t>(0ULL)].AtStep));){
threedo_Machine_SendPadEvent(m,m->PadScript[cast<int64_t>(0ULL)].Buttons);
m->PadScript = sub(m->PadScript,cast<int64_t>(1ULL),len(m->PadScript));
}
uint32_t pc = arm60_CPU_Reg(m->CPU,cast<uint32_t>(15ULL));
if ((pc == threedo_taskExitTramp)) {
threedo_Machine_curTask(m)->state = threedo_stDone;
if ((!threedo_Machine_switchTask(m))) {
if ((!threedo_Machine_wakeByFieldTick(m))) {
return threedo_Result{steps,pc,go_fmt_Sprintf(std::string("all tasks finished (%d switches)",32),m->switches)};
}
}
continue;
}
if (((pc >= threedo_hleBase) && (pc < cast<uint32_t>((threedo_hleBase + threedo_hleSize))))) {
threedo_Machine_serviceKernelCall(m,pc);
steps++;
if (m->needSchedule) {
m->needSchedule = false;
if ((!threedo_Machine_switchTask(m))) {
if (threedo_Machine_wakeByFieldTick(m)) {
auto tmp178 = std::make_tuple(cast<uint64_t>(0ULL),cast<int64_t>(0ULL));
sinceNew = std::get<0>(tmp178);
stallSwitches = std::get<1>(tmp178);
}
else {
threedo_Machine_curTask(m)->state = threedo_stRunning;
}
}
}
continue;
}
if (m->Halted) {
return threedo_Result{steps,pc,m->HaltReason};
}
if (bool(m->OnStep)) {
m->OnStep(m,pc);
if (m->StopRequested) {
m->StopRequested = false;
return threedo_Result{steps,pc,std::string("stop requested",14)};
}
}
ring[cast<int64_t>((ri & cast<int64_t>(63ULL)))] = pc;
ri++;
if ((pc < threedo_dramSize)) {
if ((!get(seen,pc))) {
seen[pc] = true;
sinceNew = cast<uint64_t>(0ULL);
stallSwitches = cast<int64_t>(0ULL);
}
else {
sinceNew++;
if ((modi<uint64_t>(sinceNew,switchAt) == cast<uint64_t>(0ULL))) {
bool woke = false;
if (m->PaceFields) {
woke = threedo_Machine_fieldTick(m);
}
threedo_Machine_curTask(m)->state = threedo_stReady;
bool switched = threedo_Machine_switchTask(m);
if (woke) {
stallSwitches = cast<int64_t>(0ULL);
sinceNew = cast<uint64_t>(0ULL);
continue;
}
if (switched) {
stallSwitches++;
if ((stallSwitches > cast<int64_t>(((cast<int64_t>((cast<int64_t>((cast<int64_t>(32ULL) * len(m->tasks))) + cast<int64_t>(8ULL)))) * gmax(cast<int64_t>(1ULL),m->StallTolerance))))) {
return threedo_Result{steps,pc,go_fmt_Sprintf(std::string("deadlock: all %d tasks stalled near 0x%08X",42),len(m->tasks),pc)};
}
sinceNew = cast<uint64_t>(0ULL);
continue;
}
threedo_Machine_curTask(m)->state = threedo_stRunning;
if ((m->SpinBreak && threedo_Machine_isFlagSpin(m,sub(ring,0,len(ring))))) {
{
Slice<uint32_t> poked = threedo_Machine_breakSpin(m,sub(ring,0,len(ring)));
if (((len(poked) > cast<int64_t>(0ULL)) && (unstickTries < cast<int64_t>(2000ULL)))) {
unstickTries++;
m->SpinBreaks++;
sinceNew = cast<uint64_t>(0ULL);
}
}
}
}
if ((sinceNew > noProgress)) {
return threedo_Result{steps,pc,go_fmt_Sprintf(std::string("no forward progress (0x%08X, %d switches, %d tasks)",51),pc,m->switches,len(m->tasks))};
}
}
}
arm60_CPU_Step(m->CPU);
steps++;
if (m->needSchedule) {
m->needSchedule = false;
if ((!threedo_Machine_switchTask(m))) {
if (threedo_Machine_wakeByFieldTick(m)) {
auto tmp179 = std::make_tuple(cast<uint64_t>(0ULL),cast<int64_t>(0ULL));
sinceNew = std::get<0>(tmp179);
stallSwitches = std::get<1>(tmp179);
}
else {
threedo_Machine_curTask(m)->state = threedo_stRunning;
}
}
}
if (m->CPU->Halted) {
return threedo_Result{steps,arm60_CPU_CurPC(m->CPU),(std::string("cpu: ",5) + m->CPU->HaltReason)};
}
if (m->Halted) {
return threedo_Result{steps,arm60_CPU_CurPC(m->CPU),m->HaltReason};
}
}
return threedo_Result{steps,arm60_CPU_Reg(m->CPU,cast<uint32_t>(15ULL)),std::string("step budget reached",19)};
}
}
// tools/platform/threedo/run.go:202:1
void threedo_Machine_serviceKernelCall(threedo_Machine* m,uint32_t pc){
rrprof::Scope rrclock(1,"Portfolio HLE");
{
m->celCnt.folios++;
auto tmp180 = std::make_tuple(threedo_Machine_profStart(m),threedo_Machine_profGfxNs(m),m->prof.gen);
time_Time tf = std::get<0>(tmp180);
int64_t gfxBefore = std::get<1>(tmp180);
int64_t gen = std::get<2>(tmp180);
auto tmp181=defer([&](){threedo_Machine_profEndFolio(m,tf,gfxBefore,gen);});
uint32_t off = cast<uint32_t>((pc - threedo_hleBase));
m->KernelCalls = append(m->KernelCalls,threedo_KernelCall{off,cast<uint32_t>((arm60_CPU_Reg(m->CPU,cast<uint32_t>(14ULL)) - cast<uint32_t>(8ULL))),std::array<uint32_t,4>{arm60_CPU_Reg(m->CPU,cast<uint32_t>(0ULL)),arm60_CPU_Reg(m->CPU,cast<uint32_t>(1ULL)),arm60_CPU_Reg(m->CPU,cast<uint32_t>(2ULL)),arm60_CPU_Reg(m->CPU,cast<uint32_t>(3ULL))}});
if ((off >= threedo_hleAudioTag)) {
threedo_Machine_serviceAudioFolio(m,cast<uint32_t>((off - threedo_hleAudioTag)));
return ;
}
if ((off >= threedo_hleGfxTag)) {
threedo_Machine_serviceGraphicsFolio(m,cast<uint32_t>((off - threedo_hleGfxTag)));
return ;
}
if ((off >= threedo_hleOtherTag)) {
threedo_Machine_serviceOtherFolio(m,cast<uint32_t>((off - threedo_hleOtherTag)));
return ;
}
if ((off >= threedo_hleFileTag)) {
threedo_Machine_serviceFileFolio(m,cast<uint32_t>((off - threedo_hleFileTag)));
return ;
}
if ((off >= threedo_hleMathTag)) {
threedo_Machine_serviceMathFolio(m,cast<uint32_t>((off - threedo_hleMathTag)));
return ;
}
if ((!threedo_Machine_serviceFolio(m,off))) {
threedo_Machine_SetResultAndReturn(m,cast<uint32_t>(0ULL));
}
}
}
// tools/platform/threedo/run.go:247:1
void threedo_Machine_SetResultAndReturn(threedo_Machine* m,uint32_t result){
{
arm60_CPU_SetReg(m->CPU,cast<uint32_t>(0ULL),result);
arm60_CPU_SetPC(m->CPU,arm60_CPU_Reg(m->CPU,cast<uint32_t>(14ULL)));
}
}
// tools/platform/threedo/run.go:253:1
std::string threedo_Machine_DisasmAt(threedo_Machine* m,uint32_t addr){
{
if ((cast<int64_t>((cast<int64_t>(addr) + cast<int64_t>(4ULL))) > len(m->dram))) {
return go_fmt_Sprintf(std::string("%08X  ????",10),addr);
}
arm60_Inst in = arm60_Decode(sub(m->dram,addr,len(m->dram)),addr);
return go_fmt_Sprintf(std::string("%08X  %s",8),addr,in.Text);
}
}
// tools/platform/threedo/task.go:52:1
void threedo_Machine_initTasks(threedo_Machine* m){
{
m->tasks = Slice<threedo_task*>{arenaNew(threedo_task{threedo_bootTaskNum,std::string("boot",4),{},threedo_stRunning,{},{},{},{}})};
m->cur = cast<int64_t>(0ULL);
}
}
// tools/platform/threedo/task.go:60:1
int32_t threedo_Machine_spawnTask(threedo_Machine* m,uint32_t tagList){
{
auto tmp182 = threedo_Machine_parseTaskTags(m,tagList);
uint32_t entry = std::get<0>(tmp182);
uint32_t sp = std::get<1>(tmp182);
uint32_t argc = std::get<2>(tmp182);
uint32_t argp = std::get<3>(tmp182);
m->nextItem++;
arm60_Context ctx={};
ctx.Mode = arm60_ModeSYS;
ctx.R[cast<int64_t>(15ULL)] = entry;
ctx.R[cast<int64_t>(13ULL)] = sp;
ctx.R[cast<int64_t>(14ULL)] = threedo_taskExitTramp;
ctx.R[cast<int64_t>(0ULL)] = argc;
ctx.R[cast<int64_t>(1ULL)] = argp;
threedo_task* t = arenaNew(threedo_task{m->nextItem,std::string("task",4),ctx,threedo_stReady,{},{},{},{}});
m->tasks = append(m->tasks,t);
m->items[t->num] = arenaNew(threedo_item{t->num,cast<uint32_t>(261ULL),{},{},{},{},{},{},{},{}});
threedo_Machine_note(m,go_fmt_Sprintf(std::string("spawnTask #%d entry=0x%08X sp=0x%08X argc=%d argp=0x%08X",56),t->num,entry,sp,argc,argp));
return t->num;
}
}
// tools/platform/threedo/task.go:78:1
Slice<std::string> threedo_Machine_TaskSummary(threedo_Machine* m){
{
Map<threedo_taskState,std::string> names = Map<threedo_taskState,std::string>{{threedo_stReady,std::string("ready",5)},{threedo_stRunning,std::string("RUNNING",7)},{threedo_stWaiting,std::string("waiting",7)},{threedo_stDone,std::string("done",4)}};
Slice<std::string> out={};
{auto&& tmp183 = m->tasks;
for(int64_t tmp184=0;tmp184<len(tmp183);++tmp184){
auto t=tmp183[tmp184];uint32_t pc = t->ctx.R[cast<int64_t>(15ULL)];
if ((t->state == threedo_stRunning)) {
pc = arm60_CPU_Reg(m->CPU,cast<uint32_t>(15ULL));
}
out = append(out,go_fmt_Sprintf(std::string("task #%d %-7s pc=0x%08X sig=0x%X wait=0x%X",42),t->num,get(names,t->state),pc,t->sig,t->wait));
}}
return out;
}
}
// tools/platform/threedo/task.go:92:1
std::tuple<uint32_t,uint32_t,uint32_t,uint32_t> threedo_Machine_parseTaskTags(threedo_Machine* m,uint32_t p){
uint32_t entry{};
uint32_t sp{};
uint32_t argc{};
uint32_t argp{};
{
for (int64_t i = cast<int64_t>(0ULL);((i < cast<int64_t>(64ULL)) && (p != cast<uint32_t>(0ULL)));i++){
uint32_t tag = threedo_Machine_readWord(m,p);
uint32_t arg = threedo_Machine_readWord(m,cast<uint32_t>((p + cast<uint32_t>(4ULL))));
{
switch(tag){
case cast<uint32_t>(0ULL):{
return {entry,sp,argc,argp};
break;}
case cast<uint32_t>(10ULL):{
entry = arg;
break;}
case cast<uint32_t>(15ULL):{
sp = arg;
break;}
case cast<uint32_t>(13ULL):{
argc = arg;
break;}
case cast<uint32_t>(14ULL):{
argp = arg;
break;}
}}
p += cast<uint32_t>(8ULL);
}
return {entry,sp,argc,argp};
}
}
// tools/platform/threedo/task.go:114:1
uint32_t threedo_Machine_readWord(threedo_Machine* m,uint32_t a){
{
return cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((shl<uint32_t>(cast<uint32_t>(threedo_Machine_Read(m,a)),cast<int64_t>(24ULL)) | shl<uint32_t>(cast<uint32_t>(threedo_Machine_Read(m,cast<uint32_t>((a + cast<uint32_t>(1ULL))))),cast<int64_t>(16ULL)))) | shl<uint32_t>(cast<uint32_t>(threedo_Machine_Read(m,cast<uint32_t>((a + cast<uint32_t>(2ULL))))),cast<int64_t>(8ULL)))) | cast<uint32_t>(threedo_Machine_Read(m,cast<uint32_t>((a + cast<uint32_t>(3ULL)))))));
}
}
// tools/platform/threedo/task.go:119:1
threedo_task* threedo_Machine_curTask(threedo_Machine* m){
{
return m->tasks[m->cur];
}
}
// tools/platform/threedo/task.go:122:1
int32_t threedo_Machine_CurrentTaskNum(threedo_Machine* m){
{
return threedo_Machine_curTask(m)->num;
}
}
// tools/platform/threedo/task.go:125:1
threedo_task* threedo_Machine_taskByNum(threedo_Machine* m,int32_t num){
{
{auto&& tmp185 = m->tasks;
for(int64_t tmp186=0;tmp186<len(tmp185);++tmp186){
auto t=tmp185[tmp186];if ((t->num == num)) {
return t;
}
}}
return {};
}
}
// tools/platform/threedo/task.go:138:1
bool threedo_Machine_switchTask(threedo_Machine* m){
{
arm60_Context saved = arm60_CPU_SaveContext(m->CPU);
int64_t n = len(m->tasks);
for (int64_t i = cast<int64_t>(1ULL);(i <= n);i++){
int64_t j = modi<int64_t>((cast<int64_t>((m->cur + i))),n);
if ((m->tasks[j]->state == threedo_stReady)) {
m->tasks[m->cur]->ctx = saved;
arm60_CPU_RestoreContext(m->CPU,m->tasks[j]->ctx);
m->tasks[j]->state = threedo_stRunning;
m->cur = j;
m->switches++;
threedo_Machine_writeWord(m,cast<uint32_t>((threedo_osCtxBase + threedo_osCtxItem)),cast<uint32_t>(m->tasks[j]->num));
return true;
}
}
return false;
}
}
// tools/platform/threedo/task.go:165:1
bool threedo_Machine_sendSignal(threedo_Machine* m,int32_t num,uint32_t sigs){
{
{auto&& tmp187 = m->tasks;
for(int64_t tmp188=0;tmp188<len(tmp187);++tmp188){
auto t=tmp187[tmp188];if ((t->num == num)) {
t->sig |= sigs;
if (((t->state == threedo_stWaiting) && (cast<uint32_t>((t->sig & t->wait)) != cast<uint32_t>(0ULL)))) {
uint32_t got = cast<uint32_t>((t->sig & t->wait));
t->sig &= ~(got);
if ((!t->folioWait)) {
t->ctx.R[cast<int64_t>(0ULL)] = got;
}
t->state = threedo_stReady;
return true;
}
return false;
}
}}
return false;
}
}
// tools/platform/threedo/task.go:193:1
void threedo_Machine_yieldTo(threedo_Machine* m,bool woke,int32_t num){
{
if ((woke && (num != threedo_Machine_curTask(m)->num))) {
threedo_Machine_curTask(m)->state = threedo_stReady;
m->needSchedule = true;
}
}
}
// tools/platform/threedo/wrap.go:26:1
std::tuple<Slice<threedo_Resource>,Error> threedo_ParseWrap(Slice<uint8_t> data){
{
if (((len(data) < cast<int64_t>(8ULL)) || (cast<std::string>(sub(data,cast<int64_t>(0ULL),cast<int64_t>(4ULL))) != std::string("wwww",4)))) {
return {{},go_fmt_Errorf(std::string("threedo: not a wwww container",29))};
}
Slice<threedo_Resource> out={};
Map<int64_t,bool> seen = Map<int64_t,bool>{};
std::function<void(int64_t,int64_t)> walk={};
walk = [&](int64_t off,int64_t depth)->void{
if (((((off < cast<int64_t>(0ULL)) || (cast<int64_t>((off + cast<int64_t>(8ULL))) > len(data))) || get(seen,off)) || (depth > cast<int64_t>(8ULL)))) {
return ;
}
seen[off] = true;
if ((cast<std::string>(sub(data,off,cast<int64_t>((off + cast<int64_t>(4ULL))))) != std::string("wwww",4))) {
out = append(out,threedo_Resource{off,threedo_kindOf(data,off),depth});
return ;
}
int64_t n = cast<int64_t>(threedo_be32(sub(data,cast<int64_t>((off + cast<int64_t>(4ULL))),len(data))));
if (((n < cast<int64_t>(0ULL)) || (n > cast<int64_t>(100000ULL)))) {
return ;
}
for (int64_t i = cast<int64_t>(0ULL);(i < n);i++){
int64_t p = cast<int64_t>((cast<int64_t>((off + cast<int64_t>(8ULL))) + cast<int64_t>((i * cast<int64_t>(4ULL)))));
if ((cast<int64_t>((p + cast<int64_t>(4ULL))) > len(data))) {
break;
}
{
int64_t rel = cast<int64_t>(threedo_be32(sub(data,p,len(data))));
if ((rel != cast<int64_t>(0ULL))) {
walk(cast<int64_t>((off + rel)),cast<int64_t>((depth + cast<int64_t>(1ULL))));
}
}
}
}
;
walk(cast<int64_t>(0ULL),cast<int64_t>(0ULL));
return {out,{}};
}
}
// tools/platform/threedo/wrap.go:63:1
std::string threedo_kindOf(Slice<uint8_t> data,int64_t off){
{
if ((cast<int64_t>((off + cast<int64_t>(4ULL))) > len(data))) {
return std::string("unknown",7);
}
{
auto tmp190=cast<std::string>(sub(data,off,cast<int64_t>((off + cast<int64_t>(4ULL)))));
if (tmp190==(std::string("CCB ",4))){
return std::string("cel",3);
}
else if (tmp190==(std::string("ORI3",4))){
return std::string("model",5);
}
else if (tmp190==(std::string("SHPM",4))){
return std::string("shape",5);
}
else {
return std::string("unknown",7);
}
}
tmp189:;
}
}
// tools/platform/threedo/wrap.go:80:1
Map<std::string,int64_t> threedo_Inventory(Slice<threedo_Resource> res){
{
Map<std::string,int64_t> m = Map<std::string,int64_t>{};
{auto&& tmp191 = res;
for(int64_t tmp192=0;tmp192<len(tmp191);++tmp192){
auto r=tmp191[tmp192];m[r.Kind]++;
}}
return m;
}
}
// tools/platform/threedo/wrap.go:100:1
std::tuple<threedo_WrapNode*,Error> threedo_ParseWrapTree(Slice<uint8_t> data){
{
if (((len(data) < cast<int64_t>(8ULL)) || (cast<std::string>(sub(data,cast<int64_t>(0ULL),cast<int64_t>(4ULL))) != std::string("wwww",4)))) {
return {{},go_fmt_Errorf(std::string("threedo: not a wwww container",29))};
}
Map<int64_t,bool> seen = Map<int64_t,bool>{};
std::function<threedo_WrapNode*(int64_t,int64_t)> walk={};
walk = [&](int64_t off,int64_t depth)->threedo_WrapNode*{
if (((((off < cast<int64_t>(0ULL)) || (cast<int64_t>((off + cast<int64_t>(8ULL))) > len(data))) || get(seen,off)) || (depth > cast<int64_t>(8ULL)))) {
return {};
}
if ((cast<std::string>(sub(data,off,cast<int64_t>((off + cast<int64_t>(4ULL))))) != std::string("wwww",4))) {
return arenaNew(threedo_WrapNode{off,threedo_kindOf(data,off),{}});
}
seen[off] = true;
int64_t n = cast<int64_t>(threedo_be32(sub(data,cast<int64_t>((off + cast<int64_t>(4ULL))),len(data))));
if (((n < cast<int64_t>(0ULL)) || (n > cast<int64_t>(100000ULL)))) {
return {};
}
threedo_WrapNode* node = arenaNew(threedo_WrapNode{off,std::string("wwww",4),{}});
for (int64_t i = cast<int64_t>(0ULL);(i < n);i++){
int64_t p = cast<int64_t>((cast<int64_t>((off + cast<int64_t>(8ULL))) + cast<int64_t>((i * cast<int64_t>(4ULL)))));
if ((cast<int64_t>((p + cast<int64_t>(4ULL))) > len(data))) {
break;
}
threedo_WrapNode* child={};
{
int64_t rel = cast<int64_t>(threedo_be32(sub(data,p,len(data))));
if ((rel != cast<int64_t>(0ULL))) {
child = walk(cast<int64_t>((off + rel)),cast<int64_t>((depth + cast<int64_t>(1ULL))));
}
}
node->Children = append(node->Children,child);
}
return node;
}
;
threedo_WrapNode* root = walk(cast<int64_t>(0ULL),cast<int64_t>(0ULL));
if ((!root)) {
return {{},go_fmt_Errorf(std::string("threedo: malformed wwww container",33))};
}
return {root,{}};
}
}
