// Actor-binding oracle (Part V §3): instead of pattern-matching actor code for
// model references, this runs each actor's REAL create/init code on the tools/arm
// CPU over a flat DS memory image and records which files the code asks the
// loader for. The oracle is function-level: it boots the game's own crt0 to
// main() (so ITCM/DTCM/autoload sections and the pre-main heap are exactly what
// the game builds), loads overlay 0 + overlay 2 + one bank/level overlay at
// their table addresses, runs each overlay's static-initializer list natively
// (so the file-slot registrations execute for real), then calls every actor
// profile's create function and its init (vtable slot 0) with the loader choke
// points trapped:
//
//   - $0201818C(id, flag) — the load-by-internal-file-ID loader (both thunks
//     $0201816C/$0201817C funnel here, as does the decompressing wrapper at
//     $02018100). The trap records the requested ID and serves the real
//     extracted file into scratch RAM, so relocation/parse code after the load
//     runs on real data. IDs >= $8000 resolve through the archive descriptor
//     table (archives.go) and are served decompressed, exactly like the
//     RAM-resident-archive branch of the real function.
//   - $020189F0(path) — the FS path->file-ID resolver (an FNT walk that needs
//     the card). Answered from our own FNT parse of the image. Overlay 0's
//     initializer at $020AA420 calls this $80A times to build the internal-ID ->
//     FS-ID remap table at [[$0209D3B8]] (which then runs natively on the heap).
//   - $020731DC(aux, cb, node) — async-load registration (a linked-list push,
//     head at $020AA3F0). Recorded and emulated; after create+init the oracle
//     drains the queue the way the game's own pump at $02072F44 does — pop,
//     call cb(aux, -1) — so deferred loads surface too.
//
// The recorded .bmd requests are the actor's true model bindings, the .bca
// requests its true animation clips. No heuristics: if an actor's code does not
// ask for a model, it has none (it draws procedurally or not at all).
package sm64ds

import (
	"fmt"
	"path/filepath"
	"strings"

	"retroreverse.com/tools/arm"
	"retroreverse.com/tools/nds"
)

// Traced addresses (see package doc and Part V §3).
const (
	oMain     = 0x02007000 // real main()
	oSync     = 0x0205BAD8 // OS sync-with-ARM7 (the $0205BB54 rendezvous loop's fn)
	oIPCReady = 0x0205BA3C // "IPC channel ready" bit test ($027FFC00+$388) the PXI init polls
	oFSPrioGt = 0x02057E7C // FS thread-priority get — sleeps on the FS thread; stubbed to 0
	oFSPrioSt = 0x02057E84 // FS thread-priority set — ditto (ovl0's init brackets its loop with these)
	oLoader   = 0x0201818C // load-by-internal-ID (inner fn of $0201816C/$0201817C)
	oSlotLoad = 0x02017C54 // load-by-slot: slot's u16 is a FAT/archive id ($02017BC4 = acquire+refcount)
	oSlotAcq  = 0x02017BC4 // acquire a slot (refcount at +2, loads on first use)
	oSlotAnnA = 0x02017A9C // async-load cbs: publish the slot id to the loader thread's
	oSlotAnnB = 0x02017AB4 // mailbox ($0209D3BC via $02017E34); the thread then acquires it
	oWrapGet  = 0x02016E70 // get-or-create render wrapper for a model: r0 = the model
	// data pointer, which identifies the file even when the resource was preloaded
	oRoSize = 0x02046564 // render-object size from a model header (every render-object
	// build calls it with r0 = the model) — the second display-binding watch
	oPathToID = 0x020189F0 // FS path -> file ID (FNT walk)
	oAsyncReg = 0x020731DC // async-load registration
	oAsyncHd  = 0x020AA3F0 // async queue head
	oSpawnCtx = 0x02043180 // factory's spawn-context store (actor,a,b,c -> globals)
	oProfArr  = 0x02090864 // actor profile-pointer array (u32 x 783, "*ERR*" ends it)
	oProfMax  = 783
	oOvl0Init = 0x020AA420 // overlay 0's initializer (its ctor list is a lone NULL)
	oErrPrint = 0x02018E68 // error printf (message in r0), always followed by terminate
	oPanic    = 0x02019740 // terminate: IME off + infinite error-render loop
	oEngInit  = 0x0201A054 // engine init: heaps, actor system, profile-table global
	// ($0201A128), VRAM texture/palette arenas ($02045D20), ...
	oGXInit = 0x02053A8C // GX/video hardware init inside it — sleeps the thread on
	// geometry-engine acks; pure hardware programming, stubbed
	oPXIRecv = 0x0205B858 // PXI receive setup ($0205B864 = the ARM7 boot-msg wait
	// the dual-core oracle stalled on); nothing ARM7-side exists here, stubbed
	oSndInit = 0x02060890 // sound-system init: sends PXI channel-$B commands to the
	// ARM7 and sleeps the thread on the reply; stubbed (sound is ARM7-side)
	oSndData = 0x020603C8 // sound-data setup (same module, same PXI wait); stubbed
	oSaveRd  = 0x02018434 // backup/save read (card SPI via ARM7); stubbed to "no save"
	oSndArc  = 0x02050F34 // sound-archive init: FS-opens /data/sound_data.sdat through
	// the SDK archive procs FS_Init (never run here) would install; stubbed
	oSndPump  = 0x020196CC // game-side sound command pump (flush + wait on the ARM7); stubbed
	oGameLoop = 0x02007040 // main()'s frame-loop call ($020197B8, never returns); the
	// oracle runs main() up to here so OS arenas, the root heap ($0203CB04 via
	// $0203C408), FS state and the engine init all happen in the game's own order
	oHeapNew  = 0x0203C844 // heap create (size, parent [0 = current], style 4 = expheap)
	oCurHeap  = 0x020A0EA0 // current-heap global (allocs with no explicit heap use it)
	oMainHeap = 0x020A0EAC // the main heap the engine init creates
	oGXWait   = 0x0205EA10 // wait for GX status bits ($020A80CC+$36, IRQ-maintained) to clear
	oGXTest   = 0x0205E9FC // test GX status bits ($020A80CC+$34); nonzero panics the caller

	oScratch  = 0x02800000 // bump allocator for served files (outside game RAM)
	oScratchE = 0x03E00000
	oRetAddr  = 0x03F00000 // call() return sentinel
	busSize   = 0x04000000
	pageBits  = 12
)

