# Luigi's Mansion (GameCube) — technical reference

**Image:** `Luigi's Mansion (USA).iso` — 1,459,978,240 bytes, MD5 `6e3d9ae0ed2fbd2f77fa1ca09a60c494`.
Not committed (copyright); supply your own copy.

Luigi's Mansion (Nintendo EAD, 2001) was the GameCube's launch showpiece: a fully lit, fully
skinned 3D game whose opening cutscene plays in real time on the console's Flipper GPU. This
document reconstructs the shipped disc from its bytes alone: no third-party emulator, debugger
or disassembler, no released source, and nothing about the file or instruction formats taken
from external documentation. Everything is derived from the image with purpose-built tools —
above all the repository's own GameCube machine (`tools/platform/gc`, a Gekko CPU core with an
LLE Flipper pipeline and real DSP), whose boot oracle runs the game and, when asked the right
questions, explains it.

The reverse-engineering method throughout is the one the oracle makes possible: *let the game
name its own data*. `-dvd` logs every disc read with the file it lands in; `-rwatch` on a buffer
the drive just filled catches the PC of the code that parses it; `-logpc` on that code yields the
source and destination pointers of every call; and a `-dump` of the game's own RAM is the
ground truth every reimplementation is verified against.

## Contents

* **Part I** — the disc: the GameCube's headerless disc format, the apploader, the DOL
  executable, and the **FST** filesystem with its development-era tree.
* **Part II** — the opening cutscene's file chain: which archives the demo player streams,
  in what order, and what is inside them.
* **Part III** — the **Yay0** compression and the **RARC** archive container.
* **Part IV** — the **.mdl** model format: GX display lists, the node tree, packets and the
  GPU matrix slots, envelope skinning, materials and tiled textures.
* **Part V** — the **.key** animation format: nine hermite channels per node, the s16 angle
  keys, and the 30 fps timeline.
* **Part VI** — the exporters (`lmtool`) and the Studio integration.
* **Part VII** — the **.scd** script database and **.sco** cut records: the cutscene's own
  camera and lights.
* **Part VIII** — the **.sls/.slk** vertex animation: Luigi's face.
* **Open** — how the demo binds shot names to archive members; the `.txp` texture-pattern
  files; the mansion's room archives (`Iwamoto/map*/room_*.arc`); gameplay and audio.

---

## Part I — the disc

The disc has no ISO 9660 filesystem. A 0x440-byte header names the game (`GLME01`,
"LUIGI'S MANSION") and points at three structures:

| what | where | notes |
|---|---|---|
| apploader | `0x2440` | 78,060 bytes, dated 2001/08/09, entry `0x81200194` |
| DOL executable | `0x15600` | 5,089,344 bytes, 10 segments, entry `0x80003100` |
| FST | `0x3B5100` | 20,857 bytes, 912 entries / 847 files, loads at `0x803FAE80` |

The **FST** is one flat array of 12-byte entries followed by a string table. A directory entry
carries the index of the first entry *outside* it, so the hierarchy is arithmetic rather than
pointers: a directory's children are simply the entries between it and that bound. Boot is the
console's IPL running the apploader, which streams the DOL's segments into RAM and returns the
entry point; the game then reloads the FST for itself and resolves every file access to a raw
byte offset on the disc — which is why the oracle's `-dvd` flag can name every read.

The file tree still wears its development layout: directories named after the game's staff
(`Iwamoto/` holds the mansion's room archives, `Kawano/` the menu resources, `Ajioka/` the
cutscenes), and **every directory still contains its CVS bookkeeping files** (`Entries`,
`Repository`, `Root`) — the version-control metadata of 2001, burned onto every retail disc.

## Part II — the opening cutscene's file chain

The opening — the mansion reveal, Luigi's walk through the dark forest, the gate — is not
video. It is played in real time by a demo system that streams one archive per shot from
`Ajioka/ADemo/`. Watching the boot with `-dvd` while the cutscene runs gives the shot order in
the game's own words:

```
opwf.szp   the forest walk        (wf — "walk forest")
oppm.szp   the mansion reveal     (pm — the pointing map hand / mansion)
opeg.szp   entering the gate      (eg — "enter gate")
opcn.szp → opsu.szp → opdn.szp → opod.szp    (the remaining shots)
```

Each archive holds everything its shot needs. `opwf.szp` contains:

```
opwf_bg.mdl        788,620   the forest set (trees, path, ground, sky dome)
opwf_luigi.mdl     246,314   Luigi (122 nodes, envelope-skinned)
opwf_luigi.key     129,112   his 338-frame walk performance
opwf_cone.mdl / opwf_handlight.mdl   the flashlight cone and glow
opwf_bg.key / *_cone.key / *_handlight.key   the sets' own (mostly constant) animations
opwf.scd           the shot's demo script          (format not yet reversed)
walkforest.sco     the shot's camera               (format not yet reversed)
opwf_luigi.slk/.sls                               (not yet reversed)
```

`entergate.mdl` in `opeg.szp` is the same 122-node Luigi with a different clip
(`entergate.key`); `oppm.szp` carries the mansion set (`op_mansion.mdl`, 1.87 MB), the gate
torch, the lightning bolt, and the giant pointing hand of the map sequence.

The demo player loads a shot's szp into a staging buffer (physical `0xC4FF60` in the observed
runs), decompresses it, and patches the parsed models in place; the previous shot's memory is
recycled. The model object the engine keeps per actor carries, at `+0x04` the mdl pointer, at
`+0x08` the key pointer, at `+0x0C` the current **frame number as a float**, and at `+0x14` the
array of runtime world matrices the renderer consumes — the single most useful structure for
verifying everything below.

