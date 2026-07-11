package psp

// prx.go strips the ~PSP encryption wrapper off a PSP executable to recover the
// plaintext ELF, driving the KIRK engine (kirk.go) with the platform key set
// (kirk_keys.go). The ~PSP container is a 0x150-byte header over an AES-CBC body; the
// header at +0xD0 carries a 4-byte tag that selects a per-firmware XOR seed. The
// decryption rebuilds a KIRK command-1 header from the ~PSP header fields XORed and
// kirk7-descrambled with that seed, then runs KIRK command 1 over the body.
//
// Two closely-related "types" share the tag table: type 0 uses the header fields as
// they are; type 1 first kirk7-decrypts a slice of the header. The correct type is the
// one whose reconstructed header passes the SHA-1 integrity check the format embeds, so
// both are tried and the matching one is used.
//
// Algorithm reference: the PSP KIRK/PRXDecrypter specification (platform crypto, not
// game data); reimplemented here from the format. Keys are documented platform
// constants in kirk_keys.go.

import (
	"crypto/sha1"
	"encoding/binary"
	"fmt"
)

// tagInfo binds a ~PSP tag to its 0x90-byte XOR seed and its kirk7 key id ("code").
type tagInfo struct {
	seed []byte // 0x90 bytes
	code int
}

// tagTable maps the ~PSP header tag (+0xD0) to its decryption parameters. Retail 2.xx
// game EBOOT.BIN uses tag 0xC0CB167C. Further tags (other firmwares / demo / update
// modules) are added here as their seeds are transcribed.
var tagTable = map[uint32]tagInfo{
	0xC0CB167C: {seed: keyEBOOT2xx[:], code: 0x5D},
}

// DecryptPRX strips a ~PSP KIRK-encrypted container and returns the plaintext ELF
// image plus the decryption tag.
func DecryptPRX(raw []byte) (plain []byte, tag uint32, err error) {
	if len(raw) < 0x150 || string(raw[0:4]) != "~PSP" {
		return nil, 0, fmt.Errorf("prx: not a ~PSP container")
	}
	tag = binary.LittleEndian.Uint32(raw[0xD0:])
	ti, ok := tagTable[tag]
	if !ok {
		return nil, tag, fmt.Errorf("prx: unknown ~PSP tag 0x%08X", tag)
	}
	// Try type 0 (no pre-decrypt) then type 1 (pre-decrypt); the SHA-1 check selects.
	for _, preDecrypt := range []bool{false, true} {
		out, matched, derr := decryptTagType(raw, ti, preDecrypt)
		if derr != nil {
			return nil, tag, derr
		}
		if matched {
			return out, tag, nil
		}
	}
	return nil, tag, fmt.Errorf("prx: SHA-1 header check failed for tag 0x%08X", tag)
}

// decryptTagType reconstructs and runs the KIRK command-1 decryption for one type. It
// returns matched=false (with no error) when the SHA-1 header check does not hold, so
// the caller can try the other type.
func decryptTagType(raw []byte, ti tagInfo, preDecrypt bool) (out []byte, matched bool, err error) {
	seed := ti.seed
	decryptSize := int(int32(binary.LittleEndian.Uint32(raw[0xB0:])))

	// Reassemble the 0x150-byte structure the way PRXDecrypter does: tag, the embedded
	// SHA-1, the unused region, the 0x90-byte "kirk block" (split across the ~PSP
	// header), and the 0x80-byte PRX header — laid out contiguously so the cross-field
	// kirk7 below matches the reference's pointer arithmetic exactly.
	st := make([]byte, 0x150)
	copy(st[0x00:0x04], raw[0xD0:0xD4])   // tag
	copy(st[0x04:0x18], raw[0xD4:0xE8])   // sha1 (0x14)
	copy(st[0x18:0x40], raw[0xE8:0x110])  // unused (0x28)
	copy(st[0x40:0x80], raw[0x110:0x150]) // kirkBlock[0x00:0x40] (key data)
	copy(st[0x80:0xD0], raw[0x80:0xD0])   // kirkBlock[0x40:0x90]
	copy(st[0xD0:0x150], raw[0x00:0x80])  // prxHeader

	if preDecrypt {
		if err := kirk7(st[0x10:0xB0], st[0x10:0xB0], ti.code); err != nil {
			return nil, false, err
		}
	}

	// SHA-1 over seed[0:0x14] || unused || kirkBlock || prxHeader, compared to the
	// embedded (possibly pre-decrypted) SHA-1.
	h := sha1.New()
	h.Write(seed[0:0x14])
	h.Write(st[0x18:0x40])
	h.Write(st[0x40:0xD0])
	h.Write(st[0xD0:0x150])
	if !bytesEqual(h.Sum(nil), st[0x04:0x18]) {
		return nil, false, nil
	}

	// Build the KIRK command-1 header at offset 0x40 of a working copy of the file.
	const offset = 0x40 // = 0x150 - 0x90 - 0x80
	buf := make([]byte, len(raw))
	copy(buf, raw)
	copy(buf[offset:offset+0x90], st[0x40:0xD0])        // header = kirkBlock
	copy(buf[offset+0x90:offset+0x110], st[0xD0:0x150]) // + prxHeader

	// decryptKirkHeaderType0: XOR the first 0x70 header bytes with seed[+0x14],
	// kirk7-descramble, then XOR with seed[+0x20].
	tmp := make([]byte, 0x70)
	for i := 0; i < 0x70; i++ {
		tmp[i] = st[0x40+i] ^ seed[0x14+i]
	}
	if err := kirk7(tmp, tmp, ti.code); err != nil {
		return nil, false, err
	}
	for i := 0; i < 0x70; i++ {
		buf[offset+i] = tmp[i] ^ seed[0x20+i]
	}

	plain, err := kirkCMD1(buf[offset:])
	if err != nil {
		return nil, false, err
	}
	if decryptSize >= 0 && decryptSize <= len(plain) {
		plain = plain[:decryptSize]
	}
	return plain, true, nil
}

func bytesEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