// FileReq is one recorded loader request or resource reference.
type FileReq struct {
	ID    int    `json:"id"`
	Name  string `json:"name"` // internal path, or archive stem ("arc0_5")
	Flag  uint32 `json:"flag"`
	Phase string `json:"phase"`          // "create", "init", "async", "ovl<N>-init"
	Kind  string `json:"kind,omitempty"` // "" = load; "acquire" = slot touch; "display" = render object built on it
}

// Stem names the model/clip file without directories/extension (archive members
// keep their arcN_M stem).
func (f FileReq) Stem() string {
	if f.ID >= 0x8000 {
		return f.Name
	}
	return strings.TrimSuffix(filepath.Base(f.Name), filepath.Ext(f.Name))
}

// Collider is one dBgW collider the actor registered in the 24-slot table at
// $020A0C80 during create/init (Part VI). The oracle spawns actors at the
// origin with no yaw, so Mtx is the actor's OWN collider transform — its
// authored scale/rotation/offset — with the placement pose still to compose.
type Collider struct {
	KCL    string    `json:"kcl"`              // .kcl stem the collider walks
	Class  string    `json:"class"`            // "Kc", "KcMbg", "KcMbgSclY"
	Mtx    [12]int32 `json:"mtx,omitempty"`    // local->world MtxFx43 at +$134 (fx12), Mbg classes
	ScaleY int32     `json:"scaleY,omitempty"` // fx12 local-Y scale at +$1C8 (SclY class)
}

// ActorRun is the record of one create+init execution.
type ActorRun struct {
	Actor     int        `json:"actor"`
	Config    int        `json:"config"` // extra overlay loaded (-1 = engine only)
	Params    [3]int     `json:"params"` // par1, par2, par3 as placed
	Create    uint32     `json:"create"`
	Obj       uint32     `json:"obj"` // create's return (0 = refused)
	Files     []FileReq  `json:"files,omitempty"`
	Colliders []Collider `json:"colliders,omitempty"`
	Notes     []string   `json:"notes,omitempty"`
}

// obus is the oracle's machine model: flat RAM with logged I/O, an advancing
// VCOUNT, the ARM9 hardware divider/square-root blocks, and immediate DMA.
type obus struct {
	mem    []byte
	io     map[uint32]uint32
	vcount uint32

	// dirty-page tracking (for cheap state restore between actor runs)
	dirty   []bool
	touched []uint32
	track   bool

	// read watch (observe-only): every byte read in [watchLo,watchHi) is
	// reported — the read-side counterpart of the write profiler
	watchLo, watchHi uint32
	watchFn          func(addr uint32)
}

func newOBus() *obus {
	return &obus{mem: make([]byte, busSize), io: map[uint32]uint32{}, dirty: make([]bool, busSize>>pageBits)}
}

func (b *obus) mark(a uint32) {
	p := a >> pageBits
	if !b.dirty[p] {
		b.dirty[p] = true
		b.touched = append(b.touched, p)
	}
}

func (b *obus) Read(a uint32) byte {
	if a < busSize {
		if b.watchFn != nil && a >= b.watchLo && a < b.watchHi {
			b.watchFn(a)
		}
		return b.mem[a]
	}
	if a>>24 == 0x04 {
		reg := a &^ 3
		switch a {
		case 0x04000004: // DISPSTAT low: VBlank flag toggles so waiters move
			b.vcount++
			return byte(b.vcount & 1)
		case 0x04000005:
			return 0
		case 0x04000006: // VCOUNT advances so scanline-poll loops progress
			b.vcount = (b.vcount + 1) & 0x1FF
			return byte(b.vcount)
		case 0x04000007:
			return byte(b.vcount >> 8)
		}
		switch reg {
		case 0x040002A0, 0x040002A4, 0x040002A8, 0x040002AC:
			return byte(b.divResult(reg) >> (8 * (a & 3)))
		case 0x040002B4:
			return byte(b.sqrtResult() >> (8 * (a & 3)))
		}
		if reg >= 0x040000B0 && reg < 0x04000100 {
			// DMA block: transfers execute immediately on write, so busy bits read
			// clear (no DS DMA source/dest address has bit 31 set, so masking the
			// whole block is safe)
			return byte((b.io[reg] &^ 0x80000000) >> (8 * (a & 3)))
		}
		if v, ok := b.io[reg]; ok {
			return byte(v >> (8 * (a & 3)))
		}
	}
	return 0
}

func (b *obus) Write(a uint32, v byte) {
	if a < busSize {
		if b.track {
			b.mark(a)
		}
		b.mem[a] = v
		return
	}
	if a>>24 == 0x04 {
		reg := a &^ 3
		shift := 8 * (a & 3)
		b.io[reg] = b.io[reg]&^(0xFF<<shift) | uint32(v)<<shift
		// immediate DMA on enable write (top byte of a channel's control word)
		if reg >= 0x040000B0 && reg < 0x04000100 && (reg-0xB0)%12 == 8 && a&3 == 3 && v&0x80 != 0 {
			b.runDMA(reg)
		}
	}
}

// runDMA performs one DS DMA transfer immediately (start-timing bits ignored:
// anything the oracle sees is either immediate or card DMA whose data our traps
// already provided, so completing it at once is the useful model).
func (b *obus) runDMA(ctrlReg uint32) {
	base := ctrlReg - 8
	src := b.io[base]
	dst := b.io[base+4]
	ctrl := b.io[ctrlReg]
	n := ctrl & 0x1FFFFF
	if n == 0 {
		n = 0x200000
	}
	unit := uint32(2)
	if ctrl&(1<<26) != 0 {
		unit = 4
	}
	step := func(mode uint32) int64 {
		switch mode {
		case 1:
			return -int64(unit)
		case 2:
			return 0
		}
		return int64(unit)
	}
	ds, ss := step((ctrl>>21)&3), step((ctrl>>23)&3)
	for i := uint32(0); i < n; i++ {
		s := uint32(int64(src) + int64(i)*ss)
		d := uint32(int64(dst) + int64(i)*ds)
		for k := uint32(0); k < unit; k++ {
			b.Write(d+k, b.Read(s+k))
		}
	}
	b.io[ctrlReg] = ctrl &^ 0x80000000
}

