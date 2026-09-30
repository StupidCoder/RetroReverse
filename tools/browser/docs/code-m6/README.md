# M6 — Asset chains and research authoring

M6 implements bounded static asset resolution, a versioned Go decoder registry,
semantic dependencies, original-source and curated-output verification, and a
script-free site catalog at `/knowledge/` linked from the Studio's game list.
The specification and authoring workflow are in
[ASSET-RESOLUTION.md](../../../knowledge/ASSET-RESOLUTION.md).

## Real-game acceptance

Captain Toad's new single-file package pins the 512 MiB decrypted European image
and resolves:

`cartridge → NCSD partition0 → NCCH RomFS → ObjectData/Kinopio.szs → Yaz0 → SARC Kinopio.bch`

The verified BCH is 392,880 bytes, SHA-256
`afb9a8de5b5a3f0a4c135ede44fa42b1cb99b21de6217534588fa3b63a2580b4`.
Its opening-stage animation dependency independently resolves through
`ObjectData/KinopioAnimationSeason1OpeningStage.szs` to a 3,616,892-byte BCH.
Each intermediate selection and decoded archive has an expected size and hash.

The model asset links the existing curated `objects/kinopio.glb`, checking its
size, hash and GLB container. It does not assert a separate texture archive or
invent offsets for the model's embedded BCH textures/skeleton. Source resolution
exports original BCH bytes; the existing game exporter remains responsible for
animated GLB conversion. No private image or extracted BCH is committed.

## Public acceptance

Synthetic fixtures exercise ROM subranges and multiple fragments, raw 2352-byte
sectors with explicit 2048-byte payloads, a two-extent ISO file, the entire
NCSD/NCCH/RomFS/Yaz0/SARC chain, overlapping compression references, and a palette
PNG whose decoded pixel colors are checked. Negative cases cover wrong image,
missing member, mismatched stage hashes, cycles, unsupported decoder/version,
path traversal, bad extents/endianness, duplicate entries, truncation, decompression
and work limits, broken dependencies and stale previews. Format fuzz seeds are
public synthetic data. Browser data and the escaped catalog export reproducibly.

The M6 registry is deliberately static and local. The browser catalog neither
runs recipes nor claims to have verified a user's image. Unsupported filesystem,
compression and model-decoder modes fail closed; detailed limits are in the spec.
