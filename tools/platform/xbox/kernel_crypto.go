package xbox

// kernel_crypto.go implements the Xc* crypto exports the title's own content/save
// integrity layer reaches. These are standard, published algorithms (SHA-1, HMAC, RC4)
// — general cryptographic primitives, not game-specific logic — so the Go standard
// library supplies the hash core (as crypto/md5 already does in xiso.go). What is
// Xbox-specific is the HMAC *construction*, which deviates from RFC 2104, and that we
// reproduce by hand rather than deferring to crypto/hmac.
//
// The console's per-box secrets (XboxEEPROMKey, XboxHDKey, XboxSignatureKey) are burned
// into each retail unit's EEPROM/flash and are not present in a disc image; we do not
// have them and do not invent them (the HLE-fiction trap). The title reads its HMAC key
// through a data-export slot; with no provisioned secret that key is a zero block, so a
// digest computed here is self-consistent with anything this same run signed, but will
// not match a signature produced on real hardware with the real key. That is the honest
// outcome — a verify against disc-embedded, real-key content legitimately fails on a
// fresh console, exactly as the hardware behaves when the secret cannot be recovered.

import (
	"crypto/sha1"
	"encoding"
	"hash"
)

// shaLoad reconstructs the streaming SHA-1 state for a guest context address (an unknown
// context is treated as freshly initialised — the title always calls XcSHAInit first).
func (m *Machine) shaLoad(ctx uint32) hash.Hash {
	h := sha1.New()
	if b, ok := m.shaCtx[ctx]; ok {
		if u, ok := h.(encoding.BinaryUnmarshaler); ok {
			u.UnmarshalBinary(b)
		}
	}
	return h
}

// shaStore marshals a streaming SHA-1 state back into the context map (gob-friendly bytes).
func (m *Machine) shaStore(ctx uint32, h hash.Hash) {
	if mm, ok := h.(encoding.BinaryMarshaler); ok {
		if b, err := mm.MarshalBinary(); err == nil {
			m.shaCtx[ctx] = b
		}
	}
}

// readBytes copies n bytes of guest memory into a Go slice (byte-wise, so trap/MMIO
// apertures resolve through the same path the CPU uses — a key read through an
// unprovisioned data-export slot yields zeros, not a fault).
func (m *Machine) readBytes(addr, n uint32) []byte {
	b := make([]byte, n)
	for i := uint32(0); i < n; i++ {
		b[i] = m.Read(addr + i)
	}
	return b
}

// writeBytes copies a Go slice into guest memory.
func (m *Machine) writeBytes(addr uint32, b []byte) {
	for i, v := range b {
		m.Write(addr+uint32(i), v)
	}
}

const xcShaBlock = 64 // SHA-1 block size, and the HMAC key/pad length