func (b *obus) u64(reg uint32) uint64 {
	return uint64(b.io[reg]) | uint64(b.io[reg+4])<<32
}

// divResult models the ARM9 divider (DIVCNT $04000280, operands $290/$298).
func (b *obus) divResult(reg uint32) uint32 {
	mode := b.io[0x04000280] & 3
	var num, den int64
	switch mode {
	case 0:
		num, den = int64(int32(b.io[0x04000290])), int64(int32(b.io[0x04000298]))
	case 1:
		num, den = int64(b.u64(0x04000290)), int64(int32(b.io[0x04000298]))
	default:
		num, den = int64(b.u64(0x04000290)), int64(b.u64(0x04000298))
	}
	var q, r int64
	if den == 0 { // GBATEK: div0 -> +/-1 with remainder = numerator
		q, r = 1, num
		if num < 0 {
			q = -1
		}
	} else {
		q, r = num/den, num%den
	}
	switch reg {
	case 0x040002A0:
		return uint32(q)
	case 0x040002A4:
		return uint32(q >> 32)
	case 0x040002A8:
		return uint32(r)
	default:
		return uint32(r >> 32)
	}
}

// sqrtResult models the ARM9 square-root block (SQRTCNT $040002B0, param $2B8).
func (b *obus) sqrtResult() uint32 {
	v := b.u64(0x040002B8)
	if b.io[0x040002B0]&1 == 0 {
		v = uint64(b.io[0x040002B8])
	}
	var r uint64
	for bit := uint64(1) << 31; bit != 0; bit >>= 1 {
		if t := r | bit; t*t <= v {
			r = t
		}
	}
	return uint32(r)
}

func (b *obus) r32(a uint32) uint32 {
	return uint32(b.Read(a)) | uint32(b.Read(a+1))<<8 | uint32(b.Read(a+2))<<16 | uint32(b.Read(a+3))<<24
}
func (b *obus) w32(a, v uint32) {
	for i := uint32(0); i < 4; i++ {
		b.Write(a+i, byte(v>>(8*i)))
	}
}

// Oracle runs game code function-by-function over the booted image.
type Oracle struct {
	ls  *LevelSet
	bus *obus
	cpu *arm.CPU

	pathID  map[string]int // FNT path (leading slash) -> FS file ID
	fatName map[int]string // FS file ID -> path
	traps   map[uint32]func(*Oracle)
	watch   map[uint32]func(*Oracle) // observe-only (execution continues natively)
	arcKind map[int]string           // archive-member content classification cache
	pool    map[string]bool          // the engine's common preload set (by stem)

	phase   string
	cur     *ActorRun // collects FileReqs during actor runs
	initReq []FileReq // requests made during overlay static inits
	asyncQ  []uint32  // async cb/aux pairs recorded (node addrs)

	scratch uint32
	callSP  uint32
	served  []servedFile // buffer intervals handed out by the traps

	engineSnap []byte  // full RAM after engine init
	engineCPU  arm.CPU // CPU state ditto
	engScratch uint32  // scratch cursor at the engine snapshot
	engServedN int
	cfgServedN int
	cfgPages   map[uint32][]byte
	cfgScratch uint32

	Trace func(string) // optional progress logger
}

// NewOracle boots the ARM9 to main() and prepares the trap set. The LevelSet
// supplies the image, the extracted files and the archive tables.
func NewOracle(ls *LevelSet) (*Oracle, error) {
	o := &Oracle{ls: ls, bus: newOBus(), pathID: map[string]int{}, arcKind: map[int]string{}, pool: map[string]bool{}}
	for _, f := range ls.rom.Files {
		o.pathID["/"+f.Path] = f.ID
	}
	o.traps = map[uint32]func(*Oracle){
		oSync:     (*Oracle).trapSync,
		oIPCReady: func(o *Oracle) { o.cpu.R[0] = 1; o.ret() },
		oFSPrioGt: func(o *Oracle) { o.cpu.R[0] = 0; o.ret() },
		oFSPrioSt: func(o *Oracle) { o.cpu.R[0] = 0; o.ret() },
		oLoader:   (*Oracle).trapLoader,
		oSlotLoad: (*Oracle).trapSlotLoad,
		oPathToID: (*Oracle).trapPathToID,
		oAsyncReg: (*Oracle).trapAsyncReg,
		oGXWait:   func(o *Oracle) { o.ret() },
		oGXInit:   func(o *Oracle) { o.ret() },
		oPXIRecv:  func(o *Oracle) { o.cpu.R[0] = 0; o.ret() },
		oSndInit:  func(o *Oracle) { o.cpu.R[0] = 0; o.ret() },
		oSndData:  func(o *Oracle) { o.cpu.R[0] = 0; o.ret() },
		oSaveRd:   func(o *Oracle) { o.cpu.R[0] = 0; o.ret() },
		oSndArc:   func(o *Oracle) { o.cpu.R[0] = 0; o.ret() },
		oSndPump:  func(o *Oracle) { o.cpu.R[0] = 0; o.ret() },
		oGXTest:   func(o *Oracle) { o.cpu.R[0] = 0; o.ret() },
		oErrPrint: func(o *Oracle) {
			o.note("game error: %q (lr=%08X)", o.cstr(o.cpu.R[0]), o.cpu.R[14])
			o.ret()
		},
		oPanic: func(o *Oracle) {
			// Always a guarded `BL $02019740` (capacity/error checks): returning
			// resumes the caller's success path. VRAM-arena overflow checks fire
			// spuriously here (the arenas' video-init never runs); the run's notes
			// keep the event so real failures stay visible.
			var stack []string
			for i := uint32(0); i < 16; i++ {
				if w := o.bus.r32(o.cpu.R[13] + i*4); w >= 0x02004000 && w < 0x02150000 {
					stack = append(stack, fmt.Sprintf("%08X", w))
				}
			}
			o.note("game terminate called (lr=%08X r0=%08X r4=%08X r5=%08X stk=%s) — resumed",
				o.cpu.R[14], o.cpu.R[0], o.cpu.R[4], o.cpu.R[5], strings.Join(stack, " "))
			o.ret()
		},
	}
	o.fatName = map[int]string{}
	for _, f := range ls.rom.Files {
		o.fatName[f.ID] = "/" + f.Path
	}
	// A slot acquire on an already-loaded slot performs no load; the reference
	// itself is the binding, so observe every acquire.
	o.watch = map[uint32]func(*Oracle){
		oSlotAcq: func(o *Oracle) {
			if o.cur == nil {
				return // overlay-init acquires are recorded by the load trap
			}
			slot := o.cpu.R[0]
			id := int(o.bus.r32(slot) & 0xFFFF)
			name, _ := o.nameFor(id)
			o.record(FileReq{ID: id, Name: name, Phase: o.phase, Kind: "acquire"})
		},
		oWrapGet: (*Oracle).watchDisplay,
		oRoSize:  (*Oracle).watchDisplay,
	}
	o.cpu = arm.NewCPU(o.bus)
	o.cpu.SWI = o.swi
	if err := o.boot(); err != nil {
		return nil, err
	}
	return o, nil
}

