#include "runtime.h"
struct allegrex_Inst;
struct allegrex_loadSlot;
struct allegrex_CPU;
struct allegrex_CPUState;
struct psp_PixelEvent;
struct psp_MemRegion;
struct pend;
struct psp_Segment;
struct psp_Import;
struct psp_reloc;
struct psp_Module;
struct psp_GeList;
struct psp_geState;
struct psp_vert;
struct psp_vertexLayout;
struct psp_geWire;
struct psp_Image;
struct psp_subIntr;
struct psp_ioFile;
struct psp_Volume;
struct psp_Entry;
struct psp_syscall;
struct psp_kobject;
struct psp_Machine;
struct psp_PadEvent;
struct psp_mpegState;
struct psp_atracState;
struct psp_ProfileBucket;
struct psp_ProfileCounter;
struct psp_FrameProfile;
struct psp_geCounters;
struct psp_profState;
struct psp_tagInfo;
struct psp_Result;
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
struct Anon4;
struct Anon5;
struct Anon6;
struct Anon7;
struct Anon8;
struct Anon9;
using allegrex_Flow=int64_t;
struct allegrex_Inst{
uint32_t Addr{};
int64_t Len{};
std::string Mnem{};
std::string Text{};
allegrex_Flow Flow{};
uint32_t Target{};
bool HasTarget{};
bool HasDelay{};
};
struct allegrex_loadSlot{
uint32_t reg{};
uint32_t val{};
};
struct allegrex_CPU{
std::array<uint32_t,32> R{};
std::array<uint32_t,32> out{};
uint32_t HI{};
uint32_t LO{};
uint32_t PC{};
uint32_t nextPC{};
std::array<uint32_t,32> COP0{};
std::array<uint32_t,32> F{};
bool FCC{};
std::array<uint32_t,128> V{};
std::array<uint32_t,16> VfpuCtrl{};
std::function<void(uint32_t,uint32_t)> OnVFPU{};
std::function<bool(allegrex_CPU*,uint32_t)> Syscall{};
bool Halted{};
std::string HaltReason{};
uint64_t Steps{};
psp_Machine* bus{};
uint32_t curPC{};
allegrex_loadSlot ld{};
bool delaySlot{};
bool pendingDelay{};
uint32_t branchAddr{};
bool nullifyNext{};
};
struct allegrex_CPUState{
std::array<uint32_t,32> R{};
std::array<uint32_t,32> Out{};
uint32_t HI{};
uint32_t LO{};
uint32_t PC{};
uint32_t NextPC{};
std::array<uint32_t,32> COP0{};
std::array<uint32_t,32> F{};
bool FCC{};
std::array<uint32_t,128> V{};
std::array<uint32_t,16> VfpuCtrl{};
bool Halted{};
std::string HaltReason{};
uint64_t Steps{};
uint32_t CurPC{};
uint32_t LdReg{};
uint32_t LdVal{};
bool DelaySlot{};
bool PendingDelay{};
uint32_t BranchAddr{};
bool NullifyNext{};
};
struct psp_PixelEvent{
bool Drawn{};
bool ScissorReject{};
bool AlphaReject{};
bool StencilReject{};
bool ZReject{};
bool MaskReject{};
uint8_t R{};
uint8_t G{};
uint8_t B{};
uint8_t A{};
};
struct psp_MemRegion{
std::string Name{};
uint32_t Base{};
uint32_t Size{};
};
struct pend{
int64_t si{};
uint32_t off{};
};
struct psp_Segment{
uint32_t VAddr{};
uint32_t FileOff{};
Slice<uint8_t> Data{};
uint32_t MemSize{};
};
struct psp_Import{
std::string Library{};
Slice<uint32_t> NIDs{};
uint32_t StubAddr{};
};
struct psp_reloc{
uint32_t offset{};
uint32_t info{};
};
struct psp_Module{
std::string Name{};
uint16_t Type{};
uint32_t EntryPC{};
uint32_t GP{};
Slice<psp_Segment> Segments{};
Slice<psp_Import> Imports{};
Slice<psp_reloc> relocs{};
bool Encrypted{};
uint32_t Tag{};
};
struct psp_GeList{
uint32_t Start{};
Slice<uint32_t> Words{};
};
struct psp_geState{
uint32_t fbLow{};
uint32_t fbHigh{};
uint32_t fbStride{};
uint32_t fbFmt{};
uint32_t base{};
uint32_t vaddr{};
uint32_t iaddr{};
uint32_t offAddr{};
uint32_t vtype{};
uint32_t patchDivU{};
uint32_t patchDivV{};
bool lightOn{};
uint32_t matDiffuse{};
uint32_t matEmissive{};
uint32_t ambientCol{};
std::array<bool,4> lightEnable{};
std::array<uint32_t,4> lightType{};
std::array<std::array<float,3>,4> lightPos{};
std::array<uint32_t,4> lightDiff{};
float vpXS{};
float vpYS{};
float vpZS{};
float vpXC{};
float vpYC{};
float vpZC{};
float offX{};
float offY{};
std::array<float,16> world{};
std::array<float,16> view{};
std::array<float,16> proj{};
int64_t worldIdx{};
int64_t viewIdx{};
int64_t projIdx{};
uint32_t matColor{};
bool clearOn{};
bool texEnable{};
std::array<uint32_t,8> texAddrN{};
std::array<uint32_t,8> texStrideN{};
std::array<uint32_t,8> texWN{};
std::array<uint32_t,8> texHN{};
uint32_t texMaxLvl{};
uint32_t texLodMode{};
float texLodBias{};
uint32_t texAddr{};
uint32_t texStride{};
uint32_t texW{};
uint32_t texH{};
uint32_t texFmt{};
bool texSwizzle{};
float texScaleU{};
float texScaleV{};
float texOffU{};
float texOffV{};
uint32_t clutAddr{};
uint32_t clutFmt{};
std::array<uint32_t,256> clut{};
bool blendOn{};
uint32_t blendSrc{};
uint32_t blendDst{};
uint32_t blendEq{};
uint32_t blendFixA{};
uint32_t blendFixB{};
bool alphaTestOn{};
uint32_t alphaFunc{};
uint32_t alphaRef{};
uint32_t alphaTestMask{};
uint32_t texMapMode{};
uint32_t texProjSrc{};
uint32_t texWrapU{};
uint32_t texWrapV{};
bool texLinear{};
uint32_t texFunc{};
bool texUseA{};
bool texDouble{};
uint32_t texEnvCol{};
std::array<float,16> texMtx{};
int64_t texMtxIdx{};
int64_t scX0{};
int64_t scY0{};
int64_t scX1{};
int64_t scY1{};
uint32_t maskRGB{};
uint32_t maskA{};
uint32_t zLow{};
uint32_t zHigh{};
uint32_t zStride{};
bool zTestOn{};
uint32_t zFunc{};
bool zNoWrite{};
bool clearDepth{};
bool cullOn{};
uint32_t cullFace{};
bool stencilOn{};
uint32_t stFunc{};
uint32_t stRef{};
uint32_t stMask{};
uint32_t stSFail{};
uint32_t stZFail{};
uint32_t stZPass{};
bool fogOn{};
float fogEnd{};
float fogScale{};
uint32_t fogColor{};
uint32_t trSrc{};
uint32_t trDst{};
uint32_t trSrcStride{};
uint32_t trDstStride{};
uint32_t trSrcX{};
uint32_t trSrcY{};
uint32_t trDstX{};
uint32_t trDstY{};
uint32_t trW{};
uint32_t trH{};
};
struct psp_vert{
float x{};
float y{};
float z{};
float u{};
float v{};
uint8_t r{};
uint8_t g{};
uint8_t b{};
uint8_t a{};
float cx{};
float cy{};
float cz{};
float cw{};
float invW{};
float fog{};
bool clip{};
};
struct psp_vertexLayout{
uint32_t tfmt{};
uint32_t cfmt{};
uint32_t nfmt{};
uint32_t pfmt{};
uint32_t idxFmt{};
uint32_t offTex{};
uint32_t offCol{};
uint32_t offNrm{};
uint32_t offPos{};
uint32_t stride{};
bool through{};
};
struct psp_geWire{
uint32_t FbLow{};
uint32_t FbHigh{};
uint32_t FbStride{};
uint32_t FbFmt{};
uint32_t Base{};
uint32_t Vaddr{};
uint32_t Iaddr{};
uint32_t OffAddr{};
uint32_t Vtype{};
uint32_t PatchDivU{};
uint32_t PatchDivV{};
bool LightOn{};
uint32_t MatDiffuse{};
uint32_t MatEmissive{};
uint32_t AmbientCol{};
std::array<bool,4> LightEnable{};
std::array<uint32_t,4> LightType{};
std::array<std::array<float,3>,4> LightPos{};
std::array<uint32_t,4> LightDiff{};
float VpXS{};
float VpYS{};
float VpZS{};
float VpXC{};
float VpYC{};
float VpZC{};
float OffX{};
float OffY{};
std::array<float,16> World{};
std::array<float,16> View{};
std::array<float,16> Proj{};
int64_t WorldIdx{};
int64_t ViewIdx{};
int64_t ProjIdx{};
uint32_t MatColor{};
bool ClearOn{};
bool TexEnable{};
std::array<uint32_t,8> TexAddrN{};
std::array<uint32_t,8> TexStrideN{};
std::array<uint32_t,8> TexWN{};
std::array<uint32_t,8> TexHN{};
uint32_t TexMaxLvl{};
uint32_t TexLodMode{};
float TexLodBias{};
uint32_t TexAddr{};
uint32_t TexStride{};
uint32_t TexW{};
uint32_t TexH{};
uint32_t TexFmt{};
bool TexSwizzle{};
float TexScaleU{};
float TexScaleV{};
float TexOffU{};
float TexOffV{};
uint32_t ClutAddr{};
uint32_t ClutFmt{};
std::array<uint32_t,256> Clut{};
bool BlendOn{};
uint32_t BlendSrc{};
uint32_t BlendDst{};
uint32_t BlendEq{};
uint32_t BlendFixA{};
uint32_t BlendFixB{};
bool AlphaTestOn{};
uint32_t AlphaFunc{};
uint32_t AlphaRef{};
uint32_t AlphaTestMask{};
uint32_t TexMapMode{};
uint32_t TexProjSrc{};
uint32_t TexWrapU{};
uint32_t TexWrapV{};
bool TexLinear{};
uint32_t TexFunc{};
bool TexUseA{};
bool TexDouble{};
uint32_t TexEnvCol{};
std::array<float,16> TexMtx{};
int64_t TexMtxIdx{};
int64_t ScX0{};
int64_t ScY0{};
int64_t ScX1{};
int64_t ScY1{};
uint32_t MaskRGB{};
uint32_t MaskA{};
uint32_t ZLow{};
uint32_t ZHigh{};
uint32_t ZStride{};
bool ZTestOn{};
uint32_t ZFunc{};
bool ZNoWrite{};
bool ClearDepth{};
bool CullOn{};
uint32_t CullFace{};
bool StencilOn{};
uint32_t StFunc{};
uint32_t StRef{};
uint32_t StMask{};
uint32_t StSFail{};
uint32_t StZFail{};
uint32_t StZPass{};
bool FogOn{};
float FogEnd{};
float FogScale{};
uint32_t FogColor{};
uint32_t TrSrc{};
uint32_t TrDst{};
uint32_t TrSrcStride{};
uint32_t TrDstStride{};
uint32_t TrSrcX{};
uint32_t TrSrcY{};
uint32_t TrDstX{};
uint32_t TrDstY{};
uint32_t TrW{};
uint32_t TrH{};
};
struct psp_Image{
psp_Volume* Volume{};
psp_Machine* closer{};
};
struct psp_subIntr{
uint32_t intno{};
uint32_t subno{};
uint32_t handler{};
uint32_t arg{};
bool enabled{};
};
struct psp_Entry{
std::string Name{};
std::string Path{};
bool IsDir{};
int64_t Size{};
int64_t Block{};
};
struct psp_ioFile{
std::string path{};
psp_Entry ent{};
int64_t pos{};
int64_t async{};
bool hasAsync{};
Slice<psp_Entry> dir{};
int64_t dirPos{};
};
using psp_byteBlocks=Slice<uint8_t>;
struct psp_Volume{
BlockSource* src{};
std::string System{};
std::string Name{};
int64_t rootLBA{};
int64_t rootSize{};
};
struct psp_syscall{
std::string name{};
std::function<void(psp_Machine*)> handler{};
};
using psp_threadState=int64_t;
struct psp_kobject{
std::string kind{};
std::string name{};
uint32_t entry{};
uint32_t addr{};
uint32_t size{};
uint32_t used{};
uint32_t bits{};
int32_t count{};
uint32_t priority{};
uint32_t stackTop{};
psp_threadState tstate{};
allegrex_CPUState ctx{};
uint32_t waitEv{};
uint32_t waitBits{};
uint32_t waitMode{};
uint32_t waitOutPtr{};
uint32_t waitSema{};
int32_t waitNeed{};
uint32_t wakeVblank{};
};
struct psp_PadEvent{
uint32_t AtVblank{};
uint32_t Buttons{};
};
struct psp_mpegState{
uint32_t Handle{};
uint32_t Ringbuf{};
uint32_t Packets{};
uint32_t In{};
uint32_t Cb{};
uint32_t CbArg{};
uint32_t Data{};
uint32_t Pts{};
uint32_t FedTotal{};
};
struct psp_geCounters{
int64_t lists{};
int64_t prims{};
int64_t verts{};
int64_t frags{};
int64_t zKilled{};
int64_t aKilled{};
int64_t sKilled{};
int64_t scissored{};
int64_t xfers{};
};
struct psp_ProfileBucket{
std::string Name{};
double Millis{};
int64_t Count{};
};
struct psp_ProfileCounter{
std::string Name{};
int64_t Value{};
};
struct psp_FrameProfile{
double TotalMs{};
Slice<psp_ProfileBucket> Buckets{};
Slice<psp_ProfileCounter> Counters{};
};
struct psp_profState{
std::array<int64_t,5> ns{};
std::array<int64_t,5> count{};
time_Time runStart{};
bool inRun{};
int64_t frameNs{};
psp_geCounters base{};
uint64_t baseInstr{};
int64_t gen{};
psp_FrameProfile last{};
bool has{};
};
struct psp_Machine{
Slice<uint8_t> ram{};
Slice<uint8_t> vram{};
Slice<uint8_t> scratch{};
allegrex_CPU* CPU{};
Map<uint32_t,uint32_t> io{};
Map<uint32_t,psp_syscall*> syscalls{};
uint32_t nextSyscall{};
Map<uint32_t,psp_kobject*> handles{};
uint32_t nextHandle{};
uint32_t heapPtr{};
uint32_t heapEnd{};
uint32_t threadEntry{};
psp_kobject* current{};
std::string doneReason{};
Map<uint32_t,psp_subIntr*> subIntrs{};
uint32_t vblanks{};
uint32_t pad{};
uint32_t padPrev{};
Slice<psp_PadEvent> padScript{};
uint32_t savedataStatus{};
bool volatileLocked{};
psp_mpegState mpeg{};
Map<uint32_t,psp_atracState*> atrac{};
uint32_t nextAtrac{};
psp_Volume* vol{};
Map<uint32_t,psp_ioFile*> files{};
uint32_t nextFd{};
uint32_t audioCh{};
Map<std::string,int64_t> SyscallCalls{};
Slice<uint8_t> tty{};
uint32_t fbAddr{};
uint32_t fbWidth{};
uint32_t fbFormat{};
Slice<psp_GeList> GeLists{};
std::function<void(psp_GeList)> OnGeList{};
psp_geState* geSt{};
int64_t geLimit{};
int64_t geCount{};
std::string imageHash{};
uint32_t WatchLo{};
uint32_t WatchHi{};
std::function<void(uint32_t,uint32_t,uint32_t)> OnWrite{};
uint32_t RWatchLo{};
uint32_t RWatchHi{};
std::function<void(uint32_t,uint32_t,uint32_t)> OnRead{};
std::function<void(psp_Machine*,uint32_t)> OnStep{};
std::function<void(uint32_t)> OnGeCmd{};
std::function<void(uint32_t,uint32_t,psp_PixelEvent)> OnPixel{};
std::function<void(psp_Machine*)> OnDisplay{};
bool StopRequested{};
bool Profile{};
psp_profState prof{};
psp_geCounters geCnt{};
Slice<std::string> Log{};
Map<std::string,bool> logSeen{};
bool Halted{};
std::string HaltReason{};
};
struct psp_atracState{
uint32_t Buf{};
uint32_t Size{};
uint32_t Frames{};
uint32_t Pos{};
uint32_t Channels{};
};
struct psp_tagInfo{
Slice<uint8_t> seed{};
int64_t code{};
};
struct psp_Result{
uint64_t Steps{};
uint32_t PC{};
std::string Reason{};
};
struct Anon0{uint32_t Addr{};int64_t Len{};std::string Mnem{};std::string Text{};allegrex_Flow Flow{};uint32_t Target{};bool HasTarget{};bool HasDelay{};};
struct Anon1{uint32_t reg{};uint32_t val{};};
struct Anon10{int64_t si{};uint32_t off{};};
struct Anon11{uint32_t Start{};Slice<uint32_t> Words{};};
struct Anon12{uint32_t fbLow{};uint32_t fbHigh{};uint32_t fbStride{};uint32_t fbFmt{};uint32_t base{};uint32_t vaddr{};uint32_t iaddr{};uint32_t offAddr{};uint32_t vtype{};uint32_t patchDivU{};uint32_t patchDivV{};bool lightOn{};uint32_t matDiffuse{};uint32_t matEmissive{};uint32_t ambientCol{};std::array<bool,4> lightEnable{};std::array<uint32_t,4> lightType{};std::array<std::array<float,3>,4> lightPos{};std::array<uint32_t,4> lightDiff{};float vpXS{};float vpYS{};float vpZS{};float vpXC{};float vpYC{};float vpZC{};float offX{};float offY{};std::array<float,16> world{};std::array<float,16> view{};std::array<float,16> proj{};int64_t worldIdx{};int64_t viewIdx{};int64_t projIdx{};uint32_t matColor{};bool clearOn{};bool texEnable{};std::array<uint32_t,8> texAddrN{};std::array<uint32_t,8> texStrideN{};std::array<uint32_t,8> texWN{};std::array<uint32_t,8> texHN{};uint32_t texMaxLvl{};uint32_t texLodMode{};float texLodBias{};uint32_t texAddr{};uint32_t texStride{};uint32_t texW{};uint32_t texH{};uint32_t texFmt{};bool texSwizzle{};float texScaleU{};float texScaleV{};float texOffU{};float texOffV{};uint32_t clutAddr{};uint32_t clutFmt{};std::array<uint32_t,256> clut{};bool blendOn{};uint32_t blendSrc{};uint32_t blendDst{};uint32_t blendEq{};uint32_t blendFixA{};uint32_t blendFixB{};bool alphaTestOn{};uint32_t alphaFunc{};uint32_t alphaRef{};uint32_t alphaTestMask{};uint32_t texMapMode{};uint32_t texProjSrc{};uint32_t texWrapU{};uint32_t texWrapV{};bool texLinear{};uint32_t texFunc{};bool texUseA{};bool texDouble{};uint32_t texEnvCol{};std::array<float,16> texMtx{};int64_t texMtxIdx{};int64_t scX0{};int64_t scY0{};int64_t scX1{};int64_t scY1{};uint32_t maskRGB{};uint32_t maskA{};uint32_t zLow{};uint32_t zHigh{};uint32_t zStride{};bool zTestOn{};uint32_t zFunc{};bool zNoWrite{};bool clearDepth{};bool cullOn{};uint32_t cullFace{};bool stencilOn{};uint32_t stFunc{};uint32_t stRef{};uint32_t stMask{};uint32_t stSFail{};uint32_t stZFail{};uint32_t stZPass{};bool fogOn{};float fogEnd{};float fogScale{};uint32_t fogColor{};uint32_t trSrc{};uint32_t trDst{};uint32_t trSrcStride{};uint32_t trDstStride{};uint32_t trSrcX{};uint32_t trSrcY{};uint32_t trDstX{};uint32_t trDstY{};uint32_t trW{};uint32_t trH{};};
struct Anon13{std::string name{};std::array<float,16> v{};};
struct Anon14{float x{};float y{};float z{};float u{};float v{};uint8_t r{};uint8_t g{};uint8_t b{};uint8_t a{};float cx{};float cy{};float cz{};float cw{};float invW{};float fog{};bool clip{};};
struct Anon15{uint32_t tfmt{};uint32_t cfmt{};uint32_t nfmt{};uint32_t pfmt{};uint32_t idxFmt{};uint32_t offTex{};uint32_t offCol{};uint32_t offNrm{};uint32_t offPos{};uint32_t stride{};bool through{};};
struct Anon16{uint32_t FbLow{};uint32_t FbHigh{};uint32_t FbStride{};uint32_t FbFmt{};uint32_t Base{};uint32_t Vaddr{};uint32_t Iaddr{};uint32_t OffAddr{};uint32_t Vtype{};uint32_t PatchDivU{};uint32_t PatchDivV{};bool LightOn{};uint32_t MatDiffuse{};uint32_t MatEmissive{};uint32_t AmbientCol{};std::array<bool,4> LightEnable{};std::array<uint32_t,4> LightType{};std::array<std::array<float,3>,4> LightPos{};std::array<uint32_t,4> LightDiff{};float VpXS{};float VpYS{};float VpZS{};float VpXC{};float VpYC{};float VpZC{};float OffX{};float OffY{};std::array<float,16> World{};std::array<float,16> View{};std::array<float,16> Proj{};int64_t WorldIdx{};int64_t ViewIdx{};int64_t ProjIdx{};uint32_t MatColor{};bool ClearOn{};bool TexEnable{};std::array<uint32_t,8> TexAddrN{};std::array<uint32_t,8> TexStrideN{};std::array<uint32_t,8> TexWN{};std::array<uint32_t,8> TexHN{};uint32_t TexMaxLvl{};uint32_t TexLodMode{};float TexLodBias{};uint32_t TexAddr{};uint32_t TexStride{};uint32_t TexW{};uint32_t TexH{};uint32_t TexFmt{};bool TexSwizzle{};float TexScaleU{};float TexScaleV{};float TexOffU{};float TexOffV{};uint32_t ClutAddr{};uint32_t ClutFmt{};std::array<uint32_t,256> Clut{};bool BlendOn{};uint32_t BlendSrc{};uint32_t BlendDst{};uint32_t BlendEq{};uint32_t BlendFixA{};uint32_t BlendFixB{};bool AlphaTestOn{};uint32_t AlphaFunc{};uint32_t AlphaRef{};uint32_t AlphaTestMask{};uint32_t TexMapMode{};uint32_t TexProjSrc{};uint32_t TexWrapU{};uint32_t TexWrapV{};bool TexLinear{};uint32_t TexFunc{};bool TexUseA{};bool TexDouble{};uint32_t TexEnvCol{};std::array<float,16> TexMtx{};int64_t TexMtxIdx{};int64_t ScX0{};int64_t ScY0{};int64_t ScX1{};int64_t ScY1{};uint32_t MaskRGB{};uint32_t MaskA{};uint32_t ZLow{};uint32_t ZHigh{};uint32_t ZStride{};bool ZTestOn{};uint32_t ZFunc{};bool ZNoWrite{};bool ClearDepth{};bool CullOn{};uint32_t CullFace{};bool StencilOn{};uint32_t StFunc{};uint32_t StRef{};uint32_t StMask{};uint32_t StSFail{};uint32_t StZFail{};uint32_t StZPass{};bool FogOn{};float FogEnd{};float FogScale{};uint32_t FogColor{};uint32_t TrSrc{};uint32_t TrDst{};uint32_t TrSrcStride{};uint32_t TrDstStride{};uint32_t TrSrcX{};uint32_t TrSrcY{};uint32_t TrDstX{};uint32_t TrDstY{};uint32_t TrW{};uint32_t TrH{};};
struct Anon17{psp_Volume* Volume{};psp_Machine* closer{};};
struct Anon18{uint32_t intno{};uint32_t subno{};uint32_t handler{};uint32_t arg{};bool enabled{};};
struct Anon19{std::string path{};psp_Entry ent{};int64_t pos{};int64_t async{};bool hasAsync{};Slice<psp_Entry> dir{};int64_t dirPos{};};
struct Anon2{std::array<uint32_t,32> R{};std::array<uint32_t,32> out{};uint32_t HI{};uint32_t LO{};uint32_t PC{};uint32_t nextPC{};std::array<uint32_t,32> COP0{};std::array<uint32_t,32> F{};bool FCC{};std::array<uint32_t,128> V{};std::array<uint32_t,16> VfpuCtrl{};std::function<void(uint32_t,uint32_t)> OnVFPU{};std::function<bool(allegrex_CPU*,uint32_t)> Syscall{};bool Halted{};std::string HaltReason{};uint64_t Steps{};psp_Machine* bus{};uint32_t curPC{};allegrex_loadSlot ld{};bool delaySlot{};bool pendingDelay{};uint32_t branchAddr{};bool nullifyNext{};};
struct Anon20{BlockSource* src{};std::string System{};std::string Name{};int64_t rootLBA{};int64_t rootSize{};};
struct Anon21{std::string Name{};std::string Path{};bool IsDir{};int64_t Size{};int64_t Block{};};
struct Anon22{std::string name{};std::function<void(psp_Machine*)> handler{};};
struct Anon23{std::string kind{};std::string name{};uint32_t entry{};uint32_t addr{};uint32_t size{};uint32_t used{};uint32_t bits{};int32_t count{};uint32_t priority{};uint32_t stackTop{};psp_threadState tstate{};allegrex_CPUState ctx{};uint32_t waitEv{};uint32_t waitBits{};uint32_t waitMode{};uint32_t waitOutPtr{};uint32_t waitSema{};int32_t waitNeed{};uint32_t wakeVblank{};};
struct Anon24{Slice<uint8_t> ram{};Slice<uint8_t> vram{};Slice<uint8_t> scratch{};allegrex_CPU* CPU{};Map<uint32_t,uint32_t> io{};Map<uint32_t,psp_syscall*> syscalls{};uint32_t nextSyscall{};Map<uint32_t,psp_kobject*> handles{};uint32_t nextHandle{};uint32_t heapPtr{};uint32_t heapEnd{};uint32_t threadEntry{};psp_kobject* current{};std::string doneReason{};Map<uint32_t,psp_subIntr*> subIntrs{};uint32_t vblanks{};uint32_t pad{};uint32_t padPrev{};Slice<psp_PadEvent> padScript{};uint32_t savedataStatus{};bool volatileLocked{};psp_mpegState mpeg{};Map<uint32_t,psp_atracState*> atrac{};uint32_t nextAtrac{};psp_Volume* vol{};Map<uint32_t,psp_ioFile*> files{};uint32_t nextFd{};uint32_t audioCh{};Map<std::string,int64_t> SyscallCalls{};Slice<uint8_t> tty{};uint32_t fbAddr{};uint32_t fbWidth{};uint32_t fbFormat{};Slice<psp_GeList> GeLists{};std::function<void(psp_GeList)> OnGeList{};psp_geState* geSt{};int64_t geLimit{};int64_t geCount{};std::string imageHash{};uint32_t WatchLo{};uint32_t WatchHi{};std::function<void(uint32_t,uint32_t,uint32_t)> OnWrite{};uint32_t RWatchLo{};uint32_t RWatchHi{};std::function<void(uint32_t,uint32_t,uint32_t)> OnRead{};std::function<void(psp_Machine*,uint32_t)> OnStep{};std::function<void(uint32_t)> OnGeCmd{};std::function<void(uint32_t,uint32_t,psp_PixelEvent)> OnPixel{};std::function<void(psp_Machine*)> OnDisplay{};bool StopRequested{};bool Profile{};psp_profState prof{};psp_geCounters geCnt{};Slice<std::string> Log{};Map<std::string,bool> logSeen{};bool Halted{};std::string HaltReason{};};
struct Anon25{uint32_t AtVblank{};uint32_t Buttons{};};
struct Anon26{uint32_t Handle{};uint32_t Ringbuf{};uint32_t Packets{};uint32_t In{};uint32_t Cb{};uint32_t CbArg{};uint32_t Data{};uint32_t Pts{};uint32_t FedTotal{};};
struct Anon27{uint32_t Buf{};uint32_t Size{};uint32_t Frames{};uint32_t Pos{};uint32_t Channels{};};
struct Anon28{std::string Name{};double Millis{};int64_t Count{};};
struct Anon29{std::string Name{};int64_t Value{};};
struct Anon3{std::array<uint32_t,32> R{};std::array<uint32_t,32> Out{};uint32_t HI{};uint32_t LO{};uint32_t PC{};uint32_t NextPC{};std::array<uint32_t,32> COP0{};std::array<uint32_t,32> F{};bool FCC{};std::array<uint32_t,128> V{};std::array<uint32_t,16> VfpuCtrl{};bool Halted{};std::string HaltReason{};uint64_t Steps{};uint32_t CurPC{};uint32_t LdReg{};uint32_t LdVal{};bool DelaySlot{};bool PendingDelay{};uint32_t BranchAddr{};bool NullifyNext{};};
struct Anon30{double TotalMs{};Slice<psp_ProfileBucket> Buckets{};Slice<psp_ProfileCounter> Counters{};};
struct Anon31{int64_t lists{};int64_t prims{};int64_t verts{};int64_t frags{};int64_t zKilled{};int64_t aKilled{};int64_t sKilled{};int64_t scissored{};int64_t xfers{};};
struct Anon32{std::array<int64_t,5> ns{};std::array<int64_t,5> count{};time_Time runStart{};bool inRun{};int64_t frameNs{};psp_geCounters base{};uint64_t baseInstr{};int64_t gen{};psp_FrameProfile last{};bool has{};};
struct Anon33{Slice<uint8_t> seed{};int64_t code{};};
struct Anon34{uint64_t Steps{};uint32_t PC{};std::string Reason{};};
struct Anon4{bool Drawn{};bool ScissorReject{};bool AlphaReject{};bool StencilReject{};bool ZReject{};bool MaskReject{};uint8_t R{};uint8_t G{};uint8_t B{};uint8_t A{};};
struct Anon5{std::string Name{};uint32_t Base{};uint32_t Size{};};
struct Anon6{uint32_t VAddr{};uint32_t FileOff{};Slice<uint8_t> Data{};uint32_t MemSize{};};
struct Anon7{std::string Library{};Slice<uint32_t> NIDs{};uint32_t StubAddr{};};
struct Anon8{uint32_t offset{};uint32_t info{};};
struct Anon9{std::string Name{};uint16_t Type{};uint32_t EntryPC{};uint32_t GP{};Slice<psp_Segment> Segments{};Slice<psp_Import> Imports{};Slice<psp_reloc> relocs{};bool Encrypted{};uint32_t Tag{};};

#include "adapters-decl.h"
std::string allegrex_Flow_String(allegrex_Flow f);
std::string allegrex_Inst_String(allegrex_Inst in);
void allegrex_CPU_Reset(allegrex_CPU* c);
void allegrex_CPU_SetPC(allegrex_CPU* c,uint32_t pc);
void allegrex_CPU_SetReg(allegrex_CPU* c,uint32_t i,uint32_t v);
uint32_t allegrex_CPU_Reg(allegrex_CPU* c,uint32_t i);
uint32_t allegrex_CPU_CurPC(allegrex_CPU* c);
void allegrex_CPU_set(allegrex_CPU* c,uint32_t i,uint32_t v);
uint32_t allegrex_CPU_read8(allegrex_CPU* c,uint32_t a);
void allegrex_CPU_write8(allegrex_CPU* c,uint32_t a,uint32_t v);
void allegrex_CPU_Exception(allegrex_CPU* c,uint32_t code);
void allegrex_CPU_rfe(allegrex_CPU* c);
bool allegrex_CPU_Interrupt(allegrex_CPU* c,bool pending);
void allegrex_CPU_addrError(allegrex_CPU* c,uint32_t code,uint32_t addr);
std::string allegrex_reg(uint32_t i);
allegrex_Inst allegrex_Decode(Slice<uint8_t> code,uint32_t addr);
allegrex_Inst allegrex_decodeSpecial2(allegrex_Inst in,uint32_t w,uint32_t rs,uint32_t rt,uint32_t rd,uint32_t funct);
allegrex_Inst allegrex_decodeSpecial3(allegrex_Inst in,uint32_t w,uint32_t rs,uint32_t rt,uint32_t rd,uint32_t shamt,uint32_t funct);
allegrex_Inst allegrex_decodeSpecial(allegrex_Inst in,uint32_t w,uint32_t rs,uint32_t rt,uint32_t rd,uint32_t shamt,uint32_t funct);
allegrex_Inst allegrex_decodeCop0(allegrex_Inst in,uint32_t w,uint32_t rs,uint32_t rt,uint32_t rd);
allegrex_Inst allegrex_word(allegrex_Inst in,uint32_t w);
int64_t allegrex_CPU_Step(allegrex_CPU* c);
uint32_t allegrex_CPU_reg(allegrex_CPU* c,uint32_t i);
void allegrex_CPU_load(allegrex_CPU* c,uint32_t reg,uint32_t val);
void allegrex_CPU_doBranch(allegrex_CPU* c,bool taken,uint32_t target);
void allegrex_CPU_execute(allegrex_CPU* c,uint32_t w);
void allegrex_CPU_branchLikely(allegrex_CPU* c,uint32_t op,uint32_t rs,uint32_t rt,uint32_t target);
void allegrex_CPU_special2(allegrex_CPU* c,uint32_t w,uint32_t rs,uint32_t rt,uint32_t rd);
void allegrex_CPU_special3(allegrex_CPU* c,uint32_t w,uint32_t rs,uint32_t rt,uint32_t rd,uint32_t shamt);
uint64_t allegrex_CPU_hilo(allegrex_CPU* c);
void allegrex_CPU_setHilo(allegrex_CPU* c,uint64_t v);
uint32_t allegrex_clz32(uint32_t v);
uint32_t allegrex_bitrev32(uint32_t v);
void allegrex_CPU_special(allegrex_CPU* c,uint32_t w,uint32_t rs,uint32_t rt,uint32_t rd,uint32_t shamt);
void allegrex_CPU_divSigned(allegrex_CPU* c,int32_t a,int32_t b);
void allegrex_CPU_divUnsigned(allegrex_CPU* c,uint32_t a,uint32_t b);
void allegrex_CPU_loadOp(allegrex_CPU* c,uint32_t op,uint32_t rs,uint32_t rt,uint32_t simm);
void allegrex_CPU_storeOp(allegrex_CPU* c,uint32_t op,uint32_t rs,uint32_t rt,uint32_t simm);
void allegrex_CPU_cop0(allegrex_CPU* c,uint32_t w,uint32_t rs,uint32_t rt,uint32_t rd);
uint32_t allegrex_rotr32(uint32_t v,uint32_t n);
std::tuple<uint32_t,bool> allegrex_addOv(uint32_t a,uint32_t b);
std::tuple<uint32_t,bool> allegrex_subOv(uint32_t a,uint32_t b);
uint32_t allegrex_b2u(bool b);
std::string allegrex_fpr(uint32_t i);
allegrex_Inst allegrex_decodeCop1(allegrex_Inst in,uint32_t w,uint32_t rs,uint32_t rt,uint32_t rd,uint32_t shamt);
allegrex_Inst allegrex_decodeCop1Fmt(allegrex_Inst in,uint32_t w,uint32_t rs,uint32_t ft,uint32_t fs,uint32_t fd);
float allegrex_CPU_ff(allegrex_CPU* c,uint32_t i);
void allegrex_CPU_setf(allegrex_CPU* c,uint32_t i,float v);
void allegrex_CPU_cop1(allegrex_CPU* c,uint32_t w,uint32_t rs,uint32_t rt,uint32_t rd,uint32_t shamt);
void allegrex_CPU_cop1S(allegrex_CPU* c,uint32_t w,uint32_t ft,uint32_t fs,uint32_t fd);
void allegrex_CPU_cop1W(allegrex_CPU* c,uint32_t w,uint32_t ft,uint32_t fs,uint32_t fd);
bool allegrex_fcompare(uint32_t funct,float a,float b);
allegrex_CPUState allegrex_CPU_SaveState(allegrex_CPU* c);
void allegrex_CPU_LoadState(allegrex_CPU* c,allegrex_CPUState s);
uint32_t allegrex_vfpuSingle(uint32_t reg);
std::array<uint32_t,4> allegrex_vecIdx(uint32_t reg,uint32_t n);
uint32_t allegrex_matIdx(uint32_t reg,uint32_t side,uint32_t a,uint32_t b);
std::array<float,4> allegrex_CPU_vread(allegrex_CPU* c,uint32_t reg,uint32_t n);
std::array<uint32_t,4> allegrex_CPU_vreadBits(allegrex_CPU* c,uint32_t reg,uint32_t n);
void allegrex_CPU_vwrite(allegrex_CPU* c,std::array<float,4> v,uint32_t reg,uint32_t n);
void allegrex_CPU_vwriteBits(allegrex_CPU* c,std::array<uint32_t,4> v,uint32_t reg,uint32_t n);
std::array<float,16> allegrex_CPU_mread(allegrex_CPU* c,uint32_t reg,uint32_t side);
void allegrex_CPU_mwrite(allegrex_CPU* c,std::array<float,16> m,uint32_t reg,uint32_t side);
void allegrex_applyPfx(std::array<float,4>* v,uint32_t data,uint32_t n);
void allegrex_CPU_applyPfxS(allegrex_CPU* c,std::array<float,4>* v,uint32_t n);
void allegrex_CPU_applyPfxT(allegrex_CPU* c,std::array<float,4>* v,uint32_t n);
void allegrex_CPU_applyPfxD(allegrex_CPU* c,std::array<float,4>* v,uint32_t n);
float allegrex_vfpuClamp(float v,float lo,float hi);
void allegrex_CPU_eatPfx(allegrex_CPU* c);
uint32_t allegrex_vfpuVecN(uint32_t w);
float allegrex_vfpuSin(float x);
float allegrex_vfpuCos(float x);
float allegrex_float16to32(uint16_t h);
float allegrex_sign2(uint32_t signBit);
std::string allegrex_vnot(uint32_t reg,uint32_t n);
std::string allegrex_mnot(uint32_t reg,uint32_t side);
allegrex_Inst allegrex_decodeCop2(allegrex_Inst in,uint32_t w,uint32_t rs,uint32_t rt,uint32_t rd);
allegrex_Inst allegrex_decodeVFPU(allegrex_Inst in,uint32_t w,uint32_t op,uint32_t rs,uint32_t rt,uint32_t simm);
std::string allegrex_lr(uint32_t w);
void allegrex_CPU_cop2(allegrex_CPU* c,uint32_t w,uint32_t rs,uint32_t rt,uint32_t rd);
void allegrex_CPU_vfpuOp(allegrex_CPU* c,uint32_t w,uint32_t op,uint32_t rs,uint32_t rt,uint32_t simm);
void allegrex_CPU_vfpuUnimpl(allegrex_CPU* c,uint32_t w,uint32_t op);
void allegrex_CPU_vfpuALU(allegrex_CPU* c,uint32_t w,uint32_t op);
bool allegrex_isInf32(float f);
float allegrex_vfpuMinMax(float s,float t,bool wantMax);
void allegrex_CPU_vfpu4(allegrex_CPU* c,uint32_t w,uint32_t rs,uint32_t rt);
void allegrex_CPU_vfpu5(allegrex_CPU* c,uint32_t w);
void allegrex_CPU_vfpuMatrix(allegrex_CPU* c,uint32_t w,uint32_t rt);
bool psp_PixelEvent_Rejected(psp_PixelEvent e);
void psp_Machine_pixelEvent(psp_Machine* m,int64_t x,int64_t y,psp_PixelEvent ev);
psp_Result psp_Machine_RunStopAfterGeCommand(psp_Machine* m,int64_t k,uint64_t budget);
int64_t psp_Machine_GeCommands(psp_Machine* m);
std::tuple<uint32_t,uint32_t,uint32_t,bool> psp_Machine_RenderTarget(psp_Machine* m);
std::tuple<uint32_t,uint32_t,uint32_t> psp_Machine_Scanout(psp_Machine* m);
std::tuple<uint32_t,uint32_t,bool> psp_Machine_DepthTarget(psp_Machine* m);
std::tuple<image_RGBA*,Error> psp_Machine_RenderSurface(psp_Machine* m,uint32_t addr,uint32_t stride,uint32_t format,int64_t w,int64_t h);
Slice<std::string> psp_TextureFormats();
std::tuple<uint32_t,bool> psp_TextureFormat(std::string name);
Slice<psp_MemRegion> psp_Machine_MemRegions(psp_Machine* m);
psp_Volume* psp_Machine_Volume(psp_Machine* m);
std::tuple<psp_Module*,Error> psp_ParseELF(Slice<uint8_t> b);
void psp_Module_parseRelocs(psp_Module* m,Slice<uint8_t> b);
void psp_Module_Relocate(psp_Module* m,uint32_t base);
Slice<uint8_t> psp_Module_segData(psp_Module* m,uint32_t va,uint32_t n);
std::tuple<uint32_t,bool> psp_Module_fileToVA(psp_Module* m,uint32_t fileOff);
void psp_Module_parseModuleInfo(psp_Module* m,Slice<uint8_t> b);
void psp_Module_parseStubs(psp_Module* m,uint32_t stubTop,uint32_t stubEnd);
std::string psp_cstrVA(psp_Module* m,uint32_t va);
std::tuple<psp_Module*,Error> psp_LoadModuleImage(Slice<uint8_t> raw);
std::string psp_firstMagic(Slice<uint8_t> b);
image_RGBA* psp_Machine_Framebuffer(psp_Machine* m);
color_RGBA psp_Machine_readPixel(psp_Machine* m,uint32_t base,uint32_t stride,uint32_t x,uint32_t y);
color_RGBA psp_Machine_readPixelFmt(psp_Machine* m,uint32_t base,uint32_t stride,uint32_t format,uint32_t x,uint32_t y);
color_RGBA psp_decode16(uint16_t p,uint32_t fmt);
std::string psp_Machine_FramebufferInfo(psp_Machine* m);
psp_GeList psp_Machine_captureList(psp_Machine* m,uint32_t addr);
uint32_t psp_geBaseAddr(uint32_t arg);
void psp_Machine_execGeList(psp_Machine* m,psp_GeList list);
std::string psp_GeCmdName(uint32_t cmd);
void psp_Machine_rasterTriClipped(psp_Machine* m,psp_geState* s,psp_vert a,psp_vert b,psp_vert c);
void psp_Machine_rasterTri(psp_Machine* m,psp_geState* s,psp_vert a,psp_vert b,psp_vert c);
void psp_Machine_rasterSprite(psp_Machine* m,psp_geState* s,psp_vert a,psp_vert b);
std::tuple<uint8_t,uint8_t,uint8_t> psp_applyFog(psp_geState* s,float f,uint8_t r,uint8_t g,uint8_t b);
std::tuple<float,float> psp_uvAt(psp_geState* s,psp_vert a,psp_vert b,psp_vert c,float area,float px,float py);
float psp_hypot32(float x,float y);
uint32_t psp_wrapTexel(int64_t i,uint32_t size,uint32_t clamp);
std::tuple<uint8_t,uint8_t,uint8_t,uint8_t> psp_modTex(psp_Machine* m,psp_geState* s,float u,float v,float rho,uint8_t r,uint8_t g,uint8_t b,uint8_t a);
std::tuple<uint8_t,uint8_t,uint8_t,uint8_t> psp_Machine_sampleTex(psp_Machine* m,psp_geState* s,uint32_t tx,uint32_t ty);
std::tuple<uint8_t,uint8_t,uint8_t,uint8_t> psp_Machine_sampleTexLvl(psp_Machine* m,psp_geState* s,uint32_t tx,uint32_t ty,uint32_t lvl);
std::tuple<uint8_t,uint8_t,uint8_t,uint8_t> psp_Machine_sampleTexLvl_Reference(psp_Machine* m,psp_geState* s,uint32_t tx,uint32_t ty,uint32_t lvl);
uint32_t psp_geState_texLevel(psp_geState* s,float rho);
uint32_t psp_Machine_texOff(psp_Machine* m,psp_geState* s,uint32_t xb,uint32_t y,uint32_t rowBytes);
std::tuple<uint8_t,uint8_t,uint8_t,uint8_t> psp_geState_clutLookup(psp_geState* s,uint32_t raw);
std::tuple<uint8_t,uint8_t,uint8_t,uint8_t> psp_decode16a(uint16_t p,uint32_t fmt);
bool psp_geState_stencilTest(psp_geState* s,uint8_t dstA);
uint8_t psp_stencilOp(uint32_t op,uint8_t cur,uint8_t ref);
bool psp_geState_alphaPass(psp_geState* s,uint8_t a);
uint32_t psp_blendFactor(uint32_t f,uint8_t sc,uint8_t sa,uint8_t da,uint32_t fix,int64_t ch);
uint32_t psp_min32(uint32_t a,uint32_t b);
uint8_t psp_clamp255(int32_t v);
void psp_Machine_putPixel(psp_Machine* m,psp_geState* s,int64_t x,int64_t y,float z,uint8_t r,uint8_t g,uint8_t b,uint8_t a);
void psp_Machine_storePixel(psp_Machine* m,psp_geState* s,uint32_t base,uint32_t off,uint8_t r,uint8_t g,uint8_t b,uint8_t a);
uint8_t psp_Machine_dstAlpha(psp_Machine* m,psp_geState* s,uint32_t base,uint32_t off);
void psp_Machine_storeAlpha(psp_Machine* m,psp_geState* s,uint32_t base,uint32_t off,uint8_t a);
std::tuple<uint8_t,uint8_t,uint8_t> psp_Machine_dstPixel(psp_Machine* m,psp_geState* s,uint32_t base,uint32_t off);
std::tuple<uint8_t,uint8_t,uint8_t,uint8_t> psp_Machine_dstPixel4(psp_Machine* m,psp_geState* s,uint32_t base,uint32_t off);
float psp_edge(psp_vert a,psp_vert b,psp_vert c);
uint8_t psp_mul8(uint8_t a,uint8_t b);
float psp_fmin3(float a,float b,float c);
float psp_fmax3(float a,float b,float c);
float psp_fmin(float a,float b);
float psp_fmax(float a,float b);
int64_t psp_clampI(int64_t v,int64_t lo,int64_t hi);
int32_t psp_minI32(int32_t a,int32_t b);
int32_t psp_maxI32(int32_t a,int32_t b);
int32_t psp_absI32(int32_t a);
void psp_Machine_drawPatch(psp_Machine* m,psp_geState* s,uint32_t arg,bool spline);
std::tuple<float,float,float,float,float> psp_splineSurf(Slice<psp_vert> ctrl,int64_t nu,int64_t nv,float tu,float tv,bool clampU,bool clampV,bool hasUV);
Slice<float> psp_splineKnots(int64_t n,bool clamped);
Slice<float> psp_splineBasis(Slice<float> knot,int64_t n,float t);
std::tuple<float,float,float,float,float> psp_bezierSurf(Slice<psp_vert> ctrl,int64_t nu,int64_t nv,float tu,float tv,bool hasUV);
Slice<float> psp_bezierBasis(int64_t n,float t);
uint32_t psp_geState_zAddress(psp_geState* s);
void psp_Machine_blockTransfer(psp_Machine* m,psp_geState* s,bool is32);
bool psp_Machine_zTest(psp_Machine* m,psp_geState* s,int64_t x,int64_t y,float z);
void psp_Machine_zWrite(psp_Machine* m,psp_geState* s,int64_t x,int64_t y,float z);
float psp_clampF(float v,float lo,float hi);
void psp_Machine_rasterList(psp_Machine* m,psp_GeList list);
uint32_t psp_geState_fbAddress(psp_geState* s);
void psp_Machine_loadClut(psp_Machine* m,psp_geState* s,uint32_t blocks);
void psp_Machine_dumpPrim(psp_Machine* m,psp_geState* s,uint32_t ptype,int64_t count);
void psp_Machine_drawPrim(psp_Machine* m,psp_geState* s,uint32_t arg);
bool psp_vert_through(psp_vert v);
psp_vertexLayout psp_layoutOf(uint32_t vtype);
Slice<psp_vert> psp_Machine_decodeVerts(psp_Machine* m,psp_geState* s,int64_t count);
std::tuple<float,float> psp_readUV(psp_Machine* m,uint32_t p,uint32_t fmt);
std::tuple<uint8_t,uint8_t,uint8_t,uint8_t> psp_readColor(psp_Machine* m,uint32_t p,uint32_t fmt);
std::tuple<uint8_t,uint8_t,uint8_t> psp_geState_lightVertex(psp_geState* s,float nx,float ny,float nz);
std::tuple<float,float,float> psp_readNormal(psp_Machine* m,uint32_t p,uint32_t fmt);
std::tuple<float,float,float> psp_readPos(psp_Machine* m,uint32_t p,uint32_t fmt);
std::tuple<float,float,float,float> psp_clipCoords(std::array<float,16> mvp,float x,float y,float z);
std::tuple<float,float,float> psp_project(psp_geState* s,float cx,float cy,float cz,float cw);
float psp_fogCoef(psp_geState* s,float w);
float psp_invW(float w);
float psp_nearDist(psp_vert v);
Slice<std::array<psp_vert,3>> psp_clipTriNear(psp_geState* s,psp_vert a,psp_vert b,psp_vert c);
psp_vert psp_lerpVert(psp_geState* s,psp_vert p,psp_vert q,float t);
void psp_ident(std::array<float,16>* m);
void psp_matPush(std::array<float,16>* m,int64_t* idx,uint32_t arg);
void psp_matPush16(std::array<float,16>* m,int64_t* idx,uint32_t arg);
std::array<float,16> psp_mul4(std::array<float,16> a,std::array<float,16> b);
float psp_f24(uint32_t arg);
float psp_f32(psp_Machine* m,uint32_t a);
psp_geWire* psp_geWireOf(psp_geState* s);
psp_geState* psp_geWire_state(psp_geWire* w);
void psp_Machine_registerSubIntr(psp_Machine* m,uint32_t intno,uint32_t subno,uint32_t handler,uint32_t arg);
void psp_Machine_setSubIntrEnabled(psp_Machine* m,uint32_t intno,uint32_t subno,bool on);
void psp_Machine_deliverVBlank(psp_Machine* m);
void psp_Machine_SetVolume(psp_Machine* m,psp_Volume* v);
std::string psp_devicePath(std::string p);
uint32_t psp_Machine_ioOpen(psp_Machine* m,std::string path);
uint32_t psp_Machine_ioClose(psp_Machine* m,uint32_t fd);
uint32_t psp_Machine_ioOpenAsync(psp_Machine* m,std::string path);
uint32_t psp_Machine_ioReadAsync(psp_Machine* m,uint32_t fd,uint32_t buf,uint32_t n);
uint32_t psp_Machine_ioLseekAsync(psp_Machine* m,uint32_t fd,int64_t off,uint32_t whence);
uint32_t psp_Machine_ioCloseAsync(psp_Machine* m,uint32_t fd);
uint32_t psp_Machine_ioWaitAsync(psp_Machine* m,uint32_t fd,uint32_t resPtr);
uint32_t psp_Machine_ioPollAsync(psp_Machine* m,uint32_t fd,uint32_t resPtr);
uint32_t psp_Machine_ioRead(psp_Machine* m,uint32_t fd,uint32_t buf,uint32_t n);
uint32_t psp_Machine_ioWrite(psp_Machine* m,uint32_t fd,uint32_t buf,uint32_t n);
void psp_Machine_fillStat(psp_Machine* m,psp_Entry e,uint32_t stat);
uint32_t psp_Machine_ioGetstat(psp_Machine* m,std::string path,uint32_t stat);
uint32_t psp_Machine_ioDopen(psp_Machine* m,std::string path);
uint32_t psp_Machine_ioDread(psp_Machine* m,uint32_t fd,uint32_t dirent);
uint32_t psp_Machine_ioDclose(psp_Machine* m,uint32_t fd);
int64_t psp_Machine_ioLseek(psp_Machine* m,uint32_t fd,int64_t offset,uint32_t whence);
std::tuple<Slice<uint8_t>,Error> psp_byteBlocks_ReadBlock(psp_byteBlocks b,int64_t n);
std::string psp_Entry_String(psp_Entry e);
std::tuple<psp_Volume*,Error> psp_OpenVolume(BlockSource* src);
uint32_t psp_le32(Slice<uint8_t> b);
std::tuple<Slice<psp_Entry>,Error> psp_Volume_dirEntries(psp_Volume* v,int64_t lba,int64_t size,std::string dirPath);
bool psp_nameEqual(std::string entry,std::string want);
Slice<std::string> psp_splitPath(std::string p);
std::tuple<psp_Entry,Error> psp_Volume_resolve(psp_Volume* v,std::string path);
std::tuple<int64_t,int64_t,bool> psp_parseLbnPath(std::string path);
std::tuple<Slice<psp_Entry>,Error> psp_Volume_ReadDir(psp_Volume* v,std::string path);
Error psp_Volume_Walk(psp_Volume* v,std::function<Error(psp_Entry)> fn);
Error psp_Volume_walk(psp_Volume* v,int64_t lba,int64_t size,std::string dirPath,std::function<Error(psp_Entry)> fn);
std::tuple<int64_t,Error> psp_Volume_ReadFileAt(psp_Volume* v,psp_Entry e,int64_t off,Slice<uint8_t> p);
std::tuple<Slice<uint8_t>,Error> psp_Volume_ReadFile(psp_Volume* v,std::string path);
bool psp_Machine_evMatch(psp_Machine* m,psp_kobject* o,uint32_t bits,uint32_t mode,uint32_t outPtr);
uint32_t psp_nidOf(std::string name);
std::string psp_nidName(std::string lib,uint32_t nid);
void psp_Machine_installStubs(psp_Machine* m,psp_Module* mod);
bool psp_Machine_handleSyscall(psp_Machine* m,allegrex_CPU* c,uint32_t code);
uint32_t psp_Machine_arg(psp_Machine* m,uint32_t i);
void psp_Machine_setRet(psp_Machine* m,uint32_t v);
void psp_Machine_setRet64(psp_Machine* m,uint64_t v);
uint64_t psp_Machine_clock(psp_Machine* m);
uint32_t psp_Machine_newHandle(psp_Machine* m,psp_kobject* o);
std::string psp_Machine_formatPrintf(psp_Machine* m,std::string format,uint32_t firstArg);
std::string psp_Machine_cstr(psp_Machine* m,uint32_t addr);
std::function<void(psp_Machine*)> psp_handlerFor(std::string name);
Error psp_kirk7(Slice<uint8_t> dst,Slice<uint8_t> src,int64_t keyId);
std::tuple<Slice<uint8_t>,Error> psp_kirkCMD1(Slice<uint8_t> in);
uint32_t psp_phys(uint32_t addr);
psp_Machine* psp_NewMachine();
void psp_Machine_SetPadScript(psp_Machine* m,Slice<psp_PadEvent> evs);
uint32_t psp_Machine_Vblanks(psp_Machine* m);
uint8_t psp_Machine_Read(psp_Machine* m,uint32_t addr);
void psp_Machine_Write(psp_Machine* m,uint32_t addr,uint8_t v);
std::string psp_Machine_TTY(psp_Machine* m);
std::string psp_Machine_CurrentThread(psp_Machine* m);
Slice<std::string> psp_Machine_KObjects(psp_Machine* m);
Slice<std::string> psp_Machine_Threads(psp_Machine* m);
Error psp_Machine_LoadModule(psp_Machine* m,psp_Module* mod);
void psp_Machine_writeRAM(psp_Machine* m,uint32_t addr,uint8_t v);
void psp_Machine_SetImageHash(psp_Machine* m,std::string h);
uint32_t psp_Machine_beGuest32(psp_Machine* m,uint32_t addr);
void psp_Machine_writeMpegAu(psp_Machine* m,uint32_t au,uint32_t pts,uint32_t esBuf,uint32_t esSize);
uint32_t psp_Machine_mpegRingbufferConstruct(psp_Machine* m,uint32_t rb,uint32_t packets,uint32_t data,uint32_t size,uint32_t cb,uint32_t cbArg);
uint32_t psp_Machine_mpegRingbufferPut(psp_Machine* m,uint32_t rb,uint32_t n,uint32_t avail);
void psp_Machine_atracParseRiff(psp_Machine* m,psp_atracState* a);
uint32_t psp_Machine_atracDecode(psp_Machine* m,uint32_t id,uint32_t out,uint32_t samplesPtr,uint32_t endPtr,uint32_t remainPtr);
uint32_t psp_Machine_mpegGetAu(psp_Machine* m,uint32_t au);
time_Time psp_Machine_profStart(psp_Machine* m);
void psp_Machine_profEnd(psp_Machine* m,int64_t bucket,time_Time t);
int64_t psp_Machine_profGeNs(psp_Machine* m);
void psp_Machine_profEndSyscall(psp_Machine* m,time_Time t,int64_t geBefore,int64_t gen);
void psp_Machine_profRunEnter(psp_Machine* m);
void psp_Machine_profRunExit(psp_Machine* m);
void psp_Machine_profFrame(psp_Machine* m);
psp_FrameProfile psp_Machine_FrameProfile(psp_Machine* m);
void psp_Machine_SetProfile(psp_Machine* m,bool on);
std::tuple<Slice<uint8_t>,uint32_t,Error> psp_DecryptPRX(Slice<uint8_t> raw);
std::tuple<Slice<uint8_t>,bool,Error> psp_decryptTagType(Slice<uint8_t> raw,psp_tagInfo ti,bool preDecrypt);
bool psp_bytesEqual(Slice<uint8_t> a,Slice<uint8_t> b);
std::string psp_Result_String(psp_Result r);
bool psp_Machine_mapped(psp_Machine* m,uint32_t p);
void psp_Machine_startThread(psp_Machine* m,uint32_t uid,psp_kobject* o,uint32_t argLen,uint32_t argPtr);
uint32_t psp_Machine_threadK0(psp_Machine* m,uint32_t uid,uint32_t stackTop);
bool psp_Machine_schedule(psp_Machine* m,psp_threadState currentBecomes);
void psp_Machine_yieldCurrent(psp_Machine* m,psp_threadState newState);
uint32_t psp_Machine_currentThreadID(psp_Machine* m);
void psp_Machine_onThreadExit(psp_Machine* m);
constexpr allegrex_Flow allegrex_FlowSeq=0ULL;
constexpr allegrex_Flow allegrex_FlowBranch=1ULL;
constexpr allegrex_Flow allegrex_FlowJump=2ULL;
constexpr allegrex_Flow allegrex_FlowCall=3ULL;
constexpr allegrex_Flow allegrex_FlowReturn=4ULL;
constexpr allegrex_Flow allegrex_FlowIndJump=5ULL;
constexpr allegrex_Flow allegrex_FlowIndCall=6ULL;
constexpr allegrex_Flow allegrex_FlowStop=7ULL;
constexpr int64_t allegrex_cop0BadVaddr=8ULL;
constexpr int64_t allegrex_cop0Status=12ULL;
constexpr int64_t allegrex_cop0Cause=13ULL;
constexpr int64_t allegrex_cop0EPC=14ULL;
constexpr int64_t allegrex_cop0PRId=15ULL;
constexpr int64_t allegrex_excInt=0ULL;
constexpr int64_t allegrex_excAdEL=4ULL;
constexpr int64_t allegrex_excAdES=5ULL;
constexpr int64_t allegrex_excSys=8ULL;
constexpr int64_t allegrex_excBp=9ULL;
constexpr int64_t allegrex_excRI=10ULL;
constexpr int64_t allegrex_excCpU=11ULL;
constexpr int64_t allegrex_excOv=12ULL;
constexpr int64_t allegrex_vecRAM=2147483776ULL;
constexpr int64_t allegrex_vecROM=3217031552ULL;
std::array<std::string,32> allegrex_regName=std::array<std::string,32>{std::string("zero",4),std::string("at",2),std::string("v0",2),std::string("v1",2),std::string("a0",2),std::string("a1",2),std::string("a2",2),std::string("a3",2),std::string("t0",2),std::string("t1",2),std::string("t2",2),std::string("t3",2),std::string("t4",2),std::string("t5",2),std::string("t6",2),std::string("t7",2),std::string("s0",2),std::string("s1",2),std::string("s2",2),std::string("s3",2),std::string("s4",2),std::string("s5",2),std::string("s6",2),std::string("s7",2),std::string("t8",2),std::string("t9",2),std::string("k0",2),std::string("k1",2),std::string("gp",2),std::string("sp",2),std::string("fp",2),std::string("ra",2)};
constexpr int64_t allegrex_vfpuCtlSPfx=0ULL;
constexpr int64_t allegrex_vfpuCtlTPfx=1ULL;
constexpr int64_t allegrex_vfpuCtlDPfx=2ULL;
constexpr int64_t allegrex_vfpuCtlCC=3ULL;
constexpr int64_t allegrex_pfxIdentity=228ULL;
std::array<uint32_t,16> allegrex_vfpuCtrlMask=std::array<uint32_t,16>{cast<uint32_t>(1048575ULL),cast<uint32_t>(1048575ULL),cast<uint32_t>(4095ULL),cast<uint32_t>(63ULL),cast<uint32_t>(4294967295ULL),cast<uint32_t>(0ULL),cast<uint32_t>(0ULL),cast<uint32_t>(0ULL),cast<uint32_t>(1073741823ULL),cast<uint32_t>(1073741823ULL),cast<uint32_t>(1073741823ULL),cast<uint32_t>(1073741823ULL),cast<uint32_t>(1073741823ULL),cast<uint32_t>(1073741823ULL),cast<uint32_t>(1073741823ULL),cast<uint32_t>(1073741823ULL)};
std::array<float,8> allegrex_pfxConstants=std::array<float,8>{cast<float>(0ULL),cast<float>(1ULL),cast<float>(2ULL),0.5,cast<float>(3ULL),(1.0 / 3.0),0.25,(1.0 / 6.0)};
std::array<std::string,5> allegrex_vfpuSuffix=std::array<std::string,5>{std::string("",0),std::string("s",1),std::string("p",1),std::string("t",1),std::string("q",1)};
std::array<float,32> allegrex_vcstConstants=std::array<float,32>{cast<float>(0ULL),go_math_MaxFloat32,go_math_Sqrt2,cast<float>(go_math_Sqrt(0.5)),cast<float>((cast<double>(2ULL) / go_math_Sqrt(go_math_Pi))),cast<float>((cast<double>(2ULL) / go_math_Pi)),cast<float>((cast<double>(1ULL) / go_math_Pi)),(go_math_Pi / cast<double>(4ULL)),(go_math_Pi / cast<double>(2ULL)),go_math_Pi,go_math_E,go_math_Log2E,go_math_Log10E,go_math_Ln2,cast<float>(go_math_Log(cast<double>(10ULL))),(cast<double>(2ULL) * go_math_Pi),(go_math_Pi / cast<double>(6ULL)),cast<float>(go_math_Log10(cast<double>(2ULL))),cast<float>((go_math_Log(cast<double>(10ULL)) / go_math_Log(cast<double>(2ULL)))),cast<float>((go_math_Sqrt(cast<double>(3ULL)) / cast<double>(2ULL)))};
Slice<std::string> psp_textureFormats=Slice<std::string>{std::string("rgb565",6),std::string("rgba5551",8),std::string("rgba4444",8),std::string("rgba8888",8),std::string("clut4",5),std::string("clut8",5)};
constexpr int64_t psp_etExec=2ULL;
constexpr int64_t psp_etPRX=65440ULL;
constexpr int64_t psp_emMIPS=8ULL;
constexpr int64_t psp_ptLoad=1ULL;
constexpr int64_t psp_rMIPS32=2ULL;
constexpr int64_t psp_rMIPS26=4ULL;
constexpr int64_t psp_rMIPSHI16=5ULL;
constexpr int64_t psp_rMIPSLO16=6ULL;
constexpr int64_t psp_shtPRXReloc=1879048352ULL;
constexpr int64_t psp_psm5650=0ULL;
constexpr int64_t psp_psm5551=1ULL;
constexpr int64_t psp_psm4444=2ULL;
constexpr int64_t psp_psm8888=3ULL;
constexpr int64_t psp_dispW=480ULL;
constexpr int64_t psp_dispH=272ULL;
constexpr int64_t psp_geNOP=0ULL;
constexpr int64_t psp_geVADDR=1ULL;
constexpr int64_t psp_geIADDR=2ULL;
constexpr int64_t psp_gePRIM=4ULL;
constexpr int64_t psp_geJUMP=8ULL;
constexpr int64_t psp_geBJUMP=9ULL;
constexpr int64_t psp_geCALL=10ULL;
constexpr int64_t psp_geRET=11ULL;
constexpr int64_t psp_geFINISH=15ULL;
constexpr int64_t psp_geEND=12ULL;
constexpr int64_t psp_geBASE=16ULL;
constexpr int64_t psp_geVTYPE=18ULL;
constexpr int64_t psp_geOFFADDR=19ULL;
constexpr int64_t psp_geREGION1=21ULL;
constexpr int64_t psp_geREGION2=22ULL;
constexpr int64_t psp_geCLEAR=211ULL;
constexpr int64_t psp_geWORLD_N=58ULL;
constexpr int64_t psp_geWORLD_D=59ULL;
constexpr int64_t psp_geVIEW_N=60ULL;
constexpr int64_t psp_geVIEW_D=61ULL;
constexpr int64_t psp_gePROJ_N=62ULL;
constexpr int64_t psp_gePROJ_D=63ULL;
constexpr int64_t psp_geFBP=156ULL;
constexpr int64_t psp_geFBW=157ULL;
std::array<std::string,256> psp_geCmdNames=[]()->std::array<std::string,256>{
std::array<std::string,256> n = [](){std::array<std::string,256> v{};v[0]=std::string("NOP",3);v[1]=std::string("VADDR",5);v[2]=std::string("IADDR",5);v[4]=std::string("PRIM",4);v[5]=std::string("BEZIER",6);v[6]=std::string("SPLINE",6);v[7]=std::string("BBOX",4);v[8]=std::string("JUMP",4);v[9]=std::string("BJUMP",5);v[10]=std::string("CALL",4);v[11]=std::string("RET",3);v[12]=std::string("END",3);v[15]=std::string("FINISH",6);v[16]=std::string("BASE",4);v[18]=std::string("VTYPE",5);v[19]=std::string("OFFADDR",7);v[20]=std::string("ORIGIN",6);v[21]=std::string("REGION1",7);v[22]=std::string("REGION2",7);v[23]=std::string("LIGHTING",8);v[28]=std::string("CLUTON",6);v[29]=std::string("CULLON",6);v[30]=std::string("TEXENABLE",9);v[31]=std::string("FOGON",5);v[32]=std::string("DITHERON",8);v[33]=std::string("BLENDON",7);v[34]=std::string("ALPHATESTON",11);v[35]=std::string("ZTESTON",7);v[36]=std::string("STENCILON",9);v[37]=std::string("AAON",4);v[38]=std::string("PATCHCULLON",11);v[39]=std::string("COLORTESTON",11);v[40]=std::string("LOGICOPON",9);v[46]=std::string("BONEMTXN",8);v[47]=std::string("BONEMTXD",8);v[54]=std::string("PATCHDIV",8);v[55]=std::string("PATCHPRIM",9);v[56]=std::string("PATCHFACE",9);v[58]=std::string("WORLDN",6);v[59]=std::string("WORLDD",6);v[60]=std::string("VIEWN",5);v[61]=std::string("VIEWD",5);v[62]=std::string("PROJN",5);v[63]=std::string("PROJD",5);v[64]=std::string("TEXMTXN",7);v[65]=std::string("TEXMTXD",7);v[66]=std::string("VPXSCALE",8);v[67]=std::string("VPYSCALE",8);v[68]=std::string("VPZSCALE",8);v[69]=std::string("VPXCENTER",9);v[70]=std::string("VPYCENTER",9);v[71]=std::string("VPZCENTER",9);v[72]=std::string("TEXSCALEU",9);v[73]=std::string("TEXSCALEV",9);v[74]=std::string("TEXOFFSETU",10);v[75]=std::string("TEXOFFSETV",10);v[76]=std::string("OFFSETX",7);v[77]=std::string("OFFSETY",7);v[80]=std::string("SHADEMODE",9);v[81]=std::string("NORMALREV",9);v[83]=std::string("MATCOLORMODE",12);v[84]=std::string("MATEMISSIVE",11);v[85]=std::string("MATAMBIENT",10);v[86]=std::string("MATDIFFUSE",10);v[87]=std::string("MATSPECULAR",11);v[88]=std::string("MATALPHA",8);v[91]=std::string("MATSPECCOEF",11);v[92]=std::string("AMBIENTCOL",10);v[93]=std::string("AMBIENTALPHA",12);v[94]=std::string("LIGHTMODE",9);v[143]=std::string("SPOTCOEF",8);v[155]=std::string("CULLFACE",8);v[156]=std::string("FBP",3);v[157]=std::string("FBW",3);v[158]=std::string("ZBP",3);v[159]=std::string("ZBW",3);v[176]=std::string("CLUTADDR",8);v[177]=std::string("CLUTADDRH",9);v[178]=std::string("TRSRC",5);v[179]=std::string("TRSRCW",6);v[180]=std::string("TRDST",5);v[181]=std::string("TRDSTW",6);v[192]=std::string("TEXMAPMODE",10);v[193]=std::string("TEXSHADELS",10);v[194]=std::string("TEXMODE",7);v[195]=std::string("TEXFORMAT",9);v[196]=std::string("LOADCLUT",8);v[197]=std::string("CLUTFORMAT",10);v[198]=std::string("TEXFILTER",9);v[199]=std::string("TEXWRAP",7);v[200]=std::string("TEXLEVEL",8);v[201]=std::string("TEXFUNC",7);v[202]=std::string("TEXENVCOL",9);v[203]=std::string("TEXFLUSH",8);v[204]=std::string("TEXSYNC",7);v[205]=std::string("FOG1",4);v[206]=std::string("FOG2",4);v[207]=std::string("FOGCOLOR",8);v[210]=std::string("FBPIXFMT",8);v[211]=std::string("CLEARMODE",9);v[212]=std::string("SCISSOR1",8);v[213]=std::string("SCISSOR2",8);v[214]=std::string("NEARZ",5);v[215]=std::string("FARZ",4);v[219]=std::string("ATEST",5);v[220]=std::string("STEST",5);v[221]=std::string("SOP",3);v[222]=std::string("ZTEST",5);v[223]=std::string("BLENDFUNC",9);v[224]=std::string("BLENDFIXA",9);v[225]=std::string("BLENDFIXB",9);v[226]=std::string("DITH0",5);v[227]=std::string("DITH1",5);v[228]=std::string("DITH2",5);v[229]=std::string("DITH3",5);v[230]=std::string("LOGICOP",7);v[231]=std::string("ZMASK",5);v[232]=std::string("PMSKC",5);v[233]=std::string("PMSKA",5);v[234]=std::string("TRSTART",7);v[235]=std::string("TRSRCPOS",8);v[236]=std::string("TRDSTPOS",8);v[238]=std::string("TRSIZE",6);return v;}();
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(8ULL));i++){
n[cast<int64_t>((cast<int64_t>(160ULL) + i))] = go_fmt_Sprintf(std::string("TEXADDR%d",9),i);
n[cast<int64_t>((cast<int64_t>(168ULL) + i))] = go_fmt_Sprintf(std::string("TEXBW%d",7),i);
n[cast<int64_t>((cast<int64_t>(184ULL) + i))] = go_fmt_Sprintf(std::string("TEXSIZE%d",9),i);
}
}{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(4ULL));i++){
n[cast<int64_t>((cast<int64_t>(24ULL) + i))] = go_fmt_Sprintf(std::string("LIGHT%dON",9),i);
n[cast<int64_t>((cast<int64_t>(95ULL) + i))] = go_fmt_Sprintf(std::string("LIGHTTYPE%d",11),i);
n[cast<int64_t>((cast<int64_t>(144ULL) + i))] = go_fmt_Sprintf(std::string("LIGHTDIF%d",10),i);
n[cast<int64_t>((cast<int64_t>(148ULL) + i))] = go_fmt_Sprintf(std::string("LIGHTSPC%d",10),i);
n[cast<int64_t>((cast<int64_t>(152ULL) + i))] = go_fmt_Sprintf(std::string("LIGHTAMB%d",10),i);
{int64_t c = cast<int64_t>(0ULL);for (;(c < cast<int64_t>(3ULL));c++){
n[cast<int64_t>((cast<int64_t>((cast<int64_t>(99ULL) + cast<int64_t>((i * cast<int64_t>(3ULL))))) + c))] = go_fmt_Sprintf(std::string("LIGHTPOS%d%c",12),i,std::string("XYZ",3)[c]);
n[cast<int64_t>((cast<int64_t>((cast<int64_t>(111ULL) + cast<int64_t>((i * cast<int64_t>(3ULL))))) + c))] = go_fmt_Sprintf(std::string("LIGHTDIR%d%c",12),i,std::string("XYZ",3)[c]);
n[cast<int64_t>((cast<int64_t>((cast<int64_t>(123ULL) + cast<int64_t>((i * cast<int64_t>(3ULL))))) + c))] = go_fmt_Sprintf(std::string("LIGHTATT%d%c",12),i,std::string("XYZ",3)[c]);
}
}}
}return n;
}
();
auto [psp_geDebugN,tmp79]=go_strconv_Atoi(go_os_Getenv(std::string("PSP_GE_DEBUG",12)));
bool psp_geCullFlip=(go_os_Getenv(std::string("PSP_GE_CULLFLIP",15)) != std::string("",0));
bool psp_geNoCull=(go_os_Getenv(std::string("PSP_GE_NOCULL",13)) != std::string("",0));
auto [psp_geWorldDump,tmp80]=go_strconv_Atoi(go_os_Getenv(std::string("PSP_GE_WORLD",12)));
bool psp_geNoZ=(go_os_Getenv(std::string("PSP_GE_NOZ",10)) != std::string("",0));
bool psp_geNoBias=(go_os_Getenv(std::string("PSP_GE_NOBIAS",13)) != std::string("",0));
bool psp_geXferLog=(go_os_Getenv(std::string("PSP_GE_XFER",11)) != std::string("",0));
auto [psp_geOnlyTex,tmp81]=go_strconv_ParseUint(go_os_Getenv(std::string("PSP_GE_ONLYTEX",14)),cast<int64_t>(16ULL),cast<int64_t>(32ULL));
auto [psp_geOnlyVA,tmp82]=go_strconv_ParseUint(go_os_Getenv(std::string("PSP_GE_ONLYVA",13)),cast<int64_t>(16ULL),cast<int64_t>(32ULL));
int64_t psp_geProbeX=cast<int64_t>(-1ULL);
int64_t psp_geProbeY=cast<int64_t>(-1ULL);
int64_t psp_gePrimDump=cast<int64_t>(-1ULL);
constexpr int64_t psp_cVADDR=1ULL;
constexpr int64_t psp_cIADDR=2ULL;
constexpr int64_t psp_cPRIM=4ULL;
constexpr int64_t psp_cBEZIER=5ULL;
constexpr int64_t psp_cSPLINE=6ULL;
constexpr int64_t psp_cBASE=16ULL;
constexpr int64_t psp_cLIGHTING=23ULL;
constexpr int64_t psp_cPATCHDIV=54ULL;
constexpr int64_t psp_cPATCHPRIM=55ULL;
constexpr int64_t psp_cMATDIFFUSE=86ULL;
constexpr int64_t psp_cVTYPE=18ULL;
constexpr int64_t psp_cOFFADDR=19ULL;
constexpr int64_t psp_cCULLON=29ULL;
constexpr int64_t psp_cTEXENABLE=30ULL;
constexpr int64_t psp_cBLENDON=33ULL;
constexpr int64_t psp_cCULLFACE=155ULL;
constexpr int64_t psp_cTEXMTXN=64ULL;
constexpr int64_t psp_cTEXMTXD=65ULL;
constexpr int64_t psp_cWORLDN=58ULL;
constexpr int64_t psp_cWORLDD=59ULL;
constexpr int64_t psp_cVIEWN=60ULL;
constexpr int64_t psp_cVIEWD=61ULL;
constexpr int64_t psp_cPROJN=62ULL;
constexpr int64_t psp_cPROJD=63ULL;
constexpr int64_t psp_cVPXSCALE=66ULL;
constexpr int64_t psp_cVPYSCALE=67ULL;
constexpr int64_t psp_cVPZSCALE=68ULL;
constexpr int64_t psp_cVPXCENTER=69ULL;
constexpr int64_t psp_cVPYCENTER=70ULL;
constexpr int64_t psp_cVPZCENTER=71ULL;
constexpr int64_t psp_cTEXSCALEU=72ULL;
constexpr int64_t psp_cTEXSCALEV=73ULL;
constexpr int64_t psp_cTEXOFFSETU=74ULL;
constexpr int64_t psp_cTEXOFFSETV=75ULL;
constexpr int64_t psp_cOFFSETX=76ULL;
constexpr int64_t psp_cOFFSETY=77ULL;
constexpr int64_t psp_cSHADEMODE=80ULL;
constexpr int64_t psp_cMATAMBIENT=85ULL;
constexpr int64_t psp_cMATALPHA=88ULL;
constexpr int64_t psp_cTEXADDR0=160ULL;
constexpr int64_t psp_cTEXBW0=168ULL;
constexpr int64_t psp_cTEXLEVEL=200ULL;
constexpr int64_t psp_cCLUTADDR=176ULL;
constexpr int64_t psp_cCLUTADDRH=177ULL;
constexpr int64_t psp_cTEXSIZE0=184ULL;
constexpr int64_t psp_cTEXMAPMODE=192ULL;
constexpr int64_t psp_cTEXMODE=194ULL;
constexpr int64_t psp_cTEXFORMAT=195ULL;
constexpr int64_t psp_cLOADCLUT=196ULL;
constexpr int64_t psp_cCLUTFORMAT=197ULL;
constexpr int64_t psp_cTEXFILTER=198ULL;
constexpr int64_t psp_cTEXWRAP=199ULL;
constexpr int64_t psp_cTEXFUNC=201ULL;
constexpr int64_t psp_cTEXENVCOL=202ULL;
constexpr int64_t psp_cTRSRC=178ULL;
constexpr int64_t psp_cTRSRCW=179ULL;
constexpr int64_t psp_cTRDST=180ULL;
constexpr int64_t psp_cTRDSTW=181ULL;
constexpr int64_t psp_cTRSRCPOS=235ULL;
constexpr int64_t psp_cTRDSTPOS=236ULL;
constexpr int64_t psp_cTRSIZE=238ULL;
constexpr int64_t psp_cTRSTART=234ULL;
constexpr int64_t psp_cFOGON=31ULL;
constexpr int64_t psp_cFOG1=205ULL;
constexpr int64_t psp_cFOG2=206ULL;
constexpr int64_t psp_cFOGCOLOR=207ULL;
constexpr int64_t psp_cLIGHT0ON=24ULL;
constexpr int64_t psp_cAMBIENTCOL=92ULL;
constexpr int64_t psp_cMATEMISSIVE=84ULL;
constexpr int64_t psp_cLIGHTTYPE0=95ULL;
constexpr int64_t psp_cLIGHTPOS0=99ULL;
constexpr int64_t psp_cLIGHTDIF0=144ULL;
constexpr int64_t psp_cSTENCILON=36ULL;
constexpr int64_t psp_cSTEST=220ULL;
constexpr int64_t psp_cSOP=221ULL;
constexpr int64_t psp_cALPHATESTON=34ULL;
constexpr int64_t psp_cATEST=219ULL;
constexpr int64_t psp_cFBPTR=156ULL;
constexpr int64_t psp_cFBWIDTH=157ULL;
constexpr int64_t psp_cZBPTR=158ULL;
constexpr int64_t psp_cZBWIDTH=159ULL;
constexpr int64_t psp_cZTESTON=35ULL;
constexpr int64_t psp_cFBPIXFMT=210ULL;
constexpr int64_t psp_cCLEARMODE=211ULL;
constexpr int64_t psp_cSCISSOR1=212ULL;
constexpr int64_t psp_cSCISSOR2=213ULL;
constexpr int64_t psp_cZTEST=222ULL;
constexpr int64_t psp_cBLENDFUNC=223ULL;
constexpr int64_t psp_cBLENDFIXA=224ULL;
constexpr int64_t psp_cBLENDFIXB=225ULL;
constexpr int64_t psp_cZMASK=231ULL;
constexpr int64_t psp_cPMSKC=232ULL;
constexpr int64_t psp_cPMSKA=233ULL;
int64_t psp_gePrimSeq={};
Slice<uint32_t> psp_geWordTrail={};
constexpr double psp_nearEps=1.00000000000000008e-05;
constexpr int64_t psp_vblankIntr=30ULL;
constexpr int64_t psp_intrExitAddr=252706816ULL;
constexpr int64_t psp_fdStdout=1ULL;
constexpr int64_t psp_fdStderr=2ULL;
constexpr int64_t psp_fdFirstFile=4ULL;
constexpr int64_t psp_errIoNoEnt=2147549186ULL;
constexpr int64_t psp_errIoBadFd=2147549193ULL;
constexpr int64_t psp_errIoNoAsync=2147615521ULL;
constexpr int64_t psp_isoBlockSize=2048ULL;
constexpr int64_t psp_pvdLBA=16ULL;
std::string psp_semaTrace=go_os_Getenv(std::string("PSP_SEMA_TRACE",14));
std::string psp_syscallTrace=go_os_Getenv(std::string("PSP_SYSCALL_TRACE",17));
Slice<std::string> psp_knownNIDs=Slice<std::string>{std::string("sceKernelAllocPartitionMemory",29),std::string("sceKernelFreePartitionMemory",28),std::string("sceKernelGetBlockHeadAddr",25),std::string("sceKernelTotalFreeMemSize",25),std::string("sceKernelMaxFreeMemSize",23),std::string("sceKernelPrintf",15),std::string("sceKernelDevkitVersion",22),std::string("sceKernelCreateThread",21),std::string("sceKernelStartThread",20),std::string("sceKernelExitThread",19),std::string("sceKernelExitDeleteThread",25),std::string("sceKernelDeleteThread",21),std::string("sceKernelGetThreadId",20),std::string("sceKernelDelayThread",20),std::string("sceKernelDelayThreadCB",22),std::string("sceKernelSleepThread",20),std::string("sceKernelSleepThreadCB",22),std::string("sceKernelWakeupThread",21),std::string("sceKernelTerminateDeleteThread",30),std::string("sceKernelCreateCallback",23),std::string("sceKernelReferThreadStatus",26),std::string("sceKernelGetThreadExitStatus",28),std::string("sceKernelGetGPI",15),std::string("sceKernelGetGPO",15),std::string("sceKernelCreateSema",19),std::string("sceKernelDeleteSema",19),std::string("sceKernelWaitSema",17),std::string("sceKernelWaitSemaCB",19),std::string("sceKernelSignalSema",19),std::string("sceKernelDeleteCallback",23),std::string("sceAudioChReserve",17),std::string("sceAudioChRelease",17),std::string("sceAudioOutputBlocking",22),std::string("sceAudioOutputPannedBlocking",28),std::string("sceAudioSetChannelDataLen",25),std::string("sceAudioGetChannelRestLen",25),std::string("sceAudioChangeChannelVolume",27),std::string("sceAudioOutput2Reserve",22),std::string("sceAudioOutput2OutputBlocking",29),std::string("sceAudioOutput2Release",22),std::string("sceAudioSRCChReserve",20),std::string("sceAudioSRCOutputBlocking",25),std::string("sceAudioSRCChRelease",20),std::string("sceKernelCreateEventFlag",24),std::string("sceKernelWaitEventFlag",22),std::string("sceKernelSetEventFlag",21),std::string("sceKernelClearEventFlag",23),std::string("sceKernelCreateMutex",20),std::string("sceKernelLockMutex",18),std::string("sceKernelUnlockMutex",20),std::string("sceKernelCreateMsgPipe",22),std::string("sceKernelGetSystemTimeWide",26),std::string("sceKernelGetSystemTimeLow",25),std::string("sceKernelGetThreadCurrentPriority",33),std::string("sceKernelChangeThreadPriority",29),std::string("sceKernelCreateVpl",18),std::string("sceKernelDeleteVpl",18),std::string("sceKernelAllocateVpl",20),std::string("sceKernelAllocateVplCB",22),std::string("sceKernelTryAllocateVpl",23),std::string("sceKernelFreeVpl",16),std::string("sceKernelReferVplStatus",23),std::string("sceKernelCreateFpl",18),std::string("sceKernelDeleteFpl",18),std::string("sceKernelAllocateFpl",20),std::string("sceKernelTryAllocateFpl",23),std::string("sceKernelFreeFpl",16),std::string("sceKernelChangeCurrentThreadAttr",32),std::string("sceKernelCreateMbx",18),std::string("sceKernelCreateVTimer",21),std::string("sceKernelPollSema",17),std::string("sceKernelSignalSema",19),std::string("sceKernelDeleteSema",19),std::string("sceKernelWaitEventFlagCB",24),std::string("sceKernelPollEventFlag",22),std::string("sceKernelDeleteEventFlag",24),std::string("sceKernelCheckCallback",22),std::string("sceKernelNotifyCallback",23),std::string("sceKernelRotateThreadReadyQueue",31),std::string("sceKernelWaitThreadEnd",22),std::string("sceKernelWaitThreadEndCB",24),std::string("sceKernelSuspendThread",22),std::string("sceKernelResumeThread",21),std::string("sceKernelGetSystemTime",22),std::string("sceKernelUSec2SysClock",22),std::string("sceKernelSysClock2USec",22),std::string("sceKernelDelaySysClockThread",28),std::string("sceKernelSetGPO",15),std::string("sceKernelCpuSuspendIntr",23),std::string("sceKernelCpuResumeIntr",22),std::string("sceKernelCpuResumeIntrWithSync",30),std::string("sceKernelRegisterSubIntrHandler",31),std::string("sceKernelReleaseSubIntrHandler",30),std::string("sceKernelEnableSubIntr",22),std::string("sceKernelDisableSubIntr",23),std::string("sceKernelExitGame",17),std::string("sceKernelRegisterExitCallback",29),std::string("sceKernelLoadModule",19),std::string("sceKernelStartModule",20),std::string("sceKernelSelfStopUnloadModule",29),std::string("sceKernelGetModuleIdByAddress",29),std::string("sceKernelGetModuleId",20),std::string("sceKernelSetCompiledSdkVersion",30),std::string("sceKernelSetCompilerVersion",27),std::string("sceKernelDcacheWritebackAll",27),std::string("sceKernelDcacheWritebackRange",29),std::string("sceKernelDcacheWritebackInvalidateAll",37),std::string("sceKernelIcacheInvalidateAll",28),std::string("sceKernelLibcTime",17),std::string("sceKernelLibcClock",18),std::string("sceKernelLibcGettimeofday",25),std::string("sceKernelUtilsMt19937Init",25),std::string("sceKernelUtilsMt19937UInt",25),std::string("sceDisplaySetMode",17),std::string("sceDisplaySetFrameBuf",21),std::string("sceDisplayGetFrameBuf",21),std::string("sceDisplayWaitVblankStart",25),std::string("sceDisplayWaitVblankStartCB",27),std::string("sceDisplayWaitVblank",20),std::string("sceDisplayWaitVblankCB",22),std::string("sceDisplayGetVcount",19),std::string("sceDisplayIsVblank",18),std::string("sceDisplayGetCurrentHcount",26),std::string("sceDisplayGetAccumulatedHcount",30),std::string("sceDisplayGetFramePerSec",24),std::string("sceGeEdramGetAddr",17),std::string("sceGeEdramGetSize",17),std::string("sceGeListEnQueue",16),std::string("sceGeListEnQueueHead",20),std::string("sceGeListSync",13),std::string("sceGeDrawSync",13),std::string("sceGeSetCallback",16),std::string("sceGeUnsetCallback",18),std::string("sceGeListUpdateStallAddr",24),std::string("sceGeBreak",10),std::string("sceGeContinue",13),std::string("sceGeGetCmd",11),std::string("sceGeGetMtx",11),std::string("sceCtrlSetSamplingCycle",23),std::string("sceCtrlSetSamplingMode",22),std::string("sceCtrlGetSamplingCycle",23),std::string("sceCtrlGetSamplingMode",22),std::string("sceCtrlReadBufferPositive",25),std::string("sceCtrlPeekBufferPositive",25),std::string("sceCtrlReadLatch",16),std::string("sceCtrlPeekLatch",16),std::string("sceCtrlSetIdleCancelThreshold",29),std::string("sceKernelStdout",15),std::string("sceKernelStdin",14),std::string("sceKernelStderr",15),std::string("sceIoOpen",9),std::string("sceIoClose",10),std::string("sceIoRead",9),std::string("sceIoWrite",10),std::string("sceIoLseek",10),std::string("sceIoLseek32",12),std::string("sceIoOpenAsync",14),std::string("sceIoCloseAsync",15),std::string("sceIoReadAsync",14),std::string("sceIoLseekAsync",15),std::string("sceIoLseek32Async",17),std::string("sceIoWaitAsync",14),std::string("sceIoWaitAsyncCB",16),std::string("sceIoPollAsync",14),std::string("sceIoGetAsyncStat",17),std::string("sceIoChangeAsyncPriority",24),std::string("sceIoSetAsyncCallback",21),std::string("sceIoGetstat",12),std::string("sceIoChdir",10),std::string("sceIoDevctl",11),std::string("sceIoIoctl",10),std::string("sceIoIoctlAsync",15),std::string("sceIoDopen",10),std::string("sceIoDread",10),std::string("sceIoDclose",11),std::string("sceIoSync",9),std::string("sceIoRemove",11),std::string("sceIoMkdir",10),std::string("sceUmdActivate",14),std::string("sceUmdDeactivate",16),std::string("sceUmdCheckMedium",17),std::string("sceUmdGetDriveStat",18),std::string("sceUmdWaitDriveStat",19),std::string("sceUmdWaitDriveStatCB",21),std::string("sceUmdWaitDriveStatWithTimer",28),std::string("sceUmdRegisterUMDCallBack",25),std::string("sceUmdUnRegisterUMDCallBack",27),std::string("sceUmdGetErrorStat",18),std::string("sceRtcGetCurrentTick",20),std::string("scePowerGetCpuClockFrequency",28),std::string("scePowerRegisterCallback",24),std::string("sceKernelVolatileMemLock",24),std::string("sceKernelVolatileMemTryLock",27),std::string("sceKernelVolatileMemUnlock",26),std::string("sceKernelPowerTick",18),std::string("sceKernelPowerLock",18),std::string("sceKernelPowerUnlock",20),std::string("sceUtilityGetSystemParamInt",27),std::string("sceUtilityGetSystemParamString",30),std::string("sceUtilityMsgDialogInitStart",28),std::string("sceUtilityMsgDialogGetStatus",28),std::string("sceUtilityMsgDialogUpdate",25),std::string("sceUtilityMsgDialogShutdownStart",32),std::string("sceUtilitySavedataInitStart",27),std::string("sceUtilitySavedataGetStatus",27),std::string("sceUtilitySavedataUpdate",24),std::string("sceUtilitySavedataShutdownStart",31),std::string("__sceSasInit",12),std::string("__sceSasCore",12),std::string("__sceSasCoreWithMix",19),std::string("__sceSasGetEndFlag",18),std::string("__sceSasGetAllEnvelopeHeights",29),std::string("__sceSasSetVoice",16),std::string("__sceSasSetVoicePCM",19),std::string("__sceSasSetPitch",16),std::string("__sceSasSetVolume",17),std::string("__sceSasSetSL",13),std::string("__sceSasSetADSR",15),std::string("__sceSasSetADSRmode",19),std::string("__sceSasSetSimpleADSR",21),std::string("__sceSasSetKeyOn",16),std::string("__sceSasSetKeyOff",17),std::string("__sceSasSetPause",16),std::string("__sceSasGetPauseFlag",20),std::string("__sceSasSetNoise",16),std::string("__sceSasSetGrain",16),std::string("__sceSasGetGrain",16),std::string("__sceSasSetOutputmode",21),std::string("__sceSasGetOutputmode",21),std::string("__sceSasRevType",15),std::string("__sceSasRevParam",16),std::string("__sceSasRevEVOL",15),std::string("__sceSasRevVON",14),std::string("sceMpegInit",11),std::string("sceMpegFinish",13),std::string("sceMpegQueryMemSize",19),std::string("sceMpegCreate",13),std::string("sceMpegDelete",13),std::string("sceMpegRegistStream",19),std::string("sceMpegUnRegistStream",21),std::string("sceMpegMallocAvcEsBuf",21),std::string("sceMpegFreeAvcEsBuf",19),std::string("sceMpegInitAu",13),std::string("sceMpegGetAvcAu",15),std::string("sceMpegGetAtracAu",17),std::string("sceMpegGetPcmAu",15),std::string("sceMpegQueryStreamOffset",24),std::string("sceMpegQueryStreamSize",22),std::string("sceMpegQueryAtracEsSize",23),std::string("sceMpegAvcDecode",16),std::string("sceMpegAvcDecodeMode",20),std::string("sceMpegAvcDecodeStop",20),std::string("sceMpegAvcDecodeYCbCr",21),std::string("sceMpegAvcDecodeStopYCbCr",25),std::string("sceMpegAvcQueryYCbCrSize",24),std::string("sceMpegAvcInitYCbCr",19),std::string("sceMpegAvcCsc",13),std::string("sceMpegAtracDecode",18),std::string("sceMpegChangeGetAvcAuMode",25),std::string("sceMpegChangeGetAuMode",22),std::string("sceMpegRingbufferQueryMemSize",29),std::string("sceMpegRingbufferConstruct",26),std::string("sceMpegRingbufferDestruct",25),std::string("sceMpegRingbufferPut",20),std::string("sceMpegRingbufferAvailableSize",30),std::string("sceMpegFlushStream",18),std::string("sceMpegFlushAllStream",21),std::string("sceAtracGetAtracID",18),std::string("sceAtracSetDataAndGetID",23),std::string("sceAtracSetHalfwayBufferAndGetID",32),std::string("sceAtracSetData",15),std::string("sceAtracSetHalfwayBuffer",24),std::string("sceAtracDecodeData",18),std::string("sceAtracGetRemainFrame",22),std::string("sceAtracGetStreamDataInfo",25),std::string("sceAtracAddStreamData",21),std::string("sceAtracGetSecondBufferInfo",27),std::string("sceAtracSetSecondBuffer",23),std::string("sceAtracGetNextDecodePosition",29),std::string("sceAtracGetSoundSample",22),std::string("sceAtracGetChannel",18),std::string("sceAtracGetMaxSample",20),std::string("sceAtracGetNextSample",21),std::string("sceAtracGetBitrate",18),std::string("sceAtracGetLoopStatus",21),std::string("sceAtracSetLoopNum",18),std::string("sceAtracResetPlayPosition",25),std::string("sceAtracGetInternalErrorInfo",28),std::string("sceAtracReleaseAtracID",22)};
Map<uint32_t,std::string> psp_nidLookup=[]()->Map<uint32_t,std::string>{
Map<uint32_t,std::string> m = Map<uint32_t,std::string>{};
{auto&& tmp196 = psp_knownNIDs;
for(int64_t tmp197=0;tmp197<len(tmp196);++tmp197){
auto n=tmp196[tmp197];m[psp_nidOf(n)] = n;
}}
return m;
}
();
constexpr int64_t psp_cmd1Mode=96ULL;
constexpr int64_t psp_cmd1DataSize=112ULL;
constexpr int64_t psp_cmd1DataOffset=116ULL;
constexpr int64_t psp_cmd1HeaderLen=144ULL;
std::array<std::array<uint8_t,16>,128> psp_keyvault=std::array<std::array<uint8_t,16>,128>{std::array<uint8_t,16>{cast<uint8_t>(44ULL),cast<uint8_t>(146ULL),cast<uint8_t>(229ULL),cast<uint8_t>(144ULL),cast<uint8_t>(43ULL),cast<uint8_t>(134ULL),cast<uint8_t>(193ULL),cast<uint8_t>(6ULL),cast<uint8_t>(183ULL),cast<uint8_t>(46ULL),cast<uint8_t>(234ULL),cast<uint8_t>(108ULL),cast<uint8_t>(212ULL),cast<uint8_t>(236ULL),cast<uint8_t>(114ULL),cast<uint8_t>(72ULL)},std::array<uint8_t,16>{cast<uint8_t>(5ULL),cast<uint8_t>(141ULL),cast<uint8_t>(200ULL),cast<uint8_t>(11ULL),cast<uint8_t>(51ULL),cast<uint8_t>(165ULL),cast<uint8_t>(191ULL),cast<uint8_t>(157ULL),cast<uint8_t>(86ULL),cast<uint8_t>(152ULL),cast<uint8_t>(250ULL),cast<uint8_t>(224ULL),cast<uint8_t>(211ULL),cast<uint8_t>(113ULL),cast<uint8_t>(94ULL),cast<uint8_t>(31ULL)},std::array<uint8_t,16>{cast<uint8_t>(184ULL),cast<uint8_t>(19ULL),cast<uint8_t>(195ULL),cast<uint8_t>(94ULL),cast<uint8_t>(198ULL),cast<uint8_t>(68ULL),cast<uint8_t>(65ULL),cast<uint8_t>(227ULL),cast<uint8_t>(220ULL),cast<uint8_t>(60ULL),cast<uint8_t>(22ULL),cast<uint8_t>(245ULL),cast<uint8_t>(180ULL),cast<uint8_t>(94ULL),cast<uint8_t>(100ULL),cast<uint8_t>(132ULL)},std::array<uint8_t,16>{cast<uint8_t>(152ULL),cast<uint8_t>(2ULL),cast<uint8_t>(196ULL),cast<uint8_t>(230ULL),cast<uint8_t>(236ULL),cast<uint8_t>(158ULL),cast<uint8_t>(158ULL),cast<uint8_t>(47ULL),cast<uint8_t>(252ULL),cast<uint8_t>(99ULL),cast<uint8_t>(76ULL),cast<uint8_t>(228ULL),cast<uint8_t>(47ULL),cast<uint8_t>(187ULL),cast<uint8_t>(70ULL),cast<uint8_t>(104ULL)},std::array<uint8_t,16>{cast<uint8_t>(153ULL),cast<uint8_t>(36ULL),cast<uint8_t>(76ULL),cast<uint8_t>(210ULL),cast<uint8_t>(88ULL),cast<uint8_t>(245ULL),cast<uint8_t>(27ULL),cast<uint8_t>(203ULL),cast<uint8_t>(176ULL),cast<uint8_t>(97ULL),cast<uint8_t>(156ULL),cast<uint8_t>(167ULL),cast<uint8_t>(56ULL),cast<uint8_t>(48ULL),cast<uint8_t>(7ULL),cast<uint8_t>(95ULL)},std::array<uint8_t,16>{cast<uint8_t>(2ULL),cast<uint8_t>(37ULL),cast<uint8_t>(215ULL),cast<uint8_t>(186ULL),cast<uint8_t>(99ULL),cast<uint8_t>(236ULL),cast<uint8_t>(185ULL),cast<uint8_t>(74ULL),cast<uint8_t>(157ULL),cast<uint8_t>(35ULL),cast<uint8_t>(118ULL),cast<uint8_t>(1ULL),cast<uint8_t>(179ULL),cast<uint8_t>(246ULL),cast<uint8_t>(172ULL),cast<uint8_t>(23ULL)},std::array<uint8_t,16>{cast<uint8_t>(96ULL),cast<uint8_t>(153ULL),cast<uint8_t>(242ULL),cast<uint8_t>(129ULL),cast<uint8_t>(112ULL),cast<uint8_t>(86ULL),cast<uint8_t>(14ULL),cast<uint8_t>(95ULL),cast<uint8_t>(116ULL),cast<uint8_t>(124ULL),cast<uint8_t>(181ULL),cast<uint8_t>(32ULL),cast<uint8_t>(192ULL),cast<uint8_t>(205ULL),cast<uint8_t>(194ULL),cast<uint8_t>(60ULL)},std::array<uint8_t,16>{cast<uint8_t>(118ULL),cast<uint8_t>(54ULL),cast<uint8_t>(139ULL),cast<uint8_t>(67ULL),cast<uint8_t>(143ULL),cast<uint8_t>(119ULL),cast<uint8_t>(216ULL),cast<uint8_t>(126ULL),cast<uint8_t>(254ULL),cast<uint8_t>(95ULL),cast<uint8_t>(182ULL),cast<uint8_t>(17ULL),cast<uint8_t>(89ULL),cast<uint8_t>(57ULL),cast<uint8_t>(136ULL),cast<uint8_t>(92ULL)},std::array<uint8_t,16>{cast<uint8_t>(20ULL),cast<uint8_t>(161ULL),cast<uint8_t>(21ULL),cast<uint8_t>(235ULL),cast<uint8_t>(67ULL),cast<uint8_t>(74ULL),cast<uint8_t>(27ULL),cast<uint8_t>(164ULL),cast<uint8_t>(144ULL),cast<uint8_t>(94ULL),cast<uint8_t>(3ULL),cast<uint8_t>(182ULL),cast<uint8_t>(23ULL),cast<uint8_t>(161ULL),cast<uint8_t>(92ULL),cast<uint8_t>(4ULL)},std::array<uint8_t,16>{cast<uint8_t>(230ULL),cast<uint8_t>(88ULL),cast<uint8_t>(3ULL),cast<uint8_t>(217ULL),cast<uint8_t>(167ULL),cast<uint8_t>(26ULL),cast<uint8_t>(168ULL),cast<uint8_t>(127ULL),cast<uint8_t>(5ULL),cast<uint8_t>(157ULL),cast<uint8_t>(34ULL),cast<uint8_t>(157ULL),cast<uint8_t>(175ULL),cast<uint8_t>(84ULL),cast<uint8_t>(83ULL),cast<uint8_t>(208ULL)},std::array<uint8_t,16>{cast<uint8_t>(186ULL),cast<uint8_t>(52ULL),cast<uint8_t>(128ULL),cast<uint8_t>(180ULL),cast<uint8_t>(40ULL),cast<uint8_t>(167ULL),cast<uint8_t>(202ULL),cast<uint8_t>(95ULL),cast<uint8_t>(33ULL),cast<uint8_t>(100ULL),cast<uint8_t>(18ULL),cast<uint8_t>(247ULL),cast<uint8_t>(15ULL),cast<uint8_t>(187ULL),cast<uint8_t>(115ULL),cast<uint8_t>(35ULL)},std::array<uint8_t,16>{cast<uint8_t>(114ULL),cast<uint8_t>(173ULL),cast<uint8_t>(53ULL),cast<uint8_t>(172ULL),cast<uint8_t>(154ULL),cast<uint8_t>(195ULL),cast<uint8_t>(19ULL),cast<uint8_t>(10ULL),cast<uint8_t>(119ULL),cast<uint8_t>(140ULL),cast<uint8_t>(177ULL),cast<uint8_t>(157ULL),cast<uint8_t>(136ULL),cast<uint8_t>(85ULL),cast<uint8_t>(11ULL),cast<uint8_t>(12ULL)},std::array<uint8_t,16>{cast<uint8_t>(132ULL),cast<uint8_t>(133ULL),cast<uint8_t>(200ULL),cast<uint8_t>(72ULL),cast<uint8_t>(117ULL),cast<uint8_t>(8ULL),cast<uint8_t>(67ULL),cast<uint8_t>(188ULL),cast<uint8_t>(155ULL),cast<uint8_t>(154ULL),cast<uint8_t>(236ULL),cast<uint8_t>(167ULL),cast<uint8_t>(156ULL),cast<uint8_t>(127ULL),cast<uint8_t>(96ULL),cast<uint8_t>(24ULL)},std::array<uint8_t,16>{cast<uint8_t>(181ULL),cast<uint8_t>(177ULL),cast<uint8_t>(110ULL),cast<uint8_t>(222ULL),cast<uint8_t>(35ULL),cast<uint8_t>(169ULL),cast<uint8_t>(123ULL),cast<uint8_t>(14ULL),cast<uint8_t>(161ULL),cast<uint8_t>(124ULL),cast<uint8_t>(219ULL),cast<uint8_t>(162ULL),cast<uint8_t>(220ULL),cast<uint8_t>(222ULL),cast<uint8_t>(196ULL),cast<uint8_t>(110ULL)},std::array<uint8_t,16>{cast<uint8_t>(200ULL),cast<uint8_t>(113ULL),cast<uint8_t>(253ULL),cast<uint8_t>(179ULL),cast<uint8_t>(188ULL),cast<uint8_t>(197ULL),cast<uint8_t>(210ULL),cast<uint8_t>(242ULL),cast<uint8_t>(226ULL),cast<uint8_t>(215ULL),cast<uint8_t>(114ULL),cast<uint8_t>(157ULL),cast<uint8_t>(223ULL),cast<uint8_t>(130ULL),cast<uint8_t>(104ULL),cast<uint8_t>(130ULL)},std::array<uint8_t,16>{cast<uint8_t>(10ULL),cast<uint8_t>(187ULL),cast<uint8_t>(51ULL),cast<uint8_t>(108ULL),cast<uint8_t>(150ULL),cast<uint8_t>(212ULL),cast<uint8_t>(205ULL),cast<uint8_t>(216ULL),cast<uint8_t>(203ULL),cast<uint8_t>(95ULL),cast<uint8_t>(75ULL),cast<uint8_t>(224ULL),cast<uint8_t>(186ULL),cast<uint8_t>(219ULL),cast<uint8_t>(158ULL),cast<uint8_t>(3ULL)},std::array<uint8_t,16>{cast<uint8_t>(50ULL),cast<uint8_t>(41ULL),cast<uint8_t>(91ULL),cast<uint8_t>(213ULL),cast<uint8_t>(234ULL),cast<uint8_t>(247ULL),cast<uint8_t>(163ULL),cast<uint8_t>(66ULL),cast<uint8_t>(22ULL),cast<uint8_t>(200ULL),cast<uint8_t>(142ULL),cast<uint8_t>(72ULL),cast<uint8_t>(255ULL),cast<uint8_t>(80ULL),cast<uint8_t>(211ULL),cast<uint8_t>(113ULL)},std::array<uint8_t,16>{cast<uint8_t>(70ULL),cast<uint8_t>(242ULL),cast<uint8_t>(94ULL),cast<uint8_t>(142ULL),cast<uint8_t>(77ULL),cast<uint8_t>(42ULL),cast<uint8_t>(165ULL),cast<uint8_t>(64ULL),cast<uint8_t>(115ULL),cast<uint8_t>(11ULL),cast<uint8_t>(196ULL),cast<uint8_t>(110ULL),cast<uint8_t>(71ULL),cast<uint8_t>(238ULL),cast<uint8_t>(111ULL),cast<uint8_t>(10ULL)},std::array<uint8_t,16>{cast<uint8_t>(93ULL),cast<uint8_t>(199ULL),cast<uint8_t>(17ULL),cast<uint8_t>(57ULL),cast<uint8_t>(208ULL),cast<uint8_t>(25ULL),cast<uint8_t>(56ULL),cast<uint8_t>(188ULL),cast<uint8_t>(2ULL),cast<uint8_t>(127ULL),cast<uint8_t>(221ULL),cast<uint8_t>(220ULL),cast<uint8_t>(176ULL),cast<uint8_t>(131ULL),cast<uint8_t>(125ULL),cast<uint8_t>(157ULL)},std::array<uint8_t,16>{cast<uint8_t>(81ULL),cast<uint8_t>(221ULL),cast<uint8_t>(101ULL),cast<uint8_t>(240ULL),cast<uint8_t>(113ULL),cast<uint8_t>(164ULL),cast<uint8_t>(229ULL),cast<uint8_t>(234ULL),cast<uint8_t>(106ULL),cast<uint8_t>(175ULL),cast<uint8_t>(18ULL),cast<uint8_t>(25ULL),cast<uint8_t>(65ULL),cast<uint8_t>(41ULL),cast<uint8_t>(184ULL),cast<uint8_t>(244ULL)},std::array<uint8_t,16>{cast<uint8_t>(3ULL),cast<uint8_t>(118ULL),cast<uint8_t>(60ULL),cast<uint8_t>(104ULL),cast<uint8_t>(101ULL),cast<uint8_t>(198ULL),cast<uint8_t>(155ULL),cast<uint8_t>(15ULL),cast<uint8_t>(254ULL),cast<uint8_t>(143ULL),cast<uint8_t>(216ULL),cast<uint8_t>(238ULL),cast<uint8_t>(164ULL),cast<uint8_t>(54ULL),cast<uint8_t>(22ULL),cast<uint8_t>(160ULL)},std::array<uint8_t,16>{cast<uint8_t>(125ULL),cast<uint8_t>(80ULL),cast<uint8_t>(184ULL),cast<uint8_t>(92ULL),cast<uint8_t>(175ULL),cast<uint8_t>(103ULL),cast<uint8_t>(105ULL),cast<uint8_t>(240ULL),cast<uint8_t>(229ULL),cast<uint8_t>(74ULL),cast<uint8_t>(168ULL),cast<uint8_t>(9ULL),cast<uint8_t>(139ULL),cast<uint8_t>(14ULL),cast<uint8_t>(190ULL),cast<uint8_t>(28ULL)},std::array<uint8_t,16>{cast<uint8_t>(114ULL),cast<uint8_t>(104ULL),cast<uint8_t>(75ULL),cast<uint8_t>(50ULL),cast<uint8_t>(172ULL),cast<uint8_t>(59ULL),cast<uint8_t>(51ULL),cast<uint8_t>(47ULL),cast<uint8_t>(42ULL),cast<uint8_t>(122ULL),cast<uint8_t>(252ULL),cast<uint8_t>(158ULL),cast<uint8_t>(20ULL),cast<uint8_t>(213ULL),cast<uint8_t>(111ULL),cast<uint8_t>(107ULL)},std::array<uint8_t,16>{cast<uint8_t>(32ULL),cast<uint8_t>(29ULL),cast<uint8_t>(49ULL),cast<uint8_t>(150ULL),cast<uint8_t>(74ULL),cast<uint8_t>(217ULL),cast<uint8_t>(159ULL),cast<uint8_t>(191ULL),cast<uint8_t>(50ULL),cast<uint8_t>(213ULL),cast<uint8_t>(214ULL),cast<uint8_t>(28ULL),cast<uint8_t>(73ULL),cast<uint8_t>(27ULL),cast<uint8_t>(217ULL),cast<uint8_t>(252ULL)},std::array<uint8_t,16>{cast<uint8_t>(248ULL),cast<uint8_t>(216ULL),cast<uint8_t>(68ULL),cast<uint8_t>(99ULL),cast<uint8_t>(214ULL),cast<uint8_t>(16ULL),cast<uint8_t>(209ULL),cast<uint8_t>(42ULL),cast<uint8_t>(68ULL),cast<uint8_t>(142ULL),cast<uint8_t>(150ULL),cast<uint8_t>(144ULL),cast<uint8_t>(166ULL),cast<uint8_t>(187ULL),cast<uint8_t>(11ULL),cast<uint8_t>(173ULL)},std::array<uint8_t,16>{cast<uint8_t>(92ULL),cast<uint8_t>(212ULL),cast<uint8_t>(5ULL),cast<uint8_t>(127ULL),cast<uint8_t>(161ULL),cast<uint8_t>(48ULL),cast<uint8_t>(96ULL),cast<uint8_t>(68ULL),cast<uint8_t>(10ULL),cast<uint8_t>(217ULL),cast<uint8_t>(182ULL),cast<uint8_t>(116ULL),cast<uint8_t>(95ULL),cast<uint8_t>(36ULL),cast<uint8_t>(79ULL),cast<uint8_t>(78ULL)},std::array<uint8_t,16>{cast<uint8_t>(244ULL),cast<uint8_t>(138ULL),cast<uint8_t>(214ULL),cast<uint8_t>(120ULL),cast<uint8_t>(89ULL),cast<uint8_t>(156ULL),cast<uint8_t>(34ULL),cast<uint8_t>(193ULL),cast<uint8_t>(212ULL),cast<uint8_t>(17ULL),cast<uint8_t>(147ULL),cast<uint8_t>(61ULL),cast<uint8_t>(248ULL),cast<uint8_t>(69ULL),cast<uint8_t>(184ULL),cast<uint8_t>(147ULL)},std::array<uint8_t,16>{cast<uint8_t>(202ULL),cast<uint8_t>(231ULL),cast<uint8_t>(210ULL),cast<uint8_t>(135ULL),cast<uint8_t>(162ULL),cast<uint8_t>(236ULL),cast<uint8_t>(193ULL),cast<uint8_t>(205ULL),cast<uint8_t>(148ULL),cast<uint8_t>(84ULL),cast<uint8_t>(43ULL),cast<uint8_t>(94ULL),cast<uint8_t>(29ULL),cast<uint8_t>(148ULL),cast<uint8_t>(136ULL),cast<uint8_t>(178ULL)},std::array<uint8_t,16>{cast<uint8_t>(222ULL),cast<uint8_t>(38ULL),cast<uint8_t>(211ULL),cast<uint8_t>(122ULL),cast<uint8_t>(57ULL),cast<uint8_t>(149ULL),cast<uint8_t>(108ULL),cast<uint8_t>(42ULL),cast<uint8_t>(216ULL),cast<uint8_t>(195ULL),cast<uint8_t>(166ULL),cast<uint8_t>(175ULL),cast<uint8_t>(33ULL),cast<uint8_t>(235ULL),cast<uint8_t>(179ULL),cast<uint8_t>(1ULL)},std::array<uint8_t,16>{cast<uint8_t>(124ULL),cast<uint8_t>(182ULL),cast<uint8_t>(139ULL),cast<uint8_t>(77ULL),cast<uint8_t>(163ULL),cast<uint8_t>(141ULL),cast<uint8_t>(29ULL),cast<uint8_t>(217ULL),cast<uint8_t>(50ULL),cast<uint8_t>(103ULL),cast<uint8_t>(156ULL),cast<uint8_t>(169ULL),cast<uint8_t>(159ULL),cast<uint8_t>(251ULL),cast<uint8_t>(40ULL),cast<uint8_t>(82ULL)},std::array<uint8_t,16>{cast<uint8_t>(160ULL),cast<uint8_t>(181ULL),cast<uint8_t>(86ULL),cast<uint8_t>(180ULL),cast<uint8_t>(105ULL),cast<uint8_t>(171ULL),cast<uint8_t>(54ULL),cast<uint8_t>(143ULL),cast<uint8_t>(54ULL),cast<uint8_t>(222ULL),cast<uint8_t>(201ULL),cast<uint8_t>(9ULL),cast<uint8_t>(46ULL),cast<uint8_t>(203ULL),cast<uint8_t>(65ULL),cast<uint8_t>(177ULL)},std::array<uint8_t,16>{cast<uint8_t>(147ULL),cast<uint8_t>(157ULL),cast<uint8_t>(225ULL),cast<uint8_t>(155ULL),cast<uint8_t>(114ULL),cast<uint8_t>(95ULL),cast<uint8_t>(238ULL),cast<uint8_t>(226ULL),cast<uint8_t>(69ULL),cast<uint8_t>(42ULL),cast<uint8_t>(188ULL),cast<uint8_t>(23ULL),cast<uint8_t>(6ULL),cast<uint8_t>(209ULL),cast<uint8_t>(71ULL),cast<uint8_t>(105ULL)},std::array<uint8_t,16>{cast<uint8_t>(164ULL),cast<uint8_t>(164ULL),cast<uint8_t>(230ULL),cast<uint8_t>(33ULL),cast<uint8_t>(56ULL),cast<uint8_t>(46ULL),cast<uint8_t>(241ULL),cast<uint8_t>(175ULL),cast<uint8_t>(123ULL),cast<uint8_t>(23ULL),cast<uint8_t>(122ULL),cast<uint8_t>(232ULL),cast<uint8_t>(66ULL),cast<uint8_t>(173ULL),cast<uint8_t>(0ULL),cast<uint8_t>(49ULL)},std::array<uint8_t,16>{cast<uint8_t>(195ULL),cast<uint8_t>(127ULL),cast<uint8_t>(19ULL),cast<uint8_t>(232ULL),cast<uint8_t>(207ULL),cast<uint8_t>(132ULL),cast<uint8_t>(219ULL),cast<uint8_t>(52ULL),cast<uint8_t>(116ULL),cast<uint8_t>(123ULL),cast<uint8_t>(195ULL),cast<uint8_t>(160ULL),cast<uint8_t>(241ULL),cast<uint8_t>(157ULL),cast<uint8_t>(58ULL),cast<uint8_t>(115ULL)},std::array<uint8_t,16>{cast<uint8_t>(43ULL),cast<uint8_t>(247ULL),cast<uint8_t>(131ULL),cast<uint8_t>(138ULL),cast<uint8_t>(216ULL),cast<uint8_t>(152ULL),cast<uint8_t>(233ULL),cast<uint8_t>(95ULL),cast<uint8_t>(165ULL),cast<uint8_t>(249ULL),cast<uint8_t>(1ULL),cast<uint8_t>(218ULL),cast<uint8_t>(97ULL),cast<uint8_t>(254ULL),cast<uint8_t>(53ULL),cast<uint8_t>(187ULL)},std::array<uint8_t,16>{cast<uint8_t>(199ULL),cast<uint8_t>(4ULL),cast<uint8_t>(98ULL),cast<uint8_t>(30ULL),cast<uint8_t>(113ULL),cast<uint8_t>(74ULL),cast<uint8_t>(102ULL),cast<uint8_t>(234ULL),cast<uint8_t>(98ULL),cast<uint8_t>(224ULL),cast<uint8_t>(75ULL),cast<uint8_t>(32ULL),cast<uint8_t>(61ULL),cast<uint8_t>(184ULL),cast<uint8_t>(194ULL),cast<uint8_t>(229ULL)},std::array<uint8_t,16>{cast<uint8_t>(201ULL),cast<uint8_t>(51ULL),cast<uint8_t>(133ULL),cast<uint8_t>(154ULL),cast<uint8_t>(171ULL),cast<uint8_t>(0ULL),cast<uint8_t>(205ULL),cast<uint8_t>(206ULL),cast<uint8_t>(77ULL),cast<uint8_t>(139ULL),cast<uint8_t>(142ULL),cast<uint8_t>(159ULL),cast<uint8_t>(61ULL),cast<uint8_t>(230ULL),cast<uint8_t>(192ULL),cast<uint8_t>(15ULL)},std::array<uint8_t,16>{cast<uint8_t>(24ULL),cast<uint8_t>(66ULL),cast<uint8_t>(86ULL),cast<uint8_t>(31ULL),cast<uint8_t>(43ULL),cast<uint8_t>(95ULL),cast<uint8_t>(52ULL),cast<uint8_t>(227ULL),cast<uint8_t>(81ULL),cast<uint8_t>(62ULL),cast<uint8_t>(183ULL),cast<uint8_t>(137ULL),cast<uint8_t>(119ULL),cast<uint8_t>(67ULL),cast<uint8_t>(26ULL),cast<uint8_t>(101ULL)},std::array<uint8_t,16>{cast<uint8_t>(220ULL),cast<uint8_t>(176ULL),cast<uint8_t>(160ULL),cast<uint8_t>(6ULL),cast<uint8_t>(90ULL),cast<uint8_t>(80ULL),cast<uint8_t>(161ULL),cast<uint8_t>(78ULL),cast<uint8_t>(89ULL),cast<uint8_t>(172ULL),cast<uint8_t>(151ULL),cast<uint8_t>(63ULL),cast<uint8_t>(23ULL),cast<uint8_t>(88ULL),cast<uint8_t>(163ULL),cast<uint8_t>(163ULL)},std::array<uint8_t,16>{cast<uint8_t>(196ULL),cast<uint8_t>(219ULL),cast<uint8_t>(174ULL),cast<uint8_t>(131ULL),cast<uint8_t>(226ULL),cast<uint8_t>(156ULL),cast<uint8_t>(242ULL),cast<uint8_t>(84ULL),cast<uint8_t>(163ULL),cast<uint8_t>(221ULL),cast<uint8_t>(55ULL),cast<uint8_t>(78ULL),cast<uint8_t>(128ULL),cast<uint8_t>(123ULL),cast<uint8_t>(244ULL),cast<uint8_t>(37ULL)},std::array<uint8_t,16>{cast<uint8_t>(191ULL),cast<uint8_t>(174ULL),cast<uint8_t>(235ULL),cast<uint8_t>(73ULL),cast<uint8_t>(130ULL),cast<uint8_t>(101ULL),cast<uint8_t>(197ULL),cast<uint8_t>(124ULL),cast<uint8_t>(100ULL),cast<uint8_t>(184ULL),cast<uint8_t>(193ULL),cast<uint8_t>(126ULL),cast<uint8_t>(25ULL),cast<uint8_t>(6ULL),cast<uint8_t>(68ULL),cast<uint8_t>(9ULL)},std::array<uint8_t,16>{cast<uint8_t>(121ULL),cast<uint8_t>(124ULL),cast<uint8_t>(236ULL),cast<uint8_t>(195ULL),cast<uint8_t>(179ULL),cast<uint8_t>(238ULL),cast<uint8_t>(10ULL),cast<uint8_t>(192ULL),cast<uint8_t>(59ULL),cast<uint8_t>(216ULL),cast<uint8_t>(230ULL),cast<uint8_t>(193ULL),cast<uint8_t>(224ULL),cast<uint8_t>(168ULL),cast<uint8_t>(177ULL),cast<uint8_t>(164ULL)},std::array<uint8_t,16>{cast<uint8_t>(117ULL),cast<uint8_t>(52ULL),cast<uint8_t>(254ULL),cast<uint8_t>(11ULL),cast<uint8_t>(214ULL),cast<uint8_t>(208ULL),cast<uint8_t>(194ULL),cast<uint8_t>(141ULL),cast<uint8_t>(104ULL),cast<uint8_t>(212ULL),cast<uint8_t>(224ULL),cast<uint8_t>(42ULL),cast<uint8_t>(231ULL),cast<uint8_t>(213ULL),cast<uint8_t>(209ULL),cast<uint8_t>(85ULL)},std::array<uint8_t,16>{cast<uint8_t>(250ULL),cast<uint8_t>(179ULL),cast<uint8_t>(83ULL),cast<uint8_t>(38ULL),cast<uint8_t>(151ULL),cast<uint8_t>(79ULL),cast<uint8_t>(78ULL),cast<uint8_t>(223ULL),cast<uint8_t>(228ULL),cast<uint8_t>(195ULL),cast<uint8_t>(168ULL),cast<uint8_t>(20ULL),cast<uint8_t>(195ULL),cast<uint8_t>(47ULL),cast<uint8_t>(15ULL),cast<uint8_t>(136ULL)},std::array<uint8_t,16>{cast<uint8_t>(236ULL),cast<uint8_t>(151ULL),cast<uint8_t>(179ULL),cast<uint8_t>(134ULL),cast<uint8_t>(180ULL),cast<uint8_t>(51ULL),cast<uint8_t>(198ULL),cast<uint8_t>(191ULL),cast<uint8_t>(78ULL),cast<uint8_t>(83ULL),cast<uint8_t>(157ULL),cast<uint8_t>(149ULL),cast<uint8_t>(235ULL),cast<uint8_t>(185ULL),cast<uint8_t>(121ULL),cast<uint8_t>(228ULL)},std::array<uint8_t,16>{cast<uint8_t>(179ULL),cast<uint8_t>(32ULL),cast<uint8_t>(162ULL),cast<uint8_t>(4ULL),cast<uint8_t>(207ULL),cast<uint8_t>(72ULL),cast<uint8_t>(6ULL),cast<uint8_t>(41ULL),cast<uint8_t>(181ULL),cast<uint8_t>(221ULL),cast<uint8_t>(142ULL),cast<uint8_t>(252ULL),cast<uint8_t>(152ULL),cast<uint8_t>(212ULL),cast<uint8_t>(23ULL),cast<uint8_t>(123ULL)},std::array<uint8_t,16>{cast<uint8_t>(93ULL),cast<uint8_t>(252ULL),cast<uint8_t>(13ULL),cast<uint8_t>(79ULL),cast<uint8_t>(44ULL),cast<uint8_t>(57ULL),cast<uint8_t>(218ULL),cast<uint8_t>(104ULL),cast<uint8_t>(74ULL),cast<uint8_t>(51ULL),cast<uint8_t>(116ULL),cast<uint8_t>(237ULL),cast<uint8_t>(73ULL),cast<uint8_t>(88ULL),cast<uint8_t>(167ULL),cast<uint8_t>(58ULL)},std::array<uint8_t,16>{cast<uint8_t>(215ULL),cast<uint8_t>(90ULL),cast<uint8_t>(84ULL),cast<uint8_t>(34ULL),cast<uint8_t>(206ULL),cast<uint8_t>(217ULL),cast<uint8_t>(163ULL),cast<uint8_t>(214ULL),cast<uint8_t>(43ULL),cast<uint8_t>(85ULL),cast<uint8_t>(125ULL),cast<uint8_t>(141ULL),cast<uint8_t>(232ULL),cast<uint8_t>(190ULL),cast<uint8_t>(199ULL),cast<uint8_t>(236ULL)},std::array<uint8_t,16>{cast<uint8_t>(107ULL),cast<uint8_t>(74ULL),cast<uint8_t>(238ULL),cast<uint8_t>(67ULL),cast<uint8_t>(69ULL),cast<uint8_t>(174ULL),cast<uint8_t>(112ULL),cast<uint8_t>(7ULL),cast<uint8_t>(207ULL),cast<uint8_t>(141ULL),cast<uint8_t>(207ULL),cast<uint8_t>(78ULL),cast<uint8_t>(74ULL),cast<uint8_t>(233ULL),cast<uint8_t>(60ULL),cast<uint8_t>(250ULL)},std::array<uint8_t,16>{cast<uint8_t>(43ULL),cast<uint8_t>(82ULL),cast<uint8_t>(47ULL),cast<uint8_t>(102ULL),cast<uint8_t>(76ULL),cast<uint8_t>(45ULL),cast<uint8_t>(17ULL),cast<uint8_t>(76ULL),cast<uint8_t>(254ULL),cast<uint8_t>(97ULL),cast<uint8_t>(49ULL),cast<uint8_t>(140ULL),cast<uint8_t>(86ULL),cast<uint8_t>(120ULL),cast<uint8_t>(78ULL),cast<uint8_t>(166ULL)},std::array<uint8_t,16>{cast<uint8_t>(58ULL),cast<uint8_t>(163ULL),cast<uint8_t>(78ULL),cast<uint8_t>(68ULL),cast<uint8_t>(198ULL),cast<uint8_t>(111ULL),cast<uint8_t>(175ULL),cast<uint8_t>(123ULL),cast<uint8_t>(250ULL),cast<uint8_t>(229ULL),cast<uint8_t>(83ULL),cast<uint8_t>(39ULL),cast<uint8_t>(239ULL),cast<uint8_t>(207ULL),cast<uint8_t>(204ULL),cast<uint8_t>(36ULL)},std::array<uint8_t,16>{cast<uint8_t>(43ULL),cast<uint8_t>(92ULL),cast<uint8_t>(120ULL),cast<uint8_t>(191ULL),cast<uint8_t>(195ULL),cast<uint8_t>(142ULL),cast<uint8_t>(73ULL),cast<uint8_t>(157ULL),cast<uint8_t>(65ULL),cast<uint8_t>(195ULL),cast<uint8_t>(60ULL),cast<uint8_t>(92ULL),cast<uint8_t>(123ULL),cast<uint8_t>(39ULL),cast<uint8_t>(150ULL),cast<uint8_t>(206ULL)},std::array<uint8_t,16>{cast<uint8_t>(243ULL),cast<uint8_t>(126ULL),cast<uint8_t>(234ULL),cast<uint8_t>(210ULL),cast<uint8_t>(192ULL),cast<uint8_t>(200ULL),cast<uint8_t>(35ULL),cast<uint8_t>(29ULL),cast<uint8_t>(169ULL),cast<uint8_t>(155ULL),cast<uint8_t>(250ULL),cast<uint8_t>(73ULL),cast<uint8_t>(93ULL),cast<uint8_t>(183ULL),cast<uint8_t>(8ULL),cast<uint8_t>(27ULL)},std::array<uint8_t,16>{cast<uint8_t>(112ULL),cast<uint8_t>(141ULL),cast<uint8_t>(78ULL),cast<uint8_t>(111ULL),cast<uint8_t>(209ULL),cast<uint8_t>(246ULL),cast<uint8_t>(111ULL),cast<uint8_t>(29ULL),cast<uint8_t>(30ULL),cast<uint8_t>(31ULL),cast<uint8_t>(203ULL),cast<uint8_t>(2ULL),cast<uint8_t>(249ULL),cast<uint8_t>(179ULL),cast<uint8_t>(153ULL),cast<uint8_t>(38ULL)},std::array<uint8_t,16>{cast<uint8_t>(15ULL),cast<uint8_t>(103ULL),cast<uint8_t>(22ULL),cast<uint8_t>(225ULL),cast<uint8_t>(128ULL),cast<uint8_t>(105ULL),cast<uint8_t>(156ULL),cast<uint8_t>(81ULL),cast<uint8_t>(252ULL),cast<uint8_t>(199ULL),cast<uint8_t>(173ULL),cast<uint8_t>(110ULL),cast<uint8_t>(79ULL),cast<uint8_t>(184ULL),cast<uint8_t>(70ULL),cast<uint8_t>(201ULL)},std::array<uint8_t,16>{cast<uint8_t>(86ULL),cast<uint8_t>(10ULL),cast<uint8_t>(73ULL),cast<uint8_t>(74ULL),cast<uint8_t>(132ULL),cast<uint8_t>(76ULL),cast<uint8_t>(142ULL),cast<uint8_t>(217ULL),cast<uint8_t>(130ULL),cast<uint8_t>(238ULL),cast<uint8_t>(11ULL),cast<uint8_t>(109ULL),cast<uint8_t>(197ULL),cast<uint8_t>(125ULL),cast<uint8_t>(32ULL),cast<uint8_t>(141ULL)},std::array<uint8_t,16>{cast<uint8_t>(18ULL),cast<uint8_t>(70ULL),cast<uint8_t>(141ULL),cast<uint8_t>(126ULL),cast<uint8_t>(28ULL),cast<uint8_t>(66ULL),cast<uint8_t>(32ULL),cast<uint8_t>(155ULL),cast<uint8_t>(186ULL),cast<uint8_t>(84ULL),cast<uint8_t>(38ULL),cast<uint8_t>(131ULL),cast<uint8_t>(94ULL),cast<uint8_t>(176ULL),cast<uint8_t>(51ULL),cast<uint8_t>(3ULL)},std::array<uint8_t,16>{cast<uint8_t>(196ULL),cast<uint8_t>(59ULL),cast<uint8_t>(182ULL),cast<uint8_t>(214ULL),cast<uint8_t>(83ULL),cast<uint8_t>(238ULL),cast<uint8_t>(103ULL),cast<uint8_t>(73ULL),cast<uint8_t>(62ULL),cast<uint8_t>(169ULL),cast<uint8_t>(95ULL),cast<uint8_t>(188ULL),cast<uint8_t>(12ULL),cast<uint8_t>(237ULL),cast<uint8_t>(111ULL),cast<uint8_t>(138ULL)},std::array<uint8_t,16>{cast<uint8_t>(44ULL),cast<uint8_t>(195ULL),cast<uint8_t>(207ULL),cast<uint8_t>(140ULL),cast<uint8_t>(40ULL),cast<uint8_t>(120ULL),cast<uint8_t>(165ULL),cast<uint8_t>(166ULL),cast<uint8_t>(99ULL),cast<uint8_t>(226ULL),cast<uint8_t>(175ULL),cast<uint8_t>(45ULL),cast<uint8_t>(113ULL),cast<uint8_t>(94ULL),cast<uint8_t>(134ULL),cast<uint8_t>(186ULL)},std::array<uint8_t,16>{cast<uint8_t>(131ULL),cast<uint8_t>(61ULL),cast<uint8_t>(167ULL),cast<uint8_t>(12ULL),cast<uint8_t>(237ULL),cast<uint8_t>(106ULL),cast<uint8_t>(32ULL),cast<uint8_t>(18ULL),cast<uint8_t>(209ULL),cast<uint8_t>(150ULL),cast<uint8_t>(230ULL),cast<uint8_t>(254ULL),cast<uint8_t>(92ULL),cast<uint8_t>(77ULL),cast<uint8_t>(55ULL),cast<uint8_t>(197ULL)},std::array<uint8_t,16>{cast<uint8_t>(199ULL),cast<uint8_t>(67ULL),cast<uint8_t>(208ULL),cast<uint8_t>(103ULL),cast<uint8_t>(66ULL),cast<uint8_t>(238ULL),cast<uint8_t>(144ULL),cast<uint8_t>(184ULL),cast<uint8_t>(202ULL),cast<uint8_t>(117ULL),cast<uint8_t>(80ULL),cast<uint8_t>(53ULL),cast<uint8_t>(32ULL),cast<uint8_t>(173ULL),cast<uint8_t>(188ULL),cast<uint8_t>(206ULL)},std::array<uint8_t,16>{cast<uint8_t>(138ULL),cast<uint8_t>(227ULL),cast<uint8_t>(102ULL),cast<uint8_t>(63ULL),cast<uint8_t>(141ULL),cast<uint8_t>(158ULL),cast<uint8_t>(130ULL),cast<uint8_t>(161ULL),cast<uint8_t>(237ULL),cast<uint8_t>(230ULL),cast<uint8_t>(140ULL),cast<uint8_t>(156ULL),cast<uint8_t>(232ULL),cast<uint8_t>(37ULL),cast<uint8_t>(109ULL),cast<uint8_t>(170ULL)},std::array<uint8_t,16>{cast<uint8_t>(127ULL),cast<uint8_t>(201ULL),cast<uint8_t>(111ULL),cast<uint8_t>(11ULL),cast<uint8_t>(177ULL),cast<uint8_t>(72ULL),cast<uint8_t>(92ULL),cast<uint8_t>(165ULL),cast<uint8_t>(93ULL),cast<uint8_t>(211ULL),cast<uint8_t>(100ULL),cast<uint8_t>(183ULL),cast<uint8_t>(122ULL),cast<uint8_t>(245ULL),cast<uint8_t>(228ULL),cast<uint8_t>(234ULL)},std::array<uint8_t,16>{cast<uint8_t>(145ULL),cast<uint8_t>(183ULL),cast<uint8_t>(101ULL),cast<uint8_t>(120ULL),cast<uint8_t>(139ULL),cast<uint8_t>(203ULL),cast<uint8_t>(139ULL),cast<uint8_t>(212ULL),cast<uint8_t>(2ULL),cast<uint8_t>(237ULL),cast<uint8_t>(85ULL),cast<uint8_t>(58ULL),cast<uint8_t>(102ULL),cast<uint8_t>(98ULL),cast<uint8_t>(208ULL),cast<uint8_t>(173ULL)},std::array<uint8_t,16>{cast<uint8_t>(40ULL),cast<uint8_t>(36ULL),cast<uint8_t>(249ULL),cast<uint8_t>(16ULL),cast<uint8_t>(27ULL),cast<uint8_t>(141ULL),cast<uint8_t>(15ULL),cast<uint8_t>(123ULL),cast<uint8_t>(110ULL),cast<uint8_t>(178ULL),cast<uint8_t>(99ULL),cast<uint8_t>(181ULL),cast<uint8_t>(181ULL),cast<uint8_t>(91ULL),cast<uint8_t>(46ULL),cast<uint8_t>(187ULL)},std::array<uint8_t,16>{cast<uint8_t>(48ULL),cast<uint8_t>(226ULL),cast<uint8_t>(87ULL),cast<uint8_t>(93ULL),cast<uint8_t>(224ULL),cast<uint8_t>(162ULL),cast<uint8_t>(73ULL),cast<uint8_t>(206ULL),cast<uint8_t>(232ULL),cast<uint8_t>(207ULL),cast<uint8_t>(43ULL),cast<uint8_t>(94ULL),cast<uint8_t>(77ULL),cast<uint8_t>(159ULL),cast<uint8_t>(82ULL),cast<uint8_t>(199ULL)},std::array<uint8_t,16>{cast<uint8_t>(94ULL),cast<uint8_t>(229ULL),cast<uint8_t>(4ULL),cast<uint8_t>(57ULL),cast<uint8_t>(98ULL),cast<uint8_t>(50ULL),cast<uint8_t>(2ULL),cast<uint8_t>(250ULL),cast<uint8_t>(133ULL),cast<uint8_t>(57ULL),cast<uint8_t>(63ULL),cast<uint8_t>(114ULL),cast<uint8_t>(187ULL),cast<uint8_t>(119ULL),cast<uint8_t>(253ULL),cast<uint8_t>(26ULL)},std::array<uint8_t,16>{cast<uint8_t>(248ULL),cast<uint8_t>(129ULL),cast<uint8_t>(116ULL),cast<uint8_t>(177ULL),cast<uint8_t>(189ULL),cast<uint8_t>(233ULL),cast<uint8_t>(191ULL),cast<uint8_t>(221ULL),cast<uint8_t>(69ULL),cast<uint8_t>(226ULL),cast<uint8_t>(245ULL),cast<uint8_t>(85ULL),cast<uint8_t>(137ULL),cast<uint8_t>(207ULL),cast<uint8_t>(70ULL),cast<uint8_t>(171ULL)},std::array<uint8_t,16>{cast<uint8_t>(125ULL),cast<uint8_t>(244ULL),cast<uint8_t>(146ULL),cast<uint8_t>(101ULL),cast<uint8_t>(227ULL),cast<uint8_t>(250ULL),cast<uint8_t>(214ULL),cast<uint8_t>(120ULL),cast<uint8_t>(214ULL),cast<uint8_t>(254ULL),cast<uint8_t>(120ULL),cast<uint8_t>(173ULL),cast<uint8_t>(187ULL),cast<uint8_t>(61ULL),cast<uint8_t>(251ULL),cast<uint8_t>(99ULL)},std::array<uint8_t,16>{cast<uint8_t>(116ULL),cast<uint8_t>(127ULL),cast<uint8_t>(214ULL),cast<uint8_t>(45ULL),cast<uint8_t>(199ULL),cast<uint8_t>(161ULL),cast<uint8_t>(202ULL),cast<uint8_t>(150ULL),cast<uint8_t>(226ULL),cast<uint8_t>(122ULL),cast<uint8_t>(206ULL),cast<uint8_t>(255ULL),cast<uint8_t>(170ULL),cast<uint8_t>(114ULL),cast<uint8_t>(63ULL),cast<uint8_t>(247ULL)},std::array<uint8_t,16>{cast<uint8_t>(30ULL),cast<uint8_t>(88ULL),cast<uint8_t>(235ULL),cast<uint8_t>(208ULL),cast<uint8_t>(101ULL),cast<uint8_t>(187ULL),cast<uint8_t>(241ULL),cast<uint8_t>(104ULL),cast<uint8_t>(197ULL),cast<uint8_t>(189ULL),cast<uint8_t>(247ULL),cast<uint8_t>(70ULL),cast<uint8_t>(186ULL),cast<uint8_t>(123ULL),cast<uint8_t>(225ULL),cast<uint8_t>(0ULL)},std::array<uint8_t,16>{cast<uint8_t>(36ULL),cast<uint8_t>(52ULL),cast<uint8_t>(125ULL),cast<uint8_t>(175ULL),cast<uint8_t>(94ULL),cast<uint8_t>(75ULL),cast<uint8_t>(53ULL),cast<uint8_t>(114ULL),cast<uint8_t>(122ULL),cast<uint8_t>(82ULL),cast<uint8_t>(39ULL),cast<uint8_t>(107ULL),cast<uint8_t>(160ULL),cast<uint8_t>(84ULL),cast<uint8_t>(116ULL),cast<uint8_t>(219ULL)},std::array<uint8_t,16>{cast<uint8_t>(9ULL),cast<uint8_t>(177ULL),cast<uint8_t>(199ULL),cast<uint8_t>(5ULL),cast<uint8_t>(195ULL),cast<uint8_t>(95ULL),cast<uint8_t>(83ULL),cast<uint8_t>(102ULL),cast<uint8_t>(119ULL),cast<uint8_t>(192ULL),cast<uint8_t>(235ULL),cast<uint8_t>(54ULL),cast<uint8_t>(119ULL),cast<uint8_t>(223ULL),cast<uint8_t>(131ULL),cast<uint8_t>(7ULL)},std::array<uint8_t,16>{cast<uint8_t>(204ULL),cast<uint8_t>(190ULL),cast<uint8_t>(97ULL),cast<uint8_t>(92ULL),cast<uint8_t>(5ULL),cast<uint8_t>(162ULL),cast<uint8_t>(0ULL),cast<uint8_t>(51ULL),cast<uint8_t>(55ULL),cast<uint8_t>(142ULL),cast<uint8_t>(89ULL),cast<uint8_t>(100ULL),cast<uint8_t>(167ULL),cast<uint8_t>(221ULL),cast<uint8_t>(112ULL),cast<uint8_t>(61ULL)},std::array<uint8_t,16>{cast<uint8_t>(13ULL),cast<uint8_t>(71ULL),cast<uint8_t>(80ULL),cast<uint8_t>(187ULL),cast<uint8_t>(252ULL),cast<uint8_t>(176ULL),cast<uint8_t>(2ULL),cast<uint8_t>(129ULL),cast<uint8_t>(48ULL),cast<uint8_t>(225ULL),cast<uint8_t>(132ULL),cast<uint8_t>(222ULL),cast<uint8_t>(168ULL),cast<uint8_t>(212ULL),cast<uint8_t>(132ULL),cast<uint8_t>(19ULL)},std::array<uint8_t,16>{cast<uint8_t>(12ULL),cast<uint8_t>(253ULL),cast<uint8_t>(103ULL),cast<uint8_t>(154ULL),cast<uint8_t>(249ULL),cast<uint8_t>(180ULL),cast<uint8_t>(114ULL),cast<uint8_t>(79ULL),cast<uint8_t>(215ULL),cast<uint8_t>(141ULL),cast<uint8_t>(214ULL),cast<uint8_t>(233ULL),cast<uint8_t>(150ULL),cast<uint8_t>(66ULL),cast<uint8_t>(40ULL),cast<uint8_t>(139ULL)},std::array<uint8_t,16>{cast<uint8_t>(122ULL),cast<uint8_t>(211ULL),cast<uint8_t>(26ULL),cast<uint8_t>(139ULL),cast<uint8_t>(75ULL),cast<uint8_t>(239ULL),cast<uint8_t>(194ULL),cast<uint8_t>(194ULL),cast<uint8_t>(179ULL),cast<uint8_t>(153ULL),cast<uint8_t>(1ULL),cast<uint8_t>(169ULL),cast<uint8_t>(254ULL),cast<uint8_t>(118ULL),cast<uint8_t>(185ULL),cast<uint8_t>(135ULL)},std::array<uint8_t,16>{cast<uint8_t>(190ULL),cast<uint8_t>(120ULL),cast<uint8_t>(120ULL),cast<uint8_t>(23ULL),cast<uint8_t>(199ULL),cast<uint8_t>(241ULL),cast<uint8_t>(111ULL),cast<uint8_t>(26ULL),cast<uint8_t>(224ULL),cast<uint8_t>(239ULL),cast<uint8_t>(59ULL),cast<uint8_t>(222ULL),cast<uint8_t>(76ULL),cast<uint8_t>(194ULL),cast<uint8_t>(215ULL),cast<uint8_t>(134ULL)},std::array<uint8_t,16>{cast<uint8_t>(124ULL),cast<uint8_t>(216ULL),cast<uint8_t>(184ULL),cast<uint8_t>(145ULL),cast<uint8_t>(145ULL),cast<uint8_t>(10ULL),cast<uint8_t>(67ULL),cast<uint8_t>(20ULL),cast<uint8_t>(208ULL),cast<uint8_t>(83ULL),cast<uint8_t>(61ULL),cast<uint8_t>(216ULL),cast<uint8_t>(76ULL),cast<uint8_t>(69ULL),cast<uint8_t>(190ULL),cast<uint8_t>(22ULL)},std::array<uint8_t,16>{cast<uint8_t>(50ULL),cast<uint8_t>(114ULL),cast<uint8_t>(44ULL),cast<uint8_t>(136ULL),cast<uint8_t>(7ULL),cast<uint8_t>(207ULL),cast<uint8_t>(53ULL),cast<uint8_t>(125ULL),cast<uint8_t>(74ULL),cast<uint8_t>(47ULL),cast<uint8_t>(81ULL),cast<uint8_t>(25ULL),cast<uint8_t>(68ULL),cast<uint8_t>(174ULL),cast<uint8_t>(104ULL),cast<uint8_t>(218ULL)},std::array<uint8_t,16>{cast<uint8_t>(126ULL),cast<uint8_t>(107ULL),cast<uint8_t>(191ULL),cast<uint8_t>(246ULL),cast<uint8_t>(246ULL),cast<uint8_t>(135ULL),cast<uint8_t>(184ULL),cast<uint8_t>(152ULL),cast<uint8_t>(238ULL),cast<uint8_t>(181ULL),cast<uint8_t>(27ULL),cast<uint8_t>(50ULL),cast<uint8_t>(22ULL),cast<uint8_t>(228ULL),cast<uint8_t>(110ULL),cast<uint8_t>(93ULL)},std::array<uint8_t,16>{cast<uint8_t>(8ULL),cast<uint8_t>(234ULL),cast<uint8_t>(90ULL),cast<uint8_t>(131ULL),cast<uint8_t>(73ULL),cast<uint8_t>(181ULL),cast<uint8_t>(157ULL),cast<uint8_t>(181ULL),cast<uint8_t>(62ULL),cast<uint8_t>(7ULL),cast<uint8_t>(121ULL),cast<uint8_t>(177ULL),cast<uint8_t>(154ULL),cast<uint8_t>(89ULL),cast<uint8_t>(163ULL),cast<uint8_t>(84ULL)},std::array<uint8_t,16>{cast<uint8_t>(243ULL),cast<uint8_t>(18ULL),cast<uint8_t>(129ULL),cast<uint8_t>(191ULL),cast<uint8_t>(230ULL),cast<uint8_t>(159ULL),cast<uint8_t>(81ULL),cast<uint8_t>(209ULL),cast<uint8_t>(100ULL),cast<uint8_t>(8ULL),cast<uint8_t>(37ULL),cast<uint8_t>(33ULL),cast<uint8_t>(255ULL),cast<uint8_t>(187ULL),cast<uint8_t>(34ULL),cast<uint8_t>(97ULL)},std::array<uint8_t,16>{cast<uint8_t>(175ULL),cast<uint8_t>(254ULL),cast<uint8_t>(142ULL),cast<uint8_t>(177ULL),cast<uint8_t>(61ULL),cast<uint8_t>(209ULL),cast<uint8_t>(126ULL),cast<uint8_t>(216ULL),cast<uint8_t>(10ULL),cast<uint8_t>(97ULL),cast<uint8_t>(36ULL),cast<uint8_t>(28ULL),cast<uint8_t>(149ULL),cast<uint8_t>(146ULL),cast<uint8_t>(86ULL),cast<uint8_t>(182ULL)},std::array<uint8_t,16>{cast<uint8_t>(146ULL),cast<uint8_t>(205ULL),cast<uint8_t>(180ULL),cast<uint8_t>(194ULL),cast<uint8_t>(91ULL),cast<uint8_t>(242ULL),cast<uint8_t>(53ULL),cast<uint8_t>(90ULL),cast<uint8_t>(35ULL),cast<uint8_t>(9ULL),cast<uint8_t>(232ULL),cast<uint8_t>(25ULL),cast<uint8_t>(201ULL),cast<uint8_t>(20ULL),cast<uint8_t>(66ULL),cast<uint8_t>(53ULL)},std::array<uint8_t,16>{cast<uint8_t>(225ULL),cast<uint8_t>(198ULL),cast<uint8_t>(91ULL),cast<uint8_t>(34ULL),cast<uint8_t>(107ULL),cast<uint8_t>(225ULL),cast<uint8_t>(218ULL),cast<uint8_t>(2ULL),cast<uint8_t>(186ULL),cast<uint8_t>(24ULL),cast<uint8_t>(250ULL),cast<uint8_t>(33ULL),cast<uint8_t>(52ULL),cast<uint8_t>(158ULL),cast<uint8_t>(249ULL),cast<uint8_t>(109ULL)},std::array<uint8_t,16>{cast<uint8_t>(20ULL),cast<uint8_t>(236ULL),cast<uint8_t>(118ULL),cast<uint8_t>(206ULL),cast<uint8_t>(151ULL),cast<uint8_t>(243ULL),cast<uint8_t>(138ULL),cast<uint8_t>(10ULL),cast<uint8_t>(52ULL),cast<uint8_t>(80ULL),cast<uint8_t>(108ULL),cast<uint8_t>(83ULL),cast<uint8_t>(154ULL),cast<uint8_t>(92ULL),cast<uint8_t>(154ULL),cast<uint8_t>(180ULL)},std::array<uint8_t,16>{cast<uint8_t>(28ULL),cast<uint8_t>(155ULL),cast<uint8_t>(196ULL),cast<uint8_t>(144ULL),cast<uint8_t>(227ULL),cast<uint8_t>(6ULL),cast<uint8_t>(100ULL),cast<uint8_t>(129ULL),cast<uint8_t>(250ULL),cast<uint8_t>(89ULL),cast<uint8_t>(253ULL),cast<uint8_t>(182ULL),cast<uint8_t>(0ULL),cast<uint8_t>(187ULL),cast<uint8_t>(40ULL),cast<uint8_t>(112ULL)},std::array<uint8_t,16>{cast<uint8_t>(67ULL),cast<uint8_t>(165ULL),cast<uint8_t>(202ULL),cast<uint8_t>(204ULL),cast<uint8_t>(13ULL),cast<uint8_t>(108ULL),cast<uint8_t>(45ULL),cast<uint8_t>(63ULL),cast<uint8_t>(43ULL),cast<uint8_t>(217ULL),cast<uint8_t>(137ULL),cast<uint8_t>(103ULL),cast<uint8_t>(107ULL),cast<uint8_t>(63ULL),cast<uint8_t>(127ULL),cast<uint8_t>(87ULL)},std::array<uint8_t,16>{cast<uint8_t>(0ULL),cast<uint8_t>(239ULL),cast<uint8_t>(253ULL),cast<uint8_t>(24ULL),cast<uint8_t>(8ULL),cast<uint8_t>(164ULL),cast<uint8_t>(5ULL),cast<uint8_t>(137ULL),cast<uint8_t>(60ULL),cast<uint8_t>(56ULL),cast<uint8_t>(251ULL),cast<uint8_t>(37ULL),cast<uint8_t>(114ULL),cast<uint8_t>(112ULL),cast<uint8_t>(97ULL),cast<uint8_t>(6ULL)},std::array<uint8_t,16>{cast<uint8_t>(238ULL),cast<uint8_t>(175ULL),cast<uint8_t>(73ULL),cast<uint8_t>(224ULL),cast<uint8_t>(9ULL),cast<uint8_t>(135ULL),cast<uint8_t>(155ULL),cast<uint8_t>(239ULL),cast<uint8_t>(170ULL),cast<uint8_t>(214ULL),cast<uint8_t>(50ULL),cast<uint8_t>(106ULL),cast<uint8_t>(50ULL),cast<uint8_t>(19ULL),cast<uint8_t>(196ULL),cast<uint8_t>(41ULL)},std::array<uint8_t,16>{cast<uint8_t>(141ULL),cast<uint8_t>(38ULL),cast<uint8_t>(185ULL),cast<uint8_t>(15ULL),cast<uint8_t>(67ULL),cast<uint8_t>(29ULL),cast<uint8_t>(187ULL),cast<uint8_t>(8ULL),cast<uint8_t>(219ULL),cast<uint8_t>(29ULL),cast<uint8_t>(218ULL),cast<uint8_t>(197ULL),cast<uint8_t>(181ULL),cast<uint8_t>(44ULL),cast<uint8_t>(146ULL),cast<uint8_t>(237ULL)},std::array<uint8_t,16>{cast<uint8_t>(87ULL),cast<uint8_t>(124ULL),cast<uint8_t>(48ULL),cast<uint8_t>(96ULL),cast<uint8_t>(174ULL),cast<uint8_t>(110ULL),cast<uint8_t>(190ULL),cast<uint8_t>(174ULL),cast<uint8_t>(58ULL),cast<uint8_t>(171ULL),cast<uint8_t>(24ULL),cast<uint8_t>(25ULL),cast<uint8_t>(197ULL),cast<uint8_t>(113ULL),cast<uint8_t>(104ULL),cast<uint8_t>(11ULL)},std::array<uint8_t,16>{cast<uint8_t>(17ULL),cast<uint8_t>(90ULL),cast<uint8_t>(93ULL),cast<uint8_t>(32ULL),cast<uint8_t>(213ULL),cast<uint8_t>(58ULL),cast<uint8_t>(141ULL),cast<uint8_t>(211ULL),cast<uint8_t>(156ULL),cast<uint8_t>(197ULL),cast<uint8_t>(175ULL),cast<uint8_t>(65ULL),cast<uint8_t>(15ULL),cast<uint8_t>(15ULL),cast<uint8_t>(24ULL),cast<uint8_t>(111ULL)},std::array<uint8_t,16>{cast<uint8_t>(13ULL),cast<uint8_t>(77ULL),cast<uint8_t>(81ULL),cast<uint8_t>(171ULL),cast<uint8_t>(35ULL),cast<uint8_t>(121ULL),cast<uint8_t>(191ULL),cast<uint8_t>(128ULL),cast<uint8_t>(58ULL),cast<uint8_t>(191ULL),cast<uint8_t>(185ULL),cast<uint8_t>(14ULL),cast<uint8_t>(117ULL),cast<uint8_t>(252ULL),cast<uint8_t>(20ULL),cast<uint8_t>(191ULL)},std::array<uint8_t,16>{cast<uint8_t>(153ULL),cast<uint8_t>(147ULL),cast<uint8_t>(218ULL),cast<uint8_t>(62ULL),cast<uint8_t>(125ULL),cast<uint8_t>(46ULL),cast<uint8_t>(91ULL),cast<uint8_t>(21ULL),cast<uint8_t>(242ULL),cast<uint8_t>(82ULL),cast<uint8_t>(164ULL),cast<uint8_t>(230ULL),cast<uint8_t>(107ULL),cast<uint8_t>(184ULL),cast<uint8_t>(90ULL),cast<uint8_t>(152ULL)},std::array<uint8_t,16>{cast<uint8_t>(244ULL),cast<uint8_t>(40ULL),cast<uint8_t>(48ULL),cast<uint8_t>(165ULL),cast<uint8_t>(251ULL),cast<uint8_t>(13ULL),cast<uint8_t>(141ULL),cast<uint8_t>(118ULL),cast<uint8_t>(14ULL),cast<uint8_t>(166ULL),cast<uint8_t>(113ULL),cast<uint8_t>(194ULL),cast<uint8_t>(43ULL),cast<uint8_t>(222ULL),cast<uint8_t>(102ULL),cast<uint8_t>(157ULL)},std::array<uint8_t,16>{cast<uint8_t>(251ULL),cast<uint8_t>(95ULL),cast<uint8_t>(235ULL),cast<uint8_t>(127ULL),cast<uint8_t>(199ULL),cast<uint8_t>(220ULL),cast<uint8_t>(221ULL),cast<uint8_t>(105ULL),cast<uint8_t>(55ULL),cast<uint8_t>(1ULL),cast<uint8_t>(151ULL),cast<uint8_t>(155ULL),cast<uint8_t>(41ULL),cast<uint8_t>(3ULL),cast<uint8_t>(92ULL),cast<uint8_t>(71ULL)},std::array<uint8_t,16>{cast<uint8_t>(2ULL),cast<uint8_t>(50ULL),cast<uint8_t>(106ULL),cast<uint8_t>(231ULL),cast<uint8_t>(211ULL),cast<uint8_t>(150ULL),cast<uint8_t>(206ULL),cast<uint8_t>(127ULL),cast<uint8_t>(28ULL),cast<uint8_t>(65ULL),cast<uint8_t>(157ULL),cast<uint8_t>(214ULL),cast<uint8_t>(82ULL),cast<uint8_t>(7ULL),cast<uint8_t>(237ULL),cast<uint8_t>(9ULL)},std::array<uint8_t,16>{cast<uint8_t>(156ULL),cast<uint8_t>(155ULL),cast<uint8_t>(19ULL),cast<uint8_t>(114ULL),cast<uint8_t>(248ULL),cast<uint8_t>(198ULL),cast<uint8_t>(64ULL),cast<uint8_t>(207ULL),cast<uint8_t>(28ULL),cast<uint8_t>(98ULL),cast<uint8_t>(245ULL),cast<uint8_t>(213ULL),cast<uint8_t>(146ULL),cast<uint8_t>(221ULL),cast<uint8_t>(181ULL),cast<uint8_t>(130ULL)},std::array<uint8_t,16>{cast<uint8_t>(3ULL),cast<uint8_t>(179ULL),cast<uint8_t>(2ULL),cast<uint8_t>(232ULL),cast<uint8_t>(95ULL),cast<uint8_t>(243ULL),cast<uint8_t>(129ULL),cast<uint8_t>(177ULL),cast<uint8_t>(59ULL),cast<uint8_t>(141ULL),cast<uint8_t>(170ULL),cast<uint8_t>(42ULL),cast<uint8_t>(144ULL),cast<uint8_t>(255ULL),cast<uint8_t>(94ULL),cast<uint8_t>(97ULL)},std::array<uint8_t,16>{cast<uint8_t>(188ULL),cast<uint8_t>(215ULL),cast<uint8_t>(249ULL),cast<uint8_t>(211ULL),cast<uint8_t>(47ULL),cast<uint8_t>(172ULL),cast<uint8_t>(248ULL),cast<uint8_t>(71ULL),cast<uint8_t>(192ULL),cast<uint8_t>(251ULL),cast<uint8_t>(77ULL),cast<uint8_t>(47ULL),cast<uint8_t>(48ULL),cast<uint8_t>(154ULL),cast<uint8_t>(189ULL),cast<uint8_t>(166ULL)},std::array<uint8_t,16>{cast<uint8_t>(245ULL),cast<uint8_t>(85ULL),cast<uint8_t>(150ULL),cast<uint8_t>(233ULL),cast<uint8_t>(127ULL),cast<uint8_t>(175ULL),cast<uint8_t>(134ULL),cast<uint8_t>(127ULL),cast<uint8_t>(172ULL),cast<uint8_t>(179ULL),cast<uint8_t>(58ULL),cast<uint8_t>(230ULL),cast<uint8_t>(156ULL),cast<uint8_t>(139ULL),cast<uint8_t>(111ULL),cast<uint8_t>(147ULL)},std::array<uint8_t,16>{cast<uint8_t>(238ULL),cast<uint8_t>(41ULL),cast<uint8_t>(112ULL),cast<uint8_t>(147ULL),cast<uint8_t>(249ULL),cast<uint8_t>(78ULL),cast<uint8_t>(68ULL),cast<uint8_t>(89ULL),cast<uint8_t>(68ULL),cast<uint8_t>(23ULL),cast<uint8_t>(31ULL),cast<uint8_t>(142ULL),cast<uint8_t>(134ULL),cast<uint8_t>(225ULL),cast<uint8_t>(112ULL),cast<uint8_t>(252ULL)},std::array<uint8_t,16>{cast<uint8_t>(228ULL),cast<uint8_t>(52ULL),cast<uint8_t>(82ULL),cast<uint8_t>(12ULL),cast<uint8_t>(240ULL),cast<uint8_t>(136ULL),cast<uint8_t>(207ULL),cast<uint8_t>(200ULL),cast<uint8_t>(205ULL),cast<uint8_t>(120ULL),cast<uint8_t>(27ULL),cast<uint8_t>(108ULL),cast<uint8_t>(207ULL),cast<uint8_t>(140ULL),cast<uint8_t>(72ULL),cast<uint8_t>(196ULL)},std::array<uint8_t,16>{cast<uint8_t>(193ULL),cast<uint8_t>(191ULL),cast<uint8_t>(102ULL),cast<uint8_t>(129ULL),cast<uint8_t>(142ULL),cast<uint8_t>(249ULL),cast<uint8_t>(83ULL),cast<uint8_t>(242ULL),cast<uint8_t>(225ULL),cast<uint8_t>(38ULL),cast<uint8_t>(107ULL),cast<uint8_t>(111ULL),cast<uint8_t>(85ULL),cast<uint8_t>(12ULL),cast<uint8_t>(201ULL),cast<uint8_t>(205ULL)},std::array<uint8_t,16>{cast<uint8_t>(86ULL),cast<uint8_t>(15ULL),cast<uint8_t>(255ULL),cast<uint8_t>(143ULL),cast<uint8_t>(60ULL),cast<uint8_t>(150ULL),cast<uint8_t>(73ULL),cast<uint8_t>(20ULL),cast<uint8_t>(69ULL),cast<uint8_t>(22ULL),cast<uint8_t>(241ULL),cast<uint8_t>(188ULL),cast<uint8_t>(191ULL),cast<uint8_t>(206ULL),cast<uint8_t>(163ULL),cast<uint8_t>(12ULL)},std::array<uint8_t,16>{cast<uint8_t>(36ULL),cast<uint8_t>(8ULL),cast<uint8_t>(220ULL),cast<uint8_t>(117ULL),cast<uint8_t>(55ULL),cast<uint8_t>(96ULL),cast<uint8_t>(162ULL),cast<uint8_t>(159ULL),cast<uint8_t>(5ULL),cast<uint8_t>(84ULL),cast<uint8_t>(181ULL),cast<uint8_t>(242ULL),cast<uint8_t>(67ULL),cast<uint8_t>(133ULL),cast<uint8_t>(115ULL),cast<uint8_t>(153ULL)},std::array<uint8_t,16>{cast<uint8_t>(221ULL),cast<uint8_t>(213ULL),cast<uint8_t>(181ULL),cast<uint8_t>(106ULL),cast<uint8_t>(89ULL),cast<uint8_t>(197ULL),cast<uint8_t>(90ULL),cast<uint8_t>(232ULL),cast<uint8_t>(59ULL),cast<uint8_t>(150ULL),cast<uint8_t>(103ULL),cast<uint8_t>(199ULL),cast<uint8_t>(92ULL),cast<uint8_t>(42ULL),cast<uint8_t>(226ULL),cast<uint8_t>(220ULL)},std::array<uint8_t,16>{cast<uint8_t>(170ULL),cast<uint8_t>(104ULL),cast<uint8_t>(103ULL),cast<uint8_t>(114ULL),cast<uint8_t>(224ULL),cast<uint8_t>(45ULL),cast<uint8_t>(68ULL),cast<uint8_t>(213ULL),cast<uint8_t>(205ULL),cast<uint8_t>(187ULL),cast<uint8_t>(101ULL),cast<uint8_t>(4ULL),cast<uint8_t>(188ULL),cast<uint8_t>(213ULL),cast<uint8_t>(191ULL),cast<uint8_t>(78ULL)},std::array<uint8_t,16>{cast<uint8_t>(31ULL),cast<uint8_t>(23ULL),cast<uint8_t>(240ULL),cast<uint8_t>(20ULL),cast<uint8_t>(231ULL),cast<uint8_t>(119ULL),cast<uint8_t>(162ULL),cast<uint8_t>(254ULL),cast<uint8_t>(75ULL),cast<uint8_t>(19ULL),cast<uint8_t>(107ULL),cast<uint8_t>(86ULL),cast<uint8_t>(205ULL),cast<uint8_t>(126ULL),cast<uint8_t>(247ULL),cast<uint8_t>(233ULL)},std::array<uint8_t,16>{cast<uint8_t>(201ULL),cast<uint8_t>(53ULL),cast<uint8_t>(72ULL),cast<uint8_t>(207ULL),cast<uint8_t>(85ULL),cast<uint8_t>(141ULL),cast<uint8_t>(117ULL),cast<uint8_t>(3ULL),cast<uint8_t>(137ULL),cast<uint8_t>(107ULL),cast<uint8_t>(46ULL),cast<uint8_t>(235ULL),cast<uint8_t>(97ULL),cast<uint8_t>(140ULL),cast<uint8_t>(169ULL),cast<uint8_t>(2ULL)},std::array<uint8_t,16>{cast<uint8_t>(222ULL),cast<uint8_t>(52ULL),cast<uint8_t>(197ULL),cast<uint8_t>(65ULL),cast<uint8_t>(231ULL),cast<uint8_t>(202ULL),cast<uint8_t>(134ULL),cast<uint8_t>(232ULL),cast<uint8_t>(190ULL),cast<uint8_t>(167ULL),cast<uint8_t>(195ULL),cast<uint8_t>(28ULL),cast<uint8_t>(236ULL),cast<uint8_t>(228ULL),cast<uint8_t>(54ULL),cast<uint8_t>(15ULL)},std::array<uint8_t,16>{cast<uint8_t>(221ULL),cast<uint8_t>(229ULL),cast<uint8_t>(255ULL),cast<uint8_t>(85ULL),cast<uint8_t>(27ULL),cast<uint8_t>(116ULL),cast<uint8_t>(246ULL),cast<uint8_t>(244ULL),cast<uint8_t>(224ULL),cast<uint8_t>(22ULL),cast<uint8_t>(215ULL),cast<uint8_t>(171ULL),cast<uint8_t>(34ULL),cast<uint8_t>(49ULL),cast<uint8_t>(27ULL),cast<uint8_t>(106ULL)},std::array<uint8_t,16>{cast<uint8_t>(176ULL),cast<uint8_t>(233ULL),cast<uint8_t>(53ULL),cast<uint8_t>(33ULL),cast<uint8_t>(51ULL),cast<uint8_t>(63ULL),cast<uint8_t>(215ULL),cast<uint8_t>(186ULL),cast<uint8_t>(180ULL),cast<uint8_t>(118ULL),cast<uint8_t>(44ULL),cast<uint8_t>(203ULL),cast<uint8_t>(77ULL),cast<uint8_t>(128ULL),cast<uint8_t>(8ULL),cast<uint8_t>(216ULL)},std::array<uint8_t,16>{cast<uint8_t>(56ULL),cast<uint8_t>(20ULL),cast<uint8_t>(105ULL),cast<uint8_t>(196ULL),cast<uint8_t>(195ULL),cast<uint8_t>(249ULL),cast<uint8_t>(27ULL),cast<uint8_t>(150ULL),cast<uint8_t>(51ULL),cast<uint8_t>(99ULL),cast<uint8_t>(142ULL),cast<uint8_t>(77ULL),cast<uint8_t>(95ULL),cast<uint8_t>(61ULL),cast<uint8_t>(240ULL),cast<uint8_t>(41ULL)},std::array<uint8_t,16>{cast<uint8_t>(250ULL),cast<uint8_t>(72ULL),cast<uint8_t>(106ULL),cast<uint8_t>(217ULL),cast<uint8_t>(142ULL),cast<uint8_t>(103ULL),cast<uint8_t>(22ULL),cast<uint8_t>(239ULL),cast<uint8_t>(106ULL),cast<uint8_t>(176ULL),cast<uint8_t>(135ULL),cast<uint8_t>(245ULL),cast<uint8_t>(137ULL),cast<uint8_t>(69ULL),cast<uint8_t>(127ULL),cast<uint8_t>(42ULL)},std::array<uint8_t,16>{cast<uint8_t>(50ULL),cast<uint8_t>(26ULL),cast<uint8_t>(9ULL),cast<uint8_t>(18ULL),cast<uint8_t>(80ULL),cast<uint8_t>(20ULL),cast<uint8_t>(138ULL),cast<uint8_t>(62ULL),cast<uint8_t>(150ULL),cast<uint8_t>(61ULL),cast<uint8_t>(234ULL),cast<uint8_t>(2ULL),cast<uint8_t>(89ULL),cast<uint8_t>(50ULL),cast<uint8_t>(225ULL),cast<uint8_t>(143ULL)},std::array<uint8_t,16>{cast<uint8_t>(75ULL),cast<uint8_t>(0ULL),cast<uint8_t>(190ULL),cast<uint8_t>(41ULL),cast<uint8_t>(188ULL),cast<uint8_t>(176ULL),cast<uint8_t>(40ULL),cast<uint8_t>(100ULL),cast<uint8_t>(206ULL),cast<uint8_t>(253ULL),cast<uint8_t>(67ULL),cast<uint8_t>(169ULL),cast<uint8_t>(111ULL),cast<uint8_t>(217ULL),cast<uint8_t>(92ULL),cast<uint8_t>(237ULL)},std::array<uint8_t,16>{cast<uint8_t>(87ULL),cast<uint8_t>(125ULL),cast<uint8_t>(196ULL),cast<uint8_t>(255ULL),cast<uint8_t>(2ULL),cast<uint8_t>(68ULL),cast<uint8_t>(226ULL),cast<uint8_t>(128ULL),cast<uint8_t>(145ULL),cast<uint8_t>(244ULL),cast<uint8_t>(202ULL),cast<uint8_t>(10ULL),cast<uint8_t>(117ULL),cast<uint8_t>(105ULL),cast<uint8_t>(253ULL),cast<uint8_t>(168ULL)},std::array<uint8_t,16>{cast<uint8_t>(131ULL),cast<uint8_t>(83ULL),cast<uint8_t>(54ULL),cast<uint8_t>(198ULL),cast<uint8_t>(24ULL),cast<uint8_t>(3ULL),cast<uint8_t>(228ULL),cast<uint8_t>(62ULL),cast<uint8_t>(78ULL),cast<uint8_t>(179ULL),cast<uint8_t>(15ULL),cast<uint8_t>(107ULL),cast<uint8_t>(110ULL),cast<uint8_t>(121ULL),cast<uint8_t>(155ULL),cast<uint8_t>(122ULL)},std::array<uint8_t,16>{cast<uint8_t>(92ULL),cast<uint8_t>(146ULL),cast<uint8_t>(101ULL),cast<uint8_t>(253ULL),cast<uint8_t>(123ULL),cast<uint8_t>(89ULL),cast<uint8_t>(106ULL),cast<uint8_t>(163ULL),cast<uint8_t>(122ULL),cast<uint8_t>(47ULL),cast<uint8_t>(80ULL),cast<uint8_t>(157ULL),cast<uint8_t>(133ULL),cast<uint8_t>(233ULL),cast<uint8_t>(39ULL),cast<uint8_t>(248ULL)},std::array<uint8_t,16>{cast<uint8_t>(154ULL),cast<uint8_t>(57ULL),cast<uint8_t>(251ULL),cast<uint8_t>(137ULL),cast<uint8_t>(223ULL),cast<uint8_t>(85ULL),cast<uint8_t>(178ULL),cast<uint8_t>(96ULL),cast<uint8_t>(20ULL),cast<uint8_t>(36ULL),cast<uint8_t>(206ULL),cast<uint8_t>(166ULL),cast<uint8_t>(217ULL),cast<uint8_t>(101ULL),cast<uint8_t>(10ULL),cast<uint8_t>(157ULL)},std::array<uint8_t,16>{cast<uint8_t>(139ULL),cast<uint8_t>(117ULL),cast<uint8_t>(190ULL),cast<uint8_t>(145ULL),cast<uint8_t>(168ULL),cast<uint8_t>(199ULL),cast<uint8_t>(90ULL),cast<uint8_t>(210ULL),cast<uint8_t>(215ULL),cast<uint8_t>(165ULL),cast<uint8_t>(148ULL),cast<uint8_t>(160ULL),cast<uint8_t>(28ULL),cast<uint8_t>(187ULL),cast<uint8_t>(149ULL),cast<uint8_t>(145ULL)},std::array<uint8_t,16>{cast<uint8_t>(149ULL),cast<uint8_t>(194ULL),cast<uint8_t>(27ULL),cast<uint8_t>(141ULL),cast<uint8_t>(5ULL),cast<uint8_t>(172ULL),cast<uint8_t>(245ULL),cast<uint8_t>(236ULL),cast<uint8_t>(90ULL),cast<uint8_t>(238ULL),cast<uint8_t>(119ULL),cast<uint8_t>(129ULL),cast<uint8_t>(35ULL),cast<uint8_t>(149ULL),cast<uint8_t>(196ULL),cast<uint8_t>(215ULL)},std::array<uint8_t,16>{cast<uint8_t>(185ULL),cast<uint8_t>(164ULL),cast<uint8_t>(97ULL),cast<uint8_t>(100ULL),cast<uint8_t>(54ULL),cast<uint8_t>(51ULL),cast<uint8_t>(250ULL),cast<uint8_t>(93ULL),cast<uint8_t>(148ULL),cast<uint8_t>(136ULL),cast<uint8_t>(226ULL),cast<uint8_t>(211ULL),cast<uint8_t>(40ULL),cast<uint8_t>(30ULL),cast<uint8_t>(1ULL),cast<uint8_t>(162ULL)},std::array<uint8_t,16>{cast<uint8_t>(184ULL),cast<uint8_t>(176ULL),cast<uint8_t>(132ULL),cast<uint8_t>(251ULL),cast<uint8_t>(159ULL),cast<uint8_t>(76ULL),cast<uint8_t>(250ULL),cast<uint8_t>(247ULL),cast<uint8_t>(48ULL),cast<uint8_t>(254ULL),cast<uint8_t>(115ULL),cast<uint8_t>(37ULL),cast<uint8_t>(162ULL),cast<uint8_t>(171ULL),cast<uint8_t>(137ULL),cast<uint8_t>(125ULL)},std::array<uint8_t,16>{cast<uint8_t>(95ULL),cast<uint8_t>(140ULL),cast<uint8_t>(23ULL),cast<uint8_t>(159ULL),cast<uint8_t>(193ULL),cast<uint8_t>(178ULL),cast<uint8_t>(29ULL),cast<uint8_t>(241ULL),cast<uint8_t>(246ULL),cast<uint8_t>(54ULL),cast<uint8_t>(122ULL),cast<uint8_t>(156ULL),cast<uint8_t>(247ULL),cast<uint8_t>(211ULL),cast<uint8_t>(212ULL),cast<uint8_t>(124ULL)}};
std::array<uint8_t,16> psp_kirk1Key=std::array<uint8_t,16>{cast<uint8_t>(152ULL),cast<uint8_t>(201ULL),cast<uint8_t>(64ULL),cast<uint8_t>(151ULL),cast<uint8_t>(92ULL),cast<uint8_t>(29ULL),cast<uint8_t>(16ULL),cast<uint8_t>(232ULL),cast<uint8_t>(127ULL),cast<uint8_t>(230ULL),cast<uint8_t>(14ULL),cast<uint8_t>(163ULL),cast<uint8_t>(253ULL),cast<uint8_t>(3ULL),cast<uint8_t>(168ULL),cast<uint8_t>(186ULL)};
std::array<uint8_t,144> psp_keyEBOOT2xx=std::array<uint8_t,144>{cast<uint8_t>(250ULL),cast<uint8_t>(54ULL),cast<uint8_t>(142ULL),cast<uint8_t>(218ULL),cast<uint8_t>(71ULL),cast<uint8_t>(116ULL),cast<uint8_t>(217ULL),cast<uint8_t>(93ULL),cast<uint8_t>(116ULL),cast<uint8_t>(152ULL),cast<uint8_t>(193ULL),cast<uint8_t>(118ULL),cast<uint8_t>(175ULL),cast<uint8_t>(126ULL),cast<uint8_t>(229ULL),cast<uint8_t>(151ULL),cast<uint8_t>(189ULL),cast<uint8_t>(9ULL),cast<uint8_t>(171ULL),cast<uint8_t>(28ULL),cast<uint8_t>(198ULL),cast<uint8_t>(186ULL),cast<uint8_t>(53ULL),cast<uint8_t>(152ULL),cast<uint8_t>(129ULL),cast<uint8_t>(146ULL),cast<uint8_t>(211ULL),cast<uint8_t>(3ULL),cast<uint8_t>(207ULL),cast<uint8_t>(5ULL),cast<uint8_t>(178ULL),cast<uint8_t>(3ULL),cast<uint8_t>(52ULL),cast<uint8_t>(231ULL),cast<uint8_t>(130ULL),cast<uint8_t>(40ULL),cast<uint8_t>(99ULL),cast<uint8_t>(246ULL),cast<uint8_t>(20ULL),cast<uint8_t>(231ULL),cast<uint8_t>(117ULL),cast<uint8_t>(39ULL),cast<uint8_t>(110ULL),cast<uint8_t>(185ULL),cast<uint8_t>(199ULL),cast<uint8_t>(175ULL),cast<uint8_t>(138ULL),cast<uint8_t>(189ULL),cast<uint8_t>(41ULL),cast<uint8_t>(236ULL),cast<uint8_t>(211ULL),cast<uint8_t>(29ULL),cast<uint8_t>(108ULL),cast<uint8_t>(161ULL),cast<uint8_t>(164ULL),cast<uint8_t>(236ULL),cast<uint8_t>(135ULL),cast<uint8_t>(236ULL),cast<uint8_t>(105ULL),cast<uint8_t>(95ULL),cast<uint8_t>(146ULL),cast<uint8_t>(30ULL),cast<uint8_t>(152ULL),cast<uint8_t>(133ULL),cast<uint8_t>(33ULL),cast<uint8_t>(174ULL),cast<uint8_t>(252ULL),cast<uint8_t>(124ULL),cast<uint8_t>(22ULL),cast<uint8_t>(221ULL),cast<uint8_t>(233ULL),cast<uint8_t>(186ULL),cast<uint8_t>(4ULL),cast<uint8_t>(120ULL),cast<uint8_t>(169ULL),cast<uint8_t>(230ULL),cast<uint8_t>(252ULL),cast<uint8_t>(2ULL),cast<uint8_t>(238ULL),cast<uint8_t>(46ULL),cast<uint8_t>(61ULL),cast<uint8_t>(138ULL),cast<uint8_t>(223ULL),cast<uint8_t>(97ULL),cast<uint8_t>(100ULL),cast<uint8_t>(5ULL),cast<uint8_t>(49ULL),cast<uint8_t>(221ULL),cast<uint8_t>(73ULL),cast<uint8_t>(225ULL),cast<uint8_t>(151ULL),cast<uint8_t>(150ULL),cast<uint8_t>(59ULL),cast<uint8_t>(63ULL),cast<uint8_t>(69ULL),cast<uint8_t>(194ULL),cast<uint8_t>(86ULL),cast<uint8_t>(132ULL),cast<uint8_t>(29ULL),cast<uint8_t>(249ULL),cast<uint8_t>(200ULL),cast<uint8_t>(107ULL),cast<uint8_t>(218ULL),cast<uint8_t>(57ULL),cast<uint8_t>(245ULL),cast<uint8_t>(254ULL),cast<uint8_t>(229ULL),cast<uint8_t>(179ULL),cast<uint8_t>(163ULL),cast<uint8_t>(147ULL),cast<uint8_t>(197ULL),cast<uint8_t>(137ULL),cast<uint8_t>(188ULL),cast<uint8_t>(138ULL),cast<uint8_t>(92ULL),cast<uint8_t>(251ULL),cast<uint8_t>(18ULL),cast<uint8_t>(114ULL),cast<uint8_t>(11ULL),cast<uint8_t>(108ULL),cast<uint8_t>(203ULL),cast<uint8_t>(211ULL),cast<uint8_t>(13ULL),cast<uint8_t>(225ULL),cast<uint8_t>(168ULL),cast<uint8_t>(178ULL),cast<uint8_t>(208ULL),cast<uint8_t>(152ULL),cast<uint8_t>(71ULL),cast<uint8_t>(24ULL),cast<uint8_t>(214ULL),cast<uint8_t>(95ULL),cast<uint8_t>(87ULL),cast<uint8_t>(35ULL),cast<uint8_t>(220ULL),cast<uint8_t>(240ULL),cast<uint8_t>(106ULL),cast<uint8_t>(22ULL),cast<uint8_t>(1ULL),cast<uint8_t>(119ULL),cast<uint8_t>(104ULL),cast<uint8_t>(59ULL),cast<uint8_t>(92ULL),cast<uint8_t>(15ULL)};
std::array<uint8_t,144> psp_keyEBOOT1xx=std::array<uint8_t,144>{cast<uint8_t>(239ULL),cast<uint8_t>(105ULL),cast<uint8_t>(203ULL),cast<uint8_t>(24ULL),cast<uint8_t>(18ULL),cast<uint8_t>(137ULL),cast<uint8_t>(142ULL),cast<uint8_t>(21ULL),cast<uint8_t>(187ULL),cast<uint8_t>(14ULL),cast<uint8_t>(249ULL),cast<uint8_t>(222ULL),cast<uint8_t>(35ULL),cast<uint8_t>(251ULL),cast<uint8_t>(176ULL),cast<uint8_t>(76ULL),cast<uint8_t>(24ULL),cast<uint8_t>(238ULL),cast<uint8_t>(135ULL),cast<uint8_t>(54ULL),cast<uint8_t>(110ULL),cast<uint8_t>(74ULL),cast<uint8_t>(141ULL),cast<uint8_t>(134ULL),cast<uint8_t>(86ULL),cast<uint8_t>(199ULL),cast<uint8_t>(181ULL),cast<uint8_t>(25ULL),cast<uint8_t>(29ULL),cast<uint8_t>(85ULL),cast<uint8_t>(22ULL),cast<uint8_t>(238ULL),cast<uint8_t>(108ULL),cast<uint8_t>(45ULL),cast<uint8_t>(203ULL),cast<uint8_t>(231ULL),cast<uint8_t>(96ULL),cast<uint8_t>(198ULL),cast<uint8_t>(71ULL),cast<uint8_t>(151ULL),cast<uint8_t>(63ULL),cast<uint8_t>(20ULL),cast<uint8_t>(149ULL),cast<uint8_t>(206ULL),cast<uint8_t>(119ULL),cast<uint8_t>(244ULL),cast<uint8_t>(86ULL),cast<uint8_t>(41ULL),cast<uint8_t>(222ULL),cast<uint8_t>(74ULL),cast<uint8_t>(130ULL),cast<uint8_t>(3ULL),cast<uint8_t>(241ULL),cast<uint8_t>(157ULL),cast<uint8_t>(12ULL),cast<uint8_t>(33ULL),cast<uint8_t>(36ULL),cast<uint8_t>(235ULL),cast<uint8_t>(41ULL),cast<uint8_t>(80ULL),cast<uint8_t>(159ULL),cast<uint8_t>(230ULL),cast<uint8_t>(223ULL),cast<uint8_t>(129ULL),cast<uint8_t>(0ULL),cast<uint8_t>(155ULL),cast<uint8_t>(200ULL),cast<uint8_t>(57ULL),cast<uint8_t>(145ULL),cast<uint8_t>(139ULL),cast<uint8_t>(12ULL),cast<uint8_t>(176ULL),cast<uint8_t>(194ULL),cast<uint8_t>(249ULL),cast<uint8_t>(45ULL),cast<uint8_t>(239ULL),cast<uint8_t>(252ULL),cast<uint8_t>(147ULL),cast<uint8_t>(58ULL),cast<uint8_t>(225ULL),cast<uint8_t>(168ULL),cast<uint8_t>(164ULL),cast<uint8_t>(148ULL),cast<uint8_t>(139ULL),cast<uint8_t>(157ULL),cast<uint8_t>(208ULL),cast<uint8_t>(29ULL),cast<uint8_t>(73ULL),cast<uint8_t>(13ULL),cast<uint8_t>(64ULL),cast<uint8_t>(106ULL),cast<uint8_t>(104ULL),cast<uint8_t>(228ULL),cast<uint8_t>(199ULL),cast<uint8_t>(212ULL),cast<uint8_t>(206ULL),cast<uint8_t>(201ULL),cast<uint8_t>(183ULL),cast<uint8_t>(200ULL),cast<uint8_t>(150ULL),cast<uint8_t>(40ULL),cast<uint8_t>(220ULL),cast<uint8_t>(170ULL),cast<uint8_t>(30ULL),cast<uint8_t>(132ULL),cast<uint8_t>(11ULL),cast<uint8_t>(23ULL),cast<uint8_t>(164ULL),cast<uint8_t>(220ULL),cast<uint8_t>(93ULL),cast<uint8_t>(93ULL),cast<uint8_t>(80ULL),cast<uint8_t>(207ULL),cast<uint8_t>(195ULL),cast<uint8_t>(166ULL),cast<uint8_t>(93ULL),cast<uint8_t>(45ULL),cast<uint8_t>(250ULL),cast<uint8_t>(93ULL),cast<uint8_t>(14ULL),cast<uint8_t>(181ULL),cast<uint8_t>(25ULL),cast<uint8_t>(121ULL),cast<uint8_t>(110ULL),cast<uint8_t>(199ULL),cast<uint8_t>(41ULL),cast<uint8_t>(94ULL),cast<uint8_t>(206ULL),cast<uint8_t>(148ULL),cast<uint8_t>(219ULL),cast<uint8_t>(172ULL),cast<uint8_t>(170ULL),cast<uint8_t>(221ULL),cast<uint8_t>(12ULL),cast<uint8_t>(247ULL),cast<uint8_t>(69ULL),cast<uint8_t>(37ULL),cast<uint8_t>(55ULL),cast<uint8_t>(167ULL),cast<uint8_t>(98ULL),cast<uint8_t>(61ULL),cast<uint8_t>(86ULL),cast<uint8_t>(230ULL),cast<uint8_t>(204ULL)};
constexpr int64_t psp_ramBase=134217728ULL;
constexpr int64_t psp_ramSize=33554432ULL;
constexpr int64_t psp_vramBase=67108864ULL;
constexpr int64_t psp_vramSize=2097152ULL;
constexpr int64_t psp_scratchBase=65536ULL;
constexpr int64_t psp_scratchSize=16384ULL;
constexpr int64_t psp_userBase=142622720ULL;
constexpr int64_t psp_stackTop=167706624ULL;
constexpr int64_t psp_stepsPerVBlank=1000000ULL;
constexpr int64_t psp_errMpegNoData=2153873409ULL;
constexpr int64_t psp_errMpegInvalid=2153841150ULL;
constexpr int64_t psp_mpegPacketSize=2048ULL;
constexpr int64_t psp_mpegPtsPerFrame=3003ULL;
constexpr int64_t psp_errAtracAllDecoded=2153971714ULL;
constexpr int64_t psp_errAtracBadID=2153971716ULL;
constexpr int64_t psp_atracMaxSamples=2048ULL;
constexpr int64_t psp_bucketList=0ULL;
constexpr int64_t psp_bucketVertex=1ULL;
constexpr int64_t psp_bucketRaster=2ULL;
constexpr int64_t psp_bucketXfer=3ULL;
constexpr int64_t psp_bucketSyscall=4ULL;
constexpr int64_t psp_numBuckets=5ULL;
std::array<std::string,5> psp_bucketNames=[](){std::array<std::string,5> v{};v[0]=std::string("list capture",12);v[1]=std::string("vertex + transform",18);v[2]=std::string("rasterise",9);v[3]=std::string("block transfer",14);v[4]=std::string("syscall HLE",11);return v;}();
std::array<int64_t,4> psp_geBuckets=std::array<int64_t,4>{cast<int64_t>(0ULL),cast<int64_t>(1ULL),cast<int64_t>(2ULL),cast<int64_t>(3ULL)};
Map<uint32_t,psp_tagInfo> psp_tagTable=Map<uint32_t,psp_tagInfo>{{cast<uint32_t>(3234535036ULL),psp_tagInfo{sub(psp_keyEBOOT2xx,0,len(psp_keyEBOOT2xx)),cast<int64_t>(93ULL)}},{cast<uint32_t>(134217728ULL),psp_tagInfo{sub(psp_keyEBOOT1xx,0,len(psp_keyEBOOT1xx)),cast<int64_t>(75ULL)}}};
constexpr int64_t psp_threadExitAddr=251658240ULL;
constexpr psp_threadState psp_thDormant=0ULL;
constexpr psp_threadState psp_thReady=1ULL;
constexpr psp_threadState psp_thRunning=2ULL;
constexpr psp_threadState psp_thWaiting=3ULL;

#include "adapters.h"
// tools/cpu/allegrex/allegrex.go:53:1
std::string allegrex_Flow_String(allegrex_Flow f){
{
{
switch(f){
case cast<allegrex_Flow>(0ULL):{
return std::string("seq",3);
break;}
case cast<allegrex_Flow>(1ULL):{
return std::string("branch",6);
break;}
case cast<allegrex_Flow>(2ULL):{
return std::string("jump",4);
break;}
case cast<allegrex_Flow>(3ULL):{
return std::string("call",4);
break;}
case cast<allegrex_Flow>(4ULL):{
return std::string("return",6);
break;}
case cast<allegrex_Flow>(5ULL):{
return std::string("indjump",7);
break;}
case cast<allegrex_Flow>(6ULL):{
return std::string("indcall",7);
break;}
case cast<allegrex_Flow>(7ULL):{
return std::string("stop",4);
break;}
}}
return std::string("?",1);
}
}
// tools/cpu/allegrex/allegrex.go:88:1
std::string allegrex_Inst_String(allegrex_Inst in){
{
return go_fmt_Sprintf(std::string("$%08X: %s",9),in.Addr,in.Text);
}
}
// tools/cpu/allegrex/cpu.go:117:1
void allegrex_CPU_Reset(allegrex_CPU* c){
{
c->R = std::array<uint32_t,32>{};
c->out = std::array<uint32_t,32>{};
auto tmp1 = std::make_tuple(cast<uint32_t>(0ULL),cast<uint32_t>(0ULL));
c->HI = std::get<0>(tmp1);
c->LO = std::get<1>(tmp1);
auto tmp2 = std::make_tuple(cast<uint32_t>(3217031168ULL),cast<uint32_t>(3217031172ULL));
c->PC = std::get<0>(tmp2);
c->nextPC = std::get<1>(tmp2);
c->COP0 = std::array<uint32_t,32>{};
c->COP0[cast<int64_t>(15ULL)] = cast<uint32_t>(24064ULL);
c->VfpuCtrl = std::array<uint32_t,16>{};
c->VfpuCtrl[cast<int64_t>(0ULL)] = cast<uint32_t>(228ULL);
c->VfpuCtrl[cast<int64_t>(1ULL)] = cast<uint32_t>(228ULL);
c->ld = allegrex_loadSlot{};
auto tmp3 = std::make_tuple(false,false);
c->delaySlot = std::get<0>(tmp3);
c->pendingDelay = std::get<1>(tmp3);
auto tmp4 = std::make_tuple(false,std::string("",0));
c->Halted = std::get<0>(tmp4);
c->HaltReason = std::get<1>(tmp4);
}
}
// tools/cpu/allegrex/cpu.go:140:1
void allegrex_CPU_SetPC(allegrex_CPU* c,uint32_t pc){
{
auto tmp5 = std::make_tuple(pc,cast<uint32_t>((pc + cast<uint32_t>(4ULL))));
c->PC = std::get<0>(tmp5);
c->nextPC = std::get<1>(tmp5);
c->pendingDelay = false;
}
}
// tools/cpu/allegrex/cpu.go:147:1
void allegrex_CPU_SetReg(allegrex_CPU* c,uint32_t i,uint32_t v){
{
if ((i != cast<uint32_t>(0ULL))) {
auto tmp6 = std::make_tuple(v,v);
c->R[i] = std::get<0>(tmp6);
c->out[i] = std::get<1>(tmp6);
}
}
}
// tools/cpu/allegrex/cpu.go:154:1
uint32_t allegrex_CPU_Reg(allegrex_CPU* c,uint32_t i){
{
return c->R[i];
}
}
// tools/cpu/allegrex/cpu.go:158:1
uint32_t allegrex_CPU_CurPC(allegrex_CPU* c){
{
return c->curPC;
}
}
// tools/cpu/allegrex/cpu.go:163:1
void allegrex_CPU_set(allegrex_CPU* c,uint32_t i,uint32_t v){
{
if ((i != cast<uint32_t>(0ULL))) {
c->out[i] = v;
}
}
}
// tools/cpu/allegrex/cpu.go:171:1
uint32_t allegrex_CPU_read8(allegrex_CPU* c,uint32_t a){
{
return cast<uint32_t>(psp_Machine_Read(c->bus,a));
}
}
// tools/cpu/allegrex/cpu.go:172:1
void allegrex_CPU_write8(allegrex_CPU* c,uint32_t a,uint32_t v){
{
psp_Machine_Write(c->bus,a,cast<uint8_t>(v));
}
}
// tools/cpu/allegrex/cpu.go:194:1
void allegrex_CPU_Exception(allegrex_CPU* c,uint32_t code){
{
uint32_t epc = c->curPC;
uint32_t cause = shl<uint32_t>(code,cast<int64_t>(2ULL));
if (c->delaySlot) {
epc = c->branchAddr;
cause |= cast<uint32_t>(2147483648ULL);
}
c->COP0[cast<int64_t>(14ULL)] = epc;
c->COP0[cast<int64_t>(13ULL)] = cast<uint32_t>(((cast<uint32_t>((c->COP0[cast<int64_t>(13ULL)] & cast<uint32_t>(65280ULL)))) | ((cause & ~(cast<uint32_t>(65280ULL))))));
uint32_t sr = c->COP0[cast<int64_t>(12ULL)];
c->COP0[cast<int64_t>(12ULL)] = cast<uint32_t>((((sr & ~(cast<uint32_t>(63ULL)))) | (cast<uint32_t>(((shl<uint32_t>(sr,cast<int64_t>(2ULL))) & cast<uint32_t>(63ULL))))));
uint32_t target = cast<uint32_t>(2147483776ULL);
if ((cast<uint32_t>((sr & cast<uint32_t>(4194304ULL))) != cast<uint32_t>(0ULL))) {
target = cast<uint32_t>(3217031552ULL);
}
allegrex_CPU_SetPC(c,target);
}
}
// tools/cpu/allegrex/cpu.go:216:1
void allegrex_CPU_rfe(allegrex_CPU* c){
{
uint32_t sr = c->COP0[cast<int64_t>(12ULL)];
c->COP0[cast<int64_t>(12ULL)] = cast<uint32_t>((((sr & ~(cast<uint32_t>(15ULL)))) | (cast<uint32_t>(((shr<uint32_t>(sr,cast<int64_t>(2ULL))) & cast<uint32_t>(15ULL))))));
}
}
// tools/cpu/allegrex/cpu.go:226:1
bool allegrex_CPU_Interrupt(allegrex_CPU* c,bool pending){
{
if (pending) {
c->COP0[cast<int64_t>(13ULL)] |= cast<uint32_t>(1024ULL);
}
else {
c->COP0[cast<int64_t>(13ULL)] &= ~(cast<uint32_t>(1024ULL));
}
uint32_t sr = c->COP0[cast<int64_t>(12ULL)];
if ((cast<uint32_t>((sr & cast<uint32_t>(1ULL))) == cast<uint32_t>(0ULL))) {
return false;
}
if ((cast<uint32_t>((cast<uint32_t>((c->COP0[cast<int64_t>(13ULL)] & sr)) & cast<uint32_t>(65280ULL))) == cast<uint32_t>(0ULL))) {
return false;
}
if (c->pendingDelay) {
return false;
}
if ((c->ld.reg != cast<uint32_t>(0ULL))) {
c->R[c->ld.reg] = c->ld.val;
c->out[c->ld.reg] = c->ld.val;
c->ld = allegrex_loadSlot{};
}
c->curPC = c->PC;
c->delaySlot = false;
allegrex_CPU_Exception(c,cast<uint32_t>(0ULL));
return true;
}
}
// tools/cpu/allegrex/cpu.go:259:1
void allegrex_CPU_addrError(allegrex_CPU* c,uint32_t code,uint32_t addr){
{
c->COP0[cast<int64_t>(8ULL)] = addr;
allegrex_CPU_Exception(c,code);
}
}
// tools/cpu/allegrex/decode.go:16:1
std::string allegrex_reg(uint32_t i){
{
return (std::string("$",1) + allegrex_regName[cast<uint32_t>((i & cast<uint32_t>(31ULL)))]);
}
}
// tools/cpu/allegrex/decode.go:20:1
allegrex_Inst allegrex_Decode(Slice<uint8_t> code,uint32_t addr){
{
if ((len(code) < cast<int64_t>(4ULL))) {
return allegrex_Inst{addr,cast<int64_t>(0ULL),std::string(".word",5),std::string(".word <truncated>",17),cast<allegrex_Flow>(7ULL),{},{},{}};
}
uint32_t w = le_Uint32(code);
allegrex_Inst in = allegrex_Inst{addr,cast<int64_t>(4ULL),{},{},cast<allegrex_Flow>(0ULL),{},{},{}};
uint32_t op = shr<uint32_t>(w,cast<int64_t>(26ULL));
uint32_t rs = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(21ULL))) & cast<uint32_t>(31ULL)));
uint32_t rt = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(31ULL)));
uint32_t rd = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(11ULL))) & cast<uint32_t>(31ULL)));
uint32_t shamt = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(6ULL))) & cast<uint32_t>(31ULL)));
uint32_t funct = cast<uint32_t>((w & cast<uint32_t>(63ULL)));
uint32_t imm = cast<uint32_t>((w & cast<uint32_t>(65535ULL)));
int32_t simm = cast<int32_t>(cast<int16_t>(imm));
uint32_t branchT = cast<uint32_t>((cast<uint32_t>((addr + cast<uint32_t>(4ULL))) + cast<uint32_t>((cast<uint32_t>(simm) * cast<uint32_t>(4ULL)))));
uint32_t jumpT = cast<uint32_t>((cast<uint32_t>(((cast<uint32_t>((addr + cast<uint32_t>(4ULL)))) & cast<uint32_t>(4026531840ULL))) | shl<uint32_t>((cast<uint32_t>((w & cast<uint32_t>(67108863ULL)))),cast<int64_t>(2ULL))));
auto set = [&](std::string mnem,std::string text)->void{
auto tmp7 = std::make_tuple(mnem,text);
in.Mnem = std::get<0>(tmp7);
in.Text = std::get<1>(tmp7);
}
;
{
switch(op){
case cast<uint32_t>(0ULL):{
return allegrex_decodeSpecial(in,w,rs,rt,rd,shamt,funct);
break;}
case cast<uint32_t>(1ULL):{
auto tmp8 = std::make_tuple(cast<allegrex_Flow>(1ULL),branchT,true,true);
in.Flow = std::get<0>(tmp8);
in.Target = std::get<1>(tmp8);
in.HasTarget = std::get<2>(tmp8);
in.HasDelay = std::get<3>(tmp8);
{
switch(rt){
case cast<uint32_t>(0ULL):{
set(std::string("bltz",4),go_fmt_Sprintf(std::string("bltz %s, $%08X",14),allegrex_reg(rs),branchT));
break;}
case cast<uint32_t>(1ULL):{
set(std::string("bgez",4),go_fmt_Sprintf(std::string("bgez %s, $%08X",14),allegrex_reg(rs),branchT));
break;}
case cast<uint32_t>(2ULL):{
set(std::string("bltzl",5),go_fmt_Sprintf(std::string("bltzl %s, $%08X",15),allegrex_reg(rs),branchT));
break;}
case cast<uint32_t>(3ULL):{
set(std::string("bgezl",5),go_fmt_Sprintf(std::string("bgezl %s, $%08X",15),allegrex_reg(rs),branchT));
break;}
case cast<uint32_t>(16ULL):{
set(std::string("bltzal",6),go_fmt_Sprintf(std::string("bltzal %s, $%08X",16),allegrex_reg(rs),branchT));
break;}
case cast<uint32_t>(17ULL):{
set(std::string("bgezal",6),go_fmt_Sprintf(std::string("bgezal %s, $%08X",16),allegrex_reg(rs),branchT));
break;}
case cast<uint32_t>(18ULL):{
set(std::string("bltzall",7),go_fmt_Sprintf(std::string("bltzall %s, $%08X",17),allegrex_reg(rs),branchT));
break;}
case cast<uint32_t>(19ULL):{
set(std::string("bgezall",7),go_fmt_Sprintf(std::string("bgezall %s, $%08X",17),allegrex_reg(rs),branchT));
break;}
default:{
return allegrex_word(in,w);
break;}
}}
break;}
case cast<uint32_t>(2ULL):{
auto tmp9 = std::make_tuple(cast<allegrex_Flow>(2ULL),jumpT,true,true);
in.Flow = std::get<0>(tmp9);
in.Target = std::get<1>(tmp9);
in.HasTarget = std::get<2>(tmp9);
in.HasDelay = std::get<3>(tmp9);
set(std::string("j",1),go_fmt_Sprintf(std::string("j $%08X",7),jumpT));
break;}
case cast<uint32_t>(3ULL):{
auto tmp10 = std::make_tuple(cast<allegrex_Flow>(3ULL),jumpT,true,true);
in.Flow = std::get<0>(tmp10);
in.Target = std::get<1>(tmp10);
in.HasTarget = std::get<2>(tmp10);
in.HasDelay = std::get<3>(tmp10);
set(std::string("jal",3),go_fmt_Sprintf(std::string("jal $%08X",9),jumpT));
break;}
case cast<uint32_t>(4ULL):{
auto tmp11 = std::make_tuple(cast<allegrex_Flow>(1ULL),branchT,true,true);
in.Flow = std::get<0>(tmp11);
in.Target = std::get<1>(tmp11);
in.HasTarget = std::get<2>(tmp11);
in.HasDelay = std::get<3>(tmp11);
if (((rs == cast<uint32_t>(0ULL)) && (rt == cast<uint32_t>(0ULL)))) {
set(std::string("b",1),go_fmt_Sprintf(std::string("b $%08X",7),branchT));
in.Flow = cast<allegrex_Flow>(2ULL);
}
else {
set(std::string("beq",3),go_fmt_Sprintf(std::string("beq %s, %s, $%08X",17),allegrex_reg(rs),allegrex_reg(rt),branchT));
}
break;}
case cast<uint32_t>(5ULL):{
auto tmp12 = std::make_tuple(cast<allegrex_Flow>(1ULL),branchT,true,true);
in.Flow = std::get<0>(tmp12);
in.Target = std::get<1>(tmp12);
in.HasTarget = std::get<2>(tmp12);
in.HasDelay = std::get<3>(tmp12);
set(std::string("bne",3),go_fmt_Sprintf(std::string("bne %s, %s, $%08X",17),allegrex_reg(rs),allegrex_reg(rt),branchT));
break;}
case cast<uint32_t>(6ULL):{
auto tmp13 = std::make_tuple(cast<allegrex_Flow>(1ULL),branchT,true,true);
in.Flow = std::get<0>(tmp13);
in.Target = std::get<1>(tmp13);
in.HasTarget = std::get<2>(tmp13);
in.HasDelay = std::get<3>(tmp13);
set(std::string("blez",4),go_fmt_Sprintf(std::string("blez %s, $%08X",14),allegrex_reg(rs),branchT));
break;}
case cast<uint32_t>(7ULL):{
auto tmp14 = std::make_tuple(cast<allegrex_Flow>(1ULL),branchT,true,true);
in.Flow = std::get<0>(tmp14);
in.Target = std::get<1>(tmp14);
in.HasTarget = std::get<2>(tmp14);
in.HasDelay = std::get<3>(tmp14);
set(std::string("bgtz",4),go_fmt_Sprintf(std::string("bgtz %s, $%08X",14),allegrex_reg(rs),branchT));
break;}
case cast<uint32_t>(8ULL):{
set(std::string("addi",4),go_fmt_Sprintf(std::string("addi %s, %s, %d",15),allegrex_reg(rt),allegrex_reg(rs),simm));
break;}
case cast<uint32_t>(9ULL):{
set(std::string("addiu",5),go_fmt_Sprintf(std::string("addiu %s, %s, %d",16),allegrex_reg(rt),allegrex_reg(rs),simm));
break;}
case cast<uint32_t>(10ULL):{
set(std::string("slti",4),go_fmt_Sprintf(std::string("slti %s, %s, %d",15),allegrex_reg(rt),allegrex_reg(rs),simm));
break;}
case cast<uint32_t>(11ULL):{
set(std::string("sltiu",5),go_fmt_Sprintf(std::string("sltiu %s, %s, %d",16),allegrex_reg(rt),allegrex_reg(rs),simm));
break;}
case cast<uint32_t>(12ULL):{
set(std::string("andi",4),go_fmt_Sprintf(std::string("andi %s, %s, 0x%X",17),allegrex_reg(rt),allegrex_reg(rs),imm));
break;}
case cast<uint32_t>(13ULL):{
set(std::string("ori",3),go_fmt_Sprintf(std::string("ori %s, %s, 0x%X",16),allegrex_reg(rt),allegrex_reg(rs),imm));
break;}
case cast<uint32_t>(14ULL):{
set(std::string("xori",4),go_fmt_Sprintf(std::string("xori %s, %s, 0x%X",17),allegrex_reg(rt),allegrex_reg(rs),imm));
break;}
case cast<uint32_t>(15ULL):{
set(std::string("lui",3),go_fmt_Sprintf(std::string("lui %s, 0x%X",12),allegrex_reg(rt),imm));
break;}
case cast<uint32_t>(16ULL):{
return allegrex_decodeCop0(in,w,rs,rt,rd);
break;}
case cast<uint32_t>(17ULL):{
return allegrex_decodeCop1(in,w,rs,rt,rd,shamt);
break;}
case cast<uint32_t>(18ULL):{
return allegrex_decodeCop2(in,w,rs,rt,rd);
break;}
case cast<uint32_t>(20ULL):case cast<uint32_t>(21ULL):case cast<uint32_t>(22ULL):case cast<uint32_t>(23ULL):{
auto tmp15 = std::make_tuple(cast<allegrex_Flow>(1ULL),branchT,true,true);
in.Flow = std::get<0>(tmp15);
in.Target = std::get<1>(tmp15);
in.HasTarget = std::get<2>(tmp15);
in.HasDelay = std::get<3>(tmp15);
std::string m = [](){std::array<std::string,24> v{};v[20]=std::string("beql",4);v[21]=std::string("bnel",4);v[22]=std::string("blezl",5);v[23]=std::string("bgtzl",5);return v;}()[op];
if (((op == cast<uint32_t>(20ULL)) || (op == cast<uint32_t>(21ULL)))) {
set(m,go_fmt_Sprintf(std::string("%s %s, %s, $%08X",16),m,allegrex_reg(rs),allegrex_reg(rt),branchT));
}
else {
set(m,go_fmt_Sprintf(std::string("%s %s, $%08X",12),m,allegrex_reg(rs),branchT));
}
break;}
case cast<uint32_t>(28ULL):{
return allegrex_decodeSpecial2(in,w,rs,rt,rd,funct);
break;}
case cast<uint32_t>(31ULL):{
return allegrex_decodeSpecial3(in,w,rs,rt,rd,shamt,funct);
break;}
case cast<uint32_t>(32ULL):case cast<uint32_t>(33ULL):case cast<uint32_t>(34ULL):case cast<uint32_t>(35ULL):case cast<uint32_t>(36ULL):case cast<uint32_t>(37ULL):case cast<uint32_t>(38ULL):{
std::string m = [](){std::array<std::string,39> v{};v[32]=std::string("lb",2);v[33]=std::string("lh",2);v[34]=std::string("lwl",3);v[35]=std::string("lw",2);v[36]=std::string("lbu",3);v[37]=std::string("lhu",3);v[38]=std::string("lwr",3);return v;}()[op];
set(m,go_fmt_Sprintf(std::string("%s %s, %d(%s)",13),m,allegrex_reg(rt),simm,allegrex_reg(rs)));
break;}
case cast<uint32_t>(40ULL):case cast<uint32_t>(41ULL):case cast<uint32_t>(42ULL):case cast<uint32_t>(43ULL):case cast<uint32_t>(46ULL):{
std::string m = get(Map<uint32_t,std::string>{{cast<uint32_t>(40ULL),std::string("sb",2)},{cast<uint32_t>(41ULL),std::string("sh",2)},{cast<uint32_t>(42ULL),std::string("swl",3)},{cast<uint32_t>(43ULL),std::string("sw",2)},{cast<uint32_t>(46ULL),std::string("swr",3)}},op);
set(m,go_fmt_Sprintf(std::string("%s %s, %d(%s)",13),m,allegrex_reg(rt),simm,allegrex_reg(rs)));
break;}
case cast<uint32_t>(48ULL):{
set(std::string("ll",2),go_fmt_Sprintf(std::string("ll %s, %d(%s)",13),allegrex_reg(rt),simm,allegrex_reg(rs)));
break;}
case cast<uint32_t>(56ULL):{
set(std::string("sc",2),go_fmt_Sprintf(std::string("sc %s, %d(%s)",13),allegrex_reg(rt),simm,allegrex_reg(rs)));
break;}
case cast<uint32_t>(49ULL):{
set(std::string("lwc1",4),go_fmt_Sprintf(std::string("lwc1 $f%d, %d(%s)",17),rt,simm,allegrex_reg(rs)));
break;}
case cast<uint32_t>(57ULL):{
set(std::string("swc1",4),go_fmt_Sprintf(std::string("swc1 $f%d, %d(%s)",17),rt,simm,allegrex_reg(rs)));
break;}
case cast<uint32_t>(50ULL):case cast<uint32_t>(54ULL):case cast<uint32_t>(53ULL):case cast<uint32_t>(58ULL):case cast<uint32_t>(62ULL):case cast<uint32_t>(61ULL):case cast<uint32_t>(52ULL):case cast<uint32_t>(55ULL):case cast<uint32_t>(60ULL):case cast<uint32_t>(24ULL):case cast<uint32_t>(25ULL):case cast<uint32_t>(27ULL):{
return allegrex_decodeVFPU(in,w,op,rs,rt,cast<uint32_t>(simm));
break;}
default:{
return allegrex_word(in,w);
break;}
}}
return in;
}
}
// tools/cpu/allegrex/decode.go:154:1
allegrex_Inst allegrex_decodeSpecial2(allegrex_Inst in,uint32_t w,uint32_t rs,uint32_t rt,uint32_t rd,uint32_t funct){
{
auto set = [&](std::string mnem,std::string text)->allegrex_Inst{
auto tmp16 = std::make_tuple(mnem,text);
in.Mnem = std::get<0>(tmp16);
in.Text = std::get<1>(tmp16);
return in;
}
;
{
switch(funct){
case cast<uint32_t>(0ULL):{
return set(std::string("madd",4),go_fmt_Sprintf(std::string("madd %s, %s",11),allegrex_reg(rs),allegrex_reg(rt)));
break;}
case cast<uint32_t>(1ULL):{
return set(std::string("maddu",5),go_fmt_Sprintf(std::string("maddu %s, %s",12),allegrex_reg(rs),allegrex_reg(rt)));
break;}
case cast<uint32_t>(2ULL):{
return set(std::string("mul",3),go_fmt_Sprintf(std::string("mul %s, %s, %s",14),allegrex_reg(rd),allegrex_reg(rs),allegrex_reg(rt)));
break;}
case cast<uint32_t>(4ULL):{
return set(std::string("msub",4),go_fmt_Sprintf(std::string("msub %s, %s",11),allegrex_reg(rs),allegrex_reg(rt)));
break;}
case cast<uint32_t>(5ULL):{
return set(std::string("msubu",5),go_fmt_Sprintf(std::string("msubu %s, %s",12),allegrex_reg(rs),allegrex_reg(rt)));
break;}
case cast<uint32_t>(32ULL):{
return set(std::string("clz",3),go_fmt_Sprintf(std::string("clz %s, %s",10),allegrex_reg(rd),allegrex_reg(rs)));
break;}
case cast<uint32_t>(33ULL):{
return set(std::string("clo",3),go_fmt_Sprintf(std::string("clo %s, %s",10),allegrex_reg(rd),allegrex_reg(rs)));
break;}
}}
return allegrex_word(in,w);
}
}
// tools/cpu/allegrex/decode.go:177:1
allegrex_Inst allegrex_decodeSpecial3(allegrex_Inst in,uint32_t w,uint32_t rs,uint32_t rt,uint32_t rd,uint32_t shamt,uint32_t funct){
{
auto set = [&](std::string mnem,std::string text)->allegrex_Inst{
auto tmp17 = std::make_tuple(mnem,text);
in.Mnem = std::get<0>(tmp17);
in.Text = std::get<1>(tmp17);
return in;
}
;
{
switch(funct){
case cast<uint32_t>(0ULL):{
return set(std::string("ext",3),go_fmt_Sprintf(std::string("ext %s, %s, %d, %d",18),allegrex_reg(rt),allegrex_reg(rs),shamt,cast<uint32_t>((rd + cast<uint32_t>(1ULL)))));
break;}
case cast<uint32_t>(4ULL):{
return set(std::string("ins",3),go_fmt_Sprintf(std::string("ins %s, %s, %d, %d",18),allegrex_reg(rt),allegrex_reg(rs),shamt,cast<uint32_t>((cast<uint32_t>((rd - shamt)) + cast<uint32_t>(1ULL)))));
break;}
case cast<uint32_t>(32ULL):{
{
switch(shamt){
case cast<uint32_t>(2ULL):{
return set(std::string("wsbh",4),go_fmt_Sprintf(std::string("wsbh %s, %s",11),allegrex_reg(rd),allegrex_reg(rt)));
break;}
case cast<uint32_t>(16ULL):{
return set(std::string("seb",3),go_fmt_Sprintf(std::string("seb %s, %s",10),allegrex_reg(rd),allegrex_reg(rt)));
break;}
case cast<uint32_t>(24ULL):{
return set(std::string("seh",3),go_fmt_Sprintf(std::string("seh %s, %s",10),allegrex_reg(rd),allegrex_reg(rt)));
break;}
case cast<uint32_t>(20ULL):{
return set(std::string("bitrev",6),go_fmt_Sprintf(std::string("bitrev %s, %s",13),allegrex_reg(rd),allegrex_reg(rt)));
break;}
}}
break;}
}}
return allegrex_word(in,w);
}
}
// tools/cpu/allegrex/decode.go:201:1
allegrex_Inst allegrex_decodeSpecial(allegrex_Inst in,uint32_t w,uint32_t rs,uint32_t rt,uint32_t rd,uint32_t shamt,uint32_t funct){
{
auto set = [&](std::string mnem,std::string text)->allegrex_Inst{
auto tmp18 = std::make_tuple(mnem,text);
in.Mnem = std::get<0>(tmp18);
in.Text = std::get<1>(tmp18);
return in;
}
;
{
switch(funct){
case cast<uint32_t>(0ULL):{
if ((w == cast<uint32_t>(0ULL))) {
return set(std::string("nop",3),std::string("nop",3));
}
return set(std::string("sll",3),go_fmt_Sprintf(std::string("sll %s, %s, %d",14),allegrex_reg(rd),allegrex_reg(rt),shamt));
break;}
case cast<uint32_t>(2ULL):{
if ((cast<uint32_t>((rs & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL))) {
return set(std::string("rotr",4),go_fmt_Sprintf(std::string("rotr %s, %s, %d",15),allegrex_reg(rd),allegrex_reg(rt),shamt));
}
return set(std::string("srl",3),go_fmt_Sprintf(std::string("srl %s, %s, %d",14),allegrex_reg(rd),allegrex_reg(rt),shamt));
break;}
case cast<uint32_t>(3ULL):{
return set(std::string("sra",3),go_fmt_Sprintf(std::string("sra %s, %s, %d",14),allegrex_reg(rd),allegrex_reg(rt),shamt));
break;}
case cast<uint32_t>(4ULL):{
return set(std::string("sllv",4),go_fmt_Sprintf(std::string("sllv %s, %s, %s",15),allegrex_reg(rd),allegrex_reg(rt),allegrex_reg(rs)));
break;}
case cast<uint32_t>(6ULL):{
if ((cast<uint32_t>((shamt & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL))) {
return set(std::string("rotrv",5),go_fmt_Sprintf(std::string("rotrv %s, %s, %s",16),allegrex_reg(rd),allegrex_reg(rt),allegrex_reg(rs)));
}
return set(std::string("srlv",4),go_fmt_Sprintf(std::string("srlv %s, %s, %s",15),allegrex_reg(rd),allegrex_reg(rt),allegrex_reg(rs)));
break;}
case cast<uint32_t>(7ULL):{
return set(std::string("srav",4),go_fmt_Sprintf(std::string("srav %s, %s, %s",15),allegrex_reg(rd),allegrex_reg(rt),allegrex_reg(rs)));
break;}
case cast<uint32_t>(8ULL):{
in.HasDelay = true;
if ((rs == cast<uint32_t>(31ULL))) {
in.Flow = cast<allegrex_Flow>(4ULL);
}
else {
in.Flow = cast<allegrex_Flow>(5ULL);
}
return set(std::string("jr",2),go_fmt_Sprintf(std::string("jr %s",5),allegrex_reg(rs)));
break;}
case cast<uint32_t>(9ULL):{
auto tmp19 = std::make_tuple(cast<allegrex_Flow>(6ULL),true);
in.Flow = std::get<0>(tmp19);
in.HasDelay = std::get<1>(tmp19);
if ((rd == cast<uint32_t>(31ULL))) {
return set(std::string("jalr",4),go_fmt_Sprintf(std::string("jalr %s",7),allegrex_reg(rs)));
}
return set(std::string("jalr",4),go_fmt_Sprintf(std::string("jalr %s, %s",11),allegrex_reg(rd),allegrex_reg(rs)));
break;}
case cast<uint32_t>(12ULL):{
in.Flow = cast<allegrex_Flow>(7ULL);
return set(std::string("syscall",7),std::string("syscall",7));
break;}
case cast<uint32_t>(10ULL):{
return set(std::string("movz",4),go_fmt_Sprintf(std::string("movz %s, %s, %s",15),allegrex_reg(rd),allegrex_reg(rs),allegrex_reg(rt)));
break;}
case cast<uint32_t>(11ULL):{
return set(std::string("movn",4),go_fmt_Sprintf(std::string("movn %s, %s, %s",15),allegrex_reg(rd),allegrex_reg(rs),allegrex_reg(rt)));
break;}
case cast<uint32_t>(13ULL):{
in.Flow = cast<allegrex_Flow>(7ULL);
return set(std::string("break",5),std::string("break",5));
break;}
case cast<uint32_t>(15ULL):{
return set(std::string("sync",4),std::string("sync",4));
break;}
case cast<uint32_t>(16ULL):{
return set(std::string("mfhi",4),go_fmt_Sprintf(std::string("mfhi %s",7),allegrex_reg(rd)));
break;}
case cast<uint32_t>(17ULL):{
return set(std::string("mthi",4),go_fmt_Sprintf(std::string("mthi %s",7),allegrex_reg(rs)));
break;}
case cast<uint32_t>(18ULL):{
return set(std::string("mflo",4),go_fmt_Sprintf(std::string("mflo %s",7),allegrex_reg(rd)));
break;}
case cast<uint32_t>(19ULL):{
return set(std::string("mtlo",4),go_fmt_Sprintf(std::string("mtlo %s",7),allegrex_reg(rs)));
break;}
case cast<uint32_t>(24ULL):{
return set(std::string("mult",4),go_fmt_Sprintf(std::string("mult %s, %s",11),allegrex_reg(rs),allegrex_reg(rt)));
break;}
case cast<uint32_t>(25ULL):{
return set(std::string("multu",5),go_fmt_Sprintf(std::string("multu %s, %s",12),allegrex_reg(rs),allegrex_reg(rt)));
break;}
case cast<uint32_t>(26ULL):{
return set(std::string("div",3),go_fmt_Sprintf(std::string("div %s, %s",10),allegrex_reg(rs),allegrex_reg(rt)));
break;}
case cast<uint32_t>(27ULL):{
return set(std::string("divu",4),go_fmt_Sprintf(std::string("divu %s, %s",11),allegrex_reg(rs),allegrex_reg(rt)));
break;}
case cast<uint32_t>(32ULL):{
return set(std::string("add",3),go_fmt_Sprintf(std::string("add %s, %s, %s",14),allegrex_reg(rd),allegrex_reg(rs),allegrex_reg(rt)));
break;}
case cast<uint32_t>(33ULL):{
return set(std::string("addu",4),go_fmt_Sprintf(std::string("addu %s, %s, %s",15),allegrex_reg(rd),allegrex_reg(rs),allegrex_reg(rt)));
break;}
case cast<uint32_t>(34ULL):{
return set(std::string("sub",3),go_fmt_Sprintf(std::string("sub %s, %s, %s",14),allegrex_reg(rd),allegrex_reg(rs),allegrex_reg(rt)));
break;}
case cast<uint32_t>(35ULL):{
return set(std::string("subu",4),go_fmt_Sprintf(std::string("subu %s, %s, %s",15),allegrex_reg(rd),allegrex_reg(rs),allegrex_reg(rt)));
break;}
case cast<uint32_t>(36ULL):{
return set(std::string("and",3),go_fmt_Sprintf(std::string("and %s, %s, %s",14),allegrex_reg(rd),allegrex_reg(rs),allegrex_reg(rt)));
break;}
case cast<uint32_t>(37ULL):{
if ((rt == cast<uint32_t>(0ULL))) {
return set(std::string("move",4),go_fmt_Sprintf(std::string("move %s, %s",11),allegrex_reg(rd),allegrex_reg(rs)));
}
return set(std::string("or",2),go_fmt_Sprintf(std::string("or %s, %s, %s",13),allegrex_reg(rd),allegrex_reg(rs),allegrex_reg(rt)));
break;}
case cast<uint32_t>(38ULL):{
return set(std::string("xor",3),go_fmt_Sprintf(std::string("xor %s, %s, %s",14),allegrex_reg(rd),allegrex_reg(rs),allegrex_reg(rt)));
break;}
case cast<uint32_t>(39ULL):{
return set(std::string("nor",3),go_fmt_Sprintf(std::string("nor %s, %s, %s",14),allegrex_reg(rd),allegrex_reg(rs),allegrex_reg(rt)));
break;}
case cast<uint32_t>(42ULL):{
return set(std::string("slt",3),go_fmt_Sprintf(std::string("slt %s, %s, %s",14),allegrex_reg(rd),allegrex_reg(rs),allegrex_reg(rt)));
break;}
case cast<uint32_t>(43ULL):{
return set(std::string("sltu",4),go_fmt_Sprintf(std::string("sltu %s, %s, %s",15),allegrex_reg(rd),allegrex_reg(rs),allegrex_reg(rt)));
break;}
case cast<uint32_t>(22ULL):{
return set(std::string("clz",3),go_fmt_Sprintf(std::string("clz %s, %s",10),allegrex_reg(rd),allegrex_reg(rs)));
break;}
case cast<uint32_t>(23ULL):{
return set(std::string("clo",3),go_fmt_Sprintf(std::string("clo %s, %s",10),allegrex_reg(rd),allegrex_reg(rs)));
break;}
case cast<uint32_t>(44ULL):{
return set(std::string("max",3),go_fmt_Sprintf(std::string("max %s, %s, %s",14),allegrex_reg(rd),allegrex_reg(rs),allegrex_reg(rt)));
break;}
case cast<uint32_t>(45ULL):{
return set(std::string("min",3),go_fmt_Sprintf(std::string("min %s, %s, %s",14),allegrex_reg(rd),allegrex_reg(rs),allegrex_reg(rt)));
break;}
}}
return allegrex_word(in,w);
}
}
// tools/cpu/allegrex/decode.go:303:1
allegrex_Inst allegrex_decodeCop0(allegrex_Inst in,uint32_t w,uint32_t rs,uint32_t rt,uint32_t rd){
{
auto set = [&](std::string mnem,std::string text)->allegrex_Inst{
auto tmp20 = std::make_tuple(mnem,text);
in.Mnem = std::get<0>(tmp20);
in.Text = std::get<1>(tmp20);
return in;
}
;
if ((cast<uint32_t>((w & cast<uint32_t>(33554432ULL))) != cast<uint32_t>(0ULL))) {
if ((cast<uint32_t>((w & cast<uint32_t>(63ULL))) == cast<uint32_t>(16ULL))) {
return set(std::string("rfe",3),std::string("rfe",3));
}
return allegrex_word(in,w);
}
{
switch(rs){
case cast<uint32_t>(0ULL):{
return set(std::string("mfc0",4),go_fmt_Sprintf(std::string("mfc0 %s, $%d",12),allegrex_reg(rt),rd));
break;}
case cast<uint32_t>(4ULL):{
return set(std::string("mtc0",4),go_fmt_Sprintf(std::string("mtc0 %s, $%d",12),allegrex_reg(rt),rd));
break;}
}}
return allegrex_word(in,w);
}
}
// tools/cpu/allegrex/decode.go:321:1
allegrex_Inst allegrex_word(allegrex_Inst in,uint32_t w){
{
auto tmp21 = std::make_tuple(std::string(".word",5),go_fmt_Sprintf(std::string(".word 0x%08X",12),w),cast<allegrex_Flow>(7ULL));
in.Mnem = std::get<0>(tmp21);
in.Text = std::get<1>(tmp21);
in.Flow = std::get<2>(tmp21);
auto tmp22 = std::make_tuple(false,false);
in.HasTarget = std::get<0>(tmp22);
in.HasDelay = std::get<1>(tmp22);
return in;
}
}
// tools/cpu/allegrex/exec.go:12:1
int64_t allegrex_CPU_Step(allegrex_CPU* c){
{
if (c->Halted) {
return cast<int64_t>(0ULL);
}
c->curPC = c->PC;
c->delaySlot = c->pendingDelay;
c->pendingDelay = false;
uint32_t w = allegrex_CPU_read32(c,c->PC);
c->PC = c->nextPC;
c->nextPC += cast<uint32_t>(4ULL);
if (c->nullifyNext) {
c->nullifyNext = false;
c->R = c->out;
c->Steps++;
return cast<int64_t>(1ULL);
}
allegrex_CPU_execute(c,w);
c->R = c->out;
c->Steps++;
return cast<int64_t>(1ULL);
}
}
// tools/cpu/allegrex/exec.go:41:1
uint32_t allegrex_CPU_reg(allegrex_CPU* c,uint32_t i){
{
return c->R[i];
}
}
// tools/cpu/allegrex/exec.go:45:1
void allegrex_CPU_load(allegrex_CPU* c,uint32_t reg,uint32_t val){
{
allegrex_CPU_set(c,reg,val);
}
}
// tools/cpu/allegrex/exec.go:49:1
void allegrex_CPU_doBranch(allegrex_CPU* c,bool taken,uint32_t target){
{
c->pendingDelay = true;
c->branchAddr = c->curPC;
if (taken) {
c->nextPC = target;
}
}
}
// tools/cpu/allegrex/exec.go:57:1
void allegrex_CPU_execute(allegrex_CPU* c,uint32_t w){
{
uint32_t op = shr<uint32_t>(w,cast<int64_t>(26ULL));
uint32_t rs = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(21ULL))) & cast<uint32_t>(31ULL)));
uint32_t rt = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(31ULL)));
uint32_t rd = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(11ULL))) & cast<uint32_t>(31ULL)));
uint32_t shamt = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(6ULL))) & cast<uint32_t>(31ULL)));
uint32_t imm = cast<uint32_t>((w & cast<uint32_t>(65535ULL)));
uint32_t simm = cast<uint32_t>(cast<int32_t>(cast<int16_t>(imm)));
uint32_t branchT = cast<uint32_t>((cast<uint32_t>((c->curPC + cast<uint32_t>(4ULL))) + shl<uint32_t>(simm,cast<int64_t>(2ULL))));
uint32_t jumpT = cast<uint32_t>((cast<uint32_t>(((cast<uint32_t>((c->curPC + cast<uint32_t>(4ULL)))) & cast<uint32_t>(4026531840ULL))) | shl<uint32_t>((cast<uint32_t>((w & cast<uint32_t>(67108863ULL)))),cast<int64_t>(2ULL))));
{
switch(op){
case cast<uint32_t>(0ULL):{
allegrex_CPU_special(c,w,rs,rt,rd,shamt);
break;}
case cast<uint32_t>(1ULL):{
{
switch(rt){
case cast<uint32_t>(0ULL):case cast<uint32_t>(1ULL):case cast<uint32_t>(2ULL):case cast<uint32_t>(3ULL):case cast<uint32_t>(16ULL):case cast<uint32_t>(17ULL):case cast<uint32_t>(18ULL):case cast<uint32_t>(19ULL):{
int32_t s = cast<int32_t>(allegrex_CPU_reg(c,rs));
if ((cast<uint32_t>((rt & cast<uint32_t>(16ULL))) != cast<uint32_t>(0ULL))) {
allegrex_CPU_set(c,cast<uint32_t>(31ULL),cast<uint32_t>((c->curPC + cast<uint32_t>(8ULL))));
}
bool taken = ((((cast<uint32_t>((rt & cast<uint32_t>(1ULL))) == cast<uint32_t>(0ULL)) && (s < cast<int32_t>(0ULL)))) || (((cast<uint32_t>((rt & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL)) && (s >= cast<int32_t>(0ULL)))));
if (((cast<uint32_t>((rt & cast<uint32_t>(2ULL))) != cast<uint32_t>(0ULL)) && (!taken))) {
c->nullifyNext = true;
}
else {
allegrex_CPU_doBranch(c,taken,branchT);
}
break;}
default:{
allegrex_CPU_Halt(c,std::string("unimplemented regimm rt=0x%02X (word 0x%08X) at 0x%08X",54),rt,w,c->curPC);
break;}
}}
break;}
case cast<uint32_t>(2ULL):{
allegrex_CPU_doBranch(c,true,jumpT);
break;}
case cast<uint32_t>(3ULL):{
allegrex_CPU_set(c,cast<uint32_t>(31ULL),cast<uint32_t>((c->curPC + cast<uint32_t>(8ULL))));
allegrex_CPU_doBranch(c,true,jumpT);
break;}
case cast<uint32_t>(4ULL):{
allegrex_CPU_doBranch(c,(allegrex_CPU_reg(c,rs) == allegrex_CPU_reg(c,rt)),branchT);
break;}
case cast<uint32_t>(5ULL):{
allegrex_CPU_doBranch(c,(allegrex_CPU_reg(c,rs) != allegrex_CPU_reg(c,rt)),branchT);
break;}
case cast<uint32_t>(6ULL):{
allegrex_CPU_doBranch(c,(cast<int32_t>(allegrex_CPU_reg(c,rs)) <= cast<int32_t>(0ULL)),branchT);
break;}
case cast<uint32_t>(7ULL):{
allegrex_CPU_doBranch(c,(cast<int32_t>(allegrex_CPU_reg(c,rs)) > cast<int32_t>(0ULL)),branchT);
break;}
case cast<uint32_t>(8ULL):{
{
auto tmp23 = allegrex_addOv(allegrex_CPU_reg(c,rs),simm);
uint32_t r = std::get<0>(tmp23);
bool ov = std::get<1>(tmp23);
if (ov) {
allegrex_CPU_Exception(c,cast<uint32_t>(12ULL));
}
else {
allegrex_CPU_set(c,rt,r);
}
}
break;}
case cast<uint32_t>(9ULL):{
allegrex_CPU_set(c,rt,cast<uint32_t>((allegrex_CPU_reg(c,rs) + simm)));
break;}
case cast<uint32_t>(10ULL):{
allegrex_CPU_set(c,rt,allegrex_b2u((cast<int32_t>(allegrex_CPU_reg(c,rs)) < cast<int32_t>(simm))));
break;}
case cast<uint32_t>(11ULL):{
allegrex_CPU_set(c,rt,allegrex_b2u((allegrex_CPU_reg(c,rs) < simm)));
break;}
case cast<uint32_t>(12ULL):{
allegrex_CPU_set(c,rt,cast<uint32_t>((allegrex_CPU_reg(c,rs) & imm)));
break;}
case cast<uint32_t>(13ULL):{
allegrex_CPU_set(c,rt,cast<uint32_t>((allegrex_CPU_reg(c,rs) | imm)));
break;}
case cast<uint32_t>(14ULL):{
allegrex_CPU_set(c,rt,cast<uint32_t>((allegrex_CPU_reg(c,rs) ^ imm)));
break;}
case cast<uint32_t>(15ULL):{
allegrex_CPU_set(c,rt,shl<uint32_t>(imm,cast<int64_t>(16ULL)));
break;}
case cast<uint32_t>(16ULL):{
allegrex_CPU_cop0(c,w,rs,rt,rd);
break;}
case cast<uint32_t>(17ULL):{
allegrex_CPU_cop1(c,w,rs,rt,rd,shamt);
break;}
case cast<uint32_t>(18ULL):{
allegrex_CPU_cop2(c,w,rs,rt,rd);
break;}
case cast<uint32_t>(20ULL):case cast<uint32_t>(21ULL):case cast<uint32_t>(22ULL):case cast<uint32_t>(23ULL):{
allegrex_CPU_branchLikely(c,op,rs,rt,branchT);
break;}
case cast<uint32_t>(28ULL):{
allegrex_CPU_special2(c,w,rs,rt,rd);
break;}
case cast<uint32_t>(31ULL):{
allegrex_CPU_special3(c,w,rs,rt,rd,shamt);
break;}
case cast<uint32_t>(32ULL):case cast<uint32_t>(33ULL):case cast<uint32_t>(34ULL):case cast<uint32_t>(35ULL):case cast<uint32_t>(36ULL):case cast<uint32_t>(37ULL):case cast<uint32_t>(38ULL):{
allegrex_CPU_loadOp(c,op,rs,rt,simm);
break;}
case cast<uint32_t>(40ULL):case cast<uint32_t>(41ULL):case cast<uint32_t>(42ULL):case cast<uint32_t>(43ULL):case cast<uint32_t>(46ULL):{
allegrex_CPU_storeOp(c,op,rs,rt,simm);
break;}
case cast<uint32_t>(48ULL):{
uint32_t addr = cast<uint32_t>((allegrex_CPU_reg(c,rs) + simm));
if ((cast<uint32_t>((addr & cast<uint32_t>(3ULL))) != cast<uint32_t>(0ULL))) {
allegrex_CPU_addrError(c,cast<uint32_t>(4ULL),addr);
return ;
}
allegrex_CPU_load(c,rt,allegrex_CPU_read32(c,addr));
break;}
case cast<uint32_t>(56ULL):{
uint32_t addr = cast<uint32_t>((allegrex_CPU_reg(c,rs) + simm));
if ((cast<uint32_t>((addr & cast<uint32_t>(3ULL))) != cast<uint32_t>(0ULL))) {
allegrex_CPU_addrError(c,cast<uint32_t>(5ULL),addr);
return ;
}
allegrex_CPU_write32(c,addr,allegrex_CPU_reg(c,rt));
allegrex_CPU_set(c,rt,cast<uint32_t>(1ULL));
break;}
case cast<uint32_t>(49ULL):{
uint32_t addr = cast<uint32_t>((allegrex_CPU_reg(c,rs) + simm));
if ((cast<uint32_t>((addr & cast<uint32_t>(3ULL))) != cast<uint32_t>(0ULL))) {
allegrex_CPU_addrError(c,cast<uint32_t>(4ULL),addr);
return ;
}
c->F[cast<uint32_t>((rt & cast<uint32_t>(31ULL)))] = allegrex_CPU_read32(c,addr);
break;}
case cast<uint32_t>(57ULL):{
uint32_t addr = cast<uint32_t>((allegrex_CPU_reg(c,rs) + simm));
if ((cast<uint32_t>((addr & cast<uint32_t>(3ULL))) != cast<uint32_t>(0ULL))) {
allegrex_CPU_addrError(c,cast<uint32_t>(5ULL),addr);
return ;
}
allegrex_CPU_write32(c,addr,c->F[cast<uint32_t>((rt & cast<uint32_t>(31ULL)))]);
break;}
case cast<uint32_t>(50ULL):case cast<uint32_t>(54ULL):case cast<uint32_t>(53ULL):case cast<uint32_t>(58ULL):case cast<uint32_t>(62ULL):case cast<uint32_t>(61ULL):case cast<uint32_t>(52ULL):case cast<uint32_t>(55ULL):case cast<uint32_t>(60ULL):case cast<uint32_t>(24ULL):case cast<uint32_t>(25ULL):case cast<uint32_t>(27ULL):{
allegrex_CPU_vfpuOp(c,w,op,rs,rt,simm);
break;}
case cast<uint32_t>(63ULL):{
if ((shr<uint32_t>(w,cast<int64_t>(16ULL)) != cast<uint32_t>(65535ULL))) {
allegrex_CPU_Halt(c,std::string("unimplemented opcode 0x3F (word 0x%08X) at 0x%08X",49),w,c->curPC);
}
break;}
default:{
allegrex_CPU_Halt(c,std::string("unimplemented opcode 0x%02X (word 0x%08X) at 0x%08X",51),op,w,c->curPC);
break;}
}}
}
}
// tools/cpu/allegrex/exec.go:178:1
void allegrex_CPU_branchLikely(allegrex_CPU* c,uint32_t op,uint32_t rs,uint32_t rt,uint32_t target){
{
bool taken={};
{
switch(op){
case cast<uint32_t>(20ULL):{
taken = (allegrex_CPU_reg(c,rs) == allegrex_CPU_reg(c,rt));
break;}
case cast<uint32_t>(21ULL):{
taken = (allegrex_CPU_reg(c,rs) != allegrex_CPU_reg(c,rt));
break;}
case cast<uint32_t>(22ULL):{
taken = (cast<int32_t>(allegrex_CPU_reg(c,rs)) <= cast<int32_t>(0ULL));
break;}
case cast<uint32_t>(23ULL):{
taken = (cast<int32_t>(allegrex_CPU_reg(c,rs)) > cast<int32_t>(0ULL));
break;}
}}
if (taken) {
allegrex_CPU_doBranch(c,true,target);
}
else {
c->nullifyNext = true;
}
}
}
// tools/cpu/allegrex/exec.go:198:1
void allegrex_CPU_special2(allegrex_CPU* c,uint32_t w,uint32_t rs,uint32_t rt,uint32_t rd){
{
{
switch(cast<uint32_t>((w & cast<uint32_t>(63ULL)))){
case cast<uint32_t>(0ULL):{
int64_t p = cast<int64_t>((cast<int64_t>(allegrex_CPU_hilo(c)) + cast<int64_t>((cast<int64_t>(cast<int32_t>(allegrex_CPU_reg(c,rs))) * cast<int64_t>(cast<int32_t>(allegrex_CPU_reg(c,rt)))))));
allegrex_CPU_setHilo(c,cast<uint64_t>(p));
break;}
case cast<uint32_t>(1ULL):{
uint64_t p = cast<uint64_t>((allegrex_CPU_hilo(c) + cast<uint64_t>((cast<uint64_t>(allegrex_CPU_reg(c,rs)) * cast<uint64_t>(allegrex_CPU_reg(c,rt))))));
allegrex_CPU_setHilo(c,p);
break;}
case cast<uint32_t>(2ULL):{
allegrex_CPU_set(c,rd,cast<uint32_t>(cast<int32_t>((cast<int32_t>(allegrex_CPU_reg(c,rs)) * cast<int32_t>(allegrex_CPU_reg(c,rt))))));
break;}
case cast<uint32_t>(4ULL):{
int64_t p = cast<int64_t>((cast<int64_t>(allegrex_CPU_hilo(c)) - cast<int64_t>((cast<int64_t>(cast<int32_t>(allegrex_CPU_reg(c,rs))) * cast<int64_t>(cast<int32_t>(allegrex_CPU_reg(c,rt)))))));
allegrex_CPU_setHilo(c,cast<uint64_t>(p));
break;}
case cast<uint32_t>(5ULL):{
uint64_t p = cast<uint64_t>((allegrex_CPU_hilo(c) - cast<uint64_t>((cast<uint64_t>(allegrex_CPU_reg(c,rs)) * cast<uint64_t>(allegrex_CPU_reg(c,rt))))));
allegrex_CPU_setHilo(c,p);
break;}
case cast<uint32_t>(32ULL):{
allegrex_CPU_set(c,rd,allegrex_clz32(allegrex_CPU_reg(c,rs)));
break;}
case cast<uint32_t>(33ULL):{
allegrex_CPU_set(c,rd,allegrex_clz32(cast<uint32_t>(~allegrex_CPU_reg(c,rs))));
break;}
default:{
allegrex_CPU_Halt(c,std::string("unimplemented special2 funct 0x%02X at 0x%08X",45),cast<uint32_t>((w & cast<uint32_t>(63ULL))),c->curPC);
break;}
}}
}
}
// tools/cpu/allegrex/exec.go:225:1
void allegrex_CPU_special3(allegrex_CPU* c,uint32_t w,uint32_t rs,uint32_t rt,uint32_t rd,uint32_t shamt){
{
{
switch(cast<uint32_t>((w & cast<uint32_t>(63ULL)))){
case cast<uint32_t>(0ULL):{
auto tmp24 = std::make_tuple(shamt,cast<uint32_t>((rd + cast<uint32_t>(1ULL))));
uint32_t pos = std::get<0>(tmp24);
uint32_t size = std::get<1>(tmp24);
uint32_t mask = cast<uint32_t>(cast<uint64_t>(((shl<uint64_t>(cast<uint64_t>(1ULL),size)) - cast<uint64_t>(1ULL))));
allegrex_CPU_set(c,rt,cast<uint32_t>(((shr<uint32_t>(allegrex_CPU_reg(c,rs),pos)) & mask)));
break;}
case cast<uint32_t>(4ULL):{
auto tmp25 = std::make_tuple(shamt,rd);
uint32_t pos = std::get<0>(tmp25);
uint32_t msb = std::get<1>(tmp25);
uint32_t size = cast<uint32_t>((cast<uint32_t>((msb - pos)) + cast<uint32_t>(1ULL)));
uint32_t mask = shl<uint32_t>(cast<uint32_t>(cast<uint64_t>(((shl<uint64_t>(cast<uint64_t>(1ULL),size)) - cast<uint64_t>(1ULL)))),pos);
allegrex_CPU_set(c,rt,cast<uint32_t>((((allegrex_CPU_reg(c,rt) & ~(mask))) | (cast<uint32_t>(((shl<uint32_t>(allegrex_CPU_reg(c,rs),pos)) & mask))))));
break;}
case cast<uint32_t>(32ULL):{
{
switch(shamt){
case cast<uint32_t>(2ULL):{
uint32_t v = allegrex_CPU_reg(c,rt);
allegrex_CPU_set(c,rd,cast<uint32_t>((shl<uint32_t>((cast<uint32_t>((v & cast<uint32_t>(16711935ULL)))),cast<int64_t>(8ULL)) | shr<uint32_t>((cast<uint32_t>((v & cast<uint32_t>(4278255360ULL)))),cast<int64_t>(8ULL)))));
break;}
case cast<uint32_t>(16ULL):{
allegrex_CPU_set(c,rd,cast<uint32_t>(cast<int32_t>(cast<int8_t>(cast<uint8_t>(allegrex_CPU_reg(c,rt))))));
break;}
case cast<uint32_t>(24ULL):{
allegrex_CPU_set(c,rd,cast<uint32_t>(cast<int32_t>(cast<int16_t>(cast<uint16_t>(allegrex_CPU_reg(c,rt))))));
break;}
case cast<uint32_t>(20ULL):{
allegrex_CPU_set(c,rd,allegrex_bitrev32(allegrex_CPU_reg(c,rt)));
break;}
default:{
allegrex_CPU_Halt(c,std::string("unimplemented bshfl shamt 0x%02X at 0x%08X",42),shamt,c->curPC);
break;}
}}
break;}
default:{
allegrex_CPU_Halt(c,std::string("unimplemented special3 funct 0x%02X at 0x%08X",45),cast<uint32_t>((w & cast<uint32_t>(63ULL))),c->curPC);
break;}
}}
}
}
// tools/cpu/allegrex/exec.go:255:1
uint64_t allegrex_CPU_hilo(allegrex_CPU* c){
{
return cast<uint64_t>((shl<uint64_t>(cast<uint64_t>(c->HI),cast<int64_t>(32ULL)) | cast<uint64_t>(c->LO)));
}
}
// tools/cpu/allegrex/exec.go:256:1
void allegrex_CPU_setHilo(allegrex_CPU* c,uint64_t v){
{
auto tmp26 = std::make_tuple(cast<uint32_t>(shr<uint64_t>(v,cast<int64_t>(32ULL))),cast<uint32_t>(v));
c->HI = std::get<0>(tmp26);
c->LO = std::get<1>(tmp26);
}
}
// tools/cpu/allegrex/exec.go:258:1
uint32_t allegrex_clz32(uint32_t v){
{
uint32_t n = cast<uint32_t>(0ULL);
{int64_t i = cast<int64_t>(31ULL);for (;(i >= cast<int64_t>(0ULL));i--){
if ((cast<uint32_t>((v & (shl<uint32_t>(cast<uint32_t>(1ULL),cast<uint64_t>(i))))) != cast<uint32_t>(0ULL))) {
break;
}
n++;
}
}return n;
}
}
// tools/cpu/allegrex/exec.go:269:1
uint32_t allegrex_bitrev32(uint32_t v){
{
uint32_t r={};
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(32ULL));i++){
r = cast<uint32_t>(((shl<uint32_t>(r,cast<int64_t>(1ULL))) | (cast<uint32_t>((v & cast<uint32_t>(1ULL))))));
v = shr<uint32_t>(v,cast<int64_t>(1ULL));
}
}return r;
}
}
// tools/cpu/allegrex/exec.go:278:1
void allegrex_CPU_special(allegrex_CPU* c,uint32_t w,uint32_t rs,uint32_t rt,uint32_t rd,uint32_t shamt){
{
uint32_t funct = cast<uint32_t>((w & cast<uint32_t>(63ULL)));
{
switch(funct){
case cast<uint32_t>(0ULL):{
allegrex_CPU_set(c,rd,shl<uint32_t>(allegrex_CPU_reg(c,rt),shamt));
break;}
case cast<uint32_t>(2ULL):{
if ((cast<uint32_t>((rs & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL))) {
allegrex_CPU_set(c,rd,allegrex_rotr32(allegrex_CPU_reg(c,rt),shamt));
}
else {
allegrex_CPU_set(c,rd,shr<uint32_t>(allegrex_CPU_reg(c,rt),shamt));
}
break;}
case cast<uint32_t>(3ULL):{
allegrex_CPU_set(c,rd,cast<uint32_t>(shr<int32_t>(cast<int32_t>(allegrex_CPU_reg(c,rt)),shamt)));
break;}
case cast<uint32_t>(4ULL):{
allegrex_CPU_set(c,rd,shl<uint32_t>(allegrex_CPU_reg(c,rt),(cast<uint32_t>((allegrex_CPU_reg(c,rs) & cast<uint32_t>(31ULL))))));
break;}
case cast<uint32_t>(6ULL):{
if ((cast<uint32_t>((shamt & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL))) {
allegrex_CPU_set(c,rd,allegrex_rotr32(allegrex_CPU_reg(c,rt),cast<uint32_t>((allegrex_CPU_reg(c,rs) & cast<uint32_t>(31ULL)))));
}
else {
allegrex_CPU_set(c,rd,shr<uint32_t>(allegrex_CPU_reg(c,rt),(cast<uint32_t>((allegrex_CPU_reg(c,rs) & cast<uint32_t>(31ULL))))));
}
break;}
case cast<uint32_t>(7ULL):{
allegrex_CPU_set(c,rd,cast<uint32_t>(shr<int32_t>(cast<int32_t>(allegrex_CPU_reg(c,rt)),(cast<uint32_t>((allegrex_CPU_reg(c,rs) & cast<uint32_t>(31ULL)))))));
break;}
case cast<uint32_t>(8ULL):{
allegrex_CPU_doBranch(c,true,allegrex_CPU_reg(c,rs));
break;}
case cast<uint32_t>(9ULL):{
allegrex_CPU_set(c,rd,cast<uint32_t>((c->curPC + cast<uint32_t>(8ULL))));
allegrex_CPU_doBranch(c,true,allegrex_CPU_reg(c,rs));
break;}
case cast<uint32_t>(10ULL):{
if ((allegrex_CPU_reg(c,rt) == cast<uint32_t>(0ULL))) {
allegrex_CPU_set(c,rd,allegrex_CPU_reg(c,rs));
}
break;}
case cast<uint32_t>(11ULL):{
if ((allegrex_CPU_reg(c,rt) != cast<uint32_t>(0ULL))) {
allegrex_CPU_set(c,rd,allegrex_CPU_reg(c,rs));
}
break;}
case cast<uint32_t>(12ULL):{
if ((bool(c->Syscall) && c->Syscall(c,cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(6ULL))) & cast<uint32_t>(1048575ULL)))))) {
return ;
}
allegrex_CPU_Exception(c,cast<uint32_t>(8ULL));
break;}
case cast<uint32_t>(15ULL):{
return ;
break;}
case cast<uint32_t>(13ULL):{
allegrex_CPU_Exception(c,cast<uint32_t>(9ULL));
break;}
case cast<uint32_t>(16ULL):{
allegrex_CPU_set(c,rd,c->HI);
break;}
case cast<uint32_t>(17ULL):{
c->HI = allegrex_CPU_reg(c,rs);
break;}
case cast<uint32_t>(18ULL):{
allegrex_CPU_set(c,rd,c->LO);
break;}
case cast<uint32_t>(19ULL):{
c->LO = allegrex_CPU_reg(c,rs);
break;}
case cast<uint32_t>(24ULL):{
int64_t p = cast<int64_t>((cast<int64_t>(cast<int32_t>(allegrex_CPU_reg(c,rs))) * cast<int64_t>(cast<int32_t>(allegrex_CPU_reg(c,rt)))));
auto tmp27 = std::make_tuple(cast<uint32_t>(p),cast<uint32_t>(shr<uint64_t>(cast<uint64_t>(p),cast<int64_t>(32ULL))));
c->LO = std::get<0>(tmp27);
c->HI = std::get<1>(tmp27);
break;}
case cast<uint32_t>(25ULL):{
uint64_t p = cast<uint64_t>((cast<uint64_t>(allegrex_CPU_reg(c,rs)) * cast<uint64_t>(allegrex_CPU_reg(c,rt))));
auto tmp28 = std::make_tuple(cast<uint32_t>(p),cast<uint32_t>(shr<uint64_t>(p,cast<int64_t>(32ULL))));
c->LO = std::get<0>(tmp28);
c->HI = std::get<1>(tmp28);
break;}
case cast<uint32_t>(26ULL):{
allegrex_CPU_divSigned(c,cast<int32_t>(allegrex_CPU_reg(c,rs)),cast<int32_t>(allegrex_CPU_reg(c,rt)));
break;}
case cast<uint32_t>(27ULL):{
allegrex_CPU_divUnsigned(c,allegrex_CPU_reg(c,rs),allegrex_CPU_reg(c,rt));
break;}
case cast<uint32_t>(32ULL):{
{
auto tmp29 = allegrex_addOv(allegrex_CPU_reg(c,rs),allegrex_CPU_reg(c,rt));
uint32_t r = std::get<0>(tmp29);
bool ov = std::get<1>(tmp29);
if (ov) {
allegrex_CPU_Exception(c,cast<uint32_t>(12ULL));
}
else {
allegrex_CPU_set(c,rd,r);
}
}
break;}
case cast<uint32_t>(33ULL):{
allegrex_CPU_set(c,rd,cast<uint32_t>((allegrex_CPU_reg(c,rs) + allegrex_CPU_reg(c,rt))));
break;}
case cast<uint32_t>(34ULL):{
{
auto tmp30 = allegrex_subOv(allegrex_CPU_reg(c,rs),allegrex_CPU_reg(c,rt));
uint32_t r = std::get<0>(tmp30);
bool ov = std::get<1>(tmp30);
if (ov) {
allegrex_CPU_Exception(c,cast<uint32_t>(12ULL));
}
else {
allegrex_CPU_set(c,rd,r);
}
}
break;}
case cast<uint32_t>(35ULL):{
allegrex_CPU_set(c,rd,cast<uint32_t>((allegrex_CPU_reg(c,rs) - allegrex_CPU_reg(c,rt))));
break;}
case cast<uint32_t>(36ULL):{
allegrex_CPU_set(c,rd,cast<uint32_t>((allegrex_CPU_reg(c,rs) & allegrex_CPU_reg(c,rt))));
break;}
case cast<uint32_t>(37ULL):{
allegrex_CPU_set(c,rd,cast<uint32_t>((allegrex_CPU_reg(c,rs) | allegrex_CPU_reg(c,rt))));
break;}
case cast<uint32_t>(38ULL):{
allegrex_CPU_set(c,rd,cast<uint32_t>((allegrex_CPU_reg(c,rs) ^ allegrex_CPU_reg(c,rt))));
break;}
case cast<uint32_t>(39ULL):{
allegrex_CPU_set(c,rd,cast<uint32_t>(~(cast<uint32_t>((allegrex_CPU_reg(c,rs) | allegrex_CPU_reg(c,rt))))));
break;}
case cast<uint32_t>(42ULL):{
allegrex_CPU_set(c,rd,allegrex_b2u((cast<int32_t>(allegrex_CPU_reg(c,rs)) < cast<int32_t>(allegrex_CPU_reg(c,rt)))));
break;}
case cast<uint32_t>(43ULL):{
allegrex_CPU_set(c,rd,allegrex_b2u((allegrex_CPU_reg(c,rs) < allegrex_CPU_reg(c,rt))));
break;}
case cast<uint32_t>(22ULL):{
allegrex_CPU_set(c,rd,allegrex_clz32(allegrex_CPU_reg(c,rs)));
break;}
case cast<uint32_t>(23ULL):{
allegrex_CPU_set(c,rd,allegrex_clz32(cast<uint32_t>(~allegrex_CPU_reg(c,rs))));
break;}
case cast<uint32_t>(44ULL):{
if ((cast<int32_t>(allegrex_CPU_reg(c,rs)) > cast<int32_t>(allegrex_CPU_reg(c,rt)))) {
allegrex_CPU_set(c,rd,allegrex_CPU_reg(c,rs));
}
else {
allegrex_CPU_set(c,rd,allegrex_CPU_reg(c,rt));
}
break;}
case cast<uint32_t>(45ULL):{
if ((cast<int32_t>(allegrex_CPU_reg(c,rs)) < cast<int32_t>(allegrex_CPU_reg(c,rt)))) {
allegrex_CPU_set(c,rd,allegrex_CPU_reg(c,rs));
}
else {
allegrex_CPU_set(c,rd,allegrex_CPU_reg(c,rt));
}
break;}
default:{
allegrex_CPU_Halt(c,std::string("unimplemented special funct 0x%02X (word 0x%08X) at 0x%08X",58),funct,w,c->curPC);
break;}
}}
}
}
// tools/cpu/allegrex/exec.go:392:1
void allegrex_CPU_divSigned(allegrex_CPU* c,int32_t a,int32_t b){
{
{
if ((b == cast<int32_t>(0ULL))){
c->HI = cast<uint32_t>(a);
if ((a >= cast<int32_t>(0ULL))) {
c->LO = cast<uint32_t>(4294967295ULL);
}
else {
c->LO = cast<uint32_t>(1ULL);
}
}
else if (((a == cast<int32_t>(-2147483648ULL)) && (b == cast<int32_t>(-1ULL)))){
auto tmp32 = std::make_tuple(cast<uint32_t>(2147483648ULL),cast<uint32_t>(0ULL));
c->LO = std::get<0>(tmp32);
c->HI = std::get<1>(tmp32);
}
else {
auto tmp33 = std::make_tuple(cast<uint32_t>(divi<int32_t>(a,b)),cast<uint32_t>(modi<int32_t>(a,b)));
c->LO = std::get<0>(tmp33);
c->HI = std::get<1>(tmp33);
}
}
tmp31:;
}
}
// tools/cpu/allegrex/exec.go:408:1
void allegrex_CPU_divUnsigned(allegrex_CPU* c,uint32_t a,uint32_t b){
{
if ((b == cast<uint32_t>(0ULL))) {
auto tmp34 = std::make_tuple(cast<uint32_t>(4294967295ULL),a);
c->LO = std::get<0>(tmp34);
c->HI = std::get<1>(tmp34);
return ;
}
auto tmp35 = std::make_tuple(divi<uint32_t>(a,b),modi<uint32_t>(a,b));
c->LO = std::get<0>(tmp35);
c->HI = std::get<1>(tmp35);
}
}
// tools/cpu/allegrex/exec.go:416:1
void allegrex_CPU_loadOp(allegrex_CPU* c,uint32_t op,uint32_t rs,uint32_t rt,uint32_t simm){
{
uint32_t addr = cast<uint32_t>((allegrex_CPU_reg(c,rs) + simm));
{
switch(op){
case cast<uint32_t>(32ULL):{
allegrex_CPU_load(c,rt,cast<uint32_t>(cast<int32_t>(cast<int8_t>(cast<uint8_t>(allegrex_CPU_read8(c,addr))))));
break;}
case cast<uint32_t>(36ULL):{
allegrex_CPU_load(c,rt,allegrex_CPU_read8(c,addr));
break;}
case cast<uint32_t>(33ULL):{
if ((cast<uint32_t>((addr & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL))) {
allegrex_CPU_addrError(c,cast<uint32_t>(4ULL),addr);
return ;
}
allegrex_CPU_load(c,rt,cast<uint32_t>(cast<int32_t>(cast<int16_t>(cast<uint16_t>(allegrex_CPU_read16(c,addr))))));
break;}
case cast<uint32_t>(37ULL):{
if ((cast<uint32_t>((addr & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL))) {
allegrex_CPU_addrError(c,cast<uint32_t>(4ULL),addr);
return ;
}
allegrex_CPU_load(c,rt,allegrex_CPU_read16(c,addr));
break;}
case cast<uint32_t>(35ULL):{
if ((cast<uint32_t>((addr & cast<uint32_t>(3ULL))) != cast<uint32_t>(0ULL))) {
allegrex_CPU_addrError(c,cast<uint32_t>(4ULL),addr);
return ;
}
allegrex_CPU_load(c,rt,allegrex_CPU_read32(c,addr));
break;}
case cast<uint32_t>(34ULL):{
uint32_t word = allegrex_CPU_read32(c,(addr & ~(cast<uint32_t>(3ULL))));
uint32_t shift = cast<uint32_t>(((cast<uint32_t>((addr & cast<uint32_t>(3ULL)))) * cast<uint32_t>(8ULL)));
uint32_t cur = c->out[rt];
allegrex_CPU_load(c,rt,cast<uint32_t>(((cast<uint32_t>((cur & (shr<uint32_t>(cast<uint32_t>(16777215ULL),shift))))) | (shl<uint32_t>(word,(cast<uint32_t>((cast<uint32_t>(24ULL) - shift))))))));
break;}
case cast<uint32_t>(38ULL):{
uint32_t word = allegrex_CPU_read32(c,(addr & ~(cast<uint32_t>(3ULL))));
uint32_t shift = cast<uint32_t>(((cast<uint32_t>((addr & cast<uint32_t>(3ULL)))) * cast<uint32_t>(8ULL)));
uint32_t cur = c->out[rt];
allegrex_CPU_load(c,rt,cast<uint32_t>(((cast<uint32_t>((cur & (shl<uint32_t>(cast<uint32_t>(4294967040ULL),(cast<uint32_t>((cast<uint32_t>(24ULL) - shift)))))))) | (shr<uint32_t>(word,shift)))));
break;}
}}
}
}
// tools/cpu/allegrex/exec.go:454:1
void allegrex_CPU_storeOp(allegrex_CPU* c,uint32_t op,uint32_t rs,uint32_t rt,uint32_t simm){
{
uint32_t addr = cast<uint32_t>((allegrex_CPU_reg(c,rs) + simm));
uint32_t v = allegrex_CPU_reg(c,rt);
{
switch(op){
case cast<uint32_t>(40ULL):{
allegrex_CPU_write8(c,addr,cast<uint32_t>((v & cast<uint32_t>(255ULL))));
break;}
case cast<uint32_t>(41ULL):{
if ((cast<uint32_t>((addr & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL))) {
allegrex_CPU_addrError(c,cast<uint32_t>(5ULL),addr);
return ;
}
allegrex_CPU_write16(c,addr,cast<uint32_t>((v & cast<uint32_t>(65535ULL))));
break;}
case cast<uint32_t>(43ULL):{
if ((cast<uint32_t>((addr & cast<uint32_t>(3ULL))) != cast<uint32_t>(0ULL))) {
allegrex_CPU_addrError(c,cast<uint32_t>(5ULL),addr);
return ;
}
allegrex_CPU_write32(c,addr,v);
break;}
case cast<uint32_t>(42ULL):{
uint32_t word = allegrex_CPU_read32(c,(addr & ~(cast<uint32_t>(3ULL))));
uint32_t shift = cast<uint32_t>(((cast<uint32_t>((addr & cast<uint32_t>(3ULL)))) * cast<uint32_t>(8ULL)));
allegrex_CPU_write32(c,(addr & ~(cast<uint32_t>(3ULL))),cast<uint32_t>(((cast<uint32_t>((word & (shl<uint32_t>(cast<uint32_t>(4294967040ULL),shift))))) | (shr<uint32_t>(v,(cast<uint32_t>((cast<uint32_t>(24ULL) - shift))))))));
break;}
case cast<uint32_t>(46ULL):{
uint32_t word = allegrex_CPU_read32(c,(addr & ~(cast<uint32_t>(3ULL))));
uint32_t shift = cast<uint32_t>(((cast<uint32_t>((addr & cast<uint32_t>(3ULL)))) * cast<uint32_t>(8ULL)));
allegrex_CPU_write32(c,(addr & ~(cast<uint32_t>(3ULL))),cast<uint32_t>(((cast<uint32_t>((word & (shr<uint32_t>(cast<uint32_t>(16777215ULL),(cast<uint32_t>((cast<uint32_t>(24ULL) - shift)))))))) | (shl<uint32_t>(v,shift)))));
break;}
}}
}
}
// tools/cpu/allegrex/exec.go:483:1
void allegrex_CPU_cop0(allegrex_CPU* c,uint32_t w,uint32_t rs,uint32_t rt,uint32_t rd){
{
if ((cast<uint32_t>((w & cast<uint32_t>(33554432ULL))) != cast<uint32_t>(0ULL))) {
if ((cast<uint32_t>((w & cast<uint32_t>(63ULL))) == cast<uint32_t>(16ULL))) {
allegrex_CPU_rfe(c);
return ;
}
allegrex_CPU_Halt(c,std::string("unimplemented cop0 command 0x%08X at 0x%08X",43),w,c->curPC);
return ;
}
{
switch(rs){
case cast<uint32_t>(0ULL):{
allegrex_CPU_load(c,rt,c->COP0[rd]);
break;}
case cast<uint32_t>(4ULL):{
c->COP0[rd] = allegrex_CPU_reg(c,rt);
break;}
default:{
allegrex_CPU_Halt(c,std::string("unimplemented cop0 rs=0x%02X (word 0x%08X) at 0x%08X",52),rs,w,c->curPC);
break;}
}}
}
}
// tools/cpu/allegrex/exec.go:502:1
uint32_t allegrex_rotr32(uint32_t v,uint32_t n){
{
n &= cast<uint32_t>(31ULL);
if ((n == cast<uint32_t>(0ULL))) {
return v;
}
return cast<uint32_t>((shr<uint32_t>(v,n) | shl<uint32_t>(v,(cast<uint32_t>((cast<uint32_t>(32ULL) - n))))));
}
}
// tools/cpu/allegrex/exec.go:511:1
std::tuple<uint32_t,bool> allegrex_addOv(uint32_t a,uint32_t b){
{
uint32_t r = cast<uint32_t>((a + b));
return {r,(cast<uint32_t>((cast<uint32_t>(((cast<uint32_t>((a ^ r))) & (cast<uint32_t>((b ^ r))))) & cast<uint32_t>(2147483648ULL))) != cast<uint32_t>(0ULL))};
}
}
// tools/cpu/allegrex/exec.go:515:1
std::tuple<uint32_t,bool> allegrex_subOv(uint32_t a,uint32_t b){
{
uint32_t r = cast<uint32_t>((a - b));
return {r,(cast<uint32_t>((cast<uint32_t>(((cast<uint32_t>((a ^ b))) & (cast<uint32_t>((a ^ r))))) & cast<uint32_t>(2147483648ULL))) != cast<uint32_t>(0ULL))};
}
}
// tools/cpu/allegrex/exec.go:520:1
uint32_t allegrex_b2u(bool b){
{
if (b) {
return cast<uint32_t>(1ULL);
}
return cast<uint32_t>(0ULL);
}
}
// tools/cpu/allegrex/fpu.go:14:1
std::string allegrex_fpr(uint32_t i){
{
return go_fmt_Sprintf(std::string("$f%d",4),cast<uint32_t>((i & cast<uint32_t>(31ULL))));
}
}
// tools/cpu/allegrex/fpu.go:17:1
allegrex_Inst allegrex_decodeCop1(allegrex_Inst in,uint32_t w,uint32_t rs,uint32_t rt,uint32_t rd,uint32_t shamt){
{
auto set = [&](std::string mnem,std::string text)->allegrex_Inst{
auto tmp36 = std::make_tuple(mnem,text);
in.Mnem = std::get<0>(tmp36);
in.Text = std::get<1>(tmp36);
return in;
}
;
auto tmp37 = std::make_tuple(rd,shamt);
uint32_t fs = std::get<0>(tmp37);
uint32_t fd = std::get<1>(tmp37);
{
switch(rs){
case cast<uint32_t>(0ULL):{
return set(std::string("mfc1",4),go_fmt_Sprintf(std::string("mfc1 %s, %s",11),allegrex_reg(rt),allegrex_fpr(fs)));
break;}
case cast<uint32_t>(2ULL):{
return set(std::string("cfc1",4),go_fmt_Sprintf(std::string("cfc1 %s, $%d",12),allegrex_reg(rt),fs));
break;}
case cast<uint32_t>(4ULL):{
return set(std::string("mtc1",4),go_fmt_Sprintf(std::string("mtc1 %s, %s",11),allegrex_reg(rt),allegrex_fpr(fs)));
break;}
case cast<uint32_t>(6ULL):{
return set(std::string("ctc1",4),go_fmt_Sprintf(std::string("ctc1 %s, $%d",12),allegrex_reg(rt),fs));
break;}
case cast<uint32_t>(8ULL):{
auto tmp38 = std::make_tuple(cast<allegrex_Flow>(1ULL),cast<uint32_t>((cast<uint32_t>((in.Addr + cast<uint32_t>(4ULL))) + cast<uint32_t>((cast<uint32_t>(cast<int32_t>(cast<int16_t>(w))) * cast<uint32_t>(4ULL))))),true,true);
in.Flow = std::get<0>(tmp38);
in.Target = std::get<1>(tmp38);
in.HasTarget = std::get<2>(tmp38);
in.HasDelay = std::get<3>(tmp38);
std::string mnem = std::array<std::string,4>{std::string("bc1f",4),std::string("bc1t",4),std::string("bc1fl",5),std::string("bc1tl",5)}[cast<uint32_t>((rt & cast<uint32_t>(3ULL)))];
return set(mnem,go_fmt_Sprintf(std::string("%s $%08X",8),mnem,in.Target));
break;}
case cast<uint32_t>(16ULL):case cast<uint32_t>(20ULL):{
return allegrex_decodeCop1Fmt(in,w,rs,rt,fs,fd);
break;}
}}
return allegrex_word(in,w);
}
}
// tools/cpu/allegrex/fpu.go:39:1
allegrex_Inst allegrex_decodeCop1Fmt(allegrex_Inst in,uint32_t w,uint32_t rs,uint32_t ft,uint32_t fs,uint32_t fd){
{
auto set = [&](std::string mnem,std::string text)->allegrex_Inst{
auto tmp39 = std::make_tuple(mnem,text);
in.Mnem = std::get<0>(tmp39);
in.Text = std::get<1>(tmp39);
return in;
}
;
uint32_t funct = cast<uint32_t>((w & cast<uint32_t>(63ULL)));
std::string sfx = std::string("s",1);
if ((rs == cast<uint32_t>(20ULL))) {
sfx = std::string("w",1);
}
auto tri = [&](std::string name)->allegrex_Inst{
return set(((name + std::string(".",1)) + sfx),go_fmt_Sprintf(std::string("%s.%s %s, %s, %s",16),name,sfx,allegrex_fpr(fd),allegrex_fpr(fs),allegrex_fpr(ft)));
}
;
auto bin = [&](std::string name)->allegrex_Inst{
return set(((name + std::string(".",1)) + sfx),go_fmt_Sprintf(std::string("%s.%s %s, %s",12),name,sfx,allegrex_fpr(fd),allegrex_fpr(fs)));
}
;
{
switch(funct){
case cast<uint32_t>(0ULL):{
return tri(std::string("add",3));
break;}
case cast<uint32_t>(1ULL):{
return tri(std::string("sub",3));
break;}
case cast<uint32_t>(2ULL):{
return tri(std::string("mul",3));
break;}
case cast<uint32_t>(3ULL):{
return tri(std::string("div",3));
break;}
case cast<uint32_t>(4ULL):{
return bin(std::string("sqrt",4));
break;}
case cast<uint32_t>(5ULL):{
return bin(std::string("abs",3));
break;}
case cast<uint32_t>(6ULL):{
return bin(std::string("mov",3));
break;}
case cast<uint32_t>(7ULL):{
return bin(std::string("neg",3));
break;}
case cast<uint32_t>(12ULL):{
return bin(std::string("round.w",7));
break;}
case cast<uint32_t>(13ULL):{
return bin(std::string("trunc.w",7));
break;}
case cast<uint32_t>(14ULL):{
return bin(std::string("ceil.w",6));
break;}
case cast<uint32_t>(15ULL):{
return bin(std::string("floor.w",7));
break;}
case cast<uint32_t>(32ULL):{
return set(std::string("cvt.s.w",7),go_fmt_Sprintf(std::string("cvt.s.w %s, %s",14),allegrex_fpr(fd),allegrex_fpr(fs)));
break;}
case cast<uint32_t>(36ULL):{
return set(std::string("cvt.w.s",7),go_fmt_Sprintf(std::string("cvt.w.s %s, %s",14),allegrex_fpr(fd),allegrex_fpr(fs)));
break;}
}}
if ((funct >= cast<uint32_t>(48ULL))) {
std::string cond = std::array<std::string,16>{std::string("f",1),std::string("un",2),std::string("eq",2),std::string("ueq",3),std::string("olt",3),std::string("ult",3),std::string("ole",3),std::string("ule",3),std::string("sf",2),std::string("ngle",4),std::string("seq",3),std::string("ngl",3),std::string("lt",2),std::string("nge",3),std::string("le",2),std::string("ngt",3)}[cast<uint32_t>((funct & cast<uint32_t>(15ULL)))];
std::string mnem = (((std::string("c.",2) + cond) + std::string(".",1)) + sfx);
return set(mnem,go_fmt_Sprintf(std::string("%s %s, %s",9),mnem,allegrex_fpr(fs),allegrex_fpr(ft)));
}
return allegrex_word(in,w);
}
}
// tools/cpu/allegrex/fpu.go:93:1
float allegrex_CPU_ff(allegrex_CPU* c,uint32_t i){
{
return go_math_Float32frombits(c->F[cast<uint32_t>((i & cast<uint32_t>(31ULL)))]);
}
}
// tools/cpu/allegrex/fpu.go:94:1
void allegrex_CPU_setf(allegrex_CPU* c,uint32_t i,float v){
{
c->F[cast<uint32_t>((i & cast<uint32_t>(31ULL)))] = go_math_Float32bits(v);
}
}
// tools/cpu/allegrex/fpu.go:97:1
void allegrex_CPU_cop1(allegrex_CPU* c,uint32_t w,uint32_t rs,uint32_t rt,uint32_t rd,uint32_t shamt){
{
auto tmp40 = std::make_tuple(rd,rt,shamt);
uint32_t fs = std::get<0>(tmp40);
uint32_t ft = std::get<1>(tmp40);
uint32_t fd = std::get<2>(tmp40);
{
switch(rs){
case cast<uint32_t>(0ULL):{
allegrex_CPU_load(c,rt,c->F[cast<uint32_t>((fs & cast<uint32_t>(31ULL)))]);
break;}
case cast<uint32_t>(2ULL):{
uint32_t v={};
if (((fs == cast<uint32_t>(31ULL)) && c->FCC)) {
v = cast<uint32_t>(8388608ULL);
}
allegrex_CPU_load(c,rt,v);
break;}
case cast<uint32_t>(4ULL):{
c->F[cast<uint32_t>((fs & cast<uint32_t>(31ULL)))] = allegrex_CPU_reg(c,rt);
break;}
case cast<uint32_t>(6ULL):{
if ((fs == cast<uint32_t>(31ULL))) {
c->FCC = (cast<uint32_t>((allegrex_CPU_reg(c,rt) & cast<uint32_t>(8388608ULL))) != cast<uint32_t>(0ULL));
}
break;}
case cast<uint32_t>(8ULL):{
uint32_t simm = cast<uint32_t>(cast<int32_t>(cast<int16_t>(w)));
uint32_t target = cast<uint32_t>((cast<uint32_t>((c->curPC + cast<uint32_t>(4ULL))) + shl<uint32_t>(simm,cast<int64_t>(2ULL))));
bool taken = (((cast<uint32_t>((rt & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL))) == c->FCC);
if (((cast<uint32_t>((rt & cast<uint32_t>(2ULL))) != cast<uint32_t>(0ULL)) && (!taken))) {
c->nullifyNext = true;
}
else {
allegrex_CPU_doBranch(c,taken,target);
}
break;}
case cast<uint32_t>(16ULL):{
allegrex_CPU_cop1S(c,w,ft,fs,fd);
break;}
case cast<uint32_t>(20ULL):{
allegrex_CPU_cop1W(c,w,ft,fs,fd);
break;}
default:{
allegrex_CPU_Halt(c,std::string("unimplemented cop1 rs=0x%02X (word 0x%08X) at 0x%08X",52),rs,w,c->curPC);
break;}
}}
}
}
// tools/cpu/allegrex/fpu.go:132:1
void allegrex_CPU_cop1S(allegrex_CPU* c,uint32_t w,uint32_t ft,uint32_t fs,uint32_t fd){
{
uint32_t funct = cast<uint32_t>((w & cast<uint32_t>(63ULL)));
{
switch(funct){
case cast<uint32_t>(0ULL):{
allegrex_CPU_setf(c,fd,(allegrex_CPU_ff(c,fs) + allegrex_CPU_ff(c,ft)));
break;}
case cast<uint32_t>(1ULL):{
allegrex_CPU_setf(c,fd,(allegrex_CPU_ff(c,fs) - allegrex_CPU_ff(c,ft)));
break;}
case cast<uint32_t>(2ULL):{
allegrex_CPU_setf(c,fd,(allegrex_CPU_ff(c,fs) * allegrex_CPU_ff(c,ft)));
break;}
case cast<uint32_t>(3ULL):{
allegrex_CPU_setf(c,fd,(allegrex_CPU_ff(c,fs) / allegrex_CPU_ff(c,ft)));
break;}
case cast<uint32_t>(4ULL):{
allegrex_CPU_setf(c,fd,cast<float>(go_math_Sqrt(cast<double>(allegrex_CPU_ff(c,fs)))));
break;}
case cast<uint32_t>(5ULL):{
allegrex_CPU_setf(c,fd,cast<float>(go_math_Abs(cast<double>(allegrex_CPU_ff(c,fs)))));
break;}
case cast<uint32_t>(6ULL):{
c->F[cast<uint32_t>((fd & cast<uint32_t>(31ULL)))] = c->F[cast<uint32_t>((fs & cast<uint32_t>(31ULL)))];
break;}
case cast<uint32_t>(7ULL):{
allegrex_CPU_setf(c,fd,cast<float>(-allegrex_CPU_ff(c,fs)));
break;}
case cast<uint32_t>(12ULL):{
c->F[cast<uint32_t>((fd & cast<uint32_t>(31ULL)))] = cast<uint32_t>(cast<int32_t>(go_math_RoundToEven(cast<double>(allegrex_CPU_ff(c,fs)))));
break;}
case cast<uint32_t>(13ULL):{
c->F[cast<uint32_t>((fd & cast<uint32_t>(31ULL)))] = cast<uint32_t>(cast<int32_t>(go_math_Trunc(cast<double>(allegrex_CPU_ff(c,fs)))));
break;}
case cast<uint32_t>(14ULL):{
c->F[cast<uint32_t>((fd & cast<uint32_t>(31ULL)))] = cast<uint32_t>(cast<int32_t>(go_math_Ceil(cast<double>(allegrex_CPU_ff(c,fs)))));
break;}
case cast<uint32_t>(15ULL):{
c->F[cast<uint32_t>((fd & cast<uint32_t>(31ULL)))] = cast<uint32_t>(cast<int32_t>(go_math_Floor(cast<double>(allegrex_CPU_ff(c,fs)))));
break;}
case cast<uint32_t>(36ULL):{
c->F[cast<uint32_t>((fd & cast<uint32_t>(31ULL)))] = cast<uint32_t>(cast<int32_t>(go_math_RoundToEven(cast<double>(allegrex_CPU_ff(c,fs)))));
break;}
default:{
if ((funct >= cast<uint32_t>(48ULL))) {
c->FCC = allegrex_fcompare(funct,allegrex_CPU_ff(c,fs),allegrex_CPU_ff(c,ft));
return ;
}
allegrex_CPU_Halt(c,std::string("unimplemented cop1.s funct 0x%02X at 0x%08X",43),funct,c->curPC);
break;}
}}
}
}
// tools/cpu/allegrex/fpu.go:170:1
void allegrex_CPU_cop1W(allegrex_CPU* c,uint32_t w,uint32_t ft,uint32_t fs,uint32_t fd){
{
uint32_t funct = cast<uint32_t>((w & cast<uint32_t>(63ULL)));
{
switch(funct){
case cast<uint32_t>(32ULL):{
allegrex_CPU_setf(c,fd,cast<float>(cast<int32_t>(c->F[cast<uint32_t>((fs & cast<uint32_t>(31ULL)))])));
break;}
default:{
allegrex_CPU_Halt(c,std::string("unimplemented cop1.w funct 0x%02X at 0x%08X",43),funct,c->curPC);
break;}
}}
}
}
// tools/cpu/allegrex/fpu.go:184:1
bool allegrex_fcompare(uint32_t funct,float a,float b){
{
uint32_t cond = cast<uint32_t>((funct & cast<uint32_t>(15ULL)));
if (((a != a) || (b != b))) {
return (cast<uint32_t>((cond & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL));
}
if (((cast<uint32_t>((cond & cast<uint32_t>(4ULL))) != cast<uint32_t>(0ULL)) && (a < b))) {
return true;
}
return ((cast<uint32_t>((cond & cast<uint32_t>(2ULL))) != cast<uint32_t>(0ULL)) && (a == b));
}
}
// tools/cpu/allegrex/state.go:32:1
allegrex_CPUState allegrex_CPU_SaveState(allegrex_CPU* c){
{
return allegrex_CPUState{c->R,c->out,c->HI,c->LO,c->PC,c->nextPC,c->COP0,c->F,c->FCC,c->V,c->VfpuCtrl,c->Halted,c->HaltReason,c->Steps,c->curPC,c->ld.reg,c->ld.val,c->delaySlot,c->pendingDelay,c->branchAddr,c->nullifyNext};
}
}
// tools/cpu/allegrex/state.go:45:1
void allegrex_CPU_LoadState(allegrex_CPU* c,allegrex_CPUState s){
{
auto tmp41 = std::make_tuple(s.R,s.Out,s.HI,s.LO);
c->R = std::get<0>(tmp41);
c->out = std::get<1>(tmp41);
c->HI = std::get<2>(tmp41);
c->LO = std::get<3>(tmp41);
auto tmp42 = std::make_tuple(s.PC,s.NextPC,s.COP0);
c->PC = std::get<0>(tmp42);
c->nextPC = std::get<1>(tmp42);
c->COP0 = std::get<2>(tmp42);
auto tmp43 = std::make_tuple(s.F,s.FCC,s.V,s.VfpuCtrl);
c->F = std::get<0>(tmp43);
c->FCC = std::get<1>(tmp43);
c->V = std::get<2>(tmp43);
c->VfpuCtrl = std::get<3>(tmp43);
auto tmp44 = std::make_tuple(s.Halted,s.HaltReason,s.Steps);
c->Halted = std::get<0>(tmp44);
c->HaltReason = std::get<1>(tmp44);
c->Steps = std::get<2>(tmp44);
auto tmp45 = std::make_tuple(s.CurPC,allegrex_loadSlot{s.LdReg,s.LdVal});
c->curPC = std::get<0>(tmp45);
c->ld = std::get<1>(tmp45);
auto tmp46 = std::make_tuple(s.DelaySlot,s.PendingDelay,s.BranchAddr);
c->delaySlot = std::get<0>(tmp46);
c->pendingDelay = std::get<1>(tmp46);
c->branchAddr = std::get<2>(tmp46);
c->nullifyNext = s.NullifyNext;
}
}
// tools/cpu/allegrex/vfpu.go:68:1
uint32_t allegrex_vfpuSingle(uint32_t reg){
{
uint32_t m = cast<uint32_t>(((shr<uint32_t>(reg,cast<int64_t>(2ULL))) & cast<uint32_t>(7ULL)));
uint32_t c = cast<uint32_t>((reg & cast<uint32_t>(3ULL)));
uint32_t r = cast<uint32_t>(((shr<uint32_t>(reg,cast<int64_t>(5ULL))) & cast<uint32_t>(3ULL)));
return cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((m * cast<uint32_t>(16ULL))) + cast<uint32_t>((c * cast<uint32_t>(4ULL))))) + r));
}
}
// tools/cpu/allegrex/vfpu.go:76:1
std::array<uint32_t,4> allegrex_vecIdx(uint32_t reg,uint32_t n){
std::array<uint32_t,4> idx{};
{
if ((n == cast<uint32_t>(1ULL))) {
idx[cast<int64_t>(0ULL)] = allegrex_vfpuSingle(reg);
return idx;
}
uint32_t row={};
{
switch(n){
case cast<uint32_t>(2ULL):case cast<uint32_t>(4ULL):{
row = cast<uint32_t>(((shr<uint32_t>(reg,cast<int64_t>(5ULL))) & cast<uint32_t>(2ULL)));
break;}
case cast<uint32_t>(3ULL):{
row = cast<uint32_t>(((shr<uint32_t>(reg,cast<int64_t>(6ULL))) & cast<uint32_t>(1ULL)));
break;}
}}
uint32_t mtx = cast<uint32_t>(((shr<uint32_t>(reg,cast<int64_t>(2ULL))) & cast<uint32_t>(7ULL)));
uint32_t col = cast<uint32_t>((reg & cast<uint32_t>(3ULL)));
if ((cast<uint32_t>((reg & cast<uint32_t>(32ULL))) != cast<uint32_t>(0ULL))) {
uint32_t base = cast<uint32_t>((cast<uint32_t>((mtx * cast<uint32_t>(16ULL))) + col));
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < n);i++){
idx[i] = cast<uint32_t>((base + cast<uint32_t>(((cast<uint32_t>(((cast<uint32_t>((row + i))) & cast<uint32_t>(3ULL)))) * cast<uint32_t>(4ULL)))));
}
}}
else {
uint32_t base = cast<uint32_t>((cast<uint32_t>((mtx * cast<uint32_t>(16ULL))) + cast<uint32_t>((col * cast<uint32_t>(4ULL)))));
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < n);i++){
idx[i] = cast<uint32_t>((base + (cast<uint32_t>(((cast<uint32_t>((row + i))) & cast<uint32_t>(3ULL))))));
}
}}
return idx;
}
}
// tools/cpu/allegrex/vfpu.go:106:1
uint32_t allegrex_matIdx(uint32_t reg,uint32_t side,uint32_t a,uint32_t b){
{
uint32_t row={};
{
switch(side){
case cast<uint32_t>(2ULL):case cast<uint32_t>(4ULL):{
row = cast<uint32_t>(((shr<uint32_t>(reg,cast<int64_t>(5ULL))) & cast<uint32_t>(2ULL)));
break;}
case cast<uint32_t>(3ULL):{
row = cast<uint32_t>(((shr<uint32_t>(reg,cast<int64_t>(6ULL))) & cast<uint32_t>(1ULL)));
break;}
}}
uint32_t mtx = cast<uint32_t>(((shr<uint32_t>(reg,cast<int64_t>(2ULL))) & cast<uint32_t>(7ULL)));
uint32_t col = cast<uint32_t>((reg & cast<uint32_t>(3ULL)));
if ((cast<uint32_t>((reg & cast<uint32_t>(32ULL))) != cast<uint32_t>(0ULL))) {
return cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((mtx * cast<uint32_t>(16ULL))) + cast<uint32_t>(((cast<uint32_t>(((cast<uint32_t>((row + b))) & cast<uint32_t>(3ULL)))) * cast<uint32_t>(4ULL))))) + (cast<uint32_t>(((cast<uint32_t>((col + a))) & cast<uint32_t>(3ULL))))));
}
return cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((mtx * cast<uint32_t>(16ULL))) + cast<uint32_t>(((cast<uint32_t>(((cast<uint32_t>((col + a))) & cast<uint32_t>(3ULL)))) * cast<uint32_t>(4ULL))))) + (cast<uint32_t>(((cast<uint32_t>((row + b))) & cast<uint32_t>(3ULL))))));
}
}
// tools/cpu/allegrex/vfpu.go:122:1
std::array<float,4> allegrex_CPU_vread(allegrex_CPU* c,uint32_t reg,uint32_t n){
std::array<float,4> v{};
{
std::array<uint32_t,4> idx = allegrex_vecIdx(reg,n);
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < n);i++){
v[i] = go_math_Float32frombits(c->V[cast<uint32_t>((idx[i] & cast<uint32_t>(127ULL)))]);
}
}return v;
}
}
// tools/cpu/allegrex/vfpu.go:130:1
std::array<uint32_t,4> allegrex_CPU_vreadBits(allegrex_CPU* c,uint32_t reg,uint32_t n){
std::array<uint32_t,4> v{};
{
std::array<uint32_t,4> idx = allegrex_vecIdx(reg,n);
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < n);i++){
v[i] = c->V[cast<uint32_t>((idx[i] & cast<uint32_t>(127ULL)))];
}
}return v;
}
}
// tools/cpu/allegrex/vfpu.go:140:1
void allegrex_CPU_vwrite(allegrex_CPU* c,std::array<float,4> v,uint32_t reg,uint32_t n){
{
std::array<uint32_t,4> bits={};
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < n);i++){
bits[i] = go_math_Float32bits(v[i]);
}
}allegrex_CPU_vwriteBits(c,bits,reg,n);
}
}
// tools/cpu/allegrex/vfpu.go:148:1
void allegrex_CPU_vwriteBits(allegrex_CPU* c,std::array<uint32_t,4> v,uint32_t reg,uint32_t n){
{
uint32_t mask = shr<uint32_t>(c->VfpuCtrl[cast<int64_t>(2ULL)],cast<int64_t>(8ULL));
std::array<uint32_t,4> idx = allegrex_vecIdx(reg,n);
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < n);i++){
if ((cast<uint32_t>((mask & (shl<uint32_t>(cast<uint32_t>(1ULL),i)))) == cast<uint32_t>(0ULL))) {
c->V[cast<uint32_t>((idx[i] & cast<uint32_t>(127ULL)))] = v[i];
}
}
}}
}
// tools/cpu/allegrex/vfpu.go:161:1
std::array<float,16> allegrex_CPU_mread(allegrex_CPU* c,uint32_t reg,uint32_t side){
std::array<float,16> m{};
{
{uint32_t a = cast<uint32_t>(0ULL);for (;(a < side);a++){
{uint32_t b = cast<uint32_t>(0ULL);for (;(b < side);b++){
m[cast<uint32_t>((cast<uint32_t>((a * cast<uint32_t>(4ULL))) + b))] = go_math_Float32frombits(c->V[cast<uint32_t>((allegrex_matIdx(reg,side,a,b) & cast<uint32_t>(127ULL)))]);
}
}}
}return m;
}
}
// tools/cpu/allegrex/vfpu.go:171:1
void allegrex_CPU_mwrite(allegrex_CPU* c,std::array<float,16> m,uint32_t reg,uint32_t side){
{
{uint32_t a = cast<uint32_t>(0ULL);for (;(a < side);a++){
{uint32_t b = cast<uint32_t>(0ULL);for (;(b < side);b++){
c->V[cast<uint32_t>((allegrex_matIdx(reg,side,a,b) & cast<uint32_t>(127ULL)))] = go_math_Float32bits(m[cast<uint32_t>((cast<uint32_t>((a * cast<uint32_t>(4ULL))) + b))]);
}
}}
}}
}
// tools/cpu/allegrex/vfpu.go:194:1
void allegrex_applyPfx(std::array<float,4>* v,uint32_t data,uint32_t n){
{
if ((data == cast<uint32_t>(228ULL))) {
return ;
}
std::array<float,4> orig={};
gcopy(sub(orig,0,len(orig)),sub(v,0,n));
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < n);i++){
uint32_t swz = cast<uint32_t>(((shr<uint32_t>(data,(cast<uint32_t>((i * cast<uint32_t>(2ULL)))))) & cast<uint32_t>(3ULL)));
uint32_t abs = cast<uint32_t>(((shr<uint32_t>(data,(cast<uint32_t>((cast<uint32_t>(8ULL) + i))))) & cast<uint32_t>(1ULL)));
uint32_t neg = cast<uint32_t>(((shr<uint32_t>(data,(cast<uint32_t>((cast<uint32_t>(16ULL) + i))))) & cast<uint32_t>(1ULL)));
uint32_t cst = cast<uint32_t>(((shr<uint32_t>(data,(cast<uint32_t>((cast<uint32_t>(12ULL) + i))))) & cast<uint32_t>(1ULL)));
float r={};
if ((cst != cast<uint32_t>(0ULL))) {
r = allegrex_pfxConstants[cast<uint32_t>((swz | shl<uint32_t>(abs,cast<int64_t>(2ULL))))];
}
else {
r = orig[cast<uint32_t>((swz & cast<uint32_t>(3ULL)))];
if ((abs != cast<uint32_t>(0ULL))) {
r = go_math_Float32frombits((go_math_Float32bits(r) & ~(cast<uint32_t>(2147483648ULL))));
}
}
if ((neg != cast<uint32_t>(0ULL))) {
r = go_math_Float32frombits(cast<uint32_t>((go_math_Float32bits(r) ^ cast<uint32_t>(2147483648ULL))));
}
(*v)[i] = r;
}
}}
}
// tools/cpu/allegrex/vfpu.go:221:1
void allegrex_CPU_applyPfxS(allegrex_CPU* c,std::array<float,4>* v,uint32_t n){
{
allegrex_applyPfx(v,c->VfpuCtrl[cast<int64_t>(0ULL)],n);
}
}
// tools/cpu/allegrex/vfpu.go:222:1
void allegrex_CPU_applyPfxT(allegrex_CPU* c,std::array<float,4>* v,uint32_t n){
{
allegrex_applyPfx(v,c->VfpuCtrl[cast<int64_t>(1ULL)],n);
}
}
// tools/cpu/allegrex/vfpu.go:226:1
void allegrex_CPU_applyPfxD(allegrex_CPU* c,std::array<float,4>* v,uint32_t n){
{
uint32_t data = c->VfpuCtrl[cast<int64_t>(2ULL)];
if ((data == cast<uint32_t>(0ULL))) {
return ;
}
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < n);i++){
{
switch(cast<uint32_t>(((shr<uint32_t>(data,(cast<uint32_t>((i * cast<uint32_t>(2ULL)))))) & cast<uint32_t>(3ULL)))){
case cast<uint32_t>(1ULL):{
(*v)[i] = allegrex_vfpuClamp((*v)[i],cast<float>(0ULL),cast<float>(1ULL));
break;}
case cast<uint32_t>(3ULL):{
(*v)[i] = allegrex_vfpuClamp((*v)[i],cast<float>(-cast<int64_t>(1ULL)),cast<float>(1ULL));
break;}
}}
}
}}
}
// tools/cpu/allegrex/vfpu.go:242:1
float allegrex_vfpuClamp(float v,float lo,float hi){
{
if ((v >= hi)) {
return hi;
}
if ((v <= lo)) {
return lo;
}
return v;
}
}
// tools/cpu/allegrex/vfpu.go:253:1
void allegrex_CPU_eatPfx(allegrex_CPU* c){
{
c->VfpuCtrl[cast<int64_t>(0ULL)] = cast<uint32_t>(228ULL);
c->VfpuCtrl[cast<int64_t>(1ULL)] = cast<uint32_t>(228ULL);
c->VfpuCtrl[cast<int64_t>(2ULL)] = cast<uint32_t>(0ULL);
}
}
// tools/cpu/allegrex/vfpu.go:262:1
uint32_t allegrex_vfpuVecN(uint32_t w){
{
return cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(7ULL))) & cast<uint32_t>(1ULL))) + cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(14ULL))) & cast<uint32_t>(2ULL))))) + cast<uint32_t>(1ULL)));
}
}
// tools/cpu/allegrex/vfpu.go:269:1
float allegrex_vfpuSin(float x){
{
double k = go_math_Mod(cast<double>(x),cast<double>(4ULL));
if ((k < cast<double>(0ULL))) {
k += cast<double>(4ULL);
}
{
auto tmp48=k;
if (tmp48==(cast<double>(0ULL)) || tmp48==(cast<double>(2ULL))){
return cast<float>(0ULL);
}
else if (tmp48==(cast<double>(1ULL))){
return cast<float>(1ULL);
}
else if (tmp48==(cast<double>(3ULL))){
return cast<float>(-cast<int64_t>(1ULL));
}
}
tmp47:;
return cast<float>(go_math_Sin(((k * go_math_Pi) / cast<double>(2ULL))));
}
}
// tools/cpu/allegrex/vfpu.go:285:1
float allegrex_vfpuCos(float x){
{
double k = go_math_Mod(cast<double>(x),cast<double>(4ULL));
if ((k < cast<double>(0ULL))) {
k += cast<double>(4ULL);
}
{
auto tmp50=k;
if (tmp50==(cast<double>(0ULL))){
return cast<float>(1ULL);
}
else if (tmp50==(cast<double>(2ULL))){
return cast<float>(-cast<int64_t>(1ULL));
}
else if (tmp50==(cast<double>(1ULL)) || tmp50==(cast<double>(3ULL))){
return cast<float>(0ULL);
}
}
tmp49:;
return cast<float>(go_math_Cos(((k * go_math_Pi) / cast<double>(2ULL))));
}
}
// tools/cpu/allegrex/vfpu.go:302:1
float allegrex_float16to32(uint16_t h){
{
uint32_t sign = shl<uint32_t>(cast<uint32_t>(shr<uint16_t>(h,cast<int64_t>(15ULL))),cast<int64_t>(31ULL));
uint32_t exp = cast<uint32_t>((cast<uint32_t>(shr<uint16_t>(h,cast<int64_t>(10ULL))) & cast<uint32_t>(31ULL)));
uint32_t frac = cast<uint32_t>((cast<uint32_t>(h) & cast<uint32_t>(1023ULL)));
{
switch(exp){
case cast<uint32_t>(0ULL):{
if ((frac == cast<uint32_t>(0ULL))) {
return go_math_Float32frombits(sign);
}
return (go_math_Float32frombits(sign) + ((cast<float>(frac) * cast<float>(go_math_Pow(cast<double>(2ULL),cast<double>(-cast<int64_t>(24ULL))))) * allegrex_sign2(sign)));
break;}
case cast<uint32_t>(31ULL):{
return go_math_Float32frombits(cast<uint32_t>((cast<uint32_t>((sign | cast<uint32_t>(2139095040ULL))) | shl<uint32_t>(frac,cast<int64_t>(13ULL)))));
break;}
}}
return go_math_Float32frombits(cast<uint32_t>((cast<uint32_t>((sign | shl<uint32_t>((cast<uint32_t>((exp + cast<uint32_t>(112ULL)))),cast<int64_t>(23ULL)))) | shl<uint32_t>(frac,cast<int64_t>(13ULL)))));
}
}
// tools/cpu/allegrex/vfpu.go:318:1
float allegrex_sign2(uint32_t signBit){
{
if ((signBit != cast<uint32_t>(0ULL))) {
return cast<float>(-cast<int64_t>(1ULL));
}
return cast<float>(1ULL);
}
}
// tools/cpu/allegrex/vfpu.go:352:1
std::string allegrex_vnot(uint32_t reg,uint32_t n){
{
uint32_t mtx = cast<uint32_t>(((shr<uint32_t>(reg,cast<int64_t>(2ULL))) & cast<uint32_t>(7ULL)));
uint32_t col = cast<uint32_t>((reg & cast<uint32_t>(3ULL)));
uint32_t row={};
uint8_t ch = cast<uint8_t>(67ULL);
{
switch(n){
case cast<uint32_t>(1ULL):{
return go_fmt_Sprintf(std::string("S%d%d%d",7),mtx,col,cast<uint32_t>(((shr<uint32_t>(reg,cast<int64_t>(5ULL))) & cast<uint32_t>(3ULL))));
break;}
case cast<uint32_t>(2ULL):case cast<uint32_t>(4ULL):{
row = cast<uint32_t>(((shr<uint32_t>(reg,cast<int64_t>(5ULL))) & cast<uint32_t>(2ULL)));
break;}
case cast<uint32_t>(3ULL):{
row = cast<uint32_t>(((shr<uint32_t>(reg,cast<int64_t>(6ULL))) & cast<uint32_t>(1ULL)));
break;}
}}
if ((cast<uint32_t>((reg & cast<uint32_t>(32ULL))) != cast<uint32_t>(0ULL))) {
ch = cast<uint8_t>(82ULL);
return go_fmt_Sprintf(std::string("%c%d%d%d",8),ch,mtx,row,col);
}
return go_fmt_Sprintf(std::string("%c%d%d%d",8),ch,mtx,col,row);
}
}
// tools/cpu/allegrex/vfpu.go:373:1
std::string allegrex_mnot(uint32_t reg,uint32_t side){
{
uint32_t mtx = cast<uint32_t>(((shr<uint32_t>(reg,cast<int64_t>(2ULL))) & cast<uint32_t>(7ULL)));
uint32_t col = cast<uint32_t>((reg & cast<uint32_t>(3ULL)));
uint32_t row={};
{
switch(side){
case cast<uint32_t>(2ULL):case cast<uint32_t>(4ULL):{
row = cast<uint32_t>(((shr<uint32_t>(reg,cast<int64_t>(5ULL))) & cast<uint32_t>(2ULL)));
break;}
case cast<uint32_t>(3ULL):{
row = cast<uint32_t>(((shr<uint32_t>(reg,cast<int64_t>(6ULL))) & cast<uint32_t>(1ULL)));
break;}
}}
if ((cast<uint32_t>((reg & cast<uint32_t>(32ULL))) != cast<uint32_t>(0ULL))) {
return go_fmt_Sprintf(std::string("E%d%d%d",7),mtx,row,col);
}
return go_fmt_Sprintf(std::string("M%d%d%d",7),mtx,col,row);
}
}
// tools/cpu/allegrex/vfpu.go:390:1
allegrex_Inst allegrex_decodeCop2(allegrex_Inst in,uint32_t w,uint32_t rs,uint32_t rt,uint32_t rd){
{
auto set = [&](std::string mnem,std::string text)->allegrex_Inst{
auto tmp51 = std::make_tuple(mnem,text);
in.Mnem = std::get<0>(tmp51);
in.Text = std::get<1>(tmp51);
return in;
}
;
uint32_t imm = cast<uint32_t>((w & cast<uint32_t>(255ULL)));
{
switch(rs){
case cast<uint32_t>(3ULL):{
if ((imm >= cast<uint32_t>(128ULL))) {
return set(std::string("mfvc",4),go_fmt_Sprintf(std::string("mfvc %s, $%d",12),allegrex_reg(rt),cast<uint32_t>((imm - cast<uint32_t>(128ULL)))));
}
return set(std::string("mfv",3),go_fmt_Sprintf(std::string("mfv %s, %s",10),allegrex_reg(rt),allegrex_vnot(imm,cast<uint32_t>(1ULL))));
break;}
case cast<uint32_t>(7ULL):{
if ((imm >= cast<uint32_t>(128ULL))) {
return set(std::string("mtvc",4),go_fmt_Sprintf(std::string("mtvc %s, $%d",12),allegrex_reg(rt),cast<uint32_t>((imm - cast<uint32_t>(128ULL)))));
}
return set(std::string("mtv",3),go_fmt_Sprintf(std::string("mtv %s, %s",10),allegrex_reg(rt),allegrex_vnot(imm,cast<uint32_t>(1ULL))));
break;}
case cast<uint32_t>(8ULL):{
auto tmp52 = std::make_tuple(cast<allegrex_Flow>(1ULL),true,true);
in.Flow = std::get<0>(tmp52);
in.HasTarget = std::get<1>(tmp52);
in.HasDelay = std::get<2>(tmp52);
in.Target = cast<uint32_t>((cast<uint32_t>((in.Addr + cast<uint32_t>(4ULL))) + cast<uint32_t>((cast<uint32_t>(cast<int32_t>(cast<int16_t>(w))) * cast<uint32_t>(4ULL)))));
std::string m = std::array<std::string,4>{std::string("bvf",3),std::string("bvt",3),std::string("bvfl",4),std::string("bvtl",4)}[cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(3ULL)))];
return set(m,go_fmt_Sprintf(std::string("%s %d, $%08X",12),m,cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(18ULL))) & cast<uint32_t>(7ULL))),in.Target));
break;}
}}
return set(std::string("cop2",4),go_fmt_Sprintf(std::string("cop2 0x%08X",11),w));
}
}
// tools/cpu/allegrex/vfpu.go:416:1
allegrex_Inst allegrex_decodeVFPU(allegrex_Inst in,uint32_t w,uint32_t op,uint32_t rs,uint32_t rt,uint32_t simm){
{
auto set = [&](std::string mnem,std::string text)->allegrex_Inst{
auto tmp53 = std::make_tuple(mnem,text);
in.Mnem = std::get<0>(tmp53);
in.Text = std::get<1>(tmp53);
return in;
}
;
uint32_t n = allegrex_vfpuVecN(w);
std::string sfx = allegrex_vfpuSuffix[n];
uint32_t vd = cast<uint32_t>((w & cast<uint32_t>(127ULL)));
uint32_t vs = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(8ULL))) & cast<uint32_t>(127ULL)));
uint32_t vt = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(127ULL)));
int32_t off = (cast<int32_t>(simm) & ~(cast<int32_t>(3ULL)));
auto vv = [&](std::string name)->allegrex_Inst{
return set(((name + std::string(".",1)) + sfx),go_fmt_Sprintf(std::string("%s.%s %s, %s",12),name,sfx,allegrex_vnot(vd,n),allegrex_vnot(vs,n)));
}
;
auto vvv = [&](std::string name)->allegrex_Inst{
return set(((name + std::string(".",1)) + sfx),go_fmt_Sprintf(std::string("%s.%s %s, %s, %s",16),name,sfx,allegrex_vnot(vd,n),allegrex_vnot(vs,n),allegrex_vnot(vt,n)));
}
;
auto v1 = [&](std::string name)->allegrex_Inst{
return set(((name + std::string(".",1)) + sfx),go_fmt_Sprintf(std::string("%s.%s %s",8),name,sfx,allegrex_vnot(vd,n)));
}
;
{
switch(op){
case cast<uint32_t>(50ULL):{
uint32_t vt = cast<uint32_t>((rt | shl<uint32_t>((cast<uint32_t>((w & cast<uint32_t>(3ULL)))),cast<int64_t>(5ULL))));
return set(std::string("lv.s",4),go_fmt_Sprintf(std::string("lv.s %s, %d(%s)",15),allegrex_vnot(vt,cast<uint32_t>(1ULL)),off,allegrex_reg(rs)));
break;}
case cast<uint32_t>(58ULL):{
uint32_t vt = cast<uint32_t>((rt | shl<uint32_t>((cast<uint32_t>((w & cast<uint32_t>(3ULL)))),cast<int64_t>(5ULL))));
return set(std::string("sv.s",4),go_fmt_Sprintf(std::string("sv.s %s, %d(%s)",15),allegrex_vnot(vt,cast<uint32_t>(1ULL)),off,allegrex_reg(rs)));
break;}
case cast<uint32_t>(54ULL):{
uint32_t vt = cast<uint32_t>((rt | shl<uint32_t>((cast<uint32_t>((w & cast<uint32_t>(1ULL)))),cast<int64_t>(5ULL))));
return set(std::string("lv.q",4),go_fmt_Sprintf(std::string("lv.q %s, %d(%s)",15),allegrex_vnot(vt,cast<uint32_t>(4ULL)),off,allegrex_reg(rs)));
break;}
case cast<uint32_t>(62ULL):{
uint32_t vt = cast<uint32_t>((rt | shl<uint32_t>((cast<uint32_t>((w & cast<uint32_t>(1ULL)))),cast<int64_t>(5ULL))));
return set(std::string("sv.q",4),go_fmt_Sprintf(std::string("sv.q %s, %d(%s)",15),allegrex_vnot(vt,cast<uint32_t>(4ULL)),off,allegrex_reg(rs)));
break;}
case cast<uint32_t>(53ULL):{
uint32_t vt = cast<uint32_t>((rt | shl<uint32_t>((cast<uint32_t>((w & cast<uint32_t>(1ULL)))),cast<int64_t>(5ULL))));
return set(((std::string("lv",2) + allegrex_lr(w)) + std::string(".q",2)),go_fmt_Sprintf(std::string("lv%s.q %s, %d(%s)",17),allegrex_lr(w),allegrex_vnot(vt,cast<uint32_t>(4ULL)),off,allegrex_reg(rs)));
break;}
case cast<uint32_t>(61ULL):{
uint32_t vt = cast<uint32_t>((rt | shl<uint32_t>((cast<uint32_t>((w & cast<uint32_t>(1ULL)))),cast<int64_t>(5ULL))));
return set(((std::string("sv",2) + allegrex_lr(w)) + std::string(".q",2)),go_fmt_Sprintf(std::string("sv%s.q %s, %d(%s)",17),allegrex_lr(w),allegrex_vnot(vt,cast<uint32_t>(4ULL)),off,allegrex_reg(rs)));
break;}
case cast<uint32_t>(24ULL):{
{
switch(cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(23ULL))) & cast<uint32_t>(7ULL)))){
case cast<uint32_t>(0ULL):{
return vvv(std::string("vadd",4));
break;}
case cast<uint32_t>(1ULL):{
return vvv(std::string("vsub",4));
break;}
case cast<uint32_t>(2ULL):{
return vvv(std::string("vsbn",4));
break;}
case cast<uint32_t>(7ULL):{
return vvv(std::string("vdiv",4));
break;}
}}
break;}
case cast<uint32_t>(25ULL):{
{
switch(cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(23ULL))) & cast<uint32_t>(7ULL)))){
case cast<uint32_t>(0ULL):{
return vvv(std::string("vmul",4));
break;}
case cast<uint32_t>(1ULL):{
return set((std::string("vdot.",5) + sfx),go_fmt_Sprintf(std::string("vdot.%s %s, %s, %s",18),sfx,allegrex_vnot(vd,cast<uint32_t>(1ULL)),allegrex_vnot(vs,n),allegrex_vnot(vt,n)));
break;}
case cast<uint32_t>(2ULL):{
return set((std::string("vscl.",5) + sfx),go_fmt_Sprintf(std::string("vscl.%s %s, %s, %s",18),sfx,allegrex_vnot(vd,n),allegrex_vnot(vs,n),allegrex_vnot(vt,cast<uint32_t>(1ULL))));
break;}
case cast<uint32_t>(4ULL):{
return set((std::string("vhdp.",5) + sfx),go_fmt_Sprintf(std::string("vhdp.%s %s, %s, %s",18),sfx,allegrex_vnot(vd,cast<uint32_t>(1ULL)),allegrex_vnot(vs,n),allegrex_vnot(vt,n)));
break;}
case cast<uint32_t>(5ULL):{
return vvv(std::string("vcrs",4));
break;}
case cast<uint32_t>(6ULL):{
return set((std::string("vdet.",5) + sfx),go_fmt_Sprintf(std::string("vdet.%s %s, %s, %s",18),sfx,allegrex_vnot(vd,cast<uint32_t>(1ULL)),allegrex_vnot(vs,n),allegrex_vnot(vt,n)));
break;}
}}
break;}
case cast<uint32_t>(27ULL):{
{
switch(cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(23ULL))) & cast<uint32_t>(7ULL)))){
case cast<uint32_t>(0ULL):{
return set((std::string("vcmp.",5) + sfx),go_fmt_Sprintf(std::string("vcmp.%s %d, %s, %s",18),sfx,cast<uint32_t>((w & cast<uint32_t>(15ULL))),allegrex_vnot(vs,n),allegrex_vnot(vt,n)));
break;}
case cast<uint32_t>(2ULL):{
return vvv(std::string("vmin",4));
break;}
case cast<uint32_t>(3ULL):{
return vvv(std::string("vmax",4));
break;}
case cast<uint32_t>(5ULL):{
return vvv(std::string("vscmp",5));
break;}
case cast<uint32_t>(6ULL):{
return vvv(std::string("vsge",4));
break;}
case cast<uint32_t>(7ULL):{
return vvv(std::string("vslt",4));
break;}
}}
break;}
case cast<uint32_t>(52ULL):{
{
switch(rs){
case cast<uint32_t>(0ULL):{
Map<uint32_t,std::string> names = Map<uint32_t,std::string>{{cast<uint32_t>(0ULL),std::string("vmov",4)},{cast<uint32_t>(1ULL),std::string("vabs",4)},{cast<uint32_t>(2ULL),std::string("vneg",4)},{cast<uint32_t>(3ULL),std::string("vidt",4)},{cast<uint32_t>(4ULL),std::string("vsat0",5)},{cast<uint32_t>(5ULL),std::string("vsat1",5)},{cast<uint32_t>(6ULL),std::string("vzero",5)},{cast<uint32_t>(7ULL),std::string("vone",4)},{cast<uint32_t>(16ULL),std::string("vrcp",4)},{cast<uint32_t>(17ULL),std::string("vrsq",4)},{cast<uint32_t>(18ULL),std::string("vsin",4)},{cast<uint32_t>(19ULL),std::string("vcos",4)},{cast<uint32_t>(20ULL),std::string("vexp2",5)},{cast<uint32_t>(21ULL),std::string("vlog2",5)},{cast<uint32_t>(22ULL),std::string("vsqrt",5)},{cast<uint32_t>(23ULL),std::string("vasin",5)},{cast<uint32_t>(24ULL),std::string("vnrcp",5)},{cast<uint32_t>(26ULL),std::string("vnsin",5)},{cast<uint32_t>(28ULL),std::string("vrexp2",6)}};
{
auto tmp54 = lookup(names,rt);
std::string name = std::get<0>(tmp54);
bool ok = std::get<1>(tmp54);
if (ok) {
{
switch(rt){
case cast<uint32_t>(3ULL):case cast<uint32_t>(6ULL):case cast<uint32_t>(7ULL):{
return v1(name);
break;}
}}
return vv(name);
}
}
break;}
case cast<uint32_t>(1ULL):{
Map<uint32_t,std::string> names = Map<uint32_t,std::string>{{cast<uint32_t>(0ULL),std::string("vrnds",5)},{cast<uint32_t>(1ULL),std::string("vrndi",5)},{cast<uint32_t>(2ULL),std::string("vrndf1",6)},{cast<uint32_t>(3ULL),std::string("vrndf2",6)},{cast<uint32_t>(18ULL),std::string("vf2h",4)},{cast<uint32_t>(19ULL),std::string("vh2f",4)},{cast<uint32_t>(22ULL),std::string("vsbz",4)},{cast<uint32_t>(23ULL),std::string("vlgb",4)},{cast<uint32_t>(24ULL),std::string("vuc2ifs",7)},{cast<uint32_t>(25ULL),std::string("vc2i",4)},{cast<uint32_t>(26ULL),std::string("vus2i",5)},{cast<uint32_t>(27ULL),std::string("vs2i",4)},{cast<uint32_t>(28ULL),std::string("vi2uc",5)},{cast<uint32_t>(29ULL),std::string("vi2c",4)},{cast<uint32_t>(30ULL),std::string("vi2us",5)},{cast<uint32_t>(31ULL),std::string("vi2s",4)}};
{
auto tmp55 = lookup(names,rt);
std::string name = std::get<0>(tmp55);
bool ok = std::get<1>(tmp55);
if (ok) {
return vv(name);
}
}
break;}
case cast<uint32_t>(2ULL):{
Map<uint32_t,std::string> names = Map<uint32_t,std::string>{{cast<uint32_t>(0ULL),std::string("vsrt1",5)},{cast<uint32_t>(1ULL),std::string("vsrt2",5)},{cast<uint32_t>(2ULL),std::string("vbfy1",5)},{cast<uint32_t>(3ULL),std::string("vbfy2",5)},{cast<uint32_t>(4ULL),std::string("vocp",4)},{cast<uint32_t>(5ULL),std::string("vsocp",5)},{cast<uint32_t>(6ULL),std::string("vfad",4)},{cast<uint32_t>(7ULL),std::string("vavg",4)},{cast<uint32_t>(8ULL),std::string("vsrt3",5)},{cast<uint32_t>(9ULL),std::string("vsrt4",5)},{cast<uint32_t>(10ULL),std::string("vsgn",4)},{cast<uint32_t>(16ULL),std::string("vmfvc",5)},{cast<uint32_t>(17ULL),std::string("vmtvc",5)},{cast<uint32_t>(25ULL),std::string("vt4444",6)},{cast<uint32_t>(26ULL),std::string("vt5551",6)},{cast<uint32_t>(27ULL),std::string("vt5650",6)}};
{
auto tmp56 = lookup(names,rt);
std::string name = std::get<0>(tmp56);
bool ok = std::get<1>(tmp56);
if (ok) {
return vv(name);
}
}
break;}
case cast<uint32_t>(3ULL):{
return set((std::string("vcst.",5) + sfx),go_fmt_Sprintf(std::string("vcst.%s %s, %d",14),sfx,allegrex_vnot(vd,n),rt));
break;}
case cast<uint32_t>(16ULL):case cast<uint32_t>(17ULL):case cast<uint32_t>(18ULL):case cast<uint32_t>(19ULL):{
std::string name = std::array<std::string,4>{std::string("vf2in",5),std::string("vf2iz",5),std::string("vf2iu",5),std::string("vf2id",5)}[cast<uint32_t>((rs - cast<uint32_t>(16ULL)))];
return set(((name + std::string(".",1)) + sfx),go_fmt_Sprintf(std::string("%s.%s %s, %s, %d",16),name,sfx,allegrex_vnot(vd,n),allegrex_vnot(vs,n),rt));
break;}
case cast<uint32_t>(20ULL):{
return set((std::string("vi2f.",5) + sfx),go_fmt_Sprintf(std::string("vi2f.%s %s, %s, %d",18),sfx,allegrex_vnot(vd,n),allegrex_vnot(vs,n),rt));
break;}
case cast<uint32_t>(21ULL):{
std::string name = std::string("vcmovt",6);
if ((cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(19ULL))) & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL))) {
name = std::string("vcmovf",6);
}
return set(((name + std::string(".",1)) + sfx),go_fmt_Sprintf(std::string("%s.%s %s, %s, %d",16),name,sfx,allegrex_vnot(vd,n),allegrex_vnot(vs,n),cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(7ULL)))));
break;}
}}
if ((rs >= cast<uint32_t>(24ULL))) {
return vv(std::string("vwbn",4));
}
break;}
case cast<uint32_t>(55ULL):{
{
switch(cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(24ULL))) & cast<uint32_t>(3ULL)))){
case cast<uint32_t>(0ULL):{
return set(std::string("vpfxs",5),go_fmt_Sprintf(std::string("vpfxs 0x%05X",12),cast<uint32_t>((w & cast<uint32_t>(1048575ULL)))));
break;}
case cast<uint32_t>(1ULL):{
return set(std::string("vpfxt",5),go_fmt_Sprintf(std::string("vpfxt 0x%05X",12),cast<uint32_t>((w & cast<uint32_t>(1048575ULL)))));
break;}
case cast<uint32_t>(2ULL):{
return set(std::string("vpfxd",5),go_fmt_Sprintf(std::string("vpfxd 0x%03X",12),cast<uint32_t>((w & cast<uint32_t>(4095ULL)))));
break;}
case cast<uint32_t>(3ULL):{
uint32_t vt7 = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(127ULL)));
if ((cast<uint32_t>((w & cast<uint32_t>(8388608ULL))) != cast<uint32_t>(0ULL))) {
return set(std::string("vfim.s",6),go_fmt_Sprintf(std::string("vfim.s %s, 0x%04X",17),allegrex_vnot(vt7,cast<uint32_t>(1ULL)),cast<uint32_t>((w & cast<uint32_t>(65535ULL)))));
}
return set(std::string("viim.s",6),go_fmt_Sprintf(std::string("viim.s %s, %d",13),allegrex_vnot(vt7,cast<uint32_t>(1ULL)),cast<int32_t>(cast<int16_t>(w))));
break;}
}}
break;}
case cast<uint32_t>(60ULL):{
{
switch(cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(21ULL))) & cast<uint32_t>(31ULL)))){
case cast<uint32_t>(0ULL):case cast<uint32_t>(1ULL):case cast<uint32_t>(2ULL):case cast<uint32_t>(3ULL):{
return set((std::string("vmmul.",6) + sfx),go_fmt_Sprintf(std::string("vmmul.%s %s, %s, %s",19),sfx,allegrex_mnot(vd,n),allegrex_mnot(vs,n),allegrex_mnot(vt,n)));
break;}
case cast<uint32_t>(16ULL):case cast<uint32_t>(17ULL):case cast<uint32_t>(18ULL):case cast<uint32_t>(19ULL):{
return set((std::string("vmscl.",6) + sfx),go_fmt_Sprintf(std::string("vmscl.%s %s, %s, %s",19),sfx,allegrex_mnot(vd,n),allegrex_mnot(vs,n),allegrex_vnot(vt,cast<uint32_t>(1ULL))));
break;}
case cast<uint32_t>(20ULL):case cast<uint32_t>(21ULL):case cast<uint32_t>(22ULL):case cast<uint32_t>(23ULL):{
std::string name = std::string("vcrsp",5);
if ((n == cast<uint32_t>(4ULL))) {
name = std::string("vqmul",5);
}
return vvv(name);
break;}
case cast<uint32_t>(28ULL):{
{
switch(cast<uint32_t>((rt & cast<uint32_t>(15ULL)))){
case cast<uint32_t>(0ULL):{
return set((std::string("vmmov.",6) + sfx),go_fmt_Sprintf(std::string("vmmov.%s %s, %s",15),sfx,allegrex_mnot(vd,n),allegrex_mnot(vs,n)));
break;}
case cast<uint32_t>(3ULL):{
return set((std::string("vmidt.",6) + sfx),go_fmt_Sprintf(std::string("vmidt.%s %s",11),sfx,allegrex_mnot(vd,n)));
break;}
case cast<uint32_t>(6ULL):{
return set((std::string("vmzero.",7) + sfx),go_fmt_Sprintf(std::string("vmzero.%s %s",12),sfx,allegrex_mnot(vd,n)));
break;}
case cast<uint32_t>(7ULL):{
return set((std::string("vmone.",6) + sfx),go_fmt_Sprintf(std::string("vmone.%s %s",11),sfx,allegrex_mnot(vd,n)));
break;}
}}
break;}
case cast<uint32_t>(29ULL):{
return set((std::string("vrot.",5) + sfx),go_fmt_Sprintf(std::string("vrot.%s %s, %s, %d",18),sfx,allegrex_vnot(vd,n),allegrex_vnot(vs,cast<uint32_t>(1ULL)),cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(31ULL)))));
break;}
default:{
uint32_t ins = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(23ULL))) & cast<uint32_t>(3ULL)));
if ((ins >= cast<uint32_t>(1ULL))) {
std::string name = go_fmt_Sprintf(std::string("vtfm%d",6),cast<uint32_t>((ins + cast<uint32_t>(1ULL))));
if ((n != cast<uint32_t>((ins + cast<uint32_t>(1ULL))))) {
name = go_fmt_Sprintf(std::string("vhtfm%d",7),cast<uint32_t>((ins + cast<uint32_t>(1ULL))));
}
return set(name,go_fmt_Sprintf(std::string("%s.%s %s, %s, %s",16),name,sfx,allegrex_vnot(vd,cast<uint32_t>((ins + cast<uint32_t>(1ULL)))),allegrex_mnot(vs,cast<uint32_t>((ins + cast<uint32_t>(1ULL)))),allegrex_vnot(vt,n)));
}
break;}
}}
break;}
}}
return set(std::string("vfpu",4),go_fmt_Sprintf(std::string("vfpu.%02X 0x%08X",16),op,w));
}
}
// tools/cpu/allegrex/vfpu.go:601:1
std::string allegrex_lr(uint32_t w){
{
if ((cast<uint32_t>((w & cast<uint32_t>(2ULL))) != cast<uint32_t>(0ULL))) {
return std::string("r",1);
}
return std::string("l",1);
}
}
// tools/cpu/allegrex/vfpu.go:611:1
void allegrex_CPU_cop2(allegrex_CPU* c,uint32_t w,uint32_t rs,uint32_t rt,uint32_t rd){
{
uint32_t imm = cast<uint32_t>((w & cast<uint32_t>(255ULL)));
{
switch(rs){
case cast<uint32_t>(3ULL):{
if ((imm >= cast<uint32_t>(128ULL))) {
uint32_t v={};
if ((imm < cast<uint32_t>(144ULL))) {
v = c->VfpuCtrl[cast<uint32_t>((imm - cast<uint32_t>(128ULL)))];
}
allegrex_CPU_load(c,rt,v);
return ;
}
allegrex_CPU_load(c,rt,c->V[cast<uint32_t>((allegrex_vfpuSingle(imm) & cast<uint32_t>(127ULL)))]);
break;}
case cast<uint32_t>(7ULL):{
if ((imm >= cast<uint32_t>(128ULL))) {
if ((imm < cast<uint32_t>(144ULL))) {
uint32_t i = cast<uint32_t>((imm - cast<uint32_t>(128ULL)));
if ((allegrex_vfpuCtrlMask[i] != cast<uint32_t>(0ULL))) {
c->VfpuCtrl[i] = cast<uint32_t>((allegrex_CPU_reg(c,rt) & allegrex_vfpuCtrlMask[i]));
}
}
return ;
}
c->V[cast<uint32_t>((allegrex_vfpuSingle(imm) & cast<uint32_t>(127ULL)))] = allegrex_CPU_reg(c,rt);
break;}
case cast<uint32_t>(8ULL):{
uint32_t target = cast<uint32_t>((cast<uint32_t>((c->curPC + cast<uint32_t>(4ULL))) + shl<uint32_t>(cast<uint32_t>(cast<int32_t>(cast<int16_t>(w))),cast<int64_t>(2ULL))));
bool val = (cast<uint32_t>(((shr<uint32_t>(c->VfpuCtrl[cast<int64_t>(3ULL)],(cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(18ULL))) & cast<uint32_t>(7ULL)))))) & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL));
{
switch(cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(3ULL)))){
case cast<uint32_t>(0ULL):{
allegrex_CPU_doBranch(c,(!val),target);
break;}
case cast<uint32_t>(1ULL):{
allegrex_CPU_doBranch(c,val,target);
break;}
case cast<uint32_t>(2ULL):{
if ((!val)) {
allegrex_CPU_doBranch(c,true,target);
}
else {
c->nullifyNext = true;
}
break;}
case cast<uint32_t>(3ULL):{
if (val) {
allegrex_CPU_doBranch(c,true,target);
}
else {
c->nullifyNext = true;
}
break;}
}}
break;}
default:{
allegrex_CPU_Halt(c,std::string("unimplemented cop2 rs=0x%02X (word 0x%08X) at 0x%08X",52),rs,w,c->curPC);
break;}
}}
}
}
// tools/cpu/allegrex/vfpu.go:662:1
void allegrex_CPU_vfpuOp(allegrex_CPU* c,uint32_t w,uint32_t op,uint32_t rs,uint32_t rt,uint32_t simm){
{
{
switch(op){
case cast<uint32_t>(50ULL):{
uint32_t vt = cast<uint32_t>((rt | shl<uint32_t>((cast<uint32_t>((w & cast<uint32_t>(3ULL)))),cast<int64_t>(5ULL))));
uint32_t addr = cast<uint32_t>((allegrex_CPU_reg(c,rs) + ((simm & ~(cast<uint32_t>(3ULL))))));
if ((cast<uint32_t>((addr & cast<uint32_t>(3ULL))) != cast<uint32_t>(0ULL))) {
allegrex_CPU_addrError(c,cast<uint32_t>(4ULL),addr);
return ;
}
c->V[cast<uint32_t>((allegrex_vfpuSingle(vt) & cast<uint32_t>(127ULL)))] = allegrex_CPU_read32(c,addr);
break;}
case cast<uint32_t>(58ULL):{
uint32_t vt = cast<uint32_t>((rt | shl<uint32_t>((cast<uint32_t>((w & cast<uint32_t>(3ULL)))),cast<int64_t>(5ULL))));
uint32_t addr = cast<uint32_t>((allegrex_CPU_reg(c,rs) + ((simm & ~(cast<uint32_t>(3ULL))))));
if ((cast<uint32_t>((addr & cast<uint32_t>(3ULL))) != cast<uint32_t>(0ULL))) {
allegrex_CPU_addrError(c,cast<uint32_t>(5ULL),addr);
return ;
}
allegrex_CPU_write32(c,addr,c->V[cast<uint32_t>((allegrex_vfpuSingle(vt) & cast<uint32_t>(127ULL)))]);
break;}
case cast<uint32_t>(54ULL):{
uint32_t vt = cast<uint32_t>((rt | shl<uint32_t>((cast<uint32_t>((w & cast<uint32_t>(1ULL)))),cast<int64_t>(5ULL))));
uint32_t addr = cast<uint32_t>((allegrex_CPU_reg(c,rs) + ((simm & ~(cast<uint32_t>(3ULL))))));
if ((cast<uint32_t>((addr & cast<uint32_t>(15ULL))) != cast<uint32_t>(0ULL))) {
allegrex_CPU_addrError(c,cast<uint32_t>(4ULL),addr);
return ;
}
std::array<uint32_t,4> v={};
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < cast<uint32_t>(4ULL));i++){
v[i] = allegrex_CPU_read32(c,cast<uint32_t>((addr + cast<uint32_t>((i * cast<uint32_t>(4ULL))))));
}
}std::array<uint32_t,4> idx = allegrex_vecIdx(vt,cast<uint32_t>(4ULL));
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < cast<uint32_t>(4ULL));i++){
c->V[cast<uint32_t>((idx[i] & cast<uint32_t>(127ULL)))] = v[i];
}
}break;}
case cast<uint32_t>(62ULL):{
uint32_t vt = cast<uint32_t>((rt | shl<uint32_t>((cast<uint32_t>((w & cast<uint32_t>(1ULL)))),cast<int64_t>(5ULL))));
uint32_t addr = cast<uint32_t>((allegrex_CPU_reg(c,rs) + ((simm & ~(cast<uint32_t>(3ULL))))));
if ((cast<uint32_t>((addr & cast<uint32_t>(15ULL))) != cast<uint32_t>(0ULL))) {
allegrex_CPU_addrError(c,cast<uint32_t>(5ULL),addr);
return ;
}
std::array<uint32_t,4> idx = allegrex_vecIdx(vt,cast<uint32_t>(4ULL));
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < cast<uint32_t>(4ULL));i++){
allegrex_CPU_write32(c,cast<uint32_t>((addr + cast<uint32_t>((i * cast<uint32_t>(4ULL))))),c->V[cast<uint32_t>((idx[i] & cast<uint32_t>(127ULL)))]);
}
}break;}
case cast<uint32_t>(53ULL):{
uint32_t vt = cast<uint32_t>((rt | shl<uint32_t>((cast<uint32_t>((w & cast<uint32_t>(1ULL)))),cast<int64_t>(5ULL))));
uint32_t addr = cast<uint32_t>((allegrex_CPU_reg(c,rs) + ((simm & ~(cast<uint32_t>(3ULL))))));
if ((cast<uint32_t>((addr & cast<uint32_t>(3ULL))) != cast<uint32_t>(0ULL))) {
allegrex_CPU_addrError(c,cast<uint32_t>(4ULL),addr);
return ;
}
std::array<uint32_t,4> idx = allegrex_vecIdx(vt,cast<uint32_t>(4ULL));
uint32_t offset = cast<uint32_t>(((shr<uint32_t>(addr,cast<int64_t>(2ULL))) & cast<uint32_t>(3ULL)));
if ((cast<uint32_t>((w & cast<uint32_t>(2ULL))) == cast<uint32_t>(0ULL))) {
{uint32_t i = cast<uint32_t>(0ULL);for (;(i <= offset);i++){
c->V[cast<uint32_t>((idx[cast<uint32_t>((cast<uint32_t>(3ULL) - i))] & cast<uint32_t>(127ULL)))] = allegrex_CPU_read32(c,cast<uint32_t>((addr - cast<uint32_t>((cast<uint32_t>(4ULL) * i)))));
}
}}
else {
{uint32_t i = cast<uint32_t>(0ULL);for (;(i <= cast<uint32_t>((cast<uint32_t>(3ULL) - offset)));i++){
c->V[cast<uint32_t>((idx[i] & cast<uint32_t>(127ULL)))] = allegrex_CPU_read32(c,cast<uint32_t>((addr + cast<uint32_t>((cast<uint32_t>(4ULL) * i)))));
}
}}
break;}
case cast<uint32_t>(61ULL):{
uint32_t vt = cast<uint32_t>((rt | shl<uint32_t>((cast<uint32_t>((w & cast<uint32_t>(1ULL)))),cast<int64_t>(5ULL))));
uint32_t addr = cast<uint32_t>((allegrex_CPU_reg(c,rs) + ((simm & ~(cast<uint32_t>(3ULL))))));
if ((cast<uint32_t>((addr & cast<uint32_t>(3ULL))) != cast<uint32_t>(0ULL))) {
allegrex_CPU_addrError(c,cast<uint32_t>(5ULL),addr);
return ;
}
std::array<uint32_t,4> idx = allegrex_vecIdx(vt,cast<uint32_t>(4ULL));
uint32_t offset = cast<uint32_t>(((shr<uint32_t>(addr,cast<int64_t>(2ULL))) & cast<uint32_t>(3ULL)));
if ((cast<uint32_t>((w & cast<uint32_t>(2ULL))) == cast<uint32_t>(0ULL))) {
{uint32_t i = cast<uint32_t>(0ULL);for (;(i <= offset);i++){
allegrex_CPU_write32(c,cast<uint32_t>((addr - cast<uint32_t>((cast<uint32_t>(4ULL) * i)))),c->V[cast<uint32_t>((idx[cast<uint32_t>((cast<uint32_t>(3ULL) - i))] & cast<uint32_t>(127ULL)))]);
}
}}
else {
{uint32_t i = cast<uint32_t>(0ULL);for (;(i <= cast<uint32_t>((cast<uint32_t>(3ULL) - offset)));i++){
allegrex_CPU_write32(c,cast<uint32_t>((addr + cast<uint32_t>((cast<uint32_t>(4ULL) * i)))),c->V[cast<uint32_t>((idx[i] & cast<uint32_t>(127ULL)))]);
}
}}
break;}
case cast<uint32_t>(24ULL):case cast<uint32_t>(25ULL):case cast<uint32_t>(27ULL):{
allegrex_CPU_vfpuALU(c,w,op);
break;}
case cast<uint32_t>(52ULL):{
allegrex_CPU_vfpu4(c,w,rs,rt);
break;}
case cast<uint32_t>(55ULL):{
allegrex_CPU_vfpu5(c,w);
break;}
case cast<uint32_t>(60ULL):{
allegrex_CPU_vfpuMatrix(c,w,rt);
break;}
default:{
allegrex_CPU_vfpuUnimpl(c,w,op);
break;}
}}
}
}
// tools/cpu/allegrex/vfpu.go:757:1
void allegrex_CPU_vfpuUnimpl(allegrex_CPU* c,uint32_t w,uint32_t op){
{
if (bool(c->OnVFPU)) {
c->OnVFPU(w,op);
return ;
}
allegrex_CPU_Halt(c,std::string("unimplemented VFPU op 0x%02X %q (word 0x%08X) at 0x%08X",55),op,allegrex_decodeVFPU(allegrex_Inst{},w,op,cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(21ULL))) & cast<uint32_t>(31ULL))),cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(31ULL))),cast<uint32_t>(cast<int32_t>(cast<int16_t>(w)))).Mnem,w,c->curPC);
}
}
// tools/cpu/allegrex/vfpu.go:767:1
void allegrex_CPU_vfpuALU(allegrex_CPU* c,uint32_t w,uint32_t op){
{
uint32_t n = allegrex_vfpuVecN(w);
uint32_t vd = cast<uint32_t>((w & cast<uint32_t>(127ULL)));
uint32_t vs = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(8ULL))) & cast<uint32_t>(127ULL)));
uint32_t vt = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(127ULL)));
uint32_t sub_ = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(23ULL))) & cast<uint32_t>(7ULL)));
std::array<float,4> d={};
{
switch(op){
case cast<uint32_t>(24ULL):{
std::array<float,4> s = allegrex_CPU_vread(c,vs,n);
std::array<float,4> t = allegrex_CPU_vread(c,vt,n);
allegrex_CPU_applyPfxS(c,(&s),n);
allegrex_CPU_applyPfxT(c,(&t),n);
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < n);i++){
{
switch(sub_){
case cast<uint32_t>(0ULL):{
d[i] = (s[i] + t[i]);
break;}
case cast<uint32_t>(1ULL):{
d[i] = (s[i] - t[i]);
break;}
case cast<uint32_t>(7ULL):{
d[i] = (s[i] / t[i]);
break;}
default:{
allegrex_CPU_vfpuUnimpl(c,w,op);
return ;
break;}
}}
}
}break;}
case cast<uint32_t>(25ULL):{
{
switch(sub_){
case cast<uint32_t>(0ULL):{
std::array<float,4> s = allegrex_CPU_vread(c,vs,n);
std::array<float,4> t = allegrex_CPU_vread(c,vt,n);
allegrex_CPU_applyPfxS(c,(&s),n);
allegrex_CPU_applyPfxT(c,(&t),n);
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < n);i++){
d[i] = (s[i] * t[i]);
}
}break;}
case cast<uint32_t>(1ULL):{
std::array<float,4> s = allegrex_CPU_vread(c,vs,n);
std::array<float,4> t = allegrex_CPU_vread(c,vt,n);
allegrex_CPU_applyPfxS(c,(&s),n);
allegrex_CPU_applyPfxT(c,(&t),n);
float sum={};
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < n);i++){
sum += (s[i] * t[i]);
}
}d[cast<int64_t>(0ULL)] = sum;
allegrex_CPU_applyPfxD(c,(&d),cast<uint32_t>(1ULL));
allegrex_CPU_vwrite(c,d,vd,cast<uint32_t>(1ULL));
allegrex_CPU_eatPfx(c);
return ;
break;}
case cast<uint32_t>(2ULL):{
std::array<float,4> s = allegrex_CPU_vread(c,vs,n);
std::array<float,4> t = allegrex_CPU_vread(c,vt,cast<uint32_t>(1ULL));
allegrex_CPU_applyPfxS(c,(&s),n);
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < n);i++){
d[i] = (s[i] * t[cast<int64_t>(0ULL)]);
}
}break;}
case cast<uint32_t>(4ULL):{
std::array<float,4> s = allegrex_CPU_vread(c,vs,n);
std::array<float,4> t = allegrex_CPU_vread(c,vt,n);
allegrex_CPU_applyPfxT(c,(&t),n);
float sum={};
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < cast<uint32_t>((n - cast<uint32_t>(1ULL))));i++){
sum += (s[i] * t[i]);
}
}sum += t[cast<uint32_t>((n - cast<uint32_t>(1ULL)))];
d[cast<int64_t>(0ULL)] = sum;
allegrex_CPU_applyPfxD(c,(&d),cast<uint32_t>(1ULL));
allegrex_CPU_vwrite(c,d,vd,cast<uint32_t>(1ULL));
allegrex_CPU_eatPfx(c);
return ;
break;}
case cast<uint32_t>(5ULL):{
std::array<float,4> s = allegrex_CPU_vread(c,vs,n);
std::array<float,4> t = allegrex_CPU_vread(c,vt,n);
d[cast<int64_t>(0ULL)] = (s[cast<int64_t>(1ULL)] * t[cast<int64_t>(2ULL)]);
d[cast<int64_t>(1ULL)] = (s[cast<int64_t>(2ULL)] * t[cast<int64_t>(0ULL)]);
d[cast<int64_t>(2ULL)] = (s[cast<int64_t>(0ULL)] * t[cast<int64_t>(1ULL)]);
break;}
case cast<uint32_t>(6ULL):{
std::array<float,4> s = allegrex_CPU_vread(c,vs,n);
std::array<float,4> t = allegrex_CPU_vread(c,vt,n);
d[cast<int64_t>(0ULL)] = ((s[cast<int64_t>(0ULL)] * t[cast<int64_t>(1ULL)]) - (s[cast<int64_t>(1ULL)] * t[cast<int64_t>(0ULL)]));
allegrex_CPU_applyPfxD(c,(&d),cast<uint32_t>(1ULL));
allegrex_CPU_vwrite(c,d,vd,cast<uint32_t>(1ULL));
allegrex_CPU_eatPfx(c);
return ;
break;}
default:{
allegrex_CPU_vfpuUnimpl(c,w,op);
return ;
break;}
}}
break;}
case cast<uint32_t>(27ULL):{
std::array<float,4> s = allegrex_CPU_vread(c,vs,n);
std::array<float,4> t = allegrex_CPU_vread(c,vt,n);
allegrex_CPU_applyPfxS(c,(&s),n);
allegrex_CPU_applyPfxT(c,(&t),n);
{
switch(sub_){
case cast<uint32_t>(0ULL):{
uint32_t cond = cast<uint32_t>((w & cast<uint32_t>(15ULL)));
uint32_t cc={};
uint32_t orv={};
uint32_t andv = cast<uint32_t>(1ULL);
uint32_t affected = cast<uint32_t>(48ULL);
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < n);i++){
bool ci={};
{
switch(cond){
case cast<uint32_t>(0ULL):{
ci = false;
break;}
case cast<uint32_t>(1ULL):{
ci = (s[i] == t[i]);
break;}
case cast<uint32_t>(2ULL):{
ci = (s[i] < t[i]);
break;}
case cast<uint32_t>(3ULL):{
ci = (s[i] <= t[i]);
break;}
case cast<uint32_t>(4ULL):{
ci = true;
break;}
case cast<uint32_t>(5ULL):{
ci = (s[i] != t[i]);
break;}
case cast<uint32_t>(6ULL):{
ci = (s[i] >= t[i]);
break;}
case cast<uint32_t>(7ULL):{
ci = (s[i] > t[i]);
break;}
case cast<uint32_t>(8ULL):{
ci = (s[i] == cast<float>(0ULL));
break;}
case cast<uint32_t>(9ULL):{
ci = (s[i] != s[i]);
break;}
case cast<uint32_t>(10ULL):{
ci = allegrex_isInf32(s[i]);
break;}
case cast<uint32_t>(11ULL):{
ci = ((s[i] != s[i]) || allegrex_isInf32(s[i]));
break;}
case cast<uint32_t>(12ULL):{
ci = (s[i] != cast<float>(0ULL));
break;}
case cast<uint32_t>(13ULL):{
ci = (s[i] == s[i]);
break;}
case cast<uint32_t>(14ULL):{
ci = (!allegrex_isInf32(s[i]));
break;}
case cast<uint32_t>(15ULL):{
ci = (!(((s[i] != s[i]) || allegrex_isInf32(s[i]))));
break;}
}}
uint32_t cb={};
if (ci) {
cb = cast<uint32_t>(1ULL);
}
cc |= shl<uint32_t>(cb,i);
orv |= cb;
andv &= cb;
affected |= shl<uint32_t>(cast<uint32_t>(1ULL),i);
}
}c->VfpuCtrl[cast<int64_t>(3ULL)] = cast<uint32_t>((((c->VfpuCtrl[cast<int64_t>(3ULL)] & ~(affected))) | (cast<uint32_t>(((cast<uint32_t>((cast<uint32_t>((cc | shl<uint32_t>(orv,cast<int64_t>(4ULL)))) | shl<uint32_t>(andv,cast<int64_t>(5ULL))))) & affected)))));
allegrex_CPU_eatPfx(c);
return ;
break;}
case cast<uint32_t>(2ULL):case cast<uint32_t>(3ULL):{
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < n);i++){
d[i] = allegrex_vfpuMinMax(s[i],t[i],(sub_ == cast<uint32_t>(3ULL)));
}
}break;}
case cast<uint32_t>(6ULL):{
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < n);i++){
if (((s[i] != s[i]) || (t[i] != t[i]))) {
d[i] = cast<float>(0ULL);
}
else if ((s[i] >= t[i])) {
d[i] = cast<float>(1ULL);
}
else {
d[i] = cast<float>(0ULL);
}
}
}allegrex_CPU_vwrite(c,d,vd,n);
allegrex_CPU_eatPfx(c);
return ;
break;}
case cast<uint32_t>(7ULL):{
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < n);i++){
if (((s[i] != s[i]) || (t[i] != t[i]))) {
d[i] = cast<float>(0ULL);
}
else if ((s[i] < t[i])) {
d[i] = cast<float>(1ULL);
}
else {
d[i] = cast<float>(0ULL);
}
}
}allegrex_CPU_vwrite(c,d,vd,n);
allegrex_CPU_eatPfx(c);
return ;
break;}
case cast<uint32_t>(5ULL):{
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < n);i++){
{
if ((s[i] < t[i])){
d[i] = cast<float>(-cast<int64_t>(1ULL));
}
else if ((s[i] > t[i])){
d[i] = cast<float>(1ULL);
}
else {
d[i] = cast<float>(0ULL);
}
}
tmp57:;
}
}break;}
default:{
allegrex_CPU_vfpuUnimpl(c,w,op);
return ;
break;}
}}
break;}
}}
allegrex_CPU_applyPfxD(c,(&d),n);
allegrex_CPU_vwrite(c,d,vd,n);
allegrex_CPU_eatPfx(c);
}
}
// tools/cpu/allegrex/vfpu.go:968:1
bool allegrex_isInf32(float f){
{
return (cast<uint32_t>((go_math_Float32bits(f) & cast<uint32_t>(2147483647ULL))) == cast<uint32_t>(2139095040ULL));
}
}
// tools/cpu/allegrex/vfpu.go:972:1
float allegrex_vfpuMinMax(float s,float t,bool wantMax){
{
if (((((s != s) || (t != t)) || allegrex_isInf32(s)) || allegrex_isInf32(t))) {
auto tmp58 = std::make_tuple(cast<int32_t>(go_math_Float32bits(s)),cast<int32_t>(go_math_Float32bits(t)));
int32_t si = std::get<0>(tmp58);
int32_t ti = std::get<1>(tmp58);
bool flip = ((si < cast<int32_t>(0ULL)) && (ti < cast<int32_t>(0ULL)));
if ((((si < ti)) != ((wantMax != flip)))) {
return s;
}
return t;
}
if ((((s < t)) != wantMax)) {
return s;
}
return t;
}
}
// tools/cpu/allegrex/vfpu.go:988:1
void allegrex_CPU_vfpu4(allegrex_CPU* c,uint32_t w,uint32_t rs,uint32_t rt){
{
uint32_t n = allegrex_vfpuVecN(w);
uint32_t vd = cast<uint32_t>((w & cast<uint32_t>(127ULL)));
uint32_t vs = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(8ULL))) & cast<uint32_t>(127ULL)));
std::array<float,4> d={};
{
switch(rs){
case cast<uint32_t>(0ULL):{
std::array<float,4> s = allegrex_CPU_vread(c,vs,n);
allegrex_CPU_applyPfxS(c,(&s),n);
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < n);i++){
float x = s[i];
{
switch(rt){
case cast<uint32_t>(0ULL):{
d[i] = x;
break;}
case cast<uint32_t>(1ULL):{
d[i] = go_math_Float32frombits((go_math_Float32bits(x) & ~(cast<uint32_t>(2147483648ULL))));
break;}
case cast<uint32_t>(2ULL):{
d[i] = go_math_Float32frombits(cast<uint32_t>((go_math_Float32bits(x) ^ cast<uint32_t>(2147483648ULL))));
break;}
case cast<uint32_t>(4ULL):{
if ((x <= cast<float>(0ULL))) {
d[i] = cast<float>(0ULL);
}
else if ((x > cast<float>(1ULL))) {
d[i] = cast<float>(1ULL);
}
else {
d[i] = x;
}
break;}
case cast<uint32_t>(5ULL):{
if ((x < cast<float>(-cast<int64_t>(1ULL)))) {
d[i] = cast<float>(-cast<int64_t>(1ULL));
}
else if ((x > cast<float>(1ULL))) {
d[i] = cast<float>(1ULL);
}
else {
d[i] = x;
}
break;}
case cast<uint32_t>(16ULL):{
d[i] = (cast<float>(1ULL) / x);
break;}
case cast<uint32_t>(17ULL):{
d[i] = cast<float>((cast<double>(1ULL) / go_math_Sqrt(cast<double>(x))));
break;}
case cast<uint32_t>(18ULL):{
d[i] = allegrex_vfpuSin(x);
break;}
case cast<uint32_t>(19ULL):{
d[i] = allegrex_vfpuCos(x);
break;}
case cast<uint32_t>(20ULL):{
d[i] = cast<float>(go_math_Exp2(cast<double>(x)));
break;}
case cast<uint32_t>(21ULL):{
d[i] = cast<float>(go_math_Log2(cast<double>(x)));
break;}
case cast<uint32_t>(22ULL):{
d[i] = cast<float>(go_math_Abs(go_math_Sqrt(cast<double>(x))));
break;}
case cast<uint32_t>(23ULL):{
d[i] = cast<float>((go_math_Asin(cast<double>(x)) / ((go_math_Pi / cast<double>(2ULL)))));
break;}
case cast<uint32_t>(24ULL):{
d[i] = cast<float>(-((cast<float>(1ULL) / x)));
break;}
case cast<uint32_t>(26ULL):{
d[i] = cast<float>(-allegrex_vfpuSin(x));
break;}
case cast<uint32_t>(28ULL):{
d[i] = cast<float>((cast<double>(1ULL) / go_math_Exp2(cast<double>(x))));
break;}
default:{
{
switch(rt){
case cast<uint32_t>(3ULL):{
uint32_t off = cast<uint32_t>((vd & cast<uint32_t>(3ULL)));
if ((n < cast<uint32_t>(3ULL))) {
off = cast<uint32_t>((vd & cast<uint32_t>(1ULL)));
}
{uint32_t j = cast<uint32_t>(0ULL);for (;(j < n);j++){
if ((j == off)) {
d[j] = cast<float>(1ULL);
}
else {
d[j] = cast<float>(0ULL);
}
}
}break;}
case cast<uint32_t>(6ULL):{
{uint32_t j = cast<uint32_t>(0ULL);for (;(j < n);j++){
d[j] = cast<float>(0ULL);
}
}break;}
case cast<uint32_t>(7ULL):{
{uint32_t j = cast<uint32_t>(0ULL);for (;(j < n);j++){
d[j] = cast<float>(1ULL);
}
}break;}
default:{
allegrex_CPU_vfpuUnimpl(c,w,cast<uint32_t>(52ULL));
return ;
break;}
}}
allegrex_CPU_applyPfxD(c,(&d),n);
allegrex_CPU_vwrite(c,d,vd,n);
allegrex_CPU_eatPfx(c);
return ;
break;}
}}
}
}allegrex_CPU_applyPfxD(c,(&d),n);
allegrex_CPU_vwrite(c,d,vd,n);
allegrex_CPU_eatPfx(c);
break;}
case cast<uint32_t>(1ULL):{
{
switch(rt){
case cast<uint32_t>(27ULL):{
std::array<uint32_t,4> s = allegrex_CPU_vreadBits(c,vs,n);
std::array<uint32_t,4> bits={};
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < n);i++){
bits[cast<uint32_t>((cast<uint32_t>(2ULL) * i))] = shl<uint32_t>(s[i],cast<int64_t>(16ULL));
bits[cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(2ULL) * i)) + cast<uint32_t>(1ULL)))] = cast<uint32_t>((s[i] & cast<uint32_t>(4294901760ULL)));
}
}allegrex_CPU_vwriteBits(c,bits,vd,cast<uint32_t>((cast<uint32_t>(2ULL) * n)));
allegrex_CPU_eatPfx(c);
break;}
case cast<uint32_t>(26ULL):{
std::array<uint32_t,4> s = allegrex_CPU_vreadBits(c,vs,n);
std::array<uint32_t,4> bits={};
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < n);i++){
bits[cast<uint32_t>((cast<uint32_t>(2ULL) * i))] = shl<uint32_t>((cast<uint32_t>((s[i] & cast<uint32_t>(65535ULL)))),cast<int64_t>(15ULL));
bits[cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(2ULL) * i)) + cast<uint32_t>(1ULL)))] = shl<uint32_t>((shr<uint32_t>(s[i],cast<int64_t>(16ULL))),cast<int64_t>(15ULL));
}
}allegrex_CPU_vwriteBits(c,bits,vd,cast<uint32_t>((cast<uint32_t>(2ULL) * n)));
allegrex_CPU_eatPfx(c);
break;}
case cast<uint32_t>(30ULL):case cast<uint32_t>(31ULL):{
std::array<uint32_t,4> s = allegrex_CPU_vreadBits(c,vs,n);
std::array<uint32_t,4> bits={};
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < n);i += cast<uint32_t>(2ULL)){
uint32_t lo={};
uint32_t hi={};
if ((rt == cast<uint32_t>(31ULL))) {
lo = shr<uint32_t>(s[i],cast<int64_t>(16ULL));
hi = shr<uint32_t>(s[cast<uint32_t>((i + cast<uint32_t>(1ULL)))],cast<int64_t>(16ULL));
}
else {
if ((cast<int32_t>(s[i]) > cast<int32_t>(0ULL))) {
lo = shr<uint32_t>(s[i],cast<int64_t>(15ULL));
}
if ((cast<int32_t>(s[cast<uint32_t>((i + cast<uint32_t>(1ULL)))]) > cast<int32_t>(0ULL))) {
hi = shr<uint32_t>(s[cast<uint32_t>((i + cast<uint32_t>(1ULL)))],cast<int64_t>(15ULL));
}
}
bits[divi<uint32_t>(i,cast<uint32_t>(2ULL))] = cast<uint32_t>((cast<uint32_t>((lo & cast<uint32_t>(65535ULL))) | shl<uint32_t>(hi,cast<int64_t>(16ULL))));
}
}allegrex_CPU_vwriteBits(c,bits,vd,divi<uint32_t>(n,cast<uint32_t>(2ULL)));
allegrex_CPU_eatPfx(c);
break;}
case cast<uint32_t>(28ULL):case cast<uint32_t>(29ULL):{
std::array<uint32_t,4> s = allegrex_CPU_vreadBits(c,vs,n);
std::array<uint32_t,4> bits={};
uint32_t out={};
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < n);i++){
uint32_t b={};
if ((rt == cast<uint32_t>(29ULL))) {
b = shr<uint32_t>(s[i],cast<int64_t>(24ULL));
}
else if ((cast<int32_t>(s[i]) > cast<int32_t>(0ULL))) {
b = shr<uint32_t>(s[i],cast<int64_t>(23ULL));
if ((b > cast<uint32_t>(255ULL))) {
b = cast<uint32_t>(255ULL);
}
}
out |= shl<uint32_t>(b,(cast<uint32_t>((cast<uint32_t>(8ULL) * i))));
}
}bits[cast<int64_t>(0ULL)] = out;
allegrex_CPU_vwriteBits(c,bits,vd,cast<uint32_t>(1ULL));
allegrex_CPU_eatPfx(c);
break;}
default:{
allegrex_CPU_vfpuUnimpl(c,w,cast<uint32_t>(52ULL));
break;}
}}
break;}
case cast<uint32_t>(2ULL):{
{
switch(rt){
case cast<uint32_t>(2ULL):{
std::array<float,4> s = allegrex_CPU_vread(c,vs,n);
allegrex_CPU_applyPfxS(c,(&s),n);
auto tmp59 = std::make_tuple((s[cast<int64_t>(0ULL)] + s[cast<int64_t>(1ULL)]),(s[cast<int64_t>(0ULL)] - s[cast<int64_t>(1ULL)]));
d[cast<int64_t>(0ULL)] = std::get<0>(tmp59);
d[cast<int64_t>(1ULL)] = std::get<1>(tmp59);
if ((n == cast<uint32_t>(4ULL))) {
auto tmp60 = std::make_tuple((s[cast<int64_t>(2ULL)] + s[cast<int64_t>(3ULL)]),(s[cast<int64_t>(2ULL)] - s[cast<int64_t>(3ULL)]));
d[cast<int64_t>(2ULL)] = std::get<0>(tmp60);
d[cast<int64_t>(3ULL)] = std::get<1>(tmp60);
}
break;}
case cast<uint32_t>(3ULL):{
std::array<float,4> s = allegrex_CPU_vread(c,vs,n);
allegrex_CPU_applyPfxS(c,(&s),n);
auto tmp61 = std::make_tuple((s[cast<int64_t>(0ULL)] + s[cast<int64_t>(2ULL)]),(s[cast<int64_t>(1ULL)] + s[cast<int64_t>(3ULL)]));
d[cast<int64_t>(0ULL)] = std::get<0>(tmp61);
d[cast<int64_t>(1ULL)] = std::get<1>(tmp61);
auto tmp62 = std::make_tuple((s[cast<int64_t>(0ULL)] - s[cast<int64_t>(2ULL)]),(s[cast<int64_t>(1ULL)] - s[cast<int64_t>(3ULL)]));
d[cast<int64_t>(2ULL)] = std::get<0>(tmp62);
d[cast<int64_t>(3ULL)] = std::get<1>(tmp62);
break;}
case cast<uint32_t>(4ULL):{
std::array<float,4> s = allegrex_CPU_vread(c,vs,n);
allegrex_CPU_applyPfxS(c,(&s),n);
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < n);i++){
d[i] = (cast<float>(1ULL) - s[i]);
}
}break;}
case cast<uint32_t>(6ULL):case cast<uint32_t>(7ULL):{
std::array<float,4> s = allegrex_CPU_vread(c,vs,n);
allegrex_CPU_applyPfxS(c,(&s),n);
float sum={};
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < n);i++){
sum += s[i];
}
}if ((rt == cast<uint32_t>(7ULL))) {
sum /= cast<float>(n);
}
d[cast<int64_t>(0ULL)] = sum;
allegrex_CPU_applyPfxD(c,(&d),cast<uint32_t>(1ULL));
allegrex_CPU_vwrite(c,d,vd,cast<uint32_t>(1ULL));
allegrex_CPU_eatPfx(c);
return ;
break;}
case cast<uint32_t>(10ULL):{
std::array<float,4> s = allegrex_CPU_vread(c,vs,n);
allegrex_CPU_applyPfxS(c,(&s),n);
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < n);i++){
{
if ((s[i] > cast<float>(0ULL))){
d[i] = cast<float>(1ULL);
}
else if ((s[i] < cast<float>(0ULL))){
d[i] = cast<float>(-cast<int64_t>(1ULL));
}
else {
d[i] = cast<float>(0ULL);
}
}
tmp63:;
}
}break;}
case cast<uint32_t>(16ULL):{
uint32_t imm = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(8ULL))) & cast<uint32_t>(127ULL)));
uint32_t v={};
if ((imm < cast<uint32_t>(16ULL))) {
v = c->VfpuCtrl[imm];
}
c->V[cast<uint32_t>((allegrex_vfpuSingle(vd) & cast<uint32_t>(127ULL)))] = v;
return ;
break;}
case cast<uint32_t>(17ULL):{
uint32_t imm = cast<uint32_t>((w & cast<uint32_t>(127ULL)));
if (((imm < cast<uint32_t>(16ULL)) && (allegrex_vfpuCtrlMask[imm] != cast<uint32_t>(0ULL)))) {
c->VfpuCtrl[imm] = cast<uint32_t>((c->V[cast<uint32_t>((allegrex_vfpuSingle(vs) & cast<uint32_t>(127ULL)))] & allegrex_vfpuCtrlMask[imm]));
}
return ;
break;}
default:{
allegrex_CPU_vfpuUnimpl(c,w,cast<uint32_t>(52ULL));
return ;
break;}
}}
allegrex_CPU_applyPfxD(c,(&d),n);
allegrex_CPU_vwrite(c,d,vd,n);
allegrex_CPU_eatPfx(c);
break;}
case cast<uint32_t>(3ULL):{
float cst = allegrex_vcstConstants[cast<uint32_t>((rt & cast<uint32_t>(31ULL)))];
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < n);i++){
d[i] = cst;
}
}allegrex_CPU_applyPfxD(c,(&d),n);
allegrex_CPU_vwrite(c,d,vd,n);
allegrex_CPU_eatPfx(c);
break;}
case cast<uint32_t>(16ULL):case cast<uint32_t>(17ULL):case cast<uint32_t>(18ULL):case cast<uint32_t>(19ULL):{
std::array<float,4> s = allegrex_CPU_vread(c,vs,n);
allegrex_CPU_applyPfxS(c,(&s),n);
double mult = cast<double>(shl<uint64_t>(cast<uint64_t>(1ULL),(cast<uint32_t>((rt & cast<uint32_t>(31ULL))))));
std::array<uint32_t,4> bits={};
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < n);i++){
if ((s[i] != s[i])) {
bits[i] = cast<uint32_t>(2147483647ULL);
continue;
}
double sv = (cast<double>(s[i]) * mult);
{
if ((sv > cast<double>(cast<double>(2147483647ULL)))){
bits[i] = cast<uint32_t>(2147483647ULL);
}
else if ((sv <= cast<double>(cast<double>(-2147483648ULL)))){
bits[i] = cast<uint32_t>(2147483648ULL);
}
else {
double r={};
{
switch(rs){
case cast<uint32_t>(16ULL):{
r = go_math_RoundToEven(sv);
break;}
case cast<uint32_t>(17ULL):{
r = go_math_Trunc(sv);
break;}
case cast<uint32_t>(18ULL):{
r = go_math_Ceil(sv);
break;}
case cast<uint32_t>(19ULL):{
r = go_math_Floor(sv);
break;}
}}
bits[i] = cast<uint32_t>(cast<int32_t>(r));
}
}
tmp64:;
}
}allegrex_CPU_vwriteBits(c,bits,vd,n);
allegrex_CPU_eatPfx(c);
break;}
case cast<uint32_t>(20ULL):{
std::array<uint32_t,4> s = allegrex_CPU_vreadBits(c,vs,n);
float mult = (cast<float>(cast<float>(1ULL)) / cast<float>(shl<uint64_t>(cast<uint64_t>(1ULL),(cast<uint32_t>((rt & cast<uint32_t>(31ULL)))))));
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < n);i++){
d[i] = (cast<float>(cast<int32_t>(s[i])) * mult);
}
}allegrex_CPU_applyPfxD(c,(&d),n);
allegrex_CPU_vwrite(c,d,vd,n);
allegrex_CPU_eatPfx(c);
break;}
case cast<uint32_t>(21ULL):{
uint32_t tf = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(19ULL))) & cast<uint32_t>(1ULL)));
uint32_t imm3 = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(7ULL)));
std::array<float,4> s = allegrex_CPU_vread(c,vs,n);
allegrex_CPU_applyPfxS(c,(&s),n);
d = allegrex_CPU_vread(c,vd,n);
allegrex_CPU_applyPfxT(c,(&d),n);
uint32_t cc = c->VfpuCtrl[cast<int64_t>(3ULL)];
if ((imm3 < cast<uint32_t>(6ULL))) {
if ((cast<uint32_t>(((shr<uint32_t>(cc,imm3)) & cast<uint32_t>(1ULL))) == cast<uint32_t>((cast<uint32_t>(1ULL) - tf)))) {
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < n);i++){
d[i] = s[i];
}
}}
}
else if ((imm3 == cast<uint32_t>(6ULL))) {
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < n);i++){
if ((cast<uint32_t>(((shr<uint32_t>(cc,i)) & cast<uint32_t>(1ULL))) == cast<uint32_t>((cast<uint32_t>(1ULL) - tf)))) {
d[i] = s[i];
}
}
}}
allegrex_CPU_applyPfxD(c,(&d),n);
allegrex_CPU_vwrite(c,d,vd,n);
allegrex_CPU_eatPfx(c);
break;}
default:{
allegrex_CPU_vfpuUnimpl(c,w,cast<uint32_t>(52ULL));
break;}
}}
}
}
// tools/cpu/allegrex/vfpu.go:1298:1
void allegrex_CPU_vfpu5(allegrex_CPU* c,uint32_t w){
{
{
switch(cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(24ULL))) & cast<uint32_t>(3ULL)))){
case cast<uint32_t>(0ULL):{
c->VfpuCtrl[cast<int64_t>(0ULL)] = cast<uint32_t>((w & cast<uint32_t>(1048575ULL)));
break;}
case cast<uint32_t>(1ULL):{
c->VfpuCtrl[cast<int64_t>(1ULL)] = cast<uint32_t>((w & cast<uint32_t>(1048575ULL)));
break;}
case cast<uint32_t>(2ULL):{
c->VfpuCtrl[cast<int64_t>(2ULL)] = cast<uint32_t>((w & cast<uint32_t>(4095ULL)));
break;}
case cast<uint32_t>(3ULL):{
uint32_t vt = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(127ULL)));
std::array<float,4> d={};
if ((cast<uint32_t>((w & cast<uint32_t>(8388608ULL))) != cast<uint32_t>(0ULL))) {
d[cast<int64_t>(0ULL)] = allegrex_float16to32(cast<uint16_t>(w));
}
else {
d[cast<int64_t>(0ULL)] = cast<float>(cast<int32_t>(cast<int16_t>(w)));
}
allegrex_CPU_applyPfxD(c,(&d),cast<uint32_t>(1ULL));
allegrex_CPU_vwrite(c,d,vt,cast<uint32_t>(1ULL));
allegrex_CPU_eatPfx(c);
break;}
}}
}
}
// tools/cpu/allegrex/vfpu.go:1323:1
void allegrex_CPU_vfpuMatrix(allegrex_CPU* c,uint32_t w,uint32_t rt){
{
uint32_t n = allegrex_vfpuVecN(w);
uint32_t vd = cast<uint32_t>((w & cast<uint32_t>(127ULL)));
uint32_t vs = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(8ULL))) & cast<uint32_t>(127ULL)));
uint32_t vt = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(127ULL)));
{
switch(cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(21ULL))) & cast<uint32_t>(31ULL)))){
case cast<uint32_t>(0ULL):case cast<uint32_t>(1ULL):case cast<uint32_t>(2ULL):case cast<uint32_t>(3ULL):{
std::array<float,16> s = allegrex_CPU_mread(c,vs,n);
std::array<float,16> t = allegrex_CPU_mread(c,vt,n);
std::array<float,16> d={};
{uint32_t a = cast<uint32_t>(0ULL);for (;(a < n);a++){
{uint32_t b = cast<uint32_t>(0ULL);for (;(b < n);b++){
float sum={};
{uint32_t k = cast<uint32_t>(0ULL);for (;(k < n);k++){
sum += (s[cast<uint32_t>((cast<uint32_t>((b * cast<uint32_t>(4ULL))) + k))] * t[cast<uint32_t>((cast<uint32_t>((a * cast<uint32_t>(4ULL))) + k))]);
}
}d[cast<uint32_t>((cast<uint32_t>((a * cast<uint32_t>(4ULL))) + b))] = sum;
}
}}
}allegrex_CPU_mwrite(c,d,vd,n);
allegrex_CPU_eatPfx(c);
break;}
case cast<uint32_t>(16ULL):case cast<uint32_t>(17ULL):case cast<uint32_t>(18ULL):case cast<uint32_t>(19ULL):{
std::array<float,16> s = allegrex_CPU_mread(c,vs,n);
std::array<float,4> t = allegrex_CPU_vread(c,vt,cast<uint32_t>(1ULL));
std::array<float,16> d={};
{uint32_t a = cast<uint32_t>(0ULL);for (;(a < n);a++){
{uint32_t b = cast<uint32_t>(0ULL);for (;(b < n);b++){
d[cast<uint32_t>((cast<uint32_t>((a * cast<uint32_t>(4ULL))) + b))] = (s[cast<uint32_t>((cast<uint32_t>((a * cast<uint32_t>(4ULL))) + b))] * t[cast<int64_t>(0ULL)]);
}
}}
}allegrex_CPU_mwrite(c,d,vd,n);
allegrex_CPU_eatPfx(c);
break;}
case cast<uint32_t>(20ULL):case cast<uint32_t>(21ULL):case cast<uint32_t>(22ULL):case cast<uint32_t>(23ULL):{
std::array<float,4> s = allegrex_CPU_vread(c,vs,n);
std::array<float,4> t = allegrex_CPU_vread(c,vt,n);
std::array<float,4> d={};
{
switch(n){
case cast<uint32_t>(3ULL):{
d[cast<int64_t>(0ULL)] = ((s[cast<int64_t>(1ULL)] * t[cast<int64_t>(2ULL)]) - (s[cast<int64_t>(2ULL)] * t[cast<int64_t>(1ULL)]));
d[cast<int64_t>(1ULL)] = ((s[cast<int64_t>(2ULL)] * t[cast<int64_t>(0ULL)]) - (s[cast<int64_t>(0ULL)] * t[cast<int64_t>(2ULL)]));
d[cast<int64_t>(2ULL)] = ((s[cast<int64_t>(0ULL)] * t[cast<int64_t>(1ULL)]) - (s[cast<int64_t>(1ULL)] * t[cast<int64_t>(0ULL)]));
break;}
case cast<uint32_t>(4ULL):{
d[cast<int64_t>(0ULL)] = ((((s[cast<int64_t>(0ULL)] * t[cast<int64_t>(3ULL)]) + (s[cast<int64_t>(1ULL)] * t[cast<int64_t>(2ULL)])) - (s[cast<int64_t>(2ULL)] * t[cast<int64_t>(1ULL)])) + (s[cast<int64_t>(3ULL)] * t[cast<int64_t>(0ULL)]));
d[cast<int64_t>(1ULL)] = ((((cast<float>(-s[cast<int64_t>(0ULL)]) * t[cast<int64_t>(2ULL)]) + (s[cast<int64_t>(1ULL)] * t[cast<int64_t>(3ULL)])) + (s[cast<int64_t>(2ULL)] * t[cast<int64_t>(0ULL)])) + (s[cast<int64_t>(3ULL)] * t[cast<int64_t>(1ULL)]));
d[cast<int64_t>(2ULL)] = ((((s[cast<int64_t>(0ULL)] * t[cast<int64_t>(1ULL)]) - (s[cast<int64_t>(1ULL)] * t[cast<int64_t>(0ULL)])) + (s[cast<int64_t>(2ULL)] * t[cast<int64_t>(3ULL)])) + (s[cast<int64_t>(3ULL)] * t[cast<int64_t>(2ULL)]));
d[cast<int64_t>(3ULL)] = ((((cast<float>(-s[cast<int64_t>(0ULL)]) * t[cast<int64_t>(0ULL)]) - (s[cast<int64_t>(1ULL)] * t[cast<int64_t>(1ULL)])) - (s[cast<int64_t>(2ULL)] * t[cast<int64_t>(2ULL)])) + (s[cast<int64_t>(3ULL)] * t[cast<int64_t>(3ULL)]));
break;}
default:{
allegrex_CPU_vfpuUnimpl(c,w,cast<uint32_t>(60ULL));
return ;
break;}
}}
allegrex_CPU_applyPfxD(c,(&d),n);
allegrex_CPU_vwrite(c,d,vd,n);
allegrex_CPU_eatPfx(c);
break;}
case cast<uint32_t>(28ULL):{
{
switch(cast<uint32_t>((rt & cast<uint32_t>(15ULL)))){
case cast<uint32_t>(0ULL):{
allegrex_CPU_mwrite(c,allegrex_CPU_mread(c,vs,n),vd,n);
break;}
case cast<uint32_t>(3ULL):{
std::array<float,16> d={};
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < n);i++){
d[cast<uint32_t>((cast<uint32_t>((i * cast<uint32_t>(4ULL))) + i))] = cast<float>(1ULL);
}
}allegrex_CPU_mwrite(c,d,vd,n);
break;}
case cast<uint32_t>(6ULL):{
allegrex_CPU_mwrite(c,std::array<float,16>{},vd,n);
break;}
case cast<uint32_t>(7ULL):{
std::array<float,16> d={};
{uint32_t a = cast<uint32_t>(0ULL);for (;(a < n);a++){
{uint32_t b = cast<uint32_t>(0ULL);for (;(b < n);b++){
d[cast<uint32_t>((cast<uint32_t>((a * cast<uint32_t>(4ULL))) + b))] = cast<float>(1ULL);
}
}}
}allegrex_CPU_mwrite(c,d,vd,n);
break;}
default:{
allegrex_CPU_vfpuUnimpl(c,w,cast<uint32_t>(60ULL));
return ;
break;}
}}
allegrex_CPU_eatPfx(c);
break;}
case cast<uint32_t>(29ULL):{
uint32_t imm = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(31ULL)));
float x = allegrex_CPU_vread(c,vs,cast<uint32_t>(1ULL))[cast<int64_t>(0ULL)];
auto tmp65 = std::make_tuple(allegrex_vfpuSin(x),allegrex_vfpuCos(x));
float sin = std::get<0>(tmp65);
float cos = std::get<1>(tmp65);
if ((cast<uint32_t>((imm & cast<uint32_t>(16ULL))) != cast<uint32_t>(0ULL))) {
sin = cast<float>(-sin);
}
uint32_t sinLane = cast<uint32_t>(((shr<uint32_t>(imm,cast<int64_t>(2ULL))) & cast<uint32_t>(3ULL)));
uint32_t cosLane = cast<uint32_t>((imm & cast<uint32_t>(3ULL)));
std::array<float,4> d={};
if ((sinLane == cosLane)) {
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < n);i++){
d[i] = sin;
}
}}
else {
d[sinLane] = sin;
}
d[cosLane] = cos;
allegrex_CPU_applyPfxD(c,(&d),n);
allegrex_CPU_vwrite(c,d,vd,n);
allegrex_CPU_eatPfx(c);
break;}
default:{
uint32_t ins = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(23ULL))) & cast<uint32_t>(3ULL)));
if ((ins == cast<uint32_t>(0ULL))) {
allegrex_CPU_vfpuUnimpl(c,w,cast<uint32_t>(60ULL));
return ;
}
uint32_t side = cast<uint32_t>((ins + cast<uint32_t>(1ULL)));
std::array<float,16> s = allegrex_CPU_mread(c,vs,side);
std::array<float,4> t = allegrex_CPU_vread(c,vt,n);
bool hom = (ins >= n);
uint32_t tn = n;
if ((side < tn)) {
tn = side;
}
std::array<float,4> d={};
{uint32_t i = cast<uint32_t>(0ULL);for (;(i <= ins);i++){
float sum={};
{uint32_t k = cast<uint32_t>(0ULL);for (;(k < tn);k++){
sum += (s[cast<uint32_t>((cast<uint32_t>((i * cast<uint32_t>(4ULL))) + k))] * t[k]);
}
}if (hom) {
sum += s[cast<uint32_t>((cast<uint32_t>((i * cast<uint32_t>(4ULL))) + ins))];
}
d[i] = sum;
}
}allegrex_CPU_applyPfxD(c,(&d),side);
allegrex_CPU_vwrite(c,d,vd,side);
allegrex_CPU_eatPfx(c);
break;}
}}
}
}
// tools/platform/psp/debug.go:47:1
bool psp_PixelEvent_Rejected(psp_PixelEvent e){
{
return (!e.Drawn);
}
}
// tools/platform/psp/debug.go:53:1
void psp_Machine_pixelEvent(psp_Machine* m,int64_t x,int64_t y,psp_PixelEvent ev){
{
{
if (ev.Drawn){
m->geCnt.frags++;
}
else if (ev.ZReject){
m->geCnt.zKilled++;
}
else if (ev.AlphaReject){
m->geCnt.aKilled++;
}
else if (ev.StencilReject){
m->geCnt.sKilled++;
}
else if (ev.ScissorReject){
m->geCnt.scissored++;
}
}
tmp1:;
if (((bool(m->OnPixel) && (x >= cast<int64_t>(0ULL))) && (y >= cast<int64_t>(0ULL)))) {
m->OnPixel(cast<uint32_t>(x),cast<uint32_t>(y),ev);
}
}
}
// tools/platform/psp/debug.go:80:1
psp_Result psp_Machine_RunStopAfterGeCommand(psp_Machine* m,int64_t k,uint64_t budget){
{
auto tmp2 = std::make_tuple(k,cast<int64_t>(0ULL));
m->geLimit = std::get<0>(tmp2);
m->geCount = std::get<1>(tmp2);
psp_Result r = psp_Machine_Run(m,budget);
auto tmp3 = std::make_tuple(cast<int64_t>(0ULL),cast<int64_t>(0ULL));
m->geLimit = std::get<0>(tmp3);
m->geCount = std::get<1>(tmp3);
return r;
}
}
// tools/platform/psp/debug.go:88:1
int64_t psp_Machine_GeCommands(psp_Machine* m){
{
return m->geCount;
}
}
// tools/platform/psp/debug.go:95:1
std::tuple<uint32_t,uint32_t,uint32_t,bool> psp_Machine_RenderTarget(psp_Machine* m){
uint32_t addr{};
uint32_t stride{};
uint32_t format{};
bool ok{};
{
if (((!m->geSt) || (m->geSt->fbStride == cast<uint32_t>(0ULL)))) {
return {cast<uint32_t>(0ULL),cast<uint32_t>(0ULL),cast<uint32_t>(0ULL),false};
}
return {psp_geState_fbAddress(m->geSt),m->geSt->fbStride,m->geSt->fbFmt,true};
}
}
// tools/platform/psp/debug.go:106:1
std::tuple<uint32_t,uint32_t,uint32_t> psp_Machine_Scanout(psp_Machine* m){
uint32_t addr{};
uint32_t stride{};
uint32_t format{};
{
addr = m->fbAddr;
if ((addr == cast<uint32_t>(0ULL))) {
addr = cast<uint32_t>(67108864ULL);
}
stride = m->fbWidth;
if ((stride == cast<uint32_t>(0ULL))) {
stride = cast<uint32_t>(480ULL);
}
return {addr,stride,m->fbFormat};
}
}
// tools/platform/psp/debug.go:119:1
std::tuple<uint32_t,uint32_t,bool> psp_Machine_DepthTarget(psp_Machine* m){
uint32_t addr{};
uint32_t stride{};
bool ok{};
{
if (((!m->geSt) || (m->geSt->zStride == cast<uint32_t>(0ULL)))) {
return {cast<uint32_t>(0ULL),cast<uint32_t>(0ULL),false};
}
return {psp_geState_zAddress(m->geSt),m->geSt->zStride,true};
}
}
// tools/platform/psp/debug.go:129:1
std::tuple<image_RGBA*,Error> psp_Machine_RenderSurface(psp_Machine* m,uint32_t addr,uint32_t stride,uint32_t format,int64_t w,int64_t h){
{
if (((((w <= cast<int64_t>(0ULL)) || (h <= cast<int64_t>(0ULL))) || (w > cast<int64_t>(4096ULL))) || (h > cast<int64_t>(4096ULL)))) {
return {{},go_fmt_Errorf(std::string("psp: render target size %dx%d out of range",42),w,h)};
}
if ((format > cast<uint32_t>(3ULL))) {
return {{},go_fmt_Errorf(std::string("psp: 0x%X is not a display pixel format",39),format)};
}
if ((stride == cast<uint32_t>(0ULL))) {
stride = cast<uint32_t>(480ULL);
}
image_RGBA* img = go_image_NewRGBA(go_image_Rect(cast<int64_t>(0ULL),cast<int64_t>(0ULL),w,h));
{int64_t y = cast<int64_t>(0ULL);for (;(y < h);y++){
{int64_t x = cast<int64_t>(0ULL);for (;(x < w);x++){
image_RGBA_Set(img,x,y,psp_Machine_readPixelFmt(m,addr,stride,format,cast<uint32_t>(x),cast<uint32_t>(y)));
}
}}
}return {img,{}};
}
}
// tools/platform/psp/debug.go:182:1
Slice<std::string> psp_TextureFormats(){
{
return append(Slice<std::string>{},psp_textureFormats);
}
}
// tools/platform/psp/debug.go:185:1
std::tuple<uint32_t,bool> psp_TextureFormat(std::string name){
{
{auto&& tmp4 = psp_textureFormats;
for(int64_t tmp5=0;tmp5<len(tmp4);++tmp5){
auto i=tmp5;auto f=tmp4[tmp5];if ((f == name)) {
return {cast<uint32_t>(i),true};
}
}}
return {cast<uint32_t>(0ULL),false};
}
}
// tools/platform/psp/debug.go:245:1
Slice<psp_MemRegion> psp_Machine_MemRegions(psp_Machine* m){
{
return Slice<psp_MemRegion>{psp_MemRegion{std::string("scratchpad",10),cast<uint32_t>(65536ULL),cast<uint32_t>(16384ULL)},psp_MemRegion{std::string("vram",4),cast<uint32_t>(67108864ULL),cast<uint32_t>(2097152ULL)},psp_MemRegion{std::string("ram",3),cast<uint32_t>(134217728ULL),cast<uint32_t>(33554432ULL)}};
}
}
// tools/platform/psp/debug.go:254:1
psp_Volume* psp_Machine_Volume(psp_Machine* m){
{
return m->vol;
}
}
// tools/platform/psp/elf.go:76:1
std::tuple<psp_Module*,Error> psp_ParseELF(Slice<uint8_t> b){
{
if (((len(b) < cast<int64_t>(52ULL)) || (cast<std::string>(sub(b,cast<int64_t>(0ULL),cast<int64_t>(4ULL))) != std::string("\177ELF",4)))) {
return {{},go_fmt_Errorf(std::string("elf: bad magic",14))};
}
if (((b[cast<int64_t>(4ULL)] != cast<uint8_t>(1ULL)) || (b[cast<int64_t>(5ULL)] != cast<uint8_t>(1ULL)))) {
return {{},go_fmt_Errorf(std::string("elf: not 32-bit little-endian",29))};
}
uint16_t eType = le_Uint16(sub(b,cast<int64_t>(16ULL),len(b)));
uint16_t eMachine = le_Uint16(sub(b,cast<int64_t>(18ULL),len(b)));
if ((eMachine != cast<uint16_t>(8ULL))) {
return {{},go_fmt_Errorf(std::string("elf: machine %d is not MIPS",27),eMachine)};
}
uint32_t entry = le_Uint32(sub(b,cast<int64_t>(24ULL),len(b)));
uint32_t phoff = le_Uint32(sub(b,cast<int64_t>(28ULL),len(b)));
int64_t phentsize = cast<int64_t>(le_Uint16(sub(b,cast<int64_t>(42ULL),len(b))));
int64_t phnum = cast<int64_t>(le_Uint16(sub(b,cast<int64_t>(44ULL),len(b))));
psp_Module* m = arenaNew(psp_Module{{},eType,entry,{},{},{},{},{},{}});
{int64_t i = cast<int64_t>(0ULL);for (;(i < phnum);i++){
int64_t ph = cast<int64_t>((cast<int64_t>(phoff) + cast<int64_t>((i * phentsize))));
if ((cast<int64_t>((ph + cast<int64_t>(32ULL))) > len(b))) {
return {{},go_fmt_Errorf(std::string("elf: program header %d out of range",35),i)};
}
uint32_t pType = le_Uint32(sub(b,ph,len(b)));
if ((pType != cast<uint32_t>(1ULL))) {
continue;
}
uint32_t off = le_Uint32(sub(b,cast<int64_t>((ph + cast<int64_t>(4ULL))),len(b)));
uint32_t vaddr = le_Uint32(sub(b,cast<int64_t>((ph + cast<int64_t>(8ULL))),len(b)));
uint32_t filesz = le_Uint32(sub(b,cast<int64_t>((ph + cast<int64_t>(16ULL))),len(b)));
uint32_t memsz = le_Uint32(sub(b,cast<int64_t>((ph + cast<int64_t>(20ULL))),len(b)));
if ((cast<int64_t>((cast<int64_t>(off) + cast<int64_t>(filesz))) > len(b))) {
return {{},go_fmt_Errorf(std::string("elf: segment %d data out of range",33),i)};
}
psp_Segment seg = psp_Segment{vaddr,off,append(Slice<uint8_t>{},sub(b,off,cast<uint32_t>((off + filesz)))),memsz};
m->Segments = append(m->Segments,seg);
}
}if ((len(m->Segments) == cast<int64_t>(0ULL))) {
return {{},go_fmt_Errorf(std::string("elf: no PT_LOAD segments",24))};
}
psp_Module_parseModuleInfo(m,b);
psp_Module_parseRelocs(m,b);
return {m,{}};
}
}
// tools/platform/psp/elf.go:122:1
void psp_Module_parseRelocs(psp_Module* m,Slice<uint8_t> b){
{
uint32_t shoff = le_Uint32(sub(b,cast<int64_t>(32ULL),len(b)));
int64_t shentsize = cast<int64_t>(le_Uint16(sub(b,cast<int64_t>(46ULL),len(b))));
int64_t shnum = cast<int64_t>(le_Uint16(sub(b,cast<int64_t>(48ULL),len(b))));
{int64_t i = cast<int64_t>(0ULL);for (;(i < shnum);i++){
int64_t sh = cast<int64_t>((cast<int64_t>(shoff) + cast<int64_t>((i * shentsize))));
if ((cast<int64_t>((sh + cast<int64_t>(40ULL))) > len(b))) {
return ;
}
if ((le_Uint32(sub(b,cast<int64_t>((sh + cast<int64_t>(4ULL))),len(b))) != cast<uint32_t>(1879048352ULL))) {
continue;
}
int64_t off = cast<int64_t>(le_Uint32(sub(b,cast<int64_t>((sh + cast<int64_t>(16ULL))),len(b))));
int64_t size = cast<int64_t>(le_Uint32(sub(b,cast<int64_t>((sh + cast<int64_t>(20ULL))),len(b))));
{int64_t p = off;for (;((cast<int64_t>((p + cast<int64_t>(8ULL))) <= cast<int64_t>((off + size))) && (cast<int64_t>((p + cast<int64_t>(8ULL))) <= len(b)));p += cast<int64_t>(8ULL)){
m->relocs = append(m->relocs,psp_reloc{le_Uint32(sub(b,p,len(b))),le_Uint32(sub(b,cast<int64_t>((p + cast<int64_t>(4ULL))),len(b)))});
}
}}
}}
}
// tools/platform/psp/elf.go:149:1
void psp_Module_Relocate(psp_Module* m,uint32_t base){
{
if ((len(m->Segments) == cast<int64_t>(0ULL))) {
return ;
}
auto read = [&](int64_t si,uint32_t off)->uint32_t{
Slice<uint8_t> d = m->Segments[si].Data;
if ((cast<int64_t>((cast<int64_t>(off) + cast<int64_t>(4ULL))) > len(d))) {
return cast<uint32_t>(0ULL);
}
return le_Uint32(sub(d,off,len(d)));
}
;
auto write = [&](int64_t si,uint32_t off,uint32_t v)->void{
Slice<uint8_t> d = m->Segments[si].Data;
if ((cast<int64_t>((cast<int64_t>(off) + cast<int64_t>(4ULL))) <= len(d))) {
le_PutUint32(sub(d,off,len(d)),v);
}
}
;
auto segAddr = [&](int64_t si)->uint32_t{
if ((si < len(m->Segments))) {
return cast<uint32_t>((base + m->Segments[si].VAddr));
}
return base;
}
;
Slice<pend> hiPending={};
{auto&& tmp6 = m->relocs;
for(int64_t tmp7=0;tmp7<len(tmp6);++tmp7){
auto r=tmp6[tmp7];int64_t ofsSeg = cast<int64_t>(cast<uint32_t>(((shr<uint32_t>(r.info,cast<int64_t>(8ULL))) & cast<uint32_t>(255ULL))));
uint32_t addrBase = segAddr(cast<int64_t>(cast<uint32_t>(((shr<uint32_t>(r.info,cast<int64_t>(16ULL))) & cast<uint32_t>(255ULL)))));
if ((ofsSeg >= len(m->Segments))) {
continue;
}
uint32_t off = r.offset;
{
switch(cast<uint32_t>((r.info & cast<uint32_t>(255ULL)))){
case cast<uint32_t>(2ULL):{
write(ofsSeg,off,cast<uint32_t>((read(ofsSeg,off) + addrBase)));
break;}
case cast<uint32_t>(4ULL):{
uint32_t w = read(ofsSeg,off);
uint32_t target = cast<uint32_t>(((shl<uint32_t>((cast<uint32_t>((w & cast<uint32_t>(67108863ULL)))),cast<int64_t>(2ULL))) + addrBase));
write(ofsSeg,off,cast<uint32_t>(((cast<uint32_t>((w & cast<uint32_t>(4227858432ULL)))) | (cast<uint32_t>(((shr<uint32_t>(target,cast<int64_t>(2ULL))) & cast<uint32_t>(67108863ULL)))))));
break;}
case cast<uint32_t>(5ULL):{
hiPending = append(hiPending,pend{ofsSeg,off});
break;}
case cast<uint32_t>(6ULL):{
int32_t lo = cast<int32_t>(cast<int16_t>(read(ofsSeg,off)));
{auto&& tmp8 = hiPending;
for(int64_t tmp9=0;tmp9<len(tmp8);++tmp9){
auto h=tmp8[tmp9];uint32_t hi = read(h.si,h.off);
uint32_t val = cast<uint32_t>((cast<uint32_t>(((shl<uint32_t>((cast<uint32_t>((hi & cast<uint32_t>(65535ULL)))),cast<int64_t>(16ULL))) + cast<uint32_t>(lo))) + addrBase));
uint32_t hiField = shr<uint32_t>((cast<uint32_t>((val - cast<uint32_t>(cast<int32_t>(cast<int16_t>(val)))))),cast<int64_t>(16ULL));
write(h.si,h.off,cast<uint32_t>(((cast<uint32_t>((hi & cast<uint32_t>(4294901760ULL)))) | (cast<uint32_t>((hiField & cast<uint32_t>(65535ULL)))))));
}}
hiPending = sub(hiPending,0,cast<int64_t>(0ULL));
write(ofsSeg,off,cast<uint32_t>(((cast<uint32_t>((read(ofsSeg,off) & cast<uint32_t>(4294901760ULL)))) | (cast<uint32_t>((cast<uint32_t>(cast<int32_t>((cast<int32_t>(cast<int16_t>(read(ofsSeg,off))) + cast<int32_t>(addrBase)))) & cast<uint32_t>(65535ULL)))))));
break;}
}}
}}
{auto&& tmp10 = m->Segments;
for(int64_t tmp11=0;tmp11<len(tmp10);++tmp11){
auto i=tmp11;m->Segments[i].VAddr += base;
}}
m->EntryPC += base;
m->GP += base;
}
}
// tools/platform/psp/elf.go:223:1
Slice<uint8_t> psp_Module_segData(psp_Module* m,uint32_t va,uint32_t n){
{
{auto&& tmp12 = m->Segments;
for(int64_t tmp13=0;tmp13<len(tmp12);++tmp13){
auto s=tmp12[tmp13];if (((va >= s.VAddr) && (cast<uint32_t>((va + n)) <= cast<uint32_t>((s.VAddr + cast<uint32_t>(len(s.Data))))))) {
uint32_t off = cast<uint32_t>((va - s.VAddr));
return sub(s.Data,off,cast<uint32_t>((off + n)));
}
}}
return {};
}
}
// tools/platform/psp/elf.go:234:1
std::tuple<uint32_t,bool> psp_Module_fileToVA(psp_Module* m,uint32_t fileOff){
{
{auto&& tmp14 = m->Segments;
for(int64_t tmp15=0;tmp15<len(tmp14);++tmp15){
auto s=tmp14[tmp15];if (((fileOff >= s.FileOff) && (fileOff < cast<uint32_t>((s.FileOff + cast<uint32_t>(len(s.Data))))))) {
return {cast<uint32_t>((s.VAddr + (cast<uint32_t>((fileOff - s.FileOff))))),true};
}
}}
return {cast<uint32_t>(0ULL),false};
}
}
// tools/platform/psp/elf.go:247:1
void psp_Module_parseModuleInfo(psp_Module* m,Slice<uint8_t> b){
{
uint32_t phoff = le_Uint32(sub(b,cast<int64_t>(28ULL),len(b)));
int64_t phentsize = cast<int64_t>(le_Uint16(sub(b,cast<int64_t>(42ULL),len(b))));
int64_t phnum = cast<int64_t>(le_Uint16(sub(b,cast<int64_t>(44ULL),len(b))));
uint32_t infoVA={};
bool found = false;
{int64_t i = cast<int64_t>(0ULL);for (;(i < phnum);i++){
int64_t ph = cast<int64_t>((cast<int64_t>(phoff) + cast<int64_t>((i * phentsize))));
if (((cast<int64_t>((ph + cast<int64_t>(32ULL))) > len(b)) || (le_Uint32(sub(b,ph,len(b))) != cast<uint32_t>(1ULL)))) {
continue;
}
uint32_t paddr = cast<uint32_t>((le_Uint32(sub(b,cast<int64_t>((ph + cast<int64_t>(12ULL))),len(b))) & cast<uint32_t>(2147483647ULL)));
auto tmp16 = psp_Module_fileToVA(m,paddr);
infoVA = std::get<0>(tmp16);
found = std::get<1>(tmp16);
break;
}
}if ((!found)) {
return ;
}
Slice<uint8_t> info = psp_Module_segData(m,infoVA,cast<uint32_t>(52ULL));
if ((!info)) {
return ;
}
m->Name = go_strings_TrimRight(cast<std::string>(sub(info,cast<int64_t>(4ULL),cast<int64_t>(32ULL))),std::string("\000",1));
m->GP = le_Uint32(sub(info,cast<int64_t>(32ULL),len(info)));
uint32_t stubTop = le_Uint32(sub(info,cast<int64_t>(44ULL),len(info)));
uint32_t stubEnd = le_Uint32(sub(info,cast<int64_t>(48ULL),len(info)));
psp_Module_parseStubs(m,stubTop,stubEnd);
}
}
// tools/platform/psp/elf.go:288:1
void psp_Module_parseStubs(psp_Module* m,uint32_t stubTop,uint32_t stubEnd){
{
{uint32_t addr = stubTop;for (;(cast<uint32_t>((addr + cast<uint32_t>(20ULL))) <= stubEnd);){
Slice<uint8_t> hdr = psp_Module_segData(m,addr,cast<uint32_t>(24ULL));
if ((!hdr)) {
return ;
}
uint32_t namePtr = le_Uint32(sub(hdr,cast<int64_t>(0ULL),len(hdr)));
uint8_t lenWords = hdr[cast<int64_t>(8ULL)];
uint16_t nfunc={};
uint32_t nidPtr={};
uint32_t stubPtr={};
if ((lenWords >= cast<uint8_t>(6ULL))) {
nfunc = le_Uint16(sub(hdr,cast<int64_t>(14ULL),len(hdr)));
nidPtr = le_Uint32(sub(hdr,cast<int64_t>(16ULL),len(hdr)));
stubPtr = le_Uint32(sub(hdr,cast<int64_t>(20ULL),len(hdr)));
}
else {
nfunc = le_Uint16(sub(hdr,cast<int64_t>(10ULL),len(hdr)));
nidPtr = le_Uint32(sub(hdr,cast<int64_t>(12ULL),len(hdr)));
stubPtr = le_Uint32(sub(hdr,cast<int64_t>(16ULL),len(hdr)));
}
psp_Import imp = psp_Import{psp_cstrVA(m,namePtr),{},stubPtr};
{
Slice<uint8_t> nids = psp_Module_segData(m,nidPtr,cast<uint32_t>((cast<uint32_t>(nfunc) * cast<uint32_t>(4ULL))));
if (bool(nids)) {
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(nfunc));i++){
imp.NIDs = append(imp.NIDs,le_Uint32(sub(nids,cast<int64_t>((i * cast<int64_t>(4ULL))),len(nids))));
}
}}
}
m->Imports = append(m->Imports,imp);
uint32_t step = cast<uint32_t>((cast<uint32_t>(lenWords) * cast<uint32_t>(4ULL)));
if ((step < cast<uint32_t>(20ULL))) {
step = cast<uint32_t>(20ULL);
}
addr += step;
}
}}
}
// tools/platform/psp/elf.go:326:1
std::string psp_cstrVA(psp_Module* m,uint32_t va){
{
Slice<uint8_t> b = psp_Module_segData(m,va,cast<uint32_t>(1ULL));
if ((!b)) {
return std::string("",0);
}
{auto&& tmp17 = m->Segments;
for(int64_t tmp18=0;tmp18<len(tmp17);++tmp18){
auto s=tmp17[tmp18];if (((va >= s.VAddr) && (va < cast<uint32_t>((s.VAddr + cast<uint32_t>(len(s.Data))))))) {
uint32_t off = cast<uint32_t>((va - s.VAddr));
uint32_t end = off;
{;for (;((end < cast<uint32_t>(len(s.Data))) && (s.Data[end] != cast<uint8_t>(0ULL)));){
end++;
}
}return cast<std::string>(sub(s.Data,off,end));
}
}}
return std::string("",0);
}
}
// tools/platform/psp/elf.go:379:1
std::tuple<psp_Module*,Error> psp_LoadModuleImage(Slice<uint8_t> raw){
{
if (((len(raw) >= cast<int64_t>(4ULL)) && (cast<std::string>(sub(raw,cast<int64_t>(0ULL),cast<int64_t>(4ULL))) == std::string("\177ELF",4)))) {
return psp_ParseELF(raw);
}
if (((len(raw) >= cast<int64_t>(4ULL)) && (cast<std::string>(sub(raw,cast<int64_t>(0ULL),cast<int64_t>(4ULL))) == std::string("~PSP",4)))) {
auto tmp19 = psp_DecryptPRX(raw);
Slice<uint8_t> plain = std::get<0>(tmp19);
uint32_t tag = std::get<1>(tmp19);
Error err = std::get<2>(tmp19);
if (bool(err)) {
return {{},go_fmt_Errorf(std::string("decrypt ~PSP: %w",16),err)};
}
auto tmp20 = psp_ParseELF(plain);
psp_Module* m = std::get<0>(tmp20);
err = std::get<1>(tmp20);
if (bool(err)) {
return {{},err};
}
m->Encrypted = true;
m->Tag = tag;
return {m,{}};
}
return {{},go_fmt_Errorf(std::string("elf: not an ELF or ~PSP image (magic %q)",40),psp_firstMagic(raw))};
}
}
// tools/platform/psp/elf.go:399:1
std::string psp_firstMagic(Slice<uint8_t> b){
{
int64_t n = cast<int64_t>(4ULL);
if ((len(b) < n)) {
n = len(b);
}
return go_fmt_Sprintf(std::string("% X",3),sub(b,0,n));
}
}
// tools/platform/psp/framebuffer.go:31:1
image_RGBA* psp_Machine_Framebuffer(psp_Machine* m){
{
uint32_t addr = m->fbAddr;
if ((addr == cast<uint32_t>(0ULL))) {
addr = cast<uint32_t>(67108864ULL);
}
uint32_t stride = m->fbWidth;
if ((stride == cast<uint32_t>(0ULL))) {
stride = cast<uint32_t>(480ULL);
}
image_RGBA* img = go_image_NewRGBA(go_image_Rect(cast<int64_t>(0ULL),cast<int64_t>(0ULL),cast<int64_t>(480ULL),cast<int64_t>(272ULL)));
{int64_t y = cast<int64_t>(0ULL);for (;(y < cast<int64_t>(272ULL));y++){
{int64_t x = cast<int64_t>(0ULL);for (;(x < cast<int64_t>(480ULL));x++){
image_RGBA_Set(img,x,y,psp_Machine_readPixel(m,addr,stride,cast<uint32_t>(x),cast<uint32_t>(y)));
}
}}
}return img;
}
}
// tools/platform/psp/framebuffer.go:50:1
color_RGBA psp_Machine_readPixel(psp_Machine* m,uint32_t base,uint32_t stride,uint32_t x,uint32_t y){
{
return psp_Machine_readPixelFmt(m,base,stride,m->fbFormat,x,y);
}
}
// tools/platform/psp/framebuffer.go:56:1
color_RGBA psp_Machine_readPixelFmt(psp_Machine* m,uint32_t base,uint32_t stride,uint32_t format,uint32_t x,uint32_t y){
{
{
switch(format){
case cast<uint32_t>(3ULL):{
uint32_t a = cast<uint32_t>((base + cast<uint32_t>(((cast<uint32_t>((cast<uint32_t>((y * stride)) + x))) * cast<uint32_t>(4ULL)))));
uint32_t p = psp_Machine_read32(m,a);
return color_RGBA{cast<uint8_t>(p),cast<uint8_t>(shr<uint32_t>(p,cast<int64_t>(8ULL))),cast<uint8_t>(shr<uint32_t>(p,cast<int64_t>(16ULL))),cast<uint8_t>(255ULL)};
break;}
default:{
uint32_t a = cast<uint32_t>((base + cast<uint32_t>(((cast<uint32_t>((cast<uint32_t>((y * stride)) + x))) * cast<uint32_t>(2ULL)))));
uint16_t p = cast<uint16_t>((cast<uint16_t>(psp_Machine_Read(m,a)) | shl<uint16_t>(cast<uint16_t>(psp_Machine_Read(m,cast<uint32_t>((a + cast<uint32_t>(1ULL))))),cast<int64_t>(8ULL))));
return psp_decode16(p,format);
break;}
}}
}
}
// tools/platform/psp/framebuffer.go:70:1
color_RGBA psp_decode16(uint16_t p,uint32_t fmt){
{
auto ext = [&](uint16_t v,uint16_t bits)->uint8_t{
v &= cast<uint16_t>(((shl<uint16_t>(cast<uint16_t>(1ULL),bits)) - cast<uint16_t>(1ULL)));
return cast<uint8_t>(divi<uint32_t>((cast<uint32_t>((cast<uint32_t>(v) * cast<uint32_t>(255ULL)))),(cast<uint32_t>(((shl<uint32_t>(cast<uint32_t>(1ULL),bits)) - cast<uint32_t>(1ULL))))));
}
;
{
switch(fmt){
case cast<uint32_t>(1ULL):{
return color_RGBA{ext(p,cast<uint16_t>(5ULL)),ext(shr<uint16_t>(p,cast<int64_t>(5ULL)),cast<uint16_t>(5ULL)),ext(shr<uint16_t>(p,cast<int64_t>(10ULL)),cast<uint16_t>(5ULL)),cast<uint8_t>(255ULL)};
break;}
case cast<uint32_t>(2ULL):{
return color_RGBA{ext(p,cast<uint16_t>(4ULL)),ext(shr<uint16_t>(p,cast<int64_t>(4ULL)),cast<uint16_t>(4ULL)),ext(shr<uint16_t>(p,cast<int64_t>(8ULL)),cast<uint16_t>(4ULL)),cast<uint8_t>(255ULL)};
break;}
default:{
return color_RGBA{ext(p,cast<uint16_t>(5ULL)),ext(shr<uint16_t>(p,cast<int64_t>(5ULL)),cast<uint16_t>(6ULL)),ext(shr<uint16_t>(p,cast<int64_t>(11ULL)),cast<uint16_t>(5ULL)),cast<uint8_t>(255ULL)};
break;}
}}
}
}
// tools/platform/psp/framebuffer.go:120:1
std::string psp_Machine_FramebufferInfo(psp_Machine* m){
{
return go_fmt_Sprintf(std::string("fb=0x%08X stride=%d psm=%d",26),m->fbAddr,m->fbWidth,m->fbFormat);
}
}
// tools/platform/psp/ge.go:50:1
psp_GeList psp_Machine_captureList(psp_Machine* m,uint32_t addr){
{
constexpr int64_t cap=1048576ULL;
psp_GeList out = psp_GeList{addr,{}};
uint32_t pc = addr;
uint32_t base = cast<uint32_t>((addr & cast<uint32_t>(4278190080ULL)));
Slice<uint32_t> callStack={};
int64_t seen = cast<int64_t>(0ULL);
{;for (;(seen < cast<int64_t>(1048576ULL));){
uint32_t w = psp_Machine_read32(m,pc);
out.Words = append(out.Words,w);
uint32_t cmd = shr<uint32_t>(w,cast<int64_t>(24ULL));
uint32_t arg = cast<uint32_t>((w & cast<uint32_t>(16777215ULL)));
pc += cast<uint32_t>(4ULL);
seen++;
{
switch(cmd){
case cast<uint32_t>(16ULL):{
base = psp_geBaseAddr(arg);
break;}
case cast<uint32_t>(8ULL):{
pc = ((cast<uint32_t>((base | (cast<uint32_t>((arg & cast<uint32_t>(16777215ULL))))))) & ~(cast<uint32_t>(3ULL)));
break;}
case cast<uint32_t>(10ULL):{
callStack = append(callStack,pc);
pc = ((cast<uint32_t>((base | (cast<uint32_t>((arg & cast<uint32_t>(16777215ULL))))))) & ~(cast<uint32_t>(3ULL)));
break;}
case cast<uint32_t>(11ULL):{
if ((len(callStack) > cast<int64_t>(0ULL))) {
pc = callStack[cast<int64_t>((len(callStack) - cast<int64_t>(1ULL)))];
callStack = sub(callStack,0,cast<int64_t>((len(callStack) - cast<int64_t>(1ULL))));
}
break;}
case cast<uint32_t>(15ULL):{
break;}
case cast<uint32_t>(12ULL):{
return out;
break;}
}}
if (((pc < cast<uint32_t>(134217728ULL)) && (pc >= cast<uint32_t>(69206016ULL)))) {
return out;
}
}
}return out;
}
}
// tools/platform/psp/ge.go:96:1
uint32_t psp_geBaseAddr(uint32_t arg){
{
return shl<uint32_t>((cast<uint32_t>((arg & cast<uint32_t>(983040ULL)))),cast<int64_t>(8ULL));
}
}
// tools/platform/psp/ge.go:99:1
void psp_Machine_execGeList(psp_Machine* m,psp_GeList list){
{
psp_Machine_rasterList(m,list);
}
}
// tools/platform/psp/ge.go:164:1
std::string psp_GeCmdName(uint32_t cmd){
{
if ((cmd < cast<uint32_t>(256ULL))) {
{
std::string s = psp_geCmdNames[cmd];
if ((s != std::string("",0))) {
return s;
}
}
}
return go_fmt_Sprintf(std::string("0x%02X",6),cmd);
}
}
// tools/platform/psp/ge_draw.go:14:1
void psp_Machine_rasterTriClipped(psp_Machine* m,psp_geState* s,psp_vert a,psp_vert b,psp_vert c){
{
{auto&& tmp21 = psp_clipTriNear(s,a,b,c);
for(int64_t tmp22=0;tmp22<len(tmp21);++tmp22){
auto t=tmp21[tmp22];psp_Machine_rasterTri(m,s,t[cast<int64_t>(0ULL)],t[cast<int64_t>(1ULL)],t[cast<int64_t>(2ULL)]);
}}
}
}
// tools/platform/psp/ge_draw.go:22:1
void psp_Machine_rasterTri(psp_Machine* m,psp_geState* s,psp_vert a,psp_vert b,psp_vert c){
{
int64_t minX = cast<int64_t>(psp_fmin3(a.x,b.x,c.x));
int64_t maxX = cast<int64_t>((cast<int64_t>(psp_fmax3(a.x,b.x,c.x)) + cast<int64_t>(1ULL)));
int64_t minY = cast<int64_t>(psp_fmin3(a.y,b.y,c.y));
int64_t maxY = cast<int64_t>((cast<int64_t>(psp_fmax3(a.y,b.y,c.y)) + cast<int64_t>(1ULL)));
auto tmp23 = std::make_tuple(psp_clampI(minX,cast<int64_t>(0ULL),cast<int64_t>(480ULL)),psp_clampI(minY,cast<int64_t>(0ULL),cast<int64_t>(272ULL)));
minX = std::get<0>(tmp23);
minY = std::get<1>(tmp23);
auto tmp24 = std::make_tuple(psp_clampI(maxX,cast<int64_t>(0ULL),cast<int64_t>(480ULL)),psp_clampI(maxY,cast<int64_t>(0ULL),cast<int64_t>(272ULL)));
maxX = std::get<0>(tmp24);
maxY = std::get<1>(tmp24);
float area = psp_edge(a,b,c);
if ((area == cast<float>(0ULL))) {
return ;
}
if ((((s->cullOn && (!psp_geNoCull)) && (!s->clearOn)) && (!psp_vert_through(a)))) {
if ((((((s->cullFace == cast<uint32_t>(0ULL))) == ((area < cast<float>(0ULL))))) != psp_geCullFlip)) {
return ;
}
}
{int64_t y = minY;for (;(y < maxY);y++){
{int64_t x = minX;for (;(x < maxX);x++){
psp_vert px = psp_vert{cast<float>((cast<float>(x) + 0.5)),cast<float>((cast<float>(y) + 0.5)),{},{},{},{},{},{},{},{},{},{},{},{},{},{}};
float w0 = psp_edge(b,c,px);
float w1 = psp_edge(c,a,px);
float w2 = psp_edge(a,b,px);
if ((((((w0 < cast<float>(0ULL)) || (w1 < cast<float>(0ULL))) || (w2 < cast<float>(0ULL)))) && ((((w0 > cast<float>(0ULL)) || (w1 > cast<float>(0ULL))) || (w2 > cast<float>(0ULL)))))) {
continue;
}
auto tmp25 = std::make_tuple((w0 / area),(w1 / area),(w2 / area));
float l0 = std::get<0>(tmp25);
float l1 = std::get<1>(tmp25);
float l2 = std::get<2>(tmp25);
float z = (((l0 * a.z) + (l1 * b.z)) + (l2 * c.z));
uint8_t r = cast<uint8_t>((((l0 * cast<float>(a.r)) + (l1 * cast<float>(b.r))) + (l2 * cast<float>(c.r))));
uint8_t g = cast<uint8_t>((((l0 * cast<float>(a.g)) + (l1 * cast<float>(b.g))) + (l2 * cast<float>(c.g))));
uint8_t bl = cast<uint8_t>((((l0 * cast<float>(a.b)) + (l1 * cast<float>(b.b))) + (l2 * cast<float>(c.b))));
uint8_t al = cast<uint8_t>((((l0 * cast<float>(a.a)) + (l1 * cast<float>(b.a))) + (l2 * cast<float>(c.a))));
if ((s->texEnable && (!s->clearOn))) {
auto tmp26 = psp_uvAt(s,a,b,c,area,(cast<float>(x) + 0.5),(cast<float>(y) + 0.5));
float u = std::get<0>(tmp26);
float v = std::get<1>(tmp26);
float rho=cast<float>(1ULL);
if ((s->texMaxLvl > cast<uint32_t>(0ULL))) {
auto tmp27 = psp_uvAt(s,a,b,c,area,(cast<float>(x) + 1.5),(cast<float>(y) + 0.5));
float ux = std::get<0>(tmp27);
float vx = std::get<1>(tmp27);
auto tmp28 = psp_uvAt(s,a,b,c,area,(cast<float>(x) + 0.5),(cast<float>(y) + 1.5));
float uy = std::get<0>(tmp28);
float vy = std::get<1>(tmp28);
auto tmp29 = std::make_tuple(cast<float>(s->texW),cast<float>(s->texH));
float fw = std::get<0>(tmp29);
float fh = std::get<1>(tmp29);
float dx = psp_hypot32((((ux - u)) * fw),(((vx - v)) * fh));
float dy = psp_hypot32((((uy - u)) * fw),(((vy - v)) * fh));
rho = dx;
if ((dy > rho)) {
rho = dy;
}
}
auto tmp30 = psp_modTex(m,s,u,v,rho,r,g,bl,al);
r = std::get<0>(tmp30);
g = std::get<1>(tmp30);
bl = std::get<2>(tmp30);
al = std::get<3>(tmp30);
}
if (((s->fogOn && a.clip) && (!s->clearOn))) {
auto tmp31 = psp_applyFog(s,(((l0 * a.fog) + (l1 * b.fog)) + (l2 * c.fog)),r,g,bl);
r = std::get<0>(tmp31);
g = std::get<1>(tmp31);
bl = std::get<2>(tmp31);
}
psp_Machine_putPixel(m,s,x,y,z,r,g,bl,al);
}
}}
}}
}
// tools/platform/psp/ge_draw.go:94:1
void psp_Machine_rasterSprite(psp_Machine* m,psp_geState* s,psp_vert a,psp_vert b){
{
auto tmp32 = std::make_tuple(cast<int64_t>(a.x),cast<int64_t>(b.x));
int64_t x0 = std::get<0>(tmp32);
int64_t x1 = std::get<1>(tmp32);
auto tmp33 = std::make_tuple(cast<int64_t>(a.y),cast<int64_t>(b.y));
int64_t y0 = std::get<0>(tmp33);
int64_t y1 = std::get<1>(tmp33);
if ((x0 > x1)) {
auto tmp34 = std::make_tuple(x1,x0);
x0 = std::get<0>(tmp34);
x1 = std::get<1>(tmp34);
}
if ((y0 > y1)) {
auto tmp35 = std::make_tuple(y1,y0);
y0 = std::get<0>(tmp35);
y1 = std::get<1>(tmp35);
}
auto tmp36 = std::make_tuple(a.u,b.u);
float u0 = std::get<0>(tmp36);
float u1 = std::get<1>(tmp36);
auto tmp37 = std::make_tuple(a.v,b.v);
float v0 = std::get<0>(tmp37);
float v1 = std::get<1>(tmp37);
auto tmp38 = std::make_tuple(psp_clampI(x0,cast<int64_t>(0ULL),cast<int64_t>(480ULL)),psp_clampI(y0,cast<int64_t>(0ULL),cast<int64_t>(272ULL)));
int64_t x0c = std::get<0>(tmp38);
int64_t y0c = std::get<1>(tmp38);
auto tmp39 = std::make_tuple(psp_clampI(x1,cast<int64_t>(0ULL),cast<int64_t>(480ULL)),psp_clampI(y1,cast<int64_t>(0ULL),cast<int64_t>(272ULL)));
int64_t x1c = std::get<0>(tmp39);
int64_t y1c = std::get<1>(tmp39);
float dw = cast<float>(cast<int64_t>((x1 - x0)));
float dh = cast<float>(cast<int64_t>((y1 - y0)));
{int64_t y = y0c;for (;(y < y1c);y++){
{int64_t x = x0c;for (;(x < x1c);x++){
auto tmp40 = std::make_tuple(b.r,b.g,b.b,b.a);
uint8_t r = std::get<0>(tmp40);
uint8_t g = std::get<1>(tmp40);
uint8_t bl = std::get<2>(tmp40);
uint8_t al = std::get<3>(tmp40);
if ((((s->texEnable && (!s->clearOn)) && (dw != cast<float>(0ULL))) && (dh != cast<float>(0ULL)))) {
float u = (u0 + (((u1 - u0)) * ((cast<float>(cast<int64_t>((x - x0))) / dw))));
float v = (v0 + (((v1 - v0)) * ((cast<float>(cast<int64_t>((y - y0))) / dh))));
auto tmp41 = psp_modTex(m,s,u,v,cast<float>(1ULL),r,g,bl,al);
r = std::get<0>(tmp41);
g = std::get<1>(tmp41);
bl = std::get<2>(tmp41);
al = std::get<3>(tmp41);
}
if (((s->fogOn && b.clip) && (!s->clearOn))) {
auto tmp42 = psp_applyFog(s,b.fog,r,g,bl);
r = std::get<0>(tmp42);
g = std::get<1>(tmp42);
bl = std::get<2>(tmp42);
}
psp_Machine_putPixel(m,s,x,y,b.z,r,g,bl,al);
}
}}
}}
}
// tools/platform/psp/ge_draw.go:128:1
std::tuple<uint8_t,uint8_t,uint8_t> psp_applyFog(psp_geState* s,float f,uint8_t r,uint8_t g,uint8_t b){
{
f = psp_clampF(f,cast<float>(0ULL),cast<float>(1ULL));
auto tmp43 = std::make_tuple(cast<float>(cast<uint8_t>(s->fogColor)),cast<float>(cast<uint8_t>(shr<uint32_t>(s->fogColor,cast<int64_t>(8ULL)))),cast<float>(cast<uint8_t>(shr<uint32_t>(s->fogColor,cast<int64_t>(16ULL)))));
float fr = std::get<0>(tmp43);
float fg = std::get<1>(tmp43);
float fb = std::get<2>(tmp43);
auto mix = [&](float c,float fc)->uint8_t{
return cast<uint8_t>(((c * f) + (fc * ((cast<float>(1ULL) - f)))));
}
;
return {mix(cast<float>(r),fr),mix(cast<float>(g),fg),mix(cast<float>(b),fb)};
}
}
// tools/platform/psp/ge_draw.go:137:1
std::tuple<float,float> psp_uvAt(psp_geState* s,psp_vert a,psp_vert b,psp_vert c,float area,float px,float py){
{
psp_vert p = psp_vert{cast<float>(px),cast<float>(py),{},{},{},{},{},{},{},{},{},{},{},{},{},{}};
float l0 = (psp_edge(b,c,p) / area);
float l1 = (psp_edge(c,a,p) / area);
float l2 = (psp_edge(a,b,p) / area);
if ((!a.clip)) {
return {(((l0 * a.u) + (l1 * b.u)) + (l2 * c.u)),(((l0 * a.v) + (l1 * b.v)) + (l2 * c.v))};
}
float iw = (((l0 * a.invW) + (l1 * b.invW)) + (l2 * c.invW));
if ((iw == cast<float>(0ULL))) {
return {cast<float>(0ULL),cast<float>(0ULL)};
}
return {((((((l0 * a.u) * a.invW) + ((l1 * b.u) * b.invW)) + ((l2 * c.u) * c.invW))) / iw),((((((l0 * a.v) * a.invW) + ((l1 * b.v) * b.invW)) + ((l2 * c.v) * c.invW))) / iw)};
}
}
// tools/platform/psp/ge_draw.go:153:1
float psp_hypot32(float x,float y){
{
return cast<float>(go_math_Hypot(cast<double>(x),cast<double>(y)));
}
}
// tools/platform/psp/ge_draw.go:158:1
uint32_t psp_wrapTexel(int64_t i,uint32_t size,uint32_t clamp){
{
if ((clamp != cast<uint32_t>(0ULL))) {
if ((i < cast<int64_t>(0ULL))) {
i = cast<int64_t>(0ULL);
}
if ((i > cast<int64_t>((cast<int64_t>(size) - cast<int64_t>(1ULL))))) {
i = cast<int64_t>((cast<int64_t>(size) - cast<int64_t>(1ULL)));
}
return cast<uint32_t>(i);
}
i &= cast<int64_t>(cast<uint32_t>((size - cast<uint32_t>(1ULL))));
if ((i < cast<int64_t>(0ULL))) {
i += cast<int64_t>(size);
}
return cast<uint32_t>(i);
}
}
// tools/platform/psp/ge_draw.go:178:1
std::tuple<uint8_t,uint8_t,uint8_t,uint8_t> psp_modTex(psp_Machine* m,psp_geState* s,float u,float v,float rho,uint8_t r,uint8_t g,uint8_t b,uint8_t a){
{
if ((((s->texW == cast<uint32_t>(0ULL)) || (s->texH == cast<uint32_t>(0ULL))) || (s->texAddr == cast<uint32_t>(0ULL)))) {
return {r,g,b,a};
}
uint32_t lvl = psp_geState_texLevel(s,rho);
auto tmp44 = std::make_tuple(shr<uint32_t>(s->texW,lvl),shr<uint32_t>(s->texH,lvl));
uint32_t texW = std::get<0>(tmp44);
uint32_t texH = std::get<1>(tmp44);
if (((lvl > cast<uint32_t>(0ULL)) && (s->texWN[lvl] != cast<uint32_t>(0ULL)))) {
auto tmp45 = std::make_tuple(s->texWN[lvl],s->texHN[lvl]);
texW = std::get<0>(tmp45);
texH = std::get<1>(tmp45);
}
if (((texW == cast<uint32_t>(0ULL)) || (texH == cast<uint32_t>(0ULL)))) {
auto tmp46 = std::make_tuple(cast<uint32_t>(0ULL),s->texW,s->texH);
lvl = std::get<0>(tmp46);
texW = std::get<1>(tmp46);
texH = std::get<2>(tmp46);
}
auto wrap = [&](float t,uint32_t size,uint32_t clamp)->uint32_t{
int64_t i = cast<int64_t>((t * cast<float>(size)));
if ((clamp != cast<uint32_t>(0ULL))) {
if ((i < cast<int64_t>(0ULL))) {
i = cast<int64_t>(0ULL);
}
if ((i > cast<int64_t>((cast<int64_t>(size) - cast<int64_t>(1ULL))))) {
i = cast<int64_t>((cast<int64_t>(size) - cast<int64_t>(1ULL)));
}
return cast<uint32_t>(i);
}
i &= cast<int64_t>(cast<uint32_t>((size - cast<uint32_t>(1ULL))));
if ((i < cast<int64_t>(0ULL))) {
i += cast<int64_t>(size);
}
return cast<uint32_t>(i);
}
;
uint8_t tr={};
uint8_t tg={};
uint8_t tb={};
uint8_t ta={};
if (s->texLinear) {
float fu = ((u * cast<float>(texW)) - 0.5);
float fv = ((v * cast<float>(texH)) - 0.5);
float fx = (fu - cast<float>(go_math_Floor(cast<double>(fu))));
float fy = (fv - cast<float>(go_math_Floor(cast<double>(fv))));
uint32_t u0 = psp_wrapTexel(cast<int64_t>(go_math_Floor(cast<double>(fu))),texW,s->texWrapU);
uint32_t v0 = psp_wrapTexel(cast<int64_t>(go_math_Floor(cast<double>(fv))),texH,s->texWrapV);
uint32_t u1 = psp_wrapTexel(cast<int64_t>((cast<int64_t>(go_math_Floor(cast<double>(fu))) + cast<int64_t>(1ULL))),texW,s->texWrapU);
uint32_t v1 = psp_wrapTexel(cast<int64_t>((cast<int64_t>(go_math_Floor(cast<double>(fv))) + cast<int64_t>(1ULL))),texH,s->texWrapV);
auto tmp47 = psp_Machine_sampleTexLvl(m,s,u0,v0,lvl);
uint8_t r00 = std::get<0>(tmp47);
uint8_t g00 = std::get<1>(tmp47);
uint8_t b00 = std::get<2>(tmp47);
uint8_t a00 = std::get<3>(tmp47);
auto tmp48 = psp_Machine_sampleTexLvl(m,s,u1,v0,lvl);
uint8_t r10 = std::get<0>(tmp48);
uint8_t g10 = std::get<1>(tmp48);
uint8_t b10 = std::get<2>(tmp48);
uint8_t a10 = std::get<3>(tmp48);
auto tmp49 = psp_Machine_sampleTexLvl(m,s,u0,v1,lvl);
uint8_t r01 = std::get<0>(tmp49);
uint8_t g01 = std::get<1>(tmp49);
uint8_t b01 = std::get<2>(tmp49);
uint8_t a01 = std::get<3>(tmp49);
auto tmp50 = psp_Machine_sampleTexLvl(m,s,u1,v1,lvl);
uint8_t r11 = std::get<0>(tmp50);
uint8_t g11 = std::get<1>(tmp50);
uint8_t b11 = std::get<2>(tmp50);
uint8_t a11 = std::get<3>(tmp50);
auto lerp = [&](uint8_t c00,uint8_t c10,uint8_t c01,uint8_t c11)->uint8_t{
float top = ((cast<float>(c00) * ((cast<float>(1ULL) - fx))) + (cast<float>(c10) * fx));
float bot = ((cast<float>(c01) * ((cast<float>(1ULL) - fx))) + (cast<float>(c11) * fx));
return cast<uint8_t>(((top * ((cast<float>(1ULL) - fy))) + (bot * fy)));
}
;
auto tmp51 = std::make_tuple(lerp(r00,r10,r01,r11),lerp(g00,g10,g01,g11),lerp(b00,b10,b01,b11),lerp(a00,a10,a01,a11));
tr = std::get<0>(tmp51);
tg = std::get<1>(tmp51);
tb = std::get<2>(tmp51);
ta = std::get<3>(tmp51);
}
else {
auto tmp52 = psp_Machine_sampleTex(m,s,wrap(u,texW,s->texWrapU),wrap(v,texH,s->texWrapV));
tr = std::get<0>(tmp52);
tg = std::get<1>(tmp52);
tb = std::get<2>(tmp52);
ta = std::get<3>(tmp52);
}
uint8_t or_={};
uint8_t og={};
uint8_t ob={};
uint8_t oa={};
{
switch(s->texFunc){
case cast<uint32_t>(1ULL):{
if (s->texUseA) {
or_ = cast<uint8_t>(divi<uint32_t>((cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(r) * (cast<uint32_t>((cast<uint32_t>(255ULL) - cast<uint32_t>(ta)))))) + cast<uint32_t>((cast<uint32_t>(tr) * cast<uint32_t>(ta)))))),cast<uint32_t>(255ULL)));
og = cast<uint8_t>(divi<uint32_t>((cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(g) * (cast<uint32_t>((cast<uint32_t>(255ULL) - cast<uint32_t>(ta)))))) + cast<uint32_t>((cast<uint32_t>(tg) * cast<uint32_t>(ta)))))),cast<uint32_t>(255ULL)));
ob = cast<uint8_t>(divi<uint32_t>((cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(b) * (cast<uint32_t>((cast<uint32_t>(255ULL) - cast<uint32_t>(ta)))))) + cast<uint32_t>((cast<uint32_t>(tb) * cast<uint32_t>(ta)))))),cast<uint32_t>(255ULL)));
oa = a;
}
else {
auto tmp53 = std::make_tuple(tr,tg,tb,a);
or_ = std::get<0>(tmp53);
og = std::get<1>(tmp53);
ob = std::get<2>(tmp53);
oa = std::get<3>(tmp53);
}
break;}
case cast<uint32_t>(2ULL):{
auto tmp54 = std::make_tuple(cast<uint8_t>(s->texEnvCol),cast<uint8_t>(shr<uint32_t>(s->texEnvCol,cast<int64_t>(8ULL))),cast<uint8_t>(shr<uint32_t>(s->texEnvCol,cast<int64_t>(16ULL))));
uint8_t er = std::get<0>(tmp54);
uint8_t eg = std::get<1>(tmp54);
uint8_t eb = std::get<2>(tmp54);
auto mix = [&](uint8_t c,uint8_t e,uint8_t t)->uint8_t{
return cast<uint8_t>(divi<uint32_t>((cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(c) * (cast<uint32_t>((cast<uint32_t>(255ULL) - cast<uint32_t>(t)))))) + cast<uint32_t>((cast<uint32_t>(e) * cast<uint32_t>(t)))))),cast<uint32_t>(255ULL)));
}
;
auto tmp55 = std::make_tuple(mix(r,er,tr),mix(g,eg,tg),mix(b,eb,tb));
or_ = std::get<0>(tmp55);
og = std::get<1>(tmp55);
ob = std::get<2>(tmp55);
oa = a;
if (s->texUseA) {
oa = psp_mul8(ta,a);
}
break;}
case cast<uint32_t>(3ULL):{
auto tmp56 = std::make_tuple(tr,tg,tb);
or_ = std::get<0>(tmp56);
og = std::get<1>(tmp56);
ob = std::get<2>(tmp56);
oa = a;
if (s->texUseA) {
oa = ta;
}
break;}
case cast<uint32_t>(4ULL):{
auto tmp57 = std::make_tuple(psp_clamp255(cast<int32_t>((cast<int32_t>(r) + cast<int32_t>(tr)))),psp_clamp255(cast<int32_t>((cast<int32_t>(g) + cast<int32_t>(tg)))),psp_clamp255(cast<int32_t>((cast<int32_t>(b) + cast<int32_t>(tb)))));
or_ = std::get<0>(tmp57);
og = std::get<1>(tmp57);
ob = std::get<2>(tmp57);
oa = a;
if (s->texUseA) {
oa = psp_mul8(ta,a);
}
break;}
default:{
auto tmp58 = std::make_tuple(psp_mul8(tr,r),psp_mul8(tg,g),psp_mul8(tb,b));
or_ = std::get<0>(tmp58);
og = std::get<1>(tmp58);
ob = std::get<2>(tmp58);
oa = a;
if (s->texUseA) {
oa = psp_mul8(ta,a);
}
break;}
}}
if (s->texDouble) {
or_ = psp_clamp255(cast<int32_t>((cast<int32_t>(or_) * cast<int32_t>(2ULL))));
og = psp_clamp255(cast<int32_t>((cast<int32_t>(og) * cast<int32_t>(2ULL))));
ob = psp_clamp255(cast<int32_t>((cast<int32_t>(ob) * cast<int32_t>(2ULL))));
}
return {or_,og,ob,oa};
}
}
// tools/platform/psp/ge_draw.go:292:1
std::tuple<uint8_t,uint8_t,uint8_t,uint8_t> psp_Machine_sampleTex(psp_Machine* m,psp_geState* s,uint32_t tx,uint32_t ty){
{
return psp_Machine_sampleTexLvl(m,s,tx,ty,cast<uint32_t>(0ULL));
}
}
// tools/platform/psp/ge_draw.go:299:1
std::tuple<uint8_t,uint8_t,uint8_t,uint8_t> psp_Machine_sampleTexLvl_Reference(psp_Machine* m,psp_geState* s,uint32_t tx,uint32_t ty,uint32_t lvl){
{
auto tmp59 = std::make_tuple(s->texAddr,s->texStride);
uint32_t addr = std::get<0>(tmp59);
uint32_t stride = std::get<1>(tmp59);
if (((lvl > cast<uint32_t>(0ULL)) && (s->texAddrN[lvl] != cast<uint32_t>(0ULL)))) {
auto tmp60 = std::make_tuple(s->texAddrN[lvl],s->texStrideN[lvl]);
addr = std::get<0>(tmp60);
stride = std::get<1>(tmp60);
}
if ((stride == cast<uint32_t>(0ULL))) {
stride = shr<uint32_t>(s->texW,lvl);
}
if (((addr == cast<uint32_t>(0ULL)) || (stride == cast<uint32_t>(0ULL)))) {
return {cast<uint8_t>(255ULL),cast<uint8_t>(255ULL),cast<uint8_t>(255ULL),cast<uint8_t>(255ULL)};
}
{
switch(s->texFmt){
case cast<uint32_t>(3ULL):{
uint32_t off = psp_Machine_texOff(m,s,cast<uint32_t>((tx * cast<uint32_t>(4ULL))),ty,cast<uint32_t>((stride * cast<uint32_t>(4ULL))));
uint32_t c = psp_Machine_read32(m,cast<uint32_t>((addr + off)));
return {cast<uint8_t>(c),cast<uint8_t>(shr<uint32_t>(c,cast<int64_t>(8ULL))),cast<uint8_t>(shr<uint32_t>(c,cast<int64_t>(16ULL))),cast<uint8_t>(shr<uint32_t>(c,cast<int64_t>(24ULL)))};
break;}
case cast<uint32_t>(0ULL):case cast<uint32_t>(1ULL):case cast<uint32_t>(2ULL):{
uint32_t off = psp_Machine_texOff(m,s,cast<uint32_t>((tx * cast<uint32_t>(2ULL))),ty,cast<uint32_t>((stride * cast<uint32_t>(2ULL))));
return psp_decode16a(psp_u16(m,cast<uint32_t>((addr + off))),s->texFmt);
break;}
case cast<uint32_t>(4ULL):{
uint32_t off = psp_Machine_texOff(m,s,divi<uint32_t>(tx,cast<uint32_t>(2ULL)),ty,divi<uint32_t>(stride,cast<uint32_t>(2ULL)));
uint32_t raw = cast<uint32_t>((shr<uint32_t>(cast<uint32_t>(psp_Machine_Read(m,cast<uint32_t>((addr + off)))),(cast<uint32_t>((cast<uint32_t>(4ULL) * (cast<uint32_t>((tx & cast<uint32_t>(1ULL)))))))) & cast<uint32_t>(15ULL)));
return psp_geState_clutLookup(s,raw);
break;}
case cast<uint32_t>(5ULL):{
uint32_t off = psp_Machine_texOff(m,s,tx,ty,stride);
return psp_geState_clutLookup(s,cast<uint32_t>(psp_Machine_Read(m,cast<uint32_t>((addr + off)))));
break;}
}}
return {cast<uint8_t>(255ULL),cast<uint8_t>(255ULL),cast<uint8_t>(255ULL),cast<uint8_t>(255ULL)};
}
}
// tools/platform/psp/ge_draw.go:332:1
uint32_t psp_geState_texLevel(psp_geState* s,float rho){
{
if (((s->texMaxLvl == cast<uint32_t>(0ULL)) || (rho <= cast<float>(1ULL)))) {
return cast<uint32_t>(0ULL);
}
float lod = (cast<float>(go_math_Log2(cast<double>(rho))) + s->texLodBias);
if ((lod <= cast<float>(0ULL))) {
return cast<uint32_t>(0ULL);
}
uint32_t l = cast<uint32_t>((lod + 0.5));
if ((l > s->texMaxLvl)) {
l = s->texMaxLvl;
}
if ((l > cast<uint32_t>(7ULL))) {
l = cast<uint32_t>(7ULL);
}
return l;
}
}
// tools/platform/psp/ge_draw.go:352:1
uint32_t psp_Machine_texOff(psp_Machine* m,psp_geState* s,uint32_t xb,uint32_t y,uint32_t rowBytes){
{
if (((!s->texSwizzle) || (rowBytes < cast<uint32_t>(16ULL)))) {
return cast<uint32_t>((cast<uint32_t>((y * rowBytes)) + xb));
}
uint32_t rowBlocks = divi<uint32_t>(rowBytes,cast<uint32_t>(16ULL));
uint32_t block = cast<uint32_t>((cast<uint32_t>(((divi<uint32_t>(y,cast<uint32_t>(8ULL))) * rowBlocks)) + divi<uint32_t>(xb,cast<uint32_t>(16ULL))));
return cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((block * cast<uint32_t>(128ULL))) + cast<uint32_t>(((modi<uint32_t>(y,cast<uint32_t>(8ULL))) * cast<uint32_t>(16ULL))))) + modi<uint32_t>(xb,cast<uint32_t>(16ULL))));
}
}
// tools/platform/psp/ge_draw.go:363:1
std::tuple<uint8_t,uint8_t,uint8_t,uint8_t> psp_geState_clutLookup(psp_geState* s,uint32_t raw){
{
uint32_t shift = cast<uint32_t>(((shr<uint32_t>(s->clutFmt,cast<int64_t>(2ULL))) & cast<uint32_t>(31ULL)));
uint32_t mask = cast<uint32_t>(((shr<uint32_t>(s->clutFmt,cast<int64_t>(8ULL))) & cast<uint32_t>(255ULL)));
uint32_t base = cast<uint32_t>(((shr<uint32_t>(s->clutFmt,cast<int64_t>(16ULL))) & cast<uint32_t>(31ULL)));
uint32_t c = s->clut[cast<uint32_t>(((cast<uint32_t>((cast<uint32_t>(((shr<uint32_t>(raw,shift)) & mask)) + shl<uint32_t>(base,cast<int64_t>(4ULL))))) & cast<uint32_t>(255ULL)))];
return {cast<uint8_t>(c),cast<uint8_t>(shr<uint32_t>(c,cast<int64_t>(8ULL))),cast<uint8_t>(shr<uint32_t>(c,cast<int64_t>(16ULL))),cast<uint8_t>(shr<uint32_t>(c,cast<int64_t>(24ULL)))};
}
}
// tools/platform/psp/ge_draw.go:373:1
std::tuple<uint8_t,uint8_t,uint8_t,uint8_t> psp_decode16a(uint16_t p,uint32_t fmt){
{
auto ext = [&](uint16_t v,uint16_t bits)->uint8_t{
v &= cast<uint16_t>(((shl<uint16_t>(cast<uint16_t>(1ULL),bits)) - cast<uint16_t>(1ULL)));
return cast<uint8_t>(divi<uint32_t>((cast<uint32_t>((cast<uint32_t>(v) * cast<uint32_t>(255ULL)))),(cast<uint32_t>(((shl<uint32_t>(cast<uint32_t>(1ULL),bits)) - cast<uint32_t>(1ULL))))));
}
;
{
switch(fmt){
case cast<uint32_t>(1ULL):{
uint8_t a = cast<uint8_t>(0ULL);
if ((cast<uint16_t>((p & cast<uint16_t>(32768ULL))) != cast<uint16_t>(0ULL))) {
a = cast<uint8_t>(255ULL);
}
return {ext(p,cast<uint16_t>(5ULL)),ext(shr<uint16_t>(p,cast<int64_t>(5ULL)),cast<uint16_t>(5ULL)),ext(shr<uint16_t>(p,cast<int64_t>(10ULL)),cast<uint16_t>(5ULL)),a};
break;}
case cast<uint32_t>(2ULL):{
return {ext(p,cast<uint16_t>(4ULL)),ext(shr<uint16_t>(p,cast<int64_t>(4ULL)),cast<uint16_t>(4ULL)),ext(shr<uint16_t>(p,cast<int64_t>(8ULL)),cast<uint16_t>(4ULL)),ext(shr<uint16_t>(p,cast<int64_t>(12ULL)),cast<uint16_t>(4ULL))};
break;}
default:{
return {ext(p,cast<uint16_t>(5ULL)),ext(shr<uint16_t>(p,cast<int64_t>(5ULL)),cast<uint16_t>(6ULL)),ext(shr<uint16_t>(p,cast<int64_t>(11ULL)),cast<uint16_t>(5ULL)),cast<uint8_t>(255ULL)};
break;}
}}
}
}
// tools/platform/psp/ge_draw.go:395:1
bool psp_geState_stencilTest(psp_geState* s,uint8_t dstA){
{
if (((!s->stencilOn) || s->clearOn)) {
return true;
}
uint32_t ref = cast<uint32_t>((s->stRef & s->stMask));
uint32_t cur = cast<uint32_t>((cast<uint32_t>(dstA) & s->stMask));
bool pass = false;
{
switch(s->stFunc){
case cast<uint32_t>(0ULL):{
break;}
case cast<uint32_t>(1ULL):{
pass = true;
break;}
case cast<uint32_t>(2ULL):{
pass = (ref == cur);
break;}
case cast<uint32_t>(3ULL):{
pass = (ref != cur);
break;}
case cast<uint32_t>(4ULL):{
pass = (ref < cur);
break;}
case cast<uint32_t>(5ULL):{
pass = (ref <= cur);
break;}
case cast<uint32_t>(6ULL):{
pass = (ref > cur);
break;}
default:{
pass = (ref >= cur);
break;}
}}
return pass;
}
}
// tools/platform/psp/ge_draw.go:423:1
uint8_t psp_stencilOp(uint32_t op,uint8_t cur,uint8_t ref){
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
return cast<uint8_t>(~cur);
break;}
case cast<uint32_t>(4ULL):{
if ((cur < cast<uint8_t>(255ULL))) {
return cast<uint8_t>((cur + cast<uint8_t>(1ULL)));
}
return cast<uint8_t>(255ULL);
break;}
case cast<uint32_t>(5ULL):{
if ((cur > cast<uint8_t>(0ULL))) {
return cast<uint8_t>((cur - cast<uint8_t>(1ULL)));
}
return cast<uint8_t>(0ULL);
break;}
default:{
return cur;
break;}
}}
}
}
// tools/platform/psp/ge_draw.go:449:1
bool psp_geState_alphaPass(psp_geState* s,uint8_t a){
{
if (((!s->alphaTestOn) || s->clearOn)) {
return true;
}
uint32_t src = cast<uint32_t>((cast<uint32_t>(a) & s->alphaTestMask));
uint32_t ref = cast<uint32_t>((s->alphaRef & s->alphaTestMask));
{
switch(s->alphaFunc){
case cast<uint32_t>(0ULL):{
return false;
break;}
case cast<uint32_t>(1ULL):{
return true;
break;}
case cast<uint32_t>(2ULL):{
return (src == ref);
break;}
case cast<uint32_t>(3ULL):{
return (src != ref);
break;}
case cast<uint32_t>(4ULL):{
return (src < ref);
break;}
case cast<uint32_t>(5ULL):{
return (src <= ref);
break;}
case cast<uint32_t>(6ULL):{
return (src > ref);
break;}
default:{
return (src >= ref);
break;}
}}
}
}
// tools/platform/psp/ge_draw.go:477:1
uint32_t psp_blendFactor(uint32_t f,uint8_t sc,uint8_t sa,uint8_t da,uint32_t fix,int64_t ch){
{
uint8_t fixCh = cast<uint8_t>(shr<uint32_t>(fix,(cast<int64_t>((cast<int64_t>(8ULL) * ch)))));
{
switch(f){
case cast<uint32_t>(0ULL):{
return cast<uint32_t>(sc);
break;}
case cast<uint32_t>(1ULL):{
return cast<uint32_t>((cast<uint32_t>(255ULL) - cast<uint32_t>(sc)));
break;}
case cast<uint32_t>(2ULL):{
return cast<uint32_t>(sa);
break;}
case cast<uint32_t>(3ULL):{
return cast<uint32_t>((cast<uint32_t>(255ULL) - cast<uint32_t>(sa)));
break;}
case cast<uint32_t>(4ULL):{
return cast<uint32_t>(da);
break;}
case cast<uint32_t>(5ULL):{
return cast<uint32_t>((cast<uint32_t>(255ULL) - cast<uint32_t>(da)));
break;}
case cast<uint32_t>(6ULL):{
return psp_min32(cast<uint32_t>((cast<uint32_t>(2ULL) * cast<uint32_t>(sa))),cast<uint32_t>(255ULL));
break;}
case cast<uint32_t>(7ULL):{
return cast<uint32_t>((cast<uint32_t>(255ULL) - psp_min32(cast<uint32_t>((cast<uint32_t>(2ULL) * cast<uint32_t>(sa))),cast<uint32_t>(255ULL))));
break;}
case cast<uint32_t>(8ULL):{
return psp_min32(cast<uint32_t>((cast<uint32_t>(2ULL) * cast<uint32_t>(da))),cast<uint32_t>(255ULL));
break;}
case cast<uint32_t>(9ULL):{
return cast<uint32_t>((cast<uint32_t>(255ULL) - psp_min32(cast<uint32_t>((cast<uint32_t>(2ULL) * cast<uint32_t>(da))),cast<uint32_t>(255ULL))));
break;}
default:{
return cast<uint32_t>(fixCh);
break;}
}}
}
}
// tools/platform/psp/ge_draw.go:505:1
uint32_t psp_min32(uint32_t a,uint32_t b){
{
if ((a < b)) {
return a;
}
return b;
}
}
// tools/platform/psp/ge_draw.go:512:1
uint8_t psp_clamp255(int32_t v){
{
if ((v < cast<int32_t>(0ULL))) {
return cast<uint8_t>(0ULL);
}
if ((v > cast<int32_t>(255ULL))) {
return cast<uint8_t>(255ULL);
}
return cast<uint8_t>(v);
}
}
// tools/platform/psp/ge_draw.go:530:1
void psp_Machine_putPixel(psp_Machine* m,psp_geState* s,int64_t x,int64_t y,float z,uint8_t r,uint8_t g,uint8_t b,uint8_t a){
{
if ((((!s->clearOn) && (((((x < s->scX0) || (x > s->scX1)) || (y < s->scY0)) || (y > s->scY1)))) && (s->scX1 > cast<int64_t>(0ULL)))) {
psp_Machine_pixelEvent(m,x,y,psp_PixelEvent{{},true,{},{},{},{},r,g,b,a});
return ;
}
if ((!psp_geState_alphaPass(s,a))) {
psp_Machine_pixelEvent(m,x,y,psp_PixelEvent{{},{},true,{},{},{},r,g,b,a});
return ;
}
uint32_t base = psp_geState_fbAddress(s);
uint32_t stride = s->fbStride;
if ((stride == cast<uint32_t>(0ULL))) {
stride = cast<uint32_t>(480ULL);
}
uint32_t off = cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(y) * stride)) + cast<uint32_t>(x)));
if (((x == psp_geProbeX) && (y == psp_geProbeY))) {
go_fmt_Printf(std::string("PIXEL(%d,%d) prim#%d rgba=%02X%02X%02X%02X blend=%v(%d,%d) atest=%v(fn%d ref%d) wrap=(%d,%d) map=%d/src%d maxlvl=%d bias=%.2f tex=%08X f%d fn%d dbl=%v mat=%08X\012",160),x,y,psp_gePrimSeq,r,g,b,a,s->blendOn,s->blendSrc,s->blendDst,s->alphaTestOn,s->alphaFunc,s->alphaRef,s->texWrapU,s->texWrapV,s->texMapMode,s->texProjSrc,s->texMaxLvl,s->texLodBias,s->texAddr,s->texFmt,s->texFunc,s->texDouble,s->matColor);
}
uint8_t dstA = psp_Machine_dstAlpha(m,s,base,off);
bool stPass = psp_geState_stencilTest(s,dstA);
if ((!stPass)) {
if ((!s->clearOn)) {
psp_Machine_storeAlpha(m,s,base,off,psp_stencilOp(s->stSFail,dstA,cast<uint8_t>(s->stRef)));
}
psp_Machine_pixelEvent(m,x,y,psp_PixelEvent{{},{},{},true,{},{},r,g,b,a});
return ;
}
bool zPass = psp_Machine_zTest(m,s,x,y,z);
if ((!zPass)) {
if (((!s->clearOn) && s->stencilOn)) {
psp_Machine_storeAlpha(m,s,base,off,psp_stencilOp(s->stZFail,dstA,cast<uint8_t>(s->stRef)));
}
psp_Machine_pixelEvent(m,x,y,psp_PixelEvent{{},{},{},{},true,{},r,g,b,a});
return ;
}
psp_Machine_zWrite(m,s,x,y,z);
uint8_t outA = a;
if ((s->stencilOn && (!s->clearOn))) {
outA = psp_stencilOp(s->stZPass,dstA,cast<uint8_t>(s->stRef));
}
if ((s->blendOn && (!s->clearOn))) {
auto tmp61 = psp_Machine_dstPixel4(m,s,base,off);
uint8_t dr = std::get<0>(tmp61);
uint8_t dg = std::get<1>(tmp61);
uint8_t db = std::get<2>(tmp61);
uint8_t da = std::get<3>(tmp61);
auto ch = [&](uint8_t sc,uint8_t dc,int64_t i)->uint8_t{
uint32_t sf = psp_blendFactor(s->blendSrc,sc,a,da,s->blendFixA,i);
uint32_t df = psp_blendFactor(s->blendDst,sc,a,da,s->blendFixB,i);
int32_t sv = cast<int32_t>(divi<uint32_t>(cast<uint32_t>((cast<uint32_t>(sc) * sf)),cast<uint32_t>(255ULL)));
int32_t dv = cast<int32_t>(divi<uint32_t>(cast<uint32_t>((cast<uint32_t>(dc) * df)),cast<uint32_t>(255ULL)));
{
switch(s->blendEq){
case cast<uint32_t>(1ULL):{
return psp_clamp255(cast<int32_t>((sv - dv)));
break;}
case cast<uint32_t>(2ULL):{
return psp_clamp255(cast<int32_t>((dv - sv)));
break;}
case cast<uint32_t>(3ULL):{
return psp_clamp255(psp_minI32(cast<int32_t>(sc),cast<int32_t>(dc)));
break;}
case cast<uint32_t>(4ULL):{
return psp_clamp255(psp_maxI32(cast<int32_t>(sc),cast<int32_t>(dc)));
break;}
case cast<uint32_t>(5ULL):{
return psp_clamp255(psp_absI32(cast<int32_t>((cast<int32_t>(sc) - cast<int32_t>(dc)))));
break;}
default:{
return psp_clamp255(cast<int32_t>((sv + dv)));
break;}
}}
}
;
auto tmp62 = std::make_tuple(ch(r,dr,cast<int64_t>(0ULL)),ch(g,dg,cast<int64_t>(1ULL)),ch(b,db,cast<int64_t>(2ULL)));
r = std::get<0>(tmp62);
g = std::get<1>(tmp62);
b = std::get<2>(tmp62);
}
if (((s->maskRGB != cast<uint32_t>(0ULL)) || (s->maskA != cast<uint32_t>(0ULL)))) {
if (((s->maskRGB == cast<uint32_t>(16777215ULL)) && (s->maskA == cast<uint32_t>(255ULL)))) {
psp_Machine_pixelEvent(m,x,y,psp_PixelEvent{{},{},{},{},{},true,r,g,b,a});
return ;
}
auto tmp63 = psp_Machine_dstPixel4(m,s,base,off);
uint8_t dr = std::get<0>(tmp63);
uint8_t dg = std::get<1>(tmp63);
uint8_t db = std::get<2>(tmp63);
uint8_t da = std::get<3>(tmp63);
auto tmp64 = std::make_tuple(cast<uint8_t>(s->maskRGB),cast<uint8_t>(shr<uint32_t>(s->maskRGB,cast<int64_t>(8ULL))),cast<uint8_t>(shr<uint32_t>(s->maskRGB,cast<int64_t>(16ULL))),cast<uint8_t>(s->maskA));
uint8_t mr = std::get<0>(tmp64);
uint8_t mg = std::get<1>(tmp64);
uint8_t mb = std::get<2>(tmp64);
uint8_t ma = std::get<3>(tmp64);
r = cast<uint8_t>(((r & ~(mr)) | cast<uint8_t>((dr & mr))));
g = cast<uint8_t>(((g & ~(mg)) | cast<uint8_t>((dg & mg))));
b = cast<uint8_t>(((b & ~(mb)) | cast<uint8_t>((db & mb))));
outA = cast<uint8_t>(((outA & ~(ma)) | cast<uint8_t>((da & ma))));
}
psp_Machine_storePixel(m,s,base,off,r,g,b,outA);
psp_Machine_pixelEvent(m,x,y,psp_PixelEvent{true,{},{},{},{},{},r,g,b,outA});
}
}
// tools/platform/psp/ge_draw.go:627:1
void psp_Machine_storePixel(psp_Machine* m,psp_geState* s,uint32_t base,uint32_t off,uint8_t r,uint8_t g,uint8_t b,uint8_t a){
{
{
switch(s->fbFmt){
case cast<uint32_t>(3ULL):{
psp_Machine_write32(m,cast<uint32_t>((base + cast<uint32_t>((off * cast<uint32_t>(4ULL))))),cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(r) | shl<uint32_t>(cast<uint32_t>(g),cast<int64_t>(8ULL)))) | shl<uint32_t>(cast<uint32_t>(b),cast<int64_t>(16ULL)))) | shl<uint32_t>(cast<uint32_t>(a),cast<int64_t>(24ULL)))));
break;}
default:{
uint16_t p={};
{
switch(s->fbFmt){
case cast<uint32_t>(1ULL):{
p = cast<uint16_t>((cast<uint16_t>((cast<uint16_t>(shr<uint8_t>(r,cast<int64_t>(3ULL))) | shl<uint16_t>(cast<uint16_t>(shr<uint8_t>(g,cast<int64_t>(3ULL))),cast<int64_t>(5ULL)))) | shl<uint16_t>(cast<uint16_t>(shr<uint8_t>(b,cast<int64_t>(3ULL))),cast<int64_t>(10ULL))));
if ((a >= cast<uint8_t>(128ULL))) {
p |= cast<uint16_t>(32768ULL);
}
break;}
case cast<uint32_t>(2ULL):{
p = cast<uint16_t>((cast<uint16_t>((cast<uint16_t>((cast<uint16_t>(shr<uint8_t>(r,cast<int64_t>(4ULL))) | shl<uint16_t>(cast<uint16_t>(shr<uint8_t>(g,cast<int64_t>(4ULL))),cast<int64_t>(4ULL)))) | shl<uint16_t>(cast<uint16_t>(shr<uint8_t>(b,cast<int64_t>(4ULL))),cast<int64_t>(8ULL)))) | shl<uint16_t>(cast<uint16_t>(shr<uint8_t>(a,cast<int64_t>(4ULL))),cast<int64_t>(12ULL))));
break;}
default:{
p = cast<uint16_t>((cast<uint16_t>((cast<uint16_t>(shr<uint8_t>(r,cast<int64_t>(3ULL))) | shl<uint16_t>(cast<uint16_t>(shr<uint8_t>(g,cast<int64_t>(2ULL))),cast<int64_t>(5ULL)))) | shl<uint16_t>(cast<uint16_t>(shr<uint8_t>(b,cast<int64_t>(3ULL))),cast<int64_t>(11ULL))));
break;}
}}
uint32_t addr = cast<uint32_t>((base + cast<uint32_t>((off * cast<uint32_t>(2ULL)))));
psp_Machine_write16(m,addr,p);
break;}
}}
}
}
// tools/platform/psp/ge_draw.go:651:1
uint8_t psp_Machine_dstAlpha(psp_Machine* m,psp_geState* s,uint32_t base,uint32_t off){
{
auto tmp65 = psp_Machine_dstPixel4(m,s,base,off);
uint8_t a = std::get<3>(tmp65);
return a;
}
}
// tools/platform/psp/ge_draw.go:658:1
void psp_Machine_storeAlpha(psp_Machine* m,psp_geState* s,uint32_t base,uint32_t off,uint8_t a){
{
auto tmp66 = psp_Machine_dstPixel4(m,s,base,off);
uint8_t r = std::get<0>(tmp66);
uint8_t g = std::get<1>(tmp66);
uint8_t b = std::get<2>(tmp66);
psp_Machine_storePixel(m,s,base,off,r,g,b,a);
}
}
// tools/platform/psp/ge_draw.go:664:1
std::tuple<uint8_t,uint8_t,uint8_t> psp_Machine_dstPixel(psp_Machine* m,psp_geState* s,uint32_t base,uint32_t off){
{
auto tmp67 = psp_Machine_dstPixel4(m,s,base,off);
uint8_t r = std::get<0>(tmp67);
uint8_t g = std::get<1>(tmp67);
uint8_t b = std::get<2>(tmp67);
return {r,g,b};
}
}
// tools/platform/psp/ge_draw.go:670:1
std::tuple<uint8_t,uint8_t,uint8_t,uint8_t> psp_Machine_dstPixel4(psp_Machine* m,psp_geState* s,uint32_t base,uint32_t off){
{
if ((s->fbFmt == cast<uint32_t>(3ULL))) {
uint32_t c = psp_Machine_read32(m,cast<uint32_t>((base + cast<uint32_t>((off * cast<uint32_t>(4ULL))))));
return {cast<uint8_t>(c),cast<uint8_t>(shr<uint32_t>(c,cast<int64_t>(8ULL))),cast<uint8_t>(shr<uint32_t>(c,cast<int64_t>(16ULL))),cast<uint8_t>(shr<uint32_t>(c,cast<int64_t>(24ULL)))};
}
return psp_decode16a(psp_u16(m,cast<uint32_t>((base + cast<uint32_t>((off * cast<uint32_t>(2ULL)))))),s->fbFmt);
}
}
// tools/platform/psp/ge_draw.go:680:1
float psp_edge(psp_vert a,psp_vert b,psp_vert c){
{
return ((((c.x - a.x)) * ((b.y - a.y))) - (((c.y - a.y)) * ((b.x - a.x))));
}
}
// tools/platform/psp/ge_draw.go:682:1
uint8_t psp_mul8(uint8_t a,uint8_t b){
{
return cast<uint8_t>(divi<uint32_t>(cast<uint32_t>((cast<uint32_t>(a) * cast<uint32_t>(b))),cast<uint32_t>(255ULL)));
}
}
// tools/platform/psp/ge_draw.go:684:1
float psp_fmin3(float a,float b,float c){
{
return psp_fmin(psp_fmin(a,b),c);
}
}
// tools/platform/psp/ge_draw.go:685:1
float psp_fmax3(float a,float b,float c){
{
return psp_fmax(psp_fmax(a,b),c);
}
}
// tools/platform/psp/ge_draw.go:686:1
float psp_fmin(float a,float b){
{
if ((a < b)) {
return a;
}
return b;
}
}
// tools/platform/psp/ge_draw.go:692:1
float psp_fmax(float a,float b){
{
if ((a > b)) {
return a;
}
return b;
}
}
// tools/platform/psp/ge_draw.go:698:1
int64_t psp_clampI(int64_t v,int64_t lo,int64_t hi){
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
// tools/platform/psp/ge_draw.go:708:1
int32_t psp_minI32(int32_t a,int32_t b){
{
if ((a < b)) {
return a;
}
return b;
}
}
// tools/platform/psp/ge_draw.go:715:1
int32_t psp_maxI32(int32_t a,int32_t b){
{
if ((a > b)) {
return a;
}
return b;
}
}
// tools/platform/psp/ge_draw.go:722:1
int32_t psp_absI32(int32_t a){
{
if ((a < cast<int32_t>(0ULL))) {
return cast<int32_t>(-a);
}
return a;
}
}
// tools/platform/psp/ge_patch.go:14:1
void psp_Machine_drawPatch(psp_Machine* m,psp_geState* s,uint32_t arg,bool spline){
{rrprof::Scope timing(2,"GE vertices and software rasterizer");
{rrTextureScope textures(m,s);
{
int64_t nu = cast<int64_t>(cast<uint32_t>((arg & cast<uint32_t>(255ULL))));
int64_t nv = cast<int64_t>(cast<uint32_t>(((shr<uint32_t>(arg,cast<int64_t>(8ULL))) & cast<uint32_t>(255ULL))));
if ((((nu < cast<int64_t>(4ULL)) || (nv < cast<int64_t>(4ULL))) || (cast<int64_t>((nu * nv)) > cast<int64_t>(4096ULL)))) {
return ;
}
psp_gePrimSeq++;
if ((psp_gePrimSeq == psp_gePrimDump)) {
psp_Machine_dumpPrim(m,s,cast<uint32_t>(9ULL),cast<int64_t>((nu * nv)));
}
if ((psp_gePrimDump >= cast<int64_t>(0ULL))) {
psp_geWordTrail = sub(psp_geWordTrail,0,cast<int64_t>(0ULL));
}
Slice<psp_vert> ctrl = psp_Machine_decodeVerts(m,s,cast<int64_t>((nu * nv)));
if ((len(ctrl) < cast<int64_t>((nu * nv)))) {
return ;
}
if ((psp_geDebugN > cast<int64_t>(0ULL))) {
psp_geDebugN--;
std::string kind = std::string("BEZIER",6);
if (spline) {
kind = std::string("SPLINE",6);
}
std::string logf = std::string("%s#%d %dx%d vt=%06X va=%08X ia=%08X div=%d,%d tex=%v@%08X mat=%08X dif=%06X light=%v c0=(%.1f,%.1f,%.1f)\012",105);
psp_fmtPrintf(logf,kind,psp_gePrimSeq,nu,nv,s->vtype,s->vaddr,s->iaddr,s->patchDivU,s->patchDivV,s->texEnable,s->texAddr,s->matColor,s->matDiffuse,s->lightOn,ctrl[cast<int64_t>(0ULL)].x,ctrl[cast<int64_t>(0ULL)].y,ctrl[cast<int64_t>(0ULL)].z);
}
int64_t divU = cast<int64_t>(s->patchDivU);
int64_t divV = cast<int64_t>(s->patchDivV);
if ((divU < cast<int64_t>(1ULL))) {
divU = cast<int64_t>(4ULL);
}
if ((divV < cast<int64_t>(1ULL))) {
divV = cast<int64_t>(4ULL);
}
int64_t su={};
int64_t sv={};
if (spline) {
su = cast<int64_t>((cast<int64_t>(((cast<int64_t>((nu - cast<int64_t>(3ULL)))) * divU)) + cast<int64_t>(1ULL)));
sv = cast<int64_t>((cast<int64_t>(((cast<int64_t>((nv - cast<int64_t>(3ULL)))) * divV)) + cast<int64_t>(1ULL)));
}
else {
su = cast<int64_t>(((cast<int64_t>((divi<int64_t>(nu,cast<int64_t>(3ULL)) * divU))) + cast<int64_t>(1ULL)));
sv = cast<int64_t>(((cast<int64_t>((divi<int64_t>(nv,cast<int64_t>(3ULL)) * divV))) + cast<int64_t>(1ULL)));
}
if ((((su < cast<int64_t>(2ULL)) || (sv < cast<int64_t>(2ULL))) || (cast<int64_t>((su * sv)) > cast<int64_t>(65536ULL)))) {
return ;
}
bool clampU = ((!spline) || (cast<uint32_t>((arg & cast<uint32_t>(65536ULL))) != cast<uint32_t>(0ULL)));
bool clampV = ((!spline) || (cast<uint32_t>((arg & cast<uint32_t>(131072ULL))) != cast<uint32_t>(0ULL)));
auto tmp68 = std::make_tuple(cast<uint8_t>(s->matColor),cast<uint8_t>(shr<uint32_t>(s->matColor,cast<int64_t>(8ULL))),cast<uint8_t>(shr<uint32_t>(s->matColor,cast<int64_t>(16ULL))));
uint8_t cr = std::get<0>(tmp68);
uint8_t cg = std::get<1>(tmp68);
uint8_t cb = std::get<2>(tmp68);
uint8_t ca = cast<uint8_t>(shr<uint32_t>(s->matColor,cast<int64_t>(24ULL)));
if ((s->lightOn && (cast<uint32_t>(((shr<uint32_t>(s->vtype,cast<int64_t>(2ULL))) & cast<uint32_t>(7ULL))) == cast<uint32_t>(0ULL)))) {
auto tmp69 = std::make_tuple(cast<uint8_t>(s->matDiffuse),cast<uint8_t>(shr<uint32_t>(s->matDiffuse,cast<int64_t>(8ULL))),cast<uint8_t>(shr<uint32_t>(s->matDiffuse,cast<int64_t>(16ULL))));
cr = std::get<0>(tmp69);
cg = std::get<1>(tmp69);
cb = std::get<2>(tmp69);
}
bool hasUV = (cast<uint32_t>((s->vtype & cast<uint32_t>(3ULL))) != cast<uint32_t>(0ULL));
Slice<psp_vert> grid = Slice<psp_vert>::make(cast<int64_t>((su * sv)));
{int64_t j = cast<int64_t>(0ULL);for (;(j < sv);j++){
float tv = (cast<float>(j) / cast<float>(cast<int64_t>((sv - cast<int64_t>(1ULL)))));
{int64_t i = cast<int64_t>(0ULL);for (;(i < su);i++){
float tu = (cast<float>(i) / cast<float>(cast<int64_t>((su - cast<int64_t>(1ULL)))));
float x={};
float y={};
float z={};
float uu={};
float vv={};
if (spline) {
auto tmp70 = psp_splineSurf(ctrl,nu,nv,tu,tv,clampU,clampV,hasUV);
x = std::get<0>(tmp70);
y = std::get<1>(tmp70);
z = std::get<2>(tmp70);
uu = std::get<3>(tmp70);
vv = std::get<4>(tmp70);
}
else {
auto tmp71 = psp_bezierSurf(ctrl,nu,nv,tu,tv,hasUV);
x = std::get<0>(tmp71);
y = std::get<1>(tmp71);
z = std::get<2>(tmp71);
uu = std::get<3>(tmp71);
vv = std::get<4>(tmp71);
}
psp_vert out={};
auto tmp72 = std::make_tuple(x,y,z);
out.x = std::get<0>(tmp72);
out.y = std::get<1>(tmp72);
out.z = std::get<2>(tmp72);
if (hasUV) {
auto tmp73 = std::make_tuple(uu,vv);
out.u = std::get<0>(tmp73);
out.v = std::get<1>(tmp73);
}
else {
out.u = ((tu * s->texScaleU) + s->texOffU);
out.v = ((tv * s->texScaleV) + s->texOffV);
}
auto tmp74 = std::make_tuple(cr,cg,cb,ca);
out.r = std::get<0>(tmp74);
out.g = std::get<1>(tmp74);
out.b = std::get<2>(tmp74);
out.a = std::get<3>(tmp74);
grid[cast<int64_t>((cast<int64_t>((j * su)) + i))] = out;
}
}}
}{int64_t j = cast<int64_t>(0ULL);for (;(cast<int64_t>((j + cast<int64_t>(1ULL))) < sv);j++){
{int64_t i = cast<int64_t>(0ULL);for (;(cast<int64_t>((i + cast<int64_t>(1ULL))) < su);i++){
psp_vert a = grid[cast<int64_t>((cast<int64_t>((j * su)) + i))];
psp_vert b = grid[cast<int64_t>((cast<int64_t>((cast<int64_t>((j * su)) + i)) + cast<int64_t>(1ULL)))];
psp_vert c = grid[cast<int64_t>((cast<int64_t>(((cast<int64_t>((j + cast<int64_t>(1ULL)))) * su)) + i))];
psp_vert d = grid[cast<int64_t>((cast<int64_t>((cast<int64_t>(((cast<int64_t>((j + cast<int64_t>(1ULL)))) * su)) + i)) + cast<int64_t>(1ULL)))];
psp_Machine_rasterTri(m,s,a,b,c);
psp_Machine_rasterTri(m,s,b,d,c);
}
}}
}}
}
}
}
// tools/platform/psp/ge_patch.go:118:1
std::tuple<float,float,float,float,float> psp_splineSurf(Slice<psp_vert> ctrl,int64_t nu,int64_t nv,float tu,float tv,bool clampU,bool clampV,bool hasUV){
float x{};
float y{};
float z{};
float u{};
float v{};
{
Slice<float> ku = psp_splineKnots(nu,clampU);
Slice<float> kv = psp_splineKnots(nv,clampV);
float pu = (ku[cast<int64_t>(3ULL)] + (tu * ((ku[nu] - ku[cast<int64_t>(3ULL)]))));
float pv = (kv[cast<int64_t>(3ULL)] + (tv * ((kv[nv] - kv[cast<int64_t>(3ULL)]))));
Slice<float> wu = psp_splineBasis(ku,nu,pu);
Slice<float> wv = psp_splineBasis(kv,nv,pv);
{int64_t j = cast<int64_t>(0ULL);for (;(j < nv);j++){
if ((wv[j] == cast<float>(0ULL))) {
continue;
}
{int64_t i = cast<int64_t>(0ULL);for (;(i < nu);i++){
float w = (wu[i] * wv[j]);
if ((w == cast<float>(0ULL))) {
continue;
}
psp_vert p = ctrl[cast<int64_t>((cast<int64_t>((j * nu)) + i))];
x += (w * p.x);
y += (w * p.y);
z += (w * p.z);
if (hasUV) {
u += (w * p.u);
v += (w * p.v);
}
}
}}
}return {x,y,z,u,v};
}
}
// tools/platform/psp/ge_patch.go:151:1
Slice<float> psp_splineKnots(int64_t n,bool clamped){
{
Slice<float> k = Slice<float>::make(cast<int64_t>((n + cast<int64_t>(4ULL))));
{auto&& tmp75 = k;
for(int64_t tmp76=0;tmp76<len(tmp75);++tmp76){
auto i=tmp76;k[i] = cast<float>(i);
}}
if (clamped) {
auto tmp77 = std::make_tuple(k[cast<int64_t>(3ULL)],k[cast<int64_t>(3ULL)],k[cast<int64_t>(3ULL)]);
k[cast<int64_t>(0ULL)] = std::get<0>(tmp77);
k[cast<int64_t>(1ULL)] = std::get<1>(tmp77);
k[cast<int64_t>(2ULL)] = std::get<2>(tmp77);
auto tmp78 = std::make_tuple(k[n],k[n],k[n]);
k[cast<int64_t>((n + cast<int64_t>(1ULL)))] = std::get<0>(tmp78);
k[cast<int64_t>((n + cast<int64_t>(2ULL)))] = std::get<1>(tmp78);
k[cast<int64_t>((n + cast<int64_t>(3ULL)))] = std::get<2>(tmp78);
}
return k;
}
}
// tools/platform/psp/ge_patch.go:165:1
Slice<float> psp_splineBasis(Slice<float> knot,int64_t n,float t){
{
if ((t >= knot[n])) {
t = (knot[n] - 1e-4);
}
if ((t < knot[cast<int64_t>(3ULL)])) {
t = knot[cast<int64_t>(3ULL)];
}
Slice<float> w = Slice<float>::make(cast<int64_t>((n + cast<int64_t>(3ULL))));
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>((n + cast<int64_t>(3ULL))));i++){
if (((knot[i] <= t) && (t < knot[cast<int64_t>((i + cast<int64_t>(1ULL)))]))) {
w[i] = cast<float>(1ULL);
}
}
}{int64_t d = cast<int64_t>(1ULL);for (;(d <= cast<int64_t>(3ULL));d++){
{int64_t i = cast<int64_t>(0ULL);for (;(cast<int64_t>((cast<int64_t>((i + d)) + cast<int64_t>(1ULL))) < cast<int64_t>((n + cast<int64_t>(4ULL))));i++){
float a={};
float b={};
{
float den = (knot[cast<int64_t>((i + d))] - knot[i]);
if ((den != cast<float>(0ULL))) {
a = ((((t - knot[i])) / den) * w[i]);
}
}
{
float den = (knot[cast<int64_t>((cast<int64_t>((i + d)) + cast<int64_t>(1ULL)))] - knot[cast<int64_t>((i + cast<int64_t>(1ULL)))]);
if ((den != cast<float>(0ULL))) {
b = ((((knot[cast<int64_t>((cast<int64_t>((i + d)) + cast<int64_t>(1ULL)))] - t)) / den) * w[cast<int64_t>((i + cast<int64_t>(1ULL)))]);
}
}
w[i] = (a + b);
}
}}
}return sub(w,0,n);
}
}
// tools/platform/psp/ge_patch.go:198:1
std::tuple<float,float,float,float,float> psp_bezierSurf(Slice<psp_vert> ctrl,int64_t nu,int64_t nv,float tu,float tv,bool hasUV){
float x{};
float y{};
float z{};
float u{};
float v{};
{
Slice<float> wu = psp_bezierBasis(nu,tu);
Slice<float> wv = psp_bezierBasis(nv,tv);
{int64_t j = cast<int64_t>(0ULL);for (;(j < nv);j++){
if ((wv[j] == cast<float>(0ULL))) {
continue;
}
{int64_t i = cast<int64_t>(0ULL);for (;(i < nu);i++){
float w = (wu[i] * wv[j]);
if ((w == cast<float>(0ULL))) {
continue;
}
psp_vert p = ctrl[cast<int64_t>((cast<int64_t>((j * nu)) + i))];
x += (w * p.x);
y += (w * p.y);
z += (w * p.z);
if (hasUV) {
u += (w * p.u);
v += (w * p.v);
}
}
}}
}return {x,y,z,u,v};
}
}
// tools/platform/psp/ge_patch.go:226:1
Slice<float> psp_bezierBasis(int64_t n,float t){
{
Slice<float> w = Slice<float>::make(n);
int64_t patches = divi<int64_t>((cast<int64_t>((n - cast<int64_t>(1ULL)))),cast<int64_t>(3ULL));
float pt = (t * cast<float>(patches));
int64_t pi = cast<int64_t>(pt);
if ((pi >= patches)) {
pi = cast<int64_t>((patches - cast<int64_t>(1ULL)));
}
float lt = (pt - cast<float>(pi));
float it = (cast<float>(1ULL) - lt);
float b0 = ((it * it) * it);
float b1 = (((cast<float>(3ULL) * lt) * it) * it);
float b2 = (((cast<float>(3ULL) * lt) * lt) * it);
float b3 = ((lt * lt) * lt);
int64_t base = cast<int64_t>((pi * cast<int64_t>(3ULL)));
w[base] += b0;
w[cast<int64_t>((base + cast<int64_t>(1ULL)))] += b1;
w[cast<int64_t>((base + cast<int64_t>(2ULL)))] += b2;
w[cast<int64_t>((base + cast<int64_t>(3ULL)))] += b3;
return w;
}
}
// tools/platform/psp/ge_raster.go:297:1
uint32_t psp_geState_zAddress(psp_geState* s){
{
return cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(67108864ULL) | (shl<uint32_t>(s->zHigh,cast<int64_t>(24ULL))))) | s->zLow));
}
}
// tools/platform/psp/ge_raster.go:302:1
void psp_Machine_blockTransfer(psp_Machine* m,psp_geState* s,bool is32){
{
if (((((s->trSrcStride == cast<uint32_t>(0ULL)) || (s->trDstStride == cast<uint32_t>(0ULL))) || (s->trW == cast<uint32_t>(0ULL))) || (s->trH == cast<uint32_t>(0ULL)))) {
return ;
}
m->geCnt.xfers++;
auto tmp83=defer([&](){psp_Machine_profEnd(m,cast<int64_t>(3ULL),psp_Machine_profStart(m));});
uint32_t bpp = cast<uint32_t>(2ULL);
if (is32) {
bpp = cast<uint32_t>(4ULL);
}
if (psp_geXferLog) {
go_fmt_Printf(std::string("XFER src=%08X(+%d,%d stride %d) -> dst=%08X(+%d,%d stride %d)  %dx%d  %d bpp\012",77),s->trSrc,s->trSrcX,s->trSrcY,s->trSrcStride,s->trDst,s->trDstX,s->trDstY,s->trDstStride,s->trW,s->trH,bpp);
}
{uint32_t y = cast<uint32_t>(0ULL);for (;(y < s->trH);y++){
uint32_t src = cast<uint32_t>((s->trSrc + cast<uint32_t>(((cast<uint32_t>((cast<uint32_t>(((cast<uint32_t>((s->trSrcY + y))) * s->trSrcStride)) + s->trSrcX))) * bpp))));
uint32_t dst = cast<uint32_t>((s->trDst + cast<uint32_t>(((cast<uint32_t>((cast<uint32_t>(((cast<uint32_t>((s->trDstY + y))) * s->trDstStride)) + s->trDstX))) * bpp))));
{uint32_t x = cast<uint32_t>(0ULL);for (;(x < cast<uint32_t>((s->trW * bpp)));x++){
psp_Machine_Write(m,cast<uint32_t>((dst + x)),psp_Machine_Read(m,cast<uint32_t>((src + x))));
}
}}
}}
}
// tools/platform/psp/ge_raster.go:329:1
bool psp_Machine_zTest(psp_Machine* m,psp_geState* s,int64_t x,int64_t y,float z){
{
if ((((((s->zLow == cast<uint32_t>(0ULL)) || (s->zStride == cast<uint32_t>(0ULL))) || s->clearOn) || (!s->zTestOn)) || psp_geNoZ)) {
return true;
}
uint32_t zi = cast<uint32_t>(psp_clampF(z,cast<float>(0ULL),cast<float>(65535ULL)));
uint32_t old = cast<uint32_t>(psp_u16(m,cast<uint32_t>((psp_geState_zAddress(s) + cast<uint32_t>(((cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(y) * s->zStride)) + cast<uint32_t>(x)))) * cast<uint32_t>(2ULL)))))));
{
switch(s->zFunc){
case cast<uint32_t>(0ULL):{
return false;
break;}
case cast<uint32_t>(1ULL):{
return true;
break;}
case cast<uint32_t>(2ULL):{
return (zi == old);
break;}
case cast<uint32_t>(3ULL):{
return (zi != old);
break;}
case cast<uint32_t>(4ULL):{
return (zi < old);
break;}
case cast<uint32_t>(5ULL):{
return (zi <= old);
break;}
case cast<uint32_t>(6ULL):{
return (zi > old);
break;}
default:{
return (zi >= old);
break;}
}}
}
}
// tools/platform/psp/ge_raster.go:357:1
void psp_Machine_zWrite(psp_Machine* m,psp_geState* s,int64_t x,int64_t y,float z){
{
if (((s->zLow == cast<uint32_t>(0ULL)) || (s->zStride == cast<uint32_t>(0ULL)))) {
return ;
}
if (s->clearOn) {
if (s->clearDepth) {
psp_Machine_write16(m,cast<uint32_t>((psp_geState_zAddress(s) + cast<uint32_t>(((cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(y) * s->zStride)) + cast<uint32_t>(x)))) * cast<uint32_t>(2ULL))))),cast<uint16_t>(psp_clampF(z,cast<float>(0ULL),cast<float>(65535ULL))));
}
return ;
}
if (((!s->zTestOn) || s->zNoWrite)) {
return ;
}
psp_Machine_write16(m,cast<uint32_t>((psp_geState_zAddress(s) + cast<uint32_t>(((cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(y) * s->zStride)) + cast<uint32_t>(x)))) * cast<uint32_t>(2ULL))))),cast<uint16_t>(psp_clampF(z,cast<float>(0ULL),cast<float>(65535ULL))));
}
}
// tools/platform/psp/ge_raster.go:373:1
float psp_clampF(float v,float lo,float hi){
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
// tools/platform/psp/ge_raster.go:383:1
void psp_Machine_rasterList(psp_Machine* m,psp_GeList list){
{rrprof::Scope timing(1,"GE command processor");
{rrGETraceScope traceScope;
{
if ((!m->geSt)) {
psp_geState* s = arenaNew(psp_geState{});
psp_ident((&s->world));
psp_ident((&s->view));
psp_ident((&s->proj));
auto tmp84 = std::make_tuple((cast<int64_t>(480ULL) / cast<int64_t>(2ULL)),(cast<int64_t>(-272ULL) / cast<int64_t>(2ULL)));
s->vpXS = std::get<0>(tmp84);
s->vpYS = std::get<1>(tmp84);
auto tmp85 = std::make_tuple((cast<int64_t>(2048ULL) + cast<int64_t>(240ULL)),(cast<int64_t>(2048ULL) + cast<int64_t>(136ULL)));
s->vpXC = std::get<0>(tmp85);
s->vpYC = std::get<1>(tmp85);
auto tmp86 = std::make_tuple(cast<float>(2048ULL),cast<float>(2048ULL));
s->offX = std::get<0>(tmp86);
s->offY = std::get<1>(tmp86);
auto tmp87 = std::make_tuple(cast<float>(1ULL),cast<float>(1ULL));
s->texScaleU = std::get<0>(tmp87);
s->texScaleV = std::get<1>(tmp87);
s->matColor = cast<uint32_t>(4294967295ULL);
s->stencilOn = true;
s->stFunc = cast<uint32_t>(1ULL);
s->stMask = cast<uint32_t>(255ULL);
s->zFunc = cast<uint32_t>(7ULL);
auto tmp88 = std::make_tuple(cast<int64_t>(479ULL),cast<int64_t>(271ULL));
s->scX1 = std::get<0>(tmp88);
s->scY1 = std::get<1>(tmp88);
m->geSt = s;
}
psp_geState* s = m->geSt;
{auto&& tmp89 = list.Words;
for(int64_t tmp90=0;tmp90<len(tmp89);++tmp90){
auto w=tmp89[tmp90];uint32_t cmd = shr<uint32_t>(w,cast<int64_t>(24ULL));
uint32_t arg = cast<uint32_t>((w & cast<uint32_t>(16777215ULL)));
if (((m->geLimit > cast<int64_t>(0ULL)) && (m->geCount >= m->geLimit))) {
m->StopRequested = true;
return ;
}
if (bool(m->OnGeCmd)) {
m->OnGeCmd(w);
}
m->geCount++;
if ((psp_gePrimDump >= cast<int64_t>(0ULL))) {
psp_geWordTrail = append(psp_geWordTrail,w);
if ((len(psp_geWordTrail) > cast<int64_t>(400ULL))) {
psp_geWordTrail = sub(psp_geWordTrail,cast<int64_t>((len(psp_geWordTrail) - cast<int64_t>(400ULL))),len(psp_geWordTrail));
}
}
{
switch(cmd){
case cast<uint32_t>(16ULL):{
s->base = psp_geBaseAddr(arg);
break;}
case cast<uint32_t>(1ULL):{
s->vaddr = cast<uint32_t>((s->base | arg));
break;}
case cast<uint32_t>(2ULL):{
s->iaddr = cast<uint32_t>((s->base | arg));
break;}
case cast<uint32_t>(19ULL):{
s->offAddr = shl<uint32_t>(arg,cast<int64_t>(8ULL));
break;}
case cast<uint32_t>(23ULL):{
s->lightOn = (cast<uint32_t>((arg & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL));
break;}
case cast<uint32_t>(92ULL):{
s->ambientCol = arg;
break;}
case cast<uint32_t>(84ULL):{
s->matEmissive = arg;
break;}
case cast<uint32_t>(24ULL):case cast<uint32_t>(25ULL):case cast<uint32_t>(26ULL):case cast<uint32_t>(27ULL):{
s->lightEnable[cast<uint32_t>((cmd - cast<uint32_t>(24ULL)))] = (cast<uint32_t>((arg & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL));
break;}
case cast<uint32_t>(95ULL):case cast<uint32_t>(96ULL):case cast<uint32_t>(97ULL):case cast<uint32_t>(98ULL):{
s->lightType[cast<uint32_t>((cmd - cast<uint32_t>(95ULL)))] = cast<uint32_t>(((shr<uint32_t>(arg,cast<int64_t>(8ULL))) & cast<uint32_t>(3ULL)));
break;}
case cast<uint32_t>(99ULL):case cast<uint32_t>(100ULL):case cast<uint32_t>(101ULL):case cast<uint32_t>(102ULL):case cast<uint32_t>(103ULL):case cast<uint32_t>(104ULL):case cast<uint32_t>(105ULL):case cast<uint32_t>(106ULL):case cast<uint32_t>(107ULL):case cast<uint32_t>(108ULL):case cast<uint32_t>(109ULL):case cast<uint32_t>(110ULL):{
uint32_t i = divi<uint32_t>((cast<uint32_t>((cmd - cast<uint32_t>(99ULL)))),cast<uint32_t>(3ULL));
s->lightPos[i][modi<uint32_t>((cast<uint32_t>((cmd - cast<uint32_t>(99ULL)))),cast<uint32_t>(3ULL))] = go_math_Float32frombits(shl<uint32_t>(arg,cast<int64_t>(8ULL)));
break;}
case cast<uint32_t>(144ULL):case cast<uint32_t>(147ULL):case cast<uint32_t>(150ULL):case cast<uint32_t>(153ULL):{
s->lightDiff[divi<uint32_t>((cast<uint32_t>((cmd - cast<uint32_t>(144ULL)))),cast<uint32_t>(3ULL))] = arg;
break;}
case cast<uint32_t>(54ULL):{
s->patchDivU = cast<uint32_t>((arg & cast<uint32_t>(255ULL)));
s->patchDivV = cast<uint32_t>(((shr<uint32_t>(arg,cast<int64_t>(8ULL))) & cast<uint32_t>(255ULL)));
break;}
case cast<uint32_t>(86ULL):{
s->matDiffuse = arg;
break;}
case cast<uint32_t>(18ULL):{
s->vtype = arg;
break;}
case cast<uint32_t>(156ULL):{
s->fbLow = arg;
break;}
case cast<uint32_t>(157ULL):{
s->fbStride = cast<uint32_t>((arg & cast<uint32_t>(65535ULL)));
s->fbHigh = cast<uint32_t>(((shr<uint32_t>(arg,cast<int64_t>(16ULL))) & cast<uint32_t>(255ULL)));
break;}
case cast<uint32_t>(29ULL):{
s->cullOn = (cast<uint32_t>((arg & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL));
break;}
case cast<uint32_t>(155ULL):{
s->cullFace = cast<uint32_t>((arg & cast<uint32_t>(1ULL)));
break;}
case cast<uint32_t>(158ULL):{
s->zLow = arg;
break;}
case cast<uint32_t>(159ULL):{
s->zStride = cast<uint32_t>((arg & cast<uint32_t>(65535ULL)));
s->zHigh = cast<uint32_t>(((shr<uint32_t>(arg,cast<int64_t>(16ULL))) & cast<uint32_t>(255ULL)));
break;}
case cast<uint32_t>(35ULL):{
s->zTestOn = (cast<uint32_t>((arg & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL));
break;}
case cast<uint32_t>(222ULL):{
s->zFunc = cast<uint32_t>((arg & cast<uint32_t>(7ULL)));
break;}
case cast<uint32_t>(231ULL):{
s->zNoWrite = (cast<uint32_t>((arg & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL));
break;}
case cast<uint32_t>(178ULL):{
s->trSrc = cast<uint32_t>(((cast<uint32_t>((s->trSrc & cast<uint32_t>(4278190080ULL)))) | arg));
break;}
case cast<uint32_t>(179ULL):{
s->trSrcStride = cast<uint32_t>((arg & cast<uint32_t>(65535ULL)));
s->trSrc = cast<uint32_t>(((cast<uint32_t>((s->trSrc & cast<uint32_t>(16777215ULL)))) | (shl<uint32_t>(cast<uint32_t>(((shr<uint32_t>(arg,cast<int64_t>(16ULL))) & cast<uint32_t>(255ULL))),cast<int64_t>(24ULL)))));
break;}
case cast<uint32_t>(180ULL):{
s->trDst = cast<uint32_t>(((cast<uint32_t>((s->trDst & cast<uint32_t>(4278190080ULL)))) | arg));
break;}
case cast<uint32_t>(181ULL):{
s->trDstStride = cast<uint32_t>((arg & cast<uint32_t>(65535ULL)));
s->trDst = cast<uint32_t>(((cast<uint32_t>((s->trDst & cast<uint32_t>(16777215ULL)))) | (shl<uint32_t>(cast<uint32_t>(((shr<uint32_t>(arg,cast<int64_t>(16ULL))) & cast<uint32_t>(255ULL))),cast<int64_t>(24ULL)))));
break;}
case cast<uint32_t>(235ULL):{
auto tmp91 = std::make_tuple(cast<uint32_t>((arg & cast<uint32_t>(1023ULL))),cast<uint32_t>(((shr<uint32_t>(arg,cast<int64_t>(10ULL))) & cast<uint32_t>(1023ULL))));
s->trSrcX = std::get<0>(tmp91);
s->trSrcY = std::get<1>(tmp91);
break;}
case cast<uint32_t>(236ULL):{
auto tmp92 = std::make_tuple(cast<uint32_t>((arg & cast<uint32_t>(1023ULL))),cast<uint32_t>(((shr<uint32_t>(arg,cast<int64_t>(10ULL))) & cast<uint32_t>(1023ULL))));
s->trDstX = std::get<0>(tmp92);
s->trDstY = std::get<1>(tmp92);
break;}
case cast<uint32_t>(238ULL):{
auto tmp93 = std::make_tuple(cast<uint32_t>(((cast<uint32_t>((arg & cast<uint32_t>(1023ULL)))) + cast<uint32_t>(1ULL))),cast<uint32_t>(((cast<uint32_t>(((shr<uint32_t>(arg,cast<int64_t>(10ULL))) & cast<uint32_t>(1023ULL)))) + cast<uint32_t>(1ULL))));
s->trW = std::get<0>(tmp93);
s->trH = std::get<1>(tmp93);
break;}
case cast<uint32_t>(234ULL):{
psp_Machine_blockTransfer(m,s,(cast<uint32_t>((arg & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL)));
break;}
case cast<uint32_t>(31ULL):{
s->fogOn = (cast<uint32_t>((arg & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL));
break;}
case cast<uint32_t>(205ULL):{
s->fogEnd = go_math_Float32frombits(shl<uint32_t>(arg,cast<int64_t>(8ULL)));
break;}
case cast<uint32_t>(206ULL):{
s->fogScale = go_math_Float32frombits(shl<uint32_t>(arg,cast<int64_t>(8ULL)));
break;}
case cast<uint32_t>(207ULL):{
s->fogColor = arg;
break;}
case cast<uint32_t>(36ULL):{
s->stencilOn = (cast<uint32_t>((arg & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL));
break;}
case cast<uint32_t>(220ULL):{
s->stFunc = cast<uint32_t>((arg & cast<uint32_t>(7ULL)));
s->stRef = cast<uint32_t>(((shr<uint32_t>(arg,cast<int64_t>(8ULL))) & cast<uint32_t>(255ULL)));
s->stMask = cast<uint32_t>(((shr<uint32_t>(arg,cast<int64_t>(16ULL))) & cast<uint32_t>(255ULL)));
break;}
case cast<uint32_t>(221ULL):{
s->stSFail = cast<uint32_t>((arg & cast<uint32_t>(255ULL)));
s->stZFail = cast<uint32_t>(((shr<uint32_t>(arg,cast<int64_t>(8ULL))) & cast<uint32_t>(255ULL)));
s->stZPass = cast<uint32_t>(((shr<uint32_t>(arg,cast<int64_t>(16ULL))) & cast<uint32_t>(255ULL)));
break;}
case cast<uint32_t>(34ULL):{
s->alphaTestOn = (cast<uint32_t>((arg & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL));
break;}
case cast<uint32_t>(219ULL):{
s->alphaFunc = cast<uint32_t>((arg & cast<uint32_t>(7ULL)));
s->alphaRef = cast<uint32_t>(((shr<uint32_t>(arg,cast<int64_t>(8ULL))) & cast<uint32_t>(255ULL)));
s->alphaTestMask = cast<uint32_t>(((shr<uint32_t>(arg,cast<int64_t>(16ULL))) & cast<uint32_t>(255ULL)));
break;}
case cast<uint32_t>(223ULL):{
s->blendSrc = cast<uint32_t>((arg & cast<uint32_t>(15ULL)));
s->blendDst = cast<uint32_t>(((shr<uint32_t>(arg,cast<int64_t>(4ULL))) & cast<uint32_t>(15ULL)));
s->blendEq = cast<uint32_t>(((shr<uint32_t>(arg,cast<int64_t>(8ULL))) & cast<uint32_t>(7ULL)));
break;}
case cast<uint32_t>(224ULL):{
s->blendFixA = arg;
break;}
case cast<uint32_t>(225ULL):{
s->blendFixB = arg;
break;}
case cast<uint32_t>(212ULL):{
auto tmp94 = std::make_tuple(cast<int64_t>(cast<uint32_t>((arg & cast<uint32_t>(1023ULL)))),cast<int64_t>(cast<uint32_t>(((shr<uint32_t>(arg,cast<int64_t>(10ULL))) & cast<uint32_t>(1023ULL)))));
s->scX0 = std::get<0>(tmp94);
s->scY0 = std::get<1>(tmp94);
break;}
case cast<uint32_t>(213ULL):{
auto tmp95 = std::make_tuple(cast<int64_t>(cast<uint32_t>((arg & cast<uint32_t>(1023ULL)))),cast<int64_t>(cast<uint32_t>(((shr<uint32_t>(arg,cast<int64_t>(10ULL))) & cast<uint32_t>(1023ULL)))));
s->scX1 = std::get<0>(tmp95);
s->scY1 = std::get<1>(tmp95);
break;}
case cast<uint32_t>(192ULL):{
s->texMapMode = cast<uint32_t>((arg & cast<uint32_t>(3ULL)));
s->texProjSrc = cast<uint32_t>(((shr<uint32_t>(arg,cast<int64_t>(8ULL))) & cast<uint32_t>(3ULL)));
break;}
case cast<uint32_t>(199ULL):{
auto tmp96 = std::make_tuple(cast<uint32_t>((arg & cast<uint32_t>(255ULL))),cast<uint32_t>(((shr<uint32_t>(arg,cast<int64_t>(8ULL))) & cast<uint32_t>(255ULL))));
s->texWrapU = std::get<0>(tmp96);
s->texWrapV = std::get<1>(tmp96);
break;}
case cast<uint32_t>(198ULL):{
s->texLinear = (cast<uint32_t>(((shr<uint32_t>(arg,cast<int64_t>(8ULL))) & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL));
break;}
case cast<uint32_t>(201ULL):{
s->texFunc = cast<uint32_t>((arg & cast<uint32_t>(7ULL)));
s->texUseA = (cast<uint32_t>(((shr<uint32_t>(arg,cast<int64_t>(8ULL))) & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL));
s->texDouble = (cast<uint32_t>(((shr<uint32_t>(arg,cast<int64_t>(16ULL))) & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL));
break;}
case cast<uint32_t>(202ULL):{
s->texEnvCol = arg;
break;}
case cast<uint32_t>(64ULL):{
s->texMtxIdx = cast<int64_t>(cast<uint32_t>((arg & cast<uint32_t>(15ULL))));
break;}
case cast<uint32_t>(65ULL):{
psp_matPush((&s->texMtx),(&s->texMtxIdx),arg);
break;}
case cast<uint32_t>(210ULL):{
s->fbFmt = cast<uint32_t>((arg & cast<uint32_t>(3ULL)));
break;}
case cast<uint32_t>(211ULL):{
s->clearOn = (cast<uint32_t>((arg & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL));
s->clearDepth = (cast<uint32_t>((arg & cast<uint32_t>(1024ULL))) != cast<uint32_t>(0ULL));
break;}
case cast<uint32_t>(85ULL):{
s->matColor = cast<uint32_t>(((cast<uint32_t>((s->matColor & cast<uint32_t>(4278190080ULL)))) | arg));
break;}
case cast<uint32_t>(88ULL):{
s->matColor = cast<uint32_t>(((cast<uint32_t>((s->matColor & cast<uint32_t>(16777215ULL)))) | shl<uint32_t>((cast<uint32_t>((arg & cast<uint32_t>(255ULL)))),cast<int64_t>(24ULL))));
break;}
case cast<uint32_t>(33ULL):{
s->blendOn = (cast<uint32_t>((arg & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL));
break;}
case cast<uint32_t>(66ULL):{
s->vpXS = psp_f24(arg);
break;}
case cast<uint32_t>(67ULL):{
s->vpYS = psp_f24(arg);
break;}
case cast<uint32_t>(68ULL):{
s->vpZS = psp_f24(arg);
break;}
case cast<uint32_t>(69ULL):{
s->vpXC = psp_f24(arg);
break;}
case cast<uint32_t>(70ULL):{
s->vpYC = psp_f24(arg);
break;}
case cast<uint32_t>(71ULL):{
s->vpZC = psp_f24(arg);
break;}
case cast<uint32_t>(76ULL):{
s->offX = (cast<float>(arg) / cast<float>(16ULL));
break;}
case cast<uint32_t>(77ULL):{
s->offY = (cast<float>(arg) / cast<float>(16ULL));
break;}
case cast<uint32_t>(58ULL):{
s->worldIdx = cast<int64_t>(cast<uint32_t>((arg & cast<uint32_t>(15ULL))));
break;}
case cast<uint32_t>(59ULL):{
psp_matPush((&s->world),(&s->worldIdx),arg);
break;}
case cast<uint32_t>(60ULL):{
s->viewIdx = cast<int64_t>(cast<uint32_t>((arg & cast<uint32_t>(15ULL))));
break;}
case cast<uint32_t>(61ULL):{
psp_matPush((&s->view),(&s->viewIdx),arg);
break;}
case cast<uint32_t>(62ULL):{
s->projIdx = cast<int64_t>(cast<uint32_t>((arg & cast<uint32_t>(31ULL))));
break;}
case cast<uint32_t>(63ULL):{
psp_matPush16((&s->proj),(&s->projIdx),arg);
break;}
case cast<uint32_t>(30ULL):{
s->texEnable = (cast<uint32_t>((arg & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL));
break;}
case cast<uint32_t>(160ULL):case cast<uint32_t>(161ULL):case cast<uint32_t>(162ULL):case cast<uint32_t>(163ULL):case cast<uint32_t>(164ULL):case cast<uint32_t>(165ULL):case cast<uint32_t>(166ULL):case cast<uint32_t>(167ULL):{
uint32_t l = cast<uint32_t>((cmd - cast<uint32_t>(160ULL)));
s->texAddrN[l] = cast<uint32_t>(((cast<uint32_t>((s->texAddrN[l] & cast<uint32_t>(4278190080ULL)))) | (cast<uint32_t>((arg & cast<uint32_t>(16777215ULL))))));
if ((l == cast<uint32_t>(0ULL))) {
s->texAddr = s->texAddrN[cast<int64_t>(0ULL)];
}
break;}
case cast<uint32_t>(168ULL):case cast<uint32_t>(169ULL):case cast<uint32_t>(170ULL):case cast<uint32_t>(171ULL):case cast<uint32_t>(172ULL):case cast<uint32_t>(173ULL):case cast<uint32_t>(174ULL):case cast<uint32_t>(175ULL):{
uint32_t l = cast<uint32_t>((cmd - cast<uint32_t>(168ULL)));
s->texStrideN[l] = cast<uint32_t>((arg & cast<uint32_t>(65535ULL)));
s->texAddrN[l] = cast<uint32_t>(((cast<uint32_t>((s->texAddrN[l] & cast<uint32_t>(16777215ULL)))) | (shl<uint32_t>((cast<uint32_t>((arg & cast<uint32_t>(16711680ULL)))),cast<int64_t>(8ULL)))));
if ((l == cast<uint32_t>(0ULL))) {
auto tmp97 = std::make_tuple(s->texStrideN[cast<int64_t>(0ULL)],s->texAddrN[cast<int64_t>(0ULL)]);
s->texStride = std::get<0>(tmp97);
s->texAddr = std::get<1>(tmp97);
}
break;}
case cast<uint32_t>(184ULL):case cast<uint32_t>(185ULL):case cast<uint32_t>(186ULL):case cast<uint32_t>(187ULL):case cast<uint32_t>(188ULL):case cast<uint32_t>(189ULL):case cast<uint32_t>(190ULL):case cast<uint32_t>(191ULL):{
uint32_t l = cast<uint32_t>((cmd - cast<uint32_t>(184ULL)));
s->texWN[l] = shl<uint32_t>(cast<uint32_t>(1ULL),(cast<uint32_t>((arg & cast<uint32_t>(255ULL)))));
s->texHN[l] = shl<uint32_t>(cast<uint32_t>(1ULL),(cast<uint32_t>(((shr<uint32_t>(arg,cast<int64_t>(8ULL))) & cast<uint32_t>(255ULL)))));
if ((l == cast<uint32_t>(0ULL))) {
auto tmp98 = std::make_tuple(s->texWN[cast<int64_t>(0ULL)],s->texHN[cast<int64_t>(0ULL)]);
s->texW = std::get<0>(tmp98);
s->texH = std::get<1>(tmp98);
}
break;}
case cast<uint32_t>(200ULL):{
s->texLodMode = cast<uint32_t>((arg & cast<uint32_t>(3ULL)));
s->texLodBias = (cast<float>(cast<int8_t>(cast<uint8_t>(shr<uint32_t>(arg,cast<int64_t>(16ULL))))) / cast<float>(16ULL));
if (psp_geNoBias) {
s->texLodBias = cast<float>(0ULL);
}
break;}
case cast<uint32_t>(195ULL):{
s->texFmt = cast<uint32_t>((arg & cast<uint32_t>(15ULL)));
break;}
case cast<uint32_t>(194ULL):{
s->texSwizzle = (cast<uint32_t>((arg & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL));
s->texMaxLvl = cast<uint32_t>(((shr<uint32_t>(arg,cast<int64_t>(16ULL))) & cast<uint32_t>(7ULL)));
break;}
case cast<uint32_t>(72ULL):{
s->texScaleU = psp_f24(arg);
break;}
case cast<uint32_t>(73ULL):{
s->texScaleV = psp_f24(arg);
break;}
case cast<uint32_t>(74ULL):{
s->texOffU = psp_f24(arg);
break;}
case cast<uint32_t>(75ULL):{
s->texOffV = psp_f24(arg);
break;}
case cast<uint32_t>(176ULL):{
s->clutAddr = cast<uint32_t>(((cast<uint32_t>((s->clutAddr & cast<uint32_t>(4278190080ULL)))) | (cast<uint32_t>((arg & cast<uint32_t>(16777215ULL))))));
break;}
case cast<uint32_t>(177ULL):{
s->clutAddr = cast<uint32_t>(((cast<uint32_t>((s->clutAddr & cast<uint32_t>(16777215ULL)))) | (shl<uint32_t>((cast<uint32_t>((arg & cast<uint32_t>(16711680ULL)))),cast<int64_t>(8ULL)))));
break;}
case cast<uint32_t>(197ULL):{
s->clutFmt = arg;
break;}
case cast<uint32_t>(196ULL):{
psp_Machine_loadClut(m,s,cast<uint32_t>((arg & cast<uint32_t>(63ULL))));
break;}
case cast<uint32_t>(232ULL):{
s->maskRGB = arg;
break;}
case cast<uint32_t>(233ULL):{
s->maskA = cast<uint32_t>((arg & cast<uint32_t>(255ULL)));
break;}
case cast<uint32_t>(4ULL):{
psp_Machine_drawPrim(m,s,arg);
break;}
case cast<uint32_t>(5ULL):case cast<uint32_t>(6ULL):{
time_Time tp = psp_Machine_profStart(m);
psp_Machine_drawPatch(m,s,arg,(cmd == cast<uint32_t>(6ULL)));
psp_Machine_profEnd(m,cast<int64_t>(2ULL),tp);
break;}
}}
}}
if ((s->fbStride != cast<uint32_t>(0ULL))) {
m->fbAddr = psp_geState_fbAddress(s);
m->fbWidth = s->fbStride;
m->fbFormat = s->fbFmt;
}
}
}
}
}
// tools/platform/psp/ge_raster.go:685:1
uint32_t psp_geState_fbAddress(psp_geState* s){
{
return cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(67108864ULL) | (shl<uint32_t>(s->fbHigh,cast<int64_t>(24ULL))))) | s->fbLow));
}
}
// tools/platform/psp/ge_raster.go:690:1
void psp_Machine_loadClut(psp_Machine* m,psp_geState* s,uint32_t blocks){
{
if ((s->clutAddr == cast<uint32_t>(0ULL))) {
return ;
}
uint32_t entryFmt = cast<uint32_t>((s->clutFmt & cast<uint32_t>(3ULL)));
uint32_t n = cast<uint32_t>((blocks * cast<uint32_t>(8ULL)));
if ((entryFmt != cast<uint32_t>(3ULL))) {
n = cast<uint32_t>((blocks * cast<uint32_t>(16ULL)));
}
if ((n > cast<uint32_t>(256ULL))) {
n = cast<uint32_t>(256ULL);
}
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < n);i++){
if ((entryFmt == cast<uint32_t>(3ULL))) {
s->clut[i] = psp_Machine_read32(m,cast<uint32_t>((s->clutAddr + cast<uint32_t>((i * cast<uint32_t>(4ULL))))));
}
else {
auto tmp99 = psp_decode16a(psp_u16(m,cast<uint32_t>((s->clutAddr + cast<uint32_t>((i * cast<uint32_t>(2ULL)))))),entryFmt);
uint8_t r = std::get<0>(tmp99);
uint8_t g = std::get<1>(tmp99);
uint8_t b = std::get<2>(tmp99);
uint8_t a = std::get<3>(tmp99);
s->clut[i] = cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(r) | shl<uint32_t>(cast<uint32_t>(g),cast<int64_t>(8ULL)))) | shl<uint32_t>(cast<uint32_t>(b),cast<int64_t>(16ULL)))) | shl<uint32_t>(cast<uint32_t>(a),cast<int64_t>(24ULL))));
}
}
}}
}
// tools/platform/psp/ge_raster.go:722:1
void psp_Machine_dumpPrim(psp_Machine* m,psp_geState* s,uint32_t ptype,int64_t count){
{
go_fmt_Printf(std::string("=== PRIM#%d t%d n%d ===\012",24),psp_gePrimSeq,ptype,count);
go_fmt_Printf(std::string("state: vt=%06X va=%08X clear=%v blend=%v texEn=%v tex@%08X f%d sw=%v %dx%d stride=%d clut@%08X clutFmt=%06X mat=%08X fb=%08X/%d fmt=%d\012",135),s->vtype,s->vaddr,s->clearOn,s->blendOn,s->texEnable,s->texAddr,s->texFmt,s->texSwizzle,s->texW,s->texH,s->texStride,s->clutAddr,s->clutFmt,s->matColor,psp_geState_fbAddress(s),s->fbStride,s->fbFmt);
go_fmt_Printf(std::string("viewport: S(%.1f,%.1f,%.1f) C(%.1f,%.1f,%.1f) off(%.1f,%.1f)\012",61),s->vpXS,s->vpYS,s->vpZS,s->vpXC,s->vpYC,s->vpZC,s->offX,s->offY);
{auto&& tmp100 = Slice<Anon13>{Anon13{std::string("world",5),s->world},Anon13{std::string("view",4),s->view},Anon13{std::string("proj",4),s->proj}};
for(int64_t tmp101=0;tmp101<len(tmp100);++tmp101){
auto m=tmp100[tmp101];go_fmt_Printf(std::string("%s:",3),m.name);
{auto&& tmp102 = m.v;
for(int64_t tmp103=0;tmp103<len(tmp102);++tmp103){
auto f=tmp102[tmp103];go_fmt_Printf(std::string(" %g",3),f);
}}
go_fmt_Println();
}}
go_fmt_Printf(std::string("trail (%d words since previous PRIM):\012",38),len(psp_geWordTrail));
{auto&& tmp104 = psp_geWordTrail;
for(int64_t tmp105=0;tmp105<len(tmp104);++tmp105){
auto w=tmp104[tmp105];go_fmt_Printf(std::string("  %08X  %s %06X\012",16),w,psp_GeCmdName(shr<uint32_t>(w,cast<int64_t>(24ULL))),cast<uint32_t>((w & cast<uint32_t>(16777215ULL))));
}}
{auto&& tmp106 = psp_Machine_decodeVerts(m,s,count);
for(int64_t tmp107=0;tmp107<len(tmp106);++tmp107){
auto i=tmp107;auto v=tmp106[tmp107];go_fmt_Printf(std::string("  v%-3d (%8.2f,%8.2f,%8.2f) uv(%.3f,%.3f) rgba %02X%02X%02X%02X\012",64),i,v.x,v.y,v.z,v.u,v.v,v.r,v.g,v.b,v.a);
}}
}
}
// tools/platform/psp/ge_raster.go:754:1
void psp_Machine_drawPrim(psp_Machine* m,psp_geState* s,uint32_t arg){
{rrprof::Scope timing(2,"GE vertices and software rasterizer");
{rrTextureScope textures(m,s);
{
int64_t count = cast<int64_t>(cast<uint32_t>((arg & cast<uint32_t>(65535ULL))));
uint32_t ptype = cast<uint32_t>(((shr<uint32_t>(arg,cast<int64_t>(16ULL))) & cast<uint32_t>(7ULL)));
if (((count == cast<int64_t>(0ULL)) || (s->vaddr == cast<uint32_t>(0ULL)))) {
return ;
}
psp_gePrimSeq++;
if ((psp_gePrimSeq == psp_gePrimDump)) {
psp_Machine_dumpPrim(m,s,ptype,count);
}
if ((psp_gePrimDump >= cast<int64_t>(0ULL))) {
psp_geWordTrail = sub(psp_geWordTrail,0,cast<int64_t>(0ULL));
}
if ((psp_geWorldDump > cast<int64_t>(0ULL))) {
psp_geWorldDump--;
go_fmt_Printf(std::string("WORLD va=%08X vt=%06X",21),s->vaddr,s->vtype);
{auto&& tmp108 = s->world;
for(int64_t tmp109=0;tmp109<len(tmp108);++tmp109){
auto f=tmp108[tmp109];go_fmt_Printf(std::string(" %g",3),f);
}}
go_fmt_Println();
}
if ((psp_geDebugN > cast<int64_t>(0ULL))) {
psp_geDebugN--;
Slice<psp_vert> vs = psp_Machine_decodeVerts(m,s,gmin(count,cast<int64_t>(2ULL)));
psp_vert v0={};
if ((len(vs) > cast<int64_t>(0ULL))) {
v0 = vs[cast<int64_t>(0ULL)];
}
go_fmt_Printf(std::string("PRIM#%d t%d n%d vt=%06X va=%08X tex=%v@%08X f%d sw%v %dx%d clut@%08X mat=%08X wT=(%.1f,%.1f,%.1f) v0=(%.1f,%.1f,%.1f uv %.2f,%.2f)\012",131),psp_gePrimSeq,ptype,count,s->vtype,s->vaddr,s->texEnable,s->texAddr,s->texFmt,s->texSwizzle,s->texW,s->texH,s->clutAddr,s->matColor,s->world[cast<int64_t>(12ULL)],s->world[cast<int64_t>(13ULL)],s->world[cast<int64_t>(14ULL)],v0.x,v0.y,v0.z,v0.u,v0.v);
}
if (((psp_geOnlyTex != cast<uint64_t>(0ULL)) && (cast<uint64_t>(s->texAddr) != psp_geOnlyTex))) {
return ;
}
if (((psp_geOnlyVA != cast<uint64_t>(0ULL)) && (shr<uint64_t>(cast<uint64_t>(s->vaddr),cast<int64_t>(16ULL)) != shr<uint64_t>(psp_geOnlyVA,cast<int64_t>(16ULL))))) {
return ;
}
m->geCnt.prims++;
m->geCnt.verts += count;
time_Time tv = psp_Machine_profStart(m);
Slice<psp_vert> verts = psp_Machine_decodeVerts(m,s,count);
psp_Machine_profEnd(m,cast<int64_t>(1ULL),tv);
{
psp_vertexLayout l = psp_layoutOf(s->vtype);
if ((l.stride != cast<uint32_t>(0ULL))) {
if ((l.idxFmt != cast<uint32_t>(0ULL))) {
s->iaddr += cast<uint32_t>((cast<uint32_t>(count) * l.idxFmt));
}
else {
s->vaddr += cast<uint32_t>((cast<uint32_t>(count) * l.stride));
}
}
}
if ((len(verts) < cast<int64_t>(1ULL))) {
return ;
}
time_Time tr = psp_Machine_profStart(m);
{
switch(ptype){
case cast<uint32_t>(3ULL):{
{int64_t i = cast<int64_t>(0ULL);for (;(cast<int64_t>((i + cast<int64_t>(2ULL))) < len(verts));i += cast<int64_t>(3ULL)){
psp_Machine_rasterTriClipped(m,s,verts[i],verts[cast<int64_t>((i + cast<int64_t>(1ULL)))],verts[cast<int64_t>((i + cast<int64_t>(2ULL)))]);
}
}break;}
case cast<uint32_t>(4ULL):{
{int64_t i = cast<int64_t>(0ULL);for (;(cast<int64_t>((i + cast<int64_t>(2ULL))) < len(verts));i++){
auto tmp110 = std::make_tuple(verts[i],verts[cast<int64_t>((i + cast<int64_t>(1ULL)))],verts[cast<int64_t>((i + cast<int64_t>(2ULL)))]);
psp_vert a = std::get<0>(tmp110);
psp_vert b = std::get<1>(tmp110);
psp_vert c = std::get<2>(tmp110);
if ((cast<int64_t>((i & cast<int64_t>(1ULL))) == cast<int64_t>(1ULL))) {
auto tmp111 = std::make_tuple(c,b);
b = std::get<0>(tmp111);
c = std::get<1>(tmp111);
}
psp_Machine_rasterTriClipped(m,s,a,b,c);
}
}break;}
case cast<uint32_t>(5ULL):{
{int64_t i = cast<int64_t>(1ULL);for (;(cast<int64_t>((i + cast<int64_t>(1ULL))) < len(verts));i++){
psp_Machine_rasterTriClipped(m,s,verts[cast<int64_t>(0ULL)],verts[i],verts[cast<int64_t>((i + cast<int64_t>(1ULL)))]);
}
}break;}
case cast<uint32_t>(6ULL):{
{int64_t i = cast<int64_t>(0ULL);for (;(cast<int64_t>((i + cast<int64_t>(1ULL))) < len(verts));i += cast<int64_t>(2ULL)){
psp_Machine_rasterSprite(m,s,verts[i],verts[cast<int64_t>((i + cast<int64_t>(1ULL)))]);
}
}break;}
}}
psp_Machine_profEnd(m,cast<int64_t>(2ULL),tr);
}
}
}
}
// tools/platform/psp/ge_raster.go:861:1
bool psp_vert_through(psp_vert v){
{
return (!v.clip);
}
}
// tools/platform/psp/ge_raster.go:877:1
psp_vertexLayout psp_layoutOf(uint32_t vtype){
{
psp_vertexLayout l={};
l.tfmt = cast<uint32_t>((vtype & cast<uint32_t>(3ULL)));
l.cfmt = cast<uint32_t>(((shr<uint32_t>(vtype,cast<int64_t>(2ULL))) & cast<uint32_t>(7ULL)));
l.nfmt = cast<uint32_t>(((shr<uint32_t>(vtype,cast<int64_t>(5ULL))) & cast<uint32_t>(3ULL)));
l.pfmt = cast<uint32_t>(((shr<uint32_t>(vtype,cast<int64_t>(7ULL))) & cast<uint32_t>(3ULL)));
uint32_t wfmt = cast<uint32_t>(((shr<uint32_t>(vtype,cast<int64_t>(9ULL))) & cast<uint32_t>(3ULL)));
l.idxFmt = cast<uint32_t>(((shr<uint32_t>(vtype,cast<int64_t>(11ULL))) & cast<uint32_t>(3ULL)));
uint32_t wcount = cast<uint32_t>(((cast<uint32_t>(((shr<uint32_t>(vtype,cast<int64_t>(14ULL))) & cast<uint32_t>(7ULL)))) + cast<uint32_t>(1ULL)));
l.through = (cast<uint32_t>((vtype & cast<uint32_t>(8388608ULL))) != cast<uint32_t>(0ULL));
Slice<uint32_t> elem = Slice<uint32_t>{cast<uint32_t>(0ULL),cast<uint32_t>(1ULL),cast<uint32_t>(2ULL),cast<uint32_t>(4ULL)};
uint32_t texSz = cast<uint32_t>((elem[l.tfmt] * cast<uint32_t>(2ULL)));
uint32_t colSz = get(Map<uint32_t,uint32_t>{{cast<uint32_t>(0ULL),cast<uint32_t>(0ULL)},{cast<uint32_t>(4ULL),cast<uint32_t>(2ULL)},{cast<uint32_t>(5ULL),cast<uint32_t>(2ULL)},{cast<uint32_t>(6ULL),cast<uint32_t>(2ULL)},{cast<uint32_t>(7ULL),cast<uint32_t>(4ULL)}},l.cfmt);
uint32_t nrmSz = cast<uint32_t>((elem[l.nfmt] * cast<uint32_t>(3ULL)));
uint32_t posSz = cast<uint32_t>((elem[l.pfmt] * cast<uint32_t>(3ULL)));
uint32_t wgtSz = cast<uint32_t>((elem[wfmt] * wcount));
if ((wfmt == cast<uint32_t>(0ULL))) {
wgtSz = cast<uint32_t>(0ULL);
}
uint32_t align = cast<uint32_t>(1ULL);
{auto&& tmp112 = Slice<uint32_t>{elem[l.tfmt],elem[l.nfmt],elem[l.pfmt],elem[wfmt],colSz};
for(int64_t tmp113=0;tmp113<len(tmp112);++tmp113){
auto e=tmp112[tmp113];if ((e > align)) {
align = e;
}
}}
auto pad = [&](uint32_t off,uint32_t a)->uint32_t{
if ((a == cast<uint32_t>(0ULL))) {
return off;
}
return ((cast<uint32_t>((cast<uint32_t>((off + a)) - cast<uint32_t>(1ULL)))) & ~((cast<uint32_t>((a - cast<uint32_t>(1ULL))))));
}
;
uint32_t off = cast<uint32_t>((pad(cast<uint32_t>(0ULL),elem[wfmt]) + wgtSz));
l.offTex = pad(off,elem[l.tfmt]);
off = cast<uint32_t>((l.offTex + texSz));
l.offCol = pad(off,colSz);
off = cast<uint32_t>((l.offCol + colSz));
l.offNrm = pad(off,elem[l.nfmt]);
off = cast<uint32_t>((l.offNrm + nrmSz));
l.offPos = pad(off,elem[l.pfmt]);
off = cast<uint32_t>((l.offPos + posSz));
l.stride = pad(off,align);
return l;
}
}
// tools/platform/psp/ge_raster.go:923:1
Slice<psp_vert> psp_Machine_decodeVerts(psp_Machine* m,psp_geState* s,int64_t count){
{
psp_vertexLayout l = psp_layoutOf(s->vtype);
auto tmp114 = std::make_tuple(l.tfmt,l.cfmt,l.nfmt,l.pfmt,l.idxFmt);
uint32_t tfmt = std::get<0>(tmp114);
uint32_t cfmt = std::get<1>(tmp114);
uint32_t nfmt = std::get<2>(tmp114);
uint32_t pfmt = std::get<3>(tmp114);
uint32_t idxFmt = std::get<4>(tmp114);
auto tmp115 = std::make_tuple(l.offTex,l.offCol,l.offNrm,l.offPos);
uint32_t offTex = std::get<0>(tmp115);
uint32_t offCol = std::get<1>(tmp115);
uint32_t offNrm = std::get<2>(tmp115);
uint32_t offPos = std::get<3>(tmp115);
auto tmp116 = std::make_tuple(l.stride,l.through);
uint32_t stride = std::get<0>(tmp116);
bool through = std::get<1>(tmp116);
if ((stride == cast<uint32_t>(0ULL))) {
return {};
}
std::array<float,16> mvp = psp_mul4(psp_mul4(s->proj,s->view),s->world);
Slice<psp_vert> out = Slice<psp_vert>::make(cast<int64_t>(0ULL),count);
uint32_t addr = s->vaddr;
{int64_t i = cast<int64_t>(0ULL);for (;(i < count);i++){
uint32_t p = addr;
{
switch(idxFmt){
case cast<uint32_t>(1ULL):{
p = cast<uint32_t>((s->vaddr + cast<uint32_t>((cast<uint32_t>(psp_Machine_Read(m,cast<uint32_t>((s->iaddr + cast<uint32_t>(i))))) * stride))));
break;}
case cast<uint32_t>(2ULL):{
p = cast<uint32_t>((s->vaddr + cast<uint32_t>((cast<uint32_t>(psp_u16(m,cast<uint32_t>((s->iaddr + cast<uint32_t>((cast<uint32_t>(i) * cast<uint32_t>(2ULL))))))) * stride))));
break;}
}}
psp_vert vv={};
vv.a = cast<uint8_t>(255ULL);
if ((tfmt != cast<uint32_t>(0ULL))) {
auto tmp117 = psp_readUV(m,cast<uint32_t>((p + offTex)),tfmt);
vv.u = std::get<0>(tmp117);
vv.v = std::get<1>(tmp117);
if (through) {
if (((s->texW != cast<uint32_t>(0ULL)) && (s->texH != cast<uint32_t>(0ULL)))) {
vv.u /= cast<float>(s->texW);
vv.v /= cast<float>(s->texH);
}
}
else {
{
switch(tfmt){
case cast<uint32_t>(1ULL):{
vv.u /= cast<float>(128ULL);
vv.v /= cast<float>(128ULL);
break;}
case cast<uint32_t>(2ULL):{
vv.u /= cast<float>(32768ULL);
vv.v /= cast<float>(32768ULL);
break;}
}}
vv.u = ((vv.u * s->texScaleU) + s->texOffU);
vv.v = ((vv.v * s->texScaleV) + s->texOffV);
}
}
if (((!through) && (s->texMapMode == cast<uint32_t>(1ULL)))) {
float sx={};
float sy={};
float sz={};
{
switch(s->texProjSrc){
case cast<uint32_t>(1ULL):{
auto tmp118 = std::make_tuple(vv.u,vv.v,cast<float>(0ULL));
sx = std::get<0>(tmp118);
sy = std::get<1>(tmp118);
sz = std::get<2>(tmp118);
break;}
case cast<uint32_t>(2ULL):case cast<uint32_t>(3ULL):{
auto tmp119 = psp_readNormal(m,cast<uint32_t>((p + offNrm)),nfmt);
sx = std::get<0>(tmp119);
sy = std::get<1>(tmp119);
sz = std::get<2>(tmp119);
if ((s->texProjSrc == cast<uint32_t>(2ULL))) {
{
float l = cast<float>(go_math_Sqrt(cast<double>((((sx * sx) + (sy * sy)) + (sz * sz)))));
if ((l > 1e-8)) {
auto tmp120 = std::make_tuple((sx / l),(sy / l),(sz / l));
sx = std::get<0>(tmp120);
sy = std::get<1>(tmp120);
sz = std::get<2>(tmp120);
}
}
}
break;}
default:{
auto tmp121 = psp_readPos(m,cast<uint32_t>((p + offPos)),pfmt);
sx = std::get<0>(tmp121);
sy = std::get<1>(tmp121);
sz = std::get<2>(tmp121);
{
switch(pfmt){
case cast<uint32_t>(1ULL):{
auto tmp122 = std::make_tuple((sx / cast<float>(128ULL)),(sy / cast<float>(128ULL)),(sz / cast<float>(128ULL)));
sx = std::get<0>(tmp122);
sy = std::get<1>(tmp122);
sz = std::get<2>(tmp122);
break;}
case cast<uint32_t>(2ULL):{
auto tmp123 = std::make_tuple((sx / cast<float>(32768ULL)),(sy / cast<float>(32768ULL)),(sz / cast<float>(32768ULL)));
sx = std::get<0>(tmp123);
sy = std::get<1>(tmp123);
sz = std::get<2>(tmp123);
break;}
}}
break;}
}}
float tu = ((((s->texMtx[cast<int64_t>(0ULL)] * sx) + (s->texMtx[cast<int64_t>(4ULL)] * sy)) + (s->texMtx[cast<int64_t>(8ULL)] * sz)) + s->texMtx[cast<int64_t>(12ULL)]);
float tv = ((((s->texMtx[cast<int64_t>(1ULL)] * sx) + (s->texMtx[cast<int64_t>(5ULL)] * sy)) + (s->texMtx[cast<int64_t>(9ULL)] * sz)) + s->texMtx[cast<int64_t>(13ULL)]);
float tq = ((((s->texMtx[cast<int64_t>(2ULL)] * sx) + (s->texMtx[cast<int64_t>(6ULL)] * sy)) + (s->texMtx[cast<int64_t>(10ULL)] * sz)) + s->texMtx[cast<int64_t>(14ULL)]);
if (((tq > 1e-6) || (tq < cast<float>(-1e-6)))) {
auto tmp124 = std::make_tuple((tu / tq),(tv / tq));
tu = std::get<0>(tmp124);
tv = std::get<1>(tmp124);
}
auto tmp125 = std::make_tuple(tu,tv);
vv.u = std::get<0>(tmp125);
vv.v = std::get<1>(tmp125);
}
if ((cfmt != cast<uint32_t>(0ULL))) {
auto tmp126 = psp_readColor(m,cast<uint32_t>((p + offCol)),cfmt);
vv.r = std::get<0>(tmp126);
vv.g = std::get<1>(tmp126);
vv.b = std::get<2>(tmp126);
vv.a = std::get<3>(tmp126);
}
else {
auto tmp127 = std::make_tuple(cast<uint8_t>(s->matColor),cast<uint8_t>(shr<uint32_t>(s->matColor,cast<int64_t>(8ULL))),cast<uint8_t>(shr<uint32_t>(s->matColor,cast<int64_t>(16ULL))));
vv.r = std::get<0>(tmp127);
vv.g = std::get<1>(tmp127);
vv.b = std::get<2>(tmp127);
vv.a = cast<uint8_t>(shr<uint32_t>(s->matColor,cast<int64_t>(24ULL)));
}
if (((s->lightOn && (!through)) && (nfmt != cast<uint32_t>(0ULL)))) {
auto tmp128 = psp_readNormal(m,cast<uint32_t>((p + offNrm)),nfmt);
float nx = std::get<0>(tmp128);
float ny = std::get<1>(tmp128);
float nz = std::get<2>(tmp128);
auto tmp129 = psp_geState_lightVertex(s,nx,ny,nz);
vv.r = std::get<0>(tmp129);
vv.g = std::get<1>(tmp129);
vv.b = std::get<2>(tmp129);
}
auto tmp130 = psp_readPos(m,cast<uint32_t>((p + offPos)),pfmt);
float px = std::get<0>(tmp130);
float py = std::get<1>(tmp130);
float pz = std::get<2>(tmp130);
if (through) {
auto tmp131 = std::make_tuple(px,py,pz);
vv.x = std::get<0>(tmp131);
vv.y = std::get<1>(tmp131);
vv.z = std::get<2>(tmp131);
}
else {
{
switch(pfmt){
case cast<uint32_t>(1ULL):{
auto tmp132 = std::make_tuple((px / cast<float>(128ULL)),(py / cast<float>(128ULL)),(pz / cast<float>(128ULL)));
px = std::get<0>(tmp132);
py = std::get<1>(tmp132);
pz = std::get<2>(tmp132);
break;}
case cast<uint32_t>(2ULL):{
auto tmp133 = std::make_tuple((px / cast<float>(32768ULL)),(py / cast<float>(32768ULL)),(pz / cast<float>(32768ULL)));
px = std::get<0>(tmp133);
py = std::get<1>(tmp133);
pz = std::get<2>(tmp133);
break;}
}}
auto tmp134 = psp_clipCoords(mvp,px,py,pz);
vv.cx = std::get<0>(tmp134);
vv.cy = std::get<1>(tmp134);
vv.cz = std::get<2>(tmp134);
vv.cw = std::get<3>(tmp134);
vv.clip = true;
auto tmp135 = psp_project(s,vv.cx,vv.cy,vv.cz,vv.cw);
vv.x = std::get<0>(tmp135);
vv.y = std::get<1>(tmp135);
vv.z = std::get<2>(tmp135);
vv.invW = psp_invW(vv.cw);
vv.fog = psp_fogCoef(s,vv.cw);
}
out = append(out,vv);
addr += stride;
}
}return out;
}
}
// tools/platform/psp/ge_raster.go:1043:1
std::tuple<float,float> psp_readUV(psp_Machine* m,uint32_t p,uint32_t fmt){
{
{
switch(fmt){
case cast<uint32_t>(1ULL):{
return {cast<float>(psp_Machine_Read(m,p)),cast<float>(psp_Machine_Read(m,cast<uint32_t>((p + cast<uint32_t>(1ULL)))))};
break;}
case cast<uint32_t>(2ULL):{
return {cast<float>(psp_u16(m,p)),cast<float>(psp_u16(m,cast<uint32_t>((p + cast<uint32_t>(2ULL)))))};
break;}
default:{
return {psp_f32(m,p),psp_f32(m,cast<uint32_t>((p + cast<uint32_t>(4ULL))))};
break;}
}}
}
}
// tools/platform/psp/ge_raster.go:1054:1
std::tuple<uint8_t,uint8_t,uint8_t,uint8_t> psp_readColor(psp_Machine* m,uint32_t p,uint32_t fmt){
{
{
switch(fmt){
case cast<uint32_t>(7ULL):{
uint32_t c = psp_Machine_read32(m,p);
return {cast<uint8_t>(c),cast<uint8_t>(shr<uint32_t>(c,cast<int64_t>(8ULL))),cast<uint8_t>(shr<uint32_t>(c,cast<int64_t>(16ULL))),cast<uint8_t>(shr<uint32_t>(c,cast<int64_t>(24ULL)))};
break;}
case cast<uint32_t>(6ULL):{
uint16_t c = psp_u16(m,p);
auto e = [&](uint16_t v)->uint8_t{
return cast<uint8_t>(cast<uint16_t>(((cast<uint16_t>((v & cast<uint16_t>(15ULL)))) * cast<uint16_t>(17ULL))));
}
;
return {e(c),e(shr<uint16_t>(c,cast<int64_t>(4ULL))),e(shr<uint16_t>(c,cast<int64_t>(8ULL))),e(shr<uint16_t>(c,cast<int64_t>(12ULL)))};
break;}
case cast<uint32_t>(5ULL):{
uint16_t c = psp_u16(m,p);
auto e = [&](uint16_t v,uint16_t b)->uint8_t{
return cast<uint8_t>(divi<uint32_t>(cast<uint32_t>((cast<uint32_t>(cast<uint16_t>((v & (cast<uint16_t>(((shl<uint16_t>(cast<uint16_t>(1ULL),b)) - cast<uint16_t>(1ULL))))))) * cast<uint32_t>(255ULL))),(cast<uint32_t>(((shl<uint32_t>(cast<uint32_t>(1ULL),b)) - cast<uint32_t>(1ULL))))));
}
;
uint8_t a = cast<uint8_t>(0ULL);
if ((cast<uint16_t>((c & cast<uint16_t>(32768ULL))) != cast<uint16_t>(0ULL))) {
a = cast<uint8_t>(255ULL);
}
return {e(c,cast<uint16_t>(5ULL)),e(shr<uint16_t>(c,cast<int64_t>(5ULL)),cast<uint16_t>(5ULL)),e(shr<uint16_t>(c,cast<int64_t>(10ULL)),cast<uint16_t>(5ULL)),a};
break;}
default:{
uint16_t c = psp_u16(m,p);
auto e = [&](uint16_t v,uint16_t b)->uint8_t{
return cast<uint8_t>(divi<uint32_t>(cast<uint32_t>((cast<uint32_t>(cast<uint16_t>((v & (cast<uint16_t>(((shl<uint16_t>(cast<uint16_t>(1ULL),b)) - cast<uint16_t>(1ULL))))))) * cast<uint32_t>(255ULL))),(cast<uint32_t>(((shl<uint32_t>(cast<uint32_t>(1ULL),b)) - cast<uint32_t>(1ULL))))));
}
;
return {e(c,cast<uint16_t>(5ULL)),e(shr<uint16_t>(c,cast<int64_t>(5ULL)),cast<uint16_t>(6ULL)),e(shr<uint16_t>(c,cast<int64_t>(11ULL)),cast<uint16_t>(5ULL)),cast<uint8_t>(255ULL)};
break;}
}}
}
}
// tools/platform/psp/ge_raster.go:1082:1
std::tuple<uint8_t,uint8_t,uint8_t> psp_geState_lightVertex(psp_geState* s,float nx,float ny,float nz){
{
float wx = (((s->world[cast<int64_t>(0ULL)] * nx) + (s->world[cast<int64_t>(4ULL)] * ny)) + (s->world[cast<int64_t>(8ULL)] * nz));
float wy = (((s->world[cast<int64_t>(1ULL)] * nx) + (s->world[cast<int64_t>(5ULL)] * ny)) + (s->world[cast<int64_t>(9ULL)] * nz));
float wz = (((s->world[cast<int64_t>(2ULL)] * nx) + (s->world[cast<int64_t>(6ULL)] * ny)) + (s->world[cast<int64_t>(10ULL)] * nz));
{
float l = cast<float>(go_math_Sqrt(cast<double>((((wx * wx) + (wy * wy)) + (wz * wz)))));
if ((l > 1e-8)) {
auto tmp136 = std::make_tuple((wx / l),(wy / l),(wz / l));
wx = std::get<0>(tmp136);
wy = std::get<1>(tmp136);
wz = std::get<2>(tmp136);
}
}
auto chOf = [&](uint32_t c,int64_t i)->float{
return (cast<float>(cast<uint8_t>(shr<uint32_t>(c,(cast<int64_t>((cast<int64_t>(8ULL) * i)))))) / cast<float>(255ULL));
}
;
std::array<float,3> out={};
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(3ULL));i++){
out[i] = (chOf(s->matEmissive,i) + (chOf(s->ambientCol,i) * chOf(s->matColor,i)));
}
}{int64_t li = cast<int64_t>(0ULL);for (;(li < cast<int64_t>(4ULL));li++){
if ((!s->lightEnable[li])) {
continue;
}
auto tmp137 = std::make_tuple(s->lightPos[li][cast<int64_t>(0ULL)],s->lightPos[li][cast<int64_t>(1ULL)],s->lightPos[li][cast<int64_t>(2ULL)]);
float lx = std::get<0>(tmp137);
float ly = std::get<1>(tmp137);
float lz = std::get<2>(tmp137);
{
float l = cast<float>(go_math_Sqrt(cast<double>((((lx * lx) + (ly * ly)) + (lz * lz)))));
if ((l > 1e-8)) {
auto tmp138 = std::make_tuple((lx / l),(ly / l),(lz / l));
lx = std::get<0>(tmp138);
ly = std::get<1>(tmp138);
lz = std::get<2>(tmp138);
}
}
float nd = (((wx * lx) + (wy * ly)) + (wz * lz));
if ((nd <= cast<float>(0ULL))) {
continue;
}
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(3ULL));i++){
out[i] += ((chOf(s->lightDiff[li],i) * chOf(s->matDiffuse,i)) * nd);
}
}}
}return {psp_clamp255(cast<int32_t>((out[cast<int64_t>(0ULL)] * cast<float>(255ULL)))),psp_clamp255(cast<int32_t>((out[cast<int64_t>(1ULL)] * cast<float>(255ULL)))),psp_clamp255(cast<int32_t>((out[cast<int64_t>(2ULL)] * cast<float>(255ULL))))};
}
}
// tools/platform/psp/ge_raster.go:1118:1
std::tuple<float,float,float> psp_readNormal(psp_Machine* m,uint32_t p,uint32_t fmt){
{
{
switch(fmt){
case cast<uint32_t>(1ULL):{
return {(cast<float>(cast<int8_t>(psp_Machine_Read(m,p))) / cast<float>(128ULL)),(cast<float>(cast<int8_t>(psp_Machine_Read(m,cast<uint32_t>((p + cast<uint32_t>(1ULL)))))) / cast<float>(128ULL)),(cast<float>(cast<int8_t>(psp_Machine_Read(m,cast<uint32_t>((p + cast<uint32_t>(2ULL)))))) / cast<float>(128ULL))};
break;}
case cast<uint32_t>(2ULL):{
return {(cast<float>(cast<int16_t>(psp_u16(m,p))) / cast<float>(32768ULL)),(cast<float>(cast<int16_t>(psp_u16(m,cast<uint32_t>((p + cast<uint32_t>(2ULL)))))) / cast<float>(32768ULL)),(cast<float>(cast<int16_t>(psp_u16(m,cast<uint32_t>((p + cast<uint32_t>(4ULL)))))) / cast<float>(32768ULL))};
break;}
case cast<uint32_t>(3ULL):{
return {psp_f32(m,p),psp_f32(m,cast<uint32_t>((p + cast<uint32_t>(4ULL)))),psp_f32(m,cast<uint32_t>((p + cast<uint32_t>(8ULL))))};
break;}
}}
return {cast<float>(0ULL),cast<float>(0ULL),cast<float>(0ULL)};
}
}
// tools/platform/psp/ge_raster.go:1130:1
std::tuple<float,float,float> psp_readPos(psp_Machine* m,uint32_t p,uint32_t fmt){
{
{
switch(fmt){
case cast<uint32_t>(1ULL):{
return {cast<float>(cast<int8_t>(psp_Machine_Read(m,p))),cast<float>(cast<int8_t>(psp_Machine_Read(m,cast<uint32_t>((p + cast<uint32_t>(1ULL)))))),cast<float>(cast<int8_t>(psp_Machine_Read(m,cast<uint32_t>((p + cast<uint32_t>(2ULL))))))};
break;}
case cast<uint32_t>(2ULL):{
return {cast<float>(cast<int16_t>(psp_u16(m,p))),cast<float>(cast<int16_t>(psp_u16(m,cast<uint32_t>((p + cast<uint32_t>(2ULL)))))),cast<float>(cast<int16_t>(psp_u16(m,cast<uint32_t>((p + cast<uint32_t>(4ULL))))))};
break;}
default:{
return {psp_f32(m,p),psp_f32(m,cast<uint32_t>((p + cast<uint32_t>(4ULL)))),psp_f32(m,cast<uint32_t>((p + cast<uint32_t>(8ULL))))};
break;}
}}
}
}
// tools/platform/psp/ge_raster.go:1142:1
std::tuple<float,float,float,float> psp_clipCoords(std::array<float,16> mvp,float x,float y,float z){
{
return {((((mvp[cast<int64_t>(0ULL)] * x) + (mvp[cast<int64_t>(4ULL)] * y)) + (mvp[cast<int64_t>(8ULL)] * z)) + mvp[cast<int64_t>(12ULL)]),((((mvp[cast<int64_t>(1ULL)] * x) + (mvp[cast<int64_t>(5ULL)] * y)) + (mvp[cast<int64_t>(9ULL)] * z)) + mvp[cast<int64_t>(13ULL)]),((((mvp[cast<int64_t>(2ULL)] * x) + (mvp[cast<int64_t>(6ULL)] * y)) + (mvp[cast<int64_t>(10ULL)] * z)) + mvp[cast<int64_t>(14ULL)]),((((mvp[cast<int64_t>(3ULL)] * x) + (mvp[cast<int64_t>(7ULL)] * y)) + (mvp[cast<int64_t>(11ULL)] * z)) + mvp[cast<int64_t>(15ULL)])};
}
}
// tools/platform/psp/ge_raster.go:1150:1
std::tuple<float,float,float> psp_project(psp_geState* s,float cx,float cy,float cz,float cw){
{
if ((cw == cast<float>(0ULL))) {
cw = 1e-6;
}
auto tmp139 = std::make_tuple((cx / cw),(cy / cw),(cz / cw));
float nx = std::get<0>(tmp139);
float ny = std::get<1>(tmp139);
float nz = std::get<2>(tmp139);
float sx = (((nx * s->vpXS) + s->vpXC) - s->offX);
float sy = (((ny * s->vpYS) + s->vpYC) - s->offY);
float sz = ((nz * s->vpZS) + s->vpZC);
return {sx,sy,sz};
}
}
// tools/platform/psp/ge_raster.go:1165:1
float psp_fogCoef(psp_geState* s,float w){
{
if ((!s->fogOn)) {
return cast<float>(1ULL);
}
float f = (((s->fogEnd - w)) * s->fogScale);
return psp_clampF(f,cast<float>(0ULL),cast<float>(1ULL));
}
}
// tools/platform/psp/ge_raster.go:1175:1
float psp_invW(float w){
{
if ((w == cast<float>(0ULL))) {
return cast<float>(1ULL);
}
return (cast<float>(1ULL) / w);
}
}
// tools/platform/psp/ge_raster.go:1195:1
float psp_nearDist(psp_vert v){
{
return (v.cz + v.cw);
}
}
// tools/platform/psp/ge_raster.go:1206:1
Slice<std::array<psp_vert,3>> psp_clipTriNear(psp_geState* s,psp_vert a,psp_vert b,psp_vert c){
{
if ((((!a.clip) || (!b.clip)) || (!c.clip))) {
return Slice<std::array<psp_vert,3>>{std::array<psp_vert,3>{a,b,c}};
}
auto tmp140 = std::make_tuple(psp_nearDist(a),psp_nearDist(b),psp_nearDist(c));
float da = std::get<0>(tmp140);
float db = std::get<1>(tmp140);
float dc = std::get<2>(tmp140);
if ((((da > psp_nearEps) && (db > psp_nearEps)) && (dc > psp_nearEps))) {
return Slice<std::array<psp_vert,3>>{std::array<psp_vert,3>{a,b,c}};
}
Slice<psp_vert> in = Slice<psp_vert>{a,b,c};
Slice<float> din = Slice<float>{da,db,dc};
Slice<psp_vert> out={};
{int64_t i = cast<int64_t>(0ULL);for (;(i < len(in));i++){
int64_t j = modi<int64_t>((cast<int64_t>((i + cast<int64_t>(1ULL)))),len(in));
auto tmp141 = std::make_tuple(in[i],in[j]);
psp_vert cur = std::get<0>(tmp141);
psp_vert nxt = std::get<1>(tmp141);
auto tmp142 = std::make_tuple(din[i],din[j]);
float dCur = std::get<0>(tmp142);
float dNxt = std::get<1>(tmp142);
auto tmp143 = std::make_tuple((dCur > psp_nearEps),(dNxt > psp_nearEps));
bool curIn = std::get<0>(tmp143);
bool nxtIn = std::get<1>(tmp143);
if (curIn) {
out = append(out,cur);
}
if ((curIn != nxtIn)) {
float t = (((psp_nearEps - dCur)) / ((dNxt - dCur)));
out = append(out,psp_lerpVert(s,cur,nxt,t));
}
}
}if ((len(out) < cast<int64_t>(3ULL))) {
return {};
}
Slice<std::array<psp_vert,3>> tris = Slice<std::array<psp_vert,3>>::make(cast<int64_t>(0ULL),cast<int64_t>(2ULL));
{int64_t i = cast<int64_t>(1ULL);for (;(cast<int64_t>((i + cast<int64_t>(1ULL))) < len(out));i++){
tris = append(tris,std::array<psp_vert,3>{out[cast<int64_t>(0ULL)],out[i],out[cast<int64_t>((i + cast<int64_t>(1ULL)))]});
}
}return tris;
}
}
// tools/platform/psp/ge_raster.go:1241:1
psp_vert psp_lerpVert(psp_geState* s,psp_vert p,psp_vert q,float t){
{
auto li = [&](float a,float b)->float{
return (a + (((b - a)) * t));
}
;
auto lb = [&](uint8_t a,uint8_t b)->uint8_t{
return cast<uint8_t>((cast<float>(a) + (((cast<float>(b) - cast<float>(a))) * t)));
}
;
psp_vert v={};
v.clip = true;
auto tmp144 = std::make_tuple(li(p.cx,q.cx),li(p.cy,q.cy));
v.cx = std::get<0>(tmp144);
v.cy = std::get<1>(tmp144);
auto tmp145 = std::make_tuple(li(p.cz,q.cz),li(p.cw,q.cw));
v.cz = std::get<0>(tmp145);
v.cw = std::get<1>(tmp145);
auto tmp146 = std::make_tuple(li(p.u,q.u),li(p.v,q.v));
v.u = std::get<0>(tmp146);
v.v = std::get<1>(tmp146);
auto tmp147 = std::make_tuple(lb(p.r,q.r),lb(p.g,q.g),lb(p.b,q.b),lb(p.a,q.a));
v.r = std::get<0>(tmp147);
v.g = std::get<1>(tmp147);
v.b = std::get<2>(tmp147);
v.a = std::get<3>(tmp147);
auto tmp148 = psp_project(s,v.cx,v.cy,v.cz,v.cw);
v.x = std::get<0>(tmp148);
v.y = std::get<1>(tmp148);
v.z = std::get<2>(tmp148);
v.invW = psp_invW(v.cw);
v.fog = psp_fogCoef(s,v.cw);
return v;
}
}
// tools/platform/psp/ge_raster.go:1258:1
void psp_ident(std::array<float,16>* m){
{
(*m) = std::array<float,16>{cast<float>(1ULL),cast<float>(0ULL),cast<float>(0ULL),cast<float>(0ULL),cast<float>(0ULL),cast<float>(1ULL),cast<float>(0ULL),cast<float>(0ULL),cast<float>(0ULL),cast<float>(0ULL),cast<float>(1ULL),cast<float>(0ULL),cast<float>(0ULL),cast<float>(0ULL),cast<float>(0ULL),cast<float>(1ULL)};
}
}
// tools/platform/psp/ge_raster.go:1263:1
void psp_matPush(std::array<float,16>* m,int64_t* idx,uint32_t arg){
{
if (((*idx) >= cast<int64_t>(12ULL))) {
return ;
}
int64_t col = divi<int64_t>((*idx),cast<int64_t>(3ULL));
int64_t row = modi<int64_t>((*idx),cast<int64_t>(3ULL));
(*m)[cast<int64_t>((cast<int64_t>((col * cast<int64_t>(4ULL))) + row))] = go_math_Float32frombits(shl<uint32_t>(arg,cast<int64_t>(8ULL)));
if ((row == cast<int64_t>(2ULL))) {
(*m)[cast<int64_t>((cast<int64_t>((col * cast<int64_t>(4ULL))) + cast<int64_t>(3ULL)))] = cast<float>(0ULL);
}
if (((*idx) == cast<int64_t>(11ULL))) {
(*m)[cast<int64_t>(15ULL)] = cast<float>(1ULL);
}
(*idx)++;
}
}
// tools/platform/psp/ge_raster.go:1280:1
void psp_matPush16(std::array<float,16>* m,int64_t* idx,uint32_t arg){
{
if (((*idx) >= cast<int64_t>(16ULL))) {
return ;
}
(*m)[(*idx)] = go_math_Float32frombits(shl<uint32_t>(arg,cast<int64_t>(8ULL)));
(*idx)++;
}
}
// tools/platform/psp/ge_raster.go:1288:1
std::array<float,16> psp_mul4(std::array<float,16> a,std::array<float,16> b){
{
std::array<float,16> r={};
{int64_t c = cast<int64_t>(0ULL);for (;(c < cast<int64_t>(4ULL));c++){
{int64_t rr = cast<int64_t>(0ULL);for (;(rr < cast<int64_t>(4ULL));rr++){
float s={};
{int64_t k = cast<int64_t>(0ULL);for (;(k < cast<int64_t>(4ULL));k++){
s += (a[cast<int64_t>((cast<int64_t>((k * cast<int64_t>(4ULL))) + rr))] * b[cast<int64_t>((cast<int64_t>((c * cast<int64_t>(4ULL))) + k))]);
}
}r[cast<int64_t>((cast<int64_t>((c * cast<int64_t>(4ULL))) + rr))] = s;
}
}}
}return r;
}
}
// tools/platform/psp/ge_raster.go:1303:1
float psp_f24(uint32_t arg){
{
return go_math_Float32frombits(shl<uint32_t>(arg,cast<int64_t>(8ULL)));
}
}
// tools/platform/psp/ge_raster.go:1306:1
float psp_f32(psp_Machine* m,uint32_t a){
{
return go_math_Float32frombits(psp_Machine_read32(m,a));
}
}
// tools/platform/psp/ge_wire.go:137:1
psp_geWire* psp_geWireOf(psp_geState* s){
{
return arenaNew(psp_geWire{s->fbLow,s->fbHigh,s->fbStride,s->fbFmt,s->base,s->vaddr,s->iaddr,s->offAddr,s->vtype,s->patchDivU,s->patchDivV,s->lightOn,s->matDiffuse,s->matEmissive,s->ambientCol,s->lightEnable,s->lightType,s->lightPos,s->lightDiff,cast<float>(s->vpXS),cast<float>(s->vpYS),cast<float>(s->vpZS),cast<float>(s->vpXC),cast<float>(s->vpYC),cast<float>(s->vpZC),cast<float>(s->offX),cast<float>(s->offY),s->world,s->view,s->proj,s->worldIdx,s->viewIdx,s->projIdx,s->matColor,s->clearOn,s->texEnable,s->texAddrN,s->texStrideN,s->texWN,s->texHN,s->texMaxLvl,s->texLodMode,cast<float>(s->texLodBias),s->texAddr,s->texStride,s->texW,s->texH,s->texFmt,s->texSwizzle,cast<float>(s->texScaleU),cast<float>(s->texScaleV),cast<float>(s->texOffU),cast<float>(s->texOffV),s->clutAddr,s->clutFmt,s->clut,s->blendOn,s->blendSrc,s->blendDst,s->blendEq,s->blendFixA,s->blendFixB,s->alphaTestOn,s->alphaFunc,s->alphaRef,s->alphaTestMask,s->texMapMode,s->texProjSrc,s->texWrapU,s->texWrapV,s->texLinear,s->texFunc,s->texUseA,s->texDouble,s->texEnvCol,s->texMtx,s->texMtxIdx,s->scX0,s->scY0,s->scX1,s->scY1,s->maskRGB,s->maskA,s->zLow,s->zHigh,s->zStride,s->zTestOn,s->zFunc,s->zNoWrite,s->clearDepth,s->cullOn,s->cullFace,s->stencilOn,s->stFunc,s->stRef,s->stMask,s->stSFail,s->stZFail,s->stZPass,s->fogOn,cast<float>(s->fogEnd),cast<float>(s->fogScale),s->fogColor,s->trSrc,s->trDst,s->trSrcStride,s->trDstStride,s->trSrcX,s->trSrcY,s->trDstX,s->trDstY,s->trW,s->trH});
}
}
// tools/platform/psp/ge_wire.go:256:1
psp_geState* psp_geWire_state(psp_geWire* w){
{
return arenaNew(psp_geState{w->FbLow,w->FbHigh,w->FbStride,w->FbFmt,w->Base,w->Vaddr,w->Iaddr,w->OffAddr,w->Vtype,w->PatchDivU,w->PatchDivV,w->LightOn,w->MatDiffuse,w->MatEmissive,w->AmbientCol,w->LightEnable,w->LightType,w->LightPos,w->LightDiff,cast<float>(w->VpXS),cast<float>(w->VpYS),cast<float>(w->VpZS),cast<float>(w->VpXC),cast<float>(w->VpYC),cast<float>(w->VpZC),cast<float>(w->OffX),cast<float>(w->OffY),w->World,w->View,w->Proj,w->WorldIdx,w->ViewIdx,w->ProjIdx,w->MatColor,w->ClearOn,w->TexEnable,w->TexAddrN,w->TexStrideN,w->TexWN,w->TexHN,w->TexMaxLvl,w->TexLodMode,cast<float>(w->TexLodBias),w->TexAddr,w->TexStride,w->TexW,w->TexH,w->TexFmt,w->TexSwizzle,cast<float>(w->TexScaleU),cast<float>(w->TexScaleV),cast<float>(w->TexOffU),cast<float>(w->TexOffV),w->ClutAddr,w->ClutFmt,w->Clut,w->BlendOn,w->BlendSrc,w->BlendDst,w->BlendEq,w->BlendFixA,w->BlendFixB,w->AlphaTestOn,w->AlphaFunc,w->AlphaRef,w->AlphaTestMask,w->TexMapMode,w->TexProjSrc,w->TexWrapU,w->TexWrapV,w->TexLinear,w->TexFunc,w->TexUseA,w->TexDouble,w->TexEnvCol,w->TexMtx,w->TexMtxIdx,w->ScX0,w->ScY0,w->ScX1,w->ScY1,w->MaskRGB,w->MaskA,w->ZLow,w->ZHigh,w->ZStride,w->ZTestOn,w->ZFunc,w->ZNoWrite,w->ClearDepth,w->CullOn,w->CullFace,w->StencilOn,w->StFunc,w->StRef,w->StMask,w->StSFail,w->StZFail,w->StZPass,w->FogOn,cast<float>(w->FogEnd),cast<float>(w->FogScale),w->FogColor,w->TrSrc,w->TrDst,w->TrSrcStride,w->TrDstStride,w->TrSrcX,w->TrSrcY,w->TrDstX,w->TrDstY,w->TrW,w->TrH});
}
}
// tools/platform/psp/intr.go:30:1
void psp_Machine_registerSubIntr(psp_Machine* m,uint32_t intno,uint32_t subno,uint32_t handler,uint32_t arg){
{
m->subIntrs[cast<uint32_t>((shl<uint32_t>(intno,cast<int64_t>(16ULL)) | subno))] = arenaNew(psp_subIntr{intno,subno,handler,arg,{}});
}
}
// tools/platform/psp/intr.go:34:1
void psp_Machine_setSubIntrEnabled(psp_Machine* m,uint32_t intno,uint32_t subno,bool on){
{
{
psp_subIntr* si = get(m->subIntrs,cast<uint32_t>((shl<uint32_t>(intno,cast<int64_t>(16ULL)) | subno)));
if (bool(si)) {
si->enabled = on;
}
}
}
}
// tools/platform/psp/intr.go:42:1
void psp_Machine_deliverVBlank(psp_Machine* m){
{
m->vblanks++;
{;for (;((len(m->padScript) > cast<int64_t>(0ULL)) && (m->vblanks >= m->padScript[cast<int64_t>(0ULL)].AtVblank));){
m->pad = m->padScript[cast<int64_t>(0ULL)].Buttons;
m->padScript = sub(m->padScript,cast<int64_t>(1ULL),len(m->padScript));
}
}{auto&& tmp149 = m->handles;
for(auto [tmp150,tmp151]:tmp149){
auto o=tmp151;if (((((o->kind == std::string("thread",6)) && (o->tstate == cast<psp_threadState>(3ULL))) && (o->wakeVblank != cast<uint32_t>(0ULL))) && (m->vblanks >= o->wakeVblank))) {
o->wakeVblank = cast<uint32_t>(0ULL);
o->tstate = cast<psp_threadState>(1ULL);
}
}}
{auto&& tmp152 = m->subIntrs;
for(auto [tmp153,tmp154]:tmp152){
auto si=tmp154;if ((si->enabled && (si->intno == cast<uint32_t>(30ULL)))) {
psp_Machine_callGuest(m,si->handler,si->subno,si->arg);
}
}}
}
}
// tools/platform/psp/io.go:32:1
void psp_Machine_SetVolume(psp_Machine* m,psp_Volume* v){
{
m->vol = v;
}
}
// tools/platform/psp/io.go:36:1
std::string psp_devicePath(std::string p){
{
{
int64_t i = go_strings_IndexByte(p,cast<uint8_t>(58ULL));
if ((i >= cast<int64_t>(0ULL))) {
p = sub(p,cast<int64_t>((i + cast<int64_t>(1ULL))),len(p));
}
}
return go_strings_TrimLeft(p,std::string("/",1));
}
}
// tools/platform/psp/io.go:43:1
uint32_t psp_Machine_ioOpen(psp_Machine* m,std::string path){
{
if ((!m->vol)) {
psp_Machine_note(m,std::string("sceIoOpen(%q) with no volume mounted",36),path);
return cast<uint32_t>(2147549186ULL);
}
std::string vp = psp_devicePath(path);
auto tmp155 = psp_Volume_resolve(m->vol,vp);
psp_Entry e = std::get<0>(tmp155);
Error err = std::get<1>(tmp155);
if ((bool(err) || e.IsDir)) {
psp_Machine_note(m,std::string("sceIoOpen(%q): not found",24),path);
return cast<uint32_t>(2147549186ULL);
}
uint32_t fd = m->nextFd;
m->nextFd++;
m->files[fd] = arenaNew(psp_ioFile{vp,e,cast<int64_t>(0ULL),{},{},{},{}});
psp_Machine_note(m,std::string("sceIoOpen(%q) -> fd %d (lba %d size %d)",39),path,fd,e.Block,e.Size);
return fd;
}
}
// tools/platform/psp/io.go:61:1
uint32_t psp_Machine_ioClose(psp_Machine* m,uint32_t fd){
{
{
auto tmp156 = lookup(m->files,fd);
bool ok = std::get<1>(tmp156);
if ((!ok)) {
return cast<uint32_t>(2147549193ULL);
}
}
removeKey(m->files,fd);
return cast<uint32_t>(0ULL);
}
}
// tools/platform/psp/io.go:79:1
uint32_t psp_Machine_ioOpenAsync(psp_Machine* m,std::string path){
{
uint32_t fd = psp_Machine_ioOpen(m,path);
if ((cast<uint32_t>((fd & cast<uint32_t>(2147483648ULL))) == cast<uint32_t>(0ULL))) {
psp_ioFile* f = get(m->files,fd);
auto tmp157 = std::make_tuple(cast<int64_t>(fd),true);
f->async = std::get<0>(tmp157);
f->hasAsync = std::get<1>(tmp157);
}
return fd;
}
}
// tools/platform/psp/io.go:88:1
uint32_t psp_Machine_ioReadAsync(psp_Machine* m,uint32_t fd,uint32_t buf,uint32_t n){
{
auto tmp158 = lookup(m->files,fd);
psp_ioFile* f = std::get<0>(tmp158);
bool ok = std::get<1>(tmp158);
if ((!ok)) {
return cast<uint32_t>(2147549193ULL);
}
auto tmp159 = std::make_tuple(cast<int64_t>(cast<int32_t>(psp_Machine_ioRead(m,fd,buf,n))),true);
f->async = std::get<0>(tmp159);
f->hasAsync = std::get<1>(tmp159);
return cast<uint32_t>(0ULL);
}
}
// tools/platform/psp/io.go:97:1
uint32_t psp_Machine_ioLseekAsync(psp_Machine* m,uint32_t fd,int64_t off,uint32_t whence){
{
auto tmp160 = lookup(m->files,fd);
psp_ioFile* f = std::get<0>(tmp160);
bool ok = std::get<1>(tmp160);
if ((!ok)) {
return cast<uint32_t>(2147549193ULL);
}
auto tmp161 = std::make_tuple(psp_Machine_ioLseek(m,fd,off,whence),true);
f->async = std::get<0>(tmp161);
f->hasAsync = std::get<1>(tmp161);
return cast<uint32_t>(0ULL);
}
}
// tools/platform/psp/io.go:106:1
uint32_t psp_Machine_ioCloseAsync(psp_Machine* m,uint32_t fd){
{
auto tmp162 = lookup(m->files,fd);
psp_ioFile* f = std::get<0>(tmp162);
bool ok = std::get<1>(tmp162);
if ((!ok)) {
return cast<uint32_t>(2147549193ULL);
}
auto tmp163 = std::make_tuple(cast<int64_t>(0ULL),true);
f->async = std::get<0>(tmp163);
f->hasAsync = std::get<1>(tmp163);
return cast<uint32_t>(0ULL);
}
}
// tools/platform/psp/io.go:121:1
uint32_t psp_Machine_ioWaitAsync(psp_Machine* m,uint32_t fd,uint32_t resPtr){
{
auto tmp164 = lookup(m->files,fd);
psp_ioFile* f = std::get<0>(tmp164);
bool ok = std::get<1>(tmp164);
if ((!ok)) {
return cast<uint32_t>(2147549193ULL);
}
if ((!f->hasAsync)) {
return cast<uint32_t>(2147615521ULL);
}
if ((resPtr != cast<uint32_t>(0ULL))) {
psp_Machine_write32(m,resPtr,cast<uint32_t>(f->async));
psp_Machine_write32(m,cast<uint32_t>((resPtr + cast<uint32_t>(4ULL))),cast<uint32_t>(shr<int64_t>(f->async,cast<int64_t>(32ULL))));
}
f->hasAsync = false;
return cast<uint32_t>(0ULL);
}
}
// tools/platform/psp/io.go:141:1
uint32_t psp_Machine_ioPollAsync(psp_Machine* m,uint32_t fd,uint32_t resPtr){
{
auto tmp165 = lookup(m->files,fd);
psp_ioFile* f = std::get<0>(tmp165);
bool ok = std::get<1>(tmp165);
if ((!ok)) {
return cast<uint32_t>(2147549193ULL);
}
if ((!f->hasAsync)) {
return cast<uint32_t>(2147615521ULL);
}
if ((resPtr != cast<uint32_t>(0ULL))) {
psp_Machine_write32(m,resPtr,cast<uint32_t>(f->async));
psp_Machine_write32(m,cast<uint32_t>((resPtr + cast<uint32_t>(4ULL))),cast<uint32_t>(shr<int64_t>(f->async,cast<int64_t>(32ULL))));
}
f->hasAsync = false;
return cast<uint32_t>(0ULL);
}
}
// tools/platform/psp/io.go:158:1
uint32_t psp_Machine_ioRead(psp_Machine* m,uint32_t fd,uint32_t buf,uint32_t n){
{
auto tmp166 = lookup(m->files,fd);
psp_ioFile* f = std::get<0>(tmp166);
bool ok = std::get<1>(tmp166);
if ((!ok)) {
return cast<uint32_t>(2147549193ULL);
}
Slice<uint8_t> p = Slice<uint8_t>::make(n);
auto tmp167 = psp_Volume_ReadFileAt(m->vol,f->ent,cast<int64_t>(f->pos),p);
int64_t got = std::get<0>(tmp167);
Error err = std::get<1>(tmp167);
if (bool(err)) {
psp_Machine_note(m,std::string("sceIoRead(%q at %d): %v",23),f->path,f->pos,err);
return cast<uint32_t>(2147549193ULL);
}
{int64_t i = cast<int64_t>(0ULL);for (;(i < got);i++){
psp_Machine_Write(m,cast<uint32_t>((buf + cast<uint32_t>(i))),p[i]);
}
}f->pos += cast<int64_t>(got);
psp_Machine_note(m,std::string("sceIoRead(%q at %d, %d) -> %d",29),f->path,cast<int64_t>((f->pos - cast<int64_t>(got))),n,got);
return cast<uint32_t>(got);
}
}
// tools/platform/psp/io.go:179:1
uint32_t psp_Machine_ioWrite(psp_Machine* m,uint32_t fd,uint32_t buf,uint32_t n){
{
if (((fd == cast<uint32_t>(1ULL)) || (fd == cast<uint32_t>(2ULL)))) {
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < n);i++){
m->tty = append(m->tty,psp_Machine_Read(m,cast<uint32_t>((buf + i))));
}
}return n;
}
psp_Machine_note(m,std::string("sceIoWrite to fd %d refused (read-only volume)",46),fd);
return cast<uint32_t>(2147549193ULL);
}
}
// tools/platform/psp/io.go:193:1
void psp_Machine_fillStat(psp_Machine* m,psp_Entry e,uint32_t stat){
{
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < cast<uint32_t>(88ULL));i += cast<uint32_t>(4ULL)){
psp_Machine_write32(m,cast<uint32_t>((stat + i)),cast<uint32_t>(0ULL));
}
}auto tmp168 = std::make_tuple(cast<uint32_t>(8703ULL),cast<uint32_t>(32ULL));
uint32_t mode = std::get<0>(tmp168);
uint32_t attr = std::get<1>(tmp168);
if (e.IsDir) {
auto tmp169 = std::make_tuple(cast<uint32_t>(4607ULL),cast<uint32_t>(16ULL));
mode = std::get<0>(tmp169);
attr = std::get<1>(tmp169);
}
psp_Machine_write32(m,cast<uint32_t>((stat + cast<uint32_t>(0ULL))),mode);
psp_Machine_write32(m,cast<uint32_t>((stat + cast<uint32_t>(4ULL))),attr);
psp_Machine_write32(m,cast<uint32_t>((stat + cast<uint32_t>(8ULL))),cast<uint32_t>(e.Size));
psp_Machine_write32(m,cast<uint32_t>((stat + cast<uint32_t>(12ULL))),cast<uint32_t>(0ULL));
psp_Machine_write32(m,cast<uint32_t>((stat + cast<uint32_t>(64ULL))),cast<uint32_t>(e.Block));
}
}
// tools/platform/psp/io.go:209:1
uint32_t psp_Machine_ioGetstat(psp_Machine* m,std::string path,uint32_t stat){
{
if ((!m->vol)) {
return cast<uint32_t>(2147549186ULL);
}
std::string vp = psp_devicePath(path);
auto tmp170 = psp_Volume_resolve(m->vol,vp);
psp_Entry e = std::get<0>(tmp170);
Error err = std::get<1>(tmp170);
if (bool(err)) {
psp_Machine_note(m,std::string("sceIoGetstat(%q): not found",27),path);
return cast<uint32_t>(2147549186ULL);
}
psp_Machine_fillStat(m,e,stat);
return cast<uint32_t>(0ULL);
}
}
// tools/platform/psp/io.go:232:1
uint32_t psp_Machine_ioDopen(psp_Machine* m,std::string path){
{
if ((!m->vol)) {
return cast<uint32_t>(2147549186ULL);
}
std::string vp = psp_devicePath(path);
auto tmp171 = psp_Volume_ReadDir(m->vol,vp);
Slice<psp_Entry> ents = std::get<0>(tmp171);
Error err = std::get<1>(tmp171);
if (bool(err)) {
psp_Machine_note(m,std::string("sceIoDopen(%q): %v",18),path,err);
return cast<uint32_t>(2147549186ULL);
}
uint32_t fd = m->nextFd;
m->nextFd++;
m->files[fd] = arenaNew(psp_ioFile{vp,{},{},{},{},ents,{}});
psp_Machine_note(m,std::string("sceIoDopen(%q) -> fd %d (%d entries)",36),path,fd,len(ents));
return fd;
}
}
// tools/platform/psp/io.go:249:1
uint32_t psp_Machine_ioDread(psp_Machine* m,uint32_t fd,uint32_t dirent){
{
auto tmp172 = lookup(m->files,fd);
psp_ioFile* f = std::get<0>(tmp172);
bool ok = std::get<1>(tmp172);
if (((!ok) || (!f->dir))) {
return cast<uint32_t>(2147549193ULL);
}
int64_t left = cast<int64_t>((len(f->dir) - f->dirPos));
if ((left <= cast<int64_t>(0ULL))) {
return cast<uint32_t>(0ULL);
}
psp_Entry e = f->dir[f->dirPos];
f->dirPos++;
psp_Machine_fillStat(m,e,dirent);
std::string name = e.Name;
{
int64_t i = go_strings_IndexByte(name,cast<uint8_t>(59ULL));
if ((i >= cast<int64_t>(0ULL))) {
name = sub(name,0,i);
}
}
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(256ULL));i++){
uint8_t b={};
if ((i < len(name))) {
b = name[i];
}
psp_Machine_Write(m,cast<uint32_t>((cast<uint32_t>((dirent + cast<uint32_t>(88ULL))) + cast<uint32_t>(i))),b);
}
}psp_Machine_write32(m,cast<uint32_t>((dirent + cast<uint32_t>(344ULL))),cast<uint32_t>(0ULL));
psp_Machine_write32(m,cast<uint32_t>((dirent + cast<uint32_t>(348ULL))),cast<uint32_t>(0ULL));
psp_Machine_note(m,std::string("sceIoDread(fd %d) -> %q (size %d)",33),fd,name,e.Size);
return cast<uint32_t>(left);
}
}
// tools/platform/psp/io.go:278:1
uint32_t psp_Machine_ioDclose(psp_Machine* m,uint32_t fd){
{
auto tmp173 = lookup(m->files,fd);
psp_ioFile* f = std::get<0>(tmp173);
bool ok = std::get<1>(tmp173);
if (((!ok) || (!f->dir))) {
return cast<uint32_t>(2147549193ULL);
}
removeKey(m->files,fd);
return cast<uint32_t>(0ULL);
}
}
// tools/platform/psp/io.go:288:1
int64_t psp_Machine_ioLseek(psp_Machine* m,uint32_t fd,int64_t offset,uint32_t whence){
{
auto tmp174 = lookup(m->files,fd);
psp_ioFile* f = std::get<0>(tmp174);
bool ok = std::get<1>(tmp174);
if ((!ok)) {
return cast<int64_t>(-1ULL);
}
{
switch(whence){
case cast<uint32_t>(0ULL):{
f->pos = offset;
break;}
case cast<uint32_t>(1ULL):{
f->pos += offset;
break;}
case cast<uint32_t>(2ULL):{
f->pos = cast<int64_t>((cast<int64_t>(f->ent.Size) + offset));
break;}
}}
if ((f->pos < cast<int64_t>(0ULL))) {
f->pos = cast<int64_t>(0ULL);
}
return f->pos;
}
}
// tools/platform/psp/iso.go:36:1
std::tuple<Slice<uint8_t>,Error> psp_byteBlocks_ReadBlock(psp_byteBlocks b,int64_t n){
{
int64_t off = cast<int64_t>((n * cast<int64_t>(2048ULL)));
if (((off < cast<int64_t>(0ULL)) || (cast<int64_t>((off + cast<int64_t>(2048ULL))) > len(b)))) {
return {{},go_fmt_Errorf(std::string("iso: block %d out of range",26),n)};
}
return {sub(b,off,cast<int64_t>((off + cast<int64_t>(2048ULL)))),{}};
}
}
// tools/platform/psp/iso.go:64:1
std::string psp_Entry_String(psp_Entry e){
{
if (e.IsDir) {
return go_fmt_Sprintf(std::string("%-32s  <dir>   (lba %d)",23),e.Path,e.Block);
}
return go_fmt_Sprintf(std::string("%-32s  %10d  (lba %d)",21),e.Path,e.Size,e.Block);
}
}
// tools/platform/psp/iso.go:73:1
std::tuple<psp_Volume*,Error> psp_OpenVolume(BlockSource* src){
{
psp_Volume* v = arenaNew(psp_Volume{src,{},{},{},{}});
auto tmp175 = blockSource_ReadBlock(src,cast<int64_t>(16ULL));
Slice<uint8_t> pvd = std::get<0>(tmp175);
Error err = std::get<1>(tmp175);
if (bool(err)) {
return {{},go_fmt_Errorf(std::string("iso: reading PVD: %w",20),err)};
}
if (((pvd[cast<int64_t>(0ULL)] != cast<uint8_t>(1ULL)) || (cast<std::string>(sub(pvd,cast<int64_t>(1ULL),cast<int64_t>(6ULL))) != std::string("CD001",5)))) {
return {{},go_errors_New(std::string("iso: no ISO 9660 Primary Volume Descriptor at LBA 16",52))};
}
v->System = go_strings_TrimRight(cast<std::string>(sub(pvd,cast<int64_t>(8ULL),cast<int64_t>(40ULL))),std::string(" \000",2));
v->Name = go_strings_TrimRight(cast<std::string>(sub(pvd,cast<int64_t>(40ULL),cast<int64_t>(72ULL))),std::string(" \000",2));
Slice<uint8_t> root = sub(pvd,cast<int64_t>(156ULL),cast<int64_t>(190ULL));
v->rootLBA = cast<int64_t>(psp_le32(sub(root,cast<int64_t>(2ULL),len(root))));
v->rootSize = cast<int64_t>(psp_le32(sub(root,cast<int64_t>(10ULL),len(root))));
return {v,{}};
}
}
// tools/platform/psp/iso.go:101:1
uint32_t psp_le32(Slice<uint8_t> b){
{
return le_Uint32(sub(b,0,cast<int64_t>(4ULL)));
}
}
// tools/platform/psp/iso.go:105:1
std::tuple<Slice<psp_Entry>,Error> psp_Volume_dirEntries(psp_Volume* v,int64_t lba,int64_t size,std::string dirPath){
{
Slice<psp_Entry> out={};
int64_t remaining = size;
{int64_t sect = cast<int64_t>(0ULL);for (;(remaining > cast<int64_t>(0ULL));sect++){
auto tmp176 = blockSource_ReadBlock(v->src,cast<int64_t>((lba + sect)));
Slice<uint8_t> b = std::get<0>(tmp176);
Error err = std::get<1>(tmp176);
if (bool(err)) {
return {{},err};
}
int64_t n = cast<int64_t>(2048ULL);
if ((remaining < n)) {
n = remaining;
}
remaining -= cast<int64_t>(2048ULL);
{int64_t p = cast<int64_t>(0ULL);for (;(p < n);){
int64_t recLen = cast<int64_t>(b[p]);
if ((recLen == cast<int64_t>(0ULL))) {
break;
}
if ((cast<int64_t>((p + recLen)) > cast<int64_t>(2048ULL))) {
return {{},go_fmt_Errorf(std::string("iso: directory record overruns sector at lba %d",47),cast<int64_t>((lba + sect)))};
}
Slice<uint8_t> rec = sub(b,p,cast<int64_t>((p + recLen)));
p += recLen;
int64_t idLen = cast<int64_t>(rec[cast<int64_t>(32ULL)]);
Slice<uint8_t> id = sub(rec,cast<int64_t>(33ULL),cast<int64_t>((cast<int64_t>(33ULL) + idLen)));
if (((idLen == cast<int64_t>(1ULL)) && (((id[cast<int64_t>(0ULL)] == cast<uint8_t>(0ULL)) || (id[cast<int64_t>(0ULL)] == cast<uint8_t>(1ULL)))))) {
continue;
}
uint8_t flags = rec[cast<int64_t>(25ULL)];
psp_Entry e = psp_Entry{cast<std::string>(id),{},(cast<uint8_t>((flags & cast<uint8_t>(2ULL))) != cast<uint8_t>(0ULL)),cast<int64_t>(psp_le32(sub(rec,cast<int64_t>(10ULL),len(rec)))),cast<int64_t>(psp_le32(sub(rec,cast<int64_t>(2ULL),len(rec))))};
if ((dirPath == std::string("",0))) {
e.Path = e.Name;
}
else {
e.Path = ((dirPath + std::string("/",1)) + e.Name);
}
out = append(out,e);
}
}}
}return {out,{}};
}
}
// tools/platform/psp/iso.go:155:1
bool psp_nameEqual(std::string entry,std::string want){
{
if (go_strings_EqualFold(entry,want)) {
return true;
}
if ((!go_strings_Contains(want,std::string(";",1)))) {
{
int64_t i = go_strings_IndexByte(entry,cast<uint8_t>(59ULL));
if ((i >= cast<int64_t>(0ULL))) {
return go_strings_EqualFold(sub(entry,0,i),want);
}
}
}
return false;
}
}
// tools/platform/psp/iso.go:167:1
Slice<std::string> psp_splitPath(std::string p){
{
Slice<std::string> parts={};
{auto&& tmp177 = go_strings_Split(go_strings_Trim(p,std::string("/",1)),std::string("/",1));
for(int64_t tmp178=0;tmp178<len(tmp177);++tmp178){
auto c=tmp177[tmp178];if ((c != std::string("",0))) {
parts = append(parts,c);
}
}}
return parts;
}
}
// tools/platform/psp/iso.go:181:1
std::tuple<psp_Entry,Error> psp_Volume_resolve(psp_Volume* v,std::string path){
{
{
auto tmp179 = psp_parseLbnPath(path);
int64_t lbn = std::get<0>(tmp179);
int64_t size = std::get<1>(tmp179);
bool ok = std::get<2>(tmp179);
if (ok) {
return {psp_Entry{path,path,{},size,lbn},{}};
}
}
psp_Entry cur = psp_Entry{v->Name,std::string("",0),true,v->rootSize,v->rootLBA};
{auto&& tmp180 = psp_splitPath(path);
for(int64_t tmp181=0;tmp181<len(tmp180);++tmp181){
auto want=tmp180[tmp181];if ((!cur.IsDir)) {
return {psp_Entry{},go_fmt_Errorf(std::string("iso: %q is not a directory",26),cur.Path)};
}
auto tmp182 = psp_Volume_dirEntries(v,cur.Block,cur.Size,cur.Path);
Slice<psp_Entry> entries = std::get<0>(tmp182);
Error err = std::get<1>(tmp182);
if (bool(err)) {
return {psp_Entry{},err};
}
bool found = false;
{auto&& tmp183 = entries;
for(int64_t tmp184=0;tmp184<len(tmp183);++tmp184){
auto e=tmp183[tmp184];if (psp_nameEqual(e.Name,want)) {
auto tmp185 = std::make_tuple(e,true);
cur = std::get<0>(tmp185);
found = std::get<1>(tmp185);
break;
}
}}
if ((!found)) {
return {psp_Entry{},go_fmt_Errorf(std::string("iso: %q not found",17),path)};
}
}}
return {cur,{}};
}
}
// tools/platform/psp/iso.go:211:1
std::tuple<int64_t,int64_t,bool> psp_parseLbnPath(std::string path){
int64_t lbn{};
int64_t size{};
bool ok{};
{
std::string p = go_strings_ToLower(go_strings_Trim(path,std::string("/",1)));
if ((!go_strings_HasPrefix(p,std::string("sce_lbn",7)))) {
return {cast<int64_t>(0ULL),cast<int64_t>(0ULL),false};
}
std::string rest = sub(p,cast<int64_t>(7ULL),len(p));
int64_t i = go_strings_Index(rest,std::string("_size",5));
if ((i < cast<int64_t>(0ULL))) {
return {cast<int64_t>(0ULL),cast<int64_t>(0ULL),false};
}
auto tmp186 = go_strconv_ParseInt(go_strings_TrimPrefix(sub(rest,0,i),std::string("0x",2)),cast<int64_t>(16ULL),cast<int64_t>(64ULL));
int64_t l = std::get<0>(tmp186);
Error err1 = std::get<1>(tmp186);
auto tmp187 = go_strconv_ParseInt(go_strings_TrimPrefix(sub(rest,cast<int64_t>((i + cast<int64_t>(5ULL))),len(rest)),std::string("0x",2)),cast<int64_t>(16ULL),cast<int64_t>(64ULL));
int64_t s = std::get<0>(tmp187);
Error err2 = std::get<1>(tmp187);
if ((bool(err1) || bool(err2))) {
return {cast<int64_t>(0ULL),cast<int64_t>(0ULL),false};
}
return {cast<int64_t>(l),cast<int64_t>(s),true};
}
}
// tools/platform/psp/iso.go:230:1
std::tuple<Slice<psp_Entry>,Error> psp_Volume_ReadDir(psp_Volume* v,std::string path){
{
auto tmp188 = psp_Volume_resolve(v,path);
psp_Entry e = std::get<0>(tmp188);
Error err = std::get<1>(tmp188);
if (bool(err)) {
return {{},err};
}
if ((!e.IsDir)) {
return {{},go_fmt_Errorf(std::string("iso: %q is not a directory",26),path)};
}
return psp_Volume_dirEntries(v,e.Block,e.Size,e.Path);
}
}
// tools/platform/psp/iso.go:242:1
Error psp_Volume_Walk(psp_Volume* v,std::function<Error(psp_Entry)> fn){
{
return psp_Volume_walk(v,v->rootLBA,v->rootSize,std::string("",0),fn);
}
}
// tools/platform/psp/iso.go:246:1
Error psp_Volume_walk(psp_Volume* v,int64_t lba,int64_t size,std::string dirPath,std::function<Error(psp_Entry)> fn){
{
auto tmp189 = psp_Volume_dirEntries(v,lba,size,dirPath);
Slice<psp_Entry> entries = std::get<0>(tmp189);
Error err = std::get<1>(tmp189);
if (bool(err)) {
return err;
}
{auto&& tmp190 = entries;
for(int64_t tmp191=0;tmp191<len(tmp190);++tmp191){
auto e=tmp190[tmp191];{
Error err = fn(e);
if (bool(err)) {
return err;
}
}
if (e.IsDir) {
{
Error err = psp_Volume_walk(v,e.Block,e.Size,e.Path,fn);
if (bool(err)) {
return err;
}
}
}
}}
return {};
}
}
// tools/platform/psp/iso.go:267:1
std::tuple<int64_t,Error> psp_Volume_ReadFileAt(psp_Volume* v,psp_Entry e,int64_t off,Slice<uint8_t> p){
{
if ((off >= e.Size)) {
return {cast<int64_t>(0ULL),{}};
}
if ((cast<int64_t>((off + len(p))) > e.Size)) {
p = sub(p,0,cast<int64_t>((e.Size - off)));
}
int64_t got = cast<int64_t>(0ULL);
{;for (;(got < len(p));){
int64_t lba = cast<int64_t>((e.Block + divi<int64_t>((cast<int64_t>((off + got))),cast<int64_t>(2048ULL))));
auto tmp192 = blockSource_ReadBlock(v->src,lba);
Slice<uint8_t> b = std::get<0>(tmp192);
Error err = std::get<1>(tmp192);
if (bool(err)) {
return {got,err};
}
got += gcopy(sub(p,got,len(p)),sub(b,modi<int64_t>((cast<int64_t>((off + got))),cast<int64_t>(2048ULL)),len(b)));
}
}return {got,{}};
}
}
// tools/platform/psp/iso.go:286:1
std::tuple<Slice<uint8_t>,Error> psp_Volume_ReadFile(psp_Volume* v,std::string path){
{
auto tmp193 = psp_Volume_resolve(v,path);
psp_Entry e = std::get<0>(tmp193);
Error err = std::get<1>(tmp193);
if (bool(err)) {
return {{},err};
}
if (e.IsDir) {
return {{},go_fmt_Errorf(std::string("iso: %q is a directory",22),path)};
}
Slice<uint8_t> out = Slice<uint8_t>::make(cast<int64_t>(0ULL),e.Size);
{int64_t got = cast<int64_t>(0ULL);for (;(got < e.Size);got += cast<int64_t>(2048ULL)){
auto tmp194 = blockSource_ReadBlock(v->src,cast<int64_t>((e.Block + divi<int64_t>(got,cast<int64_t>(2048ULL)))));
Slice<uint8_t> b = std::get<0>(tmp194);
Error err = std::get<1>(tmp194);
if (bool(err)) {
return {{},err};
}
int64_t n = cast<int64_t>((e.Size - got));
if ((n > cast<int64_t>(2048ULL))) {
n = cast<int64_t>(2048ULL);
}
out = append(out,sub(b,0,n));
}
}return {out,{}};
}
}
// tools/platform/psp/kernel.go:80:1
bool psp_Machine_evMatch(psp_Machine* m,psp_kobject* o,uint32_t bits,uint32_t mode,uint32_t outPtr){
{
bool ok = (cast<uint32_t>((o->bits & bits)) == bits);
if ((cast<uint32_t>((mode & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL))) {
ok = (cast<uint32_t>((o->bits & bits)) != cast<uint32_t>(0ULL));
}
if ((!ok)) {
return false;
}
if ((outPtr != cast<uint32_t>(0ULL))) {
psp_Machine_write32(m,outPtr,o->bits);
}
if ((cast<uint32_t>((mode & cast<uint32_t>(32ULL))) != cast<uint32_t>(0ULL))) {
o->bits &= ~(bits);
}
if ((cast<uint32_t>((mode & cast<uint32_t>(64ULL))) != cast<uint32_t>(0ULL))) {
o->bits = cast<uint32_t>(0ULL);
}
return true;
}
}
// tools/platform/psp/kernel.go:101:1
uint32_t psp_nidOf(std::string name){
{
std::array<uint8_t,20> h = go_sha1_Sum(cast<Slice<uint8_t>>(name));
return le_Uint32(sub(h,0,cast<int64_t>(4ULL)));
}
}
// tools/platform/psp/kernel.go:220:1
std::string psp_nidName(std::string lib,uint32_t nid){
{
{
auto tmp195 = lookup(psp_nidLookup,nid);
std::string name = std::get<0>(tmp195);
bool ok = std::get<1>(tmp195);
if (ok) {
return name;
}
}
return go_fmt_Sprintf(std::string("%s:0x%08X",9),lib,nid);
}
}
// tools/platform/psp/kernel.go:237:1
void psp_Machine_installStubs(psp_Machine* m,psp_Module* mod){
{
{auto&& tmp198 = mod->Imports;
for(int64_t tmp199=0;tmp199<len(tmp198);++tmp199){
auto imp=tmp198[tmp199];{auto&& tmp200 = imp.NIDs;
for(int64_t tmp201=0;tmp201<len(tmp200);++tmp201){
auto i=tmp201;auto nid=tmp200[tmp201];uint32_t stub = cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(142622720ULL) + imp.StubAddr)) + cast<uint32_t>((cast<uint32_t>(i) * cast<uint32_t>(8ULL)))));
std::string name = psp_nidName(imp.Library,nid);
uint32_t code = m->nextSyscall;
m->nextSyscall++;
m->syscalls[code] = arenaNew(psp_syscall{name,psp_handlerFor(name)});
psp_Machine_write32(m,cast<uint32_t>((stub + cast<uint32_t>(0ULL))),cast<uint32_t>(65011720ULL));
psp_Machine_write32(m,cast<uint32_t>((stub + cast<uint32_t>(4ULL))),cast<uint32_t>(((shl<uint32_t>(code,cast<int64_t>(6ULL))) | cast<uint32_t>(12ULL))));
}}
}}
}
}
// tools/platform/psp/kernel.go:252:1
bool psp_Machine_handleSyscall(psp_Machine* m,allegrex_CPU* c,uint32_t code){
{rrprof::Scope timing(3,"Kernel HLE");
{
psp_syscall* sc = get(m->syscalls,code);
if ((!sc)) {
psp_Machine_note(m,std::string("unknown syscall code 0x%X at pc 0x%08X",38),code,allegrex_CPU_CurPC(m->CPU));
psp_Machine_setRet(m,cast<uint32_t>(0ULL));
return true;
}
m->SyscallCalls[sc->name]++;
if (((psp_syscallTrace != std::string("",0)) && go_strings_Contains(sc->name,psp_syscallTrace))) {
psp_Machine_note(m,std::string("SYSCALL %s(%08X, %08X, %08X, %08X) ra=%08X",42),sc->name,psp_Machine_arg(m,cast<uint32_t>(0ULL)),psp_Machine_arg(m,cast<uint32_t>(1ULL)),psp_Machine_arg(m,cast<uint32_t>(2ULL)),psp_Machine_arg(m,cast<uint32_t>(3ULL)),allegrex_CPU_Reg(m->CPU,cast<uint32_t>(31ULL)));
}
auto tmp202 = std::make_tuple(psp_Machine_profStart(m),psp_Machine_profGeNs(m),m->prof.gen);
time_Time ts = std::get<0>(tmp202);
int64_t geBefore = std::get<1>(tmp202);
int64_t gen = std::get<2>(tmp202);
if (bool(sc->handler)) {
sc->handler(m);
}
else {
psp_Machine_setRet(m,cast<uint32_t>(0ULL));
}
psp_Machine_profEndSyscall(m,ts,geBefore,gen);
if (((psp_syscallTrace != std::string("",0)) && go_strings_Contains(sc->name,psp_syscallTrace))) {
psp_Machine_note(m,std::string("SYSCALL %s -> %08X",18),sc->name,allegrex_CPU_Reg(m->CPU,cast<uint32_t>(2ULL)));
}
return true;
}
}
}
// tools/platform/psp/kernel.go:282:1
uint32_t psp_Machine_arg(psp_Machine* m,uint32_t i){
{
return allegrex_CPU_Reg(m->CPU,cast<uint32_t>((cast<uint32_t>(4ULL) + i)));
}
}
// tools/platform/psp/kernel.go:283:1
void psp_Machine_setRet(psp_Machine* m,uint32_t v){
{
allegrex_CPU_SetReg(m->CPU,cast<uint32_t>(2ULL),v);
}
}
// tools/platform/psp/kernel.go:284:1
void psp_Machine_setRet64(psp_Machine* m,uint64_t v){
{
allegrex_CPU_SetReg(m->CPU,cast<uint32_t>(2ULL),cast<uint32_t>(v));
allegrex_CPU_SetReg(m->CPU,cast<uint32_t>(3ULL),cast<uint32_t>(shr<uint64_t>(v,cast<int64_t>(32ULL))));
}
}
// tools/platform/psp/kernel.go:292:1
uint64_t psp_Machine_clock(psp_Machine* m){
{
return divi<uint64_t>(cast<uint64_t>((m->CPU->Steps * cast<uint64_t>(16667ULL))),cast<uint64_t>(1000000ULL));
}
}
// tools/platform/psp/kernel.go:295:1
uint32_t psp_Machine_newHandle(psp_Machine* m,psp_kobject* o){
{
uint32_t h = m->nextHandle;
m->nextHandle++;
m->handles[h] = o;
return h;
}
}
// tools/platform/psp/kernel.go:305:1
std::string psp_Machine_formatPrintf(psp_Machine* m,std::string format,uint32_t firstArg){
{
uint32_t argn = firstArg;
auto nextArg = [&]()->uint32_t{
uint32_t v={};
if ((argn < cast<uint32_t>(4ULL))) {
v = allegrex_CPU_Reg(m->CPU,cast<uint32_t>((cast<uint32_t>(4ULL) + argn)));
}
else {
v = psp_Machine_read32(m,cast<uint32_t>((allegrex_CPU_Reg(m->CPU,cast<uint32_t>(29ULL)) + cast<uint32_t>((argn * cast<uint32_t>(4ULL))))));
}
argn++;
return v;
}
;
Slice<uint8_t> out={};
{int64_t i = cast<int64_t>(0ULL);for (;(i < len(format));i++){
uint8_t ch = format[i];
if ((ch != cast<uint8_t>(37ULL))) {
out = append(out,ch);
continue;
}
int64_t j = cast<int64_t>((i + cast<int64_t>(1ULL)));
{;for (;((j < len(format)) && (((((((format[j] == cast<uint8_t>(45ULL)) || (format[j] == cast<uint8_t>(48ULL))) || (format[j] == cast<uint8_t>(43ULL))) || (format[j] == cast<uint8_t>(32ULL))) || (format[j] == cast<uint8_t>(46ULL))) || (((format[j] >= cast<uint8_t>(48ULL)) && (format[j] <= cast<uint8_t>(57ULL)))))));){
j++;
}
}if ((j >= len(format))) {
break;
}
{
switch(format[j]){
case cast<uint8_t>(37ULL):{
out = append(out,cast<uint8_t>(37ULL));
break;}
case cast<uint8_t>(115ULL):{
out = append(out,psp_Machine_cstr(m,nextArg()));
break;}
case cast<uint8_t>(100ULL):case cast<uint8_t>(105ULL):{
out = append(out,go_fmt_Sprintf(std::string("%d",2),cast<int32_t>(nextArg())));
break;}
case cast<uint8_t>(117ULL):{
out = append(out,go_fmt_Sprintf(std::string("%d",2),nextArg()));
break;}
case cast<uint8_t>(120ULL):case cast<uint8_t>(88ULL):case cast<uint8_t>(112ULL):{
out = append(out,go_fmt_Sprintf(std::string("%x",2),nextArg()));
break;}
case cast<uint8_t>(99ULL):{
out = append(out,cast<uint8_t>(nextArg()));
break;}
case cast<uint8_t>(102ULL):{
out = append(out,go_fmt_Sprintf(std::string("<%08X>",6),nextArg()));
break;}
default:{
out = append(out,cast<uint8_t>(37ULL),format[j]);
break;}
}}
i = j;
}
}return cast<std::string>(out);
}
}
// tools/platform/psp/kernel.go:356:1
std::string psp_Machine_cstr(psp_Machine* m,uint32_t addr){
{
Slice<uint8_t> b={};
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < cast<uint32_t>(256ULL));i++){
uint8_t ch = psp_Machine_Read(m,cast<uint32_t>((addr + i)));
if ((ch == cast<uint8_t>(0ULL))) {
break;
}
b = append(b,ch);
}
}return cast<std::string>(b);
}
}
// tools/platform/psp/kernel.go:370:1
std::function<void(psp_Machine*)> psp_handlerFor(std::string name){
{
{
auto tmp204=name;
if (tmp204==(std::string("sceKernelAllocPartitionMemory",29))){
return [=](psp_Machine* m)->void{
uint32_t size = psp_Machine_arg(m,cast<uint32_t>(3ULL));
uint32_t addr = m->heapPtr;
m->heapPtr = ((cast<uint32_t>((cast<uint32_t>((m->heapPtr + size)) + cast<uint32_t>(255ULL)))) & ~(cast<uint32_t>(255ULL)));
uint32_t h = psp_Machine_newHandle(m,arenaNew(psp_kobject{std::string("block",5),psp_Machine_cstr(m,psp_Machine_arg(m,cast<uint32_t>(1ULL))),{},addr,{},{},{},{},{},{},{},{},{},{},{},{},{},{},{}}));
psp_Machine_setRet(m,h);
}
;
}
else if (tmp204==(std::string("sceKernelGetBlockHeadAddr",25))){
return [=](psp_Machine* m)->void{
{
psp_kobject* o = get(m->handles,psp_Machine_arg(m,cast<uint32_t>(0ULL)));
if (bool(o)) {
psp_Machine_setRet(m,o->addr);
return ;
}
}
psp_Machine_setRet(m,cast<uint32_t>(0ULL));
}
;
}
else if (tmp204==(std::string("sceKernelMaxFreeMemSize",23)) || tmp204==(std::string("sceKernelTotalFreeMemSize",25))){
return [=](psp_Machine* m)->void{
psp_Machine_setRet(m,cast<uint32_t>((m->heapEnd - m->heapPtr)));
}
;
}
else if (tmp204==(std::string("sceKernelVolatileMemLock",24)) || tmp204==(std::string("sceKernelVolatileMemTryLock",27))){
return [=](psp_Machine* m)->void{
if (m->volatileLocked) {
psp_Machine_note(m,std::string("sceKernelVolatileMemLock while already locked (caller 0x%08X)",61),allegrex_CPU_Reg(m->CPU,cast<uint32_t>(31ULL)));
psp_Machine_setRet(m,cast<uint32_t>(2150302208ULL));
return ;
}
m->volatileLocked = true;
{
uint32_t p = psp_Machine_arg(m,cast<uint32_t>(1ULL));
if ((p != cast<uint32_t>(0ULL))) {
psp_Machine_write32(m,p,cast<uint32_t>(138412032ULL));
}
}
{
uint32_t p = psp_Machine_arg(m,cast<uint32_t>(2ULL));
if ((p != cast<uint32_t>(0ULL))) {
psp_Machine_write32(m,p,cast<uint32_t>(4194304ULL));
}
}
psp_Machine_setRet(m,cast<uint32_t>(0ULL));
}
;
}
else if (tmp204==(std::string("sceKernelVolatileMemUnlock",26))){
return [=](psp_Machine* m)->void{
m->volatileLocked = false;
psp_Machine_setRet(m,cast<uint32_t>(0ULL));
}
;
}
else if (tmp204==(std::string("sceKernelCreateMutex",20)) || tmp204==(std::string("sceKernelCreateMsgPipe",22)) || tmp204==(std::string("sceKernelCreateMbx",18)) || tmp204==(std::string("sceKernelCreateVTimer",21))){
return [=](psp_Machine* m)->void{
psp_Machine_setRet(m,psp_Machine_newHandle(m,arenaNew(psp_kobject{std::string("sync",4),psp_Machine_cstr(m,psp_Machine_arg(m,cast<uint32_t>(0ULL))),{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{}})));
}
;
}
else if (tmp204==(std::string("sceKernelCreateSema",19))){
return [=](psp_Machine* m)->void{
psp_Machine_setRet(m,psp_Machine_newHandle(m,arenaNew(psp_kobject{std::string("sema",4),psp_Machine_cstr(m,psp_Machine_arg(m,cast<uint32_t>(0ULL))),{},{},{},{},{},cast<int32_t>(psp_Machine_arg(m,cast<uint32_t>(2ULL))),{},{},{},{},{},{},{},{},{},{},{}})));
}
;
}
else if (tmp204==(std::string("sceKernelWaitSema",17)) || tmp204==(std::string("sceKernelWaitSemaCB",19))){
return [=](psp_Machine* m)->void{
psp_kobject* o = get(m->handles,psp_Machine_arg(m,cast<uint32_t>(0ULL)));
if (((!o) || (o->kind != std::string("sema",4)))) {
psp_Machine_setRet(m,cast<uint32_t>(2147615171ULL));
return ;
}
int32_t need = cast<int32_t>(psp_Machine_arg(m,cast<uint32_t>(1ULL)));
if (((psp_semaTrace != std::string("",0)) && (o->name == psp_semaTrace))) {
psp_Machine_note(m,std::string("TRACE take %q by %q count=%d need=%d pc=%08X ra=%08X s0=%08X",60),o->name,psp_Machine_CurrentThread(m),o->count,need,m->CPU->PC,allegrex_CPU_Reg(m->CPU,cast<uint32_t>(31ULL)),allegrex_CPU_Reg(m->CPU,cast<uint32_t>(16ULL)));
}
if ((o->count >= need)) {
o->count -= need;
psp_Machine_setRet(m,cast<uint32_t>(0ULL));
return ;
}
if ((!m->current)) {
psp_Machine_note(m,std::string("WaitSema would block the module-start context",45));
psp_Machine_setRet(m,cast<uint32_t>(2147615160ULL));
return ;
}
psp_Machine_setRet(m,cast<uint32_t>(0ULL));
m->current->waitSema = psp_Machine_arg(m,cast<uint32_t>(0ULL));
m->current->waitNeed = need;
psp_Machine_yieldCurrent(m,cast<psp_threadState>(3ULL));
}
;
}
else if (tmp204==(std::string("sceKernelPollSema",17))){
return [=](psp_Machine* m)->void{
psp_kobject* o = get(m->handles,psp_Machine_arg(m,cast<uint32_t>(0ULL)));
if (((!o) || (o->kind != std::string("sema",4)))) {
psp_Machine_setRet(m,cast<uint32_t>(2147615171ULL));
return ;
}
int32_t need = cast<int32_t>(psp_Machine_arg(m,cast<uint32_t>(1ULL)));
if ((o->count >= need)) {
o->count -= need;
psp_Machine_setRet(m,cast<uint32_t>(0ULL));
}
else {
psp_Machine_setRet(m,cast<uint32_t>(2147615149ULL));
}
}
;
}
else if (tmp204==(std::string("sceKernelSignalSema",19))){
return [=](psp_Machine* m)->void{
psp_kobject* o = get(m->handles,psp_Machine_arg(m,cast<uint32_t>(0ULL)));
if (((!o) || (o->kind != std::string("sema",4)))) {
psp_Machine_setRet(m,cast<uint32_t>(2147615171ULL));
return ;
}
o->count += cast<int32_t>(psp_Machine_arg(m,cast<uint32_t>(1ULL)));
if (((psp_semaTrace != std::string("",0)) && (o->name == psp_semaTrace))) {
psp_Machine_note(m,std::string("TRACE give %q by %q count=%d pc=%08X",36),o->name,psp_Machine_CurrentThread(m),o->count,m->CPU->PC);
}
{auto&& tmp205 = m->handles;
for(auto [tmp206,tmp207]:tmp205){
auto th=tmp207;if (((((th->kind == std::string("thread",6)) && (th->tstate == cast<psp_threadState>(3ULL))) && (th->waitSema == psp_Machine_arg(m,cast<uint32_t>(0ULL)))) && (o->count >= th->waitNeed))) {
o->count -= th->waitNeed;
th->waitSema = cast<uint32_t>(0ULL);
th->tstate = cast<psp_threadState>(1ULL);
}
}}
psp_Machine_setRet(m,cast<uint32_t>(0ULL));
}
;
}
else if (tmp204==(std::string("sceKernelDeleteSema",19))){
return [=](psp_Machine* m)->void{
removeKey(m->handles,psp_Machine_arg(m,cast<uint32_t>(0ULL)));
psp_Machine_setRet(m,cast<uint32_t>(0ULL));
}
;
}
else if (tmp204==(std::string("sceKernelCreateEventFlag",24))){
return [=](psp_Machine* m)->void{
psp_Machine_setRet(m,psp_Machine_newHandle(m,arenaNew(psp_kobject{std::string("evflag",6),psp_Machine_cstr(m,psp_Machine_arg(m,cast<uint32_t>(0ULL))),{},{},{},{},psp_Machine_arg(m,cast<uint32_t>(2ULL)),{},{},{},{},{},{},{},{},{},{},{},{}})));
}
;
}
else if (tmp204==(std::string("sceKernelSetEventFlag",21))){
return [=](psp_Machine* m)->void{
psp_kobject* o = get(m->handles,psp_Machine_arg(m,cast<uint32_t>(0ULL)));
if (((!o) || (o->kind != std::string("evflag",6)))) {
psp_Machine_setRet(m,cast<uint32_t>(2147615177ULL));
return ;
}
o->bits |= psp_Machine_arg(m,cast<uint32_t>(1ULL));
{auto&& tmp208 = m->handles;
for(auto [tmp209,tmp210]:tmp208){
auto th=tmp210;if ((((th->kind == std::string("thread",6)) && (th->tstate == cast<psp_threadState>(3ULL))) && (th->waitEv == psp_Machine_arg(m,cast<uint32_t>(0ULL))))) {
if (psp_Machine_evMatch(m,o,th->waitBits,th->waitMode,th->waitOutPtr)) {
th->waitEv = cast<uint32_t>(0ULL);
th->tstate = cast<psp_threadState>(1ULL);
}
}
}}
psp_Machine_setRet(m,cast<uint32_t>(0ULL));
}
;
}
else if (tmp204==(std::string("sceKernelClearEventFlag",23))){
return [=](psp_Machine* m)->void{
{
psp_kobject* o = get(m->handles,psp_Machine_arg(m,cast<uint32_t>(0ULL)));
if ((bool(o) && (o->kind == std::string("evflag",6)))) {
o->bits &= psp_Machine_arg(m,cast<uint32_t>(1ULL));
}
}
psp_Machine_setRet(m,cast<uint32_t>(0ULL));
}
;
}
else if (tmp204==(std::string("sceKernelWaitEventFlag",22)) || tmp204==(std::string("sceKernelWaitEventFlagCB",24))){
return [=](psp_Machine* m)->void{
psp_kobject* o = get(m->handles,psp_Machine_arg(m,cast<uint32_t>(0ULL)));
if (((!o) || (o->kind != std::string("evflag",6)))) {
psp_Machine_setRet(m,cast<uint32_t>(2147615177ULL));
return ;
}
if (psp_Machine_evMatch(m,o,psp_Machine_arg(m,cast<uint32_t>(1ULL)),psp_Machine_arg(m,cast<uint32_t>(2ULL)),psp_Machine_arg(m,cast<uint32_t>(3ULL)))) {
psp_Machine_setRet(m,cast<uint32_t>(0ULL));
return ;
}
if ((!m->current)) {
psp_Machine_note(m,std::string("WaitEventFlag would block the module-start context",50));
psp_Machine_setRet(m,cast<uint32_t>(2147615160ULL));
return ;
}
psp_Machine_setRet(m,cast<uint32_t>(0ULL));
m->current->waitEv = psp_Machine_arg(m,cast<uint32_t>(0ULL));
m->current->waitBits = psp_Machine_arg(m,cast<uint32_t>(1ULL));
m->current->waitMode = psp_Machine_arg(m,cast<uint32_t>(2ULL));
m->current->waitOutPtr = psp_Machine_arg(m,cast<uint32_t>(3ULL));
psp_Machine_yieldCurrent(m,cast<psp_threadState>(3ULL));
}
;
}
else if (tmp204==(std::string("sceKernelPollEventFlag",22))){
return [=](psp_Machine* m)->void{
psp_kobject* o = get(m->handles,psp_Machine_arg(m,cast<uint32_t>(0ULL)));
if (((!o) || (o->kind != std::string("evflag",6)))) {
psp_Machine_setRet(m,cast<uint32_t>(2147615177ULL));
return ;
}
if (psp_Machine_evMatch(m,o,psp_Machine_arg(m,cast<uint32_t>(1ULL)),psp_Machine_arg(m,cast<uint32_t>(2ULL)),psp_Machine_arg(m,cast<uint32_t>(3ULL)))) {
psp_Machine_setRet(m,cast<uint32_t>(0ULL));
}
else {
psp_Machine_setRet(m,cast<uint32_t>(2147615169ULL));
}
}
;
}
else if (tmp204==(std::string("sceKernelDeleteEventFlag",24))){
return [=](psp_Machine* m)->void{
removeKey(m->handles,psp_Machine_arg(m,cast<uint32_t>(0ULL)));
psp_Machine_setRet(m,cast<uint32_t>(0ULL));
}
;
}
else if (tmp204==(std::string("sceAudioChReserve",17))){
return [=](psp_Machine* m)->void{
uint32_t ch = psp_Machine_arg(m,cast<uint32_t>(0ULL));
if ((cast<int32_t>(ch) < cast<int32_t>(0ULL))) {
ch = m->audioCh;
}
m->audioCh = cast<uint32_t>(((cast<uint32_t>((m->audioCh + cast<uint32_t>(1ULL)))) & cast<uint32_t>(7ULL)));
psp_Machine_setRet(m,ch);
}
;
}
else if (tmp204==(std::string("sceAudioOutputBlocking",22)) || tmp204==(std::string("sceAudioOutputPannedBlocking",28)) || tmp204==(std::string("sceAudioSRCOutputBlocking",25)) || tmp204==(std::string("sceAudioOutput2OutputBlocking",29))){
return [=](psp_Machine* m)->void{
psp_Machine_setRet(m,cast<uint32_t>(0ULL));
if (bool(m->current)) {
m->current->wakeVblank = cast<uint32_t>((m->vblanks + cast<uint32_t>(1ULL)));
psp_Machine_yieldCurrent(m,cast<psp_threadState>(3ULL));
}
}
;
}
else if (tmp204==(std::string("sceKernelCreateCallback",23))){
return [=](psp_Machine* m)->void{
psp_Machine_setRet(m,psp_Machine_newHandle(m,arenaNew(psp_kobject{std::string("callback",8),psp_Machine_cstr(m,psp_Machine_arg(m,cast<uint32_t>(0ULL))),psp_Machine_arg(m,cast<uint32_t>(1ULL)),psp_Machine_arg(m,cast<uint32_t>(2ULL)),{},{},{},{},{},{},{},{},{},{},{},{},{},{},{}})));
}
;
}
else if (tmp204==(std::string("sceKernelCreateVpl",18)) || tmp204==(std::string("sceKernelCreateFpl",18))){
return [=](psp_Machine* m)->void{
uint32_t size = psp_Machine_arg(m,cast<uint32_t>(3ULL));
uint32_t addr = m->heapPtr;
m->heapPtr = ((cast<uint32_t>((cast<uint32_t>((m->heapPtr + size)) + cast<uint32_t>(255ULL)))) & ~(cast<uint32_t>(255ULL)));
psp_Machine_setRet(m,psp_Machine_newHandle(m,arenaNew(psp_kobject{std::string("vpl",3),psp_Machine_cstr(m,psp_Machine_arg(m,cast<uint32_t>(0ULL))),{},addr,size,{},{},{},{},{},{},{},{},{},{},{},{},{},{}})));
}
;
}
else if (tmp204==(std::string("sceKernelAllocateVpl",20)) || tmp204==(std::string("sceKernelAllocateVplCB",22)) || tmp204==(std::string("sceKernelTryAllocateVpl",23))){
return [=](psp_Machine* m)->void{
psp_kobject* o = get(m->handles,psp_Machine_arg(m,cast<uint32_t>(0ULL)));
if (((!o) || (o->kind != std::string("vpl",3)))) {
psp_Machine_setRet(m,cast<uint32_t>(2147615175ULL));
return ;
}
uint32_t size = ((cast<uint32_t>((psp_Machine_arg(m,cast<uint32_t>(1ULL)) + cast<uint32_t>(7ULL)))) & ~(cast<uint32_t>(7ULL)));
if ((cast<uint32_t>((o->used + size)) > o->size)) {
psp_Machine_setRet(m,cast<uint32_t>(2147615193ULL));
return ;
}
psp_Machine_write32(m,psp_Machine_arg(m,cast<uint32_t>(2ULL)),cast<uint32_t>((o->addr + o->used)));
o->used += size;
psp_Machine_setRet(m,cast<uint32_t>(0ULL));
}
;
}
else if (tmp204==(std::string("sceKernelFreeVpl",16))){
return [=](psp_Machine* m)->void{
psp_Machine_setRet(m,cast<uint32_t>(0ULL));
}
;
}
else if (tmp204==(std::string("sceKernelReferVplStatus",23))){
return [=](psp_Machine* m)->void{
psp_kobject* o = get(m->handles,psp_Machine_arg(m,cast<uint32_t>(0ULL)));
if (((!o) || (o->kind != std::string("vpl",3)))) {
psp_Machine_setRet(m,cast<uint32_t>(2147615175ULL));
return ;
}
uint32_t p = psp_Machine_arg(m,cast<uint32_t>(1ULL));
psp_Machine_write32(m,p,cast<uint32_t>(52ULL));
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < cast<uint32_t>(32ULL));i++){
uint8_t b={};
if ((cast<int64_t>(i) < len(o->name))) {
b = o->name[i];
}
psp_Machine_Write(m,cast<uint32_t>((cast<uint32_t>((p + cast<uint32_t>(4ULL))) + i)),b);
}
}psp_Machine_write32(m,cast<uint32_t>((p + cast<uint32_t>(36ULL))),cast<uint32_t>(0ULL));
psp_Machine_write32(m,cast<uint32_t>((p + cast<uint32_t>(40ULL))),o->size);
psp_Machine_write32(m,cast<uint32_t>((p + cast<uint32_t>(44ULL))),cast<uint32_t>((o->size - o->used)));
psp_Machine_write32(m,cast<uint32_t>((p + cast<uint32_t>(48ULL))),cast<uint32_t>(0ULL));
psp_Machine_setRet(m,cast<uint32_t>(0ULL));
}
;
}
else if (tmp204==(std::string("sceKernelCreateThread",21))){
return [=](psp_Machine* m)->void{
uint32_t stackSize = psp_Machine_arg(m,cast<uint32_t>(3ULL));
if ((stackSize < cast<uint32_t>(4096ULL))) {
stackSize = cast<uint32_t>(16384ULL);
}
uint32_t base = m->heapPtr;
m->heapPtr = ((cast<uint32_t>((cast<uint32_t>((base + stackSize)) + cast<uint32_t>(255ULL)))) & ~(cast<uint32_t>(255ULL)));
psp_kobject* o = arenaNew(psp_kobject{std::string("thread",6),psp_Machine_cstr(m,psp_Machine_arg(m,cast<uint32_t>(0ULL))),psp_Machine_arg(m,cast<uint32_t>(1ULL)),base,{},{},{},{},psp_Machine_arg(m,cast<uint32_t>(2ULL)),cast<uint32_t>((m->heapPtr - cast<uint32_t>(16ULL))),cast<psp_threadState>(0ULL),{},{},{},{},{},{},{},{}});
m->threadEntry = o->entry;
psp_Machine_setRet(m,psp_Machine_newHandle(m,o));
}
;
}
else if (tmp204==(std::string("sceKernelStartThread",20))){
return [=](psp_Machine* m)->void{
psp_kobject* o = get(m->handles,psp_Machine_arg(m,cast<uint32_t>(0ULL)));
if (((!o) || (o->kind != std::string("thread",6)))) {
psp_Machine_setRet(m,cast<uint32_t>(2147615128ULL));
return ;
}
psp_Machine_startThread(m,psp_Machine_arg(m,cast<uint32_t>(0ULL)),o,psp_Machine_arg(m,cast<uint32_t>(1ULL)),psp_Machine_arg(m,cast<uint32_t>(2ULL)));
psp_Machine_setRet(m,cast<uint32_t>(0ULL));
}
;
}
else if (tmp204==(std::string("sceKernelExitGame",17)) || tmp204==(std::string("sceKernelSelfStopUnloadModule",29))){
return [=](psp_Machine* m)->void{
m->Halted = true;
m->HaltReason = std::string("game requested exit (sceKernelExitGame)",39);
}
;
}
else if (tmp204==(std::string("sceKernelExitThread",19)) || tmp204==(std::string("sceKernelExitDeleteThread",25)) || tmp204==(std::string("sceKernelTerminateDeleteThread",30))){
return [=](psp_Machine* m)->void{
psp_Machine_onThreadExit(m);
}
;
}
else if (tmp204==(std::string("sceKernelRegisterSubIntrHandler",31))){
return [=](psp_Machine* m)->void{
psp_Machine_registerSubIntr(m,psp_Machine_arg(m,cast<uint32_t>(0ULL)),psp_Machine_arg(m,cast<uint32_t>(1ULL)),psp_Machine_arg(m,cast<uint32_t>(2ULL)),psp_Machine_arg(m,cast<uint32_t>(3ULL)));
psp_Machine_setRet(m,cast<uint32_t>(0ULL));
}
;
}
else if (tmp204==(std::string("sceKernelReleaseSubIntrHandler",30))){
return [=](psp_Machine* m)->void{
removeKey(m->subIntrs,cast<uint32_t>((shl<uint32_t>(psp_Machine_arg(m,cast<uint32_t>(0ULL)),cast<int64_t>(16ULL)) | psp_Machine_arg(m,cast<uint32_t>(1ULL)))));
psp_Machine_setRet(m,cast<uint32_t>(0ULL));
}
;
}
else if (tmp204==(std::string("sceKernelEnableSubIntr",22))){
return [=](psp_Machine* m)->void{
psp_Machine_setSubIntrEnabled(m,psp_Machine_arg(m,cast<uint32_t>(0ULL)),psp_Machine_arg(m,cast<uint32_t>(1ULL)),true);
psp_Machine_setRet(m,cast<uint32_t>(0ULL));
}
;
}
else if (tmp204==(std::string("sceKernelDisableSubIntr",23))){
return [=](psp_Machine* m)->void{
psp_Machine_setSubIntrEnabled(m,psp_Machine_arg(m,cast<uint32_t>(0ULL)),psp_Machine_arg(m,cast<uint32_t>(1ULL)),false);
psp_Machine_setRet(m,cast<uint32_t>(0ULL));
}
;
}
else if (tmp204==(std::string("sceDisplayGetVcount",19))){
return [=](psp_Machine* m)->void{
psp_Machine_setRet(m,m->vblanks);
}
;
}
else if (tmp204==(std::string("sceKernelGetSystemTimeWide",26))){
return [=](psp_Machine* m)->void{
psp_Machine_setRet64(m,psp_Machine_clock(m));
}
;
}
else if (tmp204==(std::string("sceKernelGetSystemTimeLow",25))){
return [=](psp_Machine* m)->void{
psp_Machine_setRet(m,cast<uint32_t>(psp_Machine_clock(m)));
}
;
}
else if (tmp204==(std::string("sceKernelGetSystemTime",22)) || tmp204==(std::string("sceRtcGetCurrentTick",20))){
return [=](psp_Machine* m)->void{
uint64_t t = psp_Machine_clock(m);
psp_Machine_write32(m,psp_Machine_arg(m,cast<uint32_t>(0ULL)),cast<uint32_t>(t));
psp_Machine_write32(m,cast<uint32_t>((psp_Machine_arg(m,cast<uint32_t>(0ULL)) + cast<uint32_t>(4ULL))),cast<uint32_t>(shr<uint64_t>(t,cast<int64_t>(32ULL))));
psp_Machine_setRet(m,cast<uint32_t>(0ULL));
}
;
}
else if (tmp204==(std::string("sceCtrlReadBufferPositive",25)) || tmp204==(std::string("sceCtrlPeekBufferPositive",25))){
return [=](psp_Machine* m)->void{
uint32_t p = psp_Machine_arg(m,cast<uint32_t>(0ULL));
psp_Machine_write32(m,cast<uint32_t>((p + cast<uint32_t>(0ULL))),cast<uint32_t>(psp_Machine_clock(m)));
psp_Machine_write32(m,cast<uint32_t>((p + cast<uint32_t>(4ULL))),m->pad);
psp_Machine_Write(m,cast<uint32_t>((p + cast<uint32_t>(8ULL))),rrAnalogX);
psp_Machine_Write(m,cast<uint32_t>((p + cast<uint32_t>(9ULL))),rrAnalogY);
{uint32_t i = cast<uint32_t>(10ULL);for (;(i < cast<uint32_t>(16ULL));i++){
psp_Machine_Write(m,cast<uint32_t>((p + i)),cast<uint8_t>(0ULL));
}
}psp_Machine_setRet(m,cast<uint32_t>(1ULL));
}
;
}
else if (tmp204==(std::string("sceCtrlReadLatch",16)) || tmp204==(std::string("sceCtrlPeekLatch",16))){
return [=](psp_Machine* m)->void{
uint32_t p = psp_Machine_arg(m,cast<uint32_t>(0ULL));
psp_Machine_write32(m,cast<uint32_t>((p + cast<uint32_t>(0ULL))),(m->pad & ~(m->padPrev)));
psp_Machine_write32(m,cast<uint32_t>((p + cast<uint32_t>(4ULL))),(m->padPrev & ~(m->pad)));
psp_Machine_write32(m,cast<uint32_t>((p + cast<uint32_t>(8ULL))),m->pad);
psp_Machine_write32(m,cast<uint32_t>((p + cast<uint32_t>(12ULL))),cast<uint32_t>(~m->pad));
if (go_strings_HasSuffix(name,std::string("ReadLatch",9))) {
m->padPrev = m->pad;
}
psp_Machine_setRet(m,cast<uint32_t>(1ULL));
}
;
}
else if (tmp204==(std::string("sceKernelSleepThread",20)) || tmp204==(std::string("sceKernelSleepThreadCB",22))){
return [=](psp_Machine* m)->void{
psp_Machine_yieldCurrent(m,cast<psp_threadState>(3ULL));
}
;
}
else if (tmp204==(std::string("sceKernelWakeupThread",21))){
return [=](psp_Machine* m)->void{
{
psp_kobject* o = get(m->handles,psp_Machine_arg(m,cast<uint32_t>(0ULL)));
if (((bool(o) && (o->kind == std::string("thread",6))) && (o->tstate == cast<psp_threadState>(3ULL)))) {
o->tstate = cast<psp_threadState>(1ULL);
}
}
psp_Machine_setRet(m,cast<uint32_t>(0ULL));
}
;
}
else if (tmp204==(std::string("sceKernelDelayThread",20)) || tmp204==(std::string("sceKernelDelayThreadCB",22))){
return [=](psp_Machine* m)->void{
uint32_t usec = psp_Machine_arg(m,cast<uint32_t>(0ULL));
psp_Machine_setRet(m,cast<uint32_t>(0ULL));
if ((!m->current)) {
return ;
}
uint32_t frames = divi<uint32_t>(usec,cast<uint32_t>(16667ULL));
if ((frames == cast<uint32_t>(0ULL))) {
frames = cast<uint32_t>(1ULL);
}
m->current->wakeVblank = cast<uint32_t>((m->vblanks + frames));
psp_Machine_yieldCurrent(m,cast<psp_threadState>(3ULL));
}
;
}
else if (tmp204==(std::string("sceDisplayWaitVblank",20)) || tmp204==(std::string("sceDisplayWaitVblankCB",22)) || tmp204==(std::string("sceDisplayWaitVblankStart",25)) || tmp204==(std::string("sceDisplayWaitVblankStartCB",27))){
return [=](psp_Machine* m)->void{
psp_Machine_setRet(m,cast<uint32_t>(0ULL));
if ((!m->current)) {
return ;
}
m->current->wakeVblank = cast<uint32_t>((m->vblanks + cast<uint32_t>(1ULL)));
psp_Machine_yieldCurrent(m,cast<psp_threadState>(3ULL));
}
;
}
else if (tmp204==(std::string("sceDisplaySetFrameBuf",21))){
return [=](psp_Machine* m)->void{
m->fbAddr = psp_Machine_arg(m,cast<uint32_t>(0ULL));
m->fbWidth = psp_Machine_arg(m,cast<uint32_t>(1ULL));
m->fbFormat = psp_Machine_arg(m,cast<uint32_t>(2ULL));
psp_Machine_setRet(m,cast<uint32_t>(0ULL));
psp_Machine_profFrame(m);
if (bool(m->OnDisplay)) {
m->OnDisplay(m);
}
}
;
}
else if (tmp204==(std::string("sceDisplayGetFrameBuf",21))){
return [=](psp_Machine* m)->void{
psp_Machine_write32(m,psp_Machine_arg(m,cast<uint32_t>(0ULL)),m->fbAddr);
psp_Machine_write32(m,psp_Machine_arg(m,cast<uint32_t>(1ULL)),m->fbWidth);
psp_Machine_write32(m,psp_Machine_arg(m,cast<uint32_t>(2ULL)),m->fbFormat);
psp_Machine_setRet(m,cast<uint32_t>(0ULL));
}
;
}
else if (tmp204==(std::string("sceGeListEnQueue",16)) || tmp204==(std::string("sceGeListEnQueueHead",20))){
return [=](psp_Machine* m)->void{
time_Time tl = psp_Machine_profStart(m);
psp_GeList list = psp_Machine_captureList(m,psp_Machine_arg(m,cast<uint32_t>(0ULL)));
psp_Machine_profEnd(m,cast<int64_t>(0ULL),tl);
m->geCnt.lists++;
m->GeLists = append(m->GeLists,list);
if (bool(m->OnGeList)) {
m->OnGeList(list);
}
psp_Machine_execGeList(m,list);
psp_Machine_setRet(m,cast<uint32_t>(len(m->GeLists)));
}
;
}
else if (tmp204==(std::string("sceGeEdramGetAddr",17))){
return [=](psp_Machine* m)->void{
psp_Machine_setRet(m,cast<uint32_t>(67108864ULL));
}
;
}
else if (tmp204==(std::string("sceGeEdramGetSize",17))){
return [=](psp_Machine* m)->void{
psp_Machine_setRet(m,cast<uint32_t>(2097152ULL));
}
;
}
else if (tmp204==(std::string("sceKernelGetThreadId",20))){
return [=](psp_Machine* m)->void{
psp_Machine_setRet(m,psp_Machine_currentThreadID(m));
}
;
}
else if (tmp204==(std::string("sceKernelGetModuleIdByAddress",29)) || tmp204==(std::string("sceKernelGetModuleId",20))){
return [=](psp_Machine* m)->void{
psp_Machine_setRet(m,cast<uint32_t>(1ULL));
}
;
}
else if (tmp204==(std::string("sceKernelGetGPI",15))){
return [=](psp_Machine* m)->void{
psp_Machine_setRet(m,cast<uint32_t>(1ULL));
}
;
}
else if (tmp204==(std::string("sceKernelPrintf",15))){
return [=](psp_Machine* m)->void{
m->tty = append(m->tty,psp_Machine_formatPrintf(m,psp_Machine_cstr(m,psp_Machine_arg(m,cast<uint32_t>(0ULL))),cast<uint32_t>(1ULL)));
psp_Machine_setRet(m,cast<uint32_t>(0ULL));
}
;
}
else if (tmp204==(std::string("sceKernelStdin",14))){
return [=](psp_Machine* m)->void{
psp_Machine_setRet(m,cast<uint32_t>(0ULL));
}
;
}
else if (tmp204==(std::string("sceKernelStdout",15))){
return [=](psp_Machine* m)->void{
psp_Machine_setRet(m,cast<uint32_t>(1ULL));
}
;
}
else if (tmp204==(std::string("sceKernelStderr",15))){
return [=](psp_Machine* m)->void{
psp_Machine_setRet(m,cast<uint32_t>(2ULL));
}
;
}
else if (tmp204==(std::string("sceIoOpen",9))){
return [=](psp_Machine* m)->void{
psp_Machine_setRet(m,psp_Machine_ioOpen(m,psp_Machine_cstr(m,psp_Machine_arg(m,cast<uint32_t>(0ULL)))));
}
;
}
else if (tmp204==(std::string("sceIoClose",10))){
return [=](psp_Machine* m)->void{
psp_Machine_setRet(m,psp_Machine_ioClose(m,psp_Machine_arg(m,cast<uint32_t>(0ULL))));
}
;
}
else if (tmp204==(std::string("sceIoRead",9))){
return [=](psp_Machine* m)->void{
psp_Machine_setRet(m,psp_Machine_ioRead(m,psp_Machine_arg(m,cast<uint32_t>(0ULL)),psp_Machine_arg(m,cast<uint32_t>(1ULL)),psp_Machine_arg(m,cast<uint32_t>(2ULL))));
}
;
}
else if (tmp204==(std::string("sceIoWrite",10))){
return [=](psp_Machine* m)->void{
psp_Machine_setRet(m,psp_Machine_ioWrite(m,psp_Machine_arg(m,cast<uint32_t>(0ULL)),psp_Machine_arg(m,cast<uint32_t>(1ULL)),psp_Machine_arg(m,cast<uint32_t>(2ULL))));
}
;
}
else if (tmp204==(std::string("sceIoLseek",10))){
return [=](psp_Machine* m)->void{
int64_t off = cast<int64_t>(cast<uint64_t>((shl<uint64_t>(cast<uint64_t>(psp_Machine_arg(m,cast<uint32_t>(3ULL))),cast<int64_t>(32ULL)) | cast<uint64_t>(psp_Machine_arg(m,cast<uint32_t>(2ULL))))));
psp_Machine_setRet64(m,cast<uint64_t>(psp_Machine_ioLseek(m,psp_Machine_arg(m,cast<uint32_t>(0ULL)),off,psp_Machine_arg(m,cast<uint32_t>(4ULL)))));
}
;
}
else if (tmp204==(std::string("sceIoGetstat",12))){
return [=](psp_Machine* m)->void{
psp_Machine_setRet(m,psp_Machine_ioGetstat(m,psp_Machine_cstr(m,psp_Machine_arg(m,cast<uint32_t>(0ULL))),psp_Machine_arg(m,cast<uint32_t>(1ULL))));
}
;
}
else if (tmp204==(std::string("sceIoDopen",10))){
return [=](psp_Machine* m)->void{
psp_Machine_setRet(m,psp_Machine_ioDopen(m,psp_Machine_cstr(m,psp_Machine_arg(m,cast<uint32_t>(0ULL)))));
}
;
}
else if (tmp204==(std::string("sceIoDread",10))){
return [=](psp_Machine* m)->void{
psp_Machine_setRet(m,psp_Machine_ioDread(m,psp_Machine_arg(m,cast<uint32_t>(0ULL)),psp_Machine_arg(m,cast<uint32_t>(1ULL))));
}
;
}
else if (tmp204==(std::string("sceIoDclose",11))){
return [=](psp_Machine* m)->void{
psp_Machine_setRet(m,psp_Machine_ioDclose(m,psp_Machine_arg(m,cast<uint32_t>(0ULL))));
}
;
}
else if (tmp204==(std::string("sceIoOpenAsync",14))){
return [=](psp_Machine* m)->void{
psp_Machine_setRet(m,psp_Machine_ioOpenAsync(m,psp_Machine_cstr(m,psp_Machine_arg(m,cast<uint32_t>(0ULL)))));
}
;
}
else if (tmp204==(std::string("sceIoReadAsync",14))){
return [=](psp_Machine* m)->void{
psp_Machine_setRet(m,psp_Machine_ioReadAsync(m,psp_Machine_arg(m,cast<uint32_t>(0ULL)),psp_Machine_arg(m,cast<uint32_t>(1ULL)),psp_Machine_arg(m,cast<uint32_t>(2ULL))));
}
;
}
else if (tmp204==(std::string("sceIoCloseAsync",15))){
return [=](psp_Machine* m)->void{
psp_Machine_setRet(m,psp_Machine_ioCloseAsync(m,psp_Machine_arg(m,cast<uint32_t>(0ULL))));
}
;
}
else if (tmp204==(std::string("sceIoLseekAsync",15))){
return [=](psp_Machine* m)->void{
int64_t off = cast<int64_t>(cast<uint64_t>((shl<uint64_t>(cast<uint64_t>(psp_Machine_arg(m,cast<uint32_t>(3ULL))),cast<int64_t>(32ULL)) | cast<uint64_t>(psp_Machine_arg(m,cast<uint32_t>(2ULL))))));
psp_Machine_setRet(m,psp_Machine_ioLseekAsync(m,psp_Machine_arg(m,cast<uint32_t>(0ULL)),off,psp_Machine_arg(m,cast<uint32_t>(4ULL))));
}
;
}
else if (tmp204==(std::string("sceIoLseek32Async",17))){
return [=](psp_Machine* m)->void{
psp_Machine_setRet(m,psp_Machine_ioLseekAsync(m,psp_Machine_arg(m,cast<uint32_t>(0ULL)),cast<int64_t>(cast<int32_t>(psp_Machine_arg(m,cast<uint32_t>(1ULL)))),psp_Machine_arg(m,cast<uint32_t>(2ULL))));
}
;
}
else if (tmp204==(std::string("sceIoWaitAsync",14)) || tmp204==(std::string("sceIoWaitAsyncCB",16)) || tmp204==(std::string("sceIoGetAsyncStat",17))){
return [=](psp_Machine* m)->void{
psp_Machine_setRet(m,psp_Machine_ioWaitAsync(m,psp_Machine_arg(m,cast<uint32_t>(0ULL)),psp_Machine_arg(m,cast<uint32_t>(1ULL))));
}
;
}
else if (tmp204==(std::string("sceIoPollAsync",14))){
return [=](psp_Machine* m)->void{
psp_Machine_setRet(m,psp_Machine_ioPollAsync(m,psp_Machine_arg(m,cast<uint32_t>(0ULL)),psp_Machine_arg(m,cast<uint32_t>(1ULL))));
}
;
}
else if (tmp204==(std::string("sceIoChangeAsyncPriority",24)) || tmp204==(std::string("sceIoSetAsyncCallback",21))){
return [=](psp_Machine* m)->void{
psp_Machine_setRet(m,cast<uint32_t>(0ULL));
}
;
}
else if (tmp204==(std::string("sceIoLseek32",12))){
return [=](psp_Machine* m)->void{
psp_Machine_setRet(m,cast<uint32_t>(psp_Machine_ioLseek(m,psp_Machine_arg(m,cast<uint32_t>(0ULL)),cast<int64_t>(cast<int32_t>(psp_Machine_arg(m,cast<uint32_t>(1ULL)))),psp_Machine_arg(m,cast<uint32_t>(2ULL)))));
}
;
}
else if (tmp204==(std::string("sceUmdActivate",14)) || tmp204==(std::string("sceUmdDeactivate",16)) || tmp204==(std::string("sceUmdWaitDriveStat",19)) || tmp204==(std::string("sceUmdWaitDriveStatCB",21)) || tmp204==(std::string("sceUmdWaitDriveStatWithTimer",28))){
return [=](psp_Machine* m)->void{
psp_Machine_setRet(m,cast<uint32_t>(0ULL));
}
;
}
else if (tmp204==(std::string("sceMpegQueryMemSize",19))){
return [=](psp_Machine* m)->void{
psp_Machine_setRet(m,cast<uint32_t>(65536ULL));
}
;
}
else if (tmp204==(std::string("sceMpegRingbufferQueryMemSize",29))){
return [=](psp_Machine* m)->void{
psp_Machine_setRet(m,cast<uint32_t>((psp_Machine_arg(m,cast<uint32_t>(0ULL)) * cast<uint32_t>(2152ULL))));
}
;
}
else if (tmp204==(std::string("sceMpegCreate",13))){
return [=](psp_Machine* m)->void{
psp_Machine_write32(m,psp_Machine_arg(m,cast<uint32_t>(0ULL)),psp_Machine_arg(m,cast<uint32_t>(1ULL)));
m->mpeg.Handle = psp_Machine_arg(m,cast<uint32_t>(1ULL));
{
uint32_t rb = psp_Machine_arg(m,cast<uint32_t>(3ULL));
if ((rb != cast<uint32_t>(0ULL))) {
psp_Machine_write32(m,cast<uint32_t>((rb + cast<uint32_t>(40ULL))),psp_Machine_arg(m,cast<uint32_t>(1ULL)));
}
}
psp_Machine_note(m,std::string("sceMpegCreate: handle 0x%08X, ringbuffer 0x%08X",47),psp_Machine_arg(m,cast<uint32_t>(1ULL)),psp_Machine_arg(m,cast<uint32_t>(3ULL)));
psp_Machine_setRet(m,cast<uint32_t>(0ULL));
}
;
}
else if (tmp204==(std::string("sceMpegDelete",13)) || tmp204==(std::string("sceMpegFinish",13)) || tmp204==(std::string("sceMpegRingbufferDestruct",25))){
return [=](psp_Machine* m)->void{
m->mpeg = psp_mpegState{};
psp_Machine_setRet(m,cast<uint32_t>(0ULL));
}
;
}
else if (tmp204==(std::string("sceMpegRingbufferConstruct",26))){
return [=](psp_Machine* m)->void{
psp_Machine_setRet(m,psp_Machine_mpegRingbufferConstruct(m,psp_Machine_arg(m,cast<uint32_t>(0ULL)),psp_Machine_arg(m,cast<uint32_t>(1ULL)),psp_Machine_arg(m,cast<uint32_t>(2ULL)),psp_Machine_arg(m,cast<uint32_t>(3ULL)),psp_Machine_arg(m,cast<uint32_t>(4ULL)),psp_Machine_arg(m,cast<uint32_t>(5ULL))));
}
;
}
else if (tmp204==(std::string("sceMpegRingbufferAvailableSize",30))){
return [=](psp_Machine* m)->void{
psp_Machine_setRet(m,cast<uint32_t>((m->mpeg.Packets - m->mpeg.In)));
}
;
}
else if (tmp204==(std::string("sceMpegRingbufferPut",20))){
return [=](psp_Machine* m)->void{
psp_Machine_setRet(m,psp_Machine_mpegRingbufferPut(m,psp_Machine_arg(m,cast<uint32_t>(0ULL)),psp_Machine_arg(m,cast<uint32_t>(1ULL)),psp_Machine_arg(m,cast<uint32_t>(2ULL))));
}
;
}
else if (tmp204==(std::string("sceMpegQueryStreamOffset",24))){
return [=](psp_Machine* m)->void{
auto tmp211 = std::make_tuple(psp_Machine_arg(m,cast<uint32_t>(1ULL)),psp_Machine_arg(m,cast<uint32_t>(2ULL)));
uint32_t buf = std::get<0>(tmp211);
uint32_t out = std::get<1>(tmp211);
if (((((psp_Machine_Read(m,buf) != cast<uint8_t>(80ULL)) || (psp_Machine_Read(m,cast<uint32_t>((buf + cast<uint32_t>(1ULL)))) != cast<uint8_t>(83ULL))) || (psp_Machine_Read(m,cast<uint32_t>((buf + cast<uint32_t>(2ULL)))) != cast<uint8_t>(77ULL))) || (psp_Machine_Read(m,cast<uint32_t>((buf + cast<uint32_t>(3ULL)))) != cast<uint8_t>(70ULL)))) {
psp_Machine_note(m,std::string("sceMpegQueryStreamOffset: no PSMF magic at 0x%08X",49),buf);
psp_Machine_write32(m,out,cast<uint32_t>(0ULL));
psp_Machine_setRet(m,cast<uint32_t>(2153841150ULL));
return ;
}
psp_Machine_write32(m,out,psp_Machine_beGuest32(m,cast<uint32_t>((buf + cast<uint32_t>(8ULL)))));
psp_Machine_setRet(m,cast<uint32_t>(0ULL));
}
;
}
else if (tmp204==(std::string("sceMpegQueryStreamSize",22))){
return [=](psp_Machine* m)->void{
psp_Machine_write32(m,psp_Machine_arg(m,cast<uint32_t>(1ULL)),psp_Machine_beGuest32(m,cast<uint32_t>((psp_Machine_arg(m,cast<uint32_t>(0ULL)) + cast<uint32_t>(12ULL)))));
psp_Machine_setRet(m,cast<uint32_t>(0ULL));
}
;
}
else if (tmp204==(std::string("sceMpegAvcQueryYCbCrSize",24))){
return [=](psp_Machine* m)->void{
auto tmp212 = std::make_tuple(psp_Machine_arg(m,cast<uint32_t>(2ULL)),psp_Machine_arg(m,cast<uint32_t>(3ULL)));
uint32_t w = std::get<0>(tmp212);
uint32_t h = std::get<1>(tmp212);
psp_Machine_write32(m,psp_Machine_arg(m,cast<uint32_t>(4ULL)),divi<uint32_t>(cast<uint32_t>((cast<uint32_t>((w * h)) * cast<uint32_t>(3ULL))),cast<uint32_t>(2ULL)));
psp_Machine_setRet(m,cast<uint32_t>(0ULL));
}
;
}
else if (tmp204==(std::string("sceMpegRegistStream",19))){
return [=](psp_Machine* m)->void{
psp_Machine_setRet(m,cast<uint32_t>((cast<uint32_t>(8192ULL) + psp_Machine_arg(m,cast<uint32_t>(1ULL)))));
}
;
}
else if (tmp204==(std::string("sceMpegMallocAvcEsBuf",21))){
return [=](psp_Machine* m)->void{
psp_Machine_setRet(m,cast<uint32_t>(1ULL));
}
;
}
else if (tmp204==(std::string("sceMpegInitAu",13))){
return [=](psp_Machine* m)->void{
uint32_t au = psp_Machine_arg(m,cast<uint32_t>(2ULL));
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < cast<uint32_t>(16ULL));i += cast<uint32_t>(4ULL)){
psp_Machine_write32(m,cast<uint32_t>((au + i)),cast<uint32_t>(4294967295ULL));
}
}psp_Machine_write32(m,cast<uint32_t>((au + cast<uint32_t>(16ULL))),psp_Machine_arg(m,cast<uint32_t>(1ULL)));
psp_Machine_write32(m,cast<uint32_t>((au + cast<uint32_t>(20ULL))),cast<uint32_t>(0ULL));
psp_Machine_setRet(m,cast<uint32_t>(0ULL));
}
;
}
else if (tmp204==(std::string("sceMpegGetAvcAu",15)) || tmp204==(std::string("sceMpegGetAtracAu",17)) || tmp204==(std::string("sceMpegGetPcmAu",15))){
return [=](psp_Machine* m)->void{
uint32_t r = psp_Machine_mpegGetAu(m,psp_Machine_arg(m,cast<uint32_t>(2ULL)));
if (((r == cast<uint32_t>(0ULL)) && (psp_Machine_arg(m,cast<uint32_t>(3ULL)) != cast<uint32_t>(0ULL)))) {
psp_Machine_write32(m,psp_Machine_arg(m,cast<uint32_t>(3ULL)),cast<uint32_t>(0ULL));
}
psp_Machine_setRet(m,r);
}
;
}
else if (tmp204==(std::string("sceMpegAvcDecodeYCbCr",21)) || tmp204==(std::string("sceMpegAtracDecode",18))){
return [=](psp_Machine* m)->void{
{
uint32_t p = psp_Machine_arg(m,cast<uint32_t>(3ULL));
if ((p != cast<uint32_t>(0ULL))) {
psp_Machine_write32(m,p,cast<uint32_t>(1ULL));
}
}
psp_Machine_setRet(m,cast<uint32_t>(0ULL));
}
;
}
else if (tmp204==(std::string("sceMpegAvcDecode",16))){
return [=](psp_Machine* m)->void{
{
uint32_t p = psp_Machine_arg(m,cast<uint32_t>(4ULL));
if ((p != cast<uint32_t>(0ULL))) {
psp_Machine_write32(m,p,cast<uint32_t>(1ULL));
}
}
psp_Machine_setRet(m,cast<uint32_t>(0ULL));
}
;
}
else if (tmp204==(std::string("sceMpegFlushStream",18)) || tmp204==(std::string("sceMpegFlushAllStream",21))){
return [=](psp_Machine* m)->void{
m->mpeg.In = cast<uint32_t>(0ULL);
{
uint32_t rb = m->mpeg.Ringbuf;
if ((rb != cast<uint32_t>(0ULL))) {
psp_Machine_write32(m,cast<uint32_t>((rb + cast<uint32_t>(12ULL))),m->mpeg.Packets);
}
}
psp_Machine_setRet(m,cast<uint32_t>(0ULL));
}
;
}
else if (tmp204==(std::string("sceAtracSetDataAndGetID",23)) || tmp204==(std::string("sceAtracSetHalfwayBufferAndGetID",32))){
return [=](psp_Machine* m)->void{
uint32_t id = m->nextAtrac;
m->nextAtrac++;
psp_atracState* a = arenaNew(psp_atracState{psp_Machine_arg(m,cast<uint32_t>(0ULL)),psp_Machine_arg(m,cast<uint32_t>(1ULL)),{},{},{}});
psp_Machine_atracParseRiff(m,a);
m->atrac[id] = a;
psp_Machine_note(m,std::string("sceAtracSetDataAndGetID(0x%08X, %d) -> id %d (%d frames, %d ch)",63),a->Buf,a->Size,id,a->Frames,a->Channels);
psp_Machine_setRet(m,id);
}
;
}
else if (tmp204==(std::string("sceAtracDecodeData",18))){
return [=](psp_Machine* m)->void{
psp_Machine_setRet(m,psp_Machine_atracDecode(m,psp_Machine_arg(m,cast<uint32_t>(0ULL)),psp_Machine_arg(m,cast<uint32_t>(1ULL)),psp_Machine_arg(m,cast<uint32_t>(2ULL)),psp_Machine_arg(m,cast<uint32_t>(3ULL)),psp_Machine_arg(m,cast<uint32_t>(4ULL))));
}
;
}
else if (tmp204==(std::string("sceAtracGetRemainFrame",22))){
return [=](psp_Machine* m)->void{
{
psp_atracState* a = get(m->atrac,psp_Machine_arg(m,cast<uint32_t>(0ULL)));
if (bool(a)) {
psp_Machine_write32(m,psp_Machine_arg(m,cast<uint32_t>(1ULL)),cast<uint32_t>((a->Frames - a->Pos)));
psp_Machine_setRet(m,cast<uint32_t>(0ULL));
}
else {
psp_Machine_setRet(m,cast<uint32_t>(2153971716ULL));
}
}
}
;
}
else if (tmp204==(std::string("sceAtracGetStreamDataInfo",25))){
return [=](psp_Machine* m)->void{
psp_atracState* a = get(m->atrac,psp_Machine_arg(m,cast<uint32_t>(0ULL)));
if ((!a)) {
psp_Machine_setRet(m,cast<uint32_t>(2153971716ULL));
return ;
}
{
uint32_t p = psp_Machine_arg(m,cast<uint32_t>(1ULL));
if ((p != cast<uint32_t>(0ULL))) {
psp_Machine_write32(m,p,a->Buf);
}
}
{
uint32_t p = psp_Machine_arg(m,cast<uint32_t>(2ULL));
if ((p != cast<uint32_t>(0ULL))) {
psp_Machine_write32(m,p,cast<uint32_t>(0ULL));
}
}
{
uint32_t p = psp_Machine_arg(m,cast<uint32_t>(3ULL));
if ((p != cast<uint32_t>(0ULL))) {
psp_Machine_write32(m,p,a->Size);
}
}
psp_Machine_setRet(m,cast<uint32_t>(0ULL));
}
;
}
else if (tmp204==(std::string("sceAtracGetNextSample",21)) || tmp204==(std::string("sceAtracGetMaxSample",20))){
return [=](psp_Machine* m)->void{
{
psp_atracState* a = get(m->atrac,psp_Machine_arg(m,cast<uint32_t>(0ULL)));
if (bool(a)) {
uint32_t n = cast<uint32_t>(2048ULL);
if ((a->Pos >= a->Frames)) {
n = cast<uint32_t>(0ULL);
}
psp_Machine_write32(m,psp_Machine_arg(m,cast<uint32_t>(1ULL)),n);
psp_Machine_setRet(m,cast<uint32_t>(0ULL));
}
else {
psp_Machine_setRet(m,cast<uint32_t>(2153971716ULL));
}
}
}
;
}
else if (tmp204==(std::string("sceAtracGetChannel",18))){
return [=](psp_Machine* m)->void{
{
psp_atracState* a = get(m->atrac,psp_Machine_arg(m,cast<uint32_t>(0ULL)));
if (bool(a)) {
psp_Machine_write32(m,psp_Machine_arg(m,cast<uint32_t>(1ULL)),a->Channels);
psp_Machine_setRet(m,cast<uint32_t>(0ULL));
}
else {
psp_Machine_setRet(m,cast<uint32_t>(2153971716ULL));
}
}
}
;
}
else if (tmp204==(std::string("sceAtracGetNextDecodePosition",29))){
return [=](psp_Machine* m)->void{
psp_atracState* a = get(m->atrac,psp_Machine_arg(m,cast<uint32_t>(0ULL)));
if ((!a)) {
psp_Machine_setRet(m,cast<uint32_t>(2153971716ULL));
return ;
}
if ((a->Pos >= a->Frames)) {
psp_Machine_setRet(m,cast<uint32_t>(2153971714ULL));
return ;
}
psp_Machine_write32(m,psp_Machine_arg(m,cast<uint32_t>(1ULL)),cast<uint32_t>((a->Pos * cast<uint32_t>(2048ULL))));
psp_Machine_setRet(m,cast<uint32_t>(0ULL));
}
;
}
else if (tmp204==(std::string("sceAtracReleaseAtracID",22))){
return [=](psp_Machine* m)->void{
removeKey(m->atrac,psp_Machine_arg(m,cast<uint32_t>(0ULL)));
psp_Machine_setRet(m,cast<uint32_t>(0ULL));
}
;
}
else if (tmp204==(std::string("sceUmdCheckMedium",17))){
return [=](psp_Machine* m)->void{
psp_Machine_setRet(m,cast<uint32_t>(1ULL));
}
;
}
else if (tmp204==(std::string("sceUtilitySavedataInitStart",27))){
return [=](psp_Machine* m)->void{
uint32_t p = psp_Machine_arg(m,cast<uint32_t>(0ULL));
uint32_t mode = psp_Machine_read32(m,cast<uint32_t>((p + cast<uint32_t>(48ULL))));
uint32_t result={};
{
switch(mode){
case cast<uint32_t>(0ULL):case cast<uint32_t>(2ULL):case cast<uint32_t>(4ULL):{
result = cast<uint32_t>(2148598535ULL);
break;}
}}
psp_Machine_write32(m,cast<uint32_t>((p + cast<uint32_t>(28ULL))),result);
m->savedataStatus = cast<uint32_t>(1ULL);
psp_Machine_setRet(m,cast<uint32_t>(0ULL));
}
;
}
else if (tmp204==(std::string("sceUtilitySavedataGetStatus",27))){
return [=](psp_Machine* m)->void{
uint32_t st = m->savedataStatus;
psp_Machine_setRet(m,st);
{
switch(st){
case cast<uint32_t>(1ULL):case cast<uint32_t>(2ULL):{
m->savedataStatus = cast<uint32_t>((st + cast<uint32_t>(1ULL)));
break;}
case cast<uint32_t>(4ULL):{
m->savedataStatus = cast<uint32_t>(0ULL);
break;}
}}
}
;
}
else if (tmp204==(std::string("sceUtilitySavedataUpdate",24))){
return [=](psp_Machine* m)->void{
psp_Machine_setRet(m,cast<uint32_t>(0ULL));
}
;
}
else if (tmp204==(std::string("sceUtilitySavedataShutdownStart",31))){
return [=](psp_Machine* m)->void{
m->savedataStatus = cast<uint32_t>(4ULL);
psp_Machine_setRet(m,cast<uint32_t>(0ULL));
}
;
}
else if (tmp204==(std::string("sceUtilityGetSystemParamInt",27))){
return [=](psp_Machine* m)->void{
uint32_t v={};
{
switch(psp_Machine_arg(m,cast<uint32_t>(0ULL))){
case cast<uint32_t>(8ULL):case cast<uint32_t>(9ULL):{
v = cast<uint32_t>(1ULL);
break;}
}}
psp_Machine_write32(m,psp_Machine_arg(m,cast<uint32_t>(1ULL)),v);
psp_Machine_setRet(m,cast<uint32_t>(0ULL));
}
;
}
else if (tmp204==(std::string("sceUmdGetDriveStat",18))){
return [=](psp_Machine* m)->void{
psp_Machine_setRet(m,cast<uint32_t>(50ULL));
}
;
}
else if (tmp204==(std::string("sceKernelDcacheWritebackAll",27)) || tmp204==(std::string("sceKernelDcacheWritebackRange",29)) || tmp204==(std::string("sceKernelDcacheWritebackInvalidateAll",37)) || tmp204==(std::string("sceKernelIcacheInvalidateAll",28)) || tmp204==(std::string("sceDisplaySetMode",17))){
return [=](psp_Machine* m)->void{
psp_Machine_setRet(m,cast<uint32_t>(0ULL));
}
;
}
}
tmp203:;
return {};
}
}
// tools/platform/psp/kirk.go:46:1
Error psp_kirk7(Slice<uint8_t> dst,Slice<uint8_t> src,int64_t keyId){
{
if (((keyId < cast<int64_t>(0ULL)) || (keyId >= cast<int64_t>(128ULL)))) {
return go_fmt_Errorf(std::string("kirk: key id 0x%X out of range",30),keyId);
}
auto tmp213 = psp_cbcDecryptZero(sub(psp_keyvault[keyId],0,len(psp_keyvault[keyId])),src);
Slice<uint8_t> out = std::get<0>(tmp213);
Error err = std::get<1>(tmp213);
if (bool(err)) {
return err;
}
gcopy(dst,out);
return {};
}
}
// tools/platform/psp/kirk.go:69:1
std::tuple<Slice<uint8_t>,Error> psp_kirkCMD1(Slice<uint8_t> in){
{
if ((len(in) < cast<int64_t>(144ULL))) {
return {{},go_fmt_Errorf(std::string("kirk: CMD1 input too small (%d)",31),len(in))};
}
{
uint32_t mode = le_Uint32(sub(in,cast<int64_t>(96ULL),len(in)));
if ((mode != cast<uint32_t>(1ULL))) {
return {{},go_fmt_Errorf(std::string("kirk: CMD1 mode %d != 1 (header decrypt likely wrong)",53),mode)};
}
}
auto tmp214 = psp_cbcDecryptZero(sub(psp_kirk1Key,0,len(psp_kirk1Key)),sub(in,cast<int64_t>(0ULL),cast<int64_t>(32ULL)));
Slice<uint8_t> keys = std::get<0>(tmp214);
Error err = std::get<1>(tmp214);
if (bool(err)) {
return {{},err};
}
Slice<uint8_t> aesKey = sub(keys,cast<int64_t>(0ULL),cast<int64_t>(16ULL));
int64_t dataSize = cast<int64_t>(le_Uint32(sub(in,cast<int64_t>(112ULL),len(in))));
int64_t dataOffset = cast<int64_t>(le_Uint32(sub(in,cast<int64_t>(116ULL),len(in))));
int64_t start = cast<int64_t>((cast<int64_t>(144ULL) + dataOffset));
if ((((start < cast<int64_t>(0ULL)) || (dataSize < cast<int64_t>(0ULL))) || (cast<int64_t>((start + dataSize)) > len(in)))) {
return {{},go_fmt_Errorf(std::string("kirk: CMD1 body [%#x:+%#x] out of range (%d)",44),start,dataSize,len(in))};
}
return psp_cbcDecryptZero(aesKey,sub(in,start,cast<int64_t>((start + dataSize))));
}
}
// tools/platform/psp/machine.go:38:1
uint32_t psp_phys(uint32_t addr){
{
return cast<uint32_t>((addr & cast<uint32_t>(536870911ULL)));
}
}
// tools/platform/psp/machine.go:125:1
psp_Machine* psp_NewMachine(){
{
psp_Machine* m = arenaNew(psp_Machine{Slice<uint8_t>::make(cast<int64_t>(33554432ULL)),Slice<uint8_t>::make(cast<int64_t>(2097152ULL)),Slice<uint8_t>::make(cast<int64_t>(16384ULL)),{},Map<uint32_t,uint32_t>{},Map<uint32_t,psp_syscall*>{},cast<uint32_t>(4096ULL),Map<uint32_t,psp_kobject*>{},cast<uint32_t>(1ULL),{},{},{},{},{},Map<uint32_t,psp_subIntr*>{},{},{},{},{},{},{},{},Map<uint32_t,psp_atracState*>{},{},{},Map<uint32_t,psp_ioFile*>{},cast<uint32_t>(4ULL),{},Map<std::string,int64_t>{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},Map<std::string,bool>{},{},{}});
m->CPU = allegrex_NewCPU(m);
m->CPU->Syscall = [=](auto...args){return psp_Machine_handleSyscall(m,args...);};
return m;
}
}
// tools/platform/psp/machine.go:155:1
void psp_Machine_SetPadScript(psp_Machine* m,Slice<psp_PadEvent> evs){
{
m->padScript = evs;
}
}
// tools/platform/psp/machine.go:158:1
uint32_t psp_Machine_Vblanks(psp_Machine* m){
{
return m->vblanks;
}
}
// tools/platform/psp/machine.go:173:1
uint8_t psp_Machine_Read(psp_Machine* m,uint32_t addr){
{if(!m->OnRead)if(auto*p=rrMemory(m,addr,1))return *p;
{
uint32_t p = psp_phys(addr);
uint8_t v={};
{
if (((p >= cast<uint32_t>(134217728ULL)) && (p < cast<uint32_t>(167772160ULL)))){
v = m->ram[cast<uint32_t>((p - cast<uint32_t>(134217728ULL)))];
}
else if (((p >= cast<uint32_t>(67108864ULL)) && (p < cast<uint32_t>(69206016ULL)))){
v = m->vram[cast<uint32_t>((p - cast<uint32_t>(67108864ULL)))];
}
else if (((p >= cast<uint32_t>(65536ULL)) && (p < cast<uint32_t>(81920ULL)))){
v = m->scratch[cast<uint32_t>((p - cast<uint32_t>(65536ULL)))];
}
else {
psp_Machine_note(m,std::string("read from unmapped 0x%08X at pc 0x%08X",38),addr,allegrex_CPU_CurPC(m->CPU));
v = cast<uint8_t>(0ULL);
}
}
tmp215:;
if (((bool(m->OnRead) && (addr >= m->RWatchLo)) && (addr < m->RWatchHi))) {
m->OnRead(addr,cast<uint32_t>(v),allegrex_CPU_CurPC(m->CPU));
}
return v;
}
}
}
// tools/platform/psp/machine.go:194:1
void psp_Machine_Write(psp_Machine* m,uint32_t addr,uint8_t v){
{
uint32_t p = psp_phys(addr);
{
if (((p >= cast<uint32_t>(134217728ULL)) && (p < cast<uint32_t>(167772160ULL)))){
m->ram[cast<uint32_t>((p - cast<uint32_t>(134217728ULL)))] = v;
}
else if (((p >= cast<uint32_t>(67108864ULL)) && (p < cast<uint32_t>(69206016ULL)))){
m->vram[cast<uint32_t>((p - cast<uint32_t>(67108864ULL)))] = v;
}
else if (((p >= cast<uint32_t>(65536ULL)) && (p < cast<uint32_t>(81920ULL)))){
m->scratch[cast<uint32_t>((p - cast<uint32_t>(65536ULL)))] = v;
}
else {
m->io[p] = cast<uint32_t>(v);
psp_Machine_note(m,std::string("write 0x%02X to unmapped 0x%08X at pc 0x%08X",44),v,addr,allegrex_CPU_CurPC(m->CPU));
}
}
tmp216:;
if (((bool(m->OnWrite) && (addr >= m->WatchLo)) && (addr < m->WatchHi))) {
m->OnWrite(addr,cast<uint32_t>(v),allegrex_CPU_CurPC(m->CPU));
}
}
}
// tools/platform/psp/machine.go:238:1
std::string psp_Machine_TTY(psp_Machine* m){
{
return cast<std::string>(m->tty);
}
}
// tools/platform/psp/machine.go:242:1
std::string psp_Machine_CurrentThread(psp_Machine* m){
{
if ((!m->current)) {
return std::string("",0);
}
return m->current->name;
}
}
// tools/platform/psp/machine.go:251:1
Slice<std::string> psp_Machine_KObjects(psp_Machine* m){
{
Slice<std::string> out={};
{auto&& tmp217 = m->handles;
for(auto [tmp218,tmp219]:tmp217){
auto h=tmp218;auto o=tmp219;{
auto tmp221=o->kind;
if (tmp221==(std::string("sema",4))){
out = append(out,go_fmt_Sprintf(std::string("sema %d %q count %d",19),h,o->name,o->count));
}
else if (tmp221==(std::string("evflag",6))){
out = append(out,go_fmt_Sprintf(std::string("evflag %d %q bits 0x%X",22),h,o->name,o->bits));
}
else if (tmp221==(std::string("vpl",3))){
out = append(out,go_fmt_Sprintf(std::string("vpl %d %q at 0x%08X used %d/%d",30),h,o->name,o->addr,o->used,o->size));
}
else if (tmp221==(std::string("thread",6))){
if ((o->tstate == cast<psp_threadState>(3ULL))) {
{
if ((o->waitEv != cast<uint32_t>(0ULL))){
out = append(out,go_fmt_Sprintf(std::string("thread %q waits evflag %d (bits 0x%X mode 0x%X)",47),o->name,o->waitEv,o->waitBits,o->waitMode));
}
else if ((o->waitSema != cast<uint32_t>(0ULL))){
out = append(out,go_fmt_Sprintf(std::string("thread %q waits sema %d (need %d)",33),o->name,o->waitSema,o->waitNeed));
}
else if ((o->wakeVblank != cast<uint32_t>(0ULL))){
out = append(out,go_fmt_Sprintf(std::string("thread %q sleeps until vblank %d (now %d)",41),o->name,o->wakeVblank,m->vblanks));
}
else {
out = append(out,go_fmt_Sprintf(std::string("thread %q sleeps (WakeupThread)",31),o->name));
}
}
tmp222:;
}
}
}
tmp220:;
}}
return out;
}
}
// tools/platform/psp/machine.go:284:1
Slice<std::string> psp_Machine_Threads(psp_Machine* m){
{
std::array<std::string,4> states = std::array<std::string,4>{std::string("dormant",7),std::string("ready",5),std::string("running",7),std::string("waiting",7)};
Slice<std::string> out={};
{auto&& tmp223 = m->handles;
for(auto [tmp224,tmp225]:tmp223){
auto h=tmp224;auto o=tmp225;if ((o->kind != std::string("thread",6))) {
continue;
}
uint32_t pc = o->ctx.PC;
if ((o == m->current)) {
pc = m->CPU->PC;
}
out = append(out,go_fmt_Sprintf(std::string("thread %d %q entry 0x%08X prio %d %s pc 0x%08X",46),h,o->name,o->entry,o->priority,states[o->tstate],pc));
}}
return out;
}
}
// tools/platform/psp/machine.go:305:1
Error psp_Machine_LoadModule(psp_Machine* m,psp_Module* mod){
{
if ((mod->Type == cast<uint16_t>(65440ULL))) {
psp_Module_Relocate(mod,cast<uint32_t>(142622720ULL));
}
{auto&& tmp226 = mod->Segments;
for(int64_t tmp227=0;tmp227<len(tmp226);++tmp227){
auto s=tmp226[tmp227];{auto&& tmp228 = s.Data;
for(int64_t tmp229=0;tmp229<len(tmp228);++tmp229){
auto i=tmp229;auto b=tmp228[tmp229];psp_Machine_writeRAM(m,cast<uint32_t>((s.VAddr + cast<uint32_t>(i))),b);
}}
}}
uint32_t k0 = psp_Machine_threadK0(m,cast<uint32_t>(4096ULL),cast<uint32_t>(167706624ULL));
allegrex_CPU_SetPC(m->CPU,mod->EntryPC);
allegrex_CPU_SetReg(m->CPU,cast<uint32_t>(26ULL),k0);
allegrex_CPU_SetReg(m->CPU,cast<uint32_t>(28ULL),mod->GP);
allegrex_CPU_SetReg(m->CPU,cast<uint32_t>(29ULL),k0);
allegrex_CPU_SetReg(m->CPU,cast<uint32_t>(30ULL),k0);
allegrex_CPU_SetReg(m->CPU,cast<uint32_t>(31ULL),cast<uint32_t>(251658240ULL));
allegrex_CPU_SetReg(m->CPU,cast<uint32_t>(4ULL),cast<uint32_t>(0ULL));
allegrex_CPU_SetReg(m->CPU,cast<uint32_t>(5ULL),cast<uint32_t>(0ULL));
uint32_t top={};
{auto&& tmp230 = mod->Segments;
for(int64_t tmp231=0;tmp231<len(tmp230);++tmp231){
auto s=tmp230[tmp231];{
uint32_t end = cast<uint32_t>((s.VAddr + s.MemSize));
if ((end > top)) {
top = end;
}
}
}}
m->heapPtr = ((cast<uint32_t>((top + cast<uint32_t>(4095ULL)))) & ~(cast<uint32_t>(4095ULL)));
m->heapEnd = cast<uint32_t>(166658048ULL);
psp_Machine_installStubs(m,mod);
return {};
}
}
// tools/platform/psp/machine.go:341:1
void psp_Machine_writeRAM(psp_Machine* m,uint32_t addr,uint8_t v){
{
uint32_t p = psp_phys(addr);
{
if (((p >= cast<uint32_t>(134217728ULL)) && (p < cast<uint32_t>(167772160ULL)))){
m->ram[cast<uint32_t>((p - cast<uint32_t>(134217728ULL)))] = v;
}
else if (((p >= cast<uint32_t>(67108864ULL)) && (p < cast<uint32_t>(69206016ULL)))){
m->vram[cast<uint32_t>((p - cast<uint32_t>(67108864ULL)))] = v;
}
else if (((p >= cast<uint32_t>(65536ULL)) && (p < cast<uint32_t>(81920ULL)))){
m->scratch[cast<uint32_t>((p - cast<uint32_t>(65536ULL)))] = v;
}
}
tmp232:;
}
}
// tools/platform/psp/machine.go:354:1
void psp_Machine_SetImageHash(psp_Machine* m,std::string h){
{
m->imageHash = h;
}
}
// tools/platform/psp/mpeg.go:38:1
uint32_t psp_Machine_beGuest32(psp_Machine* m,uint32_t addr){
{
return cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((shl<uint32_t>(cast<uint32_t>(psp_Machine_Read(m,addr)),cast<int64_t>(24ULL)) | shl<uint32_t>(cast<uint32_t>(psp_Machine_Read(m,cast<uint32_t>((addr + cast<uint32_t>(1ULL))))),cast<int64_t>(16ULL)))) | shl<uint32_t>(cast<uint32_t>(psp_Machine_Read(m,cast<uint32_t>((addr + cast<uint32_t>(2ULL))))),cast<int64_t>(8ULL)))) | cast<uint32_t>(psp_Machine_Read(m,cast<uint32_t>((addr + cast<uint32_t>(3ULL)))))));
}
}
// tools/platform/psp/mpeg.go:44:1
void psp_Machine_writeMpegAu(psp_Machine* m,uint32_t au,uint32_t pts,uint32_t esBuf,uint32_t esSize){
{
psp_Machine_write32(m,cast<uint32_t>((au + cast<uint32_t>(0ULL))),cast<uint32_t>(0ULL));
psp_Machine_write32(m,cast<uint32_t>((au + cast<uint32_t>(4ULL))),pts);
psp_Machine_write32(m,cast<uint32_t>((au + cast<uint32_t>(8ULL))),cast<uint32_t>(0ULL));
psp_Machine_write32(m,cast<uint32_t>((au + cast<uint32_t>(12ULL))),pts);
psp_Machine_write32(m,cast<uint32_t>((au + cast<uint32_t>(16ULL))),esBuf);
psp_Machine_write32(m,cast<uint32_t>((au + cast<uint32_t>(20ULL))),esSize);
}
}
// tools/platform/psp/mpeg.go:56:1
uint32_t psp_Machine_mpegRingbufferConstruct(psp_Machine* m,uint32_t rb,uint32_t packets,uint32_t data,uint32_t size,uint32_t cb,uint32_t cbArg){
{
auto tmp233 = std::make_tuple(rb,packets,data);
m->mpeg.Ringbuf = std::get<0>(tmp233);
m->mpeg.Packets = std::get<1>(tmp233);
m->mpeg.Data = std::get<2>(tmp233);
auto tmp234 = std::make_tuple(cb,cbArg);
m->mpeg.Cb = std::get<0>(tmp234);
m->mpeg.CbArg = std::get<1>(tmp234);
m->mpeg.In = cast<uint32_t>(0ULL);
psp_Machine_write32(m,cast<uint32_t>((rb + cast<uint32_t>(0ULL))),packets);
psp_Machine_write32(m,cast<uint32_t>((rb + cast<uint32_t>(4ULL))),cast<uint32_t>(0ULL));
psp_Machine_write32(m,cast<uint32_t>((rb + cast<uint32_t>(8ULL))),cast<uint32_t>(0ULL));
psp_Machine_write32(m,cast<uint32_t>((rb + cast<uint32_t>(12ULL))),packets);
psp_Machine_write32(m,cast<uint32_t>((rb + cast<uint32_t>(16ULL))),cast<uint32_t>(2048ULL));
psp_Machine_write32(m,cast<uint32_t>((rb + cast<uint32_t>(20ULL))),data);
psp_Machine_write32(m,cast<uint32_t>((rb + cast<uint32_t>(24ULL))),cb);
psp_Machine_write32(m,cast<uint32_t>((rb + cast<uint32_t>(28ULL))),cbArg);
psp_Machine_write32(m,cast<uint32_t>((rb + cast<uint32_t>(32ULL))),cast<uint32_t>((data + size)));
psp_Machine_write32(m,cast<uint32_t>((rb + cast<uint32_t>(36ULL))),cast<uint32_t>(0ULL));
psp_Machine_write32(m,cast<uint32_t>((rb + cast<uint32_t>(40ULL))),cast<uint32_t>(0ULL));
psp_Machine_note(m,std::string("sceMpegRingbufferConstruct: %d packets, data 0x%08X, cb 0x%08X",62),packets,data,cb);
return cast<uint32_t>(0ULL);
}
}
// tools/platform/psp/mpeg.go:78:1
uint32_t psp_Machine_mpegRingbufferPut(psp_Machine* m,uint32_t rb,uint32_t n,uint32_t avail){
{
uint32_t free = cast<uint32_t>((m->mpeg.Packets - m->mpeg.In));
if ((n > free)) {
n = free;
}
if ((n > avail)) {
n = avail;
}
if (((n == cast<uint32_t>(0ULL)) || (m->mpeg.Cb == cast<uint32_t>(0ULL)))) {
return cast<uint32_t>(0ULL);
}
uint32_t got = psp_Machine_callGuest(m,m->mpeg.Cb,m->mpeg.Data,n,m->mpeg.CbArg);
if ((cast<uint32_t>((got & cast<uint32_t>(2147483648ULL))) != cast<uint32_t>(0ULL))) {
psp_Machine_note(m,std::string("sceMpegRingbufferPut: callback returned 0x%08X",46),got);
return got;
}
if ((got > n)) {
got = n;
}
m->mpeg.In += got;
m->mpeg.FedTotal += got;
psp_Machine_write32(m,cast<uint32_t>((rb + cast<uint32_t>(12ULL))),cast<uint32_t>((m->mpeg.Packets - m->mpeg.In)));
psp_Machine_write32(m,cast<uint32_t>((rb + cast<uint32_t>(8ULL))),cast<uint32_t>((psp_Machine_read32(m,cast<uint32_t>((rb + cast<uint32_t>(8ULL)))) + got)));
return got;
}
}
// tools/platform/psp/mpeg.go:129:1
void psp_Machine_atracParseRiff(psp_Machine* m,psp_atracState* a){
{
a->Channels = cast<uint32_t>(2ULL);
a->Frames = divi<uint32_t>(a->Size,cast<uint32_t>(512ULL));
if (((((psp_Machine_Read(m,a->Buf) != cast<uint8_t>(82ULL)) || (psp_Machine_Read(m,cast<uint32_t>((a->Buf + cast<uint32_t>(1ULL)))) != cast<uint8_t>(73ULL))) || (psp_Machine_Read(m,cast<uint32_t>((a->Buf + cast<uint32_t>(2ULL)))) != cast<uint8_t>(70ULL))) || (psp_Machine_Read(m,cast<uint32_t>((a->Buf + cast<uint32_t>(3ULL)))) != cast<uint8_t>(70ULL)))) {
return ;
}
uint32_t blockAlign={};
uint32_t dataSize={};
{uint32_t p = cast<uint32_t>((a->Buf + cast<uint32_t>(12ULL)));for (;(cast<uint32_t>((p + cast<uint32_t>(8ULL))) < cast<uint32_t>((a->Buf + a->Size)));){
std::string tag = cast<std::string>(Slice<uint8_t>{psp_Machine_Read(m,p),psp_Machine_Read(m,cast<uint32_t>((p + cast<uint32_t>(1ULL)))),psp_Machine_Read(m,cast<uint32_t>((p + cast<uint32_t>(2ULL)))),psp_Machine_Read(m,cast<uint32_t>((p + cast<uint32_t>(3ULL))))});
uint32_t sz = psp_Machine_read32(m,cast<uint32_t>((p + cast<uint32_t>(4ULL))));
{
auto tmp236=tag;
if (tmp236==(std::string("fmt ",4))){
a->Channels = cast<uint32_t>((cast<uint32_t>(psp_Machine_Read(m,cast<uint32_t>((p + cast<uint32_t>(10ULL))))) | shl<uint32_t>(cast<uint32_t>(psp_Machine_Read(m,cast<uint32_t>((p + cast<uint32_t>(11ULL))))),cast<int64_t>(8ULL))));
blockAlign = cast<uint32_t>((cast<uint32_t>(psp_Machine_Read(m,cast<uint32_t>((p + cast<uint32_t>(20ULL))))) | shl<uint32_t>(cast<uint32_t>(psp_Machine_Read(m,cast<uint32_t>((p + cast<uint32_t>(21ULL))))),cast<int64_t>(8ULL))));
}
else if (tmp236==(std::string("data",4))){
dataSize = sz;
}
}
tmp235:;
p += cast<uint32_t>((cast<uint32_t>(8ULL) + ((cast<uint32_t>((sz + cast<uint32_t>(1ULL)))) & ~(cast<uint32_t>(1ULL)))));
}
}if (((blockAlign != cast<uint32_t>(0ULL)) && (dataSize != cast<uint32_t>(0ULL)))) {
a->Frames = divi<uint32_t>(dataSize,blockAlign);
}
if (((a->Channels == cast<uint32_t>(0ULL)) || (a->Channels > cast<uint32_t>(2ULL)))) {
a->Channels = cast<uint32_t>(2ULL);
}
}
}
// tools/platform/psp/mpeg.go:158:1
uint32_t psp_Machine_atracDecode(psp_Machine* m,uint32_t id,uint32_t out,uint32_t samplesPtr,uint32_t endPtr,uint32_t remainPtr){
{
psp_atracState* a = get(m->atrac,id);
if ((!a)) {
return cast<uint32_t>(2153971716ULL);
}
if ((a->Pos >= a->Frames)) {
if ((samplesPtr != cast<uint32_t>(0ULL))) {
psp_Machine_write32(m,samplesPtr,cast<uint32_t>(0ULL));
}
if ((endPtr != cast<uint32_t>(0ULL))) {
psp_Machine_write32(m,endPtr,cast<uint32_t>(1ULL));
}
if ((remainPtr != cast<uint32_t>(0ULL))) {
psp_Machine_write32(m,remainPtr,cast<uint32_t>(0ULL));
}
return cast<uint32_t>(2153971714ULL);
}
a->Pos++;
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < cast<uint32_t>((cast<uint32_t>(4096ULL) * a->Channels)));i += cast<uint32_t>(4ULL)){
psp_Machine_write32(m,cast<uint32_t>((out + i)),cast<uint32_t>(0ULL));
}
}if ((samplesPtr != cast<uint32_t>(0ULL))) {
psp_Machine_write32(m,samplesPtr,cast<uint32_t>(2048ULL));
}
uint32_t end = cast<uint32_t>(0ULL);
if ((a->Pos >= a->Frames)) {
end = cast<uint32_t>(1ULL);
}
if ((endPtr != cast<uint32_t>(0ULL))) {
psp_Machine_write32(m,endPtr,end);
}
if ((remainPtr != cast<uint32_t>(0ULL))) {
psp_Machine_write32(m,remainPtr,cast<uint32_t>((a->Frames - a->Pos)));
}
return cast<uint32_t>(0ULL);
}
}
// tools/platform/psp/mpeg.go:199:1
uint32_t psp_Machine_mpegGetAu(psp_Machine* m,uint32_t au){
{
if ((m->mpeg.In == cast<uint32_t>(0ULL))) {
return cast<uint32_t>(2153873409ULL);
}
m->mpeg.In--;
{
uint32_t rb = m->mpeg.Ringbuf;
if ((rb != cast<uint32_t>(0ULL))) {
psp_Machine_write32(m,cast<uint32_t>((rb + cast<uint32_t>(12ULL))),cast<uint32_t>((m->mpeg.Packets - m->mpeg.In)));
psp_Machine_write32(m,cast<uint32_t>((rb + cast<uint32_t>(4ULL))),cast<uint32_t>((psp_Machine_read32(m,cast<uint32_t>((rb + cast<uint32_t>(4ULL)))) + cast<uint32_t>(1ULL))));
}
}
m->mpeg.Pts += cast<uint32_t>(3003ULL);
psp_Machine_writeMpegAu(m,au,m->mpeg.Pts,psp_Machine_read32(m,cast<uint32_t>((au + cast<uint32_t>(16ULL)))),cast<uint32_t>(2048ULL));
return cast<uint32_t>(0ULL);
}
}
// tools/platform/psp/profile.go:109:1
time_Time psp_Machine_profStart(psp_Machine* m){
{
if ((!m->Profile)) {
return time_Time{};
}
return go_time_Now();
}
}
// tools/platform/psp/profile.go:116:1
void psp_Machine_profEnd(psp_Machine* m,int64_t bucket,time_Time t){
{
if (time_Time_IsZero(t)) {
return ;
}
m->prof.ns[bucket] += cast<int64_t>(go_time_Since(t));
m->prof.count[bucket]++;
}
}
// tools/platform/psp/profile.go:125:1
int64_t psp_Machine_profGeNs(psp_Machine* m){
{
int64_t n={};
{auto&& tmp237 = psp_geBuckets;
for(int64_t tmp238=0;tmp238<len(tmp237);++tmp238){
auto b=tmp237[tmp238];n += m->prof.ns[b];
}}
return n;
}
}
// tools/platform/psp/profile.go:143:1
void psp_Machine_profEndSyscall(psp_Machine* m,time_Time t,int64_t geBefore,int64_t gen){
{
if ((time_Time_IsZero(t) || (gen != m->prof.gen))) {
return ;
}
int64_t ns = cast<int64_t>((cast<int64_t>(go_time_Since(t)) - (cast<int64_t>((psp_Machine_profGeNs(m) - geBefore)))));
if ((ns < cast<int64_t>(0ULL))) {
ns = cast<int64_t>(0ULL);
}
m->prof.ns[cast<int64_t>(4ULL)] += ns;
m->prof.count[cast<int64_t>(4ULL)]++;
}
}
// tools/platform/psp/profile.go:157:1
void psp_Machine_profRunEnter(psp_Machine* m){
{
if ((!m->Profile)) {
return ;
}
auto tmp239 = std::make_tuple(go_time_Now(),true);
m->prof.runStart = std::get<0>(tmp239);
m->prof.inRun = std::get<1>(tmp239);
}
}
// tools/platform/psp/profile.go:164:1
void psp_Machine_profRunExit(psp_Machine* m){
{
if (((!m->Profile) || (!m->prof.inRun))) {
return ;
}
m->prof.frameNs += cast<int64_t>(go_time_Since(m->prof.runStart));
m->prof.inRun = false;
}
}
// tools/platform/psp/profile.go:175:1
void psp_Machine_profFrame(psp_Machine* m){
{
if ((!m->Profile)) {
return ;
}
psp_profState* p = (&m->prof);
int64_t total = p->frameNs;
if (p->inRun) {
total += cast<int64_t>(go_time_Since(p->runStart));
p->runStart = go_time_Now();
}
p->frameNs = cast<int64_t>(0ULL);
auto ms = [&](int64_t ns)->double{
return (cast<double>(ns) / 1e6);
}
;
Slice<psp_ProfileBucket> buckets = Slice<psp_ProfileBucket>::make(cast<int64_t>(0ULL),cast<int64_t>(6ULL));
int64_t summed={};
{int64_t b = cast<int64_t>(0ULL);for (;(b < cast<int64_t>(5ULL));b++){
summed += p->ns[b];
buckets = append(buckets,psp_ProfileBucket{psp_bucketNames[b],ms(p->ns[b]),cast<int64_t>(p->count[b])});
}
}int64_t other = cast<int64_t>((total - summed));
if ((other < cast<int64_t>(0ULL))) {
other = cast<int64_t>(0ULL);
}
buckets = append(buckets,psp_ProfileBucket{std::string("allegrex + rest (derived)",25),ms(other),cast<int64_t>(0ULL)});
psp_geCounters now = m->geCnt;
psp_geCounters d = psp_geCounters{cast<int64_t>((now.lists - p->base.lists)),cast<int64_t>((now.prims - p->base.prims)),cast<int64_t>((now.verts - p->base.verts)),cast<int64_t>((now.frags - p->base.frags)),cast<int64_t>((now.zKilled - p->base.zKilled)),cast<int64_t>((now.aKilled - p->base.aKilled)),cast<int64_t>((now.sKilled - p->base.sKilled)),cast<int64_t>((now.scissored - p->base.scissored)),cast<int64_t>((now.xfers - p->base.xfers))};
p->last = psp_FrameProfile{ms(total),buckets,Slice<psp_ProfileCounter>{psp_ProfileCounter{std::string("display lists",13),d.lists},psp_ProfileCounter{std::string("primitives",10),d.prims},psp_ProfileCounter{std::string("vertices",8),d.verts},psp_ProfileCounter{std::string("fragments drawn",15),d.frags},psp_ProfileCounter{std::string("depth-killed",12),d.zKilled},psp_ProfileCounter{std::string("alpha-killed",12),d.aKilled},psp_ProfileCounter{std::string("stencil-killed",14),d.sKilled},psp_ProfileCounter{std::string("scissored",9),d.scissored},psp_ProfileCounter{std::string("block transfers",15),d.xfers},psp_ProfileCounter{std::string("allegrex instructions",21),cast<int64_t>(cast<uint64_t>((m->CPU->Steps - p->baseInstr)))}}};
p->has = true;
auto tmp240 = std::make_tuple(std::array<int64_t,5>{},std::array<int64_t,5>{});
p->ns = std::get<0>(tmp240);
p->count = std::get<1>(tmp240);
auto tmp241 = std::make_tuple(now,m->CPU->Steps);
p->base = std::get<0>(tmp241);
p->baseInstr = std::get<1>(tmp241);
p->gen++;
}
}
// tools/platform/psp/profile.go:239:1
psp_FrameProfile psp_Machine_FrameProfile(psp_Machine* m){
{
if ((!m->prof.has)) {
return psp_FrameProfile{};
}
return m->prof.last;
}
}
// tools/platform/psp/profile.go:249:1
void psp_Machine_SetProfile(psp_Machine* m,bool on){
{
m->Profile = on;
m->prof = psp_profState{};
if (on) {
auto tmp242 = std::make_tuple(m->geCnt,m->CPU->Steps);
m->prof.base = std::get<0>(tmp242);
m->prof.baseInstr = std::get<1>(tmp242);
}
}
}
// tools/platform/psp/prx.go:41:1
std::tuple<Slice<uint8_t>,uint32_t,Error> psp_DecryptPRX(Slice<uint8_t> raw){
Slice<uint8_t> plain{};
uint32_t tag{};
Error err{};
{
if (((len(raw) < cast<int64_t>(336ULL)) || (cast<std::string>(sub(raw,cast<int64_t>(0ULL),cast<int64_t>(4ULL))) != std::string("~PSP",4)))) {
return {{},cast<uint32_t>(0ULL),go_fmt_Errorf(std::string("prx: not a ~PSP container",25))};
}
tag = le_Uint32(sub(raw,cast<int64_t>(208ULL),len(raw)));
auto tmp243 = lookup(psp_tagTable,tag);
psp_tagInfo ti = std::get<0>(tmp243);
bool ok = std::get<1>(tmp243);
if ((!ok)) {
return {{},tag,go_fmt_Errorf(std::string("prx: unknown ~PSP tag 0x%08X",28),tag)};
}
{auto&& tmp244 = Slice<bool>{false,true};
for(int64_t tmp245=0;tmp245<len(tmp244);++tmp245){
auto preDecrypt=tmp244[tmp245];auto tmp246 = psp_decryptTagType(raw,ti,preDecrypt);
Slice<uint8_t> out = std::get<0>(tmp246);
bool matched = std::get<1>(tmp246);
Error derr = std::get<2>(tmp246);
if (bool(derr)) {
return {{},tag,derr};
}
if (matched) {
return {out,tag,{}};
}
}}
return {{},tag,go_fmt_Errorf(std::string("prx: SHA-1 header check failed for tag 0x%08X",45),tag)};
}
}
// tools/platform/psp/prx.go:66:1
std::tuple<Slice<uint8_t>,bool,Error> psp_decryptTagType(Slice<uint8_t> raw,psp_tagInfo ti,bool preDecrypt){
Slice<uint8_t> out{};
bool matched{};
Error err{};
{
Slice<uint8_t> seed = ti.seed;
int64_t decryptSize = cast<int64_t>(cast<int32_t>(le_Uint32(sub(raw,cast<int64_t>(176ULL),len(raw)))));
Slice<uint8_t> st = Slice<uint8_t>::make(cast<int64_t>(336ULL));
gcopy(sub(st,cast<int64_t>(0ULL),cast<int64_t>(4ULL)),sub(raw,cast<int64_t>(208ULL),cast<int64_t>(212ULL)));
gcopy(sub(st,cast<int64_t>(4ULL),cast<int64_t>(24ULL)),sub(raw,cast<int64_t>(212ULL),cast<int64_t>(232ULL)));
gcopy(sub(st,cast<int64_t>(24ULL),cast<int64_t>(64ULL)),sub(raw,cast<int64_t>(232ULL),cast<int64_t>(272ULL)));
gcopy(sub(st,cast<int64_t>(64ULL),cast<int64_t>(128ULL)),sub(raw,cast<int64_t>(272ULL),cast<int64_t>(336ULL)));
gcopy(sub(st,cast<int64_t>(128ULL),cast<int64_t>(208ULL)),sub(raw,cast<int64_t>(128ULL),cast<int64_t>(208ULL)));
gcopy(sub(st,cast<int64_t>(208ULL),cast<int64_t>(336ULL)),sub(raw,cast<int64_t>(0ULL),cast<int64_t>(128ULL)));
if (preDecrypt) {
{
Error err = psp_kirk7(sub(st,cast<int64_t>(16ULL),cast<int64_t>(176ULL)),sub(st,cast<int64_t>(16ULL),cast<int64_t>(176ULL)),ti.code);
if (bool(err)) {
return {{},false,err};
}
}
}
SHA1* h = go_sha1_New();
hash_Write(h,sub(seed,cast<int64_t>(0ULL),cast<int64_t>(20ULL)));
hash_Write(h,sub(st,cast<int64_t>(24ULL),cast<int64_t>(64ULL)));
hash_Write(h,sub(st,cast<int64_t>(64ULL),cast<int64_t>(208ULL)));
hash_Write(h,sub(st,cast<int64_t>(208ULL),cast<int64_t>(336ULL)));
if ((!psp_bytesEqual(hash_Sum(h,{}),sub(st,cast<int64_t>(4ULL),cast<int64_t>(24ULL))))) {
return {{},false,{}};
}
constexpr int64_t offset=64ULL;
Slice<uint8_t> buf = Slice<uint8_t>::make(len(raw));
gcopy(buf,raw);
gcopy(sub(buf,cast<int64_t>(64ULL),cast<int64_t>(208ULL)),sub(st,cast<int64_t>(64ULL),cast<int64_t>(208ULL)));
gcopy(sub(buf,cast<int64_t>(208ULL),cast<int64_t>(336ULL)),sub(st,cast<int64_t>(208ULL),cast<int64_t>(336ULL)));
Slice<uint8_t> tmp = Slice<uint8_t>::make(cast<int64_t>(112ULL));
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(112ULL));i++){
tmp[i] = cast<uint8_t>((st[cast<int64_t>((cast<int64_t>(64ULL) + i))] ^ seed[cast<int64_t>((cast<int64_t>(20ULL) + i))]));
}
}{
Error err = psp_kirk7(tmp,tmp,ti.code);
if (bool(err)) {
return {{},false,err};
}
}
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(112ULL));i++){
buf[cast<int64_t>((cast<int64_t>(64ULL) + i))] = cast<uint8_t>((tmp[i] ^ seed[cast<int64_t>((cast<int64_t>(32ULL) + i))]));
}
}auto tmp247 = psp_kirkCMD1(sub(buf,cast<int64_t>(64ULL),len(buf)));
Slice<uint8_t> plain = std::get<0>(tmp247);
err = std::get<1>(tmp247);
if (bool(err)) {
return {{},false,err};
}
if (((decryptSize >= cast<int64_t>(0ULL)) && (decryptSize <= len(plain)))) {
plain = sub(plain,0,decryptSize);
}
return {plain,true,{}};
}
}
// tools/platform/psp/prx.go:129:1
bool psp_bytesEqual(Slice<uint8_t> a,Slice<uint8_t> b){
{
if ((len(a) != len(b))) {
return false;
}
{auto&& tmp248 = a;
for(int64_t tmp249=0;tmp249<len(tmp248);++tmp249){
auto i=tmp249;if ((a[i] != b[i])) {
return false;
}
}}
return true;
}
}
// tools/platform/psp/run.go:15:1
std::string psp_Result_String(psp_Result r){
{
return go_fmt_Sprintf(std::string("stopped at 0x%08X after %d steps: %s",36),r.PC,r.Steps,r.Reason);
}
}
// tools/platform/psp/run.go:96:1
bool psp_Machine_mapped(psp_Machine* m,uint32_t p){
{
{
if (((p >= cast<uint32_t>(134217728ULL)) && (p < cast<uint32_t>(167772160ULL)))){
return true;
}
else if (((p >= cast<uint32_t>(67108864ULL)) && (p < cast<uint32_t>(69206016ULL)))){
return true;
}
else if (((p >= cast<uint32_t>(65536ULL)) && (p < cast<uint32_t>(81920ULL)))){
return true;
}
}
tmp250:;
return false;
}
}
// tools/platform/psp/sched.go:36:1
void psp_Machine_startThread(psp_Machine* m,uint32_t uid,psp_kobject* o,uint32_t argLen,uint32_t argPtr){
{
uint32_t k0 = psp_Machine_threadK0(m,uid,o->stackTop);
allegrex_CPUState ctx = allegrex_CPUState{};
auto tmp251 = std::make_tuple(cast<uint32_t>(228ULL),cast<uint32_t>(228ULL));
ctx.VfpuCtrl[cast<int64_t>(0ULL)] = std::get<0>(tmp251);
ctx.VfpuCtrl[cast<int64_t>(1ULL)] = std::get<1>(tmp251);
ctx.R[cast<int64_t>(26ULL)] = k0;
ctx.R[cast<int64_t>(28ULL)] = allegrex_CPU_Reg(m->CPU,cast<uint32_t>(28ULL));
uint32_t sp = k0;
if (((argLen > cast<uint32_t>(0ULL)) && (argPtr != cast<uint32_t>(0ULL)))) {
uint32_t cp = ((cast<uint32_t>((k0 - argLen))) & ~(cast<uint32_t>(15ULL)));
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < argLen);i++){
psp_Machine_Write(m,cast<uint32_t>((cp + i)),psp_Machine_Read(m,cast<uint32_t>((argPtr + i))));
}
}argPtr = cp;
sp = cp;
}
ctx.R[cast<int64_t>(29ULL)] = sp;
ctx.R[cast<int64_t>(30ULL)] = sp;
ctx.R[cast<int64_t>(31ULL)] = cast<uint32_t>(251658240ULL);
ctx.R[cast<int64_t>(4ULL)] = argLen;
ctx.R[cast<int64_t>(5ULL)] = argPtr;
ctx.Out = ctx.R;
ctx.PC = o->entry;
ctx.NextPC = cast<uint32_t>((o->entry + cast<uint32_t>(4ULL)));
ctx.Steps = m->CPU->Steps;
o->ctx = ctx;
o->tstate = cast<psp_threadState>(1ULL);
}
}
// tools/platform/psp/sched.go:69:1
uint32_t psp_Machine_threadK0(psp_Machine* m,uint32_t uid,uint32_t stackTop){
{
uint32_t k0 = ((cast<uint32_t>((stackTop - cast<uint32_t>(256ULL)))) & ~(cast<uint32_t>(15ULL)));
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < cast<uint32_t>(256ULL));i++){
psp_Machine_Write(m,cast<uint32_t>((k0 + i)),cast<uint8_t>(0ULL));
}
}psp_Machine_write32(m,cast<uint32_t>((k0 + cast<uint32_t>(192ULL))),uid);
uint32_t stackBase = k0;
{
psp_kobject* o = get(m->handles,uid);
if ((bool(o) && (o->kind == std::string("thread",6)))) {
stackBase = o->addr;
}
}
psp_Machine_write32(m,cast<uint32_t>((k0 + cast<uint32_t>(200ULL))),stackBase);
psp_Machine_write32(m,cast<uint32_t>((k0 + cast<uint32_t>(248ULL))),cast<uint32_t>(4294967295ULL));
psp_Machine_write32(m,cast<uint32_t>((k0 + cast<uint32_t>(252ULL))),cast<uint32_t>(4294967295ULL));
psp_Machine_write32(m,stackBase,uid);
return k0;
}
}
// tools/platform/psp/sched.go:91:1
bool psp_Machine_schedule(psp_Machine* m,psp_threadState currentBecomes){
{
if (bool(m->current)) {
m->current->ctx = allegrex_CPU_SaveState(m->CPU);
if ((m->current->tstate == cast<psp_threadState>(2ULL))) {
m->current->tstate = currentBecomes;
}
}
{int64_t tries = cast<int64_t>(0ULL);for (;(tries < cast<int64_t>(600ULL));tries++){
psp_kobject* best={};
{auto&& tmp252 = m->handles;
for(auto [tmp253,tmp254]:tmp252){
auto o=tmp254;if (((o->kind == std::string("thread",6)) && (o->tstate == cast<psp_threadState>(1ULL)))) {
if (((!best) || (o->priority < best->priority))) {
best = o;
}
}
}}
if (bool(best)) {
m->current = best;
best->tstate = cast<psp_threadState>(2ULL);
allegrex_CPU_LoadState(m->CPU,best->ctx);
return true;
}
bool timed = false;
{auto&& tmp255 = m->handles;
for(auto [tmp256,tmp257]:tmp255){
auto o=tmp257;if ((((o->kind == std::string("thread",6)) && (o->tstate == cast<psp_threadState>(3ULL))) && (o->wakeVblank != cast<uint32_t>(0ULL)))) {
timed = true;
break;
}
}}
if ((!timed)) {
break;
}
psp_Machine_deliverVBlank(m);
}
}if ((m->doneReason == std::string("",0))) {
m->doneReason = std::string("no runnable threads",19);
}
return false;
}
}
// tools/platform/psp/sched.go:134:1
void psp_Machine_yieldCurrent(psp_Machine* m,psp_threadState newState){
{
if ((!psp_Machine_schedule(m,newState))) {
m->Halted = true;
m->HaltReason = m->doneReason;
}
}
}
// tools/platform/psp/sched.go:143:1
uint32_t psp_Machine_currentThreadID(psp_Machine* m){
{
if ((!m->current)) {
return cast<uint32_t>(4096ULL);
}
{auto&& tmp258 = m->handles;
for(auto [tmp259,tmp260]:tmp258){
auto h=tmp259;auto o=tmp260;if ((o == m->current)) {
return h;
}
}}
return cast<uint32_t>(4096ULL);
}
}
// tools/platform/psp/sched.go:156:1
void psp_Machine_onThreadExit(psp_Machine* m){
{
if (bool(m->current)) {
m->current->tstate = cast<psp_threadState>(0ULL);
}
psp_kobject* saved = m->current;
m->current = {};
if ((!psp_Machine_schedule(m,cast<psp_threadState>(0ULL)))) {
m->Halted = true;
if ((m->doneReason == std::string("",0))) {
m->doneReason = std::string("all threads exited",18);
}
m->HaltReason = m->doneReason;
}
(void)(saved);
}
}

#include "fast.h"
