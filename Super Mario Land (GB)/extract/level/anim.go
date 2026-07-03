package level

// BG tile animation (the 1-3 wall torches, the world-2 water ripple, the 3-2
// waterfall). One routine drives it all: $23F8, the last call in the VBlank chain.
// When the level's enable flag $D014 (loaded at level init, $2445, from the table at
// $242D[ffe4]) is set, every 8th frame ($FFAC&7==0) it rewrites the HIGH bitplane of
// BG tile $5D (the 8 odd bytes $95D1,$95D3..$95DF), picking the pattern by $FFAC
// bit 3:
//
//	bit3=0 -> the per-world 8-byte accent at $3FAF+(world-1)*8 (bank 0)   [phase A]
//	bit3=1 -> RAM $C600 = the tile's own resting high plane               [phase B]
//
// $C600 is filled at tile-load time from the world's BG tileset source (traced from
// $0D30/$0D64): world 1 loads 128 tiles from bank2:$5032 to $9000 ($05D0, tile $5D at
// +$5D0; $05E8 copies its odd bytes), worlds 2-4 overlay 63 tiles from the $0DEA
// table source to $9310 (w2 bank1:$4402, w3 bank3:$4402, w4 bank1:$4BC2; tile $5D at
// +$2C0, odd bytes copied at $0DBE). The low bitplane never changes, so the tile
// flickers between two shapes with a 16-frame full cycle.
//
// Oracle-verified by extract/cmd/tileanimverify (live VRAM $95D0-$95DF vs this
// decode, per level).

const (
	animEnableTable = 0x242D // bank 0: per-ffe4 enable flag -> $D014 ($2445)
	animAccentTable = 0x3FAF // bank 0: per-world 8-byte phase-A high plane
	// AnimTile is the one BG tile id the engine animates.
	AnimTile = 0x5D
	// AnimPeriod is the frames each phase is shown ($FFAC&7==0 gate + bit 3).
	AnimPeriod = 8
)

// TileAnim is a level's BG tile animation: tile AnimTile alternates between
// Frames[0] (the accent phase) and Frames[1] (the resting tile as loaded), each
// shown for AnimPeriod frames. Each frame is full 2bpp tile data (16 bytes).
type TileAnim struct {
	Frames [2][16]byte
}

// animTileBase returns the resting (as-loaded) 16-byte data of tile AnimTile for a
// world, read from the traced tileset sources above.
func animTileBase(rom []byte, world int) (base [16]byte) {
	var bank int
	var src uint16
	if world == 1 {
		bank, src = 2, 0x5032+AnimTile*16
	} else {
		e := uint16(world-2) * 2
		ptr := uint16(rom[0x0DEA+e]) | uint16(rom[0x0DEA+e+1])<<8
		bank, src = []int{1, 3, 1}[world-2], ptr+0x2C0
	}
	for i := range base {
		base[i] = bankByte(rom, bank, src+uint16(i))
	}
	return base
}

// DecodeTileAnim returns the tile animation for a level id (0x11..0x43), or nil for
// levels whose $242D flag is off (the tile stays as loaded).
func DecodeTileAnim(rom []byte, id byte) *TileAnim {
	world := int(id >> 4)
	ffe4 := (world-1)*3 + int(id&0x0F) - 1
	if rom[animEnableTable+ffe4] == 0 {
		return nil
	}
	var a TileAnim
	a.Frames[1] = animTileBase(rom, world)
	a.Frames[0] = a.Frames[1]
	for i := 0; i < 8; i++ { // accent replaces the high (odd) plane only
		a.Frames[0][i*2+1] = rom[animAccentTable+(world-1)*8+i]
	}
	return &a
}