func (o *Oracle) trace(f string, a ...any) {
	if o.Trace != nil {
		o.Trace(fmt.Sprintf(f, a...))
	}
}

// boot runs the compressed image's crt0 (self-decompression, section setup, SDK
// init) to main(), with the ARM7 rendezvous stubbed.
func (o *Oracle) boot() error {
	h := o.ls.rom.Header
	arm9 := o.ls.rom.ARM9()
	for i, v := range arm9 {
		o.bus.mem[h.ARM9RAMAddr+uint32(i)] = v
	}
	c := o.cpu
	c.Mode = arm.ModeSVC
	c.R[15] = h.ARM9Entry
	const budget = 60_000_000
	for i := 0; i < budget; i++ {
		pc := c.R[15]
		if pc == oMain {
			o.callSP = c.R[13]
			o.trace("boot: reached main() after %d instrs, sp=%08X", i, o.callSP)
			return nil
		}
		if t, ok := o.traps[pc]; ok {
			t(o)
			continue
		}
		c.Step()
		if c.Halted {
			return fmt.Errorf("boot halted at %08X: %s", c.R[15], c.HaltReason)
		}
	}
	return fmt.Errorf("boot: main() not reached in %d instrs (pc=%08X)", budget, c.R[15])
}

// ret returns from the trapped function: BX LR.
func (o *Oracle) ret() {
	c := o.cpu
	c.R[15] = c.R[14] &^ 1
	c.Thumb = c.R[14]&1 != 0
}

func (o *Oracle) trapSync() { o.cpu.R[0] = 0; o.ret() }

// trapPathToID answers the FNT path->ID walk from our own FNT parse.
func (o *Oracle) trapPathToID() {
	c := o.cpu
	path := o.cstr(c.R[0])
	id, ok := o.pathID[path]
	if !ok {
		id, ok = o.pathID["/"+path]
	}
	if !ok {
		id = 0
	}
	c.R[0] = uint32(id)
	o.ret()
}

// trapLoader records the request and serves the real file bytes into scratch.
func (o *Oracle) trapLoader() {
	c := o.cpu
	id, flag := int(c.R[0]), c.R[1]
	name, data := o.fileFor(id)
	o.record(FileReq{ID: id, Name: name, Flag: flag, Phase: o.phase})
	if data == nil {
		c.R[0] = 0
	} else {
		c.R[0] = o.serve(data)
		o.served = append(o.served, servedFile{c.R[0], c.R[0] + uint32(len(data)), id, name})
	}
	o.ret()
}

// trapSlotLoad replaces $02017C54: load the file named by the slot's u16
// (a FAT id from the remap table, or an archive-flagged id), store the buffer
// at slot+4 and return it. The acquire wrapper $02017BC4 (refcount at +2) runs
// natively around this.
func (o *Oracle) trapSlotLoad() {
	c := o.cpu
	slot := c.R[0]
	id := int(o.bus.r32(slot) & 0xFFFF)
	name, data := o.nameFor(id)
	o.record(FileReq{ID: id, Name: name, Flag: 0, Phase: o.phase})
	var addr uint32
	if data != nil {
		addr = o.serve(data)
		o.served = append(o.served, servedFile{addr, addr + uint32(len(data)), id, name})
	}
	o.bus.w32(slot+4, addr)
	c.R[0] = addr
	o.ret()
}

// watchDisplay records a display binding: a render object being built on a
// served model buffer (r0 = the model pointer).
func (o *Oracle) watchDisplay() {
	if o.cur == nil {
		return
	}
	if s := o.servedBy(o.cpu.R[0]); s != nil {
		o.record(FileReq{ID: s.id, Name: s.name, Phase: o.phase, Kind: "display"})
	}
}

// nameFor resolves a FAT or archive-flagged file ID to its name and content,
// matching what the slot loader's branches produce (archive members
// decompressed; regular files LZ-decompressed when tagged, per $02017D84).
func (o *Oracle) nameFor(id int) (string, []byte) {
	if id >= 0x8000 {
		return o.fileFor(id)
	}
	name := o.fatName[id]
	if name == "" {
		name = fmt.Sprintf("fat:%d", id)
	}
	data := o.ls.rom.File(id)
	if len(data) > 4 && string(data[:4]) == "LZ77" {
		data = nds.Decompress(data[4:])
	}
	return name, data
}

// fileFor resolves an internal or archive-flagged file ID to its name and
// content, matching what the real loader would hand back (archive members
// decompressed, card files raw).
func (o *Oracle) fileFor(id int) (string, []byte) {
	if id >= 0x8000 {
		ref, ok := o.ls.ResolveArchiveID(id)
		if !ok {
			return fmt.Sprintf("archive:%#x", id), nil
		}
		data, err := o.ls.ArchiveMember(ref)
		if err != nil {
			return ref.Stem(), nil
		}
		return ref.Stem(), data
	}
	name := o.ls.InternalName(id)
	if name == "" {
		return fmt.Sprintf("file:%d", id), nil
	}
	data, err := readExtractedFile(o.ls.extDir, "files/"+strings.TrimPrefix(name, "/"))
	if err != nil {
		return name, nil
	}
	// LZ77-tagged card files come out of the loader decompressed ($02017D84);
	// mirror that here (the .kcl collision maps are tagged, .bmd/.bca are not)
	if len(data) > 4 && string(data[:4]) == "LZ77" {
		data = nds.Decompress(data[4:])
	}
	return name, data
}

