"""The bank-5 scene/level descriptor — decoded against its consumers.

`scene_run` ($1414) indexes a table at bank-5 $5600 by the scene/act number ($D238)
to a 40-byte descriptor, which $185D copies to work RAM at $D355. The descriptor is
read through a pushed pointer, so the fields below are named by *following each one to
the code that consumes it* (which twice corrected an earlier guess):

    +0        -> $D2D5  zone number (0..5; three acts per zone)        [unpacked $1913]
    +1/+2     -> $D232  a word (level extent — likely the vertical bound)
    +3/+4     -> $D234  a word (extent)
    +5/+6     -> $D26D  the level's LEFT scroll bound  (camera clamp, $4F33)
    +7/+8     -> $D26F  the level's RIGHT scroll bound = level width (camera clamp, $4F5B;
                        e.g. scene 0 = $18C0 ~ 6336 px). NOT a map pointer.
    +23       graphics bank ($09)
    +24/+25   pointer to the zone's compressed tile set (128 tiles) — VERIFIED by
              decompressing it to coherent level graphics
    +13..+19  per-act pointers into slot-2 data ($8634-style) — the actual level MAP /
              object data; consumed after $185D's $1930, format not yet decoded

So the descriptor is the level's BOUNDING BOX + tile-set pointer + (per-act) data
pointers. The level MAP itself is a block-based, streamed structure in the gameplay
engine (scroll_draw $3282 redraws a column when the camera moves >=8 px, expanding the
map from $D2AF); decoding that format is the next, large piece.
"""

from dataclasses import dataclass

import machine


@dataclass
class SceneDescriptor:
    zone: int          # +0    zone number 0..5, mirrored at +29
    extent_v: int      # +1/+2 -> $D232  (vertical extent / bound)
    extent2: int       # +3/+4 -> $D234
    scroll_left: int   # +5/+6 -> $D26D  left scroll bound
    scroll_right: int  # +7/+8 -> $D26F  right scroll bound (= level width)
    gfx_bank: int      # +23   ROM bank of this zone's graphics ($09)
    tiles_ptr: int     # +24/+25 -> compressed tile set in gfx_bank (128 tiles)
    raw: bytes         # all 40 bytes (incl. the +13..+19 map/object pointers, undecoded)

    @classmethod
    def decode(cls, file_off):
        d = machine.rom[file_off:file_off + 0x28]
        return cls(
            zone=d[0],
            extent_v=d[1] | d[2] << 8,
            extent2=d[3] | d[4] << 8,
            scroll_left=d[5] | d[6] << 8,
            scroll_right=d[7] | d[8] << 8,
            gfx_bank=d[23],
            tiles_ptr=d[24] | d[25] << 8,
            raw=bytes(d),
        )

    def tiles_file_off(self):
        """Flat ROM file offset of this zone's compressed tile set."""
        bank, addr = self.gfx_bank, self.tiles_ptr
        while addr >= 0x4000:        # normalise like the $0406 source setup
            addr -= 0x4000
            bank += 1
        return bank * 0x4000 + addr
