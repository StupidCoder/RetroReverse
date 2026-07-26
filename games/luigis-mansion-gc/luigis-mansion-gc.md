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

## Open items

* The demo's binding of shots to archives and archive members to actors (the `.scd` names
  only lights; the model actors and their mirror placement matrix are set up in code).
* The `.sls` normal pass's target indexing (exists, unexercised in the opening shots).
* `.txp` — texture-pattern animation (the torch's flame), consumed by the evaluator at
  `0x8005C7AC` through the `.scd`-style channel interpolators.
* The mansion itself: `Iwamoto/map2/room_*.arc` room archives, `Map/map*.szp`, the in-game
  actor models in `model/*.szp` (same .mdl format family, unverified).
* Audio: `AudioRes/` (JAudio banks, sequences, and the `.afc` streams the cutscene plays).