// servedFile remembers which file a served buffer holds, so a model POINTER
// seen later (an actor referencing a preloaded resource) identifies its file.
type servedFile struct {
	lo, hi uint32
	id     int
	name   string
}

func (o *Oracle) serve(data []byte) uint32 {
	addr := (o.scratch + 31) &^ 31
	if addr+uint32(len(data)) >= oScratchE {
		return 0
	}
	for i, v := range data {
		o.bus.Write(addr+uint32(i), v)
	}
	o.scratch = addr + uint32(len(data))
	return addr
}

// servedBy finds the served buffer containing p (newest wins).
func (o *Oracle) servedBy(p uint32) *servedFile {
	for i := len(o.served) - 1; i >= 0; i-- {
		if s := &o.served[i]; p >= s.lo && p < s.hi {
			return s
		}
	}
	return nil
}

// trapAsyncReg emulates the queue push and remembers the node.
func (o *Oracle) trapAsyncReg() {
	c := o.cpu
	aux, cb, node := c.R[0], c.R[1], c.R[2]
	o.bus.w32(node, o.bus.r32(oAsyncHd))
	o.bus.w32(node+4, cb)
	o.bus.w32(node+8, aux)
	o.bus.w32(oAsyncHd, node)
	o.asyncQ = append(o.asyncQ, node)
	o.ret()
}

func (o *Oracle) note(f string, a ...any) {
	if o.cur != nil {
		o.cur.Notes = append(o.cur.Notes, fmt.Sprintf(f, a...))
	} else {
		o.trace(f, a...)
	}
}

func (o *Oracle) record(f FileReq) {
	if o.cur != nil {
		o.cur.Files = append(o.cur.Files, f)
	} else {
		o.initReq = append(o.initReq, f)
	}
}

func (o *Oracle) cstr(p uint32) string {
	var b []byte
	for i := uint32(0); i < 256; i++ {
		ch := o.bus.Read(p + i)
		if ch == 0 {
			break
		}
		b = append(b, ch)
	}
	return string(b)
}

// swi services the BIOS calls the game issues (the crt0 and SDK use a handful).
func (o *Oracle) swi(c *arm.CPU, comment uint32) bool {
	n := comment & 0xFF
	if !c.Thumb && n == 0 {
		n = (comment >> 16) & 0xFF
	}
	switch n {
	case 0x03: // WaitByLoop
		return true
	case 0x04, 0x05, 0x06: // IntrWait / VBlankIntrWait / Halt: pretend it happened
		return true
	case 0x09: // Div
		num, den := int32(c.R[0]), int32(c.R[1])
		if den == 0 {
			c.R[0], c.R[1] = uint32(sign32(num)), uint32(num)
		} else {
			c.R[0], c.R[1] = uint32(num/den), uint32(num%den)
		}
		c.R[3] = c.R[0]
		if int32(c.R[3]) < 0 {
			c.R[3] = uint32(-int32(c.R[3]))
		}
		return true
	case 0x0B:
		o.cpuSet(false)
		return true
	case 0x0C:
		o.cpuSet(true)
		return true
	case 0x0D: // Sqrt
		v, r := uint64(c.R[0]), uint64(0)
		for bit := uint64(1) << 31; bit != 0; bit >>= 1 {
			if t := r | bit; t*t <= v {
				r = t
			}
		}
		c.R[0] = uint32(r)
		return true
	default:
		return true // benign: record-and-continue like bootoracle
	}
}

// cpuSet implements the CpuSet/CpuFastSet BIOS copy/fill.
func (o *Oracle) cpuSet(fast bool) {
	c, b := o.cpu, o.bus
	src, dst, ctrl := c.R[0], c.R[1], c.R[2]
	fill := ctrl&(1<<24) != 0
	n := ctrl & 0x1FFFFF
	if fast || ctrl&(1<<26) != 0 {
		var v uint32
		if fill {
			v = b.r32(src)
		}
		for i := uint32(0); i < n; i++ {
			if !fill {
				v = b.r32(src + i*4)
			}
			b.w32(dst+i*4, v)
		}
		return
	}
	var v uint32
	if fill {
		v = uint32(b.Read(src)) | uint32(b.Read(src+1))<<8
	}
	for i := uint32(0); i < n; i++ {
		if !fill {
			v = uint32(b.Read(src+i*2)) | uint32(b.Read(src+i*2+1))<<8
		}
		b.Write(dst+i*2, byte(v))
		b.Write(dst+i*2+1, byte(v>>8))
	}
}

func sign32(v int32) int32 {
	if v < 0 {
		return -1
	}
	return 1
}

// call runs fn(a0..a3) on the CPU until it returns, a trap notwithstanding.
func (o *Oracle) call(fn uint32, a0, a1, a2, a3 uint32, budget int) (uint32, string) {
	c := o.cpu
	c.R[0], c.R[1], c.R[2], c.R[3] = a0, a1, a2, a3
	c.R[13] = o.callSP
	c.R[14] = oRetAddr
	c.R[15] = fn &^ 1
	c.Thumb = fn&1 != 0
	for i := 0; i < budget; i++ {
		pc := c.R[15]
		if pc == oRetAddr {
			return c.R[0], ""
		}
		if t, ok := o.traps[pc]; ok {
			t(o)
			continue
		}
		if w, ok := o.watch[pc]; ok {
			w(o)
		}
		c.Step()
		if c.Halted {
			reason := c.HaltReason
			c.Halted = false
			return c.R[0], fmt.Sprintf("halt@%08X: %s", c.R[15], reason)
		}
	}
	var stack []string
	for i := uint32(0); i < 12; i++ {
		stack = append(stack, fmt.Sprintf("%08X", o.bus.r32(c.R[13]+i*4)))
	}
	return c.R[0], fmt.Sprintf("budget(%d)@%08X lr=%08X sp=%08X [%s]",
		budget, c.R[15], c.R[14], c.R[13], strings.Join(stack, " "))
}