// kernelCryptoHandler returns the handler for an Xc* crypto ordinal, or nil.
func kernelCryptoHandler(ord uint16) func(*Machine) int {
	switch ord {
	case 335: // XcSHAInit(pbSHAContext) -> void. Verified from the SHA wrapper at 0x20CA48
		// (LEA [EBP-30]; PUSH; CALL) — one context pointer, followed by XcSHAUpdate/Final on
		// the same buffer. Table-328 + the Xc block's +7 drift = 335. We keep the digest
		// state host-side keyed by the context address; a fresh Init resets it.
		return func(m *Machine) int {
			m.shaStore(m.arg(0), sha1.New())
			m.setRet(0)
			return 1
		}
	case 336: // XcSHAUpdate(pbSHAContext, pbInput, cbInput) -> void. Verified: three args at
		// 0x20CA6F (context, data, len), called once per data buffer. Table-329 + 7 = 336.
		return func(m *Machine) int {
			ctx, data, n := m.arg(0), m.arg(1), m.arg(2)
			h := m.shaLoad(ctx)
			h.Write(m.readBytes(data, n))
			m.shaStore(ctx, h)
			m.setRet(0)
			return 3
		}
	case 337: // XcSHAFinal(pbSHAContext, pbDigest) -> void. Verified from 0x20CAA4 (PUSH
		// digest-out [EBP+44]; PUSH context [EBP-30]; CALL), the 20-byte digest then copied
		// to the caller's buffer. Table-330 + 7 = 337. Finalise and release the context.
		return func(m *Machine) int {
			ctx, out := m.arg(0), m.arg(1)
			h := m.shaLoad(ctx)
			m.writeBytes(out, h.Sum(nil)[:20])
			delete(m.shaCtx, ctx)
			m.setRet(0)
			return 2
		}
	case 338: // XcRC4Key(pbKeyTable, cbKeyMaterial, pbKeyMaterial) -> void. Verified from the
		// crypto wrapper at 0x20CE2E (PUSH keyTable; PUSH 0x14; PUSH keyData), the keyData
		// being the 20-byte SHA digest just produced. Table-331 + the Xc block's +7 drift =
		// 338. Standard RC4 key scheduling; the 258-byte state (S then i,j) is kept host-side
		// keyed by the guest table address, since that buffer's layout is opaque here.
		return func(m *Machine) int {
			keyTable, cbKey, pbKey := m.arg(0), m.arg(1), m.arg(2)
			key := m.readBytes(pbKey, cbKey)
			var s [256]byte
			for i := range s {
				s[i] = byte(i)
			}
			if len(key) > 0 {
				j := 0
				for i := 0; i < 256; i++ {
					j = (j + int(s[i]) + int(key[i%len(key)])) & 0xFF
					s[i], s[j] = s[j], s[i]
				}
			}
			state := make([]byte, 258)
			copy(state, s[:]) // state[256]=i, state[257]=j both start 0
			m.rc4Ctx[keyTable] = state
			m.setRet(0)
			return 3
		}
	case 339: // XcRC4Crypt(pbKeyTable, cbData, pbData) -> void. Verified from 0x20CE40 /
		// 0x20CE49 (PUSH keyTable; PUSH len; PUSH data), applied in place and continuing the
		// keystream across successive calls on the same table. Table-332 + 7 = 339. RC4 is
		// its own inverse, so this both encrypts and decrypts.
		return func(m *Machine) int {
			keyTable, n, data := m.arg(0), m.arg(1), m.arg(2)
			state, ok := m.rc4Ctx[keyTable]
			if !ok || len(state) < 258 {
				// No prior XcRC4Key on this table: identity state (S[i]=i), the honest
				// "uninitialised" behaviour rather than a fabricated keystream.
				state = make([]byte, 258)
				for i := range state[:256] {
					state[i] = byte(i)
				}
			}
			var s [256]byte
			copy(s[:], state[:256])
			i, j := state[256], state[257]
			for k := uint32(0); k < n; k++ {
				i++
				j += s[i]
				s[i], s[j] = s[j], s[i]
				ks := s[byte(s[i]+s[j])]
				m.Write(data+k, m.Read(data+k)^ks)
			}
			copy(state[:256], s[:])
			state[256], state[257] = i, j
			m.rc4Ctx[keyTable] = state
			m.setRet(0)
			return 3
		}
	case 346: // XcDESKeyParity(pbKey, dwKeyLength) — set each key byte's DES parity bit.
		// Verified from its call site (0x207A7F via the thunk at 0x85D80): two args, a
		// 24-byte (3DES) key the caller just assembled by REP MOVSD and 0x18, result
		// ignored — the canonical in-place parity fix. DES keys carry odd parity in the
		// low bit of every byte (FIPS 46-3, a published standard): the low bit is set so
		// the byte's total population count is odd.
		return func(m *Machine) int {
			key, n := m.arg(0), m.arg(1)
			for i := uint32(0); i < n; i++ {
				b := m.Read(key + i)
				b = b&0xFE | (popcount7(b>>1) & 1) ^ 1
				m.Write(key+i, b)
			}
			m.setRet(0)
			return 2
		}
	case 340: // XcHMAC(pbKeyMaterial, cbKeyMaterial, pbData, cbData, pbData2, cbData2, pbDigest)
		// Verified from the two library wrappers at 0x20CAC9 (compute → copy digest out)
		// and 0x20CB09 (compute → REP CMPSB compare), each passing seven __stdcall args
		// and taking a 20-byte SHA-1 digest. This is HMAC-SHA1 with the Xbox construction:
		// the key is used as-is up to the 64-byte block size and TRUNCATED (not pre-hashed)
		// if longer — the documented deviation from RFC 2104. Both data buffers are fed to
		// the inner hash in sequence (the API lets a caller sign header+body without
		// concatenating them). Returns void; callers read the digest, not EAX.
		return func(m *Machine) int {
			pbKey, cbKey := m.arg(0), m.arg(1)
			pbData, cbData := m.arg(2), m.arg(3)
			pbData2, cbData2 := m.arg(4), m.arg(5)
			pbDigest := m.arg(6)

			if cbKey > xcShaBlock {
				cbKey = xcShaBlock // Xbox: truncate an over-long key, do not hash it (RFC 2104 would hash)
			}
			key := m.readBytes(pbKey, cbKey) // zero-padded to the block below

			var ipad, opad [xcShaBlock]byte
			for i := 0; i < xcShaBlock; i++ {
				var k byte
				if i < len(key) {
					k = key[i]
				}
				ipad[i] = k ^ 0x36
				opad[i] = k ^ 0x5C
			}

			inner := sha1.New()
			inner.Write(ipad[:])
			inner.Write(m.readBytes(pbData, cbData))
			inner.Write(m.readBytes(pbData2, cbData2))
			innerDigest := inner.Sum(nil)

			outer := sha1.New()
			outer.Write(opad[:])
			outer.Write(innerDigest)
			digest := outer.Sum(nil)

			m.writeBytes(pbDigest, digest[:20])
			m.setRet(0)
			return 7
		}
	}
	return nil
}

// popcount7 counts the set bits of a 7-bit value (a DES key byte sans parity bit).
func popcount7(b byte) byte {
	var n byte
	for ; b != 0; b >>= 1 {
		n += b & 1
	}
	return n
}
