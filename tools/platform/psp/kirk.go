package psp

// kirk.go reimplements the subset of the PSP's KIRK security-coprocessor commands
// needed to decrypt a ~PSP executable, from the format up. The KIRK operations used
// are all AES-128-CBC with a zero IV over the platform key set (kirk_keys.go):
//
//   - kirk7(keyId): CBC-decrypt with keyvault[keyId]      (KIRK command 7)
//   - kirkCMD1:     CBC-decrypt a 0x90-byte-header block  (KIRK command 1, "decrypt
//                   private") — the header's first 0x20 bytes are the AES+CMAC key
//                   pair, themselves CBC-encrypted under the KIRK master key; decrypt
//                   those, then CBC-decrypt the body with the recovered AES key.
//
// The CMAC/ECDSA integrity checks KIRK also performs are verification-only and are
// not needed to recover plaintext, so they are omitted; the SHA-1 header check that
// selects the ~PSP decryption "type" lives in prx.go.

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/binary"
	"fmt"
)

// cbcDecryptZero AES-128-CBC-decrypts the whole-block prefix of src under key with a
// zero IV, returning a fresh buffer. Any trailing sub-block bytes are copied through
// unchanged (KIRK processes only full 16-byte blocks).
func cbcDecryptZero(key, src []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	out := make([]byte, len(src))
	n := len(src) &^ 15
	if n > 0 {
		var iv [16]byte
		cipher.NewCBCDecrypter(block, iv[:]).CryptBlocks(out[:n], src[:n])
	}
	if n < len(src) {
		copy(out[n:], src[n:])
	}
	return out, nil
}

// kirk7 is KIRK command 7: CBC-decrypt size bytes of in under keyvault[keyId], in
// place (in and out may alias, as the callers rely on).
func kirk7(dst, src []byte, keyId int) error {
	if keyId < 0 || keyId >= len(keyvault) {
		return fmt.Errorf("kirk: key id 0x%X out of range", keyId)
	}
	out, err := cbcDecryptZero(keyvault[keyId][:], src)
	if err != nil {
		return err
	}
	copy(dst, out)
	return nil
}

// KIRK_CMD1_HEADER field offsets (0x90-byte header).
const (
	cmd1Mode       = 0x60 // u32, must be 1
	cmd1DataSize   = 0x70 // u32
	cmd1DataOffset = 0x74 // u32
	cmd1HeaderLen  = 0x90
)

// kirkCMD1 is KIRK command 1: in is a KIRK_CMD1_HEADER (0x90 bytes) followed by the
// encrypted body. It returns the decrypted body (data_size bytes). The header's first
// 0x20 bytes carry the AES and CMAC keys encrypted under the KIRK master key.
func kirkCMD1(in []byte) ([]byte, error) {
	if len(in) < cmd1HeaderLen {
		return nil, fmt.Errorf("kirk: CMD1 input too small (%d)", len(in))
	}
	if mode := binary.LittleEndian.Uint32(in[cmd1Mode:]); mode != 1 {
		return nil, fmt.Errorf("kirk: CMD1 mode %d != 1 (header decrypt likely wrong)", mode)
	}
	// Recover the AES/CMAC key pair (first 0x20 header bytes) under the master key.
	keys, err := cbcDecryptZero(kirk1Key[:], in[0:0x20])
	if err != nil {
		return nil, err
	}
	aesKey := keys[0:16]

	dataSize := int(binary.LittleEndian.Uint32(in[cmd1DataSize:]))
	dataOffset := int(binary.LittleEndian.Uint32(in[cmd1DataOffset:]))
	start := cmd1HeaderLen + dataOffset
	if start < 0 || dataSize < 0 || start+dataSize > len(in) {
		return nil, fmt.Errorf("kirk: CMD1 body [%#x:+%#x] out of range (%d)", start, dataSize, len(in))
	}
	return cbcDecryptZero(aesKey, in[start:start+dataSize])
}