// LoadOverlay copies a decompressed overlay to its table address, zeroes its
// BSS and runs its static-initializer list (NULL entries skipped, as
// FS_StartOverlay does). Overlay 0's initializer is its base function.
func (o *Oracle) LoadOverlay(id int) error {
	ov, ok := o.ls.ovls[id]
	if !ok {
		return fmt.Errorf("overlay %d not in table", id)
	}
	data, err := o.ls.overlayData(id)
	if err != nil {
		return err
	}
	o.phase = fmt.Sprintf("ovl%d-init", id)
	for i, v := range data {
		o.bus.Write(ov.RAMAddr+uint32(i), v)
	}
	for i := uint32(0); i < ov.BSSSize; i++ {
		o.bus.Write(ov.RAMAddr+uint32(len(data))+i, 0)
	}
	ctors := []uint32{}
	for p := ov.StaticInitStart; p < ov.StaticInitEnd; p += 4 {
		if fn := o.bus.r32(p); fn != 0 {
			ctors = append(ctors, fn)
		}
	}
	if id == 0 {
		ctors = append(ctors, oOvl0Init)
	}
	for _, fn := range ctors {
		if _, note := o.call(fn, 0, 0, 0, 0, 8_000_000); note != "" {
			o.trace("ovl%d ctor %08X: %s", id, fn, note)
		}
	}
	// ctors queue async loads of shared models (e.g. ovl2's first ctor at
	// $02100560 registers the coin/star/item slots and queues their loads); the
	// game's frame pump would service them, so drain here
	for _, note := range o.drainAsync() {
		o.trace("ovl%d async: %s", id, note)
	}
	return nil
}

// InitEngine runs main() up to (not into) the frame loop — OS arenas, root
// heap, engine init, all in the game's own order — then loads overlay 0 (the
// file-table init; replaced afterwards — its range is overlay 1's) and the
// resident engine pair: overlay 1 ($020AA420) and overlay 2 ($020AD660), and
// snapshots the result.
func (o *Oracle) InitEngine() error {
	o.phase = "main-init"
	o.scratch = oScratch
	c := o.cpu
	c.R[13] = o.callSP
	c.R[14] = oRetAddr
	c.R[15] = oMain
	const budget = 40_000_000
	i := 0
	for ; i < budget && c.R[15] != oGameLoop && c.R[15] != oRetAddr; i++ {
		if t, ok := o.traps[c.R[15]]; ok {
			t(o)
			continue
		}
		c.Step()
		if c.Halted {
			return fmt.Errorf("main-init halted at %08X: %s", c.R[15], c.HaltReason)
		}
	}
	if c.R[15] != oGameLoop {
		return fmt.Errorf("main-init: frame loop not reached (pc=%08X after %d instrs)", c.R[15], i)
	}
	o.trace("main-init: reached the frame-loop call after %d instrs", i)
	for _, id := range []int{0, 1, 2} {
		if err := o.LoadOverlay(id); err != nil {
			return err
		}
	}
	// The current-heap global still points at the fully-carved root heap; only a
	// scene start would re-point it. Give it a scene-style child of the main
	// heap, made by the game's own heap creator.
	// The current-heap global points at the root heap after main-init; a scene
	// start would re-point it at a scene heap. Make one with the game's own
	// creator (the root has ~1.25MB spare after the main and sound heaps).
	scene, note := o.call(oHeapNew, 0x100000, 0, 4, 0, 1_000_000)
	if scene != 0 && note == "" {
		o.bus.w32(oCurHeap, scene)
	} else {
		o.trace("scene heap create failed (%08X %s); keeping the root as current", scene, note)
	}
	for _, f := range o.initReq {
		o.pool[f.Stem()] = true
	}
	o.engineSnap = append([]byte(nil), o.bus.mem...)
	o.engineCPU = *o.cpu
	o.engScratch = o.scratch
	o.engServedN = len(o.served)
	o.bus.track = true
	o.clearDirty()
	return nil
}

func (o *Oracle) clearDirty() {
	for _, p := range o.bus.touched {
		o.bus.dirty[p] = false
	}
	o.bus.touched = o.bus.touched[:0]
}

// LoadConfig restores the engine state and loads one extra overlay (a level or
// enemy bank; -1 for none), then marks the config baseline for actor runs.
func (o *Oracle) LoadConfig(ovl int) error {
	// restore full engine state
	copy(o.bus.mem, o.engineSnap)
	*o.cpu = o.engineCPU
	o.clearDirty()
	o.scratch = o.engScratch
	o.served = o.served[:o.engServedN]
	o.asyncQ = nil
	if ovl >= 0 {
		if err := o.LoadOverlay(ovl); err != nil {
			return err
		}
	}
	// snapshot the pages the config init touched
	o.cfgPages = map[uint32][]byte{}
	for _, p := range o.bus.touched {
		pg := make([]byte, 1<<pageBits)
		copy(pg, o.bus.mem[p<<pageBits:])
		o.cfgPages[p] = pg
	}
	o.clearDirty()
	o.cfgScratch = o.scratch
	o.cfgServedN = len(o.served)
	return nil
}

// restoreConfig undoes one actor run: every dirtied page reverts to the
// config baseline (config-touched pages) or the engine snapshot.
func (o *Oracle) restoreConfig() {
	for _, p := range o.bus.touched {
		if pg, ok := o.cfgPages[p]; ok {
			copy(o.bus.mem[p<<pageBits:(p+1)<<pageBits], pg)
		} else {
			copy(o.bus.mem[p<<pageBits:(p+1)<<pageBits], o.engineSnap[p<<pageBits:(p+1)<<pageBits])
		}
		o.bus.dirty[p] = false
	}
	o.bus.touched = o.bus.touched[:0]
	o.scratch = o.cfgScratch
	o.served = o.served[:o.cfgServedN]
	o.asyncQ = nil
}

