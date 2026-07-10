// Package n3ds reads the Nintendo 3DS cartridge image and the containers nested
// inside it, and models enough of the machine to execute the application's ARM11
// code.
//
// The medium is a CCI ("CTR Cart Image", extension .cci or .3ds): a flat dump of
// the cartridge, framed by an NCSD header and carrying up to eight partitions.
// Each partition is an NCCH container. Nesting, outermost first:
//
//	CCI file
//	└── NCSD header (0x100 signature + 0x100 header at 0x100)
//	    └── partition 0 — NCCH (the application; "CXI")
//	        ├── ExHeader   — code-set info: segment load addresses and sizes
//	        ├── ExeFS      — the executable filesystem: .code, banner, icon, logo
//	        │   └── .code  — the ARM11 program, optionally BLZ-compressed (blz.go)
//	        └── RomFS      — the read-only asset filesystem, under an IVFC hash tree
//	    ├── partition 1 — NCCH (the electronic manual)
//	    └── partition 7 — NCCH (the system update data)
//
// Everything in an NCCH is normally AES-CTR encrypted with keys that live in the
// console, not on the cartridge. A "decrypted" dump clears that by setting the
// NoCrypto bit in the NCCH flags; this package reads such dumps directly and
// reports an explicit error for an encrypted one rather than returning garbage —
// see NCCH.Encrypted. No key material is embedded here or attempted.
//
// All multi-byte fields are little-endian. Offsets and sizes in the NCSD and NCCH
// headers are counted in "media units" whose size is derived from the header flags
// (MediaUnitSize), not assumed to be 512.
package n3ds