## Part III — Yay0 and RARC

A `.szp` is a **Yay0** stream wrapping a **RARC** archive.

**Yay0** (magic `Yay0`): four big-endian header words — magic, decompressed size, the offset of
the *link table*, the offset of the *literal bytes*. The control stream starts at `+0x10`: each
32-bit word gives 32 flags, MSB first. A set bit copies one literal byte; a clear bit reads one
16-bit link — count in the top nibble, distance in the low 12 bits — and copies `count` bytes
from `dst - distance - 1`. A zero count nibble fetches the real count from the literal stream
plus 18. The three streams (flags, links, literals) advance independently, which is the format's
whole trick. The decompressor lives at `0x800071D0` in the DOL (found by read-watching the
buffer the drive had just filled); the reimplementation in `extract/lm/yay0.go` was verified
**byte-exact** against the game's own output in RAM for the full 2.6 MB `oppm` stream.

**RARC**: a 0x20-byte outer header (magic, total size, header size, data offset), an inner
header with directory-node and file-entry counts and offsets, then 0x10-byte directory nodes,
0x14-byte file entries and a string table. A file entry whose type halfword has bit `0x0200`
set is a subdirectory reference; everything else carries data offset and size. Parser:
`extract/lm/rarc.go`.

## Part IV — the .mdl model format

Read out of the game's renderer at `0x80058A00..0x8005A100` (found by read-watching the parsed
model while the cutscene drew it). The file is a header of counts and section offsets over raw
**GX display lists** — the Flipper GPU's native command stream, stored ready to feed the
write-gather FIFO. At load the engine patches every section offset into a pointer in place; on
disc they are plain file offsets:

```
+0x04 u16 faceCount            +0x08 u16 nodeCount      +0x0c u16 envelopeCount
+0x10 u16 posCount             +0x12 u16 nrmCount
+0x14 u16 clrCount             +0x16 u16 texCount
+0x20 u16 texHdrCount          +0x24 u16 samplerCount   +0x28 u16 materialCount
+0x2a u16 drawPairCount
+0x30 → nodes     16 B: {s16 id, s16 child, s16 sibling, u16 mode,
                         u16 pairCount, u16 firstPair}   (child/sibling relative)
+0x34 → packets   32 B: {u32 dlOffset, u32 dlSize, u16 -, u16 mtxCount, s16 mtxIdx[10]}
+0x38 → matrices  48 B: 3x4 float32 row-major, one per node — INVERSE BIND matrices
+0x3c → envelope weights (f32)   +0x40 → envelope joints (u16)   +0x44 → counts (u8)
+0x48 → positions f32[3]         +0x4c → normals f32[3]
+0x50 → colours rgba8            +0x54 → texcoords f32[2]
+0x60 → texture-header pointer table (u32 each)
+0x68 → materials 0x120 B        +0x6c → samplers 8 B: {u16 texIdx, u16 -, u8 wrapS, u8 wrapT}
+0x70 → shapes    8 B: {u32 flags, u16 packetCount, u16 firstPacket}
+0x74 → draw pairs {u16 materialIdx, u16 shapeIdx}
```

**The scene graph.** Nodes form a tree by *relative* child (`+2`) and sibling (`+4`) counts —
pinned by composing the animated locals down each candidate hierarchy and matching the game's
runtime world-matrix array (the wrong assignment leaves hundreds of units of error; the right
one agrees to the sine-table quantisation, max 0.06). Each node lists (material, shape) pairs
to draw.

**Packets and matrix slots.** A shape is a run of 32-byte packets: a display-list extent plus a
table of up to ten matrix indices. Entry *i* is loaded into GX PN-matrix slot *3i*; an index of
`-1` keeps whatever an earlier packet loaded — **the slots persist across packets**, so a
correct reader walks the packets in draw order carrying slot state. An index below `nodeCount`
names a node's world matrix; an index at or above it names a blended **envelope** matrix.