// Profile returns the actor's create function under the currently loaded
// config, validating that the profile pointer lands in loaded memory and that
// the profile's +4 actor ID matches. Loaded means: ARM9 static, overlay 2, or
// the config overlay's range.
func (o *Oracle) Profile(actor, cfg int) (create uint32, ok bool) {
	if actor < 0 || actor >= oProfMax {
		return 0, false
	}
	pp := o.bus.r32(oProfArr + uint32(actor)*4)
	if !o.loaded(pp, cfg) || !o.loaded(pp+4, cfg) {
		return 0, false
	}
	if int(o.bus.r32(pp+4)&0xFFFF) != actor {
		return 0, false
	}
	create = o.bus.r32(pp)
	if !o.loaded(create, cfg) {
		return 0, false
	}
	return create, true
}

// loaded reports whether an address falls in ARM9 static+heap space or in a
// currently-loaded overlay (2 or the config overlay).
func (o *Oracle) loaded(p uint32, cfg int) bool {
	if p >= 0x02004000 && p < o.ls.ovls[2].RAMAddr {
		return true
	}
	for _, id := range []int{2, cfg} {
		if id < 0 {
			continue
		}
		if ov, ok := o.ls.ovls[id]; ok && p >= ov.RAMAddr && p < ov.RAMAddr+ov.RAMSize+ov.BSSSize {
			return true
		}
	}
	return false
}

// RunActor spawns one actor the way the factory at $02043098 does: store the
// spawn context via the factory's own helper, call the profile's create
// function, then the new object's init (vtable slot 0), then drain the async
// load queue like the pump at $02072F44. State is rolled back afterwards.
func (o *Oracle) RunActor(actor, cfg int, par [3]int) *ActorRun {
	run := &ActorRun{Actor: actor, Config: cfg, Params: par}
	create, ok := o.Profile(actor, cfg)
	if !ok {
		run.Notes = append(run.Notes, "no profile under this config")
		return run
	}
	run.Create = create
	o.cur = run
	defer func() { o.cur = nil; o.restoreConfig() }()

	// spawn context: the factory stores (actor, link, param, layer-byte) through
	// $02043180 before calling create; par1 is the u32 param word.
	o.phase = "create"
	o.call(oSpawnCtx, uint32(actor), 0, uint32(par[0])|uint32(par[1])<<16, uint32(par[2]), 10_000)
	obj, note := o.call(create, create, 0, 0, 0, 3_000_000)
	if note != "" {
		run.Notes = append(run.Notes, "create: "+note)
	}
	run.Obj = obj
	if obj != 0 && obj < busSize {
		vt := o.bus.r32(obj)
		if init := o.bus.r32(vt); o.loaded(vt, cfg) && (o.loaded(init, cfg) || init >= 0x01FF8000 && init < 0x02000000) {
			o.phase = "init"
			if _, note := o.call(init, obj, 0, 0, 0, 6_000_000); note != "" {
				run.Notes = append(run.Notes, "init: "+note)
			}
		} else {
			run.Notes = append(run.Notes, fmt.Sprintf("no vtable (obj=%08X vt=%08X)", obj, vt))
		}
	}
	// drain async loads the actor queued
	o.phase = "async"
	for _, note := range o.drainAsync() {
		run.Notes = append(run.Notes, "async: "+note)
	}
	run.Colliders = o.colliders()
	// platform actors register their collider on the first step, not in init:
	// when the actor loaded a .kcl but no collider is in the table yet, run
	// one step frame and read the table again
	if len(run.Colliders) == 0 && len(o.KCLs(run)) > 0 {
		o.StepActor(run)
	}
	return run
}

// StepActor runs one frame of an actor's step (vtable +$18) — some actors
// (platforms) register their collider on the first step, not in init. Returns
// the collider table contents afterwards.
func (o *Oracle) StepActor(run *ActorRun) {
	if run.Obj == 0 {
		return
	}
	vt := o.bus.r32(run.Obj)
	step := o.bus.r32(vt + 0x18)
	if step == 0 {
		return
	}
	o.phase = "step"
	o.cur = run // record files the step itself loads (lazy loaders)
	defer func() { o.cur = nil }()
	if _, note := o.call(step, run.Obj, 0, 0, 0, 3_000_000); note != "" {
		run.Notes = append(run.Notes, "step: "+note)
	}
	run.Colliders = o.colliders()
}

// dBgW collider vtables (Part VI): the class of a registered collider is its
// vtable pointer; Mbg classes carry the local->world MtxFx43 at +$134 and the
// SclY variant a local-Y fx12 scale at +$1C8.
const (
	colSlots    = 0x020A0C80 // the 24-slot collider table $02039184 fills
	vtKc        = 0x020993DC
	vtKcMbg     = 0x02099434
	vtKcMbgSclY = 0x02099490
)

// colliders reads back every collider the run registered: which .kcl it
// walks (the buffer at ctx+$20 identifies the served file) and its transform.
func (o *Oracle) colliders() []Collider {
	var out []Collider
	for i := uint32(0); i < 24; i++ {
		ctx := o.bus.r32(colSlots + i*4)
		if ctx == 0 {
			continue
		}
		c := Collider{}
		if s := o.servedBy(o.bus.r32(ctx + 0x20)); s != nil {
			c.KCL = strings.TrimSuffix(filepath.Base(s.name), ".kcl")
		}
		switch o.bus.r32(ctx) {
		case vtKc:
			c.Class = "Kc"
		case vtKcMbg, vtKcMbgSclY:
			c.Class = "KcMbg"
			for j := 0; j < 12; j++ {
				c.Mtx[j] = int32(o.bus.r32(ctx + 0x134 + uint32(j)*4))
			}
			if o.bus.r32(ctx) == vtKcMbgSclY {
				c.Class = "KcMbgSclY"
				c.ScaleY = int32(o.bus.r32(ctx + 0x1C8))
			}
		default:
			c.Class = fmt.Sprintf("vt%08X", o.bus.r32(ctx))
		}
		out = append(out, c)
	}
	return out
}

