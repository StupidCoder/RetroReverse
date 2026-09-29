#include "runtime.h"
struct r5900_TLBEntry;
struct r5900_CPU;
struct r5900_mmiOp;
struct r5900_Quad;
struct r5900_Inst;
struct r5900_State;
struct mips_loadSlot;
struct mips_CPU;
struct mips_GTE;
struct mips_Inst;
struct mips_CPUState;
struct mips_GTEState;
struct vu_Inst;
struct vu_VU;
struct vu_flagVals;
struct vu_upperResult;
struct vu_FlagVals;
struct vu_State;
struct iso9660_Geometry;
struct iso9660_Source;
struct iso9660_Volume;
struct iso9660_Entry;
struct ps2_dmacChan;
struct ps2_eeTimer;
struct ps2_Segment;
struct ps2_Symbol;
struct ps2_Executable;
struct ps2_GS;
struct ps2_gsXfer;
struct ps2_weaveField;
struct ps2_gsVertex;
struct ps2_gsTarget;
struct ps2_gsTex;
struct ps2_gsSampler;
struct ps2_idleSnap;
struct ps2_idleState;
struct ps2_handler;
struct ps2_iop_946_kv;
struct ps2_iop_988_kv;
struct ps2_IOP;
struct ps2_IOPIntrEvent;
struct ps2_isRunWrite;
struct ps2_ioTouch;
struct ps2_cdvd;
struct ps2_CDVDState;
struct ps2_iopDMAChan;
struct ps2_iopDMADone;
struct ps2_iopHandler;
struct ps2_iopBlock;
struct ps2_iopHeap;
struct ps2_iopCall;
struct ps2_iopLibrary;
struct ps2_iopFunc;
struct ps2_iopBinding;
struct ps2_IOPModule;
struct ps2_sio2xfer;
struct ps2_sio2PadState;
struct ps2_spuVoice;
struct ps2_spu2;
struct ps2_iopThread;
struct ps2_iopTimer;
struct ps2_iopVblankHandler;
struct ps2_IRXImport;
struct ps2_IRXExport;
struct ps2_IRXReloc;
struct ps2_IRX;
struct ps2_kernel_401_kv;
struct ps2_syscallEntry;
struct ps2_machine_810_unit;
struct ps2_machine_826_kv;
struct ps2_Machine;
struct ps2_PadPress;
struct ps2_padLiveState;
struct ps2_ProfileBucket;
struct ps2_ProfileCounter;
struct ps2_FrameProfile;
struct ps2_profState;
struct ps2_profCounters;
struct ps2_RomEntry;
struct ps2_Result;
struct ps2_thread;
struct ps2_sema;
struct ps2_sifPacket;
struct ps2_sifRPCKey;
struct ps2_vif;
struct ps2_mpgInfo;
struct ps2_countKV;
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
struct Anon69;
struct Anon7;
struct Anon70;
struct Anon71;
struct Anon72;
struct Anon73;
struct Anon74;
struct Anon75;
struct Anon76;
struct Anon77;
struct Anon78;
struct Anon79;
struct Anon8;
struct Anon80;
struct Anon81;
struct Anon82;
struct Anon83;
struct Anon84;
struct Anon85;
struct Anon86;
struct Anon9;
struct r5900_TLBEntry{
uint32_t PageMask{};
uint64_t EntryHi{};
uint64_t EntryLo0{};
uint64_t EntryLo1{};
};
struct r5900_Quad{
uint64_t Lo{};
uint64_t Hi{};
bool operator==(const r5900_Quad&)const=default;
};
struct r5900_CPU{
std::array<r5900_Quad,32> R{};
uint64_t HI{};
uint64_t LO{};
uint64_t HI1{};
uint64_t LO1{};
uint32_t SA{};
uint64_t PC{};
uint64_t nextPC{};
std::array<uint64_t,32> COP0{};
std::array<r5900_TLBEntry,48> TLB{};
std::array<uint32_t,32> FPR{};
uint32_t ACC{};
uint32_t FCR31{};
vu_VU* COP2{};
bool LLBit{};
bool Halted{};
std::string HaltReason{};
uint64_t Steps{};
std::function<bool(r5900_CPU*)> Syscall{};
ps2_Machine* bus{};
std::function<uint32_t(uint32_t)> fetch{};
uint64_t curPC{};
bool delaySlot{};
bool pendingDelay{};
uint64_t branchAddr{};
uint64_t countFrac{};
};
using r5900_mmiForm=int64_t;
struct r5900_mmiOp{
std::string name{};
r5900_mmiForm form{};
};
using r5900_Flow=int64_t;
struct r5900_Inst{
uint32_t Addr{};
int64_t Len{};
std::string Mnem{};
std::string Text{};
r5900_Flow Flow{};
uint32_t Target{};
bool HasTarget{};
bool HasDelay{};
bool Annul{};
};
struct r5900_State{
std::array<r5900_Quad,32> R{};
uint64_t HI{};
uint64_t LO{};
uint64_t HI1{};
uint64_t LO1{};
uint32_t SA{};
uint64_t PC{};
uint64_t NextPC{};
std::array<uint64_t,32> COP0{};
std::array<r5900_TLBEntry,48> TLB{};
std::array<uint32_t,32> FPR{};
uint32_t ACC{};
uint32_t FCR31{};
bool LLBit{};
bool Halted{};
std::string HaltReason{};
uint64_t Steps{};
uint64_t CurPC{};
bool DelaySlot{};
bool PendingDelay{};
uint64_t BranchAddr{};
uint64_t CountFrac{};
};
struct mips_loadSlot{
uint32_t reg{};
uint32_t val{};
};
struct mips_CPU{
std::array<uint32_t,32> R{};
std::array<uint32_t,32> out{};
uint32_t HI{};
uint32_t LO{};
uint32_t PC{};
uint32_t nextPC{};
std::array<uint32_t,32> COP0{};
RRNullGTE* GTE{};
std::function<bool(mips_CPU*)> Syscall{};
bool Halted{};
std::string HaltReason{};
uint64_t Steps{};
ps2_IOP* bus{};
uint32_t curPC{};
mips_loadSlot ld{};
bool delaySlot{};
bool pendingDelay{};
uint32_t branchAddr{};
};
struct mips_GTE{
std::array<uint32_t,32> data{};
std::array<uint32_t,32> ctrl{};
};
using mips_Flow=int64_t;
struct mips_Inst{
uint32_t Addr{};
int64_t Len{};
std::string Mnem{};
std::string Text{};
mips_Flow Flow{};
uint32_t Target{};
bool HasTarget{};
bool HasDelay{};
};
struct mips_CPUState{
std::array<uint32_t,32> R{};
std::array<uint32_t,32> Out{};
uint32_t HI{};
uint32_t LO{};
uint32_t PC{};
uint32_t NextPC{};
std::array<uint32_t,32> COP0{};
bool Halted{};
std::string HaltReason{};
uint64_t Steps{};
uint32_t CurPC{};
uint32_t LdReg{};
uint32_t LdVal{};
bool DelaySlot{};
bool PendingDelay{};
uint32_t BranchAddr{};
};
struct mips_GTEState{
std::array<uint32_t,32> Data{};
std::array<uint32_t,32> Ctrl{};
};
struct vu_Inst{
std::string Upper{};
std::string Lower{};
bool I{};
bool E{};
uint64_t Raw{};
};
struct vu_flagVals{
uint16_t mac{};
uint16_t status{};
uint32_t clip{};
};
struct vu_VU{
std::array<std::array<uint32_t,4>,32> VF{};
std::array<uint16_t,16> VI{};
std::array<float,4> ACC{};
float Q{};
float P{};
float I{};
uint32_t R{};
Slice<uint8_t> Micro{};
Slice<uint8_t> Data{};
uint32_t PC{};
uint16_t Mac{};
uint16_t Status{};
uint32_t Clip{};
std::array<vu_flagVals,4> flagPipe{};
uint16_t visMac{};
uint16_t visStatus{};
uint32_t visClip{};
uint16_t Top{};
uint16_t ITop{};
std::function<void(uint32_t)> XGKick{};
std::function<void(vu_VU*,uint32_t,uint64_t)> Trace{};
std::function<void(uint32_t,uint32_t,uint32_t)> OnMaxStore{};
uint16_t CMSAR0{};
std::function<void(uint32_t)> StartVU1{};
uint64_t Steps{};
Slice<std::array<uint32_t,2>> BranchLog{};
int64_t branchNext{};
};
struct vu_upperResult{
uint8_t kind{};
uint32_t reg{};
uint32_t mask{};
std::array<float,4> val{};
std::array<uint32_t,4> bits{};
uint32_t clip{};
};
struct vu_FlagVals{
uint16_t Mac{};
uint16_t Status{};
uint32_t Clip{};
};
struct vu_State{
std::array<std::array<uint32_t,4>,32> VF{};
std::array<uint16_t,16> VI{};
std::array<float,4> ACC{};
float Q{};
float P{};
float I{};
uint32_t R{};
uint32_t PC{};
uint16_t Mac{};
uint16_t Status{};
uint32_t Clip{};
std::array<vu_FlagVals,4> FlagPipe{};
uint16_t VisMac{};
uint16_t VisStatus{};
uint32_t VisClip{};
uint16_t Top{};
uint16_t ITop{};
uint16_t CMSAR0{};
uint64_t Steps{};
};
struct iso9660_Geometry{
int64_t SectorSize{};
int64_t DataOffset{};
};
struct iso9660_Source{
LocalSource* ra{};
int64_t size{};
iso9660_Geometry geom{};
Slice<uint8_t> buf{};
};
struct iso9660_Volume{
LocalSource* src{};
std::string System{};
std::string Name{};
int64_t Blocks{};
int64_t rootLBA{};
int64_t rootSize{};
};
struct iso9660_Entry{
std::string Name{};
std::string Path{};
bool IsDir{};
int64_t Size{};
int64_t Block{};
};
using iso9660_byteReaderAt=Slice<uint8_t>;
struct ps2_dmacChan{
uint32_t chcr{};
uint32_t madr{};
uint32_t qwc{};
uint32_t tadr{};
uint32_t asr0{};
uint32_t asr1{};
uint32_t sadr{};
};
struct ps2_eeTimer{
uint32_t base{};
uint64_t baseSteps{};
uint32_t mode{};
uint32_t comp{};
uint32_t hold{};
};
struct ps2_Segment{
uint32_t VAddr{};
Slice<uint8_t> Data{};
uint32_t MemSz{};
};
struct ps2_Symbol{
std::string Name{};
uint32_t Addr{};
uint32_t Size{};
bool Func{};
};
struct ps2_Executable{
uint32_t Entry{};
Slice<ps2_Segment> Segments{};
Slice<ps2_Symbol> Symbols{};
Slice<ps2_Symbol> byAddr{};
};
struct ps2_gsXfer{
bool active{};
uint32_t dbp{};
uint32_t dbw{};
uint32_t dpsm{};
uint32_t dsax{};
uint32_t dsay{};
uint32_t rrw{};
uint32_t rrh{};
uint32_t x{};
uint32_t y{};
Slice<uint8_t> partial{};
};
struct ps2_gsVertex{
int32_t x{};
int32_t y{};
uint32_t z{};
uint32_t rgba{};
int32_t u{};
int32_t v{};
float s{};
float t{};
float q{};
uint32_t f{};
};
struct ps2_GS{
Slice<uint8_t> vram{};
std::array<uint64_t,128> reg{};
ps2_gsXfer xfer{};
uint64_t csr{};
int64_t uploads{};
int64_t prims{};
int64_t primsRunBase{};
std::array<ps2_gsVertex,3> vq{};
int64_t vqN{};
uint32_t q{};
std::array<int64_t,8> primCount{};
int64_t curCmd{};
Map<std::string,int64_t> drawCensus{};
std::string src{};
Slice<uint8_t> curPacket{};
Slice<uint8_t> srcData{};
Slice<uint8_t> srcIn{};
Slice<uint8_t> srcMicro{};
Slice<uint8_t> srcVUData{};
bool srcDumped{};
std::array<uint32_t,512> clut{};
uint32_t cbp0{};
uint32_t cbp1{};
uint64_t plotted{};
uint64_t rejScissor{};
uint64_t rejZ{};
uint64_t rejAlpha{};
uint64_t rejDate{};
std::array<uint64_t,8> plotNonBlack{};
uint64_t parFills{};
uint64_t serFills{};
uint64_t rgbaqA0{};
uint64_t rgbaqA{};
uint64_t rgbaqRGB0{};
uint64_t rgbaqRGB{};
int64_t t8Dumped{};
uint64_t texBlack{};
uint64_t texColor{};
std::array<uint64_t,64> texBlackPSM{};
std::array<uint64_t,64> texColorPSM{};
int64_t path2ImageRemain{};
int64_t path2SkipRemain{};
Slice<uint8_t> path2Carry{};
ps2_Machine* m{};
};
struct ps2_weaveField{
Slice<uint8_t> pix{};
int64_t w{};
int64_t h{};
};
struct ps2_gsTarget{
uint32_t fbp{};
uint32_t fbw{};
uint32_t psm{};
uint32_t fbmsk{};
int32_t sx0{};
int32_t sy0{};
int32_t sx1{};
int32_t sy1{};
bool abe{};
uint64_t alpha{};
bool pabe{};
bool colclamp{};
uint32_t fba{};
uint64_t test{};
bool fge{};
uint32_t fogcol{};
uint32_t zbp{};
uint32_t zpsm{};
bool zmsk{};
uint32_t ztst{};
int64_t primType{};
};
struct ps2_gsTex{
uint32_t tbp{};
uint32_t tbw{};
uint32_t psm{};
uint32_t tw{};
uint32_t th{};
bool tcc{};
uint32_t tfx{};
uint32_t cbp{};
uint32_t cpsm{};
uint32_t csm{};
uint32_t csa{};
uint32_t cld{};
};
struct ps2_gsSampler{
ps2_GS* gs{};
ps2_gsTex tex{};
int32_t w{};
int32_t h{};
uint32_t wms{};
uint32_t wmt{};
int32_t minu{};
int32_t maxu{};
int32_t minv{};
int32_t maxv{};
bool linear{};
bool probe{};
bool parallel{};
};
struct ps2_idleSnap{
uint64_t PC{};
uint64_t NextPC{};
uint64_t BranchAddr{};
uint64_t HI{};
uint64_t LO{};
uint64_t HI1{};
uint64_t LO1{};
uint32_t SA{};
uint32_t ACC{};
uint32_t FCR31{};
bool DelaySlot{};
bool PendingDelay{};
bool LLBit{};
std::array<r5900_Quad,32> R{};
std::array<uint32_t,32> FPR{};
bool operator==(const ps2_idleSnap&)const=default;
};
struct ps2_idleState{
bool armed{};
ps2_idleSnap snap{};
uint64_t stores{};
int64_t insns{};
int64_t period{};
uint64_t next{};
Slice<r5900_State> phase{};
uint64_t Skipped{};
uint64_t Hits{};
};
struct ps2_handler{
uint32_t cause{};
uint32_t addr{};
uint32_t arg{};
uint32_t next{};
};
struct ps2_iop_946_kv{
std::string name{};
int64_t n{};
};
struct ps2_iop_988_kv{
std::string name{};
int64_t n{};
};
struct ps2_iopHandler{
uint32_t fn{};
uint32_t arg{};
};
struct ps2_iopBlock{
uint32_t base{};
uint32_t size{};
};
struct ps2_iopVblankHandler{
uint32_t edge{};
uint32_t prio{};
uint32_t fn{};
uint32_t arg{};
};
struct ps2_sio2xfer{
Slice<uint8_t> in{};
Slice<uint8_t> out{};
std::array<uint32_t,2> dmaAddr{};
std::array<uint32_t,2> dmaWords{};
bool outArmed{};
int64_t dumps{};
};
struct ps2_sio2PadState{
bool config{};
bool analog{};
bool locked{};
std::array<uint8_t,6> actMap{};
};
struct ps2_iopDMAChan{
uint32_t madr{};
uint32_t bcr{};
uint32_t chcr{};
uint32_t tadr{};
};
struct ps2_iopDMADone{
uint64_t at{};
int64_t ch{};
};
struct ps2_iopTimer{
uint32_t count{};
uint32_t mode{};
uint32_t target{};
bool fired{};
};
struct ps2_ioTouch{
uint32_t addr{};
uint32_t pc{};
bool write{};
};
struct ps2_isRunWrite{
uint32_t pc{};
uint32_t val{};
bool ought{};
};
struct ps2_IOP{
Slice<uint8_t> ram{};
Slice<uint8_t> spr{};
mips_CPU* CPU{};
ps2_Machine* ps2{};
Slice<ps2_IOPModule*> modules{};
Map<std::string,ps2_IRXExport*> exports{};
Map<std::string,std::string> owner{};
uint32_t allocPtr{};
Slice<ps2_iopCall*> calls{};
Map<ps2_iopBinding,uint32_t> bound{};
Map<uint32_t,std::string> stubName{};
Slice<uint8_t> tty{};
std::array<ps2_iopHandler,64> handlers{};
uint64_t imask{};
bool intrEnabled{};
Slice<ps2_iopBlock> blocks{};
Slice<ps2_iopBlock> freeBlocks{};
uint32_t allocHighPtr{};
Map<uint32_t,ps2_iopHeap*> heaps{};
uint32_t schedSwitch{};
uint32_t schedResched{};
uint64_t pending{};
int64_t inIntr{};
Slice<ps2_iopVblankHandler> vblankHandlers{};
bool vblankPending{};
ps2_sio2xfer sio2{};
ps2_sio2PadState pad{};
std::array<ps2_iopDMAChan,13> dma{};
uint32_t dpcr{};
uint32_t dpcr2{};
uint32_t dicr{};
uint32_t dicr2{};
Slice<ps2_iopDMADone> dmaPending{};
ps2_spu2* spu{};
ps2_cdvd* cdvd{};
std::array<ps2_iopTimer,6> timers{};
uint32_t timerAck{};
uint64_t steps{};
Map<std::string,int64_t> prof{};
Map<uint32_t,Anon38> logPC{};
Map<std::string,int64_t> unmodelledCalls{};
Slice<uint32_t> bootCallbacks{};
Map<uint32_t,uint32_t> io{};
Map<uint32_t,int64_t> unmodelledIO{};
std::function<void(uint32_t,uint32_t,bool,uint32_t)> OnIO{};
uint32_t WatchLo{};
uint32_t WatchHi{};
std::function<void(uint32_t,uint32_t,uint32_t)> OnWrite{};
std::function<void(std::string,std::array<uint32_t,4>,uint32_t)> OnCall{};
std::function<void(ps2_IOPIntrEvent)> OnIntrState{};
Slice<ps2_ioTouch> ioPend{};
uint32_t lastPC{};
std::array<uint32_t,24> trail{};
int64_t trailN{};
uint32_t schedIsRun{};
std::array<ps2_isRunWrite,40> isRunLog{};
int64_t isRunLogN{};
uint32_t Trap{};
std::string TrapSym{};
bool running{};
std::array<int64_t,64> raised{};
std::array<int64_t,64> delivered{};
int64_t switches{};
int64_t callDepth{};
bool Halted{};
std::string HaltReason{};
};
struct ps2_IOPIntrEvent{
uint64_t Step{};
std::string Kind{};
uint32_t PC{};
uint32_t RA{};
uint32_t Addr{};
uint32_t Val{};
bool Enabled{};
int64_t Depth{};
int64_t InIntr{};
};
struct ps2_cdvd{
ps2_Machine* ps2{};
uint8_t nCommand{};
Slice<uint8_t> nParams{};
uint8_t nStatus{};
uint8_t nError{};
Slice<uint8_t> lastParams{};
uint8_t nMode{};
bool nBusy{};
uint64_t nDoneAt{};
uint8_t sCommand{};
Slice<uint8_t> sParams{};
Slice<uint8_t> sResult{};
uint8_t intr{};
Slice<uint8_t> data{};
uint32_t dmaMadr{};
uint32_t dmaLen{};
bool dmaArmed{};
Map<uint8_t,int64_t> unknownN{};
Map<uint8_t,int64_t> unknownS{};
};
struct ps2_CDVDState{
uint8_t NCommand{};
uint8_t NStatus{};
uint8_t NError{};
uint8_t NMode{};
uint8_t Intr{};
Slice<uint8_t> NParams{};
Slice<uint8_t> LastParams{};
bool NBusy{};
uint64_t NDoneAt{};
uint8_t SCommand{};
Slice<uint8_t> SParams{};
Slice<uint8_t> SResult{};
Slice<uint8_t> Data{};
uint32_t DMAMadr{};
uint32_t DMALen{};
bool DMAArmed{};
};
struct ps2_iopHeap{
uint32_t chunk{};
uint32_t base{};
uint32_t size{};
uint32_t ptr{};
uint32_t total{};
};
struct ps2_iopCall{
std::string name{};
std::function<void(ps2_IOP*)> fn{};
};
struct ps2_iopLibrary{
Map<uint16_t,ps2_iopFunc> funcs{};
};
struct ps2_iopFunc{
std::string name{};
std::function<void(ps2_IOP*)> fn{};
};
struct ps2_iopBinding{
std::string library{};
uint16_t id{};
bool operator==(const ps2_iopBinding&)const=default;
auto operator<=>(const ps2_iopBinding&)const=default;
};
namespace std {template<>struct hash<ps2_iopBinding>{size_t operator()(const ps2_iopBinding&v)const{size_t h=0;h=h*31+std::hash<std::string>{}(v.library);h=h*31+std::hash<uint16_t>{}(v.id);return h;}};}
struct ps2_IOPModule{
std::string Name{};
uint32_t Base{};
uint32_t Size{};
ps2_IRX* IRX{};
};
struct ps2_spuVoice{
bool playing{};
uint64_t acc{};
uint64_t lastStep{};
};
struct ps2_spu2{
Slice<uint8_t> regs{};
Slice<uint8_t> ram{};
std::array<std::array<ps2_spuVoice,24>,2> voice{};
};
struct ps2_iopThread{
uint32_t tcb{};
uint32_t prio{};
uint32_t state{};
uint32_t entry{};
uint32_t stack{};
uint32_t frame{};
uint32_t savedPC{};
uint32_t waitTyp{};
uint32_t waitObj{};
uint32_t waitMsk{};
bool running{};
};
struct ps2_IRXImport{
std::string Library{};
uint16_t Version{};
uint32_t Addr{};
Slice<uint16_t> IDs{};
Slice<uint32_t> Stubs{};
};
struct ps2_IRXExport{
std::string Library{};
uint16_t Version{};
uint32_t Addr{};
Slice<uint32_t> Entries{};
};
struct ps2_IRXReloc{
uint32_t Offset{};
uint8_t Type{};
};
struct ps2_IRX{
std::string Name{};
uint16_t Version{};
uint32_t Entry{};
uint32_t GP{};
Slice<uint8_t> Image{};
uint32_t MemSz{};
Slice<ps2_IRXImport> Imports{};
Slice<ps2_IRXExport> Exports{};
Slice<ps2_IRXReloc> Relocs{};
Slice<ps2_Symbol> Symbols{};
};
struct ps2_kernel_401_kv{
std::string name{};
int64_t n{};
};
struct ps2_syscallEntry{
std::string name{};
std::function<void(ps2_Machine*)> fn{};
};
struct ps2_machine_810_unit{
int64_t hits{};
Map<uint32_t,int64_t> regs{};
};
struct ps2_machine_826_kv{
std::string name{};
ps2_machine_810_unit* u{};
};
struct ps2_sifPacket{
uint32_t dest{};
Slice<uint8_t> data{};
bool cmd{};
};
struct ps2_PadPress{
uint16_t Buttons{};
uint32_t At{};
uint32_t Hold{};
};
struct ps2_padLiveState{
uint16_t buttons{};
int8_t lx{};
int8_t ly{};
int8_t rx{};
int8_t ry{};
};
struct ps2_profCounters{
uint64_t eeSteps{};
uint64_t iopSteps{};
uint64_t prims{};
uint64_t frags{};
uint64_t rejZ{};
uint64_t rejScissor{};
uint64_t rejAlpha{};
uint64_t rejDate{};
};
struct ps2_ProfileBucket{
std::string Name{};
double Millis{};
int64_t Count{};
};
struct ps2_ProfileCounter{
std::string Name{};
int64_t Value{};
};
struct ps2_FrameProfile{
double TotalMs{};
Slice<ps2_ProfileBucket> Buckets{};
Slice<ps2_ProfileCounter> Counters{};
bool Drew{};
};
struct ps2_profState{
std::array<int64_t,3> ns{};
std::array<int64_t,3> count{};
time_Time runStart{};
bool inRun{};
int64_t frameNs{};
time_Time drainStart{};
int64_t drainDepth{};
ps2_profCounters base{};
ps2_FrameProfile last{};
bool has{};
};
struct ps2_Machine{
Slice<uint8_t> ram{};
Slice<uint8_t> spram{};
Slice<uint8_t> iopRAM{};
r5900_CPU* CPU{};
Map<uint32_t,uint32_t> io{};
Map<uint32_t,int64_t> unmodelled{};
ps2_Executable* exe{};
Slice<ps2_Symbol> extraSyms{};
iso9660_Volume* vol{};
Slice<uint8_t> bios{};
Map<std::string,int64_t> SyscallCalls{};
Slice<uint8_t> tty{};
uint32_t heapPtr{};
uint32_t heapEnd{};
Map<uint32_t,ps2_thread*> threads{};
uint32_t nextThreadID{};
uint32_t currentThread{};
Slice<ps2_handler> intcHandlers{};
Slice<ps2_handler> dmacHandlers{};
uint32_t dEnable{};
Slice<int64_t> dmacIRQPending{};
uint32_t intcMask{};
uint32_t intcStat{};
uint32_t dmacMask{};
uint32_t dmacStat{};
std::array<ps2_dmacChan,10> dmac{};
uint32_t dCtrl{};
uint32_t dPcr{};
uint32_t dSqwc{};
uint32_t dRbsr{};
uint32_t dRbor{};
ps2_GS* gs{};
ps2_gsWorkPool* rasterPool{};
bool SingleThreaded{};
bool gsWeave{};
std::array<ps2_weaveField,2> weaveRing{};
int8_t weaveLast{};
std::array<ps2_vif*,2> vifs{};
uint32_t vsyncFlagPtr{};
uint32_t vsyncFlag2Ptr{};
uint32_t gsInterlace{};
uint32_t gsVideoMode{};
uint32_t gsFieldMode{};
uint32_t gsIMR{};
Map<uint32_t,ps2_sema*> semas{};
uint32_t nextSemaID{};
Map<uint32_t,uint32_t> userSyscalls{};
uint32_t argc{};
uint32_t argv{};
uint32_t deci2Sockets{};
Map<uint32_t,uint32_t> deci2Desc{};
Map<uint32_t,uint32_t> sifRegs{};
Map<uint32_t,int64_t> sifUnmodelledReg{};
uint32_t sifDmaID{};
uint32_t sifCmdBuf{};
uint32_t sifCmdHandler{};
Slice<ps2_sifPacket> sifToIOPQueue{};
ps2_IOP* IOP{};
std::function<void(ps2_IOP*)> OnIOPStart{};
std::function<void(ps2_IOP*,std::string)> OnIOPModule{};
int64_t sifToIOPCount{};
int64_t sifFromIOPCount{};
std::array<uint32_t,8> sbus{};
Map<uint32_t,int64_t> sifSent{};
Map<uint32_t,int64_t> sifBack{};
Map<uint32_t,int64_t> rpcBinds{};
Map<ps2_sifRPCKey,int64_t> rpcCalls{};
bool idle{};
std::array<ps2_eeTimer,4> eeTimers{};
uint64_t steps{};
uint64_t stores{};
uint64_t eeDisturbGen{};
ps2_idleState idleDet{};
bool noIdleSkip{};
uint32_t vblanks{};
std::string imageHash{};
uint32_t WatchLo{};
uint32_t WatchHi{};
std::function<void(uint32_t,uint32_t,uint32_t)> OnWrite{};
uint32_t RWatchLo{};
uint32_t RWatchHi{};
std::function<void(uint32_t,uint32_t,uint32_t)> OnRead{};
std::function<void(ps2_Machine*,uint32_t)> OnStep{};
bool hookMuted{};
int64_t GSVertDump{};
int64_t GSBigDump{};
int32_t GSPixelX{};
int32_t GSPixelY{};
int64_t GSPixelN{};
uint8_t GSRegLog{};
int64_t GSRegLogN{};
Map<uint8_t,int64_t> GSRegLogs{};
uint8_t GSPktArmReg{};
uint64_t GSPktArmVal{};
int64_t GSPktTraceN{};
int64_t GSPktTraceOn{};
bool GSRegDumpPacket{};
uint64_t GSRegDumpVal{};
int64_t VU1DumpIn{};
bool vif1Pending{};
int64_t VIFTinyN{};
uint32_t feedMadr{};
Slice<ps2_PadPress> PadScript{};
ps2_padLiveState padLive{};
Map<uint32_t,bool> breakpoints{};
bool StopRequested{};
bool Profile{};
ps2_profState prof{};
std::function<void(ps2_Machine*)> OnVBlank{};
std::function<int64_t(int64_t,std::string)> OnGSPrim{};
std::function<void(int64_t,uint32_t,int32_t,int32_t)> OnGSPixel{};
std::function<void(uint32_t,uint32_t,uint32_t)> OnGSFlip{};
bool LogDISPFB{};
Slice<uint8_t> iopTTYLine{};
std::string iopRebootImage{};
uint64_t iopRebootAt{};
Map<uint32_t,uint32_t> IOPPokes{};
Slice<std::string> Log{};
Map<std::string,bool> logSeen{};
bool Halted{};
std::string HaltReason{};
uint64_t rrVblAcc{},rrIopAcc{};
};
struct ps2_RomEntry{
std::string Name{};
Slice<uint8_t> Data{};
};
struct ps2_Result{
uint64_t Steps{};
uint32_t PC{};
std::string Reason{};
};
using ps2_threadState=int64_t;
struct ps2_thread{
uint32_t id{};
uint32_t entry{};
uint32_t stack{};
uint32_t stackSz{};
uint32_t gp{};
uint32_t priority{};
ps2_threadState state{};
int32_t wakeupCount{};
r5900_State ctx{};
};
struct ps2_sema{
uint32_t id{};
int32_t count{};
int32_t maxCount{};
Slice<uint32_t> waiting{};
};
struct ps2_sifRPCKey{
uint32_t sid{};
uint32_t fno{};
bool operator==(const ps2_sifRPCKey&)const=default;
auto operator<=>(const ps2_sifRPCKey&)const=default;
};
namespace std {template<>struct hash<ps2_sifRPCKey>{size_t operator()(const ps2_sifRPCKey&v)const{size_t h=0;h=h*31+std::hash<uint32_t>{}(v.sid);h=h*31+std::hash<uint32_t>{}(v.fno);return h;}};}
struct ps2_vif{
int64_t idx{};
ps2_Machine* m{};
uint32_t cl{};
uint32_t wl{};
uint32_t mode{};
uint32_t mask{};
std::array<uint32_t,4> row{};
std::array<uint32_t,4> col{};
uint32_t base{};
uint32_t ofst{};
uint32_t itop{};
uint32_t mark{};
uint32_t tops{};
vu_VU* vu{};
uint64_t vuSteps{};
bool kickDumped{};
uint32_t lastStart{};
int64_t dumpN{};
int64_t runawayDumps{};
int64_t maxStores{};
int64_t maxUnpacks{};
uint32_t cmd{};
int64_t pending{};
Slice<uint8_t> buf{};
uint32_t payloadAddr{};
int64_t tinyUnpacks{};
Slice<std::string> codeLog{};
int64_t codeTrailsN{};
Slice<uint8_t> micro{};
Slice<uint8_t> data{};
Map<std::string,int64_t> census{};
Map<uint32_t,int64_t> mscal{};
Map<uint64_t,ps2_mpgInfo*> mpgSeen{};
};
struct ps2_mpgInfo{
uint32_t addr{};
int64_t size{};
int64_t count{};
};
struct ps2_countKV{
std::string name{};
int64_t n{};
};
struct Anon0{uint32_t PageMask{};uint64_t EntryHi{};uint64_t EntryLo0{};uint64_t EntryLo1{};};
struct Anon1{std::array<r5900_Quad,32> R{};uint64_t HI{};uint64_t LO{};uint64_t HI1{};uint64_t LO1{};uint32_t SA{};uint64_t PC{};uint64_t nextPC{};std::array<uint64_t,32> COP0{};std::array<r5900_TLBEntry,48> TLB{};std::array<uint32_t,32> FPR{};uint32_t ACC{};uint32_t FCR31{};vu_VU* COP2{};bool LLBit{};bool Halted{};std::string HaltReason{};uint64_t Steps{};std::function<bool(r5900_CPU*)> Syscall{};ps2_Machine* bus{};std::function<uint32_t(uint32_t)> fetch{};uint64_t curPC{};bool delaySlot{};bool pendingDelay{};uint64_t branchAddr{};uint64_t countFrac{};};
struct Anon10{std::array<uint32_t,32> R{};std::array<uint32_t,32> Out{};uint32_t HI{};uint32_t LO{};uint32_t PC{};uint32_t NextPC{};std::array<uint32_t,32> COP0{};bool Halted{};std::string HaltReason{};uint64_t Steps{};uint32_t CurPC{};uint32_t LdReg{};uint32_t LdVal{};bool DelaySlot{};bool PendingDelay{};uint32_t BranchAddr{};};
struct Anon11{std::array<uint32_t,32> Data{};std::array<uint32_t,32> Ctrl{};};
struct Anon12{std::string Upper{};std::string Lower{};bool I{};bool E{};uint64_t Raw{};};
struct Anon13{std::array<std::array<uint32_t,4>,32> VF{};std::array<uint16_t,16> VI{};std::array<float,4> ACC{};float Q{};float P{};float I{};uint32_t R{};Slice<uint8_t> Micro{};Slice<uint8_t> Data{};uint32_t PC{};uint16_t Mac{};uint16_t Status{};uint32_t Clip{};std::array<vu_flagVals,4> flagPipe{};uint16_t visMac{};uint16_t visStatus{};uint32_t visClip{};uint16_t Top{};uint16_t ITop{};std::function<void(uint32_t)> XGKick{};std::function<void(vu_VU*,uint32_t,uint64_t)> Trace{};std::function<void(uint32_t,uint32_t,uint32_t)> OnMaxStore{};uint16_t CMSAR0{};std::function<void(uint32_t)> StartVU1{};uint64_t Steps{};Slice<std::array<uint32_t,2>> BranchLog{};int64_t branchNext{};};
struct Anon14{uint16_t mac{};uint16_t status{};uint32_t clip{};};
struct Anon15{uint8_t kind{};uint32_t reg{};uint32_t mask{};std::array<float,4> val{};std::array<uint32_t,4> bits{};uint32_t clip{};};
struct Anon16{uint16_t Mac{};uint16_t Status{};uint32_t Clip{};};
struct Anon17{std::array<std::array<uint32_t,4>,32> VF{};std::array<uint16_t,16> VI{};std::array<float,4> ACC{};float Q{};float P{};float I{};uint32_t R{};uint32_t PC{};uint16_t Mac{};uint16_t Status{};uint32_t Clip{};std::array<vu_FlagVals,4> FlagPipe{};uint16_t VisMac{};uint16_t VisStatus{};uint32_t VisClip{};uint16_t Top{};uint16_t ITop{};uint16_t CMSAR0{};uint64_t Steps{};};
struct Anon18{int64_t SectorSize{};int64_t DataOffset{};};
struct Anon19{LocalSource* ra{};int64_t size{};iso9660_Geometry geom{};Slice<uint8_t> buf{};};
struct Anon2{std::string name{};r5900_mmiForm form{};};
struct Anon20{LocalSource* src{};std::string System{};std::string Name{};int64_t Blocks{};int64_t rootLBA{};int64_t rootSize{};};
struct Anon21{std::string Name{};std::string Path{};bool IsDir{};int64_t Size{};int64_t Block{};};
struct Anon22{uint32_t chcr{};uint32_t madr{};uint32_t qwc{};uint32_t tadr{};uint32_t asr0{};uint32_t asr1{};uint32_t sadr{};};
struct Anon23{uint32_t base{};uint64_t baseSteps{};uint32_t mode{};uint32_t comp{};uint32_t hold{};};
struct Anon24{uint32_t VAddr{};Slice<uint8_t> Data{};uint32_t MemSz{};};
struct Anon25{std::string Name{};uint32_t Addr{};uint32_t Size{};bool Func{};};
struct Anon26{uint32_t Entry{};Slice<ps2_Segment> Segments{};Slice<ps2_Symbol> Symbols{};Slice<ps2_Symbol> byAddr{};};
struct Anon27{Slice<uint8_t> vram{};std::array<uint64_t,128> reg{};ps2_gsXfer xfer{};uint64_t csr{};int64_t uploads{};int64_t prims{};int64_t primsRunBase{};std::array<ps2_gsVertex,3> vq{};int64_t vqN{};uint32_t q{};std::array<int64_t,8> primCount{};int64_t curCmd{};Map<std::string,int64_t> drawCensus{};std::string src{};Slice<uint8_t> curPacket{};Slice<uint8_t> srcData{};Slice<uint8_t> srcIn{};Slice<uint8_t> srcMicro{};Slice<uint8_t> srcVUData{};bool srcDumped{};std::array<uint32_t,512> clut{};uint32_t cbp0{};uint32_t cbp1{};uint64_t plotted{};uint64_t rejScissor{};uint64_t rejZ{};uint64_t rejAlpha{};uint64_t rejDate{};std::array<uint64_t,8> plotNonBlack{};uint64_t parFills{};uint64_t serFills{};uint64_t rgbaqA0{};uint64_t rgbaqA{};uint64_t rgbaqRGB0{};uint64_t rgbaqRGB{};int64_t t8Dumped{};uint64_t texBlack{};uint64_t texColor{};std::array<uint64_t,64> texBlackPSM{};std::array<uint64_t,64> texColorPSM{};int64_t path2ImageRemain{};int64_t path2SkipRemain{};Slice<uint8_t> path2Carry{};ps2_Machine* m{};};
struct Anon28{bool active{};uint32_t dbp{};uint32_t dbw{};uint32_t dpsm{};uint32_t dsax{};uint32_t dsay{};uint32_t rrw{};uint32_t rrh{};uint32_t x{};uint32_t y{};Slice<uint8_t> partial{};};
struct Anon29{Slice<uint8_t> pix{};int64_t w{};int64_t h{};};
struct Anon3{uint64_t Lo{};uint64_t Hi{};};
struct Anon30{int32_t x{};int32_t y{};uint32_t z{};uint32_t rgba{};int32_t u{};int32_t v{};float s{};float t{};float q{};uint32_t f{};};
struct Anon31{uint32_t fbp{};uint32_t fbw{};uint32_t psm{};uint32_t fbmsk{};int32_t sx0{};int32_t sy0{};int32_t sx1{};int32_t sy1{};bool abe{};uint64_t alpha{};bool pabe{};bool colclamp{};uint32_t fba{};uint64_t test{};bool fge{};uint32_t fogcol{};uint32_t zbp{};uint32_t zpsm{};bool zmsk{};uint32_t ztst{};int64_t primType{};};
struct Anon32{uint32_t tbp{};uint32_t tbw{};uint32_t psm{};uint32_t tw{};uint32_t th{};bool tcc{};uint32_t tfx{};uint32_t cbp{};uint32_t cpsm{};uint32_t csm{};uint32_t csa{};uint32_t cld{};};
struct Anon33{ps2_GS* gs{};ps2_gsTex tex{};int32_t w{};int32_t h{};uint32_t wms{};uint32_t wmt{};int32_t minu{};int32_t maxu{};int32_t minv{};int32_t maxv{};bool linear{};bool probe{};bool parallel{};};
struct Anon34{uint64_t PC{};uint64_t NextPC{};uint64_t BranchAddr{};uint64_t HI{};uint64_t LO{};uint64_t HI1{};uint64_t LO1{};uint32_t SA{};uint32_t ACC{};uint32_t FCR31{};bool DelaySlot{};bool PendingDelay{};bool LLBit{};std::array<r5900_Quad,32> R{};std::array<uint32_t,32> FPR{};};
struct Anon35{bool armed{};ps2_idleSnap snap{};uint64_t stores{};int64_t insns{};int64_t period{};uint64_t next{};Slice<r5900_State> phase{};uint64_t Skipped{};uint64_t Hits{};};
struct Anon36{uint32_t cause{};uint32_t addr{};uint32_t arg{};uint32_t next{};};
struct Anon37{Slice<uint8_t> ram{};Slice<uint8_t> spr{};mips_CPU* CPU{};ps2_Machine* ps2{};Slice<ps2_IOPModule*> modules{};Map<std::string,ps2_IRXExport*> exports{};Map<std::string,std::string> owner{};uint32_t allocPtr{};Slice<ps2_iopCall*> calls{};Map<ps2_iopBinding,uint32_t> bound{};Map<uint32_t,std::string> stubName{};Slice<uint8_t> tty{};std::array<ps2_iopHandler,64> handlers{};uint64_t imask{};bool intrEnabled{};Slice<ps2_iopBlock> blocks{};Slice<ps2_iopBlock> freeBlocks{};uint32_t allocHighPtr{};Map<uint32_t,ps2_iopHeap*> heaps{};uint32_t schedSwitch{};uint32_t schedResched{};uint64_t pending{};int64_t inIntr{};Slice<ps2_iopVblankHandler> vblankHandlers{};bool vblankPending{};ps2_sio2xfer sio2{};ps2_sio2PadState pad{};std::array<ps2_iopDMAChan,13> dma{};uint32_t dpcr{};uint32_t dpcr2{};uint32_t dicr{};uint32_t dicr2{};Slice<ps2_iopDMADone> dmaPending{};ps2_spu2* spu{};ps2_cdvd* cdvd{};std::array<ps2_iopTimer,6> timers{};uint32_t timerAck{};uint64_t steps{};Map<std::string,int64_t> prof{};Map<uint32_t,Anon38> logPC{};Map<std::string,int64_t> unmodelledCalls{};Slice<uint32_t> bootCallbacks{};Map<uint32_t,uint32_t> io{};Map<uint32_t,int64_t> unmodelledIO{};std::function<void(uint32_t,uint32_t,bool,uint32_t)> OnIO{};uint32_t WatchLo{};uint32_t WatchHi{};std::function<void(uint32_t,uint32_t,uint32_t)> OnWrite{};std::function<void(std::string,std::array<uint32_t,4>,uint32_t)> OnCall{};std::function<void(ps2_IOPIntrEvent)> OnIntrState{};Slice<ps2_ioTouch> ioPend{};uint32_t lastPC{};std::array<uint32_t,24> trail{};int64_t trailN{};uint32_t schedIsRun{};std::array<ps2_isRunWrite,40> isRunLog{};int64_t isRunLogN{};uint32_t Trap{};std::string TrapSym{};bool running{};std::array<int64_t,64> raised{};std::array<int64_t,64> delivered{};int64_t switches{};int64_t callDepth{};bool Halted{};std::string HaltReason{};};
struct Anon38{};
struct Anon39{uint64_t Step{};std::string Kind{};uint32_t PC{};uint32_t RA{};uint32_t Addr{};uint32_t Val{};bool Enabled{};int64_t Depth{};int64_t InIntr{};};
struct Anon4{uint32_t Addr{};int64_t Len{};std::string Mnem{};std::string Text{};r5900_Flow Flow{};uint32_t Target{};bool HasTarget{};bool HasDelay{};bool Annul{};};
struct Anon40{uint32_t pc{};uint32_t val{};bool ought{};};
struct Anon41{std::string name{};int64_t n{};};
struct Anon42{uint32_t addr{};uint32_t pc{};bool write{};};
struct Anon43{ps2_Machine* ps2{};uint8_t nCommand{};Slice<uint8_t> nParams{};uint8_t nStatus{};uint8_t nError{};Slice<uint8_t> lastParams{};uint8_t nMode{};bool nBusy{};uint64_t nDoneAt{};uint8_t sCommand{};Slice<uint8_t> sParams{};Slice<uint8_t> sResult{};uint8_t intr{};Slice<uint8_t> data{};uint32_t dmaMadr{};uint32_t dmaLen{};bool dmaArmed{};Map<uint8_t,int64_t> unknownN{};Map<uint8_t,int64_t> unknownS{};};
struct Anon44{std::string what{};Map<uint8_t,int64_t> m{};};
struct Anon45{uint8_t NCommand{};uint8_t NStatus{};uint8_t NError{};uint8_t NMode{};uint8_t Intr{};Slice<uint8_t> NParams{};Slice<uint8_t> LastParams{};bool NBusy{};uint64_t NDoneAt{};uint8_t SCommand{};Slice<uint8_t> SParams{};Slice<uint8_t> SResult{};Slice<uint8_t> Data{};uint32_t DMAMadr{};uint32_t DMALen{};bool DMAArmed{};};
struct Anon46{uint32_t madr{};uint32_t bcr{};uint32_t chcr{};uint32_t tadr{};};
struct Anon47{uint64_t at{};int64_t ch{};};
struct Anon48{uint32_t fn{};uint32_t arg{};};
struct Anon49{uint32_t base{};uint32_t size{};};
struct Anon5{std::array<r5900_Quad,32> R{};uint64_t HI{};uint64_t LO{};uint64_t HI1{};uint64_t LO1{};uint32_t SA{};uint64_t PC{};uint64_t NextPC{};std::array<uint64_t,32> COP0{};std::array<r5900_TLBEntry,48> TLB{};std::array<uint32_t,32> FPR{};uint32_t ACC{};uint32_t FCR31{};bool LLBit{};bool Halted{};std::string HaltReason{};uint64_t Steps{};uint64_t CurPC{};bool DelaySlot{};bool PendingDelay{};uint64_t BranchAddr{};uint64_t CountFrac{};};
struct Anon50{uint32_t chunk{};uint32_t base{};uint32_t size{};uint32_t ptr{};uint32_t total{};};
struct Anon51{std::string name{};std::function<void(ps2_IOP*)> fn{};};
struct Anon52{Map<uint16_t,ps2_iopFunc> funcs{};};
struct Anon53{std::string library{};uint16_t id{};};
struct Anon54{std::string Name{};uint32_t Base{};uint32_t Size{};ps2_IRX* IRX{};};
struct Anon55{Slice<uint8_t> in{};Slice<uint8_t> out{};std::array<uint32_t,2> dmaAddr{};std::array<uint32_t,2> dmaWords{};bool outArmed{};int64_t dumps{};};
struct Anon56{bool config{};bool analog{};bool locked{};std::array<uint8_t,6> actMap{};};
struct Anon57{bool playing{};uint64_t acc{};uint64_t lastStep{};};
struct Anon58{Slice<uint8_t> regs{};Slice<uint8_t> ram{};std::array<std::array<ps2_spuVoice,24>,2> voice{};};
struct Anon59{uint32_t tcb{};uint32_t prio{};uint32_t state{};uint32_t entry{};uint32_t stack{};uint32_t frame{};uint32_t savedPC{};uint32_t waitTyp{};uint32_t waitObj{};uint32_t waitMsk{};bool running{};};
struct Anon6{uint32_t reg{};uint32_t val{};};
struct Anon60{uint32_t count{};uint32_t mode{};uint32_t target{};bool fired{};};
struct Anon61{uint32_t edge{};uint32_t prio{};uint32_t fn{};uint32_t arg{};};
struct Anon62{std::string Library{};uint16_t Version{};uint32_t Addr{};Slice<uint16_t> IDs{};Slice<uint32_t> Stubs{};};
struct Anon63{std::string Library{};uint16_t Version{};uint32_t Addr{};Slice<uint32_t> Entries{};};
struct Anon64{uint32_t Offset{};uint8_t Type{};};
struct Anon65{std::string Name{};uint16_t Version{};uint32_t Entry{};uint32_t GP{};Slice<uint8_t> Image{};uint32_t MemSz{};Slice<ps2_IRXImport> Imports{};Slice<ps2_IRXExport> Exports{};Slice<ps2_IRXReloc> Relocs{};Slice<ps2_Symbol> Symbols{};};
struct Anon66{std::string name{};std::function<void(ps2_Machine*)> fn{};};
struct Anon67{Slice<uint8_t> ram{};Slice<uint8_t> spram{};Slice<uint8_t> iopRAM{};r5900_CPU* CPU{};Map<uint32_t,uint32_t> io{};Map<uint32_t,int64_t> unmodelled{};ps2_Executable* exe{};Slice<ps2_Symbol> extraSyms{};iso9660_Volume* vol{};Slice<uint8_t> bios{};Map<std::string,int64_t> SyscallCalls{};Slice<uint8_t> tty{};uint32_t heapPtr{};uint32_t heapEnd{};Map<uint32_t,ps2_thread*> threads{};uint32_t nextThreadID{};uint32_t currentThread{};Slice<ps2_handler> intcHandlers{};Slice<ps2_handler> dmacHandlers{};uint32_t dEnable{};Slice<int64_t> dmacIRQPending{};uint32_t intcMask{};uint32_t intcStat{};uint32_t dmacMask{};uint32_t dmacStat{};std::array<ps2_dmacChan,10> dmac{};uint32_t dCtrl{};uint32_t dPcr{};uint32_t dSqwc{};uint32_t dRbsr{};uint32_t dRbor{};ps2_GS* gs{};ps2_gsWorkPool* rasterPool{};bool SingleThreaded{};bool gsWeave{};std::array<ps2_weaveField,2> weaveRing{};int8_t weaveLast{};std::array<ps2_vif*,2> vifs{};uint32_t vsyncFlagPtr{};uint32_t vsyncFlag2Ptr{};uint32_t gsInterlace{};uint32_t gsVideoMode{};uint32_t gsFieldMode{};uint32_t gsIMR{};Map<uint32_t,ps2_sema*> semas{};uint32_t nextSemaID{};Map<uint32_t,uint32_t> userSyscalls{};uint32_t argc{};uint32_t argv{};uint32_t deci2Sockets{};Map<uint32_t,uint32_t> deci2Desc{};Map<uint32_t,uint32_t> sifRegs{};Map<uint32_t,int64_t> sifUnmodelledReg{};uint32_t sifDmaID{};uint32_t sifCmdBuf{};uint32_t sifCmdHandler{};Slice<ps2_sifPacket> sifToIOPQueue{};ps2_IOP* IOP{};std::function<void(ps2_IOP*)> OnIOPStart{};std::function<void(ps2_IOP*,std::string)> OnIOPModule{};int64_t sifToIOPCount{};int64_t sifFromIOPCount{};std::array<uint32_t,8> sbus{};Map<uint32_t,int64_t> sifSent{};Map<uint32_t,int64_t> sifBack{};Map<uint32_t,int64_t> rpcBinds{};Map<ps2_sifRPCKey,int64_t> rpcCalls{};bool idle{};std::array<ps2_eeTimer,4> eeTimers{};uint64_t steps{};uint64_t stores{};uint64_t eeDisturbGen{};ps2_idleState idleDet{};bool noIdleSkip{};uint32_t vblanks{};std::string imageHash{};uint32_t WatchLo{};uint32_t WatchHi{};std::function<void(uint32_t,uint32_t,uint32_t)> OnWrite{};uint32_t RWatchLo{};uint32_t RWatchHi{};std::function<void(uint32_t,uint32_t,uint32_t)> OnRead{};std::function<void(ps2_Machine*,uint32_t)> OnStep{};bool hookMuted{};int64_t GSVertDump{};int64_t GSBigDump{};int32_t GSPixelX{};int32_t GSPixelY{};int64_t GSPixelN{};uint8_t GSRegLog{};int64_t GSRegLogN{};Map<uint8_t,int64_t> GSRegLogs{};uint8_t GSPktArmReg{};uint64_t GSPktArmVal{};int64_t GSPktTraceN{};int64_t GSPktTraceOn{};bool GSRegDumpPacket{};uint64_t GSRegDumpVal{};int64_t VU1DumpIn{};bool vif1Pending{};int64_t VIFTinyN{};uint32_t feedMadr{};Slice<ps2_PadPress> PadScript{};ps2_padLiveState padLive{};Map<uint32_t,bool> breakpoints{};bool StopRequested{};bool Profile{};ps2_profState prof{};std::function<void(ps2_Machine*)> OnVBlank{};std::function<int64_t(int64_t,std::string)> OnGSPrim{};std::function<void(int64_t,uint32_t,int32_t,int32_t)> OnGSPixel{};std::function<void(uint32_t,uint32_t,uint32_t)> OnGSFlip{};bool LogDISPFB{};Slice<uint8_t> iopTTYLine{};std::string iopRebootImage{};uint64_t iopRebootAt{};Map<uint32_t,uint32_t> IOPPokes{};Slice<std::string> Log{};Map<std::string,bool> logSeen{};bool Halted{};std::string HaltReason{};};
struct Anon68{int64_t hits{};Map<uint32_t,int64_t> regs{};};
struct Anon69{std::string name{};ps2_machine_810_unit* u{};};
struct Anon7{std::array<uint32_t,32> R{};std::array<uint32_t,32> out{};uint32_t HI{};uint32_t LO{};uint32_t PC{};uint32_t nextPC{};std::array<uint32_t,32> COP0{};RRNullGTE* GTE{};std::function<bool(mips_CPU*)> Syscall{};bool Halted{};std::string HaltReason{};uint64_t Steps{};ps2_IOP* bus{};uint32_t curPC{};mips_loadSlot ld{};bool delaySlot{};bool pendingDelay{};uint32_t branchAddr{};};
struct Anon70{uint16_t Buttons{};uint32_t At{};uint32_t Hold{};};
struct Anon71{uint16_t buttons{};int8_t lx{};int8_t ly{};int8_t rx{};int8_t ry{};};
struct Anon72{std::string Name{};double Millis{};int64_t Count{};};
struct Anon73{std::string Name{};int64_t Value{};};
struct Anon74{double TotalMs{};Slice<ps2_ProfileBucket> Buckets{};Slice<ps2_ProfileCounter> Counters{};bool Drew{};};
struct Anon75{std::array<int64_t,3> ns{};std::array<int64_t,3> count{};time_Time runStart{};bool inRun{};int64_t frameNs{};time_Time drainStart{};int64_t drainDepth{};ps2_profCounters base{};ps2_FrameProfile last{};bool has{};};
struct Anon76{uint64_t eeSteps{};uint64_t iopSteps{};uint64_t prims{};uint64_t frags{};uint64_t rejZ{};uint64_t rejScissor{};uint64_t rejAlpha{};uint64_t rejDate{};};
struct Anon77{std::string Name{};Slice<uint8_t> Data{};};
struct Anon78{uint64_t Steps{};uint32_t PC{};std::string Reason{};};
struct Anon79{uint32_t id{};uint32_t entry{};uint32_t stack{};uint32_t stackSz{};uint32_t gp{};uint32_t priority{};ps2_threadState state{};int32_t wakeupCount{};r5900_State ctx{};};
struct Anon8{std::array<uint32_t,32> data{};std::array<uint32_t,32> ctrl{};};
struct Anon80{uint32_t id{};int32_t count{};int32_t maxCount{};Slice<uint32_t> waiting{};};
struct Anon81{uint32_t dest{};Slice<uint8_t> data{};bool cmd{};};
struct Anon82{uint32_t sid{};uint32_t fno{};};
struct Anon83{uint32_t bit{};std::string name{};};
struct Anon84{std::string way{};Map<uint32_t,int64_t> m{};};
struct Anon85{int64_t idx{};ps2_Machine* m{};uint32_t cl{};uint32_t wl{};uint32_t mode{};uint32_t mask{};std::array<uint32_t,4> row{};std::array<uint32_t,4> col{};uint32_t base{};uint32_t ofst{};uint32_t itop{};uint32_t mark{};uint32_t tops{};vu_VU* vu{};uint64_t vuSteps{};bool kickDumped{};uint32_t lastStart{};int64_t dumpN{};int64_t runawayDumps{};int64_t maxStores{};int64_t maxUnpacks{};uint32_t cmd{};int64_t pending{};Slice<uint8_t> buf{};uint32_t payloadAddr{};int64_t tinyUnpacks{};Slice<std::string> codeLog{};int64_t codeTrailsN{};Slice<uint8_t> micro{};Slice<uint8_t> data{};Map<std::string,int64_t> census{};Map<uint32_t,int64_t> mscal{};Map<uint64_t,ps2_mpgInfo*> mpgSeen{};};
struct Anon86{uint32_t addr{};int64_t size{};int64_t count{};};
struct Anon9{uint32_t Addr{};int64_t Len{};std::string Mnem{};std::string Text{};mips_Flow Flow{};uint32_t Target{};bool HasTarget{};bool HasDelay{};};

#include "adapters-decl.h"
uint64_t r5900_TLBEntry_pairSize(r5900_TLBEntry* e);
uint64_t r5900_TLBEntry_pageSize(r5900_TLBEntry* e);
std::tuple<uint32_t,bool> r5900_CPU_Translate(r5900_CPU* c,uint64_t vaddr,bool store);
std::tuple<uint32_t,bool> r5900_CPU_tlbTranslate(r5900_CPU* c,uint64_t vaddr,bool store);
void r5900_CPU_tlbException(r5900_CPU* c,uint64_t vaddr,bool store,bool refill);
void r5900_CPU_setEntryHiVPN(r5900_CPU* c,uint64_t vaddr);
void r5900_CPU_setFaultAddress(r5900_CPU* c,uint64_t vaddr);
void r5900_CPU_tlbr(r5900_CPU* c);
void r5900_CPU_tlbw(r5900_CPU* c,uint64_t i);
void r5900_CPU_SetTLB(r5900_CPU* c,int64_t i,r5900_TLBEntry e);
void r5900_CPU_tlbp(r5900_CPU* c);
uint64_t r5900_CPU_random(r5900_CPU* c);
uint64_t r5900_CPU_readCop0(r5900_CPU* c,uint32_t i);
void r5900_CPU_writeCop0(r5900_CPU* c,uint32_t i,uint64_t v);
void r5900_CPU_cop0(r5900_CPU* c,uint32_t w,uint32_t rs,uint32_t rt,uint32_t rd);
std::string r5900_CPU_DumpTLB(r5900_CPU* c);
void r5900_CPU_Reset(r5900_CPU* c);
void r5900_CPU_SetPC(r5900_CPU* c,uint64_t pc);
void r5900_CPU_SetReg(r5900_CPU* c,uint32_t i,uint64_t v);
uint64_t r5900_CPU_Reg(r5900_CPU* c,uint32_t i);
void r5900_CPU_SetQuad(r5900_CPU* c,uint32_t i,r5900_Quad q);
r5900_Quad r5900_CPU_Quad(r5900_CPU* c,uint32_t i);
uint64_t r5900_CPU_CurPC(r5900_CPU* c);
uint64_t r5900_CPU_NextPC(r5900_CPU* c);
bool r5900_CPU_InDelaySlot(r5900_CPU* c);
bool r5900_CPU_PendingDelay(r5900_CPU* c);
uint64_t r5900_CPU_BranchAddr(r5900_CPU* c);
void r5900_CPU_set(r5900_CPU* c,uint32_t i,uint64_t v);
void r5900_CPU_setQ(r5900_CPU* c,uint32_t i,uint64_t lo,uint64_t hi);
uint64_t r5900_sext32(uint32_t v);
uint32_t r5900_CPU_read8(r5900_CPU* c,uint32_t a);
void r5900_CPU_write8(r5900_CPU* c,uint32_t a,uint32_t v);
uint32_t r5900_CPU_read16(r5900_CPU* c,uint32_t a);
void r5900_CPU_write16(r5900_CPU* c,uint32_t a,uint32_t v);
uint32_t r5900_CPU_read32(r5900_CPU* c,uint32_t a);
void r5900_CPU_write32(r5900_CPU* c,uint32_t a,uint32_t v);
uint64_t r5900_CPU_read64(r5900_CPU* c,uint32_t a);
void r5900_CPU_write64(r5900_CPU* c,uint32_t a,uint64_t v);
r5900_Quad r5900_CPU_read128(r5900_CPU* c,uint32_t a);
void r5900_CPU_write128(r5900_CPU* c,uint32_t a,r5900_Quad q);
void r5900_CPU_Exception(r5900_CPU* c,uint32_t code);
void r5900_CPU_exceptionAt(r5900_CPU* c,uint32_t code,bool tlbRefill);
uint64_t r5900_sext64(uint64_t v);
void r5900_CPU_eret(r5900_CPU* c);
void r5900_CPU_coprocessorUnusable(r5900_CPU* c,uint32_t unit);
void r5900_CPU_addrError(r5900_CPU* c,uint32_t code,uint64_t vaddr);
bool r5900_CPU_Interrupt(r5900_CPU* c,bool int0,bool int1);
bool r5900_CPU_checkInterrupt(r5900_CPU* c);
void r5900_CPU_tickCount(r5900_CPU* c);
void r5900_CPU_SkipInstructions(r5900_CPU* c,uint64_t n);
uint32_t r5900_CPU_Count(r5900_CPU* c);
uint32_t r5900_CPU_Compare(r5900_CPU* c);
uint64_t r5900_CPU_CountFrac(r5900_CPU* c);
bool r5900_CPU_TimerIRQDeliverable(r5900_CPU* c);
bool r5900_CPU_InterruptDeliverable(r5900_CPU* c);
std::string r5900_reg(uint32_t i);
std::string r5900_freg(uint32_t i);
std::string r5900_vfreg(uint32_t i);
std::string r5900_cop0Reg(uint32_t i);
r5900_Inst r5900_Decode(Slice<uint8_t> code,uint32_t addr);
r5900_Inst r5900_DecodeWord(uint32_t w,uint32_t addr);
r5900_Inst r5900_decodeSpecial(r5900_Inst in,uint32_t w,uint32_t rs,uint32_t rt,uint32_t rd,uint32_t shamt);
r5900_Inst r5900_decodeRegimm(r5900_Inst in,uint32_t w,uint32_t rs,uint32_t rt,uint32_t branchT,int32_t simm);
r5900_Inst r5900_decodeCop0(r5900_Inst in,uint32_t w,uint32_t rs,uint32_t rt,uint32_t rd);
r5900_Inst r5900_decodeCop1(r5900_Inst in,uint32_t w,uint32_t rs,uint32_t rt,uint32_t rd,uint32_t shamt,uint32_t branchT);
r5900_Inst r5900_decodeCop2(r5900_Inst in,uint32_t w,uint32_t rs,uint32_t rt,uint32_t rd,uint32_t branchT);
r5900_Inst r5900_call(r5900_Inst in);
r5900_Inst r5900_word(r5900_Inst in,uint32_t w);
Slice<std::string> r5900_Disassemble(Slice<uint8_t> code,uint32_t base);
int64_t r5900_CPU_Step(r5900_CPU* c);
std::tuple<uint32_t,bool> r5900_CPU_translateFetch(r5900_CPU* c,uint64_t vaddr);
void r5900_CPU_doBranch(r5900_CPU* c,bool taken,uint64_t target);
void r5900_CPU_doBranchLikely(r5900_CPU* c,bool taken,uint64_t target);
void r5900_CPU_execute(r5900_CPU* c,uint32_t w);
void r5900_CPU_cop2(r5900_CPU* c,uint32_t w,uint32_t rs,uint32_t rt,uint32_t rd,uint64_t branchT);
void r5900_CPU_cop2Mem(r5900_CPU* c,uint32_t op,uint32_t rs,uint32_t rt,uint64_t simm);
void r5900_CPU_special(r5900_CPU* c,uint32_t w,uint32_t rs,uint32_t rt);
void r5900_CPU_regimm(r5900_CPU* c,uint32_t w,uint32_t rs,uint32_t rt,uint64_t branchT,uint64_t simm);
void r5900_CPU_trapIf(r5900_CPU* c,bool cond);
void r5900_CPU_loadOp(r5900_CPU* c,uint32_t op,uint32_t rs,uint32_t rt,uint64_t simm);
void r5900_CPU_storeOp(r5900_CPU* c,uint32_t op,uint32_t rs,uint32_t rt,uint64_t simm);
std::tuple<uint32_t,bool> r5900_addOv32(uint32_t a,uint32_t b);
std::tuple<uint32_t,bool> r5900_subOv32(uint32_t a,uint32_t b);
std::tuple<uint64_t,bool> r5900_addOv64(uint64_t a,uint64_t b);
std::tuple<uint64_t,bool> r5900_subOv64(uint64_t a,uint64_t b);
uint64_t r5900_b2u(bool b);
double r5900_fdec(uint32_t b);
uint32_t r5900_fenc(double v);
double r5900_fdiv(double a,double b);
double r5900_CPU_f(r5900_CPU* c,uint32_t i);
void r5900_CPU_setF(r5900_CPU* c,uint32_t i,double v);
void r5900_CPU_cop1(r5900_CPU* c,uint32_t w,uint32_t rs,uint32_t rt,uint32_t rd,uint32_t shamt,uint64_t branchT);
void r5900_CPU_cop1Single(r5900_CPU* c,uint32_t w,uint32_t ft,uint32_t fs,uint32_t fd);
void r5900_CPU_setCond(r5900_CPU* c,bool b);
uint32_t r5900_cvtW(double v);
void r5900_CPU_cop1Mem(r5900_CPU* c,uint32_t op,uint32_t rs,uint32_t rt,uint64_t simm);
std::array<uint32_t,4> r5900_words(r5900_Quad q);
r5900_Quad r5900_fromWords(std::array<uint32_t,4> w);
std::array<uint16_t,8> r5900_halves(r5900_Quad q);
r5900_Quad r5900_fromHalves(std::array<uint16_t,8> h);
std::array<uint8_t,16> r5900_octets(r5900_Quad q);
r5900_Quad r5900_fromOctets(std::array<uint8_t,16> b);
uint32_t r5900_satS32(int64_t v);
uint16_t r5900_satS16(int32_t v);
uint8_t r5900_satS8(int32_t v);
uint32_t r5900_satU32(int64_t v);
uint16_t r5900_satU16(int32_t v);
uint8_t r5900_satU8(int32_t v);
r5900_Quad r5900_CPU_loQ(r5900_CPU* c);
r5900_Quad r5900_CPU_hiQ(r5900_CPU* c);
void r5900_CPU_setLoQ(r5900_CPU* c,r5900_Quad q);
void r5900_CPU_setHiQ(r5900_CPU* c,r5900_Quad q);
void r5900_CPU_mmi(r5900_CPU* c,uint32_t w,uint32_t rs,uint32_t rt,uint32_t rd,uint32_t shamt);
void r5900_CPU_maddAcc(r5900_CPU* c,uint32_t rd,int64_t p,bool second);
uint32_t r5900_lzcw(uint32_t v);
std::tuple<uint64_t,uint64_t> r5900_divSigned32(int32_t a,int32_t b);
std::tuple<uint64_t,uint64_t> r5900_divUnsigned32(uint32_t a,uint32_t b);
void r5900_CPU_mmi0(r5900_CPU* c,uint32_t rs,uint32_t rt,uint32_t rd,uint32_t sub_);
void r5900_CPU_mmi1(r5900_CPU* c,uint32_t rs,uint32_t rt,uint32_t rd,uint32_t sub_);
void r5900_CPU_mmi2(r5900_CPU* c,uint32_t rs,uint32_t rt,uint32_t rd,uint32_t sub_);
void r5900_CPU_mmi3(r5900_CPU* c,uint32_t rs,uint32_t rt,uint32_t rd,uint32_t sub_);
void r5900_CPU_pmulW(r5900_CPU* c,uint32_t rd,r5900_Quad s,r5900_Quad t,bool accumulate,bool subtract,bool signed_);
void r5900_CPU_pmulH(r5900_CPU* c,uint32_t rd,r5900_Quad s,r5900_Quad t,bool accumulate,bool subtract);
void r5900_CPU_phmulH(r5900_CPU* c,uint32_t rd,r5900_Quad s,r5900_Quad t,bool subtract);
std::array<uint32_t,8> r5900_CPU_accWords(r5900_CPU* c);
void r5900_CPU_setAccWords(r5900_CPU* c,std::array<uint32_t,8> p);
void r5900_CPU_pmfhl(r5900_CPU* c,uint32_t rd,uint32_t sub_);
void r5900_CPU_pmthl(r5900_CPU* c,uint32_t rs,uint32_t sub_);
uint64_t r5900_satS64(int64_t v);
uint32_t r5900_maskIf(bool b);
uint32_t r5900_absW(uint32_t v);
uint16_t r5900_absH(uint16_t v);
r5900_Quad r5900_funnelRight(r5900_Quad hi,r5900_Quad lo,uint32_t n);
r5900_Quad r5900_shiftRight128(r5900_Quad q,uint32_t n);
r5900_Quad r5900_shiftLeft128(r5900_Quad q,uint32_t n);
void r5900_CPU_putW(r5900_CPU* c,uint32_t rd,std::array<uint32_t,4> v);
void r5900_CPU_putH(r5900_CPU* c,uint32_t rd,std::array<uint16_t,8> v);
void r5900_CPU_putB(r5900_CPU* c,uint32_t rd,std::array<uint8_t,16> v);
void r5900_CPU_putQ(r5900_CPU* c,uint32_t rd,r5900_Quad q);
r5900_Inst r5900_decodeMMI(r5900_Inst in,uint32_t w,uint32_t rs,uint32_t rt,uint32_t rd,uint32_t shamt);
r5900_Inst r5900_mmiSub(r5900_Inst in,uint32_t w,Map<uint32_t,r5900_mmiOp> tab,uint32_t rs,uint32_t rt,uint32_t rd,uint32_t shamt);
std::string r5900_Quad_String(r5900_Quad q);
std::string r5900_Flow_String(r5900_Flow f);
std::string r5900_Inst_String(r5900_Inst in);
r5900_State r5900_CPU_Snapshot(r5900_CPU* c);
void r5900_CPU_SetThreadRegs(r5900_CPU* c,r5900_State s);
void r5900_CPU_Restore(r5900_CPU* c,r5900_State s);
void mips_CPU_Reset(mips_CPU* c);
void mips_CPU_SetPC(mips_CPU* c,uint32_t pc);
void mips_CPU_SetReg(mips_CPU* c,uint32_t i,uint32_t v);
uint32_t mips_CPU_Reg(mips_CPU* c,uint32_t i);
uint32_t mips_CPU_CurPC(mips_CPU* c);
void mips_CPU_set(mips_CPU* c,uint32_t i,uint32_t v);
uint32_t mips_CPU_read8(mips_CPU* c,uint32_t a);
void mips_CPU_write8(mips_CPU* c,uint32_t a,uint32_t v);
uint32_t mips_CPU_read16(mips_CPU* c,uint32_t a);
void mips_CPU_write16(mips_CPU* c,uint32_t a,uint32_t v);
uint32_t mips_CPU_read32(mips_CPU* c,uint32_t a);
void mips_CPU_write32(mips_CPU* c,uint32_t a,uint32_t v);
void mips_CPU_Exception(mips_CPU* c,uint32_t code);
void mips_CPU_rfe(mips_CPU* c);
bool mips_CPU_Interrupt(mips_CPU* c,bool pending);
void mips_CPU_addrError(mips_CPU* c,uint32_t code,uint32_t addr);
std::string mips_reg(uint32_t i);
mips_Inst mips_Decode(Slice<uint8_t> code,uint32_t addr);
mips_Inst mips_decodeSpecial(mips_Inst in,uint32_t w,uint32_t rs,uint32_t rt,uint32_t rd,uint32_t shamt,uint32_t funct);
mips_Inst mips_decodeCop0(mips_Inst in,uint32_t w,uint32_t rs,uint32_t rt,uint32_t rd);
mips_Inst mips_decodeCop2(mips_Inst in,uint32_t w,uint32_t rs,uint32_t rt,uint32_t rd);
mips_Inst mips_word(mips_Inst in,uint32_t w);
Slice<std::string> mips_Disassemble(Slice<uint8_t> code,uint32_t base);
int64_t mips_CPU_Step(mips_CPU* c);
uint32_t mips_CPU_reg(mips_CPU* c,uint32_t i);
void mips_CPU_load(mips_CPU* c,uint32_t reg,uint32_t val);
void mips_CPU_doBranch(mips_CPU* c,bool taken,uint32_t target);
void mips_CPU_execute(mips_CPU* c,uint32_t w);
void mips_CPU_special(mips_CPU* c,uint32_t w,uint32_t rs,uint32_t rt,uint32_t rd,uint32_t shamt);
void mips_CPU_divSigned(mips_CPU* c,int32_t a,int32_t b);
void mips_CPU_divUnsigned(mips_CPU* c,uint32_t a,uint32_t b);
void mips_CPU_loadOp(mips_CPU* c,uint32_t op,uint32_t rs,uint32_t rt,uint32_t simm);
void mips_CPU_storeOp(mips_CPU* c,uint32_t op,uint32_t rs,uint32_t rt,uint32_t simm);
void mips_CPU_cop0(mips_CPU* c,uint32_t w,uint32_t rs,uint32_t rt,uint32_t rd);
void mips_CPU_cop2(mips_CPU* c,uint32_t w,uint32_t rs,uint32_t rt,uint32_t rd);
std::tuple<uint32_t,bool> mips_addOv(uint32_t a,uint32_t b);
std::tuple<uint32_t,bool> mips_subOv(uint32_t a,uint32_t b);
uint32_t mips_b2u(bool b);
mips_GTE* mips_NewGTE();
int32_t mips_s16(uint32_t v);
uint32_t mips_GTE_Read(mips_GTE* g,uint32_t reg);
void mips_GTE_Write(mips_GTE* g,uint32_t reg,uint32_t v);
uint32_t mips_GTE_ReadCtrl(mips_GTE* g,uint32_t reg);
void mips_GTE_WriteCtrl(mips_GTE* g,uint32_t reg,uint32_t v);
int32_t mips_clamp5(int32_t v);
uint32_t mips_lzc(uint32_t v);
std::tuple<int32_t,int32_t,int32_t> mips_GTE_vec(mips_GTE* g,int64_t n);
int32_t mips_GTE_rt(mips_GTE* g,int64_t r,int64_t c);
int32_t mips_GTE_light(mips_GTE* g,int64_t r,int64_t c);
int32_t mips_GTE_color(mips_GTE* g,int64_t r,int64_t c);
int32_t mips_matEntry(Slice<uint32_t> w,int64_t r,int64_t c);
int32_t mips_GTE_tr(mips_GTE* g,int64_t i);
int32_t mips_GTE_bk(mips_GTE* g,int64_t i);
int32_t mips_GTE_fc(mips_GTE* g,int64_t i);
int32_t mips_GTE_ofx(mips_GTE* g);
int32_t mips_GTE_ofy(mips_GTE* g);
uint32_t mips_GTE_h(mips_GTE* g);
int32_t mips_GTE_dqa(mips_GTE* g);
int32_t mips_GTE_dqb(mips_GTE* g);
int32_t mips_GTE_zsf3(mips_GTE* g);
int32_t mips_GTE_zsf4(mips_GTE* g);
int32_t mips_GTE_irVal(mips_GTE* g,int64_t i);
void mips_GTE_setFlag(mips_GTE* g,uint32_t bits);
void mips_GTE_clearFlags(mips_GTE* g);
void mips_GTE_finishFlags(mips_GTE* g);
void mips_GTE_macCheck(mips_GTE* g,int64_t i,int64_t val);
void mips_GTE_setMac(mips_GTE* g,int64_t i,int32_t v);
void mips_GTE_setIR(mips_GTE* g,int64_t i,int32_t v,bool lm);
void mips_GTE_setMacIR(mips_GTE* g,int64_t i,int64_t val,uint64_t sf,bool lm);
void mips_GTE_setIR0(mips_GTE* g,int32_t v);
void mips_GTE_setMac0(mips_GTE* g,int64_t val);
void mips_GTE_pushSZ3(mips_GTE* g,int32_t v);
void mips_GTE_setOTZ(mips_GTE* g,int32_t v);
void mips_GTE_pushSXY(mips_GTE* g,int32_t x,int32_t y);
int32_t mips_clampSXY(int32_t v,uint32_t bit,mips_GTE* g);
void mips_GTE_Command(mips_GTE* g,uint32_t cmd);
void mips_GTE_rtp(mips_GTE* g,int64_t n,uint64_t sf,bool lm,bool depth);
void mips_GTE_nclip(mips_GTE* g);
void mips_GTE_avsz(mips_GTE* g,bool four);
void mips_GTE_mvmva(mips_GTE* g,uint32_t cmd,uint64_t sf,bool lm);
void mips_GTE_normalColor(mips_GTE* g,int64_t n,uint64_t sf,bool lm,bool primary,bool depth);
void mips_GTE_normalColorT(mips_GTE* g,uint64_t sf,bool lm,bool primary,bool depth);
void mips_GTE_lightVector(mips_GTE* g,int64_t n,uint64_t sf,bool lm);
void mips_GTE_lightColor(mips_GTE* g,uint64_t sf,bool lm);
void mips_GTE_primaryColor(mips_GTE* g,uint64_t sf,bool lm,bool depth);
void mips_GTE_interpolate(mips_GTE* g,uint64_t sf,bool lm,std::array<int64_t,3> start);
void mips_GTE_pushColorFIFO(mips_GTE* g);
int32_t mips_GTE_clampColor(mips_GTE* g,int32_t v,uint64_t bit);
uint32_t mips_GTE_divide(mips_GTE* g,uint32_t h,uint32_t sz3);
uint32_t mips_lzc16(uint16_t v);
std::string mips_Flow_String(mips_Flow f);
std::string mips_Inst_String(mips_Inst in);
mips_CPUState mips_CPU_SaveState(mips_CPU* c);
void mips_CPU_LoadState(mips_CPU* c,mips_CPUState s);
mips_GTEState mips_GTE_SaveState(mips_GTE* g);
void mips_GTE_LoadState(mips_GTE* g,mips_GTEState s);
vu_Inst vu_Decode(uint64_t raw,uint32_t addr);
std::string vu_DisasmMacro(uint32_t w);
std::string vu_dest(uint32_t i);
uint32_t vu_ft(uint32_t i);
uint32_t vu_fs(uint32_t i);
uint32_t vu_fd(uint32_t i);
std::string vu_decodeUpper(uint32_t i);
std::string vu_decodeUpperWide(uint32_t i);
std::string vu_decodeLower(uint32_t i,uint32_t addr);
std::string vu_decodeLowerSpecial(uint32_t i);
void vu_VU_ResetBranchLog(vu_VU* v,int64_t size);
int64_t vu_VU_BranchCount(vu_VU* v);
void vu_VU_logBranch(vu_VU* v,uint32_t from,uint32_t to);
Slice<std::array<uint32_t,2>> vu_VU_BranchTrail(vu_VU* v);
vu_VU* vu_New(Slice<uint8_t> micro,Slice<uint8_t> data);
std::tuple<int64_t,bool> vu_VU_Run(vu_VU* v,uint32_t start,int64_t maxSteps);
float vu_VU_getF(vu_VU* v,uint32_t r,uint32_t lane);
void vu_VU_setLanes(vu_VU* v,uint32_t r,uint32_t destMask,std::array<float,4> val);
void vu_VU_setBits(vu_VU* v,uint32_t r,uint32_t destMask,std::array<uint32_t,4> val);
void vu_VU_setVI(vu_VU* v,uint32_t r,uint16_t val);
float vu_sane(float f);
uint64_t vu_le64m(Slice<uint8_t> b);
void vu_VU_execUpper(vu_VU* v,uint32_t i,vu_upperResult* res);
void vu_VU_commitUpper(vu_VU* v,vu_upperResult* res);
void vu_VU_macFlags(vu_VU* v,uint32_t destMask,std::array<float,4> val);
int64_t vu_VU_execLower(vu_VU* v,uint32_t i);
int64_t vu_VU_execLowerSpecial(vu_VU* v,uint32_t i);
void vu_VU_loadQ(vu_VU* v,uint32_t r,uint32_t destMask,uint32_t addr);
void vu_VU_storeQ(vu_VU* v,uint32_t r,uint32_t destMask,uint32_t addr);
uint32_t vu_laneOf(uint32_t destMask);
float vu_vuDiv(float a,float b);
float vu_dot3(std::array<uint32_t,4> r);
void vu_VU_rNext(vu_VU* v);
uint16_t vu_b2u(bool b);
void vu_VU_Macro(vu_VU* v,uint32_t w);
uint32_t vu_VU_ReadVF(vu_VU* v,uint32_t reg,uint32_t field);
void vu_VU_WriteVF(vu_VU* v,uint32_t reg,uint32_t field,uint32_t val);
uint32_t vu_VU_ReadCtrl(vu_VU* v,uint32_t reg);
void vu_VU_WriteCtrl(vu_VU* v,uint32_t reg,uint32_t val);
vu_State vu_VU_Snapshot(vu_VU* v);
void vu_VU_Restore(vu_VU* v,vu_State s);
float vu_float32frombits(uint32_t b);
uint32_t vu_float32bits(float f);
std::string iso9660_Geometry_String(iso9660_Geometry g);
iso9660_Geometry iso9660_Source_Geometry(iso9660_Source* s);
std::tuple<Slice<uint8_t>,Error> iso9660_Volume_ReadBlock(iso9660_Volume* v,int64_t n);
std::string iso9660_Entry_String(iso9660_Entry e);
std::tuple<iso9660_Volume*,Error> iso9660_OpenVolume(LocalSource* src);
std::tuple<iso9660_Volume*,Error> iso9660_OpenVolumeAt(LocalSource* src,int64_t pvd);
uint32_t iso9660_le32(Slice<uint8_t> b);
std::tuple<Slice<iso9660_Entry>,Error> iso9660_Volume_dirEntries(iso9660_Volume* v,int64_t lba,int64_t size,std::string dirPath);
bool iso9660_nameEqual(std::string entry,std::string want);
Slice<std::string> iso9660_splitPath(std::string p);
std::tuple<iso9660_Entry,Error> iso9660_Volume_Resolve(iso9660_Volume* v,std::string path);
std::string iso9660_normalisePath(std::string p);
std::tuple<int64_t,int64_t,bool> iso9660_parseLbnPath(std::string path);
std::tuple<Slice<iso9660_Entry>,Error> iso9660_Volume_ReadDir(iso9660_Volume* v,std::string path);
Error iso9660_Volume_Walk(iso9660_Volume* v,std::function<Error(iso9660_Entry)> fn);
Error iso9660_Volume_walk(iso9660_Volume* v,int64_t lba,int64_t size,std::string dirPath,std::function<Error(iso9660_Entry)> fn);
std::tuple<int64_t,Error> iso9660_Volume_ReadFileAt(iso9660_Volume* v,iso9660_Entry e,int64_t off,Slice<uint8_t> p);
std::tuple<Slice<uint8_t>,Error> iso9660_Volume_ReadFile(iso9660_Volume* v,std::string path);
std::tuple<iso9660_Entry,bool> iso9660_Volume_FileAt(iso9660_Volume* v,int64_t lba);
iso9660_byteReaderAt iso9660_newByteReaderAt(Slice<uint8_t> b);
std::tuple<int64_t,Error> iso9660_byteReaderAt_ReadAt(iso9660_byteReaderAt b,Slice<uint8_t> p,int64_t off);
std::tuple<int64_t,uint32_t,bool> ps2_dmacChanReg(uint32_t a);
void ps2_Machine_drainVIF1(ps2_Machine* m);
std::tuple<uint32_t,bool> ps2_Machine_dmacRead(ps2_Machine* m,uint32_t a);
bool ps2_Machine_dmacWrite(ps2_Machine* m,uint32_t a,uint32_t v);
void ps2_Machine_dmacStart(ps2_Machine* m,int64_t ch);
void ps2_Machine_dmacComplete(ps2_Machine* m,int64_t ch);
void ps2_Machine_queueDmacIRQ(ps2_Machine* m,int64_t ch);
void ps2_Machine_dmacRetrigger(ps2_Machine* m);
void ps2_Machine_deliverDmacIRQs(ps2_Machine* m);
void ps2_Machine_dmacSPR(ps2_Machine* m,ps2_dmacChan* c,bool fromSPR);
void ps2_Machine_dmacSPRInterleave(ps2_Machine* m,ps2_dmacChan* c,bool fromSPR);
void ps2_Machine_dmacSPRChainTo(ps2_Machine* m,ps2_dmacChan* c);
void ps2_Machine_dmacSourceChain(ps2_Machine* m,int64_t ch,ps2_dmacChan* c,std::function<void(Slice<uint8_t>)> feed,bool wholeTag);
Slice<uint8_t> ps2_Machine_dmaBytes(ps2_Machine* m,uint32_t madr,uint32_t qwc);
uint64_t ps2_eeTimer_rate(ps2_eeTimer* t);
uint32_t ps2_eeTimer_countAt(ps2_eeTimer* t,uint64_t steps);
void ps2_eeTimer_rebase(ps2_eeTimer* t,uint64_t steps);
std::tuple<int64_t,uint32_t> ps2_eeTimerAt(uint32_t p);
std::tuple<uint32_t,bool> ps2_Machine_eeTimerRead(ps2_Machine* m,uint32_t p);
bool ps2_Machine_eeTimerWrite(ps2_Machine* m,uint32_t p,uint32_t v);
std::tuple<std::string,uint32_t,bool> ps2_Executable_Lookup(ps2_Executable* e,uint32_t addr);
std::string ps2_Executable_Describe(ps2_Executable* e);
std::tuple<uint32_t,Slice<uint8_t>> ps2_Executable_Flat(ps2_Executable* e);
void ps2_Machine_gifStart(ps2_Machine* m,ps2_dmacChan* c);
void ps2_Machine_gifStream(ps2_Machine* m,Slice<uint8_t> data);
void ps2_Machine_gifPacket(ps2_Machine* m,Slice<uint8_t> data);
void ps2_Machine_gifPacked(ps2_Machine* m,ps2_GS* gs,uint64_t desc,uint64_t lo,uint64_t hi);
int64_t ps2_gifPacketLen(Slice<uint8_t> data);
uint64_t ps2_le64(Slice<uint8_t> b);
ps2_GS* ps2_Machine_ensureGS(ps2_Machine* m);
void ps2_GS_write(ps2_GS* gs,uint8_t reg,uint64_t val);
void ps2_GS_writePacked(ps2_GS* gs,uint8_t reg,uint64_t lo,uint64_t hi);
void ps2_GS_beginTransfer(ps2_GS* gs,uint64_t dir);
void ps2_GS_localCopy(ps2_GS* gs,uint64_t bitbltbuf,uint64_t trxpos,uint64_t trxreg);
std::tuple<uint32_t,bool> ps2_GS_readTexel(ps2_GS* gs,uint32_t psm,uint32_t bp,uint32_t bw,uint32_t x,uint32_t y);
void ps2_GS_writeTexel(ps2_GS* gs,uint32_t psm,uint32_t bp,uint32_t bw,uint32_t x,uint32_t y,uint32_t v);
void ps2_GS_imageData(ps2_GS* gs,Slice<uint8_t> data);
uint32_t ps2_addrPSMCT32(uint32_t bp,uint32_t bw,uint32_t x,uint32_t y);
std::tuple<uint32_t,bool> ps2_Machine_gsPrivRead(ps2_Machine* m,uint32_t a);
bool ps2_Machine_gsPrivWrite(ps2_Machine* m,uint32_t a,uint32_t v);
void ps2_Machine_SetGSWeave(ps2_Machine* m,bool on);
std::tuple<Slice<uint8_t>,int64_t,int64_t,int64_t> ps2_Machine_GSFrameWoven(ps2_Machine* m);
std::string ps2_Machine_GSStatus(ps2_Machine* m);
std::tuple<Slice<uint8_t>,int64_t,int64_t> ps2_Machine_GSFrame(ps2_Machine* m);
std::tuple<uint32_t,uint32_t,uint32_t,int64_t,int64_t,bool> ps2_Machine_GSScanout(ps2_Machine* m);
std::tuple<Slice<uint8_t>,int64_t> ps2_Machine_GSBuffer(ps2_Machine* m,uint32_t base,uint32_t bw,int64_t h);
std::tuple<Slice<uint8_t>,int64_t> ps2_Machine_GSBufferAlpha(ps2_Machine* m,uint32_t base,uint32_t bw,int64_t h);
void ps2_Machine_gsVSync(ps2_Machine* m);
Slice<uint8_t> ps2_u64bytes(uint64_t v);
uint32_t ps2_le32gs(Slice<uint8_t> b);
Error ps2_writeFile(std::string name,Slice<uint8_t> data);
void ps2_GS_pushVertex(ps2_GS* gs,int32_t x,int32_t y,uint32_t z,uint32_t f,bool kick);
uint64_t ps2_GS_prim(ps2_GS* gs);
int64_t ps2_GS_ctxt(ps2_GS* gs);
void ps2_GS_kick(ps2_GS* gs,bool draw);
time_Time ps2_GS_rasterStart(ps2_GS* gs);
void ps2_GS_rasterEnd(ps2_GS* gs,time_Time t);
void ps2_GS_drawn(ps2_GS* gs,int64_t typ);
void ps2_GS_count(ps2_GS* gs,std::string what);
ps2_gsTarget ps2_GS_target(ps2_GS* gs,uint64_t p);
void ps2_GS_plot(ps2_GS* gs,ps2_gsTarget* t,int32_t x,int32_t y,uint32_t z,uint32_t rgba,ps2_gsStats* st);
uint32_t ps2_fogPixel(uint32_t rgba,uint32_t f,uint32_t fogcol);
void ps2_GS_point(ps2_GS* gs,ps2_gsVertex v);
void ps2_GS_line(ps2_GS* gs,ps2_gsVertex a,ps2_gsVertex b,uint64_t p);
int32_t ps2_texAxis(int32_t pc,int32_t p0,int32_t p1,int32_t uv0,int32_t uv1);
void ps2_GS_sprite(ps2_GS* gs,ps2_gsVertex a,ps2_gsVertex b,uint64_t p);
void ps2_GS_triangle(ps2_GS* gs,ps2_gsVertex v0,ps2_gsVertex v1,ps2_gsVertex v2,uint64_t p);
void ps2_GS_noteFeatures(ps2_GS* gs,uint64_t p);
std::tuple<int64_t,int64_t,int64_t,int64_t> ps2_unpackRGBA(uint32_t c);
int32_t ps2_min3(int32_t a,int32_t b,int32_t c);
int32_t ps2_max3(int32_t a,int32_t b,int32_t c);
ps2_gsTex ps2_decodeTEX0(uint64_t v);
uint32_t ps2_addrPSMT8(uint32_t bp,uint32_t bw,uint32_t x,uint32_t y);
std::tuple<uint32_t,uint32_t> ps2_addrPSMT4(uint32_t bp,uint32_t bw,uint32_t x,uint32_t y);
uint32_t ps2_addrPSMZ32(uint32_t bp,uint32_t bw,uint32_t x,uint32_t y);
uint32_t ps2_addrPSMZ16(uint32_t bp,uint32_t bw,uint32_t x,uint32_t y,bool s);
uint32_t ps2_addrPSMCT16(uint32_t bp,uint32_t bw,uint32_t x,uint32_t y,bool s);
void ps2_GS_clutLoad(ps2_GS* gs,ps2_gsTex t);
uint32_t ps2_GS_clutEntry(ps2_GS* gs,ps2_gsTex* t,uint32_t idx);
uint32_t ps2_GS_expand16(ps2_GS* gs,uint32_t px);
std::optional<ps2_gsSampler> ps2_GS_sampler(ps2_GS* gs,uint64_t p);
uint32_t ps2_gsSampler_pick(ps2_gsSampler* s,int32_t u,int32_t v);
int32_t ps2_wrapTexel(int32_t c,int32_t size,uint32_t mode,int32_t min,int32_t max);
uint32_t ps2_gsSampler_at(ps2_gsSampler* s,int32_t u,int32_t v);
std::tuple<Slice<uint8_t>,int64_t,int64_t> ps2_Machine_GSTexture(ps2_Machine* m,uint64_t tex0);
uint32_t ps2_gsSampler_combine(ps2_gsSampler* s,uint32_t tex,uint32_t frag,ps2_gsStats* st);
int64_t ps2_clamp255(int64_t v);
uint32_t ps2_blendPixel(uint64_t alpha,uint32_t src,uint32_t dst,int64_t dstA,bool colclamp);
ps2_idleSnap ps2_Machine_snapshotEE(ps2_Machine* m);
bool ps2_Machine_idleStep(ps2_Machine* m);
uint64_t ps2_Machine_idleDeadline(ps2_Machine* m,uint64_t vblAcc);
uint64_t ps2_Machine_idleSkip(ps2_Machine* m,uint64_t* vblAcc,uint64_t* iopAcc,uint64_t budget);
void ps2_Machine_captureIdlePhases(ps2_Machine* m,uint64_t P);
void ps2_Machine_SetIdleSkip(ps2_Machine* m,bool on);
bool ps2_Machine_IdleSkip(ps2_Machine* m);
std::tuple<uint64_t,uint64_t> ps2_Machine_IdleStats(ps2_Machine* m);
void ps2_Machine_addIntcHandler(ps2_Machine* m);
void ps2_Machine_addDmacHandler(ps2_Machine* m);
void ps2_Machine_deliverVBlank(ps2_Machine* m);
uint32_t ps2_Machine_callGuest(ps2_Machine* m,uint32_t entry,Slice<uint32_t> args);
void ps2_Machine_raiseINTC(ps2_Machine* m,uint32_t cause);
void ps2_IOP_ieEvent(ps2_IOP* p,std::string kind,uint32_t addr,uint32_t val,uint32_t ra);
std::string ps2_IOP_IOPInterrupts(ps2_IOP* p);
ps2_IOP* ps2_newIOP(ps2_Machine* m,Slice<uint8_t> ram);
uint32_t ps2_iopPhys(uint32_t addr);
uint8_t ps2_IOP_Read(ps2_IOP* p,uint32_t addr);
void ps2_IOP_Write(ps2_IOP* p,uint32_t addr,uint8_t v);
uint32_t ps2_IOP_ioPeek(ps2_IOP* p,uint32_t a);
uint32_t ps2_IOP_Read32(ps2_IOP* p,uint32_t addr);
void ps2_IOP_Write32(ps2_IOP* p,uint32_t addr,uint32_t v);
std::string ps2_IOP_CString(ps2_IOP* p,uint32_t addr);
uint32_t ps2_IOP_ioRead(ps2_IOP* p,uint32_t a);
void ps2_IOP_ioWrite(ps2_IOP* p,uint32_t a,uint32_t v);
uint32_t ps2_IOP_alloc(ps2_IOP* p,uint32_t n);
void ps2_IOP_installIdle(ps2_IOP* p);
void ps2_IOP_Step(ps2_IOP* p);
void ps2_IOP_LogPC(ps2_IOP* p,uint32_t addr);
std::string ps2_IOP_TTY(ps2_IOP* p);
std::string ps2_IOP_DisasmAt(ps2_IOP* p,uint32_t addr);
uint32_t ps2_IOP_stubTarget(ps2_IOP* p,uint32_t addr);
std::string ps2_IOP_callName(ps2_IOP* p,uint32_t addr);
Slice<ps2_IOPModule*> ps2_IOP_Modules(ps2_IOP* p);
std::string ps2_IOP_Sym(ps2_IOP* p,uint32_t addr);
std::string ps2_IOP_SymGrep(ps2_IOP* p,std::string substr);
std::tuple<uint32_t,bool> ps2_IOP_SymAddr(ps2_IOP* p,std::string name);
void ps2_IOP_resolveTrap(ps2_IOP* p);
std::string ps2_IOP_IOPTrail(ps2_IOP* p);
std::string ps2_IOP_isRunHistory(ps2_IOP* p);
std::string ps2_IOP_threadLabel(ps2_IOP* p,uint32_t tcb);
std::string ps2_IOP_IOPProfile(ps2_IOP* p);
std::string ps2_IOP_IOPCensus(ps2_IOP* p);
std::string ps2_IOPRegionName(uint32_t a);
std::string ps2_IOP_symFunc(ps2_IOP* p,uint32_t addr);
int64_t ps2_indexByte(std::string s,uint8_t c);
void ps2_IOP_ioTrace(ps2_IOP* p,uint32_t a,bool write);
void ps2_IOP_ioTraceFlush(ps2_IOP* p);
void ps2_IOP_Run(ps2_IOP* p,uint64_t n);
uint64_t ps2_IOP_Steps(ps2_IOP* p);
bool ps2_IOP_BusyToEE(ps2_IOP* p);
Error ps2_Machine_RebootIOP(ps2_Machine* m);
Error ps2_Machine_RebootIOPFrom(ps2_Machine* m,std::string image);
std::tuple<std::string,Error> ps2_iopRebootImage(std::string cmd);
Error ps2_IOP_LoadModuleFromDisc(ps2_IOP* p,std::string path);
void ps2_IOP_runBootCallbacks(ps2_IOP* p);
Error ps2_IOP_LoadAndStart(ps2_IOP* p,std::string name,Slice<uint8_t> raw);
ps2_cdvd* ps2_newCDVD(ps2_Machine* m);
uint8_t ps2_cdvd_discType(ps2_cdvd* c);
uint8_t ps2_discTypeFor(iso9660_Geometry g,bool geomKnown,int64_t blocks);
bool ps2_cdvd_contains(ps2_cdvd* c,uint32_t a);
uint8_t ps2_cdvd_read(ps2_cdvd* c,uint32_t a);
uint8_t ps2_cdvd_peek(ps2_cdvd* c,uint32_t a);
void ps2_cdvd_write(ps2_cdvd* c,uint32_t a,uint8_t v);
void ps2_cdvd_startN(ps2_cdvd* c,uint8_t cmd);
bool ps2_cdvd_exec(ps2_cdvd* c,uint8_t cmd,Slice<uint8_t> params);
void ps2_cdvd_readSectors(ps2_cdvd* c,uint32_t lba,uint32_t count,bool dvd);
uint32_t ps2_le32(Slice<uint8_t> b);
void ps2_cdvd_startS(ps2_cdvd* c,uint8_t cmd);
bool ps2_cdvd_execS(ps2_cdvd* c,uint8_t cmd,Slice<uint8_t> params);
void ps2_cdvd_tick(ps2_cdvd* c,ps2_IOP* p);
void ps2_cdvd_arm(ps2_cdvd* c,uint32_t madr,uint32_t n);
void ps2_cdvd_pump(ps2_cdvd* c,ps2_IOP* p);
std::string ps2_cdvd_census(ps2_cdvd* c);
ps2_CDVDState ps2_cdvd_saveState(ps2_cdvd* c);
void ps2_cdvd_loadState(ps2_cdvd* c,ps2_CDVDState s);
uint8_t ps2_mode8(Slice<uint8_t> p);
std::string ps2_hexBytes(Slice<uint8_t> b);
std::tuple<int64_t,uint32_t,bool> ps2_iopDMAReg(uint32_t a);
std::tuple<uint32_t,bool> ps2_IOP_dmaRead(ps2_IOP* p,uint32_t a);
bool ps2_IOP_dmaWrite(ps2_IOP* p,uint32_t a,uint32_t v);
bool ps2_IOP_dmaEnabled(ps2_IOP* p,int64_t ch);
void ps2_IOP_dmaStart(ps2_IOP* p,int64_t ch);
void ps2_IOP_dmaTick(ps2_IOP* p);
uint32_t ps2_iopDMAIRQ(int64_t ch);
uint32_t ps2_IOP_findThreadExit(ps2_IOP* p);
void ps2_IOP_exitLoaderThread(ps2_IOP* p);
void ps2_IOP_saveFrame(ps2_IOP* p,uint32_t at);
void ps2_IOP_loadFrame(ps2_IOP* p,uint32_t at);
void ps2_init_iopintr();
void ps2_IOP_intrRegister(ps2_IOP* p);
void ps2_IOP_intrRelease(ps2_IOP* p);
uint32_t ps2_IOP_intrLine(ps2_IOP* p,uint32_t arg);
void ps2_IOP_intrEnable(ps2_IOP* p);
void ps2_IOP_intrDisable(ps2_IOP* p);
void ps2_IOP_intrSuspend(ps2_IOP* p);
void ps2_IOP_intrResume(ps2_IOP* p);
void ps2_IOP_intrCpuDisable(ps2_IOP* p);
void ps2_IOP_intrCpuEnable(ps2_IOP* p);
void ps2_IOP_intrSetSwitchHook(ps2_IOP* p);
void ps2_IOP_intrSetReschedHook(ps2_IOP* p);
void ps2_IOP_deriveSchedIsRun(ps2_IOP* p);
void ps2_IOP_intrQueryContext(ps2_IOP* p);
void ps2_IOP_intrInvokeInKmode(ps2_IOP* p);
void ps2_IOP_raiseIRQ(ps2_IOP* p,uint32_t irq);
void ps2_IOP_serviceIntr(ps2_IOP* p);
void ps2_IOP_intrDeliver(ps2_IOP* p,uint32_t irq);
std::tuple<uint32_t,bool> ps2_IOP_intrReschedule(ps2_IOP* p,uint32_t frame);
void ps2_IOP_tick(ps2_IOP* p);
uint32_t ps2_b2u(bool b);
void ps2_IOP_sysmemAlloc(ps2_IOP* p);
uint32_t ps2_IOP_allocHigh(ps2_IOP* p,uint32_t size);
uint32_t ps2_IOP_allocReuse(ps2_IOP* p,uint32_t size,bool high);
void ps2_IOP_sysmemFree(ps2_IOP* p);
void ps2_IOP_freeInsert(ps2_IOP* p,ps2_iopBlock b);
void ps2_IOP_sysmemBlockTop(ps2_IOP* p);
void ps2_IOP_sysmemBlockSize(ps2_IOP* p);
std::tuple<ps2_iopBlock,bool> ps2_IOP_blockOf(ps2_IOP* p,uint32_t ptr);
void ps2_IOP_sysmemKprintf(ps2_IOP* p);
void ps2_IOP_heapCreate(ps2_IOP* p);
void ps2_IOP_heapAlloc(ps2_IOP* p);
void ps2_IOP_clibMemcpy(ps2_IOP* p);
void ps2_IOP_clibMemmove(ps2_IOP* p);
void ps2_IOP_clibBzero(ps2_IOP* p);
void ps2_IOP_clibStrcmp(ps2_IOP* p);
void ps2_IOP_clibStrncpy(ps2_IOP* p);
void ps2_IOP_clibCtype(ps2_IOP* p);
void ps2_IOP_clibStrncmp(ps2_IOP* p);
void ps2_IOP_clibStrlen(ps2_IOP* p);
void ps2_IOP_clibStrcat(ps2_IOP* p);
void ps2_IOP_clibStrcpy(ps2_IOP* p);
void ps2_IOP_clibStrchr(ps2_IOP* p);
void ps2_IOP_clibMemset(ps2_IOP* p);
bool ps2_IOP_kernelSyscall(ps2_IOP* p);
void ps2_IOP_yield(ps2_IOP* p);
void ps2_lib(std::string name,Map<uint16_t,ps2_iopFunc> funcs);
ps2_iopFunc ps2_unknown();
void ps2_init_iopkernel();
void ps2_IOP_registerLibraries(ps2_IOP* p);
bool ps2_IOP_hasGoLibrary(ps2_IOP* p,std::string name);
uint32_t ps2_IOP_bindCall(ps2_IOP* p,std::string library,uint16_t id);
uint32_t ps2_IOP_addCall(ps2_IOP* p,std::string name,std::function<void(ps2_IOP*)> fn);
bool ps2_IOP_handleSyscall(ps2_IOP* p,mips_CPU* c);
uint32_t ps2_IOP_arg(ps2_IOP* p,int64_t i);
void ps2_IOP_setRet(ps2_IOP* p,uint32_t v);
void ps2_IOP_loadcoreRegisterLibrary(ps2_IOP* p);
void ps2_IOP_loadcoreRegisterBootCallback(ps2_IOP* p);
void ps2_IOP_stdioPrintf(ps2_IOP* p);
std::string ps2_IOP_formatArgs(ps2_IOP* p,std::string format,int64_t argi);
uint32_t ps2_insnJ(uint32_t target);
uint32_t ps2_insnSyscall(uint32_t n);
uint32_t ps2_insnNop();
uint32_t ps2_insnJR(uint32_t reg);
uint32_t ps2_regRA();
std::tuple<ps2_IOPModule*,Error> ps2_IOP_LoadIRX(ps2_IOP* p,std::string name,Slice<uint8_t> raw);
std::tuple<ps2_IOPModule*,Error> ps2_IOP_placeAndLink(ps2_IOP* p,std::string name,ps2_IRX* x,uint32_t base);
Error ps2_IOP_link(ps2_IOP* p,ps2_IOPModule* mod);
void ps2_IOP_loadcoreProbeModule(ps2_IOP* p);
void ps2_IOP_loadcoreLinkModule(ps2_IOP* p);
void ps2_IOP_loadcoreLinkCheck(ps2_IOP* p);
void ps2_IOP_loadcoreFlushIcache(ps2_IOP* p);
void ps2_IOP_loadcoreRegisterModule(ps2_IOP* p);
std::tuple<ps2_IRX*,Error> ps2_IOP_readGuestIRX(ps2_IOP* p,uint32_t addr);
std::string ps2_IOP_guestModuleName(ps2_IOP* p,uint32_t addr,ps2_IRX* x);
uint32_t ps2_IOP_exportAddr(ps2_IOP* p,std::string library,uint16_t id);
std::tuple<uint32_t,Error> ps2_IOP_Start(ps2_IOP* p,ps2_IOPModule* mod,Slice<uint32_t> args);
std::tuple<uint32_t,Error> ps2_IOP_callGuest(ps2_IOP* p,uint32_t entry);
std::tuple<uint32_t,Error> ps2_IOP_callGuestOn(ps2_IOP* p,uint32_t entry,uint32_t stack);
std::string ps2_IOP_reason(ps2_IOP* p);
bool ps2_sio2Contains(uint32_t a);
uint32_t ps2_IOP_sio2Read(ps2_IOP* p,uint32_t a);
void ps2_IOP_sio2Write(ps2_IOP* p,uint32_t a,uint32_t v);
void ps2_IOP_sio2Start(ps2_IOP* p);
Slice<uint8_t> ps2_IOP_sio2Pad(ps2_IOP* p,Slice<uint8_t> cmd,int64_t n);
void ps2_IOP_dmacmanSetSlice(ps2_IOP* p);
void ps2_IOP_dmacmanStart(ps2_IOP* p);
void ps2_IOP_dmacmanChanSetup(ps2_IOP* p);
void ps2_IOP_dmacmanChanEnable(ps2_IOP* p);
ps2_spu2* ps2_newSPU2();
uint32_t ps2_spu2_read(ps2_spu2* s,uint32_t off);
void ps2_spu2_write(ps2_spu2* s,uint32_t off,uint32_t v);
uint32_t ps2_spu2_half(ps2_spu2* s,uint32_t off);
void ps2_spu2_setHalf(ps2_spu2* s,uint32_t off,uint32_t v);
uint32_t ps2_spu2_tsa(ps2_spu2* s,int64_t core);
void ps2_spu2_dma(ps2_spu2* s,int64_t core,ps2_IOP* p,uint32_t madr,uint32_t n,bool toRAM);
void ps2_spu2_complete(ps2_spu2* s,int64_t core,ps2_IOP* p);
uint32_t ps2_spu2_readReg(ps2_spu2* s,uint32_t off,uint64_t now);
void ps2_spu2_writeReg(ps2_spu2* s,uint32_t off,uint32_t v,uint64_t now);
void ps2_spu2_keyOnOff(ps2_spu2* s,int64_t core,uint32_t r,uint64_t now,bool on);
void ps2_spu2_tick(ps2_spu2* s,uint64_t now);
void ps2_spu2_advance(ps2_spu2* s,int64_t core,int64_t v,uint64_t now);
uint32_t ps2_spu2_wordAddr(ps2_spu2* s,uint32_t off);
void ps2_spu2_setWordAddr(ps2_spu2* s,uint32_t off,uint32_t w);
uint32_t ps2_IOP_currentTCB(ps2_IOP* p);
bool ps2_IOP_looksLikeTCB(ps2_IOP* p,uint32_t a);
bool ps2_IOP_inModule(ps2_IOP* p,uint32_t addr);
std::string ps2_IOP_IOPThreads(ps2_IOP* p);
ps2_iopThread ps2_IOP_readThread(ps2_IOP* p,uint32_t a,uint32_t cur);
uint32_t ps2_iopTimerIRQ(int64_t n);
uint32_t ps2_iopTimerMax(int64_t n);
std::tuple<int64_t,uint32_t,bool> ps2_iopTimerReg(uint32_t a);
std::tuple<uint32_t,bool> ps2_IOP_timerRead(ps2_IOP* p,uint32_t a);
void ps2_IOP_timerAckFlush(ps2_IOP* p);
std::tuple<uint32_t,bool> ps2_IOP_timerPeek(ps2_IOP* p,uint32_t a);
bool ps2_IOP_timerWrite(ps2_IOP* p,uint32_t a,uint32_t v);
void ps2_IOP_timerTick(ps2_IOP* p);
void ps2_IOP_vblankRegister(ps2_IOP* p);
void ps2_IOP_vblankRelease(ps2_IOP* p);
void ps2_IOP_vblankTick(ps2_IOP* p);
void ps2_IOP_vblankDeliver(ps2_IOP* p);
void ps2_IRX_scanLibraries(ps2_IRX* m);
uint32_t ps2_IRX_word(ps2_IRX* m,uint32_t o);
std::string ps2_libraryName(Slice<uint8_t> b);
std::string ps2_cstring(Slice<uint8_t> b);
std::tuple<Slice<uint8_t>,Error> ps2_IRX_Relocate(ps2_IRX* m,uint32_t base);
uint32_t ps2_Machine_arg(ps2_Machine* m,uint32_t i);
uint32_t ps2_Machine_argT0(ps2_Machine* m);
void ps2_Machine_setRet(ps2_Machine* m,uint32_t v);
bool ps2_Machine_handleSyscall(ps2_Machine* m,r5900_CPU* c);
void ps2_Machine_setupThread(ps2_Machine* m);
void ps2_Machine_setSyscall(ps2_Machine* m);
void ps2_Machine_setupHeap(ps2_Machine* m);
void ps2_Machine_deci2Call(ps2_Machine* m);
std::string ps2_Machine_SyscallCensus(ps2_Machine* m);
std::string ps2_plural(int64_t n);
ps2_Machine* ps2_NewMachine();
ps2_IOP* ps2_Machine_StartIOP(ps2_Machine* m);
void ps2_Machine_iopPrint(ps2_Machine* m,std::string s);
void ps2_Machine_SetImageHash(ps2_Machine* m,std::string md5);
void ps2_Machine_SetVolume(ps2_Machine* m,iso9660_Volume* v);
iso9660_Volume* ps2_Machine_Volume(ps2_Machine* m);
void ps2_Machine_SetBIOS(ps2_Machine* m,Slice<uint8_t> raw);
Slice<uint8_t> ps2_Machine_BIOS(ps2_Machine* m);
ps2_Executable* ps2_Machine_Exe(ps2_Machine* m);
void ps2_Machine_LoadExecutable(ps2_Machine* m,ps2_Executable* e);
void ps2_Machine_mapMemory(ps2_Machine* m);
uint32_t ps2_phys(uint32_t addr);
bool ps2_Machine_mapped(ps2_Machine* m,uint32_t p);
std::tuple<Slice<uint8_t>,uint32_t,bool> ps2_Machine_slice(ps2_Machine* m,uint32_t p);
uint8_t ps2_Machine_Read(ps2_Machine* m,uint32_t addr);
void ps2_Machine_Write(ps2_Machine* m,uint32_t addr,uint8_t v);
uint32_t ps2_Machine_Read32(ps2_Machine* m,uint32_t addr);
void ps2_Machine_Write32(ps2_Machine* m,uint32_t addr,uint32_t v);
uint32_t ps2_Machine_Fetch32(ps2_Machine* m,uint32_t addr);
void ps2_Machine_noteRead(ps2_Machine* m,uint32_t p,uint32_t v);
void ps2_Machine_noteWrite(ps2_Machine* m,uint32_t p,uint32_t v);
uint32_t ps2_Machine_ioRead(ps2_Machine* m,uint32_t p);
void ps2_Machine_ioWrite(ps2_Machine* m,uint32_t p,uint32_t v);
Map<uint32_t,int64_t> ps2_Machine_Unmodelled(ps2_Machine* m);
std::string ps2_Machine_HardwareCensus(ps2_Machine* m);
std::string ps2_RegionName(uint32_t p);
void ps2_Machine_iopUnknownCDVD(ps2_Machine* m,uint32_t a,bool write);
void ps2_Machine_SetBreakpoint(ps2_Machine* m,uint32_t vaddr);
void ps2_Machine_ClearBreakpoints(ps2_Machine* m);
std::string ps2_Machine_TTY(ps2_Machine* m);
uint32_t ps2_Machine_VBlanks(ps2_Machine* m);
uint16_t ps2_Machine_padButtons(ps2_Machine* m);
void ps2_Machine_SetPadButtons(ps2_Machine* m,uint16_t b);
uint16_t ps2_Machine_PadButtons(ps2_Machine* m);
void ps2_Machine_SetPadStick(ps2_Machine* m,uint8_t lx,uint8_t ly,uint8_t rx,uint8_t ry);
std::tuple<uint8_t,uint8_t,uint8_t,uint8_t> ps2_Machine_PadStick(ps2_Machine* m);
std::tuple<uint8_t,uint8_t,uint8_t,uint8_t> ps2_Machine_padSticks(ps2_Machine* m);
std::tuple<uint16_t,bool> ps2_PadButton(std::string name);
Slice<std::string> ps2_PadButtonNames();
Slice<uint8_t> ps2_Machine_ReadMem(ps2_Machine* m,uint32_t addr,int64_t n);
std::string ps2_Machine_CString(ps2_Machine* m,uint32_t addr);
std::string ps2_Machine_DisasmAt(ps2_Machine* m,uint32_t vaddr);
std::string ps2_Machine_Sym(ps2_Machine* m,uint32_t addr);
void ps2_Machine_AddSymbols(ps2_Machine* m,Slice<ps2_Symbol> syms);
std::string ps2_Machine_Registers(ps2_Machine* m);
ps2_profCounters ps2_Machine_profCounters(ps2_Machine* m);
time_Time ps2_Machine_profStart(ps2_Machine* m);
void ps2_Machine_profEnd(ps2_Machine* m,int64_t bucket,time_Time t);
void ps2_Machine_profDrainEnter(ps2_Machine* m);
void ps2_Machine_profDrainExit(ps2_Machine* m);
std::tuple<time_Time,int64_t> ps2_Machine_profVU1Start(ps2_Machine* m);
void ps2_Machine_profVU1End(ps2_Machine* m,time_Time t,int64_t rasterBefore);
void ps2_Machine_profRunEnter(ps2_Machine* m);
void ps2_Machine_profRunExit(ps2_Machine* m);
void ps2_Machine_profFrame(ps2_Machine* m);
ps2_FrameProfile ps2_Machine_FrameProfile(ps2_Machine* m);
void ps2_Machine_SetProfile(ps2_Machine* m,bool on);
std::tuple<int64_t,bool> ps2_FindROMDIR(Slice<uint8_t> raw);
std::tuple<Slice<ps2_RomEntry>,Error> ps2_ReadROMDIR(Slice<uint8_t> raw);
std::tuple<Slice<uint8_t>,bool> ps2_ROMDIREntry(Slice<uint8_t> raw,std::string name);
std::tuple<Slice<std::string>,Error> ps2_IOPBootConf(Slice<uint8_t> bios);
std::tuple<Slice<ps2_RomEntry>,Error> ps2_ROMDIRModules(Slice<uint8_t> raw);
std::string ps2_Result_String(ps2_Result r);
ps2_Result ps2_Machine_Run(ps2_Machine* m,uint64_t maxSteps);
ps2_Result ps2_Machine_result(ps2_Machine* m,uint64_t steps,std::string reason);
std::string ps2_Machine_formatTrail(ps2_Machine* m,Slice<std::array<uint32_t,2>> trail,int64_t pos);
std::tuple<std::string,bool> ps2_Machine_unhandledException(ps2_Machine* m,uint32_t pc);
std::string ps2_excName(uint32_t code);
std::string ps2_threadState_String(ps2_threadState s);
void ps2_Machine_createThread(ps2_Machine* m);
void ps2_Machine_startThread(ps2_Machine* m);
void ps2_Machine_sleepThread(ps2_Machine* m);
void ps2_Machine_wakeupThread(ps2_Machine* m,uint32_t id);
void ps2_Machine_switchAway(ps2_Machine* m,ps2_threadState newState);
void ps2_Machine_preemptIfOutranked(ps2_Machine* m);
bool ps2_Machine_resume(ps2_Machine* m);
bool ps2_Machine_blocked(ps2_Machine* m);
void ps2_Machine_switchTo(ps2_Machine* m,ps2_thread* next,ps2_threadState curState);
ps2_thread* ps2_Machine_pickReady(ps2_Machine* m);
void ps2_Machine_onThreadExit(ps2_Machine* m);
std::string ps2_Machine_Threads(ps2_Machine* m);
void ps2_Machine_exitThread(ps2_Machine* m);
void ps2_Machine_referThreadStatus(ps2_Machine* m);
int64_t ps2_eeThreadStatus(ps2_threadState s);
void ps2_Machine_createSema(ps2_Machine* m);
void ps2_Machine_waitSema(ps2_Machine* m,uint32_t id);
void ps2_Machine_signalSema(ps2_Machine* m,uint32_t id);
void ps2_Machine_pollSema(ps2_Machine* m,uint32_t id);
void ps2_Machine_sifSetDma(ps2_Machine* m);
void ps2_Machine_sifData(ps2_Machine* m,uint32_t src,uint32_t dest,uint32_t size);
void ps2_Machine_iopReceive(ps2_Machine* m,uint32_t src,uint32_t dest,uint32_t size);
void ps2_Machine_iopReset(ps2_Machine* m,uint32_t src);
void ps2_Machine_sifToIOP(ps2_Machine* m,uint32_t src,uint32_t dest,uint32_t size);
void ps2_Machine_sifPump(ps2_Machine* m);
void ps2_Machine_sifFromIOP(ps2_Machine* m);
void ps2_Machine_sifWatch(ps2_Machine* m,uint32_t src,uint32_t cid,bool toIOP);
void ps2_Machine_sifSetReg(ps2_Machine* m);
void ps2_Machine_sifGetReg(ps2_Machine* m);
std::string ps2_sifCmdName(uint32_t cid);
std::string ps2_smflgBits(uint32_t v);
std::string ps2_Machine_SIFCensus(ps2_Machine* m);
void ps2_Machine_sbusFlagSet(ps2_Machine* m,uint32_t reg,uint32_t bits);
void ps2_Machine_sbusFlagClear(ps2_Machine* m,uint32_t reg,uint32_t bits);
uint32_t ps2_Machine_sbusRead(ps2_Machine* m,uint32_t off);
void ps2_Machine_sbusWriteIOP(ps2_Machine* m,uint32_t off,uint32_t v);
void ps2_Machine_sbusWrite(ps2_Machine* m,uint32_t off,uint32_t v);
uint32_t ps2_TexAddrPSMCT32(uint32_t bp,uint32_t bw,uint32_t x,uint32_t y);
uint32_t ps2_TexAddrPSMCT16(uint32_t bp,uint32_t bw,uint32_t x,uint32_t y,bool s);
uint32_t ps2_TexAddrPSMT8(uint32_t bp,uint32_t bw,uint32_t x,uint32_t y);
std::tuple<uint32_t,uint32_t> ps2_TexAddrPSMT4(uint32_t bp,uint32_t bw,uint32_t x,uint32_t y);
std::tuple<uint32_t,uint32_t> ps2_CSM1ClutXY(uint32_t i,uint32_t n);
float ps2_f32(Slice<uint8_t> b);
ps2_vif* ps2_Machine_ensureVIF(ps2_Machine* m,int64_t idx);
void ps2_Machine_vifStart(ps2_Machine* m,int64_t idx,ps2_dmacChan* c);
int64_t ps2_dmacChanForVIF(int64_t idx);
void ps2_vif_feed(ps2_vif* v,Slice<uint8_t> data);
void ps2_vif_logCode(ps2_vif* v,std::string s);
void ps2_vif_code(ps2_vif* v,uint32_t w);
void ps2_vif_arm(ps2_vif* v,uint32_t w,int64_t n);
void ps2_vif_finish(ps2_vif* v);
int64_t ps2_vif_unpackBytes(ps2_vif* v,uint32_t cmd,uint32_t num);
void ps2_vif_unpack(ps2_vif* v,uint32_t cmd,Slice<uint8_t> data);
void ps2_vif_runVU(ps2_vif* v,uint32_t start,bool cont);
void ps2_vif_xgkick(ps2_vif* v,uint32_t qw);
void ps2_vif_count(ps2_vif* v,std::string what);
uint64_t ps2_fnv64(Slice<uint8_t> b);
vu_VU* ps2_Machine_VU(ps2_Machine* m,int64_t idx);
Slice<uint8_t> ps2_Machine_VUMicro(ps2_Machine* m,int64_t idx);
Slice<uint8_t> ps2_Machine_VUDataMem(ps2_Machine* m,int64_t idx);
std::string ps2_Machine_VURegs(ps2_Machine* m,int64_t idx);
std::string ps2_Machine_VIFCensus(ps2_Machine* m);
Slice<ps2_countKV> ps2_sortedCounts(Map<std::string,int64_t> m);
uint32_t ps2_min32(uint32_t a,uint32_t b);
constexpr int64_t r5900_TLBSize=48ULL;
constexpr int64_t r5900_entryLoG=1ULL;
constexpr int64_t r5900_entryLoV=2ULL;
constexpr int64_t r5900_entryLoD=4ULL;
constexpr int64_t r5900_kuseg=0ULL;
constexpr int64_t r5900_kseg0=2147483648ULL;
constexpr int64_t r5900_kseg1=2684354560ULL;
constexpr int64_t r5900_ksseg=3221225472ULL;
constexpr int64_t r5900_kseg3=3758096384ULL;
constexpr int64_t r5900_spramBase=1879048192ULL;
constexpr int64_t r5900_spramSize=16384ULL;
constexpr int64_t r5900_cop0Index=0ULL;
constexpr int64_t r5900_cop0Random=1ULL;
constexpr int64_t r5900_cop0EntryLo0=2ULL;
constexpr int64_t r5900_cop0EntryLo1=3ULL;
constexpr int64_t r5900_cop0Context=4ULL;
constexpr int64_t r5900_cop0PageMask=5ULL;
constexpr int64_t r5900_cop0Wired=6ULL;
constexpr int64_t r5900_cop0BadVAddr=8ULL;
constexpr int64_t r5900_cop0Count=9ULL;
constexpr int64_t r5900_cop0EntryHi=10ULL;
constexpr int64_t r5900_cop0Compare=11ULL;
constexpr int64_t r5900_cop0Status=12ULL;
constexpr int64_t r5900_cop0Cause=13ULL;
constexpr int64_t r5900_cop0EPC=14ULL;
constexpr int64_t r5900_cop0PRId=15ULL;
constexpr int64_t r5900_cop0Config=16ULL;
constexpr int64_t r5900_cop0ErrorEPC=30ULL;
constexpr int64_t r5900_statusIE=1ULL;
constexpr int64_t r5900_statusEXL=2ULL;
constexpr int64_t r5900_statusERL=4ULL;
constexpr int64_t r5900_statusBEV=4194304ULL;
constexpr int64_t r5900_statusCU0=268435456ULL;
constexpr int64_t r5900_statusCU1=536870912ULL;
constexpr int64_t r5900_statusCU2=1073741824ULL;
constexpr int64_t r5900_statusEIE=65536ULL;
constexpr int64_t r5900_causeBD=2147483648ULL;
constexpr int64_t r5900_causeIP2=1024ULL;
constexpr int64_t r5900_causeIP3=2048ULL;
constexpr int64_t r5900_causeIP7=32768ULL;
constexpr int64_t r5900_excInt=0ULL;
constexpr int64_t r5900_excMod=1ULL;
constexpr int64_t r5900_excTLBL=2ULL;
constexpr int64_t r5900_excTLBS=3ULL;
constexpr int64_t r5900_excAdEL=4ULL;
constexpr int64_t r5900_excAdES=5ULL;
constexpr int64_t r5900_excSys=8ULL;
constexpr int64_t r5900_excBp=9ULL;
constexpr int64_t r5900_excRI=10ULL;
constexpr int64_t r5900_excCpU=11ULL;
constexpr int64_t r5900_excOv=12ULL;
constexpr int64_t r5900_excTrap=13ULL;
constexpr int64_t r5900_vecRAM=2147483648ULL;
constexpr int64_t r5900_vecROM=3217031680ULL;
constexpr int64_t r5900_Cop0Count=9ULL;
constexpr int64_t r5900_Cop0Compare=11ULL;
constexpr int64_t r5900_Cop0Status=12ULL;
constexpr int64_t r5900_Cop0Cause=13ULL;
constexpr int64_t r5900_Cop0EPC=14ULL;
constexpr int64_t r5900_Cop0BadVAddr=8ULL;
constexpr int64_t r5900_StatusIE=1ULL;
constexpr int64_t r5900_StatusEXL=2ULL;
constexpr int64_t r5900_StatusERL=4ULL;
constexpr int64_t r5900_StatusBEV=4194304ULL;
constexpr int64_t r5900_StatusEIE=65536ULL;
constexpr int64_t r5900_StatusCU0=268435456ULL;
constexpr int64_t r5900_StatusCU1=536870912ULL;
constexpr int64_t r5900_StatusCU2=1073741824ULL;
constexpr int64_t r5900_CauseIP2=1024ULL;
constexpr int64_t r5900_CauseIP3=2048ULL;
constexpr int64_t r5900_CauseIP7=32768ULL;
std::array<std::string,32> r5900_regName=std::array<std::string,32>{std::string("zero",4),std::string("at",2),std::string("v0",2),std::string("v1",2),std::string("a0",2),std::string("a1",2),std::string("a2",2),std::string("a3",2),std::string("t0",2),std::string("t1",2),std::string("t2",2),std::string("t3",2),std::string("t4",2),std::string("t5",2),std::string("t6",2),std::string("t7",2),std::string("s0",2),std::string("s1",2),std::string("s2",2),std::string("s3",2),std::string("s4",2),std::string("s5",2),std::string("s6",2),std::string("s7",2),std::string("t8",2),std::string("t9",2),std::string("k0",2),std::string("k1",2),std::string("gp",2),std::string("sp",2),std::string("fp",2),std::string("ra",2)};
std::array<std::string,32> r5900_cop0Name=[](){std::array<std::string,32> v{};v[0]=std::string("Index",5);v[1]=std::string("Random",6);v[2]=std::string("EntryLo0",8);v[3]=std::string("EntryLo1",8);v[4]=std::string("Context",7);v[5]=std::string("PageMask",8);v[6]=std::string("Wired",5);v[8]=std::string("BadVAddr",8);v[9]=std::string("Count",5);v[10]=std::string("EntryHi",7);v[11]=std::string("Compare",7);v[12]=std::string("Status",6);v[13]=std::string("Cause",5);v[14]=std::string("EPC",3);v[15]=std::string("PRId",4);v[16]=std::string("Config",6);v[23]=std::string("BadPAddr",8);v[24]=std::string("Debug",5);v[25]=std::string("Perf",4);v[30]=std::string("ErrorEPC",8);return v;}();
constexpr int64_t r5900_vuCtrlCMSAR=18ULL;
constexpr int64_t r5900_fcr31C=8388608ULL;
constexpr int64_t r5900_fMax=2147483647ULL;
constexpr r5900_mmiForm r5900_mDST=0ULL;
constexpr r5900_mmiForm r5900_mDT=1ULL;
constexpr r5900_mmiForm r5900_mDTS=2ULL;
constexpr r5900_mmiForm r5900_mDTSA=3ULL;
constexpr r5900_mmiForm r5900_mD=4ULL;
constexpr r5900_mmiForm r5900_mS=5ULL;
constexpr r5900_mmiForm r5900_mST=6ULL;
Map<uint32_t,r5900_mmiOp> r5900_mmi0Tab=Map<uint32_t,r5900_mmiOp>{{cast<uint32_t>(0ULL),r5900_mmiOp{std::string("paddw",5),cast<r5900_mmiForm>(0ULL)}},{cast<uint32_t>(1ULL),r5900_mmiOp{std::string("psubw",5),cast<r5900_mmiForm>(0ULL)}},{cast<uint32_t>(2ULL),r5900_mmiOp{std::string("pcgtw",5),cast<r5900_mmiForm>(0ULL)}},{cast<uint32_t>(3ULL),r5900_mmiOp{std::string("pmaxw",5),cast<r5900_mmiForm>(0ULL)}},{cast<uint32_t>(4ULL),r5900_mmiOp{std::string("paddh",5),cast<r5900_mmiForm>(0ULL)}},{cast<uint32_t>(5ULL),r5900_mmiOp{std::string("psubh",5),cast<r5900_mmiForm>(0ULL)}},{cast<uint32_t>(6ULL),r5900_mmiOp{std::string("pcgth",5),cast<r5900_mmiForm>(0ULL)}},{cast<uint32_t>(7ULL),r5900_mmiOp{std::string("pmaxh",5),cast<r5900_mmiForm>(0ULL)}},{cast<uint32_t>(8ULL),r5900_mmiOp{std::string("paddb",5),cast<r5900_mmiForm>(0ULL)}},{cast<uint32_t>(9ULL),r5900_mmiOp{std::string("psubb",5),cast<r5900_mmiForm>(0ULL)}},{cast<uint32_t>(10ULL),r5900_mmiOp{std::string("pcgtb",5),cast<r5900_mmiForm>(0ULL)}},{cast<uint32_t>(16ULL),r5900_mmiOp{std::string("paddsw",6),cast<r5900_mmiForm>(0ULL)}},{cast<uint32_t>(17ULL),r5900_mmiOp{std::string("psubsw",6),cast<r5900_mmiForm>(0ULL)}},{cast<uint32_t>(18ULL),r5900_mmiOp{std::string("pextlw",6),cast<r5900_mmiForm>(0ULL)}},{cast<uint32_t>(19ULL),r5900_mmiOp{std::string("ppacw",5),cast<r5900_mmiForm>(0ULL)}},{cast<uint32_t>(20ULL),r5900_mmiOp{std::string("paddsh",6),cast<r5900_mmiForm>(0ULL)}},{cast<uint32_t>(21ULL),r5900_mmiOp{std::string("psubsh",6),cast<r5900_mmiForm>(0ULL)}},{cast<uint32_t>(22ULL),r5900_mmiOp{std::string("pextlh",6),cast<r5900_mmiForm>(0ULL)}},{cast<uint32_t>(23ULL),r5900_mmiOp{std::string("ppach",5),cast<r5900_mmiForm>(0ULL)}},{cast<uint32_t>(24ULL),r5900_mmiOp{std::string("paddsb",6),cast<r5900_mmiForm>(0ULL)}},{cast<uint32_t>(25ULL),r5900_mmiOp{std::string("psubsb",6),cast<r5900_mmiForm>(0ULL)}},{cast<uint32_t>(26ULL),r5900_mmiOp{std::string("pextlb",6),cast<r5900_mmiForm>(0ULL)}},{cast<uint32_t>(27ULL),r5900_mmiOp{std::string("ppacb",5),cast<r5900_mmiForm>(0ULL)}},{cast<uint32_t>(30ULL),r5900_mmiOp{std::string("pext5",5),cast<r5900_mmiForm>(1ULL)}},{cast<uint32_t>(31ULL),r5900_mmiOp{std::string("ppac5",5),cast<r5900_mmiForm>(1ULL)}}};
Map<uint32_t,r5900_mmiOp> r5900_mmi1Tab=Map<uint32_t,r5900_mmiOp>{{cast<uint32_t>(1ULL),r5900_mmiOp{std::string("pabsw",5),cast<r5900_mmiForm>(1ULL)}},{cast<uint32_t>(2ULL),r5900_mmiOp{std::string("pceqw",5),cast<r5900_mmiForm>(0ULL)}},{cast<uint32_t>(3ULL),r5900_mmiOp{std::string("pminw",5),cast<r5900_mmiForm>(0ULL)}},{cast<uint32_t>(4ULL),r5900_mmiOp{std::string("padsbh",6),cast<r5900_mmiForm>(0ULL)}},{cast<uint32_t>(5ULL),r5900_mmiOp{std::string("pabsh",5),cast<r5900_mmiForm>(1ULL)}},{cast<uint32_t>(6ULL),r5900_mmiOp{std::string("pceqh",5),cast<r5900_mmiForm>(0ULL)}},{cast<uint32_t>(7ULL),r5900_mmiOp{std::string("pminh",5),cast<r5900_mmiForm>(0ULL)}},{cast<uint32_t>(10ULL),r5900_mmiOp{std::string("pceqb",5),cast<r5900_mmiForm>(0ULL)}},{cast<uint32_t>(16ULL),r5900_mmiOp{std::string("padduw",6),cast<r5900_mmiForm>(0ULL)}},{cast<uint32_t>(17ULL),r5900_mmiOp{std::string("psubuw",6),cast<r5900_mmiForm>(0ULL)}},{cast<uint32_t>(18ULL),r5900_mmiOp{std::string("pextuw",6),cast<r5900_mmiForm>(0ULL)}},{cast<uint32_t>(20ULL),r5900_mmiOp{std::string("padduh",6),cast<r5900_mmiForm>(0ULL)}},{cast<uint32_t>(21ULL),r5900_mmiOp{std::string("psubuh",6),cast<r5900_mmiForm>(0ULL)}},{cast<uint32_t>(22ULL),r5900_mmiOp{std::string("pextuh",6),cast<r5900_mmiForm>(0ULL)}},{cast<uint32_t>(24ULL),r5900_mmiOp{std::string("paddub",6),cast<r5900_mmiForm>(0ULL)}},{cast<uint32_t>(25ULL),r5900_mmiOp{std::string("psubub",6),cast<r5900_mmiForm>(0ULL)}},{cast<uint32_t>(26ULL),r5900_mmiOp{std::string("pextub",6),cast<r5900_mmiForm>(0ULL)}},{cast<uint32_t>(27ULL),r5900_mmiOp{std::string("qfsrv",5),cast<r5900_mmiForm>(0ULL)}}};
Map<uint32_t,r5900_mmiOp> r5900_mmi2Tab=Map<uint32_t,r5900_mmiOp>{{cast<uint32_t>(0ULL),r5900_mmiOp{std::string("pmaddw",6),cast<r5900_mmiForm>(0ULL)}},{cast<uint32_t>(2ULL),r5900_mmiOp{std::string("psllvw",6),cast<r5900_mmiForm>(2ULL)}},{cast<uint32_t>(3ULL),r5900_mmiOp{std::string("psrlvw",6),cast<r5900_mmiForm>(2ULL)}},{cast<uint32_t>(4ULL),r5900_mmiOp{std::string("pmsubw",6),cast<r5900_mmiForm>(0ULL)}},{cast<uint32_t>(8ULL),r5900_mmiOp{std::string("pmfhi",5),cast<r5900_mmiForm>(4ULL)}},{cast<uint32_t>(9ULL),r5900_mmiOp{std::string("pmflo",5),cast<r5900_mmiForm>(4ULL)}},{cast<uint32_t>(10ULL),r5900_mmiOp{std::string("pinth",5),cast<r5900_mmiForm>(0ULL)}},{cast<uint32_t>(12ULL),r5900_mmiOp{std::string("pmultw",6),cast<r5900_mmiForm>(0ULL)}},{cast<uint32_t>(13ULL),r5900_mmiOp{std::string("pdivw",5),cast<r5900_mmiForm>(6ULL)}},{cast<uint32_t>(14ULL),r5900_mmiOp{std::string("pcpyld",6),cast<r5900_mmiForm>(0ULL)}},{cast<uint32_t>(16ULL),r5900_mmiOp{std::string("pmaddh",6),cast<r5900_mmiForm>(0ULL)}},{cast<uint32_t>(17ULL),r5900_mmiOp{std::string("phmadh",6),cast<r5900_mmiForm>(0ULL)}},{cast<uint32_t>(18ULL),r5900_mmiOp{std::string("pand",4),cast<r5900_mmiForm>(0ULL)}},{cast<uint32_t>(19ULL),r5900_mmiOp{std::string("pxor",4),cast<r5900_mmiForm>(0ULL)}},{cast<uint32_t>(20ULL),r5900_mmiOp{std::string("pmsubh",6),cast<r5900_mmiForm>(0ULL)}},{cast<uint32_t>(21ULL),r5900_mmiOp{std::string("phmsbh",6),cast<r5900_mmiForm>(0ULL)}},{cast<uint32_t>(26ULL),r5900_mmiOp{std::string("pexeh",5),cast<r5900_mmiForm>(1ULL)}},{cast<uint32_t>(27ULL),r5900_mmiOp{std::string("prevh",5),cast<r5900_mmiForm>(1ULL)}},{cast<uint32_t>(28ULL),r5900_mmiOp{std::string("pmulth",6),cast<r5900_mmiForm>(0ULL)}},{cast<uint32_t>(29ULL),r5900_mmiOp{std::string("pdivbw",6),cast<r5900_mmiForm>(6ULL)}},{cast<uint32_t>(30ULL),r5900_mmiOp{std::string("pexew",5),cast<r5900_mmiForm>(1ULL)}},{cast<uint32_t>(31ULL),r5900_mmiOp{std::string("prot3w",6),cast<r5900_mmiForm>(1ULL)}}};
Map<uint32_t,r5900_mmiOp> r5900_mmi3Tab=Map<uint32_t,r5900_mmiOp>{{cast<uint32_t>(0ULL),r5900_mmiOp{std::string("pmadduw",7),cast<r5900_mmiForm>(0ULL)}},{cast<uint32_t>(3ULL),r5900_mmiOp{std::string("psravw",6),cast<r5900_mmiForm>(2ULL)}},{cast<uint32_t>(8ULL),r5900_mmiOp{std::string("pmthi",5),cast<r5900_mmiForm>(5ULL)}},{cast<uint32_t>(9ULL),r5900_mmiOp{std::string("pmtlo",5),cast<r5900_mmiForm>(5ULL)}},{cast<uint32_t>(10ULL),r5900_mmiOp{std::string("pinteh",6),cast<r5900_mmiForm>(0ULL)}},{cast<uint32_t>(12ULL),r5900_mmiOp{std::string("pmultuw",7),cast<r5900_mmiForm>(0ULL)}},{cast<uint32_t>(13ULL),r5900_mmiOp{std::string("pdivuw",6),cast<r5900_mmiForm>(6ULL)}},{cast<uint32_t>(14ULL),r5900_mmiOp{std::string("pcpyud",6),cast<r5900_mmiForm>(0ULL)}},{cast<uint32_t>(18ULL),r5900_mmiOp{std::string("por",3),cast<r5900_mmiForm>(0ULL)}},{cast<uint32_t>(19ULL),r5900_mmiOp{std::string("pnor",4),cast<r5900_mmiForm>(0ULL)}},{cast<uint32_t>(26ULL),r5900_mmiOp{std::string("pexch",5),cast<r5900_mmiForm>(1ULL)}},{cast<uint32_t>(27ULL),r5900_mmiOp{std::string("pcpyh",5),cast<r5900_mmiForm>(1ULL)}},{cast<uint32_t>(30ULL),r5900_mmiOp{std::string("pexcw",5),cast<r5900_mmiForm>(1ULL)}}};
std::array<std::string,8> r5900_pmfhlSub=std::array<std::string,8>{std::string("lw",2),std::string("uw",2),std::string("slw",3),std::string("lh",2),std::string("sh",2)};
constexpr r5900_Flow r5900_FlowSeq=0ULL;
constexpr r5900_Flow r5900_FlowBranch=1ULL;
constexpr r5900_Flow r5900_FlowJump=2ULL;
constexpr r5900_Flow r5900_FlowCall=3ULL;
constexpr r5900_Flow r5900_FlowReturn=4ULL;
constexpr r5900_Flow r5900_FlowIndJump=5ULL;
constexpr r5900_Flow r5900_FlowIndCall=6ULL;
constexpr r5900_Flow r5900_FlowStop=7ULL;
constexpr int64_t mips_cop0BadVaddr=8ULL;
constexpr int64_t mips_cop0Status=12ULL;
constexpr int64_t mips_cop0Cause=13ULL;
constexpr int64_t mips_cop0EPC=14ULL;
constexpr int64_t mips_cop0PRId=15ULL;
constexpr int64_t mips_excInt=0ULL;
constexpr int64_t mips_excAdEL=4ULL;
constexpr int64_t mips_excAdES=5ULL;
constexpr int64_t mips_excSys=8ULL;
constexpr int64_t mips_excBp=9ULL;
constexpr int64_t mips_excRI=10ULL;
constexpr int64_t mips_excCpU=11ULL;
constexpr int64_t mips_excOv=12ULL;
constexpr int64_t mips_vecRAM=2147483776ULL;
constexpr int64_t mips_vecROM=3217031552ULL;
std::array<std::string,32> mips_regName=std::array<std::string,32>{std::string("zero",4),std::string("at",2),std::string("v0",2),std::string("v1",2),std::string("a0",2),std::string("a1",2),std::string("a2",2),std::string("a3",2),std::string("t0",2),std::string("t1",2),std::string("t2",2),std::string("t3",2),std::string("t4",2),std::string("t5",2),std::string("t6",2),std::string("t7",2),std::string("s0",2),std::string("s1",2),std::string("s2",2),std::string("s3",2),std::string("s4",2),std::string("s5",2),std::string("s6",2),std::string("s7",2),std::string("t8",2),std::string("t9",2),std::string("k0",2),std::string("k1",2),std::string("gp",2),std::string("sp",2),std::string("fp",2),std::string("ra",2)};
Map<uint32_t,std::string> mips_gteCmd=Map<uint32_t,std::string>{{cast<uint32_t>(1ULL),std::string("rtps",4)},{cast<uint32_t>(6ULL),std::string("nclip",5)},{cast<uint32_t>(12ULL),std::string("op",2)},{cast<uint32_t>(16ULL),std::string("dpcs",4)},{cast<uint32_t>(17ULL),std::string("intpl",5)},{cast<uint32_t>(18ULL),std::string("mvmva",5)},{cast<uint32_t>(19ULL),std::string("ncds",4)},{cast<uint32_t>(20ULL),std::string("cdp",3)},{cast<uint32_t>(22ULL),std::string("ncdt",4)},{cast<uint32_t>(27ULL),std::string("nccs",4)},{cast<uint32_t>(28ULL),std::string("cc",2)},{cast<uint32_t>(30ULL),std::string("ncs",3)},{cast<uint32_t>(32ULL),std::string("nct",3)},{cast<uint32_t>(40ULL),std::string("sqr",3)},{cast<uint32_t>(41ULL),std::string("dcpl",4)},{cast<uint32_t>(42ULL),std::string("dpct",4)},{cast<uint32_t>(45ULL),std::string("avsz3",5)},{cast<uint32_t>(46ULL),std::string("avsz4",5)},{cast<uint32_t>(48ULL),std::string("rtpt",4)},{cast<uint32_t>(61ULL),std::string("gpf",3)},{cast<uint32_t>(62ULL),std::string("gpl",3)},{cast<uint32_t>(63ULL),std::string("ncct",4)}};
constexpr int64_t mips_flagIR0=4096ULL;
constexpr int64_t mips_flagSY2=8192ULL;
constexpr int64_t mips_flagSX2=16384ULL;
constexpr int64_t mips_flagMac0Neg=32768ULL;
constexpr int64_t mips_flagMac0Pos=65536ULL;
constexpr int64_t mips_flagDivOvf=131072ULL;
constexpr int64_t mips_flagSZ3OTZ=262144ULL;
constexpr int64_t mips_flagIR3=4194304ULL;
constexpr int64_t mips_flagIR2=8388608ULL;
constexpr int64_t mips_flagIR1=16777216ULL;
constexpr int64_t mips_flagErrMask=2139611136ULL;
std::array<uint8_t,257> mips_unrTable=std::array<uint8_t,257>{cast<uint8_t>(255ULL),cast<uint8_t>(253ULL),cast<uint8_t>(251ULL),cast<uint8_t>(249ULL),cast<uint8_t>(247ULL),cast<uint8_t>(245ULL),cast<uint8_t>(243ULL),cast<uint8_t>(241ULL),cast<uint8_t>(239ULL),cast<uint8_t>(238ULL),cast<uint8_t>(236ULL),cast<uint8_t>(234ULL),cast<uint8_t>(232ULL),cast<uint8_t>(230ULL),cast<uint8_t>(228ULL),cast<uint8_t>(227ULL),cast<uint8_t>(225ULL),cast<uint8_t>(223ULL),cast<uint8_t>(221ULL),cast<uint8_t>(220ULL),cast<uint8_t>(218ULL),cast<uint8_t>(216ULL),cast<uint8_t>(214ULL),cast<uint8_t>(213ULL),cast<uint8_t>(211ULL),cast<uint8_t>(209ULL),cast<uint8_t>(208ULL),cast<uint8_t>(206ULL),cast<uint8_t>(205ULL),cast<uint8_t>(203ULL),cast<uint8_t>(201ULL),cast<uint8_t>(200ULL),cast<uint8_t>(198ULL),cast<uint8_t>(197ULL),cast<uint8_t>(195ULL),cast<uint8_t>(193ULL),cast<uint8_t>(192ULL),cast<uint8_t>(190ULL),cast<uint8_t>(189ULL),cast<uint8_t>(187ULL),cast<uint8_t>(186ULL),cast<uint8_t>(184ULL),cast<uint8_t>(183ULL),cast<uint8_t>(181ULL),cast<uint8_t>(180ULL),cast<uint8_t>(178ULL),cast<uint8_t>(177ULL),cast<uint8_t>(176ULL),cast<uint8_t>(174ULL),cast<uint8_t>(173ULL),cast<uint8_t>(171ULL),cast<uint8_t>(170ULL),cast<uint8_t>(169ULL),cast<uint8_t>(167ULL),cast<uint8_t>(166ULL),cast<uint8_t>(164ULL),cast<uint8_t>(163ULL),cast<uint8_t>(162ULL),cast<uint8_t>(160ULL),cast<uint8_t>(159ULL),cast<uint8_t>(158ULL),cast<uint8_t>(156ULL),cast<uint8_t>(155ULL),cast<uint8_t>(154ULL),cast<uint8_t>(153ULL),cast<uint8_t>(151ULL),cast<uint8_t>(150ULL),cast<uint8_t>(149ULL),cast<uint8_t>(148ULL),cast<uint8_t>(146ULL),cast<uint8_t>(145ULL),cast<uint8_t>(144ULL),cast<uint8_t>(143ULL),cast<uint8_t>(141ULL),cast<uint8_t>(140ULL),cast<uint8_t>(139ULL),cast<uint8_t>(138ULL),cast<uint8_t>(137ULL),cast<uint8_t>(135ULL),cast<uint8_t>(134ULL),cast<uint8_t>(133ULL),cast<uint8_t>(132ULL),cast<uint8_t>(131ULL),cast<uint8_t>(130ULL),cast<uint8_t>(129ULL),cast<uint8_t>(127ULL),cast<uint8_t>(126ULL),cast<uint8_t>(125ULL),cast<uint8_t>(124ULL),cast<uint8_t>(123ULL),cast<uint8_t>(122ULL),cast<uint8_t>(121ULL),cast<uint8_t>(120ULL),cast<uint8_t>(119ULL),cast<uint8_t>(117ULL),cast<uint8_t>(116ULL),cast<uint8_t>(115ULL),cast<uint8_t>(114ULL),cast<uint8_t>(113ULL),cast<uint8_t>(112ULL),cast<uint8_t>(111ULL),cast<uint8_t>(110ULL),cast<uint8_t>(109ULL),cast<uint8_t>(108ULL),cast<uint8_t>(107ULL),cast<uint8_t>(106ULL),cast<uint8_t>(105ULL),cast<uint8_t>(104ULL),cast<uint8_t>(103ULL),cast<uint8_t>(102ULL),cast<uint8_t>(101ULL),cast<uint8_t>(100ULL),cast<uint8_t>(99ULL),cast<uint8_t>(98ULL),cast<uint8_t>(97ULL),cast<uint8_t>(96ULL),cast<uint8_t>(95ULL),cast<uint8_t>(94ULL),cast<uint8_t>(93ULL),cast<uint8_t>(93ULL),cast<uint8_t>(92ULL),cast<uint8_t>(91ULL),cast<uint8_t>(90ULL),cast<uint8_t>(89ULL),cast<uint8_t>(88ULL),cast<uint8_t>(87ULL),cast<uint8_t>(86ULL),cast<uint8_t>(85ULL),cast<uint8_t>(84ULL),cast<uint8_t>(83ULL),cast<uint8_t>(83ULL),cast<uint8_t>(82ULL),cast<uint8_t>(81ULL),cast<uint8_t>(80ULL),cast<uint8_t>(79ULL),cast<uint8_t>(78ULL),cast<uint8_t>(77ULL),cast<uint8_t>(77ULL),cast<uint8_t>(76ULL),cast<uint8_t>(75ULL),cast<uint8_t>(74ULL),cast<uint8_t>(73ULL),cast<uint8_t>(72ULL),cast<uint8_t>(72ULL),cast<uint8_t>(71ULL),cast<uint8_t>(70ULL),cast<uint8_t>(69ULL),cast<uint8_t>(68ULL),cast<uint8_t>(67ULL),cast<uint8_t>(67ULL),cast<uint8_t>(66ULL),cast<uint8_t>(65ULL),cast<uint8_t>(64ULL),cast<uint8_t>(63ULL),cast<uint8_t>(63ULL),cast<uint8_t>(62ULL),cast<uint8_t>(61ULL),cast<uint8_t>(60ULL),cast<uint8_t>(60ULL),cast<uint8_t>(59ULL),cast<uint8_t>(58ULL),cast<uint8_t>(57ULL),cast<uint8_t>(57ULL),cast<uint8_t>(56ULL),cast<uint8_t>(55ULL),cast<uint8_t>(54ULL),cast<uint8_t>(54ULL),cast<uint8_t>(53ULL),cast<uint8_t>(52ULL),cast<uint8_t>(51ULL),cast<uint8_t>(51ULL),cast<uint8_t>(50ULL),cast<uint8_t>(49ULL),cast<uint8_t>(49ULL),cast<uint8_t>(48ULL),cast<uint8_t>(47ULL),cast<uint8_t>(46ULL),cast<uint8_t>(46ULL),cast<uint8_t>(45ULL),cast<uint8_t>(44ULL),cast<uint8_t>(44ULL),cast<uint8_t>(43ULL),cast<uint8_t>(42ULL),cast<uint8_t>(42ULL),cast<uint8_t>(41ULL),cast<uint8_t>(40ULL),cast<uint8_t>(40ULL),cast<uint8_t>(39ULL),cast<uint8_t>(38ULL),cast<uint8_t>(38ULL),cast<uint8_t>(37ULL),cast<uint8_t>(36ULL),cast<uint8_t>(36ULL),cast<uint8_t>(35ULL),cast<uint8_t>(34ULL),cast<uint8_t>(34ULL),cast<uint8_t>(33ULL),cast<uint8_t>(32ULL),cast<uint8_t>(32ULL),cast<uint8_t>(31ULL),cast<uint8_t>(30ULL),cast<uint8_t>(30ULL),cast<uint8_t>(29ULL),cast<uint8_t>(29ULL),cast<uint8_t>(28ULL),cast<uint8_t>(27ULL),cast<uint8_t>(27ULL),cast<uint8_t>(26ULL),cast<uint8_t>(25ULL),cast<uint8_t>(25ULL),cast<uint8_t>(24ULL),cast<uint8_t>(24ULL),cast<uint8_t>(23ULL),cast<uint8_t>(22ULL),cast<uint8_t>(22ULL),cast<uint8_t>(21ULL),cast<uint8_t>(21ULL),cast<uint8_t>(20ULL),cast<uint8_t>(20ULL),cast<uint8_t>(19ULL),cast<uint8_t>(18ULL),cast<uint8_t>(18ULL),cast<uint8_t>(17ULL),cast<uint8_t>(17ULL),cast<uint8_t>(16ULL),cast<uint8_t>(15ULL),cast<uint8_t>(15ULL),cast<uint8_t>(14ULL),cast<uint8_t>(14ULL),cast<uint8_t>(13ULL),cast<uint8_t>(13ULL),cast<uint8_t>(12ULL),cast<uint8_t>(12ULL),cast<uint8_t>(11ULL),cast<uint8_t>(10ULL),cast<uint8_t>(10ULL),cast<uint8_t>(9ULL),cast<uint8_t>(9ULL),cast<uint8_t>(8ULL),cast<uint8_t>(8ULL),cast<uint8_t>(7ULL),cast<uint8_t>(7ULL),cast<uint8_t>(6ULL),cast<uint8_t>(6ULL),cast<uint8_t>(5ULL),cast<uint8_t>(5ULL),cast<uint8_t>(4ULL),cast<uint8_t>(4ULL),cast<uint8_t>(3ULL),cast<uint8_t>(3ULL),cast<uint8_t>(2ULL),cast<uint8_t>(2ULL),cast<uint8_t>(1ULL),cast<uint8_t>(1ULL),cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),cast<uint8_t>(0ULL)};
constexpr mips_Flow mips_FlowSeq=0ULL;
constexpr mips_Flow mips_FlowBranch=1ULL;
constexpr mips_Flow mips_FlowJump=2ULL;
constexpr mips_Flow mips_FlowCall=3ULL;
constexpr mips_Flow mips_FlowReturn=4ULL;
constexpr mips_Flow mips_FlowIndJump=5ULL;
constexpr mips_Flow mips_FlowIndCall=6ULL;
constexpr mips_Flow mips_FlowStop=7ULL;
std::array<std::string,16> vu_destNames=std::array<std::string,16>{std::string("",0),std::string("w",1),std::string("z",1),std::string("zw",2),std::string("y",1),std::string("yw",2),std::string("yz",2),std::string("yzw",3),std::string("x",1),std::string("xw",2),std::string("xz",2),std::string("xzw",3),std::string("xy",2),std::string("xyw",3),std::string("xyz",3),std::string("xyzw",4)};
std::array<std::string,4> vu_bcNames=std::array<std::string,4>{std::string("x",1),std::string("y",1),std::string("z",1),std::string("w",1)};
constexpr int64_t vu_upNone=0ULL;
constexpr int64_t vu_upVF=1ULL;
constexpr int64_t vu_upACC=2ULL;
constexpr int64_t vu_upBits=3ULL;
constexpr int64_t vu_upClip=4ULL;
constexpr int64_t vu_ctrlStatus=16ULL;
constexpr int64_t vu_ctrlMac=17ULL;
constexpr int64_t vu_ctrlClip=18ULL;
constexpr int64_t vu_ctrlR=20ULL;
constexpr int64_t vu_ctrlI=21ULL;
constexpr int64_t vu_ctrlQ=22ULL;
constexpr int64_t vu_ctrlTPC=26ULL;
constexpr int64_t vu_ctrlCMSAR0=27ULL;
constexpr int64_t vu_ctrlFBRST=28ULL;
constexpr int64_t vu_ctrlVPU=29ULL;
constexpr int64_t vu_ctrlCMSAR1=31ULL;
constexpr int64_t iso9660_BlockSize=2048ULL;
constexpr int64_t iso9660_pvdLBA=16ULL;
Slice<iso9660_Geometry> iso9660_geometries=Slice<iso9660_Geometry>{iso9660_Geometry{cast<int64_t>(2048ULL),cast<int64_t>(0ULL)},iso9660_Geometry{cast<int64_t>(2352ULL),cast<int64_t>(16ULL)},iso9660_Geometry{cast<int64_t>(2352ULL),cast<int64_t>(24ULL)},iso9660_Geometry{cast<int64_t>(2448ULL),cast<int64_t>(16ULL)},iso9660_Geometry{cast<int64_t>(2448ULL),cast<int64_t>(24ULL)},iso9660_Geometry{cast<int64_t>(2448ULL),cast<int64_t>(0ULL)},iso9660_Geometry{cast<int64_t>(2336ULL),cast<int64_t>(8ULL)},iso9660_Geometry{cast<int64_t>(2336ULL),cast<int64_t>(0ULL)}};
constexpr int64_t ps2_dmacBase=268468224ULL;
constexpr int64_t ps2_dmacEnd=268498432ULL;
constexpr int64_t ps2_dmacChannels=10ULL;
constexpr int64_t ps2_dCTRL=268492800ULL;
constexpr int64_t ps2_dSTAT=268492816ULL;
constexpr int64_t ps2_dPCR=268492832ULL;
constexpr int64_t ps2_dSQWC=268492848ULL;
constexpr int64_t ps2_dRBSR=268492864ULL;
constexpr int64_t ps2_dRBOR=268492880ULL;
constexpr int64_t ps2_dSTADR=268492896ULL;
constexpr int64_t ps2_dENABLER=268498208ULL;
constexpr int64_t ps2_dENABLEW=268498320ULL;
constexpr int64_t ps2_dChcr=0ULL;
constexpr int64_t ps2_dMadr=16ULL;
constexpr int64_t ps2_dQwc=32ULL;
constexpr int64_t ps2_dTadr=48ULL;
constexpr int64_t ps2_dAsr0=64ULL;
constexpr int64_t ps2_dAsr1=80ULL;
constexpr int64_t ps2_dSadr=128ULL;
std::array<uint32_t,10> ps2_dmacChanBase=std::array<uint32_t,10>{cast<uint32_t>(268468224ULL),cast<uint32_t>(268472320ULL),cast<uint32_t>(268476416ULL),cast<uint32_t>(268480512ULL),cast<uint32_t>(268481536ULL),cast<uint32_t>(268484608ULL),cast<uint32_t>(268485632ULL),cast<uint32_t>(268486656ULL),cast<uint32_t>(268488704ULL),cast<uint32_t>(268489728ULL)};
constexpr int64_t ps2_dChcrDir=1ULL;
constexpr int64_t ps2_dChcrModeM=12ULL;
constexpr int64_t ps2_dChcrTTE=64ULL;
constexpr int64_t ps2_dChcrTIE=128ULL;
constexpr int64_t ps2_dChcrStart=256ULL;
constexpr int64_t ps2_dmacChVIF0=0ULL;
constexpr int64_t ps2_dmacChVIF1=1ULL;
constexpr int64_t ps2_dmacChGIF=2ULL;
constexpr int64_t ps2_dmacChSIF0=5ULL;
constexpr int64_t ps2_dmacChSIF1=6ULL;
constexpr int64_t ps2_dmacChSPRfrom=8ULL;
constexpr int64_t ps2_dmacChSPRto=9ULL;
constexpr int64_t ps2_dtagREFE=0ULL;
constexpr int64_t ps2_dtagCNT=1ULL;
constexpr int64_t ps2_dtagNEXT=2ULL;
constexpr int64_t ps2_dtagREF=3ULL;
constexpr int64_t ps2_dtagREFS=4ULL;
constexpr int64_t ps2_dtagCALL=5ULL;
constexpr int64_t ps2_dtagRET=6ULL;
constexpr int64_t ps2_dtagEND=7ULL;
bool ps2_kickLog=(go_os_Getenv(std::string("PS2_KICKLOG",11)) != std::string("",0));
bool ps2_xferLog=(go_os_Getenv(std::string("PS2_XFERLOG",11)) != std::string("",0));
int64_t ps2_chainLogN=[]()->int64_t{
auto tmp19 = go_strconv_Atoi(go_os_Getenv(std::string("PS2_CHAINLOG",12)));
int64_t n = std::get<0>(tmp19);
return n;
}
();
constexpr int64_t ps2_eeTimerBase=268435456ULL;
constexpr int64_t ps2_eeTimerEnd=268441648ULL;
constexpr int64_t ps2_busclkPerField=2460180ULL;
constexpr int64_t ps2_gifPacked=0ULL;
constexpr int64_t ps2_gifReglist=1ULL;
constexpr int64_t ps2_gifImage=2ULL;
constexpr int64_t ps2_gifDisable=3ULL;
constexpr int64_t ps2_gifRegPrim=0ULL;
constexpr int64_t ps2_gifRegRGBAQ=1ULL;
constexpr int64_t ps2_gifRegST=2ULL;
constexpr int64_t ps2_gifRegUV=3ULL;
constexpr int64_t ps2_gifRegXYZF2=4ULL;
constexpr int64_t ps2_gifRegXYZ2=5ULL;
constexpr int64_t ps2_gifRegAD=14ULL;
constexpr int64_t ps2_gifRegNOP=15ULL;
constexpr int64_t ps2_gsPRIM=0ULL;
constexpr int64_t ps2_gsRGBAQ=1ULL;
constexpr int64_t ps2_gsXYZ2=5ULL;
constexpr int64_t ps2_gsTEX0_1=6ULL;
constexpr int64_t ps2_gsTEX0_2=7ULL;
constexpr int64_t ps2_gsCLAMP1=8ULL;
constexpr int64_t ps2_gsCLAMP2=9ULL;
constexpr int64_t ps2_gsFOG=10ULL;
constexpr int64_t ps2_gsTEX1_1=20ULL;
constexpr int64_t ps2_gsTEX1_2=21ULL;
constexpr int64_t ps2_gsTEX2_1=22ULL;
constexpr int64_t ps2_gsTEX2_2=23ULL;
constexpr int64_t ps2_gsXYOFFSET1=24ULL;
constexpr int64_t ps2_gsPRMODECONT=26ULL;
constexpr int64_t ps2_gsTEXCLUT=28ULL;
constexpr int64_t ps2_gsTEXA=59ULL;
constexpr int64_t ps2_gsFOGCOL=61ULL;
constexpr int64_t ps2_gsTEXFLUSH=63ULL;
constexpr int64_t ps2_gsSCISSOR1=64ULL;
constexpr int64_t ps2_gsALPHA1=66ULL;
constexpr int64_t ps2_gsALPHA2=67ULL;
constexpr int64_t ps2_gsCOLCLAMP=70ULL;
constexpr int64_t ps2_gsTEST1=71ULL;
constexpr int64_t ps2_gsTEST2=72ULL;
constexpr int64_t ps2_gsPABE=73ULL;
constexpr int64_t ps2_gsFBA1=74ULL;
constexpr int64_t ps2_gsFBA2=75ULL;
constexpr int64_t ps2_gsFRAME1=76ULL;
constexpr int64_t ps2_gsZBUF1=78ULL;
constexpr int64_t ps2_gsZBUF2=79ULL;
constexpr int64_t ps2_gsBITBLTBUF=80ULL;
constexpr int64_t ps2_gsTRXPOS=81ULL;
constexpr int64_t ps2_gsTRXREG=82ULL;
constexpr int64_t ps2_gsTRXDIR=83ULL;
constexpr int64_t ps2_gsHWREG=84ULL;
constexpr int64_t ps2_gsFINISH=97ULL;
constexpr int64_t ps2_gsVRAMSize=4194304ULL;
constexpr int64_t ps2_psmCT32=0ULL;
constexpr int64_t ps2_psmCT24=1ULL;
constexpr int64_t ps2_psmCT16=2ULL;
std::array<std::array<uint8_t,8>,4> ps2_blockPSMCT32=std::array<std::array<uint8_t,8>,4>{std::array<uint8_t,8>{cast<uint8_t>(0ULL),cast<uint8_t>(1ULL),cast<uint8_t>(4ULL),cast<uint8_t>(5ULL),cast<uint8_t>(16ULL),cast<uint8_t>(17ULL),cast<uint8_t>(20ULL),cast<uint8_t>(21ULL)},std::array<uint8_t,8>{cast<uint8_t>(2ULL),cast<uint8_t>(3ULL),cast<uint8_t>(6ULL),cast<uint8_t>(7ULL),cast<uint8_t>(18ULL),cast<uint8_t>(19ULL),cast<uint8_t>(22ULL),cast<uint8_t>(23ULL)},std::array<uint8_t,8>{cast<uint8_t>(8ULL),cast<uint8_t>(9ULL),cast<uint8_t>(12ULL),cast<uint8_t>(13ULL),cast<uint8_t>(24ULL),cast<uint8_t>(25ULL),cast<uint8_t>(28ULL),cast<uint8_t>(29ULL)},std::array<uint8_t,8>{cast<uint8_t>(10ULL),cast<uint8_t>(11ULL),cast<uint8_t>(14ULL),cast<uint8_t>(15ULL),cast<uint8_t>(26ULL),cast<uint8_t>(27ULL),cast<uint8_t>(30ULL),cast<uint8_t>(31ULL)}};
std::array<std::array<uint8_t,8>,8> ps2_columnPSMCT32=std::array<std::array<uint8_t,8>,8>{std::array<uint8_t,8>{cast<uint8_t>(0ULL),cast<uint8_t>(1ULL),cast<uint8_t>(4ULL),cast<uint8_t>(5ULL),cast<uint8_t>(8ULL),cast<uint8_t>(9ULL),cast<uint8_t>(12ULL),cast<uint8_t>(13ULL)},std::array<uint8_t,8>{cast<uint8_t>(2ULL),cast<uint8_t>(3ULL),cast<uint8_t>(6ULL),cast<uint8_t>(7ULL),cast<uint8_t>(10ULL),cast<uint8_t>(11ULL),cast<uint8_t>(14ULL),cast<uint8_t>(15ULL)},std::array<uint8_t,8>{cast<uint8_t>(16ULL),cast<uint8_t>(17ULL),cast<uint8_t>(20ULL),cast<uint8_t>(21ULL),cast<uint8_t>(24ULL),cast<uint8_t>(25ULL),cast<uint8_t>(28ULL),cast<uint8_t>(29ULL)},std::array<uint8_t,8>{cast<uint8_t>(18ULL),cast<uint8_t>(19ULL),cast<uint8_t>(22ULL),cast<uint8_t>(23ULL),cast<uint8_t>(26ULL),cast<uint8_t>(27ULL),cast<uint8_t>(30ULL),cast<uint8_t>(31ULL)},std::array<uint8_t,8>{cast<uint8_t>(32ULL),cast<uint8_t>(33ULL),cast<uint8_t>(36ULL),cast<uint8_t>(37ULL),cast<uint8_t>(40ULL),cast<uint8_t>(41ULL),cast<uint8_t>(44ULL),cast<uint8_t>(45ULL)},std::array<uint8_t,8>{cast<uint8_t>(34ULL),cast<uint8_t>(35ULL),cast<uint8_t>(38ULL),cast<uint8_t>(39ULL),cast<uint8_t>(42ULL),cast<uint8_t>(43ULL),cast<uint8_t>(46ULL),cast<uint8_t>(47ULL)},std::array<uint8_t,8>{cast<uint8_t>(48ULL),cast<uint8_t>(49ULL),cast<uint8_t>(52ULL),cast<uint8_t>(53ULL),cast<uint8_t>(56ULL),cast<uint8_t>(57ULL),cast<uint8_t>(60ULL),cast<uint8_t>(61ULL)},std::array<uint8_t,8>{cast<uint8_t>(50ULL),cast<uint8_t>(51ULL),cast<uint8_t>(54ULL),cast<uint8_t>(55ULL),cast<uint8_t>(58ULL),cast<uint8_t>(59ULL),cast<uint8_t>(62ULL),cast<uint8_t>(63ULL)}};
constexpr int64_t ps2_gsPMODE=301989888ULL;
constexpr int64_t ps2_gsDISPFB1=301990000ULL;
constexpr int64_t ps2_gsDISPLAY1=301990016ULL;
constexpr int64_t ps2_gsDISPFB2=301990032ULL;
constexpr int64_t ps2_gsDISPLAY2=301990048ULL;
constexpr int64_t ps2_gsBGCOLOR=301990112ULL;
constexpr int64_t ps2_gsCSR=301993984ULL;
constexpr int64_t ps2_gsIMR=301994000ULL;
constexpr int64_t ps2_gsST=2ULL;
constexpr int64_t ps2_gsUV=3ULL;
constexpr int64_t ps2_gsXYZF2=4ULL;
constexpr int64_t ps2_gsXYZF3=12ULL;
constexpr int64_t ps2_gsXYZ3=13ULL;
constexpr int64_t ps2_gsPRMODE=27ULL;
constexpr int64_t ps2_gsFRAME2=77ULL;
constexpr int64_t ps2_gsXYOFFSET2=25ULL;
constexpr int64_t ps2_gsSCISSOR2=65ULL;
constexpr int64_t ps2_primPoint=0ULL;
constexpr int64_t ps2_primLine=1ULL;
constexpr int64_t ps2_primLineStrip=2ULL;
constexpr int64_t ps2_primTri=3ULL;
constexpr int64_t ps2_primTriStrip=4ULL;
constexpr int64_t ps2_primTriFan=5ULL;
constexpr int64_t ps2_primSprite=6ULL;
std::array<std::string,8> ps2_primNames=std::array<std::string,8>{std::string("point",5),std::string("line",4),std::string("linestrip",9),std::string("tri",3),std::string("tristrip",8),std::string("trifan",6),std::string("sprite",6),std::string("prim7",5)};
constexpr int64_t ps2_psmCT16S=10ULL;
constexpr int64_t ps2_psmT8=19ULL;
constexpr int64_t ps2_psmT4=20ULL;
constexpr int64_t ps2_psmT8H=27ULL;
constexpr int64_t ps2_psmT4HL=36ULL;
constexpr int64_t ps2_psmT4HH=44ULL;
constexpr int64_t ps2_psmZ32=48ULL;
constexpr int64_t ps2_psmZ24=49ULL;
constexpr int64_t ps2_psmZ16=50ULL;
constexpr int64_t ps2_psmZ16S=58ULL;
std::array<std::array<uint8_t,4>,8> ps2_blockPSMT4=std::array<std::array<uint8_t,4>,8>{std::array<uint8_t,4>{cast<uint8_t>(0ULL),cast<uint8_t>(2ULL),cast<uint8_t>(8ULL),cast<uint8_t>(10ULL)},std::array<uint8_t,4>{cast<uint8_t>(1ULL),cast<uint8_t>(3ULL),cast<uint8_t>(9ULL),cast<uint8_t>(11ULL)},std::array<uint8_t,4>{cast<uint8_t>(4ULL),cast<uint8_t>(6ULL),cast<uint8_t>(12ULL),cast<uint8_t>(14ULL)},std::array<uint8_t,4>{cast<uint8_t>(5ULL),cast<uint8_t>(7ULL),cast<uint8_t>(13ULL),cast<uint8_t>(15ULL)},std::array<uint8_t,4>{cast<uint8_t>(16ULL),cast<uint8_t>(18ULL),cast<uint8_t>(24ULL),cast<uint8_t>(26ULL)},std::array<uint8_t,4>{cast<uint8_t>(17ULL),cast<uint8_t>(19ULL),cast<uint8_t>(25ULL),cast<uint8_t>(27ULL)},std::array<uint8_t,4>{cast<uint8_t>(20ULL),cast<uint8_t>(22ULL),cast<uint8_t>(28ULL),cast<uint8_t>(30ULL)},std::array<uint8_t,4>{cast<uint8_t>(21ULL),cast<uint8_t>(23ULL),cast<uint8_t>(29ULL),cast<uint8_t>(31ULL)}};
std::array<std::array<uint8_t,4>,8> ps2_blockPSMCT16=std::array<std::array<uint8_t,4>,8>{std::array<uint8_t,4>{cast<uint8_t>(0ULL),cast<uint8_t>(2ULL),cast<uint8_t>(8ULL),cast<uint8_t>(10ULL)},std::array<uint8_t,4>{cast<uint8_t>(1ULL),cast<uint8_t>(3ULL),cast<uint8_t>(9ULL),cast<uint8_t>(11ULL)},std::array<uint8_t,4>{cast<uint8_t>(4ULL),cast<uint8_t>(6ULL),cast<uint8_t>(12ULL),cast<uint8_t>(14ULL)},std::array<uint8_t,4>{cast<uint8_t>(5ULL),cast<uint8_t>(7ULL),cast<uint8_t>(13ULL),cast<uint8_t>(15ULL)},std::array<uint8_t,4>{cast<uint8_t>(16ULL),cast<uint8_t>(18ULL),cast<uint8_t>(24ULL),cast<uint8_t>(26ULL)},std::array<uint8_t,4>{cast<uint8_t>(17ULL),cast<uint8_t>(19ULL),cast<uint8_t>(25ULL),cast<uint8_t>(27ULL)},std::array<uint8_t,4>{cast<uint8_t>(20ULL),cast<uint8_t>(22ULL),cast<uint8_t>(28ULL),cast<uint8_t>(30ULL)},std::array<uint8_t,4>{cast<uint8_t>(21ULL),cast<uint8_t>(23ULL),cast<uint8_t>(29ULL),cast<uint8_t>(31ULL)}};
std::array<std::array<uint8_t,4>,8> ps2_blockPSMCT16S=std::array<std::array<uint8_t,4>,8>{std::array<uint8_t,4>{cast<uint8_t>(0ULL),cast<uint8_t>(2ULL),cast<uint8_t>(16ULL),cast<uint8_t>(18ULL)},std::array<uint8_t,4>{cast<uint8_t>(1ULL),cast<uint8_t>(3ULL),cast<uint8_t>(17ULL),cast<uint8_t>(19ULL)},std::array<uint8_t,4>{cast<uint8_t>(8ULL),cast<uint8_t>(10ULL),cast<uint8_t>(24ULL),cast<uint8_t>(26ULL)},std::array<uint8_t,4>{cast<uint8_t>(9ULL),cast<uint8_t>(11ULL),cast<uint8_t>(25ULL),cast<uint8_t>(27ULL)},std::array<uint8_t,4>{cast<uint8_t>(4ULL),cast<uint8_t>(6ULL),cast<uint8_t>(20ULL),cast<uint8_t>(22ULL)},std::array<uint8_t,4>{cast<uint8_t>(5ULL),cast<uint8_t>(7ULL),cast<uint8_t>(21ULL),cast<uint8_t>(23ULL)},std::array<uint8_t,4>{cast<uint8_t>(12ULL),cast<uint8_t>(14ULL),cast<uint8_t>(28ULL),cast<uint8_t>(30ULL)},std::array<uint8_t,4>{cast<uint8_t>(13ULL),cast<uint8_t>(15ULL),cast<uint8_t>(29ULL),cast<uint8_t>(31ULL)}};
std::array<std::array<uint8_t,64>,2> ps2_columnWordT8=std::array<std::array<uint8_t,64>,2>{std::array<uint8_t,64>{cast<uint8_t>(0ULL),cast<uint8_t>(1ULL),cast<uint8_t>(4ULL),cast<uint8_t>(5ULL),cast<uint8_t>(8ULL),cast<uint8_t>(9ULL),cast<uint8_t>(12ULL),cast<uint8_t>(13ULL),cast<uint8_t>(0ULL),cast<uint8_t>(1ULL),cast<uint8_t>(4ULL),cast<uint8_t>(5ULL),cast<uint8_t>(8ULL),cast<uint8_t>(9ULL),cast<uint8_t>(12ULL),cast<uint8_t>(13ULL),cast<uint8_t>(2ULL),cast<uint8_t>(3ULL),cast<uint8_t>(6ULL),cast<uint8_t>(7ULL),cast<uint8_t>(10ULL),cast<uint8_t>(11ULL),cast<uint8_t>(14ULL),cast<uint8_t>(15ULL),cast<uint8_t>(2ULL),cast<uint8_t>(3ULL),cast<uint8_t>(6ULL),cast<uint8_t>(7ULL),cast<uint8_t>(10ULL),cast<uint8_t>(11ULL),cast<uint8_t>(14ULL),cast<uint8_t>(15ULL),cast<uint8_t>(8ULL),cast<uint8_t>(9ULL),cast<uint8_t>(12ULL),cast<uint8_t>(13ULL),cast<uint8_t>(0ULL),cast<uint8_t>(1ULL),cast<uint8_t>(4ULL),cast<uint8_t>(5ULL),cast<uint8_t>(8ULL),cast<uint8_t>(9ULL),cast<uint8_t>(12ULL),cast<uint8_t>(13ULL),cast<uint8_t>(0ULL),cast<uint8_t>(1ULL),cast<uint8_t>(4ULL),cast<uint8_t>(5ULL),cast<uint8_t>(10ULL),cast<uint8_t>(11ULL),cast<uint8_t>(14ULL),cast<uint8_t>(15ULL),cast<uint8_t>(2ULL),cast<uint8_t>(3ULL),cast<uint8_t>(6ULL),cast<uint8_t>(7ULL),cast<uint8_t>(10ULL),cast<uint8_t>(11ULL),cast<uint8_t>(14ULL),cast<uint8_t>(15ULL),cast<uint8_t>(2ULL),cast<uint8_t>(3ULL),cast<uint8_t>(6ULL),cast<uint8_t>(7ULL)},std::array<uint8_t,64>{cast<uint8_t>(8ULL),cast<uint8_t>(9ULL),cast<uint8_t>(12ULL),cast<uint8_t>(13ULL),cast<uint8_t>(0ULL),cast<uint8_t>(1ULL),cast<uint8_t>(4ULL),cast<uint8_t>(5ULL),cast<uint8_t>(8ULL),cast<uint8_t>(9ULL),cast<uint8_t>(12ULL),cast<uint8_t>(13ULL),cast<uint8_t>(0ULL),cast<uint8_t>(1ULL),cast<uint8_t>(4ULL),cast<uint8_t>(5ULL),cast<uint8_t>(10ULL),cast<uint8_t>(11ULL),cast<uint8_t>(14ULL),cast<uint8_t>(15ULL),cast<uint8_t>(2ULL),cast<uint8_t>(3ULL),cast<uint8_t>(6ULL),cast<uint8_t>(7ULL),cast<uint8_t>(10ULL),cast<uint8_t>(11ULL),cast<uint8_t>(14ULL),cast<uint8_t>(15ULL),cast<uint8_t>(2ULL),cast<uint8_t>(3ULL),cast<uint8_t>(6ULL),cast<uint8_t>(7ULL),cast<uint8_t>(0ULL),cast<uint8_t>(1ULL),cast<uint8_t>(4ULL),cast<uint8_t>(5ULL),cast<uint8_t>(8ULL),cast<uint8_t>(9ULL),cast<uint8_t>(12ULL),cast<uint8_t>(13ULL),cast<uint8_t>(0ULL),cast<uint8_t>(1ULL),cast<uint8_t>(4ULL),cast<uint8_t>(5ULL),cast<uint8_t>(8ULL),cast<uint8_t>(9ULL),cast<uint8_t>(12ULL),cast<uint8_t>(13ULL),cast<uint8_t>(2ULL),cast<uint8_t>(3ULL),cast<uint8_t>(6ULL),cast<uint8_t>(7ULL),cast<uint8_t>(10ULL),cast<uint8_t>(11ULL),cast<uint8_t>(14ULL),cast<uint8_t>(15ULL),cast<uint8_t>(2ULL),cast<uint8_t>(3ULL),cast<uint8_t>(6ULL),cast<uint8_t>(7ULL),cast<uint8_t>(10ULL),cast<uint8_t>(11ULL),cast<uint8_t>(14ULL),cast<uint8_t>(15ULL)}};
std::array<uint8_t,64> ps2_columnByteT8=std::array<uint8_t,64>{cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),cast<uint8_t>(2ULL),cast<uint8_t>(2ULL),cast<uint8_t>(2ULL),cast<uint8_t>(2ULL),cast<uint8_t>(2ULL),cast<uint8_t>(2ULL),cast<uint8_t>(2ULL),cast<uint8_t>(2ULL),cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),cast<uint8_t>(2ULL),cast<uint8_t>(2ULL),cast<uint8_t>(2ULL),cast<uint8_t>(2ULL),cast<uint8_t>(2ULL),cast<uint8_t>(2ULL),cast<uint8_t>(2ULL),cast<uint8_t>(2ULL),cast<uint8_t>(1ULL),cast<uint8_t>(1ULL),cast<uint8_t>(1ULL),cast<uint8_t>(1ULL),cast<uint8_t>(1ULL),cast<uint8_t>(1ULL),cast<uint8_t>(1ULL),cast<uint8_t>(1ULL),cast<uint8_t>(3ULL),cast<uint8_t>(3ULL),cast<uint8_t>(3ULL),cast<uint8_t>(3ULL),cast<uint8_t>(3ULL),cast<uint8_t>(3ULL),cast<uint8_t>(3ULL),cast<uint8_t>(3ULL),cast<uint8_t>(1ULL),cast<uint8_t>(1ULL),cast<uint8_t>(1ULL),cast<uint8_t>(1ULL),cast<uint8_t>(1ULL),cast<uint8_t>(1ULL),cast<uint8_t>(1ULL),cast<uint8_t>(1ULL),cast<uint8_t>(3ULL),cast<uint8_t>(3ULL),cast<uint8_t>(3ULL),cast<uint8_t>(3ULL),cast<uint8_t>(3ULL),cast<uint8_t>(3ULL),cast<uint8_t>(3ULL),cast<uint8_t>(3ULL)};
std::array<std::array<uint8_t,128>,2> ps2_columnWordT4=std::array<std::array<uint8_t,128>,2>{std::array<uint8_t,128>{cast<uint8_t>(0ULL),cast<uint8_t>(1ULL),cast<uint8_t>(4ULL),cast<uint8_t>(5ULL),cast<uint8_t>(8ULL),cast<uint8_t>(9ULL),cast<uint8_t>(12ULL),cast<uint8_t>(13ULL),cast<uint8_t>(0ULL),cast<uint8_t>(1ULL),cast<uint8_t>(4ULL),cast<uint8_t>(5ULL),cast<uint8_t>(8ULL),cast<uint8_t>(9ULL),cast<uint8_t>(12ULL),cast<uint8_t>(13ULL),cast<uint8_t>(0ULL),cast<uint8_t>(1ULL),cast<uint8_t>(4ULL),cast<uint8_t>(5ULL),cast<uint8_t>(8ULL),cast<uint8_t>(9ULL),cast<uint8_t>(12ULL),cast<uint8_t>(13ULL),cast<uint8_t>(0ULL),cast<uint8_t>(1ULL),cast<uint8_t>(4ULL),cast<uint8_t>(5ULL),cast<uint8_t>(8ULL),cast<uint8_t>(9ULL),cast<uint8_t>(12ULL),cast<uint8_t>(13ULL),cast<uint8_t>(2ULL),cast<uint8_t>(3ULL),cast<uint8_t>(6ULL),cast<uint8_t>(7ULL),cast<uint8_t>(10ULL),cast<uint8_t>(11ULL),cast<uint8_t>(14ULL),cast<uint8_t>(15ULL),cast<uint8_t>(2ULL),cast<uint8_t>(3ULL),cast<uint8_t>(6ULL),cast<uint8_t>(7ULL),cast<uint8_t>(10ULL),cast<uint8_t>(11ULL),cast<uint8_t>(14ULL),cast<uint8_t>(15ULL),cast<uint8_t>(2ULL),cast<uint8_t>(3ULL),cast<uint8_t>(6ULL),cast<uint8_t>(7ULL),cast<uint8_t>(10ULL),cast<uint8_t>(11ULL),cast<uint8_t>(14ULL),cast<uint8_t>(15ULL),cast<uint8_t>(2ULL),cast<uint8_t>(3ULL),cast<uint8_t>(6ULL),cast<uint8_t>(7ULL),cast<uint8_t>(10ULL),cast<uint8_t>(11ULL),cast<uint8_t>(14ULL),cast<uint8_t>(15ULL),cast<uint8_t>(8ULL),cast<uint8_t>(9ULL),cast<uint8_t>(12ULL),cast<uint8_t>(13ULL),cast<uint8_t>(0ULL),cast<uint8_t>(1ULL),cast<uint8_t>(4ULL),cast<uint8_t>(5ULL),cast<uint8_t>(8ULL),cast<uint8_t>(9ULL),cast<uint8_t>(12ULL),cast<uint8_t>(13ULL),cast<uint8_t>(0ULL),cast<uint8_t>(1ULL),cast<uint8_t>(4ULL),cast<uint8_t>(5ULL),cast<uint8_t>(8ULL),cast<uint8_t>(9ULL),cast<uint8_t>(12ULL),cast<uint8_t>(13ULL),cast<uint8_t>(0ULL),cast<uint8_t>(1ULL),cast<uint8_t>(4ULL),cast<uint8_t>(5ULL),cast<uint8_t>(8ULL),cast<uint8_t>(9ULL),cast<uint8_t>(12ULL),cast<uint8_t>(13ULL),cast<uint8_t>(0ULL),cast<uint8_t>(1ULL),cast<uint8_t>(4ULL),cast<uint8_t>(5ULL),cast<uint8_t>(10ULL),cast<uint8_t>(11ULL),cast<uint8_t>(14ULL),cast<uint8_t>(15ULL),cast<uint8_t>(2ULL),cast<uint8_t>(3ULL),cast<uint8_t>(6ULL),cast<uint8_t>(7ULL),cast<uint8_t>(10ULL),cast<uint8_t>(11ULL),cast<uint8_t>(14ULL),cast<uint8_t>(15ULL),cast<uint8_t>(2ULL),cast<uint8_t>(3ULL),cast<uint8_t>(6ULL),cast<uint8_t>(7ULL),cast<uint8_t>(10ULL),cast<uint8_t>(11ULL),cast<uint8_t>(14ULL),cast<uint8_t>(15ULL),cast<uint8_t>(2ULL),cast<uint8_t>(3ULL),cast<uint8_t>(6ULL),cast<uint8_t>(7ULL),cast<uint8_t>(10ULL),cast<uint8_t>(11ULL),cast<uint8_t>(14ULL),cast<uint8_t>(15ULL),cast<uint8_t>(2ULL),cast<uint8_t>(3ULL),cast<uint8_t>(6ULL),cast<uint8_t>(7ULL)},std::array<uint8_t,128>{cast<uint8_t>(8ULL),cast<uint8_t>(9ULL),cast<uint8_t>(12ULL),cast<uint8_t>(13ULL),cast<uint8_t>(0ULL),cast<uint8_t>(1ULL),cast<uint8_t>(4ULL),cast<uint8_t>(5ULL),cast<uint8_t>(8ULL),cast<uint8_t>(9ULL),cast<uint8_t>(12ULL),cast<uint8_t>(13ULL),cast<uint8_t>(0ULL),cast<uint8_t>(1ULL),cast<uint8_t>(4ULL),cast<uint8_t>(5ULL),cast<uint8_t>(8ULL),cast<uint8_t>(9ULL),cast<uint8_t>(12ULL),cast<uint8_t>(13ULL),cast<uint8_t>(0ULL),cast<uint8_t>(1ULL),cast<uint8_t>(4ULL),cast<uint8_t>(5ULL),cast<uint8_t>(8ULL),cast<uint8_t>(9ULL),cast<uint8_t>(12ULL),cast<uint8_t>(13ULL),cast<uint8_t>(0ULL),cast<uint8_t>(1ULL),cast<uint8_t>(4ULL),cast<uint8_t>(5ULL),cast<uint8_t>(10ULL),cast<uint8_t>(11ULL),cast<uint8_t>(14ULL),cast<uint8_t>(15ULL),cast<uint8_t>(2ULL),cast<uint8_t>(3ULL),cast<uint8_t>(6ULL),cast<uint8_t>(7ULL),cast<uint8_t>(10ULL),cast<uint8_t>(11ULL),cast<uint8_t>(14ULL),cast<uint8_t>(15ULL),cast<uint8_t>(2ULL),cast<uint8_t>(3ULL),cast<uint8_t>(6ULL),cast<uint8_t>(7ULL),cast<uint8_t>(10ULL),cast<uint8_t>(11ULL),cast<uint8_t>(14ULL),cast<uint8_t>(15ULL),cast<uint8_t>(2ULL),cast<uint8_t>(3ULL),cast<uint8_t>(6ULL),cast<uint8_t>(7ULL),cast<uint8_t>(10ULL),cast<uint8_t>(11ULL),cast<uint8_t>(14ULL),cast<uint8_t>(15ULL),cast<uint8_t>(2ULL),cast<uint8_t>(3ULL),cast<uint8_t>(6ULL),cast<uint8_t>(7ULL),cast<uint8_t>(0ULL),cast<uint8_t>(1ULL),cast<uint8_t>(4ULL),cast<uint8_t>(5ULL),cast<uint8_t>(8ULL),cast<uint8_t>(9ULL),cast<uint8_t>(12ULL),cast<uint8_t>(13ULL),cast<uint8_t>(0ULL),cast<uint8_t>(1ULL),cast<uint8_t>(4ULL),cast<uint8_t>(5ULL),cast<uint8_t>(8ULL),cast<uint8_t>(9ULL),cast<uint8_t>(12ULL),cast<uint8_t>(13ULL),cast<uint8_t>(0ULL),cast<uint8_t>(1ULL),cast<uint8_t>(4ULL),cast<uint8_t>(5ULL),cast<uint8_t>(8ULL),cast<uint8_t>(9ULL),cast<uint8_t>(12ULL),cast<uint8_t>(13ULL),cast<uint8_t>(0ULL),cast<uint8_t>(1ULL),cast<uint8_t>(4ULL),cast<uint8_t>(5ULL),cast<uint8_t>(8ULL),cast<uint8_t>(9ULL),cast<uint8_t>(12ULL),cast<uint8_t>(13ULL),cast<uint8_t>(2ULL),cast<uint8_t>(3ULL),cast<uint8_t>(6ULL),cast<uint8_t>(7ULL),cast<uint8_t>(10ULL),cast<uint8_t>(11ULL),cast<uint8_t>(14ULL),cast<uint8_t>(15ULL),cast<uint8_t>(2ULL),cast<uint8_t>(3ULL),cast<uint8_t>(6ULL),cast<uint8_t>(7ULL),cast<uint8_t>(10ULL),cast<uint8_t>(11ULL),cast<uint8_t>(14ULL),cast<uint8_t>(15ULL),cast<uint8_t>(2ULL),cast<uint8_t>(3ULL),cast<uint8_t>(6ULL),cast<uint8_t>(7ULL),cast<uint8_t>(10ULL),cast<uint8_t>(11ULL),cast<uint8_t>(14ULL),cast<uint8_t>(15ULL),cast<uint8_t>(2ULL),cast<uint8_t>(3ULL),cast<uint8_t>(6ULL),cast<uint8_t>(7ULL),cast<uint8_t>(10ULL),cast<uint8_t>(11ULL),cast<uint8_t>(14ULL),cast<uint8_t>(15ULL)}};
std::array<uint8_t,128> ps2_columnNibbleT4=std::array<uint8_t,128>{cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),cast<uint8_t>(2ULL),cast<uint8_t>(2ULL),cast<uint8_t>(2ULL),cast<uint8_t>(2ULL),cast<uint8_t>(2ULL),cast<uint8_t>(2ULL),cast<uint8_t>(2ULL),cast<uint8_t>(2ULL),cast<uint8_t>(4ULL),cast<uint8_t>(4ULL),cast<uint8_t>(4ULL),cast<uint8_t>(4ULL),cast<uint8_t>(4ULL),cast<uint8_t>(4ULL),cast<uint8_t>(4ULL),cast<uint8_t>(4ULL),cast<uint8_t>(6ULL),cast<uint8_t>(6ULL),cast<uint8_t>(6ULL),cast<uint8_t>(6ULL),cast<uint8_t>(6ULL),cast<uint8_t>(6ULL),cast<uint8_t>(6ULL),cast<uint8_t>(6ULL),cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),cast<uint8_t>(2ULL),cast<uint8_t>(2ULL),cast<uint8_t>(2ULL),cast<uint8_t>(2ULL),cast<uint8_t>(2ULL),cast<uint8_t>(2ULL),cast<uint8_t>(2ULL),cast<uint8_t>(2ULL),cast<uint8_t>(4ULL),cast<uint8_t>(4ULL),cast<uint8_t>(4ULL),cast<uint8_t>(4ULL),cast<uint8_t>(4ULL),cast<uint8_t>(4ULL),cast<uint8_t>(4ULL),cast<uint8_t>(4ULL),cast<uint8_t>(6ULL),cast<uint8_t>(6ULL),cast<uint8_t>(6ULL),cast<uint8_t>(6ULL),cast<uint8_t>(6ULL),cast<uint8_t>(6ULL),cast<uint8_t>(6ULL),cast<uint8_t>(6ULL),cast<uint8_t>(1ULL),cast<uint8_t>(1ULL),cast<uint8_t>(1ULL),cast<uint8_t>(1ULL),cast<uint8_t>(1ULL),cast<uint8_t>(1ULL),cast<uint8_t>(1ULL),cast<uint8_t>(1ULL),cast<uint8_t>(3ULL),cast<uint8_t>(3ULL),cast<uint8_t>(3ULL),cast<uint8_t>(3ULL),cast<uint8_t>(3ULL),cast<uint8_t>(3ULL),cast<uint8_t>(3ULL),cast<uint8_t>(3ULL),cast<uint8_t>(5ULL),cast<uint8_t>(5ULL),cast<uint8_t>(5ULL),cast<uint8_t>(5ULL),cast<uint8_t>(5ULL),cast<uint8_t>(5ULL),cast<uint8_t>(5ULL),cast<uint8_t>(5ULL),cast<uint8_t>(7ULL),cast<uint8_t>(7ULL),cast<uint8_t>(7ULL),cast<uint8_t>(7ULL),cast<uint8_t>(7ULL),cast<uint8_t>(7ULL),cast<uint8_t>(7ULL),cast<uint8_t>(7ULL),cast<uint8_t>(1ULL),cast<uint8_t>(1ULL),cast<uint8_t>(1ULL),cast<uint8_t>(1ULL),cast<uint8_t>(1ULL),cast<uint8_t>(1ULL),cast<uint8_t>(1ULL),cast<uint8_t>(1ULL),cast<uint8_t>(3ULL),cast<uint8_t>(3ULL),cast<uint8_t>(3ULL),cast<uint8_t>(3ULL),cast<uint8_t>(3ULL),cast<uint8_t>(3ULL),cast<uint8_t>(3ULL),cast<uint8_t>(3ULL),cast<uint8_t>(5ULL),cast<uint8_t>(5ULL),cast<uint8_t>(5ULL),cast<uint8_t>(5ULL),cast<uint8_t>(5ULL),cast<uint8_t>(5ULL),cast<uint8_t>(5ULL),cast<uint8_t>(5ULL),cast<uint8_t>(7ULL),cast<uint8_t>(7ULL),cast<uint8_t>(7ULL),cast<uint8_t>(7ULL),cast<uint8_t>(7ULL),cast<uint8_t>(7ULL),cast<uint8_t>(7ULL),cast<uint8_t>(7ULL)};
std::array<uint8_t,32> ps2_columnWordCT16=std::array<uint8_t,32>{cast<uint8_t>(0ULL),cast<uint8_t>(1ULL),cast<uint8_t>(4ULL),cast<uint8_t>(5ULL),cast<uint8_t>(8ULL),cast<uint8_t>(9ULL),cast<uint8_t>(12ULL),cast<uint8_t>(13ULL),cast<uint8_t>(0ULL),cast<uint8_t>(1ULL),cast<uint8_t>(4ULL),cast<uint8_t>(5ULL),cast<uint8_t>(8ULL),cast<uint8_t>(9ULL),cast<uint8_t>(12ULL),cast<uint8_t>(13ULL),cast<uint8_t>(2ULL),cast<uint8_t>(3ULL),cast<uint8_t>(6ULL),cast<uint8_t>(7ULL),cast<uint8_t>(10ULL),cast<uint8_t>(11ULL),cast<uint8_t>(14ULL),cast<uint8_t>(15ULL),cast<uint8_t>(2ULL),cast<uint8_t>(3ULL),cast<uint8_t>(6ULL),cast<uint8_t>(7ULL),cast<uint8_t>(10ULL),cast<uint8_t>(11ULL),cast<uint8_t>(14ULL),cast<uint8_t>(15ULL)};
std::array<uint8_t,32> ps2_columnHalfCT16=std::array<uint8_t,32>{cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),cast<uint8_t>(1ULL),cast<uint8_t>(1ULL),cast<uint8_t>(1ULL),cast<uint8_t>(1ULL),cast<uint8_t>(1ULL),cast<uint8_t>(1ULL),cast<uint8_t>(1ULL),cast<uint8_t>(1ULL),cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),cast<uint8_t>(1ULL),cast<uint8_t>(1ULL),cast<uint8_t>(1ULL),cast<uint8_t>(1ULL),cast<uint8_t>(1ULL),cast<uint8_t>(1ULL),cast<uint8_t>(1ULL),cast<uint8_t>(1ULL)};
constexpr int64_t ps2_idleArmEvery=4096ULL;
constexpr int64_t ps2_idleWatch=64ULL;
constexpr int64_t ps2_intcGS=0ULL;
constexpr int64_t ps2_intcSBUS=1ULL;
constexpr int64_t ps2_intcVBlankOn=2ULL;
constexpr int64_t ps2_intcVBlankOff=3ULL;
constexpr int64_t ps2_intcISTAT=268496896ULL;
constexpr int64_t ps2_intcIMASK=268496912ULL;
constexpr int64_t ps2_intcVIF0=4ULL;
constexpr int64_t ps2_intcVIF1=5ULL;
constexpr int64_t ps2_intcVU0=6ULL;
constexpr int64_t ps2_intcVU1=7ULL;
constexpr int64_t ps2_intcIPU=8ULL;
constexpr int64_t ps2_intcTimer0=9ULL;
constexpr int64_t ps2_intcTimer1=10ULL;
constexpr int64_t ps2_intrExitAddr=251658240ULL;
constexpr int64_t ps2_dmacChannelSIF0=5ULL;
constexpr int64_t ps2_iopRAMSizeBytes=2097152ULL;
constexpr int64_t ps2_iopSPRAMBase=528482304ULL;
constexpr int64_t ps2_iopSPRAMSize=1024ULL;
constexpr int64_t ps2_iopIOBase=528486400ULL;
constexpr int64_t ps2_iopIOEnd=530579456ULL;
constexpr int64_t ps2_iopModuleBase=4096ULL;
constexpr int64_t ps2_iopStackArea=2031616ULL;
constexpr int64_t ps2_iopIdleStack=2079744ULL;
constexpr int64_t ps2_iopIdleLoop=512ULL;
constexpr int64_t ps2_iopIdleInsn=268500991ULL;
constexpr int64_t ps2_iopStepRatio=8ULL;
constexpr int64_t ps2_iopTrailLen=24ULL;
constexpr int64_t ps2_iopIsRunLogLen=40ULL;
constexpr int64_t ps2_iopProfileEvery=4096ULL;
Slice<std::string> ps2_iopBootOrder=Slice<std::string>{std::string("TIMEMANI",8),std::string("THREADMAN",9),std::string("IOMAN",5),std::string("MODLOAD",7),std::string("ROMDRV",6),std::string("SIFMAN",6),std::string("SIFCMD",6),std::string("LOADFILE",8),std::string("CDVDMAN",7),std::string("CDVDFSV",7),std::string("FILEIO",6),std::string("EESYNC",6)};
const std::string ps2_iopBootImage=std::string("/DRIVERS/IOPRP221.IMG",21);
bool ps2_cdvdLog=(go_os_Getenv(std::string("PS2_CDVDLOG",11)) != std::string("",0));
constexpr int64_t ps2_cdvdBase=524296192ULL;
constexpr int64_t ps2_cdvdEnd=524296224ULL;
constexpr int64_t ps2_cdvdNCommand=4ULL;
constexpr int64_t ps2_cdvdNStatus=5ULL;
constexpr int64_t ps2_cdvdNMode=6ULL;
constexpr int64_t ps2_cdvdNAbort=7ULL;
constexpr int64_t ps2_cdvdIntr=8ULL;
constexpr int64_t ps2_cdvdDriveSt=10ULL;
constexpr int64_t ps2_cdvdTrayStat=11ULL;
constexpr int64_t ps2_cdvdDiscType=15ULL;
constexpr int64_t ps2_cdvdSCommand=22ULL;
constexpr int64_t ps2_cdvdSStatus=23ULL;
constexpr int64_t ps2_cdvdSResult=24ULL;
constexpr int64_t ps2_cdvdNCmdReadCD=6ULL;
constexpr int64_t ps2_cdvdNCmdReadDVD=8ULL;
constexpr int64_t ps2_cdvdRawSectorBytes=2064ULL;
constexpr int64_t ps2_cdvdSectorHeader=12ULL;
constexpr int64_t ps2_cdvdDVDDataStart=196608ULL;
constexpr int64_t ps2_cdvdSCmdVersion=3ULL;
constexpr int64_t ps2_cdvdSCmdReady=5ULL;
constexpr int64_t ps2_cdvdNStatusError=1ULL;
constexpr int64_t ps2_cdvdNStatusReady=64ULL;
constexpr int64_t ps2_cdvdNStatusBusy=128ULL;
constexpr int64_t ps2_cdvdSStatusNoData=64ULL;
constexpr int64_t ps2_cdvdSStatusBusy=128ULL;
constexpr int64_t ps2_cdvdIntrDone=1ULL;
constexpr int64_t ps2_cdvdIntrDataRdy=4ULL;
constexpr int64_t ps2_cdvdIntrLine=2ULL;
constexpr int64_t ps2_cdvdDriveReady=10ULL;
constexpr int64_t ps2_cdvdDriveBusy=6ULL;
constexpr int64_t ps2_cdvdDiscTypeDVD=20ULL;
constexpr int64_t ps2_cdvdDiscTypePS2CD=18ULL;
constexpr int64_t ps2_cdvdSectorBytes=2048ULL;
constexpr int64_t ps2_cdvdCmdLatency=2000ULL;
constexpr int64_t ps2_cdMaxBlocks=360000ULL;
constexpr int64_t ps2_iopDMA1Base=528486528ULL;
constexpr int64_t ps2_iopDMA1End=528486640ULL;
constexpr int64_t ps2_iopDPCR=528486640ULL;
constexpr int64_t ps2_iopDICR=528486644ULL;
constexpr int64_t ps2_iopDMA2Base=528487680ULL;
constexpr int64_t ps2_iopDMA2End=528487792ULL;
constexpr int64_t ps2_iopDPCR2=528487792ULL;
constexpr int64_t ps2_iopDICR2=528487796ULL;
constexpr int64_t ps2_iopDMACEN=528487800ULL;
constexpr int64_t ps2_iopDMACINTEN=528487804ULL;
constexpr int64_t ps2_iopDMAChannels=13ULL;
constexpr int64_t ps2_iopDMAChSIF2=2ULL;
constexpr int64_t ps2_iopDMAChCDVD=3ULL;
constexpr int64_t ps2_iopDMAChSPU0=4ULL;
constexpr int64_t ps2_iopDMAChSPU1=7ULL;
constexpr int64_t ps2_iopDMAChSIF0=9ULL;
constexpr int64_t ps2_iopDMAChSIF1=10ULL;
constexpr int64_t ps2_iopDMAMadr=0ULL;
constexpr int64_t ps2_iopDMABcr=4ULL;
constexpr int64_t ps2_iopDMAChcr=8ULL;
constexpr int64_t ps2_iopDMATadr=12ULL;
constexpr int64_t ps2_iopChcrFromRAM=1ULL;
constexpr int64_t ps2_iopChcrStart=16777216ULL;
constexpr int64_t ps2_iopChcrTrigger=268435456ULL;
constexpr int64_t ps2_iopChcrSyncMask=1536ULL;
constexpr int64_t ps2_iopDMALatency=400ULL;
constexpr int64_t ps2_iopFrameSize=184ULL;
constexpr int64_t ps2_iopFrameTag=0ULL;
constexpr int64_t ps2_iopFrameHI=128ULL;
constexpr int64_t ps2_iopFrameLO=132ULL;
constexpr int64_t ps2_iopFrameSR=136ULL;
constexpr int64_t ps2_iopFrameEPC=140ULL;
constexpr int64_t ps2_iopFrameWaitResult=8ULL;
constexpr int64_t ps2_iopFrameFresh=4294967294ULL;
constexpr int64_t ps2_iopIRQs=64ULL;
constexpr int64_t ps2_iopIntrStack=2063360ULL;
constexpr int64_t ps2_iopAllocLow=0ULL;
constexpr int64_t ps2_iopAllocHigh=1ULL;
constexpr int64_t ps2_iopAllocAddr=2ULL;
constexpr int64_t ps2_iopCtypeDigit=4ULL;
constexpr int64_t ps2_iopCStringMax=1024ULL;
constexpr int64_t ps2_iopSyscallReschedule=32ULL;
Map<std::string,ps2_iopLibrary*> ps2_goLibraries=Map<std::string,ps2_iopLibrary*>{};
constexpr int64_t ps2_iopModInfoInit=12ULL;
constexpr int64_t ps2_iopModInfoText=16ULL;
constexpr int64_t ps2_iopModInfoSize=28ULL;
constexpr int64_t ps2_iopModRecEntry=16ULL;
constexpr int64_t ps2_iopModRecGP=20ULL;
constexpr int64_t ps2_iopModRecArg0=24ULL;
constexpr int64_t ps2_iopModRecArg1=28ULL;
constexpr int64_t ps2_iopModRecSize=48ULL;
constexpr int64_t ps2_iopModRelocatable=2ULL;
constexpr uint32_t ps2_iopModBadImage=4294967095ULL;
constexpr int64_t ps2_iopReturnSentinel=268369920ULL;
constexpr int64_t ps2_iopCallStack=2096128ULL;
constexpr int64_t ps2_iopCallBudget=200000000ULL;
constexpr int64_t ps2_iopSpinWindow=1048576ULL;
constexpr int64_t ps2_iopSpinDistinct=8ULL;
constexpr int64_t ps2_iopSIO2Base=528515584ULL;
constexpr int64_t ps2_iopSIO2End=528515716ULL;
constexpr int64_t ps2_sio2SEND3=528515584ULL;
constexpr int64_t ps2_sio2DATAin=528515680ULL;
constexpr int64_t ps2_sio2CTRL=528515688ULL;
constexpr int64_t ps2_sio2RECV1=528515692ULL;
constexpr int64_t ps2_sio2DATAout=528515684ULL;
constexpr int64_t ps2_iopSIO2IRQ=17ULL;
constexpr int64_t ps2_sio2NoDevice=119040ULL;
constexpr int64_t ps2_sio2Device=4352ULL;
constexpr int64_t ps2_iopSPU2Base=529530880ULL;
constexpr int64_t ps2_iopSPU2End=529532928ULL;
constexpr int64_t ps2_iopSPU2RAMSize=2097152ULL;
constexpr int64_t ps2_iopSPU2CoreSpan=1024ULL;
constexpr int64_t ps2_iopSPU2TSAHi=424ULL;
constexpr int64_t ps2_iopSPU2TSALo=426ULL;
constexpr int64_t ps2_iopSPU2Stat=1986ULL;
constexpr int64_t ps2_iopSPU2CoreStat=836ULL;
constexpr int64_t ps2_iopSPU2CoreDone=128ULL;
constexpr int64_t ps2_iopSPU2Pitch=4ULL;
constexpr int64_t ps2_iopSPU2KON=416ULL;
constexpr int64_t ps2_iopSPU2KOFF=420ULL;
constexpr int64_t ps2_iopSPU2SSA=448ULL;
constexpr int64_t ps2_iopSPU2LSAX=452ULL;
constexpr int64_t ps2_iopSPU2NAX=456ULL;
constexpr int64_t ps2_iopSPU2VoiceStride=16ULL;
constexpr int64_t ps2_iopSPU2AddrStride=12ULL;
constexpr int64_t ps2_iopSPU2Voices=24ULL;
constexpr int64_t ps2_iopSPU2Cores=2ULL;
constexpr int64_t ps2_iopSPU2BlockWords=8ULL;
constexpr int64_t ps2_iopSPU2BlockSamples=28ULL;
constexpr int64_t ps2_spuSampleRate=48000ULL;
constexpr int64_t ps2_spuFieldRate=60ULL;
constexpr int64_t ps2_spuIOPStepsPerField=125000ULL;
constexpr int64_t ps2_spuAccPerBlock=17920000ULL;
constexpr int64_t ps2_spuMaxBlocksPerAdvance=16384ULL;
constexpr int64_t ps2_iopIRQSPU=9ULL;
constexpr int64_t ps2_iopTCBStatus=12ULL;
constexpr int64_t ps2_iopTCBFrame=16ULL;
constexpr int64_t ps2_iopTCBWaitTyp=28ULL;
constexpr int64_t ps2_iopTCBWaitObj=32ULL;
constexpr int64_t ps2_iopTCBWaitMsk=40ULL;
constexpr int64_t ps2_iopTCBChain=36ULL;
constexpr int64_t ps2_iopTCBEntry=56ULL;
constexpr int64_t ps2_iopTCBStack=60ULL;
constexpr int64_t ps2_iopTCBStackSz=64ULL;
Map<uint32_t,std::string> ps2_iopThreadStates=Map<uint32_t,std::string>{{cast<uint32_t>(1ULL),std::string("RUN",3)},{cast<uint32_t>(2ULL),std::string("READY",5)},{cast<uint32_t>(4ULL),std::string("WAIT",4)},{cast<uint32_t>(8ULL),std::string("SUSPEND",7)},{cast<uint32_t>(12ULL),std::string("WAIT|SUSPEND",12)},{cast<uint32_t>(16ULL),std::string("DORMANT",7)}};
constexpr int64_t ps2_iopTimerLoBase=528486656ULL;
constexpr int64_t ps2_iopTimerLoEnd=528486704ULL;
constexpr int64_t ps2_iopTimerHiBase=528487552ULL;
constexpr int64_t ps2_iopTimerHiEnd=528487600ULL;
constexpr int64_t ps2_iopTimers=6ULL;
constexpr int64_t ps2_iopTimerCount=0ULL;
constexpr int64_t ps2_iopTimerMode=4ULL;
constexpr int64_t ps2_iopTimerTarget=8ULL;
constexpr int64_t ps2_iopTimerResetOnTarget=8ULL;
constexpr int64_t ps2_iopTimerIRQOnTarget=16ULL;
constexpr int64_t ps2_iopTimerIRQOnOverflow=32ULL;
constexpr int64_t ps2_iopTimerRepeat=64ULL;
constexpr int64_t ps2_iopTimerHitTarget=2048ULL;
constexpr int64_t ps2_iopTimerHitOverflow=4096ULL;
constexpr int64_t ps2_iopVblankLine=0ULL;
constexpr int64_t ps2_irxImportMagic=1105199104ULL;
constexpr int64_t ps2_irxExportMagic=1103101952ULL;
constexpr int64_t ps2_irxStubJR=65011720ULL;
constexpr int64_t ps2_irxExportHooks=4ULL;
constexpr int64_t ps2_rMIPS32=2ULL;
constexpr int64_t ps2_rMIPS26=4ULL;
constexpr int64_t ps2_rMIPSHI16=5ULL;
constexpr int64_t ps2_rMIPSLO16=6ULL;
Map<uint32_t,ps2_syscallEntry> ps2_syscalls=Map<uint32_t,ps2_syscallEntry>{{cast<uint32_t>(0ULL),ps2_syscallEntry{std::string("RFU000_FullReset",16),{}}},{cast<uint32_t>(1ULL),ps2_syscallEntry{std::string("ResetEE",7),[](ps2_Machine* m)->void{
ps2_Machine_setRet(m,cast<uint32_t>(1ULL));
}
}},{cast<uint32_t>(2ULL),ps2_syscallEntry{std::string("SetGsCrt",8),[](ps2_Machine* m)->void{
auto tmp311 = std::make_tuple(ps2_Machine_arg(m,cast<uint32_t>(0ULL)),ps2_Machine_arg(m,cast<uint32_t>(1ULL)),ps2_Machine_arg(m,cast<uint32_t>(2ULL)));
m->gsInterlace = std::get<0>(tmp311);
m->gsVideoMode = std::get<1>(tmp311);
m->gsFieldMode = std::get<2>(tmp311);
ps2_Machine_note(m,std::string("SetGsCrt: interlace=%d mode=%d field=%d",39),ps2_Machine_arg(m,cast<uint32_t>(0ULL)),ps2_Machine_arg(m,cast<uint32_t>(1ULL)),ps2_Machine_arg(m,cast<uint32_t>(2ULL)));
ps2_Machine_setRet(m,cast<uint32_t>(0ULL));
}
}},{cast<uint32_t>(4ULL),ps2_syscallEntry{std::string("Exit",4),[](ps2_Machine* m)->void{
ps2_Machine_Halt(m,std::string("the game called Exit(%d)",24),ps2_Machine_arg(m,cast<uint32_t>(0ULL)));
}
}},{cast<uint32_t>(6ULL),ps2_syscallEntry{std::string("LoadExecPS2",11),{}}},{cast<uint32_t>(7ULL),ps2_syscallEntry{std::string("ExecPS2",7),{}}},{cast<uint32_t>(10ULL),ps2_syscallEntry{std::string("AddSbusIntcHandler",18),{}}},{cast<uint32_t>(11ULL),ps2_syscallEntry{std::string("RemoveSbusIntcHandler",21),{}}},{cast<uint32_t>(12ULL),ps2_syscallEntry{std::string("Interrupt2Iop",13),[](ps2_Machine* m)->void{
ps2_Machine_setRet(m,cast<uint32_t>(0ULL));
}
}},{cast<uint32_t>(13ULL),ps2_syscallEntry{std::string("SetVTLBRefillHandler",20),[](ps2_Machine* m)->void{
ps2_Machine_setRet(m,cast<uint32_t>(0ULL));
}
}},{cast<uint32_t>(14ULL),ps2_syscallEntry{std::string("SetVCommonHandler",17),[](ps2_Machine* m)->void{
ps2_Machine_setRet(m,cast<uint32_t>(0ULL));
}
}},{cast<uint32_t>(15ULL),ps2_syscallEntry{std::string("SetVInterruptHandler",20),[](ps2_Machine* m)->void{
ps2_Machine_setRet(m,cast<uint32_t>(0ULL));
}
}},{cast<uint32_t>(16ULL),ps2_syscallEntry{std::string("AddIntcHandler",14),[](ps2_Machine* m)->void{
ps2_Machine_addIntcHandler(m);
}
}},{cast<uint32_t>(17ULL),ps2_syscallEntry{std::string("RemoveIntcHandler",17),[](ps2_Machine* m)->void{
ps2_Machine_setRet(m,cast<uint32_t>(0ULL));
}
}},{cast<uint32_t>(18ULL),ps2_syscallEntry{std::string("AddDmacHandler",14),[](ps2_Machine* m)->void{
ps2_Machine_addDmacHandler(m);
}
}},{cast<uint32_t>(19ULL),ps2_syscallEntry{std::string("RemoveDmacHandler",17),[](ps2_Machine* m)->void{
ps2_Machine_setRet(m,cast<uint32_t>(0ULL));
}
}},{cast<uint32_t>(20ULL),ps2_syscallEntry{std::string("EnableIntc",10),[](ps2_Machine* m)->void{
m->intcMask |= shl<uint32_t>(cast<uint32_t>(1ULL),(cast<uint32_t>((ps2_Machine_arg(m,cast<uint32_t>(0ULL)) & cast<uint32_t>(31ULL)))));
ps2_Machine_setRet(m,cast<uint32_t>(1ULL));
}
}},{cast<uint32_t>(21ULL),ps2_syscallEntry{std::string("DisableIntc",11),[](ps2_Machine* m)->void{
m->intcMask &= ~(shl<uint32_t>(cast<uint32_t>(1ULL),(cast<uint32_t>((ps2_Machine_arg(m,cast<uint32_t>(0ULL)) & cast<uint32_t>(31ULL))))));
ps2_Machine_setRet(m,cast<uint32_t>(1ULL));
}
}},{cast<uint32_t>(22ULL),ps2_syscallEntry{std::string("EnableDmac",10),[](ps2_Machine* m)->void{
m->dmacMask |= shl<uint32_t>(cast<uint32_t>(1ULL),(cast<uint32_t>((ps2_Machine_arg(m,cast<uint32_t>(0ULL)) & cast<uint32_t>(31ULL)))));
ps2_Machine_dmacRetrigger(m);
ps2_Machine_setRet(m,cast<uint32_t>(1ULL));
}
}},{cast<uint32_t>(23ULL),ps2_syscallEntry{std::string("DisableDmac",11),[](ps2_Machine* m)->void{
m->dmacMask &= ~(shl<uint32_t>(cast<uint32_t>(1ULL),(cast<uint32_t>((ps2_Machine_arg(m,cast<uint32_t>(0ULL)) & cast<uint32_t>(31ULL))))));
ps2_Machine_setRet(m,cast<uint32_t>(1ULL));
}
}},{cast<uint32_t>(24ULL),ps2_syscallEntry{std::string("SetAlarm",8),{}}},{cast<uint32_t>(25ULL),ps2_syscallEntry{std::string("ReleaseAlarm",12),{}}},{cast<uint32_t>(26ULL),ps2_syscallEntry{std::string("iEnableIntc",11),[](ps2_Machine* m)->void{
m->intcMask |= shl<uint32_t>(cast<uint32_t>(1ULL),(cast<uint32_t>((ps2_Machine_arg(m,cast<uint32_t>(0ULL)) & cast<uint32_t>(31ULL)))));
ps2_Machine_setRet(m,cast<uint32_t>(1ULL));
}
}},{cast<uint32_t>(27ULL),ps2_syscallEntry{std::string("iDisableIntc",12),[](ps2_Machine* m)->void{
m->intcMask &= ~(shl<uint32_t>(cast<uint32_t>(1ULL),(cast<uint32_t>((ps2_Machine_arg(m,cast<uint32_t>(0ULL)) & cast<uint32_t>(31ULL))))));
ps2_Machine_setRet(m,cast<uint32_t>(1ULL));
}
}},{cast<uint32_t>(28ULL),ps2_syscallEntry{std::string("iEnableDmac",11),[](ps2_Machine* m)->void{
m->dmacMask |= shl<uint32_t>(cast<uint32_t>(1ULL),(cast<uint32_t>((ps2_Machine_arg(m,cast<uint32_t>(0ULL)) & cast<uint32_t>(31ULL)))));
ps2_Machine_dmacRetrigger(m);
ps2_Machine_setRet(m,cast<uint32_t>(1ULL));
}
}},{cast<uint32_t>(29ULL),ps2_syscallEntry{std::string("iDisableDmac",12),[](ps2_Machine* m)->void{
m->dmacMask &= ~(shl<uint32_t>(cast<uint32_t>(1ULL),(cast<uint32_t>((ps2_Machine_arg(m,cast<uint32_t>(0ULL)) & cast<uint32_t>(31ULL))))));
ps2_Machine_setRet(m,cast<uint32_t>(1ULL));
}
}},{cast<uint32_t>(32ULL),ps2_syscallEntry{std::string("CreateThread",12),[](ps2_Machine* m)->void{
ps2_Machine_createThread(m);
}
}},{cast<uint32_t>(33ULL),ps2_syscallEntry{std::string("DeleteThread",12),[](ps2_Machine* m)->void{
ps2_Machine_setRet(m,cast<uint32_t>(0ULL));
}
}},{cast<uint32_t>(34ULL),ps2_syscallEntry{std::string("StartThread",11),[](ps2_Machine* m)->void{
ps2_Machine_startThread(m);
}
}},{cast<uint32_t>(35ULL),ps2_syscallEntry{std::string("ExitThread",10),[](ps2_Machine* m)->void{
ps2_Machine_exitThread(m);
}
}},{cast<uint32_t>(36ULL),ps2_syscallEntry{std::string("ExitDeleteThread",16),[](ps2_Machine* m)->void{
ps2_Machine_exitThread(m);
}
}},{cast<uint32_t>(37ULL),ps2_syscallEntry{std::string("TerminateThread",15),[](ps2_Machine* m)->void{
ps2_Machine_setRet(m,cast<uint32_t>(0ULL));
}
}},{cast<uint32_t>(39ULL),ps2_syscallEntry{std::string("DisableDispatchThread",21),[](ps2_Machine* m)->void{
ps2_Machine_setRet(m,cast<uint32_t>(0ULL));
}
}},{cast<uint32_t>(40ULL),ps2_syscallEntry{std::string("EnableDispatchThread",20),[](ps2_Machine* m)->void{
ps2_Machine_setRet(m,cast<uint32_t>(0ULL));
}
}},{cast<uint32_t>(41ULL),ps2_syscallEntry{std::string("ChangeThreadPriority",20),[](ps2_Machine* m)->void{
ps2_Machine_setRet(m,cast<uint32_t>(0ULL));
}
}},{cast<uint32_t>(43ULL),ps2_syscallEntry{std::string("RotateThreadReadyQueue",22),[](ps2_Machine* m)->void{
ps2_Machine_setRet(m,cast<uint32_t>(0ULL));
}
}},{cast<uint32_t>(45ULL),ps2_syscallEntry{std::string("ReleaseWaitThread",17),[](ps2_Machine* m)->void{
ps2_Machine_wakeupThread(m,ps2_Machine_arg(m,cast<uint32_t>(0ULL)));
}
}},{cast<uint32_t>(47ULL),ps2_syscallEntry{std::string("GetThreadId",11),[](ps2_Machine* m)->void{
ps2_Machine_setRet(m,m->currentThread);
}
}},{cast<uint32_t>(48ULL),ps2_syscallEntry{std::string("ReferThreadStatus",17),[](ps2_Machine* m)->void{
ps2_Machine_referThreadStatus(m);
}
}},{cast<uint32_t>(50ULL),ps2_syscallEntry{std::string("SleepThread",11),[](ps2_Machine* m)->void{
ps2_Machine_sleepThread(m);
}
}},{cast<uint32_t>(51ULL),ps2_syscallEntry{std::string("WakeupThread",12),[](ps2_Machine* m)->void{
ps2_Machine_wakeupThread(m,ps2_Machine_arg(m,cast<uint32_t>(0ULL)));
}
}},{cast<uint32_t>(52ULL),ps2_syscallEntry{std::string("iWakeupThread",13),[](ps2_Machine* m)->void{
ps2_Machine_wakeupThread(m,ps2_Machine_arg(m,cast<uint32_t>(0ULL)));
}
}},{cast<uint32_t>(53ULL),ps2_syscallEntry{std::string("CancelWakeupThread",18),[](ps2_Machine* m)->void{
ps2_Machine_setRet(m,cast<uint32_t>(0ULL));
}
}},{cast<uint32_t>(55ULL),ps2_syscallEntry{std::string("SuspendThread",13),[](ps2_Machine* m)->void{
ps2_Machine_setRet(m,cast<uint32_t>(0ULL));
}
}},{cast<uint32_t>(57ULL),ps2_syscallEntry{std::string("ResumeThread",12),[](ps2_Machine* m)->void{
ps2_Machine_wakeupThread(m,ps2_Machine_arg(m,cast<uint32_t>(0ULL)));
}
}},{cast<uint32_t>(59ULL),ps2_syscallEntry{std::string("JoinThread",10),[](ps2_Machine* m)->void{
ps2_Machine_setRet(m,cast<uint32_t>(0ULL));
}
}},{cast<uint32_t>(60ULL),ps2_syscallEntry{std::string("SetupThread",11),[](ps2_Machine* m)->void{
ps2_Machine_setupThread(m);
}
}},{cast<uint32_t>(61ULL),ps2_syscallEntry{std::string("SetupHeap",9),[](ps2_Machine* m)->void{
ps2_Machine_setupHeap(m);
}
}},{cast<uint32_t>(62ULL),ps2_syscallEntry{std::string("EndOfHeap",9),[](ps2_Machine* m)->void{
ps2_Machine_setRet(m,m->heapEnd);
}
}},{cast<uint32_t>(64ULL),ps2_syscallEntry{std::string("CreateSema",10),[](ps2_Machine* m)->void{
ps2_Machine_createSema(m);
}
}},{cast<uint32_t>(65ULL),ps2_syscallEntry{std::string("DeleteSema",10),[](ps2_Machine* m)->void{
ps2_Machine_setRet(m,cast<uint32_t>(0ULL));
}
}},{cast<uint32_t>(66ULL),ps2_syscallEntry{std::string("SignalSema",10),[](ps2_Machine* m)->void{
ps2_Machine_signalSema(m,ps2_Machine_arg(m,cast<uint32_t>(0ULL)));
}
}},{cast<uint32_t>(67ULL),ps2_syscallEntry{std::string("iSignalSema",11),[](ps2_Machine* m)->void{
ps2_Machine_signalSema(m,ps2_Machine_arg(m,cast<uint32_t>(0ULL)));
}
}},{cast<uint32_t>(68ULL),ps2_syscallEntry{std::string("WaitSema",8),[](ps2_Machine* m)->void{
ps2_Machine_waitSema(m,ps2_Machine_arg(m,cast<uint32_t>(0ULL)));
}
}},{cast<uint32_t>(69ULL),ps2_syscallEntry{std::string("PollSema",8),[](ps2_Machine* m)->void{
ps2_Machine_pollSema(m,ps2_Machine_arg(m,cast<uint32_t>(0ULL)));
}
}},{cast<uint32_t>(70ULL),ps2_syscallEntry{std::string("iPollSema",9),[](ps2_Machine* m)->void{
ps2_Machine_pollSema(m,ps2_Machine_arg(m,cast<uint32_t>(0ULL)));
}
}},{cast<uint32_t>(71ULL),ps2_syscallEntry{std::string("ReferSemaStatus",15),{}}},{cast<uint32_t>(74ULL),ps2_syscallEntry{std::string("SetOsdConfigParam",17),{}}},{cast<uint32_t>(75ULL),ps2_syscallEntry{std::string("GetOsdConfigParam",17),[](ps2_Machine* m)->void{
ps2_Machine_setRet(m,cast<uint32_t>(0ULL));
}
}},{cast<uint32_t>(76ULL),ps2_syscallEntry{std::string("GetGsHParam",11),{}}},{cast<uint32_t>(77ULL),ps2_syscallEntry{std::string("GetGsVParam",11),{}}},{cast<uint32_t>(78ULL),ps2_syscallEntry{std::string("SetGsHParam",11),{}}},{cast<uint32_t>(79ULL),ps2_syscallEntry{std::string("SetGsVParam",11),{}}},{cast<uint32_t>(85ULL),ps2_syscallEntry{std::string("PutTLBEntry",11),[](ps2_Machine* m)->void{
ps2_Machine_setRet(m,cast<uint32_t>(0ULL));
}
}},{cast<uint32_t>(86ULL),ps2_syscallEntry{std::string("SetTLBEntry",11),[](ps2_Machine* m)->void{
ps2_Machine_setRet(m,cast<uint32_t>(0ULL));
}
}},{cast<uint32_t>(87ULL),ps2_syscallEntry{std::string("GetTLBEntry",11),[](ps2_Machine* m)->void{
ps2_Machine_setRet(m,cast<uint32_t>(0ULL));
}
}},{cast<uint32_t>(88ULL),ps2_syscallEntry{std::string("ProbeTLBEntry",13),[](ps2_Machine* m)->void{
ps2_Machine_setRet(m,cast<uint32_t>(0ULL));
}
}},{cast<uint32_t>(89ULL),ps2_syscallEntry{std::string("ExpandScratchPad",16),[](ps2_Machine* m)->void{
ps2_Machine_setRet(m,cast<uint32_t>(0ULL));
}
}},{cast<uint32_t>(90ULL),ps2_syscallEntry{std::string("Copy",4),{}}},{cast<uint32_t>(91ULL),ps2_syscallEntry{std::string("GetEntryAddress",15),[](ps2_Machine* m)->void{
ps2_Machine_setRet(m,cast<uint32_t>(0ULL));
}
}},{cast<uint32_t>(92ULL),ps2_syscallEntry{std::string("EnableIntcHandler",17),[](ps2_Machine* m)->void{
m->intcMask |= shl<uint32_t>(cast<uint32_t>(1ULL),(cast<uint32_t>((ps2_Machine_arg(m,cast<uint32_t>(0ULL)) & cast<uint32_t>(31ULL)))));
ps2_Machine_setRet(m,cast<uint32_t>(1ULL));
}
}},{cast<uint32_t>(93ULL),ps2_syscallEntry{std::string("DisableIntcHandler",18),[](ps2_Machine* m)->void{
m->intcMask &= ~(shl<uint32_t>(cast<uint32_t>(1ULL),(cast<uint32_t>((ps2_Machine_arg(m,cast<uint32_t>(0ULL)) & cast<uint32_t>(31ULL))))));
ps2_Machine_setRet(m,cast<uint32_t>(1ULL));
}
}},{cast<uint32_t>(94ULL),ps2_syscallEntry{std::string("EnableDmacHandler",17),[](ps2_Machine* m)->void{
m->dmacMask |= shl<uint32_t>(cast<uint32_t>(1ULL),(cast<uint32_t>((ps2_Machine_arg(m,cast<uint32_t>(0ULL)) & cast<uint32_t>(31ULL)))));
ps2_Machine_dmacRetrigger(m);
ps2_Machine_setRet(m,cast<uint32_t>(1ULL));
}
}},{cast<uint32_t>(95ULL),ps2_syscallEntry{std::string("DisableDmacHandler",18),[](ps2_Machine* m)->void{
m->dmacMask &= ~(shl<uint32_t>(cast<uint32_t>(1ULL),(cast<uint32_t>((ps2_Machine_arg(m,cast<uint32_t>(0ULL)) & cast<uint32_t>(31ULL))))));
ps2_Machine_setRet(m,cast<uint32_t>(1ULL));
}
}},{cast<uint32_t>(96ULL),ps2_syscallEntry{std::string("KSeg0",5),[](ps2_Machine* m)->void{
ps2_Machine_setRet(m,cast<uint32_t>(0ULL));
}
}},{cast<uint32_t>(97ULL),ps2_syscallEntry{std::string("EnableCache",11),[](ps2_Machine* m)->void{
ps2_Machine_setRet(m,cast<uint32_t>(0ULL));
}
}},{cast<uint32_t>(98ULL),ps2_syscallEntry{std::string("DisableCache",12),[](ps2_Machine* m)->void{
ps2_Machine_setRet(m,cast<uint32_t>(0ULL));
}
}},{cast<uint32_t>(99ULL),ps2_syscallEntry{std::string("GetCop0",7),[](ps2_Machine* m)->void{
ps2_Machine_setRet(m,cast<uint32_t>(m->CPU->COP0[cast<uint32_t>((ps2_Machine_arg(m,cast<uint32_t>(0ULL)) & cast<uint32_t>(31ULL)))]));
}
}},{cast<uint32_t>(100ULL),ps2_syscallEntry{std::string("FlushCache",10),[](ps2_Machine* m)->void{
ps2_Machine_setRet(m,cast<uint32_t>(0ULL));
}
}},{cast<uint32_t>(102ULL),ps2_syscallEntry{std::string("CpuConfig",9),[](ps2_Machine* m)->void{
ps2_Machine_setRet(m,cast<uint32_t>(0ULL));
}
}},{cast<uint32_t>(104ULL),ps2_syscallEntry{std::string("iFlushCache",11),[](ps2_Machine* m)->void{
ps2_Machine_setRet(m,cast<uint32_t>(0ULL));
}
}},{cast<uint32_t>(107ULL),ps2_syscallEntry{std::string("sceSifStopDma",13),[](ps2_Machine* m)->void{
ps2_Machine_setRet(m,cast<uint32_t>(0ULL));
}
}},{cast<uint32_t>(108ULL),ps2_syscallEntry{std::string("SetCPUTimerHandler",18),[](ps2_Machine* m)->void{
ps2_Machine_setRet(m,cast<uint32_t>(0ULL));
}
}},{cast<uint32_t>(109ULL),ps2_syscallEntry{std::string("SetCPUTimer",11),[](ps2_Machine* m)->void{
ps2_Machine_setRet(m,cast<uint32_t>(0ULL));
}
}},{cast<uint32_t>(110ULL),ps2_syscallEntry{std::string("SetOsdConfigParam2",18),{}}},{cast<uint32_t>(111ULL),ps2_syscallEntry{std::string("GetOsdConfigParam2",18),[](ps2_Machine* m)->void{
ps2_Machine_setRet(m,cast<uint32_t>(0ULL));
}
}},{cast<uint32_t>(112ULL),ps2_syscallEntry{std::string("GsGetIMR",8),[](ps2_Machine* m)->void{
ps2_Machine_setRet(m,m->gsIMR);
}
}},{cast<uint32_t>(113ULL),ps2_syscallEntry{std::string("GsPutIMR",8),[](ps2_Machine* m)->void{
m->gsIMR = ps2_Machine_arg(m,cast<uint32_t>(0ULL));
ps2_Machine_setRet(m,cast<uint32_t>(0ULL));
}
}},{cast<uint32_t>(114ULL),ps2_syscallEntry{std::string("SetPgifHandler",14),[](ps2_Machine* m)->void{
ps2_Machine_setRet(m,cast<uint32_t>(0ULL));
}
}},{cast<uint32_t>(115ULL),ps2_syscallEntry{std::string("SetVSyncFlag",12),[](ps2_Machine* m)->void{
auto tmp312 = std::make_tuple(ps2_Machine_arg(m,cast<uint32_t>(0ULL)),ps2_Machine_arg(m,cast<uint32_t>(1ULL)));
m->vsyncFlagPtr = std::get<0>(tmp312);
m->vsyncFlag2Ptr = std::get<1>(tmp312);
ps2_Machine_note(m,std::string("SetVSyncFlag: flags at 0x%08X and 0x%08X",40),ps2_Machine_arg(m,cast<uint32_t>(0ULL)),ps2_Machine_arg(m,cast<uint32_t>(1ULL)));
ps2_Machine_setRet(m,cast<uint32_t>(0ULL));
}
}},{cast<uint32_t>(116ULL),ps2_syscallEntry{std::string("SetSyscall",10),[](ps2_Machine* m)->void{
ps2_Machine_setSyscall(m);
}
}},{cast<uint32_t>(118ULL),ps2_syscallEntry{std::string("sceSifDmaStat",13),[](ps2_Machine* m)->void{
ps2_Machine_setRet(m,cast<uint32_t>(4294967295ULL));
}
}},{cast<uint32_t>(119ULL),ps2_syscallEntry{std::string("sceSifSetDma",12),[](ps2_Machine* m)->void{
ps2_Machine_sifSetDma(m);
}
}},{cast<uint32_t>(120ULL),ps2_syscallEntry{std::string("sceSifSetDChain",15),[](ps2_Machine* m)->void{
ps2_Machine_setRet(m,cast<uint32_t>(0ULL));
}
}},{cast<uint32_t>(121ULL),ps2_syscallEntry{std::string("sceSifSetReg",12),[](ps2_Machine* m)->void{
ps2_Machine_sifSetReg(m);
}
}},{cast<uint32_t>(122ULL),ps2_syscallEntry{std::string("sceSifGetReg",12),[](ps2_Machine* m)->void{
ps2_Machine_sifGetReg(m);
}
}},{cast<uint32_t>(123ULL),ps2_syscallEntry{std::string("ExecOSD",7),{}}},{cast<uint32_t>(124ULL),ps2_syscallEntry{std::string("Deci2Call",9),[](ps2_Machine* m)->void{
ps2_Machine_deci2Call(m);
}
}},{cast<uint32_t>(125ULL),ps2_syscallEntry{std::string("PSMode",6),[](ps2_Machine* m)->void{
ps2_Machine_setRet(m,cast<uint32_t>(0ULL));
}
}},{cast<uint32_t>(126ULL),ps2_syscallEntry{std::string("MachineType",11),[](ps2_Machine* m)->void{
ps2_Machine_setRet(m,cast<uint32_t>(0ULL));
}
}},{cast<uint32_t>(127ULL),ps2_syscallEntry{std::string("GetMemorySize",13),[](ps2_Machine* m)->void{
ps2_Machine_setRet(m,cast<uint32_t>(33554432ULL));
}
}}};
constexpr int64_t ps2_argBlockSize=672ULL;
constexpr int64_t ps2_ramBase=0ULL;
constexpr int64_t ps2_ramSize=33554432ULL;
constexpr int64_t ps2_spramBase=1879048192ULL;
constexpr int64_t ps2_spramSize=16384ULL;
constexpr int64_t ps2_ioBase=268435456ULL;
constexpr int64_t ps2_ioEnd=268500992ULL;
constexpr int64_t ps2_vuBase=285212672ULL;
constexpr int64_t ps2_vuEnd=285278208ULL;
constexpr int64_t ps2_gsRegBase=301989888ULL;
constexpr int64_t ps2_gsRegEnd=301998080ULL;
constexpr int64_t ps2_iopRAMBase=469762048ULL;
constexpr int64_t ps2_iopRAMSize=2097152ULL;
constexpr int64_t ps2_stepsPerVBlank=1000000ULL;
constexpr int64_t ps2_PadStickCentre=128ULL;
constexpr int64_t ps2_PadStickFull=127ULL;
Map<std::string,uint16_t> ps2_padButtonBits=Map<std::string,uint16_t>{{std::string("select",6),cast<uint16_t>(1ULL)},{std::string("l3",2),cast<uint16_t>(2ULL)},{std::string("r3",2),cast<uint16_t>(4ULL)},{std::string("start",5),cast<uint16_t>(8ULL)},{std::string("up",2),cast<uint16_t>(16ULL)},{std::string("right",5),cast<uint16_t>(32ULL)},{std::string("down",4),cast<uint16_t>(64ULL)},{std::string("left",4),cast<uint16_t>(128ULL)},{std::string("l2",2),cast<uint16_t>(256ULL)},{std::string("r2",2),cast<uint16_t>(512ULL)},{std::string("l1",2),cast<uint16_t>(1024ULL)},{std::string("r1",2),cast<uint16_t>(2048ULL)},{std::string("triangle",8),cast<uint16_t>(4096ULL)},{std::string("circle",6),cast<uint16_t>(8192ULL)},{std::string("cross",5),cast<uint16_t>(16384ULL)},{std::string("x",1),cast<uint16_t>(16384ULL)},{std::string("square",6),cast<uint16_t>(32768ULL)}};
constexpr int64_t ps2_bucketDrain=0ULL;
constexpr int64_t ps2_bucketVU1=1ULL;
constexpr int64_t ps2_bucketRaster=2ULL;
constexpr int64_t ps2_numBuckets=3ULL;
constexpr int64_t ps2_romEntrySize=16ULL;
constexpr int64_t ps2_spinWindow=2097152ULL;
constexpr int64_t ps2_spinDistinct=6ULL;
constexpr int64_t ps2_trailLen=24ULL;
Map<uint32_t,std::string> ps2_exceptionVectors=Map<uint32_t,std::string>{{cast<uint32_t>(2147483648ULL),std::string("TLB refill",10)},{cast<uint32_t>(2147483776ULL),std::string("counter",7)},{cast<uint32_t>(2147484032ULL),std::string("general",7)}};
constexpr ps2_threadState ps2_thRunning=0ULL;
constexpr ps2_threadState ps2_thReady=1ULL;
constexpr ps2_threadState ps2_thSleeping=2ULL;
constexpr ps2_threadState ps2_thWaitSema=3ULL;
constexpr ps2_threadState ps2_thDormant=4ULL;
constexpr ps2_threadState ps2_thDead=5ULL;
constexpr int64_t ps2_threadExitAddr=252706816ULL;
bool ps2_sifLog=(go_os_Getenv(std::string("PS2_SIFLOG",10)) != std::string("",0));
constexpr int64_t ps2_sifCmdChangeSaddr=2147483648ULL;
constexpr int64_t ps2_sifCmdSetSreg=2147483649ULL;
constexpr int64_t ps2_sifCmdInitCmd=2147483650ULL;
constexpr int64_t ps2_sifCmdReset=2147483651ULL;
constexpr int64_t ps2_sifCmdRpcEnd=2147483656ULL;
constexpr int64_t ps2_sifCmdRpcBind=2147483657ULL;
constexpr int64_t ps2_sifCmdRpcCall=2147483658ULL;
constexpr int64_t ps2_sifCmdRpcRData=2147483660ULL;
constexpr int64_t ps2_sifDmaInt=64ULL;
constexpr int64_t ps2_iopRebootLatency=100000ULL;
constexpr int64_t ps2_sifChainTagSize=16ULL;
constexpr int64_t ps2_sifChainAddrMask=16777215ULL;
constexpr int64_t ps2_sifChainEnd=2147483648ULL;
constexpr int64_t ps2_sifEETagIRQ=2147483648ULL;
constexpr int64_t ps2_rpcPktServerID=32ULL;
constexpr int64_t ps2_rpcPktFuncNo=32ULL;
constexpr int64_t ps2_rpcPktServer=52ULL;
constexpr int64_t ps2_sifRegIOPCmdBufHW=2ULL;
constexpr int64_t ps2_sifRegIOPFlags=4ULL;
constexpr int64_t ps2_sifRegIOPCmdBuf=2147483648ULL;
constexpr int64_t ps2_sifRegEECmdBuf=2147483649ULL;
constexpr int64_t ps2_sifRegRPCUp=2147483650ULL;
constexpr int64_t ps2_sifIOPSIFUp=65536ULL;
constexpr int64_t ps2_sifIOPCmdUp=131072ULL;
constexpr int64_t ps2_sifIOPRebootDone=262144ULL;
constexpr int64_t ps2_sbusMSCOM=0ULL;
constexpr int64_t ps2_sbusSMCOM=16ULL;
constexpr int64_t ps2_sbusMSFLG=32ULL;
constexpr int64_t ps2_sbusSMFLG=48ULL;
constexpr int64_t ps2_sbusCTRL=64ULL;
constexpr int64_t ps2_sbusBD6=96ULL;
constexpr int64_t ps2_sbusSpan=128ULL;
constexpr int64_t ps2_sbusRegs=8ULL;
constexpr int64_t ps2_sbusEEBase=268497408ULL;
constexpr int64_t ps2_sbusIOPBase=486539264ULL;
constexpr int64_t ps2_sifEESIFReady=65536ULL;
constexpr int64_t ps2_vifNOP=0ULL;
constexpr int64_t ps2_vifSTCYCL=1ULL;
constexpr int64_t ps2_vifOFFSET=2ULL;
constexpr int64_t ps2_vifBASE=3ULL;
constexpr int64_t ps2_vifITOP=4ULL;
constexpr int64_t ps2_vifSTMOD=5ULL;
constexpr int64_t ps2_vifMSKPATH3=6ULL;
constexpr int64_t ps2_vifMARK=7ULL;
constexpr int64_t ps2_vifFLUSHE=16ULL;
constexpr int64_t ps2_vifFLUSH=17ULL;
constexpr int64_t ps2_vifFLUSHA=19ULL;
constexpr int64_t ps2_vifMSCAL=20ULL;
constexpr int64_t ps2_vifMSCALF=21ULL;
constexpr int64_t ps2_vifMSCNT=23ULL;
constexpr int64_t ps2_vifSTMASK=32ULL;
constexpr int64_t ps2_vifSTROW=48ULL;
constexpr int64_t ps2_vifSTCOL=49ULL;
constexpr int64_t ps2_vifMPG=74ULL;
constexpr int64_t ps2_vifDIRECT=80ULL;
constexpr int64_t ps2_vifDIRECTHL=81ULL;
constexpr int64_t ps2_vu0MemSize=4096ULL;
constexpr int64_t ps2_vu1MemSize=16384ULL;

#include "adapters.h"
// tools/cpu/r5900/cop0.go:41:1
uint64_t r5900_TLBEntry_pairSize(r5900_TLBEntry* e){
{
return cast<uint64_t>((cast<uint64_t>(cast<uint32_t>((e->PageMask | cast<uint32_t>(8191ULL)))) + cast<uint64_t>(1ULL)));
}
}
// tools/cpu/r5900/cop0.go:44:1
uint64_t r5900_TLBEntry_pageSize(r5900_TLBEntry* e){
{
return divi<uint64_t>(r5900_TLBEntry_pairSize(e),cast<uint64_t>(2ULL));
}
}
// tools/cpu/r5900/cop0.go:66:1
std::tuple<uint32_t,bool> r5900_CPU_Translate(r5900_CPU* c,uint64_t vaddr,bool store){
{
uint32_t v = cast<uint32_t>(vaddr);
{
if (((v >= cast<uint32_t>(1879048192ULL)) && (v < cast<uint32_t>(1879064576ULL)))){
return {v,true};
}
else if (((v >= cast<uint32_t>(2147483648ULL)) && (v < cast<uint32_t>(2684354560ULL)))){
return {cast<uint32_t>((v - cast<uint32_t>(2147483648ULL))),true};
}
else if (((v >= cast<uint32_t>(2684354560ULL)) && (v < cast<uint32_t>(3221225472ULL)))){
return {cast<uint32_t>((v - cast<uint32_t>(2684354560ULL))),true};
}
}
tmp1:;
return r5900_CPU_tlbTranslate(c,vaddr,store);
}
}
// tools/cpu/r5900/cop0.go:82:1
std::tuple<uint32_t,bool> r5900_CPU_tlbTranslate(r5900_CPU* c,uint64_t vaddr,bool store){
{
uint64_t asid = cast<uint64_t>((c->COP0[cast<int64_t>(10ULL)] & cast<uint64_t>(255ULL)));
{auto&& tmp2 = c->TLB;
for(int64_t tmp3=0;tmp3<len(tmp2);++tmp3){
auto i=tmp3;r5900_TLBEntry* e = (&c->TLB[i]);
uint64_t ps = r5900_TLBEntry_pageSize(e);
uint64_t pairMask = cast<uint64_t>(~(cast<uint64_t>((r5900_TLBEntry_pairSize(e) - cast<uint64_t>(1ULL)))));
if (((cast<uint64_t>((cast<uint64_t>((vaddr & pairMask)) & cast<uint64_t>(4294967295ULL)))) != (cast<uint64_t>((cast<uint64_t>((e->EntryHi & pairMask)) & cast<uint64_t>(4294967295ULL)))))) {
continue;
}
bool global = (cast<uint64_t>((cast<uint64_t>((e->EntryLo0 & e->EntryLo1)) & cast<uint64_t>(1ULL))) != cast<uint64_t>(0ULL));
if (((!global) && (cast<uint64_t>((e->EntryHi & cast<uint64_t>(255ULL))) != asid))) {
continue;
}
uint64_t lo = e->EntryLo0;
if ((cast<uint64_t>((vaddr & ps)) != cast<uint64_t>(0ULL))) {
lo = e->EntryLo1;
}
if ((cast<uint64_t>((lo & cast<uint64_t>(2ULL))) == cast<uint64_t>(0ULL))) {
r5900_CPU_tlbException(c,vaddr,store,false);
return {cast<uint32_t>(0ULL),false};
}
if ((store && (cast<uint64_t>((lo & cast<uint64_t>(4ULL))) == cast<uint64_t>(0ULL)))) {
r5900_CPU_setFaultAddress(c,vaddr);
r5900_CPU_setEntryHiVPN(c,vaddr);
r5900_CPU_Exception(c,cast<uint32_t>(1ULL));
return {cast<uint32_t>(0ULL),false};
}
uint64_t pfn = cast<uint64_t>(((shr<uint64_t>(lo,cast<int64_t>(6ULL))) & cast<uint64_t>(1048575ULL)));
return {cast<uint32_t>(cast<uint64_t>((((shl<uint64_t>(pfn,cast<int64_t>(12ULL))) & ~(cast<uint64_t>(cast<uint64_t>((ps - cast<uint64_t>(1ULL)))))) | (cast<uint64_t>((vaddr & (cast<uint64_t>((ps - cast<uint64_t>(1ULL)))))))))),true};
}}
r5900_CPU_tlbException(c,vaddr,store,true);
return {cast<uint32_t>(0ULL),false};
}
}
// tools/cpu/r5900/cop0.go:128:1
void r5900_CPU_tlbException(r5900_CPU* c,uint64_t vaddr,bool store,bool refill){
{
uint32_t code = cast<uint32_t>(2ULL);
if (store) {
code = cast<uint32_t>(3ULL);
}
r5900_CPU_setFaultAddress(c,vaddr);
r5900_CPU_setEntryHiVPN(c,vaddr);
r5900_CPU_exceptionAt(c,code,refill);
}
}
// tools/cpu/r5900/cop0.go:140:1
void r5900_CPU_setEntryHiVPN(r5900_CPU* c,uint64_t vaddr){
{
c->COP0[cast<int64_t>(10ULL)] = cast<uint64_t>(((cast<uint64_t>((c->COP0[cast<int64_t>(10ULL)] & cast<uint64_t>(255ULL)))) | ((vaddr & ~(cast<uint64_t>(8191ULL))))));
}
}
// tools/cpu/r5900/cop0.go:156:1
void r5900_CPU_setFaultAddress(r5900_CPU* c,uint64_t vaddr){
{
c->COP0[cast<int64_t>(8ULL)] = vaddr;
uint64_t badVPN2 = cast<uint64_t>(((shr<uint64_t>(vaddr,cast<int64_t>(13ULL))) & cast<uint64_t>(524287ULL)));
c->COP0[cast<int64_t>(4ULL)] = cast<uint64_t>((((c->COP0[cast<int64_t>(4ULL)] & ~(cast<uint64_t>(8388607ULL)))) | (shl<uint64_t>(badVPN2,cast<int64_t>(4ULL)))));
}
}
// tools/cpu/r5900/cop0.go:168:1
void r5900_CPU_tlbr(r5900_CPU* c){
{
uint64_t i = cast<uint64_t>((c->COP0[cast<int64_t>(0ULL)] & cast<uint64_t>(63ULL)));
if ((i >= cast<uint64_t>(48ULL))) {
r5900_CPU_Halt(c,std::string("tlbr: Index %d out of range at 0x%08X",37),i,cast<uint32_t>(c->curPC));
return ;
}
r5900_TLBEntry* e = (&c->TLB[i]);
c->COP0[cast<int64_t>(5ULL)] = cast<uint64_t>(e->PageMask);
c->COP0[cast<int64_t>(10ULL)] = e->EntryHi;
c->COP0[cast<int64_t>(2ULL)] = e->EntryLo0;
c->COP0[cast<int64_t>(3ULL)] = e->EntryLo1;
}
}
// tools/cpu/r5900/cop0.go:182:1
void r5900_CPU_tlbw(r5900_CPU* c,uint64_t i){
{
if ((i >= cast<uint64_t>(48ULL))) {
r5900_CPU_Halt(c,std::string("tlbw: index %d out of range at 0x%08X",37),i,cast<uint32_t>(c->curPC));
return ;
}
uint32_t pm = cast<uint32_t>((cast<uint32_t>(c->COP0[cast<int64_t>(5ULL)]) & cast<uint32_t>(33546240ULL)));
c->TLB[i] = r5900_TLBEntry{pm,(c->COP0[cast<int64_t>(10ULL)] & ~(cast<uint64_t>(pm))),c->COP0[cast<int64_t>(2ULL)],c->COP0[cast<int64_t>(3ULL)]};
}
}
// tools/cpu/r5900/cop0.go:200:1
void r5900_CPU_SetTLB(r5900_CPU* c,int64_t i,r5900_TLBEntry e){
{
if (((i >= cast<int64_t>(0ULL)) && (i < cast<int64_t>(48ULL)))) {
c->TLB[i] = e;
}
}
}
// tools/cpu/r5900/cop0.go:208:1
void r5900_CPU_tlbp(r5900_CPU* c){
{
uint64_t hi = c->COP0[cast<int64_t>(10ULL)];
uint64_t asid = cast<uint64_t>((hi & cast<uint64_t>(255ULL)));
{auto&& tmp4 = c->TLB;
for(int64_t tmp5=0;tmp5<len(tmp4);++tmp5){
auto i=tmp5;r5900_TLBEntry* e = (&c->TLB[i]);
uint64_t pairMask = cast<uint64_t>(~(cast<uint64_t>((r5900_TLBEntry_pairSize(e) - cast<uint64_t>(1ULL)))));
if (((cast<uint64_t>((cast<uint64_t>((hi & pairMask)) & cast<uint64_t>(4294967295ULL)))) != (cast<uint64_t>((cast<uint64_t>((e->EntryHi & pairMask)) & cast<uint64_t>(4294967295ULL)))))) {
continue;
}
bool global = (cast<uint64_t>((cast<uint64_t>((e->EntryLo0 & e->EntryLo1)) & cast<uint64_t>(1ULL))) != cast<uint64_t>(0ULL));
if (((!global) && (cast<uint64_t>((e->EntryHi & cast<uint64_t>(255ULL))) != asid))) {
continue;
}
c->COP0[cast<int64_t>(0ULL)] = cast<uint64_t>(i);
return ;
}}
c->COP0[cast<int64_t>(0ULL)] = cast<uint64_t>(2147483648ULL);
}
}
// tools/cpu/r5900/cop0.go:229:1
uint64_t r5900_CPU_random(r5900_CPU* c){
{
uint64_t wired = cast<uint64_t>((c->COP0[cast<int64_t>(6ULL)] & cast<uint64_t>(63ULL)));
if ((wired >= cast<uint64_t>(48ULL))) {
return cast<uint64_t>(47ULL);
}
uint64_t r = cast<uint64_t>((c->COP0[cast<int64_t>(1ULL)] & cast<uint64_t>(63ULL)));
if (((r <= wired) || (r >= cast<uint64_t>(48ULL)))) {
r = cast<uint64_t>(47ULL);
}
else {
r--;
}
c->COP0[cast<int64_t>(1ULL)] = r;
return r;
}
}
// tools/cpu/r5900/cop0.go:248:1
uint64_t r5900_CPU_readCop0(r5900_CPU* c,uint32_t i){
{
{
switch(i){
case cast<uint32_t>(1ULL):{
return c->COP0[cast<int64_t>(1ULL)];
break;}
case cast<uint32_t>(9ULL):case cast<uint32_t>(11ULL):case cast<uint32_t>(12ULL):case cast<uint32_t>(13ULL):case cast<uint32_t>(15ULL):case cast<uint32_t>(16ULL):case cast<uint32_t>(0ULL):case cast<uint32_t>(6ULL):case cast<uint32_t>(5ULL):{
return r5900_sext32(cast<uint32_t>(c->COP0[i]));
break;}
}}
return c->COP0[i];
}
}
// tools/cpu/r5900/cop0.go:261:1
void r5900_CPU_writeCop0(r5900_CPU* c,uint32_t i,uint64_t v){
{
{
switch(i){
case cast<uint32_t>(11ULL):{
c->COP0[cast<int64_t>(11ULL)] = cast<uint64_t>(cast<uint32_t>(v));
c->COP0[cast<int64_t>(13ULL)] &= ~(cast<uint64_t>(32768ULL));
return ;
break;}
case cast<uint32_t>(13ULL):{
c->COP0[cast<int64_t>(13ULL)] = cast<uint64_t>((((c->COP0[cast<int64_t>(13ULL)] & ~(cast<uint64_t>(768ULL)))) | (cast<uint64_t>((v & cast<uint64_t>(768ULL))))));
return ;
break;}
case cast<uint32_t>(9ULL):{
c->COP0[cast<int64_t>(9ULL)] = cast<uint64_t>(cast<uint32_t>(v));
c->countFrac = cast<uint64_t>(0ULL);
return ;
break;}
case cast<uint32_t>(15ULL):case cast<uint32_t>(1ULL):{
return ;
break;}
}}
c->COP0[i] = v;
}
}
// tools/cpu/r5900/cop0.go:282:1
void r5900_CPU_cop0(r5900_CPU* c,uint32_t w,uint32_t rs,uint32_t rt,uint32_t rd){
{
if ((cast<uint32_t>((w & cast<uint32_t>(33554432ULL))) != cast<uint32_t>(0ULL))) {
{
switch(cast<uint32_t>((w & cast<uint32_t>(63ULL)))){
case cast<uint32_t>(1ULL):{
r5900_CPU_tlbr(c);
break;}
case cast<uint32_t>(2ULL):{
r5900_CPU_tlbw(c,cast<uint64_t>((c->COP0[cast<int64_t>(0ULL)] & cast<uint64_t>(63ULL))));
break;}
case cast<uint32_t>(6ULL):{
r5900_CPU_tlbw(c,r5900_CPU_random(c));
break;}
case cast<uint32_t>(8ULL):{
r5900_CPU_tlbp(c);
break;}
case cast<uint32_t>(24ULL):{
r5900_CPU_eret(c);
break;}
case cast<uint32_t>(56ULL):{
c->COP0[cast<int64_t>(12ULL)] |= cast<uint64_t>(65536ULL);
break;}
case cast<uint32_t>(57ULL):{
c->COP0[cast<int64_t>(12ULL)] &= ~(cast<uint64_t>(65536ULL));
break;}
default:{
break;}
}}
return ;
}
{
switch(rs){
case cast<uint32_t>(0ULL):{
r5900_CPU_set(c,rt,r5900_sext32(cast<uint32_t>(r5900_CPU_readCop0(c,rd))));
break;}
case cast<uint32_t>(4ULL):{
r5900_CPU_writeCop0(c,rd,r5900_sext32(cast<uint32_t>(c->R[rt].Lo)));
break;}
default:{
r5900_CPU_Exception(c,cast<uint32_t>(10ULL));
break;}
}}
}
}
// tools/cpu/r5900/cop0.go:318:1
std::string r5900_CPU_DumpTLB(r5900_CPU* c){
{
std::string s = std::string("",0);
{auto&& tmp6 = c->TLB;
for(int64_t tmp7=0;tmp7<len(tmp6);++tmp7){
auto i=tmp7;r5900_TLBEntry* e = (&c->TLB[i]);
if (((cast<uint64_t>((e->EntryLo0 & cast<uint64_t>(2ULL))) == cast<uint64_t>(0ULL)) && (cast<uint64_t>((e->EntryLo1 & cast<uint64_t>(2ULL))) == cast<uint64_t>(0ULL)))) {
continue;
}
s += go_fmt_Sprintf(std::string("%2d: hi=%016X lo0=%016X lo1=%016X mask=%08X size=%d\012",52),i,e->EntryHi,e->EntryLo0,e->EntryLo1,e->PageMask,r5900_TLBEntry_pageSize(e));
}}
if ((s == std::string("",0))) {
return std::string("TLB: no valid entries\012",22);
}
return s;
}
}
// tools/cpu/r5900/cpu.go:225:1
void r5900_CPU_Reset(r5900_CPU* c){
{
c->R = std::array<r5900_Quad,32>{};
auto tmp8 = std::make_tuple(cast<uint64_t>(0ULL),cast<uint64_t>(0ULL),cast<uint64_t>(0ULL),cast<uint64_t>(0ULL),cast<uint32_t>(0ULL));
c->HI = std::get<0>(tmp8);
c->LO = std::get<1>(tmp8);
c->HI1 = std::get<2>(tmp8);
c->LO1 = std::get<3>(tmp8);
c->SA = std::get<4>(tmp8);
auto tmp9 = std::make_tuple(cast<uint64_t>(18446744072631615488ULL),cast<uint64_t>(18446744072631615492ULL));
c->PC = std::get<0>(tmp9);
c->nextPC = std::get<1>(tmp9);
c->COP0 = std::array<uint64_t,32>{};
c->COP0[cast<int64_t>(12ULL)] = cast<uint64_t>(4194308ULL);
c->COP0[cast<int64_t>(15ULL)] = cast<uint64_t>(11808ULL);
c->COP0[cast<int64_t>(16ULL)] = cast<uint64_t>(1088ULL);
c->COP0[cast<int64_t>(1ULL)] = cast<uint64_t>(47ULL);
c->TLB = std::array<r5900_TLBEntry,48>{};
c->FPR = std::array<uint32_t,32>{};
auto tmp10 = std::make_tuple(cast<uint32_t>(0ULL),cast<uint32_t>(0ULL));
c->ACC = std::get<0>(tmp10);
c->FCR31 = std::get<1>(tmp10);
c->LLBit = false;
auto tmp11 = std::make_tuple(false,false);
c->delaySlot = std::get<0>(tmp11);
c->pendingDelay = std::get<1>(tmp11);
auto tmp12 = std::make_tuple(false,std::string("",0));
c->Halted = std::get<0>(tmp12);
c->HaltReason = std::get<1>(tmp12);
auto tmp13 = std::make_tuple(cast<uint64_t>(0ULL),cast<uint64_t>(0ULL));
c->Steps = std::get<0>(tmp13);
c->countFrac = std::get<1>(tmp13);
}
}
// tools/cpu/r5900/cpu.go:253:1
void r5900_CPU_SetPC(r5900_CPU* c,uint64_t pc){
{
auto tmp14 = std::make_tuple(pc,cast<uint64_t>((pc + cast<uint64_t>(4ULL))));
c->PC = std::get<0>(tmp14);
c->nextPC = std::get<1>(tmp14);
c->pendingDelay = false;
}
}
// tools/cpu/r5900/cpu.go:260:1
void r5900_CPU_SetReg(r5900_CPU* c,uint32_t i,uint64_t v){
{
r5900_CPU_set(c,i,v);
}
}
// tools/cpu/r5900/cpu.go:264:1
uint64_t r5900_CPU_Reg(r5900_CPU* c,uint32_t i){
{
return c->R[i].Lo;
}
}
// tools/cpu/r5900/cpu.go:267:1
void r5900_CPU_SetQuad(r5900_CPU* c,uint32_t i,r5900_Quad q){
{
if ((i != cast<uint32_t>(0ULL))) {
c->R[i] = q;
}
}
}
// tools/cpu/r5900/cpu.go:274:1
r5900_Quad r5900_CPU_Quad(r5900_CPU* c,uint32_t i){
{
return c->R[i];
}
}
// tools/cpu/r5900/cpu.go:278:1
uint64_t r5900_CPU_CurPC(r5900_CPU* c){
{
return c->curPC;
}
}
// tools/cpu/r5900/cpu.go:284:1
uint64_t r5900_CPU_NextPC(r5900_CPU* c){
{
return c->nextPC;
}
}
// tools/cpu/r5900/cpu.go:285:1
bool r5900_CPU_InDelaySlot(r5900_CPU* c){
{
return c->delaySlot;
}
}
// tools/cpu/r5900/cpu.go:286:1
bool r5900_CPU_PendingDelay(r5900_CPU* c){
{
return c->pendingDelay;
}
}
// tools/cpu/r5900/cpu.go:287:1
uint64_t r5900_CPU_BranchAddr(r5900_CPU* c){
{
return c->branchAddr;
}
}
// tools/cpu/r5900/cpu.go:292:1
void r5900_CPU_set(r5900_CPU* c,uint32_t i,uint64_t v){
{
if ((i != cast<uint32_t>(0ULL))) {
c->R[i].Lo = v;
}
}
}
// tools/cpu/r5900/cpu.go:299:1
void r5900_CPU_setQ(r5900_CPU* c,uint32_t i,uint64_t lo,uint64_t hi){
{
if ((i != cast<uint32_t>(0ULL))) {
auto tmp15 = std::make_tuple(lo,hi);
c->R[i].Lo = std::get<0>(tmp15);
c->R[i].Hi = std::get<1>(tmp15);
}
}
}
// tools/cpu/r5900/cpu.go:308:1
uint64_t r5900_sext32(uint32_t v){
{
return cast<uint64_t>(cast<int64_t>(cast<int32_t>(v)));
}
}
// tools/cpu/r5900/cpu.go:312:1
uint32_t r5900_CPU_read8(r5900_CPU* c,uint32_t a){
{
return cast<uint32_t>(ps2_Machine_Read(c->bus,a));
}
}
// tools/cpu/r5900/cpu.go:313:1
void r5900_CPU_write8(r5900_CPU* c,uint32_t a,uint32_t v){
{
ps2_Machine_Write(c->bus,a,cast<uint8_t>(v));
}
}
// tools/cpu/r5900/cpu.go:315:1
uint32_t r5900_CPU_read16(r5900_CPU* c,uint32_t a){
{
return cast<uint32_t>((cast<uint32_t>(ps2_Machine_Read(c->bus,a)) | shl<uint32_t>(cast<uint32_t>(ps2_Machine_Read(c->bus,cast<uint32_t>((a + cast<uint32_t>(1ULL))))),cast<int64_t>(8ULL))));
}
}
// tools/cpu/r5900/cpu.go:318:1
void r5900_CPU_write16(r5900_CPU* c,uint32_t a,uint32_t v){
{
ps2_Machine_Write(c->bus,a,cast<uint8_t>(v));
ps2_Machine_Write(c->bus,cast<uint32_t>((a + cast<uint32_t>(1ULL))),cast<uint8_t>(shr<uint32_t>(v,cast<int64_t>(8ULL))));
}
}
// tools/cpu/r5900/cpu.go:323:1
uint32_t r5900_CPU_read32(r5900_CPU* c,uint32_t a){
{
return ps2_Machine_Read32(c->bus,a);
}
}
// tools/cpu/r5900/cpu.go:324:1
void r5900_CPU_write32(r5900_CPU* c,uint32_t a,uint32_t v){
{
ps2_Machine_Write32(c->bus,a,v);
}
}
// tools/cpu/r5900/cpu.go:327:1
uint64_t r5900_CPU_read64(r5900_CPU* c,uint32_t a){
{
return cast<uint64_t>((cast<uint64_t>(ps2_Machine_Read32(c->bus,a)) | shl<uint64_t>(cast<uint64_t>(ps2_Machine_Read32(c->bus,cast<uint32_t>((a + cast<uint32_t>(4ULL))))),cast<int64_t>(32ULL))));
}
}
// tools/cpu/r5900/cpu.go:330:1
void r5900_CPU_write64(r5900_CPU* c,uint32_t a,uint64_t v){
{
ps2_Machine_Write32(c->bus,a,cast<uint32_t>(v));
ps2_Machine_Write32(c->bus,cast<uint32_t>((a + cast<uint32_t>(4ULL))),cast<uint32_t>(shr<uint64_t>(v,cast<int64_t>(32ULL))));
}
}
// tools/cpu/r5900/cpu.go:335:1
r5900_Quad r5900_CPU_read128(r5900_CPU* c,uint32_t a){
{
return r5900_Quad{r5900_CPU_read64(c,a),r5900_CPU_read64(c,cast<uint32_t>((a + cast<uint32_t>(8ULL))))};
}
}
// tools/cpu/r5900/cpu.go:338:1
void r5900_CPU_write128(r5900_CPU* c,uint32_t a,r5900_Quad q){
{
r5900_CPU_write64(c,a,q.Lo);
r5900_CPU_write64(c,cast<uint32_t>((a + cast<uint32_t>(8ULL))),q.Hi);
}
}
// tools/cpu/r5900/cpu.go:348:1
void r5900_CPU_Exception(r5900_CPU* c,uint32_t code){
{
r5900_CPU_exceptionAt(c,code,false);
}
}
// tools/cpu/r5900/cpu.go:352:1
void r5900_CPU_exceptionAt(r5900_CPU* c,uint32_t code,bool tlbRefill){
{
uint64_t sr = c->COP0[cast<int64_t>(12ULL)];
if ((cast<uint64_t>((sr & cast<uint64_t>(2ULL))) == cast<uint64_t>(0ULL))) {
uint64_t epc = c->curPC;
uint64_t cause = (c->COP0[cast<int64_t>(13ULL)] & ~(cast<uint64_t>(18446744069414584320ULL)));
if (c->delaySlot) {
epc = c->branchAddr;
cause |= cast<uint64_t>(2147483648ULL);
}
else {
cause &= ~(cast<uint64_t>(2147483648ULL));
}
c->COP0[cast<int64_t>(14ULL)] = epc;
c->COP0[cast<int64_t>(13ULL)] = cast<uint64_t>((((cause & ~(cast<uint64_t>(124ULL)))) | cast<uint64_t>(shl<uint32_t>(code,cast<int64_t>(2ULL)))));
}
else {
c->COP0[cast<int64_t>(13ULL)] = cast<uint64_t>((((c->COP0[cast<int64_t>(13ULL)] & ~(cast<uint64_t>(124ULL)))) | cast<uint64_t>(shl<uint32_t>(code,cast<int64_t>(2ULL)))));
}
c->COP0[cast<int64_t>(13ULL)] &= ~(cast<uint64_t>(805306368ULL));
uint64_t base = cast<uint64_t>(2147483648ULL);
if ((cast<uint64_t>((sr & cast<uint64_t>(4194304ULL))) != cast<uint64_t>(0ULL))) {
base = cast<uint64_t>(3217031680ULL);
}
uint64_t offset = cast<uint64_t>(384ULL);
if ((tlbRefill && (cast<uint64_t>((sr & cast<uint64_t>(2ULL))) == cast<uint64_t>(0ULL)))) {
offset = cast<uint64_t>(0ULL);
}
c->COP0[cast<int64_t>(12ULL)] = cast<uint64_t>((sr | cast<uint64_t>(2ULL)));
c->LLBit = false;
r5900_CPU_SetPC(c,r5900_sext64(cast<uint64_t>((base + offset))));
c->delaySlot = false;
}
}
// tools/cpu/r5900/cpu.go:395:1
uint64_t r5900_sext64(uint64_t v){
{
return cast<uint64_t>(cast<int64_t>(cast<int32_t>(cast<uint32_t>(v))));
}
}
// tools/cpu/r5900/cpu.go:399:1
void r5900_CPU_eret(r5900_CPU* c){
{
uint64_t sr = c->COP0[cast<int64_t>(12ULL)];
if ((cast<uint64_t>((sr & cast<uint64_t>(4ULL))) != cast<uint64_t>(0ULL))) {
r5900_CPU_SetPC(c,c->COP0[cast<int64_t>(30ULL)]);
c->COP0[cast<int64_t>(12ULL)] = (sr & ~(cast<uint64_t>(4ULL)));
}
else {
r5900_CPU_SetPC(c,c->COP0[cast<int64_t>(14ULL)]);
c->COP0[cast<int64_t>(12ULL)] = (sr & ~(cast<uint64_t>(2ULL)));
}
c->LLBit = false;
c->pendingDelay = false;
}
}
// tools/cpu/r5900/cpu.go:416:1
void r5900_CPU_coprocessorUnusable(r5900_CPU* c,uint32_t unit){
{
r5900_CPU_Exception(c,cast<uint32_t>(11ULL));
c->COP0[cast<int64_t>(13ULL)] = cast<uint64_t>((((c->COP0[cast<int64_t>(13ULL)] & ~(cast<uint64_t>(805306368ULL)))) | shl<uint64_t>(cast<uint64_t>(unit),cast<int64_t>(28ULL))));
}
}
// tools/cpu/r5900/cpu.go:422:1
void r5900_CPU_addrError(r5900_CPU* c,uint32_t code,uint64_t vaddr){
{
r5900_CPU_setFaultAddress(c,vaddr);
r5900_CPU_Exception(c,code);
}
}
// tools/cpu/r5900/cpu.go:436:1
bool r5900_CPU_Interrupt(r5900_CPU* c,bool int0,bool int1){
{
if (int0) {
c->COP0[cast<int64_t>(13ULL)] |= cast<uint64_t>(1024ULL);
}
else {
c->COP0[cast<int64_t>(13ULL)] &= ~(cast<uint64_t>(1024ULL));
}
if (int1) {
c->COP0[cast<int64_t>(13ULL)] |= cast<uint64_t>(2048ULL);
}
else {
c->COP0[cast<int64_t>(13ULL)] &= ~(cast<uint64_t>(2048ULL));
}
return r5900_CPU_checkInterrupt(c);
}
}
// tools/cpu/r5900/cpu.go:455:1
bool r5900_CPU_checkInterrupt(r5900_CPU* c){
{
uint64_t sr = c->COP0[cast<int64_t>(12ULL)];
if ((((cast<uint64_t>((sr & cast<uint64_t>(1ULL))) == cast<uint64_t>(0ULL)) || (cast<uint64_t>((sr & cast<uint64_t>(65536ULL))) == cast<uint64_t>(0ULL))) || (cast<uint64_t>((sr & cast<uint64_t>(6ULL))) != cast<uint64_t>(0ULL)))) {
return false;
}
if ((cast<uint64_t>((cast<uint64_t>((c->COP0[cast<int64_t>(13ULL)] & sr)) & cast<uint64_t>(65280ULL))) == cast<uint64_t>(0ULL))) {
return false;
}
if (c->pendingDelay) {
return false;
}
c->curPC = c->PC;
c->delaySlot = false;
r5900_CPU_Exception(c,cast<uint32_t>(0ULL));
return true;
}
}
// tools/cpu/r5900/cpu.go:479:1
void r5900_CPU_tickCount(r5900_CPU* c){
{
c->countFrac++;
if ((cast<uint64_t>((c->countFrac & cast<uint64_t>(1ULL))) != cast<uint64_t>(0ULL))) {
return ;
}
uint32_t count = cast<uint32_t>((cast<uint32_t>(c->COP0[cast<int64_t>(9ULL)]) + cast<uint32_t>(1ULL)));
c->COP0[cast<int64_t>(9ULL)] = cast<uint64_t>(count);
if ((count == cast<uint32_t>(c->COP0[cast<int64_t>(11ULL)]))) {
c->COP0[cast<int64_t>(13ULL)] |= cast<uint64_t>(32768ULL);
}
}
}
// tools/cpu/r5900/cpu.go:504:1
void r5900_CPU_SkipInstructions(r5900_CPU* c,uint64_t n){
{
if ((n == cast<uint64_t>(0ULL))) {
return ;
}
uint64_t f0 = c->countFrac;
c->countFrac = cast<uint64_t>((f0 + n));
uint64_t inc = cast<uint64_t>((divi<uint64_t>((cast<uint64_t>((f0 + n))),cast<uint64_t>(2ULL)) - divi<uint64_t>(f0,cast<uint64_t>(2ULL))));
if ((inc > cast<uint64_t>(0ULL))) {
uint32_t count = cast<uint32_t>(c->COP0[cast<int64_t>(9ULL)]);
uint32_t comp = cast<uint32_t>(c->COP0[cast<int64_t>(11ULL)]);
if ((inc >= cast<uint64_t>(4294967296ULL))) {
c->COP0[cast<int64_t>(13ULL)] |= cast<uint64_t>(32768ULL);
}
else {
uint64_t d = cast<uint64_t>(cast<uint32_t>((comp - count)));
if (((d != cast<uint64_t>(0ULL)) && (d <= inc))) {
c->COP0[cast<int64_t>(13ULL)] |= cast<uint64_t>(32768ULL);
}
}
c->COP0[cast<int64_t>(9ULL)] = cast<uint64_t>(cast<uint32_t>((count + cast<uint32_t>(inc))));
}
c->Steps += n;
}
}
// tools/cpu/r5900/cpu.go:529:1
uint32_t r5900_CPU_Count(r5900_CPU* c){
{
return cast<uint32_t>(c->COP0[cast<int64_t>(9ULL)]);
}
}
// tools/cpu/r5900/cpu.go:530:1
uint32_t r5900_CPU_Compare(r5900_CPU* c){
{
return cast<uint32_t>(c->COP0[cast<int64_t>(11ULL)]);
}
}
// tools/cpu/r5900/cpu.go:534:1
uint64_t r5900_CPU_CountFrac(r5900_CPU* c){
{
return c->countFrac;
}
}
// tools/cpu/r5900/cpu.go:540:1
bool r5900_CPU_TimerIRQDeliverable(r5900_CPU* c){
{
uint64_t sr = c->COP0[cast<int64_t>(12ULL)];
return ((((cast<uint64_t>((sr & cast<uint64_t>(1ULL))) != cast<uint64_t>(0ULL)) && (cast<uint64_t>((sr & cast<uint64_t>(65536ULL))) != cast<uint64_t>(0ULL))) && (cast<uint64_t>((sr & cast<uint64_t>(6ULL))) == cast<uint64_t>(0ULL))) && (cast<uint64_t>((sr & cast<uint64_t>(32768ULL))) != cast<uint64_t>(0ULL)));
}
}
// tools/cpu/r5900/cpu.go:549:1
bool r5900_CPU_InterruptDeliverable(r5900_CPU* c){
{
uint64_t sr = c->COP0[cast<int64_t>(12ULL)];
if ((((cast<uint64_t>((sr & cast<uint64_t>(1ULL))) == cast<uint64_t>(0ULL)) || (cast<uint64_t>((sr & cast<uint64_t>(65536ULL))) == cast<uint64_t>(0ULL))) || (cast<uint64_t>((sr & cast<uint64_t>(6ULL))) != cast<uint64_t>(0ULL)))) {
return false;
}
return (cast<uint64_t>((cast<uint64_t>((c->COP0[cast<int64_t>(13ULL)] & sr)) & cast<uint64_t>(65280ULL))) != cast<uint64_t>(0ULL));
}
}
// tools/cpu/r5900/decode.go:18:1
std::string r5900_reg(uint32_t i){
{
return (std::string("$",1) + r5900_regName[cast<uint32_t>((i & cast<uint32_t>(31ULL)))]);
}
}
// tools/cpu/r5900/decode.go:20:1
std::string r5900_freg(uint32_t i){
{
return go_fmt_Sprintf(std::string("$f%d",4),cast<uint32_t>((i & cast<uint32_t>(31ULL))));
}
}
// tools/cpu/r5900/decode.go:22:1
std::string r5900_vfreg(uint32_t i){
{
return go_fmt_Sprintf(std::string("$vf%d",5),cast<uint32_t>((i & cast<uint32_t>(31ULL))));
}
}
// tools/cpu/r5900/decode.go:32:1
std::string r5900_cop0Reg(uint32_t i){
{
{
std::string n = r5900_cop0Name[cast<uint32_t>((i & cast<uint32_t>(31ULL)))];
if ((n != std::string("",0))) {
return n;
}
}
return go_fmt_Sprintf(std::string("$%d",3),cast<uint32_t>((i & cast<uint32_t>(31ULL))));
}
}
// tools/cpu/r5900/decode.go:41:1
r5900_Inst r5900_Decode(Slice<uint8_t> code,uint32_t addr){
{
if ((len(code) < cast<int64_t>(4ULL))) {
return r5900_Inst{addr,cast<int64_t>(0ULL),std::string(".word",5),std::string(".word <truncated>",17),cast<r5900_Flow>(7ULL),{},{},{},{}};
}
return r5900_DecodeWord(le_Uint32(code),addr);
}
}
// tools/cpu/r5900/decode.go:50:1
r5900_Inst r5900_DecodeWord(uint32_t w,uint32_t addr){
{
r5900_Inst in = r5900_Inst{addr,cast<int64_t>(4ULL),{},{},cast<r5900_Flow>(0ULL),{},{},{},{}};
uint32_t op = shr<uint32_t>(w,cast<int64_t>(26ULL));
uint32_t rs = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(21ULL))) & cast<uint32_t>(31ULL)));
uint32_t rt = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(31ULL)));
uint32_t rd = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(11ULL))) & cast<uint32_t>(31ULL)));
uint32_t shamt = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(6ULL))) & cast<uint32_t>(31ULL)));
uint32_t imm = cast<uint32_t>((w & cast<uint32_t>(65535ULL)));
int32_t simm = cast<int32_t>(cast<int16_t>(imm));
uint32_t branchT = cast<uint32_t>((cast<uint32_t>((addr + cast<uint32_t>(4ULL))) + cast<uint32_t>((cast<uint32_t>(simm) * cast<uint32_t>(4ULL)))));
uint32_t jumpT = cast<uint32_t>((cast<uint32_t>(((cast<uint32_t>((addr + cast<uint32_t>(4ULL)))) & cast<uint32_t>(4026531840ULL))) | shl<uint32_t>((cast<uint32_t>((w & cast<uint32_t>(67108863ULL)))),cast<int64_t>(2ULL))));
auto set = [&](std::string mnem,std::string text)->void{
auto tmp16 = std::make_tuple(mnem,text);
in.Mnem = std::get<0>(tmp16);
in.Text = std::get<1>(tmp16);
}
;
auto branch = [&](std::string mnem,std::string text,bool likely)->void{
auto tmp17 = std::make_tuple(cast<r5900_Flow>(1ULL),branchT,true,true);
in.Flow = std::get<0>(tmp17);
in.Target = std::get<1>(tmp17);
in.HasTarget = std::get<2>(tmp17);
in.HasDelay = std::get<3>(tmp17);
in.Annul = likely;
set(mnem,text);
}
;
auto mem = [&](std::string m,std::string r)->void{
set(m,go_fmt_Sprintf(std::string("%s %s, %d(%s)",13),m,r,simm,r5900_reg(rs)));
}
;
{
switch(op){
case cast<uint32_t>(0ULL):{
return r5900_decodeSpecial(in,w,rs,rt,rd,shamt);
break;}
case cast<uint32_t>(1ULL):{
return r5900_decodeRegimm(in,w,rs,rt,branchT,simm);
break;}
case cast<uint32_t>(28ULL):{
return r5900_decodeMMI(in,w,rs,rt,rd,shamt);
break;}
case cast<uint32_t>(2ULL):{
auto tmp18 = std::make_tuple(cast<r5900_Flow>(2ULL),jumpT,true,true);
in.Flow = std::get<0>(tmp18);
in.Target = std::get<1>(tmp18);
in.HasTarget = std::get<2>(tmp18);
in.HasDelay = std::get<3>(tmp18);
set(std::string("j",1),go_fmt_Sprintf(std::string("j $%08X",7),jumpT));
break;}
case cast<uint32_t>(3ULL):{
auto tmp19 = std::make_tuple(cast<r5900_Flow>(3ULL),jumpT,true,true);
in.Flow = std::get<0>(tmp19);
in.Target = std::get<1>(tmp19);
in.HasTarget = std::get<2>(tmp19);
in.HasDelay = std::get<3>(tmp19);
set(std::string("jal",3),go_fmt_Sprintf(std::string("jal $%08X",9),jumpT));
break;}
case cast<uint32_t>(4ULL):{
if (((rs == cast<uint32_t>(0ULL)) && (rt == cast<uint32_t>(0ULL)))) {
auto tmp20 = std::make_tuple(cast<r5900_Flow>(2ULL),branchT,true,true);
in.Flow = std::get<0>(tmp20);
in.Target = std::get<1>(tmp20);
in.HasTarget = std::get<2>(tmp20);
in.HasDelay = std::get<3>(tmp20);
set(std::string("b",1),go_fmt_Sprintf(std::string("b $%08X",7),branchT));
}
else {
branch(std::string("beq",3),go_fmt_Sprintf(std::string("beq %s, %s, $%08X",17),r5900_reg(rs),r5900_reg(rt),branchT),false);
}
break;}
case cast<uint32_t>(5ULL):{
branch(std::string("bne",3),go_fmt_Sprintf(std::string("bne %s, %s, $%08X",17),r5900_reg(rs),r5900_reg(rt),branchT),false);
break;}
case cast<uint32_t>(6ULL):{
branch(std::string("blez",4),go_fmt_Sprintf(std::string("blez %s, $%08X",14),r5900_reg(rs),branchT),false);
break;}
case cast<uint32_t>(7ULL):{
branch(std::string("bgtz",4),go_fmt_Sprintf(std::string("bgtz %s, $%08X",14),r5900_reg(rs),branchT),false);
break;}
case cast<uint32_t>(20ULL):{
branch(std::string("beql",4),go_fmt_Sprintf(std::string("beql %s, %s, $%08X",18),r5900_reg(rs),r5900_reg(rt),branchT),true);
break;}
case cast<uint32_t>(21ULL):{
branch(std::string("bnel",4),go_fmt_Sprintf(std::string("bnel %s, %s, $%08X",18),r5900_reg(rs),r5900_reg(rt),branchT),true);
break;}
case cast<uint32_t>(22ULL):{
branch(std::string("blezl",5),go_fmt_Sprintf(std::string("blezl %s, $%08X",15),r5900_reg(rs),branchT),true);
break;}
case cast<uint32_t>(23ULL):{
branch(std::string("bgtzl",5),go_fmt_Sprintf(std::string("bgtzl %s, $%08X",15),r5900_reg(rs),branchT),true);
break;}
case cast<uint32_t>(8ULL):{
set(std::string("addi",4),go_fmt_Sprintf(std::string("addi %s, %s, %d",15),r5900_reg(rt),r5900_reg(rs),simm));
break;}
case cast<uint32_t>(9ULL):{
set(std::string("addiu",5),go_fmt_Sprintf(std::string("addiu %s, %s, %d",16),r5900_reg(rt),r5900_reg(rs),simm));
break;}
case cast<uint32_t>(10ULL):{
set(std::string("slti",4),go_fmt_Sprintf(std::string("slti %s, %s, %d",15),r5900_reg(rt),r5900_reg(rs),simm));
break;}
case cast<uint32_t>(11ULL):{
set(std::string("sltiu",5),go_fmt_Sprintf(std::string("sltiu %s, %s, %d",16),r5900_reg(rt),r5900_reg(rs),simm));
break;}
case cast<uint32_t>(12ULL):{
set(std::string("andi",4),go_fmt_Sprintf(std::string("andi %s, %s, 0x%X",17),r5900_reg(rt),r5900_reg(rs),imm));
break;}
case cast<uint32_t>(13ULL):{
set(std::string("ori",3),go_fmt_Sprintf(std::string("ori %s, %s, 0x%X",16),r5900_reg(rt),r5900_reg(rs),imm));
break;}
case cast<uint32_t>(14ULL):{
set(std::string("xori",4),go_fmt_Sprintf(std::string("xori %s, %s, 0x%X",17),r5900_reg(rt),r5900_reg(rs),imm));
break;}
case cast<uint32_t>(15ULL):{
set(std::string("lui",3),go_fmt_Sprintf(std::string("lui %s, 0x%X",12),r5900_reg(rt),imm));
break;}
case cast<uint32_t>(24ULL):{
set(std::string("daddi",5),go_fmt_Sprintf(std::string("daddi %s, %s, %d",16),r5900_reg(rt),r5900_reg(rs),simm));
break;}
case cast<uint32_t>(25ULL):{
set(std::string("daddiu",6),go_fmt_Sprintf(std::string("daddiu %s, %s, %d",17),r5900_reg(rt),r5900_reg(rs),simm));
break;}
case cast<uint32_t>(16ULL):{
return r5900_decodeCop0(in,w,rs,rt,rd);
break;}
case cast<uint32_t>(17ULL):{
return r5900_decodeCop1(in,w,rs,rt,rd,shamt,branchT);
break;}
case cast<uint32_t>(18ULL):{
return r5900_decodeCop2(in,w,rs,rt,rd,branchT);
break;}
case cast<uint32_t>(26ULL):{
mem(std::string("ldl",3),r5900_reg(rt));
break;}
case cast<uint32_t>(27ULL):{
mem(std::string("ldr",3),r5900_reg(rt));
break;}
case cast<uint32_t>(30ULL):{
mem(std::string("lq",2),r5900_reg(rt));
break;}
case cast<uint32_t>(31ULL):{
mem(std::string("sq",2),r5900_reg(rt));
break;}
case cast<uint32_t>(32ULL):case cast<uint32_t>(33ULL):case cast<uint32_t>(34ULL):case cast<uint32_t>(35ULL):case cast<uint32_t>(36ULL):case cast<uint32_t>(37ULL):case cast<uint32_t>(38ULL):case cast<uint32_t>(39ULL):{
std::string m = [](){std::array<std::string,40> v{};v[32]=std::string("lb",2);v[33]=std::string("lh",2);v[34]=std::string("lwl",3);v[35]=std::string("lw",2);v[36]=std::string("lbu",3);v[37]=std::string("lhu",3);v[38]=std::string("lwr",3);v[39]=std::string("lwu",3);return v;}()[op];
mem(m,r5900_reg(rt));
break;}
case cast<uint32_t>(40ULL):case cast<uint32_t>(41ULL):case cast<uint32_t>(42ULL):case cast<uint32_t>(43ULL):case cast<uint32_t>(44ULL):case cast<uint32_t>(45ULL):case cast<uint32_t>(46ULL):{
std::string m = [](){std::array<std::string,47> v{};v[40]=std::string("sb",2);v[41]=std::string("sh",2);v[42]=std::string("swl",3);v[43]=std::string("sw",2);v[44]=std::string("sdl",3);v[45]=std::string("sdr",3);v[46]=std::string("swr",3);return v;}()[op];
mem(m,r5900_reg(rt));
break;}
case cast<uint32_t>(47ULL):{
set(std::string("cache",5),go_fmt_Sprintf(std::string("cache 0x%X, %d(%s)",18),rt,simm,r5900_reg(rs)));
break;}
case cast<uint32_t>(51ULL):{
set(std::string("pref",4),go_fmt_Sprintf(std::string("pref 0x%X, %d(%s)",17),rt,simm,r5900_reg(rs)));
break;}
case cast<uint32_t>(48ULL):{
mem(std::string("ll",2),r5900_reg(rt));
break;}
case cast<uint32_t>(52ULL):{
mem(std::string("lld",3),r5900_reg(rt));
break;}
case cast<uint32_t>(55ULL):{
mem(std::string("ld",2),r5900_reg(rt));
break;}
case cast<uint32_t>(56ULL):{
mem(std::string("sc",2),r5900_reg(rt));
break;}
case cast<uint32_t>(60ULL):{
mem(std::string("scd",3),r5900_reg(rt));
break;}
case cast<uint32_t>(63ULL):{
mem(std::string("sd",2),r5900_reg(rt));
break;}
case cast<uint32_t>(49ULL):{
mem(std::string("lwc1",4),r5900_freg(rt));
break;}
case cast<uint32_t>(57ULL):{
mem(std::string("swc1",4),r5900_freg(rt));
break;}
case cast<uint32_t>(54ULL):{
mem(std::string("lqc2",4),r5900_vfreg(rt));
break;}
case cast<uint32_t>(62ULL):{
mem(std::string("sqc2",4),r5900_vfreg(rt));
break;}
default:{
return r5900_word(in,w);
break;}
}}
return in;
}
}
// tools/cpu/r5900/decode.go:199:1
r5900_Inst r5900_decodeSpecial(r5900_Inst in,uint32_t w,uint32_t rs,uint32_t rt,uint32_t rd,uint32_t shamt){
{
uint32_t funct = cast<uint32_t>((w & cast<uint32_t>(63ULL)));
auto set = [&](std::string mnem,std::string text)->r5900_Inst{
auto tmp21 = std::make_tuple(mnem,text);
in.Mnem = std::get<0>(tmp21);
in.Text = std::get<1>(tmp21);
return in;
}
;
auto r3 = [&](std::string m)->r5900_Inst{
return set(m,go_fmt_Sprintf(std::string("%s %s, %s, %s",13),m,r5900_reg(rd),r5900_reg(rs),r5900_reg(rt)));
}
;
auto sh = [&](std::string m)->r5900_Inst{
return set(m,go_fmt_Sprintf(std::string("%s %s, %s, %d",13),m,r5900_reg(rd),r5900_reg(rt),shamt));
}
;
auto shv = [&](std::string m)->r5900_Inst{
return set(m,go_fmt_Sprintf(std::string("%s %s, %s, %s",13),m,r5900_reg(rd),r5900_reg(rt),r5900_reg(rs)));
}
;
auto r2 = [&](std::string m)->r5900_Inst{
return set(m,go_fmt_Sprintf(std::string("%s %s, %s",9),m,r5900_reg(rs),r5900_reg(rt)));
}
;
{
switch(funct){
case cast<uint32_t>(0ULL):{
if ((w == cast<uint32_t>(0ULL))) {
return set(std::string("nop",3),std::string("nop",3));
}
return sh(std::string("sll",3));
break;}
case cast<uint32_t>(2ULL):{
return sh(std::string("srl",3));
break;}
case cast<uint32_t>(3ULL):{
return sh(std::string("sra",3));
break;}
case cast<uint32_t>(4ULL):{
return shv(std::string("sllv",4));
break;}
case cast<uint32_t>(6ULL):{
return shv(std::string("srlv",4));
break;}
case cast<uint32_t>(7ULL):{
return shv(std::string("srav",4));
break;}
case cast<uint32_t>(8ULL):{
in.HasDelay = true;
if ((rs == cast<uint32_t>(31ULL))) {
in.Flow = cast<r5900_Flow>(4ULL);
}
else {
in.Flow = cast<r5900_Flow>(5ULL);
}
return set(std::string("jr",2),go_fmt_Sprintf(std::string("jr %s",5),r5900_reg(rs)));
break;}
case cast<uint32_t>(9ULL):{
auto tmp22 = std::make_tuple(cast<r5900_Flow>(6ULL),true);
in.Flow = std::get<0>(tmp22);
in.HasDelay = std::get<1>(tmp22);
if ((rd == cast<uint32_t>(31ULL))) {
return set(std::string("jalr",4),go_fmt_Sprintf(std::string("jalr %s",7),r5900_reg(rs)));
}
return set(std::string("jalr",4),go_fmt_Sprintf(std::string("jalr %s, %s",11),r5900_reg(rd),r5900_reg(rs)));
break;}
case cast<uint32_t>(10ULL):{
return r3(std::string("movz",4));
break;}
case cast<uint32_t>(11ULL):{
return r3(std::string("movn",4));
break;}
case cast<uint32_t>(12ULL):{
in.Flow = cast<r5900_Flow>(7ULL);
return set(std::string("syscall",7),std::string("syscall",7));
break;}
case cast<uint32_t>(13ULL):{
in.Flow = cast<r5900_Flow>(7ULL);
return set(std::string("break",5),std::string("break",5));
break;}
case cast<uint32_t>(15ULL):{
return set(std::string("sync",4),std::string("sync",4));
break;}
case cast<uint32_t>(16ULL):{
return set(std::string("mfhi",4),go_fmt_Sprintf(std::string("mfhi %s",7),r5900_reg(rd)));
break;}
case cast<uint32_t>(17ULL):{
return set(std::string("mthi",4),go_fmt_Sprintf(std::string("mthi %s",7),r5900_reg(rs)));
break;}
case cast<uint32_t>(18ULL):{
return set(std::string("mflo",4),go_fmt_Sprintf(std::string("mflo %s",7),r5900_reg(rd)));
break;}
case cast<uint32_t>(19ULL):{
return set(std::string("mtlo",4),go_fmt_Sprintf(std::string("mtlo %s",7),r5900_reg(rs)));
break;}
case cast<uint32_t>(20ULL):{
return shv(std::string("dsllv",5));
break;}
case cast<uint32_t>(22ULL):{
return shv(std::string("dsrlv",5));
break;}
case cast<uint32_t>(23ULL):{
return shv(std::string("dsrav",5));
break;}
case cast<uint32_t>(24ULL):{
return set(std::string("mult",4),go_fmt_Sprintf(std::string("mult %s, %s, %s",15),r5900_reg(rd),r5900_reg(rs),r5900_reg(rt)));
break;}
case cast<uint32_t>(25ULL):{
return set(std::string("multu",5),go_fmt_Sprintf(std::string("multu %s, %s, %s",16),r5900_reg(rd),r5900_reg(rs),r5900_reg(rt)));
break;}
case cast<uint32_t>(26ULL):{
return r2(std::string("div",3));
break;}
case cast<uint32_t>(27ULL):{
return r2(std::string("divu",4));
break;}
case cast<uint32_t>(32ULL):{
return r3(std::string("add",3));
break;}
case cast<uint32_t>(33ULL):{
if ((rt == cast<uint32_t>(0ULL))) {
return set(std::string("move",4),go_fmt_Sprintf(std::string("move %s, %s",11),r5900_reg(rd),r5900_reg(rs)));
}
return r3(std::string("addu",4));
break;}
case cast<uint32_t>(34ULL):{
return r3(std::string("sub",3));
break;}
case cast<uint32_t>(35ULL):{
return r3(std::string("subu",4));
break;}
case cast<uint32_t>(36ULL):{
return r3(std::string("and",3));
break;}
case cast<uint32_t>(37ULL):{
if ((rt == cast<uint32_t>(0ULL))) {
return set(std::string("move",4),go_fmt_Sprintf(std::string("move %s, %s",11),r5900_reg(rd),r5900_reg(rs)));
}
return r3(std::string("or",2));
break;}
case cast<uint32_t>(38ULL):{
return r3(std::string("xor",3));
break;}
case cast<uint32_t>(39ULL):{
return r3(std::string("nor",3));
break;}
case cast<uint32_t>(40ULL):{
return set(std::string("mfsa",4),go_fmt_Sprintf(std::string("mfsa %s",7),r5900_reg(rd)));
break;}
case cast<uint32_t>(41ULL):{
return set(std::string("mtsa",4),go_fmt_Sprintf(std::string("mtsa %s",7),r5900_reg(rs)));
break;}
case cast<uint32_t>(42ULL):{
return r3(std::string("slt",3));
break;}
case cast<uint32_t>(43ULL):{
return r3(std::string("sltu",4));
break;}
case cast<uint32_t>(44ULL):{
return r3(std::string("dadd",4));
break;}
case cast<uint32_t>(45ULL):{
return r3(std::string("daddu",5));
break;}
case cast<uint32_t>(46ULL):{
return r3(std::string("dsub",4));
break;}
case cast<uint32_t>(47ULL):{
return r3(std::string("dsubu",5));
break;}
case cast<uint32_t>(48ULL):{
return r2(std::string("tge",3));
break;}
case cast<uint32_t>(49ULL):{
return r2(std::string("tgeu",4));
break;}
case cast<uint32_t>(50ULL):{
return r2(std::string("tlt",3));
break;}
case cast<uint32_t>(51ULL):{
return r2(std::string("tltu",4));
break;}
case cast<uint32_t>(52ULL):{
return r2(std::string("teq",3));
break;}
case cast<uint32_t>(54ULL):{
return r2(std::string("tne",3));
break;}
case cast<uint32_t>(56ULL):{
return sh(std::string("dsll",4));
break;}
case cast<uint32_t>(58ULL):{
return sh(std::string("dsrl",4));
break;}
case cast<uint32_t>(59ULL):{
return sh(std::string("dsra",4));
break;}
case cast<uint32_t>(60ULL):{
return set(std::string("dsll32",6),go_fmt_Sprintf(std::string("dsll32 %s, %s, %d",17),r5900_reg(rd),r5900_reg(rt),shamt));
break;}
case cast<uint32_t>(62ULL):{
return set(std::string("dsrl32",6),go_fmt_Sprintf(std::string("dsrl32 %s, %s, %d",17),r5900_reg(rd),r5900_reg(rt),shamt));
break;}
case cast<uint32_t>(63ULL):{
return set(std::string("dsra32",6),go_fmt_Sprintf(std::string("dsra32 %s, %s, %d",17),r5900_reg(rd),r5900_reg(rt),shamt));
break;}
}}
return r5900_word(in,w);
}
}
// tools/cpu/r5900/decode.go:355:1
r5900_Inst r5900_decodeRegimm(r5900_Inst in,uint32_t w,uint32_t rs,uint32_t rt,uint32_t branchT,int32_t simm){
{
auto set = [&](std::string mnem,std::string text)->r5900_Inst{
auto tmp23 = std::make_tuple(mnem,text);
in.Mnem = std::get<0>(tmp23);
in.Text = std::get<1>(tmp23);
return in;
}
;
auto br = [&](std::string m,bool likely)->r5900_Inst{
auto tmp24 = std::make_tuple(cast<r5900_Flow>(1ULL),branchT,true,true);
in.Flow = std::get<0>(tmp24);
in.Target = std::get<1>(tmp24);
in.HasTarget = std::get<2>(tmp24);
in.HasDelay = std::get<3>(tmp24);
in.Annul = likely;
return set(m,go_fmt_Sprintf(std::string("%s %s, $%08X",12),m,r5900_reg(rs),branchT));
}
;
auto trap = [&](std::string m)->r5900_Inst{
return set(m,go_fmt_Sprintf(std::string("%s %s, %d",9),m,r5900_reg(rs),simm));
}
;
{
switch(rt){
case cast<uint32_t>(0ULL):{
return br(std::string("bltz",4),false);
break;}
case cast<uint32_t>(1ULL):{
return br(std::string("bgez",4),false);
break;}
case cast<uint32_t>(2ULL):{
return br(std::string("bltzl",5),true);
break;}
case cast<uint32_t>(3ULL):{
return br(std::string("bgezl",5),true);
break;}
case cast<uint32_t>(8ULL):{
return trap(std::string("tgei",4));
break;}
case cast<uint32_t>(9ULL):{
return trap(std::string("tgeiu",5));
break;}
case cast<uint32_t>(10ULL):{
return trap(std::string("tlti",4));
break;}
case cast<uint32_t>(11ULL):{
return trap(std::string("tltiu",5));
break;}
case cast<uint32_t>(12ULL):{
return trap(std::string("teqi",4));
break;}
case cast<uint32_t>(14ULL):{
return trap(std::string("tnei",4));
break;}
case cast<uint32_t>(16ULL):{
return r5900_call(br(std::string("bltzal",6),false));
break;}
case cast<uint32_t>(17ULL):{
return r5900_call(br(std::string("bgezal",6),false));
break;}
case cast<uint32_t>(18ULL):{
return r5900_call(br(std::string("bltzall",7),true));
break;}
case cast<uint32_t>(19ULL):{
return r5900_call(br(std::string("bgezall",7),true));
break;}
case cast<uint32_t>(24ULL):{
return set(std::string("mtsab",5),go_fmt_Sprintf(std::string("mtsab %s, 0x%X",14),r5900_reg(rs),cast<uint16_t>(simm)));
break;}
case cast<uint32_t>(25ULL):{
return set(std::string("mtsah",5),go_fmt_Sprintf(std::string("mtsah %s, 0x%X",14),r5900_reg(rs),cast<uint16_t>(simm)));
break;}
}}
return r5900_word(in,w);
}
}
// tools/cpu/r5900/decode.go:408:1
r5900_Inst r5900_decodeCop0(r5900_Inst in,uint32_t w,uint32_t rs,uint32_t rt,uint32_t rd){
{
auto set = [&](std::string mnem,std::string text)->r5900_Inst{
auto tmp25 = std::make_tuple(mnem,text);
in.Mnem = std::get<0>(tmp25);
in.Text = std::get<1>(tmp25);
return in;
}
;
if ((cast<uint32_t>((w & cast<uint32_t>(33554432ULL))) != cast<uint32_t>(0ULL))) {
{
switch(cast<uint32_t>((w & cast<uint32_t>(63ULL)))){
case cast<uint32_t>(1ULL):{
return set(std::string("tlbr",4),std::string("tlbr",4));
break;}
case cast<uint32_t>(2ULL):{
return set(std::string("tlbwi",5),std::string("tlbwi",5));
break;}
case cast<uint32_t>(6ULL):{
return set(std::string("tlbwr",5),std::string("tlbwr",5));
break;}
case cast<uint32_t>(8ULL):{
return set(std::string("tlbp",4),std::string("tlbp",4));
break;}
case cast<uint32_t>(24ULL):{
in.Flow = cast<r5900_Flow>(7ULL);
return set(std::string("eret",4),std::string("eret",4));
break;}
case cast<uint32_t>(56ULL):{
return set(std::string("ei",2),std::string("ei",2));
break;}
case cast<uint32_t>(57ULL):{
return set(std::string("di",2),std::string("di",2));
break;}
}}
return r5900_word(in,w);
}
{
switch(rs){
case cast<uint32_t>(0ULL):{
return set(std::string("mfc0",4),go_fmt_Sprintf(std::string("mfc0 %s, %s",11),r5900_reg(rt),r5900_cop0Reg(rd)));
break;}
case cast<uint32_t>(4ULL):{
return set(std::string("mtc0",4),go_fmt_Sprintf(std::string("mtc0 %s, %s",11),r5900_reg(rt),r5900_cop0Reg(rd)));
break;}
}}
return r5900_word(in,w);
}
}
// tools/cpu/r5900/decode.go:442:1
r5900_Inst r5900_decodeCop1(r5900_Inst in,uint32_t w,uint32_t rs,uint32_t rt,uint32_t rd,uint32_t shamt,uint32_t branchT){
{
auto set = [&](std::string mnem,std::string text)->r5900_Inst{
auto tmp26 = std::make_tuple(mnem,text);
in.Mnem = std::get<0>(tmp26);
in.Text = std::get<1>(tmp26);
return in;
}
;
{
switch(rs){
case cast<uint32_t>(0ULL):{
return set(std::string("mfc1",4),go_fmt_Sprintf(std::string("mfc1 %s, %s",11),r5900_reg(rt),r5900_freg(rd)));
break;}
case cast<uint32_t>(2ULL):{
return set(std::string("cfc1",4),go_fmt_Sprintf(std::string("cfc1 %s, $%d",12),r5900_reg(rt),rd));
break;}
case cast<uint32_t>(4ULL):{
return set(std::string("mtc1",4),go_fmt_Sprintf(std::string("mtc1 %s, %s",11),r5900_reg(rt),r5900_freg(rd)));
break;}
case cast<uint32_t>(6ULL):{
return set(std::string("ctc1",4),go_fmt_Sprintf(std::string("ctc1 %s, $%d",12),r5900_reg(rt),rd));
break;}
case cast<uint32_t>(8ULL):{
auto tmp27 = std::make_tuple(cast<r5900_Flow>(1ULL),branchT,true,true);
in.Flow = std::get<0>(tmp27);
in.Target = std::get<1>(tmp27);
in.HasTarget = std::get<2>(tmp27);
in.HasDelay = std::get<3>(tmp27);
std::string m={};
{
switch(cast<uint32_t>((rt & cast<uint32_t>(3ULL)))){
case cast<uint32_t>(0ULL):{
m = std::string("bc1f",4);
break;}
case cast<uint32_t>(1ULL):{
m = std::string("bc1t",4);
break;}
case cast<uint32_t>(2ULL):{
auto tmp28 = std::make_tuple(std::string("bc1fl",5),true);
m = std::get<0>(tmp28);
in.Annul = std::get<1>(tmp28);
break;}
case cast<uint32_t>(3ULL):{
auto tmp29 = std::make_tuple(std::string("bc1tl",5),true);
m = std::get<0>(tmp29);
in.Annul = std::get<1>(tmp29);
break;}
}}
return set(m,go_fmt_Sprintf(std::string("%s $%08X",8),m,branchT));
break;}
case cast<uint32_t>(16ULL):{
auto tmp30 = std::make_tuple(rt,rd,shamt);
uint32_t ft = std::get<0>(tmp30);
uint32_t fs = std::get<1>(tmp30);
uint32_t fd = std::get<2>(tmp30);
uint32_t funct = cast<uint32_t>((w & cast<uint32_t>(63ULL)));
{
switch(funct){
case cast<uint32_t>(48ULL):{
return set(std::string("c.f.s",5),go_fmt_Sprintf(std::string("c.f.s %s, %s",12),r5900_freg(fs),r5900_freg(ft)));
break;}
case cast<uint32_t>(50ULL):{
return set(std::string("c.eq.s",6),go_fmt_Sprintf(std::string("c.eq.s %s, %s",13),r5900_freg(fs),r5900_freg(ft)));
break;}
case cast<uint32_t>(52ULL):{
return set(std::string("c.lt.s",6),go_fmt_Sprintf(std::string("c.lt.s %s, %s",13),r5900_freg(fs),r5900_freg(ft)));
break;}
case cast<uint32_t>(54ULL):{
return set(std::string("c.le.s",6),go_fmt_Sprintf(std::string("c.le.s %s, %s",13),r5900_freg(fs),r5900_freg(ft)));
break;}
}}
{
auto tmp31 = lookup(Map<uint32_t,std::string>{{cast<uint32_t>(0ULL),std::string("add.s",5)},{cast<uint32_t>(1ULL),std::string("sub.s",5)},{cast<uint32_t>(2ULL),std::string("mul.s",5)},{cast<uint32_t>(3ULL),std::string("div.s",5)},{cast<uint32_t>(40ULL),std::string("max.s",5)},{cast<uint32_t>(41ULL),std::string("min.s",5)}},funct);
std::string n = std::get<0>(tmp31);
bool ok = std::get<1>(tmp31);
if (ok) {
return set(n,go_fmt_Sprintf(std::string("%s %s, %s, %s",13),n,r5900_freg(fd),r5900_freg(fs),r5900_freg(ft)));
}
}
{
auto tmp32 = lookup(Map<uint32_t,std::string>{{cast<uint32_t>(4ULL),std::string("sqrt.s",6)},{cast<uint32_t>(5ULL),std::string("abs.s",5)},{cast<uint32_t>(6ULL),std::string("mov.s",5)},{cast<uint32_t>(7ULL),std::string("neg.s",5)},{cast<uint32_t>(22ULL),std::string("rsqrt.s",7)},{cast<uint32_t>(36ULL),std::string("cvt.w.s",7)}},funct);
std::string n = std::get<0>(tmp32);
bool ok = std::get<1>(tmp32);
if (ok) {
return set(n,go_fmt_Sprintf(std::string("%s %s, %s",9),n,r5900_freg(fd),r5900_freg(fs)));
}
}
{
auto tmp33 = lookup(Map<uint32_t,std::string>{{cast<uint32_t>(24ULL),std::string("adda.s",6)},{cast<uint32_t>(25ULL),std::string("suba.s",6)},{cast<uint32_t>(26ULL),std::string("mula.s",6)},{cast<uint32_t>(30ULL),std::string("madda.s",7)},{cast<uint32_t>(31ULL),std::string("msuba.s",7)}},funct);
std::string n = std::get<0>(tmp33);
bool ok = std::get<1>(tmp33);
if (ok) {
return set(n,go_fmt_Sprintf(std::string("%s %s, %s",9),n,r5900_freg(fs),r5900_freg(ft)));
}
}
{
auto tmp34 = lookup(Map<uint32_t,std::string>{{cast<uint32_t>(28ULL),std::string("madd.s",6)},{cast<uint32_t>(29ULL),std::string("msub.s",6)}},funct);
std::string n = std::get<0>(tmp34);
bool ok = std::get<1>(tmp34);
if (ok) {
return set(n,go_fmt_Sprintf(std::string("%s %s, %s, %s",13),n,r5900_freg(fd),r5900_freg(fs),r5900_freg(ft)));
}
}
return r5900_word(in,w);
break;}
case cast<uint32_t>(20ULL):{
if ((cast<uint32_t>((w & cast<uint32_t>(63ULL))) == cast<uint32_t>(32ULL))) {
return set(std::string("cvt.s.w",7),go_fmt_Sprintf(std::string("cvt.s.w %s, %s",14),r5900_freg(shamt),r5900_freg(rd)));
}
return r5900_word(in,w);
break;}
}}
return r5900_word(in,w);
}
}
// tools/cpu/r5900/decode.go:526:1
r5900_Inst r5900_decodeCop2(r5900_Inst in,uint32_t w,uint32_t rs,uint32_t rt,uint32_t rd,uint32_t branchT){
{
auto set = [&](std::string mnem,std::string text)->r5900_Inst{
auto tmp35 = std::make_tuple(mnem,text);
in.Mnem = std::get<0>(tmp35);
in.Text = std::get<1>(tmp35);
return in;
}
;
if ((cast<uint32_t>((w & cast<uint32_t>(33554432ULL))) != cast<uint32_t>(0ULL))) {
return set(std::string("vu",2),vu_DisasmMacro(w));
}
{
switch(rs){
case cast<uint32_t>(1ULL):{
return set(std::string("qmfc2",5),go_fmt_Sprintf(std::string("qmfc2 %s, %s",12),r5900_reg(rt),r5900_vfreg(rd)));
break;}
case cast<uint32_t>(2ULL):{
return set(std::string("cfc2",4),go_fmt_Sprintf(std::string("cfc2 %s, $vi%d",14),r5900_reg(rt),rd));
break;}
case cast<uint32_t>(5ULL):{
return set(std::string("qmtc2",5),go_fmt_Sprintf(std::string("qmtc2 %s, %s",12),r5900_reg(rt),r5900_vfreg(rd)));
break;}
case cast<uint32_t>(6ULL):{
return set(std::string("ctc2",4),go_fmt_Sprintf(std::string("ctc2 %s, $vi%d",14),r5900_reg(rt),rd));
break;}
case cast<uint32_t>(8ULL):{
auto tmp36 = std::make_tuple(cast<r5900_Flow>(1ULL),branchT,true,true);
in.Flow = std::get<0>(tmp36);
in.Target = std::get<1>(tmp36);
in.HasTarget = std::get<2>(tmp36);
in.HasDelay = std::get<3>(tmp36);
std::string m={};
{
switch(cast<uint32_t>((rt & cast<uint32_t>(3ULL)))){
case cast<uint32_t>(0ULL):{
m = std::string("bc2f",4);
break;}
case cast<uint32_t>(1ULL):{
m = std::string("bc2t",4);
break;}
case cast<uint32_t>(2ULL):{
auto tmp37 = std::make_tuple(std::string("bc2fl",5),true);
m = std::get<0>(tmp37);
in.Annul = std::get<1>(tmp37);
break;}
case cast<uint32_t>(3ULL):{
auto tmp38 = std::make_tuple(std::string("bc2tl",5),true);
m = std::get<0>(tmp38);
in.Annul = std::get<1>(tmp38);
break;}
}}
return set(m,go_fmt_Sprintf(std::string("%s $%08X",8),m,branchT));
break;}
}}
return r5900_word(in,w);
}
}
// tools/cpu/r5900/decode.go:560:1
r5900_Inst r5900_call(r5900_Inst in){
{
in.Flow = cast<r5900_Flow>(3ULL);
return in;
}
}
// tools/cpu/r5900/decode.go:563:1
r5900_Inst r5900_word(r5900_Inst in,uint32_t w){
{
auto tmp39 = std::make_tuple(std::string(".word",5),go_fmt_Sprintf(std::string(".word 0x%08X",12),w),cast<r5900_Flow>(7ULL));
in.Mnem = std::get<0>(tmp39);
in.Text = std::get<1>(tmp39);
in.Flow = std::get<2>(tmp39);
auto tmp40 = std::make_tuple(false,false,false);
in.HasTarget = std::get<0>(tmp40);
in.HasDelay = std::get<1>(tmp40);
in.Annul = std::get<2>(tmp40);
return in;
}
}
// tools/cpu/r5900/disasm.go:8:1
Slice<std::string> r5900_Disassemble(Slice<uint8_t> code,uint32_t base){
{
Slice<std::string> out={};
{int64_t i = cast<int64_t>(0ULL);for (;(cast<int64_t>((i + cast<int64_t>(4ULL))) <= len(code));i += cast<int64_t>(4ULL)){
uint32_t addr = cast<uint32_t>((base + cast<uint32_t>(i)));
r5900_Inst in = r5900_Decode(sub(code,i,len(code)),addr);
out = append(out,Slice<std::string>{go_fmt_Sprintf(std::string("%08X  %02X %02X %02X %02X  %s",29),addr,code[i],code[cast<int64_t>((i + cast<int64_t>(1ULL)))],code[cast<int64_t>((i + cast<int64_t>(2ULL)))],code[cast<int64_t>((i + cast<int64_t>(3ULL)))],in.Text)});
}
}return out;
}
}
// tools/cpu/r5900/exec.go:20:1
int64_t r5900_CPU_Step(r5900_CPU* c){
{
if (c->Halted) {
return cast<int64_t>(0ULL);
}
r5900_CPU_tickCount(c);
if (r5900_CPU_checkInterrupt(c)) {
return cast<int64_t>(1ULL);
}
c->curPC = c->PC;
c->delaySlot = c->pendingDelay;
c->pendingDelay = false;
auto tmp41 = r5900_CPU_translateFetch(c,c->PC);
uint32_t paddr = std::get<0>(tmp41);
bool ok = std::get<1>(tmp41);
if ((!ok)) {
return cast<int64_t>(1ULL);
}
uint32_t w = c->fetch(paddr);
c->PC = c->nextPC;
c->nextPC += cast<uint64_t>(4ULL);
r5900_CPU_execute(c,w);
c->Steps++;
return cast<int64_t>(1ULL);
}
}
// tools/cpu/r5900/exec.go:49:1
std::tuple<uint32_t,bool> r5900_CPU_translateFetch(r5900_CPU* c,uint64_t vaddr){
{
if ((cast<uint64_t>((vaddr & cast<uint64_t>(3ULL))) != cast<uint64_t>(0ULL))) {
r5900_CPU_addrError(c,cast<uint32_t>(4ULL),vaddr);
return {cast<uint32_t>(0ULL),false};
}
return r5900_CPU_Translate(c,vaddr,false);
}
}
// tools/cpu/r5900/exec.go:59:1
void r5900_CPU_doBranch(r5900_CPU* c,bool taken,uint64_t target){
{
c->pendingDelay = true;
c->branchAddr = c->curPC;
if (taken) {
c->nextPC = target;
}
}
}
// tools/cpu/r5900/exec.go:70:1
void r5900_CPU_doBranchLikely(r5900_CPU* c,bool taken,uint64_t target){
{
if (taken) {
r5900_CPU_doBranch(c,true,target);
return ;
}
c->PC = c->nextPC;
c->nextPC += cast<uint64_t>(4ULL);
}
}
// tools/cpu/r5900/exec.go:79:1
void r5900_CPU_execute(r5900_CPU* c,uint32_t w){
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
r5900_CPU_special(c,w,rs,rt);
break;}
case cast<uint32_t>(1ULL):{
r5900_CPU_regimm(c,w,rs,rt,branchT,simm);
break;}
case cast<uint32_t>(28ULL):{
r5900_CPU_mmi(c,w,rs,rt,cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(11ULL))) & cast<uint32_t>(31ULL))),cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(6ULL))) & cast<uint32_t>(31ULL))));
break;}
case cast<uint32_t>(2ULL):{
r5900_CPU_doBranch(c,true,jumpT);
break;}
case cast<uint32_t>(3ULL):{
r5900_CPU_set(c,cast<uint32_t>(31ULL),cast<uint64_t>((c->curPC + cast<uint64_t>(8ULL))));
r5900_CPU_doBranch(c,true,jumpT);
break;}
case cast<uint32_t>(4ULL):{
r5900_CPU_doBranch(c,(c->R[rs].Lo == c->R[rt].Lo),branchT);
break;}
case cast<uint32_t>(5ULL):{
r5900_CPU_doBranch(c,(c->R[rs].Lo != c->R[rt].Lo),branchT);
break;}
case cast<uint32_t>(6ULL):{
r5900_CPU_doBranch(c,(cast<int64_t>(c->R[rs].Lo) <= cast<int64_t>(0ULL)),branchT);
break;}
case cast<uint32_t>(7ULL):{
r5900_CPU_doBranch(c,(cast<int64_t>(c->R[rs].Lo) > cast<int64_t>(0ULL)),branchT);
break;}
case cast<uint32_t>(20ULL):{
r5900_CPU_doBranchLikely(c,(c->R[rs].Lo == c->R[rt].Lo),branchT);
break;}
case cast<uint32_t>(21ULL):{
r5900_CPU_doBranchLikely(c,(c->R[rs].Lo != c->R[rt].Lo),branchT);
break;}
case cast<uint32_t>(22ULL):{
r5900_CPU_doBranchLikely(c,(cast<int64_t>(c->R[rs].Lo) <= cast<int64_t>(0ULL)),branchT);
break;}
case cast<uint32_t>(23ULL):{
r5900_CPU_doBranchLikely(c,(cast<int64_t>(c->R[rs].Lo) > cast<int64_t>(0ULL)),branchT);
break;}
case cast<uint32_t>(8ULL):{
{
auto tmp42 = r5900_addOv32(cast<uint32_t>(c->R[rs].Lo),cast<uint32_t>(simm));
uint32_t r = std::get<0>(tmp42);
bool ov = std::get<1>(tmp42);
if (ov) {
r5900_CPU_Exception(c,cast<uint32_t>(12ULL));
}
else {
r5900_CPU_set(c,rt,r5900_sext32(r));
}
}
break;}
case cast<uint32_t>(9ULL):{
r5900_CPU_set(c,rt,r5900_sext32(cast<uint32_t>((cast<uint32_t>(c->R[rs].Lo) + cast<uint32_t>(simm)))));
break;}
case cast<uint32_t>(10ULL):{
r5900_CPU_set(c,rt,r5900_b2u((cast<int64_t>(c->R[rs].Lo) < cast<int64_t>(simm))));
break;}
case cast<uint32_t>(11ULL):{
r5900_CPU_set(c,rt,r5900_b2u((c->R[rs].Lo < simm)));
break;}
case cast<uint32_t>(12ULL):{
r5900_CPU_set(c,rt,cast<uint64_t>((c->R[rs].Lo & cast<uint64_t>(imm))));
break;}
case cast<uint32_t>(13ULL):{
r5900_CPU_set(c,rt,cast<uint64_t>((c->R[rs].Lo | cast<uint64_t>(imm))));
break;}
case cast<uint32_t>(14ULL):{
r5900_CPU_set(c,rt,cast<uint64_t>((c->R[rs].Lo ^ cast<uint64_t>(imm))));
break;}
case cast<uint32_t>(15ULL):{
r5900_CPU_set(c,rt,r5900_sext32(shl<uint32_t>(imm,cast<int64_t>(16ULL))));
break;}
case cast<uint32_t>(24ULL):{
{
auto tmp43 = r5900_addOv64(c->R[rs].Lo,simm);
uint64_t r = std::get<0>(tmp43);
bool ov = std::get<1>(tmp43);
if (ov) {
r5900_CPU_Exception(c,cast<uint32_t>(12ULL));
}
else {
r5900_CPU_set(c,rt,r);
}
}
break;}
case cast<uint32_t>(25ULL):{
r5900_CPU_set(c,rt,cast<uint64_t>((c->R[rs].Lo + simm)));
break;}
case cast<uint32_t>(16ULL):{
r5900_CPU_cop0(c,w,rs,rt,cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(11ULL))) & cast<uint32_t>(31ULL))));
break;}
case cast<uint32_t>(17ULL):{
if ((cast<uint64_t>((c->COP0[cast<int64_t>(12ULL)] & cast<uint64_t>(536870912ULL))) == cast<uint64_t>(0ULL))) {
r5900_CPU_coprocessorUnusable(c,cast<uint32_t>(1ULL));
return ;
}
r5900_CPU_cop1(c,w,rs,rt,cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(11ULL))) & cast<uint32_t>(31ULL))),cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(6ULL))) & cast<uint32_t>(31ULL))),branchT);
break;}
case cast<uint32_t>(18ULL):{
r5900_CPU_cop2(c,w,rs,rt,cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(11ULL))) & cast<uint32_t>(31ULL))),branchT);
break;}
case cast<uint32_t>(32ULL):case cast<uint32_t>(33ULL):case cast<uint32_t>(34ULL):case cast<uint32_t>(35ULL):case cast<uint32_t>(36ULL):case cast<uint32_t>(37ULL):case cast<uint32_t>(38ULL):case cast<uint32_t>(39ULL):case cast<uint32_t>(26ULL):case cast<uint32_t>(27ULL):case cast<uint32_t>(48ULL):case cast<uint32_t>(52ULL):case cast<uint32_t>(55ULL):case cast<uint32_t>(30ULL):{
r5900_CPU_loadOp(c,op,rs,rt,simm);
break;}
case cast<uint32_t>(40ULL):case cast<uint32_t>(41ULL):case cast<uint32_t>(42ULL):case cast<uint32_t>(43ULL):case cast<uint32_t>(44ULL):case cast<uint32_t>(45ULL):case cast<uint32_t>(46ULL):case cast<uint32_t>(56ULL):case cast<uint32_t>(60ULL):case cast<uint32_t>(63ULL):case cast<uint32_t>(31ULL):{
r5900_CPU_storeOp(c,op,rs,rt,simm);
break;}
case cast<uint32_t>(47ULL):case cast<uint32_t>(51ULL):{
break;}
case cast<uint32_t>(49ULL):case cast<uint32_t>(57ULL):{
if ((cast<uint64_t>((c->COP0[cast<int64_t>(12ULL)] & cast<uint64_t>(536870912ULL))) == cast<uint64_t>(0ULL))) {
r5900_CPU_coprocessorUnusable(c,cast<uint32_t>(1ULL));
return ;
}
r5900_CPU_cop1Mem(c,op,rs,rt,simm);
break;}
case cast<uint32_t>(54ULL):case cast<uint32_t>(62ULL):{
r5900_CPU_cop2Mem(c,op,rs,rt,simm);
break;}
default:{
r5900_CPU_Exception(c,cast<uint32_t>(10ULL));
break;}
}}
}
}
// tools/cpu/r5900/exec.go:191:1
void r5900_CPU_cop2(r5900_CPU* c,uint32_t w,uint32_t rs,uint32_t rt,uint32_t rd,uint64_t branchT){
{
if ((cast<uint64_t>((c->COP0[cast<int64_t>(12ULL)] & cast<uint64_t>(1073741824ULL))) == cast<uint64_t>(0ULL))) {
r5900_CPU_coprocessorUnusable(c,cast<uint32_t>(2ULL));
return ;
}
if ((!c->COP2)) {
{
switch(rs){
case cast<uint32_t>(1ULL):{
r5900_CPU_setQ(c,rt,cast<uint64_t>(0ULL),cast<uint64_t>(0ULL));
break;}
case cast<uint32_t>(2ULL):{
r5900_CPU_set(c,rt,cast<uint64_t>(0ULL));
break;}
}}
return ;
}
if ((cast<uint32_t>((w & cast<uint32_t>(33554432ULL))) != cast<uint32_t>(0ULL))) {
vu_VU_Macro(c->COP2,w);
return ;
}
{
switch(rs){
case cast<uint32_t>(1ULL):{
r5900_CPU_setQ(c,rt,cast<uint64_t>((cast<uint64_t>(vu_VU_ReadVF(c->COP2,rd,cast<uint32_t>(0ULL))) | shl<uint64_t>(cast<uint64_t>(vu_VU_ReadVF(c->COP2,rd,cast<uint32_t>(1ULL))),cast<int64_t>(32ULL)))),cast<uint64_t>((cast<uint64_t>(vu_VU_ReadVF(c->COP2,rd,cast<uint32_t>(2ULL))) | shl<uint64_t>(cast<uint64_t>(vu_VU_ReadVF(c->COP2,rd,cast<uint32_t>(3ULL))),cast<int64_t>(32ULL)))));
break;}
case cast<uint32_t>(2ULL):{
r5900_CPU_set(c,rt,r5900_sext32(vu_VU_ReadCtrl(c->COP2,rd)));
break;}
case cast<uint32_t>(5ULL):{
r5900_Quad q = c->R[rt];
vu_VU_WriteVF(c->COP2,rd,cast<uint32_t>(0ULL),cast<uint32_t>(q.Lo));
vu_VU_WriteVF(c->COP2,rd,cast<uint32_t>(1ULL),cast<uint32_t>(shr<uint64_t>(q.Lo,cast<int64_t>(32ULL))));
vu_VU_WriteVF(c->COP2,rd,cast<uint32_t>(2ULL),cast<uint32_t>(q.Hi));
vu_VU_WriteVF(c->COP2,rd,cast<uint32_t>(3ULL),cast<uint32_t>(shr<uint64_t>(q.Hi,cast<int64_t>(32ULL))));
break;}
case cast<uint32_t>(6ULL):{
vu_VU_WriteCtrl(c->COP2,rd,cast<uint32_t>(c->R[rt].Lo));
break;}
case cast<uint32_t>(8ULL):{
bool cond = (vu_VU_ReadCtrl(c->COP2,cast<uint32_t>(18ULL)) != cast<uint32_t>(0ULL));
{
switch(cast<uint32_t>((rt & cast<uint32_t>(3ULL)))){
case cast<uint32_t>(0ULL):{
r5900_CPU_doBranch(c,(!cond),branchT);
break;}
case cast<uint32_t>(1ULL):{
r5900_CPU_doBranch(c,cond,branchT);
break;}
case cast<uint32_t>(2ULL):{
r5900_CPU_doBranchLikely(c,(!cond),branchT);
break;}
case cast<uint32_t>(3ULL):{
r5900_CPU_doBranchLikely(c,cond,branchT);
break;}
}}
break;}
default:{
r5900_CPU_Exception(c,cast<uint32_t>(10ULL));
break;}
}}
}
}
// tools/cpu/r5900/exec.go:258:1
void r5900_CPU_cop2Mem(r5900_CPU* c,uint32_t op,uint32_t rs,uint32_t rt,uint64_t simm){
{
if ((!c->COP2)) {
return ;
}
uint64_t vaddr = cast<uint64_t>((c->R[rs].Lo + simm));
if ((cast<uint64_t>((vaddr & cast<uint64_t>(15ULL))) != cast<uint64_t>(0ULL))) {
uint32_t code = cast<uint32_t>(4ULL);
if ((op == cast<uint32_t>(62ULL))) {
code = cast<uint32_t>(5ULL);
}
r5900_CPU_addrError(c,code,vaddr);
return ;
}
{
switch(op){
case cast<uint32_t>(54ULL):{
auto tmp44 = r5900_CPU_Translate(c,vaddr,false);
uint32_t p = std::get<0>(tmp44);
bool ok = std::get<1>(tmp44);
if ((!ok)) {
return ;
}
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < cast<uint32_t>(4ULL));i++){
vu_VU_WriteVF(c->COP2,rt,i,r5900_CPU_read32(c,cast<uint32_t>((p + cast<uint32_t>((i * cast<uint32_t>(4ULL)))))));
}
}break;}
case cast<uint32_t>(62ULL):{
auto tmp45 = r5900_CPU_Translate(c,vaddr,true);
uint32_t p = std::get<0>(tmp45);
bool ok = std::get<1>(tmp45);
if ((!ok)) {
return ;
}
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < cast<uint32_t>(4ULL));i++){
r5900_CPU_write32(c,cast<uint32_t>((p + cast<uint32_t>((i * cast<uint32_t>(4ULL))))),vu_VU_ReadVF(c->COP2,rt,i));
}
}break;}
}}
}
}
// tools/cpu/r5900/exec.go:291:1
void r5900_CPU_special(r5900_CPU* c,uint32_t w,uint32_t rs,uint32_t rt){
{
uint32_t rd = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(11ULL))) & cast<uint32_t>(31ULL)));
uint32_t shamt = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(6ULL))) & cast<uint32_t>(31ULL)));
uint32_t funct = cast<uint32_t>((w & cast<uint32_t>(63ULL)));
{
switch(funct){
case cast<uint32_t>(0ULL):{
r5900_CPU_set(c,rd,r5900_sext32(shl<uint32_t>(cast<uint32_t>(c->R[rt].Lo),shamt)));
break;}
case cast<uint32_t>(2ULL):{
r5900_CPU_set(c,rd,r5900_sext32(shr<uint32_t>(cast<uint32_t>(c->R[rt].Lo),shamt)));
break;}
case cast<uint32_t>(3ULL):{
r5900_CPU_set(c,rd,r5900_sext32(cast<uint32_t>(shr<int32_t>(cast<int32_t>(cast<uint32_t>(c->R[rt].Lo)),shamt))));
break;}
case cast<uint32_t>(4ULL):{
r5900_CPU_set(c,rd,r5900_sext32(shl<uint32_t>(cast<uint32_t>(c->R[rt].Lo),(cast<uint64_t>((c->R[rs].Lo & cast<uint64_t>(31ULL)))))));
break;}
case cast<uint32_t>(6ULL):{
r5900_CPU_set(c,rd,r5900_sext32(shr<uint32_t>(cast<uint32_t>(c->R[rt].Lo),(cast<uint64_t>((c->R[rs].Lo & cast<uint64_t>(31ULL)))))));
break;}
case cast<uint32_t>(7ULL):{
r5900_CPU_set(c,rd,r5900_sext32(cast<uint32_t>(shr<int32_t>(cast<int32_t>(cast<uint32_t>(c->R[rt].Lo)),(cast<uint64_t>((c->R[rs].Lo & cast<uint64_t>(31ULL))))))));
break;}
case cast<uint32_t>(8ULL):{
r5900_CPU_doBranch(c,true,c->R[rs].Lo);
break;}
case cast<uint32_t>(9ULL):{
r5900_CPU_set(c,rd,cast<uint64_t>((c->curPC + cast<uint64_t>(8ULL))));
r5900_CPU_doBranch(c,true,c->R[rs].Lo);
break;}
case cast<uint32_t>(10ULL):{
if ((c->R[rt].Lo == cast<uint64_t>(0ULL))) {
r5900_CPU_set(c,rd,c->R[rs].Lo);
}
break;}
case cast<uint32_t>(11ULL):{
if ((c->R[rt].Lo != cast<uint64_t>(0ULL))) {
r5900_CPU_set(c,rd,c->R[rs].Lo);
}
break;}
case cast<uint32_t>(12ULL):{
if ((bool(c->Syscall) && c->Syscall(c))) {
return ;
}
r5900_CPU_Exception(c,cast<uint32_t>(8ULL));
break;}
case cast<uint32_t>(13ULL):{
r5900_CPU_Exception(c,cast<uint32_t>(9ULL));
break;}
case cast<uint32_t>(15ULL):{
break;}
case cast<uint32_t>(16ULL):{
r5900_CPU_set(c,rd,c->HI);
break;}
case cast<uint32_t>(17ULL):{
c->HI = c->R[rs].Lo;
break;}
case cast<uint32_t>(18ULL):{
r5900_CPU_set(c,rd,c->LO);
break;}
case cast<uint32_t>(19ULL):{
c->LO = c->R[rs].Lo;
break;}
case cast<uint32_t>(20ULL):{
r5900_CPU_set(c,rd,shl<uint64_t>(c->R[rt].Lo,(cast<uint64_t>((c->R[rs].Lo & cast<uint64_t>(63ULL))))));
break;}
case cast<uint32_t>(22ULL):{
r5900_CPU_set(c,rd,shr<uint64_t>(c->R[rt].Lo,(cast<uint64_t>((c->R[rs].Lo & cast<uint64_t>(63ULL))))));
break;}
case cast<uint32_t>(23ULL):{
r5900_CPU_set(c,rd,cast<uint64_t>(shr<int64_t>(cast<int64_t>(c->R[rt].Lo),(cast<uint64_t>((c->R[rs].Lo & cast<uint64_t>(63ULL)))))));
break;}
case cast<uint32_t>(24ULL):{
int64_t p = cast<int64_t>((cast<int64_t>(cast<int32_t>(cast<uint32_t>(c->R[rs].Lo))) * cast<int64_t>(cast<int32_t>(cast<uint32_t>(c->R[rt].Lo)))));
auto tmp46 = std::make_tuple(r5900_sext32(cast<uint32_t>(p)),r5900_sext32(cast<uint32_t>(shr<uint64_t>(cast<uint64_t>(p),cast<int64_t>(32ULL)))));
c->LO = std::get<0>(tmp46);
c->HI = std::get<1>(tmp46);
r5900_CPU_set(c,rd,c->LO);
break;}
case cast<uint32_t>(25ULL):{
uint64_t p = cast<uint64_t>((cast<uint64_t>(cast<uint32_t>(c->R[rs].Lo)) * cast<uint64_t>(cast<uint32_t>(c->R[rt].Lo))));
auto tmp47 = std::make_tuple(r5900_sext32(cast<uint32_t>(p)),r5900_sext32(cast<uint32_t>(shr<uint64_t>(p,cast<int64_t>(32ULL)))));
c->LO = std::get<0>(tmp47);
c->HI = std::get<1>(tmp47);
r5900_CPU_set(c,rd,c->LO);
break;}
case cast<uint32_t>(26ULL):{
auto tmp48 = r5900_divSigned32(cast<int32_t>(cast<uint32_t>(c->R[rs].Lo)),cast<int32_t>(cast<uint32_t>(c->R[rt].Lo)));
c->LO = std::get<0>(tmp48);
c->HI = std::get<1>(tmp48);
break;}
case cast<uint32_t>(27ULL):{
auto tmp49 = r5900_divUnsigned32(cast<uint32_t>(c->R[rs].Lo),cast<uint32_t>(c->R[rt].Lo));
c->LO = std::get<0>(tmp49);
c->HI = std::get<1>(tmp49);
break;}
case cast<uint32_t>(32ULL):{
{
auto tmp50 = r5900_addOv32(cast<uint32_t>(c->R[rs].Lo),cast<uint32_t>(c->R[rt].Lo));
uint32_t r = std::get<0>(tmp50);
bool ov = std::get<1>(tmp50);
if (ov) {
r5900_CPU_Exception(c,cast<uint32_t>(12ULL));
}
else {
r5900_CPU_set(c,rd,r5900_sext32(r));
}
}
break;}
case cast<uint32_t>(33ULL):{
r5900_CPU_set(c,rd,r5900_sext32(cast<uint32_t>((cast<uint32_t>(c->R[rs].Lo) + cast<uint32_t>(c->R[rt].Lo)))));
break;}
case cast<uint32_t>(34ULL):{
{
auto tmp51 = r5900_subOv32(cast<uint32_t>(c->R[rs].Lo),cast<uint32_t>(c->R[rt].Lo));
uint32_t r = std::get<0>(tmp51);
bool ov = std::get<1>(tmp51);
if (ov) {
r5900_CPU_Exception(c,cast<uint32_t>(12ULL));
}
else {
r5900_CPU_set(c,rd,r5900_sext32(r));
}
}
break;}
case cast<uint32_t>(35ULL):{
r5900_CPU_set(c,rd,r5900_sext32(cast<uint32_t>((cast<uint32_t>(c->R[rs].Lo) - cast<uint32_t>(c->R[rt].Lo)))));
break;}
case cast<uint32_t>(36ULL):{
r5900_CPU_set(c,rd,cast<uint64_t>((c->R[rs].Lo & c->R[rt].Lo)));
break;}
case cast<uint32_t>(37ULL):{
r5900_CPU_set(c,rd,cast<uint64_t>((c->R[rs].Lo | c->R[rt].Lo)));
break;}
case cast<uint32_t>(38ULL):{
r5900_CPU_set(c,rd,cast<uint64_t>((c->R[rs].Lo ^ c->R[rt].Lo)));
break;}
case cast<uint32_t>(39ULL):{
r5900_CPU_set(c,rd,cast<uint64_t>(~(cast<uint64_t>((c->R[rs].Lo | c->R[rt].Lo)))));
break;}
case cast<uint32_t>(40ULL):{
r5900_CPU_set(c,rd,cast<uint64_t>(c->SA));
break;}
case cast<uint32_t>(41ULL):{
c->SA = cast<uint32_t>((cast<uint32_t>(c->R[rs].Lo) & cast<uint32_t>(127ULL)));
break;}
case cast<uint32_t>(42ULL):{
r5900_CPU_set(c,rd,r5900_b2u((cast<int64_t>(c->R[rs].Lo) < cast<int64_t>(c->R[rt].Lo))));
break;}
case cast<uint32_t>(43ULL):{
r5900_CPU_set(c,rd,r5900_b2u((c->R[rs].Lo < c->R[rt].Lo)));
break;}
case cast<uint32_t>(44ULL):{
{
auto tmp52 = r5900_addOv64(c->R[rs].Lo,c->R[rt].Lo);
uint64_t r = std::get<0>(tmp52);
bool ov = std::get<1>(tmp52);
if (ov) {
r5900_CPU_Exception(c,cast<uint32_t>(12ULL));
}
else {
r5900_CPU_set(c,rd,r);
}
}
break;}
case cast<uint32_t>(45ULL):{
r5900_CPU_set(c,rd,cast<uint64_t>((c->R[rs].Lo + c->R[rt].Lo)));
break;}
case cast<uint32_t>(46ULL):{
{
auto tmp53 = r5900_subOv64(c->R[rs].Lo,c->R[rt].Lo);
uint64_t r = std::get<0>(tmp53);
bool ov = std::get<1>(tmp53);
if (ov) {
r5900_CPU_Exception(c,cast<uint32_t>(12ULL));
}
else {
r5900_CPU_set(c,rd,r);
}
}
break;}
case cast<uint32_t>(47ULL):{
r5900_CPU_set(c,rd,cast<uint64_t>((c->R[rs].Lo - c->R[rt].Lo)));
break;}
case cast<uint32_t>(48ULL):{
r5900_CPU_trapIf(c,(cast<int64_t>(c->R[rs].Lo) >= cast<int64_t>(c->R[rt].Lo)));
break;}
case cast<uint32_t>(49ULL):{
r5900_CPU_trapIf(c,(c->R[rs].Lo >= c->R[rt].Lo));
break;}
case cast<uint32_t>(50ULL):{
r5900_CPU_trapIf(c,(cast<int64_t>(c->R[rs].Lo) < cast<int64_t>(c->R[rt].Lo)));
break;}
case cast<uint32_t>(51ULL):{
r5900_CPU_trapIf(c,(c->R[rs].Lo < c->R[rt].Lo));
break;}
case cast<uint32_t>(52ULL):{
r5900_CPU_trapIf(c,(c->R[rs].Lo == c->R[rt].Lo));
break;}
case cast<uint32_t>(54ULL):{
r5900_CPU_trapIf(c,(c->R[rs].Lo != c->R[rt].Lo));
break;}
case cast<uint32_t>(56ULL):{
r5900_CPU_set(c,rd,shl<uint64_t>(c->R[rt].Lo,shamt));
break;}
case cast<uint32_t>(58ULL):{
r5900_CPU_set(c,rd,shr<uint64_t>(c->R[rt].Lo,shamt));
break;}
case cast<uint32_t>(59ULL):{
r5900_CPU_set(c,rd,cast<uint64_t>(shr<int64_t>(cast<int64_t>(c->R[rt].Lo),shamt)));
break;}
case cast<uint32_t>(60ULL):{
r5900_CPU_set(c,rd,shl<uint64_t>(c->R[rt].Lo,(cast<uint32_t>((shamt + cast<uint32_t>(32ULL))))));
break;}
case cast<uint32_t>(62ULL):{
r5900_CPU_set(c,rd,shr<uint64_t>(c->R[rt].Lo,(cast<uint32_t>((shamt + cast<uint32_t>(32ULL))))));
break;}
case cast<uint32_t>(63ULL):{
r5900_CPU_set(c,rd,cast<uint64_t>(shr<int64_t>(cast<int64_t>(c->R[rt].Lo),(cast<uint32_t>((shamt + cast<uint32_t>(32ULL)))))));
break;}
default:{
r5900_CPU_Exception(c,cast<uint32_t>(10ULL));
break;}
}}
}
}
// tools/cpu/r5900/exec.go:466:1
void r5900_CPU_regimm(r5900_CPU* c,uint32_t w,uint32_t rs,uint32_t rt,uint64_t branchT,uint64_t simm){
{
int64_t s = cast<int64_t>(c->R[rs].Lo);
{
switch(rt){
case cast<uint32_t>(0ULL):{
r5900_CPU_doBranch(c,(s < cast<int64_t>(0ULL)),branchT);
break;}
case cast<uint32_t>(1ULL):{
r5900_CPU_doBranch(c,(s >= cast<int64_t>(0ULL)),branchT);
break;}
case cast<uint32_t>(2ULL):{
r5900_CPU_doBranchLikely(c,(s < cast<int64_t>(0ULL)),branchT);
break;}
case cast<uint32_t>(3ULL):{
r5900_CPU_doBranchLikely(c,(s >= cast<int64_t>(0ULL)),branchT);
break;}
case cast<uint32_t>(8ULL):{
r5900_CPU_trapIf(c,(s >= cast<int64_t>(simm)));
break;}
case cast<uint32_t>(9ULL):{
r5900_CPU_trapIf(c,(c->R[rs].Lo >= simm));
break;}
case cast<uint32_t>(10ULL):{
r5900_CPU_trapIf(c,(s < cast<int64_t>(simm)));
break;}
case cast<uint32_t>(11ULL):{
r5900_CPU_trapIf(c,(c->R[rs].Lo < simm));
break;}
case cast<uint32_t>(12ULL):{
r5900_CPU_trapIf(c,(c->R[rs].Lo == simm));
break;}
case cast<uint32_t>(14ULL):{
r5900_CPU_trapIf(c,(c->R[rs].Lo != simm));
break;}
case cast<uint32_t>(16ULL):{
r5900_CPU_set(c,cast<uint32_t>(31ULL),cast<uint64_t>((c->curPC + cast<uint64_t>(8ULL))));
r5900_CPU_doBranch(c,(s < cast<int64_t>(0ULL)),branchT);
break;}
case cast<uint32_t>(17ULL):{
r5900_CPU_set(c,cast<uint32_t>(31ULL),cast<uint64_t>((c->curPC + cast<uint64_t>(8ULL))));
r5900_CPU_doBranch(c,(s >= cast<int64_t>(0ULL)),branchT);
break;}
case cast<uint32_t>(18ULL):{
r5900_CPU_set(c,cast<uint32_t>(31ULL),cast<uint64_t>((c->curPC + cast<uint64_t>(8ULL))));
r5900_CPU_doBranchLikely(c,(s < cast<int64_t>(0ULL)),branchT);
break;}
case cast<uint32_t>(19ULL):{
r5900_CPU_set(c,cast<uint32_t>(31ULL),cast<uint64_t>((c->curPC + cast<uint64_t>(8ULL))));
r5900_CPU_doBranchLikely(c,(s >= cast<int64_t>(0ULL)),branchT);
break;}
case cast<uint32_t>(24ULL):{
c->SA = cast<uint32_t>(((cast<uint32_t>(((cast<uint32_t>((cast<uint32_t>(c->R[rs].Lo) & cast<uint32_t>(15ULL)))) ^ (cast<uint32_t>((cast<uint32_t>(simm) & cast<uint32_t>(15ULL))))))) * cast<uint32_t>(8ULL)));
break;}
case cast<uint32_t>(25ULL):{
c->SA = cast<uint32_t>(((cast<uint32_t>(((cast<uint32_t>((cast<uint32_t>(c->R[rs].Lo) & cast<uint32_t>(7ULL)))) ^ (cast<uint32_t>((cast<uint32_t>(simm) & cast<uint32_t>(7ULL))))))) * cast<uint32_t>(16ULL)));
break;}
default:{
r5900_CPU_Exception(c,cast<uint32_t>(10ULL));
break;}
}}
}
}
// tools/cpu/r5900/exec.go:518:1
void r5900_CPU_trapIf(r5900_CPU* c,bool cond){
{
if (cond) {
r5900_CPU_Exception(c,cast<uint32_t>(13ULL));
}
}
}
// tools/cpu/r5900/exec.go:532:1
void r5900_CPU_loadOp(r5900_CPU* c,uint32_t op,uint32_t rs,uint32_t rt,uint64_t simm){
{
uint64_t vaddr = cast<uint64_t>((c->R[rs].Lo + simm));
auto align = [&](uint64_t n)->std::tuple<uint32_t,bool>{
if ((cast<uint64_t>((vaddr & (cast<uint64_t>((n - cast<uint64_t>(1ULL)))))) != cast<uint64_t>(0ULL))) {
r5900_CPU_addrError(c,cast<uint32_t>(4ULL),vaddr);
return {cast<uint32_t>(0ULL),false};
}
return r5900_CPU_Translate(c,vaddr,false);
}
;
{
switch(op){
case cast<uint32_t>(32ULL):{
auto tmp54 = r5900_CPU_Translate(c,vaddr,false);
uint32_t p = std::get<0>(tmp54);
bool ok = std::get<1>(tmp54);
if ((!ok)) {
return ;
}
r5900_CPU_set(c,rt,cast<uint64_t>(cast<int64_t>(cast<int8_t>(cast<uint8_t>(r5900_CPU_read8(c,p))))));
break;}
case cast<uint32_t>(36ULL):{
auto tmp55 = r5900_CPU_Translate(c,vaddr,false);
uint32_t p = std::get<0>(tmp55);
bool ok = std::get<1>(tmp55);
if ((!ok)) {
return ;
}
r5900_CPU_set(c,rt,cast<uint64_t>(r5900_CPU_read8(c,p)));
break;}
case cast<uint32_t>(33ULL):{
auto tmp56 = align(cast<uint64_t>(2ULL));
uint32_t p = std::get<0>(tmp56);
bool ok = std::get<1>(tmp56);
if ((!ok)) {
return ;
}
r5900_CPU_set(c,rt,cast<uint64_t>(cast<int64_t>(cast<int16_t>(cast<uint16_t>(r5900_CPU_read16(c,p))))));
break;}
case cast<uint32_t>(37ULL):{
auto tmp57 = align(cast<uint64_t>(2ULL));
uint32_t p = std::get<0>(tmp57);
bool ok = std::get<1>(tmp57);
if ((!ok)) {
return ;
}
r5900_CPU_set(c,rt,cast<uint64_t>(r5900_CPU_read16(c,p)));
break;}
case cast<uint32_t>(35ULL):{
auto tmp58 = align(cast<uint64_t>(4ULL));
uint32_t p = std::get<0>(tmp58);
bool ok = std::get<1>(tmp58);
if ((!ok)) {
return ;
}
r5900_CPU_set(c,rt,r5900_sext32(r5900_CPU_read32(c,p)));
break;}
case cast<uint32_t>(39ULL):{
auto tmp59 = align(cast<uint64_t>(4ULL));
uint32_t p = std::get<0>(tmp59);
bool ok = std::get<1>(tmp59);
if ((!ok)) {
return ;
}
r5900_CPU_set(c,rt,cast<uint64_t>(r5900_CPU_read32(c,p)));
break;}
case cast<uint32_t>(55ULL):{
auto tmp60 = align(cast<uint64_t>(8ULL));
uint32_t p = std::get<0>(tmp60);
bool ok = std::get<1>(tmp60);
if ((!ok)) {
return ;
}
r5900_CPU_set(c,rt,r5900_CPU_read64(c,p));
break;}
case cast<uint32_t>(30ULL):{
auto tmp61 = r5900_CPU_Translate(c,(vaddr & ~(cast<uint64_t>(15ULL))),false);
uint32_t p = std::get<0>(tmp61);
bool ok = std::get<1>(tmp61);
if ((!ok)) {
return ;
}
r5900_Quad q = r5900_CPU_read128(c,p);
r5900_CPU_setQ(c,rt,q.Lo,q.Hi);
break;}
case cast<uint32_t>(48ULL):{
auto tmp62 = align(cast<uint64_t>(4ULL));
uint32_t p = std::get<0>(tmp62);
bool ok = std::get<1>(tmp62);
if ((!ok)) {
return ;
}
r5900_CPU_set(c,rt,r5900_sext32(r5900_CPU_read32(c,p)));
c->LLBit = true;
break;}
case cast<uint32_t>(52ULL):{
auto tmp63 = align(cast<uint64_t>(8ULL));
uint32_t p = std::get<0>(tmp63);
bool ok = std::get<1>(tmp63);
if ((!ok)) {
return ;
}
r5900_CPU_set(c,rt,r5900_CPU_read64(c,p));
c->LLBit = true;
break;}
case cast<uint32_t>(34ULL):{
auto tmp64 = r5900_CPU_Translate(c,(vaddr & ~(cast<uint64_t>(3ULL))),false);
uint32_t p = std::get<0>(tmp64);
bool ok = std::get<1>(tmp64);
if ((!ok)) {
return ;
}
uint64_t shift = cast<uint64_t>(((cast<uint64_t>((cast<uint64_t>(3ULL) - cast<uint64_t>((vaddr & cast<uint64_t>(3ULL)))))) * cast<uint64_t>(8ULL)));
uint32_t word = r5900_CPU_read32(c,p);
uint32_t cur = cast<uint32_t>(c->R[rt].Lo);
r5900_CPU_set(c,rt,r5900_sext32(cast<uint32_t>(((cast<uint32_t>((cur & (cast<uint32_t>((shl<uint32_t>(cast<uint32_t>(1ULL),shift) - cast<uint32_t>(1ULL))))))) | (shl<uint32_t>(word,shift))))));
break;}
case cast<uint32_t>(38ULL):{
auto tmp65 = r5900_CPU_Translate(c,(vaddr & ~(cast<uint64_t>(3ULL))),false);
uint32_t p = std::get<0>(tmp65);
bool ok = std::get<1>(tmp65);
if ((!ok)) {
return ;
}
uint64_t shift = cast<uint64_t>(((cast<uint64_t>((vaddr & cast<uint64_t>(3ULL)))) * cast<uint64_t>(8ULL)));
uint32_t word = r5900_CPU_read32(c,p);
uint32_t cur = cast<uint32_t>(c->R[rt].Lo);
r5900_CPU_set(c,rt,r5900_sext32(cast<uint32_t>((((cur & ~((shr<uint32_t>(cast<uint32_t>(4294967295ULL),shift))))) | (shr<uint32_t>(word,shift))))));
break;}
case cast<uint32_t>(26ULL):{
auto tmp66 = r5900_CPU_Translate(c,(vaddr & ~(cast<uint64_t>(7ULL))),false);
uint32_t p = std::get<0>(tmp66);
bool ok = std::get<1>(tmp66);
if ((!ok)) {
return ;
}
uint64_t shift = cast<uint64_t>(((cast<uint64_t>((cast<uint64_t>(7ULL) - cast<uint64_t>((vaddr & cast<uint64_t>(7ULL)))))) * cast<uint64_t>(8ULL)));
uint64_t d = r5900_CPU_read64(c,p);
r5900_CPU_set(c,rt,cast<uint64_t>(((cast<uint64_t>((c->R[rt].Lo & (cast<uint64_t>((shl<uint64_t>(cast<uint64_t>(1ULL),shift) - cast<uint64_t>(1ULL))))))) | (shl<uint64_t>(d,shift)))));
break;}
case cast<uint32_t>(27ULL):{
auto tmp67 = r5900_CPU_Translate(c,(vaddr & ~(cast<uint64_t>(7ULL))),false);
uint32_t p = std::get<0>(tmp67);
bool ok = std::get<1>(tmp67);
if ((!ok)) {
return ;
}
uint64_t shift = cast<uint64_t>(((cast<uint64_t>((vaddr & cast<uint64_t>(7ULL)))) * cast<uint64_t>(8ULL)));
uint64_t d = r5900_CPU_read64(c,p);
r5900_CPU_set(c,rt,cast<uint64_t>((((c->R[rt].Lo & ~((shr<uint64_t>(cast<uint64_t>(18446744073709551615ULL),shift))))) | (shr<uint64_t>(d,shift)))));
break;}
}}
}
}
// tools/cpu/r5900/exec.go:648:1
void r5900_CPU_storeOp(r5900_CPU* c,uint32_t op,uint32_t rs,uint32_t rt,uint64_t simm){
{
uint64_t vaddr = cast<uint64_t>((c->R[rs].Lo + simm));
uint64_t v = c->R[rt].Lo;
auto align = [&](uint64_t n)->std::tuple<uint32_t,bool>{
if ((cast<uint64_t>((vaddr & (cast<uint64_t>((n - cast<uint64_t>(1ULL)))))) != cast<uint64_t>(0ULL))) {
r5900_CPU_addrError(c,cast<uint32_t>(5ULL),vaddr);
return {cast<uint32_t>(0ULL),false};
}
return r5900_CPU_Translate(c,vaddr,true);
}
;
{
switch(op){
case cast<uint32_t>(40ULL):{
auto tmp68 = r5900_CPU_Translate(c,vaddr,true);
uint32_t p = std::get<0>(tmp68);
bool ok = std::get<1>(tmp68);
if ((!ok)) {
return ;
}
r5900_CPU_write8(c,p,cast<uint32_t>((cast<uint32_t>(v) & cast<uint32_t>(255ULL))));
break;}
case cast<uint32_t>(41ULL):{
auto tmp69 = align(cast<uint64_t>(2ULL));
uint32_t p = std::get<0>(tmp69);
bool ok = std::get<1>(tmp69);
if ((!ok)) {
return ;
}
r5900_CPU_write16(c,p,cast<uint32_t>((cast<uint32_t>(v) & cast<uint32_t>(65535ULL))));
break;}
case cast<uint32_t>(43ULL):{
auto tmp70 = align(cast<uint64_t>(4ULL));
uint32_t p = std::get<0>(tmp70);
bool ok = std::get<1>(tmp70);
if ((!ok)) {
return ;
}
r5900_CPU_write32(c,p,cast<uint32_t>(v));
break;}
case cast<uint32_t>(63ULL):{
auto tmp71 = align(cast<uint64_t>(8ULL));
uint32_t p = std::get<0>(tmp71);
bool ok = std::get<1>(tmp71);
if ((!ok)) {
return ;
}
r5900_CPU_write64(c,p,v);
break;}
case cast<uint32_t>(31ULL):{
auto tmp72 = r5900_CPU_Translate(c,(vaddr & ~(cast<uint64_t>(15ULL))),true);
uint32_t p = std::get<0>(tmp72);
bool ok = std::get<1>(tmp72);
if ((!ok)) {
return ;
}
r5900_CPU_write128(c,p,c->R[rt]);
break;}
case cast<uint32_t>(56ULL):{
auto tmp73 = align(cast<uint64_t>(4ULL));
uint32_t p = std::get<0>(tmp73);
bool ok = std::get<1>(tmp73);
if ((!ok)) {
return ;
}
if (c->LLBit) {
r5900_CPU_write32(c,p,cast<uint32_t>(v));
}
r5900_CPU_set(c,rt,r5900_b2u(c->LLBit));
c->LLBit = false;
break;}
case cast<uint32_t>(60ULL):{
auto tmp74 = align(cast<uint64_t>(8ULL));
uint32_t p = std::get<0>(tmp74);
bool ok = std::get<1>(tmp74);
if ((!ok)) {
return ;
}
if (c->LLBit) {
r5900_CPU_write64(c,p,v);
}
r5900_CPU_set(c,rt,r5900_b2u(c->LLBit));
c->LLBit = false;
break;}
case cast<uint32_t>(42ULL):{
auto tmp75 = r5900_CPU_Translate(c,(vaddr & ~(cast<uint64_t>(3ULL))),true);
uint32_t p = std::get<0>(tmp75);
bool ok = std::get<1>(tmp75);
if ((!ok)) {
return ;
}
uint64_t shift = cast<uint64_t>(((cast<uint64_t>((cast<uint64_t>(3ULL) - cast<uint64_t>((vaddr & cast<uint64_t>(3ULL)))))) * cast<uint64_t>(8ULL)));
uint32_t word = r5900_CPU_read32(c,p);
r5900_CPU_write32(c,p,cast<uint32_t>((((word & ~((shr<uint32_t>(cast<uint32_t>(4294967295ULL),shift))))) | (shr<uint32_t>(cast<uint32_t>(v),shift)))));
break;}
case cast<uint32_t>(46ULL):{
auto tmp76 = r5900_CPU_Translate(c,(vaddr & ~(cast<uint64_t>(3ULL))),true);
uint32_t p = std::get<0>(tmp76);
bool ok = std::get<1>(tmp76);
if ((!ok)) {
return ;
}
uint64_t shift = cast<uint64_t>(((cast<uint64_t>((vaddr & cast<uint64_t>(3ULL)))) * cast<uint64_t>(8ULL)));
uint32_t word = r5900_CPU_read32(c,p);
r5900_CPU_write32(c,p,cast<uint32_t>((((word & ~((shl<uint32_t>(cast<uint32_t>(4294967295ULL),shift))))) | (shl<uint32_t>(cast<uint32_t>(v),shift)))));
break;}
case cast<uint32_t>(44ULL):{
auto tmp77 = r5900_CPU_Translate(c,(vaddr & ~(cast<uint64_t>(7ULL))),true);
uint32_t p = std::get<0>(tmp77);
bool ok = std::get<1>(tmp77);
if ((!ok)) {
return ;
}
uint64_t shift = cast<uint64_t>(((cast<uint64_t>((cast<uint64_t>(7ULL) - cast<uint64_t>((vaddr & cast<uint64_t>(7ULL)))))) * cast<uint64_t>(8ULL)));
r5900_CPU_write64(c,p,cast<uint64_t>((((r5900_CPU_read64(c,p) & ~((shr<uint64_t>(cast<uint64_t>(18446744073709551615ULL),shift))))) | (shr<uint64_t>(v,shift)))));
break;}
case cast<uint32_t>(45ULL):{
auto tmp78 = r5900_CPU_Translate(c,(vaddr & ~(cast<uint64_t>(7ULL))),true);
uint32_t p = std::get<0>(tmp78);
bool ok = std::get<1>(tmp78);
if ((!ok)) {
return ;
}
uint64_t shift = cast<uint64_t>(((cast<uint64_t>((vaddr & cast<uint64_t>(7ULL)))) * cast<uint64_t>(8ULL)));
r5900_CPU_write64(c,p,cast<uint64_t>((((r5900_CPU_read64(c,p) & ~((shl<uint64_t>(cast<uint64_t>(18446744073709551615ULL),shift))))) | (shl<uint64_t>(v,shift)))));
break;}
}}
}
}
// tools/cpu/r5900/exec.go:750:1
std::tuple<uint32_t,bool> r5900_addOv32(uint32_t a,uint32_t b){
{
uint32_t r = cast<uint32_t>((a + b));
return {r,(cast<uint32_t>((cast<uint32_t>(((cast<uint32_t>((a ^ r))) & (cast<uint32_t>((b ^ r))))) & cast<uint32_t>(2147483648ULL))) != cast<uint32_t>(0ULL))};
}
}
// tools/cpu/r5900/exec.go:754:1
std::tuple<uint32_t,bool> r5900_subOv32(uint32_t a,uint32_t b){
{
uint32_t r = cast<uint32_t>((a - b));
return {r,(cast<uint32_t>((cast<uint32_t>(((cast<uint32_t>((a ^ b))) & (cast<uint32_t>((a ^ r))))) & cast<uint32_t>(2147483648ULL))) != cast<uint32_t>(0ULL))};
}
}
// tools/cpu/r5900/exec.go:758:1
std::tuple<uint64_t,bool> r5900_addOv64(uint64_t a,uint64_t b){
{
uint64_t r = cast<uint64_t>((a + b));
return {r,(cast<uint64_t>((cast<uint64_t>(((cast<uint64_t>((a ^ r))) & (cast<uint64_t>((b ^ r))))) & cast<uint64_t>(9223372036854775808ULL))) != cast<uint64_t>(0ULL))};
}
}
// tools/cpu/r5900/exec.go:762:1
std::tuple<uint64_t,bool> r5900_subOv64(uint64_t a,uint64_t b){
{
uint64_t r = cast<uint64_t>((a - b));
return {r,(cast<uint64_t>((cast<uint64_t>(((cast<uint64_t>((a ^ b))) & (cast<uint64_t>((a ^ r))))) & cast<uint64_t>(9223372036854775808ULL))) != cast<uint64_t>(0ULL))};
}
}
// tools/cpu/r5900/exec.go:767:1
uint64_t r5900_b2u(bool b){
{
if (b) {
return cast<uint64_t>(1ULL);
}
return cast<uint64_t>(0ULL);
}
}
// tools/cpu/r5900/fpu.go:41:1
double r5900_fdec(uint32_t b){
{
int64_t exp = cast<int64_t>(cast<uint32_t>(((shr<uint32_t>(b,cast<int64_t>(23ULL))) & cast<uint32_t>(255ULL))));
uint32_t man = cast<uint32_t>((b & cast<uint32_t>(8388607ULL)));
double sign = 1.0;
if ((shr<uint32_t>(b,cast<int64_t>(31ULL)) != cast<uint32_t>(0ULL))) {
sign = cast<double>(-1.0);
}
if ((exp == cast<int64_t>(0ULL))) {
return (sign * cast<double>(0ULL));
}
return (sign * go_math_Ldexp((cast<double>(1ULL) + (cast<double>(man) / (cast<double>(8388608ULL)))),cast<int64_t>((exp - cast<int64_t>(127ULL)))));
}
}
// tools/cpu/r5900/fpu.go:57:1
uint32_t r5900_fenc(double v){
{
uint32_t sign={};
if (go_math_Signbit(v)) {
sign = cast<uint32_t>(2147483648ULL);
}
if (go_math_IsNaN(v)) {
return sign;
}
double a = go_math_Abs(v);
if (go_math_IsInf(a,cast<int64_t>(0ULL))) {
return cast<uint32_t>((sign | cast<uint32_t>(2147483647ULL)));
}
if ((a == cast<double>(0ULL))) {
return sign;
}
auto tmp79 = go_math_Frexp(a);
double frac = std::get<0>(tmp79);
int64_t exp = std::get<1>(tmp79);
int64_t e = cast<int64_t>((exp - cast<int64_t>(1ULL)));
if ((e > cast<int64_t>(128ULL))) {
return cast<uint32_t>((sign | cast<uint32_t>(2147483647ULL)));
}
if ((e < cast<int64_t>(-127ULL))) {
return sign;
}
uint32_t man = cast<uint32_t>(((((frac * cast<double>(2ULL)) - cast<double>(1ULL))) * (cast<double>(8388608ULL))));
return cast<uint32_t>((cast<uint32_t>((sign | shl<uint32_t>(cast<uint32_t>(cast<int64_t>((e + cast<int64_t>(127ULL)))),cast<int64_t>(23ULL)))) | cast<uint32_t>((man & cast<uint32_t>(8388607ULL)))));
}
}
// tools/cpu/r5900/fpu.go:89:1
double r5900_fdiv(double a,double b){
{
if ((b == cast<double>(0ULL))) {
double s = 1.0;
if ((go_math_Signbit(a) != go_math_Signbit(b))) {
s = cast<double>(-1.0);
}
return (s * r5900_fdec(cast<uint32_t>(2147483647ULL)));
}
return (a / b);
}
}
// tools/cpu/r5900/fpu.go:102:1
double r5900_CPU_f(r5900_CPU* c,uint32_t i){
{
return r5900_fdec(c->FPR[i]);
}
}
// tools/cpu/r5900/fpu.go:103:1
void r5900_CPU_setF(r5900_CPU* c,uint32_t i,double v){
{
c->FPR[i] = r5900_fenc(v);
}
}
// tools/cpu/r5900/fpu.go:106:1
void r5900_CPU_cop1(r5900_CPU* c,uint32_t w,uint32_t rs,uint32_t rt,uint32_t rd,uint32_t shamt,uint64_t branchT){
{
{
switch(rs){
case cast<uint32_t>(0ULL):{
r5900_CPU_set(c,rt,r5900_sext32(c->FPR[rd]));
break;}
case cast<uint32_t>(2ULL):{
if ((rd == cast<uint32_t>(31ULL))) {
r5900_CPU_set(c,rt,r5900_sext32(c->FCR31));
}
else {
r5900_CPU_set(c,rt,r5900_sext32(cast<uint32_t>(11776ULL)));
}
break;}
case cast<uint32_t>(4ULL):{
c->FPR[rd] = cast<uint32_t>(c->R[rt].Lo);
break;}
case cast<uint32_t>(6ULL):{
if ((rd == cast<uint32_t>(31ULL))) {
c->FCR31 = cast<uint32_t>(c->R[rt].Lo);
}
break;}
case cast<uint32_t>(8ULL):{
bool cond = (cast<uint32_t>((c->FCR31 & cast<uint32_t>(8388608ULL))) != cast<uint32_t>(0ULL));
{
switch(cast<uint32_t>((rt & cast<uint32_t>(3ULL)))){
case cast<uint32_t>(0ULL):{
r5900_CPU_doBranch(c,(!cond),branchT);
break;}
case cast<uint32_t>(1ULL):{
r5900_CPU_doBranch(c,cond,branchT);
break;}
case cast<uint32_t>(2ULL):{
r5900_CPU_doBranchLikely(c,(!cond),branchT);
break;}
case cast<uint32_t>(3ULL):{
r5900_CPU_doBranchLikely(c,cond,branchT);
break;}
}}
break;}
case cast<uint32_t>(16ULL):{
r5900_CPU_cop1Single(c,w,rt,rd,shamt);
break;}
case cast<uint32_t>(20ULL):{
if ((cast<uint32_t>((w & cast<uint32_t>(63ULL))) == cast<uint32_t>(32ULL))) {
r5900_CPU_setF(c,shamt,cast<double>(cast<int32_t>(c->FPR[rd])));
return ;
}
r5900_CPU_Exception(c,cast<uint32_t>(10ULL));
break;}
default:{
r5900_CPU_Exception(c,cast<uint32_t>(10ULL));
break;}
}}
}
}
// tools/cpu/r5900/fpu.go:152:1
void r5900_CPU_cop1Single(r5900_CPU* c,uint32_t w,uint32_t ft,uint32_t fs,uint32_t fd){
{
auto tmp80 = std::make_tuple(r5900_CPU_f(c,fs),r5900_CPU_f(c,ft));
double s = std::get<0>(tmp80);
double t = std::get<1>(tmp80);
{
switch(cast<uint32_t>((w & cast<uint32_t>(63ULL)))){
case cast<uint32_t>(0ULL):{
r5900_CPU_setF(c,fd,(s + t));
break;}
case cast<uint32_t>(1ULL):{
r5900_CPU_setF(c,fd,(s - t));
break;}
case cast<uint32_t>(2ULL):{
r5900_CPU_setF(c,fd,(s * t));
break;}
case cast<uint32_t>(3ULL):{
r5900_CPU_setF(c,fd,r5900_fdiv(s,t));
break;}
case cast<uint32_t>(4ULL):{
r5900_CPU_setF(c,fd,go_math_Sqrt(go_math_Abs(t)));
break;}
case cast<uint32_t>(5ULL):{
r5900_CPU_setF(c,fd,go_math_Abs(s));
break;}
case cast<uint32_t>(6ULL):{
c->FPR[fd] = c->FPR[fs];
break;}
case cast<uint32_t>(7ULL):{
c->FPR[fd] = cast<uint32_t>((c->FPR[fs] ^ cast<uint32_t>(2147483648ULL)));
break;}
case cast<uint32_t>(22ULL):{
r5900_CPU_setF(c,fd,r5900_fdiv(s,go_math_Sqrt(go_math_Abs(t))));
break;}
case cast<uint32_t>(24ULL):{
c->ACC = r5900_fenc((s + t));
break;}
case cast<uint32_t>(25ULL):{
c->ACC = r5900_fenc((s - t));
break;}
case cast<uint32_t>(26ULL):{
c->ACC = r5900_fenc((s * t));
break;}
case cast<uint32_t>(28ULL):{
r5900_CPU_setF(c,fd,(r5900_fdec(c->ACC) + (s * t)));
break;}
case cast<uint32_t>(29ULL):{
r5900_CPU_setF(c,fd,(r5900_fdec(c->ACC) - (s * t)));
break;}
case cast<uint32_t>(30ULL):{
c->ACC = r5900_fenc((r5900_fdec(c->ACC) + (s * t)));
break;}
case cast<uint32_t>(31ULL):{
c->ACC = r5900_fenc((r5900_fdec(c->ACC) - (s * t)));
break;}
case cast<uint32_t>(36ULL):{
c->FPR[fd] = r5900_cvtW(s);
break;}
case cast<uint32_t>(40ULL):{
r5900_CPU_setF(c,fd,go_math_Max(s,t));
break;}
case cast<uint32_t>(41ULL):{
r5900_CPU_setF(c,fd,go_math_Min(s,t));
break;}
case cast<uint32_t>(48ULL):{
r5900_CPU_setCond(c,false);
break;}
case cast<uint32_t>(50ULL):{
r5900_CPU_setCond(c,(s == t));
break;}
case cast<uint32_t>(52ULL):{
r5900_CPU_setCond(c,(s < t));
break;}
case cast<uint32_t>(54ULL):{
r5900_CPU_setCond(c,(s <= t));
break;}
default:{
r5900_CPU_Exception(c,cast<uint32_t>(10ULL));
break;}
}}
}
}
// tools/cpu/r5900/fpu.go:214:1
void r5900_CPU_setCond(r5900_CPU* c,bool b){
{
if (b) {
c->FCR31 |= cast<uint32_t>(8388608ULL);
}
else {
c->FCR31 &= ~(cast<uint32_t>(8388608ULL));
}
}
}
// tools/cpu/r5900/fpu.go:224:1
uint32_t r5900_cvtW(double v){
{
double t = go_math_Trunc(v);
if ((t >= cast<double>(2147483647ULL))) {
return cast<uint32_t>(2147483647ULL);
}
if ((t <= cast<double>(-cast<int64_t>(2147483648ULL)))) {
return cast<uint32_t>(2147483648ULL);
}
return cast<uint32_t>(cast<int32_t>(t));
}
}
// tools/cpu/r5900/fpu.go:236:1
void r5900_CPU_cop1Mem(r5900_CPU* c,uint32_t op,uint32_t rs,uint32_t rt,uint64_t simm){
{
uint64_t vaddr = cast<uint64_t>((c->R[rs].Lo + simm));
if ((cast<uint64_t>((vaddr & cast<uint64_t>(3ULL))) != cast<uint64_t>(0ULL))) {
uint32_t code = cast<uint32_t>(4ULL);
if ((op == cast<uint32_t>(57ULL))) {
code = cast<uint32_t>(5ULL);
}
r5900_CPU_addrError(c,code,vaddr);
return ;
}
{
switch(op){
case cast<uint32_t>(49ULL):{
auto tmp81 = r5900_CPU_Translate(c,vaddr,false);
uint32_t p = std::get<0>(tmp81);
bool ok = std::get<1>(tmp81);
if ((!ok)) {
return ;
}
c->FPR[rt] = r5900_CPU_read32(c,p);
break;}
case cast<uint32_t>(57ULL):{
auto tmp82 = r5900_CPU_Translate(c,vaddr,true);
uint32_t p = std::get<0>(tmp82);
bool ok = std::get<1>(tmp82);
if ((!ok)) {
return ;
}
r5900_CPU_write32(c,p,c->FPR[rt]);
break;}
}}
}
}
// tools/cpu/r5900/mmi.go:35:1
std::array<uint32_t,4> r5900_words(r5900_Quad q){
{
return std::array<uint32_t,4>{cast<uint32_t>(q.Lo),cast<uint32_t>(shr<uint64_t>(q.Lo,cast<int64_t>(32ULL))),cast<uint32_t>(q.Hi),cast<uint32_t>(shr<uint64_t>(q.Hi,cast<int64_t>(32ULL)))};
}
}
// tools/cpu/r5900/mmi.go:39:1
r5900_Quad r5900_fromWords(std::array<uint32_t,4> w){
{
return r5900_Quad{cast<uint64_t>((cast<uint64_t>(w[cast<int64_t>(0ULL)]) | shl<uint64_t>(cast<uint64_t>(w[cast<int64_t>(1ULL)]),cast<int64_t>(32ULL)))),cast<uint64_t>((cast<uint64_t>(w[cast<int64_t>(2ULL)]) | shl<uint64_t>(cast<uint64_t>(w[cast<int64_t>(3ULL)]),cast<int64_t>(32ULL))))};
}
}
// tools/cpu/r5900/mmi.go:46:1
std::array<uint16_t,8> r5900_halves(r5900_Quad q){
{
std::array<uint16_t,8> h={};
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(4ULL));i++){
h[i] = cast<uint16_t>(shr<uint64_t>(q.Lo,(cast<uint64_t>((cast<uint64_t>(16ULL) * cast<uint64_t>(i))))));
h[cast<int64_t>((i + cast<int64_t>(4ULL)))] = cast<uint16_t>(shr<uint64_t>(q.Hi,(cast<uint64_t>((cast<uint64_t>(16ULL) * cast<uint64_t>(i))))));
}
}return h;
}
}
// tools/cpu/r5900/mmi.go:55:1
r5900_Quad r5900_fromHalves(std::array<uint16_t,8> h){
{
r5900_Quad q={};
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(4ULL));i++){
q.Lo |= shl<uint64_t>(cast<uint64_t>(h[i]),(cast<uint64_t>((cast<uint64_t>(16ULL) * cast<uint64_t>(i)))));
q.Hi |= shl<uint64_t>(cast<uint64_t>(h[cast<int64_t>((i + cast<int64_t>(4ULL)))]),(cast<uint64_t>((cast<uint64_t>(16ULL) * cast<uint64_t>(i)))));
}
}return q;
}
}
// tools/cpu/r5900/mmi.go:64:1
std::array<uint8_t,16> r5900_octets(r5900_Quad q){
{
std::array<uint8_t,16> b={};
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(8ULL));i++){
b[i] = cast<uint8_t>(shr<uint64_t>(q.Lo,(cast<uint64_t>((cast<uint64_t>(8ULL) * cast<uint64_t>(i))))));
b[cast<int64_t>((i + cast<int64_t>(8ULL)))] = cast<uint8_t>(shr<uint64_t>(q.Hi,(cast<uint64_t>((cast<uint64_t>(8ULL) * cast<uint64_t>(i))))));
}
}return b;
}
}
// tools/cpu/r5900/mmi.go:73:1
r5900_Quad r5900_fromOctets(std::array<uint8_t,16> b){
{
r5900_Quad q={};
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(8ULL));i++){
q.Lo |= shl<uint64_t>(cast<uint64_t>(b[i]),(cast<uint64_t>((cast<uint64_t>(8ULL) * cast<uint64_t>(i)))));
q.Hi |= shl<uint64_t>(cast<uint64_t>(b[cast<int64_t>((i + cast<int64_t>(8ULL)))]),(cast<uint64_t>((cast<uint64_t>(8ULL) * cast<uint64_t>(i)))));
}
}return q;
}
}
// tools/cpu/r5900/mmi.go:84:1
uint32_t r5900_satS32(int64_t v){
{
if ((v > cast<int64_t>(2147483647ULL))) {
return cast<uint32_t>(2147483647ULL);
}
if ((v < cast<int64_t>(-2147483648ULL))) {
return cast<uint32_t>(2147483648ULL);
}
return cast<uint32_t>(cast<int32_t>(v));
}
}
// tools/cpu/r5900/mmi.go:94:1
uint16_t r5900_satS16(int32_t v){
{
if ((v > cast<int32_t>(32767ULL))) {
return cast<uint16_t>(32767ULL);
}
if ((v < cast<int32_t>(-32768ULL))) {
return cast<uint16_t>(32768ULL);
}
return cast<uint16_t>(cast<int16_t>(v));
}
}
// tools/cpu/r5900/mmi.go:104:1
uint8_t r5900_satS8(int32_t v){
{
if ((v > cast<int32_t>(127ULL))) {
return cast<uint8_t>(127ULL);
}
if ((v < cast<int32_t>(-128ULL))) {
return cast<uint8_t>(128ULL);
}
return cast<uint8_t>(cast<int8_t>(v));
}
}
// tools/cpu/r5900/mmi.go:114:1
uint32_t r5900_satU32(int64_t v){
{
if ((v > cast<int64_t>(4294967295ULL))) {
return cast<uint32_t>(4294967295ULL);
}
if ((v < cast<int64_t>(0ULL))) {
return cast<uint32_t>(0ULL);
}
return cast<uint32_t>(v);
}
}
// tools/cpu/r5900/mmi.go:124:1
uint16_t r5900_satU16(int32_t v){
{
if ((v > cast<int32_t>(65535ULL))) {
return cast<uint16_t>(65535ULL);
}
if ((v < cast<int32_t>(0ULL))) {
return cast<uint16_t>(0ULL);
}
return cast<uint16_t>(v);
}
}
// tools/cpu/r5900/mmi.go:134:1
uint8_t r5900_satU8(int32_t v){
{
if ((v > cast<int32_t>(255ULL))) {
return cast<uint8_t>(255ULL);
}
if ((v < cast<int32_t>(0ULL))) {
return cast<uint8_t>(0ULL);
}
return cast<uint8_t>(v);
}
}
// tools/cpu/r5900/mmi.go:149:1
r5900_Quad r5900_CPU_loQ(r5900_CPU* c){
{
return r5900_Quad{c->LO,c->LO1};
}
}
// tools/cpu/r5900/mmi.go:150:1
r5900_Quad r5900_CPU_hiQ(r5900_CPU* c){
{
return r5900_Quad{c->HI,c->HI1};
}
}
// tools/cpu/r5900/mmi.go:152:1
void r5900_CPU_setLoQ(r5900_CPU* c,r5900_Quad q){
{
auto tmp83 = std::make_tuple(q.Lo,q.Hi);
c->LO = std::get<0>(tmp83);
c->LO1 = std::get<1>(tmp83);
}
}
// tools/cpu/r5900/mmi.go:153:1
void r5900_CPU_setHiQ(r5900_CPU* c,r5900_Quad q){
{
auto tmp84 = std::make_tuple(q.Lo,q.Hi);
c->HI = std::get<0>(tmp84);
c->HI1 = std::get<1>(tmp84);
}
}
// tools/cpu/r5900/mmi.go:157:1
void r5900_CPU_mmi(r5900_CPU* c,uint32_t w,uint32_t rs,uint32_t rt,uint32_t rd,uint32_t shamt){
{
{
switch(cast<uint32_t>((w & cast<uint32_t>(63ULL)))){
case cast<uint32_t>(0ULL):{
r5900_CPU_maddAcc(c,rd,cast<int64_t>((cast<int64_t>(cast<int32_t>(cast<uint32_t>(c->R[rs].Lo))) * cast<int64_t>(cast<int32_t>(cast<uint32_t>(c->R[rt].Lo))))),false);
break;}
case cast<uint32_t>(1ULL):{
r5900_CPU_maddAcc(c,rd,cast<int64_t>(cast<uint64_t>((cast<uint64_t>(cast<uint32_t>(c->R[rs].Lo)) * cast<uint64_t>(cast<uint32_t>(c->R[rt].Lo))))),false);
break;}
case cast<uint32_t>(4ULL):{
std::array<uint32_t,4> s = r5900_words(c->R[rs]);
r5900_CPU_set(c,rd,cast<uint64_t>((cast<uint64_t>(r5900_lzcw(s[cast<int64_t>(0ULL)])) | shl<uint64_t>(cast<uint64_t>(r5900_lzcw(s[cast<int64_t>(1ULL)])),cast<int64_t>(32ULL)))));
break;}
case cast<uint32_t>(8ULL):{
r5900_CPU_mmi0(c,rs,rt,rd,shamt);
break;}
case cast<uint32_t>(9ULL):{
r5900_CPU_mmi2(c,rs,rt,rd,shamt);
break;}
case cast<uint32_t>(16ULL):{
r5900_CPU_set(c,rd,c->HI1);
break;}
case cast<uint32_t>(17ULL):{
c->HI1 = c->R[rs].Lo;
break;}
case cast<uint32_t>(18ULL):{
r5900_CPU_set(c,rd,c->LO1);
break;}
case cast<uint32_t>(19ULL):{
c->LO1 = c->R[rs].Lo;
break;}
case cast<uint32_t>(24ULL):{
int64_t p = cast<int64_t>((cast<int64_t>(cast<int32_t>(cast<uint32_t>(c->R[rs].Lo))) * cast<int64_t>(cast<int32_t>(cast<uint32_t>(c->R[rt].Lo)))));
auto tmp85 = std::make_tuple(r5900_sext32(cast<uint32_t>(p)),r5900_sext32(cast<uint32_t>(shr<uint64_t>(cast<uint64_t>(p),cast<int64_t>(32ULL)))));
c->LO1 = std::get<0>(tmp85);
c->HI1 = std::get<1>(tmp85);
r5900_CPU_set(c,rd,c->LO1);
break;}
case cast<uint32_t>(25ULL):{
uint64_t p = cast<uint64_t>((cast<uint64_t>(cast<uint32_t>(c->R[rs].Lo)) * cast<uint64_t>(cast<uint32_t>(c->R[rt].Lo))));
auto tmp86 = std::make_tuple(r5900_sext32(cast<uint32_t>(p)),r5900_sext32(cast<uint32_t>(shr<uint64_t>(p,cast<int64_t>(32ULL)))));
c->LO1 = std::get<0>(tmp86);
c->HI1 = std::get<1>(tmp86);
r5900_CPU_set(c,rd,c->LO1);
break;}
case cast<uint32_t>(26ULL):{
auto tmp87 = r5900_divSigned32(cast<int32_t>(cast<uint32_t>(c->R[rs].Lo)),cast<int32_t>(cast<uint32_t>(c->R[rt].Lo)));
uint64_t lo = std::get<0>(tmp87);
uint64_t hi = std::get<1>(tmp87);
auto tmp88 = std::make_tuple(lo,hi);
c->LO1 = std::get<0>(tmp88);
c->HI1 = std::get<1>(tmp88);
break;}
case cast<uint32_t>(27ULL):{
auto tmp89 = r5900_divUnsigned32(cast<uint32_t>(c->R[rs].Lo),cast<uint32_t>(c->R[rt].Lo));
uint64_t lo = std::get<0>(tmp89);
uint64_t hi = std::get<1>(tmp89);
auto tmp90 = std::make_tuple(lo,hi);
c->LO1 = std::get<0>(tmp90);
c->HI1 = std::get<1>(tmp90);
break;}
case cast<uint32_t>(32ULL):{
r5900_CPU_maddAcc(c,rd,cast<int64_t>((cast<int64_t>(cast<int32_t>(cast<uint32_t>(c->R[rs].Lo))) * cast<int64_t>(cast<int32_t>(cast<uint32_t>(c->R[rt].Lo))))),true);
break;}
case cast<uint32_t>(33ULL):{
r5900_CPU_maddAcc(c,rd,cast<int64_t>(cast<uint64_t>((cast<uint64_t>(cast<uint32_t>(c->R[rs].Lo)) * cast<uint64_t>(cast<uint32_t>(c->R[rt].Lo))))),true);
break;}
case cast<uint32_t>(40ULL):{
r5900_CPU_mmi1(c,rs,rt,rd,shamt);
break;}
case cast<uint32_t>(41ULL):{
r5900_CPU_mmi3(c,rs,rt,rd,shamt);
break;}
case cast<uint32_t>(48ULL):{
r5900_CPU_pmfhl(c,rd,shamt);
break;}
case cast<uint32_t>(49ULL):{
r5900_CPU_pmthl(c,rs,shamt);
break;}
case cast<uint32_t>(52ULL):{
std::array<uint16_t,8> h = r5900_halves(c->R[rt]);
{auto&& tmp91 = h;
for(int64_t tmp92=0;tmp92<len(tmp91);++tmp92){
auto i=tmp92;h[i] = shl<uint16_t>(h[i],cast<uint32_t>((shamt & cast<uint32_t>(15ULL))));
}}
r5900_CPU_putH(c,rd,h);
break;}
case cast<uint32_t>(54ULL):{
std::array<uint16_t,8> h = r5900_halves(c->R[rt]);
{auto&& tmp93 = h;
for(int64_t tmp94=0;tmp94<len(tmp93);++tmp94){
auto i=tmp94;h[i] = shr<uint16_t>(h[i],cast<uint32_t>((shamt & cast<uint32_t>(15ULL))));
}}
r5900_CPU_putH(c,rd,h);
break;}
case cast<uint32_t>(55ULL):{
std::array<uint16_t,8> h = r5900_halves(c->R[rt]);
{auto&& tmp95 = h;
for(int64_t tmp96=0;tmp96<len(tmp95);++tmp96){
auto i=tmp96;h[i] = cast<uint16_t>(shr<int16_t>(cast<int16_t>(h[i]),(cast<uint32_t>((shamt & cast<uint32_t>(15ULL))))));
}}
r5900_CPU_putH(c,rd,h);
break;}
case cast<uint32_t>(60ULL):{
std::array<uint32_t,4> v = r5900_words(c->R[rt]);
{auto&& tmp97 = v;
for(int64_t tmp98=0;tmp98<len(tmp97);++tmp98){
auto i=tmp98;v[i] = shl<uint32_t>(v[i],cast<uint32_t>((shamt & cast<uint32_t>(31ULL))));
}}
r5900_CPU_putW(c,rd,v);
break;}
case cast<uint32_t>(62ULL):{
std::array<uint32_t,4> v = r5900_words(c->R[rt]);
{auto&& tmp99 = v;
for(int64_t tmp100=0;tmp100<len(tmp99);++tmp100){
auto i=tmp100;v[i] = shr<uint32_t>(v[i],cast<uint32_t>((shamt & cast<uint32_t>(31ULL))));
}}
r5900_CPU_putW(c,rd,v);
break;}
case cast<uint32_t>(63ULL):{
std::array<uint32_t,4> v = r5900_words(c->R[rt]);
{auto&& tmp101 = v;
for(int64_t tmp102=0;tmp102<len(tmp101);++tmp102){
auto i=tmp102;v[i] = cast<uint32_t>(shr<int32_t>(cast<int32_t>(v[i]),(cast<uint32_t>((shamt & cast<uint32_t>(31ULL))))));
}}
r5900_CPU_putW(c,rd,v);
break;}
default:{
r5900_CPU_Exception(c,cast<uint32_t>(10ULL));
break;}
}}
}
}
// tools/cpu/r5900/mmi.go:253:1
void r5900_CPU_maddAcc(r5900_CPU* c,uint32_t rd,int64_t p,bool second){
{
auto tmp103 = std::make_tuple((&c->HI),(&c->LO));
uint64_t* hi = std::get<0>(tmp103);
uint64_t* lo = std::get<1>(tmp103);
if (second) {
auto tmp104 = std::make_tuple((&c->HI1),(&c->LO1));
hi = std::get<0>(tmp104);
lo = std::get<1>(tmp104);
}
int64_t acc = cast<int64_t>(cast<uint64_t>((cast<uint64_t>(cast<uint32_t>((*lo))) | shl<uint64_t>(cast<uint64_t>(cast<uint32_t>((*hi))),cast<int64_t>(32ULL)))));
uint64_t sum = cast<uint64_t>(cast<int64_t>((acc + p)));
auto tmp105 = std::make_tuple(r5900_sext32(cast<uint32_t>(sum)),r5900_sext32(cast<uint32_t>(shr<uint64_t>(sum,cast<int64_t>(32ULL)))));
(*lo) = std::get<0>(tmp105);
(*hi) = std::get<1>(tmp105);
r5900_CPU_set(c,rd,(*lo));
}
}
// tools/cpu/r5900/mmi.go:266:1
uint32_t r5900_lzcw(uint32_t v){
{
if ((cast<uint32_t>((v & cast<uint32_t>(2147483648ULL))) != cast<uint32_t>(0ULL))) {
v = cast<uint32_t>(~v);
}
return cast<uint32_t>((cast<uint32_t>(go_bits_LeadingZeros32(v)) - cast<uint32_t>(1ULL)));
}
}
// tools/cpu/r5900/mmi.go:273:1
std::tuple<uint64_t,uint64_t> r5900_divSigned32(int32_t a,int32_t b){
uint64_t lo{};
uint64_t hi{};
{
{
if ((b == cast<int32_t>(0ULL))){
hi = r5900_sext32(cast<uint32_t>(a));
if ((a >= cast<int32_t>(0ULL))) {
lo = r5900_sext32(cast<uint32_t>(4294967295ULL));
}
else {
lo = cast<uint64_t>(1ULL);
}
}
else if (((a == cast<int32_t>(-2147483648ULL)) && (b == cast<int32_t>(-1ULL)))){
auto tmp107 = std::make_tuple(r5900_sext32(cast<uint32_t>(2147483648ULL)),cast<uint64_t>(0ULL));
lo = std::get<0>(tmp107);
hi = std::get<1>(tmp107);
}
else {
auto tmp108 = std::make_tuple(r5900_sext32(cast<uint32_t>(divi<int32_t>(a,b))),r5900_sext32(cast<uint32_t>(modi<int32_t>(a,b))));
lo = std::get<0>(tmp108);
hi = std::get<1>(tmp108);
}
}
tmp106:;
return {lo,hi};
}
}
// tools/cpu/r5900/mmi.go:290:1
std::tuple<uint64_t,uint64_t> r5900_divUnsigned32(uint32_t a,uint32_t b){
uint64_t lo{};
uint64_t hi{};
{
if ((b == cast<uint32_t>(0ULL))) {
return {r5900_sext32(cast<uint32_t>(4294967295ULL)),r5900_sext32(a)};
}
return {r5900_sext32(divi<uint32_t>(a,b)),r5900_sext32(modi<uint32_t>(a,b))};
}
}
// tools/cpu/r5900/mmi.go:300:1
void r5900_CPU_mmi0(r5900_CPU* c,uint32_t rs,uint32_t rt,uint32_t rd,uint32_t sub_){
{
auto tmp109 = std::make_tuple(c->R[rs],c->R[rt]);
r5900_Quad s = std::get<0>(tmp109);
r5900_Quad t = std::get<1>(tmp109);
{
switch(sub_){
case cast<uint32_t>(0ULL):{
auto tmp110 = std::make_tuple(r5900_words(s),r5900_words(t));
std::array<uint32_t,4> a = std::get<0>(tmp110);
std::array<uint32_t,4> b = std::get<1>(tmp110);
{auto&& tmp111 = a;
for(int64_t tmp112=0;tmp112<len(tmp111);++tmp112){
auto i=tmp112;a[i] += b[i];
}}
r5900_CPU_putW(c,rd,a);
break;}
case cast<uint32_t>(1ULL):{
auto tmp113 = std::make_tuple(r5900_words(s),r5900_words(t));
std::array<uint32_t,4> a = std::get<0>(tmp113);
std::array<uint32_t,4> b = std::get<1>(tmp113);
{auto&& tmp114 = a;
for(int64_t tmp115=0;tmp115<len(tmp114);++tmp115){
auto i=tmp115;a[i] -= b[i];
}}
r5900_CPU_putW(c,rd,a);
break;}
case cast<uint32_t>(2ULL):{
auto tmp116 = std::make_tuple(r5900_words(s),r5900_words(t));
std::array<uint32_t,4> a = std::get<0>(tmp116);
std::array<uint32_t,4> b = std::get<1>(tmp116);
{auto&& tmp117 = a;
for(int64_t tmp118=0;tmp118<len(tmp117);++tmp118){
auto i=tmp118;a[i] = r5900_maskIf((cast<int32_t>(a[i]) > cast<int32_t>(b[i])));
}}
r5900_CPU_putW(c,rd,a);
break;}
case cast<uint32_t>(3ULL):{
auto tmp119 = std::make_tuple(r5900_words(s),r5900_words(t));
std::array<uint32_t,4> a = std::get<0>(tmp119);
std::array<uint32_t,4> b = std::get<1>(tmp119);
{auto&& tmp120 = a;
for(int64_t tmp121=0;tmp121<len(tmp120);++tmp121){
auto i=tmp121;if ((cast<int32_t>(b[i]) > cast<int32_t>(a[i]))) {
a[i] = b[i];
}
}}
r5900_CPU_putW(c,rd,a);
break;}
case cast<uint32_t>(4ULL):{
auto tmp122 = std::make_tuple(r5900_halves(s),r5900_halves(t));
std::array<uint16_t,8> a = std::get<0>(tmp122);
std::array<uint16_t,8> b = std::get<1>(tmp122);
{auto&& tmp123 = a;
for(int64_t tmp124=0;tmp124<len(tmp123);++tmp124){
auto i=tmp124;a[i] += b[i];
}}
r5900_CPU_putH(c,rd,a);
break;}
case cast<uint32_t>(5ULL):{
auto tmp125 = std::make_tuple(r5900_halves(s),r5900_halves(t));
std::array<uint16_t,8> a = std::get<0>(tmp125);
std::array<uint16_t,8> b = std::get<1>(tmp125);
{auto&& tmp126 = a;
for(int64_t tmp127=0;tmp127<len(tmp126);++tmp127){
auto i=tmp127;a[i] -= b[i];
}}
r5900_CPU_putH(c,rd,a);
break;}
case cast<uint32_t>(6ULL):{
auto tmp128 = std::make_tuple(r5900_halves(s),r5900_halves(t));
std::array<uint16_t,8> a = std::get<0>(tmp128);
std::array<uint16_t,8> b = std::get<1>(tmp128);
{auto&& tmp129 = a;
for(int64_t tmp130=0;tmp130<len(tmp129);++tmp130){
auto i=tmp130;a[i] = cast<uint16_t>(r5900_maskIf((cast<int16_t>(a[i]) > cast<int16_t>(b[i]))));
}}
r5900_CPU_putH(c,rd,a);
break;}
case cast<uint32_t>(7ULL):{
auto tmp131 = std::make_tuple(r5900_halves(s),r5900_halves(t));
std::array<uint16_t,8> a = std::get<0>(tmp131);
std::array<uint16_t,8> b = std::get<1>(tmp131);
{auto&& tmp132 = a;
for(int64_t tmp133=0;tmp133<len(tmp132);++tmp133){
auto i=tmp133;if ((cast<int16_t>(b[i]) > cast<int16_t>(a[i]))) {
a[i] = b[i];
}
}}
r5900_CPU_putH(c,rd,a);
break;}
case cast<uint32_t>(8ULL):{
auto tmp134 = std::make_tuple(r5900_octets(s),r5900_octets(t));
std::array<uint8_t,16> a = std::get<0>(tmp134);
std::array<uint8_t,16> b = std::get<1>(tmp134);
{auto&& tmp135 = a;
for(int64_t tmp136=0;tmp136<len(tmp135);++tmp136){
auto i=tmp136;a[i] += b[i];
}}
r5900_CPU_putB(c,rd,a);
break;}
case cast<uint32_t>(9ULL):{
auto tmp137 = std::make_tuple(r5900_octets(s),r5900_octets(t));
std::array<uint8_t,16> a = std::get<0>(tmp137);
std::array<uint8_t,16> b = std::get<1>(tmp137);
{auto&& tmp138 = a;
for(int64_t tmp139=0;tmp139<len(tmp138);++tmp139){
auto i=tmp139;a[i] -= b[i];
}}
r5900_CPU_putB(c,rd,a);
break;}
case cast<uint32_t>(10ULL):{
auto tmp140 = std::make_tuple(r5900_octets(s),r5900_octets(t));
std::array<uint8_t,16> a = std::get<0>(tmp140);
std::array<uint8_t,16> b = std::get<1>(tmp140);
{auto&& tmp141 = a;
for(int64_t tmp142=0;tmp142<len(tmp141);++tmp142){
auto i=tmp142;a[i] = cast<uint8_t>(r5900_maskIf((cast<int8_t>(a[i]) > cast<int8_t>(b[i]))));
}}
r5900_CPU_putB(c,rd,a);
break;}
case cast<uint32_t>(16ULL):{
auto tmp143 = std::make_tuple(r5900_words(s),r5900_words(t));
std::array<uint32_t,4> a = std::get<0>(tmp143);
std::array<uint32_t,4> b = std::get<1>(tmp143);
{auto&& tmp144 = a;
for(int64_t tmp145=0;tmp145<len(tmp144);++tmp145){
auto i=tmp145;a[i] = r5900_satS32(cast<int64_t>((cast<int64_t>(cast<int32_t>(a[i])) + cast<int64_t>(cast<int32_t>(b[i])))));
}}
r5900_CPU_putW(c,rd,a);
break;}
case cast<uint32_t>(17ULL):{
auto tmp146 = std::make_tuple(r5900_words(s),r5900_words(t));
std::array<uint32_t,4> a = std::get<0>(tmp146);
std::array<uint32_t,4> b = std::get<1>(tmp146);
{auto&& tmp147 = a;
for(int64_t tmp148=0;tmp148<len(tmp147);++tmp148){
auto i=tmp148;a[i] = r5900_satS32(cast<int64_t>((cast<int64_t>(cast<int32_t>(a[i])) - cast<int64_t>(cast<int32_t>(b[i])))));
}}
r5900_CPU_putW(c,rd,a);
break;}
case cast<uint32_t>(18ULL):{
auto tmp149 = std::make_tuple(r5900_words(s),r5900_words(t));
std::array<uint32_t,4> a = std::get<0>(tmp149);
std::array<uint32_t,4> b = std::get<1>(tmp149);
r5900_CPU_putW(c,rd,std::array<uint32_t,4>{b[cast<int64_t>(0ULL)],a[cast<int64_t>(0ULL)],b[cast<int64_t>(1ULL)],a[cast<int64_t>(1ULL)]});
break;}
case cast<uint32_t>(19ULL):{
auto tmp150 = std::make_tuple(r5900_words(s),r5900_words(t));
std::array<uint32_t,4> a = std::get<0>(tmp150);
std::array<uint32_t,4> b = std::get<1>(tmp150);
r5900_CPU_putW(c,rd,std::array<uint32_t,4>{b[cast<int64_t>(0ULL)],b[cast<int64_t>(2ULL)],a[cast<int64_t>(0ULL)],a[cast<int64_t>(2ULL)]});
break;}
case cast<uint32_t>(20ULL):{
auto tmp151 = std::make_tuple(r5900_halves(s),r5900_halves(t));
std::array<uint16_t,8> a = std::get<0>(tmp151);
std::array<uint16_t,8> b = std::get<1>(tmp151);
{auto&& tmp152 = a;
for(int64_t tmp153=0;tmp153<len(tmp152);++tmp153){
auto i=tmp153;a[i] = r5900_satS16(cast<int32_t>((cast<int32_t>(cast<int16_t>(a[i])) + cast<int32_t>(cast<int16_t>(b[i])))));
}}
r5900_CPU_putH(c,rd,a);
break;}
case cast<uint32_t>(21ULL):{
auto tmp154 = std::make_tuple(r5900_halves(s),r5900_halves(t));
std::array<uint16_t,8> a = std::get<0>(tmp154);
std::array<uint16_t,8> b = std::get<1>(tmp154);
{auto&& tmp155 = a;
for(int64_t tmp156=0;tmp156<len(tmp155);++tmp156){
auto i=tmp156;a[i] = r5900_satS16(cast<int32_t>((cast<int32_t>(cast<int16_t>(a[i])) - cast<int32_t>(cast<int16_t>(b[i])))));
}}
r5900_CPU_putH(c,rd,a);
break;}
case cast<uint32_t>(22ULL):{
auto tmp157 = std::make_tuple(r5900_halves(s),r5900_halves(t));
std::array<uint16_t,8> a = std::get<0>(tmp157);
std::array<uint16_t,8> b = std::get<1>(tmp157);
r5900_CPU_putH(c,rd,std::array<uint16_t,8>{b[cast<int64_t>(0ULL)],a[cast<int64_t>(0ULL)],b[cast<int64_t>(1ULL)],a[cast<int64_t>(1ULL)],b[cast<int64_t>(2ULL)],a[cast<int64_t>(2ULL)],b[cast<int64_t>(3ULL)],a[cast<int64_t>(3ULL)]});
break;}
case cast<uint32_t>(23ULL):{
auto tmp158 = std::make_tuple(r5900_halves(s),r5900_halves(t));
std::array<uint16_t,8> a = std::get<0>(tmp158);
std::array<uint16_t,8> b = std::get<1>(tmp158);
r5900_CPU_putH(c,rd,std::array<uint16_t,8>{b[cast<int64_t>(0ULL)],b[cast<int64_t>(2ULL)],b[cast<int64_t>(4ULL)],b[cast<int64_t>(6ULL)],a[cast<int64_t>(0ULL)],a[cast<int64_t>(2ULL)],a[cast<int64_t>(4ULL)],a[cast<int64_t>(6ULL)]});
break;}
case cast<uint32_t>(24ULL):{
auto tmp159 = std::make_tuple(r5900_octets(s),r5900_octets(t));
std::array<uint8_t,16> a = std::get<0>(tmp159);
std::array<uint8_t,16> b = std::get<1>(tmp159);
{auto&& tmp160 = a;
for(int64_t tmp161=0;tmp161<len(tmp160);++tmp161){
auto i=tmp161;a[i] = r5900_satS8(cast<int32_t>((cast<int32_t>(cast<int8_t>(a[i])) + cast<int32_t>(cast<int8_t>(b[i])))));
}}
r5900_CPU_putB(c,rd,a);
break;}
case cast<uint32_t>(25ULL):{
auto tmp162 = std::make_tuple(r5900_octets(s),r5900_octets(t));
std::array<uint8_t,16> a = std::get<0>(tmp162);
std::array<uint8_t,16> b = std::get<1>(tmp162);
{auto&& tmp163 = a;
for(int64_t tmp164=0;tmp164<len(tmp163);++tmp164){
auto i=tmp164;a[i] = r5900_satS8(cast<int32_t>((cast<int32_t>(cast<int8_t>(a[i])) - cast<int32_t>(cast<int8_t>(b[i])))));
}}
r5900_CPU_putB(c,rd,a);
break;}
case cast<uint32_t>(26ULL):{
auto tmp165 = std::make_tuple(r5900_octets(s),r5900_octets(t));
std::array<uint8_t,16> a = std::get<0>(tmp165);
std::array<uint8_t,16> b = std::get<1>(tmp165);
std::array<uint8_t,16> o={};
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(8ULL));i++){
auto tmp166 = std::make_tuple(b[i],a[i]);
o[cast<int64_t>((cast<int64_t>(2ULL) * i))] = std::get<0>(tmp166);
o[cast<int64_t>((cast<int64_t>((cast<int64_t>(2ULL) * i)) + cast<int64_t>(1ULL)))] = std::get<1>(tmp166);
}
}r5900_CPU_putB(c,rd,o);
break;}
case cast<uint32_t>(27ULL):{
auto tmp167 = std::make_tuple(r5900_octets(s),r5900_octets(t));
std::array<uint8_t,16> a = std::get<0>(tmp167);
std::array<uint8_t,16> b = std::get<1>(tmp167);
std::array<uint8_t,16> o={};
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(8ULL));i++){
auto tmp168 = std::make_tuple(b[cast<int64_t>((cast<int64_t>(2ULL) * i))],a[cast<int64_t>((cast<int64_t>(2ULL) * i))]);
o[i] = std::get<0>(tmp168);
o[cast<int64_t>((i + cast<int64_t>(8ULL)))] = std::get<1>(tmp168);
}
}r5900_CPU_putB(c,rd,o);
break;}
case cast<uint32_t>(30ULL):{
std::array<uint32_t,4> b = r5900_words(t);
std::array<uint32_t,4> o={};
{auto&& tmp169 = b;
for(int64_t tmp170=0;tmp170<len(tmp169);++tmp170){
auto i=tmp170;uint32_t v = cast<uint32_t>((b[i] & cast<uint32_t>(65535ULL)));
o[i] = cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((shl<uint32_t>((cast<uint32_t>((v & cast<uint32_t>(31ULL)))),cast<int64_t>(3ULL)) | shl<uint32_t>((cast<uint32_t>(((shr<uint32_t>(v,cast<int64_t>(5ULL))) & cast<uint32_t>(31ULL)))),cast<int64_t>(11ULL)))) | shl<uint32_t>((cast<uint32_t>(((shr<uint32_t>(v,cast<int64_t>(10ULL))) & cast<uint32_t>(31ULL)))),cast<int64_t>(19ULL)))) | shl<uint32_t>((cast<uint32_t>(((shr<uint32_t>(v,cast<int64_t>(15ULL))) & cast<uint32_t>(1ULL)))),cast<int64_t>(31ULL))));
}}
r5900_CPU_putW(c,rd,o);
break;}
case cast<uint32_t>(31ULL):{
std::array<uint32_t,4> b = r5900_words(t);
std::array<uint32_t,4> o={};
{auto&& tmp171 = b;
for(int64_t tmp172=0;tmp172<len(tmp171);++tmp172){
auto i=tmp172;uint32_t v = b[i];
o[i] = cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(((shr<uint32_t>(v,cast<int64_t>(3ULL))) & cast<uint32_t>(31ULL))) | shl<uint32_t>((cast<uint32_t>(((shr<uint32_t>(v,cast<int64_t>(11ULL))) & cast<uint32_t>(31ULL)))),cast<int64_t>(5ULL)))) | shl<uint32_t>((cast<uint32_t>(((shr<uint32_t>(v,cast<int64_t>(19ULL))) & cast<uint32_t>(31ULL)))),cast<int64_t>(10ULL)))) | shl<uint32_t>((cast<uint32_t>(((shr<uint32_t>(v,cast<int64_t>(31ULL))) & cast<uint32_t>(1ULL)))),cast<int64_t>(15ULL))));
}}
r5900_CPU_putW(c,rd,o);
break;}
default:{
r5900_CPU_Exception(c,cast<uint32_t>(10ULL));
break;}
}}
}
}
// tools/cpu/r5900/mmi.go:467:1
void r5900_CPU_mmi1(r5900_CPU* c,uint32_t rs,uint32_t rt,uint32_t rd,uint32_t sub_){
{
auto tmp173 = std::make_tuple(c->R[rs],c->R[rt]);
r5900_Quad s = std::get<0>(tmp173);
r5900_Quad t = std::get<1>(tmp173);
{
switch(sub_){
case cast<uint32_t>(1ULL):{
std::array<uint32_t,4> b = r5900_words(t);
{auto&& tmp174 = b;
for(int64_t tmp175=0;tmp175<len(tmp174);++tmp175){
auto i=tmp175;b[i] = r5900_absW(b[i]);
}}
r5900_CPU_putW(c,rd,b);
break;}
case cast<uint32_t>(2ULL):{
auto tmp176 = std::make_tuple(r5900_words(s),r5900_words(t));
std::array<uint32_t,4> a = std::get<0>(tmp176);
std::array<uint32_t,4> b = std::get<1>(tmp176);
{auto&& tmp177 = a;
for(int64_t tmp178=0;tmp178<len(tmp177);++tmp178){
auto i=tmp178;a[i] = r5900_maskIf((a[i] == b[i]));
}}
r5900_CPU_putW(c,rd,a);
break;}
case cast<uint32_t>(3ULL):{
auto tmp179 = std::make_tuple(r5900_words(s),r5900_words(t));
std::array<uint32_t,4> a = std::get<0>(tmp179);
std::array<uint32_t,4> b = std::get<1>(tmp179);
{auto&& tmp180 = a;
for(int64_t tmp181=0;tmp181<len(tmp180);++tmp181){
auto i=tmp181;if ((cast<int32_t>(b[i]) < cast<int32_t>(a[i]))) {
a[i] = b[i];
}
}}
r5900_CPU_putW(c,rd,a);
break;}
case cast<uint32_t>(4ULL):{
auto tmp182 = std::make_tuple(r5900_halves(s),r5900_halves(t));
std::array<uint16_t,8> a = std::get<0>(tmp182);
std::array<uint16_t,8> b = std::get<1>(tmp182);
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(4ULL));i++){
a[i] -= b[i];
}
}{int64_t i = cast<int64_t>(4ULL);for (;(i < cast<int64_t>(8ULL));i++){
a[i] += b[i];
}
}r5900_CPU_putH(c,rd,a);
break;}
case cast<uint32_t>(5ULL):{
std::array<uint16_t,8> b = r5900_halves(t);
{auto&& tmp183 = b;
for(int64_t tmp184=0;tmp184<len(tmp183);++tmp184){
auto i=tmp184;b[i] = r5900_absH(b[i]);
}}
r5900_CPU_putH(c,rd,b);
break;}
case cast<uint32_t>(6ULL):{
auto tmp185 = std::make_tuple(r5900_halves(s),r5900_halves(t));
std::array<uint16_t,8> a = std::get<0>(tmp185);
std::array<uint16_t,8> b = std::get<1>(tmp185);
{auto&& tmp186 = a;
for(int64_t tmp187=0;tmp187<len(tmp186);++tmp187){
auto i=tmp187;a[i] = cast<uint16_t>(r5900_maskIf((a[i] == b[i])));
}}
r5900_CPU_putH(c,rd,a);
break;}
case cast<uint32_t>(7ULL):{
auto tmp188 = std::make_tuple(r5900_halves(s),r5900_halves(t));
std::array<uint16_t,8> a = std::get<0>(tmp188);
std::array<uint16_t,8> b = std::get<1>(tmp188);
{auto&& tmp189 = a;
for(int64_t tmp190=0;tmp190<len(tmp189);++tmp190){
auto i=tmp190;if ((cast<int16_t>(b[i]) < cast<int16_t>(a[i]))) {
a[i] = b[i];
}
}}
r5900_CPU_putH(c,rd,a);
break;}
case cast<uint32_t>(10ULL):{
auto tmp191 = std::make_tuple(r5900_octets(s),r5900_octets(t));
std::array<uint8_t,16> a = std::get<0>(tmp191);
std::array<uint8_t,16> b = std::get<1>(tmp191);
{auto&& tmp192 = a;
for(int64_t tmp193=0;tmp193<len(tmp192);++tmp193){
auto i=tmp193;a[i] = cast<uint8_t>(r5900_maskIf((a[i] == b[i])));
}}
r5900_CPU_putB(c,rd,a);
break;}
case cast<uint32_t>(16ULL):{
auto tmp194 = std::make_tuple(r5900_words(s),r5900_words(t));
std::array<uint32_t,4> a = std::get<0>(tmp194);
std::array<uint32_t,4> b = std::get<1>(tmp194);
{auto&& tmp195 = a;
for(int64_t tmp196=0;tmp196<len(tmp195);++tmp196){
auto i=tmp196;a[i] = r5900_satU32(cast<int64_t>((cast<int64_t>(a[i]) + cast<int64_t>(b[i]))));
}}
r5900_CPU_putW(c,rd,a);
break;}
case cast<uint32_t>(17ULL):{
auto tmp197 = std::make_tuple(r5900_words(s),r5900_words(t));
std::array<uint32_t,4> a = std::get<0>(tmp197);
std::array<uint32_t,4> b = std::get<1>(tmp197);
{auto&& tmp198 = a;
for(int64_t tmp199=0;tmp199<len(tmp198);++tmp199){
auto i=tmp199;a[i] = r5900_satU32(cast<int64_t>((cast<int64_t>(a[i]) - cast<int64_t>(b[i]))));
}}
r5900_CPU_putW(c,rd,a);
break;}
case cast<uint32_t>(18ULL):{
auto tmp200 = std::make_tuple(r5900_words(s),r5900_words(t));
std::array<uint32_t,4> a = std::get<0>(tmp200);
std::array<uint32_t,4> b = std::get<1>(tmp200);
r5900_CPU_putW(c,rd,std::array<uint32_t,4>{b[cast<int64_t>(2ULL)],a[cast<int64_t>(2ULL)],b[cast<int64_t>(3ULL)],a[cast<int64_t>(3ULL)]});
break;}
case cast<uint32_t>(20ULL):{
auto tmp201 = std::make_tuple(r5900_halves(s),r5900_halves(t));
std::array<uint16_t,8> a = std::get<0>(tmp201);
std::array<uint16_t,8> b = std::get<1>(tmp201);
{auto&& tmp202 = a;
for(int64_t tmp203=0;tmp203<len(tmp202);++tmp203){
auto i=tmp203;a[i] = r5900_satU16(cast<int32_t>((cast<int32_t>(a[i]) + cast<int32_t>(b[i]))));
}}
r5900_CPU_putH(c,rd,a);
break;}
case cast<uint32_t>(21ULL):{
auto tmp204 = std::make_tuple(r5900_halves(s),r5900_halves(t));
std::array<uint16_t,8> a = std::get<0>(tmp204);
std::array<uint16_t,8> b = std::get<1>(tmp204);
{auto&& tmp205 = a;
for(int64_t tmp206=0;tmp206<len(tmp205);++tmp206){
auto i=tmp206;a[i] = r5900_satU16(cast<int32_t>((cast<int32_t>(a[i]) - cast<int32_t>(b[i]))));
}}
r5900_CPU_putH(c,rd,a);
break;}
case cast<uint32_t>(22ULL):{
auto tmp207 = std::make_tuple(r5900_halves(s),r5900_halves(t));
std::array<uint16_t,8> a = std::get<0>(tmp207);
std::array<uint16_t,8> b = std::get<1>(tmp207);
r5900_CPU_putH(c,rd,std::array<uint16_t,8>{b[cast<int64_t>(4ULL)],a[cast<int64_t>(4ULL)],b[cast<int64_t>(5ULL)],a[cast<int64_t>(5ULL)],b[cast<int64_t>(6ULL)],a[cast<int64_t>(6ULL)],b[cast<int64_t>(7ULL)],a[cast<int64_t>(7ULL)]});
break;}
case cast<uint32_t>(24ULL):{
auto tmp208 = std::make_tuple(r5900_octets(s),r5900_octets(t));
std::array<uint8_t,16> a = std::get<0>(tmp208);
std::array<uint8_t,16> b = std::get<1>(tmp208);
{auto&& tmp209 = a;
for(int64_t tmp210=0;tmp210<len(tmp209);++tmp210){
auto i=tmp210;a[i] = r5900_satU8(cast<int32_t>((cast<int32_t>(a[i]) + cast<int32_t>(b[i]))));
}}
r5900_CPU_putB(c,rd,a);
break;}
case cast<uint32_t>(25ULL):{
auto tmp211 = std::make_tuple(r5900_octets(s),r5900_octets(t));
std::array<uint8_t,16> a = std::get<0>(tmp211);
std::array<uint8_t,16> b = std::get<1>(tmp211);
{auto&& tmp212 = a;
for(int64_t tmp213=0;tmp213<len(tmp212);++tmp213){
auto i=tmp213;a[i] = r5900_satU8(cast<int32_t>((cast<int32_t>(a[i]) - cast<int32_t>(b[i]))));
}}
r5900_CPU_putB(c,rd,a);
break;}
case cast<uint32_t>(26ULL):{
auto tmp214 = std::make_tuple(r5900_octets(s),r5900_octets(t));
std::array<uint8_t,16> a = std::get<0>(tmp214);
std::array<uint8_t,16> b = std::get<1>(tmp214);
std::array<uint8_t,16> o={};
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(8ULL));i++){
auto tmp215 = std::make_tuple(b[cast<int64_t>((i + cast<int64_t>(8ULL)))],a[cast<int64_t>((i + cast<int64_t>(8ULL)))]);
o[cast<int64_t>((cast<int64_t>(2ULL) * i))] = std::get<0>(tmp215);
o[cast<int64_t>((cast<int64_t>((cast<int64_t>(2ULL) * i)) + cast<int64_t>(1ULL)))] = std::get<1>(tmp215);
}
}r5900_CPU_putB(c,rd,o);
break;}
case cast<uint32_t>(27ULL):{
uint32_t n = cast<uint32_t>((c->SA & cast<uint32_t>(127ULL)));
r5900_CPU_putQ(c,rd,r5900_funnelRight(s,t,n));
break;}
default:{
r5900_CPU_Exception(c,cast<uint32_t>(10ULL));
break;}
}}
}
}
// tools/cpu/r5900/mmi.go:595:1
void r5900_CPU_mmi2(r5900_CPU* c,uint32_t rs,uint32_t rt,uint32_t rd,uint32_t sub_){
{
auto tmp216 = std::make_tuple(c->R[rs],c->R[rt]);
r5900_Quad s = std::get<0>(tmp216);
r5900_Quad t = std::get<1>(tmp216);
{
switch(sub_){
case cast<uint32_t>(0ULL):{
r5900_CPU_pmulW(c,rd,s,t,true,false,true);
break;}
case cast<uint32_t>(2ULL):{
auto tmp217 = std::make_tuple(r5900_words(s),r5900_words(t));
std::array<uint32_t,4> a = std::get<0>(tmp217);
std::array<uint32_t,4> b = std::get<1>(tmp217);
r5900_CPU_putQ(c,rd,r5900_Quad{r5900_sext32(shl<uint32_t>(b[cast<int64_t>(0ULL)],(cast<uint32_t>((a[cast<int64_t>(0ULL)] & cast<uint32_t>(31ULL)))))),r5900_sext32(shl<uint32_t>(b[cast<int64_t>(2ULL)],(cast<uint32_t>((a[cast<int64_t>(2ULL)] & cast<uint32_t>(31ULL))))))});
break;}
case cast<uint32_t>(3ULL):{
auto tmp218 = std::make_tuple(r5900_words(s),r5900_words(t));
std::array<uint32_t,4> a = std::get<0>(tmp218);
std::array<uint32_t,4> b = std::get<1>(tmp218);
r5900_CPU_putQ(c,rd,r5900_Quad{r5900_sext32(shr<uint32_t>(b[cast<int64_t>(0ULL)],(cast<uint32_t>((a[cast<int64_t>(0ULL)] & cast<uint32_t>(31ULL)))))),r5900_sext32(shr<uint32_t>(b[cast<int64_t>(2ULL)],(cast<uint32_t>((a[cast<int64_t>(2ULL)] & cast<uint32_t>(31ULL))))))});
break;}
case cast<uint32_t>(4ULL):{
r5900_CPU_pmulW(c,rd,s,t,true,true,true);
break;}
case cast<uint32_t>(8ULL):{
r5900_CPU_putQ(c,rd,r5900_CPU_hiQ(c));
break;}
case cast<uint32_t>(9ULL):{
r5900_CPU_putQ(c,rd,r5900_CPU_loQ(c));
break;}
case cast<uint32_t>(10ULL):{
auto tmp219 = std::make_tuple(r5900_halves(s),r5900_halves(t));
std::array<uint16_t,8> a = std::get<0>(tmp219);
std::array<uint16_t,8> b = std::get<1>(tmp219);
r5900_CPU_putH(c,rd,std::array<uint16_t,8>{b[cast<int64_t>(0ULL)],a[cast<int64_t>(4ULL)],b[cast<int64_t>(1ULL)],a[cast<int64_t>(5ULL)],b[cast<int64_t>(2ULL)],a[cast<int64_t>(6ULL)],b[cast<int64_t>(3ULL)],a[cast<int64_t>(7ULL)]});
break;}
case cast<uint32_t>(12ULL):{
r5900_CPU_pmulW(c,rd,s,t,false,false,true);
break;}
case cast<uint32_t>(13ULL):{
auto tmp220 = std::make_tuple(r5900_words(s),r5900_words(t));
std::array<uint32_t,4> a = std::get<0>(tmp220);
std::array<uint32_t,4> b = std::get<1>(tmp220);
auto tmp221 = r5900_divSigned32(cast<int32_t>(a[cast<int64_t>(0ULL)]),cast<int32_t>(b[cast<int64_t>(0ULL)]));
uint64_t lo0 = std::get<0>(tmp221);
uint64_t hi0 = std::get<1>(tmp221);
auto tmp222 = r5900_divSigned32(cast<int32_t>(a[cast<int64_t>(2ULL)]),cast<int32_t>(b[cast<int64_t>(2ULL)]));
uint64_t lo1 = std::get<0>(tmp222);
uint64_t hi1 = std::get<1>(tmp222);
auto tmp223 = std::make_tuple(lo0,hi0,lo1,hi1);
c->LO = std::get<0>(tmp223);
c->HI = std::get<1>(tmp223);
c->LO1 = std::get<2>(tmp223);
c->HI1 = std::get<3>(tmp223);
break;}
case cast<uint32_t>(14ULL):{
r5900_CPU_putQ(c,rd,r5900_Quad{t.Lo,s.Lo});
break;}
case cast<uint32_t>(16ULL):{
r5900_CPU_pmulH(c,rd,s,t,true,false);
break;}
case cast<uint32_t>(17ULL):{
r5900_CPU_phmulH(c,rd,s,t,false);
break;}
case cast<uint32_t>(18ULL):{
r5900_CPU_putQ(c,rd,r5900_Quad{cast<uint64_t>((s.Lo & t.Lo)),cast<uint64_t>((s.Hi & t.Hi))});
break;}
case cast<uint32_t>(19ULL):{
r5900_CPU_putQ(c,rd,r5900_Quad{cast<uint64_t>((s.Lo ^ t.Lo)),cast<uint64_t>((s.Hi ^ t.Hi))});
break;}
case cast<uint32_t>(20ULL):{
r5900_CPU_pmulH(c,rd,s,t,true,true);
break;}
case cast<uint32_t>(21ULL):{
r5900_CPU_phmulH(c,rd,s,t,true);
break;}
case cast<uint32_t>(26ULL):{
std::array<uint16_t,8> h = r5900_halves(t);
r5900_CPU_putH(c,rd,std::array<uint16_t,8>{h[cast<int64_t>(2ULL)],h[cast<int64_t>(1ULL)],h[cast<int64_t>(0ULL)],h[cast<int64_t>(3ULL)],h[cast<int64_t>(6ULL)],h[cast<int64_t>(5ULL)],h[cast<int64_t>(4ULL)],h[cast<int64_t>(7ULL)]});
break;}
case cast<uint32_t>(27ULL):{
std::array<uint16_t,8> h = r5900_halves(t);
r5900_CPU_putH(c,rd,std::array<uint16_t,8>{h[cast<int64_t>(3ULL)],h[cast<int64_t>(2ULL)],h[cast<int64_t>(1ULL)],h[cast<int64_t>(0ULL)],h[cast<int64_t>(7ULL)],h[cast<int64_t>(6ULL)],h[cast<int64_t>(5ULL)],h[cast<int64_t>(4ULL)]});
break;}
case cast<uint32_t>(28ULL):{
r5900_CPU_pmulH(c,rd,s,t,false,false);
break;}
case cast<uint32_t>(29ULL):{
auto tmp224 = std::make_tuple(r5900_words(s),r5900_words(t));
std::array<uint32_t,4> a = std::get<0>(tmp224);
std::array<uint32_t,4> b = std::get<1>(tmp224);
int32_t d = cast<int32_t>(cast<int16_t>(cast<uint16_t>(b[cast<int64_t>(0ULL)])));
std::array<uint32_t,4> lo={};
std::array<uint32_t,4> hi={};
{auto&& tmp225 = a;
for(int64_t tmp226=0;tmp226<len(tmp225);++tmp226){
auto i=tmp226;auto tmp227 = r5900_divSigned32(cast<int32_t>(a[i]),d);
uint64_t q = std::get<0>(tmp227);
uint64_t r = std::get<1>(tmp227);
auto tmp228 = std::make_tuple(cast<uint32_t>(q),cast<uint32_t>(r));
lo[i] = std::get<0>(tmp228);
hi[i] = std::get<1>(tmp228);
}}
r5900_CPU_setLoQ(c,r5900_fromWords(lo));
r5900_CPU_setHiQ(c,r5900_fromWords(hi));
break;}
case cast<uint32_t>(30ULL):{
std::array<uint32_t,4> v = r5900_words(t);
r5900_CPU_putW(c,rd,std::array<uint32_t,4>{v[cast<int64_t>(2ULL)],v[cast<int64_t>(1ULL)],v[cast<int64_t>(0ULL)],v[cast<int64_t>(3ULL)]});
break;}
case cast<uint32_t>(31ULL):{
std::array<uint32_t,4> v = r5900_words(t);
r5900_CPU_putW(c,rd,std::array<uint32_t,4>{v[cast<int64_t>(1ULL)],v[cast<int64_t>(2ULL)],v[cast<int64_t>(0ULL)],v[cast<int64_t>(3ULL)]});
break;}
default:{
r5900_CPU_Exception(c,cast<uint32_t>(10ULL));
break;}
}}
}
}
// tools/cpu/r5900/mmi.go:676:1
void r5900_CPU_mmi3(r5900_CPU* c,uint32_t rs,uint32_t rt,uint32_t rd,uint32_t sub_){
{
auto tmp229 = std::make_tuple(c->R[rs],c->R[rt]);
r5900_Quad s = std::get<0>(tmp229);
r5900_Quad t = std::get<1>(tmp229);
{
switch(sub_){
case cast<uint32_t>(0ULL):{
r5900_CPU_pmulW(c,rd,s,t,true,false,false);
break;}
case cast<uint32_t>(3ULL):{
auto tmp230 = std::make_tuple(r5900_words(s),r5900_words(t));
std::array<uint32_t,4> a = std::get<0>(tmp230);
std::array<uint32_t,4> b = std::get<1>(tmp230);
r5900_CPU_putQ(c,rd,r5900_Quad{r5900_sext32(cast<uint32_t>(shr<int32_t>(cast<int32_t>(b[cast<int64_t>(0ULL)]),(cast<uint32_t>((a[cast<int64_t>(0ULL)] & cast<uint32_t>(31ULL))))))),r5900_sext32(cast<uint32_t>(shr<int32_t>(cast<int32_t>(b[cast<int64_t>(2ULL)]),(cast<uint32_t>((a[cast<int64_t>(2ULL)] & cast<uint32_t>(31ULL)))))))});
break;}
case cast<uint32_t>(8ULL):{
auto tmp231 = std::make_tuple(s.Lo,s.Hi);
c->HI = std::get<0>(tmp231);
c->HI1 = std::get<1>(tmp231);
break;}
case cast<uint32_t>(9ULL):{
auto tmp232 = std::make_tuple(s.Lo,s.Hi);
c->LO = std::get<0>(tmp232);
c->LO1 = std::get<1>(tmp232);
break;}
case cast<uint32_t>(10ULL):{
auto tmp233 = std::make_tuple(r5900_halves(s),r5900_halves(t));
std::array<uint16_t,8> a = std::get<0>(tmp233);
std::array<uint16_t,8> b = std::get<1>(tmp233);
r5900_CPU_putH(c,rd,std::array<uint16_t,8>{b[cast<int64_t>(0ULL)],a[cast<int64_t>(0ULL)],b[cast<int64_t>(2ULL)],a[cast<int64_t>(2ULL)],b[cast<int64_t>(4ULL)],a[cast<int64_t>(4ULL)],b[cast<int64_t>(6ULL)],a[cast<int64_t>(6ULL)]});
break;}
case cast<uint32_t>(12ULL):{
r5900_CPU_pmulW(c,rd,s,t,false,false,false);
break;}
case cast<uint32_t>(13ULL):{
auto tmp234 = std::make_tuple(r5900_words(s),r5900_words(t));
std::array<uint32_t,4> a = std::get<0>(tmp234);
std::array<uint32_t,4> b = std::get<1>(tmp234);
auto tmp235 = r5900_divUnsigned32(a[cast<int64_t>(0ULL)],b[cast<int64_t>(0ULL)]);
uint64_t lo0 = std::get<0>(tmp235);
uint64_t hi0 = std::get<1>(tmp235);
auto tmp236 = r5900_divUnsigned32(a[cast<int64_t>(2ULL)],b[cast<int64_t>(2ULL)]);
uint64_t lo1 = std::get<0>(tmp236);
uint64_t hi1 = std::get<1>(tmp236);
auto tmp237 = std::make_tuple(lo0,hi0,lo1,hi1);
c->LO = std::get<0>(tmp237);
c->HI = std::get<1>(tmp237);
c->LO1 = std::get<2>(tmp237);
c->HI1 = std::get<3>(tmp237);
break;}
case cast<uint32_t>(14ULL):{
r5900_CPU_putQ(c,rd,r5900_Quad{s.Hi,t.Hi});
break;}
case cast<uint32_t>(18ULL):{
r5900_CPU_putQ(c,rd,r5900_Quad{cast<uint64_t>((s.Lo | t.Lo)),cast<uint64_t>((s.Hi | t.Hi))});
break;}
case cast<uint32_t>(19ULL):{
r5900_CPU_putQ(c,rd,r5900_Quad{cast<uint64_t>(~(cast<uint64_t>((s.Lo | t.Lo)))),cast<uint64_t>(~(cast<uint64_t>((s.Hi | t.Hi))))});
break;}
case cast<uint32_t>(26ULL):{
std::array<uint16_t,8> h = r5900_halves(t);
r5900_CPU_putH(c,rd,std::array<uint16_t,8>{h[cast<int64_t>(0ULL)],h[cast<int64_t>(2ULL)],h[cast<int64_t>(1ULL)],h[cast<int64_t>(3ULL)],h[cast<int64_t>(4ULL)],h[cast<int64_t>(6ULL)],h[cast<int64_t>(5ULL)],h[cast<int64_t>(7ULL)]});
break;}
case cast<uint32_t>(27ULL):{
std::array<uint16_t,8> h = r5900_halves(t);
r5900_CPU_putH(c,rd,std::array<uint16_t,8>{h[cast<int64_t>(0ULL)],h[cast<int64_t>(0ULL)],h[cast<int64_t>(0ULL)],h[cast<int64_t>(0ULL)],h[cast<int64_t>(4ULL)],h[cast<int64_t>(4ULL)],h[cast<int64_t>(4ULL)],h[cast<int64_t>(4ULL)]});
break;}
case cast<uint32_t>(30ULL):{
std::array<uint32_t,4> v = r5900_words(t);
r5900_CPU_putW(c,rd,std::array<uint32_t,4>{v[cast<int64_t>(0ULL)],v[cast<int64_t>(2ULL)],v[cast<int64_t>(1ULL)],v[cast<int64_t>(3ULL)]});
break;}
default:{
r5900_CPU_Exception(c,cast<uint32_t>(10ULL));
break;}
}}
}
}
// tools/cpu/r5900/mmi.go:729:1
void r5900_CPU_pmulW(r5900_CPU* c,uint32_t rd,r5900_Quad s,r5900_Quad t,bool accumulate,bool subtract,bool signed_){
{
auto tmp238 = std::make_tuple(r5900_words(s),r5900_words(t));
std::array<uint32_t,4> a = std::get<0>(tmp238);
std::array<uint32_t,4> b = std::get<1>(tmp238);
auto prod = [&](int64_t i)->uint64_t{
if (signed_) {
return cast<uint64_t>(cast<int64_t>((cast<int64_t>(cast<int32_t>(a[i])) * cast<int64_t>(cast<int32_t>(b[i])))));
}
return cast<uint64_t>((cast<uint64_t>(a[i]) * cast<uint64_t>(b[i])));
}
;
auto tmp239 = std::make_tuple(prod(cast<int64_t>(0ULL)),prod(cast<int64_t>(2ULL)));
uint64_t p0 = std::get<0>(tmp239);
uint64_t p1 = std::get<1>(tmp239);
if (accumulate) {
uint64_t acc0 = cast<uint64_t>((cast<uint64_t>(cast<uint32_t>(c->LO)) | shl<uint64_t>(cast<uint64_t>(cast<uint32_t>(c->HI)),cast<int64_t>(32ULL))));
uint64_t acc1 = cast<uint64_t>((cast<uint64_t>(cast<uint32_t>(c->LO1)) | shl<uint64_t>(cast<uint64_t>(cast<uint32_t>(c->HI1)),cast<int64_t>(32ULL))));
if (subtract) {
auto tmp240 = std::make_tuple(cast<uint64_t>((acc0 - p0)),cast<uint64_t>((acc1 - p1)));
p0 = std::get<0>(tmp240);
p1 = std::get<1>(tmp240);
}
else {
auto tmp241 = std::make_tuple(cast<uint64_t>((acc0 + p0)),cast<uint64_t>((acc1 + p1)));
p0 = std::get<0>(tmp241);
p1 = std::get<1>(tmp241);
}
}
auto tmp242 = std::make_tuple(r5900_sext32(cast<uint32_t>(p0)),r5900_sext32(cast<uint32_t>(shr<uint64_t>(p0,cast<int64_t>(32ULL)))));
c->LO = std::get<0>(tmp242);
c->HI = std::get<1>(tmp242);
auto tmp243 = std::make_tuple(r5900_sext32(cast<uint32_t>(p1)),r5900_sext32(cast<uint32_t>(shr<uint64_t>(p1,cast<int64_t>(32ULL)))));
c->LO1 = std::get<0>(tmp243);
c->HI1 = std::get<1>(tmp243);
r5900_CPU_putQ(c,rd,r5900_Quad{p0,p1});
}
}
// tools/cpu/r5900/mmi.go:761:1
void r5900_CPU_pmulH(r5900_CPU* c,uint32_t rd,r5900_Quad s,r5900_Quad t,bool accumulate,bool subtract){
{
auto tmp244 = std::make_tuple(r5900_halves(s),r5900_halves(t));
std::array<uint16_t,8> a = std::get<0>(tmp244);
std::array<uint16_t,8> b = std::get<1>(tmp244);
std::array<uint32_t,8> p={};
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(8ULL));i++){
p[i] = cast<uint32_t>(cast<int32_t>((cast<int32_t>(cast<int16_t>(a[i])) * cast<int32_t>(cast<int16_t>(b[i])))));
}
}if (accumulate) {
std::array<uint32_t,8> acc = r5900_CPU_accWords(c);
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(8ULL));i++){
if (subtract) {
p[i] = cast<uint32_t>((acc[i] - p[i]));
}
else {
p[i] = cast<uint32_t>((acc[i] + p[i]));
}
}
}}
r5900_CPU_setAccWords(c,p);
r5900_CPU_putW(c,rd,std::array<uint32_t,4>{p[cast<int64_t>(0ULL)],p[cast<int64_t>(1ULL)],p[cast<int64_t>(4ULL)],p[cast<int64_t>(5ULL)]});
}
}
// tools/cpu/r5900/mmi.go:786:1
void r5900_CPU_phmulH(r5900_CPU* c,uint32_t rd,r5900_Quad s,r5900_Quad t,bool subtract){
{
auto tmp245 = std::make_tuple(r5900_halves(s),r5900_halves(t));
std::array<uint16_t,8> a = std::get<0>(tmp245);
std::array<uint16_t,8> b = std::get<1>(tmp245);
std::array<uint32_t,4> r={};
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(4ULL));i++){
int32_t lo = cast<int32_t>((cast<int32_t>(cast<int16_t>(a[cast<int64_t>((cast<int64_t>(2ULL) * i))])) * cast<int32_t>(cast<int16_t>(b[cast<int64_t>((cast<int64_t>(2ULL) * i))]))));
int32_t hi = cast<int32_t>((cast<int32_t>(cast<int16_t>(a[cast<int64_t>((cast<int64_t>((cast<int64_t>(2ULL) * i)) + cast<int64_t>(1ULL)))])) * cast<int32_t>(cast<int16_t>(b[cast<int64_t>((cast<int64_t>((cast<int64_t>(2ULL) * i)) + cast<int64_t>(1ULL)))]))));
if (subtract) {
r[i] = cast<uint32_t>(cast<int32_t>((hi - lo)));
}
else {
r[i] = cast<uint32_t>(cast<int32_t>((hi + lo)));
}
}
}r5900_CPU_setAccWords(c,std::array<uint32_t,8>{r[cast<int64_t>(0ULL)],r[cast<int64_t>(0ULL)],r[cast<int64_t>(1ULL)],r[cast<int64_t>(1ULL)],r[cast<int64_t>(2ULL)],r[cast<int64_t>(2ULL)],r[cast<int64_t>(3ULL)],r[cast<int64_t>(3ULL)]});
r5900_CPU_putW(c,rd,std::array<uint32_t,4>{r[cast<int64_t>(0ULL)],r[cast<int64_t>(1ULL)],r[cast<int64_t>(2ULL)],r[cast<int64_t>(3ULL)]});
}
}
// tools/cpu/r5900/mmi.go:805:1
std::array<uint32_t,8> r5900_CPU_accWords(r5900_CPU* c){
{
return std::array<uint32_t,8>{cast<uint32_t>(c->LO),cast<uint32_t>(shr<uint64_t>(c->LO,cast<int64_t>(32ULL))),cast<uint32_t>(c->HI),cast<uint32_t>(shr<uint64_t>(c->HI,cast<int64_t>(32ULL))),cast<uint32_t>(c->LO1),cast<uint32_t>(shr<uint64_t>(c->LO1,cast<int64_t>(32ULL))),cast<uint32_t>(c->HI1),cast<uint32_t>(shr<uint64_t>(c->HI1,cast<int64_t>(32ULL)))};
}
}
// tools/cpu/r5900/mmi.go:814:1
void r5900_CPU_setAccWords(r5900_CPU* c,std::array<uint32_t,8> p){
{
c->LO = cast<uint64_t>((cast<uint64_t>(p[cast<int64_t>(0ULL)]) | shl<uint64_t>(cast<uint64_t>(p[cast<int64_t>(1ULL)]),cast<int64_t>(32ULL))));
c->HI = cast<uint64_t>((cast<uint64_t>(p[cast<int64_t>(2ULL)]) | shl<uint64_t>(cast<uint64_t>(p[cast<int64_t>(3ULL)]),cast<int64_t>(32ULL))));
c->LO1 = cast<uint64_t>((cast<uint64_t>(p[cast<int64_t>(4ULL)]) | shl<uint64_t>(cast<uint64_t>(p[cast<int64_t>(5ULL)]),cast<int64_t>(32ULL))));
c->HI1 = cast<uint64_t>((cast<uint64_t>(p[cast<int64_t>(6ULL)]) | shl<uint64_t>(cast<uint64_t>(p[cast<int64_t>(7ULL)]),cast<int64_t>(32ULL))));
}
}
// tools/cpu/r5900/mmi.go:826:1
void r5900_CPU_pmfhl(r5900_CPU* c,uint32_t rd,uint32_t sub_){
{
std::array<uint32_t,8> acc = r5900_CPU_accWords(c);
{
switch(sub_){
case cast<uint32_t>(0ULL):{
r5900_CPU_putW(c,rd,std::array<uint32_t,4>{acc[cast<int64_t>(0ULL)],acc[cast<int64_t>(2ULL)],acc[cast<int64_t>(4ULL)],acc[cast<int64_t>(6ULL)]});
break;}
case cast<uint32_t>(1ULL):{
r5900_CPU_putW(c,rd,std::array<uint32_t,4>{acc[cast<int64_t>(1ULL)],acc[cast<int64_t>(3ULL)],acc[cast<int64_t>(5ULL)],acc[cast<int64_t>(7ULL)]});
break;}
case cast<uint32_t>(2ULL):{
r5900_CPU_putQ(c,rd,r5900_Quad{r5900_satS64(cast<int64_t>((shl<int64_t>(cast<int64_t>(cast<int32_t>(acc[cast<int64_t>(2ULL)])),cast<int64_t>(32ULL)) | cast<int64_t>(cast<uint32_t>(acc[cast<int64_t>(0ULL)]))))),r5900_satS64(cast<int64_t>((shl<int64_t>(cast<int64_t>(cast<int32_t>(acc[cast<int64_t>(6ULL)])),cast<int64_t>(32ULL)) | cast<int64_t>(cast<uint32_t>(acc[cast<int64_t>(4ULL)])))))});
break;}
case cast<uint32_t>(3ULL):{
std::array<uint16_t,8> h={};
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(8ULL));i++){
h[i] = cast<uint16_t>(acc[i]);
}
}r5900_CPU_putH(c,rd,h);
break;}
case cast<uint32_t>(4ULL):{
std::array<uint16_t,8> h={};
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(8ULL));i++){
h[i] = r5900_satS16(cast<int32_t>(acc[i]));
}
}r5900_CPU_putH(c,rd,h);
break;}
default:{
r5900_CPU_Exception(c,cast<uint32_t>(10ULL));
break;}
}}
}
}
// tools/cpu/r5900/mmi.go:857:1
void r5900_CPU_pmthl(r5900_CPU* c,uint32_t rs,uint32_t sub_){
{
if ((sub_ != cast<uint32_t>(0ULL))) {
r5900_CPU_Exception(c,cast<uint32_t>(10ULL));
return ;
}
std::array<uint32_t,4> v = r5900_words(c->R[rs]);
std::array<uint32_t,8> acc = r5900_CPU_accWords(c);
auto tmp246 = std::make_tuple(v[cast<int64_t>(0ULL)],v[cast<int64_t>(1ULL)],v[cast<int64_t>(2ULL)],v[cast<int64_t>(3ULL)]);
acc[cast<int64_t>(0ULL)] = std::get<0>(tmp246);
acc[cast<int64_t>(2ULL)] = std::get<1>(tmp246);
acc[cast<int64_t>(4ULL)] = std::get<2>(tmp246);
acc[cast<int64_t>(6ULL)] = std::get<3>(tmp246);
r5900_CPU_setAccWords(c,acc);
}
}
// tools/cpu/r5900/mmi.go:870:1
uint64_t r5900_satS64(int64_t v){
{
return cast<uint64_t>(v);
}
}
// tools/cpu/r5900/mmi.go:876:1
uint32_t r5900_maskIf(bool b){
{
if (b) {
return cast<uint32_t>(4294967295ULL);
}
return cast<uint32_t>(0ULL);
}
}
// tools/cpu/r5900/mmi.go:883:1
uint32_t r5900_absW(uint32_t v){
{
if ((v == cast<uint32_t>(2147483648ULL))) {
return cast<uint32_t>(2147483647ULL);
}
if ((cast<int32_t>(v) < cast<int32_t>(0ULL))) {
return cast<uint32_t>(cast<int32_t>(-cast<int32_t>(v)));
}
return v;
}
}
// tools/cpu/r5900/mmi.go:893:1
uint16_t r5900_absH(uint16_t v){
{
if ((v == cast<uint16_t>(32768ULL))) {
return cast<uint16_t>(32767ULL);
}
if ((cast<int16_t>(v) < cast<int16_t>(0ULL))) {
return cast<uint16_t>(cast<int16_t>(-cast<int16_t>(v)));
}
return v;
}
}
// tools/cpu/r5900/mmi.go:905:1
r5900_Quad r5900_funnelRight(r5900_Quad hi,r5900_Quad lo,uint32_t n){
{
if ((n == cast<uint32_t>(0ULL))) {
return lo;
}
if ((n >= cast<uint32_t>(128ULL))) {
return r5900_shiftRight128(hi,cast<uint32_t>((n - cast<uint32_t>(128ULL))));
}
r5900_Quad r = r5900_shiftRight128(lo,n);
r5900_Quad l = r5900_shiftLeft128(hi,cast<uint32_t>((cast<uint32_t>(128ULL) - n)));
return r5900_Quad{cast<uint64_t>((r.Lo | l.Lo)),cast<uint64_t>((r.Hi | l.Hi))};
}
}
// tools/cpu/r5900/mmi.go:918:1
r5900_Quad r5900_shiftRight128(r5900_Quad q,uint32_t n){
{
{
if ((n == cast<uint32_t>(0ULL))){
return q;
}
else if ((n >= cast<uint32_t>(128ULL))){
return r5900_Quad{};
}
else if ((n >= cast<uint32_t>(64ULL))){
return r5900_Quad{shr<uint64_t>(q.Hi,(cast<uint32_t>((n - cast<uint32_t>(64ULL))))),{}};
}
else {
return r5900_Quad{cast<uint64_t>((shr<uint64_t>(q.Lo,n) | shl<uint64_t>(q.Hi,(cast<uint32_t>((cast<uint32_t>(64ULL) - n)))))),shr<uint64_t>(q.Hi,n)};
}
}
tmp247:;
}
}
// tools/cpu/r5900/mmi.go:931:1
r5900_Quad r5900_shiftLeft128(r5900_Quad q,uint32_t n){
{
{
if ((n == cast<uint32_t>(0ULL))){
return q;
}
else if ((n >= cast<uint32_t>(128ULL))){
return r5900_Quad{};
}
else if ((n >= cast<uint32_t>(64ULL))){
return r5900_Quad{{},shl<uint64_t>(q.Lo,(cast<uint32_t>((n - cast<uint32_t>(64ULL)))))};
}
else {
return r5900_Quad{shl<uint64_t>(q.Lo,n),cast<uint64_t>((shl<uint64_t>(q.Hi,n) | shr<uint64_t>(q.Lo,(cast<uint32_t>((cast<uint32_t>(64ULL) - n))))))};
}
}
tmp248:;
}
}
// tools/cpu/r5900/mmi.go:945:1
void r5900_CPU_putW(r5900_CPU* c,uint32_t rd,std::array<uint32_t,4> v){
{
r5900_CPU_putQ(c,rd,r5900_fromWords(v));
}
}
// tools/cpu/r5900/mmi.go:946:1
void r5900_CPU_putH(r5900_CPU* c,uint32_t rd,std::array<uint16_t,8> v){
{
r5900_CPU_putQ(c,rd,r5900_fromHalves(v));
}
}
// tools/cpu/r5900/mmi.go:947:1
void r5900_CPU_putB(r5900_CPU* c,uint32_t rd,std::array<uint8_t,16> v){
{
r5900_CPU_putQ(c,rd,r5900_fromOctets(v));
}
}
// tools/cpu/r5900/mmi.go:949:1
void r5900_CPU_putQ(r5900_CPU* c,uint32_t rd,r5900_Quad q){
{
r5900_CPU_setQ(c,rd,q.Lo,q.Hi);
}
}
// tools/cpu/r5900/mmi_decode.go:72:1
r5900_Inst r5900_decodeMMI(r5900_Inst in,uint32_t w,uint32_t rs,uint32_t rt,uint32_t rd,uint32_t shamt){
{
auto set = [&](std::string mnem,std::string text)->r5900_Inst{
auto tmp249 = std::make_tuple(mnem,text);
in.Mnem = std::get<0>(tmp249);
in.Text = std::get<1>(tmp249);
return in;
}
;
{
switch(cast<uint32_t>((w & cast<uint32_t>(63ULL)))){
case cast<uint32_t>(0ULL):{
return set(std::string("madd",4),go_fmt_Sprintf(std::string("madd %s, %s, %s",15),r5900_reg(rd),r5900_reg(rs),r5900_reg(rt)));
break;}
case cast<uint32_t>(1ULL):{
return set(std::string("maddu",5),go_fmt_Sprintf(std::string("maddu %s, %s, %s",16),r5900_reg(rd),r5900_reg(rs),r5900_reg(rt)));
break;}
case cast<uint32_t>(4ULL):{
return set(std::string("plzcw",5),go_fmt_Sprintf(std::string("plzcw %s, %s",12),r5900_reg(rd),r5900_reg(rs)));
break;}
case cast<uint32_t>(8ULL):{
return r5900_mmiSub(in,w,r5900_mmi0Tab,rs,rt,rd,shamt);
break;}
case cast<uint32_t>(9ULL):{
return r5900_mmiSub(in,w,r5900_mmi2Tab,rs,rt,rd,shamt);
break;}
case cast<uint32_t>(16ULL):{
return set(std::string("mfhi1",5),go_fmt_Sprintf(std::string("mfhi1 %s",8),r5900_reg(rd)));
break;}
case cast<uint32_t>(17ULL):{
return set(std::string("mthi1",5),go_fmt_Sprintf(std::string("mthi1 %s",8),r5900_reg(rs)));
break;}
case cast<uint32_t>(18ULL):{
return set(std::string("mflo1",5),go_fmt_Sprintf(std::string("mflo1 %s",8),r5900_reg(rd)));
break;}
case cast<uint32_t>(19ULL):{
return set(std::string("mtlo1",5),go_fmt_Sprintf(std::string("mtlo1 %s",8),r5900_reg(rs)));
break;}
case cast<uint32_t>(24ULL):{
return set(std::string("mult1",5),go_fmt_Sprintf(std::string("mult1 %s, %s, %s",16),r5900_reg(rd),r5900_reg(rs),r5900_reg(rt)));
break;}
case cast<uint32_t>(25ULL):{
return set(std::string("multu1",6),go_fmt_Sprintf(std::string("multu1 %s, %s, %s",17),r5900_reg(rd),r5900_reg(rs),r5900_reg(rt)));
break;}
case cast<uint32_t>(26ULL):{
return set(std::string("div1",4),go_fmt_Sprintf(std::string("div1 %s, %s",11),r5900_reg(rs),r5900_reg(rt)));
break;}
case cast<uint32_t>(27ULL):{
return set(std::string("divu1",5),go_fmt_Sprintf(std::string("divu1 %s, %s",12),r5900_reg(rs),r5900_reg(rt)));
break;}
case cast<uint32_t>(32ULL):{
return set(std::string("madd1",5),go_fmt_Sprintf(std::string("madd1 %s, %s, %s",16),r5900_reg(rd),r5900_reg(rs),r5900_reg(rt)));
break;}
case cast<uint32_t>(33ULL):{
return set(std::string("maddu1",6),go_fmt_Sprintf(std::string("maddu1 %s, %s, %s",17),r5900_reg(rd),r5900_reg(rs),r5900_reg(rt)));
break;}
case cast<uint32_t>(40ULL):{
return r5900_mmiSub(in,w,r5900_mmi1Tab,rs,rt,rd,shamt);
break;}
case cast<uint32_t>(41ULL):{
return r5900_mmiSub(in,w,r5900_mmi3Tab,rs,rt,rd,shamt);
break;}
case cast<uint32_t>(48ULL):{
if (((shamt < cast<uint32_t>(8ULL)) && (r5900_pmfhlSub[shamt] != std::string("",0)))) {
std::string m = (std::string("pmfhl.",6) + r5900_pmfhlSub[shamt]);
return set(m,go_fmt_Sprintf(std::string("%s %s",5),m,r5900_reg(rd)));
}
return r5900_word(in,w);
break;}
case cast<uint32_t>(49ULL):{
if ((shamt == cast<uint32_t>(0ULL))) {
return set(std::string("pmthl.lw",8),go_fmt_Sprintf(std::string("pmthl.lw %s",11),r5900_reg(rs)));
}
return r5900_word(in,w);
break;}
case cast<uint32_t>(52ULL):case cast<uint32_t>(54ULL):case cast<uint32_t>(55ULL):case cast<uint32_t>(60ULL):case cast<uint32_t>(62ULL):case cast<uint32_t>(63ULL):{
std::string m = get(Map<uint32_t,std::string>{{cast<uint32_t>(52ULL),std::string("psllh",5)},{cast<uint32_t>(54ULL),std::string("psrlh",5)},{cast<uint32_t>(55ULL),std::string("psrah",5)},{cast<uint32_t>(60ULL),std::string("psllw",5)},{cast<uint32_t>(62ULL),std::string("psrlw",5)},{cast<uint32_t>(63ULL),std::string("psraw",5)}},cast<uint32_t>((w & cast<uint32_t>(63ULL))));
return set(m,go_fmt_Sprintf(std::string("%s %s, %s, %d",13),m,r5900_reg(rd),r5900_reg(rt),shamt));
break;}
}}
return r5900_word(in,w);
}
}
// tools/cpu/r5900/mmi_decode.go:133:1
r5900_Inst r5900_mmiSub(r5900_Inst in,uint32_t w,Map<uint32_t,r5900_mmiOp> tab,uint32_t rs,uint32_t rt,uint32_t rd,uint32_t shamt){
{
auto tmp250 = lookup(tab,shamt);
r5900_mmiOp o = std::get<0>(tmp250);
bool ok = std::get<1>(tmp250);
if ((!ok)) {
return r5900_word(in,w);
}
in.Mnem = o.name;
{
switch(o.form){
case cast<r5900_mmiForm>(0ULL):{
in.Text = go_fmt_Sprintf(std::string("%s %s, %s, %s",13),o.name,r5900_reg(rd),r5900_reg(rs),r5900_reg(rt));
break;}
case cast<r5900_mmiForm>(1ULL):{
in.Text = go_fmt_Sprintf(std::string("%s %s, %s",9),o.name,r5900_reg(rd),r5900_reg(rt));
break;}
case cast<r5900_mmiForm>(2ULL):{
in.Text = go_fmt_Sprintf(std::string("%s %s, %s, %s",13),o.name,r5900_reg(rd),r5900_reg(rt),r5900_reg(rs));
break;}
case cast<r5900_mmiForm>(3ULL):{
in.Text = go_fmt_Sprintf(std::string("%s %s, %s, %d",13),o.name,r5900_reg(rd),r5900_reg(rt),shamt);
break;}
case cast<r5900_mmiForm>(4ULL):{
in.Text = go_fmt_Sprintf(std::string("%s %s",5),o.name,r5900_reg(rd));
break;}
case cast<r5900_mmiForm>(5ULL):{
in.Text = go_fmt_Sprintf(std::string("%s %s",5),o.name,r5900_reg(rs));
break;}
case cast<r5900_mmiForm>(6ULL):{
in.Text = go_fmt_Sprintf(std::string("%s %s, %s",9),o.name,r5900_reg(rs),r5900_reg(rt));
break;}
}}
return in;
}
}
// tools/cpu/r5900/r5900.go:44:1
std::string r5900_Quad_String(r5900_Quad q){
{
return go_fmt_Sprintf(std::string("%016X_%016X",11),q.Hi,q.Lo);
}
}
// tools/cpu/r5900/r5900.go:62:1
std::string r5900_Flow_String(r5900_Flow f){
{
{
switch(f){
case cast<r5900_Flow>(0ULL):{
return std::string("seq",3);
break;}
case cast<r5900_Flow>(1ULL):{
return std::string("branch",6);
break;}
case cast<r5900_Flow>(2ULL):{
return std::string("jump",4);
break;}
case cast<r5900_Flow>(3ULL):{
return std::string("call",4);
break;}
case cast<r5900_Flow>(4ULL):{
return std::string("return",6);
break;}
case cast<r5900_Flow>(5ULL):{
return std::string("indjump",7);
break;}
case cast<r5900_Flow>(6ULL):{
return std::string("indcall",7);
break;}
case cast<r5900_Flow>(7ULL):{
return std::string("stop",4);
break;}
}}
return std::string("?",1);
}
}
// tools/cpu/r5900/r5900.go:102:1
std::string r5900_Inst_String(r5900_Inst in){
{
return go_fmt_Sprintf(std::string("$%08X: %s",9),in.Addr,in.Text);
}
}
// tools/cpu/r5900/state.go:42:1
r5900_State r5900_CPU_Snapshot(r5900_CPU* c){
{
return r5900_State{c->R,c->HI,c->LO,c->HI1,c->LO1,c->SA,c->PC,c->nextPC,c->COP0,c->TLB,c->FPR,c->ACC,c->FCR31,c->LLBit,c->Halted,c->HaltReason,c->Steps,c->curPC,c->delaySlot,c->pendingDelay,c->branchAddr,c->countFrac};
}
}
// tools/cpu/r5900/state.go:60:1
void r5900_CPU_SetThreadRegs(r5900_CPU* c,r5900_State s){
{
auto tmp251 = std::make_tuple(s.R,s.HI,s.LO,s.HI1,s.LO1,s.SA);
c->R = std::get<0>(tmp251);
c->HI = std::get<1>(tmp251);
c->LO = std::get<2>(tmp251);
c->HI1 = std::get<3>(tmp251);
c->LO1 = std::get<4>(tmp251);
c->SA = std::get<5>(tmp251);
auto tmp252 = std::make_tuple(s.PC,s.NextPC,s.CurPC);
c->PC = std::get<0>(tmp252);
c->nextPC = std::get<1>(tmp252);
c->curPC = std::get<2>(tmp252);
auto tmp253 = std::make_tuple(s.FPR,s.ACC,s.FCR31,s.LLBit);
c->FPR = std::get<0>(tmp253);
c->ACC = std::get<1>(tmp253);
c->FCR31 = std::get<2>(tmp253);
c->LLBit = std::get<3>(tmp253);
auto tmp254 = std::make_tuple(s.DelaySlot,s.PendingDelay,s.BranchAddr);
c->delaySlot = std::get<0>(tmp254);
c->pendingDelay = std::get<1>(tmp254);
c->branchAddr = std::get<2>(tmp254);
}
}
// tools/cpu/r5900/state.go:68:1
void r5900_CPU_Restore(r5900_CPU* c,r5900_State s){
{
auto tmp255 = std::make_tuple(s.R,s.HI,s.LO,s.HI1,s.LO1,s.SA);
c->R = std::get<0>(tmp255);
c->HI = std::get<1>(tmp255);
c->LO = std::get<2>(tmp255);
c->HI1 = std::get<3>(tmp255);
c->LO1 = std::get<4>(tmp255);
c->SA = std::get<5>(tmp255);
auto tmp256 = std::make_tuple(s.PC,s.NextPC);
c->PC = std::get<0>(tmp256);
c->nextPC = std::get<1>(tmp256);
auto tmp257 = std::make_tuple(s.COP0,s.TLB);
c->COP0 = std::get<0>(tmp257);
c->TLB = std::get<1>(tmp257);
auto tmp258 = std::make_tuple(s.FPR,s.ACC,s.FCR31,s.LLBit);
c->FPR = std::get<0>(tmp258);
c->ACC = std::get<1>(tmp258);
c->FCR31 = std::get<2>(tmp258);
c->LLBit = std::get<3>(tmp258);
auto tmp259 = std::make_tuple(s.Halted,s.HaltReason,s.Steps);
c->Halted = std::get<0>(tmp259);
c->HaltReason = std::get<1>(tmp259);
c->Steps = std::get<2>(tmp259);
auto tmp260 = std::make_tuple(s.CurPC,s.DelaySlot,s.PendingDelay);
c->curPC = std::get<0>(tmp260);
c->delaySlot = std::get<1>(tmp260);
c->pendingDelay = std::get<2>(tmp260);
auto tmp261 = std::make_tuple(s.BranchAddr,s.CountFrac);
c->branchAddr = std::get<0>(tmp261);
c->countFrac = std::get<1>(tmp261);
}
}
// tools/cpu/mips/cpu.go:115:1
void mips_CPU_Reset(mips_CPU* c){
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
c->COP0[cast<int64_t>(15ULL)] = cast<uint32_t>(2ULL);
c->ld = mips_loadSlot{};
auto tmp3 = std::make_tuple(false,false);
c->delaySlot = std::get<0>(tmp3);
c->pendingDelay = std::get<1>(tmp3);
auto tmp4 = std::make_tuple(false,std::string("",0));
c->Halted = std::get<0>(tmp4);
c->HaltReason = std::get<1>(tmp4);
}
}
// tools/cpu/mips/cpu.go:135:1
void mips_CPU_SetPC(mips_CPU* c,uint32_t pc){
{
auto tmp5 = std::make_tuple(pc,cast<uint32_t>((pc + cast<uint32_t>(4ULL))));
c->PC = std::get<0>(tmp5);
c->nextPC = std::get<1>(tmp5);
c->pendingDelay = false;
}
}
// tools/cpu/mips/cpu.go:142:1
void mips_CPU_SetReg(mips_CPU* c,uint32_t i,uint32_t v){
{
if ((i != cast<uint32_t>(0ULL))) {
auto tmp6 = std::make_tuple(v,v);
c->R[i] = std::get<0>(tmp6);
c->out[i] = std::get<1>(tmp6);
}
}
}
// tools/cpu/mips/cpu.go:149:1
uint32_t mips_CPU_Reg(mips_CPU* c,uint32_t i){
{
return c->R[i];
}
}
// tools/cpu/mips/cpu.go:153:1
uint32_t mips_CPU_CurPC(mips_CPU* c){
{
return c->curPC;
}
}
// tools/cpu/mips/cpu.go:158:1
void mips_CPU_set(mips_CPU* c,uint32_t i,uint32_t v){
{
if ((i != cast<uint32_t>(0ULL))) {
c->out[i] = v;
}
}
}
// tools/cpu/mips/cpu.go:166:1
uint32_t mips_CPU_read8(mips_CPU* c,uint32_t a){
{
return cast<uint32_t>(ps2_IOP_Read(c->bus,a));
}
}
// tools/cpu/mips/cpu.go:167:1
void mips_CPU_write8(mips_CPU* c,uint32_t a,uint32_t v){
{
ps2_IOP_Write(c->bus,a,cast<uint8_t>(v));
}
}
// tools/cpu/mips/cpu.go:168:1
uint32_t mips_CPU_read16(mips_CPU* c,uint32_t a){
{
return cast<uint32_t>((cast<uint32_t>(ps2_IOP_Read(c->bus,a)) | shl<uint32_t>(cast<uint32_t>(ps2_IOP_Read(c->bus,cast<uint32_t>((a + cast<uint32_t>(1ULL))))),cast<int64_t>(8ULL))));
}
}
// tools/cpu/mips/cpu.go:169:1
void mips_CPU_write16(mips_CPU* c,uint32_t a,uint32_t v){
{
ps2_IOP_Write(c->bus,a,cast<uint8_t>(v));
ps2_IOP_Write(c->bus,cast<uint32_t>((a + cast<uint32_t>(1ULL))),cast<uint8_t>(shr<uint32_t>(v,cast<int64_t>(8ULL))));
}
}
// tools/cpu/mips/cpu.go:173:1
uint32_t mips_CPU_read32(mips_CPU* c,uint32_t a){
{
return cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(ps2_IOP_Read(c->bus,a)) | shl<uint32_t>(cast<uint32_t>(ps2_IOP_Read(c->bus,cast<uint32_t>((a + cast<uint32_t>(1ULL))))),cast<int64_t>(8ULL)))) | shl<uint32_t>(cast<uint32_t>(ps2_IOP_Read(c->bus,cast<uint32_t>((a + cast<uint32_t>(2ULL))))),cast<int64_t>(16ULL)))) | shl<uint32_t>(cast<uint32_t>(ps2_IOP_Read(c->bus,cast<uint32_t>((a + cast<uint32_t>(3ULL))))),cast<int64_t>(24ULL))));
}
}
// tools/cpu/mips/cpu.go:176:1
void mips_CPU_write32(mips_CPU* c,uint32_t a,uint32_t v){
{
ps2_IOP_Write(c->bus,a,cast<uint8_t>(v));
ps2_IOP_Write(c->bus,cast<uint32_t>((a + cast<uint32_t>(1ULL))),cast<uint8_t>(shr<uint32_t>(v,cast<int64_t>(8ULL))));
ps2_IOP_Write(c->bus,cast<uint32_t>((a + cast<uint32_t>(2ULL))),cast<uint8_t>(shr<uint32_t>(v,cast<int64_t>(16ULL))));
ps2_IOP_Write(c->bus,cast<uint32_t>((a + cast<uint32_t>(3ULL))),cast<uint8_t>(shr<uint32_t>(v,cast<int64_t>(24ULL))));
}
}
// tools/cpu/mips/cpu.go:189:1
void mips_CPU_Exception(mips_CPU* c,uint32_t code){
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
mips_CPU_SetPC(c,target);
}
}
// tools/cpu/mips/cpu.go:211:1
void mips_CPU_rfe(mips_CPU* c){
{
uint32_t sr = c->COP0[cast<int64_t>(12ULL)];
c->COP0[cast<int64_t>(12ULL)] = cast<uint32_t>((((sr & ~(cast<uint32_t>(15ULL)))) | (cast<uint32_t>(((shr<uint32_t>(sr,cast<int64_t>(2ULL))) & cast<uint32_t>(15ULL))))));
}
}
// tools/cpu/mips/cpu.go:221:1
bool mips_CPU_Interrupt(mips_CPU* c,bool pending){
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
c->ld = mips_loadSlot{};
}
c->curPC = c->PC;
c->delaySlot = false;
mips_CPU_Exception(c,cast<uint32_t>(0ULL));
return true;
}
}
// tools/cpu/mips/cpu.go:254:1
void mips_CPU_addrError(mips_CPU* c,uint32_t code,uint32_t addr){
{
c->COP0[cast<int64_t>(8ULL)] = addr;
mips_CPU_Exception(c,code);
}
}
// tools/cpu/mips/decode.go:16:1
std::string mips_reg(uint32_t i){
{
return (std::string("$",1) + mips_regName[cast<uint32_t>((i & cast<uint32_t>(31ULL)))]);
}
}
// tools/cpu/mips/decode.go:31:1
mips_Inst mips_Decode(Slice<uint8_t> code,uint32_t addr){
{
if ((len(code) < cast<int64_t>(4ULL))) {
return mips_Inst{addr,cast<int64_t>(0ULL),std::string(".word",5),std::string(".word <truncated>",17),cast<mips_Flow>(7ULL),{},{},{}};
}
uint32_t w = le_Uint32(code);
mips_Inst in = mips_Inst{addr,cast<int64_t>(4ULL),{},{},cast<mips_Flow>(0ULL),{},{},{}};
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
return mips_decodeSpecial(in,w,rs,rt,rd,shamt,funct);
break;}
case cast<uint32_t>(1ULL):{
auto tmp8 = std::make_tuple(cast<mips_Flow>(1ULL),branchT,true,true);
in.Flow = std::get<0>(tmp8);
in.Target = std::get<1>(tmp8);
in.HasTarget = std::get<2>(tmp8);
in.HasDelay = std::get<3>(tmp8);
{
switch(rt){
case cast<uint32_t>(0ULL):{
set(std::string("bltz",4),go_fmt_Sprintf(std::string("bltz %s, $%08X",14),mips_reg(rs),branchT));
break;}
case cast<uint32_t>(1ULL):{
set(std::string("bgez",4),go_fmt_Sprintf(std::string("bgez %s, $%08X",14),mips_reg(rs),branchT));
break;}
case cast<uint32_t>(16ULL):{
set(std::string("bltzal",6),go_fmt_Sprintf(std::string("bltzal %s, $%08X",16),mips_reg(rs),branchT));
break;}
case cast<uint32_t>(17ULL):{
set(std::string("bgezal",6),go_fmt_Sprintf(std::string("bgezal %s, $%08X",16),mips_reg(rs),branchT));
break;}
default:{
return mips_word(in,w);
break;}
}}
break;}
case cast<uint32_t>(2ULL):{
auto tmp9 = std::make_tuple(cast<mips_Flow>(2ULL),jumpT,true,true);
in.Flow = std::get<0>(tmp9);
in.Target = std::get<1>(tmp9);
in.HasTarget = std::get<2>(tmp9);
in.HasDelay = std::get<3>(tmp9);
set(std::string("j",1),go_fmt_Sprintf(std::string("j $%08X",7),jumpT));
break;}
case cast<uint32_t>(3ULL):{
auto tmp10 = std::make_tuple(cast<mips_Flow>(3ULL),jumpT,true,true);
in.Flow = std::get<0>(tmp10);
in.Target = std::get<1>(tmp10);
in.HasTarget = std::get<2>(tmp10);
in.HasDelay = std::get<3>(tmp10);
set(std::string("jal",3),go_fmt_Sprintf(std::string("jal $%08X",9),jumpT));
break;}
case cast<uint32_t>(4ULL):{
auto tmp11 = std::make_tuple(cast<mips_Flow>(1ULL),branchT,true,true);
in.Flow = std::get<0>(tmp11);
in.Target = std::get<1>(tmp11);
in.HasTarget = std::get<2>(tmp11);
in.HasDelay = std::get<3>(tmp11);
if (((rs == cast<uint32_t>(0ULL)) && (rt == cast<uint32_t>(0ULL)))) {
set(std::string("b",1),go_fmt_Sprintf(std::string("b $%08X",7),branchT));
in.Flow = cast<mips_Flow>(2ULL);
}
else {
set(std::string("beq",3),go_fmt_Sprintf(std::string("beq %s, %s, $%08X",17),mips_reg(rs),mips_reg(rt),branchT));
}
break;}
case cast<uint32_t>(5ULL):{
auto tmp12 = std::make_tuple(cast<mips_Flow>(1ULL),branchT,true,true);
in.Flow = std::get<0>(tmp12);
in.Target = std::get<1>(tmp12);
in.HasTarget = std::get<2>(tmp12);
in.HasDelay = std::get<3>(tmp12);
set(std::string("bne",3),go_fmt_Sprintf(std::string("bne %s, %s, $%08X",17),mips_reg(rs),mips_reg(rt),branchT));
break;}
case cast<uint32_t>(6ULL):{
auto tmp13 = std::make_tuple(cast<mips_Flow>(1ULL),branchT,true,true);
in.Flow = std::get<0>(tmp13);
in.Target = std::get<1>(tmp13);
in.HasTarget = std::get<2>(tmp13);
in.HasDelay = std::get<3>(tmp13);
set(std::string("blez",4),go_fmt_Sprintf(std::string("blez %s, $%08X",14),mips_reg(rs),branchT));
break;}
case cast<uint32_t>(7ULL):{
auto tmp14 = std::make_tuple(cast<mips_Flow>(1ULL),branchT,true,true);
in.Flow = std::get<0>(tmp14);
in.Target = std::get<1>(tmp14);
in.HasTarget = std::get<2>(tmp14);
in.HasDelay = std::get<3>(tmp14);
set(std::string("bgtz",4),go_fmt_Sprintf(std::string("bgtz %s, $%08X",14),mips_reg(rs),branchT));
break;}
case cast<uint32_t>(8ULL):{
set(std::string("addi",4),go_fmt_Sprintf(std::string("addi %s, %s, %d",15),mips_reg(rt),mips_reg(rs),simm));
break;}
case cast<uint32_t>(9ULL):{
set(std::string("addiu",5),go_fmt_Sprintf(std::string("addiu %s, %s, %d",16),mips_reg(rt),mips_reg(rs),simm));
break;}
case cast<uint32_t>(10ULL):{
set(std::string("slti",4),go_fmt_Sprintf(std::string("slti %s, %s, %d",15),mips_reg(rt),mips_reg(rs),simm));
break;}
case cast<uint32_t>(11ULL):{
set(std::string("sltiu",5),go_fmt_Sprintf(std::string("sltiu %s, %s, %d",16),mips_reg(rt),mips_reg(rs),simm));
break;}
case cast<uint32_t>(12ULL):{
set(std::string("andi",4),go_fmt_Sprintf(std::string("andi %s, %s, 0x%X",17),mips_reg(rt),mips_reg(rs),imm));
break;}
case cast<uint32_t>(13ULL):{
set(std::string("ori",3),go_fmt_Sprintf(std::string("ori %s, %s, 0x%X",16),mips_reg(rt),mips_reg(rs),imm));
break;}
case cast<uint32_t>(14ULL):{
set(std::string("xori",4),go_fmt_Sprintf(std::string("xori %s, %s, 0x%X",17),mips_reg(rt),mips_reg(rs),imm));
break;}
case cast<uint32_t>(15ULL):{
set(std::string("lui",3),go_fmt_Sprintf(std::string("lui %s, 0x%X",12),mips_reg(rt),imm));
break;}
case cast<uint32_t>(16ULL):{
return mips_decodeCop0(in,w,rs,rt,rd);
break;}
case cast<uint32_t>(18ULL):{
return mips_decodeCop2(in,w,rs,rt,rd);
break;}
case cast<uint32_t>(32ULL):case cast<uint32_t>(33ULL):case cast<uint32_t>(34ULL):case cast<uint32_t>(35ULL):case cast<uint32_t>(36ULL):case cast<uint32_t>(37ULL):case cast<uint32_t>(38ULL):{
std::string m = [](){std::array<std::string,39> v{};v[32]=std::string("lb",2);v[33]=std::string("lh",2);v[34]=std::string("lwl",3);v[35]=std::string("lw",2);v[36]=std::string("lbu",3);v[37]=std::string("lhu",3);v[38]=std::string("lwr",3);return v;}()[op];
set(m,go_fmt_Sprintf(std::string("%s %s, %d(%s)",13),m,mips_reg(rt),simm,mips_reg(rs)));
break;}
case cast<uint32_t>(40ULL):case cast<uint32_t>(41ULL):case cast<uint32_t>(42ULL):case cast<uint32_t>(43ULL):case cast<uint32_t>(46ULL):{
std::string m = get(Map<uint32_t,std::string>{{cast<uint32_t>(40ULL),std::string("sb",2)},{cast<uint32_t>(41ULL),std::string("sh",2)},{cast<uint32_t>(42ULL),std::string("swl",3)},{cast<uint32_t>(43ULL),std::string("sw",2)},{cast<uint32_t>(46ULL),std::string("swr",3)}},op);
set(m,go_fmt_Sprintf(std::string("%s %s, %d(%s)",13),m,mips_reg(rt),simm,mips_reg(rs)));
break;}
case cast<uint32_t>(50ULL):{
set(std::string("lwc2",4),go_fmt_Sprintf(std::string("lwc2 $%d, %d(%s)",16),rt,simm,mips_reg(rs)));
break;}
case cast<uint32_t>(58ULL):{
set(std::string("swc2",4),go_fmt_Sprintf(std::string("swc2 $%d, %d(%s)",16),rt,simm,mips_reg(rs)));
break;}
default:{
return mips_word(in,w);
break;}
}}
return in;
}
}
// tools/cpu/mips/decode.go:136:1
mips_Inst mips_decodeSpecial(mips_Inst in,uint32_t w,uint32_t rs,uint32_t rt,uint32_t rd,uint32_t shamt,uint32_t funct){
{
auto set = [&](std::string mnem,std::string text)->mips_Inst{
auto tmp15 = std::make_tuple(mnem,text);
in.Mnem = std::get<0>(tmp15);
in.Text = std::get<1>(tmp15);
return in;
}
;
{
switch(funct){
case cast<uint32_t>(0ULL):{
if ((w == cast<uint32_t>(0ULL))) {
return set(std::string("nop",3),std::string("nop",3));
}
return set(std::string("sll",3),go_fmt_Sprintf(std::string("sll %s, %s, %d",14),mips_reg(rd),mips_reg(rt),shamt));
break;}
case cast<uint32_t>(2ULL):{
return set(std::string("srl",3),go_fmt_Sprintf(std::string("srl %s, %s, %d",14),mips_reg(rd),mips_reg(rt),shamt));
break;}
case cast<uint32_t>(3ULL):{
return set(std::string("sra",3),go_fmt_Sprintf(std::string("sra %s, %s, %d",14),mips_reg(rd),mips_reg(rt),shamt));
break;}
case cast<uint32_t>(4ULL):{
return set(std::string("sllv",4),go_fmt_Sprintf(std::string("sllv %s, %s, %s",15),mips_reg(rd),mips_reg(rt),mips_reg(rs)));
break;}
case cast<uint32_t>(6ULL):{
return set(std::string("srlv",4),go_fmt_Sprintf(std::string("srlv %s, %s, %s",15),mips_reg(rd),mips_reg(rt),mips_reg(rs)));
break;}
case cast<uint32_t>(7ULL):{
return set(std::string("srav",4),go_fmt_Sprintf(std::string("srav %s, %s, %s",15),mips_reg(rd),mips_reg(rt),mips_reg(rs)));
break;}
case cast<uint32_t>(8ULL):{
in.HasDelay = true;
if ((rs == cast<uint32_t>(31ULL))) {
in.Flow = cast<mips_Flow>(4ULL);
}
else {
in.Flow = cast<mips_Flow>(5ULL);
}
return set(std::string("jr",2),go_fmt_Sprintf(std::string("jr %s",5),mips_reg(rs)));
break;}
case cast<uint32_t>(9ULL):{
auto tmp16 = std::make_tuple(cast<mips_Flow>(6ULL),true);
in.Flow = std::get<0>(tmp16);
in.HasDelay = std::get<1>(tmp16);
if ((rd == cast<uint32_t>(31ULL))) {
return set(std::string("jalr",4),go_fmt_Sprintf(std::string("jalr %s",7),mips_reg(rs)));
}
return set(std::string("jalr",4),go_fmt_Sprintf(std::string("jalr %s, %s",11),mips_reg(rd),mips_reg(rs)));
break;}
case cast<uint32_t>(12ULL):{
in.Flow = cast<mips_Flow>(7ULL);
return set(std::string("syscall",7),std::string("syscall",7));
break;}
case cast<uint32_t>(13ULL):{
in.Flow = cast<mips_Flow>(7ULL);
return set(std::string("break",5),std::string("break",5));
break;}
case cast<uint32_t>(16ULL):{
return set(std::string("mfhi",4),go_fmt_Sprintf(std::string("mfhi %s",7),mips_reg(rd)));
break;}
case cast<uint32_t>(17ULL):{
return set(std::string("mthi",4),go_fmt_Sprintf(std::string("mthi %s",7),mips_reg(rs)));
break;}
case cast<uint32_t>(18ULL):{
return set(std::string("mflo",4),go_fmt_Sprintf(std::string("mflo %s",7),mips_reg(rd)));
break;}
case cast<uint32_t>(19ULL):{
return set(std::string("mtlo",4),go_fmt_Sprintf(std::string("mtlo %s",7),mips_reg(rs)));
break;}
case cast<uint32_t>(24ULL):{
return set(std::string("mult",4),go_fmt_Sprintf(std::string("mult %s, %s",11),mips_reg(rs),mips_reg(rt)));
break;}
case cast<uint32_t>(25ULL):{
return set(std::string("multu",5),go_fmt_Sprintf(std::string("multu %s, %s",12),mips_reg(rs),mips_reg(rt)));
break;}
case cast<uint32_t>(26ULL):{
return set(std::string("div",3),go_fmt_Sprintf(std::string("div %s, %s",10),mips_reg(rs),mips_reg(rt)));
break;}
case cast<uint32_t>(27ULL):{
return set(std::string("divu",4),go_fmt_Sprintf(std::string("divu %s, %s",11),mips_reg(rs),mips_reg(rt)));
break;}
case cast<uint32_t>(32ULL):{
return set(std::string("add",3),go_fmt_Sprintf(std::string("add %s, %s, %s",14),mips_reg(rd),mips_reg(rs),mips_reg(rt)));
break;}
case cast<uint32_t>(33ULL):{
return set(std::string("addu",4),go_fmt_Sprintf(std::string("addu %s, %s, %s",15),mips_reg(rd),mips_reg(rs),mips_reg(rt)));
break;}
case cast<uint32_t>(34ULL):{
return set(std::string("sub",3),go_fmt_Sprintf(std::string("sub %s, %s, %s",14),mips_reg(rd),mips_reg(rs),mips_reg(rt)));
break;}
case cast<uint32_t>(35ULL):{
return set(std::string("subu",4),go_fmt_Sprintf(std::string("subu %s, %s, %s",15),mips_reg(rd),mips_reg(rs),mips_reg(rt)));
break;}
case cast<uint32_t>(36ULL):{
return set(std::string("and",3),go_fmt_Sprintf(std::string("and %s, %s, %s",14),mips_reg(rd),mips_reg(rs),mips_reg(rt)));
break;}
case cast<uint32_t>(37ULL):{
if ((rt == cast<uint32_t>(0ULL))) {
return set(std::string("move",4),go_fmt_Sprintf(std::string("move %s, %s",11),mips_reg(rd),mips_reg(rs)));
}
return set(std::string("or",2),go_fmt_Sprintf(std::string("or %s, %s, %s",13),mips_reg(rd),mips_reg(rs),mips_reg(rt)));
break;}
case cast<uint32_t>(38ULL):{
return set(std::string("xor",3),go_fmt_Sprintf(std::string("xor %s, %s, %s",14),mips_reg(rd),mips_reg(rs),mips_reg(rt)));
break;}
case cast<uint32_t>(39ULL):{
return set(std::string("nor",3),go_fmt_Sprintf(std::string("nor %s, %s, %s",14),mips_reg(rd),mips_reg(rs),mips_reg(rt)));
break;}
case cast<uint32_t>(42ULL):{
return set(std::string("slt",3),go_fmt_Sprintf(std::string("slt %s, %s, %s",14),mips_reg(rd),mips_reg(rs),mips_reg(rt)));
break;}
case cast<uint32_t>(43ULL):{
return set(std::string("sltu",4),go_fmt_Sprintf(std::string("sltu %s, %s, %s",15),mips_reg(rd),mips_reg(rs),mips_reg(rt)));
break;}
}}
return mips_word(in,w);
}
}
// tools/cpu/mips/decode.go:218:1
mips_Inst mips_decodeCop0(mips_Inst in,uint32_t w,uint32_t rs,uint32_t rt,uint32_t rd){
{
auto set = [&](std::string mnem,std::string text)->mips_Inst{
auto tmp17 = std::make_tuple(mnem,text);
in.Mnem = std::get<0>(tmp17);
in.Text = std::get<1>(tmp17);
return in;
}
;
if ((cast<uint32_t>((w & cast<uint32_t>(33554432ULL))) != cast<uint32_t>(0ULL))) {
if ((cast<uint32_t>((w & cast<uint32_t>(63ULL))) == cast<uint32_t>(16ULL))) {
return set(std::string("rfe",3),std::string("rfe",3));
}
return mips_word(in,w);
}
{
switch(rs){
case cast<uint32_t>(0ULL):{
return set(std::string("mfc0",4),go_fmt_Sprintf(std::string("mfc0 %s, $%d",12),mips_reg(rt),rd));
break;}
case cast<uint32_t>(4ULL):{
return set(std::string("mtc0",4),go_fmt_Sprintf(std::string("mtc0 %s, $%d",12),mips_reg(rt),rd));
break;}
}}
return mips_word(in,w);
}
}
// tools/cpu/mips/decode.go:236:1
mips_Inst mips_decodeCop2(mips_Inst in,uint32_t w,uint32_t rs,uint32_t rt,uint32_t rd){
{
auto set = [&](std::string mnem,std::string text)->mips_Inst{
auto tmp18 = std::make_tuple(mnem,text);
in.Mnem = std::get<0>(tmp18);
in.Text = std::get<1>(tmp18);
return in;
}
;
if ((cast<uint32_t>((w & cast<uint32_t>(33554432ULL))) != cast<uint32_t>(0ULL))) {
uint32_t cmd = cast<uint32_t>((w & cast<uint32_t>(33554431ULL)));
std::string name = get(mips_gteCmd,cast<uint32_t>((w & cast<uint32_t>(63ULL))));
if ((name == std::string("",0))) {
return set(std::string("cop2",4),go_fmt_Sprintf(std::string("cop2 0x%07X",11),cmd));
}
return set(name,go_fmt_Sprintf(std::string("%s\011; cop2 0x%07X",16),name,cmd));
}
{
switch(rs){
case cast<uint32_t>(0ULL):{
return set(std::string("mfc2",4),go_fmt_Sprintf(std::string("mfc2 %s, $%d",12),mips_reg(rt),rd));
break;}
case cast<uint32_t>(2ULL):{
return set(std::string("cfc2",4),go_fmt_Sprintf(std::string("cfc2 %s, $%d",12),mips_reg(rt),rd));
break;}
case cast<uint32_t>(4ULL):{
return set(std::string("mtc2",4),go_fmt_Sprintf(std::string("mtc2 %s, $%d",12),mips_reg(rt),rd));
break;}
case cast<uint32_t>(6ULL):{
return set(std::string("ctc2",4),go_fmt_Sprintf(std::string("ctc2 %s, $%d",12),mips_reg(rt),rd));
break;}
}}
return mips_word(in,w);
}
}
// tools/cpu/mips/decode.go:260:1
mips_Inst mips_word(mips_Inst in,uint32_t w){
{
auto tmp19 = std::make_tuple(std::string(".word",5),go_fmt_Sprintf(std::string(".word 0x%08X",12),w),cast<mips_Flow>(7ULL));
in.Mnem = std::get<0>(tmp19);
in.Text = std::get<1>(tmp19);
in.Flow = std::get<2>(tmp19);
auto tmp20 = std::make_tuple(false,false);
in.HasTarget = std::get<0>(tmp20);
in.HasDelay = std::get<1>(tmp20);
return in;
}
}
// tools/cpu/mips/disasm.go:13:1
Slice<std::string> mips_Disassemble(Slice<uint8_t> code,uint32_t base){
{
Slice<std::string> out={};
{int64_t pc = cast<int64_t>(0ULL);for (;(cast<int64_t>((pc + cast<int64_t>(4ULL))) <= len(code));pc += cast<int64_t>(4ULL)){
mips_Inst in = mips_Decode(sub(code,pc,len(code)),cast<uint32_t>((base + cast<uint32_t>(pc))));
std::string raw = go_fmt_Sprintf(std::string("%02X %02X %02X %02X",19),code[pc],code[cast<int64_t>((pc + cast<int64_t>(1ULL)))],code[cast<int64_t>((pc + cast<int64_t>(2ULL)))],code[cast<int64_t>((pc + cast<int64_t>(3ULL)))]);
out = append(out,Slice<std::string>{go_fmt_Sprintf(std::string("$%08X: %s  %s",13),cast<uint32_t>((base + cast<uint32_t>(pc))),raw,go_strings_TrimSpace(in.Text))});
}
}return out;
}
}
// tools/cpu/mips/exec.go:9:1
int64_t mips_CPU_Step(mips_CPU* c){
{
if (c->Halted) {
return cast<int64_t>(0ULL);
}
c->curPC = c->PC;
c->delaySlot = c->pendingDelay;
c->pendingDelay = false;
uint32_t w = mips_CPU_read32(c,c->PC);
c->PC = c->nextPC;
c->nextPC += cast<uint32_t>(4ULL);
if ((c->ld.reg != cast<uint32_t>(0ULL))) {
c->out[c->ld.reg] = c->ld.val;
}
c->ld = mips_loadSlot{};
mips_CPU_execute(c,w);
c->R = c->out;
c->Steps++;
return cast<int64_t>(1ULL);
}
}
// tools/cpu/mips/exec.go:36:1
uint32_t mips_CPU_reg(mips_CPU* c,uint32_t i){
{
return c->R[i];
}
}
// tools/cpu/mips/exec.go:39:1
void mips_CPU_load(mips_CPU* c,uint32_t reg,uint32_t val){
{
c->ld = mips_loadSlot{reg,val};
}
}
// tools/cpu/mips/exec.go:43:1
void mips_CPU_doBranch(mips_CPU* c,bool taken,uint32_t target){
{
c->pendingDelay = true;
c->branchAddr = c->curPC;
if (taken) {
c->nextPC = target;
}
}
}
// tools/cpu/mips/exec.go:51:1
void mips_CPU_execute(mips_CPU* c,uint32_t w){
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
mips_CPU_special(c,w,rs,rt,rd,shamt);
break;}
case cast<uint32_t>(1ULL):{
int32_t s = cast<int32_t>(mips_CPU_reg(c,rs));
bool link = (cast<uint32_t>((rt & cast<uint32_t>(30ULL))) == cast<uint32_t>(16ULL));
if (link) {
mips_CPU_set(c,cast<uint32_t>(31ULL),cast<uint32_t>((c->curPC + cast<uint32_t>(8ULL))));
}
bool taken = ((((cast<uint32_t>((rt & cast<uint32_t>(1ULL))) == cast<uint32_t>(0ULL)) && (s < cast<int32_t>(0ULL)))) || (((cast<uint32_t>((rt & cast<uint32_t>(1ULL))) == cast<uint32_t>(1ULL)) && (s >= cast<int32_t>(0ULL)))));
mips_CPU_doBranch(c,taken,branchT);
break;}
case cast<uint32_t>(2ULL):{
mips_CPU_doBranch(c,true,jumpT);
break;}
case cast<uint32_t>(3ULL):{
mips_CPU_set(c,cast<uint32_t>(31ULL),cast<uint32_t>((c->curPC + cast<uint32_t>(8ULL))));
mips_CPU_doBranch(c,true,jumpT);
break;}
case cast<uint32_t>(4ULL):{
mips_CPU_doBranch(c,(mips_CPU_reg(c,rs) == mips_CPU_reg(c,rt)),branchT);
break;}
case cast<uint32_t>(5ULL):{
mips_CPU_doBranch(c,(mips_CPU_reg(c,rs) != mips_CPU_reg(c,rt)),branchT);
break;}
case cast<uint32_t>(6ULL):{
mips_CPU_doBranch(c,(cast<int32_t>(mips_CPU_reg(c,rs)) <= cast<int32_t>(0ULL)),branchT);
break;}
case cast<uint32_t>(7ULL):{
mips_CPU_doBranch(c,(cast<int32_t>(mips_CPU_reg(c,rs)) > cast<int32_t>(0ULL)),branchT);
break;}
case cast<uint32_t>(8ULL):{
{
auto tmp21 = mips_addOv(mips_CPU_reg(c,rs),simm);
uint32_t r = std::get<0>(tmp21);
bool ov = std::get<1>(tmp21);
if (ov) {
mips_CPU_Exception(c,cast<uint32_t>(12ULL));
}
else {
mips_CPU_set(c,rt,r);
}
}
break;}
case cast<uint32_t>(9ULL):{
mips_CPU_set(c,rt,cast<uint32_t>((mips_CPU_reg(c,rs) + simm)));
break;}
case cast<uint32_t>(10ULL):{
mips_CPU_set(c,rt,mips_b2u((cast<int32_t>(mips_CPU_reg(c,rs)) < cast<int32_t>(simm))));
break;}
case cast<uint32_t>(11ULL):{
mips_CPU_set(c,rt,mips_b2u((mips_CPU_reg(c,rs) < simm)));
break;}
case cast<uint32_t>(12ULL):{
mips_CPU_set(c,rt,cast<uint32_t>((mips_CPU_reg(c,rs) & imm)));
break;}
case cast<uint32_t>(13ULL):{
mips_CPU_set(c,rt,cast<uint32_t>((mips_CPU_reg(c,rs) | imm)));
break;}
case cast<uint32_t>(14ULL):{
mips_CPU_set(c,rt,cast<uint32_t>((mips_CPU_reg(c,rs) ^ imm)));
break;}
case cast<uint32_t>(15ULL):{
mips_CPU_set(c,rt,shl<uint32_t>(imm,cast<int64_t>(16ULL)));
break;}
case cast<uint32_t>(16ULL):{
mips_CPU_cop0(c,w,rs,rt,rd);
break;}
case cast<uint32_t>(18ULL):{
mips_CPU_cop2(c,w,rs,rt,rd);
break;}
case cast<uint32_t>(32ULL):case cast<uint32_t>(33ULL):case cast<uint32_t>(34ULL):case cast<uint32_t>(35ULL):case cast<uint32_t>(36ULL):case cast<uint32_t>(37ULL):case cast<uint32_t>(38ULL):{
mips_CPU_loadOp(c,op,rs,rt,simm);
break;}
case cast<uint32_t>(40ULL):case cast<uint32_t>(41ULL):case cast<uint32_t>(42ULL):case cast<uint32_t>(43ULL):case cast<uint32_t>(46ULL):{
mips_CPU_storeOp(c,op,rs,rt,simm);
break;}
case cast<uint32_t>(50ULL):{
uint32_t addr = cast<uint32_t>((mips_CPU_reg(c,rs) + simm));
if ((cast<uint32_t>((addr & cast<uint32_t>(3ULL))) != cast<uint32_t>(0ULL))) {
mips_CPU_addrError(c,cast<uint32_t>(4ULL),addr);
return ;
}
uint32_t v = mips_CPU_read32(c,addr);
if (bool(c->GTE)) {
rrGTE_Write(c->GTE,rt,v);
}
break;}
case cast<uint32_t>(58ULL):{
uint32_t addr = cast<uint32_t>((mips_CPU_reg(c,rs) + simm));
if ((cast<uint32_t>((addr & cast<uint32_t>(3ULL))) != cast<uint32_t>(0ULL))) {
mips_CPU_addrError(c,cast<uint32_t>(5ULL),addr);
return ;
}
uint32_t v={};
if (bool(c->GTE)) {
v = rrGTE_Read(c->GTE,rt);
}
mips_CPU_write32(c,addr,v);
break;}
default:{
mips_CPU_Halt(c,std::string("unimplemented opcode 0x%02X (word 0x%08X) at 0x%08X",51),op,w,c->curPC);
break;}
}}
}
}
// tools/cpu/mips/exec.go:140:1
void mips_CPU_special(mips_CPU* c,uint32_t w,uint32_t rs,uint32_t rt,uint32_t rd,uint32_t shamt){
{
uint32_t funct = cast<uint32_t>((w & cast<uint32_t>(63ULL)));
{
switch(funct){
case cast<uint32_t>(0ULL):{
mips_CPU_set(c,rd,shl<uint32_t>(mips_CPU_reg(c,rt),shamt));
break;}
case cast<uint32_t>(2ULL):{
mips_CPU_set(c,rd,shr<uint32_t>(mips_CPU_reg(c,rt),shamt));
break;}
case cast<uint32_t>(3ULL):{
mips_CPU_set(c,rd,cast<uint32_t>(shr<int32_t>(cast<int32_t>(mips_CPU_reg(c,rt)),shamt)));
break;}
case cast<uint32_t>(4ULL):{
mips_CPU_set(c,rd,shl<uint32_t>(mips_CPU_reg(c,rt),(cast<uint32_t>((mips_CPU_reg(c,rs) & cast<uint32_t>(31ULL))))));
break;}
case cast<uint32_t>(6ULL):{
mips_CPU_set(c,rd,shr<uint32_t>(mips_CPU_reg(c,rt),(cast<uint32_t>((mips_CPU_reg(c,rs) & cast<uint32_t>(31ULL))))));
break;}
case cast<uint32_t>(7ULL):{
mips_CPU_set(c,rd,cast<uint32_t>(shr<int32_t>(cast<int32_t>(mips_CPU_reg(c,rt)),(cast<uint32_t>((mips_CPU_reg(c,rs) & cast<uint32_t>(31ULL)))))));
break;}
case cast<uint32_t>(8ULL):{
mips_CPU_doBranch(c,true,mips_CPU_reg(c,rs));
break;}
case cast<uint32_t>(9ULL):{
mips_CPU_set(c,rd,cast<uint32_t>((c->curPC + cast<uint32_t>(8ULL))));
mips_CPU_doBranch(c,true,mips_CPU_reg(c,rs));
break;}
case cast<uint32_t>(12ULL):{
if ((bool(c->Syscall) && c->Syscall(c))) {
return ;
}
mips_CPU_Exception(c,cast<uint32_t>(8ULL));
break;}
case cast<uint32_t>(13ULL):{
mips_CPU_Exception(c,cast<uint32_t>(9ULL));
break;}
case cast<uint32_t>(16ULL):{
mips_CPU_set(c,rd,c->HI);
break;}
case cast<uint32_t>(17ULL):{
c->HI = mips_CPU_reg(c,rs);
break;}
case cast<uint32_t>(18ULL):{
mips_CPU_set(c,rd,c->LO);
break;}
case cast<uint32_t>(19ULL):{
c->LO = mips_CPU_reg(c,rs);
break;}
case cast<uint32_t>(24ULL):{
int64_t p = cast<int64_t>((cast<int64_t>(cast<int32_t>(mips_CPU_reg(c,rs))) * cast<int64_t>(cast<int32_t>(mips_CPU_reg(c,rt)))));
auto tmp22 = std::make_tuple(cast<uint32_t>(p),cast<uint32_t>(shr<uint64_t>(cast<uint64_t>(p),cast<int64_t>(32ULL))));
c->LO = std::get<0>(tmp22);
c->HI = std::get<1>(tmp22);
break;}
case cast<uint32_t>(25ULL):{
uint64_t p = cast<uint64_t>((cast<uint64_t>(mips_CPU_reg(c,rs)) * cast<uint64_t>(mips_CPU_reg(c,rt))));
auto tmp23 = std::make_tuple(cast<uint32_t>(p),cast<uint32_t>(shr<uint64_t>(p,cast<int64_t>(32ULL))));
c->LO = std::get<0>(tmp23);
c->HI = std::get<1>(tmp23);
break;}
case cast<uint32_t>(26ULL):{
mips_CPU_divSigned(c,cast<int32_t>(mips_CPU_reg(c,rs)),cast<int32_t>(mips_CPU_reg(c,rt)));
break;}
case cast<uint32_t>(27ULL):{
mips_CPU_divUnsigned(c,mips_CPU_reg(c,rs),mips_CPU_reg(c,rt));
break;}
case cast<uint32_t>(32ULL):{
{
auto tmp24 = mips_addOv(mips_CPU_reg(c,rs),mips_CPU_reg(c,rt));
uint32_t r = std::get<0>(tmp24);
bool ov = std::get<1>(tmp24);
if (ov) {
mips_CPU_Exception(c,cast<uint32_t>(12ULL));
}
else {
mips_CPU_set(c,rd,r);
}
}
break;}
case cast<uint32_t>(33ULL):{
mips_CPU_set(c,rd,cast<uint32_t>((mips_CPU_reg(c,rs) + mips_CPU_reg(c,rt))));
break;}
case cast<uint32_t>(34ULL):{
{
auto tmp25 = mips_subOv(mips_CPU_reg(c,rs),mips_CPU_reg(c,rt));
uint32_t r = std::get<0>(tmp25);
bool ov = std::get<1>(tmp25);
if (ov) {
mips_CPU_Exception(c,cast<uint32_t>(12ULL));
}
else {
mips_CPU_set(c,rd,r);
}
}
break;}
case cast<uint32_t>(35ULL):{
mips_CPU_set(c,rd,cast<uint32_t>((mips_CPU_reg(c,rs) - mips_CPU_reg(c,rt))));
break;}
case cast<uint32_t>(36ULL):{
mips_CPU_set(c,rd,cast<uint32_t>((mips_CPU_reg(c,rs) & mips_CPU_reg(c,rt))));
break;}
case cast<uint32_t>(37ULL):{
mips_CPU_set(c,rd,cast<uint32_t>((mips_CPU_reg(c,rs) | mips_CPU_reg(c,rt))));
break;}
case cast<uint32_t>(38ULL):{
mips_CPU_set(c,rd,cast<uint32_t>((mips_CPU_reg(c,rs) ^ mips_CPU_reg(c,rt))));
break;}
case cast<uint32_t>(39ULL):{
mips_CPU_set(c,rd,cast<uint32_t>(~(cast<uint32_t>((mips_CPU_reg(c,rs) | mips_CPU_reg(c,rt))))));
break;}
case cast<uint32_t>(42ULL):{
mips_CPU_set(c,rd,mips_b2u((cast<int32_t>(mips_CPU_reg(c,rs)) < cast<int32_t>(mips_CPU_reg(c,rt)))));
break;}
case cast<uint32_t>(43ULL):{
mips_CPU_set(c,rd,mips_b2u((mips_CPU_reg(c,rs) < mips_CPU_reg(c,rt))));
break;}
default:{
mips_CPU_Halt(c,std::string("unimplemented special funct 0x%02X (word 0x%08X) at 0x%08X",58),funct,w,c->curPC);
break;}
}}
}
}
// tools/cpu/mips/exec.go:220:1
void mips_CPU_divSigned(mips_CPU* c,int32_t a,int32_t b){
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
auto tmp27 = std::make_tuple(cast<uint32_t>(2147483648ULL),cast<uint32_t>(0ULL));
c->LO = std::get<0>(tmp27);
c->HI = std::get<1>(tmp27);
}
else {
auto tmp28 = std::make_tuple(cast<uint32_t>(divi<int32_t>(a,b)),cast<uint32_t>(modi<int32_t>(a,b)));
c->LO = std::get<0>(tmp28);
c->HI = std::get<1>(tmp28);
}
}
tmp26:;
}
}
// tools/cpu/mips/exec.go:236:1
void mips_CPU_divUnsigned(mips_CPU* c,uint32_t a,uint32_t b){
{
if ((b == cast<uint32_t>(0ULL))) {
auto tmp29 = std::make_tuple(cast<uint32_t>(4294967295ULL),a);
c->LO = std::get<0>(tmp29);
c->HI = std::get<1>(tmp29);
return ;
}
auto tmp30 = std::make_tuple(divi<uint32_t>(a,b),modi<uint32_t>(a,b));
c->LO = std::get<0>(tmp30);
c->HI = std::get<1>(tmp30);
}
}
// tools/cpu/mips/exec.go:244:1
void mips_CPU_loadOp(mips_CPU* c,uint32_t op,uint32_t rs,uint32_t rt,uint32_t simm){
{
uint32_t addr = cast<uint32_t>((mips_CPU_reg(c,rs) + simm));
{
switch(op){
case cast<uint32_t>(32ULL):{
mips_CPU_load(c,rt,cast<uint32_t>(cast<int32_t>(cast<int8_t>(cast<uint8_t>(mips_CPU_read8(c,addr))))));
break;}
case cast<uint32_t>(36ULL):{
mips_CPU_load(c,rt,mips_CPU_read8(c,addr));
break;}
case cast<uint32_t>(33ULL):{
if ((cast<uint32_t>((addr & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL))) {
mips_CPU_addrError(c,cast<uint32_t>(4ULL),addr);
return ;
}
mips_CPU_load(c,rt,cast<uint32_t>(cast<int32_t>(cast<int16_t>(cast<uint16_t>(mips_CPU_read16(c,addr))))));
break;}
case cast<uint32_t>(37ULL):{
if ((cast<uint32_t>((addr & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL))) {
mips_CPU_addrError(c,cast<uint32_t>(4ULL),addr);
return ;
}
mips_CPU_load(c,rt,mips_CPU_read16(c,addr));
break;}
case cast<uint32_t>(35ULL):{
if ((cast<uint32_t>((addr & cast<uint32_t>(3ULL))) != cast<uint32_t>(0ULL))) {
mips_CPU_addrError(c,cast<uint32_t>(4ULL),addr);
return ;
}
mips_CPU_load(c,rt,mips_CPU_read32(c,addr));
break;}
case cast<uint32_t>(34ULL):{
uint32_t word = mips_CPU_read32(c,(addr & ~(cast<uint32_t>(3ULL))));
uint32_t shift = cast<uint32_t>(((cast<uint32_t>((addr & cast<uint32_t>(3ULL)))) * cast<uint32_t>(8ULL)));
uint32_t cur = c->out[rt];
mips_CPU_load(c,rt,cast<uint32_t>(((cast<uint32_t>((cur & (shr<uint32_t>(cast<uint32_t>(16777215ULL),shift))))) | (shl<uint32_t>(word,(cast<uint32_t>((cast<uint32_t>(24ULL) - shift))))))));
break;}
case cast<uint32_t>(38ULL):{
uint32_t word = mips_CPU_read32(c,(addr & ~(cast<uint32_t>(3ULL))));
uint32_t shift = cast<uint32_t>(((cast<uint32_t>((addr & cast<uint32_t>(3ULL)))) * cast<uint32_t>(8ULL)));
uint32_t cur = c->out[rt];
mips_CPU_load(c,rt,cast<uint32_t>(((cast<uint32_t>((cur & (shl<uint32_t>(cast<uint32_t>(4294967040ULL),(cast<uint32_t>((cast<uint32_t>(24ULL) - shift)))))))) | (shr<uint32_t>(word,shift)))));
break;}
}}
}
}
// tools/cpu/mips/exec.go:282:1
void mips_CPU_storeOp(mips_CPU* c,uint32_t op,uint32_t rs,uint32_t rt,uint32_t simm){
{
uint32_t addr = cast<uint32_t>((mips_CPU_reg(c,rs) + simm));
uint32_t v = mips_CPU_reg(c,rt);
{
switch(op){
case cast<uint32_t>(40ULL):{
mips_CPU_write8(c,addr,cast<uint32_t>((v & cast<uint32_t>(255ULL))));
break;}
case cast<uint32_t>(41ULL):{
if ((cast<uint32_t>((addr & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL))) {
mips_CPU_addrError(c,cast<uint32_t>(5ULL),addr);
return ;
}
mips_CPU_write16(c,addr,cast<uint32_t>((v & cast<uint32_t>(65535ULL))));
break;}
case cast<uint32_t>(43ULL):{
if ((cast<uint32_t>((addr & cast<uint32_t>(3ULL))) != cast<uint32_t>(0ULL))) {
mips_CPU_addrError(c,cast<uint32_t>(5ULL),addr);
return ;
}
mips_CPU_write32(c,addr,v);
break;}
case cast<uint32_t>(42ULL):{
uint32_t word = mips_CPU_read32(c,(addr & ~(cast<uint32_t>(3ULL))));
uint32_t shift = cast<uint32_t>(((cast<uint32_t>((addr & cast<uint32_t>(3ULL)))) * cast<uint32_t>(8ULL)));
mips_CPU_write32(c,(addr & ~(cast<uint32_t>(3ULL))),cast<uint32_t>(((cast<uint32_t>((word & (shl<uint32_t>(cast<uint32_t>(4294967040ULL),shift))))) | (shr<uint32_t>(v,(cast<uint32_t>((cast<uint32_t>(24ULL) - shift))))))));
break;}
case cast<uint32_t>(46ULL):{
uint32_t word = mips_CPU_read32(c,(addr & ~(cast<uint32_t>(3ULL))));
uint32_t shift = cast<uint32_t>(((cast<uint32_t>((addr & cast<uint32_t>(3ULL)))) * cast<uint32_t>(8ULL)));
mips_CPU_write32(c,(addr & ~(cast<uint32_t>(3ULL))),cast<uint32_t>(((cast<uint32_t>((word & (shr<uint32_t>(cast<uint32_t>(16777215ULL),(cast<uint32_t>((cast<uint32_t>(24ULL) - shift)))))))) | (shl<uint32_t>(v,shift)))));
break;}
}}
}
}
// tools/cpu/mips/exec.go:311:1
void mips_CPU_cop0(mips_CPU* c,uint32_t w,uint32_t rs,uint32_t rt,uint32_t rd){
{
if ((cast<uint32_t>((w & cast<uint32_t>(33554432ULL))) != cast<uint32_t>(0ULL))) {
if ((cast<uint32_t>((w & cast<uint32_t>(63ULL))) == cast<uint32_t>(16ULL))) {
mips_CPU_rfe(c);
return ;
}
mips_CPU_Halt(c,std::string("unimplemented cop0 command 0x%08X at 0x%08X",43),w,c->curPC);
return ;
}
{
switch(rs){
case cast<uint32_t>(0ULL):{
mips_CPU_load(c,rt,c->COP0[rd]);
break;}
case cast<uint32_t>(4ULL):{
c->COP0[rd] = mips_CPU_reg(c,rt);
break;}
default:{
mips_CPU_Halt(c,std::string("unimplemented cop0 rs=0x%02X (word 0x%08X) at 0x%08X",52),rs,w,c->curPC);
break;}
}}
}
}
// tools/cpu/mips/exec.go:330:1
void mips_CPU_cop2(mips_CPU* c,uint32_t w,uint32_t rs,uint32_t rt,uint32_t rd){
{
if ((cast<uint32_t>((w & cast<uint32_t>(33554432ULL))) != cast<uint32_t>(0ULL))) {
if (bool(c->GTE)) {
rrGTE_Command(c->GTE,cast<uint32_t>((w & cast<uint32_t>(33554431ULL))));
}
return ;
}
{
switch(rs){
case cast<uint32_t>(0ULL):{
uint32_t v={};
if (bool(c->GTE)) {
v = rrGTE_Read(c->GTE,rd);
}
mips_CPU_load(c,rt,v);
break;}
case cast<uint32_t>(2ULL):{
uint32_t v={};
if (bool(c->GTE)) {
v = rrGTE_ReadCtrl(c->GTE,rd);
}
mips_CPU_load(c,rt,v);
break;}
case cast<uint32_t>(4ULL):{
if (bool(c->GTE)) {
rrGTE_Write(c->GTE,rd,mips_CPU_reg(c,rt));
}
break;}
case cast<uint32_t>(6ULL):{
if (bool(c->GTE)) {
rrGTE_WriteCtrl(c->GTE,rd,mips_CPU_reg(c,rt));
}
break;}
default:{
mips_CPU_Halt(c,std::string("unimplemented cop2 rs=0x%02X (word 0x%08X) at 0x%08X",52),rs,w,c->curPC);
break;}
}}
}
}
// tools/cpu/mips/exec.go:364:1
std::tuple<uint32_t,bool> mips_addOv(uint32_t a,uint32_t b){
{
uint32_t r = cast<uint32_t>((a + b));
return {r,(cast<uint32_t>((cast<uint32_t>(((cast<uint32_t>((a ^ r))) & (cast<uint32_t>((b ^ r))))) & cast<uint32_t>(2147483648ULL))) != cast<uint32_t>(0ULL))};
}
}
// tools/cpu/mips/exec.go:368:1
std::tuple<uint32_t,bool> mips_subOv(uint32_t a,uint32_t b){
{
uint32_t r = cast<uint32_t>((a - b));
return {r,(cast<uint32_t>((cast<uint32_t>(((cast<uint32_t>((a ^ b))) & (cast<uint32_t>((a ^ r))))) & cast<uint32_t>(2147483648ULL))) != cast<uint32_t>(0ULL))};
}
}
// tools/cpu/mips/exec.go:373:1
uint32_t mips_b2u(bool b){
{
if (b) {
return cast<uint32_t>(1ULL);
}
return cast<uint32_t>(0ULL);
}
}
// tools/cpu/mips/gte.go:48:1
mips_GTE* mips_NewGTE(){
{
return arenaNew(mips_GTE{});
}
}
// tools/cpu/mips/gte.go:51:1
int32_t mips_s16(uint32_t v){
{
return cast<int32_t>(cast<int16_t>(v));
}
}
// tools/cpu/mips/gte.go:56:1
uint32_t mips_GTE_Read(mips_GTE* g,uint32_t reg){
{
{
switch(cast<uint32_t>((reg & cast<uint32_t>(31ULL)))){
case cast<uint32_t>(15ULL):{
return g->data[cast<int64_t>(14ULL)];
break;}
case cast<uint32_t>(28ULL):case cast<uint32_t>(29ULL):{
int32_t r = mips_clamp5(shr<int32_t>(mips_s16(g->data[cast<int64_t>(9ULL)]),cast<int64_t>(7ULL)));
int32_t gg = mips_clamp5(shr<int32_t>(mips_s16(g->data[cast<int64_t>(10ULL)]),cast<int64_t>(7ULL)));
int32_t b = mips_clamp5(shr<int32_t>(mips_s16(g->data[cast<int64_t>(11ULL)]),cast<int64_t>(7ULL)));
return cast<uint32_t>(cast<int32_t>((cast<int32_t>((r | shl<int32_t>(gg,cast<int64_t>(5ULL)))) | shl<int32_t>(b,cast<int64_t>(10ULL)))));
break;}
default:{
return g->data[cast<uint32_t>((reg & cast<uint32_t>(31ULL)))];
break;}
}}
}
}
// tools/cpu/mips/gte.go:72:1
void mips_GTE_Write(mips_GTE* g,uint32_t reg,uint32_t v){
{
{
switch(cast<uint32_t>((reg & cast<uint32_t>(31ULL)))){
case cast<uint32_t>(15ULL):{
g->data[cast<int64_t>(12ULL)] = g->data[cast<int64_t>(13ULL)];
g->data[cast<int64_t>(13ULL)] = g->data[cast<int64_t>(14ULL)];
g->data[cast<int64_t>(14ULL)] = v;
break;}
case cast<uint32_t>(28ULL):{
g->data[cast<int64_t>(9ULL)] = cast<uint32_t>(shl<int32_t>(cast<int32_t>(cast<uint32_t>((v & cast<uint32_t>(31ULL)))),cast<int64_t>(7ULL)));
g->data[cast<int64_t>(10ULL)] = cast<uint32_t>(shl<int32_t>(cast<int32_t>(cast<uint32_t>(((shr<uint32_t>(v,cast<int64_t>(5ULL))) & cast<uint32_t>(31ULL)))),cast<int64_t>(7ULL)));
g->data[cast<int64_t>(11ULL)] = cast<uint32_t>(shl<int32_t>(cast<int32_t>(cast<uint32_t>(((shr<uint32_t>(v,cast<int64_t>(10ULL))) & cast<uint32_t>(31ULL)))),cast<int64_t>(7ULL)));
g->data[cast<int64_t>(28ULL)] = cast<uint32_t>((v & cast<uint32_t>(32767ULL)));
break;}
case cast<uint32_t>(30ULL):{
g->data[cast<int64_t>(30ULL)] = v;
g->data[cast<int64_t>(31ULL)] = mips_lzc(v);
break;}
case cast<uint32_t>(7ULL):case cast<uint32_t>(23ULL):case cast<uint32_t>(29ULL):case cast<uint32_t>(31ULL):{
g->data[cast<uint32_t>((reg & cast<uint32_t>(31ULL)))] = v;
break;}
default:{
g->data[cast<uint32_t>((reg & cast<uint32_t>(31ULL)))] = v;
break;}
}}
}
}
// tools/cpu/mips/gte.go:94:1
uint32_t mips_GTE_ReadCtrl(mips_GTE* g,uint32_t reg){
{
return g->ctrl[cast<uint32_t>((reg & cast<uint32_t>(31ULL)))];
}
}
// tools/cpu/mips/gte.go:97:1
void mips_GTE_WriteCtrl(mips_GTE* g,uint32_t reg,uint32_t v){
{
g->ctrl[cast<uint32_t>((reg & cast<uint32_t>(31ULL)))] = v;
}
}
// tools/cpu/mips/gte.go:100:1
int32_t mips_clamp5(int32_t v){
{
if ((v < cast<int32_t>(0ULL))) {
return cast<int32_t>(0ULL);
}
if ((v > cast<int32_t>(31ULL))) {
return cast<int32_t>(31ULL);
}
return v;
}
}
// tools/cpu/mips/gte.go:112:1
uint32_t mips_lzc(uint32_t v){
{
uint32_t top = cast<uint32_t>((v & cast<uint32_t>(2147483648ULL)));
uint32_t n = cast<uint32_t>(0ULL);
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(32ULL));i++){
if ((cast<uint32_t>((v & cast<uint32_t>(2147483648ULL))) != top)) {
break;
}
n++;
v = shl<uint32_t>(v,cast<int64_t>(1ULL));
}
}return n;
}
}
// tools/cpu/mips/gte.go:127:1
std::tuple<int32_t,int32_t,int32_t> mips_GTE_vec(mips_GTE* g,int64_t n){
{
{
switch(n){
case cast<int64_t>(0ULL):{
return {mips_s16(g->data[cast<int64_t>(0ULL)]),mips_s16(shr<uint32_t>(g->data[cast<int64_t>(0ULL)],cast<int64_t>(16ULL))),mips_s16(g->data[cast<int64_t>(1ULL)])};
break;}
case cast<int64_t>(1ULL):{
return {mips_s16(g->data[cast<int64_t>(2ULL)]),mips_s16(shr<uint32_t>(g->data[cast<int64_t>(2ULL)],cast<int64_t>(16ULL))),mips_s16(g->data[cast<int64_t>(3ULL)])};
break;}
default:{
return {mips_s16(g->data[cast<int64_t>(4ULL)]),mips_s16(shr<uint32_t>(g->data[cast<int64_t>(4ULL)],cast<int64_t>(16ULL))),mips_s16(g->data[cast<int64_t>(5ULL)])};
break;}
}}
}
}
// tools/cpu/mips/gte.go:139:1
int32_t mips_GTE_rt(mips_GTE* g,int64_t r,int64_t c){
{
return mips_matEntry(sub(g->ctrl,cast<int64_t>(0ULL),cast<int64_t>(5ULL)),r,c);
}
}
// tools/cpu/mips/gte.go:142:1
int32_t mips_GTE_light(mips_GTE* g,int64_t r,int64_t c){
{
return mips_matEntry(sub(g->ctrl,cast<int64_t>(8ULL),cast<int64_t>(13ULL)),r,c);
}
}
// tools/cpu/mips/gte.go:143:1
int32_t mips_GTE_color(mips_GTE* g,int64_t r,int64_t c){
{
return mips_matEntry(sub(g->ctrl,cast<int64_t>(16ULL),cast<int64_t>(21ULL)),r,c);
}
}
// tools/cpu/mips/gte.go:147:1
int32_t mips_matEntry(Slice<uint32_t> w,int64_t r,int64_t c){
{
int64_t idx = cast<int64_t>((cast<int64_t>((r * cast<int64_t>(3ULL))) + c));
{
switch(idx){
case cast<int64_t>(0ULL):{
return mips_s16(w[cast<int64_t>(0ULL)]);
break;}
case cast<int64_t>(1ULL):{
return mips_s16(shr<uint32_t>(w[cast<int64_t>(0ULL)],cast<int64_t>(16ULL)));
break;}
case cast<int64_t>(2ULL):{
return mips_s16(w[cast<int64_t>(1ULL)]);
break;}
case cast<int64_t>(3ULL):{
return mips_s16(shr<uint32_t>(w[cast<int64_t>(1ULL)],cast<int64_t>(16ULL)));
break;}
case cast<int64_t>(4ULL):{
return mips_s16(w[cast<int64_t>(2ULL)]);
break;}
case cast<int64_t>(5ULL):{
return mips_s16(shr<uint32_t>(w[cast<int64_t>(2ULL)],cast<int64_t>(16ULL)));
break;}
case cast<int64_t>(6ULL):{
return mips_s16(w[cast<int64_t>(3ULL)]);
break;}
case cast<int64_t>(7ULL):{
return mips_s16(shr<uint32_t>(w[cast<int64_t>(3ULL)],cast<int64_t>(16ULL)));
break;}
default:{
return mips_s16(w[cast<int64_t>(4ULL)]);
break;}
}}
}
}
// tools/cpu/mips/gte.go:171:1
int32_t mips_GTE_tr(mips_GTE* g,int64_t i){
{
return cast<int32_t>(g->ctrl[cast<int64_t>((cast<int64_t>(5ULL) + i))]);
}
}
// tools/cpu/mips/gte.go:172:1
int32_t mips_GTE_bk(mips_GTE* g,int64_t i){
{
return cast<int32_t>(g->ctrl[cast<int64_t>((cast<int64_t>(13ULL) + i))]);
}
}
// tools/cpu/mips/gte.go:173:1
int32_t mips_GTE_fc(mips_GTE* g,int64_t i){
{
return cast<int32_t>(g->ctrl[cast<int64_t>((cast<int64_t>(21ULL) + i))]);
}
}
// tools/cpu/mips/gte.go:174:1
int32_t mips_GTE_ofx(mips_GTE* g){
{
return cast<int32_t>(g->ctrl[cast<int64_t>(24ULL)]);
}
}
// tools/cpu/mips/gte.go:175:1
int32_t mips_GTE_ofy(mips_GTE* g){
{
return cast<int32_t>(g->ctrl[cast<int64_t>(25ULL)]);
}
}
// tools/cpu/mips/gte.go:176:1
uint32_t mips_GTE_h(mips_GTE* g){
{
return cast<uint32_t>((g->ctrl[cast<int64_t>(26ULL)] & cast<uint32_t>(65535ULL)));
}
}
// tools/cpu/mips/gte.go:177:1
int32_t mips_GTE_dqa(mips_GTE* g){
{
return mips_s16(g->ctrl[cast<int64_t>(27ULL)]);
}
}
// tools/cpu/mips/gte.go:178:1
int32_t mips_GTE_dqb(mips_GTE* g){
{
return cast<int32_t>(g->ctrl[cast<int64_t>(28ULL)]);
}
}
// tools/cpu/mips/gte.go:179:1
int32_t mips_GTE_zsf3(mips_GTE* g){
{
return mips_s16(g->ctrl[cast<int64_t>(29ULL)]);
}
}
// tools/cpu/mips/gte.go:180:1
int32_t mips_GTE_zsf4(mips_GTE* g){
{
return mips_s16(g->ctrl[cast<int64_t>(30ULL)]);
}
}
// tools/cpu/mips/gte.go:181:1
int32_t mips_GTE_irVal(mips_GTE* g,int64_t i){
{
return mips_s16(g->data[cast<int64_t>((cast<int64_t>(8ULL) + i))]);
}
}
// tools/cpu/mips/gte.go:185:1
void mips_GTE_setFlag(mips_GTE* g,uint32_t bits){
{
g->ctrl[cast<int64_t>(31ULL)] |= bits;
}
}
// tools/cpu/mips/gte.go:186:1
void mips_GTE_clearFlags(mips_GTE* g){
{
g->ctrl[cast<int64_t>(31ULL)] = cast<uint32_t>(0ULL);
}
}
// tools/cpu/mips/gte.go:187:1
void mips_GTE_finishFlags(mips_GTE* g){
{
if ((cast<uint32_t>((g->ctrl[cast<int64_t>(31ULL)] & cast<uint32_t>(2139611136ULL))) != cast<uint32_t>(0ULL))) {
g->ctrl[cast<int64_t>(31ULL)] |= cast<uint32_t>(2147483648ULL);
}
}
}
// tools/cpu/mips/gte.go:194:1
void mips_GTE_macCheck(mips_GTE* g,int64_t i,int64_t val){
{
if ((val > cast<int64_t>(8796093022207ULL))) {
mips_GTE_setFlag(g,shl<uint32_t>(cast<uint32_t>(1ULL),(cast<uint64_t>((cast<uint64_t>(31ULL) - cast<uint64_t>(i))))));
}
else if ((val < cast<int64_t>(-8796093022208ULL))) {
mips_GTE_setFlag(g,shl<uint32_t>(cast<uint32_t>(1ULL),(cast<uint64_t>((cast<uint64_t>(28ULL) - cast<uint64_t>(i))))));
}
}
}
// tools/cpu/mips/gte.go:202:1
void mips_GTE_setMac(mips_GTE* g,int64_t i,int32_t v){
{
g->data[cast<int64_t>((cast<int64_t>(24ULL) + i))] = cast<uint32_t>(v);
}
}
// tools/cpu/mips/gte.go:205:1
void mips_GTE_setIR(mips_GTE* g,int64_t i,int32_t v,bool lm){
{
int32_t min = cast<int32_t>(-32768ULL);
if (lm) {
min = cast<int32_t>(0ULL);
}
if ((v > cast<int32_t>(32767ULL))) {
v = cast<int32_t>(32767ULL);
mips_GTE_setFlag(g,shr<uint32_t>(cast<uint32_t>(16777216ULL),cast<uint64_t>(cast<int64_t>((i - cast<int64_t>(1ULL))))));
}
else if ((v < min)) {
v = min;
mips_GTE_setFlag(g,shr<uint32_t>(cast<uint32_t>(16777216ULL),cast<uint64_t>(cast<int64_t>((i - cast<int64_t>(1ULL))))));
}
g->data[cast<int64_t>((cast<int64_t>(8ULL) + i))] = cast<uint32_t>(v);
}
}
// tools/cpu/mips/gte.go:221:1
void mips_GTE_setMacIR(mips_GTE* g,int64_t i,int64_t val,uint64_t sf,bool lm){
{
mips_GTE_macCheck(g,i,val);
int32_t m = cast<int32_t>(shr<int64_t>(val,sf));
mips_GTE_setMac(g,i,m);
mips_GTE_setIR(g,i,m,lm);
}
}
// tools/cpu/mips/gte.go:228:1
void mips_GTE_setIR0(mips_GTE* g,int32_t v){
{
if ((v < cast<int32_t>(0ULL))) {
v = cast<int32_t>(0ULL);
mips_GTE_setFlag(g,cast<uint32_t>(4096ULL));
}
else if ((v > cast<int32_t>(4096ULL))) {
v = cast<int32_t>(4096ULL);
mips_GTE_setFlag(g,cast<uint32_t>(4096ULL));
}
g->data[cast<int64_t>(8ULL)] = cast<uint32_t>(v);
}
}
// tools/cpu/mips/gte.go:239:1
void mips_GTE_setMac0(mips_GTE* g,int64_t val){
{
if ((val > cast<int64_t>(2147483647ULL))) {
mips_GTE_setFlag(g,cast<uint32_t>(65536ULL));
}
else if ((val < cast<int64_t>(-2147483648ULL))) {
mips_GTE_setFlag(g,cast<uint32_t>(32768ULL));
}
g->data[cast<int64_t>(24ULL)] = cast<uint32_t>(cast<int32_t>(val));
}
}
// tools/cpu/mips/gte.go:249:1
void mips_GTE_pushSZ3(mips_GTE* g,int32_t v){
{
if ((v < cast<int32_t>(0ULL))) {
v = cast<int32_t>(0ULL);
mips_GTE_setFlag(g,cast<uint32_t>(262144ULL));
}
else if ((v > cast<int32_t>(65535ULL))) {
v = cast<int32_t>(65535ULL);
mips_GTE_setFlag(g,cast<uint32_t>(262144ULL));
}
g->data[cast<int64_t>(16ULL)] = g->data[cast<int64_t>(17ULL)];
g->data[cast<int64_t>(17ULL)] = g->data[cast<int64_t>(18ULL)];
g->data[cast<int64_t>(18ULL)] = g->data[cast<int64_t>(19ULL)];
g->data[cast<int64_t>(19ULL)] = cast<uint32_t>(v);
}
}
// tools/cpu/mips/gte.go:263:1
void mips_GTE_setOTZ(mips_GTE* g,int32_t v){
{
if ((v < cast<int32_t>(0ULL))) {
v = cast<int32_t>(0ULL);
mips_GTE_setFlag(g,cast<uint32_t>(262144ULL));
}
else if ((v > cast<int32_t>(65535ULL))) {
v = cast<int32_t>(65535ULL);
mips_GTE_setFlag(g,cast<uint32_t>(262144ULL));
}
g->data[cast<int64_t>(7ULL)] = cast<uint32_t>(v);
}
}
// tools/cpu/mips/gte.go:275:1
void mips_GTE_pushSXY(mips_GTE* g,int32_t x,int32_t y){
{
int32_t sx = mips_clampSXY(x,cast<uint32_t>(16384ULL),g);
int32_t sy = mips_clampSXY(y,cast<uint32_t>(8192ULL),g);
g->data[cast<int64_t>(12ULL)] = g->data[cast<int64_t>(13ULL)];
g->data[cast<int64_t>(13ULL)] = g->data[cast<int64_t>(14ULL)];
g->data[cast<int64_t>(14ULL)] = cast<uint32_t>((cast<uint32_t>(cast<uint16_t>(cast<int16_t>(sx))) | shl<uint32_t>(cast<uint32_t>(cast<uint16_t>(cast<int16_t>(sy))),cast<int64_t>(16ULL))));
}
}
// tools/cpu/mips/gte.go:283:1
int32_t mips_clampSXY(int32_t v,uint32_t bit,mips_GTE* g){
{
if ((v < cast<int32_t>(-1024ULL))) {
mips_GTE_setFlag(g,bit);
return cast<int32_t>(-1024ULL);
}
if ((v > cast<int32_t>(1023ULL))) {
mips_GTE_setFlag(g,bit);
return cast<int32_t>(1023ULL);
}
return v;
}
}
// tools/cpu/mips/gte.go:298:1
void mips_GTE_Command(mips_GTE* g,uint32_t cmd){
{
mips_GTE_clearFlags(g);
uint64_t sf = cast<uint64_t>(0ULL);
if ((cast<uint32_t>((cmd & cast<uint32_t>(524288ULL))) != cast<uint32_t>(0ULL))) {
sf = cast<uint64_t>(12ULL);
}
bool lm = (cast<uint32_t>((cmd & cast<uint32_t>(1024ULL))) != cast<uint32_t>(0ULL));
{
switch(cast<uint32_t>((cmd & cast<uint32_t>(63ULL)))){
case cast<uint32_t>(1ULL):{
mips_GTE_rtp(g,cast<int64_t>(0ULL),sf,lm,true);
break;}
case cast<uint32_t>(48ULL):{
mips_GTE_rtp(g,cast<int64_t>(0ULL),sf,lm,false);
mips_GTE_rtp(g,cast<int64_t>(1ULL),sf,lm,false);
mips_GTE_rtp(g,cast<int64_t>(2ULL),sf,lm,true);
break;}
case cast<uint32_t>(6ULL):{
mips_GTE_nclip(g);
break;}
case cast<uint32_t>(18ULL):{
mips_GTE_mvmva(g,cmd,sf,lm);
break;}
case cast<uint32_t>(45ULL):{
mips_GTE_avsz(g,false);
break;}
case cast<uint32_t>(46ULL):{
mips_GTE_avsz(g,true);
break;}
case cast<uint32_t>(30ULL):{
mips_GTE_normalColor(g,cast<int64_t>(0ULL),sf,lm,false,false);
break;}
case cast<uint32_t>(32ULL):{
mips_GTE_normalColorT(g,sf,lm,false,false);
break;}
case cast<uint32_t>(27ULL):{
mips_GTE_normalColor(g,cast<int64_t>(0ULL),sf,lm,true,false);
break;}
case cast<uint32_t>(63ULL):{
mips_GTE_normalColorT(g,sf,lm,true,false);
break;}
case cast<uint32_t>(19ULL):{
mips_GTE_normalColor(g,cast<int64_t>(0ULL),sf,lm,true,true);
break;}
case cast<uint32_t>(22ULL):{
mips_GTE_normalColorT(g,sf,lm,true,true);
break;}
case cast<uint32_t>(28ULL):{
mips_GTE_lightColor(g,sf,lm);
mips_GTE_primaryColor(g,sf,lm,false);
mips_GTE_pushColorFIFO(g);
break;}
case cast<uint32_t>(20ULL):{
mips_GTE_lightColor(g,sf,lm);
mips_GTE_primaryColor(g,sf,lm,true);
mips_GTE_pushColorFIFO(g);
break;}
case cast<uint32_t>(17ULL):{
mips_GTE_interpolate(g,sf,lm,std::array<int64_t,3>{shl<int64_t>(cast<int64_t>(mips_GTE_irVal(g,cast<int64_t>(1ULL))),cast<int64_t>(12ULL)),shl<int64_t>(cast<int64_t>(mips_GTE_irVal(g,cast<int64_t>(2ULL))),cast<int64_t>(12ULL)),shl<int64_t>(cast<int64_t>(mips_GTE_irVal(g,cast<int64_t>(3ULL))),cast<int64_t>(12ULL))});
break;}
case cast<uint32_t>(16ULL):{
uint32_t c = g->data[cast<int64_t>(6ULL)];
mips_GTE_interpolate(g,sf,lm,std::array<int64_t,3>{shl<int64_t>(cast<int64_t>(cast<uint32_t>((c & cast<uint32_t>(255ULL)))),cast<int64_t>(16ULL)),shl<int64_t>(cast<int64_t>(cast<uint32_t>(((shr<uint32_t>(c,cast<int64_t>(8ULL))) & cast<uint32_t>(255ULL)))),cast<int64_t>(16ULL)),shl<int64_t>(cast<int64_t>(cast<uint32_t>(((shr<uint32_t>(c,cast<int64_t>(16ULL))) & cast<uint32_t>(255ULL)))),cast<int64_t>(16ULL))});
break;}
default:{
break;}
}}
mips_GTE_finishFlags(g);
}
}
// tools/cpu/mips/gte.go:355:1
void mips_GTE_rtp(mips_GTE* g,int64_t n,uint64_t sf,bool lm,bool depth){
{
auto tmp31 = mips_GTE_vec(g,n);
int32_t vx = std::get<0>(tmp31);
int32_t vy = std::get<1>(tmp31);
int32_t vz = std::get<2>(tmp31);
mips_GTE_setMacIR(g,cast<int64_t>(1ULL),cast<int64_t>((cast<int64_t>((cast<int64_t>((shl<int64_t>(cast<int64_t>(mips_GTE_tr(g,cast<int64_t>(0ULL))),cast<int64_t>(12ULL)) + cast<int64_t>((cast<int64_t>(mips_GTE_rt(g,cast<int64_t>(0ULL),cast<int64_t>(0ULL))) * cast<int64_t>(vx))))) + cast<int64_t>((cast<int64_t>(mips_GTE_rt(g,cast<int64_t>(0ULL),cast<int64_t>(1ULL))) * cast<int64_t>(vy))))) + cast<int64_t>((cast<int64_t>(mips_GTE_rt(g,cast<int64_t>(0ULL),cast<int64_t>(2ULL))) * cast<int64_t>(vz))))),sf,lm);
mips_GTE_setMacIR(g,cast<int64_t>(2ULL),cast<int64_t>((cast<int64_t>((cast<int64_t>((shl<int64_t>(cast<int64_t>(mips_GTE_tr(g,cast<int64_t>(1ULL))),cast<int64_t>(12ULL)) + cast<int64_t>((cast<int64_t>(mips_GTE_rt(g,cast<int64_t>(1ULL),cast<int64_t>(0ULL))) * cast<int64_t>(vx))))) + cast<int64_t>((cast<int64_t>(mips_GTE_rt(g,cast<int64_t>(1ULL),cast<int64_t>(1ULL))) * cast<int64_t>(vy))))) + cast<int64_t>((cast<int64_t>(mips_GTE_rt(g,cast<int64_t>(1ULL),cast<int64_t>(2ULL))) * cast<int64_t>(vz))))),sf,lm);
int64_t mac3full = cast<int64_t>((cast<int64_t>((cast<int64_t>((shl<int64_t>(cast<int64_t>(mips_GTE_tr(g,cast<int64_t>(2ULL))),cast<int64_t>(12ULL)) + cast<int64_t>((cast<int64_t>(mips_GTE_rt(g,cast<int64_t>(2ULL),cast<int64_t>(0ULL))) * cast<int64_t>(vx))))) + cast<int64_t>((cast<int64_t>(mips_GTE_rt(g,cast<int64_t>(2ULL),cast<int64_t>(1ULL))) * cast<int64_t>(vy))))) + cast<int64_t>((cast<int64_t>(mips_GTE_rt(g,cast<int64_t>(2ULL),cast<int64_t>(2ULL))) * cast<int64_t>(vz)))));
mips_GTE_macCheck(g,cast<int64_t>(3ULL),mac3full);
mips_GTE_setMac(g,cast<int64_t>(3ULL),cast<int32_t>(shr<int64_t>(mac3full,sf)));
mips_GTE_setIR(g,cast<int64_t>(3ULL),cast<int32_t>(shr<int64_t>(mac3full,sf)),lm);
mips_GTE_pushSZ3(g,cast<int32_t>(shr<int64_t>(mac3full,cast<int64_t>(12ULL))));
uint32_t div = mips_GTE_divide(g,mips_GTE_h(g),cast<uint32_t>((g->data[cast<int64_t>(19ULL)] & cast<uint32_t>(65535ULL))));
int64_t macx = cast<int64_t>((cast<int64_t>((cast<int64_t>(div) * cast<int64_t>(mips_GTE_irVal(g,cast<int64_t>(1ULL))))) + cast<int64_t>(mips_GTE_ofx(g))));
mips_GTE_setMac0(g,macx);
int64_t macy = cast<int64_t>((cast<int64_t>((cast<int64_t>(div) * cast<int64_t>(mips_GTE_irVal(g,cast<int64_t>(2ULL))))) + cast<int64_t>(mips_GTE_ofy(g))));
int32_t sx = cast<int32_t>(shr<int64_t>(macx,cast<int64_t>(16ULL)));
mips_GTE_setMac0(g,macy);
int32_t sy = cast<int32_t>(shr<int64_t>(macy,cast<int64_t>(16ULL)));
mips_GTE_pushSXY(g,sx,sy);
if (depth) {
int64_t macz = cast<int64_t>((cast<int64_t>((cast<int64_t>(div) * cast<int64_t>(mips_GTE_dqa(g)))) + cast<int64_t>(mips_GTE_dqb(g))));
mips_GTE_setMac0(g,macz);
mips_GTE_setIR0(g,cast<int32_t>(shr<int64_t>(macz,cast<int64_t>(12ULL))));
}
}
}
// tools/cpu/mips/gte.go:385:1
void mips_GTE_nclip(mips_GTE* g){
{
auto tmp32 = std::make_tuple(mips_s16(g->data[cast<int64_t>(12ULL)]),mips_s16(shr<uint32_t>(g->data[cast<int64_t>(12ULL)],cast<int64_t>(16ULL))));
int32_t sx0 = std::get<0>(tmp32);
int32_t sy0 = std::get<1>(tmp32);
auto tmp33 = std::make_tuple(mips_s16(g->data[cast<int64_t>(13ULL)]),mips_s16(shr<uint32_t>(g->data[cast<int64_t>(13ULL)],cast<int64_t>(16ULL))));
int32_t sx1 = std::get<0>(tmp33);
int32_t sy1 = std::get<1>(tmp33);
auto tmp34 = std::make_tuple(mips_s16(g->data[cast<int64_t>(14ULL)]),mips_s16(shr<uint32_t>(g->data[cast<int64_t>(14ULL)],cast<int64_t>(16ULL))));
int32_t sx2 = std::get<0>(tmp34);
int32_t sy2 = std::get<1>(tmp34);
mips_GTE_setMac0(g,cast<int64_t>((cast<int64_t>((cast<int64_t>((cast<int64_t>((cast<int64_t>((cast<int64_t>((cast<int64_t>(sx0) * cast<int64_t>(sy1))) + cast<int64_t>((cast<int64_t>(sx1) * cast<int64_t>(sy2))))) + cast<int64_t>((cast<int64_t>(sx2) * cast<int64_t>(sy0))))) - cast<int64_t>((cast<int64_t>(sx0) * cast<int64_t>(sy2))))) - cast<int64_t>((cast<int64_t>(sx1) * cast<int64_t>(sy0))))) - cast<int64_t>((cast<int64_t>(sx2) * cast<int64_t>(sy1))))));
}
}
// tools/cpu/mips/gte.go:394:1
void mips_GTE_avsz(mips_GTE* g,bool four){
{
int64_t sum={};
int32_t zsf={};
if (four) {
zsf = mips_GTE_zsf4(g);
sum = cast<int64_t>((cast<int64_t>((cast<int64_t>((cast<int64_t>(cast<uint32_t>((g->data[cast<int64_t>(16ULL)] & cast<uint32_t>(65535ULL)))) + cast<int64_t>(cast<uint32_t>((g->data[cast<int64_t>(17ULL)] & cast<uint32_t>(65535ULL)))))) + cast<int64_t>(cast<uint32_t>((g->data[cast<int64_t>(18ULL)] & cast<uint32_t>(65535ULL)))))) + cast<int64_t>(cast<uint32_t>((g->data[cast<int64_t>(19ULL)] & cast<uint32_t>(65535ULL))))));
}
else {
zsf = mips_GTE_zsf3(g);
sum = cast<int64_t>((cast<int64_t>((cast<int64_t>(cast<uint32_t>((g->data[cast<int64_t>(17ULL)] & cast<uint32_t>(65535ULL)))) + cast<int64_t>(cast<uint32_t>((g->data[cast<int64_t>(18ULL)] & cast<uint32_t>(65535ULL)))))) + cast<int64_t>(cast<uint32_t>((g->data[cast<int64_t>(19ULL)] & cast<uint32_t>(65535ULL))))));
}
int64_t mac0 = cast<int64_t>((cast<int64_t>(zsf) * sum));
mips_GTE_setMac0(g,mac0);
mips_GTE_setOTZ(g,cast<int32_t>(shr<int64_t>(mac0,cast<int64_t>(12ULL))));
}
}
// tools/cpu/mips/gte.go:411:1
void mips_GTE_mvmva(mips_GTE* g,uint32_t cmd,uint64_t sf,bool lm){
{
uint32_t mx = cast<uint32_t>(((shr<uint32_t>(cmd,cast<int64_t>(17ULL))) & cast<uint32_t>(3ULL)));
uint32_t vsel = cast<uint32_t>(((shr<uint32_t>(cmd,cast<int64_t>(15ULL))) & cast<uint32_t>(3ULL)));
uint32_t cv = cast<uint32_t>(((shr<uint32_t>(cmd,cast<int64_t>(13ULL))) & cast<uint32_t>(3ULL)));
auto mat = [&](int64_t r,int64_t c)->int32_t{
{
switch(mx){
case cast<uint32_t>(0ULL):{
return mips_GTE_rt(g,r,c);
break;}
case cast<uint32_t>(1ULL):{
return mips_GTE_light(g,r,c);
break;}
default:{
return mips_GTE_color(g,r,c);
break;}
}}
}
;
int32_t vx={};
int32_t vy={};
int32_t vz={};
{
switch(vsel){
case cast<uint32_t>(0ULL):case cast<uint32_t>(1ULL):case cast<uint32_t>(2ULL):{
auto tmp35 = mips_GTE_vec(g,cast<int64_t>(vsel));
vx = std::get<0>(tmp35);
vy = std::get<1>(tmp35);
vz = std::get<2>(tmp35);
break;}
default:{
auto tmp36 = std::make_tuple(mips_GTE_irVal(g,cast<int64_t>(1ULL)),mips_GTE_irVal(g,cast<int64_t>(2ULL)),mips_GTE_irVal(g,cast<int64_t>(3ULL)));
vx = std::get<0>(tmp36);
vy = std::get<1>(tmp36);
vz = std::get<2>(tmp36);
break;}
}}
auto tr = [&](int64_t i)->int64_t{
{
switch(cv){
case cast<uint32_t>(0ULL):{
return cast<int64_t>(mips_GTE_tr(g,i));
break;}
case cast<uint32_t>(1ULL):{
return cast<int64_t>(mips_GTE_bk(g,i));
break;}
case cast<uint32_t>(2ULL):{
return cast<int64_t>(mips_GTE_fc(g,i));
break;}
default:{
return cast<int64_t>(0ULL);
break;}
}}
}
;
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(3ULL));i++){
int64_t val = cast<int64_t>((cast<int64_t>((cast<int64_t>((shl<int64_t>(tr(i),cast<int64_t>(12ULL)) + cast<int64_t>((cast<int64_t>(mat(i,cast<int64_t>(0ULL))) * cast<int64_t>(vx))))) + cast<int64_t>((cast<int64_t>(mat(i,cast<int64_t>(1ULL))) * cast<int64_t>(vy))))) + cast<int64_t>((cast<int64_t>(mat(i,cast<int64_t>(2ULL))) * cast<int64_t>(vz)))));
mips_GTE_setMacIR(g,cast<int64_t>((i + cast<int64_t>(1ULL))),val,sf,lm);
}
}}
}
// tools/cpu/mips/gte.go:458:1
void mips_GTE_normalColor(mips_GTE* g,int64_t n,uint64_t sf,bool lm,bool primary,bool depth){
{
mips_GTE_lightVector(g,n,sf,lm);
mips_GTE_lightColor(g,sf,lm);
if (primary) {
mips_GTE_primaryColor(g,sf,lm,depth);
}
mips_GTE_pushColorFIFO(g);
}
}
// tools/cpu/mips/gte.go:467:1
void mips_GTE_normalColorT(mips_GTE* g,uint64_t sf,bool lm,bool primary,bool depth){
{
{int64_t n = cast<int64_t>(0ULL);for (;(n < cast<int64_t>(3ULL));n++){
mips_GTE_normalColor(g,n,sf,lm,primary,depth);
}
}}
}
// tools/cpu/mips/gte.go:475:1
void mips_GTE_lightVector(mips_GTE* g,int64_t n,uint64_t sf,bool lm){
{
auto tmp37 = mips_GTE_vec(g,n);
int32_t vx = std::get<0>(tmp37);
int32_t vy = std::get<1>(tmp37);
int32_t vz = std::get<2>(tmp37);
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(3ULL));i++){
int64_t val = cast<int64_t>((cast<int64_t>((cast<int64_t>((cast<int64_t>(mips_GTE_light(g,i,cast<int64_t>(0ULL))) * cast<int64_t>(vx))) + cast<int64_t>((cast<int64_t>(mips_GTE_light(g,i,cast<int64_t>(1ULL))) * cast<int64_t>(vy))))) + cast<int64_t>((cast<int64_t>(mips_GTE_light(g,i,cast<int64_t>(2ULL))) * cast<int64_t>(vz)))));
mips_GTE_setMacIR(g,cast<int64_t>((i + cast<int64_t>(1ULL))),val,sf,lm);
}
}}
}
// tools/cpu/mips/gte.go:485:1
void mips_GTE_lightColor(mips_GTE* g,uint64_t sf,bool lm){
{
auto tmp38 = std::make_tuple(mips_GTE_irVal(g,cast<int64_t>(1ULL)),mips_GTE_irVal(g,cast<int64_t>(2ULL)),mips_GTE_irVal(g,cast<int64_t>(3ULL)));
int32_t i1 = std::get<0>(tmp38);
int32_t i2 = std::get<1>(tmp38);
int32_t i3 = std::get<2>(tmp38);
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(3ULL));i++){
int64_t val = cast<int64_t>((cast<int64_t>((cast<int64_t>((shl<int64_t>(cast<int64_t>(mips_GTE_bk(g,i)),cast<int64_t>(12ULL)) + cast<int64_t>((cast<int64_t>(mips_GTE_color(g,i,cast<int64_t>(0ULL))) * cast<int64_t>(i1))))) + cast<int64_t>((cast<int64_t>(mips_GTE_color(g,i,cast<int64_t>(1ULL))) * cast<int64_t>(i2))))) + cast<int64_t>((cast<int64_t>(mips_GTE_color(g,i,cast<int64_t>(2ULL))) * cast<int64_t>(i3)))));
mips_GTE_setMacIR(g,cast<int64_t>((i + cast<int64_t>(1ULL))),val,sf,lm);
}
}}
}
// tools/cpu/mips/gte.go:495:1
void mips_GTE_primaryColor(mips_GTE* g,uint64_t sf,bool lm,bool depth){
{
std::array<int64_t,3> col = std::array<int64_t,3>{cast<int64_t>(cast<uint32_t>((g->data[cast<int64_t>(6ULL)] & cast<uint32_t>(255ULL)))),cast<int64_t>(cast<uint32_t>(((shr<uint32_t>(g->data[cast<int64_t>(6ULL)],cast<int64_t>(8ULL))) & cast<uint32_t>(255ULL)))),cast<int64_t>(cast<uint32_t>(((shr<uint32_t>(g->data[cast<int64_t>(6ULL)],cast<int64_t>(16ULL))) & cast<uint32_t>(255ULL))))};
std::array<int64_t,3> mac={};
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(3ULL));i++){
mac[i] = shl<int64_t>(cast<int64_t>((col[i] * cast<int64_t>(mips_GTE_irVal(g,cast<int64_t>((i + cast<int64_t>(1ULL))))))),cast<int64_t>(4ULL));
}
}if (depth) {
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(3ULL));i++){
int64_t diff = cast<int64_t>(((shl<int64_t>(cast<int64_t>(mips_GTE_fc(g,i)),cast<int64_t>(12ULL))) - mac[i]));
mips_GTE_setIR(g,cast<int64_t>((i + cast<int64_t>(1ULL))),cast<int32_t>(shr<int64_t>(diff,sf)),false);
mac[i] += cast<int64_t>((cast<int64_t>(mips_GTE_irVal(g,cast<int64_t>((i + cast<int64_t>(1ULL))))) * cast<int64_t>(mips_GTE_irVal(g,cast<int64_t>(0ULL)))));
}
}}
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(3ULL));i++){
mips_GTE_macCheck(g,cast<int64_t>((i + cast<int64_t>(1ULL))),mac[i]);
int32_t m = cast<int32_t>(shr<int64_t>(mac[i],sf));
mips_GTE_setMac(g,cast<int64_t>((i + cast<int64_t>(1ULL))),m);
mips_GTE_setIR(g,cast<int64_t>((i + cast<int64_t>(1ULL))),m,lm);
}
}}
}
// tools/cpu/mips/gte.go:520:1
void mips_GTE_interpolate(mips_GTE* g,uint64_t sf,bool lm,std::array<int64_t,3> start){
{
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(3ULL));i++){
int64_t mac = start[i];
int64_t diff = cast<int64_t>(((shl<int64_t>(cast<int64_t>(mips_GTE_fc(g,i)),cast<int64_t>(12ULL))) - mac));
mips_GTE_setIR(g,cast<int64_t>((i + cast<int64_t>(1ULL))),cast<int32_t>(shr<int64_t>(diff,sf)),false);
mac += cast<int64_t>((cast<int64_t>(mips_GTE_irVal(g,cast<int64_t>((i + cast<int64_t>(1ULL))))) * cast<int64_t>(mips_GTE_irVal(g,cast<int64_t>(0ULL)))));
mips_GTE_macCheck(g,cast<int64_t>((i + cast<int64_t>(1ULL))),mac);
int32_t m = cast<int32_t>(shr<int64_t>(mac,sf));
mips_GTE_setMac(g,cast<int64_t>((i + cast<int64_t>(1ULL))),m);
mips_GTE_setIR(g,cast<int64_t>((i + cast<int64_t>(1ULL))),m,lm);
}
}mips_GTE_pushColorFIFO(g);
}
}
// tools/cpu/mips/gte.go:536:1
void mips_GTE_pushColorFIFO(mips_GTE* g){
{
uint32_t code = cast<uint32_t>(((shr<uint32_t>(g->data[cast<int64_t>(6ULL)],cast<int64_t>(24ULL))) & cast<uint32_t>(255ULL)));
int32_t r = mips_GTE_clampColor(g,shr<int32_t>(cast<int32_t>(g->data[cast<int64_t>(25ULL)]),cast<int64_t>(4ULL)),cast<uint64_t>(21ULL));
int32_t gg = mips_GTE_clampColor(g,shr<int32_t>(cast<int32_t>(g->data[cast<int64_t>(26ULL)]),cast<int64_t>(4ULL)),cast<uint64_t>(20ULL));
int32_t b = mips_GTE_clampColor(g,shr<int32_t>(cast<int32_t>(g->data[cast<int64_t>(27ULL)]),cast<int64_t>(4ULL)),cast<uint64_t>(19ULL));
g->data[cast<int64_t>(20ULL)] = g->data[cast<int64_t>(21ULL)];
g->data[cast<int64_t>(21ULL)] = g->data[cast<int64_t>(22ULL)];
g->data[cast<int64_t>(22ULL)] = cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(r) | shl<uint32_t>(cast<uint32_t>(gg),cast<int64_t>(8ULL)))) | shl<uint32_t>(cast<uint32_t>(b),cast<int64_t>(16ULL)))) | shl<uint32_t>(code,cast<int64_t>(24ULL))));
}
}
// tools/cpu/mips/gte.go:547:1
int32_t mips_GTE_clampColor(mips_GTE* g,int32_t v,uint64_t bit){
{
if ((v < cast<int32_t>(0ULL))) {
mips_GTE_setFlag(g,shl<uint32_t>(cast<uint32_t>(1ULL),bit));
return cast<int32_t>(0ULL);
}
if ((v > cast<int32_t>(255ULL))) {
mips_GTE_setFlag(g,shl<uint32_t>(cast<uint32_t>(1ULL),bit));
return cast<int32_t>(255ULL);
}
return v;
}
}
// tools/cpu/mips/gte.go:564:1
uint32_t mips_GTE_divide(mips_GTE* g,uint32_t h,uint32_t sz3){
{
if ((h >= cast<uint32_t>((sz3 * cast<uint32_t>(2ULL))))) {
mips_GTE_setFlag(g,cast<uint32_t>(131072ULL));
return cast<uint32_t>(131071ULL);
}
uint32_t z = mips_lzc16(cast<uint16_t>(sz3));
uint32_t n = shl<uint32_t>(h,z);
uint32_t d = shl<uint32_t>(sz3,z);
uint32_t u = cast<uint32_t>((cast<uint32_t>(mips_unrTable[shr<uint32_t>((cast<uint32_t>((d - cast<uint32_t>(32704ULL)))),cast<int64_t>(7ULL))]) + cast<uint32_t>(257ULL)));
d = shr<uint32_t>((cast<uint32_t>((cast<uint32_t>(33554560ULL) - cast<uint32_t>((d * u))))),cast<int64_t>(8ULL));
d = shr<uint32_t>((cast<uint32_t>((cast<uint32_t>(128ULL) + cast<uint32_t>((d * u))))),cast<int64_t>(8ULL));
uint64_t res = shr<uint64_t>((cast<uint64_t>((cast<uint64_t>((cast<uint64_t>(n) * cast<uint64_t>(d))) + cast<uint64_t>(32768ULL)))),cast<int64_t>(16ULL));
if ((res > cast<uint64_t>(131071ULL))) {
return cast<uint32_t>(131071ULL);
}
return cast<uint32_t>(res);
}
}
// tools/cpu/mips/gte.go:583:1
uint32_t mips_lzc16(uint16_t v){
{
uint32_t n = cast<uint32_t>(0ULL);
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(16ULL));i++){
if ((cast<uint16_t>((v & cast<uint16_t>(32768ULL))) != cast<uint16_t>(0ULL))) {
break;
}
n++;
v = shl<uint16_t>(v,cast<int64_t>(1ULL));
}
}return n;
}
}
// tools/cpu/mips/mips.go:57:1
std::string mips_Flow_String(mips_Flow f){
{
{
switch(f){
case cast<mips_Flow>(0ULL):{
return std::string("seq",3);
break;}
case cast<mips_Flow>(1ULL):{
return std::string("branch",6);
break;}
case cast<mips_Flow>(2ULL):{
return std::string("jump",4);
break;}
case cast<mips_Flow>(3ULL):{
return std::string("call",4);
break;}
case cast<mips_Flow>(4ULL):{
return std::string("return",6);
break;}
case cast<mips_Flow>(5ULL):{
return std::string("indjump",7);
break;}
case cast<mips_Flow>(6ULL):{
return std::string("indcall",7);
break;}
case cast<mips_Flow>(7ULL):{
return std::string("stop",4);
break;}
}}
return std::string("?",1);
}
}
// tools/cpu/mips/mips.go:92:1
std::string mips_Inst_String(mips_Inst in){
{
return go_fmt_Sprintf(std::string("$%08X: %s",9),in.Addr,in.Text);
}
}
// tools/cpu/mips/state.go:26:1
mips_CPUState mips_CPU_SaveState(mips_CPU* c){
{
return mips_CPUState{c->R,c->out,c->HI,c->LO,c->PC,c->nextPC,c->COP0,c->Halted,c->HaltReason,c->Steps,c->curPC,c->ld.reg,c->ld.val,c->delaySlot,c->pendingDelay,c->branchAddr};
}
}
// tools/cpu/mips/state.go:37:1
void mips_CPU_LoadState(mips_CPU* c,mips_CPUState s){
{
auto tmp39 = std::make_tuple(s.R,s.Out,s.HI,s.LO);
c->R = std::get<0>(tmp39);
c->out = std::get<1>(tmp39);
c->HI = std::get<2>(tmp39);
c->LO = std::get<3>(tmp39);
auto tmp40 = std::make_tuple(s.PC,s.NextPC,s.COP0);
c->PC = std::get<0>(tmp40);
c->nextPC = std::get<1>(tmp40);
c->COP0 = std::get<2>(tmp40);
auto tmp41 = std::make_tuple(s.Halted,s.HaltReason,s.Steps);
c->Halted = std::get<0>(tmp41);
c->HaltReason = std::get<1>(tmp41);
c->Steps = std::get<2>(tmp41);
auto tmp42 = std::make_tuple(s.CurPC,mips_loadSlot{s.LdReg,s.LdVal});
c->curPC = std::get<0>(tmp42);
c->ld = std::get<1>(tmp42);
auto tmp43 = std::make_tuple(s.DelaySlot,s.PendingDelay,s.BranchAddr);
c->delaySlot = std::get<0>(tmp43);
c->pendingDelay = std::get<1>(tmp43);
c->branchAddr = std::get<2>(tmp43);
}
}
// tools/cpu/mips/state.go:51:1
mips_GTEState mips_GTE_SaveState(mips_GTE* g){
{
return mips_GTEState{g->data,g->ctrl};
}
}
// tools/cpu/mips/state.go:54:1
void mips_GTE_LoadState(mips_GTE* g,mips_GTEState s){
{
auto tmp44 = std::make_tuple(s.Data,s.Ctrl);
g->data = std::get<0>(tmp44);
g->ctrl = std::get<1>(tmp44);
}
}
// tools/cpu/vu/disasm.go:32:1
vu_Inst vu_Decode(uint64_t raw,uint32_t addr){
{
uint32_t up = cast<uint32_t>(shr<uint64_t>(raw,cast<int64_t>(32ULL)));
uint32_t lo = cast<uint32_t>(raw);
vu_Inst inst = vu_Inst{{},{},(cast<uint32_t>((up & cast<uint32_t>(2147483648ULL))) != cast<uint32_t>(0ULL)),(cast<uint32_t>((up & cast<uint32_t>(1073741824ULL))) != cast<uint32_t>(0ULL)),raw};
inst.Upper = vu_decodeUpper(up);
if (inst.I) {
inst.Lower = go_fmt_Sprintf(std::string("loi %g",6),vu_float32frombits(lo));
}
else {
inst.Lower = vu_decodeLower(lo,addr);
}
return inst;
}
}
// tools/cpu/vu/disasm.go:53:1
std::string vu_DisasmMacro(uint32_t w){
{
uint32_t op = cast<uint32_t>((w & cast<uint32_t>(63ULL)));
{
if ((op < cast<uint32_t>(48ULL))){
return (std::string("v",1) + vu_decodeUpper(w));
}
else if ((op == cast<uint32_t>(56ULL))){
return go_fmt_Sprintf(std::string("vcallms 0x%X",12),cast<uint32_t>((cast<uint32_t>((shr<uint32_t>(w,cast<int64_t>(6ULL)) & cast<uint32_t>(32767ULL))) * cast<uint32_t>(8ULL))));
}
else if ((op == cast<uint32_t>(57ULL))){
return std::string("vcallmsr",8);
}
else if ((op < cast<uint32_t>(60ULL))){
return (std::string("v",1) + vu_decodeLowerSpecial(w));
}
else {
{
uint32_t op2 = cast<uint32_t>((cast<uint32_t>((shr<uint32_t>(w,cast<int64_t>(4ULL)) & cast<uint32_t>(124ULL))) | cast<uint32_t>((w & cast<uint32_t>(3ULL)))));
if ((op2 <= cast<uint32_t>(47ULL))) {
return (std::string("v",1) + vu_decodeUpper(w));
}
}
return (std::string("v",1) + vu_decodeLowerSpecial(w));
}
}
tmp1:;
}
}
// tools/cpu/vu/disasm.go:80:1
std::string vu_dest(uint32_t i){
{
return vu_destNames[cast<uint32_t>((shr<uint32_t>(i,cast<int64_t>(21ULL)) & cast<uint32_t>(15ULL)))];
}
}
// tools/cpu/vu/disasm.go:81:1
uint32_t vu_ft(uint32_t i){
{
return cast<uint32_t>((shr<uint32_t>(i,cast<int64_t>(16ULL)) & cast<uint32_t>(31ULL)));
}
}
// tools/cpu/vu/disasm.go:82:1
uint32_t vu_fs(uint32_t i){
{
return cast<uint32_t>((shr<uint32_t>(i,cast<int64_t>(11ULL)) & cast<uint32_t>(31ULL)));
}
}
// tools/cpu/vu/disasm.go:83:1
uint32_t vu_fd(uint32_t i){
{
return cast<uint32_t>((shr<uint32_t>(i,cast<int64_t>(6ULL)) & cast<uint32_t>(31ULL)));
}
}
// tools/cpu/vu/disasm.go:86:1
std::string vu_decodeUpper(uint32_t i){
{
uint32_t op = cast<uint32_t>((i & cast<uint32_t>(63ULL)));
auto tmp2 = std::make_tuple(vu_dest(i),vu_ft(i),vu_fs(i),vu_fd(i));
std::string d = std::get<0>(tmp2);
uint32_t t = std::get<1>(tmp2);
uint32_t s = std::get<2>(tmp2);
uint32_t f = std::get<3>(tmp2);
std::string bc = vu_bcNames[cast<uint32_t>((i & cast<uint32_t>(3ULL)))];
auto three = [&](std::string name)->std::string{
return go_fmt_Sprintf(std::string("%s.%s vf%02d, vf%02d, vf%02d",28),name,d,f,s,t);
}
;
auto threeBC = [&](std::string name)->std::string{
return go_fmt_Sprintf(std::string("%s%s.%s vf%02d, vf%02d, vf%02d%s",32),name,bc,d,f,s,t,bc);
}
;
auto threeQ = [&](std::string name)->std::string{
return go_fmt_Sprintf(std::string("%sq.%s vf%02d, vf%02d, q",24),name,d,f,s);
}
;
auto threeI = [&](std::string name)->std::string{
return go_fmt_Sprintf(std::string("%si.%s vf%02d, vf%02d, i",24),name,d,f,s);
}
;
{
if ((op < cast<uint32_t>(4ULL))){
return threeBC(std::string("add",3));
}
else if ((op < cast<uint32_t>(8ULL))){
return threeBC(std::string("sub",3));
}
else if ((op < cast<uint32_t>(12ULL))){
return threeBC(std::string("madd",4));
}
else if ((op < cast<uint32_t>(16ULL))){
return threeBC(std::string("msub",4));
}
else if ((op < cast<uint32_t>(20ULL))){
return threeBC(std::string("max",3));
}
else if ((op < cast<uint32_t>(24ULL))){
return threeBC(std::string("mini",4));
}
else if ((op < cast<uint32_t>(28ULL))){
return threeBC(std::string("mul",3));
}
else if ((op == cast<uint32_t>(28ULL))){
return threeQ(std::string("mul",3));
}
else if ((op == cast<uint32_t>(29ULL))){
return threeI(std::string("max",3));
}
else if ((op == cast<uint32_t>(30ULL))){
return threeI(std::string("mul",3));
}
else if ((op == cast<uint32_t>(31ULL))){
return threeI(std::string("mini",4));
}
else if ((op == cast<uint32_t>(32ULL))){
return threeQ(std::string("add",3));
}
else if ((op == cast<uint32_t>(33ULL))){
return threeQ(std::string("madd",4));
}
else if ((op == cast<uint32_t>(34ULL))){
return threeI(std::string("add",3));
}
else if ((op == cast<uint32_t>(35ULL))){
return threeI(std::string("madd",4));
}
else if ((op == cast<uint32_t>(36ULL))){
return threeQ(std::string("sub",3));
}
else if ((op == cast<uint32_t>(37ULL))){
return threeQ(std::string("msub",4));
}
else if ((op == cast<uint32_t>(38ULL))){
return threeI(std::string("sub",3));
}
else if ((op == cast<uint32_t>(39ULL))){
return threeI(std::string("msub",4));
}
else if ((op == cast<uint32_t>(40ULL))){
return three(std::string("add",3));
}
else if ((op == cast<uint32_t>(41ULL))){
return three(std::string("madd",4));
}
else if ((op == cast<uint32_t>(42ULL))){
return three(std::string("mul",3));
}
else if ((op == cast<uint32_t>(43ULL))){
return three(std::string("max",3));
}
else if ((op == cast<uint32_t>(44ULL))){
return three(std::string("sub",3));
}
else if ((op == cast<uint32_t>(45ULL))){
return three(std::string("msub",4));
}
else if ((op == cast<uint32_t>(46ULL))){
return three(std::string("opmsub",6));
}
else if ((op == cast<uint32_t>(47ULL))){
return three(std::string("mini",4));
}
else if ((op >= cast<uint32_t>(60ULL))){
return vu_decodeUpperWide(i);
}
}
tmp3:;
return go_fmt_Sprintf(std::string("upper? 0x%08X",13),i);
}
}
// tools/cpu/vu/disasm.go:166:1
std::string vu_decodeUpperWide(uint32_t i){
{
uint32_t op2 = cast<uint32_t>((cast<uint32_t>((shr<uint32_t>(i,cast<int64_t>(4ULL)) & cast<uint32_t>(124ULL))) | cast<uint32_t>((i & cast<uint32_t>(3ULL)))));
auto tmp4 = std::make_tuple(vu_dest(i),vu_ft(i),vu_fs(i));
std::string d = std::get<0>(tmp4);
uint32_t t = std::get<1>(tmp4);
uint32_t s = std::get<2>(tmp4);
std::string bc = vu_bcNames[cast<uint32_t>((i & cast<uint32_t>(3ULL)))];
auto acc = [&](std::string name)->std::string{
return go_fmt_Sprintf(std::string("%s.%s acc, vf%02d, vf%02d",25),name,d,s,t);
}
;
auto accBC = [&](std::string name)->std::string{
return go_fmt_Sprintf(std::string("%s%s.%s acc, vf%02d, vf%02d%s",29),name,bc,d,s,t,bc);
}
;
auto accQ = [&](std::string name)->std::string{
return go_fmt_Sprintf(std::string("%sq.%s acc, vf%02d, q",21),name,d,s);
}
;
auto accI = [&](std::string name)->std::string{
return go_fmt_Sprintf(std::string("%si.%s acc, vf%02d, i",21),name,d,s);
}
;
auto conv = [&](std::string name)->std::string{
return go_fmt_Sprintf(std::string("%s.%s vf%02d, vf%02d",20),name,d,t,s);
}
;
{
if ((op2 < cast<uint32_t>(4ULL))){
return accBC(std::string("adda",4));
}
else if ((op2 < cast<uint32_t>(8ULL))){
return accBC(std::string("suba",4));
}
else if ((op2 < cast<uint32_t>(12ULL))){
return accBC(std::string("madda",5));
}
else if ((op2 < cast<uint32_t>(16ULL))){
return accBC(std::string("msuba",5));
}
else if ((op2 == cast<uint32_t>(16ULL))){
return conv(std::string("itof0",5));
}
else if ((op2 == cast<uint32_t>(17ULL))){
return conv(std::string("itof4",5));
}
else if ((op2 == cast<uint32_t>(18ULL))){
return conv(std::string("itof12",6));
}
else if ((op2 == cast<uint32_t>(19ULL))){
return conv(std::string("itof15",6));
}
else if ((op2 == cast<uint32_t>(20ULL))){
return conv(std::string("ftoi0",5));
}
else if ((op2 == cast<uint32_t>(21ULL))){
return conv(std::string("ftoi4",5));
}
else if ((op2 == cast<uint32_t>(22ULL))){
return conv(std::string("ftoi12",6));
}
else if ((op2 == cast<uint32_t>(23ULL))){
return conv(std::string("ftoi15",6));
}
else if ((op2 < cast<uint32_t>(28ULL))){
return accBC(std::string("mula",4));
}
else if ((op2 == cast<uint32_t>(28ULL))){
return accQ(std::string("mula",4));
}
else if ((op2 == cast<uint32_t>(29ULL))){
return conv(std::string("abs",3));
}
else if ((op2 == cast<uint32_t>(30ULL))){
return accI(std::string("mula",4));
}
else if ((op2 == cast<uint32_t>(31ULL))){
return go_fmt_Sprintf(std::string("clipw.xyz vf%02d, vf%02dw",25),vu_fs(i),vu_ft(i));
}
else if ((op2 == cast<uint32_t>(32ULL))){
return accQ(std::string("adda",4));
}
else if ((op2 == cast<uint32_t>(33ULL))){
return accQ(std::string("madda",5));
}
else if ((op2 == cast<uint32_t>(34ULL))){
return accI(std::string("adda",4));
}
else if ((op2 == cast<uint32_t>(35ULL))){
return accI(std::string("madda",5));
}
else if ((op2 == cast<uint32_t>(36ULL))){
return accQ(std::string("suba",4));
}
else if ((op2 == cast<uint32_t>(37ULL))){
return accQ(std::string("msuba",5));
}
else if ((op2 == cast<uint32_t>(38ULL))){
return accI(std::string("suba",4));
}
else if ((op2 == cast<uint32_t>(39ULL))){
return accI(std::string("msuba",5));
}
else if ((op2 == cast<uint32_t>(40ULL))){
return acc(std::string("adda",4));
}
else if ((op2 == cast<uint32_t>(41ULL))){
return acc(std::string("madda",5));
}
else if ((op2 == cast<uint32_t>(42ULL))){
return acc(std::string("mula",4));
}
else if ((op2 == cast<uint32_t>(44ULL))){
return acc(std::string("suba",4));
}
else if ((op2 == cast<uint32_t>(45ULL))){
return acc(std::string("msuba",5));
}
else if ((op2 == cast<uint32_t>(46ULL))){
return acc(std::string("opmula",6));
}
else if ((op2 == cast<uint32_t>(47ULL))){
return std::string("nop",3);
}
}
tmp5:;
return go_fmt_Sprintf(std::string("upper? 0x%08X (wide 0x%02X)",27),i,op2);
}
}
// tools/cpu/vu/disasm.go:257:1
std::string vu_decodeLower(uint32_t i,uint32_t addr){
{
uint32_t op7 = cast<uint32_t>((shr<uint32_t>(i,cast<int64_t>(25ULL)) & cast<uint32_t>(127ULL)));
auto tmp6 = std::make_tuple(vu_dest(i),vu_ft(i),vu_fs(i));
std::string d = std::get<0>(tmp6);
uint32_t t = std::get<1>(tmp6);
uint32_t s = std::get<2>(tmp6);
int32_t imm11 = shr<int32_t>(cast<int32_t>(shl<uint32_t>(i,cast<int64_t>(21ULL))),cast<int64_t>(21ULL));
uint32_t imm15 = cast<uint32_t>((cast<uint32_t>((shr<uint32_t>(i,cast<int64_t>(10ULL)) & cast<uint32_t>(30720ULL))) | cast<uint32_t>((i & cast<uint32_t>(2047ULL)))));
auto branch = [&](std::string name,std::string args)->std::string{
uint32_t target = cast<uint32_t>((cast<uint32_t>((addr + cast<uint32_t>(8ULL))) + cast<uint32_t>((cast<uint32_t>(imm11) * cast<uint32_t>(8ULL)))));
if ((args != std::string("",0))) {
args += std::string(", ",2);
}
return go_fmt_Sprintf(std::string("%s %s0x%X",9),name,args,target);
}
;
{
switch(op7){
case cast<uint32_t>(0ULL):{
return go_fmt_Sprintf(std::string("lq.%s vf%02d, %d(vi%02d)",24),d,t,imm11,s);
break;}
case cast<uint32_t>(1ULL):{
return go_fmt_Sprintf(std::string("sq.%s vf%02d, %d(vi%02d)",24),d,s,imm11,t);
break;}
case cast<uint32_t>(4ULL):{
return go_fmt_Sprintf(std::string("ilw.%s vi%02d, %d(vi%02d)",25),d,t,imm11,s);
break;}
case cast<uint32_t>(5ULL):{
return go_fmt_Sprintf(std::string("isw.%s vi%02d, %d(vi%02d)",25),d,t,imm11,s);
break;}
case cast<uint32_t>(8ULL):{
return go_fmt_Sprintf(std::string("iaddiu vi%02d, vi%02d, %d",25),t,s,imm15);
break;}
case cast<uint32_t>(9ULL):{
return go_fmt_Sprintf(std::string("isubiu vi%02d, vi%02d, %d",25),t,s,imm15);
break;}
case cast<uint32_t>(16ULL):{
return go_fmt_Sprintf(std::string("fceq 0x%06X",11),cast<uint32_t>((i & cast<uint32_t>(16777215ULL))));
break;}
case cast<uint32_t>(17ULL):{
return go_fmt_Sprintf(std::string("fcset 0x%06X",12),cast<uint32_t>((i & cast<uint32_t>(16777215ULL))));
break;}
case cast<uint32_t>(18ULL):{
return go_fmt_Sprintf(std::string("fcand 0x%06X",12),cast<uint32_t>((i & cast<uint32_t>(16777215ULL))));
break;}
case cast<uint32_t>(19ULL):{
return go_fmt_Sprintf(std::string("fcor 0x%06X",11),cast<uint32_t>((i & cast<uint32_t>(16777215ULL))));
break;}
case cast<uint32_t>(20ULL):{
return go_fmt_Sprintf(std::string("fseq vi%02d, 0x%03X",19),t,cast<uint32_t>((cast<uint32_t>((shr<uint32_t>(i,cast<int64_t>(10ULL)) & cast<uint32_t>(2048ULL))) | cast<uint32_t>((i & cast<uint32_t>(2047ULL))))));
break;}
case cast<uint32_t>(21ULL):{
return go_fmt_Sprintf(std::string("fsset 0x%03X",12),cast<uint32_t>((cast<uint32_t>((shr<uint32_t>(i,cast<int64_t>(10ULL)) & cast<uint32_t>(2048ULL))) | cast<uint32_t>((i & cast<uint32_t>(2047ULL))))));
break;}
case cast<uint32_t>(22ULL):{
return go_fmt_Sprintf(std::string("fsand vi%02d, 0x%03X",20),t,cast<uint32_t>((cast<uint32_t>((shr<uint32_t>(i,cast<int64_t>(10ULL)) & cast<uint32_t>(2048ULL))) | cast<uint32_t>((i & cast<uint32_t>(2047ULL))))));
break;}
case cast<uint32_t>(23ULL):{
return go_fmt_Sprintf(std::string("fsor vi%02d, 0x%03X",19),t,cast<uint32_t>((cast<uint32_t>((shr<uint32_t>(i,cast<int64_t>(10ULL)) & cast<uint32_t>(2048ULL))) | cast<uint32_t>((i & cast<uint32_t>(2047ULL))))));
break;}
case cast<uint32_t>(24ULL):{
return go_fmt_Sprintf(std::string("fmeq vi%02d, vi%02d",19),t,s);
break;}
case cast<uint32_t>(26ULL):{
return go_fmt_Sprintf(std::string("fmand vi%02d, vi%02d",20),t,s);
break;}
case cast<uint32_t>(27ULL):{
return go_fmt_Sprintf(std::string("fmor vi%02d, vi%02d",19),t,s);
break;}
case cast<uint32_t>(28ULL):{
return go_fmt_Sprintf(std::string("fcget vi%02d",12),t);
break;}
case cast<uint32_t>(32ULL):{
return branch(std::string("b",1),std::string("",0));
break;}
case cast<uint32_t>(33ULL):{
return branch(std::string("bal",3),go_fmt_Sprintf(std::string("vi%02d",6),t));
break;}
case cast<uint32_t>(36ULL):{
return go_fmt_Sprintf(std::string("jr vi%02d",9),s);
break;}
case cast<uint32_t>(37ULL):{
return go_fmt_Sprintf(std::string("jalr vi%02d, vi%02d",19),t,s);
break;}
case cast<uint32_t>(40ULL):{
return branch(std::string("ibeq",4),go_fmt_Sprintf(std::string("vi%02d, vi%02d",14),t,s));
break;}
case cast<uint32_t>(41ULL):{
return branch(std::string("ibne",4),go_fmt_Sprintf(std::string("vi%02d, vi%02d",14),t,s));
break;}
case cast<uint32_t>(44ULL):{
return branch(std::string("ibltz",5),go_fmt_Sprintf(std::string("vi%02d",6),s));
break;}
case cast<uint32_t>(45ULL):{
return branch(std::string("ibgtz",5),go_fmt_Sprintf(std::string("vi%02d",6),s));
break;}
case cast<uint32_t>(46ULL):{
return branch(std::string("iblez",5),go_fmt_Sprintf(std::string("vi%02d",6),s));
break;}
case cast<uint32_t>(47ULL):{
return branch(std::string("ibgez",5),go_fmt_Sprintf(std::string("vi%02d",6),s));
break;}
case cast<uint32_t>(64ULL):{
return vu_decodeLowerSpecial(i);
break;}
}}
if ((i == cast<uint32_t>(0ULL))) {
return std::string("nop",3);
}
return go_fmt_Sprintf(std::string("lower? 0x%08X",13),i);
}
}
// tools/cpu/vu/disasm.go:337:1
std::string vu_decodeLowerSpecial(uint32_t i){
{
auto tmp7 = std::make_tuple(vu_dest(i),vu_ft(i),vu_fs(i),vu_fd(i));
std::string d = std::get<0>(tmp7);
uint32_t t = std::get<1>(tmp7);
uint32_t s = std::get<2>(tmp7);
uint32_t f = std::get<3>(tmp7);
uint32_t op = cast<uint32_t>((i & cast<uint32_t>(63ULL)));
{
switch(op){
case cast<uint32_t>(48ULL):{
return go_fmt_Sprintf(std::string("iadd vi%02d, vi%02d, vi%02d",27),f,s,t);
break;}
case cast<uint32_t>(49ULL):{
return go_fmt_Sprintf(std::string("isub vi%02d, vi%02d, vi%02d",27),f,s,t);
break;}
case cast<uint32_t>(50ULL):{
return go_fmt_Sprintf(std::string("iaddi vi%02d, vi%02d, %d",24),t,s,shr<int32_t>(cast<int32_t>(shl<uint32_t>(i,cast<int64_t>(21ULL))),cast<int64_t>(27ULL)));
break;}
case cast<uint32_t>(52ULL):{
return go_fmt_Sprintf(std::string("iand vi%02d, vi%02d, vi%02d",27),f,s,t);
break;}
case cast<uint32_t>(53ULL):{
return go_fmt_Sprintf(std::string("ior vi%02d, vi%02d, vi%02d",26),f,s,t);
break;}
}}
if ((op < cast<uint32_t>(60ULL))) {
if ((i == cast<uint32_t>(0ULL))) {
return std::string("nop",3);
}
return go_fmt_Sprintf(std::string("lower? 0x%08X",13),i);
}
uint32_t op2 = cast<uint32_t>((cast<uint32_t>((shr<uint32_t>(i,cast<int64_t>(4ULL)) & cast<uint32_t>(124ULL))) | cast<uint32_t>((i & cast<uint32_t>(3ULL)))));
std::string fsf = vu_bcNames[cast<uint32_t>((shr<uint32_t>(i,cast<int64_t>(21ULL)) & cast<uint32_t>(3ULL)))];
std::string ftf = vu_bcNames[cast<uint32_t>((shr<uint32_t>(i,cast<int64_t>(23ULL)) & cast<uint32_t>(3ULL)))];
{
switch(op2){
case cast<uint32_t>(48ULL):{
return go_fmt_Sprintf(std::string("move.%s vf%02d, vf%02d",22),d,t,s);
break;}
case cast<uint32_t>(49ULL):{
return go_fmt_Sprintf(std::string("mr32.%s vf%02d, vf%02d",22),d,t,s);
break;}
case cast<uint32_t>(52ULL):{
return go_fmt_Sprintf(std::string("lqi.%s vf%02d, (vi%02d++)",25),d,t,s);
break;}
case cast<uint32_t>(53ULL):{
return go_fmt_Sprintf(std::string("sqi.%s vf%02d, (vi%02d++)",25),d,s,t);
break;}
case cast<uint32_t>(54ULL):{
return go_fmt_Sprintf(std::string("lqd.%s vf%02d, (--vi%02d)",25),d,t,s);
break;}
case cast<uint32_t>(55ULL):{
return go_fmt_Sprintf(std::string("sqd.%s vf%02d, (--vi%02d)",25),d,s,t);
break;}
case cast<uint32_t>(56ULL):{
return go_fmt_Sprintf(std::string("div q, vf%02d%s, vf%02d%s",25),s,fsf,t,ftf);
break;}
case cast<uint32_t>(57ULL):{
return go_fmt_Sprintf(std::string("sqrt q, vf%02d%s",16),t,ftf);
break;}
case cast<uint32_t>(58ULL):{
return go_fmt_Sprintf(std::string("rsqrt q, vf%02d%s, vf%02d%s",27),s,fsf,t,ftf);
break;}
case cast<uint32_t>(59ULL):{
return std::string("waitq",5);
break;}
case cast<uint32_t>(60ULL):{
return go_fmt_Sprintf(std::string("mtir vi%02d, vf%02d%s",21),t,s,fsf);
break;}
case cast<uint32_t>(61ULL):{
return go_fmt_Sprintf(std::string("mfir.%s vf%02d, vi%02d",22),d,t,s);
break;}
case cast<uint32_t>(62ULL):{
return go_fmt_Sprintf(std::string("ilwr.%s vi%02d, (vi%02d)",24),d,t,s);
break;}
case cast<uint32_t>(63ULL):{
return go_fmt_Sprintf(std::string("iswr.%s vi%02d, (vi%02d)",24),d,t,s);
break;}
case cast<uint32_t>(64ULL):{
return std::string("rnext",5);
break;}
case cast<uint32_t>(65ULL):{
return go_fmt_Sprintf(std::string("rget.%s vf%02d, r",17),d,t);
break;}
case cast<uint32_t>(66ULL):{
return go_fmt_Sprintf(std::string("rinit r, vf%02d%s",17),s,fsf);
break;}
case cast<uint32_t>(67ULL):{
return go_fmt_Sprintf(std::string("rxor r, vf%02d%s",16),s,fsf);
break;}
case cast<uint32_t>(100ULL):{
return go_fmt_Sprintf(std::string("mfp.%s vf%02d, p",16),d,t);
break;}
case cast<uint32_t>(104ULL):{
return go_fmt_Sprintf(std::string("xtop vi%02d",11),t);
break;}
case cast<uint32_t>(105ULL):{
return go_fmt_Sprintf(std::string("xitop vi%02d",12),t);
break;}
case cast<uint32_t>(108ULL):{
return go_fmt_Sprintf(std::string("xgkick vi%02d",13),s);
break;}
case cast<uint32_t>(112ULL):{
return go_fmt_Sprintf(std::string("esadd p, vf%02d",15),s);
break;}
case cast<uint32_t>(113ULL):{
return go_fmt_Sprintf(std::string("ersadd p, vf%02d",16),s);
break;}
case cast<uint32_t>(114ULL):{
return go_fmt_Sprintf(std::string("eleng p, vf%02d",15),s);
break;}
case cast<uint32_t>(115ULL):{
return go_fmt_Sprintf(std::string("erleng p, vf%02d",16),s);
break;}
case cast<uint32_t>(116ULL):{
return go_fmt_Sprintf(std::string("eatanxy p, vf%02d",17),s);
break;}
case cast<uint32_t>(117ULL):{
return go_fmt_Sprintf(std::string("eatanxz p, vf%02d",17),s);
break;}
case cast<uint32_t>(118ULL):{
return go_fmt_Sprintf(std::string("esum p, vf%02d",14),s);
break;}
case cast<uint32_t>(120ULL):{
return go_fmt_Sprintf(std::string("esqrt p, vf%02d%s",17),s,fsf);
break;}
case cast<uint32_t>(121ULL):{
return go_fmt_Sprintf(std::string("ersqrt p, vf%02d%s",18),s,fsf);
break;}
case cast<uint32_t>(122ULL):{
return go_fmt_Sprintf(std::string("ercpr p, vf%02d%s",17),s,fsf);
break;}
case cast<uint32_t>(123ULL):{
return std::string("waitp",5);
break;}
case cast<uint32_t>(124ULL):{
return go_fmt_Sprintf(std::string("esin p, vf%02d%s",16),s,fsf);
break;}
case cast<uint32_t>(125ULL):{
return go_fmt_Sprintf(std::string("eatan p, vf%02d%s",17),s,fsf);
break;}
case cast<uint32_t>(126ULL):{
return go_fmt_Sprintf(std::string("eexp p, vf%02d%s",16),s,fsf);
break;}
}}
return go_fmt_Sprintf(std::string("lower? 0x%08X (special2 0x%02X)",31),i,op2);
}
}
// tools/cpu/vu/exec.go:83:1
void vu_VU_ResetBranchLog(vu_VU* v,int64_t size){
{
if ((!v->BranchLog)) {
v->BranchLog = Slice<std::array<uint32_t,2>>::make(size);
}
v->branchNext = cast<int64_t>(0ULL);
}
}
// tools/cpu/vu/exec.go:91:1
int64_t vu_VU_BranchCount(vu_VU* v){
{
return v->branchNext;
}
}
// tools/cpu/vu/exec.go:94:1
void vu_VU_logBranch(vu_VU* v,uint32_t from,uint32_t to){
{
if ((!v->BranchLog)) {
return ;
}
v->BranchLog[modi<int64_t>(v->branchNext,len(v->BranchLog))] = std::array<uint32_t,2>{from,to};
v->branchNext++;
}
}
// tools/cpu/vu/exec.go:103:1
Slice<std::array<uint32_t,2>> vu_VU_BranchTrail(vu_VU* v){
{
if ((!v->BranchLog)) {
return {};
}
int64_t n = v->branchNext;
if ((n > len(v->BranchLog))) {
n = len(v->BranchLog);
}
Slice<std::array<uint32_t,2>> out = Slice<std::array<uint32_t,2>>::make(cast<int64_t>(0ULL),n);
{int64_t i = cast<int64_t>((v->branchNext - n));for (;(i < v->branchNext);i++){
out = append(out,Slice<std::array<uint32_t,2>>{v->BranchLog[modi<int64_t>(i,len(v->BranchLog))]});
}
}return out;
}
}
// tools/cpu/vu/exec.go:126:1
vu_VU* vu_New(Slice<uint8_t> micro,Slice<uint8_t> data){
{
vu_VU* v = arenaNew(vu_VU{{},{},{},{},{},{},{},micro,data,{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{}});
v->VF[cast<int64_t>(0ULL)][cast<int64_t>(3ULL)] = go_math_Float32bits(1.0);
return v;
}
}
// tools/cpu/vu/exec.go:135:1
std::tuple<int64_t,bool> vu_VU_Run(vu_VU* v,uint32_t start,int64_t maxSteps){
{
if ((len(v->Micro) == cast<int64_t>(0ULL))) {
return {cast<int64_t>(0ULL),false};
}
uint32_t mask = (cast<uint32_t>(cast<int64_t>((len(v->Micro) - cast<int64_t>(1ULL)))) & ~(cast<uint32_t>(7ULL)));
v->PC = cast<uint32_t>((start & mask));
int64_t delayed = cast<int64_t>(-1ULL);
int64_t endIn = cast<int64_t>(-1ULL);
vu_upperResult res={};
{int64_t steps = cast<int64_t>(0ULL);for (;(steps < maxSteps);steps++){
uint64_t raw = vu_le64m(rrBorrow(v->Micro,v->PC,len(v->Micro)));
uint32_t up = cast<uint32_t>(shr<uint64_t>(raw,cast<int64_t>(32ULL)));
uint32_t lo = cast<uint32_t>(raw);
if (bool(v->Trace)) {
v->Trace(v,v->PC,raw);
}
v->visMac = v->flagPipe[cast<int64_t>(0ULL)].mac;
v->visStatus = v->flagPipe[cast<int64_t>(0ULL)].status;
v->visClip = v->flagPipe[cast<int64_t>(0ULL)].clip;
vu_VU_execUpper(v,up,(&res));
int64_t taken=cast<int64_t>(-1ULL);
if ((cast<uint32_t>((up & cast<uint32_t>(2147483648ULL))) != cast<uint32_t>(0ULL))) {
v->I = vu_sane(go_math_Float32frombits(lo));
}
else {
taken = vu_VU_execLower(v,lo);
}
vu_VU_commitUpper(v,(&res));
auto tmp8 = std::make_tuple(v->flagPipe[cast<int64_t>(1ULL)],v->flagPipe[cast<int64_t>(2ULL)],v->flagPipe[cast<int64_t>(3ULL)]);
v->flagPipe[cast<int64_t>(0ULL)] = std::get<0>(tmp8);
v->flagPipe[cast<int64_t>(1ULL)] = std::get<1>(tmp8);
v->flagPipe[cast<int64_t>(2ULL)] = std::get<2>(tmp8);
v->flagPipe[cast<int64_t>(3ULL)] = vu_flagVals{v->Mac,v->Status,v->Clip};
v->Steps++;
uint32_t next = cast<uint32_t>(((cast<uint32_t>((v->PC + cast<uint32_t>(8ULL)))) & mask));
if ((delayed >= cast<int64_t>(0ULL))) {
next = cast<uint32_t>((cast<uint32_t>(delayed) & mask));
delayed = cast<int64_t>(-1ULL);
vu_VU_logBranch(v,v->PC,next);
}
if ((taken >= cast<int64_t>(0ULL))) {
delayed = taken;
}
v->PC = next;
if ((endIn >= cast<int64_t>(0ULL))) {
endIn--;
if ((endIn < cast<int64_t>(0ULL))) {
return {cast<int64_t>((steps + cast<int64_t>(1ULL))),true};
}
}
else if ((cast<uint32_t>((up & cast<uint32_t>(1073741824ULL))) != cast<uint32_t>(0ULL))) {
endIn = cast<int64_t>(0ULL);
}
}
}return {maxSteps,false};
}
}
// tools/cpu/vu/exec.go:195:1
float vu_VU_getF(vu_VU* v,uint32_t r,uint32_t lane){
{
return go_math_Float32frombits(v->VF[r][lane]);
}
}
// tools/cpu/vu/exec.go:200:1
void vu_VU_setLanes(vu_VU* v,uint32_t r,uint32_t destMask,std::array<float,4> val){
{
if ((r == cast<uint32_t>(0ULL))) {
return ;
}
{uint32_t lane = cast<uint32_t>(0ULL);for (;(lane < cast<uint32_t>(4ULL));lane++){
if ((cast<uint32_t>((destMask & (shr<uint32_t>(cast<uint32_t>(8ULL),lane)))) != cast<uint32_t>(0ULL))) {
v->VF[r][lane] = go_math_Float32bits(vu_sane(val[lane]));
}
}
}}
}
// tools/cpu/vu/exec.go:211:1
void vu_VU_setBits(vu_VU* v,uint32_t r,uint32_t destMask,std::array<uint32_t,4> val){
{
if ((r == cast<uint32_t>(0ULL))) {
return ;
}
{uint32_t lane = cast<uint32_t>(0ULL);for (;(lane < cast<uint32_t>(4ULL));lane++){
if ((cast<uint32_t>((destMask & (shr<uint32_t>(cast<uint32_t>(8ULL),lane)))) != cast<uint32_t>(0ULL))) {
v->VF[r][lane] = val[lane];
}
}
}}
}
// tools/cpu/vu/exec.go:222:1
void vu_VU_setVI(vu_VU* v,uint32_t r,uint16_t val){
{
if (((r != cast<uint32_t>(0ULL)) && (r < cast<uint32_t>(16ULL)))) {
v->VI[cast<uint32_t>((r & cast<uint32_t>(15ULL)))] = val;
}
}
}
// tools/cpu/vu/exec.go:230:1
float vu_sane(float f){
{
if (go_math_IsNaN(cast<double>(f))) {
return cast<float>(0ULL);
}
if ((f > go_math_MaxFloat32)) {
return go_math_MaxFloat32;
}
if ((f < cast<float>(-go_math_MaxFloat32))) {
return cast<float>(-go_math_MaxFloat32);
}
return f;
}
}
// tools/cpu/vu/exec.go:243:1
uint64_t vu_le64m(Slice<uint8_t> b){
{
return cast<uint64_t>((cast<uint64_t>((cast<uint64_t>((cast<uint64_t>((cast<uint64_t>((cast<uint64_t>((cast<uint64_t>((cast<uint64_t>(b[cast<int64_t>(0ULL)]) | shl<uint64_t>(cast<uint64_t>(b[cast<int64_t>(1ULL)]),cast<int64_t>(8ULL)))) | shl<uint64_t>(cast<uint64_t>(b[cast<int64_t>(2ULL)]),cast<int64_t>(16ULL)))) | shl<uint64_t>(cast<uint64_t>(b[cast<int64_t>(3ULL)]),cast<int64_t>(24ULL)))) | shl<uint64_t>(cast<uint64_t>(b[cast<int64_t>(4ULL)]),cast<int64_t>(32ULL)))) | shl<uint64_t>(cast<uint64_t>(b[cast<int64_t>(5ULL)]),cast<int64_t>(40ULL)))) | shl<uint64_t>(cast<uint64_t>(b[cast<int64_t>(6ULL)]),cast<int64_t>(48ULL)))) | shl<uint64_t>(cast<uint64_t>(b[cast<int64_t>(7ULL)]),cast<int64_t>(56ULL))));
}
}
// tools/cpu/vu/exec.go:270:1
void vu_VU_execUpper(vu_VU* v,uint32_t i,vu_upperResult* res){
{
res->kind = cast<uint8_t>(0ULL);
uint32_t op = cast<uint32_t>((i & cast<uint32_t>(63ULL)));
if (((op < cast<uint32_t>(48ULL)) || (op >= cast<uint32_t>(60ULL)))) {
}
else {
return ;
}
uint32_t d = cast<uint32_t>((shr<uint32_t>(i,cast<int64_t>(21ULL)) & cast<uint32_t>(15ULL)));
auto tmp9 = std::make_tuple(cast<uint32_t>((shr<uint32_t>(i,cast<int64_t>(16ULL)) & cast<uint32_t>(31ULL))),cast<uint32_t>((shr<uint32_t>(i,cast<int64_t>(11ULL)) & cast<uint32_t>(31ULL))),cast<uint32_t>((shr<uint32_t>(i,cast<int64_t>(6ULL)) & cast<uint32_t>(31ULL))));
uint32_t t = std::get<0>(tmp9);
uint32_t s = std::get<1>(tmp9);
uint32_t f = std::get<2>(tmp9);
std::array<float,4> vs={};
std::array<float,4> vt={};
{uint32_t l = cast<uint32_t>(0ULL);for (;(l < cast<uint32_t>(4ULL));l++){
vs[l] = vu_VU_getF(v,s,l);
vt[l] = vu_VU_getF(v,t,l);
}
}auto setB = [&](float x)->std::array<float,4>{
return std::array<float,4>{x,x,x,x};
}
;
auto vf = [&](uint32_t r,std::array<float,4> val)->void{
auto tmp10 = std::make_tuple(cast<uint8_t>(1ULL),r,d,val);
res->kind = std::get<0>(tmp10);
res->reg = std::get<1>(tmp10);
res->mask = std::get<2>(tmp10);
res->val = std::get<3>(tmp10);
}
;
auto acc = [&](std::array<float,4> val)->void{
auto tmp11 = std::make_tuple(cast<uint8_t>(2ULL),d,val);
res->kind = std::get<0>(tmp11);
res->mask = std::get<1>(tmp11);
res->val = std::get<2>(tmp11);
}
;
auto addv = [&](std::array<float,4> b)->std::array<float,4>{
return std::array<float,4>{(vs[cast<int64_t>(0ULL)] + b[cast<int64_t>(0ULL)]),(vs[cast<int64_t>(1ULL)] + b[cast<int64_t>(1ULL)]),(vs[cast<int64_t>(2ULL)] + b[cast<int64_t>(2ULL)]),(vs[cast<int64_t>(3ULL)] + b[cast<int64_t>(3ULL)])};
}
;
auto subv = [&](std::array<float,4> b)->std::array<float,4>{
return std::array<float,4>{(vs[cast<int64_t>(0ULL)] - b[cast<int64_t>(0ULL)]),(vs[cast<int64_t>(1ULL)] - b[cast<int64_t>(1ULL)]),(vs[cast<int64_t>(2ULL)] - b[cast<int64_t>(2ULL)]),(vs[cast<int64_t>(3ULL)] - b[cast<int64_t>(3ULL)])};
}
;
auto mulv = [&](std::array<float,4> b)->std::array<float,4>{
return std::array<float,4>{(vs[cast<int64_t>(0ULL)] * b[cast<int64_t>(0ULL)]),(vs[cast<int64_t>(1ULL)] * b[cast<int64_t>(1ULL)]),(vs[cast<int64_t>(2ULL)] * b[cast<int64_t>(2ULL)]),(vs[cast<int64_t>(3ULL)] * b[cast<int64_t>(3ULL)])};
}
;
auto maddv = [&](std::array<float,4> b)->std::array<float,4>{
return std::array<float,4>{(v->ACC[cast<int64_t>(0ULL)] + (vs[cast<int64_t>(0ULL)] * b[cast<int64_t>(0ULL)])),(v->ACC[cast<int64_t>(1ULL)] + (vs[cast<int64_t>(1ULL)] * b[cast<int64_t>(1ULL)])),(v->ACC[cast<int64_t>(2ULL)] + (vs[cast<int64_t>(2ULL)] * b[cast<int64_t>(2ULL)])),(v->ACC[cast<int64_t>(3ULL)] + (vs[cast<int64_t>(3ULL)] * b[cast<int64_t>(3ULL)]))};
}
;
auto msubv = [&](std::array<float,4> b)->std::array<float,4>{
return std::array<float,4>{(v->ACC[cast<int64_t>(0ULL)] - (vs[cast<int64_t>(0ULL)] * b[cast<int64_t>(0ULL)])),(v->ACC[cast<int64_t>(1ULL)] - (vs[cast<int64_t>(1ULL)] * b[cast<int64_t>(1ULL)])),(v->ACC[cast<int64_t>(2ULL)] - (vs[cast<int64_t>(2ULL)] * b[cast<int64_t>(2ULL)])),(v->ACC[cast<int64_t>(3ULL)] - (vs[cast<int64_t>(3ULL)] * b[cast<int64_t>(3ULL)]))};
}
;
auto maxv = [&](std::array<float,4> b)->std::array<float,4>{
std::array<float,4> r={};
{int64_t l = cast<int64_t>(0ULL);for (;(l < cast<int64_t>(4ULL));l++){
if ((vs[l] > b[l])) {
r[l] = vs[l];
}
else {
r[l] = b[l];
}
}
}return r;
}
;
auto minv = [&](std::array<float,4> b)->std::array<float,4>{
std::array<float,4> r={};
{int64_t l = cast<int64_t>(0ULL);for (;(l < cast<int64_t>(4ULL));l++){
if ((vs[l] < b[l])) {
r[l] = vs[l];
}
else {
r[l] = b[l];
}
}
}return r;
}
;
float bcVal = vt[cast<uint32_t>((i & cast<uint32_t>(3ULL)))];
{
if ((op < cast<uint32_t>(4ULL))){
vf(f,addv(setB(bcVal)));
}
else if ((op < cast<uint32_t>(8ULL))){
vf(f,subv(setB(bcVal)));
}
else if ((op < cast<uint32_t>(12ULL))){
vf(f,maddv(setB(bcVal)));
}
else if ((op < cast<uint32_t>(16ULL))){
vf(f,msubv(setB(bcVal)));
}
else if ((op < cast<uint32_t>(20ULL))){
vf(f,maxv(setB(bcVal)));
}
else if ((op < cast<uint32_t>(24ULL))){
vf(f,minv(setB(bcVal)));
}
else if ((op < cast<uint32_t>(28ULL))){
vf(f,mulv(setB(bcVal)));
}
else if ((op == cast<uint32_t>(28ULL))){
vf(f,mulv(setB(v->Q)));
}
else if ((op == cast<uint32_t>(29ULL))){
vf(f,maxv(setB(v->I)));
}
else if ((op == cast<uint32_t>(30ULL))){
vf(f,mulv(setB(v->I)));
}
else if ((op == cast<uint32_t>(31ULL))){
vf(f,minv(setB(v->I)));
}
else if ((op == cast<uint32_t>(32ULL))){
vf(f,addv(setB(v->Q)));
}
else if ((op == cast<uint32_t>(33ULL))){
vf(f,maddv(setB(v->Q)));
}
else if ((op == cast<uint32_t>(34ULL))){
vf(f,addv(setB(v->I)));
}
else if ((op == cast<uint32_t>(35ULL))){
vf(f,maddv(setB(v->I)));
}
else if ((op == cast<uint32_t>(36ULL))){
vf(f,subv(setB(v->Q)));
}
else if ((op == cast<uint32_t>(37ULL))){
vf(f,msubv(setB(v->Q)));
}
else if ((op == cast<uint32_t>(38ULL))){
vf(f,subv(setB(v->I)));
}
else if ((op == cast<uint32_t>(39ULL))){
vf(f,msubv(setB(v->I)));
}
else if ((op == cast<uint32_t>(40ULL))){
vf(f,addv(vt));
}
else if ((op == cast<uint32_t>(41ULL))){
vf(f,maddv(vt));
}
else if ((op == cast<uint32_t>(42ULL))){
vf(f,mulv(vt));
}
else if ((op == cast<uint32_t>(43ULL))){
vf(f,maxv(vt));
}
else if ((op == cast<uint32_t>(44ULL))){
vf(f,subv(vt));
}
else if ((op == cast<uint32_t>(45ULL))){
vf(f,msubv(vt));
}
else if ((op == cast<uint32_t>(46ULL))){
auto tmp13 = std::make_tuple(cast<uint8_t>(1ULL),f,cast<uint32_t>((d & cast<uint32_t>(14ULL))));
res->kind = std::get<0>(tmp13);
res->reg = std::get<1>(tmp13);
res->mask = std::get<2>(tmp13);
res->val = std::array<float,4>{(v->ACC[cast<int64_t>(0ULL)] - (vs[cast<int64_t>(1ULL)] * vt[cast<int64_t>(2ULL)])),(v->ACC[cast<int64_t>(1ULL)] - (vs[cast<int64_t>(2ULL)] * vt[cast<int64_t>(0ULL)])),(v->ACC[cast<int64_t>(2ULL)] - (vs[cast<int64_t>(0ULL)] * vt[cast<int64_t>(1ULL)]))};
}
else if ((op == cast<uint32_t>(47ULL))){
vf(f,minv(vt));
}
else {
uint32_t op2 = cast<uint32_t>((cast<uint32_t>((shr<uint32_t>(i,cast<int64_t>(4ULL)) & cast<uint32_t>(124ULL))) | cast<uint32_t>((i & cast<uint32_t>(3ULL)))));
{
if ((op2 < cast<uint32_t>(4ULL))){
acc(addv(setB(bcVal)));
}
else if ((op2 < cast<uint32_t>(8ULL))){
acc(subv(setB(bcVal)));
}
else if ((op2 < cast<uint32_t>(12ULL))){
acc(maddv(setB(bcVal)));
}
else if ((op2 < cast<uint32_t>(16ULL))){
acc(msubv(setB(bcVal)));
}
else if ((op2 <= cast<uint32_t>(19ULL))){
float shift = std::array<float,4>{cast<float>(1ULL),cast<float>(16ULL),cast<float>(4096ULL),cast<float>(32768ULL)}[cast<uint32_t>((op2 - cast<uint32_t>(16ULL)))];
auto tmp15 = std::make_tuple(cast<uint8_t>(3ULL),t,d);
res->kind = std::get<0>(tmp15);
res->reg = std::get<1>(tmp15);
res->mask = std::get<2>(tmp15);
{uint32_t l = cast<uint32_t>(0ULL);for (;(l < cast<uint32_t>(4ULL));l++){
res->bits[l] = go_math_Float32bits(vu_sane((cast<float>(cast<int32_t>(v->VF[s][l])) / shift)));
}
}}
else if ((op2 <= cast<uint32_t>(23ULL))){
float shift = std::array<float,4>{cast<float>(1ULL),cast<float>(16ULL),cast<float>(4096ULL),cast<float>(32768ULL)}[cast<uint32_t>((op2 - cast<uint32_t>(20ULL)))];
auto tmp16 = std::make_tuple(cast<uint8_t>(3ULL),t,d);
res->kind = std::get<0>(tmp16);
res->reg = std::get<1>(tmp16);
res->mask = std::get<2>(tmp16);
{uint32_t l = cast<uint32_t>(0ULL);for (;(l < cast<uint32_t>(4ULL));l++){
double fv = (cast<double>(vu_VU_getF(v,s,l)) * cast<double>(shift));
{
if ((fv > go_math_MaxInt32)){
fv = go_math_MaxInt32;
}
else if ((fv < go_math_MinInt32)){
fv = go_math_MinInt32;
}
else if (go_math_IsNaN(fv)){
fv = cast<double>(0ULL);
}
}
tmp17:;
res->bits[l] = cast<uint32_t>(cast<int32_t>(fv));
}
}}
else if ((op2 < cast<uint32_t>(28ULL))){
acc(mulv(setB(bcVal)));
}
else if ((op2 == cast<uint32_t>(28ULL))){
acc(mulv(setB(v->Q)));
}
else if ((op2 == cast<uint32_t>(29ULL))){
auto tmp18 = std::make_tuple(cast<uint8_t>(3ULL),t,d);
res->kind = std::get<0>(tmp18);
res->reg = std::get<1>(tmp18);
res->mask = std::get<2>(tmp18);
{int64_t l = cast<int64_t>(0ULL);for (;(l < cast<int64_t>(4ULL));l++){
res->bits[l] = (v->VF[s][l] & ~(cast<uint32_t>(2147483648ULL)));
}
}}
else if ((op2 == cast<uint32_t>(30ULL))){
acc(mulv(setB(v->I)));
}
else if ((op2 == cast<uint32_t>(31ULL))){
float w = vt[cast<int64_t>(3ULL)];
if ((w < cast<float>(0ULL))) {
w = cast<float>(-w);
}
uint32_t bits={};
if ((vs[cast<int64_t>(0ULL)] > w)) {
bits |= cast<uint32_t>(1ULL);
}
if ((vs[cast<int64_t>(0ULL)] < cast<float>(-w))) {
bits |= cast<uint32_t>(2ULL);
}
if ((vs[cast<int64_t>(1ULL)] > w)) {
bits |= cast<uint32_t>(4ULL);
}
if ((vs[cast<int64_t>(1ULL)] < cast<float>(-w))) {
bits |= cast<uint32_t>(8ULL);
}
if ((vs[cast<int64_t>(2ULL)] > w)) {
bits |= cast<uint32_t>(16ULL);
}
if ((vs[cast<int64_t>(2ULL)] < cast<float>(-w))) {
bits |= cast<uint32_t>(32ULL);
}
auto tmp19 = std::make_tuple(cast<uint8_t>(4ULL),bits);
res->kind = std::get<0>(tmp19);
res->clip = std::get<1>(tmp19);
}
else if ((op2 == cast<uint32_t>(32ULL))){
acc(addv(setB(v->Q)));
}
else if ((op2 == cast<uint32_t>(33ULL))){
acc(maddv(setB(v->Q)));
}
else if ((op2 == cast<uint32_t>(34ULL))){
acc(addv(setB(v->I)));
}
else if ((op2 == cast<uint32_t>(35ULL))){
acc(maddv(setB(v->I)));
}
else if ((op2 == cast<uint32_t>(36ULL))){
acc(subv(setB(v->Q)));
}
else if ((op2 == cast<uint32_t>(37ULL))){
acc(msubv(setB(v->Q)));
}
else if ((op2 == cast<uint32_t>(38ULL))){
acc(subv(setB(v->I)));
}
else if ((op2 == cast<uint32_t>(39ULL))){
acc(msubv(setB(v->I)));
}
else if ((op2 == cast<uint32_t>(40ULL))){
acc(addv(vt));
}
else if ((op2 == cast<uint32_t>(41ULL))){
acc(maddv(vt));
}
else if ((op2 == cast<uint32_t>(42ULL))){
acc(mulv(vt));
}
else if ((op2 == cast<uint32_t>(44ULL))){
acc(subv(vt));
}
else if ((op2 == cast<uint32_t>(45ULL))){
acc(msubv(vt));
}
else if ((op2 == cast<uint32_t>(46ULL))){
auto tmp20 = std::make_tuple(cast<uint8_t>(2ULL),cast<uint32_t>((d & cast<uint32_t>(14ULL))));
res->kind = std::get<0>(tmp20);
res->mask = std::get<1>(tmp20);
res->val = std::array<float,4>{(vs[cast<int64_t>(1ULL)] * vt[cast<int64_t>(2ULL)]),(vs[cast<int64_t>(2ULL)] * vt[cast<int64_t>(0ULL)]),(vs[cast<int64_t>(0ULL)] * vt[cast<int64_t>(1ULL)])};
}
}
tmp14:;
}
}
tmp12:;
}
}
// tools/cpu/vu/exec.go:500:1
void vu_VU_commitUpper(vu_VU* v,vu_upperResult* res){
{
{
switch(res->kind){
case cast<uint8_t>(1ULL):{
vu_VU_setLanes(v,res->reg,res->mask,res->val);
vu_VU_macFlags(v,res->mask,res->val);
break;}
case cast<uint8_t>(2ULL):{
{uint32_t l = cast<uint32_t>(0ULL);for (;(l < cast<uint32_t>(4ULL));l++){
if ((cast<uint32_t>((res->mask & (shr<uint32_t>(cast<uint32_t>(8ULL),l)))) != cast<uint32_t>(0ULL))) {
v->ACC[l] = vu_sane(res->val[l]);
}
}
}vu_VU_macFlags(v,res->mask,res->val);
break;}
case cast<uint8_t>(3ULL):{
vu_VU_setBits(v,res->reg,res->mask,res->bits);
break;}
case cast<uint8_t>(4ULL):{
v->Clip = cast<uint32_t>(((cast<uint32_t>((shl<uint32_t>(v->Clip,cast<int64_t>(6ULL)) | res->clip))) & cast<uint32_t>(16777215ULL)));
break;}
}}
}
}
// tools/cpu/vu/exec.go:522:1
void vu_VU_macFlags(vu_VU* v,uint32_t destMask,std::array<float,4> val){
{
uint16_t z={};
uint16_t sgn={};
{uint32_t l = cast<uint32_t>(0ULL);for (;(l < cast<uint32_t>(4ULL));l++){
if ((cast<uint32_t>((destMask & (shr<uint32_t>(cast<uint32_t>(8ULL),l)))) == cast<uint32_t>(0ULL))) {
continue;
}
uint16_t bit = cast<uint16_t>(shr<uint16_t>(cast<uint16_t>(8ULL),l));
if ((val[l] == cast<float>(0ULL))) {
z |= bit;
}
if (go_math_Signbit(cast<double>(val[l]))) {
sgn |= bit;
}
}
}v->Mac = cast<uint16_t>((z | shl<uint16_t>(sgn,cast<int64_t>(4ULL))));
v->Status = cast<uint16_t>((v->Status & cast<uint16_t>(4032ULL)));
if ((z != cast<uint16_t>(0ULL))) {
v->Status |= cast<uint16_t>(65ULL);
}
if ((sgn != cast<uint16_t>(0ULL))) {
v->Status |= cast<uint16_t>(130ULL);
}
}
}
// tools/cpu/vu/exec.go:550:1
int64_t vu_VU_execLower(vu_VU* v,uint32_t i){
{
if ((i == cast<uint32_t>(0ULL))) {
return cast<int64_t>(-1ULL);
}
uint32_t op7 = cast<uint32_t>((shr<uint32_t>(i,cast<int64_t>(25ULL)) & cast<uint32_t>(127ULL)));
uint32_t d = cast<uint32_t>((shr<uint32_t>(i,cast<int64_t>(21ULL)) & cast<uint32_t>(15ULL)));
auto tmp21 = std::make_tuple(cast<uint32_t>((shr<uint32_t>(i,cast<int64_t>(16ULL)) & cast<uint32_t>(31ULL))),cast<uint32_t>((shr<uint32_t>(i,cast<int64_t>(11ULL)) & cast<uint32_t>(31ULL))));
uint32_t t = std::get<0>(tmp21);
uint32_t s = std::get<1>(tmp21);
int32_t imm11 = shr<int32_t>(cast<int32_t>(shl<uint32_t>(i,cast<int64_t>(21ULL))),cast<int64_t>(21ULL));
uint32_t imm15 = cast<uint32_t>((cast<uint32_t>((shr<uint32_t>(i,cast<int64_t>(10ULL)) & cast<uint32_t>(30720ULL))) | cast<uint32_t>((i & cast<uint32_t>(2047ULL)))));
uint32_t dmask = (cast<uint32_t>(cast<int64_t>((len(v->Data) - cast<int64_t>(1ULL)))) & ~(cast<uint32_t>(15ULL)));
auto target = [&]()->int64_t{
return cast<int64_t>((cast<int64_t>(cast<uint32_t>((cast<uint32_t>((v->PC + cast<uint32_t>(8ULL))) + cast<uint32_t>((cast<uint32_t>(imm11) * cast<uint32_t>(8ULL)))))) & cast<int64_t>(cast<uint32_t>(cast<int64_t>((len(v->Micro) - cast<int64_t>(1ULL)))))));
}
;
{
switch(op7){
case cast<uint32_t>(0ULL):{
vu_VU_loadQ(v,t,d,cast<uint32_t>(((cast<uint32_t>((cast<uint32_t>(cast<int32_t>((cast<int32_t>(v->VI[cast<uint32_t>((s & cast<uint32_t>(15ULL)))]) + imm11))) * cast<uint32_t>(16ULL)))) & dmask)));
break;}
case cast<uint32_t>(1ULL):{
vu_VU_storeQ(v,s,d,cast<uint32_t>(((cast<uint32_t>((cast<uint32_t>(cast<int32_t>((cast<int32_t>(v->VI[cast<uint32_t>((t & cast<uint32_t>(15ULL)))]) + imm11))) * cast<uint32_t>(16ULL)))) & dmask)));
break;}
case cast<uint32_t>(4ULL):{
uint32_t a = cast<uint32_t>(((cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(cast<int32_t>((cast<int32_t>(v->VI[cast<uint32_t>((s & cast<uint32_t>(15ULL)))]) + imm11))) * cast<uint32_t>(16ULL))) + cast<uint32_t>((vu_laneOf(d) * cast<uint32_t>(4ULL)))))) & ((cast<uint32_t>(cast<int64_t>((len(v->Data) - cast<int64_t>(1ULL)))) & ~(cast<uint32_t>(3ULL))))));
vu_VU_setVI(v,t,cast<uint16_t>(cast<uint32_t>((cast<uint32_t>(v->Data[a]) | shl<uint32_t>(cast<uint32_t>(v->Data[cast<uint32_t>((a + cast<uint32_t>(1ULL)))]),cast<int64_t>(8ULL))))));
break;}
case cast<uint32_t>(5ULL):{
uint32_t a = cast<uint32_t>(((cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(cast<int32_t>((cast<int32_t>(v->VI[cast<uint32_t>((s & cast<uint32_t>(15ULL)))]) + imm11))) * cast<uint32_t>(16ULL))) + cast<uint32_t>((vu_laneOf(d) * cast<uint32_t>(4ULL)))))) & ((cast<uint32_t>(cast<int64_t>((len(v->Data) - cast<int64_t>(1ULL)))) & ~(cast<uint32_t>(3ULL))))));
uint16_t val = v->VI[cast<uint32_t>((t & cast<uint32_t>(15ULL)))];
v->Data[a] = cast<uint8_t>(val);
v->Data[cast<uint32_t>((a + cast<uint32_t>(1ULL)))] = cast<uint8_t>(shr<uint16_t>(val,cast<int64_t>(8ULL)));
v->Data[cast<uint32_t>((a + cast<uint32_t>(2ULL)))] = cast<uint8_t>(0ULL);
v->Data[cast<uint32_t>((a + cast<uint32_t>(3ULL)))] = cast<uint8_t>(0ULL);
break;}
case cast<uint32_t>(8ULL):{
vu_VU_setVI(v,t,cast<uint16_t>((v->VI[cast<uint32_t>((s & cast<uint32_t>(15ULL)))] + cast<uint16_t>(imm15))));
break;}
case cast<uint32_t>(9ULL):{
vu_VU_setVI(v,t,cast<uint16_t>((v->VI[cast<uint32_t>((s & cast<uint32_t>(15ULL)))] - cast<uint16_t>(imm15))));
break;}
case cast<uint32_t>(16ULL):{
vu_VU_setVI(v,cast<uint32_t>(1ULL),vu_b2u((cast<uint32_t>((v->visClip & cast<uint32_t>(16777215ULL))) == cast<uint32_t>((i & cast<uint32_t>(16777215ULL))))));
break;}
case cast<uint32_t>(17ULL):{
v->Clip = cast<uint32_t>((i & cast<uint32_t>(16777215ULL)));
{auto&& tmp22 = v->flagPipe;
for(int64_t tmp23=0;tmp23<len(tmp22);++tmp23){
auto st=tmp23;v->flagPipe[st].clip = v->Clip;
}}
break;}
case cast<uint32_t>(18ULL):{
vu_VU_setVI(v,cast<uint32_t>(1ULL),vu_b2u((cast<uint32_t>((cast<uint32_t>((v->visClip & i)) & cast<uint32_t>(16777215ULL))) != cast<uint32_t>(0ULL))));
break;}
case cast<uint32_t>(19ULL):{
vu_VU_setVI(v,cast<uint32_t>(1ULL),vu_b2u((cast<uint32_t>(((cast<uint32_t>((v->visClip | cast<uint32_t>((i & cast<uint32_t>(16777215ULL)))))) & cast<uint32_t>(16777215ULL))) == cast<uint32_t>(16777215ULL))));
break;}
case cast<uint32_t>(20ULL):{
vu_VU_setVI(v,t,vu_b2u((cast<uint32_t>(v->visStatus) == cast<uint32_t>((cast<uint32_t>((shr<uint32_t>(i,cast<int64_t>(10ULL)) & cast<uint32_t>(2048ULL))) | cast<uint32_t>((i & cast<uint32_t>(2047ULL))))))));
break;}
case cast<uint32_t>(21ULL):{
v->Status = cast<uint16_t>((cast<uint16_t>((v->Status & cast<uint16_t>(63ULL))) | cast<uint16_t>((cast<uint16_t>(cast<uint32_t>((cast<uint32_t>((shr<uint32_t>(i,cast<int64_t>(10ULL)) & cast<uint32_t>(2048ULL))) | cast<uint32_t>((i & cast<uint32_t>(2047ULL)))))) & cast<uint16_t>(4032ULL)))));
{auto&& tmp24 = v->flagPipe;
for(int64_t tmp25=0;tmp25<len(tmp24);++tmp25){
auto st=tmp25;v->flagPipe[st].status = cast<uint16_t>((cast<uint16_t>((v->flagPipe[st].status & cast<uint16_t>(63ULL))) | cast<uint16_t>((v->Status & cast<uint16_t>(4032ULL)))));
}}
break;}
case cast<uint32_t>(22ULL):{
vu_VU_setVI(v,t,cast<uint16_t>((v->visStatus & cast<uint16_t>(cast<uint32_t>((cast<uint32_t>((shr<uint32_t>(i,cast<int64_t>(10ULL)) & cast<uint32_t>(2048ULL))) | cast<uint32_t>((i & cast<uint32_t>(2047ULL)))))))));
break;}
case cast<uint32_t>(23ULL):{
vu_VU_setVI(v,t,cast<uint16_t>((v->visStatus | cast<uint16_t>(cast<uint32_t>((cast<uint32_t>((shr<uint32_t>(i,cast<int64_t>(10ULL)) & cast<uint32_t>(2048ULL))) | cast<uint32_t>((i & cast<uint32_t>(2047ULL)))))))));
break;}
case cast<uint32_t>(24ULL):{
vu_VU_setVI(v,t,vu_b2u((v->visMac == v->VI[cast<uint32_t>((s & cast<uint32_t>(15ULL)))])));
break;}
case cast<uint32_t>(26ULL):{
vu_VU_setVI(v,t,cast<uint16_t>((v->visMac & v->VI[cast<uint32_t>((s & cast<uint32_t>(15ULL)))])));
break;}
case cast<uint32_t>(27ULL):{
vu_VU_setVI(v,t,cast<uint16_t>((v->visMac | v->VI[cast<uint32_t>((s & cast<uint32_t>(15ULL)))])));
break;}
case cast<uint32_t>(28ULL):{
vu_VU_setVI(v,t,cast<uint16_t>(cast<uint32_t>((v->visClip & cast<uint32_t>(4095ULL)))));
break;}
case cast<uint32_t>(32ULL):{
return target();
break;}
case cast<uint32_t>(33ULL):{
vu_VU_setVI(v,t,cast<uint16_t>(divi<uint32_t>((cast<uint32_t>((v->PC + cast<uint32_t>(16ULL)))),cast<uint32_t>(8ULL))));
return target();
break;}
case cast<uint32_t>(36ULL):{
return cast<int64_t>((cast<int64_t>(cast<uint32_t>((cast<uint32_t>(v->VI[cast<uint32_t>((s & cast<uint32_t>(15ULL)))]) * cast<uint32_t>(8ULL)))) & cast<int64_t>(cast<uint32_t>(cast<int64_t>((len(v->Micro) - cast<int64_t>(1ULL)))))));
break;}
case cast<uint32_t>(37ULL):{
vu_VU_setVI(v,t,cast<uint16_t>(divi<uint32_t>((cast<uint32_t>((v->PC + cast<uint32_t>(16ULL)))),cast<uint32_t>(8ULL))));
return cast<int64_t>((cast<int64_t>(cast<uint32_t>((cast<uint32_t>(v->VI[cast<uint32_t>((s & cast<uint32_t>(15ULL)))]) * cast<uint32_t>(8ULL)))) & cast<int64_t>(cast<uint32_t>(cast<int64_t>((len(v->Micro) - cast<int64_t>(1ULL)))))));
break;}
case cast<uint32_t>(40ULL):{
if ((v->VI[cast<uint32_t>((t & cast<uint32_t>(15ULL)))] == v->VI[cast<uint32_t>((s & cast<uint32_t>(15ULL)))])) {
return target();
}
break;}
case cast<uint32_t>(41ULL):{
if ((v->VI[cast<uint32_t>((t & cast<uint32_t>(15ULL)))] != v->VI[cast<uint32_t>((s & cast<uint32_t>(15ULL)))])) {
return target();
}
break;}
case cast<uint32_t>(44ULL):{
if ((cast<int16_t>(v->VI[cast<uint32_t>((s & cast<uint32_t>(15ULL)))]) < cast<int16_t>(0ULL))) {
return target();
}
break;}
case cast<uint32_t>(45ULL):{
if ((cast<int16_t>(v->VI[cast<uint32_t>((s & cast<uint32_t>(15ULL)))]) > cast<int16_t>(0ULL))) {
return target();
}
break;}
case cast<uint32_t>(46ULL):{
if ((cast<int16_t>(v->VI[cast<uint32_t>((s & cast<uint32_t>(15ULL)))]) <= cast<int16_t>(0ULL))) {
return target();
}
break;}
case cast<uint32_t>(47ULL):{
if ((cast<int16_t>(v->VI[cast<uint32_t>((s & cast<uint32_t>(15ULL)))]) >= cast<int16_t>(0ULL))) {
return target();
}
break;}
case cast<uint32_t>(64ULL):{
return vu_VU_execLowerSpecial(v,i);
break;}
}}
return cast<int64_t>(-1ULL);
}
}
// tools/cpu/vu/exec.go:653:1
int64_t vu_VU_execLowerSpecial(vu_VU* v,uint32_t i){
{
uint32_t d = cast<uint32_t>((shr<uint32_t>(i,cast<int64_t>(21ULL)) & cast<uint32_t>(15ULL)));
auto tmp26 = std::make_tuple(cast<uint32_t>((shr<uint32_t>(i,cast<int64_t>(16ULL)) & cast<uint32_t>(31ULL))),cast<uint32_t>((shr<uint32_t>(i,cast<int64_t>(11ULL)) & cast<uint32_t>(31ULL))),cast<uint32_t>((shr<uint32_t>(i,cast<int64_t>(6ULL)) & cast<uint32_t>(31ULL))));
uint32_t t = std::get<0>(tmp26);
uint32_t s = std::get<1>(tmp26);
uint32_t f = std::get<2>(tmp26);
uint32_t dmask = (cast<uint32_t>(cast<int64_t>((len(v->Data) - cast<int64_t>(1ULL)))) & ~(cast<uint32_t>(15ULL)));
{
switch(cast<uint32_t>((i & cast<uint32_t>(63ULL)))){
case cast<uint32_t>(48ULL):{
vu_VU_setVI(v,f,cast<uint16_t>((v->VI[cast<uint32_t>((s & cast<uint32_t>(15ULL)))] + v->VI[cast<uint32_t>((t & cast<uint32_t>(15ULL)))])));
return cast<int64_t>(-1ULL);
break;}
case cast<uint32_t>(49ULL):{
vu_VU_setVI(v,f,cast<uint16_t>((v->VI[cast<uint32_t>((s & cast<uint32_t>(15ULL)))] - v->VI[cast<uint32_t>((t & cast<uint32_t>(15ULL)))])));
return cast<int64_t>(-1ULL);
break;}
case cast<uint32_t>(50ULL):{
vu_VU_setVI(v,t,cast<uint16_t>(cast<int16_t>((cast<int16_t>(v->VI[cast<uint32_t>((s & cast<uint32_t>(15ULL)))]) + cast<int16_t>(shr<int32_t>(cast<int32_t>(shl<uint32_t>(i,cast<int64_t>(21ULL))),cast<int64_t>(27ULL)))))));
return cast<int64_t>(-1ULL);
break;}
case cast<uint32_t>(52ULL):{
vu_VU_setVI(v,f,cast<uint16_t>((v->VI[cast<uint32_t>((s & cast<uint32_t>(15ULL)))] & v->VI[cast<uint32_t>((t & cast<uint32_t>(15ULL)))])));
return cast<int64_t>(-1ULL);
break;}
case cast<uint32_t>(53ULL):{
vu_VU_setVI(v,f,cast<uint16_t>((v->VI[cast<uint32_t>((s & cast<uint32_t>(15ULL)))] | v->VI[cast<uint32_t>((t & cast<uint32_t>(15ULL)))])));
return cast<int64_t>(-1ULL);
break;}
}}
if ((cast<uint32_t>((i & cast<uint32_t>(60ULL))) != cast<uint32_t>(60ULL))) {
return cast<int64_t>(-1ULL);
}
uint32_t op2 = cast<uint32_t>((cast<uint32_t>((shr<uint32_t>(i,cast<int64_t>(4ULL)) & cast<uint32_t>(124ULL))) | cast<uint32_t>((i & cast<uint32_t>(3ULL)))));
uint32_t fsf = cast<uint32_t>((shr<uint32_t>(i,cast<int64_t>(21ULL)) & cast<uint32_t>(3ULL)));
uint32_t ftf = cast<uint32_t>((shr<uint32_t>(i,cast<int64_t>(23ULL)) & cast<uint32_t>(3ULL)));
{
switch(op2){
case cast<uint32_t>(48ULL):{
vu_VU_setBits(v,t,d,v->VF[s]);
break;}
case cast<uint32_t>(49ULL):{
std::array<uint32_t,4> src = v->VF[s];
vu_VU_setBits(v,t,d,std::array<uint32_t,4>{src[cast<int64_t>(1ULL)],src[cast<int64_t>(2ULL)],src[cast<int64_t>(3ULL)],src[cast<int64_t>(0ULL)]});
break;}
case cast<uint32_t>(52ULL):{
vu_VU_loadQ(v,t,d,cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(v->VI[cast<uint32_t>((s & cast<uint32_t>(15ULL)))]) * cast<uint32_t>(16ULL))) & dmask)));
vu_VU_setVI(v,s,cast<uint16_t>((v->VI[cast<uint32_t>((s & cast<uint32_t>(15ULL)))] + cast<uint16_t>(1ULL))));
break;}
case cast<uint32_t>(53ULL):{
vu_VU_storeQ(v,s,d,cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(v->VI[cast<uint32_t>((t & cast<uint32_t>(15ULL)))]) * cast<uint32_t>(16ULL))) & dmask)));
vu_VU_setVI(v,t,cast<uint16_t>((v->VI[cast<uint32_t>((t & cast<uint32_t>(15ULL)))] + cast<uint16_t>(1ULL))));
break;}
case cast<uint32_t>(54ULL):{
vu_VU_setVI(v,s,cast<uint16_t>((v->VI[cast<uint32_t>((s & cast<uint32_t>(15ULL)))] - cast<uint16_t>(1ULL))));
vu_VU_loadQ(v,t,d,cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(v->VI[cast<uint32_t>((s & cast<uint32_t>(15ULL)))]) * cast<uint32_t>(16ULL))) & dmask)));
break;}
case cast<uint32_t>(55ULL):{
vu_VU_setVI(v,t,cast<uint16_t>((v->VI[cast<uint32_t>((t & cast<uint32_t>(15ULL)))] - cast<uint16_t>(1ULL))));
vu_VU_storeQ(v,s,d,cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(v->VI[cast<uint32_t>((t & cast<uint32_t>(15ULL)))]) * cast<uint32_t>(16ULL))) & dmask)));
break;}
case cast<uint32_t>(56ULL):{
v->Q = vu_vuDiv(vu_VU_getF(v,s,fsf),vu_VU_getF(v,t,ftf));
break;}
case cast<uint32_t>(57ULL):{
v->Q = vu_sane(cast<float>(go_math_Sqrt(go_math_Abs(cast<double>(vu_VU_getF(v,t,ftf))))));
break;}
case cast<uint32_t>(58ULL):{
v->Q = vu_vuDiv(vu_VU_getF(v,s,fsf),cast<float>(go_math_Sqrt(go_math_Abs(cast<double>(vu_VU_getF(v,t,ftf))))));
break;}
case cast<uint32_t>(59ULL):{
break;}
case cast<uint32_t>(60ULL):{
vu_VU_setVI(v,t,cast<uint16_t>(v->VF[s][fsf]));
break;}
case cast<uint32_t>(61ULL):{
uint32_t bits = cast<uint32_t>(cast<int32_t>(cast<int16_t>(v->VI[cast<uint32_t>((s & cast<uint32_t>(15ULL)))])));
vu_VU_setBits(v,t,d,std::array<uint32_t,4>{bits,bits,bits,bits});
break;}
case cast<uint32_t>(62ULL):{
uint32_t a = cast<uint32_t>(((cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(v->VI[cast<uint32_t>((s & cast<uint32_t>(15ULL)))]) * cast<uint32_t>(16ULL))) + cast<uint32_t>((vu_laneOf(d) * cast<uint32_t>(4ULL)))))) & ((cast<uint32_t>(cast<int64_t>((len(v->Data) - cast<int64_t>(1ULL)))) & ~(cast<uint32_t>(3ULL))))));
vu_VU_setVI(v,t,cast<uint16_t>(cast<uint32_t>((cast<uint32_t>(v->Data[a]) | shl<uint32_t>(cast<uint32_t>(v->Data[cast<uint32_t>((a + cast<uint32_t>(1ULL)))]),cast<int64_t>(8ULL))))));
break;}
case cast<uint32_t>(63ULL):{
uint32_t a = cast<uint32_t>(((cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(v->VI[cast<uint32_t>((s & cast<uint32_t>(15ULL)))]) * cast<uint32_t>(16ULL))) + cast<uint32_t>((vu_laneOf(d) * cast<uint32_t>(4ULL)))))) & ((cast<uint32_t>(cast<int64_t>((len(v->Data) - cast<int64_t>(1ULL)))) & ~(cast<uint32_t>(3ULL))))));
uint16_t val = v->VI[cast<uint32_t>((t & cast<uint32_t>(15ULL)))];
v->Data[a] = cast<uint8_t>(val);
v->Data[cast<uint32_t>((a + cast<uint32_t>(1ULL)))] = cast<uint8_t>(shr<uint16_t>(val,cast<int64_t>(8ULL)));
v->Data[cast<uint32_t>((a + cast<uint32_t>(2ULL)))] = cast<uint8_t>(0ULL);
v->Data[cast<uint32_t>((a + cast<uint32_t>(3ULL)))] = cast<uint8_t>(0ULL);
break;}
case cast<uint32_t>(64ULL):{
vu_VU_rNext(v);
}
case cast<uint32_t>(65ULL):{
uint32_t bits = cast<uint32_t>((cast<uint32_t>(1065353216ULL) | cast<uint32_t>((v->R & cast<uint32_t>(8388607ULL)))));
vu_VU_setBits(v,t,d,std::array<uint32_t,4>{bits,bits,bits,bits});
break;}
case cast<uint32_t>(66ULL):{
v->R = cast<uint32_t>((v->VF[s][fsf] & cast<uint32_t>(8388607ULL)));
break;}
case cast<uint32_t>(67ULL):{
v->R = cast<uint32_t>(((cast<uint32_t>((v->R ^ v->VF[s][fsf]))) & cast<uint32_t>(8388607ULL)));
break;}
case cast<uint32_t>(100ULL):{
uint32_t bits = go_math_Float32bits(v->P);
vu_VU_setBits(v,t,d,std::array<uint32_t,4>{bits,bits,bits,bits});
break;}
case cast<uint32_t>(104ULL):{
vu_VU_setVI(v,t,v->Top);
break;}
case cast<uint32_t>(105ULL):{
vu_VU_setVI(v,t,v->ITop);
break;}
case cast<uint32_t>(108ULL):{
if (bool(v->XGKick)) {
v->XGKick(cast<uint32_t>((cast<uint32_t>(v->VI[cast<uint32_t>((s & cast<uint32_t>(15ULL)))]) & cast<uint32_t>(cast<int64_t>((divi<int64_t>(len(v->Data),cast<int64_t>(16ULL)) - cast<int64_t>(1ULL)))))));
}
break;}
case cast<uint32_t>(112ULL):{
v->P = vu_sane(vu_dot3(v->VF[s]));
break;}
case cast<uint32_t>(113ULL):{
v->P = vu_vuDiv(cast<float>(1ULL),vu_dot3(v->VF[s]));
break;}
case cast<uint32_t>(114ULL):{
v->P = vu_sane(cast<float>(go_math_Sqrt(cast<double>(vu_dot3(v->VF[s])))));
break;}
case cast<uint32_t>(115ULL):{
v->P = vu_vuDiv(cast<float>(1ULL),cast<float>(go_math_Sqrt(cast<double>(vu_dot3(v->VF[s])))));
break;}
case cast<uint32_t>(116ULL):{
v->P = vu_sane(cast<float>(go_math_Atan2(cast<double>(vu_VU_getF(v,s,cast<uint32_t>(1ULL))),cast<double>(vu_VU_getF(v,s,cast<uint32_t>(0ULL))))));
break;}
case cast<uint32_t>(117ULL):{
v->P = vu_sane(cast<float>(go_math_Atan2(cast<double>(vu_VU_getF(v,s,cast<uint32_t>(2ULL))),cast<double>(vu_VU_getF(v,s,cast<uint32_t>(0ULL))))));
break;}
case cast<uint32_t>(118ULL):{
v->P = vu_sane((((vu_VU_getF(v,s,cast<uint32_t>(0ULL)) + vu_VU_getF(v,s,cast<uint32_t>(1ULL))) + vu_VU_getF(v,s,cast<uint32_t>(2ULL))) + vu_VU_getF(v,s,cast<uint32_t>(3ULL))));
break;}
case cast<uint32_t>(120ULL):{
v->P = vu_sane(cast<float>(go_math_Sqrt(go_math_Abs(cast<double>(vu_VU_getF(v,s,fsf))))));
break;}
case cast<uint32_t>(121ULL):{
v->P = vu_vuDiv(cast<float>(1ULL),cast<float>(go_math_Sqrt(go_math_Abs(cast<double>(vu_VU_getF(v,s,fsf))))));
break;}
case cast<uint32_t>(122ULL):{
v->P = vu_vuDiv(cast<float>(1ULL),vu_VU_getF(v,s,fsf));
break;}
case cast<uint32_t>(123ULL):{
break;}
case cast<uint32_t>(124ULL):{
v->P = vu_sane(cast<float>(go_math_Sin(cast<double>(vu_VU_getF(v,s,fsf)))));
break;}
case cast<uint32_t>(125ULL):{
v->P = vu_sane(cast<float>(go_math_Atan(cast<double>(vu_VU_getF(v,s,fsf)))));
break;}
case cast<uint32_t>(126ULL):{
v->P = vu_sane(cast<float>(go_math_Exp(cast<double>(-cast<double>(vu_VU_getF(v,s,fsf))))));
break;}
}}
return cast<int64_t>(-1ULL);
}
}
// tools/cpu/vu/exec.go:775:1
void vu_VU_loadQ(vu_VU* v,uint32_t r,uint32_t destMask,uint32_t addr){
{
std::array<uint32_t,4> val={};
{uint32_t l = cast<uint32_t>(0ULL);for (;(l < cast<uint32_t>(4ULL));l++){
uint32_t o = cast<uint32_t>((addr + cast<uint32_t>((l * cast<uint32_t>(4ULL)))));
val[l] = cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(v->Data[o]) | shl<uint32_t>(cast<uint32_t>(v->Data[cast<uint32_t>((o + cast<uint32_t>(1ULL)))]),cast<int64_t>(8ULL)))) | shl<uint32_t>(cast<uint32_t>(v->Data[cast<uint32_t>((o + cast<uint32_t>(2ULL)))]),cast<int64_t>(16ULL)))) | shl<uint32_t>(cast<uint32_t>(v->Data[cast<uint32_t>((o + cast<uint32_t>(3ULL)))]),cast<int64_t>(24ULL))));
}
}vu_VU_setBits(v,r,destMask,val);
}
}
// tools/cpu/vu/exec.go:785:1
void vu_VU_storeQ(vu_VU* v,uint32_t r,uint32_t destMask,uint32_t addr){
{
{uint32_t l = cast<uint32_t>(0ULL);for (;(l < cast<uint32_t>(4ULL));l++){
if ((cast<uint32_t>((destMask & (shr<uint32_t>(cast<uint32_t>(8ULL),l)))) == cast<uint32_t>(0ULL))) {
continue;
}
uint32_t o = cast<uint32_t>((addr + cast<uint32_t>((l * cast<uint32_t>(4ULL)))));
uint32_t bits = v->VF[r][l];
if ((bool(v->OnMaxStore) && (cast<uint32_t>((bits & cast<uint32_t>(2147483647ULL))) == cast<uint32_t>(2139095039ULL)))) {
v->OnMaxStore(v->PC,divi<uint32_t>(addr,cast<uint32_t>(16ULL)),bits);
}
v->Data[o] = cast<uint8_t>(bits);
v->Data[cast<uint32_t>((o + cast<uint32_t>(1ULL)))] = cast<uint8_t>(shr<uint32_t>(bits,cast<int64_t>(8ULL)));
v->Data[cast<uint32_t>((o + cast<uint32_t>(2ULL)))] = cast<uint8_t>(shr<uint32_t>(bits,cast<int64_t>(16ULL)));
v->Data[cast<uint32_t>((o + cast<uint32_t>(3ULL)))] = cast<uint8_t>(shr<uint32_t>(bits,cast<int64_t>(24ULL)));
}
}}
}
// tools/cpu/vu/exec.go:803:1
uint32_t vu_laneOf(uint32_t destMask){
{
{
switch(destMask){
case cast<uint32_t>(8ULL):{
return cast<uint32_t>(0ULL);
break;}
case cast<uint32_t>(4ULL):{
return cast<uint32_t>(1ULL);
break;}
case cast<uint32_t>(2ULL):{
return cast<uint32_t>(2ULL);
break;}
default:{
return cast<uint32_t>(3ULL);
break;}
}}
}
}
// tools/cpu/vu/exec.go:818:1
float vu_vuDiv(float a,float b){
{
if ((b == cast<float>(0ULL))) {
if ((go_math_Signbit(cast<double>(a)) != go_math_Signbit(cast<double>(b)))) {
return cast<float>(-go_math_MaxFloat32);
}
return go_math_MaxFloat32;
}
return vu_sane((a / b));
}
}
// tools/cpu/vu/exec.go:828:1
float vu_dot3(std::array<uint32_t,4> r){
{
float x = go_math_Float32frombits(r[cast<int64_t>(0ULL)]);
float y = go_math_Float32frombits(r[cast<int64_t>(1ULL)]);
float z = go_math_Float32frombits(r[cast<int64_t>(2ULL)]);
return (((x * x) + (y * y)) + (z * z));
}
}
// tools/cpu/vu/exec.go:838:1
void vu_VU_rNext(vu_VU* v){
{
uint32_t bit = cast<uint32_t>(((shr<uint32_t>(v->R,cast<int64_t>(22ULL))) ^ (shr<uint32_t>(v->R,cast<int64_t>(17ULL)))));
v->R = cast<uint32_t>(((cast<uint32_t>((shl<uint32_t>(v->R,cast<int64_t>(1ULL)) | cast<uint32_t>((bit & cast<uint32_t>(1ULL)))))) & cast<uint32_t>(8388607ULL)));
}
}
// tools/cpu/vu/exec.go:844:1
uint16_t vu_b2u(bool b){
{
if (b) {
return cast<uint16_t>(1ULL);
}
return cast<uint16_t>(0ULL);
}
}
// tools/cpu/vu/macro.go:15:1
void vu_VU_Macro(vu_VU* v,uint32_t w){
{
uint32_t op = cast<uint32_t>((w & cast<uint32_t>(63ULL)));
if ((op < cast<uint32_t>(48ULL))) {
vu_upperResult res={};
vu_VU_execUpper(v,w,(&res));
vu_VU_commitUpper(v,(&res));
return ;
}
if ((op < cast<uint32_t>(60ULL))) {
{
switch(op){
case cast<uint32_t>(56ULL):{
vu_VU_Run(v,cast<uint32_t>((cast<uint32_t>((shr<uint32_t>(w,cast<int64_t>(6ULL)) & cast<uint32_t>(32767ULL))) * cast<uint32_t>(8ULL))),cast<int64_t>(1048576ULL));
break;}
case cast<uint32_t>(57ULL):{
vu_VU_Run(v,cast<uint32_t>((cast<uint32_t>(v->CMSAR0) * cast<uint32_t>(8ULL))),cast<int64_t>(1048576ULL));
break;}
default:{
vu_VU_execLowerSpecial(v,w);
break;}
}}
return ;
}
uint32_t op2 = cast<uint32_t>((cast<uint32_t>((shr<uint32_t>(w,cast<int64_t>(4ULL)) & cast<uint32_t>(124ULL))) | cast<uint32_t>((w & cast<uint32_t>(3ULL)))));
if ((op2 <= cast<uint32_t>(47ULL))) {
vu_upperResult res={};
vu_VU_execUpper(v,w,(&res));
vu_VU_commitUpper(v,(&res));
return ;
}
vu_VU_execLowerSpecial(v,w);
}
}
// tools/cpu/vu/macro.go:50:1
uint32_t vu_VU_ReadVF(vu_VU* v,uint32_t reg,uint32_t field){
{
return v->VF[cast<uint32_t>((reg & cast<uint32_t>(31ULL)))][cast<uint32_t>((field & cast<uint32_t>(3ULL)))];
}
}
// tools/cpu/vu/macro.go:55:1
void vu_VU_WriteVF(vu_VU* v,uint32_t reg,uint32_t field,uint32_t val){
{
if ((cast<uint32_t>((reg & cast<uint32_t>(31ULL))) != cast<uint32_t>(0ULL))) {
v->VF[cast<uint32_t>((reg & cast<uint32_t>(31ULL)))][cast<uint32_t>((field & cast<uint32_t>(3ULL)))] = val;
}
}
}
// tools/cpu/vu/macro.go:78:1
uint32_t vu_VU_ReadCtrl(vu_VU* v,uint32_t reg){
{
reg &= cast<uint32_t>(31ULL);
if ((reg < cast<uint32_t>(16ULL))) {
return cast<uint32_t>(v->VI[reg]);
}
{
switch(reg){
case cast<uint32_t>(16ULL):{
return cast<uint32_t>(v->Status);
break;}
case cast<uint32_t>(17ULL):{
return cast<uint32_t>(v->Mac);
break;}
case cast<uint32_t>(18ULL):{
return v->Clip;
break;}
case cast<uint32_t>(20ULL):{
return v->R;
break;}
case cast<uint32_t>(21ULL):{
return vu_float32bits(v->I);
break;}
case cast<uint32_t>(22ULL):{
return vu_float32bits(v->Q);
break;}
case cast<uint32_t>(26ULL):{
return divi<uint32_t>(v->PC,cast<uint32_t>(8ULL));
break;}
case cast<uint32_t>(27ULL):{
return cast<uint32_t>(v->CMSAR0);
break;}
case cast<uint32_t>(29ULL):{
return cast<uint32_t>(0ULL);
break;}
}}
return cast<uint32_t>(0ULL);
}
}
// tools/cpu/vu/macro.go:109:1
void vu_VU_WriteCtrl(vu_VU* v,uint32_t reg,uint32_t val){
{
reg &= cast<uint32_t>(31ULL);
if ((reg == cast<uint32_t>(0ULL))) {
return ;
}
if ((reg < cast<uint32_t>(16ULL))) {
v->VI[reg] = cast<uint16_t>(val);
return ;
}
{
switch(reg){
case cast<uint32_t>(16ULL):{
v->Status = cast<uint16_t>(val);
break;}
case cast<uint32_t>(17ULL):{
v->Mac = cast<uint16_t>(val);
break;}
case cast<uint32_t>(18ULL):{
v->Clip = cast<uint32_t>((val & cast<uint32_t>(16777215ULL)));
break;}
case cast<uint32_t>(20ULL):{
v->R = cast<uint32_t>((val & cast<uint32_t>(8388607ULL)));
break;}
case cast<uint32_t>(21ULL):{
v->I = vu_float32frombits(val);
break;}
case cast<uint32_t>(22ULL):{
v->Q = vu_float32frombits(val);
break;}
case cast<uint32_t>(27ULL):{
v->CMSAR0 = cast<uint16_t>(val);
break;}
case cast<uint32_t>(28ULL):{
break;}
case cast<uint32_t>(31ULL):{
if (bool(v->StartVU1)) {
v->StartVU1(val);
}
break;}
}}
}
}
// tools/cpu/vu/state.go:48:1
vu_State vu_VU_Snapshot(vu_VU* v){
{
vu_State s = vu_State{v->VF,v->VI,v->ACC,cast<float>(v->Q),cast<float>(v->P),cast<float>(v->I),v->R,v->PC,v->Mac,v->Status,v->Clip,{},v->visMac,v->visStatus,v->visClip,v->Top,v->ITop,v->CMSAR0,v->Steps};
{auto&& tmp27 = v->flagPipe;
for(int64_t tmp28=0;tmp28<len(tmp27);++tmp28){
auto i=tmp28;auto f=tmp27[tmp28];s.FlagPipe[i] = vu_FlagVals{f.mac,f.status,f.clip};
}}
return s;
}
}
// tools/cpu/vu/state.go:65:1
void vu_VU_Restore(vu_VU* v,vu_State s){
{
auto tmp29 = std::make_tuple(s.VF,s.VI,s.ACC);
v->VF = std::get<0>(tmp29);
v->VI = std::get<1>(tmp29);
v->ACC = std::get<2>(tmp29);
auto tmp30 = std::make_tuple(s.Q,s.P,s.I,s.R);
v->Q = std::get<0>(tmp30);
v->P = std::get<1>(tmp30);
v->I = std::get<2>(tmp30);
v->R = std::get<3>(tmp30);
v->PC = s.PC;
auto tmp31 = std::make_tuple(s.Mac,s.Status,s.Clip);
v->Mac = std::get<0>(tmp31);
v->Status = std::get<1>(tmp31);
v->Clip = std::get<2>(tmp31);
auto tmp32 = std::make_tuple(s.VisMac,s.VisStatus,s.VisClip);
v->visMac = std::get<0>(tmp32);
v->visStatus = std::get<1>(tmp32);
v->visClip = std::get<2>(tmp32);
auto tmp33 = std::make_tuple(s.Top,s.ITop,s.CMSAR0,s.Steps);
v->Top = std::get<0>(tmp33);
v->ITop = std::get<1>(tmp33);
v->CMSAR0 = std::get<2>(tmp33);
v->Steps = std::get<3>(tmp33);
{auto&& tmp34 = s.FlagPipe;
for(int64_t tmp35=0;tmp35<len(tmp34);++tmp35){
auto i=tmp35;auto f=tmp34[tmp35];v->flagPipe[i] = vu_flagVals{f.Mac,f.Status,f.Clip};
}}
}
}
// tools/cpu/vu/vu.go:12:1
float vu_float32frombits(uint32_t b){
{
return go_math_Float32frombits(b);
}
}
// tools/cpu/vu/vu.go:13:1
uint32_t vu_float32bits(float f){
{
return go_math_Float32bits(f);
}
}
// tools/lib/iso9660/iso9660.go:58:1
std::string iso9660_Geometry_String(iso9660_Geometry g){
{
return go_fmt_Sprintf(std::string("%d-byte sectors, data at +%d",28),g.SectorSize,g.DataOffset);
}
}
// tools/lib/iso9660/iso9660.go:81:1
iso9660_Geometry iso9660_Source_Geometry(iso9660_Source* s){
{
return s->geom;
}
}
// tools/lib/iso9660/iso9660.go:162:1
std::tuple<Slice<uint8_t>,Error> iso9660_Volume_ReadBlock(iso9660_Volume* v,int64_t n){
{
return localSource_ReadBlock(v->src,n);
}
}
// tools/lib/iso9660/iso9660.go:173:1
std::string iso9660_Entry_String(iso9660_Entry e){
{
if (e.IsDir) {
return go_fmt_Sprintf(std::string("%-32s  <dir>   (lba %d)",23),e.Path,e.Block);
}
return go_fmt_Sprintf(std::string("%-32s  %10d  (lba %d)",21),e.Path,e.Size,e.Block);
}
}
// tools/lib/iso9660/iso9660.go:197:1
std::tuple<iso9660_Volume*,Error> iso9660_OpenVolume(LocalSource* src){
{
return iso9660_OpenVolumeAt(src,cast<int64_t>(16ULL));
}
}
// tools/lib/iso9660/iso9660.go:205:1
std::tuple<iso9660_Volume*,Error> iso9660_OpenVolumeAt(LocalSource* src,int64_t pvd){
{
auto tmp1 = localSource_ReadBlock(src,pvd);
Slice<uint8_t> b = std::get<0>(tmp1);
Error err = std::get<1>(tmp1);
if (bool(err)) {
return {{},go_fmt_Errorf(std::string("iso9660: reading PVD: %w",24),err)};
}
if (((b[cast<int64_t>(0ULL)] != cast<uint8_t>(1ULL)) || (cast<std::string>(sub(b,cast<int64_t>(1ULL),cast<int64_t>(6ULL))) != std::string("CD001",5)))) {
return {{},go_fmt_Errorf(std::string("iso9660: no Primary Volume Descriptor at LBA %d",47),pvd)};
}
iso9660_Volume* v = arenaNew(iso9660_Volume{src,go_strings_TrimRight(cast<std::string>(sub(b,cast<int64_t>(8ULL),cast<int64_t>(40ULL))),std::string(" \000",2)),go_strings_TrimRight(cast<std::string>(sub(b,cast<int64_t>(40ULL),cast<int64_t>(72ULL))),std::string(" \000",2)),cast<int64_t>(iso9660_le32(rrBorrow(b,cast<int64_t>(80ULL),len(b)))),{},{}});
Slice<uint8_t> root = sub(b,cast<int64_t>(156ULL),cast<int64_t>(190ULL));
v->rootLBA = cast<int64_t>(iso9660_le32(rrBorrow(root,cast<int64_t>(2ULL),len(root))));
v->rootSize = cast<int64_t>(iso9660_le32(rrBorrow(root,cast<int64_t>(10ULL),len(root))));
return {v,{}};
}
}
// tools/lib/iso9660/iso9660.go:227:1
uint32_t iso9660_le32(Slice<uint8_t> b){
{
return le_Uint32(rrBorrow(b,0,cast<int64_t>(4ULL)));
}
}
// tools/lib/iso9660/iso9660.go:231:1
std::tuple<Slice<iso9660_Entry>,Error> iso9660_Volume_dirEntries(iso9660_Volume* v,int64_t lba,int64_t size,std::string dirPath){
{
Slice<iso9660_Entry> out={};
int64_t remaining = size;
{int64_t sect = cast<int64_t>(0ULL);for (;(remaining > cast<int64_t>(0ULL));sect++){
auto tmp2 = localSource_ReadBlock(v->src,cast<int64_t>((lba + sect)));
Slice<uint8_t> b = std::get<0>(tmp2);
Error err = std::get<1>(tmp2);
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
return {{},go_fmt_Errorf(std::string("iso9660: directory record overruns block at lba %d",50),cast<int64_t>((lba + sect)))};
}
Slice<uint8_t> rec = sub(b,p,cast<int64_t>((p + recLen)));
p += recLen;
int64_t idLen = cast<int64_t>(rec[cast<int64_t>(32ULL)]);
Slice<uint8_t> id = sub(rec,cast<int64_t>(33ULL),cast<int64_t>((cast<int64_t>(33ULL) + idLen)));
if (((idLen == cast<int64_t>(1ULL)) && (((id[cast<int64_t>(0ULL)] == cast<uint8_t>(0ULL)) || (id[cast<int64_t>(0ULL)] == cast<uint8_t>(1ULL)))))) {
continue;
}
iso9660_Entry e = iso9660_Entry{cast<std::string>(id),{},(cast<uint8_t>((rec[cast<int64_t>(25ULL)] & cast<uint8_t>(2ULL))) != cast<uint8_t>(0ULL)),cast<int64_t>(iso9660_le32(rrBorrow(rec,cast<int64_t>(10ULL),len(rec)))),cast<int64_t>(iso9660_le32(rrBorrow(rec,cast<int64_t>(2ULL),len(rec))))};
if ((dirPath == std::string("",0))) {
e.Path = e.Name;
}
else {
e.Path = ((dirPath + std::string("/",1)) + e.Name);
}
out = append(out,Slice<iso9660_Entry>{e});
}
}}
}return {out,{}};
}
}
// tools/lib/iso9660/iso9660.go:280:1
bool iso9660_nameEqual(std::string entry,std::string want){
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
// tools/lib/iso9660/iso9660.go:292:1
Slice<std::string> iso9660_splitPath(std::string p){
{
Slice<std::string> parts={};
{auto&& tmp3 = go_strings_Split(go_strings_Trim(p,std::string("/",1)),std::string("/",1));
for(int64_t tmp4=0;tmp4<len(tmp3);++tmp4){
auto c=tmp3[tmp4];if ((c != std::string("",0))) {
parts = append(parts,Slice<std::string>{c});
}
}}
return parts;
}
}
// tools/lib/iso9660/iso9660.go:309:1
std::tuple<iso9660_Entry,Error> iso9660_Volume_Resolve(iso9660_Volume* v,std::string path){
{
path = iso9660_normalisePath(path);
{
auto tmp5 = iso9660_parseLbnPath(path);
int64_t lbn = std::get<0>(tmp5);
int64_t size = std::get<1>(tmp5);
bool ok = std::get<2>(tmp5);
if (ok) {
return {iso9660_Entry{path,path,{},size,lbn},{}};
}
}
iso9660_Entry cur = iso9660_Entry{v->Name,std::string("",0),true,v->rootSize,v->rootLBA};
{auto&& tmp6 = iso9660_splitPath(path);
for(int64_t tmp7=0;tmp7<len(tmp6);++tmp7){
auto want=tmp6[tmp7];if ((!cur.IsDir)) {
return {iso9660_Entry{},go_fmt_Errorf(std::string("iso9660: %q is not a directory",30),cur.Path)};
}
auto tmp8 = iso9660_Volume_dirEntries(v,cur.Block,cur.Size,cur.Path);
Slice<iso9660_Entry> entries = std::get<0>(tmp8);
Error err = std::get<1>(tmp8);
if (bool(err)) {
return {iso9660_Entry{},err};
}
bool found = false;
{auto&& tmp9 = entries;
for(int64_t tmp10=0;tmp10<len(tmp9);++tmp10){
auto e=tmp9[tmp10];if (iso9660_nameEqual(e.Name,want)) {
auto tmp11 = std::make_tuple(e,true);
cur = std::get<0>(tmp11);
found = std::get<1>(tmp11);
break;
}
}}
if ((!found)) {
return {iso9660_Entry{},go_fmt_Errorf(std::string("iso9660: %q not found",21),path)};
}
}}
return {cur,{}};
}
}
// tools/lib/iso9660/iso9660.go:338:1
std::string iso9660_normalisePath(std::string p){
{
{
int64_t i = go_strings_IndexByte(p,cast<uint8_t>(58ULL));
if (((i >= cast<int64_t>(0ULL)) && (!go_strings_ContainsAny(sub(p,0,i),std::string("/\\",2))))) {
p = sub(p,cast<int64_t>((i + cast<int64_t>(1ULL))),len(p));
}
}
return go_strings_ReplaceAll(p,std::string("\\",1),std::string("/",1));
}
}
// tools/lib/iso9660/iso9660.go:347:1
std::tuple<int64_t,int64_t,bool> iso9660_parseLbnPath(std::string path){
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
auto tmp12 = go_strconv_ParseInt(go_strings_TrimPrefix(sub(rest,0,i),std::string("0x",2)),cast<int64_t>(16ULL),cast<int64_t>(64ULL));
int64_t l = std::get<0>(tmp12);
Error err1 = std::get<1>(tmp12);
auto tmp13 = go_strconv_ParseInt(go_strings_TrimPrefix(sub(rest,cast<int64_t>((i + cast<int64_t>(5ULL))),len(rest)),std::string("0x",2)),cast<int64_t>(16ULL),cast<int64_t>(64ULL));
int64_t s = std::get<0>(tmp13);
Error err2 = std::get<1>(tmp13);
if ((bool(err1) || bool(err2))) {
return {cast<int64_t>(0ULL),cast<int64_t>(0ULL),false};
}
return {cast<int64_t>(l),cast<int64_t>(s),true};
}
}
// tools/lib/iso9660/iso9660.go:366:1
std::tuple<Slice<iso9660_Entry>,Error> iso9660_Volume_ReadDir(iso9660_Volume* v,std::string path){
{
auto tmp14 = iso9660_Volume_Resolve(v,path);
iso9660_Entry e = std::get<0>(tmp14);
Error err = std::get<1>(tmp14);
if (bool(err)) {
return {{},err};
}
if ((!e.IsDir)) {
return {{},go_fmt_Errorf(std::string("iso9660: %q is not a directory",30),path)};
}
return iso9660_Volume_dirEntries(v,e.Block,e.Size,e.Path);
}
}
// tools/lib/iso9660/iso9660.go:378:1
Error iso9660_Volume_Walk(iso9660_Volume* v,std::function<Error(iso9660_Entry)> fn){
{
return iso9660_Volume_walk(v,v->rootLBA,v->rootSize,std::string("",0),fn);
}
}
// tools/lib/iso9660/iso9660.go:382:1
Error iso9660_Volume_walk(iso9660_Volume* v,int64_t lba,int64_t size,std::string dirPath,std::function<Error(iso9660_Entry)> fn){
{
auto tmp15 = iso9660_Volume_dirEntries(v,lba,size,dirPath);
Slice<iso9660_Entry> entries = std::get<0>(tmp15);
Error err = std::get<1>(tmp15);
if (bool(err)) {
return err;
}
{auto&& tmp16 = entries;
for(int64_t tmp17=0;tmp17<len(tmp16);++tmp17){
auto e=tmp16[tmp17];{
Error err = fn(e);
if (bool(err)) {
return err;
}
}
if (e.IsDir) {
{
Error err = iso9660_Volume_walk(v,e.Block,e.Size,e.Path,fn);
if (bool(err)) {
return err;
}
}
}
}}
return {};
}
}
// tools/lib/iso9660/iso9660.go:402:1
std::tuple<int64_t,Error> iso9660_Volume_ReadFileAt(iso9660_Volume* v,iso9660_Entry e,int64_t off,Slice<uint8_t> p){
{
if ((off >= e.Size)) {
return {cast<int64_t>(0ULL),{}};
}
if ((cast<int64_t>((off + len(p))) > e.Size)) {
p = sub(p,0,cast<int64_t>((e.Size - off)));
}
int64_t got = cast<int64_t>(0ULL);
{;for (;(got < len(p));){
auto tmp18 = localSource_ReadBlock(v->src,cast<int64_t>((e.Block + divi<int64_t>((cast<int64_t>((off + got))),cast<int64_t>(2048ULL)))));
Slice<uint8_t> b = std::get<0>(tmp18);
Error err = std::get<1>(tmp18);
if (bool(err)) {
return {got,err};
}
got += gcopy(sub(p,got,len(p)),sub(b,modi<int64_t>((cast<int64_t>((off + got))),cast<int64_t>(2048ULL)),len(b)));
}
}return {got,{}};
}
}
// tools/lib/iso9660/iso9660.go:421:1
std::tuple<Slice<uint8_t>,Error> iso9660_Volume_ReadFile(iso9660_Volume* v,std::string path){
{
auto tmp19 = iso9660_Volume_Resolve(v,path);
iso9660_Entry e = std::get<0>(tmp19);
Error err = std::get<1>(tmp19);
if (bool(err)) {
return {{},err};
}
if (e.IsDir) {
return {{},go_fmt_Errorf(std::string("iso9660: %q is a directory",26),path)};
}
Slice<uint8_t> out = Slice<uint8_t>::make(e.Size);
{
auto tmp20 = iso9660_Volume_ReadFileAt(v,e,cast<int64_t>(0ULL),out);
Error err = std::get<1>(tmp20);
if (bool(err)) {
return {{},err};
}
}
return {out,{}};
}
}
// tools/lib/iso9660/iso9660.go:439:1
std::tuple<iso9660_Entry,bool> iso9660_Volume_FileAt(iso9660_Volume* v,int64_t lba){
{
iso9660_Entry hit={};
bool found = false;
iso9660_Volume_Walk(v,[&](iso9660_Entry e)->Error{
if (e.IsDir) {
return {};
}
int64_t blocks = divi<int64_t>((cast<int64_t>((cast<int64_t>((e.Size + cast<int64_t>(2048ULL))) - cast<int64_t>(1ULL)))),cast<int64_t>(2048ULL));
if (((lba >= e.Block) && (lba < cast<int64_t>((e.Block + blocks))))) {
auto tmp21 = std::make_tuple(e,true);
hit = std::get<0>(tmp21);
found = std::get<1>(tmp21);
}
return {};
}
);
return {hit,found};
}
}
// tools/lib/iso9660/iso9660.go:458:1
iso9660_byteReaderAt iso9660_newByteReaderAt(Slice<uint8_t> b){
{
return cast<iso9660_byteReaderAt>(b);
}
}
// tools/lib/iso9660/iso9660.go:460:1
std::tuple<int64_t,Error> iso9660_byteReaderAt_ReadAt(iso9660_byteReaderAt b,Slice<uint8_t> p,int64_t off){
{
if (((off < cast<int64_t>(0ULL)) || (off >= cast<int64_t>(len(b))))) {
return {cast<int64_t>(0ULL),go_io_EOF};
}
int64_t n = gcopy(p,sub(b,off,len(b)));
if ((n < len(p))) {
return {n,go_io_EOF};
}
return {n,{}};
}
}
// tools/platform/ps2/dmac.go:124:1
std::tuple<int64_t,uint32_t,bool> ps2_dmacChanReg(uint32_t a){
int64_t ch{};
uint32_t reg{};
bool ok{};
{
if (((a < cast<uint32_t>(268468224ULL)) || (a >= cast<uint32_t>(268492800ULL)))) {
return {cast<int64_t>(0ULL),cast<uint32_t>(0ULL),false};
}
{auto&& tmp1 = ps2_dmacChanBase;
for(int64_t tmp2=0;tmp2<len(tmp1);++tmp2){
auto i=tmp2;auto base=tmp1[tmp2];if (((a >= base) && (a < cast<uint32_t>((base + cast<uint32_t>(256ULL)))))) {
return {i,cast<uint32_t>((a - base)),true};
}
}}
return {cast<int64_t>(0ULL),cast<uint32_t>(0ULL),false};
}
}
// tools/platform/ps2/dmac.go:138:1
void ps2_Machine_drainVIF1(ps2_Machine* m){
{
if ((!m->vif1Pending)) {
return ;
}
m->vif1Pending = false;
ps2_Machine_dmacStart(m,cast<int64_t>(1ULL));
}
}
// tools/platform/ps2/dmac.go:147:1
std::tuple<uint32_t,bool> ps2_Machine_dmacRead(ps2_Machine* m,uint32_t a){
{
ps2_Machine_drainVIF1(m);
{
auto tmp3 = ps2_dmacChanReg(a);
int64_t ch = std::get<0>(tmp3);
uint32_t reg = std::get<1>(tmp3);
bool ok = std::get<2>(tmp3);
if (ok) {
ps2_dmacChan* c = (&m->dmac[ch]);
{
switch(reg){
case cast<uint32_t>(0ULL):{
return {c->chcr,true};
break;}
case cast<uint32_t>(16ULL):{
return {c->madr,true};
break;}
case cast<uint32_t>(32ULL):{
return {c->qwc,true};
break;}
case cast<uint32_t>(48ULL):{
return {c->tadr,true};
break;}
case cast<uint32_t>(64ULL):{
return {c->asr0,true};
break;}
case cast<uint32_t>(80ULL):{
return {c->asr1,true};
break;}
case cast<uint32_t>(128ULL):{
return {c->sadr,true};
break;}
}}
return {cast<uint32_t>(0ULL),true};
}
}
{
switch(a){
case cast<uint32_t>(268492800ULL):{
return {m->dCtrl,true};
break;}
case cast<uint32_t>(268492816ULL):{
return {cast<uint32_t>((m->dmacStat | shl<uint32_t>(m->dmacMask,cast<int64_t>(16ULL)))),true};
break;}
case cast<uint32_t>(268492832ULL):{
return {m->dPcr,true};
break;}
case cast<uint32_t>(268492848ULL):{
return {m->dSqwc,true};
break;}
case cast<uint32_t>(268492864ULL):{
return {m->dRbsr,true};
break;}
case cast<uint32_t>(268492880ULL):{
return {m->dRbor,true};
break;}
case cast<uint32_t>(268498208ULL):{
return {m->dEnable,true};
break;}
}}
return {cast<uint32_t>(0ULL),false};
}
}
// tools/platform/ps2/dmac.go:204:1
bool ps2_Machine_dmacWrite(ps2_Machine* m,uint32_t a,uint32_t v){
{
if (m->vif1Pending) {
{
auto tmp4 = ps2_dmacChanReg(a);
int64_t ch = std::get<0>(tmp4);
uint32_t reg = std::get<1>(tmp4);
bool ok = std::get<2>(tmp4);
if (((((ok && (ch == cast<int64_t>(2ULL))) && (reg == cast<uint32_t>(0ULL))) && (cast<uint32_t>((v & cast<uint32_t>(256ULL))) != cast<uint32_t>(0ULL))) && (cast<uint32_t>((v & cast<uint32_t>(12ULL))) == cast<uint32_t>(4ULL)))) {
m->dmac[ch].chcr = v;
ps2_Machine_dmacStart(m,cast<int64_t>(2ULL));
ps2_Machine_drainVIF1(m);
return true;
}
}
ps2_Machine_drainVIF1(m);
}
{
auto tmp5 = ps2_dmacChanReg(a);
int64_t ch = std::get<0>(tmp5);
uint32_t reg = std::get<1>(tmp5);
bool ok = std::get<2>(tmp5);
if (ok) {
ps2_dmacChan* c = (&m->dmac[ch]);
{
switch(reg){
case cast<uint32_t>(0ULL):{
c->chcr = v;
if ((ps2_kickLog && (ch <= cast<int64_t>(2ULL)))) {
go_fmt_Printf(std::string("  kick ch%d CHCR=%08X TADR=%08X MADR=%08X QWC=%d pc=%08X\012",57),ch,v,c->tadr,c->madr,c->qwc,cast<uint32_t>(m->CPU->PC));
}
if ((cast<uint32_t>((v & cast<uint32_t>(256ULL))) != cast<uint32_t>(0ULL))) {
if (((ch == cast<int64_t>(1ULL)) && (cast<uint32_t>((v & cast<uint32_t>(12ULL))) == cast<uint32_t>(4ULL)))) {
m->vif1Pending = true;
return true;
}
ps2_Machine_dmacStart(m,ch);
}
break;}
case cast<uint32_t>(16ULL):{
c->madr = v;
break;}
case cast<uint32_t>(32ULL):{
c->qwc = cast<uint32_t>((v & cast<uint32_t>(65535ULL)));
break;}
case cast<uint32_t>(48ULL):{
c->tadr = v;
if ((ps2_kickLog && (ch <= cast<int64_t>(2ULL)))) {
go_fmt_Printf(std::string("  kick ch%d TADR=%08X pc=%08X ra=%08X\012",38),ch,v,cast<uint32_t>(m->CPU->PC),cast<uint32_t>(m->CPU->R[cast<int64_t>(31ULL)].Lo));
}
break;}
case cast<uint32_t>(64ULL):{
c->asr0 = v;
break;}
case cast<uint32_t>(80ULL):{
c->asr1 = v;
break;}
case cast<uint32_t>(128ULL):{
c->sadr = v;
break;}
}}
return true;
}
}
{
switch(a){
case cast<uint32_t>(268492800ULL):{
m->dCtrl = v;
break;}
case cast<uint32_t>(268492816ULL):{
m->dmacStat &= ~(cast<uint32_t>((v & cast<uint32_t>(65535ULL))));
m->dmacMask ^= cast<uint32_t>(((shr<uint32_t>(v,cast<int64_t>(16ULL))) & cast<uint32_t>(65535ULL)));
ps2_Machine_dmacRetrigger(m);
break;}
case cast<uint32_t>(268492832ULL):{
m->dPcr = v;
break;}
case cast<uint32_t>(268492848ULL):{
m->dSqwc = v;
break;}
case cast<uint32_t>(268492864ULL):{
m->dRbsr = v;
break;}
case cast<uint32_t>(268492880ULL):{
m->dRbor = v;
break;}
case cast<uint32_t>(268498320ULL):{
m->dEnable = v;
break;}
default:{
return false;
break;}
}}
return true;
}
}
// tools/platform/ps2/dmac.go:284:1
void ps2_Machine_dmacStart(ps2_Machine* m,int64_t ch){
{
ps2_Machine_profDrainEnter(m);
auto tmp6=defer([&](){ps2_Machine_profDrainExit(m);});
ps2_dmacChan* c = (&m->dmac[ch]);
{
switch(ch){
case cast<int64_t>(2ULL):{
ps2_Machine_gifStart(m,c);
break;}
case cast<int64_t>(0ULL):{
ps2_Machine_vifStart(m,cast<int64_t>(0ULL),c);
break;}
case cast<int64_t>(1ULL):{
ps2_Machine_vifStart(m,cast<int64_t>(1ULL),c);
break;}
case cast<int64_t>(5ULL):case cast<int64_t>(6ULL):{
break;}
case cast<int64_t>(9ULL):{
{
switch(cast<uint32_t>((c->chcr & cast<uint32_t>(12ULL)))){
case cast<uint32_t>(4ULL):{
ps2_Machine_dmacSPRChainTo(m,c);
break;}
case cast<uint32_t>(8ULL):{
ps2_Machine_dmacSPRInterleave(m,c,false);
break;}
default:{
ps2_Machine_dmacSPR(m,c,false);
break;}
}}
break;}
case cast<int64_t>(8ULL):{
if ((cast<uint32_t>((c->chcr & cast<uint32_t>(12ULL))) == cast<uint32_t>(8ULL))) {
ps2_Machine_dmacSPRInterleave(m,c,true);
}
else {
ps2_Machine_dmacSPR(m,c,true);
}
break;}
default:{
ps2_Machine_note(m,std::string("EE DMA: channel %d started (mode 0x%X, %d qwords at 0x%08X), and nothing models it",82),ch,shr<uint32_t>((cast<uint32_t>((c->chcr & cast<uint32_t>(12ULL)))),cast<int64_t>(2ULL)),c->qwc,c->madr);
break;}
}}
ps2_Machine_dmacComplete(m,ch);
}
}
// tools/platform/ps2/dmac.go:363:1
void ps2_Machine_dmacComplete(ps2_Machine* m,int64_t ch){
{
m->dmac[ch].chcr &= ~(cast<uint32_t>(256ULL));
m->dmacStat |= shl<uint32_t>(cast<uint32_t>(1ULL),cast<uint64_t>(ch));
if ((cast<uint32_t>((m->dmacMask & (shl<uint32_t>(cast<uint32_t>(1ULL),cast<uint64_t>(ch))))) != cast<uint32_t>(0ULL))) {
ps2_Machine_queueDmacIRQ(m,ch);
}
}
}
// tools/platform/ps2/dmac.go:386:1
void ps2_Machine_queueDmacIRQ(ps2_Machine* m,int64_t ch){
{
{auto&& tmp7 = m->dmacHandlers;
for(int64_t tmp8=0;tmp8<len(tmp7);++tmp8){
auto h=tmp7[tmp8];if ((cast<int64_t>(h.cause) != ch)) {
continue;
}
{auto&& tmp9 = m->dmacIRQPending;
for(int64_t tmp10=0;tmp10<len(tmp9);++tmp10){
auto q=tmp9[tmp10];if ((q == ch)) {
return ;
}
}}
m->dmacIRQPending = append(m->dmacIRQPending,Slice<int64_t>{ch});
return ;
}}
}
}
// tools/platform/ps2/dmac.go:405:1
void ps2_Machine_dmacRetrigger(ps2_Machine* m){
{
uint32_t live = cast<uint32_t>((m->dmacStat & m->dmacMask));
if ((live == cast<uint32_t>(0ULL))) {
return ;
}
{int64_t ch = cast<int64_t>(0ULL);for (;(ch < cast<int64_t>(16ULL));ch++){
if ((cast<uint32_t>((live & (shl<uint32_t>(cast<uint32_t>(1ULL),cast<uint64_t>(ch))))) != cast<uint32_t>(0ULL))) {
ps2_Machine_queueDmacIRQ(m,ch);
}
}
}}
}
// tools/platform/ps2/dmac.go:419:1
void ps2_Machine_deliverDmacIRQs(ps2_Machine* m){
{
Slice<int64_t> pending = m->dmacIRQPending;
m->dmacIRQPending = {};
{auto&& tmp11 = pending;
for(int64_t tmp12=0;tmp12<len(tmp11);++tmp12){
auto ch=tmp11[tmp12];{auto&& tmp13 = m->dmacHandlers;
for(int64_t tmp14=0;tmp14<len(tmp13);++tmp14){
auto h=tmp13[tmp14];if ((cast<int64_t>(h.cause) == ch)) {
ps2_Machine_callGuest(m,h.addr,Slice<uint32_t>{cast<uint32_t>(ch),h.arg});
}
}}
}}
ps2_Machine_preemptIfOutranked(m);
}
}
// tools/platform/ps2/dmac.go:444:1
void ps2_Machine_dmacSPR(ps2_Machine* m,ps2_dmacChan* c,bool fromSPR){
{
uint32_t n = cast<uint32_t>((c->qwc * cast<uint32_t>(16ULL)));
uint32_t madr = cast<uint32_t>((c->madr & cast<uint32_t>(268435455ULL)));
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < n);i++){
uint32_t s = cast<uint32_t>(((cast<uint32_t>((c->sadr + i))) & cast<uint32_t>(16383ULL)));
if (fromSPR) {
ps2_Machine_Write(m,cast<uint32_t>((madr + i)),m->spram[s]);
}
else {
m->hookMuted = true;
m->spram[s] = ps2_Machine_Read(m,cast<uint32_t>((madr + i)));
m->hookMuted = false;
}
}
}c->madr += n;
c->sadr = cast<uint32_t>(((cast<uint32_t>((c->sadr + n))) & cast<uint32_t>(16383ULL)));
c->qwc = cast<uint32_t>(0ULL);
}
}
// tools/platform/ps2/dmac.go:489:1
void ps2_Machine_dmacSPRInterleave(ps2_Machine* m,ps2_dmacChan* c,bool fromSPR){
{
uint32_t tqwc = cast<uint32_t>((shr<uint32_t>(m->dSqwc,cast<int64_t>(16ULL)) & cast<uint32_t>(255ULL)));
uint32_t sqwc = cast<uint32_t>((m->dSqwc & cast<uint32_t>(255ULL)));
if ((tqwc == cast<uint32_t>(0ULL))) {
ps2_Machine_dmacSPR(m,c,fromSPR);
return ;
}
uint32_t madr = cast<uint32_t>((c->madr & cast<uint32_t>(268435455ULL)));
uint32_t sadr = c->sadr;
uint32_t left = c->qwc;
{;for (;(left > cast<uint32_t>(0ULL));){
uint32_t n = tqwc;
if ((n > left)) {
n = left;
}
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < cast<uint32_t>((n * cast<uint32_t>(16ULL))));i++){
uint32_t s = cast<uint32_t>(((cast<uint32_t>((sadr + i))) & cast<uint32_t>(16383ULL)));
if (fromSPR) {
ps2_Machine_Write(m,cast<uint32_t>((madr + i)),m->spram[s]);
}
else {
m->hookMuted = true;
m->spram[s] = ps2_Machine_Read(m,cast<uint32_t>((madr + i)));
m->hookMuted = false;
}
}
}sadr += cast<uint32_t>((n * cast<uint32_t>(16ULL)));
madr += cast<uint32_t>(((cast<uint32_t>((n + sqwc))) * cast<uint32_t>(16ULL)));
left -= n;
}
}}
}
// tools/platform/ps2/dmac.go:529:1
void ps2_Machine_dmacSPRChainTo(ps2_Machine* m,ps2_dmacChan* c){
{
uint32_t sadr = cast<uint32_t>((c->sadr & cast<uint32_t>(16383ULL)));
ps2_Machine_dmacSourceChain(m,cast<int64_t>(9ULL),c,[&](Slice<uint8_t> b)->void{
{auto&& tmp15 = b;
for(int64_t tmp16=0;tmp16<len(tmp15);++tmp16){
auto x=tmp15[tmp16];m->spram[sadr] = x;
sadr = cast<uint32_t>(((cast<uint32_t>((sadr + cast<uint32_t>(1ULL)))) & cast<uint32_t>(16383ULL)));
}}
}
,true);
c->sadr = sadr;
}
}
// tools/platform/ps2/dmac.go:574:1
void ps2_Machine_dmacSourceChain(ps2_Machine* m,int64_t ch,ps2_dmacChan* c,std::function<void(Slice<uint8_t>)> feed,bool wholeTag){
{
bool tte = (cast<uint32_t>((c->chcr & cast<uint32_t>(64ULL))) != cast<uint32_t>(0ULL));
constexpr int64_t maxLinks=65536ULL;
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(65536ULL));i++){
Slice<uint8_t> tag = ps2_Machine_dmaBytes(m,c->tadr,cast<uint32_t>(1ULL));
uint64_t lo = ps2_le64(tag);
uint32_t qwc = cast<uint32_t>(cast<uint64_t>((lo & cast<uint64_t>(65535ULL))));
uint32_t id = cast<uint32_t>((cast<uint32_t>(shr<uint64_t>(lo,cast<int64_t>(28ULL))) & cast<uint32_t>(7ULL)));
bool irq = (cast<uint64_t>((lo & cast<uint64_t>(2147483648ULL))) != cast<uint64_t>(0ULL));
uint32_t addr = cast<uint32_t>(((cast<uint32_t>(shr<uint64_t>(lo,cast<int64_t>(32ULL))) & ~(cast<uint32_t>(15ULL))) & cast<uint32_t>(2147483647ULL)));
if ((cast<uint64_t>((lo & cast<uint64_t>(9223372036854775808ULL))) != cast<uint64_t>(0ULL))) {
addr |= cast<uint32_t>(2147483648ULL);
}
if (((ps2_chainLogN > cast<int64_t>(0ULL)) && (((ch == cast<int64_t>(1ULL)) || (ch == cast<int64_t>(2ULL)))))) {
ps2_chainLogN--;
go_fmt_Printf(std::string("  chain ch%d tag@%08X id=%d qwc=%d addr=%08X\012",45),ch,c->tadr,id,qwc,addr);
}
if (tte) {
m->feedMadr = cast<uint32_t>((c->tadr + cast<uint32_t>(8ULL)));
if (wholeTag) {
m->feedMadr = c->tadr;
feed(sub(tag,cast<int64_t>(0ULL),cast<int64_t>(16ULL)));
}
else {
feed(sub(tag,cast<int64_t>(8ULL),cast<int64_t>(16ULL)));
}
}
uint32_t data = cast<uint32_t>((cast<uint32_t>((c->tadr & cast<uint32_t>(2147483648ULL))) | (cast<uint32_t>((cast<uint32_t>((c->tadr & cast<uint32_t>(2147483647ULL))) + cast<uint32_t>(16ULL))))));
bool end = false;
{
switch(id){
case cast<uint32_t>(0ULL):{
data = addr;
end = true;
break;}
case cast<uint32_t>(1ULL):{
c->tadr = cast<uint32_t>((data + cast<uint32_t>((qwc * cast<uint32_t>(16ULL)))));
break;}
case cast<uint32_t>(2ULL):{
c->tadr = addr;
break;}
case cast<uint32_t>(3ULL):case cast<uint32_t>(4ULL):{
data = addr;
c->tadr += cast<uint32_t>(16ULL);
break;}
case cast<uint32_t>(5ULL):{
c->asr1 = c->asr0;
c->asr0 = cast<uint32_t>((cast<uint32_t>((data & cast<uint32_t>(2147483648ULL))) | (cast<uint32_t>((cast<uint32_t>((data & cast<uint32_t>(2147483647ULL))) + cast<uint32_t>((qwc * cast<uint32_t>(16ULL))))))));
c->tadr = addr;
break;}
case cast<uint32_t>(6ULL):{
if (((c->asr0 == cast<uint32_t>(0ULL)) && (c->asr1 == cast<uint32_t>(0ULL)))) {
end = true;
break;
}
c->tadr = c->asr0;
auto tmp17 = std::make_tuple(c->asr1,cast<uint32_t>(0ULL));
c->asr0 = std::get<0>(tmp17);
c->asr1 = std::get<1>(tmp17);
break;}
case cast<uint32_t>(7ULL):{
end = true;
break;}
}}
if ((qwc > cast<uint32_t>(0ULL))) {
m->feedMadr = data;
feed(ps2_Machine_dmaBytes(m,data,qwc));
}
if (end) {
return ;
}
if ((irq && (cast<uint32_t>((c->chcr & cast<uint32_t>(128ULL))) != cast<uint32_t>(0ULL)))) {
return ;
}
if ((i == cast<int64_t>(65535ULL))) {
ps2_Machine_note(m,std::string("EE DMA: channel %d's chain ran %d links without ending \342\200\224 truncated at TADR 0x%08X",83),ch,cast<int64_t>(65536ULL),c->tadr);
}
}
}}
}
// tools/platform/ps2/dmac.go:658:1
Slice<uint8_t> ps2_Machine_dmaBytes(ps2_Machine* m,uint32_t madr,uint32_t qwc){
{
uint32_t n = cast<uint32_t>((qwc * cast<uint32_t>(16ULL)));
Slice<uint8_t> out = Slice<uint8_t>::make(n);
bool spr = (cast<uint32_t>((madr & cast<uint32_t>(2147483648ULL))) != cast<uint32_t>(0ULL));
uint32_t addr = cast<uint32_t>((madr & cast<uint32_t>(268435455ULL)));
m->hookMuted = true;
auto tmp18=defer([&](){[&]()->void{
m->hookMuted = false;
}
();});
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < n);i++){
if (spr) {
out[i] = m->spram[cast<uint32_t>(((cast<uint32_t>((addr + i))) & cast<uint32_t>(16383ULL)))];
}
else {
out[i] = ps2_Machine_Read(m,cast<uint32_t>((addr + i)));
}
}
}return out;
}
}
// tools/platform/ps2/eetimer.go:48:1
uint64_t ps2_eeTimer_rate(ps2_eeTimer* t){
{
{
switch(cast<uint32_t>((t->mode & cast<uint32_t>(3ULL)))){
case cast<uint32_t>(0ULL):{
return cast<uint64_t>(2460180ULL);
break;}
case cast<uint32_t>(1ULL):{
return cast<uint64_t>(153761ULL);
break;}
case cast<uint32_t>(2ULL):{
return cast<uint64_t>(9610ULL);
break;}
default:{
return cast<uint64_t>(263ULL);
break;}
}}
}
}
// tools/platform/ps2/eetimer.go:62:1
uint32_t ps2_eeTimer_countAt(ps2_eeTimer* t,uint64_t steps){
{
if ((cast<uint32_t>((t->mode & cast<uint32_t>(128ULL))) == cast<uint32_t>(0ULL))) {
return cast<uint32_t>((t->base & cast<uint32_t>(65535ULL)));
}
uint64_t elapsed = divi<uint64_t>(cast<uint64_t>(((cast<uint64_t>((steps - t->baseSteps))) * ps2_eeTimer_rate(t))),cast<uint64_t>(1000000ULL));
uint64_t c = cast<uint64_t>((cast<uint64_t>(t->base) + elapsed));
if (((cast<uint32_t>((t->mode & cast<uint32_t>(64ULL))) != cast<uint32_t>(0ULL)) && (t->comp > cast<uint32_t>(0ULL)))) {
return cast<uint32_t>(modi<uint64_t>(c,cast<uint64_t>(t->comp)));
}
return cast<uint32_t>(cast<uint64_t>((c & cast<uint64_t>(65535ULL))));
}
}
// tools/platform/ps2/eetimer.go:76:1
void ps2_eeTimer_rebase(ps2_eeTimer* t,uint64_t steps){
{
t->base = ps2_eeTimer_countAt(t,steps);
t->baseSteps = steps;
}
}
// tools/platform/ps2/eetimer.go:82:1
std::tuple<int64_t,uint32_t> ps2_eeTimerAt(uint32_t p){
int64_t n{};
uint32_t reg{};
{
if (((p < cast<uint32_t>(268435456ULL)) || (p >= cast<uint32_t>(268441648ULL)))) {
return {cast<int64_t>(-1ULL),cast<uint32_t>(0ULL)};
}
uint32_t off = cast<uint32_t>((p - cast<uint32_t>(268435456ULL)));
if (((cast<uint32_t>((off & cast<uint32_t>(2047ULL))) >= cast<uint32_t>(64ULL)) || (cast<uint32_t>((off & cast<uint32_t>(15ULL))) != cast<uint32_t>(0ULL)))) {
return {cast<int64_t>(-1ULL),cast<uint32_t>(0ULL)};
}
return {cast<int64_t>(shr<uint32_t>(off,cast<int64_t>(11ULL))),cast<uint32_t>((off & cast<uint32_t>(63ULL)))};
}
}
// tools/platform/ps2/eetimer.go:93:1
std::tuple<uint32_t,bool> ps2_Machine_eeTimerRead(ps2_Machine* m,uint32_t p){
{
auto tmp20 = ps2_eeTimerAt(p);
int64_t n = std::get<0>(tmp20);
uint32_t reg = std::get<1>(tmp20);
if ((n < cast<int64_t>(0ULL))) {
return {cast<uint32_t>(0ULL),false};
}
ps2_eeTimer* t = (&m->eeTimers[n]);
{
switch(reg){
case cast<uint32_t>(0ULL):{
return {ps2_eeTimer_countAt(t,m->steps),true};
break;}
case cast<uint32_t>(16ULL):{
return {t->mode,true};
break;}
case cast<uint32_t>(32ULL):{
return {t->comp,true};
break;}
case cast<uint32_t>(48ULL):{
return {t->hold,true};
break;}
}}
return {cast<uint32_t>(0ULL),false};
}
}
// tools/platform/ps2/eetimer.go:112:1
bool ps2_Machine_eeTimerWrite(ps2_Machine* m,uint32_t p,uint32_t v){
{
auto tmp21 = ps2_eeTimerAt(p);
int64_t n = std::get<0>(tmp21);
uint32_t reg = std::get<1>(tmp21);
if ((n < cast<int64_t>(0ULL))) {
return false;
}
ps2_eeTimer* t = (&m->eeTimers[n]);
{
switch(reg){
case cast<uint32_t>(0ULL):{
t->base = cast<uint32_t>((v & cast<uint32_t>(65535ULL)));
t->baseSteps = m->steps;
break;}
case cast<uint32_t>(16ULL):{
ps2_eeTimer_rebase(t,m->steps);
t->mode = (v & ~(cast<uint32_t>(3072ULL)));
if ((cast<uint32_t>((v & cast<uint32_t>(768ULL))) != cast<uint32_t>(0ULL))) {
ps2_Machine_note(m,std::string("EE timer %d MODE 0x%X enables interrupts \342\200\224 not delivered",58),n,v);
}
break;}
case cast<uint32_t>(32ULL):{
t->comp = cast<uint32_t>((v & cast<uint32_t>(65535ULL)));
break;}
case cast<uint32_t>(48ULL):{
t->hold = cast<uint32_t>((v & cast<uint32_t>(65535ULL)));
break;}
}}
return true;
}
}
// tools/platform/ps2/elf.go:104:1
std::tuple<std::string,uint32_t,bool> ps2_Executable_Lookup(ps2_Executable* e,uint32_t addr){
std::string name{};
uint32_t off{};
bool ok{};
{
int64_t i = cast<int64_t>((go_sort_Search(len(e->byAddr),[&](int64_t i)->bool{
return (e->byAddr[i].Addr > addr);
}
) - cast<int64_t>(1ULL)));
if ((i < cast<int64_t>(0ULL))) {
return {std::string("",0),cast<uint32_t>(0ULL),false};
}
ps2_Symbol s = e->byAddr[i];
uint32_t end = cast<uint32_t>((s.Addr + s.Size));
if ((s.Size == cast<uint32_t>(0ULL))) {
if ((cast<int64_t>((i + cast<int64_t>(1ULL))) < len(e->byAddr))) {
end = e->byAddr[cast<int64_t>((i + cast<int64_t>(1ULL)))].Addr;
}
else {
end = cast<uint32_t>((s.Addr + cast<uint32_t>(4ULL)));
}
}
if ((addr >= end)) {
return {std::string("",0),cast<uint32_t>(0ULL),false};
}
return {s.Name,cast<uint32_t>((addr - s.Addr)),true};
}
}
// tools/platform/ps2/elf.go:127:1
std::string ps2_Executable_Describe(ps2_Executable* e){
{
std::string s = go_fmt_Sprintf(std::string("entry:    0x%08X\012",17),e->Entry);
{auto&& tmp22 = e->Segments;
for(int64_t tmp23=0;tmp23<len(tmp22);++tmp23){
auto i=tmp23;auto seg=tmp22[tmp23];uint32_t bss = cast<uint32_t>((seg.MemSz - cast<uint32_t>(len(seg.Data))));
s += go_fmt_Sprintf(std::string("segment %d: 0x%08X..0x%08X  %d bytes in file, %d bytes bss\012",59),i,seg.VAddr,cast<uint32_t>((cast<uint32_t>((seg.VAddr + seg.MemSz)) - cast<uint32_t>(1ULL))),len(seg.Data),bss);
}}
int64_t funcs = cast<int64_t>(0ULL);
{auto&& tmp24 = e->Symbols;
for(int64_t tmp25=0;tmp25<len(tmp24);++tmp25){
auto sym=tmp24[tmp25];if (sym.Func) {
funcs++;
}
}}
s += go_fmt_Sprintf(std::string("symbols:  %d (%d functions)\012",28),len(e->Symbols),funcs);
return s;
}
}
// tools/platform/ps2/elf.go:147:1
std::tuple<uint32_t,Slice<uint8_t>> ps2_Executable_Flat(ps2_Executable* e){
uint32_t base{};
Slice<uint8_t> mem{};
{
auto tmp26 = std::make_tuple(cast<uint32_t>(4294967295ULL),cast<uint32_t>(0ULL));
uint32_t lo = std::get<0>(tmp26);
uint32_t hi = std::get<1>(tmp26);
{auto&& tmp27 = e->Segments;
for(int64_t tmp28=0;tmp28<len(tmp27);++tmp28){
auto s=tmp27[tmp28];if ((s.VAddr < lo)) {
lo = s.VAddr;
}
if ((cast<uint32_t>((s.VAddr + s.MemSz)) > hi)) {
hi = cast<uint32_t>((s.VAddr + s.MemSz));
}
}}
mem = Slice<uint8_t>::make(cast<uint32_t>((hi - lo)));
{auto&& tmp29 = e->Segments;
for(int64_t tmp30=0;tmp30<len(tmp29);++tmp30){
auto s=tmp29[tmp30];gcopy(sub(mem,cast<uint32_t>((s.VAddr - lo)),len(mem)),s.Data);
}}
return {lo,mem};
}
}
// tools/platform/ps2/gif.go:55:1
void ps2_Machine_gifStart(ps2_Machine* m,ps2_dmacChan* c){
{
{
switch(shr<uint32_t>((cast<uint32_t>((c->chcr & cast<uint32_t>(12ULL)))),cast<int64_t>(2ULL))){
case cast<uint32_t>(1ULL):{
Slice<uint8_t> all={};
ps2_Machine_dmacSourceChain(m,cast<int64_t>(2ULL),c,[&](Slice<uint8_t> b)->void{
if ((len(b) == cast<int64_t>(8ULL))) {
ps2_Machine_note(m,std::string("GS: the GIF channel forwarded a chain tag (TTE) \342\200\224 unhandled",61));
return ;
}
all = append(all,b);
}
,false);
if ((len(all) > cast<int64_t>(0ULL))) {
ps2_Machine_gifPacket(m,all);
}
break;}
default:{
if ((c->qwc == cast<uint32_t>(0ULL))) {
return ;
}
ps2_Machine_gifPacket(m,ps2_Machine_dmaBytes(m,c->madr,c->qwc));
break;}
}}
}
}
// tools/platform/ps2/gif.go:89:1
void ps2_Machine_gifStream(ps2_Machine* m,Slice<uint8_t> data){
{
ps2_GS* gs = ps2_Machine_ensureGS(m);
if ((gs->path2SkipRemain > cast<int64_t>(0ULL))) {
int64_t n = gs->path2SkipRemain;
if ((n > len(data))) {
n = len(data);
}
gs->path2SkipRemain -= n;
data = sub(data,n,len(data));
}
if ((gs->path2ImageRemain > cast<int64_t>(0ULL))) {
int64_t n = gs->path2ImageRemain;
if ((n > len(data))) {
n = len(data);
}
ps2_GS_imageData(gs,sub(data,0,n));
gs->path2ImageRemain -= n;
data = sub(data,n,len(data));
}
if ((len(gs->path2Carry) > cast<int64_t>(0ULL))) {
data = append(gs->path2Carry,data);
gs->path2Carry = {};
}
int64_t pos = cast<int64_t>(0ULL);
{;for (;(cast<int64_t>((len(data) - pos)) >= cast<int64_t>(16ULL));){
uint64_t lo = ps2_le64(rrBorrow(data,pos,len(data)));
int64_t nloop = cast<int64_t>(cast<uint64_t>((lo & cast<uint64_t>(32767ULL))));
uint64_t flg = cast<uint64_t>(((shr<uint64_t>(lo,cast<int64_t>(58ULL))) & cast<uint64_t>(3ULL)));
int64_t nreg = cast<int64_t>(cast<uint64_t>(((shr<uint64_t>(lo,cast<int64_t>(60ULL))) & cast<uint64_t>(15ULL))));
if ((nreg == cast<int64_t>(0ULL))) {
nreg = cast<int64_t>(16ULL);
}
int64_t need={};
{
switch(flg){
case cast<uint64_t>(0ULL):{
need = cast<int64_t>((cast<int64_t>(16ULL) + cast<int64_t>((cast<int64_t>((nloop * nreg)) * cast<int64_t>(16ULL)))));
break;}
case cast<uint64_t>(1ULL):{
int64_t n = cast<int64_t>((nloop * nreg));
need = cast<int64_t>((cast<int64_t>(16ULL) + cast<int64_t>(((cast<int64_t>((n + cast<int64_t>((n & cast<int64_t>(1ULL)))))) * cast<int64_t>(8ULL)))));
break;}
default:{
need = cast<int64_t>((cast<int64_t>(16ULL) + cast<int64_t>((nloop * cast<int64_t>(16ULL)))));
break;}
}}
if ((cast<int64_t>((len(data) - pos)) >= need)) {
ps2_Machine_gifPacket(m,sub(data,pos,cast<int64_t>((pos + need))));
pos += need;
continue;
}
{
switch(flg){
case cast<uint64_t>(2ULL):{
ps2_Machine_gifPacket(m,sub(data,pos,len(data)));
gs->path2ImageRemain = cast<int64_t>((need - (cast<int64_t>((len(data) - pos)))));
break;}
case cast<uint64_t>(3ULL):{
gs->path2SkipRemain = cast<int64_t>((need - (cast<int64_t>((len(data) - pos)))));
break;}
default:{
gs->path2Carry = append(Slice<uint8_t>{},sub(data,pos,len(data)));
break;}
}}
return ;
}
}{
int64_t rest = cast<int64_t>((len(data) - pos));
if ((rest > cast<int64_t>(0ULL))) {
gs->path2Carry = append(Slice<uint8_t>{},sub(data,pos,len(data)));
(void)(rest);
}
}
}
}
// tools/platform/ps2/gif.go:157:1
void ps2_Machine_gifPacket(ps2_Machine* m,Slice<uint8_t> data){
{
ps2_GS* gs = ps2_Machine_ensureGS(m);
gs->curPacket = data;
auto tmp31=defer([&](){[&]()->void{
gs->curPacket = {};
}
();});
int64_t pos = cast<int64_t>(0ULL);
{;for (;(cast<int64_t>((pos + cast<int64_t>(16ULL))) <= len(data));){
uint64_t lo = ps2_le64(rrBorrow(data,pos,len(data)));
uint64_t hi = ps2_le64(rrBorrow(data,cast<int64_t>((pos + cast<int64_t>(8ULL))),len(data)));
pos += cast<int64_t>(16ULL);
uint32_t nloop = cast<uint32_t>(cast<uint64_t>((lo & cast<uint64_t>(32767ULL))));
bool eop = (cast<uint64_t>((lo & cast<uint64_t>(32768ULL))) != cast<uint64_t>(0ULL));
bool pre = (cast<uint64_t>((lo & cast<uint64_t>(70368744177664ULL))) != cast<uint64_t>(0ULL));
uint32_t prim = cast<uint32_t>(cast<uint64_t>(((shr<uint64_t>(lo,cast<int64_t>(47ULL))) & cast<uint64_t>(2047ULL))));
uint64_t flg = cast<uint64_t>(((shr<uint64_t>(lo,cast<int64_t>(58ULL))) & cast<uint64_t>(3ULL)));
int64_t nreg = cast<int64_t>(cast<uint64_t>(((shr<uint64_t>(lo,cast<int64_t>(60ULL))) & cast<uint64_t>(15ULL))));
if ((nreg == cast<int64_t>(0ULL))) {
nreg = cast<int64_t>(16ULL);
}
uint64_t regs = hi;
if (pre) {
ps2_GS_write(gs,cast<uint8_t>(0ULL),cast<uint64_t>(prim));
}
{
switch(flg){
case cast<uint64_t>(0ULL):{
{uint32_t n = cast<uint32_t>(0ULL);for (;(n < nloop);n++){
{int64_t r = cast<int64_t>(0ULL);for (;(r < nreg);r++){
if ((cast<int64_t>((pos + cast<int64_t>(16ULL))) > len(data))) {
return ;
}
uint64_t desc = cast<uint64_t>(((shr<uint64_t>(regs,(cast<uint64_t>((cast<uint64_t>(4ULL) * cast<uint64_t>(r)))))) & cast<uint64_t>(15ULL)));
uint64_t dlo = ps2_le64(rrBorrow(data,pos,len(data)));
uint64_t dhi = ps2_le64(rrBorrow(data,cast<int64_t>((pos + cast<int64_t>(8ULL))),len(data)));
pos += cast<int64_t>(16ULL);
ps2_Machine_gifPacked(m,gs,desc,dlo,dhi);
}
}}
}break;}
case cast<uint64_t>(1ULL):{
uint32_t total = cast<uint32_t>((nloop * cast<uint32_t>(nreg)));
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < total);i++){
if ((cast<int64_t>((pos + cast<int64_t>(8ULL))) > len(data))) {
return ;
}
uint64_t desc = cast<uint64_t>(((shr<uint64_t>(regs,(cast<uint64_t>((cast<uint64_t>(4ULL) * cast<uint64_t>(modi<uint32_t>(i,cast<uint32_t>(nreg)))))))) & cast<uint64_t>(15ULL)));
uint64_t val = ps2_le64(rrBorrow(data,pos,len(data)));
pos += cast<int64_t>(8ULL);
if ((desc != cast<uint64_t>(15ULL))) {
ps2_GS_write(gs,cast<uint8_t>(desc),val);
}
}
}if ((cast<uint32_t>((total & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL))) {
pos += cast<int64_t>(8ULL);
}
break;}
case cast<uint64_t>(2ULL):{
int64_t n = cast<int64_t>((cast<int64_t>(nloop) * cast<int64_t>(16ULL)));
if ((cast<int64_t>((pos + n)) > len(data))) {
n = cast<int64_t>((len(data) - pos));
}
ps2_GS_imageData(gs,sub(data,pos,cast<int64_t>((pos + n))));
pos += n;
break;}
case cast<uint64_t>(3ULL):{
pos += cast<int64_t>((cast<int64_t>(nloop) * cast<int64_t>(16ULL)));
break;}
}}
if (eop) {
(void)(eop);
}
}
}}
}
// tools/platform/ps2/gif.go:236:1
void ps2_Machine_gifPacked(ps2_Machine* m,ps2_GS* gs,uint64_t desc,uint64_t lo,uint64_t hi){
{
{
switch(desc){
case cast<uint64_t>(14ULL):{
uint8_t reg = cast<uint8_t>(cast<uint64_t>((hi & cast<uint64_t>(255ULL))));
ps2_GS_write(gs,reg,lo);
break;}
case cast<uint64_t>(15ULL):{
break;}
default:{
ps2_GS_writePacked(gs,cast<uint8_t>(desc),lo,hi);
break;}
}}
}
}
// tools/platform/ps2/gif.go:253:1
int64_t ps2_gifPacketLen(Slice<uint8_t> data){
{
int64_t pos = cast<int64_t>(0ULL);
{;for (;(cast<int64_t>((pos + cast<int64_t>(16ULL))) <= len(data));){
uint64_t lo = ps2_le64(rrBorrow(data,pos,len(data)));
int64_t nloop = cast<int64_t>(cast<uint64_t>((lo & cast<uint64_t>(32767ULL))));
bool eop = (cast<uint64_t>((lo & cast<uint64_t>(32768ULL))) != cast<uint64_t>(0ULL));
uint64_t flg = cast<uint64_t>((shr<uint64_t>(lo,cast<int64_t>(58ULL)) & cast<uint64_t>(3ULL)));
int64_t nreg = cast<int64_t>((cast<int64_t>(shr<uint64_t>(lo,cast<int64_t>(60ULL))) & cast<int64_t>(15ULL)));
if ((nreg == cast<int64_t>(0ULL))) {
nreg = cast<int64_t>(16ULL);
}
pos += cast<int64_t>(16ULL);
{
switch(flg){
case cast<uint64_t>(0ULL):{
pos += cast<int64_t>((cast<int64_t>((nloop * nreg)) * cast<int64_t>(16ULL)));
break;}
case cast<uint64_t>(1ULL):{
int64_t n = cast<int64_t>((nloop * nreg));
pos += cast<int64_t>(((cast<int64_t>((n + cast<int64_t>((n & cast<int64_t>(1ULL)))))) * cast<int64_t>(8ULL)));
break;}
default:{
pos += cast<int64_t>((nloop * cast<int64_t>(16ULL)));
break;}
}}
if (eop) {
break;
}
}
}if ((pos > len(data))) {
pos = len(data);
}
return pos;
}
}
// tools/platform/ps2/gif.go:285:1
uint64_t ps2_le64(Slice<uint8_t> b){
{
return cast<uint64_t>((cast<uint64_t>((cast<uint64_t>((cast<uint64_t>((cast<uint64_t>((cast<uint64_t>((cast<uint64_t>((cast<uint64_t>(b[cast<int64_t>(0ULL)]) | shl<uint64_t>(cast<uint64_t>(b[cast<int64_t>(1ULL)]),cast<int64_t>(8ULL)))) | shl<uint64_t>(cast<uint64_t>(b[cast<int64_t>(2ULL)]),cast<int64_t>(16ULL)))) | shl<uint64_t>(cast<uint64_t>(b[cast<int64_t>(3ULL)]),cast<int64_t>(24ULL)))) | shl<uint64_t>(cast<uint64_t>(b[cast<int64_t>(4ULL)]),cast<int64_t>(32ULL)))) | shl<uint64_t>(cast<uint64_t>(b[cast<int64_t>(5ULL)]),cast<int64_t>(40ULL)))) | shl<uint64_t>(cast<uint64_t>(b[cast<int64_t>(6ULL)]),cast<int64_t>(48ULL)))) | shl<uint64_t>(cast<uint64_t>(b[cast<int64_t>(7ULL)]),cast<int64_t>(56ULL))));
}
}
// tools/platform/ps2/gs.go:170:1
ps2_GS* ps2_Machine_ensureGS(ps2_Machine* m){
{
if ((!m->gs)) {
m->gs = arenaNew(ps2_GS{Slice<uint8_t>::make(cast<int64_t>(4194304ULL)),{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},m});
}
return m->gs;
}
}
// tools/platform/ps2/gs.go:180:1
void ps2_GS_write(ps2_GS* gs,uint8_t reg,uint64_t val){
{
bool logThis = ((bool(gs->m) && (gs->m->GSRegLogN > cast<int64_t>(0ULL))) && (reg == gs->m->GSRegLog));
if ((((!logThis) && bool(gs->m)) && bool(gs->m->GSRegLogs))) {
{
int64_t n = get(gs->m->GSRegLogs,reg);
if ((n > cast<int64_t>(0ULL))) {
gs->m->GSRegLogs[reg] = cast<int64_t>((n - cast<int64_t>(1ULL)));
logThis = true;
}
}
}
if (logThis) {
if (((gs->m->GSRegLogN > cast<int64_t>(0ULL)) && (reg == gs->m->GSRegLog))) {
gs->m->GSRegLogN--;
}
std::string srcEE = std::string("",0);
{
ps2_vif* v = gs->m->vifs[cast<int64_t>(1ULL)];
if ((bool(v) && go_strings_HasPrefix(gs->src,std::string("path2",5)))) {
srcEE = ps2_sprintf(std::string(" (payload EE 0x%08X)",20),v->payloadAddr);
}
}
go_fmt_Printf(std::string("  GS reg 0x%02X <- %016X at prim %d from %s%s\012",46),reg,val,gs->prims,gs->src,srcEE);
if (((gs->m->GSRegDumpPacket && (val == gs->m->GSRegDumpVal)) && bool(gs->curPacket))) {
gs->m->GSRegDumpPacket = false;
int64_t n = divi<int64_t>(len(gs->curPacket),cast<int64_t>(16ULL));
if ((n > cast<int64_t>(48ULL))) {
n = cast<int64_t>(48ULL);
}
go_fmt_Printf(std::string("  the packet carrying that write (%d qw, first %d):\012",52),divi<int64_t>(len(gs->curPacket),cast<int64_t>(16ULL)),n);
{int64_t i = cast<int64_t>(0ULL);for (;(i < n);i++){
go_fmt_Printf(std::string("    qw %2d: %016X %016X\012",24),i,ps2_le64(rrBorrow(gs->curPacket,cast<int64_t>((cast<int64_t>((i * cast<int64_t>(16ULL))) + cast<int64_t>(8ULL))),len(gs->curPacket))),ps2_le64(rrBorrow(gs->curPacket,cast<int64_t>((i * cast<int64_t>(16ULL))),len(gs->curPacket))));
}
}}
}
if ((bool(gs->m) && (gs->m->GSPktTraceN > cast<int64_t>(0ULL)))) {
if ((((gs->m->GSPktTraceOn == cast<int64_t>(0ULL)) && (reg == gs->m->GSPktArmReg)) && (val == gs->m->GSPktArmVal))) {
gs->m->GSPktTraceOn = gs->m->GSPktTraceN;
gs->m->GSPktTraceN = cast<int64_t>(0ULL);
go_fmt_Printf(std::string("  pkttrace armed at prim %d from %s\012",36),gs->prims,gs->src);
}
}
if ((bool(gs->m) && (gs->m->GSPktTraceOn > cast<int64_t>(0ULL)))) {
gs->m->GSPktTraceOn--;
go_fmt_Printf(std::string("  pkt reg 0x%02X <- %016X (prim %d)\012",36),reg,val,gs->prims);
}
if ((cast<int64_t>(reg) < cast<int64_t>(128ULL))) {
gs->reg[reg] = val;
}
{
switch(reg){
case cast<uint8_t>(0ULL):{
gs->vqN = cast<int64_t>(0ULL);
break;}
case cast<uint8_t>(1ULL):{
gs->q = cast<uint32_t>(shr<uint64_t>(val,cast<int64_t>(32ULL)));
break;}
case cast<uint8_t>(6ULL):case cast<uint8_t>(7ULL):{
ps2_GS_clutLoad(gs,ps2_decodeTEX0(val));
break;}
case cast<uint8_t>(22ULL):case cast<uint8_t>(23ULL):{
constexpr uint64_t tex2Mask=18446743936336658432ULL;
uint8_t t0 = cast<uint8_t>((cast<uint8_t>(6ULL) + (cast<uint8_t>((reg - cast<uint8_t>(22ULL))))));
ps2_GS_write(gs,t0,cast<uint64_t>(((gs->reg[t0] & ~(cast<uint64_t>(18446743936336658432ULL))) | cast<uint64_t>((val & cast<uint64_t>(18446743936336658432ULL))))));
break;}
case cast<uint8_t>(5ULL):{
ps2_GS_pushVertex(gs,cast<int32_t>(cast<uint64_t>((val & cast<uint64_t>(65535ULL)))),cast<int32_t>(cast<uint64_t>((shr<uint64_t>(val,cast<int64_t>(16ULL)) & cast<uint64_t>(65535ULL)))),cast<uint32_t>(shr<uint64_t>(val,cast<int64_t>(32ULL))),cast<uint32_t>(shr<uint64_t>(gs->reg[cast<int64_t>(10ULL)],cast<int64_t>(56ULL))),true);
break;}
case cast<uint8_t>(13ULL):{
ps2_GS_pushVertex(gs,cast<int32_t>(cast<uint64_t>((val & cast<uint64_t>(65535ULL)))),cast<int32_t>(cast<uint64_t>((shr<uint64_t>(val,cast<int64_t>(16ULL)) & cast<uint64_t>(65535ULL)))),cast<uint32_t>(shr<uint64_t>(val,cast<int64_t>(32ULL))),cast<uint32_t>(shr<uint64_t>(gs->reg[cast<int64_t>(10ULL)],cast<int64_t>(56ULL))),false);
break;}
case cast<uint8_t>(4ULL):{
ps2_GS_pushVertex(gs,cast<int32_t>(cast<uint64_t>((val & cast<uint64_t>(65535ULL)))),cast<int32_t>(cast<uint64_t>((shr<uint64_t>(val,cast<int64_t>(16ULL)) & cast<uint64_t>(65535ULL)))),cast<uint32_t>(cast<uint64_t>((shr<uint64_t>(val,cast<int64_t>(32ULL)) & cast<uint64_t>(16777215ULL)))),cast<uint32_t>((cast<uint32_t>(shr<uint64_t>(val,cast<int64_t>(56ULL))) & cast<uint32_t>(255ULL))),true);
break;}
case cast<uint8_t>(12ULL):{
ps2_GS_pushVertex(gs,cast<int32_t>(cast<uint64_t>((val & cast<uint64_t>(65535ULL)))),cast<int32_t>(cast<uint64_t>((shr<uint64_t>(val,cast<int64_t>(16ULL)) & cast<uint64_t>(65535ULL)))),cast<uint32_t>(cast<uint64_t>((shr<uint64_t>(val,cast<int64_t>(32ULL)) & cast<uint64_t>(16777215ULL)))),cast<uint32_t>((cast<uint32_t>(shr<uint64_t>(val,cast<int64_t>(56ULL))) & cast<uint32_t>(255ULL))),false);
break;}
case cast<uint8_t>(83ULL):{
ps2_GS_beginTransfer(gs,cast<uint64_t>((val & cast<uint64_t>(3ULL))));
break;}
case cast<uint8_t>(84ULL):{
ps2_GS_imageData(gs,ps2_u64bytes(val));
break;}
case cast<uint8_t>(97ULL):{
gs->csr |= cast<uint64_t>(2ULL);
break;}
}}
}
}
// tools/platform/ps2/gs.go:263:1
void ps2_GS_writePacked(ps2_GS* gs,uint8_t reg,uint64_t lo,uint64_t hi){
{
{
switch(reg){
case cast<uint8_t>(0ULL):{
ps2_GS_write(gs,cast<uint8_t>(0ULL),cast<uint64_t>((lo & cast<uint64_t>(2047ULL))));
break;}
case cast<uint8_t>(1ULL):{
auto tmp32 = std::make_tuple(cast<uint64_t>((lo & cast<uint64_t>(255ULL))),cast<uint64_t>((shr<uint64_t>(lo,cast<int64_t>(32ULL)) & cast<uint64_t>(255ULL))));
uint64_t r = std::get<0>(tmp32);
uint64_t g = std::get<1>(tmp32);
auto tmp33 = std::make_tuple(cast<uint64_t>((hi & cast<uint64_t>(255ULL))),cast<uint64_t>((shr<uint64_t>(hi,cast<int64_t>(32ULL)) & cast<uint64_t>(255ULL))));
uint64_t b = std::get<0>(tmp33);
uint64_t a = std::get<1>(tmp33);
if ((a == cast<uint64_t>(0ULL))) {
gs->rgbaqA0++;
}
else {
gs->rgbaqA++;
}
if ((cast<uint64_t>((cast<uint64_t>((r | g)) | b)) == cast<uint64_t>(0ULL))) {
gs->rgbaqRGB0++;
}
else {
gs->rgbaqRGB++;
}
ps2_GS_write(gs,cast<uint8_t>(1ULL),cast<uint64_t>((cast<uint64_t>((cast<uint64_t>((cast<uint64_t>((r | shl<uint64_t>(g,cast<int64_t>(8ULL)))) | shl<uint64_t>(b,cast<int64_t>(16ULL)))) | shl<uint64_t>(a,cast<int64_t>(24ULL)))) | shl<uint64_t>(cast<uint64_t>(gs->q),cast<int64_t>(32ULL)))));
break;}
case cast<uint8_t>(2ULL):{
gs->q = cast<uint32_t>(hi);
ps2_GS_write(gs,cast<uint8_t>(2ULL),lo);
break;}
case cast<uint8_t>(3ULL):{
auto tmp34 = std::make_tuple(cast<uint64_t>((lo & cast<uint64_t>(16383ULL))),cast<uint64_t>((shr<uint64_t>(lo,cast<int64_t>(32ULL)) & cast<uint64_t>(16383ULL))));
uint64_t u = std::get<0>(tmp34);
uint64_t v = std::get<1>(tmp34);
ps2_GS_write(gs,cast<uint8_t>(3ULL),cast<uint64_t>((u | shl<uint64_t>(v,cast<int64_t>(16ULL)))));
break;}
case cast<uint8_t>(4ULL):{
auto tmp35 = std::make_tuple(cast<uint64_t>((lo & cast<uint64_t>(65535ULL))),cast<uint64_t>((shr<uint64_t>(lo,cast<int64_t>(32ULL)) & cast<uint64_t>(65535ULL))));
uint64_t x = std::get<0>(tmp35);
uint64_t y = std::get<1>(tmp35);
uint64_t z = cast<uint64_t>((shr<uint64_t>(hi,cast<int64_t>(4ULL)) & cast<uint64_t>(16777215ULL)));
uint64_t f = cast<uint64_t>((shr<uint64_t>(hi,cast<int64_t>(36ULL)) & cast<uint64_t>(255ULL)));
uint64_t val = cast<uint64_t>((cast<uint64_t>((cast<uint64_t>((x | shl<uint64_t>(y,cast<int64_t>(16ULL)))) | shl<uint64_t>(z,cast<int64_t>(32ULL)))) | shl<uint64_t>(f,cast<int64_t>(56ULL))));
if ((cast<uint64_t>((hi & cast<uint64_t>(140737488355328ULL))) != cast<uint64_t>(0ULL))) {
ps2_GS_write(gs,cast<uint8_t>(12ULL),val);
}
else {
ps2_GS_write(gs,cast<uint8_t>(4ULL),val);
}
break;}
case cast<uint8_t>(5ULL):{
auto tmp36 = std::make_tuple(cast<uint64_t>((lo & cast<uint64_t>(65535ULL))),cast<uint64_t>((shr<uint64_t>(lo,cast<int64_t>(32ULL)) & cast<uint64_t>(65535ULL))));
uint64_t x = std::get<0>(tmp36);
uint64_t y = std::get<1>(tmp36);
uint64_t z = cast<uint64_t>((hi & cast<uint64_t>(4294967295ULL)));
uint64_t val = cast<uint64_t>((cast<uint64_t>((x | shl<uint64_t>(y,cast<int64_t>(16ULL)))) | shl<uint64_t>(z,cast<int64_t>(32ULL))));
if ((cast<uint64_t>((hi & cast<uint64_t>(140737488355328ULL))) != cast<uint64_t>(0ULL))) {
ps2_GS_write(gs,cast<uint8_t>(13ULL),val);
}
else {
ps2_GS_write(gs,cast<uint8_t>(5ULL),val);
}
break;}
case cast<uint8_t>(10ULL):{
ps2_GS_write(gs,cast<uint8_t>(10ULL),shl<uint64_t>(cast<uint64_t>((shr<uint64_t>(hi,cast<int64_t>(36ULL)) & cast<uint64_t>(255ULL))),cast<int64_t>(56ULL)));
break;}
default:{
ps2_GS_write(gs,reg,lo);
break;}
}}
}
}
// tools/platform/ps2/gs.go:327:1
void ps2_GS_beginTransfer(ps2_GS* gs,uint64_t dir){
{
uint64_t bitbltbuf = gs->reg[cast<int64_t>(80ULL)];
uint64_t trxpos = gs->reg[cast<int64_t>(81ULL)];
uint64_t trxreg = gs->reg[cast<int64_t>(82ULL)];
if ((dir != cast<uint64_t>(0ULL))) {
gs->xfer.active = false;
if ((dir == cast<uint64_t>(2ULL))) {
ps2_GS_localCopy(gs,bitbltbuf,trxpos,trxreg);
}
else {
ps2_GS_count(gs,std::string("local->host transfer (not served)",33));
}
return ;
}
gs->xfer = ps2_gsXfer{true,cast<uint32_t>((cast<uint32_t>(shr<uint64_t>(bitbltbuf,cast<int64_t>(32ULL))) & cast<uint32_t>(16383ULL))),cast<uint32_t>((cast<uint32_t>(shr<uint64_t>(bitbltbuf,cast<int64_t>(48ULL))) & cast<uint32_t>(63ULL))),cast<uint32_t>((cast<uint32_t>(shr<uint64_t>(bitbltbuf,cast<int64_t>(56ULL))) & cast<uint32_t>(63ULL))),cast<uint32_t>((cast<uint32_t>(shr<uint64_t>(trxpos,cast<int64_t>(32ULL))) & cast<uint32_t>(2047ULL))),cast<uint32_t>((cast<uint32_t>(shr<uint64_t>(trxpos,cast<int64_t>(48ULL))) & cast<uint32_t>(2047ULL))),cast<uint32_t>((cast<uint32_t>(trxreg) & cast<uint32_t>(4095ULL))),cast<uint32_t>((cast<uint32_t>(shr<uint64_t>(trxreg,cast<int64_t>(32ULL))) & cast<uint32_t>(4095ULL))),{},{},{}};
gs->uploads++;
ps2_Machine_note(gs->m,std::string("GS: image upload %dx%d to base 0x%X (width %d, format 0x%X) at (%d,%d)",70),gs->xfer.rrw,gs->xfer.rrh,cast<uint32_t>((gs->xfer.dbp * cast<uint32_t>(64ULL))),cast<uint32_t>((gs->xfer.dbw * cast<uint32_t>(64ULL))),gs->xfer.dpsm,gs->xfer.dsax,gs->xfer.dsay);
if (ps2_xferLog) {
print(ps2_sprintf(std::string("  XFER upload %dx%d dbp 0x%X (blk 0x%X) dbw %d psm 0x%X at (%d,%d) from %s\012",75),gs->xfer.rrw,gs->xfer.rrh,cast<uint32_t>((gs->xfer.dbp * cast<uint32_t>(64ULL))),gs->xfer.dbp,cast<uint32_t>((gs->xfer.dbw * cast<uint32_t>(64ULL))),gs->xfer.dpsm,gs->xfer.dsax,gs->xfer.dsay,gs->src));
}
}
}
// tools/platform/ps2/gs.go:367:1
void ps2_GS_localCopy(ps2_GS* gs,uint64_t bitbltbuf,uint64_t trxpos,uint64_t trxreg){
{auto restore=rrGSTransfer(gs,"ps2_GS_localCopy");
{
uint32_t sbp = cast<uint32_t>((cast<uint32_t>(bitbltbuf) & cast<uint32_t>(16383ULL)));
uint32_t sbw = cast<uint32_t>((cast<uint32_t>(shr<uint64_t>(bitbltbuf,cast<int64_t>(16ULL))) & cast<uint32_t>(63ULL)));
uint32_t spsm = cast<uint32_t>((cast<uint32_t>(shr<uint64_t>(bitbltbuf,cast<int64_t>(24ULL))) & cast<uint32_t>(63ULL)));
uint32_t dbp = cast<uint32_t>((cast<uint32_t>(shr<uint64_t>(bitbltbuf,cast<int64_t>(32ULL))) & cast<uint32_t>(16383ULL)));
uint32_t dbw = cast<uint32_t>((cast<uint32_t>(shr<uint64_t>(bitbltbuf,cast<int64_t>(48ULL))) & cast<uint32_t>(63ULL)));
uint32_t dpsm = cast<uint32_t>((cast<uint32_t>(shr<uint64_t>(bitbltbuf,cast<int64_t>(56ULL))) & cast<uint32_t>(63ULL)));
uint32_t ssax = cast<uint32_t>((cast<uint32_t>(trxpos) & cast<uint32_t>(2047ULL)));
uint32_t ssay = cast<uint32_t>((cast<uint32_t>(shr<uint64_t>(trxpos,cast<int64_t>(16ULL))) & cast<uint32_t>(2047ULL)));
uint32_t dsax = cast<uint32_t>((cast<uint32_t>(shr<uint64_t>(trxpos,cast<int64_t>(32ULL))) & cast<uint32_t>(2047ULL)));
uint32_t dsay = cast<uint32_t>((cast<uint32_t>(shr<uint64_t>(trxpos,cast<int64_t>(48ULL))) & cast<uint32_t>(2047ULL)));
uint32_t rrw = cast<uint32_t>((cast<uint32_t>(trxreg) & cast<uint32_t>(4095ULL)));
uint32_t rrh = cast<uint32_t>((cast<uint32_t>(shr<uint64_t>(trxreg,cast<int64_t>(32ULL))) & cast<uint32_t>(4095ULL)));
ps2_Machine_note(gs->m,std::string("GS: local copy %dx%d from 0x%X (psm 0x%02X) to 0x%X (psm 0x%02X)",64),rrw,rrh,cast<uint32_t>((sbp * cast<uint32_t>(64ULL))),spsm,cast<uint32_t>((dbp * cast<uint32_t>(64ULL))),dpsm);
ps2_GS_count(gs,ps2_sprintf(std::string("local->local 0x%02X->0x%02X",27),spsm,dpsm));
if (ps2_xferLog) {
print(ps2_sprintf(std::string("  XFER copy %dx%d sbp 0x%X psm 0x%X -> dbp 0x%X psm 0x%X at (%d,%d)\012",68),rrw,rrh,cast<uint32_t>((sbp * cast<uint32_t>(64ULL))),spsm,cast<uint32_t>((dbp * cast<uint32_t>(64ULL))),dpsm,dsax,dsay));
}
Slice<uint32_t> row = Slice<uint32_t>::make(rrw);
{uint32_t y = cast<uint32_t>(0ULL);for (;(y < rrh);y++){
{uint32_t x = cast<uint32_t>(0ULL);for (;(x < rrw);x++){
auto tmp37 = ps2_GS_readTexel(gs,spsm,sbp,sbw,cast<uint32_t>((ssax + x)),cast<uint32_t>((ssay + y)));
row[x] = std::get<0>(tmp37);
}
}{uint32_t x = cast<uint32_t>(0ULL);for (;(x < rrw);x++){
ps2_GS_writeTexel(gs,dpsm,dbp,dbw,cast<uint32_t>((dsax + x)),cast<uint32_t>((dsay + y)),row[x]);
}
}}
}}
}
}
// tools/platform/ps2/gs.go:405:1
std::tuple<uint32_t,bool> ps2_GS_readTexel(ps2_GS* gs,uint32_t psm,uint32_t bp,uint32_t bw,uint32_t x,uint32_t y){
{
{
switch(psm){
case cast<uint32_t>(0ULL):case cast<uint32_t>(1ULL):{
uint32_t a = ps2_addrPSMCT32(bp,bw,x,y);
if ((cast<uint32_t>((a + cast<uint32_t>(4ULL))) <= cast<uint32_t>(len(gs->vram)))) {
return {ps2_le32gs(rrBorrow(gs->vram,a,len(gs->vram))),true};
}
break;}
case cast<uint32_t>(2ULL):case cast<uint32_t>(10ULL):{
uint32_t a = ps2_addrPSMCT16(bp,bw,x,y,(psm == cast<uint32_t>(10ULL)));
if ((cast<uint32_t>((a + cast<uint32_t>(2ULL))) <= cast<uint32_t>(len(gs->vram)))) {
return {cast<uint32_t>((cast<uint32_t>(gs->vram[a]) | shl<uint32_t>(cast<uint32_t>(gs->vram[cast<uint32_t>((a + cast<uint32_t>(1ULL)))]),cast<int64_t>(8ULL)))),true};
}
break;}
case cast<uint32_t>(19ULL):{
uint32_t a = ps2_addrPSMT8(bp,bw,x,y);
if ((a < cast<uint32_t>(len(gs->vram)))) {
return {cast<uint32_t>(gs->vram[a]),true};
}
break;}
case cast<uint32_t>(20ULL):{
auto tmp38 = ps2_addrPSMT4(bp,bw,x,y);
uint32_t a = std::get<0>(tmp38);
uint32_t nib = std::get<1>(tmp38);
if ((a < cast<uint32_t>(len(gs->vram)))) {
return {cast<uint32_t>((shr<uint32_t>(cast<uint32_t>(gs->vram[a]),(cast<uint32_t>((cast<uint32_t>(4ULL) * nib)))) & cast<uint32_t>(15ULL))),true};
}
break;}
case cast<uint32_t>(27ULL):{
uint32_t a = ps2_addrPSMCT32(bp,bw,x,y);
if ((cast<uint32_t>((a + cast<uint32_t>(4ULL))) <= cast<uint32_t>(len(gs->vram)))) {
return {cast<uint32_t>(gs->vram[cast<uint32_t>((a + cast<uint32_t>(3ULL)))]),true};
}
break;}
case cast<uint32_t>(36ULL):{
uint32_t a = ps2_addrPSMCT32(bp,bw,x,y);
if ((cast<uint32_t>((a + cast<uint32_t>(4ULL))) <= cast<uint32_t>(len(gs->vram)))) {
return {cast<uint32_t>((cast<uint32_t>(gs->vram[cast<uint32_t>((a + cast<uint32_t>(3ULL)))]) & cast<uint32_t>(15ULL))),true};
}
break;}
case cast<uint32_t>(44ULL):{
uint32_t a = ps2_addrPSMCT32(bp,bw,x,y);
if ((cast<uint32_t>((a + cast<uint32_t>(4ULL))) <= cast<uint32_t>(len(gs->vram)))) {
return {shr<uint32_t>(cast<uint32_t>(gs->vram[cast<uint32_t>((a + cast<uint32_t>(3ULL)))]),cast<int64_t>(4ULL)),true};
}
break;}
}}
return {cast<uint32_t>(0ULL),false};
}
}
// tools/platform/ps2/gs.go:449:1
void ps2_GS_writeTexel(ps2_GS* gs,uint32_t psm,uint32_t bp,uint32_t bw,uint32_t x,uint32_t y,uint32_t v){
{
{
switch(psm){
case cast<uint32_t>(0ULL):{
uint32_t a = ps2_addrPSMCT32(bp,bw,x,y);
if ((cast<uint32_t>((a + cast<uint32_t>(4ULL))) <= cast<uint32_t>(len(gs->vram)))) {
gs->vram[cast<uint32_t>((a + cast<uint32_t>(0ULL)))] = cast<uint8_t>(v);
if(rrcapture::trace.active)rrGSWrite(gs,cast<uint32_t>((a + cast<uint32_t>(0ULL))));
gs->vram[cast<uint32_t>((a + cast<uint32_t>(1ULL)))] = cast<uint8_t>(shr<uint32_t>(v,cast<int64_t>(8ULL)));
if(rrcapture::trace.active)rrGSWrite(gs,cast<uint32_t>((a + cast<uint32_t>(1ULL))));
gs->vram[cast<uint32_t>((a + cast<uint32_t>(2ULL)))] = cast<uint8_t>(shr<uint32_t>(v,cast<int64_t>(16ULL)));
if(rrcapture::trace.active)rrGSWrite(gs,cast<uint32_t>((a + cast<uint32_t>(2ULL))));
gs->vram[cast<uint32_t>((a + cast<uint32_t>(3ULL)))] = cast<uint8_t>(shr<uint32_t>(v,cast<int64_t>(24ULL)));
if(rrcapture::trace.active)rrGSWrite(gs,cast<uint32_t>((a + cast<uint32_t>(3ULL))));
}
break;}
case cast<uint32_t>(1ULL):{
uint32_t a = ps2_addrPSMCT32(bp,bw,x,y);
if ((cast<uint32_t>((a + cast<uint32_t>(4ULL))) <= cast<uint32_t>(len(gs->vram)))) {
gs->vram[cast<uint32_t>((a + cast<uint32_t>(0ULL)))] = cast<uint8_t>(v);
if(rrcapture::trace.active)rrGSWrite(gs,cast<uint32_t>((a + cast<uint32_t>(0ULL))));
gs->vram[cast<uint32_t>((a + cast<uint32_t>(1ULL)))] = cast<uint8_t>(shr<uint32_t>(v,cast<int64_t>(8ULL)));
if(rrcapture::trace.active)rrGSWrite(gs,cast<uint32_t>((a + cast<uint32_t>(1ULL))));
gs->vram[cast<uint32_t>((a + cast<uint32_t>(2ULL)))] = cast<uint8_t>(shr<uint32_t>(v,cast<int64_t>(16ULL)));
if(rrcapture::trace.active)rrGSWrite(gs,cast<uint32_t>((a + cast<uint32_t>(2ULL))));
}
break;}
case cast<uint32_t>(2ULL):case cast<uint32_t>(10ULL):{
uint32_t a = ps2_addrPSMCT16(bp,bw,x,y,(psm == cast<uint32_t>(10ULL)));
if ((cast<uint32_t>((a + cast<uint32_t>(2ULL))) <= cast<uint32_t>(len(gs->vram)))) {
gs->vram[cast<uint32_t>((a + cast<uint32_t>(0ULL)))] = cast<uint8_t>(v);
if(rrcapture::trace.active)rrGSWrite(gs,cast<uint32_t>((a + cast<uint32_t>(0ULL))));
gs->vram[cast<uint32_t>((a + cast<uint32_t>(1ULL)))] = cast<uint8_t>(shr<uint32_t>(v,cast<int64_t>(8ULL)));
if(rrcapture::trace.active)rrGSWrite(gs,cast<uint32_t>((a + cast<uint32_t>(1ULL))));
}
break;}
case cast<uint32_t>(19ULL):{
uint32_t a = ps2_addrPSMT8(bp,bw,x,y);
if ((a < cast<uint32_t>(len(gs->vram)))) {
gs->vram[a] = cast<uint8_t>(v);
if(rrcapture::trace.active)rrGSWrite(gs,a);
}
break;}
case cast<uint32_t>(20ULL):{
auto tmp39 = ps2_addrPSMT4(bp,bw,x,y);
uint32_t a = std::get<0>(tmp39);
uint32_t nib = std::get<1>(tmp39);
if ((a < cast<uint32_t>(len(gs->vram)))) {
gs->vram[a] = cast<uint8_t>(((gs->vram[a] & ~((shl<uint8_t>(cast<uint8_t>(15ULL),(cast<uint32_t>((cast<uint32_t>(4ULL) * nib))))))) | shl<uint8_t>(cast<uint8_t>(cast<uint32_t>((v & cast<uint32_t>(15ULL)))),(cast<uint32_t>((cast<uint32_t>(4ULL) * nib))))));
if(rrcapture::trace.active)rrGSWrite(gs,a);
}
break;}
case cast<uint32_t>(27ULL):{
uint32_t a = ps2_addrPSMCT32(bp,bw,x,y);
if ((cast<uint32_t>((a + cast<uint32_t>(4ULL))) <= cast<uint32_t>(len(gs->vram)))) {
gs->vram[cast<uint32_t>((a + cast<uint32_t>(3ULL)))] = cast<uint8_t>(v);
if(rrcapture::trace.active)rrGSWrite(gs,cast<uint32_t>((a + cast<uint32_t>(3ULL))));
}
break;}
case cast<uint32_t>(36ULL):{
uint32_t a = ps2_addrPSMCT32(bp,bw,x,y);
if ((cast<uint32_t>((a + cast<uint32_t>(4ULL))) <= cast<uint32_t>(len(gs->vram)))) {
gs->vram[cast<uint32_t>((a + cast<uint32_t>(3ULL)))] = cast<uint8_t>((cast<uint8_t>((gs->vram[cast<uint32_t>((a + cast<uint32_t>(3ULL)))] & cast<uint8_t>(240ULL))) | cast<uint8_t>(cast<uint32_t>((v & cast<uint32_t>(15ULL))))));
if(rrcapture::trace.active)rrGSWrite(gs,cast<uint32_t>((a + cast<uint32_t>(3ULL))));
}
break;}
case cast<uint32_t>(44ULL):{
uint32_t a = ps2_addrPSMCT32(bp,bw,x,y);
if ((cast<uint32_t>((a + cast<uint32_t>(4ULL))) <= cast<uint32_t>(len(gs->vram)))) {
gs->vram[cast<uint32_t>((a + cast<uint32_t>(3ULL)))] = cast<uint8_t>((cast<uint8_t>((gs->vram[cast<uint32_t>((a + cast<uint32_t>(3ULL)))] & cast<uint8_t>(15ULL))) | shl<uint8_t>(cast<uint8_t>(cast<uint32_t>((v & cast<uint32_t>(15ULL)))),cast<int64_t>(4ULL))));
if(rrcapture::trace.active)rrGSWrite(gs,cast<uint32_t>((a + cast<uint32_t>(3ULL))));
}
break;}
}}
}
}
// tools/platform/ps2/gs.go:504:1
void ps2_GS_imageData(ps2_GS* gs,Slice<uint8_t> data){
{rrprof::Scope timing(3,"GS transfers");
{auto restore=rrGSTransfer(gs,"ps2_GS_imageData");
{
if ((((!gs->xfer.active) || (gs->xfer.rrw == cast<uint32_t>(0ULL))) || (gs->xfer.rrh == cast<uint32_t>(0ULL)))) {
return ;
}
ps2_gsXfer x = gs->xfer;
Slice<uint8_t> buf = append(x.partial,data);
x.partial = {};
auto advance = [&]()->bool{
x.x++;
if ((x.x >= x.rrw)) {
x.x = cast<uint32_t>(0ULL);
x.y++;
if ((x.y >= x.rrh)) {
x.active = false;
return false;
}
}
return true;
}
;
{
switch(x.dpsm){
case cast<uint32_t>(0ULL):case cast<uint32_t>(1ULL):{
int64_t i = cast<int64_t>(0ULL);
{;for (;((cast<int64_t>((i + cast<int64_t>(4ULL))) <= len(buf)) && x.active);){
uint32_t px = ps2_le32gs(rrBorrow(buf,i,len(buf)));
i += cast<int64_t>(4ULL);
uint32_t addr = ps2_addrPSMCT32(x.dbp,x.dbw,cast<uint32_t>((x.dsax + x.x)),cast<uint32_t>((x.dsay + x.y)));
if ((cast<uint32_t>((addr + cast<uint32_t>(4ULL))) <= cast<uint32_t>(len(gs->vram)))) {
gs->vram[cast<uint32_t>((addr + cast<uint32_t>(0ULL)))] = cast<uint8_t>(px);
if(rrcapture::trace.active)rrGSWrite(gs,cast<uint32_t>((addr + cast<uint32_t>(0ULL))));
gs->vram[cast<uint32_t>((addr + cast<uint32_t>(1ULL)))] = cast<uint8_t>(shr<uint32_t>(px,cast<int64_t>(8ULL)));
if(rrcapture::trace.active)rrGSWrite(gs,cast<uint32_t>((addr + cast<uint32_t>(1ULL))));
gs->vram[cast<uint32_t>((addr + cast<uint32_t>(2ULL)))] = cast<uint8_t>(shr<uint32_t>(px,cast<int64_t>(16ULL)));
if(rrcapture::trace.active)rrGSWrite(gs,cast<uint32_t>((addr + cast<uint32_t>(2ULL))));
gs->vram[cast<uint32_t>((addr + cast<uint32_t>(3ULL)))] = cast<uint8_t>(shr<uint32_t>(px,cast<int64_t>(24ULL)));
if(rrcapture::trace.active)rrGSWrite(gs,cast<uint32_t>((addr + cast<uint32_t>(3ULL))));
}
advance();
}
}x.partial = append(x.partial,sub(buf,i,len(buf)));
break;}
case cast<uint32_t>(2ULL):case cast<uint32_t>(10ULL):{
int64_t i = cast<int64_t>(0ULL);
{;for (;((cast<int64_t>((i + cast<int64_t>(2ULL))) <= len(buf)) && x.active);){
uint32_t px = cast<uint32_t>((cast<uint32_t>(buf[i]) | shl<uint32_t>(cast<uint32_t>(buf[cast<int64_t>((i + cast<int64_t>(1ULL)))]),cast<int64_t>(8ULL))));
i += cast<int64_t>(2ULL);
uint32_t addr = ps2_addrPSMCT16(x.dbp,x.dbw,cast<uint32_t>((x.dsax + x.x)),cast<uint32_t>((x.dsay + x.y)),(x.dpsm == cast<uint32_t>(10ULL)));
if ((cast<uint32_t>((addr + cast<uint32_t>(2ULL))) <= cast<uint32_t>(len(gs->vram)))) {
gs->vram[cast<uint32_t>((addr + cast<uint32_t>(0ULL)))] = cast<uint8_t>(px);
if(rrcapture::trace.active)rrGSWrite(gs,cast<uint32_t>((addr + cast<uint32_t>(0ULL))));
gs->vram[cast<uint32_t>((addr + cast<uint32_t>(1ULL)))] = cast<uint8_t>(shr<uint32_t>(px,cast<int64_t>(8ULL)));
if(rrcapture::trace.active)rrGSWrite(gs,cast<uint32_t>((addr + cast<uint32_t>(1ULL))));
}
advance();
}
}x.partial = append(x.partial,sub(buf,i,len(buf)));
break;}
case cast<uint32_t>(19ULL):{
int64_t i = cast<int64_t>(0ULL);
{;for (;((i < len(buf)) && x.active);){
uint32_t addr = ps2_addrPSMT8(x.dbp,x.dbw,cast<uint32_t>((x.dsax + x.x)),cast<uint32_t>((x.dsay + x.y)));
if ((addr < cast<uint32_t>(len(gs->vram)))) {
gs->vram[addr] = buf[i];
if(rrcapture::trace.active)rrGSWrite(gs,addr);
}
i++;
advance();
}
}break;}
case cast<uint32_t>(20ULL):{
int64_t i = cast<int64_t>(0ULL);
{;for (;((i < len(buf)) && x.active);){
{int64_t half = cast<int64_t>(0ULL);for (;((half < cast<int64_t>(2ULL)) && x.active);half++){
auto tmp40 = ps2_addrPSMT4(x.dbp,x.dbw,cast<uint32_t>((x.dsax + x.x)),cast<uint32_t>((x.dsay + x.y)));
uint32_t addr = std::get<0>(tmp40);
uint32_t nib = std::get<1>(tmp40);
if ((addr < cast<uint32_t>(len(gs->vram)))) {
uint8_t v = cast<uint8_t>((shr<uint8_t>(buf[i],(cast<int64_t>((cast<int64_t>(4ULL) * half)))) & cast<uint8_t>(15ULL)));
gs->vram[addr] = cast<uint8_t>(((gs->vram[addr] & ~((shl<uint8_t>(cast<uint8_t>(15ULL),(cast<uint32_t>((cast<uint32_t>(4ULL) * nib))))))) | shl<uint8_t>(v,(cast<uint32_t>((cast<uint32_t>(4ULL) * nib))))));
if(rrcapture::trace.active)rrGSWrite(gs,addr);
}
advance();
}
}i++;
}
}break;}
default:{
ps2_Machine_note(gs->m,std::string("GS: image upload in format 0x%X (%d bytes) \342\200\224 consumed, not yet placed",71),x.dpsm,len(buf));
x.active = false;
break;}
}}
gs->xfer = x;
}
}
}
}
// tools/platform/ps2/gs.go:607:1
uint32_t ps2_addrPSMCT32(uint32_t bp,uint32_t bw,uint32_t x,uint32_t y){
{
if ((bw == cast<uint32_t>(0ULL))) {
bw = cast<uint32_t>(1ULL);
}
constexpr int64_t pageW=64ULL;
constexpr int64_t pageH=32ULL;
uint32_t pageX = divi<uint32_t>(x,cast<uint32_t>(64ULL));
uint32_t pageY = divi<uint32_t>(y,cast<uint32_t>(32ULL));
uint32_t pagesPerRow = bw;
uint32_t page = cast<uint32_t>((cast<uint32_t>((pageY * pagesPerRow)) + pageX));
uint32_t px = modi<uint32_t>(x,cast<uint32_t>(64ULL));
uint32_t py = modi<uint32_t>(y,cast<uint32_t>(32ULL));
uint32_t bx = divi<uint32_t>(px,cast<uint32_t>(8ULL));
uint32_t by = divi<uint32_t>(py,cast<uint32_t>(8ULL));
uint8_t block = ps2_blockPSMCT32[by][bx];
uint32_t cx = modi<uint32_t>(px,cast<uint32_t>(8ULL));
uint32_t cy = modi<uint32_t>(py,cast<uint32_t>(8ULL));
uint8_t col = ps2_columnPSMCT32[cy][cx];
uint32_t word = cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((bp * cast<uint32_t>(64ULL))) + cast<uint32_t>((page * cast<uint32_t>(2048ULL))))) + cast<uint32_t>((cast<uint32_t>(block) * cast<uint32_t>(64ULL))))) + cast<uint32_t>(col)));
return cast<uint32_t>((word * cast<uint32_t>(4ULL)));
}
}
// tools/platform/ps2/gs.go:674:1
std::tuple<uint32_t,bool> ps2_Machine_gsPrivRead(ps2_Machine* m,uint32_t a){
{
ps2_Machine_drainVIF1(m);
ps2_GS* gs = ps2_Machine_ensureGS(m);
{
switch((a & ~(cast<uint32_t>(4ULL)))){
case cast<uint32_t>(301993984ULL):{
if ((cast<uint32_t>((a & cast<uint32_t>(4ULL))) != cast<uint32_t>(0ULL))) {
return {cast<uint32_t>(shr<uint64_t>(gs->csr,cast<int64_t>(32ULL))),true};
}
return {cast<uint32_t>(gs->csr),true};
break;}
}}
return {get(m->io,a),true};
}
}
// tools/platform/ps2/gs.go:690:1
bool ps2_Machine_gsPrivWrite(ps2_Machine* m,uint32_t a,uint32_t v){
{
ps2_Machine_drainVIF1(m);
ps2_GS* gs = ps2_Machine_ensureGS(m);
{
switch(a){
case cast<uint32_t>(301993984ULL):{
if ((cast<uint32_t>((v & cast<uint32_t>(512ULL))) != cast<uint32_t>(0ULL))) {
gs->csr = cast<uint64_t>(0ULL);
}
gs->csr &= ~(cast<uint64_t>((cast<uint64_t>(v) & cast<uint64_t>(31ULL))));
break;}
case cast<uint32_t>(301993988ULL):{
break;}
}}
if (((m->LogDISPFB && ((((a == cast<uint32_t>(301990032ULL)) || (a == cast<uint32_t>(301990000ULL))) || (a == cast<uint32_t>(301989888ULL))))) && (get(m->io,a) != v))) {
go_fmt_Printf(std::string("  dispfb: [%08X] <- %08X (was %08X) at prim %d (run %d) vblank %d\012",66),a,v,get(m->io,a),gs->prims,cast<int64_t>((gs->prims - gs->primsRunBase)),ps2_Machine_VBlanks(m));
}
bool isFlip = ((((a == cast<uint32_t>(301990032ULL)) || (a == cast<uint32_t>(301990000ULL)))) && (get(m->io,a) != v));
uint32_t old = get(m->io,a);
m->io[a] = v;
if ((isFlip && m->gsWeave)) {
{
auto tmp41 = ps2_Machine_GSFrame(m);
Slice<uint8_t> pix = std::get<0>(tmp41);
int64_t w = std::get<1>(tmp41);
int64_t h = std::get<2>(tmp41);
if (bool(pix)) {
uint32_t p = cast<uint32_t>((ps2_Machine_VBlanks(m) & cast<uint32_t>(1ULL)));
m->weaveRing[p] = ps2_weaveField{pix,w,h};
m->weaveLast = cast<int8_t>(p);
}
}
}
if ((bool(m->OnGSFlip) && isFlip)) {
m->OnGSFlip(a,v,old);
}
return true;
}
}
// tools/platform/ps2/gs.go:736:1
void ps2_Machine_SetGSWeave(ps2_Machine* m,bool on){
{
m->gsWeave = on;
}
}
// tools/platform/ps2/gs.go:745:1
std::tuple<Slice<uint8_t>,int64_t,int64_t,int64_t> ps2_Machine_GSFrameWoven(ps2_Machine* m){
Slice<uint8_t> pix{};
int64_t w{};
int64_t h{};
int64_t newestParity{};
{
auto tmp42 = std::make_tuple((&m->weaveRing[cast<int64_t>(0ULL)]),(&m->weaveRing[cast<int64_t>(1ULL)]));
ps2_weaveField* e = std::get<0>(tmp42);
ps2_weaveField* o = std::get<1>(tmp42);
if (((((e->w == cast<int64_t>(0ULL)) || (o->w == cast<int64_t>(0ULL))) || (e->w != o->w)) || (e->h != o->h))) {
auto tmp43 = ps2_Machine_GSFrame(m);
Slice<uint8_t> p = std::get<0>(tmp43);
int64_t fw = std::get<1>(tmp43);
int64_t fh = std::get<2>(tmp43);
return {p,fw,fh,cast<int64_t>(-1ULL)};
}
auto tmp44 = std::make_tuple(e->w,cast<int64_t>((e->h * cast<int64_t>(2ULL))));
w = std::get<0>(tmp44);
h = std::get<1>(tmp44);
pix = Slice<uint8_t>::make(cast<int64_t>((cast<int64_t>((w * h)) * cast<int64_t>(4ULL))));
int64_t row = cast<int64_t>((w * cast<int64_t>(4ULL)));
{int64_t y = cast<int64_t>(0ULL);for (;(y < e->h);y++){
gcopy(sub(pix,cast<int64_t>(((cast<int64_t>((cast<int64_t>(2ULL) * y))) * row)),cast<int64_t>((cast<int64_t>(((cast<int64_t>((cast<int64_t>(2ULL) * y))) * row)) + row))),sub(e->pix,cast<int64_t>((y * row)),cast<int64_t>((cast<int64_t>((y * row)) + row))));
gcopy(sub(pix,cast<int64_t>(((cast<int64_t>((cast<int64_t>((cast<int64_t>(2ULL) * y)) + cast<int64_t>(1ULL)))) * row)),cast<int64_t>((cast<int64_t>(((cast<int64_t>((cast<int64_t>((cast<int64_t>(2ULL) * y)) + cast<int64_t>(1ULL)))) * row)) + row))),sub(o->pix,cast<int64_t>((y * row)),cast<int64_t>((cast<int64_t>((y * row)) + row))));
}
}return {pix,w,h,cast<int64_t>(m->weaveLast)};
}
}
// tools/platform/ps2/gs.go:765:1
std::string ps2_Machine_GSStatus(ps2_Machine* m){
{
if ((!m->gs)) {
return std::string("the GS was never touched\012",25);
}
std::string s = ps2_sprintf(std::string("the GS: %d image uploads, %d vertex kicks\012",42),m->gs->uploads,m->gs->prims);
uint64_t dispfb2 = cast<uint64_t>((cast<uint64_t>(get(m->io,cast<uint32_t>(301990032ULL))) | shl<uint64_t>(cast<uint64_t>(get(m->io,cast<uint32_t>(301990036ULL))),cast<int64_t>(32ULL))));
uint64_t dispfb1 = cast<uint64_t>((cast<uint64_t>(get(m->io,cast<uint32_t>(301990000ULL))) | shl<uint64_t>(cast<uint64_t>(get(m->io,cast<uint32_t>(301990004ULL))),cast<int64_t>(32ULL))));
s += ps2_sprintf(std::string("      display: PMODE 0x%X, DISPFB1 fb 0x%05X psm 0x%02X, DISPFB2 fb 0x%05X psm 0x%02X\012",86),get(m->io,cast<uint32_t>(301989888ULL)),cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(dispfb1) & cast<uint32_t>(511ULL))) * cast<uint32_t>(2048ULL))),cast<uint32_t>((cast<uint32_t>(shr<uint64_t>(dispfb1,cast<int64_t>(15ULL))) & cast<uint32_t>(31ULL))),cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(dispfb2) & cast<uint32_t>(511ULL))) * cast<uint32_t>(2048ULL))),cast<uint32_t>((cast<uint32_t>(shr<uint64_t>(dispfb2,cast<int64_t>(15ULL))) & cast<uint32_t>(31ULL))));
s += ps2_sprintf(std::string("      pixels: %d plotted; rejected %d scissor, %d ztest, %d alphatest, %d datest\012",81),m->gs->plotted,m->gs->rejScissor,m->gs->rejZ,m->gs->rejAlpha,m->gs->rejDate);
s += ps2_sprintf(std::string("      FOGCOL 0x%06X, XYOFFSET1 (%d,%d), XYOFFSET2 (%d,%d), PABE %d, COLCLAMP %d\012",80),cast<uint32_t>((cast<uint32_t>(m->gs->reg[cast<int64_t>(61ULL)]) & cast<uint32_t>(16777215ULL))),shr<uint32_t>(cast<uint32_t>((cast<uint32_t>(m->gs->reg[cast<int64_t>(24ULL)]) & cast<uint32_t>(65535ULL))),cast<int64_t>(4ULL)),shr<uint32_t>(cast<uint32_t>((cast<uint32_t>(shr<uint64_t>(m->gs->reg[cast<int64_t>(24ULL)],cast<int64_t>(32ULL))) & cast<uint32_t>(65535ULL))),cast<int64_t>(4ULL)),shr<uint32_t>(cast<uint32_t>((cast<uint32_t>(m->gs->reg[cast<int64_t>(25ULL)]) & cast<uint32_t>(65535ULL))),cast<int64_t>(4ULL)),shr<uint32_t>(cast<uint32_t>((cast<uint32_t>(shr<uint64_t>(m->gs->reg[cast<int64_t>(25ULL)],cast<int64_t>(32ULL))) & cast<uint32_t>(65535ULL))),cast<int64_t>(4ULL)),cast<uint32_t>((cast<uint32_t>(m->gs->reg[cast<int64_t>(73ULL)]) & cast<uint32_t>(1ULL))),cast<uint32_t>((cast<uint32_t>(m->gs->reg[cast<int64_t>(70ULL)]) & cast<uint32_t>(1ULL))));
{auto&& tmp45 = m->gs->plotNonBlack;
for(int64_t tmp46=0;tmp46<len(tmp45);++tmp46){
auto i=tmp46;auto n=tmp45[tmp46];if ((n > cast<uint64_t>(0ULL))) {
s += ps2_sprintf(std::string("      non-black pixels by %-10s %d\012",35),ps2_primNames[i],n);
}
}}
s += ps2_sprintf(std::string("      vertex colours: %d with alpha 0, %d with alpha > 0; %d with black RGB, %d with colour\012",92),m->gs->rgbaqA0,m->gs->rgbaqA,m->gs->rgbaqRGB0,m->gs->rgbaqRGB);
s += ps2_sprintf(std::string("      texel samples: %d black, %d coloured\012",43),m->gs->texBlack,m->gs->texColor);
{int64_t psm = cast<int64_t>(0ULL);for (;(psm < cast<int64_t>(64ULL));psm++){
if ((cast<uint64_t>((m->gs->texBlackPSM[psm] + m->gs->texColorPSM[psm])) > cast<uint64_t>(0ULL))) {
s += ps2_sprintf(std::string("        psm 0x%02X: %d black, %d coloured\012",42),psm,m->gs->texBlackPSM[psm],m->gs->texColorPSM[psm]);
}
}
}{auto&& tmp47 = m->gs->primCount;
for(int64_t tmp48=0;tmp48<len(tmp47);++tmp48){
auto i=tmp48;auto n=tmp47[tmp48];if ((n > cast<int64_t>(0ULL))) {
s += ps2_sprintf(std::string("      %-24s %d\012",15),ps2_primNames[i],n);
}
}}
{auto&& tmp49 = ps2_sortedCounts(m->gs->drawCensus);
for(int64_t tmp50=0;tmp50<len(tmp49);++tmp50){
auto kv=tmp49[tmp50];s += ps2_sprintf(std::string("      %-24s %d\012",15),kv.name,kv.n);
}}
return s;
}
}
// tools/platform/ps2/gs.go:817:1
std::tuple<Slice<uint8_t>,int64_t,int64_t> ps2_Machine_GSFrame(ps2_Machine* m){
Slice<uint8_t> pix{};
int64_t w{};
int64_t h{};
{
if ((!m->gs)) {
return {{},cast<int64_t>(0ULL),cast<int64_t>(0ULL)};
}
uint64_t dispfb = cast<uint64_t>((cast<uint64_t>(get(m->io,cast<uint32_t>(301990032ULL))) | shl<uint64_t>(cast<uint64_t>(get(m->io,cast<uint32_t>(301990036ULL))),cast<int64_t>(32ULL))));
uint64_t display = cast<uint64_t>((cast<uint64_t>(get(m->io,cast<uint32_t>(301990048ULL))) | shl<uint64_t>(cast<uint64_t>(get(m->io,cast<uint32_t>(301990052ULL))),cast<int64_t>(32ULL))));
{
uint32_t pmode = get(m->io,cast<uint32_t>(301989888ULL));
if (((cast<uint32_t>((pmode & cast<uint32_t>(2ULL))) == cast<uint32_t>(0ULL)) && (cast<uint32_t>((pmode & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL)))) {
dispfb = cast<uint64_t>((cast<uint64_t>(get(m->io,cast<uint32_t>(301990000ULL))) | shl<uint64_t>(cast<uint64_t>(get(m->io,cast<uint32_t>(301990004ULL))),cast<int64_t>(32ULL))));
display = cast<uint64_t>((cast<uint64_t>(get(m->io,cast<uint32_t>(301990016ULL))) | shl<uint64_t>(cast<uint64_t>(get(m->io,cast<uint32_t>(301990020ULL))),cast<int64_t>(32ULL))));
}
}
uint32_t fbp = cast<uint32_t>((cast<uint32_t>(dispfb) & cast<uint32_t>(511ULL)));
uint32_t fbw = cast<uint32_t>((cast<uint32_t>(shr<uint64_t>(dispfb,cast<int64_t>(9ULL))) & cast<uint32_t>(63ULL)));
uint32_t psm = cast<uint32_t>((cast<uint32_t>(shr<uint64_t>(dispfb,cast<int64_t>(15ULL))) & cast<uint32_t>(31ULL)));
uint32_t dbx = cast<uint32_t>((cast<uint32_t>(shr<uint64_t>(dispfb,cast<int64_t>(32ULL))) & cast<uint32_t>(2047ULL)));
uint32_t dby = cast<uint32_t>((cast<uint32_t>(shr<uint64_t>(dispfb,cast<int64_t>(43ULL))) & cast<uint32_t>(2047ULL)));
uint32_t magh = cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(shr<uint64_t>(display,cast<int64_t>(23ULL))) & cast<uint32_t>(15ULL))) + cast<uint32_t>(1ULL)));
uint32_t magv = cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(shr<uint64_t>(display,cast<int64_t>(27ULL))) & cast<uint32_t>(3ULL))) + cast<uint32_t>(1ULL)));
uint32_t dw = divi<uint32_t>((cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(shr<uint64_t>(display,cast<int64_t>(32ULL))) & cast<uint32_t>(4095ULL))) + cast<uint32_t>(1ULL)))),magh);
uint32_t dh = divi<uint32_t>((cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(shr<uint64_t>(display,cast<int64_t>(44ULL))) & cast<uint32_t>(2047ULL))) + cast<uint32_t>(1ULL)))),magv);
if (((m->gsInterlace == cast<uint32_t>(1ULL)) && (m->gsFieldMode == cast<uint32_t>(1ULL)))) {
dh /= cast<uint32_t>(2ULL);
}
if ((((fbw == cast<uint32_t>(0ULL)) || (dw == cast<uint32_t>(0ULL))) || (dh == cast<uint32_t>(0ULL)))) {
return {{},cast<int64_t>(0ULL),cast<int64_t>(0ULL)};
}
if (((psm != cast<uint32_t>(0ULL)) && (psm != cast<uint32_t>(1ULL)))) {
ps2_Machine_note(m,std::string("GS: GSFrame asked to read a PSM 0x%X display buffer \342\200\224 only CT32/CT24 read back yet",84),psm);
return {{},cast<int64_t>(0ULL),cast<int64_t>(0ULL)};
}
auto tmp51 = std::make_tuple(cast<int64_t>(dw),cast<int64_t>(dh));
w = std::get<0>(tmp51);
h = std::get<1>(tmp51);
pix = Slice<uint8_t>::make(cast<int64_t>((cast<int64_t>((w * h)) * cast<int64_t>(4ULL))));
{int64_t y = cast<int64_t>(0ULL);for (;(y < h);y++){
{int64_t x = cast<int64_t>(0ULL);for (;(x < w);x++){
uint32_t addr = ps2_addrPSMCT32(cast<uint32_t>((fbp * cast<uint32_t>(32ULL))),fbw,cast<uint32_t>((dbx + cast<uint32_t>(x))),cast<uint32_t>((dby + cast<uint32_t>(y))));
int64_t o = cast<int64_t>(((cast<int64_t>((cast<int64_t>((y * w)) + x))) * cast<int64_t>(4ULL)));
if ((cast<uint32_t>((addr + cast<uint32_t>(4ULL))) <= cast<uint32_t>(len(m->gs->vram)))) {
pix[cast<int64_t>((o + cast<int64_t>(0ULL)))] = m->gs->vram[cast<uint32_t>((addr + cast<uint32_t>(0ULL)))];
pix[cast<int64_t>((o + cast<int64_t>(1ULL)))] = m->gs->vram[cast<uint32_t>((addr + cast<uint32_t>(1ULL)))];
pix[cast<int64_t>((o + cast<int64_t>(2ULL)))] = m->gs->vram[cast<uint32_t>((addr + cast<uint32_t>(2ULL)))];
pix[cast<int64_t>((o + cast<int64_t>(3ULL)))] = cast<uint8_t>(255ULL);
}
}
}}
}return {pix,w,h};
}
}
// tools/platform/ps2/gs.go:880:1
std::tuple<uint32_t,uint32_t,uint32_t,int64_t,int64_t,bool> ps2_Machine_GSScanout(ps2_Machine* m){
uint32_t fbWord{};
uint32_t dbx{};
uint32_t dby{};
int64_t w{};
int64_t h{};
bool ok{};
{
if ((!m->gs)) {
return {cast<uint32_t>(0ULL),cast<uint32_t>(0ULL),cast<uint32_t>(0ULL),cast<int64_t>(0ULL),cast<int64_t>(0ULL),false};
}
uint64_t dispfb = cast<uint64_t>((cast<uint64_t>(get(m->io,cast<uint32_t>(301990032ULL))) | shl<uint64_t>(cast<uint64_t>(get(m->io,cast<uint32_t>(301990036ULL))),cast<int64_t>(32ULL))));
uint64_t display = cast<uint64_t>((cast<uint64_t>(get(m->io,cast<uint32_t>(301990048ULL))) | shl<uint64_t>(cast<uint64_t>(get(m->io,cast<uint32_t>(301990052ULL))),cast<int64_t>(32ULL))));
{
uint32_t pmode = get(m->io,cast<uint32_t>(301989888ULL));
if (((cast<uint32_t>((pmode & cast<uint32_t>(2ULL))) == cast<uint32_t>(0ULL)) && (cast<uint32_t>((pmode & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL)))) {
dispfb = cast<uint64_t>((cast<uint64_t>(get(m->io,cast<uint32_t>(301990000ULL))) | shl<uint64_t>(cast<uint64_t>(get(m->io,cast<uint32_t>(301990004ULL))),cast<int64_t>(32ULL))));
display = cast<uint64_t>((cast<uint64_t>(get(m->io,cast<uint32_t>(301990016ULL))) | shl<uint64_t>(cast<uint64_t>(get(m->io,cast<uint32_t>(301990020ULL))),cast<int64_t>(32ULL))));
}
}
uint32_t fbp = cast<uint32_t>((cast<uint32_t>(dispfb) & cast<uint32_t>(511ULL)));
uint32_t fbw = cast<uint32_t>((cast<uint32_t>(shr<uint64_t>(dispfb,cast<int64_t>(9ULL))) & cast<uint32_t>(63ULL)));
dbx = cast<uint32_t>((cast<uint32_t>(shr<uint64_t>(dispfb,cast<int64_t>(32ULL))) & cast<uint32_t>(2047ULL)));
dby = cast<uint32_t>((cast<uint32_t>(shr<uint64_t>(dispfb,cast<int64_t>(43ULL))) & cast<uint32_t>(2047ULL)));
uint32_t magh = cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(shr<uint64_t>(display,cast<int64_t>(23ULL))) & cast<uint32_t>(15ULL))) + cast<uint32_t>(1ULL)));
uint32_t magv = cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(shr<uint64_t>(display,cast<int64_t>(27ULL))) & cast<uint32_t>(3ULL))) + cast<uint32_t>(1ULL)));
uint32_t dw = divi<uint32_t>((cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(shr<uint64_t>(display,cast<int64_t>(32ULL))) & cast<uint32_t>(4095ULL))) + cast<uint32_t>(1ULL)))),magh);
uint32_t dh = divi<uint32_t>((cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(shr<uint64_t>(display,cast<int64_t>(44ULL))) & cast<uint32_t>(2047ULL))) + cast<uint32_t>(1ULL)))),magv);
if (((m->gsInterlace == cast<uint32_t>(1ULL)) && (m->gsFieldMode == cast<uint32_t>(1ULL)))) {
dh /= cast<uint32_t>(2ULL);
}
if ((((fbw == cast<uint32_t>(0ULL)) || (dw == cast<uint32_t>(0ULL))) || (dh == cast<uint32_t>(0ULL)))) {
return {cast<uint32_t>(0ULL),cast<uint32_t>(0ULL),cast<uint32_t>(0ULL),cast<int64_t>(0ULL),cast<int64_t>(0ULL),false};
}
return {cast<uint32_t>((fbp * cast<uint32_t>(2048ULL))),dbx,dby,cast<int64_t>(dw),cast<int64_t>(dh),true};
}
}
// tools/platform/ps2/gs.go:911:1
std::tuple<Slice<uint8_t>,int64_t> ps2_Machine_GSBuffer(ps2_Machine* m,uint32_t base,uint32_t bw,int64_t h){
Slice<uint8_t> pix{};
int64_t w{};
{
if ((((!m->gs) || (bw == cast<uint32_t>(0ULL))) || (h <= cast<int64_t>(0ULL)))) {
return {{},cast<int64_t>(0ULL)};
}
w = cast<int64_t>((cast<int64_t>(bw) * cast<int64_t>(64ULL)));
pix = Slice<uint8_t>::make(cast<int64_t>((cast<int64_t>((w * h)) * cast<int64_t>(4ULL))));
{int64_t y = cast<int64_t>(0ULL);for (;(y < h);y++){
{int64_t x = cast<int64_t>(0ULL);for (;(x < w);x++){
uint32_t addr = ps2_addrPSMCT32(divi<uint32_t>(base,cast<uint32_t>(64ULL)),bw,cast<uint32_t>(x),cast<uint32_t>(y));
int64_t o = cast<int64_t>(((cast<int64_t>((cast<int64_t>((y * w)) + x))) * cast<int64_t>(4ULL)));
if ((cast<uint32_t>((addr + cast<uint32_t>(4ULL))) <= cast<uint32_t>(len(m->gs->vram)))) {
pix[cast<int64_t>((o + cast<int64_t>(0ULL)))] = m->gs->vram[cast<uint32_t>((addr + cast<uint32_t>(0ULL)))];
pix[cast<int64_t>((o + cast<int64_t>(1ULL)))] = m->gs->vram[cast<uint32_t>((addr + cast<uint32_t>(1ULL)))];
pix[cast<int64_t>((o + cast<int64_t>(2ULL)))] = m->gs->vram[cast<uint32_t>((addr + cast<uint32_t>(2ULL)))];
pix[cast<int64_t>((o + cast<int64_t>(3ULL)))] = cast<uint8_t>(255ULL);
}
}
}}
}return {pix,w};
}
}
// tools/platform/ps2/gs.go:935:1
std::tuple<Slice<uint8_t>,int64_t> ps2_Machine_GSBufferAlpha(ps2_Machine* m,uint32_t base,uint32_t bw,int64_t h){
Slice<uint8_t> pix{};
int64_t w{};
{
if ((((!m->gs) || (bw == cast<uint32_t>(0ULL))) || (h <= cast<int64_t>(0ULL)))) {
return {{},cast<int64_t>(0ULL)};
}
w = cast<int64_t>((cast<int64_t>(bw) * cast<int64_t>(64ULL)));
pix = Slice<uint8_t>::make(cast<int64_t>((cast<int64_t>((w * h)) * cast<int64_t>(4ULL))));
{int64_t y = cast<int64_t>(0ULL);for (;(y < h);y++){
{int64_t x = cast<int64_t>(0ULL);for (;(x < w);x++){
uint32_t addr = ps2_addrPSMCT32(divi<uint32_t>(base,cast<uint32_t>(64ULL)),bw,cast<uint32_t>(x),cast<uint32_t>(y));
int64_t o = cast<int64_t>(((cast<int64_t>((cast<int64_t>((y * w)) + x))) * cast<int64_t>(4ULL)));
if ((cast<uint32_t>((addr + cast<uint32_t>(4ULL))) <= cast<uint32_t>(len(m->gs->vram)))) {
uint8_t a = m->gs->vram[cast<uint32_t>((addr + cast<uint32_t>(3ULL)))];
auto tmp52 = std::make_tuple(a,a,a,cast<uint8_t>(255ULL));
pix[cast<int64_t>((o + cast<int64_t>(0ULL)))] = std::get<0>(tmp52);
pix[cast<int64_t>((o + cast<int64_t>(1ULL)))] = std::get<1>(tmp52);
pix[cast<int64_t>((o + cast<int64_t>(2ULL)))] = std::get<2>(tmp52);
pix[cast<int64_t>((o + cast<int64_t>(3ULL)))] = std::get<3>(tmp52);
}
}
}}
}return {pix,w};
}
}
// tools/platform/ps2/gs.go:956:1
void ps2_Machine_gsVSync(ps2_Machine* m){
{
if ((!m->gs)) {
return ;
}
m->gs->csr |= cast<uint64_t>(8ULL);
m->gs->csr ^= cast<uint64_t>(8192ULL);
}
}
// tools/platform/ps2/gs.go:965:1
Slice<uint8_t> ps2_u64bytes(uint64_t v){
{
return Slice<uint8_t>{cast<uint8_t>(v),cast<uint8_t>(shr<uint64_t>(v,cast<int64_t>(8ULL))),cast<uint8_t>(shr<uint64_t>(v,cast<int64_t>(16ULL))),cast<uint8_t>(shr<uint64_t>(v,cast<int64_t>(24ULL))),cast<uint8_t>(shr<uint64_t>(v,cast<int64_t>(32ULL))),cast<uint8_t>(shr<uint64_t>(v,cast<int64_t>(40ULL))),cast<uint8_t>(shr<uint64_t>(v,cast<int64_t>(48ULL))),cast<uint8_t>(shr<uint64_t>(v,cast<int64_t>(56ULL)))};
}
}
// tools/platform/ps2/gs.go:971:1
uint32_t ps2_le32gs(Slice<uint8_t> b){
{
return cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(b[cast<int64_t>(0ULL)]) | shl<uint32_t>(cast<uint32_t>(b[cast<int64_t>(1ULL)]),cast<int64_t>(8ULL)))) | shl<uint32_t>(cast<uint32_t>(b[cast<int64_t>(2ULL)]),cast<int64_t>(16ULL)))) | shl<uint32_t>(cast<uint32_t>(b[cast<int64_t>(3ULL)]),cast<int64_t>(24ULL))));
}
}
// tools/platform/ps2/gsdraw.go:29:1
Error ps2_writeFile(std::string name,Slice<uint8_t> data){
{
return go_os_WriteFile(name,data,cast<fs_FileMode>(420ULL));
}
}
// tools/platform/ps2/gsdraw.go:75:1
void ps2_GS_pushVertex(ps2_GS* gs,int32_t x,int32_t y,uint32_t z,uint32_t f,bool kick){
{
uint64_t xyoff = gs->reg[cast<int64_t>(24ULL)];
if ((ps2_GS_ctxt(gs) == cast<int64_t>(1ULL))) {
xyoff = gs->reg[cast<int64_t>(25ULL)];
}
uint64_t uv = gs->reg[cast<int64_t>(3ULL)];
uint64_t st = gs->reg[cast<int64_t>(2ULL)];
ps2_gsVertex v = ps2_gsVertex{cast<int32_t>((x - cast<int32_t>(cast<uint32_t>((cast<uint32_t>(xyoff) & cast<uint32_t>(65535ULL)))))),cast<int32_t>((y - cast<int32_t>(cast<uint32_t>((cast<uint32_t>(shr<uint64_t>(xyoff,cast<int64_t>(32ULL))) & cast<uint32_t>(65535ULL)))))),z,cast<uint32_t>(gs->reg[cast<int64_t>(1ULL)]),cast<int32_t>(cast<uint32_t>((cast<uint32_t>(uv) & cast<uint32_t>(16383ULL)))),cast<int32_t>(cast<uint32_t>((cast<uint32_t>(shr<uint64_t>(uv,cast<int64_t>(16ULL))) & cast<uint32_t>(16383ULL)))),cast<float>(go_math_Float32frombits(cast<uint32_t>(st))),cast<float>(go_math_Float32frombits(cast<uint32_t>(shr<uint64_t>(st,cast<int64_t>(32ULL))))),cast<float>(go_math_Float32frombits(gs->q)),f};
if ((gs->vqN < cast<int64_t>(3ULL))) {
gs->vq[gs->vqN] = v;
gs->vqN++;
}
ps2_GS_kick(gs,kick);
}
}
// tools/platform/ps2/gsdraw.go:114:1
uint64_t ps2_GS_prim(ps2_GS* gs){
{
uint64_t p = gs->reg[cast<int64_t>(0ULL)];
if ((cast<uint64_t>((gs->reg[cast<int64_t>(26ULL)] & cast<uint64_t>(1ULL))) == cast<uint64_t>(0ULL))) {
p = cast<uint64_t>((cast<uint64_t>((p & cast<uint64_t>(7ULL))) | (gs->reg[cast<int64_t>(27ULL)] & ~(cast<uint64_t>(7ULL)))));
}
return p;
}
}
// tools/platform/ps2/gsdraw.go:124:1
int64_t ps2_GS_ctxt(ps2_GS* gs){
{
return cast<int64_t>(cast<uint64_t>((shr<uint64_t>(ps2_GS_prim(gs),cast<int64_t>(9ULL)) & cast<uint64_t>(1ULL))));
}
}
// tools/platform/ps2/gsdraw.go:130:1
void ps2_GS_kick(ps2_GS* gs,bool draw){
{
uint64_t p = ps2_GS_prim(gs);
int64_t typ = cast<int64_t>(cast<uint64_t>((p & cast<uint64_t>(7ULL))));
{
switch(typ){
case cast<int64_t>(0ULL):{
if ((gs->vqN >= cast<int64_t>(1ULL))) {
if (draw) {
ps2_GS_drawn(gs,typ);
ps2_GS_point(gs,gs->vq[cast<int64_t>(0ULL)]);
}
gs->vqN = cast<int64_t>(0ULL);
}
break;}
case cast<int64_t>(1ULL):{
if ((gs->vqN >= cast<int64_t>(2ULL))) {
if (draw) {
ps2_GS_drawn(gs,typ);
ps2_GS_line(gs,gs->vq[cast<int64_t>(0ULL)],gs->vq[cast<int64_t>(1ULL)],p);
}
gs->vqN = cast<int64_t>(0ULL);
}
break;}
case cast<int64_t>(2ULL):{
if ((gs->vqN >= cast<int64_t>(2ULL))) {
if (draw) {
ps2_GS_drawn(gs,typ);
ps2_GS_line(gs,gs->vq[cast<int64_t>(0ULL)],gs->vq[cast<int64_t>(1ULL)],p);
}
gs->vq[cast<int64_t>(0ULL)] = gs->vq[cast<int64_t>(1ULL)];
gs->vqN = cast<int64_t>(1ULL);
}
break;}
case cast<int64_t>(3ULL):{
if ((gs->vqN >= cast<int64_t>(3ULL))) {
if (draw) {
ps2_GS_drawn(gs,typ);
ps2_GS_triangle(gs,gs->vq[cast<int64_t>(0ULL)],gs->vq[cast<int64_t>(1ULL)],gs->vq[cast<int64_t>(2ULL)],p);
}
gs->vqN = cast<int64_t>(0ULL);
}
break;}
case cast<int64_t>(4ULL):{
if ((gs->vqN >= cast<int64_t>(3ULL))) {
if (draw) {
ps2_GS_drawn(gs,typ);
ps2_GS_triangle(gs,gs->vq[cast<int64_t>(0ULL)],gs->vq[cast<int64_t>(1ULL)],gs->vq[cast<int64_t>(2ULL)],p);
}
auto tmp53 = std::make_tuple(gs->vq[cast<int64_t>(1ULL)],gs->vq[cast<int64_t>(2ULL)]);
gs->vq[cast<int64_t>(0ULL)] = std::get<0>(tmp53);
gs->vq[cast<int64_t>(1ULL)] = std::get<1>(tmp53);
gs->vqN = cast<int64_t>(2ULL);
}
break;}
case cast<int64_t>(5ULL):{
if ((gs->vqN >= cast<int64_t>(3ULL))) {
if (draw) {
ps2_GS_drawn(gs,typ);
ps2_GS_triangle(gs,gs->vq[cast<int64_t>(0ULL)],gs->vq[cast<int64_t>(1ULL)],gs->vq[cast<int64_t>(2ULL)],p);
}
gs->vq[cast<int64_t>(1ULL)] = gs->vq[cast<int64_t>(2ULL)];
gs->vqN = cast<int64_t>(2ULL);
}
break;}
case cast<int64_t>(6ULL):{
if ((gs->vqN >= cast<int64_t>(2ULL))) {
if (draw) {
ps2_GS_drawn(gs,typ);
ps2_GS_sprite(gs,gs->vq[cast<int64_t>(0ULL)],gs->vq[cast<int64_t>(1ULL)],p);
}
gs->vqN = cast<int64_t>(0ULL);
}
break;}
default:{
gs->vqN = cast<int64_t>(0ULL);
break;}
}}
}
}
// tools/platform/ps2/gsdraw.go:203:1
time_Time ps2_GS_rasterStart(ps2_GS* gs){
{
if ((!gs->m)) {
return time_Time{};
}
return ps2_Machine_profStart(gs->m);
}
}
// tools/platform/ps2/gsdraw.go:210:1
void ps2_GS_rasterEnd(ps2_GS* gs,time_Time t){
{
if (bool(gs->m)) {
ps2_Machine_profEnd(gs->m,cast<int64_t>(2ULL),t);
}
}
}
// tools/platform/ps2/gsdraw.go:217:1
void ps2_GS_drawn(ps2_GS* gs,int64_t typ){
{
gs->primCount[typ]++;
gs->prims++;
gs->curCmd = cast<int64_t>(-1ULL);
if ((bool(gs->m) && bool(gs->m->OnGSPrim))) {
gs->curCmd = gs->m->OnGSPrim(typ,gs->src);
}
bool dump = (bool(gs->m) && (gs->m->GSVertDump > cast<int64_t>(0ULL)));
if (((((!dump) && bool(gs->m)) && (gs->m->GSBigDump > cast<int64_t>(0ULL))) && (gs->vqN >= cast<int64_t>(2ULL)))) {
auto tmp54 = std::make_tuple(gs->vq[cast<int64_t>(0ULL)].x,gs->vq[cast<int64_t>(0ULL)].x,gs->vq[cast<int64_t>(0ULL)].y,gs->vq[cast<int64_t>(0ULL)].y);
int32_t minX = std::get<0>(tmp54);
int32_t maxX = std::get<1>(tmp54);
int32_t minY = std::get<2>(tmp54);
int32_t maxY = std::get<3>(tmp54);
{int64_t i = cast<int64_t>(1ULL);for (;((i < gs->vqN) && (i < cast<int64_t>(3ULL)));i++){
ps2_gsVertex v = gs->vq[i];
if ((v.x < minX)) {
minX = v.x;
}
if ((v.x > maxX)) {
maxX = v.x;
}
if ((v.y < minY)) {
minY = v.y;
}
if ((v.y > maxY)) {
maxY = v.y;
}
}
}if (((cast<int32_t>((maxX - minX)) > cast<int32_t>(16384ULL)) || (cast<int32_t>((maxY - minY)) > cast<int32_t>(16384ULL)))) {
gs->m->GSBigDump--;
go_fmt_Printf(std::string("  BIG (%dx%d px) from %s\012",25),divi<int32_t>((cast<int32_t>((maxX - minX))),cast<int32_t>(16ULL)),divi<int32_t>((cast<int32_t>((maxY - minY))),cast<int32_t>(16ULL)),gs->src);
dump = true;
gs->m->GSVertDump++;
if (((!gs->srcDumped) && bool(gs->srcData))) {
gs->srcDumped = true;
int64_t n = len(gs->srcData);
if ((n > cast<int64_t>(640ULL))) {
n = cast<int64_t>(640ULL);
}
{int64_t o = cast<int64_t>(0ULL);for (;(cast<int64_t>((o + cast<int64_t>(16ULL))) <= n);o += cast<int64_t>(16ULL)){
go_fmt_Printf(std::string("    pkt qw+%-3d %08X %08X %08X %08X\012",36),divi<int64_t>(o,cast<int64_t>(16ULL)),ps2_le32gs(rrBorrow(gs->srcData,o,len(gs->srcData))),ps2_le32gs(rrBorrow(gs->srcData,cast<int64_t>((o + cast<int64_t>(4ULL))),len(gs->srcData))),ps2_le32gs(rrBorrow(gs->srcData,cast<int64_t>((o + cast<int64_t>(8ULL))),len(gs->srcData))),ps2_le32gs(rrBorrow(gs->srcData,cast<int64_t>((o + cast<int64_t>(12ULL))),len(gs->srcData))));
}
}n = len(gs->srcIn);
if ((n > cast<int64_t>(384ULL))) {
n = cast<int64_t>(384ULL);
}
{int64_t o = cast<int64_t>(0ULL);for (;(cast<int64_t>((o + cast<int64_t>(16ULL))) <= n);o += cast<int64_t>(16ULL)){
go_fmt_Printf(std::string("    in  qw+%-3d %08X %08X %08X %08X\012",36),divi<int64_t>(o,cast<int64_t>(16ULL)),ps2_le32gs(rrBorrow(gs->srcIn,o,len(gs->srcIn))),ps2_le32gs(rrBorrow(gs->srcIn,cast<int64_t>((o + cast<int64_t>(4ULL))),len(gs->srcIn))),ps2_le32gs(rrBorrow(gs->srcIn,cast<int64_t>((o + cast<int64_t>(8ULL))),len(gs->srcIn))),ps2_le32gs(rrBorrow(gs->srcIn,cast<int64_t>((o + cast<int64_t>(12ULL))),len(gs->srcIn))));
}
}if (bool(gs->srcMicro)) {
(void)(ps2_writeFile(std::string("bigkick-micro.bin",17),gs->srcMicro));
(void)(ps2_writeFile(std::string("bigkick-data.bin",16),gs->srcVUData));
go_fmt_Printf(std::string("    wrote bigkick-micro.bin / bigkick-data.bin\012",47));
}
}
}
}
if (dump) {
gs->m->GSVertDump--;
uint64_t p = ps2_GS_prim(gs);
uint64_t xyoff = gs->reg[cast<int64_t>(24ULL)];
uint64_t scis = gs->reg[cast<int64_t>(64ULL)];
if ((ps2_GS_ctxt(gs) == cast<int64_t>(1ULL))) {
xyoff = gs->reg[cast<int64_t>(25ULL)];
scis = gs->reg[cast<int64_t>(65ULL)];
}
uint64_t tex0 = gs->reg[cast<int64_t>(6ULL)];
uint64_t frame = gs->reg[cast<int64_t>(76ULL)];
uint64_t alpha = gs->reg[cast<int64_t>(66ULL)];
uint64_t test = gs->reg[cast<int64_t>(71ULL)];
uint64_t zbuf = gs->reg[cast<int64_t>(78ULL)];
if ((ps2_GS_ctxt(gs) == cast<int64_t>(1ULL))) {
tex0 = gs->reg[cast<int64_t>(7ULL)];
frame = gs->reg[cast<int64_t>(77ULL)];
alpha = gs->reg[cast<int64_t>(67ULL)];
test = gs->reg[cast<int64_t>(72ULL)];
zbuf = gs->reg[cast<int64_t>(79ULL)];
}
std::string texS = std::string("",0);
if ((cast<uint64_t>((p & cast<uint64_t>(16ULL))) != cast<uint64_t>(0ULL))) {
texS = ps2_sprintf(std::string(" tex0 %016X",11),tex0);
}
if ((cast<uint64_t>((p & cast<uint64_t>(64ULL))) != cast<uint64_t>(0ULL))) {
texS += ps2_sprintf(std::string(" alpha %010X",12),alpha);
}
go_fmt_Printf(std::string("  prim %-9s PRIM=0x%03X ctx%d%s%s%s fb 0x%05X psm 0x%02X fbmsk %08X test %05X zbuf %09X xyoff (%d,%d) scissor x %d..%d y %d..%d%s from %s:\012",139),ps2_primNames[cast<int64_t>((typ & cast<int64_t>(7ULL)))],p,cast<int64_t>((ps2_GS_ctxt(gs) + cast<int64_t>(1ULL))),get(Map<bool,std::string>{{true,std::string(" TME",4)},{false,std::string("",0)}},(cast<uint64_t>((p & cast<uint64_t>(16ULL))) != cast<uint64_t>(0ULL))),get(Map<bool,std::string>{{true,std::string(" ABE",4)},{false,std::string("",0)}},(cast<uint64_t>((p & cast<uint64_t>(64ULL))) != cast<uint64_t>(0ULL))),get(Map<bool,std::string>{{true,std::string(" FGE",4)},{false,std::string("",0)}},(cast<uint64_t>((p & cast<uint64_t>(32ULL))) != cast<uint64_t>(0ULL))),cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(frame) & cast<uint32_t>(511ULL))) * cast<uint32_t>(2048ULL))),cast<uint32_t>((cast<uint32_t>(shr<uint64_t>(frame,cast<int64_t>(24ULL))) & cast<uint32_t>(63ULL))),cast<uint32_t>(shr<uint64_t>(frame,cast<int64_t>(32ULL))),cast<uint32_t>((cast<uint32_t>(test) & cast<uint32_t>(524287ULL))),cast<uint64_t>((cast<uint64_t>(zbuf) & cast<uint64_t>(8589934591ULL))),shr<uint32_t>(cast<uint32_t>((cast<uint32_t>(xyoff) & cast<uint32_t>(65535ULL))),cast<int64_t>(4ULL)),shr<uint32_t>(cast<uint32_t>((cast<uint32_t>(shr<uint64_t>(xyoff,cast<int64_t>(32ULL))) & cast<uint32_t>(65535ULL))),cast<int64_t>(4ULL)),cast<uint32_t>((cast<uint32_t>(scis) & cast<uint32_t>(2047ULL))),cast<uint32_t>((cast<uint32_t>(shr<uint64_t>(scis,cast<int64_t>(16ULL))) & cast<uint32_t>(2047ULL))),cast<uint32_t>((cast<uint32_t>(shr<uint64_t>(scis,cast<int64_t>(32ULL))) & cast<uint32_t>(2047ULL))),cast<uint32_t>((cast<uint32_t>(shr<uint64_t>(scis,cast<int64_t>(48ULL))) & cast<uint32_t>(2047ULL))),texS,gs->src);
{int64_t i = cast<int64_t>(0ULL);for (;((i < gs->vqN) && (i < cast<int64_t>(3ULL)));i++){
ps2_gsVertex v = gs->vq[i];
go_fmt_Printf(std::string("    v%d xy (%8.2f,%8.2f) z %08X rgba %08X stq (%g, %g, %g) uv (%.1f,%.1f)\012",74),i,(cast<double>(v.x) / cast<double>(16ULL)),(cast<double>(v.y) / cast<double>(16ULL)),v.z,v.rgba,v.s,v.t,v.q,(cast<double>(v.u) / cast<double>(16ULL)),(cast<double>(v.v) / cast<double>(16ULL)));
}
}}
}
}
// tools/platform/ps2/gsdraw.go:341:1
void ps2_GS_count(ps2_GS* gs,std::string what){
{
if ((!gs->drawCensus)) {
gs->drawCensus = Map<std::string,int64_t>{};
}
gs->drawCensus[what]++;
}
}
// tools/platform/ps2/gsdraw.go:376:1
ps2_gsTarget ps2_GS_target(ps2_GS* gs,uint64_t p){
{
uint64_t frame = gs->reg[cast<int64_t>(76ULL)];
uint64_t scis = gs->reg[cast<int64_t>(64ULL)];
uint64_t alpha = gs->reg[cast<int64_t>(66ULL)];
uint64_t test = gs->reg[cast<int64_t>(71ULL)];
uint64_t fba = gs->reg[cast<int64_t>(74ULL)];
uint64_t zbuf = gs->reg[cast<int64_t>(78ULL)];
if ((cast<uint64_t>((shr<uint64_t>(p,cast<int64_t>(9ULL)) & cast<uint64_t>(1ULL))) == cast<uint64_t>(1ULL))) {
frame = gs->reg[cast<int64_t>(77ULL)];
scis = gs->reg[cast<int64_t>(65ULL)];
alpha = gs->reg[cast<int64_t>(67ULL)];
test = gs->reg[cast<int64_t>(72ULL)];
fba = gs->reg[cast<int64_t>(75ULL)];
zbuf = gs->reg[cast<int64_t>(79ULL)];
}
uint32_t ztst = cast<uint32_t>(1ULL);
if ((cast<uint64_t>((shr<uint64_t>(test,cast<int64_t>(16ULL)) & cast<uint64_t>(1ULL))) != cast<uint64_t>(0ULL))) {
ztst = cast<uint32_t>((cast<uint32_t>(shr<uint64_t>(test,cast<int64_t>(17ULL))) & cast<uint32_t>(3ULL)));
}
return ps2_gsTarget{cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(frame) & cast<uint32_t>(511ULL))) * cast<uint32_t>(32ULL))),cast<uint32_t>((cast<uint32_t>(shr<uint64_t>(frame,cast<int64_t>(16ULL))) & cast<uint32_t>(63ULL))),cast<uint32_t>((cast<uint32_t>(shr<uint64_t>(frame,cast<int64_t>(24ULL))) & cast<uint32_t>(63ULL))),cast<uint32_t>(shr<uint64_t>(frame,cast<int64_t>(32ULL))),cast<int32_t>(cast<uint32_t>((cast<uint32_t>(scis) & cast<uint32_t>(2047ULL)))),cast<int32_t>(cast<uint32_t>((cast<uint32_t>(shr<uint64_t>(scis,cast<int64_t>(32ULL))) & cast<uint32_t>(2047ULL)))),cast<int32_t>(cast<uint32_t>((cast<uint32_t>(shr<uint64_t>(scis,cast<int64_t>(16ULL))) & cast<uint32_t>(2047ULL)))),cast<int32_t>(cast<uint32_t>((cast<uint32_t>(shr<uint64_t>(scis,cast<int64_t>(48ULL))) & cast<uint32_t>(2047ULL)))),(cast<uint64_t>((p & cast<uint64_t>(64ULL))) != cast<uint64_t>(0ULL)),alpha,(cast<uint64_t>((gs->reg[cast<int64_t>(73ULL)] & cast<uint64_t>(1ULL))) != cast<uint64_t>(0ULL)),(cast<uint64_t>((gs->reg[cast<int64_t>(70ULL)] & cast<uint64_t>(1ULL))) != cast<uint64_t>(0ULL)),cast<uint32_t>((cast<uint32_t>(fba) & cast<uint32_t>(1ULL))),test,(cast<uint64_t>((p & cast<uint64_t>(32ULL))) != cast<uint64_t>(0ULL)),cast<uint32_t>((cast<uint32_t>(gs->reg[cast<int64_t>(61ULL)]) & cast<uint32_t>(16777215ULL))),cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(zbuf) & cast<uint32_t>(511ULL))) * cast<uint32_t>(32ULL))),cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(shr<uint64_t>(zbuf,cast<int64_t>(24ULL))) & cast<uint32_t>(15ULL))) | cast<uint32_t>(48ULL))),(cast<uint64_t>((shr<uint64_t>(zbuf,cast<int64_t>(32ULL)) & cast<uint64_t>(1ULL))) != cast<uint64_t>(0ULL)),ztst,cast<int64_t>(cast<uint64_t>((p & cast<uint64_t>(7ULL))))};
}
}
// tools/platform/ps2/gsdraw.go:426:1
void ps2_GS_plot(ps2_GS* gs,ps2_gsTarget* t,int32_t x,int32_t y,uint32_t z,uint32_t rgba,ps2_gsStats* st){
{
if (((((((x < t->sx0) || (x > t->sx1)) || (y < t->sy0)) || (y > t->sy1)) || (x < cast<int32_t>(0ULL))) || (y < cast<int32_t>(0ULL)))) {
st->rejScissor++;
return ;
}
bool frameZ = ((t->psm == cast<uint32_t>(48ULL)) || (t->psm == cast<uint32_t>(49ULL)));
if ((((t->psm != cast<uint32_t>(0ULL)) && (t->psm != cast<uint32_t>(1ULL))) && (!frameZ))) {
ps2_GS_count(gs,ps2_sprintf(std::string("DROPPED pixel writes: frame psm 0x%02X at fb 0x%05X",51),t->psm,cast<uint32_t>((t->fbp * cast<uint32_t>(64ULL)))));
return ;
}
uint32_t fbmsk = t->fbmsk;
bool writeZ = (!t->zmsk);
uint32_t srcA = shr<uint32_t>(rgba,cast<int64_t>(24ULL));
if ((cast<uint64_t>((t->test & cast<uint64_t>(1ULL))) != cast<uint64_t>(0ULL))) {
uint64_t atst = cast<uint64_t>((shr<uint64_t>(t->test,cast<int64_t>(1ULL)) & cast<uint64_t>(7ULL)));
uint32_t aref = cast<uint32_t>((cast<uint32_t>(shr<uint64_t>(t->test,cast<int64_t>(4ULL))) & cast<uint32_t>(255ULL)));
bool pass = false;
{
switch(atst){
case cast<uint64_t>(0ULL):{
break;}
case cast<uint64_t>(1ULL):{
pass = true;
break;}
case cast<uint64_t>(2ULL):{
pass = (srcA < aref);
break;}
case cast<uint64_t>(3ULL):{
pass = (srcA <= aref);
break;}
case cast<uint64_t>(4ULL):{
pass = (srcA == aref);
break;}
case cast<uint64_t>(5ULL):{
pass = (srcA >= aref);
break;}
case cast<uint64_t>(6ULL):{
pass = (srcA > aref);
break;}
case cast<uint64_t>(7ULL):{
pass = (srcA != aref);
break;}
}}
if ((!pass)) {
{
switch(cast<uint64_t>((shr<uint64_t>(t->test,cast<int64_t>(12ULL)) & cast<uint64_t>(3ULL)))){
case cast<uint64_t>(0ULL):{
st->rejAlpha++;
return ;
break;}
case cast<uint64_t>(1ULL):{
writeZ = false;
break;}
case cast<uint64_t>(2ULL):{
fbmsk = cast<uint32_t>(4294967295ULL);
break;}
case cast<uint64_t>(3ULL):{
fbmsk |= cast<uint32_t>(4278190080ULL);
writeZ = false;
break;}
}}
}
}
uint32_t zaddr = cast<uint32_t>(4294967295ULL);
if (((t->ztst != cast<uint32_t>(1ULL)) || writeZ)) {
uint32_t zold={};
{
switch(t->zpsm){
case cast<uint32_t>(48ULL):case cast<uint32_t>(49ULL):{
zaddr = ps2_addrPSMZ32(t->zbp,t->fbw,cast<uint32_t>(x),cast<uint32_t>(y));
if ((cast<uint32_t>((zaddr + cast<uint32_t>(4ULL))) > cast<uint32_t>(len(gs->vram)))) {
return ;
}
zold = ps2_le32gs(rrBorrow(gs->vram,zaddr,len(gs->vram)));
if ((t->zpsm == cast<uint32_t>(49ULL))) {
zold &= cast<uint32_t>(16777215ULL);
z &= cast<uint32_t>(16777215ULL);
}
break;}
case cast<uint32_t>(50ULL):case cast<uint32_t>(58ULL):{
zaddr = ps2_addrPSMZ16(t->zbp,t->fbw,cast<uint32_t>(x),cast<uint32_t>(y),(t->zpsm == cast<uint32_t>(58ULL)));
if ((cast<uint32_t>((zaddr + cast<uint32_t>(2ULL))) > cast<uint32_t>(len(gs->vram)))) {
return ;
}
zold = cast<uint32_t>((cast<uint32_t>(gs->vram[zaddr]) | shl<uint32_t>(cast<uint32_t>(gs->vram[cast<uint32_t>((zaddr + cast<uint32_t>(1ULL)))]),cast<int64_t>(8ULL))));
if ((z > cast<uint32_t>(65535ULL))) {
z = cast<uint32_t>(65535ULL);
}
break;}
}}
{
switch(t->ztst){
case cast<uint32_t>(0ULL):{
st->rejZ++;
return ;
break;}
case cast<uint32_t>(2ULL):{
if ((z < zold)) {
st->rejZ++;
return ;
}
break;}
case cast<uint32_t>(3ULL):{
if ((z <= zold)) {
st->rejZ++;
return ;
}
break;}
}}
}
uint32_t addr = ps2_addrPSMCT32(t->fbp,t->fbw,cast<uint32_t>(x),cast<uint32_t>(y));
if (frameZ) {
addr = ps2_addrPSMZ32(t->fbp,t->fbw,cast<uint32_t>(x),cast<uint32_t>(y));
}
if ((cast<uint32_t>((addr + cast<uint32_t>(4ULL))) > cast<uint32_t>(len(gs->vram)))) {
return ;
}
uint32_t old = ps2_le32gs(rrBorrow(gs->vram,addr,len(gs->vram)));
if ((cast<uint64_t>((shr<uint64_t>(t->test,cast<int64_t>(14ULL)) & cast<uint64_t>(1ULL))) != cast<uint64_t>(0ULL))) {
uint32_t datm = cast<uint32_t>((cast<uint32_t>(shr<uint64_t>(t->test,cast<int64_t>(15ULL))) & cast<uint32_t>(1ULL)));
if ((shr<uint32_t>(old,cast<int64_t>(31ULL)) != datm)) {
st->rejDate++;
return ;
}
}
st->plotted++;
uint32_t preBlend = rgba;
bool blended = false;
if ((t->abe && (((!t->pabe) || (cast<uint32_t>((srcA & cast<uint32_t>(128ULL))) != cast<uint32_t>(0ULL)))))) {
int64_t dstA = cast<int64_t>(shr<uint32_t>(old,cast<int64_t>(24ULL)));
if (((t->psm == cast<uint32_t>(1ULL)) || (t->psm == cast<uint32_t>(49ULL)))) {
dstA = cast<int64_t>(128ULL);
}
rgba = ps2_blendPixel(t->alpha,rgba,old,dstA,t->colclamp);
blended = true;
}
if ((t->fba != cast<uint32_t>(0ULL))) {
rgba |= cast<uint32_t>(2147483648ULL);
}
if ((writeZ && (zaddr != cast<uint32_t>(4294967295ULL)))) {
{
switch(t->zpsm){
case cast<uint32_t>(48ULL):case cast<uint32_t>(49ULL):{
gs->vram[cast<uint32_t>((zaddr + cast<uint32_t>(0ULL)))] = cast<uint8_t>(z);
if(rrcapture::trace.active)rrGSWrite(gs,cast<uint32_t>((zaddr + cast<uint32_t>(0ULL))));
gs->vram[cast<uint32_t>((zaddr + cast<uint32_t>(1ULL)))] = cast<uint8_t>(shr<uint32_t>(z,cast<int64_t>(8ULL)));
if(rrcapture::trace.active)rrGSWrite(gs,cast<uint32_t>((zaddr + cast<uint32_t>(1ULL))));
gs->vram[cast<uint32_t>((zaddr + cast<uint32_t>(2ULL)))] = cast<uint8_t>(shr<uint32_t>(z,cast<int64_t>(16ULL)));
if(rrcapture::trace.active)rrGSWrite(gs,cast<uint32_t>((zaddr + cast<uint32_t>(2ULL))));
if ((t->zpsm == cast<uint32_t>(48ULL))) {
gs->vram[cast<uint32_t>((zaddr + cast<uint32_t>(3ULL)))] = cast<uint8_t>(shr<uint32_t>(z,cast<int64_t>(24ULL)));
if(rrcapture::trace.active)rrGSWrite(gs,cast<uint32_t>((zaddr + cast<uint32_t>(3ULL))));
}
break;}
case cast<uint32_t>(50ULL):case cast<uint32_t>(58ULL):{
gs->vram[cast<uint32_t>((zaddr + cast<uint32_t>(0ULL)))] = cast<uint8_t>(z);
if(rrcapture::trace.active)rrGSWrite(gs,cast<uint32_t>((zaddr + cast<uint32_t>(0ULL))));
gs->vram[cast<uint32_t>((zaddr + cast<uint32_t>(1ULL)))] = cast<uint8_t>(shr<uint32_t>(z,cast<int64_t>(8ULL)));
if(rrcapture::trace.active)rrGSWrite(gs,cast<uint32_t>((zaddr + cast<uint32_t>(1ULL))));
break;}
}}
}
if ((fbmsk == cast<uint32_t>(4294967295ULL))) {
return ;
}
uint32_t px = cast<uint32_t>(((rgba & ~(fbmsk)) | cast<uint32_t>((old & fbmsk))));
if ((cast<uint32_t>((px & cast<uint32_t>(16777215ULL))) != cast<uint32_t>(0ULL))) {
st->plotNonBlack[t->primType]++;
}
if ((((bool(gs->m) && (gs->m->GSPixelN > cast<int64_t>(0ULL))) && (x == gs->m->GSPixelX)) && (y == gs->m->GSPixelY))) {
gs->m->GSPixelN--;
std::string blendS = std::string("",0);
if (blended) {
blendS = ps2_sprintf(std::string(" blend(%08X %X->%08X)",21),preBlend,cast<uint32_t>(t->alpha),rgba);
}
go_fmt_Printf(std::string("  pixel (%d,%d) fb 0x%05X psm 0x%X fbmsk %08X <- %08X (src %08X over %08X, %s%s%s) from %s\012",91),x,y,cast<uint32_t>((t->fbp * cast<uint32_t>(64ULL))),t->psm,t->fbmsk,px,rgba,old,ps2_primNames[t->primType],get(Map<bool,std::string>{{true,std::string(" ABE",4)},{false,std::string("",0)}},t->abe),blendS,gs->src);
}
gs->vram[cast<uint32_t>((addr + cast<uint32_t>(0ULL)))] = cast<uint8_t>(px);
if(rrcapture::trace.active)rrGSWrite(gs,cast<uint32_t>((addr + cast<uint32_t>(0ULL))));
gs->vram[cast<uint32_t>((addr + cast<uint32_t>(1ULL)))] = cast<uint8_t>(shr<uint32_t>(px,cast<int64_t>(8ULL)));
if(rrcapture::trace.active)rrGSWrite(gs,cast<uint32_t>((addr + cast<uint32_t>(1ULL))));
gs->vram[cast<uint32_t>((addr + cast<uint32_t>(2ULL)))] = cast<uint8_t>(shr<uint32_t>(px,cast<int64_t>(16ULL)));
if(rrcapture::trace.active)rrGSWrite(gs,cast<uint32_t>((addr + cast<uint32_t>(2ULL))));
gs->vram[cast<uint32_t>((addr + cast<uint32_t>(3ULL)))] = cast<uint8_t>(shr<uint32_t>(px,cast<int64_t>(24ULL)));
if(rrcapture::trace.active)rrGSWrite(gs,cast<uint32_t>((addr + cast<uint32_t>(3ULL))));
if ((bool(gs->m) && bool(gs->m->OnGSPixel))) {
gs->m->OnGSPixel(gs->curCmd,cast<uint32_t>((t->fbp * cast<uint32_t>(64ULL))),x,y);
}
}
}
// tools/platform/ps2/gsdraw.go:612:1
uint32_t ps2_fogPixel(uint32_t rgba,uint32_t f,uint32_t fogcol){
{
uint32_t g = cast<uint32_t>((cast<uint32_t>(255ULL) - f));
uint32_t r = shr<uint32_t>((cast<uint32_t>((cast<uint32_t>((f * (cast<uint32_t>((rgba & cast<uint32_t>(255ULL)))))) + cast<uint32_t>((g * (cast<uint32_t>((fogcol & cast<uint32_t>(255ULL))))))))),cast<int64_t>(8ULL));
uint32_t gr = shr<uint32_t>((cast<uint32_t>((cast<uint32_t>((f * (cast<uint32_t>((shr<uint32_t>(rgba,cast<int64_t>(8ULL)) & cast<uint32_t>(255ULL)))))) + cast<uint32_t>((g * (cast<uint32_t>((shr<uint32_t>(fogcol,cast<int64_t>(8ULL)) & cast<uint32_t>(255ULL))))))))),cast<int64_t>(8ULL));
uint32_t b = shr<uint32_t>((cast<uint32_t>((cast<uint32_t>((f * (cast<uint32_t>((shr<uint32_t>(rgba,cast<int64_t>(16ULL)) & cast<uint32_t>(255ULL)))))) + cast<uint32_t>((g * (cast<uint32_t>((shr<uint32_t>(fogcol,cast<int64_t>(16ULL)) & cast<uint32_t>(255ULL))))))))),cast<int64_t>(8ULL));
return cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((r | shl<uint32_t>(gr,cast<int64_t>(8ULL)))) | shl<uint32_t>(b,cast<int64_t>(16ULL)))) | cast<uint32_t>((rgba & cast<uint32_t>(4278190080ULL)))));
}
}
// tools/platform/ps2/gsdraw.go:621:1
void ps2_GS_point(ps2_GS* gs,ps2_gsVertex v){
{
auto tmp55=defer([&](){ps2_GS_rasterEnd(gs,ps2_GS_rasterStart(gs));});
uint64_t p = ps2_GS_prim(gs);
ps2_gsTarget t = ps2_GS_target(gs,p);
uint32_t rgba = v.rgba;
if (t.fge) {
rgba = ps2_fogPixel(rgba,cast<uint32_t>((v.f & cast<uint32_t>(255ULL))),t.fogcol);
}
ps2_gsStats st={};
ps2_GS_plot(gs,(&t),shr<int32_t>(v.x,cast<int64_t>(4ULL)),shr<int32_t>(v.y,cast<int64_t>(4ULL)),v.z,rgba,(&st));
ps2_GS_mergeStats(gs,(&st));
}
}
// tools/platform/ps2/gsdraw.go:642:1
void ps2_GS_line(ps2_GS* gs,ps2_gsVertex a,ps2_gsVertex b,uint64_t p){
{
auto tmp56=defer([&](){ps2_GS_rasterEnd(gs,ps2_GS_rasterStart(gs));});
ps2_GS_noteFeatures(gs,p);
ps2_gsTarget t = ps2_GS_target(gs,p);
auto samplerStorage = ps2_GS_sampler(gs,p); auto*smp = samplerStorage ? &*samplerStorage : nullptr;
bool fst = (cast<uint64_t>((p & cast<uint64_t>(256ULL))) != cast<uint64_t>(0ULL));
bool gouraud = (cast<uint64_t>((p & cast<uint64_t>(8ULL))) != cast<uint64_t>(0ULL));
auto tmp57 = std::make_tuple(shr<int32_t>(a.x,cast<int64_t>(4ULL)),shr<int32_t>(a.y,cast<int64_t>(4ULL)));
int32_t x0 = std::get<0>(tmp57);
int32_t y0 = std::get<1>(tmp57);
auto tmp58 = std::make_tuple(shr<int32_t>(b.x,cast<int64_t>(4ULL)),shr<int32_t>(b.y,cast<int64_t>(4ULL)));
int32_t x1 = std::get<0>(tmp58);
int32_t y1 = std::get<1>(tmp58);
auto tmp59 = std::make_tuple(cast<int32_t>((x1 - x0)),cast<int32_t>((y1 - y0)));
int32_t dx = std::get<0>(tmp59);
int32_t dy = std::get<1>(tmp59);
int32_t steps = dx;
if ((steps < cast<int32_t>(0ULL))) {
steps = cast<int32_t>(-steps);
}
{
int32_t ady = dy;
if (((ady < cast<int32_t>(0ULL)) && (cast<int32_t>(-ady) > steps))) {
steps = cast<int32_t>(-ady);
}
else if ((ady > steps)) {
steps = ady;
}
}
auto tmp60 = ps2_unpackRGBA(a.rgba);
int64_t ar = std::get<0>(tmp60);
int64_t ag = std::get<1>(tmp60);
int64_t ab = std::get<2>(tmp60);
int64_t aa = std::get<3>(tmp60);
auto tmp61 = ps2_unpackRGBA(b.rgba);
int64_t br = std::get<0>(tmp61);
int64_t bg = std::get<1>(tmp61);
int64_t bbl = std::get<2>(tmp61);
int64_t ba = std::get<3>(tmp61);
ps2_gsStats st={};
{int32_t i = cast<int32_t>(0ULL);for (;(i <= steps);i++){
auto tmp62 = std::make_tuple(cast<int64_t>(i),cast<int64_t>(steps));
int64_t num = std::get<0>(tmp62);
int64_t den = std::get<1>(tmp62);
if ((den == cast<int64_t>(0ULL))) {
den = cast<int64_t>(1ULL);
}
int32_t x = cast<int32_t>((x0 + cast<int32_t>(divi<int64_t>(cast<int64_t>((cast<int64_t>(dx) * num)),den))));
int32_t y = cast<int32_t>((y0 + cast<int32_t>(divi<int64_t>(cast<int64_t>((cast<int64_t>(dy) * num)),den))));
uint32_t z = cast<uint32_t>(cast<int64_t>((cast<int64_t>(a.z) + divi<int64_t>(cast<int64_t>(((cast<int64_t>((cast<int64_t>(b.z) - cast<int64_t>(a.z)))) * num)),den))));
uint32_t rgba = b.rgba;
if (gouraud) {
int64_t r = cast<int64_t>((ar + divi<int64_t>(cast<int64_t>(((cast<int64_t>((br - ar))) * num)),den)));
int64_t g = cast<int64_t>((ag + divi<int64_t>(cast<int64_t>(((cast<int64_t>((bg - ag))) * num)),den)));
int64_t bl = cast<int64_t>((ab + divi<int64_t>(cast<int64_t>(((cast<int64_t>((bbl - ab))) * num)),den)));
int64_t al = cast<int64_t>((aa + divi<int64_t>(cast<int64_t>(((cast<int64_t>((ba - aa))) * num)),den)));
rgba = cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(r) | shl<uint32_t>(cast<uint32_t>(g),cast<int64_t>(8ULL)))) | shl<uint32_t>(cast<uint32_t>(bl),cast<int64_t>(16ULL)))) | shl<uint32_t>(cast<uint32_t>(al),cast<int64_t>(24ULL))));
}
if (bool(smp)) {
int32_t tu={};
int32_t tv={};
if (fst) {
tu = cast<int32_t>(cast<int64_t>((cast<int64_t>(a.u) + divi<int64_t>(cast<int64_t>(((cast<int64_t>((cast<int64_t>(b.u) - cast<int64_t>(a.u)))) * num)),den))));
tv = cast<int32_t>(cast<int64_t>((cast<int64_t>(a.v) + divi<int64_t>(cast<int64_t>(((cast<int64_t>((cast<int64_t>(b.v) - cast<int64_t>(a.v)))) * num)),den))));
}
else {
float f = (cast<float>(num) / cast<float>(den));
float s = (a.s + (((b.s - a.s)) * f));
float tt = (a.t + (((b.t - a.t)) * f));
float q = (a.q + (((b.q - a.q)) * f));
if ((q != cast<float>(0ULL))) {
tu = cast<int32_t>((((s / q) * cast<float>(shl<int32_t>(cast<int32_t>(1ULL),smp->tex.tw))) * cast<float>(16ULL)));
tv = cast<int32_t>((((tt / q) * cast<float>(shl<int32_t>(cast<int32_t>(1ULL),smp->tex.th))) * cast<float>(16ULL)));
}
}
rgba = ps2_gsSampler_combine(smp,ps2_gsSampler_pick(smp,tu,tv),rgba,(&st));
}
if (t.fge) {
uint32_t ff = cast<uint32_t>(cast<int64_t>((cast<int64_t>(cast<uint32_t>((a.f & cast<uint32_t>(255ULL)))) + divi<int64_t>(cast<int64_t>(((cast<int64_t>((cast<int64_t>(cast<uint32_t>((b.f & cast<uint32_t>(255ULL)))) - cast<int64_t>(cast<uint32_t>((a.f & cast<uint32_t>(255ULL))))))) * num)),den))));
rgba = ps2_fogPixel(rgba,ff,t.fogcol);
}
ps2_GS_plot(gs,(&t),x,y,z,rgba,(&st));
}
}ps2_GS_mergeStats(gs,(&st));
}
}
// tools/platform/ps2/gsdraw.go:714:1
int32_t ps2_texAxis(int32_t pc,int32_t p0,int32_t p1,int32_t uv0,int32_t uv1){
{
if ((p1 == p0)) {
return uv0;
}
return cast<int32_t>((uv0 + cast<int32_t>(divi<int64_t>(cast<int64_t>((cast<int64_t>(cast<int32_t>((pc - p0))) * cast<int64_t>(cast<int32_t>((uv1 - uv0))))),cast<int64_t>(cast<int32_t>((p1 - p0)))))));
}
}
// tools/platform/ps2/gsdraw.go:726:1
void ps2_GS_sprite(ps2_GS* gs,ps2_gsVertex a,ps2_gsVertex b,uint64_t p){
{
auto tmp63=defer([&](){ps2_GS_rasterEnd(gs,ps2_GS_rasterStart(gs));});
ps2_GS_noteFeatures(gs,p);
ps2_gsTarget t = ps2_GS_target(gs,p);
auto samplerStorage = ps2_GS_sampler(gs,p); auto*smp = samplerStorage ? &*samplerStorage : nullptr;
bool fst = (cast<uint64_t>((p & cast<uint64_t>(256ULL))) != cast<uint64_t>(0ULL));
auto tmp64 = std::make_tuple(a.u,a.v,b.u,b.v);
int32_t au = std::get<0>(tmp64);
int32_t av = std::get<1>(tmp64);
int32_t bu = std::get<2>(tmp64);
int32_t bv = std::get<3>(tmp64);
if ((bool(smp) && (!fst))) {
if ((a.q != cast<float>(0ULL))) {
au = cast<int32_t>((((a.s / a.q) * cast<float>(shl<int32_t>(cast<int32_t>(1ULL),smp->tex.tw))) * cast<float>(16ULL)));
av = cast<int32_t>((((a.t / a.q) * cast<float>(shl<int32_t>(cast<int32_t>(1ULL),smp->tex.th))) * cast<float>(16ULL)));
}
if ((b.q != cast<float>(0ULL))) {
bu = cast<int32_t>((((b.s / b.q) * cast<float>(shl<int32_t>(cast<int32_t>(1ULL),smp->tex.tw))) * cast<float>(16ULL)));
bv = cast<int32_t>((((b.t / b.q) * cast<float>(shl<int32_t>(cast<int32_t>(1ULL),smp->tex.th))) * cast<float>(16ULL)));
}
}
auto tmp65 = std::make_tuple(shr<int32_t>(a.x,cast<int64_t>(4ULL)),shr<int32_t>(b.x,cast<int64_t>(4ULL)));
int32_t x0 = std::get<0>(tmp65);
int32_t x1 = std::get<1>(tmp65);
auto tmp66 = std::make_tuple(shr<int32_t>(a.y,cast<int64_t>(4ULL)),shr<int32_t>(b.y,cast<int64_t>(4ULL)));
int32_t y0 = std::get<0>(tmp66);
int32_t y1 = std::get<1>(tmp66);
if ((x0 > x1)) {
auto tmp67 = std::make_tuple(x1,x0);
x0 = std::get<0>(tmp67);
x1 = std::get<1>(tmp67);
}
if ((y0 > y1)) {
auto tmp68 = std::make_tuple(y1,y0);
y0 = std::get<0>(tmp68);
y1 = std::get<1>(tmp68);
}
ps2_GS_rasterFill(gs,(&t),smp,x0,x1,y0,y1,[&](int32_t yLo,int32_t yHi,ps2_gsStats* st)->void{
{int32_t y = yLo;for (;(y < yHi);y++){
int32_t v={};
if (bool(smp)) {
v = ps2_texAxis(shl<int32_t>(y,cast<int64_t>(4ULL)),a.y,b.y,av,bv);
}
{int32_t x = x0;for (;(x < x1);x++){
uint32_t rgba = b.rgba;
if (bool(smp)) {
int32_t u = ps2_texAxis(shl<int32_t>(x,cast<int64_t>(4ULL)),a.x,b.x,au,bu);
if ((((bool(gs->m) && (gs->m->GSPixelN > cast<int64_t>(0ULL))) && (x == gs->m->GSPixelX)) && (y == gs->m->GSPixelY))) {
smp->probe = true;
uint32_t texel = ps2_gsSampler_pick(smp,u,v);
smp->probe = false;
int64_t ctxt = cast<int64_t>(cast<uint64_t>((shr<uint64_t>(p,cast<int64_t>(9ULL)) & cast<uint64_t>(1ULL))));
go_fmt_Printf(std::string("  sample (%d,%d) tbp 0x%05X psm 0x%X %dx%d tbw %d wms/wmt %d/%d cbp 0x%X csa %d uv (%.2f,%.2f) texel %08X vtx %08X tex1 %016X\012",126),x,y,cast<uint32_t>((smp->tex.tbp * cast<uint32_t>(64ULL))),smp->tex.psm,smp->w,smp->h,smp->tex.tbw,smp->wms,smp->wmt,smp->tex.cbp,smp->tex.csa,(cast<double>(u) / cast<double>(16ULL)),(cast<double>(v) / cast<double>(16ULL)),texel,rgba,gs->reg[cast<int64_t>((cast<int64_t>(20ULL) + ctxt))]);
}
rgba = ps2_gsSampler_combine(smp,ps2_gsSampler_pick(smp,u,v),rgba,st);
}
if (t.fge) {
rgba = ps2_fogPixel(rgba,cast<uint32_t>((b.f & cast<uint32_t>(255ULL))),t.fogcol);
}
ps2_GS_plot(gs,(&t),x,y,b.z,rgba,st);
}
}}
}}
);
}
}
// tools/platform/ps2/gsdraw.go:800:1
void ps2_GS_triangle(ps2_GS* gs,ps2_gsVertex v0,ps2_gsVertex v1,ps2_gsVertex v2,uint64_t p){
{
auto tmp69=defer([&](){ps2_GS_rasterEnd(gs,ps2_GS_rasterStart(gs));});
ps2_GS_noteFeatures(gs,p);
ps2_gsTarget t = ps2_GS_target(gs,p);
auto samplerStorage = ps2_GS_sampler(gs,p); auto*smp = samplerStorage ? &*samplerStorage : nullptr;
bool fst = (cast<uint64_t>((p & cast<uint64_t>(256ULL))) != cast<uint64_t>(0ULL));
bool gouraud = (cast<uint64_t>((p & cast<uint64_t>(8ULL))) != cast<uint64_t>(0ULL));
int32_t minX = shr<int32_t>(ps2_min3(v0.x,v1.x,v2.x),cast<int64_t>(4ULL));
int32_t maxX = shr<int32_t>((cast<int32_t>((ps2_max3(v0.x,v1.x,v2.x) + cast<int32_t>(15ULL)))),cast<int64_t>(4ULL));
int32_t minY = shr<int32_t>(ps2_min3(v0.y,v1.y,v2.y),cast<int64_t>(4ULL));
int32_t maxY = shr<int32_t>((cast<int32_t>((ps2_max3(v0.y,v1.y,v2.y) + cast<int32_t>(15ULL)))),cast<int64_t>(4ULL));
if ((minX < t.sx0)) {
minX = t.sx0;
}
if ((maxX > cast<int32_t>((t.sx1 + cast<int32_t>(1ULL))))) {
maxX = cast<int32_t>((t.sx1 + cast<int32_t>(1ULL)));
}
if ((minY < t.sy0)) {
minY = t.sy0;
}
if ((maxY > cast<int32_t>((t.sy1 + cast<int32_t>(1ULL))))) {
maxY = cast<int32_t>((t.sy1 + cast<int32_t>(1ULL)));
}
if (((minX >= maxX) || (minY >= maxY))) {
return ;
}
int64_t area = cast<int64_t>((cast<int64_t>((cast<int64_t>(cast<int32_t>((v1.x - v0.x))) * cast<int64_t>(cast<int32_t>((v2.y - v0.y))))) - cast<int64_t>((cast<int64_t>(cast<int32_t>((v1.y - v0.y))) * cast<int64_t>(cast<int32_t>((v2.x - v0.x)))))));
if ((area == cast<int64_t>(0ULL))) {
return ;
}
if ((area < cast<int64_t>(0ULL))) {
auto tmp70 = std::make_tuple(v2,v1);
v1 = std::get<0>(tmp70);
v2 = std::get<1>(tmp70);
area = cast<int64_t>(-area);
}
auto tmp71 = ps2_unpackRGBA(v0.rgba);
int64_t c0r = std::get<0>(tmp71);
int64_t c0g = std::get<1>(tmp71);
int64_t c0b = std::get<2>(tmp71);
int64_t c0a = std::get<3>(tmp71);
auto tmp72 = ps2_unpackRGBA(v1.rgba);
int64_t c1r = std::get<0>(tmp72);
int64_t c1g = std::get<1>(tmp72);
int64_t c1b = std::get<2>(tmp72);
int64_t c1a = std::get<3>(tmp72);
auto tmp73 = ps2_unpackRGBA(v2.rgba);
int64_t c2r = std::get<0>(tmp73);
int64_t c2g = std::get<1>(tmp73);
int64_t c2b = std::get<2>(tmp73);
int64_t c2a = std::get<3>(tmp73);
ps2_GS_rasterFill(gs,(&t),smp,minX,maxX,minY,maxY,[&](int32_t yLo,int32_t yHi,ps2_gsStats* st)->void{
{int32_t y = yLo;for (;(y < yHi);y++){
int32_t py = cast<int32_t>((shl<int32_t>(y,cast<int64_t>(4ULL)) + cast<int32_t>(8ULL)));
{int32_t x = minX;for (;(x < maxX);x++){
int32_t px = cast<int32_t>((shl<int32_t>(x,cast<int64_t>(4ULL)) + cast<int32_t>(8ULL)));
int64_t w0 = cast<int64_t>((cast<int64_t>((cast<int64_t>(cast<int32_t>((v1.x - px))) * cast<int64_t>(cast<int32_t>((v2.y - py))))) - cast<int64_t>((cast<int64_t>(cast<int32_t>((v1.y - py))) * cast<int64_t>(cast<int32_t>((v2.x - px)))))));
int64_t w1 = cast<int64_t>((cast<int64_t>((cast<int64_t>(cast<int32_t>((v2.x - px))) * cast<int64_t>(cast<int32_t>((v0.y - py))))) - cast<int64_t>((cast<int64_t>(cast<int32_t>((v2.y - py))) * cast<int64_t>(cast<int32_t>((v0.x - px)))))));
int64_t w2 = cast<int64_t>((cast<int64_t>((cast<int64_t>(cast<int32_t>((v0.x - px))) * cast<int64_t>(cast<int32_t>((v1.y - py))))) - cast<int64_t>((cast<int64_t>(cast<int32_t>((v0.y - py))) * cast<int64_t>(cast<int32_t>((v1.x - px)))))));
if ((((w0 < cast<int64_t>(0ULL)) || (w1 < cast<int64_t>(0ULL))) || (w2 < cast<int64_t>(0ULL)))) {
continue;
}
uint32_t rgba = v2.rgba;
if (gouraud) {
int64_t r = divi<int64_t>((cast<int64_t>((cast<int64_t>((cast<int64_t>((w0 * c0r)) + cast<int64_t>((w1 * c1r)))) + cast<int64_t>((w2 * c2r))))),area);
int64_t g = divi<int64_t>((cast<int64_t>((cast<int64_t>((cast<int64_t>((w0 * c0g)) + cast<int64_t>((w1 * c1g)))) + cast<int64_t>((w2 * c2g))))),area);
int64_t b = divi<int64_t>((cast<int64_t>((cast<int64_t>((cast<int64_t>((w0 * c0b)) + cast<int64_t>((w1 * c1b)))) + cast<int64_t>((w2 * c2b))))),area);
int64_t a = divi<int64_t>((cast<int64_t>((cast<int64_t>((cast<int64_t>((w0 * c0a)) + cast<int64_t>((w1 * c1a)))) + cast<int64_t>((w2 * c2a))))),area);
rgba = cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(r) | shl<uint32_t>(cast<uint32_t>(g),cast<int64_t>(8ULL)))) | shl<uint32_t>(cast<uint32_t>(b),cast<int64_t>(16ULL)))) | shl<uint32_t>(cast<uint32_t>(a),cast<int64_t>(24ULL))));
}
if (bool(smp)) {
int32_t tu={};
int32_t tv={};
if (fst) {
tu = cast<int32_t>(divi<int64_t>((cast<int64_t>((cast<int64_t>((cast<int64_t>((w0 * cast<int64_t>(v0.u))) + cast<int64_t>((w1 * cast<int64_t>(v1.u))))) + cast<int64_t>((w2 * cast<int64_t>(v2.u)))))),area));
tv = cast<int32_t>(divi<int64_t>((cast<int64_t>((cast<int64_t>((cast<int64_t>((w0 * cast<int64_t>(v0.v))) + cast<int64_t>((w1 * cast<int64_t>(v1.v))))) + cast<int64_t>((w2 * cast<int64_t>(v2.v)))))),area));
}
else {
auto tmp74 = std::make_tuple(cast<float>(w0),cast<float>(w1),cast<float>(w2));
float fw0 = std::get<0>(tmp74);
float fw1 = std::get<1>(tmp74);
float fw2 = std::get<2>(tmp74);
float fa = cast<float>(area);
float s = (((((fw0 * v0.s) + (fw1 * v1.s)) + (fw2 * v2.s))) / fa);
float tt = (((((fw0 * v0.t) + (fw1 * v1.t)) + (fw2 * v2.t))) / fa);
float q = (((((fw0 * v0.q) + (fw1 * v1.q)) + (fw2 * v2.q))) / fa);
if ((q != cast<float>(0ULL))) {
tu = cast<int32_t>((((s / q) * cast<float>(shl<int32_t>(cast<int32_t>(1ULL),smp->tex.tw))) * cast<float>(16ULL)));
tv = cast<int32_t>((((tt / q) * cast<float>(shl<int32_t>(cast<int32_t>(1ULL),smp->tex.th))) * cast<float>(16ULL)));
}
}
if ((((bool(gs->m) && (gs->m->GSPixelN > cast<int64_t>(0ULL))) && (x == gs->m->GSPixelX)) && (y == gs->m->GSPixelY))) {
smp->probe = true;
uint32_t texel = ps2_gsSampler_pick(smp,tu,tv);
smp->probe = false;
int64_t ctxt = cast<int64_t>(cast<uint64_t>((shr<uint64_t>(p,cast<int64_t>(9ULL)) & cast<uint64_t>(1ULL))));
go_fmt_Printf(std::string("  sample (%d,%d) tbp 0x%05X psm 0x%X %dx%d tbw %d wms/wmt %d/%d cbp 0x%X csa %d uv (%.2f,%.2f) texel %08X vtx %08X tex1 %016X miptbp1 %016X miptbp2 %016X\012",154),x,y,cast<uint32_t>((smp->tex.tbp * cast<uint32_t>(64ULL))),smp->tex.psm,smp->w,smp->h,smp->tex.tbw,smp->wms,smp->wmt,smp->tex.cbp,smp->tex.csa,(cast<double>(tu) / cast<double>(16ULL)),(cast<double>(tv) / cast<double>(16ULL)),texel,rgba,gs->reg[cast<int64_t>((cast<int64_t>(20ULL) + ctxt))],gs->reg[cast<int64_t>((cast<int64_t>(52ULL) + ctxt))],gs->reg[cast<int64_t>((cast<int64_t>(54ULL) + ctxt))]);
}
rgba = ps2_gsSampler_combine(smp,ps2_gsSampler_pick(smp,tu,tv),rgba,st);
}
if (t.fge) {
uint32_t f = cast<uint32_t>(divi<int64_t>((cast<int64_t>((cast<int64_t>((cast<int64_t>((w0 * cast<int64_t>(cast<uint32_t>((v0.f & cast<uint32_t>(255ULL)))))) + cast<int64_t>((w1 * cast<int64_t>(cast<uint32_t>((v1.f & cast<uint32_t>(255ULL)))))))) + cast<int64_t>((w2 * cast<int64_t>(cast<uint32_t>((v2.f & cast<uint32_t>(255ULL))))))))),area));
rgba = ps2_fogPixel(rgba,f,t.fogcol);
}
uint32_t z = cast<uint32_t>(divi<int64_t>((cast<int64_t>((cast<int64_t>((cast<int64_t>((w0 * cast<int64_t>(v0.z))) + cast<int64_t>((w1 * cast<int64_t>(v1.z))))) + cast<int64_t>((w2 * cast<int64_t>(v2.z)))))),area));
ps2_GS_plot(gs,(&t),x,y,z,rgba,st);
}
}}
}}
);
}
}
// tools/platform/ps2/gsdraw.go:909:1
void ps2_GS_noteFeatures(ps2_GS* gs,uint64_t p){
{
if ((cast<uint64_t>((p & cast<uint64_t>(16ULL))) != cast<uint64_t>(0ULL))) {
uint64_t tex0 = gs->reg[cast<int64_t>(6ULL)];
if ((cast<uint64_t>((shr<uint64_t>(p,cast<int64_t>(9ULL)) & cast<uint64_t>(1ULL))) == cast<uint64_t>(1ULL))) {
tex0 = gs->reg[cast<int64_t>(7ULL)];
}
ps2_gsTex t = ps2_decodeTEX0(tex0);
ps2_GS_count(gs,ps2_sprintf(std::string("textured PSM 0x%02X",19),t.psm));
ps2_GS_count(gs,ps2_sprintf(std::string("  tex base 0x%05X %dx%d psm 0x%02X tfx %d cbp 0x%05X csa %d cld %d tex0 %016X",77),cast<uint32_t>((t.tbp * cast<uint32_t>(64ULL))),shl<int64_t>(cast<int64_t>(1ULL),t.tw),shl<int64_t>(cast<int64_t>(1ULL),t.th),t.psm,t.tfx,cast<uint32_t>((t.cbp * cast<uint32_t>(64ULL))),t.csa,t.cld,tex0));
}
if ((cast<uint64_t>((p & cast<uint64_t>(64ULL))) != cast<uint64_t>(0ULL))) {
ps2_GS_count(gs,std::string("alpha-blended",13));
}
if ((cast<uint64_t>((p & cast<uint64_t>(32ULL))) != cast<uint64_t>(0ULL))) {
ps2_GS_count(gs,std::string("fogged",6));
}
if ((cast<uint64_t>((p & cast<uint64_t>(512ULL))) != cast<uint64_t>(0ULL))) {
ps2_GS_count(gs,std::string("context 2",9));
}
uint64_t test = gs->reg[cast<int64_t>(71ULL)];
uint64_t frame = gs->reg[cast<int64_t>(76ULL)];
uint64_t zbuf = gs->reg[cast<int64_t>(78ULL)];
if ((cast<uint64_t>((shr<uint64_t>(p,cast<int64_t>(9ULL)) & cast<uint64_t>(1ULL))) == cast<uint64_t>(1ULL))) {
test = gs->reg[cast<int64_t>(72ULL)];
frame = gs->reg[cast<int64_t>(77ULL)];
zbuf = gs->reg[cast<int64_t>(79ULL)];
}
if (((cast<uint64_t>((shr<uint64_t>(test,cast<int64_t>(16ULL)) & cast<uint64_t>(1ULL))) != cast<uint64_t>(0ULL)) && (cast<uint64_t>((shr<uint64_t>(test,cast<int64_t>(17ULL)) & cast<uint64_t>(3ULL))) >= cast<uint64_t>(2ULL)))) {
ps2_GS_count(gs,std::string("depth-tested",12));
}
ps2_GS_count(gs,ps2_sprintf(std::string("  draw target fb 0x%05X z 0x%05X",32),cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(frame) & cast<uint32_t>(511ULL))) * cast<uint32_t>(2048ULL))),cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(zbuf) & cast<uint32_t>(511ULL))) * cast<uint32_t>(2048ULL)))));
}
}
// tools/platform/ps2/gsdraw.go:949:1
std::tuple<int64_t,int64_t,int64_t,int64_t> ps2_unpackRGBA(uint32_t c){
int64_t r{};
int64_t g{};
int64_t b{};
int64_t a{};
{
return {cast<int64_t>(cast<uint32_t>((c & cast<uint32_t>(255ULL)))),cast<int64_t>(cast<uint32_t>((shr<uint32_t>(c,cast<int64_t>(8ULL)) & cast<uint32_t>(255ULL)))),cast<int64_t>(cast<uint32_t>((shr<uint32_t>(c,cast<int64_t>(16ULL)) & cast<uint32_t>(255ULL)))),cast<int64_t>(cast<uint32_t>((shr<uint32_t>(c,cast<int64_t>(24ULL)) & cast<uint32_t>(255ULL))))};
}
}
// tools/platform/ps2/gsdraw.go:953:1
int32_t ps2_min3(int32_t a,int32_t b,int32_t c){
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
// tools/platform/ps2/gsdraw.go:963:1
int32_t ps2_max3(int32_t a,int32_t b,int32_t c){
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
// tools/platform/ps2/gstex.go:60:1
ps2_gsTex ps2_decodeTEX0(uint64_t v){
{
return ps2_gsTex{cast<uint32_t>((cast<uint32_t>(v) & cast<uint32_t>(16383ULL))),cast<uint32_t>((cast<uint32_t>(shr<uint64_t>(v,cast<int64_t>(14ULL))) & cast<uint32_t>(63ULL))),cast<uint32_t>((cast<uint32_t>(shr<uint64_t>(v,cast<int64_t>(20ULL))) & cast<uint32_t>(63ULL))),cast<uint32_t>((cast<uint32_t>(shr<uint64_t>(v,cast<int64_t>(26ULL))) & cast<uint32_t>(15ULL))),cast<uint32_t>((cast<uint32_t>(shr<uint64_t>(v,cast<int64_t>(30ULL))) & cast<uint32_t>(15ULL))),(cast<uint64_t>((shr<uint64_t>(v,cast<int64_t>(34ULL)) & cast<uint64_t>(1ULL))) != cast<uint64_t>(0ULL)),cast<uint32_t>((cast<uint32_t>(shr<uint64_t>(v,cast<int64_t>(35ULL))) & cast<uint32_t>(3ULL))),cast<uint32_t>((cast<uint32_t>(shr<uint64_t>(v,cast<int64_t>(37ULL))) & cast<uint32_t>(16383ULL))),cast<uint32_t>((cast<uint32_t>(shr<uint64_t>(v,cast<int64_t>(51ULL))) & cast<uint32_t>(15ULL))),cast<uint32_t>((cast<uint32_t>(shr<uint64_t>(v,cast<int64_t>(55ULL))) & cast<uint32_t>(1ULL))),cast<uint32_t>((cast<uint32_t>(shr<uint64_t>(v,cast<int64_t>(56ULL))) & cast<uint32_t>(31ULL))),cast<uint32_t>((cast<uint32_t>(shr<uint64_t>(v,cast<int64_t>(61ULL))) & cast<uint32_t>(7ULL)))};
}
}
// tools/platform/ps2/gstex.go:184:1
uint32_t ps2_addrPSMT8(uint32_t bp,uint32_t bw,uint32_t x,uint32_t y){
{
uint32_t pagesPerRow = divi<uint32_t>(bw,cast<uint32_t>(2ULL));
if ((pagesPerRow == cast<uint32_t>(0ULL))) {
pagesPerRow = cast<uint32_t>(1ULL);
}
uint32_t page = cast<uint32_t>((cast<uint32_t>(((divi<uint32_t>(y,cast<uint32_t>(64ULL))) * pagesPerRow)) + divi<uint32_t>(x,cast<uint32_t>(128ULL))));
auto tmp75 = std::make_tuple(modi<uint32_t>(x,cast<uint32_t>(128ULL)),modi<uint32_t>(y,cast<uint32_t>(64ULL)));
uint32_t px = std::get<0>(tmp75);
uint32_t py = std::get<1>(tmp75);
uint8_t block = ps2_blockPSMCT32[divi<uint32_t>(py,cast<uint32_t>(16ULL))][divi<uint32_t>(px,cast<uint32_t>(16ULL))];
auto tmp76 = std::make_tuple(modi<uint32_t>(px,cast<uint32_t>(16ULL)),modi<uint32_t>(py,cast<uint32_t>(16ULL)));
uint32_t bx = std::get<0>(tmp76);
uint32_t by = std::get<1>(tmp76);
uint32_t column = divi<uint32_t>(by,cast<uint32_t>(4ULL));
auto tmp77 = std::make_tuple(bx,modi<uint32_t>(by,cast<uint32_t>(4ULL)));
uint32_t cx = std::get<0>(tmp77);
uint32_t cy = std::get<1>(tmp77);
uint8_t cw = ps2_columnWordT8[cast<uint32_t>((column & cast<uint32_t>(1ULL)))][cast<uint32_t>((cast<uint32_t>((cy * cast<uint32_t>(16ULL))) + cx))];
uint8_t cb = ps2_columnByteT8[cast<uint32_t>((cast<uint32_t>((cy * cast<uint32_t>(16ULL))) + cx))];
return cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((bp * cast<uint32_t>(256ULL))) + cast<uint32_t>((page * cast<uint32_t>(8192ULL))))) + cast<uint32_t>((cast<uint32_t>(block) * cast<uint32_t>(256ULL))))) + cast<uint32_t>((column * cast<uint32_t>(64ULL))))) + cast<uint32_t>((cast<uint32_t>(cw) * cast<uint32_t>(4ULL))))) + cast<uint32_t>(cb)));
}
}
// tools/platform/ps2/gstex.go:202:1
std::tuple<uint32_t,uint32_t> ps2_addrPSMT4(uint32_t bp,uint32_t bw,uint32_t x,uint32_t y){
{
uint32_t pagesPerRow = divi<uint32_t>(bw,cast<uint32_t>(2ULL));
if ((pagesPerRow == cast<uint32_t>(0ULL))) {
pagesPerRow = cast<uint32_t>(1ULL);
}
uint32_t page = cast<uint32_t>((cast<uint32_t>(((divi<uint32_t>(y,cast<uint32_t>(128ULL))) * pagesPerRow)) + divi<uint32_t>(x,cast<uint32_t>(128ULL))));
auto tmp78 = std::make_tuple(modi<uint32_t>(x,cast<uint32_t>(128ULL)),modi<uint32_t>(y,cast<uint32_t>(128ULL)));
uint32_t px = std::get<0>(tmp78);
uint32_t py = std::get<1>(tmp78);
uint8_t block = ps2_blockPSMT4[divi<uint32_t>(py,cast<uint32_t>(16ULL))][divi<uint32_t>(px,cast<uint32_t>(32ULL))];
auto tmp79 = std::make_tuple(modi<uint32_t>(px,cast<uint32_t>(32ULL)),modi<uint32_t>(py,cast<uint32_t>(16ULL)));
uint32_t bx = std::get<0>(tmp79);
uint32_t by = std::get<1>(tmp79);
uint32_t column = divi<uint32_t>(by,cast<uint32_t>(4ULL));
auto tmp80 = std::make_tuple(bx,modi<uint32_t>(by,cast<uint32_t>(4ULL)));
uint32_t cx = std::get<0>(tmp80);
uint32_t cy = std::get<1>(tmp80);
uint8_t cw = ps2_columnWordT4[cast<uint32_t>((column & cast<uint32_t>(1ULL)))][cast<uint32_t>((cast<uint32_t>((cy * cast<uint32_t>(32ULL))) + cx))];
uint8_t cn = ps2_columnNibbleT4[cast<uint32_t>((cast<uint32_t>((cy * cast<uint32_t>(32ULL))) + cx))];
return {cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((bp * cast<uint32_t>(256ULL))) + cast<uint32_t>((page * cast<uint32_t>(8192ULL))))) + cast<uint32_t>((cast<uint32_t>(block) * cast<uint32_t>(256ULL))))) + cast<uint32_t>((column * cast<uint32_t>(64ULL))))) + cast<uint32_t>((cast<uint32_t>(cw) * cast<uint32_t>(4ULL))))) + cast<uint32_t>(shr<uint8_t>(cn,cast<int64_t>(1ULL))))),cast<uint32_t>(cast<uint8_t>((cn & cast<uint8_t>(1ULL))))};
}
}
// tools/platform/ps2/gstex.go:222:1
uint32_t ps2_addrPSMZ32(uint32_t bp,uint32_t bw,uint32_t x,uint32_t y){
{
if ((bw == cast<uint32_t>(0ULL))) {
bw = cast<uint32_t>(1ULL);
}
constexpr int64_t pageW=64ULL;
constexpr int64_t pageH=32ULL;
uint32_t page = cast<uint32_t>((cast<uint32_t>(((divi<uint32_t>(y,cast<uint32_t>(32ULL))) * bw)) + divi<uint32_t>(x,cast<uint32_t>(64ULL))));
auto tmp81 = std::make_tuple(modi<uint32_t>(x,cast<uint32_t>(64ULL)),modi<uint32_t>(y,cast<uint32_t>(32ULL)));
uint32_t px = std::get<0>(tmp81);
uint32_t py = std::get<1>(tmp81);
uint8_t block = cast<uint8_t>((ps2_blockPSMCT32[divi<uint32_t>(py,cast<uint32_t>(8ULL))][divi<uint32_t>(px,cast<uint32_t>(8ULL))] ^ cast<uint8_t>(24ULL)));
auto tmp82 = std::make_tuple(modi<uint32_t>(px,cast<uint32_t>(8ULL)),modi<uint32_t>(py,cast<uint32_t>(8ULL)));
uint32_t cx = std::get<0>(tmp82);
uint32_t cy = std::get<1>(tmp82);
uint8_t col = ps2_columnPSMCT32[cy][cx];
return cast<uint32_t>(((cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((bp * cast<uint32_t>(64ULL))) + cast<uint32_t>((page * cast<uint32_t>(2048ULL))))) + cast<uint32_t>((cast<uint32_t>(block) * cast<uint32_t>(64ULL))))) + cast<uint32_t>(col)))) * cast<uint32_t>(4ULL)));
}
}
// tools/platform/ps2/gstex.go:237:1
uint32_t ps2_addrPSMZ16(uint32_t bp,uint32_t bw,uint32_t x,uint32_t y,bool s){
{
if ((bw == cast<uint32_t>(0ULL))) {
bw = cast<uint32_t>(1ULL);
}
uint32_t page = cast<uint32_t>((cast<uint32_t>(((divi<uint32_t>(y,cast<uint32_t>(64ULL))) * bw)) + divi<uint32_t>(x,cast<uint32_t>(64ULL))));
auto tmp83 = std::make_tuple(modi<uint32_t>(x,cast<uint32_t>(64ULL)),modi<uint32_t>(y,cast<uint32_t>(64ULL)));
uint32_t px = std::get<0>(tmp83);
uint32_t py = std::get<1>(tmp83);
uint8_t block={};
if (s) {
block = cast<uint8_t>((ps2_blockPSMCT16S[divi<uint32_t>(py,cast<uint32_t>(8ULL))][divi<uint32_t>(px,cast<uint32_t>(16ULL))] ^ cast<uint8_t>(24ULL)));
}
else {
block = cast<uint8_t>((ps2_blockPSMCT16[divi<uint32_t>(py,cast<uint32_t>(8ULL))][divi<uint32_t>(px,cast<uint32_t>(16ULL))] ^ cast<uint8_t>(24ULL)));
}
auto tmp84 = std::make_tuple(modi<uint32_t>(px,cast<uint32_t>(16ULL)),modi<uint32_t>(py,cast<uint32_t>(8ULL)));
uint32_t bx = std::get<0>(tmp84);
uint32_t by = std::get<1>(tmp84);
uint32_t column = divi<uint32_t>(by,cast<uint32_t>(2ULL));
auto tmp85 = std::make_tuple(bx,modi<uint32_t>(by,cast<uint32_t>(2ULL)));
uint32_t cx = std::get<0>(tmp85);
uint32_t cy = std::get<1>(tmp85);
uint8_t cw = ps2_columnWordCT16[cast<uint32_t>((cast<uint32_t>((cy * cast<uint32_t>(16ULL))) + cx))];
uint8_t ch = ps2_columnHalfCT16[cast<uint32_t>((cast<uint32_t>((cy * cast<uint32_t>(16ULL))) + cx))];
return cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((bp * cast<uint32_t>(256ULL))) + cast<uint32_t>((page * cast<uint32_t>(8192ULL))))) + cast<uint32_t>((cast<uint32_t>(block) * cast<uint32_t>(256ULL))))) + cast<uint32_t>((column * cast<uint32_t>(64ULL))))) + cast<uint32_t>((cast<uint32_t>(cw) * cast<uint32_t>(4ULL))))) + cast<uint32_t>((cast<uint32_t>(ch) * cast<uint32_t>(2ULL)))));
}
}
// tools/platform/ps2/gstex.go:259:1
uint32_t ps2_addrPSMCT16(uint32_t bp,uint32_t bw,uint32_t x,uint32_t y,bool s){
{
if ((bw == cast<uint32_t>(0ULL))) {
bw = cast<uint32_t>(1ULL);
}
uint32_t page = cast<uint32_t>((cast<uint32_t>(((divi<uint32_t>(y,cast<uint32_t>(64ULL))) * bw)) + divi<uint32_t>(x,cast<uint32_t>(64ULL))));
auto tmp86 = std::make_tuple(modi<uint32_t>(x,cast<uint32_t>(64ULL)),modi<uint32_t>(y,cast<uint32_t>(64ULL)));
uint32_t px = std::get<0>(tmp86);
uint32_t py = std::get<1>(tmp86);
uint8_t block={};
if (s) {
block = ps2_blockPSMCT16S[divi<uint32_t>(py,cast<uint32_t>(8ULL))][divi<uint32_t>(px,cast<uint32_t>(16ULL))];
}
else {
block = ps2_blockPSMCT16[divi<uint32_t>(py,cast<uint32_t>(8ULL))][divi<uint32_t>(px,cast<uint32_t>(16ULL))];
}
auto tmp87 = std::make_tuple(modi<uint32_t>(px,cast<uint32_t>(16ULL)),modi<uint32_t>(py,cast<uint32_t>(8ULL)));
uint32_t bx = std::get<0>(tmp87);
uint32_t by = std::get<1>(tmp87);
uint32_t column = divi<uint32_t>(by,cast<uint32_t>(2ULL));
auto tmp88 = std::make_tuple(bx,modi<uint32_t>(by,cast<uint32_t>(2ULL)));
uint32_t cx = std::get<0>(tmp88);
uint32_t cy = std::get<1>(tmp88);
uint8_t cw = ps2_columnWordCT16[cast<uint32_t>((cast<uint32_t>((cy * cast<uint32_t>(16ULL))) + cx))];
uint8_t ch = ps2_columnHalfCT16[cast<uint32_t>((cast<uint32_t>((cy * cast<uint32_t>(16ULL))) + cx))];
return cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((bp * cast<uint32_t>(256ULL))) + cast<uint32_t>((page * cast<uint32_t>(8192ULL))))) + cast<uint32_t>((cast<uint32_t>(block) * cast<uint32_t>(256ULL))))) + cast<uint32_t>((column * cast<uint32_t>(64ULL))))) + cast<uint32_t>((cast<uint32_t>(cw) * cast<uint32_t>(4ULL))))) + cast<uint32_t>((cast<uint32_t>(ch) * cast<uint32_t>(2ULL)))));
}
}
// tools/platform/ps2/gstex.go:289:1
void ps2_GS_clutLoad(ps2_GS* gs,ps2_gsTex t){
{
{
switch(t.cld){
case cast<uint32_t>(0ULL):{
return ;
break;}
case cast<uint32_t>(1ULL):{
break;}
case cast<uint32_t>(2ULL):{
gs->cbp0 = t.cbp;
break;}
case cast<uint32_t>(3ULL):{
gs->cbp1 = t.cbp;
break;}
case cast<uint32_t>(4ULL):{
if ((t.cbp == gs->cbp0)) {
return ;
}
gs->cbp0 = t.cbp;
break;}
case cast<uint32_t>(5ULL):{
if ((t.cbp == gs->cbp1)) {
return ;
}
gs->cbp1 = t.cbp;
break;}
default:{
return ;
break;}
}}
uint32_t n={};
{
switch(t.psm){
case cast<uint32_t>(19ULL):case cast<uint32_t>(27ULL):{
n = cast<uint32_t>(256ULL);
break;}
case cast<uint32_t>(20ULL):case cast<uint32_t>(36ULL):case cast<uint32_t>(44ULL):{
n = cast<uint32_t>(16ULL);
break;}
default:{
return ;
break;}
}}
uint32_t base = cast<uint32_t>((t.csa * cast<uint32_t>(16ULL)));
auto tmp89=defer([&](){[&]()->void{
uint32_t nz = cast<uint32_t>(0ULL);
{uint32_t i = cast<uint32_t>(0ULL);for (;((i < n) && (cast<uint32_t>((base + i)) < cast<uint32_t>(512ULL)));i++){
if ((cast<uint32_t>((gs->clut[cast<uint32_t>((base + i))] & cast<uint32_t>(16777215ULL))) != cast<uint32_t>(0ULL))) {
nz++;
}
}
}ps2_GS_count(gs,ps2_sprintf(std::string("clut load cld=%d cbp=0x%05X cpsm=0x%02X csa=%d for psm 0x%02X rgb-nonzero %d/%d",79),t.cld,cast<uint32_t>((t.cbp * cast<uint32_t>(64ULL))),t.cpsm,t.csa,t.psm,nz,n));
ps2_Machine_note(gs->m,std::string("GS: clut at 0x%05X loads %08X %08X %08X %08X %08X %08X %08X %08X",64),cast<uint32_t>((t.cbp * cast<uint32_t>(64ULL))),gs->clut[base],gs->clut[cast<uint32_t>((base + cast<uint32_t>(1ULL)))],gs->clut[cast<uint32_t>((base + cast<uint32_t>(2ULL)))],gs->clut[cast<uint32_t>((base + cast<uint32_t>(3ULL)))],gs->clut[cast<uint32_t>((base + cast<uint32_t>(4ULL)))],gs->clut[cast<uint32_t>((base + cast<uint32_t>(5ULL)))],gs->clut[cast<uint32_t>((base + cast<uint32_t>(6ULL)))],gs->clut[cast<uint32_t>((base + cast<uint32_t>(7ULL)))]);
}
();});
{uint32_t i = cast<uint32_t>(0ULL);for (;((i < n) && (cast<uint32_t>((base + i)) < cast<uint32_t>(512ULL)));i++){
uint32_t x={};
uint32_t y={};
if ((t.csm == cast<uint32_t>(0ULL))) {
if ((n == cast<uint32_t>(256ULL))) {
x = cast<uint32_t>((cast<uint32_t>((i & cast<uint32_t>(7ULL))) | shr<uint32_t>(cast<uint32_t>((i & cast<uint32_t>(16ULL))),cast<int64_t>(1ULL))));
y = cast<uint32_t>((cast<uint32_t>((shr<uint32_t>(i,cast<int64_t>(3ULL)) & cast<uint32_t>(1ULL))) | cast<uint32_t>((shr<uint32_t>(i,cast<int64_t>(4ULL)) & cast<uint32_t>(14ULL)))));
}
else {
x = cast<uint32_t>((i & cast<uint32_t>(7ULL)));
y = cast<uint32_t>((shr<uint32_t>(i,cast<int64_t>(3ULL)) & cast<uint32_t>(1ULL)));
}
}
else {
uint64_t texclut = gs->reg[cast<int64_t>(28ULL)];
uint32_t cou = cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(shr<uint64_t>(texclut,cast<int64_t>(6ULL))) & cast<uint32_t>(63ULL))) * cast<uint32_t>(16ULL)));
uint32_t cov = cast<uint32_t>((cast<uint32_t>(shr<uint64_t>(texclut,cast<int64_t>(12ULL))) & cast<uint32_t>(1023ULL)));
x = cast<uint32_t>((cou + i));
y = cov;
}
{
switch(t.cpsm){
case cast<uint32_t>(0ULL):{
uint32_t bw = cast<uint32_t>(1ULL);
if ((t.csm == cast<uint32_t>(1ULL))) {
bw = cast<uint32_t>((cast<uint32_t>(gs->reg[cast<int64_t>(28ULL)]) & cast<uint32_t>(63ULL)));
}
uint32_t a = ps2_addrPSMCT32(t.cbp,bw,x,y);
if ((cast<uint32_t>((a + cast<uint32_t>(4ULL))) <= cast<uint32_t>(len(gs->vram)))) {
gs->clut[cast<uint32_t>((base + i))] = ps2_le32gs(rrBorrow(gs->vram,a,len(gs->vram)));
}
break;}
case cast<uint32_t>(2ULL):case cast<uint32_t>(10ULL):{
uint32_t bw = cast<uint32_t>(1ULL);
if ((t.csm == cast<uint32_t>(1ULL))) {
bw = cast<uint32_t>((cast<uint32_t>(gs->reg[cast<int64_t>(28ULL)]) & cast<uint32_t>(63ULL)));
}
uint32_t a = ps2_addrPSMCT16(t.cbp,bw,x,y,(t.cpsm == cast<uint32_t>(10ULL)));
if ((cast<uint32_t>((a + cast<uint32_t>(2ULL))) <= cast<uint32_t>(len(gs->vram)))) {
gs->clut[cast<uint32_t>((base + i))] = cast<uint32_t>((cast<uint32_t>(gs->vram[a]) | shl<uint32_t>(cast<uint32_t>(gs->vram[cast<uint32_t>((a + cast<uint32_t>(1ULL)))]),cast<int64_t>(8ULL))));
}
break;}
}}
}
}}
}
// tools/platform/ps2/gstex.go:386:1
uint32_t ps2_GS_clutEntry(ps2_GS* gs,ps2_gsTex* t,uint32_t idx){
{
uint32_t slot = cast<uint32_t>(((cast<uint32_t>((cast<uint32_t>((t->csa * cast<uint32_t>(16ULL))) + idx))) & cast<uint32_t>(511ULL)));
uint32_t raw = gs->clut[slot];
if (((t->cpsm == cast<uint32_t>(2ULL)) || (t->cpsm == cast<uint32_t>(10ULL)))) {
return ps2_GS_expand16(gs,raw);
}
return raw;
}
}
// tools/platform/ps2/gstex.go:398:1
uint32_t ps2_GS_expand16(ps2_GS* gs,uint32_t px){
{
uint64_t texa = gs->reg[cast<int64_t>(59ULL)];
uint32_t r = shl<uint32_t>(cast<uint32_t>((px & cast<uint32_t>(31ULL))),cast<int64_t>(3ULL));
uint32_t g = shl<uint32_t>(cast<uint32_t>((shr<uint32_t>(px,cast<int64_t>(5ULL)) & cast<uint32_t>(31ULL))),cast<int64_t>(3ULL));
uint32_t b = shl<uint32_t>(cast<uint32_t>((shr<uint32_t>(px,cast<int64_t>(10ULL)) & cast<uint32_t>(31ULL))),cast<int64_t>(3ULL));
uint32_t a={};
if ((cast<uint32_t>((px & cast<uint32_t>(32768ULL))) != cast<uint32_t>(0ULL))) {
a = cast<uint32_t>((cast<uint32_t>(shr<uint64_t>(texa,cast<int64_t>(32ULL))) & cast<uint32_t>(255ULL)));
}
else if (((cast<uint64_t>((texa & cast<uint64_t>(32768ULL))) != cast<uint64_t>(0ULL)) && (cast<uint32_t>((px & cast<uint32_t>(32767ULL))) == cast<uint32_t>(0ULL)))) {
a = cast<uint32_t>(0ULL);
}
else {
a = cast<uint32_t>((cast<uint32_t>(texa) & cast<uint32_t>(255ULL)));
}
return cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((r | shl<uint32_t>(g,cast<int64_t>(8ULL)))) | shl<uint32_t>(b,cast<int64_t>(16ULL)))) | shl<uint32_t>(a,cast<int64_t>(24ULL))));
}
}
// tools/platform/ps2/gstex.go:440:1
std::optional<ps2_gsSampler> ps2_GS_sampler(ps2_GS* gs,uint64_t p){
{
if ((cast<uint64_t>((p & cast<uint64_t>(16ULL))) == cast<uint64_t>(0ULL))) {
return {};
}
int64_t ctxt = cast<int64_t>(cast<uint64_t>((shr<uint64_t>(p,cast<int64_t>(9ULL)) & cast<uint64_t>(1ULL))));
uint64_t tex0 = gs->reg[cast<int64_t>(6ULL)];
uint64_t clamp = gs->reg[cast<int64_t>(8ULL)];
uint64_t tex1 = gs->reg[cast<int64_t>(20ULL)];
if ((ctxt == cast<int64_t>(1ULL))) {
tex0 = gs->reg[cast<int64_t>(7ULL)];
clamp = gs->reg[cast<int64_t>(9ULL)];
tex1 = gs->reg[cast<int64_t>(21ULL)];
}
ps2_gsTex t = ps2_decodeTEX0(tex0);
{
switch(t.psm){
case cast<uint32_t>(0ULL):case cast<uint32_t>(1ULL):case cast<uint32_t>(2ULL):case cast<uint32_t>(10ULL):case cast<uint32_t>(19ULL):case cast<uint32_t>(20ULL):case cast<uint32_t>(27ULL):case cast<uint32_t>(36ULL):case cast<uint32_t>(44ULL):{
break;}
default:{
ps2_GS_count(gs,ps2_sprintf(std::string("textured PSM 0x%02X (vertex colour \342\200\224 unsampled)",49),t.psm));
return {};
break;}
}}
return ps2_gsSampler{gs,t,shl<int32_t>(cast<int32_t>(1ULL),t.tw),shl<int32_t>(cast<int32_t>(1ULL),t.th),cast<uint32_t>((cast<uint32_t>(clamp) & cast<uint32_t>(3ULL))),cast<uint32_t>((cast<uint32_t>(shr<uint64_t>(clamp,cast<int64_t>(2ULL))) & cast<uint32_t>(3ULL))),cast<int32_t>(cast<uint32_t>((cast<uint32_t>(shr<uint64_t>(clamp,cast<int64_t>(4ULL))) & cast<uint32_t>(1023ULL)))),cast<int32_t>(cast<uint32_t>((cast<uint32_t>(shr<uint64_t>(clamp,cast<int64_t>(14ULL))) & cast<uint32_t>(1023ULL)))),cast<int32_t>(cast<uint32_t>((cast<uint32_t>(shr<uint64_t>(clamp,cast<int64_t>(24ULL))) & cast<uint32_t>(1023ULL)))),cast<int32_t>(cast<uint32_t>((cast<uint32_t>(shr<uint64_t>(clamp,cast<int64_t>(34ULL))) & cast<uint32_t>(1023ULL)))),(cast<uint64_t>((shr<uint64_t>(tex1,cast<int64_t>(5ULL)) & cast<uint64_t>(1ULL))) != cast<uint64_t>(0ULL)),{},{}};
}
}
// tools/platform/ps2/gstex.go:483:1
uint32_t ps2_gsSampler_pick(ps2_gsSampler* s,int32_t u,int32_t v){
{
if (s->linear) {
auto tmp90 = std::make_tuple(cast<int32_t>((u - cast<int32_t>(8ULL))),cast<int32_t>((v - cast<int32_t>(8ULL))));
int32_t us = std::get<0>(tmp90);
int32_t vs = std::get<1>(tmp90);
auto tmp91 = std::make_tuple(shr<int32_t>(us,cast<int64_t>(4ULL)),shr<int32_t>(vs,cast<int64_t>(4ULL)));
int32_t x0 = std::get<0>(tmp91);
int32_t y0 = std::get<1>(tmp91);
auto tmp92 = std::make_tuple(cast<uint32_t>(cast<int32_t>((us & cast<int32_t>(15ULL)))),cast<uint32_t>(cast<int32_t>((vs & cast<int32_t>(15ULL)))));
uint32_t fx = std::get<0>(tmp92);
uint32_t fy = std::get<1>(tmp92);
uint32_t c00 = ps2_gsSampler_at(s,x0,y0);
uint32_t c10 = ps2_gsSampler_at(s,cast<int32_t>((x0 + cast<int32_t>(1ULL))),y0);
uint32_t c01 = ps2_gsSampler_at(s,x0,cast<int32_t>((y0 + cast<int32_t>(1ULL))));
uint32_t c11 = ps2_gsSampler_at(s,cast<int32_t>((x0 + cast<int32_t>(1ULL))),cast<int32_t>((y0 + cast<int32_t>(1ULL))));
auto lerp = [&](uint32_t a,uint32_t b,uint32_t f)->uint32_t{
uint32_t out={};
{int64_t sh = cast<int64_t>(0ULL);for (;(sh < cast<int64_t>(32ULL));sh += cast<int64_t>(8ULL)){
auto tmp93 = std::make_tuple(cast<int32_t>(cast<uint32_t>((shr<uint32_t>(a,sh) & cast<uint32_t>(255ULL)))),cast<int32_t>(cast<uint32_t>((shr<uint32_t>(b,sh) & cast<uint32_t>(255ULL)))));
int32_t av = std::get<0>(tmp93);
int32_t bv = std::get<1>(tmp93);
out |= shl<uint32_t>(cast<uint32_t>((cast<uint32_t>(cast<int32_t>((av + shr<int32_t>(cast<int32_t>(((cast<int32_t>((bv - av))) * cast<int32_t>(f))),cast<int64_t>(4ULL))))) & cast<uint32_t>(255ULL))),sh);
}
}return out;
}
;
uint32_t top = lerp(c00,c10,fx);
uint32_t bot = lerp(c01,c11,fx);
return lerp(top,bot,fy);
}
return ps2_gsSampler_at(s,shr<int32_t>((cast<int32_t>((u - cast<int32_t>(8ULL)))),cast<int64_t>(4ULL)),shr<int32_t>((cast<int32_t>((v - cast<int32_t>(8ULL)))),cast<int64_t>(4ULL)));
}
}
// tools/platform/ps2/gstex.go:509:1
int32_t ps2_wrapTexel(int32_t c,int32_t size,uint32_t mode,int32_t min,int32_t max){
{
{
switch(mode){
case cast<uint32_t>(1ULL):{
if ((c < cast<int32_t>(0ULL))) {
c = cast<int32_t>(0ULL);
}
if ((c >= size)) {
c = cast<int32_t>((size - cast<int32_t>(1ULL)));
}
return c;
break;}
case cast<uint32_t>(2ULL):{
if ((c < min)) {
c = min;
}
if ((c > max)) {
c = max;
}
return c;
break;}
case cast<uint32_t>(3ULL):{
return cast<int32_t>((cast<int32_t>((c & min)) | max));
break;}
default:{
return cast<int32_t>((c & (cast<int32_t>((size - cast<int32_t>(1ULL))))));
break;}
}}
}
}
// tools/platform/ps2/gstex.go:535:1
uint32_t ps2_gsSampler_at(ps2_gsSampler* s,int32_t u,int32_t v){
{
u = ps2_wrapTexel(u,s->w,s->wms,s->minu,s->maxu);
v = ps2_wrapTexel(v,s->h,s->wmt,s->minv,s->maxv);
auto tmp94 = std::make_tuple(cast<uint32_t>(u),cast<uint32_t>(v));
uint32_t x = std::get<0>(tmp94);
uint32_t y = std::get<1>(tmp94);
auto tmp95 = std::make_tuple(s->gs,(&s->tex));
ps2_GS* gs = std::get<0>(tmp95);
ps2_gsTex* t = std::get<1>(tmp95);
{
switch(t->psm){
case cast<uint32_t>(0ULL):{
uint32_t a = ps2_addrPSMCT32(t->tbp,t->tbw,x,y);
if ((cast<uint32_t>((a + cast<uint32_t>(4ULL))) <= cast<uint32_t>(len(gs->vram)))) {
return ps2_le32gs(rrBorrow(gs->vram,a,len(gs->vram)));
}
break;}
case cast<uint32_t>(1ULL):{
uint32_t a = ps2_addrPSMCT32(t->tbp,t->tbw,x,y);
if ((cast<uint32_t>((a + cast<uint32_t>(4ULL))) <= cast<uint32_t>(len(gs->vram)))) {
uint32_t rgb = cast<uint32_t>((ps2_le32gs(rrBorrow(gs->vram,a,len(gs->vram))) & cast<uint32_t>(16777215ULL)));
uint64_t texa = gs->reg[cast<int64_t>(59ULL)];
uint32_t alpha = cast<uint32_t>((cast<uint32_t>(texa) & cast<uint32_t>(255ULL)));
if (((cast<uint64_t>((texa & cast<uint64_t>(32768ULL))) != cast<uint64_t>(0ULL)) && (rgb == cast<uint32_t>(0ULL)))) {
alpha = cast<uint32_t>(0ULL);
}
return cast<uint32_t>((rgb | shl<uint32_t>(alpha,cast<int64_t>(24ULL))));
}
break;}
case cast<uint32_t>(2ULL):case cast<uint32_t>(10ULL):{
uint32_t a = ps2_addrPSMCT16(t->tbp,t->tbw,x,y,(t->psm == cast<uint32_t>(10ULL)));
if ((cast<uint32_t>((a + cast<uint32_t>(2ULL))) <= cast<uint32_t>(len(gs->vram)))) {
return ps2_GS_expand16(gs,cast<uint32_t>((cast<uint32_t>(gs->vram[a]) | shl<uint32_t>(cast<uint32_t>(gs->vram[cast<uint32_t>((a + cast<uint32_t>(1ULL)))]),cast<int64_t>(8ULL)))));
}
break;}
case cast<uint32_t>(19ULL):{
uint32_t a = ps2_addrPSMT8(t->tbp,t->tbw,x,y);
if ((a < cast<uint32_t>(len(gs->vram)))) {
uint32_t idx = cast<uint32_t>(gs->vram[a]);
if (s->probe) {
print(ps2_sprintf(std::string("    T8 probe (%d,%d) addr 0x%X idx %d clut(cbp 0x%X csa %d cpsm 0x%X csm %d) entry %08X\012",88),x,y,a,idx,t->cbp,t->csa,t->cpsm,t->csm,ps2_GS_clutEntry(gs,t,idx)));
}
uint32_t entry = ps2_GS_clutEntry(gs,t,idx);
if ((((cast<uint32_t>((entry & cast<uint32_t>(16777215ULL))) == cast<uint32_t>(0ULL)) && (!s->parallel)) && (gs->t8Dumped < cast<int64_t>(16ULL)))) {
gs->t8Dumped++;
ps2_Machine_note(gs->m,std::string("GS: T8 sample tbp 0x%05X (%d,%d) -> idx %d entry %08X (csa %d)",62),cast<uint32_t>((t->tbp * cast<uint32_t>(64ULL))),x,y,idx,entry,t->csa);
}
return entry;
}
break;}
case cast<uint32_t>(20ULL):{
auto tmp96 = ps2_addrPSMT4(t->tbp,t->tbw,x,y);
uint32_t a = std::get<0>(tmp96);
uint32_t nib = std::get<1>(tmp96);
if ((a < cast<uint32_t>(len(gs->vram)))) {
uint32_t idx = cast<uint32_t>((shr<uint32_t>(cast<uint32_t>(gs->vram[a]),(cast<uint32_t>((cast<uint32_t>(4ULL) * nib)))) & cast<uint32_t>(15ULL)));
if (s->probe) {
print(ps2_sprintf(std::string("    T4 probe (%d,%d) addr 0x%X nib %d idx %d clut(cbp 0x%X csa %d cpsm 0x%X csm %d) entry %08X\012",95),x,y,a,nib,idx,t->cbp,t->csa,t->cpsm,t->csm,ps2_GS_clutEntry(gs,t,idx)));
}
return ps2_GS_clutEntry(gs,t,idx);
}
break;}
case cast<uint32_t>(27ULL):{
uint32_t a = ps2_addrPSMCT32(t->tbp,t->tbw,x,y);
if ((cast<uint32_t>((a + cast<uint32_t>(4ULL))) <= cast<uint32_t>(len(gs->vram)))) {
return ps2_GS_clutEntry(gs,t,cast<uint32_t>(gs->vram[cast<uint32_t>((a + cast<uint32_t>(3ULL)))]));
}
break;}
case cast<uint32_t>(36ULL):{
uint32_t a = ps2_addrPSMCT32(t->tbp,t->tbw,x,y);
if ((cast<uint32_t>((a + cast<uint32_t>(4ULL))) <= cast<uint32_t>(len(gs->vram)))) {
return ps2_GS_clutEntry(gs,t,cast<uint32_t>((cast<uint32_t>(gs->vram[cast<uint32_t>((a + cast<uint32_t>(3ULL)))]) & cast<uint32_t>(15ULL))));
}
break;}
case cast<uint32_t>(44ULL):{
uint32_t a = ps2_addrPSMCT32(t->tbp,t->tbw,x,y);
if ((cast<uint32_t>((a + cast<uint32_t>(4ULL))) <= cast<uint32_t>(len(gs->vram)))) {
return ps2_GS_clutEntry(gs,t,shr<uint32_t>(cast<uint32_t>(gs->vram[cast<uint32_t>((a + cast<uint32_t>(3ULL)))]),cast<int64_t>(4ULL)));
}
break;}
}}
return cast<uint32_t>(0ULL);
}
}
// tools/platform/ps2/gstex.go:616:1
std::tuple<Slice<uint8_t>,int64_t,int64_t> ps2_Machine_GSTexture(ps2_Machine* m,uint64_t tex0){
Slice<uint8_t> pix{};
int64_t w{};
int64_t h{};
{
ps2_GS* gs = ps2_Machine_ensureGS(m);
ps2_gsTex t = ps2_decodeTEX0(tex0);
if ((t.cld == cast<uint32_t>(0ULL))) {
t.cld = cast<uint32_t>(1ULL);
}
ps2_GS_clutLoad(gs,t);
auto samplerStorage = ps2_gsSampler{gs,t,shl<int32_t>(cast<int32_t>(1ULL),t.tw),shl<int32_t>(cast<int32_t>(1ULL),t.th),{},{},{},{},{},{},{},{},{}}; auto*s = &samplerStorage;
auto tmp97 = std::make_tuple(cast<int64_t>(s->w),cast<int64_t>(s->h));
w = std::get<0>(tmp97);
h = std::get<1>(tmp97);
pix = Slice<uint8_t>::make(cast<int64_t>((cast<int64_t>((w * h)) * cast<int64_t>(4ULL))));
{int64_t y = cast<int64_t>(0ULL);for (;(y < h);y++){
{int64_t x = cast<int64_t>(0ULL);for (;(x < w);x++){
uint32_t c = ps2_gsSampler_at(s,cast<int32_t>(x),cast<int32_t>(y));
int64_t o = cast<int64_t>(((cast<int64_t>((cast<int64_t>((y * w)) + x))) * cast<int64_t>(4ULL)));
pix[cast<int64_t>((o + cast<int64_t>(0ULL)))] = cast<uint8_t>(c);
pix[cast<int64_t>((o + cast<int64_t>(1ULL)))] = cast<uint8_t>(shr<uint32_t>(c,cast<int64_t>(8ULL)));
pix[cast<int64_t>((o + cast<int64_t>(2ULL)))] = cast<uint8_t>(shr<uint32_t>(c,cast<int64_t>(16ULL)));
pix[cast<int64_t>((o + cast<int64_t>(3ULL)))] = cast<uint8_t>(shr<uint32_t>(c,cast<int64_t>(24ULL)));
}
}}
}return {pix,w,h};
}
}
// tools/platform/ps2/gstex.go:642:1
uint32_t ps2_gsSampler_combine(ps2_gsSampler* s,uint32_t tex,uint32_t frag,ps2_gsStats* st){
{
if ((cast<uint32_t>((tex & cast<uint32_t>(16777215ULL))) == cast<uint32_t>(0ULL))) {
st->texBlackPSM[cast<uint32_t>((s->tex.psm & cast<uint32_t>(63ULL)))]++;
st->texBlack++;
}
else {
st->texColorPSM[cast<uint32_t>((s->tex.psm & cast<uint32_t>(63ULL)))]++;
st->texColor++;
}
auto tmp98 = ps2_unpackRGBA(tex);
int64_t tr = std::get<0>(tmp98);
int64_t tg = std::get<1>(tmp98);
int64_t tb = std::get<2>(tmp98);
int64_t ta = std::get<3>(tmp98);
auto tmp99 = ps2_unpackRGBA(frag);
int64_t fr = std::get<0>(tmp99);
int64_t fg = std::get<1>(tmp99);
int64_t fb = std::get<2>(tmp99);
int64_t fa = std::get<3>(tmp99);
int64_t r={};
int64_t g={};
int64_t b={};
int64_t a={};
{
switch(s->tex.tfx){
case cast<uint32_t>(0ULL):{
auto tmp100 = std::make_tuple(ps2_clamp255(shr<int64_t>(cast<int64_t>((tr * fr)),cast<int64_t>(7ULL))),ps2_clamp255(shr<int64_t>(cast<int64_t>((tg * fg)),cast<int64_t>(7ULL))),ps2_clamp255(shr<int64_t>(cast<int64_t>((tb * fb)),cast<int64_t>(7ULL))));
r = std::get<0>(tmp100);
g = std::get<1>(tmp100);
b = std::get<2>(tmp100);
if (s->tex.tcc) {
a = ps2_clamp255(shr<int64_t>(cast<int64_t>((ta * fa)),cast<int64_t>(7ULL)));
}
else {
a = fa;
}
break;}
case cast<uint32_t>(1ULL):{
auto tmp101 = std::make_tuple(tr,tg,tb);
r = std::get<0>(tmp101);
g = std::get<1>(tmp101);
b = std::get<2>(tmp101);
if (s->tex.tcc) {
a = ta;
}
else {
a = fa;
}
break;}
case cast<uint32_t>(2ULL):{
auto tmp102 = std::make_tuple(ps2_clamp255(cast<int64_t>((shr<int64_t>(cast<int64_t>((tr * fr)),cast<int64_t>(7ULL)) + fa))),ps2_clamp255(cast<int64_t>((shr<int64_t>(cast<int64_t>((tg * fg)),cast<int64_t>(7ULL)) + fa))),ps2_clamp255(cast<int64_t>((shr<int64_t>(cast<int64_t>((tb * fb)),cast<int64_t>(7ULL)) + fa))));
r = std::get<0>(tmp102);
g = std::get<1>(tmp102);
b = std::get<2>(tmp102);
if (s->tex.tcc) {
a = ps2_clamp255(cast<int64_t>((ta + fa)));
}
else {
a = fa;
}
break;}
case cast<uint32_t>(3ULL):{
auto tmp103 = std::make_tuple(ps2_clamp255(cast<int64_t>((shr<int64_t>(cast<int64_t>((tr * fr)),cast<int64_t>(7ULL)) + fa))),ps2_clamp255(cast<int64_t>((shr<int64_t>(cast<int64_t>((tg * fg)),cast<int64_t>(7ULL)) + fa))),ps2_clamp255(cast<int64_t>((shr<int64_t>(cast<int64_t>((tb * fb)),cast<int64_t>(7ULL)) + fa))));
r = std::get<0>(tmp103);
g = std::get<1>(tmp103);
b = std::get<2>(tmp103);
if (s->tex.tcc) {
a = ta;
}
else {
a = fa;
}
break;}
}}
return cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(r) | shl<uint32_t>(cast<uint32_t>(g),cast<int64_t>(8ULL)))) | shl<uint32_t>(cast<uint32_t>(b),cast<int64_t>(16ULL)))) | shl<uint32_t>(cast<uint32_t>(a),cast<int64_t>(24ULL))));
}
}
// tools/platform/ps2/gstex.go:686:1
int64_t ps2_clamp255(int64_t v){
{
if ((v < cast<int64_t>(0ULL))) {
return cast<int64_t>(0ULL);
}
if ((v > cast<int64_t>(255ULL))) {
return cast<int64_t>(255ULL);
}
return v;
}
}
// tools/platform/ps2/gstex.go:701:1
uint32_t ps2_blendPixel(uint64_t alpha,uint32_t src,uint32_t dst,int64_t dstA,bool colclamp){
{
uint64_t selA = cast<uint64_t>((alpha & cast<uint64_t>(3ULL)));
uint64_t selB = cast<uint64_t>((shr<uint64_t>(alpha,cast<int64_t>(2ULL)) & cast<uint64_t>(3ULL)));
uint64_t selC = cast<uint64_t>((shr<uint64_t>(alpha,cast<int64_t>(4ULL)) & cast<uint64_t>(3ULL)));
uint64_t selD = cast<uint64_t>((shr<uint64_t>(alpha,cast<int64_t>(6ULL)) & cast<uint64_t>(3ULL)));
int64_t fix = cast<int64_t>((cast<int64_t>(shr<uint64_t>(alpha,cast<int64_t>(32ULL))) & cast<int64_t>(255ULL)));
auto tmp104 = ps2_unpackRGBA(src);
int64_t sr = std::get<0>(tmp104);
int64_t sg = std::get<1>(tmp104);
int64_t sb = std::get<2>(tmp104);
int64_t sa = std::get<3>(tmp104);
auto tmp105 = ps2_unpackRGBA(dst);
int64_t dr = std::get<0>(tmp105);
int64_t dg = std::get<1>(tmp105);
int64_t db = std::get<2>(tmp105);
auto pick = [&](uint64_t sel)->std::tuple<int64_t,int64_t,int64_t>{
{
switch(sel){
case cast<uint64_t>(0ULL):{
return {sr,sg,sb};
break;}
case cast<uint64_t>(1ULL):{
return {dr,dg,db};
break;}
}}
return {cast<int64_t>(0ULL),cast<int64_t>(0ULL),cast<int64_t>(0ULL)};
}
;
auto tmp106 = pick(selA);
int64_t ar = std::get<0>(tmp106);
int64_t ag = std::get<1>(tmp106);
int64_t ab = std::get<2>(tmp106);
auto tmp107 = pick(selB);
int64_t br = std::get<0>(tmp107);
int64_t bg = std::get<1>(tmp107);
int64_t bb = std::get<2>(tmp107);
auto tmp108 = pick(selD);
int64_t drr = std::get<0>(tmp108);
int64_t dgg = std::get<1>(tmp108);
int64_t dbb = std::get<2>(tmp108);
int64_t c={};
{
switch(selC){
case cast<uint64_t>(0ULL):{
c = sa;
break;}
case cast<uint64_t>(1ULL):{
c = dstA;
break;}
default:{
c = fix;
break;}
}}
auto out = [&](int64_t a,int64_t b,int64_t d)->int64_t{
int64_t v = cast<int64_t>((shr<int64_t>(cast<int64_t>(((cast<int64_t>((a - b))) * c)),cast<int64_t>(7ULL)) + d));
if (colclamp) {
return ps2_clamp255(v);
}
return cast<int64_t>((v & cast<int64_t>(255ULL)));
}
;
return cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(out(ar,br,drr)) | shl<uint32_t>(cast<uint32_t>(out(ag,bg,dgg)),cast<int64_t>(8ULL)))) | shl<uint32_t>(cast<uint32_t>(out(ab,bb,dbb)),cast<int64_t>(16ULL)))) | shl<uint32_t>(cast<uint32_t>(sa),cast<int64_t>(24ULL))));
}
}
// tools/platform/ps2/idle.go:119:1
ps2_idleSnap ps2_Machine_snapshotEE(ps2_Machine* m){
{
r5900_CPU* c = m->CPU;
return ps2_idleSnap{c->PC,r5900_CPU_NextPC(c),r5900_CPU_BranchAddr(c),c->HI,c->LO,c->HI1,c->LO1,c->SA,c->ACC,c->FCR31,r5900_CPU_InDelaySlot(c),r5900_CPU_PendingDelay(c),c->LLBit,c->R,c->FPR};
}
}
// tools/platform/ps2/idle.go:133:1
bool ps2_Machine_idleStep(ps2_Machine* m){
{
ps2_idleState* d = (&m->idleDet);
if ((!d->armed)) {
if ((m->steps < d->next)) {
return false;
}
d->armed = true;
d->insns = cast<int64_t>(0ULL);
d->stores = m->stores;
d->snap = ps2_Machine_snapshotEE(m);
return false;
}
d->insns++;
if ((d->insns > cast<int64_t>(64ULL))) {
d->armed = false;
d->next = cast<uint64_t>((m->steps + cast<uint64_t>(4096ULL)));
return false;
}
if (((m->CPU->PC != d->snap.PC) || (m->stores != d->stores))) {
return false;
}
if ((ps2_Machine_snapshotEE(m) != d->snap)) {
return false;
}
d->period = d->insns;
d->armed = false;
d->next = cast<uint64_t>((m->steps + cast<uint64_t>(4096ULL)));
return true;
}
}
// tools/platform/ps2/idle.go:169:1
uint64_t ps2_Machine_idleDeadline(ps2_Machine* m,uint64_t vblAcc){
{
uint64_t D = cast<uint64_t>((cast<uint64_t>(1000000ULL) - vblAcc));
{
uint64_t half = cast<uint64_t>(500000ULL);
if ((vblAcc < half)) {
{
uint64_t d = cast<uint64_t>((half - vblAcc));
if ((d < D)) {
D = d;
}
}
}
}
if (r5900_CPU_TimerIRQDeliverable(m->CPU)) {
uint64_t ticks = cast<uint64_t>(cast<uint32_t>((r5900_CPU_Compare(m->CPU) - r5900_CPU_Count(m->CPU))));
if ((ticks == cast<uint64_t>(0ULL))) {
ticks = cast<uint64_t>(4294967296ULL);
}
uint64_t steps = cast<uint64_t>((ticks * cast<uint64_t>(2ULL)));
if ((cast<uint64_t>((r5900_CPU_CountFrac(m->CPU) & cast<uint64_t>(1ULL))) != cast<uint64_t>(0ULL))) {
steps--;
}
if ((steps < D)) {
D = steps;
}
}
if ((m->iopRebootImage != std::string("",0))) {
if ((m->steps >= m->iopRebootAt)) {
return cast<uint64_t>(0ULL);
}
{
uint64_t d = cast<uint64_t>((m->iopRebootAt - m->steps));
if ((d < D)) {
D = d;
}
}
}
return D;
}
}
// tools/platform/ps2/idle.go:216:1
uint64_t ps2_Machine_idleSkip(ps2_Machine* m,uint64_t* vblAcc,uint64_t* iopAcc,uint64_t budget){
{
uint64_t P = cast<uint64_t>(m->idleDet.period);
if ((P == cast<uint64_t>(0ULL))) {
return cast<uint64_t>(0ULL);
}
if (r5900_CPU_InterruptDeliverable(m->CPU)) {
return cast<uint64_t>(0ULL);
}
if ((bool(m->IOP) && ps2_IOP_BusyToEE(m->IOP))) {
return cast<uint64_t>(0ULL);
}
uint64_t D = ps2_Machine_idleDeadline(m,(*vblAcc));
if ((D <= P)) {
return cast<uint64_t>(0ULL);
}
uint64_t N = cast<uint64_t>(((divi<uint64_t>((cast<uint64_t>((D - cast<uint64_t>(1ULL)))),P)) * P));
if ((N > budget)) {
N = cast<uint64_t>(((divi<uint64_t>(budget,P)) * P));
}
if ((N == cast<uint64_t>(0ULL))) {
return cast<uint64_t>(0ULL);
}
ps2_Machine_captureIdlePhases(m,P);
uint64_t done = cast<uint64_t>(0ULL);
bool broke = false;
{;for (;(done < N);){
uint64_t chunk = cast<uint64_t>((N - done));
if (bool(m->IOP)) {
{
uint64_t toIOP = cast<uint64_t>((cast<uint64_t>(8ULL) - (*iopAcc)));
if ((toIOP < chunk)) {
chunk = toIOP;
}
}
}
bool stepsIOP = (bool(m->IOP) && (cast<uint64_t>(((*iopAcc) + chunk)) >= cast<uint64_t>(8ULL)));
if ((!stepsIOP)) {
(*vblAcc) += chunk;
(*iopAcc) += chunk;
r5900_CPU_SkipInstructions(m->CPU,chunk);
m->steps += chunk;
done += chunk;
continue;
}
(*vblAcc) += chunk;
(*iopAcc) = cast<uint64_t>(0ULL);
r5900_CPU_SkipInstructions(m->CPU,cast<uint64_t>((chunk - cast<uint64_t>(1ULL))));
m->steps += cast<uint64_t>((chunk - cast<uint64_t>(1ULL)));
done += chunk;
r5900_CPU_SetThreadRegs(m->CPU,m->idleDet.phase[modi<uint64_t>((cast<uint64_t>((done - cast<uint64_t>(1ULL)))),P)]);
uint64_t gen = m->eeDisturbGen;
ps2_IOP_Step(m->IOP);
if ((m->eeDisturbGen != gen)) {
r5900_CPU_Step(m->CPU);
m->steps++;
broke = true;
break;
}
r5900_CPU_SkipInstructions(m->CPU,cast<uint64_t>(1ULL));
m->steps++;
}
}if ((!broke)) {
r5900_CPU_SetThreadRegs(m->CPU,m->idleDet.phase[modi<uint64_t>(done,P)]);
}
m->idleDet.Skipped += done;
m->idleDet.Hits++;
return done;
}
}
// tools/platform/ps2/idle.go:327:1
void ps2_Machine_captureIdlePhases(ps2_Machine* m,uint64_t P){
{
if ((cast<uint64_t>(len(m->idleDet.phase)) < P)) {
m->idleDet.phase = Slice<r5900_State>::make(P);
}
m->idleDet.phase = sub(m->idleDet.phase,0,P);
r5900_State base = r5900_CPU_Snapshot(m->CPU);
m->idleDet.phase[cast<int64_t>(0ULL)] = base;
{uint64_t i = cast<uint64_t>(1ULL);for (;(i < P);i++){
r5900_CPU_Step(m->CPU);
m->idleDet.phase[i] = r5900_CPU_Snapshot(m->CPU);
}
}r5900_CPU_Restore(m->CPU,base);
}
}
// tools/platform/ps2/idle.go:355:1
void ps2_Machine_SetIdleSkip(ps2_Machine* m,bool on){
{
m->noIdleSkip = (!on);
}
}
// tools/platform/ps2/idle.go:359:1
bool ps2_Machine_IdleSkip(ps2_Machine* m){
{
return (!m->noIdleSkip);
}
}
// tools/platform/ps2/idle.go:363:1
std::tuple<uint64_t,uint64_t> ps2_Machine_IdleStats(ps2_Machine* m){
uint64_t skipped{};
uint64_t hits{};
{
return {m->idleDet.Skipped,m->idleDet.Hits};
}
}
// tools/platform/ps2/intr.go:47:1
void ps2_Machine_addIntcHandler(ps2_Machine* m){
{
auto tmp109 = std::make_tuple(ps2_Machine_arg(m,cast<uint32_t>(0ULL)),ps2_Machine_arg(m,cast<uint32_t>(1ULL)),ps2_Machine_arg(m,cast<uint32_t>(2ULL)),ps2_Machine_arg(m,cast<uint32_t>(3ULL)));
uint32_t cause = std::get<0>(tmp109);
uint32_t addr = std::get<1>(tmp109);
uint32_t arg = std::get<3>(tmp109);
m->intcHandlers = append(m->intcHandlers,Slice<ps2_handler>{ps2_handler{cause,addr,arg,{}}});
ps2_Machine_note(m,std::string("AddIntcHandler cause=%d -> %s",29),cause,ps2_Machine_Sym(m,addr));
ps2_Machine_setRet(m,cast<uint32_t>(len(m->intcHandlers)));
}
}
// tools/platform/ps2/intr.go:54:1
void ps2_Machine_addDmacHandler(ps2_Machine* m){
{
auto tmp110 = std::make_tuple(ps2_Machine_arg(m,cast<uint32_t>(0ULL)),ps2_Machine_arg(m,cast<uint32_t>(1ULL)),ps2_Machine_arg(m,cast<uint32_t>(2ULL)),ps2_Machine_arg(m,cast<uint32_t>(3ULL)));
uint32_t cause = std::get<0>(tmp110);
uint32_t addr = std::get<1>(tmp110);
uint32_t arg = std::get<3>(tmp110);
m->dmacHandlers = append(m->dmacHandlers,Slice<ps2_handler>{ps2_handler{cause,addr,arg,{}}});
ps2_Machine_note(m,std::string("AddDmacHandler channel=%d -> %s",31),cause,ps2_Machine_Sym(m,addr));
if ((cause == cast<uint32_t>(5ULL))) {
m->sifCmdHandler = addr;
}
ps2_Machine_setRet(m,cast<uint32_t>(len(m->dmacHandlers)));
}
}
// tools/platform/ps2/intr.go:79:1
void ps2_Machine_deliverVBlank(ps2_Machine* m){
{
ps2_Machine_profFrame(m);
ps2_Machine_drainVIF1(m);
m->vblanks++;
ps2_Machine_gsVSync(m);
m->intcStat |= cast<uint32_t>(4ULL);
if (bool(m->OnVBlank)) {
m->OnVBlank(m);
}
if (bool(m->IOP)) {
ps2_IOP_vblankTick(m->IOP);
}
if ((m->vsyncFlagPtr != cast<uint32_t>(0ULL))) {
ps2_Machine_Write32(m,m->vsyncFlagPtr,cast<uint32_t>(1ULL));
}
if ((m->vsyncFlag2Ptr != cast<uint32_t>(0ULL))) {
ps2_Machine_Write32(m,m->vsyncFlag2Ptr,cast<uint32_t>(1ULL));
}
{auto&& tmp111 = m->intcHandlers;
for(int64_t tmp112=0;tmp112<len(tmp111);++tmp112){
auto h=tmp111[tmp112];if (((h.cause != cast<uint32_t>(2ULL)) && (h.cause != cast<uint32_t>(3ULL)))) {
continue;
}
if ((cast<uint32_t>((m->intcMask & (shl<uint32_t>(cast<uint32_t>(1ULL),h.cause)))) == cast<uint32_t>(0ULL))) {
continue;
}
ps2_Machine_callGuest(m,h.addr,Slice<uint32_t>{h.arg});
}}
ps2_Machine_preemptIfOutranked(m);
}
}
// tools/platform/ps2/intr.go:142:1
uint32_t ps2_Machine_callGuest(ps2_Machine* m,uint32_t entry,Slice<uint32_t> args){
{
if (((entry == cast<uint32_t>(0ULL)) || m->Halted)) {
return cast<uint32_t>(0ULL);
}
m->eeDisturbGen++;
r5900_State saved = r5900_CPU_Snapshot(m->CPU);
{auto&& tmp113 = args;
for(int64_t tmp114=0;tmp114<len(tmp113);++tmp114){
auto i=tmp114;auto a=tmp113[tmp114];if ((i < cast<int64_t>(4ULL))) {
r5900_CPU_SetReg(m->CPU,cast<uint32_t>(cast<int64_t>((cast<int64_t>(4ULL) + i))),cast<uint64_t>(cast<int64_t>(cast<int32_t>(a))));
}
}}
r5900_CPU_SetReg(m->CPU,cast<uint32_t>(29ULL),cast<uint64_t>((saved.R[cast<int64_t>(29ULL)].Lo - cast<uint64_t>(2048ULL))));
r5900_CPU_SetReg(m->CPU,cast<uint32_t>(31ULL),cast<uint64_t>(251658240ULL));
r5900_CPU_SetPC(m->CPU,cast<uint64_t>(entry));
constexpr int64_t budget=8388608ULL;
uint64_t ret={};
bool done = false;
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(8388608ULL));i++){
uint32_t pc = cast<uint32_t>(m->CPU->PC);
if ((pc == cast<uint32_t>(251658240ULL))) {
ret = r5900_CPU_Reg(m->CPU,cast<uint32_t>(2ULL));
done = true;
break;
}
if ((m->CPU->Halted || m->Halted)) {
break;
}
if (bool(m->OnStep)) {
m->OnStep(m,pc);
}
{
auto tmp115 = ps2_Machine_unhandledException(m,pc);
std::string r = std::get<0>(tmp115);
bool faulted = std::get<1>(tmp115);
if (faulted) {
ps2_Machine_note(m,std::string("callGuest: %s faulted \342\200\224 %s",28),ps2_Machine_Sym(m,entry),r);
break;
}
}
if ((!ps2_Machine_mapped(m,ps2_phys(pc)))) {
ps2_Machine_note(m,std::string("callGuest: %s left mapped memory at 0x%08X",42),ps2_Machine_Sym(m,entry),pc);
break;
}
r5900_CPU_Step(m->CPU);
m->steps++;
}
}if ((!done)) {
ps2_Machine_note(m,std::string("callGuest: %s did not return within its budget",46),ps2_Machine_Sym(m,entry));
}
r5900_CPU_Restore(m->CPU,saved);
return cast<uint32_t>(ret);
}
}
// tools/platform/ps2/intr.go:208:1
void ps2_Machine_raiseINTC(ps2_Machine* m,uint32_t cause){
{
m->intcStat |= shl<uint32_t>(cast<uint32_t>(1ULL),cause);
r5900_CPU_Interrupt(m->CPU,(cast<uint32_t>((m->intcStat & m->intcMask)) != cast<uint32_t>(0ULL)),(cast<uint32_t>((m->dmacStat & m->dmacMask)) != cast<uint32_t>(0ULL)));
}
}
// tools/platform/ps2/iop.go:317:1
void ps2_IOP_ieEvent(ps2_IOP* p,std::string kind,uint32_t addr,uint32_t val,uint32_t ra){
{
if ((!p->OnIntrState)) {
return ;
}
p->OnIntrState(ps2_IOPIntrEvent{p->steps,kind,mips_CPU_CurPC(p->CPU),ra,addr,val,p->intrEnabled,p->callDepth,p->inIntr});
}
}
// tools/platform/ps2/iop.go:336:1
std::string ps2_IOP_IOPInterrupts(ps2_IOP* p){
{
std::string s = go_fmt_Sprintf(std::string("the IOP's interrupts (%d thread switches by THREADMAN on interrupt exit):\012",74),p->switches);
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(64ULL));i++){
if (((p->raised[i] == cast<int64_t>(0ULL)) && (p->delivered[i] == cast<int64_t>(0ULL)))) {
continue;
}
std::string masked = std::string("",0);
if ((cast<uint64_t>((shr<uint64_t>(p->imask,cast<uint64_t>(i)) & cast<uint64_t>(1ULL))) == cast<uint64_t>(0ULL))) {
masked = std::string("  [masked]",10);
}
s += go_fmt_Sprintf(std::string("      %2d  raised %7d  delivered %7d   %s%s\012",44),i,p->raised[i],p->delivered[i],ps2_IOP_Sym(p,p->handlers[i].fn),masked);
}
}return s;
}
}
// tools/platform/ps2/iop.go:353:1
ps2_IOP* ps2_newIOP(ps2_Machine* m,Slice<uint8_t> ram){
{
ps2_IOP* p = arenaNew(ps2_IOP{ram,Slice<uint8_t>::make(cast<int64_t>(1024ULL)),{},m,{},Map<std::string,ps2_IRXExport*>{},Map<std::string,std::string>{},cast<uint32_t>(4096ULL),{},{},Map<uint32_t,std::string>{},{},{},{},true,{},{},{},Map<uint32_t,ps2_iopHeap*>{},{},{},{},{},{},{},{},ps2_sio2PadState{{},{},{},std::array<uint8_t,6>{cast<uint8_t>(255ULL),cast<uint8_t>(255ULL),cast<uint8_t>(255ULL),cast<uint8_t>(255ULL),cast<uint8_t>(255ULL),cast<uint8_t>(255ULL)}},{},{},{},{},{},{},ps2_newSPU2(),{},{},{},{},{},{},Map<std::string,int64_t>{},{},Map<uint32_t,uint32_t>{},Map<uint32_t,int64_t>{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{}});
p->cdvd = ps2_newCDVD(m);
p->CPU = mips_NewCPU(p);
p->CPU->Syscall = [=](auto...args){return ps2_IOP_handleSyscall(p,args...);};
ps2_IOP_registerLibraries(p);
ps2_IOP_installIdle(p);
return p;
}
}
// tools/platform/ps2/iop.go:392:1
uint32_t ps2_iopPhys(uint32_t addr){
{
return cast<uint32_t>((addr & cast<uint32_t>(536870911ULL)));
}
}
// tools/platform/ps2/iop.go:394:1
uint8_t ps2_IOP_Read(ps2_IOP* p,uint32_t addr){
{
uint32_t a = ps2_iopPhys(addr);
{
if ((a < cast<uint32_t>(2097152ULL))){
return p->ram[a];
}
else if (((a >= cast<uint32_t>(528482304ULL)) && (a < cast<uint32_t>(528483328ULL)))){
return p->spr[cast<uint32_t>((a - cast<uint32_t>(528482304ULL)))];
}
else if (ps2_cdvd_contains(p->cdvd,a)){
ps2_IOP_ioTrace(p,a,false);
return ps2_cdvd_read(p->cdvd,a);
}
}
tmp116:;
uint32_t v = ps2_IOP_ioRead(p,(a & ~(cast<uint32_t>(3ULL))));
ps2_IOP_ioTrace(p,(a & ~(cast<uint32_t>(3ULL))),false);
return cast<uint8_t>(shr<uint32_t>(v,(cast<uint32_t>((cast<uint32_t>(8ULL) * (cast<uint32_t>((a & cast<uint32_t>(3ULL)))))))));
}
}
// tools/platform/ps2/iop.go:413:1
void ps2_IOP_Write(ps2_IOP* p,uint32_t addr,uint8_t v){
{
uint32_t a = ps2_iopPhys(addr);
if (((bool(p->OnWrite) && (a >= p->WatchLo)) && (a < p->WatchHi))) {
p->OnWrite(a,cast<uint32_t>(v),mips_CPU_CurPC(p->CPU));
}
{
if ((a < cast<uint32_t>(2097152ULL))){
p->ram[a] = v;
if ((p->schedIsRun != cast<uint32_t>(0ULL))) {
{
auto tmp119=a;
if (tmp119==(cast<uint32_t>((p->schedIsRun + cast<uint32_t>(3ULL))))){
p->isRunLog[modi<int64_t>(p->isRunLogN,cast<int64_t>(40ULL))] = ps2_isRunWrite{mips_CPU_CurPC(p->CPU),ps2_IOP_Read32(p,p->schedIsRun),false};
p->isRunLogN++;
}
else if (tmp119==(cast<uint32_t>((p->schedIsRun + cast<uint32_t>(7ULL))))){
p->isRunLog[modi<int64_t>(p->isRunLogN,cast<int64_t>(40ULL))] = ps2_isRunWrite{mips_CPU_CurPC(p->CPU),ps2_IOP_Read32(p,cast<uint32_t>((p->schedIsRun + cast<uint32_t>(4ULL)))),true};
p->isRunLogN++;
}
}
tmp118:;
}
return ;
}
else if (((a >= cast<uint32_t>(528482304ULL)) && (a < cast<uint32_t>(528483328ULL)))){
p->spr[cast<uint32_t>((a - cast<uint32_t>(528482304ULL)))] = v;
return ;
}
else if (ps2_cdvd_contains(p->cdvd,a)){
ps2_cdvd_write(p->cdvd,a,v);
ps2_IOP_ioTrace(p,a,true);
return ;
}
}
tmp117:;
uint32_t sh = cast<uint32_t>((cast<uint32_t>(8ULL) * (cast<uint32_t>((a & cast<uint32_t>(3ULL))))));
uint32_t w = cast<uint32_t>(((ps2_IOP_ioPeek(p,(a & ~(cast<uint32_t>(3ULL)))) & ~((shl<uint32_t>(cast<uint32_t>(255ULL),sh)))) | shl<uint32_t>(cast<uint32_t>(v),sh)));
ps2_IOP_ioWrite(p,(a & ~(cast<uint32_t>(3ULL))),w);
ps2_IOP_ioTrace(p,(a & ~(cast<uint32_t>(3ULL))),true);
}
}
// tools/platform/ps2/iop.go:458:1
uint32_t ps2_IOP_ioPeek(ps2_IOP* p,uint32_t a){
{
if (((a >= cast<uint32_t>(486539264ULL)) && (a < cast<uint32_t>(486539392ULL)))) {
return ps2_Machine_sbusRead(p->ps2,cast<uint32_t>((a - cast<uint32_t>(486539264ULL))));
}
if (ps2_cdvd_contains(p->cdvd,a)) {
return cast<uint32_t>(ps2_cdvd_peek(p->cdvd,a));
}
{
auto tmp120 = ps2_IOP_dmaRead(p,a);
uint32_t v = std::get<0>(tmp120);
bool ok = std::get<1>(tmp120);
if (ok) {
return v;
}
}
{
auto tmp121 = ps2_IOP_timerPeek(p,a);
uint32_t v = std::get<0>(tmp121);
bool ok = std::get<1>(tmp121);
if (ok) {
return v;
}
}
if (((a >= cast<uint32_t>(529530880ULL)) && (a < cast<uint32_t>(529532928ULL)))) {
return ps2_spu2_read(p->spu,cast<uint32_t>((a - cast<uint32_t>(529530880ULL))));
}
return get(p->io,a);
}
}
// tools/platform/ps2/iop.go:492:1
uint32_t ps2_IOP_Read32(ps2_IOP* p,uint32_t addr){
{
uint32_t a = ps2_iopPhys(addr);
{
if ((cast<uint32_t>((a + cast<uint32_t>(4ULL))) <= cast<uint32_t>(2097152ULL))){
return cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(p->ram[a]) | shl<uint32_t>(cast<uint32_t>(p->ram[cast<uint32_t>((a + cast<uint32_t>(1ULL)))]),cast<int64_t>(8ULL)))) | shl<uint32_t>(cast<uint32_t>(p->ram[cast<uint32_t>((a + cast<uint32_t>(2ULL)))]),cast<int64_t>(16ULL)))) | shl<uint32_t>(cast<uint32_t>(p->ram[cast<uint32_t>((a + cast<uint32_t>(3ULL)))]),cast<int64_t>(24ULL))));
}
else if (((a >= cast<uint32_t>(528482304ULL)) && (cast<uint32_t>((a + cast<uint32_t>(4ULL))) <= cast<uint32_t>(528483328ULL)))){
uint32_t o = cast<uint32_t>((a - cast<uint32_t>(528482304ULL)));
return cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(p->spr[o]) | shl<uint32_t>(cast<uint32_t>(p->spr[cast<uint32_t>((o + cast<uint32_t>(1ULL)))]),cast<int64_t>(8ULL)))) | shl<uint32_t>(cast<uint32_t>(p->spr[cast<uint32_t>((o + cast<uint32_t>(2ULL)))]),cast<int64_t>(16ULL)))) | shl<uint32_t>(cast<uint32_t>(p->spr[cast<uint32_t>((o + cast<uint32_t>(3ULL)))]),cast<int64_t>(24ULL))));
}
}
tmp122:;
return ps2_IOP_ioRead(p,(a & ~(cast<uint32_t>(3ULL))));
}
}
// tools/platform/ps2/iop.go:504:1
void ps2_IOP_Write32(ps2_IOP* p,uint32_t addr,uint32_t v){
{
uint32_t a = ps2_iopPhys(addr);
{
if ((cast<uint32_t>((a + cast<uint32_t>(4ULL))) <= cast<uint32_t>(2097152ULL))){
p->ram[a] = cast<uint8_t>(v);
p->ram[cast<uint32_t>((a + cast<uint32_t>(1ULL)))] = cast<uint8_t>(shr<uint32_t>(v,cast<int64_t>(8ULL)));
p->ram[cast<uint32_t>((a + cast<uint32_t>(2ULL)))] = cast<uint8_t>(shr<uint32_t>(v,cast<int64_t>(16ULL)));
p->ram[cast<uint32_t>((a + cast<uint32_t>(3ULL)))] = cast<uint8_t>(shr<uint32_t>(v,cast<int64_t>(24ULL)));
}
else if (((a >= cast<uint32_t>(528482304ULL)) && (cast<uint32_t>((a + cast<uint32_t>(4ULL))) <= cast<uint32_t>(528483328ULL)))){
uint32_t o = cast<uint32_t>((a - cast<uint32_t>(528482304ULL)));
p->spr[o] = cast<uint8_t>(v);
p->spr[cast<uint32_t>((o + cast<uint32_t>(1ULL)))] = cast<uint8_t>(shr<uint32_t>(v,cast<int64_t>(8ULL)));
p->spr[cast<uint32_t>((o + cast<uint32_t>(2ULL)))] = cast<uint8_t>(shr<uint32_t>(v,cast<int64_t>(16ULL)));
p->spr[cast<uint32_t>((o + cast<uint32_t>(3ULL)))] = cast<uint8_t>(shr<uint32_t>(v,cast<int64_t>(24ULL)));
}
else {
ps2_IOP_ioWrite(p,(a & ~(cast<uint32_t>(3ULL))),v);
}
}
tmp123:;
}
}
// tools/platform/ps2/iop.go:524:1
std::string ps2_IOP_CString(ps2_IOP* p,uint32_t addr){
{
Slice<uint8_t> b={};
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < cast<uint32_t>(1024ULL));i++){
uint8_t c = ps2_IOP_Read(p,cast<uint32_t>((addr + i)));
if ((c == cast<uint8_t>(0ULL))) {
break;
}
b = append(b,Slice<uint8_t>{c});
}
}return cast<std::string>(b);
}
}
// tools/platform/ps2/iop.go:539:1
uint32_t ps2_IOP_ioRead(ps2_IOP* p,uint32_t a){
{
if (((a >= cast<uint32_t>(486539264ULL)) && (a < cast<uint32_t>(486539392ULL)))) {
return ps2_Machine_sbusRead(p->ps2,cast<uint32_t>((a - cast<uint32_t>(486539264ULL))));
}
{
auto tmp124 = ps2_IOP_dmaRead(p,a);
uint32_t v = std::get<0>(tmp124);
bool ok = std::get<1>(tmp124);
if (ok) {
return v;
}
}
{
auto tmp125 = ps2_IOP_timerRead(p,a);
uint32_t v = std::get<0>(tmp125);
bool ok = std::get<1>(tmp125);
if (ok) {
return v;
}
}
if (((a >= cast<uint32_t>(529530880ULL)) && (a < cast<uint32_t>(529532928ULL)))) {
return ps2_spu2_readReg(p->spu,cast<uint32_t>((a - cast<uint32_t>(529530880ULL))),p->steps);
}
if (ps2_sio2Contains(a)) {
return ps2_IOP_sio2Read(p,a);
}
p->unmodelledIO[a]++;
return get(p->io,a);
}
}
// tools/platform/ps2/iop.go:559:1
void ps2_IOP_ioWrite(ps2_IOP* p,uint32_t a,uint32_t v){
{
if (((a >= cast<uint32_t>(486539264ULL)) && (a < cast<uint32_t>(486539392ULL)))) {
ps2_Machine_sbusWriteIOP(p->ps2,cast<uint32_t>((a - cast<uint32_t>(486539264ULL))),v);
return ;
}
if (ps2_IOP_dmaWrite(p,a,v)) {
return ;
}
if (ps2_IOP_timerWrite(p,a,v)) {
return ;
}
if (((a >= cast<uint32_t>(529530880ULL)) && (a < cast<uint32_t>(529532928ULL)))) {
ps2_spu2_writeReg(p->spu,cast<uint32_t>((a - cast<uint32_t>(529530880ULL))),v,p->steps);
return ;
}
if (ps2_sio2Contains(a)) {
ps2_IOP_sio2Write(p,a,v);
return ;
}
p->unmodelledIO[a]++;
p->io[a] = v;
}
}
// tools/platform/ps2/iop.go:590:1
uint32_t ps2_IOP_alloc(ps2_IOP* p,uint32_t n){
{
uint32_t a = ((cast<uint32_t>((p->allocPtr + cast<uint32_t>(63ULL)))) & ~(cast<uint32_t>(63ULL)));
p->allocPtr = cast<uint32_t>((a + n));
uint32_t ceiling = cast<uint32_t>(2031616ULL);
if (((p->allocHighPtr != cast<uint32_t>(0ULL)) && (p->allocHighPtr < ceiling))) {
ceiling = p->allocHighPtr;
}
if ((p->allocPtr > ceiling)) {
ps2_IOP_halt(p,std::string("out of IOP memory: %d bytes wanted, and the low allocations have reached 0x%08X",79),n,ceiling);
return cast<uint32_t>(0ULL);
}
return a;
}
}
// tools/platform/ps2/iop.go:666:1
void ps2_IOP_installIdle(ps2_IOP* p){
{
ps2_IOP_Write32(p,cast<uint32_t>(512ULL),cast<uint32_t>(268500991ULL));
ps2_IOP_Write32(p,cast<uint32_t>(516ULL),ps2_insnNop());
mips_CPU_SetPC(p->CPU,cast<uint32_t>(512ULL));
mips_CPU_SetReg(p->CPU,cast<uint32_t>(29ULL),cast<uint32_t>(2079744ULL));
}
}
// tools/platform/ps2/iop.go:684:1
void ps2_IOP_Step(ps2_IOP* p){
{
if ((((!p->running) || p->Halted) || p->CPU->Halted)) {
return ;
}
ps2_IOP_tick(p);
if (bool(p->logPC)) {
{
auto tmp126 = lookup(p->logPC,cast<uint32_t>(mips_CPU_CurPC(p->CPU)));
bool hit = std::get<1>(tmp126);
if (hit) {
uint32_t pc = cast<uint32_t>(mips_CPU_CurPC(p->CPU));
go_fmt_Printf(std::string("ioplogpc vbl=%d %-24s a0=%08X a1=%08X a2=%08X a3=%08X sp+10=%08X sp+14=%08X ra=%s\012",82),ps2_Machine_VBlanks(p->ps2),ps2_IOP_Sym(p,pc),mips_CPU_Reg(p->CPU,cast<uint32_t>(4ULL)),mips_CPU_Reg(p->CPU,cast<uint32_t>(5ULL)),mips_CPU_Reg(p->CPU,cast<uint32_t>(6ULL)),mips_CPU_Reg(p->CPU,cast<uint32_t>(7ULL)),ps2_IOP_Read32(p,cast<uint32_t>((cast<uint32_t>(mips_CPU_Reg(p->CPU,cast<uint32_t>(29ULL))) + cast<uint32_t>(16ULL)))),ps2_IOP_Read32(p,cast<uint32_t>((cast<uint32_t>(mips_CPU_Reg(p->CPU,cast<uint32_t>(29ULL))) + cast<uint32_t>(20ULL)))),ps2_IOP_Sym(p,cast<uint32_t>(mips_CPU_Reg(p->CPU,cast<uint32_t>(31ULL)))));
}
}
}
mips_CPU_Step(p->CPU);
}
}
// tools/platform/ps2/iop.go:704:1
void ps2_IOP_LogPC(ps2_IOP* p,uint32_t addr){
{
if ((!p->logPC)) {
p->logPC = Map<uint32_t,Anon38>{};
}
p->logPC[addr] = Anon38{};
}
}
// tools/platform/ps2/iop.go:718:1
std::string ps2_IOP_TTY(ps2_IOP* p){
{
return cast<std::string>(p->tty);
}
}
// tools/platform/ps2/iop.go:728:1
std::string ps2_IOP_DisasmAt(ps2_IOP* p,uint32_t addr){
{
Slice<uint8_t> code = Slice<uint8_t>{ps2_IOP_Read(p,addr),ps2_IOP_Read(p,cast<uint32_t>((addr + cast<uint32_t>(1ULL)))),ps2_IOP_Read(p,cast<uint32_t>((addr + cast<uint32_t>(2ULL)))),ps2_IOP_Read(p,cast<uint32_t>((addr + cast<uint32_t>(3ULL))))};
mips_Inst in = mips_Decode(code,addr);
if (in.HasTarget) {
{
std::string name = ps2_IOP_callName(p,in.Target);
if ((name != std::string("",0))) {
return go_fmt_Sprintf(std::string("%-28s ; %s",10),in.Text,name);
}
}
{
std::string name = get(p->stubName,in.Target);
if ((name != std::string("",0))) {
return go_fmt_Sprintf(std::string("%-28s ; %s -> %s",16),in.Text,name,ps2_IOP_Sym(p,ps2_IOP_stubTarget(p,in.Target)));
}
}
{
std::string s = ps2_IOP_Sym(p,in.Target);
if ((s != std::string("",0))) {
return go_fmt_Sprintf(std::string("%-28s ; %s",10),in.Text,s);
}
}
}
return in.Text;
}
}
// tools/platform/ps2/iop.go:755:1
uint32_t ps2_IOP_stubTarget(ps2_IOP* p,uint32_t addr){
{
uint32_t w = ps2_IOP_Read32(p,addr);
if ((shr<uint32_t>(w,cast<int64_t>(26ULL)) != cast<uint32_t>(2ULL))) {
return addr;
}
return cast<uint32_t>((((addr & ~(cast<uint32_t>(268435455ULL)))) | shl<uint32_t>((cast<uint32_t>((w & cast<uint32_t>(67108863ULL)))),cast<int64_t>(2ULL))));
}
}
// tools/platform/ps2/iop.go:764:1
std::string ps2_IOP_callName(ps2_IOP* p,uint32_t addr){
{
if ((ps2_IOP_Read32(p,addr) != ps2_insnJR(ps2_regRA()))) {
return std::string("",0);
}
uint32_t w = ps2_IOP_Read32(p,cast<uint32_t>((addr + cast<uint32_t>(4ULL))));
if ((cast<uint32_t>((w & cast<uint32_t>(63ULL))) != cast<uint32_t>(12ULL))) {
return std::string("",0);
}
uint32_t code = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(6ULL))) & cast<uint32_t>(1048575ULL)));
if (((code == cast<uint32_t>(0ULL)) || (cast<int64_t>(code) >= len(p->calls)))) {
return std::string("",0);
}
return p->calls[code]->name;
}
}
// tools/platform/ps2/iop.go:780:1
Slice<ps2_IOPModule*> ps2_IOP_Modules(ps2_IOP* p){
{
return p->modules;
}
}
// tools/platform/ps2/iop.go:786:1
std::string ps2_IOP_Sym(ps2_IOP* p,uint32_t addr){
{
{auto&& tmp127 = p->modules;
for(int64_t tmp128=0;tmp128<len(tmp127);++tmp128){
auto m=tmp127[tmp128];if (((addr < m->Base) || (addr >= cast<uint32_t>((m->Base + m->Size))))) {
continue;
}
uint32_t off = cast<uint32_t>((addr - m->Base));
ps2_Symbol* best={};
{auto&& tmp129 = m->IRX->Symbols;
for(int64_t tmp130=0;tmp130<len(tmp129);++tmp130){
auto i=tmp130;ps2_Symbol* s = (&m->IRX->Symbols[i]);
if (((s->Func && (s->Addr <= off)) && (((!best) || (s->Addr > best->Addr))))) {
best = s;
}
}}
if (bool(best)) {
{
uint32_t d = cast<uint32_t>((off - best->Addr));
if ((d != cast<uint32_t>(0ULL))) {
return go_fmt_Sprintf(std::string("%s+0x%X",7),best->Name,d);
}
}
return best->Name;
}
return go_fmt_Sprintf(std::string("%s+0x%X",7),m->Name,off);
}}
return go_fmt_Sprintf(std::string("0x%08X",6),addr);
}
}
// tools/platform/ps2/iop.go:813:1
std::string ps2_IOP_SymGrep(ps2_IOP* p,std::string substr){
{
Slice<std::string> out={};
{auto&& tmp131 = p->modules;
for(int64_t tmp132=0;tmp132<len(tmp131);++tmp132){
auto m=tmp131[tmp132];{auto&& tmp133 = m->IRX->Symbols;
for(int64_t tmp134=0;tmp134<len(tmp133);++tmp134){
auto i=tmp134;ps2_Symbol* s = (&m->IRX->Symbols[i]);
if (((substr == std::string("",0)) || go_strings_Contains(s->Name,substr))) {
std::string kind = std::string("data",4);
if (s->Func) {
kind = std::string("func",4);
}
out = append(out,Slice<std::string>{go_fmt_Sprintf(std::string("  %08X  %-6s %-10s %s",21),cast<uint32_t>((m->Base + s->Addr)),kind,m->Name,s->Name)});
}
}}
}}
go_sort_Strings(out);
return go_strings_Join(out,std::string("\012",1));
}
}
// tools/platform/ps2/iop.go:837:1
std::tuple<uint32_t,bool> ps2_IOP_SymAddr(ps2_IOP* p,std::string name){
{
{auto&& tmp135 = p->modules;
for(int64_t tmp136=0;tmp136<len(tmp135);++tmp136){
auto m=tmp135[tmp136];{auto&& tmp137 = m->IRX->Symbols;
for(int64_t tmp138=0;tmp138<len(tmp137);++tmp138){
auto i=tmp138;{
ps2_Symbol* s = (&m->IRX->Symbols[i]);
if ((s->Name == name)) {
return {cast<uint32_t>((m->Base + s->Addr)),true};
}
}
}}
}}
return {cast<uint32_t>(0ULL),false};
}
}
// tools/platform/ps2/iop.go:852:1
void ps2_IOP_resolveTrap(ps2_IOP* p){
{
if (((p->Trap != cast<uint32_t>(0ULL)) || (p->TrapSym == std::string("",0)))) {
return ;
}
{
auto tmp139 = ps2_IOP_SymAddr(p,p->TrapSym);
uint32_t a = std::get<0>(tmp139);
bool ok = std::get<1>(tmp139);
if (ok) {
p->Trap = a;
ps2_Machine_note(p->ps2,std::string("IOP: the trap is armed at %s (0x%08X)",37),p->TrapSym,a);
}
}
}
}
// tools/platform/ps2/iop.go:884:1
std::string ps2_IOP_IOPTrail(ps2_IOP* p){
{
int64_t n = cast<int64_t>(24ULL);
if ((p->trailN < n)) {
n = p->trailN;
}
std::string s = go_fmt_Sprintf(std::string("the last %d instructions the IOP ran:\012",38),n);
{int64_t i = n;for (;(i > cast<int64_t>(0ULL));i--){
uint32_t pc = p->trail[modi<int64_t>((cast<int64_t>((p->trailN - i))),cast<int64_t>(24ULL))];
s += go_fmt_Sprintf(std::string("      %-24s %08X  %s\012",21),ps2_IOP_Sym(p,pc),pc,ps2_IOP_DisasmAt(p,pc));
}
}s += go_fmt_Sprintf(std::string("      with $sp=%08X $ra=%s $v0=%08X, %d deep in a call, %d deep in an interrupt\012",80),mips_CPU_Reg(p->CPU,cast<uint32_t>(29ULL)),ps2_IOP_Sym(p,mips_CPU_Reg(p->CPU,cast<uint32_t>(31ULL))),mips_CPU_Reg(p->CPU,cast<uint32_t>(2ULL)),p->callDepth,p->inIntr);
s += go_fmt_Sprintf(std::string("      args $a0=%08X $a1=%08X $a2=%08X $a3=%08X\012",47),mips_CPU_Reg(p->CPU,cast<uint32_t>(4ULL)),mips_CPU_Reg(p->CPU,cast<uint32_t>(5ULL)),mips_CPU_Reg(p->CPU,cast<uint32_t>(6ULL)),mips_CPU_Reg(p->CPU,cast<uint32_t>(7ULL)));
s += go_fmt_Sprintf(std::string("      intrEnabled=%v pending=%016X imask=%016X\012",47),p->intrEnabled,p->pending,p->imask);
s += ps2_IOP_isRunHistory(p);
return s;
}
}
// tools/platform/ps2/iop.go:907:1
std::string ps2_IOP_isRunHistory(ps2_IOP* p){
{
if (((p->schedIsRun == cast<uint32_t>(0ULL)) || (p->isRunLogN == cast<int64_t>(0ULL)))) {
return std::string("",0);
}
int64_t n = cast<int64_t>(40ULL);
if ((p->isRunLogN < n)) {
n = p->isRunLogN;
}
std::string s = go_fmt_Sprintf(std::string("      the scheduler's is-running/ought-to-run pointers (at %08X), oldest first:\012",80),p->schedIsRun);
{int64_t i = n;for (;(i > cast<int64_t>(0ULL));i--){
ps2_isRunWrite w = p->isRunLog[modi<int64_t>((cast<int64_t>((p->isRunLogN - i))),cast<int64_t>(40ULL))];
std::string which = std::string("is-run",6);
if (w.ought) {
which = std::string("ought ",6);
}
s += go_fmt_Sprintf(std::string("        %s = %-28s by %s\012",25),which,ps2_IOP_threadLabel(p,w.val),ps2_IOP_Sym(p,w.pc));
}
}return s;
}
}
// tools/platform/ps2/iop.go:929:1
std::string ps2_IOP_threadLabel(ps2_IOP* p,uint32_t tcb){
{
if ((tcb == cast<uint32_t>(0ULL))) {
return std::string("0",1);
}
return go_fmt_Sprintf(std::string("%08X (%s)",9),tcb,ps2_IOP_Sym(p,ps2_IOP_Read32(p,cast<uint32_t>((tcb + cast<uint32_t>(56ULL))))));
}
}
// tools/platform/ps2/iop.go:942:1
std::string ps2_IOP_IOPProfile(ps2_IOP* p){
{
if ((len(p->prof) == cast<int64_t>(0ULL))) {
return std::string("",0);
}
Slice<ps2_iop_946_kv> all={};
int64_t total = cast<int64_t>(0ULL);
{auto&& tmp140 = p->prof;
for(auto [tmp141,tmp142]:tmp140){
auto k=tmp141;auto n=tmp142;all = append(all,Slice<ps2_iop_946_kv>{ps2_iop_946_kv{k,n}});
total += n;
}}
go_sort_Slice(all,[&](int64_t i,int64_t j)->bool{
if ((all[i].n != all[j].n)) {
return (all[i].n > all[j].n);
}
return (all[i].name < all[j].name);
}
);
std::string s = go_fmt_Sprintf(std::string("where the IOP spent its time (%d samples, one per %d instructions):\012",68),total,cast<int64_t>(4096ULL));
{auto&& tmp143 = all;
for(int64_t tmp144=0;tmp144<len(tmp143);++tmp144){
auto i=tmp144;auto e=tmp143[tmp144];if ((i == cast<int64_t>(12ULL))) {
s += go_fmt_Sprintf(std::string("      ... and %d more routines\012",31),cast<int64_t>((len(all) - cast<int64_t>(12ULL))));
break;
}
s += go_fmt_Sprintf(std::string("      %5.1f%%  %s\012",18),((cast<double>(100ULL) * cast<double>(e.n)) / cast<double>(total)),e.name);
}}
return s;
}
}
// tools/platform/ps2/iop.go:981:1
std::string ps2_IOP_IOPCensus(ps2_IOP* p){
{
if (((len(p->unmodelledCalls) == cast<int64_t>(0ULL)) && (len(p->unmodelledIO) == cast<int64_t>(0ULL)))) {
return std::string("",0);
}
std::string s = std::string("the IOP's unanswered requests (the work list):\012",47);
if ((len(p->unmodelledCalls) > cast<int64_t>(0ULL))) {
Slice<ps2_iop_988_kv> all={};
{auto&& tmp145 = p->unmodelledCalls;
for(auto [tmp146,tmp147]:tmp145){
auto k=tmp146;auto n=tmp147;all = append(all,Slice<ps2_iop_988_kv>{ps2_iop_988_kv{k,n}});
}}
go_sort_Slice(all,[&](int64_t i,int64_t j)->bool{
if ((all[i].n != all[j].n)) {
return (all[i].n > all[j].n);
}
return (all[i].name < all[j].name);
}
);
s += std::string("  kernel functions nothing models:\012",35);
{auto&& tmp148 = all;
for(int64_t tmp149=0;tmp149<len(tmp148);++tmp149){
auto e=tmp148[tmp149];s += go_fmt_Sprintf(std::string("      %-20s %d call%s\012",22),e.name,e.n,ps2_plural(e.n));
}}
}
{
std::string d = ps2_cdvd_census(p->cdvd);
if ((d != std::string("",0))) {
s += (std::string("  ",2) + d);
}
}
if ((len(p->unmodelledIO) > cast<int64_t>(0ULL))) {
Slice<uint32_t> regs={};
{auto&& tmp150 = p->unmodelledIO;
for(auto [tmp151,tmp152]:tmp150){
auto a=tmp151;regs = append(regs,Slice<uint32_t>{a});
}}
go_sort_Slice(regs,[&](int64_t i,int64_t j)->bool{
return (get(p->unmodelledIO,regs[i]) > get(p->unmodelledIO,regs[j]));
}
);
s += go_fmt_Sprintf(std::string("  IOP hardware touched: %d registers\012",37),len(regs));
{auto&& tmp153 = regs;
for(int64_t tmp154=0;tmp154<len(tmp153);++tmp154){
auto i=tmp154;auto a=tmp153[tmp154];if ((i == cast<int64_t>(8ULL))) {
s += go_fmt_Sprintf(std::string("      ... and %d more\012",22),cast<int64_t>((len(regs) - cast<int64_t>(8ULL))));
break;
}
s += go_fmt_Sprintf(std::string("      0x%08X  %s  %d\012",21),a,ps2_IOPRegionName(a),get(p->unmodelledIO,a));
}}
}
return s;
}
}
// tools/platform/ps2/iop.go:1031:1
std::string ps2_IOPRegionName(uint32_t a){
{
{
if (((a >= cast<uint32_t>(528486464ULL)) && (a < cast<uint32_t>(528486496ULL)))){
return std::string("SIO",3);
}
else if (((a >= cast<uint32_t>(528486512ULL)) && (a < cast<uint32_t>(528486528ULL)))){
return std::string("INTC",4);
}
else if (((a >= cast<uint32_t>(528486528ULL)) && (a < cast<uint32_t>(528486656ULL)))){
return std::string("DMA",3);
}
else if (((a >= cast<uint32_t>(528486656ULL)) && (a < cast<uint32_t>(528486720ULL)))){
return std::string("TIMER",5);
}
else if (((a >= cast<uint32_t>(528487504ULL)) && (a < cast<uint32_t>(528487520ULL)))){
return std::string("?",1);
}
else if (((a >= cast<uint32_t>(528487680ULL)) && (a < cast<uint32_t>(528487808ULL)))){
return std::string("DMA2",4);
}
else if (((a >= cast<uint32_t>(528490608ULL)) && (a < cast<uint32_t>(528490624ULL)))){
return std::string("POST",4);
}
else if (((a >= cast<uint32_t>(528515584ULL)) && (a < cast<uint32_t>(528515840ULL)))){
return std::string("SIO2",4);
}
else if (((a >= cast<uint32_t>(529530880ULL)) && (a < cast<uint32_t>(530579456ULL)))){
return std::string("SPU2",4);
}
else if (((a >= cast<uint32_t>(486539264ULL)) && (a < cast<uint32_t>(486539520ULL)))){
return std::string("SIF",3);
}
}
tmp155:;
return std::string("?",1);
}
}
// tools/platform/ps2/iop.go:1059:1
std::string ps2_IOP_symFunc(ps2_IOP* p,uint32_t addr){
{
std::string s = ps2_IOP_Sym(p,addr);
{
int64_t i = cast<int64_t>((len(s) - cast<int64_t>(1ULL)));
if ((i > cast<int64_t>(0ULL))) {
{
int64_t j = ps2_indexByte(s,cast<uint8_t>(43ULL));
if ((j > cast<int64_t>(0ULL))) {
return sub(s,0,j);
}
}
}
}
return s;
}
}
// tools/platform/ps2/iop.go:1069:1
int64_t ps2_indexByte(std::string s,uint8_t c){
{
{int64_t i = cast<int64_t>(0ULL);for (;(i < len(s));i++){
if ((cast<uint8_t>(s[i]) == c)) {
return i;
}
}
}return cast<int64_t>(-1ULL);
}
}
// tools/platform/ps2/iop.go:1094:1
void ps2_IOP_ioTrace(ps2_IOP* p,uint32_t a,bool write){
{
if ((!p->OnIO)) {
return ;
}
p->ioPend = append(sub(p->ioPend,0,cast<int64_t>(0ULL)),Slice<ps2_ioTouch>{ps2_ioTouch{a,mips_CPU_CurPC(p->CPU),write}});
}
}
// tools/platform/ps2/iop.go:1109:1
void ps2_IOP_ioTraceFlush(ps2_IOP* p){
{
if (((!p->OnIO) || (len(p->ioPend) == cast<int64_t>(0ULL)))) {
return ;
}
ps2_ioTouch t = p->ioPend[cast<int64_t>(0ULL)];
p->ioPend = sub(p->ioPend,0,cast<int64_t>(0ULL));
p->OnIO(t.addr,ps2_IOP_ioPeek(p,t.addr),t.write,t.pc);
}
}
// tools/platform/ps2/iop.go:1129:1
void ps2_IOP_Run(ps2_IOP* p,uint64_t n){
{
{uint64_t i = cast<uint64_t>(0ULL);for (;(i < n);i++){
if ((((!p->running) || p->Halted) || p->CPU->Halted)) {
return ;
}
ps2_IOP_tick(p);
mips_CPU_Step(p->CPU);
}
}}
}
// tools/platform/ps2/iop.go:1140:1
uint64_t ps2_IOP_Steps(ps2_IOP* p){
{
return p->steps;
}
}
// tools/platform/ps2/iop.go:1149:1
bool ps2_IOP_BusyToEE(ps2_IOP* p){
{
{
ps2_cdvd* c = p->cdvd;
if (bool(c)) {
if (((c->nBusy || c->dmaArmed) || (len(c->data) > cast<int64_t>(0ULL)))) {
return true;
}
}
}
if ((len(p->ps2->sifToIOPQueue) > cast<int64_t>(0ULL))) {
return true;
}
return (cast<uint32_t>((p->dma[cast<int64_t>(9ULL)].chcr & cast<uint32_t>(16777216ULL))) != cast<uint32_t>(0ULL));
}
}
// tools/platform/ps2/iopboot.go:78:1
Error ps2_Machine_RebootIOP(ps2_Machine* m){
{
if (bool(m->bios)) {
return ps2_Machine_RebootIOPFrom(m,std::string("",0));
}
return ps2_Machine_RebootIOPFrom(m,ps2_iopBootImage);
}
}
// tools/platform/ps2/iopboot.go:93:1
Error ps2_Machine_RebootIOPFrom(ps2_Machine* m,std::string image){
{
if ((!m->vol)) {
return go_fmt_Errorf(std::string("ps2: no disc is mounted, so the IOP has nothing to boot from",60));
}
Map<std::string,Slice<uint8_t>> byName = Map<std::string,Slice<uint8_t>>{};
if (bool(m->bios)) {
auto tmp156 = ps2_ROMDIRModules(m->bios);
Slice<ps2_RomEntry> entries = std::get<0>(tmp156);
Error err = std::get<1>(tmp156);
if (bool(err)) {
return go_fmt_Errorf(std::string("ps2: the supplied BIOS is not a ROMDIR image: %w",48),err);
}
{auto&& tmp157 = entries;
for(int64_t tmp158=0;tmp158<len(tmp157);++tmp158){
auto e=tmp157[tmp158];byName[e.Name] = e.Data;
}}
}
if ((image == std::string("",0))) {
if ((!m->bios)) {
return go_fmt_Errorf(std::string("ps2: a ROM-only boot was requested but no BIOS is mounted",57));
}
}
else {
auto tmp159 = iso9660_Volume_ReadFile(m->vol,image);
Slice<uint8_t> raw = std::get<0>(tmp159);
Error err = std::get<1>(tmp159);
if (bool(err)) {
return go_fmt_Errorf(std::string("ps2: reading the IOP's boot image %s: %w",40),image,err);
}
auto tmp160 = ps2_ROMDIRModules(raw);
Slice<ps2_RomEntry> entries = std::get<0>(tmp160);
err = std::get<1>(tmp160);
if (bool(err)) {
return go_fmt_Errorf(std::string("ps2: %s is not a ROMDIR archive: %w",35),image,err);
}
{auto&& tmp161 = entries;
for(int64_t tmp162=0;tmp162<len(tmp161);++tmp162){
auto e=tmp161[tmp162];byName[e.Name] = e.Data;
}}
}
ps2_Machine_sbusFlagSet(m,cast<uint32_t>(32ULL),cast<uint32_t>(65536ULL));
ps2_Machine_StartIOP(m);
{auto&& tmp163 = ps2_iopBootOrder;
for(int64_t tmp164=0;tmp164<len(tmp163);++tmp164){
auto name=tmp163[tmp164];auto tmp165 = lookup(byName,name);
Slice<uint8_t> raw = std::get<0>(tmp165);
bool ok = std::get<1>(tmp165);
if ((!ok)) {
if ((!m->bios)) {
return go_fmt_Errorf(std::string("ps2: %s holds no module called %s, and no BIOS was supplied to fall back on",75),image,name);
}
return go_fmt_Errorf(std::string("ps2: neither %s nor the BIOS holds a module called %s",53),image,name);
}
{
Error err = ps2_IOP_LoadAndStart(m->IOP,name,raw);
if (bool(err)) {
return err;
}
}
}}
{
if ((image == std::string("",0))){
ps2_Machine_note(m,std::string("IOP: power-on boot from ROM (%d modules)",40),len(m->IOP->modules));
}
else if (bool(m->bios)){
ps2_Machine_note(m,std::string("IOP: booted on ROM + %s (%d modules)",36),image,len(m->IOP->modules));
}
else {
ps2_Machine_note(m,std::string("IOP: booted on %s (%d modules, image only)",42),image,len(m->IOP->modules));
}
}
tmp166:;
ps2_IOP_runBootCallbacks(m->IOP);
{auto&& tmp167 = m->IOPPokes;
for(auto [tmp168,tmp169]:tmp167){
auto addr=tmp168;auto val=tmp169;ps2_IOP_Write32(m->IOP,addr,val);
ps2_Machine_note(m,std::string("IOP: poked 0x%08X = 0x%08X (%s)",31),addr,val,ps2_IOP_Sym(m->IOP,addr));
}}
ps2_IOP_exitLoaderThread(m->IOP);
return {};
}
}
// tools/platform/ps2/iopboot.go:220:1
std::tuple<std::string,Error> ps2_iopRebootImage(std::string cmd){
{
std::string arg = cmd;
{
int64_t i = go_strings_IndexByte(arg,cast<uint8_t>(32ULL));
if ((i >= cast<int64_t>(0ULL))) {
arg = sub(arg,cast<int64_t>((i + cast<int64_t>(1ULL))),len(arg));
}
}
if ((!go_strings_HasPrefix(arg,std::string("cdrom0:",7)))) {
return {std::string("",0),go_fmt_Errorf(std::string("ps2: the EE asked the IOP to boot %q, which is not on the disc",62),cmd)};
}
std::string path = go_strings_TrimPrefix(arg,std::string("cdrom0:",7));
path = go_strings_ReplaceAll(path,std::string("\\",1),std::string("/",1));
{
int64_t i = go_strings_IndexByte(path,cast<uint8_t>(59ULL));
if ((i >= cast<int64_t>(0ULL))) {
path = sub(path,0,i);
}
}
return {path,{}};
}
}
// tools/platform/ps2/iopboot.go:239:1
Error ps2_IOP_LoadModuleFromDisc(ps2_IOP* p,std::string path){
{
auto tmp170 = iso9660_Volume_ReadFile(p->ps2->vol,path);
Slice<uint8_t> raw = std::get<0>(tmp170);
Error err = std::get<1>(tmp170);
if (bool(err)) {
return go_fmt_Errorf(std::string("ps2: reading %s: %w",19),path,err);
}
std::string name = path;
{
int64_t i = go_strings_LastIndexAny(name,std::string("/\\",2));
if ((i >= cast<int64_t>(0ULL))) {
name = sub(name,cast<int64_t>((i + cast<int64_t>(1ULL))),len(name));
}
}
return ps2_IOP_LoadAndStart(p,go_strings_TrimSuffix(name,std::string(";1",2)),raw);
}
}
// tools/platform/ps2/iopboot.go:259:1
void ps2_IOP_runBootCallbacks(ps2_IOP* p){
{
{auto&& tmp171 = p->bootCallbacks;
for(int64_t tmp172=0;tmp172<len(tmp171);++tmp172){
auto fn=tmp171[tmp172];{
auto tmp173 = ps2_IOP_callGuest(p,fn);
Error err = std::get<1>(tmp173);
if (bool(err)) {
ps2_Machine_note(p->ps2,std::string("IOP: the boot callback at %s did not return: %v",47),ps2_IOP_Sym(p,fn),err);
return ;
}
}
ps2_Machine_note(p->ps2,std::string("IOP: the boot is over, and %s was told so",41),ps2_IOP_Sym(p,fn));
}}
}
}
// tools/platform/ps2/iopboot.go:269:1
Error ps2_IOP_LoadAndStart(ps2_IOP* p,std::string name,Slice<uint8_t> raw){
{
auto tmp174 = ps2_IOP_LoadIRX(p,name,raw);
ps2_IOPModule* mod = std::get<0>(tmp174);
Error err = std::get<1>(tmp174);
if (bool(err)) {
return err;
}
if (bool(p->ps2->OnIOPModule)) {
p->ps2->OnIOPModule(p,name);
}
auto tmp175 = ps2_IOP_Start(p,mod,Slice<uint32_t>{});
uint32_t res = std::get<0>(tmp175);
err = std::get<1>(tmp175);
if (bool(err)) {
return err;
}
ps2_Machine_note(p->ps2,std::string("IOP: %s loaded at 0x%08X (%d KiB), started -> %d",48),name,mod->Base,divi<uint32_t>(mod->Size,cast<uint32_t>(1024ULL)),cast<int32_t>(res));
return {};
}
}
// tools/platform/ps2/iopcdvd.go:229:1
ps2_cdvd* ps2_newCDVD(ps2_Machine* m){
{
return arenaNew(ps2_cdvd{m,{},{},cast<uint8_t>(64ULL),{},{},{},{},{},{},{},{},{},{},{},{},{},Map<uint8_t,int64_t>{},Map<uint8_t,int64_t>{}});
}
}
// tools/platform/ps2/iopcdvd.go:257:1
uint8_t ps2_cdvd_discType(ps2_cdvd* c){
{
{
iso9660_Volume* vol = c->ps2->vol;
if (bool(vol)) {
auto tmp176 = iso9660_Volume_Geometry(vol);
iso9660_Geometry g = std::get<0>(tmp176);
bool ok = std::get<1>(tmp176);
return ps2_discTypeFor(g,ok,vol->Blocks);
}
}
return cast<uint8_t>(20ULL);
}
}
// tools/platform/ps2/iopcdvd.go:267:1
uint8_t ps2_discTypeFor(iso9660_Geometry g,bool geomKnown,int64_t blocks){
{
if ((geomKnown && (g.DataOffset != cast<int64_t>(0ULL)))) {
return cast<uint8_t>(18ULL);
}
if (((blocks > cast<int64_t>(0ULL)) && (blocks <= cast<int64_t>(360000ULL)))) {
return cast<uint8_t>(18ULL);
}
return cast<uint8_t>(20ULL);
}
}
// tools/platform/ps2/iopcdvd.go:288:1
bool ps2_cdvd_contains(ps2_cdvd* c,uint32_t a){
{
return ((a >= cast<uint32_t>(524296192ULL)) && (a < cast<uint32_t>(524296224ULL)));
}
}
// tools/platform/ps2/iopcdvd.go:290:1
uint8_t ps2_cdvd_read(ps2_cdvd* c,uint32_t a){
{
{
switch(cast<uint32_t>((a - cast<uint32_t>(524296192ULL)))){
case cast<uint32_t>(4ULL):{
return c->nCommand;
break;}
case cast<uint32_t>(5ULL):{
return c->nStatus;
break;}
case cast<uint32_t>(6ULL):{
return c->nError;
break;}
case cast<uint32_t>(8ULL):{
return c->intr;
break;}
case cast<uint32_t>(10ULL):{
if (c->nBusy) {
return cast<uint8_t>(6ULL);
}
return cast<uint8_t>(10ULL);
break;}
case cast<uint32_t>(11ULL):{
return cast<uint8_t>(0ULL);
break;}
case cast<uint32_t>(15ULL):{
return ps2_cdvd_discType(c);
break;}
case cast<uint32_t>(22ULL):{
return c->sCommand;
break;}
case cast<uint32_t>(23ULL):{
uint8_t st={};
if ((len(c->sResult) == cast<int64_t>(0ULL))) {
st |= cast<uint8_t>(64ULL);
}
return st;
break;}
case cast<uint32_t>(24ULL):{
if ((len(c->sResult) == cast<int64_t>(0ULL))) {
return cast<uint8_t>(0ULL);
}
uint8_t v = c->sResult[cast<int64_t>(0ULL)];
c->sResult = sub(c->sResult,cast<int64_t>(1ULL),len(c->sResult));
return v;
break;}
}}
ps2_Machine_iopUnknownCDVD(c->ps2,a,false);
return cast<uint8_t>(0ULL);
}
}
// tools/platform/ps2/iopcdvd.go:362:1
uint8_t ps2_cdvd_peek(ps2_cdvd* c,uint32_t a){
{
if ((cast<uint32_t>((a - cast<uint32_t>(524296192ULL))) == cast<uint32_t>(24ULL))) {
if ((len(c->sResult) == cast<int64_t>(0ULL))) {
return cast<uint8_t>(0ULL);
}
return c->sResult[cast<int64_t>(0ULL)];
}
return ps2_cdvd_read(c,a);
}
}
// tools/platform/ps2/iopcdvd.go:372:1
void ps2_cdvd_write(ps2_cdvd* c,uint32_t a,uint8_t v){
{
{
switch(cast<uint32_t>((a - cast<uint32_t>(524296192ULL)))){
case cast<uint32_t>(4ULL):{
ps2_cdvd_startN(c,v);
break;}
case cast<uint32_t>(5ULL):{
c->nParams = append(c->nParams,Slice<uint8_t>{v});
break;}
case cast<uint32_t>(6ULL):{
c->nMode = v;
break;}
case cast<uint32_t>(7ULL):{
break;}
case cast<uint32_t>(8ULL):{
c->intr &= ~(v);
break;}
case cast<uint32_t>(22ULL):{
ps2_cdvd_startS(c,v);
break;}
case cast<uint32_t>(23ULL):{
c->sParams = append(c->sParams,Slice<uint8_t>{v});
break;}
default:{
ps2_Machine_iopUnknownCDVD(c->ps2,a,true);
break;}
}}
}
}
// tools/platform/ps2/iopcdvd.go:399:1
void ps2_cdvd_startN(ps2_cdvd* c,uint8_t cmd){
{
c->nCommand = cmd;
c->nError = cast<uint8_t>(0ULL);
Slice<uint8_t> params = c->nParams;
c->nParams = {};
if (ps2_cdvdLog) {
auto tmp177 = std::make_tuple(cast<uint32_t>(0ULL),cast<uint32_t>(0ULL));
uint32_t lba = std::get<0>(tmp177);
uint32_t count = std::get<1>(tmp177);
if ((len(params) >= cast<int64_t>(8ULL))) {
auto tmp178 = std::make_tuple(ps2_le32(rrBorrow(params,cast<int64_t>(0ULL),len(params))),ps2_le32(rrBorrow(params,cast<int64_t>(4ULL),len(params))));
lba = std::get<0>(tmp178);
count = std::get<1>(tmp178);
}
go_fmt_Printf(std::string("CDVD N-cmd 0x%02X params=%s lba=%d count=%d mode=%02X vbl=%d IOPstep=%d\012",72),cmd,ps2_hexBytes(params),lba,count,ps2_mode8(params),ps2_Machine_VBlanks(c->ps2),c->ps2->IOP->steps);
}
if ((!ps2_cdvd_exec(c,cmd,params))) {
c->unknownN[cmd]++;
if ((get(c->unknownN,cmd) == cast<int64_t>(1ULL))) {
ps2_Machine_note(c->ps2,std::string("CDVD: N-command 0x%02X (%s) \342\200\224 nothing models it",49),cmd,ps2_hexBytes(params));
}
}
ps2_cdvd_pump(c,c->ps2->IOP);
c->nBusy = true;
c->nDoneAt = cast<uint64_t>((c->ps2->IOP->steps + cast<uint64_t>(2000ULL)));
}
}
// tools/platform/ps2/iopcdvd.go:433:1
bool ps2_cdvd_exec(ps2_cdvd* c,uint8_t cmd,Slice<uint8_t> params){
{
{
switch(cmd){
case cast<uint8_t>(6ULL):case cast<uint8_t>(8ULL):{
if ((len(params) < cast<int64_t>(8ULL))) {
return false;
}
uint32_t lba = ps2_le32(rrBorrow(params,cast<int64_t>(0ULL),len(params)));
uint32_t count = ps2_le32(rrBorrow(params,cast<int64_t>(4ULL),len(params)));
ps2_cdvd_readSectors(c,lba,count,(cmd == cast<uint8_t>(8ULL)));
return true;
break;}
}}
return false;
}
}
// tools/platform/ps2/iopcdvd.go:467:1
void ps2_cdvd_readSectors(ps2_cdvd* c,uint32_t lba,uint32_t count,bool dvd){
{
iso9660_Volume* vol = c->ps2->vol;
if ((!vol)) {
ps2_Machine_note(c->ps2,std::string("CDVD: asked to read %d sectors at LBA %d, and no disc is mounted",64),count,lba);
return ;
}
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < count);i++){
int64_t data = cast<int64_t>(12ULL);
int64_t secLen = cast<int64_t>(2064ULL);
if ((!dvd)) {
auto tmp179 = std::make_tuple(cast<int64_t>(0ULL),cast<int64_t>(2048ULL));
data = std::get<0>(tmp179);
secLen = std::get<1>(tmp179);
}
Slice<uint8_t> sec = Slice<uint8_t>::make(secLen);
if (dvd) {
uint32_t phys = cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(196608ULL) + lba)) + i));
sec[cast<int64_t>(1ULL)] = cast<uint8_t>(shr<uint32_t>(phys,cast<int64_t>(16ULL)));
sec[cast<int64_t>(2ULL)] = cast<uint8_t>(shr<uint32_t>(phys,cast<int64_t>(8ULL)));
sec[cast<int64_t>(3ULL)] = cast<uint8_t>(phys);
}
auto tmp180 = iso9660_Volume_ReadBlock(vol,cast<int64_t>(cast<uint32_t>((lba + i))));
Slice<uint8_t> blk = std::get<0>(tmp180);
Error err = std::get<1>(tmp180);
if (bool(err)) {
ps2_Machine_note(c->ps2,std::string("CDVD: reading LBA %d: %v",24),cast<uint32_t>((lba + i)),err);
}
else {
gcopy(sub(sec,data,len(sec)),blk);
}
c->data = append(c->data,sec);
}
}}
}
// tools/platform/ps2/iopcdvd.go:513:1
uint32_t ps2_le32(Slice<uint8_t> b){
{
return cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(b[cast<int64_t>(0ULL)]) | shl<uint32_t>(cast<uint32_t>(b[cast<int64_t>(1ULL)]),cast<int64_t>(8ULL)))) | shl<uint32_t>(cast<uint32_t>(b[cast<int64_t>(2ULL)]),cast<int64_t>(16ULL)))) | shl<uint32_t>(cast<uint32_t>(b[cast<int64_t>(3ULL)]),cast<int64_t>(24ULL))));
}
}
// tools/platform/ps2/iopcdvd.go:520:1
void ps2_cdvd_startS(ps2_cdvd* c,uint8_t cmd){
{
c->sCommand = cmd;
Slice<uint8_t> params = c->sParams;
c->sParams = {};
c->sResult = {};
if ((!ps2_cdvd_execS(c,cmd,params))) {
c->unknownS[cmd]++;
if ((get(c->unknownS,cmd) == cast<int64_t>(1ULL))) {
ps2_Machine_note(c->ps2,std::string("CDVD: S-command 0x%02X (%s) \342\200\224 nothing models it",49),cmd,ps2_hexBytes(params));
}
}
}
}
// tools/platform/ps2/iopcdvd.go:534:1
bool ps2_cdvd_execS(ps2_cdvd* c,uint8_t cmd,Slice<uint8_t> params){
{
{
switch(cmd){
case cast<uint8_t>(3ULL):{
c->sResult = Slice<uint8_t>{cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),cast<uint8_t>(0ULL),cast<uint8_t>(0ULL)};
return true;
break;}
case cast<uint8_t>(5ULL):{
c->sResult = Slice<uint8_t>{cast<uint8_t>(0ULL)};
return true;
break;}
}}
(void)(params);
return false;
}
}
// tools/platform/ps2/iopcdvd.go:564:1
void ps2_cdvd_tick(ps2_cdvd* c,ps2_IOP* p){
{
if (((!c->nBusy) || (p->steps < c->nDoneAt))) {
return ;
}
c->nBusy = false;
c->intr |= cast<uint8_t>(1ULL);
ps2_IOP_raiseIRQ(p,cast<uint32_t>(2ULL));
}
}
// tools/platform/ps2/iopcdvd.go:577:1
void ps2_cdvd_arm(ps2_cdvd* c,uint32_t madr,uint32_t n){
{
auto tmp181 = std::make_tuple(madr,n,true);
c->dmaMadr = std::get<0>(tmp181);
c->dmaLen = std::get<1>(tmp181);
c->dmaArmed = std::get<2>(tmp181);
}
}
// tools/platform/ps2/iopcdvd.go:588:1
void ps2_cdvd_pump(ps2_cdvd* c,ps2_IOP* p){
{
if (((!c->dmaArmed) || (cast<uint32_t>(len(c->data)) < c->dmaLen))) {
return ;
}
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < c->dmaLen);i++){
p->ram[cast<uint32_t>(((cast<uint32_t>((c->dmaMadr + i))) & cast<uint32_t>(2097151ULL)))] = c->data[i];
}
}c->data = sub(c->data,c->dmaLen,len(c->data));
c->dmaArmed = false;
p->dmaPending = append(p->dmaPending,Slice<ps2_iopDMADone>{ps2_iopDMADone{cast<uint64_t>((p->steps + cast<uint64_t>(400ULL))),cast<int64_t>(3ULL)}});
}
}
// tools/platform/ps2/iopcdvd.go:606:1
std::string ps2_cdvd_census(ps2_cdvd* c){
{
if (((len(c->unknownN) == cast<int64_t>(0ULL)) && (len(c->unknownS) == cast<int64_t>(0ULL)))) {
return std::string("",0);
}
std::string s = std::string("the CD/DVD drive's unanswered commands (the work list):\012",56);
{auto&& tmp182 = Slice<Anon44>{Anon44{std::string("N",1),c->unknownN},Anon44{std::string("S",1),c->unknownS}};
for(int64_t tmp183=0;tmp183<len(tmp182);++tmp183){
auto e=tmp182[tmp183];{auto&& tmp184 = e.m;
for(auto [tmp185,tmp186]:tmp184){
auto cmd=tmp185;auto n=tmp186;s += go_fmt_Sprintf(std::string("      %s-command 0x%02X   %d time%s\012",36),e.what,cmd,n,ps2_plural(n));
}}
}}
return s;
}
}
// tools/platform/ps2/iopcdvd.go:648:1
ps2_CDVDState ps2_cdvd_saveState(ps2_cdvd* c){
{
return ps2_CDVDState{c->nCommand,c->nStatus,c->nError,c->nMode,c->intr,append(Slice<uint8_t>{},c->nParams),append(Slice<uint8_t>{},c->lastParams),c->nBusy,c->nDoneAt,c->sCommand,append(Slice<uint8_t>{},c->sParams),append(Slice<uint8_t>{},c->sResult),append(Slice<uint8_t>{},c->data),c->dmaMadr,c->dmaLen,c->dmaArmed};
}
}
// tools/platform/ps2/iopcdvd.go:670:1
void ps2_cdvd_loadState(ps2_cdvd* c,ps2_CDVDState s){
{
if ((s.NStatus == cast<uint8_t>(0ULL))) {
return ;
}
auto tmp187 = std::make_tuple(s.NCommand,s.NStatus,s.NError,s.NMode,s.Intr);
c->nCommand = std::get<0>(tmp187);
c->nStatus = std::get<1>(tmp187);
c->nError = std::get<2>(tmp187);
c->nMode = std::get<3>(tmp187);
c->intr = std::get<4>(tmp187);
c->nParams = append(Slice<uint8_t>{},s.NParams);
c->lastParams = append(Slice<uint8_t>{},s.LastParams);
auto tmp188 = std::make_tuple(s.NBusy,s.NDoneAt);
c->nBusy = std::get<0>(tmp188);
c->nDoneAt = std::get<1>(tmp188);
c->sCommand = s.SCommand;
c->sParams = append(Slice<uint8_t>{},s.SParams);
c->sResult = append(Slice<uint8_t>{},s.SResult);
c->data = append(Slice<uint8_t>{},s.Data);
auto tmp189 = std::make_tuple(s.DMAMadr,s.DMALen,s.DMAArmed);
c->dmaMadr = std::get<0>(tmp189);
c->dmaLen = std::get<1>(tmp189);
c->dmaArmed = std::get<2>(tmp189);
}
}
// tools/platform/ps2/iopcdvd.go:685:1
uint8_t ps2_mode8(Slice<uint8_t> p){
{
if ((len(p) > cast<int64_t>(8ULL))) {
return p[cast<int64_t>(8ULL)];
}
return cast<uint8_t>(0ULL);
}
}
// tools/platform/ps2/iopcdvd.go:692:1
std::string ps2_hexBytes(Slice<uint8_t> b){
{
if ((len(b) == cast<int64_t>(0ULL))) {
return std::string("no parameters",13);
}
std::string s = std::string("",0);
{auto&& tmp190 = b;
for(int64_t tmp191=0;tmp191<len(tmp190);++tmp191){
auto i=tmp191;auto v=tmp190[tmp191];if ((i > cast<int64_t>(0ULL))) {
s += std::string(" ",1);
}
s += go_fmt_Sprintf(std::string("%02X",4),v);
}}
return s;
}
}
// tools/platform/ps2/iopdma.go:111:1
std::tuple<int64_t,uint32_t,bool> ps2_iopDMAReg(uint32_t a){
int64_t ch{};
uint32_t reg{};
bool ok{};
{
{
if (((a >= cast<uint32_t>(528486528ULL)) && (a < cast<uint32_t>(528486640ULL)))){
return {divi<int64_t>(cast<int64_t>(cast<uint32_t>((a - cast<uint32_t>(528486528ULL)))),cast<int64_t>(16ULL)),cast<uint32_t>((a & cast<uint32_t>(12ULL))),true};
}
else if (((a >= cast<uint32_t>(528487680ULL)) && (a < cast<uint32_t>(528487792ULL)))){
return {cast<int64_t>((cast<int64_t>(7ULL) + divi<int64_t>(cast<int64_t>(cast<uint32_t>((a - cast<uint32_t>(528487680ULL)))),cast<int64_t>(16ULL)))),cast<uint32_t>((a & cast<uint32_t>(12ULL))),true};
}
}
tmp192:;
return {cast<int64_t>(0ULL),cast<uint32_t>(0ULL),false};
}
}
// tools/platform/ps2/iopdma.go:122:1
std::tuple<uint32_t,bool> ps2_IOP_dmaRead(ps2_IOP* p,uint32_t a){
{
{
auto tmp193 = ps2_iopDMAReg(a);
int64_t ch = std::get<0>(tmp193);
uint32_t reg = std::get<1>(tmp193);
bool ok = std::get<2>(tmp193);
if (ok) {
ps2_iopDMAChan* c = (&p->dma[ch]);
{
switch(reg){
case cast<uint32_t>(0ULL):{
return {c->madr,true};
break;}
case cast<uint32_t>(4ULL):{
return {c->bcr,true};
break;}
case cast<uint32_t>(8ULL):{
return {c->chcr,true};
break;}
case cast<uint32_t>(12ULL):{
return {c->tadr,true};
break;}
}}
}
}
{
switch(a){
case cast<uint32_t>(528486640ULL):{
return {p->dpcr,true};
break;}
case cast<uint32_t>(528487792ULL):{
return {p->dpcr2,true};
break;}
case cast<uint32_t>(528486644ULL):{
return {p->dicr,true};
break;}
case cast<uint32_t>(528487796ULL):{
return {p->dicr2,true};
break;}
}}
return {cast<uint32_t>(0ULL),false};
}
}
// tools/platform/ps2/iopdma.go:153:1
bool ps2_IOP_dmaWrite(ps2_IOP* p,uint32_t a,uint32_t v){
{
{
auto tmp194 = ps2_iopDMAReg(a);
int64_t ch = std::get<0>(tmp194);
uint32_t reg = std::get<1>(tmp194);
bool ok = std::get<2>(tmp194);
if (ok) {
ps2_iopDMAChan* c = (&p->dma[ch]);
{
switch(reg){
case cast<uint32_t>(0ULL):{
c->madr = cast<uint32_t>((v & cast<uint32_t>(16777215ULL)));
break;}
case cast<uint32_t>(4ULL):{
c->bcr = v;
break;}
case cast<uint32_t>(8ULL):{
c->chcr = v;
if ((cast<uint32_t>((v & cast<uint32_t>(16777216ULL))) != cast<uint32_t>(0ULL))) {
ps2_IOP_dmaStart(p,ch);
}
break;}
case cast<uint32_t>(12ULL):{
c->tadr = v;
break;}
}}
return true;
}
}
{
switch(a){
case cast<uint32_t>(528486640ULL):{
p->dpcr = v;
return true;
break;}
case cast<uint32_t>(528487792ULL):{
p->dpcr2 = v;
return true;
break;}
case cast<uint32_t>(528486644ULL):{
p->dicr = cast<uint32_t>((((p->dicr & ~((cast<uint32_t>((v & cast<uint32_t>(2130706432ULL))))))) | ((v & ~(cast<uint32_t>(2130706432ULL))))));
return true;
break;}
case cast<uint32_t>(528487796ULL):{
p->dicr2 = cast<uint32_t>((((p->dicr2 & ~((cast<uint32_t>((v & cast<uint32_t>(1056964608ULL))))))) | ((v & ~(cast<uint32_t>(1056964608ULL))))));
return true;
break;}
}}
return false;
}
}
// tools/platform/ps2/iopdma.go:196:1
bool ps2_IOP_dmaEnabled(ps2_IOP* p,int64_t ch){
{
if ((ch < cast<int64_t>(7ULL))) {
return (cast<uint32_t>((shr<uint32_t>(p->dpcr,(cast<uint64_t>((cast<uint64_t>(ch) * cast<uint64_t>(4ULL))))) & cast<uint32_t>(8ULL))) != cast<uint32_t>(0ULL));
}
return (cast<uint32_t>((shr<uint32_t>(p->dpcr2,(cast<uint64_t>((cast<uint64_t>(cast<int64_t>((ch - cast<int64_t>(7ULL)))) * cast<uint64_t>(4ULL))))) & cast<uint32_t>(8ULL))) != cast<uint32_t>(0ULL));
}
}
// tools/platform/ps2/iopdma.go:209:1
void ps2_IOP_dmaStart(ps2_IOP* p,int64_t ch){
{
ps2_iopDMAChan* c = (&p->dma[ch]);
if ((!ps2_IOP_dmaEnabled(p,ch))) {
ps2_Machine_note(p->ps2,std::string("IOP DMA: channel %d started but DPCR never enabled it \342\200\224 check the decode",74),ch);
}
uint32_t words = cast<uint32_t>((c->bcr & cast<uint32_t>(65535ULL)));
uint32_t blocks = shr<uint32_t>(c->bcr,cast<int64_t>(16ULL));
if ((shr<uint32_t>((cast<uint32_t>((c->chcr & cast<uint32_t>(1536ULL)))),cast<int64_t>(9ULL)) == cast<uint32_t>(0ULL))) {
if ((words == cast<uint32_t>(0ULL))) {
words = cast<uint32_t>(65536ULL);
}
blocks = cast<uint32_t>(1ULL);
}
if ((blocks == cast<uint32_t>(0ULL))) {
blocks = cast<uint32_t>(1ULL);
}
uint32_t n = cast<uint32_t>((cast<uint32_t>((words * blocks)) * cast<uint32_t>(4ULL)));
bool toRAM = (cast<uint32_t>((c->chcr & cast<uint32_t>(1ULL))) == cast<uint32_t>(0ULL));
{
switch(ch){
case cast<int64_t>(4ULL):{
ps2_spu2_dma(p->spu,cast<int64_t>(0ULL),p,c->madr,n,toRAM);
break;}
case cast<int64_t>(7ULL):{
ps2_spu2_dma(p->spu,cast<int64_t>(1ULL),p,c->madr,n,toRAM);
break;}
case cast<int64_t>(3ULL):{
ps2_cdvd_arm(p->cdvd,c->madr,n);
ps2_cdvd_pump(p->cdvd,p);
return ;
break;}
case cast<int64_t>(9ULL):{
if ((!toRAM)) {
ps2_Machine_sifFromIOP(p->ps2);
}
break;}
case cast<int64_t>(10ULL):{
ps2_Machine_sifPump(p->ps2);
return ;
break;}
default:{
p->unmodelledCalls[go_fmt_Sprintf(std::string("DMA channel %d",14),ch)]++;
if ((get(p->unmodelledCalls,go_fmt_Sprintf(std::string("DMA channel %d",14),ch)) == cast<int64_t>(1ULL))) {
ps2_Machine_note(p->ps2,std::string("IOP DMA: channel %d moved %d bytes at 0x%08X, and nothing models that channel",77),ch,n,c->madr);
}
break;}
}}
p->dmaPending = append(p->dmaPending,Slice<ps2_iopDMADone>{ps2_iopDMADone{cast<uint64_t>((p->steps + cast<uint64_t>(400ULL))),ch}});
}
}
// tools/platform/ps2/iopdma.go:305:1
void ps2_IOP_dmaTick(ps2_IOP* p){
{
{;for (;((len(p->dmaPending) > cast<int64_t>(0ULL)) && (p->steps >= p->dmaPending[cast<int64_t>(0ULL)].at));){
ps2_iopDMADone d = p->dmaPending[cast<int64_t>(0ULL)];
p->dmaPending = sub(p->dmaPending,cast<int64_t>(1ULL),len(p->dmaPending));
p->dma[d.ch].chcr &= ~(cast<uint32_t>(285212672ULL));
{
switch(d.ch){
case cast<int64_t>(4ULL):{
ps2_spu2_complete(p->spu,cast<int64_t>(0ULL),p);
break;}
case cast<int64_t>(7ULL):{
ps2_spu2_complete(p->spu,cast<int64_t>(1ULL),p);
break;}
}}
ps2_IOP_raiseIRQ(p,ps2_iopDMAIRQ(d.ch));
}
}}
}
// tools/platform/ps2/iopdma.go:337:1
uint32_t ps2_iopDMAIRQ(int64_t ch){
{
if ((ch < cast<int64_t>(7ULL))) {
return cast<uint32_t>(cast<int64_t>((cast<int64_t>(32ULL) + ch)));
}
return cast<uint32_t>(cast<int64_t>((cast<int64_t>((cast<int64_t>(40ULL) + ch)) - cast<int64_t>(7ULL))));
}
}
// tools/platform/ps2/iopframe.go:97:1
uint32_t ps2_IOP_findThreadExit(ps2_IOP* p){
{
Map<uint32_t,int64_t> counts = Map<uint32_t,int64_t>{};
{auto&& tmp195 = p->blocks;
for(int64_t tmp196=0;tmp196<len(tmp195);++tmp196){
auto b=tmp195[tmp196];if ((b.size < cast<uint32_t>(184ULL))) {
continue;
}
uint32_t frame = ((cast<uint32_t>((cast<uint32_t>((b.base + b.size)) - cast<uint32_t>(184ULL)))) & ~(cast<uint32_t>(7ULL)));
if ((ps2_IOP_Read32(p,cast<uint32_t>((frame + cast<uint32_t>(0ULL)))) != cast<uint32_t>(4294967294ULL))) {
continue;
}
counts[ps2_IOP_Read32(p,cast<uint32_t>((frame + cast<uint32_t>(124ULL))))]++;
}}
auto tmp197 = std::make_tuple(cast<uint32_t>(0ULL),cast<int64_t>(0ULL));
uint32_t best = std::get<0>(tmp197);
int64_t n = std::get<1>(tmp197);
{auto&& tmp198 = counts;
for(auto [tmp199,tmp200]:tmp198){
auto addr=tmp199;auto c=tmp200;if ((c > n)) {
auto tmp201 = std::make_tuple(addr,c);
best = std::get<0>(tmp201);
n = std::get<1>(tmp201);
}
}}
return best;
}
}
// tools/platform/ps2/iopframe.go:136:1
void ps2_IOP_exitLoaderThread(ps2_IOP* p){
{
uint32_t exit = ps2_IOP_findThreadExit(p);
if ((exit == cast<uint32_t>(0ULL))) {
ps2_Machine_note(p->ps2,std::string("IOP: no thread was ever started, so the loader has nothing to stand down through",80));
return ;
}
mips_CPU_SetPC(p->CPU,exit);
mips_CPU_SetReg(p->CPU,cast<uint32_t>(31ULL),cast<uint32_t>(512ULL));
ps2_Machine_note(p->ps2,std::string("IOP: the loader stands down through %s, and the scheduler takes over",68),ps2_IOP_Sym(p,exit));
}
}
// tools/platform/ps2/iopframe.go:148:1
void ps2_IOP_saveFrame(ps2_IOP* p,uint32_t at){
{
mips_CPUState st = mips_CPU_SaveState(p->CPU);
ps2_IOP_Write32(p,cast<uint32_t>((at + cast<uint32_t>(0ULL))),cast<uint32_t>(4294967294ULL));
{uint32_t i = cast<uint32_t>(1ULL);for (;(i < cast<uint32_t>(32ULL));i++){
ps2_IOP_Write32(p,cast<uint32_t>((at + cast<uint32_t>((cast<uint32_t>(4ULL) * i)))),st.R[i]);
}
}ps2_IOP_Write32(p,cast<uint32_t>((at + cast<uint32_t>(128ULL))),st.HI);
ps2_IOP_Write32(p,cast<uint32_t>((at + cast<uint32_t>(132ULL))),st.LO);
ps2_IOP_Write32(p,cast<uint32_t>((at + cast<uint32_t>(136ULL))),shl<uint32_t>(ps2_b2u(p->intrEnabled),cast<int64_t>(2ULL)));
ps2_IOP_Write32(p,cast<uint32_t>((at + cast<uint32_t>(140ULL))),st.PC);
ps2_IOP_ieEvent(p,std::string("save",4),at,st.PC,st.R[cast<int64_t>(31ULL)]);
}
}
// tools/platform/ps2/iopframe.go:177:1
void ps2_IOP_loadFrame(ps2_IOP* p,uint32_t at){
{
mips_CPUState st = mips_CPU_SaveState(p->CPU);
{uint32_t i = cast<uint32_t>(1ULL);for (;(i < cast<uint32_t>(32ULL));i++){
st.R[i] = ps2_IOP_Read32(p,cast<uint32_t>((at + cast<uint32_t>((cast<uint32_t>(4ULL) * i)))));
}
}st.R[cast<int64_t>(0ULL)] = cast<uint32_t>(0ULL);
st.Out = st.R;
st.HI = ps2_IOP_Read32(p,cast<uint32_t>((at + cast<uint32_t>(128ULL))));
st.LO = ps2_IOP_Read32(p,cast<uint32_t>((at + cast<uint32_t>(132ULL))));
uint32_t pc = ps2_IOP_Read32(p,cast<uint32_t>((at + cast<uint32_t>(140ULL))));
auto tmp202 = std::make_tuple(pc,cast<uint32_t>((pc + cast<uint32_t>(4ULL))),pc);
st.PC = std::get<0>(tmp202);
st.NextPC = std::get<1>(tmp202);
st.CurPC = std::get<2>(tmp202);
auto tmp203 = std::make_tuple(false,false,cast<uint32_t>(0ULL));
st.DelaySlot = std::get<0>(tmp203);
st.PendingDelay = std::get<1>(tmp203);
st.BranchAddr = std::get<2>(tmp203);
auto tmp204 = std::make_tuple(cast<uint32_t>(0ULL),cast<uint32_t>(0ULL));
st.LdReg = std::get<0>(tmp204);
st.LdVal = std::get<1>(tmp204);
mips_CPU_LoadState(p->CPU,st);
p->intrEnabled = (cast<uint32_t>((shr<uint32_t>(ps2_IOP_Read32(p,cast<uint32_t>((at + cast<uint32_t>(136ULL)))),cast<int64_t>(2ULL)) & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL));
ps2_IOP_ieEvent(p,std::string("load",4),at,pc,st.R[cast<int64_t>(31ULL)]);
}
}
// tools/platform/ps2/iopintr.go:19:1
void ps2_init_iopintr(){
{
ps2_lib(std::string("intrman",7),Map<uint16_t,ps2_iopFunc>{{cast<uint16_t>(4ULL),ps2_iopFunc{std::string("RegisterIntrHandler",19),ps2_IOP_intrRegister}},{cast<uint16_t>(5ULL),ps2_iopFunc{std::string("ReleaseIntrHandler",18),ps2_IOP_intrRelease}},{cast<uint16_t>(6ULL),ps2_iopFunc{std::string("EnableIntr",10),ps2_IOP_intrEnable}},{cast<uint16_t>(7ULL),ps2_iopFunc{std::string("DisableIntr",11),ps2_IOP_intrDisable}},{cast<uint16_t>(8ULL),ps2_iopFunc{std::string("CpuDisableIntr",14),ps2_IOP_intrCpuDisable}},{cast<uint16_t>(9ULL),ps2_iopFunc{std::string("CpuEnableIntr",13),ps2_IOP_intrCpuEnable}},{cast<uint16_t>(14ULL),ps2_iopFunc{std::string("CpuInvokeInKmode",16),ps2_IOP_intrInvokeInKmode}},{cast<uint16_t>(17ULL),ps2_iopFunc{std::string("CpuSuspendIntr",14),ps2_IOP_intrSuspend}},{cast<uint16_t>(18ULL),ps2_iopFunc{std::string("CpuResumeIntr",13),ps2_IOP_intrResume}},{cast<uint16_t>(23ULL),ps2_iopFunc{std::string("QueryIntrContext",16),ps2_IOP_intrQueryContext}},{cast<uint16_t>(28ULL),ps2_iopFunc{std::string("<register scheduler hook>",25),ps2_IOP_intrSetSwitchHook}},{cast<uint16_t>(30ULL),ps2_iopFunc{std::string("<register reschedule predicate>",31),ps2_IOP_intrSetReschedHook}}});
ps2_lib(std::string("sysmem",6),Map<uint16_t,ps2_iopFunc>{{cast<uint16_t>(4ULL),ps2_iopFunc{std::string("AllocSysMemory",14),ps2_IOP_sysmemAlloc}},{cast<uint16_t>(5ULL),ps2_iopFunc{std::string("FreeSysMemory",13),ps2_IOP_sysmemFree}},{cast<uint16_t>(6ULL),ps2_unknown()},{cast<uint16_t>(7ULL),ps2_unknown()},{cast<uint16_t>(8ULL),ps2_unknown()},{cast<uint16_t>(9ULL),ps2_iopFunc{std::string("QueryBlockTopAddress",20),ps2_IOP_sysmemBlockTop}},{cast<uint16_t>(10ULL),ps2_iopFunc{std::string("QueryBlockSize",14),ps2_IOP_sysmemBlockSize}},{cast<uint16_t>(14ULL),ps2_iopFunc{std::string("Kprintf",7),ps2_IOP_sysmemKprintf}}});
ps2_lib(std::string("heaplib",7),Map<uint16_t,ps2_iopFunc>{{cast<uint16_t>(4ULL),ps2_iopFunc{std::string("CreateHeap",10),ps2_IOP_heapCreate}},{cast<uint16_t>(5ULL),ps2_unknown()},{cast<uint16_t>(6ULL),ps2_iopFunc{std::string("AllocHeapMemory",15),ps2_IOP_heapAlloc}},{cast<uint16_t>(7ULL),ps2_unknown()},{cast<uint16_t>(8ULL),ps2_unknown()}});
ps2_lib(std::string("sysclib",7),Map<uint16_t,ps2_iopFunc>{{cast<uint16_t>(8ULL),ps2_iopFunc{std::string("look_ctype_table",16),ps2_IOP_clibCtype}},{cast<uint16_t>(11ULL),ps2_unknown()},{cast<uint16_t>(12ULL),ps2_iopFunc{std::string("memcpy",6),ps2_IOP_clibMemcpy}},{cast<uint16_t>(13ULL),ps2_iopFunc{std::string("memmove",7),ps2_IOP_clibMemmove}},{cast<uint16_t>(14ULL),ps2_iopFunc{std::string("memset",6),ps2_IOP_clibMemset}},{cast<uint16_t>(17ULL),ps2_iopFunc{std::string("bzero",5),ps2_IOP_clibBzero}},{cast<uint16_t>(19ULL),ps2_unknown()},{cast<uint16_t>(20ULL),ps2_iopFunc{std::string("strcat",6),ps2_IOP_clibStrcat}},{cast<uint16_t>(22ULL),ps2_iopFunc{std::string("strcmp",6),ps2_IOP_clibStrcmp}},{cast<uint16_t>(23ULL),ps2_iopFunc{std::string("strcpy",6),ps2_IOP_clibStrcpy}},{cast<uint16_t>(25ULL),ps2_iopFunc{std::string("strchr",6),ps2_IOP_clibStrchr}},{cast<uint16_t>(27ULL),ps2_iopFunc{std::string("strlen",6),ps2_IOP_clibStrlen}},{cast<uint16_t>(29ULL),ps2_iopFunc{std::string("strncmp",7),ps2_IOP_clibStrncmp}},{cast<uint16_t>(30ULL),ps2_iopFunc{std::string("strncpy",7),ps2_IOP_clibStrncpy}},{cast<uint16_t>(36ULL),ps2_unknown()}});
}
}
// tools/platform/ps2/iopintr.go:218:1
void ps2_IOP_intrRegister(ps2_IOP* p){
{
auto tmp205 = std::make_tuple(ps2_IOP_arg(p,cast<int64_t>(0ULL)),ps2_IOP_arg(p,cast<int64_t>(1ULL)),ps2_IOP_arg(p,cast<int64_t>(2ULL)),ps2_IOP_arg(p,cast<int64_t>(3ULL)));
uint32_t irq = std::get<0>(tmp205);
uint32_t fn = std::get<2>(tmp205);
uint32_t arg = std::get<3>(tmp205);
if ((irq >= cast<uint32_t>(64ULL))) {
ps2_Machine_note(p->ps2,std::string("IOP: a handler was registered on interrupt %d, which is past the end of the table",81),irq);
ps2_IOP_setRet(p,cast<uint32_t>(4294967194ULL));
return ;
}
p->handlers[irq] = ps2_iopHandler{fn,arg};
ps2_Machine_note(p->ps2,std::string("IOP: interrupt %d handled by %s",31),irq,ps2_IOP_Sym(p,fn));
ps2_IOP_setRet(p,cast<uint32_t>(0ULL));
}
}
// tools/platform/ps2/iopintr.go:230:1
void ps2_IOP_intrRelease(ps2_IOP* p){
{
{
uint32_t irq = ps2_IOP_arg(p,cast<int64_t>(0ULL));
if ((irq < cast<uint32_t>(64ULL))) {
p->handlers[irq] = ps2_iopHandler{};
}
}
ps2_IOP_setRet(p,cast<uint32_t>(0ULL));
}
}
// tools/platform/ps2/iopintr.go:255:1
uint32_t ps2_IOP_intrLine(ps2_IOP* p,uint32_t arg){
{
return cast<uint32_t>((arg & cast<uint32_t>(63ULL)));
}
}
// tools/platform/ps2/iopintr.go:257:1
void ps2_IOP_intrEnable(ps2_IOP* p){
{
p->imask |= shl<uint64_t>(cast<uint64_t>(1ULL),cast<uint64_t>(ps2_IOP_intrLine(p,ps2_IOP_arg(p,cast<int64_t>(0ULL)))));
ps2_IOP_setRet(p,cast<uint32_t>(0ULL));
}
}
// tools/platform/ps2/iopintr.go:262:1
void ps2_IOP_intrDisable(ps2_IOP* p){
{
p->imask &= ~(shl<uint64_t>(cast<uint64_t>(1ULL),cast<uint64_t>(ps2_IOP_intrLine(p,ps2_IOP_arg(p,cast<int64_t>(0ULL))))));
ps2_IOP_setRet(p,cast<uint32_t>(0ULL));
}
}
// tools/platform/ps2/iopintr.go:273:1
void ps2_IOP_intrSuspend(ps2_IOP* p){
{
uint32_t old = ps2_b2u(p->intrEnabled);
{
uint32_t ptr = ps2_IOP_arg(p,cast<int64_t>(0ULL));
if ((ptr != cast<uint32_t>(0ULL))) {
ps2_IOP_Write32(p,ptr,old);
}
}
p->intrEnabled = false;
ps2_IOP_ieEvent(p,std::string("suspend",7),ps2_IOP_arg(p,cast<int64_t>(0ULL)),old,mips_CPU_Reg(p->CPU,cast<uint32_t>(31ULL)));
ps2_IOP_setRet(p,cast<uint32_t>(0ULL));
}
}
// tools/platform/ps2/iopintr.go:284:1
void ps2_IOP_intrResume(ps2_IOP* p){
{
p->intrEnabled = (ps2_IOP_arg(p,cast<int64_t>(0ULL)) != cast<uint32_t>(0ULL));
ps2_IOP_ieEvent(p,std::string("resume",6),cast<uint32_t>(0ULL),ps2_IOP_arg(p,cast<int64_t>(0ULL)),mips_CPU_Reg(p->CPU,cast<uint32_t>(31ULL)));
ps2_IOP_setRet(p,cast<uint32_t>(0ULL));
}
}
// tools/platform/ps2/iopintr.go:305:1
void ps2_IOP_intrCpuDisable(ps2_IOP* p){
{
p->intrEnabled = false;
ps2_IOP_ieEvent(p,std::string("cpu-off",7),cast<uint32_t>(0ULL),cast<uint32_t>(0ULL),mips_CPU_Reg(p->CPU,cast<uint32_t>(31ULL)));
ps2_IOP_setRet(p,cast<uint32_t>(0ULL));
}
}
// tools/platform/ps2/iopintr.go:311:1
void ps2_IOP_intrCpuEnable(ps2_IOP* p){
{
p->intrEnabled = true;
ps2_IOP_ieEvent(p,std::string("cpu-on",6),cast<uint32_t>(0ULL),cast<uint32_t>(0ULL),mips_CPU_Reg(p->CPU,cast<uint32_t>(31ULL)));
ps2_IOP_setRet(p,cast<uint32_t>(0ULL));
}
}
// tools/platform/ps2/iopintr.go:320:1
void ps2_IOP_intrSetSwitchHook(ps2_IOP* p){
{
p->schedSwitch = ps2_IOP_arg(p,cast<int64_t>(0ULL));
ps2_Machine_note(p->ps2,std::string("IOP: the scheduler's switch routine is %s",41),ps2_IOP_Sym(p,p->schedSwitch));
ps2_IOP_setRet(p,cast<uint32_t>(0ULL));
}
}
// tools/platform/ps2/iopintr.go:326:1
void ps2_IOP_intrSetReschedHook(ps2_IOP* p){
{
p->schedResched = ps2_IOP_arg(p,cast<int64_t>(0ULL));
ps2_Machine_note(p->ps2,std::string("IOP: the scheduler's reschedule predicate is %s",47),ps2_IOP_Sym(p,p->schedResched));
ps2_IOP_deriveSchedIsRun(p);
ps2_IOP_setRet(p,cast<uint32_t>(0ULL));
}
}
// tools/platform/ps2/iopintr.go:340:1
void ps2_IOP_deriveSchedIsRun(ps2_IOP* p){
{
if ((p->schedResched == cast<uint32_t>(0ULL))) {
return ;
}
auto tmp206 = std::make_tuple(ps2_IOP_Read32(p,p->schedResched),ps2_IOP_Read32(p,cast<uint32_t>((p->schedResched + cast<uint32_t>(4ULL)))));
uint32_t lui = std::get<0>(tmp206);
uint32_t addiu = std::get<1>(tmp206);
if (((shr<uint32_t>(lui,cast<int64_t>(26ULL)) == cast<uint32_t>(15ULL)) && (shr<uint32_t>(addiu,cast<int64_t>(26ULL)) == cast<uint32_t>(9ULL)))) {
uint32_t g = cast<uint32_t>((shl<uint32_t>((cast<uint32_t>((lui & cast<uint32_t>(65535ULL)))),cast<int64_t>(16ULL)) + cast<uint32_t>(cast<int32_t>(cast<int16_t>(cast<uint16_t>(cast<uint32_t>((addiu & cast<uint32_t>(65535ULL)))))))));
p->schedIsRun = cast<uint32_t>((g - cast<uint32_t>(4ULL)));
}
}
}
// tools/platform/ps2/iopintr.go:352:1
void ps2_IOP_intrQueryContext(ps2_IOP* p){
{
ps2_IOP_setRet(p,ps2_b2u((p->inIntr > cast<int64_t>(0ULL))));
}
}
// tools/platform/ps2/iopintr.go:368:1
void ps2_IOP_intrInvokeInKmode(ps2_IOP* p){
{
uint32_t fn = ps2_IOP_arg(p,cast<int64_t>(0ULL));
if ((fn == cast<uint32_t>(0ULL))) {
ps2_IOP_setRet(p,cast<uint32_t>(0ULL));
return ;
}
auto tmp207 = std::make_tuple(ps2_IOP_arg(p,cast<int64_t>(1ULL)),ps2_IOP_arg(p,cast<int64_t>(2ULL)),ps2_IOP_arg(p,cast<int64_t>(3ULL)));
uint32_t a1 = std::get<0>(tmp207);
uint32_t a2 = std::get<1>(tmp207);
uint32_t a3 = std::get<2>(tmp207);
mips_CPU_SetReg(p->CPU,cast<uint32_t>(4ULL),a1);
mips_CPU_SetReg(p->CPU,cast<uint32_t>(5ULL),a2);
mips_CPU_SetReg(p->CPU,cast<uint32_t>(6ULL),a3);
mips_CPU_SetReg(p->CPU,cast<uint32_t>(7ULL),cast<uint32_t>(0ULL));
auto tmp208 = ps2_IOP_callGuestOn(p,fn,((cast<uint32_t>((mips_CPU_Reg(p->CPU,cast<uint32_t>(29ULL)) - cast<uint32_t>(64ULL)))) & ~(cast<uint32_t>(7ULL))));
uint32_t res = std::get<0>(tmp208);
Error err = std::get<1>(tmp208);
if (bool(err)) {
ps2_IOP_halt(p,std::string("the routine %s, invoked through intrman #14, did not return: %v",63),ps2_IOP_Sym(p,fn),err);
return ;
}
ps2_IOP_setRet(p,res);
}
}
// tools/platform/ps2/iopintr.go:404:1
void ps2_IOP_raiseIRQ(ps2_IOP* p,uint32_t irq){
{
p->pending |= shl<uint64_t>(cast<uint64_t>(1ULL),cast<uint64_t>(irq));
p->raised[irq]++;
}
}
// tools/platform/ps2/iopintr.go:414:1
void ps2_IOP_serviceIntr(ps2_IOP* p){
{
if ((((p->inIntr > cast<int64_t>(0ULL)) || (!p->intrEnabled)) || (((p->pending == cast<uint64_t>(0ULL)) && (!p->vblankPending))))) {
return ;
}
{
mips_CPUState st = mips_CPU_SaveState(p->CPU);
if ((st.PendingDelay || (st.LdReg != cast<uint32_t>(0ULL)))) {
return ;
}
}
if (p->vblankPending) {
p->vblankPending = false;
ps2_IOP_vblankDeliver(p);
return ;
}
{uint32_t irq = cast<uint32_t>(0ULL);for (;(irq < cast<uint32_t>(64ULL));irq++){
uint64_t bit = shl<uint64_t>(cast<uint64_t>(1ULL),cast<uint64_t>(irq));
if ((((cast<uint64_t>((p->pending & bit)) == cast<uint64_t>(0ULL)) || (cast<uint64_t>((p->imask & bit)) == cast<uint64_t>(0ULL))) || (p->handlers[irq].fn == cast<uint32_t>(0ULL)))) {
continue;
}
p->pending &= ~(bit);
p->delivered[irq]++;
ps2_IOP_intrDeliver(p,irq);
return ;
}
}}
}
// tools/platform/ps2/iopintr.go:465:1
void ps2_IOP_intrDeliver(ps2_IOP* p,uint32_t irq){
{
ps2_iopHandler h = p->handlers[irq];
uint32_t frame = ((cast<uint32_t>((mips_CPU_Reg(p->CPU,cast<uint32_t>(29ULL)) - cast<uint32_t>(184ULL)))) & ~(cast<uint32_t>(7ULL)));
ps2_IOP_saveFrame(p,frame);
bool preemptible = (p->callDepth == cast<int64_t>(0ULL));
p->inIntr++;
p->intrEnabled = false;
ps2_IOP_ieEvent(p,std::string("deliver",7),frame,irq,mips_CPU_Reg(p->CPU,cast<uint32_t>(31ULL)));
mips_CPU_SetReg(p->CPU,cast<uint32_t>(4ULL),h.arg);
auto tmp209 = ps2_IOP_callGuestOn(p,h.fn,cast<uint32_t>(2063360ULL));
Error err = std::get<1>(tmp209);
uint32_t resume = frame;
if (((!err) && preemptible)) {
{
auto tmp210 = ps2_IOP_intrReschedule(p,frame);
uint32_t next = std::get<0>(tmp210);
bool ok = std::get<1>(tmp210);
if (ok) {
resume = next;
}
}
}
p->inIntr--;
ps2_IOP_loadFrame(p,resume);
if (bool(err)) {
ps2_IOP_halt(p,std::string("the handler for interrupt %d (%s) did not return: %v",52),irq,ps2_IOP_Sym(p,h.fn),err);
}
}
}
// tools/platform/ps2/iopintr.go:538:1
std::tuple<uint32_t,bool> ps2_IOP_intrReschedule(ps2_IOP* p,uint32_t frame){
{
if (((p->schedResched == cast<uint32_t>(0ULL)) || (p->schedSwitch == cast<uint32_t>(0ULL)))) {
return {cast<uint32_t>(0ULL),false};
}
auto tmp211 = ps2_IOP_callGuestOn(p,p->schedResched,cast<uint32_t>(2063360ULL));
uint32_t want = std::get<0>(tmp211);
Error err = std::get<1>(tmp211);
if ((bool(err) || (want == cast<uint32_t>(0ULL)))) {
return {cast<uint32_t>(0ULL),false};
}
mips_CPU_SetReg(p->CPU,cast<uint32_t>(4ULL),frame);
auto tmp212 = ps2_IOP_callGuestOn(p,p->schedSwitch,cast<uint32_t>(2063360ULL));
uint32_t next = std::get<0>(tmp212);
err = std::get<1>(tmp212);
if (bool(err)) {
ps2_IOP_halt(p,std::string("THREADMAN's thread switch did not return: %v",44),err);
return {cast<uint32_t>(0ULL),false};
}
if ((next == cast<uint32_t>(0ULL))) {
return {cast<uint32_t>(0ULL),false};
}
p->switches++;
return {next,true};
}
}
// tools/platform/ps2/iopintr.go:565:1
void ps2_IOP_tick(ps2_IOP* p){
{
ps2_IOP_ioTraceFlush(p);
if ((p->timerAck != cast<uint32_t>(0ULL))) {
ps2_IOP_timerAckFlush(p);
}
p->steps++;
{
uint32_t pc = ps2_iopPhys(p->CPU->PC);
if ((((pc < cast<uint32_t>(4096ULL)) && (pc != cast<uint32_t>(512ULL))) && (pc != cast<uint32_t>(516ULL)))) {
ps2_IOP_halt(p,std::string("jumped into the empty kernel area at 0x%08X, from %s\012%s",55),p->CPU->PC,ps2_IOP_Sym(p,p->lastPC),ps2_IOP_IOPTrail(p));
return ;
}
}
p->lastPC = p->CPU->PC;
p->trail[modi<int64_t>(p->trailN,cast<int64_t>(24ULL))] = p->CPU->PC;
p->trailN++;
if (((p->Trap != cast<uint32_t>(0ULL)) && (p->CPU->PC == p->Trap))) {
ps2_IOP_halt(p,std::string("reached the trap at %s\012%s",25),ps2_IOP_Sym(p,p->Trap),ps2_IOP_IOPTrail(p));
return ;
}
if (bool(p->OnCall)) {
{
auto tmp213 = lookup(p->stubName,ps2_iopPhys(p->CPU->PC));
std::string name = std::get<0>(tmp213);
bool ok = std::get<1>(tmp213);
if (ok) {
p->OnCall(name,std::array<uint32_t,4>{mips_CPU_Reg(p->CPU,cast<uint32_t>(4ULL)),mips_CPU_Reg(p->CPU,cast<uint32_t>(5ULL)),mips_CPU_Reg(p->CPU,cast<uint32_t>(6ULL)),mips_CPU_Reg(p->CPU,cast<uint32_t>(7ULL))},mips_CPU_Reg(p->CPU,cast<uint32_t>(31ULL)));
}
}
}
if ((modi<uint64_t>(p->steps,cast<uint64_t>(4096ULL)) == cast<uint64_t>(0ULL))) {
if ((!p->prof)) {
p->prof = Map<std::string,int64_t>{};
}
p->prof[ps2_IOP_symFunc(p,p->CPU->PC)]++;
}
ps2_IOP_timerTick(p);
if (p->cdvd->nBusy) {
ps2_cdvd_tick(p->cdvd,p);
}
if ((len(p->dmaPending) > cast<int64_t>(0ULL))) {
ps2_IOP_dmaTick(p);
}
if ((p->pending != cast<uint64_t>(0ULL))) {
ps2_IOP_serviceIntr(p);
}
}
}
// tools/platform/ps2/iopintr.go:618:1
uint32_t ps2_b2u(bool b){
{
if (b) {
return cast<uint32_t>(1ULL);
}
return cast<uint32_t>(0ULL);
}
}
// tools/platform/ps2/iopintr.go:668:1
void ps2_IOP_sysmemAlloc(ps2_IOP* p){
{
auto tmp214 = std::make_tuple(ps2_IOP_arg(p,cast<int64_t>(0ULL)),ps2_IOP_arg(p,cast<int64_t>(1ULL)),ps2_IOP_arg(p,cast<int64_t>(2ULL)));
uint32_t mode = std::get<0>(tmp214);
uint32_t size = std::get<1>(tmp214);
uint32_t addr = std::get<2>(tmp214);
size = ((cast<uint32_t>((size + cast<uint32_t>(63ULL)))) & ~(cast<uint32_t>(63ULL)));
if (((mode == cast<uint32_t>(2ULL)) && (addr != cast<uint32_t>(0ULL)))) {
p->blocks = append(p->blocks,Slice<ps2_iopBlock>{ps2_iopBlock{addr,size}});
ps2_IOP_setRet(p,addr);
return ;
}
uint32_t base = ps2_IOP_allocReuse(p,size,(mode == cast<uint32_t>(1ULL)));
if ((base == cast<uint32_t>(0ULL))) {
if ((mode == cast<uint32_t>(1ULL))) {
base = ps2_IOP_allocHigh(p,size);
}
else {
base = ps2_IOP_alloc(p,size);
}
}
if ((base == cast<uint32_t>(0ULL))) {
uint32_t free={};
{auto&& tmp215 = p->freeBlocks;
for(int64_t tmp216=0;tmp216<len(tmp215);++tmp216){
auto f=tmp215[tmp216];free += f.size;
}}
ps2_Machine_note(p->ps2,std::string("IOP: AllocSysMemory(mode=%d) could not find %d bytes; low=0x%X high=0x%X free-list=%d bytes in %d blocks",104),mode,size,p->allocPtr,p->allocHighPtr,free,len(p->freeBlocks));
ps2_IOP_setRet(p,cast<uint32_t>(0ULL));
return ;
}
p->blocks = append(p->blocks,Slice<ps2_iopBlock>{ps2_iopBlock{base,size}});
ps2_IOP_setRet(p,base);
}
}
// tools/platform/ps2/iopintr.go:703:1
uint32_t ps2_IOP_allocHigh(ps2_IOP* p,uint32_t size){
{
uint32_t top = p->allocHighPtr;
if ((top == cast<uint32_t>(0ULL))) {
top = cast<uint32_t>(2031616ULL);
}
uint32_t base = ((cast<uint32_t>((top - size))) & ~(cast<uint32_t>(63ULL)));
if ((base < p->allocPtr)) {
ps2_IOP_halt(p,std::string("out of IOP memory: %d bytes wanted from the high end, and it has reached the low allocations at 0x%08X",102),size,p->allocPtr);
return cast<uint32_t>(0ULL);
}
p->allocHighPtr = base;
return base;
}
}
// tools/platform/ps2/iopintr.go:722:1
uint32_t ps2_IOP_allocReuse(ps2_IOP* p,uint32_t size,bool high){
{
int64_t best = cast<int64_t>(-1ULL);
{auto&& tmp217 = p->freeBlocks;
for(int64_t tmp218=0;tmp218<len(tmp217);++tmp218){
auto i=tmp218;auto f=tmp217[tmp218];if (((f.size >= size) && (((best < cast<int64_t>(0ULL)) || (f.size < p->freeBlocks[best].size))))) {
best = i;
}
}}
if ((best < cast<int64_t>(0ULL))) {
return cast<uint32_t>(0ULL);
}
ps2_iopBlock f = p->freeBlocks[best];
if ((f.size == size)) {
p->freeBlocks = append(sub(p->freeBlocks,0,best),sub(p->freeBlocks,cast<int64_t>((best + cast<int64_t>(1ULL))),len(p->freeBlocks)));
return f.base;
}
if (high) {
p->freeBlocks[best] = ps2_iopBlock{f.base,cast<uint32_t>((f.size - size))};
return cast<uint32_t>((cast<uint32_t>((f.base + f.size)) - size));
}
p->freeBlocks[best] = ps2_iopBlock{cast<uint32_t>((f.base + size)),cast<uint32_t>((f.size - size))};
return f.base;
}
}
// tools/platform/ps2/iopintr.go:748:1
void ps2_IOP_sysmemFree(ps2_IOP* p){
{
uint32_t ptr = ps2_IOP_arg(p,cast<int64_t>(0ULL));
{auto&& tmp219 = p->blocks;
for(int64_t tmp220=0;tmp220<len(tmp219);++tmp220){
auto i=tmp220;auto b=tmp219[tmp220];if ((b.base == ptr)) {
p->blocks = append(sub(p->blocks,0,i),sub(p->blocks,cast<int64_t>((i + cast<int64_t>(1ULL))),len(p->blocks)));
ps2_IOP_freeInsert(p,b);
ps2_IOP_setRet(p,cast<uint32_t>(0ULL));
return ;
}
}}
ps2_IOP_setRet(p,cast<uint32_t>(0ULL));
}
}
// tools/platform/ps2/iopintr.go:773:1
void ps2_IOP_freeInsert(ps2_IOP* p,ps2_iopBlock b){
{
{;for (;;){
bool merged = false;
{auto&& tmp221 = p->freeBlocks;
for(int64_t tmp222=0;tmp222<len(tmp221);++tmp222){
auto i=tmp222;auto f=tmp221[tmp222];if ((cast<uint32_t>((f.base + f.size)) == b.base)) {
auto tmp223 = std::make_tuple(f.base,cast<uint32_t>((f.size + b.size)));
b.base = std::get<0>(tmp223);
b.size = std::get<1>(tmp223);
}
else if ((cast<uint32_t>((b.base + b.size)) == f.base)) {
b.size += f.size;
}
else {
continue;
}
p->freeBlocks = append(sub(p->freeBlocks,0,i),sub(p->freeBlocks,cast<int64_t>((i + cast<int64_t>(1ULL))),len(p->freeBlocks)));
merged = true;
break;
}}
if ((!merged)) {
break;
}
}
}if ((cast<uint32_t>((b.base + b.size)) == p->allocPtr)) {
p->allocPtr = b.base;
return ;
}
if (((p->allocHighPtr != cast<uint32_t>(0ULL)) && (b.base == p->allocHighPtr))) {
p->allocHighPtr = cast<uint32_t>((b.base + b.size));
return ;
}
p->freeBlocks = append(p->freeBlocks,Slice<ps2_iopBlock>{b});
}
}
// tools/platform/ps2/iopintr.go:812:1
void ps2_IOP_sysmemBlockTop(ps2_IOP* p){
{
{
auto tmp224 = ps2_IOP_blockOf(p,ps2_IOP_arg(p,cast<int64_t>(0ULL)));
ps2_iopBlock b = std::get<0>(tmp224);
bool ok = std::get<1>(tmp224);
if (ok) {
ps2_IOP_setRet(p,b.base);
return ;
}
}
ps2_IOP_setRet(p,cast<uint32_t>(0ULL));
}
}
// tools/platform/ps2/iopintr.go:820:1
void ps2_IOP_sysmemBlockSize(ps2_IOP* p){
{
{
auto tmp225 = ps2_IOP_blockOf(p,ps2_IOP_arg(p,cast<int64_t>(0ULL)));
ps2_iopBlock b = std::get<0>(tmp225);
bool ok = std::get<1>(tmp225);
if (ok) {
ps2_IOP_setRet(p,b.size);
return ;
}
}
ps2_IOP_setRet(p,cast<uint32_t>(0ULL));
}
}
// tools/platform/ps2/iopintr.go:828:1
std::tuple<ps2_iopBlock,bool> ps2_IOP_blockOf(ps2_IOP* p,uint32_t ptr){
{
{auto&& tmp226 = p->blocks;
for(int64_t tmp227=0;tmp227<len(tmp226);++tmp227){
auto b=tmp226[tmp227];if (((ptr >= b.base) && (ptr < cast<uint32_t>((b.base + b.size))))) {
return {b,true};
}
}}
return {ps2_iopBlock{},false};
}
}
// tools/platform/ps2/iopintr.go:838:1
void ps2_IOP_sysmemKprintf(ps2_IOP* p){
{
std::string s = ps2_IOP_formatArgs(p,ps2_IOP_CString(p,ps2_IOP_arg(p,cast<int64_t>(0ULL))),cast<int64_t>(1ULL));
ps2_Machine_iopPrint(p->ps2,s);
ps2_IOP_setRet(p,cast<uint32_t>(len(s)));
}
}
// tools/platform/ps2/iopintr.go:866:1
void ps2_IOP_heapCreate(ps2_IOP* p){
{
uint32_t chunk = ps2_IOP_arg(p,cast<int64_t>(0ULL));
uint32_t base = ps2_IOP_alloc(p,chunk);
if ((base == cast<uint32_t>(0ULL))) {
ps2_IOP_setRet(p,cast<uint32_t>(0ULL));
return ;
}
p->heaps[base] = arenaNew(ps2_iopHeap{chunk,base,chunk,base,{}});
ps2_IOP_setRet(p,base);
}
}
// tools/platform/ps2/iopintr.go:879:1
void ps2_IOP_heapAlloc(ps2_IOP* p){
{
auto tmp228 = std::make_tuple(get(p->heaps,ps2_IOP_arg(p,cast<int64_t>(0ULL))),ps2_IOP_arg(p,cast<int64_t>(1ULL)));
ps2_iopHeap* h = std::get<0>(tmp228);
uint32_t size = std::get<1>(tmp228);
if ((!h)) {
ps2_IOP_halt(p,std::string("AllocHeapMemory from heap 0x%08X, which was never created",57),ps2_IOP_arg(p,cast<int64_t>(0ULL)));
ps2_IOP_setRet(p,cast<uint32_t>(0ULL));
return ;
}
uint32_t a = ((cast<uint32_t>((h->ptr + cast<uint32_t>(7ULL)))) & ~(cast<uint32_t>(7ULL)));
if ((cast<uint32_t>((a + size)) > cast<uint32_t>((h->base + h->size)))) {
uint32_t n = h->chunk;
if ((size > n)) {
n = size;
}
uint32_t base = ps2_IOP_alloc(p,n);
if ((base == cast<uint32_t>(0ULL))) {
ps2_IOP_setRet(p,cast<uint32_t>(0ULL));
return ;
}
auto tmp229 = std::make_tuple(base,n,base);
h->base = std::get<0>(tmp229);
h->size = std::get<1>(tmp229);
h->ptr = std::get<2>(tmp229);
a = base;
}
h->ptr = cast<uint32_t>((a + size));
h->total += size;
ps2_IOP_setRet(p,a);
}
}
// tools/platform/ps2/iopintr.go:930:1
void ps2_IOP_clibMemcpy(ps2_IOP* p){
{
auto tmp230 = std::make_tuple(ps2_IOP_arg(p,cast<int64_t>(0ULL)),ps2_IOP_arg(p,cast<int64_t>(1ULL)),ps2_IOP_arg(p,cast<int64_t>(2ULL)));
uint32_t dst = std::get<0>(tmp230);
uint32_t src = std::get<1>(tmp230);
uint32_t n = std::get<2>(tmp230);
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < n);i++){
ps2_IOP_Write(p,cast<uint32_t>((dst + i)),ps2_IOP_Read(p,cast<uint32_t>((src + i))));
}
}ps2_IOP_setRet(p,dst);
}
}
// tools/platform/ps2/iopintr.go:947:1
void ps2_IOP_clibMemmove(ps2_IOP* p){
{
auto tmp231 = std::make_tuple(ps2_IOP_arg(p,cast<int64_t>(0ULL)),ps2_IOP_arg(p,cast<int64_t>(1ULL)),ps2_IOP_arg(p,cast<int64_t>(2ULL)));
uint32_t dst = std::get<0>(tmp231);
uint32_t src = std::get<1>(tmp231);
uint32_t n = std::get<2>(tmp231);
if (((dst > src) && (dst < cast<uint32_t>((src + n))))) {
{uint32_t i = n;for (;(i > cast<uint32_t>(0ULL));i--){
ps2_IOP_Write(p,cast<uint32_t>((cast<uint32_t>((dst + i)) - cast<uint32_t>(1ULL))),ps2_IOP_Read(p,cast<uint32_t>((cast<uint32_t>((src + i)) - cast<uint32_t>(1ULL)))));
}
}}
else {
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < n);i++){
ps2_IOP_Write(p,cast<uint32_t>((dst + i)),ps2_IOP_Read(p,cast<uint32_t>((src + i))));
}
}}
ps2_IOP_setRet(p,dst);
}
}
// tools/platform/ps2/iopintr.go:963:1
void ps2_IOP_clibBzero(ps2_IOP* p){
{
auto tmp232 = std::make_tuple(ps2_IOP_arg(p,cast<int64_t>(0ULL)),ps2_IOP_arg(p,cast<int64_t>(1ULL)));
uint32_t dst = std::get<0>(tmp232);
uint32_t n = std::get<1>(tmp232);
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < n);i++){
ps2_IOP_Write(p,cast<uint32_t>((dst + i)),cast<uint8_t>(0ULL));
}
}ps2_IOP_setRet(p,dst);
}
}
// tools/platform/ps2/iopintr.go:973:1
void ps2_IOP_clibStrcmp(ps2_IOP* p){
{
auto tmp233 = std::make_tuple(ps2_IOP_arg(p,cast<int64_t>(0ULL)),ps2_IOP_arg(p,cast<int64_t>(1ULL)));
uint32_t a = std::get<0>(tmp233);
uint32_t b = std::get<1>(tmp233);
{uint32_t i = cast<uint32_t>(0ULL);for (;;i++){
auto tmp234 = std::make_tuple(ps2_IOP_Read(p,cast<uint32_t>((a + i))),ps2_IOP_Read(p,cast<uint32_t>((b + i))));
uint8_t ca = std::get<0>(tmp234);
uint8_t cb = std::get<1>(tmp234);
if ((ca != cb)) {
ps2_IOP_setRet(p,cast<uint32_t>(cast<int32_t>((cast<int32_t>(ca) - cast<int32_t>(cb)))));
return ;
}
if ((ca == cast<uint8_t>(0ULL))) {
ps2_IOP_setRet(p,cast<uint32_t>(0ULL));
return ;
}
}
}}
}
// tools/platform/ps2/iopintr.go:990:1
void ps2_IOP_clibStrncpy(ps2_IOP* p){
{
auto tmp235 = std::make_tuple(ps2_IOP_arg(p,cast<int64_t>(0ULL)),ps2_IOP_arg(p,cast<int64_t>(1ULL)),ps2_IOP_arg(p,cast<int64_t>(2ULL)));
uint32_t dst = std::get<0>(tmp235);
uint32_t src = std::get<1>(tmp235);
uint32_t n = std::get<2>(tmp235);
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < n);i++){
uint8_t c = ps2_IOP_Read(p,cast<uint32_t>((src + i)));
ps2_IOP_Write(p,cast<uint32_t>((dst + i)),c);
if ((c == cast<uint8_t>(0ULL))) {
{;for (;(i < n);i++){
ps2_IOP_Write(p,cast<uint32_t>((dst + i)),cast<uint8_t>(0ULL));
}
}break;
}
}
}ps2_IOP_setRet(p,dst);
}
}
// tools/platform/ps2/iopintr.go:1011:1
void ps2_IOP_clibCtype(ps2_IOP* p){
{
uint8_t c = cast<uint8_t>(ps2_IOP_arg(p,cast<int64_t>(0ULL)));
uint32_t class_={};
if (((c >= cast<uint8_t>(48ULL)) && (c <= cast<uint8_t>(57ULL)))) {
class_ |= cast<uint32_t>(4ULL);
}
ps2_IOP_setRet(p,class_);
}
}
// tools/platform/ps2/iopintr.go:1032:1
void ps2_IOP_clibStrncmp(ps2_IOP* p){
{
auto tmp236 = std::make_tuple(ps2_IOP_arg(p,cast<int64_t>(0ULL)),ps2_IOP_arg(p,cast<int64_t>(1ULL)),ps2_IOP_arg(p,cast<int64_t>(2ULL)));
uint32_t a = std::get<0>(tmp236);
uint32_t b = std::get<1>(tmp236);
uint32_t n = std::get<2>(tmp236);
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < n);i++){
auto tmp237 = std::make_tuple(ps2_IOP_Read(p,cast<uint32_t>((a + i))),ps2_IOP_Read(p,cast<uint32_t>((b + i))));
uint8_t ca = std::get<0>(tmp237);
uint8_t cb = std::get<1>(tmp237);
if ((ca != cb)) {
ps2_IOP_setRet(p,cast<uint32_t>(cast<int32_t>((cast<int32_t>(ca) - cast<int32_t>(cb)))));
return ;
}
if ((ca == cast<uint8_t>(0ULL))) {
break;
}
}
}ps2_IOP_setRet(p,cast<uint32_t>(0ULL));
}
}
// tools/platform/ps2/iopintr.go:1048:1
void ps2_IOP_clibStrlen(ps2_IOP* p){
{
uint32_t s = ps2_IOP_arg(p,cast<int64_t>(0ULL));
uint32_t n = cast<uint32_t>(0ULL);
{;for (;((n < cast<uint32_t>(1024ULL)) && (ps2_IOP_Read(p,cast<uint32_t>((s + n))) != cast<uint8_t>(0ULL)));n++){
}
}ps2_IOP_setRet(p,n);
}
}
// tools/platform/ps2/iopintr.go:1057:1
void ps2_IOP_clibStrcat(ps2_IOP* p){
{
auto tmp238 = std::make_tuple(ps2_IOP_arg(p,cast<int64_t>(0ULL)),ps2_IOP_arg(p,cast<int64_t>(1ULL)));
uint32_t dst = std::get<0>(tmp238);
uint32_t src = std::get<1>(tmp238);
uint32_t end = dst;
{;for (;((cast<uint32_t>((end - dst)) < cast<uint32_t>(1024ULL)) && (ps2_IOP_Read(p,end) != cast<uint8_t>(0ULL)));end++){
}
}{uint32_t i = cast<uint32_t>(0ULL);for (;(i < cast<uint32_t>(1024ULL));i++){
uint8_t c = ps2_IOP_Read(p,cast<uint32_t>((src + i)));
ps2_IOP_Write(p,cast<uint32_t>((end + i)),c);
if ((c == cast<uint8_t>(0ULL))) {
break;
}
}
}ps2_IOP_setRet(p,dst);
}
}
// tools/platform/ps2/iopintr.go:1087:1
void ps2_IOP_clibStrcpy(ps2_IOP* p){
{
auto tmp239 = std::make_tuple(ps2_IOP_arg(p,cast<int64_t>(0ULL)),ps2_IOP_arg(p,cast<int64_t>(1ULL)));
uint32_t dst = std::get<0>(tmp239);
uint32_t src = std::get<1>(tmp239);
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < cast<uint32_t>(1024ULL));i++){
uint8_t c = ps2_IOP_Read(p,cast<uint32_t>((src + i)));
ps2_IOP_Write(p,cast<uint32_t>((dst + i)),c);
if ((c == cast<uint8_t>(0ULL))) {
break;
}
}
}ps2_IOP_setRet(p,dst);
}
}
// tools/platform/ps2/iopintr.go:1106:1
void ps2_IOP_clibStrchr(ps2_IOP* p){
{
auto tmp240 = std::make_tuple(ps2_IOP_arg(p,cast<int64_t>(0ULL)),cast<uint8_t>(ps2_IOP_arg(p,cast<int64_t>(1ULL))));
uint32_t s = std::get<0>(tmp240);
uint8_t c = std::get<1>(tmp240);
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < cast<uint32_t>(1024ULL));i++){
uint8_t ch = ps2_IOP_Read(p,cast<uint32_t>((s + i)));
if ((ch == c)) {
ps2_IOP_setRet(p,cast<uint32_t>((s + i)));
return ;
}
if ((ch == cast<uint8_t>(0ULL))) {
break;
}
}
}ps2_IOP_setRet(p,cast<uint32_t>(0ULL));
}
}
// tools/platform/ps2/iopintr.go:1122:1
void ps2_IOP_clibMemset(ps2_IOP* p){
{
auto tmp241 = std::make_tuple(ps2_IOP_arg(p,cast<int64_t>(0ULL)),ps2_IOP_arg(p,cast<int64_t>(1ULL)),ps2_IOP_arg(p,cast<int64_t>(2ULL)));
uint32_t dst = std::get<0>(tmp241);
uint32_t c = std::get<1>(tmp241);
uint32_t n = std::get<2>(tmp241);
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < n);i++){
ps2_IOP_Write(p,cast<uint32_t>((dst + i)),cast<uint8_t>(c));
}
}ps2_IOP_setRet(p,dst);
}
}
// tools/platform/ps2/iopintr.go:1159:1
bool ps2_IOP_kernelSyscall(ps2_IOP* p){
{
{
uint32_t svc = mips_CPU_Reg(p->CPU,cast<uint32_t>(2ULL));
switch(svc){
case cast<uint32_t>(32ULL):{
ps2_IOP_yield(p);
break;}
default:{
std::string name = go_fmt_Sprintf(std::string("kernel syscall %d",17),svc);
p->unmodelledCalls[name]++;
if ((get(p->unmodelledCalls,name) == cast<int64_t>(1ULL))) {
ps2_Machine_note(p->ps2,std::string("IOP: %s from %s \342\200\224 unmodelled",30),name,ps2_IOP_Sym(p,mips_CPU_CurPC(p->CPU)));
}
break;}
}}
return true;
}
}
// tools/platform/ps2/iopintr.go:1205:1
void ps2_IOP_yield(ps2_IOP* p){
{
uint32_t frame = ((cast<uint32_t>((mips_CPU_Reg(p->CPU,cast<uint32_t>(29ULL)) - cast<uint32_t>(184ULL)))) & ~(cast<uint32_t>(7ULL)));
ps2_IOP_saveFrame(p,frame);
ps2_IOP_Write32(p,cast<uint32_t>((frame + cast<uint32_t>(136ULL))),shl<uint32_t>(ps2_b2u((mips_CPU_Reg(p->CPU,cast<uint32_t>(6ULL)) != cast<uint32_t>(0ULL))),cast<int64_t>(2ULL)));
ps2_IOP_Write32(p,cast<uint32_t>((frame + cast<uint32_t>(8ULL))),mips_CPU_Reg(p->CPU,cast<uint32_t>(4ULL)));
ps2_IOP_ieEvent(p,std::string("yield",5),frame,mips_CPU_Reg(p->CPU,cast<uint32_t>(6ULL)),mips_CPU_Reg(p->CPU,cast<uint32_t>(31ULL)));
uint32_t resume = frame;
{
auto tmp242 = ps2_IOP_intrReschedule(p,frame);
uint32_t next = std::get<0>(tmp242);
bool ok = std::get<1>(tmp242);
if (ok) {
resume = next;
}
}
ps2_IOP_loadFrame(p,resume);
}
}
// tools/platform/ps2/iopkernel.go:70:1
void ps2_lib(std::string name,Map<uint16_t,ps2_iopFunc> funcs){
{
ps2_goLibraries[name] = arenaNew(ps2_iopLibrary{funcs});
}
}
// tools/platform/ps2/iopkernel.go:77:1
ps2_iopFunc ps2_unknown(){
{
return ps2_iopFunc{};
}
}
// tools/platform/ps2/iopkernel.go:79:1
void ps2_init_iopkernel(){
{
ps2_lib(std::string("loadcore",8),Map<uint16_t,ps2_iopFunc>{{cast<uint16_t>(3ULL),ps2_unknown()},{cast<uint16_t>(4ULL),ps2_iopFunc{std::string("FlushIcache",11),ps2_IOP_loadcoreFlushIcache}},{cast<uint16_t>(5ULL),ps2_unknown()},{cast<uint16_t>(6ULL),ps2_iopFunc{std::string("RegisterLibraryEntries",22),ps2_IOP_loadcoreRegisterLibrary}},{cast<uint16_t>(8ULL),ps2_iopFunc{std::string("LinkImports",11),ps2_IOP_loadcoreLinkCheck}},{cast<uint16_t>(9ULL),ps2_unknown()},{cast<uint16_t>(10ULL),ps2_unknown()},{cast<uint16_t>(12ULL),ps2_unknown()},{cast<uint16_t>(16ULL),ps2_iopFunc{std::string("RegisterModule",14),ps2_IOP_loadcoreRegisterModule}},{cast<uint16_t>(17ULL),ps2_unknown()},{cast<uint16_t>(20ULL),ps2_iopFunc{std::string("RegisterBootCallback",20),ps2_IOP_loadcoreRegisterBootCallback}},{cast<uint16_t>(21ULL),ps2_unknown()},{cast<uint16_t>(22ULL),ps2_iopFunc{std::string("ProbeModule",11),ps2_IOP_loadcoreProbeModule}},{cast<uint16_t>(23ULL),ps2_iopFunc{std::string("LinkModule",10),ps2_IOP_loadcoreLinkModule}},{cast<uint16_t>(24ULL),ps2_unknown()}});
ps2_lib(std::string("stdio",5),Map<uint16_t,ps2_iopFunc>{{cast<uint16_t>(4ULL),ps2_iopFunc{std::string("printf",6),ps2_IOP_stdioPrintf}}});
ps2_lib(std::string("dmacman",7),Map<uint16_t,ps2_iopFunc>{{cast<uint16_t>(28ULL),ps2_iopFunc{std::string("SetSliceDMA",11),ps2_IOP_dmacmanSetSlice}},{cast<uint16_t>(32ULL),ps2_iopFunc{std::string("StartDMA",8),ps2_IOP_dmacmanStart}},{cast<uint16_t>(33ULL),ps2_iopFunc{std::string("ChanSetup",9),ps2_IOP_dmacmanChanSetup}},{cast<uint16_t>(34ULL),ps2_iopFunc{std::string("ChanEnable",10),ps2_IOP_dmacmanChanEnable}},{cast<uint16_t>(35ULL),ps2_unknown()}});
ps2_lib(std::string("vblank",6),Map<uint16_t,ps2_iopFunc>{{cast<uint16_t>(8ULL),ps2_iopFunc{std::string("RegisterVblankHandler",21),ps2_IOP_vblankRegister}},{cast<uint16_t>(9ULL),ps2_iopFunc{std::string("ReleaseVblankHandler",20),ps2_IOP_vblankRelease}}});
ps2_lib(std::string("secrman",7),Map<uint16_t,ps2_iopFunc>{{cast<uint16_t>(4ULL),ps2_unknown()},{cast<uint16_t>(5ULL),ps2_unknown()},{cast<uint16_t>(6ULL),ps2_unknown()}});
}
}
// tools/platform/ps2/iopkernel.go:159:1
void ps2_IOP_registerLibraries(ps2_IOP* p){
{
p->calls = append(p->calls,Slice<ps2_iopCall*>{arenaNew(ps2_iopCall{std::string("<invalid>",9),{}})});
p->bound = Map<ps2_iopBinding,uint32_t>{};
}
}
// tools/platform/ps2/iopkernel.go:171:1
bool ps2_IOP_hasGoLibrary(ps2_IOP* p,std::string name){
{
auto tmp243 = lookup(ps2_goLibraries,name);
bool ok = std::get<1>(tmp243);
return ok;
}
}
// tools/platform/ps2/iopkernel.go:179:1
uint32_t ps2_IOP_bindCall(ps2_IOP* p,std::string library,uint16_t id){
{
ps2_iopBinding key = ps2_iopBinding{library,id};
{
auto tmp244 = lookup(p->bound,key);
uint32_t code = std::get<0>(tmp244);
bool ok = std::get<1>(tmp244);
if (ok) {
return code;
}
}
ps2_iopLibrary* l = get(ps2_goLibraries,library);
auto tmp245 = lookup(l->funcs,id);
ps2_iopFunc f = std::get<0>(tmp245);
bool known = std::get<1>(tmp245);
std::string name = go_fmt_Sprintf(std::string("%s#%d",5),library,id);
if ((known && (f.name != std::string("",0)))) {
name = go_fmt_Sprintf(std::string("%s.%s",5),library,f.name);
}
uint32_t code = ps2_IOP_addCall(p,name,f.fn);
p->bound[key] = code;
return code;
}
}
// tools/platform/ps2/iopkernel.go:197:1
uint32_t ps2_IOP_addCall(ps2_IOP* p,std::string name,std::function<void(ps2_IOP*)> fn){
{
p->calls = append(p->calls,Slice<ps2_iopCall*>{arenaNew(ps2_iopCall{name,fn})});
return cast<uint32_t>(cast<int64_t>((len(p->calls) - cast<int64_t>(1ULL))));
}
}
// tools/platform/ps2/iopkernel.go:208:1
bool ps2_IOP_handleSyscall(ps2_IOP* p,mips_CPU* c){
{
uint32_t code = cast<uint32_t>(((shr<uint32_t>(ps2_IOP_Read32(p,mips_CPU_CurPC(c)),cast<int64_t>(6ULL))) & cast<uint32_t>(1048575ULL)));
if ((code == cast<uint32_t>(0ULL))) {
return ps2_IOP_kernelSyscall(p);
}
if ((cast<int64_t>(code) >= len(p->calls))) {
return false;
}
ps2_iopCall* call = p->calls[code];
if ((!call->fn)) {
p->unmodelledCalls[call->name]++;
if ((get(p->unmodelledCalls,call->name) == cast<int64_t>(1ULL))) {
ps2_Machine_note(p->ps2,std::string("IOP: %s(0x%X, 0x%X, 0x%X, 0x%X) from %s \342\200\224 unmodelled",54),call->name,ps2_IOP_arg(p,cast<int64_t>(0ULL)),ps2_IOP_arg(p,cast<int64_t>(1ULL)),ps2_IOP_arg(p,cast<int64_t>(2ULL)),ps2_IOP_arg(p,cast<int64_t>(3ULL)),ps2_IOP_Sym(p,mips_CPU_Reg(p->CPU,cast<uint32_t>(31ULL))));
}
ps2_IOP_setRet(p,cast<uint32_t>(0ULL));
return true;
}
call->fn(p);
return true;
}
}
// tools/platform/ps2/iopkernel.go:240:1
uint32_t ps2_IOP_arg(ps2_IOP* p,int64_t i){
{
if ((i < cast<int64_t>(4ULL))) {
return mips_CPU_Reg(p->CPU,cast<uint32_t>(cast<int64_t>((cast<int64_t>(4ULL) + i))));
}
return ps2_IOP_Read32(p,cast<uint32_t>((mips_CPU_Reg(p->CPU,cast<uint32_t>(29ULL)) + cast<uint32_t>(cast<int64_t>((cast<int64_t>(4ULL) * i))))));
}
}
// tools/platform/ps2/iopkernel.go:247:1
void ps2_IOP_setRet(ps2_IOP* p,uint32_t v){
{
mips_CPU_SetReg(p->CPU,cast<uint32_t>(2ULL),v);
}
}
// tools/platform/ps2/iopkernel.go:258:1
void ps2_IOP_loadcoreRegisterLibrary(ps2_IOP* p){
{
ps2_IOP_setRet(p,cast<uint32_t>(0ULL));
}
}
// tools/platform/ps2/iopkernel.go:286:1
void ps2_IOP_loadcoreRegisterBootCallback(ps2_IOP* p){
{
p->bootCallbacks = append(p->bootCallbacks,Slice<uint32_t>{ps2_IOP_arg(p,cast<int64_t>(0ULL))});
ps2_IOP_setRet(p,cast<uint32_t>(0ULL));
}
}
// tools/platform/ps2/iopkernel.go:295:1
void ps2_IOP_stdioPrintf(ps2_IOP* p){
{
std::string s = ps2_IOP_formatArgs(p,ps2_IOP_CString(p,ps2_IOP_arg(p,cast<int64_t>(0ULL))),cast<int64_t>(1ULL));
ps2_Machine_iopPrint(p->ps2,s);
ps2_IOP_setRet(p,cast<uint32_t>(len(s)));
}
}
// tools/platform/ps2/iopkernel.go:306:1
std::string ps2_IOP_formatArgs(ps2_IOP* p,std::string format,int64_t argi){
{
Slice<uint8_t> out={};
auto next = [&]()->uint32_t{
uint32_t v = ps2_IOP_arg(p,argi);
argi++;
return v;
}
;
{int64_t i = cast<int64_t>(0ULL);for (;(i < len(format));i++){
if (((cast<uint8_t>(format[i]) != cast<uint8_t>(37ULL)) || (cast<int64_t>((i + cast<int64_t>(1ULL))) >= len(format)))) {
out = append(out,Slice<uint8_t>{cast<uint8_t>(format[i])});
continue;
}
int64_t j = cast<int64_t>((i + cast<int64_t>(1ULL)));
{;for (;((j < len(format)) && ((((((((((cast<uint8_t>(format[j]) == cast<uint8_t>(45ULL)) || (cast<uint8_t>(format[j]) == cast<uint8_t>(43ULL))) || (cast<uint8_t>(format[j]) == cast<uint8_t>(32ULL))) || (cast<uint8_t>(format[j]) == cast<uint8_t>(35ULL))) || (cast<uint8_t>(format[j]) == cast<uint8_t>(48ULL))) || (cast<uint8_t>(format[j]) == cast<uint8_t>(46ULL))) || (((cast<uint8_t>(format[j]) >= cast<uint8_t>(49ULL)) && (cast<uint8_t>(format[j]) <= cast<uint8_t>(57ULL))))) || (cast<uint8_t>(format[j]) == cast<uint8_t>(108ULL))) || (cast<uint8_t>(format[j]) == cast<uint8_t>(104ULL)))));){
j++;
}
}if ((j >= len(format))) {
out = append(out,sub(format,i,len(format)));
break;
}
uint8_t verb = cast<uint8_t>(format[j]);
std::string spec = sub(format,i,cast<int64_t>((j + cast<int64_t>(1ULL))));
{
switch(verb){
case cast<uint8_t>(37ULL):{
out = append(out,Slice<uint8_t>{cast<uint8_t>(37ULL)});
break;}
case cast<uint8_t>(100ULL):case cast<uint8_t>(105ULL):{
out = append(out,go_fmt_Sprintf(std::string("%d",2),cast<int32_t>(next())));
break;}
case cast<uint8_t>(117ULL):{
out = append(out,go_fmt_Sprintf(std::string("%d",2),next()));
break;}
case cast<uint8_t>(120ULL):case cast<uint8_t>(88ULL):case cast<uint8_t>(111ULL):{
out = append(out,go_fmt_Sprintf(cast<std::string>(Slice<uint8_t>{cast<uint8_t>(37ULL),verb}),next()));
break;}
case cast<uint8_t>(99ULL):{
out = append(out,Slice<uint8_t>{cast<uint8_t>(next())});
break;}
case cast<uint8_t>(112ULL):{
out = append(out,go_fmt_Sprintf(std::string("0x%08X",6),next()));
break;}
case cast<uint8_t>(115ULL):{
out = append(out,ps2_IOP_CString(p,next()));
break;}
default:{
out = append(out,spec);
break;}
}}
i = j;
}
}return cast<std::string>(out);
}
}
// tools/platform/ps2/iopload.go:51:1
uint32_t ps2_insnJ(uint32_t target){
{
return cast<uint32_t>((cast<uint32_t>(134217728ULL) | cast<uint32_t>(((shr<uint32_t>(target,cast<int64_t>(2ULL))) & cast<uint32_t>(67108863ULL)))));
}
}
// tools/platform/ps2/iopload.go:52:1
uint32_t ps2_insnSyscall(uint32_t n){
{
return cast<uint32_t>(((shl<uint32_t>(n,cast<int64_t>(6ULL))) | cast<uint32_t>(12ULL)));
}
}
// tools/platform/ps2/iopload.go:53:1
uint32_t ps2_insnNop(){
{
return cast<uint32_t>(0ULL);
}
}
// tools/platform/ps2/iopload.go:54:1
uint32_t ps2_insnJR(uint32_t reg){
{
return cast<uint32_t>(((shl<uint32_t>(reg,cast<int64_t>(21ULL))) | cast<uint32_t>(8ULL)));
}
}
// tools/platform/ps2/iopload.go:55:1
uint32_t ps2_regRA(){
{
return cast<uint32_t>(31ULL);
}
}
// tools/platform/ps2/iopload.go:65:1
std::tuple<ps2_IOPModule*,Error> ps2_IOP_LoadIRX(ps2_IOP* p,std::string name,Slice<uint8_t> raw){
{
auto tmp246 = ps2_ReadIRX(raw);
ps2_IRX* x = std::get<0>(tmp246);
Error err = std::get<1>(tmp246);
if (bool(err)) {
return {{},go_fmt_Errorf(std::string("%s: %w",6),name,err)};
}
uint32_t base = ps2_IOP_alloc(p,x->MemSz);
return ps2_IOP_placeAndLink(p,name,x,base);
}
}
// tools/platform/ps2/iopload.go:78:1
std::tuple<ps2_IOPModule*,Error> ps2_IOP_placeAndLink(ps2_IOP* p,std::string name,ps2_IRX* x,uint32_t base){
{
auto tmp247 = ps2_IRX_Relocate(x,base);
Slice<uint8_t> img = std::get<0>(tmp247);
Error err = std::get<1>(tmp247);
if (bool(err)) {
return {{},go_fmt_Errorf(std::string("%s: %w",6),name,err)};
}
gcopy(sub(p->ram,base,len(p->ram)),img);
ps2_IOPModule* mod = arenaNew(ps2_IOPModule{name,base,x->MemSz,x});
p->modules = append(p->modules,Slice<ps2_IOPModule*>{mod});
{auto&& tmp248 = x->Exports;
for(int64_t tmp249=0;tmp249<len(tmp248);++tmp249){
auto i=tmp249;ps2_IRXExport* e = (&x->Exports[i]);
{
auto tmp250 = lookup(p->owner,e->Library);
std::string prev = std::get<0>(tmp250);
bool ok = std::get<1>(tmp250);
if ((ok && (prev != name))) {
ps2_Machine_note(p->ps2,std::string("IOP: %s re-exports %s, which %s already provides",48),name,e->Library,prev);
}
}
p->exports[e->Library] = e;
p->owner[e->Library] = name;
}}
{
Error err = ps2_IOP_link(p,mod);
if (bool(err)) {
return {{},err};
}
}
ps2_IOP_resolveTrap(p);
p->running = true;
return {mod,{}};
}
}
// tools/platform/ps2/iopload.go:111:1
Error ps2_IOP_link(ps2_IOP* p,ps2_IOPModule* mod){
{
{auto&& tmp251 = mod->IRX->Imports;
for(int64_t tmp252=0;tmp252<len(tmp251);++tmp252){
auto imp=tmp251[tmp252];auto tmp253 = lookup(p->exports,imp.Library);
ps2_IRXExport* exp = std::get<0>(tmp253);
bool onDisc = std::get<1>(tmp253);
{auto&& tmp254 = imp.Stubs;
for(int64_t tmp255=0;tmp255<len(tmp254);++tmp255){
auto i=tmp255;auto stub=tmp254[tmp255];uint16_t id = imp.IDs[i];
uint32_t addr = cast<uint32_t>((mod->Base + stub));
p->stubName[addr] = go_fmt_Sprintf(std::string("%s#%d",5),imp.Library,id);
{
if (onDisc){
if ((cast<int64_t>(id) >= len(exp->Entries))) {
return go_fmt_Errorf(std::string("%s imports %s function %d, but %s's export table has only %d",60),mod->Name,imp.Library,id,get(p->owner,imp.Library),len(exp->Entries));
}
uint32_t target = ps2_IOP_exportAddr(p,imp.Library,id);
if ((target == cast<uint32_t>(0ULL))) {
return go_fmt_Errorf(std::string("%s imports %s function %d, which %s exports as a null pointer",61),mod->Name,imp.Library,id,get(p->owner,imp.Library));
}
ps2_IOP_Write32(p,cast<uint32_t>((addr + cast<uint32_t>(0ULL))),ps2_insnJ(target));
ps2_IOP_Write32(p,cast<uint32_t>((addr + cast<uint32_t>(4ULL))),ps2_insnNop());
}
else if (ps2_IOP_hasGoLibrary(p,imp.Library)){
uint32_t code = ps2_IOP_bindCall(p,imp.Library,id);
ps2_IOP_Write32(p,cast<uint32_t>((addr + cast<uint32_t>(0ULL))),ps2_insnJR(ps2_regRA()));
ps2_IOP_Write32(p,cast<uint32_t>((addr + cast<uint32_t>(4ULL))),ps2_insnSyscall(code));
}
else {
return go_fmt_Errorf(std::string("%s imports library %q, which nothing on the disc exports and nothing in Go models",81),mod->Name,imp.Library);
}
}
tmp256:;
}}
}}
return {};
}
}
// tools/platform/ps2/iopload.go:217:1
void ps2_IOP_loadcoreProbeModule(ps2_IOP* p){
{
auto tmp257 = std::make_tuple(ps2_IOP_arg(p,cast<int64_t>(0ULL)),ps2_IOP_arg(p,cast<int64_t>(1ULL)));
uint32_t image = std::get<0>(tmp257);
uint32_t info = std::get<1>(tmp257);
auto tmp258 = ps2_IOP_readGuestIRX(p,image);
ps2_IRX* x = std::get<0>(tmp258);
Error err = std::get<1>(tmp258);
if (bool(err)) {
ps2_Machine_note(p->ps2,std::string("IOP: loadcore#22: %s is not a loadable module: %v",49),ps2_IOP_Sym(p,image),err);
ps2_IOP_setRet(p,cast<uint32_t>(4294967095ULL));
return ;
}
ps2_IOP_Write32(p,cast<uint32_t>((info + cast<uint32_t>(28ULL))),x->MemSz);
ps2_IOP_setRet(p,cast<uint32_t>(2ULL));
}
}
// tools/platform/ps2/iopload.go:232:1
void ps2_IOP_loadcoreLinkModule(ps2_IOP* p){
{
auto tmp259 = std::make_tuple(ps2_IOP_arg(p,cast<int64_t>(0ULL)),ps2_IOP_arg(p,cast<int64_t>(1ULL)));
uint32_t image = std::get<0>(tmp259);
uint32_t info = std::get<1>(tmp259);
uint32_t base = ps2_IOP_Read32(p,cast<uint32_t>((info + cast<uint32_t>(12ULL))));
auto tmp260 = ps2_IOP_readGuestIRX(p,image);
ps2_IRX* x = std::get<0>(tmp260);
Error err = std::get<1>(tmp260);
if (bool(err)) {
ps2_Machine_note(p->ps2,std::string("IOP: loadcore#23: %s: %v",24),ps2_IOP_Sym(p,image),err);
ps2_IOP_setRet(p,cast<uint32_t>(4294967095ULL));
return ;
}
auto tmp261 = ps2_IOP_placeAndLink(p,ps2_IOP_guestModuleName(p,image,x),x,base);
ps2_IOPModule* mod = std::get<0>(tmp261);
err = std::get<1>(tmp261);
if (bool(err)) {
ps2_Machine_note(p->ps2,std::string("IOP: loadcore#23: %v",20),err);
ps2_IOP_setRet(p,cast<uint32_t>(4294967095ULL));
return ;
}
uint32_t rec = cast<uint32_t>((base - cast<uint32_t>(48ULL)));
ps2_IOP_Write32(p,cast<uint32_t>((rec + cast<uint32_t>(16ULL))),cast<uint32_t>((mod->Base + x->Entry)));
ps2_IOP_Write32(p,cast<uint32_t>((rec + cast<uint32_t>(20ULL))),cast<uint32_t>((mod->Base + x->GP)));
ps2_IOP_Write32(p,cast<uint32_t>((rec + cast<uint32_t>(24ULL))),mod->Base);
ps2_IOP_Write32(p,cast<uint32_t>((rec + cast<uint32_t>(28ULL))),mod->Size);
ps2_IOP_Write32(p,cast<uint32_t>((info + cast<uint32_t>(16ULL))),mod->Base);
ps2_Machine_note(p->ps2,std::string("IOP: MODLOAD loaded %s at 0x%08X (%d KiB) via loadcore",54),mod->Name,mod->Base,divi<uint32_t>(mod->Size,cast<uint32_t>(1024ULL)));
ps2_IOP_setRet(p,cast<uint32_t>(0ULL));
}
}
// tools/platform/ps2/iopload.go:267:1
void ps2_IOP_loadcoreLinkCheck(ps2_IOP* p){
{
ps2_IOP_setRet(p,cast<uint32_t>(0ULL));
}
}
// tools/platform/ps2/iopload.go:272:1
void ps2_IOP_loadcoreFlushIcache(ps2_IOP* p){
{
ps2_IOP_setRet(p,cast<uint32_t>(0ULL));
}
}
// tools/platform/ps2/iopload.go:277:1
void ps2_IOP_loadcoreRegisterModule(ps2_IOP* p){
{
ps2_IOP_setRet(p,cast<uint32_t>(0ULL));
}
}
// tools/platform/ps2/iopload.go:287:1
std::tuple<ps2_IRX*,Error> ps2_IOP_readGuestIRX(ps2_IOP* p,uint32_t addr){
{
{
uint32_t m = ps2_IOP_Read32(p,addr);
if ((m != cast<uint32_t>(1179403647ULL))) {
return {{},go_fmt_Errorf(std::string("no ELF magic at 0x%08X (found 0x%08X)",37),addr,m)};
}
}
auto rd16 = [&](uint32_t o)->uint32_t{
return cast<uint32_t>((cast<uint32_t>(ps2_IOP_Read(p,cast<uint32_t>((addr + o)))) | shl<uint32_t>(cast<uint32_t>(ps2_IOP_Read(p,cast<uint32_t>((cast<uint32_t>((addr + o)) + cast<uint32_t>(1ULL))))),cast<int64_t>(8ULL))));
}
;
auto tmp262 = std::make_tuple(ps2_IOP_Read32(p,cast<uint32_t>((addr + cast<uint32_t>(28ULL)))),rd16(cast<uint32_t>(42ULL)),rd16(cast<uint32_t>(44ULL)));
uint32_t phoff = std::get<0>(tmp262);
uint32_t phentsize = std::get<1>(tmp262);
uint32_t phnum = std::get<2>(tmp262);
auto tmp263 = std::make_tuple(ps2_IOP_Read32(p,cast<uint32_t>((addr + cast<uint32_t>(32ULL)))),rd16(cast<uint32_t>(46ULL)),rd16(cast<uint32_t>(48ULL)));
uint32_t shoff = std::get<0>(tmp263);
uint32_t shentsize = std::get<1>(tmp263);
uint32_t shnum = std::get<2>(tmp263);
uint32_t size = cast<uint32_t>((phoff + cast<uint32_t>((phentsize * phnum))));
{
uint32_t e = cast<uint32_t>((shoff + cast<uint32_t>((shentsize * shnum))));
if ((e > size)) {
size = e;
}
}
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < shnum);i++){
uint32_t sh = cast<uint32_t>((cast<uint32_t>((addr + shoff)) + cast<uint32_t>((i * shentsize))));
if ((ps2_IOP_Read32(p,cast<uint32_t>((sh + cast<uint32_t>(4ULL)))) == cast<uint32_t>(8ULL))) {
continue;
}
{
uint32_t e = cast<uint32_t>((ps2_IOP_Read32(p,cast<uint32_t>((sh + cast<uint32_t>(16ULL)))) + ps2_IOP_Read32(p,cast<uint32_t>((sh + cast<uint32_t>(20ULL))))));
if ((e > size)) {
size = e;
}
}
}
}{uint32_t i = cast<uint32_t>(0ULL);for (;(i < phnum);i++){
uint32_t ph = cast<uint32_t>((cast<uint32_t>((addr + phoff)) + cast<uint32_t>((i * phentsize))));
{
uint32_t e = cast<uint32_t>((ps2_IOP_Read32(p,cast<uint32_t>((ph + cast<uint32_t>(4ULL)))) + ps2_IOP_Read32(p,cast<uint32_t>((ph + cast<uint32_t>(16ULL))))));
if ((e > size)) {
size = e;
}
}
}
}if (((size == cast<uint32_t>(0ULL)) || (cast<uint32_t>((addr + size)) > cast<uint32_t>(2097152ULL)))) {
return {{},go_fmt_Errorf(std::string("the ELF header at 0x%08X gives an implausible file size of %d bytes",67),addr,size)};
}
Slice<uint8_t> raw = Slice<uint8_t>::make(size);
gcopy(raw,sub(p->ram,addr,cast<uint32_t>((addr + size))));
return ps2_ReadIRX(raw);
}
}
// tools/platform/ps2/iopload.go:326:1
std::string ps2_IOP_guestModuleName(ps2_IOP* p,uint32_t addr,ps2_IRX* x){
{
if ((x->Name != std::string("",0))) {
return x->Name;
}
return go_fmt_Sprintf(std::string("module@0x%08X",13),addr);
}
}
// tools/platform/ps2/iopload.go:334:1
uint32_t ps2_IOP_exportAddr(ps2_IOP* p,std::string library,uint16_t id){
{
ps2_IRXExport* exp = get(p->exports,library);
std::string owner = get(p->owner,library);
{auto&& tmp264 = p->modules;
for(int64_t tmp265=0;tmp265<len(tmp264);++tmp265){
auto m=tmp264[tmp265];if ((m->Name == owner)) {
return cast<uint32_t>((m->Base + exp->Entries[id]));
}
}}
return cast<uint32_t>(0ULL);
}
}
// tools/platform/ps2/iopload.go:352:1
std::tuple<uint32_t,Error> ps2_IOP_Start(ps2_IOP* p,ps2_IOPModule* mod,Slice<uint32_t> args){
{
uint32_t entry = cast<uint32_t>((mod->Base + mod->IRX->Entry));
uint32_t gp = cast<uint32_t>((mod->Base + mod->IRX->GP));
uint32_t argv = cast<uint32_t>(0ULL);
if ((len(args) > cast<int64_t>(0ULL))) {
argv = ps2_IOP_alloc(p,cast<uint32_t>(cast<int64_t>((cast<int64_t>(4ULL) * len(args)))));
{auto&& tmp266 = args;
for(int64_t tmp267=0;tmp267<len(tmp266);++tmp267){
auto i=tmp267;auto a=tmp266[tmp267];ps2_IOP_Write32(p,cast<uint32_t>((argv + cast<uint32_t>(cast<int64_t>((cast<int64_t>(4ULL) * i))))),a);
}}
}
mips_CPU_SetReg(p->CPU,cast<uint32_t>(4ULL),cast<uint32_t>(len(args)));
mips_CPU_SetReg(p->CPU,cast<uint32_t>(5ULL),argv);
mips_CPU_SetReg(p->CPU,cast<uint32_t>(6ULL),cast<uint32_t>(0ULL));
mips_CPU_SetReg(p->CPU,cast<uint32_t>(7ULL),cast<uint32_t>(0ULL));
mips_CPU_SetReg(p->CPU,cast<uint32_t>(28ULL),gp);
auto tmp268 = ps2_IOP_callGuest(p,entry);
uint32_t res = std::get<0>(tmp268);
Error err = std::get<1>(tmp268);
if (bool(err)) {
return {cast<uint32_t>(0ULL),go_fmt_Errorf(std::string("starting %s: %w",15),mod->Name,err)};
}
return {res,{}};
}
}
// tools/platform/ps2/iopload.go:403:1
std::tuple<uint32_t,Error> ps2_IOP_callGuest(ps2_IOP* p,uint32_t entry){
{
return ps2_IOP_callGuestOn(p,entry,cast<uint32_t>(2096128ULL));
}
}
// tools/platform/ps2/iopload.go:410:1
std::tuple<uint32_t,Error> ps2_IOP_callGuestOn(ps2_IOP* p,uint32_t entry,uint32_t stack){
{
p->callDepth++;
auto tmp269=defer([&](){[&]()->void{
p->callDepth--;
}
();});
uint32_t saved = p->CPU->PC;
uint32_t savedSP = mips_CPU_Reg(p->CPU,cast<uint32_t>(29ULL));
uint32_t savedRA = mips_CPU_Reg(p->CPU,cast<uint32_t>(31ULL));
mips_CPU_SetReg(p->CPU,cast<uint32_t>(29ULL),stack);
mips_CPU_SetReg(p->CPU,cast<uint32_t>(31ULL),cast<uint32_t>(268369920ULL));
mips_CPU_SetPC(p->CPU,entry);
Map<uint32_t,int64_t> seen = Map<uint32_t,int64_t>{};
int64_t acc = cast<int64_t>(0ULL);
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(200000000ULL));i++){
if ((p->CPU->PC == cast<uint32_t>(268369920ULL))) {
uint32_t v = mips_CPU_Reg(p->CPU,cast<uint32_t>(2ULL));
mips_CPU_SetPC(p->CPU,saved);
mips_CPU_SetReg(p->CPU,cast<uint32_t>(29ULL),savedSP);
mips_CPU_SetReg(p->CPU,cast<uint32_t>(31ULL),savedRA);
return {v,{}};
}
if ((p->Halted || p->CPU->Halted)) {
return {cast<uint32_t>(0ULL),go_fmt_Errorf(std::string("the IOP stopped at %s: %s",25),ps2_IOP_Sym(p,p->CPU->PC),ps2_IOP_reason(p))};
}
seen[p->CPU->PC]++;
acc++;
if ((acc >= cast<int64_t>(1048576ULL))) {
if ((len(seen) <= cast<int64_t>(8ULL))) {
return {cast<uint32_t>(0ULL),go_fmt_Errorf(std::string("%s is spinning on %d addresses around %s \342\200\224 waiting for something that never happens",85),ps2_IOP_Sym(p,entry),len(seen),ps2_IOP_Sym(p,p->CPU->PC))};
}
seen = Map<uint32_t,int64_t>{};
acc = cast<int64_t>(0ULL);
}
ps2_IOP_tick(p);
mips_CPU_Step(p->CPU);
}
}return {cast<uint32_t>(0ULL),go_fmt_Errorf(std::string("the routine at %s ran %d instructions without returning, and is at %s",69),ps2_IOP_Sym(p,entry),cast<int64_t>(200000000ULL),ps2_IOP_Sym(p,p->CPU->PC))};
}
}
// tools/platform/ps2/iopload.go:461:1
std::string ps2_IOP_reason(ps2_IOP* p){
{
if ((p->HaltReason != std::string("",0))) {
return p->HaltReason;
}
return p->CPU->HaltReason;
}
}
// tools/platform/ps2/iopsio2.go:104:1
bool ps2_sio2Contains(uint32_t a){
{
return ((a >= cast<uint32_t>(528515584ULL)) && (a < cast<uint32_t>(528515716ULL)));
}
}
// tools/platform/ps2/iopsio2.go:112:1
uint32_t ps2_IOP_sio2Read(ps2_IOP* p,uint32_t a){
{
if ((a == cast<uint32_t>(528515684ULL))) {
if ((len(p->sio2.out) > cast<int64_t>(0ULL))) {
uint32_t b = cast<uint32_t>(p->sio2.out[cast<int64_t>(0ULL)]);
p->sio2.out = sub(p->sio2.out,cast<int64_t>(1ULL),len(p->sio2.out));
return b;
}
return cast<uint32_t>(255ULL);
}
return get(p->io,a);
}
}
// tools/platform/ps2/iopsio2.go:133:1
void ps2_IOP_sio2Write(ps2_IOP* p,uint32_t a,uint32_t v){
{
if ((a == cast<uint32_t>(528515680ULL))) {
p->sio2.in = append(p->sio2.in,Slice<uint8_t>{cast<uint8_t>(v)});
return ;
}
if ((a == cast<uint32_t>(528515688ULL))) {
p->io[a] = (v & ~(cast<uint32_t>(1ULL)));
if ((cast<uint32_t>((v & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL))) {
ps2_IOP_sio2Start(p);
}
return ;
}
p->io[a] = v;
}
}
// tools/platform/ps2/iopsio2.go:166:1
void ps2_IOP_sio2Start(ps2_IOP* p){
{
Slice<uint8_t> in = p->sio2.in;
p->sio2.in = {};
p->sio2.out = sub(p->sio2.out,0,cast<int64_t>(0ULL));
uint32_t recv1 = cast<uint32_t>(119040ULL);
bool dump = ((p->sio2.dumps < cast<int64_t>(4ULL)) && (len(in) > cast<int64_t>(0ULL)));
if (dump) {
p->sio2.dumps++;
ps2_Machine_note(p->ps2,std::string("SIO2: transfer, %d command bytes",32),len(in));
}
int64_t off = cast<int64_t>(0ULL);
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(16ULL));i++){
uint32_t s3 = get(p->io,cast<uint32_t>((cast<uint32_t>(528515584ULL) + cast<uint32_t>((cast<uint32_t>(i) * cast<uint32_t>(4ULL))))));
if ((s3 == cast<uint32_t>(0ULL))) {
break;
}
uint32_t port = cast<uint32_t>((s3 & cast<uint32_t>(3ULL)));
int64_t n = cast<int64_t>((cast<int64_t>(shr<uint32_t>(s3,cast<int64_t>(8ULL))) & cast<int64_t>(511ULL)));
if (dump) {
int64_t end = cast<int64_t>((off + n));
if ((end > len(in))) {
end = len(in);
}
ps2_Machine_note(p->ps2,std::string("SIO2:   send3[%d] = %08X (port %d, len %d): % x",47),i,s3,port,n,sub(in,off,end));
}
if ((n == cast<int64_t>(0ULL))) {
continue;
}
Slice<uint8_t> cmd = Slice<uint8_t>{};
if ((off < len(in))) {
int64_t end = cast<int64_t>((off + n));
if ((end > len(in))) {
end = len(in);
}
cmd = sub(in,off,end);
}
off += n;
if ((port == cast<uint32_t>(0ULL))) {
p->sio2.out = append(p->sio2.out,ps2_IOP_sio2Pad(p,cmd,n));
recv1 = cast<uint32_t>(4352ULL);
}
else {
{int64_t j = cast<int64_t>(0ULL);for (;(j < n);j++){
p->sio2.out = append(p->sio2.out,Slice<uint8_t>{cast<uint8_t>(255ULL)});
}
}}
}
}p->io[cast<uint32_t>(528515692ULL)] = recv1;
if (p->sio2.outArmed) {
p->sio2.outArmed = false;
uint32_t a = ps2_iopPhys(p->sio2.dmaAddr[cast<int64_t>(1ULL)]);
int64_t n = cast<int64_t>(cast<uint32_t>((p->sio2.dmaWords[cast<int64_t>(1ULL)] * cast<uint32_t>(4ULL))));
if ((n > len(p->sio2.out))) {
n = len(p->sio2.out);
}
{int64_t j = cast<int64_t>(0ULL);for (;((j < n) && (cast<int64_t>((cast<int64_t>(a) + j)) < len(p->ram)));j++){
p->ram[cast<int64_t>((cast<int64_t>(a) + j))] = p->sio2.out[j];
}
}if (dump) {
int64_t d = n;
if ((d > cast<int64_t>(24ULL))) {
d = cast<int64_t>(24ULL);
}
ps2_Machine_note(p->ps2,std::string("SIO2:   response, %d bytes to 0x%X: % x",39),n,p->sio2.dmaAddr[cast<int64_t>(1ULL)],sub(p->sio2.out,0,d));
}
p->sio2.out = sub(p->sio2.out,n,len(p->sio2.out));
}
ps2_IOP_raiseIRQ(p,cast<uint32_t>(17ULL));
}
}
// tools/platform/ps2/iopsio2.go:266:1
Slice<uint8_t> ps2_IOP_sio2Pad(ps2_IOP* p,Slice<uint8_t> cmd,int64_t n){
{
Slice<uint8_t> resp = Slice<uint8_t>::make(n);
{auto&& tmp270 = resp;
for(int64_t tmp271=0;tmp271<len(tmp270);++tmp271){
auto i=tmp271;resp[i] = cast<uint8_t>(255ULL);
}}
if (((n < cast<int64_t>(2ULL)) || (len(cmd) < cast<int64_t>(2ULL)))) {
return resp;
}
uint8_t id = cast<uint8_t>(65ULL);
if (p->pad.analog) {
id = cast<uint8_t>(115ULL);
}
if (p->pad.config) {
id = cast<uint8_t>(243ULL);
}
resp[cast<int64_t>(1ULL)] = id;
if ((n > cast<int64_t>(2ULL))) {
resp[cast<int64_t>(2ULL)] = cast<uint8_t>(90ULL);
}
auto put = [&](int64_t i,uint8_t b)->void{
if ((cast<int64_t>((cast<int64_t>(3ULL) + i)) < n)) {
resp[cast<int64_t>((cast<int64_t>(3ULL) + i))] = b;
}
}
;
auto zero = [&](int64_t k)->void{
{int64_t i = cast<int64_t>(0ULL);for (;(i < k);i++){
put(i,cast<uint8_t>(0ULL));
}
}}
;
auto arg = [&](int64_t i)->uint8_t{
if ((cast<int64_t>((cast<int64_t>(3ULL) + i)) < len(cmd))) {
return cmd[cast<int64_t>((cast<int64_t>(3ULL) + i))];
}
return cast<uint8_t>(0ULL);
}
;
auto pollData = [&]()->void{
uint16_t buttons = ps2_Machine_padButtons(p->ps2);
put(cast<int64_t>(0ULL),cast<uint8_t>(~cast<uint8_t>(buttons)));
put(cast<int64_t>(1ULL),cast<uint8_t>(~cast<uint8_t>(shr<uint16_t>(buttons,cast<int64_t>(8ULL)))));
if (p->pad.analog) {
auto tmp272 = ps2_Machine_padSticks(p->ps2);
uint8_t lx = std::get<0>(tmp272);
uint8_t ly = std::get<1>(tmp272);
uint8_t rx = std::get<2>(tmp272);
uint8_t ry = std::get<3>(tmp272);
put(cast<int64_t>(2ULL),rx);
put(cast<int64_t>(3ULL),ry);
put(cast<int64_t>(4ULL),lx);
put(cast<int64_t>(5ULL),ly);
}
}
;
{
switch(cmd[cast<int64_t>(1ULL)]){
case cast<uint8_t>(66ULL):{
pollData();
break;}
case cast<uint8_t>(67ULL):{
if (p->pad.config) {
zero(cast<int64_t>(6ULL));
}
else {
pollData();
}
p->pad.config = (arg(cast<int64_t>(0ULL)) == cast<uint8_t>(1ULL));
break;}
case cast<uint8_t>(68ULL):{
if (p->pad.config) {
zero(cast<int64_t>(6ULL));
p->pad.analog = (arg(cast<int64_t>(0ULL)) == cast<uint8_t>(1ULL));
p->pad.locked = (arg(cast<int64_t>(1ULL)) == cast<uint8_t>(3ULL));
}
break;}
case cast<uint8_t>(69ULL):{
if (p->pad.config) {
uint8_t mode = cast<uint8_t>(0ULL);
if (p->pad.analog) {
mode = cast<uint8_t>(1ULL);
}
put(cast<int64_t>(0ULL),cast<uint8_t>(3ULL));
put(cast<int64_t>(1ULL),cast<uint8_t>(2ULL));
put(cast<int64_t>(2ULL),mode);
put(cast<int64_t>(3ULL),cast<uint8_t>(2ULL));
put(cast<int64_t>(4ULL),cast<uint8_t>(1ULL));
put(cast<int64_t>(5ULL),cast<uint8_t>(0ULL));
}
break;}
case cast<uint8_t>(70ULL):{
if (p->pad.config) {
if ((arg(cast<int64_t>(0ULL)) == cast<uint8_t>(0ULL))) {
put(cast<int64_t>(0ULL),cast<uint8_t>(0ULL));
put(cast<int64_t>(1ULL),cast<uint8_t>(0ULL));
put(cast<int64_t>(2ULL),cast<uint8_t>(1ULL));
put(cast<int64_t>(3ULL),cast<uint8_t>(2ULL));
put(cast<int64_t>(4ULL),cast<uint8_t>(0ULL));
put(cast<int64_t>(5ULL),cast<uint8_t>(10ULL));
}
else {
put(cast<int64_t>(0ULL),cast<uint8_t>(0ULL));
put(cast<int64_t>(1ULL),cast<uint8_t>(0ULL));
put(cast<int64_t>(2ULL),cast<uint8_t>(1ULL));
put(cast<int64_t>(3ULL),cast<uint8_t>(1ULL));
put(cast<int64_t>(4ULL),cast<uint8_t>(1ULL));
put(cast<int64_t>(5ULL),cast<uint8_t>(20ULL));
}
}
break;}
case cast<uint8_t>(71ULL):{
if (p->pad.config) {
zero(cast<int64_t>(6ULL));
put(cast<int64_t>(2ULL),cast<uint8_t>(2ULL));
put(cast<int64_t>(4ULL),cast<uint8_t>(1ULL));
}
break;}
case cast<uint8_t>(76ULL):{
if (p->pad.config) {
zero(cast<int64_t>(6ULL));
if ((arg(cast<int64_t>(0ULL)) == cast<uint8_t>(0ULL))) {
put(cast<int64_t>(3ULL),cast<uint8_t>(4ULL));
}
else {
put(cast<int64_t>(3ULL),cast<uint8_t>(7ULL));
}
}
break;}
case cast<uint8_t>(77ULL):{
if (p->pad.config) {
{auto&& tmp273 = p->pad.actMap;
for(int64_t tmp274=0;tmp274<len(tmp273);++tmp274){
auto i=tmp274;auto b=tmp273[tmp274];put(i,b);
}}
{auto&& tmp275 = p->pad.actMap;
for(int64_t tmp276=0;tmp276<len(tmp275);++tmp276){
auto i=tmp276;p->pad.actMap[i] = arg(i);
}}
}
break;}
case cast<uint8_t>(79ULL):{
if (p->pad.config) {
zero(cast<int64_t>(5ULL));
put(cast<int64_t>(5ULL),cast<uint8_t>(90ULL));
}
break;}
}}
return resp;
}
}
// tools/platform/ps2/iopsio2.go:429:1
void ps2_IOP_dmacmanSetSlice(ps2_IOP* p){
{
auto tmp277 = std::make_tuple(ps2_IOP_arg(p,cast<int64_t>(0ULL)),ps2_IOP_arg(p,cast<int64_t>(1ULL)),ps2_IOP_arg(p,cast<int64_t>(2ULL)),ps2_IOP_arg(p,cast<int64_t>(3ULL)));
uint32_t ch = std::get<0>(tmp277);
uint32_t addr = std::get<1>(tmp277);
uint32_t size = std::get<2>(tmp277);
uint32_t count = std::get<3>(tmp277);
if (((ch == cast<uint32_t>(11ULL)) || (ch == cast<uint32_t>(12ULL)))) {
uint32_t i = cast<uint32_t>((ch - cast<uint32_t>(11ULL)));
p->sio2.dmaAddr[i] = addr;
p->sio2.dmaWords[i] = cast<uint32_t>((size * count));
}
ps2_IOP_setRet(p,cast<uint32_t>(1ULL));
}
}
// tools/platform/ps2/iopsio2.go:439:1
void ps2_IOP_dmacmanStart(ps2_IOP* p){
{
uint32_t ch = ps2_IOP_arg(p,cast<int64_t>(0ULL));
{
switch(ch){
case cast<uint32_t>(11ULL):{
uint32_t a = ps2_iopPhys(p->sio2.dmaAddr[cast<int64_t>(0ULL)]);
int64_t n = cast<int64_t>(cast<uint32_t>((p->sio2.dmaWords[cast<int64_t>(0ULL)] * cast<uint32_t>(4ULL))));
{int64_t j = cast<int64_t>(0ULL);for (;((j < n) && (cast<int64_t>((cast<int64_t>(a) + j)) < len(p->ram)));j++){
p->sio2.in = append(p->sio2.in,Slice<uint8_t>{p->ram[cast<int64_t>((cast<int64_t>(a) + j))]});
}
}break;}
case cast<uint32_t>(12ULL):{
p->sio2.outArmed = true;
break;}
}}
ps2_IOP_setRet(p,cast<uint32_t>(1ULL));
}
}
// tools/platform/ps2/iopsio2.go:458:1
void ps2_IOP_dmacmanChanSetup(ps2_IOP* p){
{
ps2_IOP_setRet(p,cast<uint32_t>(0ULL));
}
}
// tools/platform/ps2/iopsio2.go:460:1
void ps2_IOP_dmacmanChanEnable(ps2_IOP* p){
{
ps2_IOP_setRet(p,cast<uint32_t>(0ULL));
}
}
// tools/platform/ps2/iopspu.go:154:1
ps2_spu2* ps2_newSPU2(){
{
return arenaNew(ps2_spu2{Slice<uint8_t>::make(cast<int64_t>(2048ULL)),Slice<uint8_t>::make(cast<int64_t>(2097152ULL)),{}});
}
}
// tools/platform/ps2/iopspu.go:164:1
uint32_t ps2_spu2_read(ps2_spu2* s,uint32_t off){
{
if ((cast<uint32_t>((off + cast<uint32_t>(4ULL))) > cast<uint32_t>(len(s->regs)))) {
return cast<uint32_t>(0ULL);
}
return cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(s->regs[off]) | shl<uint32_t>(cast<uint32_t>(s->regs[cast<uint32_t>((off + cast<uint32_t>(1ULL)))]),cast<int64_t>(8ULL)))) | shl<uint32_t>(cast<uint32_t>(s->regs[cast<uint32_t>((off + cast<uint32_t>(2ULL)))]),cast<int64_t>(16ULL)))) | shl<uint32_t>(cast<uint32_t>(s->regs[cast<uint32_t>((off + cast<uint32_t>(3ULL)))]),cast<int64_t>(24ULL))));
}
}
// tools/platform/ps2/iopspu.go:172:1
void ps2_spu2_write(ps2_spu2* s,uint32_t off,uint32_t v){
{
if ((cast<uint32_t>((off + cast<uint32_t>(4ULL))) > cast<uint32_t>(len(s->regs)))) {
return ;
}
s->regs[off] = cast<uint8_t>(v);
s->regs[cast<uint32_t>((off + cast<uint32_t>(1ULL)))] = cast<uint8_t>(shr<uint32_t>(v,cast<int64_t>(8ULL)));
s->regs[cast<uint32_t>((off + cast<uint32_t>(2ULL)))] = cast<uint8_t>(shr<uint32_t>(v,cast<int64_t>(16ULL)));
s->regs[cast<uint32_t>((off + cast<uint32_t>(3ULL)))] = cast<uint8_t>(shr<uint32_t>(v,cast<int64_t>(24ULL)));
}
}
// tools/platform/ps2/iopspu.go:183:1
uint32_t ps2_spu2_half(ps2_spu2* s,uint32_t off){
{
return cast<uint32_t>((cast<uint32_t>(s->regs[off]) | shl<uint32_t>(cast<uint32_t>(s->regs[cast<uint32_t>((off + cast<uint32_t>(1ULL)))]),cast<int64_t>(8ULL))));
}
}
// tools/platform/ps2/iopspu.go:187:1
void ps2_spu2_setHalf(ps2_spu2* s,uint32_t off,uint32_t v){
{
s->regs[off] = cast<uint8_t>(v);
s->regs[cast<uint32_t>((off + cast<uint32_t>(1ULL)))] = cast<uint8_t>(shr<uint32_t>(v,cast<int64_t>(8ULL)));
}
}
// tools/platform/ps2/iopspu.go:200:1
uint32_t ps2_spu2_tsa(ps2_spu2* s,int64_t core){
{
uint32_t base = cast<uint32_t>((cast<uint32_t>(core) * cast<uint32_t>(1024ULL)));
return cast<uint32_t>(((cast<uint32_t>((shl<uint32_t>(ps2_spu2_half(s,cast<uint32_t>((base + cast<uint32_t>(424ULL)))),cast<int64_t>(16ULL)) | ps2_spu2_half(s,cast<uint32_t>((base + cast<uint32_t>(426ULL))))))) & cast<uint32_t>(1048575ULL)));
}
}
// tools/platform/ps2/iopspu.go:212:1
void ps2_spu2_dma(ps2_spu2* s,int64_t core,ps2_IOP* p,uint32_t madr,uint32_t n,bool toRAM){
{
uint32_t base = cast<uint32_t>((cast<uint32_t>(core) * cast<uint32_t>(1024ULL)));
ps2_spu2_setHalf(s,cast<uint32_t>(1986ULL),(ps2_spu2_half(s,cast<uint32_t>(1986ULL)) & ~((shl<uint32_t>(cast<uint32_t>(1ULL),(cast<uint64_t>((cast<uint64_t>(2ULL) + cast<uint64_t>(core)))))))));
ps2_spu2_setHalf(s,cast<uint32_t>((base + cast<uint32_t>(836ULL))),(ps2_spu2_half(s,cast<uint32_t>((base + cast<uint32_t>(836ULL)))) & ~(cast<uint32_t>(128ULL))));
uint32_t off = cast<uint32_t>((ps2_spu2_tsa(s,core) * cast<uint32_t>(2ULL)));
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < n);i++){
if ((cast<uint32_t>((off + i)) >= cast<uint32_t>(len(s->ram)))) {
break;
}
if (toRAM) {
ps2_IOP_Write(p,cast<uint32_t>((madr + i)),s->ram[cast<uint32_t>((off + i))]);
}
else {
s->ram[cast<uint32_t>((off + i))] = ps2_IOP_Read(p,cast<uint32_t>((madr + i)));
}
}
}ps2_Machine_note(p->ps2,std::string("IOP SPU2: core %d took %d bytes from 0x%08X into sound memory at 0x%05X",71),core,n,madr,off);
}
}
// tools/platform/ps2/iopspu.go:245:1
void ps2_spu2_complete(ps2_spu2* s,int64_t core,ps2_IOP* p){
{
uint32_t base = cast<uint32_t>((cast<uint32_t>(core) * cast<uint32_t>(1024ULL)));
ps2_spu2_setHalf(s,cast<uint32_t>(1986ULL),cast<uint32_t>((ps2_spu2_half(s,cast<uint32_t>(1986ULL)) | shl<uint32_t>(cast<uint32_t>(1ULL),(cast<uint64_t>((cast<uint64_t>(2ULL) + cast<uint64_t>(core))))))));
ps2_spu2_setHalf(s,cast<uint32_t>((base + cast<uint32_t>(836ULL))),cast<uint32_t>((ps2_spu2_half(s,cast<uint32_t>((base + cast<uint32_t>(836ULL)))) | cast<uint32_t>(128ULL))));
}
}
// tools/platform/ps2/iopspu.go:266:1
uint32_t ps2_spu2_readReg(ps2_spu2* s,uint32_t off,uint64_t now){
{
ps2_spu2_tick(s,now);
return ps2_spu2_read(s,off);
}
}
// tools/platform/ps2/iopspu.go:274:1
void ps2_spu2_writeReg(ps2_spu2* s,uint32_t off,uint32_t v,uint64_t now){
{
ps2_spu2_tick(s,now);
ps2_spu2_write(s,off,v);
int64_t core = cast<int64_t>(divi<uint32_t>(off,cast<uint32_t>(1024ULL)));
if ((core >= cast<int64_t>(2ULL))) {
return ;
}
{
switch(modi<uint32_t>(off,cast<uint32_t>(1024ULL))){
case cast<uint32_t>(416ULL):{
ps2_spu2_keyOnOff(s,core,cast<uint32_t>(416ULL),now,true);
break;}
case cast<uint32_t>(420ULL):{
ps2_spu2_keyOnOff(s,core,cast<uint32_t>(420ULL),now,false);
break;}
}}
}
}
// tools/platform/ps2/iopspu.go:295:1
void ps2_spu2_keyOnOff(ps2_spu2* s,int64_t core,uint32_t r,uint64_t now,bool on){
{
uint32_t base = cast<uint32_t>((cast<uint32_t>(core) * cast<uint32_t>(1024ULL)));
uint32_t mask = cast<uint32_t>((ps2_spu2_half(s,cast<uint32_t>((base + r))) | shl<uint32_t>(ps2_spu2_half(s,cast<uint32_t>((cast<uint32_t>((base + r)) + cast<uint32_t>(2ULL)))),cast<int64_t>(16ULL))));
{int64_t v = cast<int64_t>(0ULL);for (;(v < cast<int64_t>(24ULL));v++){
if ((cast<uint32_t>((mask & (shl<uint32_t>(cast<uint32_t>(1ULL),cast<uint64_t>(v))))) == cast<uint32_t>(0ULL))) {
continue;
}
ps2_spuVoice* vs = (&s->voice[core][v]);
if (on) {
uint32_t ab = cast<uint32_t>((base + cast<uint32_t>((cast<uint32_t>(v) * cast<uint32_t>(12ULL)))));
ps2_spu2_setWordAddr(s,cast<uint32_t>((ab + cast<uint32_t>(456ULL))),ps2_spu2_wordAddr(s,cast<uint32_t>((ab + cast<uint32_t>(448ULL)))));
vs->playing = true;
vs->acc = cast<uint64_t>(0ULL);
vs->lastStep = now;
}
else {
vs->playing = false;
}
}
}ps2_spu2_setHalf(s,cast<uint32_t>((base + r)),cast<uint32_t>(0ULL));
ps2_spu2_setHalf(s,cast<uint32_t>((cast<uint32_t>((base + r)) + cast<uint32_t>(2ULL))),cast<uint32_t>(0ULL));
}
}
// tools/platform/ps2/iopspu.go:318:1
void ps2_spu2_tick(ps2_spu2* s,uint64_t now){
{
{int64_t core = cast<int64_t>(0ULL);for (;(core < cast<int64_t>(2ULL));core++){
{int64_t v = cast<int64_t>(0ULL);for (;(v < cast<int64_t>(24ULL));v++){
ps2_spu2_advance(s,core,v,now);
}
}}
}}
}
// tools/platform/ps2/iopspu.go:328:1
void ps2_spu2_advance(ps2_spu2* s,int64_t core,int64_t v,uint64_t now){
{
ps2_spuVoice* vs = (&s->voice[core][v]);
if ((!vs->playing)) {
vs->lastStep = now;
return ;
}
if ((now <= vs->lastStep)) {
return ;
}
uint64_t pitch = cast<uint64_t>(cast<uint32_t>((ps2_spu2_half(s,cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(core) * cast<uint32_t>(1024ULL))) + cast<uint32_t>((cast<uint32_t>(v) * cast<uint32_t>(16ULL))))) + cast<uint32_t>(4ULL)))) & cast<uint32_t>(16383ULL))));
vs->acc += cast<uint64_t>(((cast<uint64_t>((now - vs->lastStep))) * pitch));
vs->lastStep = now;
if ((pitch == cast<uint64_t>(0ULL))) {
return ;
}
uint32_t ab = cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(core) * cast<uint32_t>(1024ULL))) + cast<uint32_t>((cast<uint32_t>(v) * cast<uint32_t>(12ULL)))));
auto tmp278 = std::make_tuple(cast<uint32_t>((ab + cast<uint32_t>(456ULL))),cast<uint32_t>((ab + cast<uint32_t>(452ULL))));
uint32_t naxOff = std::get<0>(tmp278);
uint32_t lsaxOff = std::get<1>(tmp278);
{int64_t n = cast<int64_t>(0ULL);for (;((vs->acc >= cast<uint64_t>(17920000ULL)) && (n < cast<int64_t>(16384ULL)));n++){
vs->acc -= cast<uint64_t>(17920000ULL);
uint32_t nax = ps2_spu2_wordAddr(s,naxOff);
uint8_t flag = s->ram[cast<uint32_t>(((cast<uint32_t>((cast<uint32_t>((nax * cast<uint32_t>(2ULL))) + cast<uint32_t>(1ULL)))) & cast<uint32_t>(cast<int64_t>((len(s->ram) - cast<int64_t>(1ULL))))))];
if ((cast<uint8_t>((flag & cast<uint8_t>(4ULL))) != cast<uint8_t>(0ULL))) {
ps2_spu2_setWordAddr(s,lsaxOff,nax);
}
if ((cast<uint8_t>((flag & cast<uint8_t>(1ULL))) != cast<uint8_t>(0ULL))) {
ps2_spu2_setWordAddr(s,naxOff,ps2_spu2_wordAddr(s,lsaxOff));
if ((cast<uint8_t>((flag & cast<uint8_t>(2ULL))) == cast<uint8_t>(0ULL))) {
vs->playing = false;
break;
}
}
else {
ps2_spu2_setWordAddr(s,naxOff,cast<uint32_t>((nax + cast<uint32_t>(8ULL))));
}
}
}if ((vs->acc >= cast<uint64_t>(17920000ULL))) {
vs->acc %= cast<uint64_t>(17920000ULL);
}
}
}
// tools/platform/ps2/iopspu.go:376:1
uint32_t ps2_spu2_wordAddr(ps2_spu2* s,uint32_t off){
{
return cast<uint32_t>(((cast<uint32_t>((shl<uint32_t>(ps2_spu2_half(s,off),cast<int64_t>(16ULL)) | ps2_spu2_half(s,cast<uint32_t>((off + cast<uint32_t>(2ULL))))))) & cast<uint32_t>(1048575ULL)));
}
}
// tools/platform/ps2/iopspu.go:380:1
void ps2_spu2_setWordAddr(ps2_spu2* s,uint32_t off,uint32_t w){
{
ps2_spu2_setHalf(s,off,cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(65535ULL))));
ps2_spu2_setHalf(s,cast<uint32_t>((off + cast<uint32_t>(2ULL))),cast<uint32_t>((w & cast<uint32_t>(65535ULL))));
}
}
// tools/platform/ps2/iopthreads.go:87:1
uint32_t ps2_IOP_currentTCB(ps2_IOP* p){
{
if ((p->schedIsRun == cast<uint32_t>(0ULL))) {
return cast<uint32_t>(0ULL);
}
return ps2_IOP_Read32(p,p->schedIsRun);
}
}
// tools/platform/ps2/iopthreads.go:99:1
bool ps2_IOP_looksLikeTCB(ps2_IOP* p,uint32_t a){
{
uint32_t status = ps2_IOP_Read32(p,cast<uint32_t>((a + cast<uint32_t>(12ULL))));
uint32_t state = cast<uint32_t>((status & cast<uint32_t>(255ULL)));
uint32_t prio = shr<uint32_t>(status,cast<int64_t>(16ULL));
{
auto tmp279 = lookup(ps2_iopThreadStates,state);
bool ok = std::get<1>(tmp279);
if ((!ok)) {
return false;
}
}
if (((cast<uint32_t>((status & cast<uint32_t>(65280ULL))) != cast<uint32_t>(0ULL)) || (prio > cast<uint32_t>(127ULL)))) {
return false;
}
if ((!ps2_IOP_inModule(p,ps2_IOP_Read32(p,cast<uint32_t>((a + cast<uint32_t>(56ULL))))))) {
return false;
}
uint32_t stack = ps2_IOP_Read32(p,cast<uint32_t>((a + cast<uint32_t>(60ULL))));
uint32_t size = ps2_IOP_Read32(p,cast<uint32_t>((a + cast<uint32_t>(64ULL))));
if (((((stack < cast<uint32_t>(4096ULL)) || (stack >= cast<uint32_t>(2097152ULL))) || (size == cast<uint32_t>(0ULL))) || (size > cast<uint32_t>(262144ULL)))) {
return false;
}
uint32_t frame = ps2_IOP_Read32(p,cast<uint32_t>((a + cast<uint32_t>(16ULL))));
if (((frame != cast<uint32_t>(0ULL)) && (((frame < stack) || (frame >= cast<uint32_t>((stack + size))))))) {
return false;
}
return true;
}
}
// tools/platform/ps2/iopthreads.go:126:1
bool ps2_IOP_inModule(ps2_IOP* p,uint32_t addr){
{
{auto&& tmp280 = p->modules;
for(int64_t tmp281=0;tmp281<len(tmp280);++tmp281){
auto m=tmp280[tmp281];if (((addr >= m->Base) && (addr < cast<uint32_t>((m->Base + m->Size))))) {
return true;
}
}}
return false;
}
}
// tools/platform/ps2/iopthreads.go:144:1
std::string ps2_IOP_IOPThreads(ps2_IOP* p){
{
uint32_t cur = ps2_IOP_currentTCB(p);
Slice<uint32_t> tcbs={};
Map<uint32_t,bool> seen = Map<uint32_t,bool>{};
{uint32_t a = cast<uint32_t>(4096ULL);for (;(a < cast<uint32_t>(2097152ULL));a += cast<uint32_t>(4ULL)){
if (ps2_IOP_looksLikeTCB(p,a)) {
tcbs = append(tcbs,Slice<uint32_t>{a});
seen[a] = true;
}
}
}go_sort_Slice(tcbs,[&](int64_t i,int64_t j)->bool{
return (tcbs[i] < tcbs[j]);
}
);
auto tmp282 = std::make_tuple(cast<int64_t>(0ULL),true);
int64_t chainLen = std::get<0>(tmp282);
bool chainOK = std::get<1>(tmp282);
if ((len(tcbs) > cast<int64_t>(0ULL))) {
Map<uint32_t,bool> visited = Map<uint32_t,bool>{};
{uint32_t a = tcbs[cast<int64_t>((len(tcbs) - cast<int64_t>(1ULL)))];for (;(((a != cast<uint32_t>(0ULL)) && get(seen,a)) && (!get(visited,a)));a = ps2_IOP_Read32(p,cast<uint32_t>((a + cast<uint32_t>(36ULL))))){
visited[a] = true;
chainLen++;
}
}chainOK = (chainLen == len(tcbs));
}
strings_Builder b={};
go_fmt_Fprintf((&b),std::string("the IOP's threads (%d; running = %s):\012",38),len(tcbs),ps2_IOP_Sym(p,p->CPU->PC));
{auto&& tmp283 = tcbs;
for(int64_t tmp284=0;tmp284<len(tmp283);++tmp284){
auto a=tmp283[tmp284];ps2_iopThread t = ps2_IOP_readThread(p,a,cur);
std::string mark = std::string(" ",1);
if (t.running) {
mark = std::string("*",1);
}
go_fmt_Fprintf((&b),std::string("  %s %08X  p%-3d %-12s  %-28s  @ %s\012",36),mark,a,t.prio,get(ps2_iopThreadStates,t.state),ps2_IOP_Sym(p,t.entry),ps2_IOP_Sym(p,t.savedPC));
if (((t.state == cast<uint32_t>(4ULL)) || (t.state == cast<uint32_t>(12ULL)))) {
if ((t.waitTyp == cast<uint32_t>(4ULL))) {
go_fmt_Fprintf((&b),std::string("        waiting on event flag %08X, bits %X\012",44),t.waitObj,t.waitMsk);
}
else {
go_fmt_Fprintf((&b),std::string("        waiting (type %d) on %08X\012",34),t.waitTyp,t.waitObj);
}
}
}}
if ((!chainOK)) {
go_fmt_Fprintf((&b),std::string("  (note: the +0x24 chain from the newest block visits %d of %d \342\200\224 the scan and the chain disagree)\012",100),chainLen,len(tcbs));
}
return strings_Builder_String(&(b));
}
}
// tools/platform/ps2/iopthreads.go:199:1
ps2_iopThread ps2_IOP_readThread(ps2_IOP* p,uint32_t a,uint32_t cur){
{
uint32_t status = ps2_IOP_Read32(p,cast<uint32_t>((a + cast<uint32_t>(12ULL))));
ps2_iopThread t = ps2_iopThread{a,shr<uint32_t>(status,cast<int64_t>(16ULL)),cast<uint32_t>((status & cast<uint32_t>(255ULL))),ps2_IOP_Read32(p,cast<uint32_t>((a + cast<uint32_t>(56ULL)))),ps2_IOP_Read32(p,cast<uint32_t>((a + cast<uint32_t>(60ULL)))),ps2_IOP_Read32(p,cast<uint32_t>((a + cast<uint32_t>(16ULL)))),{},ps2_IOP_Read32(p,cast<uint32_t>((a + cast<uint32_t>(28ULL)))),ps2_IOP_Read32(p,cast<uint32_t>((a + cast<uint32_t>(32ULL)))),ps2_IOP_Read32(p,cast<uint32_t>((a + cast<uint32_t>(40ULL)))),(a == cur)};
if (t.running) {
t.savedPC = p->CPU->PC;
}
else if ((t.frame != cast<uint32_t>(0ULL))) {
t.savedPC = ps2_IOP_Read32(p,cast<uint32_t>((t.frame + cast<uint32_t>(140ULL))));
}
return t;
}
}
// tools/platform/ps2/ioptimer.go:72:1
uint32_t ps2_iopTimerIRQ(int64_t n){
{
if ((n < cast<int64_t>(3ULL))) {
return cast<uint32_t>(cast<int64_t>((cast<int64_t>(4ULL) + n)));
}
return cast<uint32_t>(cast<int64_t>((cast<int64_t>(11ULL) + n)));
}
}
// tools/platform/ps2/ioptimer.go:82:1
uint32_t ps2_iopTimerMax(int64_t n){
{
if ((n < cast<int64_t>(3ULL))) {
return cast<uint32_t>(65535ULL);
}
return cast<uint32_t>(4294967295ULL);
}
}
// tools/platform/ps2/ioptimer.go:90:1
std::tuple<int64_t,uint32_t,bool> ps2_iopTimerReg(uint32_t a){
int64_t n{};
uint32_t reg{};
bool ok{};
{
{
if (((a >= cast<uint32_t>(528486656ULL)) && (a < cast<uint32_t>(528486704ULL)))){
return {divi<int64_t>(cast<int64_t>(cast<uint32_t>((a - cast<uint32_t>(528486656ULL)))),cast<int64_t>(16ULL)),cast<uint32_t>((a & cast<uint32_t>(12ULL))),true};
}
else if (((a >= cast<uint32_t>(528487552ULL)) && (a < cast<uint32_t>(528487600ULL)))){
return {cast<int64_t>((cast<int64_t>(3ULL) + divi<int64_t>(cast<int64_t>(cast<uint32_t>((a - cast<uint32_t>(528487552ULL)))),cast<int64_t>(16ULL)))),cast<uint32_t>((a & cast<uint32_t>(12ULL))),true};
}
}
tmp285:;
return {cast<int64_t>(0ULL),cast<uint32_t>(0ULL),false};
}
}
// tools/platform/ps2/ioptimer.go:101:1
std::tuple<uint32_t,bool> ps2_IOP_timerRead(ps2_IOP* p,uint32_t a){
{
auto tmp286 = ps2_iopTimerReg(a);
int64_t n = std::get<0>(tmp286);
uint32_t reg = std::get<1>(tmp286);
bool ok = std::get<2>(tmp286);
if ((!ok)) {
return {cast<uint32_t>(0ULL),false};
}
auto tmp287 = ps2_IOP_timerPeek(p,a);
uint32_t v = std::get<0>(tmp287);
if ((reg == cast<uint32_t>(4ULL))) {
p->timerAck |= shl<uint32_t>(cast<uint32_t>(1ULL),cast<uint64_t>(n));
}
return {v,true};
}
}
// tools/platform/ps2/ioptimer.go:140:1
void ps2_IOP_timerAckFlush(ps2_IOP* p){
{
{int64_t n = cast<int64_t>(0ULL);for (;((p->timerAck != cast<uint32_t>(0ULL)) && (n < cast<int64_t>(6ULL)));n++){
if ((cast<uint32_t>((p->timerAck & (shl<uint32_t>(cast<uint32_t>(1ULL),cast<uint64_t>(n))))) == cast<uint32_t>(0ULL))) {
continue;
}
p->timers[n].mode &= ~(cast<uint32_t>(6144ULL));
}
}p->timerAck = cast<uint32_t>(0ULL);
}
}
// tools/platform/ps2/ioptimer.go:151:1
std::tuple<uint32_t,bool> ps2_IOP_timerPeek(ps2_IOP* p,uint32_t a){
{
auto tmp288 = ps2_iopTimerReg(a);
int64_t n = std::get<0>(tmp288);
uint32_t reg = std::get<1>(tmp288);
bool ok = std::get<2>(tmp288);
if ((!ok)) {
return {cast<uint32_t>(0ULL),false};
}
ps2_iopTimer* t = (&p->timers[n]);
{
switch(reg){
case cast<uint32_t>(0ULL):{
return {t->count,true};
break;}
case cast<uint32_t>(4ULL):{
return {t->mode,true};
break;}
case cast<uint32_t>(8ULL):{
return {t->target,true};
break;}
}}
return {cast<uint32_t>(0ULL),false};
}
}
// tools/platform/ps2/ioptimer.go:169:1
bool ps2_IOP_timerWrite(ps2_IOP* p,uint32_t a,uint32_t v){
{
auto tmp289 = ps2_iopTimerReg(a);
int64_t n = std::get<0>(tmp289);
uint32_t reg = std::get<1>(tmp289);
bool ok = std::get<2>(tmp289);
if ((!ok)) {
return false;
}
ps2_iopTimer* t = (&p->timers[n]);
{
switch(reg){
case cast<uint32_t>(0ULL):{
t->count = v;
break;}
case cast<uint32_t>(4ULL):{
t->mode = v;
t->count = cast<uint32_t>(0ULL);
break;}
case cast<uint32_t>(8ULL):{
t->target = v;
break;}
}}
t->fired = false;
return true;
}
}
// tools/platform/ps2/ioptimer.go:196:1
void ps2_IOP_timerTick(ps2_IOP* p){
{
{auto&& tmp290 = p->timers;
for(int64_t tmp291=0;tmp291<len(tmp290);++tmp291){
auto n=tmp291;ps2_iopTimer* t = (&p->timers[n]);
if ((t->mode == cast<uint32_t>(0ULL))) {
continue;
}
uint32_t max = ps2_iopTimerMax(n);
if ((t->count == max)) {
t->count = cast<uint32_t>(0ULL);
t->mode |= cast<uint32_t>(4096ULL);
if ((cast<uint32_t>((t->mode & cast<uint32_t>(32ULL))) != cast<uint32_t>(0ULL))) {
ps2_IOP_raiseIRQ(p,ps2_iopTimerIRQ(n));
}
continue;
}
t->count++;
if (((t->target == cast<uint32_t>(0ULL)) || (t->count < t->target))) {
continue;
}
if ((cast<uint32_t>((t->mode & cast<uint32_t>(8ULL))) != cast<uint32_t>(0ULL))) {
t->count = cast<uint32_t>(0ULL);
t->fired = false;
}
if (t->fired) {
continue;
}
t->mode |= cast<uint32_t>(2048ULL);
if ((cast<uint32_t>((t->mode & cast<uint32_t>(16ULL))) != cast<uint32_t>(0ULL))) {
ps2_IOP_raiseIRQ(p,ps2_iopTimerIRQ(n));
}
if (((cast<uint32_t>((t->mode & cast<uint32_t>(64ULL))) == cast<uint32_t>(0ULL)) || (cast<uint32_t>((t->mode & cast<uint32_t>(8ULL))) == cast<uint32_t>(0ULL)))) {
t->fired = true;
}
}}
}
}
// tools/platform/ps2/iopvblank.go:53:1
void ps2_IOP_vblankRegister(ps2_IOP* p){
{
auto tmp292 = std::make_tuple(ps2_IOP_arg(p,cast<int64_t>(0ULL)),ps2_IOP_arg(p,cast<int64_t>(1ULL)),ps2_IOP_arg(p,cast<int64_t>(2ULL)),ps2_IOP_arg(p,cast<int64_t>(3ULL)));
uint32_t edge = std::get<0>(tmp292);
uint32_t prio = std::get<1>(tmp292);
uint32_t fn = std::get<2>(tmp292);
uint32_t arg = std::get<3>(tmp292);
{auto&& tmp293 = p->vblankHandlers;
for(int64_t tmp294=0;tmp294<len(tmp293);++tmp294){
auto i=tmp294;auto h=tmp293[tmp294];if (((h.fn == fn) && (h.edge == edge))) {
p->vblankHandlers[i] = ps2_iopVblankHandler{edge,prio,fn,arg};
ps2_IOP_setRet(p,cast<uint32_t>(0ULL));
return ;
}
}}
p->vblankHandlers = append(p->vblankHandlers,Slice<ps2_iopVblankHandler>{ps2_iopVblankHandler{edge,prio,fn,arg}});
{int64_t i = cast<int64_t>((len(p->vblankHandlers) - cast<int64_t>(1ULL)));for (;(i > cast<int64_t>(0ULL));i--){
auto tmp295 = std::make_tuple(p->vblankHandlers[cast<int64_t>((i - cast<int64_t>(1ULL)))],p->vblankHandlers[i]);
ps2_iopVblankHandler a = std::get<0>(tmp295);
ps2_iopVblankHandler b = std::get<1>(tmp295);
if (((((a.edge != cast<uint32_t>(0ULL)) && (b.edge == cast<uint32_t>(0ULL)))) || (((a.edge == b.edge) && (a.prio > b.prio))))) {
auto tmp296 = std::make_tuple(b,a);
p->vblankHandlers[cast<int64_t>((i - cast<int64_t>(1ULL)))] = std::get<0>(tmp296);
p->vblankHandlers[i] = std::get<1>(tmp296);
}
}
}ps2_Machine_note(p->ps2,std::string("IOP: vblank handler %s registered (edge %d, priority 0x%X, arg 0x%X)",68),ps2_IOP_Sym(p,fn),edge,prio,arg);
ps2_IOP_setRet(p,cast<uint32_t>(0ULL));
}
}
// tools/platform/ps2/iopvblank.go:82:1
void ps2_IOP_vblankRelease(ps2_IOP* p){
{
auto tmp297 = std::make_tuple(ps2_IOP_arg(p,cast<int64_t>(0ULL)),ps2_IOP_arg(p,cast<int64_t>(1ULL)));
uint32_t edge = std::get<0>(tmp297);
uint32_t fn = std::get<1>(tmp297);
{auto&& tmp298 = p->vblankHandlers;
for(int64_t tmp299=0;tmp299<len(tmp298);++tmp299){
auto i=tmp299;auto h=tmp298[tmp299];if (((h.fn == fn) && (h.edge == edge))) {
p->vblankHandlers = append(sub(p->vblankHandlers,0,i),sub(p->vblankHandlers,cast<int64_t>((i + cast<int64_t>(1ULL))),len(p->vblankHandlers)));
ps2_Machine_note(p->ps2,std::string("IOP: vblank handler %s released (edge %d)",41),ps2_IOP_Sym(p,fn),edge);
ps2_IOP_setRet(p,cast<uint32_t>(0ULL));
return ;
}
}}
ps2_IOP_setRet(p,cast<uint32_t>(0ULL));
}
}
// tools/platform/ps2/iopvblank.go:98:1
void ps2_IOP_vblankTick(ps2_IOP* p){
{
if ((len(p->vblankHandlers) == cast<int64_t>(0ULL))) {
return ;
}
p->vblankPending = true;
p->raised[cast<int64_t>(0ULL)]++;
}
}
// tools/platform/ps2/iopvblank.go:111:1
void ps2_IOP_vblankDeliver(ps2_IOP* p){
{
uint32_t frame = ((cast<uint32_t>((mips_CPU_Reg(p->CPU,cast<uint32_t>(29ULL)) - cast<uint32_t>(184ULL)))) & ~(cast<uint32_t>(7ULL)));
ps2_IOP_saveFrame(p,frame);
bool preemptible = (p->callDepth == cast<int64_t>(0ULL));
p->inIntr++;
p->intrEnabled = false;
ps2_IOP_ieEvent(p,std::string("deliver",7),frame,cast<uint32_t>(0ULL),mips_CPU_Reg(p->CPU,cast<uint32_t>(31ULL)));
p->delivered[cast<int64_t>(0ULL)]++;
uint32_t failed={};
Error failErr={};
{auto&& tmp300 = p->vblankHandlers;
for(int64_t tmp301=0;tmp301<len(tmp300);++tmp301){
auto h=tmp300[tmp301];mips_CPU_SetReg(p->CPU,cast<uint32_t>(4ULL),h.arg);
{
auto tmp302 = ps2_IOP_callGuestOn(p,h.fn,cast<uint32_t>(2063360ULL));
Error err = std::get<1>(tmp302);
if (bool(err)) {
auto tmp303 = std::make_tuple(h.fn,err);
failed = std::get<0>(tmp303);
failErr = std::get<1>(tmp303);
break;
}
}
}}
uint32_t resume = frame;
if (((!failErr) && preemptible)) {
{
auto tmp304 = ps2_IOP_intrReschedule(p,frame);
uint32_t next = std::get<0>(tmp304);
bool ok = std::get<1>(tmp304);
if (ok) {
resume = next;
}
}
}
p->inIntr--;
ps2_IOP_loadFrame(p,resume);
if (bool(failErr)) {
ps2_IOP_halt(p,std::string("the vblank handler %s did not return: %v",40),ps2_IOP_Sym(p,failed),failErr);
}
}
}
// tools/platform/ps2/irx.go:245:1
void ps2_IRX_scanLibraries(ps2_IRX* m){
{
{uint32_t o = cast<uint32_t>(0ULL);for (;(cast<uint32_t>((o + cast<uint32_t>(20ULL))) <= cast<uint32_t>(len(m->Image)));o += cast<uint32_t>(4ULL)){
uint32_t magic = ps2_IRX_word(m,o);
if (((magic != cast<uint32_t>(1105199104ULL)) && (magic != cast<uint32_t>(1103101952ULL)))) {
continue;
}
if ((ps2_IRX_word(m,cast<uint32_t>((o + cast<uint32_t>(4ULL)))) != cast<uint32_t>(0ULL))) {
continue;
}
std::string name = ps2_libraryName(sub(m->Image,cast<uint32_t>((o + cast<uint32_t>(12ULL))),cast<uint32_t>((o + cast<uint32_t>(20ULL)))));
if ((name == std::string("",0))) {
continue;
}
uint16_t version = cast<uint16_t>(ps2_IRX_word(m,cast<uint32_t>((o + cast<uint32_t>(8ULL)))));
if ((magic == cast<uint32_t>(1105199104ULL))) {
ps2_IRXImport imp = ps2_IRXImport{name,version,o,{},{}};
{uint32_t p = cast<uint32_t>((o + cast<uint32_t>(20ULL)));for (;(cast<uint32_t>((p + cast<uint32_t>(8ULL))) <= cast<uint32_t>(len(m->Image)));p += cast<uint32_t>(8ULL)){
auto tmp305 = std::make_tuple(ps2_IRX_word(m,p),ps2_IRX_word(m,cast<uint32_t>((p + cast<uint32_t>(4ULL)))));
uint32_t lo = std::get<0>(tmp305);
uint32_t hi = std::get<1>(tmp305);
if (((lo == cast<uint32_t>(0ULL)) && (hi == cast<uint32_t>(0ULL)))) {
break;
}
if ((lo != cast<uint32_t>(65011720ULL))) {
break;
}
imp.IDs = append(imp.IDs,Slice<uint16_t>{cast<uint16_t>(hi)});
imp.Stubs = append(imp.Stubs,Slice<uint32_t>{p});
}
}if ((len(imp.IDs) > cast<int64_t>(0ULL))) {
m->Imports = append(m->Imports,Slice<ps2_IRXImport>{imp});
}
continue;
}
ps2_IRXExport exp = ps2_IRXExport{name,version,o,{}};
{uint32_t p = cast<uint32_t>((o + cast<uint32_t>(20ULL)));for (;(cast<uint32_t>((p + cast<uint32_t>(4ULL))) <= cast<uint32_t>(len(m->Image)));p += cast<uint32_t>(4ULL)){
uint32_t w = ps2_IRX_word(m,p);
if (((w == cast<uint32_t>(0ULL)) && (len(exp.Entries) >= cast<int64_t>(4ULL)))) {
break;
}
exp.Entries = append(exp.Entries,Slice<uint32_t>{w});
}
}if ((len(exp.Entries) >= cast<int64_t>(4ULL))) {
m->Exports = append(m->Exports,Slice<ps2_IRXExport>{exp});
}
}
}}
}
// tools/platform/ps2/irx.go:299:1
uint32_t ps2_IRX_word(ps2_IRX* m,uint32_t o){
{
return le_Uint32(rrBorrow(m->Image,o,len(m->Image)));
}
}
// tools/platform/ps2/irx.go:305:1
std::string ps2_libraryName(Slice<uint8_t> b){
{
int64_t n = cast<int64_t>(0ULL);
{;for (;((n < len(b)) && (b[n] != cast<uint8_t>(0ULL)));){
n++;
}
}std::string s = go_strings_TrimRight(cast<std::string>(sub(b,0,n)),std::string(" ",1));
if ((s == std::string("",0))) {
return std::string("",0);
}
{int64_t i = cast<int64_t>(0ULL);for (;(i < len(s));i++){
if (((cast<uint8_t>(s[i]) < cast<uint8_t>(48ULL)) || (cast<uint8_t>(s[i]) > cast<uint8_t>(122ULL)))) {
return std::string("",0);
}
}
}return s;
}
}
// tools/platform/ps2/irx.go:322:1
std::string ps2_cstring(Slice<uint8_t> b){
{
{
int64_t i = go_bytes_IndexByte(b,cast<uint8_t>(0ULL));
if ((i >= cast<int64_t>(0ULL))) {
b = sub(b,0,i);
}
}
return cast<std::string>(b);
}
}
// tools/platform/ps2/irx.go:339:1
std::tuple<Slice<uint8_t>,Error> ps2_IRX_Relocate(ps2_IRX* m,uint32_t base){
{
Slice<uint8_t> img = Slice<uint8_t>::make(m->MemSz);
gcopy(img,m->Image);
auto get = [&](uint32_t o)->uint32_t{
return le_Uint32(rrBorrow(img,o,len(img)));
}
;
auto put = [&](uint32_t o,uint32_t v)->void{
le_PutUint32(rrBorrow(img,o,len(img)),v);
}
;
Slice<uint32_t> pendingHI={};
{auto&& tmp306 = m->Relocs;
for(int64_t tmp307=0;tmp307<len(tmp306);++tmp307){
auto r=tmp306[tmp307];if ((cast<uint32_t>((r.Offset + cast<uint32_t>(4ULL))) > m->MemSz)) {
return {{},go_fmt_Errorf(std::string("ps2: a relocation in %s points outside the module (0x%X)",56),m->Name,r.Offset)};
}
{
switch(r.Type){
case cast<uint8_t>(2ULL):{
put(r.Offset,cast<uint32_t>((get(r.Offset) + base)));
break;}
case cast<uint8_t>(4ULL):{
uint32_t insn = get(r.Offset);
uint32_t target = cast<uint32_t>((shl<uint32_t>((cast<uint32_t>((insn & cast<uint32_t>(67108863ULL)))),cast<int64_t>(2ULL)) + base));
put(r.Offset,cast<uint32_t>(((insn & ~(cast<uint32_t>(67108863ULL))) | cast<uint32_t>(((shr<uint32_t>(target,cast<int64_t>(2ULL))) & cast<uint32_t>(67108863ULL))))));
break;}
case cast<uint8_t>(5ULL):{
pendingHI = append(pendingHI,Slice<uint32_t>{r.Offset});
break;}
case cast<uint8_t>(6ULL):{
uint32_t lo = get(r.Offset);
int32_t addend = cast<int32_t>(cast<int16_t>(cast<uint32_t>((lo & cast<uint32_t>(65535ULL)))));
{auto&& tmp308 = pendingHI;
for(int64_t tmp309=0;tmp309<len(tmp308);++tmp309){
auto hiOff=tmp308[tmp309];uint32_t hi = get(hiOff);
uint32_t full = cast<uint32_t>((cast<uint32_t>(cast<int32_t>((shl<int32_t>(cast<int32_t>(cast<uint32_t>((hi & cast<uint32_t>(65535ULL)))),cast<int64_t>(16ULL)) + addend))) + base));
put(hiOff,cast<uint32_t>(((hi & ~(cast<uint32_t>(65535ULL))) | cast<uint32_t>((shr<uint32_t>((cast<uint32_t>((full + cast<uint32_t>(32768ULL)))),cast<int64_t>(16ULL)) & cast<uint32_t>(65535ULL))))));
}}
pendingHI = sub(pendingHI,0,cast<int64_t>(0ULL));
uint32_t full = cast<uint32_t>((cast<uint32_t>(addend) + base));
put(r.Offset,cast<uint32_t>(((lo & ~(cast<uint32_t>(65535ULL))) | cast<uint32_t>((full & cast<uint32_t>(65535ULL))))));
break;}
default:{
return {{},go_fmt_Errorf(std::string("ps2: %s uses relocation type %d, which nothing on this disc does",64),m->Name,r.Type)};
break;}
}}
}}
if ((len(pendingHI) > cast<int64_t>(0ULL))) {
return {{},go_fmt_Errorf(std::string("ps2: %s ends with %d unpaired HI16 relocations",46),m->Name,len(pendingHI))};
}
return {img,{}};
}
}
// tools/platform/ps2/kernel.go:37:1
uint32_t ps2_Machine_arg(ps2_Machine* m,uint32_t i){
{
return cast<uint32_t>(r5900_CPU_Reg(m->CPU,cast<uint32_t>((cast<uint32_t>(4ULL) + i))));
}
}
// tools/platform/ps2/kernel.go:38:1
uint32_t ps2_Machine_argT0(ps2_Machine* m){
{
return cast<uint32_t>(r5900_CPU_Reg(m->CPU,cast<uint32_t>(8ULL)));
}
}
// tools/platform/ps2/kernel.go:41:1
void ps2_Machine_setRet(ps2_Machine* m,uint32_t v){
{
r5900_CPU_SetReg(m->CPU,cast<uint32_t>(2ULL),cast<uint64_t>(cast<int64_t>(cast<int32_t>(v))));
}
}
// tools/platform/ps2/kernel.go:45:1
bool ps2_Machine_handleSyscall(ps2_Machine* m,r5900_CPU* c){
{
int32_t num = cast<int32_t>(cast<uint32_t>(r5900_CPU_Reg(c,cast<uint32_t>(3ULL))));
if ((num < cast<int32_t>(0ULL))) {
num = cast<int32_t>(-num);
}
auto tmp310 = lookup(ps2_syscalls,cast<uint32_t>(num));
ps2_syscallEntry e = std::get<0>(tmp310);
bool known = std::get<1>(tmp310);
std::string name = e.name;
if ((!known)) {
name = go_fmt_Sprintf(std::string("syscall_%d",10),num);
}
m->SyscallCalls[name]++;
{
uint32_t h = get(m->userSyscalls,cast<uint32_t>(num));
if ((h != cast<uint32_t>(0ULL))) {
r5900_CPU_SetPC(c,cast<uint64_t>(h));
return true;
}
}
if ((!e.fn)) {
ps2_Machine_note(m,std::string("%s (%d) unmodelled: a0=0x%08X a1=0x%08X a2=0x%08X a3=0x%08X, called from %s",75),name,num,ps2_Machine_arg(m,cast<uint32_t>(0ULL)),ps2_Machine_arg(m,cast<uint32_t>(1ULL)),ps2_Machine_arg(m,cast<uint32_t>(2ULL)),ps2_Machine_arg(m,cast<uint32_t>(3ULL)),ps2_Machine_Sym(m,cast<uint32_t>(r5900_CPU_CurPC(c))));
ps2_Machine_setRet(m,cast<uint32_t>(0ULL));
return true;
}
e.fn(m);
return true;
}
}
// tools/platform/ps2/kernel.go:236:1
void ps2_Machine_setupThread(ps2_Machine* m){
{
auto tmp313 = std::make_tuple(ps2_Machine_arg(m,cast<uint32_t>(0ULL)),ps2_Machine_arg(m,cast<uint32_t>(1ULL)),ps2_Machine_arg(m,cast<uint32_t>(2ULL)),ps2_Machine_arg(m,cast<uint32_t>(3ULL)));
uint32_t gp = std::get<0>(tmp313);
uint32_t stack = std::get<1>(tmp313);
uint32_t stackSize = std::get<2>(tmp313);
uint32_t args = std::get<3>(tmp313);
uint32_t root = ps2_Machine_argT0(m);
uint32_t top = cast<uint32_t>((stack + stackSize));
if ((stack == cast<uint32_t>(4294967295ULL))) {
top = cast<uint32_t>(33554432ULL);
}
uint32_t sp = ((cast<uint32_t>((top - cast<uint32_t>(672ULL)))) & ~(cast<uint32_t>(15ULL)));
ps2_thread* t = get(m->threads,m->currentThread);
if (bool(t)) {
t->gp = gp;
t->stack = cast<uint32_t>((top - stackSize));
t->stackSz = stackSize;
t->entry = root;
}
ps2_Machine_note(m,std::string("SetupThread: gp=0x%08X stack=0x%08X+0x%X args=0x%08X root=%s -> sp=0x%08X",73),gp,cast<uint32_t>((top - stackSize)),stackSize,args,ps2_Machine_Sym(m,root),sp);
m->argc = cast<uint32_t>(0ULL);
m->argv = cast<uint32_t>(0ULL);
ps2_Machine_setRet(m,sp);
}
}
// tools/platform/ps2/kernel.go:278:1
void ps2_Machine_setSyscall(ps2_Machine* m){
{
auto tmp314 = std::make_tuple(ps2_Machine_arg(m,cast<uint32_t>(0ULL)),ps2_Machine_arg(m,cast<uint32_t>(1ULL)));
uint32_t num = std::get<0>(tmp314);
uint32_t handler = std::get<1>(tmp314);
if ((handler == cast<uint32_t>(0ULL))) {
removeKey(m->userSyscalls,num);
ps2_Machine_note(m,std::string("SetSyscall %d: back to the kernel's",35),num);
}
else {
m->userSyscalls[num] = handler;
ps2_Machine_note(m,std::string("SetSyscall %d -> %s",19),num,ps2_Machine_Sym(m,handler));
}
ps2_Machine_setRet(m,cast<uint32_t>(0ULL));
}
}
// tools/platform/ps2/kernel.go:293:1
void ps2_Machine_setupHeap(ps2_Machine* m){
{
auto tmp315 = std::make_tuple(ps2_Machine_arg(m,cast<uint32_t>(0ULL)),ps2_Machine_arg(m,cast<uint32_t>(1ULL)));
uint32_t start = std::get<0>(tmp315);
uint32_t size = std::get<1>(tmp315);
m->heapPtr = ((cast<uint32_t>((start + cast<uint32_t>(15ULL)))) & ~(cast<uint32_t>(15ULL)));
if ((size == cast<uint32_t>(4294967295ULL))) {
ps2_thread* t = get(m->threads,m->currentThread);
if ((bool(t) && (t->stack != cast<uint32_t>(0ULL)))) {
m->heapEnd = t->stack;
}
else {
m->heapEnd = cast<uint32_t>(33538048ULL);
}
}
else {
m->heapEnd = cast<uint32_t>((m->heapPtr + size));
}
ps2_Machine_note(m,std::string("SetupHeap: 0x%08X..0x%08X (%d KiB)",34),m->heapPtr,m->heapEnd,divi<uint32_t>((cast<uint32_t>((m->heapEnd - m->heapPtr))),cast<uint32_t>(1024ULL)));
ps2_Machine_setRet(m,m->heapEnd);
}
}
// tools/platform/ps2/kernel.go:336:1
void ps2_Machine_deci2Call(ps2_Machine* m){
{
auto tmp316 = std::make_tuple(ps2_Machine_arg(m,cast<uint32_t>(0ULL)),ps2_Machine_arg(m,cast<uint32_t>(1ULL)));
uint32_t fn = std::get<0>(tmp316);
uint32_t param = std::get<1>(tmp316);
constexpr int64_t deci2Open=1ULL;
constexpr int64_t deci2Close=2ULL;
constexpr int64_t deci2ReqSend=3ULL;
constexpr int64_t deci2Poll=4ULL;
constexpr int64_t deci2ExRecv=5ULL;
constexpr int64_t deci2ExSend=6ULL;
constexpr int64_t deci2KPuts=7ULL;
constexpr int64_t deci2HeaderSize=12ULL;
{
switch(fn){
case cast<uint32_t>(1ULL):{
m->deci2Sockets++;
m->deci2Desc[m->deci2Sockets] = ps2_Machine_Read32(m,cast<uint32_t>((param + cast<uint32_t>(4ULL))));
ps2_Machine_setRet(m,m->deci2Sockets);
break;}
case cast<uint32_t>(3ULL):{
uint32_t desc = get(m->deci2Desc,ps2_Machine_Read32(m,param));
if ((desc == cast<uint32_t>(0ULL))) {
ps2_Machine_setRet(m,cast<uint32_t>(4294967295ULL));
return ;
}
uint32_t pkt = ps2_Machine_Read32(m,cast<uint32_t>((desc + cast<uint32_t>(16ULL))));
int64_t n = cast<int64_t>(cast<uint32_t>((ps2_Machine_Read32(m,pkt) & cast<uint32_t>(65535ULL))));
if ((n > cast<int64_t>(12ULL))) {
m->tty = append(m->tty,ps2_Machine_ReadMem(m,cast<uint32_t>((pkt + cast<uint32_t>(12ULL))),cast<int64_t>((n - cast<int64_t>(12ULL)))));
}
ps2_Machine_Write32(m,cast<uint32_t>((desc + cast<uint32_t>(12ULL))),cast<uint32_t>(0ULL));
ps2_Machine_setRet(m,cast<uint32_t>(0ULL));
break;}
case cast<uint32_t>(4ULL):{
ps2_Machine_setRet(m,cast<uint32_t>(0ULL));
break;}
case cast<uint32_t>(7ULL):{
m->tty = append(m->tty,ps2_Machine_CString(m,ps2_Machine_Read32(m,param)));
ps2_Machine_setRet(m,cast<uint32_t>(1ULL));
break;}
default:{
ps2_Machine_setRet(m,cast<uint32_t>(0ULL));
break;}
}}
}
}
// tools/platform/ps2/kernel.go:390:1
std::string ps2_Machine_SyscallCensus(ps2_Machine* m){
{
if ((len(m->SyscallCalls) == cast<int64_t>(0ULL))) {
return std::string("no kernel calls\012",16);
}
Map<std::string,bool> handled = Map<std::string,bool>{};
{auto&& tmp317 = ps2_syscalls;
for(auto [tmp318,tmp319]:tmp317){
auto e=tmp319;if (bool(e.fn)) {
handled[e.name] = true;
}
}}
Slice<ps2_kernel_401_kv> all={};
{auto&& tmp320 = m->SyscallCalls;
for(auto [tmp321,tmp322]:tmp320){
auto k=tmp321;auto v=tmp322;all = append(all,Slice<ps2_kernel_401_kv>{ps2_kernel_401_kv{k,v}});
}}
go_sort_Slice(all,[&](int64_t i,int64_t j)->bool{
if ((all[i].n != all[j].n)) {
return (all[i].n > all[j].n);
}
return (all[i].name < all[j].name);
}
);
std::string s = std::string("kernel calls:\012",14);
int64_t unmodelled = cast<int64_t>(0ULL);
{auto&& tmp323 = all;
for(int64_t tmp324=0;tmp324<len(tmp323);++tmp324){
auto e=tmp323[tmp324];std::string mark = std::string(" ",1);
if ((!get(handled,e.name))) {
mark = std::string("*",1);
unmodelled++;
}
s += go_fmt_Sprintf(std::string("  %s %-24s %d\012",14),mark,e.name,e.n);
}}
if ((unmodelled > cast<int64_t>(0ULL))) {
s += go_fmt_Sprintf(std::string("  (* = unmodelled: %d call%s to write)\012",39),unmodelled,ps2_plural(unmodelled));
}
return s;
}
}
// tools/platform/ps2/kernel.go:432:1
std::string ps2_plural(int64_t n){
{
if ((n == cast<int64_t>(1ULL))) {
return std::string("",0);
}
return std::string("s",1);
}
}
// tools/platform/ps2/machine.go:407:1
ps2_Machine* ps2_NewMachine(){
{
ps2_Machine* m = arenaNew(ps2_Machine{Slice<uint8_t>::make(cast<int64_t>(33554432ULL)),Slice<uint8_t>::make(cast<int64_t>(16384ULL)),Slice<uint8_t>::make(cast<int64_t>(2097152ULL)),{},Map<uint32_t,uint32_t>{},Map<uint32_t,int64_t>{},{},{},{},{},Map<std::string,int64_t>{},{},{},{},Map<uint32_t,ps2_thread*>{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},Map<uint32_t,ps2_sema*>{},cast<uint32_t>(1ULL),Map<uint32_t,uint32_t>{},{},{},{},Map<uint32_t,uint32_t>{},Map<uint32_t,uint32_t>{},Map<uint32_t,int64_t>{},{},{},{},{},{},{},{},{},{},{},Map<uint32_t,int64_t>{},Map<uint32_t,int64_t>{},Map<uint32_t,int64_t>{},Map<ps2_sifRPCKey,int64_t>{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},cast<int64_t>(-1ULL),{},{},{},{},{},Map<uint32_t,bool>{},{},{},{},{},{},{},{},{},{},{},{},{},{},Map<std::string,bool>{},{},{}});
m->CPU = r5900_NewCPU(m);
m->CPU->Syscall = [=](auto...args){return ps2_Machine_handleSyscall(m,args...);};
m->noIdleSkip = true;
ps2_Machine_ensureVIF(m,cast<int64_t>(0ULL));
return m;
}
}
// tools/platform/ps2/machine.go:446:1
ps2_IOP* ps2_Machine_StartIOP(ps2_Machine* m){
{
ps2_IOP* old = m->IOP;
m->IOP = ps2_newIOP(m,m->iopRAM);
if (bool(old)) {
m->IOP->logPC = old->logPC;
}
if (bool(m->OnIOPStart)) {
m->OnIOPStart(m->IOP);
}
return m->IOP;
}
}
// tools/platform/ps2/machine.go:464:1
void ps2_Machine_iopPrint(ps2_Machine* m,std::string s){
{
m->iopTTYLine = append(m->iopTTYLine,s);
m->IOP->tty = append(m->IOP->tty,s);
{;for (;;){
int64_t i = go_bytes_IndexByte(m->iopTTYLine,cast<uint8_t>(10ULL));
if ((i < cast<int64_t>(0ULL))) {
break;
}
ps2_Machine_note(m,std::string("iop: %s",7),cast<std::string>(sub(m->iopTTYLine,0,i)));
m->iopTTYLine = sub(m->iopTTYLine,cast<int64_t>((i + cast<int64_t>(1ULL))),len(m->iopTTYLine));
}
}}
}
// tools/platform/ps2/machine.go:479:1
void ps2_Machine_SetImageHash(ps2_Machine* m,std::string md5){
{
m->imageHash = md5;
}
}
// tools/platform/ps2/machine.go:482:1
void ps2_Machine_SetVolume(ps2_Machine* m,iso9660_Volume* v){
{
m->vol = v;
}
}
// tools/platform/ps2/machine.go:485:1
iso9660_Volume* ps2_Machine_Volume(ps2_Machine* m){
{
return m->vol;
}
}
// tools/platform/ps2/machine.go:497:1
void ps2_Machine_SetBIOS(ps2_Machine* m,Slice<uint8_t> raw){
{
m->bios = raw;
}
}
// tools/platform/ps2/machine.go:500:1
Slice<uint8_t> ps2_Machine_BIOS(ps2_Machine* m){
{
return m->bios;
}
}
// tools/platform/ps2/machine.go:503:1
ps2_Executable* ps2_Machine_Exe(ps2_Machine* m){
{
return m->exe;
}
}
// tools/platform/ps2/machine.go:509:1
void ps2_Machine_LoadExecutable(ps2_Machine* m,ps2_Executable* e){
{
m->exe = e;
{auto&& tmp325 = e->Segments;
for(int64_t tmp326=0;tmp326<len(tmp325);++tmp326){
auto seg=tmp325[tmp326];gcopy(sub(m->ram,seg.VAddr,len(m->ram)),seg.Data);
{uint32_t i = cast<uint32_t>(len(seg.Data));for (;(i < seg.MemSz);i++){
m->ram[cast<uint32_t>((seg.VAddr + i))] = cast<uint8_t>(0ULL);
}
}}}
ps2_Machine_mapMemory(m);
m->CPU->COP0[cast<int64_t>(12ULL)] = cast<uint64_t>(1879179008ULL);
r5900_CPU_SetPC(m->CPU,cast<uint64_t>(e->Entry));
r5900_CPU_SetReg(m->CPU,cast<uint32_t>(29ULL),cast<uint64_t>(33538048ULL));
r5900_CPU_SetReg(m->CPU,cast<uint32_t>(28ULL),cast<uint64_t>(0ULL));
m->heapPtr = cast<uint32_t>(29360128ULL);
m->heapEnd = cast<uint32_t>(32505856ULL);
m->threads = Map<uint32_t,ps2_thread*>{{cast<uint32_t>(1ULL),arenaNew(ps2_thread{cast<uint32_t>(1ULL),e->Entry,{},{},{},cast<uint32_t>(64ULL),cast<ps2_threadState>(0ULL),{},{}})}};
m->currentThread = cast<uint32_t>(1ULL);
m->nextThreadID = cast<uint32_t>(2ULL);
}
}
// tools/platform/ps2/machine.go:565:1
void ps2_Machine_mapMemory(ps2_Machine* m){
{
constexpr int64_t pageMask16M=33546240ULL;
constexpr int64_t page16M=4096ULL;
constexpr int64_t flags=31ULL;
auto entry = [&](int64_t i,uint32_t vaddr,uint32_t paddr)->void{
uint64_t pfn = cast<uint64_t>(shr<uint32_t>(paddr,cast<int64_t>(12ULL)));
r5900_CPU_SetTLB(m->CPU,i,r5900_TLBEntry{cast<uint32_t>(33546240ULL),cast<uint64_t>(vaddr),cast<uint64_t>(((shl<uint64_t>(pfn,cast<int64_t>(6ULL))) | cast<uint64_t>(31ULL))),cast<uint64_t>(((shl<uint64_t>((cast<uint64_t>((pfn + cast<uint64_t>(4096ULL)))),cast<int64_t>(6ULL))) | cast<uint64_t>(31ULL)))});
}
;
entry(cast<int64_t>(0ULL),cast<uint32_t>(0ULL),cast<uint32_t>(0ULL));
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(8ULL));i++){
entry(cast<int64_t>((cast<int64_t>(1ULL) + i)),cast<uint32_t>((cast<uint32_t>(268435456ULL) + cast<uint32_t>((cast<uint32_t>(i) * cast<uint32_t>(33554432ULL))))),cast<uint32_t>((cast<uint32_t>(268435456ULL) + cast<uint32_t>((cast<uint32_t>(i) * cast<uint32_t>(33554432ULL))))));
}
}entry(cast<int64_t>(9ULL),cast<uint32_t>(536870912ULL),cast<uint32_t>(0ULL));
entry(cast<int64_t>(10ULL),cast<uint32_t>(805306368ULL),cast<uint32_t>(0ULL));
}
}
// tools/platform/ps2/machine.go:616:1
uint32_t ps2_phys(uint32_t addr){
{
if (((addr >= cast<uint32_t>(1879048192ULL)) && (addr < cast<uint32_t>(1879064576ULL)))) {
return addr;
}
return cast<uint32_t>((addr & cast<uint32_t>(536870911ULL)));
}
}
// tools/platform/ps2/machine.go:624:1
bool ps2_Machine_mapped(ps2_Machine* m,uint32_t p){
{
return ((p < cast<uint32_t>(33554432ULL)) || (((p >= cast<uint32_t>(1879048192ULL)) && (p < cast<uint32_t>(1879064576ULL)))));
}
}
// tools/platform/ps2/machine.go:629:1
std::tuple<Slice<uint8_t>,uint32_t,bool> ps2_Machine_slice(ps2_Machine* m,uint32_t p){
{
{
if ((p < cast<uint32_t>(33554432ULL))){
return {m->ram,p,true};
}
else if (((p >= cast<uint32_t>(1879048192ULL)) && (p < cast<uint32_t>(1879064576ULL)))){
return {m->spram,cast<uint32_t>((p - cast<uint32_t>(1879048192ULL))),true};
}
else if (((p >= cast<uint32_t>(469762048ULL)) && (p < cast<uint32_t>(471859200ULL)))){
return {m->iopRAM,cast<uint32_t>((p - cast<uint32_t>(469762048ULL))),true};
}
else if (((p >= cast<uint32_t>(285212672ULL)) && (p < cast<uint32_t>(285278208ULL)))){
uint32_t q = cast<uint32_t>((shr<uint32_t>(p,cast<int64_t>(14ULL)) & cast<uint32_t>(3ULL)));
ps2_vif* v = ps2_Machine_ensureVIF(m,cast<int64_t>(shr<uint32_t>(q,cast<int64_t>(1ULL))));
Slice<uint8_t> buf = v->micro;
if ((cast<uint32_t>((q & cast<uint32_t>(1ULL))) != cast<uint32_t>(0ULL))) {
buf = v->data;
}
return {buf,cast<uint32_t>(((cast<uint32_t>((p - cast<uint32_t>(285212672ULL)))) & cast<uint32_t>(cast<int64_t>((len(buf) - cast<int64_t>(1ULL)))))),true};
}
}
tmp327:;
return {{},cast<uint32_t>(0ULL),false};
}
}
// tools/platform/ps2/machine.go:654:1
uint8_t ps2_Machine_Read(ps2_Machine* m,uint32_t addr){
{
uint32_t p = ps2_phys(addr);
{
auto tmp328 = ps2_Machine_slice(m,p);
Slice<uint8_t> buf = std::get<0>(tmp328);
uint32_t off = std::get<1>(tmp328);
bool ok = std::get<2>(tmp328);
if (ok) {
uint8_t v = buf[off];
ps2_Machine_noteRead(m,p,cast<uint32_t>(v));
return v;
}
}
uint32_t v = ps2_Machine_ioRead(m,p);
return cast<uint8_t>(shr<uint32_t>(v,(cast<uint32_t>((cast<uint32_t>(8ULL) * (cast<uint32_t>((p & cast<uint32_t>(3ULL)))))))));
}
}
// tools/platform/ps2/machine.go:665:1
void ps2_Machine_Write(ps2_Machine* m,uint32_t addr,uint8_t v){
{
uint32_t p = ps2_phys(addr);
m->stores++;
{
auto tmp329 = ps2_Machine_slice(m,p);
Slice<uint8_t> buf = std::get<0>(tmp329);
uint32_t off = std::get<1>(tmp329);
bool ok = std::get<2>(tmp329);
if (ok) {
buf[off] = v;
ps2_Machine_noteWrite(m,p,cast<uint32_t>(v));
return ;
}
}
uint32_t sh = cast<uint32_t>((cast<uint32_t>(8ULL) * (cast<uint32_t>((p & cast<uint32_t>(3ULL))))));
uint32_t w = cast<uint32_t>(((get(m->io,(p & ~(cast<uint32_t>(3ULL)))) & ~((shl<uint32_t>(cast<uint32_t>(255ULL),sh)))) | shl<uint32_t>(cast<uint32_t>(v),sh)));
ps2_Machine_ioWrite(m,(p & ~(cast<uint32_t>(3ULL))),w);
}
}
// tools/platform/ps2/machine.go:679:1
uint32_t ps2_Machine_Read32(ps2_Machine* m,uint32_t addr){
{
uint32_t p = ps2_phys(addr);
{
auto tmp330 = ps2_Machine_slice(m,p);
Slice<uint8_t> buf = std::get<0>(tmp330);
uint32_t off = std::get<1>(tmp330);
bool ok = std::get<2>(tmp330);
if ((ok && (cast<uint32_t>((off + cast<uint32_t>(4ULL))) <= cast<uint32_t>(len(buf))))) {
uint32_t v = cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(buf[off]) | shl<uint32_t>(cast<uint32_t>(buf[cast<uint32_t>((off + cast<uint32_t>(1ULL)))]),cast<int64_t>(8ULL)))) | shl<uint32_t>(cast<uint32_t>(buf[cast<uint32_t>((off + cast<uint32_t>(2ULL)))]),cast<int64_t>(16ULL)))) | shl<uint32_t>(cast<uint32_t>(buf[cast<uint32_t>((off + cast<uint32_t>(3ULL)))]),cast<int64_t>(24ULL))));
ps2_Machine_noteRead(m,p,v);
return v;
}
}
return ps2_Machine_ioRead(m,p);
}
}
// tools/platform/ps2/machine.go:689:1
void ps2_Machine_Write32(ps2_Machine* m,uint32_t addr,uint32_t v){
{
uint32_t p = ps2_phys(addr);
m->stores++;
{
auto tmp331 = ps2_Machine_slice(m,p);
Slice<uint8_t> buf = std::get<0>(tmp331);
uint32_t off = std::get<1>(tmp331);
bool ok = std::get<2>(tmp331);
if ((ok && (cast<uint32_t>((off + cast<uint32_t>(4ULL))) <= cast<uint32_t>(len(buf))))) {
buf[off] = cast<uint8_t>(v);
buf[cast<uint32_t>((off + cast<uint32_t>(1ULL)))] = cast<uint8_t>(shr<uint32_t>(v,cast<int64_t>(8ULL)));
buf[cast<uint32_t>((off + cast<uint32_t>(2ULL)))] = cast<uint8_t>(shr<uint32_t>(v,cast<int64_t>(16ULL)));
buf[cast<uint32_t>((off + cast<uint32_t>(3ULL)))] = cast<uint8_t>(shr<uint32_t>(v,cast<int64_t>(24ULL)));
ps2_Machine_noteWrite(m,p,v);
return ;
}
}
ps2_Machine_ioWrite(m,p,v);
}
}
// tools/platform/ps2/machine.go:706:1
uint32_t ps2_Machine_Fetch32(ps2_Machine* m,uint32_t addr){
{const uint32_t p=addr&0x1fffffffu;if(p+3u<32u*1024*1024){uint32_t value;std::memcpy(&value,m->ram.p+p,4);return value;}
{
uint32_t p = ps2_phys(addr);
{
auto tmp332 = ps2_Machine_slice(m,p);
Slice<uint8_t> buf = std::get<0>(tmp332);
uint32_t off = std::get<1>(tmp332);
bool ok = std::get<2>(tmp332);
if ((ok && (cast<uint32_t>((off + cast<uint32_t>(4ULL))) <= cast<uint32_t>(len(buf))))) {
return cast<uint32_t>((cast<uint32_t>((cast<uint32_t>((cast<uint32_t>(buf[off]) | shl<uint32_t>(cast<uint32_t>(buf[cast<uint32_t>((off + cast<uint32_t>(1ULL)))]),cast<int64_t>(8ULL)))) | shl<uint32_t>(cast<uint32_t>(buf[cast<uint32_t>((off + cast<uint32_t>(2ULL)))]),cast<int64_t>(16ULL)))) | shl<uint32_t>(cast<uint32_t>(buf[cast<uint32_t>((off + cast<uint32_t>(3ULL)))]),cast<int64_t>(24ULL))));
}
}
return cast<uint32_t>(0ULL);
}
}
}
// tools/platform/ps2/machine.go:714:1
void ps2_Machine_noteRead(ps2_Machine* m,uint32_t p,uint32_t v){
{
if ((((bool(m->OnRead) && (!m->hookMuted)) && (p >= m->RWatchLo)) && (p < m->RWatchHi))) {
m->OnRead(p,v,cast<uint32_t>(r5900_CPU_CurPC(m->CPU)));
}
}
}
// tools/platform/ps2/machine.go:720:1
void ps2_Machine_noteWrite(ps2_Machine* m,uint32_t p,uint32_t v){
{
if (((bool(m->OnWrite) && (p >= m->WatchLo)) && (p < m->WatchHi))) {
m->OnWrite(p,v,cast<uint32_t>(r5900_CPU_CurPC(m->CPU)));
}
}
}
// tools/platform/ps2/machine.go:736:1
uint32_t ps2_Machine_ioRead(ps2_Machine* m,uint32_t p){
{
if (((p >= cast<uint32_t>(268497408ULL)) && (p < cast<uint32_t>(268497536ULL)))) {
return ps2_Machine_sbusRead(m,cast<uint32_t>((p - cast<uint32_t>(268497408ULL))));
}
if (((p >= cast<uint32_t>(268468224ULL)) && (p < cast<uint32_t>(268498432ULL)))) {
{
auto tmp333 = ps2_Machine_dmacRead(m,p);
uint32_t v = std::get<0>(tmp333);
bool ok = std::get<1>(tmp333);
if (ok) {
return v;
}
}
}
{
switch(p){
case cast<uint32_t>(268496896ULL):{
return m->intcStat;
break;}
case cast<uint32_t>(268496912ULL):{
return m->intcMask;
break;}
}}
if (((p >= cast<uint32_t>(301989888ULL)) && (p < cast<uint32_t>(301998080ULL)))) {
{
auto tmp334 = ps2_Machine_gsPrivRead(m,p);
uint32_t v = std::get<0>(tmp334);
bool ok = std::get<1>(tmp334);
if (ok) {
return v;
}
}
}
{
auto tmp335 = ps2_Machine_eeTimerRead(m,p);
uint32_t v = std::get<0>(tmp335);
bool ok = std::get<1>(tmp335);
if (ok) {
return v;
}
}
m->unmodelled[p]++;
return get(m->io,p);
}
}
// tools/platform/ps2/machine.go:766:1
void ps2_Machine_ioWrite(ps2_Machine* m,uint32_t p,uint32_t v){
{
if (((p >= cast<uint32_t>(268497408ULL)) && (p < cast<uint32_t>(268497536ULL)))) {
ps2_Machine_sbusWrite(m,cast<uint32_t>((p - cast<uint32_t>(268497408ULL))),v);
return ;
}
if (((p >= cast<uint32_t>(268468224ULL)) && (p < cast<uint32_t>(268498432ULL)))) {
if (ps2_Machine_dmacWrite(m,p,v)) {
return ;
}
}
{
switch(p){
case cast<uint32_t>(268496896ULL):{
m->intcStat &= ~(v);
return ;
break;}
case cast<uint32_t>(268496912ULL):{
m->intcMask ^= v;
return ;
break;}
}}
if (((p >= cast<uint32_t>(301989888ULL)) && (p < cast<uint32_t>(301998080ULL)))) {
if (ps2_Machine_gsPrivWrite(m,p,v)) {
return ;
}
}
if (ps2_Machine_eeTimerWrite(m,p,v)) {
return ;
}
m->unmodelled[p]++;
m->io[p] = v;
}
}
// tools/platform/ps2/machine.go:800:1
Map<uint32_t,int64_t> ps2_Machine_Unmodelled(ps2_Machine* m){
{
return m->unmodelled;
}
}
// tools/platform/ps2/machine.go:806:1
std::string ps2_Machine_HardwareCensus(ps2_Machine* m){
{
if ((len(m->unmodelled) == cast<int64_t>(0ULL))) {
return std::string("no unmodelled hardware touched\012",31);
}
Map<std::string,ps2_machine_810_unit*> units = Map<std::string,ps2_machine_810_unit*>{};
{auto&& tmp336 = m->unmodelled;
for(auto [tmp337,tmp338]:tmp336){
auto addr=tmp337;auto n=tmp338;std::string name = ps2_RegionName(addr);
ps2_machine_810_unit* u = get(units,name);
if ((!u)) {
u = arenaNew(ps2_machine_810_unit{{},Map<uint32_t,int64_t>{}});
units[name] = u;
}
u->hits += n;
u->regs[addr] += n;
}}
Slice<ps2_machine_826_kv> all={};
{auto&& tmp339 = units;
for(auto [tmp340,tmp341]:tmp339){
auto k=tmp340;auto v=tmp341;all = append(all,Slice<ps2_machine_826_kv>{ps2_machine_826_kv{k,v}});
}}
go_sort_Slice(all,[&](int64_t i,int64_t j)->bool{
return (all[i].u->hits > all[j].u->hits);
}
);
std::string s = std::string("unmodelled hardware touched (the work list):\012",45);
{auto&& tmp342 = all;
for(int64_t tmp343=0;tmp343<len(tmp342);++tmp343){
auto e=tmp342[tmp343];s += go_fmt_Sprintf(std::string("  %-8s %8d accesses over %d registers\012",38),e.name,e.u->hits,len(e.u->regs));
Slice<uint32_t> regs={};
{auto&& tmp344 = e.u->regs;
for(auto [tmp345,tmp346]:tmp344){
auto r=tmp345;regs = append(regs,Slice<uint32_t>{r});
}}
go_sort_Slice(regs,[&](int64_t i,int64_t j)->bool{
return (get(e.u->regs,regs[i]) > get(e.u->regs,regs[j]));
}
);
{auto&& tmp347 = regs;
for(int64_t tmp348=0;tmp348<len(tmp347);++tmp348){
auto i=tmp348;auto r=tmp347[tmp348];if ((i == cast<int64_t>(4ULL))) {
s += go_fmt_Sprintf(std::string("      ... and %d more\012",22),cast<int64_t>((len(regs) - cast<int64_t>(4ULL))));
break;
}
s += go_fmt_Sprintf(std::string("      0x%08X  %d\012",17),r,get(e.u->regs,r));
}}
}}
return s;
}
}
// tools/platform/ps2/machine.go:859:1
std::string ps2_RegionName(uint32_t p){
{
{
if (((p >= cast<uint32_t>(268435456ULL)) && (p < cast<uint32_t>(268443648ULL)))){
return std::string("TIMER",5);
}
else if (((p >= cast<uint32_t>(268443648ULL)) && (p < cast<uint32_t>(268447744ULL)))){
return std::string("IPU",3);
}
else if (((p >= cast<uint32_t>(268447744ULL)) && (p < cast<uint32_t>(268449792ULL)))){
return std::string("GIF",3);
}
else if (((p >= cast<uint32_t>(268449792ULL)) && (p < cast<uint32_t>(268451840ULL)))){
return std::string("VIF",3);
}
else if (((p >= cast<uint32_t>(268451840ULL)) && (p < cast<uint32_t>(268468224ULL)))){
return std::string("VIF-FIFO",8);
}
else if (((p >= cast<uint32_t>(268468224ULL)) && (p < cast<uint32_t>(268496896ULL)))){
return std::string("DMAC",4);
}
else if (((p >= cast<uint32_t>(268496896ULL)) && (p < cast<uint32_t>(268497408ULL)))){
return std::string("INTC",4);
}
else if (((p >= cast<uint32_t>(268497408ULL)) && (p < cast<uint32_t>(268498432ULL)))){
return std::string("SIF",3);
}
else if (((p >= cast<uint32_t>(268497920ULL)) && (p < cast<uint32_t>(268500992ULL)))){
return std::string("MCH",3);
}
else if (((p >= cast<uint32_t>(285212672ULL)) && (p < cast<uint32_t>(285278208ULL)))){
return std::string("VU",2);
}
else if (((p >= cast<uint32_t>(301989888ULL)) && (p < cast<uint32_t>(301998080ULL)))){
return std::string("GS",2);
}
else if (((p >= cast<uint32_t>(469762048ULL)) && (p < cast<uint32_t>(471859200ULL)))){
return std::string("IOP-RAM",7);
}
else if ((p >= cast<uint32_t>(532676608ULL))){
return std::string("BIOS",4);
}
}
tmp349:;
return std::string("?",1);
}
}
// tools/platform/ps2/machine.go:897:1
void ps2_Machine_iopUnknownCDVD(ps2_Machine* m,uint32_t a,bool write){
{
std::string rw = std::string("read",4);
if (write) {
rw = std::string("written",7);
}
ps2_Machine_note(m,std::string("CDVD: register 0x%08X %s, and nothing models it",47),a,rw);
m->IOP->unmodelledIO[a]++;
}
}
// tools/platform/ps2/machine.go:922:1
void ps2_Machine_SetBreakpoint(ps2_Machine* m,uint32_t vaddr){
{
m->breakpoints[vaddr] = true;
}
}
// tools/platform/ps2/machine.go:925:1
void ps2_Machine_ClearBreakpoints(ps2_Machine* m){
{
m->breakpoints = Map<uint32_t,bool>{};
}
}
// tools/platform/ps2/machine.go:928:1
std::string ps2_Machine_TTY(ps2_Machine* m){
{
return cast<std::string>(m->tty);
}
}
// tools/platform/ps2/machine.go:931:1
uint32_t ps2_Machine_VBlanks(ps2_Machine* m){
{
return m->vblanks;
}
}
// tools/platform/ps2/machine.go:950:1
uint16_t ps2_Machine_padButtons(ps2_Machine* m){
{
uint16_t b = m->padLive.buttons;
{auto&& tmp350 = m->PadScript;
for(int64_t tmp351=0;tmp351<len(tmp350);++tmp351){
auto pr=tmp350[tmp351];if (((m->vblanks >= pr.At) && (m->vblanks < cast<uint32_t>((pr.At + pr.Hold))))) {
b |= pr.Buttons;
}
}}
return b;
}
}
// tools/platform/ps2/machine.go:998:1
void ps2_Machine_SetPadButtons(ps2_Machine* m,uint16_t b){
{
m->padLive.buttons = b;
}
}
// tools/platform/ps2/machine.go:1002:1
uint16_t ps2_Machine_PadButtons(ps2_Machine* m){
{
return m->padLive.buttons;
}
}
// tools/platform/ps2/machine.go:1010:1
void ps2_Machine_SetPadStick(ps2_Machine* m,uint8_t lx,uint8_t ly,uint8_t rx,uint8_t ry){
{
auto d = [&](uint8_t v)->int8_t{
return cast<int8_t>(cast<int64_t>((cast<int64_t>(v) - cast<int64_t>(128ULL))));
}
;
auto tmp352 = std::make_tuple(d(lx),d(ly),d(rx),d(ry));
m->padLive.lx = std::get<0>(tmp352);
m->padLive.ly = std::get<1>(tmp352);
m->padLive.rx = std::get<2>(tmp352);
m->padLive.ry = std::get<3>(tmp352);
}
}
// tools/platform/ps2/machine.go:1016:1
std::tuple<uint8_t,uint8_t,uint8_t,uint8_t> ps2_Machine_PadStick(ps2_Machine* m){
uint8_t lx{};
uint8_t ly{};
uint8_t rx{};
uint8_t ry{};
{
return ps2_Machine_padSticks(m);
}
}
// tools/platform/ps2/machine.go:1019:1
std::tuple<uint8_t,uint8_t,uint8_t,uint8_t> ps2_Machine_padSticks(ps2_Machine* m){
uint8_t lx{};
uint8_t ly{};
uint8_t rx{};
uint8_t ry{};
{
auto w = [&](int8_t v)->uint8_t{
return cast<uint8_t>(cast<int64_t>((cast<int64_t>(128ULL) + cast<int64_t>(v))));
}
;
return {w(m->padLive.lx),w(m->padLive.ly),w(m->padLive.rx),w(m->padLive.ry)};
}
}
// tools/platform/ps2/machine.go:1040:1
std::tuple<uint16_t,bool> ps2_PadButton(std::string name){
{
auto tmp353 = lookup(ps2_padButtonBits,go_strings_ToLower(name));
uint16_t b = std::get<0>(tmp353);
bool ok = std::get<1>(tmp353);
return {b,ok};
}
}
// tools/platform/ps2/machine.go:1047:1
Slice<std::string> ps2_PadButtonNames(){
{
Slice<std::string> out = Slice<std::string>::make(cast<int64_t>(0ULL),len(ps2_padButtonBits));
{auto&& tmp354 = ps2_padButtonBits;
for(auto [tmp355,tmp356]:tmp354){
auto n=tmp355;out = append(out,Slice<std::string>{n});
}}
go_sort_Strings(out);
return out;
}
}
// tools/platform/ps2/machine.go:1057:1
Slice<uint8_t> ps2_Machine_ReadMem(ps2_Machine* m,uint32_t addr,int64_t n){
{
m->hookMuted = true;
auto tmp357=defer([&](){[&]()->void{
m->hookMuted = false;
}
();});
Slice<uint8_t> out = Slice<uint8_t>::make(n);
{auto&& tmp358 = out;
for(int64_t tmp359=0;tmp359<len(tmp358);++tmp359){
auto i=tmp359;out[i] = ps2_Machine_Read(m,cast<uint32_t>((addr + cast<uint32_t>(i))));
}}
return out;
}
}
// tools/platform/ps2/machine.go:1068:1
std::string ps2_Machine_CString(ps2_Machine* m,uint32_t addr){
{
m->hookMuted = true;
auto tmp360=defer([&](){[&]()->void{
m->hookMuted = false;
}
();});
Slice<uint8_t> b={};
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(1024ULL));i++){
uint8_t c = ps2_Machine_Read(m,cast<uint32_t>((addr + cast<uint32_t>(i))));
if ((c == cast<uint8_t>(0ULL))) {
break;
}
b = append(b,Slice<uint8_t>{c});
}
}return cast<std::string>(b);
}
}
// tools/platform/ps2/machine.go:1083:1
std::string ps2_Machine_DisasmAt(ps2_Machine* m,uint32_t vaddr){
{
m->hookMuted = true;
auto tmp361=defer([&](){[&]()->void{
m->hookMuted = false;
}
();});
return r5900_DecodeWord(ps2_Machine_Fetch32(m,vaddr),vaddr).Text;
}
}
// tools/platform/ps2/machine.go:1093:1
std::string ps2_Machine_Sym(ps2_Machine* m,uint32_t addr){
{
if (bool(m->exe)) {
{
auto tmp362 = ps2_Executable_Lookup(m->exe,addr);
std::string name = std::get<0>(tmp362);
uint32_t off = std::get<1>(tmp362);
bool ok = std::get<2>(tmp362);
if (ok) {
if ((off == cast<uint32_t>(0ULL))) {
return name;
}
return go_fmt_Sprintf(std::string("%s+0x%X",7),name,off);
}
}
}
if ((len(m->extraSyms) > cast<int64_t>(0ULL))) {
int64_t i = cast<int64_t>((go_sort_Search(len(m->extraSyms),[&](int64_t i)->bool{
return (m->extraSyms[i].Addr > addr);
}
) - cast<int64_t>(1ULL)));
if (((i >= cast<int64_t>(0ULL)) && (cast<uint32_t>((addr - m->extraSyms[i].Addr)) < cast<uint32_t>(16384ULL)))) {
ps2_Symbol s = m->extraSyms[i];
if ((s.Addr == addr)) {
return s.Name;
}
return go_fmt_Sprintf(std::string("%s+0x%X",7),s.Name,cast<uint32_t>((addr - s.Addr)));
}
}
return go_fmt_Sprintf(std::string("0x%08X",6),addr);
}
}
// tools/platform/ps2/machine.go:1120:1
void ps2_Machine_AddSymbols(ps2_Machine* m,Slice<ps2_Symbol> syms){
{
m->extraSyms = append(m->extraSyms,syms);
go_sort_Slice(m->extraSyms,[&](int64_t i,int64_t j)->bool{
return (m->extraSyms[i].Addr < m->extraSyms[j].Addr);
}
);
}
}
// tools/platform/ps2/machine.go:1126:1
std::string ps2_Machine_Registers(ps2_Machine* m){
{
std::array<std::string,32> names = std::array<std::string,32>{std::string("zero",4),std::string("at",2),std::string("v0",2),std::string("v1",2),std::string("a0",2),std::string("a1",2),std::string("a2",2),std::string("a3",2),std::string("t0",2),std::string("t1",2),std::string("t2",2),std::string("t3",2),std::string("t4",2),std::string("t5",2),std::string("t6",2),std::string("t7",2),std::string("s0",2),std::string("s1",2),std::string("s2",2),std::string("s3",2),std::string("s4",2),std::string("s5",2),std::string("s6",2),std::string("s7",2),std::string("t8",2),std::string("t9",2),std::string("k0",2),std::string("k1",2),std::string("gp",2),std::string("sp",2),std::string("fp",2),std::string("ra",2)};
std::string s = go_fmt_Sprintf(std::string("pc=%s\012",6),ps2_Machine_Sym(m,cast<uint32_t>(m->CPU->PC)));
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(32ULL));i += cast<int64_t>(4ULL)){
{int64_t j = cast<int64_t>(0ULL);for (;(j < cast<int64_t>(4ULL));j++){
s += go_fmt_Sprintf(std::string("%4s=%016X  ",11),names[cast<int64_t>((i + j))],r5900_CPU_Reg(m->CPU,cast<uint32_t>(cast<int64_t>((i + j)))));
}
}s += std::string("\012",1);
}
}return s;
}
}
// tools/platform/ps2/profile.go:133:1
ps2_profCounters ps2_Machine_profCounters(ps2_Machine* m){
{
ps2_profCounters c = ps2_profCounters{m->steps,{},{},{},{},{},{},{}};
if (bool(m->IOP)) {
c.iopSteps = ps2_IOP_Steps(m->IOP);
}
{
ps2_GS* g = m->gs;
if (bool(g)) {
c.prims = cast<uint64_t>(g->prims);
c.frags = g->plotted;
auto tmp363 = std::make_tuple(g->rejZ,g->rejScissor,g->rejAlpha,g->rejDate);
c.rejZ = std::get<0>(tmp363);
c.rejScissor = std::get<1>(tmp363);
c.rejAlpha = std::get<2>(tmp363);
c.rejDate = std::get<3>(tmp363);
}
}
return c;
}
}
// tools/platform/ps2/profile.go:148:1
time_Time ps2_Machine_profStart(ps2_Machine* m){
{
if ((!m->Profile)) {
return time_Time{};
}
return go_time_Now();
}
}
// tools/platform/ps2/profile.go:155:1
void ps2_Machine_profEnd(ps2_Machine* m,int64_t bucket,time_Time t){
{
if (time_Time_IsZero(t)) {
return ;
}
m->prof.ns[bucket] += cast<int64_t>(go_time_Since(t));
m->prof.count[bucket]++;
}
}
// tools/platform/ps2/profile.go:165:1
void ps2_Machine_profDrainEnter(ps2_Machine* m){
{
if ((!m->Profile)) {
return ;
}
if ((m->prof.drainDepth == cast<int64_t>(0ULL))) {
m->prof.drainStart = go_time_Now();
}
m->prof.drainDepth++;
}
}
// tools/platform/ps2/profile.go:175:1
void ps2_Machine_profDrainExit(ps2_Machine* m){
{
if (((!m->Profile) || (m->prof.drainDepth == cast<int64_t>(0ULL)))) {
return ;
}
m->prof.drainDepth--;
if ((m->prof.drainDepth == cast<int64_t>(0ULL))) {
m->prof.ns[cast<int64_t>(0ULL)] += cast<int64_t>(go_time_Since(m->prof.drainStart));
m->prof.count[cast<int64_t>(0ULL)]++;
}
}
}
// tools/platform/ps2/profile.go:192:1
std::tuple<time_Time,int64_t> ps2_Machine_profVU1Start(ps2_Machine* m){
{
if ((!m->Profile)) {
return {time_Time{},cast<int64_t>(0ULL)};
}
return {go_time_Now(),m->prof.ns[cast<int64_t>(2ULL)]};
}
}
// tools/platform/ps2/profile.go:199:1
void ps2_Machine_profVU1End(ps2_Machine* m,time_Time t,int64_t rasterBefore){
{
if (time_Time_IsZero(t)) {
return ;
}
int64_t nested = cast<int64_t>((m->prof.ns[cast<int64_t>(2ULL)] - rasterBefore));
m->prof.ns[cast<int64_t>(1ULL)] += cast<int64_t>((cast<int64_t>(go_time_Since(t)) - nested));
m->prof.count[cast<int64_t>(1ULL)]++;
}
}
// tools/platform/ps2/profile.go:210:1
void ps2_Machine_profRunEnter(ps2_Machine* m){
{
if ((!m->Profile)) {
return ;
}
auto tmp364 = std::make_tuple(go_time_Now(),true);
m->prof.runStart = std::get<0>(tmp364);
m->prof.inRun = std::get<1>(tmp364);
}
}
// tools/platform/ps2/profile.go:217:1
void ps2_Machine_profRunExit(ps2_Machine* m){
{
if (((!m->Profile) || (!m->prof.inRun))) {
return ;
}
m->prof.frameNs += cast<int64_t>(go_time_Since(m->prof.runStart));
m->prof.inRun = false;
}
}
// tools/platform/ps2/profile.go:234:1
void ps2_Machine_profFrame(ps2_Machine* m){
{
if ((!m->Profile)) {
return ;
}
ps2_profState* p = (&m->prof);
int64_t total = p->frameNs;
if (p->inRun) {
total += cast<int64_t>(go_time_Since(p->runStart));
p->runStart = go_time_Now();
}
p->frameNs = cast<int64_t>(0ULL);
if ((p->drainDepth > cast<int64_t>(0ULL))) {
p->ns[cast<int64_t>(0ULL)] += cast<int64_t>(go_time_Since(p->drainStart));
p->drainStart = go_time_Now();
}
auto ms = [&](int64_t ns)->double{
return (cast<double>(ns) / 1e6);
}
;
auto tmp365 = std::make_tuple(p->ns[cast<int64_t>(0ULL)],p->ns[cast<int64_t>(1ULL)],p->ns[cast<int64_t>(2ULL)]);
int64_t drain = std::get<0>(tmp365);
int64_t vu1 = std::get<1>(tmp365);
int64_t raster = std::get<2>(tmp365);
int64_t decode = cast<int64_t>((cast<int64_t>((drain - vu1)) - raster));
if ((decode < cast<int64_t>(0ULL))) {
decode = cast<int64_t>(0ULL);
}
Slice<ps2_ProfileBucket> buckets = Slice<ps2_ProfileBucket>{ps2_ProfileBucket{std::string("gif/vif/dma decode (derived)",28),ms(decode),cast<int64_t>(p->count[cast<int64_t>(0ULL)])},ps2_ProfileBucket{std::string("vu1 (geometry)",14),ms(vu1),cast<int64_t>(p->count[cast<int64_t>(1ULL)])},ps2_ProfileBucket{std::string("rasterise",9),ms(raster),cast<int64_t>(p->count[cast<int64_t>(2ULL)])}};
int64_t other = cast<int64_t>((total - drain));
if ((other < cast<int64_t>(0ULL))) {
other = cast<int64_t>(0ULL);
}
buckets = append(buckets,Slice<ps2_ProfileBucket>{ps2_ProfileBucket{std::string("ee + iop + rest (derived)",25),ms(other),{}}});
ps2_profCounters now = ps2_Machine_profCounters(m);
ps2_profCounters d = ps2_profCounters{cast<uint64_t>((now.eeSteps - p->base.eeSteps)),cast<uint64_t>((now.iopSteps - p->base.iopSteps)),cast<uint64_t>((now.prims - p->base.prims)),cast<uint64_t>((now.frags - p->base.frags)),cast<uint64_t>((now.rejZ - p->base.rejZ)),cast<uint64_t>((now.rejScissor - p->base.rejScissor)),cast<uint64_t>((now.rejAlpha - p->base.rejAlpha)),cast<uint64_t>((now.rejDate - p->base.rejDate))};
p->last = ps2_FrameProfile{ms(total),buckets,Slice<ps2_ProfileCounter>{ps2_ProfileCounter{std::string("gs primitives",13),cast<int64_t>(d.prims)},ps2_ProfileCounter{std::string("fragments drawn",15),cast<int64_t>(d.frags)},ps2_ProfileCounter{std::string("depth-rejected",14),cast<int64_t>(d.rejZ)},ps2_ProfileCounter{std::string("scissor-rejected",16),cast<int64_t>(d.rejScissor)},ps2_ProfileCounter{std::string("alpha-rejected",14),cast<int64_t>(d.rejAlpha)},ps2_ProfileCounter{std::string("date-rejected",13),cast<int64_t>(d.rejDate)},ps2_ProfileCounter{std::string("vu1 kicks",9),cast<int64_t>(p->count[cast<int64_t>(1ULL)])},ps2_ProfileCounter{std::string("ee instructions",15),cast<int64_t>(d.eeSteps)},ps2_ProfileCounter{std::string("iop instructions",16),cast<int64_t>(d.iopSteps)}},(d.prims > cast<uint64_t>(0ULL))};
p->has = true;
auto tmp366 = std::make_tuple(std::array<int64_t,3>{},std::array<int64_t,3>{});
p->ns = std::get<0>(tmp366);
p->count = std::get<1>(tmp366);
p->base = now;
}
}
// tools/platform/ps2/profile.go:309:1
ps2_FrameProfile ps2_Machine_FrameProfile(ps2_Machine* m){
{
if ((!m->prof.has)) {
return ps2_FrameProfile{};
}
return m->prof.last;
}
}
// tools/platform/ps2/profile.go:320:1
void ps2_Machine_SetProfile(ps2_Machine* m,bool on){
{
m->Profile = on;
m->prof = ps2_profState{};
if (on) {
m->prof.base = ps2_Machine_profCounters(m);
}
}
}
// tools/platform/ps2/romdir.go:59:1
std::tuple<int64_t,bool> ps2_FindROMDIR(Slice<uint8_t> raw){
{
auto name = [&](int64_t off)->std::string{
if ((cast<int64_t>((off + cast<int64_t>(16ULL))) > len(raw))) {
return std::string("",0);
}
return go_strings_TrimRight(cast<std::string>(sub(raw,off,cast<int64_t>((off + cast<int64_t>(10ULL))))),std::string("\000",1));
}
;
{int64_t off = cast<int64_t>(0ULL);for (;(cast<int64_t>((off + cast<int64_t>(48ULL))) <= len(raw));off += cast<int64_t>(16ULL)){
if ((((name(off) == std::string("RESET",5)) && (name(cast<int64_t>((off + cast<int64_t>(16ULL)))) == std::string("ROMDIR",6))) && (name(cast<int64_t>((off + cast<int64_t>(32ULL)))) == std::string("EXTINFO",7)))) {
return {off,true};
}
}
}return {cast<int64_t>(0ULL),false};
}
}
// tools/platform/ps2/romdir.go:75:1
std::tuple<Slice<ps2_RomEntry>,Error> ps2_ReadROMDIR(Slice<uint8_t> raw){
{
Slice<ps2_RomEntry> out={};
auto tmp367 = ps2_FindROMDIR(raw);
int64_t off = std::get<0>(tmp367);
bool ok = std::get<1>(tmp367);
if ((!ok)) {
return {{},go_errors_New(std::string("ps2: no ROMDIR directory (no RESET/ROMDIR/EXTINFO records) in this image",72))};
}
int64_t body = cast<int64_t>(0ULL);
{;for (;;){
if ((cast<int64_t>((off + cast<int64_t>(16ULL))) > len(raw))) {
return {{},go_errors_New(std::string("ps2: ROMDIR ran off the end of the image without a terminator",61))};
}
std::string name = go_strings_TrimRight(cast<std::string>(sub(raw,off,cast<int64_t>((off + cast<int64_t>(10ULL))))),std::string("\000",1));
int64_t size = cast<int64_t>(le_Uint32(rrBorrow(raw,cast<int64_t>((off + cast<int64_t>(12ULL))),len(raw))));
if ((name == std::string("",0))) {
break;
}
if ((cast<int64_t>((body + size)) > len(raw))) {
return {{},go_errors_New(std::string("ps2: a ROMDIR entry claims more bytes than the image holds",58))};
}
out = append(out,Slice<ps2_RomEntry>{ps2_RomEntry{name,sub(raw,body,cast<int64_t>((body + size)))}});
body += ((cast<int64_t>((size + cast<int64_t>(15ULL)))) & ~(cast<int64_t>(15ULL)));
off += cast<int64_t>(16ULL);
}
}if ((len(out) == cast<int64_t>(0ULL))) {
return {{},go_errors_New(std::string("ps2: the ROMDIR is empty",24))};
}
return {out,{}};
}
}
// tools/platform/ps2/romdir.go:108:1
std::tuple<Slice<uint8_t>,bool> ps2_ROMDIREntry(Slice<uint8_t> raw,std::string name){
{
auto tmp368 = ps2_ReadROMDIR(raw);
Slice<ps2_RomEntry> all = std::get<0>(tmp368);
Error err = std::get<1>(tmp368);
if (bool(err)) {
return {{},false};
}
{auto&& tmp369 = all;
for(int64_t tmp370=0;tmp370<len(tmp369);++tmp370){
auto e=tmp369[tmp370];if ((e.Name == name)) {
return {e.Data,true};
}
}}
return {{},false};
}
}
// tools/platform/ps2/romdir.go:127:1
std::tuple<Slice<std::string>,Error> ps2_IOPBootConf(Slice<uint8_t> bios){
{
auto tmp371 = ps2_ROMDIREntry(bios,std::string("IOPBTCONF",9));
Slice<uint8_t> blob = std::get<0>(tmp371);
bool ok = std::get<1>(tmp371);
if ((!ok)) {
return {{},go_errors_New(std::string("ps2: this ROM has no IOPBTCONF entry",36))};
}
Slice<std::string> out={};
{auto&& tmp372 = go_strings_Fields(cast<std::string>(blob));
for(int64_t tmp373=0;tmp373<len(tmp372);++tmp373){
auto line=tmp372[tmp373];if (((line == std::string("",0)) || go_strings_HasPrefix(line,std::string("@",1)))) {
continue;
}
out = append(out,Slice<std::string>{line});
}}
return {out,{}};
}
}
// tools/platform/ps2/romdir.go:145:1
std::tuple<Slice<ps2_RomEntry>,Error> ps2_ROMDIRModules(Slice<uint8_t> raw){
{
auto tmp374 = ps2_ReadROMDIR(raw);
Slice<ps2_RomEntry> all = std::get<0>(tmp374);
Error err = std::get<1>(tmp374);
if (bool(err)) {
return {{},err};
}
Slice<ps2_RomEntry> out={};
{auto&& tmp375 = all;
for(int64_t tmp376=0;tmp376<len(tmp375);++tmp376){
auto e=tmp375[tmp376];if (((len(e.Data) >= cast<int64_t>(4ULL)) && (cast<std::string>(sub(e.Data,0,cast<int64_t>(4ULL))) == std::string("\177ELF",4)))) {
out = append(out,Slice<ps2_RomEntry>{e});
}
}}
return {out,{}};
}
}
// tools/platform/ps2/run.go:32:1
std::string ps2_Result_String(ps2_Result r){
{
return go_fmt_Sprintf(std::string("stopped after %d steps at 0x%08X: %s",36),r.Steps,r.PC,r.Reason);
}
}
// tools/platform/ps2/run.go:53:1
ps2_Result ps2_Machine_Run(ps2_Machine* m,uint64_t maxSteps){
{rrprof::Scope timing(0,"EE / IOP and scheduler");
{
ps2_Machine_profRunEnter(m);
auto tmp377=defer([&](){ps2_Machine_profRunExit(m);});
uint64_t steps={};
uint64_t&vblAcc=m->rrVblAcc;
uint64_t&iopAcc=m->rrIopAcc;
RRSpinSet spinSeen{};
int64_t spinAcc={};
std::array<std::array<uint32_t,2>,24> trail={};
int64_t trailPos={};
uint32_t trailPrev={};
{;for (;(steps < maxSteps);){
if ((m->Halted || m->CPU->Halted)) {
break;
}
if (m->StopRequested) {
m->StopRequested = false;
return ps2_Machine_result(m,steps,std::string("stop requested",14));
}
if (((m->iopRebootImage != std::string("",0)) && (m->steps >= m->iopRebootAt))) {
std::string image = m->iopRebootImage;
m->iopRebootImage = std::string("",0);
{
Error err = ps2_Machine_RebootIOPFrom(m,image);
if (bool(err)) {
ps2_Machine_note(m,std::string("SIF: the IOP did not reboot: %v",31),err);
}
}
}
if (((((!m->noIdleSkip) && (!m->idle)) && (!m->OnStep)) && ps2_Machine_idleStep(m))) {
{
uint64_t n = ps2_Machine_idleSkip(m,(&vblAcc),(&iopAcc),cast<uint64_t>((maxSteps - steps)));
if ((n > cast<uint64_t>(0ULL))) {
uint32_t pc = cast<uint32_t>(m->CPU->PC);
steps += n;
spinSeen.add(pc);
spinAcc += cast<int64_t>(n);
if ((spinAcc >= cast<int64_t>(2097152ULL))) {
if ((len(spinSeen) <= cast<int64_t>(6ULL))) {
return ps2_Machine_result(m,steps,go_fmt_Sprintf(std::string("spinning on %d addresses around %s \342\200\224 waiting for something that never happens",79),len(spinSeen),ps2_Machine_Sym(m,pc)));
}
spinSeen = RRSpinSet{};
spinAcc = cast<int64_t>(0ULL);
}
continue;
}
}
}
if (bool(m->IOP)) {
iopAcc++;
if ((iopAcc >= cast<uint64_t>(8ULL))) {
iopAcc = cast<uint64_t>(0ULL);
ps2_IOP_Step(m->IOP);
}
}
vblAcc++;
if ((vblAcc >= cast<uint64_t>(1000000ULL))) {
vblAcc = cast<uint64_t>(0ULL);
ps2_Machine_deliverVBlank(m);
if (m->Halted) {
break;
}
}
else if ((vblAcc == cast<uint64_t>(500000ULL))) {
m->intcStat |= cast<uint32_t>(8ULL);
}
if (m->idle) {
steps++;
m->steps++;
if ((!ps2_Machine_resume(m))) {
if (ps2_Machine_blocked(m)) {
return ps2_Machine_result(m,steps,std::string("deadlocked: every thread is blocked, and nothing left can wake one",66));
}
continue;
}
}
uint32_t pc = cast<uint32_t>(m->CPU->PC);
if ((pc == cast<uint32_t>(252706816ULL))) {
ps2_Machine_onThreadExit(m);
continue;
}
if (get(m->breakpoints,pc)) {
return ps2_Machine_result(m,steps,(std::string("breakpoint at ",14) + ps2_Machine_Sym(m,pc)));
}
{
auto tmp378 = ps2_Machine_unhandledException(m,pc);
std::string r = std::get<0>(tmp378);
bool caught = std::get<1>(tmp378);
if (caught) {
ps2_Machine_note(m,std::string("jump trail before the exception:%s",34),ps2_Machine_formatTrail(m,sub(trail,0,len(trail)),trailPos));
return ps2_Machine_result(m,steps,r);
}
}
if ((!ps2_Machine_mapped(m,ps2_phys(pc)))) {
ps2_Machine_note(m,std::string("jump trail before the PC left memory:%s",39),ps2_Machine_formatTrail(m,sub(trail,0,len(trail)),trailPos));
return ps2_Machine_result(m,steps,go_fmt_Sprintf(std::string("the PC left mapped memory (0x%08X) \342\200\224 an unimplemented kernel call, or a jump through an unwritten pointer",107),pc));
}
if (((trailPrev != cast<uint32_t>(0ULL)) && (pc != cast<uint32_t>((trailPrev + cast<uint32_t>(4ULL)))))) {
trail[trailPos] = std::array<uint32_t,2>{trailPrev,pc};
trailPos = modi<int64_t>((cast<int64_t>((trailPos + cast<int64_t>(1ULL)))),cast<int64_t>(24ULL));
}
trailPrev = pc;
if (bool(m->OnStep)) {
m->OnStep(m,pc);
if (m->Halted) {
break;
}
}
spinSeen.add(pc);
spinAcc++;
if ((spinAcc >= cast<int64_t>(2097152ULL))) {
if ((len(spinSeen) <= cast<int64_t>(6ULL))) {
return ps2_Machine_result(m,steps,go_fmt_Sprintf(std::string("spinning on %d addresses around %s \342\200\224 waiting for something that never happens",79),len(spinSeen),ps2_Machine_Sym(m,pc)));
}
spinSeen = RRSpinSet{};
spinAcc = cast<int64_t>(0ULL);
}
r5900_CPU_Step(m->CPU);
steps++;
m->steps++;
if ((len(m->dmacIRQPending) > cast<int64_t>(0ULL))) {
ps2_Machine_deliverDmacIRQs(m);
}
}
}if (m->CPU->Halted) {
return ps2_Machine_result(m,steps,(std::string("cpu: ",5) + m->CPU->HaltReason));
}
if (m->Halted) {
return ps2_Machine_result(m,steps,m->HaltReason);
}
return ps2_Machine_result(m,steps,std::string("step budget exhausted",21));
}
}
}
// tools/platform/ps2/run.go:253:1
ps2_Result ps2_Machine_result(ps2_Machine* m,uint64_t steps,std::string reason){
{
ps2_Machine_drainVIF1(m);
return ps2_Result{steps,cast<uint32_t>(m->CPU->PC),reason};
}
}
// tools/platform/ps2/run.go:261:1
std::string ps2_Machine_formatTrail(ps2_Machine* m,Slice<std::array<uint32_t,2>> trail,int64_t pos){
{
std::string s = std::string("",0);
{int64_t i = cast<int64_t>(0ULL);for (;(i < len(trail));i++){
std::array<uint32_t,2> e = trail[modi<int64_t>((cast<int64_t>((pos + i))),len(trail))];
if (((e[cast<int64_t>(0ULL)] == cast<uint32_t>(0ULL)) && (e[cast<int64_t>(1ULL)] == cast<uint32_t>(0ULL)))) {
continue;
}
s += go_fmt_Sprintf(std::string("\012  %-34s -> %s",14),ps2_Machine_Sym(m,e[cast<int64_t>(0ULL)]),ps2_Machine_Sym(m,e[cast<int64_t>(1ULL)]));
}
}return s;
}
}
// tools/platform/ps2/run.go:286:1
std::tuple<std::string,bool> ps2_Machine_unhandledException(ps2_Machine* m,uint32_t pc){
{if(pc!=0x80000000u && pc!=0x80000080u && pc!=0x80000180u)return {std::string{},false};
{
auto tmp379 = lookup(ps2_exceptionVectors,pc);
std::string name = std::get<0>(tmp379);
bool isVector = std::get<1>(tmp379);
if ((!isVector)) {
return {std::string("",0),false};
}
if ((ps2_Machine_Fetch32(m,pc) != cast<uint32_t>(0ULL))) {
return {std::string("",0),false};
}
uint64_t cause = m->CPU->COP0[cast<int64_t>(13ULL)];
uint64_t code = cast<uint64_t>(((shr<uint64_t>(cause,cast<int64_t>(2ULL))) & cast<uint64_t>(31ULL)));
uint32_t epc = cast<uint32_t>(m->CPU->COP0[cast<int64_t>(14ULL)]);
uint32_t bad = cast<uint32_t>(m->CPU->COP0[cast<int64_t>(8ULL)]);
return {go_fmt_Sprintf(std::string("unhandled %s exception (%s) at %s \342\200\224 faulting address 0x%08X, and nothing is installed at the vector",101),name,ps2_excName(cast<uint32_t>(code)),ps2_Machine_Sym(m,epc),bad),true};
}
}
}
// tools/platform/ps2/run.go:304:1
std::string ps2_excName(uint32_t code){
{
{
switch(code){
case cast<uint32_t>(0ULL):{
return std::string("interrupt",9);
break;}
case cast<uint32_t>(1ULL):{
return std::string("TLB modification",16);
break;}
case cast<uint32_t>(2ULL):{
return std::string("TLB miss on a load or fetch",27);
break;}
case cast<uint32_t>(3ULL):{
return std::string("TLB miss on a store",19);
break;}
case cast<uint32_t>(4ULL):{
return std::string("address error on a load or fetch",32);
break;}
case cast<uint32_t>(5ULL):{
return std::string("address error on a store",24);
break;}
case cast<uint32_t>(8ULL):{
return std::string("syscall",7);
break;}
case cast<uint32_t>(9ULL):{
return std::string("breakpoint",10);
break;}
case cast<uint32_t>(10ULL):{
return std::string("reserved instruction",20);
break;}
case cast<uint32_t>(11ULL):{
return std::string("coprocessor unusable",20);
break;}
case cast<uint32_t>(12ULL):{
return std::string("arithmetic overflow",19);
break;}
case cast<uint32_t>(13ULL):{
return std::string("trap",4);
break;}
}}
return go_fmt_Sprintf(std::string("code %d",7),code);
}
}
// tools/platform/ps2/sched.go:27:1
std::string ps2_threadState_String(ps2_threadState s){
{
{
switch(s){
case cast<ps2_threadState>(0ULL):{
return std::string("running",7);
break;}
case cast<ps2_threadState>(1ULL):{
return std::string("ready",5);
break;}
case cast<ps2_threadState>(2ULL):{
return std::string("sleeping",8);
break;}
case cast<ps2_threadState>(3ULL):{
return std::string("wait-sema",9);
break;}
case cast<ps2_threadState>(4ULL):{
return std::string("dormant",7);
break;}
}}
return std::string("dead",4);
}
}
// tools/platform/ps2/sched.go:76:1
void ps2_Machine_createThread(ps2_Machine* m){
{
uint32_t p = ps2_Machine_arg(m,cast<uint32_t>(0ULL));
ps2_thread* t = arenaNew(ps2_thread{m->nextThreadID,ps2_Machine_Read32(m,cast<uint32_t>((p + cast<uint32_t>(4ULL)))),ps2_Machine_Read32(m,cast<uint32_t>((p + cast<uint32_t>(8ULL)))),ps2_Machine_Read32(m,cast<uint32_t>((p + cast<uint32_t>(12ULL)))),ps2_Machine_Read32(m,cast<uint32_t>((p + cast<uint32_t>(16ULL)))),ps2_Machine_Read32(m,cast<uint32_t>((p + cast<uint32_t>(20ULL)))),cast<ps2_threadState>(4ULL),{},{}});
m->nextThreadID++;
m->threads[t->id] = t;
ps2_Machine_note(m,std::string("CreateThread %d: entry=%s stack=0x%08X+0x%X prio=%d",51),t->id,ps2_Machine_Sym(m,t->entry),t->stack,t->stackSz,t->priority);
ps2_Machine_setRet(m,t->id);
}
}
// tools/platform/ps2/sched.go:97:1
void ps2_Machine_startThread(ps2_Machine* m){
{
auto tmp380 = std::make_tuple(ps2_Machine_arg(m,cast<uint32_t>(0ULL)),ps2_Machine_arg(m,cast<uint32_t>(1ULL)));
uint32_t id = std::get<0>(tmp380);
uint32_t arg = std::get<1>(tmp380);
ps2_thread* t = get(m->threads,id);
if ((!t)) {
ps2_Machine_note(m,std::string("StartThread %d: no such thread",30),id);
ps2_Machine_setRet(m,cast<uint32_t>(4294967295ULL));
return ;
}
t->ctx = r5900_CPU_Snapshot(m->CPU);
t->ctx.R = std::array<r5900_Quad,32>{};
auto tmp381 = std::make_tuple(cast<uint64_t>(0ULL),cast<uint64_t>(0ULL),cast<uint64_t>(0ULL),cast<uint64_t>(0ULL),cast<uint32_t>(0ULL));
t->ctx.HI = std::get<0>(tmp381);
t->ctx.LO = std::get<1>(tmp381);
t->ctx.HI1 = std::get<2>(tmp381);
t->ctx.LO1 = std::get<3>(tmp381);
t->ctx.SA = std::get<4>(tmp381);
auto tmp382 = std::make_tuple(false,false);
t->ctx.DelaySlot = std::get<0>(tmp382);
t->ctx.PendingDelay = std::get<1>(tmp382);
t->ctx.PC = cast<uint64_t>(t->entry);
t->ctx.NextPC = cast<uint64_t>((cast<uint64_t>(t->entry) + cast<uint64_t>(4ULL)));
t->ctx.R[cast<int64_t>(4ULL)] = r5900_Quad{cast<uint64_t>(arg),{}};
t->ctx.R[cast<int64_t>(28ULL)] = r5900_Quad{cast<uint64_t>(t->gp),{}};
t->ctx.R[cast<int64_t>(29ULL)] = r5900_Quad{cast<uint64_t>(cast<uint32_t>((cast<uint32_t>((t->stack + t->stackSz)) - cast<uint32_t>(16ULL)))),{}};
t->ctx.R[cast<int64_t>(31ULL)] = r5900_Quad{cast<uint64_t>(252706816ULL),{}};
t->state = cast<ps2_threadState>(1ULL);
ps2_Machine_setRet(m,id);
ps2_thread* cur = get(m->threads,m->currentThread);
if ((bool(cur) && (t->priority < cur->priority))) {
ps2_Machine_switchTo(m,t,cast<ps2_threadState>(1ULL));
}
}
}
// tools/platform/ps2/sched.go:139:1
void ps2_Machine_sleepThread(ps2_Machine* m){
{
ps2_thread* t = get(m->threads,m->currentThread);
if ((!t)) {
ps2_Machine_setRet(m,cast<uint32_t>(0ULL));
return ;
}
if ((t->wakeupCount > cast<int32_t>(0ULL))) {
t->wakeupCount--;
ps2_Machine_setRet(m,cast<uint32_t>(0ULL));
return ;
}
ps2_Machine_setRet(m,cast<uint32_t>(0ULL));
ps2_Machine_switchAway(m,cast<ps2_threadState>(2ULL));
}
}
// tools/platform/ps2/sched.go:155:1
void ps2_Machine_wakeupThread(ps2_Machine* m,uint32_t id){
{
ps2_thread* t = get(m->threads,id);
if ((!t)) {
ps2_Machine_setRet(m,cast<uint32_t>(4294967295ULL));
return ;
}
m->eeDisturbGen++;
if ((t->state == cast<ps2_threadState>(2ULL))) {
t->state = cast<ps2_threadState>(1ULL);
}
else {
t->wakeupCount++;
}
ps2_Machine_setRet(m,id);
}
}
// tools/platform/ps2/sched.go:180:1
void ps2_Machine_switchAway(ps2_Machine* m,ps2_threadState newState){
{
{
ps2_thread* cur = get(m->threads,m->currentThread);
if (bool(cur)) {
cur->ctx = r5900_CPU_Snapshot(m->CPU);
cur->state = newState;
}
}
ps2_thread* next = ps2_Machine_pickReady(m);
if ((!next)) {
m->idle = true;
return ;
}
next->state = cast<ps2_threadState>(0ULL);
m->currentThread = next->id;
r5900_CPU_Restore(m->CPU,next->ctx);
}
}
// tools/platform/ps2/sched.go:204:1
void ps2_Machine_preemptIfOutranked(ps2_Machine* m){
{
ps2_thread* cur = get(m->threads,m->currentThread);
if (((m->idle || (!cur)) || (cur->state != cast<ps2_threadState>(0ULL)))) {
return ;
}
{
ps2_thread* next = ps2_Machine_pickReady(m);
if ((bool(next) && (next->priority < cur->priority))) {
ps2_Machine_switchTo(m,next,cast<ps2_threadState>(1ULL));
}
}
}
}
// tools/platform/ps2/sched.go:216:1
bool ps2_Machine_resume(ps2_Machine* m){
{
ps2_thread* next = ps2_Machine_pickReady(m);
if ((!next)) {
return false;
}
m->idle = false;
next->state = cast<ps2_threadState>(0ULL);
m->currentThread = next->id;
r5900_CPU_Restore(m->CPU,next->ctx);
return true;
}
}
// tools/platform/ps2/sched.go:231:1
bool ps2_Machine_blocked(ps2_Machine* m){
{
if ((bool(m->IOP) || (len(m->intcHandlers) > cast<int64_t>(0ULL)))) {
return false;
}
if ((((bool(m->IOP) && m->IOP->running) && (!m->IOP->Halted)) && (!m->IOP->CPU->Halted))) {
return false;
}
return (!ps2_Machine_pickReady(m));
}
}
// tools/platform/ps2/sched.go:244:1
void ps2_Machine_switchTo(ps2_Machine* m,ps2_thread* next,ps2_threadState curState){
{
m->eeDisturbGen++;
{
ps2_thread* cur = get(m->threads,m->currentThread);
if (bool(cur)) {
cur->ctx = r5900_CPU_Snapshot(m->CPU);
if ((cur->state == cast<ps2_threadState>(0ULL))) {
cur->state = curState;
}
}
}
next->state = cast<ps2_threadState>(0ULL);
m->currentThread = next->id;
m->idle = false;
r5900_CPU_Restore(m->CPU,next->ctx);
}
}
// tools/platform/ps2/sched.go:259:1
ps2_thread* ps2_Machine_pickReady(ps2_Machine* m){
{
ps2_thread* best={};
{auto&& tmp383 = m->threads;
for(auto [tmp384,tmp385]:tmp383){
auto t=tmp385;if ((t->state != cast<ps2_threadState>(1ULL))) {
continue;
}
if ((((!best) || (t->priority < best->priority)) || (((t->priority == best->priority) && (t->id < best->id))))) {
best = t;
}
}}
return best;
}
}
// tools/platform/ps2/sched.go:274:1
void ps2_Machine_onThreadExit(ps2_Machine* m){
{
{
ps2_thread* t = get(m->threads,m->currentThread);
if (bool(t)) {
t->state = cast<ps2_threadState>(5ULL);
ps2_Machine_note(m,std::string("thread %d exited",16),t->id);
}
}
if ((!ps2_Machine_resume(m))) {
m->idle = true;
}
}
}
// tools/platform/ps2/sched.go:286:1
std::string ps2_Machine_Threads(ps2_Machine* m){
{
if ((len(m->threads) == cast<int64_t>(0ULL))) {
return std::string("no threads\012",11);
}
std::string s = std::string("threads:\012",9);
{uint32_t id = cast<uint32_t>(1ULL);for (;(id < m->nextThreadID);id++){
ps2_thread* t = get(m->threads,id);
if ((!t)) {
continue;
}
std::string mark = std::string(" ",1);
if ((id == m->currentThread)) {
mark = std::string(">",1);
}
uint32_t pc = cast<uint32_t>(t->ctx.PC);
if ((id == m->currentThread)) {
pc = cast<uint32_t>(m->CPU->PC);
}
s += ps2_sprintf(std::string("%s %2d  %-9s prio=%-3d pc=%-32s entry=%s\012",41),mark,t->id,t->state,t->priority,ps2_Machine_Sym(m,pc),ps2_Machine_Sym(m,t->entry));
}
}return s;
}
}
// tools/platform/ps2/sched.go:312:1
void ps2_Machine_exitThread(ps2_Machine* m){
{
ps2_Machine_onThreadExit(m);
}
}
// tools/platform/ps2/sched.go:318:1
void ps2_Machine_referThreadStatus(ps2_Machine* m){
{
auto tmp386 = std::make_tuple(ps2_Machine_arg(m,cast<uint32_t>(0ULL)),ps2_Machine_arg(m,cast<uint32_t>(1ULL)));
uint32_t id = std::get<0>(tmp386);
uint32_t p = std::get<1>(tmp386);
if ((id == cast<uint32_t>(0ULL))) {
id = m->currentThread;
}
ps2_thread* t = get(m->threads,id);
if ((!t)) {
ps2_Machine_setRet(m,cast<uint32_t>(4294967295ULL));
return ;
}
if ((p != cast<uint32_t>(0ULL))) {
ps2_Machine_Write32(m,cast<uint32_t>((p + cast<uint32_t>(0ULL))),cast<uint32_t>(ps2_eeThreadStatus(t->state)));
ps2_Machine_Write32(m,cast<uint32_t>((p + cast<uint32_t>(4ULL))),t->entry);
ps2_Machine_Write32(m,cast<uint32_t>((p + cast<uint32_t>(8ULL))),t->stack);
ps2_Machine_Write32(m,cast<uint32_t>((p + cast<uint32_t>(12ULL))),t->gp);
ps2_Machine_Write32(m,cast<uint32_t>((p + cast<uint32_t>(16ULL))),t->priority);
ps2_Machine_Write32(m,cast<uint32_t>((p + cast<uint32_t>(20ULL))),t->priority);
}
ps2_Machine_setRet(m,cast<uint32_t>(ps2_eeThreadStatus(t->state)));
}
}
// tools/platform/ps2/sched.go:341:1
int64_t ps2_eeThreadStatus(ps2_threadState s){
{
{
switch(s){
case cast<ps2_threadState>(0ULL):{
return cast<int64_t>(1ULL);
break;}
case cast<ps2_threadState>(1ULL):{
return cast<int64_t>(2ULL);
break;}
case cast<ps2_threadState>(3ULL):case cast<ps2_threadState>(2ULL):{
return cast<int64_t>(4ULL);
break;}
case cast<ps2_threadState>(4ULL):{
return cast<int64_t>(16ULL);
break;}
}}
return cast<int64_t>(0ULL);
}
}
// tools/platform/ps2/sema.go:32:1
void ps2_Machine_createSema(ps2_Machine* m){
{
uint32_t p = ps2_Machine_arg(m,cast<uint32_t>(0ULL));
ps2_sema* s = arenaNew(ps2_sema{m->nextSemaID,cast<int32_t>(ps2_Machine_Read32(m,cast<uint32_t>((p + cast<uint32_t>(8ULL))))),cast<int32_t>(ps2_Machine_Read32(m,cast<uint32_t>((p + cast<uint32_t>(4ULL))))),{}});
m->nextSemaID++;
m->semas[s->id] = s;
ps2_Machine_setRet(m,s->id);
}
}
// tools/platform/ps2/sema.go:45:1
void ps2_Machine_waitSema(ps2_Machine* m,uint32_t id){
{
ps2_sema* s = get(m->semas,id);
if ((!s)) {
ps2_Machine_setRet(m,cast<uint32_t>(4294967295ULL));
return ;
}
ps2_Machine_setRet(m,id);
if ((s->count > cast<int32_t>(0ULL))) {
s->count--;
return ;
}
s->waiting = append(s->waiting,Slice<uint32_t>{m->currentThread});
ps2_Machine_switchAway(m,cast<ps2_threadState>(3ULL));
}
}
// tools/platform/ps2/sema.go:63:1
void ps2_Machine_signalSema(ps2_Machine* m,uint32_t id){
{
ps2_sema* s = get(m->semas,id);
if ((!s)) {
ps2_Machine_setRet(m,cast<uint32_t>(4294967295ULL));
return ;
}
ps2_Machine_setRet(m,id);
if ((len(s->waiting) > cast<int64_t>(0ULL))) {
uint32_t tid = s->waiting[cast<int64_t>(0ULL)];
s->waiting = sub(s->waiting,cast<int64_t>(1ULL),len(s->waiting));
{
ps2_thread* t = get(m->threads,tid);
if ((bool(t) && (t->state == cast<ps2_threadState>(3ULL)))) {
t->state = cast<ps2_threadState>(1ULL);
m->eeDisturbGen++;
}
}
return ;
}
if (((s->count < s->maxCount) || (s->maxCount == cast<int32_t>(0ULL)))) {
s->count++;
}
}
}
// tools/platform/ps2/sema.go:87:1
void ps2_Machine_pollSema(ps2_Machine* m,uint32_t id){
{
ps2_sema* s = get(m->semas,id);
if (((!s) || (s->count <= cast<int32_t>(0ULL)))) {
ps2_Machine_setRet(m,cast<uint32_t>(4294967295ULL));
return ;
}
s->count--;
ps2_Machine_setRet(m,id);
}
}
// tools/platform/ps2/sif.go:88:1
void ps2_Machine_sifSetDma(ps2_Machine* m){
{
auto tmp387 = std::make_tuple(ps2_Machine_arg(m,cast<uint32_t>(0ULL)),ps2_Machine_arg(m,cast<uint32_t>(1ULL)));
uint32_t desc = std::get<0>(tmp387);
uint32_t count = std::get<1>(tmp387);
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < count);i++){
uint32_t d = cast<uint32_t>((desc + cast<uint32_t>((i * cast<uint32_t>(16ULL)))));
uint32_t src = ps2_Machine_Read32(m,cast<uint32_t>((d + cast<uint32_t>(0ULL))));
uint32_t dest = ps2_Machine_Read32(m,cast<uint32_t>((d + cast<uint32_t>(4ULL))));
uint32_t size = ps2_Machine_Read32(m,cast<uint32_t>((d + cast<uint32_t>(8ULL))));
uint32_t attr = ps2_Machine_Read32(m,cast<uint32_t>((d + cast<uint32_t>(12ULL))));
if ((cast<uint32_t>((attr & cast<uint32_t>(64ULL))) != cast<uint32_t>(0ULL))) {
ps2_Machine_iopReceive(m,src,dest,size);
continue;
}
ps2_Machine_sifData(m,src,dest,size);
}
}m->sifDmaID++;
ps2_Machine_setRet(m,m->sifDmaID);
}
}
// tools/platform/ps2/sif.go:116:1
void ps2_Machine_sifData(ps2_Machine* m,uint32_t src,uint32_t dest,uint32_t size){
{
if (((dest == cast<uint32_t>(0ULL)) || (size == cast<uint32_t>(0ULL)))) {
return ;
}
ps2_sifPacket p = ps2_sifPacket{dest,Slice<uint8_t>::make(size),{}};
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < size);i++){
p.data[i] = ps2_Machine_Read(m,cast<uint32_t>((src + i)));
}
}m->sifToIOPQueue = append(m->sifToIOPQueue,Slice<ps2_sifPacket>{p});
ps2_Machine_sifPump(m);
}
}
// tools/platform/ps2/sif.go:129:1
void ps2_Machine_iopReceive(ps2_Machine* m,uint32_t src,uint32_t dest,uint32_t size){
{
if ((size < cast<uint32_t>(16ULL))) {
return ;
}
uint32_t cid = ps2_Machine_Read32(m,cast<uint32_t>((src + cast<uint32_t>(8ULL))));
uint32_t opt = ps2_Machine_Read32(m,cast<uint32_t>((src + cast<uint32_t>(12ULL))));
if ((cid == cast<uint32_t>(2147483651ULL))) {
ps2_Machine_iopReset(m,src);
return ;
}
if (((cid == cast<uint32_t>(2147483648ULL)) || (((cid == cast<uint32_t>(2147483650ULL)) && (opt == cast<uint32_t>(0ULL)))))) {
m->sifCmdBuf = ps2_Machine_Read32(m,cast<uint32_t>((src + cast<uint32_t>(16ULL))));
}
m->sifSent[cid]++;
ps2_Machine_sifWatch(m,src,cid,true);
ps2_Machine_sifToIOP(m,src,dest,size);
}
}
// tools/platform/ps2/sif.go:193:1
void ps2_Machine_iopReset(ps2_Machine* m,uint32_t src){
{
std::string cmd = ps2_Machine_CString(m,cast<uint32_t>((src + cast<uint32_t>(24ULL))));
auto tmp388 = ps2_iopRebootImage(cmd);
std::string image = std::get<0>(tmp388);
Error err = std::get<1>(tmp388);
if (bool(err)) {
ps2_Machine_note(m,std::string("SIF: %v",7),err);
return ;
}
ps2_Machine_note(m,std::string("SIF: the EE is rebooting the IOP \342\200\224 %q, so the image is %s",59),cmd,image);
m->iopRebootImage = image;
m->iopRebootAt = cast<uint64_t>((m->steps + cast<uint64_t>(100000ULL)));
}
}
// tools/platform/ps2/sif.go:282:1
void ps2_Machine_sifToIOP(ps2_Machine* m,uint32_t src,uint32_t dest,uint32_t size){
{
if ((dest == cast<uint32_t>(0ULL))) {
ps2_Machine_note(m,std::string("SIF: the EE sent the IOP a packet before it knew where the IOP wanted it (dest is 0)",84));
return ;
}
ps2_sifPacket p = ps2_sifPacket{dest,Slice<uint8_t>::make(size),true};
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < size);i++){
p.data[i] = ps2_Machine_Read(m,cast<uint32_t>((src + i)));
}
}m->sifToIOPQueue = append(m->sifToIOPQueue,Slice<ps2_sifPacket>{p});
ps2_Machine_sifPump(m);
}
}
// tools/platform/ps2/sif.go:305:1
void ps2_Machine_sifPump(ps2_Machine* m){
{
if (((!m->IOP) || (len(m->sifToIOPQueue) == cast<int64_t>(0ULL)))) {
return ;
}
ps2_iopDMAChan* c = (&m->IOP->dma[cast<int64_t>(10ULL)]);
{;for (;(len(m->sifToIOPQueue) > cast<int64_t>(0ULL));){
ps2_sifPacket p = m->sifToIOPQueue[cast<int64_t>(0ULL)];
if ((p.cmd && (cast<uint32_t>((c->chcr & cast<uint32_t>(16777216ULL))) == cast<uint32_t>(0ULL)))) {
return ;
}
m->sifToIOPQueue = sub(m->sifToIOPQueue,cast<int64_t>(1ULL),len(m->sifToIOPQueue));
{auto&& tmp389 = p.data;
for(int64_t tmp390=0;tmp390<len(tmp389);++tmp390){
auto i=tmp390;auto b=tmp389[tmp390];ps2_IOP_Write(m->IOP,cast<uint32_t>((p.dest + cast<uint32_t>(i))),b);
}}
m->sifToIOPCount++;
if ((!p.cmd)) {
continue;
}
c->madr = cast<uint32_t>((p.dest + cast<uint32_t>(len(p.data))));
c->chcr &= ~(cast<uint32_t>(285212672ULL));
ps2_IOP_raiseIRQ(m->IOP,ps2_iopDMAIRQ(cast<int64_t>(10ULL)));
}
}}
}
// tools/platform/ps2/sif.go:367:1
void ps2_Machine_sifFromIOP(ps2_Machine* m){
{
m->eeDisturbGen++;
ps2_iopDMAChan* c = (&m->IOP->dma[cast<int64_t>(9ULL)]);
{uint32_t tag = c->tadr;for (;;tag += cast<uint32_t>(16ULL)){
uint32_t src = cast<uint32_t>((ps2_IOP_Read32(m->IOP,cast<uint32_t>((tag + cast<uint32_t>(0ULL)))) & cast<uint32_t>(16777215ULL)));
uint32_t words = ps2_IOP_Read32(m->IOP,cast<uint32_t>((tag + cast<uint32_t>(4ULL))));
uint32_t eeTag = ps2_IOP_Read32(m->IOP,cast<uint32_t>((tag + cast<uint32_t>(8ULL))));
uint32_t dest = ps2_IOP_Read32(m->IOP,cast<uint32_t>((tag + cast<uint32_t>(12ULL))));
if (((words == cast<uint32_t>(0ULL)) || (dest == cast<uint32_t>(0ULL)))) {
break;
}
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < cast<uint32_t>((words * cast<uint32_t>(4ULL))));i++){
ps2_Machine_Write(m,cast<uint32_t>((dest + i)),ps2_IOP_Read(m->IOP,cast<uint32_t>((src + i))));
}
}m->sifFromIOPCount++;
if ((dest == m->sifCmdBuf)) {
m->sifBack[ps2_Machine_Read32(m,cast<uint32_t>((dest + cast<uint32_t>(8ULL))))]++;
if (ps2_sifLog) {
ps2_sprintfLog(std::string("SIF vbl=%d  <- IOP cid=%08X (%s) to cmdbuf",42),m->vblanks,ps2_Machine_Read32(m,cast<uint32_t>((dest + cast<uint32_t>(8ULL)))),ps2_sifCmdName(ps2_Machine_Read32(m,cast<uint32_t>((dest + cast<uint32_t>(8ULL))))));
}
}
else if (ps2_sifLog) {
ps2_sprintfLog(std::string("SIF vbl=%d  <- IOP reply -> EE %08X (%d words)",46),m->vblanks,dest,words);
}
if (((cast<uint32_t>((eeTag & cast<uint32_t>(2147483648ULL))) != cast<uint32_t>(0ULL)) && (m->sifCmdHandler != cast<uint32_t>(0ULL)))) {
ps2_Machine_callGuest(m,m->sifCmdHandler,Slice<uint32_t>{cast<uint32_t>(0ULL)});
ps2_Machine_preemptIfOutranked(m);
}
if ((cast<uint32_t>((ps2_IOP_Read32(m->IOP,cast<uint32_t>((tag + cast<uint32_t>(0ULL)))) & cast<uint32_t>(2147483648ULL))) != cast<uint32_t>(0ULL))) {
break;
}
}
}}
}
// tools/platform/ps2/sif.go:462:1
void ps2_Machine_sifWatch(ps2_Machine* m,uint32_t src,uint32_t cid,bool toIOP){
{
{
if ((toIOP && (cid == cast<uint32_t>(2147483657ULL)))){
m->rpcBinds[ps2_Machine_Read32(m,cast<uint32_t>((src + cast<uint32_t>(32ULL))))]++;
}
else if ((toIOP && (cid == cast<uint32_t>(2147483658ULL)))){
m->rpcCalls[ps2_sifRPCKey{ps2_Machine_Read32(m,cast<uint32_t>((src + cast<uint32_t>(52ULL)))),ps2_Machine_Read32(m,cast<uint32_t>((src + cast<uint32_t>(32ULL))))}]++;
if (ps2_sifLog) {
uint32_t sp = cast<uint32_t>(r5900_CPU_Reg(m->CPU,cast<uint32_t>(29ULL)));
std::string callers={};
{uint32_t i = cast<uint32_t>(0ULL);for (;((i < cast<uint32_t>(96ULL)) && (len(callers) < cast<int64_t>(120ULL)));i += cast<uint32_t>(4ULL)){
uint32_t w = ps2_Machine_Read32(m,cast<uint32_t>((sp + i)));
{
std::string s = ps2_Machine_Sym(m,w);
if (((s != std::string("",0)) && (!go_strings_HasPrefix(s,std::string("0x",2))))) {
callers += (std::string(" ",1) + s);
}
}
}
}ps2_sprintfLog(std::string("SIF vbl=%d RPC_CALL handle=%08X fno=%d ra=%s callers:%s",55),m->vblanks,ps2_Machine_Read32(m,cast<uint32_t>((src + cast<uint32_t>(52ULL)))),ps2_Machine_Read32(m,cast<uint32_t>((src + cast<uint32_t>(32ULL)))),ps2_Machine_Sym(m,cast<uint32_t>(r5900_CPU_Reg(m->CPU,cast<uint32_t>(31ULL)))),callers);
}
}
}
tmp391:;
}
}
// tools/platform/ps2/sif.go:544:1
void ps2_Machine_sifSetReg(ps2_Machine* m){
{
auto tmp392 = std::make_tuple(ps2_Machine_arg(m,cast<uint32_t>(0ULL)),ps2_Machine_arg(m,cast<uint32_t>(1ULL)));
uint32_t reg = std::get<0>(tmp392);
uint32_t val = std::get<1>(tmp392);
{
switch(reg){
case cast<uint32_t>(4ULL):{
ps2_Machine_sbusFlagClear(m,cast<uint32_t>(48ULL),val);
break;}
case cast<uint32_t>(2ULL):{
ps2_Machine_sbusWrite(m,cast<uint32_t>(16ULL),val);
break;}
case cast<uint32_t>(2147483648ULL):case cast<uint32_t>(2147483649ULL):case cast<uint32_t>(2147483650ULL):{
m->sifRegs[reg] = val;
break;}
default:{
m->sifUnmodelledReg[reg]++;
break;}
}}
ps2_Machine_setRet(m,cast<uint32_t>(0ULL));
}
}
// tools/platform/ps2/sif.go:575:1
void ps2_Machine_sifGetReg(ps2_Machine* m){
{
uint32_t reg = ps2_Machine_arg(m,cast<uint32_t>(0ULL));
{
switch(reg){
case cast<uint32_t>(4ULL):{
ps2_Machine_setRet(m,ps2_Machine_sbusRead(m,cast<uint32_t>(48ULL)));
break;}
case cast<uint32_t>(2ULL):{
ps2_Machine_setRet(m,ps2_Machine_sbusRead(m,cast<uint32_t>(16ULL)));
break;}
case cast<uint32_t>(2147483648ULL):case cast<uint32_t>(2147483649ULL):case cast<uint32_t>(2147483650ULL):{
ps2_Machine_setRet(m,get(m->sifRegs,reg));
break;}
default:{
m->sifUnmodelledReg[reg]++;
ps2_Machine_setRet(m,cast<uint32_t>(0ULL));
break;}
}}
}
}
// tools/platform/ps2/sif.go:599:1
std::string ps2_sifCmdName(uint32_t cid){
{
{
switch(cid){
case cast<uint32_t>(2147483648ULL):{
return std::string("CHANGE_SADDR",12);
break;}
case cast<uint32_t>(2147483649ULL):{
return std::string("SET_SREG",8);
break;}
case cast<uint32_t>(2147483650ULL):{
return std::string("INIT_CMD",8);
break;}
case cast<uint32_t>(2147483651ULL):{
return std::string("RESET",5);
break;}
case cast<uint32_t>(2147483656ULL):{
return std::string("RPC_END",7);
break;}
case cast<uint32_t>(2147483657ULL):{
return std::string("RPC_BIND",8);
break;}
case cast<uint32_t>(2147483658ULL):{
return std::string("RPC_CALL",8);
break;}
case cast<uint32_t>(2147483660ULL):{
return std::string("RPC_RDATA",9);
break;}
}}
return std::string("?",1);
}
}
// tools/platform/ps2/sif.go:623:1
std::string ps2_smflgBits(uint32_t v){
{
Slice<std::string> names={};
{auto&& tmp393 = Slice<Anon83>{Anon83{cast<uint32_t>(65536ULL),std::string("SIF up",6)},Anon83{cast<uint32_t>(131072ULL),std::string("command layer listening",23)},Anon83{cast<uint32_t>(262144ULL),std::string("reboot done",11)}};
for(int64_t tmp394=0;tmp394<len(tmp393);++tmp394){
auto b=tmp393[tmp394];if ((cast<uint32_t>((v & b.bit)) != cast<uint32_t>(0ULL))) {
names = append(names,Slice<std::string>{b.name});
}
}}
if ((len(names) == cast<int64_t>(0ULL))) {
return std::string("  (the IOP has raised nothing)",30);
}
return ((std::string("  (",3) + go_strings_Join(names,std::string(", ",2))) + std::string(")",1));
}
}
// tools/platform/ps2/sif.go:650:1
std::string ps2_Machine_SIFCensus(ps2_Machine* m){
{
std::string s = ps2_sprintf(std::string("the SIF: %d packets crossed to the IOP, %d came back.\012",54),m->sifToIOPCount,m->sifFromIOPCount);
s += ps2_sprintf(std::string("  SMCOM 0x%08X (the IOP's command buffer)   SMFLG 0x%08X%s\012",59),ps2_Machine_sbusRead(m,cast<uint32_t>(16ULL)),ps2_Machine_sbusRead(m,cast<uint32_t>(48ULL)),ps2_smflgBits(ps2_Machine_sbusRead(m,cast<uint32_t>(48ULL))));
s += ps2_sprintf(std::string("  the EE's command buffer 0x%08X, its handler %s\012",49),m->sifCmdBuf,ps2_Machine_Sym(m,m->sifCmdHandler));
{auto&& tmp395 = Slice<Anon84>{Anon84{std::string("->",2),m->sifSent},Anon84{std::string("<-",2),m->sifBack}};
for(int64_t tmp396=0;tmp396<len(tmp395);++tmp396){
auto d=tmp395[tmp396];Slice<uint32_t> cids={};
{auto&& tmp397 = d.m;
for(auto [tmp398,tmp399]:tmp397){
auto cid=tmp398;cids = append(cids,Slice<uint32_t>{cid});
}}
go_sort_Slice(cids,[&](int64_t i,int64_t j)->bool{
return (cids[i] < cids[j]);
}
);
{auto&& tmp400 = cids;
for(int64_t tmp401=0;tmp401<len(tmp400);++tmp401){
auto cid=tmp400[tmp401];s += ps2_sprintf(std::string("  %s  %-13s 0x%08X  %d\012",23),d.way,ps2_sifCmdName(cid),cid,get(d.m,cid));
}}
}}
if ((len(m->rpcBinds) > cast<int64_t>(0ULL))) {
s += std::string("  the servers the EE bound to, on the IOP:\012",43);
Slice<uint32_t> sids={};
{auto&& tmp402 = m->rpcBinds;
for(auto [tmp403,tmp404]:tmp402){
auto sid=tmp403;sids = append(sids,Slice<uint32_t>{sid});
}}
go_sort_Slice(sids,[&](int64_t i,int64_t j)->bool{
return (sids[i] < sids[j]);
}
);
{auto&& tmp405 = sids;
for(int64_t tmp406=0;tmp406<len(tmp405);++tmp406){
auto sid=tmp405[tmp406];s += ps2_sprintf(std::string("      0x%08X  bound %d time%s\012",30),sid,get(m->rpcBinds,sid),ps2_plural(get(m->rpcBinds,sid)));
}}
}
if ((len(m->rpcCalls) > cast<int64_t>(0ULL))) {
s += std::string("  and the functions it called, by the handle the IOP gave it:\012",62);
Slice<ps2_sifRPCKey> keys={};
{auto&& tmp407 = m->rpcCalls;
for(auto [tmp408,tmp409]:tmp407){
auto k=tmp408;keys = append(keys,Slice<ps2_sifRPCKey>{k});
}}
go_sort_Slice(keys,[&](int64_t i,int64_t j)->bool{
if ((keys[i].sid != keys[j].sid)) {
return (keys[i].sid < keys[j].sid);
}
return (keys[i].fno < keys[j].fno);
}
);
{auto&& tmp410 = keys;
for(int64_t tmp411=0;tmp411<len(tmp410);++tmp411){
auto k=tmp410[tmp411];s += ps2_sprintf(std::string("      handle 0x%08X  fn %-4d  %d call%s\012",40),k.sid,k.fno,get(m->rpcCalls,k),ps2_plural(get(m->rpcCalls,k)));
}}
}
{auto&& tmp412 = m->sifUnmodelledReg;
for(auto [tmp413,tmp414]:tmp412){
auto reg=tmp413;auto n=tmp414;s += ps2_sprintf(std::string("  SIF register 0x%08X asked for %d time%s, and nothing models it\012",65),reg,n,ps2_plural(n));
}}
return s;
}
}
// tools/platform/ps2/sifbus.go:85:1
void ps2_Machine_sbusFlagSet(ps2_Machine* m,uint32_t reg,uint32_t bits){
{
m->sbus[divi<uint32_t>(reg,cast<uint32_t>(16ULL))] |= bits;
}
}
// tools/platform/ps2/sifbus.go:86:1
void ps2_Machine_sbusFlagClear(ps2_Machine* m,uint32_t reg,uint32_t bits){
{
m->sbus[divi<uint32_t>(reg,cast<uint32_t>(16ULL))] &= ~(bits);
}
}
// tools/platform/ps2/sifbus.go:89:1
uint32_t ps2_Machine_sbusRead(ps2_Machine* m,uint32_t off){
{
{
uint32_t i = divi<uint32_t>(off,cast<uint32_t>(16ULL));
if ((i < cast<uint32_t>(8ULL))) {
return m->sbus[i];
}
}
return cast<uint32_t>(0ULL);
}
}
// tools/platform/ps2/sifbus.go:99:1
void ps2_Machine_sbusWriteIOP(ps2_Machine* m,uint32_t off,uint32_t v){
{
{
switch(off){
case cast<uint32_t>(48ULL):{
ps2_Machine_sbusFlagSet(m,cast<uint32_t>(48ULL),v);
break;}
case cast<uint32_t>(32ULL):{
ps2_Machine_sbusFlagClear(m,cast<uint32_t>(32ULL),v);
break;}
default:{
ps2_Machine_sbusWrite(m,off,v);
break;}
}}
}
}
// tools/platform/ps2/sifbus.go:112:1
void ps2_Machine_sbusWrite(ps2_Machine* m,uint32_t off,uint32_t v){
{
{
uint32_t i = divi<uint32_t>(off,cast<uint32_t>(16ULL));
if ((i < cast<uint32_t>(8ULL))) {
m->sbus[i] = v;
}
}
}
}
// tools/platform/ps2/texaddr.go:10:1
uint32_t ps2_TexAddrPSMCT32(uint32_t bp,uint32_t bw,uint32_t x,uint32_t y){
{
return ps2_addrPSMCT32(bp,bw,x,y);
}
}
// tools/platform/ps2/texaddr.go:13:1
uint32_t ps2_TexAddrPSMCT16(uint32_t bp,uint32_t bw,uint32_t x,uint32_t y,bool s){
{
return ps2_addrPSMCT16(bp,bw,x,y,s);
}
}
// tools/platform/ps2/texaddr.go:16:1
uint32_t ps2_TexAddrPSMT8(uint32_t bp,uint32_t bw,uint32_t x,uint32_t y){
{
return ps2_addrPSMT8(bp,bw,x,y);
}
}
// tools/platform/ps2/texaddr.go:19:1
std::tuple<uint32_t,uint32_t> ps2_TexAddrPSMT4(uint32_t bp,uint32_t bw,uint32_t x,uint32_t y){
{
return ps2_addrPSMT4(bp,bw,x,y);
}
}
// tools/platform/ps2/texaddr.go:24:1
std::tuple<uint32_t,uint32_t> ps2_CSM1ClutXY(uint32_t i,uint32_t n){
uint32_t x{};
uint32_t y{};
{
if ((n == cast<uint32_t>(256ULL))) {
return {cast<uint32_t>((cast<uint32_t>((i & cast<uint32_t>(7ULL))) | shr<uint32_t>(cast<uint32_t>((i & cast<uint32_t>(16ULL))),cast<int64_t>(1ULL)))),cast<uint32_t>((cast<uint32_t>((shr<uint32_t>(i,cast<int64_t>(3ULL)) & cast<uint32_t>(1ULL))) | cast<uint32_t>((shr<uint32_t>(i,cast<int64_t>(4ULL)) & cast<uint32_t>(14ULL)))))};
}
return {cast<uint32_t>((i & cast<uint32_t>(7ULL))),cast<uint32_t>((shr<uint32_t>(i,cast<int64_t>(3ULL)) & cast<uint32_t>(1ULL)))};
}
}
// tools/platform/ps2/vif.go:34:1
float ps2_f32(Slice<uint8_t> b){
{
return go_math_Float32frombits(ps2_le32gs(b));
}
}
// tools/platform/ps2/vif.go:141:1
ps2_vif* ps2_Machine_ensureVIF(ps2_Machine* m,int64_t idx){
{
if ((!m->vifs[idx])) {
uint32_t size = cast<uint32_t>(4096ULL);
if ((idx == cast<int64_t>(1ULL))) {
size = cast<uint32_t>(16384ULL);
}
ps2_vif* v = arenaNew(ps2_vif{idx,m,cast<uint32_t>(1ULL),cast<uint32_t>(1ULL),{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},Slice<uint8_t>::make(size),Slice<uint8_t>::make(size),Map<std::string,int64_t>{},Map<uint32_t,int64_t>{},Map<uint64_t,ps2_mpgInfo*>{}});
v->vu = vu_New(v->micro,v->data);
if ((idx == cast<int64_t>(1ULL))) {
v->vu->XGKick = [=](auto...args){return ps2_vif_xgkick(v,args...);};
v->vu->OnMaxStore = [=](uint32_t pc,uint32_t qwAddr,uint32_t bits)->void{
{
v->maxStores++;
if ((v->maxStores <= cast<int64_t>(8ULL))) {
ps2_Machine_note(v->m,std::string("VU1 \302\261FLT_MAX store: pc 0x%X -> data qw %d (bits %08X, program 0x%X, top %d)",76),pc,qwAddr,bits,v->lastStart,v->vu->Top);
}
}
}
;
}
else {
m->CPU->COP2 = v->vu;
}
m->vifs[idx] = v;
}
return m->vifs[idx];
}
}
// tools/platform/ps2/vif.go:173:1
void ps2_Machine_vifStart(ps2_Machine* m,int64_t idx,ps2_dmacChan* c){
{
ps2_vif* v = ps2_Machine_ensureVIF(m,idx);
{
switch(shr<uint32_t>((cast<uint32_t>((c->chcr & cast<uint32_t>(12ULL)))),cast<int64_t>(2ULL))){
case cast<uint32_t>(1ULL):{
ps2_Machine_dmacSourceChain(m,ps2_dmacChanForVIF(idx),c,[=](auto...args){return ps2_vif_feed(v,args...);},false);
break;}
default:{
if ((c->qwc == cast<uint32_t>(0ULL))) {
return ;
}
m->feedMadr = c->madr;
ps2_vif_feed(v,ps2_Machine_dmaBytes(m,c->madr,c->qwc));
break;}
}}
}
}
// tools/platform/ps2/vif.go:187:1
int64_t ps2_dmacChanForVIF(int64_t idx){
{
if ((idx == cast<int64_t>(1ULL))) {
return cast<int64_t>(1ULL);
}
return cast<int64_t>(0ULL);
}
}
// tools/platform/ps2/vif.go:195:1
void ps2_vif_feed(ps2_vif* v,Slice<uint8_t> data){
{rrprof::Scope timing(2,"VIF / DMA");
{
int64_t i = cast<int64_t>(0ULL);
{;for (;(i < len(data));){
if ((v->pending > cast<int64_t>(0ULL))) {
int64_t n = v->pending;
if ((n > cast<int64_t>((len(data) - i)))) {
n = cast<int64_t>((len(data) - i));
}
if ((len(v->buf) == cast<int64_t>(0ULL))) {
v->payloadAddr = cast<uint32_t>((v->m->feedMadr + cast<uint32_t>(i)));
}
v->buf = append(v->buf,sub(data,i,cast<int64_t>((i + n))));
v->pending -= n;
i += n;
if ((v->pending == cast<int64_t>(0ULL))) {
ps2_vif_finish(v);
}
continue;
}
if ((cast<int64_t>((len(data) - i)) < cast<int64_t>(4ULL))) {
return ;
}
uint32_t w = ps2_le32gs(rrBorrow(data,i,len(data)));
i += cast<int64_t>(4ULL);
ps2_vif_code(v,w);
}
}}
}
}
// tools/platform/ps2/vif.go:224:1
void ps2_vif_logCode(ps2_vif* v,std::string s){
{
if (((v->idx != cast<int64_t>(1ULL)) || (v->m->VU1DumpIn < cast<int64_t>(0ULL)))) {
return ;
}
v->codeLog = append(v->codeLog,Slice<std::string>{s});
if ((len(v->codeLog) > cast<int64_t>(96ULL))) {
v->codeLog = sub(v->codeLog,cast<int64_t>((len(v->codeLog) - cast<int64_t>(96ULL))),len(v->codeLog));
}
}
}
// tools/platform/ps2/vif.go:235:1
void ps2_vif_code(ps2_vif* v,uint32_t w){
{
uint32_t cmd = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(24ULL))) & cast<uint32_t>(127ULL)));
uint32_t num = cast<uint32_t>(((shr<uint32_t>(w,cast<int64_t>(16ULL))) & cast<uint32_t>(255ULL)));
uint32_t imm = cast<uint32_t>((w & cast<uint32_t>(65535ULL)));
{
switch(cmd){
case cast<uint32_t>(0ULL):{
break;}
case cast<uint32_t>(1ULL):{
v->cl = cast<uint32_t>((imm & cast<uint32_t>(255ULL)));
v->wl = shr<uint32_t>(imm,cast<int64_t>(8ULL));
if ((v->cl == cast<uint32_t>(0ULL))) {
v->cl = cast<uint32_t>(1ULL);
}
ps2_vif_logCode(v,ps2_sprintf(std::string("stcycl cl%d wl%d",16),v->cl,v->wl));
break;}
case cast<uint32_t>(2ULL):{
v->ofst = cast<uint32_t>((imm & cast<uint32_t>(1023ULL)));
ps2_vif_logCode(v,ps2_sprintf(std::string("offset %d (raw 0x%X)",20),v->ofst,imm));
break;}
case cast<uint32_t>(3ULL):{
v->base = cast<uint32_t>((imm & cast<uint32_t>(1023ULL)));
v->tops = v->base;
ps2_vif_logCode(v,ps2_sprintf(std::string("base %d",7),v->base));
break;}
case cast<uint32_t>(4ULL):{
v->itop = imm;
break;}
case cast<uint32_t>(5ULL):{
v->mode = cast<uint32_t>((imm & cast<uint32_t>(3ULL)));
break;}
case cast<uint32_t>(6ULL):{
ps2_vif_count(v,std::string("mskpath3",8));
break;}
case cast<uint32_t>(7ULL):{
v->mark = imm;
break;}
case cast<uint32_t>(16ULL):case cast<uint32_t>(17ULL):case cast<uint32_t>(19ULL):{
break;}
case cast<uint32_t>(20ULL):case cast<uint32_t>(21ULL):{
ps2_vif_count(v,std::string("mscal",5));
{
v->mscal[imm]++;
if ((get(v->mscal,imm) == cast<int64_t>(1ULL))) {
ps2_Machine_note(v->m,std::string("VIF%d: MSCAL of the microprogram at 0x%X (VU%d micro address 0x%X)",66),v->idx,imm,v->idx,cast<uint32_t>((imm * cast<uint32_t>(8ULL))));
}
}
ps2_vif_logCode(v,ps2_sprintf(std::string("mscal 0x%X (tops %d -> top)",27),cast<uint32_t>((imm * cast<uint32_t>(8ULL))),v->tops));
ps2_vif_runVU(v,cast<uint32_t>((imm * cast<uint32_t>(8ULL))),false);
break;}
case cast<uint32_t>(23ULL):{
ps2_vif_count(v,std::string("mscnt",5));
ps2_vif_logCode(v,ps2_sprintf(std::string("mscnt (pc 0x%X, tops %d -> top)",31),v->vu->PC,v->tops));
ps2_vif_runVU(v,cast<uint32_t>(0ULL),true);
break;}
case cast<uint32_t>(32ULL):{
ps2_vif_arm(v,w,cast<int64_t>(4ULL));
break;}
case cast<uint32_t>(48ULL):case cast<uint32_t>(49ULL):{
ps2_vif_arm(v,w,cast<int64_t>(16ULL));
break;}
case cast<uint32_t>(74ULL):{
uint32_t n = num;
if ((n == cast<uint32_t>(0ULL))) {
n = cast<uint32_t>(256ULL);
}
ps2_vif_arm(v,w,cast<int64_t>((cast<int64_t>(n) * cast<int64_t>(8ULL))));
break;}
case cast<uint32_t>(80ULL):case cast<uint32_t>(81ULL):{
uint32_t n = cast<uint32_t>(imm);
if ((n == cast<uint32_t>(0ULL))) {
n = cast<uint32_t>(65536ULL);
}
ps2_vif_arm(v,w,cast<int64_t>((cast<int64_t>(n) * cast<int64_t>(16ULL))));
break;}
default:{
if (((cmd >= cast<uint32_t>(96ULL)) && (cmd < cast<uint32_t>(128ULL)))) {
ps2_vif_arm(v,w,ps2_vif_unpackBytes(v,cmd,num));
return ;
}
ps2_vif_count(v,ps2_sprintf(std::string("cmd 0x%02X unknown",18),cmd));
break;}
}}
}
}
// tools/platform/ps2/vif.go:321:1
void ps2_vif_arm(ps2_vif* v,uint32_t w,int64_t n){
{
v->cmd = w;
v->pending = n;
v->buf = sub(v->buf,0,cast<int64_t>(0ULL));
if ((n == cast<int64_t>(0ULL))) {
ps2_vif_finish(v);
}
}
}
// tools/platform/ps2/vif.go:331:1
void ps2_vif_finish(ps2_vif* v){
{
uint32_t cmd = cast<uint32_t>(((shr<uint32_t>(v->cmd,cast<int64_t>(24ULL))) & cast<uint32_t>(127ULL)));
uint32_t imm = cast<uint32_t>((v->cmd & cast<uint32_t>(65535ULL)));
{
switch(cmd){
case cast<uint32_t>(32ULL):{
v->mask = ps2_le32gs(v->buf);
break;}
case cast<uint32_t>(48ULL):{
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(4ULL));i++){
v->row[i] = ps2_le32gs(rrBorrow(v->buf,cast<int64_t>((i * cast<int64_t>(4ULL))),len(v->buf)));
}
}break;}
case cast<uint32_t>(49ULL):{
{int64_t i = cast<int64_t>(0ULL);for (;(i < cast<int64_t>(4ULL));i++){
v->col[i] = ps2_le32gs(rrBorrow(v->buf,cast<int64_t>((i * cast<int64_t>(4ULL))),len(v->buf)));
}
}break;}
case cast<uint32_t>(74ULL):{
ps2_vif_count(v,std::string("mpg",3));
ps2_vif_logCode(v,ps2_sprintf(std::string("mpg -> micro 0x%X (%d bytes)",28),cast<uint32_t>((imm * cast<uint32_t>(8ULL))),len(v->buf)));
gcopy(sub(v->micro,ps2_min32(cast<uint32_t>((imm * cast<uint32_t>(8ULL))),cast<uint32_t>(len(v->micro))),len(v->micro)),v->buf);
uint64_t h = ps2_fnv64(v->buf);
{
ps2_mpgInfo* info = get(v->mpgSeen,h);
if (bool(info)) {
info->count++;
}
else {
v->mpgSeen[h] = arenaNew(ps2_mpgInfo{cast<uint32_t>((imm * cast<uint32_t>(8ULL))),len(v->buf),cast<int64_t>(1ULL)});
}
}
break;}
case cast<uint32_t>(80ULL):case cast<uint32_t>(81ULL):{
ps2_vif_count(v,std::string("direct",6));
ps2_vif_logCode(v,ps2_sprintf(std::string("direct (%d qw)",14),divi<int64_t>(len(v->buf),cast<int64_t>(16ULL))));
if ((ps2_xferLog && (len(v->buf) >= cast<int64_t>(4096ULL)))) {
print(ps2_sprintf(std::string("  VIF%d DIRECT %d qw from 0x%08X\012",33),v->idx,divi<int64_t>(len(v->buf),cast<int64_t>(16ULL)),v->payloadAddr));
}
ps2_Machine_ensureGS(v->m)->src = std::string("path2 direct",12);
ps2_Machine_gifStream(v->m,v->buf);
break;}
default:{
if (((cmd >= cast<uint32_t>(96ULL)) && (cmd < cast<uint32_t>(128ULL)))) {
ps2_vif_unpack(v,cmd,v->buf);
}
break;}
}}
v->buf = sub(v->buf,0,cast<int64_t>(0ULL));
}
}
// tools/platform/ps2/vif.go:386:1
int64_t ps2_vif_unpackBytes(ps2_vif* v,uint32_t cmd,uint32_t num){
{
if ((num == cast<uint32_t>(0ULL))) {
num = cast<uint32_t>(256ULL);
}
uint32_t vn = cast<uint32_t>(((shr<uint32_t>(cmd,cast<int64_t>(2ULL))) & cast<uint32_t>(3ULL)));
uint32_t vl = cast<uint32_t>((cmd & cast<uint32_t>(3ULL)));
uint32_t vectors = num;
if ((v->wl > v->cl)) {
uint32_t full = divi<uint32_t>(num,v->wl);
uint32_t rem = modi<uint32_t>(num,v->wl);
if ((rem > v->cl)) {
rem = v->cl;
}
vectors = cast<uint32_t>((cast<uint32_t>((full * v->cl)) + rem));
}
uint32_t bytesPer={};
if ((vl == cast<uint32_t>(3ULL))) {
bytesPer = cast<uint32_t>(2ULL);
}
else {
bytesPer = cast<uint32_t>(((cast<uint32_t>((vn + cast<uint32_t>(1ULL)))) * (shr<uint32_t>(cast<uint32_t>(4ULL),vl))));
}
uint32_t n = cast<uint32_t>((vectors * bytesPer));
return cast<int64_t>(((cast<uint32_t>((n + cast<uint32_t>(3ULL)))) & ~(cast<uint32_t>(3ULL))));
}
}
// tools/platform/ps2/vif.go:417:1
void ps2_vif_unpack(ps2_vif* v,uint32_t cmd,Slice<uint8_t> data){
{
uint32_t imm = cast<uint32_t>((v->cmd & cast<uint32_t>(65535ULL)));
uint32_t num = cast<uint32_t>(((shr<uint32_t>(v->cmd,cast<int64_t>(16ULL))) & cast<uint32_t>(255ULL)));
if ((num == cast<uint32_t>(0ULL))) {
num = cast<uint32_t>(256ULL);
}
uint32_t vn = cast<uint32_t>(((shr<uint32_t>(cmd,cast<int64_t>(2ULL))) & cast<uint32_t>(3ULL)));
uint32_t vl = cast<uint32_t>((cmd & cast<uint32_t>(3ULL)));
bool usn = (cast<uint32_t>((v->cmd & cast<uint32_t>(16384ULL))) != cast<uint32_t>(0ULL));
bool flg = (cast<uint32_t>((v->cmd & cast<uint32_t>(32768ULL))) != cast<uint32_t>(0ULL));
std::string cyc = std::string("",0);
if ((v->wl != v->cl)) {
cyc = ps2_sprintf(std::string(" cl%d wl%d",10),v->cl,v->wl);
}
ps2_vif_count(v,ps2_sprintf(std::string("unpack V%d-%d%s",15),cast<uint32_t>((vn + cast<uint32_t>(1ULL))),shr<int64_t>(cast<int64_t>(32ULL),vl),cyc));
if (((v->mask != cast<uint32_t>(0ULL)) && (cast<uint32_t>((v->cmd & cast<uint32_t>(268435456ULL))) != cast<uint32_t>(0ULL)))) {
ps2_vif_count(v,ps2_sprintf(std::string("unpack with STMASK %08X (mask not applied)",42),v->mask));
}
uint32_t qwMask = cast<uint32_t>(cast<int64_t>((divi<int64_t>(len(v->data),cast<int64_t>(16ULL)) - cast<int64_t>(1ULL))));
uint32_t addrQW = cast<uint32_t>((imm & cast<uint32_t>(1023ULL)));
if (flg) {
addrQW += v->tops;
}
addrQW &= qwMask;
uint32_t addr = cast<uint32_t>((addrQW * cast<uint32_t>(16ULL)));
if (((v->idx == cast<int64_t>(1ULL)) && (v->m->VU1DumpIn >= cast<int64_t>(0ULL)))) {
std::array<uint8_t,16> head={};
gcopy(sub(head,0,len(head)),data);
std::string mode = std::string("",0);
if (flg) {
mode = ps2_sprintf(std::string(" (flg, tops %d)",15),v->tops);
}
ps2_vif_logCode(v,ps2_sprintf(std::string("unpack V%d-%d num %d -> qw %d%s from EE 0x%08X  first %08X %08X %08X %08X",73),cast<uint32_t>((vn + cast<uint32_t>(1ULL))),shr<int64_t>(cast<int64_t>(32ULL),vl),num,divi<uint32_t>(addr,cast<uint32_t>(16ULL)),mode,v->payloadAddr,ps2_le32gs(rrBorrow(head,0,len(head))),ps2_le32gs(rrBorrow(head,cast<int64_t>(4ULL),len(head))),ps2_le32gs(rrBorrow(head,cast<int64_t>(8ULL),len(head))),ps2_le32gs(rrBorrow(head,cast<int64_t>(12ULL),len(head)))));
}
int64_t pos = cast<int64_t>(0ULL);
auto read = [&]()->uint32_t{
uint32_t e={};
{
switch(vl){
case cast<uint32_t>(0ULL):{
if ((cast<int64_t>((pos + cast<int64_t>(4ULL))) <= len(data))) {
e = ps2_le32gs(rrBorrow(data,pos,len(data)));
}
pos += cast<int64_t>(4ULL);
break;}
case cast<uint32_t>(1ULL):{
if ((cast<int64_t>((pos + cast<int64_t>(2ULL))) <= len(data))) {
e = cast<uint32_t>((cast<uint32_t>(data[pos]) | shl<uint32_t>(cast<uint32_t>(data[cast<int64_t>((pos + cast<int64_t>(1ULL)))]),cast<int64_t>(8ULL))));
if (((!usn) && (cast<uint32_t>((e & cast<uint32_t>(32768ULL))) != cast<uint32_t>(0ULL)))) {
e |= cast<uint32_t>(4294901760ULL);
}
}
pos += cast<int64_t>(2ULL);
break;}
case cast<uint32_t>(2ULL):{
if ((pos < len(data))) {
e = cast<uint32_t>(data[pos]);
if (((!usn) && (cast<uint32_t>((e & cast<uint32_t>(128ULL))) != cast<uint32_t>(0ULL)))) {
e |= cast<uint32_t>(4294967040ULL);
}
}
pos++;
break;}
}}
return e;
}
;
uint32_t wl = v->wl;
if ((wl == cast<uint32_t>(0ULL))) {
wl = v->cl;
}
{uint32_t i = cast<uint32_t>(0ULL);for (;(i < num);i++){
std::array<uint32_t,4> f={};
bool inCycle = ((wl <= v->cl) || (modi<uint32_t>(i,wl) < v->cl));
{
if (((vl == cast<uint32_t>(3ULL)) && (vn == cast<uint32_t>(3ULL)))){
uint32_t h={};
if ((cast<int64_t>((pos + cast<int64_t>(2ULL))) <= len(data))) {
h = cast<uint32_t>((cast<uint32_t>(data[pos]) | shl<uint32_t>(cast<uint32_t>(data[cast<int64_t>((pos + cast<int64_t>(1ULL)))]),cast<int64_t>(8ULL))));
}
pos += cast<int64_t>(2ULL);
f[cast<int64_t>(0ULL)] = shl<uint32_t>((cast<uint32_t>((h & cast<uint32_t>(31ULL)))),cast<int64_t>(3ULL));
f[cast<int64_t>(1ULL)] = shl<uint32_t>((cast<uint32_t>((shr<uint32_t>(h,cast<int64_t>(5ULL)) & cast<uint32_t>(31ULL)))),cast<int64_t>(3ULL));
f[cast<int64_t>(2ULL)] = shl<uint32_t>((cast<uint32_t>((shr<uint32_t>(h,cast<int64_t>(10ULL)) & cast<uint32_t>(31ULL)))),cast<int64_t>(3ULL));
f[cast<int64_t>(3ULL)] = shl<uint32_t>((shr<uint32_t>(h,cast<int64_t>(15ULL))),cast<int64_t>(7ULL));
}
else if ((!inCycle)){
f = v->row;
}
else {
{uint32_t e = cast<uint32_t>(0ULL);for (;(e <= vn);e++){
f[e] = read();
}
}if ((vn == cast<uint32_t>(0ULL))) {
auto tmp416 = std::make_tuple(f[cast<int64_t>(0ULL)],f[cast<int64_t>(0ULL)],f[cast<int64_t>(0ULL)]);
f[cast<int64_t>(1ULL)] = std::get<0>(tmp416);
f[cast<int64_t>(2ULL)] = std::get<1>(tmp416);
f[cast<int64_t>(3ULL)] = std::get<2>(tmp416);
}
uint32_t last = vn;
if ((vn == cast<uint32_t>(0ULL))) {
last = cast<uint32_t>(3ULL);
}
{
switch(v->mode){
case cast<uint32_t>(1ULL):{
{uint32_t e = cast<uint32_t>(0ULL);for (;(e <= last);e++){
f[e] += v->row[e];
}
}break;}
case cast<uint32_t>(2ULL):{
{uint32_t e = cast<uint32_t>(0ULL);for (;(e <= last);e++){
f[e] += v->row[e];
v->row[e] = f[e];
}
}break;}
}}
}
}
tmp415:;
uint32_t atQW = cast<uint32_t>((addrQW + i));
if ((v->cl > wl)) {
atQW = cast<uint32_t>((cast<uint32_t>((addrQW + cast<uint32_t>((divi<uint32_t>(i,wl) * v->cl)))) + modi<uint32_t>(i,wl)));
}
uint32_t at = cast<uint32_t>(((cast<uint32_t>((atQW & qwMask))) * cast<uint32_t>(16ULL)));
{int64_t e = cast<int64_t>(0ULL);for (;(e < cast<int64_t>(4ULL));e++){
if (((v->idx == cast<int64_t>(1ULL)) && (cast<uint32_t>((f[e] & cast<uint32_t>(2147483647ULL))) == cast<uint32_t>(2139095039ULL)))) {
{
v->maxUnpacks++;
if ((v->maxUnpacks <= cast<int64_t>(8ULL))) {
ps2_Machine_note(v->m,std::string("VIF1 unpacked a \302\261FLT_MAX word into data qw %d (bits %08X) \342\200\224 the garbage was authored EE-side",95),divi<uint32_t>(at,cast<uint32_t>(16ULL)),f[e]);
}
}
}
if ((((v->idx == cast<int64_t>(1ULL)) && (v->m->VIFTinyN > cast<int64_t>(0ULL))) && (vl == cast<uint32_t>(0ULL)))) {
{
uint32_t exp = cast<uint32_t>((f[e] & cast<uint32_t>(2139095040ULL)));
if ((((exp != cast<uint32_t>(0ULL)) && (exp < cast<uint32_t>(872415232ULL))) && (v->tinyUnpacks < v->m->VIFTinyN))) {
v->tinyUnpacks++;
ps2_Machine_note(v->m,std::string("VIF1 unpacked tiny float %g (bits %08X) into data qw %d elem %d \342\200\224 payload from EE 0x%08X (+0x%X)",98),go_math_Float32frombits(f[e]),f[e],divi<uint32_t>(at,cast<uint32_t>(16ULL)),e,v->payloadAddr,pos);
}
}
}
v->data[cast<uint32_t>((cast<uint32_t>((at + cast<uint32_t>((cast<uint32_t>(e) * cast<uint32_t>(4ULL))))) + cast<uint32_t>(0ULL)))] = cast<uint8_t>(f[e]);
v->data[cast<uint32_t>((cast<uint32_t>((at + cast<uint32_t>((cast<uint32_t>(e) * cast<uint32_t>(4ULL))))) + cast<uint32_t>(1ULL)))] = cast<uint8_t>(shr<uint32_t>(f[e],cast<int64_t>(8ULL)));
v->data[cast<uint32_t>((cast<uint32_t>((at + cast<uint32_t>((cast<uint32_t>(e) * cast<uint32_t>(4ULL))))) + cast<uint32_t>(2ULL)))] = cast<uint8_t>(shr<uint32_t>(f[e],cast<int64_t>(16ULL)));
v->data[cast<uint32_t>((cast<uint32_t>((at + cast<uint32_t>((cast<uint32_t>(e) * cast<uint32_t>(4ULL))))) + cast<uint32_t>(3ULL)))] = cast<uint8_t>(shr<uint32_t>(f[e],cast<int64_t>(24ULL)));
}
}}
}}
}
// tools/platform/ps2/vif.go:579:1
void ps2_vif_runVU(ps2_vif* v,uint32_t start,bool cont){
{rrprof::Scope timing(1,"Vector units");
{
v->vu->Top = cast<uint16_t>(v->tops);
v->vu->ITop = cast<uint16_t>(v->itop);
if ((v->ofst != cast<uint32_t>(0ULL))) {
if ((v->tops == v->base)) {
v->tops = cast<uint32_t>(((cast<uint32_t>((v->base + v->ofst))) & cast<uint32_t>(1023ULL)));
}
else {
v->tops = v->base;
}
}
if (cont) {
start = v->vu->PC;
}
v->lastStart = start;
if ((((v->idx == cast<int64_t>(1ULL)) && (v->m->VU1DumpIn == cast<int64_t>(start))) && (v->dumpN < cast<int64_t>(12ULL)))) {
std::string name = ps2_sprintf(std::string("vu1in-%02d.bin",14),v->dumpN);
std::string mname = ps2_sprintf(std::string("vu1in-%02d-micro.bin",20),v->dumpN);
std::string rname = ps2_sprintf(std::string("vu1in-%02d-regs.txt",19),v->dumpN);
v->dumpN++;
Slice<uint8_t> snap = append(Slice<uint8_t>{},v->data);
(void)(ps2_writeFile(name,snap));
(void)(ps2_writeFile(mname,append(Slice<uint8_t>{},v->micro)));
(void)(ps2_writeFile(rname,cast<Slice<uint8_t>>(ps2_Machine_VURegs(v->m,v->idx))));
ps2_Machine_note(v->m,std::string("VU1 input snapshot at MSCAL of 0x%X (top %d): %s + %s + %s",58),start,v->vu->Top,name,mname,rname);
if ((v->codeTrailsN < cast<int64_t>(3ULL))) {
v->codeTrailsN++;
ps2_Machine_note(v->m,std::string("VIF1 code trail %d into this MSCAL (oldest first):",50),v->codeTrailsN);
{auto&& tmp417 = v->codeLog;
for(int64_t tmp418=0;tmp418<len(tmp417);++tmp418){
auto i=tmp418;auto s=tmp417[tmp418];ps2_Machine_note(v->m,std::string("VIF1 t%d.%02d  %s",17),v->codeTrailsN,i,s);
}}
}
v->codeLog = sub(v->codeLog,0,cast<int64_t>(0ULL));
}
if ((v->idx == cast<int64_t>(1ULL))) {
vu_VU_ResetBranchLog(v->vu,cast<int64_t>(48ULL));
}
auto tmp419 = ps2_Machine_profVU1Start(v->m);
time_Time pt = std::get<0>(tmp419);
int64_t rasterBefore = std::get<1>(tmp419);
auto tmp420 = vu_VU_Run(v->vu,start,cast<int64_t>(1048576ULL));
int64_t steps = std::get<0>(tmp420);
bool ended = std::get<1>(tmp420);
ps2_Machine_profVU1End(v->m,pt,rasterBefore);
v->vuSteps += cast<uint64_t>(steps);
if ((!ended)) {
ps2_vif_count(v,ps2_sprintf(std::string("vu1 program 0x%X hit the step budget (pc 0x%X)",46),start,v->vu->PC));
if ((v->runawayDumps < cast<int64_t>(2ULL))) {
v->runawayDumps++;
ps2_Machine_note(v->m,std::string("VU1 runaway: entry 0x%X (top %d), stopped at 0x%X after %d steps, %d taken branches; the last (delay-slot -> target):",117),start,v->vu->Top,v->vu->PC,steps,vu_VU_BranchCount(v->vu));
{auto&& tmp421 = vu_VU_BranchTrail(v->vu);
for(int64_t tmp422=0;tmp422<len(tmp421);++tmp422){
auto br=tmp421[tmp422];ps2_Machine_note(v->m,std::string("VU1   0x%04X -> 0x%04X",22),br[cast<int64_t>(0ULL)],br[cast<int64_t>(1ULL)]);
}}
std::string mName = ps2_sprintf(std::string("vu1-runaway-%d-micro.bin",24),v->runawayDumps);
std::string dName = ps2_sprintf(std::string("vu1-runaway-%d-data.bin",23),v->runawayDumps);
(void)(ps2_writeFile(mName,append(Slice<uint8_t>{},v->micro)));
(void)(ps2_writeFile(dName,append(Slice<uint8_t>{},v->data)));
ps2_Machine_note(v->m,std::string("VU1 runaway memories: %s / %s",29),mName,dName);
}
}
}
}
}
// tools/platform/ps2/vif.go:653:1
void ps2_vif_xgkick(ps2_vif* v,uint32_t qw){
{
ps2_vif_count(v,std::string("xgkick",6));
ps2_GS* gs = ps2_Machine_ensureGS(v->m);
gs->src = ps2_sprintf(std::string("vu1 program 0x%X (kick at qw %d from pc 0x%X, top %d)",53),v->lastStart,qw,v->vu->PC,v->vu->Top);
Slice<uint8_t> data = sub(v->data,cast<uint32_t>((qw * cast<uint32_t>(16ULL))),len(v->data));
int64_t n = ps2_gifPacketLen(data);
gs->srcData = sub(data,0,n);
uint32_t in = cast<uint32_t>((cast<uint32_t>(v->vu->Top) * cast<uint32_t>(16ULL)));
if ((cast<int64_t>(in) < len(v->data))) {
gs->srcIn = sub(v->data,in,len(v->data));
}
auto tmp423 = std::make_tuple(v->micro,v->data);
gs->srcMicro = std::get<0>(tmp423);
gs->srcVUData = std::get<1>(tmp423);
if ((!v->kickDumped)) {
v->kickDumped = true;
int64_t dump = n;
if ((dump > cast<int64_t>(192ULL))) {
dump = cast<int64_t>(192ULL);
}
{int64_t o = cast<int64_t>(0ULL);for (;(o < dump);o += cast<int64_t>(16ULL)){
ps2_Machine_note(v->m,std::string("VU1 first XGKICK qw+%d: %08X %08X %08X %08X",43),divi<int64_t>(o,cast<int64_t>(16ULL)),ps2_le32gs(rrBorrow(data,o,len(data))),ps2_le32gs(rrBorrow(data,cast<int64_t>((o + cast<int64_t>(4ULL))),len(data))),ps2_le32gs(rrBorrow(data,cast<int64_t>((o + cast<int64_t>(8ULL))),len(data))),ps2_le32gs(rrBorrow(data,cast<int64_t>((o + cast<int64_t>(12ULL))),len(data))));
}
}}
ps2_Machine_gifPacket(v->m,sub(data,0,n));
}
}
// tools/platform/ps2/vif.go:681:1
void ps2_vif_count(ps2_vif* v,std::string what){
{
v->census[ps2_sprintf(std::string("VIF%d %s",8),v->idx,what)]++;
}
}
// tools/platform/ps2/vif.go:684:1
uint64_t ps2_fnv64(Slice<uint8_t> b){
{
uint64_t h = cast<uint64_t>(14695981039346656037ULL);
{auto&& tmp424 = b;
for(int64_t tmp425=0;tmp425<len(tmp424);++tmp425){
auto c=tmp424[tmp425];h = cast<uint64_t>(((cast<uint64_t>((h ^ cast<uint64_t>(c)))) * cast<uint64_t>(1099511628211ULL)));
}}
return h;
}
}
// tools/platform/ps2/vif.go:696:1
vu_VU* ps2_Machine_VU(ps2_Machine* m,int64_t idx){
{
if (((idx < cast<int64_t>(0ULL)) || (idx > cast<int64_t>(1ULL)))) {
return {};
}
return ps2_Machine_ensureVIF(m,idx)->vu;
}
}
// tools/platform/ps2/vif.go:705:1
Slice<uint8_t> ps2_Machine_VUMicro(ps2_Machine* m,int64_t idx){
{
if ((((idx < cast<int64_t>(0ULL)) || (idx > cast<int64_t>(1ULL))) || (!m->vifs[idx]))) {
return {};
}
return m->vifs[idx]->micro;
}
}
// tools/platform/ps2/vif.go:714:1
Slice<uint8_t> ps2_Machine_VUDataMem(ps2_Machine* m,int64_t idx){
{
if ((((idx < cast<int64_t>(0ULL)) || (idx > cast<int64_t>(1ULL))) || (!m->vifs[idx]))) {
return {};
}
return m->vifs[idx]->data;
}
}
// tools/platform/ps2/vif.go:725:1
std::string ps2_Machine_VURegs(ps2_Machine* m,int64_t idx){
{
if ((((idx < cast<int64_t>(0ULL)) || (idx > cast<int64_t>(1ULL))) || (!m->vifs[idx]))) {
return std::string("",0);
}
vu_VU* u = m->vifs[idx]->vu;
std::string s = ps2_sprintf(std::string("VU%d registers (pc=%04x)\012",25),idx,u->PC);
auto f = [&](uint32_t b)->float{
return go_math_Float32frombits(b);
}
;
{int64_t r = cast<int64_t>(0ULL);for (;(r < cast<int64_t>(32ULL));r++){
std::array<uint32_t,4> v = u->VF[r];
s += ps2_sprintf(std::string("  vf%02d  %08x %08x %08x %08x   %13.6g %13.6g %13.6g %13.6g\012",60),r,v[cast<int64_t>(0ULL)],v[cast<int64_t>(1ULL)],v[cast<int64_t>(2ULL)],v[cast<int64_t>(3ULL)],f(v[cast<int64_t>(0ULL)]),f(v[cast<int64_t>(1ULL)]),f(v[cast<int64_t>(2ULL)]),f(v[cast<int64_t>(3ULL)]));
}
}{int64_t r = cast<int64_t>(0ULL);for (;(r < cast<int64_t>(16ULL));r += cast<int64_t>(8ULL)){
s += std::string(" ",1);
{int64_t j = cast<int64_t>(0ULL);for (;(j < cast<int64_t>(8ULL));j++){
s += ps2_sprintf(std::string(" vi%02d=%04x",12),cast<int64_t>((r + j)),u->VI[cast<int64_t>((r + j))]);
}
}s += std::string("\012",1);
}
}s += ps2_sprintf(std::string("  acc = %g %g %g %g   q=%g i=%g\012",32),u->ACC[cast<int64_t>(0ULL)],u->ACC[cast<int64_t>(1ULL)],u->ACC[cast<int64_t>(2ULL)],u->ACC[cast<int64_t>(3ULL)],u->Q,u->I);
return s;
}
}
// tools/platform/ps2/vif.go:751:1
std::string ps2_Machine_VIFCensus(ps2_Machine* m){
{
std::string s = std::string("",0);
{auto&& tmp426 = m->vifs;
for(int64_t tmp427=0;tmp427<len(tmp426);++tmp427){
auto v=tmp426[tmp427];if ((!v)) {
continue;
}
{auto&& tmp428 = ps2_sortedCounts(v->census);
for(int64_t tmp429=0;tmp429<len(tmp428);++tmp429){
auto kv=tmp428[tmp429];s += ps2_sprintf(std::string("      %-24s %d\012",15),kv.name,kv.n);
}}
if ((v->vuSteps > cast<uint64_t>(0ULL))) {
s += ps2_sprintf(std::string("      VU%d instructions run    %d\012",34),v->idx,v->vuSteps);
}
if ((len(v->mpgSeen) > cast<int64_t>(0ULL))) {
Slice<ps2_mpgInfo*> progs={};
{auto&& tmp430 = v->mpgSeen;
for(auto [tmp431,tmp432]:tmp430){
auto i=tmp432;progs = append(progs,Slice<ps2_mpgInfo*>{i});
}}
{int64_t i = cast<int64_t>(1ULL);for (;(i < len(progs));i++){
{int64_t j = i;for (;((j > cast<int64_t>(0ULL)) && (((progs[j]->addr < progs[cast<int64_t>((j - cast<int64_t>(1ULL)))]->addr) || (((progs[j]->addr == progs[cast<int64_t>((j - cast<int64_t>(1ULL)))]->addr) && (progs[j]->size < progs[cast<int64_t>((j - cast<int64_t>(1ULL)))]->size))))));j--){
auto tmp433 = std::make_tuple(progs[cast<int64_t>((j - cast<int64_t>(1ULL)))],progs[j]);
progs[j] = std::get<0>(tmp433);
progs[cast<int64_t>((j - cast<int64_t>(1ULL)))] = std::get<1>(tmp433);
}
}}
}s += ps2_sprintf(std::string("      VIF%d distinct microprograms: %d\012",39),v->idx,len(progs));
{auto&& tmp434 = progs;
for(int64_t tmp435=0;tmp435<len(tmp434);++tmp435){
auto i=tmp434[tmp435];s += ps2_sprintf(std::string("        micro 0x%04X  %5d bytes  uploaded %d times\012",51),i->addr,i->size,i->count);
}}
}
}}
if ((s == std::string("",0))) {
return std::string("",0);
}
return (std::string("what the VPU interfaces were asked (the render transport):\012",59) + s);
}
}
// tools/platform/ps2/vif.go:791:1
Slice<ps2_countKV> ps2_sortedCounts(Map<std::string,int64_t> m){
{
Slice<ps2_countKV> all={};
{auto&& tmp436 = m;
for(auto [tmp437,tmp438]:tmp436){
auto k=tmp437;auto n=tmp438;all = append(all,Slice<ps2_countKV>{ps2_countKV{k,n}});
}}
{int64_t i = cast<int64_t>(1ULL);for (;(i < len(all));i++){
{int64_t j = i;for (;((j > cast<int64_t>(0ULL)) && (((all[j].n > all[cast<int64_t>((j - cast<int64_t>(1ULL)))].n) || (((all[j].n == all[cast<int64_t>((j - cast<int64_t>(1ULL)))].n) && (all[j].name < all[cast<int64_t>((j - cast<int64_t>(1ULL)))].name))))));j--){
auto tmp439 = std::make_tuple(all[cast<int64_t>((j - cast<int64_t>(1ULL)))],all[j]);
all[j] = std::get<0>(tmp439);
all[cast<int64_t>((j - cast<int64_t>(1ULL)))] = std::get<1>(tmp439);
}
}}
}return all;
}
}
// tools/platform/ps2/vif.go:804:1
uint32_t ps2_min32(uint32_t a,uint32_t b){
{
if ((a < b)) {
return a;
}
return b;
}
}
