#include "runtime.h"
struct x86_CPU;
struct x86_ea;
struct x86_FPUState;
struct xbox_mmioLatch;
struct xbox_irqSource;
struct xbox_dpcEntry;
struct xbox_kernel_132_kv;
struct xbox_cacheFile;
struct xbox_fileObject;
struct xbox_dirEntry;
struct xbox_pendingIO;
struct xbox_Machine;
struct xbox_nv2a;
struct xbox_surfaces2D;
struct xbox_imageBlit;
struct xbox_combInput;
struct xbox_combStage;
struct xbox_combState;
struct xbox_combRegs;
struct xbox_RegionSpec;
struct xbox_presentedSurface;
struct xbox_pusherState;
struct xbox_nv2a_pgraph_320_row;
struct xbox_pgraph;
struct xbox_zetaBucket;
struct xbox_PixelEvent;
struct xbox_rasterState;
struct xbox_rstats;
struct xbox_texKey;
struct xbox_texEntry;
struct xbox_texImage;
struct xbox_shadowFragStats;
struct xbox_kelvinVtx;
struct xbox_vtxFmt;
struct xbox_vshSrc;
struct xbox_vshInst;
struct xbox_vshState;
struct xbox_ProfileBucket;
struct xbox_ProfileCounter;
struct xbox_FrameProfile;
struct xbox_profCounters;
struct xbox_profState;
struct xbox_thread;
struct xbox_kobject;
struct xbox_ktimer;
struct xbox_PadControl;
struct xbox_PadState;
struct xbox_xidDevice;
struct xbox_Section;
struct xbox_XBE;
struct xbox_Image;
struct xbox_Entry;
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
struct xbox_mmioLatch{
std::string name{};
Map<uint32_t,uint32_t> reg{};
Map<uint32_t,bool> seenCold{};
};
struct xbox_irqSource{
std::string name{};
uint32_t vector{};
std::function<bool(xbox_Machine*)> pending{};
};
struct xbox_dpcEntry{
uint32_t Dpc{};
uint32_t Arg1{};
uint32_t Arg2{};
};
struct xbox_kernel_132_kv{
uint32_t pc{};
uint64_t c{};
};
struct xbox_cacheFile{
Slice<uint8_t> Data{};
};
struct xbox_Entry{
std::string Name{};
std::string Path{};
bool IsDir{};
uint32_t Size{};
uint32_t Sector{};
uint8_t Attr{};
};
struct xbox_fileObject{
xbox_Entry entry{};
xbox_cacheFile* cache{};
std::string key{};
bool dir{};
uint32_t off{};
int64_t scan{};
};
struct xbox_dirEntry{
std::string Name{};
uint32_t Size{};
bool IsDir{};
};
struct xbox_pendingIO{
uint64_t Due{};
uint32_t IOSB{};
uint32_t Event{};
uint32_t Info{};
uint32_t Status{};
uint32_t Handle{};
};
struct xbox_nv2a{
Map<uint32_t,uint32_t> reg{};
uint32_t dmaPut{};
uint32_t dmaGet{};
bool kicked{};
uint32_t pcrtcIntr{};
uint32_t dispMode{};
uint32_t dispFormat{};
uint32_t fbPitch{};
uint32_t fbAddr{};
};
struct xbox_pusherState{
uint32_t method{};
uint32_t subchan{};
uint32_t count{};
bool nonInc{};
uint32_t subReturn{};
bool subActive{};
bool running{};
};
struct xbox_ktimer{
uint32_t Timer{};
uint32_t Dpc{};
uint64_t Due{};
uint32_t Period{};
};
struct xbox_profCounters{
int64_t draws{};
int64_t methods{};
int64_t frags{};
int64_t zRej{};
int64_t aRej{};
};
struct xbox_ProfileBucket{
std::string Name{};
double Millis{};
int64_t Count{};
};
struct xbox_ProfileCounter{
std::string Name{};
int64_t Value{};
};
struct xbox_FrameProfile{
double TotalMs{};
Slice<xbox_ProfileBucket> Buckets{};
Slice<xbox_ProfileCounter> Counters{};
bool Drew{};
};
struct xbox_profState{
std::array<int64_t,4> ns{};
std::array<int64_t,4> count{};
time_Time runStart{};
bool inRun{};
int64_t frameNs{};
time_Time pushStart{};
bool inPush{};
xbox_profCounters base{};
uint64_t baseInstr{};
xbox_FrameProfile last{};
bool has{};
};
struct xbox_Machine{
Slice<uint8_t> RAM{};
x86_CPU* CPU{};
xbox_XBE* XBE{};
xbox_Image* Disc{};
uint32_t poolNext{};
uint32_t heapNext{};
uint32_t heapTop{};
Map<uint32_t,xbox_kobject*> objects{};
Map<uint32_t,xbox_fileObject*> files{};
Map<std::string,xbox_cacheFile*> cacheFS{};
Map<std::string,Slice<uint8_t>> fileBasic{};
Slice<xbox_pendingIO> pendingIO{};
Map<uint32_t,uint32_t> poolSizes{};
uint32_t nextObjAddr{};
uint32_t kbandNext{};
Slice<xbox_thread*> threads{};
xbox_thread* current{};
uint32_t nextTID{};
int64_t rrCursor{};
bool reschedule{};
int64_t quantumLeft{};
Map<uint32_t,uint32_t> interrupts{};
uint64_t nextVBlank{};
bool isrActive{};
x86_CPU isrSaved{};
Slice<xbox_dpcEntry> dpcQueue{};
uint64_t tick{};
uint64_t clockBaseTick{};
uint64_t clockBase100ns{};
uint64_t tscBase{};
uint32_t tickCountAddr{};
uint32_t systemTimeAddr{};
Map<uint32_t,Slice<uint8_t>> shaCtx{};
Map<uint32_t,Slice<uint8_t>> rc4Ctx{};
xbox_nv2a nv{};
xbox_pusherState push{};
xbox_pgraph* pgraph{};
xbox_mmioLatch apu{};
xbox_mmioLatch ac97{};
xbox_mmioLatch usb{};
xbox_mmioLatch nic{};
uint64_t usbFrameServed{};
uint32_t usbWrDword{};
Slice<xbox_ktimer> timers{};
std::array<xbox_xidDevice*,4> usbDev{};
Slice<uint32_t> usbDone{};
Slice<uint8_t> usbCtrlData{};
int64_t usbCtrlOff{};
uint32_t pciAddr{};
Map<uint32_t,uint8_t> pciSpace{};
Slice<std::string> Log{};
Map<uint16_t,int64_t> OrdinalHits{};
Map<uint16_t,bool> dataDeref{};
Map<uint32_t,bool> dosErrWarned{};
bool Halted{};
std::string HaltReason{};
bool firstPush{};
bool pusherEnabled{};
uint32_t wWLo{};
uint32_t wWHi{};
std::function<void(uint32_t,uint32_t,uint32_t)> onW{};
uint32_t wRLo{};
uint32_t wRHi{};
std::function<void(uint32_t,uint32_t,uint32_t)> onR{};
std::function<void(uint32_t,uint32_t,xbox_PixelEvent)> OnPixel{};
std::function<void(xbox_Machine*,uint32_t,uint32_t,uint32_t)> OnNVMethod{};
std::function<void(xbox_Machine*)> OnFlip{};
bool FlipVSync{};
bool StopRequested{};
int64_t stopAfterMethod{};
bool stopAfterArmed{};
Map<uint32_t,bool> bps{};
bool verbose{};
int64_t traceLeft{};
uint32_t bpAddr{};
int64_t bpLeft{};
Map<uint32_t,uint64_t> hotpc{};
bool Profile{};
xbox_profState prof{};
};
struct xbox_surfaces2D{
uint32_t format{};
uint32_t srcPitch{};
uint32_t dstPitch{};
uint32_t srcOffset{};
uint32_t dstOffset{};
};
struct xbox_imageBlit{
uint32_t inX{};
uint32_t inY{};
uint32_t outX{};
uint32_t outY{};
};
struct xbox_combInput{
std::array<std::array<float,4>,4> tex{};
std::array<float,4> col0{};
std::array<float,4> col1{};
};
struct xbox_combStage{
uint32_t colorICW{};
uint32_t alphaICW{};
uint32_t colorOCW{};
uint32_t alphaOCW{};
std::array<float,4> factor0{};
std::array<float,4> factor1{};
};
struct xbox_combState{
Slice<xbox_combStage> stages{};
uint32_t cw0{};
uint32_t cw1{};
std::array<float,4> fogColor{};
std::array<xbox_combStage,8> buf{};
};
struct xbox_combRegs{
std::array<std::array<float,4>,16> r{};
};
struct xbox_RegionSpec{
uint32_t Addr{};
int64_t W{};
int64_t H{};
int64_t Stride{};
std::string Format{};
};
struct xbox_presentedSurface{
bool valid{};
uint32_t base{};
uint32_t pitch{};
int64_t w{};
int64_t h{};
int64_t ax{};
int64_t ay{};
};
struct xbox_nv2a_pgraph_320_row{
uint32_t key{};
int64_t count{};
};
struct xbox_rasterState{
uint32_t colorPhys{};
uint32_t zetaPhys{};
uint32_t colorPitch{};
uint32_t zetaPitch{};
int64_t aaX{};
int64_t aaY{};
int64_t x0{};
int64_t y0{};
int64_t x1{};
int64_t y1{};
int64_t surfW{};
int64_t surfH{};
bool swizzle{};
int64_t swW{};
int64_t swH{};
bool hasZeta{};
bool depthTest{};
bool depthWrite{};
uint32_t depthFunc{};
bool alphaTest{};
uint32_t alphaFunc{};
float alphaRef{};
bool blend{};
uint32_t blendSrc{};
uint32_t blendDst{};
uint32_t blendEq{};
std::array<float,4> blendConst{};
std::array<bool,4> colorMask{};
bool cull{};
uint32_t cullFace{};
bool frontCW{};
bool flatShade{};
xbox_combState comb{};
std::array<bool,4> texEnable{};
std::array<xbox_texImage*,4> texImg{};
std::array<uint32_t,4> texStage{};
std::array<uint32_t,4> texWrapU{};
std::array<uint32_t,4> texWrapV{};
std::array<bool,4> texBilinear{};
std::array<bool,4> texRect{};
};
struct xbox_vshSrc{
int64_t mux{};
int64_t reg{};
bool neg{};
std::array<int64_t,4> swz{};
};
struct xbox_vshInst{
int64_t mac{};
int64_t ilu{};
int64_t constIdx{};
int64_t inputReg{};
xbox_vshSrc a{};
xbox_vshSrc b{};
xbox_vshSrc c{};
uint32_t macMask{};
int64_t macDst{};
uint32_t iluMask{};
uint32_t outMask{};
bool outIsO{};
int64_t outAddr{};
bool outFromILU{};
bool relConst{};
bool final{};
};
struct xbox_kelvinVtx{
std::array<float,4> pos{};
std::array<float,4> d0{};
std::array<float,4> d1{};
float fog{};
std::array<std::array<float,4>,4> uv{};
};
struct xbox_pgraph{
xbox_Machine* m{};
std::array<uint32_t,8> subObject{};
std::array<uint32_t,8> subClass{};
std::array<uint32_t,2048> Regs{};
std::array<std::array<uint32_t,4>,136> Prog{};
std::array<std::array<uint32_t,4>,192> Const{};
uint32_t ProgLoad{};
uint32_t ConstLoad{};
std::array<uint32_t,4> progBuf{};
int64_t progBufN{};
std::array<uint32_t,4> constBuf{};
int64_t constBufN{};
uint32_t prim{};
Slice<uint32_t> inline_{};
Slice<uint32_t> elems{};
Slice<std::array<uint32_t,2>> ranges{};
std::array<std::array<float,4>,16> vtxAttr{};
int64_t Draws{};
int64_t pixWritten{};
int64_t pixZRej{};
int64_t pixARej{};
int64_t lowWriteDraw{};
std::string ffFragHalt{};
xbox_presentedSurface presented{};
xbox_rasterState rast{};
bool rastValid{};
Map<xbox_texKey,xbox_texEntry*> texCache{};
uint64_t texRun{};
xbox_surfaces2D surf2D{};
xbox_imageBlit blit{};
bool survey{};
Map<uint32_t,int64_t> seen{};
Map<uint32_t,uint32_t> firstArg{};
int64_t Methods{};
int64_t SetObjs{};
Map<uint32_t,int64_t> unhandled{};
Map<uint32_t,xbox_zetaBucket*> zetaHist{};
bool shadowDumped{};
xbox_shadowFragStats* shadowFrag{};
Slice<xbox_vshInst> vshProg{};
bool vshWritesConst{};
Slice<std::array<std::array<float,4>,16>> vin{};
Slice<xbox_kelvinVtx> vertsBuf{};
Slice<std::array<int64_t,3>> triScratch{};
Slice<std::array<int64_t,3>> clipTris{};
Slice<xbox_kelvinVtx> clipVerts{};
};
struct xbox_zetaBucket{
int64_t pix{};
uint32_t zmin{};
uint32_t zmax{};
int64_t draws{};
int64_t lastDraw{};
};
struct xbox_PixelEvent{
bool Drawn{};
bool ZReject{};
bool AlphaReject{};
uint8_t R{};
uint8_t G{};
uint8_t B{};
uint8_t A{};
};
struct xbox_rstats{
int64_t written{};
int64_t zRej{};
int64_t aRej{};
};
struct xbox_texKey{
uint32_t offset{};
uint32_t format{};
uint32_t rect{};
uint32_t ctl1{};
bool operator==(const xbox_texKey&)const=default;
auto operator<=>(const xbox_texKey&)const=default;
};
namespace std {template<>struct hash<xbox_texKey>{size_t operator()(const xbox_texKey&v)const{size_t h=0;h=h*31+std::hash<uint32_t>{}(v.offset);h=h*31+std::hash<uint32_t>{}(v.format);h=h*31+std::hash<uint32_t>{}(v.rect);h=h*31+std::hash<uint32_t>{}(v.ctl1);return h;}};}
struct xbox_texEntry{
xbox_texImage* img{};
uint64_t srcHash{};
uint32_t srcLen{};
uint64_t validated{};
};
struct xbox_texImage{
int64_t w{};
int64_t h{};
bool cube{};
Slice<uint8_t> pix{};
Slice<uint32_t> depth{};
};
struct xbox_shadowFragStats{
int64_t draw{};
int64_t near{};
int64_t far{};
int64_t behind{};
int64_t pxMin{};
int64_t pxMax{};
int64_t pyMin{};
int64_t pyMax{};
float rMin{};
float rMax{};
uint32_t dMin{};
uint32_t dMax{};
float diffMin{};
float diffMax{};
};
struct xbox_vtxFmt{
int64_t typ{};
int64_t size{};
uint32_t stride{};
int64_t words{};
};
struct xbox_vshState{
xbox_pgraph* g{};
std::array<std::array<float,4>,16>* v{};
std::array<std::array<float,4>,13> r{};
std::array<std::array<float,4>,13> o{};
int32_t a{};
};
using xbox_StopReason=int64_t;
using xbox_threadState=int64_t;
struct xbox_thread{
uint32_t id{};
uint32_t kthread{};
x86_CPU ctx{};
int32_t priority{};
xbox_threadState state{};
uint64_t wakeTick{};
int32_t suspendCount{};
bool waitAll{};
Slice<uint32_t> waitObjs{};
int64_t waitReg{};
uint32_t stackTop{};
uint32_t stackLimit{};
};
struct xbox_kobject{
std::string kind{};
uint32_t addr{};
bool signaled{};
int32_t count{};
int32_t limit{};
xbox_thread* thread{};
std::string lastSigWho{};
uint32_t lastSigFrom{};
uint64_t lastSigTick{};
};
using xbox_usbEndpointDir=int64_t;
using xbox_PadControlKind=int64_t;
struct xbox_PadControl{
xbox_PadControlKind Kind{};
uint16_t Bit{};
int64_t Index{};
int64_t Sign{};
};
struct xbox_PadState{
uint16_t Buttons{};
std::array<uint8_t,8> Analog{};
std::array<int16_t,4> Axes{};
};
struct xbox_xidDevice{
uint32_t Addr{};
uint32_t Config{};
uint16_t Buttons{};
std::array<uint8_t,8> Analog{};
std::array<int16_t,4> Axes{};
std::array<uint8_t,20> SentReport{};
bool Fresh{};
uint32_t AddrNext{};
bool AddrArmed{};
};
struct xbox_Section{
std::string Name{};
uint32_t Flags{};
uint32_t VAddr{};
uint32_t VSize{};
uint32_t RawAddr{};
uint32_t RawSize{};
uint32_t NameAddr{};
};
struct xbox_XBE{
uint32_t Base{};
uint32_t ImageSize{};
uint32_t Entry{};
uint32_t ThunkAddr{};
bool Retail{};
std::string TitleName{};
uint32_t TitleID{};
Slice<xbox_Section> Sections{};
Slice<uint16_t> Ordinals{};
Slice<uint8_t> raw{};
};
struct xbox_Image{
std::string Path{};
int64_t Size{};
int64_t Base{};
uint64_t CreationTime{};
uint32_t rootSector{};
uint32_t rootSize{};
os_File* f{};
std::string md5{};
Slice<xbox_Entry> index{};
};
struct Anon0{std::array<uint32_t,8> Regs{};std::array<uint16_t,8> Seg{};uint32_t IP{};uint32_t instrIP{};bool CF{};bool PF{};bool AF{};bool ZF{};bool SF{};bool TF{};bool IF{};bool DF{};bool OF{};bool Halted{};std::string HaltReason{};uint64_t Steps{};uint64_t Ext386{};int64_t Mode{};std::array<uint32_t,8> SegBase{};std::function<uint32_t(uint16_t)> SegResolve{};std::function<bool(x86_CPU*,uint8_t)> IntHook{};std::function<void(x86_CPU*)> OnStep{};std::function<uint32_t(uint16_t,int64_t)> PortIn{};std::function<void(uint16_t,int64_t,uint32_t)> PortOut{};RRX86Bus bus{};bool ssShadow{};x86_FPUState FPU{};std::array<std::array<uint8_t,16>,8> XMM{};std::array<std::array<uint8_t,8>,8> MMX{};uint32_t MXCSR{};uint64_t TSCMul{};std::function<uint64_t()> TSCFunc{};int64_t dSeg{};int64_t dOpsize{};int64_t dAddrsize{};};
struct Anon1{bool isReg{};uint8_t reg{};uint32_t base{};uint32_t off{};};
struct Anon10{uint64_t Due{};uint32_t IOSB{};uint32_t Event{};uint32_t Info{};uint32_t Status{};uint32_t Handle{};};
struct Anon11{Slice<uint8_t> RAM{};x86_CPU* CPU{};xbox_XBE* XBE{};xbox_Image* Disc{};uint32_t poolNext{};uint32_t heapNext{};uint32_t heapTop{};Map<uint32_t,xbox_kobject*> objects{};Map<uint32_t,xbox_fileObject*> files{};Map<std::string,xbox_cacheFile*> cacheFS{};Map<std::string,Slice<uint8_t>> fileBasic{};Slice<xbox_pendingIO> pendingIO{};Map<uint32_t,uint32_t> poolSizes{};uint32_t nextObjAddr{};uint32_t kbandNext{};Slice<xbox_thread*> threads{};xbox_thread* current{};uint32_t nextTID{};int64_t rrCursor{};bool reschedule{};int64_t quantumLeft{};Map<uint32_t,uint32_t> interrupts{};uint64_t nextVBlank{};bool isrActive{};x86_CPU isrSaved{};Slice<xbox_dpcEntry> dpcQueue{};uint64_t tick{};uint64_t clockBaseTick{};uint64_t clockBase100ns{};uint64_t tscBase{};uint32_t tickCountAddr{};uint32_t systemTimeAddr{};Map<uint32_t,Slice<uint8_t>> shaCtx{};Map<uint32_t,Slice<uint8_t>> rc4Ctx{};xbox_nv2a nv{};xbox_pusherState push{};xbox_pgraph* pgraph{};xbox_mmioLatch apu{};xbox_mmioLatch ac97{};xbox_mmioLatch usb{};xbox_mmioLatch nic{};uint64_t usbFrameServed{};uint32_t usbWrDword{};Slice<xbox_ktimer> timers{};std::array<xbox_xidDevice*,4> usbDev{};Slice<uint32_t> usbDone{};Slice<uint8_t> usbCtrlData{};int64_t usbCtrlOff{};uint32_t pciAddr{};Map<uint32_t,uint8_t> pciSpace{};Slice<std::string> Log{};Map<uint16_t,int64_t> OrdinalHits{};Map<uint16_t,bool> dataDeref{};Map<uint32_t,bool> dosErrWarned{};bool Halted{};std::string HaltReason{};bool firstPush{};bool pusherEnabled{};uint32_t wWLo{};uint32_t wWHi{};std::function<void(uint32_t,uint32_t,uint32_t)> onW{};uint32_t wRLo{};uint32_t wRHi{};std::function<void(uint32_t,uint32_t,uint32_t)> onR{};std::function<void(uint32_t,uint32_t,xbox_PixelEvent)> OnPixel{};std::function<void(xbox_Machine*,uint32_t,uint32_t,uint32_t)> OnNVMethod{};std::function<void(xbox_Machine*)> OnFlip{};bool FlipVSync{};bool StopRequested{};int64_t stopAfterMethod{};bool stopAfterArmed{};Map<uint32_t,bool> bps{};bool verbose{};int64_t traceLeft{};uint32_t bpAddr{};int64_t bpLeft{};Map<uint32_t,uint64_t> hotpc{};bool Profile{};xbox_profState prof{};};
struct Anon12{Map<uint32_t,uint32_t> reg{};uint32_t dmaPut{};uint32_t dmaGet{};bool kicked{};uint32_t pcrtcIntr{};uint32_t dispMode{};uint32_t dispFormat{};uint32_t fbPitch{};uint32_t fbAddr{};};
struct Anon13{uint32_t format{};uint32_t srcPitch{};uint32_t dstPitch{};uint32_t srcOffset{};uint32_t dstOffset{};};
struct Anon14{uint32_t inX{};uint32_t inY{};uint32_t outX{};uint32_t outY{};};
struct Anon15{std::array<std::array<float,4>,4> tex{};std::array<float,4> col0{};std::array<float,4> col1{};};
struct Anon16{uint32_t colorICW{};uint32_t alphaICW{};uint32_t colorOCW{};uint32_t alphaOCW{};std::array<float,4> factor0{};std::array<float,4> factor1{};};
struct Anon17{Slice<xbox_combStage> stages{};uint32_t cw0{};uint32_t cw1{};std::array<float,4> fogColor{};std::array<xbox_combStage,8> buf{};};
struct Anon18{std::array<std::array<float,4>,16> r{};};
struct Anon19{uint32_t Addr{};int64_t W{};int64_t H{};int64_t Stride{};std::string Format{};};
struct Anon2{std::array<double,8> St{};std::array<uint8_t,8> Tag{};int64_t Top{};uint16_t Ctrl{};uint16_t Stat{};};
struct Anon20{bool valid{};uint32_t base{};uint32_t pitch{};int64_t w{};int64_t h{};int64_t ax{};int64_t ay{};};
struct Anon21{uint32_t method{};uint32_t subchan{};uint32_t count{};bool nonInc{};uint32_t subReturn{};bool subActive{};bool running{};};
struct Anon22{xbox_Machine* m{};std::array<uint32_t,8> subObject{};std::array<uint32_t,8> subClass{};std::array<uint32_t,2048> Regs{};std::array<std::array<uint32_t,4>,136> Prog{};std::array<std::array<uint32_t,4>,192> Const{};uint32_t ProgLoad{};uint32_t ConstLoad{};std::array<uint32_t,4> progBuf{};int64_t progBufN{};std::array<uint32_t,4> constBuf{};int64_t constBufN{};uint32_t prim{};Slice<uint32_t> inline_{};Slice<uint32_t> elems{};Slice<std::array<uint32_t,2>> ranges{};std::array<std::array<float,4>,16> vtxAttr{};int64_t Draws{};int64_t pixWritten{};int64_t pixZRej{};int64_t pixARej{};int64_t lowWriteDraw{};std::string ffFragHalt{};xbox_presentedSurface presented{};xbox_rasterState rast{};bool rastValid{};Map<xbox_texKey,xbox_texEntry*> texCache{};uint64_t texRun{};xbox_surfaces2D surf2D{};xbox_imageBlit blit{};bool survey{};Map<uint32_t,int64_t> seen{};Map<uint32_t,uint32_t> firstArg{};int64_t Methods{};int64_t SetObjs{};Map<uint32_t,int64_t> unhandled{};Map<uint32_t,xbox_zetaBucket*> zetaHist{};bool shadowDumped{};xbox_shadowFragStats* shadowFrag{};Slice<xbox_vshInst> vshProg{};bool vshWritesConst{};Slice<std::array<std::array<float,4>,16>> vin{};Slice<xbox_kelvinVtx> vertsBuf{};Slice<std::array<int64_t,3>> triScratch{};Slice<std::array<int64_t,3>> clipTris{};Slice<xbox_kelvinVtx> clipVerts{};};
struct Anon23{int64_t pix{};uint32_t zmin{};uint32_t zmax{};int64_t draws{};int64_t lastDraw{};};
struct Anon24{uint32_t key{};int64_t count{};};
struct Anon25{bool Drawn{};bool ZReject{};bool AlphaReject{};uint8_t R{};uint8_t G{};uint8_t B{};uint8_t A{};};
struct Anon26{uint32_t colorPhys{};uint32_t zetaPhys{};uint32_t colorPitch{};uint32_t zetaPitch{};int64_t aaX{};int64_t aaY{};int64_t x0{};int64_t y0{};int64_t x1{};int64_t y1{};int64_t surfW{};int64_t surfH{};bool swizzle{};int64_t swW{};int64_t swH{};bool hasZeta{};bool depthTest{};bool depthWrite{};uint32_t depthFunc{};bool alphaTest{};uint32_t alphaFunc{};float alphaRef{};bool blend{};uint32_t blendSrc{};uint32_t blendDst{};uint32_t blendEq{};std::array<float,4> blendConst{};std::array<bool,4> colorMask{};bool cull{};uint32_t cullFace{};bool frontCW{};bool flatShade{};xbox_combState comb{};std::array<bool,4> texEnable{};std::array<xbox_texImage*,4> texImg{};std::array<uint32_t,4> texStage{};std::array<uint32_t,4> texWrapU{};std::array<uint32_t,4> texWrapV{};std::array<bool,4> texBilinear{};std::array<bool,4> texRect{};};
struct Anon27{int64_t written{};int64_t zRej{};int64_t aRej{};};
struct Anon28{uint32_t offset{};uint32_t format{};uint32_t rect{};uint32_t ctl1{};};
struct Anon29{xbox_texImage* img{};uint64_t srcHash{};uint32_t srcLen{};uint64_t validated{};};
struct Anon3{std::string name{};Map<uint32_t,uint32_t> reg{};Map<uint32_t,bool> seenCold{};};
struct Anon30{int64_t w{};int64_t h{};bool cube{};Slice<uint8_t> pix{};Slice<uint32_t> depth{};};
struct Anon31{int64_t draw{};int64_t near{};int64_t far{};int64_t behind{};int64_t pxMin{};int64_t pxMax{};int64_t pyMin{};int64_t pyMax{};float rMin{};float rMax{};uint32_t dMin{};uint32_t dMax{};float diffMin{};float diffMax{};};
struct Anon32{std::array<float,4> pos{};std::array<float,4> d0{};std::array<float,4> d1{};float fog{};std::array<std::array<float,4>,4> uv{};};
struct Anon33{int64_t typ{};int64_t size{};uint32_t stride{};int64_t words{};};
struct Anon34{int64_t mux{};int64_t reg{};bool neg{};std::array<int64_t,4> swz{};};
struct Anon35{int64_t mac{};int64_t ilu{};int64_t constIdx{};int64_t inputReg{};xbox_vshSrc a{};xbox_vshSrc b{};xbox_vshSrc c{};uint32_t macMask{};int64_t macDst{};uint32_t iluMask{};uint32_t outMask{};bool outIsO{};int64_t outAddr{};bool outFromILU{};bool relConst{};bool final{};};
struct Anon36{xbox_pgraph* g{};std::array<std::array<float,4>,16>* v{};std::array<std::array<float,4>,13> r{};std::array<std::array<float,4>,13> o{};int32_t a{};};
struct Anon37{std::string Name{};double Millis{};int64_t Count{};};
struct Anon38{std::string Name{};int64_t Value{};};
struct Anon39{double TotalMs{};Slice<xbox_ProfileBucket> Buckets{};Slice<xbox_ProfileCounter> Counters{};bool Drew{};};
struct Anon4{std::string name{};uint32_t vector{};std::function<bool(xbox_Machine*)> pending{};};
struct Anon40{int64_t draws{};int64_t methods{};int64_t frags{};int64_t zRej{};int64_t aRej{};};
struct Anon41{std::array<int64_t,4> ns{};std::array<int64_t,4> count{};time_Time runStart{};bool inRun{};int64_t frameNs{};time_Time pushStart{};bool inPush{};xbox_profCounters base{};uint64_t baseInstr{};xbox_FrameProfile last{};bool has{};};
struct Anon42{uint32_t id{};uint32_t kthread{};x86_CPU ctx{};int32_t priority{};xbox_threadState state{};uint64_t wakeTick{};int32_t suspendCount{};bool waitAll{};Slice<uint32_t> waitObjs{};int64_t waitReg{};uint32_t stackTop{};uint32_t stackLimit{};};
struct Anon43{std::string kind{};uint32_t addr{};bool signaled{};int32_t count{};int32_t limit{};xbox_thread* thread{};std::string lastSigWho{};uint32_t lastSigFrom{};uint64_t lastSigTick{};};
struct Anon44{uint32_t Timer{};uint32_t Dpc{};uint64_t Due{};uint32_t Period{};};
struct Anon45{xbox_PadControlKind Kind{};uint16_t Bit{};int64_t Index{};int64_t Sign{};};
struct Anon46{uint16_t Buttons{};std::array<uint8_t,8> Analog{};std::array<int16_t,4> Axes{};};
struct Anon47{uint32_t Addr{};uint32_t Config{};uint16_t Buttons{};std::array<uint8_t,8> Analog{};std::array<int16_t,4> Axes{};std::array<uint8_t,20> SentReport{};bool Fresh{};uint32_t AddrNext{};bool AddrArmed{};};
struct Anon48{std::string Name{};uint32_t Flags{};uint32_t VAddr{};uint32_t VSize{};uint32_t RawAddr{};uint32_t RawSize{};uint32_t NameAddr{};};
struct Anon49{uint32_t Base{};uint32_t ImageSize{};uint32_t Entry{};uint32_t ThunkAddr{};bool Retail{};std::string TitleName{};uint32_t TitleID{};Slice<xbox_Section> Sections{};Slice<uint16_t> Ordinals{};Slice<uint8_t> raw{};};
struct Anon5{uint32_t Dpc{};uint32_t Arg1{};uint32_t Arg2{};};
struct Anon50{std::string Path{};int64_t Size{};int64_t Base{};uint64_t CreationTime{};uint32_t rootSector{};uint32_t rootSize{};os_File* f{};std::string md5{};Slice<xbox_Entry> index{};};
struct Anon51{std::string Name{};std::string Path{};bool IsDir{};uint32_t Size{};uint32_t Sector{};uint8_t Attr{};};
struct Anon6{uint32_t pc{};uint64_t c{};};
struct Anon7{Slice<uint8_t> Data{};};
struct Anon8{xbox_Entry entry{};xbox_cacheFile* cache{};std::string key{};bool dir{};uint32_t off{};int64_t scan{};};
struct Anon9{std::string Name{};uint32_t Size{};bool IsDir{};};

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
xbox_mmioLatch xbox_newMMIOLatch(std::string name);
uint8_t xbox_Machine_latchRead(xbox_Machine* m,xbox_mmioLatch* l,uint32_t off,uint32_t dw,bool written);
void xbox_Machine_latchWrite(xbox_Machine* m,xbox_mmioLatch* l,uint32_t off,uint8_t v);
void xbox_Machine_latchTrace(xbox_Machine* m,xbox_mmioLatch* l,uint32_t off,uint32_t dw);
uint8_t xbox_Machine_apuRead(xbox_Machine* m,uint32_t off);
void xbox_Machine_apuWrite(xbox_Machine* m,uint32_t off,uint8_t v);
uint8_t xbox_Machine_ac97Read(xbox_Machine* m,uint32_t off);
void xbox_Machine_ac97Write(xbox_Machine* m,uint32_t off,uint8_t v);
uint8_t xbox_Machine_nicRead(xbox_Machine* m,uint32_t off);
void xbox_Machine_nicWrite(xbox_Machine* m,uint32_t off,uint8_t v);
void xbox_Machine_apuTick(xbox_Machine* m);
void xbox_Machine_creditFlipVBlank(xbox_Machine* m);
void xbox_Machine_vblankTick(xbox_Machine* m);
bool xbox_Machine_irqPending(xbox_Machine* m);
void xbox_Machine_deliverPending(xbox_Machine* m);
void xbox_Machine_isrReturn(xbox_Machine* m);
uint32_t xbox_Machine_DebugInterruptKI(xbox_Machine* m,uint32_t vector);
void xbox_Machine_setupKPCR(xbox_Machine* m);
void xbox_Machine_patchThunks(xbox_Machine* m);
void xbox_Machine_onStep(xbox_Machine* m,x86_CPU* c);
void xbox_Machine_EnableHotPC(xbox_Machine* m);
Slice<std::string> xbox_Machine_HotPCReport(xbox_Machine* m,int64_t n);
void xbox_Machine_dispatchKernel(xbox_Machine* m,uint16_t ord);
uint32_t xbox_Machine_retAddr(xbox_Machine* m);
uint32_t xbox_Machine_arg(xbox_Machine* m,int64_t i);
void xbox_Machine_setRet(xbox_Machine* m,uint32_t v);
void xbox_Machine_kret(xbox_Machine* m,int64_t argWords);
uint32_t xbox_Machine_allocPool(xbox_Machine* m,uint32_t size);
uint32_t xbox_Machine_allocPoolAligned(xbox_Machine* m,uint32_t size,uint32_t align);
uint32_t xbox_Machine_allocVirtual(xbox_Machine* m,uint32_t size);
uint32_t xbox_Machine_allocKObject(xbox_Machine* m,uint32_t size);
std::function<int64_t(xbox_Machine*)> xbox_kernelHandler(uint16_t ord);
Slice<uint8_t> xbox_Machine_readBytes(xbox_Machine* m,uint32_t addr,uint32_t n);
void xbox_Machine_writeBytes(xbox_Machine* m,uint32_t addr,Slice<uint8_t> b);
std::function<int64_t(xbox_Machine*)> xbox_kernelCryptoHandler(uint16_t ord);
uint8_t xbox_popcount7(uint8_t b);
std::tuple<int64_t,bool> xbox_dataExportSize(uint16_t ord);
void xbox_Machine_initDataExport(xbox_Machine* m,uint16_t ord,uint32_t addr);
bool xbox_fileObject_isDir(xbox_fileObject* fo);
uint64_t xbox_fnv64(std::string s);
uint32_t xbox_fileObject_size(xbox_fileObject* fo);
std::tuple<std::string,bool> xbox_resolveDiscPath(std::string p);
std::tuple<std::string,bool> xbox_resolveCachePath(std::string p);
std::tuple<std::string,bool> xbox_rawDevicePath(std::string p);
xbox_cacheFile* xbox_Machine_rawDevice(xbox_Machine* m,std::string key);
std::tuple<uint32_t,bool,bool> xbox_Machine_statPath(xbox_Machine* m,std::string path);
Slice<xbox_dirEntry> xbox_Machine_listDir(xbox_Machine* m,xbox_fileObject* fo);
bool xbox_matchPattern(std::string pat,std::string name);
std::tuple<uint64_t,uint64_t> xbox_Machine_volumeUnits(xbox_Machine* m,xbox_fileObject* fo);
void xbox_Machine_writeNetworkOpenInfo(xbox_Machine* m,uint32_t buf,uint32_t size,bool isDir);
Map<std::string,int64_t> xbox_Machine_DebugCacheFS(xbox_Machine* m);
Slice<uint8_t> xbox_Machine_DebugCacheFile(xbox_Machine* m,std::string key);
std::string xbox_Machine_readObjectAttributesPath(xbox_Machine* m,uint32_t oa);
std::string xbox_Machine_readObjectString(xbox_Machine* m,uint32_t p);
uint32_t xbox_Machine_openFile(xbox_Machine* m,uint32_t handleOut,uint32_t oa,uint32_t iosb,uint32_t disposition);
uint32_t xbox_Machine_finishOpen(xbox_Machine* m,uint32_t iosb,uint32_t _handle,uint32_t info,uint32_t status);
uint64_t xbox_ioCompletionTicks(uint32_t n);
uint32_t xbox_Machine_readFile(xbox_Machine* m,uint32_t handle,uint32_t event,uint32_t apcRoutine,uint32_t apcCtx,uint32_t iosb,uint32_t buffer,uint32_t length,uint32_t byteOffsetPtr);
uint32_t xbox_Machine_writeFile(xbox_Machine* m,uint32_t handle,uint32_t event,uint32_t iosb,uint32_t buffer,uint32_t length,uint32_t byteOffsetPtr);
uint32_t xbox_Machine_setFileLength(xbox_Machine* m,xbox_fileObject* fo,uint32_t n);
void xbox_Machine_ioTick(xbox_Machine* m);
uint32_t xbox_Machine_newObject(xbox_Machine* m,std::string kind,uint8_t typ,bool signaled,int32_t count,int32_t limit);
void xbox_Machine_writeSignal(xbox_Machine* m,uint32_t addr,bool signaled);
xbox_kobject* xbox_Machine_objAt(xbox_Machine* m,uint32_t handle);
xbox_kobject* xbox_Machine_guestObjAt(xbox_Machine* m,uint32_t addr);
bool xbox_Machine_satisfyWait(xbox_Machine* m,xbox_kobject* o);
std::function<int64_t(xbox_Machine*)> xbox_kernelObjectHandler(uint16_t ord);
void xbox_Machine_doWait(xbox_Machine* m,uint32_t handle,int64_t reg);
void xbox_Machine_doWaitTimed(xbox_Machine* m,uint32_t handle,uint32_t timeoutPtr);
void xbox_Machine_wakeWaiters(xbox_Machine* m,uint32_t handle);
uint32_t xbox_boolU32(bool b);
std::string xbox_ordinalName(uint16_t ord);
std::tuple<xbox_Machine*,Error> xbox_NewMachine(xbox_XBE* xbe,xbox_Image* disc);
uint32_t xbox_Machine_resolveSel(xbox_Machine* m,uint16_t sel);
Error xbox_Machine_loadImage(xbox_Machine* m);
void xbox_Machine_setupMemoryLayout(xbox_Machine* m);
std::tuple<uint32_t,bool,bool> xbox_Machine_translate(xbox_Machine* m,uint32_t a);
uint8_t xbox_Machine_Read(xbox_Machine* m,uint32_t a);
void xbox_Machine_Write(xbox_Machine* m,uint32_t a,uint8_t v);
uint32_t xbox_Machine_watchPC(xbox_Machine* m);
void xbox_Machine_SetWriteWatch(xbox_Machine* m,uint32_t lo,uint32_t hi,std::function<void(uint32_t,uint32_t,uint32_t)> cb);
void xbox_Machine_SetReadWatch(xbox_Machine* m,uint32_t lo,uint32_t hi,std::function<void(uint32_t,uint32_t,uint32_t)> cb);
void xbox_Machine_fault(xbox_Machine* m,std::string kind,uint32_t a);
uint32_t xbox_Machine_read32(xbox_Machine* m,uint32_t a);
uint16_t xbox_Machine_read16(xbox_Machine* m,uint32_t a);
void xbox_Machine_write32(xbox_Machine* m,uint32_t a,uint32_t v);
void xbox_Machine_write16(xbox_Machine* m,uint32_t a,uint16_t v);
std::string xbox_Machine_cstr(xbox_Machine* m,uint32_t a);
void xbox_Machine_SetVerbose(xbox_Machine* m,bool v);
void xbox_Machine_EnableGPU(xbox_Machine* m);
xbox_pgraph* xbox_Machine_PGraph(xbox_Machine* m);
void xbox_Machine_SetTrace(xbox_Machine* m,int64_t n);
void xbox_Machine_SetPCBreak(xbox_Machine* m,uint32_t addr,int64_t n);
uint8_t xbox_Machine_MemReadByte(xbox_Machine* m,uint32_t a);
uint32_t xbox_Machine_MemRead32(xbox_Machine* m,uint32_t a);
void xbox_Machine_Poke(xbox_Machine* m,uint32_t a,uint32_t v);
void xbox_Machine_ReadRAM(xbox_Machine* m,uint32_t a,Slice<uint8_t> buf);
void xbox_Machine_ReadCode(xbox_Machine* m,uint32_t a,Slice<uint8_t> buf);
uint32_t xbox_Machine_NVSubchannelClass(xbox_Machine* m,uint32_t subchan);
std::tuple<uint32_t,uint32_t,uint32_t> xbox_Machine_AllocStats(xbox_Machine* m);
uint32_t xbox_Machine_CallerReturnAddr(xbox_Machine* m);
uint32_t xbox_align32(uint32_t v,uint32_t a);
std::string xbox_nvRegName(uint32_t off);
uint8_t xbox_Machine_nvRead(xbox_Machine* m,uint32_t off);
void xbox_Machine_nvWrite(xbox_Machine* m,uint32_t off,uint8_t v);
bool xbox_Machine_FirstPushReached(xbox_Machine* m);
void xbox_pgraph_surf2DMethod(xbox_pgraph* g,uint32_t method,uint32_t arg);
void xbox_pgraph_blitMethod(xbox_pgraph* g,uint32_t method,uint32_t arg);
void xbox_pgraph_doBlit(xbox_pgraph* g,uint32_t w,uint32_t h);
uint32_t xbox_surf2DBpp(uint32_t format);
std::tuple<Slice<xbox_kelvinVtx>,Slice<std::array<int64_t,3>>> xbox_pgraph_clipNearPlane(xbox_pgraph* g,Slice<xbox_kelvinVtx> verts,Slice<std::array<int64_t,3>> tris);
xbox_kelvinVtx xbox_clipVertex(xbox_kelvinVtx* p,xbox_kelvinVtx* q,std::array<float,4> cp,std::array<float,4> cq,float t,std::function<std::array<float,4>(std::array<float,4>)> project);
void xbox_pgraph_combDecode(xbox_pgraph* g,xbox_combState* cs);
std::array<float,4> xbox_argb8Norm(uint32_t v);
std::array<float,4> xbox_argb8Vec(uint32_t v);
std::array<float,4> xbox_combRegs_read(xbox_combRegs* r,uint32_t reg);
void xbox_combRegs_write(xbox_combRegs* r,uint32_t reg,std::array<float,4> v,bool alpha);
void xbox_combMap(std::array<float,4>* v,uint32_t mapping);
void xbox_combMap_reference(std::array<float,4>* v,uint32_t mapping);
std::array<float,4> xbox_combFetch(xbox_combRegs* r,uint32_t b,bool alphaSide);
std::array<float,4> xbox_pgraph_combine(xbox_pgraph* g,xbox_combState* cs,xbox_combInput* in);
std::array<float,4> xbox_finalFetch(xbox_combRegs* r,uint32_t b,std::array<float,4> ef);
void xbox_combOp(std::array<float,4>* v,uint32_t op);
void xbox_combOp_reference(std::array<float,4>* v,uint32_t op);
float xbox_clampSigned(float x);
std::array<float,4> xbox_mul4(std::array<float,4> a,std::array<float,4> b);
std::array<float,4> xbox_add4(std::array<float,4> a,std::array<float,4> b);
float xbox_clamp01(float x);
std::string xbox_NVMethodName(uint32_t class_,uint32_t method);
std::string xbox_NVMethodDecode(uint32_t class_,uint32_t method,uint32_t arg);
std::string xbox_nvPrimName(uint32_t p);
std::string xbox_nvTexFormatName(uint32_t code);
std::tuple<int64_t,int64_t> xbox_Machine_SurfaceAAScale(xbox_Machine* m);
bool xbox_Machine_TextureBound(xbox_Machine* m,int64_t u);
std::tuple<image_RGBA*,Error> xbox_Machine_DumpTexture(xbox_Machine* m,int64_t u);
Slice<std::string> xbox_RegionFormats();
std::tuple<image_RGBA*,Error> xbox_Machine_RenderRegion(xbox_Machine* m,xbox_RegionSpec s);
std::tuple<int64_t,bool> xbox_regionBPP(std::string format);
std::tuple<uint8_t,uint8_t,uint8_t,uint8_t> xbox_decodeRegionTexel(std::string format,Slice<uint8_t> b);
std::tuple<Slice<std::string>,bool> xbox_Machine_VshProgramDisasm(xbox_Machine* m);
std::array<float,4> xbox_Machine_VshConst(xbox_Machine* m,int64_t c);
void xbox_pgraph_clearSurface(xbox_pgraph* g,uint32_t mask);
void xbox_pgraph_recordPresented(xbox_pgraph* g);
std::tuple<image_RGBA*,Error> xbox_Machine_RenderPresented(xbox_Machine* m);
std::tuple<image_RGBA*,Error> xbox_Machine_RenderScanout(xbox_Machine* m);
std::tuple<image_RGBA*,Error> xbox_Machine_RenderDrawTarget(xbox_Machine* m);
std::tuple<image_RGBA*,Error> xbox_Machine_renderSwizzledSurface(xbox_Machine* m,uint32_t base,int64_t sw,int64_t sh,int64_t w,int64_t h);
std::tuple<image_RGBA*,Error> xbox_Machine_renderRawSurface(xbox_Machine* m,uint32_t base,uint32_t pitch,int64_t w,int64_t h,int64_t ax,int64_t ay);
void xbox_pgraph_kelvinMethod(xbox_pgraph* g,uint32_t method,uint32_t arg);
void xbox_Machine_runPusher(xbox_Machine* m);
xbox_pgraph* xbox_newPgraph(xbox_Machine* m);
void xbox_pgraph_SetSurvey(xbox_pgraph* g,bool v);
void xbox_Machine_pgraphMethod(xbox_Machine* m,uint32_t subchan,uint32_t method,uint32_t arg);
void xbox_pgraph_DumpZetaHist(xbox_pgraph* g);
Slice<std::string> xbox_pgraph_SurveyReport(xbox_pgraph* g);
uint32_t xbox_Machine_ramhtInstance(xbox_Machine* m,uint32_t handle);
uint32_t xbox_Machine_ramhtClass(xbox_Machine* m,uint32_t handle);
std::tuple<uint32_t,uint32_t> xbox_Machine_dmaObjectTarget(xbox_Machine* m,uint32_t handle);
std::tuple<int64_t,int64_t> xbox_surfaceAAScale(uint32_t format);
std::tuple<int64_t,int64_t,std::string> xbox_swizzleGeom(uint32_t format);
bool xbox_pgraph_decodeSwizzleExtent(xbox_pgraph* g,xbox_rasterState* st,uint32_t format);
bool xbox_pgraph_rasterStateDecode(xbox_pgraph* g,xbox_rasterState* st);
std::tuple<uint32_t,bool> xbox_pgraph_surfPhys(xbox_pgraph* g,uint32_t addr);
uint32_t xbox_rasterState_colorAddr(xbox_rasterState* st,int64_t px,int64_t py);
uint32_t xbox_rasterState_zetaAddr(xbox_rasterState* st,int64_t px,int64_t py);
void xbox_pgraph_rasterTri(xbox_pgraph* g,xbox_kelvinVtx* v0,xbox_kelvinVtx* v1,xbox_kelvinVtx* v2,xbox_rstats* rs);
void xbox_pgraph_mergeStats(xbox_pgraph* g,xbox_rstats* rs);
bool xbox_pgraph_rasterParallelOK(xbox_pgraph* g,xbox_rasterState* st,Slice<xbox_kelvinVtx> verts,Slice<std::array<int64_t,3>> tris);
void xbox_pgraph_rasterTriBand(xbox_pgraph* g,xbox_kelvinVtx* v0,xbox_kelvinVtx* v1,xbox_kelvinVtx* v2,int64_t lane,int64_t stride,xbox_rstats* rs);
void xbox_pgraph_shadePixel(xbox_pgraph* g,xbox_rasterState* st,int64_t px,int64_t py,float b0,float b1,float b2,float iw0,float iw1,float iw2,xbox_kelvinVtx* v0,xbox_kelvinVtx* v1,xbox_kelvinVtx* v2,xbox_rstats* rs);
bool xbox_depthPass(uint32_t fn,uint32_t z,uint32_t old);
bool xbox_alphaPass(uint32_t fn,float a,float ref);
std::array<float,4> xbox_blendPixel(xbox_rasterState* st,std::array<float,4> src,std::array<float,4> dst);
uint8_t xbox_u8(float f);
float xbox_invW(float w);
double xbox_min3(double a,double b,double c);
double xbox_max3(double a,double b,double c);
std::array<float,4> xbox_pgraph_texReflectSpecular(xbox_pgraph* g,xbox_rasterState* st,int64_t u,std::array<std::array<float,4>,4>* uvw,xbox_combInput* in);
uint64_t xbox_hashRAM(Slice<uint8_t> ram,uint32_t phys,uint32_t n);
std::tuple<uint32_t,bool> xbox_texSpan(uint32_t colorFmt,bool cube,int64_t w,int64_t h,uint32_t pitch);
std::tuple<uint32_t,uint32_t,bool> xbox_pgraph_texSource(xbox_pgraph* g,int64_t u);
void xbox_pgraph_cacheTex(xbox_pgraph* g,xbox_texKey key,xbox_texImage* img,int64_t u);
bool xbox_pgraph_texStateDecode(xbox_pgraph* g,xbox_rasterState* st);
std::tuple<xbox_texImage*,bool,bool> xbox_pgraph_texDecode(xbox_pgraph* g,int64_t u);
std::tuple<int64_t,int64_t,Slice<uint8_t>,bool> xbox_pgraph_DebugDecodeTexture(xbox_pgraph* g,int64_t u);
bool xbox_isLinearTexFmt(uint32_t colorFmt);
uint8_t xbox_exp5(uint16_t v);
uint8_t xbox_exp6(uint16_t v);
void xbox_decode565(Slice<uint8_t> pix,int64_t o,Slice<uint8_t> b);
void xbox_decode1555(Slice<uint8_t> pix,int64_t o,Slice<uint8_t> b);
void xbox_decode4444(Slice<uint8_t> pix,int64_t o,Slice<uint8_t> b);
void xbox_decodeSwizzled(xbox_texImage* img,Slice<uint8_t> ram,uint32_t phys,int64_t bpp,std::function<void(Slice<uint8_t>,int64_t,Slice<uint8_t>)> put);
int64_t xbox_swizzleOffset(int64_t x,int64_t y,int64_t w,int64_t h);
void xbox_decodeLinear(xbox_texImage* img,Slice<uint8_t> ram,uint32_t phys,uint32_t pitch,int64_t bpp,std::function<void(Slice<uint8_t>,int64_t,Slice<uint8_t>)> put);
void xbox_decodeDXT(xbox_texImage* img,Slice<uint8_t> ram,uint32_t phys,int64_t variant);
uint8_t xbox_dxt5Alpha(Slice<uint8_t> blk,int64_t i);
float xbox_texWrapCoord(float v,int64_t size,uint32_t mode);
std::array<float,4> xbox_pgraph_texSampleCube(xbox_pgraph* g,xbox_rasterState* st,int64_t u,float x,float y,float z);
float xbox_absf32(float v);
void xbox_pgraph_traceShadowFrag(xbox_pgraph* g,xbox_rasterState* st,int64_t u,int64_t px,int64_t py,float s,float t,float r,float q);
void xbox_shadowFragStats_print(xbox_shadowFragStats* f);
bool xbox_shadowComparePass(float r,uint32_t d);
std::array<float,4> xbox_pgraph_texSampleShadow(xbox_pgraph* g,xbox_rasterState* st,int64_t u,float s,float t,float r);
std::array<float,4> xbox_pgraph_texSample(xbox_pgraph* g,xbox_rasterState* st,int64_t u,float s,float t);
xbox_vtxFmt xbox_pgraph_attrFormat(xbox_pgraph* g,int64_t i);
std::array<float,4> xbox_decodeAttr(xbox_vtxFmt* vf,Slice<uint32_t> w);
void xbox_pgraph_beginEnd(xbox_pgraph* g,uint32_t arg);
void xbox_pgraph_runDraw(xbox_pgraph* g);
Slice<std::array<std::array<float,4>,16>> xbox_ensureVin(Slice<std::array<std::array<float,4>,16>> buf,int64_t n);
std::tuple<Slice<xbox_kelvinVtx>,bool> xbox_pgraph_transformVerts(xbox_pgraph* g,int64_t nin,bool ff,bool trace);
std::tuple<std::array<std::array<float,4>,16>,bool> xbox_pgraph_fetchArrayVertex(xbox_pgraph* g,std::array<xbox_vtxFmt,16>* fmts,uint32_t ix);
xbox_kelvinVtx xbox_pgraph_transformFF(xbox_pgraph* g,std::array<std::array<float,4>,16>* in);
std::tuple<xbox_kelvinVtx,bool> xbox_pgraph_transform(xbox_pgraph* g,std::array<std::array<float,4>,16>* in,bool trace);
void xbox_pgraph_assemble(xbox_pgraph* g,Slice<xbox_kelvinVtx> verts);
void xbox_pgraph_traceDraw(xbox_pgraph* g,std::array<xbox_vtxFmt,16>* fmts,int64_t stride);
uint32_t xbox_vshField(std::array<uint32_t,4>* w,int64_t dword,int64_t lo,int64_t n);
xbox_vshInst xbox_pgraph_vshDecode(xbox_pgraph* g,int64_t i);
std::array<float,4> xbox_vshState_src(xbox_vshState* s,xbox_vshInst* inst,xbox_vshSrc* o);
std::array<float,4> xbox_f32vec(std::array<uint32_t,4>* c);
void xbox_maskWrite(std::array<float,4>* dst,std::array<float,4> val,uint32_t mask);
bool xbox_pgraph_vshCompile(xbox_pgraph* g);
bool xbox_pgraph_vshRun(xbox_pgraph* g,std::array<std::array<float,4>,16>* v,std::array<std::array<float,4>,13>* out);
void xbox_maskWrite32(std::array<uint32_t,4>* dst,std::array<uint32_t,4> val,uint32_t mask);
void xbox_pgraph_progData(xbox_pgraph* g,uint32_t arg);
void xbox_pgraph_constData(xbox_pgraph* g,uint32_t arg);
std::string xbox_pgraph_vshDisasm(xbox_pgraph* g,int64_t i);
std::string xbox_outName(bool isO,int64_t addr);
float xbox_minf32(float a,float b);
float xbox_maxf32(float a,float b);
float xbox_floorf32(float f);
float xbox_exp2f32(float f);
float xbox_rcpf32(float x);
float xbox_rccf32(float x);
float xbox_rsqf32(float x);
std::array<float,4> xbox_log2partial(float x);
std::array<float,4> xbox_litf32(std::array<float,4> c);
uint32_t xbox_Machine_portIn(xbox_Machine* m,uint16_t port,int64_t size);
void xbox_Machine_portOut(xbox_Machine* m,uint16_t port,int64_t size,uint32_t v);
uint32_t xbox_Machine_pciConfigRead(xbox_Machine* m,int64_t size);
xbox_profCounters xbox_Machine_profCounters(xbox_Machine* m);
time_Time xbox_Machine_profStart(xbox_Machine* m);
void xbox_Machine_profEnd(xbox_Machine* m,int64_t bucket,time_Time t);
void xbox_Machine_profPusherEnter(xbox_Machine* m);
void xbox_Machine_profPusherExit(xbox_Machine* m);
void xbox_Machine_profRunEnter(xbox_Machine* m);
void xbox_Machine_profRunExit(xbox_Machine* m);
void xbox_Machine_profFrame(xbox_Machine* m);
xbox_FrameProfile xbox_Machine_FrameProfile(xbox_Machine* m);
void xbox_Machine_SetProfile(xbox_Machine* m,bool on);
std::string xbox_StopReason_String(xbox_StopReason r);
std::tuple<xbox_StopReason,uint64_t> xbox_Machine_Run(xbox_Machine* m,uint64_t maxSteps);
std::tuple<xbox_StopReason,uint64_t> xbox_Machine_RunStopAfterNVMethod(xbox_Machine* m,int64_t k,uint64_t maxSteps);
uint32_t xbox_Machine_PC(xbox_Machine* m);
void xbox_Machine_SetBreakpoint(xbox_Machine* m,uint32_t pc);
void xbox_Machine_ClearBreakpoint(xbox_Machine* m,uint32_t pc);
void xbox_Machine_ClearBreakpoints(xbox_Machine* m);
Slice<uint32_t> xbox_Machine_Breakpoints(xbox_Machine* m);
void xbox_Machine_ClearHalt(xbox_Machine* m);
std::string xbox_Machine_Report(xbox_Machine* m);
Slice<std::string> xbox_Machine_OrdinalHistogram(xbox_Machine* m);
uint64_t xbox_Machine_systemTime100ns(xbox_Machine* m);
uint64_t xbox_Machine_guestMs(xbox_Machine* m);
uint64_t xbox_Machine_guestTSC(xbox_Machine* m);
void xbox_Machine_schedTick(xbox_Machine* m);
void xbox_Machine_yieldCurrent(xbox_Machine* m,xbox_threadState state);
void xbox_Machine_dispatch(xbox_Machine* m);
bool xbox_Machine_idleAdvance(xbox_Machine* m);
xbox_thread* xbox_Machine_pickRunnable(xbox_Machine* m);
void xbox_Machine_switchTo(xbox_Machine* m,xbox_thread* t);
void xbox_Machine_wakeDueSleepers(xbox_Machine* m);
int64_t xbox_Machine_aliveThreads(xbox_Machine* m);
std::string xbox_threadState_String(xbox_threadState s);
bool xbox_thread_runnable(xbox_thread* t);
void xbox_kobject_noteSignal(xbox_kobject* o,std::string who,uint32_t from,uint64_t tick);
Slice<std::string> xbox_Machine_DebugThreads(xbox_Machine* m);
int64_t xbox_Machine_DebugCurrentTid(xbox_Machine* m);
void xbox_Machine_bootThread(xbox_Machine* m);
uint32_t xbox_Machine_currentKThread(xbox_Machine* m);
xbox_thread* xbox_Machine_createThread(xbox_Machine* m,uint32_t entry,uint32_t ctx1,uint32_t ctx2,uint32_t stackSize,uint32_t tlsSize,int32_t priority,bool suspended);
void xbox_Machine_initTLSArea(xbox_Machine* m,uint32_t addr,uint32_t size);
void xbox_Machine_pushCtx(xbox_Machine* m,x86_CPU* c,uint32_t v);
void xbox_Machine_exitCurrentThread(xbox_Machine* m);
uint32_t xbox_Machine_threadID(xbox_Machine* m);
uint32_t xbox_b2u(bool b);
bool xbox_Machine_armTimer(xbox_Machine* m,uint32_t tm,uint32_t dpc,uint32_t dueLo,uint32_t dueHi,uint32_t period);
bool xbox_Machine_cancelTimer(xbox_Machine* m,uint32_t tm);
bool xbox_Machine_queueDPC(xbox_Machine* m,uint32_t dpc,uint32_t arg1,uint32_t arg2);
void xbox_Machine_timerTick(xbox_Machine* m);
void xbox_Machine_deliverDPC(xbox_Machine* m);
uint64_t xbox_Machine_usbFrame(xbox_Machine* m);
uint8_t xbox_Machine_usbRead(xbox_Machine* m,uint32_t off);
void xbox_Machine_usbWrite(xbox_Machine* m,uint32_t off,uint8_t v);
void xbox_Machine_usbPortWrite(xbox_Machine* m,uint32_t port,uint32_t bits);
void xbox_Machine_usbSetPortConnected(xbox_Machine* m,uint32_t port,bool connected);
void xbox_Machine_usbRaise(xbox_Machine* m,uint32_t bit);
bool xbox_Machine_usbIRQ(xbox_Machine* m);
void xbox_Machine_usbUpdateIRQ(xbox_Machine* m);
void xbox_Machine_usbTick(xbox_Machine* m);
void xbox_Machine_usbFrameTick(xbox_Machine* m);
void xbox_Machine_usbWalkPeriodic(xbox_Machine* m);
xbox_xidDevice* xbox_Machine_usbDeviceFor(xbox_Machine* m,uint32_t fa);
void xbox_Machine_usbWalkList(xbox_Machine* m,uint32_t head);
void xbox_Machine_usbRunEndpoint(xbox_Machine* m,uint32_t ed);
std::tuple<uint32_t,bool> xbox_Machine_usbRunTD(xbox_Machine* m,uint32_t ed,uint32_t edCtrl,uint32_t td);
void xbox_Machine_usbControlStatus(xbox_Machine* m,xbox_xidDevice* dev,uint32_t endpoint,uint32_t length);
void xbox_Machine_usbRetire(xbox_Machine* m,uint32_t td,uint32_t ctrl,uint32_t cbp,uint32_t moved,uint32_t length,uint32_t cc);
uint32_t xbox_usbCurrentBuffer(uint32_t cbp,uint32_t moved,uint32_t length);
void xbox_Machine_usbStall(xbox_Machine* m,uint32_t ed,uint32_t td,uint32_t ctrl,uint32_t cbp,uint32_t length,uint32_t nextTD);
void xbox_Machine_usbHaltEndpoint(xbox_Machine* m,uint32_t ed,uint32_t nextTD);
void xbox_Machine_usbWriteback(xbox_Machine* m);
void xbox_Machine_usbFrameNumberToHCCA(xbox_Machine* m);
std::tuple<xbox_PadControl,bool> xbox_PadControlByName(std::string name);
Slice<std::string> xbox_PadControlNames();
xbox_PadState xbox_PadStateOf(Map<std::string,bool> held);
void xbox_Machine_AttachPad(xbox_Machine* m,int64_t port);
void xbox_Machine_SetPadButtons(xbox_Machine* m,int64_t port,uint16_t buttons);
xbox_xidDevice* xbox_Machine_pad(xbox_Machine* m,int64_t port);
void xbox_Machine_SetPadAnalog(xbox_Machine* m,int64_t port,int64_t i,uint8_t pressure);
uint8_t xbox_Machine_PadAnalog(xbox_Machine* m,int64_t port,int64_t i);
void xbox_Machine_SetPadAxis(xbox_Machine* m,int64_t port,int64_t i,int16_t v);
int16_t xbox_Machine_PadAxis(xbox_Machine* m,int64_t port,int64_t i);
void xbox_Machine_SetPad(xbox_Machine* m,int64_t port,xbox_PadState s);
uint16_t xbox_Machine_PadButtons(xbox_Machine* m,int64_t port);
uint32_t xbox_xidDevice_address(xbox_xidDevice* d);
void xbox_xidDevice_controlStatusDone(xbox_xidDevice* d,xbox_Machine* m);
std::tuple<Slice<uint8_t>,Error> xbox_xidDevice_setup(xbox_xidDevice* d,xbox_Machine* m,Slice<uint8_t> pkt);
std::tuple<Slice<uint8_t>,Error> xbox_xidDevice_descriptor(xbox_xidDevice* d,xbox_Machine* m,uint16_t dtype,uint8_t index);
Slice<uint8_t> xbox_xidDevice_interruptIn(xbox_xidDevice* d,xbox_Machine* m,uint32_t endpoint);
std::array<uint8_t,20> xbox_xidDevice_report(xbox_xidDevice* d);
std::string xbox_Section_FlagString(xbox_Section s);
std::tuple<xbox_XBE*,Error> xbox_ParseXBE(Slice<uint8_t> b);
bool xbox_XBE_inImage(xbox_XBE* x,uint32_t va);
std::tuple<int64_t,bool> xbox_XBE_atVA(xbox_XBE* x,uint32_t va);
Error xbox_XBE_parseSections(xbox_XBE* x,uint32_t n,uint32_t addr);
std::string xbox_XBE_cstrVA(xbox_XBE* x,uint32_t va);
void xbox_XBE_parseCertificate(xbox_XBE* x,uint32_t va);
Error xbox_XBE_parseThunks(xbox_XBE* x);
uint16_t xbox_le16(Slice<uint8_t> b);
uint32_t xbox_le32(Slice<uint8_t> b);
std::string xbox_Entry_String(xbox_Entry e);
int64_t xbox_Entry_Offset(xbox_Entry e);
std::tuple<xbox_Entry,int64_t,bool> xbox_Image_EntryAt(xbox_Image* img,int64_t off);
std::tuple<xbox_Image*,Error> xbox_Open(std::string path);
std::tuple<int64_t,Slice<uint8_t>,Error> xbox_Image_findVolumeDescriptor(xbox_Image* img);
Error xbox_Image_Close(xbox_Image* img);
std::tuple<Slice<uint8_t>,Error> xbox_Image_readAt(xbox_Image* img,int64_t off,int64_t n);
std::tuple<Slice<uint8_t>,Error> xbox_Image_Read(xbox_Image* img,int64_t off,int64_t n);
std::tuple<Slice<xbox_Entry>,Error> xbox_Image_ReadDir(xbox_Image* img,std::string path);
Error xbox_Image_Walk(xbox_Image* img,std::function<Error(xbox_Entry)> fn);
Error xbox_Image_walk(xbox_Image* img,uint32_t sector,uint32_t size,std::string dirPath,std::function<Error(xbox_Entry)> fn);
std::tuple<Slice<uint8_t>,Error> xbox_Image_ReadFile(xbox_Image* img,std::string path);
std::tuple<Slice<uint8_t>,Error> xbox_Image_ReadFileEntry(xbox_Image* img,xbox_Entry e);
std::tuple<Slice<xbox_Entry>,Error> xbox_Image_dirEntries(xbox_Image* img,uint32_t sector,uint32_t size,std::string dirPath);
std::tuple<xbox_Entry,Error> xbox_Image_resolve(xbox_Image* img,std::string path);
Slice<std::string> xbox_splitPath(std::string p);
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
constexpr int64_t xbox_apuBase=4269801472ULL;
constexpr int64_t xbox_apuSize=524288ULL;
constexpr int64_t xbox_apuTop=4270325760ULL;
constexpr int64_t xbox_ac97Base=4273995776ULL;
constexpr int64_t xbox_ac97Size=4096ULL;
constexpr int64_t xbox_ac97Top=4273999872ULL;
constexpr int64_t xbox_usbBase=4275044352ULL;
constexpr int64_t xbox_usbSize=4096ULL;
constexpr int64_t xbox_usbTop=4275048448ULL;
constexpr int64_t xbox_nicBase=4277141504ULL;
constexpr int64_t xbox_nicSize=4096ULL;
constexpr int64_t xbox_nicTop=4277145600ULL;
constexpr int64_t xbox_apuHandshake=131088ULL;
constexpr int64_t xbox_apuCounterShift=10ULL;
constexpr int64_t xbox_epAliveWord=368664ULL;
constexpr int64_t xbox_epRunControl=393212ULL;
constexpr int64_t xbox_epAliveMagic=13421772ULL;
constexpr int64_t xbox_ac97GlobalControl=300ULL;
constexpr int64_t xbox_ac97GlobalStatus=304ULL;
bool xbox_apuTrace=(go_os_Getenv(std::string("RR_APU_TRACE",12)) != std::string("",0));
constexpr int64_t xbox_apuGPControl=8192ULL;
constexpr int64_t xbox_apuGPScratchBase=8256ULL;
constexpr int64_t xbox_gpCommandSlot=18448ULL;
constexpr int64_t xbox_isrExitAddr=2399141376ULL;
constexpr int64_t xbox_vblankPeriod=12224433ULL;
Slice<xbox_irqSource> xbox_irqSources=Slice<xbox_irqSource>{xbox_irqSource{std::string("nv2a",4),cast<uint32_t>(3ULL),[](xbox_Machine* m)->bool{
return (cast<uint32_t>((cast<uint32_t>((m->nv.pcrtcIntr & get(m->nv.reg,cast<uint32_t>(1572944ULL)))) & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL));
}
},xbox_irqSource{std::string("usb",3),cast<uint32_t>(1ULL),xbox_Machine_usbIRQ}};
constexpr int64_t xbox_gpuInterruptVector=3ULL;
constexpr int64_t xbox_usbInterruptVector=1ULL;
constexpr int64_t xbox_kpcrSize=704ULL;
constexpr int64_t xbox_kpcrExceptionList=0ULL;
constexpr int64_t xbox_kpcrStackBase=4ULL;
constexpr int64_t xbox_kpcrStackLimit=8ULL;
constexpr int64_t xbox_kpcrSelfPcr=28ULL;
constexpr int64_t xbox_kpcrPrcb=32ULL;
constexpr int64_t xbox_kpcrIrql=36ULL;
constexpr int64_t xbox_kpcrPrcbData=40ULL;
constexpr int64_t xbox_prcbCurrentThread=0ULL;
bool xbox_pumpTrace=(go_os_Getenv(std::string("RR_PUMP",7)) != std::string("",0));
constexpr int64_t xbox_kretNone=-1ULL;
constexpr int64_t xbox_xcShaBlock=64ULL;
const std::string xbox_rawPartition0Key=std::string("\000RAW/Device/Harddisk0/partition0",32);
constexpr int64_t xbox_rawPartition0Size=65536ULL;
constexpr int64_t xbox_hddBytesPerSector=512ULL;
constexpr int64_t xbox_hddSectorsPerUnit=32ULL;
constexpr int64_t xbox_hddBytesPerUnit=16384ULL;
constexpr int64_t xbox_hddTotalUnits=262144ULL;
constexpr int64_t xbox_discBytesPerSector=2048ULL;
constexpr int64_t xbox_dispSupersede=0ULL;
constexpr int64_t xbox_dispOpen=1ULL;
constexpr int64_t xbox_dispCreate=2ULL;
constexpr int64_t xbox_dispOpenIf=3ULL;
constexpr int64_t xbox_dispOverwrite=4ULL;
constexpr int64_t xbox_dispOverwriteIf=5ULL;
constexpr int64_t xbox_infoOpened=1ULL;
constexpr int64_t xbox_infoCreated=2ULL;
constexpr int64_t xbox_infoOverwritten=3ULL;
constexpr int64_t xbox_statusPending=259ULL;
constexpr int64_t xbox_dhType=0ULL;
constexpr int64_t xbox_dhSignalState=4ULL;
constexpr int64_t xbox_dhWaitListHead=8ULL;
Map<uint16_t,std::string> xbox_verifiedNames=Map<uint16_t,std::string>{{cast<uint16_t>(24ULL),std::string("ExQueryNonVolatileSetting",25)},{cast<uint16_t>(37ULL),std::string("FscSetCacheSize",15)},{cast<uint16_t>(44ULL),std::string("HalGetInterruptVector",21)},{cast<uint16_t>(46ULL),std::string("HalReadWritePCISpace",20)},{cast<uint16_t>(47ULL),std::string("HalRegisterShutdownNotification",31)},{cast<uint16_t>(65ULL),std::string("IoCreateDevice",14)},{cast<uint16_t>(98ULL),std::string("KeConnectInterrupt",18)},{cast<uint16_t>(107ULL),std::string("KeInitializeDpc",15)},{cast<uint16_t>(109ULL),std::string("KeInitializeInterrupt",21)},{cast<uint16_t>(113ULL),std::string("KeInitializeTimerEx",19)},{cast<uint16_t>(149ULL),std::string("KeSetTimer",10)},{cast<uint16_t>(150ULL),std::string("KeSetTimerEx",12)},{cast<uint16_t>(165ULL),std::string("MmAllocateContiguousMemory",26)},{cast<uint16_t>(166ULL),std::string("MmAllocateContiguousMemoryEx",28)},{cast<uint16_t>(168ULL),std::string("MmClaimGpuInstanceMemory",24)},{cast<uint16_t>(173ULL),std::string("MmGetPhysicalAddress",20)},{cast<uint16_t>(175ULL),std::string("MmLockUnlockBufferPages",23)},{cast<uint16_t>(180ULL),std::string("MmQueryAllocationSize",21)},{cast<uint16_t>(182ULL),std::string("MmSetAddressProtect",19)},{cast<uint16_t>(184ULL),std::string("NtAllocateVirtualMemory",23)},{cast<uint16_t>(199ULL),std::string("NtFreeVirtualMemory",19)},{cast<uint16_t>(202ULL),std::string("NtOpenFile",10)},{cast<uint16_t>(301ULL),std::string("RtlNtStatusToDosError",21)},{cast<uint16_t>(2ULL),std::string("AvSendTVEncoderOption",21)},{cast<uint16_t>(15ULL),std::string("ExAllocatePoolWithTag",21)},{cast<uint16_t>(23ULL),std::string("ExQueryPoolBlockSize",20)},{cast<uint16_t>(126ULL),std::string("KeQueryPerformanceCounter",25)},{cast<uint16_t>(127ULL),std::string("KeQueryPerformanceFrequency",27)},{cast<uint16_t>(129ULL),std::string("KeRaiseIrqlToDpcLevel",21)},{cast<uint16_t>(143ULL),std::string("KeSetBasePriorityThread",23)},{cast<uint16_t>(151ULL),std::string("KeStallExecutionProcessor",25)},{cast<uint16_t>(160ULL),std::string("KfRaiseIrql",11)},{cast<uint16_t>(161ULL),std::string("KfLowerIrql",11)},{cast<uint16_t>(189ULL),std::string("NtCreateEvent",13)},{cast<uint16_t>(190ULL),std::string("NtCreateFile",12)},{cast<uint16_t>(193ULL),std::string("NtCreateSemaphore",17)},{cast<uint16_t>(207ULL),std::string("NtQueryDirectoryFile",20)},{cast<uint16_t>(210ULL),std::string("NtQueryFullAttributesFile",25)},{cast<uint16_t>(211ULL),std::string("NtQueryInformationFile",22)},{cast<uint16_t>(218ULL),std::string("NtQueryVolumeInformationFile",28)},{cast<uint16_t>(198ULL),std::string("NtFlushBuffersFile",18)},{cast<uint16_t>(219ULL),std::string("NtReadFile",10)},{cast<uint16_t>(222ULL),std::string("NtReleaseSemaphore",18)},{cast<uint16_t>(224ULL),std::string("NtResumeThread",14)},{cast<uint16_t>(225ULL),std::string("NtSetEvent",10)},{cast<uint16_t>(226ULL),std::string("NtSetInformationFile",20)},{cast<uint16_t>(231ULL),std::string("NtSuspendThread",15)},{cast<uint16_t>(236ULL),std::string("NtWriteFile",11)},{cast<uint16_t>(234ULL),std::string("NtWaitForSingleObjectEx",23)},{cast<uint16_t>(246ULL),std::string("ObReferenceObjectByHandle",25)},{cast<uint16_t>(250ULL),std::string("ObfDereferenceObject",20)},{cast<uint16_t>(17ULL),std::string("ExFreePool",10)},{cast<uint16_t>(97ULL),std::string("KeCancelTimer",13)},{cast<uint16_t>(128ULL),std::string("KeQuerySystemTime",17)},{cast<uint16_t>(137ULL),std::string("KeRemoveQueueDpc",16)},{cast<uint16_t>(181ULL),std::string("MmQueryStatistics",17)},{cast<uint16_t>(335ULL),std::string("XcSHAInit",9)},{cast<uint16_t>(336ULL),std::string("XcSHAUpdate",11)},{cast<uint16_t>(337ULL),std::string("XcSHAFinal",10)},{cast<uint16_t>(338ULL),std::string("XcRC4Key",8)},{cast<uint16_t>(339ULL),std::string("XcRC4Crypt",10)},{cast<uint16_t>(154ULL),std::string("KeSystemTime",12)},{cast<uint16_t>(156ULL),std::string("KeTickCount",11)},{cast<uint16_t>(346ULL),std::string("XcDESKeyParity",14)},{cast<uint16_t>(340ULL),std::string("XcHMAC",6)},{cast<uint16_t>(357ULL),std::string("IdexChannelObject",17)},{cast<uint16_t>(1ULL),std::string("AvGetSavedDataAddress",21)},{cast<uint16_t>(3ULL),std::string("AvSetDisplayMode",16)},{cast<uint16_t>(99ULL),std::string("KeGetCurrentThread",18)},{cast<uint16_t>(119ULL),std::string("KeInsertQueueDpc",16)},{cast<uint16_t>(145ULL),std::string("KeSetEvent",10)},{cast<uint16_t>(159ULL),std::string("KeWaitForSingleObject",21)},{cast<uint16_t>(187ULL),std::string("NtClose",7)},{cast<uint16_t>(252ULL),std::string("PhyGetLinkState",15)},{cast<uint16_t>(253ULL),std::string("PhyInitialize",13)},{cast<uint16_t>(255ULL),std::string("PsCreateSystemThreadEx",22)},{cast<uint16_t>(277ULL),std::string("RtlEnterCriticalSection",23)},{cast<uint16_t>(289ULL),std::string("RtlInitAnsiString",17)},{cast<uint16_t>(291ULL),std::string("RtlInitializeCriticalSection",28)},{cast<uint16_t>(294ULL),std::string("RtlLeaveCriticalSection",23)}};
Map<uint16_t,std::string> xbox_ordinalNames=Map<uint16_t,std::string>{{cast<uint16_t>(1ULL),std::string("AvGetSavedDataAddress",21)},{cast<uint16_t>(2ULL),std::string("AvSendTVEncoderOption",21)},{cast<uint16_t>(3ULL),std::string("AvSetDisplayMode",16)},{cast<uint16_t>(4ULL),std::string("AvSetSavedDataAddress",21)},{cast<uint16_t>(5ULL),std::string("DbgBreakPoint",13)},{cast<uint16_t>(6ULL),std::string("DbgBreakPointWithStatus",23)},{cast<uint16_t>(7ULL),std::string("DbgLoadImageSymbols",19)},{cast<uint16_t>(8ULL),std::string("DbgPrint",8)},{cast<uint16_t>(9ULL),std::string("HalReadSMCTrayState",19)},{cast<uint16_t>(10ULL),std::string("DbgPrompt",9)},{cast<uint16_t>(11ULL),std::string("DbgUnLoadImageSymbols",21)},{cast<uint16_t>(12ULL),std::string("ExAcquireReadWriteLockExclusive",31)},{cast<uint16_t>(13ULL),std::string("ExAcquireReadWriteLockShared",28)},{cast<uint16_t>(14ULL),std::string("ExAllocatePool",14)},{cast<uint16_t>(15ULL),std::string("ExAllocatePoolWithTag",21)},{cast<uint16_t>(16ULL),std::string("ExFreePool",10)},{cast<uint16_t>(17ULL),std::string("ExInitializeReadWriteLock",25)},{cast<uint16_t>(18ULL),std::string("ExInterlockedAddLargeInteger",28)},{cast<uint16_t>(19ULL),std::string("ExInterlockedAddLargeStatistic",30)},{cast<uint16_t>(20ULL),std::string("ExInterlockedCompareExchange64",30)},{cast<uint16_t>(21ULL),std::string("ExQueryPoolBlockSize",20)},{cast<uint16_t>(22ULL),std::string("ExQueryNonVolatileSetting",25)},{cast<uint16_t>(23ULL),std::string("ExReadWriteRefurbInfo",21)},{cast<uint16_t>(24ULL),std::string("ExRaiseException",16)},{cast<uint16_t>(25ULL),std::string("ExRaiseStatus",13)},{cast<uint16_t>(26ULL),std::string("ExReleaseReadWriteLock",22)},{cast<uint16_t>(27ULL),std::string("ExSaveNonVolatileSetting",24)},{cast<uint16_t>(28ULL),std::string("ExSemaphoreObjectType",21)},{cast<uint16_t>(29ULL),std::string("ExTimerObjectType",17)},{cast<uint16_t>(30ULL),std::string("ExfInterlockedInsertHeadList",28)},{cast<uint16_t>(31ULL),std::string("ExfInterlockedInsertTailList",28)},{cast<uint16_t>(32ULL),std::string("ExfInterlockedRemoveHeadList",28)},{cast<uint16_t>(33ULL),std::string("FscGetCacheSize",15)},{cast<uint16_t>(34ULL),std::string("FscInvalidateIdleBlocks",23)},{cast<uint16_t>(35ULL),std::string("FscSetCacheSize",15)},{cast<uint16_t>(36ULL),std::string("HalClearSoftwareInterrupt",25)},{cast<uint16_t>(37ULL),std::string("HalDisableSystemInterrupt",25)},{cast<uint16_t>(38ULL),std::string("HalDiskCachePartitionCount",26)},{cast<uint16_t>(39ULL),std::string("HalDiskModelNumber",18)},{cast<uint16_t>(40ULL),std::string("HalDiskSerialNumber",19)},{cast<uint16_t>(41ULL),std::string("HalEnableSystemInterrupt",24)},{cast<uint16_t>(42ULL),std::string("HalGetInterruptVector",21)},{cast<uint16_t>(43ULL),std::string("HalReadSMBusValue",17)},{cast<uint16_t>(44ULL),std::string("HalReadWritePCISpace",20)},{cast<uint16_t>(45ULL),std::string("HalRegisterShutdownNotification",31)},{cast<uint16_t>(46ULL),std::string("HalRequestSoftwareInterrupt",27)},{cast<uint16_t>(47ULL),std::string("HalReturnToFirmware",19)},{cast<uint16_t>(48ULL),std::string("HalWriteSMBusValue",18)},{cast<uint16_t>(49ULL),std::string("InterlockedCompareExchange",26)},{cast<uint16_t>(50ULL),std::string("InterlockedDecrement",20)},{cast<uint16_t>(51ULL),std::string("InterlockedIncrement",20)},{cast<uint16_t>(52ULL),std::string("InterlockedExchange",19)},{cast<uint16_t>(53ULL),std::string("InterlockedExchangeAdd",22)},{cast<uint16_t>(54ULL),std::string("InterlockedFlushSList",21)},{cast<uint16_t>(55ULL),std::string("InterlockedPopEntrySList",24)},{cast<uint16_t>(56ULL),std::string("InterlockedPushEntrySList",25)},{cast<uint16_t>(57ULL),std::string("IoAllocateIrp",13)},{cast<uint16_t>(58ULL),std::string("IoBuildAsynchronousFsdRequest",29)},{cast<uint16_t>(59ULL),std::string("IoBuildDeviceIoControlRequest",29)},{cast<uint16_t>(60ULL),std::string("IoBuildSynchronousFsdRequest",28)},{cast<uint16_t>(61ULL),std::string("IoCheckShareAccess",18)},{cast<uint16_t>(62ULL),std::string("IoCompletionObjectType",22)},{cast<uint16_t>(63ULL),std::string("IoCreateDevice",14)},{cast<uint16_t>(64ULL),std::string("IoCreateFile",12)},{cast<uint16_t>(65ULL),std::string("IoCreateSymbolicLink",20)},{cast<uint16_t>(66ULL),std::string("IoDeleteDevice",14)},{cast<uint16_t>(67ULL),std::string("IoDeleteSymbolicLink",20)},{cast<uint16_t>(68ULL),std::string("IoFreeIrp",9)},{cast<uint16_t>(69ULL),std::string("IoInitializeIrp",15)},{cast<uint16_t>(70ULL),std::string("IoInvalidDeviceRequest",22)},{cast<uint16_t>(71ULL),std::string("IoQueryFileInformation",22)},{cast<uint16_t>(72ULL),std::string("IoQueryVolumeInformation",24)},{cast<uint16_t>(73ULL),std::string("IoQueueThreadIrp",16)},{cast<uint16_t>(74ULL),std::string("IoRemoveShareAccess",19)},{cast<uint16_t>(75ULL),std::string("IoSetIoCompletion",17)},{cast<uint16_t>(76ULL),std::string("IoSetShareAccess",16)},{cast<uint16_t>(77ULL),std::string("IoStartNextPacket",17)},{cast<uint16_t>(78ULL),std::string("IoStartNextPacketByKey",22)},{cast<uint16_t>(79ULL),std::string("IoStartPacket",13)},{cast<uint16_t>(80ULL),std::string("IoSynchronousDeviceIoControlRequest",35)},{cast<uint16_t>(81ULL),std::string("IoSynchronousFsdRequest",23)},{cast<uint16_t>(82ULL),std::string("IofCallDriver",13)},{cast<uint16_t>(83ULL),std::string("IofCompleteRequest",18)},{cast<uint16_t>(84ULL),std::string("KdDebuggerEnabled",17)},{cast<uint16_t>(85ULL),std::string("KdDebuggerNotPresent",20)},{cast<uint16_t>(86ULL),std::string("IoDismountVolume",16)},{cast<uint16_t>(87ULL),std::string("IoDismountVolumeByName",22)},{cast<uint16_t>(88ULL),std::string("KeAlertResumeThread",19)},{cast<uint16_t>(89ULL),std::string("KeAlertThread",13)},{cast<uint16_t>(90ULL),std::string("KeBoostPriorityThread",21)},{cast<uint16_t>(91ULL),std::string("KeBugCheck",10)},{cast<uint16_t>(92ULL),std::string("KeBugCheckEx",12)},{cast<uint16_t>(93ULL),std::string("KeCancelTimer",13)},{cast<uint16_t>(94ULL),std::string("KeConnectInterrupt",18)},{cast<uint16_t>(95ULL),std::string("KeDelayExecutionThread",22)},{cast<uint16_t>(96ULL),std::string("KeDisconnectInterrupt",21)},{cast<uint16_t>(97ULL),std::string("KeEnterCriticalRegion",21)},{cast<uint16_t>(98ULL),std::string("KeGetCurrentIrql",16)},{cast<uint16_t>(99ULL),std::string("KeGetCurrentThread",18)},{cast<uint16_t>(100ULL),std::string("KeInitializeApc",15)},{cast<uint16_t>(101ULL),std::string("KeInitializeDeviceQueue",23)},{cast<uint16_t>(102ULL),std::string("KeInitializeDpc",15)},{cast<uint16_t>(103ULL),std::string("KeInitializeEvent",17)},{cast<uint16_t>(104ULL),std::string("KeInitializeInterrupt",21)},{cast<uint16_t>(105ULL),std::string("KeInitializeMutant",18)},{cast<uint16_t>(106ULL),std::string("KeInitializeQueue",17)},{cast<uint16_t>(107ULL),std::string("KeInitializeSemaphore",21)},{cast<uint16_t>(108ULL),std::string("KeInitializeTimerEx",19)},{cast<uint16_t>(109ULL),std::string("KeInsertByKeyDeviceQueue",24)},{cast<uint16_t>(110ULL),std::string("KeInsertDeviceQueue",19)},{cast<uint16_t>(111ULL),std::string("KeInsertHeadQueue",17)},{cast<uint16_t>(112ULL),std::string("KeInsertQueue",13)},{cast<uint16_t>(113ULL),std::string("KeInsertQueueApc",16)},{cast<uint16_t>(114ULL),std::string("KeInsertQueueDpc",16)},{cast<uint16_t>(115ULL),std::string("KeInterruptTime",15)},{cast<uint16_t>(116ULL),std::string("KeIsExecutingDpc",16)},{cast<uint16_t>(117ULL),std::string("KeLeaveCriticalRegion",21)},{cast<uint16_t>(118ULL),std::string("KePulseEvent",12)},{cast<uint16_t>(119ULL),std::string("KeQueryBasePriorityThread",25)},{cast<uint16_t>(120ULL),std::string("KeQueryInterruptTime",20)},{cast<uint16_t>(121ULL),std::string("KeQueryPerformanceCounter",25)},{cast<uint16_t>(122ULL),std::string("KeQueryPerformanceFrequency",27)},{cast<uint16_t>(123ULL),std::string("KeQuerySystemTime",17)},{cast<uint16_t>(124ULL),std::string("KeRaiseIrqlToDpcLevel",21)},{cast<uint16_t>(125ULL),std::string("KeRaiseIrqlToSynchLevel",23)},{cast<uint16_t>(126ULL),std::string("KeReleaseMutant",15)},{cast<uint16_t>(127ULL),std::string("KeReleaseSemaphore",18)},{cast<uint16_t>(128ULL),std::string("KeRemoveByKeyDeviceQueue",24)},{cast<uint16_t>(129ULL),std::string("KeRemoveDeviceQueue",19)},{cast<uint16_t>(130ULL),std::string("KeRemoveEntryDeviceQueue",24)},{cast<uint16_t>(131ULL),std::string("KeRemoveQueue",13)},{cast<uint16_t>(132ULL),std::string("KeRemoveQueueDpc",16)},{cast<uint16_t>(133ULL),std::string("KeResetEvent",12)},{cast<uint16_t>(134ULL),std::string("KeRestoreFloatingPointState",27)},{cast<uint16_t>(135ULL),std::string("KeResumeThread",14)},{cast<uint16_t>(136ULL),std::string("KeRundownQueue",14)},{cast<uint16_t>(137ULL),std::string("KeSaveFloatingPointState",24)},{cast<uint16_t>(138ULL),std::string("KeSetBasePriorityThread",23)},{cast<uint16_t>(139ULL),std::string("KeSetDisableBoostThread",23)},{cast<uint16_t>(140ULL),std::string("KeSetEvent",10)},{cast<uint16_t>(141ULL),std::string("KeSetEventBoostPriority",23)},{cast<uint16_t>(142ULL),std::string("KeSetPriorityProcess",20)},{cast<uint16_t>(143ULL),std::string("KeSetPriorityThread",19)},{cast<uint16_t>(144ULL),std::string("KeSetTimer",10)},{cast<uint16_t>(145ULL),std::string("KeSetTimerEx",12)},{cast<uint16_t>(146ULL),std::string("KeStallExecutionProcessor",25)},{cast<uint16_t>(147ULL),std::string("KeSuspendThread",15)},{cast<uint16_t>(148ULL),std::string("KeSynchronizeExecution",22)},{cast<uint16_t>(149ULL),std::string("KeSystemTime",12)},{cast<uint16_t>(150ULL),std::string("KeTestAlertThread",17)},{cast<uint16_t>(151ULL),std::string("KeTickCount",11)},{cast<uint16_t>(152ULL),std::string("KeTimeIncrement",15)},{cast<uint16_t>(153ULL),std::string("KeWaitForMultipleObjects",24)},{cast<uint16_t>(154ULL),std::string("KeWaitForSingleObject",21)},{cast<uint16_t>(155ULL),std::string("KfRaiseIrql",11)},{cast<uint16_t>(156ULL),std::string("KfLowerIrql",11)},{cast<uint16_t>(157ULL),std::string("KiBugCheckData",14)},{cast<uint16_t>(158ULL),std::string("KiUnlockDispatcherDatabase",26)},{cast<uint16_t>(159ULL),std::string("LaunchDataPage",14)},{cast<uint16_t>(160ULL),std::string("MmAllocateContiguousMemory",26)},{cast<uint16_t>(161ULL),std::string("MmAllocateContiguousMemoryEx",28)},{cast<uint16_t>(162ULL),std::string("MmAllocateSystemMemory",22)},{cast<uint16_t>(163ULL),std::string("MmClaimGpuInstanceMemory",24)},{cast<uint16_t>(164ULL),std::string("MmCreateKernelStack",19)},{cast<uint16_t>(165ULL),std::string("MmDeleteKernelStack",19)},{cast<uint16_t>(166ULL),std::string("MmFreeContiguousMemory",22)},{cast<uint16_t>(167ULL),std::string("MmFreeSystemMemory",18)},{cast<uint16_t>(168ULL),std::string("MmGetPhysicalAddress",20)},{cast<uint16_t>(169ULL),std::string("MmIsAddressValid",16)},{cast<uint16_t>(170ULL),std::string("MmLockUnlockBufferPages",23)},{cast<uint16_t>(171ULL),std::string("MmLockUnlockPhysicalPage",24)},{cast<uint16_t>(172ULL),std::string("MmMapIoSpace",12)},{cast<uint16_t>(173ULL),std::string("MmPersistContiguousMemory",25)},{cast<uint16_t>(174ULL),std::string("MmQueryAddressProtect",21)},{cast<uint16_t>(175ULL),std::string("MmQueryAllocationSize",21)},{cast<uint16_t>(176ULL),std::string("MmQueryStatistics",17)},{cast<uint16_t>(177ULL),std::string("MmSetAddressProtect",19)},{cast<uint16_t>(178ULL),std::string("MmUnmapIoSpace",14)},{cast<uint16_t>(179ULL),std::string("NtAllocateVirtualMemory",23)},{cast<uint16_t>(180ULL),std::string("NtCancelTimer",13)},{cast<uint16_t>(181ULL),std::string("NtClearEvent",12)},{cast<uint16_t>(182ULL),std::string("NtClose",7)},{cast<uint16_t>(183ULL),std::string("NtCreateDirectoryObject",23)},{cast<uint16_t>(184ULL),std::string("NtCreateEvent",13)},{cast<uint16_t>(185ULL),std::string("NtCreateFile",12)},{cast<uint16_t>(186ULL),std::string("NtCreateIoCompletion",20)},{cast<uint16_t>(187ULL),std::string("NtCreateMutant",14)},{cast<uint16_t>(188ULL),std::string("NtCreateSemaphore",17)},{cast<uint16_t>(189ULL),std::string("NtCreateTimer",13)},{cast<uint16_t>(190ULL),std::string("NtDeleteFile",12)},{cast<uint16_t>(191ULL),std::string("NtDeviceIoControlFile",21)},{cast<uint16_t>(192ULL),std::string("NtDuplicateObject",17)},{cast<uint16_t>(193ULL),std::string("NtFlushBuffersFile",18)},{cast<uint16_t>(194ULL),std::string("NtFreeVirtualMemory",19)},{cast<uint16_t>(195ULL),std::string("NtFsControlFile",15)},{cast<uint16_t>(196ULL),std::string("NtOpenDirectoryObject",21)},{cast<uint16_t>(197ULL),std::string("NtOpenFile",10)},{cast<uint16_t>(198ULL),std::string("NtOpenSymbolicLinkObject",24)},{cast<uint16_t>(199ULL),std::string("NtProtectVirtualMemory",22)},{cast<uint16_t>(200ULL),std::string("NtPulseEvent",12)},{cast<uint16_t>(201ULL),std::string("NtQueueApcThread",16)},{cast<uint16_t>(202ULL),std::string("NtQueryDirectoryFile",20)},{cast<uint16_t>(203ULL),std::string("NtQueryDirectoryObject",22)},{cast<uint16_t>(204ULL),std::string("NtQueryEvent",12)},{cast<uint16_t>(205ULL),std::string("NtQueryFullAttributesFile",25)},{cast<uint16_t>(206ULL),std::string("NtQueryInformationFile",22)},{cast<uint16_t>(207ULL),std::string("NtQueryIoCompletion",19)},{cast<uint16_t>(208ULL),std::string("NtQueryMutant",13)},{cast<uint16_t>(209ULL),std::string("NtQuerySemaphore",16)},{cast<uint16_t>(210ULL),std::string("NtQuerySymbolicLinkObject",25)},{cast<uint16_t>(211ULL),std::string("NtQueryTimer",12)},{cast<uint16_t>(212ULL),std::string("NtQueryVirtualMemory",20)},{cast<uint16_t>(213ULL),std::string("NtQueryVolumeInformationFile",28)},{cast<uint16_t>(214ULL),std::string("NtReadFile",10)},{cast<uint16_t>(215ULL),std::string("NtReadFileScatter",17)},{cast<uint16_t>(216ULL),std::string("NtReleaseMutant",15)},{cast<uint16_t>(217ULL),std::string("NtReleaseSemaphore",18)},{cast<uint16_t>(218ULL),std::string("NtRemoveIoCompletion",20)},{cast<uint16_t>(219ULL),std::string("NtResumeThread",14)},{cast<uint16_t>(220ULL),std::string("NtSetEvent",10)},{cast<uint16_t>(221ULL),std::string("NtSetInformationFile",20)},{cast<uint16_t>(222ULL),std::string("NtSetIoCompletion",17)},{cast<uint16_t>(223ULL),std::string("NtSetSystemTime",15)},{cast<uint16_t>(224ULL),std::string("NtSetTimerEx",12)},{cast<uint16_t>(225ULL),std::string("NtSignalAndWaitForSingleObjectEx",32)},{cast<uint16_t>(226ULL),std::string("NtSuspendThread",15)},{cast<uint16_t>(227ULL),std::string("NtUserIoApcDispatcher",21)},{cast<uint16_t>(228ULL),std::string("NtWaitForSingleObject",21)},{cast<uint16_t>(229ULL),std::string("NtWaitForSingleObjectEx",23)},{cast<uint16_t>(230ULL),std::string("NtWaitForMultipleObjectsEx",26)},{cast<uint16_t>(231ULL),std::string("NtWriteFile",11)},{cast<uint16_t>(232ULL),std::string("NtWriteFileGather",17)},{cast<uint16_t>(233ULL),std::string("NtYieldExecution",16)},{cast<uint16_t>(234ULL),std::string("ObCreateObject",14)},{cast<uint16_t>(235ULL),std::string("ObDirectoryObjectType",21)},{cast<uint16_t>(236ULL),std::string("ObInsertObject",14)},{cast<uint16_t>(237ULL),std::string("ObMakeTemporaryObject",21)},{cast<uint16_t>(238ULL),std::string("ObOpenObjectByName",18)},{cast<uint16_t>(239ULL),std::string("ObOpenObjectByPointer",21)},{cast<uint16_t>(240ULL),std::string("ObpObjectHandleTable",20)},{cast<uint16_t>(241ULL),std::string("ObReferenceObjectByHandle",25)},{cast<uint16_t>(242ULL),std::string("ObReferenceObjectByName",23)},{cast<uint16_t>(243ULL),std::string("ObReferenceObjectByPointer",26)},{cast<uint16_t>(244ULL),std::string("ObSymbolicLinkObjectType",24)},{cast<uint16_t>(245ULL),std::string("ObfDereferenceObject",20)},{cast<uint16_t>(246ULL),std::string("ObfReferenceObject",18)},{cast<uint16_t>(249ULL),std::string("PsCreateSystemThread?",21)},{cast<uint16_t>(250ULL),std::string("PsCreateSystemThread?",21)},{cast<uint16_t>(251ULL),std::string("PsCreateSystemThread?",21)},{cast<uint16_t>(252ULL),std::string("PsCreateSystemThread?",21)},{cast<uint16_t>(253ULL),std::string("PsCreateSystemThread",20)},{cast<uint16_t>(254ULL),std::string("PsCreateSystemThread",20)},{cast<uint16_t>(255ULL),std::string("PsCreateSystemThreadEx",22)},{cast<uint16_t>(256ULL),std::string("PsQueryStatistics",17)},{cast<uint16_t>(257ULL),std::string("PsSetCreateThreadNotifyRoutine",30)},{cast<uint16_t>(258ULL),std::string("PsTerminateSystemThread",23)},{cast<uint16_t>(259ULL),std::string("PsThreadObjectType",18)},{cast<uint16_t>(260ULL),std::string("RtlAnsiStringToUnicodeString",28)},{cast<uint16_t>(261ULL),std::string("RtlAppendStringToString",23)},{cast<uint16_t>(262ULL),std::string("RtlAppendUnicodeStringToString",30)},{cast<uint16_t>(263ULL),std::string("RtlAppendUnicodeToString",24)},{cast<uint16_t>(264ULL),std::string("RtlAssert",9)},{cast<uint16_t>(265ULL),std::string("RtlCaptureContext",17)},{cast<uint16_t>(266ULL),std::string("RtlCaptureStackBackTrace",24)},{cast<uint16_t>(267ULL),std::string("RtlCharToInteger",16)},{cast<uint16_t>(268ULL),std::string("RtlCompareMemory",16)},{cast<uint16_t>(269ULL),std::string("RtlCompareMemoryUlong",21)},{cast<uint16_t>(270ULL),std::string("RtlCompareString",16)},{cast<uint16_t>(271ULL),std::string("RtlCompareUnicodeString",23)},{cast<uint16_t>(272ULL),std::string("RtlCopyString",13)},{cast<uint16_t>(273ULL),std::string("RtlCopyUnicodeString",20)},{cast<uint16_t>(274ULL),std::string("RtlCreateUnicodeString",22)},{cast<uint16_t>(275ULL),std::string("RtlDowncaseUnicodeChar",22)},{cast<uint16_t>(276ULL),std::string("RtlDowncaseUnicodeString",24)},{cast<uint16_t>(277ULL),std::string("RtlEnterCriticalSection",23)},{cast<uint16_t>(278ULL),std::string("RtlEnterCriticalSectionAndRegion",32)},{cast<uint16_t>(279ULL),std::string("RtlEqualString",14)},{cast<uint16_t>(280ULL),std::string("RtlEqualUnicodeString",21)},{cast<uint16_t>(281ULL),std::string("RtlExtendedIntegerMultiply",26)},{cast<uint16_t>(282ULL),std::string("RtlExtendedLargeIntegerDivide",29)},{cast<uint16_t>(283ULL),std::string("RtlExtendedMagicDivide",22)},{cast<uint16_t>(284ULL),std::string("RtlFillMemory",13)},{cast<uint16_t>(285ULL),std::string("RtlFillMemoryUlong",18)},{cast<uint16_t>(286ULL),std::string("RtlFreeAnsiString",17)},{cast<uint16_t>(287ULL),std::string("RtlFreeUnicodeString",20)},{cast<uint16_t>(288ULL),std::string("RtlGetCallersAddress",20)},{cast<uint16_t>(289ULL),std::string("RtlInitAnsiString",17)},{cast<uint16_t>(290ULL),std::string("RtlInitUnicodeString",20)},{cast<uint16_t>(291ULL),std::string("RtlInitializeCriticalSection",28)},{cast<uint16_t>(292ULL),std::string("RtlIntegerToChar",16)},{cast<uint16_t>(293ULL),std::string("RtlIntegerToUnicodeString",25)},{cast<uint16_t>(294ULL),std::string("RtlLeaveCriticalSection",23)},{cast<uint16_t>(295ULL),std::string("RtlLeaveCriticalSectionAndRegion",32)},{cast<uint16_t>(296ULL),std::string("RtlLowerChar",12)},{cast<uint16_t>(297ULL),std::string("RtlMapGenericMask",17)},{cast<uint16_t>(298ULL),std::string("RtlMoveMemory",13)},{cast<uint16_t>(299ULL),std::string("RtlMultiByteToUnicodeN",22)},{cast<uint16_t>(300ULL),std::string("RtlMultiByteToUnicodeSize",25)},{cast<uint16_t>(301ULL),std::string("RtlNtStatusToDosError",21)},{cast<uint16_t>(302ULL),std::string("RtlRaiseException",17)},{cast<uint16_t>(303ULL),std::string("RtlRaiseStatus",14)},{cast<uint16_t>(304ULL),std::string("RtlTimeFieldsToTime",19)},{cast<uint16_t>(305ULL),std::string("RtlTimeToTimeFields",19)},{cast<uint16_t>(306ULL),std::string("RtlTryEnterCriticalSection",26)},{cast<uint16_t>(307ULL),std::string("RtlUlongByteSwap",16)},{cast<uint16_t>(308ULL),std::string("RtlUnicodeStringToAnsiString",28)},{cast<uint16_t>(309ULL),std::string("RtlUnicodeStringToInteger",25)},{cast<uint16_t>(310ULL),std::string("RtlUnicodeToMultiByteN",22)},{cast<uint16_t>(311ULL),std::string("RtlUnicodeToMultiByteSize",25)},{cast<uint16_t>(312ULL),std::string("RtlUpcaseUnicodeChar",20)},{cast<uint16_t>(313ULL),std::string("RtlUpcaseUnicodeString",22)},{cast<uint16_t>(314ULL),std::string("RtlUpcaseUnicodeToMultiByteN",28)},{cast<uint16_t>(315ULL),std::string("RtlUpperChar",12)},{cast<uint16_t>(316ULL),std::string("RtlUpperString",14)},{cast<uint16_t>(317ULL),std::string("RtlUshortByteSwap",17)},{cast<uint16_t>(318ULL),std::string("RtlWalkFrameChain",17)},{cast<uint16_t>(319ULL),std::string("RtlZeroMemory",13)},{cast<uint16_t>(320ULL),std::string("XboxEEPROMKey",13)},{cast<uint16_t>(321ULL),std::string("XboxHardwareInfo",16)},{cast<uint16_t>(322ULL),std::string("XboxHDKey",9)},{cast<uint16_t>(323ULL),std::string("XboxKrnlVersion",15)},{cast<uint16_t>(324ULL),std::string("XboxSignatureKey",16)},{cast<uint16_t>(325ULL),std::string("XeImageFileName",15)},{cast<uint16_t>(326ULL),std::string("XeLoadSection",13)},{cast<uint16_t>(327ULL),std::string("XeUnloadSection",15)},{cast<uint16_t>(328ULL),std::string("XcSHAInit",9)},{cast<uint16_t>(329ULL),std::string("XcSHAUpdate",11)},{cast<uint16_t>(330ULL),std::string("XcSHAFinal",10)},{cast<uint16_t>(331ULL),std::string("XcRC4Key",8)},{cast<uint16_t>(332ULL),std::string("XcRC4Crypt",10)},{cast<uint16_t>(333ULL),std::string("XcHMAC",6)},{cast<uint16_t>(334ULL),std::string("XcPKEncPublic",13)},{cast<uint16_t>(335ULL),std::string("XcPKDecPrivate",14)},{cast<uint16_t>(336ULL),std::string("XcPKGetKeyLen",13)},{cast<uint16_t>(337ULL),std::string("XcVerifyPKCS1Signature",22)},{cast<uint16_t>(338ULL),std::string("XcModExp",8)},{cast<uint16_t>(339ULL),std::string("XcDESKeyParity",14)},{cast<uint16_t>(340ULL),std::string("XcKeyTable",10)},{cast<uint16_t>(341ULL),std::string("XcBlockCrypt",12)},{cast<uint16_t>(342ULL),std::string("XcBlockCryptCBC",15)},{cast<uint16_t>(343ULL),std::string("XcCryptService",14)},{cast<uint16_t>(344ULL),std::string("XcUpdateCrypto",14)},{cast<uint16_t>(345ULL),std::string("RtlRip",6)},{cast<uint16_t>(346ULL),std::string("XboxLANKey",10)},{cast<uint16_t>(347ULL),std::string("XboxAlternateSignatureKeys",26)},{cast<uint16_t>(348ULL),std::string("XePublicKeyData",15)},{cast<uint16_t>(349ULL),std::string("HalBootSMCVideoMode",19)},{cast<uint16_t>(350ULL),std::string("IdexChannelObject",17)},{cast<uint16_t>(351ULL),std::string("HalIsResetOrShutdownPending",27)},{cast<uint16_t>(352ULL),std::string("IoMarkIrpMustComplete",21)},{cast<uint16_t>(353ULL),std::string("HalInitiateShutdown",19)},{cast<uint16_t>(354ULL),std::string("RtlSnprintf",11)},{cast<uint16_t>(355ULL),std::string("RtlSprintf",10)},{cast<uint16_t>(356ULL),std::string("RtlVsnprintf",12)},{cast<uint16_t>(357ULL),std::string("RtlVsprintf",11)},{cast<uint16_t>(358ULL),std::string("HalEnableSecureTrayEject",24)},{cast<uint16_t>(359ULL),std::string("HalWriteSMCScratchRegister",26)},{cast<uint16_t>(360ULL),std::string("MmDbgAllocateMemory",19)}};
constexpr int64_t xbox_ramSize=67108864ULL;
constexpr int64_t xbox_cachedBase=2147483648ULL;
constexpr int64_t xbox_uncachedBase=2952790016ULL;
constexpr int64_t xbox_physBase=3489660928ULL;
constexpr int64_t xbox_wcBase=4026531840ULL;
constexpr int64_t xbox_windowMask=67108863ULL;
constexpr int64_t xbox_mmioBase=4244635648ULL;
constexpr int64_t xbox_mmioTop=4261412864ULL;
constexpr int64_t xbox_kernelBandBase=66846720ULL;
constexpr int64_t xbox_kpcrAddr=66846720ULL;
constexpr int64_t xbox_trapBase=2399141888ULL;
constexpr int64_t xbox_trapStride=16ULL;
constexpr int64_t xbox_trapCount=512ULL;
constexpr int64_t xbox_trapTop=2399150080ULL;
constexpr int64_t xbox_titleStackTop=66842624ULL;
constexpr int64_t xbox_titleStackSize=65536ULL;
constexpr int64_t xbox_fsSelector=56ULL;
constexpr int64_t xbox_nvApertureSize=16777216ULL;
constexpr int64_t xbox_nvPMC_ID=0ULL;
constexpr int64_t xbox_nvPFIFO_RUNOUT=9216ULL;
constexpr int64_t xbox_nvPFIFO_C1_STATUS=12820ULL;
constexpr int64_t xbox_nvPFIFO_DMA_PUSH=12832ULL;
constexpr int64_t xbox_nvPFIFO_DMA_PUT=12864ULL;
constexpr int64_t xbox_nvPFIFO_DMA_GET=12868ULL;
constexpr int64_t xbox_nvUSER_PUT=8388672ULL;
constexpr int64_t xbox_nvUSER_GET=8388676ULL;
constexpr int64_t xbox_nvPFB_FLUSH=1049616ULL;
constexpr int64_t xbox_nvPMC_INTR=256ULL;
constexpr int64_t xbox_nvPFIFO_INTR=8448ULL;
constexpr int64_t xbox_nvPGRAPH_INTR=4194560ULL;
constexpr int64_t xbox_nvPCRTC_INTR=6291712ULL;
constexpr int64_t xbox_nvPMC_INTR_EN=320ULL;
constexpr int64_t xbox_nvPCRTC_INTR_EN=6291776ULL;
constexpr int64_t xbox_nvPGRAPH_SEMAPHORE=4197136ULL;
bool xbox_nvTrace=(go_os_Getenv(std::string("RR_NV_TRACE",11)) != std::string("",0));
constexpr int64_t xbox_class2DSurfaces=98ULL;
constexpr int64_t xbox_classImageBlit=159ULL;
constexpr int64_t xbox_surf2DFormat=768ULL;
constexpr int64_t xbox_surf2DPitch=772ULL;
constexpr int64_t xbox_surf2DSrcOff=776ULL;
constexpr int64_t xbox_surf2DDstOff=780ULL;
constexpr int64_t xbox_blitPointIn=768ULL;
constexpr int64_t xbox_blitPointOut=772ULL;
constexpr int64_t xbox_blitSize=776ULL;
constexpr double xbox_clipWEps=1.52587890625000000e-05;
constexpr int64_t xbox_kelvinCombinerAlphaICW=608ULL;
constexpr int64_t xbox_kelvinSpecFogCW0=648ULL;
constexpr int64_t xbox_kelvinSpecFogCW1=652ULL;
constexpr int64_t xbox_kelvinCombinerFactor0=2656ULL;
constexpr int64_t xbox_kelvinCombinerFactor1=2688ULL;
constexpr int64_t xbox_kelvinCombinerAlphaOCW=2720ULL;
constexpr int64_t xbox_kelvinCombinerColorICW=2752ULL;
constexpr int64_t xbox_kelvinCombinerColorOCW=7744ULL;
constexpr int64_t xbox_kelvinCombinerControl=7776ULL;
Map<uint32_t,std::string> xbox_kelvinMethodNames=Map<uint32_t,std::string>{{cast<uint32_t>(388ULL),std::string("SET_CONTEXT_DMA_A",17)},{cast<uint32_t>(392ULL),std::string("SET_CONTEXT_DMA_B",17)},{cast<uint32_t>(412ULL),std::string("SET_CONTEXT_DMA_VERTEX_A",24)},{cast<uint32_t>(416ULL),std::string("SET_CONTEXT_DMA_VERTEX_B",24)},{cast<uint32_t>(420ULL),std::string("SET_CONTEXT_DMA_SEMAPHORE",25)},{cast<uint32_t>(512ULL),std::string("SET_SURFACE_CLIP_HORIZONTAL",27)},{cast<uint32_t>(516ULL),std::string("SET_SURFACE_CLIP_VERTICAL",25)},{cast<uint32_t>(520ULL),std::string("SET_SURFACE_FORMAT",18)},{cast<uint32_t>(524ULL),std::string("SET_SURFACE_PITCH",17)},{cast<uint32_t>(528ULL),std::string("SET_SURFACE_COLOR_OFFSET",24)},{cast<uint32_t>(532ULL),std::string("SET_SURFACE_ZETA_OFFSET",23)},{cast<uint32_t>(692ULL),std::string("SET_WINDOW_CLIP_TYPE",20)},{cast<uint32_t>(768ULL),std::string("SET_ALPHA_TEST_ENABLE",21)},{cast<uint32_t>(772ULL),std::string("SET_BLEND_ENABLE",16)},{cast<uint32_t>(776ULL),std::string("SET_CULL_FACE_ENABLE",20)},{cast<uint32_t>(780ULL),std::string("SET_DEPTH_TEST_ENABLE",21)},{cast<uint32_t>(828ULL),std::string("SET_ALPHA_FUNC",14)},{cast<uint32_t>(832ULL),std::string("SET_ALPHA_REF",13)},{cast<uint32_t>(836ULL),std::string("SET_BLEND_FUNC_SFACTOR",22)},{cast<uint32_t>(840ULL),std::string("SET_BLEND_FUNC_DFACTOR",22)},{cast<uint32_t>(844ULL),std::string("SET_BLEND_COLOR",15)},{cast<uint32_t>(848ULL),std::string("SET_BLEND_EQUATION",18)},{cast<uint32_t>(852ULL),std::string("SET_DEPTH_FUNC",14)},{cast<uint32_t>(856ULL),std::string("SET_COLOR_MASK",14)},{cast<uint32_t>(860ULL),std::string("SET_DEPTH_MASK",14)},{cast<uint32_t>(892ULL),std::string("SET_SHADE_MODE",14)},{cast<uint32_t>(924ULL),std::string("SET_CULL_FACE",13)},{cast<uint32_t>(928ULL),std::string("SET_FRONT_FACE",14)},{cast<uint32_t>(6140ULL),std::string("SET_BEGIN_END",13)},{cast<uint32_t>(6144ULL),std::string("ARRAY_ELEMENT16",15)},{cast<uint32_t>(6152ULL),std::string("ARRAY_ELEMENT32",15)},{cast<uint32_t>(6160ULL),std::string("DRAW_ARRAYS",11)},{cast<uint32_t>(6168ULL),std::string("INLINE_ARRAY",12)},{cast<uint32_t>(7532ULL),std::string("SET_SEMAPHORE_OFFSET",20)},{cast<uint32_t>(7536ULL),std::string("BACK_END_WRITE_SEMAPHORE_RELEASE",32)},{cast<uint32_t>(7564ULL),std::string("SET_ZSTENCIL_CLEAR_VALUE",24)},{cast<uint32_t>(7568ULL),std::string("SET_COLOR_CLEAR_VALUE",21)},{cast<uint32_t>(7572ULL),std::string("CLEAR_SURFACE",13)},{cast<uint32_t>(7576ULL),std::string("SET_CLEAR_RECT_HORIZONTAL",25)},{cast<uint32_t>(7580ULL),std::string("SET_CLEAR_RECT_VERTICAL",23)},{cast<uint32_t>(7776ULL),std::string("SET_COMBINER_CONTROL",20)},{cast<uint32_t>(7828ULL),std::string("SET_TRANSFORM_EXECUTION_MODE",28)},{cast<uint32_t>(7832ULL),std::string("SET_TRANSFORM_CONSTANT_LOAD_CXT",31)},{cast<uint32_t>(7836ULL),std::string("SET_TRANSFORM_PROGRAM_LOAD",26)},{cast<uint32_t>(7840ULL),std::string("SET_TRANSFORM_PROGRAM_START",27)},{cast<uint32_t>(7844ULL),std::string("SET_TRANSFORM_CONSTANT_LOAD",27)},{cast<uint32_t>(7792ULL),std::string("SET_SHADER_STAGE_PROGRAM",24)}};
constexpr int64_t xbox_NVTextureUnits=4ULL;
constexpr int64_t xbox_kelvinSurfaceClipH=512ULL;
constexpr int64_t xbox_kelvinSurfaceClipV=516ULL;
constexpr int64_t xbox_kelvinSurfaceFormat=520ULL;
constexpr int64_t xbox_kelvinSurfacePitch=524ULL;
constexpr int64_t xbox_kelvinSurfaceColorOffset=528ULL;
constexpr int64_t xbox_kelvinSurfaceZetaOffset=532ULL;
constexpr int64_t xbox_kelvinZStencilClearValue=7564ULL;
constexpr int64_t xbox_kelvinColorClearValue=7568ULL;
constexpr int64_t xbox_kelvinClearSurface=7572ULL;
constexpr int64_t xbox_kelvinClearRectH=7576ULL;
constexpr int64_t xbox_kelvinClearRectV=7580ULL;
bool xbox_nvSemTrace=(go_os_Getenv(std::string("RR_NV_SEM",9)) != std::string("",0));
bool xbox_nvVPTrace=(go_os_Getenv(std::string("RR_NV_VP",8)) != std::string("",0));
bool xbox_nvSurfTrace=(go_os_Getenv(std::string("RR_NV_SURF",10)) != std::string("",0));
bool xbox_shadowTrace=(go_os_Getenv(std::string("RR_SHADOW",9)) != std::string("",0));
bool xbox_lowWriteTrace=(go_os_Getenv(std::string("RR_LOWWRITE",11)) != std::string("",0));
std::string xbox_shadowCmpEnv=go_os_Getenv(std::string("RR_SHADOWCMP",12));
bool xbox_shadowFragTrace=(go_os_Getenv(std::string("RR_SHADOWFRAG",13)) != std::string("",0));
bool xbox_texShaderTrace=(go_os_Getenv(std::string("RR_TEXSHADER",12)) != std::string("",0));
bool xbox_rasterSerial=(go_os_Getenv(std::string("RR_NV_SERIAL",12)) != std::string("",0));
bool xbox_flipVSyncDefault=(go_os_Getenv(std::string("RR_FLIP_VSYNC",13)) != std::string("0",1));
constexpr int64_t xbox_kelvinCtxDmaSemaphore=420ULL;
constexpr int64_t xbox_kelvinSemaphoreOffset=7532ULL;
constexpr int64_t xbox_kelvinSemaphoreRelease=7536ULL;
constexpr int64_t xbox_kelvinFlipStall=304ULL;
constexpr int64_t xbox_kelvinViewportOffset=2592ULL;
constexpr int64_t xbox_kelvinViewportScale=2800ULL;
constexpr int64_t xbox_vshSlotViewportScale=58ULL;
constexpr int64_t xbox_vshSlotViewportOffset=59ULL;
constexpr int64_t xbox_maxPushWords=67108864ULL;
constexpr int64_t xbox_classKelvin=151ULL;
constexpr int64_t xbox_nvSetObject=0ULL;
constexpr int64_t xbox_pramin=7340032ULL;
constexpr int64_t xbox_ramhtScan=4096ULL;
constexpr int64_t xbox_kelvinAlphaTestEnable=768ULL;
constexpr int64_t xbox_kelvinBlendEnable=772ULL;
constexpr int64_t xbox_kelvinCullFaceEnable=776ULL;
constexpr int64_t xbox_kelvinDepthTestEnable=780ULL;
constexpr int64_t xbox_kelvinAlphaFunc=828ULL;
constexpr int64_t xbox_kelvinAlphaRef=832ULL;
constexpr int64_t xbox_kelvinBlendSrcFactor=836ULL;
constexpr int64_t xbox_kelvinBlendDstFactor=840ULL;
constexpr int64_t xbox_kelvinBlendColor=844ULL;
constexpr int64_t xbox_kelvinBlendEquation=848ULL;
constexpr int64_t xbox_kelvinDepthFunc=852ULL;
constexpr int64_t xbox_kelvinColorMask=856ULL;
constexpr int64_t xbox_kelvinDepthWriteMask=860ULL;
constexpr int64_t xbox_kelvinShadeMode=892ULL;
constexpr int64_t xbox_kelvinCullFace=924ULL;
constexpr int64_t xbox_kelvinFrontFace=928ULL;
constexpr int64_t xbox_kelvinWindowClipType=692ULL;
constexpr int64_t xbox_kelvinWindowClipH=704ULL;
constexpr int64_t xbox_kelvinWindowClipV=736ULL;
constexpr int64_t xbox_rasterParallelMinArea=16384ULL;
constexpr int64_t xbox_kelvinCtxDmaTexA=388ULL;
constexpr int64_t xbox_kelvinCtxDmaTexB=392ULL;
constexpr int64_t xbox_kelvinTexOffset=6912ULL;
constexpr int64_t xbox_kelvinTexFormat=6916ULL;
constexpr int64_t xbox_kelvinTexAddress=6920ULL;
constexpr int64_t xbox_kelvinTexControl=6924ULL;
constexpr int64_t xbox_kelvinTexFilter=6932ULL;
constexpr int64_t xbox_kelvinTexRect=6940ULL;
constexpr int64_t xbox_kelvinTexCtl1=6928ULL;
constexpr int64_t xbox_texFmtSZ_A1R5G5B5=2ULL;
constexpr int64_t xbox_texFmtSZ_A4R4G4B4=4ULL;
constexpr int64_t xbox_texFmtSZ_R5G6B5=5ULL;
constexpr int64_t xbox_texFmtSZ_A8R8G8B8=6ULL;
constexpr int64_t xbox_texFmtSZ_X8R8G8B8=7ULL;
constexpr int64_t xbox_texFmtDXT1=12ULL;
constexpr int64_t xbox_texFmtDXT3=14ULL;
constexpr int64_t xbox_texFmtDXT5=15ULL;
constexpr int64_t xbox_texFmtLU_R5G6B5=17ULL;
constexpr int64_t xbox_texFmtLU_A8R8G8B8=18ULL;
constexpr int64_t xbox_texFmtSZ_A8=25ULL;
constexpr int64_t xbox_texFmtLU_X8R8G8B8=30ULL;
constexpr int64_t xbox_texFmtLU_DepthX8Y24=46ULL;
auto [xbox_nvDrawTrace,tmp319]=go_strconv_Atoi(go_os_Getenv(std::string("RR_NV_DRAW",10)));
constexpr int64_t xbox_kelvinCtxDmaVertexA=412ULL;
constexpr int64_t xbox_kelvinCtxDmaVertexB=416ULL;
constexpr int64_t xbox_kelvinVtxArrayOffset=5920ULL;
constexpr int64_t xbox_kelvinVtxArrayFormat=5984ULL;
constexpr int64_t xbox_kelvinBeginEnd=6140ULL;
constexpr int64_t xbox_kelvinElement16=6144ULL;
constexpr int64_t xbox_kelvinElement32=6152ULL;
constexpr int64_t xbox_kelvinDrawArrays=6160ULL;
constexpr int64_t xbox_kelvinInlineArray=6168ULL;
constexpr int64_t xbox_kelvinVertexData4C=6464ULL;
constexpr int64_t xbox_vtxTypeUBD3D=0ULL;
constexpr int64_t xbox_vtxTypeS1=1ULL;
constexpr int64_t xbox_vtxTypeF=2ULL;
constexpr int64_t xbox_vtxTypeUBOGL=4ULL;
constexpr int64_t xbox_vtxTypeS32K=5ULL;
constexpr int64_t xbox_vtxTypeCMP=6ULL;
constexpr int64_t xbox_primPoints=1ULL;
constexpr int64_t xbox_primLines=2ULL;
constexpr int64_t xbox_primLineLoop=3ULL;
constexpr int64_t xbox_primLineStrip=4ULL;
constexpr int64_t xbox_primTriangles=5ULL;
constexpr int64_t xbox_primTriStrip=6ULL;
constexpr int64_t xbox_primTriFan=7ULL;
constexpr int64_t xbox_primQuads=8ULL;
constexpr int64_t xbox_primQuadStrip=9ULL;
constexpr int64_t xbox_primPolygon=10ULL;
constexpr int64_t xbox_vshParallelMinVerts=256ULL;
bool xbox_nvVSTrace=(go_os_Getenv(std::string("RR_NV_VS",8)) != std::string("",0));
constexpr int64_t xbox_kelvinProgData=2816ULL;
constexpr int64_t xbox_kelvinConstData=2944ULL;
constexpr int64_t xbox_kelvinTransformExecMode=7828ULL;
constexpr int64_t xbox_kelvinCompositeMatrix=1664ULL;
constexpr int64_t xbox_kelvinLightingEnable=788ULL;
constexpr int64_t xbox_kelvinCxtWriteEnable=7832ULL;
constexpr int64_t xbox_kelvinProgLoad=7836ULL;
constexpr int64_t xbox_kelvinProgStart=7840ULL;
constexpr int64_t xbox_kelvinConstLoad=7844ULL;
constexpr int64_t xbox_vshProgSlots=136ULL;
constexpr int64_t xbox_vshConstSlots=192ULL;
constexpr int64_t xbox_macNOP=0ULL;
constexpr int64_t xbox_macMOV=1ULL;
constexpr int64_t xbox_macMUL=2ULL;
constexpr int64_t xbox_macADD=3ULL;
constexpr int64_t xbox_macMAD=4ULL;
constexpr int64_t xbox_macDP3=5ULL;
constexpr int64_t xbox_macDPH=6ULL;
constexpr int64_t xbox_macDP4=7ULL;
constexpr int64_t xbox_macDST=8ULL;
constexpr int64_t xbox_macMIN=9ULL;
constexpr int64_t xbox_macMAX=10ULL;
constexpr int64_t xbox_macSLT=11ULL;
constexpr int64_t xbox_macSGE=12ULL;
constexpr int64_t xbox_macARL=13ULL;
constexpr int64_t xbox_iluNOP=0ULL;
constexpr int64_t xbox_iluMOV=1ULL;
constexpr int64_t xbox_iluRCP=2ULL;
constexpr int64_t xbox_iluRCC=3ULL;
constexpr int64_t xbox_iluRSQ=4ULL;
constexpr int64_t xbox_iluEXP=5ULL;
constexpr int64_t xbox_iluLOG=6ULL;
constexpr int64_t xbox_iluLIT=7ULL;
std::array<std::string,14> xbox_macNames=std::array<std::string,14>{std::string("NOP",3),std::string("MOV",3),std::string("MUL",3),std::string("ADD",3),std::string("MAD",3),std::string("DP3",3),std::string("DPH",3),std::string("DP4",3),std::string("DST",3),std::string("MIN",3),std::string("MAX",3),std::string("SLT",3),std::string("SGE",3),std::string("ARL",3)};
std::array<std::string,8> xbox_iluNames=std::array<std::string,8>{std::string("NOP",3),std::string("MOV",3),std::string("RCP",3),std::string("RCC",3),std::string("RSQ",3),std::string("EXP",3),std::string("LOG",3),std::string("LIT",3)};
constexpr int64_t xbox_bucketVertex=0ULL;
constexpr int64_t xbox_bucketRaster=1ULL;
constexpr int64_t xbox_bucketClear=2ULL;
constexpr int64_t xbox_bucketPusher=3ULL;
constexpr int64_t xbox_numBuckets=4ULL;
constexpr xbox_StopReason xbox_StopBudget=0ULL;
constexpr xbox_StopReason xbox_StopHalt=1ULL;
constexpr xbox_StopReason xbox_StopFirstPush=2ULL;
constexpr xbox_StopReason xbox_StopRequest=3ULL;
constexpr xbox_StopReason xbox_StopBreak=4ULL;
constexpr int64_t xbox_schedQuantum=4000ULL;
constexpr int64_t xbox_instrsPerMs=733466ULL;
constexpr xbox_threadState xbox_tsReady=0ULL;
constexpr xbox_threadState xbox_tsRunning=1ULL;
constexpr xbox_threadState xbox_tsWaiting=2ULL;
constexpr xbox_threadState xbox_tsSleeping=3ULL;
constexpr xbox_threadState xbox_tsDead=4ULL;
constexpr int64_t xbox_threadExitAddr=2399141632ULL;
constexpr int64_t xbox_kthreadTlsData=40ULL;
constexpr int64_t xbox_hcRevision=0ULL;
constexpr int64_t xbox_hcControl=4ULL;
constexpr int64_t xbox_hcCommandStatus=8ULL;
constexpr int64_t xbox_hcInterruptStatus=12ULL;
constexpr int64_t xbox_hcInterruptEnable=16ULL;
constexpr int64_t xbox_hcInterruptDisable=20ULL;
constexpr int64_t xbox_hcHCCA=24ULL;
constexpr int64_t xbox_hcPeriodCurrentED=28ULL;
constexpr int64_t xbox_hcControlHeadED=32ULL;
constexpr int64_t xbox_hcControlCurrentED=36ULL;
constexpr int64_t xbox_hcBulkHeadED=40ULL;
constexpr int64_t xbox_hcBulkCurrentED=44ULL;
constexpr int64_t xbox_hcDoneHead=48ULL;
constexpr int64_t xbox_hcFmInterval=52ULL;
constexpr int64_t xbox_hcFmRemaining=56ULL;
constexpr int64_t xbox_hcFmNumber=60ULL;
constexpr int64_t xbox_hcPeriodicStart=64ULL;
constexpr int64_t xbox_hcLSThreshold=68ULL;
constexpr int64_t xbox_hcRhDescriptorA=72ULL;
constexpr int64_t xbox_hcRhDescriptorB=76ULL;
constexpr int64_t xbox_hcRhStatus=80ULL;
constexpr int64_t xbox_hcRhPortStatus1=84ULL;
constexpr int64_t xbox_usbPorts=4ULL;
constexpr int64_t xbox_ctrlPLE=4ULL;
constexpr int64_t xbox_ctrlIE=8ULL;
constexpr int64_t xbox_ctrlCLE=16ULL;
constexpr int64_t xbox_ctrlBLE=32ULL;
constexpr int64_t xbox_ctrlIR=256ULL;
constexpr int64_t xbox_hcfsMask=192ULL;
constexpr int64_t xbox_hcfsReset=0ULL;
constexpr int64_t xbox_hcfsResume=64ULL;
constexpr int64_t xbox_hcfsOperational=128ULL;
constexpr int64_t xbox_hcfsSuspend=192ULL;
constexpr int64_t xbox_ohciIntSO=1ULL;
constexpr int64_t xbox_ohciIntWDH=2ULL;
constexpr int64_t xbox_ohciIntSF=4ULL;
constexpr int64_t xbox_ohciIntRD=8ULL;
constexpr int64_t xbox_ohciIntUE=16ULL;
constexpr int64_t xbox_ohciIntFNO=32ULL;
constexpr int64_t xbox_ohciIntRHSC=64ULL;
constexpr int64_t xbox_ohciIntOC=1073741824ULL;
constexpr int64_t xbox_ohciIntMIE=2147483648ULL;
constexpr int64_t xbox_portCCS=1ULL;
constexpr int64_t xbox_portPES=2ULL;
constexpr int64_t xbox_portPSS=4ULL;
constexpr int64_t xbox_portPOCI=8ULL;
constexpr int64_t xbox_portPRS=16ULL;
constexpr int64_t xbox_portPPS=256ULL;
constexpr int64_t xbox_portLSDA=512ULL;
constexpr int64_t xbox_portCSC=65536ULL;
constexpr int64_t xbox_portPESC=131072ULL;
constexpr int64_t xbox_portPSSC=262144ULL;
constexpr int64_t xbox_portOCIC=524288ULL;
constexpr int64_t xbox_portPRSC=1048576ULL;
constexpr int64_t xbox_portChangeMask=2031616ULL;
constexpr int64_t xbox_portWClearEnable=1ULL;
constexpr int64_t xbox_portWSetEnable=2ULL;
constexpr int64_t xbox_portWSetSuspend=4ULL;
constexpr int64_t xbox_portWClearSuspend=8ULL;
constexpr int64_t xbox_portWSetReset=16ULL;
constexpr int64_t xbox_portWSetPower=256ULL;
constexpr int64_t xbox_portWClearPower=512ULL;
constexpr int64_t xbox_edFA=127ULL;
constexpr int64_t xbox_edEN=1920ULL;
constexpr int64_t xbox_edD=6144ULL;
constexpr int64_t xbox_edSkip=16384ULL;
constexpr int64_t xbox_edISO=32768ULL;
constexpr int64_t xbox_edMPS=134152192ULL;
constexpr int64_t xbox_edHeadHalted=1ULL;
constexpr int64_t xbox_edHeadToggle=2ULL;
constexpr uint32_t xbox_edPtrMask=4294967280ULL;
constexpr int64_t xbox_tdDP=1572864ULL;
constexpr int64_t xbox_tdDPOut=524288ULL;
constexpr int64_t xbox_tdDPIn=1048576ULL;
constexpr int64_t xbox_tdRounding=262144ULL;
constexpr int64_t xbox_tdCC=4026531840ULL;
constexpr int64_t xbox_ccNoError=0ULL;
constexpr int64_t xbox_ccStall=4ULL;
constexpr int64_t xbox_ccDataUnderrun=9ULL;
constexpr int64_t xbox_ccNotAccessed=14ULL;
constexpr xbox_usbEndpointDir xbox_dirSetup=0ULL;
constexpr xbox_usbEndpointDir xbox_dirIn=1ULL;
constexpr xbox_usbEndpointDir xbox_dirOut=2ULL;
Error xbox_errUSBUnsupported=go_fmt_Errorf(std::string("usb: unmodelled request",23));
Map<std::string,xbox_PadControl> xbox_padControls=Map<std::string,xbox_PadControl>{{std::string("up",2),xbox_PadControl{cast<xbox_PadControlKind>(0ULL),cast<uint16_t>(1ULL),{},{}}},{std::string("down",4),xbox_PadControl{cast<xbox_PadControlKind>(0ULL),cast<uint16_t>(2ULL),{},{}}},{std::string("left",4),xbox_PadControl{cast<xbox_PadControlKind>(0ULL),cast<uint16_t>(4ULL),{},{}}},{std::string("right",5),xbox_PadControl{cast<xbox_PadControlKind>(0ULL),cast<uint16_t>(8ULL),{},{}}},{std::string("start",5),xbox_PadControl{cast<xbox_PadControlKind>(0ULL),cast<uint16_t>(16ULL),{},{}}},{std::string("a",1),xbox_PadControl{cast<xbox_PadControlKind>(1ULL),{},cast<int64_t>(0ULL),{}}},{std::string("b",1),xbox_PadControl{cast<xbox_PadControlKind>(1ULL),{},cast<int64_t>(1ULL),{}}},{std::string("rtrigger",8),xbox_PadControl{cast<xbox_PadControlKind>(1ULL),{},cast<int64_t>(7ULL),{}}},{std::string("stickup",7),xbox_PadControl{cast<xbox_PadControlKind>(2ULL),{},cast<int64_t>(1ULL),cast<int64_t>(1ULL)}},{std::string("stickdown",9),xbox_PadControl{cast<xbox_PadControlKind>(2ULL),{},cast<int64_t>(1ULL),cast<int64_t>(-1ULL)}},{std::string("stickleft",9),xbox_PadControl{cast<xbox_PadControlKind>(2ULL),{},cast<int64_t>(0ULL),cast<int64_t>(-1ULL)}},{std::string("stickright",10),xbox_PadControl{cast<xbox_PadControlKind>(2ULL),{},cast<int64_t>(0ULL),cast<int64_t>(1ULL)}},{std::string("an2",3),xbox_PadControl{cast<xbox_PadControlKind>(1ULL),{},cast<int64_t>(2ULL),{}}},{std::string("an3",3),xbox_PadControl{cast<xbox_PadControlKind>(1ULL),{},cast<int64_t>(3ULL),{}}},{std::string("an4",3),xbox_PadControl{cast<xbox_PadControlKind>(1ULL),{},cast<int64_t>(4ULL),{}}},{std::string("an5",3),xbox_PadControl{cast<xbox_PadControlKind>(1ULL),{},cast<int64_t>(5ULL),{}}},{std::string("an6",3),xbox_PadControl{cast<xbox_PadControlKind>(1ULL),{},cast<int64_t>(6ULL),{}}},{std::string("an7",3),xbox_PadControl{cast<xbox_PadControlKind>(1ULL),{},cast<int64_t>(7ULL),{}}},{std::string("ax2plus",7),xbox_PadControl{cast<xbox_PadControlKind>(2ULL),{},cast<int64_t>(2ULL),cast<int64_t>(1ULL)}},{std::string("ax2minus",8),xbox_PadControl{cast<xbox_PadControlKind>(2ULL),{},cast<int64_t>(2ULL),cast<int64_t>(-1ULL)}},{std::string("ax3plus",7),xbox_PadControl{cast<xbox_PadControlKind>(2ULL),{},cast<int64_t>(3ULL),cast<int64_t>(1ULL)}},{std::string("ax3minus",8),xbox_PadControl{cast<xbox_PadControlKind>(2ULL),{},cast<int64_t>(3ULL),cast<int64_t>(-1ULL)}},{std::string("bit20",5),xbox_PadControl{cast<xbox_PadControlKind>(0ULL),cast<uint16_t>(32ULL),{},{}}},{std::string("bit40",5),xbox_PadControl{cast<xbox_PadControlKind>(0ULL),cast<uint16_t>(64ULL),{},{}}},{std::string("bit80",5),xbox_PadControl{cast<xbox_PadControlKind>(0ULL),cast<uint16_t>(128ULL),{},{}}}};
constexpr xbox_PadControlKind xbox_PadDigitalButton=0ULL;
constexpr xbox_PadControlKind xbox_PadAnalogButton=1ULL;
constexpr xbox_PadControlKind xbox_PadAxisDirection=2ULL;
constexpr int64_t xbox_PadPressed=255ULL;
constexpr int64_t xbox_PadStickFull=32767ULL;
constexpr int64_t xbox_usbReqGetStatus=0ULL;
constexpr int64_t xbox_usbReqClearFeature=1ULL;
constexpr int64_t xbox_usbReqSetFeature=3ULL;
constexpr int64_t xbox_usbReqSetAddress=5ULL;
constexpr int64_t xbox_usbReqGetDescriptor=6ULL;
constexpr int64_t xbox_usbReqSetDescriptor=7ULL;
constexpr int64_t xbox_usbReqGetConfiguration=8ULL;
constexpr int64_t xbox_usbReqSetConfiguration=9ULL;
constexpr int64_t xbox_usbReqGetInterface=10ULL;
constexpr int64_t xbox_usbReqSetInterface=11ULL;
constexpr int64_t xbox_xidReqGetReport=1ULL;
constexpr int64_t xbox_usbTypeMask=96ULL;
constexpr int64_t xbox_usbTypeStandard=0ULL;
constexpr int64_t xbox_usbTypeClass=32ULL;
constexpr int64_t xbox_usbTypeVendor=64ULL;
constexpr int64_t xbox_usbDescDevice=1ULL;
constexpr int64_t xbox_usbDescConfiguration=2ULL;
constexpr int64_t xbox_usbDescInterface=4ULL;
constexpr int64_t xbox_usbDescEndpoint=5ULL;
std::array<uint8_t,18> xbox_deviceDescriptor=std::array<uint8_t,18>{cast<uint8_t>(18ULL),cast<uint8_t>(1ULL),cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),cast<uint8_t>(229ULL),cast<uint8_t>(230ULL),cast<uint8_t>(8ULL)};
constexpr int64_t xbox_xidInterfaceClass=88ULL;
std::array<uint8_t,25> xbox_configDescriptor=std::array<uint8_t,25>{cast<uint8_t>(9ULL),cast<uint8_t>(2ULL),cast<uint8_t>(25ULL),cast<uint8_t>(0ULL),cast<uint8_t>(1ULL),cast<uint8_t>(197ULL),cast<uint8_t>(232ULL),cast<uint8_t>(233ULL),cast<uint8_t>(234ULL),cast<uint8_t>(9ULL),cast<uint8_t>(4ULL),cast<uint8_t>(210ULL),cast<uint8_t>(235ULL),cast<uint8_t>(236ULL),cast<uint8_t>(88ULL),cast<uint8_t>(237ULL),cast<uint8_t>(238ULL),cast<uint8_t>(239ULL),cast<uint8_t>(7ULL),cast<uint8_t>(5ULL),cast<uint8_t>(129ULL),cast<uint8_t>(227ULL),cast<uint8_t>(32ULL),cast<uint8_t>(240ULL),cast<uint8_t>(241ULL)};
constexpr int64_t xbox_xidDescType=66ULL;
constexpr int64_t xbox_xidReportSize=20ULL;
std::array<uint8_t,8> xbox_xidDescriptor=std::array<uint8_t,8>{cast<uint8_t>(8ULL),cast<uint8_t>(66ULL),cast<uint8_t>(242ULL),cast<uint8_t>(243ULL),cast<uint8_t>(1ULL),cast<uint8_t>(244ULL),cast<uint8_t>(20ULL),cast<uint8_t>(0ULL)};
constexpr int64_t xbox_entryKeyRetail=2835109803ULL;
constexpr int64_t xbox_entryKeyDebug=2491784523ULL;
constexpr int64_t xbox_thunkKeyRetail=1533886646ULL;
constexpr int64_t xbox_thunkKeyDebug=4021416274ULL;
const std::string xbox_xbeMagic=std::string("XBEH",4);
constexpr int64_t xbox_SecWritable=1ULL;
constexpr int64_t xbox_SecPreload=2ULL;
constexpr int64_t xbox_SecExecutable=4ULL;
constexpr int64_t xbox_SecInserted=8ULL;
constexpr int64_t xbox_SecHeadPageRO=16ULL;
constexpr int64_t xbox_SecTailPageRO=32ULL;
constexpr int64_t xbox_sectorSize=2048ULL;
constexpr int64_t xbox_volumeDescriptorSector=32ULL;
constexpr int64_t xbox_volumeDescriptorOffset=65536ULL;
const std::string xbox_xdvdfsMagic=std::string("MICROSOFT*XBOX*MEDIA",20);
constexpr int64_t xbox_attrDirectory=16ULL;
Slice<int64_t> xbox_candidateBases=Slice<int64_t>{cast<int64_t>(0ULL),cast<int64_t>(265879552ULL),cast<int64_t>(405798912ULL)};

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
// tools/platform/xbox/apu.go:110:1
xbox_mmioLatch xbox_newMMIOLatch(std::string name){
{
return xbox_mmioLatch{name,Map<uint32_t,uint32_t>{},Map<uint32_t,bool>{}};
}
}
// tools/platform/xbox/apu.go:116:1
uint8_t xbox_Machine_latchRead(xbox_Machine* m,xbox_mmioLatch* l,uint32_t off,uint32_t dw,bool written){
{
if ((((!written) && (!get(l->seenCold,shr<uint32_t>(off,cast<int64_t>(2ULL))))) && (len(l->seenCold) < cast<int64_t>(64ULL)))) {
l->seenCold[shr<uint32_t>(off,cast<int64_t>(2ULL))] = true;
xbox_Machine_logf(m,std::string("%s: unwritten register %05X read -> 0 (PC %08X)",47),l->name,(off & ~(cast<uint32_t>(3ULL))),x86_CPU_LinearPC(m->CPU));
}
if ((xbox_apuTrace && (cast<uint32_t>((off & cast<uint32_t>(3ULL))) == cast<uint32_t>(0ULL)))) {
go_fmt_Printf(std::string("%srd %05X -> %08X  PC=%08X\012",27),l->name,off,dw,x86_CPU_LinearPC(m->CPU));
}
return cast<uint8_t>(shr<uint32_t>(dw,(cast<uint32_t>((cast<uint32_t>(8ULL) * (cast<uint32_t>((off & cast<uint32_t>(3ULL)))))))));
}
}
// tools/platform/xbox/apu.go:128:1
void xbox_Machine_latchWrite(xbox_Machine* m,xbox_mmioLatch* l,uint32_t off,uint8_t v){
{
uint32_t idx = shr<uint32_t>(off,cast<int64_t>(2ULL));
uint32_t shift = cast<uint32_t>((cast<uint32_t>(8ULL) * (cast<uint32_t>((off & cast<uint32_t>(3ULL))))));
l->reg[idx] = cast<uint32_t>((((get(l->reg,idx) & ~((shl<uint32_t>(cast<uint32_t>(255ULL),shift))))) | shl<uint32_t>(cast<uint32_t>(v),shift)));
xbox_Machine_latchTrace(m,l,off,get(l->reg,idx));
}
}
// tools/platform/xbox/apu.go:148:1
void xbox_Machine_latchTrace(xbox_Machine* m,xbox_mmioLatch* l,uint32_t off,uint32_t dw){
{
if ((xbox_apuTrace && (cast<uint32_t>((off & cast<uint32_t>(3ULL))) == cast<uint32_t>(3ULL)))) {
go_fmt_Printf(std::string("%swr %05X <- %08X  PC=%08X\012",27),l->name,(off & ~(cast<uint32_t>(3ULL))),dw,x86_CPU_LinearPC(m->CPU));
}
}
}
// tools/platform/xbox/apu.go:155:1
uint8_t xbox_Machine_apuRead(xbox_Machine* m,uint32_t off){
{
auto tmp1 = lookup(m->apu.reg,shr<uint32_t>(off,cast<int64_t>(2ULL)));
uint32_t dw = std::get<0>(tmp1);
bool written = std::get<1>(tmp1);
if (((off & ~(cast<uint32_t>(3ULL))) == cast<uint32_t>(131088ULL))) {
auto tmp2 = std::make_tuple(cast<uint32_t>(shr<uint64_t>(m->tick,cast<int64_t>(10ULL))),true);
dw = std::get<0>(tmp2);
written = std::get<1>(tmp2);
}
if ((((off & ~(cast<uint32_t>(3ULL))) == cast<uint32_t>(368664ULL)) && (get(m->apu.reg,cast<uint32_t>(98303ULL)) == cast<uint32_t>(3ULL)))) {
auto tmp3 = std::make_tuple(cast<uint32_t>(13421772ULL),true);
dw = std::get<0>(tmp3);
written = std::get<1>(tmp3);
}
return xbox_Machine_latchRead(m,(&m->apu),off,dw,written);
}
}
// tools/platform/xbox/apu.go:167:1
void xbox_Machine_apuWrite(xbox_Machine* m,uint32_t off,uint8_t v){
{
xbox_Machine_latchWrite(m,(&m->apu),off,v);
}
}
// tools/platform/xbox/apu.go:182:1
uint8_t xbox_Machine_ac97Read(xbox_Machine* m,uint32_t off){
{
auto tmp4 = lookup(m->ac97.reg,shr<uint32_t>(off,cast<int64_t>(2ULL)));
uint32_t dw = std::get<0>(tmp4);
bool written = std::get<1>(tmp4);
if ((((off & ~(cast<uint32_t>(3ULL))) == cast<uint32_t>(304ULL)) && (cast<uint32_t>((get(m->ac97.reg,cast<uint32_t>(75ULL)) & cast<uint32_t>(2ULL))) != cast<uint32_t>(0ULL)))) {
auto tmp5 = std::make_tuple(cast<uint32_t>((dw | cast<uint32_t>(256ULL))),true);
dw = std::get<0>(tmp5);
written = std::get<1>(tmp5);
}
uint8_t b = xbox_Machine_latchRead(m,(&m->ac97),off,dw,written);
if ((cast<uint32_t>((off & cast<uint32_t>(15ULL))) == cast<uint32_t>(11ULL))) {
b &= ~(cast<uint8_t>(2ULL));
}
return b;
}
}
// tools/platform/xbox/apu.go:193:1
void xbox_Machine_ac97Write(xbox_Machine* m,uint32_t off,uint8_t v){
{
xbox_Machine_latchWrite(m,(&m->ac97),off,v);
}
}
// tools/platform/xbox/apu.go:200:1
uint8_t xbox_Machine_nicRead(xbox_Machine* m,uint32_t off){
{
auto tmp6 = lookup(m->nic.reg,shr<uint32_t>(off,cast<int64_t>(2ULL)));
uint32_t dw = std::get<0>(tmp6);
bool written = std::get<1>(tmp6);
return xbox_Machine_latchRead(m,(&m->nic),off,dw,written);
}
}
// tools/platform/xbox/apu.go:204:1
void xbox_Machine_nicWrite(xbox_Machine* m,uint32_t off,uint8_t v){
{
xbox_Machine_latchWrite(m,(&m->nic),off,v);
}
}
// tools/platform/xbox/apu.go:224:1
void xbox_Machine_apuTick(xbox_Machine* m){
{
if ((get(m->apu.reg,cast<uint32_t>(2048ULL)) == cast<uint32_t>(0ULL))) {
return ;
}
uint32_t base = get(m->apu.reg,cast<uint32_t>(2064ULL));
if ((base == cast<uint32_t>(0ULL))) {
return ;
}
{
uint32_t addr = cast<uint32_t>((base + cast<uint32_t>(18448ULL)));
if ((xbox_Machine_read32(m,addr) != cast<uint32_t>(0ULL))) {
xbox_Machine_write32(m,addr,cast<uint32_t>(0ULL));
}
}
}
}
// tools/platform/xbox/interrupt.go:42:1
void xbox_Machine_creditFlipVBlank(xbox_Machine* m){
{
if ((m->tick < m->nextVBlank)) {
m->tick = m->nextVBlank;
}
}
}
// tools/platform/xbox/interrupt.go:52:1
void xbox_Machine_vblankTick(xbox_Machine* m){
{
if ((m->tick >= m->nextVBlank)) {
m->nextVBlank = cast<uint64_t>((m->tick + cast<uint64_t>(12224433ULL)));
m->nv.pcrtcIntr |= cast<uint32_t>(1ULL);
}
if ((m->nv.pcrtcIntr != cast<uint32_t>(0ULL))) {
xbox_Machine_deliverPending(m);
}
}
}
// tools/platform/xbox/interrupt.go:91:1
bool xbox_Machine_irqPending(xbox_Machine* m){
{
{auto&& tmp7 = xbox_irqSources;
for(int64_t tmp8=0;tmp8<len(tmp7);++tmp8){
auto i=tmp8;if (xbox_irqSources[i].pending(m)) {
return true;
}
}}
return false;
}
}
// tools/platform/xbox/interrupt.go:109:1
void xbox_Machine_deliverPending(xbox_Machine* m){
{
if ((((m->isrActive || (!m->CPU)) || m->CPU->Halted) || (!m->CPU->IF))) {
return ;
}
if ((xbox_Machine_Read(m,cast<uint32_t>(66846756ULL)) != cast<uint8_t>(0ULL))) {
return ;
}
{auto&& tmp9 = xbox_irqSources;
for(int64_t tmp10=0;tmp10<len(tmp9);++tmp10){
auto i=tmp10;xbox_irqSource* s = (&xbox_irqSources[i]);
if ((!s->pending(m))) {
continue;
}
auto tmp11 = lookup(m->interrupts,s->vector);
uint32_t ki = std::get<0>(tmp11);
bool ok = std::get<1>(tmp11);
if ((!ok)) {
continue;
}
uint32_t routine = xbox_Machine_read32(m,cast<uint32_t>((ki + cast<uint32_t>(0ULL))));
if ((routine == cast<uint32_t>(0ULL))) {
continue;
}
m->isrSaved = (*m->CPU);
m->isrActive = true;
x86_CPU* c = m->CPU;
auto push = [&](uint32_t v)->void{
c->Regs[cast<int64_t>(4ULL)] -= cast<uint32_t>(4ULL);
xbox_Machine_write32(m,c->Regs[cast<int64_t>(4ULL)],v);
}
;
push(xbox_Machine_read32(m,cast<uint32_t>((ki + cast<uint32_t>(4ULL)))));
push(ki);
push(cast<uint32_t>(2399141376ULL));
c->IP = routine;
c->IF = false;
return ;
}}
}
}
// tools/platform/xbox/interrupt.go:156:1
void xbox_Machine_isrReturn(xbox_Machine* m){
{
if ((len(m->dpcQueue) > cast<int64_t>(0ULL))) {
xbox_dpcEntry d = m->dpcQueue[cast<int64_t>(0ULL)];
m->dpcQueue = sub(m->dpcQueue,cast<int64_t>(1ULL),len(m->dpcQueue));
uint32_t routine = xbox_Machine_read32(m,cast<uint32_t>((d.Dpc + cast<uint32_t>(12ULL))));
if ((routine != cast<uint32_t>(0ULL))) {
x86_CPU* c = m->CPU;
auto push = [&](uint32_t v)->void{
c->Regs[cast<int64_t>(4ULL)] -= cast<uint32_t>(4ULL);
xbox_Machine_write32(m,c->Regs[cast<int64_t>(4ULL)],v);
}
;
push(d.Arg2);
push(d.Arg1);
push(xbox_Machine_read32(m,cast<uint32_t>((d.Dpc + cast<uint32_t>(16ULL)))));
push(d.Dpc);
push(cast<uint32_t>(2399141376ULL));
c->IP = routine;
return ;
}
}
uint64_t steps = m->CPU->Steps;
(*m->CPU) = m->isrSaved;
m->CPU->Steps = steps;
m->isrActive = false;
}
}
// tools/platform/xbox/interrupt.go:186:1
uint32_t xbox_Machine_DebugInterruptKI(xbox_Machine* m,uint32_t vector){
{
return get(m->interrupts,vector);
}
}
// tools/platform/xbox/kernel.go:51:1
void xbox_Machine_setupKPCR(xbox_Machine* m){
{
xbox_Machine_write32(m,cast<uint32_t>(66846720ULL),cast<uint32_t>(4294967295ULL));
xbox_Machine_write32(m,cast<uint32_t>(66846724ULL),cast<uint32_t>(66842624ULL));
xbox_Machine_write32(m,cast<uint32_t>(66846728ULL),cast<uint32_t>(66777088ULL));
xbox_Machine_write32(m,cast<uint32_t>(66846748ULL),cast<uint32_t>(66846720ULL));
xbox_Machine_write32(m,cast<uint32_t>(66846752ULL),cast<uint32_t>(66846760ULL));
xbox_Machine_write32(m,cast<uint32_t>(66846756ULL),cast<uint32_t>(0ULL));
}
}
// tools/platform/xbox/kernel.go:64:1
void xbox_Machine_patchThunks(xbox_Machine* m){
{
auto tmp12 = std::make_tuple(cast<int64_t>(0ULL),cast<int64_t>(0ULL));
int64_t code = std::get<0>(tmp12);
int64_t data = std::get<1>(tmp12);
{uint32_t p = m->XBE->ThunkAddr;for (;;p += cast<uint32_t>(4ULL)){
uint32_t v = xbox_Machine_read32(m,p);
if ((v == cast<uint32_t>(0ULL))) {
break;
}
if ((cast<uint32_t>((v & cast<uint32_t>(2147483648ULL))) == cast<uint32_t>(0ULL))) {
continue;
}
uint16_t ord = cast<uint16_t>(cast<uint32_t>((v & cast<uint32_t>(65535ULL))));
{
auto tmp13 = xbox_dataExportSize(ord);
int64_t size = std::get<0>(tmp13);
bool isData = std::get<1>(tmp13);
if (isData) {
uint32_t addr = xbox_Machine_allocKObject(m,cast<uint32_t>(size));
xbox_Machine_initDataExport(m,ord,addr);
xbox_Machine_write32(m,p,addr);
data++;
}
else {
xbox_Machine_write32(m,p,cast<uint32_t>((cast<uint32_t>(2399141888ULL) + cast<uint32_t>((cast<uint32_t>(ord) * cast<uint32_t>(16ULL))))));
code++;
}
}
}
}xbox_Machine_logf(m,std::string("kernel: patched %d function + %d data import thunks",51),code,data);
}
}
// tools/platform/xbox/kernel.go:92:1
void xbox_Machine_onStep(xbox_Machine* m,x86_CPU* c){
{
uint32_t pc = cast<uint32_t>((c->SegBase[cast<int64_t>(1ULL)] + c->IP));
if (((pc >= cast<uint32_t>(2399141888ULL)) && (pc < cast<uint32_t>(2399150080ULL)))) {
uint16_t ord = cast<uint16_t>(divi<uint32_t>((cast<uint32_t>((pc - cast<uint32_t>(2399141888ULL)))),cast<uint32_t>(16ULL)));
xbox_Machine_dispatchKernel(m,ord);
return ;
}
if ((pc == cast<uint32_t>(2399141632ULL))) {
xbox_Machine_exitCurrentThread(m);
return ;
}
if ((pc == cast<uint32_t>(2399141376ULL))) {
xbox_Machine_isrReturn(m);
return ;
}
if ((m->traceLeft > cast<int64_t>(0ULL))) {
m->traceLeft--;
go_fmt_Printf(std::string("%08X  %s\012",9),pc,xbox_Machine_disasmAt(m,pc));
}
if (((m->bpLeft > cast<int64_t>(0ULL)) && (pc == m->bpAddr))) {
m->bpLeft--;
std::array<uint32_t,8> r = c->Regs;
go_fmt_Printf(std::string("bpstack %08X: EAX=%08X ECX=%08X EDX=%08X EBX=%08X ESP=%08X EBP=%08X ESI=%08X EDI=%08X\012",86),pc,r[cast<int64_t>(0ULL)],r[cast<int64_t>(1ULL)],r[cast<int64_t>(2ULL)],r[cast<int64_t>(3ULL)],r[cast<int64_t>(4ULL)],r[cast<int64_t>(5ULL)],r[cast<int64_t>(6ULL)],r[cast<int64_t>(7ULL)]);
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < cast<uint32_t>(24ULL));i++){
go_fmt_Printf(std::string("  [ESP+%02X] %08X\012",18),cast<uint32_t>((i * cast<uint32_t>(4ULL))),xbox_Machine_read32(m,cast<uint32_t>((r[cast<int64_t>(4ULL)] + cast<uint32_t>((i * cast<uint32_t>(4ULL)))))));
}
}}
if ((bool(m->hotpc) && (cast<uint64_t>((m->tick & cast<uint64_t>(255ULL))) == cast<uint64_t>(0ULL)))) {
m->hotpc[pc]++;
}
m->tick++;
xbox_Machine_schedTick(m);
}
}
// tools/platform/xbox/kernel.go:128:1
void xbox_Machine_EnableHotPC(xbox_Machine* m){
{
m->hotpc = Map<uint32_t,uint64_t>{};
}
}
// tools/platform/xbox/kernel.go:131:1
Slice<std::string> xbox_Machine_HotPCReport(xbox_Machine* m,int64_t n){
{
Slice<xbox_kernel_132_kv> all = Slice<xbox_kernel_132_kv>::make(cast<int64_t>(0ULL),len(m->hotpc));
uint64_t total={};
{auto&& tmp14 = m->hotpc;
for(auto [tmp15,tmp16]:tmp14){
auto pc=tmp15;auto c=tmp16;all = append(all,Slice<xbox_kernel_132_kv>{xbox_kernel_132_kv{pc,c}});
total += c;
}}
go_sort_Slice(all,[&](int64_t i,int64_t j)->bool{
return (all[i].c > all[j].c);
}
);
if ((len(all) > n)) {
all = sub(all,0,n);
}
Slice<std::string> out = Slice<std::string>::make(cast<int64_t>(0ULL),len(all));
{auto&& tmp17 = all;
for(int64_t tmp18=0;tmp18<len(tmp17);++tmp18){
auto e=tmp17[tmp18];out = append(out,Slice<std::string>{go_fmt_Sprintf(std::string("  %08X  %6.2f%%  (%d samples)  %s",33),e.pc,((cast<double>(e.c) * cast<double>(1.00000000000000000e+02)) / cast<double>(total)),e.c,xbox_Machine_disasmAt(m,e.pc))});
}}
return out;
}
}
// tools/platform/xbox/kernel.go:160:1
void xbox_Machine_dispatchKernel(xbox_Machine* m,uint16_t ord){
{
m->OrdinalHits[ord]++;
std::string name = xbox_ordinalName(ord);
if (xbox_pumpTrace) {
{
switch(ord){
case cast<uint16_t>(145ULL):case cast<uint16_t>(151ULL):case cast<uint16_t>(222ULL):case cast<uint16_t>(224ULL):case cast<uint16_t>(225ULL):case cast<uint16_t>(231ULL):case cast<uint16_t>(234ULL):case cast<uint16_t>(159ULL):{
go_fmt_Printf(std::string("PUMP tid=%d %s(%08X) from=%08X tick=%d\012",39),xbox_Machine_threadID(m),name,xbox_Machine_arg(m,cast<int64_t>(0ULL)),xbox_Machine_retAddr(m),m->tick);
break;}
}}
}
std::function<int64_t(xbox_Machine*)> h = xbox_kernelHandler(ord);
if ((!h)) {
x86_CPU_Halt(m->CPU,std::string("unimplemented xboxkrnl ordinal %d (%s), called from %08X",56),ord,name,xbox_Machine_retAddr(m));
auto tmp19 = std::make_tuple(true,m->CPU->HaltReason);
m->Halted = std::get<0>(tmp19);
m->HaltReason = std::get<1>(tmp19);
return ;
}
int64_t argWords = h(m);
if ((argWords == cast<int64_t>(-1ULL))) {
return ;
}
xbox_Machine_kret(m,argWords);
if (m->reschedule) {
m->reschedule = false;
xbox_Machine_dispatch(m);
}
}
}
// tools/platform/xbox/kernel.go:200:1
uint32_t xbox_Machine_retAddr(xbox_Machine* m){
{
return xbox_Machine_read32(m,m->CPU->Regs[cast<int64_t>(4ULL)]);
}
}
// tools/platform/xbox/kernel.go:203:1
uint32_t xbox_Machine_arg(xbox_Machine* m,int64_t i){
{
return xbox_Machine_read32(m,cast<uint32_t>((cast<uint32_t>((m->CPU->Regs[cast<int64_t>(4ULL)] + cast<uint32_t>(4ULL))) + cast<uint32_t>((cast<uint32_t>(i) * cast<uint32_t>(4ULL))))));
}
}
// tools/platform/xbox/kernel.go:206:1
void xbox_Machine_setRet(xbox_Machine* m,uint32_t v){
{
m->CPU->Regs[cast<int64_t>(0ULL)] = v;
}
}
// tools/platform/xbox/kernel.go:211:1
void xbox_Machine_kret(xbox_Machine* m,int64_t argWords){
{
uint32_t sp = m->CPU->Regs[cast<int64_t>(4ULL)];
uint32_t ret = xbox_Machine_read32(m,sp);
m->CPU->Regs[cast<int64_t>(4ULL)] = cast<uint32_t>((cast<uint32_t>((sp + cast<uint32_t>(4ULL))) + cast<uint32_t>((cast<uint32_t>(argWords) * cast<uint32_t>(4ULL)))));
m->CPU->IP = ret;
}
}
// tools/platform/xbox/kernel.go:222:1
uint32_t xbox_Machine_allocPool(xbox_Machine* m,uint32_t size){
{
return xbox_Machine_allocPoolAligned(m,size,cast<uint32_t>(16ULL));
}
}
// tools/platform/xbox/kernel.go:228:1
uint32_t xbox_Machine_allocPoolAligned(xbox_Machine* m,uint32_t size,uint32_t align){
{
if ((align < cast<uint32_t>(16ULL))) {
align = cast<uint32_t>(16ULL);
}
size = xbox_align32(size,cast<uint32_t>(16ULL));
if (((size == cast<uint32_t>(0ULL)) || (size > m->poolNext))) {
return cast<uint32_t>(0ULL);
}
uint32_t base = ((cast<uint32_t>((m->poolNext - size))) & ~((cast<uint32_t>((align - cast<uint32_t>(1ULL))))));
if ((base < m->heapNext)) {
return cast<uint32_t>(0ULL);
}
m->poolNext = base;
m->poolSizes[base] = size;
xbox_Machine_logf(m,std::string("allocPool: %X bytes (align %X) -> %08X (pool free %d KiB)",57),size,align,base,divi<uint32_t>((cast<uint32_t>((base - m->heapNext))),cast<uint32_t>(1024ULL)));
return base;
}
}
// tools/platform/xbox/kernel.go:249:1
uint32_t xbox_Machine_allocVirtual(xbox_Machine* m,uint32_t size){
{
size = xbox_align32(size,cast<uint32_t>(4096ULL));
if (((size == cast<uint32_t>(0ULL)) || (cast<uint32_t>((m->heapNext + size)) > m->heapTop))) {
return cast<uint32_t>(0ULL);
}
uint32_t a = m->heapNext;
m->heapNext += size;
return a;
}
}
// tools/platform/xbox/kernel.go:261:1
uint32_t xbox_Machine_allocKObject(xbox_Machine* m,uint32_t size){
{
size = xbox_align32(size,cast<uint32_t>(16ULL));
uint32_t a = m->nextObjAddr;
m->nextObjAddr += size;
return a;
}
}
// tools/platform/xbox/kernel.go:278:1
std::function<int64_t(xbox_Machine*)> xbox_kernelHandler(uint16_t ord){
{
{
std::function<int64_t(xbox_Machine*)> h = xbox_kernelObjectHandler(ord);
if (bool(h)) {
return h;
}
}
{
std::function<int64_t(xbox_Machine*)> h = xbox_kernelCryptoHandler(ord);
if (bool(h)) {
return h;
}
}
{
switch(ord){
case cast<uint16_t>(1ULL):{
return [=](xbox_Machine* m)->int64_t{
xbox_Machine_setRet(m,cast<uint32_t>(0ULL));
return cast<int64_t>(0ULL);
}
;
break;}
case cast<uint16_t>(3ULL):{
return [=](xbox_Machine* m)->int64_t{
m->nv.dispMode = xbox_Machine_arg(m,cast<int64_t>(2ULL));
m->nv.dispFormat = xbox_Machine_arg(m,cast<int64_t>(3ULL));
m->nv.fbPitch = xbox_Machine_arg(m,cast<int64_t>(4ULL));
m->nv.fbAddr = xbox_Machine_arg(m,cast<int64_t>(5ULL));
xbox_Machine_logf(m,std::string("AvSetDisplayMode: mode %X format %X pitch %d fb %08X",52),m->nv.dispMode,m->nv.dispFormat,m->nv.fbPitch,m->nv.fbAddr);
xbox_Machine_setRet(m,cast<uint32_t>(0ULL));
return cast<int64_t>(6ULL);
}
;
break;}
case cast<uint16_t>(2ULL):{
return [=](xbox_Machine* m)->int64_t{
{
uint32_t p = xbox_Machine_arg(m,cast<int64_t>(3ULL));
if ((p != cast<uint32_t>(0ULL))) {
xbox_Machine_write32(m,p,cast<uint32_t>(0ULL));
}
}
xbox_Machine_setRet(m,cast<uint32_t>(0ULL));
return cast<int64_t>(4ULL);
}
;
break;}
case cast<uint16_t>(15ULL):{
return [=](xbox_Machine* m)->int64_t{
xbox_Machine_setRet(m,xbox_Machine_allocPool(m,xbox_Machine_arg(m,cast<int64_t>(0ULL))));
return cast<int64_t>(2ULL);
}
;
break;}
case cast<uint16_t>(151ULL):{
return [=](xbox_Machine* m)->int64_t{
uint32_t us = xbox_Machine_arg(m,cast<int64_t>(0ULL));
m->tick += divi<uint64_t>(cast<uint64_t>((cast<uint64_t>(us) * cast<uint64_t>(733466ULL))),cast<uint64_t>(1000ULL));
xbox_Machine_setRet(m,cast<uint32_t>(0ULL));
return cast<int64_t>(1ULL);
}
;
break;}
case cast<uint16_t>(129ULL):{
return [=](xbox_Machine* m)->int64_t{
uint8_t old = xbox_Machine_Read(m,cast<uint32_t>(66846756ULL));
xbox_Machine_Write(m,cast<uint32_t>(66846756ULL),cast<uint8_t>(2ULL));
xbox_Machine_setRet(m,cast<uint32_t>(old));
return cast<int64_t>(0ULL);
}
;
break;}
case cast<uint16_t>(160ULL):{
return [=](xbox_Machine* m)->int64_t{
uint8_t old = xbox_Machine_Read(m,cast<uint32_t>(66846756ULL));
xbox_Machine_Write(m,cast<uint32_t>(66846756ULL),cast<uint8_t>(m->CPU->Regs[cast<int64_t>(1ULL)]));
xbox_Machine_setRet(m,cast<uint32_t>(old));
return cast<int64_t>(0ULL);
}
;
break;}
case cast<uint16_t>(161ULL):{
return [=](xbox_Machine* m)->int64_t{
xbox_Machine_Write(m,cast<uint32_t>(66846756ULL),cast<uint8_t>(m->CPU->Regs[cast<int64_t>(1ULL)]));
xbox_Machine_setRet(m,cast<uint32_t>(0ULL));
return cast<int64_t>(0ULL);
}
;
break;}
case cast<uint16_t>(23ULL):{
return [=](xbox_Machine* m)->int64_t{
xbox_Machine_setRet(m,get(m->poolSizes,xbox_Machine_arg(m,cast<int64_t>(0ULL))));
return cast<int64_t>(1ULL);
}
;
break;}
case cast<uint16_t>(8ULL):{
return [=](xbox_Machine* m)->int64_t{
xbox_Machine_logf(m,std::string("DbgPrint: %s",12),xbox_Machine_cstr(m,xbox_Machine_arg(m,cast<int64_t>(0ULL))));
xbox_Machine_setRet(m,cast<uint32_t>(0ULL));
return cast<int64_t>(0ULL);
}
;
break;}
case cast<uint16_t>(24ULL):{
return [=](xbox_Machine* m)->int64_t{
{
uint32_t p = xbox_Machine_arg(m,cast<int64_t>(1ULL));
if ((p != cast<uint32_t>(0ULL))) {
xbox_Machine_write32(m,p,cast<uint32_t>(4ULL));
}
}
{
uint32_t p = xbox_Machine_arg(m,cast<int64_t>(2ULL));
if ((p != cast<uint32_t>(0ULL))) {
xbox_Machine_write32(m,p,cast<uint32_t>(0ULL));
}
}
{
uint32_t p = xbox_Machine_arg(m,cast<int64_t>(4ULL));
if ((p != cast<uint32_t>(0ULL))) {
xbox_Machine_write32(m,p,cast<uint32_t>(4ULL));
}
}
xbox_Machine_setRet(m,cast<uint32_t>(0ULL));
return cast<int64_t>(5ULL);
}
;
break;}
case cast<uint16_t>(37ULL):{
return [=](xbox_Machine* m)->int64_t{
xbox_Machine_setRet(m,cast<uint32_t>(0ULL));
return cast<int64_t>(1ULL);
}
;
break;}
case cast<uint16_t>(47ULL):{
return [=](xbox_Machine* m)->int64_t{
xbox_Machine_setRet(m,cast<uint32_t>(0ULL));
return cast<int64_t>(2ULL);
}
;
break;}
case cast<uint16_t>(289ULL):{
return [=](xbox_Machine* m)->int64_t{
auto tmp20 = std::make_tuple(xbox_Machine_arg(m,cast<int64_t>(0ULL)),xbox_Machine_arg(m,cast<int64_t>(1ULL)));
uint32_t dst = std::get<0>(tmp20);
uint32_t src = std::get<1>(tmp20);
if ((dst != cast<uint32_t>(0ULL))) {
uint32_t n = cast<uint32_t>(len(xbox_Machine_cstr(m,src)));
xbox_Machine_write16(m,cast<uint32_t>((dst + cast<uint32_t>(0ULL))),cast<uint16_t>(n));
xbox_Machine_write16(m,cast<uint32_t>((dst + cast<uint32_t>(2ULL))),cast<uint16_t>(cast<uint32_t>((n + cast<uint32_t>(1ULL)))));
xbox_Machine_write32(m,cast<uint32_t>((dst + cast<uint32_t>(4ULL))),src);
}
xbox_Machine_setRet(m,cast<uint32_t>(0ULL));
return cast<int64_t>(2ULL);
}
;
break;}
case cast<uint16_t>(252ULL):{
return [=](xbox_Machine* m)->int64_t{
xbox_Machine_setRet(m,cast<uint32_t>(0ULL));
return cast<int64_t>(1ULL);
}
;
break;}
case cast<uint16_t>(253ULL):{
return [=](xbox_Machine* m)->int64_t{
xbox_Machine_setRet(m,cast<uint32_t>(0ULL));
return cast<int64_t>(2ULL);
}
;
break;}
case cast<uint16_t>(255ULL):{
return [=](xbox_Machine* m)->int64_t{
uint32_t handleOut = xbox_Machine_arg(m,cast<int64_t>(0ULL));
uint32_t stackSize = xbox_Machine_arg(m,cast<int64_t>(2ULL));
uint32_t tlsSize = xbox_Machine_arg(m,cast<int64_t>(3ULL));
uint32_t threadIDOut = xbox_Machine_arg(m,cast<int64_t>(4ULL));
uint32_t ctx1 = xbox_Machine_arg(m,cast<int64_t>(5ULL));
uint32_t ctx2 = xbox_Machine_arg(m,cast<int64_t>(6ULL));
bool suspended = (xbox_Machine_arg(m,cast<int64_t>(7ULL)) != cast<uint32_t>(0ULL));
uint32_t startRoutine = xbox_Machine_arg(m,cast<int64_t>(9ULL));
xbox_thread* t = xbox_Machine_createThread(m,startRoutine,ctx1,ctx2,stackSize,tlsSize,cast<int32_t>(16ULL),suspended);
if ((handleOut != cast<uint32_t>(0ULL))) {
xbox_Machine_write32(m,handleOut,t->kthread);
}
if ((threadIDOut != cast<uint32_t>(0ULL))) {
xbox_Machine_write32(m,threadIDOut,t->id);
}
xbox_Machine_setRet(m,cast<uint32_t>(0ULL));
return cast<int64_t>(10ULL);
}
;
break;}
case cast<uint16_t>(258ULL):{
return [=](xbox_Machine* m)->int64_t{
xbox_Machine_logf(m,std::string("PsTerminateSystemThread: thread %d exit status %08X (tick %d)",61),xbox_Machine_threadID(m),xbox_Machine_arg(m,cast<int64_t>(0ULL)),m->tick);
xbox_Machine_exitCurrentThread(m);
return cast<int64_t>(-1ULL);
}
;
break;}
}}
return {};
}
}
// tools/platform/xbox/kernel_crypto.go:49:1
Slice<uint8_t> xbox_Machine_readBytes(xbox_Machine* m,uint32_t addr,uint32_t n){
{
Slice<uint8_t> b = Slice<uint8_t>::make(n);
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < n);i++){
b[i] = xbox_Machine_Read(m,cast<uint32_t>((addr + i)));
}
}return b;
}
}
// tools/platform/xbox/kernel_crypto.go:58:1
void xbox_Machine_writeBytes(xbox_Machine* m,uint32_t addr,Slice<uint8_t> b){
{
{auto&& tmp21 = b;
for(int64_t tmp22=0;tmp22<len(tmp21);++tmp22){
auto i=tmp22;auto v=tmp21[tmp22];xbox_Machine_Write(m,cast<uint32_t>((addr + cast<uint32_t>(i))),v);
}}
}
}
// tools/platform/xbox/kernel_crypto.go:67:1
std::function<int64_t(xbox_Machine*)> xbox_kernelCryptoHandler(uint16_t ord){
{
{
switch(ord){
case cast<uint16_t>(335ULL):{
return [=](xbox_Machine* m)->int64_t{
xbox_Machine_shaStore(m,xbox_Machine_arg(m,cast<int64_t>(0ULL)),go_sha1_New());
xbox_Machine_setRet(m,cast<uint32_t>(0ULL));
return cast<int64_t>(1ULL);
}
;
break;}
case cast<uint16_t>(336ULL):{
return [=](xbox_Machine* m)->int64_t{
auto tmp23 = std::make_tuple(xbox_Machine_arg(m,cast<int64_t>(0ULL)),xbox_Machine_arg(m,cast<int64_t>(1ULL)),xbox_Machine_arg(m,cast<int64_t>(2ULL)));
uint32_t ctx = std::get<0>(tmp23);
uint32_t data = std::get<1>(tmp23);
uint32_t n = std::get<2>(tmp23);
SHA1* h = xbox_Machine_shaLoad(m,ctx);
hash_Write(h,xbox_Machine_readBytes(m,data,n));
xbox_Machine_shaStore(m,ctx,h);
xbox_Machine_setRet(m,cast<uint32_t>(0ULL));
return cast<int64_t>(3ULL);
}
;
break;}
case cast<uint16_t>(337ULL):{
return [=](xbox_Machine* m)->int64_t{
auto tmp24 = std::make_tuple(xbox_Machine_arg(m,cast<int64_t>(0ULL)),xbox_Machine_arg(m,cast<int64_t>(1ULL)));
uint32_t ctx = std::get<0>(tmp24);
uint32_t out = std::get<1>(tmp24);
SHA1* h = xbox_Machine_shaLoad(m,ctx);
xbox_Machine_writeBytes(m,out,sub(hash_Sum(h,{}),0,cast<int64_t>(20ULL)));
removeKey(m->shaCtx,ctx);
xbox_Machine_setRet(m,cast<uint32_t>(0ULL));
return cast<int64_t>(2ULL);
}
;
break;}
case cast<uint16_t>(338ULL):{
return [=](xbox_Machine* m)->int64_t{
auto tmp25 = std::make_tuple(xbox_Machine_arg(m,cast<int64_t>(0ULL)),xbox_Machine_arg(m,cast<int64_t>(1ULL)),xbox_Machine_arg(m,cast<int64_t>(2ULL)));
uint32_t keyTable = std::get<0>(tmp25);
uint32_t cbKey = std::get<1>(tmp25);
uint32_t pbKey = std::get<2>(tmp25);
Slice<uint8_t> key = xbox_Machine_readBytes(m,pbKey,cbKey);
std::array<uint8_t,256> s={};
{auto&& tmp26 = s;
for(int64_t tmp27=0;tmp27<len(tmp26);++tmp27){
auto i=tmp27;s[i] = cast<uint8_t>(i);
}}
if ((len(key) > cast<int64_t>(0ULL))) {
int64_t j = cast<int64_t>(0ULL);
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(256ULL));i++){
j = cast<int64_t>(((cast<int64_t>((cast<int64_t>((j + cast<int64_t>(s[i]))) + cast<int64_t>(key[modi<int64_t>(i,len(key))])))) & cast<int64_t>(255ULL)));
auto tmp28 = std::make_tuple(s[j],s[i]);
s[i] = std::get<0>(tmp28);
s[j] = std::get<1>(tmp28);
}
}}
Slice<uint8_t> state = Slice<uint8_t>::make(cast<int64_t>(258ULL));
gcopy(state,sub(s,0,len(s)));
m->rc4Ctx[keyTable] = state;
xbox_Machine_setRet(m,cast<uint32_t>(0ULL));
return cast<int64_t>(3ULL);
}
;
break;}
case cast<uint16_t>(339ULL):{
return [=](xbox_Machine* m)->int64_t{
auto tmp29 = std::make_tuple(xbox_Machine_arg(m,cast<int64_t>(0ULL)),xbox_Machine_arg(m,cast<int64_t>(1ULL)),xbox_Machine_arg(m,cast<int64_t>(2ULL)));
uint32_t keyTable = std::get<0>(tmp29);
uint32_t n = std::get<1>(tmp29);
uint32_t data = std::get<2>(tmp29);
auto tmp30 = lookup(m->rc4Ctx,keyTable);
Slice<uint8_t> state = std::get<0>(tmp30);
bool ok = std::get<1>(tmp30);
if (((!ok) || (len(state) < cast<int64_t>(258ULL)))) {
state = Slice<uint8_t>::make(cast<int64_t>(258ULL));
{auto&& tmp31 = sub(state,0,cast<int64_t>(256ULL));
for(int64_t tmp32=0;tmp32<len(tmp31);++tmp32){
auto i=tmp32;state[i] = cast<uint8_t>(i);
}}
}
std::array<uint8_t,256> s={};
gcopy(sub(s,0,len(s)),sub(state,0,cast<int64_t>(256ULL)));
auto tmp33 = std::make_tuple(state[cast<int64_t>(256ULL)],state[cast<int64_t>(257ULL)]);
uint8_t i = std::get<0>(tmp33);
uint8_t j = std::get<1>(tmp33);
{uint32_t k = cast<uint32_t>(0ULL);for (;(k < n);k++){
i++;
j += s[i];
auto tmp34 = std::make_tuple(s[j],s[i]);
s[i] = std::get<0>(tmp34);
s[j] = std::get<1>(tmp34);
uint8_t ks = s[cast<uint8_t>(cast<uint8_t>((s[i] + s[j])))];
xbox_Machine_Write(m,cast<uint32_t>((data + k)),cast<uint8_t>((xbox_Machine_Read(m,cast<uint32_t>((data + k))) ^ ks)));
}
}gcopy(sub(state,0,cast<int64_t>(256ULL)),sub(s,0,len(s)));
auto tmp35 = std::make_tuple(i,j);
state[cast<int64_t>(256ULL)] = std::get<0>(tmp35);
state[cast<int64_t>(257ULL)] = std::get<1>(tmp35);
m->rc4Ctx[keyTable] = state;
xbox_Machine_setRet(m,cast<uint32_t>(0ULL));
return cast<int64_t>(3ULL);
}
;
break;}
case cast<uint16_t>(346ULL):{
return [=](xbox_Machine* m)->int64_t{
auto tmp36 = std::make_tuple(xbox_Machine_arg(m,cast<int64_t>(0ULL)),xbox_Machine_arg(m,cast<int64_t>(1ULL)));
uint32_t key = std::get<0>(tmp36);
uint32_t n = std::get<1>(tmp36);
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < n);i++){
uint8_t b = xbox_Machine_Read(m,cast<uint32_t>((key + i)));
b = cast<uint8_t>((cast<uint8_t>((cast<uint8_t>((b & cast<uint8_t>(254ULL))) | (cast<uint8_t>((xbox_popcount7(shr<uint8_t>(b,cast<int64_t>(1ULL))) & cast<uint8_t>(1ULL)))))) ^ cast<uint8_t>(1ULL)));
xbox_Machine_Write(m,cast<uint32_t>((key + i)),b);
}
}xbox_Machine_setRet(m,cast<uint32_t>(0ULL));
return cast<int64_t>(2ULL);
}
;
break;}
case cast<uint16_t>(340ULL):{
return [=](xbox_Machine* m)->int64_t{
auto tmp37 = std::make_tuple(xbox_Machine_arg(m,cast<int64_t>(0ULL)),xbox_Machine_arg(m,cast<int64_t>(1ULL)));
uint32_t pbKey = std::get<0>(tmp37);
uint32_t cbKey = std::get<1>(tmp37);
auto tmp38 = std::make_tuple(xbox_Machine_arg(m,cast<int64_t>(2ULL)),xbox_Machine_arg(m,cast<int64_t>(3ULL)));
uint32_t pbData = std::get<0>(tmp38);
uint32_t cbData = std::get<1>(tmp38);
auto tmp39 = std::make_tuple(xbox_Machine_arg(m,cast<int64_t>(4ULL)),xbox_Machine_arg(m,cast<int64_t>(5ULL)));
uint32_t pbData2 = std::get<0>(tmp39);
uint32_t cbData2 = std::get<1>(tmp39);
uint32_t pbDigest = xbox_Machine_arg(m,cast<int64_t>(6ULL));
if ((cbKey > cast<uint32_t>(64ULL))) {
cbKey = cast<uint32_t>(64ULL);
}
Slice<uint8_t> key = xbox_Machine_readBytes(m,pbKey,cbKey);
std::array<uint8_t,64> ipad={};
std::array<uint8_t,64> opad={};
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(64ULL));i++){
uint8_t k={};
if ((i < len(key))) {
k = key[i];
}
ipad[i] = cast<uint8_t>((k ^ cast<uint8_t>(54ULL)));
opad[i] = cast<uint8_t>((k ^ cast<uint8_t>(92ULL)));
}
}SHA1* inner = go_sha1_New();
hash_Write(inner,sub(ipad,0,len(ipad)));
hash_Write(inner,xbox_Machine_readBytes(m,pbData,cbData));
hash_Write(inner,xbox_Machine_readBytes(m,pbData2,cbData2));
Slice<uint8_t> innerDigest = hash_Sum(inner,{});
SHA1* outer = go_sha1_New();
hash_Write(outer,sub(opad,0,len(opad)));
hash_Write(outer,innerDigest);
Slice<uint8_t> digest = hash_Sum(outer,{});
xbox_Machine_writeBytes(m,pbDigest,sub(digest,0,cast<int64_t>(20ULL)));
xbox_Machine_setRet(m,cast<uint32_t>(0ULL));
return cast<int64_t>(7ULL);
}
;
break;}
}}
return {};
}
}
// tools/platform/xbox/kernel_crypto.go:220:1
uint8_t xbox_popcount7(uint8_t b){
{
uint8_t n={};
{;for (;(b != cast<uint8_t>(0ULL));b = shr<uint8_t>(b,cast<int64_t>(1ULL))){
n += cast<uint8_t>((b & cast<uint8_t>(1ULL)));
}
}return n;
}
}
// tools/platform/xbox/kernel_data.go:28:1
std::tuple<int64_t,bool> xbox_dataExportSize(uint16_t ord){
{
{
switch(ord){
case cast<uint16_t>(154ULL):{
return {cast<int64_t>(16ULL),true};
break;}
case cast<uint16_t>(156ULL):{
return {cast<int64_t>(8ULL),true};
break;}
case cast<uint16_t>(357ULL):{
return {cast<int64_t>(64ULL),true};
break;}
}}
return {cast<int64_t>(0ULL),false};
}
}
// tools/platform/xbox/kernel_data.go:54:1
void xbox_Machine_initDataExport(xbox_Machine* m,uint16_t ord,uint32_t addr){
{
{
switch(ord){
case cast<uint16_t>(84ULL):{
xbox_Machine_Write(m,addr,cast<uint8_t>(0ULL));
break;}
case cast<uint16_t>(85ULL):{
xbox_Machine_Write(m,addr,cast<uint8_t>(1ULL));
break;}
case cast<uint16_t>(152ULL):{
xbox_Machine_write32(m,addr,cast<uint32_t>(10000ULL));
break;}
case cast<uint16_t>(323ULL):{
xbox_Machine_write16(m,cast<uint32_t>((addr + cast<uint32_t>(0ULL))),cast<uint16_t>(1ULL));
xbox_Machine_write16(m,cast<uint32_t>((addr + cast<uint32_t>(2ULL))),cast<uint16_t>(0ULL));
xbox_Machine_write16(m,cast<uint32_t>((addr + cast<uint32_t>(4ULL))),cast<uint16_t>(5838ULL));
xbox_Machine_write16(m,cast<uint32_t>((addr + cast<uint32_t>(6ULL))),cast<uint16_t>(1ULL));
break;}
case cast<uint16_t>(320ULL):case cast<uint16_t>(321ULL):case cast<uint16_t>(322ULL):case cast<uint16_t>(324ULL):case cast<uint16_t>(325ULL):{
break;}
case cast<uint16_t>(357ULL):{
xbox_Machine_write32(m,cast<uint32_t>((addr + cast<uint32_t>(40ULL))),cast<uint32_t>((addr + cast<uint32_t>(40ULL))));
xbox_Machine_write32(m,cast<uint32_t>((addr + cast<uint32_t>(44ULL))),cast<uint32_t>((addr + cast<uint32_t>(40ULL))));
break;}
}}
if ((ord == cast<uint16_t>(156ULL))) {
m->tickCountAddr = addr;
}
if ((ord == cast<uint16_t>(154ULL))) {
m->systemTimeAddr = addr;
}
}
}
// tools/platform/xbox/kernel_file.go:54:1
bool xbox_fileObject_isDir(xbox_fileObject* fo){
{
return (fo->dir || fo->entry.IsDir);
}
}
// tools/platform/xbox/kernel_file.go:59:1
uint64_t xbox_fnv64(std::string s){
{
uint64_t h = cast<uint64_t>(14695981039346656037ULL);
{int64_t i = cast<int64_t>(0ULL);for (;(i < len(s));i++){
h ^= cast<uint64_t>(cast<uint8_t>(s[i]));
h *= cast<uint64_t>(1099511628211ULL);
}
}return h;
}
}
// tools/platform/xbox/kernel_file.go:69:1
uint32_t xbox_fileObject_size(xbox_fileObject* fo){
{
if (bool(fo->cache)) {
return cast<uint32_t>(len(fo->cache->Data));
}
return fo->entry.Size;
}
}
// tools/platform/xbox/kernel_file.go:78:1
std::tuple<std::string,bool> xbox_resolveDiscPath(std::string p){
{
p = go_strings_ReplaceAll(p,std::string("\\",1),std::string("/",1));
{auto&& tmp40 = Slice<std::string>{std::string("/Device/CdRom0",14),std::string("/??/D:",6),std::string("/??/d:",6),std::string("D:",2),std::string("d:",2)};
for(int64_t tmp41=0;tmp41<len(tmp40);++tmp41){
auto pre=tmp40[tmp41];if (((len(p) >= len(pre)) && go_strings_EqualFold(sub(p,0,len(pre)),pre))) {
std::string rest = sub(p,len(pre),len(p));
if ((rest == std::string("",0))) {
rest = std::string("/",1);
}
if ((!go_strings_HasPrefix(rest,std::string("/",1)))) {
rest = (std::string("/",1) + rest);
}
return {rest,true};
}
}}
return {std::string("",0),false};
}
}
// tools/platform/xbox/kernel_file.go:98:1
std::tuple<std::string,bool> xbox_resolveCachePath(std::string p){
{
p = go_strings_ReplaceAll(p,std::string("\\",1),std::string("/",1));
{auto&& tmp42 = Slice<std::string>{std::string("/??/T:",6),std::string("/??/U:",6),std::string("/??/Z:",6),std::string("T:",2),std::string("U:",2),std::string("Z:",2)};
for(int64_t tmp43=0;tmp43<len(tmp42);++tmp43){
auto pre=tmp42[tmp43];if (((len(p) >= len(pre)) && go_strings_EqualFold(sub(p,0,len(pre)),pre))) {
std::string drive = sub(pre,cast<int64_t>((len(pre) - cast<int64_t>(2ULL))),len(pre));
std::string rest = sub(p,len(pre),len(p));
if ((!go_strings_HasPrefix(rest,std::string("/",1)))) {
rest = (std::string("/",1) + rest);
}
return {go_strings_ToUpper((drive + rest)),true};
}
}}
return {std::string("",0),false};
}
}
// tools/platform/xbox/kernel_file.go:168:1
std::tuple<std::string,bool> xbox_rawDevicePath(std::string p){
{
std::string q = go_strings_ToLower(go_strings_ReplaceAll(p,std::string("\\",1),std::string("/",1)));
q = go_strings_TrimSuffix(q,std::string("/",1));
if ((q == std::string("/device/harddisk0/partition0",28))) {
return {xbox_rawPartition0Key,true};
}
return {std::string("",0),false};
}
}
// tools/platform/xbox/kernel_file.go:212:1
xbox_cacheFile* xbox_Machine_rawDevice(xbox_Machine* m,std::string key){
{
{
xbox_cacheFile* cf = get(m->cacheFS,key);
if (bool(cf)) {
return cf;
}
}
xbox_cacheFile* cf = arenaNew(xbox_cacheFile{Slice<uint8_t>::make(cast<int64_t>(65536ULL))});
m->cacheFS[key] = cf;
return cf;
}
}
// tools/platform/xbox/kernel_file.go:231:1
std::tuple<uint32_t,bool,bool> xbox_Machine_statPath(xbox_Machine* m,std::string path){
uint32_t size{};
bool isDir{};
bool found{};
{
{
auto tmp44 = xbox_resolveCachePath(path);
std::string key = std::get<0>(tmp44);
bool ok = std::get<1>(tmp44);
if (ok) {
{
xbox_cacheFile* cf = get(m->cacheFS,key);
if (bool(cf)) {
return {cast<uint32_t>(len(cf->Data)),false,true};
}
}
std::string dir = go_strings_TrimSuffix(key,std::string("/",1));
if ((len(dir) == cast<int64_t>(2ULL))) {
return {cast<uint32_t>(0ULL),true,true};
}
{auto&& tmp45 = m->cacheFS;
for(auto [tmp46,tmp47]:tmp45){
auto k=tmp46;if (go_strings_HasPrefix(k,(dir + std::string("/",1)))) {
return {cast<uint32_t>(0ULL),true,true};
}
}}
return {cast<uint32_t>(0ULL),false,false};
}
}
auto tmp48 = xbox_resolveDiscPath(path);
std::string disc = std::get<0>(tmp48);
bool ok = std::get<1>(tmp48);
if (((!ok) || (!m->Disc))) {
return {cast<uint32_t>(0ULL),false,false};
}
auto tmp49 = xbox_Image_resolve(m->Disc,disc);
xbox_Entry e = std::get<0>(tmp49);
Error err = std::get<1>(tmp49);
if (bool(err)) {
return {cast<uint32_t>(0ULL),false,false};
}
return {e.Size,e.IsDir,true};
}
}
// tools/platform/xbox/kernel_file.go:270:1
Slice<xbox_dirEntry> xbox_Machine_listDir(xbox_Machine* m,xbox_fileObject* fo){
{
if (fo->dir) {
std::string prefix = (go_strings_TrimSuffix(fo->key,std::string("/",1)) + std::string("/",1));
Map<std::string,xbox_dirEntry> byName = Map<std::string,xbox_dirEntry>{};
{auto&& tmp50 = m->cacheFS;
for(auto [tmp51,tmp52]:tmp50){
auto k=tmp51;auto cf=tmp52;if ((!go_strings_HasPrefix(k,prefix))) {
continue;
}
std::string rest = sub(k,len(prefix),len(k));
{
int64_t i = go_strings_Index(rest,std::string("/",1));
if ((i >= cast<int64_t>(0ULL))) {
byName[sub(rest,0,i)] = xbox_dirEntry{sub(rest,0,i),{},true};
}
else if ((rest != std::string("",0))) {
byName[rest] = xbox_dirEntry{rest,cast<uint32_t>(len(cf->Data)),{}};
}
}
}}
Slice<xbox_dirEntry> out = Slice<xbox_dirEntry>::make(cast<int64_t>(0ULL),len(byName));
{auto&& tmp53 = byName;
for(auto [tmp54,tmp55]:tmp53){
auto e=tmp55;out = append(out,Slice<xbox_dirEntry>{e});
}}
go_sort_Slice(out,[&](int64_t i,int64_t j)->bool{
return (out[i].Name < out[j].Name);
}
);
return out;
}
if (((!m->Disc) || (!fo->entry.IsDir))) {
return {};
}
auto tmp56 = xbox_Image_ReadDir(m->Disc,fo->entry.Path);
Slice<xbox_Entry> es = std::get<0>(tmp56);
Error err = std::get<1>(tmp56);
if (bool(err)) {
return {};
}
Slice<xbox_dirEntry> out = Slice<xbox_dirEntry>::make(cast<int64_t>(0ULL),len(es));
{auto&& tmp57 = es;
for(int64_t tmp58=0;tmp58<len(tmp57);++tmp58){
auto e=tmp57[tmp58];out = append(out,Slice<xbox_dirEntry>{xbox_dirEntry{e.Name,e.Size,e.IsDir}});
}}
go_sort_Slice(out,[&](int64_t i,int64_t j)->bool{
return (out[i].Name < out[j].Name);
}
);
return out;
}
}
// tools/platform/xbox/kernel_file.go:312:1
bool xbox_matchPattern(std::string pat,std::string name){
{
auto tmp59 = std::make_tuple(go_strings_ToUpper(pat),go_strings_ToUpper(name));
pat = std::get<0>(tmp59);
name = std::get<1>(tmp59);
auto tmp60 = std::make_tuple(cast<int64_t>(0ULL),cast<int64_t>(0ULL),cast<int64_t>(-1ULL),cast<int64_t>(0ULL));
int64_t p = std::get<0>(tmp60);
int64_t n = std::get<1>(tmp60);
int64_t star = std::get<2>(tmp60);
int64_t mark = std::get<3>(tmp60);
{;for (;(n < len(name));){
{
if (((p < len(pat)) && (((cast<uint8_t>(pat[p]) == cast<uint8_t>(63ULL)) || (cast<uint8_t>(pat[p]) == cast<uint8_t>(name[n])))))){
auto tmp62 = std::make_tuple(cast<int64_t>((p + cast<int64_t>(1ULL))),cast<int64_t>((n + cast<int64_t>(1ULL))));
p = std::get<0>(tmp62);
n = std::get<1>(tmp62);
}
else if (((p < len(pat)) && (cast<uint8_t>(pat[p]) == cast<uint8_t>(42ULL)))){
auto tmp63 = std::make_tuple(p,n);
star = std::get<0>(tmp63);
mark = std::get<1>(tmp63);
p++;
}
else if ((star >= cast<int64_t>(0ULL))){
auto tmp64 = std::make_tuple(cast<int64_t>((star + cast<int64_t>(1ULL))),cast<int64_t>((mark + cast<int64_t>(1ULL))));
p = std::get<0>(tmp64);
mark = std::get<1>(tmp64);
n = mark;
}
else {
return false;
}
}
tmp61:;
}
}{;for (;((p < len(pat)) && (cast<uint8_t>(pat[p]) == cast<uint8_t>(42ULL)));){
p++;
}
}return (p == len(pat));
}
}
// tools/platform/xbox/kernel_file.go:359:1
std::tuple<uint64_t,uint64_t> xbox_Machine_volumeUnits(xbox_Machine* m,xbox_fileObject* fo){
uint64_t total{};
uint64_t avail{};
{
if (((!fo->cache) && (!fo->dir))) {
if ((!m->Disc)) {
return {cast<uint64_t>(0ULL),cast<uint64_t>(0ULL)};
}
return {divi<uint64_t>(cast<uint64_t>(m->Disc->Size),cast<uint64_t>(2048ULL)),cast<uint64_t>(0ULL)};
}
uint64_t used = cast<uint64_t>(0ULL);
{auto&& tmp65 = m->cacheFS;
for(auto [tmp66,tmp67]:tmp65){
auto k=tmp66;auto cf=tmp67;if ((k == xbox_rawPartition0Key)) {
continue;
}
used += divi<uint64_t>((cast<uint64_t>((cast<uint64_t>((cast<uint64_t>(len(cf->Data)) + cast<uint64_t>(16384ULL))) - cast<uint64_t>(1ULL)))),cast<uint64_t>(16384ULL));
}}
if ((used > cast<uint64_t>(262144ULL))) {
return {cast<uint64_t>(262144ULL),cast<uint64_t>(0ULL)};
}
return {cast<uint64_t>(262144ULL),cast<uint64_t>((cast<uint64_t>(262144ULL) - used))};
}
}
// tools/platform/xbox/kernel_file.go:389:1
void xbox_Machine_writeNetworkOpenInfo(xbox_Machine* m,uint32_t buf,uint32_t size,bool isDir){
{
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < cast<uint32_t>(56ULL));i += cast<uint32_t>(4ULL)){
xbox_Machine_write32(m,cast<uint32_t>((buf + i)),cast<uint32_t>(0ULL));
}
}xbox_Machine_write32(m,cast<uint32_t>((buf + cast<uint32_t>(32ULL))),size);
xbox_Machine_write32(m,cast<uint32_t>((buf + cast<uint32_t>(40ULL))),size);
uint32_t attrs = cast<uint32_t>(129ULL);
if (isDir) {
attrs = cast<uint32_t>(17ULL);
}
xbox_Machine_write32(m,cast<uint32_t>((buf + cast<uint32_t>(48ULL))),attrs);
}
}
// tools/platform/xbox/kernel_file.go:405:1
Map<std::string,int64_t> xbox_Machine_DebugCacheFS(xbox_Machine* m){
{
Map<std::string,int64_t> out = Map<std::string,int64_t>{};
{auto&& tmp68 = m->cacheFS;
for(auto [tmp69,tmp70]:tmp68){
auto k=tmp69;auto v=tmp70;out[k] = len(v->Data);
}}
return out;
}
}
// tools/platform/xbox/kernel_file.go:413:1
Slice<uint8_t> xbox_Machine_DebugCacheFile(xbox_Machine* m,std::string key){
{
{
xbox_cacheFile* cf = get(m->cacheFS,key);
if (bool(cf)) {
return cf->Data;
}
}
return {};
}
}
// tools/platform/xbox/kernel_file.go:423:1
std::string xbox_Machine_readObjectAttributesPath(xbox_Machine* m,uint32_t oa){
{
if ((oa == cast<uint32_t>(0ULL))) {
return std::string("",0);
}
return xbox_Machine_readObjectString(m,xbox_Machine_read32(m,cast<uint32_t>((oa + cast<uint32_t>(4ULL)))));
}
}
// tools/platform/xbox/kernel_file.go:432:1
std::string xbox_Machine_readObjectString(xbox_Machine* m,uint32_t p){
{
if ((p == cast<uint32_t>(0ULL))) {
return std::string("",0);
}
uint32_t length = cast<uint32_t>(xbox_Machine_read16(m,p));
uint32_t buf = xbox_Machine_read32(m,cast<uint32_t>((p + cast<uint32_t>(4ULL))));
if ((((buf == cast<uint32_t>(0ULL)) || (length == cast<uint32_t>(0ULL))) || (length > cast<uint32_t>(1024ULL)))) {
return std::string("",0);
}
Slice<uint8_t> b = Slice<uint8_t>::make(length);
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < length);i++){
b[i] = xbox_Machine_Read(m,cast<uint32_t>((buf + i)));
}
}return cast<std::string>(b);
}
}
// tools/platform/xbox/kernel_file.go:466:1
uint32_t xbox_Machine_openFile(xbox_Machine* m,uint32_t handleOut,uint32_t oa,uint32_t iosb,uint32_t disposition){
{
std::string path = xbox_Machine_readObjectAttributesPath(m,oa);
auto newHandle = [&](xbox_fileObject* fo,uint32_t info)->uint32_t{
uint32_t h = xbox_Machine_allocKObject(m,cast<uint32_t>(64ULL));
m->objects[h] = arenaNew(xbox_kobject{std::string("file",4),h,true,{},{},{},{},{},{}});
xbox_Machine_writeSignal(m,h,true);
m->files[h] = fo;
if ((handleOut != cast<uint32_t>(0ULL))) {
xbox_Machine_write32(m,handleOut,h);
}
return xbox_Machine_finishOpen(m,iosb,h,info,cast<uint32_t>(0ULL));
}
;
{
auto tmp71 = xbox_rawDevicePath(path);
std::string key = std::get<0>(tmp71);
bool ok = std::get<1>(tmp71);
if (ok) {
xbox_cacheFile* cf = xbox_Machine_rawDevice(m,key);
xbox_Machine_logf(m,std::string("NtOpenFile: %q -> raw device (%d bytes, blank: no XONLINE account)",66),path,len(cf->Data));
return newHandle(arenaNew(xbox_fileObject{{},cf,key,{},{},{}}),cast<uint32_t>(1ULL));
}
}
{
auto tmp72 = xbox_resolveCachePath(path);
std::string key = std::get<0>(tmp72);
bool ok = std::get<1>(tmp72);
if (ok) {
xbox_cacheFile* cf = get(m->cacheFS,key);
if (((!cf) && (disposition != cast<uint32_t>(2ULL)))) {
{
auto tmp73 = xbox_Machine_statPath(m,path);
bool isDir = std::get<1>(tmp73);
bool found = std::get<2>(tmp73);
if ((found && isDir)) {
xbox_Machine_logf(m,std::string("NtOpen/CreateFile: %q -> hdd %q (directory)",43),path,key);
return newHandle(arenaNew(xbox_fileObject{{},{},key,true,{},{}}),cast<uint32_t>(1ULL));
}
}
}
{
if ((!cf)){
if (((disposition == cast<uint32_t>(1ULL)) || (disposition == cast<uint32_t>(4ULL)))) {
xbox_Machine_logf(m,std::string("NtOpen/CreateFile: %q (hdd %q) -> not found (disp %d)",53),path,key,disposition);
return xbox_Machine_finishOpen(m,iosb,cast<uint32_t>(0ULL),cast<uint32_t>(0ULL),cast<uint32_t>(3221225524ULL));
}
cf = arenaNew(xbox_cacheFile{});
m->cacheFS[key] = cf;
xbox_Machine_logf(m,std::string("NtCreateFile: %q -> hdd %q CREATED",34),path,key);
return newHandle(arenaNew(xbox_fileObject{{},cf,key,{},{},{}}),cast<uint32_t>(2ULL));
}
else if ((disposition == cast<uint32_t>(2ULL))){
xbox_Machine_logf(m,std::string("NtCreateFile: %q (hdd %q) -> collision",38),path,key);
return xbox_Machine_finishOpen(m,iosb,cast<uint32_t>(0ULL),cast<uint32_t>(0ULL),cast<uint32_t>(3221225525ULL));
}
else if ((disposition == cast<uint32_t>(0ULL)) || (disposition == cast<uint32_t>(4ULL)) || (disposition == cast<uint32_t>(5ULL))){
cf->Data = sub(cf->Data,0,cast<int64_t>(0ULL));
xbox_Machine_logf(m,std::string("NtCreateFile: %q -> hdd %q overwritten",38),path,key);
return newHandle(arenaNew(xbox_fileObject{{},cf,key,{},{},{}}),cast<uint32_t>(3ULL));
}
else {
xbox_Machine_logf(m,std::string("NtOpenFile: %q -> hdd %q (%d bytes)",35),path,key,len(cf->Data));
return newHandle(arenaNew(xbox_fileObject{{},cf,key,{},{},{}}),cast<uint32_t>(1ULL));
}
}
tmp74:;
}
}
auto tmp75 = xbox_resolveDiscPath(path);
std::string disc = std::get<0>(tmp75);
bool ok = std::get<1>(tmp75);
if (((!ok) || (!m->Disc))) {
xbox_Machine_logf(m,std::string("NtOpenFile: %q -> not on disc",29),path);
return xbox_Machine_finishOpen(m,iosb,cast<uint32_t>(0ULL),cast<uint32_t>(0ULL),cast<uint32_t>(3221225524ULL));
}
auto tmp76 = xbox_Image_resolve(m->Disc,disc);
xbox_Entry e = std::get<0>(tmp76);
Error err = std::get<1>(tmp76);
if (bool(err)) {
xbox_Machine_logf(m,std::string("NtOpenFile: %q (disc %q) -> not found",37),path,disc);
return xbox_Machine_finishOpen(m,iosb,cast<uint32_t>(0ULL),cast<uint32_t>(0ULL),cast<uint32_t>(3221225524ULL));
}
xbox_Machine_logf(m,std::string("NtOpenFile: %q -> disc %q (%d bytes)",36),path,disc,e.Size);
return newHandle(arenaNew(xbox_fileObject{e,{},{},{},{},{}}),cast<uint32_t>(1ULL));
}
}
// tools/platform/xbox/kernel_file.go:546:1
uint32_t xbox_Machine_finishOpen(xbox_Machine* m,uint32_t iosb,uint32_t _handle,uint32_t info,uint32_t status){
{
if ((iosb != cast<uint32_t>(0ULL))) {
xbox_Machine_write32(m,cast<uint32_t>((iosb + cast<uint32_t>(0ULL))),status);
xbox_Machine_write32(m,cast<uint32_t>((iosb + cast<uint32_t>(4ULL))),info);
}
xbox_Machine_setRet(m,status);
return status;
}
}
// tools/platform/xbox/kernel_file.go:566:1
uint64_t xbox_ioCompletionTicks(uint32_t n){
{
return cast<uint64_t>((cast<uint64_t>(733466ULL) * (cast<uint64_t>((cast<uint64_t>(10ULL) + divi<uint64_t>(cast<uint64_t>(n),cast<uint64_t>(8192ULL)))))));
}
}
// tools/platform/xbox/kernel_file.go:593:1
uint32_t xbox_Machine_readFile(xbox_Machine* m,uint32_t handle,uint32_t event,uint32_t apcRoutine,uint32_t apcCtx,uint32_t iosb,uint32_t buffer,uint32_t length,uint32_t byteOffsetPtr){
{
if ((apcRoutine != cast<uint32_t>(0ULL))) {
x86_CPU_Halt(m->CPU,std::string("NtReadFile: ApcRoutine %08X (ctx %08X) passed from %08X \342\200\224 APC completion not implemented",90),apcRoutine,apcCtx,xbox_Machine_retAddr(m));
return cast<uint32_t>(0ULL);
}
xbox_fileObject* fo = get(m->files,handle);
if ((!fo)) {
return xbox_Machine_finishOpen(m,iosb,handle,cast<uint32_t>(0ULL),cast<uint32_t>(3221225480ULL));
}
auto tmp77 = fo->off;
uint32_t off = tmp77;
if ((byteOffsetPtr != cast<uint32_t>(0ULL))) {
off = xbox_Machine_read32(m,byteOffsetPtr);
}
auto tmp78 = xbox_fileObject_size(fo);
uint32_t size = tmp78;
if ((off > size)) {
off = size;
}
uint32_t n = length;
if ((cast<uint32_t>((off + n)) > size)) {
n = cast<uint32_t>((size - off));
}
if ((n > cast<uint32_t>(0ULL))) {
Slice<uint8_t> data={};
if (bool(fo->cache)) {
data = sub(fo->cache->Data,off,cast<uint32_t>((off + n)));
}
else {
Error err={};
auto tmp79 = xbox_Image_Read(m->Disc,cast<int64_t>((cast<int64_t>((cast<int64_t>(fo->entry.Sector) * cast<int64_t>(2048ULL))) + cast<int64_t>(off))),cast<int64_t>(n));
data = std::get<0>(tmp79);
err = std::get<1>(tmp79);
if (bool(err)) {
return xbox_Machine_finishOpen(m,iosb,handle,cast<uint32_t>(0ULL),cast<uint32_t>(3221225480ULL));
}
}
{auto&& tmp80 = data;
for(int64_t tmp81=0;tmp81<len(tmp80);++tmp81){
auto i=tmp81;auto b=tmp80[tmp81];xbox_Machine_Write(m,cast<uint32_t>((buffer + cast<uint32_t>(i))),b);
}}
}
fo->off = cast<uint32_t>((off + n));
if (((iosb != cast<uint32_t>(0ULL)) && (xbox_Machine_read32(m,iosb) == cast<uint32_t>(259ULL)))) {
{
xbox_kobject* o = get(m->objects,handle);
if (bool(o)) {
o->signaled = false;
xbox_Machine_writeSignal(m,handle,false);
}
}
m->pendingIO = append(m->pendingIO,Slice<xbox_pendingIO>{xbox_pendingIO{cast<uint64_t>((m->tick + xbox_ioCompletionTicks(n))),iosb,event,n,{},handle}});
xbox_Machine_logf(m,std::string("NtReadFile: handle %08X off %d len %d -> %d bytes (async, event %08X, iosb %08X, buf %08X, due +%d, from %08X)",110),handle,off,length,n,event,iosb,buffer,xbox_ioCompletionTicks(n),xbox_Machine_retAddr(m));
xbox_Machine_setRet(m,cast<uint32_t>(259ULL));
return cast<uint32_t>(259ULL);
}
xbox_Machine_logf(m,std::string("NtReadFile: handle %08X off %d len %d -> %d bytes (sync, tick %d)",65),handle,off,length,n,m->tick);
return xbox_Machine_finishOpen(m,iosb,handle,n,cast<uint32_t>(0ULL));
}
}
// tools/platform/xbox/kernel_file.go:657:1
uint32_t xbox_Machine_writeFile(xbox_Machine* m,uint32_t handle,uint32_t event,uint32_t iosb,uint32_t buffer,uint32_t length,uint32_t byteOffsetPtr){
{
xbox_fileObject* fo = get(m->files,handle);
if ((!fo)) {
return xbox_Machine_finishOpen(m,iosb,handle,cast<uint32_t>(0ULL),cast<uint32_t>(3221225480ULL));
}
if ((!fo->cache)) {
x86_CPU_Halt(m->CPU,std::string("NtWriteFile: write to a read-only disc file (handle %08X) from %08X",67),handle,xbox_Machine_retAddr(m));
return cast<uint32_t>(0ULL);
}
auto tmp82 = fo->off;
uint32_t off = tmp82;
if ((byteOffsetPtr != cast<uint32_t>(0ULL))) {
off = xbox_Machine_read32(m,byteOffsetPtr);
}
{
int64_t need = cast<int64_t>((cast<int64_t>(off) + cast<int64_t>(length)));
if ((need > len(fo->cache->Data))) {
fo->cache->Data = append(fo->cache->Data,Slice<uint8_t>::make(cast<int64_t>((need - len(fo->cache->Data)))));
}
}
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < length);i++){
fo->cache->Data[cast<uint32_t>((off + i))] = xbox_Machine_Read(m,cast<uint32_t>((buffer + i)));
}
}fo->off = cast<uint32_t>((off + length));
xbox_Machine_logf(m,std::string("NtWriteFile: handle %08X off %d len %d (file now %d bytes)",58),handle,off,length,len(fo->cache->Data));
return xbox_Machine_finishOpen(m,iosb,handle,length,cast<uint32_t>(0ULL));
}
}
// tools/platform/xbox/kernel_file.go:694:1
uint32_t xbox_Machine_setFileLength(xbox_Machine* m,xbox_fileObject* fo,uint32_t n){
{
if ((!fo->cache)) {
x86_CPU_Halt(m->CPU,std::string("NtSetInformationFile: set length on a read-only disc file (%q) from %08X",72),fo->entry.Path,xbox_Machine_retAddr(m));
return cast<uint32_t>(3221225485ULL);
}
if (xbox_fileObject_isDir(fo)) {
x86_CPU_Halt(m->CPU,std::string("NtSetInformationFile: set length on a directory (%q) from %08X",62),fo->key,xbox_Machine_retAddr(m));
return cast<uint32_t>(3221225485ULL);
}
{
uint32_t cur = cast<uint32_t>(len(fo->cache->Data));
if ((n < cur)){
fo->cache->Data = sub(fo->cache->Data,0,n);
}
else if ((n > cur)){
fo->cache->Data = append(fo->cache->Data,Slice<uint8_t>::make(cast<uint32_t>((n - cur))));
}
}
tmp83:;
xbox_Machine_logf(m,std::string("NtSetInformationFile: handle file %q length -> %d bytes",55),fo->key,n);
return cast<uint32_t>(0ULL);
}
}
// tools/platform/xbox/kernel_file.go:727:1
void xbox_Machine_ioTick(xbox_Machine* m){
{
if ((len(m->pendingIO) == cast<int64_t>(0ULL))) {
return ;
}
Slice<xbox_pendingIO> kept = sub(m->pendingIO,0,cast<int64_t>(0ULL));
{auto&& tmp84 = m->pendingIO;
for(int64_t tmp85=0;tmp85<len(tmp84);++tmp85){
auto p=tmp84[tmp85];if ((p.Due > m->tick)) {
kept = append(kept,Slice<xbox_pendingIO>{p});
continue;
}
if ((p.IOSB != cast<uint32_t>(0ULL))) {
xbox_Machine_write32(m,cast<uint32_t>((p.IOSB + cast<uint32_t>(0ULL))),p.Status);
xbox_Machine_write32(m,cast<uint32_t>((p.IOSB + cast<uint32_t>(4ULL))),p.Info);
}
if ((p.Event != cast<uint32_t>(0ULL))) {
{
xbox_kobject* o = xbox_Machine_objAt(m,p.Event);
if (bool(o)) {
o->signaled = true;
xbox_Machine_writeSignal(m,o->addr,true);
xbox_kobject_noteSignal(o,std::string("io",2),cast<uint32_t>(0ULL),m->tick);
xbox_Machine_wakeWaiters(m,p.Event);
}
}
}
if ((p.Handle != cast<uint32_t>(0ULL))) {
{
xbox_kobject* o = get(m->objects,p.Handle);
if (bool(o)) {
o->signaled = true;
xbox_Machine_writeSignal(m,p.Handle,true);
xbox_kobject_noteSignal(o,std::string("io",2),cast<uint32_t>(0ULL),m->tick);
xbox_Machine_wakeWaiters(m,p.Handle);
}
}
}
}}
m->pendingIO = kept;
}
}
// tools/platform/xbox/kernel_objects.go:30:1
uint32_t xbox_Machine_newObject(xbox_Machine* m,std::string kind,uint8_t typ,bool signaled,int32_t count,int32_t limit){
{
uint32_t addr = xbox_Machine_allocKObject(m,cast<uint32_t>(32ULL));
xbox_Machine_Write(m,cast<uint32_t>((addr + cast<uint32_t>(0ULL))),typ);
xbox_Machine_writeSignal(m,addr,signaled);
xbox_Machine_write32(m,cast<uint32_t>((addr + cast<uint32_t>(8ULL))),cast<uint32_t>((addr + cast<uint32_t>(8ULL))));
xbox_Machine_write32(m,cast<uint32_t>((cast<uint32_t>((addr + cast<uint32_t>(8ULL))) + cast<uint32_t>(4ULL))),cast<uint32_t>((addr + cast<uint32_t>(8ULL))));
m->objects[addr] = arenaNew(xbox_kobject{kind,addr,signaled,count,limit,{},{},{},{}});
xbox_Machine_logf(m,std::string("newObject: %s @%08X signaled=%v from %08X",41),kind,addr,signaled,xbox_Machine_retAddr(m));
return addr;
}
}
// tools/platform/xbox/kernel_objects.go:41:1
void xbox_Machine_writeSignal(xbox_Machine* m,uint32_t addr,bool signaled){
{
if (signaled) {
xbox_Machine_write32(m,cast<uint32_t>((addr + cast<uint32_t>(4ULL))),cast<uint32_t>(1ULL));
}
else {
xbox_Machine_write32(m,cast<uint32_t>((addr + cast<uint32_t>(4ULL))),cast<uint32_t>(0ULL));
}
}
}
// tools/platform/xbox/kernel_objects.go:52:1
xbox_kobject* xbox_Machine_objAt(xbox_Machine* m,uint32_t handle){
{
xbox_kobject* o = get(m->objects,handle);
if (bool(o)) {
o->signaled = (xbox_Machine_read32(m,cast<uint32_t>((handle + cast<uint32_t>(4ULL)))) != cast<uint32_t>(0ULL));
}
return o;
}
}
// tools/platform/xbox/kernel_objects.go:67:1
xbox_kobject* xbox_Machine_guestObjAt(xbox_Machine* m,uint32_t addr){
{
if ((addr == cast<uint32_t>(0ULL))) {
return {};
}
{
xbox_kobject* o = xbox_Machine_objAt(m,addr);
if (bool(o)) {
return o;
}
}
xbox_kobject* o = arenaNew(xbox_kobject{std::string("event",5),addr,(xbox_Machine_read32(m,cast<uint32_t>((addr + cast<uint32_t>(4ULL)))) != cast<uint32_t>(0ULL)),{},{},{},{},{},{}});
m->objects[addr] = o;
return o;
}
}
// tools/platform/xbox/kernel_objects.go:82:1
bool xbox_Machine_satisfyWait(xbox_Machine* m,xbox_kobject* o){
{
if ((!o)) {
return true;
}
{
auto tmp87=o->kind;
if (tmp87==(std::string("semaphore",9))){
if ((o->count > cast<int32_t>(0ULL))) {
o->count--;
if ((o->count == cast<int32_t>(0ULL))) {
o->signaled = false;
xbox_Machine_writeSignal(m,o->addr,false);
}
return true;
}
return false;
}
else if (tmp87==(std::string("event-auto",10))){
if (o->signaled) {
o->signaled = false;
xbox_Machine_writeSignal(m,o->addr,false);
return true;
}
return false;
}
else {
return o->signaled;
}
}
tmp86:;
}
}
// tools/platform/xbox/kernel_objects.go:121:1
std::function<int64_t(xbox_Machine*)> xbox_kernelObjectHandler(uint16_t ord){
{
{
switch(ord){
case cast<uint16_t>(202ULL):{
return [=](xbox_Machine* m)->int64_t{
xbox_Machine_openFile(m,xbox_Machine_arg(m,cast<int64_t>(0ULL)),xbox_Machine_arg(m,cast<int64_t>(2ULL)),xbox_Machine_arg(m,cast<int64_t>(3ULL)),cast<uint32_t>(1ULL));
return cast<int64_t>(6ULL);
}
;
break;}
case cast<uint16_t>(190ULL):{
return [=](xbox_Machine* m)->int64_t{
xbox_Machine_openFile(m,xbox_Machine_arg(m,cast<int64_t>(0ULL)),xbox_Machine_arg(m,cast<int64_t>(2ULL)),xbox_Machine_arg(m,cast<int64_t>(3ULL)),xbox_Machine_arg(m,cast<int64_t>(7ULL)));
return cast<int64_t>(9ULL);
}
;
break;}
case cast<uint16_t>(219ULL):{
return [=](xbox_Machine* m)->int64_t{
xbox_Machine_readFile(m,xbox_Machine_arg(m,cast<int64_t>(0ULL)),xbox_Machine_arg(m,cast<int64_t>(1ULL)),xbox_Machine_arg(m,cast<int64_t>(2ULL)),xbox_Machine_arg(m,cast<int64_t>(3ULL)),xbox_Machine_arg(m,cast<int64_t>(4ULL)),xbox_Machine_arg(m,cast<int64_t>(5ULL)),xbox_Machine_arg(m,cast<int64_t>(6ULL)),xbox_Machine_arg(m,cast<int64_t>(7ULL)));
return cast<int64_t>(8ULL);
}
;
break;}
case cast<uint16_t>(198ULL):{
return [=](xbox_Machine* m)->int64_t{
auto tmp88 = std::make_tuple(xbox_Machine_arg(m,cast<int64_t>(0ULL)),xbox_Machine_arg(m,cast<int64_t>(1ULL)));
uint32_t h = std::get<0>(tmp88);
uint32_t iosb = std::get<1>(tmp88);
if ((!get(m->files,h))) {
xbox_Machine_finishOpen(m,iosb,h,cast<uint32_t>(0ULL),cast<uint32_t>(3221225480ULL));
return cast<int64_t>(2ULL);
}
xbox_Machine_finishOpen(m,iosb,h,cast<uint32_t>(0ULL),cast<uint32_t>(0ULL));
return cast<int64_t>(2ULL);
}
;
break;}
case cast<uint16_t>(236ULL):{
return [=](xbox_Machine* m)->int64_t{
xbox_Machine_writeFile(m,xbox_Machine_arg(m,cast<int64_t>(0ULL)),xbox_Machine_arg(m,cast<int64_t>(1ULL)),xbox_Machine_arg(m,cast<int64_t>(4ULL)),xbox_Machine_arg(m,cast<int64_t>(5ULL)),xbox_Machine_arg(m,cast<int64_t>(6ULL)),xbox_Machine_arg(m,cast<int64_t>(7ULL)));
return cast<int64_t>(8ULL);
}
;
break;}
case cast<uint16_t>(246ULL):{
return [=](xbox_Machine* m)->int64_t{
auto tmp89 = std::make_tuple(xbox_Machine_arg(m,cast<int64_t>(0ULL)),xbox_Machine_arg(m,cast<int64_t>(2ULL)));
uint32_t h = std::get<0>(tmp89);
uint32_t out = std::get<1>(tmp89);
if ((!get(m->objects,h))) {
xbox_Machine_setRet(m,cast<uint32_t>(3221225480ULL));
return cast<int64_t>(3ULL);
}
if ((out != cast<uint32_t>(0ULL))) {
xbox_Machine_write32(m,out,h);
}
xbox_Machine_setRet(m,cast<uint32_t>(0ULL));
return cast<int64_t>(3ULL);
}
;
break;}
case cast<uint16_t>(126ULL):case cast<uint16_t>(127ULL):{
auto tmp90 = ord;
uint16_t ord = tmp90;
return [=](xbox_Machine* m)->int64_t{
uint64_t v = xbox_Machine_guestTSC(m);
if ((ord == cast<uint16_t>(127ULL))) {
v = cast<uint64_t>(733466000ULL);
}
m->CPU->Regs[cast<int64_t>(0ULL)] = cast<uint32_t>(v);
m->CPU->Regs[cast<int64_t>(2ULL)] = cast<uint32_t>(shr<uint64_t>(v,cast<int64_t>(32ULL)));
return cast<int64_t>(0ULL);
}
;
break;}
case cast<uint16_t>(128ULL):{
return [=](xbox_Machine* m)->int64_t{
{
uint32_t p = xbox_Machine_arg(m,cast<int64_t>(0ULL));
if ((p != cast<uint32_t>(0ULL))) {
uint64_t t = xbox_Machine_systemTime100ns(m);
xbox_Machine_write32(m,p,cast<uint32_t>(t));
xbox_Machine_write32(m,cast<uint32_t>((p + cast<uint32_t>(4ULL))),cast<uint32_t>(shr<uint64_t>(t,cast<int64_t>(32ULL))));
}
}
xbox_Machine_setRet(m,cast<uint32_t>(0ULL));
return cast<int64_t>(1ULL);
}
;
break;}
case cast<uint16_t>(143ULL):{
return [=](xbox_Machine* m)->int64_t{
auto tmp91 = std::make_tuple(xbox_Machine_arg(m,cast<int64_t>(0ULL)),cast<int32_t>(xbox_Machine_arg(m,cast<int64_t>(1ULL))));
uint32_t kt = std::get<0>(tmp91);
int32_t inc = std::get<1>(tmp91);
int32_t old = cast<int32_t>(16ULL);
{
xbox_kobject* o = get(m->objects,kt);
if ((bool(o) && bool(o->thread))) {
old = o->thread->priority;
int32_t p = cast<int32_t>((cast<int32_t>(16ULL) + inc));
if ((p < cast<int32_t>(0ULL))) {
p = cast<int32_t>(0ULL);
}
else if ((p > cast<int32_t>(31ULL))) {
p = cast<int32_t>(31ULL);
}
o->thread->priority = p;
}
}
xbox_Machine_setRet(m,cast<uint32_t>(old));
return cast<int64_t>(2ULL);
}
;
break;}
case cast<uint16_t>(250ULL):{
return [=](xbox_Machine* m)->int64_t{
xbox_Machine_setRet(m,cast<uint32_t>(0ULL));
return cast<int64_t>(0ULL);
}
;
break;}
case cast<uint16_t>(226ULL):{
return [=](xbox_Machine* m)->int64_t{
auto tmp92 = std::make_tuple(xbox_Machine_arg(m,cast<int64_t>(0ULL)),xbox_Machine_arg(m,cast<int64_t>(1ULL)),xbox_Machine_arg(m,cast<int64_t>(2ULL)),xbox_Machine_arg(m,cast<int64_t>(3ULL)),xbox_Machine_arg(m,cast<int64_t>(4ULL)));
uint32_t h = std::get<0>(tmp92);
uint32_t iosb = std::get<1>(tmp92);
uint32_t buf = std::get<2>(tmp92);
uint32_t ln = std::get<3>(tmp92);
uint32_t class_ = std::get<4>(tmp92);
xbox_fileObject* fo = get(m->files,h);
if ((!fo)) {
xbox_Machine_finishOpen(m,iosb,h,cast<uint32_t>(0ULL),cast<uint32_t>(3221225480ULL));
return cast<int64_t>(5ULL);
}
{
if (((class_ == cast<uint32_t>(14ULL)) && (ln >= cast<uint32_t>(8ULL)))){
fo->off = xbox_Machine_read32(m,buf);
xbox_Machine_finishOpen(m,iosb,h,cast<uint32_t>(0ULL),cast<uint32_t>(0ULL));
}
else if (((class_ == cast<uint32_t>(4ULL)) && (ln >= cast<uint32_t>(40ULL)))){
if (((!fo->cache) || (fo->key == std::string("",0)))) {
x86_CPU_Halt(m->CPU,std::string("NtSetInformationFile: set FileBasicInformation on a non-store file (%q) from %08X",81),fo->entry.Path,xbox_Machine_retAddr(m));
return cast<int64_t>(5ULL);
}
Slice<uint8_t> blob = Slice<uint8_t>::make(cast<int64_t>(40ULL));
{auto&& tmp94 = blob;
for(int64_t tmp95=0;tmp95<len(tmp94);++tmp95){
auto i=tmp95;blob[i] = xbox_Machine_Read(m,cast<uint32_t>((buf + cast<uint32_t>(i))));
}}
m->fileBasic[fo->key] = blob;
xbox_Machine_logf(m,std::string("NtSetInformationFile: %q FileBasicInformation stored (attrs %08X)",65),fo->key,xbox_Machine_read32(m,cast<uint32_t>((buf + cast<uint32_t>(32ULL)))));
xbox_Machine_finishOpen(m,iosb,h,cast<uint32_t>(0ULL),cast<uint32_t>(0ULL));
}
else if (((((class_ == cast<uint32_t>(19ULL)) || (class_ == cast<uint32_t>(20ULL)))) && (ln >= cast<uint32_t>(8ULL)))){
{
uint32_t hi = xbox_Machine_read32(m,cast<uint32_t>((buf + cast<uint32_t>(4ULL))));
if ((hi != cast<uint32_t>(0ULL))) {
x86_CPU_Halt(m->CPU,(std::string("NtSetInformationFile: class %d length %08X_%08X exceeds ",56) + std::string("this store's 32-bit files, from %08X",36)),class_,hi,xbox_Machine_read32(m,buf),xbox_Machine_retAddr(m));
return cast<int64_t>(5ULL);
}
}
{
uint32_t err = xbox_Machine_setFileLength(m,fo,xbox_Machine_read32(m,buf));
if ((err != cast<uint32_t>(0ULL))) {
xbox_Machine_finishOpen(m,iosb,h,cast<uint32_t>(0ULL),err);
return cast<int64_t>(5ULL);
}
}
xbox_Machine_finishOpen(m,iosb,h,cast<uint32_t>(0ULL),cast<uint32_t>(0ULL));
}
else {
x86_CPU_Halt(m->CPU,std::string("NtSetInformationFile: unmodelled class %d (len %d) from %08X",60),class_,ln,xbox_Machine_retAddr(m));
}
}
tmp93:;
return cast<int64_t>(5ULL);
}
;
break;}
case cast<uint16_t>(224ULL):case cast<uint16_t>(231ULL):{
auto tmp96 = ord;
uint16_t ord = tmp96;
return [=](xbox_Machine* m)->int64_t{
auto tmp97 = std::make_tuple(xbox_Machine_arg(m,cast<int64_t>(0ULL)),xbox_Machine_arg(m,cast<int64_t>(1ULL)));
uint32_t h = std::get<0>(tmp97);
uint32_t prevOut = std::get<1>(tmp97);
xbox_kobject* o = get(m->objects,h);
if (((!o) || (!o->thread))) {
xbox_Machine_setRet(m,cast<uint32_t>(3221225480ULL));
return cast<int64_t>(2ULL);
}
xbox_thread* t = o->thread;
int32_t prev = t->suspendCount;
if ((ord == cast<uint16_t>(224ULL))) {
if ((t->suspendCount > cast<int32_t>(0ULL))) {
t->suspendCount--;
}
}
else {
t->suspendCount++;
if ((t == m->current)) {
m->reschedule = true;
}
}
if ((prevOut != cast<uint32_t>(0ULL))) {
xbox_Machine_write32(m,prevOut,cast<uint32_t>(prev));
}
xbox_Machine_setRet(m,cast<uint32_t>(0ULL));
return cast<int64_t>(2ULL);
}
;
break;}
case cast<uint16_t>(211ULL):{
return [=](xbox_Machine* m)->int64_t{
auto tmp98 = std::make_tuple(xbox_Machine_arg(m,cast<int64_t>(0ULL)),xbox_Machine_arg(m,cast<int64_t>(1ULL)),xbox_Machine_arg(m,cast<int64_t>(2ULL)),xbox_Machine_arg(m,cast<int64_t>(3ULL)),xbox_Machine_arg(m,cast<int64_t>(4ULL)));
uint32_t h = std::get<0>(tmp98);
uint32_t iosb = std::get<1>(tmp98);
uint32_t buf = std::get<2>(tmp98);
uint32_t ln = std::get<3>(tmp98);
uint32_t class_ = std::get<4>(tmp98);
xbox_fileObject* fo = get(m->files,h);
if ((!fo)) {
xbox_Machine_finishOpen(m,iosb,h,cast<uint32_t>(0ULL),cast<uint32_t>(3221225480ULL));
return cast<int64_t>(5ULL);
}
if (((class_ == cast<uint32_t>(14ULL)) && (ln >= cast<uint32_t>(8ULL)))) {
xbox_Machine_write32(m,cast<uint32_t>((buf + cast<uint32_t>(0ULL))),fo->off);
xbox_Machine_write32(m,cast<uint32_t>((buf + cast<uint32_t>(4ULL))),cast<uint32_t>(0ULL));
xbox_Machine_finishOpen(m,iosb,h,cast<uint32_t>(8ULL),cast<uint32_t>(0ULL));
return cast<int64_t>(5ULL);
}
if (((class_ == cast<uint32_t>(6ULL)) && (ln >= cast<uint32_t>(8ULL)))) {
uint64_t id = cast<uint64_t>(fo->entry.Sector);
if (bool(fo->cache)) {
id = xbox_fnv64(fo->key);
}
xbox_Machine_write32(m,cast<uint32_t>((buf + cast<uint32_t>(0ULL))),cast<uint32_t>(id));
xbox_Machine_write32(m,cast<uint32_t>((buf + cast<uint32_t>(4ULL))),cast<uint32_t>(shr<uint64_t>(id,cast<int64_t>(32ULL))));
xbox_Machine_finishOpen(m,iosb,h,cast<uint32_t>(8ULL),cast<uint32_t>(0ULL));
return cast<int64_t>(5ULL);
}
if (((class_ != cast<uint32_t>(34ULL)) || (ln < cast<uint32_t>(56ULL)))) {
x86_CPU_Halt(m->CPU,std::string("NtQueryInformationFile: unmodelled class %d (len %d) from %08X",62),class_,ln,xbox_Machine_retAddr(m));
return cast<int64_t>(5ULL);
}
xbox_Machine_writeNetworkOpenInfo(m,buf,xbox_fileObject_size(fo),xbox_fileObject_isDir(fo));
xbox_Machine_finishOpen(m,iosb,h,cast<uint32_t>(56ULL),cast<uint32_t>(0ULL));
return cast<int64_t>(5ULL);
}
;
break;}
case cast<uint16_t>(207ULL):{
return [=](xbox_Machine* m)->int64_t{
auto tmp99 = std::make_tuple(xbox_Machine_arg(m,cast<int64_t>(0ULL)),xbox_Machine_arg(m,cast<int64_t>(4ULL)),xbox_Machine_arg(m,cast<int64_t>(5ULL)),xbox_Machine_arg(m,cast<int64_t>(6ULL)));
uint32_t h = std::get<0>(tmp99);
uint32_t iosb = std::get<1>(tmp99);
uint32_t buf = std::get<2>(tmp99);
uint32_t ln = std::get<3>(tmp99);
auto tmp100 = std::make_tuple(xbox_Machine_arg(m,cast<int64_t>(7ULL)),xbox_Machine_arg(m,cast<int64_t>(8ULL)),xbox_Machine_arg(m,cast<int64_t>(9ULL)));
uint32_t class_ = std::get<0>(tmp100);
uint32_t namePtr = std::get<1>(tmp100);
uint32_t restart = std::get<2>(tmp100);
xbox_fileObject* fo = get(m->files,h);
if (((!fo) || (!xbox_fileObject_isDir(fo)))) {
xbox_Machine_finishOpen(m,iosb,h,cast<uint32_t>(0ULL),cast<uint32_t>(3221225480ULL));
return cast<int64_t>(10ULL);
}
if ((class_ != cast<uint32_t>(1ULL))) {
x86_CPU_Halt(m->CPU,std::string("NtQueryDirectoryFile: unmodelled class %d (len %d) from %08X",60),class_,ln,xbox_Machine_retAddr(m));
return cast<int64_t>(10ULL);
}
std::string pat = std::string("*",1);
if ((namePtr != cast<uint32_t>(0ULL))) {
{
std::string s = xbox_Machine_readObjectString(m,namePtr);
if ((s != std::string("",0))) {
pat = s;
}
}
}
if ((restart != cast<uint32_t>(0ULL))) {
fo->scan = cast<int64_t>(0ULL);
}
Slice<xbox_dirEntry> entries = xbox_Machine_listDir(m,fo);
int64_t i = fo->scan;
{;for (;((i < len(entries)) && (!xbox_matchPattern(pat,entries[i].Name)));){
i++;
}
}if ((i >= len(entries))) {
fo->scan = len(entries);
uint32_t status = cast<uint32_t>(2147483654ULL);
if (((restart != cast<uint32_t>(0ULL)) || (i == cast<int64_t>(0ULL)))) {
status = cast<uint32_t>(3221225487ULL);
}
xbox_Machine_logf(m,std::string("NtQueryDirectoryFile: %q pattern %q -> end (%08X)",49),fo->key,pat,status);
xbox_Machine_finishOpen(m,iosb,h,cast<uint32_t>(0ULL),status);
return cast<int64_t>(10ULL);
}
xbox_dirEntry e = entries[i];
fo->scan = cast<int64_t>((i + cast<int64_t>(1ULL)));
uint32_t need = cast<uint32_t>(cast<int64_t>((cast<int64_t>(64ULL) + len(e.Name))));
if ((ln < need)) {
xbox_Machine_finishOpen(m,iosb,h,cast<uint32_t>(0ULL),cast<uint32_t>(2147483653ULL));
return cast<int64_t>(10ULL);
}
{uint32_t o = cast<uint32_t>(0ULL);for (;(o < cast<uint32_t>(64ULL));o += cast<uint32_t>(4ULL)){
xbox_Machine_write32(m,cast<uint32_t>((buf + o)),cast<uint32_t>(0ULL));
}
}xbox_Machine_write32(m,cast<uint32_t>((buf + cast<uint32_t>(0ULL))),cast<uint32_t>(0ULL));
xbox_Machine_write32(m,cast<uint32_t>((buf + cast<uint32_t>(4ULL))),cast<uint32_t>(i));
uint32_t attrs = cast<uint32_t>(129ULL);
if (e.IsDir) {
attrs = cast<uint32_t>(17ULL);
}
else {
xbox_Machine_write32(m,cast<uint32_t>((buf + cast<uint32_t>(40ULL))),e.Size);
xbox_Machine_write32(m,cast<uint32_t>((buf + cast<uint32_t>(48ULL))),e.Size);
}
xbox_Machine_write32(m,cast<uint32_t>((buf + cast<uint32_t>(56ULL))),attrs);
xbox_Machine_write32(m,cast<uint32_t>((buf + cast<uint32_t>(60ULL))),cast<uint32_t>(len(e.Name)));
{int64_t j = cast<int64_t>(0ULL);for (;(j < len(e.Name));j++){
xbox_Machine_Write(m,cast<uint32_t>((cast<uint32_t>((buf + cast<uint32_t>(64ULL))) + cast<uint32_t>(j))),cast<uint8_t>(e.Name[j]));
}
}xbox_Machine_logf(m,std::string("NtQueryDirectoryFile: %q pattern %q -> %q (%d bytes, dir=%v)",60),fo->key,pat,e.Name,e.Size,e.IsDir);
xbox_Machine_finishOpen(m,iosb,h,need,cast<uint32_t>(0ULL));
return cast<int64_t>(10ULL);
}
;
break;}
case cast<uint16_t>(210ULL):{
return [=](xbox_Machine* m)->int64_t{
auto tmp101 = std::make_tuple(xbox_Machine_arg(m,cast<int64_t>(0ULL)),xbox_Machine_arg(m,cast<int64_t>(1ULL)));
uint32_t oa = std::get<0>(tmp101);
uint32_t buf = std::get<1>(tmp101);
std::string path = xbox_Machine_readObjectAttributesPath(m,oa);
auto tmp102 = xbox_Machine_statPath(m,path);
uint32_t size = std::get<0>(tmp102);
bool isDir = std::get<1>(tmp102);
bool found = std::get<2>(tmp102);
if ((!found)) {
xbox_Machine_logf(m,std::string("NtQueryFullAttributesFile: %q -> not found (from %08X)",54),path,xbox_Machine_retAddr(m));
xbox_Machine_setRet(m,cast<uint32_t>(3221225524ULL));
return cast<int64_t>(2ULL);
}
xbox_Machine_writeNetworkOpenInfo(m,buf,size,isDir);
xbox_Machine_logf(m,std::string("NtQueryFullAttributesFile: %q -> %d bytes, dir=%v (from %08X)",61),path,size,isDir,xbox_Machine_retAddr(m));
xbox_Machine_setRet(m,cast<uint32_t>(0ULL));
return cast<int64_t>(2ULL);
}
;
break;}
case cast<uint16_t>(218ULL):{
return [=](xbox_Machine* m)->int64_t{
auto tmp103 = std::make_tuple(xbox_Machine_arg(m,cast<int64_t>(0ULL)),xbox_Machine_arg(m,cast<int64_t>(1ULL)),xbox_Machine_arg(m,cast<int64_t>(2ULL)),xbox_Machine_arg(m,cast<int64_t>(3ULL)),xbox_Machine_arg(m,cast<int64_t>(4ULL)));
uint32_t h = std::get<0>(tmp103);
uint32_t iosb = std::get<1>(tmp103);
uint32_t buf = std::get<2>(tmp103);
uint32_t ln = std::get<3>(tmp103);
uint32_t class_ = std::get<4>(tmp103);
xbox_fileObject* fo = get(m->files,h);
if ((!fo)) {
xbox_Machine_finishOpen(m,iosb,h,cast<uint32_t>(0ULL),cast<uint32_t>(3221225480ULL));
return cast<int64_t>(5ULL);
}
if (((class_ == cast<uint32_t>(3ULL)) && (ln >= cast<uint32_t>(24ULL)))) {
auto tmp104 = xbox_Machine_volumeUnits(m,fo);
uint64_t total = std::get<0>(tmp104);
uint64_t avail = std::get<1>(tmp104);
xbox_Machine_write32(m,cast<uint32_t>((buf + cast<uint32_t>(0ULL))),cast<uint32_t>(total));
xbox_Machine_write32(m,cast<uint32_t>((buf + cast<uint32_t>(4ULL))),cast<uint32_t>(shr<uint64_t>(total,cast<int64_t>(32ULL))));
xbox_Machine_write32(m,cast<uint32_t>((buf + cast<uint32_t>(8ULL))),cast<uint32_t>(avail));
xbox_Machine_write32(m,cast<uint32_t>((buf + cast<uint32_t>(12ULL))),cast<uint32_t>(shr<uint64_t>(avail,cast<int64_t>(32ULL))));
xbox_Machine_write32(m,cast<uint32_t>((buf + cast<uint32_t>(16ULL))),cast<uint32_t>(32ULL));
xbox_Machine_write32(m,cast<uint32_t>((buf + cast<uint32_t>(20ULL))),cast<uint32_t>(512ULL));
xbox_Machine_logf(m,std::string("NtQueryVolumeInformationFile: handle %08X size -> %d/%d units free from %08X",76),h,avail,total,xbox_Machine_retAddr(m));
xbox_Machine_finishOpen(m,iosb,h,cast<uint32_t>(24ULL),cast<uint32_t>(0ULL));
return cast<int64_t>(5ULL);
}
if (((class_ != cast<uint32_t>(1ULL)) || (ln < cast<uint32_t>(20ULL)))) {
x86_CPU_Halt(m->CPU,std::string("NtQueryVolumeInformationFile: unmodelled class %d (len %d) from %08X",68),class_,ln,xbox_Machine_retAddr(m));
return cast<int64_t>(5ULL);
}
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < cast<uint32_t>(24ULL));i += cast<uint32_t>(4ULL)){
xbox_Machine_write32(m,cast<uint32_t>((buf + i)),cast<uint32_t>(0ULL));
}
}if (((!fo->cache) && bool(m->Disc))) {
xbox_Machine_write32(m,cast<uint32_t>((buf + cast<uint32_t>(0ULL))),cast<uint32_t>(m->Disc->CreationTime));
xbox_Machine_write32(m,cast<uint32_t>((buf + cast<uint32_t>(4ULL))),cast<uint32_t>(shr<uint64_t>(m->Disc->CreationTime,cast<int64_t>(32ULL))));
}
xbox_Machine_write32(m,cast<uint32_t>((buf + cast<uint32_t>(8ULL))),cast<uint32_t>(3512491585ULL));
xbox_Machine_write32(m,cast<uint32_t>((buf + cast<uint32_t>(12ULL))),cast<uint32_t>(0ULL));
xbox_Machine_write32(m,cast<uint32_t>((buf + cast<uint32_t>(16ULL))),cast<uint32_t>(0ULL));
xbox_Machine_logf(m,std::string("NtQueryVolumeInformationFile: handle %08X class %d from %08X (tick %d)",70),h,class_,xbox_Machine_retAddr(m),m->tick);
xbox_Machine_finishOpen(m,iosb,h,cast<uint32_t>(20ULL),cast<uint32_t>(0ULL));
return cast<int64_t>(5ULL);
}
;
break;}
case cast<uint16_t>(165ULL):{
return [=](xbox_Machine* m)->int64_t{
xbox_Machine_setRet(m,xbox_Machine_allocPool(m,xbox_Machine_arg(m,cast<int64_t>(0ULL))));
return cast<int64_t>(1ULL);
}
;
break;}
case cast<uint16_t>(166ULL):{
return [=](xbox_Machine* m)->int64_t{
if ((xbox_Machine_arg(m,cast<int64_t>(0ULL)) >= cast<uint32_t>(1048576ULL))) {
xbox_Machine_logf(m,std::string("MmAllocateContiguousMemoryEx: %X bytes (lo=%X hi=%X align=%X prot=%X) from %08X (caller %08X)",93),xbox_Machine_arg(m,cast<int64_t>(0ULL)),xbox_Machine_arg(m,cast<int64_t>(1ULL)),xbox_Machine_arg(m,cast<int64_t>(2ULL)),xbox_Machine_arg(m,cast<int64_t>(3ULL)),xbox_Machine_arg(m,cast<int64_t>(4ULL)),xbox_Machine_retAddr(m),xbox_Machine_read32(m,cast<uint32_t>((m->CPU->Regs[cast<int64_t>(5ULL)] + cast<uint32_t>(4ULL)))));
}
xbox_Machine_setRet(m,xbox_Machine_allocPoolAligned(m,xbox_Machine_arg(m,cast<int64_t>(0ULL)),xbox_Machine_arg(m,cast<int64_t>(3ULL))));
return cast<int64_t>(5ULL);
}
;
break;}
case cast<uint16_t>(168ULL):{
return [=](xbox_Machine* m)->int64_t{
uint32_t size = xbox_Machine_arg(m,cast<int64_t>(0ULL));
uint32_t base = xbox_Machine_allocPool(m,size);
{
uint32_t p = xbox_Machine_arg(m,cast<int64_t>(1ULL));
if ((p != cast<uint32_t>(0ULL))) {
xbox_Machine_write32(m,p,cast<uint32_t>(0ULL));
}
}
if ((base == cast<uint32_t>(0ULL))) {
xbox_Machine_setRet(m,cast<uint32_t>(0ULL));
}
else {
xbox_Machine_setRet(m,cast<uint32_t>((base + xbox_align32(size,cast<uint32_t>(16ULL)))));
}
return cast<int64_t>(2ULL);
}
;
break;}
case cast<uint16_t>(65ULL):{
return [=](xbox_Machine* m)->int64_t{
uint32_t extSize = xbox_Machine_arg(m,cast<int64_t>(1ULL));
uint32_t out = xbox_Machine_arg(m,cast<int64_t>(5ULL));
constexpr int64_t hdr=64ULL;
uint32_t dev = xbox_Machine_allocPool(m,cast<uint32_t>((cast<uint32_t>(64ULL) + extSize)));
if ((dev == cast<uint32_t>(0ULL))) {
xbox_Machine_setRet(m,cast<uint32_t>(3221225626ULL));
return cast<int64_t>(6ULL);
}
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < cast<uint32_t>((cast<uint32_t>(64ULL) + extSize)));i += cast<uint32_t>(4ULL)){
xbox_Machine_write32(m,cast<uint32_t>((dev + i)),cast<uint32_t>(0ULL));
}
}xbox_Machine_write16(m,cast<uint32_t>((dev + cast<uint32_t>(0ULL))),cast<uint16_t>(3ULL));
xbox_Machine_write16(m,cast<uint32_t>((dev + cast<uint32_t>(2ULL))),cast<uint16_t>(64ULL));
xbox_Machine_write32(m,cast<uint32_t>((dev + cast<uint32_t>(24ULL))),cast<uint32_t>((dev + cast<uint32_t>(64ULL))));
if ((out != cast<uint32_t>(0ULL))) {
xbox_Machine_write32(m,out,dev);
}
std::string name = std::string("",0);
{
uint32_t ns = xbox_Machine_arg(m,cast<int64_t>(2ULL));
if ((ns != cast<uint32_t>(0ULL))) {
name = xbox_Machine_cstr(m,xbox_Machine_read32(m,cast<uint32_t>((ns + cast<uint32_t>(4ULL)))));
}
}
xbox_Machine_logf(m,std::string("IoCreateDevice: %q type %02X ext %d -> %08X",43),name,xbox_Machine_arg(m,cast<int64_t>(3ULL)),extSize,dev);
xbox_Machine_setRet(m,cast<uint32_t>(0ULL));
return cast<int64_t>(6ULL);
}
;
break;}
case cast<uint16_t>(173ULL):{
return [=](xbox_Machine* m)->int64_t{
uint32_t va = xbox_Machine_arg(m,cast<int64_t>(0ULL));
auto tmp105 = xbox_Machine_translate(m,va);
uint32_t phys = std::get<0>(tmp105);
bool mmio = std::get<1>(tmp105);
bool ok = std::get<2>(tmp105);
if (((!ok) || mmio)) {
x86_CPU_Halt(m->CPU,std::string("MmGetPhysicalAddress of non-RAM address %08X (from %08X)",56),va,xbox_Machine_retAddr(m));
return cast<int64_t>(1ULL);
}
xbox_Machine_setRet(m,phys);
return cast<int64_t>(1ULL);
}
;
break;}
case cast<uint16_t>(175ULL):{
return [=](xbox_Machine* m)->int64_t{
xbox_Machine_setRet(m,cast<uint32_t>(0ULL));
return cast<int64_t>(3ULL);
}
;
break;}
case cast<uint16_t>(180ULL):{
return [=](xbox_Machine* m)->int64_t{
xbox_Machine_setRet(m,get(m->poolSizes,xbox_Machine_arg(m,cast<int64_t>(0ULL))));
return cast<int64_t>(1ULL);
}
;
break;}
case cast<uint16_t>(181ULL):{
return [=](xbox_Machine* m)->int64_t{
uint32_t p = xbox_Machine_arg(m,cast<int64_t>(0ULL));
if ((p != cast<uint32_t>(0ULL))) {
constexpr int64_t pageShift=12ULL;
uint32_t total = cast<uint32_t>(16384ULL);
uint32_t free = cast<uint32_t>(0ULL);
if ((m->poolNext > m->heapNext)) {
free = shr<uint32_t>((cast<uint32_t>((m->poolNext - m->heapNext))),cast<int64_t>(12ULL));
}
uint32_t poolPages = shr<uint32_t>((cast<uint32_t>((cast<uint32_t>(66846720ULL) - m->poolNext))),cast<int64_t>(12ULL));
xbox_Machine_write32(m,cast<uint32_t>((p + cast<uint32_t>(0ULL))),cast<uint32_t>(36ULL));
xbox_Machine_write32(m,cast<uint32_t>((p + cast<uint32_t>(4ULL))),total);
xbox_Machine_write32(m,cast<uint32_t>((p + cast<uint32_t>(8ULL))),free);
xbox_Machine_write32(m,cast<uint32_t>((p + cast<uint32_t>(12ULL))),cast<uint32_t>((total - free)));
xbox_Machine_write32(m,cast<uint32_t>((p + cast<uint32_t>(16ULL))),cast<uint32_t>(0ULL));
xbox_Machine_write32(m,cast<uint32_t>((p + cast<uint32_t>(20ULL))),cast<uint32_t>(0ULL));
xbox_Machine_write32(m,cast<uint32_t>((p + cast<uint32_t>(24ULL))),poolPages);
xbox_Machine_write32(m,cast<uint32_t>((p + cast<uint32_t>(28ULL))),cast<uint32_t>(0ULL));
xbox_Machine_write32(m,cast<uint32_t>((p + cast<uint32_t>(32ULL))),cast<uint32_t>(0ULL));
}
xbox_Machine_setRet(m,cast<uint32_t>(0ULL));
return cast<int64_t>(1ULL);
}
;
break;}
case cast<uint16_t>(182ULL):{
return [=](xbox_Machine* m)->int64_t{
xbox_Machine_setRet(m,cast<uint32_t>(0ULL));
return cast<int64_t>(3ULL);
}
;
break;}
case cast<uint16_t>(184ULL):{
return [=](xbox_Machine* m)->int64_t{
auto tmp106 = std::make_tuple(xbox_Machine_arg(m,cast<int64_t>(0ULL)),xbox_Machine_arg(m,cast<int64_t>(2ULL)));
uint32_t baseOut = std::get<0>(tmp106);
uint32_t sizeP = std::get<1>(tmp106);
uint32_t size = xbox_Machine_read32(m,sizeP);
uint32_t addr = cast<uint32_t>(0ULL);
if (((baseOut != cast<uint32_t>(0ULL)) && (xbox_Machine_read32(m,baseOut) != cast<uint32_t>(0ULL)))) {
addr = xbox_Machine_read32(m,baseOut);
}
else {
addr = xbox_Machine_allocVirtual(m,size);
xbox_Machine_logf(m,std::string("NtAllocateVirtualMemory: %X bytes type=%X prot=%X -> %08X (from %08X)",69),size,xbox_Machine_arg(m,cast<int64_t>(3ULL)),xbox_Machine_arg(m,cast<int64_t>(4ULL)),addr,xbox_Machine_retAddr(m));
}
if ((baseOut != cast<uint32_t>(0ULL))) {
xbox_Machine_write32(m,baseOut,addr);
}
if ((sizeP != cast<uint32_t>(0ULL))) {
xbox_Machine_write32(m,sizeP,xbox_align32(size,cast<uint32_t>(4096ULL)));
}
xbox_Machine_setRet(m,cast<uint32_t>(0ULL));
return cast<int64_t>(5ULL);
}
;
break;}
case cast<uint16_t>(189ULL):{
return [=](xbox_Machine* m)->int64_t{
uint32_t handleOut = xbox_Machine_arg(m,cast<int64_t>(0ULL));
auto tmp107 = std::make_tuple(std::string("event",5),cast<uint8_t>(0ULL));
std::string kind = std::get<0>(tmp107);
uint8_t typ = std::get<1>(tmp107);
if ((xbox_Machine_arg(m,cast<int64_t>(2ULL)) == cast<uint32_t>(1ULL))) {
auto tmp108 = std::make_tuple(std::string("event-auto",10),cast<uint8_t>(1ULL));
kind = std::get<0>(tmp108);
typ = std::get<1>(tmp108);
}
uint32_t h = xbox_Machine_newObject(m,kind,typ,(xbox_Machine_arg(m,cast<int64_t>(3ULL)) != cast<uint32_t>(0ULL)),cast<int32_t>(0ULL),cast<int32_t>(0ULL));
if ((handleOut != cast<uint32_t>(0ULL))) {
xbox_Machine_write32(m,handleOut,h);
}
xbox_Machine_setRet(m,cast<uint32_t>(0ULL));
return cast<int64_t>(4ULL);
}
;
break;}
case cast<uint16_t>(193ULL):{
return [=](xbox_Machine* m)->int64_t{
uint32_t handleOut = xbox_Machine_arg(m,cast<int64_t>(0ULL));
int32_t initial = cast<int32_t>(xbox_Machine_arg(m,cast<int64_t>(2ULL)));
int32_t limit = cast<int32_t>(xbox_Machine_arg(m,cast<int64_t>(3ULL)));
uint32_t h = xbox_Machine_newObject(m,std::string("semaphore",9),cast<uint8_t>(5ULL),(initial > cast<int32_t>(0ULL)),initial,limit);
if ((handleOut != cast<uint32_t>(0ULL))) {
xbox_Machine_write32(m,handleOut,h);
}
xbox_Machine_setRet(m,cast<uint32_t>(0ULL));
return cast<int64_t>(4ULL);
}
;
break;}
case cast<uint16_t>(222ULL):{
return [=](xbox_Machine* m)->int64_t{
auto tmp109 = std::make_tuple(xbox_Machine_arg(m,cast<int64_t>(0ULL)),xbox_Machine_arg(m,cast<int64_t>(1ULL)),xbox_Machine_arg(m,cast<int64_t>(2ULL)));
uint32_t handle = std::get<0>(tmp109);
uint32_t release = std::get<1>(tmp109);
uint32_t prevOut = std::get<2>(tmp109);
xbox_Machine_logf(m,std::string("NtReleaseSemaphore: handle %08X +%d from %08X (tick %d)",55),handle,release,xbox_Machine_retAddr(m),m->tick);
{
xbox_kobject* o = xbox_Machine_objAt(m,handle);
if (bool(o)) {
if ((prevOut != cast<uint32_t>(0ULL))) {
xbox_Machine_write32(m,prevOut,cast<uint32_t>(o->count));
}
o->count += cast<int32_t>(release);
o->signaled = (o->count > cast<int32_t>(0ULL));
xbox_Machine_writeSignal(m,o->addr,o->signaled);
xbox_kobject_noteSignal(o,std::string("NtReleaseSemaphore",18),xbox_Machine_retAddr(m),m->tick);
xbox_Machine_wakeWaiters(m,handle);
}
else if ((prevOut != cast<uint32_t>(0ULL))) {
xbox_Machine_write32(m,prevOut,cast<uint32_t>(0ULL));
}
}
xbox_Machine_setRet(m,cast<uint32_t>(0ULL));
return cast<int64_t>(3ULL);
}
;
break;}
case cast<uint16_t>(225ULL):{
return [=](xbox_Machine* m)->int64_t{
auto tmp110 = std::make_tuple(xbox_Machine_arg(m,cast<int64_t>(0ULL)),xbox_Machine_arg(m,cast<int64_t>(1ULL)));
uint32_t h = std::get<0>(tmp110);
uint32_t prevOut = std::get<1>(tmp110);
xbox_kobject* o = xbox_Machine_objAt(m,h);
if ((!o)) {
x86_CPU_Halt(m->CPU,std::string("NtSetEvent: no object for handle %08X, from %08X",48),h,xbox_Machine_retAddr(m));
return cast<int64_t>(2ULL);
}
if ((prevOut != cast<uint32_t>(0ULL))) {
xbox_Machine_write32(m,prevOut,xbox_boolU32(o->signaled));
}
o->signaled = true;
xbox_Machine_writeSignal(m,o->addr,true);
xbox_kobject_noteSignal(o,std::string("NtSetEvent",10),xbox_Machine_retAddr(m),m->tick);
xbox_Machine_wakeWaiters(m,h);
xbox_Machine_setRet(m,cast<uint32_t>(0ULL));
return cast<int64_t>(2ULL);
}
;
break;}
case cast<uint16_t>(234ULL):{
return [=](xbox_Machine* m)->int64_t{
xbox_Machine_doWaitTimed(m,xbox_Machine_arg(m,cast<int64_t>(0ULL)),xbox_Machine_arg(m,cast<int64_t>(3ULL)));
return cast<int64_t>(4ULL);
}
;
break;}
case cast<uint16_t>(199ULL):{
return [=](xbox_Machine* m)->int64_t{
xbox_Machine_setRet(m,cast<uint32_t>(0ULL));
return cast<int64_t>(3ULL);
}
;
break;}
case cast<uint16_t>(187ULL):{
return [=](xbox_Machine* m)->int64_t{
uint32_t h = xbox_Machine_arg(m,cast<int64_t>(0ULL));
removeKey(m->files,h);
xbox_Machine_setRet(m,cast<uint32_t>(0ULL));
return cast<int64_t>(1ULL);
}
;
break;}
case cast<uint16_t>(46ULL):{
return [=](xbox_Machine* m)->int64_t{
auto tmp111 = std::make_tuple(xbox_Machine_arg(m,cast<int64_t>(0ULL)),xbox_Machine_arg(m,cast<int64_t>(1ULL)),xbox_Machine_arg(m,cast<int64_t>(2ULL)));
uint32_t bus = std::get<0>(tmp111);
uint32_t slot = std::get<1>(tmp111);
uint32_t reg = std::get<2>(tmp111);
auto tmp112 = std::make_tuple(xbox_Machine_arg(m,cast<int64_t>(3ULL)),xbox_Machine_arg(m,cast<int64_t>(4ULL)),xbox_Machine_arg(m,cast<int64_t>(5ULL)));
uint32_t buf = std::get<0>(tmp112);
uint32_t length = std::get<1>(tmp112);
uint32_t write = std::get<2>(tmp112);
uint32_t base = cast<uint32_t>((shl<uint32_t>(bus,cast<int64_t>(24ULL)) | shl<uint32_t>(slot,cast<int64_t>(16ULL))));
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < length);i++){
uint32_t key = cast<uint32_t>((base | (cast<uint32_t>(((cast<uint32_t>((reg + i))) & cast<uint32_t>(65535ULL))))));
if ((write != cast<uint32_t>(0ULL))) {
m->pciSpace[key] = xbox_Machine_Read(m,cast<uint32_t>((buf + i)));
}
else {
xbox_Machine_Write(m,cast<uint32_t>((buf + i)),get(m->pciSpace,key));
}
}
}xbox_Machine_setRet(m,cast<uint32_t>(0ULL));
return cast<int64_t>(6ULL);
}
;
break;}
case cast<uint16_t>(44ULL):{
return [=](xbox_Machine* m)->int64_t{
uint32_t level = xbox_Machine_arg(m,cast<int64_t>(0ULL));
{
uint32_t p = xbox_Machine_arg(m,cast<int64_t>(1ULL));
if ((p != cast<uint32_t>(0ULL))) {
xbox_Machine_Write(m,p,cast<uint8_t>(level));
}
}
xbox_Machine_setRet(m,level);
return cast<int64_t>(2ULL);
}
;
break;}
case cast<uint16_t>(109ULL):{
return [=](xbox_Machine* m)->int64_t{
uint32_t ki = xbox_Machine_arg(m,cast<int64_t>(0ULL));
if ((ki != cast<uint32_t>(0ULL))) {
xbox_Machine_write32(m,cast<uint32_t>((ki + cast<uint32_t>(0ULL))),xbox_Machine_arg(m,cast<int64_t>(1ULL)));
xbox_Machine_write32(m,cast<uint32_t>((ki + cast<uint32_t>(4ULL))),xbox_Machine_arg(m,cast<int64_t>(2ULL)));
xbox_Machine_write32(m,cast<uint32_t>((ki + cast<uint32_t>(8ULL))),xbox_Machine_arg(m,cast<int64_t>(3ULL)));
}
xbox_Machine_setRet(m,cast<uint32_t>(0ULL));
return cast<int64_t>(7ULL);
}
;
break;}
case cast<uint16_t>(98ULL):{
return [=](xbox_Machine* m)->int64_t{
uint32_t ki = xbox_Machine_arg(m,cast<int64_t>(0ULL));
if ((ki != cast<uint32_t>(0ULL))) {
uint32_t vec = xbox_Machine_read32(m,cast<uint32_t>((ki + cast<uint32_t>(8ULL))));
m->interrupts[vec] = ki;
xbox_Machine_logf(m,std::string("KeConnectInterrupt: vector %d -> KINTERRUPT %08X (routine %08X ctx %08X)",72),vec,ki,xbox_Machine_read32(m,ki),xbox_Machine_read32(m,cast<uint32_t>((ki + cast<uint32_t>(4ULL)))));
}
xbox_Machine_setRet(m,cast<uint32_t>(1ULL));
return cast<int64_t>(1ULL);
}
;
break;}
case cast<uint16_t>(99ULL):{
return [=](xbox_Machine* m)->int64_t{
xbox_Machine_setRet(m,xbox_Machine_currentKThread(m));
return cast<int64_t>(0ULL);
}
;
break;}
case cast<uint16_t>(119ULL):{
return [=](xbox_Machine* m)->int64_t{
uint32_t dpc = xbox_Machine_arg(m,cast<int64_t>(0ULL));
{auto&& tmp113 = m->dpcQueue;
for(int64_t tmp114=0;tmp114<len(tmp113);++tmp114){
auto d=tmp113[tmp114];if ((d.Dpc == dpc)) {
xbox_Machine_setRet(m,cast<uint32_t>(0ULL));
return cast<int64_t>(3ULL);
}
}}
m->dpcQueue = append(m->dpcQueue,Slice<xbox_dpcEntry>{xbox_dpcEntry{dpc,xbox_Machine_arg(m,cast<int64_t>(1ULL)),xbox_Machine_arg(m,cast<int64_t>(2ULL))}});
xbox_Machine_setRet(m,cast<uint32_t>(1ULL));
return cast<int64_t>(3ULL);
}
;
break;}
case cast<uint16_t>(145ULL):{
return [=](xbox_Machine* m)->int64_t{
xbox_kobject* o = xbox_Machine_guestObjAt(m,xbox_Machine_arg(m,cast<int64_t>(0ULL)));
int32_t prev = cast<int32_t>(0ULL);
if (bool(o)) {
if (o->signaled) {
prev = cast<int32_t>(1ULL);
}
o->signaled = true;
xbox_Machine_writeSignal(m,o->addr,true);
xbox_kobject_noteSignal(o,std::string("KeSetEvent",10),xbox_Machine_retAddr(m),m->tick);
xbox_Machine_wakeWaiters(m,o->addr);
}
xbox_Machine_setRet(m,cast<uint32_t>(prev));
return cast<int64_t>(3ULL);
}
;
break;}
case cast<uint16_t>(159ULL):{
return [=](xbox_Machine* m)->int64_t{
xbox_Machine_guestObjAt(m,xbox_Machine_arg(m,cast<int64_t>(0ULL)));
xbox_Machine_doWaitTimed(m,xbox_Machine_arg(m,cast<int64_t>(0ULL)),xbox_Machine_arg(m,cast<int64_t>(4ULL)));
return cast<int64_t>(5ULL);
}
;
break;}
case cast<uint16_t>(107ULL):{
return [=](xbox_Machine* m)->int64_t{
uint32_t dpc = xbox_Machine_arg(m,cast<int64_t>(0ULL));
if ((dpc != cast<uint32_t>(0ULL))) {
xbox_Machine_write16(m,cast<uint32_t>((dpc + cast<uint32_t>(0ULL))),cast<uint16_t>(19ULL));
xbox_Machine_write32(m,cast<uint32_t>((dpc + cast<uint32_t>(4ULL))),cast<uint32_t>(0ULL));
xbox_Machine_write32(m,cast<uint32_t>((dpc + cast<uint32_t>(8ULL))),cast<uint32_t>(0ULL));
xbox_Machine_write32(m,cast<uint32_t>((dpc + cast<uint32_t>(12ULL))),xbox_Machine_arg(m,cast<int64_t>(1ULL)));
xbox_Machine_write32(m,cast<uint32_t>((dpc + cast<uint32_t>(16ULL))),xbox_Machine_arg(m,cast<int64_t>(2ULL)));
}
xbox_Machine_setRet(m,cast<uint32_t>(0ULL));
return cast<int64_t>(3ULL);
}
;
break;}
case cast<uint16_t>(149ULL):{
return [=](xbox_Machine* m)->int64_t{
bool was = xbox_Machine_armTimer(m,xbox_Machine_arg(m,cast<int64_t>(0ULL)),xbox_Machine_arg(m,cast<int64_t>(3ULL)),xbox_Machine_arg(m,cast<int64_t>(1ULL)),xbox_Machine_arg(m,cast<int64_t>(2ULL)),cast<uint32_t>(0ULL));
xbox_Machine_setRet(m,xbox_b2u(was));
return cast<int64_t>(4ULL);
}
;
break;}
case cast<uint16_t>(150ULL):{
return [=](xbox_Machine* m)->int64_t{
bool was = xbox_Machine_armTimer(m,xbox_Machine_arg(m,cast<int64_t>(0ULL)),xbox_Machine_arg(m,cast<int64_t>(4ULL)),xbox_Machine_arg(m,cast<int64_t>(1ULL)),xbox_Machine_arg(m,cast<int64_t>(2ULL)),xbox_Machine_arg(m,cast<int64_t>(3ULL)));
xbox_Machine_setRet(m,xbox_b2u(was));
return cast<int64_t>(5ULL);
}
;
break;}
case cast<uint16_t>(113ULL):{
return [=](xbox_Machine* m)->int64_t{
uint32_t tm = xbox_Machine_arg(m,cast<int64_t>(0ULL));
if ((tm != cast<uint32_t>(0ULL))) {
xbox_Machine_write16(m,cast<uint32_t>((tm + cast<uint32_t>(0ULL))),cast<uint16_t>(8ULL));
xbox_Machine_write32(m,cast<uint32_t>((tm + cast<uint32_t>(4ULL))),cast<uint32_t>(0ULL));
xbox_Machine_write32(m,cast<uint32_t>((tm + cast<uint32_t>(12ULL))),cast<uint32_t>(0ULL));
xbox_Machine_write32(m,cast<uint32_t>((tm + cast<uint32_t>(16ULL))),cast<uint32_t>(0ULL));
}
xbox_Machine_setRet(m,cast<uint32_t>(0ULL));
return cast<int64_t>(2ULL);
}
;
break;}
case cast<uint16_t>(97ULL):{
return [=](xbox_Machine* m)->int64_t{
uint32_t tm = xbox_Machine_arg(m,cast<int64_t>(0ULL));
bool was = xbox_Machine_cancelTimer(m,tm);
if ((tm != cast<uint32_t>(0ULL))) {
xbox_Machine_write32(m,cast<uint32_t>((tm + cast<uint32_t>(4ULL))),cast<uint32_t>(0ULL));
}
xbox_Machine_setRet(m,xbox_b2u(was));
return cast<int64_t>(1ULL);
}
;
break;}
case cast<uint16_t>(137ULL):{
return [=](xbox_Machine* m)->int64_t{
uint32_t dpc = xbox_Machine_arg(m,cast<int64_t>(0ULL));
bool found = false;
{auto&& tmp115 = m->dpcQueue;
for(int64_t tmp116=0;tmp116<len(tmp115);++tmp116){
auto i=tmp116;auto d=tmp115[tmp116];if ((d.Dpc == dpc)) {
m->dpcQueue = append(sub(m->dpcQueue,0,i),sub(m->dpcQueue,cast<int64_t>((i + cast<int64_t>(1ULL))),len(m->dpcQueue)));
found = true;
break;
}
}}
xbox_Machine_setRet(m,xbox_b2u(found));
return cast<int64_t>(1ULL);
}
;
break;}
case cast<uint16_t>(17ULL):{
return [=](xbox_Machine* m)->int64_t{
xbox_Machine_setRet(m,cast<uint32_t>(0ULL));
return cast<int64_t>(1ULL);
}
;
break;}
case cast<uint16_t>(277ULL):{
return [=](xbox_Machine* m)->int64_t{
xbox_Machine_setRet(m,cast<uint32_t>(0ULL));
return cast<int64_t>(1ULL);
}
;
break;}
case cast<uint16_t>(291ULL):{
return [=](xbox_Machine* m)->int64_t{
uint32_t cs = xbox_Machine_arg(m,cast<int64_t>(0ULL));
if ((cs != cast<uint32_t>(0ULL))) {
xbox_Machine_write32(m,cast<uint32_t>((cs + cast<uint32_t>(0ULL))),cast<uint32_t>(0ULL));
xbox_Machine_write32(m,cast<uint32_t>((cs + cast<uint32_t>(4ULL))),cast<uint32_t>(4294967295ULL));
xbox_Machine_write32(m,cast<uint32_t>((cs + cast<uint32_t>(8ULL))),cast<uint32_t>(0ULL));
xbox_Machine_write32(m,cast<uint32_t>((cs + cast<uint32_t>(12ULL))),cast<uint32_t>(0ULL));
}
xbox_Machine_setRet(m,cast<uint32_t>(0ULL));
return cast<int64_t>(1ULL);
}
;
break;}
case cast<uint16_t>(294ULL):{
return [=](xbox_Machine* m)->int64_t{
xbox_Machine_setRet(m,cast<uint32_t>(0ULL));
return cast<int64_t>(1ULL);
}
;
break;}
case cast<uint16_t>(301ULL):{
return [=](xbox_Machine* m)->int64_t{
uint32_t st = xbox_Machine_arg(m,cast<int64_t>(0ULL));
uint32_t w={};
{
switch(st){
case cast<uint32_t>(0ULL):{
w = cast<uint32_t>(0ULL);
break;}
case cast<uint32_t>(258ULL):{
w = cast<uint32_t>(258ULL);
break;}
case cast<uint32_t>(259ULL):{
w = cast<uint32_t>(997ULL);
break;}
case cast<uint32_t>(3221225480ULL):{
w = cast<uint32_t>(6ULL);
break;}
case cast<uint32_t>(3221225489ULL):{
w = cast<uint32_t>(38ULL);
break;}
case cast<uint32_t>(3221225524ULL):{
w = cast<uint32_t>(2ULL);
break;}
case cast<uint32_t>(3221225525ULL):{
w = cast<uint32_t>(183ULL);
break;}
case cast<uint32_t>(3221225530ULL):{
w = cast<uint32_t>(3ULL);
break;}
default:{
w = cast<uint32_t>(317ULL);
if ((!get(m->dosErrWarned,st))) {
if ((!m->dosErrWarned)) {
m->dosErrWarned = Map<uint32_t,bool>{};
}
m->dosErrWarned[st] = true;
xbox_Machine_logf(m,std::string("RtlNtStatusToDosError: unmapped NTSTATUS %08X -> 317 (from %08X)",64),st,xbox_Machine_retAddr(m));
}
break;}
}}
xbox_Machine_setRet(m,w);
return cast<int64_t>(1ULL);
}
;
break;}
}}
return {};
}
}
// tools/platform/xbox/kernel_objects.go:1334:1
void xbox_Machine_doWait(xbox_Machine* m,uint32_t handle,int64_t reg){
{
xbox_kobject* o = xbox_Machine_objAt(m,handle);
if (xbox_Machine_satisfyWait(m,o)) {
xbox_Machine_setRet(m,cast<uint32_t>(0ULL));
return ;
}
if ((!m->current)) {
xbox_Machine_setRet(m,cast<uint32_t>(0ULL));
return ;
}
m->current->waitObjs = Slice<uint32_t>{handle};
xbox_Machine_setRet(m,cast<uint32_t>(0ULL));
xbox_Machine_yieldCurrent(m,cast<xbox_threadState>(2ULL));
}
}
// tools/platform/xbox/kernel_objects.go:1358:1
void xbox_Machine_doWaitTimed(xbox_Machine* m,uint32_t handle,uint32_t timeoutPtr){
{
xbox_kobject* o = xbox_Machine_objAt(m,handle);
if (xbox_Machine_satisfyWait(m,o)) {
xbox_Machine_setRet(m,cast<uint32_t>(0ULL));
return ;
}
if ((!m->current)) {
xbox_Machine_setRet(m,cast<uint32_t>(0ULL));
return ;
}
uint64_t wake = cast<uint64_t>(0ULL);
if ((timeoutPtr != cast<uint32_t>(0ULL))) {
int64_t v = cast<int64_t>(cast<uint64_t>((shl<uint64_t>(cast<uint64_t>(xbox_Machine_read32(m,cast<uint32_t>((timeoutPtr + cast<uint32_t>(4ULL))))),cast<int64_t>(32ULL)) | cast<uint64_t>(xbox_Machine_read32(m,timeoutPtr)))));
{
if ((v < cast<int64_t>(0ULL))){
wake = cast<uint64_t>((cast<uint64_t>((m->tick + cast<uint64_t>((divi<uint64_t>(cast<uint64_t>(cast<int64_t>(-v)),cast<uint64_t>(10000ULL)) * cast<uint64_t>(733466ULL))))) + cast<uint64_t>(1ULL)));
}
else if ((v == cast<int64_t>(0ULL))){
xbox_Machine_setRet(m,cast<uint32_t>(258ULL));
return ;
}
else {
uint64_t now = xbox_Machine_systemTime100ns(m);
if ((cast<uint64_t>(v) <= now)) {
xbox_Machine_setRet(m,cast<uint32_t>(258ULL));
return ;
}
wake = cast<uint64_t>((cast<uint64_t>((m->tick + cast<uint64_t>((divi<uint64_t>((cast<uint64_t>((cast<uint64_t>(v) - now))),cast<uint64_t>(10000ULL)) * cast<uint64_t>(733466ULL))))) + cast<uint64_t>(1ULL)));
}
}
tmp117:;
}
xbox_thread* t = m->current;
t->waitObjs = Slice<uint32_t>{handle};
t->wakeTick = wake;
std::string kind = std::string("guest",5);
if (bool(o)) {
kind = o->kind;
}
xbox_Machine_logf(m,std::string("wait: tid=%d parks on %08X (%s) wake=%d from %08X",49),t->id,handle,kind,wake,xbox_Machine_retAddr(m));
xbox_Machine_setRet(m,cast<uint32_t>(258ULL));
xbox_Machine_yieldCurrent(m,cast<xbox_threadState>(2ULL));
}
}
// tools/platform/xbox/kernel_objects.go:1400:1
void xbox_Machine_wakeWaiters(xbox_Machine* m,uint32_t handle){
{
xbox_kobject* o = get(m->objects,handle);
{auto&& tmp118 = m->threads;
for(int64_t tmp119=0;tmp119<len(tmp118);++tmp119){
auto t=tmp118[tmp119];if ((t->state != cast<xbox_threadState>(2ULL))) {
continue;
}
{auto&& tmp120 = t->waitObjs;
for(int64_t tmp121=0;tmp121<len(tmp120);++tmp121){
auto h=tmp120[tmp121];if (((h == handle) && xbox_Machine_satisfyWait(m,o))) {
t->state = cast<xbox_threadState>(0ULL);
t->waitObjs = {};
t->wakeTick = cast<uint64_t>(0ULL);
t->ctx.Regs[cast<int64_t>(0ULL)] = cast<uint32_t>(0ULL);
break;
}
}}
}}
}
}
// tools/platform/xbox/kernel_objects.go:1420:1
uint32_t xbox_boolU32(bool b){
{
if (b) {
return cast<uint32_t>(1ULL);
}
return cast<uint32_t>(0ULL);
}
}
// tools/platform/xbox/kernel_ordinals.go:19:1
std::string xbox_ordinalName(uint16_t ord){
{
{
auto tmp122 = lookup(xbox_verifiedNames,ord);
std::string n = std::get<0>(tmp122);
bool ok = std::get<1>(tmp122);
if (ok) {
return n;
}
}
{
auto tmp123 = lookup(xbox_ordinalNames,ord);
std::string n = std::get<0>(tmp123);
bool ok = std::get<1>(tmp123);
if (ok) {
return n;
}
}
return go_fmt_Sprintf(std::string("ordinal_%d",10),ord);
}
}
// tools/platform/xbox/machine.go:272:1
std::tuple<xbox_Machine*,Error> xbox_NewMachine(xbox_XBE* xbe,xbox_Image* disc){
{
xbox_Machine* m = arenaNew(xbox_Machine{Slice<uint8_t>::make(cast<int64_t>(67108864ULL)),{},xbe,disc,{},{},{},Map<uint32_t,xbox_kobject*>{},Map<uint32_t,xbox_fileObject*>{},Map<std::string,xbox_cacheFile*>{},Map<std::string,Slice<uint8_t>>{},{},Map<uint32_t,uint32_t>{},{},{},{},{},{},{},{},{},Map<uint32_t,uint32_t>{},{},{},{},{},{},{},{},{},{},{},Map<uint32_t,Slice<uint8_t>>{},Map<uint32_t,Slice<uint8_t>>{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},Map<uint16_t,int64_t>{},Map<uint16_t,bool>{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{}});
m->FlipVSync = xbox_flipVSyncDefault;
m->nv.reg = Map<uint32_t,uint32_t>{};
m->apu = xbox_newMMIOLatch(std::string("APU",3));
m->ac97 = xbox_newMMIOLatch(std::string("AC97",4));
m->usb = xbox_newMMIOLatch(std::string("USB",3));
m->nic = xbox_newMMIOLatch(std::string("NIC",3));
m->pciSpace = Map<uint32_t,uint8_t>{};
m->pgraph = xbox_newPgraph(m);
{
Error err = xbox_Machine_loadImage(m);
if (bool(err)) {
return {{},err};
}
}
xbox_Machine_setupMemoryLayout(m);
xbox_Machine_setupKPCR(m);
xbox_Machine_patchThunks(m);
x86_CPU* c = x86_NewCPU(m);
c->Mode = cast<int64_t>(1ULL);
{auto&& tmp124 = c->SegBase;
for(int64_t tmp125=0;tmp125<len(tmp124);++tmp125){
auto i=tmp125;c->SegBase[i] = cast<uint32_t>(0ULL);
}}
c->Seg[cast<int64_t>(1ULL)] = cast<uint16_t>(8ULL);
c->Seg[cast<int64_t>(3ULL)] = cast<uint16_t>(16ULL);
c->Seg[cast<int64_t>(0ULL)] = cast<uint16_t>(16ULL);
c->Seg[cast<int64_t>(2ULL)] = cast<uint16_t>(16ULL);
c->Seg[cast<int64_t>(5ULL)] = cast<uint16_t>(16ULL);
c->Seg[cast<int64_t>(4ULL)] = cast<uint16_t>(56ULL);
c->SegBase[cast<int64_t>(4ULL)] = cast<uint32_t>(66846720ULL);
c->SegResolve = [=](auto...args){return xbox_Machine_resolveSel(m,args...);};
c->IP = xbe->Entry;
c->Regs[cast<int64_t>(4ULL)] = cast<uint32_t>(66842620ULL);
xbox_Machine_write32(m,cast<uint32_t>(66842620ULL),cast<uint32_t>(2399141632ULL));
c->IF = true;
c->TSCFunc = [=](auto...args){return xbox_Machine_guestTSC(m,args...);};
c->OnStep = [=](auto...args){return xbox_Machine_onStep(m,args...);};
c->PortIn = [=](auto...args){return xbox_Machine_portIn(m,args...);};
c->PortOut = [=](auto...args){return xbox_Machine_portOut(m,args...);};
m->CPU = c;
xbox_Machine_bootThread(m);
return {m,{}};
}
}
// tools/platform/xbox/machine.go:355:1
uint32_t xbox_Machine_resolveSel(xbox_Machine* m,uint16_t sel){
{
if ((sel == cast<uint16_t>(56ULL))) {
return cast<uint32_t>(66846720ULL);
}
return cast<uint32_t>(0ULL);
}
}
// tools/platform/xbox/machine.go:365:1
Error xbox_Machine_loadImage(xbox_Machine* m){
{
xbox_XBE* x = m->XBE;
{auto&& tmp126 = x->Sections;
for(int64_t tmp127=0;tmp127<len(tmp126);++tmp127){
auto s=tmp126[tmp127];if ((s.RawSize == cast<uint32_t>(0ULL))) {
continue;
}
auto tmp128 = xbox_XBE_atVA(x,s.VAddr);
int64_t off = std::get<0>(tmp128);
bool ok = std::get<1>(tmp128);
if ((!ok)) {
return go_fmt_Errorf(std::string("xbox: section %q VA %#x has no file mapping",43),s.Name,s.VAddr);
}
if ((cast<int64_t>((cast<int64_t>(s.VAddr) + cast<int64_t>(s.RawSize))) > len(m->RAM))) {
return go_fmt_Errorf(std::string("xbox: section %q at VA %#x..%#x exceeds 64 MB RAM",49),s.Name,s.VAddr,cast<uint32_t>((s.VAddr + s.RawSize)));
}
if ((cast<int64_t>((off + cast<int64_t>(s.RawSize))) > len(x->raw))) {
return go_fmt_Errorf(std::string("xbox: section %q raw range %#x..%#x overruns the %d-byte image",62),s.Name,off,cast<int64_t>((off + cast<int64_t>(s.RawSize))),len(x->raw));
}
gcopy(sub(m->RAM,s.VAddr,cast<uint32_t>((s.VAddr + s.RawSize))),sub(x->raw,off,cast<int64_t>((off + cast<int64_t>(s.RawSize)))));
}}
return {};
}
}
// tools/platform/xbox/machine.go:391:1
void xbox_Machine_setupMemoryLayout(xbox_Machine* m){
{
uint32_t end = cast<uint32_t>((m->XBE->Base + m->XBE->ImageSize));
m->heapNext = xbox_align32(end,cast<uint32_t>(4096ULL));
m->poolNext = cast<uint32_t>(66777088ULL);
m->heapTop = m->poolNext;
m->kbandNext = cast<uint32_t>(66847424ULL);
m->nextObjAddr = m->kbandNext;
}
}
// tools/platform/xbox/machine.go:405:1
std::tuple<uint32_t,bool,bool> xbox_Machine_translate(xbox_Machine* m,uint32_t a){
uint32_t phys{};
bool mmio{};
bool ok{};
{
{
if ((a < cast<uint32_t>(67108864ULL))){
return {a,false,true};
}
else if (((a >= cast<uint32_t>(2147483648ULL)) && (a < cast<uint32_t>(2214592512ULL)))){
return {cast<uint32_t>((a - cast<uint32_t>(2147483648ULL))),false,true};
}
else if (((a >= cast<uint32_t>(2952790016ULL)) && (a < cast<uint32_t>(3019898880ULL)))){
return {cast<uint32_t>((a - cast<uint32_t>(2952790016ULL))),false,true};
}
else if (((a >= cast<uint32_t>(3489660928ULL)) && (a < cast<uint32_t>(3556769792ULL)))){
return {cast<uint32_t>((a & cast<uint32_t>(67108863ULL))),false,true};
}
else if (((a >= cast<uint32_t>(4026531840ULL)) && (a < cast<uint32_t>(4093640704ULL)))){
return {cast<uint32_t>((a - cast<uint32_t>(4026531840ULL))),false,true};
}
else if (((a >= cast<uint32_t>(4244635648ULL)) && (a < cast<uint32_t>(4261412864ULL)))){
return {cast<uint32_t>((a - cast<uint32_t>(4244635648ULL))),true,true};
}
else {
return {cast<uint32_t>(0ULL),false,false};
}
}
tmp129:;
}
}
// tools/platform/xbox/machine.go:428:1
uint8_t xbox_Machine_Read(xbox_Machine* m,uint32_t a){
{
if (((a >= cast<uint32_t>(2399141888ULL)) && (a < cast<uint32_t>(2399150080ULL)))) {
uint16_t ord = cast<uint16_t>(divi<uint32_t>((cast<uint32_t>((a - cast<uint32_t>(2399141888ULL)))),cast<uint32_t>(16ULL)));
if ((!get(m->dataDeref,ord))) {
m->dataDeref[ord] = true;
xbox_Machine_logf(m,std::string("kernel: data export ordinal %d (%s) dereferenced -> 0",53),ord,xbox_ordinalName(ord));
}
return cast<uint8_t>(0ULL);
}
if (((a >= cast<uint32_t>(4269801472ULL)) && (a < cast<uint32_t>(4270325760ULL)))) {
return xbox_Machine_apuRead(m,cast<uint32_t>((a - cast<uint32_t>(4269801472ULL))));
}
if (((a >= cast<uint32_t>(4273995776ULL)) && (a < cast<uint32_t>(4273999872ULL)))) {
return xbox_Machine_ac97Read(m,cast<uint32_t>((a - cast<uint32_t>(4273995776ULL))));
}
if (((a >= cast<uint32_t>(4275044352ULL)) && (a < cast<uint32_t>(4275048448ULL)))) {
return xbox_Machine_usbRead(m,cast<uint32_t>((a - cast<uint32_t>(4275044352ULL))));
}
if (((a >= cast<uint32_t>(4277141504ULL)) && (a < cast<uint32_t>(4277145600ULL)))) {
return xbox_Machine_nicRead(m,cast<uint32_t>((a - cast<uint32_t>(4277141504ULL))));
}
auto tmp130 = xbox_Machine_translate(m,a);
uint32_t phys = std::get<0>(tmp130);
bool mmio = std::get<1>(tmp130);
bool ok = std::get<2>(tmp130);
if ((!ok)) {
xbox_Machine_fault(m,std::string("read",4),a);
return cast<uint8_t>(255ULL);
}
if (mmio) {
return xbox_Machine_nvRead(m,phys);
}
uint8_t v = m->RAM[phys];
if (((bool(m->onR) && (a >= m->wRLo)) && (a < m->wRHi))) {
m->onR(a,cast<uint32_t>(v),xbox_Machine_watchPC(m));
}
return v;
}
}
// tools/platform/xbox/machine.go:470:1
void xbox_Machine_Write(xbox_Machine* m,uint32_t a,uint8_t v){
{
if (((a >= cast<uint32_t>(2399141888ULL)) && (a < cast<uint32_t>(2399150080ULL)))) {
return ;
}
if (((a >= cast<uint32_t>(4269801472ULL)) && (a < cast<uint32_t>(4270325760ULL)))) {
xbox_Machine_apuWrite(m,cast<uint32_t>((a - cast<uint32_t>(4269801472ULL))),v);
return ;
}
if (((a >= cast<uint32_t>(4273995776ULL)) && (a < cast<uint32_t>(4273999872ULL)))) {
xbox_Machine_ac97Write(m,cast<uint32_t>((a - cast<uint32_t>(4273995776ULL))),v);
return ;
}
if (((a >= cast<uint32_t>(4275044352ULL)) && (a < cast<uint32_t>(4275048448ULL)))) {
xbox_Machine_usbWrite(m,cast<uint32_t>((a - cast<uint32_t>(4275044352ULL))),v);
return ;
}
if (((a >= cast<uint32_t>(4277141504ULL)) && (a < cast<uint32_t>(4277145600ULL)))) {
xbox_Machine_nicWrite(m,cast<uint32_t>((a - cast<uint32_t>(4277141504ULL))),v);
return ;
}
auto tmp131 = xbox_Machine_translate(m,a);
uint32_t phys = std::get<0>(tmp131);
bool mmio = std::get<1>(tmp131);
bool ok = std::get<2>(tmp131);
if ((!ok)) {
xbox_Machine_fault(m,std::string("write",5),a);
return ;
}
if (mmio) {
xbox_Machine_nvWrite(m,phys,v);
return ;
}
m->RAM[phys] = v;
if(rrcapture::trace.active)rrXboxWrite(m,m->RAM,phys);
if (((bool(m->onW) && (a >= m->wWLo)) && (a < m->wWHi))) {
m->onW(a,cast<uint32_t>(v),xbox_Machine_watchPC(m));
}
}
}
// tools/platform/xbox/machine.go:506:1
uint32_t xbox_Machine_watchPC(xbox_Machine* m){
{
if ((!m->CPU)) {
return cast<uint32_t>(0ULL);
}
return x86_CPU_LinearPC(m->CPU);
}
}
// tools/platform/xbox/machine.go:514:1
void xbox_Machine_SetWriteWatch(xbox_Machine* m,uint32_t lo,uint32_t hi,std::function<void(uint32_t,uint32_t,uint32_t)> cb){
{
auto tmp132 = std::make_tuple(lo,hi,cb);
m->wWLo = std::get<0>(tmp132);
m->wWHi = std::get<1>(tmp132);
m->onW = std::get<2>(tmp132);
}
}
// tools/platform/xbox/machine.go:517:1
void xbox_Machine_SetReadWatch(xbox_Machine* m,uint32_t lo,uint32_t hi,std::function<void(uint32_t,uint32_t,uint32_t)> cb){
{
auto tmp133 = std::make_tuple(lo,hi,cb);
m->wRLo = std::get<0>(tmp133);
m->wRHi = std::get<1>(tmp133);
m->onR = std::get<2>(tmp133);
}
}
// tools/platform/xbox/machine.go:521:1
void xbox_Machine_fault(xbox_Machine* m,std::string kind,uint32_t a){
{
if ((bool(m->CPU) && (!m->CPU->Halted))) {
x86_CPU_Halt(m->CPU,std::string("out-of-range %s at %08X (PC %08X)",33),kind,a,x86_CPU_LinearPC(m->CPU));
}
}
}
// tools/platform/xbox/machine.go:529:1
uint32_t xbox_Machine_read32(xbox_Machine* m,uint32_t a){
{
return cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(xbox_Machine_Read(m,a)) | shl<uint32_t>(cast<uint32_t>(xbox_Machine_Read(m,cast<uint32_t>((a + cast<uint32_t>(1ULL))))),cast<int64_t>(8ULL)))) | shl<uint32_t>(cast<uint32_t>(xbox_Machine_Read(m,cast<uint32_t>((a + cast<uint32_t>(2ULL))))),cast<int64_t>(16ULL)))) | shl<uint32_t>(cast<uint32_t>(xbox_Machine_Read(m,cast<uint32_t>((a + cast<uint32_t>(3ULL))))),cast<int64_t>(24ULL))));
}
}
// tools/platform/xbox/machine.go:532:1
uint16_t xbox_Machine_read16(xbox_Machine* m,uint32_t a){
{
return cast<uint16_t>((cast<uint16_t>(xbox_Machine_Read(m,a)) | shl<uint16_t>(cast<uint16_t>(xbox_Machine_Read(m,cast<uint32_t>((a + cast<uint32_t>(1ULL))))),cast<int64_t>(8ULL))));
}
}
// tools/platform/xbox/machine.go:533:1
void xbox_Machine_write32(xbox_Machine* m,uint32_t a,uint32_t v){
{
xbox_Machine_Write(m,a,cast<uint8_t>(v));
xbox_Machine_Write(m,cast<uint32_t>((a + cast<uint32_t>(1ULL))),cast<uint8_t>(shr<uint32_t>(v,cast<int64_t>(8ULL))));
xbox_Machine_Write(m,cast<uint32_t>((a + cast<uint32_t>(2ULL))),cast<uint8_t>(shr<uint32_t>(v,cast<int64_t>(16ULL))));
xbox_Machine_Write(m,cast<uint32_t>((a + cast<uint32_t>(3ULL))),cast<uint8_t>(shr<uint32_t>(v,cast<int64_t>(24ULL))));
}
}
// tools/platform/xbox/machine.go:539:1
void xbox_Machine_write16(xbox_Machine* m,uint32_t a,uint16_t v){
{
xbox_Machine_Write(m,a,cast<uint8_t>(v));
xbox_Machine_Write(m,cast<uint32_t>((a + cast<uint32_t>(1ULL))),cast<uint8_t>(shr<uint16_t>(v,cast<int64_t>(8ULL))));
}
}
// tools/platform/xbox/machine.go:545:1
std::string xbox_Machine_cstr(xbox_Machine* m,uint32_t a){
{
Slice<uint8_t> b={};
{uint32_t i = cast<uint32_t>(0ULL);for (;((i < cast<uint32_t>(1024ULL)) && (a != cast<uint32_t>(0ULL)));i++){
uint8_t ch = xbox_Machine_Read(m,cast<uint32_t>((a + i)));
if ((ch == cast<uint8_t>(0ULL))) {
break;
}
b = append(b,Slice<uint8_t>{ch});
}
}return cast<std::string>(b);
}
}
// tools/platform/xbox/machine.go:566:1
void xbox_Machine_SetVerbose(xbox_Machine* m,bool v){
{
m->verbose = v;
}
}
// tools/platform/xbox/machine.go:572:1
void xbox_Machine_EnableGPU(xbox_Machine* m){
{
m->pusherEnabled = true;
}
}
// tools/platform/xbox/machine.go:575:1
xbox_pgraph* xbox_Machine_PGraph(xbox_Machine* m){
{
return m->pgraph;
}
}
// tools/platform/xbox/machine.go:578:1
void xbox_Machine_SetTrace(xbox_Machine* m,int64_t n){
{
m->traceLeft = n;
}
}
// tools/platform/xbox/machine.go:583:1
void xbox_Machine_SetPCBreak(xbox_Machine* m,uint32_t addr,int64_t n){
{
auto tmp134 = std::make_tuple(addr,n);
m->bpAddr = std::get<0>(tmp134);
m->bpLeft = std::get<1>(tmp134);
}
}
// tools/platform/xbox/machine.go:587:1
uint8_t xbox_Machine_MemReadByte(xbox_Machine* m,uint32_t a){
{
return xbox_Machine_Read(m,a);
}
}
// tools/platform/xbox/machine.go:588:1
uint32_t xbox_Machine_MemRead32(xbox_Machine* m,uint32_t a){
{
return xbox_Machine_read32(m,a);
}
}
// tools/platform/xbox/machine.go:601:1
void xbox_Machine_Poke(xbox_Machine* m,uint32_t a,uint32_t v){
{
xbox_Machine_write32(m,a,v);
}
}
// tools/platform/xbox/machine.go:612:1
void xbox_Machine_ReadRAM(xbox_Machine* m,uint32_t a,Slice<uint8_t> buf){
{
{auto&& tmp135 = buf;
for(int64_t tmp136=0;tmp136<len(tmp135);++tmp136){
auto i=tmp136;buf[i] = cast<uint8_t>(0ULL);
{
auto tmp137 = xbox_Machine_translate(m,cast<uint32_t>((a + cast<uint32_t>(i))));
uint32_t phys = std::get<0>(tmp137);
bool mmio = std::get<1>(tmp137);
bool ok = std::get<2>(tmp137);
if (((ok && (!mmio)) && (cast<int64_t>(phys) < len(m->RAM)))) {
buf[i] = m->RAM[phys];
}
}
}}
}
}
// tools/platform/xbox/machine.go:623:1
void xbox_Machine_ReadCode(xbox_Machine* m,uint32_t a,Slice<uint8_t> buf){
{
xbox_Machine_ReadRAM(m,a,buf);
}
}
// tools/platform/xbox/machine.go:628:1
uint32_t xbox_Machine_NVSubchannelClass(xbox_Machine* m,uint32_t subchan){
{
return m->pgraph->subClass[cast<uint32_t>((subchan & cast<uint32_t>(7ULL)))];
}
}
// tools/platform/xbox/machine.go:633:1
std::tuple<uint32_t,uint32_t,uint32_t> xbox_Machine_AllocStats(xbox_Machine* m){
uint32_t heapNext{};
uint32_t heapTop{};
uint32_t poolNext{};
{
return {m->heapNext,m->heapTop,m->poolNext};
}
}
// tools/platform/xbox/machine.go:671:1
uint32_t xbox_Machine_CallerReturnAddr(xbox_Machine* m){
{
return xbox_Machine_read32(m,m->CPU->Regs[cast<int64_t>(4ULL)]);
}
}
// tools/platform/xbox/machine.go:673:1
uint32_t xbox_align32(uint32_t v,uint32_t a){
{
return ((cast<uint32_t>((cast<uint32_t>((v + a)) - cast<uint32_t>(1ULL)))) & ~((cast<uint32_t>((a - cast<uint32_t>(1ULL))))));
}
}
// tools/platform/xbox/nv2a.go:77:1
std::string xbox_nvRegName(uint32_t off){
{
{
switch((off & ~(cast<uint32_t>(3ULL)))){
case cast<uint32_t>(0ULL):{
return std::string("PMC_BOOT_0",10);
break;}
case cast<uint32_t>(12832ULL):{
return std::string("PFIFO_CACHE1_DMA_PUSH",21);
break;}
case cast<uint32_t>(12864ULL):{
return std::string("PFIFO_CACHE1_DMA_PUT",20);
break;}
case cast<uint32_t>(12868ULL):{
return std::string("PFIFO_CACHE1_DMA_GET",20);
break;}
case cast<uint32_t>(8388672ULL):{
return std::string("USER_DMA_PUT",12);
break;}
case cast<uint32_t>(8388676ULL):{
return std::string("USER_DMA_GET",12);
break;}
}}
{
if ((off < cast<uint32_t>(4096ULL))){
return std::string("PMC",3);
}
else if (((off >= cast<uint32_t>(8192ULL)) && (off < cast<uint32_t>(16384ULL)))){
return std::string("PFIFO",5);
}
else if (((off >= cast<uint32_t>(4194304ULL)) && (off < cast<uint32_t>(4198400ULL)))){
return std::string("PGRAPH",6);
}
else if (((off >= cast<uint32_t>(6291456ULL)) && (off < cast<uint32_t>(6299648ULL)))){
return std::string("PCRTC/PRMCIO",12);
}
else if (((off >= cast<uint32_t>(7340032ULL)) && (off < cast<uint32_t>(8388608ULL)))){
return std::string("PRAMIN",6);
}
else if ((off >= cast<uint32_t>(8388608ULL))){
return std::string("USER",4);
}
}
tmp138:;
return std::string("?",1);
}
}
// tools/platform/xbox/nv2a.go:134:1
uint8_t xbox_Machine_nvRead(xbox_Machine* m,uint32_t off){
{
if ((off >= cast<uint32_t>(16777216ULL))) {
return cast<uint8_t>(255ULL);
}
uint32_t dw = get(m->nv.reg,shr<uint32_t>(off,cast<int64_t>(2ULL)));
{
switch((off & ~(cast<uint32_t>(3ULL)))){
case cast<uint32_t>(0ULL):{
dw = cast<uint32_t>(44040354ULL);
break;}
case cast<uint32_t>(12868ULL):case cast<uint32_t>(8388676ULL):{
dw = m->nv.dmaGet;
break;}
case cast<uint32_t>(12864ULL):case cast<uint32_t>(8388672ULL):{
dw = m->nv.dmaPut;
break;}
case cast<uint32_t>(1049616ULL):{
dw &= ~(cast<uint32_t>(65536ULL));
break;}
case cast<uint32_t>(8448ULL):case cast<uint32_t>(4194560ULL):{
dw = cast<uint32_t>(0ULL);
break;}
case cast<uint32_t>(6291712ULL):{
dw = m->nv.pcrtcIntr;
break;}
case cast<uint32_t>(256ULL):{
dw = cast<uint32_t>(0ULL);
if ((m->nv.pcrtcIntr != cast<uint32_t>(0ULL))) {
dw |= cast<uint32_t>(16777216ULL);
}
break;}
case cast<uint32_t>(12820ULL):case cast<uint32_t>(9216ULL):{
dw |= cast<uint32_t>(16ULL);
break;}
}}
if ((xbox_nvTrace && ((cast<uint32_t>((off & cast<uint32_t>(3ULL)))) == cast<uint32_t>(0ULL)))) {
go_fmt_Printf(std::string("NVrd %06X %-22s -> %08X  PC=%08X\012",33),off,xbox_nvRegName(off),dw,x86_CPU_LinearPC(m->CPU));
}
return cast<uint8_t>(shr<uint32_t>(dw,(cast<uint32_t>((cast<uint32_t>(8ULL) * (cast<uint32_t>((off & cast<uint32_t>(3ULL)))))))));
}
}
// tools/platform/xbox/nv2a.go:175:1
void xbox_Machine_nvWrite(xbox_Machine* m,uint32_t off,uint8_t v){
{
if ((off >= cast<uint32_t>(16777216ULL))) {
return ;
}
uint32_t idx = shr<uint32_t>(off,cast<int64_t>(2ULL));
uint32_t shift = cast<uint32_t>((cast<uint32_t>(8ULL) * (cast<uint32_t>((off & cast<uint32_t>(3ULL))))));
m->nv.reg[idx] = cast<uint32_t>((((get(m->nv.reg,idx) & ~((shl<uint32_t>(cast<uint32_t>(255ULL),shift))))) | shl<uint32_t>(cast<uint32_t>(v),shift)));
if ((cast<uint32_t>((off & cast<uint32_t>(3ULL))) != cast<uint32_t>(3ULL))) {
return ;
}
if (xbox_nvTrace) {
go_fmt_Printf(std::string("NVwr %06X %-22s <- %08X  PC=%08X\012",33),(off & ~(cast<uint32_t>(3ULL))),xbox_nvRegName(off),get(m->nv.reg,idx),x86_CPU_LinearPC(m->CPU));
}
{
switch((off & ~(cast<uint32_t>(3ULL)))){
case cast<uint32_t>(12864ULL):case cast<uint32_t>(8388672ULL):{
m->nv.dmaPut = get(m->nv.reg,idx);
if (((!m->nv.kicked) && (m->nv.dmaPut != m->nv.dmaGet))) {
m->nv.kicked = true;
m->firstPush = true;
xbox_Machine_logf(m,std::string("NV2A: first push-buffer kick \342\200\224 DMA_PUT=%08X (GET=%08X) at PC %08X",67),m->nv.dmaPut,m->nv.dmaGet,x86_CPU_LinearPC(m->CPU));
}
if (m->pusherEnabled) {
xbox_Machine_runPusher(m);
}
break;}
case cast<uint32_t>(12868ULL):case cast<uint32_t>(8388676ULL):{
m->nv.dmaGet = get(m->nv.reg,idx);
break;}
case cast<uint32_t>(6291712ULL):{
m->nv.pcrtcIntr &= ~(get(m->nv.reg,idx));
break;}
}}
}
}
// tools/platform/xbox/nv2a.go:222:1
bool xbox_Machine_FirstPushReached(xbox_Machine* m){
{
return m->firstPush;
}
}
// tools/platform/xbox/nv2a_blit.go:54:1
void xbox_pgraph_surf2DMethod(xbox_pgraph* g,uint32_t method,uint32_t arg){
{
{
switch(method){
case cast<uint32_t>(768ULL):{
g->surf2D.format = arg;
break;}
case cast<uint32_t>(772ULL):{
g->surf2D.srcPitch = cast<uint32_t>((arg & cast<uint32_t>(65535ULL)));
g->surf2D.dstPitch = shr<uint32_t>(arg,cast<int64_t>(16ULL));
break;}
case cast<uint32_t>(776ULL):{
g->surf2D.srcOffset = arg;
break;}
case cast<uint32_t>(780ULL):{
g->surf2D.dstOffset = arg;
break;}
}}
}
}
// tools/platform/xbox/nv2a_blit.go:68:1
void xbox_pgraph_blitMethod(xbox_pgraph* g,uint32_t method,uint32_t arg){
{
{
switch(method){
case cast<uint32_t>(768ULL):{
auto tmp139 = std::make_tuple(cast<uint32_t>((arg & cast<uint32_t>(65535ULL))),shr<uint32_t>(arg,cast<int64_t>(16ULL)));
g->blit.inX = std::get<0>(tmp139);
g->blit.inY = std::get<1>(tmp139);
break;}
case cast<uint32_t>(772ULL):{
auto tmp140 = std::make_tuple(cast<uint32_t>((arg & cast<uint32_t>(65535ULL))),shr<uint32_t>(arg,cast<int64_t>(16ULL)));
g->blit.outX = std::get<0>(tmp140);
g->blit.outY = std::get<1>(tmp140);
break;}
case cast<uint32_t>(776ULL):{
xbox_pgraph_doBlit(g,cast<uint32_t>((arg & cast<uint32_t>(65535ULL))),shr<uint32_t>(arg,cast<int64_t>(16ULL)));
break;}
}}
}
}
// tools/platform/xbox/nv2a_blit.go:83:1
void xbox_pgraph_doBlit(xbox_pgraph* g,uint32_t w,uint32_t h){
{
uint32_t bpp = xbox_surf2DBpp(g->surf2D.format);
if ((((bpp == cast<uint32_t>(0ULL)) || (w == cast<uint32_t>(0ULL))) || (h == cast<uint32_t>(0ULL)))) {
return ;
}
auto tmp141 = g->m;
xbox_Machine* m = tmp141;
auto tmp142 = std::make_tuple((&g->surf2D),(&g->blit));
xbox_surfaces2D* s = std::get<0>(tmp142);
xbox_imageBlit* b = std::get<1>(tmp142);
int64_t n = cast<int64_t>(cast<uint32_t>((w * bpp)));
{uint32_t y = cast<uint32_t>(0ULL);for (;(y < h);y++){
int64_t src = cast<int64_t>(cast<uint32_t>((cast<uint32_t>((s->srcOffset + cast<uint32_t>(((cast<uint32_t>((b->inY + y))) * s->srcPitch)))) + cast<uint32_t>((b->inX * bpp)))));
int64_t dst = cast<int64_t>(cast<uint32_t>((cast<uint32_t>((s->dstOffset + cast<uint32_t>(((cast<uint32_t>((b->outY + y))) * s->dstPitch)))) + cast<uint32_t>((b->outX * bpp)))));
if (((((src < cast<int64_t>(0ULL)) || (dst < cast<int64_t>(0ULL))) || (cast<int64_t>((src + n)) > len(m->RAM))) || (cast<int64_t>((dst + n)) > len(m->RAM)))) {
return ;
}
gcopy(sub(m->RAM,dst,cast<int64_t>((dst + n))),sub(m->RAM,src,cast<int64_t>((src + n))));
}
}}
}
// tools/platform/xbox/nv2a_blit.go:103:1
uint32_t xbox_surf2DBpp(uint32_t format){
{
{
switch(cast<uint32_t>((format & cast<uint32_t>(255ULL)))){
case cast<uint32_t>(1ULL):{
return cast<uint32_t>(1ULL);
break;}
case cast<uint32_t>(3ULL):case cast<uint32_t>(4ULL):case cast<uint32_t>(5ULL):case cast<uint32_t>(6ULL):{
return cast<uint32_t>(2ULL);
break;}
case cast<uint32_t>(7ULL):case cast<uint32_t>(8ULL):case cast<uint32_t>(9ULL):case cast<uint32_t>(10ULL):case cast<uint32_t>(11ULL):{
return cast<uint32_t>(4ULL);
break;}
default:{
return cast<uint32_t>(4ULL);
break;}
}}
}
}
// tools/platform/xbox/nv2a_clip.go:39:1
std::tuple<Slice<xbox_kelvinVtx>,Slice<std::array<int64_t,3>>> xbox_pgraph_clipNearPlane(xbox_pgraph* g,Slice<xbox_kelvinVtx> verts,Slice<std::array<int64_t,3>> tris){
{
bool crosses = false;
{auto&& tmp143 = verts;
for(int64_t tmp144=0;tmp144<len(tmp143);++tmp144){
auto i=tmp144;if ((verts[i].pos[cast<int64_t>(3ULL)] <= cast<float>(1.52587890625000000e-05))) {
crosses = true;
break;
}
}}
if ((!crosses)) {
return {verts,tris};
}
auto tmp145 = std::make_tuple((&g->Const[cast<int64_t>(58ULL)]),(&g->Const[cast<int64_t>(59ULL)]));
std::array<uint32_t,4>* sc = std::get<0>(tmp145);
std::array<uint32_t,4>* off = std::get<1>(tmp145);
auto tmp146 = std::make_tuple(go_math_Float32frombits((*sc)[cast<int64_t>(0ULL)]),go_math_Float32frombits((*sc)[cast<int64_t>(1ULL)]),go_math_Float32frombits((*sc)[cast<int64_t>(2ULL)]));
float sx = std::get<0>(tmp146);
float sy = std::get<1>(tmp146);
float sz = std::get<2>(tmp146);
auto tmp147 = std::make_tuple(go_math_Float32frombits((*off)[cast<int64_t>(0ULL)]),go_math_Float32frombits((*off)[cast<int64_t>(1ULL)]),go_math_Float32frombits((*off)[cast<int64_t>(2ULL)]));
float ox = std::get<0>(tmp147);
float oy = std::get<1>(tmp147);
float oz = std::get<2>(tmp147);
if ((((sx == cast<float>(0.00000000000000000e+00)) || (sy == cast<float>(0.00000000000000000e+00))) || (sz == cast<float>(0.00000000000000000e+00)))) {
return {verts,tris};
}
auto clipOf = [&](xbox_kelvinVtx* v)->std::array<float,4>{
float w = v->pos[cast<int64_t>(3ULL)];
return std::array<float,4>{((((v->pos[cast<int64_t>(0ULL)] - ox)) / sx) * w),((((v->pos[cast<int64_t>(1ULL)] - oy)) / sy) * w),((((v->pos[cast<int64_t>(2ULL)] - oz)) / sz) * w),w};
}
;
auto project = [&](std::array<float,4> c)->std::array<float,4>{
float w = c[cast<int64_t>(3ULL)];
return std::array<float,4>{(((c[cast<int64_t>(0ULL)] / w) * sx) + ox),(((c[cast<int64_t>(1ULL)] / w) * sy) + oy),(((c[cast<int64_t>(2ULL)] / w) * sz) + oz),w};
}
;
g->clipVerts = append(sub(g->clipVerts,0,cast<int64_t>(0ULL)),verts);
Slice<xbox_kelvinVtx> out = g->clipVerts;
Slice<std::array<int64_t,3>> otris = sub(g->clipTris,0,cast<int64_t>(0ULL));
std::array<xbox_kelvinVtx,4> poly={};
{auto&& tmp148 = tris;
for(int64_t tmp149=0;tmp149<len(tmp148);++tmp149){
auto t=tmp148[tmp149];auto tmp150 = std::make_tuple((&verts[t[cast<int64_t>(0ULL)]]),(&verts[t[cast<int64_t>(1ULL)]]),(&verts[t[cast<int64_t>(2ULL)]]));
xbox_kelvinVtx* va = std::get<0>(tmp150);
xbox_kelvinVtx* vb = std::get<1>(tmp150);
xbox_kelvinVtx* vc = std::get<2>(tmp150);
auto tmp151 = std::make_tuple((va->pos[cast<int64_t>(3ULL)] > cast<float>(1.52587890625000000e-05)),(vb->pos[cast<int64_t>(3ULL)] > cast<float>(1.52587890625000000e-05)),(vc->pos[cast<int64_t>(3ULL)] > cast<float>(1.52587890625000000e-05)));
bool ia = std::get<0>(tmp151);
bool ib = std::get<1>(tmp151);
bool ic = std::get<2>(tmp151);
int64_t nin = cast<int64_t>(0ULL);
if (ia) {
nin++;
}
if (ib) {
nin++;
}
if (ic) {
nin++;
}
if ((nin == cast<int64_t>(3ULL))) {
otris = append(otris,Slice<std::array<int64_t,3>>{t});
continue;
}
if ((nin == cast<int64_t>(0ULL))) {
continue;
}
int64_t n = cast<int64_t>(0ULL);
auto edge = [&](xbox_kelvinVtx* p,xbox_kelvinVtx* q,bool pin,bool qin)->void{
if (pin) {
poly[n] = (*p);
n++;
}
if ((pin != qin)) {
auto tmp152 = std::make_tuple(clipOf(p),clipOf(q));
std::array<float,4> cp = std::get<0>(tmp152);
std::array<float,4> cq = std::get<1>(tmp152);
float tt = (((cast<float>(1.52587890625000000e-05) - cp[cast<int64_t>(3ULL)])) / ((cq[cast<int64_t>(3ULL)] - cp[cast<int64_t>(3ULL)])));
poly[n] = xbox_clipVertex(p,q,cp,cq,tt,project);
n++;
}
}
;
edge(va,vb,ia,ib);
edge(vb,vc,ib,ic);
edge(vc,va,ic,ia);
int64_t base = len(out);
out = append(out,sub(poly,0,n));
{int64_t k = cast<int64_t>(1ULL);for (;(cast<int64_t>((k + cast<int64_t>(1ULL))) < n);k++){
otris = append(otris,Slice<std::array<int64_t,3>>{std::array<int64_t,3>{base,cast<int64_t>((base + k)),cast<int64_t>((cast<int64_t>((base + k)) + cast<int64_t>(1ULL)))}});
}
}}}
auto tmp153 = std::make_tuple(out,otris);
g->clipVerts = std::get<0>(tmp153);
g->clipTris = std::get<1>(tmp153);
return {out,otris};
}
}
// tools/platform/xbox/nv2a_clip.go:127:1
xbox_kelvinVtx xbox_clipVertex(xbox_kelvinVtx* p,xbox_kelvinVtx* q,std::array<float,4> cp,std::array<float,4> cq,float t,std::function<std::array<float,4>(std::array<float,4>)> project){
{
std::array<float,4> cn={};
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(4ULL));i++){
cn[i] = (cp[i] + (t * ((cq[i] - cp[i]))));
}
}cn[cast<int64_t>(3ULL)] = cast<float>(1.52587890625000000e-05);
xbox_kelvinVtx v={};
v.pos = project(cn);
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(4ULL));i++){
v.d0[i] = (p->d0[i] + (t * ((q->d0[i] - p->d0[i]))));
v.d1[i] = (p->d1[i] + (t * ((q->d1[i] - p->d1[i]))));
}
}v.fog = (p->fog + (t * ((q->fog - p->fog))));
{int64_t u = cast<int64_t>(0ULL);for (;(u < cast<int64_t>(4ULL));u++){
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(4ULL));i++){
v.uv[u][i] = (p->uv[u][i] + (t * ((q->uv[u][i] - p->uv[u][i]))));
}
}}
}return v;
}
}
// tools/platform/xbox/nv2a_combiner.go:50:1
void xbox_pgraph_combDecode(xbox_pgraph* g,xbox_combState* cs){
{
int64_t n = cast<int64_t>(cast<uint32_t>((g->Regs[cast<int64_t>(1944ULL)] & cast<uint32_t>(255ULL))));
if ((n > cast<int64_t>(8ULL))) {
n = cast<int64_t>(8ULL);
}
uint32_t ctl = g->Regs[cast<int64_t>(1944ULL)];
cs->stages = sub(cs->buf,0,n);
{int64_t i = cast<int64_t>(0ULL);for (;(i < n);i++){
auto tmp154 = std::make_tuple(i,i);
int64_t f0i = std::get<0>(tmp154);
int64_t f1i = std::get<1>(tmp154);
if ((cast<uint32_t>((shr<uint32_t>(ctl,cast<int64_t>(8ULL)) & cast<uint32_t>(1ULL))) == cast<uint32_t>(0ULL))) {
f0i = cast<int64_t>(0ULL);
}
if ((cast<uint32_t>((shr<uint32_t>(ctl,cast<int64_t>(12ULL)) & cast<uint32_t>(1ULL))) == cast<uint32_t>(0ULL))) {
f1i = cast<int64_t>(0ULL);
}
cs->stages[i] = xbox_combStage{g->Regs[cast<uint32_t>((cast<uint32_t>(688ULL) + cast<uint32_t>(i)))],g->Regs[cast<uint32_t>((cast<uint32_t>(152ULL) + cast<uint32_t>(i)))],g->Regs[cast<uint32_t>((cast<uint32_t>(1936ULL) + cast<uint32_t>(i)))],g->Regs[cast<uint32_t>((cast<uint32_t>(680ULL) + cast<uint32_t>(i)))],xbox_argb8Norm(g->Regs[cast<uint32_t>((cast<uint32_t>(664ULL) + cast<uint32_t>(f0i)))]),xbox_argb8Norm(g->Regs[cast<uint32_t>((cast<uint32_t>(672ULL) + cast<uint32_t>(f1i)))])};
}
}cs->cw0 = g->Regs[cast<int64_t>(162ULL)];
cs->cw1 = g->Regs[cast<int64_t>(163ULL)];
std::array<float,4> fc = xbox_argb8Norm(g->Regs[cast<int64_t>(170ULL)]);
cs->fogColor = std::array<float,4>{fc[cast<int64_t>(0ULL)],fc[cast<int64_t>(1ULL)],fc[cast<int64_t>(2ULL)],cast<float>(1.00000000000000000e+00)};
}
}
// tools/platform/xbox/nv2a_combiner.go:82:1
std::array<float,4> xbox_argb8Norm(uint32_t v){
{
std::array<float,4> a = xbox_argb8Vec(v);
return std::array<float,4>{(a[cast<int64_t>(0ULL)] / cast<float>(2.55000000000000000e+02)),(a[cast<int64_t>(1ULL)] / cast<float>(2.55000000000000000e+02)),(a[cast<int64_t>(2ULL)] / cast<float>(2.55000000000000000e+02)),(a[cast<int64_t>(3ULL)] / cast<float>(2.55000000000000000e+02))};
}
}
// tools/platform/xbox/nv2a_combiner.go:87:1
std::array<float,4> xbox_argb8Vec(uint32_t v){
{
return std::array<float,4>{cast<float>(cast<uint32_t>((shr<uint32_t>(v,cast<int64_t>(16ULL)) & cast<uint32_t>(255ULL)))),cast<float>(cast<uint32_t>((shr<uint32_t>(v,cast<int64_t>(8ULL)) & cast<uint32_t>(255ULL)))),cast<float>(cast<uint32_t>((v & cast<uint32_t>(255ULL)))),cast<float>(cast<uint32_t>((shr<uint32_t>(v,cast<int64_t>(24ULL)) & cast<uint32_t>(255ULL))))};
}
}
// tools/platform/xbox/nv2a_combiner.go:103:1
std::array<float,4> xbox_combRegs_read(xbox_combRegs* r,uint32_t reg){
{
return r->r[reg];
}
}
// tools/platform/xbox/nv2a_combiner.go:105:1
void xbox_combRegs_write(xbox_combRegs* r,uint32_t reg,std::array<float,4> v,bool alpha){
{
if ((!(((((reg >= cast<uint32_t>(4ULL)) && (reg <= cast<uint32_t>(5ULL)))) || (((reg >= cast<uint32_t>(8ULL)) && (reg <= cast<uint32_t>(13ULL)))))))) {
return ;
}
if (alpha) {
r->r[reg][cast<int64_t>(3ULL)] = v[cast<int64_t>(3ULL)];
}
else {
auto tmp155 = std::make_tuple(v[cast<int64_t>(0ULL)],v[cast<int64_t>(1ULL)],v[cast<int64_t>(2ULL)]);
r->r[reg][cast<int64_t>(0ULL)] = std::get<0>(tmp155);
r->r[reg][cast<int64_t>(1ULL)] = std::get<1>(tmp155);
r->r[reg][cast<int64_t>(2ULL)] = std::get<2>(tmp155);
}
}
}
// tools/platform/xbox/nv2a_combiner.go:122:1
void xbox_combMap_reference(std::array<float,4>* v,uint32_t mapping){
{
{
switch(mapping){
case cast<uint32_t>(0ULL):{
{auto&& tmp156 = *(v);
for(int64_t tmp157=0;tmp157<len(tmp156);++tmp157){
auto i=tmp157;auto x=tmp156[tmp157];(*v)[i] = xbox_clamp01(x);
}}
break;}
case cast<uint32_t>(1ULL):{
{auto&& tmp158 = *(v);
for(int64_t tmp159=0;tmp159<len(tmp158);++tmp159){
auto i=tmp159;auto x=tmp158[tmp159];(*v)[i] = (cast<float>(1.00000000000000000e+00) - xbox_clamp01(x));
}}
break;}
case cast<uint32_t>(2ULL):{
{auto&& tmp160 = *(v);
for(int64_t tmp161=0;tmp161<len(tmp160);++tmp161){
auto i=tmp161;auto x=tmp160[tmp161];(*v)[i] = ((cast<float>(2.00000000000000000e+00) * xbox_clamp01(x)) - cast<float>(1.00000000000000000e+00));
}}
break;}
case cast<uint32_t>(3ULL):{
{auto&& tmp162 = *(v);
for(int64_t tmp163=0;tmp163<len(tmp162);++tmp163){
auto i=tmp163;auto x=tmp162[tmp163];(*v)[i] = cast<float>(-(((cast<float>(2.00000000000000000e+00) * xbox_clamp01(x)) - cast<float>(1.00000000000000000e+00))));
}}
break;}
case cast<uint32_t>(4ULL):{
{auto&& tmp164 = *(v);
for(int64_t tmp165=0;tmp165<len(tmp164);++tmp165){
auto i=tmp165;auto x=tmp164[tmp165];(*v)[i] = (xbox_clamp01(x) - cast<float>(5.00000000000000000e-01));
}}
break;}
case cast<uint32_t>(5ULL):{
{auto&& tmp166 = *(v);
for(int64_t tmp167=0;tmp167<len(tmp166);++tmp167){
auto i=tmp167;auto x=tmp166[tmp167];(*v)[i] = cast<float>(-((xbox_clamp01(x) - cast<float>(5.00000000000000000e-01))));
}}
break;}
case cast<uint32_t>(6ULL):{
break;}
case cast<uint32_t>(7ULL):{
{auto&& tmp168 = *(v);
for(int64_t tmp169=0;tmp169<len(tmp168);++tmp169){
auto i=tmp169;auto x=tmp168[tmp169];(*v)[i] = cast<float>(-x);
}}
break;}
}}
}
}
// tools/platform/xbox/nv2a_combiner.go:159:1
std::array<float,4> xbox_combFetch(xbox_combRegs* r,uint32_t b,bool alphaSide){
{
std::array<float,4> v = r->r[cast<uint32_t>((b & cast<uint32_t>(15ULL)))];
if ((cast<uint32_t>((shr<uint32_t>(b,cast<int64_t>(4ULL)) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
v = std::array<float,4>{v[cast<int64_t>(3ULL)],v[cast<int64_t>(3ULL)],v[cast<int64_t>(3ULL)],v[cast<int64_t>(3ULL)]};
}
else if (alphaSide) {
v = std::array<float,4>{v[cast<int64_t>(2ULL)],v[cast<int64_t>(2ULL)],v[cast<int64_t>(2ULL)],v[cast<int64_t>(2ULL)]};
}
xbox_combMap((&v),cast<uint32_t>((shr<uint32_t>(b,cast<int64_t>(5ULL)) & cast<uint32_t>(7ULL))));
return v;
}
}
// tools/platform/xbox/nv2a_combiner.go:171:1
std::array<float,4> xbox_pgraph_combine(xbox_pgraph* g,xbox_combState* cs,xbox_combInput* in){
{
xbox_combRegs r={};
r.r[cast<int64_t>(3ULL)] = cs->fogColor;
r.r[cast<int64_t>(4ULL)] = in->col0;
r.r[cast<int64_t>(5ULL)] = in->col1;
auto tmp170 = std::make_tuple(in->tex[cast<int64_t>(0ULL)],in->tex[cast<int64_t>(1ULL)],in->tex[cast<int64_t>(2ULL)],in->tex[cast<int64_t>(3ULL)]);
r.r[cast<int64_t>(8ULL)] = std::get<0>(tmp170);
r.r[cast<int64_t>(9ULL)] = std::get<1>(tmp170);
r.r[cast<int64_t>(10ULL)] = std::get<2>(tmp170);
r.r[cast<int64_t>(11ULL)] = std::get<3>(tmp170);
{auto&& tmp171 = cs->stages;
for(int64_t tmp172=0;tmp172<len(tmp171);++tmp172){
auto si=tmp172;xbox_combStage* st = (&cs->stages[si]);
r.r[cast<int64_t>(1ULL)] = st->factor0;
r.r[cast<int64_t>(2ULL)] = st->factor1;
std::array<float,4> a = xbox_combFetch((&r),cast<uint32_t>((shr<uint32_t>(st->colorICW,cast<int64_t>(24ULL)) & cast<uint32_t>(255ULL))),false);
std::array<float,4> b = xbox_combFetch((&r),cast<uint32_t>((shr<uint32_t>(st->colorICW,cast<int64_t>(16ULL)) & cast<uint32_t>(255ULL))),false);
std::array<float,4> c = xbox_combFetch((&r),cast<uint32_t>((shr<uint32_t>(st->colorICW,cast<int64_t>(8ULL)) & cast<uint32_t>(255ULL))),false);
std::array<float,4> d = xbox_combFetch((&r),cast<uint32_t>((st->colorICW & cast<uint32_t>(255ULL))),false);
std::array<float,4> ab = xbox_mul4(a,b);
std::array<float,4> cd = xbox_mul4(c,d);
if ((cast<uint32_t>((shr<uint32_t>(st->colorOCW,cast<int64_t>(13ULL)) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
float dp = (((a[cast<int64_t>(0ULL)] * b[cast<int64_t>(0ULL)]) + (a[cast<int64_t>(1ULL)] * b[cast<int64_t>(1ULL)])) + (a[cast<int64_t>(2ULL)] * b[cast<int64_t>(2ULL)]));
ab = std::array<float,4>{dp,dp,dp,dp};
}
if ((cast<uint32_t>((shr<uint32_t>(st->colorOCW,cast<int64_t>(12ULL)) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
float dp = (((c[cast<int64_t>(0ULL)] * d[cast<int64_t>(0ULL)]) + (c[cast<int64_t>(1ULL)] * d[cast<int64_t>(1ULL)])) + (c[cast<int64_t>(2ULL)] * d[cast<int64_t>(2ULL)]));
cd = std::array<float,4>{dp,dp,dp,dp};
}
std::array<float,4> sum={};
if ((cast<uint32_t>((shr<uint32_t>(st->colorOCW,cast<int64_t>(14ULL)) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
if ((r.r[cast<int64_t>(12ULL)][cast<int64_t>(3ULL)] < cast<float>(5.00000000000000000e-01))) {
sum = ab;
}
else {
sum = cd;
}
}
else {
sum = xbox_add4(ab,cd);
}
uint32_t op = cast<uint32_t>((shr<uint32_t>(st->colorOCW,cast<int64_t>(15ULL)) & cast<uint32_t>(7ULL)));
xbox_combOp((&ab),op);
xbox_combOp((&cd),op);
xbox_combOp((&sum),op);
xbox_combRegs_write(&(r),cast<uint32_t>((shr<uint32_t>(st->colorOCW,cast<int64_t>(4ULL)) & cast<uint32_t>(15ULL))),ab,false);
xbox_combRegs_write(&(r),cast<uint32_t>((st->colorOCW & cast<uint32_t>(15ULL))),cd,false);
xbox_combRegs_write(&(r),cast<uint32_t>((shr<uint32_t>(st->colorOCW,cast<int64_t>(8ULL)) & cast<uint32_t>(15ULL))),sum,false);
std::array<float,4> aa = xbox_combFetch((&r),cast<uint32_t>((shr<uint32_t>(st->alphaICW,cast<int64_t>(24ULL)) & cast<uint32_t>(255ULL))),true);
std::array<float,4> ba = xbox_combFetch((&r),cast<uint32_t>((shr<uint32_t>(st->alphaICW,cast<int64_t>(16ULL)) & cast<uint32_t>(255ULL))),true);
std::array<float,4> ca = xbox_combFetch((&r),cast<uint32_t>((shr<uint32_t>(st->alphaICW,cast<int64_t>(8ULL)) & cast<uint32_t>(255ULL))),true);
std::array<float,4> da = xbox_combFetch((&r),cast<uint32_t>((st->alphaICW & cast<uint32_t>(255ULL))),true);
std::array<float,4> abA = std::array<float,4>{cast<float>(0.00000000000000000e+00),cast<float>(0.00000000000000000e+00),cast<float>(0.00000000000000000e+00),(aa[cast<int64_t>(3ULL)] * ba[cast<int64_t>(3ULL)])};
std::array<float,4> cdA = std::array<float,4>{cast<float>(0.00000000000000000e+00),cast<float>(0.00000000000000000e+00),cast<float>(0.00000000000000000e+00),(ca[cast<int64_t>(3ULL)] * da[cast<int64_t>(3ULL)])};
std::array<float,4> sumA={};
if ((cast<uint32_t>((shr<uint32_t>(st->alphaOCW,cast<int64_t>(14ULL)) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
if ((r.r[cast<int64_t>(12ULL)][cast<int64_t>(3ULL)] < cast<float>(5.00000000000000000e-01))) {
sumA = abA;
}
else {
sumA = cdA;
}
}
else {
sumA = std::array<float,4>{cast<float>(0.00000000000000000e+00),cast<float>(0.00000000000000000e+00),cast<float>(0.00000000000000000e+00),(abA[cast<int64_t>(3ULL)] + cdA[cast<int64_t>(3ULL)])};
}
uint32_t opA = cast<uint32_t>((shr<uint32_t>(st->alphaOCW,cast<int64_t>(15ULL)) & cast<uint32_t>(7ULL)));
xbox_combOp((&abA),opA);
xbox_combOp((&cdA),opA);
xbox_combOp((&sumA),opA);
xbox_combRegs_write(&(r),cast<uint32_t>((shr<uint32_t>(st->alphaOCW,cast<int64_t>(4ULL)) & cast<uint32_t>(15ULL))),abA,true);
xbox_combRegs_write(&(r),cast<uint32_t>((st->alphaOCW & cast<uint32_t>(15ULL))),cdA,true);
xbox_combRegs_write(&(r),cast<uint32_t>((shr<uint32_t>(st->alphaOCW,cast<int64_t>(8ULL)) & cast<uint32_t>(15ULL))),sumA,true);
}}
std::array<float,4> e = xbox_finalFetch((&r),cast<uint32_t>((shr<uint32_t>(cs->cw1,cast<int64_t>(24ULL)) & cast<uint32_t>(255ULL))),std::array<float,4>{});
std::array<float,4> f = xbox_finalFetch((&r),cast<uint32_t>((shr<uint32_t>(cs->cw1,cast<int64_t>(16ULL)) & cast<uint32_t>(255ULL))),std::array<float,4>{});
std::array<float,4> ef = xbox_mul4(e,f);
std::array<float,4> a = xbox_finalFetch((&r),cast<uint32_t>((shr<uint32_t>(cs->cw0,cast<int64_t>(24ULL)) & cast<uint32_t>(255ULL))),ef);
std::array<float,4> b = xbox_finalFetch((&r),cast<uint32_t>((shr<uint32_t>(cs->cw0,cast<int64_t>(16ULL)) & cast<uint32_t>(255ULL))),ef);
std::array<float,4> c = xbox_finalFetch((&r),cast<uint32_t>((shr<uint32_t>(cs->cw0,cast<int64_t>(8ULL)) & cast<uint32_t>(255ULL))),ef);
std::array<float,4> d = xbox_finalFetch((&r),cast<uint32_t>((cs->cw0 & cast<uint32_t>(255ULL))),ef);
std::array<float,4> gv = xbox_finalFetch((&r),cast<uint32_t>((shr<uint32_t>(cs->cw1,cast<int64_t>(8ULL)) & cast<uint32_t>(255ULL))),ef);
std::array<float,4> out={};
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(3ULL));i++){
out[i] = xbox_clamp01((((a[i] * b[i]) + (((cast<float>(1.00000000000000000e+00) - a[i])) * c[i])) + d[i]));
}
}out[cast<int64_t>(3ULL)] = xbox_clamp01(gv[cast<int64_t>(3ULL)]);
return out;
}
}
// tools/platform/xbox/nv2a_combiner.go:263:1
std::array<float,4> xbox_finalFetch(xbox_combRegs* r,uint32_t b,std::array<float,4> ef){
{
std::array<float,4> v={};
{
switch(cast<uint32_t>((b & cast<uint32_t>(15ULL)))){
case cast<uint32_t>(14ULL):{
{auto&& tmp173 = v;
for(int64_t tmp174=0;tmp174<len(tmp173);++tmp174){
auto i=tmp174;v[i] = xbox_clamp01((r->r[cast<int64_t>(12ULL)][i] + r->r[cast<int64_t>(5ULL)][i]));
}}
break;}
case cast<uint32_t>(15ULL):{
v = ef;
break;}
default:{
v = xbox_combRegs_read(r,cast<uint32_t>((b & cast<uint32_t>(15ULL))));
break;}
}}
if ((cast<uint32_t>((shr<uint32_t>(b,cast<int64_t>(4ULL)) & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL))) {
v = std::array<float,4>{v[cast<int64_t>(3ULL)],v[cast<int64_t>(3ULL)],v[cast<int64_t>(3ULL)],v[cast<int64_t>(3ULL)]};
}
if ((cast<uint32_t>((shr<uint32_t>(b,cast<int64_t>(5ULL)) & cast<uint32_t>(7ULL))) == cast<uint32_t>(1ULL))) {
return std::array<float,4>{(cast<float>(1.00000000000000000e+00) - xbox_clamp01(v[cast<int64_t>(0ULL)])),(cast<float>(1.00000000000000000e+00) - xbox_clamp01(v[cast<int64_t>(1ULL)])),(cast<float>(1.00000000000000000e+00) - xbox_clamp01(v[cast<int64_t>(2ULL)])),(cast<float>(1.00000000000000000e+00) - xbox_clamp01(v[cast<int64_t>(3ULL)]))};
}
return std::array<float,4>{xbox_clamp01(v[cast<int64_t>(0ULL)]),xbox_clamp01(v[cast<int64_t>(1ULL)]),xbox_clamp01(v[cast<int64_t>(2ULL)]),xbox_clamp01(v[cast<int64_t>(3ULL)])};
}
}
// tools/platform/xbox/nv2a_combiner.go:288:1
void xbox_combOp_reference(std::array<float,4>* v,uint32_t op){
{
{
switch(op){
case cast<uint32_t>(1ULL):{
{auto&& tmp175 = *(v);
for(int64_t tmp176=0;tmp176<len(tmp175);++tmp176){
auto i=tmp176;auto x=tmp175[tmp176];(*v)[i] = xbox_clampSigned((x - cast<float>(5.00000000000000000e-01)));
}}
break;}
case cast<uint32_t>(2ULL):{
{auto&& tmp177 = *(v);
for(int64_t tmp178=0;tmp178<len(tmp177);++tmp178){
auto i=tmp178;auto x=tmp177[tmp178];(*v)[i] = xbox_clampSigned((x * cast<float>(2.00000000000000000e+00)));
}}
break;}
case cast<uint32_t>(3ULL):{
{auto&& tmp179 = *(v);
for(int64_t tmp180=0;tmp180<len(tmp179);++tmp180){
auto i=tmp180;auto x=tmp179[tmp180];(*v)[i] = xbox_clampSigned((((x - cast<float>(5.00000000000000000e-01))) * cast<float>(2.00000000000000000e+00)));
}}
break;}
case cast<uint32_t>(4ULL):{
{auto&& tmp181 = *(v);
for(int64_t tmp182=0;tmp182<len(tmp181);++tmp182){
auto i=tmp182;auto x=tmp181[tmp182];(*v)[i] = xbox_clampSigned((x * cast<float>(4.00000000000000000e+00)));
}}
break;}
case cast<uint32_t>(6ULL):{
{auto&& tmp183 = *(v);
for(int64_t tmp184=0;tmp184<len(tmp183);++tmp184){
auto i=tmp184;auto x=tmp183[tmp184];(*v)[i] = xbox_clampSigned((x * cast<float>(5.00000000000000000e-01)));
}}
break;}
default:{
{auto&& tmp185 = *(v);
for(int64_t tmp186=0;tmp186<len(tmp185);++tmp186){
auto i=tmp186;auto x=tmp185[tmp186];(*v)[i] = xbox_clampSigned(x);
}}
break;}
}}
}
}
// tools/platform/xbox/nv2a_combiner.go:317:1
float xbox_clampSigned(float x){
{
if ((x > cast<float>(1.00000000000000000e+00))) {
return cast<float>(1.00000000000000000e+00);
}
if ((x < cast<float>(-1.00000000000000000e+00))) {
return cast<float>(-1.00000000000000000e+00);
}
return x;
}
}
// tools/platform/xbox/nv2a_combiner.go:327:1
std::array<float,4> xbox_mul4(std::array<float,4> a,std::array<float,4> b){
{
return std::array<float,4>{(a[cast<int64_t>(0ULL)] * b[cast<int64_t>(0ULL)]),(a[cast<int64_t>(1ULL)] * b[cast<int64_t>(1ULL)]),(a[cast<int64_t>(2ULL)] * b[cast<int64_t>(2ULL)]),(a[cast<int64_t>(3ULL)] * b[cast<int64_t>(3ULL)])};
}
}
// tools/platform/xbox/nv2a_combiner.go:330:1
std::array<float,4> xbox_add4(std::array<float,4> a,std::array<float,4> b){
{
return std::array<float,4>{(a[cast<int64_t>(0ULL)] + b[cast<int64_t>(0ULL)]),(a[cast<int64_t>(1ULL)] + b[cast<int64_t>(1ULL)]),(a[cast<int64_t>(2ULL)] + b[cast<int64_t>(2ULL)]),(a[cast<int64_t>(3ULL)] + b[cast<int64_t>(3ULL)])};
}
}
// tools/platform/xbox/nv2a_combiner.go:333:1
float xbox_clamp01(float x){
{
if ((x < cast<float>(0.00000000000000000e+00))) {
return cast<float>(0.00000000000000000e+00);
}
if ((x > cast<float>(1.00000000000000000e+00))) {
return cast<float>(1.00000000000000000e+00);
}
return x;
}
}
// tools/platform/xbox/nv2a_debug.go:22:1
std::string xbox_NVMethodName(uint32_t class_,uint32_t method){
{
if ((method == cast<uint32_t>(0ULL))) {
return std::string("SET_OBJECT",10);
}
if ((class_ != cast<uint32_t>(151ULL))) {
return go_fmt_Sprintf(std::string("CLASS_%04X_M%04X",16),class_,method);
}
{
if (((method >= cast<uint32_t>(6912ULL)) && (method < cast<uint32_t>(7168ULL)))){
uint32_t u = divi<uint32_t>((cast<uint32_t>((method - cast<uint32_t>(6912ULL)))),cast<uint32_t>(64ULL));
{
switch(cast<uint32_t>((cast<uint32_t>(6912ULL) + modi<uint32_t>((cast<uint32_t>((method - cast<uint32_t>(6912ULL)))),cast<uint32_t>(64ULL))))){
case cast<uint32_t>(6912ULL):{
return go_fmt_Sprintf(std::string("SET_TEXTURE_OFFSET[%d]",22),u);
break;}
case cast<uint32_t>(6916ULL):{
return go_fmt_Sprintf(std::string("SET_TEXTURE_FORMAT[%d]",22),u);
break;}
case cast<uint32_t>(6920ULL):{
return go_fmt_Sprintf(std::string("SET_TEXTURE_ADDRESS[%d]",23),u);
break;}
case cast<uint32_t>(6924ULL):{
return go_fmt_Sprintf(std::string("SET_TEXTURE_CONTROL0[%d]",24),u);
break;}
case cast<uint32_t>(6928ULL):{
return go_fmt_Sprintf(std::string("SET_TEXTURE_CONTROL1[%d]",24),u);
break;}
case cast<uint32_t>(6932ULL):{
return go_fmt_Sprintf(std::string("SET_TEXTURE_FILTER[%d]",22),u);
break;}
case cast<uint32_t>(6940ULL):{
return go_fmt_Sprintf(std::string("SET_TEXTURE_IMAGE_RECT[%d]",26),u);
break;}
}}
}
else if (((method >= cast<uint32_t>(2816ULL)) && (method < cast<uint32_t>(2944ULL)))){
return std::string("TRANSFORM_PROGRAM",17);
}
else if (((method >= cast<uint32_t>(2944ULL)) && (method < cast<uint32_t>(3072ULL)))){
return std::string("TRANSFORM_CONSTANT",18);
}
else if (((method >= cast<uint32_t>(6464ULL)) && (method < cast<uint32_t>(6528ULL)))){
return go_fmt_Sprintf(std::string("SET_VERTEX_DATA4UB[%d]",22),shr<uint32_t>((cast<uint32_t>((method - cast<uint32_t>(6464ULL)))),cast<int64_t>(2ULL)));
}
else if (((method >= cast<uint32_t>(5920ULL)) && (method < cast<uint32_t>(5984ULL)))){
return go_fmt_Sprintf(std::string("SET_VERTEX_DATA_ARRAY_OFFSET[%d]",32),shr<uint32_t>((cast<uint32_t>((method - cast<uint32_t>(5920ULL)))),cast<int64_t>(2ULL)));
}
else if (((method >= cast<uint32_t>(5984ULL)) && (method < cast<uint32_t>(6048ULL)))){
return go_fmt_Sprintf(std::string("SET_VERTEX_DATA_ARRAY_FORMAT[%d]",32),shr<uint32_t>((cast<uint32_t>((method - cast<uint32_t>(5984ULL)))),cast<int64_t>(2ULL)));
}
}
tmp187:;
{
auto tmp188 = lookup(xbox_kelvinMethodNames,method);
std::string n = std::get<0>(tmp188);
bool ok = std::get<1>(tmp188);
if (ok) {
return n;
}
}
return go_fmt_Sprintf(std::string("KELVIN_M%04X",12),method);
}
}
// tools/platform/xbox/nv2a_debug.go:121:1
std::string xbox_NVMethodDecode(uint32_t class_,uint32_t method,uint32_t arg){
{
if ((class_ != cast<uint32_t>(151ULL))) {
return std::string("",0);
}
{
switch(method){
case cast<uint32_t>(6140ULL):{
if ((arg == cast<uint32_t>(0ULL))) {
return std::string("END \342\200\224 the batch draws here",28);
}
return go_fmt_Sprintf(std::string("BEGIN %s",8),xbox_nvPrimName(arg));
break;}
case cast<uint32_t>(7572ULL):{
std::string s={};
if ((cast<uint32_t>((arg & cast<uint32_t>(240ULL))) != cast<uint32_t>(0ULL))) {
s += std::string("color ",6);
}
if ((cast<uint32_t>((arg & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL))) {
s += std::string("depth ",6);
}
if ((cast<uint32_t>((arg & cast<uint32_t>(2ULL))) != cast<uint32_t>(0ULL))) {
s += std::string("stencil ",8);
}
return (std::string("clear ",6) + s);
break;}
case cast<uint32_t>(512ULL):case cast<uint32_t>(516ULL):case cast<uint32_t>(7576ULL):case cast<uint32_t>(7580ULL):{
return go_fmt_Sprintf(std::string("%d .. %d",8),cast<uint32_t>((arg & cast<uint32_t>(65535ULL))),shr<uint32_t>(arg,cast<int64_t>(16ULL)));
break;}
case cast<uint32_t>(524ULL):{
return go_fmt_Sprintf(std::string("color pitch %d, zeta pitch %d",29),cast<uint32_t>((arg & cast<uint32_t>(65535ULL))),shr<uint32_t>(arg,cast<int64_t>(16ULL)));
break;}
case cast<uint32_t>(528ULL):case cast<uint32_t>(532ULL):{
return go_fmt_Sprintf(std::string("@%08X",5),arg);
break;}
case cast<uint32_t>(7568ULL):{
return go_fmt_Sprintf(std::string("A%02X R%02X G%02X B%02X",23),shr<uint32_t>(arg,cast<int64_t>(24ULL)),cast<uint32_t>((shr<uint32_t>(arg,cast<int64_t>(16ULL)) & cast<uint32_t>(255ULL))),cast<uint32_t>((shr<uint32_t>(arg,cast<int64_t>(8ULL)) & cast<uint32_t>(255ULL))),cast<uint32_t>((arg & cast<uint32_t>(255ULL))));
break;}
case cast<uint32_t>(6160ULL):{
return go_fmt_Sprintf(std::string("start %d, %d vertices",21),cast<uint32_t>((arg & cast<uint32_t>(16777215ULL))),cast<uint32_t>(((shr<uint32_t>(arg,cast<int64_t>(24ULL))) + cast<uint32_t>(1ULL))));
break;}
case cast<uint32_t>(7536ULL):{
return go_fmt_Sprintf(std::string("release %d",10),arg);
break;}
}}
if (((method >= cast<uint32_t>(6912ULL)) && (method < cast<uint32_t>(7168ULL)))) {
uint32_t u = divi<uint32_t>((cast<uint32_t>((method - cast<uint32_t>(6912ULL)))),cast<uint32_t>(64ULL));
{
switch(cast<uint32_t>((cast<uint32_t>(6912ULL) + modi<uint32_t>((cast<uint32_t>((method - cast<uint32_t>(6912ULL)))),cast<uint32_t>(64ULL))))){
case cast<uint32_t>(6912ULL):{
return go_fmt_Sprintf(std::string("unit %d @%08X",13),u,arg);
break;}
case cast<uint32_t>(6916ULL):{
return go_fmt_Sprintf(std::string("unit %d %s, %dx%d",17),u,xbox_nvTexFormatName(cast<uint32_t>((shr<uint32_t>(arg,cast<int64_t>(8ULL)) & cast<uint32_t>(255ULL)))),shl<int64_t>(cast<int64_t>(1ULL),(cast<uint32_t>((shr<uint32_t>(arg,cast<int64_t>(20ULL)) & cast<uint32_t>(15ULL))))),shl<int64_t>(cast<int64_t>(1ULL),(cast<uint32_t>((shr<uint32_t>(arg,cast<int64_t>(24ULL)) & cast<uint32_t>(15ULL))))));
break;}
case cast<uint32_t>(6924ULL):{
return go_fmt_Sprintf(std::string("unit %d enable=%t",17),u,(cast<uint32_t>((shr<uint32_t>(arg,cast<int64_t>(30ULL)) & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL)));
break;}
}}
}
return std::string("",0);
}
}
// tools/platform/xbox/nv2a_debug.go:172:1
std::string xbox_nvPrimName(uint32_t p){
{
{
switch(p){
case cast<uint32_t>(1ULL):{
return std::string("POINTS",6);
break;}
case cast<uint32_t>(2ULL):{
return std::string("LINES",5);
break;}
case cast<uint32_t>(3ULL):{
return std::string("LINE_LOOP",9);
break;}
case cast<uint32_t>(4ULL):{
return std::string("LINE_STRIP",10);
break;}
case cast<uint32_t>(5ULL):{
return std::string("TRIANGLES",9);
break;}
case cast<uint32_t>(6ULL):{
return std::string("TRIANGLE_STRIP",14);
break;}
case cast<uint32_t>(7ULL):{
return std::string("TRIANGLE_FAN",12);
break;}
case cast<uint32_t>(8ULL):{
return std::string("QUADS",5);
break;}
case cast<uint32_t>(9ULL):{
return std::string("QUAD_STRIP",10);
break;}
case cast<uint32_t>(10ULL):{
return std::string("POLYGON",7);
break;}
}}
return go_fmt_Sprintf(std::string("PRIM_%d",7),p);
}
}
// tools/platform/xbox/nv2a_debug.go:199:1
std::string xbox_nvTexFormatName(uint32_t code){
{
{
switch(code){
case cast<uint32_t>(2ULL):{
return std::string("SZ_A1R5G5B5",11);
break;}
case cast<uint32_t>(4ULL):{
return std::string("SZ_A4R4G4B4",11);
break;}
case cast<uint32_t>(5ULL):{
return std::string("SZ_R5G6B5",9);
break;}
case cast<uint32_t>(6ULL):{
return std::string("SZ_A8R8G8B8",11);
break;}
case cast<uint32_t>(7ULL):{
return std::string("SZ_X8R8G8B8",11);
break;}
case cast<uint32_t>(12ULL):{
return std::string("DXT1",4);
break;}
case cast<uint32_t>(14ULL):{
return std::string("DXT3",4);
break;}
case cast<uint32_t>(15ULL):{
return std::string("DXT5",4);
break;}
case cast<uint32_t>(17ULL):{
return std::string("LU_R5G6B5",9);
break;}
case cast<uint32_t>(18ULL):{
return std::string("LU_A8R8G8B8",11);
break;}
case cast<uint32_t>(25ULL):{
return std::string("SZ_A8",5);
break;}
case cast<uint32_t>(30ULL):{
return std::string("LU_X8R8G8B8",11);
break;}
}}
return go_fmt_Sprintf(std::string("FMT_%02X",8),code);
}
}
// tools/platform/xbox/nv2a_debug.go:239:1
std::tuple<int64_t,int64_t> xbox_Machine_SurfaceAAScale(xbox_Machine* m){
int64_t ax{};
int64_t ay{};
{
return xbox_surfaceAAScale(m->pgraph->Regs[cast<int64_t>(130ULL)]);
}
}
// tools/platform/xbox/nv2a_debug.go:255:1
bool xbox_Machine_TextureBound(xbox_Machine* m,int64_t u){
{
if (((u < cast<int64_t>(0ULL)) || (u >= cast<int64_t>(4ULL)))) {
return false;
}
xbox_pgraph* g = m->pgraph;
uint32_t ctl = g->Regs[shr<uint32_t>((cast<uint32_t>((cast<uint32_t>(6924ULL) + cast<uint32_t>((cast<uint32_t>(u) * cast<uint32_t>(64ULL)))))),cast<int64_t>(2ULL))];
uint32_t stage = cast<uint32_t>((shr<uint32_t>(g->Regs[cast<int64_t>(1948ULL)],(cast<uint64_t>((cast<uint64_t>(5ULL) * cast<uint64_t>(u))))) & cast<uint32_t>(31ULL)));
return ((cast<uint32_t>((shr<uint32_t>(ctl,cast<int64_t>(30ULL)) & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL)) && (stage != cast<uint32_t>(0ULL)));
}
}
// tools/platform/xbox/nv2a_debug.go:267:1
std::tuple<image_RGBA*,Error> xbox_Machine_DumpTexture(xbox_Machine* m,int64_t u){
{
if ((!xbox_Machine_TextureBound(m,u))) {
return {{},go_fmt_Errorf(std::string("xbox: texture unit %d is not bound",34),u)};
}
auto tmp189 = xbox_pgraph_DebugDecodeTexture(m->pgraph,u);
int64_t w = std::get<0>(tmp189);
int64_t h = std::get<1>(tmp189);
Slice<uint8_t> pix = std::get<2>(tmp189);
bool ok = std::get<3>(tmp189);
if ((!ok)) {
return {{},go_fmt_Errorf(std::string("xbox: texture unit %d did not decode",36),u)};
}
image_RGBA* img = go_image_NewRGBA(go_image_Rect(cast<int64_t>(0ULL),cast<int64_t>(0ULL),w,h));
gcopy(img->Pix,pix);
return {img,{}};
}
}
// tools/platform/xbox/nv2a_debug.go:286:1
Slice<std::string> xbox_RegionFormats(){
{
return Slice<std::string>{std::string("a8r8g8b8",8),std::string("x8r8g8b8",8),std::string("r5g6b5",6),std::string("a1r5g5b5",8),std::string("a4r4g4b4",8),std::string("a8",2),std::string("z24s8",5)};
}
}
// tools/platform/xbox/nv2a_debug.go:303:1
std::tuple<image_RGBA*,Error> xbox_Machine_RenderRegion(xbox_Machine* m,xbox_RegionSpec s){
{
auto tmp190 = xbox_regionBPP(s.Format);
int64_t bpp = std::get<0>(tmp190);
bool ok = std::get<1>(tmp190);
if ((!ok)) {
return {{},go_fmt_Errorf(std::string("xbox: no region format %q (have %v)",35),s.Format,xbox_RegionFormats())};
}
if (((s.W <= cast<int64_t>(0ULL)) || (s.H <= cast<int64_t>(0ULL)))) {
return {{},go_fmt_Errorf(std::string("xbox: a %dx%d region has nothing to show",40),s.W,s.H)};
}
if ((cast<int64_t>((s.W * s.H)) > cast<int64_t>(4194304ULL))) {
return {{},go_fmt_Errorf(std::string("xbox: a %dx%d region is too big to draw",39),s.W,s.H)};
}
auto tmp191 = xbox_Machine_translate(m,s.Addr);
uint32_t phys = std::get<0>(tmp191);
bool mmio = std::get<1>(tmp191);
ok = std::get<2>(tmp191);
if (((!ok) || mmio)) {
return {{},go_fmt_Errorf(std::string("xbox: %08X is not RAM",21),s.Addr)};
}
int64_t stride = s.Stride;
if ((stride <= cast<int64_t>(0ULL))) {
stride = cast<int64_t>((s.W * bpp));
}
image_RGBA* img = go_image_NewRGBA(go_image_Rect(cast<int64_t>(0ULL),cast<int64_t>(0ULL),s.W,s.H));
{int64_t y = cast<int64_t>(0ULL);for (;(y < s.H);y++){
int64_t row = cast<int64_t>((cast<int64_t>(phys) + cast<int64_t>((y * stride))));
{int64_t x = cast<int64_t>(0ULL);for (;(x < s.W);x++){
int64_t o = cast<int64_t>((row + cast<int64_t>((x * bpp))));
if ((cast<int64_t>((o + bpp)) > len(m->RAM))) {
continue;
}
auto tmp192 = xbox_decodeRegionTexel(s.Format,sub(m->RAM,o,cast<int64_t>((o + bpp))));
uint8_t r = std::get<0>(tmp192);
uint8_t g = std::get<1>(tmp192);
uint8_t b = std::get<2>(tmp192);
uint8_t a = std::get<3>(tmp192);
int64_t i = image_RGBA_PixOffset(img,x,y);
auto tmp193 = std::make_tuple(r,g,b,a);
img->Pix[cast<int64_t>((i + cast<int64_t>(0ULL)))] = std::get<0>(tmp193);
img->Pix[cast<int64_t>((i + cast<int64_t>(1ULL)))] = std::get<1>(tmp193);
img->Pix[cast<int64_t>((i + cast<int64_t>(2ULL)))] = std::get<2>(tmp193);
img->Pix[cast<int64_t>((i + cast<int64_t>(3ULL)))] = std::get<3>(tmp193);
}
}}
}return {img,{}};
}
}
// tools/platform/xbox/nv2a_debug.go:338:1
std::tuple<int64_t,bool> xbox_regionBPP(std::string format){
{
{
auto tmp195=format;
if (tmp195==(std::string("a8r8g8b8",8)) || tmp195==(std::string("x8r8g8b8",8)) || tmp195==(std::string("z24s8",5))){
return {cast<int64_t>(4ULL),true};
}
else if (tmp195==(std::string("r5g6b5",6)) || tmp195==(std::string("a1r5g5b5",8)) || tmp195==(std::string("a4r4g4b4",8))){
return {cast<int64_t>(2ULL),true};
}
else if (tmp195==(std::string("a8",2))){
return {cast<int64_t>(1ULL),true};
}
}
tmp194:;
return {cast<int64_t>(0ULL),false};
}
}
// tools/platform/xbox/nv2a_debug.go:352:1
std::tuple<uint8_t,uint8_t,uint8_t,uint8_t> xbox_decodeRegionTexel(std::string format,Slice<uint8_t> b){
uint8_t r{};
uint8_t g{};
uint8_t bl{};
uint8_t a{};
{
{
auto tmp197=format;
if (tmp197==(std::string("a8r8g8b8",8))){
return {b[cast<int64_t>(2ULL)],b[cast<int64_t>(1ULL)],b[cast<int64_t>(0ULL)],b[cast<int64_t>(3ULL)]};
}
else if (tmp197==(std::string("x8r8g8b8",8))){
return {b[cast<int64_t>(2ULL)],b[cast<int64_t>(1ULL)],b[cast<int64_t>(0ULL)],cast<uint8_t>(255ULL)};
}
else if (tmp197==(std::string("z24s8",5))){
uint32_t d = cast<uint32_t>((cast<uint32_t>((shl<uint32_t>(cast<uint32_t>(b[cast<int64_t>(0ULL)]),cast<int64_t>(8ULL)) | shl<uint32_t>(cast<uint32_t>(b[cast<int64_t>(1ULL)]),cast<int64_t>(16ULL)))) | shl<uint32_t>(cast<uint32_t>(b[cast<int64_t>(2ULL)]),cast<int64_t>(24ULL))));
uint8_t v = cast<uint8_t>(shr<uint32_t>(d,cast<int64_t>(24ULL)));
return {v,v,v,cast<uint8_t>(255ULL)};
}
else if (tmp197==(std::string("r5g6b5",6))){
uint16_t v = cast<uint16_t>((cast<uint16_t>(b[cast<int64_t>(0ULL)]) | shl<uint16_t>(cast<uint16_t>(b[cast<int64_t>(1ULL)]),cast<int64_t>(8ULL))));
return {xbox_exp5(shr<uint16_t>(v,cast<int64_t>(11ULL))),xbox_exp6(cast<uint16_t>((shr<uint16_t>(v,cast<int64_t>(5ULL)) & cast<uint16_t>(63ULL)))),xbox_exp5(cast<uint16_t>((v & cast<uint16_t>(31ULL)))),cast<uint8_t>(255ULL)};
}
else if (tmp197==(std::string("a1r5g5b5",8))){
uint16_t v = cast<uint16_t>((cast<uint16_t>(b[cast<int64_t>(0ULL)]) | shl<uint16_t>(cast<uint16_t>(b[cast<int64_t>(1ULL)]),cast<int64_t>(8ULL))));
a = cast<uint8_t>(0ULL);
if ((cast<uint16_t>((v & cast<uint16_t>(32768ULL))) != cast<uint16_t>(0ULL))) {
a = cast<uint8_t>(255ULL);
}
return {xbox_exp5(cast<uint16_t>((shr<uint16_t>(v,cast<int64_t>(10ULL)) & cast<uint16_t>(31ULL)))),xbox_exp5(cast<uint16_t>((shr<uint16_t>(v,cast<int64_t>(5ULL)) & cast<uint16_t>(31ULL)))),xbox_exp5(cast<uint16_t>((v & cast<uint16_t>(31ULL)))),a};
}
else if (tmp197==(std::string("a4r4g4b4",8))){
uint16_t v = cast<uint16_t>((cast<uint16_t>(b[cast<int64_t>(0ULL)]) | shl<uint16_t>(cast<uint16_t>(b[cast<int64_t>(1ULL)]),cast<int64_t>(8ULL))));
auto ex = [&](uint16_t c)->uint8_t{
return cast<uint8_t>(cast<uint16_t>((shl<uint16_t>(c,cast<int64_t>(4ULL)) | c)));
}
;
return {ex(cast<uint16_t>((shr<uint16_t>(v,cast<int64_t>(8ULL)) & cast<uint16_t>(15ULL)))),ex(cast<uint16_t>((shr<uint16_t>(v,cast<int64_t>(4ULL)) & cast<uint16_t>(15ULL)))),ex(cast<uint16_t>((v & cast<uint16_t>(15ULL)))),ex(cast<uint16_t>((shr<uint16_t>(v,cast<int64_t>(12ULL)) & cast<uint16_t>(15ULL))))};
}
else if (tmp197==(std::string("a8",2))){
return {b[cast<int64_t>(0ULL)],b[cast<int64_t>(0ULL)],b[cast<int64_t>(0ULL)],cast<uint8_t>(255ULL)};
}
}
tmp196:;
return {cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),cast<uint8_t>(255ULL)};
}
}
// tools/platform/xbox/nv2a_debug.go:393:1
std::tuple<Slice<std::string>,bool> xbox_Machine_VshProgramDisasm(xbox_Machine* m){
{
xbox_pgraph* g = m->pgraph;
if ((cast<uint32_t>((g->Regs[cast<int64_t>(1957ULL)] & cast<uint32_t>(3ULL))) != cast<uint32_t>(2ULL))) {
return {{},false};
}
int64_t pc = modi<int64_t>(cast<int64_t>(g->Regs[cast<int64_t>(1960ULL)]),cast<int64_t>(136ULL));
Slice<std::string> out={};
{int64_t steps = cast<int64_t>(0ULL);for (;(steps < cast<int64_t>(136ULL));steps++){
xbox_vshInst inst = xbox_pgraph_vshDecode(g,pc);
out = append(out,Slice<std::string>{go_fmt_Sprintf(std::string("[%3d] %s",8),pc,xbox_pgraph_vshDisasm(g,pc))});
if (inst.final) {
return {out,true};
}
pc = modi<int64_t>((cast<int64_t>((pc + cast<int64_t>(1ULL)))),cast<int64_t>(136ULL));
}
}return {out,false};
}
}
// tools/platform/xbox/nv2a_debug.go:412:1
std::array<float,4> xbox_Machine_VshConst(xbox_Machine* m,int64_t c){
{
if (((c < cast<int64_t>(0ULL)) || (c >= cast<int64_t>(192ULL)))) {
return std::array<float,4>{};
}
return xbox_f32vec((&m->pgraph->Const[c]));
}
}
// tools/platform/xbox/nv2a_frame.go:39:1
void xbox_pgraph_clearSurface(xbox_pgraph* g,uint32_t mask){
{rrconsole::EventScope restore(rrcapture::trace.current); if(rrcapture::trace.active)rrXboxDraw(g,"xbox_pgraph_clearSurface");
{rrprof::Scope timing(4,"Surface clears");
{
auto tmp198 = g->m;
xbox_Machine* m = tmp198;
if (xbox_shadowTrace) {
go_fmt_Printf(std::string("SHADOW CLEAR mask=%02X color=%08X zeta=%08X zclear=%08X draws=%d\012",65),mask,g->Regs[cast<int64_t>(132ULL)],g->Regs[cast<int64_t>(133ULL)],g->Regs[cast<int64_t>(1891ULL)],g->Draws);
}
uint32_t format = g->Regs[cast<int64_t>(130ULL)];
auto tmp199 = xbox_surfaceAAScale(format);
int64_t ax = std::get<0>(tmp199);
int64_t ay = std::get<1>(tmp199);
uint32_t rh = g->Regs[cast<int64_t>(1894ULL)];
uint32_t rv = g->Regs[cast<int64_t>(1895ULL)];
auto tmp200 = std::make_tuple(cast<uint32_t>((cast<uint32_t>((rh & cast<uint32_t>(65535ULL))) * cast<uint32_t>(ax))),cast<uint32_t>((cast<uint32_t>(((cast<uint32_t>((shr<uint32_t>(rh,cast<int64_t>(16ULL)) + cast<uint32_t>(1ULL)))) * cast<uint32_t>(ax))) - cast<uint32_t>(1ULL))));
uint32_t x1 = std::get<0>(tmp200);
uint32_t x2 = std::get<1>(tmp200);
auto tmp201 = std::make_tuple(cast<uint32_t>((cast<uint32_t>((rv & cast<uint32_t>(65535ULL))) * cast<uint32_t>(ay))),cast<uint32_t>((cast<uint32_t>(((cast<uint32_t>((shr<uint32_t>(rv,cast<int64_t>(16ULL)) + cast<uint32_t>(1ULL)))) * cast<uint32_t>(ay))) - cast<uint32_t>(1ULL))));
uint32_t y1 = std::get<0>(tmp201);
uint32_t y2 = std::get<1>(tmp201);
bool swizzle = (cast<uint32_t>((shr<uint32_t>(format,cast<int64_t>(8ULL)) & cast<uint32_t>(15ULL))) == cast<uint32_t>(2ULL));
auto tmp202 = std::make_tuple(cast<int64_t>(0ULL),cast<int64_t>(0ULL),std::string("",0));
int64_t swW = std::get<0>(tmp202);
int64_t swH = std::get<1>(tmp202);
std::string reason = std::get<2>(tmp202);
if (swizzle) {
auto tmp203 = xbox_swizzleGeom(format);
swW = std::get<0>(tmp203);
swH = std::get<1>(tmp203);
reason = std::get<2>(tmp203);
if ((reason != std::string("",0))) {
x86_CPU_Halt(m->CPU,std::string("nv2a: clear of swizzled surface %dx%d (0x208=%08X) \342\200\224 %s",57),swW,swH,format,reason);
return ;
}
}
auto addr = [&](uint32_t base,uint32_t pitch,uint32_t x,uint32_t y)->uint32_t{
if (swizzle) {
return cast<uint32_t>((base + cast<uint32_t>((cast<uint32_t>(xbox_swizzleOffset(cast<int64_t>(x),cast<int64_t>(y),swW,swH)) * cast<uint32_t>(4ULL)))));
}
return cast<uint32_t>((cast<uint32_t>((base + cast<uint32_t>((y * pitch)))) + cast<uint32_t>((x * cast<uint32_t>(4ULL)))));
}
;
if ((cast<uint32_t>((mask & cast<uint32_t>(240ULL))) != cast<uint32_t>(0ULL))) {
uint32_t base = g->Regs[cast<int64_t>(132ULL)];
uint32_t pitch = cast<uint32_t>((g->Regs[cast<int64_t>(131ULL)] & cast<uint32_t>(65535ULL)));
uint32_t color = g->Regs[cast<int64_t>(1892ULL)];
{
auto tmp204 = xbox_Machine_translate(m,base);
uint32_t phys = std::get<0>(tmp204);
bool mmio = std::get<1>(tmp204);
bool ok = std::get<2>(tmp204);
if ((ok && (!mmio))) {
{uint32_t y = y1;for (;(y <= y2);y++){
{uint32_t x = x1;for (;(x <= x2);x++){
uint32_t row = addr(phys,pitch,x,y);
if ((cast<uint32_t>((row + cast<uint32_t>(4ULL))) > cast<uint32_t>(len(m->RAM)))) {
break;
}
m->RAM[cast<uint32_t>((row + cast<uint32_t>(0ULL)))] = cast<uint8_t>(color);
if(rrcapture::trace.active)rrXboxWrite(g->m,m->RAM,cast<uint32_t>((row + cast<uint32_t>(0ULL))));
m->RAM[cast<uint32_t>((row + cast<uint32_t>(1ULL)))] = cast<uint8_t>(shr<uint32_t>(color,cast<int64_t>(8ULL)));
if(rrcapture::trace.active)rrXboxWrite(g->m,m->RAM,cast<uint32_t>((row + cast<uint32_t>(1ULL))));
m->RAM[cast<uint32_t>((row + cast<uint32_t>(2ULL)))] = cast<uint8_t>(shr<uint32_t>(color,cast<int64_t>(16ULL)));
if(rrcapture::trace.active)rrXboxWrite(g->m,m->RAM,cast<uint32_t>((row + cast<uint32_t>(2ULL))));
m->RAM[cast<uint32_t>((row + cast<uint32_t>(3ULL)))] = cast<uint8_t>(shr<uint32_t>(color,cast<int64_t>(24ULL)));
if(rrcapture::trace.active)rrXboxWrite(g->m,m->RAM,cast<uint32_t>((row + cast<uint32_t>(3ULL))));
if (bool(m->OnPixel)) {
m->OnPixel(x,y,xbox_PixelEvent{true,{},{},cast<uint8_t>(shr<uint32_t>(color,cast<int64_t>(16ULL))),cast<uint8_t>(shr<uint32_t>(color,cast<int64_t>(8ULL))),cast<uint8_t>(color),cast<uint8_t>(shr<uint32_t>(color,cast<int64_t>(24ULL)))});
}
}
}}
}}
}
}
if ((cast<uint32_t>((mask & cast<uint32_t>(3ULL))) != cast<uint32_t>(0ULL))) {
uint32_t base = g->Regs[cast<int64_t>(133ULL)];
uint32_t pitch = shr<uint32_t>(g->Regs[cast<int64_t>(131ULL)],cast<int64_t>(16ULL));
uint32_t clear = g->Regs[cast<int64_t>(1891ULL)];
uint32_t keep = cast<uint32_t>(0ULL);
if ((cast<uint32_t>((mask & cast<uint32_t>(1ULL))) == cast<uint32_t>(0ULL))) {
keep |= cast<uint32_t>(4294967040ULL);
}
if ((cast<uint32_t>((mask & cast<uint32_t>(2ULL))) == cast<uint32_t>(0ULL))) {
keep |= cast<uint32_t>(255ULL);
}
{
auto tmp205 = xbox_Machine_translate(m,base);
uint32_t phys = std::get<0>(tmp205);
bool mmio = std::get<1>(tmp205);
bool ok = std::get<2>(tmp205);
if (((ok && (!mmio)) && (((pitch != cast<uint32_t>(0ULL)) || swizzle)))) {
{uint32_t y = y1;for (;(y <= y2);y++){
{uint32_t x = x1;for (;(x <= x2);x++){
uint32_t row = addr(phys,pitch,x,y);
if ((cast<uint32_t>((row + cast<uint32_t>(4ULL))) > cast<uint32_t>(len(m->RAM)))) {
break;
}
uint32_t old = cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(m->RAM[row]) | shl<uint32_t>(cast<uint32_t>(m->RAM[cast<uint32_t>((row + cast<uint32_t>(1ULL)))]),cast<int64_t>(8ULL)))) | shl<uint32_t>(cast<uint32_t>(m->RAM[cast<uint32_t>((row + cast<uint32_t>(2ULL)))]),cast<int64_t>(16ULL)))) | shl<uint32_t>(cast<uint32_t>(m->RAM[cast<uint32_t>((row + cast<uint32_t>(3ULL)))]),cast<int64_t>(24ULL))));
uint32_t v = cast<uint32_t>(((clear & ~(keep)) | cast<uint32_t>((old & keep))));
m->RAM[cast<uint32_t>((row + cast<uint32_t>(0ULL)))] = cast<uint8_t>(v);
if(rrcapture::trace.active)rrXboxWrite(g->m,m->RAM,cast<uint32_t>((row + cast<uint32_t>(0ULL))));
m->RAM[cast<uint32_t>((row + cast<uint32_t>(1ULL)))] = cast<uint8_t>(shr<uint32_t>(v,cast<int64_t>(8ULL)));
if(rrcapture::trace.active)rrXboxWrite(g->m,m->RAM,cast<uint32_t>((row + cast<uint32_t>(1ULL))));
m->RAM[cast<uint32_t>((row + cast<uint32_t>(2ULL)))] = cast<uint8_t>(shr<uint32_t>(v,cast<int64_t>(16ULL)));
if(rrcapture::trace.active)rrXboxWrite(g->m,m->RAM,cast<uint32_t>((row + cast<uint32_t>(2ULL))));
m->RAM[cast<uint32_t>((row + cast<uint32_t>(3ULL)))] = cast<uint8_t>(shr<uint32_t>(v,cast<int64_t>(24ULL)));
if(rrcapture::trace.active)rrXboxWrite(g->m,m->RAM,cast<uint32_t>((row + cast<uint32_t>(3ULL))));
}
}}
}}
}
}
}
}
}
}
// tools/platform/xbox/nv2a_frame.go:155:1
void xbox_pgraph_recordPresented(xbox_pgraph* g){
{
int64_t w = cast<int64_t>(shr<uint32_t>(g->Regs[cast<int64_t>(128ULL)],cast<int64_t>(16ULL)));
int64_t h = cast<int64_t>(shr<uint32_t>(g->Regs[cast<int64_t>(129ULL)],cast<int64_t>(16ULL)));
uint32_t pitch = cast<uint32_t>((g->Regs[cast<int64_t>(131ULL)] & cast<uint32_t>(65535ULL)));
uint32_t base = g->Regs[cast<int64_t>(132ULL)];
if (((((w <= cast<int64_t>(0ULL)) || (h <= cast<int64_t>(0ULL))) || (pitch == cast<uint32_t>(0ULL))) || (base == cast<uint32_t>(0ULL)))) {
return ;
}
auto tmp206 = xbox_surfaceAAScale(g->Regs[cast<int64_t>(130ULL)]);
int64_t ax = std::get<0>(tmp206);
int64_t ay = std::get<1>(tmp206);
g->presented = xbox_presentedSurface{true,base,pitch,w,h,ax,ay};
}
}
// tools/platform/xbox/nv2a_frame.go:187:1
std::tuple<image_RGBA*,Error> xbox_Machine_RenderPresented(xbox_Machine* m){
{
xbox_presentedSurface p = m->pgraph->presented;
if ((!p.valid)) {
return xbox_Machine_RenderDrawTarget(m);
}
return xbox_Machine_renderRawSurface(m,p.base,p.pitch,p.w,p.h,p.ax,p.ay);
}
}
// tools/platform/xbox/nv2a_frame.go:204:1
std::tuple<image_RGBA*,Error> xbox_Machine_RenderScanout(xbox_Machine* m){
{
if ((m->nv.fbAddr == cast<uint32_t>(0ULL))) {
return xbox_Machine_RenderDrawTarget(m);
}
uint32_t pitch = m->nv.fbPitch;
int64_t w = cast<int64_t>(divi<uint32_t>(pitch,cast<uint32_t>(4ULL)));
int64_t h = divi<int64_t>(cast<int64_t>((w * cast<int64_t>(3ULL))),cast<int64_t>(4ULL));
if (((w <= cast<int64_t>(0ULL)) || (h <= cast<int64_t>(0ULL)))) {
return {{},go_fmt_Errorf(std::string("nv2a: no display mode programmed (pitch=%d)",43),pitch)};
}
return xbox_Machine_renderRawSurface(m,m->nv.fbAddr,pitch,w,h,cast<int64_t>(1ULL),cast<int64_t>(1ULL));
}
}
// tools/platform/xbox/nv2a_frame.go:222:1
std::tuple<image_RGBA*,Error> xbox_Machine_RenderDrawTarget(xbox_Machine* m){
{
xbox_pgraph* g = m->pgraph;
uint32_t format = g->Regs[cast<int64_t>(130ULL)];
int64_t w = cast<int64_t>(shr<uint32_t>(g->Regs[cast<int64_t>(128ULL)],cast<int64_t>(16ULL)));
int64_t h = cast<int64_t>(shr<uint32_t>(g->Regs[cast<int64_t>(129ULL)],cast<int64_t>(16ULL)));
uint32_t pitch = cast<uint32_t>((g->Regs[cast<int64_t>(131ULL)] & cast<uint32_t>(65535ULL)));
uint32_t base = g->Regs[cast<int64_t>(132ULL)];
auto tmp207 = xbox_surfaceAAScale(format);
int64_t ax = std::get<0>(tmp207);
int64_t ay = std::get<1>(tmp207);
if ((((w <= cast<int64_t>(0ULL)) || (h <= cast<int64_t>(0ULL))) || (base == cast<uint32_t>(0ULL)))) {
return {{},go_fmt_Errorf(std::string("nv2a: no render surface programmed (w=%d h=%d pitch=%d base=%08X)",65),w,h,pitch,base)};
}
if ((cast<uint32_t>((shr<uint32_t>(format,cast<int64_t>(8ULL)) & cast<uint32_t>(15ULL))) == cast<uint32_t>(2ULL))) {
auto tmp208 = xbox_swizzleGeom(format);
int64_t sw = std::get<0>(tmp208);
int64_t sh = std::get<1>(tmp208);
std::string reason = std::get<2>(tmp208);
if ((reason != std::string("",0))) {
return {{},go_fmt_Errorf(std::string("nv2a: swizzled surface %dx%d not renderable (0x208=%08X): %s",60),sw,sh,format,reason)};
}
return xbox_Machine_renderSwizzledSurface(m,base,sw,sh,w,h);
}
if ((pitch == cast<uint32_t>(0ULL))) {
return {{},go_fmt_Errorf(std::string("nv2a: no render surface programmed (w=%d h=%d pitch=%d base=%08X)",65),w,h,pitch,base)};
}
return xbox_Machine_renderRawSurface(m,base,pitch,w,h,ax,ay);
}
}
// tools/platform/xbox/nv2a_frame.go:271:1
std::tuple<image_RGBA*,Error> xbox_Machine_renderSwizzledSurface(xbox_Machine* m,uint32_t base,int64_t sw,int64_t sh,int64_t w,int64_t h){
{
auto tmp209 = xbox_Machine_translate(m,base);
uint32_t phys = std::get<0>(tmp209);
bool mmio = std::get<1>(tmp209);
bool ok = std::get<2>(tmp209);
if (((!ok) || mmio)) {
return {{},go_fmt_Errorf(std::string("nv2a: swizzled surface at %08X is not RAM",41),base)};
}
image_RGBA* img = go_image_NewRGBA(go_image_Rect(cast<int64_t>(0ULL),cast<int64_t>(0ULL),w,h));
{int64_t y = cast<int64_t>(0ULL);for (;(y < h);y++){
{int64_t x = cast<int64_t>(0ULL);for (;(x < w);x++){
uint32_t o = cast<uint32_t>((phys + cast<uint32_t>((cast<uint32_t>(xbox_swizzleOffset(x,y,sw,sh)) * cast<uint32_t>(4ULL)))));
int64_t i = image_RGBA_PixOffset(img,x,y);
if ((cast<uint32_t>((o + cast<uint32_t>(4ULL))) > cast<uint32_t>(len(m->RAM)))) {
continue;
}
img->Pix[cast<int64_t>((i + cast<int64_t>(0ULL)))] = m->RAM[cast<uint32_t>((o + cast<uint32_t>(2ULL)))];
img->Pix[cast<int64_t>((i + cast<int64_t>(1ULL)))] = m->RAM[cast<uint32_t>((o + cast<uint32_t>(1ULL)))];
img->Pix[cast<int64_t>((i + cast<int64_t>(2ULL)))] = m->RAM[cast<uint32_t>((o + cast<uint32_t>(0ULL)))];
img->Pix[cast<int64_t>((i + cast<int64_t>(3ULL)))] = cast<uint8_t>(255ULL);
}
}}
}return {img,{}};
}
}
// tools/platform/xbox/nv2a_frame.go:296:1
std::tuple<image_RGBA*,Error> xbox_Machine_renderRawSurface(xbox_Machine* m,uint32_t base,uint32_t pitch,int64_t w,int64_t h,int64_t ax,int64_t ay){
{
auto tmp210 = xbox_Machine_translate(m,base);
uint32_t phys = std::get<0>(tmp210);
bool mmio = std::get<1>(tmp210);
bool ok = std::get<2>(tmp210);
if (((!ok) || mmio)) {
return {{},go_fmt_Errorf(std::string("nv2a: surface at %08X is not RAM",32),base)};
}
image_RGBA* img = go_image_NewRGBA(go_image_Rect(cast<int64_t>(0ULL),cast<int64_t>(0ULL),w,h));
uint32_t n = cast<uint32_t>(cast<int64_t>((ax * ay)));
{int64_t y = cast<int64_t>(0ULL);for (;(y < h);y++){
{int64_t x = cast<int64_t>(0ULL);for (;(x < w);x++){
uint32_t r={};
uint32_t gg={};
uint32_t b={};
{int64_t sy = cast<int64_t>(0ULL);for (;(sy < ay);sy++){
uint32_t row = cast<uint32_t>((phys + cast<uint32_t>((cast<uint32_t>(cast<int64_t>((cast<int64_t>((y * ay)) + sy))) * pitch))));
{int64_t sx = cast<int64_t>(0ULL);for (;(sx < ax);sx++){
uint32_t o = cast<uint32_t>((row + cast<uint32_t>((cast<uint32_t>(cast<int64_t>((cast<int64_t>((x * ax)) + sx))) * cast<uint32_t>(4ULL)))));
if ((cast<uint32_t>((o + cast<uint32_t>(4ULL))) > cast<uint32_t>(len(m->RAM)))) {
continue;
}
b += cast<uint32_t>(m->RAM[o]);
gg += cast<uint32_t>(m->RAM[cast<uint32_t>((o + cast<uint32_t>(1ULL)))]);
r += cast<uint32_t>(m->RAM[cast<uint32_t>((o + cast<uint32_t>(2ULL)))]);
}
}}
}int64_t i = image_RGBA_PixOffset(img,x,y);
img->Pix[cast<int64_t>((i + cast<int64_t>(0ULL)))] = cast<uint8_t>(divi<uint32_t>(r,n));
img->Pix[cast<int64_t>((i + cast<int64_t>(1ULL)))] = cast<uint8_t>(divi<uint32_t>(gg,n));
img->Pix[cast<int64_t>((i + cast<int64_t>(2ULL)))] = cast<uint8_t>(divi<uint32_t>(b,n));
img->Pix[cast<int64_t>((i + cast<int64_t>(3ULL)))] = cast<uint8_t>(255ULL);
}
}}
}return {img,{}};
}
}
// tools/platform/xbox/nv2a_kelvin.go:98:1
void xbox_pgraph_kelvinMethod(xbox_pgraph* g,uint32_t method,uint32_t arg){
{
if ((method == cast<uint32_t>(304ULL))) {
if (g->m->FlipVSync) {
xbox_Machine_creditFlipVBlank(g->m);
}
xbox_pgraph_recordPresented(g);
xbox_Machine_profFrame(g->m);
if (bool(g->m->OnFlip)) {
g->m->OnFlip(g->m);
}
}
if ((method < cast<uint32_t>(8192ULL))) {
g->Regs[shr<uint32_t>(method,cast<int64_t>(2ULL))] = arg;
}
if ((xbox_nvVPTrace && ((((method >= cast<uint32_t>(2592ULL)) && (method < cast<uint32_t>(2608ULL))) || ((method >= cast<uint32_t>(2800ULL)) && (method < cast<uint32_t>(2816ULL))))))) {
go_fmt_Printf(std::string("VP method %04X = %08X (%g) draws=%d\012",36),method,arg,go_math_Float32frombits(arg),g->Draws);
}
if (((xbox_nvSurfTrace && (method >= cast<uint32_t>(512ULL))) && (method <= cast<uint32_t>(532ULL)))) {
go_fmt_Printf(std::string("SURF method %04X = %08X draws=%d\012",33),method,arg,g->Draws);
}
if (xbox_shadowTrace) {
{
switch(method){
case cast<uint32_t>(304ULL):{
go_fmt_Printf(std::string("SHADOW FLIP draws=%d tid=%d tick=%d\012",36),g->Draws,xbox_Machine_threadID(g->m),g->m->tick);
break;}
case cast<uint32_t>(528ULL):{
go_fmt_Printf(std::string("SHADOW COLOR=%08X draws=%d\012",27),arg,g->Draws);
break;}
case cast<uint32_t>(532ULL):{
go_fmt_Printf(std::string("SHADOW ZETA=%08X draws=%d\012",26),arg,g->Draws);
break;}
case cast<uint32_t>(6912ULL):case cast<uint32_t>(6976ULL):case cast<uint32_t>(7040ULL):case cast<uint32_t>(7104ULL):{
uint32_t u = divi<uint32_t>((cast<uint32_t>((method - cast<uint32_t>(6912ULL)))),cast<uint32_t>(64ULL));
go_fmt_Printf(std::string("SHADOW TEXOFF unit=%d off=%08X draws=%d\012",40),u,arg,g->Draws);
break;}
}}
}
if (xbox_nvSemTrace) {
{
switch(method){
case cast<uint32_t>(256ULL):case cast<uint32_t>(272ULL):case cast<uint32_t>(304ULL):case cast<uint32_t>(7532ULL):case cast<uint32_t>(7536ULL):case cast<uint32_t>(6096ULL):{
go_fmt_Printf(std::string("KELVIN sync: method %04X arg %08X (steps=%d)\012",45),method,arg,g->m->CPU->Steps);
break;}
}}
}
{
if ((method == cast<uint32_t>(7572ULL))){
time_Time tc = xbox_Machine_profStart(g->m);
xbox_pgraph_clearSurface(g,arg);
xbox_Machine_profEnd(g->m,cast<int64_t>(2ULL),tc);
return ;
}
else if ((method == cast<uint32_t>(7536ULL))){
auto tmp212 = xbox_Machine_dmaObjectTarget(g->m,g->Regs[cast<int64_t>(105ULL)]);
uint32_t base = std::get<0>(tmp212);
uint32_t limit = std::get<1>(tmp212);
uint32_t off = g->Regs[cast<int64_t>(1883ULL)];
if (((base != cast<uint32_t>(0ULL)) && (off <= limit))) {
xbox_Machine_write32(g->m,cast<uint32_t>((base + off)),arg);
}
g->m->nv.reg[cast<uint32_t>(1049284ULL)] = shl<uint32_t>(arg,cast<int64_t>(2ULL));
return ;
}
else if ((method == cast<uint32_t>(6140ULL))){
g->rastValid = false;
xbox_pgraph_beginEnd(g,arg);
return ;
}
else if ((method == cast<uint32_t>(6168ULL))){
g->inline_ = append(g->inline_,Slice<uint32_t>{arg});
return ;
}
else if ((method == cast<uint32_t>(6144ULL))){
g->elems = append(g->elems,Slice<uint32_t>{cast<uint32_t>((arg & cast<uint32_t>(65535ULL))),shr<uint32_t>(arg,cast<int64_t>(16ULL))});
return ;
}
else if ((method == cast<uint32_t>(6152ULL))){
g->elems = append(g->elems,Slice<uint32_t>{arg});
return ;
}
else if ((method == cast<uint32_t>(6160ULL))){
g->ranges = append(g->ranges,Slice<std::array<uint32_t,2>>{std::array<uint32_t,2>{cast<uint32_t>((arg & cast<uint32_t>(16777215ULL))),cast<uint32_t>(((shr<uint32_t>(arg,cast<int64_t>(24ULL))) + cast<uint32_t>(1ULL)))}});
return ;
}
else if (((method >= cast<uint32_t>(6464ULL)) && (method < cast<uint32_t>(6528ULL)))){
uint32_t i = shr<uint32_t>((cast<uint32_t>((method - cast<uint32_t>(6464ULL)))),cast<int64_t>(2ULL));
g->vtxAttr[i] = std::array<float,4>{(cast<float>(cast<uint32_t>((arg & cast<uint32_t>(255ULL)))) / cast<float>(2.55000000000000000e+02)),(cast<float>(cast<uint32_t>((shr<uint32_t>(arg,cast<int64_t>(8ULL)) & cast<uint32_t>(255ULL)))) / cast<float>(2.55000000000000000e+02)),(cast<float>(cast<uint32_t>((shr<uint32_t>(arg,cast<int64_t>(16ULL)) & cast<uint32_t>(255ULL)))) / cast<float>(2.55000000000000000e+02)),(cast<float>(cast<uint32_t>((shr<uint32_t>(arg,cast<int64_t>(24ULL)) & cast<uint32_t>(255ULL)))) / cast<float>(2.55000000000000000e+02))};
return ;
}
else if (((method >= cast<uint32_t>(2816ULL)) && (method < cast<uint32_t>(2944ULL)))){
xbox_pgraph_progData(g,arg);
return ;
}
else if (((method >= cast<uint32_t>(2944ULL)) && (method < cast<uint32_t>(3072ULL)))){
xbox_pgraph_constData(g,arg);
return ;
}
else if ((method == cast<uint32_t>(7836ULL))){
g->ProgLoad = arg;
g->progBufN = cast<int64_t>(0ULL);
return ;
}
else if ((method == cast<uint32_t>(7844ULL))){
g->ConstLoad = arg;
g->constBufN = cast<int64_t>(0ULL);
return ;
}
else if (((method >= cast<uint32_t>(2592ULL)) && (method < cast<uint32_t>(2608ULL)))){
g->Const[cast<int64_t>(59ULL)][shr<uint32_t>((cast<uint32_t>((method - cast<uint32_t>(2592ULL)))),cast<int64_t>(2ULL))] = arg;
return ;
}
else if (((method >= cast<uint32_t>(2800ULL)) && (method < cast<uint32_t>(2816ULL)))){
g->Const[cast<int64_t>(58ULL)][shr<uint32_t>((cast<uint32_t>((method - cast<uint32_t>(2800ULL)))),cast<int64_t>(2ULL))] = arg;
return ;
}
}
tmp211:;
g->unhandled[cast<uint32_t>((cast<uint32_t>(9895936ULL) | (cast<uint32_t>((method & cast<uint32_t>(65535ULL))))))]++;
}
}
// tools/platform/xbox/nv2a_pfifo.go:56:1
void xbox_Machine_runPusher(xbox_Machine* m){
{rrprof::Scope timing(1,"NV2A commands");
{
xbox_pusherState* p = (&m->push);
if (p->running) {
return ;
}
p->running = true;
auto tmp213=defer([&](){[&]()->void{
p->running = false;
}
();});
xbox_Machine_profPusherEnter(m);
auto tmp214=defer([&](){xbox_Machine_profPusherExit(m);});
m->pgraph->texRun++;
if (xbox_nvTrace) {
go_fmt_Printf(std::string("PUSH run: GET=%08X PUT=%08X\012",28),m->nv.dmaGet,m->nv.dmaPut);
}
int64_t words = cast<int64_t>(0ULL);
{;for (;(m->nv.dmaGet != m->nv.dmaPut);){
if (m->StopRequested) {
return ;
}
{
words++;
if ((words > cast<int64_t>(67108864ULL))) {
x86_CPU_Halt(m->CPU,std::string("nv2a: pusher exceeded %d words (GET=%08X PUT=%08X) \342\200\224 runaway push buffer",74),cast<int64_t>(67108864ULL),m->nv.dmaGet,m->nv.dmaPut);
return ;
}
}
uint32_t word = xbox_Machine_read32(m,m->nv.dmaGet);
if ((xbox_nvTrace && (p->count == cast<uint32_t>(0ULL)))) {
go_fmt_Printf(std::string("  PUSH @%08X cmd=%08X\012",22),m->nv.dmaGet,word);
}
m->nv.dmaGet += cast<uint32_t>(4ULL);
if ((p->count > cast<uint32_t>(0ULL))) {
xbox_Machine_pgraphMethod(m,p->subchan,p->method,word);
if ((!p->nonInc)) {
p->method += cast<uint32_t>(4ULL);
}
p->count--;
if (m->CPU->Halted) {
return ;
}
continue;
}
{
if (((cast<uint32_t>((word & cast<uint32_t>(3758292995ULL))) == cast<uint32_t>(0ULL)) && (word != cast<uint32_t>(0ULL)))){
p->method = cast<uint32_t>((word & cast<uint32_t>(8191ULL)));
p->subchan = cast<uint32_t>(((shr<uint32_t>(word,cast<int64_t>(13ULL))) & cast<uint32_t>(7ULL)));
p->count = cast<uint32_t>(((shr<uint32_t>(word,cast<int64_t>(18ULL))) & cast<uint32_t>(2047ULL)));
p->nonInc = false;
}
else if ((cast<uint32_t>((word & cast<uint32_t>(3758292995ULL))) == cast<uint32_t>(1073741824ULL))){
p->method = cast<uint32_t>((word & cast<uint32_t>(8191ULL)));
p->subchan = cast<uint32_t>(((shr<uint32_t>(word,cast<int64_t>(13ULL))) & cast<uint32_t>(7ULL)));
p->count = cast<uint32_t>(((shr<uint32_t>(word,cast<int64_t>(18ULL))) & cast<uint32_t>(2047ULL)));
p->nonInc = true;
}
else if ((cast<uint32_t>((word & cast<uint32_t>(3758096387ULL))) == cast<uint32_t>(536870912ULL))){
m->nv.dmaGet = cast<uint32_t>((word & cast<uint32_t>(536870911ULL)));
}
else if ((cast<uint32_t>((word & cast<uint32_t>(3ULL))) == cast<uint32_t>(1ULL))){
m->nv.dmaGet = cast<uint32_t>((word & cast<uint32_t>(4294967292ULL)));
}
else if ((cast<uint32_t>((word & cast<uint32_t>(3ULL))) == cast<uint32_t>(2ULL))){
if (p->subActive) {
x86_CPU_Halt(m->CPU,std::string("nv2a: nested push-buffer CALL at GET=%08X (word %08X)",53),cast<uint32_t>((m->nv.dmaGet - cast<uint32_t>(4ULL))),word);
return ;
}
p->subReturn = m->nv.dmaGet;
p->subActive = true;
m->nv.dmaGet = cast<uint32_t>((word & cast<uint32_t>(4294967292ULL)));
}
else if ((word == cast<uint32_t>(131072ULL))){
if ((!p->subActive)) {
x86_CPU_Halt(m->CPU,std::string("nv2a: push-buffer RETURN with no active CALL at GET=%08X",56),cast<uint32_t>((m->nv.dmaGet - cast<uint32_t>(4ULL))));
return ;
}
m->nv.dmaGet = p->subReturn;
p->subActive = false;
}
else if ((word == cast<uint32_t>(0ULL))){
}
else {
x86_CPU_Halt(m->CPU,std::string("nv2a: unrecognised push-buffer command %08X at GET=%08X",55),word,cast<uint32_t>((m->nv.dmaGet - cast<uint32_t>(4ULL))));
return ;
}
}
tmp215:;
}
}}
}
}
// tools/platform/xbox/nv2a_pgraph.go:161:1
xbox_pgraph* xbox_newPgraph(xbox_Machine* m){
{
xbox_pgraph* g = arenaNew(xbox_pgraph{m,{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},Map<xbox_texKey,xbox_texEntry*>{},{},{},{},{},Map<uint32_t,int64_t>{},Map<uint32_t,uint32_t>{},{},{},Map<uint32_t,int64_t>{},Map<uint32_t,xbox_zetaBucket*>{},{},{},{},{},{},{},{},{},{}});
{auto&& tmp216 = g->vtxAttr;
for(int64_t tmp217=0;tmp217<len(tmp216);++tmp217){
auto i=tmp217;g->vtxAttr[i] = std::array<float,4>{cast<float>(0.00000000000000000e+00),cast<float>(0.00000000000000000e+00),cast<float>(0.00000000000000000e+00),cast<float>(1.00000000000000000e+00)};
}}
return g;
}
}
// tools/platform/xbox/nv2a_pgraph.go:177:1
void xbox_pgraph_SetSurvey(xbox_pgraph* g,bool v){
{
g->survey = v;
}
}
// tools/platform/xbox/nv2a_pgraph.go:180:1
void xbox_Machine_pgraphMethod(xbox_Machine* m,uint32_t subchan,uint32_t method,uint32_t arg){
{
xbox_pgraph* g = m->pgraph;
g->Methods++;
if (bool(m->OnNVMethod)) {
m->OnNVMethod(m,subchan,method,arg);
}
if (m->stopAfterArmed) {
auto tmp218=defer([&](){[&]()->void{
{
m->stopAfterMethod--;
if ((m->stopAfterMethod <= cast<int64_t>(0ULL))) {
m->StopRequested = true;
}
}
}
();});
}
if ((method == cast<uint32_t>(0ULL))) {
g->SetObjs++;
g->subObject[cast<uint32_t>((subchan & cast<uint32_t>(7ULL)))] = arg;
g->subClass[cast<uint32_t>((subchan & cast<uint32_t>(7ULL)))] = xbox_Machine_ramhtClass(m,arg);
return ;
}
uint32_t class_ = g->subClass[cast<uint32_t>((subchan & cast<uint32_t>(7ULL)))];
if (g->survey) {
uint32_t key = cast<uint32_t>((shl<uint32_t>(class_,cast<int64_t>(16ULL)) | (cast<uint32_t>((method & cast<uint32_t>(65535ULL))))));
if ((get(g->seen,key) == cast<int64_t>(0ULL))) {
g->firstArg[key] = arg;
}
g->seen[key]++;
}
{
switch(class_){
case cast<uint32_t>(151ULL):{
xbox_pgraph_kelvinMethod(g,method,arg);
return ;
break;}
case cast<uint32_t>(98ULL):{
xbox_pgraph_surf2DMethod(g,method,arg);
return ;
break;}
case cast<uint32_t>(159ULL):{
xbox_pgraph_blitMethod(g,method,arg);
return ;
break;}
}}
g->unhandled[cast<uint32_t>((shl<uint32_t>(class_,cast<int64_t>(16ULL)) | (cast<uint32_t>((method & cast<uint32_t>(65535ULL))))))]++;
}
}
// tools/platform/xbox/nv2a_pgraph.go:296:1
void xbox_pgraph_DumpZetaHist(xbox_pgraph* g){
{
if ((!xbox_shadowTrace)) {
return ;
}
if (bool(g->shadowFrag)) {
xbox_shadowFragStats_print(g->shadowFrag);
g->shadowFrag = {};
}
Slice<uint32_t> offs = Slice<uint32_t>::make(cast<int64_t>(0ULL),len(g->zetaHist));
{auto&& tmp219 = g->zetaHist;
for(auto [tmp220,tmp221]:tmp219){
auto o=tmp220;offs = append(offs,Slice<uint32_t>{o});
}}
go_sort_Slice(offs,[&](int64_t i,int64_t j)->bool{
return (offs[i] < offs[j]);
}
);
go_fmt_Printf(std::string("SHADOW zeta-write census: %d distinct zeta offsets\012",51),len(offs));
{auto&& tmp222 = offs;
for(int64_t tmp223=0;tmp223<len(tmp222);++tmp223){
auto o=tmp222[tmp223];xbox_zetaBucket* b = get(g->zetaHist,o);
go_fmt_Printf(std::string("  zeta=%08X pix=%-9d draws=%-6d zmin=%06X zmax=%06X\012",52),o,b->pix,b->draws,cast<uint32_t>((b->zmin & cast<uint32_t>(16777215ULL))),cast<uint32_t>((b->zmax & cast<uint32_t>(16777215ULL))));
}}
}
}
// tools/platform/xbox/nv2a_pgraph.go:319:1
Slice<std::string> xbox_pgraph_SurveyReport(xbox_pgraph* g){
{
Slice<xbox_nv2a_pgraph_320_row> rows = Slice<xbox_nv2a_pgraph_320_row>::make(cast<int64_t>(0ULL),len(g->seen));
{auto&& tmp224 = g->seen;
for(auto [tmp225,tmp226]:tmp224){
auto k=tmp225;auto c=tmp226;rows = append(rows,Slice<xbox_nv2a_pgraph_320_row>{xbox_nv2a_pgraph_320_row{k,c}});
}}
go_sort_Slice(rows,[&](int64_t i,int64_t j)->bool{
if ((shr<uint32_t>(rows[i].key,cast<int64_t>(16ULL)) != shr<uint32_t>(rows[j].key,cast<int64_t>(16ULL)))) {
return (shr<uint32_t>(rows[i].key,cast<int64_t>(16ULL)) < shr<uint32_t>(rows[j].key,cast<int64_t>(16ULL)));
}
return (cast<uint32_t>((rows[i].key & cast<uint32_t>(65535ULL))) < cast<uint32_t>((rows[j].key & cast<uint32_t>(65535ULL))));
}
);
Slice<std::string> out = Slice<std::string>::make(cast<int64_t>(0ULL),cast<int64_t>((len(rows) + cast<int64_t>(2ULL))));
out = append(out,Slice<std::string>{go_fmt_Sprintf(std::string("PGRAPH survey: %d methods, %d SET_OBJECT, %d distinct (class,method)",68),g->Methods,g->SetObjs,len(rows))});
{auto&& tmp227 = rows;
for(int64_t tmp228=0;tmp228<len(tmp227);++tmp228){
auto r=tmp227[tmp228];auto tmp229 = std::make_tuple(shr<uint32_t>(r.key,cast<int64_t>(16ULL)),cast<uint32_t>((r.key & cast<uint32_t>(65535ULL))));
uint32_t class_ = std::get<0>(tmp229);
uint32_t mthd = std::get<1>(tmp229);
out = append(out,Slice<std::string>{go_fmt_Sprintf(std::string("  class %04X method %04X  x%-6d  firstArg=%08X",46),class_,mthd,r.count,get(g->firstArg,r.key))});
}}
return out;
}
}
// tools/platform/xbox/nv2a_ramht.go:24:1
uint32_t xbox_Machine_ramhtInstance(xbox_Machine* m,uint32_t handle){
{
if ((handle == cast<uint32_t>(0ULL))) {
return cast<uint32_t>(0ULL);
}
uint32_t base = cast<uint32_t>(4251975680ULL);
{uint32_t off = cast<uint32_t>(0ULL);for (;(off < cast<uint32_t>(4096ULL));off += cast<uint32_t>(8ULL)){
if ((xbox_Machine_read32(m,cast<uint32_t>((base + off))) != handle)) {
continue;
}
uint32_t ctx = xbox_Machine_read32(m,cast<uint32_t>((cast<uint32_t>((base + off)) + cast<uint32_t>(4ULL))));
return cast<uint32_t>((base + shl<uint32_t>((cast<uint32_t>((ctx & cast<uint32_t>(65535ULL)))),cast<int64_t>(4ULL))));
}
}return cast<uint32_t>(0ULL);
}
}
// tools/platform/xbox/nv2a_ramht.go:40:1
uint32_t xbox_Machine_ramhtClass(xbox_Machine* m,uint32_t handle){
{
uint32_t inst = xbox_Machine_ramhtInstance(m,handle);
if ((inst == cast<uint32_t>(0ULL))) {
return cast<uint32_t>(0ULL);
}
return cast<uint32_t>((xbox_Machine_read32(m,inst) & cast<uint32_t>(255ULL)));
}
}
// tools/platform/xbox/nv2a_ramht.go:55:1
std::tuple<uint32_t,uint32_t> xbox_Machine_dmaObjectTarget(xbox_Machine* m,uint32_t handle){
uint32_t base{};
uint32_t limit{};
{
uint32_t inst = xbox_Machine_ramhtInstance(m,handle);
if ((inst == cast<uint32_t>(0ULL))) {
return {cast<uint32_t>(0ULL),cast<uint32_t>(0ULL)};
}
uint32_t w0 = xbox_Machine_read32(m,inst);
uint32_t w1 = xbox_Machine_read32(m,cast<uint32_t>((inst + cast<uint32_t>(4ULL))));
uint32_t w2 = xbox_Machine_read32(m,cast<uint32_t>((inst + cast<uint32_t>(8ULL))));
return {cast<uint32_t>((((w2 & ~(cast<uint32_t>(4095ULL)))) + (cast<uint32_t>((shr<uint32_t>(w0,cast<int64_t>(20ULL)) & cast<uint32_t>(4095ULL)))))),w1};
}
}
// tools/platform/xbox/nv2a_raster.go:101:1
std::tuple<int64_t,int64_t> xbox_surfaceAAScale(uint32_t format){
int64_t ax{};
int64_t ay{};
{
{
switch(cast<uint32_t>((shr<uint32_t>(format,cast<int64_t>(12ULL)) & cast<uint32_t>(15ULL)))){
case cast<uint32_t>(1ULL):{
return {cast<int64_t>(2ULL),cast<int64_t>(1ULL)};
break;}
case cast<uint32_t>(2ULL):{
return {cast<int64_t>(2ULL),cast<int64_t>(2ULL)};
break;}
}}
return {cast<int64_t>(1ULL),cast<int64_t>(1ULL)};
}
}
// tools/platform/xbox/nv2a_raster.go:130:1
std::tuple<int64_t,int64_t,std::string> xbox_swizzleGeom(uint32_t format){
int64_t w{};
int64_t h{};
std::string reason{};
{
if (((cast<uint32_t>((shr<uint32_t>(format,cast<int64_t>(20ULL)) & cast<uint32_t>(15ULL))) != cast<uint32_t>(0ULL)) || (cast<uint32_t>((shr<uint32_t>(format,cast<int64_t>(28ULL)) & cast<uint32_t>(15ULL))) != cast<uint32_t>(0ULL)))) {
return {cast<int64_t>(0ULL),cast<int64_t>(0ULL),std::string("unexpected format high nibbles",30)};
}
w = shl<int64_t>(cast<int64_t>(1ULL),(cast<uint32_t>((shr<uint32_t>(format,cast<int64_t>(16ULL)) & cast<uint32_t>(15ULL)))));
h = shl<int64_t>(cast<int64_t>(1ULL),(cast<uint32_t>((shr<uint32_t>(format,cast<int64_t>(24ULL)) & cast<uint32_t>(15ULL)))));
if ((w != h)) {
return {w,h,std::string("non-square (width/height orientation unverified)",48)};
}
return {w,h,std::string("",0)};
}
}
// tools/platform/xbox/nv2a_raster.go:142:1
bool xbox_pgraph_decodeSwizzleExtent(xbox_pgraph* g,xbox_rasterState* st,uint32_t format){
{
auto tmp230 = g->m;
xbox_Machine* m = tmp230;
auto tmp231 = xbox_swizzleGeom(format);
int64_t w = std::get<0>(tmp231);
int64_t h = std::get<1>(tmp231);
std::string reason = std::get<2>(tmp231);
if ((reason != std::string("",0))) {
x86_CPU_Halt(m->CPU,std::string("nv2a: draw %d swizzled surface %dx%d (0x208=%08X) \342\200\224 %s",56),g->Draws,w,h,format,reason);
return false;
}
st->swizzle = true;
auto tmp232 = std::make_tuple(w,h);
st->swW = std::get<0>(tmp232);
st->swH = std::get<1>(tmp232);
int64_t clipW = cast<int64_t>(shr<uint32_t>(g->Regs[cast<int64_t>(128ULL)],cast<int64_t>(16ULL)));
int64_t clipH = cast<int64_t>(shr<uint32_t>(g->Regs[cast<int64_t>(129ULL)],cast<int64_t>(16ULL)));
if (((clipW > st->swW) || (clipH > st->swH))) {
x86_CPU_Halt(m->CPU,std::string("nv2a: draw %d swizzled extent %dx%d smaller than clip %dx%d (0x208=%08X)",72),g->Draws,st->swW,st->swH,clipW,clipH,format);
return false;
}
return true;
}
}
// tools/platform/xbox/nv2a_raster.go:165:1
bool xbox_pgraph_rasterStateDecode(xbox_pgraph* g,xbox_rasterState* st){
{
auto tmp233 = g->m;
xbox_Machine* m = tmp233;
uint32_t format = g->Regs[cast<int64_t>(130ULL)];
{
uint32_t cf = cast<uint32_t>((format & cast<uint32_t>(15ULL)));
if (((cf != cast<uint32_t>(8ULL)) && (cf != cast<uint32_t>(3ULL)))) {
x86_CPU_Halt(m->CPU,std::string("nv2a: draw %d into unmodelled color surface format %X (0x208=%08X)",66),g->Draws,cf,format);
return false;
}
}
{
uint32_t ty = cast<uint32_t>((shr<uint32_t>(format,cast<int64_t>(8ULL)) & cast<uint32_t>(15ULL)));
switch(ty){
case cast<uint32_t>(1ULL):{
st->swizzle = false;
break;}
case cast<uint32_t>(2ULL):{
if ((!xbox_pgraph_decodeSwizzleExtent(g,st,format))) {
return false;
}
break;}
default:{
x86_CPU_Halt(m->CPU,std::string("nv2a: draw %d into surface type %d (0x208=%08X) \342\200\224 unmodelled layout",69),g->Draws,ty,format);
return false;
break;}
}}
auto tmp234 = xbox_surfaceAAScale(format);
st->aaX = std::get<0>(tmp234);
st->aaY = std::get<1>(tmp234);
bool ok={};
auto tmp235 = xbox_pgraph_surfPhys(g,g->Regs[cast<int64_t>(132ULL)]);
st->colorPhys = std::get<0>(tmp235);
ok = std::get<1>(tmp235);
if ((!ok)) {
x86_CPU_Halt(m->CPU,std::string("nv2a: draw %d color surface %08X is not RAM",43),g->Draws,g->Regs[cast<int64_t>(132ULL)]);
return false;
}
st->colorPitch = cast<uint32_t>((g->Regs[cast<int64_t>(131ULL)] & cast<uint32_t>(65535ULL)));
st->zetaPitch = shr<uint32_t>(g->Regs[cast<int64_t>(131ULL)],cast<int64_t>(16ULL));
st->hasZeta = (cast<uint32_t>((shr<uint32_t>(format,cast<int64_t>(4ULL)) & cast<uint32_t>(15ULL))) != cast<uint32_t>(0ULL));
if (st->hasZeta) {
auto tmp236 = xbox_pgraph_surfPhys(g,g->Regs[cast<int64_t>(133ULL)]);
st->zetaPhys = std::get<0>(tmp236);
ok = std::get<1>(tmp236);
if ((!ok)) {
st->hasZeta = false;
}
{
uint32_t zf = cast<uint32_t>((shr<uint32_t>(format,cast<int64_t>(4ULL)) & cast<uint32_t>(15ULL)));
if (((zf != cast<uint32_t>(2ULL)) && st->hasZeta)) {
x86_CPU_Halt(m->CPU,std::string("nv2a: draw %d zeta format %d unmodelled (0x208=%08X)",52),g->Draws,zf,format);
return false;
}
}
}
uint32_t clipH = g->Regs[cast<int64_t>(128ULL)];
uint32_t clipV = g->Regs[cast<int64_t>(129ULL)];
st->surfW = cast<int64_t>((cast<int64_t>(shr<uint32_t>(clipH,cast<int64_t>(16ULL))) * st->aaX));
st->surfH = cast<int64_t>((cast<int64_t>(shr<uint32_t>(clipV,cast<int64_t>(16ULL))) * st->aaY));
st->x0 = cast<int64_t>((cast<int64_t>(cast<uint32_t>((clipH & cast<uint32_t>(65535ULL)))) * st->aaX));
st->y0 = cast<int64_t>((cast<int64_t>(cast<uint32_t>((clipV & cast<uint32_t>(65535ULL)))) * st->aaY));
st->x1 = cast<int64_t>((st->x0 + st->surfW));
st->y1 = cast<int64_t>((st->y0 + st->surfH));
if ((g->Regs[cast<int64_t>(173ULL)] == cast<uint32_t>(0ULL))) {
uint32_t wh = g->Regs[cast<int64_t>(176ULL)];
uint32_t wv = g->Regs[cast<int64_t>(184ULL)];
if (((wh != cast<uint32_t>(0ULL)) || (wv != cast<uint32_t>(0ULL)))) {
auto tmp237 = std::make_tuple(cast<int64_t>((cast<int64_t>(cast<uint32_t>((wh & cast<uint32_t>(65535ULL)))) * st->aaX)),cast<int64_t>(((cast<int64_t>((cast<int64_t>(shr<uint32_t>(wh,cast<int64_t>(16ULL))) + cast<int64_t>(1ULL)))) * st->aaX)));
int64_t x0 = std::get<0>(tmp237);
int64_t x1 = std::get<1>(tmp237);
auto tmp238 = std::make_tuple(cast<int64_t>((cast<int64_t>(cast<uint32_t>((wv & cast<uint32_t>(65535ULL)))) * st->aaY)),cast<int64_t>(((cast<int64_t>((cast<int64_t>(shr<uint32_t>(wv,cast<int64_t>(16ULL))) + cast<int64_t>(1ULL)))) * st->aaY)));
int64_t y0 = std::get<0>(tmp238);
int64_t y1 = std::get<1>(tmp238);
if ((x0 > st->x0)) {
st->x0 = x0;
}
if ((y0 > st->y0)) {
st->y0 = y0;
}
if ((x1 < st->x1)) {
st->x1 = x1;
}
if ((y1 < st->y1)) {
st->y1 = y1;
}
}
}
st->depthTest = ((cast<uint32_t>((g->Regs[cast<int64_t>(195ULL)] & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL)) && st->hasZeta);
st->depthWrite = (st->depthTest && (cast<uint32_t>((g->Regs[cast<int64_t>(215ULL)] & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL)));
st->depthFunc = g->Regs[cast<int64_t>(213ULL)];
st->alphaTest = (cast<uint32_t>((g->Regs[cast<int64_t>(192ULL)] & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL));
st->alphaFunc = g->Regs[cast<int64_t>(207ULL)];
st->alphaRef = (cast<float>(cast<uint32_t>((g->Regs[cast<int64_t>(208ULL)] & cast<uint32_t>(255ULL)))) / cast<float>(2.55000000000000000e+02));
st->blend = (cast<uint32_t>((g->Regs[cast<int64_t>(193ULL)] & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL));
st->blendSrc = g->Regs[cast<int64_t>(209ULL)];
st->blendDst = g->Regs[cast<int64_t>(210ULL)];
st->blendEq = g->Regs[cast<int64_t>(212ULL)];
uint32_t bc = g->Regs[cast<int64_t>(211ULL)];
st->blendConst = std::array<float,4>{(cast<float>(cast<uint32_t>((shr<uint32_t>(bc,cast<int64_t>(24ULL)) & cast<uint32_t>(255ULL)))) / cast<float>(2.55000000000000000e+02)),(cast<float>(cast<uint32_t>((shr<uint32_t>(bc,cast<int64_t>(16ULL)) & cast<uint32_t>(255ULL)))) / cast<float>(2.55000000000000000e+02)),(cast<float>(cast<uint32_t>((shr<uint32_t>(bc,cast<int64_t>(8ULL)) & cast<uint32_t>(255ULL)))) / cast<float>(2.55000000000000000e+02)),(cast<float>(cast<uint32_t>((bc & cast<uint32_t>(255ULL)))) / cast<float>(2.55000000000000000e+02))};
uint32_t cm = g->Regs[cast<int64_t>(214ULL)];
st->colorMask = std::array<bool,4>{(cast<uint32_t>((shr<uint32_t>(cm,cast<int64_t>(24ULL)) & cast<uint32_t>(255ULL))) != cast<uint32_t>(0ULL)),(cast<uint32_t>((shr<uint32_t>(cm,cast<int64_t>(16ULL)) & cast<uint32_t>(255ULL))) != cast<uint32_t>(0ULL)),(cast<uint32_t>((shr<uint32_t>(cm,cast<int64_t>(8ULL)) & cast<uint32_t>(255ULL))) != cast<uint32_t>(0ULL)),(cast<uint32_t>((cm & cast<uint32_t>(255ULL))) != cast<uint32_t>(0ULL))};
st->cull = (cast<uint32_t>((g->Regs[cast<int64_t>(194ULL)] & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL));
st->cullFace = g->Regs[cast<int64_t>(231ULL)];
st->frontCW = (g->Regs[cast<int64_t>(232ULL)] == cast<uint32_t>(2304ULL));
st->flatShade = (g->Regs[cast<int64_t>(223ULL)] == cast<uint32_t>(7424ULL));
xbox_pgraph_combDecode(g,(&st->comb));
return xbox_pgraph_texStateDecode(g,st);
}
}
// tools/platform/xbox/nv2a_raster.go:274:1
std::tuple<uint32_t,bool> xbox_pgraph_surfPhys(xbox_pgraph* g,uint32_t addr){
{
auto tmp239 = xbox_Machine_translate(g->m,addr);
uint32_t phys = std::get<0>(tmp239);
bool mmio = std::get<1>(tmp239);
bool ok = std::get<2>(tmp239);
if (((!ok) || mmio)) {
return {cast<uint32_t>(0ULL),false};
}
return {phys,true};
}
}
// tools/platform/xbox/nv2a_raster.go:291:1
uint32_t xbox_rasterState_colorAddr(xbox_rasterState* st,int64_t px,int64_t py){
{
if (st->swizzle) {
return cast<uint32_t>((st->colorPhys + cast<uint32_t>((cast<uint32_t>(xbox_swizzleOffset(px,py,st->swW,st->swH)) * cast<uint32_t>(4ULL)))));
}
return cast<uint32_t>((cast<uint32_t>((st->colorPhys + cast<uint32_t>((cast<uint32_t>(py) * st->colorPitch)))) + cast<uint32_t>((cast<uint32_t>(px) * cast<uint32_t>(4ULL)))));
}
}
// tools/platform/xbox/nv2a_raster.go:298:1
uint32_t xbox_rasterState_zetaAddr(xbox_rasterState* st,int64_t px,int64_t py){
{
if (st->swizzle) {
return cast<uint32_t>((st->zetaPhys + cast<uint32_t>((cast<uint32_t>(xbox_swizzleOffset(px,py,st->swW,st->swH)) * cast<uint32_t>(4ULL)))));
}
return cast<uint32_t>((cast<uint32_t>((st->zetaPhys + cast<uint32_t>((cast<uint32_t>(py) * st->zetaPitch)))) + cast<uint32_t>((cast<uint32_t>(px) * cast<uint32_t>(4ULL)))));
}
}
// tools/platform/xbox/nv2a_raster.go:309:1
void xbox_pgraph_rasterTri(xbox_pgraph* g,xbox_kelvinVtx* v0,xbox_kelvinVtx* v1,xbox_kelvinVtx* v2,xbox_rstats* rs){
{
xbox_pgraph_rasterTriBand(g,v0,v1,v2,cast<int64_t>(0ULL),cast<int64_t>(1ULL),rs);
}
}
// tools/platform/xbox/nv2a_raster.go:319:1
void xbox_pgraph_mergeStats(xbox_pgraph* g,xbox_rstats* rs){
{
g->pixWritten += rs->written;
g->pixZRej += rs->zRej;
g->pixARej += rs->aRej;
}
}
// tools/platform/xbox/nv2a_raster.go:329:1
bool xbox_pgraph_rasterParallelOK(xbox_pgraph* g,xbox_rasterState* st,Slice<xbox_kelvinVtx> verts,Slice<std::array<int64_t,3>> tris){
{
if ((xbox_rasterWorkers() < cast<int64_t>(2ULL))) {
return false;
}
if (((((bool(g->m->OnPixel) || xbox_shadowTrace) || xbox_shadowFragTrace) || xbox_lowWriteTrace) || (g->ffFragHalt != std::string("",0)))) {
return false;
}
{int64_t u = cast<int64_t>(0ULL);for (;(u < cast<int64_t>(4ULL));u++){
if (((!st->texEnable[u]) || (!st->texImg[u]))) {
continue;
}
if ((st->texImg[u]->cube && (st->texStage[u] != cast<uint32_t>(3ULL)))) {
return false;
}
if ((bool(st->texImg[u]->depth) && (st->texStage[u] != cast<uint32_t>(2ULL)))) {
return false;
}
}
}int64_t area = cast<int64_t>(0ULL);
{auto&& tmp240 = tris;
for(int64_t tmp241=0;tmp241<len(tmp240);++tmp241){
auto t=tmp240[tmp241];auto tmp242 = std::make_tuple((&verts[t[cast<int64_t>(0ULL)]]),(&verts[t[cast<int64_t>(1ULL)]]),(&verts[t[cast<int64_t>(2ULL)]]));
xbox_kelvinVtx* v0 = std::get<0>(tmp242);
xbox_kelvinVtx* v1 = std::get<1>(tmp242);
xbox_kelvinVtx* v2 = std::get<2>(tmp242);
int64_t minX = cast<int64_t>(xbox_min3(cast<double>(v0->pos[cast<int64_t>(0ULL)]),cast<double>(v1->pos[cast<int64_t>(0ULL)]),cast<double>(v2->pos[cast<int64_t>(0ULL)])));
int64_t maxX = cast<int64_t>((cast<int64_t>(xbox_max3(cast<double>(v0->pos[cast<int64_t>(0ULL)]),cast<double>(v1->pos[cast<int64_t>(0ULL)]),cast<double>(v2->pos[cast<int64_t>(0ULL)]))) + cast<int64_t>(1ULL)));
int64_t minY = cast<int64_t>(xbox_min3(cast<double>(v0->pos[cast<int64_t>(1ULL)]),cast<double>(v1->pos[cast<int64_t>(1ULL)]),cast<double>(v2->pos[cast<int64_t>(1ULL)])));
int64_t maxY = cast<int64_t>((cast<int64_t>(xbox_max3(cast<double>(v0->pos[cast<int64_t>(1ULL)]),cast<double>(v1->pos[cast<int64_t>(1ULL)]),cast<double>(v2->pos[cast<int64_t>(1ULL)]))) + cast<int64_t>(1ULL)));
if ((minX < st->x0)) {
minX = st->x0;
}
if ((minY < st->y0)) {
minY = st->y0;
}
if ((maxX > st->x1)) {
maxX = st->x1;
}
if ((maxY > st->y1)) {
maxY = st->y1;
}
if (((minX < maxX) && (minY < maxY))) {
area += cast<int64_t>(((cast<int64_t>((maxX - minX))) * (cast<int64_t>((maxY - minY)))));
}
if ((area >= cast<int64_t>(16384ULL))) {
return true;
}
}}
return false;
}
}
// tools/platform/xbox/nv2a_raster.go:420:1
void xbox_pgraph_rasterTriBand(xbox_pgraph* g,xbox_kelvinVtx* v0,xbox_kelvinVtx* v1,xbox_kelvinVtx* v2,int64_t lane,int64_t stride,xbox_rstats* rs){
{
if (g->m->CPU->Halted) {
return ;
}
xbox_rasterState* st = (&g->rast);
if ((!g->rastValid)) {
if ((!xbox_pgraph_rasterStateDecode(g,st))) {
return ;
}
g->rastValid = true;
}
auto tmp243 = std::make_tuple(cast<double>(v0->pos[cast<int64_t>(0ULL)]),cast<double>(v0->pos[cast<int64_t>(1ULL)]));
double x0 = std::get<0>(tmp243);
double y0 = std::get<1>(tmp243);
auto tmp244 = std::make_tuple(cast<double>(v1->pos[cast<int64_t>(0ULL)]),cast<double>(v1->pos[cast<int64_t>(1ULL)]));
double x1 = std::get<0>(tmp244);
double y1 = std::get<1>(tmp244);
auto tmp245 = std::make_tuple(cast<double>(v2->pos[cast<int64_t>(0ULL)]),cast<double>(v2->pos[cast<int64_t>(1ULL)]));
double x2 = std::get<0>(tmp245);
double y2 = std::get<1>(tmp245);
double area = ((((x1 - x0)) * ((y2 - y0))) - (((y1 - y0)) * ((x2 - x0))));
if ((area == cast<double>(0.00000000000000000e+00))) {
return ;
}
if (st->cull) {
bool front = (((area > cast<double>(0.00000000000000000e+00))) == st->frontCW);
{
switch(st->cullFace){
case cast<uint32_t>(1028ULL):{
if (front) {
return ;
}
break;}
case cast<uint32_t>(1032ULL):{
return ;
break;}
default:{
if ((!front)) {
return ;
}
break;}
}}
}
double sign = cast<double>(1.00000000000000000e+00);
if ((area < cast<double>(0.00000000000000000e+00))) {
auto tmp246 = std::make_tuple(cast<double>(-1.00000000000000000e+00),cast<double>(-area));
sign = std::get<0>(tmp246);
area = std::get<1>(tmp246);
}
int64_t minX = cast<int64_t>(xbox_min3(x0,x1,x2));
int64_t maxX = cast<int64_t>((cast<int64_t>(xbox_max3(x0,x1,x2)) + cast<int64_t>(1ULL)));
int64_t minY = cast<int64_t>(xbox_min3(y0,y1,y2));
int64_t maxY = cast<int64_t>((cast<int64_t>(xbox_max3(y0,y1,y2)) + cast<int64_t>(1ULL)));
if ((minX < st->x0)) {
minX = st->x0;
}
if ((minY < st->y0)) {
minY = st->y0;
}
if ((maxX > st->x1)) {
maxX = st->x1;
}
if ((maxY > st->y1)) {
maxY = st->y1;
}
if (((minX >= maxX) || (minY >= maxY))) {
return ;
}
auto tmp247 = std::make_tuple(xbox_invW(v0->pos[cast<int64_t>(3ULL)]),xbox_invW(v1->pos[cast<int64_t>(3ULL)]),xbox_invW(v2->pos[cast<int64_t>(3ULL)]));
float iw0 = std::get<0>(tmp247);
float iw1 = std::get<1>(tmp247);
float iw2 = std::get<2>(tmp247);
auto edge = [&](double px,double py,double ex0,double ey0,double ex1,double ey1)->double{
return (sign * (((((ex1 - ex0)) * ((py - ey0))) - (((ey1 - ey0)) * ((px - ex0))))));
}
;
auto topLeft = [&](double ex0,double ey0,double ex1,double ey1)->bool{
auto tmp248 = std::make_tuple((sign * ((ex1 - ex0))),(sign * ((ey1 - ey0))));
double dx = std::get<0>(tmp248);
double dy = std::get<1>(tmp248);
return ((dy < cast<double>(0.00000000000000000e+00)) || (((dy == cast<double>(0.00000000000000000e+00)) && (dx > cast<double>(0.00000000000000000e+00)))));
}
;
bool tl0 = topLeft(x1,y1,x2,y2);
bool tl1 = topLeft(x2,y2,x0,y0);
bool tl2 = topLeft(x0,y0,x1,y1);
if ((stride > cast<int64_t>(1ULL))) {
int64_t off = modi<int64_t>((cast<int64_t>((cast<int64_t>((lane - modi<int64_t>(minY,stride))) + stride))),stride);
minY += off;
}
{int64_t py = minY;for (;(py < maxY);py += stride){
double fy = (cast<double>(py) + cast<double>(5.00000000000000000e-01));
{int64_t px = minX;for (;(px < maxX);px++){
double fx = (cast<double>(px) + cast<double>(5.00000000000000000e-01));
double w0 = edge(fx,fy,x1,y1,x2,y2);
double w1 = edge(fx,fy,x2,y2,x0,y0);
double w2 = edge(fx,fy,x0,y0,x1,y1);
if ((((w0 < cast<double>(0.00000000000000000e+00)) || (w1 < cast<double>(0.00000000000000000e+00))) || (w2 < cast<double>(0.00000000000000000e+00)))) {
continue;
}
if ((((((w0 == cast<double>(0.00000000000000000e+00)) && (!tl0))) || (((w1 == cast<double>(0.00000000000000000e+00)) && (!tl1)))) || (((w2 == cast<double>(0.00000000000000000e+00)) && (!tl2))))) {
continue;
}
auto tmp249 = std::make_tuple(cast<float>((w0 / area)),cast<float>((w1 / area)),cast<float>((w2 / area)));
float b0 = std::get<0>(tmp249);
float b1 = std::get<1>(tmp249);
float b2 = std::get<2>(tmp249);
xbox_pgraph_shadePixel(g,st,px,py,b0,b1,b2,iw0,iw1,iw2,v0,v1,v2,rs);
if (g->m->CPU->Halted) {
return ;
}
}
}}
}}
}
// tools/platform/xbox/nv2a_raster.go:532:1
void xbox_pgraph_shadePixel(xbox_pgraph* g,xbox_rasterState* st,int64_t px,int64_t py,float b0,float b1,float b2,float iw0,float iw1,float iw2,xbox_kelvinVtx* v0,xbox_kelvinVtx* v1,xbox_kelvinVtx* v2,xbox_rstats* rs){
{
auto tmp250 = g->m;
xbox_Machine* m = tmp250;
if ((g->ffFragHalt != std::string("",0))) {
x86_CPU_Halt(m->CPU,std::string("nv2a: draw %d fixed-function fragment at (%d,%d) \342\200\224 %s",55),g->Draws,px,py,g->ffFragHalt);
return ;
}
uint32_t zAddr={};
uint32_t zOld={};
if ((st->depthTest || st->depthWrite)) {
zAddr = xbox_rasterState_zetaAddr(st,px,py);
if ((cast<uint32_t>((zAddr + cast<uint32_t>(4ULL))) > cast<uint32_t>(len(m->RAM)))) {
return ;
}
}
if (((xbox_lowWriteTrace && (g->Draws != g->lowWriteDraw)) && (((xbox_rasterState_colorAddr(st,px,py) < cast<uint32_t>(7340032ULL)) || ((st->depthWrite && (zAddr < cast<uint32_t>(7340032ULL)))))))) {
g->lowWriteDraw = g->Draws;
go_fmt_Printf(std::string("LOWWRITE draw=%d px=%d py=%d colorPhys=%08X zetaPhys=%08X pitch=%d/%d fmt=%08X ztest=%v zwrite=%v zfunc=%03X prim=%d\012",117),g->Draws,px,py,st->colorPhys,st->zetaPhys,st->colorPitch,st->zetaPitch,g->Regs[cast<int64_t>(130ULL)],st->depthTest,st->depthWrite,st->depthFunc,g->prim);
}
float z = (((b0 * v0->pos[cast<int64_t>(2ULL)]) + (b1 * v1->pos[cast<int64_t>(2ULL)])) + (b2 * v2->pos[cast<int64_t>(2ULL)]));
uint32_t zi = cast<uint32_t>(0ULL);
if ((z > cast<float>(0.00000000000000000e+00))) {
if ((z >= cast<float>(1.67772150000000000e+07))) {
zi = cast<uint32_t>(16777215ULL);
}
else {
zi = cast<uint32_t>(z);
}
}
if (st->depthTest) {
uint32_t stored = cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(m->RAM[zAddr]) | shl<uint32_t>(cast<uint32_t>(m->RAM[cast<uint32_t>((zAddr + cast<uint32_t>(1ULL)))]),cast<int64_t>(8ULL)))) | shl<uint32_t>(cast<uint32_t>(m->RAM[cast<uint32_t>((zAddr + cast<uint32_t>(2ULL)))]),cast<int64_t>(16ULL)))) | shl<uint32_t>(cast<uint32_t>(m->RAM[cast<uint32_t>((zAddr + cast<uint32_t>(3ULL)))]),cast<int64_t>(24ULL))));
zOld = shr<uint32_t>(stored,cast<int64_t>(8ULL));
if ((!xbox_depthPass(st->depthFunc,zi,zOld))) {
rs->zRej++;
if (bool(m->OnPixel)) {
m->OnPixel(cast<uint32_t>(px),cast<uint32_t>(py),xbox_PixelEvent{{},true,{},{},{},{},{}});
}
return ;
}
}
float iw = (((b0 * iw0) + (b1 * iw1)) + (b2 * iw2));
if ((iw == cast<float>(0.00000000000000000e+00))) {
iw = cast<float>(1.00000000000000000e+00);
}
auto pc4 = [&](std::array<float,4>* a0,std::array<float,4>* a1,std::array<float,4>* a2)->std::array<float,4>{
std::array<float,4> out={};
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(4ULL));i++){
out[i] = ((((((b0 * (*a0)[i]) * iw0) + ((b1 * (*a1)[i]) * iw1)) + ((b2 * (*a2)[i]) * iw2))) / iw);
}
}return out;
}
;
xbox_combInput in={};
if (st->flatShade) {
auto tmp251 = std::make_tuple(v2->d0,v2->d1);
in.col0 = std::get<0>(tmp251);
in.col1 = std::get<1>(tmp251);
}
else {
in.col0 = pc4((&v0->d0),(&v1->d0),(&v2->d0));
in.col1 = pc4((&v0->d1),(&v1->d1),(&v2->d1));
}
std::array<std::array<float,4>,4> uvw={};
{int64_t u = cast<int64_t>(0ULL);for (;(u < cast<int64_t>(4ULL));u++){
uvw[u] = pc4((&v0->uv[u]),(&v1->uv[u]),(&v2->uv[u]));
}
}{int64_t u = cast<int64_t>(0ULL);for (;(u < cast<int64_t>(4ULL));u++){
if (((!st->texEnable[u]) || (!st->texImg[u]))) {
in.tex[u] = std::array<float,4>{cast<float>(1.00000000000000000e+00),cast<float>(1.00000000000000000e+00),cast<float>(1.00000000000000000e+00),cast<float>(1.00000000000000000e+00)};
continue;
}
if (st->texImg[u]->cube) {
{
switch(st->texStage[u]){
case cast<uint32_t>(3ULL):{
in.tex[u] = xbox_pgraph_texSampleCube(g,st,u,uvw[u][cast<int64_t>(0ULL)],uvw[u][cast<int64_t>(1ULL)],uvw[u][cast<int64_t>(2ULL)]);
break;}
case cast<uint32_t>(12ULL):{
in.tex[u] = xbox_pgraph_texReflectSpecular(g,st,u,(&uvw),(&in));
break;}
default:{
if (xbox_texShaderTrace) {
xbox_pgraph_dumpTexShaderConfig(g,st,v0,v1,v2,b0,b1,b2,iw0,iw1,iw2,iw,px,py);
}
x86_CPU_Halt(g->m->CPU,std::string("nv2a: texture unit %d samples a cube map in stage mode %d \342\200\224 only CUBEMAP (3) and DOT_REFLECT_SPECULAR (12) are modelled",121),u,st->texStage[u]);
return ;
break;}
}}
continue;
}
if (bool(st->texImg[u]->depth)) {
if ((st->texStage[u] != cast<uint32_t>(2ULL))) {
x86_CPU_Halt(g->m->CPU,std::string("nv2a: texture unit %d samples a depth texture in stage mode %d \342\200\224 only PROJECTIVE (2) is modelled",98),u,st->texStage[u]);
return ;
}
float q = uvw[u][cast<int64_t>(3ULL)];
if ((q == cast<float>(0.00000000000000000e+00))) {
q = cast<float>(1.00000000000000000e+00);
}
auto tmp252 = std::make_tuple((uvw[u][cast<int64_t>(0ULL)] / q),(uvw[u][cast<int64_t>(1ULL)] / q),(uvw[u][cast<int64_t>(2ULL)] / q));
float s = std::get<0>(tmp252);
float t = std::get<1>(tmp252);
float r = std::get<2>(tmp252);
if (xbox_shadowFragTrace) {
xbox_pgraph_traceShadowFrag(g,st,u,px,py,s,t,r,q);
}
in.tex[u] = xbox_pgraph_texSampleShadow(g,st,u,s,t,r);
continue;
}
auto tmp253 = std::make_tuple(uvw[u][cast<int64_t>(0ULL)],uvw[u][cast<int64_t>(1ULL)]);
float s = std::get<0>(tmp253);
float t = std::get<1>(tmp253);
if (((st->texStage[u] == cast<uint32_t>(2ULL)) || ((((st->texStage[u] == cast<uint32_t>(1ULL)) && (uvw[u][cast<int64_t>(3ULL)] != cast<float>(0.00000000000000000e+00))) && (uvw[u][cast<int64_t>(3ULL)] != cast<float>(1.00000000000000000e+00)))))) {
auto tmp254 = std::make_tuple((s / uvw[u][cast<int64_t>(3ULL)]),(t / uvw[u][cast<int64_t>(3ULL)]));
s = std::get<0>(tmp254);
t = std::get<1>(tmp254);
}
in.tex[u] = xbox_pgraph_texSample(g,st,u,s,t);
}
}std::array<float,4> col = xbox_pgraph_combine(g,(&st->comb),(&in));
if ((st->alphaTest && (!xbox_alphaPass(st->alphaFunc,col[cast<int64_t>(3ULL)],st->alphaRef)))) {
rs->aRej++;
if (bool(m->OnPixel)) {
m->OnPixel(cast<uint32_t>(px),cast<uint32_t>(py),xbox_PixelEvent{{},{},true,xbox_u8(col[cast<int64_t>(0ULL)]),xbox_u8(col[cast<int64_t>(1ULL)]),xbox_u8(col[cast<int64_t>(2ULL)]),xbox_u8(col[cast<int64_t>(3ULL)])});
}
return ;
}
uint32_t cAddr = xbox_rasterState_colorAddr(st,px,py);
if ((cast<uint32_t>((cAddr + cast<uint32_t>(4ULL))) > cast<uint32_t>(len(m->RAM)))) {
return ;
}
if (st->blend) {
float db = (cast<float>(m->RAM[cAddr]) / cast<float>(2.55000000000000000e+02));
float dg = (cast<float>(m->RAM[cast<uint32_t>((cAddr + cast<uint32_t>(1ULL)))]) / cast<float>(2.55000000000000000e+02));
float dr = (cast<float>(m->RAM[cast<uint32_t>((cAddr + cast<uint32_t>(2ULL)))]) / cast<float>(2.55000000000000000e+02));
float da = (cast<float>(m->RAM[cast<uint32_t>((cAddr + cast<uint32_t>(3ULL)))]) / cast<float>(2.55000000000000000e+02));
std::array<float,4> dst = std::array<float,4>{dr,dg,db,da};
col = xbox_blendPixel(st,col,dst);
}
if (st->colorMask[cast<int64_t>(3ULL)]) {
m->RAM[cAddr] = xbox_u8(col[cast<int64_t>(2ULL)]);
if(rrcapture::trace.active)rrXboxWrite(g->m,m->RAM,cAddr);
}
if (st->colorMask[cast<int64_t>(2ULL)]) {
m->RAM[cast<uint32_t>((cAddr + cast<uint32_t>(1ULL)))] = xbox_u8(col[cast<int64_t>(1ULL)]);
if(rrcapture::trace.active)rrXboxWrite(g->m,m->RAM,cast<uint32_t>((cAddr + cast<uint32_t>(1ULL))));
}
if (st->colorMask[cast<int64_t>(1ULL)]) {
m->RAM[cast<uint32_t>((cAddr + cast<uint32_t>(2ULL)))] = xbox_u8(col[cast<int64_t>(0ULL)]);
if(rrcapture::trace.active)rrXboxWrite(g->m,m->RAM,cast<uint32_t>((cAddr + cast<uint32_t>(2ULL))));
}
if (st->colorMask[cast<int64_t>(0ULL)]) {
m->RAM[cast<uint32_t>((cAddr + cast<uint32_t>(3ULL)))] = xbox_u8(col[cast<int64_t>(3ULL)]);
if(rrcapture::trace.active)rrXboxWrite(g->m,m->RAM,cast<uint32_t>((cAddr + cast<uint32_t>(3ULL))));
}
if (st->depthWrite) {
uint32_t stored = cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(m->RAM[zAddr]) | shl<uint32_t>(cast<uint32_t>(m->RAM[cast<uint32_t>((zAddr + cast<uint32_t>(1ULL)))]),cast<int64_t>(8ULL)))) | shl<uint32_t>(cast<uint32_t>(m->RAM[cast<uint32_t>((zAddr + cast<uint32_t>(2ULL)))]),cast<int64_t>(16ULL)))) | shl<uint32_t>(cast<uint32_t>(m->RAM[cast<uint32_t>((zAddr + cast<uint32_t>(3ULL)))]),cast<int64_t>(24ULL))));
stored = cast<uint32_t>((shl<uint32_t>(zi,cast<int64_t>(8ULL)) | cast<uint32_t>((stored & cast<uint32_t>(255ULL)))));
m->RAM[zAddr] = cast<uint8_t>(stored);
if(rrcapture::trace.active)rrXboxWrite(g->m,m->RAM,zAddr);
m->RAM[cast<uint32_t>((zAddr + cast<uint32_t>(1ULL)))] = cast<uint8_t>(shr<uint32_t>(stored,cast<int64_t>(8ULL)));
if(rrcapture::trace.active)rrXboxWrite(g->m,m->RAM,cast<uint32_t>((zAddr + cast<uint32_t>(1ULL))));
m->RAM[cast<uint32_t>((zAddr + cast<uint32_t>(2ULL)))] = cast<uint8_t>(shr<uint32_t>(stored,cast<int64_t>(16ULL)));
if(rrcapture::trace.active)rrXboxWrite(g->m,m->RAM,cast<uint32_t>((zAddr + cast<uint32_t>(2ULL))));
m->RAM[cast<uint32_t>((zAddr + cast<uint32_t>(3ULL)))] = cast<uint8_t>(shr<uint32_t>(stored,cast<int64_t>(24ULL)));
if(rrcapture::trace.active)rrXboxWrite(g->m,m->RAM,cast<uint32_t>((zAddr + cast<uint32_t>(3ULL))));
if (xbox_shadowTrace) {
uint32_t zo = g->Regs[cast<int64_t>(133ULL)];
xbox_zetaBucket* b = get(g->zetaHist,zo);
if ((!b)) {
b = arenaNew(xbox_zetaBucket{{},cast<uint32_t>(4294967295ULL),{},{},{}});
g->zetaHist[zo] = b;
}
b->pix++;
if ((zi < b->zmin)) {
b->zmin = zi;
}
if ((zi > b->zmax)) {
b->zmax = zi;
}
if ((b->lastDraw != g->Draws)) {
b->draws++;
b->lastDraw = g->Draws;
}
}
}
rs->written++;
if (bool(m->OnPixel)) {
m->OnPixel(cast<uint32_t>(px),cast<uint32_t>(py),xbox_PixelEvent{true,{},{},xbox_u8(col[cast<int64_t>(0ULL)]),xbox_u8(col[cast<int64_t>(1ULL)]),xbox_u8(col[cast<int64_t>(2ULL)]),xbox_u8(col[cast<int64_t>(3ULL)])});
}
}
}
// tools/platform/xbox/nv2a_raster.go:733:1
bool xbox_depthPass(uint32_t fn,uint32_t z,uint32_t old){
{
{
switch(fn){
case cast<uint32_t>(512ULL):{
return false;
break;}
case cast<uint32_t>(513ULL):{
return (z < old);
break;}
case cast<uint32_t>(514ULL):{
return (z == old);
break;}
case cast<uint32_t>(515ULL):{
return (z <= old);
break;}
case cast<uint32_t>(516ULL):{
return (z > old);
break;}
case cast<uint32_t>(517ULL):{
return (z != old);
break;}
case cast<uint32_t>(518ULL):{
return (z >= old);
break;}
default:{
return true;
break;}
}}
}
}
// tools/platform/xbox/nv2a_raster.go:754:1
bool xbox_alphaPass(uint32_t fn,float a,float ref){
{
{
switch(fn){
case cast<uint32_t>(512ULL):{
return false;
break;}
case cast<uint32_t>(513ULL):{
return (a < ref);
break;}
case cast<uint32_t>(514ULL):{
return (a == ref);
break;}
case cast<uint32_t>(515ULL):{
return (a <= ref);
break;}
case cast<uint32_t>(516ULL):{
return (a > ref);
break;}
case cast<uint32_t>(517ULL):{
return (a != ref);
break;}
case cast<uint32_t>(518ULL):{
return (a >= ref);
break;}
default:{
return true;
break;}
}}
}
}
// tools/platform/xbox/nv2a_raster.go:777:1
std::array<float,4> xbox_blendPixel(xbox_rasterState* st,std::array<float,4> src,std::array<float,4> dst){
{
auto factor = [&](uint32_t f,bool other)->std::array<float,4>{
{
switch(f){
case cast<uint32_t>(0ULL):{
return std::array<float,4>{};
break;}
case cast<uint32_t>(1ULL):{
return std::array<float,4>{cast<float>(1.00000000000000000e+00),cast<float>(1.00000000000000000e+00),cast<float>(1.00000000000000000e+00),cast<float>(1.00000000000000000e+00)};
break;}
case cast<uint32_t>(768ULL):{
return src;
break;}
case cast<uint32_t>(769ULL):{
return std::array<float,4>{(cast<float>(1.00000000000000000e+00) - src[cast<int64_t>(0ULL)]),(cast<float>(1.00000000000000000e+00) - src[cast<int64_t>(1ULL)]),(cast<float>(1.00000000000000000e+00) - src[cast<int64_t>(2ULL)]),(cast<float>(1.00000000000000000e+00) - src[cast<int64_t>(3ULL)])};
break;}
case cast<uint32_t>(770ULL):{
return std::array<float,4>{src[cast<int64_t>(3ULL)],src[cast<int64_t>(3ULL)],src[cast<int64_t>(3ULL)],src[cast<int64_t>(3ULL)]};
break;}
case cast<uint32_t>(771ULL):{
return std::array<float,4>{(cast<float>(1.00000000000000000e+00) - src[cast<int64_t>(3ULL)]),(cast<float>(1.00000000000000000e+00) - src[cast<int64_t>(3ULL)]),(cast<float>(1.00000000000000000e+00) - src[cast<int64_t>(3ULL)]),(cast<float>(1.00000000000000000e+00) - src[cast<int64_t>(3ULL)])};
break;}
case cast<uint32_t>(772ULL):{
return std::array<float,4>{dst[cast<int64_t>(3ULL)],dst[cast<int64_t>(3ULL)],dst[cast<int64_t>(3ULL)],dst[cast<int64_t>(3ULL)]};
break;}
case cast<uint32_t>(773ULL):{
return std::array<float,4>{(cast<float>(1.00000000000000000e+00) - dst[cast<int64_t>(3ULL)]),(cast<float>(1.00000000000000000e+00) - dst[cast<int64_t>(3ULL)]),(cast<float>(1.00000000000000000e+00) - dst[cast<int64_t>(3ULL)]),(cast<float>(1.00000000000000000e+00) - dst[cast<int64_t>(3ULL)])};
break;}
case cast<uint32_t>(774ULL):{
return dst;
break;}
case cast<uint32_t>(775ULL):{
return std::array<float,4>{(cast<float>(1.00000000000000000e+00) - dst[cast<int64_t>(0ULL)]),(cast<float>(1.00000000000000000e+00) - dst[cast<int64_t>(1ULL)]),(cast<float>(1.00000000000000000e+00) - dst[cast<int64_t>(2ULL)]),(cast<float>(1.00000000000000000e+00) - dst[cast<int64_t>(3ULL)])};
break;}
case cast<uint32_t>(776ULL):{
float f = xbox_minf32(src[cast<int64_t>(3ULL)],(cast<float>(1.00000000000000000e+00) - dst[cast<int64_t>(3ULL)]));
return std::array<float,4>{f,f,f,cast<float>(1.00000000000000000e+00)};
break;}
case cast<uint32_t>(32769ULL):{
return st->blendConst;
break;}
case cast<uint32_t>(32770ULL):{
return std::array<float,4>{(cast<float>(1.00000000000000000e+00) - st->blendConst[cast<int64_t>(0ULL)]),(cast<float>(1.00000000000000000e+00) - st->blendConst[cast<int64_t>(1ULL)]),(cast<float>(1.00000000000000000e+00) - st->blendConst[cast<int64_t>(2ULL)]),(cast<float>(1.00000000000000000e+00) - st->blendConst[cast<int64_t>(3ULL)])};
break;}
case cast<uint32_t>(32771ULL):{
float a = st->blendConst[cast<int64_t>(3ULL)];
return std::array<float,4>{a,a,a,a};
break;}
case cast<uint32_t>(32772ULL):{
float a = (cast<float>(1.00000000000000000e+00) - st->blendConst[cast<int64_t>(3ULL)]);
return std::array<float,4>{a,a,a,a};
break;}
}}
return std::array<float,4>{cast<float>(1.00000000000000000e+00),cast<float>(1.00000000000000000e+00),cast<float>(1.00000000000000000e+00),cast<float>(1.00000000000000000e+00)};
}
;
auto tmp255 = std::make_tuple(factor(st->blendSrc,false),factor(st->blendDst,true));
std::array<float,4> sf = std::get<0>(tmp255);
std::array<float,4> df = std::get<1>(tmp255);
std::array<float,4> out={};
{
switch(st->blendEq){
case cast<uint32_t>(32778ULL):{
{auto&& tmp256 = out;
for(int64_t tmp257=0;tmp257<len(tmp256);++tmp257){
auto i=tmp257;out[i] = ((src[i] * sf[i]) - (dst[i] * df[i]));
}}
break;}
case cast<uint32_t>(32779ULL):{
{auto&& tmp258 = out;
for(int64_t tmp259=0;tmp259<len(tmp258);++tmp259){
auto i=tmp259;out[i] = ((dst[i] * df[i]) - (src[i] * sf[i]));
}}
break;}
case cast<uint32_t>(32775ULL):{
{auto&& tmp260 = out;
for(int64_t tmp261=0;tmp261<len(tmp260);++tmp261){
auto i=tmp261;out[i] = xbox_minf32(src[i],dst[i]);
}}
break;}
case cast<uint32_t>(32776ULL):{
{auto&& tmp262 = out;
for(int64_t tmp263=0;tmp263<len(tmp262);++tmp263){
auto i=tmp263;out[i] = xbox_maxf32(src[i],dst[i]);
}}
break;}
default:{
{auto&& tmp264 = out;
for(int64_t tmp265=0;tmp265<len(tmp264);++tmp265){
auto i=tmp265;out[i] = ((src[i] * sf[i]) + (dst[i] * df[i]));
}}
break;}
}}
return out;
}
}
// tools/platform/xbox/nv2a_raster.go:843:1
uint8_t xbox_u8(float f){
{
if ((f <= cast<float>(0.00000000000000000e+00))) {
return cast<uint8_t>(0ULL);
}
if ((f >= cast<float>(1.00000000000000000e+00))) {
return cast<uint8_t>(255ULL);
}
return cast<uint8_t>(((f * cast<float>(2.55000000000000000e+02)) + cast<float>(5.00000000000000000e-01)));
}
}
// tools/platform/xbox/nv2a_raster.go:853:1
float xbox_invW(float w){
{
if ((w == cast<float>(0.00000000000000000e+00))) {
return cast<float>(1.00000000000000000e+00);
}
return (cast<float>(1.00000000000000000e+00) / w);
}
}
// tools/platform/xbox/nv2a_raster.go:860:1
double xbox_min3(double a,double b,double c){
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
// tools/platform/xbox/nv2a_raster.go:869:1
double xbox_max3(double a,double b,double c){
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
// tools/platform/xbox/nv2a_raster.go:893:1
std::array<float,4> xbox_pgraph_texReflectSpecular(xbox_pgraph* g,xbox_rasterState* st,int64_t u,std::array<std::array<float,4>,4>* uvw,xbox_combInput* in){
{
if ((((u < cast<int64_t>(2ULL)) || (st->texStage[cast<int64_t>((u - cast<int64_t>(1ULL)))] != cast<uint32_t>(17ULL))) || (st->texStage[cast<int64_t>((u - cast<int64_t>(2ULL)))] != cast<uint32_t>(17ULL)))) {
x86_CPU_Halt(g->m->CPU,std::string("nv2a: DOT_REFLECT_SPECULAR at unit %d without two preceding DOT_PRODUCT stages (saw %d,%d)",90),u,st->texStage[cast<int64_t>((u - cast<int64_t>(2ULL)))],st->texStage[cast<int64_t>((u - cast<int64_t>(1ULL)))]);
return std::array<float,4>{cast<float>(0.00000000000000000e+00),cast<float>(0.00000000000000000e+00),cast<float>(0.00000000000000000e+00),cast<float>(1.00000000000000000e+00)};
}
std::array<float,3> n = std::array<float,3>{cast<float>(0.00000000000000000e+00),cast<float>(0.00000000000000000e+00),cast<float>(1.00000000000000000e+00)};
{int64_t s = cast<int64_t>((u - cast<int64_t>(1ULL)));for (;(s >= cast<int64_t>(0ULL));s--){
if ((((st->texEnable[s] && bool(st->texImg[s])) && (!st->texImg[s]->cube)) && (!st->texImg[s]->depth))) {
n = std::array<float,3>{((in->tex[s][cast<int64_t>(0ULL)] * cast<float>(2.00000000000000000e+00)) - cast<float>(1.00000000000000000e+00)),((in->tex[s][cast<int64_t>(1ULL)] * cast<float>(2.00000000000000000e+00)) - cast<float>(1.00000000000000000e+00)),((in->tex[s][cast<int64_t>(2ULL)] * cast<float>(2.00000000000000000e+00)) - cast<float>(1.00000000000000000e+00))};
break;
}
}
}auto dot = [&](std::array<float,4> tc)->float{
return (((tc[cast<int64_t>(0ULL)] * n[cast<int64_t>(0ULL)]) + (tc[cast<int64_t>(1ULL)] * n[cast<int64_t>(1ULL)])) + (tc[cast<int64_t>(2ULL)] * n[cast<int64_t>(2ULL)]));
}
;
std::array<float,3> N = std::array<float,3>{dot((*uvw)[cast<int64_t>((u - cast<int64_t>(2ULL)))]),dot((*uvw)[cast<int64_t>((u - cast<int64_t>(1ULL)))]),dot((*uvw)[u])};
std::array<float,3> E = std::array<float,3>{(*uvw)[cast<int64_t>((u - cast<int64_t>(2ULL)))][cast<int64_t>(3ULL)],(*uvw)[cast<int64_t>((u - cast<int64_t>(1ULL)))][cast<int64_t>(3ULL)],(*uvw)[u][cast<int64_t>(3ULL)]};
std::array<float,3> R = E;
{
float nn = (((N[cast<int64_t>(0ULL)] * N[cast<int64_t>(0ULL)]) + (N[cast<int64_t>(1ULL)] * N[cast<int64_t>(1ULL)])) + (N[cast<int64_t>(2ULL)] * N[cast<int64_t>(2ULL)]));
if ((nn != cast<float>(0.00000000000000000e+00))) {
float k = ((cast<float>(2.00000000000000000e+00) * ((((N[cast<int64_t>(0ULL)] * E[cast<int64_t>(0ULL)]) + (N[cast<int64_t>(1ULL)] * E[cast<int64_t>(1ULL)])) + (N[cast<int64_t>(2ULL)] * E[cast<int64_t>(2ULL)])))) / nn);
R = std::array<float,3>{((k * N[cast<int64_t>(0ULL)]) - E[cast<int64_t>(0ULL)]),((k * N[cast<int64_t>(1ULL)]) - E[cast<int64_t>(1ULL)]),((k * N[cast<int64_t>(2ULL)]) - E[cast<int64_t>(2ULL)])};
}
}
return xbox_pgraph_texSampleCube(g,st,u,R[cast<int64_t>(0ULL)],R[cast<int64_t>(1ULL)],R[cast<int64_t>(2ULL)]);
}
}
// tools/platform/xbox/nv2a_texture.go:84:1
uint64_t xbox_hashRAM(Slice<uint8_t> ram,uint32_t phys,uint32_t n){
{
uint64_t end = cast<uint64_t>((cast<uint64_t>(phys) + cast<uint64_t>(n)));
if ((end > cast<uint64_t>(len(ram)))) {
end = cast<uint64_t>(len(ram));
}
uint64_t h = cast<uint64_t>(1469598103934665603ULL);
uint64_t i = cast<uint64_t>(phys);
{;for (;(cast<uint64_t>((i + cast<uint64_t>(8ULL))) <= end);i += cast<uint64_t>(8ULL)){
h = cast<uint64_t>(((cast<uint64_t>((h ^ le_Uint64(sub(ram,i,len(ram)))))) * cast<uint64_t>(1099511628211ULL)));
}
}{;for (;(i < end);i++){
h = cast<uint64_t>(((cast<uint64_t>((h ^ cast<uint64_t>(ram[i])))) * cast<uint64_t>(1099511628211ULL)));
}
}return h;
}
}
// tools/platform/xbox/nv2a_texture.go:105:1
std::tuple<uint32_t,bool> xbox_texSpan(uint32_t colorFmt,bool cube,int64_t w,int64_t h,uint32_t pitch){
{
uint32_t faceBytes={};
{
switch(colorFmt){
case cast<uint32_t>(12ULL):{
faceBytes = cast<uint32_t>((cast<uint32_t>(cast<int64_t>((divi<int64_t>((cast<int64_t>((w + cast<int64_t>(3ULL)))),cast<int64_t>(4ULL)) * (divi<int64_t>((cast<int64_t>((h + cast<int64_t>(3ULL)))),cast<int64_t>(4ULL)))))) * cast<uint32_t>(8ULL)));
break;}
case cast<uint32_t>(14ULL):case cast<uint32_t>(15ULL):{
faceBytes = cast<uint32_t>((cast<uint32_t>(cast<int64_t>((divi<int64_t>((cast<int64_t>((w + cast<int64_t>(3ULL)))),cast<int64_t>(4ULL)) * (divi<int64_t>((cast<int64_t>((h + cast<int64_t>(3ULL)))),cast<int64_t>(4ULL)))))) * cast<uint32_t>(16ULL)));
break;}
case cast<uint32_t>(6ULL):case cast<uint32_t>(7ULL):{
faceBytes = cast<uint32_t>(cast<int64_t>((cast<int64_t>((w * h)) * cast<int64_t>(4ULL))));
break;}
case cast<uint32_t>(5ULL):case cast<uint32_t>(2ULL):case cast<uint32_t>(4ULL):{
faceBytes = cast<uint32_t>(cast<int64_t>((cast<int64_t>((w * h)) * cast<int64_t>(2ULL))));
break;}
case cast<uint32_t>(25ULL):{
faceBytes = cast<uint32_t>(cast<int64_t>((w * h)));
break;}
case cast<uint32_t>(18ULL):case cast<uint32_t>(30ULL):{
if ((pitch == cast<uint32_t>(0ULL))) {
pitch = cast<uint32_t>(cast<int64_t>((w * cast<int64_t>(4ULL))));
}
faceBytes = cast<uint32_t>((cast<uint32_t>(h) * pitch));
break;}
case cast<uint32_t>(17ULL):{
if ((pitch == cast<uint32_t>(0ULL))) {
pitch = cast<uint32_t>(cast<int64_t>((w * cast<int64_t>(2ULL))));
}
faceBytes = cast<uint32_t>((cast<uint32_t>(h) * pitch));
break;}
case cast<uint32_t>(46ULL):{
if ((pitch == cast<uint32_t>(0ULL))) {
pitch = cast<uint32_t>(cast<int64_t>((w * cast<int64_t>(4ULL))));
}
faceBytes = cast<uint32_t>((cast<uint32_t>(h) * pitch));
break;}
default:{
return {cast<uint32_t>(0ULL),false};
break;}
}}
if (cube) {
return {cast<uint32_t>((faceBytes * cast<uint32_t>(6ULL))),true};
}
return {faceBytes,true};
}
}
// tools/platform/xbox/nv2a_texture.go:149:1
std::tuple<uint32_t,uint32_t,bool> xbox_pgraph_texSource(xbox_pgraph* g,int64_t u){
uint32_t phys{};
uint32_t span{};
bool ok{};
{
uint32_t r = cast<uint32_t>((cast<uint32_t>(u) * cast<uint32_t>(64ULL)));
uint32_t offset = g->Regs[shr<uint32_t>((cast<uint32_t>((cast<uint32_t>(6912ULL) + r))),cast<int64_t>(2ULL))];
uint32_t format = g->Regs[shr<uint32_t>((cast<uint32_t>((cast<uint32_t>(6916ULL) + r))),cast<int64_t>(2ULL))];
uint32_t rect = g->Regs[shr<uint32_t>((cast<uint32_t>((cast<uint32_t>(6940ULL) + r))),cast<int64_t>(2ULL))];
uint32_t ctl1 = g->Regs[shr<uint32_t>((cast<uint32_t>((cast<uint32_t>(6928ULL) + r))),cast<int64_t>(2ULL))];
if ((cast<uint32_t>((shr<uint32_t>(format,cast<int64_t>(4ULL)) & cast<uint32_t>(15ULL))) != cast<uint32_t>(2ULL))) {
return {cast<uint32_t>(0ULL),cast<uint32_t>(0ULL),false};
}
uint32_t colorFmt = cast<uint32_t>((shr<uint32_t>(format,cast<int64_t>(8ULL)) & cast<uint32_t>(255ULL)));
uint32_t dmaReg = cast<uint32_t>(388ULL);
if ((cast<uint32_t>((format & cast<uint32_t>(3ULL))) == cast<uint32_t>(2ULL))) {
dmaReg = cast<uint32_t>(392ULL);
}
auto tmp266 = xbox_Machine_dmaObjectTarget(g->m,g->Regs[shr<uint32_t>(dmaReg,cast<int64_t>(2ULL))]);
uint32_t base = std::get<0>(tmp266);
auto tmp267 = xbox_Machine_translate(g->m,cast<uint32_t>((base + offset)));
uint32_t p = std::get<0>(tmp267);
bool mmio = std::get<1>(tmp267);
bool okT = std::get<2>(tmp267);
if (((!okT) || mmio)) {
return {cast<uint32_t>(0ULL),cast<uint32_t>(0ULL),false};
}
int64_t w={};
int64_t h={};
if (xbox_isLinearTexFmt(colorFmt)) {
auto tmp268 = std::make_tuple(cast<int64_t>(shr<uint32_t>(rect,cast<int64_t>(16ULL))),cast<int64_t>(cast<uint32_t>((rect & cast<uint32_t>(65535ULL)))));
w = std::get<0>(tmp268);
h = std::get<1>(tmp268);
}
else {
auto tmp269 = std::make_tuple(shl<int64_t>(cast<int64_t>(1ULL),(cast<uint32_t>((shr<uint32_t>(format,cast<int64_t>(20ULL)) & cast<uint32_t>(15ULL))))),shl<int64_t>(cast<int64_t>(1ULL),(cast<uint32_t>((shr<uint32_t>(format,cast<int64_t>(24ULL)) & cast<uint32_t>(15ULL))))));
w = std::get<0>(tmp269);
h = std::get<1>(tmp269);
}
if (((((w <= cast<int64_t>(0ULL)) || (h <= cast<int64_t>(0ULL))) || (w > cast<int64_t>(4096ULL))) || (h > cast<int64_t>(4096ULL)))) {
return {cast<uint32_t>(0ULL),cast<uint32_t>(0ULL),false};
}
auto tmp270 = xbox_texSpan(colorFmt,(cast<uint32_t>((shr<uint32_t>(format,cast<int64_t>(2ULL)) & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL)),w,h,shr<uint32_t>(ctl1,cast<int64_t>(16ULL)));
span = std::get<0>(tmp270);
ok = std::get<1>(tmp270);
if (((!ok) || (cast<uint64_t>((cast<uint64_t>(p) + cast<uint64_t>(span))) > cast<uint64_t>(len(g->m->RAM))))) {
return {cast<uint32_t>(0ULL),cast<uint32_t>(0ULL),false};
}
return {p,span,true};
}
}
// tools/platform/xbox/nv2a_texture.go:186:1
void xbox_pgraph_cacheTex(xbox_pgraph* g,xbox_texKey key,xbox_texImage* img,int64_t u){
{
uint64_t hsh={};
auto tmp271 = xbox_pgraph_texSource(g,u);
uint32_t phys = std::get<0>(tmp271);
uint32_t span = std::get<1>(tmp271);
bool ok = std::get<2>(tmp271);
if (ok) {
hsh = xbox_hashRAM(g->m->RAM,phys,span);
}
g->texCache[key] = rrTextureEntryNew(xbox_texEntry{img,hsh,span,g->texRun});
}
}
// tools/platform/xbox/nv2a_texture.go:210:1
bool xbox_pgraph_texStateDecode(xbox_pgraph* g,xbox_rasterState* st){
{
uint32_t shader = g->Regs[cast<int64_t>(1948ULL)];
{int64_t u = cast<int64_t>(0ULL);for (;(u < cast<int64_t>(4ULL));u++){
uint32_t r = cast<uint32_t>((cast<uint32_t>(u) * cast<uint32_t>(64ULL)));
uint32_t ctl = g->Regs[shr<uint32_t>((cast<uint32_t>((cast<uint32_t>(6924ULL) + r))),cast<int64_t>(2ULL))];
st->texEnable[u] = (cast<uint32_t>((shr<uint32_t>(ctl,cast<int64_t>(30ULL)) & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL));
st->texStage[u] = cast<uint32_t>((shr<uint32_t>(shader,(cast<uint64_t>((cast<uint64_t>(5ULL) * cast<uint64_t>(u))))) & cast<uint32_t>(31ULL)));
if (((!st->texEnable[u]) || (st->texStage[u] == cast<uint32_t>(0ULL)))) {
st->texEnable[u] = false;
st->texImg[u] = {};
continue;
}
uint32_t addr = g->Regs[shr<uint32_t>((cast<uint32_t>((cast<uint32_t>(6920ULL) + r))),cast<int64_t>(2ULL))];
st->texWrapU[u] = cast<uint32_t>((addr & cast<uint32_t>(255ULL)));
st->texWrapV[u] = cast<uint32_t>((shr<uint32_t>(addr,cast<int64_t>(8ULL)) & cast<uint32_t>(255ULL)));
uint32_t filt = g->Regs[shr<uint32_t>((cast<uint32_t>((cast<uint32_t>(6932ULL) + r))),cast<int64_t>(2ULL))];
st->texBilinear[u] = (cast<uint32_t>((shr<uint32_t>(filt,cast<int64_t>(24ULL)) & cast<uint32_t>(255ULL))) >= cast<uint32_t>(2ULL));
auto tmp272 = xbox_pgraph_texDecode(g,u);
xbox_texImage* img = std::get<0>(tmp272);
bool rect = std::get<1>(tmp272);
bool ok = std::get<2>(tmp272);
if ((!ok)) {
return false;
}
st->texImg[u] = img;
st->texRect[u] = rect;
}
}return true;
}
}
// tools/platform/xbox/nv2a_texture.go:240:1
std::tuple<xbox_texImage*,bool,bool> xbox_pgraph_texDecode(xbox_pgraph* g,int64_t u){
{rrprof::Scope timing(5,"Texture decoding");
{
uint32_t r = cast<uint32_t>((cast<uint32_t>(u) * cast<uint32_t>(64ULL)));
uint32_t offset = g->Regs[shr<uint32_t>((cast<uint32_t>((cast<uint32_t>(6912ULL) + r))),cast<int64_t>(2ULL))];
uint32_t format = g->Regs[shr<uint32_t>((cast<uint32_t>((cast<uint32_t>(6916ULL) + r))),cast<int64_t>(2ULL))];
uint32_t rect = g->Regs[shr<uint32_t>((cast<uint32_t>((cast<uint32_t>(6940ULL) + r))),cast<int64_t>(2ULL))];
uint32_t ctl1 = g->Regs[shr<uint32_t>((cast<uint32_t>((cast<uint32_t>(6928ULL) + r))),cast<int64_t>(2ULL))];
xbox_texKey key = xbox_texKey{offset,format,rect,ctl1};
bool linear = xbox_isLinearTexFmt(cast<uint32_t>((shr<uint32_t>(format,cast<int64_t>(8ULL)) & cast<uint32_t>(255ULL))));
{
auto tmp273 = lookup(g->texCache,key);
xbox_texEntry* e = std::get<0>(tmp273);
bool ok = std::get<1>(tmp273);
if (ok) {
if ((e->validated == g->texRun)) {
return {e->img,linear,true};
}
{
auto tmp274 = xbox_pgraph_texSource(g,u);
uint32_t phys = std::get<0>(tmp274);
uint32_t span = std::get<1>(tmp274);
bool ok2 = std::get<2>(tmp274);
if ((ok2 && (xbox_hashRAM(g->m->RAM,phys,span) == e->srcHash))) {
e->validated = g->texRun;
return {e->img,linear,true};
}
}
}
}
{
uint32_t dim = cast<uint32_t>((shr<uint32_t>(format,cast<int64_t>(4ULL)) & cast<uint32_t>(15ULL)));
if ((dim != cast<uint32_t>(2ULL))) {
x86_CPU_Halt(g->m->CPU,std::string("nv2a: texture unit %d dimensionality %d unmodelled (fmt=%08X)",61),u,dim,format);
return {{},false,false};
}
}
uint32_t dmaReg = cast<uint32_t>(388ULL);
if ((cast<uint32_t>((format & cast<uint32_t>(3ULL))) == cast<uint32_t>(2ULL))) {
dmaReg = cast<uint32_t>(392ULL);
}
auto tmp275 = xbox_Machine_dmaObjectTarget(g->m,g->Regs[shr<uint32_t>(dmaReg,cast<int64_t>(2ULL))]);
uint32_t base = std::get<0>(tmp275);
auto tmp276 = xbox_Machine_translate(g->m,cast<uint32_t>((base + offset)));
uint32_t phys = std::get<0>(tmp276);
bool mmio = std::get<1>(tmp276);
bool ok = std::get<2>(tmp276);
if (((!ok) || mmio)) {
x86_CPU_Halt(g->m->CPU,std::string("nv2a: texture unit %d at %08X is not RAM",40),u,cast<uint32_t>((base + offset)));
return {{},false,false};
}
uint32_t colorFmt = cast<uint32_t>((shr<uint32_t>(format,cast<int64_t>(8ULL)) & cast<uint32_t>(255ULL)));
int64_t w={};
int64_t h={};
if (linear) {
auto tmp277 = std::make_tuple(cast<int64_t>(shr<uint32_t>(rect,cast<int64_t>(16ULL))),cast<int64_t>(cast<uint32_t>((rect & cast<uint32_t>(65535ULL)))));
w = std::get<0>(tmp277);
h = std::get<1>(tmp277);
}
else {
auto tmp278 = std::make_tuple(shl<int64_t>(cast<int64_t>(1ULL),(cast<uint32_t>((shr<uint32_t>(format,cast<int64_t>(20ULL)) & cast<uint32_t>(15ULL))))),shl<int64_t>(cast<int64_t>(1ULL),(cast<uint32_t>((shr<uint32_t>(format,cast<int64_t>(24ULL)) & cast<uint32_t>(15ULL))))));
w = std::get<0>(tmp278);
h = std::get<1>(tmp278);
}
if (((((w <= cast<int64_t>(0ULL)) || (h <= cast<int64_t>(0ULL))) || (w > cast<int64_t>(4096ULL))) || (h > cast<int64_t>(4096ULL)))) {
x86_CPU_Halt(g->m->CPU,std::string("nv2a: texture unit %d has impossible size %dx%d (fmt=%08X rect=%08X)",68),u,w,h,format,rect);
return {{},false,false};
}
Slice<uint8_t> ram = g->m->RAM;
if ((cast<uint32_t>((shr<uint32_t>(format,cast<int64_t>(2ULL)) & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL))) {
if ((w != h)) {
x86_CPU_Halt(g->m->CPU,std::string("nv2a: texture unit %d cube map %dx%d is not square (fmt=%08X)",61),u,w,h,format);
return {{},false,false};
}
{
uint32_t mips = cast<uint32_t>((shr<uint32_t>(format,cast<int64_t>(16ULL)) & cast<uint32_t>(15ULL)));
if ((mips > cast<uint32_t>(1ULL))) {
x86_CPU_Halt(g->m->CPU,std::string("nv2a: texture unit %d cube map with %d mip levels \342\200\224 face stride unmodelled (fmt=%08X)",87),u,mips,format);
return {{},false,false};
}
}
uint32_t faceBytes={};
std::function<void(xbox_texImage*,uint32_t)> decodeFace={};
{
switch(colorFmt){
case cast<uint32_t>(12ULL):case cast<uint32_t>(14ULL):case cast<uint32_t>(15ULL):{
int64_t variant = get(Map<uint32_t,int64_t>{{cast<uint32_t>(12ULL),cast<int64_t>(1ULL)},{cast<uint32_t>(14ULL),cast<int64_t>(3ULL)},{cast<uint32_t>(15ULL),cast<int64_t>(5ULL)}},colorFmt);
uint32_t blockBytes = cast<uint32_t>(8ULL);
if ((variant != cast<int64_t>(1ULL))) {
blockBytes = cast<uint32_t>(16ULL);
}
faceBytes = cast<uint32_t>((cast<uint32_t>(cast<int64_t>((divi<int64_t>((cast<int64_t>((w + cast<int64_t>(3ULL)))),cast<int64_t>(4ULL)) * (divi<int64_t>((cast<int64_t>((h + cast<int64_t>(3ULL)))),cast<int64_t>(4ULL)))))) * blockBytes));
decodeFace = [=](xbox_texImage* dst,uint32_t addr)->void{
xbox_decodeDXT(dst,ram,addr,variant);
}
;
break;}
case cast<uint32_t>(6ULL):case cast<uint32_t>(7ULL):{
faceBytes = cast<uint32_t>(cast<int64_t>((cast<int64_t>((w * h)) * cast<int64_t>(4ULL))));
decodeFace = [=](xbox_texImage* dst,uint32_t addr)->void{
xbox_decodeSwizzled(dst,ram,addr,cast<int64_t>(4ULL),[=](Slice<uint8_t> pix,int64_t o,Slice<uint8_t> b)->void{
auto tmp279 = std::make_tuple(b[cast<int64_t>(2ULL)],b[cast<int64_t>(1ULL)],b[cast<int64_t>(0ULL)],b[cast<int64_t>(3ULL)]);
pix[o] = std::get<0>(tmp279);
pix[cast<int64_t>((o + cast<int64_t>(1ULL)))] = std::get<1>(tmp279);
pix[cast<int64_t>((o + cast<int64_t>(2ULL)))] = std::get<2>(tmp279);
pix[cast<int64_t>((o + cast<int64_t>(3ULL)))] = std::get<3>(tmp279);
if ((colorFmt == cast<uint32_t>(7ULL))) {
pix[cast<int64_t>((o + cast<int64_t>(3ULL)))] = cast<uint8_t>(255ULL);
}
}
);
}
;
break;}
default:{
x86_CPU_Halt(g->m->CPU,std::string("nv2a: texture unit %d cube map color format 0x%02X unmodelled (fmt=%08X)",72),u,colorFmt,format);
return {{},false,false};
break;}
}}
xbox_texImage* img = rrTextureNew(xbox_texImage{w,h,true,Slice<uint8_t>::make(cast<int64_t>((cast<int64_t>((cast<int64_t>((w * h)) * cast<int64_t>(4ULL))) * cast<int64_t>(6ULL)))),{}});
xbox_texImage* face = rrTextureNew(xbox_texImage{w,h,{},Slice<uint8_t>::make(cast<int64_t>((cast<int64_t>((w * h)) * cast<int64_t>(4ULL)))),{}});
{int64_t f = cast<int64_t>(0ULL);for (;(f < cast<int64_t>(6ULL));f++){
decodeFace(face,cast<uint32_t>((phys + cast<uint32_t>((cast<uint32_t>(f) * faceBytes)))));
gcopy(sub(img->pix,cast<int64_t>((cast<int64_t>((cast<int64_t>((f * w)) * h)) * cast<int64_t>(4ULL))),len(img->pix)),face->pix);
}
}xbox_pgraph_cacheTex(g,key,img,u);
return {img,linear,true};
}
xbox_texImage* img = rrTextureNew(xbox_texImage{w,h,{},Slice<uint8_t>::make(cast<int64_t>((cast<int64_t>((w * h)) * cast<int64_t>(4ULL)))),{}});
{
switch(colorFmt){
case cast<uint32_t>(12ULL):{
xbox_decodeDXT(img,ram,phys,cast<int64_t>(1ULL));
break;}
case cast<uint32_t>(14ULL):{
xbox_decodeDXT(img,ram,phys,cast<int64_t>(3ULL));
break;}
case cast<uint32_t>(15ULL):{
xbox_decodeDXT(img,ram,phys,cast<int64_t>(5ULL));
break;}
case cast<uint32_t>(6ULL):case cast<uint32_t>(7ULL):{
xbox_decodeSwizzled(img,ram,phys,cast<int64_t>(4ULL),[=](Slice<uint8_t> pix,int64_t o,Slice<uint8_t> b)->void{
auto tmp280 = std::make_tuple(b[cast<int64_t>(2ULL)],b[cast<int64_t>(1ULL)],b[cast<int64_t>(0ULL)],b[cast<int64_t>(3ULL)]);
pix[o] = std::get<0>(tmp280);
pix[cast<int64_t>((o + cast<int64_t>(1ULL)))] = std::get<1>(tmp280);
pix[cast<int64_t>((o + cast<int64_t>(2ULL)))] = std::get<2>(tmp280);
pix[cast<int64_t>((o + cast<int64_t>(3ULL)))] = std::get<3>(tmp280);
if ((colorFmt == cast<uint32_t>(7ULL))) {
pix[cast<int64_t>((o + cast<int64_t>(3ULL)))] = cast<uint8_t>(255ULL);
}
}
);
break;}
case cast<uint32_t>(5ULL):{
xbox_decodeSwizzled(img,ram,phys,cast<int64_t>(2ULL),xbox_decode565);
break;}
case cast<uint32_t>(2ULL):{
xbox_decodeSwizzled(img,ram,phys,cast<int64_t>(2ULL),xbox_decode1555);
break;}
case cast<uint32_t>(4ULL):{
xbox_decodeSwizzled(img,ram,phys,cast<int64_t>(2ULL),xbox_decode4444);
break;}
case cast<uint32_t>(25ULL):{
xbox_decodeSwizzled(img,ram,phys,cast<int64_t>(1ULL),[=](Slice<uint8_t> pix,int64_t o,Slice<uint8_t> b)->void{
auto tmp281 = std::make_tuple(cast<uint8_t>(255ULL),cast<uint8_t>(255ULL),cast<uint8_t>(255ULL),b[cast<int64_t>(0ULL)]);
pix[o] = std::get<0>(tmp281);
pix[cast<int64_t>((o + cast<int64_t>(1ULL)))] = std::get<1>(tmp281);
pix[cast<int64_t>((o + cast<int64_t>(2ULL)))] = std::get<2>(tmp281);
pix[cast<int64_t>((o + cast<int64_t>(3ULL)))] = std::get<3>(tmp281);
}
);
break;}
case cast<uint32_t>(18ULL):case cast<uint32_t>(30ULL):{
xbox_decodeLinear(img,ram,phys,shr<uint32_t>(ctl1,cast<int64_t>(16ULL)),cast<int64_t>(4ULL),[=](Slice<uint8_t> pix,int64_t o,Slice<uint8_t> b)->void{
auto tmp282 = std::make_tuple(b[cast<int64_t>(2ULL)],b[cast<int64_t>(1ULL)],b[cast<int64_t>(0ULL)],b[cast<int64_t>(3ULL)]);
pix[o] = std::get<0>(tmp282);
pix[cast<int64_t>((o + cast<int64_t>(1ULL)))] = std::get<1>(tmp282);
pix[cast<int64_t>((o + cast<int64_t>(2ULL)))] = std::get<2>(tmp282);
pix[cast<int64_t>((o + cast<int64_t>(3ULL)))] = std::get<3>(tmp282);
if ((colorFmt == cast<uint32_t>(30ULL))) {
pix[cast<int64_t>((o + cast<int64_t>(3ULL)))] = cast<uint8_t>(255ULL);
}
}
);
break;}
case cast<uint32_t>(17ULL):{
xbox_decodeLinear(img,ram,phys,shr<uint32_t>(ctl1,cast<int64_t>(16ULL)),cast<int64_t>(2ULL),xbox_decode565);
break;}
case cast<uint32_t>(46ULL):{
if (xbox_shadowTrace) {
xbox_pgraph_DumpZetaHist(g);
if ((!g->shadowDumped)) {
g->shadowDumped = true;
xbox_pgraph_dumpReceiverState(g);
}
}
uint32_t pitch = shr<uint32_t>(ctl1,cast<int64_t>(16ULL));
if ((pitch == cast<uint32_t>(0ULL))) {
pitch = cast<uint32_t>(cast<int64_t>((w * cast<int64_t>(4ULL))));
}
img->depth = Slice<uint32_t>::make(cast<int64_t>((w * h)));
{int64_t y = cast<int64_t>(0ULL);for (;(y < h);y++){
uint32_t row = cast<uint32_t>((phys + cast<uint32_t>((cast<uint32_t>(y) * pitch))));
{int64_t x = cast<int64_t>(0ULL);for (;(x < w);x++){
uint32_t a = cast<uint32_t>((row + cast<uint32_t>(cast<int64_t>((x * cast<int64_t>(4ULL))))));
if ((cast<int64_t>((cast<int64_t>(a) + cast<int64_t>(4ULL))) > len(ram))) {
continue;
}
uint32_t v = cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(ram[a]) | shl<uint32_t>(cast<uint32_t>(ram[cast<uint32_t>((a + cast<uint32_t>(1ULL)))]),cast<int64_t>(8ULL)))) | shl<uint32_t>(cast<uint32_t>(ram[cast<uint32_t>((a + cast<uint32_t>(2ULL)))]),cast<int64_t>(16ULL)))) | shl<uint32_t>(cast<uint32_t>(ram[cast<uint32_t>((a + cast<uint32_t>(3ULL)))]),cast<int64_t>(24ULL))));
uint32_t d = shr<uint32_t>(v,cast<int64_t>(8ULL));
img->depth[cast<int64_t>((cast<int64_t>((y * w)) + x))] = d;
int64_t o = cast<int64_t>(((cast<int64_t>((cast<int64_t>((y * w)) + x))) * cast<int64_t>(4ULL)));
auto tmp283 = std::make_tuple(cast<uint8_t>(shr<uint32_t>(d,cast<int64_t>(16ULL))),cast<uint8_t>(shr<uint32_t>(d,cast<int64_t>(16ULL))),cast<uint8_t>(shr<uint32_t>(d,cast<int64_t>(16ULL))),cast<uint8_t>(255ULL));
img->pix[o] = std::get<0>(tmp283);
img->pix[cast<int64_t>((o + cast<int64_t>(1ULL)))] = std::get<1>(tmp283);
img->pix[cast<int64_t>((o + cast<int64_t>(2ULL)))] = std::get<2>(tmp283);
img->pix[cast<int64_t>((o + cast<int64_t>(3ULL)))] = std::get<3>(tmp283);
}
}}
}break;}
default:{
x86_CPU_Halt(g->m->CPU,std::string("nv2a: texture unit %d color format 0x%02X unmodelled (fmt=%08X)",63),u,colorFmt,format);
return {{},false,false};
break;}
}}
xbox_pgraph_cacheTex(g,key,img,u);
return {img,linear,true};
}
}
}
// tools/platform/xbox/nv2a_texture.go:434:1
std::tuple<int64_t,int64_t,Slice<uint8_t>,bool> xbox_pgraph_DebugDecodeTexture(xbox_pgraph* g,int64_t u){
int64_t w{};
int64_t h{};
Slice<uint8_t> pix{};
bool ok{};
{
auto tmp284 = xbox_pgraph_texDecode(g,u);
xbox_texImage* img = std::get<0>(tmp284);
ok = std::get<2>(tmp284);
if (((!ok) || (!img))) {
return {cast<int64_t>(0ULL),cast<int64_t>(0ULL),{},false};
}
return {img->w,img->h,img->pix,true};
}
}
// tools/platform/xbox/nv2a_texture.go:442:1
bool xbox_isLinearTexFmt(uint32_t colorFmt){
{
{
switch(colorFmt){
case cast<uint32_t>(17ULL):case cast<uint32_t>(18ULL):case cast<uint32_t>(30ULL):case cast<uint32_t>(46ULL):{
return true;
break;}
}}
return false;
}
}
// tools/platform/xbox/nv2a_texture.go:454:1
uint8_t xbox_exp5(uint16_t v){
{
return cast<uint8_t>(cast<uint16_t>((shl<uint16_t>(v,cast<int64_t>(3ULL)) | shr<uint16_t>(v,cast<int64_t>(2ULL)))));
}
}
// tools/platform/xbox/nv2a_texture.go:455:1
uint8_t xbox_exp6(uint16_t v){
{
return cast<uint8_t>(cast<uint16_t>((shl<uint16_t>(v,cast<int64_t>(2ULL)) | shr<uint16_t>(v,cast<int64_t>(4ULL)))));
}
}
// tools/platform/xbox/nv2a_texture.go:457:1
void xbox_decode565(Slice<uint8_t> pix,int64_t o,Slice<uint8_t> b){
{
uint16_t v = cast<uint16_t>((cast<uint16_t>(b[cast<int64_t>(0ULL)]) | shl<uint16_t>(cast<uint16_t>(b[cast<int64_t>(1ULL)]),cast<int64_t>(8ULL))));
pix[o] = xbox_exp5(cast<uint16_t>((shr<uint16_t>(v,cast<int64_t>(11ULL)) & cast<uint16_t>(31ULL))));
pix[cast<int64_t>((o + cast<int64_t>(1ULL)))] = xbox_exp6(cast<uint16_t>((shr<uint16_t>(v,cast<int64_t>(5ULL)) & cast<uint16_t>(63ULL))));
pix[cast<int64_t>((o + cast<int64_t>(2ULL)))] = xbox_exp5(cast<uint16_t>((v & cast<uint16_t>(31ULL))));
pix[cast<int64_t>((o + cast<int64_t>(3ULL)))] = cast<uint8_t>(255ULL);
}
}
// tools/platform/xbox/nv2a_texture.go:465:1
void xbox_decode1555(Slice<uint8_t> pix,int64_t o,Slice<uint8_t> b){
{
uint16_t v = cast<uint16_t>((cast<uint16_t>(b[cast<int64_t>(0ULL)]) | shl<uint16_t>(cast<uint16_t>(b[cast<int64_t>(1ULL)]),cast<int64_t>(8ULL))));
pix[o] = xbox_exp5(cast<uint16_t>((shr<uint16_t>(v,cast<int64_t>(10ULL)) & cast<uint16_t>(31ULL))));
pix[cast<int64_t>((o + cast<int64_t>(1ULL)))] = xbox_exp5(cast<uint16_t>((shr<uint16_t>(v,cast<int64_t>(5ULL)) & cast<uint16_t>(31ULL))));
pix[cast<int64_t>((o + cast<int64_t>(2ULL)))] = xbox_exp5(cast<uint16_t>((v & cast<uint16_t>(31ULL))));
if ((shr<uint16_t>(v,cast<int64_t>(15ULL)) != cast<uint16_t>(0ULL))) {
pix[cast<int64_t>((o + cast<int64_t>(3ULL)))] = cast<uint8_t>(255ULL);
}
}
}
// tools/platform/xbox/nv2a_texture.go:475:1
void xbox_decode4444(Slice<uint8_t> pix,int64_t o,Slice<uint8_t> b){
{
uint16_t v = cast<uint16_t>((cast<uint16_t>(b[cast<int64_t>(0ULL)]) | shl<uint16_t>(cast<uint16_t>(b[cast<int64_t>(1ULL)]),cast<int64_t>(8ULL))));
pix[o] = cast<uint8_t>(cast<uint16_t>((cast<uint16_t>((shr<uint16_t>(v,cast<int64_t>(8ULL)) & cast<uint16_t>(15ULL))) * cast<uint16_t>(17ULL))));
pix[cast<int64_t>((o + cast<int64_t>(1ULL)))] = cast<uint8_t>(cast<uint16_t>((cast<uint16_t>((shr<uint16_t>(v,cast<int64_t>(4ULL)) & cast<uint16_t>(15ULL))) * cast<uint16_t>(17ULL))));
pix[cast<int64_t>((o + cast<int64_t>(2ULL)))] = cast<uint8_t>(cast<uint16_t>((cast<uint16_t>((v & cast<uint16_t>(15ULL))) * cast<uint16_t>(17ULL))));
pix[cast<int64_t>((o + cast<int64_t>(3ULL)))] = cast<uint8_t>(cast<uint16_t>((cast<uint16_t>((shr<uint16_t>(v,cast<int64_t>(12ULL)) & cast<uint16_t>(15ULL))) * cast<uint16_t>(17ULL))));
}
}
// tools/platform/xbox/nv2a_texture.go:486:1
void xbox_decodeSwizzled(xbox_texImage* img,Slice<uint8_t> ram,uint32_t phys,int64_t bpp,std::function<void(Slice<uint8_t>,int64_t,Slice<uint8_t>)> put){
{
{int64_t y = cast<int64_t>(0ULL);for (;(y < img->h);y++){
{int64_t x = cast<int64_t>(0ULL);for (;(x < img->w);x++){
uint32_t a = cast<uint32_t>((phys + cast<uint32_t>((cast<uint32_t>(xbox_swizzleOffset(x,y,img->w,img->h)) * cast<uint32_t>(bpp)))));
if ((cast<int64_t>((cast<int64_t>(a) + bpp)) > len(ram))) {
continue;
}
put(img->pix,cast<int64_t>(((cast<int64_t>((cast<int64_t>((y * img->w)) + x))) * cast<int64_t>(4ULL))),sub(ram,a,cast<uint32_t>((a + cast<uint32_t>(bpp)))));
}
}}
}}
}
// tools/platform/xbox/nv2a_texture.go:500:1
int64_t xbox_swizzleOffset(int64_t x,int64_t y,int64_t w,int64_t h){
{
auto tmp285 = std::make_tuple(cast<int64_t>(0ULL),cast<int64_t>(0ULL));
int64_t off = std::get<0>(tmp285);
int64_t shift = std::get<1>(tmp285);
{;for (;((w > cast<int64_t>(1ULL)) || (h > cast<int64_t>(1ULL)));){
if ((w > cast<int64_t>(1ULL))) {
off |= shl<int64_t>((cast<int64_t>((x & cast<int64_t>(1ULL)))),shift);
x = shr<int64_t>(x,cast<int64_t>(1ULL));
w = shr<int64_t>(w,cast<int64_t>(1ULL));
shift++;
}
if ((h > cast<int64_t>(1ULL))) {
off |= shl<int64_t>((cast<int64_t>((y & cast<int64_t>(1ULL)))),shift);
y = shr<int64_t>(y,cast<int64_t>(1ULL));
h = shr<int64_t>(h,cast<int64_t>(1ULL));
shift++;
}
}
}return off;
}
}
// tools/platform/xbox/nv2a_texture.go:519:1
void xbox_decodeLinear(xbox_texImage* img,Slice<uint8_t> ram,uint32_t phys,uint32_t pitch,int64_t bpp,std::function<void(Slice<uint8_t>,int64_t,Slice<uint8_t>)> put){
{
if ((pitch == cast<uint32_t>(0ULL))) {
pitch = cast<uint32_t>(cast<int64_t>((img->w * bpp)));
}
{int64_t y = cast<int64_t>(0ULL);for (;(y < img->h);y++){
uint32_t row = cast<uint32_t>((phys + cast<uint32_t>((cast<uint32_t>(y) * pitch))));
{int64_t x = cast<int64_t>(0ULL);for (;(x < img->w);x++){
uint32_t a = cast<uint32_t>((row + cast<uint32_t>(cast<int64_t>((x * bpp)))));
if ((cast<int64_t>((cast<int64_t>(a) + bpp)) > len(ram))) {
continue;
}
put(img->pix,cast<int64_t>(((cast<int64_t>((cast<int64_t>((y * img->w)) + x))) * cast<int64_t>(4ULL))),sub(ram,a,cast<uint32_t>((a + cast<uint32_t>(bpp)))));
}
}}
}}
}
// tools/platform/xbox/nv2a_texture.go:538:1
void xbox_decodeDXT(xbox_texImage* img,Slice<uint8_t> ram,uint32_t phys,int64_t variant){
{
uint32_t blockBytes = cast<uint32_t>(8ULL);
if ((variant != cast<int64_t>(1ULL))) {
blockBytes = cast<uint32_t>(16ULL);
}
auto tmp286 = std::make_tuple(divi<int64_t>((cast<int64_t>((img->w + cast<int64_t>(3ULL)))),cast<int64_t>(4ULL)),divi<int64_t>((cast<int64_t>((img->h + cast<int64_t>(3ULL)))),cast<int64_t>(4ULL)));
int64_t bw = std::get<0>(tmp286);
int64_t bh = std::get<1>(tmp286);
{int64_t by = cast<int64_t>(0ULL);for (;(by < bh);by++){
{int64_t bx = cast<int64_t>(0ULL);for (;(bx < bw);bx++){
uint32_t a = cast<uint32_t>((phys + cast<uint32_t>((cast<uint32_t>(cast<int64_t>((cast<int64_t>((by * bw)) + bx))) * blockBytes))));
if ((cast<int64_t>((cast<int64_t>(a) + cast<int64_t>(blockBytes))) > len(ram))) {
continue;
}
Slice<uint8_t> blk = sub(ram,a,cast<uint32_t>((a + blockBytes)));
Slice<uint8_t> color = blk;
if ((variant != cast<int64_t>(1ULL))) {
color = sub(blk,cast<int64_t>(8ULL),len(blk));
}
uint16_t c0 = cast<uint16_t>((cast<uint16_t>(color[cast<int64_t>(0ULL)]) | shl<uint16_t>(cast<uint16_t>(color[cast<int64_t>(1ULL)]),cast<int64_t>(8ULL))));
uint16_t c1 = cast<uint16_t>((cast<uint16_t>(color[cast<int64_t>(2ULL)]) | shl<uint16_t>(cast<uint16_t>(color[cast<int64_t>(3ULL)]),cast<int64_t>(8ULL))));
std::array<std::array<uint8_t,4>,4> pal={};
auto expand = [&](uint16_t c)->std::tuple<uint8_t,uint8_t,uint8_t>{
return {xbox_exp5(cast<uint16_t>((shr<uint16_t>(c,cast<int64_t>(11ULL)) & cast<uint16_t>(31ULL)))),xbox_exp6(cast<uint16_t>((shr<uint16_t>(c,cast<int64_t>(5ULL)) & cast<uint16_t>(63ULL)))),xbox_exp5(cast<uint16_t>((c & cast<uint16_t>(31ULL))))};
}
;
auto tmp287 = expand(c0);
uint8_t r0 = std::get<0>(tmp287);
uint8_t g0 = std::get<1>(tmp287);
uint8_t b0 = std::get<2>(tmp287);
auto tmp288 = expand(c1);
uint8_t r1 = std::get<0>(tmp288);
uint8_t g1 = std::get<1>(tmp288);
uint8_t b1 = std::get<2>(tmp288);
pal[cast<int64_t>(0ULL)] = std::array<uint8_t,4>{r0,g0,b0,cast<uint8_t>(255ULL)};
pal[cast<int64_t>(1ULL)] = std::array<uint8_t,4>{r1,g1,b1,cast<uint8_t>(255ULL)};
if (((variant != cast<int64_t>(1ULL)) || (c0 > c1))) {
pal[cast<int64_t>(2ULL)] = std::array<uint8_t,4>{cast<uint8_t>(divi<int64_t>((cast<int64_t>((cast<int64_t>((cast<int64_t>(2ULL) * cast<int64_t>(r0))) + cast<int64_t>(r1)))),cast<int64_t>(3ULL))),cast<uint8_t>(divi<int64_t>((cast<int64_t>((cast<int64_t>((cast<int64_t>(2ULL) * cast<int64_t>(g0))) + cast<int64_t>(g1)))),cast<int64_t>(3ULL))),cast<uint8_t>(divi<int64_t>((cast<int64_t>((cast<int64_t>((cast<int64_t>(2ULL) * cast<int64_t>(b0))) + cast<int64_t>(b1)))),cast<int64_t>(3ULL))),cast<uint8_t>(255ULL)};
pal[cast<int64_t>(3ULL)] = std::array<uint8_t,4>{cast<uint8_t>(divi<int64_t>((cast<int64_t>((cast<int64_t>(r0) + cast<int64_t>((cast<int64_t>(2ULL) * cast<int64_t>(r1)))))),cast<int64_t>(3ULL))),cast<uint8_t>(divi<int64_t>((cast<int64_t>((cast<int64_t>(g0) + cast<int64_t>((cast<int64_t>(2ULL) * cast<int64_t>(g1)))))),cast<int64_t>(3ULL))),cast<uint8_t>(divi<int64_t>((cast<int64_t>((cast<int64_t>(b0) + cast<int64_t>((cast<int64_t>(2ULL) * cast<int64_t>(b1)))))),cast<int64_t>(3ULL))),cast<uint8_t>(255ULL)};
}
else {
pal[cast<int64_t>(2ULL)] = std::array<uint8_t,4>{cast<uint8_t>(divi<int64_t>((cast<int64_t>((cast<int64_t>(r0) + cast<int64_t>(r1)))),cast<int64_t>(2ULL))),cast<uint8_t>(divi<int64_t>((cast<int64_t>((cast<int64_t>(g0) + cast<int64_t>(g1)))),cast<int64_t>(2ULL))),cast<uint8_t>(divi<int64_t>((cast<int64_t>((cast<int64_t>(b0) + cast<int64_t>(b1)))),cast<int64_t>(2ULL))),cast<uint8_t>(255ULL)};
pal[cast<int64_t>(3ULL)] = std::array<uint8_t,4>{cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),cast<uint8_t>(0ULL)};
}
uint32_t sel = cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(color[cast<int64_t>(4ULL)]) | shl<uint32_t>(cast<uint32_t>(color[cast<int64_t>(5ULL)]),cast<int64_t>(8ULL)))) | shl<uint32_t>(cast<uint32_t>(color[cast<int64_t>(6ULL)]),cast<int64_t>(16ULL)))) | shl<uint32_t>(cast<uint32_t>(color[cast<int64_t>(7ULL)]),cast<int64_t>(24ULL))));
{int64_t ty = cast<int64_t>(0ULL);for (;(ty < cast<int64_t>(4ULL));ty++){
int64_t y = cast<int64_t>((cast<int64_t>((by * cast<int64_t>(4ULL))) + ty));
if ((y >= img->h)) {
break;
}
{int64_t tx = cast<int64_t>(0ULL);for (;(tx < cast<int64_t>(4ULL));tx++){
int64_t x = cast<int64_t>((cast<int64_t>((bx * cast<int64_t>(4ULL))) + tx));
if ((x >= img->w)) {
continue;
}
std::array<uint8_t,4> p = pal[cast<uint32_t>((shr<uint32_t>(sel,(cast<uint64_t>((cast<uint64_t>(2ULL) * cast<uint64_t>(cast<int64_t>((cast<int64_t>((ty * cast<int64_t>(4ULL))) + tx))))))) & cast<uint32_t>(3ULL)))];
{
switch(variant){
case cast<int64_t>(3ULL):{
uint8_t a4 = cast<uint8_t>((shr<uint8_t>(blk[cast<int64_t>((cast<int64_t>((ty * cast<int64_t>(2ULL))) + divi<int64_t>(tx,cast<int64_t>(2ULL))))],(cast<uint64_t>((cast<uint64_t>(4ULL) * cast<uint64_t>(cast<int64_t>((tx & cast<int64_t>(1ULL)))))))) & cast<uint8_t>(15ULL)));
p[cast<int64_t>(3ULL)] = cast<uint8_t>((a4 * cast<uint8_t>(17ULL)));
break;}
case cast<int64_t>(5ULL):{
p[cast<int64_t>(3ULL)] = xbox_dxt5Alpha(blk,cast<int64_t>((cast<int64_t>((ty * cast<int64_t>(4ULL))) + tx)));
break;}
}}
int64_t o = cast<int64_t>(((cast<int64_t>((cast<int64_t>((y * img->w)) + x))) * cast<int64_t>(4ULL)));
auto tmp289 = std::make_tuple(p[cast<int64_t>(0ULL)],p[cast<int64_t>(1ULL)],p[cast<int64_t>(2ULL)],p[cast<int64_t>(3ULL)]);
img->pix[o] = std::get<0>(tmp289);
img->pix[cast<int64_t>((o + cast<int64_t>(1ULL)))] = std::get<1>(tmp289);
img->pix[cast<int64_t>((o + cast<int64_t>(2ULL)))] = std::get<2>(tmp289);
img->pix[cast<int64_t>((o + cast<int64_t>(3ULL)))] = std::get<3>(tmp289);
}
}}
}}
}}
}}
}
// tools/platform/xbox/nv2a_texture.go:599:1
uint8_t xbox_dxt5Alpha(Slice<uint8_t> blk,int64_t i){
{
auto tmp290 = std::make_tuple(cast<int64_t>(blk[cast<int64_t>(0ULL)]),cast<int64_t>(blk[cast<int64_t>(1ULL)]));
int64_t a0 = std::get<0>(tmp290);
int64_t a1 = std::get<1>(tmp290);
uint64_t bits = cast<uint64_t>((cast<uint64_t>((cast<uint64_t>((cast<uint64_t>((cast<uint64_t>((cast<uint64_t>(blk[cast<int64_t>(2ULL)]) | shl<uint64_t>(cast<uint64_t>(blk[cast<int64_t>(3ULL)]),cast<int64_t>(8ULL)))) | shl<uint64_t>(cast<uint64_t>(blk[cast<int64_t>(4ULL)]),cast<int64_t>(16ULL)))) | shl<uint64_t>(cast<uint64_t>(blk[cast<int64_t>(5ULL)]),cast<int64_t>(24ULL)))) | shl<uint64_t>(cast<uint64_t>(blk[cast<int64_t>(6ULL)]),cast<int64_t>(32ULL)))) | shl<uint64_t>(cast<uint64_t>(blk[cast<int64_t>(7ULL)]),cast<int64_t>(40ULL))));
int64_t code = cast<int64_t>(cast<uint64_t>((shr<uint64_t>(bits,(cast<uint64_t>((cast<uint64_t>(3ULL) * cast<uint64_t>(i))))) & cast<uint64_t>(7ULL))));
{
if ((code == cast<int64_t>(0ULL))){
return cast<uint8_t>(a0);
}
else if ((code == cast<int64_t>(1ULL))){
return cast<uint8_t>(a1);
}
else if ((a0 > a1)){
return cast<uint8_t>(divi<int64_t>((cast<int64_t>((cast<int64_t>(((cast<int64_t>((cast<int64_t>(8ULL) - code))) * a0)) + cast<int64_t>(((cast<int64_t>((code - cast<int64_t>(1ULL)))) * a1))))),cast<int64_t>(7ULL)));
}
else if ((code == cast<int64_t>(6ULL))){
return cast<uint8_t>(0ULL);
}
else if ((code == cast<int64_t>(7ULL))){
return cast<uint8_t>(255ULL);
}
else {
return cast<uint8_t>(divi<int64_t>((cast<int64_t>((cast<int64_t>(((cast<int64_t>((cast<int64_t>(6ULL) - code))) * a0)) + cast<int64_t>(((cast<int64_t>((code - cast<int64_t>(1ULL)))) * a1))))),cast<int64_t>(5ULL)));
}
}
tmp291:;
}
}
// tools/platform/xbox/nv2a_texture.go:623:1
float xbox_texWrapCoord(float v,int64_t size,uint32_t mode){
{
float n = cast<float>(size);
{
switch(mode){
case cast<uint32_t>(1ULL):{
v -= (n * xbox_floorf32((v / n)));
break;}
case cast<uint32_t>(2ULL):{
float p = (n * cast<float>(2.00000000000000000e+00));
v -= (p * xbox_floorf32((v / p)));
if ((v >= n)) {
v = ((p - v) - cast<float>(3.90625000000000000e-03));
}
break;}
default:{
if ((v < cast<float>(0.00000000000000000e+00))) {
v = cast<float>(0.00000000000000000e+00);
}
if ((v > (n - cast<float>(1.00000000000000000e+00)))) {
v = (n - cast<float>(1.00000000000000000e+00));
}
break;}
}}
return v;
}
}
// tools/platform/xbox/nv2a_texture.go:649:1
std::array<float,4> xbox_pgraph_texSampleCube(xbox_pgraph* g,xbox_rasterState* st,int64_t u,float x,float y,float z){
{
xbox_texImage* img = st->texImg[u];
auto tmp292 = std::make_tuple(xbox_absf32(x),xbox_absf32(y),xbox_absf32(z));
float ax = std::get<0>(tmp292);
float ay = std::get<1>(tmp292);
float az = std::get<2>(tmp292);
int64_t face={};
float sc={};
float tc={};
float ma={};
{
if (((ax >= ay) && (ax >= az))){
if ((x >= cast<float>(0.00000000000000000e+00))) {
auto tmp294 = std::make_tuple(cast<int64_t>(0ULL),cast<float>(-z),cast<float>(-y));
face = std::get<0>(tmp294);
sc = std::get<1>(tmp294);
tc = std::get<2>(tmp294);
}
else {
auto tmp295 = std::make_tuple(cast<int64_t>(1ULL),z,cast<float>(-y));
face = std::get<0>(tmp295);
sc = std::get<1>(tmp295);
tc = std::get<2>(tmp295);
}
ma = ax;
}
else if ((ay >= az)){
if ((y >= cast<float>(0.00000000000000000e+00))) {
auto tmp296 = std::make_tuple(cast<int64_t>(2ULL),x,z);
face = std::get<0>(tmp296);
sc = std::get<1>(tmp296);
tc = std::get<2>(tmp296);
}
else {
auto tmp297 = std::make_tuple(cast<int64_t>(3ULL),x,cast<float>(-z));
face = std::get<0>(tmp297);
sc = std::get<1>(tmp297);
tc = std::get<2>(tmp297);
}
ma = ay;
}
else {
if ((z >= cast<float>(0.00000000000000000e+00))) {
auto tmp298 = std::make_tuple(cast<int64_t>(4ULL),x,cast<float>(-y));
face = std::get<0>(tmp298);
sc = std::get<1>(tmp298);
tc = std::get<2>(tmp298);
}
else {
auto tmp299 = std::make_tuple(cast<int64_t>(5ULL),cast<float>(-x),cast<float>(-y));
face = std::get<0>(tmp299);
sc = std::get<1>(tmp299);
tc = std::get<2>(tmp299);
}
ma = az;
}
}
tmp293:;
if ((ma == cast<float>(0.00000000000000000e+00))) {
return std::array<float,4>{cast<float>(0.00000000000000000e+00),cast<float>(0.00000000000000000e+00),cast<float>(0.00000000000000000e+00),cast<float>(1.00000000000000000e+00)};
}
float fx = (((((sc / ma) + cast<float>(1.00000000000000000e+00))) * cast<float>(5.00000000000000000e-01)) * cast<float>(img->w));
float fy = (((((tc / ma) + cast<float>(1.00000000000000000e+00))) * cast<float>(5.00000000000000000e-01)) * cast<float>(img->h));
auto clampi = [&](int64_t v,int64_t hi)->int64_t{
if ((v < cast<int64_t>(0ULL))) {
return cast<int64_t>(0ULL);
}
if ((v > hi)) {
return hi;
}
return v;
}
;
auto fetch = [&](int64_t xi,int64_t yi)->std::array<float,4>{
int64_t o = cast<int64_t>(((cast<int64_t>((cast<int64_t>(((cast<int64_t>((cast<int64_t>((face * img->h)) + yi))) * img->w)) + xi))) * cast<int64_t>(4ULL)));
return std::array<float,4>{(cast<float>(img->pix[o]) / cast<float>(2.55000000000000000e+02)),(cast<float>(img->pix[cast<int64_t>((o + cast<int64_t>(1ULL)))]) / cast<float>(2.55000000000000000e+02)),(cast<float>(img->pix[cast<int64_t>((o + cast<int64_t>(2ULL)))]) / cast<float>(2.55000000000000000e+02)),(cast<float>(img->pix[cast<int64_t>((o + cast<int64_t>(3ULL)))]) / cast<float>(2.55000000000000000e+02))};
}
;
if ((!st->texBilinear[u])) {
return fetch(clampi(cast<int64_t>(fx),cast<int64_t>((img->w - cast<int64_t>(1ULL)))),clampi(cast<int64_t>(fy),cast<int64_t>((img->h - cast<int64_t>(1ULL)))));
}
auto tmp300 = std::make_tuple((fx - cast<float>(5.00000000000000000e-01)),(fy - cast<float>(5.00000000000000000e-01)));
float gx = std::get<0>(tmp300);
float gy = std::get<1>(tmp300);
auto tmp301 = std::make_tuple(xbox_floorf32(gx),xbox_floorf32(gy));
float x0f = std::get<0>(tmp301);
float y0f = std::get<1>(tmp301);
auto tmp302 = std::make_tuple((gx - x0f),(gy - y0f));
float wx = std::get<0>(tmp302);
float wy = std::get<1>(tmp302);
auto tmp303 = std::make_tuple(clampi(cast<int64_t>(x0f),cast<int64_t>((img->w - cast<int64_t>(1ULL)))),clampi(cast<int64_t>((cast<int64_t>(x0f) + cast<int64_t>(1ULL))),cast<int64_t>((img->w - cast<int64_t>(1ULL)))));
int64_t x0 = std::get<0>(tmp303);
int64_t x1 = std::get<1>(tmp303);
auto tmp304 = std::make_tuple(clampi(cast<int64_t>(y0f),cast<int64_t>((img->h - cast<int64_t>(1ULL)))),clampi(cast<int64_t>((cast<int64_t>(y0f) + cast<int64_t>(1ULL))),cast<int64_t>((img->h - cast<int64_t>(1ULL)))));
int64_t y0 = std::get<0>(tmp304);
int64_t y1 = std::get<1>(tmp304);
auto tmp305 = std::make_tuple(fetch(x0,y0),fetch(x1,y0));
std::array<float,4> c00 = std::get<0>(tmp305);
std::array<float,4> c10 = std::get<1>(tmp305);
auto tmp306 = std::make_tuple(fetch(x0,y1),fetch(x1,y1));
std::array<float,4> c01 = std::get<0>(tmp306);
std::array<float,4> c11 = std::get<1>(tmp306);
std::array<float,4> out={};
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(4ULL));i++){
float top = (c00[i] + (((c10[i] - c00[i])) * wx));
float bot = (c01[i] + (((c11[i] - c01[i])) * wx));
out[i] = (top + (((bot - top)) * wy));
}
}return out;
}
}
// tools/platform/xbox/nv2a_texture.go:717:1
float xbox_absf32(float v){
{
if ((v < cast<float>(0.00000000000000000e+00))) {
return cast<float>(-v);
}
return v;
}
}
// tools/platform/xbox/nv2a_texture.go:729:1
void xbox_pgraph_traceShadowFrag(xbox_pgraph* g,xbox_rasterState* st,int64_t u,int64_t px,int64_t py,float s,float t,float r,float q){
{
xbox_texImage* img = st->texImg[u];
int64_t x = cast<int64_t>(xbox_texWrapCoord(s,img->w,st->texWrapU[u]));
int64_t y = cast<int64_t>(xbox_texWrapCoord(t,img->h,st->texWrapV[u]));
uint32_t d = img->depth[cast<int64_t>((cast<int64_t>((y * img->w)) + x))];
if ((bool(g->shadowFrag) && (g->shadowFrag->draw != g->Draws))) {
xbox_shadowFragStats_print(g->shadowFrag);
g->shadowFrag = {};
}
if ((!g->shadowFrag)) {
g->shadowFrag = arenaNew(xbox_shadowFragStats{g->Draws,{},{},{},px,px,py,py,cast<float>(r),cast<float>(r),cast<uint32_t>(16777215ULL),cast<uint32_t>(0ULL),cast<float>(cast<float>(0.00000000000000000e+00)),cast<float>(cast<float>(0.00000000000000000e+00))});
}
xbox_shadowFragStats* f = g->shadowFrag;
if ((px < f->pxMin)) {
f->pxMin = px;
}
if ((px > f->pxMax)) {
f->pxMax = px;
}
if ((py < f->pyMin)) {
f->pyMin = py;
}
if ((py > f->pyMax)) {
f->pyMax = py;
}
if ((r < f->rMin)) {
f->rMin = r;
}
if ((r > f->rMax)) {
f->rMax = r;
}
if ((d != cast<uint32_t>(16777215ULL))) {
float diff = (r - cast<float>(d));
if ((f->near == cast<int64_t>(0ULL))) {
auto tmp307 = std::make_tuple(diff,diff);
f->diffMin = std::get<0>(tmp307);
f->diffMax = std::get<1>(tmp307);
}
if ((diff < f->diffMin)) {
f->diffMin = diff;
}
if ((diff > f->diffMax)) {
f->diffMax = diff;
}
if ((d < f->dMin)) {
f->dMin = d;
}
if ((d > f->dMax)) {
f->dMax = d;
}
f->near++;
if ((diff > cast<float>(0.00000000000000000e+00))) {
f->behind++;
}
}
else {
f->far++;
}
}
}
// tools/platform/xbox/nv2a_texture.go:796:1
void xbox_shadowFragStats_print(xbox_shadowFragStats* f){
{
go_fmt_Printf(std::string("SHADOWFRAG draw=%d frags=%d near=%d behind=%d px=[%d,%d]x[%d,%d] r=[%.0f,%.0f] D=[%06X,%06X] r-D=[%.0f,%.0f]\012",109),f->draw,cast<int64_t>((f->near + f->far)),f->near,f->behind,f->pxMin,f->pxMax,f->pyMin,f->pyMax,f->rMin,f->rMax,f->dMin,f->dMax,f->diffMin,f->diffMax);
}
}
// tools/platform/xbox/nv2a_texture.go:827:1
bool xbox_shadowComparePass(float r,uint32_t d){
{
{
auto tmp309=xbox_shadowCmpEnv;
if (tmp309==(std::string("lt",2))){
return (r < cast<float>(d));
}
else if (tmp309==(std::string("gt",2))){
return (r > cast<float>(d));
}
else if (tmp309==(std::string("ge",2))){
return (r >= cast<float>(d));
}
else if (tmp309==(std::string("one",3))){
return true;
}
else if (tmp309==(std::string("zero",4))){
return false;
}
}
tmp308:;
return (r <= cast<float>(d));
}
}
// tools/platform/xbox/nv2a_texture.go:852:1
std::array<float,4> xbox_pgraph_texSampleShadow(xbox_pgraph* g,xbox_rasterState* st,int64_t u,float s,float t,float r){
{
xbox_texImage* img = st->texImg[u];
auto cmp = [&](int64_t x,int64_t y)->float{
if (xbox_shadowComparePass(r,img->depth[cast<int64_t>((cast<int64_t>((y * img->w)) + x))])) {
return cast<float>(1.00000000000000000e+00);
}
return cast<float>(0.00000000000000000e+00);
}
;
if ((!st->texBilinear[u])) {
int64_t x = cast<int64_t>(xbox_texWrapCoord(s,img->w,st->texWrapU[u]));
int64_t y = cast<int64_t>(xbox_texWrapCoord(t,img->h,st->texWrapV[u]));
float v = cmp(x,y);
return std::array<float,4>{v,v,v,v};
}
auto tmp310 = std::make_tuple((s - cast<float>(5.00000000000000000e-01)),(t - cast<float>(5.00000000000000000e-01)));
float gx = std::get<0>(tmp310);
float gy = std::get<1>(tmp310);
auto tmp311 = std::make_tuple(xbox_floorf32(gx),xbox_floorf32(gy));
float x0f = std::get<0>(tmp311);
float y0f = std::get<1>(tmp311);
auto tmp312 = std::make_tuple((gx - x0f),(gy - y0f));
float wx = std::get<0>(tmp312);
float wy = std::get<1>(tmp312);
int64_t x0 = cast<int64_t>(xbox_texWrapCoord(x0f,img->w,st->texWrapU[u]));
int64_t x1 = cast<int64_t>(xbox_texWrapCoord((x0f + cast<float>(1.00000000000000000e+00)),img->w,st->texWrapU[u]));
int64_t y0 = cast<int64_t>(xbox_texWrapCoord(y0f,img->h,st->texWrapV[u]));
int64_t y1 = cast<int64_t>(xbox_texWrapCoord((y0f + cast<float>(1.00000000000000000e+00)),img->h,st->texWrapV[u]));
float top = (cmp(x0,y0) + (((cmp(x1,y0) - cmp(x0,y0))) * wx));
float bot = (cmp(x0,y1) + (((cmp(x1,y1) - cmp(x0,y1))) * wx));
float v = (top + (((bot - top)) * wy));
return std::array<float,4>{v,v,v,v};
}
}
// tools/platform/xbox/nv2a_texture.go:881:1
std::array<float,4> xbox_pgraph_texSample(xbox_pgraph* g,xbox_rasterState* st,int64_t u,float s,float t){
{
xbox_texImage* img = st->texImg[u];
auto tmp313 = std::make_tuple(s,t);
float fx = std::get<0>(tmp313);
float fy = std::get<1>(tmp313);
if ((!st->texRect[u])) {
fx *= cast<float>(img->w);
fy *= cast<float>(img->h);
}
auto fetch = [&](int64_t x,int64_t y)->std::array<float,4>{
int64_t o = cast<int64_t>(((cast<int64_t>((cast<int64_t>((y * img->w)) + x))) * cast<int64_t>(4ULL)));
return std::array<float,4>{(cast<float>(img->pix[o]) / cast<float>(2.55000000000000000e+02)),(cast<float>(img->pix[cast<int64_t>((o + cast<int64_t>(1ULL)))]) / cast<float>(2.55000000000000000e+02)),(cast<float>(img->pix[cast<int64_t>((o + cast<int64_t>(2ULL)))]) / cast<float>(2.55000000000000000e+02)),(cast<float>(img->pix[cast<int64_t>((o + cast<int64_t>(3ULL)))]) / cast<float>(2.55000000000000000e+02))};
}
;
if ((!st->texBilinear[u])) {
int64_t x = cast<int64_t>(xbox_texWrapCoord(fx,img->w,st->texWrapU[u]));
int64_t y = cast<int64_t>(xbox_texWrapCoord(fy,img->h,st->texWrapV[u]));
return fetch(x,y);
}
auto tmp314 = std::make_tuple((fx - cast<float>(5.00000000000000000e-01)),(fy - cast<float>(5.00000000000000000e-01)));
float gx = std::get<0>(tmp314);
float gy = std::get<1>(tmp314);
auto tmp315 = std::make_tuple(xbox_floorf32(gx),xbox_floorf32(gy));
float x0f = std::get<0>(tmp315);
float y0f = std::get<1>(tmp315);
auto tmp316 = std::make_tuple((gx - x0f),(gy - y0f));
float wx = std::get<0>(tmp316);
float wy = std::get<1>(tmp316);
int64_t x0 = cast<int64_t>(xbox_texWrapCoord(x0f,img->w,st->texWrapU[u]));
int64_t x1 = cast<int64_t>(xbox_texWrapCoord((x0f + cast<float>(1.00000000000000000e+00)),img->w,st->texWrapU[u]));
int64_t y0 = cast<int64_t>(xbox_texWrapCoord(y0f,img->h,st->texWrapV[u]));
int64_t y1 = cast<int64_t>(xbox_texWrapCoord((y0f + cast<float>(1.00000000000000000e+00)),img->h,st->texWrapV[u]));
auto tmp317 = std::make_tuple(fetch(x0,y0),fetch(x1,y0));
std::array<float,4> c00 = std::get<0>(tmp317);
std::array<float,4> c10 = std::get<1>(tmp317);
auto tmp318 = std::make_tuple(fetch(x0,y1),fetch(x1,y1));
std::array<float,4> c01 = std::get<0>(tmp318);
std::array<float,4> c11 = std::get<1>(tmp318);
std::array<float,4> out={};
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(4ULL));i++){
float top = (c00[i] + (((c10[i] - c00[i])) * wx));
float bot = (c01[i] + (((c11[i] - c01[i])) * wx));
out[i] = (top + (((bot - top)) * wy));
}
}return out;
}
}
// tools/platform/xbox/nv2a_vertex.go:85:1
xbox_vtxFmt xbox_pgraph_attrFormat(xbox_pgraph* g,int64_t i){
{
uint32_t f = g->Regs[cast<uint32_t>((cast<uint32_t>(1496ULL) + cast<uint32_t>(i)))];
xbox_vtxFmt vf = xbox_vtxFmt{cast<int64_t>(cast<uint32_t>((f & cast<uint32_t>(15ULL)))),cast<int64_t>(cast<uint32_t>((shr<uint32_t>(f,cast<int64_t>(4ULL)) & cast<uint32_t>(15ULL)))),cast<uint32_t>((shr<uint32_t>(f,cast<int64_t>(8ULL)) & cast<uint32_t>(255ULL))),{}};
{
switch(vf.typ){
case cast<int64_t>(0ULL):case cast<int64_t>(4ULL):{
if ((vf.size > cast<int64_t>(0ULL))) {
vf.words = cast<int64_t>(1ULL);
}
break;}
case cast<int64_t>(1ULL):case cast<int64_t>(5ULL):{
vf.words = divi<int64_t>((cast<int64_t>((vf.size + cast<int64_t>(1ULL)))),cast<int64_t>(2ULL));
break;}
case cast<int64_t>(2ULL):{
vf.words = vf.size;
break;}
case cast<int64_t>(6ULL):{
if ((vf.size > cast<int64_t>(0ULL))) {
vf.words = cast<int64_t>(1ULL);
}
break;}
}}
return vf;
}
}
// tools/platform/xbox/nv2a_vertex.go:107:1
std::array<float,4> xbox_decodeAttr(xbox_vtxFmt* vf,Slice<uint32_t> w){
{
std::array<float,4> val = std::array<float,4>{cast<float>(0.00000000000000000e+00),cast<float>(0.00000000000000000e+00),cast<float>(0.00000000000000000e+00),cast<float>(1.00000000000000000e+00)};
{
switch(vf->typ){
case cast<int64_t>(0ULL):{
val[cast<int64_t>(0ULL)] = (cast<float>(cast<uint32_t>((shr<uint32_t>(w[cast<int64_t>(0ULL)],cast<int64_t>(16ULL)) & cast<uint32_t>(255ULL)))) / cast<float>(2.55000000000000000e+02));
val[cast<int64_t>(1ULL)] = (cast<float>(cast<uint32_t>((shr<uint32_t>(w[cast<int64_t>(0ULL)],cast<int64_t>(8ULL)) & cast<uint32_t>(255ULL)))) / cast<float>(2.55000000000000000e+02));
val[cast<int64_t>(2ULL)] = (cast<float>(cast<uint32_t>((w[cast<int64_t>(0ULL)] & cast<uint32_t>(255ULL)))) / cast<float>(2.55000000000000000e+02));
val[cast<int64_t>(3ULL)] = (cast<float>(cast<uint32_t>((shr<uint32_t>(w[cast<int64_t>(0ULL)],cast<int64_t>(24ULL)) & cast<uint32_t>(255ULL)))) / cast<float>(2.55000000000000000e+02));
break;}
case cast<int64_t>(4ULL):{
{int64_t i = cast<int64_t>(0ULL);for (;((i < vf->size) && (i < cast<int64_t>(4ULL)));i++){
val[i] = (cast<float>(cast<uint32_t>((shr<uint32_t>(w[cast<int64_t>(0ULL)],(cast<uint64_t>((cast<uint64_t>(8ULL) * cast<uint64_t>(i))))) & cast<uint32_t>(255ULL)))) / cast<float>(2.55000000000000000e+02));
}
}break;}
case cast<int64_t>(1ULL):case cast<int64_t>(5ULL):{
{int64_t i = cast<int64_t>(0ULL);for (;((i < vf->size) && (i < cast<int64_t>(4ULL)));i++){
int16_t raw = cast<int16_t>(shr<uint32_t>(w[divi<int64_t>(i,cast<int64_t>(2ULL))],(cast<uint64_t>((cast<uint64_t>(16ULL) * cast<uint64_t>(cast<int64_t>((i & cast<int64_t>(1ULL)))))))));
if ((vf->typ == cast<int64_t>(1ULL))) {
float v = (cast<float>(raw) / cast<float>(3.27670000000000000e+04));
if ((v < cast<float>(-1.00000000000000000e+00))) {
v = cast<float>(-1.00000000000000000e+00);
}
val[i] = v;
}
else {
val[i] = cast<float>(raw);
}
}
}break;}
case cast<int64_t>(2ULL):{
{int64_t i = cast<int64_t>(0ULL);for (;((i < vf->size) && (i < cast<int64_t>(4ULL)));i++){
val[i] = go_math_Float32frombits(w[i]);
}
}break;}
case cast<int64_t>(6ULL):{
auto sext = [&](uint32_t v,uint64_t bits)->float{
int32_t s = shr<int32_t>(cast<int32_t>(shl<uint32_t>(v,(cast<uint64_t>((cast<uint64_t>(32ULL) - bits))))),(cast<uint64_t>((cast<uint64_t>(32ULL) - bits))));
float f = (cast<float>(s) / cast<float>(cast<int32_t>((shl<int32_t>(cast<int32_t>(1ULL),(cast<uint64_t>((bits - cast<uint64_t>(1ULL))))) - cast<int32_t>(1ULL)))));
if ((f < cast<float>(-1.00000000000000000e+00))) {
f = cast<float>(-1.00000000000000000e+00);
}
return f;
}
;
val[cast<int64_t>(0ULL)] = sext(cast<uint32_t>((w[cast<int64_t>(0ULL)] & cast<uint32_t>(2047ULL))),cast<uint64_t>(11ULL));
val[cast<int64_t>(1ULL)] = sext(cast<uint32_t>((shr<uint32_t>(w[cast<int64_t>(0ULL)],cast<int64_t>(11ULL)) & cast<uint32_t>(2047ULL))),cast<uint64_t>(11ULL));
val[cast<int64_t>(2ULL)] = sext(cast<uint32_t>((shr<uint32_t>(w[cast<int64_t>(0ULL)],cast<int64_t>(22ULL)) & cast<uint32_t>(1023ULL))),cast<uint64_t>(10ULL));
break;}
}}
return val;
}
}
// tools/platform/xbox/nv2a_vertex.go:155:1
void xbox_pgraph_beginEnd(xbox_pgraph* g,uint32_t arg){
{
if ((arg != cast<uint32_t>(0ULL))) {
g->prim = arg;
g->inline_ = sub(g->inline_,0,cast<int64_t>(0ULL));
g->elems = sub(g->elems,0,cast<int64_t>(0ULL));
g->ranges = sub(g->ranges,0,cast<int64_t>(0ULL));
return ;
}
if ((g->prim == cast<uint32_t>(0ULL))) {
return ;
}
xbox_pgraph_runDraw(g);
g->prim = cast<uint32_t>(0ULL);
}
}
// tools/platform/xbox/nv2a_vertex.go:172:1
void xbox_pgraph_runDraw(xbox_pgraph* g){
{rrconsole::EventScope restore(rrcapture::trace.current); if(rrcapture::trace.active)rrXboxDraw(g,"xbox_pgraph_runDraw");
{rrprof::Scope timing(2,"Vertex processing");
{
g->Draws++;
time_Time tv = xbox_Machine_profStart(g->m);
bool ff = (cast<uint32_t>((g->Regs[cast<int64_t>(1957ULL)] & cast<uint32_t>(3ULL))) != cast<uint32_t>(2ULL));
g->ffFragHalt = std::string("",0);
if (ff) {
if ((cast<uint32_t>((g->Regs[cast<int64_t>(197ULL)] & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL))) {
g->ffFragHalt = std::string("lighting is enabled and the FF lighting equation/state is unmodelled",68);
}
else {
uint32_t shader = g->Regs[cast<int64_t>(1948ULL)];
{uint32_t u = cast<uint32_t>(0ULL);for (;(u < cast<uint32_t>(4ULL));u++){
if (((cast<uint32_t>((shr<uint32_t>(g->Regs[shr<uint32_t>((cast<uint32_t>((cast<uint32_t>(6924ULL) + cast<uint32_t>((u * cast<uint32_t>(64ULL)))))),cast<int64_t>(2ULL))],cast<int64_t>(30ULL)) & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL)) && (cast<uint32_t>((shr<uint32_t>(shader,(cast<uint32_t>((cast<uint32_t>(5ULL) * u)))) & cast<uint32_t>(31ULL))) != cast<uint32_t>(0ULL)))) {
g->ffFragHalt = std::string("textures are enabled and FF texgen is unmodelled",48);
}
}
}}
}
std::array<xbox_vtxFmt,16> fmts={};
int64_t stride = cast<int64_t>(0ULL);
{auto&& tmp320 = fmts;
for(int64_t tmp321=0;tmp321<len(tmp320);++tmp321){
auto i=tmp321;fmts[i] = xbox_pgraph_attrFormat(g,i);
stride += fmts[i].words;
}}
bool trace = (g->Draws <= xbox_nvDrawTrace);
if (trace) {
xbox_pgraph_traceDraw(g,(&fmts),stride);
}
if (((!ff) && (!xbox_pgraph_vshCompile(g)))) {
return ;
}
int64_t nin={};
{
if ((len(g->inline_) > cast<int64_t>(0ULL))){
if (((len(g->elems) > cast<int64_t>(0ULL)) || (len(g->ranges) > cast<int64_t>(0ULL)))) {
x86_CPU_Halt(g->m->CPU,std::string("nv2a: draw %d mixes inline and indexed vertex submission",56),g->Draws);
return ;
}
if ((stride == cast<int64_t>(0ULL))) {
x86_CPU_Halt(g->m->CPU,std::string("nv2a: draw %d has inline data but no enabled attributes",55),g->Draws);
return ;
}
if ((modi<int64_t>(len(g->inline_),stride) != cast<int64_t>(0ULL))) {
x86_CPU_Halt(g->m->CPU,std::string("nv2a: draw %d inline data %d words is not a multiple of the vertex stride %d",76),g->Draws,len(g->inline_),stride);
return ;
}
nin = divi<int64_t>(len(g->inline_),stride);
g->vin = xbox_ensureVin(g->vin,nin);
{int64_t vi = cast<int64_t>(0ULL);for (;(vi < nin);vi++){
Slice<uint32_t> w = sub(g->inline_,cast<int64_t>((vi * stride)),cast<int64_t>(((cast<int64_t>((vi + cast<int64_t>(1ULL)))) * stride)));
std::array<std::array<float,4>,16>* in = (&g->vin[vi]);
gcopy(sub(in,0,len(in)),sub(g->vtxAttr,0,len(g->vtxAttr)));
{int64_t a = cast<int64_t>(0ULL);for (;(a < cast<int64_t>(16ULL));a++){
if ((fmts[a].words == cast<int64_t>(0ULL))) {
continue;
}
(*in)[a] = xbox_decodeAttr((&fmts[a]),sub(w,0,fmts[a].words));
w = sub(w,fmts[a].words,len(w));
}
}}
}}
else if (((len(g->elems) > cast<int64_t>(0ULL)) || (len(g->ranges) > cast<int64_t>(0ULL)))){
Slice<uint32_t> idx = g->elems;
{auto&& tmp323 = g->ranges;
for(int64_t tmp324=0;tmp324<len(tmp323);++tmp324){
auto r=tmp323[tmp324];{uint32_t k = cast<uint32_t>(0ULL);for (;(k < r[cast<int64_t>(1ULL)]);k++){
idx = append(idx,Slice<uint32_t>{cast<uint32_t>((r[cast<int64_t>(0ULL)] + k))});
}
}}}
nin = len(idx);
g->vin = xbox_ensureVin(g->vin,nin);
{auto&& tmp325 = idx;
for(int64_t tmp326=0;tmp326<len(tmp325);++tmp326){
auto i=tmp326;auto ix=tmp325[tmp326];auto tmp327 = xbox_pgraph_fetchArrayVertex(g,(&fmts),ix);
std::array<std::array<float,4>,16> in = std::get<0>(tmp327);
bool ok = std::get<1>(tmp327);
if ((!ok)) {
return ;
}
g->vin[i] = in;
}}
}
else {
return ;
}
}
tmp322:;
auto tmp328 = xbox_pgraph_transformVerts(g,nin,ff,trace);
Slice<xbox_kelvinVtx> verts = std::get<0>(tmp328);
bool ok = std::get<1>(tmp328);
if ((!ok)) {
return ;
}
xbox_Machine_profEnd(g->m,cast<int64_t>(0ULL),tv);
xbox_pgraph_assemble(g,verts);
}
}
}
}
// tools/platform/xbox/nv2a_vertex.go:286:1
Slice<std::array<std::array<float,4>,16>> xbox_ensureVin(Slice<std::array<std::array<float,4>,16>> buf,int64_t n){
{
if ((len(buf) < n)) {
return Slice<std::array<std::array<float,4>,16>>::make(n);
}
return sub(buf,0,n);
}
}
// tools/platform/xbox/nv2a_vertex.go:308:1
std::tuple<Slice<xbox_kelvinVtx>,bool> xbox_pgraph_transformVerts(xbox_pgraph* g,int64_t nin,bool ff,bool trace){
{
if ((len(g->vertsBuf) < nin)) {
g->vertsBuf = Slice<xbox_kelvinVtx>::make(nin);
}
Slice<xbox_kelvinVtx> verts = sub(g->vertsBuf,0,nin);
auto do_ = [&](int64_t i)->bool{
if (ff) {
verts[i] = xbox_pgraph_transformFF(g,(&g->vin[i]));
return true;
}
auto tmp329 = xbox_pgraph_transform(g,(&g->vin[i]),(trace && (i < cast<int64_t>(8ULL))));
xbox_kelvinVtx v = std::get<0>(tmp329);
bool ok = std::get<1>(tmp329);
if (ok) {
verts[i] = v;
}
return ok;
}
;
if (((((!ff) && (!trace)) && (!g->vshWritesConst)) && (nin >= cast<int64_t>(256ULL)))) {
if ((!do_(cast<int64_t>(0ULL)))) {
return {{},false};
}
xbox_parallelChunks(cast<int64_t>(1ULL),nin,xbox_vshWorkers(),[&](int64_t a,int64_t b)->void{
{int64_t i = a;for (;(i < b);i++){
do_(i);
}
}}
);
return {verts,true};
}
{int64_t i = cast<int64_t>(0ULL);for (;(i < nin);i++){
if ((!do_(i))) {
return {{},false};
}
}
}return {verts,true};
}
}
// tools/platform/xbox/nv2a_vertex.go:391:1
std::tuple<std::array<std::array<float,4>,16>,bool> xbox_pgraph_fetchArrayVertex(xbox_pgraph* g,std::array<xbox_vtxFmt,16>* fmts,uint32_t ix){
{
std::array<std::array<float,4>,16> in={};
gcopy(sub(in,0,len(in)),sub(g->vtxAttr,0,len(g->vtxAttr)));
{int64_t a = cast<int64_t>(0ULL);for (;(a < cast<int64_t>(16ULL));a++){
xbox_vtxFmt* vf = (&(*fmts)[a]);
if ((vf->words == cast<int64_t>(0ULL))) {
continue;
}
uint32_t off = g->Regs[cast<uint32_t>((cast<uint32_t>(1480ULL) + cast<uint32_t>(a)))];
uint32_t dmaReg = cast<uint32_t>(412ULL);
if ((shr<uint32_t>(off,cast<int64_t>(31ULL)) != cast<uint32_t>(0ULL))) {
dmaReg = cast<uint32_t>(416ULL);
}
auto tmp330 = xbox_Machine_dmaObjectTarget(g->m,g->Regs[shr<uint32_t>(dmaReg,cast<int64_t>(2ULL))]);
uint32_t base = std::get<0>(tmp330);
uint32_t limit = std::get<1>(tmp330);
uint32_t addr = cast<uint32_t>((cast<uint32_t>((base + (cast<uint32_t>((off & cast<uint32_t>(2147483647ULL)))))) + cast<uint32_t>((ix * vf->stride))));
if (((base == cast<uint32_t>(0ULL)) && (limit == cast<uint32_t>(0ULL)))) {
x86_CPU_Halt(g->m->CPU,std::string("nv2a: draw %d attr %d fetches through an unbound vertex DMA context",67),g->Draws,a);
return {in,false};
}
std::array<uint32_t,4> w={};
{int64_t k = cast<int64_t>(0ULL);for (;(k < vf->words);k++){
w[k] = xbox_Machine_read32(g->m,cast<uint32_t>((addr + cast<uint32_t>((cast<uint32_t>(k) * cast<uint32_t>(4ULL))))));
}
}in[a] = xbox_decodeAttr(vf,sub(w,0,vf->words));
}
}return {in,true};
}
}
// tools/platform/xbox/nv2a_vertex.go:426:1
xbox_kelvinVtx xbox_pgraph_transformFF(xbox_pgraph* g,std::array<std::array<float,4>,16>* in){
{
std::array<float,4> pos = (*in)[cast<int64_t>(0ULL)];
std::array<float,4> clip={};
uint32_t base = cast<uint32_t>(416ULL);
{int64_t j = cast<int64_t>(0ULL);for (;(j < cast<int64_t>(4ULL));j++){
float s={};
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(4ULL));i++){
s += (pos[i] * go_math_Float32frombits(g->Regs[cast<uint32_t>((base + cast<uint32_t>(cast<int64_t>((cast<int64_t>((i * cast<int64_t>(4ULL))) + j)))))]));
}
}clip[j] = s;
}
}if ((clip[cast<int64_t>(3ULL)] == cast<float>(0.00000000000000000e+00))) {
return xbox_kelvinVtx{};
}
std::function<float(uint32_t)> f = go_math_Float32frombits;
auto tmp331 = std::make_tuple((&g->Const[cast<int64_t>(58ULL)]),(&g->Const[cast<int64_t>(59ULL)]));
std::array<uint32_t,4>* sc = std::get<0>(tmp331);
std::array<uint32_t,4>* off = std::get<1>(tmp331);
xbox_kelvinVtx v = xbox_kelvinVtx{{},(*in)[cast<int64_t>(3ULL)],(*in)[cast<int64_t>(4ULL)],{},std::array<std::array<float,4>,4>{(*in)[cast<int64_t>(9ULL)],(*in)[cast<int64_t>(10ULL)],(*in)[cast<int64_t>(11ULL)],(*in)[cast<int64_t>(12ULL)]}};
v.pos[cast<int64_t>(0ULL)] = (((clip[cast<int64_t>(0ULL)] / clip[cast<int64_t>(3ULL)]) * f((*sc)[cast<int64_t>(0ULL)])) + f((*off)[cast<int64_t>(0ULL)]));
v.pos[cast<int64_t>(1ULL)] = (((clip[cast<int64_t>(1ULL)] / clip[cast<int64_t>(3ULL)]) * f((*sc)[cast<int64_t>(1ULL)])) + f((*off)[cast<int64_t>(1ULL)]));
v.pos[cast<int64_t>(2ULL)] = (((clip[cast<int64_t>(2ULL)] / clip[cast<int64_t>(3ULL)]) * f((*sc)[cast<int64_t>(2ULL)])) + f((*off)[cast<int64_t>(2ULL)]));
v.pos[cast<int64_t>(3ULL)] = clip[cast<int64_t>(3ULL)];
return v;
}
}
// tools/platform/xbox/nv2a_vertex.go:454:1
std::tuple<xbox_kelvinVtx,bool> xbox_pgraph_transform(xbox_pgraph* g,std::array<std::array<float,4>,16>* in,bool trace){
{
std::array<std::array<float,4>,13> out={};
if ((!xbox_pgraph_vshRun(g,in,(&out)))) {
return {xbox_kelvinVtx{},false};
}
xbox_kelvinVtx v = xbox_kelvinVtx{out[cast<int64_t>(0ULL)],out[cast<int64_t>(3ULL)],out[cast<int64_t>(4ULL)],cast<float>(out[cast<int64_t>(5ULL)][cast<int64_t>(0ULL)]),std::array<std::array<float,4>,4>{out[cast<int64_t>(9ULL)],out[cast<int64_t>(10ULL)],out[cast<int64_t>(11ULL)],out[cast<int64_t>(12ULL)]}};
if (trace) {
go_fmt_Printf(std::string("  vtx in v0=%v v3=%v v9=%v -> pos=%v d0=%v uv0=%v\012",50),(*in)[cast<int64_t>(0ULL)],(*in)[cast<int64_t>(3ULL)],(*in)[cast<int64_t>(9ULL)],v.pos,v.d0,v.uv[cast<int64_t>(0ULL)]);
}
return {v,true};
}
}
// tools/platform/xbox/nv2a_vertex.go:485:1
void xbox_pgraph_assemble(xbox_pgraph* g,Slice<xbox_kelvinVtx> verts){
{rrprof::Scope timing(3,"Software rasterizer");
{
Slice<std::array<int64_t,3>> tris = sub(g->triScratch,0,cast<int64_t>(0ULL));
auto tri = [&](int64_t a,int64_t b,int64_t c)->void{
tris = append(tris,Slice<std::array<int64_t,3>>{std::array<int64_t,3>{a,b,c}});
}
;
int64_t n = len(verts);
{
switch(g->prim){
case cast<uint32_t>(5ULL):{
{int64_t i = cast<int64_t>(0ULL);for (;(cast<int64_t>((i + cast<int64_t>(2ULL))) < n);i += cast<int64_t>(3ULL)){
tri(i,cast<int64_t>((i + cast<int64_t>(1ULL))),cast<int64_t>((i + cast<int64_t>(2ULL))));
}
}break;}
case cast<uint32_t>(6ULL):{
{int64_t i = cast<int64_t>(0ULL);for (;(cast<int64_t>((i + cast<int64_t>(2ULL))) < n);i++){
if ((cast<int64_t>((i & cast<int64_t>(1ULL))) == cast<int64_t>(0ULL))) {
tri(i,cast<int64_t>((i + cast<int64_t>(1ULL))),cast<int64_t>((i + cast<int64_t>(2ULL))));
}
else {
tri(cast<int64_t>((i + cast<int64_t>(1ULL))),i,cast<int64_t>((i + cast<int64_t>(2ULL))));
}
}
}break;}
case cast<uint32_t>(7ULL):case cast<uint32_t>(10ULL):{
{int64_t i = cast<int64_t>(1ULL);for (;(cast<int64_t>((i + cast<int64_t>(1ULL))) < n);i++){
tri(cast<int64_t>(0ULL),i,cast<int64_t>((i + cast<int64_t>(1ULL))));
}
}break;}
case cast<uint32_t>(8ULL):{
{int64_t i = cast<int64_t>(0ULL);for (;(cast<int64_t>((i + cast<int64_t>(3ULL))) < n);i += cast<int64_t>(4ULL)){
tri(i,cast<int64_t>((i + cast<int64_t>(1ULL))),cast<int64_t>((i + cast<int64_t>(2ULL))));
tri(i,cast<int64_t>((i + cast<int64_t>(2ULL))),cast<int64_t>((i + cast<int64_t>(3ULL))));
}
}break;}
case cast<uint32_t>(9ULL):{
{int64_t i = cast<int64_t>(0ULL);for (;(cast<int64_t>((i + cast<int64_t>(3ULL))) < n);i += cast<int64_t>(2ULL)){
tri(i,cast<int64_t>((i + cast<int64_t>(1ULL))),cast<int64_t>((i + cast<int64_t>(3ULL))));
tri(i,cast<int64_t>((i + cast<int64_t>(3ULL))),cast<int64_t>((i + cast<int64_t>(2ULL))));
}
}break;}
case cast<uint32_t>(1ULL):{
break;}
default:{
x86_CPU_Halt(g->m->CPU,std::string("nv2a: draw %d primitive type %d not implemented",47),g->Draws,g->prim);
return ;
break;}
}}
g->triScratch = tris;
if ((len(tris) == cast<int64_t>(0ULL))) {
return ;
}
auto tmp332 = xbox_pgraph_clipNearPlane(g,verts,tris);
verts = std::get<0>(tmp332);
tris = std::get<1>(tmp332);
if ((len(tris) == cast<int64_t>(0ULL))) {
return ;
}
if ((!g->rastValid)) {
if ((!xbox_pgraph_rasterStateDecode(g,(&g->rast)))) {
return ;
}
g->rastValid = true;
}
time_Time tr = xbox_Machine_profStart(g->m);
if (xbox_pgraph_rasterParallelOK(g,(&g->rast),verts,tris)) {
xbox_pgraph_rasterParallel(g,verts,tris);
xbox_Machine_profEnd(g->m,cast<int64_t>(1ULL),tr);
return ;
}
xbox_rstats rs={};
{auto&& tmp333 = tris;
for(int64_t tmp334=0;tmp334<len(tmp333);++tmp334){
auto t=tmp333[tmp334];xbox_pgraph_rasterTri(g,(&verts[t[cast<int64_t>(0ULL)]]),(&verts[t[cast<int64_t>(1ULL)]]),(&verts[t[cast<int64_t>(2ULL)]]),(&rs));
if (g->m->CPU->Halted) {
break;
}
}}
xbox_pgraph_mergeStats(g,(&rs));
xbox_Machine_profEnd(g->m,cast<int64_t>(1ULL),tr);
}
}
}
// tools/platform/xbox/nv2a_vertex.go:567:1
void xbox_pgraph_traceDraw(xbox_pgraph* g,std::array<xbox_vtxFmt,16>* fmts,int64_t stride){
{
go_fmt_Printf(std::string("nv2a draw %d: prim=%d inline=%d words elems=%d ranges=%d stride=%d\012",67),g->Draws,g->prim,len(g->inline_),len(g->elems),len(g->ranges),stride);
{int64_t a = cast<int64_t>(0ULL);for (;(a < cast<int64_t>(16ULL));a++){
if (((*fmts)[a].words != cast<int64_t>(0ULL))) {
go_fmt_Printf(std::string("  attr%-2d type=%d size=%d stride=%d words=%d\012",46),a,(*fmts)[a].typ,(*fmts)[a].size,(*fmts)[a].stride,(*fmts)[a].words);
}
}
}go_fmt_Printf(std::string("  progStart=%d constLoadAt=%d execMode=%08X surf fmt=%08X pitch=%08X color=%08X zeta=%08X clipH=%08X clipV=%08X\012",112),g->Regs[cast<int64_t>(1960ULL)],g->ConstLoad,g->Regs[cast<int64_t>(1957ULL)],g->Regs[cast<int64_t>(130ULL)],g->Regs[cast<int64_t>(131ULL)],g->Regs[cast<int64_t>(132ULL)],g->Regs[cast<int64_t>(133ULL)],g->Regs[cast<int64_t>(128ULL)],g->Regs[cast<int64_t>(129ULL)]);
go_fmt_Printf(std::string("  blend en=%d sf=%03X df=%03X eq=%04X  alpha en=%d func=%03X ref=%d  depth en=%d func=%03X wr=%d  cull en=%d face=%03X front=%03X  combiners=%d shader=%08X\012",156),g->Regs[cast<int64_t>(193ULL)],g->Regs[cast<int64_t>(209ULL)],g->Regs[cast<int64_t>(210ULL)],g->Regs[cast<int64_t>(212ULL)],g->Regs[cast<int64_t>(192ULL)],g->Regs[cast<int64_t>(207ULL)],g->Regs[cast<int64_t>(208ULL)],g->Regs[cast<int64_t>(195ULL)],g->Regs[cast<int64_t>(213ULL)],g->Regs[cast<int64_t>(215ULL)],g->Regs[cast<int64_t>(194ULL)],g->Regs[cast<int64_t>(231ULL)],g->Regs[cast<int64_t>(232ULL)],cast<uint32_t>((g->Regs[cast<int64_t>(1944ULL)] & cast<uint32_t>(255ULL))),g->Regs[cast<int64_t>(1948ULL)]);
{int64_t u = cast<int64_t>(0ULL);for (;(u < cast<int64_t>(4ULL));u++){
uint32_t ctl = g->Regs[shr<uint32_t>((cast<uint32_t>((cast<uint32_t>(6924ULL) + cast<uint32_t>((cast<uint32_t>(u) * cast<uint32_t>(64ULL)))))),cast<int64_t>(2ULL))];
if ((cast<uint32_t>((shr<uint32_t>(ctl,cast<int64_t>(30ULL)) & cast<uint32_t>(1ULL))) == cast<uint32_t>(0ULL))) {
continue;
}
go_fmt_Printf(std::string("  tex%d off=%08X fmt=%08X addr=%08X ctl0=%08X filt=%08X rect=%08X\012",66),u,g->Regs[shr<uint32_t>((cast<uint32_t>((cast<uint32_t>(6912ULL) + cast<uint32_t>((cast<uint32_t>(u) * cast<uint32_t>(64ULL)))))),cast<int64_t>(2ULL))],g->Regs[shr<uint32_t>((cast<uint32_t>((cast<uint32_t>(6916ULL) + cast<uint32_t>((cast<uint32_t>(u) * cast<uint32_t>(64ULL)))))),cast<int64_t>(2ULL))],g->Regs[shr<uint32_t>((cast<uint32_t>((cast<uint32_t>(6920ULL) + cast<uint32_t>((cast<uint32_t>(u) * cast<uint32_t>(64ULL)))))),cast<int64_t>(2ULL))],ctl,g->Regs[shr<uint32_t>((cast<uint32_t>((cast<uint32_t>(6932ULL) + cast<uint32_t>((cast<uint32_t>(u) * cast<uint32_t>(64ULL)))))),cast<int64_t>(2ULL))],g->Regs[shr<uint32_t>((cast<uint32_t>((cast<uint32_t>(6940ULL) + cast<uint32_t>((cast<uint32_t>(u) * cast<uint32_t>(64ULL)))))),cast<int64_t>(2ULL))]);
}
}if (xbox_nvVSTrace) {
int64_t pc = modi<int64_t>(cast<int64_t>(g->Regs[cast<int64_t>(1960ULL)]),cast<int64_t>(136ULL));
{int64_t k = cast<int64_t>(0ULL);for (;(k < cast<int64_t>(136ULL));k++){
int64_t i = modi<int64_t>((cast<int64_t>((pc + k))),cast<int64_t>(136ULL));
go_fmt_Printf(std::string("  vsh[%3d] %s\012",14),i,xbox_pgraph_vshDisasm(g,i));
if (xbox_pgraph_vshDecode(g,i).final) {
break;
}
}
}}
}
}
// tools/platform/xbox/nv2a_vsh.go:125:1
uint32_t xbox_vshField(std::array<uint32_t,4>* w,int64_t dword,int64_t lo,int64_t n){
{
return cast<uint32_t>((shr<uint32_t>((*w)[dword],lo) & (cast<uint32_t>((shl<uint32_t>(cast<uint32_t>(1ULL),cast<uint64_t>(n)) - cast<uint32_t>(1ULL))))));
}
}
// tools/platform/xbox/nv2a_vsh.go:130:1
xbox_vshInst xbox_pgraph_vshDecode(xbox_pgraph* g,int64_t i){
{
std::array<uint32_t,4>* w = (&g->Prog[i]);
auto swz = [&](int64_t dword,int64_t lo)->std::array<int64_t,4>{
uint32_t v = xbox_vshField(w,dword,lo,cast<int64_t>(8ULL));
return std::array<int64_t,4>{cast<int64_t>(cast<uint32_t>((shr<uint32_t>(v,cast<int64_t>(6ULL)) & cast<uint32_t>(3ULL)))),cast<int64_t>(cast<uint32_t>((shr<uint32_t>(v,cast<int64_t>(4ULL)) & cast<uint32_t>(3ULL)))),cast<int64_t>(cast<uint32_t>((shr<uint32_t>(v,cast<int64_t>(2ULL)) & cast<uint32_t>(3ULL)))),cast<int64_t>(cast<uint32_t>((v & cast<uint32_t>(3ULL))))};
}
;
return xbox_vshInst{cast<int64_t>(xbox_vshField(w,cast<int64_t>(1ULL),cast<int64_t>(21ULL),cast<int64_t>(4ULL))),cast<int64_t>(xbox_vshField(w,cast<int64_t>(1ULL),cast<int64_t>(25ULL),cast<int64_t>(3ULL))),cast<int64_t>(xbox_vshField(w,cast<int64_t>(1ULL),cast<int64_t>(13ULL),cast<int64_t>(8ULL))),cast<int64_t>(xbox_vshField(w,cast<int64_t>(1ULL),cast<int64_t>(9ULL),cast<int64_t>(4ULL))),xbox_vshSrc{cast<int64_t>(xbox_vshField(w,cast<int64_t>(2ULL),cast<int64_t>(26ULL),cast<int64_t>(2ULL))),cast<int64_t>(xbox_vshField(w,cast<int64_t>(2ULL),cast<int64_t>(28ULL),cast<int64_t>(4ULL))),(xbox_vshField(w,cast<int64_t>(1ULL),cast<int64_t>(8ULL),cast<int64_t>(1ULL)) != cast<uint32_t>(0ULL)),swz(cast<int64_t>(1ULL),cast<int64_t>(0ULL))},xbox_vshSrc{cast<int64_t>(xbox_vshField(w,cast<int64_t>(2ULL),cast<int64_t>(11ULL),cast<int64_t>(2ULL))),cast<int64_t>(xbox_vshField(w,cast<int64_t>(2ULL),cast<int64_t>(13ULL),cast<int64_t>(4ULL))),(xbox_vshField(w,cast<int64_t>(2ULL),cast<int64_t>(25ULL),cast<int64_t>(1ULL)) != cast<uint32_t>(0ULL)),swz(cast<int64_t>(2ULL),cast<int64_t>(17ULL))},xbox_vshSrc{cast<int64_t>(xbox_vshField(w,cast<int64_t>(3ULL),cast<int64_t>(28ULL),cast<int64_t>(2ULL))),cast<int64_t>(cast<uint32_t>((shl<uint32_t>(xbox_vshField(w,cast<int64_t>(2ULL),cast<int64_t>(0ULL),cast<int64_t>(2ULL)),cast<int64_t>(2ULL)) | xbox_vshField(w,cast<int64_t>(3ULL),cast<int64_t>(30ULL),cast<int64_t>(2ULL))))),(xbox_vshField(w,cast<int64_t>(2ULL),cast<int64_t>(10ULL),cast<int64_t>(1ULL)) != cast<uint32_t>(0ULL)),swz(cast<int64_t>(2ULL),cast<int64_t>(2ULL))},xbox_vshField(w,cast<int64_t>(3ULL),cast<int64_t>(24ULL),cast<int64_t>(4ULL)),cast<int64_t>(xbox_vshField(w,cast<int64_t>(3ULL),cast<int64_t>(20ULL),cast<int64_t>(4ULL))),xbox_vshField(w,cast<int64_t>(3ULL),cast<int64_t>(16ULL),cast<int64_t>(4ULL)),xbox_vshField(w,cast<int64_t>(3ULL),cast<int64_t>(12ULL),cast<int64_t>(4ULL)),(xbox_vshField(w,cast<int64_t>(3ULL),cast<int64_t>(11ULL),cast<int64_t>(1ULL)) != cast<uint32_t>(0ULL)),cast<int64_t>(xbox_vshField(w,cast<int64_t>(3ULL),cast<int64_t>(3ULL),cast<int64_t>(8ULL))),(xbox_vshField(w,cast<int64_t>(3ULL),cast<int64_t>(2ULL),cast<int64_t>(1ULL)) != cast<uint32_t>(0ULL)),(xbox_vshField(w,cast<int64_t>(3ULL),cast<int64_t>(1ULL),cast<int64_t>(1ULL)) != cast<uint32_t>(0ULL)),(xbox_vshField(w,cast<int64_t>(3ULL),cast<int64_t>(0ULL),cast<int64_t>(1ULL)) != cast<uint32_t>(0ULL))};
}
}
// tools/platform/xbox/nv2a_vsh.go:182:1
std::array<float,4> xbox_vshState_src(xbox_vshState* s,xbox_vshInst* inst,xbox_vshSrc* o){
{
std::array<float,4> base={};
{
switch(o->mux){
case cast<int64_t>(1ULL):{
base = s->r[modi<int64_t>(cast<int64_t>((o->reg & cast<int64_t>(15ULL))),cast<int64_t>(13ULL))];
break;}
case cast<int64_t>(2ULL):{
base = (*s->v)[inst->inputReg];
break;}
case cast<int64_t>(3ULL):{
int64_t c = inst->constIdx;
if (inst->relConst) {
c += cast<int64_t>(s->a);
}
if (((c >= cast<int64_t>(0ULL)) && (c < cast<int64_t>(192ULL)))) {
base = xbox_f32vec((&s->g->Const[c]));
}
break;}
default:{
break;}
}}
std::array<float,4> out = std::array<float,4>{base[o->swz[cast<int64_t>(0ULL)]],base[o->swz[cast<int64_t>(1ULL)]],base[o->swz[cast<int64_t>(2ULL)]],base[o->swz[cast<int64_t>(3ULL)]]};
if (o->neg) {
auto tmp335 = std::make_tuple(cast<float>(-out[cast<int64_t>(0ULL)]),cast<float>(-out[cast<int64_t>(1ULL)]),cast<float>(-out[cast<int64_t>(2ULL)]),cast<float>(-out[cast<int64_t>(3ULL)]));
out[cast<int64_t>(0ULL)] = std::get<0>(tmp335);
out[cast<int64_t>(1ULL)] = std::get<1>(tmp335);
out[cast<int64_t>(2ULL)] = std::get<2>(tmp335);
out[cast<int64_t>(3ULL)] = std::get<3>(tmp335);
}
return out;
}
}
// tools/platform/xbox/nv2a_vsh.go:209:1
std::array<float,4> xbox_f32vec(std::array<uint32_t,4>* c){
{
return std::array<float,4>{go_math_Float32frombits((*c)[cast<int64_t>(0ULL)]),go_math_Float32frombits((*c)[cast<int64_t>(1ULL)]),go_math_Float32frombits((*c)[cast<int64_t>(2ULL)]),go_math_Float32frombits((*c)[cast<int64_t>(3ULL)])};
}
}
// tools/platform/xbox/nv2a_vsh.go:216:1
void xbox_maskWrite(std::array<float,4>* dst,std::array<float,4> val,uint32_t mask){
{
if ((cast<uint32_t>((mask & cast<uint32_t>(8ULL))) != cast<uint32_t>(0ULL))) {
(*dst)[cast<int64_t>(0ULL)] = val[cast<int64_t>(0ULL)];
}
if ((cast<uint32_t>((mask & cast<uint32_t>(4ULL))) != cast<uint32_t>(0ULL))) {
(*dst)[cast<int64_t>(1ULL)] = val[cast<int64_t>(1ULL)];
}
if ((cast<uint32_t>((mask & cast<uint32_t>(2ULL))) != cast<uint32_t>(0ULL))) {
(*dst)[cast<int64_t>(2ULL)] = val[cast<int64_t>(2ULL)];
}
if ((cast<uint32_t>((mask & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL))) {
(*dst)[cast<int64_t>(3ULL)] = val[cast<int64_t>(3ULL)];
}
}
}
// tools/platform/xbox/nv2a_vsh.go:240:1
bool xbox_pgraph_vshCompile(xbox_pgraph* g){
{
g->vshProg = sub(g->vshProg,0,cast<int64_t>(0ULL));
g->vshWritesConst = false;
int64_t pc = modi<int64_t>(cast<int64_t>(g->Regs[cast<int64_t>(1960ULL)]),cast<int64_t>(136ULL));
{int64_t steps = cast<int64_t>(0ULL);for (;;steps++){
if ((steps >= cast<int64_t>(136ULL))) {
x86_CPU_Halt(g->m->CPU,std::string("nv2a vsh: no FINAL instruction within %d slots from start %d",60),cast<int64_t>(136ULL),g->Regs[cast<int64_t>(1960ULL)]);
return false;
}
xbox_vshInst inst = xbox_pgraph_vshDecode(g,pc);
g->vshProg = append(g->vshProg,Slice<xbox_vshInst>{inst});
if (((inst.outMask != cast<uint32_t>(0ULL)) && (!inst.outIsO))) {
g->vshWritesConst = true;
}
pc = modi<int64_t>((cast<int64_t>((pc + cast<int64_t>(1ULL)))),cast<int64_t>(136ULL));
if (inst.final) {
return true;
}
}
}}
}
// tools/platform/xbox/nv2a_vsh.go:261:1
bool xbox_pgraph_vshRun(xbox_pgraph* g,std::array<std::array<float,4>,16>* v,std::array<std::array<float,4>,13>* out){
{
xbox_vshState s = xbox_vshState{g,v,{},{},{}};
{auto&& tmp336 = g->vshProg;
for(int64_t tmp337=0;tmp337<len(tmp336);++tmp337){
auto pi=tmp337;xbox_vshInst inst = g->vshProg[pi];
std::array<float,4> macRes={};
if ((inst.mac != cast<int64_t>(0ULL))) {
auto tmp338 = xbox_vshState_src(&(s),(&inst),(&inst.a));
std::array<float,4> a = tmp338;
{
switch(inst.mac){
case cast<int64_t>(1ULL):{
macRes = a;
break;}
case cast<int64_t>(2ULL):{
auto tmp339 = xbox_vshState_src(&(s),(&inst),(&inst.b));
std::array<float,4> b = tmp339;
{auto&& tmp340 = macRes;
for(int64_t tmp341=0;tmp341<len(tmp340);++tmp341){
auto i=tmp341;macRes[i] = (a[i] * b[i]);
}}
break;}
case cast<int64_t>(3ULL):{
auto tmp342 = xbox_vshState_src(&(s),(&inst),(&inst.c));
std::array<float,4> c = tmp342;
{auto&& tmp343 = macRes;
for(int64_t tmp344=0;tmp344<len(tmp343);++tmp344){
auto i=tmp344;macRes[i] = (a[i] + c[i]);
}}
break;}
case cast<int64_t>(4ULL):{
auto tmp345 = std::make_tuple(xbox_vshState_src(&(s),(&inst),(&inst.b)),xbox_vshState_src(&(s),(&inst),(&inst.c)));
std::array<float,4> b = std::get<0>(tmp345);
std::array<float,4> c = std::get<1>(tmp345);
{auto&& tmp346 = macRes;
for(int64_t tmp347=0;tmp347<len(tmp346);++tmp347){
auto i=tmp347;macRes[i] = ((a[i] * b[i]) + c[i]);
}}
break;}
case cast<int64_t>(5ULL):{
auto tmp348 = xbox_vshState_src(&(s),(&inst),(&inst.b));
std::array<float,4> b = tmp348;
float dp = (((a[cast<int64_t>(0ULL)] * b[cast<int64_t>(0ULL)]) + (a[cast<int64_t>(1ULL)] * b[cast<int64_t>(1ULL)])) + (a[cast<int64_t>(2ULL)] * b[cast<int64_t>(2ULL)]));
macRes = std::array<float,4>{dp,dp,dp,dp};
break;}
case cast<int64_t>(6ULL):{
auto tmp349 = xbox_vshState_src(&(s),(&inst),(&inst.b));
std::array<float,4> b = tmp349;
float dp = ((((a[cast<int64_t>(0ULL)] * b[cast<int64_t>(0ULL)]) + (a[cast<int64_t>(1ULL)] * b[cast<int64_t>(1ULL)])) + (a[cast<int64_t>(2ULL)] * b[cast<int64_t>(2ULL)])) + b[cast<int64_t>(3ULL)]);
macRes = std::array<float,4>{dp,dp,dp,dp};
break;}
case cast<int64_t>(7ULL):{
auto tmp350 = xbox_vshState_src(&(s),(&inst),(&inst.b));
std::array<float,4> b = tmp350;
float dp = ((((a[cast<int64_t>(0ULL)] * b[cast<int64_t>(0ULL)]) + (a[cast<int64_t>(1ULL)] * b[cast<int64_t>(1ULL)])) + (a[cast<int64_t>(2ULL)] * b[cast<int64_t>(2ULL)])) + (a[cast<int64_t>(3ULL)] * b[cast<int64_t>(3ULL)]));
macRes = std::array<float,4>{dp,dp,dp,dp};
break;}
case cast<int64_t>(8ULL):{
auto tmp351 = xbox_vshState_src(&(s),(&inst),(&inst.b));
std::array<float,4> b = tmp351;
macRes = std::array<float,4>{cast<float>(1.00000000000000000e+00),(a[cast<int64_t>(1ULL)] * b[cast<int64_t>(1ULL)]),a[cast<int64_t>(2ULL)],b[cast<int64_t>(3ULL)]};
break;}
case cast<int64_t>(9ULL):{
auto tmp352 = xbox_vshState_src(&(s),(&inst),(&inst.b));
std::array<float,4> b = tmp352;
{auto&& tmp353 = macRes;
for(int64_t tmp354=0;tmp354<len(tmp353);++tmp354){
auto i=tmp354;macRes[i] = xbox_minf32(a[i],b[i]);
}}
break;}
case cast<int64_t>(10ULL):{
auto tmp355 = xbox_vshState_src(&(s),(&inst),(&inst.b));
std::array<float,4> b = tmp355;
{auto&& tmp356 = macRes;
for(int64_t tmp357=0;tmp357<len(tmp356);++tmp357){
auto i=tmp357;macRes[i] = xbox_maxf32(a[i],b[i]);
}}
break;}
case cast<int64_t>(11ULL):{
auto tmp358 = xbox_vshState_src(&(s),(&inst),(&inst.b));
std::array<float,4> b = tmp358;
{auto&& tmp359 = macRes;
for(int64_t tmp360=0;tmp360<len(tmp359);++tmp360){
auto i=tmp360;if ((a[i] < b[i])) {
macRes[i] = cast<float>(1.00000000000000000e+00);
}
}}
break;}
case cast<int64_t>(12ULL):{
auto tmp361 = xbox_vshState_src(&(s),(&inst),(&inst.b));
std::array<float,4> b = tmp361;
{auto&& tmp362 = macRes;
for(int64_t tmp363=0;tmp363<len(tmp362);++tmp363){
auto i=tmp363;if ((a[i] >= b[i])) {
macRes[i] = cast<float>(1.00000000000000000e+00);
}
}}
break;}
case cast<int64_t>(13ULL):{
s.a = cast<int32_t>(xbox_floorf32(a[cast<int64_t>(0ULL)]));
break;}
default:{
x86_CPU_Halt(g->m->CPU,std::string("nv2a vsh: MAC op %d unimplemented",33),inst.mac);
return false;
break;}
}}
if (((inst.mac != cast<int64_t>(13ULL)) && (inst.macMask != cast<uint32_t>(0ULL)))) {
xbox_maskWrite((&s.r[modi<int64_t>(inst.macDst,cast<int64_t>(13ULL))]),macRes,inst.macMask);
}
}
std::array<float,4> iluRes={};
if ((inst.ilu != cast<int64_t>(0ULL))) {
auto tmp364 = xbox_vshState_src(&(s),(&inst),(&inst.c));
std::array<float,4> c = tmp364;
{
switch(inst.ilu){
case cast<int64_t>(1ULL):{
iluRes = c;
break;}
case cast<int64_t>(2ULL):{
float r = xbox_rcpf32(c[cast<int64_t>(0ULL)]);
iluRes = std::array<float,4>{r,r,r,r};
break;}
case cast<int64_t>(3ULL):{
float r = xbox_rccf32(c[cast<int64_t>(0ULL)]);
iluRes = std::array<float,4>{r,r,r,r};
break;}
case cast<int64_t>(4ULL):{
float r = xbox_rsqf32(c[cast<int64_t>(0ULL)]);
iluRes = std::array<float,4>{r,r,r,r};
break;}
case cast<int64_t>(5ULL):{
float e = xbox_floorf32(c[cast<int64_t>(0ULL)]);
iluRes = std::array<float,4>{xbox_exp2f32(e),(c[cast<int64_t>(0ULL)] - e),xbox_exp2f32(c[cast<int64_t>(0ULL)]),cast<float>(1.00000000000000000e+00)};
break;}
case cast<int64_t>(6ULL):{
iluRes = xbox_log2partial(c[cast<int64_t>(0ULL)]);
break;}
case cast<int64_t>(7ULL):{
iluRes = xbox_litf32(c);
break;}
default:{
x86_CPU_Halt(g->m->CPU,std::string("nv2a vsh: ILU op %d unimplemented",33),inst.ilu);
return false;
break;}
}}
if ((inst.iluMask != cast<uint32_t>(0ULL))) {
xbox_maskWrite((&s.r[cast<int64_t>(1ULL)]),iluRes,inst.iluMask);
}
}
if ((inst.outMask != cast<uint32_t>(0ULL))) {
std::array<float,4> val = macRes;
if (inst.outFromILU) {
val = iluRes;
}
if (inst.outIsO) {
int64_t oreg = cast<int64_t>((inst.outAddr & cast<int64_t>(15ULL)));
if ((oreg < cast<int64_t>(13ULL))) {
xbox_maskWrite((&s.o[oreg]),val,inst.outMask);
if ((oreg == cast<int64_t>(0ULL))) {
xbox_maskWrite((&s.r[cast<int64_t>(12ULL)]),val,inst.outMask);
}
}
}
else if (((g->Regs[cast<int64_t>(1958ULL)] != cast<uint32_t>(0ULL)) && (inst.outAddr < cast<int64_t>(192ULL)))) {
std::array<uint32_t,4> bits={};
{auto&& tmp365 = val;
for(int64_t tmp366=0;tmp366<len(tmp365);++tmp366){
auto i=tmp366;auto f=tmp365[tmp366];bits[i] = go_math_Float32bits(f);
}}
xbox_maskWrite32((&g->Const[inst.outAddr]),bits,inst.outMask);
}
}
if (inst.final) {
break;
}
}}
s.o[cast<int64_t>(0ULL)] = s.r[cast<int64_t>(12ULL)];
(*out) = s.o;
return true;
}
}
// tools/platform/xbox/nv2a_vsh.go:402:1
void xbox_maskWrite32(std::array<uint32_t,4>* dst,std::array<uint32_t,4> val,uint32_t mask){
{
if ((cast<uint32_t>((mask & cast<uint32_t>(8ULL))) != cast<uint32_t>(0ULL))) {
(*dst)[cast<int64_t>(0ULL)] = val[cast<int64_t>(0ULL)];
}
if ((cast<uint32_t>((mask & cast<uint32_t>(4ULL))) != cast<uint32_t>(0ULL))) {
(*dst)[cast<int64_t>(1ULL)] = val[cast<int64_t>(1ULL)];
}
if ((cast<uint32_t>((mask & cast<uint32_t>(2ULL))) != cast<uint32_t>(0ULL))) {
(*dst)[cast<int64_t>(2ULL)] = val[cast<int64_t>(2ULL)];
}
if ((cast<uint32_t>((mask & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL))) {
(*dst)[cast<int64_t>(3ULL)] = val[cast<int64_t>(3ULL)];
}
}
}
// tools/platform/xbox/nv2a_vsh.go:422:1
void xbox_pgraph_progData(xbox_pgraph* g,uint32_t arg){
{
g->progBuf[g->progBufN] = arg;
g->progBufN++;
if ((g->progBufN < cast<int64_t>(4ULL))) {
return ;
}
g->progBufN = cast<int64_t>(0ULL);
uint32_t slot = modi<uint32_t>(g->ProgLoad,cast<uint32_t>(136ULL));
g->Prog[slot] = g->progBuf;
g->ProgLoad = cast<uint32_t>((slot + cast<uint32_t>(1ULL)));
if (xbox_nvVSTrace) {
go_fmt_Printf(std::string("VSH prog[%3d] = %08X %08X %08X %08X  %s\012",40),slot,g->progBuf[cast<int64_t>(0ULL)],g->progBuf[cast<int64_t>(1ULL)],g->progBuf[cast<int64_t>(2ULL)],g->progBuf[cast<int64_t>(3ULL)],xbox_pgraph_vshDisasm(g,cast<int64_t>(slot)));
}
}
}
// tools/platform/xbox/nv2a_vsh.go:440:1
void xbox_pgraph_constData(xbox_pgraph* g,uint32_t arg){
{
g->constBuf[g->constBufN] = arg;
g->constBufN++;
if ((g->constBufN < cast<int64_t>(4ULL))) {
return ;
}
g->constBufN = cast<int64_t>(0ULL);
uint32_t slot = modi<uint32_t>(g->ConstLoad,cast<uint32_t>(192ULL));
g->Const[slot] = g->constBuf;
g->ConstLoad = cast<uint32_t>((slot + cast<uint32_t>(1ULL)));
if (xbox_nvVSTrace) {
std::array<float,4> v = xbox_f32vec((&g->Const[slot]));
go_fmt_Printf(std::string("VSH const[%3d] = (%g, %g, %g, %g)\012",34),slot,v[cast<int64_t>(0ULL)],v[cast<int64_t>(1ULL)],v[cast<int64_t>(2ULL)],v[cast<int64_t>(3ULL)]);
}
if (((xbox_nvVPTrace && (slot >= cast<uint32_t>(56ULL))) && (slot <= cast<uint32_t>(60ULL)))) {
std::array<float,4> v = xbox_f32vec((&g->Const[slot]));
go_fmt_Printf(std::string("VP const-load c%d = (%g, %g, %g, %g) draws=%d\012",46),slot,v[cast<int64_t>(0ULL)],v[cast<int64_t>(1ULL)],v[cast<int64_t>(2ULL)],v[cast<int64_t>(3ULL)],g->Draws);
}
}
}
// tools/platform/xbox/nv2a_vsh.go:462:1
std::string xbox_pgraph_vshDisasm(xbox_pgraph* g,int64_t i){
{
xbox_vshInst in = xbox_pgraph_vshDecode(g,i);
auto srcStr = [&](xbox_vshSrc* o)->std::string{
std::string name={};
{
switch(o->mux){
case cast<int64_t>(1ULL):{
name = go_fmt_Sprintf(std::string("R%d",3),modi<int64_t>(o->reg,cast<int64_t>(13ULL)));
break;}
case cast<int64_t>(2ULL):{
name = go_fmt_Sprintf(std::string("v%d",3),in.inputReg);
break;}
case cast<int64_t>(3ULL):{
if (in.relConst) {
name = go_fmt_Sprintf(std::string("c[a0+%d]",8),in.constIdx);
}
else {
name = go_fmt_Sprintf(std::string("c%d",3),in.constIdx);
}
break;}
default:{
name = std::string("?",1);
break;}
}}
std::string sw = std::string("",0);
if ((o->swz != std::array<int64_t,4>{cast<int64_t>(0ULL),cast<int64_t>(1ULL),cast<int64_t>(2ULL),cast<int64_t>(3ULL)})) {
std::string comp = std::string("xyzw",4);
sw = (std::string(".",1) + cast<std::string>(Slice<uint8_t>{cast<uint8_t>(comp[o->swz[cast<int64_t>(0ULL)]]),cast<uint8_t>(comp[o->swz[cast<int64_t>(1ULL)]]),cast<uint8_t>(comp[o->swz[cast<int64_t>(2ULL)]]),cast<uint8_t>(comp[o->swz[cast<int64_t>(3ULL)]])}));
}
if (o->neg) {
return ((std::string("-",1) + name) + sw);
}
return (name + sw);
}
;
auto maskStr = [&](uint32_t m)->std::string{
if ((m == cast<uint32_t>(15ULL))) {
return std::string("",0);
}
std::string s = std::string(".",1);
{auto&& tmp367 = cast<Slice<uint8_t>>(std::string("xyzw",4));
for(int64_t tmp368=0;tmp368<len(tmp367);++tmp368){
auto b=tmp368;auto c=tmp367[tmp368];if ((cast<uint32_t>((m & (shr<uint32_t>(cast<uint32_t>(8ULL),cast<uint64_t>(b))))) != cast<uint32_t>(0ULL))) {
s += cast<std::string>(c);
}
}}
return s;
}
;
Slice<std::string> parts={};
if ((in.mac != cast<int64_t>(0ULL))) {
std::string p = xbox_macNames[in.mac];
if ((in.mac == cast<int64_t>(13ULL))) {
p += (std::string(" a0.x, ",7) + srcStr((&in.a)));
}
else {
Slice<std::string> dsts = Slice<std::string>{};
if ((in.macMask != cast<uint32_t>(0ULL))) {
dsts = append(dsts,Slice<std::string>{go_fmt_Sprintf(std::string("R%d%s",5),modi<int64_t>(in.macDst,cast<int64_t>(13ULL)),maskStr(in.macMask))});
}
if (((in.outMask != cast<uint32_t>(0ULL)) && (!in.outFromILU))) {
dsts = append(dsts,Slice<std::string>{(xbox_outName(in.outIsO,in.outAddr) + maskStr(in.outMask))});
}
if ((len(dsts) == cast<int64_t>(0ULL))) {
dsts = append(dsts,Slice<std::string>{std::string("_",1)});
}
p += (((std::string(" ",1) + go_strings_Join(dsts,std::string("/",1))) + std::string(", ",2)) + srcStr((&in.a)));
{
switch(in.mac){
case cast<int64_t>(2ULL):case cast<int64_t>(5ULL):case cast<int64_t>(6ULL):case cast<int64_t>(7ULL):case cast<int64_t>(8ULL):case cast<int64_t>(9ULL):case cast<int64_t>(10ULL):case cast<int64_t>(11ULL):case cast<int64_t>(12ULL):{
p += (std::string(", ",2) + srcStr((&in.b)));
break;}
case cast<int64_t>(3ULL):{
p += (std::string(", ",2) + srcStr((&in.c)));
break;}
case cast<int64_t>(4ULL):{
p += (((std::string(", ",2) + srcStr((&in.b))) + std::string(", ",2)) + srcStr((&in.c)));
break;}
}}
}
parts = append(parts,Slice<std::string>{p});
}
if ((in.ilu != cast<int64_t>(0ULL))) {
Slice<std::string> dsts = Slice<std::string>{};
if ((in.iluMask != cast<uint32_t>(0ULL))) {
dsts = append(dsts,Slice<std::string>{(std::string("R1",2) + maskStr(in.iluMask))});
}
if (((in.outMask != cast<uint32_t>(0ULL)) && in.outFromILU)) {
dsts = append(dsts,Slice<std::string>{(xbox_outName(in.outIsO,in.outAddr) + maskStr(in.outMask))});
}
if ((len(dsts) == cast<int64_t>(0ULL))) {
dsts = append(dsts,Slice<std::string>{std::string("_",1)});
}
parts = append(parts,Slice<std::string>{((((xbox_iluNames[in.ilu] + std::string(" ",1)) + go_strings_Join(dsts,std::string("/",1))) + std::string(", ",2)) + srcStr((&in.c)))});
}
std::string out = go_strings_Join(parts,std::string(" ; ",3));
if ((out == std::string("",0))) {
out = std::string("NOP",3);
}
if (in.final) {
out += std::string(" [FINAL]",8);
}
return out;
}
}
// tools/platform/xbox/nv2a_vsh.go:553:1
std::string xbox_outName(bool isO,int64_t addr){
{
if ((!isO)) {
return go_fmt_Sprintf(std::string("c%d",3),addr);
}
std::array<std::string,13> names = std::array<std::string,13>{std::string("oPos",4),std::string("o1",2),std::string("o2",2),std::string("oD0",3),std::string("oD1",3),std::string("oFog",4),std::string("oPts",4),std::string("oB0",3),std::string("oB1",3),std::string("oT0",3),std::string("oT1",3),std::string("oT2",3),std::string("oT3",3)};
{
int64_t a = cast<int64_t>((addr & cast<int64_t>(15ULL)));
if ((a < cast<int64_t>(13ULL))) {
return names[a];
}
}
return go_fmt_Sprintf(std::string("o%d",3),cast<int64_t>((addr & cast<int64_t>(15ULL))));
}
}
// tools/platform/xbox/nv2a_vsh.go:566:1
float xbox_minf32(float a,float b){
{
if ((a < b)) {
return a;
}
return b;
}
}
// tools/platform/xbox/nv2a_vsh.go:572:1
float xbox_maxf32(float a,float b){
{
if ((a > b)) {
return a;
}
return b;
}
}
// tools/platform/xbox/nv2a_vsh.go:578:1
float xbox_floorf32(float f){
{
return cast<float>(go_math_Floor(cast<double>(f)));
}
}
// tools/platform/xbox/nv2a_vsh.go:579:1
float xbox_exp2f32(float f){
{
return cast<float>(go_math_Exp2(cast<double>(f)));
}
}
// tools/platform/xbox/nv2a_vsh.go:581:1
float xbox_rcpf32(float x){
{
if ((x == cast<float>(0.00000000000000000e+00))) {
return cast<float>(go_math_Inf(cast<int64_t>(1ULL)));
}
return (cast<float>(1.00000000000000000e+00) / x);
}
}
// tools/platform/xbox/nv2a_vsh.go:591:1
float xbox_rccf32(float x){
{
constexpr double lo=5.42100999999999970e-20;
constexpr double hi=1.88446700000000000e+19;
double y = (cast<double>(1.00000000000000000e+00) / cast<double>(x));
bool neg = go_math_Signbit(y);
double a = go_math_Abs(y);
if ((((a > cast<double>(1.88446700000000000e+19)) || go_math_IsInf(a,cast<int64_t>(0ULL))) || go_math_IsNaN(a))) {
a = cast<double>(1.88446700000000000e+19);
}
else if ((a < cast<double>(5.42100999999999970e-20))) {
a = cast<double>(5.42100999999999970e-20);
}
if (neg) {
return cast<float>(cast<double>(-a));
}
return cast<float>(a);
}
}
// tools/platform/xbox/nv2a_vsh.go:607:1
float xbox_rsqf32(float x){
{
double a = go_math_Abs(cast<double>(x));
if ((a == cast<double>(0.00000000000000000e+00))) {
return cast<float>(go_math_Inf(cast<int64_t>(1ULL)));
}
return cast<float>((cast<double>(1.00000000000000000e+00) / go_math_Sqrt(a)));
}
}
// tools/platform/xbox/nv2a_vsh.go:616:1
std::array<float,4> xbox_log2partial(float x){
{
double a = go_math_Abs(cast<double>(x));
if ((a == cast<double>(0.00000000000000000e+00))) {
return std::array<float,4>{cast<float>(go_math_Inf(cast<int64_t>(-1ULL))),cast<float>(1.00000000000000000e+00),cast<float>(go_math_Inf(cast<int64_t>(-1ULL))),cast<float>(1.00000000000000000e+00)};
}
auto tmp369 = go_math_Frexp(a);
double fr = std::get<0>(tmp369);
int64_t exp = std::get<1>(tmp369);
return std::array<float,4>{cast<float>(cast<int64_t>((exp - cast<int64_t>(1ULL)))),cast<float>((fr * cast<double>(2.00000000000000000e+00))),cast<float>(go_math_Log2(a)),cast<float>(1.00000000000000000e+00)};
}
}
// tools/platform/xbox/nv2a_vsh.go:627:1
std::array<float,4> xbox_litf32(std::array<float,4> c){
{
std::array<float,4> out = std::array<float,4>{cast<float>(1.00000000000000000e+00),cast<float>(0.00000000000000000e+00),cast<float>(0.00000000000000000e+00),cast<float>(1.00000000000000000e+00)};
double p = cast<double>(c[cast<int64_t>(3ULL)]);
if ((p < cast<double>(-1.28000000000000000e+02))) {
p = cast<double>(-1.28000000000000000e+02);
}
else if ((p > cast<double>(1.28000000000000000e+02))) {
p = cast<double>(1.28000000000000000e+02);
}
if ((c[cast<int64_t>(0ULL)] > cast<float>(0.00000000000000000e+00))) {
out[cast<int64_t>(1ULL)] = c[cast<int64_t>(0ULL)];
if ((c[cast<int64_t>(1ULL)] > cast<float>(0.00000000000000000e+00))) {
out[cast<int64_t>(2ULL)] = cast<float>(go_math_Pow(cast<double>(c[cast<int64_t>(1ULL)]),p));
}
else if ((p == cast<double>(0.00000000000000000e+00))) {
out[cast<int64_t>(2ULL)] = cast<float>(1.00000000000000000e+00);
}
}
return out;
}
}
// tools/platform/xbox/ports.go:11:1
uint32_t xbox_Machine_portIn(xbox_Machine* m,uint16_t port,int64_t size){
{
{
switch(port){
case cast<uint16_t>(3324ULL):case cast<uint16_t>(3325ULL):case cast<uint16_t>(3326ULL):case cast<uint16_t>(3327ULL):{
return xbox_Machine_pciConfigRead(m,size);
break;}
case cast<uint16_t>(32776ULL):{
return cast<uint32_t>((cast<uint32_t>(m->tick) & cast<uint32_t>(16777215ULL)));
break;}
}}
return shr<uint32_t>(cast<uint32_t>(4294967295ULL),cast<uint64_t>(cast<int64_t>((cast<int64_t>(32ULL) - cast<int64_t>((size * cast<int64_t>(8ULL)))))));
}
}
// tools/platform/xbox/ports.go:23:1
void xbox_Machine_portOut(xbox_Machine* m,uint16_t port,int64_t size,uint32_t v){
{
{
switch(port){
case cast<uint16_t>(3320ULL):{
m->pciAddr = v;
break;}
}}
}
}
// tools/platform/xbox/ports.go:34:1
uint32_t xbox_Machine_pciConfigRead(xbox_Machine* m,int64_t size){
{
uint32_t reg = cast<uint32_t>((m->pciAddr & cast<uint32_t>(252ULL)));
uint32_t bus = cast<uint32_t>(((shr<uint32_t>(m->pciAddr,cast<int64_t>(16ULL))) & cast<uint32_t>(255ULL)));
uint32_t dev = cast<uint32_t>(((shr<uint32_t>(m->pciAddr,cast<int64_t>(11ULL))) & cast<uint32_t>(31ULL)));
uint32_t v={};
{
if ((((bus == cast<uint32_t>(0ULL)) && (dev == cast<uint32_t>(0ULL))) && (reg == cast<uint32_t>(0ULL)))){
v = cast<uint32_t>(44044510ULL);
}
else if ((((bus == cast<uint32_t>(1ULL)) && (dev == cast<uint32_t>(0ULL))) && (reg == cast<uint32_t>(0ULL)))){
v = cast<uint32_t>(44044510ULL);
}
else {
v = cast<uint32_t>(4294967295ULL);
}
}
tmp370:;
return cast<uint32_t>((shr<uint32_t>(v,cast<uint64_t>(cast<uint32_t>(((cast<uint32_t>((m->pciAddr & cast<uint32_t>(3ULL)))) * cast<uint32_t>(8ULL))))) & (shr<uint32_t>(cast<uint32_t>(4294967295ULL),cast<uint64_t>(cast<int64_t>((cast<int64_t>(32ULL) - cast<int64_t>((size * cast<int64_t>(8ULL))))))))));
}
}
// tools/platform/xbox/profile.go:94:1
xbox_profCounters xbox_Machine_profCounters(xbox_Machine* m){
{
xbox_pgraph* g = m->pgraph;
return xbox_profCounters{g->Draws,g->Methods,g->pixWritten,g->pixZRej,g->pixARej};
}
}
// tools/platform/xbox/profile.go:125:1
time_Time xbox_Machine_profStart(xbox_Machine* m){
{
if ((!m->Profile)) {
return time_Time{};
}
return go_time_Now();
}
}
// tools/platform/xbox/profile.go:132:1
void xbox_Machine_profEnd(xbox_Machine* m,int64_t bucket,time_Time t){
{
if (time_Time_IsZero(t)) {
return ;
}
m->prof.ns[bucket] += cast<int64_t>(go_time_Since(t));
m->prof.count[bucket]++;
}
}
// tools/platform/xbox/profile.go:144:1
void xbox_Machine_profPusherEnter(xbox_Machine* m){
{
if (((!m->Profile) || m->prof.inPush)) {
return ;
}
auto tmp371 = std::make_tuple(go_time_Now(),true);
m->prof.pushStart = std::get<0>(tmp371);
m->prof.inPush = std::get<1>(tmp371);
}
}
// tools/platform/xbox/profile.go:151:1
void xbox_Machine_profPusherExit(xbox_Machine* m){
{
if (((!m->Profile) || (!m->prof.inPush))) {
return ;
}
m->prof.ns[cast<int64_t>(3ULL)] += cast<int64_t>(go_time_Since(m->prof.pushStart));
m->prof.count[cast<int64_t>(3ULL)]++;
m->prof.inPush = false;
}
}
// tools/platform/xbox/profile.go:162:1
void xbox_Machine_profRunEnter(xbox_Machine* m){
{
if ((!m->Profile)) {
return ;
}
auto tmp372 = std::make_tuple(go_time_Now(),true);
m->prof.runStart = std::get<0>(tmp372);
m->prof.inRun = std::get<1>(tmp372);
}
}
// tools/platform/xbox/profile.go:169:1
void xbox_Machine_profRunExit(xbox_Machine* m){
{
if (((!m->Profile) || (!m->prof.inRun))) {
return ;
}
m->prof.frameNs += cast<int64_t>(go_time_Since(m->prof.runStart));
m->prof.inRun = false;
}
}
// tools/platform/xbox/profile.go:192:1
void xbox_Machine_profFrame(xbox_Machine* m){
{
if ((!m->Profile)) {
return ;
}
xbox_profState* p = (&m->prof);
int64_t total = p->frameNs;
if (p->inRun) {
total += cast<int64_t>(go_time_Since(p->runStart));
p->runStart = go_time_Now();
}
p->frameNs = cast<int64_t>(0ULL);
if (p->inPush) {
p->ns[cast<int64_t>(3ULL)] += cast<int64_t>(go_time_Since(p->pushStart));
p->count[cast<int64_t>(3ULL)]++;
p->pushStart = go_time_Now();
}
auto ms = [&](int64_t ns)->double{
return (cast<double>(ns) / cast<double>(1.00000000000000000e+06));
}
;
int64_t nested = cast<int64_t>((cast<int64_t>((p->ns[cast<int64_t>(0ULL)] + p->ns[cast<int64_t>(1ULL)])) + p->ns[cast<int64_t>(2ULL)]));
int64_t decode = cast<int64_t>((p->ns[cast<int64_t>(3ULL)] - nested));
if ((decode < cast<int64_t>(0ULL))) {
decode = cast<int64_t>(0ULL);
}
Slice<xbox_ProfileBucket> buckets = Slice<xbox_ProfileBucket>{xbox_ProfileBucket{std::string("vertex + xf",11),ms(p->ns[cast<int64_t>(0ULL)]),cast<int64_t>(p->count[cast<int64_t>(0ULL)])},xbox_ProfileBucket{std::string("rasterise",9),ms(p->ns[cast<int64_t>(1ULL)]),cast<int64_t>(p->count[cast<int64_t>(1ULL)])},xbox_ProfileBucket{std::string("clear",5),ms(p->ns[cast<int64_t>(2ULL)]),cast<int64_t>(p->count[cast<int64_t>(2ULL)])},xbox_ProfileBucket{std::string("command decode (derived)",24),ms(decode),cast<int64_t>(p->count[cast<int64_t>(3ULL)])}};
int64_t other = cast<int64_t>((total - p->ns[cast<int64_t>(3ULL)]));
if ((other < cast<int64_t>(0ULL))) {
other = cast<int64_t>(0ULL);
}
buckets = append(buckets,Slice<xbox_ProfileBucket>{xbox_ProfileBucket{std::string("x86 + rest (derived)",20),ms(other),cast<int64_t>(0ULL)}});
xbox_profCounters now = xbox_Machine_profCounters(m);
xbox_profCounters d = xbox_profCounters{cast<int64_t>((now.draws - p->base.draws)),cast<int64_t>((now.methods - p->base.methods)),cast<int64_t>((now.frags - p->base.frags)),cast<int64_t>((now.zRej - p->base.zRej)),cast<int64_t>((now.aRej - p->base.aRej))};
int64_t instrs = cast<int64_t>(cast<uint64_t>((m->CPU->Steps - p->baseInstr)));
p->last = xbox_FrameProfile{ms(total),buckets,Slice<xbox_ProfileCounter>{xbox_ProfileCounter{std::string("draws",5),d.draws},xbox_ProfileCounter{std::string("nv methods",10),d.methods},xbox_ProfileCounter{std::string("fragments drawn",15),d.frags},xbox_ProfileCounter{std::string("depth-killed",12),d.zRej},xbox_ProfileCounter{std::string("alpha-killed",12),d.aRej},xbox_ProfileCounter{std::string("x86 instructions",16),instrs}},(d.draws > cast<int64_t>(0ULL))};
p->has = true;
auto tmp373 = std::make_tuple(std::array<int64_t,4>{},std::array<int64_t,4>{});
p->ns = std::get<0>(tmp373);
p->count = std::get<1>(tmp373);
auto tmp374 = std::make_tuple(now,m->CPU->Steps);
p->base = std::get<0>(tmp374);
p->baseInstr = std::get<1>(tmp374);
}
}
// tools/platform/xbox/profile.go:263:1
xbox_FrameProfile xbox_Machine_FrameProfile(xbox_Machine* m){
{
if ((!m->prof.has)) {
return xbox_FrameProfile{};
}
return m->prof.last;
}
}
// tools/platform/xbox/profile.go:273:1
void xbox_Machine_SetProfile(xbox_Machine* m,bool on){
{
m->Profile = on;
m->prof = xbox_profState{};
if (on) {
auto tmp375 = std::make_tuple(xbox_Machine_profCounters(m),m->CPU->Steps);
m->prof.base = std::get<0>(tmp375);
m->prof.baseInstr = std::get<1>(tmp375);
}
}
}
// tools/platform/xbox/run.go:25:1
std::string xbox_StopReason_String(xbox_StopReason r){
{
{
switch(r){
case cast<xbox_StopReason>(0ULL):{
return std::string("budget exhausted",16);
break;}
case cast<xbox_StopReason>(1ULL):{
return std::string("halted",6);
break;}
case cast<xbox_StopReason>(2ULL):{
return std::string("first NV2A push reached",23);
break;}
case cast<xbox_StopReason>(3ULL):{
return std::string("stop requested",14);
break;}
case cast<xbox_StopReason>(4ULL):{
return std::string("breakpoint",10);
break;}
default:{
return std::string("?",1);
break;}
}}
}
}
// tools/platform/xbox/run.go:45:1
std::tuple<xbox_StopReason,uint64_t> xbox_Machine_Run(xbox_Machine* m,uint64_t maxSteps){
{rrprof::Scope timing(0,"x86 CPU and devices");
{
m->StopRequested = false;
xbox_Machine_profRunEnter(m);
auto tmp376=defer([&](){xbox_Machine_profRunExit(m);});
uint64_t n={};
{;for (;(n < maxSteps);){
if (m->CPU->Halted) {
return {cast<xbox_StopReason>(1ULL),n};
}
if ((m->firstPush && (!m->pusherEnabled))) {
return {cast<xbox_StopReason>(2ULL),n};
}
if ((((n > cast<uint64_t>(0ULL)) && (len(m->bps) > cast<int64_t>(0ULL))) && get(m->bps,xbox_Machine_PC(m)))) {
return {cast<xbox_StopReason>(4ULL),n};
}
x86_CPU_Step(m->CPU);
n++;
if (m->StopRequested) {
m->StopRequested = false;
return {cast<xbox_StopReason>(3ULL),n};
}
}
}return {cast<xbox_StopReason>(0ULL),n};
}
}
}
// tools/platform/xbox/run.go:89:1
std::tuple<xbox_StopReason,uint64_t> xbox_Machine_RunStopAfterNVMethod(xbox_Machine* m,int64_t k,uint64_t maxSteps){
{
auto tmp377 = std::make_tuple(k,true);
m->stopAfterMethod = std::get<0>(tmp377);
m->stopAfterArmed = std::get<1>(tmp377);
auto tmp379=defer([&](){[&]()->void{
auto tmp378 = std::make_tuple(cast<int64_t>(0ULL),false);
m->stopAfterMethod = std::get<0>(tmp378);
m->stopAfterArmed = std::get<1>(tmp378);
}
();});
return xbox_Machine_Run(m,maxSteps);
}
}
// tools/platform/xbox/run.go:99:1
uint32_t xbox_Machine_PC(xbox_Machine* m){
{
return cast<uint32_t>((m->CPU->SegBase[cast<int64_t>(1ULL)] + m->CPU->IP));
}
}
// tools/platform/xbox/run.go:103:1
void xbox_Machine_SetBreakpoint(xbox_Machine* m,uint32_t pc){
{
if ((!m->bps)) {
m->bps = Map<uint32_t,bool>{};
}
m->bps[pc] = true;
}
}
// tools/platform/xbox/run.go:110:1
void xbox_Machine_ClearBreakpoint(xbox_Machine* m,uint32_t pc){
{
removeKey(m->bps,pc);
}
}
// tools/platform/xbox/run.go:112:1
void xbox_Machine_ClearBreakpoints(xbox_Machine* m){
{
m->bps = {};
}
}
// tools/platform/xbox/run.go:114:1
Slice<uint32_t> xbox_Machine_Breakpoints(xbox_Machine* m){
{
Slice<uint32_t> out = Slice<uint32_t>::make(cast<int64_t>(0ULL),len(m->bps));
{auto&& tmp380 = m->bps;
for(auto [tmp381,tmp382]:tmp380){
auto pc=tmp381;out = append(out,Slice<uint32_t>{pc});
}}
xbox_sortU32(out);
return out;
}
}
// tools/platform/xbox/run.go:128:1
void xbox_Machine_ClearHalt(xbox_Machine* m){
{
auto tmp383 = std::make_tuple(false,std::string("",0));
m->CPU->Halted = std::get<0>(tmp383);
m->CPU->HaltReason = std::get<1>(tmp383);
auto tmp384 = std::make_tuple(false,std::string("",0));
m->Halted = std::get<0>(tmp384);
m->HaltReason = std::get<1>(tmp384);
}
}
// tools/platform/xbox/run.go:135:1
std::string xbox_Machine_Report(xbox_Machine* m){
{
x86_CPU* c = m->CPU;
std::string s = go_fmt_Sprintf(std::string("xbox: %q (title id %08X)\012",25),m->XBE->TitleName,m->XBE->TitleID);
s += go_fmt_Sprintf(std::string("  entry %08X  base %08X  imageSize %08X\012",40),m->XBE->Entry,m->XBE->Base,m->XBE->ImageSize);
s += go_fmt_Sprintf(std::string("  steps=%d  PC=%08X  EAX=%08X EBX=%08X ECX=%08X EDX=%08X ESP=%08X\012",66),c->Steps,x86_CPU_LinearPC(c),c->Regs[cast<int64_t>(0ULL)],c->Regs[cast<int64_t>(3ULL)],c->Regs[cast<int64_t>(1ULL)],c->Regs[cast<int64_t>(2ULL)],c->Regs[cast<int64_t>(4ULL)]);
if (c->Halted) {
s += ((std::string("  HALT: ",8) + c->HaltReason) + std::string("\012",1));
}
if (m->firstPush) {
s += std::string("  \342\230\205 first NV2A push-buffer kick reached\012",42);
}
s += go_fmt_Sprintf(std::string("  distinct ordinals called: %d\012",31),len(m->OrdinalHits));
return s;
}
}
// tools/platform/xbox/run.go:153:1
Slice<std::string> xbox_Machine_OrdinalHistogram(xbox_Machine* m){
{
Slice<int64_t> keys = Slice<int64_t>::make(cast<int64_t>(0ULL),len(m->OrdinalHits));
{auto&& tmp385 = m->OrdinalHits;
for(auto [tmp386,tmp387]:tmp385){
auto o=tmp386;keys = append(keys,Slice<int64_t>{cast<int64_t>(o)});
}}
{int64_t i = cast<int64_t>(1ULL);for (;(i < len(keys));i++){
{int64_t j = i;for (;((j > cast<int64_t>(0ULL)) && (keys[cast<int64_t>((j - cast<int64_t>(1ULL)))] > keys[j]));j--){
auto tmp388 = std::make_tuple(keys[j],keys[cast<int64_t>((j - cast<int64_t>(1ULL)))]);
keys[cast<int64_t>((j - cast<int64_t>(1ULL)))] = std::get<0>(tmp388);
keys[j] = std::get<1>(tmp388);
}
}}
}Slice<std::string> out = Slice<std::string>::make(cast<int64_t>(0ULL),len(keys));
{auto&& tmp389 = keys;
for(int64_t tmp390=0;tmp390<len(tmp389);++tmp390){
auto o=tmp389[tmp390];out = append(out,Slice<std::string>{go_fmt_Sprintf(std::string("  ord %3d %-34s x%d",19),o,xbox_ordinalName(cast<uint16_t>(o)),get(m->OrdinalHits,cast<uint16_t>(o)))});
}}
return out;
}
}
// tools/platform/xbox/sched.go:47:1
uint64_t xbox_Machine_systemTime100ns(xbox_Machine* m){
{
return cast<uint64_t>((m->clockBase100ns + divi<uint64_t>(cast<uint64_t>(((cast<uint64_t>((m->tick - m->clockBaseTick))) * cast<uint64_t>(10000ULL))),cast<uint64_t>(733466ULL))));
}
}
// tools/platform/xbox/sched.go:52:1
uint64_t xbox_Machine_guestMs(xbox_Machine* m){
{
return divi<uint64_t>(xbox_Machine_systemTime100ns(m),cast<uint64_t>(10000ULL));
}
}
// tools/platform/xbox/sched.go:59:1
uint64_t xbox_Machine_guestTSC(xbox_Machine* m){
{
return cast<uint64_t>((m->tscBase + (cast<uint64_t>((m->tick - m->clockBaseTick)))));
}
}
// tools/platform/xbox/sched.go:66:1
void xbox_Machine_schedTick(xbox_Machine* m){
{
if ((cast<uint64_t>((m->tick & cast<uint64_t>(1023ULL))) == cast<uint64_t>(0ULL))) {
if ((m->tickCountAddr != cast<uint32_t>(0ULL))) {
xbox_Machine_write32(m,m->tickCountAddr,cast<uint32_t>(xbox_Machine_guestMs(m)));
}
if ((m->systemTimeAddr != cast<uint32_t>(0ULL))) {
uint64_t t = xbox_Machine_systemTime100ns(m);
xbox_Machine_write32(m,m->systemTimeAddr,cast<uint32_t>(t));
xbox_Machine_write32(m,cast<uint32_t>((m->systemTimeAddr + cast<uint32_t>(4ULL))),cast<uint32_t>(shr<uint64_t>(t,cast<int64_t>(32ULL))));
}
xbox_Machine_apuTick(m);
xbox_Machine_ioTick(m);
xbox_Machine_timerTick(m);
xbox_Machine_usbTick(m);
xbox_Machine_vblankTick(m);
}
if (m->isrActive) {
return ;
}
xbox_Machine_wakeDueSleepers(m);
if (m->reschedule) {
m->reschedule = false;
xbox_Machine_dispatch(m);
return ;
}
m->quantumLeft--;
if ((m->quantumLeft <= cast<int64_t>(0ULL))) {
m->quantumLeft = cast<int64_t>(4000ULL);
if ((len(m->threads) > cast<int64_t>(1ULL))) {
xbox_Machine_dispatch(m);
}
}
}
}
// tools/platform/xbox/sched.go:106:1
void xbox_Machine_yieldCurrent(xbox_Machine* m,xbox_threadState state){
{
if (bool(m->current)) {
m->current->state = state;
}
m->reschedule = true;
}
}
// tools/platform/xbox/sched.go:118:1
void xbox_Machine_dispatch(xbox_Machine* m){
{
xbox_thread* next = xbox_Machine_pickRunnable(m);
if (((((!next) && bool(m->current)) && (m->current->state == cast<xbox_threadState>(1ULL))) && (m->current->suspendCount == cast<int32_t>(0ULL)))) {
return ;
}
{;for (;(!next);){
if ((!xbox_Machine_idleAdvance(m))) {
x86_CPU_Halt(m->CPU,std::string("scheduler: no runnable thread (deadlock); %d threads",52),len(m->threads));
auto tmp391 = std::make_tuple(true,m->CPU->HaltReason);
m->Halted = std::get<0>(tmp391);
m->HaltReason = std::get<1>(tmp391);
return ;
}
next = xbox_Machine_pickRunnable(m);
}
}xbox_Machine_switchTo(m,next);
}
}
// tools/platform/xbox/sched.go:138:1
bool xbox_Machine_idleAdvance(xbox_Machine* m){
{
uint64_t min={};
{auto&& tmp392 = m->threads;
for(int64_t tmp393=0;tmp393<len(tmp392);++tmp393){
auto t=tmp392[tmp393];if (((((t->state == cast<xbox_threadState>(3ULL)) || (((t->state == cast<xbox_threadState>(2ULL)) && (t->wakeTick != cast<uint64_t>(0ULL)))))) && (((min == cast<uint64_t>(0ULL)) || (t->wakeTick < min))))) {
min = t->wakeTick;
}
}}
{auto&& tmp394 = m->pendingIO;
for(int64_t tmp395=0;tmp395<len(tmp394);++tmp395){
auto p=tmp394[tmp395];if (((min == cast<uint64_t>(0ULL)) || (p.Due < min))) {
min = p.Due;
}
}}
if ((min == cast<uint64_t>(0ULL))) {
return false;
}
if ((min > m->tick)) {
m->tick = min;
}
xbox_Machine_wakeDueSleepers(m);
xbox_Machine_ioTick(m);
return true;
}
}
// tools/platform/xbox/sched.go:163:1
xbox_thread* xbox_Machine_pickRunnable(xbox_Machine* m){
{
xbox_thread* best={};
int64_t n = len(m->threads);
if ((n == cast<int64_t>(0ULL))) {
return {};
}
{int64_t i = cast<int64_t>(0ULL);for (;(i < n);i++){
xbox_thread* t = m->threads[modi<int64_t>((cast<int64_t>((m->rrCursor + i))),n)];
if ((xbox_thread_runnable(t) && (((!best) || (t->priority > best->priority))))) {
best = t;
}
}
}if (bool(best)) {
m->rrCursor = modi<int64_t>((cast<int64_t>((m->rrCursor + cast<int64_t>(1ULL)))),n);
}
return best;
}
}
// tools/platform/xbox/sched.go:182:1
void xbox_Machine_switchTo(xbox_Machine* m,xbox_thread* t){
{
if ((m->current == t)) {
t->state = cast<xbox_threadState>(1ULL);
return ;
}
if (bool(m->current)) {
m->current->ctx = (*m->CPU);
if ((m->current->state == cast<xbox_threadState>(1ULL))) {
m->current->state = cast<xbox_threadState>(0ULL);
}
}
(*m->CPU) = t->ctx;
m->current = t;
t->state = cast<xbox_threadState>(1ULL);
m->quantumLeft = cast<int64_t>(4000ULL);
xbox_Machine_write32(m,cast<uint32_t>(66846760ULL),t->kthread);
if ((t->stackTop != cast<uint32_t>(0ULL))) {
xbox_Machine_write32(m,cast<uint32_t>(66846724ULL),t->stackTop);
xbox_Machine_write32(m,cast<uint32_t>(66846728ULL),t->stackLimit);
}
}
}
// tools/platform/xbox/sched.go:215:1
void xbox_Machine_wakeDueSleepers(xbox_Machine* m){
{
{auto&& tmp396 = m->threads;
for(int64_t tmp397=0;tmp397<len(tmp396);++tmp397){
auto t=tmp396[tmp397];{
if (((t->state == cast<xbox_threadState>(3ULL)) && (t->wakeTick <= m->tick))){
t->state = cast<xbox_threadState>(0ULL);
}
else if ((((t->state == cast<xbox_threadState>(2ULL)) && (t->wakeTick != cast<uint64_t>(0ULL))) && (t->wakeTick <= m->tick))){
t->state = cast<xbox_threadState>(0ULL);
t->waitObjs = {};
t->wakeTick = cast<uint64_t>(0ULL);
}
}
tmp398:;
}}
}
}
// tools/platform/xbox/sched.go:229:1
int64_t xbox_Machine_aliveThreads(xbox_Machine* m){
{
int64_t n = cast<int64_t>(0ULL);
{auto&& tmp399 = m->threads;
for(int64_t tmp400=0;tmp400<len(tmp399);++tmp400){
auto t=tmp399[tmp400];if ((t->state != cast<xbox_threadState>(4ULL))) {
n++;
}
}}
return n;
}
}
// tools/platform/xbox/thread.go:33:1
std::string xbox_threadState_String(xbox_threadState s){
{
return std::array<std::string,5>{std::string("ready",5),std::string("running",7),std::string("waiting",7),std::string("sleeping",8),std::string("dead",4)}[s];
}
}
// tools/platform/xbox/thread.go:78:1
bool xbox_thread_runnable(xbox_thread* t){
{
return ((t->state == cast<xbox_threadState>(0ULL)) && (t->suspendCount == cast<int32_t>(0ULL)));
}
}
// tools/platform/xbox/thread.go:100:1
void xbox_kobject_noteSignal(xbox_kobject* o,std::string who,uint32_t from,uint64_t tick){
{
auto tmp401 = std::make_tuple(who,from,tick);
o->lastSigWho = std::get<0>(tmp401);
o->lastSigFrom = std::get<1>(tmp401);
o->lastSigTick = std::get<2>(tmp401);
}
}
// tools/platform/xbox/thread.go:107:1
Slice<std::string> xbox_Machine_DebugThreads(xbox_Machine* m){
{
Slice<std::string> out = Slice<std::string>::make(cast<int64_t>(0ULL),len(m->threads));
{auto&& tmp402 = m->threads;
for(int64_t tmp403=0;tmp403<len(tmp402);++tmp403){
auto t=tmp402[tmp403];uint32_t pc = cast<uint32_t>((t->ctx.SegBase[cast<int64_t>(1ULL)] + t->ctx.IP));
std::string mark = std::string(" ",1);
if ((t == m->current)) {
pc = cast<uint32_t>((m->CPU->SegBase[cast<int64_t>(1ULL)] + m->CPU->IP));
mark = std::string("*",1);
}
out = append(out,Slice<std::string>{go_fmt_Sprintf(std::string("%s tid=%d %-8s prio=%d susp=%d PC=%08X wakeTick=%d waitObjs=%v",62),mark,t->id,t->state,t->priority,t->suspendCount,pc,t->wakeTick,t->waitObjs)});
{auto&& tmp404 = t->waitObjs;
for(int64_t tmp405=0;tmp405<len(tmp404);++tmp405){
auto h=tmp404[tmp405];xbox_kobject* o = get(m->objects,h);
if ((!o)) {
out = append(out,Slice<std::string>{go_fmt_Sprintf(std::string("      waits on %08X: (no kobject; guest header signal=%d)",57),h,xbox_Machine_read32(m,cast<uint32_t>((h + cast<uint32_t>(4ULL)))))});
continue;
}
std::string sig = std::string("never",5);
if ((o->lastSigWho != std::string("",0))) {
sig = go_fmt_Sprintf(std::string("%s from %08X at tick %d",23),o->lastSigWho,o->lastSigFrom,o->lastSigTick);
}
out = append(out,Slice<std::string>{go_fmt_Sprintf(std::string("      waits on %08X: %s signaled=%v count=%d lastSignal=%s",58),h,o->kind,o->signaled,o->count,sig)});
}}
}}
return out;
}
}
// tools/platform/xbox/thread.go:141:1
int64_t xbox_Machine_DebugCurrentTid(xbox_Machine* m){
{
if ((!m->current)) {
return cast<int64_t>(-1ULL);
}
return cast<int64_t>(m->current->id);
}
}
// tools/platform/xbox/thread.go:151:1
void xbox_Machine_bootThread(xbox_Machine* m){
{
uint32_t kt = xbox_Machine_allocKObject(m,cast<uint32_t>(256ULL));
xbox_thread* t = arenaNew(xbox_thread{cast<uint32_t>(0ULL),kt,{},cast<int32_t>(16ULL),cast<xbox_threadState>(1ULL),{},{},{},{},{},cast<uint32_t>(66842624ULL),cast<uint32_t>(66777088ULL)});
m->nextTID = cast<uint32_t>(1ULL);
m->threads = append(m->threads,Slice<xbox_thread*>{t});
m->current = t;
m->objects[kt] = arenaNew(xbox_kobject{std::string("thread",6),kt,{},{},{},t,{},{},{}});
xbox_Machine_write32(m,cast<uint32_t>((kt + cast<uint32_t>(40ULL))),cast<uint32_t>(66842624ULL));
xbox_Machine_write32(m,cast<uint32_t>(66846760ULL),kt);
}
}
// tools/platform/xbox/thread.go:175:1
uint32_t xbox_Machine_currentKThread(xbox_Machine* m){
{
if ((!m->current)) {
return cast<uint32_t>(0ULL);
}
return m->current->kthread;
}
}
// tools/platform/xbox/thread.go:197:1
xbox_thread* xbox_Machine_createThread(xbox_Machine* m,uint32_t entry,uint32_t ctx1,uint32_t ctx2,uint32_t stackSize,uint32_t tlsSize,int32_t priority,bool suspended){
{
if ((stackSize == cast<uint32_t>(0ULL))) {
stackSize = cast<uint32_t>(16384ULL);
}
stackSize = xbox_align32(stackSize,cast<uint32_t>(4096ULL));
uint32_t stackBase = xbox_Machine_allocPool(m,stackSize);
if ((stackBase == cast<uint32_t>(0ULL))) {
x86_CPU_Halt(m->CPU,std::string("createThread: pool exhausted allocating a %X-byte stack",55),stackSize);
}
uint32_t tlsData = cast<uint32_t>((stackBase + stackSize));
if ((tlsSize != cast<uint32_t>(0ULL))) {
tlsSize = xbox_align32(tlsSize,cast<uint32_t>(4ULL));
tlsData -= tlsSize;
xbox_Machine_initTLSArea(m,tlsData,tlsSize);
}
uint32_t sp = cast<uint32_t>((tlsData - cast<uint32_t>(16ULL)));
uint32_t kt = xbox_Machine_allocKObject(m,cast<uint32_t>(256ULL));
xbox_Machine_write32(m,cast<uint32_t>((kt + cast<uint32_t>(40ULL))),tlsData);
xbox_thread* t = arenaNew(xbox_thread{m->nextTID,kt,{},priority,cast<xbox_threadState>(0ULL),{},{},{},{},{},cast<uint32_t>((stackBase + stackSize)),stackBase});
if (suspended) {
t->suspendCount = cast<int32_t>(1ULL);
}
m->nextTID++;
t->ctx = (*m->CPU);
t->ctx.Regs = std::array<uint32_t,8>{};
t->ctx.IP = entry;
t->ctx.Regs[cast<int64_t>(4ULL)] = sp;
xbox_Machine_pushCtx(m,(&t->ctx),ctx2);
xbox_Machine_pushCtx(m,(&t->ctx),ctx1);
xbox_Machine_pushCtx(m,(&t->ctx),cast<uint32_t>(2399141632ULL));
t->ctx.Halted = false;
t->ctx.HaltReason = std::string("",0);
m->threads = append(m->threads,Slice<xbox_thread*>{t});
m->objects[kt] = arenaNew(xbox_kobject{std::string("thread",6),kt,{},{},{},t,{},{},{}});
xbox_Machine_logf(m,std::string("createThread id=%d entry=%08X ctx1=%08X ctx2=%08X sp=%08X prio=%d susp=%v -> KTHREAD %08X",89),t->id,entry,ctx1,ctx2,sp,priority,suspended,kt);
return t;
}
}
// tools/platform/xbox/thread.go:267:1
void xbox_Machine_initTLSArea(xbox_Machine* m,uint32_t addr,uint32_t size){
{
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < size);i++){
xbox_Machine_Write(m,cast<uint32_t>((addr + i)),cast<uint8_t>(0ULL));
}
}}
}
// tools/platform/xbox/thread.go:274:1
void xbox_Machine_pushCtx(xbox_Machine* m,x86_CPU* c,uint32_t v){
{
c->Regs[cast<int64_t>(4ULL)] -= cast<uint32_t>(4ULL);
xbox_Machine_write32(m,c->Regs[cast<int64_t>(4ULL)],v);
}
}
// tools/platform/xbox/thread.go:280:1
void xbox_Machine_exitCurrentThread(xbox_Machine* m){
{
xbox_Machine_logf(m,std::string("thread %d exited (returned to sentinel) at tick %d",50),xbox_Machine_threadID(m),m->tick);
if (bool(m->current)) {
m->current->state = cast<xbox_threadState>(4ULL);
removeKey(m->objects,m->current->kthread);
}
m->reschedule = true;
xbox_Machine_dispatch(m);
}
}
// tools/platform/xbox/thread.go:290:1
uint32_t xbox_Machine_threadID(xbox_Machine* m){
{
if ((!m->current)) {
return cast<uint32_t>(0ULL);
}
return m->current->id;
}
}
// tools/platform/xbox/timer.go:47:1
uint32_t xbox_b2u(bool b){
{
if (b) {
return cast<uint32_t>(1ULL);
}
return cast<uint32_t>(0ULL);
}
}
// tools/platform/xbox/timer.go:77:1
bool xbox_Machine_armTimer(xbox_Machine* m,uint32_t tm,uint32_t dpc,uint32_t dueLo,uint32_t dueHi,uint32_t period){
{
if ((tm == cast<uint32_t>(0ULL))) {
return false;
}
uint64_t due = cast<uint64_t>((shl<uint64_t>(cast<uint64_t>(dueHi),cast<int64_t>(32ULL)) | cast<uint64_t>(dueLo)));
uint64_t deadline={};
if ((cast<int64_t>(due) < cast<int64_t>(0ULL))) {
deadline = cast<uint64_t>((xbox_Machine_systemTime100ns(m) + cast<uint64_t>(cast<int64_t>(-cast<int64_t>(due)))));
}
else {
deadline = due;
}
bool was = xbox_Machine_cancelTimer(m,tm);
xbox_Machine_write32(m,cast<uint32_t>((tm + cast<uint32_t>(4ULL))),cast<uint32_t>(0ULL));
m->timers = append(m->timers,Slice<xbox_ktimer>{xbox_ktimer{tm,dpc,deadline,period}});
return was;
}
}
// tools/platform/xbox/timer.go:95:1
bool xbox_Machine_cancelTimer(xbox_Machine* m,uint32_t tm){
{
{auto&& tmp406 = m->timers;
for(int64_t tmp407=0;tmp407<len(tmp406);++tmp407){
auto i=tmp407;if ((m->timers[i].Timer == tm)) {
m->timers = append(sub(m->timers,0,i),sub(m->timers,cast<int64_t>((i + cast<int64_t>(1ULL))),len(m->timers)));
return true;
}
}}
return false;
}
}
// tools/platform/xbox/timer.go:106:1
bool xbox_Machine_queueDPC(xbox_Machine* m,uint32_t dpc,uint32_t arg1,uint32_t arg2){
{
if ((dpc == cast<uint32_t>(0ULL))) {
return false;
}
{auto&& tmp408 = m->dpcQueue;
for(int64_t tmp409=0;tmp409<len(tmp408);++tmp409){
auto d=tmp408[tmp409];if ((d.Dpc == dpc)) {
return false;
}
}}
m->dpcQueue = append(m->dpcQueue,Slice<xbox_dpcEntry>{xbox_dpcEntry{dpc,arg1,arg2}});
return true;
}
}
// tools/platform/xbox/timer.go:126:1
void xbox_Machine_timerTick(xbox_Machine* m){
{
if ((len(m->timers) == cast<int64_t>(0ULL))) {
return ;
}
uint64_t now = xbox_Machine_systemTime100ns(m);
{int64_t i = cast<int64_t>(0ULL);for (;(i < len(m->timers));){
xbox_ktimer t = m->timers[i];
if ((now < t.Due)) {
i++;
continue;
}
{
xbox_kobject* o = xbox_Machine_guestObjAt(m,t.Timer);
if (bool(o)) {
o->signaled = true;
xbox_Machine_writeSignal(m,o->addr,true);
xbox_kobject_noteSignal(o,std::string("timer",5),cast<uint32_t>(0ULL),m->tick);
xbox_Machine_wakeWaiters(m,o->addr);
}
}
xbox_Machine_queueDPC(m,t.Dpc,cast<uint32_t>(0ULL),cast<uint32_t>(0ULL));
if ((t.Period != cast<uint32_t>(0ULL))) {
m->timers[i].Due = cast<uint64_t>((now + cast<uint64_t>((cast<uint64_t>(t.Period) * cast<uint64_t>(10000ULL)))));
i++;
}
else {
m->timers = append(sub(m->timers,0,i),sub(m->timers,cast<int64_t>((i + cast<int64_t>(1ULL))),len(m->timers)));
}
}
}xbox_Machine_deliverDPC(m);
}
}
// tools/platform/xbox/timer.go:172:1
void xbox_Machine_deliverDPC(xbox_Machine* m){
{
if ((((m->isrActive || (!m->CPU)) || m->CPU->Halted) || (!m->CPU->IF))) {
return ;
}
if ((xbox_Machine_Read(m,cast<uint32_t>(66846756ULL)) != cast<uint8_t>(0ULL))) {
return ;
}
if ((len(m->dpcQueue) == cast<int64_t>(0ULL))) {
return ;
}
xbox_dpcEntry d = m->dpcQueue[cast<int64_t>(0ULL)];
uint32_t routine = xbox_Machine_read32(m,cast<uint32_t>((d.Dpc + cast<uint32_t>(12ULL))));
if ((routine == cast<uint32_t>(0ULL))) {
m->dpcQueue = sub(m->dpcQueue,cast<int64_t>(1ULL),len(m->dpcQueue));
return ;
}
m->dpcQueue = sub(m->dpcQueue,cast<int64_t>(1ULL),len(m->dpcQueue));
m->isrSaved = (*m->CPU);
m->isrActive = true;
x86_CPU* c = m->CPU;
auto push = [&](uint32_t v)->void{
c->Regs[cast<int64_t>(4ULL)] -= cast<uint32_t>(4ULL);
xbox_Machine_write32(m,c->Regs[cast<int64_t>(4ULL)],v);
}
;
push(d.Arg2);
push(d.Arg1);
push(xbox_Machine_read32(m,cast<uint32_t>((d.Dpc + cast<uint32_t>(16ULL)))));
push(d.Dpc);
push(cast<uint32_t>(2399141376ULL));
c->IP = routine;
}
}
// tools/platform/xbox/usb.go:150:1
uint64_t xbox_Machine_usbFrame(xbox_Machine* m){
{
return xbox_Machine_guestMs(m);
}
}
// tools/platform/xbox/usb.go:155:1
uint8_t xbox_Machine_usbRead(xbox_Machine* m,uint32_t off){
{
auto tmp410 = lookup(m->usb.reg,shr<uint32_t>(off,cast<int64_t>(2ULL)));
uint32_t dw = std::get<0>(tmp410);
bool written = std::get<1>(tmp410);
{
switch((off & ~(cast<uint32_t>(3ULL)))){
case cast<uint32_t>(0ULL):{
auto tmp411 = std::make_tuple(cast<uint32_t>(16ULL),true);
dw = std::get<0>(tmp411);
written = std::get<1>(tmp411);
break;}
case cast<uint32_t>(8ULL):{
auto tmp412 = std::make_tuple(cast<uint32_t>(0ULL),true);
dw = std::get<0>(tmp412);
written = std::get<1>(tmp412);
break;}
case cast<uint32_t>(20ULL):{
auto tmp413 = std::make_tuple(get(m->usb.reg,cast<uint32_t>(4ULL)),true);
dw = std::get<0>(tmp413);
written = std::get<1>(tmp413);
break;}
case cast<uint32_t>(60ULL):{
auto tmp414 = std::make_tuple(cast<uint32_t>(cast<uint64_t>((xbox_Machine_usbFrame(m) & cast<uint64_t>(65535ULL)))),true);
dw = std::get<0>(tmp414);
written = std::get<1>(tmp414);
break;}
case cast<uint32_t>(56ULL):{
uint32_t fi = cast<uint32_t>((get(m->usb.reg,cast<uint32_t>(13ULL)) & cast<uint32_t>(16383ULL)));
if ((fi != cast<uint32_t>(0ULL))) {
uint64_t pos = modi<uint64_t>(m->tick,cast<uint64_t>(733466ULL));
uint32_t rem = cast<uint32_t>((fi - cast<uint32_t>(divi<uint64_t>(cast<uint64_t>((cast<uint64_t>(fi) * pos)),cast<uint64_t>(733466ULL)))));
dw = cast<uint32_t>((rem | shl<uint32_t>((cast<uint32_t>((get(m->usb.reg,cast<uint32_t>(15ULL)) & cast<uint32_t>(1ULL)))),cast<int64_t>(31ULL))));
}
written = true;
break;}
case cast<uint32_t>(72ULL):{
dw = cast<uint32_t>((((dw & ~(cast<uint32_t>(255ULL)))) | cast<uint32_t>(4ULL)));
written = true;
break;}
}}
return xbox_Machine_latchRead(m,(&m->usb),off,dw,written);
}
}
// tools/platform/xbox/usb.go:201:1
void xbox_Machine_usbWrite(xbox_Machine* m,uint32_t off,uint8_t v){
{
uint32_t reg = (off & ~(cast<uint32_t>(3ULL)));
uint32_t bits = shl<uint32_t>(cast<uint32_t>(v),(cast<uint32_t>((cast<uint32_t>(8ULL) * (cast<uint32_t>((off & cast<uint32_t>(3ULL))))))));
if ((cast<uint32_t>((off & cast<uint32_t>(3ULL))) == cast<uint32_t>(0ULL))) {
m->usbWrDword = cast<uint32_t>(0ULL);
}
m->usbWrDword |= bits;
uint32_t wrote = m->usbWrDword;
{
switch(reg){
case cast<uint32_t>(16ULL):{
m->usb.reg[cast<uint32_t>(4ULL)] |= bits;
xbox_Machine_latchTrace(m,(&m->usb),off,wrote);
xbox_Machine_usbUpdateIRQ(m);
return ;
break;}
case cast<uint32_t>(20ULL):{
m->usb.reg[cast<uint32_t>(4ULL)] &= ~(bits);
xbox_Machine_latchTrace(m,(&m->usb),off,wrote);
xbox_Machine_usbUpdateIRQ(m);
return ;
break;}
case cast<uint32_t>(12ULL):{
m->usb.reg[cast<uint32_t>(3ULL)] &= ~(bits);
xbox_Machine_latchTrace(m,(&m->usb),off,wrote);
xbox_Machine_usbUpdateIRQ(m);
return ;
break;}
case cast<uint32_t>(8ULL):{
xbox_Machine_latchTrace(m,(&m->usb),off,wrote);
return ;
break;}
case cast<uint32_t>(48ULL):{
xbox_Machine_latchWrite(m,(&m->usb),off,v);
return ;
break;}
}}
if (((reg >= cast<uint32_t>(84ULL)) && (reg < cast<uint32_t>(100ULL)))) {
xbox_Machine_usbPortWrite(m,divi<uint32_t>((cast<uint32_t>((reg - cast<uint32_t>(84ULL)))),cast<uint32_t>(4ULL)),bits);
xbox_Machine_latchTrace(m,(&m->usb),off,wrote);
return ;
}
xbox_Machine_latchWrite(m,(&m->usb),off,v);
}
}
// tools/platform/xbox/usb.go:266:1
void xbox_Machine_usbPortWrite(xbox_Machine* m,uint32_t port,uint32_t bits){
{
uint32_t off = cast<uint32_t>((cast<uint32_t>(84ULL) + cast<uint32_t>((port * cast<uint32_t>(4ULL)))));
uint32_t st = get(m->usb.reg,shr<uint32_t>(off,cast<int64_t>(2ULL)));
uint32_t was = st;
{
if ((cast<uint32_t>((bits & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL))){
st &= ~(cast<uint32_t>(2ULL));
}
else if ((cast<uint32_t>((bits & cast<uint32_t>(2ULL))) != cast<uint32_t>(0ULL))){
if ((cast<uint32_t>((st & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL))) {
st |= cast<uint32_t>(2ULL);
}
}
}
tmp415:;
if (((cast<uint32_t>((bits & cast<uint32_t>(4ULL))) != cast<uint32_t>(0ULL)) && (cast<uint32_t>((st & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL)))) {
st |= cast<uint32_t>(4ULL);
}
if ((cast<uint32_t>((bits & cast<uint32_t>(8ULL))) != cast<uint32_t>(0ULL))) {
st &= ~(cast<uint32_t>(4ULL));
}
if (((cast<uint32_t>((bits & cast<uint32_t>(16ULL))) != cast<uint32_t>(0ULL)) && (cast<uint32_t>((st & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL)))) {
st &= ~(cast<uint32_t>(16ULL));
st |= cast<uint32_t>(1048578ULL);
}
if ((cast<uint32_t>((bits & cast<uint32_t>(256ULL))) != cast<uint32_t>(0ULL))) {
st |= cast<uint32_t>(256ULL);
}
if ((cast<uint32_t>((bits & cast<uint32_t>(512ULL))) != cast<uint32_t>(0ULL))) {
st &= ~(cast<uint32_t>(258ULL));
}
st &= ~(cast<uint32_t>((bits & cast<uint32_t>(2031616ULL))));
m->usb.reg[shr<uint32_t>(off,cast<int64_t>(2ULL))] = st;
if ((cast<uint32_t>(((st & ~(was)) & cast<uint32_t>(2031616ULL))) != cast<uint32_t>(0ULL))) {
xbox_Machine_usbRaise(m,cast<uint32_t>(64ULL));
return ;
}
xbox_Machine_usbUpdateIRQ(m);
}
}
// tools/platform/xbox/usb.go:331:1
void xbox_Machine_usbSetPortConnected(xbox_Machine* m,uint32_t port,bool connected){
{
if ((port >= cast<uint32_t>(4ULL))) {
return ;
}
uint32_t off = cast<uint32_t>((cast<uint32_t>(84ULL) + cast<uint32_t>((port * cast<uint32_t>(4ULL)))));
uint32_t st = get(m->usb.reg,shr<uint32_t>(off,cast<int64_t>(2ULL)));
bool was = (cast<uint32_t>((st & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL));
if ((was == connected)) {
return ;
}
if (connected) {
st |= cast<uint32_t>(1ULL);
}
else {
st &= ~(cast<uint32_t>(3ULL));
}
st |= cast<uint32_t>(65536ULL);
m->usb.reg[shr<uint32_t>(off,cast<int64_t>(2ULL))] = st;
xbox_Machine_usbRaise(m,cast<uint32_t>(64ULL));
}
}
// tools/platform/xbox/usb.go:354:1
void xbox_Machine_usbRaise(xbox_Machine* m,uint32_t bit){
{
m->usb.reg[cast<uint32_t>(3ULL)] |= bit;
xbox_Machine_usbUpdateIRQ(m);
}
}
// tools/platform/xbox/usb.go:366:1
bool xbox_Machine_usbIRQ(xbox_Machine* m){
{
uint32_t en = get(m->usb.reg,cast<uint32_t>(4ULL));
if ((cast<uint32_t>((en & cast<uint32_t>(2147483648ULL))) == cast<uint32_t>(0ULL))) {
return false;
}
return ((cast<uint32_t>((get(m->usb.reg,cast<uint32_t>(3ULL)) & en)) & ~(cast<uint32_t>(2147483648ULL))) != cast<uint32_t>(0ULL));
}
}
// tools/platform/xbox/usb.go:376:1
void xbox_Machine_usbUpdateIRQ(xbox_Machine* m){
{
if (xbox_Machine_usbIRQ(m)) {
xbox_Machine_deliverPending(m);
}
}
}
// tools/platform/xbox/usb.go:390:1
void xbox_Machine_usbTick(xbox_Machine* m){
{
uint64_t now = xbox_Machine_usbFrame(m);
if (((m->usbFrameServed == cast<uint64_t>(0ULL)) || (now < m->usbFrameServed))) {
m->usbFrameServed = now;
}
{
uint64_t n = cast<uint64_t>((now - m->usbFrameServed));
if ((n > cast<uint64_t>(64ULL))) {
m->usbFrameServed = cast<uint64_t>((now - cast<uint64_t>(64ULL)));
}
}
{;for (;(m->usbFrameServed < now);){
m->usbFrameServed++;
xbox_Machine_usbFrameTick(m);
}
}if (xbox_Machine_usbIRQ(m)) {
xbox_Machine_deliverPending(m);
}
}
}
// tools/platform/xbox/usb.go:412:1
void xbox_Machine_usbFrameTick(xbox_Machine* m){
{
uint32_t ctrl = get(m->usb.reg,cast<uint32_t>(1ULL));
if ((cast<uint32_t>((ctrl & cast<uint32_t>(192ULL))) != cast<uint32_t>(128ULL))) {
return ;
}
xbox_Machine_usbFrameNumberToHCCA(m);
xbox_Machine_usbRaise(m,cast<uint32_t>(4ULL));
if ((cast<uint32_t>((ctrl & cast<uint32_t>(16ULL))) != cast<uint32_t>(0ULL))) {
xbox_Machine_usbWalkList(m,get(m->usb.reg,cast<uint32_t>(8ULL)));
}
if ((cast<uint32_t>((ctrl & cast<uint32_t>(32ULL))) != cast<uint32_t>(0ULL))) {
xbox_Machine_usbWalkList(m,get(m->usb.reg,cast<uint32_t>(10ULL)));
}
if ((cast<uint32_t>((ctrl & cast<uint32_t>(4ULL))) != cast<uint32_t>(0ULL))) {
xbox_Machine_usbWalkPeriodic(m);
}
xbox_Machine_usbWriteback(m);
}
}
// tools/platform/xbox/usb.go:461:1
void xbox_Machine_usbWalkPeriodic(xbox_Machine* m){
{
uint32_t hcca = get(m->usb.reg,cast<uint32_t>(6ULL));
if ((hcca == cast<uint32_t>(0ULL))) {
return ;
}
xbox_Machine_usbWalkList(m,xbox_Machine_read32(m,cast<uint32_t>((hcca + cast<uint32_t>(((cast<uint32_t>((cast<uint32_t>(xbox_Machine_usbFrame(m)) & cast<uint32_t>(31ULL)))) * cast<uint32_t>(4ULL)))))));
}
}
// tools/platform/xbox/usb_ohci.go:129:1
xbox_xidDevice* xbox_Machine_usbDeviceFor(xbox_Machine* m,uint32_t fa){
{
{auto&& tmp416 = m->usbDev;
for(int64_t tmp417=0;tmp417<len(tmp416);++tmp417){
auto i=tmp417;xbox_xidDevice* d = m->usbDev[i];
if ((bool(d) && (xbox_xidDevice_address(d) == fa))) {
return d;
}
}}
return {};
}
}
// tools/platform/xbox/usb_ohci.go:143:1
void xbox_Machine_usbWalkList(xbox_Machine* m,uint32_t head){
{
{auto tmp418 = std::make_tuple(cast<uint32_t>((head & cast<uint32_t>(4294967280ULL))),cast<int64_t>(0ULL));
uint32_t ed = std::get<0>(tmp418);
int64_t n = std::get<1>(tmp418);for (;((ed != cast<uint32_t>(0ULL)) && (n < cast<int64_t>(64ULL)));n++){
uint32_t next = cast<uint32_t>((xbox_Machine_read32(m,cast<uint32_t>((ed + cast<uint32_t>(12ULL)))) & cast<uint32_t>(4294967280ULL)));
xbox_Machine_usbRunEndpoint(m,ed);
ed = next;
}
}}
}
// tools/platform/xbox/usb_ohci.go:152:1
void xbox_Machine_usbRunEndpoint(xbox_Machine* m,uint32_t ed){
{
uint32_t ctrl = xbox_Machine_read32(m,cast<uint32_t>((ed + cast<uint32_t>(0ULL))));
if (((cast<uint32_t>((ctrl & cast<uint32_t>(16384ULL))) != cast<uint32_t>(0ULL)) || (cast<uint32_t>((ctrl & cast<uint32_t>(32768ULL))) != cast<uint32_t>(0ULL)))) {
return ;
}
uint32_t head = xbox_Machine_read32(m,cast<uint32_t>((ed + cast<uint32_t>(8ULL))));
if ((cast<uint32_t>((head & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL))) {
return ;
}
uint32_t tail = cast<uint32_t>((xbox_Machine_read32(m,cast<uint32_t>((ed + cast<uint32_t>(4ULL)))) & cast<uint32_t>(4294967280ULL)));
{int64_t n = cast<int64_t>(0ULL);for (;(n < cast<int64_t>(64ULL));n++){
uint32_t td = cast<uint32_t>((head & cast<uint32_t>(4294967280ULL)));
if (((td == cast<uint32_t>(0ULL)) || (td == tail))) {
return ;
}
auto tmp419 = xbox_Machine_usbRunTD(m,ed,ctrl,td);
uint32_t nextTD = std::get<0>(tmp419);
bool done = std::get<1>(tmp419);
if ((!done)) {
return ;
}
head = cast<uint32_t>((nextTD | (cast<uint32_t>((head & cast<uint32_t>(2ULL))))));
xbox_Machine_write32(m,cast<uint32_t>((ed + cast<uint32_t>(8ULL))),head);
}
}}
}
// tools/platform/xbox/usb_ohci.go:181:1
std::tuple<uint32_t,bool> xbox_Machine_usbRunTD(xbox_Machine* m,uint32_t ed,uint32_t edCtrl,uint32_t td){
uint32_t next{};
bool retired{};
{
uint32_t ctrl = xbox_Machine_read32(m,cast<uint32_t>((td + cast<uint32_t>(0ULL))));
uint32_t cbp = xbox_Machine_read32(m,cast<uint32_t>((td + cast<uint32_t>(4ULL))));
uint32_t nextTD = cast<uint32_t>((xbox_Machine_read32(m,cast<uint32_t>((td + cast<uint32_t>(8ULL)))) & cast<uint32_t>(4294967280ULL)));
uint32_t be = xbox_Machine_read32(m,cast<uint32_t>((td + cast<uint32_t>(12ULL))));
uint32_t fa = cast<uint32_t>((edCtrl & cast<uint32_t>(127ULL)));
uint32_t endpoint = shr<uint32_t>((cast<uint32_t>((edCtrl & cast<uint32_t>(1920ULL)))),cast<int64_t>(7ULL));
xbox_xidDevice* dev = xbox_Machine_usbDeviceFor(m,fa);
if ((!dev)) {
return {nextTD,false};
}
uint32_t length={};
if ((cbp != cast<uint32_t>(0ULL))) {
length = cast<uint32_t>((cast<uint32_t>((be - cbp)) + cast<uint32_t>(1ULL)));
}
{
switch(cast<uint32_t>((ctrl & cast<uint32_t>(1572864ULL)))){
case cast<uint32_t>(0ULL):{
Slice<uint8_t> pkt = Slice<uint8_t>::make(cast<int64_t>(8ULL));
{auto&& tmp420 = pkt;
for(int64_t tmp421=0;tmp421<len(tmp420);++tmp421){
auto i=tmp421;pkt[i] = xbox_Machine_Read(m,cast<uint32_t>((cbp + cast<uint32_t>(i))));
}}
auto tmp422 = xbox_xidDevice_setup(dev,m,pkt);
Slice<uint8_t> data = std::get<0>(tmp422);
Error err = std::get<1>(tmp422);
if (bool(err)) {
xbox_Machine_usbStall(m,ed,td,ctrl,cbp,length,nextTD);
return {nextTD,false};
}
auto tmp423 = std::make_tuple(data,cast<int64_t>(0ULL));
m->usbCtrlData = std::get<0>(tmp423);
m->usbCtrlOff = std::get<1>(tmp423);
xbox_Machine_usbRetire(m,td,ctrl,cbp,length,length,cast<uint32_t>(0ULL));
break;}
case cast<uint32_t>(1048576ULL):{
Slice<uint8_t> src={};
if ((endpoint == cast<uint32_t>(0ULL))) {
src = sub(m->usbCtrlData,gmin(m->usbCtrlOff,len(m->usbCtrlData)),len(m->usbCtrlData));
}
else {
src = xbox_xidDevice_interruptIn(dev,m,endpoint);
if ((!src)) {
return {nextTD,false};
}
}
uint32_t n = gmin(cast<uint32_t>(len(src)),length);
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < n);i++){
xbox_Machine_Write(m,cast<uint32_t>((cbp + i)),src[i]);
}
}if ((endpoint == cast<uint32_t>(0ULL))) {
m->usbCtrlOff += cast<int64_t>(n);
}
uint32_t cc = cast<uint32_t>(0ULL);
if (((n < length) && (cast<uint32_t>((ctrl & cast<uint32_t>(262144ULL))) == cast<uint32_t>(0ULL)))) {
cc = cast<uint32_t>(9ULL);
}
xbox_Machine_usbRetire(m,td,ctrl,cbp,n,length,cc);
if ((cc != cast<uint32_t>(0ULL))) {
xbox_Machine_usbHaltEndpoint(m,ed,nextTD);
return {nextTD,false};
}
xbox_Machine_usbControlStatus(m,dev,endpoint,length);
break;}
case cast<uint32_t>(524288ULL):{
xbox_Machine_usbRetire(m,td,ctrl,cbp,length,length,cast<uint32_t>(0ULL));
xbox_Machine_usbControlStatus(m,dev,endpoint,length);
break;}
default:{
x86_CPU_Halt(m->CPU,std::string("USB: TD %08X has reserved direction %d (ED %08X)",48),td,shr<uint32_t>((cast<uint32_t>((ctrl & cast<uint32_t>(1572864ULL)))),cast<int64_t>(19ULL)),ed);
auto tmp424 = std::make_tuple(true,m->CPU->HaltReason);
m->Halted = std::get<0>(tmp424);
m->HaltReason = std::get<1>(tmp424);
return {nextTD,false};
break;}
}}
return {nextTD,true};
}
}
// tools/platform/xbox/usb_ohci.go:266:1
void xbox_Machine_usbControlStatus(xbox_Machine* m,xbox_xidDevice* dev,uint32_t endpoint,uint32_t length){
{
if (((endpoint == cast<uint32_t>(0ULL)) && (length == cast<uint32_t>(0ULL)))) {
xbox_xidDevice_controlStatusDone(dev,m);
}
}
}
// tools/platform/xbox/usb_ohci.go:280:1
void xbox_Machine_usbRetire(xbox_Machine* m,uint32_t td,uint32_t ctrl,uint32_t cbp,uint32_t moved,uint32_t length,uint32_t cc){
{
ctrl = cast<uint32_t>((((ctrl & ~(cast<uint32_t>(4026531840ULL)))) | (shl<uint32_t>(cc,cast<int64_t>(28ULL)))));
xbox_Machine_write32(m,cast<uint32_t>((td + cast<uint32_t>(0ULL))),ctrl);
xbox_Machine_write32(m,cast<uint32_t>((td + cast<uint32_t>(4ULL))),xbox_usbCurrentBuffer(cbp,moved,length));
m->usbDone = append(m->usbDone,Slice<uint32_t>{td});
}
}
// tools/platform/xbox/usb_ohci.go:308:1
uint32_t xbox_usbCurrentBuffer(uint32_t cbp,uint32_t moved,uint32_t length){
{
if ((moved >= length)) {
return cast<uint32_t>(0ULL);
}
return cast<uint32_t>((cbp + moved));
}
}
// tools/platform/xbox/usb_ohci.go:322:1
void xbox_Machine_usbStall(xbox_Machine* m,uint32_t ed,uint32_t td,uint32_t ctrl,uint32_t cbp,uint32_t length,uint32_t nextTD){
{
xbox_Machine_usbRetire(m,td,ctrl,cbp,cast<uint32_t>(0ULL),length,cast<uint32_t>(4ULL));
xbox_Machine_usbHaltEndpoint(m,ed,nextTD);
}
}
// tools/platform/xbox/usb_ohci.go:353:1
void xbox_Machine_usbHaltEndpoint(xbox_Machine* m,uint32_t ed,uint32_t nextTD){
{
uint32_t head = xbox_Machine_read32(m,cast<uint32_t>((ed + cast<uint32_t>(8ULL))));
xbox_Machine_write32(m,cast<uint32_t>((ed + cast<uint32_t>(8ULL))),cast<uint32_t>((cast<uint32_t>((nextTD | (cast<uint32_t>((head & cast<uint32_t>(2ULL)))))) | cast<uint32_t>(1ULL))));
}
}
// tools/platform/xbox/usb_ohci.go:363:1
void xbox_Machine_usbWriteback(xbox_Machine* m){
{
if ((len(m->usbDone) == cast<int64_t>(0ULL))) {
return ;
}
uint32_t head={};
{auto&& tmp425 = m->usbDone;
for(int64_t tmp426=0;tmp426<len(tmp425);++tmp426){
auto td=tmp425[tmp426];xbox_Machine_write32(m,cast<uint32_t>((td + cast<uint32_t>(8ULL))),head);
head = td;
}}
m->usbDone = sub(m->usbDone,0,cast<int64_t>(0ULL));
{
uint32_t hcca = get(m->usb.reg,cast<uint32_t>(6ULL));
if ((hcca != cast<uint32_t>(0ULL))) {
xbox_Machine_write32(m,cast<uint32_t>((hcca + cast<uint32_t>(132ULL))),head);
}
}
m->usb.reg[cast<uint32_t>(12ULL)] = head;
xbox_Machine_usbRaise(m,cast<uint32_t>(2ULL));
}
}
// tools/platform/xbox/usb_ohci.go:385:1
void xbox_Machine_usbFrameNumberToHCCA(xbox_Machine* m){
{
{
uint32_t hcca = get(m->usb.reg,cast<uint32_t>(6ULL));
if ((hcca != cast<uint32_t>(0ULL))) {
xbox_Machine_write16(m,cast<uint32_t>((hcca + cast<uint32_t>(128ULL))),cast<uint16_t>(xbox_Machine_usbFrame(m)));
}
}
}
}
// tools/platform/xbox/usb_xid.go:213:1
std::tuple<xbox_PadControl,bool> xbox_PadControlByName(std::string name){
{
auto tmp427 = lookup(xbox_padControls,name);
xbox_PadControl c = std::get<0>(tmp427);
bool ok = std::get<1>(tmp427);
return {c,ok};
}
}
// tools/platform/xbox/usb_xid.go:220:1
Slice<std::string> xbox_PadControlNames(){
{
Slice<std::string> names = Slice<std::string>::make(cast<int64_t>(0ULL),len(xbox_padControls));
{auto&& tmp428 = xbox_padControls;
for(auto [tmp429,tmp430]:tmp428){
auto n=tmp429;names = append(names,Slice<std::string>{n});
}}
go_sort_Strings(names);
return names;
}
}
// tools/platform/xbox/usb_xid.go:299:1
xbox_PadState xbox_PadStateOf(Map<std::string,bool> held){
{
xbox_PadState s={};
{auto&& tmp431 = held;
for(auto [tmp432,tmp433]:tmp431){
auto n=tmp432;auto down=tmp433;if ((!down)) {
continue;
}
auto tmp434 = lookup(xbox_padControls,n);
xbox_PadControl c = std::get<0>(tmp434);
bool ok = std::get<1>(tmp434);
if ((!ok)) {
continue;
}
{
switch(c.Kind){
case cast<xbox_PadControlKind>(0ULL):{
s.Buttons |= c.Bit;
break;}
case cast<xbox_PadControlKind>(1ULL):{
s.Analog[c.Index] = cast<uint8_t>(255ULL);
break;}
case cast<xbox_PadControlKind>(2ULL):{
s.Axes[c.Index] += cast<int16_t>(c.Sign);
break;}
}}
}}
{auto&& tmp435 = s.Axes;
for(int64_t tmp436=0;tmp436<len(tmp435);++tmp436){
auto i=tmp436;auto v=tmp435[tmp436];{
if ((v > cast<int16_t>(0ULL))){
s.Axes[i] = cast<int16_t>(32767ULL);
}
else if ((v < cast<int16_t>(0ULL))){
s.Axes[i] = cast<int16_t>(-32767ULL);
}
else {
s.Axes[i] = cast<int16_t>(0ULL);
}
}
tmp437:;
}}
return s;
}
}
// tools/platform/xbox/usb_xid.go:348:1
void xbox_Machine_AttachPad(xbox_Machine* m,int64_t port){
{
if ((((port < cast<int64_t>(0ULL)) || (port >= cast<int64_t>(4ULL))) || bool(m->usbDev[port]))) {
return ;
}
m->usbDev[port] = arenaNew(xbox_xidDevice{});
xbox_Machine_usbSetPortConnected(m,cast<uint32_t>(port),true);
}
}
// tools/platform/xbox/usb_xid.go:362:1
void xbox_Machine_SetPadButtons(xbox_Machine* m,int64_t port,uint16_t buttons){
{
if (((port < cast<int64_t>(0ULL)) || (port >= cast<int64_t>(4ULL)))) {
return ;
}
auto tmp438 = std::make_tuple(m->usbDev[port],m->usbDev[port]!=nullptr);
xbox_xidDevice* d = std::get<0>(tmp438);
if ((!d)) {
return ;
}
if ((d->Buttons != buttons)) {
auto tmp439 = std::make_tuple(buttons,true);
d->Buttons = std::get<0>(tmp439);
d->Fresh = std::get<1>(tmp439);
}
}
}
// tools/platform/xbox/usb_xid.go:376:1
xbox_xidDevice* xbox_Machine_pad(xbox_Machine* m,int64_t port){
{
if (((port < cast<int64_t>(0ULL)) || (port >= cast<int64_t>(4ULL)))) {
return {};
}
auto tmp440 = std::make_tuple(m->usbDev[port],m->usbDev[port]!=nullptr);
xbox_xidDevice* d = std::get<0>(tmp440);
return d;
}
}
// tools/platform/xbox/usb_xid.go:392:1
void xbox_Machine_SetPadAnalog(xbox_Machine* m,int64_t port,int64_t i,uint8_t pressure){
{
xbox_xidDevice* d = xbox_Machine_pad(m,port);
if ((((!d) || (i < cast<int64_t>(0ULL))) || (i >= cast<int64_t>(8ULL)))) {
return ;
}
if ((d->Analog[i] != pressure)) {
auto tmp441 = std::make_tuple(pressure,true);
d->Analog[i] = std::get<0>(tmp441);
d->Fresh = std::get<1>(tmp441);
}
}
}
// tools/platform/xbox/usb_xid.go:403:1
uint8_t xbox_Machine_PadAnalog(xbox_Machine* m,int64_t port,int64_t i){
{
xbox_xidDevice* d = xbox_Machine_pad(m,port);
if ((((!d) || (i < cast<int64_t>(0ULL))) || (i >= cast<int64_t>(8ULL)))) {
return cast<uint8_t>(0ULL);
}
return d->Analog[i];
}
}
// tools/platform/xbox/usb_xid.go:415:1
void xbox_Machine_SetPadAxis(xbox_Machine* m,int64_t port,int64_t i,int16_t v){
{
xbox_xidDevice* d = xbox_Machine_pad(m,port);
if ((((!d) || (i < cast<int64_t>(0ULL))) || (i >= cast<int64_t>(4ULL)))) {
return ;
}
if ((d->Axes[i] != v)) {
auto tmp442 = std::make_tuple(v,true);
d->Axes[i] = std::get<0>(tmp442);
d->Fresh = std::get<1>(tmp442);
}
}
}
// tools/platform/xbox/usb_xid.go:426:1
int16_t xbox_Machine_PadAxis(xbox_Machine* m,int64_t port,int64_t i){
{
xbox_xidDevice* d = xbox_Machine_pad(m,port);
if ((((!d) || (i < cast<int64_t>(0ULL))) || (i >= cast<int64_t>(4ULL)))) {
return cast<int16_t>(0ULL);
}
return d->Axes[i];
}
}
// tools/platform/xbox/usb_xid.go:438:1
void xbox_Machine_SetPad(xbox_Machine* m,int64_t port,xbox_PadState s){
{
xbox_Machine_SetPadButtons(m,port,s.Buttons);
{auto&& tmp443 = s.Analog;
for(int64_t tmp444=0;tmp444<len(tmp443);++tmp444){
auto i=tmp444;auto p=tmp443[tmp444];xbox_Machine_SetPadAnalog(m,port,i,p);
}}
{auto&& tmp445 = s.Axes;
for(int64_t tmp446=0;tmp446<len(tmp445);++tmp446){
auto i=tmp446;auto v=tmp445[tmp446];xbox_Machine_SetPadAxis(m,port,i,v);
}}
}
}
// tools/platform/xbox/usb_xid.go:449:1
uint16_t xbox_Machine_PadButtons(xbox_Machine* m,int64_t port){
{
if (((port < cast<int64_t>(0ULL)) || (port >= cast<int64_t>(4ULL)))) {
return cast<uint16_t>(0ULL);
}
{
auto tmp447 = std::make_tuple(m->usbDev[port],m->usbDev[port]!=nullptr);
xbox_xidDevice* d = std::get<0>(tmp447);
if (bool(d)) {
return d->Buttons;
}
}
return cast<uint16_t>(0ULL);
}
}
// tools/platform/xbox/usb_xid.go:495:1
uint32_t xbox_xidDevice_address(xbox_xidDevice* d){
{
return d->Addr;
}
}
// tools/platform/xbox/usb_xid.go:499:1
void xbox_xidDevice_controlStatusDone(xbox_xidDevice* d,xbox_Machine* m){
{
if (d->AddrArmed) {
auto tmp448 = std::make_tuple(d->AddrNext,false);
d->Addr = std::get<0>(tmp448);
d->AddrArmed = std::get<1>(tmp448);
}
}
}
// tools/platform/xbox/usb_xid.go:532:1
std::tuple<Slice<uint8_t>,Error> xbox_xidDevice_setup(xbox_xidDevice* d,xbox_Machine* m,Slice<uint8_t> pkt){
{
uint8_t bmType = pkt[cast<int64_t>(0ULL)];
uint8_t bReq = pkt[cast<int64_t>(1ULL)];
uint16_t wValue = cast<uint16_t>((cast<uint16_t>(pkt[cast<int64_t>(2ULL)]) | shl<uint16_t>(cast<uint16_t>(pkt[cast<int64_t>(3ULL)]),cast<int64_t>(8ULL))));
if ((((bmType == cast<uint8_t>(193ULL)) && (bReq == cast<uint8_t>(6ULL))) && (shr<uint16_t>(wValue,cast<int64_t>(8ULL)) == cast<uint16_t>(66ULL)))) {
return {sub(xbox_xidDescriptor,0,len(xbox_xidDescriptor)),{}};
}
if ((((bmType == cast<uint8_t>(161ULL)) && (bReq == cast<uint8_t>(1ULL))) && (wValue == cast<uint16_t>(256ULL)))) {
std::array<uint8_t,20> r = xbox_xidDevice_report(d);
return {sub(r,0,len(r)),{}};
}
if ((((bmType == cast<uint8_t>(193ULL)) && (bReq == cast<uint8_t>(1ULL))) && (((wValue == cast<uint16_t>(256ULL)) || (wValue == cast<uint16_t>(512ULL)))))) {
int64_t n = cast<int64_t>(20ULL);
if ((wValue == cast<uint16_t>(512ULL))) {
n = cast<int64_t>(6ULL);
}
Slice<uint8_t> r = Slice<uint8_t>::make(n);
{auto&& tmp449 = r;
for(int64_t tmp450=0;tmp450<len(tmp449);++tmp450){
auto i=tmp450;r[i] = cast<uint8_t>(cast<int64_t>((cast<int64_t>(208ULL) + i)));
}}
return {r,{}};
}
if ((cast<uint8_t>((bmType & cast<uint8_t>(96ULL))) != cast<uint8_t>(0ULL))) {
return {{},xbox_Machine_usbUnsupported(m,std::string("unmodelled control request bmRequestType=%02X bRequest=%02X wValue=%04X",71),bmType,bReq,wValue)};
}
{
switch(bReq){
case cast<uint8_t>(5ULL):{
auto tmp451 = std::make_tuple(cast<uint32_t>((cast<uint32_t>(wValue) & cast<uint32_t>(127ULL))),true);
d->AddrNext = std::get<0>(tmp451);
d->AddrArmed = std::get<1>(tmp451);
return {{},{}};
break;}
case cast<uint8_t>(9ULL):{
d->Config = cast<uint32_t>((cast<uint32_t>(wValue) & cast<uint32_t>(255ULL)));
return {{},{}};
break;}
case cast<uint8_t>(8ULL):{
return {Slice<uint8_t>{cast<uint8_t>(d->Config)},{}};
break;}
case cast<uint8_t>(6ULL):{
return xbox_xidDevice_descriptor(d,m,shr<uint16_t>(wValue,cast<int64_t>(8ULL)),cast<uint8_t>(wValue));
break;}
}}
return {{},xbox_Machine_usbUnsupported(m,std::string("unmodelled standard control request bmRequestType=%02X bRequest=%02X wValue=%04X",80),bmType,bReq,wValue)};
}
}
// tools/platform/xbox/usb_xid.go:678:1
std::tuple<Slice<uint8_t>,Error> xbox_xidDevice_descriptor(xbox_xidDevice* d,xbox_Machine* m,uint16_t dtype,uint8_t index){
{
{
switch(dtype){
case cast<uint16_t>(1ULL):{
return {sub(xbox_deviceDescriptor,0,len(xbox_deviceDescriptor)),{}};
break;}
case cast<uint16_t>(2ULL):{
return {sub(xbox_configDescriptor,0,len(xbox_configDescriptor)),{}};
break;}
}}
return {{},xbox_Machine_usbUnsupported(m,std::string("unmodelled GET_DESCRIPTOR type=%02X index=%02X \342\200\224 read XAPI's comparison before answering",90),dtype,index)};
}
}
// tools/platform/xbox/usb_xid.go:1051:1
Slice<uint8_t> xbox_xidDevice_interruptIn(xbox_xidDevice* d,xbox_Machine* m,uint32_t endpoint){
{
if ((!d->Fresh)) {
return {};
}
d->Fresh = false;
std::array<uint8_t,20> r = xbox_xidDevice_report(d);
if ((r == d->SentReport)) {
return {};
}
d->SentReport = r;
return sub(r,0,len(r));
}
}
// tools/platform/xbox/usb_xid.go:1070:1
std::array<uint8_t,20> xbox_xidDevice_report(xbox_xidDevice* d){
{
std::array<uint8_t,20> r={};
auto tmp452 = std::make_tuple(cast<uint8_t>(245ULL),cast<uint8_t>(246ULL));
r[cast<int64_t>(0ULL)] = std::get<0>(tmp452);
r[cast<int64_t>(1ULL)] = std::get<1>(tmp452);
auto tmp453 = std::make_tuple(cast<uint8_t>(d->Buttons),cast<uint8_t>(shr<uint16_t>(d->Buttons,cast<int64_t>(8ULL))));
r[cast<int64_t>(2ULL)] = std::get<0>(tmp453);
r[cast<int64_t>(3ULL)] = std::get<1>(tmp453);
gcopy(sub(r,cast<int64_t>(4ULL),cast<int64_t>(12ULL)),sub(d->Analog,0,len(d->Analog)));
{auto&& tmp454 = d->Axes;
for(int64_t tmp455=0;tmp455<len(tmp454);++tmp455){
auto i=tmp455;auto v=tmp454[tmp455];r[cast<int64_t>((cast<int64_t>(12ULL) + cast<int64_t>((cast<int64_t>(2ULL) * i))))] = cast<uint8_t>(cast<uint16_t>(v));
r[cast<int64_t>((cast<int64_t>(13ULL) + cast<int64_t>((cast<int64_t>(2ULL) * i))))] = cast<uint8_t>(shr<uint16_t>(cast<uint16_t>(v),cast<int64_t>(8ULL)));
}}
return r;
}
}
// tools/platform/xbox/xbe.go:68:1
std::string xbox_Section_FlagString(xbox_Section s){
{
Slice<std::string> f={};
if ((cast<uint32_t>((s.Flags & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL))) {
f = append(f,Slice<std::string>{std::string("W",1)});
}
if ((cast<uint32_t>((s.Flags & cast<uint32_t>(4ULL))) != cast<uint32_t>(0ULL))) {
f = append(f,Slice<std::string>{std::string("X",1)});
}
if ((cast<uint32_t>((s.Flags & cast<uint32_t>(2ULL))) != cast<uint32_t>(0ULL))) {
f = append(f,Slice<std::string>{std::string("preload",7)});
}
if ((cast<uint32_t>((s.Flags & cast<uint32_t>(8ULL))) != cast<uint32_t>(0ULL))) {
f = append(f,Slice<std::string>{std::string("inserted",8)});
}
if ((len(f) == cast<int64_t>(0ULL))) {
return std::string("-",1);
}
return go_strings_Join(f,std::string("|",1));
}
}
// tools/platform/xbox/xbe.go:105:1
std::tuple<xbox_XBE*,Error> xbox_ParseXBE(Slice<uint8_t> b){
{
if ((len(b) < cast<int64_t>(376ULL))) {
return {{},go_fmt_Errorf(std::string("xbe: too small (%d bytes) to hold a header",42),len(b))};
}
if ((cast<std::string>(sub(b,cast<int64_t>(0ULL),cast<int64_t>(4ULL))) != xbox_xbeMagic)) {
return {{},go_fmt_Errorf(std::string("xbe: bad magic %q, want %q",26),sub(b,cast<int64_t>(0ULL),cast<int64_t>(4ULL)),xbox_xbeMagic)};
}
xbox_XBE* x = arenaNew(xbox_XBE{{},{},{},{},{},{},{},{},{},b});
x->Base = xbox_le32(rrBorrow(b,cast<int64_t>(260ULL),len(b)));
x->ImageSize = xbox_le32(rrBorrow(b,cast<int64_t>(268ULL),len(b)));
uint32_t entryRaw = xbox_le32(rrBorrow(b,cast<int64_t>(296ULL),len(b)));
uint32_t thunkRaw = xbox_le32(rrBorrow(b,cast<int64_t>(344ULL),len(b)));
uint32_t certAddr = xbox_le32(rrBorrow(b,cast<int64_t>(280ULL),len(b)));
uint32_t nSections = xbox_le32(rrBorrow(b,cast<int64_t>(284ULL),len(b)));
uint32_t secHdrAddr = xbox_le32(rrBorrow(b,cast<int64_t>(288ULL),len(b)));
{
Error err = xbox_XBE_parseSections(x,nSections,secHdrAddr);
if (bool(err)) {
return {{},err};
}
}
{
uint32_t e = cast<uint32_t>((entryRaw ^ cast<uint32_t>(2835109803ULL)));
if (xbox_XBE_inImage(x,e)) {
auto tmp456 = std::make_tuple(e,true);
x->Entry = std::get<0>(tmp456);
x->Retail = std::get<1>(tmp456);
}
else {
uint32_t e = cast<uint32_t>((entryRaw ^ cast<uint32_t>(2491784523ULL)));
if (xbox_XBE_inImage(x,e)) {
auto tmp457 = std::make_tuple(e,false);
x->Entry = std::get<0>(tmp457);
x->Retail = std::get<1>(tmp457);
}
else {
x->Entry = cast<uint32_t>((entryRaw ^ cast<uint32_t>(2835109803ULL)));
}
}
}
{
uint32_t t = cast<uint32_t>((thunkRaw ^ cast<uint32_t>(1533886646ULL)));
if (xbox_XBE_inImage(x,t)) {
x->ThunkAddr = t;
}
else {
uint32_t t = cast<uint32_t>((thunkRaw ^ cast<uint32_t>(4021416274ULL)));
if (xbox_XBE_inImage(x,t)) {
x->ThunkAddr = t;
}
else {
x->ThunkAddr = cast<uint32_t>((thunkRaw ^ cast<uint32_t>(1533886646ULL)));
}
}
}
xbox_XBE_parseCertificate(x,certAddr);
{
Error err = xbox_XBE_parseThunks(x);
if (bool(err)) {
return {{},err};
}
}
return {x,{}};
}
}
// tools/platform/xbox/xbe.go:154:1
bool xbox_XBE_inImage(xbox_XBE* x,uint32_t va){
{
return ((va >= x->Base) && (va < cast<uint32_t>((x->Base + x->ImageSize))));
}
}
// tools/platform/xbox/xbe.go:161:1
std::tuple<int64_t,bool> xbox_XBE_atVA(xbox_XBE* x,uint32_t va){
{
{auto&& tmp458 = x->Sections;
for(int64_t tmp459=0;tmp459<len(tmp458);++tmp459){
auto s=tmp458[tmp459];if (((va >= s.VAddr) && (va < cast<uint32_t>((s.VAddr + s.RawSize))))) {
return {cast<int64_t>(cast<uint32_t>((s.RawAddr + (cast<uint32_t>((va - s.VAddr)))))),true};
}
}}
if ((va >= x->Base)) {
int64_t off = cast<int64_t>(cast<uint32_t>((va - x->Base)));
if ((off < len(x->raw))) {
return {off,true};
}
}
return {cast<int64_t>(0ULL),false};
}
}
// tools/platform/xbox/xbe.go:180:1
Error xbox_XBE_parseSections(xbox_XBE* x,uint32_t n,uint32_t addr){
{
if (((n == cast<uint32_t>(0ULL)) || (n > cast<uint32_t>(4096ULL)))) {
return go_fmt_Errorf(std::string("xbe: implausible section count %d",33),n);
}
auto tmp460 = xbox_XBE_atVA(x,addr);
int64_t off = std::get<0>(tmp460);
bool ok = std::get<1>(tmp460);
if ((!ok)) {
if ((addr < x->Base)) {
return go_fmt_Errorf(std::string("xbe: section headers at %#x precede the image base %#x",54),addr,x->Base);
}
off = cast<int64_t>(cast<uint32_t>((addr - x->Base)));
}
constexpr int64_t secHdrSize=56ULL;
if ((cast<int64_t>((off + cast<int64_t>((cast<int64_t>(n) * cast<int64_t>(56ULL))))) > len(x->raw))) {
return go_fmt_Errorf(std::string("xbe: %d section headers at %#x overrun the %d-byte image",56),n,off,len(x->raw));
}
x->Sections = Slice<xbox_Section>::make(n);
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(n));i++){
Slice<uint8_t> h = sub(x->raw,cast<int64_t>((off + cast<int64_t>((i * cast<int64_t>(56ULL))))),len(x->raw));
xbox_Section s = xbox_Section{{},xbox_le32(rrBorrow(h,cast<int64_t>(0ULL),len(h))),xbox_le32(rrBorrow(h,cast<int64_t>(4ULL),len(h))),xbox_le32(rrBorrow(h,cast<int64_t>(8ULL),len(h))),xbox_le32(rrBorrow(h,cast<int64_t>(12ULL),len(h))),xbox_le32(rrBorrow(h,cast<int64_t>(16ULL),len(h))),xbox_le32(rrBorrow(h,cast<int64_t>(20ULL),len(h)))};
x->Sections[i] = s;
}
}{auto&& tmp461 = x->Sections;
for(int64_t tmp462=0;tmp462<len(tmp461);++tmp462){
auto i=tmp462;x->Sections[i].Name = xbox_XBE_cstrVA(x,x->Sections[i].NameAddr);
}}
return {};
}
}
// tools/platform/xbox/xbe.go:218:1
std::string xbox_XBE_cstrVA(xbox_XBE* x,uint32_t va){
{
auto tmp463 = xbox_XBE_atVA(x,va);
int64_t off = std::get<0>(tmp463);
bool ok = std::get<1>(tmp463);
if ((!ok)) {
return std::string("",0);
}
int64_t end = off;
{;for (;((end < len(x->raw)) && (x->raw[end] != cast<uint8_t>(0ULL)));){
end++;
}
}return cast<std::string>(sub(x->raw,off,end));
}
}
// tools/platform/xbox/xbe.go:231:1
void xbox_XBE_parseCertificate(xbox_XBE* x,uint32_t va){
{
auto tmp464 = xbox_XBE_atVA(x,va);
int64_t off = std::get<0>(tmp464);
bool ok = std::get<1>(tmp464);
if (((!ok) || (cast<int64_t>((cast<int64_t>((off + cast<int64_t>(12ULL))) + cast<int64_t>(80ULL))) > len(x->raw)))) {
return ;
}
x->TitleID = xbox_le32(rrBorrow(x->raw,cast<int64_t>((off + cast<int64_t>(8ULL))),len(x->raw)));
Slice<uint16_t> u = Slice<uint16_t>::make(cast<int64_t>(0ULL),cast<int64_t>(40ULL));
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(40ULL));i++){
uint16_t c = xbox_le16(rrBorrow(x->raw,cast<int64_t>((cast<int64_t>((off + cast<int64_t>(12ULL))) + cast<int64_t>((i * cast<int64_t>(2ULL))))),len(x->raw)));
if ((c == cast<uint16_t>(0ULL))) {
break;
}
u = append(u,Slice<uint16_t>{c});
}
}x->TitleName = go_strings_TrimRight(cast<std::string>(go_utf16_Decode(u)),std::string("\000 ",2));
}
}
// tools/platform/xbox/xbe.go:250:1
Error xbox_XBE_parseThunks(xbox_XBE* x){
{
auto tmp465 = xbox_XBE_atVA(x,x->ThunkAddr);
int64_t off = std::get<0>(tmp465);
bool ok = std::get<1>(tmp465);
if ((!ok)) {
return go_fmt_Errorf(std::string("xbe: kernel thunk table VA %#x is not inside the image",54),x->ThunkAddr);
}
Map<uint16_t,bool> set = Map<uint16_t,bool>{};
{int64_t p = off;for (;(cast<int64_t>((p + cast<int64_t>(4ULL))) <= len(x->raw));p += cast<int64_t>(4ULL)){
uint32_t v = xbox_le32(rrBorrow(x->raw,p,len(x->raw)));
if ((v == cast<uint32_t>(0ULL))) {
break;
}
if ((cast<uint32_t>((v & cast<uint32_t>(2147483648ULL))) != cast<uint32_t>(0ULL))) {
set[cast<uint16_t>(cast<uint32_t>((v & cast<uint32_t>(65535ULL))))] = true;
}
}
}x->Ordinals = Slice<uint16_t>::make(cast<int64_t>(0ULL),len(set));
{auto&& tmp466 = set;
for(auto [tmp467,tmp468]:tmp466){
auto o=tmp467;x->Ordinals = append(x->Ordinals,Slice<uint16_t>{o});
}}
go_sort_Slice(x->Ordinals,[&](int64_t i,int64_t j)->bool{
return (x->Ordinals[i] < x->Ordinals[j]);
}
);
return {};
}
}
// tools/platform/xbox/xbe.go:275:1
uint16_t xbox_le16(Slice<uint8_t> b){
{
return le_Uint16(sub(b,0,cast<int64_t>(2ULL)));
}
}
// tools/platform/xbox/xbe.go:276:1
uint32_t xbox_le32(Slice<uint8_t> b){
{
return le_Uint32(rrBorrow(b,0,cast<int64_t>(4ULL)));
}
}
// tools/platform/xbox/xiso.go:88:1
std::string xbox_Entry_String(xbox_Entry e){
{
if (e.IsDir) {
return go_fmt_Sprintf(std::string("%-40s  <dir>   (sector %d)",26),e.Path,e.Sector);
}
return go_fmt_Sprintf(std::string("%-40s  %10d  (sector %d)",24),e.Path,e.Size,e.Sector);
}
}
// tools/platform/xbox/xiso.go:100:1
int64_t xbox_Entry_Offset(xbox_Entry e){
{
return cast<int64_t>((cast<int64_t>(e.Sector) * cast<int64_t>(2048ULL)));
}
}
// tools/platform/xbox/xiso.go:110:1
std::tuple<xbox_Entry,int64_t,bool> xbox_Image_EntryAt(xbox_Image* img,int64_t off){
{
if ((!img->index)) {
img->index = Slice<xbox_Entry>{};
(void)(xbox_Image_Walk(img,[&](xbox_Entry e)->Error{
if (((!e.IsDir) && (e.Size > cast<uint32_t>(0ULL)))) {
img->index = append(img->index,Slice<xbox_Entry>{e});
}
return {};
}
));
go_sort_Slice(img->index,[&](int64_t i,int64_t j)->bool{
return (xbox_Entry_Offset(img->index[i]) < xbox_Entry_Offset(img->index[j]));
}
);
}
auto tmp469 = go_sort_Search(len(img->index),[&](int64_t i)->bool{
return (xbox_Entry_Offset(img->index[i]) > off);
}
);
int64_t i = tmp469;
if ((i == cast<int64_t>(0ULL))) {
return {xbox_Entry{},cast<int64_t>(0ULL),false};
}
xbox_Entry e = img->index[cast<int64_t>((i - cast<int64_t>(1ULL)))];
int64_t within = cast<int64_t>((off - xbox_Entry_Offset(e)));
if ((within >= cast<int64_t>(e.Size))) {
return {xbox_Entry{},cast<int64_t>(0ULL),false};
}
return {e,within,true};
}
}
// tools/platform/xbox/xiso.go:135:1
std::tuple<xbox_Image*,Error> xbox_Open(std::string path){
{
auto tmp470 = go_os_Open(path);
os_File* f = std::get<0>(tmp470);
Error err = std::get<1>(tmp470);
if (bool(err)) {
return {{},err};
}
auto tmp471 = os_File_Stat(f);
RRFileInfo st = std::get<0>(tmp471);
err = std::get<1>(tmp471);
if (bool(err)) {
os_File_Close(f);
return {{},err};
}
xbox_Image* img = arenaNew(xbox_Image{path,rrFileInfo_Size(st),{},{},{},{},f,{},{}});
auto tmp472 = xbox_Image_findVolumeDescriptor(img);
int64_t base = std::get<0>(tmp472);
Slice<uint8_t> vd = std::get<1>(tmp472);
err = std::get<2>(tmp472);
if (bool(err)) {
os_File_Close(f);
return {{},err};
}
img->Base = base;
img->rootSector = xbox_le32(rrBorrow(vd,cast<int64_t>(20ULL),len(vd)));
img->rootSize = xbox_le32(rrBorrow(vd,cast<int64_t>(24ULL),len(vd)));
img->CreationTime = cast<uint64_t>((cast<uint64_t>(xbox_le32(rrBorrow(vd,cast<int64_t>(28ULL),len(vd)))) | shl<uint64_t>(cast<uint64_t>(xbox_le32(rrBorrow(vd,cast<int64_t>(32ULL),len(vd)))),cast<int64_t>(32ULL))));
return {img,{}};
}
}
// tools/platform/xbox/xiso.go:166:1
std::tuple<int64_t,Slice<uint8_t>,Error> xbox_Image_findVolumeDescriptor(xbox_Image* img){
int64_t base{};
Slice<uint8_t> vd{};
Error err{};
{
auto valid = [&](Slice<uint8_t> sector)->bool{
return (((len(sector) == cast<int64_t>(2048ULL)) && (cast<std::string>(sub(sector,cast<int64_t>(0ULL),cast<int64_t>(20ULL))) == xbox_xdvdfsMagic)) && (cast<std::string>(sub(sector,cast<int64_t>(2028ULL),len(sector))) == xbox_xdvdfsMagic));
}
;
{auto&& tmp473 = xbox_candidateBases;
for(int64_t tmp474=0;tmp474<len(tmp473);++tmp474){
auto b=tmp473[tmp474];int64_t off = cast<int64_t>((b + cast<int64_t>(65536ULL)));
if ((cast<int64_t>((off + cast<int64_t>(2048ULL))) > img->Size)) {
continue;
}
auto tmp475 = xbox_Image_readAt(img,off,cast<int64_t>(2048ULL));
Slice<uint8_t> sec = std::get<0>(tmp475);
Error e = std::get<1>(tmp475);
if (((!e) && valid(sec))) {
return {b,sec,{}};
}
}}
constexpr int64_t chunk=8388608ULL;
Slice<uint8_t> buf = Slice<uint8_t>::make(cast<int64_t>(8388608ULL));
{int64_t off = cast<int64_t>(0ULL);for (;(cast<int64_t>((off + cast<int64_t>(2048ULL))) <= img->Size);){
int64_t n = cast<int64_t>(len(buf));
if ((cast<int64_t>((off + n)) > img->Size)) {
n = cast<int64_t>((img->Size - off));
}
auto tmp476 = os_File_ReadAt(img->f,sub(buf,0,n),off);
int64_t got = std::get<0>(tmp476);
Error e = std::get<1>(tmp476);
if ((((got == cast<int64_t>(0ULL)) && bool(e)) && (!go_errors_Is(e,go_io_EOF)))) {
return {cast<int64_t>(0ULL),{},e};
}
{int64_t p = cast<int64_t>(0ULL);for (;(cast<int64_t>((p + cast<int64_t>(2048ULL))) <= got);p += cast<int64_t>(2048ULL)){
if (((cast<std::string>(sub(buf,p,cast<int64_t>((p + cast<int64_t>(20ULL))))) == xbox_xdvdfsMagic) && (cast<std::string>(sub(buf,cast<int64_t>((cast<int64_t>((p + cast<int64_t>(2048ULL))) - cast<int64_t>(20ULL))),cast<int64_t>((p + cast<int64_t>(2048ULL))))) == xbox_xdvdfsMagic))) {
int64_t b = cast<int64_t>((cast<int64_t>((off + cast<int64_t>(p))) - cast<int64_t>(65536ULL)));
if ((b >= cast<int64_t>(0ULL))) {
return {b,append(Slice<uint8_t>{},sub(buf,p,cast<int64_t>((p + cast<int64_t>(2048ULL))))),{}};
}
}
}
}off += cast<int64_t>((cast<int64_t>(got) - (modi<int64_t>(cast<int64_t>(got),cast<int64_t>(2048ULL)))));
if ((got < cast<int64_t>(n))) {
break;
}
}
}return {cast<int64_t>(0ULL),{},go_errors_New(std::string("xiso: no XDVDFS volume descriptor found (is this an Xbox disc image?)",69))};
}
}
// tools/platform/xbox/xiso.go:216:1
Error xbox_Image_Close(xbox_Image* img){
{
return os_File_Close(img->f);
}
}
// tools/platform/xbox/xiso.go:219:1
std::tuple<Slice<uint8_t>,Error> xbox_Image_readAt(xbox_Image* img,int64_t off,int64_t n){
{
if ((((off < cast<int64_t>(0ULL)) || (n < cast<int64_t>(0ULL))) || (cast<int64_t>((off + cast<int64_t>(n))) > img->Size))) {
return {{},go_fmt_Errorf(std::string("xiso: read of %d bytes at %#x is outside the %d-byte image",58),n,off,img->Size)};
}
Slice<uint8_t> b = Slice<uint8_t>::make(n);
{
auto tmp477 = os_File_ReadAt(img->f,b,off);
Error err = std::get<1>(tmp477);
if (bool(err)) {
return {{},err};
}
}
return {b,{}};
}
}
// tools/platform/xbox/xiso.go:232:1
std::tuple<Slice<uint8_t>,Error> xbox_Image_Read(xbox_Image* img,int64_t off,int64_t n){
{
return xbox_Image_readAt(img,cast<int64_t>((img->Base + off)),n);
}
}
// tools/platform/xbox/xiso.go:237:1
std::tuple<Slice<xbox_Entry>,Error> xbox_Image_ReadDir(xbox_Image* img,std::string path){
{
auto tmp478 = xbox_Image_resolve(img,path);
xbox_Entry e = std::get<0>(tmp478);
Error err = std::get<1>(tmp478);
if (bool(err)) {
return {{},err};
}
if ((!e.IsDir)) {
return {{},go_fmt_Errorf(std::string("xiso: %q is not a directory",27),path)};
}
return xbox_Image_dirEntries(img,e.Sector,e.Size,e.Path);
}
}
// tools/platform/xbox/xiso.go:249:1
Error xbox_Image_Walk(xbox_Image* img,std::function<Error(xbox_Entry)> fn){
{
return xbox_Image_walk(img,img->rootSector,img->rootSize,std::string("",0),fn);
}
}
// tools/platform/xbox/xiso.go:253:1
Error xbox_Image_walk(xbox_Image* img,uint32_t sector,uint32_t size,std::string dirPath,std::function<Error(xbox_Entry)> fn){
{
auto tmp479 = xbox_Image_dirEntries(img,sector,size,dirPath);
Slice<xbox_Entry> entries = std::get<0>(tmp479);
Error err = std::get<1>(tmp479);
if (bool(err)) {
return err;
}
{auto&& tmp480 = entries;
for(int64_t tmp481=0;tmp481<len(tmp480);++tmp481){
auto e=tmp480[tmp481];{
Error err = fn(e);
if (bool(err)) {
return err;
}
}
if (e.IsDir) {
{
Error err = xbox_Image_walk(img,e.Sector,e.Size,e.Path,fn);
if (bool(err)) {
return err;
}
}
}
}}
return {};
}
}
// tools/platform/xbox/xiso.go:272:1
std::tuple<Slice<uint8_t>,Error> xbox_Image_ReadFile(xbox_Image* img,std::string path){
{
auto tmp482 = xbox_Image_resolve(img,path);
xbox_Entry e = std::get<0>(tmp482);
Error err = std::get<1>(tmp482);
if (bool(err)) {
return {{},err};
}
if (e.IsDir) {
return {{},go_fmt_Errorf(std::string("xiso: %q is a directory",23),path)};
}
return xbox_Image_ReadFileEntry(img,e);
}
}
// tools/platform/xbox/xiso.go:284:1
std::tuple<Slice<uint8_t>,Error> xbox_Image_ReadFileEntry(xbox_Image* img,xbox_Entry e){
{
if (e.IsDir) {
return {{},go_fmt_Errorf(std::string("xiso: %q is a directory",23),e.Path)};
}
return xbox_Image_Read(img,cast<int64_t>((cast<int64_t>(e.Sector) * cast<int64_t>(2048ULL))),cast<int64_t>(e.Size));
}
}
// tools/platform/xbox/xiso.go:297:1
std::tuple<Slice<xbox_Entry>,Error> xbox_Image_dirEntries(xbox_Image* img,uint32_t sector,uint32_t size,std::string dirPath){
{
if ((size == cast<uint32_t>(0ULL))) {
return {{},{}};
}
int64_t nsec = divi<int64_t>((cast<int64_t>((cast<int64_t>((cast<int64_t>(size) + cast<int64_t>(2048ULL))) - cast<int64_t>(1ULL)))),cast<int64_t>(2048ULL));
auto tmp483 = xbox_Image_Read(img,cast<int64_t>((cast<int64_t>(sector) * cast<int64_t>(2048ULL))),cast<int64_t>((nsec * cast<int64_t>(2048ULL))));
Slice<uint8_t> data = std::get<0>(tmp483);
Error err = std::get<1>(tmp483);
if (bool(err)) {
return {{},err};
}
Slice<xbox_Entry> out={};
Map<uint32_t,bool> seen = Map<uint32_t,bool>{};
Slice<uint32_t> stack = Slice<uint32_t>{cast<uint32_t>(0ULL)};
{;for (;(len(stack) > cast<int64_t>(0ULL));){
uint32_t p = stack[cast<int64_t>((len(stack) - cast<int64_t>(1ULL)))];
stack = sub(stack,0,cast<int64_t>((len(stack) - cast<int64_t>(1ULL))));
if (get(seen,p)) {
continue;
}
seen[p] = true;
if ((cast<int64_t>((cast<int64_t>(p) + cast<int64_t>(14ULL))) > len(data))) {
return {{},go_fmt_Errorf(std::string("xiso: directory %q has an entry at %#x past its %d-byte extent",62),dirPath,p,len(data))};
}
Slice<uint8_t> rec = sub(data,p,len(data));
uint16_t left = xbox_le16(rrBorrow(rec,cast<int64_t>(0ULL),len(rec)));
uint16_t right = xbox_le16(rrBorrow(rec,cast<int64_t>(2ULL),len(rec)));
uint32_t startSector = xbox_le32(rrBorrow(rec,cast<int64_t>(4ULL),len(rec)));
uint32_t fsize = xbox_le32(rrBorrow(rec,cast<int64_t>(8ULL),len(rec)));
uint8_t attr = rec[cast<int64_t>(12ULL)];
int64_t nlen = cast<int64_t>(rec[cast<int64_t>(13ULL)]);
if ((cast<int64_t>((cast<int64_t>((cast<int64_t>(p) + cast<int64_t>(14ULL))) + nlen)) > len(data))) {
return {{},go_fmt_Errorf(std::string("xiso: directory %q has a name overrunning its extent at %#x",59),dirPath,p)};
}
std::string name = cast<std::string>(sub(rec,cast<int64_t>(14ULL),cast<int64_t>((cast<int64_t>(14ULL) + nlen))));
xbox_Entry e = xbox_Entry{name,{},(cast<uint8_t>((attr & cast<uint8_t>(16ULL))) != cast<uint8_t>(0ULL)),fsize,startSector,attr};
if ((dirPath == std::string("",0))) {
e.Path = (std::string("/",1) + name);
}
else {
e.Path = ((dirPath + std::string("/",1)) + name);
}
out = append(out,Slice<xbox_Entry>{e});
if (((left != cast<uint16_t>(0ULL)) && (left != cast<uint16_t>(65535ULL)))) {
stack = append(stack,Slice<uint32_t>{cast<uint32_t>((cast<uint32_t>(left) * cast<uint32_t>(4ULL)))});
}
if (((right != cast<uint16_t>(0ULL)) && (right != cast<uint16_t>(65535ULL)))) {
stack = append(stack,Slice<uint32_t>{cast<uint32_t>((cast<uint32_t>(right) * cast<uint32_t>(4ULL)))});
}
}
}return {out,{}};
}
}
// tools/platform/xbox/xiso.go:359:1
std::tuple<xbox_Entry,Error> xbox_Image_resolve(xbox_Image* img,std::string path){
{
xbox_Entry cur = xbox_Entry{std::string("/",1),std::string("",0),true,img->rootSize,img->rootSector,{}};
{auto&& tmp484 = xbox_splitPath(path);
for(int64_t tmp485=0;tmp485<len(tmp484);++tmp485){
auto want=tmp484[tmp485];if ((!cur.IsDir)) {
return {xbox_Entry{},go_fmt_Errorf(std::string("xiso: %q is not a directory",27),cur.Path)};
}
auto tmp486 = xbox_Image_dirEntries(img,cur.Sector,cur.Size,cur.Path);
Slice<xbox_Entry> entries = std::get<0>(tmp486);
Error err = std::get<1>(tmp486);
if (bool(err)) {
return {xbox_Entry{},err};
}
bool found = false;
{auto&& tmp487 = entries;
for(int64_t tmp488=0;tmp488<len(tmp487);++tmp488){
auto e=tmp487[tmp488];if (go_strings_EqualFold(e.Name,want)) {
auto tmp489 = std::make_tuple(e,true);
cur = std::get<0>(tmp489);
found = std::get<1>(tmp489);
break;
}
}}
if ((!found)) {
return {xbox_Entry{},go_fmt_Errorf(std::string("xiso: %q not found",18),path)};
}
}}
return {cur,{}};
}
}
// tools/platform/xbox/xiso.go:383:1
Slice<std::string> xbox_splitPath(std::string p){
{
Slice<std::string> parts={};
{auto&& tmp490 = go_strings_Split(go_strings_Trim(p,std::string("/",1)),std::string("/",1));
for(int64_t tmp491=0;tmp491<len(tmp490);++tmp491){
auto c=tmp490[tmp491];if ((c != std::string("",0))) {
parts = append(parts,Slice<std::string>{c});
}
}}
return parts;
}
}

#include "fast.h"