// drainAsync pops the async-load queue the way the game's pump at $02072F44
// does — {next, cb, aux}: call cb(aux, -1) — until it is empty.
func (o *Oracle) drainAsync() []string {
	var notes []string
	for guard := 0; guard < 512; guard++ {
		node := o.bus.r32(oAsyncHd)
		if node == 0 {
			break
		}
		o.bus.w32(oAsyncHd, o.bus.r32(node))
		cb, aux := o.bus.r32(node+4), o.bus.r32(node+8)
		if cb == 0 {
			continue
		}
		if _, note := o.call(cb, aux, ^uint32(0), 0, 0, 3_000_000); note != "" {
			notes = append(notes, note)
		}
		// Slot-load requests only announce the slot to the loader THREAD's
		// mailbox; the oracle has no threads, so do the thread's work: acquire
		// the slot (loads through the $02017C54 trap on first use).
		if cb == oSlotAnnA || cb == oSlotAnnB {
			if _, note := o.call(oSlotAcq, aux, 0, 0, 0, 3_000_000); note != "" {
				notes = append(notes, note)
			}
		}
	}
	return notes
}

// Models lists a run's model bindings, most specific first: resources OUTSIDE
// the engine's common preload pool (the ovl2 static-init preload set — the
// game's own definition of "shared by every stage") come before pool members,
// and within each group the actor's own loads precede references; request
// order is preserved inside each class. The first entry is the actor's model.
func (o *Oracle) Models(r *ActorRun) []string { return o.stems(r, "bmd") }

// Clips lists a run's animation bindings, in the same order.
func (o *Oracle) Clips(r *ActorRun) []string { return o.stems(r, "bca") }

// KCLs lists a run's collision-mesh bindings (the .kcl files the actor's own
// code loaded — platform/mechanism actors registering dBgW_Kc colliders).
func (o *Oracle) KCLs(r *ActorRun) []string { return o.stems(r, "kcl") }

func (o *Oracle) stems(r *ActorRun, kind string) []string {
	var out []string
	seen := map[string]bool{}
	// Building a render object on a resource ("display") is the strongest
	// binding signal; next come the actor's own loads of non-pool resources
	// (its private files); then everything else (slot touches, pool re-loads)
	// in request order.
	class := func(f FileReq) int {
		pool := o.pool[f.Stem()]
		switch {
		case f.Kind == "display" && !pool:
			return 0
		case f.Kind == "" && !pool:
			return 1
		case f.Kind == "display":
			return 2
		default:
			return 3
		}
	}
	for group := 0; group < 4; group++ {
		for _, f := range r.Files {
			if class(f) != group || seen[f.Stem()] || o.Classify(f) != kind {
				continue
			}
			seen[f.Stem()] = true
			out = append(out, f.Stem())
		}
	}
	return out
}

// Classify identifies a request's payload: "bmd", "bca", "kcl", or "" —
// by extension for named files, by content sniffing for archive members
// (.bca header: u16 bones, u16 frames, then four in-file offsets).
func (o *Oracle) Classify(f FileReq) string {
	switch {
	case strings.HasSuffix(f.Name, ".bmd"):
		return "bmd"
	case strings.HasSuffix(f.Name, ".bca"):
		return "bca"
	case strings.HasSuffix(f.Name, ".kcl"):
		return "kcl"
	case f.ID < 0x8000:
		return ""
	}
	if k, ok := o.arcKind[f.ID]; ok {
		return k
	}
	_, data := o.fileFor(f.ID)
	k := ""
	if PlausibleBMD(data) {
		k = "bmd"
	} else if len(data) >= 0x18 {
		n := len(data)
		off := [4]uint32{le.Uint32(data[8:]), le.Uint32(data[12:]), le.Uint32(data[16:]), le.Uint32(data[20:])}
		if off[0] >= 0x18 && off[0] <= off[1] && off[1] <= off[2] && off[2] <= off[3] && off[3] <= uint32(n) {
			k = "bca"
		}
	}
	o.arcKind[f.ID] = k
	return k
}

// InitRequests exposes the file requests made by overlay static initializers
// (level-overlay ctors register per-level model slots through the same loader).
func (o *Oracle) InitRequests() []FileReq { return o.initReq }

// ---- direct-drive helpers (kcltrace and other function-level probes) ----

// Call runs fn(a0..a3) on the CPU until it returns; the note is "" on a clean
// return (see call).
func (o *Oracle) Call(fn uint32, a0, a1, a2, a3 uint32, budget int) (uint32, string) {
	return o.call(fn, a0, a1, a2, a3, budget)
}

// R32/R16/W32 access the machine's memory (little-endian).
func (o *Oracle) R32(a uint32) uint32 { return o.bus.r32(a) }
func (o *Oracle) R16(a uint32) uint32 {
	return uint32(o.bus.Read(a)) | uint32(o.bus.Read(a+1))<<8
}
func (o *Oracle) W32(a, v uint32) { o.bus.w32(a, v) }

// ReadBytes copies n bytes out of the machine.
func (o *Oracle) ReadBytes(a, n uint32) []byte {
	out := make([]byte, n)
	for i := uint32(0); i < n; i++ {
		out[i] = o.bus.mem[a+i]
	}
	return out
}

// Alloc carves a zeroed block out of the scratch region.
func (o *Oracle) Alloc(n uint32) uint32 {
	addr := (o.scratch + 31) &^ 31
	for i := uint32(0); i < n; i++ {
		o.bus.Write(addr+i, 0)
	}
	o.scratch = addr + n
	return addr
}

// PC exposes the CPU's current program counter (during Call, the instruction
// a read watch fires under).
func (o *Oracle) PC() uint32 { return o.cpu.R[15] }

// SetReadWatch observes every byte read in [lo,hi); fn(addr) runs before the
// read is served. SetReadWatch(0,0,nil) clears it.
func (o *Oracle) SetReadWatch(lo, hi uint32, fn func(addr uint32)) {
	o.bus.watchLo, o.bus.watchHi, o.bus.watchFn = lo, hi, fn
}

// LastServed reports the most recently served file buffer.
func (o *Oracle) LastServed() (lo, hi uint32, name string) {
	if len(o.served) == 0 {
		return 0, 0, ""
	}
	s := o.served[len(o.served)-1]
	return s.lo, s.hi, s.name
}