**Display lists.** A GX command stream: opcode byte (`0x80` quads, `0x90` triangles, `0x98`
strip, `0xA0` fan, OR'd with the vertex-format index), u16 vertex count, then per vertex
**three direct matrix bytes** — the PN slot plus two texture-matrix slots (GX attributes 7/8 on
static models, 1/2 on skinned ones; visible as `vcd=1F81` in the oracle's draw trace) — followed
by one u16 index per enabled attribute in position, normal (three indices when the shape's
`0x02000000` NBT flag is set), colour, texcoord order. An attribute is enabled when its array
count is nonzero.

**Skinning.** The per-node file matrices are **inverse binds** (world → joint). This was proved
against the game's runtime array: `world[i] == inverse(file[i])` exactly, for the forest set's
every node — and the envelope formula was confirmed to five decimals:

```
envelope[e] = Σk  weight[k] · world[joint[k]] · invBind[joint[k]]
```

Rigid vertices are stored in their joint's local frame; envelope-skinned vertices in bind
(model) space — the same split the matrix slots express at draw time. Luigi has 122 nodes and
64 envelopes (58 of two joints, up to one of five). Even the forest's trees are "joints": each
gnarled tree is stamped into place by its node matrix, so instancing falls out of the skeleton.

**Textures.** A texture header is `{u8 formatIndex, u8 -, u16 width, u16 height}` with the
image at `+0x20`. The one-byte format index maps through the game's table at `0x80338C30` to
the GX format: `0→CI4 1→CI8 2→CI14X2 3→I4 4→I8 5→IA4 6→IA8 7→RGB565 8→RGB5A3 9→RGBA8 10→CMPR`.
The opening models use CMPR (8×8 tiles of four DXT1 blocks), RGB565 and IA8; decoders in
`extract/lm/gxtex.go`. A material (0x120 bytes) carries up to eight 32-byte texture stages,
each naming a sampler; samplers bind a texture with GX wrap modes.

**Material render state.** The record opens with `{u8 rgba[4] tint @0, u8 mode @6, u8
stageCount @7}`. The tint is loaded into the GX material colour register and the TEV
multiplies it into the texture's colour *and* alpha — draw-traced on the forest walk
(`RR_GC_DRAWTRACE=1` from `cut1.state`): the cone's lit vertices print exactly the file's
bytes (`FFE5A47F` — a warm beam at alpha 127). The mode byte selects the pass: `0` opaque,
`1` alpha-tested cutout (the forest foliage), `2` translucent — alpha blend
(`bp41=00F4AD`, srcalpha/invsrcalpha) with Z compare on but **Z write off** (`zm=07` vs the
opaque pass's `17`). Mode 2 marks both flashlight-cone materials, the torch flames, the
mansion set's window glass and one all-invisible Luigi material (tint alpha 0 — its draws
write zero pixels on hardware too). The export maps mode 2 to glTF `BLEND` plus
`extras.blend:"alpha"` (the Studio turns off depth writes), everything else to `MASK`.
**Culling.** Each packet's u16 at +8 is its GX SDK cull mode, read by the renderer at
`0x80059570` and passed to `GXSetCullMode` (the write-profiler climb from the GEN_MODE
trace bits found the load; the SDK call swaps front/back into the register encoding).
Almost everything culls: the sets ask for back, the characters for front (their bind
matrices carry the opposite winding — the same det < 0 information the exporter already
uses to flip mirror-stamped triangles), and the only cull-none packets are the sets' sky
domes and backdrops, which must render from inside. The export honours the field:
materials ship single-sided (strip order reversed — GX's front face is the opposite
winding from glTF's CCW; mirror-stamped triangles keep their order) unless one of their
packets is cull-none. This is what lets the opod camera sit inside the doorway and film
the door opening through the hill's culled back faces — double-sided export filled the
whole frame with the hillside's underside — and it also removes the flashlight cone's far
shell, which hardware never draws.

**Texture sampling.** The `.mdl` and `.bin` textures ship without mip chains, so the
hardware always samples mip 0. The Studio's earlier `linear` (bilinear + generated
mipmaps) profile banded the flashlight cone into rings: the beam texture is an 8×128
gradient whose u tiles ~16× around the cone, and the isotropic mip level that extreme
u-minification selects collapses the 128-row v-gradient into a handful of steps. The
manifest now says `bilinear` — bilinear with no mipmaps — which reproduces the console's
smooth gradient (and its honest texel shimmer).

Still open: the `.clr` sidecars are keyframed material-colour tracks (the oppm trace shows
the lightning's register riding them, alpha 0 at rest and flashing to 255), and `.txp` is a
texture-pattern flipbook (the torch fire) — neither is decoded yet, so those materials ship
at their rest colour.

## Part V — the .key animation format

Read out of the evaluator at `0x8005AF0C` and its two interpolators (`0x8005AB04` floats,
`0x8005AC34` angles):

```
+0x02 u16 trackCount (one per node)     +0x04 u16 frameCount
+0x08 u32 flags (bit 1: scale channels are honoured — SRT instead of RT matrices)
+0x0c → float pool (scale)   +0x10 → s16 pool (rotation)   +0x14 → float pool (translation)
+0x18 → per-track 9 × u32 channel data offsets
+0x1c → per-track 9 × u16 channel kinds
```

The nine channels are `sx sy sz rx ry rz tx ty tz`. A kind of 0 leaves the default (1 for
scale, 0 otherwise); 1 reads one constant; *n* > 1 is *n* keyframes. The kind's top bit selects
4-field keys (time, value, tangent-in, tangent-out) over 3-field ones. Interpolation is **cubic
hermite** with key times in frames on a **30 fps** timeline (the arithmetic rescales them by
1/30 so tangents are per-second). Rotation keys are s16 values in **360/4096-degree units**,
interpolated in degrees, then converted to the console's 65536-per-turn integer angles and
looked up in its **4096-entry sine table**; a constant rotation is simply `value << 4`. The
local matrix is `Rz·Ry·Rx` with the translation in the fourth column.

The walk (`opwf_luigi.key`) is 338 frames — eleven seconds from the first timid step to the
double-take — and drives all 122 nodes; the root track carries Luigi's path through the set.
The evaluator's per-node key-index cache (the model object's `+0x28` array) is an optimisation,
not a semantic: evaluating cold gives the same pose.

The demo places each actor through one more matrix (for the walk shot: a mirror with a
translation) between the key pose and the world — visible as the constant left-multiplier `A`
when solving `ram = A · composed` over all nodes.

**`A` solved — the placement gap is closed.** `extract/cmd/actorsolve` recovers each
actor's `A` from a savestate: every demo actor is a heap object (frame-counter float,
four ascending stage-array pointers, an `FFFFFFFF` terminator) whose stage-0 array holds
the node WORLD matrices, so `A = world[i]·composed[i]⁻¹` over all nodes, against our own
`.key` evaluation (residual ≤ the sine-table quantisation; the world arrays lag the
counter by ~2 frames; a `-state2` a few fields later separates live objects from a
previous shot's parked ones). Stage 1 is the same matrices premultiplied by the running
camera's view — chasing it produced a phantom "time-varying blocking" until its solve was
matched against the baked `.scd` camera (`A_s1(f) = View(f)·A_s0`). The answers, committed
as `cmd/webexport/blocking/<shot>.json` and verified across states ~200 frames apart:

- The demo world is y-up and almost every actor's `A` is the **identity** — the sets,
  torch, lightning, crow, the map hand, and the opsu/opod Luigis are keyed straight in
  world space where the cameras look.
- `opwf_luigi.key` and `entergate.key` (whose root tracks carry a 180° X rotation) get a
  constant **point reflection through the key's own frame-0 root position**:
  `A = −I + 2·p₀` — exactly `(0,0,−2420)` for the walk and `(60.6, 31.9, −15525.6)` at
  the gate, i.e. twice `(0,0,−1210)` and `(30.3, 15.9, −7762.8)`. "A mirror with a
  translation", now with its numbers.
- The flashlight and its cone are **attached**: `cmd/attachprobe` shows the handlight's
  world node 0 equals a Luigi hand joint's world matrix times a constant offset (joint 35,
  identity, in `opwf_luigi.mdl`; joint 53 with a fixed offset in `entergate.mdl`), and the
  cone's equals the handlight's joint 3 exactly. The exporter expands these rides into
  per-frame wrapper channels **derived from the keys** — the placement mode per actor is
  the only traced fact; every matrix in the shipped GLBs comes from the disc's own data.

The `.scd` camera roll channel is **degrees** (like fov; the door-knob shot straightens
from −12.7° to −0.7°) — the Studio player once read it as radians and spun the shot
through two full revolutions.

**The rest of the demo library.** `/Ajioka/ADemo/` holds 78 archives, all on the same
`.scd`/`.sco` + `.mdl`/`.key` machinery: the opening's seven shots; **56 door-unlock
vignettes** (`co/no/ho/oo-demo01..14` — Luigi's gloved hand `hr_mdl` at fourteen door
sets, the `ho/oo` variants adding the key `k_mdl`); **`dodb`** ("drbyebye") — Professor
E. Gadd's farewell at his lab shack, whose `db_bg` set turns out to be the whole mansion
approach plus the lab, with Gadd, Luigi, the Poltergust (`db_sojiki`) and the flashlight
pair; **`dotb`** — the thunderbolt beat over the same set; and the Game Boy Horror scenes
(`gameboy`, `gbdemo01..07`). The lab farewell and the thunderbolt ship as Studio cutscene
levels (`lab-byebye`, `thunderbolt`): every actor in them is world-keyed (Gadd's blocking
rides non-root joints, as the key demos taught earlier), so identity placement reproduces
the scenes — the 460-frame farewell runs from Gadd at his shack up the path to a
frightened Luigi shutting the front door.

The door vignettes and the Game Boy Horror ship as two more levels — 64 shots over just
twelve GLBs, because the archives dedup hard by content (verified by hashing every member
of all 56 door archives): one glove-hand model, one key model, eight unique door sets
behind the fourteen `dNN` names (01=02=12, 03=11, 04=06, 09=13, 10=14), and swing/hand/key
clips shared by content across doors within a variant (only the d03/d11 "normal open" pair
differs). `SkinnedGLBMulti` bakes any number of `.key` clips into one GLB as named glTF
animations, and each shot's script entry picks its clip — `door-demos` plays all four
variants (`cnop` "can't open", `noop`, `hkop`/`osop` with the key) against each door, and
`gbh-demos` plays the handheld's vignette plus the seven screen scenes (E. Gadd calling on
the Game Boy Horror's display) off one model with seven clips.

## Part VI — exporters and the Studio

`extract/cmd/lmtool` drives everything:

```
lmtool -image DISC.iso -ls  /Ajioka/ADemo/opwf.szp            list a RARC
lmtool -image DISC.iso -x   /Ajioka/ADemo/opwf.szp -into d/   extract it
lmtool -image DISC.iso -mdl /Ajioka/ADemo/opwf.szp:opwf_bg.mdl -o forest.glb
lmtool -image DISC.iso -mdl /Ajioka/ADemo/opwf.szp:opwf_luigi.mdl \
       -anim opwf_luigi.key [-inplace] -o luigi-walk.glb
lmtool -image DISC.iso -mdl ... -anim ... -pose 196           print pose matrices (RAM diffing)
```

The static path bakes each node's geometry through `inverse(invBind)` — the bind-pose world — and
groups triangles by material. The skinned path emits the node tree as a glTF joint hierarchy,
the file matrices as `inverseBindMatrices`, converts rigid vertices to bind space, and samples
the clip at the native 30 fps into LINEAR translation/rotation tracks (`-inplace` zeroes the
root's x/z for a viewer-friendly walk-in-place export). The `verify-the-shipped-file` rule
applies: every export was opened and rendered before shipping, which is how a mirrored forest
(bake through the matrix instead of its inverse) and a culled-away Luigi (bind bounds far from
the posed skeleton) were caught.

The Studio page (`site/src/studio/main.js`, id `luigis-mansion-gc`) ships the two sets through
an `lm-set` renderer that opens at each set's establishing camera inside its own sky dome, and
the animated actors through `lm-actor`, which frames the *posed skeleton* rather than the
bind-space geometry and disables frustum culling (three.js culls skinned meshes by their bind
bounds). Assets live in `site/public/luigis-mansion-gc/`.

The formats cover the **whole demo library**, and the Studio carries all of it: the opening's
seven shots (the forest walk plus six shots on the shared mansion set — the map, the crow, the
gate, the steps, the door knob, the door — each flying its own `.sco` camera), the fourteen
**key demos** (four `co/ho/no/oo` variants share eight unique door vignettes — `dNN_mdl` sets
posed by their own `.key`s — plus Luigi's gloved hand `hr_mdl` and the ornate key `k_mdl`), the
lab farewell `drbyebye` (E. Gadd `db_lohakase`, another Luigi, the **Poltergust** `db_sojiki`
with its 110 KB of vertex-animation weights, the lab set), the `thunderbolt` shot and the
**Game Boy Horror** vignette. One wrinkle the exports honour: model spaces are not one-handed —
the opening's actors are y-down (the demo's placement matrix flips them) while the lab scene,
the door sets and the Game Boy Horror are y-up, hence `lmtool`'s `-noflip`.

## Part VII — the .scd script database and .sco cut records

Read out of the camera evaluator at `0x801156FC` and the channel interpolators at
`0x800F1C58` (float) and `0x800F1E1C` (s16). The two files split a shot's direction between a
**channel database** (.scd) and **per-cut base records** (.sco); at load the engine patches
the .scd's section offsets into pointers, exactly as it does for models.

**.scd** (for the forest walk, `opwf.scd`):

```
+0x00 u16 frameCount (0x151 = 337 for the walk)
+0x08 → camera:  f32 aspect, then 10 channel descriptors —
        pos xyz, target xyz, roll, fov, near, far
+0x0c → simple lights, 36 B each: {char name[16], desc rgb[3], u16}   ("sunn1")
+0x10 → keyed lights, 104 B each: {char name[16], u32, desc[14]}      ("grasss1", "lightt1", …)
+0x18 → float pool          +0x1c → s16 pool
```

A channel descriptor is six bytes — `{u16 count, u16 offset, u16 stride}` — into one of the
two pools. Stride 1 is a constant at `pool[offset]`; otherwise `count` keys of `stride` words:
(time, value, tangent), with stride 4 carrying separate in/out tangents. Evaluation is the
same **cubic hermite on the 30 fps frame timeline** as the `.key` format — the channel system
is one mechanism worn three ways (model animation, camera, lights).

**.sco**: a f32, then the camera cut record at `+4` — `{s16 cut, u16 flags, f32 base pos[3],
target[3], roll, aspect, fov, near, far}` — followed by light cut records
(`{s16 lightIdx, s16, s16 rgb[3]}`, clamped to 0..255 after the channel adds). Every channel
result is **added** to its cut base; in the forest walk the bases are zero and the channels
are absolute.

The result feeds the runtime camera struct at `0x803A3820` (position, target, an up vector
built from the roll, then fov/aspect/near/far), and the reimplementation reproduces it to
float precision at every probed frame: the walk's camera slides from `(-3384, 115, -787)` to
`(-2064, 115, -1847)` over 337 frames on four hermite keys per axis, looking 30 units ahead,
at a telephoto **19.1° fov** with the far plane at 327,680 units. `lmtool -camera
"/Ajioka/ADemo/opwf.szp:opwf.scd:walkforest.sco"` exports the track as JSON, one sample per
frame, in the same space as the shot's GLB set.

## Part VIII — the .sls/.slk vertex animation: Luigi's face

Read out of the evaluator at `0x8005C464` (found by read-watching the file in RAM). The
`.sls`/`.slk` pair is the characters' **blend-shape system**: groups of vertices — Luigi's
facial regions, all bound to the head joints 87–88 — whose entries in the model's *position
array* are rebuilt every frame: zeroed, accumulated as `weight × shape` for each active
shape, then divided by the summed weight. The model object holds **eight overlay slots**
(`+0xB0`), so several of these can stack.

**.sls** — the shape geometry:

```
+0x08 u16 groupCount        +0x0c u16 posDictCount   +0x0e u16 nrmDictCount
+0x10 → group records 20 B: {u16 group, -, u16 chanBase, -[5], u16 shapeCount, u16 vertexCount}
+0x14 → shape entries  8 B: {u16 posCount, u16 posStreamOff, u16 nrmCount, u16 nrmOff}
+0x18 → position dictionary (f32[3])      +0x1c → normal dictionary (f32[3])
+0x20 → u16 shape streams   +0x28 → active-shape entry table   +0x2c → target vertex indices
```

A shape's stream holds one u16 per vertex: the low 13 bits index the shared vector
dictionary, the top three bits negate x/y/z — **mirror compression** for a symmetric face.
The walk Luigi's six groups are the mouth (174 vertices), the two eyes (41 each), the brow
region and two more, 422 vertices in all; their single shape differs from the position
array's stored neutral face by up to ~3 units — the shot's **scared expression**, which the
exporter now applies. The normal dictionary and per-shape normal counts exist in every file
but are only consumed when weights change (the walk's never do); their target indexing is
not yet traced.

**.slk** — the weight tracks: `{-, u16 frameCount, -, u16 channelCount}`, then a float pool,
per-channel offset and count tables (count 1 = constant) and a channel map — the same
hermite-on-30 fps-frames channel mechanism as `.key` and `.scd`, worn a third way. A group
with a **single** shape is hardwired to weight 1.0 and the `.slk` is never consulted — which
is why both opening Luigis ship all-constant `.slk`s. The gameplay Luigi (`model/luige.szp`)
shows the system in earnest: ten pairs — `lv0/lv1/lv2_face` (expressions by health level),
`breath01`, `biku_01/02` (the startle), `dam_f01/dam_b01` (damage) — whose groups carry up
to four alternative shapes with one-hot weight keys cross-fading mouth positions every few
frames: the talking flap.

## Part IX — the .bin room format: the mansion itself

The mansion's rooms (`Iwamoto/map2/room_00.arc`..`room_73.arc`, plain RARC archives) use a
**third model format**, `.bin` — one `room.bin` per room plus its furniture (`otukue.bin` the
tea table, `oisu.bin` the chair, `otoire.bin` the toilet, …) and `.anm` furniture animations.
Read out of the renderer at `0x8001D8C8..0x8001DD90` (all the `bl 0x801F9F94` callers): like
the `.mdl`, the file is its own runtime object — a version byte and a name, then **21 section
offsets at `+0x0C`** patched into pointers at load:

```
[0]  textures  12 B {u16 w, u16 h, u8 gxFormat, …, u32 dataOff}   (raw GX format ids;
                     dataOff is relative to the TEXTURE SECTION, not the file — the header
                     table and the pixel data share the section, so the first data offset
                     is also where the header table ends)
[1]  samplers  20 B {s16 texIdx, s16 palette, s8 wrapS, s8 wrapT, ..., u8 mip, s16 lod}
                     (wrap: 0 clamp, 1 repeat, 2 MIRROR — the foyer's floor medallion is
                     one quadrant mirrored both ways)
[2]  positions s16[3]   [3] normals f32[3]   [6] texcoords f32[2]   [7] texcoords1 f32[2]
[10] materials 40 B {…, u8 rgba[4]@3, s8 samplerIdx@9, …, s16 stageBlock@0x1a → [13]}
[11] meshes    24 B {u16, u16 dlSize/32, u32 attrMask, u16, u8, u8 nbt, u32 dlOff (section-rel)}
[12] graph    140 B {s16 parent,child,next,prev; u16 FLAGS@8 (0x80 = …); f32 scale/rot°/
                     trans @0xc/0x18/0x24; bbox+radius@0x30; u16 pairCount@0x4c; u32
                     pairs@0x50 — the pair count is @0x4c ONLY; reading @8 as a second
                     count duplicated other nodes' parts at identity (phantom window
                     frames floating in the foyer)}
```

Where the cutscene `.mdl` stores matrix-indexed fans over float arrays, the room `.bin`
stores **indexed strips over s16 positions** — no matrix bytes at all; the graph node's
composed TRS (rotation from *degrees*, through the same 65536-per-turn sine table) places
each part, and the part list is `{u16 material, u16 mesh}` pairs in two runs (the renderer's
opaque and translucent passes). A mesh's display list lives inside the mesh section; its
vertex layout comes from the mesh's **attribute mask, the u32 at +4** — the per-mesh setup at
`0x8001D5B8` reads it as literal `1<<GXAttr` bits: `0x200` POS, `0x400` NRM, `0x800` CLR0,
`0x2000` TEX0, `0x4000` TEX1 — one u16 index per present attribute, in attribute order. The
byte at +11 switches the normal to NBT3 (`GXSetVtxAttrFmt` cnt=2): *three* indices (normal,
binormal, tangent) into the normals array. A survey of all 572 `.bin` files on the disc finds
exactly five masks — 0x200, 0x600, 0x2200 (position+texcoord, *no* normal — 183 meshes an
earlier u16@+8 heuristic misread as position+normal), 0x2600, and 0x6600 with NBT in the six
two-stage room files (63/65/66/67, `b1_c_67`, `gyara_00`) — no vertex colours anywhere, and
no CI-format textures either, so neither needs an export path. The u16 at +8 the first decode
used is *not* the attribute word. Positions can also be f32 (an object flag at +0x48 bit 1
selects stride 12 over 6 in `0x8001D4C0`); every file on this disc uses s16.
`lmtool -bin "/Iwamoto/map2/room_02.arc:room.bin"` exports a room; the foyer comes out with
its staircase, webbed double door and panelled walls, and the Studio's "The mansion" section
carries the first rooms and furniture. Untextured materials (the foyer's banister cloth,
material 25 on the mask-0x600 mesh) are genuinely stage-less in the game too — white cloth
shaded by GX lighting — so the exporter leaves them lit rather than unlit. Acceptance:
the export from the game's own foyer camera (medallion → rosette arch, `foyer.state`)
matches the game frame surface for surface; the differences are the game's darkness-and-
flashlight lighting, the door leaf (a separate asset), and the furniture (separate bins).

Still open in this format: the `[13]/[14]` material-stage sections richer furniture uses
(and TEX1 lightmap blending for the six two-stage rooms), the `.anm` vertex animations, and
the sweep of all 74 rooms — see `PLAN.md` for the assembled-mansion roadmap.
`extract/cmd/binsurvey` re-runs the disc-wide survey. (A caution from this section's own history: the sampler records were first read at
stride 12, which textured the foyer's walls with window frames; the true stride, 20, was
recovered by noticing the texture-index fields landing at byte offsets 0, 20, 40, … in the
raw section. Two more decode bugs survived that fix and were caught the same way — by the
user looking at the export: texture data offsets were taken as file-absolute (every texture
decoded 0x60 bytes early, disproved because 25 headers occupy 0x0..0x18C of the section and
the first "offset" pointed inside them), and the graph word at +0x08 was read as a second
pair count when it is flags (node 3's 0x80 pulled 128 pairs from the following node —
including its upper-wall window frames — and drew them untransformed). Eyeballing an export
is not acceptance.)

## Part X — the placement database: furniture as instances

Furniture is *not* baked into the rooms: each piece is modelled around its own origin
(floor at y=0) and instanced into the mansion-global frame at load. The placements live in
`/Map/map2.szp` — the archive the game reads alongside the room arcs — under `jmp/`, a
directory of **JMap tables** (`furnitureinfo`, `roominfo`, `objinfo`, `characterinfo`, …).
The format, read straight off the file (`lm/jmp.go`): a 16-byte header `{u32 recordCount,
u32 fieldCount, u32 dataOffset, u32 recordSize}`, then 12-byte field descriptors
`{u32 nameHash, u32 bitmask, u16 offset, u8 shift, u8 type}` — type 1 a 32-byte string,
type 2 a f32, else a masked/shifted u32. Field names exist only as hashes (they lived in
Nintendo's conversion tool); the columns were named by matching values against known data:

- `jmp/furnitureinfo` — 730 records × 0xC4: actor class ("furniture"), **model name**
  (the member base name in the room's arc: `chest`, `syan`, `o_isu`, …), a free tag
  string, **pos/rot°/scale xyz** (three f32 triples at +0/+0xC/+0x18), and the **room
  number** (+0x94), plus behaviour ints (move/appear flags, hitbox). Verified against the
  foyer: its 6 placements name exactly `room_02.arc`'s furniture members, and the game's
  own frame shows the chest of drawers at (-440, 0, -560) rot 14° and the mirror at
  (420, 0, -400) rot -20° — where our assembled render puts them.
- `jmp/roominfo` — 72 records of per-room engine params (ambient RGB etc.);
  `jmp/objinfo` — 414 effect spawns ("fire": candle flames); enemy/character/observer
  tables place the szp actor models (out of scope here).

`lmtool -mansion DIR` exports the whole thing in one pass: all 75 `room_*.arc` shells
(rooms are mansion-global, so Phase 3's "assembly" is just loading them together),
every furniture `.bin` once — content-hash deduped across arcs (the eleven `o_isu`
chairs ship one GLB; a name reused for different geometry gets a numbered variant) —
and `placements.json` mapping each room to its furniture instances. 451 unique furniture
GLBs, 632 resolved placements (one orphan: `syan45`, a chandelier whose member was cut
from `room_45.arc`). The Studio's `lm-room` renderer assembles a room the same way the
game does: shell GLB + placement instances, each under a holder with the record's
TRS (rotation ZYX, degrees). Rooms are levels, not objects: the camera flies
(the shared FlyCam — WASD/arrows, virtual sticks on touch), and the furniture keeps
the game's interaction — in the game the vacuum yanks furniture open, so clicking a
piece plays its .anm clip and freezes on the last frame; clicking again plays it
backwards, closing the drawers up.

## Part XI — .anm: the furniture moves

The `.anm` files (each room arc's `anm/` directory, plus the door swings in
`/Game/game_usa.szp`) animate the `.bin` scene graph directly — the fourth wearing of the
same hermite-channel mechanism. Found statically: the DOL's door-path table
(`/iwamoto/Door/pull.anm`, …, at 0x802FF944) leads to the list loader `0x8001FB10`, the
binder `0x8001DE90`, and the evaluator `0x8001E04C`, which read:

```
+0 u8 version   +1 u8 loop   +4 u32 frame count
+8/+0xC/+0x10   pool offsets (scale / rotation / translation floats)
+0x14           descriptor table offset (0x18)
```

Descriptors are **54 bytes per graph node**: 9 channels of `{s16 count, s16 offset,
s16 fourFlag}` in the order `sx sy sz rx ry rz tx ty tz`, channel group *n* indexing
pool *n*. `count` 1 is a constant at `pool[offset]`; otherwise `count` hermite keys of
`{time, value, tangent}` (separate in/out tangents when flagged), times in 30 fps frames,
rotations in degrees — `lm/anm.go` shares `key.go`'s evaluator verbatim. The evaluated
TRS *replaces* the node's own, which is also the decode's verification: `chest_0.anm`
(the rest clip) reproduces `chest.bin`'s five nodes' graph TRS to the last float, all 45
channels. `chest_1` then slides the drawer node out (`tz` keys 19→64 over 15 frames)
— the visible "chest opens" moment. `lmtool -mansion` exports any furniture with clips
through a node-hierarchy GLB (`binGLBAnimated`) carrying one glTF animation per clip;
the Studio's gallery plays them (click cycles clips).

The `.bas` files that pair with the interaction clips (`chest_1.bas` ↔ `chest_1.anm`)
are the clips' **sound-cue tracks**: `{u16 count, u16, u32}` then 32-byte records
`{u32 soundID, f32 start, f32 end, f32 scale, u32 direction, u8 volume, u8 pan,
u16 flags}`. The timestamps live on the paired clip's timeline (chest cues at frames
0/9.03/12.93 of its 15; the chandelier at 0/60/75/171/296 of its 300), sustained cues
carry a nonzero end (the chain rattle holds 0→60), one-shots end at 0 — and the records
come in mirrored pairs with the times swapped and the direction field flipped (0x01/0x11
forward, 0x02/0x12 backward): the game plays the same clips in reverse to close the
furniture, with its own cue set, exactly the interaction the Studio viewers mimic.
Sound IDs are JAudio effect-bank entries (0x10xx/0x18xx); making them audible means
synthesizing the `AudioRes` banks — an audio project for another day.

## Part XII — the assembled mansion

Every room shell is modelled in one mansion-global frame — a bounds survey of all 75
`room.bin`s (`extract/cmd/roombounds`) shows them tiling the building rather than piling
at the origin: basement y −600..0, ground floor −50..550, first floor 500..1100, attic
1050..2100, with a handful of building-sized shells that are not rooms at all but the
**exterior**: the outer walls and grounds (`room_11/16/23/37/72`), and the roof
(`room_59/60`). Assembly is therefore exactly what Phase 3 hoped: load everything.

The Studio's `lm-mansion` renderer does that — all 75 shells streamed a few at a time
(each furnished from `placements.json` as it lands, the furniture cache shared across
rooms so repeated pieces download once), the exterior shells on their own layer
(on = the mansion on its grounds under the night sky; off = the dollhouse cutaway),
FlyCam to walk the halls, and the same click-to-vacuum furniture as the single-room
viewer. Standing in the assembled foyer, the archway that was a black hole in the
single-room export now opens into the parlor beyond it.

**The doors** live in the DOL, not in any file. The map-info header at `0x8030377C`
(GLME01/USA) carries the room table at +0x14 (74 × 48 B) and the **door list** at +0x18:
28-byte records `{u8 axis, u8 flag, …, u8 type@6, u8 id@7, s32 pos[3]@8, …,
u16 size[3]@20, u8 roomA@26, u8 roomB@27}`, terminated by axis 0 — 72 openings, each
naming the two rooms it connects. axis 1 faces z, 2 faces x, 4 is a floor opening; one
size component is zero (the thin axis). A word of 255 at +4 is a doorless archway (18 of
them, the check the panel counter at `0x8001A0C0` makes); otherwise the type byte indexes
the 20-byte table at `0x802FF95C` — `{u8 kind (2 = double), …, u8 model@19}` — whose
model names `/iwamoto/Door/{saku,door_NN}.bin` via the path table at `0x802FF868`
(54 leafed doors: door_01 ×30, door_09 ×13, three doubles). The leaf models are 200×300
`.bin`s whose leaf node sits at +100 (the GLB is centred on its opening, hinge node on
the jamb); the swing `.anm`s (`pull`/`push`) are complete open-and-shut cycles — the
game plays them as Luigi walks through, so the Studio holds a clicked door at the
swing's apex and runs it backwards on the next click. `lmtool -mansion` exports
`doors/*.glb` and a `doors` array in placements.json; `lm-mansion` hinges singles at
the record position and doubles at ±(width−200)/2, mirrored.

The mansion also peels: every shell (and its furniture and doors) is grouped by storey —
basement / ground / first / attic by the shell's lowest y — and the Studio's layer
toggles strip the dollhouse floor by floor, exterior and roof on their own switch.

`room_41` is a 10-unit dummy box far outside the building; `syan45`'s chandelier record
names a member that was cut from `room_45.arc`.

## Open items

* **The mansion tour — a collision-aware camera fly-through** of the assembled mansion.
  There is no ready-made path in the game's data (`jmp/railinfo` + the `path/*` splines
  are *actor* rails — ghost escape routes, the rat runs — and the game's own cameras are
  fixed per-room rigs), so the tour is generated offline from data we already have, and
  shipped as a `cameraTrack` JSON the existing fly-until-grabbed player consumes:
  1. **Route from the door graph.** The DOL door list is a room-adjacency graph with
     exact 3-D coordinates on every edge. A hand-picked room order (a dozen showpiece
     rooms — foyer, ballroom, kitchen, conservatory, a basement corridor; the graph
     fills in the connective legs) turns into door-to-door legs, never centroid-based:
     room centroids sit inside furniture (the billiard table) or under stair sweeps.
  2. **Free space from the shipped geometry.** The room + furniture GLBs plus the
     placement transforms are the collision proxy — no new decode needed (`col.mp`,
     905 KB in map2.szp, is the game's own collision mesh and remains unreversed; it
     would say the same thing). Per room: voxelize at ~40–50 units, mark cells cut by
     any triangle, dilate the occupied set by the camera clearance (~60–80 units).
     The remainder is the flyable volume — it excludes banisters, chandelier drops,
     tabletops by construction. Doorways get a locally reduced clearance radius
     (a 200-unit opening cannot pass an 80-unit bubble; the approach is perpendicular
     and the door has already swung open).
  3. **A\* door-to-door with a clearance-shaped cost.** Plain shortest-path hugs
     corners; adding a proximity penalty pushes the route toward the medial axis of
     each free pocket ("the cameraman's comfortable line"), and a mild band cost keeps
     it near eye height (~140–160 above the storey's floor) instead of ballooning to
     the ceiling. Vertical transitions happen inside the rooms that justify them
     (the foyer, stairwell room 36).
  4. **Smooth, then re-validate the smooth.** Shortcut the voxel staircase (drop
     waypoints while the straight segment keeps line-of-sight through free cells),
     fit a centripetal Catmull-Rom, arc-length reparameterize for constant speed —
     then sample the spline densely and check every sample against the grid,
     re-inserting waypoints where the curve bulges out of the corridor. This loop is
     the step naive smoothing skips, and why naive smoothed paths clip.
  5. **Doors open ahead of the camera.** The tour knows which opening each leg
     threads; fire the swing's open half a second before arrival, the close half once
     through — the leaves hold at their apex exactly like the click interaction.
  6. **Verify like everything else ships:** fly the generated track headless, record
     the minimum distance-to-geometry over the whole run, eyeball the worst frames.
     The deliverable is `mansion-tour.json` plus a "never closer than N units" number,
     not a spline that looked fine in three scrubbed spots. Caveat to check on the
     full watch-through: a collision-clean *path* is not yet a clean *picture* — the
     look-ahead target can drag the view across a wall edge near sharp turns; damp the
     look target harder or pull it in when a camera→target ray is occluded.
* The demo's binding of shots to archives and archive members to actors (the `.scd` names
  only lights; the model actors and their mirror placement matrix are set up in code).
* The `.sls` normal pass's target indexing (exists, unexercised in the opening shots).
* `.txp` — texture-pattern animation (the torch's flame), consumed by the evaluator at
  `0x8005C7AC` through the `.scd`-style channel interpolators.
* The mansion itself: `Map/map*.szp` beyond map2 (map0/map1/map6's own room sets), the
  in-game actor models in `model/*.szp` (same .mdl format family, unverified), and
  `col.mp` — the mansion's collision mesh.
* Audio: `AudioRes/` (JAudio banks, sequences, and the `.afc` streams the cutscene
  plays) — also the gate for making the `.bas` sound cues audible.
