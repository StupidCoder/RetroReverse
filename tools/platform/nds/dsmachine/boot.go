package dsmachine

import (
	"encoding/binary"

	"retroreverse.com/tools/platform/nds"
)

// The state the BIOS leaves behind.
//
// We do not run the DS boot ROM: we load the two binaries at their link addresses
// and start the cores at their entry points, the way every "direct boot" does. But
// the boot ROM does not only load — it *leaves things in main RAM*, and the game
// reads them. Skip that and the game boots into a machine that is subtly lying to
// it: it cannot find its own filesystem, and it does not know who owns the console.
//
// Two of these matter enough to name:
//
//   - The cartridge header copy at 0x027FFE00. NitroSDK's file system does not read
//     the FAT and FNT offsets off the card — it reads them out of this copy. Without
//     it, every file the game opens is at offset zero.
//   - The firmware user settings at 0x027FFC80. The BIOS validates the two copies in
//     the firmware flash, picks the newer, and leaves a 0x70-byte block here. The
//     game reads its language, its owner's name and — the one with teeth — the
//     touchscreen calibration from this block, not from the SPI bus.
//
// Everything here is addressed through main RAM's top mirror (0x027FFxxx aliases
// 0x023FFxxx), which is how DS software has always written it.
const (
	bootHeaderCopy   = 0x027FFE00 // the cartridge header, 0x170 bytes
	bootUserSettings = 0x027FFC80 // the firmware user-settings block, 0x70 bytes
	bootChipID       = 0x027FF800
	bootChipIDAlt    = 0x027FFC00
)

// directBoot writes the boot parameter block the BIOS would have left.
func (m *Machine) directBoot(rom *nds.ROM) {
	b := &bus{c: m.ARM9}
	le := binary.LittleEndian

	// The cartridge header, copied verbatim.
	n := 0x170
	if len(rom.Data) < n {
		n = len(rom.Data)
	}
	for i := 0; i < n; i++ {
		b.Write(bootHeaderCopy+uint32(i), rom.Data[i])
	}

	// The chip ID, which the BIOS leaves in two places, and the checksums beside it.
	const chipID = 0x00003FC2
	for _, base := range []uint32{bootChipID, bootChipIDAlt} {
		b.w32(base+0, chipID)
		b.w32(base+4, chipID)
		b.Write(base+8, byte(rom.Header.HeaderCRC))
		b.Write(base+9, byte(rom.Header.HeaderCRC>>8))
		b.Write(base+10, byte(rom.Header.SecureCRC))
		b.Write(base+11, byte(rom.Header.SecureCRC>>8))
	}
	b.Write(0x027FFC40, 1) // boot indicator: booted from a cartridge, not the firmware menu

	// The firmware user settings, chosen the way the BIOS chooses them: take the two
	// copies, keep the ones whose checksum is right, and prefer the newer counter.
	us := m.spi.currentUserSettings()
	for i, v := range us {
		b.Write(bootUserSettings+uint32(i), v)
	}
	_ = le
}

// currentUserSettings returns the live user-settings block from the firmware — the
// newer of the two valid copies, exactly as the BIOS selects it.
func (s *spibus) currentUserSettings() []byte {
	le := binary.LittleEndian
	best := []byte(nil)
	bestCount := -1
	for _, off := range []int{userSettings1, userSettings2} {
		blk := s.firmware[off : off+0x100]
		if nds.CRC16(blk[0x00:0x70]) != le.Uint16(blk[0x72:]) {
			continue // a copy that does not check out is not a copy
		}
		if c := int(le.Uint16(blk[0x70:])); c > bestCount {
			best, bestCount = blk, c
		}
	}
	if best == nil {
		return make([]byte, 0x70)
	}
	return best[:0x70]
}
