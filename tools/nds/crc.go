package nds

// CRC16 computes the DS header/secure-area checksum: CRC-16 with the reflected
// polynomial 0xA001 (CRC-16/MODBUS), initial value 0xFFFF. The cartridge header
// checksum at 0x15E covers header bytes 0x000..0x15D.
func CRC16(data []byte) uint16 {
	crc := uint16(0xFFFF)
	for _, b := range data {
		crc ^= uint16(b)
		for i := 0; i < 8; i++ {
			if crc&1 != 0 {
				crc = crc>>1 ^ 0xA001
			} else {
				crc >>= 1
			}
		}
	}
	return crc
}

// VerifyHeaderCRC recomputes the header checksum over bytes 0x000..0x15D and reports
// whether it matches the stored value at 0x15E.
func (r *ROM) VerifyHeaderCRC() (computed uint16, ok bool) {
	if len(r.Data) < 0x160 {
		return 0, false
	}
	computed = CRC16(r.Data[0x00:0x15E])
	return computed, computed == r.Header.HeaderCRC
}
