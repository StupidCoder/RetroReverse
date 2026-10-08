#include "runtime.h"
struct arm_Inst;
struct arm_CPU;
struct arm_Banks;
struct arm_vfpState;
struct n3ds_PixelEvent;
struct n3ds_MemRegion;
struct n3ds_dspHLE;
struct n3ds_dspSource;
struct n3ds_dspBuffer;
struct n3ds_dspFilters;
struct n3ds_ExeFSFile;
struct n3ds_ExeFS;
struct n3ds_CodeSegInfo;
struct n3ds_ExHeader;
struct n3ds_fsFile;
struct n3ds_fsDirEntry;
struct n3ds_fsDir;
struct n3ds_GPU;
struct n3ds_lightState;
struct n3ds_lightSource;
struct n3ds_ScreenGeom;
struct n3ds_vsOut;
struct n3ds_loaderBuf;
struct n3ds_fbState;
struct n3ds_rasterTri;
struct n3ds_rstats;
struct n3ds_scrVert;
struct n3ds_shaderState;
struct n3ds_shSrc;
struct n3ds_shInst;
struct n3ds_rgba;
struct n3ds_tevOperand;
struct n3ds_tevStage;
struct n3ds_tevState;
struct n3ds_texKey;
struct n3ds_texImage;
struct n3ds_FBPresent;
struct n3ds_GXRecord;
struct n3ds_xferRecord;
struct n3ds_gxPendingCmd;
struct n3ds_ipcHeader;
struct n3ds_ipcCall;
struct n3ds_memRegion;
struct n3ds_Machine;
struct n3ds_kobject;
struct n3ds_aptParam;
struct n3ds_watch;
struct n3ds_svcEvent;
struct n3ds_MSBTMessage;
struct n3ds_region;
struct n3ds_NCCH;
struct n3ds_Partition;
struct n3ds_NCSD;
struct n3ds_PICAWrite;
struct n3ds_ProfileBucket;
struct n3ds_ProfileCounter;
struct n3ds_FrameProfile;
struct n3ds_profState;
struct n3ds_profCounters;
struct n3ds_ivfcLevel;
struct n3ds_RomFSFile;
struct n3ds_RomFS;
struct n3ds_ivfcSpan;
struct n3ds_thread;
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
struct Anon55;
struct Anon56;
struct Anon57;
struct Anon58;
struct Anon59;
struct Anon6;
struct Anon60;
struct Anon61;
struct Anon62;
struct Anon63;
struct Anon64;
struct Anon65;
struct Anon66;
struct Anon67;
struct Anon68;
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
n3ds_Machine* bus{};
n3ds_Machine* wide{};
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
struct n3ds_PixelEvent{
bool Drawn{};
bool ZReject{};
bool AlphaReject{};
bool StencilReject{};
uint8_t R{};
uint8_t G{};
uint8_t B{};
uint8_t A{};
};
struct n3ds_MemRegion{
std::string Name{};
uint32_t Base{};
uint32_t Size{};
};
using n3ds_dspSample=std::array<int16_t,2>;
struct n3ds_dspFilters{
bool SimpleEnabled{};
bool BiquadEnabled{};
int32_t SB0{};
int32_t SA1{};
n3ds_dspSample SY1{};
int16_t BA1{};
int16_t BA2{};
int16_t BB0{};
int16_t BB1{};
int16_t BB2{};
n3ds_dspSample BX1{};
n3ds_dspSample BX2{};
n3ds_dspSample BY1{};
n3ds_dspSample BY2{};
};
struct n3ds_dspBuffer{
uint32_t PhysAddr{};
uint32_t Length{};
uint8_t AdpcmPS{};
std::array<int16_t,2> AdpcmYn{};
bool AdpcmDirty{};
bool IsLooping{};
uint16_t BufferID{};
bool Stereo{};
uint8_t Format{};
bool FromQueue{};
uint32_t PlayPosition{};
bool HasPlayed{};
};
using n3ds_dspFrame=std::array<n3ds_dspSample,160>;
struct n3ds_dspSource{
bool Enabled{};
uint16_t SyncCount{};
float Rate{};
uint8_t Interp{};
uint8_t Format{};
bool Stereo{};
std::array<std::array<float,4>,3> Gain{};
n3ds_dspFilters Filters{};
std::array<int16_t,16> AdpcmCoeffs{};
int16_t AdpcmYn1{};
int16_t AdpcmYn2{};
Slice<n3ds_dspBuffer> Queue{};
Slice<n3ds_dspSample> CurBuf{};
uint32_t CurSample{};
uint32_t CurPhysAddr{};
uint16_t CurBufferID{};
uint16_t LastBufferID{};
bool BufferUpdate{};
n3ds_dspSample Xn1{};
n3ds_dspSample Xn2{};
uint64_t FPos{};
n3ds_dspFrame Frame{};
};
struct n3ds_dspHLE{
bool ComponentLoaded{};
uint32_t ComponentSize{};
uint32_t State{};
std::array<Slice<uint8_t>,8> Pipes{};
Map<uint32_t,uint32_t> IntEvents{};
uint32_t SemEvent{};
uint32_t Semaphore{};
uint32_t SemMask{};
uint64_t NextFrame{};
uint64_t Ticks{};
std::array<n3ds_dspSource,24> Sources{};
std::array<float,3> MixVolume{};
std::array<bool,2> AuxBusEnable{};
uint16_t OutputFormat{};
uint16_t ClippingMode{};
bool Headphones{};
};
using n3ds_dspQuadFrame=std::array<std::array<int32_t,4>,160>;
struct n3ds_ExeFSFile{
std::string Name{};
int64_t Offset{};
int64_t Size{};
std::array<uint8_t,32> Hash{};
};
struct n3ds_ExeFS{
Slice<uint8_t> raw{};
Slice<n3ds_ExeFSFile> Files{};
};
struct n3ds_CodeSegInfo{
uint32_t Address{};
uint32_t NumPages{};
uint32_t Size{};
};
struct n3ds_ExHeader{
std::string Title{};
uint8_t Flag{};
uint16_t RemasterVer{};
n3ds_CodeSegInfo Text{};
uint32_t StackSize{};
n3ds_CodeSegInfo ROData{};
n3ds_CodeSegInfo Data{};
uint32_t BSSSize{};
Slice<uint64_t> Dependencies{};
uint64_t SaveDataSize{};
uint64_t JumpID{};
Slice<uint8_t> ACI{};
};
struct n3ds_fsFile{
Slice<uint8_t> data{};
std::string path{};
std::string save{};
};
struct n3ds_fsDirEntry{
std::string name{};
bool isDir{};
int64_t size{};
};
struct n3ds_fsDir{
std::string path{};
Slice<n3ds_fsDirEntry> entries{};
int64_t cursor{};
};
struct n3ds_texKey{
uint32_t addr{};
uint32_t fmt{};
uint32_t w{};
uint32_t h{};
};
inline bool operator==(const n3ds_texKey&a,const n3ds_texKey&b){return a.addr==b.addr&&a.fmt==b.fmt&&a.w==b.w&&a.h==b.h;}
namespace std{template<>struct hash<n3ds_texKey>{size_t operator()(const n3ds_texKey&k)const{return k.addr^(size_t(k.fmt)<<8)^(size_t(k.w)<<16)^(size_t(k.h)<<24);}};}
struct n3ds_shSrc{
uint8_t bank{};
uint8_t reg{};
uint8_t idx{};
bool neg{};
std::array<uint8_t,4> sw{};
bool plain{};
};
struct n3ds_shInst{
uint8_t kind{};
uint8_t op{};
std::array<n3ds_shSrc,3> src{};
bool dstTmp{};
uint8_t dst{};
std::array<bool,4> mask{};
bool maskAll{};
uint8_t cmpX{};
uint8_t cmpY{};
};
struct n3ds_PICAWrite{
uint32_t Off{};
uint16_t Reg{};
uint8_t Mask{};
uint32_t Value{};
bool Burst{};
};
struct n3ds_vsOut{
std::array<float,4> pos{};
std::array<float,4> color{};
std::array<std::array<float,2>,3> uv{};
float uv0w{};
std::array<float,4> quat{};
std::array<float,3> view{};
};
struct n3ds_scrVert{
float x{};
float y{};
float z{};
float iw{};
std::array<float,4> col{};
std::array<std::array<float,2>,3> uv{};
float uv0w{};
std::array<float,4> quat{};
std::array<float,3> view{};
};
struct n3ds_rasterTri{
n3ds_scrVert v0{};
n3ds_scrVert v1{};
n3ds_scrVert v2{};
float area{};
int64_t minX{};
int64_t maxX{};
int64_t minY{};
int64_t maxY{};
};
struct n3ds_loaderBuf{
uint32_t off{};
int64_t first{};
int64_t n{};
uint32_t stride{};
};
struct n3ds_GPU{
n3ds_Machine* m{};
std::array<uint32_t,768> Regs{};
std::array<uint32_t,4096> Code{};
std::array<uint32_t,128> Opdesc{};
std::array<std::array<float,4>,96> Float{};
uint32_t Bool{};
std::array<std::array<uint8_t,4>,4> Int{};
int64_t codeIdx{};
int64_t opdIdx{};
int64_t fltIdx{};
bool fltF32{};
Slice<uint32_t> fltBuf{};
int64_t fixedIdx{};
Slice<uint32_t> fixedBuf{};
std::array<std::array<float,4>,16> fixedVal{};
std::array<std::array<float,256>,24> LUT{};
std::array<std::array<float,256>,24> LUTDiff{};
std::array<bool,24> lutSet{};
uint32_t lutIdx{};
uint32_t lutType{};
int64_t gshCodeIdx{};
int64_t gshOpdIdx{};
int64_t gshFltIdx{};
bool gshFltF32{};
Slice<uint32_t> gshFltBuf{};
int64_t Draws{};
int64_t RejectedTris{};
int64_t ZeroAreaTris{};
int64_t CulledTris{};
int64_t DepthKilled{};
int64_t PixelsDrawn{};
int64_t ShadowWrites{};
int64_t ShadowSamples{};
int64_t ShadowOccluded{};
int64_t TraceDraws{};
int64_t TraceFrom{};
bool TraceUniforms{};
std::ostream* Census{};
Map<n3ds_texKey,n3ds_texImage*> texCache{};
uint32_t jumpAddr{};
uint32_t jumpSize{};
bool jumpPending{};
int64_t ListHops{};
std::array<n3ds_shInst,4096> dec{};
std::array<uint32_t,4096> decEpoch{};
uint32_t shEpoch{};
uint32_t decodedAll{};
Slice<uint8_t> cmdBuf{};
Slice<n3ds_PICAWrite> cmdWrite{};
Slice<n3ds_vsOut> outs{};
Slice<n3ds_rasterTri> tris{};
n3ds_workPool* workers{};
Slice<n3ds_loaderBuf> bufs{};
Slice<int64_t> comps{};
};
struct n3ds_lightSource{
int64_t index{};
std::array<float,3> specular0{};
std::array<float,3> specular1{};
std::array<float,3> diffuse{};
std::array<float,3> ambient{};
std::array<float,3> pos{};
std::array<float,3> spotDir{};
bool directional{};
bool twoSided{};
bool geo0{};
bool geo1{};
float attenBias{};
float attenScale{};
bool distAtten{};
bool spotAtten{};
bool shadowed{};
};
struct n3ds_lightState{
bool enabled{};
int64_t count{};
std::array<float,3> ambient{};
std::array<n3ds_lightSource,8> lights{};
uint32_t env{};
bool shadow{};
int64_t shadowSel{};
bool shadowInvert{};
bool shadowPrimary{};
bool shadowSecond{};
bool shadowAlpha{};
uint32_t bumpMode{};
int64_t bumpSel{};
bool noBumpRenorm{};
bool clampHighlight{};
bool primaryAlpha{};
bool secondAlpha{};
uint32_t lutIn{};
uint32_t lutAbs{};
uint32_t lutScale{};
bool noD0{};
bool noD1{};
bool noFR{};
bool noRR{};
bool noRG{};
bool noRB{};
};
struct n3ds_ScreenGeom{
uint32_t Src{};
uint32_t SrcW{};
uint32_t W{};
uint32_t H{};
bool Flip{};
bool Bottom{};
uint32_t Dst{};
uint32_t DstStride{};
uint32_t DstFmt{};
uint32_t DstBPP{};
};
struct n3ds_fbState{
uint32_t colorAddr{};
uint32_t depthAddr{};
uint32_t width{};
uint32_t height{};
bool depthTest{};
bool depthWr{};
uint32_t depthFunc{};
uint32_t colorMask{};
float vpHalfW{};
float vpHalfH{};
float depthScale{};
float depthOff{};
bool depthZBuffer{};
bool shadowMode{};
Slice<uint8_t> colorBuf{};
Slice<uint8_t> depthBuf{};
uint32_t colorOff{};
uint32_t depthOff32{};
};
struct n3ds_rstats{
int64_t pixelsDrawn{};
int64_t depthKilled{};
int64_t shadowWrites{};
int64_t shadowSamples{};
int64_t shadowOccluded{};
};
struct n3ds_shaderState{
n3ds_GPU* g{};
std::array<std::array<float,4>,16>* v{};
std::array<std::array<float,4>,16>* o{};
std::array<std::array<float,4>,16> r{};
std::array<int32_t,2> a0{};
int32_t aL{};
std::array<bool,2> cc{};
};
struct n3ds_rgba{
int32_t r{};
int32_t g{};
int32_t b{};
int32_t a{};
};
struct n3ds_tevOperand{
uint8_t src{};
uint8_t op{};
};
struct n3ds_tevStage{
std::array<n3ds_tevOperand,3> colr{};
std::array<n3ds_tevOperand,3> alph{};
uint8_t combC{};
uint8_t combA{};
uint8_t scaleC{};
uint8_t scaleA{};
n3ds_rgba konst{};
bool updC{};
bool updA{};
};
struct n3ds_tevState{
std::array<n3ds_tevStage,6> stages{};
n3ds_rgba bufColor{};
uint32_t texEnable{};
bool alphaTest{};
uint8_t alphaFunc{};
int32_t alphaRef{};
};
struct n3ds_texImage{
uint32_t w{};
uint32_t h{};
Slice<uint8_t> pix{};
};
struct n3ds_FBPresent{
uint32_t Active{};
uint32_t AddrLeft{};
uint32_t AddrRight{};
uint32_t Stride{};
uint32_t Format{};
uint32_t DispSelect{};
bool Valid{};
};
struct n3ds_GXRecord{
uint64_t Instr{};
std::array<uint32_t,8> Words{};
Slice<uint8_t> Buf{};
bool Chained{};
};
struct n3ds_xferRecord{
uint32_t dst{};
uint32_t w{};
uint32_t h{};
uint32_t format{};
uint32_t bpp{};
uint32_t stride{};
uint32_t src{};
uint32_t srcW{};
bool flip{};
};
struct n3ds_gxPendingCmd{
std::array<uint32_t,8> Words{};
uint64_t Deadline{};
};
struct n3ds_ipcHeader{
uint16_t Command{};
int64_t Normal{};
int64_t Translate{};
};
struct n3ds_ipcCall{
std::string service{};
uint16_t command{};
};
struct n3ds_memRegion{
std::string name{};
uint32_t base{};
Slice<uint8_t> data{};
};
struct n3ds_aptParam{
uint32_t Sender{};
uint32_t Command{};
uint32_t Handle{};
Slice<uint8_t> Data{};
};
struct n3ds_profCounters{
int64_t draws{};
int64_t frags{};
int64_t depthKilled{};
int64_t culled{};
int64_t rejected{};
int64_t shadowWrites{};
int64_t listHops{};
};
struct n3ds_ProfileBucket{
std::string Name{};
double Millis{};
int64_t Count{};
};
struct n3ds_ProfileCounter{
std::string Name{};
int64_t Value{};
};
struct n3ds_FrameProfile{
double TotalMs{};
Slice<n3ds_ProfileBucket> Buckets{};
Slice<n3ds_ProfileCounter> Counters{};
bool Drew{};
};
struct n3ds_profState{
std::array<int64_t,7> ns{};
std::array<int64_t,7> count{};
time_Time runStart{};
bool inRun{};
int64_t frameNs{};
int64_t idleSkips{};
n3ds_profCounters base{};
uint64_t baseInstr{};
n3ds_FrameProfile last{};
bool has{};
};
struct n3ds_watch{
uint32_t addr{};
uint32_t len{};
Map<uint32_t,uint32_t> last{};
Map<uint32_t,bool> seen{};
};
struct n3ds_svcEvent{
uint32_t PC{};
uint32_t Num{};
std::string Name{};
std::array<uint32_t,4> Args{};
};
struct n3ds_Machine{
arm_CPU* CPU{};
Slice<n3ds_memRegion*> regions{};
Slice<n3ds_memRegion*> pages{};
n3ds_memRegion* codeReg{};
n3ds_memRegion* stackReg{};
n3ds_memRegion* tlsReg{};
uint32_t entry{};
uint32_t heapPtr{};
uint32_t linearPtr{};
n3ds_memRegion* heapReg{};
n3ds_memRegion* linearReg{};
Map<uint32_t,n3ds_kobject*> handles{};
uint32_t nextHandle{};
Map<uint32_t,std::string> ports{};
Map<uint32_t,std::string> services{};
uint64_t tick{};
uint64_t instrs{};
Slice<n3ds_thread*> threads{};
n3ds_thread* curThread{};
uint32_t nextThread{};
uint32_t nextTLS{};
int64_t rrCursor{};
bool reschedule{};
bool stopped{};
n3ds_RomFS* romfs{};
Slice<uint8_t> romfsRaw{};
Map<uint32_t,n3ds_fsFile*> fsFiles{};
Map<uint32_t,n3ds_fsDir*> fsDirs{};
Map<uint32_t,uint32_t> fsArchives{};
Map<std::string,Slice<uint8_t>> saveFiles{};
bool saveFormatted{};
std::array<uint32_t,4> saveFormatInfo{};
n3ds_GPU* gpu{};
n3ds_dspHLE dsp{};
bool AudioCapture{};
Slice<int16_t> AudioPCM{};
bool MemTrace{};
bool DSPTrace{};
Slice<uint32_t> notifyWaiters{};
uint32_t aptNotifyEv{};
uint32_t aptResumeEv{};
bool aptWakePending{};
Slice<n3ds_aptParam> aptParams{};
Slice<n3ds_gxPendingCmd> gxPending{};
Slice<n3ds_ipcCall> ipcLog{};
uint32_t gspShared{};
uint32_t gspSharedAddr{};
uint32_t gspEvent{};
uint32_t hidShared{};
uint32_t hidSharedAddr{};
Slice<uint32_t> hidEvents{};
Map<uint32_t,int64_t> hidReadHist{};
Map<uint32_t,uint32_t> hidReadPC{};
bool HidTrace{};
uint32_t hidButtons{};
uint32_t hidPrevButtons{};
uint32_t hidRingIdx{};
int64_t HidPulse{};
uint16_t hidTouchX{};
uint16_t hidTouchY{};
bool hidTouchDown{};
Map<uint32_t,bool> arbTimedWarned{};
bool gspOverflowed{};
uint64_t nextFrameInstr{};
uint64_t vblankCount{};
int64_t framesSubmitted{};
int64_t framesSwapped{};
int64_t displayTransfers{};
n3ds_xferRecord lastXferTop{};
n3ds_xferRecord lastXferBottom{};
std::array<n3ds_FBPresent,2> screenFB{};
std::function<void(n3ds_PICAWrite)> OnPICACmd{};
std::function<void(uint32_t,uint32_t,n3ds_PixelEvent)> OnPixel{};
std::function<void(n3ds_Machine*)> OnFrame{};
std::function<void(std::string,n3ds_ScreenGeom)> OnPresent{};
bool StopRequested{};
uint32_t WatchLo{};
uint32_t WatchHi{};
std::function<void(uint32_t,uint32_t,uint32_t)> OnWrite{};
uint32_t RWatchLo{};
uint32_t RWatchHi{};
std::function<void(uint32_t,uint32_t,uint32_t)> OnRead{};
int64_t picaLimit{};
int64_t picaCount{};
bool SingleThreaded{};
bool Profile{};
n3ds_profState prof{};
bool GXCapture{};
Slice<n3ds_GXRecord> gxLog{};
bool Trace{};
int64_t traceN{};
int64_t traceMax{};
Map<uint32_t,bool> bps{};
Map<uint32_t,bool> logpcs{};
Map<uint32_t,bool> tracefroms{};
Slice<n3ds_watch> watches{};
Slice<n3ds_svcEvent> svcLog{};
Slice<uint8_t> debugOut{};
bool Verbose{};
uint64_t programID{};
};
struct n3ds_kobject{
std::string kind{};
std::string name{};
bool signal{};
bool manualReset{};
int32_t semCount{};
uint32_t mutexOwner{};
int64_t mutexDepth{};
Slice<uint32_t> waiters{};
uint32_t blockAddr{};
uint32_t blockSize{};
n3ds_thread* thread{};
};
struct n3ds_MSBTMessage{
int64_t Index{};
std::string Label{};
std::string Text{};
};
struct n3ds_region{
int64_t Offset{};
int64_t Size{};
};
struct n3ds_NCCH{
Slice<uint8_t> raw{};
int64_t ContentSize{};
uint64_t PartitionID{};
uint64_t ProgramID{};
std::string MakerCode{};
uint16_t Version{};
std::string ProductCode{};
int64_t MediaUnitSize{};
std::array<uint8_t,8> Flags{};
int64_t ExHeaderSize{};
n3ds_region PlainRegion{};
n3ds_region LogoRegion{};
n3ds_region ExeFSRegion{};
n3ds_region RomFSRegion{};
};
struct n3ds_Partition{
int64_t Index{};
int64_t Offset{};
int64_t Size{};
uint8_t FSType{};
uint8_t Crypt{};
uint64_t ID{};
};
struct n3ds_NCSD{
Slice<uint8_t> raw{};
int64_t ImageSize{};
uint64_t MediaID{};
int64_t MediaUnitSize{};
std::array<uint8_t,8> Flags{};
std::array<n3ds_Partition,8> Partitions{};
};
struct n3ds_ivfcLevel{
uint64_t LogicalOffset{};
uint64_t HashDataSize{};
uint32_t BlockSizeLog{};
};
struct n3ds_RomFSFile{
std::string Path{};
int64_t Offset{};
int64_t Size{};
};
struct n3ds_ivfcSpan{
int64_t Offset{};
int64_t Size{};
int64_t BlockSize{};
};
struct n3ds_RomFS{
Slice<uint8_t> raw{};
int64_t dataStart{};
Slice<uint8_t> MasterHash{};
std::array<n3ds_ivfcSpan,3> Levels{};
Slice<n3ds_RomFSFile> Files{};
Slice<std::string> Dirs{};
};
using n3ds_threadState=int64_t;
struct n3ds_thread{
uint32_t id{};
uint32_t handle{};
arm_CPU ctx{};
uint32_t tlsBase{};
uint32_t tpidr{};
int32_t priority{};
n3ds_threadState state{};
uint64_t wakeTick{};
uint64_t waitDeadline{};
bool waitAll{};
Slice<uint32_t> waitOn{};
uint32_t arbAddr{};
};
struct Anon0{uint32_t Addr{};int64_t Len{};std::string Mnem{};std::string Text{};arm_Flow Flow{};uint32_t Target{};bool HasTarget{};bool Thumb{};bool TargetThumb{};int64_t Cond{};};
struct Anon1{std::array<uint32_t,16> R{};bool N{};bool Z{};bool C{};bool V{};bool Q{};uint32_t GE{};bool Thumb{};bool BigEndian{};bool IRQDisable{};bool FIQDisable{};uint32_t Mode{};arm_Variant Arch{};arm_vfpState VFP{};bool exclValid{};uint32_t exclAddr{};std::array<uint32_t,6> bankR13{};std::array<uint32_t,6> bankR14{};std::array<uint32_t,6> bankSPSR{};std::array<uint32_t,5> fiqR8_12{};std::array<uint32_t,5> usrR8_12{};std::function<bool(arm_CPU*,uint32_t)> SWI{};std::function<void(arm_CPU*,bool,uint32_t,uint32_t,uint32_t,uint32_t,uint32_t,uint32_t*)> Coproc{};n3ds_Machine* bus{};n3ds_Machine* wide{};bool Halted{};std::string HaltReason{};uint64_t Instrs{};uint32_t cur{};bool branched{};};
struct Anon10{bool SimpleEnabled{};bool BiquadEnabled{};int32_t SB0{};int32_t SA1{};n3ds_dspSample SY1{};int16_t BA1{};int16_t BA2{};int16_t BB0{};int16_t BB1{};int16_t BB2{};n3ds_dspSample BX1{};n3ds_dspSample BX2{};n3ds_dspSample BY1{};n3ds_dspSample BY2{};};
struct Anon11{std::string Name{};int64_t Offset{};int64_t Size{};std::array<uint8_t,32> Hash{};};
struct Anon12{Slice<uint8_t> raw{};Slice<n3ds_ExeFSFile> Files{};};
struct Anon13{uint32_t Address{};uint32_t NumPages{};uint32_t Size{};};
struct Anon14{std::string Title{};uint8_t Flag{};uint16_t RemasterVer{};n3ds_CodeSegInfo Text{};uint32_t StackSize{};n3ds_CodeSegInfo ROData{};n3ds_CodeSegInfo Data{};uint32_t BSSSize{};Slice<uint64_t> Dependencies{};uint64_t SaveDataSize{};uint64_t JumpID{};Slice<uint8_t> ACI{};};
struct Anon15{std::string name{};n3ds_CodeSegInfo seg{};};
struct Anon16{Slice<uint8_t> data{};std::string path{};std::string save{};};
struct Anon17{std::string name{};bool isDir{};int64_t size{};};
struct Anon18{std::string path{};Slice<n3ds_fsDirEntry> entries{};int64_t cursor{};};
struct Anon19{n3ds_Machine* m{};std::array<uint32_t,768> Regs{};std::array<uint32_t,4096> Code{};std::array<uint32_t,128> Opdesc{};std::array<std::array<float,4>,96> Float{};uint32_t Bool{};std::array<std::array<uint8_t,4>,4> Int{};int64_t codeIdx{};int64_t opdIdx{};int64_t fltIdx{};bool fltF32{};Slice<uint32_t> fltBuf{};int64_t fixedIdx{};Slice<uint32_t> fixedBuf{};std::array<std::array<float,4>,16> fixedVal{};std::array<std::array<float,256>,24> LUT{};std::array<std::array<float,256>,24> LUTDiff{};std::array<bool,24> lutSet{};uint32_t lutIdx{};uint32_t lutType{};int64_t gshCodeIdx{};int64_t gshOpdIdx{};int64_t gshFltIdx{};bool gshFltF32{};Slice<uint32_t> gshFltBuf{};int64_t Draws{};int64_t RejectedTris{};int64_t ZeroAreaTris{};int64_t CulledTris{};int64_t DepthKilled{};int64_t PixelsDrawn{};int64_t ShadowWrites{};int64_t ShadowSamples{};int64_t ShadowOccluded{};int64_t TraceDraws{};int64_t TraceFrom{};bool TraceUniforms{};std::ostream* Census{};Map<n3ds_texKey,n3ds_texImage*> texCache{};uint32_t jumpAddr{};uint32_t jumpSize{};bool jumpPending{};int64_t ListHops{};std::array<n3ds_shInst,4096> dec{};std::array<uint32_t,4096> decEpoch{};uint32_t shEpoch{};uint32_t decodedAll{};Slice<uint8_t> cmdBuf{};Slice<n3ds_PICAWrite> cmdWrite{};Slice<n3ds_vsOut> outs{};Slice<n3ds_rasterTri> tris{};n3ds_workPool* workers{};Slice<n3ds_loaderBuf> bufs{};Slice<int64_t> comps{};};
struct Anon2{std::array<uint32_t,6> R13{};std::array<uint32_t,6> R14{};std::array<uint32_t,6> SPSR{};std::array<uint32_t,5> FIQR8_12{};std::array<uint32_t,5> USRR8_12{};};
struct Anon20{bool enabled{};int64_t count{};std::array<float,3> ambient{};std::array<n3ds_lightSource,8> lights{};uint32_t env{};bool shadow{};int64_t shadowSel{};bool shadowInvert{};bool shadowPrimary{};bool shadowSecond{};bool shadowAlpha{};uint32_t bumpMode{};int64_t bumpSel{};bool noBumpRenorm{};bool clampHighlight{};bool primaryAlpha{};bool secondAlpha{};uint32_t lutIn{};uint32_t lutAbs{};uint32_t lutScale{};bool noD0{};bool noD1{};bool noFR{};bool noRR{};bool noRG{};bool noRB{};};
struct Anon21{int64_t index{};std::array<float,3> specular0{};std::array<float,3> specular1{};std::array<float,3> diffuse{};std::array<float,3> ambient{};std::array<float,3> pos{};std::array<float,3> spotDir{};bool directional{};bool twoSided{};bool geo0{};bool geo1{};float attenBias{};float attenScale{};bool distAtten{};bool spotAtten{};bool shadowed{};};
struct Anon22{uint32_t Src{};uint32_t SrcW{};uint32_t W{};uint32_t H{};bool Flip{};bool Bottom{};uint32_t Dst{};uint32_t DstStride{};uint32_t DstFmt{};uint32_t DstBPP{};};
struct Anon23{std::array<float,4> pos{};std::array<float,4> color{};std::array<std::array<float,2>,3> uv{};float uv0w{};std::array<float,4> quat{};std::array<float,3> view{};};
struct Anon24{uint32_t off{};int64_t first{};int64_t n{};uint32_t stride{};};
struct Anon25{uint32_t colorAddr{};uint32_t depthAddr{};uint32_t width{};uint32_t height{};bool depthTest{};bool depthWr{};uint32_t depthFunc{};uint32_t colorMask{};float vpHalfW{};float vpHalfH{};float depthScale{};float depthOff{};bool depthZBuffer{};bool shadowMode{};Slice<uint8_t> colorBuf{};Slice<uint8_t> depthBuf{};uint32_t colorOff{};uint32_t depthOff32{};};
struct Anon26{n3ds_scrVert v0{};n3ds_scrVert v1{};n3ds_scrVert v2{};float area{};int64_t minX{};int64_t maxX{};int64_t minY{};int64_t maxY{};};
struct Anon27{int64_t pixelsDrawn{};int64_t depthKilled{};int64_t shadowWrites{};int64_t shadowSamples{};int64_t shadowOccluded{};};
struct Anon28{float x{};float y{};float z{};float iw{};std::array<float,4> col{};std::array<std::array<float,2>,3> uv{};float uv0w{};std::array<float,4> quat{};std::array<float,3> view{};};
struct Anon29{n3ds_GPU* g{};std::array<std::array<float,4>,16>* v{};std::array<std::array<float,4>,16>* o{};std::array<std::array<float,4>,16> r{};std::array<int32_t,2> a0{};int32_t aL{};std::array<bool,2> cc{};};
struct Anon3{uint32_t bit{};std::string name{};};
struct Anon30{uint8_t bank{};uint8_t reg{};uint8_t idx{};bool neg{};std::array<uint8_t,4> sw{};bool plain{};};
struct Anon31{uint8_t kind{};uint8_t op{};std::array<n3ds_shSrc,3> src{};bool dstTmp{};uint8_t dst{};std::array<bool,4> mask{};bool maskAll{};uint8_t cmpX{};uint8_t cmpY{};};
struct Anon32{int32_t r{};int32_t g{};int32_t b{};int32_t a{};};
struct Anon33{uint8_t src{};uint8_t op{};};
struct Anon34{std::array<n3ds_tevOperand,3> colr{};std::array<n3ds_tevOperand,3> alph{};uint8_t combC{};uint8_t combA{};uint8_t scaleC{};uint8_t scaleA{};n3ds_rgba konst{};bool updC{};bool updA{};};
struct Anon35{std::array<n3ds_tevStage,6> stages{};n3ds_rgba bufColor{};uint32_t texEnable{};bool alphaTest{};uint8_t alphaFunc{};int32_t alphaRef{};};
struct Anon36{uint32_t addr{};uint32_t fmt{};uint32_t w{};uint32_t h{};};
struct Anon37{uint32_t w{};uint32_t h{};Slice<uint8_t> pix{};};
struct Anon38{uint32_t Active{};uint32_t AddrLeft{};uint32_t AddrRight{};uint32_t Stride{};uint32_t Format{};uint32_t DispSelect{};bool Valid{};};
struct Anon39{uint64_t Instr{};std::array<uint32_t,8> Words{};Slice<uint8_t> Buf{};bool Chained{};};
struct Anon4{std::array<uint32_t,32> S{};uint32_t FPSCR{};uint32_t FPEXC{};};
struct Anon40{uint32_t dst{};uint32_t w{};uint32_t h{};uint32_t format{};uint32_t bpp{};uint32_t stride{};uint32_t src{};uint32_t srcW{};bool flip{};};
struct Anon41{std::array<uint32_t,8> Words{};uint64_t Deadline{};};
struct Anon42{uint16_t Command{};int64_t Normal{};int64_t Translate{};};
struct Anon43{std::string service{};uint16_t command{};};
struct Anon44{std::string name{};uint32_t base{};Slice<uint8_t> data{};};
struct Anon45{arm_CPU* CPU{};Slice<n3ds_memRegion*> regions{};Slice<n3ds_memRegion*> pages{};n3ds_memRegion* codeReg{};n3ds_memRegion* stackReg{};n3ds_memRegion* tlsReg{};uint32_t entry{};uint32_t heapPtr{};uint32_t linearPtr{};n3ds_memRegion* heapReg{};n3ds_memRegion* linearReg{};Map<uint32_t,n3ds_kobject*> handles{};uint32_t nextHandle{};Map<uint32_t,std::string> ports{};Map<uint32_t,std::string> services{};uint64_t tick{};uint64_t instrs{};Slice<n3ds_thread*> threads{};n3ds_thread* curThread{};uint32_t nextThread{};uint32_t nextTLS{};int64_t rrCursor{};bool reschedule{};bool stopped{};n3ds_RomFS* romfs{};Slice<uint8_t> romfsRaw{};Map<uint32_t,n3ds_fsFile*> fsFiles{};Map<uint32_t,n3ds_fsDir*> fsDirs{};Map<uint32_t,uint32_t> fsArchives{};Map<std::string,Slice<uint8_t>> saveFiles{};bool saveFormatted{};std::array<uint32_t,4> saveFormatInfo{};n3ds_GPU* gpu{};n3ds_dspHLE dsp{};bool AudioCapture{};Slice<int16_t> AudioPCM{};bool MemTrace{};bool DSPTrace{};Slice<uint32_t> notifyWaiters{};uint32_t aptNotifyEv{};uint32_t aptResumeEv{};bool aptWakePending{};Slice<n3ds_aptParam> aptParams{};Slice<n3ds_gxPendingCmd> gxPending{};Slice<n3ds_ipcCall> ipcLog{};uint32_t gspShared{};uint32_t gspSharedAddr{};uint32_t gspEvent{};uint32_t hidShared{};uint32_t hidSharedAddr{};Slice<uint32_t> hidEvents{};Map<uint32_t,int64_t> hidReadHist{};Map<uint32_t,uint32_t> hidReadPC{};bool HidTrace{};uint32_t hidButtons{};uint32_t hidPrevButtons{};uint32_t hidRingIdx{};int64_t HidPulse{};uint16_t hidTouchX{};uint16_t hidTouchY{};bool hidTouchDown{};Map<uint32_t,bool> arbTimedWarned{};bool gspOverflowed{};uint64_t nextFrameInstr{};uint64_t vblankCount{};int64_t framesSubmitted{};int64_t framesSwapped{};int64_t displayTransfers{};n3ds_xferRecord lastXferTop{};n3ds_xferRecord lastXferBottom{};std::array<n3ds_FBPresent,2> screenFB{};std::function<void(n3ds_PICAWrite)> OnPICACmd{};std::function<void(uint32_t,uint32_t,n3ds_PixelEvent)> OnPixel{};std::function<void(n3ds_Machine*)> OnFrame{};std::function<void(std::string,n3ds_ScreenGeom)> OnPresent{};bool StopRequested{};uint32_t WatchLo{};uint32_t WatchHi{};std::function<void(uint32_t,uint32_t,uint32_t)> OnWrite{};uint32_t RWatchLo{};uint32_t RWatchHi{};std::function<void(uint32_t,uint32_t,uint32_t)> OnRead{};int64_t picaLimit{};int64_t picaCount{};bool SingleThreaded{};bool Profile{};n3ds_profState prof{};bool GXCapture{};Slice<n3ds_GXRecord> gxLog{};bool Trace{};int64_t traceN{};int64_t traceMax{};Map<uint32_t,bool> bps{};Map<uint32_t,bool> logpcs{};Map<uint32_t,bool> tracefroms{};Slice<n3ds_watch> watches{};Slice<n3ds_svcEvent> svcLog{};Slice<uint8_t> debugOut{};bool Verbose{};uint64_t programID{};};
struct Anon46{std::string kind{};std::string name{};bool signal{};bool manualReset{};int32_t semCount{};uint32_t mutexOwner{};int64_t mutexDepth{};Slice<uint32_t> waiters{};uint32_t blockAddr{};uint32_t blockSize{};n3ds_thread* thread{};};
struct Anon47{uint32_t Sender{};uint32_t Command{};uint32_t Handle{};Slice<uint8_t> Data{};};
struct Anon48{uint32_t addr{};uint32_t len{};Map<uint32_t,uint32_t> last{};Map<uint32_t,bool> seen{};};
struct Anon49{uint32_t PC{};uint32_t Num{};std::string Name{};std::array<uint32_t,4> Args{};};
struct Anon5{bool Drawn{};bool ZReject{};bool AlphaReject{};bool StencilReject{};uint8_t R{};uint8_t G{};uint8_t B{};uint8_t A{};};
struct Anon50{int64_t Index{};std::string Label{};std::string Text{};};
struct Anon51{int64_t Offset{};int64_t Size{};};
struct Anon52{Slice<uint8_t> raw{};int64_t ContentSize{};uint64_t PartitionID{};uint64_t ProgramID{};std::string MakerCode{};uint16_t Version{};std::string ProductCode{};int64_t MediaUnitSize{};std::array<uint8_t,8> Flags{};int64_t ExHeaderSize{};n3ds_region PlainRegion{};n3ds_region LogoRegion{};n3ds_region ExeFSRegion{};n3ds_region RomFSRegion{};};
struct Anon53{std::string name{};n3ds_region reg{};};
struct Anon54{uint8_t bit{};std::string name{};};
struct Anon55{int64_t Index{};int64_t Offset{};int64_t Size{};uint8_t FSType{};uint8_t Crypt{};uint64_t ID{};};
struct Anon56{Slice<uint8_t> raw{};int64_t ImageSize{};uint64_t MediaID{};int64_t MediaUnitSize{};std::array<uint8_t,8> Flags{};std::array<n3ds_Partition,8> Partitions{};};
struct Anon57{uint32_t Off{};uint16_t Reg{};uint8_t Mask{};uint32_t Value{};bool Burst{};};
struct Anon58{std::string Name{};double Millis{};int64_t Count{};};
struct Anon59{std::string Name{};int64_t Value{};};
struct Anon6{std::string Name{};uint32_t Base{};uint32_t Size{};};
struct Anon60{double TotalMs{};Slice<n3ds_ProfileBucket> Buckets{};Slice<n3ds_ProfileCounter> Counters{};bool Drew{};};
struct Anon61{std::array<int64_t,7> ns{};std::array<int64_t,7> count{};time_Time runStart{};bool inRun{};int64_t frameNs{};int64_t idleSkips{};n3ds_profCounters base{};uint64_t baseInstr{};n3ds_FrameProfile last{};bool has{};};
struct Anon62{int64_t draws{};int64_t frags{};int64_t depthKilled{};int64_t culled{};int64_t rejected{};int64_t shadowWrites{};int64_t listHops{};};
struct Anon63{uint64_t LogicalOffset{};uint64_t HashDataSize{};uint32_t BlockSizeLog{};};
struct Anon64{std::string Path{};int64_t Offset{};int64_t Size{};};
struct Anon65{Slice<uint8_t> raw{};int64_t dataStart{};Slice<uint8_t> MasterHash{};std::array<n3ds_ivfcSpan,3> Levels{};Slice<n3ds_RomFSFile> Files{};Slice<std::string> Dirs{};};
struct Anon66{int64_t Offset{};int64_t Size{};int64_t BlockSize{};};
struct Anon67{std::string name{};uint32_t off{};uint32_t len{};};
struct Anon68{uint32_t id{};uint32_t handle{};arm_CPU ctx{};uint32_t tlsBase{};uint32_t tpidr{};int32_t priority{};n3ds_threadState state{};uint64_t wakeTick{};uint64_t waitDeadline{};bool waitAll{};Slice<uint32_t> waitOn{};uint32_t arbAddr{};};
struct Anon7{bool ComponentLoaded{};uint32_t ComponentSize{};uint32_t State{};std::array<Slice<uint8_t>,8> Pipes{};Map<uint32_t,uint32_t> IntEvents{};uint32_t SemEvent{};uint32_t Semaphore{};uint32_t SemMask{};uint64_t NextFrame{};uint64_t Ticks{};std::array<n3ds_dspSource,24> Sources{};std::array<float,3> MixVolume{};std::array<bool,2> AuxBusEnable{};uint16_t OutputFormat{};uint16_t ClippingMode{};bool Headphones{};};
struct Anon8{bool Enabled{};uint16_t SyncCount{};float Rate{};uint8_t Interp{};uint8_t Format{};bool Stereo{};std::array<std::array<float,4>,3> Gain{};n3ds_dspFilters Filters{};std::array<int16_t,16> AdpcmCoeffs{};int16_t AdpcmYn1{};int16_t AdpcmYn2{};Slice<n3ds_dspBuffer> Queue{};Slice<n3ds_dspSample> CurBuf{};uint32_t CurSample{};uint32_t CurPhysAddr{};uint16_t CurBufferID{};uint16_t LastBufferID{};bool BufferUpdate{};n3ds_dspSample Xn1{};n3ds_dspSample Xn2{};uint64_t FPos{};n3ds_dspFrame Frame{};};
struct Anon9{uint32_t PhysAddr{};uint32_t Length{};uint8_t AdpcmPS{};std::array<int16_t,2> AdpcmYn{};bool AdpcmDirty{};bool IsLooping{};uint16_t BufferID{};bool Stereo{};uint8_t Format{};bool FromQueue{};uint32_t PlayPosition{};bool HasPlayed{};};

#include "adapters-decl.h"
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
constexpr int64_t n3ds_blzFooterSize=8ULL;
constexpr int64_t n3ds_blzMinFooterPad=0ULL;
constexpr int64_t n3ds_blzMaxFooterPad=3ULL;
constexpr int64_t n3ds_blzThreshold=2ULL;
std::array<std::array<int32_t,2>,8> n3ds_etc1Modifiers=std::array<std::array<int32_t,2>,8>{std::array<int32_t,2>{cast<int32_t>(2ULL),cast<int32_t>(8ULL)},std::array<int32_t,2>{cast<int32_t>(5ULL),cast<int32_t>(17ULL)},std::array<int32_t,2>{cast<int32_t>(9ULL),cast<int32_t>(29ULL)},std::array<int32_t,2>{cast<int32_t>(13ULL),cast<int32_t>(42ULL)},std::array<int32_t,2>{cast<int32_t>(18ULL),cast<int32_t>(60ULL)},std::array<int32_t,2>{cast<int32_t>(24ULL),cast<int32_t>(80ULL)},std::array<int32_t,2>{cast<int32_t>(33ULL),cast<int32_t>(106ULL)},std::array<int32_t,2>{cast<int32_t>(47ULL),cast<int32_t>(183ULL)}};
Slice<std::string> n3ds_textureFormats=Slice<std::string>{std::string("rgba8",5),std::string("rgb8",4),std::string("rgba5551",8),std::string("rgb565",6),std::string("rgba4",5),std::string("la8",3),std::string("hilo8",5),std::string("l8",2),std::string("a8",2),std::string("la4",3),std::string("l4",2),std::string("a4",2),std::string("etc1",4),std::string("etc1a4",6)};
constexpr int64_t n3ds_dspRAMBase=535822336ULL;
constexpr int64_t n3ds_dspRAMSize=524288ULL;
constexpr int64_t n3ds_dspRegion0=536150016ULL;
constexpr int64_t n3ds_dspRegion1=536281088ULL;
constexpr int64_t n3ds_dspSharedSize=32768ULL;
constexpr int64_t n3ds_dspFrameTicks=1310720ULL;
constexpr int64_t n3ds_dspNumSources=24ULL;
constexpr int64_t n3ds_dspOffDSPStatus=2048ULL;
constexpr int64_t n3ds_dspOffDebug=2080ULL;
constexpr int64_t n3ds_dspOffFinalSamples=2688ULL;
constexpr int64_t n3ds_dspOffSourceStatus=3328ULL;
constexpr int64_t n3ds_dspOffCompressor=3616ULL;
constexpr int64_t n3ds_dspOffDSPConfig=10336ULL;
constexpr int64_t n3ds_dspOffIntermediate=10532ULL;
constexpr int64_t n3ds_dspOffSourceConfigs=15652ULL;
constexpr int64_t n3ds_dspOffAdpcmCoeffs=20260ULL;
constexpr int64_t n3ds_dspOffUnknown10=21028ULL;
constexpr int64_t n3ds_dspOffUnknown11=21540ULL;
constexpr int64_t n3ds_dspOffUnknown12=21924ULL;
constexpr int64_t n3ds_dspOffUnknown13=22692ULL;
constexpr int64_t n3ds_dspOffUnknown14=22712ULL;
constexpr int64_t n3ds_dspOffFrameCounter=32766ULL;
constexpr int64_t n3ds_dspStateOff=0ULL;
constexpr int64_t n3ds_dspStateOn=1ULL;
constexpr int64_t n3ds_dspStateSleeping=2ULL;
constexpr int64_t n3ds_dspPipeDebug=0ULL;
constexpr int64_t n3ds_dspPipeDMA=1ULL;
constexpr int64_t n3ds_dspPipeAudio=2ULL;
constexpr int64_t n3ds_dspPipeBinary=3ULL;
constexpr int64_t n3ds_dspIntZero=0ULL;
constexpr int64_t n3ds_dspIntOne=1ULL;
constexpr int64_t n3ds_dspIntPipe=2ULL;
constexpr int64_t n3ds_srcCfgDirty=0ULL;
constexpr int64_t n3ds_srcCfgGain=4ULL;
constexpr int64_t n3ds_srcCfgRate=52ULL;
constexpr int64_t n3ds_srcCfgInterp=56ULL;
constexpr int64_t n3ds_srcCfgFiltersOn=58ULL;
constexpr int64_t n3ds_srcCfgSimpleFilter=60ULL;
constexpr int64_t n3ds_srcCfgBiquadFilter=64ULL;
constexpr int64_t n3ds_srcCfgBuffersDirty=74ULL;
constexpr int64_t n3ds_srcCfgBuffers=76ULL;
constexpr int64_t n3ds_srcCfgEnable=160ULL;
constexpr int64_t n3ds_srcCfgSyncCount=162ULL;
constexpr int64_t n3ds_srcCfgPlayPosition=164ULL;
constexpr int64_t n3ds_srcCfgEmbPhysAddr=172ULL;
constexpr int64_t n3ds_srcCfgEmbLength=176ULL;
constexpr int64_t n3ds_srcCfgFlags1=180ULL;
constexpr int64_t n3ds_srcCfgEmbAdpcmPS=182ULL;
constexpr int64_t n3ds_srcCfgEmbAdpcmYn=184ULL;
constexpr int64_t n3ds_srcCfgFlags2=188ULL;
constexpr int64_t n3ds_srcCfgEmbBufferID=190ULL;
constexpr int64_t n3ds_srcBufPhysAddr=0ULL;
constexpr int64_t n3ds_srcBufLength=4ULL;
constexpr int64_t n3ds_srcBufAdpcmPS=8ULL;
constexpr int64_t n3ds_srcBufAdpcmYn=10ULL;
constexpr int64_t n3ds_srcBufAdpcmDirty=14ULL;
constexpr int64_t n3ds_srcBufIsLooping=15ULL;
constexpr int64_t n3ds_srcBufBufferID=16ULL;
constexpr int64_t n3ds_dirtyFormat=1ULL;
constexpr int64_t n3ds_dirtyMonoStereo=2ULL;
constexpr int64_t n3ds_dirtyAdpcmCoeffs=4ULL;
constexpr int64_t n3ds_dirtyPartialEmbedded=8ULL;
constexpr int64_t n3ds_dirtyPartialReset=16ULL;
constexpr int64_t n3ds_dirtyEnable=65536ULL;
constexpr int64_t n3ds_dirtyInterp=131072ULL;
constexpr int64_t n3ds_dirtyRate=262144ULL;
constexpr int64_t n3ds_dirtyBufferQueue=524288ULL;
constexpr int64_t n3ds_dirtyPlayPosition=2097152ULL;
constexpr int64_t n3ds_dirtyFiltersEnabled=4194304ULL;
constexpr int64_t n3ds_dirtySimpleFilter=8388608ULL;
constexpr int64_t n3ds_dirtyBiquadFilter=16777216ULL;
constexpr int64_t n3ds_dirtyGain0=33554432ULL;
constexpr int64_t n3ds_dirtyGain1=67108864ULL;
constexpr int64_t n3ds_dirtyGain2=134217728ULL;
constexpr int64_t n3ds_dirtySyncCount=268435456ULL;
constexpr int64_t n3ds_dirtyReset=536870912ULL;
constexpr int64_t n3ds_dirtyEmbeddedBuffer=1073741824ULL;
constexpr int64_t n3ds_dspSamplesPerFrame=160ULL;
constexpr int64_t n3ds_dspFmtPCM8=0ULL;
constexpr int64_t n3ds_dspFmtPCM16=1ULL;
constexpr int64_t n3ds_dspFmtADPCM=2ULL;
constexpr int64_t n3ds_dspInterpPolyphase=0ULL;
constexpr int64_t n3ds_dspInterpLinear=1ULL;
constexpr int64_t n3ds_dspInterpNone=2ULL;
constexpr int64_t n3ds_cfgDirty=0ULL;
constexpr int64_t n3ds_cfgMasterVolume=4ULL;
constexpr int64_t n3ds_cfgAuxReturnVol=8ULL;
constexpr int64_t n3ds_cfgOutputFormat=22ULL;
constexpr int64_t n3ds_cfgClippingMode=24ULL;
constexpr int64_t n3ds_cfgHeadphones=26ULL;
constexpr int64_t n3ds_cfgAuxBusEnable=40ULL;
constexpr int64_t n3ds_cfgDirty2=192ULL;
constexpr int64_t n3ds_cfgDirtyAuxBus0=256ULL;
constexpr int64_t n3ds_cfgDirtyAuxBus1=512ULL;
constexpr int64_t n3ds_cfgDirtyMasterVol=65536ULL;
constexpr int64_t n3ds_cfgDirtyAuxReturn0=16777216ULL;
constexpr int64_t n3ds_cfgDirtyAuxReturn1=33554432ULL;
constexpr int64_t n3ds_cfgDirtyOutFormat=67108864ULL;
constexpr int64_t n3ds_dspOutMono=0ULL;
constexpr int64_t n3ds_dspOutStereo=1ULL;
constexpr int64_t n3ds_dspOutSurround=2ULL;
constexpr int64_t n3ds_dspIntermediateMixSize=2560ULL;
constexpr int64_t n3ds_exefsMaxFiles=10ULL;
constexpr int64_t n3ds_exefsHeaderSize=512ULL;
constexpr int64_t n3ds_PageSize=4096ULL;
constexpr uint32_t n3ds_resultFSNotFound=3363849336ULL;
constexpr uint32_t n3ds_resultFSSaveNotFound=3363849556ULL;
constexpr int64_t n3ds_archiveRomFS=3ULL;
constexpr int64_t n3ds_regViewportWidth=65ULL;
constexpr int64_t n3ds_regViewportHeight=67ULL;
constexpr int64_t n3ds_regViewportXY=104ULL;
constexpr int64_t n3ds_regDepthMapScale=77ULL;
constexpr int64_t n3ds_regDepthMapOffset=78ULL;
constexpr int64_t n3ds_regDepthMapMode=109ULL;
constexpr int64_t n3ds_regShOutmapTotal=79ULL;
constexpr int64_t n3ds_regShOutmapO0=80ULL;
constexpr int64_t n3ds_regTexUnitConfig=128ULL;
constexpr int64_t n3ds_regTex0Param=131ULL;
constexpr int64_t n3ds_regShadowTex=139ULL;
constexpr int64_t n3ds_regLightingEnable=143ULL;
constexpr int64_t n3ds_regBlendConfig=256ULL;
constexpr int64_t n3ds_regShadowDensity=304ULL;
constexpr int64_t n3ds_regDepthColorMask=263ULL;
constexpr int64_t n3ds_regColorbufRead=274ULL;
constexpr int64_t n3ds_regColorbufWrite=275ULL;
constexpr int64_t n3ds_regDepthbufRead=276ULL;
constexpr int64_t n3ds_regDepthbufWrite=277ULL;
constexpr int64_t n3ds_regDepthbufFormat=278ULL;
constexpr int64_t n3ds_regColorbufFormat=279ULL;
constexpr int64_t n3ds_regDepthbufLoc=284ULL;
constexpr int64_t n3ds_regColorbufLoc=285ULL;
constexpr int64_t n3ds_regFramebufDim=286ULL;
constexpr int64_t n3ds_regAttrBase=512ULL;
constexpr int64_t n3ds_regAttrFmtLow=513ULL;
constexpr int64_t n3ds_regAttrFmtHigh=514ULL;
constexpr int64_t n3ds_regIndexConfig=551ULL;
constexpr int64_t n3ds_regNumVertices=552ULL;
constexpr int64_t n3ds_regGeoConfig=553ULL;
constexpr int64_t n3ds_regVertexOff=554ULL;
constexpr int64_t n3ds_regDrawArrays=558ULL;
constexpr int64_t n3ds_regDrawElems=559ULL;
constexpr int64_t n3ds_regFixedIndex=562ULL;
constexpr int64_t n3ds_regFixedData0=563ULL;
constexpr int64_t n3ds_regCmdBufSize0=568ULL;
constexpr int64_t n3ds_regCmdBufAddr0=570ULL;
constexpr int64_t n3ds_regCmdBufJump0=572ULL;
constexpr int64_t n3ds_regPrimConfig=606ULL;
constexpr int64_t n3ds_regVshBool=688ULL;
constexpr int64_t n3ds_regVshInt0=689ULL;
constexpr int64_t n3ds_regVshMaxInput=697ULL;
constexpr int64_t n3ds_regVshEntry=698ULL;
constexpr int64_t n3ds_regVshAttrPermL=699ULL;
constexpr int64_t n3ds_regVshAttrPermH=700ULL;
constexpr int64_t n3ds_regVshFloatCfg=704ULL;
constexpr int64_t n3ds_regVshFloatData=705ULL;
constexpr int64_t n3ds_regVshCodeIdx=715ULL;
constexpr int64_t n3ds_regVshCodeData=716ULL;
constexpr int64_t n3ds_regVshOpdescIdx=725ULL;
constexpr int64_t n3ds_regVshOpdescDat=726ULL;
constexpr int64_t n3ds_maxCmdBufHops=4096ULL;
constexpr int64_t n3ds_regLightBase=320ULL;
constexpr int64_t n3ds_regLightStep=16ULL;
constexpr int64_t n3ds_lightSpecular0=0ULL;
constexpr int64_t n3ds_lightSpecular1=1ULL;
constexpr int64_t n3ds_lightDiffuse=2ULL;
constexpr int64_t n3ds_lightAmbient=3ULL;
constexpr int64_t n3ds_lightPosXY=4ULL;
constexpr int64_t n3ds_lightPosZ=5ULL;
constexpr int64_t n3ds_lightSpotXY=6ULL;
constexpr int64_t n3ds_lightSpotZ=7ULL;
constexpr int64_t n3ds_lightConfig=9ULL;
constexpr int64_t n3ds_lightAttenBias=10ULL;
constexpr int64_t n3ds_lightAttenScl=11ULL;
constexpr int64_t n3ds_regLightAmbient=448ULL;
constexpr int64_t n3ds_regLightNumLight=450ULL;
constexpr int64_t n3ds_regLightConfig0=451ULL;
constexpr int64_t n3ds_regLightConfig1=452ULL;
constexpr int64_t n3ds_regLightLUTIndex=453ULL;
constexpr int64_t n3ds_regLightLUTData=456ULL;
constexpr int64_t n3ds_regLightLUTAbs=464ULL;
constexpr int64_t n3ds_regLightLUTSel=465ULL;
constexpr int64_t n3ds_regLightLUTScale=466ULL;
constexpr int64_t n3ds_regLightPermute=473ULL;
constexpr int64_t n3ds_lutD0=0ULL;
constexpr int64_t n3ds_lutD1=1ULL;
constexpr int64_t n3ds_lutFR=3ULL;
constexpr int64_t n3ds_lutRB=4ULL;
constexpr int64_t n3ds_lutRG=5ULL;
constexpr int64_t n3ds_lutRR=6ULL;
constexpr int64_t n3ds_lutSP0=8ULL;
constexpr int64_t n3ds_lutDA0=16ULL;
constexpr int64_t n3ds_numLUT=24ULL;
constexpr int64_t n3ds_lutInNH=0ULL;
constexpr int64_t n3ds_lutInVH=1ULL;
constexpr int64_t n3ds_lutInNV=2ULL;
constexpr int64_t n3ds_lutInLN=3ULL;
constexpr int64_t n3ds_lutInSP=4ULL;
constexpr int64_t n3ds_lutInCP=5ULL;
constexpr int64_t n3ds_shBad=0ULL;
constexpr int64_t n3ds_shNop=1ULL;
constexpr int64_t n3ds_shEnd=2ULL;
constexpr int64_t n3ds_shFlow=3ULL;
constexpr int64_t n3ds_shArith=4ULL;
constexpr int64_t n3ds_shMad=5ULL;
constexpr int64_t n3ds_shCmp=6ULL;
constexpr int64_t n3ds_shBankIn=0ULL;
constexpr int64_t n3ds_shBankTmp=1ULL;
constexpr int64_t n3ds_shBankUni=2ULL;
std::array<uint64_t,3> n3ds_shSrcShift=std::array<uint64_t,3>{cast<uint64_t>(4ULL),cast<uint64_t>(13ULL),cast<uint64_t>(22ULL)};
std::array<uint32_t,6> n3ds_tevStageBase=std::array<uint32_t,6>{cast<uint32_t>(192ULL),cast<uint32_t>(200ULL),cast<uint32_t>(208ULL),cast<uint32_t>(216ULL),cast<uint32_t>(240ULL),cast<uint32_t>(248ULL)};
constexpr int64_t n3ds_texType2D=0ULL;
constexpr int64_t n3ds_texTypeCube=1ULL;
constexpr int64_t n3ds_texTypeShadow2D=2ULL;
constexpr int64_t n3ds_texTypeProjection2D=3ULL;
constexpr int64_t n3ds_texTypeShadowCube=4ULL;
constexpr int64_t n3ds_texTypeDisabled=5ULL;
constexpr int64_t n3ds_gspSharedSize=4096ULL;
constexpr int64_t n3ds_hidSharedSize=4096ULL;
constexpr int64_t n3ds_vramPhysBase=402653184ULL;
constexpr int64_t n3ds_vramVirtBase=520093696ULL;
constexpr int64_t n3ds_vramSize=6291456ULL;
constexpr int64_t n3ds_stepsPerFrame=4468530ULL;
constexpr int64_t n3ds_gspIntPSC0=0ULL;
constexpr int64_t n3ds_gspIntPSC1=1ULL;
constexpr int64_t n3ds_gspIntVBlank0=2ULL;
constexpr int64_t n3ds_gspIntVBlank1=3ULL;
constexpr int64_t n3ds_gspIntPPF=4ULL;
constexpr int64_t n3ds_gspIntP3D=5ULL;
constexpr int64_t n3ds_gspIntDMA=6ULL;
constexpr int64_t n3ds_gxQueueOff=2048ULL;
constexpr int64_t n3ds_gxCmdStride=32ULL;
constexpr int64_t n3ds_gxMaxCmds=15ULL;
constexpr int64_t n3ds_gxHdrIndex=0ULL;
constexpr int64_t n3ds_gxHdrCount=1ULL;
constexpr int64_t n3ds_gxCmdRequestDMA=0ULL;
constexpr int64_t n3ds_gxCmdProcessCmdList=1ULL;
constexpr int64_t n3ds_gxCmdMemoryFill=2ULL;
constexpr int64_t n3ds_gxCmdDisplayTransfer=3ULL;
constexpr int64_t n3ds_gxCmdTextureCopy=4ULL;
constexpr int64_t n3ds_gxCmdFlushCache=5ULL;
constexpr int64_t n3ds_hidA=1ULL;
constexpr int64_t n3ds_hidB=2ULL;
constexpr int64_t n3ds_hidSelect=4ULL;
constexpr int64_t n3ds_hidStart=8ULL;
constexpr int64_t n3ds_hidRight=16ULL;
constexpr int64_t n3ds_hidLeft=32ULL;
constexpr int64_t n3ds_hidUp=64ULL;
constexpr int64_t n3ds_hidDown=128ULL;
constexpr int64_t n3ds_hidR=256ULL;
constexpr int64_t n3ds_hidL=512ULL;
constexpr int64_t n3ds_hidX=1024ULL;
constexpr int64_t n3ds_hidY=2048ULL;
Map<std::string,uint32_t> n3ds_hidButtonNames=Map<std::string,uint32_t>{{std::string("a",1),cast<uint32_t>(1ULL)},{std::string("b",1),cast<uint32_t>(2ULL)},{std::string("x",1),cast<uint32_t>(1024ULL)},{std::string("y",1),cast<uint32_t>(2048ULL)},{std::string("l",1),cast<uint32_t>(512ULL)},{std::string("r",1),cast<uint32_t>(256ULL)},{std::string("up",2),cast<uint32_t>(64ULL)},{std::string("down",4),cast<uint32_t>(128ULL)},{std::string("left",4),cast<uint32_t>(32ULL)},{std::string("right",5),cast<uint32_t>(16ULL)},{std::string("start",5),cast<uint32_t>(8ULL)},{std::string("select",6),cast<uint32_t>(4ULL)}};
constexpr int64_t n3ds_hidPadBase=0ULL;
constexpr int64_t n3ds_hidTouchBase=168ULL;
constexpr int64_t n3ds_hidAccelBase=264ULL;
constexpr int64_t n3ds_hidHdrIndex=16ULL;
constexpr int64_t n3ds_hidEntries=40ULL;
constexpr int64_t n3ds_hidEntSz=16ULL;
constexpr int64_t n3ds_hidRingLen=8ULL;
constexpr int64_t n3ds_hidTouchEntries=32ULL;
constexpr int64_t n3ds_hidTouchEntSz=8ULL;
constexpr int64_t n3ds_cfgLangEnglish=1ULL;
constexpr int64_t n3ds_cfgRegionEUR=2ULL;
constexpr int64_t n3ds_cfgConsoleSeed=5937279718600631670ULL;
std::array<float,8> n3ds_cfgStereoCamera=std::array<float,8>{cast<float>(6.20000000000000000e+01),cast<float>(2.89000000000000000e+02),cast<float>(7.68000030517578125e+01),cast<float>(4.60800018310546875e+01),cast<float>(1.00000000000000000e+01),cast<float>(5.00000000000000000e+00),cast<float>(5.55800018310546875e+01),cast<float>(2.15599994659423828e+01)};
constexpr int64_t n3ds_codeBase=1048576ULL;
constexpr int64_t n3ds_heapBase=134217728ULL;
constexpr int64_t n3ds_heapMax=234881024ULL;
constexpr int64_t n3ds_linearBase=335544320ULL;
constexpr int64_t n3ds_linearMax=469762048ULL;
constexpr int64_t n3ds_stackTop=268435456ULL;
constexpr int64_t n3ds_configPage=536346624ULL;
constexpr int64_t n3ds_tlsBase=536354816ULL;
constexpr int64_t n3ds_tlsSize=4096ULL;
constexpr int64_t n3ds_pageSize=4096ULL;
constexpr int64_t n3ds_appMemAlloc=67108864ULL;
constexpr int64_t n3ds_pageShift=12ULL;
constexpr int64_t n3ds_ncchHeaderOff=256ULL;
const std::string n3ds_ncchMagic=std::string("NCCH",4);
constexpr int64_t n3ds_flagCryptoMethod=3ULL;
constexpr int64_t n3ds_flagContentPlatfm=4ULL;
constexpr int64_t n3ds_flagContentType=5ULL;
constexpr int64_t n3ds_flagContentUnit=6ULL;
constexpr int64_t n3ds_flagOther=7ULL;
constexpr int64_t n3ds_otherFixedKey=1ULL;
constexpr int64_t n3ds_otherNoRomFS=2ULL;
constexpr int64_t n3ds_otherNoCrypto=4ULL;
constexpr int64_t n3ds_otherSeedCryp=32ULL;
constexpr int64_t n3ds_ncsdSigSize=256ULL;
constexpr int64_t n3ds_ncsdHeaderOff=256ULL;
constexpr int64_t n3ds_ncsdHeaderSize=256ULL;
const std::string n3ds_ncsdMagic=std::string("NCSD",4);
constexpr int64_t n3ds_NumPartitions=8ULL;
constexpr int64_t n3ds_bucketCmd=0ULL;
constexpr int64_t n3ds_bucketVertex=1ULL;
constexpr int64_t n3ds_bucketRaster=2ULL;
constexpr int64_t n3ds_bucketTexture=3ULL;
constexpr int64_t n3ds_bucketGX=4ULL;
constexpr int64_t n3ds_bucketDSP=5ULL;
constexpr int64_t n3ds_bucketSVC=6ULL;
constexpr int64_t n3ds_numBuckets=7ULL;
std::array<std::string,7> n3ds_bucketNames=[](){std::array<std::string,7> v{};v[0]=std::string("command decode",14);v[1]=std::string("vertex + shader",15);v[2]=std::string("rasterise",9);v[3]=std::string("texture decode",14);v[4]=std::string("gx transfers",12);v[5]=std::string("dsp",3);v[6]=std::string("svc + ipc",9);return v;}();
const std::string n3ds_ivfcMagic=std::string("IVFC",4);
constexpr int64_t n3ds_ivfcMagicNumber=65536ULL;
constexpr int64_t n3ds_ivfcHeaderSize=96ULL;
constexpr int64_t n3ds_romfsHeaderLen=40ULL;
constexpr int64_t n3ds_romfsNoEntry=4294967295ULL;
constexpr int64_t n3ds_maxIdleFrames=40ULL;
constexpr int64_t n3ds_svcControlMemory=1ULL;
constexpr int64_t n3ds_svcQueryMemory=2ULL;
constexpr int64_t n3ds_svcExitProcess=3ULL;
constexpr int64_t n3ds_svcGetProcessAffinity=4ULL;
constexpr int64_t n3ds_svcCreateThread=8ULL;
constexpr int64_t n3ds_svcExitThread=9ULL;
constexpr int64_t n3ds_svcSleepThread=10ULL;
constexpr int64_t n3ds_svcGetThreadPriority=11ULL;
constexpr int64_t n3ds_svcSetThreadPriority=12ULL;
constexpr int64_t n3ds_svcCreateMutex=19ULL;
constexpr int64_t n3ds_svcReleaseMutex=20ULL;
constexpr int64_t n3ds_svcCreateSemaphore=21ULL;
constexpr int64_t n3ds_svcReleaseSemaphore=22ULL;
constexpr int64_t n3ds_svcCreateEvent=23ULL;
constexpr int64_t n3ds_svcSignalEvent=24ULL;
constexpr int64_t n3ds_svcClearEvent=25ULL;
constexpr int64_t n3ds_svcCreateTimer=26ULL;
constexpr int64_t n3ds_svcCreateMemoryBlock=30ULL;
constexpr int64_t n3ds_svcMapMemoryBlock=31ULL;
constexpr int64_t n3ds_svcUnmapMemoryBlock=32ULL;
constexpr int64_t n3ds_svcCreateAddressArb=33ULL;
constexpr int64_t n3ds_svcArbitrateAddress=34ULL;
constexpr int64_t n3ds_svcCloseHandle=35ULL;
constexpr int64_t n3ds_svcWaitSync1=36ULL;
constexpr int64_t n3ds_svcWaitSyncN=37ULL;
constexpr int64_t n3ds_svcDuplicateHandle=39ULL;
constexpr int64_t n3ds_svcGetSystemTick=40ULL;
constexpr int64_t n3ds_svcGetHandleInfo=41ULL;
constexpr int64_t n3ds_svcGetSystemInfo=42ULL;
constexpr int64_t n3ds_svcGetProcessInfo=43ULL;
constexpr int64_t n3ds_svcGetThreadInfo=44ULL;
constexpr int64_t n3ds_svcConnectToPort=45ULL;
constexpr int64_t n3ds_svcSendSyncRequest=50ULL;
constexpr int64_t n3ds_svcGetProcessId=53ULL;
constexpr int64_t n3ds_svcGetThreadId=55ULL;
constexpr int64_t n3ds_svcGetResourceLimit=56ULL;
constexpr int64_t n3ds_svcGetResLimitCurrent=57ULL;
constexpr int64_t n3ds_svcGetResLimitLimit=58ULL;
constexpr int64_t n3ds_svcBreak=60ULL;
constexpr int64_t n3ds_svcOutputDebugString=61ULL;
constexpr uint32_t n3ds_resultSuccess=0ULL;
constexpr int64_t n3ds_committedBytes=4194304ULL;
constexpr uint32_t n3ds_resultTimeout=155196414ULL;
constexpr n3ds_threadState n3ds_ready=0ULL;
constexpr n3ds_threadState n3ds_running=1ULL;
constexpr n3ds_threadState n3ds_waiting=2ULL;
constexpr n3ds_threadState n3ds_sleeping=3ULL;
constexpr n3ds_threadState n3ds_dead=4ULL;
constexpr int64_t n3ds_quantum=2000ULL;
constexpr int64_t n3ds_threadExitSentinel=4294901760ULL;
constexpr int64_t n3ds_sysclockHz=268111856ULL;
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
std::tuple<Slice<uint8_t>,Error> n3ds_DecompressBLZ(Slice<uint8_t> pak);
image_NRGBA* n3ds_decodeETC1(Slice<uint8_t> data,int64_t w,int64_t h);
image_NRGBA* n3ds_decodeETC1A4(Slice<uint8_t> data,int64_t w,int64_t h);
void n3ds_decodeETC1Block(image_NRGBA* img,int64_t bx,int64_t by,uint64_t v);
void n3ds_Machine_handleCP15(n3ds_Machine* m,arm_CPU* c,bool load,uint32_t cp,uint32_t op1,uint32_t crn,uint32_t crm,uint32_t op2,uint32_t* rd);
void n3ds_GPU_pixelEvent(n3ds_GPU* g,uint32_t x,uint32_t y,n3ds_PixelEvent ev);
int64_t n3ds_Machine_RunStopAfterPICACommand(n3ds_Machine* m,int64_t k,int64_t budget);
int64_t n3ds_Machine_PICACommands(n3ds_Machine* m);
void n3ds_Machine_ClearBreakpoints(n3ds_Machine* m);
bool n3ds_Machine_Stopped(n3ds_Machine* m);
Slice<n3ds_MemRegion> n3ds_Machine_MemRegions(n3ds_Machine* m);
n3ds_RomFS* n3ds_Machine_RomFS(n3ds_Machine* m);
std::tuple<uint32_t,uint32_t,uint32_t> n3ds_GPU_ColorTarget(n3ds_GPU* g);
std::tuple<uint32_t,uint32_t,uint32_t> n3ds_GPU_DepthTarget(n3ds_GPU* g);
image_NRGBA* n3ds_Machine_RenderDepth(n3ds_Machine* m,uint32_t addr,uint32_t w,uint32_t h);
Slice<std::string> n3ds_TextureFormats();
std::tuple<uint32_t,bool> n3ds_TextureFormat(std::string name);
std::tuple<image_NRGBA*,Error> n3ds_Machine_RenderTexture(n3ds_Machine* m,uint32_t addr,uint32_t format,uint32_t w,uint32_t h);
bool n3ds_Machine_ipcDSP(n3ds_Machine* m,n3ds_ipcHeader hdr);
bool n3ds_Machine_dspWritePipe(n3ds_Machine* m,n3ds_ipcHeader hdr);
void n3ds_Machine_dspAnnounceStructs(n3ds_Machine* m);
void n3ds_Machine_dspReplyPipeData(n3ds_Machine* m,n3ds_ipcHeader hdr,uint32_t ch,uint32_t size,bool ifPossible);
bool n3ds_Machine_dspDue(n3ds_Machine* m);
std::tuple<uint64_t,bool> n3ds_Machine_dspDeadline(n3ds_Machine* m);
void n3ds_Machine_dspTick(n3ds_Machine* m);
uint32_t n3ds_Machine_dspReadRegion(n3ds_Machine* m);
uint32_t n3ds_Machine_dspWriteRegion(n3ds_Machine* m);
void n3ds_Machine_dspParseConfig(n3ds_Machine* m,int64_t i,uint32_t region);
void n3ds_Machine_dspWriteStatus(n3ds_Machine* m,int64_t i,uint32_t region);
void n3ds_Machine_dspSignalInterrupt(n3ds_Machine* m,uint32_t interrupt,uint32_t channel);
void n3ds_Machine_dspSignalHandle(n3ds_Machine* m,uint32_t h);
uint16_t n3ds_Machine_dspRead16(n3ds_Machine* m,uint32_t a);
void n3ds_Machine_dspWrite16(n3ds_Machine* m,uint32_t a,uint16_t v);
uint32_t n3ds_Machine_dspRead32(n3ds_Machine* m,uint32_t a);
void n3ds_Machine_dspWrite32(n3ds_Machine* m,uint32_t a,uint32_t v);
void n3ds_dspFilters_reset(n3ds_dspFilters* f);
void n3ds_dspFilters_enable(n3ds_dspFilters* f,bool simple,bool biquad);
void n3ds_dspFilters_processFrame(n3ds_dspFilters* f,n3ds_dspFrame* frame);
int16_t n3ds_clampS16(int32_t v);
int16_t n3ds_addClampS16(int16_t a,int16_t b);
void n3ds_Machine_dspSourceFrame(n3ds_Machine* m,int64_t i);
void n3ds_dspSource_resample(n3ds_dspSource* s,int64_t* outi);
n3ds_dspSample n3ds_interpolate(uint8_t mode,uint64_t frac,n3ds_dspSample x0,n3ds_dspSample x1);
bool n3ds_Machine_dspDequeue(n3ds_Machine* m,n3ds_dspSource* s);
Slice<n3ds_dspSample> n3ds_Machine_dspDecodeBuffer(n3ds_Machine* m,n3ds_dspSource* s,n3ds_dspBuffer buf);
int16_t n3ds_pcm8(uint8_t b);
int16_t n3ds_pcm16(Slice<uint8_t> b);
Slice<n3ds_dspSample> n3ds_Machine_decodeADPCM(n3ds_Machine* m,n3ds_dspSource* s,n3ds_dspBuffer buf,Slice<uint8_t> data);
void n3ds_dspSource_mixInto(n3ds_dspSource* s,n3ds_dspQuadFrame* dest,int64_t mix);
void n3ds_Machine_dspMixerConfig(n3ds_Machine* m,uint32_t region);
float n3ds_dspFloat(uint32_t w);
n3ds_dspFrame n3ds_Machine_dspMix(n3ds_Machine* m,uint32_t read,uint32_t write,std::array<n3ds_dspQuadFrame,3>* mixes);
void n3ds_Machine_dspReadQuad(n3ds_Machine* m,uint32_t base,n3ds_dspQuadFrame* q);
void n3ds_Machine_dspWriteQuad(n3ds_Machine* m,uint32_t base,n3ds_dspQuadFrame* q);
void n3ds_Machine_dspWriteFinal(n3ds_Machine* m,uint32_t write,n3ds_dspFrame* f);
std::tuple<n3ds_ExeFS*,Error> n3ds_ParseExeFS(Slice<uint8_t> b);
std::tuple<Slice<uint8_t>,Error> n3ds_ExeFS_File(n3ds_ExeFS* fs,std::string name);
std::tuple<Slice<uint8_t>,Error> n3ds_ExeFS_Code(n3ds_ExeFS* fs,n3ds_ExHeader* ex);
uint32_t n3ds_CodeSegInfo_Extent(n3ds_CodeSegInfo s);
bool n3ds_ExHeader_CompressedExeFSCode(n3ds_ExHeader* e);
bool n3ds_ExHeader_SDApplication(n3ds_ExHeader* e);
std::tuple<n3ds_ExHeader*,Error> n3ds_ParseExHeader(Slice<uint8_t> b);
Error n3ds_ExHeader_validate(n3ds_ExHeader* e);
uint32_t n3ds_ExHeader_CodeSize(n3ds_ExHeader* e);
uint32_t n3ds_ExHeader_BSSAddress(n3ds_ExHeader* e);
bool n3ds_Machine_isFileSession(n3ds_Machine* m,uint32_t handle);
bool n3ds_Machine_isDirSession(n3ds_Machine* m,uint32_t handle);
bool n3ds_Machine_fsOpenDirectory(n3ds_Machine* m,n3ds_ipcHeader hdr);
std::tuple<n3ds_fsDir*,bool> n3ds_Machine_romfsChildren(n3ds_Machine* m,std::string path);
bool n3ds_containsSlash(std::string s);
bool n3ds_Machine_ipcDir(n3ds_Machine* m,uint32_t handle,n3ds_ipcHeader hdr);
bool n3ds_Machine_fsOpenFile(n3ds_Machine* m,n3ds_ipcHeader hdr);
std::string n3ds_Machine_readFSPath(n3ds_Machine* m,uint32_t ptr,uint32_t ptype,uint32_t size);
bool n3ds_Machine_ipcFile(n3ds_Machine* m,uint32_t handle,n3ds_ipcHeader hdr);
int64_t n3ds_fileLen(n3ds_fsFile* f);
std::string n3ds_filePath(n3ds_fsFile* f);
bool n3ds_Machine_fsOpenArchive(n3ds_Machine* m,n3ds_ipcHeader hdr);
bool n3ds_Machine_fsDeleteFile(n3ds_Machine* m,n3ds_ipcHeader hdr);
bool n3ds_Machine_fsCreateFile(n3ds_Machine* m,n3ds_ipcHeader hdr);
n3ds_GPU* n3ds_newGPU(n3ds_Machine* m);
n3ds_GPU* n3ds_Machine_GPU(n3ds_Machine* m);
void n3ds_GPU_Execute(n3ds_GPU* g,uint32_t addr,uint32_t size);
void n3ds_GPU_write(n3ds_GPU* g,n3ds_PICAWrite w);
void n3ds_GPU_floatUniformWord(n3ds_GPU* g,uint32_t v);
void n3ds_GPU_drawArrays(n3ds_GPU* g);
void n3ds_GPU_drawElements(n3ds_GPU* g);
bool n3ds_GPU_checkUnsupported(n3ds_GPU* g);
float n3ds_f24bits(uint32_t v);
float n3ds_f32bits(uint32_t v);
std::array<float,4> n3ds_unpackF24x4(uint32_t w0,uint32_t w1,uint32_t w2);
float n3ds_toF24(float v);
std::string n3ds_lutName(int64_t t);
void n3ds_GPU_lutWrite(n3ds_GPU* g,uint32_t v);
float n3ds_GPU_lutLookup(n3ds_GPU* g,int64_t t,uint8_t idx,float delta);
float n3ds_f16(uint32_t v);
float n3ds_f20(uint32_t v);
float n3ds_fixed11(uint32_t v);
std::array<float,3> n3ds_rgb10(uint32_t v);
n3ds_lightState n3ds_GPU_lightstate(n3ds_GPU* g);
bool n3ds_lutSupported(uint32_t env,int64_t t);
float n3ds_lutScaleOf(uint32_t v);
std::tuple<std::array<float,4>,std::array<float,4>> n3ds_GPU_shade(n3ds_GPU* g,n3ds_lightState* ls,std::array<float,4> quat,std::array<float,3> view,std::array<n3ds_rgba,3>* tex);
std::tuple<uint8_t,float> n3ds_lutIndexAbs(float v);
std::array<float,3> n3ds_quatRotate(std::array<float,4> q,std::array<float,3> v);
float n3ds_dot3(std::array<float,3> a,std::array<float,3> b);
std::array<float,3> n3ds_scale3(std::array<float,3> v,float s);
std::array<float,3> n3ds_normalize3(std::array<float,3> v);
float n3ds_sqrt32(float f);
float n3ds_absf(float f);
float n3ds_minf(float a,float b);
float n3ds_clampf(float v,float lo,float hi);
void n3ds_GPU_dumpLighting(n3ds_GPU* g);
void n3ds_GPU_dumpFragment(n3ds_GPU* g);
std::string n3ds_tevSrcName(uint32_t s);
std::string n3ds_tevCOpName(uint32_t o);
std::string n3ds_tevAOpName(uint32_t o);
std::string n3ds_tevOpName(uint32_t o);
image_NRGBA* n3ds_Machine_PresentedImage(n3ds_Machine* m,n3ds_ScreenGeom g);
image_NRGBA* n3ds_Machine_Framebuffer(n3ds_Machine* m,std::string screen);
image_NRGBA* n3ds_Machine_decodeFB(n3ds_Machine* m,uint32_t addr,uint32_t fw,uint32_t fh,uint32_t stride,uint32_t format,uint32_t bpp);
image_NRGBA* n3ds_Machine_RenderTarget(n3ds_Machine* m,uint32_t addr,uint32_t w,uint32_t h);
std::tuple<int64_t,int64_t> n3ds_ScreenGeom_Size(n3ds_ScreenGeom g);
std::tuple<uint32_t,uint32_t,bool> n3ds_ScreenGeom_Source(n3ds_ScreenGeom g,int64_t sx,int64_t sy);
std::tuple<n3ds_ScreenGeom,bool> n3ds_Machine_ScreenGeom(n3ds_Machine* m,std::string screen);
image_NRGBA* n3ds_Machine_ScreenImage(n3ds_Machine* m,n3ds_ScreenGeom g);
uint32_t n3ds_fbBPP(uint32_t format);
void n3ds_GPU_draw(n3ds_GPU* g,bool indexed);
void n3ds_GPU_fill(n3ds_GPU* g,n3ds_fbState* fb,n3ds_lightState* ls,n3ds_tevState* tv,Slice<n3ds_rasterTri> tris);
int64_t n3ds_GPU_rasterWorkers(n3ds_GPU* g,n3ds_fbState* fb,Slice<n3ds_rasterTri> tris);
void n3ds_GPU_warmTextures(n3ds_GPU* g);
void n3ds_mapAttrsToInputs(std::array<std::array<float,4>,16>* v,std::array<std::array<float,4>,16>* attrs,uint64_t perm,int64_t nAttr);
void n3ds_GPU_mapOutputs(n3ds_GPU* g,std::array<std::array<float,4>,16>* o,n3ds_vsOut* dst);
n3ds_fbState n3ds_GPU_fbstate(n3ds_GPU* g);
void n3ds_GPU_shadowMapWrite(n3ds_GPU* g,n3ds_fbState* fb,uint32_t off,float depth,uint8_t density);
void n3ds_GPU_mergeStats(n3ds_GPU* g,n3ds_rstats* st);
float n3ds_edgeFn(float ax,float ay,float bx,float by,float px,float py);
std::tuple<n3ds_rasterTri,bool> n3ds_GPU_setupTri(n3ds_GPU* g,n3ds_vsOut* a,n3ds_vsOut* b,n3ds_vsOut* c,n3ds_fbState* fb);
void n3ds_GPU_fillTri(n3ds_GPU* g,n3ds_fbState* fb,n3ds_lightState* ls,n3ds_tevState* tv,n3ds_rasterTri* t,int64_t yLo,int64_t yHi,n3ds_rstats* st);
uint32_t n3ds_tiledOffset(uint32_t x,uint32_t y,uint32_t width);
bool n3ds_GPU_depthCompare(n3ds_GPU* g,n3ds_fbState* fb,uint32_t off,float depth);
void n3ds_GPU_depthWrite(n3ds_GPU* g,n3ds_fbState* fb,uint32_t off,float depth);
void n3ds_GPU_writePixel(n3ds_GPU* g,n3ds_fbState* fb,uint32_t x,uint32_t y,uint32_t off,uint8_t r,uint8_t gr,uint8_t b,uint8_t a);
float n3ds_min3f(float a,float b,float c);
float n3ds_max3f(float a,float b,float c);
int64_t n3ds_GPU_vertexWorkers(n3ds_GPU* g,uint32_t count);
uint32_t n3ds_GPU_stencilDepthTest(n3ds_GPU* g,n3ds_fbState* fb,uint32_t off,float depth);
uint8_t n3ds_stencilValue(uint32_t op,uint8_t old,uint8_t ref);
float n3ds_clipDistance(n3ds_vsOut v,int64_t plane);
n3ds_vsOut n3ds_clipVertex(n3ds_vsOut a,n3ds_vsOut b,float t);
Slice<n3ds_rasterTri> n3ds_GPU_clipTri(n3ds_GPU* g,Slice<n3ds_rasterTri> tris,n3ds_vsOut a,n3ds_vsOut b,n3ds_vsOut c,n3ds_fbState* fb);
bool n3ds_GPU_shaderRun(n3ds_GPU* g,std::array<std::array<float,4>,16>* v,std::array<std::array<float,4>,16>* out,int64_t entry);
bool n3ds_shaderState_exec(n3ds_shaderState* s,int64_t pc,int64_t end);
bool n3ds_shaderState_cond(n3ds_shaderState* s,uint32_t in);
bool n3ds_shaderState_boolReg(n3ds_shaderState* s,uint32_t in);
bool n3ds_shaderState_arith(n3ds_shaderState* s,n3ds_shInst* d);
std::array<float,4> n3ds_shaderState_src(n3ds_shaderState* s,n3ds_shSrc* o);
void n3ds_shaderState_writeDst(n3ds_shaderState* s,n3ds_shInst* d,std::array<float,4> val);
bool n3ds_compare(uint32_t op,float a,float b);
float n3ds_floor32(float f);
float n3ds_rsqrt32(float f);
n3ds_shInst* n3ds_GPU_shaderInst(n3ds_GPU* g,int64_t pc);
void n3ds_GPU_decodeAll(n3ds_GPU* g);
void n3ds_GPU_invalidateShaders(n3ds_GPU* g);
void n3ds_GPU_decode(n3ds_GPU* g,int64_t pc);
void n3ds_shInst_setDst(n3ds_shInst* d,int64_t reg,uint32_t desc);
n3ds_shSrc n3ds_decodeSrc(int64_t reg,int64_t idx,uint32_t desc,int64_t n);
std::tuple<uint8_t,uint8_t,uint8_t,uint8_t,bool,bool> n3ds_GPU_fragment(n3ds_GPU* g,std::array<float,4> vcol,std::array<std::array<float,2>,3> uv,float uv0w,n3ds_lightState* ls,n3ds_tevState* tv,std::array<float,4> quat,std::array<float,3> view,n3ds_rstats* st);
n3ds_rgba n3ds_tevColorOperand(n3ds_rgba s,uint32_t op);
int32_t n3ds_tevAlphaOperand(n3ds_rgba s,uint32_t op);
std::tuple<int32_t,bool> n3ds_tevCombine(uint32_t op,int32_t a,int32_t b,int32_t c);
uint8_t n3ds_logicOp(uint32_t op,uint8_t s,uint8_t d);
std::tuple<uint8_t,uint8_t,uint8_t,uint8_t> n3ds_GPU_blend(n3ds_GPU* g,uint8_t sr,uint8_t sg,uint8_t sb,uint8_t sa,uint8_t dr,uint8_t dg,uint8_t db,uint8_t da);
int32_t n3ds_clamp255(float f);
int32_t n3ds_clampi(int32_t v,int32_t lo,int32_t hi);
std::tuple<n3ds_rgba,bool> n3ds_tevState_run(n3ds_tevState* t,n3ds_rgba vertex,n3ds_rgba fragPrim,n3ds_rgba fragSec,std::array<n3ds_rgba,3> tex);
n3ds_tevState n3ds_GPU_tevstate(n3ds_GPU* g);
n3ds_tevState n3ds_tevStateFromRegs(std::array<uint32_t,768>* r);
std::tuple<uint32_t,uint32_t,uint32_t,uint32_t> n3ds_texUnitRegs(int64_t u);
std::tuple<n3ds_rgba,bool> n3ds_GPU_sampleTextureSt(n3ds_GPU* g,int64_t u,float s,float t,float q,n3ds_rstats* st);
n3ds_rgba n3ds_shadowCompare(n3ds_rgba c,float z,uint32_t cfg);
int32_t n3ds_wrapCoord(float f,int32_t n,uint32_t mode);
std::tuple<n3ds_texImage*,bool> n3ds_GPU_texture(n3ds_GPU* g,uint32_t addr,uint32_t format,uint32_t w,uint32_t h);
void n3ds_GPU_eachTexel(n3ds_GPU* g,uint32_t addr,uint32_t w,uint32_t h,uint32_t bpp,std::function<void(uint32_t,uint32_t,uint32_t)> visit);
void n3ds_GPU_decode4(n3ds_GPU* g,uint32_t addr,uint32_t w,uint32_t h,n3ds_texImage* img,std::function<std::array<uint8_t,4>(uint8_t)> px);
uint32_t n3ds_texBytes(n3ds_texKey k);
void n3ds_GPU_invalidateTextures(n3ds_GPU* g,uint32_t addr,uint32_t size);
uint32_t n3ds_Machine_gpuAddrToVirt(n3ds_Machine* m,uint32_t a);
void n3ds_Machine_svcMapMemoryBlock(n3ds_Machine* m,arm_CPU* c);
void n3ds_Machine_svcUnmapMemoryBlock(n3ds_Machine* m,arm_CPU* c);
bool n3ds_Machine_vblankDue(n3ds_Machine* m);
n3ds_FBPresent n3ds_Machine_Scanout(n3ds_Machine* m,int64_t screen);
void n3ds_Machine_deliverVBlank(n3ds_Machine* m);
void n3ds_Machine_consumeFBInfo(n3ds_Machine* m);
void n3ds_Machine_signalGSPEvent(n3ds_Machine* m);
void n3ds_Machine_pushGSPInterrupt(n3ds_Machine* m,uint8_t id);
uint64_t n3ds_Machine_VBlanks(n3ds_Machine* m);
Slice<n3ds_GXRecord> n3ds_Machine_GXLog(n3ds_Machine* m);
void n3ds_Machine_captureChainedList(n3ds_Machine* m,uint32_t addr,uint32_t size,Slice<uint8_t> buf);
void n3ds_Machine_captureGX(n3ds_Machine* m,uint32_t cmd,uint32_t id);
void n3ds_Machine_gxMemoryFill(n3ds_Machine* m,uint32_t start,uint32_t value,uint32_t end,uint32_t ctl);
void n3ds_Machine_gxTextureCopy(n3ds_Machine* m,uint32_t src,uint32_t dst,uint32_t size,uint32_t inDim,uint32_t outDim);
void n3ds_Machine_gxDisplayTransfer(n3ds_Machine* m,uint32_t src,uint32_t dst,uint32_t srcDims,uint32_t dstDims,uint32_t flags);
void n3ds_Machine_processGXQueue(n3ds_Machine* m);
uint64_t n3ds_gxLatency(uint32_t id);
std::tuple<uint64_t,bool> n3ds_Machine_gxDeadline(n3ds_Machine* m);
void n3ds_Machine_pumpGX(n3ds_Machine* m);
Error n3ds_Machine_SetKeys(n3ds_Machine* m,std::string list);
void n3ds_Machine_SetTouch(n3ds_Machine* m,int64_t x,int64_t y,bool down);
int64_t n3ds_clampInt(int64_t v,int64_t lo,int64_t hi);
void n3ds_Machine_updateHIDShared(n3ds_Machine* m);
void n3ds_Machine_DumpHIDReads(n3ds_Machine* m);
n3ds_ipcHeader n3ds_parseIPCHeader(uint32_t w);
uint32_t n3ds_Machine_ipcArg(n3ds_Machine* m,int64_t i);
void n3ds_Machine_ipcReply(n3ds_Machine* m,uint16_t cmd,Slice<uint32_t> values);
bool n3ds_Machine_handleIPC(n3ds_Machine* m,uint32_t handle);
bool n3ds_Machine_ipcSrv(n3ds_Machine* m,n3ds_ipcHeader hdr);
void n3ds_Machine_publishNotification(n3ds_Machine* m,uint32_t id);
bool n3ds_knownService(std::string name);
std::string n3ds_Machine_readServiceName(n3ds_Machine* m);
std::string n3ds_Machine_decodeName(n3ds_Machine* m,uint64_t rot);
uint32_t n3ds_Machine_newHandle(n3ds_Machine* m,std::string kind,bool signalled);
std::string n3ds_ipcCall_Service(n3ds_ipcCall c);
Slice<n3ds_ipcCall> n3ds_Machine_IPCLog(n3ds_Machine* m);
bool n3ds_Machine_ipcService(n3ds_Machine* m,std::string name,n3ds_ipcHeader hdr);
std::string n3ds_serviceBase(std::string name);
bool n3ds_Machine_ipcAPT(n3ds_Machine* m,std::string name,n3ds_ipcHeader hdr);
void n3ds_Machine_signalAPTEvents(n3ds_Machine* m);
bool n3ds_Machine_ipcGSP(n3ds_Machine* m,n3ds_ipcHeader hdr);
bool n3ds_Machine_ipcHID(n3ds_Machine* m,n3ds_ipcHeader hdr);
bool n3ds_Machine_ipcCFG(n3ds_Machine* m,n3ds_ipcHeader hdr);
void n3ds_Machine_writeConfigBlock(n3ds_Machine* m,uint32_t blkID,uint32_t out,uint32_t size);
bool n3ds_Machine_ipcFS(n3ds_Machine* m,n3ds_ipcHeader hdr);
std::tuple<int64_t,int64_t> n3ds_Machine_FrameStats(n3ds_Machine* m);
int64_t n3ds_Machine_DisplayTransfers(n3ds_Machine* m);
bool n3ds_Machine_ipcErr(n3ds_Machine* m,n3ds_ipcHeader hdr);
bool n3ds_memRegion_contains(n3ds_memRegion* r,uint32_t a);
Slice<uint8_t> n3ds_Machine_ReadBytes(n3ds_Machine* m,uint32_t addr,uint32_t length);
Slice<uint32_t> n3ds_Machine_FindBytes(n3ds_Machine* m,Slice<uint8_t> pat);
std::tuple<n3ds_Machine*,Error> n3ds_NewMachine(Slice<uint8_t> img);
n3ds_memRegion* n3ds_Machine_mapRegion(n3ds_Machine* m,std::string name,uint32_t base,Slice<uint8_t> data);
Slice<uint8_t> n3ds_Machine_buildConfigPage(n3ds_Machine* m);
void n3ds_Machine_indexRegion(n3ds_Machine* m,n3ds_memRegion* r);
void n3ds_Machine_clearPages(n3ds_Machine* m);
n3ds_memRegion* n3ds_Machine_regionOf(n3ds_Machine* m,uint32_t a);
uint8_t n3ds_Machine_Read(n3ds_Machine* m,uint32_t a);
void n3ds_Machine_Write(n3ds_Machine* m,uint32_t a,uint8_t v);
uint32_t n3ds_Machine_ReadWord(n3ds_Machine* m,uint32_t a);
void n3ds_Machine_WriteWord(n3ds_Machine* m,uint32_t a,uint32_t v);
std::tuple<Slice<uint8_t>,uint32_t> n3ds_Machine_directRange(n3ds_Machine* m,uint32_t addr,uint32_t size);
void n3ds_Machine_copyRange(n3ds_Machine* m,uint32_t dst,uint32_t src,uint32_t size);
std::tuple<Slice<uint8_t>,Error> n3ds_Yaz0(Slice<uint8_t> b);
std::tuple<Slice<Slice<uint8_t>>,Error> n3ds_NARCFiles(Slice<uint8_t> d);
std::tuple<Slice<n3ds_MSBTMessage>,Error> n3ds_ParseMSBT(Slice<uint8_t> fd);
std::string n3ds_decodeUTF16(Slice<uint8_t> b);
bool n3ds_region_empty(n3ds_region r);
std::tuple<n3ds_NCCH*,Error> n3ds_ParseNCCH(Slice<uint8_t> part);
bool n3ds_NCCH_Encrypted(n3ds_NCCH* c);
std::string n3ds_NCCH_CryptoMethod(n3ds_NCCH* c);
std::string n3ds_NCCH_ContentType(n3ds_NCCH* c);
Error n3ds_NCCH_checkPlain(n3ds_NCCH* c,std::string what);
std::tuple<Slice<uint8_t>,Error> n3ds_NCCH_ExHeaderBytes(n3ds_NCCH* c);
std::tuple<n3ds_ExHeader*,Error> n3ds_NCCH_ExHeader(n3ds_NCCH* c);
std::tuple<Slice<uint8_t>,Error> n3ds_NCCH_ExeFSBytes(n3ds_NCCH* c);
std::tuple<n3ds_ExeFS*,Error> n3ds_NCCH_ExeFS(n3ds_NCCH* c);
std::tuple<Slice<uint8_t>,Error> n3ds_NCCH_RomFSBytes(n3ds_NCCH* c);
std::tuple<n3ds_RomFS*,Error> n3ds_NCCH_RomFS(n3ds_NCCH* c);
Slice<uint8_t> n3ds_trimNul(Slice<uint8_t> b);
bool n3ds_Partition_Empty(n3ds_Partition p);
Slice<uint8_t> n3ds_NCSD_Raw(n3ds_NCSD* n);
std::tuple<n3ds_NCSD*,Error> n3ds_ParseNCSD(Slice<uint8_t> img);
Slice<uint8_t> n3ds_NCSD_Bytes(n3ds_NCSD* n,int64_t i);
std::tuple<n3ds_NCCH*,Error> n3ds_NCSD_Partition(n3ds_NCSD* n,int64_t i);
std::tuple<n3ds_NCCH*,Error> n3ds_NCSD_Executable(n3ds_NCSD* n);
std::tuple<Slice<n3ds_PICAWrite>,Error> n3ds_DecodePICA(Slice<uint8_t> buf);
std::tuple<Slice<n3ds_PICAWrite>,Error> n3ds_DecodePICAInto(Slice<n3ds_PICAWrite> dst,Slice<uint8_t> buf);
std::string n3ds_PICARegGroup(uint16_t reg);
n3ds_profCounters n3ds_Machine_profCounters(n3ds_Machine* m);
time_Time n3ds_Machine_profStart(n3ds_Machine* m);
void n3ds_Machine_profEnd(n3ds_Machine* m,int64_t bucket,time_Time t);
void n3ds_Machine_profRunEnter(n3ds_Machine* m);
void n3ds_Machine_profRunExit(n3ds_Machine* m);
void n3ds_Machine_profFrame(n3ds_Machine* m);
n3ds_FrameProfile n3ds_Machine_FrameProfile(n3ds_Machine* m);
void n3ds_Machine_SetProfile(n3ds_Machine* m,bool on);
uint64_t n3ds_ivfcLevel_blockSize(n3ds_ivfcLevel l);
uint64_t n3ds_align64(uint64_t v,uint64_t a);
std::tuple<n3ds_RomFS*,Error> n3ds_ParseRomFS(Slice<uint8_t> b);
Error n3ds_RomFS_walk(n3ds_RomFS* fs,Slice<uint8_t> dirMeta,Slice<uint8_t> fileMeta,uint32_t dirOff,std::string prefix);
std::tuple<std::string,Error> n3ds_utf16Name(Slice<uint8_t> e,int64_t off,uint32_t nameLen,int64_t avail);
int64_t n3ds_min64(int64_t a,int64_t b);
std::tuple<Slice<uint8_t>,Error> n3ds_RomFS_File(n3ds_RomFS* fs,std::string path);
Slice<uint8_t> n3ds_RomFS_Data(n3ds_RomFS* fs,n3ds_RomFSFile f);
int64_t n3ds_RomFS_L3Offset(n3ds_RomFS* fs,n3ds_RomFSFile f);
std::tuple<n3ds_RomFSFile,int64_t,bool> n3ds_RomFS_FileAt(n3ds_RomFS* fs,int64_t l3Off);
int64_t n3ds_Machine_Run(n3ds_Machine* m,int64_t budget);
int64_t n3ds_Machine_RunFrames(n3ds_Machine* m,int64_t frames,int64_t budget);
void n3ds_Machine_SetTrace(n3ds_Machine* m,bool on,int64_t max);
void n3ds_Machine_AddBreakpoint(n3ds_Machine* m,uint32_t addr);
void n3ds_Machine_AddTraceFrom(n3ds_Machine* m,uint32_t addr);
void n3ds_Machine_AddLogPC(n3ds_Machine* m,uint32_t addr);
void n3ds_Machine_AddWatch(n3ds_Machine* m,uint32_t addr,uint32_t length);
void n3ds_Machine_traceOne(n3ds_Machine* m,uint32_t pc);
void n3ds_Machine_checkWatches(n3ds_Machine* m,uint32_t pc);
std::string n3ds_Machine_HaltReason(n3ds_Machine* m);
std::string n3ds_Machine_DebugString(n3ds_Machine* m);
uint32_t n3ds_Machine_Entry(n3ds_Machine* m);
uint64_t n3ds_Machine_Instrs(n3ds_Machine* m);
std::string n3ds_Machine_argStrings(n3ds_Machine* m);
bool n3ds_Machine_handleSVC(n3ds_Machine* m,arm_CPU* c,uint32_t comment);
void n3ds_Machine_svcControlMemory(n3ds_Machine* m,arm_CPU* c);
void n3ds_Machine_svcQueryMemory(n3ds_Machine* m,arm_CPU* c);
void n3ds_Machine_svcCreateHandle(n3ds_Machine* m,arm_CPU* c,std::string kind,bool signalled,int64_t handleReg);
void n3ds_Machine_svcConnectToPort(n3ds_Machine* m,arm_CPU* c);
void n3ds_Machine_svcSendSyncRequest(n3ds_Machine* m,arm_CPU* c);
void n3ds_Machine_svcGetResourceLimitValues(n3ds_Machine* m,arm_CPU* c,bool current);
void n3ds_Machine_svcOutputDebugString(n3ds_Machine* m,arm_CPU* c);
std::string n3ds_Machine_readCString(n3ds_Machine* m,uint32_t addr,uint32_t max);
std::string n3ds_svcName(uint32_t n);
Slice<n3ds_svcEvent> n3ds_Machine_SVCLog(n3ds_Machine* m);
Map<uint32_t,std::string> n3ds_Machine_Ports(n3ds_Machine* m);
bool n3ds_Machine_objAvailable(n3ds_Machine* m,n3ds_kobject* obj);
void n3ds_Machine_consume(n3ds_Machine* m,n3ds_kobject* obj,n3ds_thread* t);
void n3ds_Machine_svcWaitSync1(n3ds_Machine* m,arm_CPU* c);
void n3ds_Machine_armWaitDeadline(n3ds_Machine* m,int64_t ns);
void n3ds_Machine_svcWaitSyncN(n3ds_Machine* m,arm_CPU* c);
bool n3ds_Machine_signalObject(n3ds_Machine* m,n3ds_kobject* obj);
bool n3ds_Machine_tryComplete(n3ds_Machine* m,n3ds_thread* t);
bool n3ds_Machine_availableFor(n3ds_Machine* m,n3ds_kobject* obj,n3ds_thread* t);
void n3ds_Machine_consumeFor(n3ds_Machine* m,n3ds_kobject* obj,n3ds_thread* t);
n3ds_thread* n3ds_Machine_threadByID(n3ds_Machine* m,uint32_t id);
void n3ds_Machine_svcSignalEvent(n3ds_Machine* m,arm_CPU* c);
void n3ds_Machine_svcClearEvent(n3ds_Machine* m,arm_CPU* c);
void n3ds_Machine_svcReleaseMutex(n3ds_Machine* m,arm_CPU* c);
void n3ds_Machine_svcReleaseSemaphore(n3ds_Machine* m,arm_CPU* c);
void n3ds_Machine_signalThreadExit(n3ds_Machine* m,n3ds_thread* t);
void n3ds_Machine_svcArbitrateAddress(n3ds_Machine* m,arm_CPU* c);
void n3ds_Machine_arbPark(n3ds_Machine* m,uint32_t addr,bool timed);
void n3ds_Machine_arbSignal(n3ds_Machine* m,uint32_t addr,int32_t count);
std::string n3ds_threadState_String(n3ds_threadState s);
uint32_t n3ds_Machine_allocTLS(n3ds_Machine* m);
uint32_t n3ds_Machine_cmdBuf(n3ds_Machine* m);
uint32_t n3ds_Machine_createThread(n3ds_Machine* m,int32_t priority,uint32_t entry,uint32_t arg,uint32_t stacktop);
void n3ds_Machine_svcSleepThread(n3ds_Machine* m,arm_CPU* c);
void n3ds_Machine_svcExitThread(n3ds_Machine* m,arm_CPU* c);
n3ds_thread* n3ds_Machine_pickRunnable(n3ds_Machine* m);
void n3ds_Machine_switchTo(n3ds_Machine* m,n3ds_thread* t);
std::tuple<uint64_t,bool> n3ds_Machine_soonestSleeper(n3ds_Machine* m);
void n3ds_Machine_wakeDueSleepers(n3ds_Machine* m);
bool n3ds_Machine_wake(n3ds_Machine* m,n3ds_thread* t);
void n3ds_Machine_setResult(n3ds_Machine* m,n3ds_thread* t,int64_t reg,uint32_t v);
void n3ds_Machine_DumpThreads(n3ds_Machine* m);
void n3ds_Machine_dumpThreads(n3ds_Machine* m);
int64_t n3ds_Machine_aliveThreads(n3ds_Machine* m);
uint64_t n3ds_nsToTick(int64_t ns);

#include "adapters.h"
#include "capture-hooks.h"
#include "performance.h"
#include "graphics-stream.h"
#include "vertex-fast.h"
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
parts = append(parts,arm_regName[i]);
}
else if ((j == cast<int64_t>((i + cast<int64_t>(1ULL))))){
parts = append(parts,arm_regName[i],arm_regName[j]);
}
else {
parts = append(parts,((arm_regName[i] + std::string("-",1)) + arm_regName[j]));
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
return n3ds_Machine_Read(c->bus,a);
}
}
// tools/cpu/arm/cpu.go:253:1
void arm_CPU_write8(arm_CPU* c,uint32_t a,uint8_t v){
{
n3ds_Machine_Write(c->bus,a,v);
}
}
// tools/cpu/arm/cpu.go:269:1
uint32_t arm_CPU_read32(arm_CPU* c,uint32_t a){
{
if (arm_Variant_isV6(c->Arch)) {
if ((bool(c->wide) && (cast<uint32_t>((a & cast<uint32_t>(3ULL))) == cast<uint32_t>(0ULL)))) {
return n3ds_Machine_Read32(c->wide,a);
}
return cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(n3ds_Machine_Read(c->bus,a)) | shl<uint32_t>(cast<uint32_t>(n3ds_Machine_Read(c->bus,cast<uint32_t>((a + cast<uint32_t>(1ULL))))),cast<int64_t>(8ULL)))) | shl<uint32_t>(cast<uint32_t>(n3ds_Machine_Read(c->bus,cast<uint32_t>((a + cast<uint32_t>(2ULL))))),cast<int64_t>(16ULL)))) | shl<uint32_t>(cast<uint32_t>(n3ds_Machine_Read(c->bus,cast<uint32_t>((a + cast<uint32_t>(3ULL))))),cast<int64_t>(24ULL))));
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
n3ds_Machine_Write32(c->wide,a,v);
return ;
}
n3ds_Machine_Write(c->bus,a,cast<uint8_t>(v));
n3ds_Machine_Write(c->bus,cast<uint32_t>((a + cast<uint32_t>(1ULL))),cast<uint8_t>(shr<uint32_t>(v,cast<int64_t>(8ULL))));
n3ds_Machine_Write(c->bus,cast<uint32_t>((a + cast<uint32_t>(2ULL))),cast<uint8_t>(shr<uint32_t>(v,cast<int64_t>(16ULL))));
n3ds_Machine_Write(c->bus,cast<uint32_t>((a + cast<uint32_t>(3ULL))),cast<uint8_t>(shr<uint32_t>(v,cast<int64_t>(24ULL))));
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
return cast<float>((a + b));
}
,[&](double a,double b)->double{
return cast<double>((a + b));
}
);
}
else {
binop([&](float a,float b)->float{
return cast<float>((a - b));
}
,[&](double a,double b)->double{
return cast<double>((a - b));
}
);
}
return true;
break;}
case cast<uint32_t>(2ULL):{
binop([&](float a,float b)->float{
return cast<float>((a * b));
}
,[&](double a,double b)->double{
return cast<double>((a * b));
}
);
if ((op == cast<uint32_t>(1ULL))) {
arm_CPU_vfpNegate(c,vd,single);
}
return true;
break;}
case cast<uint32_t>(4ULL):{
binop([&](float a,float b)->float{
return cast<float>((a / b));
}
,[&](double a,double b)->double{
return cast<double>((a / b));
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
float p = cast<float>((arm_CPU_sGet(c,vn) * arm_CPU_sGet(c,vm)));
if (sub_) {
p = cast<float>(-p);
}
arm_CPU_sSet(c,vd,cast<float>((acc + p)));
}
else {
double acc = arm_CPU_dGet(c,vd);
if (negAcc) {
acc = cast<double>(-acc);
}
double p = cast<double>((arm_CPU_dGet(c,vn) * arm_CPU_dGet(c,vm)));
if (sub_) {
p = cast<double>(-p);
}
arm_CPU_dSet(c,vd,cast<double>((acc + p)));
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
if ((f < cast<double>(0.00000000000000000e+00))) {
f = cast<double>(0.00000000000000000e+00);
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
// tools/platform/n3ds/blz.go:47:1
std::tuple<Slice<uint8_t>,Error> n3ds_DecompressBLZ(Slice<uint8_t> pak){
{
if ((len(pak) < cast<int64_t>(8ULL))) {
return {{},go_fmt_Errorf(std::string("blz: input too short for a footer (%d bytes)",44),len(pak))};
}
int64_t n = len(pak);
uint32_t bufferTopAndBottom = le_Uint32(sub(pak,cast<int64_t>((n - cast<int64_t>(8ULL))),len(pak)));
uint32_t originalBottom = le_Uint32(sub(pak,cast<int64_t>((n - cast<int64_t>(4ULL))),len(pak)));
if ((originalBottom == cast<uint32_t>(0ULL))) {
Slice<uint8_t> out = Slice<uint8_t>::make(n);
gcopy(out,pak);
return {out,{}};
}
int64_t compressedSize = cast<int64_t>(cast<uint32_t>((bufferTopAndBottom & cast<uint32_t>(16777215ULL))));
int64_t footerSize = cast<int64_t>(shr<uint32_t>(bufferTopAndBottom,cast<int64_t>(24ULL)));
if (((footerSize < cast<int64_t>(8ULL)) || (footerSize > cast<int64_t>(11ULL)))) {
return {{},go_fmt_Errorf(std::string("blz: footer size %d out of range 8..11 (bufferTopAndBottom=0x%08x)",66),footerSize,bufferTopAndBottom)};
}
if (((compressedSize < footerSize) || (compressedSize > n))) {
return {{},go_fmt_Errorf(std::string("blz: compressed size 0x%x out of range [0x%x, 0x%x]",51),compressedSize,footerSize,n)};
}
int64_t headSize = cast<int64_t>((n - compressedSize));
int64_t rawLen = cast<int64_t>((n + cast<int64_t>(originalBottom)));
Slice<uint8_t> out = Slice<uint8_t>::make(rawLen);
gcopy(out,pak);
int64_t src = cast<int64_t>((n - footerSize));
int64_t dst = rawLen;
int64_t end = headSize;
uint8_t flags={};
uint8_t mask = cast<uint8_t>(0ULL);
{;for (;(dst > end);){
{
mask = shr<uint8_t>(mask,cast<int64_t>(1ULL));
if ((mask == cast<uint8_t>(0ULL))) {
if ((src <= end)) {
break;
}
src--;
flags = pak[src];
mask = cast<uint8_t>(128ULL);
}
}
if ((cast<uint8_t>((flags & mask)) == cast<uint8_t>(0ULL))) {
if ((src <= end)) {
return {{},go_fmt_Errorf(std::string("blz: literal token underruns the compressed stream at src=0x%x",62),src)};
}
src--;
dst--;
out[dst] = pak[src];
continue;
}
if ((cast<int64_t>((src - cast<int64_t>(2ULL))) < end)) {
return {{},go_fmt_Errorf(std::string("blz: match token underruns the compressed stream at src=0x%x",60),src)};
}
src--;
uint32_t tok = shl<uint32_t>(cast<uint32_t>(pak[src]),cast<int64_t>(8ULL));
src--;
tok |= cast<uint32_t>(pak[src]);
int64_t length = cast<int64_t>((cast<int64_t>((cast<int64_t>(shr<uint32_t>(tok,cast<int64_t>(12ULL))) + cast<int64_t>(2ULL))) + cast<int64_t>(1ULL)));
int64_t dist = cast<int64_t>((cast<int64_t>(cast<uint32_t>((tok & cast<uint32_t>(4095ULL)))) + cast<int64_t>(3ULL)));
if ((cast<int64_t>((dst - end)) < length)) {
length = cast<int64_t>((dst - end));
}
if ((cast<int64_t>((dst + dist)) > rawLen)) {
return {{},go_fmt_Errorf(std::string("blz: match distance %d at dst=0x%x reads past the buffer end",60),dist,dst)};
}
{;for (;(length > cast<int64_t>(0ULL));length--){
dst--;
out[dst] = out[cast<int64_t>((dst + dist))];
}
}}
}if ((dst != end)) {
return {{},go_fmt_Errorf(std::string("blz: stream ended with 0x%x bytes still to write (dst=0x%x, head ends at 0x%x)",78),cast<int64_t>((dst - end)),dst,end)};
}
return {out,{}};
}
}
// tools/platform/n3ds/cgfx_texture.go:173:1
image_NRGBA* n3ds_decodeETC1(Slice<uint8_t> data,int64_t w,int64_t h){
{
image_NRGBA* img = go_image_NewNRGBA(go_image_Rect(cast<int64_t>(0ULL),cast<int64_t>(0ULL),w,h));
int64_t p = cast<int64_t>(0ULL);
{int64_t ty = cast<int64_t>(0ULL);for (;(ty < h);ty += cast<int64_t>(8ULL)){
{int64_t tx = cast<int64_t>(0ULL);for (;(tx < w);tx += cast<int64_t>(8ULL)){
{auto&& tmp1 = std::array<std::array<int64_t,2>,4>{std::array<int64_t,2>{cast<int64_t>(0ULL),cast<int64_t>(0ULL)},std::array<int64_t,2>{cast<int64_t>(4ULL),cast<int64_t>(0ULL)},std::array<int64_t,2>{cast<int64_t>(0ULL),cast<int64_t>(4ULL)},std::array<int64_t,2>{cast<int64_t>(4ULL),cast<int64_t>(4ULL)}};
for(int64_t tmp2=0;tmp2<len(tmp1);++tmp2){
auto sub_=tmp1[tmp2];uint64_t v = le_Uint64(sub(data,p,len(data)));
p += cast<int64_t>(8ULL);
n3ds_decodeETC1Block(img,cast<int64_t>((tx + sub_[cast<int64_t>(0ULL)])),cast<int64_t>((ty + sub_[cast<int64_t>(1ULL)])),v);
}}
}
}}
}return img;
}
}
// tools/platform/n3ds/cgfx_texture.go:193:1
image_NRGBA* n3ds_decodeETC1A4(Slice<uint8_t> data,int64_t w,int64_t h){
{
image_NRGBA* img = go_image_NewNRGBA(go_image_Rect(cast<int64_t>(0ULL),cast<int64_t>(0ULL),w,h));
int64_t p = cast<int64_t>(0ULL);
{int64_t ty = cast<int64_t>(0ULL);for (;(ty < h);ty += cast<int64_t>(8ULL)){
{int64_t tx = cast<int64_t>(0ULL);for (;(tx < w);tx += cast<int64_t>(8ULL)){
{auto&& tmp3 = std::array<std::array<int64_t,2>,4>{std::array<int64_t,2>{cast<int64_t>(0ULL),cast<int64_t>(0ULL)},std::array<int64_t,2>{cast<int64_t>(4ULL),cast<int64_t>(0ULL)},std::array<int64_t,2>{cast<int64_t>(0ULL),cast<int64_t>(4ULL)},std::array<int64_t,2>{cast<int64_t>(4ULL),cast<int64_t>(4ULL)}};
for(int64_t tmp4=0;tmp4<len(tmp3);++tmp4){
auto sub_=tmp3[tmp4];uint64_t alpha = le_Uint64(sub(data,p,len(data)));
uint64_t colour = le_Uint64(sub(data,cast<int64_t>((p + cast<int64_t>(8ULL))),len(data)));
p += cast<int64_t>(16ULL);
auto tmp5 = std::make_tuple(cast<int64_t>((tx + sub_[cast<int64_t>(0ULL)])),cast<int64_t>((ty + sub_[cast<int64_t>(1ULL)])));
int64_t bx = std::get<0>(tmp5);
int64_t by = std::get<1>(tmp5);
n3ds_decodeETC1Block(img,bx,by,colour);
{int64_t x = cast<int64_t>(0ULL);for (;(x < cast<int64_t>(4ULL));x++){
{int64_t y = cast<int64_t>(0ULL);for (;(y < cast<int64_t>(4ULL));y++){
auto tmp6 = std::make_tuple(cast<int64_t>((bx + x)),cast<int64_t>((by + y)));
int64_t px = std::get<0>(tmp6);
int64_t py = std::get<1>(tmp6);
if (((px >= w) || (py >= h))) {
continue;
}
uint8_t a = cast<uint8_t>((cast<uint8_t>(shr<uint64_t>(alpha,cast<uint64_t>(cast<int64_t>((cast<int64_t>(4ULL) * (cast<int64_t>((cast<int64_t>((x * cast<int64_t>(4ULL))) + y)))))))) & cast<uint8_t>(15ULL)));
img->Pix[cast<int64_t>((image_NRGBA_PixOffset(img,px,py) + cast<int64_t>(3ULL)))] = cast<uint8_t>((a * cast<uint8_t>(17ULL)));
}
}}
}}}
}
}}
}return img;
}
}
// tools/platform/n3ds/cgfx_texture.go:224:1
void n3ds_decodeETC1Block(image_NRGBA* img,int64_t bx,int64_t by,uint64_t v){
{
std::array<uint8_t,8> b = std::array<uint8_t,8>{};
{auto&& tmp7 = b;
for(int64_t tmp8=0;tmp8<len(tmp7);++tmp8){
auto i=tmp8;b[i] = cast<uint8_t>(shr<uint64_t>(v,(cast<int64_t>((cast<int64_t>(56ULL) - cast<int64_t>((cast<int64_t>(8ULL) * i)))))));
}}
bool diff = (cast<uint8_t>((b[cast<int64_t>(3ULL)] & cast<uint8_t>(2ULL))) != cast<uint8_t>(0ULL));
bool flip = (cast<uint8_t>((b[cast<int64_t>(3ULL)] & cast<uint8_t>(1ULL))) != cast<uint8_t>(0ULL));
int64_t table0 = cast<int64_t>(shr<uint8_t>(b[cast<int64_t>(3ULL)],cast<int64_t>(5ULL)));
int64_t table1 = cast<int64_t>(cast<uint8_t>((shr<uint8_t>(b[cast<int64_t>(3ULL)],cast<int64_t>(2ULL)) & cast<uint8_t>(7ULL))));
std::array<std::array<int32_t,3>,2> base={};
if (diff) {
{int64_t ch = cast<int64_t>(0ULL);for (;(ch < cast<int64_t>(3ULL));ch++){
int32_t c5 = cast<int32_t>(shr<uint8_t>(b[ch],cast<int64_t>(3ULL)));
int32_t d3 = cast<int32_t>(cast<uint8_t>((b[ch] & cast<uint8_t>(7ULL))));
if ((d3 >= cast<int32_t>(4ULL))) {
d3 -= cast<int32_t>(8ULL);
}
int32_t c2 = cast<int32_t>((c5 + d3));
base[cast<int64_t>(0ULL)][ch] = cast<int32_t>((shl<int32_t>(c5,cast<int64_t>(3ULL)) | shr<int32_t>(c5,cast<int64_t>(2ULL))));
base[cast<int64_t>(1ULL)][ch] = cast<int32_t>((shl<int32_t>(c2,cast<int64_t>(3ULL)) | shr<int32_t>(c2,cast<int64_t>(2ULL))));
}
}}
else {
{int64_t ch = cast<int64_t>(0ULL);for (;(ch < cast<int64_t>(3ULL));ch++){
int32_t c0 = cast<int32_t>(shr<uint8_t>(b[ch],cast<int64_t>(4ULL)));
int32_t c1 = cast<int32_t>(cast<uint8_t>((b[ch] & cast<uint8_t>(15ULL))));
base[cast<int64_t>(0ULL)][ch] = cast<int32_t>((cast<int32_t>((c0 * cast<int32_t>(16ULL))) + c0));
base[cast<int64_t>(1ULL)][ch] = cast<int32_t>((cast<int32_t>((c1 * cast<int32_t>(16ULL))) + c1));
}
}}
uint32_t msb = cast<uint32_t>((shl<uint32_t>(cast<uint32_t>(b[cast<int64_t>(4ULL)]),cast<int64_t>(8ULL)) | cast<uint32_t>(b[cast<int64_t>(5ULL)])));
uint32_t lsb = cast<uint32_t>((shl<uint32_t>(cast<uint32_t>(b[cast<int64_t>(6ULL)]),cast<int64_t>(8ULL)) | cast<uint32_t>(b[cast<int64_t>(7ULL)])));
{int64_t x = cast<int64_t>(0ULL);for (;(x < cast<int64_t>(4ULL));x++){
{int64_t y = cast<int64_t>(0ULL);for (;(y < cast<int64_t>(4ULL));y++){
uint64_t i = cast<uint64_t>(cast<int64_t>((cast<int64_t>((x * cast<int64_t>(4ULL))) + y)));
int64_t sub_ = cast<int64_t>(0ULL);
if (((((!flip) && (x >= cast<int64_t>(2ULL)))) || ((flip && (y >= cast<int64_t>(2ULL)))))) {
sub_ = cast<int64_t>(1ULL);
}
int64_t table = table0;
if ((sub_ == cast<int64_t>(1ULL))) {
table = table1;
}
int32_t mag = n3ds_etc1Modifiers[table][cast<uint32_t>((shr<uint32_t>(lsb,i) & cast<uint32_t>(1ULL)))];
if ((cast<uint32_t>((shr<uint32_t>(msb,i) & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL))) {
mag = cast<int32_t>(-mag);
}
auto tmp9 = std::make_tuple(cast<int64_t>((bx + x)),cast<int64_t>((by + y)));
int64_t px = std::get<0>(tmp9);
int64_t py = std::get<1>(tmp9);
if (((px >= img->Rect.Max.X) || (py >= img->Rect.Max.Y))) {
continue;
}
int64_t o = image_NRGBA_PixOffset(img,px,py);
{int64_t ch = cast<int64_t>(0ULL);for (;(ch < cast<int64_t>(3ULL));ch++){
int32_t c = cast<int32_t>((base[sub_][ch] + mag));
if ((c < cast<int32_t>(0ULL))) {
c = cast<int32_t>(0ULL);
}
if ((c > cast<int32_t>(255ULL))) {
c = cast<int32_t>(255ULL);
}
img->Pix[cast<int64_t>((o + ch))] = cast<uint8_t>(c);
}
}img->Pix[cast<int64_t>((o + cast<int64_t>(3ULL)))] = cast<uint8_t>(255ULL);
}
}}
}}
}
// tools/platform/n3ds/cp15.go:13:1
void n3ds_Machine_handleCP15(n3ds_Machine* m,arm_CPU* c,bool load,uint32_t cp,uint32_t op1,uint32_t crn,uint32_t crm,uint32_t op2,uint32_t* rd){
{
if ((cp != cast<uint32_t>(15ULL))) {
return ;
}
if ((((crn == cast<uint32_t>(13ULL)) && (crm == cast<uint32_t>(0ULL))) && (op2 == cast<uint32_t>(3ULL)))) {
if (load) {
(*rd) = m->curThread->tlsBase;
}
return ;
}
if ((((crn == cast<uint32_t>(13ULL)) && (crm == cast<uint32_t>(0ULL))) && (((op2 == cast<uint32_t>(2ULL)) || (op2 == cast<uint32_t>(4ULL)))))) {
if (load) {
(*rd) = m->curThread->tpidr;
}
else {
m->curThread->tpidr = (*rd);
}
return ;
}
if (load) {
(*rd) = cast<uint32_t>(0ULL);
}
}
}
// tools/platform/n3ds/debug.go:36:1
void n3ds_GPU_pixelEvent(n3ds_GPU* g,uint32_t x,uint32_t y,n3ds_PixelEvent ev){
{
if (bool(g->m->OnPixel)) {
g->m->OnPixel(x,y,ev);
}
}
}
// tools/platform/n3ds/debug.go:49:1
int64_t n3ds_Machine_RunStopAfterPICACommand(n3ds_Machine* m,int64_t k,int64_t budget){
{
auto tmp10 = std::make_tuple(k,cast<int64_t>(0ULL));
m->picaLimit = std::get<0>(tmp10);
m->picaCount = std::get<1>(tmp10);
int64_t n = n3ds_Machine_Run(m,budget);
auto tmp11 = std::make_tuple(cast<int64_t>(0ULL),cast<int64_t>(0ULL));
m->picaLimit = std::get<0>(tmp11);
m->picaCount = std::get<1>(tmp11);
return n;
}
}
// tools/platform/n3ds/debug.go:58:1
int64_t n3ds_Machine_PICACommands(n3ds_Machine* m){
{
return m->picaCount;
}
}
// tools/platform/n3ds/debug.go:62:1
void n3ds_Machine_ClearBreakpoints(n3ds_Machine* m){
{
m->bps = Map<uint32_t,bool>{};
}
}
// tools/platform/n3ds/debug.go:65:1
bool n3ds_Machine_Stopped(n3ds_Machine* m){
{
return m->stopped;
}
}
// tools/platform/n3ds/debug.go:78:1
Slice<n3ds_MemRegion> n3ds_Machine_MemRegions(n3ds_Machine* m){
{
Slice<n3ds_MemRegion> out = Slice<n3ds_MemRegion>::make(cast<int64_t>(0ULL),len(m->regions));
{auto&& tmp12 = m->regions;
for(int64_t tmp13=0;tmp13<len(tmp12);++tmp13){
auto r=tmp12[tmp13];out = append(out,n3ds_MemRegion{r->name,r->base,cast<uint32_t>(len(r->data))});
}}
return out;
}
}
// tools/platform/n3ds/debug.go:88:1
n3ds_RomFS* n3ds_Machine_RomFS(n3ds_Machine* m){
{
return m->romfs;
}
}
// tools/platform/n3ds/debug.go:95:1
std::tuple<uint32_t,uint32_t,uint32_t> n3ds_GPU_ColorTarget(n3ds_GPU* g){
uint32_t addr{};
uint32_t w{};
uint32_t h{};
{
n3ds_fbState fb = n3ds_GPU_fbstate(g);
return {fb.colorAddr,fb.width,fb.height};
}
}
// tools/platform/n3ds/debug.go:101:1
std::tuple<uint32_t,uint32_t,uint32_t> n3ds_GPU_DepthTarget(n3ds_GPU* g){
uint32_t addr{};
uint32_t w{};
uint32_t h{};
{
n3ds_fbState fb = n3ds_GPU_fbstate(g);
return {fb.depthAddr,fb.width,fb.height};
}
}
// tools/platform/n3ds/debug.go:110:1
image_NRGBA* n3ds_Machine_RenderDepth(n3ds_Machine* m,uint32_t addr,uint32_t w,uint32_t h){
{
image_NRGBA* img = go_image_NewNRGBA(go_image_Rect(cast<int64_t>(0ULL),cast<int64_t>(0ULL),cast<int64_t>(w),cast<int64_t>(h)));
{uint32_t y = cast<uint32_t>(0ULL);for (;(y < h);y++){
{uint32_t x = cast<uint32_t>(0ULL);for (;(x < w);x++){
uint32_t p = cast<uint32_t>((addr + n3ds_tiledOffset(x,y,w)));
uint32_t d = cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(n3ds_Machine_Read(m,p)) | shl<uint32_t>(cast<uint32_t>(n3ds_Machine_Read(m,cast<uint32_t>((p + cast<uint32_t>(1ULL))))),cast<int64_t>(8ULL)))) | shl<uint32_t>(cast<uint32_t>(n3ds_Machine_Read(m,cast<uint32_t>((p + cast<uint32_t>(2ULL))))),cast<int64_t>(16ULL))));
uint8_t v = cast<uint8_t>(shr<uint32_t>(d,cast<int64_t>(16ULL)));
int64_t o = image_NRGBA_PixOffset(img,cast<int64_t>(x),cast<int64_t>(y));
auto tmp14 = std::make_tuple(v,v,v,cast<uint8_t>(255ULL));
img->Pix[o] = std::get<0>(tmp14);
img->Pix[cast<int64_t>((o + cast<int64_t>(1ULL)))] = std::get<1>(tmp14);
img->Pix[cast<int64_t>((o + cast<int64_t>(2ULL)))] = std::get<2>(tmp14);
img->Pix[cast<int64_t>((o + cast<int64_t>(3ULL)))] = std::get<3>(tmp14);
}
}}
}return img;
}
}
// tools/platform/n3ds/debug.go:132:1
Slice<std::string> n3ds_TextureFormats(){
{
return append(Slice<std::string>{},n3ds_textureFormats);
}
}
// tools/platform/n3ds/debug.go:135:1
std::tuple<uint32_t,bool> n3ds_TextureFormat(std::string name){
{
{auto&& tmp15 = n3ds_textureFormats;
for(int64_t tmp16=0;tmp16<len(tmp15);++tmp16){
auto i=tmp16;auto f=tmp15[tmp16];if ((f == name)) {
return {cast<uint32_t>(i),true};
}
}}
return {cast<uint32_t>(0ULL),false};
}
}
// tools/platform/n3ds/debug.go:155:1
std::tuple<image_NRGBA*,Error> n3ds_Machine_RenderTexture(n3ds_Machine* m,uint32_t addr,uint32_t format,uint32_t w,uint32_t h){
{
if ((cast<int64_t>(format) >= len(n3ds_textureFormats))) {
return {{},go_fmt_Errorf(std::string("n3ds: texture format 0x%X is not a PICA format",46),format)};
}
if (((((w == cast<uint32_t>(0ULL)) || (h == cast<uint32_t>(0ULL))) || (w > cast<uint32_t>(2048ULL))) || (h > cast<uint32_t>(2048ULL)))) {
return {{},go_fmt_Errorf(std::string("n3ds: texture size %dx%d out of range",37),w,h)};
}
n3ds_GPU* g = m->gpu;
n3ds_texKey k = n3ds_texKey{addr,format,w,h};
auto tmp17 = lookup(g->texCache,k);
bool cached = std::get<1>(tmp17);
auto tmp18 = n3ds_GPU_texture(g,addr,format,w,h);
n3ds_texImage* tex = std::get<0>(tmp18);
bool ok = std::get<1>(tmp18);
if ((!cached)) {
removeKey(g->texCache,k);
}
if ((!ok)) {
return {{},go_fmt_Errorf(std::string("n3ds: cannot decode a %s texture at 0x%08X",42),n3ds_textureFormats[format],addr)};
}
image_NRGBA* img = go_image_NewNRGBA(go_image_Rect(cast<int64_t>(0ULL),cast<int64_t>(0ULL),cast<int64_t>(w),cast<int64_t>(h)));
gcopy(img->Pix,tex->pix);
return {img,{}};
}
}
// tools/platform/n3ds/dsp.go:142:1
bool n3ds_Machine_ipcDSP(n3ds_Machine* m,n3ds_ipcHeader hdr){
{
{
switch(hdr.Command){
case cast<uint16_t>(1ULL):{
uint32_t v = cast<uint32_t>(1ULL);
if ((m->dsp.State == cast<uint32_t>(1ULL))) {
v = cast<uint32_t>(0ULL);
}
n3ds_Machine_ipcReply(m,hdr.Command,Slice<uint32_t>{v});
return true;
break;}
case cast<uint16_t>(2ULL):{
n3ds_Machine_ipcReply(m,hdr.Command,Slice<uint32_t>{cast<uint32_t>(1ULL)});
return true;
break;}
case cast<uint16_t>(7ULL):{
m->dsp.Semaphore = cast<uint32_t>((n3ds_Machine_ipcArg(m,cast<int64_t>(1ULL)) & cast<uint32_t>(65535ULL)));
n3ds_Machine_ipcReply(m,hdr.Command,Slice<uint32_t>{});
return true;
break;}
case cast<uint16_t>(12ULL):{
n3ds_Machine_ipcReply(m,hdr.Command,Slice<uint32_t>{cast<uint32_t>((cast<uint32_t>(((shl<uint32_t>(n3ds_Machine_ipcArg(m,cast<int64_t>(1ULL)),cast<int64_t>(1ULL))) + cast<uint32_t>(535822336ULL))) + cast<uint32_t>(262144ULL)))});
return true;
break;}
case cast<uint16_t>(13ULL):{
return n3ds_Machine_dspWritePipe(m,hdr);
break;}
case cast<uint16_t>(14ULL):{
auto tmp19 = std::make_tuple(cast<uint32_t>((n3ds_Machine_ipcArg(m,cast<int64_t>(1ULL)) & cast<uint32_t>(7ULL))),cast<uint32_t>((n3ds_Machine_ipcArg(m,cast<int64_t>(3ULL)) & cast<uint32_t>(65535ULL))));
uint32_t ch = std::get<0>(tmp19);
uint32_t size = std::get<1>(tmp19);
if ((cast<uint32_t>(len(m->dsp.Pipes[ch])) < size)) {
arm_CPU_Halt(m->CPU,std::string("dsp ReadPipe channel %d wants 0x%X bytes, pipe holds 0x%X (hardware would hang) at 0x%08X",89),ch,size,len(m->dsp.Pipes[ch]),arm_CPU_PC(m->CPU));
return true;
}
n3ds_Machine_dspReplyPipeData(m,hdr,ch,size,false);
return true;
break;}
case cast<uint16_t>(15ULL):{
n3ds_Machine_ipcReply(m,hdr.Command,Slice<uint32_t>{cast<uint32_t>(len(m->dsp.Pipes[cast<uint32_t>((n3ds_Machine_ipcArg(m,cast<int64_t>(1ULL)) & cast<uint32_t>(7ULL)))]))});
return true;
break;}
case cast<uint16_t>(16ULL):{
n3ds_Machine_dspReplyPipeData(m,hdr,cast<uint32_t>((n3ds_Machine_ipcArg(m,cast<int64_t>(1ULL)) & cast<uint32_t>(7ULL))),cast<uint32_t>((n3ds_Machine_ipcArg(m,cast<int64_t>(3ULL)) & cast<uint32_t>(65535ULL))),true);
return true;
break;}
case cast<uint16_t>(17ULL):{
m->dsp.ComponentLoaded = true;
m->dsp.ComponentSize = n3ds_Machine_ipcArg(m,cast<int64_t>(1ULL));
m->dsp.NextFrame = cast<uint64_t>((m->instrs + cast<uint64_t>(1310720ULL)));
if (m->Verbose) {
go_fmt_Printf(std::string("    dsp LoadComponent size=0x%X progMask=0x%X dataMask=0x%X\012",60),n3ds_Machine_ipcArg(m,cast<int64_t>(1ULL)),n3ds_Machine_ipcArg(m,cast<int64_t>(2ULL)),n3ds_Machine_ipcArg(m,cast<int64_t>(3ULL)));
}
auto tmp20 = std::make_tuple(n3ds_Machine_ipcArg(m,cast<int64_t>(4ULL)),n3ds_Machine_ipcArg(m,cast<int64_t>(5ULL)));
uint32_t desc = std::get<0>(tmp20);
uint32_t ptr = std::get<1>(tmp20);
n3ds_Machine_WriteWord(m,n3ds_Machine_cmdBuf(m),cast<uint32_t>((cast<uint32_t>((shl<uint32_t>(cast<uint32_t>(hdr.Command),cast<int64_t>(16ULL)) | cast<uint32_t>(128ULL))) | cast<uint32_t>(2ULL))));
n3ds_Machine_WriteWord(m,cast<uint32_t>((n3ds_Machine_cmdBuf(m) + cast<uint32_t>(4ULL))),cast<uint32_t>(0ULL));
n3ds_Machine_WriteWord(m,cast<uint32_t>((n3ds_Machine_cmdBuf(m) + cast<uint32_t>(8ULL))),cast<uint32_t>(1ULL));
n3ds_Machine_WriteWord(m,cast<uint32_t>((n3ds_Machine_cmdBuf(m) + cast<uint32_t>(12ULL))),desc);
n3ds_Machine_WriteWord(m,cast<uint32_t>((n3ds_Machine_cmdBuf(m) + cast<uint32_t>(16ULL))),ptr);
return true;
break;}
case cast<uint16_t>(18ULL):{
m->dsp.ComponentLoaded = false;
m->dsp.State = cast<uint32_t>(0ULL);
n3ds_Machine_ipcReply(m,hdr.Command,Slice<uint32_t>{});
return true;
break;}
case cast<uint16_t>(19ULL):case cast<uint16_t>(20ULL):{
n3ds_Machine_ipcReply(m,hdr.Command,Slice<uint32_t>{});
return true;
break;}
case cast<uint16_t>(21ULL):{
auto tmp21 = std::make_tuple(n3ds_Machine_ipcArg(m,cast<int64_t>(1ULL)),n3ds_Machine_ipcArg(m,cast<int64_t>(2ULL)),n3ds_Machine_ipcArg(m,cast<int64_t>(4ULL)));
uint32_t intr = std::get<0>(tmp21);
uint32_t ch = std::get<1>(tmp21);
uint32_t ev = std::get<2>(tmp21);
uint32_t key = cast<uint32_t>((shl<uint32_t>(intr,cast<int64_t>(8ULL)) | ch));
if ((ev == cast<uint32_t>(0ULL))) {
removeKey(m->dsp.IntEvents,key);
}
else {
m->dsp.IntEvents[key] = ev;
}
if (m->Verbose) {
go_fmt_Printf(std::string("    dsp RegisterInterruptEvents interrupt=%d channel=%d event=0x%08X\012",69),intr,ch,ev);
}
n3ds_Machine_ipcReply(m,hdr.Command,Slice<uint32_t>{});
return true;
break;}
case cast<uint16_t>(22ULL):{
if ((m->dsp.SemEvent == cast<uint32_t>(0ULL))) {
m->dsp.SemEvent = n3ds_Machine_newHandle(m,std::string("event",5),false);
m->handles[m->dsp.SemEvent]->name = std::string("dsp-semaphore",13);
}
n3ds_Machine_WriteWord(m,n3ds_Machine_cmdBuf(m),cast<uint32_t>((cast<uint32_t>((shl<uint32_t>(cast<uint32_t>(hdr.Command),cast<int64_t>(16ULL)) | cast<uint32_t>(64ULL))) | cast<uint32_t>(2ULL))));
n3ds_Machine_WriteWord(m,cast<uint32_t>((n3ds_Machine_cmdBuf(m) + cast<uint32_t>(4ULL))),cast<uint32_t>(0ULL));
n3ds_Machine_WriteWord(m,cast<uint32_t>((n3ds_Machine_cmdBuf(m) + cast<uint32_t>(8ULL))),cast<uint32_t>(0ULL));
n3ds_Machine_WriteWord(m,cast<uint32_t>((n3ds_Machine_cmdBuf(m) + cast<uint32_t>(12ULL))),m->dsp.SemEvent);
return true;
break;}
case cast<uint16_t>(23ULL):{
m->dsp.SemMask = cast<uint32_t>((n3ds_Machine_ipcArg(m,cast<int64_t>(1ULL)) & cast<uint32_t>(65535ULL)));
n3ds_Machine_ipcReply(m,hdr.Command,Slice<uint32_t>{});
return true;
break;}
case cast<uint16_t>(31ULL):{
n3ds_Machine_ipcReply(m,hdr.Command,Slice<uint32_t>{cast<uint32_t>(0ULL)});
return true;
break;}
case cast<uint16_t>(32ULL):{
n3ds_Machine_ipcReply(m,hdr.Command,Slice<uint32_t>{});
return true;
break;}
}}
arm_CPU_Halt(m->CPU,std::string("dsp command 0x%04X unimplemented at 0x%08X after %d instructions",64),hdr.Command,arm_CPU_PC(m->CPU),m->CPU->Instrs);
return true;
}
}
// tools/platform/n3ds/dsp.go:265:1
bool n3ds_Machine_dspWritePipe(n3ds_Machine* m,n3ds_ipcHeader hdr){
{
auto tmp22 = std::make_tuple(n3ds_Machine_ipcArg(m,cast<int64_t>(1ULL)),n3ds_Machine_ipcArg(m,cast<int64_t>(2ULL)),n3ds_Machine_ipcArg(m,cast<int64_t>(4ULL)));
uint32_t ch = std::get<0>(tmp22);
uint32_t size = std::get<1>(tmp22);
uint32_t ptr = std::get<2>(tmp22);
if ((ch != cast<uint32_t>(2ULL))) {
arm_CPU_Halt(m->CPU,std::string("dsp WriteProcessPipe to unmodelled pipe %d (size 0x%X) at 0x%08X after %d instructions",86),ch,size,arm_CPU_PC(m->CPU),m->CPU->Instrs);
return true;
}
if ((size < cast<uint32_t>(4ULL))) {
arm_CPU_Halt(m->CPU,std::string("dsp WriteProcessPipe audio: %d-byte write, want 4 at 0x%08X",59),size,arm_CPU_PC(m->CPU));
return true;
}
uint8_t state = n3ds_Machine_Read(m,ptr);
if (m->Verbose) {
go_fmt_Printf(std::string("    dsp WriteProcessPipe audio state-change=%d\012",47),state);
}
{
switch(state){
case cast<uint8_t>(0ULL):case cast<uint8_t>(2ULL):{
m->dsp.Pipes = std::array<Slice<uint8_t>,8>{};
n3ds_Machine_dspAnnounceStructs(m);
m->dsp.State = cast<uint32_t>(1ULL);
break;}
case cast<uint8_t>(1ULL):{
m->dsp.State = cast<uint32_t>(0ULL);
break;}
case cast<uint8_t>(3ULL):{
n3ds_Machine_dspAnnounceStructs(m);
m->dsp.State = cast<uint32_t>(2ULL);
break;}
default:{
arm_CPU_Halt(m->CPU,std::string("dsp audio pipe: unknown state change %d at 0x%08X",49),state,arm_CPU_PC(m->CPU));
return true;
break;}
}}
n3ds_Machine_ipcReply(m,hdr.Command,Slice<uint32_t>{});
return true;
}
}
// tools/platform/n3ds/dsp.go:307:1
void n3ds_Machine_dspAnnounceStructs(n3ds_Machine* m){
{
std::array<uint32_t,15> offs = std::array<uint32_t,15>{cast<uint32_t>(32766ULL),cast<uint32_t>(15652ULL),cast<uint32_t>(3328ULL),cast<uint32_t>(20260ULL),cast<uint32_t>(10336ULL),cast<uint32_t>(2048ULL),cast<uint32_t>(2688ULL),cast<uint32_t>(10532ULL),cast<uint32_t>(3616ULL),cast<uint32_t>(2080ULL),cast<uint32_t>(21028ULL),cast<uint32_t>(21540ULL),cast<uint32_t>(21924ULL),cast<uint32_t>(22692ULL),cast<uint32_t>(22712ULL)};
Slice<uint8_t> p = Slice<uint8_t>::make(cast<int64_t>(0ULL),cast<int64_t>(32ULL));
p = append(p,cast<uint8_t>(15ULL),cast<uint8_t>(0ULL));
{auto&& tmp23 = offs;
for(int64_t tmp24=0;tmp24<len(tmp23);++tmp24){
auto off=tmp23[tmp24];uint32_t w = cast<uint32_t>((cast<uint32_t>(32768ULL) + divi<uint32_t>(off,cast<uint32_t>(2ULL))));
p = append(p,cast<uint8_t>(w),cast<uint8_t>(shr<uint32_t>(w,cast<int64_t>(8ULL))));
}}
m->dsp.Pipes[cast<int64_t>(2ULL)] = append(m->dsp.Pipes[cast<int64_t>(2ULL)],p);
n3ds_Machine_dspSignalInterrupt(m,cast<uint32_t>(2ULL),cast<uint32_t>(2ULL));
}
}
// tools/platform/n3ds/dsp.go:329:1
void n3ds_Machine_dspReplyPipeData(n3ds_Machine* m,n3ds_ipcHeader hdr,uint32_t ch,uint32_t size,bool ifPossible){
{
Slice<uint8_t> out={};
{
Slice<uint8_t> data = m->dsp.Pipes[ch];
if ((cast<uint32_t>(len(data)) >= size)) {
out = sub(data,0,size);
m->dsp.Pipes[ch] = sub(data,size,len(data));
}
}
uint32_t ptr = n3ds_Machine_ReadWord(m,cast<uint32_t>((m->curThread->tlsBase + cast<uint32_t>(388ULL))));
{
uint32_t max = shr<uint32_t>(n3ds_Machine_ReadWord(m,cast<uint32_t>((m->curThread->tlsBase + cast<uint32_t>(384ULL)))),cast<int64_t>(14ULL));
if ((cast<uint32_t>(len(out)) > max)) {
out = sub(out,0,max);
}
}
{auto&& tmp25 = out;
for(int64_t tmp26=0;tmp26<len(tmp25);++tmp26){
auto i=tmp26;auto b=tmp25[tmp26];n3ds_Machine_Write(m,cast<uint32_t>((ptr + cast<uint32_t>(i))),b);
}}
uint32_t n = cast<uint32_t>(len(out));
if (ifPossible) {
n3ds_Machine_WriteWord(m,n3ds_Machine_cmdBuf(m),cast<uint32_t>((cast<uint32_t>((shl<uint32_t>(cast<uint32_t>(hdr.Command),cast<int64_t>(16ULL)) | cast<uint32_t>(128ULL))) | cast<uint32_t>(2ULL))));
n3ds_Machine_WriteWord(m,cast<uint32_t>((n3ds_Machine_cmdBuf(m) + cast<uint32_t>(4ULL))),cast<uint32_t>(0ULL));
n3ds_Machine_WriteWord(m,cast<uint32_t>((n3ds_Machine_cmdBuf(m) + cast<uint32_t>(8ULL))),n);
n3ds_Machine_WriteWord(m,cast<uint32_t>((n3ds_Machine_cmdBuf(m) + cast<uint32_t>(12ULL))),cast<uint32_t>((shl<uint32_t>(n,cast<int64_t>(14ULL)) | cast<uint32_t>(2ULL))));
n3ds_Machine_WriteWord(m,cast<uint32_t>((n3ds_Machine_cmdBuf(m) + cast<uint32_t>(16ULL))),ptr);
return ;
}
n3ds_Machine_WriteWord(m,n3ds_Machine_cmdBuf(m),cast<uint32_t>((cast<uint32_t>((shl<uint32_t>(cast<uint32_t>(hdr.Command),cast<int64_t>(16ULL)) | cast<uint32_t>(64ULL))) | cast<uint32_t>(2ULL))));
n3ds_Machine_WriteWord(m,cast<uint32_t>((n3ds_Machine_cmdBuf(m) + cast<uint32_t>(4ULL))),cast<uint32_t>(0ULL));
n3ds_Machine_WriteWord(m,cast<uint32_t>((n3ds_Machine_cmdBuf(m) + cast<uint32_t>(8ULL))),cast<uint32_t>((shl<uint32_t>(n,cast<int64_t>(14ULL)) | cast<uint32_t>(2ULL))));
n3ds_Machine_WriteWord(m,cast<uint32_t>((n3ds_Machine_cmdBuf(m) + cast<uint32_t>(12ULL))),ptr);
}
}
// tools/platform/n3ds/dsp.go:362:1
bool n3ds_Machine_dspDue(n3ds_Machine* m){
{
return (m->dsp.ComponentLoaded && (m->instrs >= m->dsp.NextFrame));
}
}
// tools/platform/n3ds/dsp.go:368:1
std::tuple<uint64_t,bool> n3ds_Machine_dspDeadline(n3ds_Machine* m){
{
if ((!m->dsp.ComponentLoaded)) {
return {cast<uint64_t>(0ULL),false};
}
return {m->dsp.NextFrame,true};
}
}
// tools/platform/n3ds/dsp.go:380:1
void n3ds_Machine_dspTick(n3ds_Machine* m){
rrprof::Scope profile(4,"DSP HLE mixer");{
auto tmp27=defer([&](){n3ds_Machine_profEnd(m,cast<int64_t>(5ULL),n3ds_Machine_profStart(m));});
m->dsp.NextFrame = cast<uint64_t>((m->instrs + cast<uint64_t>(1310720ULL)));
m->dsp.Ticks++;
if ((m->dsp.State == cast<uint32_t>(1ULL))) {
auto tmp28 = std::make_tuple(n3ds_Machine_dspReadRegion(m),n3ds_Machine_dspWriteRegion(m));
uint32_t read = std::get<0>(tmp28);
uint32_t write = std::get<1>(tmp28);
std::array<n3ds_dspQuadFrame,3> mixes={};
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(24ULL));i++){
n3ds_Machine_dspParseConfig(m,i,read);
if (m->dsp.Sources[i].Enabled) {
n3ds_Machine_dspSourceFrame(m,i);
}
n3ds_Machine_dspWriteStatus(m,i,write);
{int64_t mix = cast<int64_t>(0ULL);for (;(mix < cast<int64_t>(3ULL));mix++){
n3ds_dspSource_mixInto(&(m->dsp.Sources[i]),(&mixes[mix]),mix);
}
}}
}n3ds_Machine_dspMixerConfig(m,read);
n3ds_dspFrame final = n3ds_Machine_dspMix(m,read,write,(&mixes));
n3ds_Machine_dspWriteFinal(m,write,(&final));
n3ds_Machine_WriteWord(m,cast<uint32_t>((write + cast<uint32_t>(2048ULL))),cast<uint32_t>(0ULL));
n3ds_Machine_dspSignalInterrupt(m,cast<uint32_t>(2ULL),cast<uint32_t>(2ULL));
}
if ((m->dsp.SemEvent != cast<uint32_t>(0ULL))) {
n3ds_Machine_dspSignalHandle(m,m->dsp.SemEvent);
}
}
}
// tools/platform/n3ds/dsp.go:426:1
uint32_t n3ds_Machine_dspReadRegion(n3ds_Machine* m){
{
uint16_t c0 = n3ds_Machine_dspRead16(m,cast<uint32_t>(536182782ULL));
uint16_t c1 = n3ds_Machine_dspRead16(m,cast<uint32_t>(536313854ULL));
{
if (((c0 == cast<uint16_t>(65535ULL)) && (c1 != cast<uint16_t>(65534ULL)))){
return cast<uint32_t>(536281088ULL);
}
else if (((c1 == cast<uint16_t>(65535ULL)) && (c0 != cast<uint16_t>(65534ULL)))){
return cast<uint32_t>(536150016ULL);
}
else if ((c0 > c1)){
return cast<uint32_t>(536150016ULL);
}
}
tmp29:;
return cast<uint32_t>(536281088ULL);
}
}
// tools/platform/n3ds/dsp.go:440:1
uint32_t n3ds_Machine_dspWriteRegion(n3ds_Machine* m){
{
if ((n3ds_Machine_dspReadRegion(m) == cast<uint32_t>(536150016ULL))) {
return cast<uint32_t>(536281088ULL);
}
return cast<uint32_t>(536150016ULL);
}
}
// tools/platform/n3ds/dsp.go:503:1
void n3ds_Machine_dspParseConfig(n3ds_Machine* m,int64_t i,uint32_t region){
{
uint32_t cfg = cast<uint32_t>((cast<uint32_t>((region + cast<uint32_t>(15652ULL))) + cast<uint32_t>((cast<uint32_t>(i) * cast<uint32_t>(192ULL)))));
uint32_t dirty = n3ds_Machine_ReadWord(m,cast<uint32_t>((cfg + cast<uint32_t>(0ULL))));
if ((dirty == cast<uint32_t>(0ULL))) {
return ;
}
n3ds_dspSource* s = (&m->dsp.Sources[i]);
if (m->DSPTrace) {
go_fmt_Printf(std::string("    dsp[%d] cfg r%d(c=%d/%d) dirty=%08X enable=%d sync=%d flags1=%04X len=%d id=%d buffers=%04X (frame %d)\012",107),i,divi<uint32_t>((cast<uint32_t>((region - cast<uint32_t>(536150016ULL)))),cast<uint32_t>(131072ULL)),n3ds_Machine_dspRead16(m,cast<uint32_t>(536182782ULL)),n3ds_Machine_dspRead16(m,cast<uint32_t>(536313854ULL)),dirty,n3ds_Machine_Read(m,cast<uint32_t>((cfg + cast<uint32_t>(160ULL)))),n3ds_Machine_dspRead16(m,cast<uint32_t>((cfg + cast<uint32_t>(162ULL)))),n3ds_Machine_dspRead16(m,cast<uint32_t>((cfg + cast<uint32_t>(180ULL)))),n3ds_Machine_dspRead32(m,cast<uint32_t>((cfg + cast<uint32_t>(176ULL)))),n3ds_Machine_dspRead16(m,cast<uint32_t>((cfg + cast<uint32_t>(190ULL)))),n3ds_Machine_dspRead16(m,cast<uint32_t>((cfg + cast<uint32_t>(74ULL)))),m->dsp.Ticks);
}
if ((cast<uint32_t>((dirty & cast<uint32_t>(536870912ULL))) != cast<uint32_t>(0ULL))) {
(*s) = n3ds_dspSource{{},{},cast<float>(1.00000000000000000e+00),{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{}};
n3ds_dspFilters_reset(&(s->Filters));
}
if ((cast<uint32_t>((dirty & cast<uint32_t>(16ULL))) != cast<uint32_t>(0ULL))) {
s->Queue = {};
}
if ((cast<uint32_t>((dirty & cast<uint32_t>(65536ULL))) != cast<uint32_t>(0ULL))) {
s->Enabled = (n3ds_Machine_Read(m,cast<uint32_t>((cfg + cast<uint32_t>(160ULL)))) != cast<uint8_t>(0ULL));
}
if ((cast<uint32_t>((dirty & cast<uint32_t>(268435456ULL))) != cast<uint32_t>(0ULL))) {
s->SyncCount = n3ds_Machine_dspRead16(m,cast<uint32_t>((cfg + cast<uint32_t>(162ULL))));
}
if ((cast<uint32_t>((dirty & cast<uint32_t>(262144ULL))) != cast<uint32_t>(0ULL))) {
s->Rate = go_math_Float32frombits(n3ds_Machine_ReadWord(m,cast<uint32_t>((cfg + cast<uint32_t>(52ULL)))));
if ((!((s->Rate > cast<float>(0.00000000000000000e+00))))) {
s->Rate = cast<float>(1.00000000000000000e+00);
}
}
if ((cast<uint32_t>((dirty & cast<uint32_t>(131072ULL))) != cast<uint32_t>(0ULL))) {
s->Interp = n3ds_Machine_Read(m,cast<uint32_t>((cfg + cast<uint32_t>(56ULL))));
}
if ((cast<uint32_t>((dirty & cast<uint32_t>(1073741825ULL))) != cast<uint32_t>(0ULL))) {
s->Format = cast<uint8_t>(cast<uint16_t>((shr<uint16_t>(n3ds_Machine_dspRead16(m,cast<uint32_t>((cfg + cast<uint32_t>(180ULL)))),cast<int64_t>(2ULL)) & cast<uint16_t>(3ULL))));
}
if ((cast<uint32_t>((dirty & cast<uint32_t>(1073741826ULL))) != cast<uint32_t>(0ULL))) {
s->Stereo = (cast<uint16_t>((n3ds_Machine_dspRead16(m,cast<uint32_t>((cfg + cast<uint32_t>(180ULL)))) & cast<uint16_t>(3ULL))) == cast<uint16_t>(2ULL));
}
if ((cast<uint32_t>((dirty & cast<uint32_t>(4ULL))) != cast<uint32_t>(0ULL))) {
uint32_t co = cast<uint32_t>((cast<uint32_t>((region + cast<uint32_t>(20260ULL))) + cast<uint32_t>((cast<uint32_t>(i) * cast<uint32_t>(32ULL)))));
{int64_t k = cast<int64_t>(0ULL);for (;(k < cast<int64_t>(16ULL));k++){
s->AdpcmCoeffs[k] = cast<int16_t>(n3ds_Machine_dspRead16(m,cast<uint32_t>((co + cast<uint32_t>((cast<uint32_t>(k) * cast<uint32_t>(2ULL)))))));
}
}}
{int64_t g = cast<int64_t>(0ULL);for (;(g < cast<int64_t>(3ULL));g++){
if ((cast<uint32_t>((dirty & (shl<uint32_t>(cast<uint32_t>(33554432ULL),g)))) == cast<uint32_t>(0ULL))) {
continue;
}
{int64_t c = cast<int64_t>(0ULL);for (;(c < cast<int64_t>(4ULL));c++){
s->Gain[g][c] = n3ds_dspFloat(n3ds_Machine_ReadWord(m,cast<uint32_t>((cast<uint32_t>((cfg + cast<uint32_t>(4ULL))) + cast<uint32_t>((cast<uint32_t>(cast<int64_t>((cast<int64_t>((g * cast<int64_t>(4ULL))) + c))) * cast<uint32_t>(4ULL)))))));
}
}}
}if ((cast<uint32_t>((dirty & cast<uint32_t>(4194304ULL))) != cast<uint32_t>(0ULL))) {
uint16_t en = n3ds_Machine_dspRead16(m,cast<uint32_t>((cfg + cast<uint32_t>(58ULL))));
n3ds_dspFilters_enable(&(s->Filters),(cast<uint16_t>((en & cast<uint16_t>(1ULL))) != cast<uint16_t>(0ULL)),(cast<uint16_t>((en & cast<uint16_t>(2ULL))) != cast<uint16_t>(0ULL)));
}
if ((cast<uint32_t>((dirty & cast<uint32_t>(8388608ULL))) != cast<uint32_t>(0ULL))) {
s->Filters.SB0 = cast<int32_t>(cast<int16_t>(n3ds_Machine_dspRead16(m,cast<uint32_t>((cfg + cast<uint32_t>(60ULL))))));
s->Filters.SA1 = cast<int32_t>(cast<int16_t>(n3ds_Machine_dspRead16(m,cast<uint32_t>((cast<uint32_t>((cfg + cast<uint32_t>(60ULL))) + cast<uint32_t>(2ULL))))));
}
if ((cast<uint32_t>((dirty & cast<uint32_t>(16777216ULL))) != cast<uint32_t>(0ULL))) {
s->Filters.BA2 = cast<int16_t>(n3ds_Machine_dspRead16(m,cast<uint32_t>((cfg + cast<uint32_t>(64ULL)))));
s->Filters.BA1 = cast<int16_t>(n3ds_Machine_dspRead16(m,cast<uint32_t>((cast<uint32_t>((cfg + cast<uint32_t>(64ULL))) + cast<uint32_t>(2ULL)))));
s->Filters.BB2 = cast<int16_t>(n3ds_Machine_dspRead16(m,cast<uint32_t>((cast<uint32_t>((cfg + cast<uint32_t>(64ULL))) + cast<uint32_t>(4ULL)))));
s->Filters.BB1 = cast<int16_t>(n3ds_Machine_dspRead16(m,cast<uint32_t>((cast<uint32_t>((cfg + cast<uint32_t>(64ULL))) + cast<uint32_t>(6ULL)))));
s->Filters.BB0 = cast<int16_t>(n3ds_Machine_dspRead16(m,cast<uint32_t>((cast<uint32_t>((cfg + cast<uint32_t>(64ULL))) + cast<uint32_t>(8ULL)))));
}
uint32_t playPos = cast<uint32_t>(0ULL);
if ((cast<uint32_t>((dirty & cast<uint32_t>(2097152ULL))) != cast<uint32_t>(0ULL))) {
playPos = n3ds_Machine_dspRead32(m,cast<uint32_t>((cfg + cast<uint32_t>(164ULL))));
}
if (((cast<uint32_t>((dirty & cast<uint32_t>(8ULL))) != cast<uint32_t>(0ULL)) && (len(s->CurBuf) > cast<int64_t>(0ULL)))) {
n3ds_dspBuffer buf = n3ds_dspBuffer{s->CurPhysAddr,n3ds_Machine_dspRead32(m,cast<uint32_t>((cfg + cast<uint32_t>(176ULL)))),{},{},{},{},{},s->Stereo,s->Format,{},{},{}};
{
Slice<n3ds_dspSample> pcm = n3ds_Machine_dspDecodeBuffer(m,s,buf);
if (bool(pcm)) {
if ((cast<int64_t>(s->CurSample) < len(pcm))) {
s->CurBuf = sub(pcm,s->CurSample,len(pcm));
}
else {
auto tmp30 = std::make_tuple(cast<uint32_t>(0ULL),pcm);
s->CurSample = std::get<0>(tmp30);
s->CurBuf = std::get<1>(tmp30);
}
}
}
}
if ((cast<uint32_t>((dirty & cast<uint32_t>(1073741824ULL))) != cast<uint32_t>(0ULL))) {
s->Queue = append(s->Queue,n3ds_dspBuffer{n3ds_Machine_dspRead32(m,cast<uint32_t>((cfg + cast<uint32_t>(172ULL)))),n3ds_Machine_dspRead32(m,cast<uint32_t>((cfg + cast<uint32_t>(176ULL)))),cast<uint8_t>(n3ds_Machine_dspRead16(m,cast<uint32_t>((cfg + cast<uint32_t>(182ULL))))),std::array<int16_t,2>{cast<int16_t>(n3ds_Machine_dspRead16(m,cast<uint32_t>((cfg + cast<uint32_t>(184ULL))))),cast<int16_t>(n3ds_Machine_dspRead16(m,cast<uint32_t>((cast<uint32_t>((cfg + cast<uint32_t>(184ULL))) + cast<uint32_t>(2ULL)))))},(cast<uint16_t>((n3ds_Machine_dspRead16(m,cast<uint32_t>((cfg + cast<uint32_t>(188ULL)))) & cast<uint16_t>(1ULL))) != cast<uint16_t>(0ULL)),(cast<uint16_t>((n3ds_Machine_dspRead16(m,cast<uint32_t>((cfg + cast<uint32_t>(188ULL)))) & cast<uint16_t>(2ULL))) != cast<uint16_t>(0ULL)),n3ds_Machine_dspRead16(m,cast<uint32_t>((cfg + cast<uint32_t>(190ULL)))),s->Stereo,s->Format,{},playPos,{}});
}
if ((cast<uint32_t>((dirty & cast<uint32_t>(524288ULL))) != cast<uint32_t>(0ULL))) {
uint16_t mask = n3ds_Machine_dspRead16(m,cast<uint32_t>((cfg + cast<uint32_t>(74ULL))));
if (m->DSPTrace) {
go_fmt_Printf(std::string("    dsp[%d] ENQUEUE r%d mask=%04X ids=[%d %d %d %d] (frame %d)\012",63),i,divi<uint32_t>((cast<uint32_t>((region - cast<uint32_t>(536150016ULL)))),cast<uint32_t>(131072ULL)),mask,n3ds_Machine_dspRead16(m,cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((cfg + cast<uint32_t>(76ULL))) + cast<uint32_t>(0ULL))) + cast<uint32_t>(16ULL)))),n3ds_Machine_dspRead16(m,cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((cfg + cast<uint32_t>(76ULL))) + cast<uint32_t>(20ULL))) + cast<uint32_t>(16ULL)))),n3ds_Machine_dspRead16(m,cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((cfg + cast<uint32_t>(76ULL))) + cast<uint32_t>(40ULL))) + cast<uint32_t>(16ULL)))),n3ds_Machine_dspRead16(m,cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((cfg + cast<uint32_t>(76ULL))) + cast<uint32_t>(60ULL))) + cast<uint32_t>(16ULL)))),m->dsp.Ticks);
}
{uint32_t b = cast<uint32_t>(0ULL);for (;(b < cast<uint32_t>(4ULL));b++){
if ((cast<uint16_t>((mask & (shl<uint16_t>(cast<uint16_t>(1ULL),b)))) == cast<uint16_t>(0ULL))) {
continue;
}
uint32_t bb = cast<uint32_t>((cast<uint32_t>((cfg + cast<uint32_t>(76ULL))) + cast<uint32_t>((b * cast<uint32_t>(20ULL)))));
s->Queue = append(s->Queue,n3ds_dspBuffer{n3ds_Machine_dspRead32(m,cast<uint32_t>((bb + cast<uint32_t>(0ULL)))),n3ds_Machine_dspRead32(m,cast<uint32_t>((bb + cast<uint32_t>(4ULL)))),cast<uint8_t>(n3ds_Machine_dspRead16(m,cast<uint32_t>((bb + cast<uint32_t>(8ULL))))),std::array<int16_t,2>{cast<int16_t>(n3ds_Machine_dspRead16(m,cast<uint32_t>((bb + cast<uint32_t>(10ULL))))),cast<int16_t>(n3ds_Machine_dspRead16(m,cast<uint32_t>((cast<uint32_t>((bb + cast<uint32_t>(10ULL))) + cast<uint32_t>(2ULL)))))},(n3ds_Machine_Read(m,cast<uint32_t>((bb + cast<uint32_t>(14ULL)))) != cast<uint8_t>(0ULL)),(n3ds_Machine_Read(m,cast<uint32_t>((bb + cast<uint32_t>(15ULL)))) != cast<uint8_t>(0ULL)),n3ds_Machine_dspRead16(m,cast<uint32_t>((bb + cast<uint32_t>(16ULL)))),s->Stereo,s->Format,true,{},{}});
}
}n3ds_Machine_dspWrite16(m,cast<uint32_t>((cfg + cast<uint32_t>(74ULL))),cast<uint16_t>(0ULL));
}
n3ds_Machine_WriteWord(m,cast<uint32_t>((cfg + cast<uint32_t>(0ULL))),cast<uint32_t>(0ULL));
}
}
// tools/platform/n3ds/dsp.go:656:1
void n3ds_Machine_dspWriteStatus(n3ds_Machine* m,int64_t i,uint32_t region){
{
uint32_t st = cast<uint32_t>((cast<uint32_t>((region + cast<uint32_t>(3328ULL))) + cast<uint32_t>((cast<uint32_t>(i) * cast<uint32_t>(12ULL)))));
n3ds_dspSource* s = (&m->dsp.Sources[i]);
if ((m->DSPTrace && (((s->Enabled || s->BufferUpdate) || (len(s->Queue) > cast<int64_t>(0ULL)))))) {
go_fmt_Printf(std::string("    dsp[%d] status enabled=%v update=%v sync=%d pos=%d cur=%d last=%d queued=%d (frame %d)\012",91),i,s->Enabled,s->BufferUpdate,s->SyncCount,s->CurSample,s->CurBufferID,s->LastBufferID,len(s->Queue),m->dsp.Ticks);
}
auto tmp31 = std::make_tuple(cast<uint8_t>(0ULL),cast<uint8_t>(0ULL));
uint8_t enabled = std::get<0>(tmp31);
uint8_t update = std::get<1>(tmp31);
if (s->Enabled) {
enabled = cast<uint8_t>(1ULL);
}
if (s->BufferUpdate) {
update = cast<uint8_t>(1ULL);
s->BufferUpdate = false;
}
n3ds_Machine_Write(m,cast<uint32_t>((st + cast<uint32_t>(0ULL))),enabled);
n3ds_Machine_Write(m,cast<uint32_t>((st + cast<uint32_t>(1ULL))),update);
n3ds_Machine_dspWrite16(m,cast<uint32_t>((st + cast<uint32_t>(2ULL))),s->SyncCount);
n3ds_Machine_dspWrite32(m,cast<uint32_t>((st + cast<uint32_t>(4ULL))),s->CurSample);
n3ds_Machine_dspWrite16(m,cast<uint32_t>((st + cast<uint32_t>(8ULL))),s->CurBufferID);
n3ds_Machine_dspWrite16(m,cast<uint32_t>((st + cast<uint32_t>(10ULL))),s->LastBufferID);
}
}
// tools/platform/n3ds/dsp.go:683:1
void n3ds_Machine_dspSignalInterrupt(n3ds_Machine* m,uint32_t interrupt,uint32_t channel){
{
{
auto tmp32 = lookup(m->dsp.IntEvents,cast<uint32_t>((shl<uint32_t>(interrupt,cast<int64_t>(8ULL)) | channel)));
uint32_t h = std::get<0>(tmp32);
bool ok = std::get<1>(tmp32);
if (ok) {
n3ds_Machine_dspSignalHandle(m,h);
}
}
}
}
// tools/platform/n3ds/dsp.go:689:1
void n3ds_Machine_dspSignalHandle(n3ds_Machine* m,uint32_t h){
{
{
n3ds_kobject* obj = get(m->handles,h);
if (bool(obj)) {
obj->signal = true;
if (n3ds_Machine_signalObject(m,obj)) {
m->reschedule = true;
}
}
}
}
}
// tools/platform/n3ds/dsp.go:698:1
uint16_t n3ds_Machine_dspRead16(n3ds_Machine* m,uint32_t a){
{
return cast<uint16_t>((cast<uint16_t>(n3ds_Machine_Read(m,a)) | shl<uint16_t>(cast<uint16_t>(n3ds_Machine_Read(m,cast<uint32_t>((a + cast<uint32_t>(1ULL))))),cast<int64_t>(8ULL))));
}
}
// tools/platform/n3ds/dsp.go:702:1
void n3ds_Machine_dspWrite16(n3ds_Machine* m,uint32_t a,uint16_t v){
{
n3ds_Machine_Write(m,a,cast<uint8_t>(v));
n3ds_Machine_Write(m,cast<uint32_t>((a + cast<uint32_t>(1ULL))),cast<uint8_t>(shr<uint16_t>(v,cast<int64_t>(8ULL))));
}
}
// tools/platform/n3ds/dsp.go:710:1
uint32_t n3ds_Machine_dspRead32(n3ds_Machine* m,uint32_t a){
{
uint32_t v = n3ds_Machine_ReadWord(m,a);
return cast<uint32_t>((shl<uint32_t>(v,cast<int64_t>(16ULL)) | shr<uint32_t>(v,cast<int64_t>(16ULL))));
}
}
// tools/platform/n3ds/dsp.go:715:1
void n3ds_Machine_dspWrite32(n3ds_Machine* m,uint32_t a,uint32_t v){
{
n3ds_Machine_WriteWord(m,a,cast<uint32_t>((shl<uint32_t>(v,cast<int64_t>(16ULL)) | shr<uint32_t>(v,cast<int64_t>(16ULL)))));
}
}
// tools/platform/n3ds/dsp_voice.go:136:1
void n3ds_dspFilters_reset(n3ds_dspFilters* f){
{
(*f) = n3ds_dspFilters{{},{},cast<int32_t>(32768ULL),{},{},{},{},cast<int16_t>(16384ULL),{},{},{},{},{},{}};
}
}
// tools/platform/n3ds/dsp_voice.go:142:1
void n3ds_dspFilters_enable(n3ds_dspFilters* f,bool simple,bool biquad){
{
auto tmp33 = std::make_tuple(simple,biquad);
f->SimpleEnabled = std::get<0>(tmp33);
f->BiquadEnabled = std::get<1>(tmp33);
if ((!simple)) {
auto tmp34 = std::make_tuple(cast<int32_t>(32768ULL),cast<int32_t>(0ULL),n3ds_dspSample{});
f->SB0 = std::get<0>(tmp34);
f->SA1 = std::get<1>(tmp34);
f->SY1 = std::get<2>(tmp34);
}
if ((!biquad)) {
auto tmp35 = std::make_tuple(cast<int16_t>(16384ULL),cast<int16_t>(0ULL),cast<int16_t>(0ULL),cast<int16_t>(0ULL),cast<int16_t>(0ULL));
f->BB0 = std::get<0>(tmp35);
f->BB1 = std::get<1>(tmp35);
f->BB2 = std::get<2>(tmp35);
f->BA1 = std::get<3>(tmp35);
f->BA2 = std::get<4>(tmp35);
auto tmp36 = std::make_tuple(n3ds_dspSample{},n3ds_dspSample{},n3ds_dspSample{},n3ds_dspSample{});
f->BX1 = std::get<0>(tmp36);
f->BX2 = std::get<1>(tmp36);
f->BY1 = std::get<2>(tmp36);
f->BY2 = std::get<3>(tmp36);
}
}
}
// tools/platform/n3ds/dsp_voice.go:153:1
void n3ds_dspFilters_processFrame(n3ds_dspFilters* f,n3ds_dspFrame* frame){
{
if (((!f->SimpleEnabled) && (!f->BiquadEnabled))) {
return ;
}
{auto&& tmp37 = (*frame);
for(int64_t tmp38=0;tmp38<len(tmp37);++tmp38){
auto i=tmp38;n3ds_dspSample x = (*frame)[i];
if (f->SimpleEnabled) {
n3ds_dspSample y={};
{int64_t c = cast<int64_t>(0ULL);for (;(c < cast<int64_t>(2ULL));c++){
y[c] = n3ds_clampS16(shr<int32_t>((cast<int32_t>((cast<int32_t>((f->SB0 * cast<int32_t>(x[c]))) + cast<int32_t>((f->SA1 * cast<int32_t>(f->SY1[c])))))),cast<int64_t>(15ULL)));
}
}auto tmp39 = std::make_tuple(y,y);
f->SY1 = std::get<0>(tmp39);
x = std::get<1>(tmp39);
}
if (f->BiquadEnabled) {
n3ds_dspSample y={};
{int64_t c = cast<int64_t>(0ULL);for (;(c < cast<int64_t>(2ULL));c++){
int32_t acc = cast<int32_t>((cast<int32_t>((cast<int32_t>((cast<int32_t>((cast<int32_t>((cast<int32_t>(f->BB0) * cast<int32_t>(x[c]))) + cast<int32_t>((cast<int32_t>(f->BB1) * cast<int32_t>(f->BX1[c]))))) + cast<int32_t>((cast<int32_t>(f->BB2) * cast<int32_t>(f->BX2[c]))))) + cast<int32_t>((cast<int32_t>(f->BA1) * cast<int32_t>(f->BY1[c]))))) + cast<int32_t>((cast<int32_t>(f->BA2) * cast<int32_t>(f->BY2[c])))));
y[c] = n3ds_clampS16(shr<int32_t>(acc,cast<int64_t>(14ULL)));
}
}auto tmp40 = std::make_tuple(f->BX1,x);
f->BX2 = std::get<0>(tmp40);
f->BX1 = std::get<1>(tmp40);
auto tmp41 = std::make_tuple(f->BY1,y);
f->BY2 = std::get<0>(tmp41);
f->BY1 = std::get<1>(tmp41);
x = y;
}
(*frame)[i] = x;
}}
}
}
// tools/platform/n3ds/dsp_voice.go:181:1
int16_t n3ds_clampS16(int32_t v){
{
{
if ((v > cast<int32_t>(32767ULL))){
return cast<int16_t>(32767ULL);
}
else if ((v < cast<int32_t>(-32768ULL))){
return cast<int16_t>(-32768ULL);
}
}
tmp42:;
return cast<int16_t>(v);
}
}
// tools/platform/n3ds/dsp_voice.go:191:1
int16_t n3ds_addClampS16(int16_t a,int16_t b){
{
return n3ds_clampS16(cast<int32_t>((cast<int32_t>(a) + cast<int32_t>(b))));
}
}
// tools/platform/n3ds/dsp_voice.go:201:1
void n3ds_Machine_dspSourceFrame(n3ds_Machine* m,int64_t i){
{
n3ds_dspSource* s = (&m->dsp.Sources[i]);
s->Frame = n3ds_dspFrame{};
if ((len(s->CurBuf) == cast<int64_t>(0ULL))) {
if (n3ds_Machine_dspDequeue(m,s)) {
return ;
}
s->Enabled = false;
s->BufferUpdate = true;
s->LastBufferID = s->CurBufferID;
s->CurBufferID = cast<uint16_t>(0ULL);
return ;
}
int64_t pos = cast<int64_t>(0ULL);
{int64_t n = cast<int64_t>(0ULL);for (;((pos < cast<int64_t>(160ULL)) && (n < cast<int64_t>(64ULL)));n++){
if (((len(s->CurBuf) == cast<int64_t>(0ULL)) && (!n3ds_Machine_dspDequeue(m,s)))) {
break;
}
n3ds_dspSource_resample(s,(&pos));
}
}s->CurSample += cast<uint32_t>(cast<double>((cast<double>(pos) * cast<double>(s->Rate))));
n3ds_dspFilters_processFrame(&(s->Filters),(&s->Frame));
}
}
// tools/platform/n3ds/dsp_voice.go:233:1
void n3ds_dspSource_resample(n3ds_dspSource* s,int64_t* outi){
{
constexpr int64_t scale=16777216ULL;
if ((len(s->CurBuf) == cast<int64_t>(0ULL))) {
return ;
}
double rate = cast<double>(s->Rate);
if ((!((rate > cast<double>(0.00000000000000000e+00))))) {
rate = cast<double>(1.00000000000000000e+00);
}
Slice<n3ds_dspSample> input = Slice<n3ds_dspSample>::make(cast<int64_t>(0ULL),cast<int64_t>((len(s->CurBuf) + cast<int64_t>(2ULL))));
input = append(input,s->Xn2,s->Xn1);
input = append(input,s->CurBuf);
uint64_t step = cast<uint64_t>(cast<double>((rate * cast<double>(1.67772160000000000e+07))));
uint64_t fpos = s->FPos;
int64_t idx = cast<int64_t>(0ULL);
{;for (;((*outi) < cast<int64_t>(160ULL));){
idx = cast<int64_t>(divi<uint64_t>(fpos,cast<uint64_t>(16777216ULL)));
if ((cast<int64_t>((idx + cast<int64_t>(2ULL))) >= len(input))) {
idx = cast<int64_t>((len(input) - cast<int64_t>(2ULL)));
break;
}
uint64_t frac = cast<uint64_t>((fpos & cast<uint64_t>(16777215ULL)));
s->Frame[(*outi)] = n3ds_interpolate(s->Interp,frac,input[idx],input[cast<int64_t>((idx + cast<int64_t>(1ULL)))]);
(*outi)++;
fpos += step;
}
}auto tmp43 = std::make_tuple(input[idx],input[cast<int64_t>((idx + cast<int64_t>(1ULL)))]);
s->Xn2 = std::get<0>(tmp43);
s->Xn1 = std::get<1>(tmp43);
s->FPos = cast<uint64_t>((fpos - cast<uint64_t>((cast<uint64_t>(idx) * cast<uint64_t>(16777216ULL)))));
s->CurBuf = sub(s->CurBuf,idx,len(s->CurBuf));
}
}
// tools/platform/n3ds/dsp_voice.go:271:1
n3ds_dspSample n3ds_interpolate(uint8_t mode,uint64_t frac,n3ds_dspSample x0,n3ds_dspSample x1){
{
if ((mode == cast<uint8_t>(2ULL))) {
return x0;
}
n3ds_dspSample out={};
{int64_t c = cast<int64_t>(0ULL);for (;(c < cast<int64_t>(2ULL));c++){
int16_t delta = n3ds_clampS16(cast<int32_t>((cast<int32_t>(x1[c]) - cast<int32_t>(x0[c]))));
out[c] = cast<int16_t>(cast<int64_t>((cast<int64_t>(x0[c]) + divi<int64_t>(cast<int64_t>((cast<int64_t>(frac) * cast<int64_t>(delta))),cast<int64_t>(16777216ULL)))));
}
}return out;
}
}
// tools/platform/n3ds/dsp_voice.go:286:1
bool n3ds_Machine_dspDequeue(n3ds_Machine* m,n3ds_dspSource* s){
{
if ((len(s->Queue) == cast<int64_t>(0ULL))) {
return false;
}
int64_t best = cast<int64_t>(0ULL);
{auto&& tmp44 = s->Queue;
for(int64_t tmp45=0;tmp45<len(tmp44);++tmp45){
auto i=tmp45;if ((s->Queue[i].BufferID < s->Queue[best].BufferID)) {
best = i;
}
}}
n3ds_dspBuffer buf = s->Queue[best];
s->Queue = append(sub(s->Queue,0,best),sub(s->Queue,cast<int64_t>((best + cast<int64_t>(1ULL))),len(s->Queue)));
if (m->DSPTrace) {
go_fmt_Printf(std::string("    dsp DEQUEUE src? id=%d addr=%08X len=%d loop=%v fromQueue=%v played=%v (queue left %d, frame %d)\012",101),buf.BufferID,buf.PhysAddr,buf.Length,buf.IsLooping,buf.FromQueue,buf.HasPlayed,len(s->Queue),m->dsp.Ticks);
}
if (buf.AdpcmDirty) {
auto tmp46 = std::make_tuple(buf.AdpcmYn[cast<int64_t>(0ULL)],buf.AdpcmYn[cast<int64_t>(1ULL)]);
s->AdpcmYn1 = std::get<0>(tmp46);
s->AdpcmYn2 = std::get<1>(tmp46);
}
s->CurBuf = n3ds_Machine_dspDecodeBuffer(m,s,buf);
s->CurSample = cast<uint32_t>(0ULL);
if ((!buf.HasPlayed)) {
s->CurSample = buf.PlayPosition;
}
s->CurPhysAddr = buf.PhysAddr;
s->CurBufferID = buf.BufferID;
s->LastBufferID = cast<uint16_t>(0ULL);
s->BufferUpdate = (buf.FromQueue && (!buf.HasPlayed));
{
int64_t n = cast<int64_t>(s->CurSample);
if ((n < len(s->CurBuf))) {
s->CurBuf = sub(s->CurBuf,n,len(s->CurBuf));
}
else {
s->CurBuf = {};
}
}
if (buf.IsLooping) {
buf.HasPlayed = true;
s->Queue = append(s->Queue,buf);
}
return true;
}
}
// tools/platform/n3ds/dsp_voice.go:341:1
Slice<n3ds_dspSample> n3ds_Machine_dspDecodeBuffer(n3ds_Machine* m,n3ds_dspSource* s,n3ds_dspBuffer buf){
{
if ((cast<int32_t>(buf.Length) <= cast<int32_t>(0ULL))) {
return {};
}
uint32_t addr = n3ds_Machine_gpuAddrToVirt(m,(buf.PhysAddr & ~(cast<uint32_t>(3ULL))));
uint32_t chans = cast<uint32_t>(1ULL);
if (buf.Stereo) {
chans = cast<uint32_t>(2ULL);
}
{
switch(buf.Format){
case cast<uint8_t>(0ULL):{
Slice<uint8_t> data = n3ds_Machine_ReadBytes(m,addr,cast<uint32_t>((buf.Length * chans)));
Slice<n3ds_dspSample> out = Slice<n3ds_dspSample>::make(buf.Length);
{auto&& tmp47 = out;
for(int64_t tmp48=0;tmp48<len(tmp47);++tmp48){
auto i=tmp48;if (buf.Stereo) {
out[i] = n3ds_dspSample{n3ds_pcm8(data[cast<int64_t>((i * cast<int64_t>(2ULL)))]),n3ds_pcm8(data[cast<int64_t>((cast<int64_t>((i * cast<int64_t>(2ULL))) + cast<int64_t>(1ULL)))])};
}
else {
int16_t v = n3ds_pcm8(data[i]);
out[i] = n3ds_dspSample{v,v};
}
}}
return out;
break;}
case cast<uint8_t>(1ULL):{
Slice<uint8_t> data = n3ds_Machine_ReadBytes(m,addr,cast<uint32_t>((cast<uint32_t>((buf.Length * chans)) * cast<uint32_t>(2ULL))));
Slice<n3ds_dspSample> out = Slice<n3ds_dspSample>::make(buf.Length);
{auto&& tmp49 = out;
for(int64_t tmp50=0;tmp50<len(tmp49);++tmp50){
auto i=tmp50;if (buf.Stereo) {
out[i] = n3ds_dspSample{n3ds_pcm16(sub(data,cast<int64_t>((i * cast<int64_t>(4ULL))),len(data))),n3ds_pcm16(sub(data,cast<int64_t>((cast<int64_t>((i * cast<int64_t>(4ULL))) + cast<int64_t>(2ULL))),len(data)))};
}
else {
int16_t v = n3ds_pcm16(sub(data,cast<int64_t>((i * cast<int64_t>(2ULL))),len(data)));
out[i] = n3ds_dspSample{v,v};
}
}}
return out;
break;}
case cast<uint8_t>(2ULL):{
uint32_t frames = divi<uint32_t>((cast<uint32_t>((buf.Length + cast<uint32_t>(13ULL)))),cast<uint32_t>(14ULL));
return n3ds_Machine_decodeADPCM(m,s,buf,n3ds_Machine_ReadBytes(m,addr,cast<uint32_t>((frames * cast<uint32_t>(8ULL)))));
break;}
}}
return {};
}
}
// tools/platform/n3ds/dsp_voice.go:382:1
int16_t n3ds_pcm8(uint8_t b){
{
return cast<int16_t>(shl<uint16_t>(cast<uint16_t>(b),cast<int64_t>(8ULL)));
}
}
// tools/platform/n3ds/dsp_voice.go:384:1
int16_t n3ds_pcm16(Slice<uint8_t> b){
{
return cast<int16_t>(cast<uint16_t>((cast<uint16_t>(b[cast<int64_t>(0ULL)]) | shl<uint16_t>(cast<uint16_t>(b[cast<int64_t>(1ULL)]),cast<int64_t>(8ULL)))));
}
}
// tools/platform/n3ds/dsp_voice.go:391:1
Slice<n3ds_dspSample> n3ds_Machine_decodeADPCM(n3ds_Machine* m,n3ds_dspSource* s,n3ds_dspBuffer buf,Slice<uint8_t> data){
{
if (buf.AdpcmDirty) {
(void)(buf.AdpcmPS);
}
auto tmp51 = std::make_tuple(cast<int32_t>(s->AdpcmYn1),cast<int32_t>(s->AdpcmYn2));
int32_t yn1 = std::get<0>(tmp51);
int32_t yn2 = std::get<1>(tmp51);
Slice<n3ds_dspSample> out = Slice<n3ds_dspSample>::make(cast<int64_t>(0ULL),buf.Length);
{int64_t blk = cast<int64_t>(0ULL);for (;((cast<int64_t>((blk * cast<int64_t>(8ULL))) < len(data)) && (cast<uint32_t>(len(out)) < buf.Length));blk++){
uint8_t hdr = data[cast<int64_t>((blk * cast<int64_t>(8ULL)))];
int32_t scale = shl<int32_t>(cast<int32_t>(1ULL),(cast<uint8_t>((hdr & cast<uint8_t>(15ULL)))));
int64_t idx = cast<int64_t>(cast<uint8_t>(((shr<uint8_t>(hdr,cast<int64_t>(4ULL))) & cast<uint8_t>(7ULL))));
auto tmp52 = std::make_tuple(cast<int32_t>(s->AdpcmCoeffs[cast<int64_t>((idx * cast<int64_t>(2ULL)))]),cast<int32_t>(s->AdpcmCoeffs[cast<int64_t>((cast<int64_t>((idx * cast<int64_t>(2ULL))) + cast<int64_t>(1ULL)))]));
int32_t c1 = std::get<0>(tmp52);
int32_t c2 = std::get<1>(tmp52);
auto decode = [&](int32_t nibble)->int16_t{
if ((nibble >= cast<int32_t>(8ULL))) {
nibble -= cast<int32_t>(16ULL);
}
int32_t v = shr<int32_t>((cast<int32_t>((cast<int32_t>((cast<int32_t>((shl<int32_t>((cast<int32_t>((nibble * scale))),cast<int64_t>(11ULL)) + cast<int32_t>(1024ULL))) + cast<int32_t>((c1 * yn1)))) + cast<int32_t>((c2 * yn2))))),cast<int64_t>(11ULL));
int16_t r = n3ds_clampS16(v);
auto tmp53 = std::make_tuple(yn1,cast<int32_t>(r));
yn2 = std::get<0>(tmp53);
yn1 = std::get<1>(tmp53);
return r;
}
;
{int64_t i = cast<int64_t>(1ULL);for (;((i < cast<int64_t>(8ULL)) && (cast<uint32_t>(len(out)) < buf.Length));i++){
uint8_t b = data[cast<int64_t>((cast<int64_t>((blk * cast<int64_t>(8ULL))) + i))];
int16_t hi = decode(cast<int32_t>(shr<uint8_t>(b,cast<int64_t>(4ULL))));
out = append(out,n3ds_dspSample{hi,hi});
if ((cast<uint32_t>(len(out)) >= buf.Length)) {
break;
}
int16_t lo = decode(cast<int32_t>(cast<uint8_t>((b & cast<uint8_t>(15ULL)))));
out = append(out,n3ds_dspSample{lo,lo});
}
}}
}auto tmp54 = std::make_tuple(cast<int16_t>(yn1),cast<int16_t>(yn2));
s->AdpcmYn1 = std::get<0>(tmp54);
s->AdpcmYn2 = std::get<1>(tmp54);
return out;
}
}
// tools/platform/n3ds/dsp_voice.go:435:1
void n3ds_dspSource_mixInto(n3ds_dspSource* s,n3ds_dspQuadFrame* dest,int64_t mix){
{
if ((!s->Enabled)) {
return ;
}
std::array<float,4>* g = (&s->Gain[mix]);
{auto&& tmp55 = (*dest);
for(int64_t tmp56=0;tmp56<len(tmp55);++tmp56){
auto i=tmp56;auto tmp57 = std::make_tuple(cast<float>(s->Frame[i][cast<int64_t>(0ULL)]),cast<float>(s->Frame[i][cast<int64_t>(1ULL)]));
float l = std::get<0>(tmp57);
float r = std::get<1>(tmp57);
(*dest)[i][cast<int64_t>(0ULL)] += cast<int32_t>(cast<float>(((*g)[cast<int64_t>(0ULL)] * l)));
(*dest)[i][cast<int64_t>(1ULL)] += cast<int32_t>(cast<float>(((*g)[cast<int64_t>(1ULL)] * r)));
(*dest)[i][cast<int64_t>(2ULL)] += cast<int32_t>(cast<float>(((*g)[cast<int64_t>(2ULL)] * l)));
(*dest)[i][cast<int64_t>(3ULL)] += cast<int32_t>(cast<float>(((*g)[cast<int64_t>(3ULL)] * r)));
}}
}
}
// tools/platform/n3ds/dsp_voice.go:481:1
void n3ds_Machine_dspMixerConfig(n3ds_Machine* m,uint32_t region){
{
uint32_t cfg = cast<uint32_t>((region + cast<uint32_t>(10336ULL)));
uint32_t dirty = n3ds_Machine_ReadWord(m,cast<uint32_t>((cfg + cast<uint32_t>(0ULL))));
if ((dirty == cast<uint32_t>(0ULL))) {
return ;
}
n3ds_dspHLE* d = (&m->dsp);
if ((cast<uint32_t>((dirty & cast<uint32_t>(256ULL))) != cast<uint32_t>(0ULL))) {
d->AuxBusEnable[cast<int64_t>(0ULL)] = (n3ds_Machine_dspRead16(m,cast<uint32_t>((cfg + cast<uint32_t>(40ULL)))) != cast<uint16_t>(0ULL));
}
if ((cast<uint32_t>((dirty & cast<uint32_t>(512ULL))) != cast<uint32_t>(0ULL))) {
d->AuxBusEnable[cast<int64_t>(1ULL)] = (n3ds_Machine_dspRead16(m,cast<uint32_t>((cast<uint32_t>((cfg + cast<uint32_t>(40ULL))) + cast<uint32_t>(2ULL)))) != cast<uint16_t>(0ULL));
}
if ((cast<uint32_t>((dirty & cast<uint32_t>(65536ULL))) != cast<uint32_t>(0ULL))) {
d->MixVolume[cast<int64_t>(0ULL)] = n3ds_dspFloat(n3ds_Machine_ReadWord(m,cast<uint32_t>((cfg + cast<uint32_t>(4ULL)))));
}
if ((cast<uint32_t>((dirty & cast<uint32_t>(16777216ULL))) != cast<uint32_t>(0ULL))) {
d->MixVolume[cast<int64_t>(1ULL)] = n3ds_dspFloat(n3ds_Machine_ReadWord(m,cast<uint32_t>((cfg + cast<uint32_t>(8ULL)))));
}
if ((cast<uint32_t>((dirty & cast<uint32_t>(33554432ULL))) != cast<uint32_t>(0ULL))) {
d->MixVolume[cast<int64_t>(2ULL)] = n3ds_dspFloat(n3ds_Machine_ReadWord(m,cast<uint32_t>((cast<uint32_t>((cfg + cast<uint32_t>(8ULL))) + cast<uint32_t>(4ULL)))));
}
if ((cast<uint32_t>((dirty & cast<uint32_t>(67108864ULL))) != cast<uint32_t>(0ULL))) {
d->OutputFormat = n3ds_Machine_dspRead16(m,cast<uint32_t>((cfg + cast<uint32_t>(22ULL))));
}
d->Headphones = (n3ds_Machine_dspRead16(m,cast<uint32_t>((cfg + cast<uint32_t>(26ULL)))) != cast<uint16_t>(0ULL));
d->ClippingMode = n3ds_Machine_dspRead16(m,cast<uint32_t>((cfg + cast<uint32_t>(24ULL))));
n3ds_Machine_WriteWord(m,cast<uint32_t>((cfg + cast<uint32_t>(0ULL))),cast<uint32_t>(0ULL));
n3ds_Machine_WriteWord(m,cast<uint32_t>((cfg + cast<uint32_t>(192ULL))),cast<uint32_t>(0ULL));
}
}
// tools/platform/n3ds/dsp_voice.go:514:1
float n3ds_dspFloat(uint32_t w){
{
float f = go_math_Float32frombits(w);
if ((go_math_IsNaN(cast<double>(f)) || go_math_IsInf(cast<double>(f),cast<int64_t>(0ULL)))) {
return cast<float>(0.00000000000000000e+00);
}
return f;
}
}
// tools/platform/n3ds/dsp_voice.go:528:1
n3ds_dspFrame n3ds_Machine_dspMix(n3ds_Machine* m,uint32_t read,uint32_t write,std::array<n3ds_dspQuadFrame,3>* mixes){
{
n3ds_dspHLE* d = (&m->dsp);
std::array<n3ds_dspQuadFrame,3> bus={};
{int64_t b = cast<int64_t>(0ULL);for (;(b < cast<int64_t>(2ULL));b++){
if (d->AuxBusEnable[b]) {
n3ds_Machine_dspReadQuad(m,cast<uint32_t>((cast<uint32_t>((read + cast<uint32_t>(10532ULL))) + cast<uint32_t>((cast<uint32_t>(b) * cast<uint32_t>(2560ULL))))),(&bus[cast<int64_t>((b + cast<int64_t>(1ULL)))]));
}
}
}bus[cast<int64_t>(0ULL)] = (*mixes)[cast<int64_t>(0ULL)];
{int64_t b = cast<int64_t>(0ULL);for (;(b < cast<int64_t>(2ULL));b++){
if (d->AuxBusEnable[b]) {
n3ds_Machine_dspWriteQuad(m,cast<uint32_t>((cast<uint32_t>((write + cast<uint32_t>(10532ULL))) + cast<uint32_t>((cast<uint32_t>(b) * cast<uint32_t>(2560ULL))))),(&(*mixes)[cast<int64_t>((b + cast<int64_t>(1ULL)))]));
}
else {
bus[cast<int64_t>((b + cast<int64_t>(1ULL)))] = (*mixes)[cast<int64_t>((b + cast<int64_t>(1ULL)))];
}
}
}n3ds_dspFrame out={};
{int64_t mix = cast<int64_t>(0ULL);for (;(mix < cast<int64_t>(3ULL));mix++){
float g = d->MixVolume[mix];
{auto&& tmp58 = out;
for(int64_t tmp59=0;tmp59<len(tmp58);++tmp59){
auto i=tmp59;std::array<int32_t,4>* s = (&bus[mix][i]);
int16_t l={};
int16_t r={};
if ((d->OutputFormat == cast<uint16_t>(0ULL))) {
int16_t mono = n3ds_clampS16(cast<int32_t>(cast<float>(((cast<float>((cast<float>((cast<float>((cast<float>((g * cast<float>((*s)[cast<int64_t>(0ULL)]))) + cast<float>((g * cast<float>((*s)[cast<int64_t>(1ULL)]))))) + cast<float>((g * cast<float>((*s)[cast<int64_t>(2ULL)]))))) + cast<float>((g * cast<float>((*s)[cast<int64_t>(3ULL)])))))) / cast<float>(2.00000000000000000e+00)))));
auto tmp60 = std::make_tuple(mono,mono);
l = std::get<0>(tmp60);
r = std::get<1>(tmp60);
}
else {
l = n3ds_clampS16(cast<int32_t>(cast<float>((cast<float>((g * cast<float>((*s)[cast<int64_t>(0ULL)]))) + cast<float>((g * cast<float>((*s)[cast<int64_t>(2ULL)])))))));
r = n3ds_clampS16(cast<int32_t>(cast<float>((cast<float>((g * cast<float>((*s)[cast<int64_t>(1ULL)]))) + cast<float>((g * cast<float>((*s)[cast<int64_t>(3ULL)])))))));
}
out[i][cast<int64_t>(0ULL)] = n3ds_addClampS16(out[i][cast<int64_t>(0ULL)],l);
out[i][cast<int64_t>(1ULL)] = n3ds_addClampS16(out[i][cast<int64_t>(1ULL)],r);
}}
}
}return out;
}
}
// tools/platform/n3ds/dsp_voice.go:574:1
void n3ds_Machine_dspReadQuad(n3ds_Machine* m,uint32_t base,n3ds_dspQuadFrame* q){
{
{int64_t c = cast<int64_t>(0ULL);for (;(c < cast<int64_t>(4ULL));c++){
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(160ULL));i++){
(*q)[i][c] = cast<int32_t>(n3ds_Machine_ReadWord(m,cast<uint32_t>((base + cast<uint32_t>((cast<uint32_t>(cast<int64_t>((cast<int64_t>((c * cast<int64_t>(160ULL))) + i))) * cast<uint32_t>(4ULL)))))));
}
}}
}}
}
// tools/platform/n3ds/dsp_voice.go:582:1
void n3ds_Machine_dspWriteQuad(n3ds_Machine* m,uint32_t base,n3ds_dspQuadFrame* q){
{
{int64_t c = cast<int64_t>(0ULL);for (;(c < cast<int64_t>(4ULL));c++){
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(160ULL));i++){
n3ds_Machine_WriteWord(m,cast<uint32_t>((base + cast<uint32_t>((cast<uint32_t>(cast<int64_t>((cast<int64_t>((c * cast<int64_t>(160ULL))) + i))) * cast<uint32_t>(4ULL))))),cast<uint32_t>((*q)[i][c]));
}
}}
}}
}
// tools/platform/n3ds/dsp_voice.go:592:1
void n3ds_Machine_dspWriteFinal(n3ds_Machine* m,uint32_t write,n3ds_dspFrame* f){
{
{auto&& tmp61 = (*f);
for(int64_t tmp62=0;tmp62<len(tmp61);++tmp62){
auto i=tmp62;n3ds_Machine_dspWrite16(m,cast<uint32_t>((cast<uint32_t>((write + cast<uint32_t>(2688ULL))) + cast<uint32_t>((cast<uint32_t>(i) * cast<uint32_t>(4ULL))))),cast<uint16_t>((*f)[i][cast<int64_t>(0ULL)]));
n3ds_Machine_dspWrite16(m,cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((write + cast<uint32_t>(2688ULL))) + cast<uint32_t>((cast<uint32_t>(i) * cast<uint32_t>(4ULL))))) + cast<uint32_t>(2ULL))),cast<uint16_t>((*f)[i][cast<int64_t>(1ULL)]));
}}
if (m->AudioCapture) {
{auto&& tmp63 = (*f);
for(int64_t tmp64=0;tmp64<len(tmp63);++tmp64){
auto i=tmp64;m->AudioPCM = append(m->AudioPCM,(*f)[i][cast<int64_t>(0ULL)],(*f)[i][cast<int64_t>(1ULL)]);
}}
}
}
}
// tools/platform/n3ds/exefs.go:31:1
std::tuple<n3ds_ExeFS*,Error> n3ds_ParseExeFS(Slice<uint8_t> b){
{
if ((len(b) < cast<int64_t>(512ULL))) {
return {{},go_fmt_Errorf(std::string("n3ds: ExeFS too short (%d bytes)",32),len(b))};
}
n3ds_ExeFS* fs = arenaNew(n3ds_ExeFS{b,{}});
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(10ULL));i++){
Slice<uint8_t> e = sub(b,cast<int64_t>((i * cast<int64_t>(16ULL))),len(b));
std::string name = cast<std::string>(n3ds_trimNul(sub(e,cast<int64_t>(0ULL),cast<int64_t>(8ULL))));
if ((name == std::string("",0))) {
continue;
}
n3ds_ExeFSFile f = n3ds_ExeFSFile{name,cast<int64_t>(le_Uint32(sub(e,cast<int64_t>(8ULL),len(e)))),cast<int64_t>(le_Uint32(sub(e,cast<int64_t>(12ULL),len(e)))),{}};
int64_t end = cast<int64_t>((cast<int64_t>((cast<int64_t>(512ULL) + f.Offset)) + f.Size));
if ((((f.Offset < cast<int64_t>(0ULL)) || (f.Size < cast<int64_t>(0ULL))) || (end > cast<int64_t>(len(b))))) {
return {{},go_fmt_Errorf(std::string("n3ds: ExeFS file %q [0x%x+0x%x] runs past the region end (0x%x)",63),name,f.Offset,f.Size,len(b))};
}
gcopy(sub(f.Hash,0,len(f.Hash)),sub(b,cast<int64_t>((cast<int64_t>(512ULL) - cast<int64_t>((cast<int64_t>(32ULL) * (cast<int64_t>((i + cast<int64_t>(1ULL)))))))),len(b)));
fs->Files = append(fs->Files,f);
}
}if ((len(fs->Files) == cast<int64_t>(0ULL))) {
return {{},go_fmt_Errorf(std::string("n3ds: ExeFS has no files",24))};
}
return {fs,{}};
}
}
// tools/platform/n3ds/exefs.go:62:1
std::tuple<Slice<uint8_t>,Error> n3ds_ExeFS_File(n3ds_ExeFS* fs,std::string name){
{
{auto&& tmp65 = fs->Files;
for(int64_t tmp66=0;tmp66<len(tmp65);++tmp66){
auto f=tmp65[tmp66];if ((f.Name == name)) {
int64_t start = cast<int64_t>((cast<int64_t>(512ULL) + f.Offset));
return {sub(fs->raw,start,cast<int64_t>((start + f.Size))),{}};
}
}}
return {{},go_fmt_Errorf(std::string("n3ds: ExeFS has no file %q",26),name)};
}
}
// tools/platform/n3ds/exefs.go:76:1
std::tuple<Slice<uint8_t>,Error> n3ds_ExeFS_Code(n3ds_ExeFS* fs,n3ds_ExHeader* ex){
{
auto tmp67 = n3ds_ExeFS_File(fs,std::string(".code",5));
Slice<uint8_t> raw = std::get<0>(tmp67);
Error err = std::get<1>(tmp67);
if (bool(err)) {
return {{},err};
}
if ((!ex)) {
return {raw,{}};
}
Slice<uint8_t> code = raw;
if (n3ds_ExHeader_CompressedExeFSCode(ex)) {
{
auto tmp68 = n3ds_DecompressBLZ(raw);
code = std::get<0>(tmp68);
err = std::get<1>(tmp68);
if (bool(err)) {
return {{},go_fmt_Errorf(std::string("n3ds: decompressing .code: %w",29),err)};
}
}
}
{
uint32_t want = n3ds_ExHeader_CodeSize(ex);
if ((cast<uint32_t>(len(code)) != want)) {
return {{},go_fmt_Errorf((std::string("n3ds: .code is 0x%x bytes but the ExHeader describes 0x%x ",58) + std::string("(text extent 0x%x + rodata extent 0x%x + data 0x%x)",51)),len(code),want,n3ds_CodeSegInfo_Extent(ex->Text),n3ds_CodeSegInfo_Extent(ex->ROData),ex->Data.Size)};
}
}
return {code,{}};
}
}
// tools/platform/n3ds/exheader.go:23:1
uint32_t n3ds_CodeSegInfo_Extent(n3ds_CodeSegInfo s){
{
return cast<uint32_t>((s.NumPages * cast<uint32_t>(4096ULL)));
}
}
// tools/platform/n3ds/exheader.go:44:1
bool n3ds_ExHeader_CompressedExeFSCode(n3ds_ExHeader* e){
{
return (cast<uint8_t>((e->Flag & cast<uint8_t>(1ULL))) != cast<uint8_t>(0ULL));
}
}
// tools/platform/n3ds/exheader.go:47:1
bool n3ds_ExHeader_SDApplication(n3ds_ExHeader* e){
{
return (cast<uint8_t>((e->Flag & cast<uint8_t>(2ULL))) != cast<uint8_t>(0ULL));
}
}
// tools/platform/n3ds/exheader.go:50:1
std::tuple<n3ds_ExHeader*,Error> n3ds_ParseExHeader(Slice<uint8_t> b){
{
if ((len(b) < cast<int64_t>(512ULL))) {
return {{},go_fmt_Errorf(std::string("n3ds: ExHeader too short (%d bytes, want >= 0x200)",50),len(b))};
}
auto u32 = [&](int64_t off)->uint32_t{
return le_Uint32(sub(b,off,len(b)));
}
;
auto seg = [&](int64_t off)->n3ds_CodeSegInfo{
return n3ds_CodeSegInfo{u32(off),u32(cast<int64_t>((off + cast<int64_t>(4ULL)))),u32(cast<int64_t>((off + cast<int64_t>(8ULL))))};
}
;
n3ds_ExHeader* e = arenaNew(n3ds_ExHeader{cast<std::string>(n3ds_trimNul(sub(b,cast<int64_t>(0ULL),cast<int64_t>(8ULL)))),b[cast<int64_t>(13ULL)],le_Uint16(sub(b,cast<int64_t>(14ULL),len(b))),seg(cast<int64_t>(16ULL)),u32(cast<int64_t>(28ULL)),seg(cast<int64_t>(32ULL)),seg(cast<int64_t>(48ULL)),u32(cast<int64_t>(60ULL)),{},{},{},{}});
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(48ULL));i++){
{
uint64_t id = le_Uint64(sub(b,cast<int64_t>((cast<int64_t>(64ULL) + cast<int64_t>((i * cast<int64_t>(8ULL))))),len(b)));
if ((id != cast<uint64_t>(0ULL))) {
e->Dependencies = append(e->Dependencies,id);
}
}
}
}e->SaveDataSize = le_Uint64(sub(b,cast<int64_t>(448ULL),len(b)));
e->JumpID = le_Uint64(sub(b,cast<int64_t>(456ULL),len(b)));
if ((len(b) >= cast<int64_t>(1024ULL))) {
e->ACI = sub(b,cast<int64_t>(512ULL),cast<int64_t>(1024ULL));
}
{
Error err = n3ds_ExHeader_validate(e);
if (bool(err)) {
return {{},err};
}
}
return {e,{}};
}
}
// tools/platform/n3ds/exheader.go:91:1
Error n3ds_ExHeader_validate(n3ds_ExHeader* e){
{
{auto&& tmp69 = Slice<Anon15>{Anon15{std::string("text",4),e->Text},Anon15{std::string("rodata",6),e->ROData},Anon15{std::string("data",4),e->Data}};
for(int64_t tmp70=0;tmp70<len(tmp69);++tmp70){
auto s=tmp69[tmp70];if ((s.seg.Size > n3ds_CodeSegInfo_Extent(s.seg))) {
return go_fmt_Errorf(std::string("n3ds: ExHeader %s size 0x%x exceeds its %d-page extent 0x%x",59),s.name,s.seg.Size,s.seg.NumPages,n3ds_CodeSegInfo_Extent(s.seg));
}
}}
{
auto tmp71 = std::make_tuple(e->ROData.Address,cast<uint32_t>((e->Text.Address + n3ds_CodeSegInfo_Extent(e->Text))));
uint32_t got = std::get<0>(tmp71);
uint32_t want = std::get<1>(tmp71);
if ((got != want)) {
return go_fmt_Errorf(std::string("n3ds: ExHeader rodata address 0x%08x does not follow text extent (expected 0x%08x)",82),got,want);
}
}
{
auto tmp72 = std::make_tuple(e->Data.Address,cast<uint32_t>((e->ROData.Address + n3ds_CodeSegInfo_Extent(e->ROData))));
uint32_t got = std::get<0>(tmp72);
uint32_t want = std::get<1>(tmp72);
if ((got != want)) {
return go_fmt_Errorf(std::string("n3ds: ExHeader data address 0x%08x does not follow rodata extent (expected 0x%08x)",82),got,want);
}
}
return {};
}
}
// tools/platform/n3ds/exheader.go:117:1
uint32_t n3ds_ExHeader_CodeSize(n3ds_ExHeader* e){
{
return cast<uint32_t>((cast<uint32_t>((n3ds_CodeSegInfo_Extent(e->Text) + n3ds_CodeSegInfo_Extent(e->ROData))) + n3ds_CodeSegInfo_Extent(e->Data)));
}
}
// tools/platform/n3ds/exheader.go:123:1
uint32_t n3ds_ExHeader_BSSAddress(n3ds_ExHeader* e){
{
return cast<uint32_t>((e->Data.Address + e->Data.Size));
}
}
// tools/platform/n3ds/fs.go:26:1
bool n3ds_Machine_isFileSession(n3ds_Machine* m,uint32_t handle){
{
auto tmp73 = lookup(m->fsFiles,handle);
bool ok = std::get<1>(tmp73);
return ok;
}
}
// tools/platform/n3ds/fs.go:46:1
bool n3ds_Machine_isDirSession(n3ds_Machine* m,uint32_t handle){
{
auto tmp74 = lookup(m->fsDirs,handle);
bool ok = std::get<1>(tmp74);
return ok;
}
}
// tools/platform/n3ds/fs.go:55:1
bool n3ds_Machine_fsOpenDirectory(n3ds_Machine* m,n3ds_ipcHeader hdr){
{
auto tmp75 = std::make_tuple(n3ds_Machine_ipcArg(m,cast<int64_t>(3ULL)),n3ds_Machine_ipcArg(m,cast<int64_t>(4ULL)));
uint32_t pathType = std::get<0>(tmp75);
uint32_t pathSize = std::get<1>(tmp75);
uint32_t pathPtr = n3ds_Machine_ipcArg(m,cast<int64_t>(6ULL));
std::string path = n3ds_Machine_readFSPath(m,pathPtr,pathType,pathSize);
if (m->Verbose) {
go_fmt_Printf(std::string("    fsOpenDirectory pathType=%d path=%q\012",40),pathType,path);
}
auto tmp76 = n3ds_Machine_romfsChildren(m,path);
n3ds_fsDir* dir = std::get<0>(tmp76);
bool ok = std::get<1>(tmp76);
if ((!ok)) {
uint32_t code = cast<uint32_t>(3363849336ULL);
{
auto tmp77 = lookup(m->fsArchives,n3ds_Machine_ipcArg(m,cast<int64_t>(1ULL)));
uint32_t id = std::get<0>(tmp77);
bool open = std::get<1>(tmp77);
if (((open && (id != cast<uint32_t>(3ULL))) && (!m->saveFormatted))) {
code = cast<uint32_t>(3363849556ULL);
}
}
n3ds_Machine_WriteWord(m,n3ds_Machine_cmdBuf(m),cast<uint32_t>((shl<uint32_t>(cast<uint32_t>(hdr.Command),cast<int64_t>(16ULL)) | cast<uint32_t>(64ULL))));
n3ds_Machine_WriteWord(m,cast<uint32_t>((n3ds_Machine_cmdBuf(m) + cast<uint32_t>(4ULL))),code);
return true;
}
uint32_t h = n3ds_Machine_newHandle(m,std::string("fs-dir",6),false);
m->fsDirs[h] = dir;
n3ds_Machine_WriteWord(m,n3ds_Machine_cmdBuf(m),cast<uint32_t>((cast<uint32_t>((shl<uint32_t>(cast<uint32_t>(hdr.Command),cast<int64_t>(16ULL)) | cast<uint32_t>(64ULL))) | cast<uint32_t>(2ULL))));
n3ds_Machine_WriteWord(m,cast<uint32_t>((n3ds_Machine_cmdBuf(m) + cast<uint32_t>(4ULL))),cast<uint32_t>(0ULL));
n3ds_Machine_WriteWord(m,cast<uint32_t>((n3ds_Machine_cmdBuf(m) + cast<uint32_t>(8ULL))),cast<uint32_t>(0ULL));
n3ds_Machine_WriteWord(m,cast<uint32_t>((n3ds_Machine_cmdBuf(m) + cast<uint32_t>(12ULL))),h);
return true;
}
}
// tools/platform/n3ds/fs.go:86:1
std::tuple<n3ds_fsDir*,bool> n3ds_Machine_romfsChildren(n3ds_Machine* m,std::string path){
{
if ((!m->romfs)) {
return {{},false};
}
if ((path == std::string("",0))) {
return {{},false};
}
if ((path[cast<int64_t>((len(path) - cast<int64_t>(1ULL)))] != cast<uint8_t>(47ULL))) {
path += std::string("/",1);
}
if ((path == std::string("/",1))) {
}
else {
bool found = false;
{auto&& tmp78 = m->romfs->Dirs;
for(int64_t tmp79=0;tmp79<len(tmp78);++tmp79){
auto d=tmp78[tmp79];if (((d + std::string("/",1)) == path)) {
found = true;
break;
}
}}
if ((!found)) {
return {{},false};
}
}
n3ds_fsDir* dir = arenaNew(n3ds_fsDir{path,{},{}});
Map<std::string,bool> seen = Map<std::string,bool>{};
{auto&& tmp80 = m->romfs->Dirs;
for(int64_t tmp81=0;tmp81<len(tmp80);++tmp81){
auto d=tmp80[tmp81];if ((((len(d) > len(path)) && (sub(d,0,len(path)) == path)) && (!n3ds_containsSlash(sub(d,len(path),len(d)))))) {
std::string name = sub(d,len(path),len(d));
if ((!get(seen,name))) {
seen[name] = true;
dir->entries = append(dir->entries,n3ds_fsDirEntry{name,true,{}});
}
}
}}
{auto&& tmp82 = m->romfs->Files;
for(int64_t tmp83=0;tmp83<len(tmp82);++tmp83){
auto f=tmp82[tmp83];if ((((len(f.Path) > len(path)) && (sub(f.Path,0,len(path)) == path)) && (!n3ds_containsSlash(sub(f.Path,len(path),len(f.Path)))))) {
dir->entries = append(dir->entries,n3ds_fsDirEntry{sub(f.Path,len(path),len(f.Path)),{},f.Size});
}
}}
return {dir,true};
}
}
// tools/platform/n3ds/fs.go:129:1
bool n3ds_containsSlash(std::string s){
{
{int64_t i = cast<int64_t>(0ULL);for (;(i < len(s));i++){
if ((s[i] == cast<uint8_t>(47ULL))) {
return true;
}
}
}return false;
}
}
// tools/platform/n3ds/fs.go:143:1
bool n3ds_Machine_ipcDir(n3ds_Machine* m,uint32_t handle,n3ds_ipcHeader hdr){
{
n3ds_fsDir* d = get(m->fsDirs,handle);
{
switch(hdr.Command){
case cast<uint16_t>(2049ULL):{
int64_t count = cast<int64_t>(n3ds_Machine_ipcArg(m,cast<int64_t>(1ULL)));
uint32_t out = n3ds_Machine_ipcArg(m,cast<int64_t>(3ULL));
int64_t n = cast<int64_t>(0ULL);
constexpr int64_t entrySize=552ULL;
{;for (;((n < count) && (d->cursor < len(d->entries)));){
n3ds_fsDirEntry e = d->entries[d->cursor];
uint32_t base = cast<uint32_t>((out + cast<uint32_t>(cast<int64_t>((n * cast<int64_t>(552ULL))))));
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < cast<uint32_t>(552ULL));i++){
n3ds_Machine_Write(m,cast<uint32_t>((base + i)),cast<uint8_t>(0ULL));
}
}{auto&& tmp84 = e.name;
for(int64_t tmp85=0;tmp85<len(tmp84);++tmp85){
auto i=tmp85;auto r=tmp84[tmp85];if ((i >= cast<int64_t>(261ULL))) {
break;
}
n3ds_Machine_Write(m,cast<uint32_t>((base + cast<uint32_t>(cast<int64_t>((i * cast<int64_t>(2ULL)))))),cast<uint8_t>(r));
n3ds_Machine_Write(m,cast<uint32_t>((cast<uint32_t>((base + cast<uint32_t>(cast<int64_t>((i * cast<int64_t>(2ULL)))))) + cast<uint32_t>(1ULL))),cast<uint8_t>(shr<uint16_t>(cast<uint16_t>(r),cast<int64_t>(8ULL))));
}}
if (e.isDir) {
n3ds_Machine_Write(m,cast<uint32_t>((base + cast<uint32_t>(540ULL))),cast<uint8_t>(1ULL));
}
n3ds_Machine_WriteWord(m,cast<uint32_t>((base + cast<uint32_t>(544ULL))),cast<uint32_t>(e.size));
n3ds_Machine_WriteWord(m,cast<uint32_t>((base + cast<uint32_t>(548ULL))),cast<uint32_t>(shr<int64_t>(e.size,cast<int64_t>(32ULL))));
d->cursor++;
n++;
}
}if (m->Verbose) {
go_fmt_Printf(std::string("    fsDirRead %q -> %d entries (cursor %d/%d)\012",46),d->path,n,d->cursor,len(d->entries));
}
n3ds_Machine_ipcReply(m,hdr.Command,Slice<uint32_t>{cast<uint32_t>(n)});
return true;
break;}
case cast<uint16_t>(2050ULL):{
removeKey(m->fsDirs,handle);
n3ds_Machine_ipcReply(m,hdr.Command,Slice<uint32_t>{});
return true;
break;}
}}
arm_CPU_Halt(m->CPU,std::string("fs dir command 0x%04X unimplemented at 0x%08X after %d instructions",67),hdr.Command,arm_CPU_PC(m->CPU),m->CPU->Instrs);
return true;
}
}
// tools/platform/n3ds/fs.go:194:1
bool n3ds_Machine_fsOpenFile(n3ds_Machine* m,n3ds_ipcHeader hdr){
{
uint32_t archive={};
uint32_t pathType={};
uint32_t pathSize={};
uint32_t pathPtr={};
bool saveArchive = false;
if ((hdr.Command == cast<uint16_t>(2051ULL))) {
archive = n3ds_Machine_ipcArg(m,cast<int64_t>(2ULL));
auto tmp86 = std::make_tuple(n3ds_Machine_ipcArg(m,cast<int64_t>(6ULL)),n3ds_Machine_ipcArg(m,cast<int64_t>(7ULL)));
pathType = std::get<0>(tmp86);
pathSize = std::get<1>(tmp86);
pathPtr = n3ds_Machine_ipcArg(m,cast<int64_t>(13ULL));
}
else {
uint32_t archiveHandle = n3ds_Machine_ipcArg(m,cast<int64_t>(2ULL));
{
auto tmp87 = lookup(m->fsArchives,archiveHandle);
uint32_t id = std::get<0>(tmp87);
bool ok = std::get<1>(tmp87);
if ((ok && (id != cast<uint32_t>(3ULL)))) {
saveArchive = true;
}
}
auto tmp88 = std::make_tuple(n3ds_Machine_ipcArg(m,cast<int64_t>(4ULL)),n3ds_Machine_ipcArg(m,cast<int64_t>(5ULL)));
pathType = std::get<0>(tmp88);
pathSize = std::get<1>(tmp88);
pathPtr = n3ds_Machine_ipcArg(m,cast<int64_t>(9ULL));
}
std::string path = n3ds_Machine_readFSPath(m,pathPtr,pathType,pathSize);
if (m->Verbose) {
go_fmt_Printf(std::string("    fsOpen cmd=0x%04X archive=0x%08X save=%v pathType=%d pathSize=%d flags=0x%X attr=0x%X path=%q\012",98),hdr.Command,archive,saveArchive,pathType,pathSize,n3ds_Machine_ipcArg(m,cast<int64_t>(6ULL)),n3ds_Machine_ipcArg(m,cast<int64_t>(7ULL)),path);
}
if (saveArchive) {
auto tmp89 = lookup(m->saveFiles,path);
Slice<uint8_t> data = std::get<0>(tmp89);
bool ok = std::get<1>(tmp89);
if ((((!ok) && (hdr.Command == cast<uint16_t>(2050ULL))) && (cast<uint32_t>((n3ds_Machine_ipcArg(m,cast<int64_t>(6ULL)) & cast<uint32_t>(4ULL))) != cast<uint32_t>(0ULL)))) {
m->saveFiles[path] = Slice<uint8_t>{};
auto tmp90 = std::make_tuple(get(m->saveFiles,path),true);
data = std::get<0>(tmp90);
ok = std::get<1>(tmp90);
}
if ((!ok)) {
uint32_t code = cast<uint32_t>(3363849556ULL);
if (m->saveFormatted) {
code = cast<uint32_t>(3363849336ULL);
}
n3ds_Machine_WriteWord(m,n3ds_Machine_cmdBuf(m),cast<uint32_t>((shl<uint32_t>(cast<uint32_t>(hdr.Command),cast<int64_t>(16ULL)) | cast<uint32_t>(64ULL))));
n3ds_Machine_WriteWord(m,cast<uint32_t>((n3ds_Machine_cmdBuf(m) + cast<uint32_t>(4ULL))),code);
return true;
}
n3ds_fsFile* sess = arenaNew(n3ds_fsFile{data,path,path});
uint32_t h = n3ds_Machine_newHandle(m,std::string("fs-file",7),false);
m->fsFiles[h] = sess;
n3ds_Machine_WriteWord(m,n3ds_Machine_cmdBuf(m),cast<uint32_t>((cast<uint32_t>((shl<uint32_t>(cast<uint32_t>(hdr.Command),cast<int64_t>(16ULL)) | cast<uint32_t>(64ULL))) | cast<uint32_t>(2ULL))));
n3ds_Machine_WriteWord(m,cast<uint32_t>((n3ds_Machine_cmdBuf(m) + cast<uint32_t>(4ULL))),cast<uint32_t>(0ULL));
n3ds_Machine_WriteWord(m,cast<uint32_t>((n3ds_Machine_cmdBuf(m) + cast<uint32_t>(8ULL))),cast<uint32_t>(0ULL));
n3ds_Machine_WriteWord(m,cast<uint32_t>((n3ds_Machine_cmdBuf(m) + cast<uint32_t>(12ULL))),h);
return true;
}
Slice<uint8_t> data={};
bool found = false;
{
if (((archive == cast<uint32_t>(3ULL)) && bool(m->romfsRaw))){
int64_t l3 = cast<int64_t>(0ULL);
if (bool(m->romfs)) {
l3 = m->romfs->Levels[cast<int64_t>(2ULL)].Offset;
}
auto tmp92 = std::make_tuple(sub(m->romfsRaw,l3,len(m->romfsRaw)),true);
data = std::get<0>(tmp92);
found = std::get<1>(tmp92);
path = std::string("<romfs-l3>",10);
}
else if ((bool(m->romfs) && (path != std::string("",0)))){
{
auto tmp93 = n3ds_RomFS_File(m->romfs,path);
Slice<uint8_t> d = std::get<0>(tmp93);
Error err = std::get<1>(tmp93);
if ((!err)) {
auto tmp94 = std::make_tuple(d,true);
data = std::get<0>(tmp94);
found = std::get<1>(tmp94);
}
}
}
}
tmp91:;
if ((!found)) {
if (m->Verbose) {
go_fmt_Printf(std::string("    fsOpenFile %q -> NOT FOUND\012",31),path);
}
n3ds_Machine_WriteWord(m,n3ds_Machine_cmdBuf(m),cast<uint32_t>((shl<uint32_t>(cast<uint32_t>(hdr.Command),cast<int64_t>(16ULL)) | cast<uint32_t>(64ULL))));
n3ds_Machine_WriteWord(m,cast<uint32_t>((n3ds_Machine_cmdBuf(m) + cast<uint32_t>(4ULL))),cast<uint32_t>(3363849336ULL));
return true;
}
n3ds_fsFile* sess = arenaNew(n3ds_fsFile{data,path,{}});
uint32_t h = n3ds_Machine_newHandle(m,std::string("fs-file",7),false);
m->fsFiles[h] = sess;
if (m->Verbose) {
go_fmt_Printf(std::string("    fsOpenFile %q -> handle 0x%08X (%d bytes)\012",46),path,h,len(sess->data));
}
n3ds_Machine_WriteWord(m,n3ds_Machine_cmdBuf(m),cast<uint32_t>((cast<uint32_t>((shl<uint32_t>(cast<uint32_t>(hdr.Command),cast<int64_t>(16ULL)) | cast<uint32_t>(64ULL))) | cast<uint32_t>(2ULL))));
n3ds_Machine_WriteWord(m,cast<uint32_t>((n3ds_Machine_cmdBuf(m) + cast<uint32_t>(4ULL))),cast<uint32_t>(0ULL));
n3ds_Machine_WriteWord(m,cast<uint32_t>((n3ds_Machine_cmdBuf(m) + cast<uint32_t>(8ULL))),cast<uint32_t>(0ULL));
n3ds_Machine_WriteWord(m,cast<uint32_t>((n3ds_Machine_cmdBuf(m) + cast<uint32_t>(12ULL))),h);
return true;
}
}
// tools/platform/n3ds/fs.go:308:1
std::string n3ds_Machine_readFSPath(n3ds_Machine* m,uint32_t ptr,uint32_t ptype,uint32_t size){
{
if ((((ptr == cast<uint32_t>(0ULL)) || (size == cast<uint32_t>(0ULL))) || (size > cast<uint32_t>(4096ULL)))) {
return std::string("",0);
}
{
switch(ptype){
case cast<uint32_t>(3ULL):{
Slice<uint8_t> b={};
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < size);i++){
uint8_t c = n3ds_Machine_Read(m,cast<uint32_t>((ptr + i)));
if ((c == cast<uint8_t>(0ULL))) {
break;
}
b = append(b,c);
}
}return cast<std::string>(b);
break;}
case cast<uint32_t>(4ULL):{
Slice<int32_t> r={};
{uint32_t i = cast<uint32_t>(0ULL);for (;(cast<uint32_t>((i + cast<uint32_t>(1ULL))) < size);i += cast<uint32_t>(2ULL)){
uint16_t u = cast<uint16_t>((cast<uint16_t>(n3ds_Machine_Read(m,cast<uint32_t>((ptr + i)))) | shl<uint16_t>(cast<uint16_t>(n3ds_Machine_Read(m,cast<uint32_t>((cast<uint32_t>((ptr + i)) + cast<uint32_t>(1ULL))))),cast<int64_t>(8ULL))));
if ((u == cast<uint16_t>(0ULL))) {
break;
}
r = append(r,cast<int32_t>(u));
}
}return cast<std::string>(r);
break;}
}}
return std::string("",0);
}
}
// tools/platform/n3ds/fs.go:338:1
bool n3ds_Machine_ipcFile(n3ds_Machine* m,uint32_t handle,n3ds_ipcHeader hdr){
{
n3ds_fsFile* f = get(m->fsFiles,handle);
{
switch(hdr.Command){
case cast<uint16_t>(2050ULL):{
int64_t off = cast<int64_t>(cast<uint64_t>((cast<uint64_t>(n3ds_Machine_ipcArg(m,cast<int64_t>(1ULL))) | shl<uint64_t>(cast<uint64_t>(n3ds_Machine_ipcArg(m,cast<int64_t>(2ULL))),cast<int64_t>(32ULL)))));
int64_t size = cast<int64_t>(n3ds_Machine_ipcArg(m,cast<int64_t>(3ULL)));
uint32_t bufPtr = n3ds_Machine_ipcArg(m,cast<int64_t>(5ULL));
int64_t n = cast<int64_t>(0ULL);
if ((bool(f) && (f->save != std::string("",0)))) {
f->data = get(m->saveFiles,f->save);
}
if (((bool(f) && (off >= cast<int64_t>(0ULL))) && (off < cast<int64_t>(len(f->data))))) {
n = size;
if ((cast<int64_t>((off + n)) > cast<int64_t>(len(f->data)))) {
n = cast<int64_t>((cast<int64_t>(len(f->data)) - off));
}
{int64_t i = cast<int64_t>(0ULL);for (;(i < n);i++){
n3ds_Machine_Write(m,cast<uint32_t>((bufPtr + cast<uint32_t>(i))),f->data[cast<int64_t>((off + i))]);
}
}}
if (m->Verbose) {
uint32_t head={};
if ((bool(f) && (n >= cast<int64_t>(4ULL)))) {
head = n3ds_Machine_ReadWord(m,bufPtr);
}
go_fmt_Printf(std::string("    IFile Read h=0x%08X off=%d size=%d -> %d bytes (flen=%d) head=%08X\012",71),handle,off,size,n,n3ds_fileLen(f),head);
}
n3ds_Machine_ipcReply(m,hdr.Command,Slice<uint32_t>{cast<uint32_t>(n)});
return true;
break;}
case cast<uint16_t>(2052ULL):{
uint64_t sz={};
if (bool(f)) {
if ((f->save != std::string("",0))) {
f->data = get(m->saveFiles,f->save);
}
sz = cast<uint64_t>(len(f->data));
}
n3ds_Machine_ipcReply(m,hdr.Command,Slice<uint32_t>{cast<uint32_t>(sz),cast<uint32_t>(shr<uint64_t>(sz,cast<int64_t>(32ULL)))});
return true;
break;}
case cast<uint16_t>(2056ULL):{
removeKey(m->fsFiles,handle);
removeKey(m->handles,handle);
n3ds_Machine_ipcReply(m,hdr.Command,Slice<uint32_t>{});
return true;
break;}
case cast<uint16_t>(2051ULL):{
int64_t off = cast<int64_t>(cast<uint64_t>((cast<uint64_t>(n3ds_Machine_ipcArg(m,cast<int64_t>(1ULL))) | shl<uint64_t>(cast<uint64_t>(n3ds_Machine_ipcArg(m,cast<int64_t>(2ULL))),cast<int64_t>(32ULL)))));
int64_t size = cast<int64_t>(n3ds_Machine_ipcArg(m,cast<int64_t>(3ULL)));
uint32_t bufPtr = n3ds_Machine_ipcArg(m,cast<int64_t>(6ULL));
if ((((bool(f) && (f->save != std::string("",0))) && (off >= cast<int64_t>(0ULL))) && (size >= cast<int64_t>(0ULL)))) {
Slice<uint8_t> data = get(m->saveFiles,f->save);
{
int64_t need = cast<int64_t>((off + size));
if ((cast<int64_t>(len(data)) < need)) {
Slice<uint8_t> grown = Slice<uint8_t>::make(need);
gcopy(grown,data);
data = grown;
}
}
{int64_t i = cast<int64_t>(0ULL);for (;(i < size);i++){
data[cast<int64_t>((off + i))] = n3ds_Machine_Read(m,cast<uint32_t>((bufPtr + cast<uint32_t>(i))));
}
}m->saveFiles[f->save] = data;
f->data = data;
if (m->Verbose) {
go_fmt_Printf(std::string("    IFile Write %q off=%d size=%d\012",34),f->save,off,size);
}
n3ds_Machine_ipcReply(m,hdr.Command,Slice<uint32_t>{cast<uint32_t>(size)});
return true;
}
arm_CPU_Halt(m->CPU,std::string("IFile Write on non-writable session (path %q) at 0x%08X",55),n3ds_filePath(f),arm_CPU_PC(m->CPU));
return true;
break;}
case cast<uint16_t>(2053ULL):{
if ((bool(f) && (f->save != std::string("",0)))) {
int64_t size = cast<int64_t>(cast<uint64_t>((cast<uint64_t>(n3ds_Machine_ipcArg(m,cast<int64_t>(1ULL))) | shl<uint64_t>(cast<uint64_t>(n3ds_Machine_ipcArg(m,cast<int64_t>(2ULL))),cast<int64_t>(32ULL)))));
Slice<uint8_t> data = get(m->saveFiles,f->save);
if ((cast<int64_t>(len(data)) != size)) {
Slice<uint8_t> resized = Slice<uint8_t>::make(size);
gcopy(resized,data);
m->saveFiles[f->save] = resized;
f->data = resized;
}
if (m->Verbose) {
go_fmt_Printf(std::string("    IFile SetSize %q -> %d\012",27),f->save,size);
}
}
n3ds_Machine_ipcReply(m,hdr.Command,Slice<uint32_t>{});
return true;
break;}
case cast<uint16_t>(2054ULL):case cast<uint16_t>(2055ULL):{
n3ds_Machine_ipcReply(m,hdr.Command,Slice<uint32_t>{});
return true;
break;}
case cast<uint16_t>(2058ULL):{
n3ds_Machine_ipcReply(m,hdr.Command,Slice<uint32_t>{});
return true;
break;}
case cast<uint16_t>(2059ULL):{
n3ds_Machine_ipcReply(m,hdr.Command,Slice<uint32_t>{cast<uint32_t>(0ULL)});
return true;
break;}
case cast<uint16_t>(2060ULL):{
if ((!f)) {
arm_CPU_Halt(m->CPU,std::string("IFile OpenLinkFile on unknown session 0x%08X at 0x%08X",54),handle,arm_CPU_PC(m->CPU));
return true;
}
n3ds_fsFile* link = arenaNew(n3ds_fsFile{f->data,f->path,f->save});
uint32_t h = n3ds_Machine_newHandle(m,std::string("fs-file",7),false);
m->fsFiles[h] = link;
if (m->Verbose) {
go_fmt_Printf(std::string("    IFile OpenLinkFile %q -> handle 0x%08X\012",43),n3ds_filePath(f),h);
}
n3ds_Machine_WriteWord(m,n3ds_Machine_cmdBuf(m),cast<uint32_t>((cast<uint32_t>((shl<uint32_t>(cast<uint32_t>(hdr.Command),cast<int64_t>(16ULL)) | cast<uint32_t>(64ULL))) | cast<uint32_t>(2ULL))));
n3ds_Machine_WriteWord(m,cast<uint32_t>((n3ds_Machine_cmdBuf(m) + cast<uint32_t>(4ULL))),cast<uint32_t>(0ULL));
n3ds_Machine_WriteWord(m,cast<uint32_t>((n3ds_Machine_cmdBuf(m) + cast<uint32_t>(8ULL))),cast<uint32_t>(0ULL));
n3ds_Machine_WriteWord(m,cast<uint32_t>((n3ds_Machine_cmdBuf(m) + cast<uint32_t>(12ULL))),h);
return true;
break;}
}}
arm_CPU_Halt(m->CPU,std::string("IFile command 0x%04X unimplemented at 0x%08X after %d instructions",66),hdr.Command,arm_CPU_PC(m->CPU),m->CPU->Instrs);
return true;
}
}
// tools/platform/n3ds/fs.go:459:1
int64_t n3ds_fileLen(n3ds_fsFile* f){
{
if ((!f)) {
return cast<int64_t>(-1ULL);
}
return len(f->data);
}
}
// tools/platform/n3ds/fs.go:466:1
std::string n3ds_filePath(n3ds_fsFile* f){
{
if ((!f)) {
return std::string("<nil>",5);
}
return f->path;
}
}
// tools/platform/n3ds/fs.go:477:1
bool n3ds_Machine_fsOpenArchive(n3ds_Machine* m,n3ds_ipcHeader hdr){
{
uint32_t id = n3ds_Machine_ipcArg(m,cast<int64_t>(1ULL));
uint32_t h = n3ds_Machine_newHandle(m,std::string("fs-archive",10),false);
m->fsArchives[h] = id;
if (m->Verbose) {
go_fmt_Printf(std::string("    fsOpenArchive id=0x%X -> handle 0x%08X\012",43),id,h);
}
n3ds_Machine_ipcReply(m,hdr.Command,Slice<uint32_t>{h,cast<uint32_t>(0ULL)});
return true;
}
}
// tools/platform/n3ds/fs.go:491:1
bool n3ds_Machine_fsDeleteFile(n3ds_Machine* m,n3ds_ipcHeader hdr){
{
std::string path = n3ds_Machine_readFSPath(m,n3ds_Machine_ipcArg(m,cast<int64_t>(7ULL)),n3ds_Machine_ipcArg(m,cast<int64_t>(4ULL)),n3ds_Machine_ipcArg(m,cast<int64_t>(5ULL)));
if (m->Verbose) {
go_fmt_Printf(std::string("    fsDeleteFile %q\012",20),path);
}
{
auto tmp95 = lookup(m->saveFiles,path);
bool ok = std::get<1>(tmp95);
if (ok) {
removeKey(m->saveFiles,path);
n3ds_Machine_ipcReply(m,hdr.Command,Slice<uint32_t>{});
}
else {
n3ds_Machine_WriteWord(m,n3ds_Machine_cmdBuf(m),cast<uint32_t>((shl<uint32_t>(cast<uint32_t>(hdr.Command),cast<int64_t>(16ULL)) | cast<uint32_t>(64ULL))));
n3ds_Machine_WriteWord(m,cast<uint32_t>((n3ds_Machine_cmdBuf(m) + cast<uint32_t>(4ULL))),cast<uint32_t>(3363849336ULL));
}
}
return true;
}
}
// tools/platform/n3ds/fs.go:511:1
bool n3ds_Machine_fsCreateFile(n3ds_Machine* m,n3ds_ipcHeader hdr){
{
auto tmp96 = std::make_tuple(n3ds_Machine_ipcArg(m,cast<int64_t>(4ULL)),n3ds_Machine_ipcArg(m,cast<int64_t>(5ULL)));
uint32_t pathType = std::get<0>(tmp96);
uint32_t pathSize = std::get<1>(tmp96);
int64_t size = cast<int64_t>(cast<uint64_t>((cast<uint64_t>(n3ds_Machine_ipcArg(m,cast<int64_t>(7ULL))) | shl<uint64_t>(cast<uint64_t>(n3ds_Machine_ipcArg(m,cast<int64_t>(8ULL))),cast<int64_t>(32ULL)))));
std::string path = n3ds_Machine_readFSPath(m,n3ds_Machine_ipcArg(m,cast<int64_t>(10ULL)),pathType,pathSize);
if (m->Verbose) {
go_fmt_Printf(std::string("    fsCreateFile %q size=%d\012",28),path,size);
}
if (((size < cast<int64_t>(0ULL)) || (size > cast<int64_t>(16777216ULL)))) {
arm_CPU_Halt(m->CPU,std::string("fs CreateFile %q with implausible size %d at 0x%08X",51),path,size,arm_CPU_PC(m->CPU));
return true;
}
m->saveFiles[path] = Slice<uint8_t>::make(size);
n3ds_Machine_ipcReply(m,hdr.Command,Slice<uint32_t>{});
return true;
}
}
// tools/platform/n3ds/gpu.go:232:1
n3ds_GPU* n3ds_newGPU(n3ds_Machine* m){
{
return arenaNew(n3ds_GPU{m,{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},cast<uint32_t>(1ULL),{},{},{},{},{},{},{},{}});
}
}
// tools/platform/n3ds/gpu.go:235:1
n3ds_GPU* n3ds_Machine_GPU(n3ds_Machine* m){
{
return m->gpu;
}
}
// tools/platform/n3ds/gpu.go:242:1
void n3ds_GPU_Execute(n3ds_GPU* g,uint32_t addr,uint32_t size){
rrprof::Scope profile(1,"PICA commands / vertex processing");rr3ds::Command command(g->m,"PICA command list",addr,size);{
g->jumpPending = false;
{int64_t hop = cast<int64_t>(0ULL);for (;;hop++){
if ((hop > cast<int64_t>(4096ULL))) {
arm_CPU_Halt(g->m->CPU,std::string("gpu: command-buffer chain exceeded %d hops (at 0x%08X) after %d instructions",76),cast<int64_t>(4096ULL),addr,g->m->CPU->Instrs);
return ;
}
time_Time t = n3ds_Machine_profStart(g->m);
if ((cast<uint32_t>(len(g->cmdBuf)) < size)) {
g->cmdBuf = Slice<uint8_t>::make(size);
}
Slice<uint8_t> buf = sub(g->cmdBuf,0,size);
{
auto tmp97 = n3ds_Machine_directRange(g->m,addr,size);
Slice<uint8_t> src = std::get<0>(tmp97);
uint32_t off = std::get<1>(tmp97);
if (bool(src)) {
gcopy(buf,sub(src,off,cast<uint32_t>((off + size))));
}
else {
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < size);i++){
buf[i] = n3ds_Machine_Read(g->m,cast<uint32_t>((addr + i)));
}
}}
}
if ((hop > cast<int64_t>(0ULL))) {
g->ListHops++;
if (g->m->GXCapture) {
n3ds_Machine_captureChainedList(g->m,addr,size,append(Slice<uint8_t>{},buf));
}
}
auto tmp98 = n3ds_DecodePICAInto(sub(g->cmdWrite,0,cast<int64_t>(0ULL)),buf);
Slice<n3ds_PICAWrite> ws = std::get<0>(tmp98);
Error err = std::get<1>(tmp98);
g->cmdWrite = ws;
n3ds_Machine_profEnd(g->m,cast<int64_t>(0ULL),t);
if (bool(err)) {
arm_CPU_Halt(g->m->CPU,std::string("gpu: %v (list at 0x%08X)",24),err,addr);
return ;
}
{auto&& tmp99 = ws;
for(int64_t tmp100=0;tmp100<len(tmp99);++tmp100){
auto w=tmp99[tmp100];n3ds_Machine* m = g->m;
if ((m->picaLimit > cast<int64_t>(0ULL))) {
if ((m->picaCount >= m->picaLimit)) {
m->StopRequested = true;
return ;
}
m->picaCount++;
}
if (bool(m->OnPICACmd)) {
m->OnPICACmd(w);
}
n3ds_GPU_write(g,w);
if (m->CPU->Halted) {
return ;
}
if (g->jumpPending) {
break;
}
}}
if ((!g->jumpPending)) {
return ;
}
auto tmp101 = std::make_tuple(g->jumpAddr,g->jumpSize);
addr = std::get<0>(tmp101);
size = std::get<1>(tmp101);
g->jumpPending = false;
}
}}
}
// tools/platform/n3ds/gpu.go:314:1
void n3ds_GPU_write(n3ds_GPU* g,n3ds_PICAWrite w){
{
if (((w.Mask == cast<uint8_t>(0ULL)) || (cast<int64_t>(w.Reg) >= cast<int64_t>(768ULL)))) {
return ;
}
uint32_t old = g->Regs[w.Reg];
uint32_t v = old;
{int64_t b = cast<int64_t>(0ULL);for (;(b < cast<int64_t>(4ULL));b++){
if ((cast<uint8_t>((w.Mask & (shl<uint8_t>(cast<uint8_t>(1ULL),b)))) != cast<uint8_t>(0ULL))) {
uint64_t sh = cast<uint64_t>(cast<int64_t>((cast<int64_t>(8ULL) * b)));
v = cast<uint32_t>(((v & ~((shl<uint32_t>(cast<uint32_t>(255ULL),sh)))) | cast<uint32_t>((w.Value & (shl<uint32_t>(cast<uint32_t>(255ULL),sh))))));
}
}
}g->Regs[w.Reg] = v;
{
if (((w.Reg == cast<uint16_t>(558ULL)) && (v != cast<uint32_t>(0ULL)))){
n3ds_GPU_drawArrays(g);
}
else if (((w.Reg == cast<uint16_t>(559ULL)) && (v != cast<uint32_t>(0ULL)))){
n3ds_GPU_drawElements(g);
}
else if (((w.Reg == cast<uint16_t>(572ULL)) || (w.Reg == cast<uint16_t>(573ULL)))){
uint16_t i = cast<uint16_t>((w.Reg - cast<uint16_t>(572ULL)));
g->jumpAddr = n3ds_Machine_gpuAddrToVirt(g->m,shl<uint32_t>(g->Regs[cast<uint16_t>((cast<uint16_t>(570ULL) + i))],cast<int64_t>(3ULL)));
g->jumpSize = shl<uint32_t>(g->Regs[cast<uint16_t>((cast<uint16_t>(568ULL) + i))],cast<int64_t>(3ULL));
g->jumpPending = (g->jumpSize != cast<uint32_t>(0ULL));
}
else if ((w.Reg == cast<uint16_t>(453ULL))){
g->lutIdx = cast<uint32_t>((v & cast<uint32_t>(255ULL)));
g->lutType = cast<uint32_t>((shr<uint32_t>(v,cast<int64_t>(8ULL)) & cast<uint32_t>(31ULL)));
}
else if (((w.Reg >= cast<uint16_t>(456ULL)) && (w.Reg < cast<uint16_t>(464ULL)))){
n3ds_GPU_lutWrite(g,v);
}
else if ((w.Reg == cast<uint16_t>(704ULL))){
g->fltIdx = cast<int64_t>(cast<uint32_t>((v & cast<uint32_t>(255ULL))));
g->fltF32 = (shr<uint32_t>(v,cast<int64_t>(31ULL)) != cast<uint32_t>(0ULL));
g->fltBuf = sub(g->fltBuf,0,cast<int64_t>(0ULL));
}
else if (((w.Reg >= cast<uint16_t>(705ULL)) && (w.Reg < cast<uint16_t>(713ULL)))){
n3ds_GPU_floatUniformWord(g,v);
}
else if ((w.Reg == cast<uint16_t>(715ULL))){
g->codeIdx = cast<int64_t>(cast<uint32_t>((v & cast<uint32_t>(4095ULL))));
}
else if (((w.Reg >= cast<uint16_t>(716ULL)) && (w.Reg < cast<uint16_t>(724ULL)))){
if ((g->codeIdx < cast<int64_t>(4096ULL))) {
g->Code[g->codeIdx] = v;
g->codeIdx++;
n3ds_GPU_invalidateShaders(g);
}
}
else if ((w.Reg == cast<uint16_t>(725ULL))){
g->opdIdx = cast<int64_t>(cast<uint32_t>((v & cast<uint32_t>(127ULL))));
}
else if (((w.Reg >= cast<uint16_t>(726ULL)) && (w.Reg < cast<uint16_t>(734ULL)))){
if ((g->opdIdx < cast<int64_t>(128ULL))) {
g->Opdesc[g->opdIdx] = v;
g->opdIdx++;
n3ds_GPU_invalidateShaders(g);
}
}
else if ((w.Reg == cast<uint16_t>(688ULL))){
g->Bool = cast<uint32_t>((v & cast<uint32_t>(65535ULL)));
}
else if (((w.Reg >= cast<uint16_t>(689ULL)) && (w.Reg < cast<uint16_t>(693ULL)))){
uint16_t i = cast<uint16_t>((w.Reg - cast<uint16_t>(689ULL)));
g->Int[i] = std::array<uint8_t,4>{cast<uint8_t>(v),cast<uint8_t>(shr<uint32_t>(v,cast<int64_t>(8ULL))),cast<uint8_t>(shr<uint32_t>(v,cast<int64_t>(16ULL))),cast<uint8_t>(shr<uint32_t>(v,cast<int64_t>(24ULL)))};
}
else if ((w.Reg == cast<uint16_t>(562ULL))){
g->fixedIdx = cast<int64_t>(cast<uint32_t>((v & cast<uint32_t>(15ULL))));
g->fixedBuf = sub(g->fixedBuf,0,cast<int64_t>(0ULL));
}
else if (((w.Reg >= cast<uint16_t>(563ULL)) && (w.Reg < cast<uint16_t>(566ULL)))){
g->fixedBuf = append(g->fixedBuf,v);
if ((len(g->fixedBuf) == cast<int64_t>(3ULL))) {
g->fixedVal[g->fixedIdx] = n3ds_unpackF24x4(g->fixedBuf[cast<int64_t>(0ULL)],g->fixedBuf[cast<int64_t>(1ULL)],g->fixedBuf[cast<int64_t>(2ULL)]);
g->fixedBuf = sub(g->fixedBuf,0,cast<int64_t>(0ULL));
}
}
else if ((w.Reg == cast<uint16_t>(656ULL))){
g->gshFltIdx = cast<int64_t>(cast<uint32_t>((v & cast<uint32_t>(255ULL))));
g->gshFltF32 = (shr<uint32_t>(v,cast<int64_t>(31ULL)) != cast<uint32_t>(0ULL));
g->gshFltBuf = sub(g->gshFltBuf,0,cast<int64_t>(0ULL));
}
else if ((w.Reg == cast<uint16_t>(667ULL))){
g->gshCodeIdx = cast<int64_t>(cast<uint32_t>((v & cast<uint32_t>(4095ULL))));
}
else if (((w.Reg >= cast<uint16_t>(668ULL)) && (w.Reg < cast<uint16_t>(676ULL)))){
if ((g->gshCodeIdx < cast<int64_t>(4096ULL))) {
g->Code[g->gshCodeIdx] = v;
g->gshCodeIdx++;
n3ds_GPU_invalidateShaders(g);
}
}
}
tmp102:;
}
}
// tools/platform/n3ds/gpu.go:413:1
void n3ds_GPU_floatUniformWord(n3ds_GPU* g,uint32_t v){
{
g->fltBuf = append(g->fltBuf,v);
if (g->fltF32) {
if ((len(g->fltBuf) == cast<int64_t>(4ULL))) {
if ((g->fltIdx < cast<int64_t>(96ULL))) {
std::array<float,4>* c = (&g->Float[g->fltIdx]);
(*c)[cast<int64_t>(0ULL)] = n3ds_toF24(n3ds_f32bits(g->fltBuf[cast<int64_t>(3ULL)]));
(*c)[cast<int64_t>(1ULL)] = n3ds_toF24(n3ds_f32bits(g->fltBuf[cast<int64_t>(2ULL)]));
(*c)[cast<int64_t>(2ULL)] = n3ds_toF24(n3ds_f32bits(g->fltBuf[cast<int64_t>(1ULL)]));
(*c)[cast<int64_t>(3ULL)] = n3ds_toF24(n3ds_f32bits(g->fltBuf[cast<int64_t>(0ULL)]));
}
g->fltIdx++;
g->fltBuf = sub(g->fltBuf,0,cast<int64_t>(0ULL));
}
return ;
}
if ((len(g->fltBuf) == cast<int64_t>(3ULL))) {
if ((g->fltIdx < cast<int64_t>(96ULL))) {
g->Float[g->fltIdx] = n3ds_unpackF24x4(g->fltBuf[cast<int64_t>(0ULL)],g->fltBuf[cast<int64_t>(1ULL)],g->fltBuf[cast<int64_t>(2ULL)]);
}
g->fltIdx++;
g->fltBuf = sub(g->fltBuf,0,cast<int64_t>(0ULL));
}
}
}
// tools/platform/n3ds/gpu.go:440:1
void n3ds_GPU_drawArrays(n3ds_GPU* g){
{
g->Draws++;
n3ds_GPU_draw(g,false);
}
}
// tools/platform/n3ds/gpu.go:445:1
void n3ds_GPU_drawElements(n3ds_GPU* g){
{
g->Draws++;
n3ds_GPU_draw(g,true);
}
}
// tools/platform/n3ds/gpu.go:452:1
bool n3ds_GPU_checkUnsupported(n3ds_GPU* g){
{
if ((cast<uint32_t>((g->Regs[cast<int64_t>(553ULL)] & cast<uint32_t>(3ULL))) != cast<uint32_t>(0ULL))) {
arm_CPU_Halt(g->m->CPU,std::string("gpu: draw with geometry-shader stage enabled (0x229=0x%08X) unimplemented",73),g->Regs[cast<int64_t>(553ULL)]);
return true;
}
return false;
}
}
// tools/platform/n3ds/gpu_float.go:12:1
float n3ds_f24bits(uint32_t v){
{
uint32_t sign = cast<uint32_t>((shr<uint32_t>(v,cast<int64_t>(23ULL)) & cast<uint32_t>(1ULL)));
uint32_t exp = cast<uint32_t>((shr<uint32_t>(v,cast<int64_t>(16ULL)) & cast<uint32_t>(127ULL)));
uint32_t man = cast<uint32_t>((v & cast<uint32_t>(65535ULL)));
if (((exp == cast<uint32_t>(0ULL)) && (man == cast<uint32_t>(0ULL)))) {
if ((sign != cast<uint32_t>(0ULL))) {
return cast<float>(go_math_Copysign(cast<double>(0.00000000000000000e+00),cast<double>(-1.00000000000000000e+00)));
}
return cast<float>(0.00000000000000000e+00);
}
uint32_t f = cast<uint32_t>((cast<uint32_t>((shl<uint32_t>(sign,cast<int64_t>(31ULL)) | shl<uint32_t>((cast<uint32_t>((exp + cast<uint32_t>(64ULL)))),cast<int64_t>(23ULL)))) | shl<uint32_t>(man,cast<int64_t>(7ULL))));
return go_math_Float32frombits(f);
}
}
// tools/platform/n3ds/gpu_float.go:28:1
float n3ds_f32bits(uint32_t v){
{
return go_math_Float32frombits(v);
}
}
// tools/platform/n3ds/gpu_float.go:34:1
std::array<float,4> n3ds_unpackF24x4(uint32_t w0,uint32_t w1,uint32_t w2){
{
return std::array<float,4>{n3ds_f24bits(cast<uint32_t>((w2 & cast<uint32_t>(16777215ULL)))),n3ds_f24bits(cast<uint32_t>((shl<uint32_t>((cast<uint32_t>((w1 & cast<uint32_t>(65535ULL)))),cast<int64_t>(8ULL)) | shr<uint32_t>(w2,cast<int64_t>(24ULL))))),n3ds_f24bits(cast<uint32_t>((shl<uint32_t>((cast<uint32_t>((w0 & cast<uint32_t>(255ULL)))),cast<int64_t>(16ULL)) | shr<uint32_t>(w1,cast<int64_t>(16ULL))))),n3ds_f24bits(shr<uint32_t>(w0,cast<int64_t>(8ULL)))};
}
}
// tools/platform/n3ds/gpu_float.go:57:1
float n3ds_toF24(float v){
{
uint32_t b = go_math_Float32bits(v);
auto tmp103 = std::make_tuple(shr<uint32_t>(b,cast<int64_t>(31ULL)),cast<int64_t>(cast<uint32_t>((shr<uint32_t>(b,cast<int64_t>(23ULL)) & cast<uint32_t>(255ULL)))),cast<uint32_t>((b & cast<uint32_t>(8388607ULL))));
uint32_t sign = std::get<0>(tmp103);
int64_t exp = std::get<1>(tmp103);
uint32_t man = std::get<2>(tmp103);
{
if ((exp == cast<int64_t>(255ULL))){
auto tmp105 = std::make_tuple(cast<int64_t>(0ULL),cast<uint32_t>(0ULL));
exp = std::get<0>(tmp105);
man = std::get<1>(tmp105);
}
else if ((exp == cast<int64_t>(0ULL))){
man = cast<uint32_t>(0ULL);
}
else {
int64_t e = cast<int64_t>((cast<int64_t>((exp - cast<int64_t>(127ULL))) + cast<int64_t>(63ULL)));
if ((e <= cast<int64_t>(0ULL))) {
auto tmp106 = std::make_tuple(cast<int64_t>(0ULL),cast<uint32_t>(0ULL));
exp = std::get<0>(tmp106);
man = std::get<1>(tmp106);
}
else if ((e >= cast<int64_t>(127ULL))) {
auto tmp107 = std::make_tuple(cast<int64_t>(189ULL),cast<uint32_t>(8323072ULL));
exp = std::get<0>(tmp107);
man = std::get<1>(tmp107);
}
else {
exp = cast<int64_t>((e + cast<int64_t>(64ULL)));
man &= ~(cast<uint32_t>(127ULL));
}
}
}
tmp104:;
if (((exp == cast<int64_t>(0ULL)) && (man == cast<uint32_t>(0ULL)))) {
if ((sign != cast<uint32_t>(0ULL))) {
return cast<float>(go_math_Copysign(cast<double>(0.00000000000000000e+00),cast<double>(-1.00000000000000000e+00)));
}
return cast<float>(0.00000000000000000e+00);
}
return go_math_Float32frombits(cast<uint32_t>((cast<uint32_t>((shl<uint32_t>(sign,cast<int64_t>(31ULL)) | shl<uint32_t>(cast<uint32_t>(exp),cast<int64_t>(23ULL)))) | man)));
}
}
// tools/platform/n3ds/gpu_light.go:82:1
std::string n3ds_lutName(int64_t t){
{
{
if ((t == cast<int64_t>(0ULL))){
return std::string("D0",2);
}
else if ((t == cast<int64_t>(1ULL))){
return std::string("D1",2);
}
else if ((t == cast<int64_t>(3ULL))){
return std::string("FR",2);
}
else if ((t == cast<int64_t>(4ULL))){
return std::string("RB",2);
}
else if ((t == cast<int64_t>(5ULL))){
return std::string("RG",2);
}
else if ((t == cast<int64_t>(6ULL))){
return std::string("RR",2);
}
else if (((t >= cast<int64_t>(8ULL)) && (t < cast<int64_t>(16ULL)))){
return go_fmt_Sprintf(std::string("SP%d",4),cast<int64_t>((t - cast<int64_t>(8ULL))));
}
else if (((t >= cast<int64_t>(16ULL)) && (t < cast<int64_t>(24ULL)))){
return go_fmt_Sprintf(std::string("DA%d",4),cast<int64_t>((t - cast<int64_t>(16ULL))));
}
}
tmp108:;
return std::string("?",1);
}
}
// tools/platform/n3ds/gpu_light.go:119:1
void n3ds_GPU_lutWrite(n3ds_GPU* g,uint32_t v){
{
auto tmp109 = std::make_tuple(cast<int64_t>(g->lutType),cast<int64_t>(g->lutIdx));
int64_t t = std::get<0>(tmp109);
int64_t i = std::get<1>(tmp109);
if (((t >= cast<int64_t>(24ULL)) || (i >= cast<int64_t>(256ULL)))) {
return ;
}
g->LUT[t][i] = cast<float>((cast<float>(cast<uint32_t>((v & cast<uint32_t>(4095ULL)))) / cast<float>(4.09500000000000000e+03)));
g->LUTDiff[t][i] = cast<float>((cast<float>(shr<int32_t>(cast<int32_t>(shl<uint32_t>(v,cast<int64_t>(8ULL))),cast<int64_t>(20ULL))) / cast<float>(4.09500000000000000e+03)));
g->lutSet[t] = true;
g->lutIdx = cast<uint32_t>(((cast<uint32_t>((g->lutIdx + cast<uint32_t>(1ULL)))) & cast<uint32_t>(255ULL)));
}
}
// tools/platform/n3ds/gpu_light.go:144:1
float n3ds_GPU_lutLookup(n3ds_GPU* g,int64_t t,uint8_t idx,float delta){
{
if ((t >= cast<int64_t>(24ULL))) {
return cast<float>(0.00000000000000000e+00);
}
return n3ds_clampf(cast<float>((g->LUT[t][idx] + cast<float>((g->LUTDiff[t][idx] * delta)))),cast<float>(0.00000000000000000e+00),cast<float>(1.00000000000000000e+00));
}
}
// tools/platform/n3ds/gpu_light.go:194:1
float n3ds_f16(uint32_t v){
{
v &= cast<uint32_t>(65535ULL);
float sign = cast<float>(1.00000000000000000e+00);
if ((cast<uint32_t>((v & cast<uint32_t>(32768ULL))) != cast<uint32_t>(0ULL))) {
sign = cast<float>(-1.00000000000000000e+00);
}
int64_t exp = cast<int64_t>((cast<int64_t>(shr<uint32_t>(v,cast<int64_t>(10ULL))) & cast<int64_t>(31ULL)));
int64_t man = cast<int64_t>(cast<uint32_t>((v & cast<uint32_t>(1023ULL))));
{
switch(exp){
case cast<int64_t>(0ULL):{
if ((man == cast<int64_t>(0ULL))) {
return cast<float>((sign * cast<float>(0.00000000000000000e+00)));
}
return cast<float>((cast<float>((cast<float>((sign * cast<float>(man))) / cast<float>(1.02400000000000000e+03))) / cast<float>(1.63840000000000000e+04)));
break;}
case cast<int64_t>(31ULL):{
return cast<float>((sign * cast<float>(1.00000001504746622e+30)));
break;}
}}
float f = cast<float>((cast<float>(1.00000000000000000e+00) + cast<float>((cast<float>(man) / cast<float>(1.02400000000000000e+03)))));
{int64_t e = cast<int64_t>((exp - cast<int64_t>(15ULL)));for (;(e > cast<int64_t>(0ULL));e--){
f *= cast<float>(2.00000000000000000e+00);
}
}{int64_t e = cast<int64_t>((exp - cast<int64_t>(15ULL)));for (;(e < cast<int64_t>(0ULL));e++){
f /= cast<float>(2.00000000000000000e+00);
}
}return cast<float>((sign * f));
}
}
// tools/platform/n3ds/gpu_light.go:224:1
float n3ds_f20(uint32_t v){
{
v &= cast<uint32_t>(1048575ULL);
int64_t exp = cast<int64_t>(cast<uint32_t>((shr<uint32_t>(v,cast<int64_t>(12ULL)) & cast<uint32_t>(127ULL))));
uint32_t man = cast<uint32_t>((v & cast<uint32_t>(4095ULL)));
if (((exp == cast<int64_t>(0ULL)) && (man == cast<uint32_t>(0ULL)))) {
return cast<float>(0.00000000000000000e+00);
}
float f = cast<float>(go_math_Ldexp(cast<double>((cast<double>(1.00000000000000000e+00) + cast<double>((cast<double>(man) / cast<double>(4.09600000000000000e+03))))),cast<int64_t>((exp - cast<int64_t>(63ULL)))));
if ((cast<uint32_t>((shr<uint32_t>(v,cast<int64_t>(19ULL)) & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL))) {
f = cast<float>(-f);
}
return f;
}
}
// tools/platform/n3ds/gpu_light.go:240:1
float n3ds_fixed11(uint32_t v){
{
return cast<float>((cast<float>(shr<int32_t>(cast<int32_t>(shl<uint32_t>(v,cast<int64_t>(19ULL))),cast<int64_t>(19ULL))) / cast<float>(2.04700000000000000e+03)));
}
}
// tools/platform/n3ds/gpu_light.go:250:1
std::array<float,3> n3ds_rgb10(uint32_t v){
{
constexpr double s=3.92156862745098034e-03;
return std::array<float,3>{cast<float>((cast<float>(cast<uint32_t>((shr<uint32_t>(v,cast<int64_t>(20ULL)) & cast<uint32_t>(1023ULL)))) * cast<float>(3.92156885936856270e-03))),cast<float>((cast<float>(cast<uint32_t>((shr<uint32_t>(v,cast<int64_t>(10ULL)) & cast<uint32_t>(1023ULL)))) * cast<float>(3.92156885936856270e-03))),cast<float>((cast<float>(cast<uint32_t>((v & cast<uint32_t>(1023ULL)))) * cast<float>(3.92156885936856270e-03)))};
}
}
// tools/platform/n3ds/gpu_light.go:260:1
n3ds_lightState n3ds_GPU_lightstate(n3ds_GPU* g){
{
auto tmp110 = std::make_tuple(g->Regs[cast<int64_t>(451ULL)],g->Regs[cast<int64_t>(452ULL)]);
uint32_t c0 = std::get<0>(tmp110);
uint32_t c1 = std::get<1>(tmp110);
n3ds_lightState ls = n3ds_lightState{(cast<uint32_t>((g->Regs[cast<int64_t>(143ULL)] & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL)),cast<int64_t>((cast<int64_t>(cast<uint32_t>((g->Regs[cast<int64_t>(450ULL)] & cast<uint32_t>(7ULL)))) + cast<int64_t>(1ULL))),n3ds_rgb10(g->Regs[cast<int64_t>(448ULL)]),{},cast<uint32_t>((shr<uint32_t>(c0,cast<int64_t>(4ULL)) & cast<uint32_t>(15ULL))),(cast<uint32_t>((c0 & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL)),cast<int64_t>(cast<uint32_t>((shr<uint32_t>(c0,cast<int64_t>(24ULL)) & cast<uint32_t>(3ULL)))),(cast<uint32_t>((shr<uint32_t>(c0,cast<int64_t>(18ULL)) & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL)),(cast<uint32_t>((shr<uint32_t>(c0,cast<int64_t>(16ULL)) & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL)),(cast<uint32_t>((shr<uint32_t>(c0,cast<int64_t>(17ULL)) & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL)),(cast<uint32_t>((shr<uint32_t>(c0,cast<int64_t>(19ULL)) & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL)),cast<uint32_t>((shr<uint32_t>(c0,cast<int64_t>(28ULL)) & cast<uint32_t>(3ULL))),cast<int64_t>(cast<uint32_t>((shr<uint32_t>(c0,cast<int64_t>(22ULL)) & cast<uint32_t>(3ULL)))),(cast<uint32_t>((shr<uint32_t>(c0,cast<int64_t>(30ULL)) & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL)),(cast<uint32_t>((shr<uint32_t>(c0,cast<int64_t>(27ULL)) & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL)),(cast<uint32_t>((shr<uint32_t>(c0,cast<int64_t>(2ULL)) & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL)),(cast<uint32_t>((shr<uint32_t>(c0,cast<int64_t>(3ULL)) & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL)),g->Regs[cast<int64_t>(465ULL)],g->Regs[cast<int64_t>(464ULL)],g->Regs[cast<int64_t>(466ULL)],(cast<uint32_t>((shr<uint32_t>(c1,cast<int64_t>(16ULL)) & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL)),(cast<uint32_t>((shr<uint32_t>(c1,cast<int64_t>(17ULL)) & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL)),(cast<uint32_t>((shr<uint32_t>(c1,cast<int64_t>(19ULL)) & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL)),(cast<uint32_t>((shr<uint32_t>(c1,cast<int64_t>(20ULL)) & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL)),(cast<uint32_t>((shr<uint32_t>(c1,cast<int64_t>(21ULL)) & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL)),(cast<uint32_t>((shr<uint32_t>(c1,cast<int64_t>(22ULL)) & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL))};
uint32_t perm = g->Regs[cast<int64_t>(473ULL)];
{int64_t slot = cast<int64_t>(0ULL);for (;(slot < ls.count);slot++){
uint32_t li = cast<uint32_t>((shr<uint32_t>(perm,(cast<uint64_t>((cast<uint64_t>(4ULL) * cast<uint64_t>(slot))))) & cast<uint32_t>(7ULL)));
uint32_t base = cast<uint32_t>((cast<uint32_t>(320ULL) + cast<uint32_t>((li * cast<uint32_t>(16ULL)))));
uint32_t cfg = g->Regs[cast<uint32_t>((base + cast<uint32_t>(9ULL)))];
ls.lights[slot] = n3ds_lightSource{cast<int64_t>(li),n3ds_rgb10(g->Regs[cast<uint32_t>((base + cast<uint32_t>(0ULL)))]),n3ds_rgb10(g->Regs[cast<uint32_t>((base + cast<uint32_t>(1ULL)))]),n3ds_rgb10(g->Regs[cast<uint32_t>((base + cast<uint32_t>(2ULL)))]),n3ds_rgb10(g->Regs[cast<uint32_t>((base + cast<uint32_t>(3ULL)))]),std::array<float,3>{n3ds_f16(g->Regs[cast<uint32_t>((base + cast<uint32_t>(4ULL)))]),n3ds_f16(shr<uint32_t>(g->Regs[cast<uint32_t>((base + cast<uint32_t>(4ULL)))],cast<int64_t>(16ULL))),n3ds_f16(g->Regs[cast<uint32_t>((base + cast<uint32_t>(5ULL)))])},std::array<float,3>{n3ds_fixed11(g->Regs[cast<uint32_t>((base + cast<uint32_t>(6ULL)))]),n3ds_fixed11(shr<uint32_t>(g->Regs[cast<uint32_t>((base + cast<uint32_t>(6ULL)))],cast<int64_t>(16ULL))),n3ds_fixed11(g->Regs[cast<uint32_t>((base + cast<uint32_t>(7ULL)))])},(cast<uint32_t>((cfg & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL)),(cast<uint32_t>((cfg & cast<uint32_t>(2ULL))) != cast<uint32_t>(0ULL)),(cast<uint32_t>((cfg & cast<uint32_t>(4ULL))) != cast<uint32_t>(0ULL)),(cast<uint32_t>((cfg & cast<uint32_t>(8ULL))) != cast<uint32_t>(0ULL)),n3ds_f20(g->Regs[cast<uint32_t>((base + cast<uint32_t>(10ULL)))]),n3ds_f20(g->Regs[cast<uint32_t>((base + cast<uint32_t>(11ULL)))]),(cast<uint32_t>((shr<uint32_t>(c1,(cast<uint32_t>((cast<uint32_t>(24ULL) + li)))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(0ULL)),(cast<uint32_t>((shr<uint32_t>(c1,(cast<uint32_t>((cast<uint32_t>(8ULL) + li)))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(0ULL)),(cast<uint32_t>((shr<uint32_t>(c1,li) & cast<uint32_t>(1ULL))) == cast<uint32_t>(0ULL))};
}
}return ls;
}
}
// tools/platform/n3ds/gpu_light.go:332:1
bool n3ds_lutSupported(uint32_t env,int64_t t){
{
{
switch(t){
case cast<int64_t>(0ULL):{
return (env != cast<uint32_t>(1ULL));
break;}
case cast<int64_t>(1ULL):{
return (((env != cast<uint32_t>(0ULL)) && (env != cast<uint32_t>(1ULL))) && (env != cast<uint32_t>(5ULL)));
break;}
case cast<int64_t>(3ULL):{
return (((env != cast<uint32_t>(0ULL)) && (env != cast<uint32_t>(2ULL))) && (env != cast<uint32_t>(4ULL)));
break;}
case cast<int64_t>(6ULL):{
return (env != cast<uint32_t>(3ULL));
break;}
case cast<int64_t>(5ULL):case cast<int64_t>(4ULL):{
return (((env == cast<uint32_t>(4ULL)) || (env == cast<uint32_t>(5ULL))) || (env == cast<uint32_t>(8ULL)));
break;}
}}
if (((t >= cast<int64_t>(8ULL)) && (t < cast<int64_t>(16ULL)))) {
return ((env != cast<uint32_t>(2ULL)) && (env != cast<uint32_t>(3ULL)));
}
return true;
}
}
// tools/platform/n3ds/gpu_light.go:352:1
float n3ds_lutScaleOf(uint32_t v){
{
{
switch(cast<uint32_t>((v & cast<uint32_t>(7ULL)))){
case cast<uint32_t>(0ULL):{
return cast<float>(1.00000000000000000e+00);
break;}
case cast<uint32_t>(1ULL):{
return cast<float>(2.00000000000000000e+00);
break;}
case cast<uint32_t>(2ULL):{
return cast<float>(4.00000000000000000e+00);
break;}
case cast<uint32_t>(3ULL):{
return cast<float>(8.00000000000000000e+00);
break;}
case cast<uint32_t>(6ULL):{
return cast<float>(2.50000000000000000e-01);
break;}
case cast<uint32_t>(7ULL):{
return cast<float>(5.00000000000000000e-01);
break;}
}}
return cast<float>(0.00000000000000000e+00);
}
}
// tools/platform/n3ds/gpu_light.go:375:1
std::tuple<std::array<float,4>,std::array<float,4>> n3ds_GPU_shade(n3ds_GPU* g,n3ds_lightState* ls,std::array<float,4> quat,std::array<float,3> view,std::array<n3ds_rgba,3>* tex){
std::array<float,4> prim{};
std::array<float,4> sec{};
{
std::array<float,4> shadow = std::array<float,4>{cast<float>(1.00000000000000000e+00),cast<float>(1.00000000000000000e+00),cast<float>(1.00000000000000000e+00),cast<float>(1.00000000000000000e+00)};
if (ls->shadow) {
n3ds_rgba s = (*tex)[n3ds_clampi(cast<int32_t>(ls->shadowSel),cast<int32_t>(0ULL),cast<int32_t>(2ULL))];
shadow = std::array<float,4>{cast<float>((cast<float>(s.r) / cast<float>(2.55000000000000000e+02))),cast<float>((cast<float>(s.g) / cast<float>(2.55000000000000000e+02))),cast<float>((cast<float>(s.b) / cast<float>(2.55000000000000000e+02))),cast<float>((cast<float>(s.a) / cast<float>(2.55000000000000000e+02)))};
if (ls->shadowInvert) {
{auto&& tmp111 = shadow;
for(int64_t tmp112=0;tmp112<len(tmp111);++tmp112){
auto i=tmp112;shadow[i] = cast<float>((cast<float>(1.00000000000000000e+00) - shadow[i]));
}}
}
}
std::array<float,3> sn = std::array<float,3>{cast<float>(0.00000000000000000e+00),cast<float>(0.00000000000000000e+00),cast<float>(1.00000000000000000e+00)};
std::array<float,3> st = std::array<float,3>{cast<float>(1.00000000000000000e+00),cast<float>(0.00000000000000000e+00),cast<float>(0.00000000000000000e+00)};
if ((ls->bumpMode != cast<uint32_t>(0ULL))) {
n3ds_rgba t = (*tex)[n3ds_clampi(cast<int32_t>(ls->bumpSel),cast<int32_t>(0ULL),cast<int32_t>(2ULL))];
std::array<float,3> p = std::array<float,3>{cast<float>((cast<float>((cast<float>(t.r) / cast<float>(1.27500000000000000e+02))) - cast<float>(1.00000000000000000e+00))),cast<float>((cast<float>((cast<float>(t.g) / cast<float>(1.27500000000000000e+02))) - cast<float>(1.00000000000000000e+00))),cast<float>((cast<float>((cast<float>(t.b) / cast<float>(1.27500000000000000e+02))) - cast<float>(1.00000000000000000e+00)))};
if ((ls->bumpMode == cast<uint32_t>(1ULL))) {
if ((!ls->noBumpRenorm)) {
float z = cast<float>((cast<float>(1.00000000000000000e+00) - (cast<float>((cast<float>((p[cast<int64_t>(0ULL)] * p[cast<int64_t>(0ULL)])) + cast<float>((p[cast<int64_t>(1ULL)] * p[cast<int64_t>(1ULL)])))))));
if ((z < cast<float>(0.00000000000000000e+00))) {
z = cast<float>(0.00000000000000000e+00);
}
p[cast<int64_t>(2ULL)] = n3ds_sqrt32(z);
}
sn = p;
}
else {
st = p;
}
}
std::array<float,3> normal = n3ds_quatRotate(quat,sn);
std::array<float,3> tangent = n3ds_quatRotate(quat,st);
std::array<float,3> nview = n3ds_normalize3(view);
prim = std::array<float,4>{cast<float>(0.00000000000000000e+00),cast<float>(0.00000000000000000e+00),cast<float>(0.00000000000000000e+00),cast<float>(1.00000000000000000e+00)};
sec = std::array<float,4>{cast<float>(0.00000000000000000e+00),cast<float>(0.00000000000000000e+00),cast<float>(0.00000000000000000e+00),cast<float>(1.00000000000000000e+00)};
{int64_t i = cast<int64_t>(0ULL);for (;(i < ls->count);i++){
n3ds_lightSource* l = (&ls->lights[i]);
std::array<float,3> ldir = l->pos;
if ((!l->directional)) {
ldir = std::array<float,3>{cast<float>((l->pos[cast<int64_t>(0ULL)] + view[cast<int64_t>(0ULL)])),cast<float>((l->pos[cast<int64_t>(1ULL)] + view[cast<int64_t>(1ULL)])),cast<float>((l->pos[cast<int64_t>(2ULL)] + view[cast<int64_t>(2ULL)]))};
}
float dist = n3ds_sqrt32(cast<float>((cast<float>((cast<float>((ldir[cast<int64_t>(0ULL)] * ldir[cast<int64_t>(0ULL)])) + cast<float>((ldir[cast<int64_t>(1ULL)] * ldir[cast<int64_t>(1ULL)])))) + cast<float>((ldir[cast<int64_t>(2ULL)] * ldir[cast<int64_t>(2ULL)])))));
ldir = n3ds_normalize3(ldir);
std::array<float,3> half = std::array<float,3>{cast<float>((nview[cast<int64_t>(0ULL)] + ldir[cast<int64_t>(0ULL)])),cast<float>((nview[cast<int64_t>(1ULL)] + ldir[cast<int64_t>(1ULL)])),cast<float>((nview[cast<int64_t>(2ULL)] + ldir[cast<int64_t>(2ULL)]))};
float distAtten = cast<float>(1.00000000000000000e+00);
if (l->distAtten) {
float loc = n3ds_clampf(cast<float>((cast<float>((l->attenScale * dist)) + l->attenBias)),cast<float>(0.00000000000000000e+00),cast<float>(1.00000000000000000e+00));
auto tmp113 = n3ds_lutIndexAbs(loc);
uint8_t idx = std::get<0>(tmp113);
float delta = std::get<1>(tmp113);
distAtten = n3ds_GPU_lutLookup(g,cast<int64_t>((cast<int64_t>(16ULL) + l->index)),idx,delta);
}
auto lut = [&](uint64_t field,int64_t t)->float{
uint32_t in = cast<uint32_t>((shr<uint32_t>(ls->lutIn,field) & cast<uint32_t>(7ULL)));
bool abs = (cast<uint32_t>((shr<uint32_t>(ls->lutAbs,(cast<uint64_t>((field + cast<uint64_t>(1ULL))))) & cast<uint32_t>(1ULL))) == cast<uint32_t>(0ULL));
float c={};
{
switch(in){
case cast<uint32_t>(0ULL):{
c = n3ds_dot3(normal,n3ds_normalize3(half));
break;}
case cast<uint32_t>(1ULL):{
c = n3ds_dot3(nview,n3ds_normalize3(half));
break;}
case cast<uint32_t>(2ULL):{
c = n3ds_dot3(normal,nview);
break;}
case cast<uint32_t>(3ULL):{
c = n3ds_dot3(ldir,normal);
break;}
case cast<uint32_t>(4ULL):{
c = n3ds_dot3(ldir,l->spotDir);
break;}
case cast<uint32_t>(5ULL):{
if ((ls->env == cast<uint32_t>(8ULL))) {
std::array<float,3> nh = n3ds_normalize3(half);
float d = n3ds_dot3(normal,nh);
std::array<float,3> proj = std::array<float,3>{cast<float>((nh[cast<int64_t>(0ULL)] - cast<float>((normal[cast<int64_t>(0ULL)] * d)))),cast<float>((nh[cast<int64_t>(1ULL)] - cast<float>((normal[cast<int64_t>(1ULL)] * d)))),cast<float>((nh[cast<int64_t>(2ULL)] - cast<float>((normal[cast<int64_t>(2ULL)] * d))))};
c = n3ds_dot3(proj,tangent);
}
break;}
}}
uint8_t idx={};
float delta={};
if (abs) {
if (l->twoSided) {
c = n3ds_absf(c);
}
else if ((c < cast<float>(0.00000000000000000e+00))) {
c = cast<float>(0.00000000000000000e+00);
}
auto tmp114 = n3ds_lutIndexAbs(c);
idx = std::get<0>(tmp114);
delta = std::get<1>(tmp114);
}
else {
float f = n3ds_floor32(cast<float>((c * cast<float>(1.28000000000000000e+02))));
int32_t si = cast<int32_t>(n3ds_clampf(f,cast<float>(-1.28000000000000000e+02),cast<float>(1.27000000000000000e+02)));
delta = cast<float>((cast<float>((c * cast<float>(1.28000000000000000e+02))) - cast<float>(si)));
idx = cast<uint8_t>(si);
}
return cast<float>((n3ds_lutScaleOf(shr<uint32_t>(ls->lutScale,field)) * n3ds_GPU_lutLookup(g,t,idx,delta)));
}
;
float spotAtten = cast<float>(1.00000000000000000e+00);
if ((l->spotAtten && n3ds_lutSupported(ls->env,cast<int64_t>(8ULL)))) {
spotAtten = lut(cast<uint64_t>(8ULL),cast<int64_t>((cast<int64_t>(8ULL) + l->index)));
}
float d0 = cast<float>(1.00000000000000000e+00);
if (((!ls->noD0) && n3ds_lutSupported(ls->env,cast<int64_t>(0ULL)))) {
d0 = lut(cast<uint64_t>(0ULL),cast<int64_t>(0ULL));
}
std::array<float,3> spec0 = n3ds_scale3(l->specular0,d0);
std::array<float,3> refl = std::array<float,3>{cast<float>(1.00000000000000000e+00),cast<float>(1.00000000000000000e+00),cast<float>(1.00000000000000000e+00)};
if (((!ls->noRR) && n3ds_lutSupported(ls->env,cast<int64_t>(6ULL)))) {
refl[cast<int64_t>(0ULL)] = lut(cast<uint64_t>(24ULL),cast<int64_t>(6ULL));
}
auto tmp115 = std::make_tuple(refl[cast<int64_t>(0ULL)],refl[cast<int64_t>(0ULL)]);
refl[cast<int64_t>(1ULL)] = std::get<0>(tmp115);
refl[cast<int64_t>(2ULL)] = std::get<1>(tmp115);
if (((!ls->noRG) && n3ds_lutSupported(ls->env,cast<int64_t>(5ULL)))) {
refl[cast<int64_t>(1ULL)] = lut(cast<uint64_t>(20ULL),cast<int64_t>(5ULL));
}
if (((!ls->noRB) && n3ds_lutSupported(ls->env,cast<int64_t>(4ULL)))) {
refl[cast<int64_t>(2ULL)] = lut(cast<uint64_t>(16ULL),cast<int64_t>(4ULL));
}
float d1 = cast<float>(1.00000000000000000e+00);
if (((!ls->noD1) && n3ds_lutSupported(ls->env,cast<int64_t>(1ULL)))) {
d1 = lut(cast<uint64_t>(4ULL),cast<int64_t>(1ULL));
}
std::array<float,3> spec1 = std::array<float,3>{cast<float>((cast<float>((d1 * refl[cast<int64_t>(0ULL)])) * l->specular1[cast<int64_t>(0ULL)])),cast<float>((cast<float>((d1 * refl[cast<int64_t>(1ULL)])) * l->specular1[cast<int64_t>(1ULL)])),cast<float>((cast<float>((d1 * refl[cast<int64_t>(2ULL)])) * l->specular1[cast<int64_t>(2ULL)]))};
if ((((i == cast<int64_t>((ls->count - cast<int64_t>(1ULL)))) && (!ls->noFR)) && n3ds_lutSupported(ls->env,cast<int64_t>(3ULL)))) {
float fr = lut(cast<uint64_t>(12ULL),cast<int64_t>(3ULL));
if (ls->primaryAlpha) {
prim[cast<int64_t>(3ULL)] = fr;
}
if (ls->secondAlpha) {
sec[cast<int64_t>(3ULL)] = fr;
}
}
float ndl = n3ds_dot3(ldir,normal);
if (l->twoSided) {
ndl = n3ds_absf(ndl);
}
else if ((ndl < cast<float>(0.00000000000000000e+00))) {
ndl = cast<float>(0.00000000000000000e+00);
}
float clampHL = cast<float>(1.00000000000000000e+00);
if ((ls->clampHighlight && (ndl == cast<float>(0.00000000000000000e+00)))) {
clampHL = cast<float>(0.00000000000000000e+00);
}
if ((l->geo0 || l->geo1)) {
float geo = n3ds_dot3(half,half);
if ((geo == cast<float>(0.00000000000000000e+00))) {
geo = cast<float>(0.00000000000000000e+00);
}
else {
geo = n3ds_minf(cast<float>((ndl / geo)),cast<float>(1.00000000000000000e+00));
}
if (l->geo0) {
spec0 = n3ds_scale3(spec0,geo);
}
if (l->geo1) {
spec1 = n3ds_scale3(spec1,geo);
}
}
auto tmp116 = std::make_tuple(std::array<float,3>{cast<float>(1.00000000000000000e+00),cast<float>(1.00000000000000000e+00),cast<float>(1.00000000000000000e+00)},std::array<float,3>{cast<float>(1.00000000000000000e+00),cast<float>(1.00000000000000000e+00),cast<float>(1.00000000000000000e+00)});
std::array<float,3> shPri = std::get<0>(tmp116);
std::array<float,3> shSec = std::get<1>(tmp116);
if ((ls->shadowPrimary && l->shadowed)) {
shPri = std::array<float,3>{shadow[cast<int64_t>(0ULL)],shadow[cast<int64_t>(1ULL)],shadow[cast<int64_t>(2ULL)]};
}
if ((ls->shadowSecond && l->shadowed)) {
shSec = std::array<float,3>{shadow[cast<int64_t>(0ULL)],shadow[cast<int64_t>(1ULL)],shadow[cast<int64_t>(2ULL)]};
}
float att = cast<float>((distAtten * spotAtten));
{int64_t c = cast<int64_t>(0ULL);for (;(c < cast<int64_t>(3ULL));c++){
prim[c] += cast<float>(((cast<float>((cast<float>((cast<float>((l->diffuse[c] * ndl)) * shPri[c])) + l->ambient[c]))) * att));
sec[c] += cast<float>((cast<float>((cast<float>(((cast<float>((spec0[c] + spec1[c]))) * clampHL)) * att)) * shSec[c]));
}
}}
}if (ls->shadowAlpha) {
if (ls->primaryAlpha) {
prim[cast<int64_t>(3ULL)] *= shadow[cast<int64_t>(3ULL)];
}
if (ls->secondAlpha) {
sec[cast<int64_t>(3ULL)] *= shadow[cast<int64_t>(3ULL)];
}
}
{int64_t c = cast<int64_t>(0ULL);for (;(c < cast<int64_t>(3ULL));c++){
prim[c] = n3ds_clampf(cast<float>((prim[c] + ls->ambient[c])),cast<float>(0.00000000000000000e+00),cast<float>(1.00000000000000000e+00));
sec[c] = n3ds_clampf(sec[c],cast<float>(0.00000000000000000e+00),cast<float>(1.00000000000000000e+00));
}
}prim[cast<int64_t>(3ULL)] = n3ds_clampf(prim[cast<int64_t>(3ULL)],cast<float>(0.00000000000000000e+00),cast<float>(1.00000000000000000e+00));
sec[cast<int64_t>(3ULL)] = n3ds_clampf(sec[cast<int64_t>(3ULL)],cast<float>(0.00000000000000000e+00),cast<float>(1.00000000000000000e+00));
return {prim,sec};
}
}
// tools/platform/n3ds/gpu_light.go:601:1
std::tuple<uint8_t,float> n3ds_lutIndexAbs(float v){
{
v = n3ds_clampf(v,cast<float>(0.00000000000000000e+00),cast<float>(1.00000000000000000e+00));
float f = n3ds_clampf(n3ds_floor32(cast<float>((v * cast<float>(2.56000000000000000e+02)))),cast<float>(0.00000000000000000e+00),cast<float>(2.55000000000000000e+02));
return {cast<uint8_t>(f),cast<float>((cast<float>((v * cast<float>(2.56000000000000000e+02))) - f))};
}
}
// tools/platform/n3ds/gpu_light.go:610:1
std::array<float,3> n3ds_quatRotate(std::array<float,4> q,std::array<float,3> v){
{
float l = n3ds_sqrt32(cast<float>((cast<float>((cast<float>((cast<float>((q[cast<int64_t>(0ULL)] * q[cast<int64_t>(0ULL)])) + cast<float>((q[cast<int64_t>(1ULL)] * q[cast<int64_t>(1ULL)])))) + cast<float>((q[cast<int64_t>(2ULL)] * q[cast<int64_t>(2ULL)])))) + cast<float>((q[cast<int64_t>(3ULL)] * q[cast<int64_t>(3ULL)])))));
if ((l == cast<float>(0.00000000000000000e+00))) {
return v;
}
auto tmp117 = std::make_tuple(cast<float>((q[cast<int64_t>(0ULL)] / l)),cast<float>((q[cast<int64_t>(1ULL)] / l)),cast<float>((q[cast<int64_t>(2ULL)] / l)),cast<float>((q[cast<int64_t>(3ULL)] / l)));
float x = std::get<0>(tmp117);
float y = std::get<1>(tmp117);
float z = std::get<2>(tmp117);
float w = std::get<3>(tmp117);
float tx = cast<float>((cast<float>(2.00000000000000000e+00) * (cast<float>((cast<float>((y * v[cast<int64_t>(2ULL)])) - cast<float>((z * v[cast<int64_t>(1ULL)])))))));
float ty = cast<float>((cast<float>(2.00000000000000000e+00) * (cast<float>((cast<float>((z * v[cast<int64_t>(0ULL)])) - cast<float>((x * v[cast<int64_t>(2ULL)])))))));
float tz = cast<float>((cast<float>(2.00000000000000000e+00) * (cast<float>((cast<float>((x * v[cast<int64_t>(1ULL)])) - cast<float>((y * v[cast<int64_t>(0ULL)])))))));
return std::array<float,3>{cast<float>((cast<float>((v[cast<int64_t>(0ULL)] + cast<float>((w * tx)))) + (cast<float>((cast<float>((y * tz)) - cast<float>((z * ty))))))),cast<float>((cast<float>((v[cast<int64_t>(1ULL)] + cast<float>((w * ty)))) + (cast<float>((cast<float>((z * tx)) - cast<float>((x * tz))))))),cast<float>((cast<float>((v[cast<int64_t>(2ULL)] + cast<float>((w * tz)))) + (cast<float>((cast<float>((x * ty)) - cast<float>((y * tx)))))))};
}
}
// tools/platform/n3ds/gpu_light.go:627:1
float n3ds_dot3(std::array<float,3> a,std::array<float,3> b){
{
return cast<float>((cast<float>((cast<float>((a[cast<int64_t>(0ULL)] * b[cast<int64_t>(0ULL)])) + cast<float>((a[cast<int64_t>(1ULL)] * b[cast<int64_t>(1ULL)])))) + cast<float>((a[cast<int64_t>(2ULL)] * b[cast<int64_t>(2ULL)]))));
}
}
// tools/platform/n3ds/gpu_light.go:629:1
std::array<float,3> n3ds_scale3(std::array<float,3> v,float s){
{
return std::array<float,3>{cast<float>((v[cast<int64_t>(0ULL)] * s)),cast<float>((v[cast<int64_t>(1ULL)] * s)),cast<float>((v[cast<int64_t>(2ULL)] * s))};
}
}
// tools/platform/n3ds/gpu_light.go:633:1
std::array<float,3> n3ds_normalize3(std::array<float,3> v){
{
float l = cast<float>((cast<float>((cast<float>((v[cast<int64_t>(0ULL)] * v[cast<int64_t>(0ULL)])) + cast<float>((v[cast<int64_t>(1ULL)] * v[cast<int64_t>(1ULL)])))) + cast<float>((v[cast<int64_t>(2ULL)] * v[cast<int64_t>(2ULL)]))));
if ((l <= cast<float>(0.00000000000000000e+00))) {
return std::array<float,3>{cast<float>(0.00000000000000000e+00),cast<float>(0.00000000000000000e+00),cast<float>(1.00000000000000000e+00)};
}
l = n3ds_sqrt32(l);
return std::array<float,3>{cast<float>((v[cast<int64_t>(0ULL)] / l)),cast<float>((v[cast<int64_t>(1ULL)] / l)),cast<float>((v[cast<int64_t>(2ULL)] / l))};
}
}
// tools/platform/n3ds/gpu_light.go:642:1
float n3ds_sqrt32(float f){
{
return cast<float>(go_math_Sqrt(cast<double>(f)));
}
}
// tools/platform/n3ds/gpu_light.go:644:1
float n3ds_absf(float f){
{
if ((f < cast<float>(0.00000000000000000e+00))) {
return cast<float>(-f);
}
return f;
}
}
// tools/platform/n3ds/gpu_light.go:651:1
float n3ds_minf(float a,float b){
{
if ((a < b)) {
return a;
}
return b;
}
}
// tools/platform/n3ds/gpu_light.go:658:1
float n3ds_clampf(float v,float lo,float hi){
{
if ((v < lo)) {
return lo;
}
if ((v > hi)) {
return hi;
}
return v;
}
}
// tools/platform/n3ds/gpu_light.go:671:1
void n3ds_GPU_dumpLighting(n3ds_GPU* g){
{
auto tmp118 = std::make_tuple(g->Regs[cast<int64_t>(451ULL)],g->Regs[cast<int64_t>(452ULL)]);
uint32_t c0 = std::get<0>(tmp118);
uint32_t c1 = std::get<1>(tmp118);
go_fmt_Printf(std::string("  lighting: on count=%d ambient=%v (raw %08X)\012",46),cast<int64_t>((cast<int64_t>(cast<uint32_t>((g->Regs[cast<int64_t>(450ULL)] & cast<uint32_t>(7ULL)))) + cast<int64_t>(1ULL))),n3ds_rgb10(g->Regs[cast<int64_t>(448ULL)]),g->Regs[cast<int64_t>(448ULL)]);
go_fmt_Printf(std::string("    cfg0=%08X shadow=%v env=%d shPri=%v shSec=%v shInv=%v shAlpha=%v shTex=%d bumpTex=%d bumpMode=%d noRenorm=%v clampHL=%v priA=%v secA=%v\012",140),c0,(cast<uint32_t>((c0 & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL)),cast<uint32_t>((shr<uint32_t>(c0,cast<int64_t>(4ULL)) & cast<uint32_t>(15ULL))),(cast<uint32_t>((shr<uint32_t>(c0,cast<int64_t>(16ULL)) & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL)),(cast<uint32_t>((shr<uint32_t>(c0,cast<int64_t>(17ULL)) & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL)),(cast<uint32_t>((shr<uint32_t>(c0,cast<int64_t>(18ULL)) & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL)),(cast<uint32_t>((shr<uint32_t>(c0,cast<int64_t>(19ULL)) & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL)),cast<uint32_t>((shr<uint32_t>(c0,cast<int64_t>(24ULL)) & cast<uint32_t>(3ULL))),cast<uint32_t>((shr<uint32_t>(c0,cast<int64_t>(22ULL)) & cast<uint32_t>(3ULL))),cast<uint32_t>((shr<uint32_t>(c0,cast<int64_t>(28ULL)) & cast<uint32_t>(3ULL))),(cast<uint32_t>((shr<uint32_t>(c0,cast<int64_t>(30ULL)) & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL)),(cast<uint32_t>((shr<uint32_t>(c0,cast<int64_t>(27ULL)) & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL)),(cast<uint32_t>((shr<uint32_t>(c0,cast<int64_t>(2ULL)) & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL)),(cast<uint32_t>((shr<uint32_t>(c0,cast<int64_t>(3ULL)) & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL)));
go_fmt_Printf(std::string("    cfg1=%08X noShadow=%02X noSpot=%02X noDist=%02X lutOff(d0=%v d1=%v fr=%v rr=%v rg=%v rb=%v)\012",96),c1,cast<uint32_t>((c1 & cast<uint32_t>(255ULL))),cast<uint32_t>((shr<uint32_t>(c1,cast<int64_t>(8ULL)) & cast<uint32_t>(255ULL))),cast<uint32_t>((shr<uint32_t>(c1,cast<int64_t>(24ULL)) & cast<uint32_t>(255ULL))),(cast<uint32_t>((shr<uint32_t>(c1,cast<int64_t>(16ULL)) & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL)),(cast<uint32_t>((shr<uint32_t>(c1,cast<int64_t>(17ULL)) & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL)),(cast<uint32_t>((shr<uint32_t>(c1,cast<int64_t>(19ULL)) & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL)),(cast<uint32_t>((shr<uint32_t>(c1,cast<int64_t>(20ULL)) & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL)),(cast<uint32_t>((shr<uint32_t>(c1,cast<int64_t>(21ULL)) & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL)),(cast<uint32_t>((shr<uint32_t>(c1,cast<int64_t>(22ULL)) & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL)));
go_fmt_Printf(std::string("    perm=%08X lutsel=%08X lutabs=%08X lutscale=%08X\012",52),g->Regs[cast<int64_t>(473ULL)],g->Regs[cast<int64_t>(465ULL)],g->Regs[cast<int64_t>(464ULL)],g->Regs[cast<int64_t>(466ULL)]);
n3ds_lightState ls = n3ds_GPU_lightstate(g);
{int64_t i = cast<int64_t>(0ULL);for (;(i < ls.count);i++){
n3ds_lightSource l = ls.lights[i];
uint32_t base = cast<uint32_t>((cast<uint32_t>(320ULL) + cast<uint32_t>((cast<uint32_t>(l.index) * cast<uint32_t>(16ULL)))));
go_fmt_Printf(std::string("    slot%d=light%d diffuse=%v(%08X) ambient=%v(%08X) spec0=%v spec1=%v\012",71),i,l.index,l.diffuse,g->Regs[cast<uint32_t>((base + cast<uint32_t>(2ULL)))],l.ambient,g->Regs[cast<uint32_t>((base + cast<uint32_t>(3ULL)))],l.specular0,l.specular1);
go_fmt_Printf(std::string("           pos=%v dir=%v twoSided=%v geo=%v/%v atten=%v(bias=%.4f scale=%.6f) spot=%v(%v) shadowed=%v\012",102),l.pos,l.directional,l.twoSided,l.geo0,l.geo1,l.distAtten,l.attenBias,l.attenScale,l.spotAtten,l.spotDir,l.shadowed);
}
}{int64_t t = cast<int64_t>(0ULL);for (;(t < cast<int64_t>(24ULL));t++){
if ((!g->lutSet[t])) {
continue;
}
go_fmt_Printf(std::string("    lut%-2d %-4s: [0]=%.3f [32]=%.3f [64]=%.3f [128]=%.3f [192]=%.3f [255]=%.3f\012",80),t,n3ds_lutName(t),g->LUT[t][cast<int64_t>(0ULL)],g->LUT[t][cast<int64_t>(32ULL)],g->LUT[t][cast<int64_t>(64ULL)],g->LUT[t][cast<int64_t>(128ULL)],g->LUT[t][cast<int64_t>(192ULL)],g->LUT[t][cast<int64_t>(255ULL)]);
}
}}
}
// tools/platform/n3ds/gpu_light.go:706:1
void n3ds_GPU_dumpFragment(n3ds_GPU* g){
{
uint32_t en = g->Regs[cast<int64_t>(128ULL)];
go_fmt_Printf(std::string("  texunits: cfg=%08X (tex0=%v tex1=%v tex2=%v) type=%d\012",55),en,(cast<uint32_t>((en & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL)),(cast<uint32_t>((shr<uint32_t>(en,cast<int64_t>(1ULL)) & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL)),(cast<uint32_t>((shr<uint32_t>(en,cast<int64_t>(2ULL)) & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL)),cast<uint32_t>((shr<uint32_t>(g->Regs[cast<int64_t>(131ULL)],cast<int64_t>(28ULL)) & cast<uint32_t>(7ULL))));
{int64_t u = cast<int64_t>(0ULL);for (;(u < cast<int64_t>(3ULL));u++){
if ((cast<uint32_t>((shr<uint32_t>(en,cast<uint64_t>(u)) & cast<uint32_t>(1ULL))) == cast<uint32_t>(0ULL))) {
continue;
}
auto tmp119 = n3ds_texUnitRegs(u);
uint32_t dimR = std::get<0>(tmp119);
uint32_t paramR = std::get<1>(tmp119);
uint32_t addrR = std::get<2>(tmp119);
uint32_t typR = std::get<3>(tmp119);
go_fmt_Printf(std::string("    tex%d %dx%d fmt=%d addr=0x%08X param=%08X\012",46),u,cast<uint32_t>((shr<uint32_t>(g->Regs[dimR],cast<int64_t>(16ULL)) & cast<uint32_t>(2047ULL))),cast<uint32_t>((g->Regs[dimR] & cast<uint32_t>(2047ULL))),cast<uint32_t>((g->Regs[typR] & cast<uint32_t>(15ULL))),n3ds_Machine_gpuAddrToVirt(g->m,shl<uint32_t>(g->Regs[addrR],cast<int64_t>(3ULL))),g->Regs[paramR]);
}
}int64_t nstage = cast<int64_t>(6ULL);
go_fmt_Printf(std::string("  tev: bufupd=%08X bufcol=%08X\012",31),g->Regs[cast<int64_t>(224ULL)],g->Regs[cast<int64_t>(253ULL)]);
{int64_t st = cast<int64_t>(0ULL);for (;(st < nstage);st++){
uint32_t b = n3ds_tevStageBase[st];
auto tmp120 = std::make_tuple(g->Regs[b],g->Regs[cast<uint32_t>((b + cast<uint32_t>(1ULL)))],g->Regs[cast<uint32_t>((b + cast<uint32_t>(2ULL)))]);
uint32_t src = std::get<0>(tmp120);
uint32_t opd = std::get<1>(tmp120);
uint32_t cmb = std::get<2>(tmp120);
uint32_t k = g->Regs[cast<uint32_t>((b + cast<uint32_t>(3ULL)))];
go_fmt_Printf(std::string("    stage%d rgb: %s(%s) %s(%s) %s(%s) op=%s | a: %s(%s) %s(%s) %s(%s) op=%s | scale=%d/%d konst=%08X rgb(%d,%d,%d)\012",115),st,n3ds_tevSrcName(cast<uint32_t>((src & cast<uint32_t>(15ULL)))),n3ds_tevCOpName(cast<uint32_t>((opd & cast<uint32_t>(15ULL)))),n3ds_tevSrcName(cast<uint32_t>((shr<uint32_t>(src,cast<int64_t>(4ULL)) & cast<uint32_t>(15ULL)))),n3ds_tevCOpName(cast<uint32_t>((shr<uint32_t>(opd,cast<int64_t>(4ULL)) & cast<uint32_t>(15ULL)))),n3ds_tevSrcName(cast<uint32_t>((shr<uint32_t>(src,cast<int64_t>(8ULL)) & cast<uint32_t>(15ULL)))),n3ds_tevCOpName(cast<uint32_t>((shr<uint32_t>(opd,cast<int64_t>(8ULL)) & cast<uint32_t>(15ULL)))),n3ds_tevOpName(cast<uint32_t>((cmb & cast<uint32_t>(15ULL)))),n3ds_tevSrcName(cast<uint32_t>((shr<uint32_t>(src,cast<int64_t>(16ULL)) & cast<uint32_t>(15ULL)))),n3ds_tevAOpName(cast<uint32_t>((shr<uint32_t>(opd,cast<int64_t>(12ULL)) & cast<uint32_t>(7ULL)))),n3ds_tevSrcName(cast<uint32_t>((shr<uint32_t>(src,cast<int64_t>(20ULL)) & cast<uint32_t>(15ULL)))),n3ds_tevAOpName(cast<uint32_t>((shr<uint32_t>(opd,cast<int64_t>(16ULL)) & cast<uint32_t>(7ULL)))),n3ds_tevSrcName(cast<uint32_t>((shr<uint32_t>(src,cast<int64_t>(24ULL)) & cast<uint32_t>(15ULL)))),n3ds_tevAOpName(cast<uint32_t>((shr<uint32_t>(opd,cast<int64_t>(20ULL)) & cast<uint32_t>(7ULL)))),n3ds_tevOpName(cast<uint32_t>((shr<uint32_t>(cmb,cast<int64_t>(16ULL)) & cast<uint32_t>(15ULL)))),shl<int64_t>(cast<int64_t>(1ULL),(cast<uint32_t>((g->Regs[cast<uint32_t>((b + cast<uint32_t>(4ULL)))] & cast<uint32_t>(3ULL))))),shl<int64_t>(cast<int64_t>(1ULL),(cast<uint32_t>((shr<uint32_t>(g->Regs[cast<uint32_t>((b + cast<uint32_t>(4ULL)))],cast<int64_t>(16ULL)) & cast<uint32_t>(3ULL))))),k,cast<uint32_t>((k & cast<uint32_t>(255ULL))),cast<uint32_t>((shr<uint32_t>(k,cast<int64_t>(8ULL)) & cast<uint32_t>(255ULL))),cast<uint32_t>((shr<uint32_t>(k,cast<int64_t>(16ULL)) & cast<uint32_t>(255ULL))));
}
}}
}
// tools/platform/n3ds/gpu_light.go:734:1
std::string n3ds_tevSrcName(uint32_t s){
{
{
switch(s){
case cast<uint32_t>(0ULL):{
return std::string("vtxcol",6);
break;}
case cast<uint32_t>(1ULL):{
return std::string("fragpri",7);
break;}
case cast<uint32_t>(2ULL):{
return std::string("fragsec",7);
break;}
case cast<uint32_t>(3ULL):{
return std::string("tex0",4);
break;}
case cast<uint32_t>(4ULL):{
return std::string("tex1",4);
break;}
case cast<uint32_t>(5ULL):{
return std::string("tex2",4);
break;}
case cast<uint32_t>(13ULL):{
return std::string("buffer",6);
break;}
case cast<uint32_t>(14ULL):{
return std::string("konst",5);
break;}
case cast<uint32_t>(15ULL):{
return std::string("prev",4);
break;}
}}
return go_fmt_Sprintf(std::string("src%d",5),s);
}
}
// tools/platform/n3ds/gpu_light.go:758:1
std::string n3ds_tevCOpName(uint32_t o){
{
{
switch(o){
case cast<uint32_t>(0ULL):{
return std::string("rgb",3);
break;}
case cast<uint32_t>(1ULL):{
return std::string("1-rgb",5);
break;}
case cast<uint32_t>(2ULL):{
return std::string("a",1);
break;}
case cast<uint32_t>(3ULL):{
return std::string("1-a",3);
break;}
case cast<uint32_t>(4ULL):{
return std::string("r",1);
break;}
case cast<uint32_t>(5ULL):{
return std::string("1-r",3);
break;}
case cast<uint32_t>(8ULL):{
return std::string("g",1);
break;}
case cast<uint32_t>(9ULL):{
return std::string("1-g",3);
break;}
case cast<uint32_t>(12ULL):{
return std::string("b",1);
break;}
case cast<uint32_t>(13ULL):{
return std::string("1-b",3);
break;}
}}
return go_fmt_Sprintf(std::string("cop%d",5),o);
}
}
// tools/platform/n3ds/gpu_light.go:784:1
std::string n3ds_tevAOpName(uint32_t o){
{
{
switch(o){
case cast<uint32_t>(0ULL):{
return std::string("a",1);
break;}
case cast<uint32_t>(1ULL):{
return std::string("1-a",3);
break;}
case cast<uint32_t>(2ULL):{
return std::string("r",1);
break;}
case cast<uint32_t>(3ULL):{
return std::string("1-r",3);
break;}
case cast<uint32_t>(4ULL):{
return std::string("g",1);
break;}
case cast<uint32_t>(5ULL):{
return std::string("1-g",3);
break;}
case cast<uint32_t>(6ULL):{
return std::string("b",1);
break;}
}}
return std::string("1-b",3);
}
}
// tools/platform/n3ds/gpu_light.go:804:1
std::string n3ds_tevOpName(uint32_t o){
{
{
switch(o){
case cast<uint32_t>(0ULL):{
return std::string("replace",7);
break;}
case cast<uint32_t>(1ULL):{
return std::string("modulate",8);
break;}
case cast<uint32_t>(2ULL):{
return std::string("add",3);
break;}
case cast<uint32_t>(3ULL):{
return std::string("addsigned",9);
break;}
case cast<uint32_t>(4ULL):{
return std::string("lerp",4);
break;}
case cast<uint32_t>(5ULL):{
return std::string("subtract",8);
break;}
case cast<uint32_t>(8ULL):{
return std::string("muladd",6);
break;}
case cast<uint32_t>(9ULL):{
return std::string("addmul",6);
break;}
}}
return go_fmt_Sprintf(std::string("op%d",4),o);
}
}
// tools/platform/n3ds/gpu_png.go:31:1
image_NRGBA* n3ds_Machine_PresentedImage(n3ds_Machine* m,n3ds_ScreenGeom g){
{
if (((((g.Dst == cast<uint32_t>(0ULL)) || (g.DstBPP == cast<uint32_t>(0ULL))) || (g.W == cast<uint32_t>(0ULL))) || (g.H == cast<uint32_t>(0ULL)))) {
return {};
}
return n3ds_Machine_decodeFB(m,g.Dst,g.W,g.H,g.DstStride,g.DstFmt,g.DstBPP);
}
}
// tools/platform/n3ds/gpu_png.go:38:1
image_NRGBA* n3ds_Machine_Framebuffer(n3ds_Machine* m,std::string screen){
{
n3ds_xferRecord rec = m->lastXferTop;
int64_t idx = cast<int64_t>(0ULL);
if ((screen == std::string("bottom",6))) {
auto tmp121 = std::make_tuple(m->lastXferBottom,cast<int64_t>(1ULL));
rec = std::get<0>(tmp121);
idx = std::get<1>(tmp121);
}
{
n3ds_FBPresent fb = n3ds_Machine_Scanout(m,idx);
if ((fb.Valid && (fb.AddrLeft != cast<uint32_t>(0ULL)))) {
rec.dst = fb.AddrLeft;
{
uint32_t bpp = n3ds_fbBPP(fb.Format);
if ((bpp != cast<uint32_t>(0ULL))) {
rec.bpp = bpp;
rec.format = cast<uint32_t>((fb.Format & cast<uint32_t>(7ULL)));
if (((bpp != cast<uint32_t>(0ULL)) && (fb.Stride != cast<uint32_t>(0ULL)))) {
rec.stride = divi<uint32_t>(fb.Stride,bpp);
}
}
}
if (((rec.w == cast<uint32_t>(0ULL)) || (rec.h == cast<uint32_t>(0ULL)))) {
rec.w = cast<uint32_t>(240ULL);
rec.h = cast<uint32_t>(400ULL);
if ((idx == cast<int64_t>(1ULL))) {
rec.h = cast<uint32_t>(320ULL);
}
}
}
}
if ((rec.dst == cast<uint32_t>(0ULL))) {
return {};
}
return n3ds_Machine_decodeFB(m,rec.dst,rec.w,rec.h,rec.stride,rec.format,rec.bpp);
}
}
// tools/platform/n3ds/gpu_png.go:72:1
image_NRGBA* n3ds_Machine_decodeFB(n3ds_Machine* m,uint32_t addr,uint32_t fw,uint32_t fh,uint32_t stride,uint32_t format,uint32_t bpp){
{
auto tmp122 = std::make_tuple(cast<int64_t>(fw),cast<int64_t>(fh));
int64_t w = std::get<0>(tmp122);
int64_t h = std::get<1>(tmp122);
image_NRGBA* img = go_image_NewNRGBA(go_image_Rect(cast<int64_t>(0ULL),cast<int64_t>(0ULL),h,w));
{int64_t y = cast<int64_t>(0ULL);for (;(y < h);y++){
{int64_t x = cast<int64_t>(0ULL);for (;(x < w);x++){
uint32_t p = cast<uint32_t>((addr + cast<uint32_t>((cast<uint32_t>(cast<int64_t>((cast<int64_t>((y * cast<int64_t>(stride))) + x))) * bpp))));
uint8_t r={};
uint8_t g={};
uint8_t b={};
{
switch(format){
case cast<uint32_t>(0ULL):{
auto tmp123 = std::make_tuple(n3ds_Machine_Read(m,cast<uint32_t>((p + cast<uint32_t>(1ULL)))),n3ds_Machine_Read(m,cast<uint32_t>((p + cast<uint32_t>(2ULL)))),n3ds_Machine_Read(m,cast<uint32_t>((p + cast<uint32_t>(3ULL)))));
b = std::get<0>(tmp123);
g = std::get<1>(tmp123);
r = std::get<2>(tmp123);
break;}
case cast<uint32_t>(1ULL):{
auto tmp124 = std::make_tuple(n3ds_Machine_Read(m,p),n3ds_Machine_Read(m,cast<uint32_t>((p + cast<uint32_t>(1ULL)))),n3ds_Machine_Read(m,cast<uint32_t>((p + cast<uint32_t>(2ULL)))));
b = std::get<0>(tmp124);
g = std::get<1>(tmp124);
r = std::get<2>(tmp124);
break;}
case cast<uint32_t>(2ULL):{
uint16_t v = cast<uint16_t>((cast<uint16_t>(n3ds_Machine_Read(m,p)) | shl<uint16_t>(cast<uint16_t>(n3ds_Machine_Read(m,cast<uint32_t>((p + cast<uint32_t>(1ULL))))),cast<int64_t>(8ULL))));
auto tmp125 = std::make_tuple(cast<uint8_t>((cast<uint8_t>(shr<uint16_t>(v,cast<int64_t>(11ULL))) & cast<uint8_t>(31ULL))),cast<uint8_t>((cast<uint8_t>(shr<uint16_t>(v,cast<int64_t>(5ULL))) & cast<uint8_t>(63ULL))),cast<uint8_t>((cast<uint8_t>(v) & cast<uint8_t>(31ULL))));
uint8_t r5 = std::get<0>(tmp125);
uint8_t g6 = std::get<1>(tmp125);
uint8_t b5 = std::get<2>(tmp125);
auto tmp126 = std::make_tuple(cast<uint8_t>((shl<uint8_t>(r5,cast<int64_t>(3ULL)) | shr<uint8_t>(r5,cast<int64_t>(2ULL)))),cast<uint8_t>((shl<uint8_t>(g6,cast<int64_t>(2ULL)) | shr<uint8_t>(g6,cast<int64_t>(4ULL)))),cast<uint8_t>((shl<uint8_t>(b5,cast<int64_t>(3ULL)) | shr<uint8_t>(b5,cast<int64_t>(2ULL)))));
r = std::get<0>(tmp126);
g = std::get<1>(tmp126);
b = std::get<2>(tmp126);
break;}
case cast<uint32_t>(3ULL):{
uint16_t v = cast<uint16_t>((cast<uint16_t>(n3ds_Machine_Read(m,p)) | shl<uint16_t>(cast<uint16_t>(n3ds_Machine_Read(m,cast<uint32_t>((p + cast<uint32_t>(1ULL))))),cast<int64_t>(8ULL))));
auto tmp127 = std::make_tuple(cast<uint8_t>((cast<uint8_t>(shr<uint16_t>(v,cast<int64_t>(11ULL))) & cast<uint8_t>(31ULL))),cast<uint8_t>((cast<uint8_t>(shr<uint16_t>(v,cast<int64_t>(6ULL))) & cast<uint8_t>(31ULL))),cast<uint8_t>((cast<uint8_t>(shr<uint16_t>(v,cast<int64_t>(1ULL))) & cast<uint8_t>(31ULL))));
uint8_t r5 = std::get<0>(tmp127);
uint8_t g5 = std::get<1>(tmp127);
uint8_t b5 = std::get<2>(tmp127);
auto tmp128 = std::make_tuple(cast<uint8_t>((shl<uint8_t>(r5,cast<int64_t>(3ULL)) | shr<uint8_t>(r5,cast<int64_t>(2ULL)))),cast<uint8_t>((shl<uint8_t>(g5,cast<int64_t>(3ULL)) | shr<uint8_t>(g5,cast<int64_t>(2ULL)))),cast<uint8_t>((shl<uint8_t>(b5,cast<int64_t>(3ULL)) | shr<uint8_t>(b5,cast<int64_t>(2ULL)))));
r = std::get<0>(tmp128);
g = std::get<1>(tmp128);
b = std::get<2>(tmp128);
break;}
case cast<uint32_t>(4ULL):{
uint16_t v = cast<uint16_t>((cast<uint16_t>(n3ds_Machine_Read(m,p)) | shl<uint16_t>(cast<uint16_t>(n3ds_Machine_Read(m,cast<uint32_t>((p + cast<uint32_t>(1ULL))))),cast<int64_t>(8ULL))));
auto tmp129 = std::make_tuple(cast<uint8_t>((cast<uint8_t>(shr<uint16_t>(v,cast<int64_t>(12ULL))) * cast<uint8_t>(17ULL))),cast<uint8_t>((cast<uint8_t>(cast<uint16_t>((shr<uint16_t>(v,cast<int64_t>(8ULL)) & cast<uint16_t>(15ULL)))) * cast<uint8_t>(17ULL))),cast<uint8_t>((cast<uint8_t>(cast<uint16_t>((shr<uint16_t>(v,cast<int64_t>(4ULL)) & cast<uint16_t>(15ULL)))) * cast<uint8_t>(17ULL))));
r = std::get<0>(tmp129);
g = std::get<1>(tmp129);
b = std::get<2>(tmp129);
break;}
}}
int64_t o = image_NRGBA_PixOffset(img,y,cast<int64_t>((cast<int64_t>((w - cast<int64_t>(1ULL))) - x)));
auto tmp130 = std::make_tuple(r,g,b,cast<uint8_t>(255ULL));
img->Pix[o] = std::get<0>(tmp130);
img->Pix[cast<int64_t>((o + cast<int64_t>(1ULL)))] = std::get<1>(tmp130);
img->Pix[cast<int64_t>((o + cast<int64_t>(2ULL)))] = std::get<2>(tmp130);
img->Pix[cast<int64_t>((o + cast<int64_t>(3ULL)))] = std::get<3>(tmp130);
}
}}
}return img;
}
}
// tools/platform/n3ds/gpu_png.go:112:1
image_NRGBA* n3ds_Machine_RenderTarget(n3ds_Machine* m,uint32_t addr,uint32_t w,uint32_t h){
{
image_NRGBA* img = go_image_NewNRGBA(go_image_Rect(cast<int64_t>(0ULL),cast<int64_t>(0ULL),cast<int64_t>(w),cast<int64_t>(h)));
{uint32_t y = cast<uint32_t>(0ULL);for (;(y < h);y++){
{uint32_t x = cast<uint32_t>(0ULL);for (;(x < w);x++){
uint32_t p = cast<uint32_t>((addr + n3ds_tiledOffset(x,y,w)));
auto tmp131 = std::make_tuple(n3ds_Machine_Read(m,cast<uint32_t>((p + cast<uint32_t>(1ULL)))),n3ds_Machine_Read(m,cast<uint32_t>((p + cast<uint32_t>(2ULL)))),n3ds_Machine_Read(m,cast<uint32_t>((p + cast<uint32_t>(3ULL)))));
uint8_t b = std::get<0>(tmp131);
uint8_t g = std::get<1>(tmp131);
uint8_t r = std::get<2>(tmp131);
int64_t o = image_NRGBA_PixOffset(img,cast<int64_t>(x),cast<int64_t>(y));
auto tmp132 = std::make_tuple(r,g,b,cast<uint8_t>(255ULL));
img->Pix[o] = std::get<0>(tmp132);
img->Pix[cast<int64_t>((o + cast<int64_t>(1ULL)))] = std::get<1>(tmp132);
img->Pix[cast<int64_t>((o + cast<int64_t>(2ULL)))] = std::get<2>(tmp132);
img->Pix[cast<int64_t>((o + cast<int64_t>(3ULL)))] = std::get<3>(tmp132);
}
}}
}return img;
}
}
// tools/platform/n3ds/gpu_png.go:183:1
std::tuple<int64_t,int64_t> n3ds_ScreenGeom_Size(n3ds_ScreenGeom g){
int64_t w{};
int64_t h{};
{
return {cast<int64_t>(g.H),cast<int64_t>(g.W)};
}
}
// tools/platform/n3ds/gpu_png.go:188:1
std::tuple<uint32_t,uint32_t,bool> n3ds_ScreenGeom_Source(n3ds_ScreenGeom g,int64_t sx,int64_t sy){
uint32_t x{};
uint32_t y{};
bool ok{};
{
auto tmp133 = n3ds_ScreenGeom_Size(g);
int64_t iw = std::get<0>(tmp133);
int64_t ih = std::get<1>(tmp133);
if (((((sx < cast<int64_t>(0ULL)) || (sy < cast<int64_t>(0ULL))) || (sx >= iw)) || (sy >= ih))) {
return {cast<uint32_t>(0ULL),cast<uint32_t>(0ULL),false};
}
int64_t fx = cast<int64_t>((cast<int64_t>((cast<int64_t>(g.W) - cast<int64_t>(1ULL))) - sy));
int64_t fy = sx;
if (g.Flip) {
fy = cast<int64_t>((cast<int64_t>((cast<int64_t>(g.H) - cast<int64_t>(1ULL))) - fy));
}
if (((((fx < cast<int64_t>(0ULL)) || (fy < cast<int64_t>(0ULL))) || (fx >= cast<int64_t>(g.W))) || (fy >= cast<int64_t>(g.H)))) {
return {cast<uint32_t>(0ULL),cast<uint32_t>(0ULL),false};
}
return {cast<uint32_t>(fx),cast<uint32_t>(fy),true};
}
}
// tools/platform/n3ds/gpu_png.go:208:1
std::tuple<n3ds_ScreenGeom,bool> n3ds_Machine_ScreenGeom(n3ds_Machine* m,std::string screen){
{
n3ds_xferRecord rec = m->lastXferTop;
if ((screen == std::string("bottom",6))) {
rec = m->lastXferBottom;
}
if (((((rec.dst == cast<uint32_t>(0ULL)) || (rec.srcW == cast<uint32_t>(0ULL))) || (rec.w == cast<uint32_t>(0ULL))) || (rec.h == cast<uint32_t>(0ULL)))) {
return {n3ds_ScreenGeom{},false};
}
return {n3ds_ScreenGeom{rec.src,rec.srcW,rec.w,rec.h,rec.flip,(screen == std::string("bottom",6)),rec.dst,rec.stride,rec.format,rec.bpp},true};
}
}
// tools/platform/n3ds/gpu_png.go:231:1
image_NRGBA* n3ds_Machine_ScreenImage(n3ds_Machine* m,n3ds_ScreenGeom g){
{
auto tmp134 = n3ds_ScreenGeom_Size(g);
int64_t iw = std::get<0>(tmp134);
int64_t ih = std::get<1>(tmp134);
image_NRGBA* img = go_image_NewNRGBA(go_image_Rect(cast<int64_t>(0ULL),cast<int64_t>(0ULL),iw,ih));
{int64_t sy = cast<int64_t>(0ULL);for (;(sy < ih);sy++){
{int64_t sx = cast<int64_t>(0ULL);for (;(sx < iw);sx++){
auto tmp135 = n3ds_ScreenGeom_Source(g,sx,sy);
uint32_t x = std::get<0>(tmp135);
uint32_t y = std::get<1>(tmp135);
bool ok = std::get<2>(tmp135);
if ((!ok)) {
continue;
}
uint32_t p = cast<uint32_t>((g.Src + n3ds_tiledOffset(x,y,g.SrcW)));
auto tmp136 = std::make_tuple(n3ds_Machine_Read(m,cast<uint32_t>((p + cast<uint32_t>(1ULL)))),n3ds_Machine_Read(m,cast<uint32_t>((p + cast<uint32_t>(2ULL)))),n3ds_Machine_Read(m,cast<uint32_t>((p + cast<uint32_t>(3ULL)))));
uint8_t b = std::get<0>(tmp136);
uint8_t gr = std::get<1>(tmp136);
uint8_t r = std::get<2>(tmp136);
int64_t o = image_NRGBA_PixOffset(img,sx,sy);
auto tmp137 = std::make_tuple(r,gr,b,cast<uint8_t>(255ULL));
img->Pix[o] = std::get<0>(tmp137);
img->Pix[cast<int64_t>((o + cast<int64_t>(1ULL)))] = std::get<1>(tmp137);
img->Pix[cast<int64_t>((o + cast<int64_t>(2ULL)))] = std::get<2>(tmp137);
img->Pix[cast<int64_t>((o + cast<int64_t>(3ULL)))] = std::get<3>(tmp137);
}
}}
}return img;
}
}
// tools/platform/n3ds/gpu_png.go:252:1
uint32_t n3ds_fbBPP(uint32_t format){
{
{
switch(cast<uint32_t>((format & cast<uint32_t>(7ULL)))){
case cast<uint32_t>(0ULL):{
return cast<uint32_t>(4ULL);
break;}
case cast<uint32_t>(1ULL):{
return cast<uint32_t>(3ULL);
break;}
case cast<uint32_t>(2ULL):case cast<uint32_t>(3ULL):case cast<uint32_t>(4ULL):{
return cast<uint32_t>(2ULL);
break;}
}}
return cast<uint32_t>(0ULL);
}
}
// tools/platform/n3ds/gpu_raster.go:46:1
void n3ds_GPU_draw(n3ds_GPU* g,bool indexed){
rrperf::Draw drawProfile(g,indexed);
{
if ((((g->TraceDraws > cast<int64_t>(0ULL)) && (g->Draws >= g->TraceFrom)) && (cast<uint32_t>((g->Regs[cast<int64_t>(143ULL)] & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL)))) {
n3ds_GPU_dumpLighting(g);
n3ds_GPU_dumpFragment(g);
}
if (n3ds_GPU_checkUnsupported(g)) {
return ;
}
n3ds_Machine* m = g->m;
uint32_t physBase = shl<uint32_t>(g->Regs[cast<int64_t>(512ULL)],cast<int64_t>(3ULL));
uint32_t base = n3ds_Machine_gpuAddrToVirt(m,physBase);
uint64_t fmtWord = cast<uint64_t>((cast<uint64_t>(g->Regs[cast<int64_t>(513ULL)]) | shl<uint64_t>(cast<uint64_t>(cast<uint32_t>((g->Regs[cast<int64_t>(514ULL)] & cast<uint32_t>(65535ULL)))),cast<int64_t>(32ULL))));
uint32_t fixedMask = cast<uint32_t>((shr<uint32_t>(g->Regs[cast<int64_t>(514ULL)],cast<int64_t>(16ULL)) & cast<uint32_t>(4095ULL)));
auto tmp138 = std::make_tuple(sub(g->bufs,0,cast<int64_t>(0ULL)),sub(g->comps,0,cast<int64_t>(0ULL)));
Slice<n3ds_loaderBuf> bufs = std::get<0>(tmp138);
Slice<int64_t> comps = std::get<1>(tmp138);
{int64_t b = cast<int64_t>(0ULL);for (;(b < cast<int64_t>(12ULL));b++){
uint32_t cfg1 = g->Regs[cast<uint32_t>((cast<uint32_t>(516ULL) + cast<uint32_t>((cast<uint32_t>(b) * cast<uint32_t>(3ULL)))))];
uint32_t cfg2 = g->Regs[cast<uint32_t>((cast<uint32_t>(517ULL) + cast<uint32_t>((cast<uint32_t>(b) * cast<uint32_t>(3ULL)))))];
int64_t n = cast<int64_t>(shr<uint32_t>(cfg2,cast<int64_t>(28ULL)));
if ((n == cast<int64_t>(0ULL))) {
continue;
}
uint64_t perm = cast<uint64_t>((cast<uint64_t>(cfg1) | shl<uint64_t>(cast<uint64_t>(cast<uint32_t>((cfg2 & cast<uint32_t>(65535ULL)))),cast<int64_t>(32ULL))));
int64_t start = len(comps);
{int64_t j = cast<int64_t>(0ULL);for (;(j < n);j++){
comps = append(comps,cast<int64_t>(cast<uint64_t>((shr<uint64_t>(perm,(cast<uint64_t>((cast<uint64_t>(4ULL) * cast<uint64_t>(j))))) & cast<uint64_t>(15ULL)))));
}
}bufs = append(bufs,n3ds_loaderBuf{g->Regs[cast<uint32_t>((cast<uint32_t>(515ULL) + cast<uint32_t>((cast<uint32_t>(b) * cast<uint32_t>(3ULL)))))],start,n,cast<uint32_t>((shr<uint32_t>(cfg2,cast<int64_t>(16ULL)) & cast<uint32_t>(255ULL)))});
}
}auto tmp139 = std::make_tuple(bufs,comps);
g->bufs = std::get<0>(tmp139);
g->comps = std::get<1>(tmp139);
uint32_t count = g->Regs[cast<int64_t>(552ULL)];
uint32_t first = g->Regs[cast<int64_t>(554ULL)];
uint32_t idxCfg = g->Regs[cast<int64_t>(551ULL)];
uint32_t idxAddr = n3ds_Machine_gpuAddrToVirt(m,cast<uint32_t>((physBase + cast<uint32_t>((idxCfg & cast<uint32_t>(268435455ULL))))));
bool idx16 = (shr<uint32_t>(idxCfg,cast<int64_t>(31ULL)) != cast<uint32_t>(0ULL));
uint64_t inPerm = cast<uint64_t>((cast<uint64_t>(g->Regs[cast<int64_t>(699ULL)]) | shl<uint64_t>(cast<uint64_t>(g->Regs[cast<int64_t>(700ULL)]),cast<int64_t>(32ULL))));
if (bool(g->Census)) {
go_fmt_Fprintf(g->Census,std::string("draw %d indexed=%v count=%d base=0x%08X fmt=%08X/%08X inPerm=%08X_%08X entry=0x%03X",83),g->Draws,indexed,count,base,g->Regs[cast<int64_t>(513ULL)],g->Regs[cast<int64_t>(514ULL)],g->Regs[cast<int64_t>(700ULL)],g->Regs[cast<int64_t>(699ULL)],cast<uint32_t>((g->Regs[cast<int64_t>(698ULL)] & cast<uint32_t>(4095ULL))));
{auto&& tmp140 = bufs;
for(int64_t tmp141=0;tmp141<len(tmp140);++tmp141){
auto b=tmp140[tmp141];go_fmt_Fprintf(g->Census,std::string(" | buf off=0x%06X stride=%d comps=%v",36),b.off,b.stride,sub(comps,b.first,cast<int64_t>((b.first + b.n))));
}}
go_fmt_Fprintln(g->Census);
}
bool trace = ((g->TraceDraws > cast<int64_t>(0ULL)) && (g->Draws >= g->TraceFrom));
if (trace) {
g->TraceDraws--;
n3ds_fbState tfb = n3ds_GPU_fbstate(g);
go_fmt_Printf(std::string("  FETCH base=0x%08X (reg 0x200=0x%08X) fmt=%08X/%08X fixedMask=%03X nattr=%d idxCfg=0x%08X (16bit=%v addr=0x%08X) inPerm=%08X_%08X maxIn=%d\012",140),base,g->Regs[cast<int64_t>(512ULL)],g->Regs[cast<int64_t>(513ULL)],g->Regs[cast<int64_t>(514ULL)],fixedMask,cast<uint32_t>((cast<uint32_t>((shr<uint32_t>(g->Regs[cast<int64_t>(514ULL)],cast<int64_t>(28ULL)) & cast<uint32_t>(15ULL))) + cast<uint32_t>(1ULL))),idxCfg,idx16,idxAddr,g->Regs[cast<int64_t>(700ULL)],g->Regs[cast<int64_t>(699ULL)],cast<uint32_t>((g->Regs[cast<int64_t>(697ULL)] & cast<uint32_t>(15ULL))));
{auto&& tmp142 = bufs;
for(int64_t tmp143=0;tmp143<len(tmp142);++tmp143){
auto bi=tmp143;auto b=tmp142[tmp143];go_fmt_Printf(std::string("    buf%d off=0x%06X stride=%d comps=%v -> addr(v0)=0x%08X\012",59),bi,b.off,b.stride,sub(comps,b.first,cast<int64_t>((b.first + b.n))),n3ds_Machine_gpuAddrToVirt(m,cast<uint32_t>((physBase + b.off))));
}}
{int64_t a = cast<int64_t>(0ULL);for (;(a < cast<int64_t>(12ULL));a++){
if ((cast<uint32_t>((shr<uint32_t>(fixedMask,cast<uint64_t>(a)) & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL))) {
go_fmt_Printf(std::string("    fixed attr%d = %v\012",22),a,g->fixedVal[a]);
}
}
}go_fmt_Printf(std::string("  VIEWPORT offset x=%d y=%d (reg 0x068=0x%08X) depthmap=0x%08X(%s) scale=%f off=%f 0x107=%08X\012",94),cast<uint32_t>((g->Regs[cast<int64_t>(104ULL)] & cast<uint32_t>(1023ULL))),cast<uint32_t>((shr<uint32_t>(g->Regs[cast<int64_t>(104ULL)],cast<int64_t>(16ULL)) & cast<uint32_t>(1023ULL))),g->Regs[cast<int64_t>(104ULL)],g->Regs[cast<int64_t>(109ULL)],get(Map<bool,std::string>{{true,std::string("Z",1)},{false,std::string("W",1)}},(cast<uint32_t>((g->Regs[cast<int64_t>(109ULL)] & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL))),n3ds_f24bits(g->Regs[cast<int64_t>(77ULL)]),n3ds_f24bits(g->Regs[cast<int64_t>(78ULL)]),g->Regs[cast<int64_t>(263ULL)]);
go_fmt_Printf(std::string("gpu draw %d: indexed=%v count=%d first=%d base=0x%08X bufs=%d prim=%d vp=%.1fx%.1f color=0x%08X depth=0x%08X dim=%dx%d test=%v wr=%v\012",133),g->Draws,indexed,count,first,base,len(bufs),cast<uint32_t>((shr<uint32_t>(g->Regs[cast<int64_t>(606ULL)],cast<int64_t>(8ULL)) & cast<uint32_t>(3ULL))),n3ds_f24bits(g->Regs[cast<int64_t>(65ULL)]),n3ds_f24bits(g->Regs[cast<int64_t>(67ULL)]),tfb.colorAddr,tfb.depthAddr,tfb.width,tfb.height,tfb.depthTest,tfb.depthWr);
if (g->TraceUniforms) {
{auto&& tmp144 = g->Float;
for(int64_t tmp145=0;tmp145<len(tmp144);++tmp145){
auto ci=tmp145;go_fmt_Printf(std::string("  c%-2d = %v\012",13),ci,g->Float[ci]);
}}
}
else {
{auto&& tmp146 = Slice<int64_t>{cast<int64_t>(0ULL),cast<int64_t>(1ULL),cast<int64_t>(2ULL),cast<int64_t>(3ULL),cast<int64_t>(4ULL),cast<int64_t>(5ULL),cast<int64_t>(6ULL),cast<int64_t>(32ULL),cast<int64_t>(33ULL),cast<int64_t>(34ULL),cast<int64_t>(35ULL),cast<int64_t>(64ULL),cast<int64_t>(65ULL)};
for(int64_t tmp147=0;tmp147<len(tmp146);++tmp147){
auto ci=tmp146[tmp147];go_fmt_Printf(std::string("  c%-2d = %v\012",13),ci,g->Float[ci]);
}}
}
Slice<int64_t> nan={};
{auto&& tmp148 = g->Float;
for(int64_t tmp149=0;tmp149<len(tmp148);++tmp149){
auto ci=tmp149;{int64_t k = cast<int64_t>(0ULL);for (;(k < cast<int64_t>(4ULL));k++){
if ((g->Float[ci][k] != g->Float[ci][k])) {
nan = append(nan,ci);
break;
}
}
}}}
go_fmt_Printf(std::string("  NaN uniforms at draw: %v; bool=%04X entry=0x%03X\012",51),nan,g->Bool,cast<uint32_t>((g->Regs[cast<int64_t>(698ULL)] & cast<uint32_t>(65535ULL))));
}
time_Time tv = n3ds_Machine_profStart(m);
if ((cast<uint32_t>(len(g->outs)) < count)) {
g->outs = Slice<n3ds_vsOut>::make(count);
}
Slice<n3ds_vsOut> outs = sub(g->outs,0,count);
int64_t entry = cast<int64_t>(cast<uint32_t>((g->Regs[cast<int64_t>(698ULL)] & cast<uint32_t>(4095ULL))));
int64_t maxIn = cast<int64_t>((cast<int64_t>(cast<uint32_t>((g->Regs[cast<int64_t>(697ULL)] & cast<uint32_t>(15ULL)))) + cast<int64_t>(1ULL)));
rrvertex::Draw vertexFast(g,indexed,physBase,fmtWord,fixedMask,bufs,comps,idxAddr,idx16,count,trace);
auto shadeVertex = [&](uint32_t i,std::array<std::array<float,4>,16>* vin,std::array<std::array<float,4>,16>* vout,std::array<std::array<float,4>,16>* attrs)->bool{
rrperf::Vertex vertexProfile;
uint32_t vi = cast<uint32_t>((first + i));
if (indexed && vertexFast.eligible) { vi=vertexFast.index(i); }
else if (indexed) {
if (idx16) {
vi = cast<uint32_t>((cast<uint32_t>(n3ds_Machine_Read(m,cast<uint32_t>((idxAddr + cast<uint32_t>((i * cast<uint32_t>(2ULL))))))) | shl<uint32_t>(cast<uint32_t>(n3ds_Machine_Read(m,cast<uint32_t>((cast<uint32_t>((idxAddr + cast<uint32_t>((i * cast<uint32_t>(2ULL))))) + cast<uint32_t>(1ULL))))),cast<int64_t>(8ULL))));
}
else {
vi = cast<uint32_t>(n3ds_Machine_Read(m,cast<uint32_t>((idxAddr + i))));
}
}
if(vertexFast.reuse(vi,i,outs)){vertexProfile.reuse(vi);return true;}
if(!vertexFast.fetch(vi,attrs)){
{int64_t a = cast<int64_t>(0ULL);for (;(a < cast<int64_t>(16ULL));a++){
(*attrs)[a] = std::array<float,4>{cast<float>(0.00000000000000000e+00),cast<float>(0.00000000000000000e+00),cast<float>(0.00000000000000000e+00),cast<float>(1.00000000000000000e+00)};
if ((cast<uint32_t>((shr<uint32_t>(fixedMask,cast<uint64_t>(a)) & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL))) {
(*attrs)[a] = g->fixedVal[a];
}
}
}{auto&& tmp150 = bufs;
for(int64_t tmp151=0;tmp151<len(tmp150);++tmp151){
auto b=tmp150[tmp151];uint32_t p = n3ds_Machine_gpuAddrToVirt(m,cast<uint32_t>((cast<uint32_t>((physBase + b.off)) + cast<uint32_t>((vi * b.stride)))));
{auto&& tmp152 = sub(comps,b.first,cast<int64_t>((b.first + b.n)));
for(int64_t tmp153=0;tmp153<len(tmp152);++tmp153){
auto c=tmp152[tmp153];if ((c >= cast<int64_t>(12ULL))) {
p += cast<uint32_t>((cast<uint32_t>(cast<int64_t>((c - cast<int64_t>(11ULL)))) * cast<uint32_t>(4ULL)));
continue;
}
uint64_t f = cast<uint64_t>((shr<uint64_t>(fmtWord,(cast<uint64_t>((cast<uint64_t>(4ULL) * cast<uint64_t>(c))))) & cast<uint64_t>(15ULL)));
int64_t n = cast<int64_t>((cast<int64_t>(shr<uint64_t>(f,cast<int64_t>(2ULL))) + cast<int64_t>(1ULL)));
std::array<float,4> val={};
val[cast<int64_t>(3ULL)] = cast<float>(1.00000000000000000e+00);
{int64_t j = cast<int64_t>(0ULL);for (;(j < n);j++){
{
switch(cast<uint64_t>((f & cast<uint64_t>(3ULL)))){
case cast<uint64_t>(0ULL):{
val[j] = cast<float>(cast<int8_t>(n3ds_Machine_Read(m,p)));
p++;
break;}
case cast<uint64_t>(1ULL):{
val[j] = cast<float>(n3ds_Machine_Read(m,p));
p++;
break;}
case cast<uint64_t>(2ULL):{
val[j] = cast<float>(cast<int16_t>(cast<uint16_t>((cast<uint16_t>(n3ds_Machine_Read(m,p)) | shl<uint16_t>(cast<uint16_t>(n3ds_Machine_Read(m,cast<uint32_t>((p + cast<uint32_t>(1ULL))))),cast<int64_t>(8ULL))))));
p += cast<uint32_t>(2ULL);
break;}
case cast<uint64_t>(3ULL):{
val[j] = n3ds_f32bits(n3ds_Machine_ReadWord(m,p));
p += cast<uint32_t>(4ULL);
break;}
}}
}
}(*attrs)[c] = val;
}}
}}
}
n3ds_mapAttrsToInputs(vin,attrs,inPerm,maxIn);
vertexProfile.shade(vi);
if ((!n3ds_GPU_shaderRun(g,vin,vout,entry))) {
return false;
}
vertexProfile.finish();
n3ds_GPU_mapOutputs(g,vout,(&outs[i]));
vertexFast.remember(vi,i);
if ((trace && (i < cast<uint32_t>(8ULL)))) {
n3ds_vsOut* r = (&outs[i]);
go_fmt_Printf(std::string("  v%-3d (i=%d) clip=%v\012",23),vi,i,r->pos);
{int64_t a = cast<int64_t>(0ULL);for (;(a < cast<int64_t>(12ULL));a++){
if (((cast<uint64_t>((shr<uint64_t>(fmtWord,(cast<uint64_t>((cast<uint64_t>(4ULL) * cast<uint64_t>(a))))) & cast<uint64_t>(15ULL))) != cast<uint64_t>(0ULL)) || (cast<uint32_t>((shr<uint32_t>(fixedMask,cast<uint64_t>(a)) & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL)))) {
go_fmt_Printf(std::string("       attr%-2d = %v\012",21),a,(*attrs)[a]);
}
}
}}
return true;
}
;
{
int64_t workers = n3ds_GPU_vertexWorkers(g,count);
if ((workers > cast<int64_t>(1ULL))) {
n3ds_GPU_decodeAll(g);
Slice<bool> failed = Slice<bool>::make(workers);
uint32_t per = divi<uint32_t>((cast<uint32_t>((cast<uint32_t>((count + cast<uint32_t>(workers))) - cast<uint32_t>(1ULL)))),cast<uint32_t>(workers));
n3ds_workPool_run(n3ds_GPU_pool(g),workers,[&](int64_t w)->void{
uint32_t lo = cast<uint32_t>((cast<uint32_t>(w) * per));
uint32_t hi = cast<uint32_t>((lo + per));
if ((hi > count)) {
hi = count;
}
std::array<std::array<float,4>,16> vin={};
std::array<std::array<float,4>,16> vout={};
std::array<std::array<float,4>,16> attrs={};
{uint32_t i = lo;for (;(i < hi);i++){
if ((!shadeVertex(i,(&vin),(&vout),(&attrs)))) {
failed[w] = true;
return ;
}
}
}}
);
{auto&& tmp154 = failed;
for(int64_t tmp155=0;tmp155<len(tmp154);++tmp155){
auto f=tmp154[tmp155];if (f) {
std::array<std::array<float,4>,16> vin={};
std::array<std::array<float,4>,16> vout={};
std::array<std::array<float,4>,16> attrs={};
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < count);i++){
if ((!shadeVertex(i,(&vin),(&vout),(&attrs)))) {
return ;
}
}
}return ;
}
}}
}
else {
std::array<std::array<float,4>,16> vin={};
std::array<std::array<float,4>,16> vout={};
std::array<std::array<float,4>,16> attrs={};
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < count);i++){
if ((!shadeVertex(i,(&vin),(&vout),(&attrs)))) {
return ;
}
}
}}
}
n3ds_Machine_profEnd(m,cast<int64_t>(1ULL),tv);
rrperf::Clip clipProfile;
time_Time tr = n3ds_Machine_profStart(m);
auto tmp156=defer([&](){n3ds_Machine_profEnd(m,cast<int64_t>(2ULL),tr);});
n3ds_fbState fb = n3ds_GPU_fbstate(g);
n3ds_lightState ls = n3ds_GPU_lightstate(g);
n3ds_tevState tev = n3ds_GPU_tevstate(g);
uint32_t prim = cast<uint32_t>((shr<uint32_t>(g->Regs[cast<int64_t>(606ULL)],cast<int64_t>(8ULL)) & cast<uint32_t>(3ULL)));
int64_t n = len(outs);
Slice<n3ds_rasterTri> tris = sub(g->tris,0,cast<int64_t>(0ULL));
auto emit = [&](int64_t a,int64_t b,int64_t c)->void{
tris = n3ds_GPU_clipTri(g,tris,outs[a],outs[b],outs[c],(&fb));
}
;
{
switch(prim){
case cast<uint32_t>(0ULL):case cast<uint32_t>(3ULL):{
{int64_t i = cast<int64_t>(0ULL);for (;(cast<int64_t>((i + cast<int64_t>(2ULL))) < n);i += cast<int64_t>(3ULL)){
emit(i,cast<int64_t>((i + cast<int64_t>(1ULL))),cast<int64_t>((i + cast<int64_t>(2ULL))));
}
}break;}
case cast<uint32_t>(1ULL):{
{int64_t i = cast<int64_t>(0ULL);for (;(cast<int64_t>((i + cast<int64_t>(2ULL))) < n);i++){
if ((cast<int64_t>((i & cast<int64_t>(1ULL))) == cast<int64_t>(0ULL))) {
emit(i,cast<int64_t>((i + cast<int64_t>(1ULL))),cast<int64_t>((i + cast<int64_t>(2ULL))));
}
else {
emit(cast<int64_t>((i + cast<int64_t>(1ULL))),i,cast<int64_t>((i + cast<int64_t>(2ULL))));
}
}
}break;}
case cast<uint32_t>(2ULL):{
{int64_t i = cast<int64_t>(1ULL);for (;(cast<int64_t>((i + cast<int64_t>(1ULL))) < n);i++){
emit(cast<int64_t>(0ULL),i,cast<int64_t>((i + cast<int64_t>(1ULL))));
}
}break;}
default:{
arm_CPU_Halt(m->CPU,std::string("gpu: primitive mode 3 (geometry) unimplemented",46));
return ;
break;}
}}
g->tris = tris;
clipProfile.finish();
n3ds_GPU_fill(g,(&fb),(&ls),(&tev),tris);
if (((fb.colorMask != cast<uint32_t>(0ULL)) || fb.shadowMode)) {
n3ds_GPU_invalidateTextures(g,fb.colorAddr,cast<uint32_t>((cast<uint32_t>(((((cast<uint32_t>((fb.width + cast<uint32_t>(7ULL)))) & ~(cast<uint32_t>(7ULL)))) * (((cast<uint32_t>((fb.height + cast<uint32_t>(7ULL)))) & ~(cast<uint32_t>(7ULL)))))) * cast<uint32_t>(4ULL))));
}
}
}
// tools/platform/n3ds/gpu_raster.go:379:1
void n3ds_GPU_fill(n3ds_GPU* g,n3ds_fbState* fb,n3ds_lightState* ls,n3ds_tevState* tv,Slice<n3ds_rasterTri> tris){
rrprof::Scope profile(2,"PICA software rasterizer");auto graphics=rrgpu::Operation::draw(g,fb,ls,tv,tris);if(graphics.execute())return;rrperf::Fallback fallbackProfile(g,fb,ls,tv,tris,graphics.target!=nullptr);{
if ((len(tris) == cast<int64_t>(0ULL))) {
return ;
}
int64_t workers = n3ds_GPU_rasterWorkers(g,fb,tris);
if ((workers <= cast<int64_t>(1ULL))) {
n3ds_rstats st={};
{auto&& tmp157 = tris;
for(int64_t tmp158=0;tmp158<len(tmp157);++tmp158){
auto i=tmp158;n3ds_GPU_fillTri(g,fb,ls,tv,(&tris[i]),tris[i].minY,tris[i].maxY,(&st));
}}
n3ds_GPU_mergeStats(g,(&st));
return ;
}
n3ds_GPU_warmTextures(g);
constexpr int64_t bandRows=8ULL;
int64_t height = cast<int64_t>(fb->height);
int64_t bands = divi<int64_t>((cast<int64_t>((cast<int64_t>((height + cast<int64_t>(8ULL))) - cast<int64_t>(1ULL)))),cast<int64_t>(8ULL));
Slice<n3ds_rstats> stats = Slice<n3ds_rstats>::make(workers);
int32_t next={};
n3ds_workPool_run(n3ds_GPU_pool(g),workers,[&](int64_t w)->void{
n3ds_rstats* st = (&stats[w]);
{;for (;;){
int64_t b = cast<int64_t>((cast<int64_t>(go_atomic_AddInt32((&next),cast<int32_t>(1ULL))) - cast<int64_t>(1ULL)));
if ((b >= bands)) {
return ;
}
int64_t yLo = cast<int64_t>((b * cast<int64_t>(8ULL)));
int64_t yHi = cast<int64_t>((yLo + cast<int64_t>(8ULL)));
if ((yHi > height)) {
yHi = height;
}
{auto&& tmp159 = tris;
for(int64_t tmp160=0;tmp160<len(tmp159);++tmp160){
auto i=tmp160;n3ds_rasterTri* t = (&tris[i]);
if (((t->maxY <= yLo) || (t->minY >= yHi))) {
continue;
}
auto tmp161 = std::make_tuple(t->minY,t->maxY);
int64_t lo = std::get<0>(tmp161);
int64_t hi = std::get<1>(tmp161);
if ((lo < yLo)) {
lo = yLo;
}
if ((hi > yHi)) {
hi = yHi;
}
n3ds_GPU_fillTri(g,fb,ls,tv,t,lo,hi,st);
}}
}
}}
);
{auto&& tmp162 = stats;
for(int64_t tmp163=0;tmp163<len(tmp162);++tmp163){
auto i=tmp163;n3ds_GPU_mergeStats(g,(&stats[i]));
}}
}
}
// tools/platform/n3ds/gpu_raster.go:446:1
int64_t n3ds_GPU_rasterWorkers(n3ds_GPU* g,n3ds_fbState* fb,Slice<n3ds_rasterTri> tris){
{
n3ds_Machine* m = g->m;
if (((((bool(m->OnPixel) || bool(m->OnRead)) || bool(m->OnWrite)) || m->HidTrace) || m->SingleThreaded)) {
return cast<int64_t>(1ULL);
}
constexpr int64_t minPixels=256ULL;
int64_t work = cast<int64_t>(0ULL);
{auto&& tmp164 = tris;
for(int64_t tmp165=0;tmp165<len(tmp164);++tmp165){
auto i=tmp165;n3ds_rasterTri* t = (&tris[i]);
work += cast<int64_t>(((cast<int64_t>((t->maxX - t->minX))) * (cast<int64_t>((t->maxY - t->minY)))));
if ((work >= cast<int64_t>(1024ULL))) {
break;
}
}}
if ((work < cast<int64_t>(256ULL))) {
return cast<int64_t>(1ULL);
}
int64_t n = n3ds_maxWorkers;
{
int64_t b = divi<int64_t>((cast<int64_t>((cast<int64_t>(fb->height) + cast<int64_t>(15ULL)))),cast<int64_t>(16ULL));
if ((n > b)) {
n = b;
}
}
if ((n < cast<int64_t>(1ULL))) {
n = cast<int64_t>(1ULL);
}
return n;
}
}
// tools/platform/n3ds/gpu_raster.go:479:1
void n3ds_GPU_warmTextures(n3ds_GPU* g){
{
{int64_t u = cast<int64_t>(0ULL);for (;(u < cast<int64_t>(3ULL));u++){
auto tmp166 = n3ds_texUnitRegs(u);
uint32_t dimR = std::get<0>(tmp166);
uint32_t addrR = std::get<2>(tmp166);
uint32_t typR = std::get<3>(tmp166);
uint32_t w = cast<uint32_t>((shr<uint32_t>(g->Regs[dimR],cast<int64_t>(16ULL)) & cast<uint32_t>(2047ULL)));
uint32_t h = cast<uint32_t>((g->Regs[dimR] & cast<uint32_t>(2047ULL)));
if (((w == cast<uint32_t>(0ULL)) || (h == cast<uint32_t>(0ULL)))) {
continue;
}
n3ds_GPU_texture(g,n3ds_Machine_gpuAddrToVirt(g->m,shl<uint32_t>(g->Regs[addrR],cast<int64_t>(3ULL))),cast<uint32_t>((g->Regs[typR] & cast<uint32_t>(15ULL))),w,h);
}
}}
}
// tools/platform/n3ds/gpu_raster.go:503:1
void n3ds_mapAttrsToInputs(std::array<std::array<float,4>,16>* v,std::array<std::array<float,4>,16>* attrs,uint64_t perm,int64_t nAttr){
{
{auto&& tmp167 = (*v);
for(int64_t tmp168=0;tmp168<len(tmp167);++tmp168){
auto j=tmp168;(*v)[j] = std::array<float,4>{cast<float>(0.00000000000000000e+00),cast<float>(0.00000000000000000e+00),cast<float>(0.00000000000000000e+00),cast<float>(1.00000000000000000e+00)};
}}
{int64_t a = cast<int64_t>(0ULL);for (;((a < nAttr) && (a < cast<int64_t>(16ULL)));a++){
(*v)[cast<uint64_t>((shr<uint64_t>(perm,(cast<uint64_t>((cast<uint64_t>(4ULL) * cast<uint64_t>(a))))) & cast<uint64_t>(15ULL)))] = (*attrs)[a];
}
}}
}
// tools/platform/n3ds/gpu_raster.go:515:1
void n3ds_GPU_mapOutputs(n3ds_GPU* g,std::array<std::array<float,4>,16>* o,n3ds_vsOut* dst){
{
n3ds_vsOut out={};
out.color = std::array<float,4>{cast<float>(1.00000000000000000e+00),cast<float>(1.00000000000000000e+00),cast<float>(1.00000000000000000e+00),cast<float>(1.00000000000000000e+00)};
int64_t total = cast<int64_t>(cast<uint32_t>((g->Regs[cast<int64_t>(79ULL)] & cast<uint32_t>(7ULL))));
{int64_t r = cast<int64_t>(0ULL);for (;(r < total);r++){
uint32_t sem = g->Regs[cast<uint32_t>((cast<uint32_t>(80ULL) + cast<uint32_t>(r)))];
{int64_t c = cast<int64_t>(0ULL);for (;(c < cast<int64_t>(4ULL));c++){
uint32_t s = cast<uint32_t>((shr<uint32_t>(sem,(cast<uint64_t>((cast<uint64_t>(8ULL) * cast<uint64_t>(c))))) & cast<uint32_t>(31ULL)));
float v = (*o)[r][c];
{
if ((s <= cast<uint32_t>(3ULL))){
out.pos[s] = v;
}
else if (((s >= cast<uint32_t>(8ULL)) && (s <= cast<uint32_t>(11ULL)))){
out.color[cast<uint32_t>((s - cast<uint32_t>(8ULL)))] = v;
}
else if (((s == cast<uint32_t>(12ULL)) || (s == cast<uint32_t>(13ULL)))){
out.uv[cast<int64_t>(0ULL)][cast<uint32_t>((s - cast<uint32_t>(12ULL)))] = v;
}
else if (((s == cast<uint32_t>(14ULL)) || (s == cast<uint32_t>(15ULL)))){
out.uv[cast<int64_t>(1ULL)][cast<uint32_t>((s - cast<uint32_t>(14ULL)))] = v;
}
else if (((s == cast<uint32_t>(22ULL)) || (s == cast<uint32_t>(23ULL)))){
out.uv[cast<int64_t>(2ULL)][cast<uint32_t>((s - cast<uint32_t>(22ULL)))] = v;
}
else if (((s >= cast<uint32_t>(4ULL)) && (s <= cast<uint32_t>(7ULL)))){
out.quat[cast<uint32_t>((s - cast<uint32_t>(4ULL)))] = v;
}
else if ((s == cast<uint32_t>(16ULL))){
out.uv0w = v;
}
else if (((s >= cast<uint32_t>(18ULL)) && (s <= cast<uint32_t>(20ULL)))){
out.view[cast<uint32_t>((s - cast<uint32_t>(18ULL)))] = v;
}
}
tmp169:;
}
}}
}(*dst) = out;
}
}
// tools/platform/n3ds/gpu_raster.go:573:1
n3ds_fbState n3ds_GPU_fbstate(n3ds_GPU* g){
{
uint32_t dim = g->Regs[cast<int64_t>(286ULL)];
uint32_t dcm = g->Regs[cast<int64_t>(263ULL)];
bool colorWr = (g->Regs[cast<int64_t>(275ULL)] != cast<uint32_t>(0ULL));
bool depthRd = (g->Regs[cast<int64_t>(276ULL)] != cast<uint32_t>(0ULL));
bool depthWr = (g->Regs[cast<int64_t>(277ULL)] != cast<uint32_t>(0ULL));
uint32_t cmask = cast<uint32_t>((shr<uint32_t>(dcm,cast<int64_t>(8ULL)) & cast<uint32_t>(15ULL)));
if ((!colorWr)) {
cmask = cast<uint32_t>(0ULL);
}
n3ds_fbState fb = n3ds_fbState{n3ds_Machine_gpuAddrToVirt(g->m,shl<uint32_t>(g->Regs[cast<int64_t>(285ULL)],cast<int64_t>(3ULL))),n3ds_Machine_gpuAddrToVirt(g->m,shl<uint32_t>(g->Regs[cast<int64_t>(284ULL)],cast<int64_t>(3ULL))),cast<uint32_t>((dim & cast<uint32_t>(2047ULL))),cast<uint32_t>(((cast<uint32_t>((shr<uint32_t>(dim,cast<int64_t>(12ULL)) & cast<uint32_t>(1023ULL)))) + cast<uint32_t>(1ULL))),((cast<uint32_t>((dcm & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL)) && depthRd),((cast<uint32_t>((shr<uint32_t>(dcm,cast<int64_t>(12ULL)) & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL)) && depthWr),cast<uint32_t>((shr<uint32_t>(dcm,cast<int64_t>(4ULL)) & cast<uint32_t>(7ULL))),cmask,n3ds_f24bits(g->Regs[cast<int64_t>(65ULL)]),n3ds_f24bits(g->Regs[cast<int64_t>(67ULL)]),n3ds_f24bits(g->Regs[cast<int64_t>(77ULL)]),n3ds_f24bits(g->Regs[cast<int64_t>(78ULL)]),(cast<uint32_t>((g->Regs[cast<int64_t>(109ULL)] & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL)),(cast<uint32_t>((g->Regs[cast<int64_t>(256ULL)] & cast<uint32_t>(3ULL))) == cast<uint32_t>(3ULL)),{},{},{},{}};
uint32_t size = cast<uint32_t>((cast<uint32_t>(((((cast<uint32_t>((fb.width + cast<uint32_t>(7ULL)))) & ~(cast<uint32_t>(7ULL)))) * (((cast<uint32_t>((fb.height + cast<uint32_t>(7ULL)))) & ~(cast<uint32_t>(7ULL)))))) * cast<uint32_t>(4ULL)));
auto tmp170 = n3ds_Machine_directRange(g->m,fb.colorAddr,size);
fb.colorBuf = std::get<0>(tmp170);
fb.colorOff = std::get<1>(tmp170);
auto tmp171 = n3ds_Machine_directRange(g->m,fb.depthAddr,size);
fb.depthBuf = std::get<0>(tmp171);
fb.depthOff32 = std::get<1>(tmp171);
return fb;
}
}
// tools/platform/n3ds/gpu_raster.go:625:1
void n3ds_GPU_shadowMapWrite(n3ds_GPU* g,n3ds_fbState* fb,uint32_t off,float depth,uint8_t density){
{
uint32_t p = cast<uint32_t>((fb->colorAddr + off));
Slice<uint8_t> dst={};
if (bool(fb->colorBuf)) {
dst = sub(fb->colorBuf,cast<uint32_t>((fb->colorOff + off)),len(fb->colorBuf));
}
uint32_t z = cast<uint32_t>(cast<float>((depth * cast<float>(1.67772150000000000e+07))));
uint32_t refZ={};
uint8_t refS={};
if (bool(dst)) {
refZ = cast<uint32_t>((cast<uint32_t>((shl<uint32_t>(cast<uint32_t>(dst[cast<int64_t>(0ULL)]),cast<int64_t>(16ULL)) | shl<uint32_t>(cast<uint32_t>(dst[cast<int64_t>(1ULL)]),cast<int64_t>(8ULL)))) | cast<uint32_t>(dst[cast<int64_t>(2ULL)])));
refS = dst[cast<int64_t>(3ULL)];
}
else {
refZ = cast<uint32_t>((cast<uint32_t>((shl<uint32_t>(cast<uint32_t>(n3ds_Machine_Read(g->m,p)),cast<int64_t>(16ULL)) | shl<uint32_t>(cast<uint32_t>(n3ds_Machine_Read(g->m,cast<uint32_t>((p + cast<uint32_t>(1ULL))))),cast<int64_t>(8ULL)))) | cast<uint32_t>(n3ds_Machine_Read(g->m,cast<uint32_t>((p + cast<uint32_t>(2ULL)))))));
refS = n3ds_Machine_Read(g->m,cast<uint32_t>((p + cast<uint32_t>(3ULL))));
}
if ((z >= refZ)) {
return ;
}
if ((density == cast<uint8_t>(0ULL))) {
if (bool(dst)) {
auto tmp172 = std::make_tuple(cast<uint8_t>(shr<uint32_t>(z,cast<int64_t>(16ULL))),cast<uint8_t>(shr<uint32_t>(z,cast<int64_t>(8ULL))),cast<uint8_t>(z));
dst[cast<int64_t>(0ULL)] = std::get<0>(tmp172);
dst[cast<int64_t>(1ULL)] = std::get<1>(tmp172);
dst[cast<int64_t>(2ULL)] = std::get<2>(tmp172);
return ;
}
n3ds_Machine_Write(g->m,p,cast<uint8_t>(shr<uint32_t>(z,cast<int64_t>(16ULL))));
n3ds_Machine_Write(g->m,cast<uint32_t>((p + cast<uint32_t>(1ULL))),cast<uint8_t>(shr<uint32_t>(z,cast<int64_t>(8ULL))));
n3ds_Machine_Write(g->m,cast<uint32_t>((p + cast<uint32_t>(2ULL))),cast<uint8_t>(z));
return ;
}
uint32_t cfg = g->Regs[cast<int64_t>(304ULL)];
auto tmp173 = std::make_tuple(n3ds_f16(cfg),n3ds_f16(shr<uint32_t>(cfg,cast<int64_t>(16ULL))));
float k = std::get<0>(tmp173);
float lin = std::get<1>(tmp173);
float d = cast<float>(density);
if ((refZ != cast<uint32_t>(0ULL))) {
{
float den = cast<float>((k + cast<float>((cast<float>((lin * cast<float>(z))) / cast<float>(refZ)))));
if ((den != cast<float>(0.00000000000000000e+00))) {
d = cast<float>((cast<float>(density) / den));
}
}
}
{
uint8_t s = cast<uint8_t>(n3ds_clampf(d,cast<float>(0.00000000000000000e+00),cast<float>(2.55000000000000000e+02)));
if ((s < refS)) {
if (bool(dst)) {
dst[cast<int64_t>(3ULL)] = s;
return ;
}
n3ds_Machine_Write(g->m,cast<uint32_t>((p + cast<uint32_t>(3ULL))),s);
}
}
}
}
// tools/platform/n3ds/gpu_raster.go:690:1
void n3ds_GPU_mergeStats(n3ds_GPU* g,n3ds_rstats* st){
{
g->PixelsDrawn += st->pixelsDrawn;
g->DepthKilled += st->depthKilled;
g->ShadowWrites += st->shadowWrites;
g->ShadowSamples += st->shadowSamples;
g->ShadowOccluded += st->shadowOccluded;
}
}
// tools/platform/n3ds/gpu_raster.go:698:1
float n3ds_edgeFn(float ax,float ay,float bx,float by,float px,float py){
{
return cast<float>((cast<float>(((cast<float>((bx - ax))) * (cast<float>((py - ay))))) - cast<float>(((cast<float>((by - ay))) * (cast<float>((px - ax)))))));
}
}
// tools/platform/n3ds/gpu_raster.go:705:1
std::tuple<n3ds_rasterTri,bool> n3ds_GPU_setupTri(n3ds_GPU* g,n3ds_vsOut* a,n3ds_vsOut* b,n3ds_vsOut* c,n3ds_fbState* fb){
{
n3ds_rasterTri t={};
if (((((((!((a->pos[cast<int64_t>(3ULL)] > cast<float>(0.00000000000000000e+00)))) || (!((b->pos[cast<int64_t>(3ULL)] > cast<float>(0.00000000000000000e+00))))) || (!((c->pos[cast<int64_t>(3ULL)] > cast<float>(0.00000000000000000e+00))))) || (a->pos[cast<int64_t>(0ULL)] != a->pos[cast<int64_t>(0ULL)])) || (b->pos[cast<int64_t>(0ULL)] != b->pos[cast<int64_t>(0ULL)])) || (c->pos[cast<int64_t>(0ULL)] != c->pos[cast<int64_t>(0ULL)]))) {
g->RejectedTris++;
return {t,false};
}
auto toScreen = [&](n3ds_vsOut* v)->n3ds_scrVert{
float iw = cast<float>((cast<float>(1.00000000000000000e+00) / v->pos[cast<int64_t>(3ULL)]));
return n3ds_scrVert{cast<float>(((cast<float>((cast<float>((v->pos[cast<int64_t>(0ULL)] * iw)) + cast<float>(1.00000000000000000e+00)))) * fb->vpHalfW)),cast<float>(((cast<float>((cast<float>((v->pos[cast<int64_t>(1ULL)] * iw)) + cast<float>(1.00000000000000000e+00)))) * fb->vpHalfH)),cast<float>((v->pos[cast<int64_t>(2ULL)] * iw)),iw,v->color,v->uv,v->uv0w,v->quat,v->view};
}
;
auto tmp174 = std::make_tuple(toScreen(a),toScreen(b),toScreen(c));
n3ds_scrVert v0 = std::get<0>(tmp174);
n3ds_scrVert v1 = std::get<1>(tmp174);
n3ds_scrVert v2 = std::get<2>(tmp174);
float area = n3ds_edgeFn(v0.x,v0.y,v1.x,v1.y,v2.x,v2.y);
if ((area == cast<float>(0.00000000000000000e+00))) {
g->ZeroAreaTris++;
return {t,false};
}
uint32_t cull = cast<uint32_t>((g->Regs[cast<int64_t>(64ULL)] & cast<uint32_t>(3ULL)));
bool ccw = (area > cast<float>(0.00000000000000000e+00));
if (((((cull == cast<uint32_t>(1ULL)) && ccw)) || (((cull == cast<uint32_t>(2ULL)) && (!ccw))))) {
g->CulledTris++;
return {t,false};
}
if ((area < cast<float>(0.00000000000000000e+00))) {
auto tmp175 = std::make_tuple(v2,v1);
v1 = std::get<0>(tmp175);
v2 = std::get<1>(tmp175);
area = cast<float>(-area);
}
int64_t minX = cast<int64_t>(n3ds_min3f(v0.x,v1.x,v2.x));
int64_t maxX = cast<int64_t>((cast<int64_t>(n3ds_max3f(v0.x,v1.x,v2.x)) + cast<int64_t>(1ULL)));
int64_t minY = cast<int64_t>(n3ds_min3f(v0.y,v1.y,v2.y));
int64_t maxY = cast<int64_t>((cast<int64_t>(n3ds_max3f(v0.y,v1.y,v2.y)) + cast<int64_t>(1ULL)));
if ((minX < cast<int64_t>(0ULL))) {
minX = cast<int64_t>(0ULL);
}
if ((minY < cast<int64_t>(0ULL))) {
minY = cast<int64_t>(0ULL);
}
if ((maxX > cast<int64_t>(fb->width))) {
maxX = cast<int64_t>(fb->width);
}
if ((maxY > cast<int64_t>(fb->height))) {
maxY = cast<int64_t>(fb->height);
}
return {n3ds_rasterTri{v0,v1,v2,area,minX,maxX,minY,maxY},true};
}
}
// tools/platform/n3ds/gpu_raster.go:773:1
void n3ds_GPU_fillTri(n3ds_GPU* g,n3ds_fbState* fb,n3ds_lightState* ls,n3ds_tevState* tv,n3ds_rasterTri* t,int64_t yLo,int64_t yHi,n3ds_rstats* st){
{
auto tmp176 = std::make_tuple((&t->v0),(&t->v1),(&t->v2));
n3ds_scrVert* v0 = std::get<0>(tmp176);
n3ds_scrVert* v1 = std::get<1>(tmp176);
n3ds_scrVert* v2 = std::get<2>(tmp176);
float area = t->area;
bool stencil = (((cast<uint32_t>((g->Regs[cast<int64_t>(261ULL)] & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL)) && (cast<uint32_t>((g->Regs[cast<int64_t>(278ULL)] & cast<uint32_t>(3ULL))) == cast<uint32_t>(3ULL))) && (!fb->shadowMode));
auto tmp177 = std::make_tuple(t->minX,t->maxX);
int64_t minX = std::get<0>(tmp177);
int64_t maxX = std::get<1>(tmp177);
auto tmp178 = std::make_tuple(yLo,yHi);
int64_t minY = std::get<0>(tmp178);
int64_t maxY = std::get<1>(tmp178);
auto edge = n3ds_edgeFn;
{int64_t y = minY;for (;(y < maxY);y++){
uint32_t ty = cast<uint32_t>(cast<int64_t>((cast<int64_t>((cast<int64_t>(fb->height) - cast<int64_t>(1ULL))) - y)));
{int64_t x = minX;for (;(x < maxX);x++){
uint32_t off = n3ds_tiledOffset(cast<uint32_t>(x),ty,fb->width);
auto tmp179 = std::make_tuple(cast<float>((cast<float>(x) + cast<float>(5.00000000000000000e-01))),cast<float>((cast<float>(y) + cast<float>(5.00000000000000000e-01))));
float px = std::get<0>(tmp179);
float py = std::get<1>(tmp179);
float w0 = edge(v1->x,v1->y,v2->x,v2->y,px,py);
float w1 = edge(v2->x,v2->y,v0->x,v0->y,px,py);
float w2 = edge(v0->x,v0->y,v1->x,v1->y,px,py);
if ((((w0 < cast<float>(0.00000000000000000e+00)) || (w1 < cast<float>(0.00000000000000000e+00))) || (w2 < cast<float>(0.00000000000000000e+00)))) {
continue;
}
auto tmp180 = std::make_tuple(cast<float>((w0 / area)),cast<float>((w1 / area)),cast<float>((w2 / area)));
float l0 = std::get<0>(tmp180);
float l1 = std::get<1>(tmp180);
float l2 = std::get<2>(tmp180);
float iw = cast<float>((cast<float>((cast<float>((l0 * v0->iw)) + cast<float>((l1 * v1->iw)))) + cast<float>((l2 * v2->iw))));
float z = cast<float>((cast<float>((cast<float>((l0 * v0->z)) + cast<float>((l1 * v1->z)))) + cast<float>((l2 * v2->z))));
float depth = cast<float>((cast<float>((z * fb->depthScale)) + fb->depthOff));
if (((!fb->depthZBuffer) && (iw != cast<float>(0.00000000000000000e+00)))) {
depth /= iw;
}
if ((depth < cast<float>(0.00000000000000000e+00))) {
depth = cast<float>(0.00000000000000000e+00);
}
if ((depth > cast<float>(1.00000000000000000e+00))) {
depth = cast<float>(1.00000000000000000e+00);
}
if ((((!stencil) && (!fb->shadowMode)) && (!n3ds_GPU_depthCompare(g,fb,off,depth)))) {
st->depthKilled++;
n3ds_GPU_pixelEvent(g,cast<uint32_t>(x),ty,n3ds_PixelEvent{{},true,{},{},{},{},{},{}});
continue;
}
auto pc = [&](float a0,float a1,float a2)->float{
return cast<float>(((cast<float>((cast<float>((cast<float>((cast<float>((l0 * a0)) * v0->iw)) + cast<float>((cast<float>((l1 * a1)) * v1->iw)))) + cast<float>((cast<float>((l2 * a2)) * v2->iw))))) / iw));
}
;
std::array<float,4> col={};
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(4ULL));i++){
col[i] = pc(v0->col[i],v1->col[i],v2->col[i]);
}
}std::array<std::array<float,2>,3> uv={};
{int64_t t = cast<int64_t>(0ULL);for (;(t < cast<int64_t>(3ULL));t++){
uv[t][cast<int64_t>(0ULL)] = pc(v0->uv[t][cast<int64_t>(0ULL)],v1->uv[t][cast<int64_t>(0ULL)],v2->uv[t][cast<int64_t>(0ULL)]);
uv[t][cast<int64_t>(1ULL)] = pc(v0->uv[t][cast<int64_t>(1ULL)],v1->uv[t][cast<int64_t>(1ULL)],v2->uv[t][cast<int64_t>(1ULL)]);
}
}float uv0w = pc(v0->uv0w,v1->uv0w,v2->uv0w);
std::array<float,4> q={};
std::array<float,3> vw={};
if (ls->enabled) {
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(4ULL));i++){
q[i] = pc(v0->quat[i],v1->quat[i],v2->quat[i]);
}
}{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(3ULL));i++){
vw[i] = pc(v0->view[i],v1->view[i],v2->view[i]);
}
}}
auto tmp181 = n3ds_GPU_fragment(g,col,uv,uv0w,ls,tv,q,vw,st);
uint8_t r8 = std::get<0>(tmp181);
uint8_t g8 = std::get<1>(tmp181);
uint8_t b8 = std::get<2>(tmp181);
uint8_t a8 = std::get<3>(tmp181);
bool discard = std::get<4>(tmp181);
bool ok = std::get<5>(tmp181);
if ((!ok)) {
return ;
}
if (fb->shadowMode) {
n3ds_GPU_shadowMapWrite(g,fb,off,depth,cast<uint8_t>(g8));
st->shadowWrites++;
continue;
}
if (discard) {
n3ds_GPU_pixelEvent(g,cast<uint32_t>(x),ty,n3ds_PixelEvent{{},{},true,{},r8,g8,b8,a8});
continue;
}
if (stencil) {
uint32_t reject = n3ds_GPU_stencilDepthTest(g,fb,off,depth);
if ((reject != cast<uint32_t>(0ULL))) {
if ((reject == cast<uint32_t>(2ULL))) {
st->depthKilled++;
}
n3ds_GPU_pixelEvent(g,cast<uint32_t>(x),ty,n3ds_PixelEvent{{},(reject == cast<uint32_t>(2ULL)),{},(reject == cast<uint32_t>(1ULL)),{},{},{},{}});
continue;
}
}
n3ds_GPU_depthWrite(g,fb,off,depth);
n3ds_GPU_writePixel(g,fb,cast<uint32_t>(x),ty,off,r8,g8,b8,a8);
st->pixelsDrawn++;
}
}}
}}
}
// tools/platform/n3ds/gpu_raster.go:940:1
uint32_t n3ds_tiledOffset(uint32_t x,uint32_t y,uint32_t width){
{
uint32_t tile = cast<uint32_t>((cast<uint32_t>(((divi<uint32_t>(y,cast<uint32_t>(8ULL))) * (divi<uint32_t>(width,cast<uint32_t>(8ULL))))) + divi<uint32_t>(x,cast<uint32_t>(8ULL))));
uint32_t mo = cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((x & cast<uint32_t>(1ULL))) | shl<uint32_t>(cast<uint32_t>((y & cast<uint32_t>(1ULL))),cast<int64_t>(1ULL)))) | shl<uint32_t>(cast<uint32_t>((x & cast<uint32_t>(2ULL))),cast<int64_t>(1ULL)))) | shl<uint32_t>(cast<uint32_t>((y & cast<uint32_t>(2ULL))),cast<int64_t>(2ULL)))) | shl<uint32_t>(cast<uint32_t>((x & cast<uint32_t>(4ULL))),cast<int64_t>(2ULL)))) | shl<uint32_t>(cast<uint32_t>((y & cast<uint32_t>(4ULL))),cast<int64_t>(3ULL))));
return cast<uint32_t>(((cast<uint32_t>((cast<uint32_t>((tile * cast<uint32_t>(64ULL))) + mo))) * cast<uint32_t>(4ULL)));
}
}
// tools/platform/n3ds/gpu_raster.go:953:1
bool n3ds_GPU_depthCompare(n3ds_GPU* g,n3ds_fbState* fb,uint32_t off,float depth){
{
if ((!fb->depthTest)) {
return true;
}
uint32_t old={};
if (bool(fb->depthBuf)) {
Slice<uint8_t> d = borrowSub(fb->depthBuf,cast<uint32_t>((fb->depthOff32 + off)),len(fb->depthBuf));
old = cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(d[cast<int64_t>(0ULL)]) | shl<uint32_t>(cast<uint32_t>(d[cast<int64_t>(1ULL)]),cast<int64_t>(8ULL)))) | shl<uint32_t>(cast<uint32_t>(d[cast<int64_t>(2ULL)]),cast<int64_t>(16ULL))));
}
else {
uint32_t p = cast<uint32_t>((fb->depthAddr + off));
old = cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(n3ds_Machine_Read(g->m,p)) | shl<uint32_t>(cast<uint32_t>(n3ds_Machine_Read(g->m,cast<uint32_t>((p + cast<uint32_t>(1ULL))))),cast<int64_t>(8ULL)))) | shl<uint32_t>(cast<uint32_t>(n3ds_Machine_Read(g->m,cast<uint32_t>((p + cast<uint32_t>(2ULL))))),cast<int64_t>(16ULL))));
}
uint32_t nv = cast<uint32_t>(cast<float>((depth * cast<float>(1.67772150000000000e+07))));
{
switch(fb->depthFunc){
case cast<uint32_t>(0ULL):{
return false;
break;}
case cast<uint32_t>(1ULL):{
return true;
break;}
case cast<uint32_t>(2ULL):{
return (nv == old);
break;}
case cast<uint32_t>(3ULL):{
return (nv != old);
break;}
case cast<uint32_t>(4ULL):{
return (nv < old);
break;}
case cast<uint32_t>(5ULL):{
return (nv <= old);
break;}
case cast<uint32_t>(6ULL):{
return (nv > old);
break;}
default:{
return (nv >= old);
break;}
}}
}
}
// tools/platform/n3ds/gpu_raster.go:986:1
void n3ds_GPU_depthWrite(n3ds_GPU* g,n3ds_fbState* fb,uint32_t off,float depth){
{
if (((!fb->depthTest) || (!fb->depthWr))) {
return ;
}
uint32_t nv = cast<uint32_t>(cast<float>((depth * cast<float>(1.67772150000000000e+07))));
if (bool(fb->depthBuf)) {
Slice<uint8_t> d = borrowSub(fb->depthBuf,cast<uint32_t>((fb->depthOff32 + off)),len(fb->depthBuf));
auto tmp182 = std::make_tuple(cast<uint8_t>(nv),cast<uint8_t>(shr<uint32_t>(nv,cast<int64_t>(8ULL))),cast<uint8_t>(shr<uint32_t>(nv,cast<int64_t>(16ULL))));
d[cast<int64_t>(0ULL)] = std::get<0>(tmp182);
d[cast<int64_t>(1ULL)] = std::get<1>(tmp182);
d[cast<int64_t>(2ULL)] = std::get<2>(tmp182);
return ;
}
uint32_t p = cast<uint32_t>((fb->depthAddr + off));
n3ds_Machine_Write(g->m,p,cast<uint8_t>(nv));
n3ds_Machine_Write(g->m,cast<uint32_t>((p + cast<uint32_t>(1ULL))),cast<uint8_t>(shr<uint32_t>(nv,cast<int64_t>(8ULL))));
n3ds_Machine_Write(g->m,cast<uint32_t>((p + cast<uint32_t>(2ULL))),cast<uint8_t>(shr<uint32_t>(nv,cast<int64_t>(16ULL))));
}
}
// tools/platform/n3ds/gpu_raster.go:1004:1
void n3ds_GPU_writePixel(n3ds_GPU* g,n3ds_fbState* fb,uint32_t x,uint32_t y,uint32_t off,uint8_t r,uint8_t gr,uint8_t b,uint8_t a){
{
Slice<uint8_t> dst={};
uint8_t da={};
uint8_t db={};
uint8_t dg={};
uint8_t dr={};
if (bool(fb->colorBuf)) {
dst = borrowSub(fb->colorBuf,cast<uint32_t>((fb->colorOff + off)),len(fb->colorBuf));
auto tmp183 = std::make_tuple(dst[cast<int64_t>(0ULL)],dst[cast<int64_t>(1ULL)],dst[cast<int64_t>(2ULL)],dst[cast<int64_t>(3ULL)]);
da = std::get<0>(tmp183);
db = std::get<1>(tmp183);
dg = std::get<2>(tmp183);
dr = std::get<3>(tmp183);
}
else {
uint32_t p = cast<uint32_t>((fb->colorAddr + off));
auto tmp184 = std::make_tuple(n3ds_Machine_Read(g->m,p),n3ds_Machine_Read(g->m,cast<uint32_t>((p + cast<uint32_t>(1ULL)))));
da = std::get<0>(tmp184);
db = std::get<1>(tmp184);
auto tmp185 = std::make_tuple(n3ds_Machine_Read(g->m,cast<uint32_t>((p + cast<uint32_t>(2ULL)))),n3ds_Machine_Read(g->m,cast<uint32_t>((p + cast<uint32_t>(3ULL)))));
dg = std::get<0>(tmp185);
dr = std::get<1>(tmp185);
}
auto tmp186 = n3ds_GPU_blend(g,r,gr,b,a,dr,dg,db,da);
r = std::get<0>(tmp186);
gr = std::get<1>(tmp186);
b = std::get<2>(tmp186);
a = std::get<3>(tmp186);
n3ds_GPU_pixelEvent(g,x,y,n3ds_PixelEvent{(fb->colorMask != cast<uint32_t>(0ULL)),{},{},{},r,gr,b,a});
if (bool(dst)) {
if ((cast<uint32_t>((fb->colorMask & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL))) {
dst[cast<int64_t>(3ULL)] = r;
}
if ((cast<uint32_t>((fb->colorMask & cast<uint32_t>(2ULL))) != cast<uint32_t>(0ULL))) {
dst[cast<int64_t>(2ULL)] = gr;
}
if ((cast<uint32_t>((fb->colorMask & cast<uint32_t>(4ULL))) != cast<uint32_t>(0ULL))) {
dst[cast<int64_t>(1ULL)] = b;
}
if ((cast<uint32_t>((fb->colorMask & cast<uint32_t>(8ULL))) != cast<uint32_t>(0ULL))) {
dst[cast<int64_t>(0ULL)] = a;
}
return ;
}
uint32_t p = cast<uint32_t>((fb->colorAddr + off));
if ((cast<uint32_t>((fb->colorMask & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL))) {
n3ds_Machine_Write(g->m,cast<uint32_t>((p + cast<uint32_t>(3ULL))),r);
}
if ((cast<uint32_t>((fb->colorMask & cast<uint32_t>(2ULL))) != cast<uint32_t>(0ULL))) {
n3ds_Machine_Write(g->m,cast<uint32_t>((p + cast<uint32_t>(2ULL))),gr);
}
if ((cast<uint32_t>((fb->colorMask & cast<uint32_t>(4ULL))) != cast<uint32_t>(0ULL))) {
n3ds_Machine_Write(g->m,cast<uint32_t>((p + cast<uint32_t>(1ULL))),b);
}
if ((cast<uint32_t>((fb->colorMask & cast<uint32_t>(8ULL))) != cast<uint32_t>(0ULL))) {
n3ds_Machine_Write(g->m,p,a);
}
}
}
// tools/platform/n3ds/gpu_raster.go:1050:1
float n3ds_min3f(float a,float b,float c){
{
if ((b < a)) {
a = b;
}
if ((c < a)) {
a = c;
}
return a;
}
}
// tools/platform/n3ds/gpu_raster.go:1060:1
float n3ds_max3f(float a,float b,float c){
{
if ((b > a)) {
a = b;
}
if ((c > a)) {
a = c;
}
return a;
}
}
// tools/platform/n3ds/gpu_raster.go:1082:1
int64_t n3ds_GPU_vertexWorkers(n3ds_GPU* g,uint32_t count){
{
constexpr int64_t minPerWorker=8ULL;
n3ds_Machine* m = g->m;
if (((((bool(m->OnRead) || bool(m->OnWrite)) || m->HidTrace) || (g->TraceDraws > cast<int64_t>(0ULL))) || m->SingleThreaded)) {
return cast<int64_t>(1ULL);
}
int64_t n = n3ds_maxWorkers;
{
int64_t w = divi<int64_t>(cast<int64_t>(count),cast<int64_t>(8ULL));
if ((w < n)) {
n = w;
}
}
if ((n < cast<int64_t>(1ULL))) {
n = cast<int64_t>(1ULL);
}
return n;
}
}
// tools/platform/n3ds/gpu_raster.go:1103:1
uint32_t n3ds_GPU_stencilDepthTest(n3ds_GPU* g,n3ds_fbState* fb,uint32_t off,float depth){
{
auto tmp187 = std::make_tuple(g->Regs[cast<int64_t>(261ULL)],g->Regs[cast<int64_t>(262ULL)]);
uint32_t cfg = std::get<0>(tmp187);
uint32_t ops = std::get<1>(tmp187);
uint8_t old={};
if (bool(fb->depthBuf)) {
old = fb->depthBuf[cast<uint32_t>((cast<uint32_t>((fb->depthOff32 + off)) + cast<uint32_t>(3ULL)))];
}
else {
old = n3ds_Machine_Read(g->m,cast<uint32_t>((cast<uint32_t>((fb->depthAddr + off)) + cast<uint32_t>(3ULL))));
}
auto tmp188 = std::make_tuple(cast<uint8_t>(shr<uint32_t>(cfg,cast<int64_t>(16ULL))),cast<uint8_t>(shr<uint32_t>(cfg,cast<int64_t>(24ULL))));
uint8_t ref = std::get<0>(tmp188);
uint8_t mask = std::get<1>(tmp188);
auto tmp189 = std::make_tuple(cast<uint8_t>((ref & mask)),cast<uint8_t>((old & mask)));
uint8_t a = std::get<0>(tmp189);
uint8_t b = std::get<1>(tmp189);
bool pass = false;
{
switch(cast<uint32_t>((shr<uint32_t>(cfg,cast<int64_t>(4ULL)) & cast<uint32_t>(7ULL)))){
case cast<uint32_t>(1ULL):{
pass = true;
break;}
case cast<uint32_t>(2ULL):{
pass = (a == b);
break;}
case cast<uint32_t>(3ULL):{
pass = (a != b);
break;}
case cast<uint32_t>(4ULL):{
pass = (a < b);
break;}
case cast<uint32_t>(5ULL):{
pass = (a <= b);
break;}
case cast<uint32_t>(6ULL):{
pass = (a > b);
break;}
case cast<uint32_t>(7ULL):{
pass = (a >= b);
break;}
}}
auto tmp190 = std::make_tuple(cast<uint32_t>(0ULL),cast<uint32_t>((shr<uint32_t>(ops,cast<int64_t>(8ULL)) & cast<uint32_t>(7ULL))));
uint32_t reject = std::get<0>(tmp190);
uint32_t op = std::get<1>(tmp190);
if ((!pass)) {
auto tmp191 = std::make_tuple(cast<uint32_t>(1ULL),cast<uint32_t>((ops & cast<uint32_t>(7ULL))));
reject = std::get<0>(tmp191);
op = std::get<1>(tmp191);
}
else if ((!n3ds_GPU_depthCompare(g,fb,off,depth))) {
auto tmp192 = std::make_tuple(cast<uint32_t>(2ULL),cast<uint32_t>((shr<uint32_t>(ops,cast<int64_t>(4ULL)) & cast<uint32_t>(7ULL))));
reject = std::get<0>(tmp192);
op = std::get<1>(tmp192);
}
if (((g->Regs[cast<int64_t>(277ULL)] != cast<uint32_t>(0ULL)) && (op != cast<uint32_t>(0ULL)))) {
uint8_t value = n3ds_stencilValue(op,old,ref);
uint8_t writeMask = cast<uint8_t>(shr<uint32_t>(cfg,cast<int64_t>(8ULL)));
value = cast<uint8_t>(((old & ~(writeMask)) | cast<uint8_t>((value & writeMask))));
if (bool(fb->depthBuf)) {
fb->depthBuf[cast<uint32_t>((cast<uint32_t>((fb->depthOff32 + off)) + cast<uint32_t>(3ULL)))] = value;
}
else {
n3ds_Machine_Write(g->m,cast<uint32_t>((cast<uint32_t>((fb->depthAddr + off)) + cast<uint32_t>(3ULL))),value);
}
}
return reject;
}
}
// tools/platform/n3ds/gpu_raster.go:1149:1
uint8_t n3ds_stencilValue(uint32_t op,uint8_t old,uint8_t ref){
{
{
switch(op){
case cast<uint32_t>(1ULL):{
return cast<uint8_t>(0ULL);
break;}
case cast<uint32_t>(2ULL):{
return ref;
break;}
case cast<uint32_t>(3ULL):{
if ((old < cast<uint8_t>(255ULL))) {
return cast<uint8_t>((old + cast<uint8_t>(1ULL)));
}
break;}
case cast<uint32_t>(4ULL):{
if ((old > cast<uint8_t>(0ULL))) {
return cast<uint8_t>((old - cast<uint8_t>(1ULL)));
}
break;}
case cast<uint32_t>(5ULL):{
return cast<uint8_t>(~old);
break;}
case cast<uint32_t>(6ULL):{
return cast<uint8_t>((old + cast<uint8_t>(1ULL)));
break;}
case cast<uint32_t>(7ULL):{
return cast<uint8_t>((old - cast<uint8_t>(1ULL)));
break;}
}}
return old;
}
}
// tools/platform/n3ds/gpu_raster.go:1176:1
float n3ds_clipDistance(n3ds_vsOut v,int64_t plane){
{
{
switch(plane){
case cast<int64_t>(0ULL):{
return cast<float>((v.pos[cast<int64_t>(3ULL)] + v.pos[cast<int64_t>(0ULL)]));
break;}
case cast<int64_t>(1ULL):{
return cast<float>((v.pos[cast<int64_t>(3ULL)] - v.pos[cast<int64_t>(0ULL)]));
break;}
case cast<int64_t>(2ULL):{
return cast<float>((v.pos[cast<int64_t>(3ULL)] + v.pos[cast<int64_t>(1ULL)]));
break;}
case cast<int64_t>(3ULL):{
return cast<float>((v.pos[cast<int64_t>(3ULL)] - v.pos[cast<int64_t>(1ULL)]));
break;}
case cast<int64_t>(4ULL):{
return cast<float>((v.pos[cast<int64_t>(3ULL)] + v.pos[cast<int64_t>(2ULL)]));
break;}
case cast<int64_t>(5ULL):{
return cast<float>(-v.pos[cast<int64_t>(2ULL)]);
break;}
default:{
return cast<float>((v.pos[cast<int64_t>(3ULL)] - cast<float>(9.99999974737875164e-06)));
break;}
}}
}
}
// tools/platform/n3ds/gpu_raster.go:1195:1
n3ds_vsOut n3ds_clipVertex(n3ds_vsOut a,n3ds_vsOut b,float t){
{
n3ds_vsOut v={};
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(4ULL));i++){
v.pos[i] = cast<float>((a.pos[i] + cast<float>(((cast<float>((b.pos[i] - a.pos[i]))) * t))));
v.color[i] = cast<float>((a.color[i] + cast<float>(((cast<float>((b.color[i] - a.color[i]))) * t))));
v.quat[i] = cast<float>((a.quat[i] + cast<float>(((cast<float>((b.quat[i] - a.quat[i]))) * t))));
}
}{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(3ULL));i++){
v.view[i] = cast<float>((a.view[i] + cast<float>(((cast<float>((b.view[i] - a.view[i]))) * t))));
{int64_t j = cast<int64_t>(0ULL);for (;(j < cast<int64_t>(2ULL));j++){
v.uv[i][j] = cast<float>((a.uv[i][j] + cast<float>(((cast<float>((b.uv[i][j] - a.uv[i][j]))) * t))));
}
}}
}v.uv0w = cast<float>((a.uv0w + cast<float>(((cast<float>((b.uv0w - a.uv0w))) * t))));
return v;
}
}
// tools/platform/n3ds/gpu_raster.go:1212:1
Slice<n3ds_rasterTri> n3ds_GPU_clipTri(n3ds_GPU* g,Slice<n3ds_rasterTri> tris,n3ds_vsOut a,n3ds_vsOut b,n3ds_vsOut c,n3ds_fbState* fb){
{
std::array<n3ds_vsOut,3> vertices = std::array<n3ds_vsOut,3>{a,b,c};
auto tmp193 = std::make_tuple(cast<uint32_t>(0ULL),cast<uint32_t>(127ULL));
uint32_t outside = std::get<0>(tmp193);
uint32_t common = std::get<1>(tmp193);
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(3ULL));i++){
{auto&& tmp194 = vertices[i].pos;
for(int64_t tmp195=0;tmp195<len(tmp194);++tmp195){
auto p=tmp194[tmp195];if ((cast<float>((p - p)) != cast<float>(0.00000000000000000e+00))) {
g->RejectedTris++;
return tris;
}
}}
uint32_t mask = cast<uint32_t>(0ULL);
{int64_t plane = cast<int64_t>(0ULL);for (;(plane < cast<int64_t>(7ULL));plane++){
if ((n3ds_clipDistance(vertices[i],plane) < cast<float>(0.00000000000000000e+00))) {
mask |= shl<uint32_t>(cast<uint32_t>(1ULL),cast<uint64_t>(plane));
}
}
}outside |= mask;
common &= mask;
}
}if ((common != cast<uint32_t>(0ULL))) {
g->RejectedTris++;
return tris;
}
if ((outside == cast<uint32_t>(0ULL))) {
{
auto tmp196 = n3ds_GPU_setupTri(g,(&a),(&b),(&c),fb);
n3ds_rasterTri t = std::get<0>(tmp196);
bool ok = std::get<1>(tmp196);
if (ok) {
tris = append(tris,t);
}
}
return tris;
}
std::array<n3ds_vsOut,12> input = std::array<n3ds_vsOut,12>{a,b,c};
std::array<n3ds_vsOut,12> output={};
int64_t count = cast<int64_t>(3ULL);
{int64_t plane = cast<int64_t>(0ULL);for (;(plane < cast<int64_t>(7ULL));plane++){
if ((cast<uint32_t>((outside & (shl<uint32_t>(cast<uint32_t>(1ULL),cast<uint64_t>(plane))))) == cast<uint32_t>(0ULL))) {
continue;
}
int64_t n = cast<int64_t>(0ULL);
n3ds_vsOut prev = input[cast<int64_t>((count - cast<int64_t>(1ULL)))];
float pd = n3ds_clipDistance(prev,plane);
{int64_t i = cast<int64_t>(0ULL);for (;(i < count);i++){
n3ds_vsOut curr = input[i];
float cd = n3ds_clipDistance(curr,plane);
if ((((pd >= cast<float>(0.00000000000000000e+00))) != ((cd >= cast<float>(0.00000000000000000e+00))))) {
output[n] = n3ds_clipVertex(prev,curr,cast<float>((pd / (cast<float>((pd - cd))))));
n++;
}
if ((cd >= cast<float>(0.00000000000000000e+00))) {
output[n] = curr;
n++;
}
auto tmp197 = std::make_tuple(curr,cd);
prev = std::get<0>(tmp197);
pd = std::get<1>(tmp197);
}
}if ((n < cast<int64_t>(3ULL))) {
g->RejectedTris++;
return tris;
}
count = n;
{int64_t i = cast<int64_t>(0ULL);for (;(i < count);i++){
input[i] = output[i];
}
}}
}{int64_t i = cast<int64_t>(1ULL);for (;(cast<int64_t>((i + cast<int64_t>(1ULL))) < count);i++){
{
auto tmp198 = n3ds_GPU_setupTri(g,(&input[cast<int64_t>(0ULL)]),(&input[i]),(&input[cast<int64_t>((i + cast<int64_t>(1ULL)))]),fb);
n3ds_rasterTri t = std::get<0>(tmp198);
bool ok = std::get<1>(tmp198);
if (ok) {
tris = append(tris,t);
}
}
}
}return tris;
}
}
// tools/platform/n3ds/gpu_shader.go:32:1
bool n3ds_GPU_shaderRun(n3ds_GPU* g,std::array<std::array<float,4>,16>* v,std::array<std::array<float,4>,16>* out,int64_t entry){
{
(*out) = std::array<std::array<float,4>,16>{};
n3ds_shaderState s = n3ds_shaderState{g,v,out,{},{},{},{}};
return n3ds_shaderState_exec(&(s),entry,cast<int64_t>(4096ULL));
}
}
// tools/platform/n3ds/gpu_shader.go:52:1
bool n3ds_shaderState_exec(n3ds_shaderState* s,int64_t pc,int64_t end){
{
n3ds_GPU* g = s->g;
{int64_t steps = cast<int64_t>(0ULL);for (;(pc < end);steps++){
if ((steps > cast<int64_t>(1048576ULL))) {
arm_CPU_Halt(g->m->CPU,std::string("gpu shader: runaway program (no END after 1M steps)",51));
return false;
}
n3ds_shInst* d = n3ds_GPU_shaderInst(g,pc);
{
switch(d->kind){
case cast<uint8_t>(1ULL):{
pc++;
continue;
break;}
case cast<uint8_t>(2ULL):{
return true;
break;}
case cast<uint8_t>(4ULL):case cast<uint8_t>(5ULL):case cast<uint8_t>(6ULL):{
if ((!n3ds_shaderState_arith(s,d))) {
return false;
}
pc++;
continue;
break;}
}}
uint32_t in = g->Code[pc];
uint32_t op = shr<uint32_t>(in,cast<int64_t>(26ULL));
{
switch(op){
case cast<uint32_t>(36ULL):{
auto tmp199 = std::make_tuple(cast<int64_t>(cast<uint32_t>((shr<uint32_t>(in,cast<int64_t>(10ULL)) & cast<uint32_t>(4095ULL)))),cast<int64_t>(cast<uint32_t>((in & cast<uint32_t>(255ULL)))));
int64_t dst = std::get<0>(tmp199);
int64_t num = std::get<1>(tmp199);
if ((!n3ds_shaderState_exec(s,dst,cast<int64_t>((dst + num))))) {
return false;
}
pc++;
break;}
case cast<uint32_t>(37ULL):case cast<uint32_t>(38ULL):{
auto tmp200 = std::make_tuple(cast<int64_t>(cast<uint32_t>((shr<uint32_t>(in,cast<int64_t>(10ULL)) & cast<uint32_t>(4095ULL)))),cast<int64_t>(cast<uint32_t>((in & cast<uint32_t>(255ULL)))));
int64_t dst = std::get<0>(tmp200);
int64_t num = std::get<1>(tmp200);
bool taken = false;
if ((op == cast<uint32_t>(38ULL))) {
taken = n3ds_shaderState_boolReg(s,in);
}
else {
taken = n3ds_shaderState_cond(s,in);
}
if (taken) {
if ((!n3ds_shaderState_exec(s,dst,cast<int64_t>((dst + num))))) {
return false;
}
}
pc++;
break;}
case cast<uint32_t>(39ULL):case cast<uint32_t>(40ULL):{
auto tmp201 = std::make_tuple(cast<int64_t>(cast<uint32_t>((shr<uint32_t>(in,cast<int64_t>(10ULL)) & cast<uint32_t>(4095ULL)))),cast<int64_t>(cast<uint32_t>((in & cast<uint32_t>(255ULL)))));
int64_t dst = std::get<0>(tmp201);
int64_t num = std::get<1>(tmp201);
bool taken = false;
if ((op == cast<uint32_t>(39ULL))) {
taken = n3ds_shaderState_boolReg(s,in);
}
else {
taken = n3ds_shaderState_cond(s,in);
}
if (taken) {
if ((!n3ds_shaderState_exec(s,cast<int64_t>((pc + cast<int64_t>(1ULL))),dst))) {
return false;
}
}
else if ((num > cast<int64_t>(0ULL))) {
if ((!n3ds_shaderState_exec(s,dst,cast<int64_t>((dst + num))))) {
return false;
}
}
pc = cast<int64_t>((dst + num));
break;}
case cast<uint32_t>(41ULL):{
int64_t dst = cast<int64_t>(cast<uint32_t>((shr<uint32_t>(in,cast<int64_t>(10ULL)) & cast<uint32_t>(4095ULL))));
std::array<uint8_t,4> ir = g->Int[cast<uint32_t>((shr<uint32_t>(in,cast<int64_t>(22ULL)) & cast<uint32_t>(3ULL)))];
s->aL = cast<int32_t>(ir[cast<int64_t>(1ULL)]);
{int64_t n = cast<int64_t>(0ULL);for (;(n <= cast<int64_t>(ir[cast<int64_t>(0ULL)]));n++){
if ((!n3ds_shaderState_exec(s,cast<int64_t>((pc + cast<int64_t>(1ULL))),cast<int64_t>((dst + cast<int64_t>(1ULL)))))) {
return false;
}
s->aL += cast<int32_t>(ir[cast<int64_t>(2ULL)]);
}
}pc = cast<int64_t>((dst + cast<int64_t>(1ULL)));
break;}
case cast<uint32_t>(44ULL):case cast<uint32_t>(45ULL):{
int64_t dst = cast<int64_t>(cast<uint32_t>((shr<uint32_t>(in,cast<int64_t>(10ULL)) & cast<uint32_t>(4095ULL))));
bool taken = false;
if ((op == cast<uint32_t>(45ULL))) {
taken = n3ds_shaderState_boolReg(s,in);
if ((cast<uint32_t>((in & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL))) {
taken = (!taken);
}
}
else {
taken = n3ds_shaderState_cond(s,in);
}
if (taken) {
pc = dst;
}
else {
pc++;
}
break;}
default:{
arm_CPU_Halt(g->m->CPU,std::string("gpu shader: opcode 0x%02X unimplemented (instr 0x%08X)",54),op,in);
return false;
break;}
}}
}
}return true;
}
}
// tools/platform/n3ds/gpu_shader.go:155:1
bool n3ds_shaderState_cond(n3ds_shaderState* s,uint32_t in){
{
bool refx = (cast<uint32_t>((shr<uint32_t>(in,cast<int64_t>(25ULL)) & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL));
bool refy = (cast<uint32_t>((shr<uint32_t>(in,cast<int64_t>(24ULL)) & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL));
bool x = (s->cc[cast<int64_t>(0ULL)] == refx);
bool y = (s->cc[cast<int64_t>(1ULL)] == refy);
{
switch(cast<uint32_t>((shr<uint32_t>(in,cast<int64_t>(22ULL)) & cast<uint32_t>(3ULL)))){
case cast<uint32_t>(0ULL):{
return (x || y);
break;}
case cast<uint32_t>(1ULL):{
return (x && y);
break;}
case cast<uint32_t>(2ULL):{
return x;
break;}
default:{
return y;
break;}
}}
}
}
// tools/platform/n3ds/gpu_shader.go:172:1
bool n3ds_shaderState_boolReg(n3ds_shaderState* s,uint32_t in){
{
return (cast<uint32_t>((shr<uint32_t>(s->g->Bool,(cast<uint32_t>((shr<uint32_t>(in,cast<int64_t>(22ULL)) & cast<uint32_t>(15ULL))))) & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL));
}
}
// tools/platform/n3ds/gpu_shader.go:179:1
bool n3ds_shaderState_arith(n3ds_shaderState* s,n3ds_shInst* d){
{
uint32_t op = cast<uint32_t>(d->op);
{
switch(d->kind){
case cast<uint8_t>(5ULL):{
auto tmp202 = std::make_tuple(n3ds_shaderState_src(s,(&d->src[cast<int64_t>(0ULL)])),n3ds_shaderState_src(s,(&d->src[cast<int64_t>(1ULL)])),n3ds_shaderState_src(s,(&d->src[cast<int64_t>(2ULL)])));
std::array<float,4> s1 = std::get<0>(tmp202);
std::array<float,4> s2 = std::get<1>(tmp202);
std::array<float,4> s3 = std::get<2>(tmp202);
std::array<float,4> out={};
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(4ULL));i++){
out[i] = cast<float>((cast<float>((s1[i] * s2[i])) + s3[i]));
}
}n3ds_shaderState_writeDst(s,d,out);
return true;
break;}
case cast<uint8_t>(6ULL):{
auto tmp203 = std::make_tuple(n3ds_shaderState_src(s,(&d->src[cast<int64_t>(0ULL)])),n3ds_shaderState_src(s,(&d->src[cast<int64_t>(1ULL)])));
std::array<float,4> s1 = std::get<0>(tmp203);
std::array<float,4> s2 = std::get<1>(tmp203);
s->cc[cast<int64_t>(0ULL)] = n3ds_compare(cast<uint32_t>(d->cmpX),s1[cast<int64_t>(0ULL)],s2[cast<int64_t>(0ULL)]);
s->cc[cast<int64_t>(1ULL)] = n3ds_compare(cast<uint32_t>(d->cmpY),s1[cast<int64_t>(1ULL)],s2[cast<int64_t>(1ULL)]);
return true;
break;}
}}
auto tmp204 = std::make_tuple(n3ds_shaderState_src(s,(&d->src[cast<int64_t>(0ULL)])),n3ds_shaderState_src(s,(&d->src[cast<int64_t>(1ULL)])));
std::array<float,4> s1 = std::get<0>(tmp204);
std::array<float,4> s2 = std::get<1>(tmp204);
std::array<float,4> out={};
{
switch(op){
case cast<uint32_t>(0ULL):{
{auto&& tmp205 = out;
for(int64_t tmp206=0;tmp206<len(tmp205);++tmp206){
auto i=tmp206;out[i] = cast<float>((s1[i] + s2[i]));
}}
break;}
case cast<uint32_t>(1ULL):{
float dp = cast<float>((cast<float>((cast<float>((s1[cast<int64_t>(0ULL)] * s2[cast<int64_t>(0ULL)])) + cast<float>((s1[cast<int64_t>(1ULL)] * s2[cast<int64_t>(1ULL)])))) + cast<float>((s1[cast<int64_t>(2ULL)] * s2[cast<int64_t>(2ULL)]))));
out = std::array<float,4>{dp,dp,dp,dp};
break;}
case cast<uint32_t>(2ULL):{
float dp = cast<float>((cast<float>((cast<float>((cast<float>((s1[cast<int64_t>(0ULL)] * s2[cast<int64_t>(0ULL)])) + cast<float>((s1[cast<int64_t>(1ULL)] * s2[cast<int64_t>(1ULL)])))) + cast<float>((s1[cast<int64_t>(2ULL)] * s2[cast<int64_t>(2ULL)])))) + cast<float>((s1[cast<int64_t>(3ULL)] * s2[cast<int64_t>(3ULL)]))));
out = std::array<float,4>{dp,dp,dp,dp};
break;}
case cast<uint32_t>(3ULL):case cast<uint32_t>(24ULL):{
float dp = cast<float>((cast<float>((cast<float>((cast<float>((s1[cast<int64_t>(0ULL)] * s2[cast<int64_t>(0ULL)])) + cast<float>((s1[cast<int64_t>(1ULL)] * s2[cast<int64_t>(1ULL)])))) + cast<float>((s1[cast<int64_t>(2ULL)] * s2[cast<int64_t>(2ULL)])))) + s2[cast<int64_t>(3ULL)]));
out = std::array<float,4>{dp,dp,dp,dp};
break;}
case cast<uint32_t>(8ULL):{
{auto&& tmp207 = out;
for(int64_t tmp208=0;tmp208<len(tmp207);++tmp208){
auto i=tmp208;out[i] = cast<float>((s1[i] * s2[i]));
}}
break;}
case cast<uint32_t>(9ULL):case cast<uint32_t>(26ULL):{
{auto&& tmp209 = out;
for(int64_t tmp210=0;tmp210<len(tmp209);++tmp210){
auto i=tmp210;if ((s1[i] >= s2[i])) {
out[i] = cast<float>(1.00000000000000000e+00);
}
}}
break;}
case cast<uint32_t>(10ULL):case cast<uint32_t>(27ULL):{
{auto&& tmp211 = out;
for(int64_t tmp212=0;tmp212<len(tmp211);++tmp212){
auto i=tmp212;if ((s1[i] < s2[i])) {
out[i] = cast<float>(1.00000000000000000e+00);
}
}}
break;}
case cast<uint32_t>(11ULL):{
{auto&& tmp213 = out;
for(int64_t tmp214=0;tmp214<len(tmp213);++tmp214){
auto i=tmp214;out[i] = n3ds_floor32(s1[i]);
}}
break;}
case cast<uint32_t>(12ULL):{
{auto&& tmp215 = out;
for(int64_t tmp216=0;tmp216<len(tmp215);++tmp216){
auto i=tmp216;if ((s1[i] > s2[i])) {
out[i] = s1[i];
}
else {
out[i] = s2[i];
}
}}
break;}
case cast<uint32_t>(13ULL):{
{auto&& tmp217 = out;
for(int64_t tmp218=0;tmp218<len(tmp217);++tmp218){
auto i=tmp218;if ((s1[i] < s2[i])) {
out[i] = s1[i];
}
else {
out[i] = s2[i];
}
}}
break;}
case cast<uint32_t>(14ULL):{
float d = cast<float>((cast<float>(1.00000000000000000e+00) / s1[cast<int64_t>(0ULL)]));
out = std::array<float,4>{d,d,d,d};
break;}
case cast<uint32_t>(15ULL):{
float d = n3ds_rsqrt32(s1[cast<int64_t>(0ULL)]);
out = std::array<float,4>{d,d,d,d};
break;}
case cast<uint32_t>(18ULL):{
if (d->mask[cast<int64_t>(0ULL)]) {
s->a0[cast<int64_t>(0ULL)] = cast<int32_t>(s1[cast<int64_t>(0ULL)]);
}
if (d->mask[cast<int64_t>(1ULL)]) {
s->a0[cast<int64_t>(1ULL)] = cast<int32_t>(s1[cast<int64_t>(1ULL)]);
}
return true;
break;}
case cast<uint32_t>(19ULL):{
out = s1;
break;}
default:{
arm_CPU_Halt(s->g->m->CPU,std::string("gpu shader: opcode 0x%02X unimplemented",39),op);
return false;
break;}
}}
n3ds_shaderState_writeDst(s,d,out);
return true;
}
}
// tools/platform/n3ds/gpu_shader.go:278:1
std::array<float,4> n3ds_shaderState_src(n3ds_shaderState* s,n3ds_shSrc* o){
{
std::array<float,4>* base={};
{
switch(o->bank){
case cast<uint8_t>(0ULL):{
base = (&(*s->v)[o->reg]);
break;}
case cast<uint8_t>(1ULL):{
base = (&s->r[o->reg]);
break;}
default:{
int64_t c = cast<int64_t>(o->reg);
{
switch(o->idx){
case cast<uint8_t>(1ULL):{
c += cast<int64_t>(s->a0[cast<int64_t>(0ULL)]);
break;}
case cast<uint8_t>(2ULL):{
c += cast<int64_t>(s->a0[cast<int64_t>(1ULL)]);
break;}
case cast<uint8_t>(3ULL):{
c += cast<int64_t>(s->aL);
break;}
}}
if (((c < cast<int64_t>(0ULL)) || (c >= cast<int64_t>(96ULL)))) {
return std::array<float,4>{};
}
base = (&s->g->Float[c]);
break;}
}}
if (o->plain) {
return (*base);
}
std::array<float,4> out = std::array<float,4>{(*base)[o->sw[cast<int64_t>(0ULL)]],(*base)[o->sw[cast<int64_t>(1ULL)]],(*base)[o->sw[cast<int64_t>(2ULL)]],(*base)[o->sw[cast<int64_t>(3ULL)]]};
if (o->neg) {
auto tmp219 = std::make_tuple(cast<float>(-out[cast<int64_t>(0ULL)]),cast<float>(-out[cast<int64_t>(1ULL)]),cast<float>(-out[cast<int64_t>(2ULL)]),cast<float>(-out[cast<int64_t>(3ULL)]));
out[cast<int64_t>(0ULL)] = std::get<0>(tmp219);
out[cast<int64_t>(1ULL)] = std::get<1>(tmp219);
out[cast<int64_t>(2ULL)] = std::get<2>(tmp219);
out[cast<int64_t>(3ULL)] = std::get<3>(tmp219);
}
return out;
}
}
// tools/platform/n3ds/gpu_shader.go:311:1
void n3ds_shaderState_writeDst(n3ds_shaderState* s,n3ds_shInst* d,std::array<float,4> val){
{
std::array<float,4>* t={};
if (d->dstTmp) {
t = (&s->r[d->dst]);
}
else {
t = (&(*s->o)[d->dst]);
}
if (d->maskAll) {
(*t) = val;
return ;
}
if (d->mask[cast<int64_t>(0ULL)]) {
(*t)[cast<int64_t>(0ULL)] = val[cast<int64_t>(0ULL)];
}
if (d->mask[cast<int64_t>(1ULL)]) {
(*t)[cast<int64_t>(1ULL)] = val[cast<int64_t>(1ULL)];
}
if (d->mask[cast<int64_t>(2ULL)]) {
(*t)[cast<int64_t>(2ULL)] = val[cast<int64_t>(2ULL)];
}
if (d->mask[cast<int64_t>(3ULL)]) {
(*t)[cast<int64_t>(3ULL)] = val[cast<int64_t>(3ULL)];
}
}
}
// tools/platform/n3ds/gpu_shader.go:336:1
bool n3ds_compare(uint32_t op,float a,float b){
{
{
switch(op){
case cast<uint32_t>(0ULL):{
return (a == b);
break;}
case cast<uint32_t>(1ULL):{
return (a != b);
break;}
case cast<uint32_t>(2ULL):{
return (a < b);
break;}
case cast<uint32_t>(3ULL):{
return (a <= b);
break;}
case cast<uint32_t>(4ULL):{
return (a > b);
break;}
case cast<uint32_t>(5ULL):{
return (a >= b);
break;}
default:{
return true;
break;}
}}
}
}
// tools/platform/n3ds/gpu_shader.go:355:1
float n3ds_floor32(float f){
{
float i = cast<float>(cast<int32_t>(f));
if (((f < cast<float>(0.00000000000000000e+00)) && (f != i))) {
i--;
}
return i;
}
}
// tools/platform/n3ds/gpu_shader.go:363:1
float n3ds_rsqrt32(float f){
{
if ((f <= cast<float>(0.00000000000000000e+00))) {
return cast<float>(go_math_Inf(cast<int64_t>(1ULL)));
}
return cast<float>(cast<double>((cast<double>(1.00000000000000000e+00) / go_math_Sqrt(cast<double>(f)))));
}
}
// tools/platform/n3ds/gpu_shader_cache.go:82:1
n3ds_shInst* n3ds_GPU_shaderInst(n3ds_GPU* g,int64_t pc){
{
if ((g->decEpoch[pc] != g->shEpoch)) {
n3ds_GPU_decode(g,pc);
g->decEpoch[pc] = g->shEpoch;
}
return (&g->dec[pc]);
}
}
// tools/platform/n3ds/gpu_shader_cache.go:99:1
void n3ds_GPU_decodeAll(n3ds_GPU* g){
{
if ((g->decodedAll == g->shEpoch)) {
return ;
}
{auto&& tmp220 = g->Code;
for(int64_t tmp221=0;tmp221<len(tmp220);++tmp221){
auto pc=tmp221;if ((g->decEpoch[pc] != g->shEpoch)) {
n3ds_GPU_decode(g,pc);
g->decEpoch[pc] = g->shEpoch;
}
}}
g->decodedAll = g->shEpoch;
}
}
// tools/platform/n3ds/gpu_shader_cache.go:115:1
void n3ds_GPU_invalidateShaders(n3ds_GPU* g){
{
g->shEpoch++;
if ((g->shEpoch == cast<uint32_t>(0ULL))) {
g->decEpoch = std::array<uint32_t,4096>{};
g->shEpoch = cast<uint32_t>(1ULL);
}
}
}
// tools/platform/n3ds/gpu_shader_cache.go:125:1
void n3ds_GPU_decode(n3ds_GPU* g,int64_t pc){
{
uint32_t in = g->Code[pc];
uint32_t op = shr<uint32_t>(in,cast<int64_t>(26ULL));
n3ds_shInst* d = (&g->dec[pc]);
(*d) = n3ds_shInst{cast<uint8_t>(0ULL),cast<uint8_t>(op),{},{},{},{},{},{},{}};
{
switch(op){
case cast<uint32_t>(33ULL):{
d->kind = cast<uint8_t>(1ULL);
return ;
break;}
case cast<uint32_t>(34ULL):{
d->kind = cast<uint8_t>(2ULL);
return ;
break;}
case cast<uint32_t>(36ULL):case cast<uint32_t>(37ULL):case cast<uint32_t>(38ULL):case cast<uint32_t>(39ULL):case cast<uint32_t>(40ULL):case cast<uint32_t>(41ULL):case cast<uint32_t>(44ULL):case cast<uint32_t>(45ULL):{
d->kind = cast<uint8_t>(3ULL);
return ;
break;}
}}
{
if ((op >= cast<uint32_t>(48ULL))){
uint32_t desc = g->Opdesc[cast<uint32_t>((in & cast<uint32_t>(31ULL)))];
d->kind = cast<uint8_t>(5ULL);
n3ds_shInst_setDst(d,cast<int64_t>(cast<uint32_t>((shr<uint32_t>(in,cast<int64_t>(24ULL)) & cast<uint32_t>(31ULL)))),desc);
d->src[cast<int64_t>(0ULL)] = n3ds_decodeSrc(cast<int64_t>(cast<uint32_t>((shr<uint32_t>(in,cast<int64_t>(17ULL)) & cast<uint32_t>(31ULL)))),cast<int64_t>(0ULL),desc,cast<int64_t>(0ULL));
if ((op >= cast<uint32_t>(56ULL))) {
d->src[cast<int64_t>(1ULL)] = n3ds_decodeSrc(cast<int64_t>(cast<uint32_t>((shr<uint32_t>(in,cast<int64_t>(10ULL)) & cast<uint32_t>(127ULL)))),cast<int64_t>(cast<uint32_t>((shr<uint32_t>(in,cast<int64_t>(22ULL)) & cast<uint32_t>(3ULL)))),desc,cast<int64_t>(1ULL));
d->src[cast<int64_t>(2ULL)] = n3ds_decodeSrc(cast<int64_t>(cast<uint32_t>((shr<uint32_t>(in,cast<int64_t>(5ULL)) & cast<uint32_t>(31ULL)))),cast<int64_t>(0ULL),desc,cast<int64_t>(2ULL));
}
else {
d->src[cast<int64_t>(1ULL)] = n3ds_decodeSrc(cast<int64_t>(cast<uint32_t>((shr<uint32_t>(in,cast<int64_t>(12ULL)) & cast<uint32_t>(31ULL)))),cast<int64_t>(0ULL),desc,cast<int64_t>(1ULL));
d->src[cast<int64_t>(2ULL)] = n3ds_decodeSrc(cast<int64_t>(cast<uint32_t>((shr<uint32_t>(in,cast<int64_t>(5ULL)) & cast<uint32_t>(127ULL)))),cast<int64_t>(cast<uint32_t>((shr<uint32_t>(in,cast<int64_t>(22ULL)) & cast<uint32_t>(3ULL)))),desc,cast<int64_t>(2ULL));
}
}
else if ((shr<uint32_t>(op,cast<int64_t>(1ULL)) == cast<uint32_t>(23ULL))){
uint32_t desc = g->Opdesc[cast<uint32_t>((in & cast<uint32_t>(127ULL)))];
d->kind = cast<uint8_t>(6ULL);
d->src[cast<int64_t>(0ULL)] = n3ds_decodeSrc(cast<int64_t>(cast<uint32_t>((shr<uint32_t>(in,cast<int64_t>(12ULL)) & cast<uint32_t>(127ULL)))),cast<int64_t>(cast<uint32_t>((shr<uint32_t>(in,cast<int64_t>(19ULL)) & cast<uint32_t>(3ULL)))),desc,cast<int64_t>(0ULL));
d->src[cast<int64_t>(1ULL)] = n3ds_decodeSrc(cast<int64_t>(cast<uint32_t>((shr<uint32_t>(in,cast<int64_t>(7ULL)) & cast<uint32_t>(31ULL)))),cast<int64_t>(0ULL),desc,cast<int64_t>(1ULL));
d->cmpX = cast<uint8_t>(cast<uint32_t>((shr<uint32_t>(in,cast<int64_t>(24ULL)) & cast<uint32_t>(7ULL))));
d->cmpY = cast<uint8_t>(cast<uint32_t>((shr<uint32_t>(in,cast<int64_t>(21ULL)) & cast<uint32_t>(7ULL))));
}
else {
uint32_t desc = g->Opdesc[cast<uint32_t>((in & cast<uint32_t>(127ULL)))];
d->kind = cast<uint8_t>(4ULL);
n3ds_shInst_setDst(d,cast<int64_t>(cast<uint32_t>((shr<uint32_t>(in,cast<int64_t>(21ULL)) & cast<uint32_t>(31ULL)))),desc);
int64_t idx = cast<int64_t>(cast<uint32_t>((shr<uint32_t>(in,cast<int64_t>(19ULL)) & cast<uint32_t>(3ULL))));
{
switch(op){
case cast<uint32_t>(24ULL):case cast<uint32_t>(25ULL):case cast<uint32_t>(26ULL):case cast<uint32_t>(27ULL):{
d->src[cast<int64_t>(0ULL)] = n3ds_decodeSrc(cast<int64_t>(cast<uint32_t>((shr<uint32_t>(in,cast<int64_t>(14ULL)) & cast<uint32_t>(31ULL)))),cast<int64_t>(0ULL),desc,cast<int64_t>(0ULL));
d->src[cast<int64_t>(1ULL)] = n3ds_decodeSrc(cast<int64_t>(cast<uint32_t>((shr<uint32_t>(in,cast<int64_t>(7ULL)) & cast<uint32_t>(127ULL)))),idx,desc,cast<int64_t>(1ULL));
break;}
default:{
d->src[cast<int64_t>(0ULL)] = n3ds_decodeSrc(cast<int64_t>(cast<uint32_t>((shr<uint32_t>(in,cast<int64_t>(12ULL)) & cast<uint32_t>(127ULL)))),idx,desc,cast<int64_t>(0ULL));
d->src[cast<int64_t>(1ULL)] = n3ds_decodeSrc(cast<int64_t>(cast<uint32_t>((shr<uint32_t>(in,cast<int64_t>(7ULL)) & cast<uint32_t>(31ULL)))),cast<int64_t>(0ULL),desc,cast<int64_t>(1ULL));
break;}
}}
}
}
tmp222:;
}
}
// tools/platform/n3ds/gpu_shader_cache.go:182:1
void n3ds_shInst_setDst(n3ds_shInst* d,int64_t reg,uint32_t desc){
{
if ((reg < cast<int64_t>(16ULL))) {
auto tmp223 = std::make_tuple(false,cast<uint8_t>(reg));
d->dstTmp = std::get<0>(tmp223);
d->dst = std::get<1>(tmp223);
}
else {
auto tmp224 = std::make_tuple(true,cast<uint8_t>(cast<int64_t>((reg - cast<int64_t>(16ULL)))));
d->dstTmp = std::get<0>(tmp224);
d->dst = std::get<1>(tmp224);
}
d->maskAll = true;
{uint64_t i = cast<uint64_t>(0ULL);for (;(i < cast<uint64_t>(4ULL));i++){
d->mask[i] = (cast<uint32_t>((shr<uint32_t>(desc,(cast<uint64_t>((cast<uint64_t>(3ULL) - i)))) & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL));
d->maskAll = (d->maskAll && d->mask[i]);
}
}}
}
// tools/platform/n3ds/gpu_shader_cache.go:201:1
n3ds_shSrc n3ds_decodeSrc(int64_t reg,int64_t idx,uint32_t desc,int64_t n){
{
n3ds_shSrc s={};
{
if ((reg < cast<int64_t>(16ULL))){
auto tmp226 = std::make_tuple(cast<uint8_t>(0ULL),cast<uint8_t>(reg));
s.bank = std::get<0>(tmp226);
s.reg = std::get<1>(tmp226);
}
else if ((reg < cast<int64_t>(32ULL))){
auto tmp227 = std::make_tuple(cast<uint8_t>(1ULL),cast<uint8_t>(cast<int64_t>((reg - cast<int64_t>(16ULL)))));
s.bank = std::get<0>(tmp227);
s.reg = std::get<1>(tmp227);
}
else {
auto tmp228 = std::make_tuple(cast<uint8_t>(2ULL),cast<uint8_t>(cast<int64_t>((reg - cast<int64_t>(32ULL)))));
s.bank = std::get<0>(tmp228);
s.reg = std::get<1>(tmp228);
s.idx = cast<uint8_t>(idx);
}
}
tmp225:;
uint64_t shift = n3ds_shSrcShift[n];
s.neg = (cast<uint32_t>((shr<uint32_t>(desc,shift) & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL));
uint32_t sw = cast<uint32_t>((shr<uint32_t>(desc,(cast<uint64_t>((shift + cast<uint64_t>(1ULL))))) & cast<uint32_t>(255ULL)));
{uint64_t i = cast<uint64_t>(0ULL);for (;(i < cast<uint64_t>(4ULL));i++){
s.sw[i] = cast<uint8_t>(cast<uint32_t>((shr<uint32_t>(sw,(cast<uint64_t>((cast<uint64_t>(6ULL) - cast<uint64_t>((cast<uint64_t>(2ULL) * i)))))) & cast<uint32_t>(3ULL))));
}
}s.plain = ((!s.neg) && (s.sw == std::array<uint8_t,4>{cast<uint8_t>(0ULL),cast<uint8_t>(1ULL),cast<uint8_t>(2ULL),cast<uint8_t>(3ULL)}));
return s;
}
}
// tools/platform/n3ds/gpu_tev.go:31:1
std::tuple<uint8_t,uint8_t,uint8_t,uint8_t,bool,bool> n3ds_GPU_fragment(n3ds_GPU* g,std::array<float,4> vcol,std::array<std::array<float,2>,3> uv,float uv0w,n3ds_lightState* ls,n3ds_tevState* tv,std::array<float,4> quat,std::array<float,3> view,n3ds_rstats* st){
uint8_t r{};
uint8_t gr{};
uint8_t b{};
uint8_t a{};
bool discard{};
bool ok{};
{
n3ds_rgba vertex = n3ds_rgba{n3ds_clamp255(cast<float>((vcol[cast<int64_t>(0ULL)] * cast<float>(2.55000000000000000e+02)))),n3ds_clamp255(cast<float>((vcol[cast<int64_t>(1ULL)] * cast<float>(2.55000000000000000e+02)))),n3ds_clamp255(cast<float>((vcol[cast<int64_t>(2ULL)] * cast<float>(2.55000000000000000e+02)))),n3ds_clamp255(cast<float>((vcol[cast<int64_t>(3ULL)] * cast<float>(2.55000000000000000e+02))))};
std::array<n3ds_rgba,3> tex={};
{int64_t u = cast<int64_t>(0ULL);for (;(u < cast<int64_t>(3ULL));u++){
if ((cast<uint32_t>((shr<uint32_t>(tv->texEnable,cast<uint64_t>(u)) & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL))) {
bool oks={};
auto tmp229 = n3ds_GPU_sampleTextureSt(g,u,uv[u][cast<int64_t>(0ULL)],uv[u][cast<int64_t>(1ULL)],uv0w,st);
tex[u] = std::get<0>(tmp229);
oks = std::get<1>(tmp229);
if ((!oks)) {
return {cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),false,false};
}
}
}
}auto tmp230 = std::make_tuple(vertex,n3ds_rgba{cast<int32_t>(0ULL),cast<int32_t>(0ULL),cast<int32_t>(0ULL),cast<int32_t>(0ULL)});
n3ds_rgba fragPrim = std::get<0>(tmp230);
n3ds_rgba fragSec = std::get<1>(tmp230);
if (ls->enabled) {
auto tmp231 = n3ds_GPU_shade(g,ls,quat,view,(&tex));
std::array<float,4> p = std::get<0>(tmp231);
std::array<float,4> sc = std::get<1>(tmp231);
fragPrim = n3ds_rgba{n3ds_clamp255(cast<float>((p[cast<int64_t>(0ULL)] * cast<float>(2.55000000000000000e+02)))),n3ds_clamp255(cast<float>((p[cast<int64_t>(1ULL)] * cast<float>(2.55000000000000000e+02)))),n3ds_clamp255(cast<float>((p[cast<int64_t>(2ULL)] * cast<float>(2.55000000000000000e+02)))),n3ds_clamp255(cast<float>((p[cast<int64_t>(3ULL)] * cast<float>(2.55000000000000000e+02))))};
fragSec = n3ds_rgba{n3ds_clamp255(cast<float>((sc[cast<int64_t>(0ULL)] * cast<float>(2.55000000000000000e+02)))),n3ds_clamp255(cast<float>((sc[cast<int64_t>(1ULL)] * cast<float>(2.55000000000000000e+02)))),n3ds_clamp255(cast<float>((sc[cast<int64_t>(2ULL)] * cast<float>(2.55000000000000000e+02)))),n3ds_clamp255(cast<float>((sc[cast<int64_t>(3ULL)] * cast<float>(2.55000000000000000e+02))))};
}
auto tmp232 = n3ds_tevState_run(tv,vertex,fragPrim,fragSec,tex);
n3ds_rgba prev = std::get<0>(tmp232);
bool okr = std::get<1>(tmp232);
if ((!okr)) {
arm_CPU_Halt(g->m->CPU,std::string("gpu tev: a stage uses a source or combine op this model does not implement",74));
return {cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),false,false};
}
if (tv->alphaTest) {
int32_t ref = tv->alphaRef;
bool pass={};
{
switch(tv->alphaFunc){
case cast<uint8_t>(0ULL):{
pass = false;
break;}
case cast<uint8_t>(1ULL):{
pass = true;
break;}
case cast<uint8_t>(2ULL):{
pass = (prev.a == ref);
break;}
case cast<uint8_t>(3ULL):{
pass = (prev.a != ref);
break;}
case cast<uint8_t>(4ULL):{
pass = (prev.a < ref);
break;}
case cast<uint8_t>(5ULL):{
pass = (prev.a <= ref);
break;}
case cast<uint8_t>(6ULL):{
pass = (prev.a > ref);
break;}
default:{
pass = (prev.a >= ref);
break;}
}}
if ((!pass)) {
return {cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),true,true};
}
}
return {cast<uint8_t>(prev.r),cast<uint8_t>(prev.g),cast<uint8_t>(prev.b),cast<uint8_t>(prev.a),false,true};
}
}
// tools/platform/n3ds/gpu_tev.go:92:1
n3ds_rgba n3ds_tevColorOperand(n3ds_rgba s,uint32_t op){
{
{
switch(op){
case cast<uint32_t>(0ULL):{
return s;
break;}
case cast<uint32_t>(1ULL):{
return n3ds_rgba{cast<int32_t>((cast<int32_t>(255ULL) - s.r)),cast<int32_t>((cast<int32_t>(255ULL) - s.g)),cast<int32_t>((cast<int32_t>(255ULL) - s.b)),s.a};
break;}
case cast<uint32_t>(2ULL):{
return n3ds_rgba{s.a,s.a,s.a,s.a};
break;}
case cast<uint32_t>(3ULL):{
return n3ds_rgba{cast<int32_t>((cast<int32_t>(255ULL) - s.a)),cast<int32_t>((cast<int32_t>(255ULL) - s.a)),cast<int32_t>((cast<int32_t>(255ULL) - s.a)),cast<int32_t>((cast<int32_t>(255ULL) - s.a))};
break;}
case cast<uint32_t>(4ULL):{
return n3ds_rgba{s.r,s.r,s.r,s.r};
break;}
case cast<uint32_t>(5ULL):{
return n3ds_rgba{cast<int32_t>((cast<int32_t>(255ULL) - s.r)),cast<int32_t>((cast<int32_t>(255ULL) - s.r)),cast<int32_t>((cast<int32_t>(255ULL) - s.r)),cast<int32_t>((cast<int32_t>(255ULL) - s.r))};
break;}
case cast<uint32_t>(8ULL):{
return n3ds_rgba{s.g,s.g,s.g,s.g};
break;}
case cast<uint32_t>(9ULL):{
return n3ds_rgba{cast<int32_t>((cast<int32_t>(255ULL) - s.g)),cast<int32_t>((cast<int32_t>(255ULL) - s.g)),cast<int32_t>((cast<int32_t>(255ULL) - s.g)),cast<int32_t>((cast<int32_t>(255ULL) - s.g))};
break;}
case cast<uint32_t>(12ULL):{
return n3ds_rgba{s.b,s.b,s.b,s.b};
break;}
case cast<uint32_t>(13ULL):{
return n3ds_rgba{cast<int32_t>((cast<int32_t>(255ULL) - s.b)),cast<int32_t>((cast<int32_t>(255ULL) - s.b)),cast<int32_t>((cast<int32_t>(255ULL) - s.b)),cast<int32_t>((cast<int32_t>(255ULL) - s.b))};
break;}
}}
return s;
}
}
// tools/platform/n3ds/gpu_tev.go:119:1
int32_t n3ds_tevAlphaOperand(n3ds_rgba s,uint32_t op){
{
{
switch(op){
case cast<uint32_t>(0ULL):{
return s.a;
break;}
case cast<uint32_t>(1ULL):{
return cast<int32_t>((cast<int32_t>(255ULL) - s.a));
break;}
case cast<uint32_t>(2ULL):{
return s.r;
break;}
case cast<uint32_t>(3ULL):{
return cast<int32_t>((cast<int32_t>(255ULL) - s.r));
break;}
case cast<uint32_t>(4ULL):{
return s.g;
break;}
case cast<uint32_t>(5ULL):{
return cast<int32_t>((cast<int32_t>(255ULL) - s.g));
break;}
case cast<uint32_t>(6ULL):{
return s.b;
break;}
default:{
return cast<int32_t>((cast<int32_t>(255ULL) - s.b));
break;}
}}
}
}
// tools/platform/n3ds/gpu_tev.go:141:1
std::tuple<int32_t,bool> n3ds_tevCombine(uint32_t op,int32_t a,int32_t b,int32_t c){
{
{
switch(op){
case cast<uint32_t>(0ULL):{
return {a,true};
break;}
case cast<uint32_t>(1ULL):{
return {divi<int32_t>(cast<int32_t>((a * b)),cast<int32_t>(255ULL)),true};
break;}
case cast<uint32_t>(2ULL):{
return {cast<int32_t>((a + b)),true};
break;}
case cast<uint32_t>(3ULL):{
return {cast<int32_t>((cast<int32_t>((a + b)) - cast<int32_t>(128ULL))),true};
break;}
case cast<uint32_t>(4ULL):{
return {divi<int32_t>((cast<int32_t>((cast<int32_t>((a * c)) + cast<int32_t>((b * (cast<int32_t>((cast<int32_t>(255ULL) - c)))))))),cast<int32_t>(255ULL)),true};
break;}
case cast<uint32_t>(5ULL):{
return {cast<int32_t>((a - b)),true};
break;}
case cast<uint32_t>(8ULL):{
return {cast<int32_t>((divi<int32_t>(cast<int32_t>((a * b)),cast<int32_t>(255ULL)) + c)),true};
break;}
case cast<uint32_t>(9ULL):{
return {divi<int32_t>(cast<int32_t>((n3ds_clampi(cast<int32_t>((a + b)),cast<int32_t>(0ULL),cast<int32_t>(255ULL)) * c)),cast<int32_t>(255ULL)),true};
break;}
}}
return {cast<int32_t>(0ULL),false};
}
}
// tools/platform/n3ds/gpu_tev.go:165:1
uint8_t n3ds_logicOp(uint32_t op,uint8_t s,uint8_t d){
{
{
switch(op){
case cast<uint32_t>(0ULL):{
return cast<uint8_t>(0ULL);
break;}
case cast<uint32_t>(1ULL):{
return cast<uint8_t>((s & d));
break;}
case cast<uint32_t>(2ULL):{
return (s & ~(d));
break;}
case cast<uint32_t>(3ULL):{
return s;
break;}
case cast<uint32_t>(4ULL):{
return cast<uint8_t>(255ULL);
break;}
case cast<uint32_t>(5ULL):{
return cast<uint8_t>(~s);
break;}
case cast<uint32_t>(6ULL):{
return d;
break;}
case cast<uint32_t>(7ULL):{
return cast<uint8_t>(~d);
break;}
case cast<uint32_t>(8ULL):{
return cast<uint8_t>(~(cast<uint8_t>((s & d))));
break;}
case cast<uint32_t>(9ULL):{
return cast<uint8_t>((s | d));
break;}
case cast<uint32_t>(10ULL):{
return cast<uint8_t>(~(cast<uint8_t>((s | d))));
break;}
case cast<uint32_t>(11ULL):{
return cast<uint8_t>((s ^ d));
break;}
case cast<uint32_t>(12ULL):{
return cast<uint8_t>(~(cast<uint8_t>((s ^ d))));
break;}
case cast<uint32_t>(13ULL):{
return (d & ~(s));
break;}
case cast<uint32_t>(14ULL):{
return cast<uint8_t>((s | cast<uint8_t>(~d)));
break;}
}}
return cast<uint8_t>((cast<uint8_t>(~s) | d));
}
}
// tools/platform/n3ds/gpu_tev.go:203:1
std::tuple<uint8_t,uint8_t,uint8_t,uint8_t> n3ds_GPU_blend(n3ds_GPU* g,uint8_t sr,uint8_t sg,uint8_t sb,uint8_t sa,uint8_t dr,uint8_t dg,uint8_t db,uint8_t da){
{
if ((cast<uint32_t>((shr<uint32_t>(g->Regs[cast<int64_t>(256ULL)],cast<int64_t>(8ULL)) & cast<uint32_t>(1ULL))) == cast<uint32_t>(0ULL))) {
uint32_t lop = cast<uint32_t>((g->Regs[cast<int64_t>(258ULL)] & cast<uint32_t>(15ULL)));
return {n3ds_logicOp(lop,sr,dr),n3ds_logicOp(lop,sg,dg),n3ds_logicOp(lop,sb,db),n3ds_logicOp(lop,sa,da)};
}
uint32_t cfg = g->Regs[cast<int64_t>(257ULL)];
auto tmp233 = std::make_tuple(cast<uint32_t>((cfg & cast<uint32_t>(7ULL))),cast<uint32_t>((shr<uint32_t>(cfg,cast<int64_t>(8ULL)) & cast<uint32_t>(7ULL))));
uint32_t eqC = std::get<0>(tmp233);
uint32_t eqA = std::get<1>(tmp233);
auto tmp234 = std::make_tuple(cast<uint32_t>((shr<uint32_t>(cfg,cast<int64_t>(16ULL)) & cast<uint32_t>(15ULL))),cast<uint32_t>((shr<uint32_t>(cfg,cast<int64_t>(20ULL)) & cast<uint32_t>(15ULL))));
uint32_t fsC = std::get<0>(tmp234);
uint32_t fdC = std::get<1>(tmp234);
auto tmp235 = std::make_tuple(cast<uint32_t>((shr<uint32_t>(cfg,cast<int64_t>(24ULL)) & cast<uint32_t>(15ULL))),cast<uint32_t>((shr<uint32_t>(cfg,cast<int64_t>(28ULL)) & cast<uint32_t>(15ULL))));
uint32_t fsA = std::get<0>(tmp235);
uint32_t fdA = std::get<1>(tmp235);
auto factor = [&](uint32_t code,int32_t s,int32_t d,int32_t sA,int32_t dA,bool isA)->std::tuple<int32_t,int32_t,int32_t>{
uint32_t cc = g->Regs[cast<int64_t>(259ULL)];
auto tmp236 = std::make_tuple(cast<int32_t>(cast<uint32_t>((cc & cast<uint32_t>(255ULL)))),cast<int32_t>(cast<uint32_t>((shr<uint32_t>(cc,cast<int64_t>(8ULL)) & cast<uint32_t>(255ULL)))),cast<int32_t>(cast<uint32_t>((shr<uint32_t>(cc,cast<int64_t>(16ULL)) & cast<uint32_t>(255ULL)))));
int32_t cr = std::get<0>(tmp236);
int32_t cg2 = std::get<1>(tmp236);
int32_t cb2 = std::get<2>(tmp236);
int32_t ca = cast<int32_t>(cast<uint32_t>((shr<uint32_t>(cc,cast<int64_t>(24ULL)) & cast<uint32_t>(255ULL))));
{
switch(code){
case cast<uint32_t>(0ULL):{
return {cast<int32_t>(0ULL),cast<int32_t>(0ULL),cast<int32_t>(0ULL)};
break;}
case cast<uint32_t>(1ULL):{
return {cast<int32_t>(255ULL),cast<int32_t>(255ULL),cast<int32_t>(255ULL)};
break;}
case cast<uint32_t>(2ULL):{
return {s,s,s};
break;}
case cast<uint32_t>(3ULL):{
return {cast<int32_t>((cast<int32_t>(255ULL) - s)),cast<int32_t>((cast<int32_t>(255ULL) - s)),cast<int32_t>((cast<int32_t>(255ULL) - s))};
break;}
case cast<uint32_t>(4ULL):{
return {d,d,d};
break;}
case cast<uint32_t>(5ULL):{
return {cast<int32_t>((cast<int32_t>(255ULL) - d)),cast<int32_t>((cast<int32_t>(255ULL) - d)),cast<int32_t>((cast<int32_t>(255ULL) - d))};
break;}
case cast<uint32_t>(6ULL):{
return {sA,sA,sA};
break;}
case cast<uint32_t>(7ULL):{
return {cast<int32_t>((cast<int32_t>(255ULL) - sA)),cast<int32_t>((cast<int32_t>(255ULL) - sA)),cast<int32_t>((cast<int32_t>(255ULL) - sA))};
break;}
case cast<uint32_t>(8ULL):{
return {dA,dA,dA};
break;}
case cast<uint32_t>(9ULL):{
return {cast<int32_t>((cast<int32_t>(255ULL) - dA)),cast<int32_t>((cast<int32_t>(255ULL) - dA)),cast<int32_t>((cast<int32_t>(255ULL) - dA))};
break;}
case cast<uint32_t>(10ULL):{
if (isA) {
return {ca,ca,ca};
}
return {cr,cg2,cb2};
break;}
case cast<uint32_t>(11ULL):{
if (isA) {
return {cast<int32_t>((cast<int32_t>(255ULL) - ca)),cast<int32_t>((cast<int32_t>(255ULL) - ca)),cast<int32_t>((cast<int32_t>(255ULL) - ca))};
}
return {cast<int32_t>((cast<int32_t>(255ULL) - cr)),cast<int32_t>((cast<int32_t>(255ULL) - cg2)),cast<int32_t>((cast<int32_t>(255ULL) - cb2))};
break;}
case cast<uint32_t>(12ULL):{
return {ca,ca,ca};
break;}
case cast<uint32_t>(13ULL):{
return {cast<int32_t>((cast<int32_t>(255ULL) - ca)),cast<int32_t>((cast<int32_t>(255ULL) - ca)),cast<int32_t>((cast<int32_t>(255ULL) - ca))};
break;}
case cast<uint32_t>(14ULL):{
int32_t m = sA;
if ((cast<int32_t>((cast<int32_t>(255ULL) - dA)) < m)) {
m = cast<int32_t>((cast<int32_t>(255ULL) - dA));
}
return {m,m,m};
break;}
}}
arm_CPU_Halt(g->m->CPU,std::string("gpu: blend factor %d unimplemented",34),code);
return {cast<int32_t>(0ULL),cast<int32_t>(0ULL),cast<int32_t>(0ULL)};
}
;
auto apply = [&](uint32_t eq,int32_t s,int32_t d,int32_t fs,int32_t fd)->int32_t{
{
switch(eq){
case cast<uint32_t>(0ULL):{
return n3ds_clampi(divi<int32_t>((cast<int32_t>((cast<int32_t>((s * fs)) + cast<int32_t>((d * fd))))),cast<int32_t>(255ULL)),cast<int32_t>(0ULL),cast<int32_t>(255ULL));
break;}
case cast<uint32_t>(1ULL):{
return n3ds_clampi(divi<int32_t>((cast<int32_t>((cast<int32_t>((s * fs)) - cast<int32_t>((d * fd))))),cast<int32_t>(255ULL)),cast<int32_t>(0ULL),cast<int32_t>(255ULL));
break;}
case cast<uint32_t>(2ULL):{
return n3ds_clampi(divi<int32_t>((cast<int32_t>((cast<int32_t>((d * fd)) - cast<int32_t>((s * fs))))),cast<int32_t>(255ULL)),cast<int32_t>(0ULL),cast<int32_t>(255ULL));
break;}
case cast<uint32_t>(3ULL):{
if ((d < s)) {
return d;
}
return s;
break;}
case cast<uint32_t>(4ULL):{
if ((d > s)) {
return d;
}
return s;
break;}
}}
arm_CPU_Halt(g->m->CPU,std::string("gpu: blend equation %d unimplemented",36),eq);
return s;
}
;
auto tmp237 = std::make_tuple(cast<int32_t>(sr),cast<int32_t>(sg),cast<int32_t>(sb),cast<int32_t>(sa));
int32_t sri = std::get<0>(tmp237);
int32_t sgi = std::get<1>(tmp237);
int32_t sbi = std::get<2>(tmp237);
int32_t sai = std::get<3>(tmp237);
auto tmp238 = std::make_tuple(cast<int32_t>(dr),cast<int32_t>(dg),cast<int32_t>(db),cast<int32_t>(da));
int32_t dri = std::get<0>(tmp238);
int32_t dgi = std::get<1>(tmp238);
int32_t dbi = std::get<2>(tmp238);
int32_t dai = std::get<3>(tmp238);
auto tmp239 = factor(fsC,sri,dri,sai,dai,false);
int32_t fsr = std::get<0>(tmp239);
auto tmp240 = factor(fsC,sgi,dgi,sai,dai,false);
int32_t fsg = std::get<1>(tmp240);
auto tmp241 = factor(fsC,sbi,dbi,sai,dai,false);
int32_t fsb = std::get<2>(tmp241);
auto tmp242 = factor(fdC,sri,dri,sai,dai,false);
int32_t fdr = std::get<0>(tmp242);
auto tmp243 = factor(fdC,sgi,dgi,sai,dai,false);
int32_t fdg = std::get<1>(tmp243);
auto tmp244 = factor(fdC,sbi,dbi,sai,dai,false);
int32_t fdb = std::get<2>(tmp244);
auto tmp245 = factor(fsA,sai,dai,sai,dai,true);
int32_t fsa = std::get<0>(tmp245);
auto tmp246 = factor(fdA,sai,dai,sai,dai,true);
int32_t fda = std::get<0>(tmp246);
return {cast<uint8_t>(apply(eqC,sri,dri,fsr,fdr)),cast<uint8_t>(apply(eqC,sgi,dgi,fsg,fdg)),cast<uint8_t>(apply(eqC,sbi,dbi,fsb,fdb)),cast<uint8_t>(apply(eqA,sai,dai,fsa,fda))};
}
}
// tools/platform/n3ds/gpu_tev.go:312:1
int32_t n3ds_clamp255(float f){
{
if ((f < cast<float>(0.00000000000000000e+00))) {
return cast<int32_t>(0ULL);
}
if ((f > cast<float>(2.55000000000000000e+02))) {
return cast<int32_t>(255ULL);
}
return cast<int32_t>(f);
}
}
// tools/platform/n3ds/gpu_tev.go:322:1
int32_t n3ds_clampi(int32_t v,int32_t lo,int32_t hi){
{
if ((v < lo)) {
return lo;
}
if ((v > hi)) {
return hi;
}
return v;
}
}
// tools/platform/n3ds/gpu_tev.go:338:1
std::tuple<n3ds_rgba,bool> n3ds_tevState_run(n3ds_tevState* t,n3ds_rgba vertex,n3ds_rgba fragPrim,n3ds_rgba fragSec,std::array<n3ds_rgba,3> tex){
{
n3ds_rgba prev = vertex;
n3ds_rgba buf={};
n3ds_rgba next = t->bufColor;
auto fetch = [&](uint8_t sel)->std::tuple<n3ds_rgba,bool>{
{
switch(sel){
case cast<uint8_t>(0ULL):{
return {vertex,true};
break;}
case cast<uint8_t>(1ULL):{
return {fragPrim,true};
break;}
case cast<uint8_t>(2ULL):{
return {fragSec,true};
break;}
case cast<uint8_t>(3ULL):{
return {tex[cast<int64_t>(0ULL)],true};
break;}
case cast<uint8_t>(4ULL):{
return {tex[cast<int64_t>(1ULL)],true};
break;}
case cast<uint8_t>(5ULL):{
return {tex[cast<int64_t>(2ULL)],true};
break;}
case cast<uint8_t>(13ULL):{
return {buf,true};
break;}
case cast<uint8_t>(15ULL):{
return {prev,true};
break;}
}}
return {n3ds_rgba{},false};
}
;
{auto&& tmp247 = t->stages;
for(int64_t tmp248=0;tmp248<len(tmp247);++tmp248){
auto i=tmp248;n3ds_tevStage* s = (&t->stages[i]);
std::array<n3ds_rgba,3> cin={};
{int64_t j = cast<int64_t>(0ULL);for (;(j < cast<int64_t>(3ULL));j++){
n3ds_tevOperand o = s->colr[j];
auto tmp249 = std::make_tuple(n3ds_rgba{},true);
n3ds_rgba src = std::get<0>(tmp249);
bool okf = std::get<1>(tmp249);
if ((o.src == cast<uint8_t>(14ULL))) {
src = s->konst;
}
else {
auto tmp250 = fetch(o.src);
src = std::get<0>(tmp250);
okf = std::get<1>(tmp250);
if ((!okf)) {
return {n3ds_rgba{},false};
}
}
cin[j] = n3ds_tevColorOperand(src,cast<uint32_t>(o.op));
}
}std::array<int32_t,3> ain={};
{int64_t j = cast<int64_t>(0ULL);for (;(j < cast<int64_t>(3ULL));j++){
n3ds_tevOperand o = s->alph[j];
auto tmp251 = std::make_tuple(n3ds_rgba{},true);
n3ds_rgba src = std::get<0>(tmp251);
bool okf = std::get<1>(tmp251);
if ((o.src == cast<uint8_t>(14ULL))) {
src = s->konst;
}
else {
auto tmp252 = fetch(o.src);
src = std::get<0>(tmp252);
okf = std::get<1>(tmp252);
if ((!okf)) {
return {n3ds_rgba{},false};
}
}
ain[j] = n3ds_tevAlphaOperand(src,cast<uint32_t>(o.op));
}
}auto tmp253 = n3ds_tevCombine(cast<uint32_t>(s->combC),cin[cast<int64_t>(0ULL)].r,cin[cast<int64_t>(1ULL)].r,cin[cast<int64_t>(2ULL)].r);
int32_t cr = std::get<0>(tmp253);
bool okc = std::get<1>(tmp253);
auto tmp254 = n3ds_tevCombine(cast<uint32_t>(s->combC),cin[cast<int64_t>(0ULL)].g,cin[cast<int64_t>(1ULL)].g,cin[cast<int64_t>(2ULL)].g);
int32_t cg = std::get<0>(tmp254);
auto tmp255 = n3ds_tevCombine(cast<uint32_t>(s->combC),cin[cast<int64_t>(0ULL)].b,cin[cast<int64_t>(1ULL)].b,cin[cast<int64_t>(2ULL)].b);
int32_t cb = std::get<0>(tmp255);
auto tmp256 = n3ds_tevCombine(cast<uint32_t>(s->combA),ain[cast<int64_t>(0ULL)],ain[cast<int64_t>(1ULL)],ain[cast<int64_t>(2ULL)]);
int32_t ca = std::get<0>(tmp256);
bool oka = std::get<1>(tmp256);
if (((!okc) || (!oka))) {
return {n3ds_rgba{},false};
}
n3ds_rgba out = n3ds_rgba{n3ds_clampi(shl<int32_t>(cr,s->scaleC),cast<int32_t>(0ULL),cast<int32_t>(255ULL)),n3ds_clampi(shl<int32_t>(cg,s->scaleC),cast<int32_t>(0ULL),cast<int32_t>(255ULL)),n3ds_clampi(shl<int32_t>(cb,s->scaleC),cast<int32_t>(0ULL),cast<int32_t>(255ULL)),n3ds_clampi(shl<int32_t>(ca,s->scaleA),cast<int32_t>(0ULL),cast<int32_t>(255ULL))};
buf = next;
if (s->updC) {
auto tmp257 = std::make_tuple(out.r,out.g,out.b);
next.r = std::get<0>(tmp257);
next.g = std::get<1>(tmp257);
next.b = std::get<2>(tmp257);
}
if (s->updA) {
next.a = out.a;
}
prev = out;
}}
return {prev,true};
}
}
// tools/platform/n3ds/gpu_tev_state.go:50:1
n3ds_tevState n3ds_GPU_tevstate(n3ds_GPU* g){
{
return n3ds_tevStateFromRegs((&g->Regs));
}
}
// tools/platform/n3ds/gpu_tev_state.go:56:1
n3ds_tevState n3ds_tevStateFromRegs(std::array<uint32_t,768>* r){
{
n3ds_tevState t={};
t.texEnable = (*r)[cast<int64_t>(128ULL)];
uint32_t upd = (*r)[cast<int64_t>(224ULL)];
t.bufColor = n3ds_rgba{cast<int32_t>(cast<uint32_t>(((*r)[cast<int64_t>(253ULL)] & cast<uint32_t>(255ULL)))),cast<int32_t>(cast<uint32_t>((shr<uint32_t>((*r)[cast<int64_t>(253ULL)],cast<int64_t>(8ULL)) & cast<uint32_t>(255ULL)))),cast<int32_t>(cast<uint32_t>((shr<uint32_t>((*r)[cast<int64_t>(253ULL)],cast<int64_t>(16ULL)) & cast<uint32_t>(255ULL)))),cast<int32_t>(cast<uint32_t>((shr<uint32_t>((*r)[cast<int64_t>(253ULL)],cast<int64_t>(24ULL)) & cast<uint32_t>(255ULL))))};
{int64_t s = cast<int64_t>(0ULL);for (;(s < cast<int64_t>(6ULL));s++){
uint32_t base = n3ds_tevStageBase[s];
uint32_t src = (*r)[base];
uint32_t opd = (*r)[cast<uint32_t>((base + cast<uint32_t>(1ULL)))];
uint32_t cmb = (*r)[cast<uint32_t>((base + cast<uint32_t>(2ULL)))];
uint32_t scale = (*r)[cast<uint32_t>((base + cast<uint32_t>(4ULL)))];
n3ds_tevStage* st = (&t.stages[s]);
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(3ULL));i++){
st->colr[i] = n3ds_tevOperand{cast<uint8_t>(cast<uint32_t>((shr<uint32_t>(src,(cast<uint64_t>((cast<uint64_t>(4ULL) * cast<uint64_t>(i))))) & cast<uint32_t>(15ULL)))),cast<uint8_t>(cast<uint32_t>((shr<uint32_t>(opd,(cast<uint64_t>((cast<uint64_t>(4ULL) * cast<uint64_t>(i))))) & cast<uint32_t>(15ULL))))};
st->alph[i] = n3ds_tevOperand{cast<uint8_t>(cast<uint32_t>((shr<uint32_t>(src,(cast<uint64_t>((cast<uint64_t>(16ULL) + cast<uint64_t>((cast<uint64_t>(4ULL) * cast<uint64_t>(i))))))) & cast<uint32_t>(15ULL)))),cast<uint8_t>(cast<uint32_t>((shr<uint32_t>(opd,(cast<uint64_t>((cast<uint64_t>(12ULL) + cast<uint64_t>((cast<uint64_t>(4ULL) * cast<uint64_t>(i))))))) & cast<uint32_t>(7ULL))))};
}
}st->combC = cast<uint8_t>(cast<uint32_t>((cmb & cast<uint32_t>(15ULL))));
st->combA = cast<uint8_t>(cast<uint32_t>((shr<uint32_t>(cmb,cast<int64_t>(16ULL)) & cast<uint32_t>(15ULL))));
st->scaleC = cast<uint8_t>(cast<uint32_t>((scale & cast<uint32_t>(3ULL))));
st->scaleA = cast<uint8_t>(cast<uint32_t>((shr<uint32_t>(scale,cast<int64_t>(16ULL)) & cast<uint32_t>(3ULL))));
st->konst = n3ds_rgba{cast<int32_t>(cast<uint32_t>(((*r)[cast<uint32_t>((base + cast<uint32_t>(3ULL)))] & cast<uint32_t>(255ULL)))),cast<int32_t>(cast<uint32_t>((shr<uint32_t>((*r)[cast<uint32_t>((base + cast<uint32_t>(3ULL)))],cast<int64_t>(8ULL)) & cast<uint32_t>(255ULL)))),cast<int32_t>(cast<uint32_t>((shr<uint32_t>((*r)[cast<uint32_t>((base + cast<uint32_t>(3ULL)))],cast<int64_t>(16ULL)) & cast<uint32_t>(255ULL)))),cast<int32_t>(cast<uint32_t>((shr<uint32_t>((*r)[cast<uint32_t>((base + cast<uint32_t>(3ULL)))],cast<int64_t>(24ULL)) & cast<uint32_t>(255ULL))))};
if ((s < cast<int64_t>(4ULL))) {
st->updC = (cast<uint32_t>((shr<uint32_t>(upd,(cast<uint64_t>((cast<uint64_t>(8ULL) + cast<uint64_t>(s))))) & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL));
st->updA = (cast<uint32_t>((shr<uint32_t>(upd,(cast<uint64_t>((cast<uint64_t>(12ULL) + cast<uint64_t>(s))))) & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL));
}
}
}uint32_t at = (*r)[cast<int64_t>(260ULL)];
t.alphaTest = (cast<uint32_t>((at & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL));
t.alphaFunc = cast<uint8_t>(cast<uint32_t>((shr<uint32_t>(at,cast<int64_t>(4ULL)) & cast<uint32_t>(7ULL))));
t.alphaRef = cast<int32_t>(cast<uint32_t>((shr<uint32_t>(at,cast<int64_t>(8ULL)) & cast<uint32_t>(255ULL))));
return t;
}
}
// tools/platform/n3ds/gpu_texture.go:29:1
std::tuple<uint32_t,uint32_t,uint32_t,uint32_t> n3ds_texUnitRegs(int64_t u){
uint32_t dim{};
uint32_t param{};
uint32_t addr{};
uint32_t typ{};
{
{
switch(u){
case cast<int64_t>(0ULL):{
return {cast<uint32_t>(130ULL),cast<uint32_t>(131ULL),cast<uint32_t>(133ULL),cast<uint32_t>(142ULL)};
break;}
case cast<int64_t>(1ULL):{
return {cast<uint32_t>(146ULL),cast<uint32_t>(147ULL),cast<uint32_t>(149ULL),cast<uint32_t>(150ULL)};
break;}
default:{
return {cast<uint32_t>(154ULL),cast<uint32_t>(155ULL),cast<uint32_t>(157ULL),cast<uint32_t>(158ULL)};
break;}
}}
}
}
// tools/platform/n3ds/gpu_texture.go:47:1
std::tuple<n3ds_rgba,bool> n3ds_GPU_sampleTextureSt(n3ds_GPU* g,int64_t u,float s,float t,float q,n3ds_rstats* st){
{
auto tmp258 = n3ds_texUnitRegs(u);
uint32_t dimR = std::get<0>(tmp258);
uint32_t paramR = std::get<1>(tmp258);
uint32_t addrR = std::get<2>(tmp258);
uint32_t typR = std::get<3>(tmp258);
uint32_t w = cast<uint32_t>((shr<uint32_t>(g->Regs[dimR],cast<int64_t>(16ULL)) & cast<uint32_t>(2047ULL)));
uint32_t h = cast<uint32_t>((g->Regs[dimR] & cast<uint32_t>(2047ULL)));
if (((w == cast<uint32_t>(0ULL)) || (h == cast<uint32_t>(0ULL)))) {
return {n3ds_rgba{cast<int32_t>(255ULL),cast<int32_t>(255ULL),cast<int32_t>(255ULL),cast<int32_t>(255ULL)},true};
}
uint32_t format = cast<uint32_t>((g->Regs[typR] & cast<uint32_t>(15ULL)));
uint32_t addr = n3ds_Machine_gpuAddrToVirt(g->m,shl<uint32_t>(g->Regs[addrR],cast<int64_t>(3ULL)));
uint32_t typ = cast<uint32_t>(0ULL);
if ((u == cast<int64_t>(0ULL))) {
typ = cast<uint32_t>((shr<uint32_t>(g->Regs[cast<int64_t>(131ULL)],cast<int64_t>(28ULL)) & cast<uint32_t>(7ULL)));
}
float shadowZ={};
{
switch(typ){
case cast<uint32_t>(2ULL):{
if (((cast<uint32_t>((g->Regs[cast<int64_t>(139ULL)] & cast<uint32_t>(1ULL))) == cast<uint32_t>(0ULL)) && (q != cast<float>(0.00000000000000000e+00)))) {
auto tmp259 = std::make_tuple(cast<float>((s / q)),cast<float>((t / q)));
s = std::get<0>(tmp259);
t = std::get<1>(tmp259);
}
shadowZ = n3ds_absf(q);
break;}
case cast<uint32_t>(3ULL):{
if ((q != cast<float>(0.00000000000000000e+00))) {
auto tmp260 = std::make_tuple(cast<float>((s / q)),cast<float>((t / q)));
s = std::get<0>(tmp260);
t = std::get<1>(tmp260);
}
break;}
}}
auto tmp261 = n3ds_GPU_texture(g,addr,format,w,h);
n3ds_texImage* img = std::get<0>(tmp261);
bool ok = std::get<1>(tmp261);
if ((!ok)) {
return {n3ds_rgba{},false};
}
uint32_t param = g->Regs[paramR];
int32_t x = n3ds_wrapCoord(s,cast<int32_t>(w),cast<uint32_t>((shr<uint32_t>(param,cast<int64_t>(12ULL)) & cast<uint32_t>(7ULL))));
int32_t v = n3ds_wrapCoord(t,cast<int32_t>(h),cast<uint32_t>((shr<uint32_t>(param,cast<int64_t>(8ULL)) & cast<uint32_t>(7ULL))));
if (((x < cast<int32_t>(0ULL)) || (v < cast<int32_t>(0ULL)))) {
return {n3ds_rgba{},true};
}
int32_t y = cast<int32_t>((cast<int32_t>((cast<int32_t>(h) - cast<int32_t>(1ULL))) - v));
uint32_t p = cast<uint32_t>(((cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(y) * w)) + cast<uint32_t>(x)))) * cast<uint32_t>(4ULL)));
n3ds_rgba c = n3ds_rgba{cast<int32_t>(img->pix[p]),cast<int32_t>(img->pix[cast<uint32_t>((p + cast<uint32_t>(1ULL)))]),cast<int32_t>(img->pix[cast<uint32_t>((p + cast<uint32_t>(2ULL)))]),cast<int32_t>(img->pix[cast<uint32_t>((p + cast<uint32_t>(3ULL)))])};
if ((typ == cast<uint32_t>(2ULL))) {
c = n3ds_shadowCompare(c,shadowZ,g->Regs[cast<int64_t>(139ULL)]);
st->shadowSamples++;
if ((c.r != cast<int32_t>(0ULL))) {
st->shadowOccluded++;
}
}
return {c,true};
}
}
// tools/platform/n3ds/gpu_texture.go:128:1
n3ds_rgba n3ds_shadowCompare(n3ds_rgba c,float z,uint32_t cfg){
{
int32_t zi = cast<int32_t>(cast<float>((n3ds_clampf(z,cast<float>(0.00000000000000000e+00),cast<float>(1.00000000000000000e+00)) * cast<float>(1.67772150000000000e+07))));
zi -= shl<int32_t>(cast<int32_t>(cast<uint32_t>((shr<uint32_t>(cfg,cast<int64_t>(1ULL)) & cast<uint32_t>(8388607ULL)))),cast<int64_t>(1ULL));
int32_t zref = cast<int32_t>((cast<int32_t>((shl<int32_t>(c.a,cast<int64_t>(16ULL)) | shl<int32_t>(c.b,cast<int64_t>(8ULL)))) | c.g));
int32_t d = cast<int32_t>(0ULL);
if ((zref >= zi)) {
d = c.r;
}
return n3ds_rgba{d,d,d,d};
}
}
// tools/platform/n3ds/gpu_texture.go:139:1
int32_t n3ds_wrapCoord(float f,int32_t n,uint32_t mode){
{
int32_t v = cast<int32_t>(n3ds_floor32(cast<float>((f * cast<float>(n)))));
{
switch(mode){
case cast<uint32_t>(0ULL):{
if ((v < cast<int32_t>(0ULL))) {
v = cast<int32_t>(0ULL);
}
if ((v >= n)) {
v = cast<int32_t>((n - cast<int32_t>(1ULL)));
}
break;}
case cast<uint32_t>(1ULL):{
if (((v < cast<int32_t>(0ULL)) || (v >= n))) {
return cast<int32_t>(-1ULL);
}
break;}
case cast<uint32_t>(2ULL):{
v %= n;
if ((v < cast<int32_t>(0ULL))) {
v += n;
}
break;}
default:{
int32_t period = cast<int32_t>((cast<int32_t>(2ULL) * n));
v %= period;
if ((v < cast<int32_t>(0ULL))) {
v += period;
}
if ((v >= n)) {
v = cast<int32_t>((cast<int32_t>((period - cast<int32_t>(1ULL))) - v));
}
break;}
}}
return v;
}
}
// tools/platform/n3ds/gpu_texture.go:174:1
std::tuple<n3ds_texImage*,bool> n3ds_GPU_texture(n3ds_GPU* g,uint32_t addr,uint32_t format,uint32_t w,uint32_t h){
{
n3ds_texKey k = n3ds_texKey{addr,format,w,h};
{
auto tmp262 = lookup(g->texCache,k);
n3ds_texImage* img = std::get<0>(tmp262);
bool hit = std::get<1>(tmp262);
if (hit) {
return {img,true};
}
}
time_Time t = n3ds_Machine_profStart(g->m);
auto tmp263=defer([&](){n3ds_Machine_profEnd(g->m,cast<int64_t>(3ULL),t);});
n3ds_texImage* img = rrNewTexture(n3ds_texImage{w,h,Slice<uint8_t>::make(cast<uint32_t>((cast<uint32_t>((w * h)) * cast<uint32_t>(4ULL))))});
auto put = [&](uint32_t x,uint32_t y,uint8_t r,uint8_t gr,uint8_t b,uint8_t a)->void{
uint32_t p = cast<uint32_t>(((cast<uint32_t>((cast<uint32_t>((y * w)) + x))) * cast<uint32_t>(4ULL)));
auto tmp264 = std::make_tuple(r,gr,b,a);
img->pix[p] = std::get<0>(tmp264);
img->pix[cast<uint32_t>((p + cast<uint32_t>(1ULL)))] = std::get<1>(tmp264);
img->pix[cast<uint32_t>((p + cast<uint32_t>(2ULL)))] = std::get<2>(tmp264);
img->pix[cast<uint32_t>((p + cast<uint32_t>(3ULL)))] = std::get<3>(tmp264);
}
;
{
switch(format){
case cast<uint32_t>(0ULL):{
n3ds_GPU_eachTexel(g,addr,w,h,cast<uint32_t>(4ULL),[&](uint32_t x,uint32_t y,uint32_t p)->void{
put(x,y,n3ds_Machine_Read(g->m,cast<uint32_t>((p + cast<uint32_t>(3ULL)))),n3ds_Machine_Read(g->m,cast<uint32_t>((p + cast<uint32_t>(2ULL)))),n3ds_Machine_Read(g->m,cast<uint32_t>((p + cast<uint32_t>(1ULL)))),n3ds_Machine_Read(g->m,p));
}
);
break;}
case cast<uint32_t>(1ULL):{
n3ds_GPU_eachTexel(g,addr,w,h,cast<uint32_t>(3ULL),[&](uint32_t x,uint32_t y,uint32_t p)->void{
put(x,y,n3ds_Machine_Read(g->m,cast<uint32_t>((p + cast<uint32_t>(2ULL)))),n3ds_Machine_Read(g->m,cast<uint32_t>((p + cast<uint32_t>(1ULL)))),n3ds_Machine_Read(g->m,p),cast<uint8_t>(255ULL));
}
);
break;}
case cast<uint32_t>(2ULL):{
n3ds_GPU_eachTexel(g,addr,w,h,cast<uint32_t>(2ULL),[&](uint32_t x,uint32_t y,uint32_t p)->void{
uint16_t v = cast<uint16_t>((cast<uint16_t>(n3ds_Machine_Read(g->m,p)) | shl<uint16_t>(cast<uint16_t>(n3ds_Machine_Read(g->m,cast<uint32_t>((p + cast<uint32_t>(1ULL))))),cast<int64_t>(8ULL))));
uint8_t r = cast<uint8_t>((cast<uint8_t>(shr<uint16_t>(v,cast<int64_t>(11ULL))) & cast<uint8_t>(31ULL)));
uint8_t gr = cast<uint8_t>((cast<uint8_t>(shr<uint16_t>(v,cast<int64_t>(6ULL))) & cast<uint8_t>(31ULL)));
uint8_t b = cast<uint8_t>((cast<uint8_t>(shr<uint16_t>(v,cast<int64_t>(1ULL))) & cast<uint8_t>(31ULL)));
put(x,y,cast<uint8_t>((shl<uint8_t>(r,cast<int64_t>(3ULL)) | shr<uint8_t>(r,cast<int64_t>(2ULL)))),cast<uint8_t>((shl<uint8_t>(gr,cast<int64_t>(3ULL)) | shr<uint8_t>(gr,cast<int64_t>(2ULL)))),cast<uint8_t>((shl<uint8_t>(b,cast<int64_t>(3ULL)) | shr<uint8_t>(b,cast<int64_t>(2ULL)))),cast<uint8_t>((cast<uint8_t>(cast<uint16_t>((v & cast<uint16_t>(1ULL)))) * cast<uint8_t>(255ULL))));
}
);
break;}
case cast<uint32_t>(3ULL):{
n3ds_GPU_eachTexel(g,addr,w,h,cast<uint32_t>(2ULL),[&](uint32_t x,uint32_t y,uint32_t p)->void{
uint16_t v = cast<uint16_t>((cast<uint16_t>(n3ds_Machine_Read(g->m,p)) | shl<uint16_t>(cast<uint16_t>(n3ds_Machine_Read(g->m,cast<uint32_t>((p + cast<uint32_t>(1ULL))))),cast<int64_t>(8ULL))));
uint8_t r = cast<uint8_t>((cast<uint8_t>(shr<uint16_t>(v,cast<int64_t>(11ULL))) & cast<uint8_t>(31ULL)));
uint8_t gr = cast<uint8_t>((cast<uint8_t>(shr<uint16_t>(v,cast<int64_t>(5ULL))) & cast<uint8_t>(63ULL)));
uint8_t b = cast<uint8_t>((cast<uint8_t>(v) & cast<uint8_t>(31ULL)));
put(x,y,cast<uint8_t>((shl<uint8_t>(r,cast<int64_t>(3ULL)) | shr<uint8_t>(r,cast<int64_t>(2ULL)))),cast<uint8_t>((shl<uint8_t>(gr,cast<int64_t>(2ULL)) | shr<uint8_t>(gr,cast<int64_t>(4ULL)))),cast<uint8_t>((shl<uint8_t>(b,cast<int64_t>(3ULL)) | shr<uint8_t>(b,cast<int64_t>(2ULL)))),cast<uint8_t>(255ULL));
}
);
break;}
case cast<uint32_t>(4ULL):{
n3ds_GPU_eachTexel(g,addr,w,h,cast<uint32_t>(2ULL),[&](uint32_t x,uint32_t y,uint32_t p)->void{
uint16_t v = cast<uint16_t>((cast<uint16_t>(n3ds_Machine_Read(g->m,p)) | shl<uint16_t>(cast<uint16_t>(n3ds_Machine_Read(g->m,cast<uint32_t>((p + cast<uint32_t>(1ULL))))),cast<int64_t>(8ULL))));
put(x,y,cast<uint8_t>((cast<uint8_t>(shr<uint16_t>(v,cast<int64_t>(12ULL))) * cast<uint8_t>(17ULL))),cast<uint8_t>((cast<uint8_t>(cast<uint16_t>((shr<uint16_t>(v,cast<int64_t>(8ULL)) & cast<uint16_t>(15ULL)))) * cast<uint8_t>(17ULL))),cast<uint8_t>((cast<uint8_t>(cast<uint16_t>((shr<uint16_t>(v,cast<int64_t>(4ULL)) & cast<uint16_t>(15ULL)))) * cast<uint8_t>(17ULL))),cast<uint8_t>((cast<uint8_t>(cast<uint16_t>((v & cast<uint16_t>(15ULL)))) * cast<uint8_t>(17ULL))));
}
);
break;}
case cast<uint32_t>(5ULL):{
n3ds_GPU_eachTexel(g,addr,w,h,cast<uint32_t>(2ULL),[&](uint32_t x,uint32_t y,uint32_t p)->void{
auto tmp265 = std::make_tuple(n3ds_Machine_Read(g->m,cast<uint32_t>((p + cast<uint32_t>(1ULL)))),n3ds_Machine_Read(g->m,p));
uint8_t l = std::get<0>(tmp265);
uint8_t a = std::get<1>(tmp265);
put(x,y,l,l,l,a);
}
);
break;}
case cast<uint32_t>(6ULL):{
n3ds_GPU_eachTexel(g,addr,w,h,cast<uint32_t>(2ULL),[&](uint32_t x,uint32_t y,uint32_t p)->void{
put(x,y,n3ds_Machine_Read(g->m,cast<uint32_t>((p + cast<uint32_t>(1ULL)))),n3ds_Machine_Read(g->m,p),cast<uint8_t>(0ULL),cast<uint8_t>(255ULL));
}
);
break;}
case cast<uint32_t>(7ULL):{
n3ds_GPU_eachTexel(g,addr,w,h,cast<uint32_t>(1ULL),[&](uint32_t x,uint32_t y,uint32_t p)->void{
uint8_t l = n3ds_Machine_Read(g->m,p);
put(x,y,l,l,l,cast<uint8_t>(255ULL));
}
);
break;}
case cast<uint32_t>(8ULL):{
n3ds_GPU_eachTexel(g,addr,w,h,cast<uint32_t>(1ULL),[&](uint32_t x,uint32_t y,uint32_t p)->void{
put(x,y,cast<uint8_t>(255ULL),cast<uint8_t>(255ULL),cast<uint8_t>(255ULL),n3ds_Machine_Read(g->m,p));
}
);
break;}
case cast<uint32_t>(9ULL):{
n3ds_GPU_eachTexel(g,addr,w,h,cast<uint32_t>(1ULL),[&](uint32_t x,uint32_t y,uint32_t p)->void{
uint8_t v = n3ds_Machine_Read(g->m,p);
uint8_t l = cast<uint8_t>((shr<uint8_t>(v,cast<int64_t>(4ULL)) * cast<uint8_t>(17ULL)));
put(x,y,l,l,l,cast<uint8_t>((cast<uint8_t>((v & cast<uint8_t>(15ULL))) * cast<uint8_t>(17ULL))));
}
);
break;}
case cast<uint32_t>(10ULL):{
n3ds_GPU_decode4(g,addr,w,h,img,[&](uint8_t n)->std::array<uint8_t,4>{
return std::array<uint8_t,4>{cast<uint8_t>((n * cast<uint8_t>(17ULL))),cast<uint8_t>((n * cast<uint8_t>(17ULL))),cast<uint8_t>((n * cast<uint8_t>(17ULL))),cast<uint8_t>(255ULL)};
}
);
break;}
case cast<uint32_t>(11ULL):{
n3ds_GPU_decode4(g,addr,w,h,img,[&](uint8_t n)->std::array<uint8_t,4>{
return std::array<uint8_t,4>{cast<uint8_t>(255ULL),cast<uint8_t>(255ULL),cast<uint8_t>(255ULL),cast<uint8_t>((n * cast<uint8_t>(17ULL)))};
}
);
break;}
case cast<uint32_t>(12ULL):{
uint32_t size = divi<uint32_t>(cast<uint32_t>((w * h)),cast<uint32_t>(2ULL));
Slice<uint8_t> data = Slice<uint8_t>::make(size);
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < size);i++){
data[i] = n3ds_Machine_Read(g->m,cast<uint32_t>((addr + i)));
}
}gcopy(img->pix,rrDecodeETC(data,w,h,false));
break;}
case cast<uint32_t>(13ULL):{
uint32_t size = cast<uint32_t>((w * h));
Slice<uint8_t> data = Slice<uint8_t>::make(size);
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < size);i++){
data[i] = n3ds_Machine_Read(g->m,cast<uint32_t>((addr + i)));
}
}gcopy(img->pix,rrDecodeETC(data,w,h,true));
break;}
default:{
arm_CPU_Halt(g->m->CPU,std::string("gpu: texture format 0x%X unimplemented (%dx%d at 0x%08X)",56),format,w,h,addr);
return {{},false};
break;}
}}
if ((!g->texCache)) {
g->texCache = Map<n3ds_texKey,n3ds_texImage*>{};
}
g->texCache[k] = img;
return {img,true};
}
}
// tools/platform/n3ds/gpu_texture.go:276:1
void n3ds_GPU_eachTexel(n3ds_GPU* g,uint32_t addr,uint32_t w,uint32_t h,uint32_t bpp,std::function<void(uint32_t,uint32_t,uint32_t)> visit){
{
uint32_t p = addr;
{uint32_t ty = cast<uint32_t>(0ULL);for (;(ty < h);ty += cast<uint32_t>(8ULL)){
{uint32_t tx = cast<uint32_t>(0ULL);for (;(tx < w);tx += cast<uint32_t>(8ULL)){
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < cast<uint32_t>(64ULL));i++){
uint32_t x = cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((i & cast<uint32_t>(1ULL))) | cast<uint32_t>((shr<uint32_t>(i,cast<int64_t>(1ULL)) & cast<uint32_t>(2ULL))))) | cast<uint32_t>((shr<uint32_t>(i,cast<int64_t>(2ULL)) & cast<uint32_t>(4ULL)))));
uint32_t y = cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((shr<uint32_t>(i,cast<int64_t>(1ULL)) & cast<uint32_t>(1ULL))) | cast<uint32_t>((shr<uint32_t>(i,cast<int64_t>(2ULL)) & cast<uint32_t>(2ULL))))) | cast<uint32_t>((shr<uint32_t>(i,cast<int64_t>(3ULL)) & cast<uint32_t>(4ULL)))));
visit(cast<uint32_t>((tx + x)),cast<uint32_t>((ty + y)),p);
p += bpp;
}
}}
}}
}}
}
// tools/platform/n3ds/gpu_texture.go:292:1
void n3ds_GPU_decode4(n3ds_GPU* g,uint32_t addr,uint32_t w,uint32_t h,n3ds_texImage* img,std::function<std::array<uint8_t,4>(uint8_t)> px){
{
uint32_t p = addr;
bool half = false;
uint8_t hold={};
{uint32_t ty = cast<uint32_t>(0ULL);for (;(ty < h);ty += cast<uint32_t>(8ULL)){
{uint32_t tx = cast<uint32_t>(0ULL);for (;(tx < w);tx += cast<uint32_t>(8ULL)){
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < cast<uint32_t>(64ULL));i++){
uint32_t x = cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((i & cast<uint32_t>(1ULL))) | cast<uint32_t>((shr<uint32_t>(i,cast<int64_t>(1ULL)) & cast<uint32_t>(2ULL))))) | cast<uint32_t>((shr<uint32_t>(i,cast<int64_t>(2ULL)) & cast<uint32_t>(4ULL)))));
uint32_t y = cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((shr<uint32_t>(i,cast<int64_t>(1ULL)) & cast<uint32_t>(1ULL))) | cast<uint32_t>((shr<uint32_t>(i,cast<int64_t>(2ULL)) & cast<uint32_t>(2ULL))))) | cast<uint32_t>((shr<uint32_t>(i,cast<int64_t>(3ULL)) & cast<uint32_t>(4ULL)))));
uint8_t n={};
if ((!half)) {
hold = n3ds_Machine_Read(g->m,p);
p++;
n = cast<uint8_t>((hold & cast<uint8_t>(15ULL)));
}
else {
n = shr<uint8_t>(hold,cast<int64_t>(4ULL));
}
half = (!half);
std::array<uint8_t,4> c = px(n);
uint32_t q = cast<uint32_t>(((cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(((cast<uint32_t>((ty + y))) * w)) + tx)) + x))) * cast<uint32_t>(4ULL)));
gcopy(sub(img->pix,q,cast<uint32_t>((q + cast<uint32_t>(4ULL)))),sub(c,0,len(c)));
}
}}
}}
}}
}
// tools/platform/n3ds/gpu_texture.go:324:1
uint32_t n3ds_texBytes(n3ds_texKey k){
{
uint32_t bpp={};
{
switch(k.fmt){
case cast<uint32_t>(0ULL):{
bpp = cast<uint32_t>(32ULL);
break;}
case cast<uint32_t>(1ULL):{
bpp = cast<uint32_t>(24ULL);
break;}
case cast<uint32_t>(2ULL):case cast<uint32_t>(3ULL):case cast<uint32_t>(4ULL):case cast<uint32_t>(5ULL):case cast<uint32_t>(6ULL):{
bpp = cast<uint32_t>(16ULL);
break;}
case cast<uint32_t>(7ULL):case cast<uint32_t>(8ULL):case cast<uint32_t>(9ULL):case cast<uint32_t>(13ULL):{
bpp = cast<uint32_t>(8ULL);
break;}
default:{
bpp = cast<uint32_t>(8ULL);
break;}
}}
return divi<uint32_t>(cast<uint32_t>((cast<uint32_t>((k.w * k.h)) * bpp)),cast<uint32_t>(8ULL));
}
}
// tools/platform/n3ds/gpu_texture.go:351:1
void n3ds_GPU_invalidateTextures(n3ds_GPU* g,uint32_t addr,uint32_t size){
{
// Erase through the returned iterator: C++ unordered_map erasure invalidates
// the current iterator, unlike deletion during a Go map range.
if(!size)return;
for(auto it=g->texCache.p->begin();it!=g->texCache.p->end();){
 auto k=it->first;
 if(uint64_t(k.addr)<uint64_t(addr)+size&&uint64_t(addr)<uint64_t(k.addr)+n3ds_texBytes(k))it=g->texCache.p->erase(it);
 else ++it;
}
}}
// tools/platform/n3ds/gsp_mem.go:45:1
uint32_t n3ds_Machine_gpuAddrToVirt(n3ds_Machine* m,uint32_t a){
{
{
if (((a >= cast<uint32_t>(402653184ULL)) && (a < cast<uint32_t>(408944640ULL)))){
return cast<uint32_t>((cast<uint32_t>((a - cast<uint32_t>(402653184ULL))) + cast<uint32_t>(520093696ULL)));
}
else if (((a >= cast<uint32_t>(536870912ULL)) && (a < cast<uint32_t>(671088640ULL)))){
return cast<uint32_t>((a - cast<uint32_t>(201326592ULL)));
}
}
tmp270:;
return a;
}
}
// tools/platform/n3ds/gsp_mem.go:60:1
void n3ds_Machine_svcMapMemoryBlock(n3ds_Machine* m,arm_CPU* c){
{
auto tmp271 = std::make_tuple(c->R[cast<int64_t>(0ULL)],c->R[cast<int64_t>(1ULL)]);
uint32_t handle = std::get<0>(tmp271);
uint32_t addr = std::get<1>(tmp271);
n3ds_kobject* obj = get(m->handles,handle);
if ((!obj)) {
arm_CPU_Halt(c,std::string("MapMemoryBlock: unknown handle 0x%08X at 0x%08X after %d instructions",69),handle,arm_CPU_PC(c),c->Instrs);
return ;
}
uint32_t size = obj->blockSize;
if ((size == cast<uint32_t>(0ULL))) {
size = cast<uint32_t>(4096ULL);
}
if ((obj->blockAddr != cast<uint32_t>(0ULL))) {
{
n3ds_memRegion* r = n3ds_Machine_regionOf(m,obj->blockAddr);
if (bool(r)) {
n3ds_Machine_mapRegion(m,obj->kind,addr,r->data);
}
}
}
else {
n3ds_Machine_mapRegion(m,obj->kind,addr,Slice<uint8_t>::make(size));
}
obj->blockAddr = addr;
obj->blockSize = size;
if ((obj->kind == std::string("gsp-shared",10))) {
m->gspSharedAddr = addr;
}
if ((obj->kind == std::string("hid-shared",10))) {
m->hidSharedAddr = addr;
}
if (m->Verbose) {
go_fmt_Printf(std::string("  MapMemoryBlock handle=0x%08X (%s) -> 0x%08X size=0x%X\012",56),handle,obj->kind,addr,size);
}
c->R[cast<int64_t>(0ULL)] = cast<uint32_t>(0ULL);
}
}
// tools/platform/n3ds/gsp_mem.go:99:1
void n3ds_Machine_svcUnmapMemoryBlock(n3ds_Machine* m,arm_CPU* c){
{
auto tmp272 = std::make_tuple(c->R[cast<int64_t>(0ULL)],c->R[cast<int64_t>(1ULL)]);
uint32_t handle = std::get<0>(tmp272);
uint32_t addr = std::get<1>(tmp272);
n3ds_kobject* obj = get(m->handles,handle);
if ((!obj)) {
arm_CPU_Halt(c,std::string("UnmapMemoryBlock: unknown handle 0x%08X at 0x%08X after %d instructions",71),handle,arm_CPU_PC(c),c->Instrs);
return ;
}
if ((obj->blockAddr != addr)) {
arm_CPU_Halt(c,std::string("UnmapMemoryBlock: handle 0x%08X (%s) mapped at 0x%08X, not 0x%08X, at 0x%08X",76),handle,obj->kind,obj->blockAddr,addr,arm_CPU_PC(c));
return ;
}
{auto&& tmp273 = m->regions;
for(int64_t tmp274=0;tmp274<len(tmp273);++tmp274){
auto i=tmp274;auto r=tmp273[tmp274];if ((r->base == addr)) {
m->regions = append(sub(m->regions,0,i),sub(m->regions,cast<int64_t>((i + cast<int64_t>(1ULL))),len(m->regions)));
break;
}
}}
obj->blockAddr = cast<uint32_t>(0ULL);
if (m->Verbose) {
go_fmt_Printf(std::string("  UnmapMemoryBlock handle=0x%08X (%s) from 0x%08X\012",50),handle,obj->kind,addr);
}
c->R[cast<int64_t>(0ULL)] = cast<uint32_t>(0ULL);
}
}
// tools/platform/n3ds/gsp_vblank.go:36:1
bool n3ds_Machine_vblankDue(n3ds_Machine* m){
{
return ((m->gspEvent != cast<uint32_t>(0ULL)) && (m->instrs >= m->nextFrameInstr));
}
}
// tools/platform/n3ds/gsp_vblank.go:61:1
n3ds_FBPresent n3ds_Machine_Scanout(n3ds_Machine* m,int64_t screen){
{
if (((screen < cast<int64_t>(0ULL)) || (screen > cast<int64_t>(1ULL)))) {
return n3ds_FBPresent{};
}
return m->screenFB[screen];
}
}
// tools/platform/n3ds/gsp_vblank.go:70:1
void n3ds_Machine_deliverVBlank(n3ds_Machine* m){
{
m->vblankCount++;
m->nextFrameInstr = cast<uint64_t>((m->instrs + cast<uint64_t>(4468530ULL)));
n3ds_Machine_consumeFBInfo(m);
n3ds_Machine_pushGSPInterrupt(m,cast<uint8_t>(2ULL));
n3ds_Machine_pushGSPInterrupt(m,cast<uint8_t>(3ULL));
n3ds_Machine_signalGSPEvent(m);
n3ds_Machine_updateHIDShared(m);
if (m->aptWakePending) {
m->aptWakePending = false;
n3ds_Machine_signalAPTEvents(m);
}
n3ds_Machine_profFrame(m);
if (bool(m->OnFrame)) {
m->OnFrame(m);
}
}
}
// tools/platform/n3ds/gsp_vblank.go:115:1
void n3ds_Machine_consumeFBInfo(n3ds_Machine* m){
{
if ((m->gspSharedAddr == cast<uint32_t>(0ULL))) {
return ;
}
{uint32_t screen = cast<uint32_t>(0ULL);for (;(screen < cast<uint32_t>(2ULL));screen++){
uint32_t base = cast<uint32_t>((cast<uint32_t>((m->gspSharedAddr + cast<uint32_t>(512ULL))) + cast<uint32_t>((screen * cast<uint32_t>(64ULL)))));
if ((n3ds_Machine_Read(m,cast<uint32_t>((base + cast<uint32_t>(1ULL)))) == cast<uint8_t>(0ULL))) {
continue;
}
uint32_t idx = cast<uint32_t>((cast<uint32_t>(n3ds_Machine_Read(m,base)) & cast<uint32_t>(1ULL)));
uint32_t e = cast<uint32_t>((cast<uint32_t>((base + cast<uint32_t>(4ULL))) + cast<uint32_t>((idx * cast<uint32_t>(28ULL)))));
m->screenFB[screen] = n3ds_FBPresent{n3ds_Machine_ReadWord(m,e),n3ds_Machine_ReadWord(m,cast<uint32_t>((e + cast<uint32_t>(4ULL)))),n3ds_Machine_ReadWord(m,cast<uint32_t>((e + cast<uint32_t>(8ULL)))),n3ds_Machine_ReadWord(m,cast<uint32_t>((e + cast<uint32_t>(12ULL)))),n3ds_Machine_ReadWord(m,cast<uint32_t>((e + cast<uint32_t>(16ULL)))),n3ds_Machine_ReadWord(m,cast<uint32_t>((e + cast<uint32_t>(20ULL)))),true};
n3ds_Machine_Write(m,cast<uint32_t>((base + cast<uint32_t>(1ULL))),cast<uint8_t>(0ULL));
m->framesSwapped++;
}
}}
}
// tools/platform/n3ds/gsp_vblank.go:142:1
void n3ds_Machine_signalGSPEvent(n3ds_Machine* m){
{
{
n3ds_kobject* obj = get(m->handles,m->gspEvent);
if (bool(obj)) {
obj->signal = true;
if (n3ds_Machine_signalObject(m,obj)) {
m->reschedule = true;
}
}
}
}
}
// tools/platform/n3ds/gsp_vblank.go:156:1
void n3ds_Machine_pushGSPInterrupt(n3ds_Machine* m,uint8_t id){
{
if ((m->gspSharedAddr == cast<uint32_t>(0ULL))) {
return ;
}
uint32_t base = m->gspSharedAddr;
uint8_t idx = n3ds_Machine_Read(m,cast<uint32_t>((base + cast<uint32_t>(0ULL))));
uint8_t cnt = n3ds_Machine_Read(m,cast<uint32_t>((base + cast<uint32_t>(1ULL))));
constexpr int64_t listLen=52ULL;
if ((cast<uint32_t>(cnt) >= cast<uint32_t>(52ULL))) {
if ((!m->gspOverflowed)) {
m->gspOverflowed = true;
go_fmt_Printf((std::string("GSP interrupt queue overflow (id=%d idx=%d cnt=%d at instr %d): the app has stopped ",84) + std::string("draining it \342\200\224 interrupts are being dropped, and something has blocked its GSP thread\012",87)),id,idx,cnt,m->instrs);
}
return ;
}
uint32_t pos = modi<uint32_t>((cast<uint32_t>((cast<uint32_t>(idx) + cast<uint32_t>(cnt)))),cast<uint32_t>(52ULL));
n3ds_Machine_Write(m,cast<uint32_t>((cast<uint32_t>((base + cast<uint32_t>(12ULL))) + pos)),id);
n3ds_Machine_Write(m,cast<uint32_t>((base + cast<uint32_t>(1ULL))),cast<uint8_t>((cnt + cast<uint8_t>(1ULL))));
}
}
// tools/platform/n3ds/gsp_vblank.go:187:1
uint64_t n3ds_Machine_VBlanks(n3ds_Machine* m){
{
return m->vblankCount;
}
}
// tools/platform/n3ds/gx.go:53:1
Slice<n3ds_GXRecord> n3ds_Machine_GXLog(n3ds_Machine* m){
{
return m->gxLog;
}
}
// tools/platform/n3ds/gx.go:58:1
void n3ds_Machine_captureChainedList(n3ds_Machine* m,uint32_t addr,uint32_t size,Slice<uint8_t> buf){
{
m->gxLog = append(m->gxLog,n3ds_GXRecord{m->CPU->Instrs,std::array<uint32_t,8>{cast<uint32_t>(1ULL),addr,size},append(Slice<uint8_t>{},buf),true});
}
}
// tools/platform/n3ds/gx.go:71:1
void n3ds_Machine_captureGX(n3ds_Machine* m,uint32_t cmd,uint32_t id){
{
n3ds_GXRecord r={};
r.Instr = m->CPU->Instrs;
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < cast<uint32_t>(8ULL));i++){
r.Words[i] = n3ds_Machine_ReadWord(m,cast<uint32_t>((cmd + cast<uint32_t>((i * cast<uint32_t>(4ULL))))));
}
}if ((id == cast<uint32_t>(1ULL))) {
auto tmp275 = std::make_tuple(r.Words[cast<int64_t>(1ULL)],r.Words[cast<int64_t>(2ULL)]);
uint32_t addr = std::get<0>(tmp275);
uint32_t size = std::get<1>(tmp275);
if (((size > cast<uint32_t>(0ULL)) && (size < cast<uint32_t>(16777216ULL)))) {
r.Buf = Slice<uint8_t>::make(size);
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < size);i++){
r.Buf[i] = n3ds_Machine_Read(m,cast<uint32_t>((addr + i)));
}
}}
}
m->gxLog = append(m->gxLog,r);
}
}
// tools/platform/n3ds/gx.go:93:1
void n3ds_Machine_gxMemoryFill(n3ds_Machine* m,uint32_t start,uint32_t value,uint32_t end,uint32_t ctl){
rrprof::Scope profile(3,"GX memory / display transfers");rr3ds::Command command(m,"GX memory fill",start,value,end,ctl);auto graphics=rrgpu::Operation::fill(m,start,value,end,ctl);if(graphics.execute()){n3ds_GPU_invalidateTextures(m->gpu,graphics.dst,graphics.size);return;}{
auto tmp276=defer([&](){n3ds_Machine_profEnd(m,cast<int64_t>(4ULL),n3ds_Machine_profStart(m));});
auto tmp277 = std::make_tuple(n3ds_Machine_gpuAddrToVirt(m,start),n3ds_Machine_gpuAddrToVirt(m,end));
start = std::get<0>(tmp277);
end = std::get<1>(tmp277);
uint32_t unit={};
{
switch(cast<uint32_t>((shr<uint32_t>(ctl,cast<int64_t>(8ULL)) & cast<uint32_t>(3ULL)))){
case cast<uint32_t>(0ULL):{
unit = cast<uint32_t>(2ULL);
break;}
case cast<uint32_t>(1ULL):{
unit = cast<uint32_t>(3ULL);
break;}
case cast<uint32_t>(2ULL):{
unit = cast<uint32_t>(4ULL);
break;}
default:{
arm_CPU_Halt(m->CPU,std::string("gx: MemoryFill control 0x%04X has width code 3 after %d instructions",68),ctl,m->CPU->Instrs);
return ;
break;}
}}
{uint32_t a = start;for (;(cast<uint32_t>((a + unit)) <= end);a += unit){
{uint32_t j = cast<uint32_t>(0ULL);for (;(j < unit);j++){
n3ds_Machine_Write(m,cast<uint32_t>((a + j)),cast<uint8_t>(shr<uint32_t>(value,(cast<uint32_t>((cast<uint32_t>(8ULL) * j))))));
}
}}
}n3ds_GPU_invalidateTextures(m->gpu,start,cast<uint32_t>((end - start)));
}
}
// tools/platform/n3ds/gx.go:119:1
void n3ds_Machine_gxTextureCopy(n3ds_Machine* m,uint32_t src,uint32_t dst,uint32_t size,uint32_t inDim,uint32_t outDim){
rrprof::Scope profile(3,"GX memory / display transfers");rr3ds::Command command(m,"GX texture copy",src,dst,size,inDim);auto graphics=rrgpu::Operation::copy(m,src,dst,size,inDim,outDim);if(graphics.execute()){n3ds_GPU_invalidateTextures(m->gpu,graphics.dst,graphics.size);return;}{
auto tmp278=defer([&](){n3ds_Machine_profEnd(m,cast<int64_t>(4ULL),n3ds_Machine_profStart(m));});
auto tmp279 = std::make_tuple(n3ds_Machine_gpuAddrToVirt(m,src),n3ds_Machine_gpuAddrToVirt(m,dst));
src = std::get<0>(tmp279);
dst = std::get<1>(tmp279);
auto tmp280 = std::make_tuple(cast<uint32_t>((cast<uint32_t>((inDim & cast<uint32_t>(65535ULL))) * cast<uint32_t>(2ULL))),cast<uint32_t>((shr<uint32_t>(inDim,cast<int64_t>(16ULL)) * cast<uint32_t>(2ULL))));
uint32_t inW = std::get<0>(tmp280);
uint32_t inGap = std::get<1>(tmp280);
auto tmp281 = std::make_tuple(cast<uint32_t>((cast<uint32_t>((outDim & cast<uint32_t>(65535ULL))) * cast<uint32_t>(2ULL))),cast<uint32_t>((shr<uint32_t>(outDim,cast<int64_t>(16ULL)) * cast<uint32_t>(2ULL))));
uint32_t outW = std::get<0>(tmp281);
uint32_t outGap = std::get<1>(tmp281);
if (((inW == cast<uint32_t>(0ULL)) && (inGap == cast<uint32_t>(0ULL)))) {
inW = size;
}
if (((outW == cast<uint32_t>(0ULL)) && (outGap == cast<uint32_t>(0ULL)))) {
outW = size;
}
if ((((inW == cast<uint32_t>(0ULL)) || (outW == cast<uint32_t>(0ULL))) || (modi<uint32_t>(size,inW) != cast<uint32_t>(0ULL)))) {
arm_CPU_Halt(m->CPU,std::string("gx: TextureCopy dims in=0x%08X out=0x%08X size=0x%X don't divide after %d instructions",86),inDim,outDim,size,m->CPU->Instrs);
return ;
}
auto tmp282 = std::make_tuple(src,dst);
uint32_t sp = std::get<0>(tmp282);
uint32_t dp = std::get<1>(tmp282);
auto tmp283 = std::make_tuple(cast<uint32_t>(0ULL),cast<uint32_t>(0ULL));
uint32_t sn = std::get<0>(tmp283);
uint32_t dn = std::get<1>(tmp283);
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < size);i++){
n3ds_Machine_Write(m,dp,n3ds_Machine_Read(m,sp));
auto tmp284 = std::make_tuple(cast<uint32_t>((sp + cast<uint32_t>(1ULL))),cast<uint32_t>((sn + cast<uint32_t>(1ULL))));
sp = std::get<0>(tmp284);
sn = std::get<1>(tmp284);
if ((sn == inW)) {
auto tmp285 = std::make_tuple(cast<uint32_t>((sp + inGap)),cast<uint32_t>(0ULL));
sp = std::get<0>(tmp285);
sn = std::get<1>(tmp285);
}
auto tmp286 = std::make_tuple(cast<uint32_t>((dp + cast<uint32_t>(1ULL))),cast<uint32_t>((dn + cast<uint32_t>(1ULL))));
dp = std::get<0>(tmp286);
dn = std::get<1>(tmp286);
if ((dn == outW)) {
auto tmp287 = std::make_tuple(cast<uint32_t>((dp + outGap)),cast<uint32_t>(0ULL));
dp = std::get<0>(tmp287);
dn = std::get<1>(tmp287);
}
}
}n3ds_GPU_invalidateTextures(m->gpu,dst,cast<uint32_t>((dp - dst)));
}
}
// tools/platform/n3ds/gx.go:162:1
void n3ds_Machine_gxDisplayTransfer(n3ds_Machine* m,uint32_t src,uint32_t dst,uint32_t srcDims,uint32_t dstDims,uint32_t flags){
rrprof::Scope profile(3,"GX memory / display transfers");rr3ds::Command command(m,"GX display transfer",src,dst,srcDims,flags);auto graphics=rrgpu::Operation::display(m,src,dst,srcDims,dstDims,flags);{
auto tmp288=defer([&](){n3ds_Machine_profEnd(m,cast<int64_t>(4ULL),n3ds_Machine_profStart(m));});
auto tmp289 = std::make_tuple(cast<uint32_t>((srcDims & cast<uint32_t>(65535ULL))),shr<uint32_t>(srcDims,cast<int64_t>(16ULL)));
uint32_t srcW = std::get<0>(tmp289);
uint32_t srcH = std::get<1>(tmp289);
auto tmp290 = std::make_tuple(cast<uint32_t>((dstDims & cast<uint32_t>(65535ULL))),shr<uint32_t>(dstDims,cast<int64_t>(16ULL)));
uint32_t dstW = std::get<0>(tmp290);
uint32_t dstH = std::get<1>(tmp290);
auto tmp291 = std::make_tuple(cast<uint32_t>((shr<uint32_t>(flags,cast<int64_t>(8ULL)) & cast<uint32_t>(7ULL))),cast<uint32_t>((shr<uint32_t>(flags,cast<int64_t>(12ULL)) & cast<uint32_t>(7ULL))));
uint32_t inFmt = std::get<0>(tmp291);
uint32_t outFmt = std::get<1>(tmp291);
bool flip = (cast<uint32_t>((flags & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL));
if ((((inFmt != cast<uint32_t>(0ULL)) || (cast<uint32_t>((shr<uint32_t>(flags,cast<int64_t>(1ULL)) & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL))) || (cast<uint32_t>((shr<uint32_t>(flags,cast<int64_t>(24ULL)) & cast<uint32_t>(3ULL))) != cast<uint32_t>(0ULL)))) {
arm_CPU_Halt(m->CPU,std::string("gx: DisplayTransfer flags 0x%08X unimplemented (in-format %d, tiled-out %d, scale %d) after %d instructions",107),flags,inFmt,cast<uint32_t>((shr<uint32_t>(flags,cast<int64_t>(1ULL)) & cast<uint32_t>(1ULL))),cast<uint32_t>((shr<uint32_t>(flags,cast<int64_t>(24ULL)) & cast<uint32_t>(3ULL))),m->CPU->Instrs);
return ;
}
auto tmp292 = std::make_tuple(n3ds_Machine_gpuAddrToVirt(m,src),n3ds_Machine_gpuAddrToVirt(m,dst));
src = std::get<0>(tmp292);
dst = std::get<1>(tmp292);
uint32_t dstBPP={};
{
switch(outFmt){
case cast<uint32_t>(0ULL):{
dstBPP = cast<uint32_t>(4ULL);
break;}
case cast<uint32_t>(1ULL):{
dstBPP = cast<uint32_t>(3ULL);
break;}
case cast<uint32_t>(2ULL):case cast<uint32_t>(3ULL):case cast<uint32_t>(4ULL):{
dstBPP = cast<uint32_t>(2ULL);
break;}
default:{
arm_CPU_Halt(m->CPU,std::string("gx: DisplayTransfer out-format %d unimplemented after %d instructions",69),outFmt,m->CPU->Instrs);
return ;
break;}
}}
auto tmp293 = std::make_tuple(srcW,srcH);
uint32_t w = std::get<0>(tmp293);
uint32_t h = std::get<1>(tmp293);
if ((dstW < w)) {
w = dstW;
}
if ((dstH < h)) {
h = dstH;
}
if(!graphics.execute()){
uint32_t tilesPerRow = divi<uint32_t>(srcW,cast<uint32_t>(8ULL));
{uint32_t y = cast<uint32_t>(0ULL);for (;(y < h);y++){
{uint32_t x = cast<uint32_t>(0ULL);for (;(x < w);x++){
uint32_t tile = cast<uint32_t>((cast<uint32_t>(((divi<uint32_t>(y,cast<uint32_t>(8ULL))) * tilesPerRow)) + divi<uint32_t>(x,cast<uint32_t>(8ULL))));
uint32_t mo = cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(((cast<uint32_t>((x & cast<uint32_t>(1ULL)))) & cast<uint32_t>(1ULL))) | shl<uint32_t>((cast<uint32_t>((y & cast<uint32_t>(1ULL)))),cast<int64_t>(1ULL)))) | shl<uint32_t>((cast<uint32_t>((x & cast<uint32_t>(2ULL)))),cast<int64_t>(1ULL)))) | shl<uint32_t>((cast<uint32_t>((y & cast<uint32_t>(2ULL)))),cast<int64_t>(2ULL)))) | shl<uint32_t>((cast<uint32_t>((x & cast<uint32_t>(4ULL)))),cast<int64_t>(2ULL)))) | shl<uint32_t>((cast<uint32_t>((y & cast<uint32_t>(4ULL)))),cast<int64_t>(3ULL))));
uint32_t p = cast<uint32_t>((src + cast<uint32_t>(((cast<uint32_t>((cast<uint32_t>((tile * cast<uint32_t>(64ULL))) + mo))) * cast<uint32_t>(4ULL)))));
auto tmp294 = std::make_tuple(n3ds_Machine_Read(m,p),n3ds_Machine_Read(m,cast<uint32_t>((p + cast<uint32_t>(1ULL)))));
uint8_t a = std::get<0>(tmp294);
uint8_t b = std::get<1>(tmp294);
auto tmp295 = std::make_tuple(n3ds_Machine_Read(m,cast<uint32_t>((p + cast<uint32_t>(2ULL)))),n3ds_Machine_Read(m,cast<uint32_t>((p + cast<uint32_t>(3ULL)))));
uint8_t g = std::get<0>(tmp295);
uint8_t r = std::get<1>(tmp295);
uint32_t dy = y;
if (flip) {
dy = cast<uint32_t>((cast<uint32_t>((h - cast<uint32_t>(1ULL))) - y));
}
uint32_t q = cast<uint32_t>((dst + cast<uint32_t>(((cast<uint32_t>((cast<uint32_t>((dy * dstW)) + x))) * dstBPP))));
{
rr3ds::source(p);
switch(outFmt){
case cast<uint32_t>(0ULL):{
n3ds_Machine_Write(m,q,a);
n3ds_Machine_Write(m,cast<uint32_t>((q + cast<uint32_t>(1ULL))),b);
n3ds_Machine_Write(m,cast<uint32_t>((q + cast<uint32_t>(2ULL))),g);
n3ds_Machine_Write(m,cast<uint32_t>((q + cast<uint32_t>(3ULL))),r);
break;}
case cast<uint32_t>(1ULL):{
n3ds_Machine_Write(m,q,b);
n3ds_Machine_Write(m,cast<uint32_t>((q + cast<uint32_t>(1ULL))),g);
n3ds_Machine_Write(m,cast<uint32_t>((q + cast<uint32_t>(2ULL))),r);
break;}
case cast<uint32_t>(2ULL):{
uint32_t v = cast<uint32_t>((cast<uint32_t>((shl<uint32_t>(cast<uint32_t>(shr<uint8_t>(r,cast<int64_t>(3ULL))),cast<int64_t>(11ULL)) | shl<uint32_t>(cast<uint32_t>(shr<uint8_t>(g,cast<int64_t>(2ULL))),cast<int64_t>(5ULL)))) | cast<uint32_t>(shr<uint8_t>(b,cast<int64_t>(3ULL)))));
n3ds_Machine_Write(m,q,cast<uint8_t>(v));
n3ds_Machine_Write(m,cast<uint32_t>((q + cast<uint32_t>(1ULL))),cast<uint8_t>(shr<uint32_t>(v,cast<int64_t>(8ULL))));
break;}
case cast<uint32_t>(3ULL):{
uint32_t v = cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((shl<uint32_t>(cast<uint32_t>(shr<uint8_t>(r,cast<int64_t>(3ULL))),cast<int64_t>(11ULL)) | shl<uint32_t>(cast<uint32_t>(shr<uint8_t>(g,cast<int64_t>(3ULL))),cast<int64_t>(6ULL)))) | shl<uint32_t>(cast<uint32_t>(shr<uint8_t>(b,cast<int64_t>(3ULL))),cast<int64_t>(1ULL)))) | cast<uint32_t>(shr<uint8_t>(a,cast<int64_t>(7ULL)))));
n3ds_Machine_Write(m,q,cast<uint8_t>(v));
n3ds_Machine_Write(m,cast<uint32_t>((q + cast<uint32_t>(1ULL))),cast<uint8_t>(shr<uint32_t>(v,cast<int64_t>(8ULL))));
break;}
case cast<uint32_t>(4ULL):{
uint32_t v = cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((shl<uint32_t>(cast<uint32_t>(shr<uint8_t>(r,cast<int64_t>(4ULL))),cast<int64_t>(12ULL)) | shl<uint32_t>(cast<uint32_t>(shr<uint8_t>(g,cast<int64_t>(4ULL))),cast<int64_t>(8ULL)))) | shl<uint32_t>(cast<uint32_t>(shr<uint8_t>(b,cast<int64_t>(4ULL))),cast<int64_t>(4ULL)))) | cast<uint32_t>(shr<uint8_t>(a,cast<int64_t>(4ULL)))));
n3ds_Machine_Write(m,q,cast<uint8_t>(v));
n3ds_Machine_Write(m,cast<uint32_t>((q + cast<uint32_t>(1ULL))),cast<uint8_t>(shr<uint32_t>(v,cast<int64_t>(8ULL))));
break;}
}}
}
}}
}}m->displayTransfers++;
n3ds_xferRecord rec = n3ds_xferRecord{dst,w,h,outFmt,dstBPP,dstW,src,srcW,flip};
std::string screen = std::string("top",3);
if ((dstH >= cast<uint32_t>(400ULL))) {
m->lastXferTop = rec;
}
else {
m->lastXferBottom = rec;
screen = std::string("bottom",6);
}
if (bool(m->OnPresent)) {
m->OnPresent(screen,n3ds_ScreenGeom{rec.src,rec.srcW,rec.w,rec.h,rec.flip,(screen == std::string("bottom",6)),rec.dst,rec.stride,rec.format,rec.bpp});
}
}
}
// tools/platform/n3ds/gx.go:290:1
void n3ds_Machine_processGXQueue(n3ds_Machine* m){
{
if ((m->gspSharedAddr == cast<uint32_t>(0ULL))) {
return ;
}
n3ds_Machine_pumpGX(m);
uint32_t hdr = cast<uint32_t>((m->gspSharedAddr + cast<uint32_t>(2048ULL)));
uint8_t count = n3ds_Machine_Read(m,cast<uint32_t>((hdr + cast<uint32_t>(1ULL))));
if ((count == cast<uint8_t>(0ULL))) {
return ;
}
uint8_t idx = n3ds_Machine_Read(m,cast<uint32_t>((hdr + cast<uint32_t>(0ULL))));
{uint8_t i = cast<uint8_t>(0ULL);for (;(i < count);i++){
uint32_t slot = modi<uint32_t>((cast<uint32_t>((cast<uint32_t>(idx) + cast<uint32_t>(i)))),cast<uint32_t>(15ULL));
uint32_t cmd = cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((m->gspSharedAddr + cast<uint32_t>(2048ULL))) + cast<uint32_t>(32ULL))) + cast<uint32_t>((slot * cast<uint32_t>(32ULL)))));
uint32_t id = cast<uint32_t>((cast<uint32_t>(n3ds_Machine_Read(m,cmd)) & cast<uint32_t>(31ULL)));
if (m->GXCapture) {
n3ds_Machine_captureGX(m,cmd,id);
}
std::array<uint32_t,8> w={};
{uint32_t j = cast<uint32_t>(0ULL);for (;(j < cast<uint32_t>(8ULL));j++){
w[j] = n3ds_Machine_ReadWord(m,cast<uint32_t>((cmd + cast<uint32_t>((j * cast<uint32_t>(4ULL))))));
}
}uint64_t base = m->instrs;
{
int64_t n = len(m->gxPending);
if (((n > cast<int64_t>(0ULL)) && (m->gxPending[cast<int64_t>((n - cast<int64_t>(1ULL)))].Deadline > base))) {
base = m->gxPending[cast<int64_t>((n - cast<int64_t>(1ULL)))].Deadline;
}
}
m->gxPending = append(m->gxPending,n3ds_gxPendingCmd{w,cast<uint64_t>((base + n3ds_gxLatency(id)))});
}
}n3ds_Machine_Write(m,cast<uint32_t>((hdr + cast<uint32_t>(0ULL))),cast<uint8_t>(modi<uint32_t>((cast<uint32_t>((cast<uint32_t>(idx) + cast<uint32_t>(count)))),cast<uint32_t>(15ULL))));
n3ds_Machine_Write(m,cast<uint32_t>((hdr + cast<uint32_t>(1ULL))),cast<uint8_t>(0ULL));
n3ds_Machine_pumpGX(m);
}
}
// tools/platform/n3ds/gx.go:357:1
uint64_t n3ds_gxLatency(uint32_t id){
{
{
switch(id){
case cast<uint32_t>(1ULL):{
return cast<uint64_t>(16384ULL);
break;}
case cast<uint32_t>(2ULL):{
return cast<uint64_t>(2048ULL);
break;}
case cast<uint32_t>(3ULL):case cast<uint32_t>(4ULL):{
return cast<uint64_t>(4096ULL);
break;}
case cast<uint32_t>(0ULL):{
return cast<uint64_t>(256ULL);
break;}
}}
return cast<uint64_t>(0ULL);
}
}
// tools/platform/n3ds/gx.go:374:1
std::tuple<uint64_t,bool> n3ds_Machine_gxDeadline(n3ds_Machine* m){
{
if ((len(m->gxPending) == cast<int64_t>(0ULL))) {
return {cast<uint64_t>(0ULL),false};
}
return {m->gxPending[cast<int64_t>(0ULL)].Deadline,true};
}
}
// tools/platform/n3ds/gx.go:383:1
void n3ds_Machine_pumpGX(n3ds_Machine* m){
{
{;for (;((len(m->gxPending) > cast<int64_t>(0ULL)) && (m->instrs >= m->gxPending[cast<int64_t>(0ULL)].Deadline));){
std::array<uint32_t,8> w = m->gxPending[cast<int64_t>(0ULL)].Words;
m->gxPending = sub(m->gxPending,cast<int64_t>(1ULL),len(m->gxPending));
Slice<uint8_t> raised={};
{
switch(cast<uint32_t>((w[cast<int64_t>(0ULL)] & cast<uint32_t>(31ULL)))){
case cast<uint32_t>(1ULL):{
m->framesSubmitted++;
n3ds_GPU_Execute(m->gpu,w[cast<int64_t>(1ULL)],w[cast<int64_t>(2ULL)]);
if (m->CPU->Halted) {
return ;
}
raised = append(raised,cast<uint8_t>(5ULL));
break;}
case cast<uint32_t>(2ULL):{
if ((w[cast<int64_t>(1ULL)] != cast<uint32_t>(0ULL))) {
n3ds_Machine_gxMemoryFill(m,w[cast<int64_t>(1ULL)],w[cast<int64_t>(2ULL)],w[cast<int64_t>(3ULL)],cast<uint32_t>((w[cast<int64_t>(7ULL)] & cast<uint32_t>(65535ULL))));
raised = append(raised,cast<uint8_t>(0ULL));
}
if ((w[cast<int64_t>(4ULL)] != cast<uint32_t>(0ULL))) {
n3ds_Machine_gxMemoryFill(m,w[cast<int64_t>(4ULL)],w[cast<int64_t>(5ULL)],w[cast<int64_t>(6ULL)],shr<uint32_t>(w[cast<int64_t>(7ULL)],cast<int64_t>(16ULL)));
raised = append(raised,cast<uint8_t>(1ULL));
}
break;}
case cast<uint32_t>(3ULL):{
n3ds_Machine_gxDisplayTransfer(m,w[cast<int64_t>(1ULL)],w[cast<int64_t>(2ULL)],w[cast<int64_t>(3ULL)],w[cast<int64_t>(4ULL)],w[cast<int64_t>(5ULL)]);
raised = append(raised,cast<uint8_t>(4ULL));
break;}
case cast<uint32_t>(4ULL):{
n3ds_Machine_gxTextureCopy(m,w[cast<int64_t>(1ULL)],w[cast<int64_t>(2ULL)],w[cast<int64_t>(3ULL)],w[cast<int64_t>(4ULL)],w[cast<int64_t>(5ULL)]);
raised = append(raised,cast<uint8_t>(4ULL));
break;}
case cast<uint32_t>(0ULL):{
n3ds_Machine_copyRange(m,w[cast<int64_t>(2ULL)],w[cast<int64_t>(1ULL)],w[cast<int64_t>(3ULL)]);
n3ds_GPU_invalidateTextures(m->gpu,w[cast<int64_t>(2ULL)],w[cast<int64_t>(3ULL)]);
raised = append(raised,cast<uint8_t>(6ULL));
break;}
}}
{auto&& tmp296 = raised;
for(int64_t tmp297=0;tmp297<len(tmp296);++tmp297){
auto id=tmp296[tmp297];n3ds_Machine_pushGSPInterrupt(m,id);
}}
if ((len(raised) > cast<int64_t>(0ULL))) {
n3ds_Machine_signalGSPEvent(m);
}
}
}}
}
// tools/platform/n3ds/hid.go:49:1
Error n3ds_Machine_SetKeys(n3ds_Machine* m,std::string list){
{
uint32_t mask={};
{auto&& tmp298 = go_strings_Split(list,std::string(",",1));
for(int64_t tmp299=0;tmp299<len(tmp298);++tmp299){
auto name=tmp298[tmp299];name = go_strings_TrimSpace(go_strings_ToLower(name));
if ((name == std::string("",0))) {
continue;
}
auto tmp300 = lookup(n3ds_hidButtonNames,name);
uint32_t bit = std::get<0>(tmp300);
bool ok = std::get<1>(tmp300);
if ((!ok)) {
return go_fmt_Errorf(std::string("unknown key %q (valid: a,b,x,y,l,r,up,down,left,right,start,select)",67),name);
}
mask |= bit;
}}
m->hidButtons = mask;
return {};
}
}
// tools/platform/n3ds/hid.go:113:1
void n3ds_Machine_SetTouch(n3ds_Machine* m,int64_t x,int64_t y,bool down){
{
m->hidTouchX = cast<uint16_t>(n3ds_clampInt(x,cast<int64_t>(0ULL),cast<int64_t>(319ULL)));
m->hidTouchY = cast<uint16_t>(n3ds_clampInt(y,cast<int64_t>(0ULL),cast<int64_t>(239ULL)));
m->hidTouchDown = down;
}
}
// tools/platform/n3ds/hid.go:119:1
int64_t n3ds_clampInt(int64_t v,int64_t lo,int64_t hi){
{
if ((v < lo)) {
return lo;
}
if ((v > hi)) {
return hi;
}
return v;
}
}
// tools/platform/n3ds/hid.go:137:1
void n3ds_Machine_updateHIDShared(n3ds_Machine* m){
{
if ((m->hidSharedAddr == cast<uint32_t>(0ULL))) {
return ;
}
uint32_t idx = m->hidRingIdx;
uint64_t tick = cast<uint64_t>((m->vblankCount + cast<uint64_t>(1ULL)));
uint32_t cur = m->hidButtons;
if ((((m->HidPulse > cast<int64_t>(0ULL)) && (cur != cast<uint32_t>(0ULL))) && (modi<uint64_t>(m->vblankCount,cast<uint64_t>(m->HidPulse)) >= cast<uint64_t>((cast<uint64_t>(m->HidPulse) - cast<uint64_t>(3ULL)))))) {
cur = cast<uint32_t>(0ULL);
}
m->hidPrevButtons = cur;
auto writeHeader = [&](uint32_t sb)->void{
uint32_t base = cast<uint32_t>((m->hidSharedAddr + sb));
n3ds_Machine_WriteWord(m,cast<uint32_t>((base + cast<uint32_t>(0ULL))),cast<uint32_t>(tick));
n3ds_Machine_WriteWord(m,cast<uint32_t>((base + cast<uint32_t>(4ULL))),cast<uint32_t>(shr<uint64_t>(tick,cast<int64_t>(32ULL))));
n3ds_Machine_WriteWord(m,cast<uint32_t>((base + cast<uint32_t>(8ULL))),cast<uint32_t>(tick));
n3ds_Machine_WriteWord(m,cast<uint32_t>((base + cast<uint32_t>(12ULL))),cast<uint32_t>(shr<uint64_t>(tick,cast<int64_t>(32ULL))));
n3ds_Machine_WriteWord(m,cast<uint32_t>((base + cast<uint32_t>(16ULL))),idx);
}
;
writeHeader(cast<uint32_t>(0ULL));
uint32_t ent = cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((m->hidSharedAddr + cast<uint32_t>(0ULL))) + cast<uint32_t>(40ULL))) + cast<uint32_t>((idx * cast<uint32_t>(16ULL)))));
n3ds_Machine_WriteWord(m,cast<uint32_t>((ent + cast<uint32_t>(0ULL))),cur);
n3ds_Machine_WriteWord(m,cast<uint32_t>((ent + cast<uint32_t>(4ULL))),cast<uint32_t>(0ULL));
n3ds_Machine_WriteWord(m,cast<uint32_t>((ent + cast<uint32_t>(8ULL))),cast<uint32_t>(0ULL));
n3ds_Machine_WriteWord(m,cast<uint32_t>((ent + cast<uint32_t>(12ULL))),rr3dsCirclePad);
writeHeader(cast<uint32_t>(168ULL));
uint32_t tent = cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((m->hidSharedAddr + cast<uint32_t>(168ULL))) + cast<uint32_t>(32ULL))) + cast<uint32_t>((idx * cast<uint32_t>(8ULL)))));
auto tmp301 = std::make_tuple(cast<uint32_t>(0ULL),cast<uint32_t>(0ULL));
uint32_t pos = std::get<0>(tmp301);
uint32_t valid = std::get<1>(tmp301);
if (m->hidTouchDown) {
pos = cast<uint32_t>((cast<uint32_t>(m->hidTouchX) | shl<uint32_t>(cast<uint32_t>(m->hidTouchY),cast<int64_t>(16ULL))));
valid = cast<uint32_t>(1ULL);
}
n3ds_Machine_WriteWord(m,cast<uint32_t>((tent + cast<uint32_t>(0ULL))),pos);
n3ds_Machine_WriteWord(m,cast<uint32_t>((tent + cast<uint32_t>(4ULL))),valid);
writeHeader(cast<uint32_t>(264ULL));
m->hidRingIdx = cast<uint32_t>(((cast<uint32_t>((idx + cast<uint32_t>(1ULL)))) & cast<uint32_t>(7ULL)));
if ((len(m->hidEvents) > cast<int64_t>(0ULL))) {
{
n3ds_kobject* obj = get(m->handles,m->hidEvents[cast<int64_t>(0ULL)]);
if (bool(obj)) {
obj->signal = true;
if (n3ds_Machine_signalObject(m,obj)) {
m->reschedule = true;
}
}
}
}
}
}
// tools/platform/n3ds/hid.go:208:1
void n3ds_Machine_DumpHIDReads(n3ds_Machine* m){
{
go_fmt_Printf(std::string("\012HID shared-memory reads (base 0x%08X):\012",40),m->hidSharedAddr);
if ((len(m->hidReadHist) == cast<int64_t>(0ULL))) {
go_fmt_Println(std::string("  (none \342\200\224 the game did not read the HID block during this run)",64));
return ;
}
Slice<uint32_t> offs = Slice<uint32_t>::make(cast<int64_t>(0ULL),len(m->hidReadHist));
{auto&& tmp302 = m->hidReadHist;
for(auto [tmp303,tmp304]:tmp302){
auto o=tmp303;offs = append(offs,o);
}}
go_sort_Slice(offs,[&](int64_t i,int64_t j)->bool{
return (offs[i] < offs[j]);
}
);
{auto&& tmp305 = offs;
for(int64_t tmp306=0;tmp306<len(tmp305);++tmp306){
auto o=tmp305[tmp306];go_fmt_Printf(std::string("  +0x%03X  %d reads (last reader pc=0x%08X)\012",44),o,get(m->hidReadHist,o),get(m->hidReadPC,o));
}}
}
}
// tools/platform/n3ds/ipc.go:27:1
n3ds_ipcHeader n3ds_parseIPCHeader(uint32_t w){
{
return n3ds_ipcHeader{cast<uint16_t>(shr<uint32_t>(w,cast<int64_t>(16ULL))),cast<int64_t>(cast<uint32_t>((shr<uint32_t>(w,cast<int64_t>(6ULL)) & cast<uint32_t>(63ULL)))),cast<int64_t>(cast<uint32_t>((w & cast<uint32_t>(63ULL))))};
}
}
// tools/platform/n3ds/ipc.go:36:1
uint32_t n3ds_Machine_ipcArg(n3ds_Machine* m,int64_t i){
{
return n3ds_Machine_ReadWord(m,cast<uint32_t>((n3ds_Machine_cmdBuf(m) + cast<uint32_t>((cast<uint32_t>(i) * cast<uint32_t>(4ULL))))));
}
}
// tools/platform/n3ds/ipc.go:40:1
void n3ds_Machine_ipcReply(n3ds_Machine* m,uint16_t cmd,Slice<uint32_t> values){
{
n3ds_Machine_WriteWord(m,n3ds_Machine_cmdBuf(m),cast<uint32_t>((shl<uint32_t>(cast<uint32_t>(cmd),cast<int64_t>(16ULL)) | shl<uint32_t>(cast<uint32_t>(cast<int64_t>((len(values) + cast<int64_t>(1ULL)))),cast<int64_t>(6ULL)))));
n3ds_Machine_WriteWord(m,cast<uint32_t>((n3ds_Machine_cmdBuf(m) + cast<uint32_t>(4ULL))),cast<uint32_t>(0ULL));
{auto&& tmp307 = values;
for(int64_t tmp308=0;tmp308<len(tmp307);++tmp308){
auto i=tmp308;auto v=tmp307[tmp308];n3ds_Machine_WriteWord(m,cast<uint32_t>((cast<uint32_t>((n3ds_Machine_cmdBuf(m) + cast<uint32_t>(8ULL))) + cast<uint32_t>((cast<uint32_t>(i) * cast<uint32_t>(4ULL))))),v);
}}
}
}
// tools/platform/n3ds/ipc.go:50:1
bool n3ds_Machine_handleIPC(n3ds_Machine* m,uint32_t handle){
{
n3ds_ipcHeader hdr = n3ds_parseIPCHeader(n3ds_Machine_ReadWord(m,n3ds_Machine_cmdBuf(m)));
std::string name = get(m->ports,handle);
{
auto tmp309 = lookup(m->services,handle);
std::string svc = std::get<0>(tmp309);
bool ok = std::get<1>(tmp309);
if (ok) {
name = svc;
}
}
m->ipcLog = append(m->ipcLog,n3ds_ipcCall{name,hdr.Command});
if (m->Verbose) {
go_fmt_Printf(std::string("  [t%d] IPC handle=0x%08X %-14s cmd 0x%04X (%d normal, %d translate)\012",69),m->curThread->id,handle,name,hdr.Command,hdr.Normal,hdr.Translate);
}
if (((name == std::string("",0)) && n3ds_Machine_isFileSession(m,handle))) {
return n3ds_Machine_ipcFile(m,handle,hdr);
}
if (((name == std::string("",0)) && n3ds_Machine_isDirSession(m,handle))) {
return n3ds_Machine_ipcDir(m,handle,hdr);
}
{
auto tmp311=name;
if (tmp311==(std::string("srv:",4)) || tmp311==(std::string("srv:pm",6))){
return n3ds_Machine_ipcSrv(m,hdr);
}
else {
return n3ds_Machine_ipcService(m,name,hdr);
}
}
tmp310:;
}
}
// tools/platform/n3ds/ipc.go:79:1
bool n3ds_Machine_ipcSrv(n3ds_Machine* m,n3ds_ipcHeader hdr){
{
{
switch(hdr.Command){
case cast<uint16_t>(1ULL):{
n3ds_Machine_ipcReply(m,hdr.Command,Slice<uint32_t>{});
return true;
break;}
case cast<uint16_t>(2ULL):{
uint32_t h = n3ds_Machine_newHandle(m,std::string("notification-semaphore",22),true);
n3ds_Machine_WriteWord(m,n3ds_Machine_cmdBuf(m),cast<uint32_t>((cast<uint32_t>((shl<uint32_t>(cast<uint32_t>(hdr.Command),cast<int64_t>(16ULL)) | cast<uint32_t>(64ULL))) | cast<uint32_t>(2ULL))));
n3ds_Machine_WriteWord(m,cast<uint32_t>((n3ds_Machine_cmdBuf(m) + cast<uint32_t>(4ULL))),cast<uint32_t>(0ULL));
n3ds_Machine_WriteWord(m,cast<uint32_t>((n3ds_Machine_cmdBuf(m) + cast<uint32_t>(8ULL))),cast<uint32_t>(0ULL));
n3ds_Machine_WriteWord(m,cast<uint32_t>((n3ds_Machine_cmdBuf(m) + cast<uint32_t>(12ULL))),h);
return true;
break;}
case cast<uint16_t>(5ULL):{
std::string name = n3ds_Machine_readServiceName(m);
uint32_t h = n3ds_Machine_newHandle(m,(std::string("service:",8) + name),false);
m->services[h] = name;
n3ds_Machine_WriteWord(m,n3ds_Machine_cmdBuf(m),cast<uint32_t>((cast<uint32_t>((shl<uint32_t>(cast<uint32_t>(hdr.Command),cast<int64_t>(16ULL)) | cast<uint32_t>(64ULL))) | cast<uint32_t>(2ULL))));
n3ds_Machine_WriteWord(m,cast<uint32_t>((n3ds_Machine_cmdBuf(m) + cast<uint32_t>(4ULL))),cast<uint32_t>(0ULL));
n3ds_Machine_WriteWord(m,cast<uint32_t>((n3ds_Machine_cmdBuf(m) + cast<uint32_t>(8ULL))),cast<uint32_t>(0ULL));
n3ds_Machine_WriteWord(m,cast<uint32_t>((n3ds_Machine_cmdBuf(m) + cast<uint32_t>(12ULL))),h);
if (m->Verbose) {
go_fmt_Printf(std::string("    GetServiceHandle %q -> 0x%08X\012",34),name,h);
}
return true;
break;}
case cast<uint16_t>(9ULL):case cast<uint16_t>(10ULL):{
n3ds_Machine_ipcReply(m,hdr.Command,Slice<uint32_t>{});
return true;
break;}
case cast<uint16_t>(11ULL):{
m->notifyWaiters = append(m->notifyWaiters,m->curThread->id);
m->curThread->state = cast<n3ds_threadState>(2ULL);
m->reschedule = true;
return true;
break;}
}}
arm_CPU_Halt(m->CPU,std::string("srv: command 0x%04X unimplemented at 0x%08X after %d instructions",65),hdr.Command,arm_CPU_PC(m->CPU),m->CPU->Instrs);
return true;
}
}
// tools/platform/n3ds/ipc.go:122:1
void n3ds_Machine_publishNotification(n3ds_Machine* m,uint32_t id){
{
Slice<uint32_t> rest = sub(m->notifyWaiters,0,cast<int64_t>(0ULL));
{auto&& tmp312 = m->notifyWaiters;
for(int64_t tmp313=0;tmp313<len(tmp312);++tmp313){
auto tid=tmp312[tmp313];n3ds_thread* t = n3ds_Machine_threadByID(m,tid);
if (((!t) || (t->state != cast<n3ds_threadState>(2ULL)))) {
continue;
}
uint32_t buf = cast<uint32_t>((t->tlsBase + cast<uint32_t>(128ULL)));
n3ds_Machine_WriteWord(m,buf,cast<uint32_t>(721024ULL));
n3ds_Machine_WriteWord(m,cast<uint32_t>((buf + cast<uint32_t>(4ULL))),cast<uint32_t>(0ULL));
n3ds_Machine_WriteWord(m,cast<uint32_t>((buf + cast<uint32_t>(8ULL))),id);
if (n3ds_Machine_wake(m,t)) {
m->reschedule = true;
}
}}
m->notifyWaiters = rest;
}
}
// tools/platform/n3ds/ipc.go:143:1
bool n3ds_knownService(std::string name){
{
{
auto tmp315=n3ds_serviceBase(name);
if (tmp315==(std::string("APT",3)) || tmp315==(std::string("gsp",3)) || tmp315==(std::string("hid",3)) || tmp315==(std::string("cfg",3)) || tmp315==(std::string("fs",2)) || tmp315==(std::string("ndm",3)) || tmp315==(std::string("ptm",3)) || tmp315==(std::string("ac",2)) || tmp315==(std::string("act",3)) || tmp315==(std::string("frd",3)) || tmp315==(std::string("cecd",4)) || tmp315==(std::string("boss",4)) || tmp315==(std::string("nim",3)) || tmp315==(std::string("mic",3)) || tmp315==(std::string("csnd",4)) || tmp315==(std::string("dsp",3)) || tmp315==(std::string("y2r",3)) || tmp315==(std::string("am",2)) || tmp315==(std::string("ns",2)) || tmp315==(std::string("nfc",3)) || tmp315==(std::string("pxi",3)) || tmp315==(std::string("srv",3)) || tmp315==(std::string("cam",3)) || tmp315==(std::string("mcu",3))){
return true;
}
}
tmp314:;
return false;
}
}
// tools/platform/n3ds/ipc.go:159:1
std::string n3ds_Machine_readServiceName(n3ds_Machine* m){
{
{auto&& tmp316 = Slice<uint64_t>{cast<uint64_t>(0ULL),cast<uint64_t>(16ULL),cast<uint64_t>(8ULL),cast<uint64_t>(24ULL)};
for(int64_t tmp317=0;tmp317<len(tmp316);++tmp317){
auto rot=tmp316[tmp317];{
std::string n = n3ds_Machine_decodeName(m,rot);
if (n3ds_knownService(n)) {
return n;
}
}
}}
return n3ds_Machine_decodeName(m,cast<uint64_t>(0ULL));
}
}
// tools/platform/n3ds/ipc.go:169:1
std::string n3ds_Machine_decodeName(n3ds_Machine* m,uint64_t rot){
{
int64_t n = cast<int64_t>(n3ds_Machine_ReadWord(m,cast<uint32_t>((n3ds_Machine_cmdBuf(m) + cast<uint32_t>(12ULL)))));
if (((n <= cast<int64_t>(0ULL)) || (n > cast<int64_t>(8ULL)))) {
n = cast<int64_t>(8ULL);
}
Slice<uint8_t> b={};
{int64_t w = cast<int64_t>(0ULL);for (;(w < cast<int64_t>(2ULL));w++){
uint32_t word = n3ds_Machine_ReadWord(m,cast<uint32_t>((cast<uint32_t>((n3ds_Machine_cmdBuf(m) + cast<uint32_t>(4ULL))) + cast<uint32_t>((cast<uint32_t>(w) * cast<uint32_t>(4ULL))))));
word = cast<uint32_t>((shr<uint32_t>(word,rot) | shl<uint32_t>(word,(cast<uint64_t>((cast<uint64_t>(32ULL) - rot))))));
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(4ULL));i++){
uint8_t ch = cast<uint8_t>(shr<uint32_t>(word,(cast<uint64_t>((cast<uint64_t>(i) * cast<uint64_t>(8ULL))))));
if (((ch == cast<uint8_t>(0ULL)) || (len(b) >= n))) {
return cast<std::string>(b);
}
b = append(b,ch);
}
}}
}return cast<std::string>(b);
}
}
// tools/platform/n3ds/ipc.go:189:1
uint32_t n3ds_Machine_newHandle(n3ds_Machine* m,std::string kind,bool signalled){
{
uint32_t h = m->nextHandle;
m->nextHandle++;
m->handles[h] = arenaNew(n3ds_kobject{kind,{},signalled,{},{},{},{},{},{},{},{}});
return h;
}
}
// tools/platform/n3ds/ipc.go:203:1
std::string n3ds_ipcCall_Service(n3ds_ipcCall c){
{
if ((c.service == std::string("",0))) {
return std::string("?",1);
}
return c.service;
}
}
// tools/platform/n3ds/ipc.go:211:1
Slice<n3ds_ipcCall> n3ds_Machine_IPCLog(n3ds_Machine* m){
{
return m->ipcLog;
}
}
// tools/platform/n3ds/ipc_services.go:17:1
bool n3ds_Machine_ipcService(n3ds_Machine* m,std::string name,n3ds_ipcHeader hdr){
{
{
auto tmp319=n3ds_serviceBase(name);
if (tmp319==(std::string("APT",3))){
return n3ds_Machine_ipcAPT(m,name,hdr);
}
else if (tmp319==(std::string("gsp",3))){
return n3ds_Machine_ipcGSP(m,hdr);
}
else if (tmp319==(std::string("hid",3))){
return n3ds_Machine_ipcHID(m,hdr);
}
else if (tmp319==(std::string("cfg",3))){
return n3ds_Machine_ipcCFG(m,hdr);
}
else if (tmp319==(std::string("fs",2))){
return n3ds_Machine_ipcFS(m,hdr);
}
else if (tmp319==(std::string("err",3))){
return n3ds_Machine_ipcErr(m,hdr);
}
else if (tmp319==(std::string("dsp",3))){
return n3ds_Machine_ipcDSP(m,hdr);
}
else if (tmp319==(std::string("act",3))){
if ((hdr.Command == cast<uint16_t>(1ULL))) {
n3ds_Machine_ipcReply(m,hdr.Command,Slice<uint32_t>{});
return true;
}
}
else if (tmp319==(std::string("nfc",3))){
{
switch(hdr.Command){
case cast<uint16_t>(1ULL):{
n3ds_Machine_ipcReply(m,hdr.Command,Slice<uint32_t>{});
return true;
break;}
case cast<uint16_t>(11ULL):case cast<uint16_t>(12ULL):case cast<uint16_t>(15ULL):{
n3ds_Machine_ipcReply(m,hdr.Command,Slice<uint32_t>{cast<uint32_t>(0ULL),cast<uint32_t>(0ULL)});
return true;
break;}
}}
}
else if (tmp319==(std::string("ndm",3)) || tmp319==(std::string("ptm",3)) || tmp319==(std::string("ac",2)) || tmp319==(std::string("frd",3)) || tmp319==(std::string("cecd",4)) || tmp319==(std::string("boss",4)) || tmp319==(std::string("nim",3)) || tmp319==(std::string("mic",3)) || tmp319==(std::string("csnd",4)) || tmp319==(std::string("y2r",3))){
n3ds_Machine_ipcReply(m,hdr.Command,Slice<uint32_t>{});
return true;
}
}
tmp318:;
arm_CPU_Halt(m->CPU,std::string("service %q command 0x%04X unimplemented at 0x%08X after %d instructions",71),name,hdr.Command,arm_CPU_PC(m->CPU),m->CPU->Instrs);
return true;
}
}
// tools/platform/n3ds/ipc_services.go:73:1
std::string n3ds_serviceBase(std::string name){
{
{auto&& tmp320 = name;
for(int64_t tmp321=0;tmp321<len(tmp320);++tmp321){
auto i=tmp321;auto ch=tmp320[tmp321];if ((ch == cast<int32_t>(58ULL))) {
return sub(name,0,i);
}
}}
return name;
}
}
// tools/platform/n3ds/ipc_services.go:87:1
bool n3ds_Machine_ipcAPT(n3ds_Machine* m,std::string name,n3ds_ipcHeader hdr){
{
{
switch(hdr.Command){
case cast<uint16_t>(1ULL):{
uint32_t h = n3ds_Machine_newHandle(m,std::string("apt-lock",8),true);
n3ds_Machine_WriteWord(m,n3ds_Machine_cmdBuf(m),cast<uint32_t>((cast<uint32_t>((shl<uint32_t>(cast<uint32_t>(hdr.Command),cast<int64_t>(16ULL)) | cast<uint32_t>(192ULL))) | cast<uint32_t>(2ULL))));
n3ds_Machine_WriteWord(m,cast<uint32_t>((n3ds_Machine_cmdBuf(m) + cast<uint32_t>(4ULL))),cast<uint32_t>(0ULL));
n3ds_Machine_WriteWord(m,cast<uint32_t>((n3ds_Machine_cmdBuf(m) + cast<uint32_t>(8ULL))),cast<uint32_t>(0ULL));
n3ds_Machine_WriteWord(m,cast<uint32_t>((n3ds_Machine_cmdBuf(m) + cast<uint32_t>(12ULL))),cast<uint32_t>(0ULL));
n3ds_Machine_WriteWord(m,cast<uint32_t>((n3ds_Machine_cmdBuf(m) + cast<uint32_t>(16ULL))),cast<uint32_t>(0ULL));
n3ds_Machine_WriteWord(m,cast<uint32_t>((n3ds_Machine_cmdBuf(m) + cast<uint32_t>(20ULL))),h);
return true;
break;}
case cast<uint16_t>(2ULL):case cast<uint16_t>(3ULL):case cast<uint16_t>(4ULL):{
if ((hdr.Command == cast<uint16_t>(2ULL))) {
uint32_t ev1 = n3ds_Machine_newHandle(m,std::string("apt-notify",10),true);
uint32_t ev2 = n3ds_Machine_newHandle(m,std::string("apt-resume",10),true);
auto tmp322 = std::make_tuple(ev1,ev2);
m->aptNotifyEv = std::get<0>(tmp322);
m->aptResumeEv = std::get<1>(tmp322);
n3ds_Machine_WriteWord(m,n3ds_Machine_cmdBuf(m),cast<uint32_t>((cast<uint32_t>((shl<uint32_t>(cast<uint32_t>(hdr.Command),cast<int64_t>(16ULL)) | cast<uint32_t>(64ULL))) | cast<uint32_t>(4ULL))));
n3ds_Machine_WriteWord(m,cast<uint32_t>((n3ds_Machine_cmdBuf(m) + cast<uint32_t>(4ULL))),cast<uint32_t>(0ULL));
n3ds_Machine_WriteWord(m,cast<uint32_t>((n3ds_Machine_cmdBuf(m) + cast<uint32_t>(8ULL))),cast<uint32_t>(0ULL));
n3ds_Machine_WriteWord(m,cast<uint32_t>((n3ds_Machine_cmdBuf(m) + cast<uint32_t>(12ULL))),ev1);
n3ds_Machine_WriteWord(m,cast<uint32_t>((n3ds_Machine_cmdBuf(m) + cast<uint32_t>(16ULL))),ev2);
return true;
}
n3ds_Machine_ipcReply(m,hdr.Command,Slice<uint32_t>{});
return true;
break;}
case cast<uint16_t>(5ULL):{
n3ds_Machine_ipcReply(m,hdr.Command,Slice<uint32_t>{cast<uint32_t>(0ULL),cast<uint32_t>(0ULL),cast<uint32_t>(768ULL),cast<uint32_t>(768ULL)});
return true;
break;}
case cast<uint16_t>(6ULL):{
n3ds_Machine_ipcReply(m,hdr.Command,Slice<uint32_t>{cast<uint32_t>(0ULL),cast<uint32_t>(0ULL),cast<uint32_t>(1ULL),cast<uint32_t>(1ULL),cast<uint32_t>(1ULL),cast<uint32_t>(0ULL)});
return true;
break;}
case cast<uint16_t>(9ULL):{
n3ds_Machine_ipcReply(m,hdr.Command,Slice<uint32_t>{cast<uint32_t>(1ULL)});
return true;
break;}
case cast<uint16_t>(11ULL):{
n3ds_Machine_ipcReply(m,hdr.Command,Slice<uint32_t>{cast<uint32_t>(1ULL)});
return true;
break;}
case cast<uint16_t>(12ULL):{
if (m->Verbose) {
go_fmt_Printf(std::string("    SendParameter sender=0x%X dest=0x%X signal=%d size=0x%X\012",60),n3ds_Machine_ReadWord(m,cast<uint32_t>((n3ds_Machine_cmdBuf(m) + cast<uint32_t>(4ULL)))),n3ds_Machine_ReadWord(m,cast<uint32_t>((n3ds_Machine_cmdBuf(m) + cast<uint32_t>(8ULL)))),n3ds_Machine_ReadWord(m,cast<uint32_t>((n3ds_Machine_cmdBuf(m) + cast<uint32_t>(12ULL)))),n3ds_Machine_ReadWord(m,cast<uint32_t>((n3ds_Machine_cmdBuf(m) + cast<uint32_t>(16ULL)))));
}
uint32_t h = n3ds_Machine_newHandle(m,std::string("apt-reply-shared",16),false);
m->handles[h]->blockSize = cast<uint32_t>(4096ULL);
m->aptParams = append(m->aptParams,n3ds_aptParam{n3ds_Machine_ReadWord(m,cast<uint32_t>((n3ds_Machine_cmdBuf(m) + cast<uint32_t>(8ULL)))),cast<uint32_t>(3ULL),h,{}});
m->aptWakePending = true;
n3ds_Machine_ipcReply(m,hdr.Command,Slice<uint32_t>{cast<uint32_t>(0ULL)});
return true;
break;}
case cast<uint16_t>(13ULL):case cast<uint16_t>(14ULL):{
n3ds_aptParam p = n3ds_aptParam{cast<uint32_t>(0ULL),cast<uint32_t>(1ULL),{},{}};
if ((len(m->aptParams) > cast<int64_t>(0ULL))) {
p = m->aptParams[cast<int64_t>(0ULL)];
if ((hdr.Command == cast<uint16_t>(13ULL))) {
m->aptParams = sub(m->aptParams,cast<int64_t>(1ULL),len(m->aptParams));
if ((len(m->aptParams) > cast<int64_t>(0ULL))) {
m->aptWakePending = true;
}
}
}
if ((len(p.Data) > cast<int64_t>(0ULL))) {
uint32_t ptr = n3ds_Machine_ReadWord(m,cast<uint32_t>((m->curThread->tlsBase + cast<uint32_t>(388ULL))));
uint32_t max = shr<uint32_t>(n3ds_Machine_ReadWord(m,cast<uint32_t>((m->curThread->tlsBase + cast<uint32_t>(384ULL)))),cast<int64_t>(14ULL));
uint32_t n = cast<uint32_t>(len(p.Data));
if ((n > max)) {
n = max;
}
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < n);i++){
n3ds_Machine_Write(m,cast<uint32_t>((ptr + i)),p.Data[i]);
}
}}
n3ds_Machine_WriteWord(m,n3ds_Machine_cmdBuf(m),cast<uint32_t>((cast<uint32_t>((shl<uint32_t>(cast<uint32_t>(hdr.Command),cast<int64_t>(16ULL)) | cast<uint32_t>(256ULL))) | cast<uint32_t>(2ULL))));
n3ds_Machine_WriteWord(m,cast<uint32_t>((n3ds_Machine_cmdBuf(m) + cast<uint32_t>(4ULL))),cast<uint32_t>(0ULL));
n3ds_Machine_WriteWord(m,cast<uint32_t>((n3ds_Machine_cmdBuf(m) + cast<uint32_t>(8ULL))),p.Sender);
n3ds_Machine_WriteWord(m,cast<uint32_t>((n3ds_Machine_cmdBuf(m) + cast<uint32_t>(12ULL))),p.Command);
n3ds_Machine_WriteWord(m,cast<uint32_t>((n3ds_Machine_cmdBuf(m) + cast<uint32_t>(16ULL))),cast<uint32_t>(len(p.Data)));
n3ds_Machine_WriteWord(m,cast<uint32_t>((n3ds_Machine_cmdBuf(m) + cast<uint32_t>(20ULL))),cast<uint32_t>(0ULL));
n3ds_Machine_WriteWord(m,cast<uint32_t>((n3ds_Machine_cmdBuf(m) + cast<uint32_t>(24ULL))),p.Handle);
return true;
break;}
case cast<uint16_t>(67ULL):case cast<uint16_t>(75ULL):case cast<uint16_t>(76ULL):{
if ((hdr.Command == cast<uint16_t>(67ULL))) {
m->aptWakePending = true;
}
n3ds_Machine_ipcReply(m,hdr.Command,Slice<uint32_t>{});
return true;
break;}
case cast<uint16_t>(24ULL):{
n3ds_Machine_ipcReply(m,hdr.Command,Slice<uint32_t>{});
return true;
break;}
case cast<uint16_t>(22ULL):case cast<uint16_t>(23ULL):{
n3ds_Machine_ipcReply(m,hdr.Command,Slice<uint32_t>{});
return true;
break;}
case cast<uint16_t>(30ULL):{
m->aptParams = append(m->aptParams,n3ds_aptParam{n3ds_Machine_ReadWord(m,cast<uint32_t>((n3ds_Machine_cmdBuf(m) + cast<uint32_t>(4ULL)))),cast<uint32_t>(10ULL),{},Slice<uint8_t>::make(cast<int64_t>(132ULL))},n3ds_aptParam{{},cast<uint32_t>(8ULL),{},{}});
m->aptWakePending = true;
n3ds_Machine_ipcReply(m,hdr.Command,Slice<uint32_t>{});
return true;
break;}
case cast<uint16_t>(64ULL):{
n3ds_Machine_ipcReply(m,hdr.Command,Slice<uint32_t>{});
return true;
break;}
case cast<uint16_t>(257ULL):case cast<uint16_t>(258ULL):{
n3ds_Machine_ipcReply(m,hdr.Command,Slice<uint32_t>{cast<uint32_t>(0ULL)});
return true;
break;}
case cast<uint16_t>(85ULL):{
n3ds_Machine_ipcReply(m,hdr.Command,Slice<uint32_t>{});
return true;
break;}
case cast<uint16_t>(59ULL):{
n3ds_Machine_ipcReply(m,hdr.Command,Slice<uint32_t>{});
return true;
break;}
case cast<uint16_t>(43ULL):{
n3ds_Machine_ipcReply(m,hdr.Command,Slice<uint32_t>{});
return true;
break;}
case cast<uint16_t>(44ULL):{
n3ds_Machine_ipcReply(m,hdr.Command,Slice<uint32_t>{});
return true;
break;}
case cast<uint16_t>(68ULL):{
n3ds_Machine_WriteWord(m,n3ds_Machine_cmdBuf(m),cast<uint32_t>((shl<uint32_t>(cast<uint32_t>(hdr.Command),cast<int64_t>(16ULL)) | cast<uint32_t>(64ULL))));
n3ds_Machine_WriteWord(m,cast<uint32_t>((n3ds_Machine_cmdBuf(m) + cast<uint32_t>(4ULL))),cast<uint32_t>(3365982204ULL));
return true;
break;}
case cast<uint16_t>(62ULL):{
n3ds_Machine_ipcReply(m,hdr.Command,Slice<uint32_t>{});
return true;
break;}
}}
arm_CPU_Halt(m->CPU,std::string("APT command 0x%04X unimplemented at 0x%08X after %d instructions",64),hdr.Command,arm_CPU_PC(m->CPU),m->CPU->Instrs);
return true;
}
}
// tools/platform/n3ds/ipc_services.go:360:1
void n3ds_Machine_signalAPTEvents(n3ds_Machine* m){
{
{
n3ds_kobject* obj = get(m->handles,m->aptResumeEv);
if (bool(obj)) {
obj->signal = true;
if (n3ds_Machine_signalObject(m,obj)) {
m->reschedule = true;
}
}
}
}
}
// tools/platform/n3ds/ipc_services.go:376:1
bool n3ds_Machine_ipcGSP(n3ds_Machine* m,n3ds_ipcHeader hdr){
{
{
switch(hdr.Command){
case cast<uint16_t>(19ULL):{
m->gspEvent = n3ds_Machine_ipcArg(m,cast<int64_t>(3ULL));
if ((m->gspShared == cast<uint32_t>(0ULL))) {
m->gspShared = n3ds_Machine_newHandle(m,std::string("gsp-shared",10),false);
m->handles[m->gspShared]->blockSize = cast<uint32_t>(4096ULL);
}
n3ds_Machine_WriteWord(m,n3ds_Machine_cmdBuf(m),cast<uint32_t>((cast<uint32_t>((shl<uint32_t>(cast<uint32_t>(hdr.Command),cast<int64_t>(16ULL)) | cast<uint32_t>(128ULL))) | cast<uint32_t>(2ULL))));
n3ds_Machine_WriteWord(m,cast<uint32_t>((n3ds_Machine_cmdBuf(m) + cast<uint32_t>(4ULL))),cast<uint32_t>(0ULL));
n3ds_Machine_WriteWord(m,cast<uint32_t>((n3ds_Machine_cmdBuf(m) + cast<uint32_t>(8ULL))),cast<uint32_t>(0ULL));
n3ds_Machine_WriteWord(m,cast<uint32_t>((n3ds_Machine_cmdBuf(m) + cast<uint32_t>(12ULL))),cast<uint32_t>(0ULL));
n3ds_Machine_WriteWord(m,cast<uint32_t>((n3ds_Machine_cmdBuf(m) + cast<uint32_t>(16ULL))),m->gspShared);
return true;
break;}
case cast<uint16_t>(1ULL):case cast<uint16_t>(2ULL):case cast<uint16_t>(3ULL):case cast<uint16_t>(4ULL):case cast<uint16_t>(5ULL):{
n3ds_Machine_ipcReply(m,hdr.Command,Slice<uint32_t>{});
return true;
break;}
case cast<uint16_t>(8ULL):{
n3ds_Machine_ipcReply(m,hdr.Command,Slice<uint32_t>{});
return true;
break;}
case cast<uint16_t>(10ULL):{
n3ds_Machine_ipcReply(m,hdr.Command,Slice<uint32_t>{});
return true;
break;}
case cast<uint16_t>(11ULL):{
m->framesSubmitted++;
n3ds_Machine_ipcReply(m,hdr.Command,Slice<uint32_t>{});
return true;
break;}
case cast<uint16_t>(12ULL):case cast<uint16_t>(13ULL):case cast<uint16_t>(14ULL):{
n3ds_Machine_ipcReply(m,hdr.Command,Slice<uint32_t>{});
return true;
break;}
case cast<uint16_t>(22ULL):case cast<uint16_t>(23ULL):{
n3ds_Machine_ipcReply(m,hdr.Command,Slice<uint32_t>{});
return true;
break;}
case cast<uint16_t>(30ULL):case cast<uint16_t>(32ULL):{
n3ds_Machine_ipcReply(m,hdr.Command,Slice<uint32_t>{});
return true;
break;}
case cast<uint16_t>(25ULL):case cast<uint16_t>(26ULL):{
n3ds_Machine_ipcReply(m,hdr.Command,Slice<uint32_t>{});
return true;
break;}
case cast<uint16_t>(24ULL):{
n3ds_Machine_ipcReply(m,hdr.Command,Slice<uint32_t>{cast<uint32_t>(0ULL),cast<uint32_t>(0ULL),cast<uint32_t>(0ULL),cast<uint32_t>(0ULL),cast<uint32_t>(0ULL),cast<uint32_t>(0ULL),cast<uint32_t>(0ULL),cast<uint32_t>(0ULL)});
return true;
break;}
case cast<uint16_t>(31ULL):{
n3ds_Machine_ipcReply(m,hdr.Command,Slice<uint32_t>{});
return true;
break;}
}}
arm_CPU_Halt(m->CPU,std::string("gsp command 0x%04X unimplemented at 0x%08X after %d instructions",64),hdr.Command,arm_CPU_PC(m->CPU),m->CPU->Instrs);
return true;
}
}
// tools/platform/n3ds/ipc_services.go:439:1
bool n3ds_Machine_ipcHID(n3ds_Machine* m,n3ds_ipcHeader hdr){
{
{
switch(hdr.Command){
case cast<uint16_t>(10ULL):{
uint32_t sh = n3ds_Machine_newHandle(m,std::string("hid-shared",10),false);
m->handles[sh]->blockSize = cast<uint32_t>(4096ULL);
m->hidShared = sh;
m->hidEvents = sub(m->hidEvents,0,cast<int64_t>(0ULL));
n3ds_Machine_WriteWord(m,n3ds_Machine_cmdBuf(m),cast<uint32_t>((cast<uint32_t>((shl<uint32_t>(cast<uint32_t>(hdr.Command),cast<int64_t>(16ULL)) | cast<uint32_t>(64ULL))) | cast<uint32_t>(13ULL))));
n3ds_Machine_WriteWord(m,cast<uint32_t>((n3ds_Machine_cmdBuf(m) + cast<uint32_t>(4ULL))),cast<uint32_t>(0ULL));
n3ds_Machine_WriteWord(m,cast<uint32_t>((n3ds_Machine_cmdBuf(m) + cast<uint32_t>(8ULL))),cast<uint32_t>(0ULL));
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(6ULL));i++){
uint32_t h = sh;
if ((i > cast<int64_t>(0ULL))) {
h = n3ds_Machine_newHandle(m,std::string("hid-event",9),true);
m->hidEvents = append(m->hidEvents,h);
}
n3ds_Machine_WriteWord(m,cast<uint32_t>((cast<uint32_t>((n3ds_Machine_cmdBuf(m) + cast<uint32_t>(12ULL))) + cast<uint32_t>((cast<uint32_t>(i) * cast<uint32_t>(4ULL))))),h);
}
}return true;
break;}
case cast<uint16_t>(17ULL):case cast<uint16_t>(18ULL):case cast<uint16_t>(19ULL):{
n3ds_Machine_ipcReply(m,hdr.Command,Slice<uint32_t>{});
return true;
break;}
}}
arm_CPU_Halt(m->CPU,std::string("hid command 0x%04X unimplemented at 0x%08X after %d instructions",64),hdr.Command,arm_CPU_PC(m->CPU),m->CPU->Instrs);
return true;
}
}
// tools/platform/n3ds/ipc_services.go:466:1
bool n3ds_Machine_ipcCFG(n3ds_Machine* m,n3ds_ipcHeader hdr){
{
{
switch(hdr.Command){
case cast<uint16_t>(1ULL):{
auto tmp323 = std::make_tuple(n3ds_Machine_ipcArg(m,cast<int64_t>(1ULL)),n3ds_Machine_ipcArg(m,cast<int64_t>(2ULL)),n3ds_Machine_ipcArg(m,cast<int64_t>(4ULL)));
uint32_t size = std::get<0>(tmp323);
uint32_t blkID = std::get<1>(tmp323);
uint32_t out = std::get<2>(tmp323);
n3ds_Machine_writeConfigBlock(m,blkID,out,size);
n3ds_Machine_ipcReply(m,hdr.Command,Slice<uint32_t>{cast<uint32_t>(0ULL)});
return true;
break;}
case cast<uint16_t>(2ULL):{
n3ds_Machine_ipcReply(m,hdr.Command,Slice<uint32_t>{cast<uint32_t>(2ULL)});
return true;
break;}
case cast<uint16_t>(3ULL):{
uint64_t h = cast<uint64_t>((cast<uint64_t>(5937279718600631670ULL) ^ cast<uint64_t>(n3ds_Machine_ipcArg(m,cast<int64_t>(1ULL)))));
h = cast<uint64_t>(((cast<uint64_t>((h ^ (shr<uint64_t>(h,cast<int64_t>(30ULL)))))) * cast<uint64_t>(13787848793156543929ULL)));
h = cast<uint64_t>(((cast<uint64_t>((h ^ (shr<uint64_t>(h,cast<int64_t>(27ULL)))))) * cast<uint64_t>(10723151780598845931ULL)));
h ^= shr<uint64_t>(h,cast<int64_t>(31ULL));
n3ds_Machine_ipcReply(m,hdr.Command,Slice<uint32_t>{cast<uint32_t>(h),cast<uint32_t>(shr<uint64_t>(h,cast<int64_t>(32ULL)))});
return true;
break;}
}}
arm_CPU_Halt(m->CPU,std::string("cfg command 0x%04X unimplemented at 0x%08X after %d instructions",64),hdr.Command,arm_CPU_PC(m->CPU),m->CPU->Instrs);
return true;
}
}
// tools/platform/n3ds/ipc_services.go:511:1
void n3ds_Machine_writeConfigBlock(n3ds_Machine* m,uint32_t blkID,uint32_t out,uint32_t size){
{
if ((out == cast<uint32_t>(0ULL))) {
return ;
}
Slice<uint8_t> buf = Slice<uint8_t>::make(size);
{
switch(blkID){
case cast<uint32_t>(655362ULL):{
if ((size >= cast<uint32_t>(1ULL))) {
buf[cast<int64_t>(0ULL)] = cast<uint8_t>(1ULL);
}
break;}
case cast<uint32_t>(720896ULL):{
if ((size >= cast<uint32_t>(1ULL))) {
buf[cast<int64_t>(0ULL)] = cast<uint8_t>(2ULL);
}
break;}
case cast<uint32_t>(458753ULL):{
if ((size >= cast<uint32_t>(1ULL))) {
buf[cast<int64_t>(0ULL)] = cast<uint8_t>(1ULL);
}
break;}
case cast<uint32_t>(1245184ULL):{
if ((size >= cast<uint32_t>(4ULL))) {
auto tmp324 = std::make_tuple(cast<uint8_t>(1ULL),cast<uint8_t>(0ULL));
buf[cast<int64_t>(0ULL)] = std::get<0>(tmp324);
buf[cast<int64_t>(1ULL)] = std::get<1>(tmp324);
auto tmp325 = std::make_tuple(cast<uint8_t>(1ULL),cast<uint8_t>(0ULL));
buf[cast<int64_t>(2ULL)] = std::get<0>(tmp325);
buf[cast<int64_t>(3ULL)] = std::get<1>(tmp325);
}
break;}
case cast<uint32_t>(327685ULL):{
{auto&& tmp326 = n3ds_cfgStereoCamera;
for(int64_t tmp327=0;tmp327<len(tmp326);++tmp327){
auto i=tmp327;auto f=tmp326[tmp327];{
uint32_t off = cast<uint32_t>((cast<uint32_t>(i) * cast<uint32_t>(4ULL)));
if ((cast<uint32_t>((off + cast<uint32_t>(4ULL))) <= size)) {
le_PutUint32(sub(buf,off,len(buf)),go_math_Float32bits(f));
}
}
}}
break;}
default:{
if (m->Verbose) {
go_fmt_Printf(std::string("cfg block 0x%08X (%d bytes) zero-filled\012",40),blkID,size);
}
break;}
}}
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < size);i++){
n3ds_Machine_Write(m,cast<uint32_t>((out + i)),buf[i]);
}
}}
}
// tools/platform/n3ds/ipc_services.go:577:1
bool n3ds_Machine_ipcFS(n3ds_Machine* m,n3ds_ipcHeader hdr){
{
{
switch(hdr.Command){
case cast<uint16_t>(2145ULL):{
n3ds_Machine_ipcReply(m,hdr.Command,Slice<uint32_t>{});
return true;
break;}
case cast<uint16_t>(2146ULL):case cast<uint16_t>(2147ULL):{
n3ds_Machine_ipcReply(m,hdr.Command,Slice<uint32_t>{cast<uint32_t>(0ULL)});
return true;
break;}
case cast<uint16_t>(2049ULL):{
n3ds_Machine_ipcReply(m,hdr.Command,Slice<uint32_t>{});
return true;
break;}
case cast<uint16_t>(2050ULL):case cast<uint16_t>(2051ULL):{
return n3ds_Machine_fsOpenFile(m,hdr);
break;}
case cast<uint16_t>(2052ULL):{
return n3ds_Machine_fsDeleteFile(m,hdr);
break;}
case cast<uint16_t>(2056ULL):{
return n3ds_Machine_fsCreateFile(m,hdr);
break;}
case cast<uint16_t>(2057ULL):{
n3ds_Machine_ipcReply(m,hdr.Command,Slice<uint32_t>{});
return true;
break;}
case cast<uint16_t>(2059ULL):{
return n3ds_Machine_fsOpenDirectory(m,hdr);
break;}
case cast<uint16_t>(2060ULL):{
return n3ds_Machine_fsOpenArchive(m,hdr);
break;}
case cast<uint16_t>(2062ULL):{
removeKey(m->fsArchives,n3ds_Machine_ipcArg(m,cast<int64_t>(1ULL)));
n3ds_Machine_ipcReply(m,hdr.Command,Slice<uint32_t>{});
return true;
break;}
case cast<uint16_t>(2068ULL):case cast<uint16_t>(2071ULL):case cast<uint16_t>(2129ULL):{
n3ds_Machine_ipcReply(m,hdr.Command,Slice<uint32_t>{cast<uint32_t>(0ULL),cast<uint32_t>(0ULL)});
return true;
break;}
case cast<uint16_t>(2061ULL):{
n3ds_Machine_ipcReply(m,hdr.Command,Slice<uint32_t>{});
return true;
break;}
case cast<uint16_t>(2117ULL):{
if ((!m->saveFormatted)) {
n3ds_Machine_WriteWord(m,n3ds_Machine_cmdBuf(m),cast<uint32_t>((shl<uint32_t>(cast<uint32_t>(hdr.Command),cast<int64_t>(16ULL)) | cast<uint32_t>(64ULL))));
n3ds_Machine_WriteWord(m,cast<uint32_t>((n3ds_Machine_cmdBuf(m) + cast<uint32_t>(4ULL))),cast<uint32_t>(3363849556ULL));
return true;
}
n3ds_Machine_ipcReply(m,hdr.Command,Slice<uint32_t>{m->saveFormatInfo[cast<int64_t>(0ULL)],m->saveFormatInfo[cast<int64_t>(1ULL)],m->saveFormatInfo[cast<int64_t>(2ULL)],m->saveFormatInfo[cast<int64_t>(3ULL)]});
return true;
break;}
case cast<uint16_t>(2124ULL):{
m->saveFormatted = true;
m->saveFormatInfo = std::array<uint32_t,4>{n3ds_Machine_ipcArg(m,cast<int64_t>(4ULL)),n3ds_Machine_ipcArg(m,cast<int64_t>(5ULL)),n3ds_Machine_ipcArg(m,cast<int64_t>(6ULL)),cast<uint32_t>((n3ds_Machine_ipcArg(m,cast<int64_t>(9ULL)) & cast<uint32_t>(255ULL)))};
m->saveFiles = Map<std::string,Slice<uint8_t>>{};
n3ds_Machine_ipcReply(m,hdr.Command,Slice<uint32_t>{});
return true;
break;}
}}
arm_CPU_Halt(m->CPU,std::string("fs command 0x%04X unimplemented at 0x%08X after %d instructions",63),hdr.Command,arm_CPU_PC(m->CPU),m->CPU->Instrs);
return true;
}
}
// tools/platform/n3ds/ipc_services.go:656:1
std::tuple<int64_t,int64_t> n3ds_Machine_FrameStats(n3ds_Machine* m){
int64_t submitted{};
int64_t swapped{};
{
return {m->framesSubmitted,m->framesSwapped};
}
}
// tools/platform/n3ds/ipc_services.go:662:1
int64_t n3ds_Machine_DisplayTransfers(n3ds_Machine* m){
{
return m->displayTransfers;
}
}
// tools/platform/n3ds/ipc_services.go:670:1
bool n3ds_Machine_ipcErr(n3ds_Machine* m,n3ds_ipcHeader hdr){
{
if ((hdr.Command == cast<uint16_t>(1ULL))) {
uint32_t errType = n3ds_Machine_ipcArg(m,cast<int64_t>(1ULL));
uint32_t code = n3ds_Machine_ipcArg(m,cast<int64_t>(3ULL));
uint32_t pc = n3ds_Machine_ipcArg(m,cast<int64_t>(5ULL));
go_fmt_Printf(std::string("err:f ThrowFatalError type=0x%X resultCode=0x%08X pc=0x%08X; cmdbuf:",68),errType,code,pc);
{int64_t i = cast<int64_t>(1ULL);for (;(i < cast<int64_t>(16ULL));i++){
go_fmt_Printf(std::string(" %08X",5),n3ds_Machine_ReadWord(m,cast<uint32_t>((n3ds_Machine_cmdBuf(m) + cast<uint32_t>((cast<uint32_t>(i) * cast<uint32_t>(4ULL)))))));
}
}go_fmt_Println();
}
n3ds_Machine_ipcReply(m,hdr.Command,Slice<uint32_t>{});
return true;
}
}
// tools/platform/n3ds/machine.go:58:1
bool n3ds_memRegion_contains(n3ds_memRegion* r,uint32_t a){
{
return ((a >= r->base) && (cast<uint32_t>((a - r->base)) < cast<uint32_t>(len(r->data))));
}
}
// tools/platform/n3ds/machine.go:65:1
Slice<uint8_t> n3ds_Machine_ReadBytes(n3ds_Machine* m,uint32_t addr,uint32_t length){
{
Slice<uint8_t> b = Slice<uint8_t>::make(length);
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < length);i++){
b[i] = n3ds_Machine_Read(m,cast<uint32_t>((addr + i)));
}
}return b;
}
}
// tools/platform/n3ds/machine.go:77:1
Slice<uint32_t> n3ds_Machine_FindBytes(n3ds_Machine* m,Slice<uint8_t> pat){
{
Slice<uint32_t> hits={};
if ((len(pat) == cast<int64_t>(0ULL))) {
return hits;
}
{auto&& tmp328 = m->regions;
for(int64_t tmp329=0;tmp329<len(tmp328);++tmp329){
auto r=tmp328[tmp329];Slice<uint8_t> d = r->data;
{int64_t i = cast<int64_t>(0ULL);for (;(cast<int64_t>((i + len(pat))) <= len(d));i++){
if (((d[i] == pat[cast<int64_t>(0ULL)]) && (cast<std::string>(sub(d,i,cast<int64_t>((i + len(pat))))) == cast<std::string>(pat)))) {
hits = append(hits,cast<uint32_t>((r->base + cast<uint32_t>(i))));
if ((len(hits) >= cast<int64_t>(256ULL))) {
return hits;
}
}
}
}}}
return hits;
}
}
// tools/platform/n3ds/machine.go:316:1
std::tuple<n3ds_Machine*,Error> n3ds_NewMachine(Slice<uint8_t> img){
{
auto tmp330 = n3ds_ParseNCSD(img);
n3ds_NCSD* ncsd = std::get<0>(tmp330);
Error err = std::get<1>(tmp330);
if (bool(err)) {
return {{},err};
}
auto tmp331 = n3ds_NCSD_Executable(ncsd);
n3ds_NCCH* cxi = std::get<0>(tmp331);
err = std::get<1>(tmp331);
if (bool(err)) {
return {{},err};
}
if (n3ds_NCCH_Encrypted(cxi)) {
return {{},go_fmt_Errorf(std::string("n3ds: partition 0 is encrypted (%s); supply a decrypted dump",60),n3ds_NCCH_CryptoMethod(cxi))};
}
auto tmp332 = n3ds_NCCH_ExHeader(cxi);
n3ds_ExHeader* ex = std::get<0>(tmp332);
err = std::get<1>(tmp332);
if (bool(err)) {
return {{},err};
}
auto tmp333 = n3ds_NCCH_ExeFS(cxi);
n3ds_ExeFS* efs = std::get<0>(tmp333);
err = std::get<1>(tmp333);
if (bool(err)) {
return {{},err};
}
auto tmp334 = n3ds_ExeFS_Code(efs,ex);
Slice<uint8_t> code = std::get<0>(tmp334);
err = std::get<1>(tmp334);
if (bool(err)) {
return {{},err};
}
n3ds_Machine* m = arenaNew(n3ds_Machine{{},{},{},{},{},{},ex->Text.Address,{},{},{},{},Map<uint32_t,n3ds_kobject*>{},cast<uint32_t>(65536ULL),Map<uint32_t,std::string>{},Map<uint32_t,std::string>{},{},{},{},{},{},{},{},{},{},{},{},Map<uint32_t,n3ds_fsFile*>{},Map<uint32_t,n3ds_fsDir*>{},Map<uint32_t,uint32_t>{},Map<std::string,Slice<uint8_t>>{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},Map<uint32_t,bool>{},{},{},{},{},{},{},cxi->ProgramID});
m->gpu = n3ds_newGPU(m);
auto tmp335 = n3ds_NCCH_RomFS(cxi);
m->romfs = std::get<0>(tmp335);
auto tmp336 = n3ds_NCCH_RomFSBytes(cxi);
m->romfsRaw = std::get<0>(tmp336);
uint32_t total = cast<uint32_t>((n3ds_ExHeader_CodeSize(ex) + ex->BSSSize));
Slice<uint8_t> codeMem = Slice<uint8_t>::make(total);
gcopy(codeMem,code);
m->codeReg = n3ds_Machine_mapRegion(m,std::string("code",4),ex->Text.Address,codeMem);
uint32_t stackSize = ((cast<uint32_t>((cast<uint32_t>((ex->StackSize + cast<uint32_t>(4096ULL))) - cast<uint32_t>(1ULL)))) & ~(cast<uint32_t>(4095ULL)));
if ((stackSize == cast<uint32_t>(0ULL))) {
stackSize = cast<uint32_t>(16384ULL);
}
m->stackReg = n3ds_Machine_mapRegion(m,std::string("stack",5),cast<uint32_t>((cast<uint32_t>(268435456ULL) - stackSize)),Slice<uint8_t>::make(stackSize));
auto tmp337 = std::make_tuple(cast<uint32_t>(134217728ULL),cast<uint32_t>(335544320ULL));
m->heapPtr = std::get<0>(tmp337);
m->linearPtr = std::get<1>(tmp337);
m->heapReg = n3ds_Machine_mapRegion(m,std::string("heap",4),cast<uint32_t>(134217728ULL),{});
m->linearReg = n3ds_Machine_mapRegion(m,std::string("linear",6),cast<uint32_t>(335544320ULL),{});
n3ds_Machine_mapRegion(m,std::string("config",6),cast<uint32_t>(536346624ULL),n3ds_Machine_buildConfigPage(m));
m->tlsReg = n3ds_Machine_mapRegion(m,std::string("tls",3),cast<uint32_t>(536354816ULL),Slice<uint8_t>::make(cast<int64_t>(4096ULL)));
n3ds_Machine_mapRegion(m,std::string("vram",4),cast<uint32_t>(520093696ULL),Slice<uint8_t>::make(cast<int64_t>(6291456ULL)));
n3ds_Machine_mapRegion(m,std::string("dspram",6),cast<uint32_t>(535822336ULL),Slice<uint8_t>::make(cast<int64_t>(524288ULL)));
m->dsp.IntEvents = Map<uint32_t,uint32_t>{};
arm_CPU* cpu = arm_NewCPU(m);
cpu->Arch = cast<arm_Variant>(1ULL);
arm_CPU_Reset(cpu);
cpu->Mode = cast<uint32_t>(31ULL);
cpu->R[cast<int64_t>(13ULL)] = cast<uint32_t>(268435456ULL);
cpu->R[cast<int64_t>(15ULL)] = m->entry;
cpu->SWI = [=](auto...args){return n3ds_Machine_handleSVC(m,args...);};
cpu->Coproc = [=](auto...args){return n3ds_Machine_handleCP15(m,args...);};
m->CPU = cpu;
m->nextTLS = cast<uint32_t>(536358912ULL);
n3ds_thread* main = arenaNew(n3ds_thread{cast<uint32_t>(1ULL),n3ds_Machine_newHandle(m,std::string("thread",6),false),{},cast<uint32_t>(536354816ULL),{},cast<int32_t>(48ULL),cast<n3ds_threadState>(0ULL),{},{},{},{},{}});
main->ctx = (*cpu);
m->nextThread = cast<uint32_t>(2ULL);
m->threads = Slice<n3ds_thread*>{main};
m->curThread = main;
m->handles[main->handle]->thread = main;
return {m,{}};
}
}
// tools/platform/n3ds/machine.go:426:1
n3ds_memRegion* n3ds_Machine_mapRegion(n3ds_Machine* m,std::string name,uint32_t base,Slice<uint8_t> data){
{
n3ds_memRegion* r = arenaNew(n3ds_memRegion{name,base,data});
m->regions = append(m->regions,r);
n3ds_Machine_indexRegion(m,r);
return r;
}
}
// tools/platform/n3ds/machine.go:437:1
Slice<uint8_t> n3ds_Machine_buildConfigPage(n3ds_Machine* m){
{
Slice<uint8_t> p = Slice<uint8_t>::make(cast<int64_t>(4096ULL));
le_PutUint32(sub(p,cast<int64_t>(0ULL),len(p)),cast<uint32_t>(36831232ULL));
le_PutUint32(sub(p,cast<int64_t>(64ULL),len(p)),cast<uint32_t>(67108864ULL));
return p;
}
}
// tools/platform/n3ds/machine.go:476:1
void n3ds_Machine_indexRegion(n3ds_Machine* m,n3ds_memRegion* r){
{
if ((!m->pages)) {
m->pages = Slice<n3ds_memRegion*>::make(cast<int64_t>(1048576ULL));
}
if ((len(r->data) == cast<int64_t>(0ULL))) {
return ;
}
uint32_t lo = shr<uint32_t>(r->base,cast<int64_t>(12ULL));
uint32_t hi = shr<uint32_t>((cast<uint32_t>((r->base + cast<uint32_t>(len(r->data))))),cast<int64_t>(12ULL));
if ((cast<uint32_t>((r->base & cast<uint32_t>(4095ULL))) != cast<uint32_t>(0ULL))) {
lo++;
}
{uint32_t p = lo;for (;(p < hi);p++){
m->pages[p] = r;
}
}}
}
// tools/platform/n3ds/machine.go:494:1
void n3ds_Machine_clearPages(n3ds_Machine* m){
{
m->pages = {};
}
}
// tools/platform/n3ds/machine.go:496:1
n3ds_memRegion* n3ds_Machine_regionOf(n3ds_Machine* m,uint32_t a){
{
if (bool(m->pages)) {
{
n3ds_memRegion* r = m->pages[shr<uint32_t>(a,cast<int64_t>(12ULL))];
if (bool(r)) {
return r;
}
}
}
{auto&& tmp338 = m->regions;
for(int64_t tmp339=0;tmp339<len(tmp338);++tmp339){
auto r=tmp338[tmp339];if (n3ds_memRegion_contains(r,a)) {
return r;
}
}}
return {};
}
}
// tools/platform/n3ds/machine.go:511:1
uint8_t n3ds_Machine_Read(n3ds_Machine* m,uint32_t a){
{
if ((((m->HidTrace && (m->hidSharedAddr != cast<uint32_t>(0ULL))) && (a >= m->hidSharedAddr)) && (a < cast<uint32_t>((m->hidSharedAddr + cast<uint32_t>(4096ULL)))))) {
if ((!m->hidReadHist)) {
m->hidReadHist = Map<uint32_t,int64_t>{};
m->hidReadPC = Map<uint32_t,uint32_t>{};
}
uint32_t off = ((cast<uint32_t>((a - m->hidSharedAddr))) & ~(cast<uint32_t>(3ULL)));
m->hidReadHist[off]++;
m->hidReadPC[off] = arm_CPU_PC(m->CPU);
}
{
n3ds_memRegion* r = n3ds_Machine_regionOf(m,a);
if (bool(r)) {
uint8_t v = r->data[cast<uint32_t>((a - r->base))];
if (((bool(m->OnRead) && (a >= m->RWatchLo)) && (a < m->RWatchHi))) {
m->OnRead(a,cast<uint32_t>(v),arm_CPU_PC(m->CPU));
}
return v;
}
}
if (m->Verbose) {
go_fmt_Printf(std::string("  [unmapped read  0x%08X pc=0x%08X]\012",36),a,arm_CPU_PC(m->CPU));
}
return cast<uint8_t>(0ULL);
}
}
// tools/platform/n3ds/machine.go:534:1
void n3ds_Machine_Write(n3ds_Machine* m,uint32_t a,uint8_t v){
{
{
n3ds_memRegion* r = n3ds_Machine_regionOf(m,a);
if (bool(r)) {
r->data[cast<uint32_t>((a - r->base))] = v;
if (((bool(m->OnWrite) && (a >= m->WatchLo)) && (a < m->WatchHi))) {
m->OnWrite(a,cast<uint32_t>(v),arm_CPU_PC(m->CPU));
}
return ;
}
}
if (m->Verbose) {
go_fmt_Printf(std::string("  [unmapped write 0x%08X=0x%02X pc=0x%08X]\012",43),a,v,arm_CPU_PC(m->CPU));
}
}
}
// tools/platform/n3ds/machine.go:549:1
uint32_t n3ds_Machine_ReadWord(n3ds_Machine* m,uint32_t a){
{
return cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(n3ds_Machine_Read(m,a)) | shl<uint32_t>(cast<uint32_t>(n3ds_Machine_Read(m,cast<uint32_t>((a + cast<uint32_t>(1ULL))))),cast<int64_t>(8ULL)))) | shl<uint32_t>(cast<uint32_t>(n3ds_Machine_Read(m,cast<uint32_t>((a + cast<uint32_t>(2ULL))))),cast<int64_t>(16ULL)))) | shl<uint32_t>(cast<uint32_t>(n3ds_Machine_Read(m,cast<uint32_t>((a + cast<uint32_t>(3ULL))))),cast<int64_t>(24ULL))));
}
}
// tools/platform/n3ds/machine.go:553:1
void n3ds_Machine_WriteWord(n3ds_Machine* m,uint32_t a,uint32_t v){
{
n3ds_Machine_Write(m,a,cast<uint8_t>(v));
n3ds_Machine_Write(m,cast<uint32_t>((a + cast<uint32_t>(1ULL))),cast<uint8_t>(shr<uint32_t>(v,cast<int64_t>(8ULL))));
n3ds_Machine_Write(m,cast<uint32_t>((a + cast<uint32_t>(2ULL))),cast<uint8_t>(shr<uint32_t>(v,cast<int64_t>(16ULL))));
n3ds_Machine_Write(m,cast<uint32_t>((a + cast<uint32_t>(3ULL))),cast<uint8_t>(shr<uint32_t>(v,cast<int64_t>(24ULL))));
}
}
// tools/platform/n3ds/machine.go:573:1
std::tuple<Slice<uint8_t>,uint32_t> n3ds_Machine_directRange(n3ds_Machine* m,uint32_t addr,uint32_t size){
{
if (((bool(m->OnRead) || bool(m->OnWrite)) || m->HidTrace)) {
return {{},cast<uint32_t>(0ULL)};
}
n3ds_memRegion* r = n3ds_Machine_regionOf(m,addr);
if ((!r)) {
return {{},cast<uint32_t>(0ULL)};
}
uint32_t off = cast<uint32_t>((addr - r->base));
if ((cast<uint64_t>((cast<uint64_t>(off) + cast<uint64_t>(size))) > cast<uint64_t>(len(r->data)))) {
return {{},cast<uint32_t>(0ULL)};
}
return {r->data,off};
}
}
// tools/platform/n3ds/machine.go:595:1
void n3ds_Machine_copyRange(n3ds_Machine* m,uint32_t dst,uint32_t src,uint32_t size){
{
if ((size == cast<uint32_t>(0ULL))) {
return ;
}
auto tmp340 = n3ds_Machine_directRange(m,dst,size);
Slice<uint8_t> d = std::get<0>(tmp340);
uint32_t doff = std::get<1>(tmp340);
auto tmp341 = n3ds_Machine_directRange(m,src,size);
Slice<uint8_t> s = std::get<0>(tmp341);
uint32_t soff = std::get<1>(tmp341);
if ((bool(d) && bool(s))) {
gcopy(sub(d,doff,cast<uint32_t>((doff + size))),sub(s,soff,cast<uint32_t>((soff + size))));
return ;
}
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < size);i++){
n3ds_Machine_Write(m,cast<uint32_t>((dst + i)),n3ds_Machine_Read(m,cast<uint32_t>((src + i))));
}
}}
}
// tools/platform/n3ds/message.go:20:1
std::tuple<Slice<uint8_t>,Error> n3ds_Yaz0(Slice<uint8_t> b){
{
if (((len(b) < cast<int64_t>(16ULL)) || (cast<std::string>(sub(b,0,cast<int64_t>(4ULL))) != std::string("Yaz0",4)))) {
return {{},go_fmt_Errorf(std::string("not a Yaz0 stream",17))};
}
uint32_t size = be_Uint32(sub(b,cast<int64_t>(4ULL),len(b)));
Slice<uint8_t> out = Slice<uint8_t>::make(cast<int64_t>(0ULL),size);
int64_t src = cast<int64_t>(16ULL);
{;for (;(cast<uint32_t>(len(out)) < size);){
if ((src >= len(b))) {
return {{},go_fmt_Errorf(std::string("Yaz0: truncated at %d/%d",24),len(out),size)};
}
uint8_t ctrl = b[src];
src++;
{int64_t i = cast<int64_t>(0ULL);for (;((i < cast<int64_t>(8ULL)) && (cast<uint32_t>(len(out)) < size));i++){
if ((cast<uint8_t>((ctrl & cast<uint8_t>(128ULL))) != cast<uint8_t>(0ULL))) {
out = append(out,b[src]);
src++;
}
else {
auto tmp342 = std::make_tuple(b[src],b[cast<int64_t>((src + cast<int64_t>(1ULL)))]);
uint8_t b1 = std::get<0>(tmp342);
uint8_t b2 = std::get<1>(tmp342);
src += cast<int64_t>(2ULL);
int64_t dist = cast<int64_t>(((cast<int64_t>((shl<int64_t>(cast<int64_t>(cast<uint8_t>((b1 & cast<uint8_t>(15ULL)))),cast<int64_t>(8ULL)) | cast<int64_t>(b2)))) + cast<int64_t>(1ULL)));
int64_t n = cast<int64_t>(shr<uint8_t>(b1,cast<int64_t>(4ULL)));
if ((n == cast<int64_t>(0ULL))) {
n = cast<int64_t>((cast<int64_t>(b[src]) + cast<int64_t>(18ULL)));
src++;
}
else {
n += cast<int64_t>(2ULL);
}
int64_t start = cast<int64_t>((len(out) - dist));
{int64_t j = cast<int64_t>(0ULL);for (;(j < n);j++){
out = append(out,out[cast<int64_t>((start + j))]);
}
}}
ctrl = shl<uint8_t>(ctrl,cast<int64_t>(1ULL));
}
}}
}return {out,{}};
}
}
// tools/platform/n3ds/message.go:63:1
std::tuple<Slice<Slice<uint8_t>>,Error> n3ds_NARCFiles(Slice<uint8_t> d){
{
if (((len(d) < cast<int64_t>(16ULL)) || (cast<std::string>(sub(d,0,cast<int64_t>(4ULL))) != std::string("NARC",4)))) {
return {{},go_fmt_Errorf(std::string("not a NARC",10))};
}
binary_littleEndian le = go_binary_LittleEndian;
if ((cast<std::string>(sub(d,cast<int64_t>(16ULL),cast<int64_t>(20ULL))) != std::string("BTAF",4))) {
return {{},go_fmt_Errorf(std::string("NARC: BTAF not at 0x10",22))};
}
uint32_t btafSize = le_Uint32(sub(d,cast<int64_t>(20ULL),len(d)));
uint32_t n = le_Uint32(sub(d,cast<int64_t>(24ULL),len(d)));
Slice<std::array<uint32_t,2>> fat = Slice<std::array<uint32_t,2>>::make(n);
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < n);i++){
uint32_t off = cast<uint32_t>((cast<uint32_t>(28ULL) + cast<uint32_t>((i * cast<uint32_t>(8ULL)))));
fat[i] = std::array<uint32_t,2>{le_Uint32(sub(d,off,len(d))),le_Uint32(sub(d,cast<uint32_t>((off + cast<uint32_t>(4ULL))),len(d)))};
}
}uint32_t btnfOff = cast<uint32_t>((cast<uint32_t>(16ULL) + btafSize));
if ((cast<std::string>(sub(d,btnfOff,cast<uint32_t>((btnfOff + cast<uint32_t>(4ULL))))) != std::string("BTNF",4))) {
return {{},go_fmt_Errorf(std::string("NARC: BTNF not after BTAF",25))};
}
uint32_t gmifOff = cast<uint32_t>((btnfOff + le_Uint32(sub(d,cast<uint32_t>((btnfOff + cast<uint32_t>(4ULL))),len(d)))));
if ((cast<std::string>(sub(d,gmifOff,cast<uint32_t>((gmifOff + cast<uint32_t>(4ULL))))) != std::string("GMIF",4))) {
return {{},go_fmt_Errorf(std::string("NARC: GMIF not after BTNF",25))};
}
uint32_t gmifData = cast<uint32_t>((gmifOff + cast<uint32_t>(8ULL)));
Slice<Slice<uint8_t>> files = Slice<Slice<uint8_t>>::make(n);
{auto&& tmp343 = fat;
for(int64_t tmp344=0;tmp344<len(tmp343);++tmp344){
auto i=tmp344;auto ext=tmp343[tmp344];files[i] = sub(d,cast<uint32_t>((gmifData + ext[cast<int64_t>(0ULL)])),cast<uint32_t>((gmifData + ext[cast<int64_t>(1ULL)])));
}}
return {files,{}};
}
}
// tools/platform/n3ds/message.go:105:1
std::tuple<Slice<n3ds_MSBTMessage>,Error> n3ds_ParseMSBT(Slice<uint8_t> fd){
{
if (((len(fd) < cast<int64_t>(32ULL)) || (cast<std::string>(sub(fd,0,cast<int64_t>(8ULL))) != std::string("MsgStdBn",8)))) {
return {{},go_fmt_Errorf(std::string("not an MSBT",11))};
}
binary_littleEndian le = go_binary_LittleEndian;
int64_t nsec = cast<int64_t>(le_Uint16(sub(fd,cast<int64_t>(14ULL),len(fd))));
Map<int64_t,std::string> labels = Map<int64_t,std::string>{};
Slice<std::string> texts={};
int64_t p = cast<int64_t>(32ULL);
{int64_t s = cast<int64_t>(0ULL);for (;((s < nsec) && (cast<int64_t>((p + cast<int64_t>(16ULL))) <= len(fd)));s++){
std::string mag = cast<std::string>(sub(fd,p,cast<int64_t>((p + cast<int64_t>(4ULL)))));
int64_t size = cast<int64_t>(le_Uint32(sub(fd,cast<int64_t>((p + cast<int64_t>(4ULL))),len(fd))));
Slice<uint8_t> body = sub(fd,cast<int64_t>((p + cast<int64_t>(16ULL))),cast<int64_t>((cast<int64_t>((p + cast<int64_t>(16ULL))) + size)));
{
auto tmp346=mag;
if (tmp346==(std::string("TXT2",4))){
int64_t cnt = cast<int64_t>(le_Uint32(sub(body,cast<int64_t>(0ULL),len(body))));
{int64_t i = cast<int64_t>(0ULL);for (;(i < cnt);i++){
int64_t start = cast<int64_t>(le_Uint32(sub(body,cast<int64_t>((cast<int64_t>(4ULL) + cast<int64_t>((i * cast<int64_t>(4ULL))))),len(body))));
int64_t end = len(body);
if ((cast<int64_t>((i + cast<int64_t>(1ULL))) < cnt)) {
end = cast<int64_t>(le_Uint32(sub(body,cast<int64_t>((cast<int64_t>(4ULL) + cast<int64_t>(((cast<int64_t>((i + cast<int64_t>(1ULL)))) * cast<int64_t>(4ULL))))),len(body))));
}
texts = append(texts,n3ds_decodeUTF16(sub(body,start,end)));
}
}}
else if (tmp346==(std::string("LBL1",4))){
int64_t ngrp = cast<int64_t>(le_Uint32(sub(body,cast<int64_t>(0ULL),len(body))));
{int64_t g = cast<int64_t>(0ULL);for (;(g < ngrp);g++){
int64_t num = cast<int64_t>(le_Uint32(sub(body,cast<int64_t>((cast<int64_t>(4ULL) + cast<int64_t>((g * cast<int64_t>(8ULL))))),len(body))));
int64_t off = cast<int64_t>(le_Uint32(sub(body,cast<int64_t>((cast<int64_t>((cast<int64_t>(4ULL) + cast<int64_t>((g * cast<int64_t>(8ULL))))) + cast<int64_t>(4ULL))),len(body))));
int64_t q = off;
{int64_t k = cast<int64_t>(0ULL);for (;(k < num);k++){
int64_t ln = cast<int64_t>(body[q]);
q++;
std::string name = cast<std::string>(sub(body,q,cast<int64_t>((q + ln))));
q += ln;
int64_t idx = cast<int64_t>(le_Uint32(sub(body,q,len(body))));
q += cast<int64_t>(4ULL);
labels[idx] = name;
}
}}
}}
}
tmp345:;
p += cast<int64_t>((cast<int64_t>(16ULL) + size));
p = ((cast<int64_t>((p + cast<int64_t>(15ULL)))) & ~(cast<int64_t>(15ULL)));
}
}Slice<n3ds_MSBTMessage> out = Slice<n3ds_MSBTMessage>::make(len(texts));
{auto&& tmp347 = texts;
for(int64_t tmp348=0;tmp348<len(tmp347);++tmp348){
auto i=tmp348;auto t=tmp347[tmp348];out[i] = n3ds_MSBTMessage{i,get(labels,i),t};
}}
return {out,{}};
}
}
// tools/platform/n3ds/message.go:156:1
std::string n3ds_decodeUTF16(Slice<uint8_t> b){
{
Slice<uint16_t> u = Slice<uint16_t>::make(cast<int64_t>(0ULL),divi<int64_t>(len(b),cast<int64_t>(2ULL)));
{int64_t i = cast<int64_t>(0ULL);for (;(cast<int64_t>((i + cast<int64_t>(1ULL))) < len(b));i += cast<int64_t>(2ULL)){
uint16_t v = cast<uint16_t>((cast<uint16_t>(b[i]) | shl<uint16_t>(cast<uint16_t>(b[cast<int64_t>((i + cast<int64_t>(1ULL)))]),cast<int64_t>(8ULL))));
if ((v == cast<uint16_t>(0ULL))) {
break;
}
u = append(u,v);
}
}return cast<std::string>(go_utf16_Decode(u));
}
}
// tools/platform/n3ds/ncch.go:34:1
bool n3ds_region_empty(n3ds_region r){
{
return (r.Size == cast<int64_t>(0ULL));
}
}
// tools/platform/n3ds/ncch.go:59:1
std::tuple<n3ds_NCCH*,Error> n3ds_ParseNCCH(Slice<uint8_t> part){
{
if ((len(part) < cast<int64_t>(768ULL))) {
return {{},go_fmt_Errorf(std::string("n3ds: partition too short for an NCCH header (%d bytes)",55),len(part))};
}
Slice<uint8_t> h = sub(part,cast<int64_t>(256ULL),len(part));
if ((cast<std::string>(sub(h,cast<int64_t>(0ULL),cast<int64_t>(4ULL))) != n3ds_ncchMagic)) {
return {{},go_fmt_Errorf(std::string("n3ds: not an NCCH partition: magic %q, want %q",46),sub(h,cast<int64_t>(0ULL),cast<int64_t>(4ULL)),n3ds_ncchMagic)};
}
n3ds_NCCH* c = arenaNew(n3ds_NCCH{part,{},{},{},{},{},{},{},{},{},{},{},{},{}});
gcopy(sub(c->Flags,0,len(c->Flags)),sub(h,cast<int64_t>(136ULL),cast<int64_t>(144ULL)));
c->MediaUnitSize = shl<int64_t>(cast<int64_t>(512ULL),c->Flags[cast<int64_t>(6ULL)]);
c->ContentSize = cast<int64_t>((cast<int64_t>(le_Uint32(sub(h,cast<int64_t>(4ULL),len(h)))) * c->MediaUnitSize));
c->PartitionID = le_Uint64(sub(h,cast<int64_t>(8ULL),len(h)));
c->MakerCode = cast<std::string>(n3ds_trimNul(sub(h,cast<int64_t>(16ULL),cast<int64_t>(18ULL))));
c->Version = le_Uint16(sub(h,cast<int64_t>(18ULL),len(h)));
c->ProgramID = le_Uint64(sub(h,cast<int64_t>(24ULL),len(h)));
c->ProductCode = cast<std::string>(n3ds_trimNul(sub(h,cast<int64_t>(80ULL),cast<int64_t>(96ULL))));
c->ExHeaderSize = cast<int64_t>(le_Uint32(sub(h,cast<int64_t>(128ULL),len(h))));
auto rd = [&](int64_t off)->n3ds_region{
return n3ds_region{cast<int64_t>((cast<int64_t>(le_Uint32(sub(h,off,len(h)))) * c->MediaUnitSize)),cast<int64_t>((cast<int64_t>(le_Uint32(sub(h,cast<int64_t>((off + cast<int64_t>(4ULL))),len(h)))) * c->MediaUnitSize))};
}
;
c->PlainRegion = rd(cast<int64_t>(144ULL));
c->LogoRegion = rd(cast<int64_t>(152ULL));
c->ExeFSRegion = rd(cast<int64_t>(160ULL));
c->RomFSRegion = rd(cast<int64_t>(176ULL));
{auto&& tmp349 = Slice<Anon53>{Anon53{std::string("plain",5),c->PlainRegion},Anon53{std::string("logo",4),c->LogoRegion},Anon53{std::string("exefs",5),c->ExeFSRegion},Anon53{std::string("romfs",5),c->RomFSRegion}};
for(int64_t tmp350=0;tmp350<len(tmp349);++tmp350){
auto r=tmp349[tmp350];if (((!n3ds_region_empty(r.reg)) && (cast<int64_t>((r.reg.Offset + r.reg.Size)) > cast<int64_t>(len(part))))) {
return {{},go_fmt_Errorf(std::string("n3ds: %s region [0x%x+0x%x] runs past the partition end (0x%x)",62),r.name,r.reg.Offset,r.reg.Size,len(part))};
}
}}
return {c,{}};
}
}
// tools/platform/n3ds/ncch.go:106:1
bool n3ds_NCCH_Encrypted(n3ds_NCCH* c){
{
return (cast<uint8_t>((c->Flags[cast<int64_t>(7ULL)] & cast<uint8_t>(4ULL))) == cast<uint8_t>(0ULL));
}
}
// tools/platform/n3ds/ncch.go:109:1
std::string n3ds_NCCH_CryptoMethod(n3ds_NCCH* c){
{
{
switch(c->Flags[cast<int64_t>(3ULL)]){
case cast<uint8_t>(0ULL):{
return std::string("standard",8);
break;}
case cast<uint8_t>(1ULL):{
return std::string("7.x",3);
break;}
case cast<uint8_t>(10ULL):{
return std::string("Secure3",7);
break;}
case cast<uint8_t>(11ULL):{
return std::string("Secure4",7);
break;}
}}
return go_fmt_Sprintf(std::string("unknown(0x%02x)",15),c->Flags[cast<int64_t>(3ULL)]);
}
}
// tools/platform/n3ds/ncch.go:124:1
std::string n3ds_NCCH_ContentType(n3ds_NCCH* c){
{
uint8_t f = c->Flags[cast<int64_t>(5ULL)];
Slice<Anon54> names = Slice<Anon54>{Anon54{cast<uint8_t>(1ULL),std::string("Data",4)},Anon54{cast<uint8_t>(2ULL),std::string("Executable",10)},Anon54{cast<uint8_t>(4ULL),std::string("SystemUpdate",12)},Anon54{cast<uint8_t>(8ULL),std::string("Manual",6)},Anon54{cast<uint8_t>(16ULL),std::string("Child",5)},Anon54{cast<uint8_t>(32ULL),std::string("Trial",5)}};
std::string out = std::string("",0);
{auto&& tmp351 = names;
for(int64_t tmp352=0;tmp352<len(tmp351);++tmp352){
auto n=tmp351[tmp352];if ((cast<uint8_t>((f & n.bit)) != cast<uint8_t>(0ULL))) {
if ((out != std::string("",0))) {
out += std::string("|",1);
}
out += n.name;
}
}}
if ((out == std::string("",0))) {
return go_fmt_Sprintf(std::string("none(0x%02x)",12),f);
}
return out;
}
}
// tools/platform/n3ds/ncch.go:150:1
Error n3ds_NCCH_checkPlain(n3ds_NCCH* c,std::string what){
{
if (n3ds_NCCH_Encrypted(c)) {
return go_fmt_Errorf(std::string("n3ds: %s is AES-CTR encrypted (crypto method %s); supply a decrypted dump",73),what,n3ds_NCCH_CryptoMethod(c));
}
return {};
}
}
// tools/platform/n3ds/ncch.go:159:1
std::tuple<Slice<uint8_t>,Error> n3ds_NCCH_ExHeaderBytes(n3ds_NCCH* c){
{
{
Error err = n3ds_NCCH_checkPlain(c,std::string("the ExHeader",12));
if (bool(err)) {
return {{},err};
}
}
if ((c->ExHeaderSize == cast<int64_t>(0ULL))) {
return {{},go_fmt_Errorf(std::string("n3ds: partition has no ExHeader",31))};
}
int64_t start = cast<int64_t>(512ULL);
if ((cast<int64_t>((start + c->ExHeaderSize)) > cast<int64_t>(len(c->raw)))) {
return {{},go_fmt_Errorf(std::string("n3ds: ExHeader runs past the partition end",42))};
}
return {sub(c->raw,start,cast<int64_t>((start + c->ExHeaderSize))),{}};
}
}
// tools/platform/n3ds/ncch.go:174:1
std::tuple<n3ds_ExHeader*,Error> n3ds_NCCH_ExHeader(n3ds_NCCH* c){
{
auto tmp353 = n3ds_NCCH_ExHeaderBytes(c);
Slice<uint8_t> b = std::get<0>(tmp353);
Error err = std::get<1>(tmp353);
if (bool(err)) {
return {{},err};
}
return n3ds_ParseExHeader(b);
}
}
// tools/platform/n3ds/ncch.go:183:1
std::tuple<Slice<uint8_t>,Error> n3ds_NCCH_ExeFSBytes(n3ds_NCCH* c){
{
{
Error err = n3ds_NCCH_checkPlain(c,std::string("the ExeFS",9));
if (bool(err)) {
return {{},err};
}
}
if (n3ds_region_empty(c->ExeFSRegion)) {
return {{},go_fmt_Errorf(std::string("n3ds: partition has no ExeFS",28))};
}
n3ds_region r = c->ExeFSRegion;
return {sub(c->raw,r.Offset,cast<int64_t>((r.Offset + r.Size))),{}};
}
}
// tools/platform/n3ds/ncch.go:195:1
std::tuple<n3ds_ExeFS*,Error> n3ds_NCCH_ExeFS(n3ds_NCCH* c){
{
auto tmp354 = n3ds_NCCH_ExeFSBytes(c);
Slice<uint8_t> b = std::get<0>(tmp354);
Error err = std::get<1>(tmp354);
if (bool(err)) {
return {{},err};
}
return n3ds_ParseExeFS(b);
}
}
// tools/platform/n3ds/ncch.go:204:1
std::tuple<Slice<uint8_t>,Error> n3ds_NCCH_RomFSBytes(n3ds_NCCH* c){
{
{
Error err = n3ds_NCCH_checkPlain(c,std::string("the RomFS",9));
if (bool(err)) {
return {{},err};
}
}
if ((n3ds_region_empty(c->RomFSRegion) || (cast<uint8_t>((c->Flags[cast<int64_t>(7ULL)] & cast<uint8_t>(2ULL))) != cast<uint8_t>(0ULL)))) {
return {{},go_fmt_Errorf(std::string("n3ds: partition has no RomFS",28))};
}
n3ds_region r = c->RomFSRegion;
return {sub(c->raw,r.Offset,cast<int64_t>((r.Offset + r.Size))),{}};
}
}
// tools/platform/n3ds/ncch.go:216:1
std::tuple<n3ds_RomFS*,Error> n3ds_NCCH_RomFS(n3ds_NCCH* c){
{
auto tmp355 = n3ds_NCCH_RomFSBytes(c);
Slice<uint8_t> b = std::get<0>(tmp355);
Error err = std::get<1>(tmp355);
if (bool(err)) {
return {{},err};
}
return n3ds_ParseRomFS(b);
}
}
// tools/platform/n3ds/ncch.go:224:1
Slice<uint8_t> n3ds_trimNul(Slice<uint8_t> b){
{
{auto&& tmp356 = b;
for(int64_t tmp357=0;tmp357<len(tmp356);++tmp357){
auto i=tmp357;auto v=tmp356[tmp357];if ((v == cast<uint8_t>(0ULL))) {
return sub(b,0,i);
}
}}
return b;
}
}
// tools/platform/n3ds/ncsd.go:32:1
bool n3ds_Partition_Empty(n3ds_Partition p){
{
return (p.Size == cast<int64_t>(0ULL));
}
}
// tools/platform/n3ds/ncsd.go:46:1
Slice<uint8_t> n3ds_NCSD_Raw(n3ds_NCSD* n){
{
return n->raw;
}
}
// tools/platform/n3ds/ncsd.go:53:1
std::tuple<n3ds_NCSD*,Error> n3ds_ParseNCSD(Slice<uint8_t> img){
{
if ((len(img) < cast<int64_t>(512ULL))) {
return {{},go_fmt_Errorf(std::string("n3ds: image too short for an NCSD header (%d bytes)",51),len(img))};
}
Slice<uint8_t> h = sub(img,cast<int64_t>(256ULL),len(img));
if ((cast<std::string>(sub(h,cast<int64_t>(0ULL),cast<int64_t>(4ULL))) != n3ds_ncsdMagic)) {
return {{},go_fmt_Errorf(std::string("n3ds: not an NCSD image: magic %q at 0x%x, want %q",50),sub(h,cast<int64_t>(0ULL),cast<int64_t>(4ULL)),cast<int64_t>(256ULL),n3ds_ncsdMagic)};
}
n3ds_NCSD* n = arenaNew(n3ds_NCSD{img,{},{},{},{},{}});
gcopy(sub(n->Flags,0,len(n->Flags)),sub(h,cast<int64_t>(136ULL),cast<int64_t>(144ULL)));
n->MediaUnitSize = shl<int64_t>(cast<int64_t>(512ULL),n->Flags[cast<int64_t>(6ULL)]);
n->ImageSize = cast<int64_t>((cast<int64_t>(le_Uint32(sub(h,cast<int64_t>(4ULL),len(h)))) * n->MediaUnitSize));
n->MediaID = le_Uint64(sub(h,cast<int64_t>(8ULL),len(h)));
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(8ULL));i++){
int64_t off = cast<int64_t>((cast<int64_t>(le_Uint32(sub(h,cast<int64_t>((cast<int64_t>(32ULL) + cast<int64_t>((i * cast<int64_t>(8ULL))))),len(h)))) * n->MediaUnitSize));
int64_t size = cast<int64_t>((cast<int64_t>(le_Uint32(sub(h,cast<int64_t>((cast<int64_t>(36ULL) + cast<int64_t>((i * cast<int64_t>(8ULL))))),len(h)))) * n->MediaUnitSize));
n3ds_Partition p = n3ds_Partition{i,off,size,h[cast<int64_t>((cast<int64_t>(16ULL) + i))],h[cast<int64_t>((cast<int64_t>(24ULL) + i))],le_Uint64(sub(h,cast<int64_t>((cast<int64_t>(144ULL) + cast<int64_t>((i * cast<int64_t>(8ULL))))),len(h)))};
if ((!n3ds_Partition_Empty(p))) {
if ((((off < cast<int64_t>(0ULL)) || (size < cast<int64_t>(0ULL))) || (cast<int64_t>((off + size)) > cast<int64_t>(len(img))))) {
return {{},go_fmt_Errorf(std::string("n3ds: partition %d [0x%x+0x%x] runs past the image end (0x%x)",61),i,off,size,len(img))};
}
}
n->Partitions[i] = p;
}
}return {n,{}};
}
}
// tools/platform/n3ds/ncsd.go:90:1
Slice<uint8_t> n3ds_NCSD_Bytes(n3ds_NCSD* n,int64_t i){
{
n3ds_Partition p = n->Partitions[i];
if (n3ds_Partition_Empty(p)) {
return {};
}
return sub(n->raw,p.Offset,cast<int64_t>((p.Offset + p.Size)));
}
}
// tools/platform/n3ds/ncsd.go:99:1
std::tuple<n3ds_NCCH*,Error> n3ds_NCSD_Partition(n3ds_NCSD* n,int64_t i){
{
if (((i < cast<int64_t>(0ULL)) || (i >= cast<int64_t>(8ULL)))) {
return {{},go_fmt_Errorf(std::string("n3ds: partition index %d out of range",37),i)};
}
Slice<uint8_t> b = n3ds_NCSD_Bytes(n,i);
if ((!b)) {
return {{},go_fmt_Errorf(std::string("n3ds: partition %d is empty",27),i)};
}
return n3ds_ParseNCCH(b);
}
}
// tools/platform/n3ds/ncsd.go:111:1
std::tuple<n3ds_NCCH*,Error> n3ds_NCSD_Executable(n3ds_NCSD* n){
{
return n3ds_NCSD_Partition(n,cast<int64_t>(0ULL));
}
}
// tools/platform/n3ds/pica.go:40:1
std::tuple<Slice<n3ds_PICAWrite>,Error> n3ds_DecodePICA(Slice<uint8_t> buf){
{
return n3ds_DecodePICAInto({},buf);
}
}
// tools/platform/n3ds/pica.go:50:1
std::tuple<Slice<n3ds_PICAWrite>,Error> n3ds_DecodePICAInto(Slice<n3ds_PICAWrite> dst,Slice<uint8_t> buf){
{
auto word = [&](uint32_t off)->uint32_t{
return cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(buf[off]) | shl<uint32_t>(cast<uint32_t>(buf[cast<uint32_t>((off + cast<uint32_t>(1ULL)))]),cast<int64_t>(8ULL)))) | shl<uint32_t>(cast<uint32_t>(buf[cast<uint32_t>((off + cast<uint32_t>(2ULL)))]),cast<int64_t>(16ULL)))) | shl<uint32_t>(cast<uint32_t>(buf[cast<uint32_t>((off + cast<uint32_t>(3ULL)))]),cast<int64_t>(24ULL))));
}
;
Slice<n3ds_PICAWrite> ws = dst;
uint32_t off = cast<uint32_t>(0ULL);
{;for (;(cast<uint32_t>((off + cast<uint32_t>(8ULL))) <= cast<uint32_t>(len(buf)));){
auto tmp358 = std::make_tuple(word(off),word(cast<uint32_t>((off + cast<uint32_t>(4ULL)))));
uint32_t param = std::get<0>(tmp358);
uint32_t hdr = std::get<1>(tmp358);
uint16_t reg = cast<uint16_t>(cast<uint32_t>((hdr & cast<uint32_t>(65535ULL))));
uint8_t mask = cast<uint8_t>(cast<uint32_t>((shr<uint32_t>(hdr,cast<int64_t>(16ULL)) & cast<uint32_t>(15ULL))));
uint32_t n = cast<uint32_t>((shr<uint32_t>(hdr,cast<int64_t>(20ULL)) & cast<uint32_t>(255ULL)));
bool consec = (shr<uint32_t>(hdr,cast<int64_t>(31ULL)) != cast<uint32_t>(0ULL));
uint32_t end = cast<uint32_t>((cast<uint32_t>((off + cast<uint32_t>(8ULL))) + cast<uint32_t>((n * cast<uint32_t>(4ULL)))));
if ((modi<uint32_t>(n,cast<uint32_t>(2ULL)) == cast<uint32_t>(1ULL))) {
end += cast<uint32_t>(4ULL);
}
if ((end > cast<uint32_t>(len(buf)))) {
return {ws,go_fmt_Errorf(std::string("pica: entry at 0x%X (reg 0x%03X, %d extras) runs past the %d-byte buffer",72),off,reg,n,len(buf))};
}
ws = append(ws,n3ds_PICAWrite{off,reg,mask,param,(n > cast<uint32_t>(0ULL))});
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < n);i++){
uint16_t r = reg;
if (consec) {
r += cast<uint16_t>((cast<uint16_t>(i) + cast<uint16_t>(1ULL)));
}
ws = append(ws,n3ds_PICAWrite{off,r,mask,word(cast<uint32_t>((cast<uint32_t>((off + cast<uint32_t>(8ULL))) + cast<uint32_t>((i * cast<uint32_t>(4ULL)))))),true});
}
}off = end;
}
}return {ws,{}};
}
}
// tools/platform/n3ds/pica.go:88:1
std::string n3ds_PICARegGroup(uint16_t reg){
{
{
if ((reg < cast<uint16_t>(64ULL))){
return std::string("misc/irq",8);
}
else if ((reg < cast<uint16_t>(128ULL))){
return std::string("rasterizer",10);
}
else if ((reg < cast<uint16_t>(192ULL))){
return std::string("texturing",9);
}
else if ((reg < cast<uint16_t>(256ULL))){
return std::string("tev",3);
}
else if ((reg < cast<uint16_t>(320ULL))){
return std::string("framebuffer",11);
}
else if ((reg < cast<uint16_t>(512ULL))){
return std::string("fragment-lighting",17);
}
else if ((reg < cast<uint16_t>(552ULL))){
return std::string("geometry-pipeline",17);
}
else if ((reg < cast<uint16_t>(608ULL))){
return std::string("geometry-pipeline",17);
}
else if ((reg < cast<uint16_t>(640ULL))){
return std::string("geometry-shader",15);
}
else if ((reg < cast<uint16_t>(704ULL))){
return std::string("vertex-shader",13);
}
else if ((reg < cast<uint16_t>(736ULL))){
return std::string("vertex-shader-upload",20);
}
else {
return std::string("unknown-2C0",11);
}
}
tmp359:;
}
}
// tools/platform/n3ds/profile.go:109:1
n3ds_profCounters n3ds_Machine_profCounters(n3ds_Machine* m){
{
n3ds_GPU* g = m->gpu;
return n3ds_profCounters{g->Draws,g->PixelsDrawn,g->DepthKilled,g->CulledTris,g->RejectedTris,g->ShadowWrites,g->ListHops};
}
}
// tools/platform/n3ds/profile.go:116:1
time_Time n3ds_Machine_profStart(n3ds_Machine* m){
{
if ((!m->Profile)) {
return time_Time{};
}
return go_time_Now();
}
}
// tools/platform/n3ds/profile.go:123:1
void n3ds_Machine_profEnd(n3ds_Machine* m,int64_t bucket,time_Time t){
{
if (time_Time_IsZero(t)) {
return ;
}
m->prof.ns[bucket] += cast<int64_t>(go_time_Since(t));
m->prof.count[bucket]++;
}
}
// tools/platform/n3ds/profile.go:133:1
void n3ds_Machine_profRunEnter(n3ds_Machine* m){
{
if ((!m->Profile)) {
return ;
}
auto tmp360 = std::make_tuple(go_time_Now(),true);
m->prof.runStart = std::get<0>(tmp360);
m->prof.inRun = std::get<1>(tmp360);
}
}
// tools/platform/n3ds/profile.go:140:1
void n3ds_Machine_profRunExit(n3ds_Machine* m){
{
if (((!m->Profile) || (!m->prof.inRun))) {
return ;
}
m->prof.frameNs += cast<int64_t>(go_time_Since(m->prof.runStart));
m->prof.inRun = false;
}
}
// tools/platform/n3ds/profile.go:151:1
void n3ds_Machine_profFrame(n3ds_Machine* m){
{
if ((!m->Profile)) {
return ;
}
n3ds_profState* p = (&m->prof);
int64_t total = p->frameNs;
if (p->inRun) {
total += cast<int64_t>(go_time_Since(p->runStart));
p->runStart = go_time_Now();
}
p->frameNs = cast<int64_t>(0ULL);
auto ms = [&](int64_t ns)->double{
return cast<double>((cast<double>(ns) / cast<double>(1.00000000000000000e+06)));
}
;
int64_t raster = cast<int64_t>((p->ns[cast<int64_t>(2ULL)] - p->ns[cast<int64_t>(3ULL)]));
if ((raster < cast<int64_t>(0ULL))) {
raster = cast<int64_t>(0ULL);
}
Slice<n3ds_ProfileBucket> buckets = Slice<n3ds_ProfileBucket>::make(cast<int64_t>(0ULL),cast<int64_t>(8ULL));
int64_t summed={};
{int64_t b = cast<int64_t>(0ULL);for (;(b < cast<int64_t>(7ULL));b++){
int64_t ns = p->ns[b];
if ((b == cast<int64_t>(2ULL))) {
ns = raster;
}
summed += ns;
buckets = append(buckets,n3ds_ProfileBucket{n3ds_bucketNames[b],ms(ns),cast<int64_t>(p->count[b])});
}
}int64_t other = cast<int64_t>((total - summed));
if ((other < cast<int64_t>(0ULL))) {
other = cast<int64_t>(0ULL);
}
buckets = append(buckets,n3ds_ProfileBucket{std::string("arm11 + rest (derived)",22),ms(other),cast<int64_t>(0ULL)});
n3ds_profCounters now = n3ds_Machine_profCounters(m);
n3ds_profCounters d = n3ds_profCounters{cast<int64_t>((now.draws - p->base.draws)),cast<int64_t>((now.frags - p->base.frags)),cast<int64_t>((now.depthKilled - p->base.depthKilled)),cast<int64_t>((now.culled - p->base.culled)),cast<int64_t>((now.rejected - p->base.rejected)),cast<int64_t>((now.shadowWrites - p->base.shadowWrites)),cast<int64_t>((now.listHops - p->base.listHops))};
int64_t instrs = cast<int64_t>(cast<uint64_t>((m->instrs - p->baseInstr)));
p->last = n3ds_FrameProfile{ms(total),buckets,Slice<n3ds_ProfileCounter>{n3ds_ProfileCounter{std::string("draws",5),d.draws},n3ds_ProfileCounter{std::string("fragments drawn",15),d.frags},n3ds_ProfileCounter{std::string("depth-killed",12),d.depthKilled},n3ds_ProfileCounter{std::string("tris culled",11),d.culled},n3ds_ProfileCounter{std::string("tris w-rejected",15),d.rejected},n3ds_ProfileCounter{std::string("shadow writes",13),d.shadowWrites},n3ds_ProfileCounter{std::string("list hops",9),d.listHops},n3ds_ProfileCounter{std::string("arm11 instructions",18),instrs},n3ds_ProfileCounter{std::string("idle skips",10),p->idleSkips}},(d.draws > cast<int64_t>(0ULL))};
p->has = true;
auto tmp361 = std::make_tuple(std::array<int64_t,7>{},std::array<int64_t,7>{});
p->ns = std::get<0>(tmp361);
p->count = std::get<1>(tmp361);
p->idleSkips = cast<int64_t>(0ULL);
auto tmp362 = std::make_tuple(now,m->instrs);
p->base = std::get<0>(tmp362);
p->baseInstr = std::get<1>(tmp362);
}
}
// tools/platform/n3ds/profile.go:225:1
n3ds_FrameProfile n3ds_Machine_FrameProfile(n3ds_Machine* m){
{
if ((!m->prof.has)) {
return n3ds_FrameProfile{};
}
return m->prof.last;
}
}
// tools/platform/n3ds/profile.go:234:1
void n3ds_Machine_SetProfile(n3ds_Machine* m,bool on){
{
m->Profile = on;
m->prof = n3ds_profState{};
if (on) {
auto tmp363 = std::make_tuple(n3ds_Machine_profCounters(m),m->instrs);
m->prof.base = std::get<0>(tmp363);
m->prof.baseInstr = std::get<1>(tmp363);
}
}
}
// tools/platform/n3ds/romfs.go:56:1
uint64_t n3ds_ivfcLevel_blockSize(n3ds_ivfcLevel l){
{
return shl<uint64_t>(cast<uint64_t>(1ULL),l.BlockSizeLog);
}
}
// tools/platform/n3ds/romfs.go:84:1
uint64_t n3ds_align64(uint64_t v,uint64_t a){
{
if ((a == cast<uint64_t>(0ULL))) {
return v;
}
return ((cast<uint64_t>((cast<uint64_t>((v + a)) - cast<uint64_t>(1ULL)))) & ~((cast<uint64_t>((a - cast<uint64_t>(1ULL))))));
}
}
// tools/platform/n3ds/romfs.go:92:1
std::tuple<n3ds_RomFS*,Error> n3ds_ParseRomFS(Slice<uint8_t> b){
{
if ((len(b) < cast<int64_t>(96ULL))) {
return {{},go_fmt_Errorf(std::string("romfs: region too short for an IVFC header (%d bytes)",53),len(b))};
}
if ((cast<std::string>(sub(b,cast<int64_t>(0ULL),cast<int64_t>(4ULL))) != n3ds_ivfcMagic)) {
return {{},go_fmt_Errorf(std::string("romfs: bad IVFC magic %q",24),sub(b,cast<int64_t>(0ULL),cast<int64_t>(4ULL)))};
}
{
uint32_t m = le_Uint32(sub(b,cast<int64_t>(4ULL),len(b)));
if ((m != cast<uint32_t>(65536ULL))) {
return {{},go_fmt_Errorf(std::string("romfs: bad IVFC magic number 0x%08x, want 0x%08x",48),m,cast<int64_t>(65536ULL))};
}
}
uint64_t masterHashSize = cast<uint64_t>(le_Uint32(sub(b,cast<int64_t>(8ULL),len(b))));
std::array<n3ds_ivfcLevel,3> lvl={};
{auto&& tmp364 = lvl;
for(int64_t tmp365=0;tmp365<len(tmp364);++tmp365){
auto i=tmp365;int64_t o = cast<int64_t>((cast<int64_t>(12ULL) + cast<int64_t>((i * cast<int64_t>(24ULL)))));
lvl[i] = n3ds_ivfcLevel{le_Uint64(sub(b,o,len(b))),le_Uint64(sub(b,cast<int64_t>((o + cast<int64_t>(8ULL))),len(b))),le_Uint32(sub(b,cast<int64_t>((o + cast<int64_t>(16ULL))),len(b)))};
}}
uint64_t l3Off = n3ds_align64(cast<uint64_t>((cast<uint64_t>(96ULL) + masterHashSize)),n3ds_ivfcLevel_blockSize(lvl[cast<int64_t>(2ULL)]));
uint64_t l1Off = n3ds_align64(cast<uint64_t>((l3Off + lvl[cast<int64_t>(2ULL)].HashDataSize)),n3ds_ivfcLevel_blockSize(lvl[cast<int64_t>(0ULL)]));
uint64_t l2Off = n3ds_align64(cast<uint64_t>((l1Off + lvl[cast<int64_t>(0ULL)].HashDataSize)),n3ds_ivfcLevel_blockSize(lvl[cast<int64_t>(1ULL)]));
uint64_t end = n3ds_align64(cast<uint64_t>((l2Off + lvl[cast<int64_t>(1ULL)].HashDataSize)),n3ds_ivfcLevel_blockSize(lvl[cast<int64_t>(1ULL)]));
if ((end != cast<uint64_t>(len(b)))) {
return {{},go_fmt_Errorf(std::string("romfs: IVFC levels end at 0x%x but the region is 0x%x bytes",59),end,len(b))};
}
uint64_t off = l3Off;
if ((cast<uint64_t>((off + cast<uint64_t>(40ULL))) > cast<uint64_t>(len(b)))) {
return {{},go_fmt_Errorf(std::string("romfs: level-3 offset 0x%x runs past the region end (0x%x)",58),off,len(b))};
}
Slice<uint8_t> l3 = sub(b,off,len(b));
{
uint32_t hl = le_Uint32(l3);
if ((hl != cast<uint32_t>(40ULL))) {
return {{},go_fmt_Errorf(std::string("romfs: level-3 header length 0x%x at offset 0x%x, want 0x%x \342\200\224 IVFC level walk mislanded",89),hl,off,cast<int64_t>(40ULL))};
}
}
auto u32 = [&](int64_t o)->uint32_t{
return le_Uint32(sub(l3,o,len(l3)));
}
;
auto tmp366 = std::make_tuple(u32(cast<int64_t>(12ULL)),u32(cast<int64_t>(16ULL)));
uint32_t dirMetaOff = std::get<0>(tmp366);
uint32_t dirMetaLen = std::get<1>(tmp366);
auto tmp367 = std::make_tuple(u32(cast<int64_t>(28ULL)),u32(cast<int64_t>(32ULL)));
uint32_t fileMetaOff = std::get<0>(tmp367);
uint32_t fileMetaLen = std::get<1>(tmp367);
uint32_t fileDataOff = u32(cast<int64_t>(36ULL));
{auto&& tmp368 = Slice<Anon67>{Anon67{std::string("dir meta",8),dirMetaOff,dirMetaLen},Anon67{std::string("file meta",9),fileMetaOff,fileMetaLen}};
for(int64_t tmp369=0;tmp369<len(tmp368);++tmp369){
auto r=tmp368[tmp369];if ((cast<uint64_t>((cast<uint64_t>((cast<uint64_t>(off) + cast<uint64_t>(r.off))) + cast<uint64_t>(r.len))) > cast<uint64_t>(len(b)))) {
return {{},go_fmt_Errorf(std::string("romfs: %s table [0x%x+0x%x] runs past the region end",52),r.name,r.off,r.len)};
}
}}
n3ds_RomFS* fs = arenaNew(n3ds_RomFS{b,cast<int64_t>((cast<int64_t>(off) + cast<int64_t>(fileDataOff))),sub(b,cast<int64_t>(96ULL),cast<uint64_t>((cast<uint64_t>(96ULL) + masterHashSize))),std::array<n3ds_ivfcSpan,3>{n3ds_ivfcSpan{cast<int64_t>(l1Off),cast<int64_t>(lvl[cast<int64_t>(0ULL)].HashDataSize),cast<int64_t>(n3ds_ivfcLevel_blockSize(lvl[cast<int64_t>(0ULL)]))},n3ds_ivfcSpan{cast<int64_t>(l2Off),cast<int64_t>(lvl[cast<int64_t>(1ULL)].HashDataSize),cast<int64_t>(n3ds_ivfcLevel_blockSize(lvl[cast<int64_t>(1ULL)]))},n3ds_ivfcSpan{cast<int64_t>(l3Off),cast<int64_t>(lvl[cast<int64_t>(2ULL)].HashDataSize),cast<int64_t>(n3ds_ivfcLevel_blockSize(lvl[cast<int64_t>(2ULL)]))}},{},{}});
Slice<uint8_t> dirMeta = sub(l3,dirMetaOff,cast<uint32_t>((dirMetaOff + dirMetaLen)));
Slice<uint8_t> fileMeta = sub(l3,fileMetaOff,cast<uint32_t>((fileMetaOff + fileMetaLen)));
{
Error err = n3ds_RomFS_walk(fs,dirMeta,fileMeta,cast<uint32_t>(0ULL),std::string("",0));
if (bool(err)) {
return {{},err};
}
}
go_sort_Slice(fs->Files,[&](int64_t i,int64_t j)->bool{
return (fs->Files[i].Path < fs->Files[j].Path);
}
);
go_sort_Strings(fs->Dirs);
return {fs,{}};
}
}
// tools/platform/n3ds/romfs.go:176:1
Error n3ds_RomFS_walk(n3ds_RomFS* fs,Slice<uint8_t> dirMeta,Slice<uint8_t> fileMeta,uint32_t dirOff,std::string prefix){
{
if ((cast<int64_t>((cast<int64_t>(dirOff) + cast<int64_t>(24ULL))) > len(dirMeta))) {
return go_fmt_Errorf(std::string("romfs: directory entry at 0x%x runs past the metadata table",59),dirOff);
}
Slice<uint8_t> d = sub(dirMeta,dirOff,len(dirMeta));
auto u32 = [&](int64_t o)->uint32_t{
return le_Uint32(sub(d,o,len(d)));
}
;
auto tmp370 = std::make_tuple(u32(cast<int64_t>(8ULL)),u32(cast<int64_t>(12ULL)));
uint32_t firstChild = std::get<0>(tmp370);
uint32_t firstFile = std::get<1>(tmp370);
{uint32_t fo = firstFile;for (;(fo != cast<uint32_t>(4294967295ULL));){
if ((cast<int64_t>((cast<int64_t>(fo) + cast<int64_t>(32ULL))) > len(fileMeta))) {
return go_fmt_Errorf(std::string("romfs: file entry at 0x%x runs past the metadata table",54),fo);
}
Slice<uint8_t> f = sub(fileMeta,fo,len(fileMeta));
int64_t dataOff = cast<int64_t>(le_Uint64(sub(f,cast<int64_t>(8ULL),len(f))));
int64_t dataLen = cast<int64_t>(le_Uint64(sub(f,cast<int64_t>(16ULL),len(f))));
uint32_t nameLen = le_Uint32(sub(f,cast<int64_t>(28ULL),len(f)));
auto tmp371 = n3ds_utf16Name(f,cast<int64_t>(32ULL),nameLen,cast<int64_t>((len(fileMeta) - cast<int64_t>(fo))));
std::string name = std::get<0>(tmp371);
Error err = std::get<1>(tmp371);
if (bool(err)) {
return go_fmt_Errorf(std::string("romfs: file entry at 0x%x: %w",29),fo,err);
}
if ((cast<int64_t>((cast<int64_t>((fs->dataStart + dataOff)) + dataLen)) > cast<int64_t>(len(fs->raw)))) {
return go_fmt_Errorf(std::string("romfs: file %q data [0x%x+0x%x] runs past the region end",56),((prefix + std::string("/",1)) + name),dataOff,dataLen);
}
fs->Files = append(fs->Files,n3ds_RomFSFile{((prefix + std::string("/",1)) + name),dataOff,dataLen});
fo = le_Uint32(sub(f,cast<int64_t>(4ULL),len(f)));
}
}{uint32_t co = firstChild;for (;(co != cast<uint32_t>(4294967295ULL));){
if ((cast<int64_t>((cast<int64_t>(co) + cast<int64_t>(24ULL))) > len(dirMeta))) {
return go_fmt_Errorf(std::string("romfs: directory entry at 0x%x runs past the metadata table",59),co);
}
Slice<uint8_t> c = sub(dirMeta,co,len(dirMeta));
uint32_t nameLen = le_Uint32(sub(c,cast<int64_t>(20ULL),len(c)));
auto tmp372 = n3ds_utf16Name(c,cast<int64_t>(24ULL),nameLen,cast<int64_t>((len(dirMeta) - cast<int64_t>(co))));
std::string name = std::get<0>(tmp372);
Error err = std::get<1>(tmp372);
if (bool(err)) {
return go_fmt_Errorf(std::string("romfs: directory entry at 0x%x: %w",34),co,err);
}
std::string path = ((prefix + std::string("/",1)) + name);
fs->Dirs = append(fs->Dirs,path);
{
Error err = n3ds_RomFS_walk(fs,dirMeta,fileMeta,co,path);
if (bool(err)) {
return err;
}
}
co = le_Uint32(sub(c,cast<int64_t>(4ULL),len(c)));
}
}return {};
}
}
// tools/platform/n3ds/romfs.go:226:1
std::tuple<std::string,Error> n3ds_utf16Name(Slice<uint8_t> e,int64_t off,uint32_t nameLen,int64_t avail){
{
if ((modi<uint32_t>(nameLen,cast<uint32_t>(2ULL)) != cast<uint32_t>(0ULL))) {
return {std::string("",0),go_fmt_Errorf(std::string("odd name length %d",18),nameLen)};
}
if ((cast<int64_t>((off + cast<int64_t>(nameLen))) > avail)) {
return {std::string("",0),go_fmt_Errorf(std::string("name of %d bytes runs past the entry",36),nameLen)};
}
Slice<uint16_t> u = Slice<uint16_t>::make(divi<uint32_t>(nameLen,cast<uint32_t>(2ULL)));
{auto&& tmp373 = u;
for(int64_t tmp374=0;tmp374<len(tmp373);++tmp374){
auto i=tmp374;u[i] = le_Uint16(sub(e,cast<int64_t>((off + cast<int64_t>((i * cast<int64_t>(2ULL))))),len(e)));
}}
std::string name = cast<std::string>(go_utf16_Decode(u));
if (((name == std::string("",0)) || go_strings_ContainsAny(name,std::string("/\000",2)))) {
return {std::string("",0),go_fmt_Errorf(std::string("invalid name %q",15),name)};
}
return {name,{}};
}
}
// tools/platform/n3ds/romfs.go:281:1
int64_t n3ds_min64(int64_t a,int64_t b){
{
if ((a < b)) {
return a;
}
return b;
}
}
// tools/platform/n3ds/romfs.go:290:1
std::tuple<Slice<uint8_t>,Error> n3ds_RomFS_File(n3ds_RomFS* fs,std::string path){
{
int64_t i = go_sort_Search(len(fs->Files),[&](int64_t i)->bool{
return (fs->Files[i].Path >= path);
}
);
if (((i == len(fs->Files)) || (fs->Files[i].Path != path))) {
return {{},go_fmt_Errorf(std::string("romfs: no file %q",17),path)};
}
return {n3ds_RomFS_Data(fs,fs->Files[i]),{}};
}
}
// tools/platform/n3ds/romfs.go:299:1
Slice<uint8_t> n3ds_RomFS_Data(n3ds_RomFS* fs,n3ds_RomFSFile f){
{
int64_t start = cast<int64_t>((fs->dataStart + f.Offset));
return sub(fs->raw,start,cast<int64_t>((start + f.Size)));
}
}
// tools/platform/n3ds/romfs.go:312:1
int64_t n3ds_RomFS_L3Offset(n3ds_RomFS* fs,n3ds_RomFSFile f){
{
return cast<int64_t>((cast<int64_t>((fs->dataStart - fs->Levels[cast<int64_t>(2ULL)].Offset)) + f.Offset));
}
}
// tools/platform/n3ds/romfs.go:321:1
std::tuple<n3ds_RomFSFile,int64_t,bool> n3ds_RomFS_FileAt(n3ds_RomFS* fs,int64_t l3Off){
n3ds_RomFSFile f{};
int64_t within{};
bool ok{};
{
{auto&& tmp375 = fs->Files;
for(int64_t tmp376=0;tmp376<len(tmp375);++tmp376){
auto cand=tmp375[tmp376];{
int64_t s = n3ds_RomFS_L3Offset(fs,cand);
if (((l3Off >= s) && (l3Off < cast<int64_t>((s + cand.Size))))) {
return {cand,cast<int64_t>((l3Off - s)),true};
}
}
}}
return {n3ds_RomFSFile{},cast<int64_t>(0ULL),false};
}
}
// tools/platform/n3ds/run.go:24:1
int64_t n3ds_Machine_Run(n3ds_Machine* m,int64_t budget){
{
int64_t n = cast<int64_t>(0ULL);
int64_t idleFrames = cast<int64_t>(0ULL);
n3ds_Machine_profRunEnter(m);
auto tmp377=defer([&](){n3ds_Machine_profRunExit(m);});
auto tmp378 = std::make_tuple(false,false);
m->stopped = std::get<0>(tmp378);
m->StopRequested = std::get<1>(tmp378);
{;for (;(n < budget);){
if ((m->CPU->Halted || m->StopRequested)) {
break;
}
if (n3ds_Machine_vblankDue(m)) {
n3ds_Machine_deliverVBlank(m);
}
if (n3ds_Machine_dspDue(m)) {
n3ds_Machine_dspTick(m);
}
n3ds_Machine_wakeDueSleepers(m);
n3ds_Machine_processGXQueue(m);
n3ds_thread* t = n3ds_Machine_pickRunnable(m);
if ((!t)) {
constexpr uint64_t never=18446744073709551615ULL;
auto tmp379 = std::make_tuple(cast<uint64_t>(18446744073709551615ULL),std::string("",0));
uint64_t next = std::get<0>(tmp379);
std::string kind = std::get<1>(tmp379);
{
auto tmp380 = n3ds_Machine_gxDeadline(m);
uint64_t dl = std::get<0>(tmp380);
bool ok = std::get<1>(tmp380);
if ((ok && (dl < next))) {
auto tmp381 = std::make_tuple(dl,std::string("gx",2));
next = std::get<0>(tmp381);
kind = std::get<1>(tmp381);
}
}
{
auto tmp382 = n3ds_Machine_dspDeadline(m);
uint64_t dl = std::get<0>(tmp382);
bool ok = std::get<1>(tmp382);
if ((ok && (dl < next))) {
auto tmp383 = std::make_tuple(dl,std::string("dsp",3));
next = std::get<0>(tmp383);
kind = std::get<1>(tmp383);
}
}
{
auto tmp384 = n3ds_Machine_soonestSleeper(m);
uint64_t wt = std::get<0>(tmp384);
bool ok = std::get<1>(tmp384);
if (ok) {
uint64_t dl = m->instrs;
if ((wt > m->tick)) {
dl += divi<uint64_t>((cast<uint64_t>((cast<uint64_t>((wt - m->tick)) + cast<uint64_t>(1ULL)))),cast<uint64_t>(2ULL));
}
if ((dl < next)) {
auto tmp385 = std::make_tuple(dl,std::string("sleep",5));
next = std::get<0>(tmp385);
kind = std::get<1>(tmp385);
}
}
}
if (((m->gspEvent != cast<uint32_t>(0ULL)) && (m->nextFrameInstr < next))) {
auto tmp386 = std::make_tuple(m->nextFrameInstr,std::string("vblank",6));
next = std::get<0>(tmp386);
kind = std::get<1>(tmp386);
}
if (((kind == std::string("",0)) || (idleFrames >= cast<int64_t>(40ULL)))) {
n3ds_Machine_dumpThreads(m);
arm_CPU_Halt(m->CPU,std::string("all threads blocked (deadlock): %d live, none runnable, after %d instructions",77),n3ds_Machine_aliveThreads(m),m->CPU->Instrs);
break;
}
if ((next > m->instrs)) {
m->tick += cast<uint64_t>((cast<uint64_t>(2ULL) * (cast<uint64_t>((next - m->instrs)))));
m->instrs = next;
m->prof.idleSkips++;
}
{
auto tmp388=kind;
if (tmp388==(std::string("gx",2))){
n3ds_Machine_pumpGX(m);
idleFrames = cast<int64_t>(0ULL);
}
else if (tmp388==(std::string("dsp",3))){
n3ds_Machine_dspTick(m);
idleFrames++;
}
else if (tmp388==(std::string("sleep",5))){
idleFrames = cast<int64_t>(0ULL);
}
else if (tmp388==(std::string("vblank",6))){
n3ds_Machine_deliverVBlank(m);
idleFrames++;
}
}
tmp387:;
n3ds_Machine_wakeDueSleepers(m);
continue;
}
idleFrames = cast<int64_t>(0ULL);
n3ds_Machine_switchTo(m,t);
{int64_t q = cast<int64_t>(0ULL);for (;((q < cast<int64_t>(2000ULL)) && (n < budget));q++){
if (m->CPU->Halted) {
break;
}
uint32_t pc = m->CPU->R[cast<int64_t>(15ULL)];
if ((pc == cast<uint32_t>(4294901760ULL))) {
n3ds_Machine_svcExitThread(m,m->CPU);
break;
}
if (get(m->bps,pc)) {
go_fmt_Printf(std::string("breakpoint [t%d] at 0x%08X r0=%08X r1=%08X r4=%08X r5=%08X lr=%08X after %d\012",76),m->curThread->id,pc,m->CPU->R[cast<int64_t>(0ULL)],m->CPU->R[cast<int64_t>(1ULL)],m->CPU->R[cast<int64_t>(4ULL)],m->CPU->R[cast<int64_t>(5ULL)],m->CPU->R[cast<int64_t>(14ULL)],n);
m->stopped = true;
break;
}
if (get(m->tracefroms,pc)) {
m->Trace = true;
m->traceN = cast<int64_t>(0ULL);
}
if (get(m->logpcs,pc)) {
uint32_t sp = m->CPU->R[cast<int64_t>(13ULL)];
go_fmt_Printf(std::string("logpc [t%d] 0x%08X r0=%08X r1=%08X r2=%08X r3=%08X r4=%08X r5=%08X r6=%08X r7=%08X lr=%08X sp=[%08X %08X %08X] instr=%d%s\012",122),m->curThread->id,pc,m->CPU->R[cast<int64_t>(0ULL)],m->CPU->R[cast<int64_t>(1ULL)],m->CPU->R[cast<int64_t>(2ULL)],m->CPU->R[cast<int64_t>(3ULL)],m->CPU->R[cast<int64_t>(4ULL)],m->CPU->R[cast<int64_t>(5ULL)],m->CPU->R[cast<int64_t>(6ULL)],m->CPU->R[cast<int64_t>(7ULL)],m->CPU->R[cast<int64_t>(14ULL)],n3ds_Machine_ReadWord(m,sp),n3ds_Machine_ReadWord(m,cast<uint32_t>((sp + cast<uint32_t>(4ULL)))),n3ds_Machine_ReadWord(m,cast<uint32_t>((sp + cast<uint32_t>(8ULL)))),m->CPU->Instrs,n3ds_Machine_argStrings(m));
}
if ((m->Trace && (m->traceN < m->traceMax))) {
n3ds_Machine_traceOne(m,pc);
m->traceN++;
}
n3ds_Machine_checkWatches(m,pc);
m->instrs++;
m->tick += cast<uint64_t>(2ULL);
m->reschedule = false;
arm_CPU_Step(m->CPU);
n++;
if (((m->reschedule || (t->state != cast<n3ds_threadState>(1ULL))) || m->StopRequested)) {
break;
}
}
}t->ctx = (*m->CPU);
if ((t->state == cast<n3ds_threadState>(1ULL))) {
t->state = cast<n3ds_threadState>(0ULL);
}
if ((m->stopped || m->StopRequested)) {
break;
}
}
}return n;
}
}
// tools/platform/n3ds/run.go:172:1
int64_t n3ds_Machine_RunFrames(n3ds_Machine* m,int64_t frames,int64_t budget){
{
if ((frames <= cast<int64_t>(0ULL))) {
return cast<int64_t>(0ULL);
}
uint64_t target = cast<uint64_t>((m->vblankCount + cast<uint64_t>(frames)));
std::function<void(n3ds_Machine*)> prev = m->OnFrame;
m->OnFrame = [&](n3ds_Machine* mm)->void{
if (bool(prev)) {
prev(mm);
}
if ((mm->vblankCount >= target)) {
mm->StopRequested = true;
}
}
;
auto tmp389=defer([&](){[&]()->void{
m->OnFrame = prev;
}
();});
int64_t n = cast<int64_t>(0ULL);
{;for (;((((n < budget) && (m->vblankCount < target)) && (!m->CPU->Halted)) && (!m->stopped));){
int64_t ran = n3ds_Machine_Run(m,cast<int64_t>((budget - n)));
n += ran;
if ((ran == cast<int64_t>(0ULL))) {
break;
}
}
}return n;
}
}
// tools/platform/n3ds/run.go:200:1
void n3ds_Machine_SetTrace(n3ds_Machine* m,bool on,int64_t max){
{
m->Trace = on;
m->traceMax = max;
}
}
// tools/platform/n3ds/run.go:206:1
void n3ds_Machine_AddBreakpoint(n3ds_Machine* m,uint32_t addr){
{
m->bps[addr] = true;
}
}
// tools/platform/n3ds/run.go:211:1
void n3ds_Machine_AddTraceFrom(n3ds_Machine* m,uint32_t addr){
{
if ((!m->tracefroms)) {
m->tracefroms = Map<uint32_t,bool>{};
}
m->tracefroms[addr] = true;
}
}
// tools/platform/n3ds/run.go:221:1
void n3ds_Machine_AddLogPC(n3ds_Machine* m,uint32_t addr){
{
if ((!m->logpcs)) {
m->logpcs = Map<uint32_t,bool>{};
}
m->logpcs[addr] = true;
}
}
// tools/platform/n3ds/run.go:229:1
void n3ds_Machine_AddWatch(n3ds_Machine* m,uint32_t addr,uint32_t length){
{
if ((length == cast<uint32_t>(0ULL))) {
length = cast<uint32_t>(1ULL);
}
m->watches = append(m->watches,n3ds_watch{addr,length,{},{}});
}
}
// tools/platform/n3ds/run.go:236:1
void n3ds_Machine_traceOne(n3ds_Machine* m,uint32_t pc){
{
std::array<uint8_t,4> buf={};
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < cast<uint32_t>(4ULL));i++){
buf[i] = n3ds_Machine_Read(m,cast<uint32_t>((pc + i)));
}
}arm_Inst in = arm_DecodeVariant(sub(buf,0,len(buf)),pc,m->CPU->Thumb,cast<arm_Variant>(1ULL));
go_fmt_Printf(std::string("[t%d] %08X: %-22s  r0=%08X r1=%08X r2=%08X r3=%08X sp=%08X lr=%08X\012",67),m->curThread->id,pc,in.Text,m->CPU->R[cast<int64_t>(0ULL)],m->CPU->R[cast<int64_t>(1ULL)],m->CPU->R[cast<int64_t>(2ULL)],m->CPU->R[cast<int64_t>(3ULL)],m->CPU->R[cast<int64_t>(13ULL)],m->CPU->R[cast<int64_t>(14ULL)]);
}
}
// tools/platform/n3ds/run.go:249:1
void n3ds_Machine_checkWatches(n3ds_Machine* m,uint32_t pc){
{
{auto&& tmp390 = m->watches;
for(int64_t tmp391=0;tmp391<len(tmp390);++tmp391){
auto i=tmp391;n3ds_watch* w = (&m->watches[i]);
{uint32_t off = cast<uint32_t>(0ULL);for (;(off < w->len);off += cast<uint32_t>(4ULL)){
uint32_t a = cast<uint32_t>((w->addr + off));
uint32_t v = n3ds_Machine_ReadWord(m,a);
if ((!w->last)) {
w->last = Map<uint32_t,uint32_t>{};
w->seen = Map<uint32_t,bool>{};
}
if ((!get(w->seen,a))) {
w->seen[a] = true;
w->last[a] = v;
continue;
}
if ((v != get(w->last,a))) {
go_fmt_Printf(std::string("watch [t%d] 0x%08X: 0x%08X -> 0x%08X at pc=0x%08X\012",50),m->curThread->id,a,get(w->last,a),v,pc);
w->last[a] = v;
}
}
}}}
}
}
// tools/platform/n3ds/run.go:273:1
std::string n3ds_Machine_HaltReason(n3ds_Machine* m){
{
return m->CPU->HaltReason;
}
}
// tools/platform/n3ds/run.go:276:1
std::string n3ds_Machine_DebugString(n3ds_Machine* m){
{
return cast<std::string>(m->debugOut);
}
}
// tools/platform/n3ds/run.go:279:1
uint32_t n3ds_Machine_Entry(n3ds_Machine* m){
{
return m->entry;
}
}
// tools/platform/n3ds/run.go:282:1
uint64_t n3ds_Machine_Instrs(n3ds_Machine* m){
{
return m->CPU->Instrs;
}
}
// tools/platform/n3ds/run.go:287:1
std::string n3ds_Machine_argStrings(n3ds_Machine* m){
{
std::string out = std::string("",0);
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(4ULL));i++){
uint32_t p = m->CPU->R[i];
if ((p < cast<uint32_t>(1048576ULL))) {
continue;
}
Slice<uint8_t> b={};
{uint32_t n = cast<uint32_t>(0ULL);for (;(n < cast<uint32_t>(64ULL));n++){
uint8_t c = n3ds_Machine_Read(m,cast<uint32_t>((p + n)));
if ((c == cast<uint8_t>(0ULL))) {
break;
}
if (((c < cast<uint8_t>(32ULL)) || (c > cast<uint8_t>(126ULL)))) {
b = {};
break;
}
b = append(b,c);
}
}if ((len(b) >= cast<int64_t>(3ULL))) {
out += go_fmt_Sprintf(std::string(" r%d=%q",7),i,cast<std::string>(b));
}
}
}return out;
}
}
// tools/platform/n3ds/svc.go:81:1
bool n3ds_Machine_handleSVC(n3ds_Machine* m,arm_CPU* c,uint32_t comment){
{
auto tmp392=defer([&](){n3ds_Machine_profEnd(m,cast<int64_t>(6ULL),n3ds_Machine_profStart(m));});
uint32_t num = cast<uint32_t>((comment & cast<uint32_t>(255ULL)));
n3ds_svcEvent ev = n3ds_svcEvent{arm_CPU_PC(c),num,n3ds_svcName(num),{}};
ev.Args = std::array<uint32_t,4>{c->R[cast<int64_t>(0ULL)],c->R[cast<int64_t>(1ULL)],c->R[cast<int64_t>(2ULL)],c->R[cast<int64_t>(3ULL)]};
m->svcLog = append(m->svcLog,ev);
if (m->Verbose) {
go_fmt_Printf(std::string("[t%d] svc 0x%02X %-20s r0=%08X r1=%08X r2=%08X r3=%08X  pc=%08X\012",64),m->curThread->id,num,ev.Name,c->R[cast<int64_t>(0ULL)],c->R[cast<int64_t>(1ULL)],c->R[cast<int64_t>(2ULL)],c->R[cast<int64_t>(3ULL)],arm_CPU_PC(c));
}
{
switch(num){
case cast<uint32_t>(1ULL):{
n3ds_Machine_svcControlMemory(m,c);
break;}
case cast<uint32_t>(2ULL):{
n3ds_Machine_svcQueryMemory(m,c);
break;}
case cast<uint32_t>(3ULL):{
arm_CPU_Halt(c,std::string("svcExitProcess after %d instructions",36),c->Instrs);
break;}
case cast<uint32_t>(40ULL):{
c->R[cast<int64_t>(0ULL)] = cast<uint32_t>(m->tick);
c->R[cast<int64_t>(1ULL)] = cast<uint32_t>(shr<uint64_t>(m->tick,cast<int64_t>(32ULL)));
break;}
case cast<uint32_t>(42ULL):{
auto tmp393 = std::make_tuple(cast<uint32_t>(0ULL),cast<uint32_t>(0ULL),cast<uint32_t>(0ULL));
c->R[cast<int64_t>(0ULL)] = std::get<0>(tmp393);
c->R[cast<int64_t>(1ULL)] = std::get<1>(tmp393);
c->R[cast<int64_t>(2ULL)] = std::get<2>(tmp393);
break;}
case cast<uint32_t>(43ULL):case cast<uint32_t>(44ULL):case cast<uint32_t>(41ULL):{
auto tmp394 = std::make_tuple(cast<uint32_t>(0ULL),cast<uint32_t>(0ULL),cast<uint32_t>(0ULL));
c->R[cast<int64_t>(0ULL)] = std::get<0>(tmp394);
c->R[cast<int64_t>(1ULL)] = std::get<1>(tmp394);
c->R[cast<int64_t>(2ULL)] = std::get<2>(tmp394);
break;}
case cast<uint32_t>(11ULL):{
auto tmp395 = std::make_tuple(cast<uint32_t>(0ULL),cast<uint32_t>(m->curThread->priority));
c->R[cast<int64_t>(0ULL)] = std::get<0>(tmp395);
c->R[cast<int64_t>(1ULL)] = std::get<1>(tmp395);
break;}
case cast<uint32_t>(12ULL):{
m->curThread->priority = cast<int32_t>(c->R[cast<int64_t>(1ULL)]);
c->R[cast<int64_t>(0ULL)] = cast<uint32_t>(0ULL);
break;}
case cast<uint32_t>(10ULL):{
n3ds_Machine_svcSleepThread(m,c);
break;}
case cast<uint32_t>(9ULL):{
n3ds_Machine_svcExitThread(m,c);
break;}
case cast<uint32_t>(24ULL):{
n3ds_Machine_svcSignalEvent(m,c);
break;}
case cast<uint32_t>(25ULL):{
n3ds_Machine_svcClearEvent(m,c);
break;}
case cast<uint32_t>(20ULL):{
n3ds_Machine_svcReleaseMutex(m,c);
break;}
case cast<uint32_t>(22ULL):{
n3ds_Machine_svcReleaseSemaphore(m,c);
break;}
case cast<uint32_t>(34ULL):{
n3ds_Machine_svcArbitrateAddress(m,c);
break;}
case cast<uint32_t>(53ULL):{
auto tmp396 = std::make_tuple(cast<uint32_t>(0ULL),cast<uint32_t>(1ULL));
c->R[cast<int64_t>(0ULL)] = std::get<0>(tmp396);
c->R[cast<int64_t>(1ULL)] = std::get<1>(tmp396);
break;}
case cast<uint32_t>(55ULL):{
auto tmp397 = std::make_tuple(cast<uint32_t>(0ULL),m->curThread->id);
c->R[cast<int64_t>(0ULL)] = std::get<0>(tmp397);
c->R[cast<int64_t>(1ULL)] = std::get<1>(tmp397);
break;}
case cast<uint32_t>(8ULL):{
uint32_t h = n3ds_Machine_createThread(m,cast<int32_t>(c->R[cast<int64_t>(0ULL)]),c->R[cast<int64_t>(1ULL)],c->R[cast<int64_t>(2ULL)],c->R[cast<int64_t>(3ULL)]);
auto tmp398 = std::make_tuple(cast<uint32_t>(0ULL),h);
c->R[cast<int64_t>(0ULL)] = std::get<0>(tmp398);
c->R[cast<int64_t>(1ULL)] = std::get<1>(tmp398);
break;}
case cast<uint32_t>(23ULL):{
uint32_t h = n3ds_Machine_newHandle(m,std::string("event",5),false);
m->handles[h]->manualReset = (c->R[cast<int64_t>(1ULL)] != cast<uint32_t>(0ULL));
auto tmp399 = std::make_tuple(cast<uint32_t>(0ULL),h);
c->R[cast<int64_t>(0ULL)] = std::get<0>(tmp399);
c->R[cast<int64_t>(1ULL)] = std::get<1>(tmp399);
break;}
case cast<uint32_t>(19ULL):{
uint32_t h = n3ds_Machine_newHandle(m,std::string("mutex",5),false);
if ((c->R[cast<int64_t>(1ULL)] != cast<uint32_t>(0ULL))) {
auto tmp400 = std::make_tuple(m->curThread->id,cast<int64_t>(1ULL));
m->handles[h]->mutexOwner = std::get<0>(tmp400);
m->handles[h]->mutexDepth = std::get<1>(tmp400);
}
auto tmp401 = std::make_tuple(cast<uint32_t>(0ULL),h);
c->R[cast<int64_t>(0ULL)] = std::get<0>(tmp401);
c->R[cast<int64_t>(1ULL)] = std::get<1>(tmp401);
break;}
case cast<uint32_t>(21ULL):{
uint32_t h = n3ds_Machine_newHandle(m,std::string("semaphore",9),false);
m->handles[h]->semCount = cast<int32_t>(c->R[cast<int64_t>(1ULL)]);
auto tmp402 = std::make_tuple(cast<uint32_t>(0ULL),h);
c->R[cast<int64_t>(0ULL)] = std::get<0>(tmp402);
c->R[cast<int64_t>(1ULL)] = std::get<1>(tmp402);
break;}
case cast<uint32_t>(26ULL):{
n3ds_Machine_svcCreateHandle(m,c,std::string("timer",5),false,cast<int64_t>(1ULL));
break;}
case cast<uint32_t>(33ULL):{
n3ds_Machine_svcCreateHandle(m,c,std::string("arbiter",7),false,cast<int64_t>(1ULL));
break;}
case cast<uint32_t>(30ULL):{
n3ds_Machine_svcCreateHandle(m,c,std::string("memblock",8),false,cast<int64_t>(0ULL));
break;}
case cast<uint32_t>(31ULL):{
n3ds_Machine_svcMapMemoryBlock(m,c);
break;}
case cast<uint32_t>(32ULL):{
n3ds_Machine_svcUnmapMemoryBlock(m,c);
break;}
case cast<uint32_t>(39ULL):{
n3ds_Machine_svcCreateHandle(m,c,std::string("dup",3),false,cast<int64_t>(1ULL));
break;}
case cast<uint32_t>(45ULL):{
n3ds_Machine_svcConnectToPort(m,c);
break;}
case cast<uint32_t>(36ULL):{
n3ds_Machine_svcWaitSync1(m,c);
break;}
case cast<uint32_t>(37ULL):{
n3ds_Machine_svcWaitSyncN(m,c);
break;}
case cast<uint32_t>(35ULL):{
removeKey(m->handles,c->R[cast<int64_t>(0ULL)]);
removeKey(m->ports,c->R[cast<int64_t>(0ULL)]);
c->R[cast<int64_t>(0ULL)] = cast<uint32_t>(0ULL);
break;}
case cast<uint32_t>(50ULL):{
n3ds_Machine_svcSendSyncRequest(m,c);
break;}
case cast<uint32_t>(56ULL):{
n3ds_Machine_svcCreateHandle(m,c,std::string("resourcelimit",13),false,cast<int64_t>(1ULL));
break;}
case cast<uint32_t>(57ULL):{
n3ds_Machine_svcGetResourceLimitValues(m,c,true);
break;}
case cast<uint32_t>(58ULL):{
n3ds_Machine_svcGetResourceLimitValues(m,c,false);
break;}
case cast<uint32_t>(61ULL):{
n3ds_Machine_svcOutputDebugString(m,c);
break;}
case cast<uint32_t>(60ULL):{
arm_CPU_Halt(c,std::string("svcBreak (reason %d) at 0x%08X after %d instructions",52),c->R[cast<int64_t>(0ULL)],arm_CPU_PC(c),c->Instrs);
break;}
default:{
arm_CPU_Halt(c,std::string("unimplemented svc 0x%02X (%s) at 0x%08X after %d instructions",61),num,n3ds_svcName(num),arm_CPU_PC(c),c->Instrs);
break;}
}}
return true;
}
}
// tools/platform/n3ds/svc.go:198:1
void n3ds_Machine_svcControlMemory(n3ds_Machine* m,arm_CPU* c){
{
uint32_t memop = c->R[cast<int64_t>(0ULL)];
uint32_t size = ((cast<uint32_t>((cast<uint32_t>((c->R[cast<int64_t>(3ULL)] + cast<uint32_t>(4096ULL))) - cast<uint32_t>(1ULL)))) & ~(cast<uint32_t>(4095ULL)));
{
switch(cast<uint32_t>((memop & cast<uint32_t>(255ULL)))){
case cast<uint32_t>(3ULL):{
break;}
case cast<uint32_t>(1ULL):case cast<uint32_t>(4ULL):case cast<uint32_t>(6ULL):{
c->R[cast<int64_t>(0ULL)] = cast<uint32_t>(0ULL);
return ;
break;}
default:{
arm_CPU_Halt(c,std::string("svcControlMemory op 0x%X unimplemented at 0x%08X",48),cast<uint32_t>((memop & cast<uint32_t>(255ULL))),arm_CPU_PC(c));
return ;
break;}
}}
if ((size > cast<uint32_t>(100663296ULL))) {
arm_CPU_Halt(c,std::string("svcControlMemory: implausible size 0x%X at 0x%08X (heap-size derivation likely wrong)",85),c->R[cast<int64_t>(3ULL)],arm_CPU_PC(c));
return ;
}
bool linear = (cast<uint32_t>((memop & cast<uint32_t>(65536ULL))) != cast<uint32_t>(0ULL));
uint32_t addr={};
if (linear) {
addr = m->linearPtr;
m->linearPtr += size;
m->linearReg->data = append(m->linearReg->data,Slice<uint8_t>::make(size));
n3ds_Machine_indexRegion(m,m->linearReg);
if ((m->linearPtr > cast<uint32_t>(469762048ULL))) {
arm_CPU_Halt(c,std::string("linear heap exhausted at 0x%08X",31),arm_CPU_PC(c));
return ;
}
}
else {
addr = m->heapPtr;
m->heapPtr += size;
m->heapReg->data = append(m->heapReg->data,Slice<uint8_t>::make(size));
n3ds_Machine_indexRegion(m,m->heapReg);
if ((m->heapPtr > cast<uint32_t>(234881024ULL))) {
arm_CPU_Halt(c,std::string("process heap exhausted at 0x%08X",32),arm_CPU_PC(c));
return ;
}
}
if (m->MemTrace) {
go_fmt_Printf(std::string("[mem] ControlMemory op=%06X addr0=%08X addr1=%08X size=%08X %s -> %08X (linear now %08X)\012",89),memop,c->R[cast<int64_t>(1ULL)],c->R[cast<int64_t>(2ULL)],size,get(Map<bool,std::string>{{true,std::string("LINEAR",6)},{false,std::string("heap",4)}},linear),addr,m->linearPtr);
}
c->R[cast<int64_t>(0ULL)] = cast<uint32_t>(0ULL);
c->R[cast<int64_t>(1ULL)] = addr;
}
}
// tools/platform/n3ds/svc.go:253:1
void n3ds_Machine_svcQueryMemory(n3ds_Machine* m,arm_CPU* c){
{
uint32_t addr = c->R[cast<int64_t>(2ULL)];
n3ds_memRegion* r = n3ds_Machine_regionOf(m,addr);
if ((!r)) {
auto tmp403 = std::make_tuple(cast<uint32_t>(0ULL),addr,cast<uint32_t>(4096ULL),cast<uint32_t>(0ULL),cast<uint32_t>(0ULL),cast<uint32_t>(0ULL));
c->R[cast<int64_t>(0ULL)] = std::get<0>(tmp403);
c->R[cast<int64_t>(1ULL)] = std::get<1>(tmp403);
c->R[cast<int64_t>(2ULL)] = std::get<2>(tmp403);
c->R[cast<int64_t>(3ULL)] = std::get<3>(tmp403);
c->R[cast<int64_t>(4ULL)] = std::get<4>(tmp403);
c->R[cast<int64_t>(5ULL)] = std::get<5>(tmp403);
return ;
}
if (m->MemTrace) {
go_fmt_Printf(std::string("[mem] QueryMemory(%08X) -> region %q base=%08X size=%08X\012",57),addr,r->name,r->base,len(r->data));
}
c->R[cast<int64_t>(0ULL)] = cast<uint32_t>(0ULL);
c->R[cast<int64_t>(1ULL)] = r->base;
c->R[cast<int64_t>(2ULL)] = cast<uint32_t>(len(r->data));
c->R[cast<int64_t>(3ULL)] = cast<uint32_t>(3ULL);
c->R[cast<int64_t>(4ULL)] = cast<uint32_t>(3ULL);
c->R[cast<int64_t>(5ULL)] = cast<uint32_t>(0ULL);
}
}
// tools/platform/n3ds/svc.go:277:1
void n3ds_Machine_svcCreateHandle(n3ds_Machine* m,arm_CPU* c,std::string kind,bool signalled,int64_t handleReg){
{
uint32_t h = m->nextHandle;
m->nextHandle++;
m->handles[h] = arenaNew(n3ds_kobject{kind,{},signalled,{},{},{},{},{},{},{},{}});
c->R[cast<int64_t>(0ULL)] = cast<uint32_t>(0ULL);
c->R[cast<uint32_t>(handleReg)] = h;
}
}
// tools/platform/n3ds/svc.go:289:1
void n3ds_Machine_svcConnectToPort(n3ds_Machine* m,arm_CPU* c){
{
uint32_t namePtr = c->R[cast<int64_t>(1ULL)];
std::string name = n3ds_Machine_readCString(m,namePtr,cast<uint32_t>(12ULL));
uint32_t h = m->nextHandle;
m->nextHandle++;
m->handles[h] = arenaNew(n3ds_kobject{std::string("port",4),name,{},{},{},{},{},{},{},{},{}});
m->ports[h] = name;
c->R[cast<int64_t>(0ULL)] = cast<uint32_t>(0ULL);
c->R[cast<int64_t>(1ULL)] = h;
if (m->Verbose) {
go_fmt_Printf(std::string("  ConnectToPort %q -> handle 0x%08X\012",36),name,h);
}
}
}
// tools/platform/n3ds/svc.go:308:1
void n3ds_Machine_svcSendSyncRequest(n3ds_Machine* m,arm_CPU* c){
{
n3ds_Machine_handleIPC(m,c->R[cast<int64_t>(0ULL)]);
c->R[cast<int64_t>(0ULL)] = cast<uint32_t>(0ULL);
}
}
// tools/platform/n3ds/svc.go:330:1
void n3ds_Machine_svcGetResourceLimitValues(n3ds_Machine* m,arm_CPU* c,bool current){
{
uint32_t values = c->R[cast<int64_t>(0ULL)];
uint32_t names = c->R[cast<int64_t>(2ULL)];
int32_t count = cast<int32_t>(c->R[cast<int64_t>(3ULL)]);
{int32_t i = cast<int32_t>(0ULL);for (;((i < count) && (i < cast<int32_t>(32ULL)));i++){
uint32_t name = n3ds_Machine_ReadWord(m,cast<uint32_t>((names + cast<uint32_t>((cast<uint32_t>(i) * cast<uint32_t>(4ULL))))));
int64_t v={};
{
switch(name){
case cast<uint32_t>(1ULL):{
v = cast<int64_t>(4194304ULL);
break;}
case cast<uint32_t>(0ULL):{
v = cast<int64_t>(24ULL);
break;}
case cast<uint32_t>(9ULL):{
v = cast<int64_t>(0ULL);
break;}
default:{
if (current) {
v = cast<int64_t>(0ULL);
}
else {
v = cast<int64_t>(256ULL);
}
break;}
}}
uint32_t off = cast<uint32_t>((values + cast<uint32_t>((cast<uint32_t>(i) * cast<uint32_t>(8ULL)))));
n3ds_Machine_WriteWord(m,off,cast<uint32_t>(v));
n3ds_Machine_WriteWord(m,cast<uint32_t>((off + cast<uint32_t>(4ULL))),cast<uint32_t>(shr<int64_t>(v,cast<int64_t>(32ULL))));
}
}c->R[cast<int64_t>(0ULL)] = cast<uint32_t>(0ULL);
}
}
// tools/platform/n3ds/svc.go:370:1
void n3ds_Machine_svcOutputDebugString(n3ds_Machine* m,arm_CPU* c){
{
auto tmp404 = std::make_tuple(c->R[cast<int64_t>(0ULL)],c->R[cast<int64_t>(1ULL)]);
uint32_t ptr = std::get<0>(tmp404);
uint32_t n = std::get<1>(tmp404);
{uint32_t i = cast<uint32_t>(0ULL);for (;((i < n) && (i < cast<uint32_t>(4096ULL)));i++){
m->debugOut = append(m->debugOut,n3ds_Machine_Read(m,cast<uint32_t>((ptr + i))));
}
}m->debugOut = append(m->debugOut,cast<uint8_t>(10ULL));
c->R[cast<int64_t>(0ULL)] = cast<uint32_t>(0ULL);
}
}
// tools/platform/n3ds/svc.go:380:1
std::string n3ds_Machine_readCString(n3ds_Machine* m,uint32_t addr,uint32_t max){
{
Slice<uint8_t> b={};
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < max);i++){
uint8_t ch = n3ds_Machine_Read(m,cast<uint32_t>((addr + i)));
if ((ch == cast<uint8_t>(0ULL))) {
break;
}
b = append(b,ch);
}
}return cast<std::string>(b);
}
}
// tools/platform/n3ds/svc.go:392:1
std::string n3ds_svcName(uint32_t n){
{
Map<uint32_t,std::string> names = Map<uint32_t,std::string>{{cast<uint32_t>(1ULL),std::string("ControlMemory",13)},{cast<uint32_t>(2ULL),std::string("QueryMemory",11)},{cast<uint32_t>(3ULL),std::string("ExitProcess",11)},{cast<uint32_t>(8ULL),std::string("CreateThread",12)},{cast<uint32_t>(9ULL),std::string("ExitThread",10)},{cast<uint32_t>(10ULL),std::string("SleepThread",11)},{cast<uint32_t>(11ULL),std::string("GetThreadPriority",17)},{cast<uint32_t>(12ULL),std::string("SetThreadPriority",17)},{cast<uint32_t>(19ULL),std::string("CreateMutex",11)},{cast<uint32_t>(20ULL),std::string("ReleaseMutex",12)},{cast<uint32_t>(21ULL),std::string("CreateSemaphore",15)},{cast<uint32_t>(22ULL),std::string("ReleaseSemaphore",16)},{cast<uint32_t>(23ULL),std::string("CreateEvent",11)},{cast<uint32_t>(24ULL),std::string("SignalEvent",11)},{cast<uint32_t>(25ULL),std::string("ClearEvent",10)},{cast<uint32_t>(26ULL),std::string("CreateTimer",11)},{cast<uint32_t>(30ULL),std::string("CreateMemoryBlock",17)},{cast<uint32_t>(31ULL),std::string("MapMemoryBlock",14)},{cast<uint32_t>(33ULL),std::string("CreateAddressArbiter",20)},{cast<uint32_t>(34ULL),std::string("ArbitrateAddress",16)},{cast<uint32_t>(35ULL),std::string("CloseHandle",11)},{cast<uint32_t>(36ULL),std::string("WaitSynchronization1",20)},{cast<uint32_t>(37ULL),std::string("WaitSynchronizationN",20)},{cast<uint32_t>(39ULL),std::string("DuplicateHandle",15)},{cast<uint32_t>(40ULL),std::string("GetSystemTick",13)},{cast<uint32_t>(42ULL),std::string("GetSystemInfo",13)},{cast<uint32_t>(43ULL),std::string("GetProcessInfo",14)},{cast<uint32_t>(44ULL),std::string("GetThreadInfo",13)},{cast<uint32_t>(45ULL),std::string("ConnectToPort",13)},{cast<uint32_t>(50ULL),std::string("SendSyncRequest",15)},{cast<uint32_t>(53ULL),std::string("GetProcessId",12)},{cast<uint32_t>(55ULL),std::string("GetThreadId",11)},{cast<uint32_t>(56ULL),std::string("GetResourceLimit",16)},{cast<uint32_t>(57ULL),std::string("GetResourceLimitCurrentValues",29)},{cast<uint32_t>(58ULL),std::string("GetResourceLimitLimitValues",27)},{cast<uint32_t>(60ULL),std::string("Break",5)},{cast<uint32_t>(61ULL),std::string("OutputDebugString",17)}};
{
auto tmp405 = lookup(names,n);
std::string s = std::get<0>(tmp405);
bool ok = std::get<1>(tmp405);
if (ok) {
return s;
}
}
return go_fmt_Sprintf(std::string("svc_0x%02X",10),n);
}
}
// tools/platform/n3ds/svc.go:421:1
Slice<n3ds_svcEvent> n3ds_Machine_SVCLog(n3ds_Machine* m){
{
return m->svcLog;
}
}
// tools/platform/n3ds/svc.go:424:1
Map<uint32_t,std::string> n3ds_Machine_Ports(n3ds_Machine* m){
{
return m->ports;
}
}
// tools/platform/n3ds/sync.go:22:1
bool n3ds_Machine_objAvailable(n3ds_Machine* m,n3ds_kobject* obj){
{
{
auto tmp407=obj->kind;
if (tmp407==(std::string("mutex",5))){
return ((obj->mutexOwner == cast<uint32_t>(0ULL)) || (obj->mutexOwner == m->curThread->id));
}
else if (tmp407==(std::string("semaphore",9))){
return (obj->semCount > cast<int32_t>(0ULL));
}
else {
return obj->signal;
}
}
tmp406:;
}
}
// tools/platform/n3ds/sync.go:35:1
void n3ds_Machine_consume(n3ds_Machine* m,n3ds_kobject* obj,n3ds_thread* t){
{
{
auto tmp409=obj->kind;
if (tmp409==(std::string("mutex",5))){
obj->mutexOwner = t->id;
obj->mutexDepth++;
}
else if (tmp409==(std::string("semaphore",9))){
obj->semCount--;
}
else if (tmp409==(std::string("event",5))){
if ((!obj->manualReset)) {
obj->signal = false;
}
}
}
tmp408:;
}
}
// tools/platform/n3ds/sync.go:57:1
void n3ds_Machine_svcWaitSync1(n3ds_Machine* m,arm_CPU* c){
{
uint32_t handle = c->R[cast<int64_t>(0ULL)];
int64_t timeoutNs = cast<int64_t>(cast<uint64_t>((cast<uint64_t>(c->R[cast<int64_t>(2ULL)]) | shl<uint64_t>(cast<uint64_t>(c->R[cast<int64_t>(3ULL)]),cast<int64_t>(32ULL)))));
n3ds_kobject* obj = get(m->handles,handle);
if ((!obj)) {
c->R[cast<int64_t>(0ULL)] = cast<uint32_t>(155196414ULL);
return ;
}
if (n3ds_Machine_objAvailable(m,obj)) {
n3ds_Machine_consume(m,obj,m->curThread);
c->R[cast<int64_t>(0ULL)] = cast<uint32_t>(0ULL);
return ;
}
if ((timeoutNs == cast<int64_t>(0ULL))) {
c->R[cast<int64_t>(0ULL)] = cast<uint32_t>(155196414ULL);
return ;
}
m->curThread->waitOn = Slice<uint32_t>{handle};
m->curThread->waitAll = false;
n3ds_Machine_armWaitDeadline(m,timeoutNs);
obj->waiters = append(obj->waiters,m->curThread->id);
m->curThread->state = cast<n3ds_threadState>(2ULL);
m->reschedule = true;
}
}
// tools/platform/n3ds/sync.go:86:1
void n3ds_Machine_armWaitDeadline(n3ds_Machine* m,int64_t ns){
{
if ((ns > cast<int64_t>(0ULL))) {
m->curThread->waitDeadline = cast<uint64_t>((m->tick + n3ds_nsToTick(ns)));
return ;
}
m->curThread->waitDeadline = cast<uint64_t>(0ULL);
}
}
// tools/platform/n3ds/sync.go:98:1
void n3ds_Machine_svcWaitSyncN(n3ds_Machine* m,arm_CPU* c){
{
uint32_t handlesPtr = c->R[cast<int64_t>(1ULL)];
int64_t count = cast<int64_t>(c->R[cast<int64_t>(2ULL)]);
bool waitAll = (c->R[cast<int64_t>(3ULL)] != cast<uint32_t>(0ULL));
int64_t timeoutNs = cast<int64_t>(cast<uint64_t>((cast<uint64_t>(c->R[cast<int64_t>(0ULL)]) | shl<uint64_t>(cast<uint64_t>(c->R[cast<int64_t>(4ULL)]),cast<int64_t>(32ULL)))));
Slice<uint32_t> handles = Slice<uint32_t>::make(count);
{int64_t i = cast<int64_t>(0ULL);for (;(i < count);i++){
handles[i] = n3ds_Machine_ReadWord(m,cast<uint32_t>((handlesPtr + cast<uint32_t>((cast<uint32_t>(i) * cast<uint32_t>(4ULL))))));
}
}if (waitAll) {
bool all = true;
{auto&& tmp410 = handles;
for(int64_t tmp411=0;tmp411<len(tmp410);++tmp411){
auto h=tmp410[tmp411];{
n3ds_kobject* obj = get(m->handles,h);
if (((!obj) || (!n3ds_Machine_objAvailable(m,obj)))) {
all = false;
break;
}
}
}}
if (all) {
{auto&& tmp412 = handles;
for(int64_t tmp413=0;tmp413<len(tmp412);++tmp413){
auto h=tmp412[tmp413];n3ds_Machine_consume(m,get(m->handles,h),m->curThread);
}}
c->R[cast<int64_t>(0ULL)] = cast<uint32_t>(0ULL);
return ;
}
}
else {
{auto&& tmp414 = handles;
for(int64_t tmp415=0;tmp415<len(tmp414);++tmp415){
auto i=tmp415;auto h=tmp414[tmp415];{
n3ds_kobject* obj = get(m->handles,h);
if ((bool(obj) && n3ds_Machine_objAvailable(m,obj))) {
n3ds_Machine_consume(m,obj,m->curThread);
auto tmp416 = std::make_tuple(cast<uint32_t>(0ULL),cast<uint32_t>(i));
c->R[cast<int64_t>(0ULL)] = std::get<0>(tmp416);
c->R[cast<int64_t>(1ULL)] = std::get<1>(tmp416);
return ;
}
}
}}
}
if ((timeoutNs == cast<int64_t>(0ULL))) {
c->R[cast<int64_t>(0ULL)] = cast<uint32_t>(155196414ULL);
return ;
}
m->curThread->waitOn = handles;
m->curThread->waitAll = waitAll;
n3ds_Machine_armWaitDeadline(m,timeoutNs);
{auto&& tmp417 = handles;
for(int64_t tmp418=0;tmp418<len(tmp417);++tmp418){
auto h=tmp417[tmp418];{
n3ds_kobject* obj = get(m->handles,h);
if (bool(obj)) {
obj->waiters = append(obj->waiters,m->curThread->id);
}
}
}}
m->curThread->state = cast<n3ds_threadState>(2ULL);
m->reschedule = true;
}
}
// tools/platform/n3ds/sync.go:153:1
bool n3ds_Machine_signalObject(n3ds_Machine* m,n3ds_kobject* obj){
{
bool resched = false;
int64_t i = cast<int64_t>(0ULL);
{;for (;(i < len(obj->waiters));){
uint32_t tid = obj->waiters[i];
n3ds_thread* t = n3ds_Machine_threadByID(m,tid);
if (((!t) || (t->state != cast<n3ds_threadState>(2ULL)))) {
obj->waiters = append(sub(obj->waiters,0,i),sub(obj->waiters,cast<int64_t>((i + cast<int64_t>(1ULL))),len(obj->waiters)));
continue;
}
if (n3ds_Machine_tryComplete(m,t)) {
obj->waiters = append(sub(obj->waiters,0,i),sub(obj->waiters,cast<int64_t>((i + cast<int64_t>(1ULL))),len(obj->waiters)));
if (n3ds_Machine_wake(m,t)) {
resched = true;
}
if ((!n3ds_Machine_objAvailable(m,obj))) {
break;
}
continue;
}
i++;
}
}return resched;
}
}
// tools/platform/n3ds/sync.go:184:1
bool n3ds_Machine_tryComplete(n3ds_Machine* m,n3ds_thread* t){
{
if (t->waitAll) {
{auto&& tmp419 = t->waitOn;
for(int64_t tmp420=0;tmp420<len(tmp419);++tmp420){
auto h=tmp419[tmp420];{
n3ds_kobject* obj = get(m->handles,h);
if (((!obj) || (!n3ds_Machine_availableFor(m,obj,t)))) {
return false;
}
}
}}
{auto&& tmp421 = t->waitOn;
for(int64_t tmp422=0;tmp422<len(tmp421);++tmp422){
auto h=tmp421[tmp422];n3ds_Machine_consumeFor(m,get(m->handles,h),t);
}}
n3ds_Machine_setResult(m,t,cast<int64_t>(0ULL),cast<uint32_t>(0ULL));
return true;
}
{auto&& tmp423 = t->waitOn;
for(int64_t tmp424=0;tmp424<len(tmp423);++tmp424){
auto i=tmp424;auto h=tmp423[tmp424];{
n3ds_kobject* obj = get(m->handles,h);
if ((bool(obj) && n3ds_Machine_availableFor(m,obj,t))) {
n3ds_Machine_consumeFor(m,obj,t);
n3ds_Machine_setResult(m,t,cast<int64_t>(0ULL),cast<uint32_t>(0ULL));
n3ds_Machine_setResult(m,t,cast<int64_t>(1ULL),cast<uint32_t>(i));
return true;
}
}
}}
return false;
}
}
// tools/platform/n3ds/sync.go:210:1
bool n3ds_Machine_availableFor(n3ds_Machine* m,n3ds_kobject* obj,n3ds_thread* t){
{
{
auto tmp426=obj->kind;
if (tmp426==(std::string("mutex",5))){
return ((obj->mutexOwner == cast<uint32_t>(0ULL)) || (obj->mutexOwner == t->id));
}
else if (tmp426==(std::string("semaphore",9))){
return (obj->semCount > cast<int32_t>(0ULL));
}
else {
return obj->signal;
}
}
tmp425:;
}
}
// tools/platform/n3ds/sync.go:221:1
void n3ds_Machine_consumeFor(n3ds_Machine* m,n3ds_kobject* obj,n3ds_thread* t){
{
{
auto tmp428=obj->kind;
if (tmp428==(std::string("mutex",5))){
obj->mutexOwner = t->id;
obj->mutexDepth++;
}
else if (tmp428==(std::string("semaphore",9))){
obj->semCount--;
}
else if (tmp428==(std::string("event",5))){
if ((!obj->manualReset)) {
obj->signal = false;
}
}
}
tmp427:;
}
}
// tools/platform/n3ds/sync.go:235:1
n3ds_thread* n3ds_Machine_threadByID(n3ds_Machine* m,uint32_t id){
{
{auto&& tmp429 = m->threads;
for(int64_t tmp430=0;tmp430<len(tmp429);++tmp430){
auto t=tmp429[tmp430];if ((t->id == id)) {
return t;
}
}}
return {};
}
}
// tools/platform/n3ds/sync.go:246:1
void n3ds_Machine_svcSignalEvent(n3ds_Machine* m,arm_CPU* c){
{
n3ds_kobject* obj = get(m->handles,c->R[cast<int64_t>(0ULL)]);
if (bool(obj)) {
obj->signal = true;
if (n3ds_Machine_signalObject(m,obj)) {
m->reschedule = true;
}
}
c->R[cast<int64_t>(0ULL)] = cast<uint32_t>(0ULL);
}
}
// tools/platform/n3ds/sync.go:257:1
void n3ds_Machine_svcClearEvent(n3ds_Machine* m,arm_CPU* c){
{
{
n3ds_kobject* obj = get(m->handles,c->R[cast<int64_t>(0ULL)]);
if (bool(obj)) {
obj->signal = false;
}
}
c->R[cast<int64_t>(0ULL)] = cast<uint32_t>(0ULL);
}
}
// tools/platform/n3ds/sync.go:264:1
void n3ds_Machine_svcReleaseMutex(n3ds_Machine* m,arm_CPU* c){
{
n3ds_kobject* obj = get(m->handles,c->R[cast<int64_t>(0ULL)]);
if ((bool(obj) && (obj->kind == std::string("mutex",5)))) {
if ((obj->mutexDepth > cast<int64_t>(0ULL))) {
obj->mutexDepth--;
}
if ((obj->mutexDepth == cast<int64_t>(0ULL))) {
obj->mutexOwner = cast<uint32_t>(0ULL);
if (n3ds_Machine_signalObject(m,obj)) {
m->reschedule = true;
}
}
}
c->R[cast<int64_t>(0ULL)] = cast<uint32_t>(0ULL);
}
}
// tools/platform/n3ds/sync.go:280:1
void n3ds_Machine_svcReleaseSemaphore(n3ds_Machine* m,arm_CPU* c){
{
n3ds_kobject* obj = get(m->handles,c->R[cast<int64_t>(1ULL)]);
if ((bool(obj) && (obj->kind == std::string("semaphore",9)))) {
int32_t prev = obj->semCount;
obj->semCount += cast<int32_t>(c->R[cast<int64_t>(2ULL)]);
if (n3ds_Machine_signalObject(m,obj)) {
m->reschedule = true;
}
auto tmp431 = std::make_tuple(cast<uint32_t>(0ULL),cast<uint32_t>(prev));
c->R[cast<int64_t>(0ULL)] = std::get<0>(tmp431);
c->R[cast<int64_t>(1ULL)] = std::get<1>(tmp431);
return ;
}
c->R[cast<int64_t>(0ULL)] = cast<uint32_t>(0ULL);
}
}
// tools/platform/n3ds/sync.go:297:1
void n3ds_Machine_signalThreadExit(n3ds_Machine* m,n3ds_thread* t){
{
{auto&& tmp432 = m->handles;
for(auto [tmp433,tmp434]:tmp432){
auto o=tmp434;if (((o->kind == std::string("thread",6)) && (o->thread == t))) {
o->signal = true;
if (n3ds_Machine_signalObject(m,o)) {
m->reschedule = true;
}
}
}}
}
}
// tools/platform/n3ds/sync.go:327:1
void n3ds_Machine_svcArbitrateAddress(n3ds_Machine* m,arm_CPU* c){
{
uint32_t addr = c->R[cast<int64_t>(1ULL)];
uint32_t typ = c->R[cast<int64_t>(2ULL)];
int32_t value = cast<int32_t>(c->R[cast<int64_t>(3ULL)]);
{
switch(typ){
case cast<uint32_t>(0ULL):{
n3ds_Machine_arbSignal(m,addr,value);
c->R[cast<int64_t>(0ULL)] = cast<uint32_t>(0ULL);
break;}
case cast<uint32_t>(1ULL):case cast<uint32_t>(3ULL):{
if ((cast<int32_t>(n3ds_Machine_ReadWord(m,addr)) < value)) {
n3ds_Machine_arbPark(m,addr,(typ == cast<uint32_t>(3ULL)));
}
else {
c->R[cast<int64_t>(0ULL)] = cast<uint32_t>(0ULL);
}
break;}
case cast<uint32_t>(2ULL):case cast<uint32_t>(4ULL):{
int32_t cur = cast<int32_t>(n3ds_Machine_ReadWord(m,addr));
if ((cur <= value)) {
n3ds_Machine_WriteWord(m,addr,cast<uint32_t>(cast<int32_t>((cur - cast<int32_t>(1ULL)))));
n3ds_Machine_arbPark(m,addr,(typ == cast<uint32_t>(4ULL)));
}
else {
c->R[cast<int64_t>(0ULL)] = cast<uint32_t>(0ULL);
}
break;}
default:{
c->R[cast<int64_t>(0ULL)] = cast<uint32_t>(0ULL);
break;}
}}
}
}
// tools/platform/n3ds/sync.go:361:1
void n3ds_Machine_arbPark(n3ds_Machine* m,uint32_t addr,bool timed){
{
if ((timed && (!get(m->arbTimedWarned,addr)))) {
if ((!m->arbTimedWarned)) {
m->arbTimedWarned = Map<uint32_t,bool>{};
}
m->arbTimedWarned[addr] = true;
go_fmt_Printf((std::string("n3ds: thread %d parked on a TIMED arbiter wait (arb@0x%08X) \342\200\224 the timeout is not modelled, ",93) + std::string("so this thread will only wake if something signals the address\012",63)),m->curThread->id,addr);
}
m->curThread->arbAddr = addr;
m->curThread->state = cast<n3ds_threadState>(2ULL);
m->reschedule = true;
}
}
// tools/platform/n3ds/sync.go:375:1
void n3ds_Machine_arbSignal(n3ds_Machine* m,uint32_t addr,int32_t count){
{
int32_t n={};
{auto&& tmp435 = m->threads;
for(int64_t tmp436=0;tmp436<len(tmp435);++tmp436){
auto t=tmp435[tmp436];if (((t->state == cast<n3ds_threadState>(2ULL)) && (t->arbAddr == addr))) {
if (((count >= cast<int32_t>(0ULL)) && (n >= count))) {
break;
}
n3ds_Machine_setResult(m,t,cast<int64_t>(0ULL),cast<uint32_t>(0ULL));
if (n3ds_Machine_wake(m,t)) {
m->reschedule = true;
}
n++;
}
}}
}
}
// tools/platform/n3ds/thread.go:39:1
std::string n3ds_threadState_String(n3ds_threadState s){
{
return std::array<std::string,5>{std::string("ready",5),std::string("running",7),std::string("waiting",7),std::string("sleeping",8),std::string("dead",4)}[s];
}
}
// tools/platform/n3ds/thread.go:73:1
uint32_t n3ds_Machine_allocTLS(n3ds_Machine* m){
{
uint32_t base = m->nextTLS;
m->nextTLS += cast<uint32_t>(4096ULL);
n3ds_Machine_mapRegion(m,std::string("tls-thread",10),base,Slice<uint8_t>::make(cast<int64_t>(4096ULL)));
return base;
}
}
// tools/platform/n3ds/thread.go:81:1
uint32_t n3ds_Machine_cmdBuf(n3ds_Machine* m){
{
return cast<uint32_t>((m->curThread->tlsBase + cast<uint32_t>(128ULL)));
}
}
// tools/platform/n3ds/thread.go:86:1
uint32_t n3ds_Machine_createThread(n3ds_Machine* m,int32_t priority,uint32_t entry,uint32_t arg,uint32_t stacktop){
{
n3ds_thread* t = arenaNew(n3ds_thread{m->nextThread,{},{},n3ds_Machine_allocTLS(m),{},priority,cast<n3ds_threadState>(0ULL),{},{},{},{},{}});
m->nextThread++;
t->ctx = (*m->CPU);
t->ctx.R = std::array<uint32_t,16>{};
t->ctx.R[cast<int64_t>(0ULL)] = arg;
t->ctx.R[cast<int64_t>(13ULL)] = stacktop;
t->ctx.R[cast<int64_t>(14ULL)] = cast<uint32_t>(4294901760ULL);
t->ctx.R[cast<int64_t>(15ULL)] = (entry & ~(cast<uint32_t>(1ULL)));
t->ctx.Thumb = (cast<uint32_t>((entry & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL));
t->ctx.Mode = cast<uint32_t>(31ULL);
t->ctx.IRQDisable = false;
t->ctx.Halted = false;
uint32_t h = n3ds_Machine_newHandle(m,std::string("thread",6),false);
m->handles[h]->thread = t;
t->handle = h;
m->threads = append(m->threads,t);
if (m->Verbose) {
go_fmt_Printf(std::string("  createThread id=%d entry=0x%08X arg=0x%08X sp=0x%08X prio=%d -> handle 0x%08X\012",80),t->id,entry,arg,stacktop,priority,h);
}
return h;
}
}
// tools/platform/n3ds/thread.go:119:1
void n3ds_Machine_svcSleepThread(n3ds_Machine* m,arm_CPU* c){
{
int64_t ns = cast<int64_t>(cast<uint64_t>((cast<uint64_t>(c->R[cast<int64_t>(0ULL)]) | shl<uint64_t>(cast<uint64_t>(c->R[cast<int64_t>(1ULL)]),cast<int64_t>(32ULL)))));
c->R[cast<int64_t>(0ULL)] = cast<uint32_t>(0ULL);
{
if ((ns == cast<int64_t>(0ULL))){
m->curThread->state = cast<n3ds_threadState>(0ULL);
}
else if ((ns < cast<int64_t>(0ULL))){
m->curThread->state = cast<n3ds_threadState>(2ULL);
}
else {
m->curThread->state = cast<n3ds_threadState>(3ULL);
m->curThread->wakeTick = cast<uint64_t>((m->tick + n3ds_nsToTick(ns)));
}
}
tmp437:;
m->reschedule = true;
}
}
// tools/platform/n3ds/thread.go:135:1
void n3ds_Machine_svcExitThread(n3ds_Machine* m,arm_CPU* c){
{
m->curThread->state = cast<n3ds_threadState>(4ULL);
n3ds_Machine_signalThreadExit(m,m->curThread);
m->reschedule = true;
}
}
// tools/platform/n3ds/thread.go:143:1
n3ds_thread* n3ds_Machine_pickRunnable(n3ds_Machine* m){
{
n3ds_thread* best={};
int64_t n = len(m->threads);
{int64_t i = cast<int64_t>(0ULL);for (;(i < n);i++){
n3ds_thread* t = m->threads[modi<int64_t>((cast<int64_t>((m->rrCursor + i))),n)];
if (((t->state == cast<n3ds_threadState>(0ULL)) && (((!best) || (t->priority < best->priority))))) {
best = t;
}
}
}if (bool(best)) {
m->rrCursor = modi<int64_t>((cast<int64_t>((m->rrCursor + cast<int64_t>(1ULL)))),n);
}
return best;
}
}
// tools/platform/n3ds/thread.go:159:1
void n3ds_Machine_switchTo(n3ds_Machine* m,n3ds_thread* t){
{
if ((m->curThread == t)) {
t->state = cast<n3ds_threadState>(1ULL);
return ;
}
(*m->CPU) = t->ctx;
arm_CPU_ClearExclusive(m->CPU);
m->curThread = t;
t->state = cast<n3ds_threadState>(1ULL);
}
}
// tools/platform/n3ds/thread.go:176:1
std::tuple<uint64_t,bool> n3ds_Machine_soonestSleeper(n3ds_Machine* m){
{
uint64_t soonest={};
bool found = false;
auto at = [&](uint64_t tick)->void{
if (((!found) || (tick < soonest))) {
auto tmp438 = std::make_tuple(tick,true);
soonest = std::get<0>(tmp438);
found = std::get<1>(tmp438);
}
}
;
{auto&& tmp439 = m->threads;
for(int64_t tmp440=0;tmp440<len(tmp439);++tmp440){
auto t=tmp439[tmp440];{
if ((t->state == cast<n3ds_threadState>(3ULL))){
at(t->wakeTick);
}
else if (((t->state == cast<n3ds_threadState>(2ULL)) && (t->waitDeadline != cast<uint64_t>(0ULL)))){
at(t->waitDeadline);
}
}
tmp441:;
}}
return {soonest,found};
}
}
// tools/platform/n3ds/thread.go:208:1
void n3ds_Machine_wakeDueSleepers(n3ds_Machine* m){
{
{auto&& tmp442 = m->threads;
for(int64_t tmp443=0;tmp443<len(tmp442);++tmp443){
auto t=tmp442[tmp443];{
if (((t->state == cast<n3ds_threadState>(3ULL)) && (t->wakeTick <= m->tick))){
t->state = cast<n3ds_threadState>(0ULL);
}
else if ((((t->state == cast<n3ds_threadState>(2ULL)) && (t->waitDeadline != cast<uint64_t>(0ULL))) && (t->waitDeadline <= m->tick))){
n3ds_Machine_setResult(m,t,cast<int64_t>(0ULL),cast<uint32_t>(155196414ULL));
n3ds_Machine_wake(m,t);
}
}
tmp444:;
}}
}
}
// tools/platform/n3ds/thread.go:226:1
bool n3ds_Machine_wake(n3ds_Machine* m,n3ds_thread* t){
{
t->state = cast<n3ds_threadState>(0ULL);
t->waitOn = {};
t->arbAddr = cast<uint32_t>(0ULL);
t->waitDeadline = cast<uint64_t>(0ULL);
return ((!m->curThread) || (t->priority < m->curThread->priority));
}
}
// tools/platform/n3ds/thread.go:237:1
void n3ds_Machine_setResult(n3ds_Machine* m,n3ds_thread* t,int64_t reg,uint32_t v){
{
if ((t == m->curThread)) {
m->CPU->R[reg] = v;
}
else {
t->ctx.R[reg] = v;
}
}
}
// tools/platform/n3ds/thread.go:248:1
void n3ds_Machine_DumpThreads(n3ds_Machine* m){
{
n3ds_Machine_dumpThreads(m);
go_fmt_Printf(std::string("dsp: componentLoaded=%v state=%d semEvent=0x%08X ticks=%d (instrs=%d)\012",70),m->dsp.ComponentLoaded,m->dsp.State,m->dsp.SemEvent,m->dsp.Ticks,m->instrs);
{auto&& tmp445 = m->dsp.Sources;
for(int64_t tmp446=0;tmp446<len(tmp445);++tmp446){
auto i=tmp446;n3ds_dspSource* s = (&m->dsp.Sources[i]);
if (((s->Enabled || (len(s->Queue) > cast<int64_t>(0ULL))) || (s->CurBufferID != cast<uint16_t>(0ULL)))) {
go_fmt_Printf(std::string("  dsp src %2d: enabled=%v sync=%d rate=%g fmt=%d stereo=%v pos=%d remain=%d cur=%d last=%d queued=%d update=%v\012",111),i,s->Enabled,s->SyncCount,s->Rate,s->Format,s->Stereo,s->CurSample,len(s->CurBuf),s->CurBufferID,s->LastBufferID,len(s->Queue),s->BufferUpdate);
}
}}
go_fmt_Printf(std::string("handles:\012",9));
{auto&& tmp447 = m->handles;
for(auto [tmp448,tmp449]:tmp447){
auto h=tmp448;auto o=tmp449;std::string extra = std::string("",0);
if (o->signal) {
extra = std::string(" signalled",10);
}
if (((o->kind == std::string("thread",6)) && bool(o->thread))) {
extra += go_fmt_Sprintf(std::string(" (thread %d)",12),o->thread->id);
}
go_fmt_Printf(std::string("  0x%08X %-24s%s\012",17),h,o->kind,extra);
}}
}
}
// tools/platform/n3ds/thread.go:275:1
void n3ds_Machine_dumpThreads(n3ds_Machine* m){
{
go_fmt_Printf(std::string("thread states at deadlock (%d GX commands pending):\012",52),len(m->gxPending));
{auto&& tmp450 = m->threads;
for(int64_t tmp451=0;tmp451<len(tmp450);++tmp451){
auto t=tmp450[tmp451];std::string wo = std::string("",0);
{auto&& tmp452 = t->waitOn;
for(int64_t tmp453=0;tmp453<len(tmp452);++tmp453){
auto h=tmp452[tmp453];std::string kind = std::string("?",1);
{
n3ds_kobject* o = get(m->handles,h);
if (bool(o)) {
kind = o->kind;
}
}
wo += go_fmt_Sprintf(std::string(" 0x%08X(%s)",11),h,kind);
}}
if ((t->arbAddr != cast<uint32_t>(0ULL))) {
wo += go_fmt_Sprintf(std::string(" arb@0x%08X",11),t->arbAddr);
}
go_fmt_Printf(std::string("  thread %d prio %d state %-8s pc=0x%08X sp=0x%08X lr=0x%08X waitOn:%s\012",71),t->id,t->priority,t->state,t->ctx.R[cast<int64_t>(15ULL)],t->ctx.R[cast<int64_t>(13ULL)],t->ctx.R[cast<int64_t>(14ULL)],wo);
}}
}
}
// tools/platform/n3ds/thread.go:295:1
int64_t n3ds_Machine_aliveThreads(n3ds_Machine* m){
{
int64_t n = cast<int64_t>(0ULL);
{auto&& tmp454 = m->threads;
for(int64_t tmp455=0;tmp455<len(tmp454);++tmp455){
auto t=tmp454[tmp455];if ((t->state != cast<n3ds_threadState>(4ULL))) {
n++;
}
}}
return n;
}
}
// tools/platform/n3ds/thread.go:311:1
uint64_t n3ds_nsToTick(int64_t ns){
{
if ((ns <= cast<int64_t>(0ULL))) {
return cast<uint64_t>(0ULL);
}
return divi<uint64_t>(cast<uint64_t>((cast<uint64_t>(ns) * cast<uint64_t>(268111856ULL))),cast<uint64_t>(1000000000ULL));
}
}
