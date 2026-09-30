# Static asset knowledge (M6)

A game's `knowledge.json` holds the exact release identities, location graph,
semantic assets, dependencies, evidence and curated output identities. The package
contains facts and bounded recipes, never code, commands or game payloads.

## Commands

```sh
# Validate structure, references and the existing curated exports.
go run ./tools/cmd/knowledgecheck -artifacts site \
  games/captain-toad-treasure-tracker-3ds/knowledge.json

# Independently resolve the original model bytes from your exact local image.
# -out and -report are optional; both refuse to overwrite existing files.
go run ./tools/cmd/knowledgeasset \
  -media 'cartridge=/path/to/Captain Toad.cci' \
  -release europe-decrypted -asset kinopio \
  -out /tmp/kinopio.bch -report /tmp/kinopio-provenance.json \
  games/captain-toad-treasure-tracker-3ds/knowledge.json

go run ./tools/cmd/knowledgeasset -decoders

# Regenerate browser metadata and the script-free /knowledge/ catalog.
python3 tools/browser/package.py --site-only
```

The resolver hashes the **complete** raw media before following any location.
Exactly the declared media roles must be supplied; filenames and input order do
not identify releases. It keeps those file handles for the session. Do not modify
media while resolving. File, member, transform and final asset hashes then verify
the selected contents independently. Wrong sizes/hashes, missing or ambiguous
members, unsupported formats and exhausted limits fail with the failing location.
It never searches for a similar release or falls back to an unverified offset.
The Go API returns `resolved` on success and distinguishes `inactive`,
`unavailable`, `mismatch` and `ambiguous` failures; the CLI exits nonzero and
names the failing stage. No failed asset is written as a successful report.

This milestone resolves static assets locally in Go. The browser catalog displays
escaped metadata and links to curated outputs; it does not execute recipes or
upload local media. There is no browser decoder counterpart requiring a port or
parity claim. Logical file-set roots, encrypted partitions, live pointers and RAM
asset provenance are not implemented by this static resolver.

## Location profile

Existing `image` and `relative` locations gain an optional positive `length`.
Without it the view extends to the end of its containing span. Relative offsets
are canonical hexadecimal byte strings. All ranges are half-open. A relative
view after decompression remains relative to decoded bytes, never a raw-image
alias.

New kinds all have a `parent` location:

| Kind | Fields and meaning |
|---|---|
| `sectors` | `first`, `count`, `stride`, `payloadOffset`, `payloadSize`: concatenate the selected bytes from each sector. Units are explicitly sectors/bytes. Framing and payload are not interchangeable. |
| `file` | `decoder`, `version:1`, exact normalized `path`, expected `size` and `sha256`. Resolve through a filesystem. |
| `member` | The same fields; resolve a container's named member. |
| `transform` | `decoder`, `version:1`, expected decoded `size` and `sha256`. Transform the complete parent view. |

Paths are case-sensitive, relative, slash-separated, and reject dot components,
control characters, drive prefixes and traversal. They select archive contents;
they are never used as extraction filenames. Decoder IDs/versions are a fixed
registry, with no plugin import, shell command, expression evaluator or network.

| Decoder | Kind | Implemented profile |
|---|---|---|
| `ncsd` v1 | member | `partition0` through `partition7`, 512-byte media units. |
| `ncch` v1 | member | `romfs` or `exefs`; explicit NoCrypto flag required. |
| `romfs` v1 | file | IVFC-wrapped Nintendo 3DS RomFS, exact UTF-16 names, bounded sibling traversal, duplicate/cycle rejection on the traversed directory lists. |
| `iso9660` v1 | file | Primary volume, 2048-byte logical blocks, exact on-disc names including `;1`, checked both-endian fields and noncontiguous multi-extent files. |
| `sarc` v1 | member | Little-endian, named entries with archive name-hash checks; duplicate names rejected. |
| `yaz0` v1 | transform | Bounded literals and overlapping back-references, checked output size before allocation. |

ISO Joliet/Rock Ridge, extended attributes, interleaving and associated files are
rejected. Supply raw-sector layout explicitly before ISO; the resolver does not
guess track geometry. RomFS authentication here comes from pinned input and
member hashes; it does not independently certify the IVFC signature/hash tree.
Format logic follows the repository's own `n3ds` and `iso9660` implementations,
using bounded random-access adapters instead of loading whole images or invoking
unbounded legacy parsers.

A provenance report records each view's size and parent; image/relative offsets;
container/filesystem extents in **parent coordinates**; sector stride/payload;
and transform IDs/hashes. Noncontiguous files retain all extents. Asset source
fragments record the actual consumed lengths, and dependency reports retain their
semantic roles. Package revision/source hash and release bind the report to its
recipe. Static resolution has no machine snapshot or mapping generation.

## Asset profile

Each asset declares:

- `label`, `description`, semantic `category`, and original `format`.
- Registered output `decoder` and `version`.
- Ordered `sources`: `{location, length}` fragments concatenated in that order.
- Exact concatenated source `size` and `sha256`.
- `dependencies`: `{asset, role}` for palette, texture, skeleton, animation or data.
  References must exist, share all applicable releases and form a bounded DAG.
- `releases` and `evidence` references.
- Optional `artifacts`: `{path, format, size, sha256}` for curated PNG/GLB outputs.
  Paths live under `public/<package-id>/` relative to the site root.

`bytes` v1 exports the verified source bytes without claiming to decode their
internal format. `indexed8` v1 requires `parameters:{width,height}` and exactly
one `palette` dependency containing 768 raw RGB bytes; it emits a deterministic
PNG. Its source size must equal width × height. Unknown formats may be named as
original formats, but cannot select an unknown executable decoder.

Artifact checks verify exact file size/hash, site-root containment including
symlinks, PNG decoding/dimensions, or GLB v2 chunk structure and embedded-resource
policy. They do not prove artistic correctness or validate every glTF semantic.
The curated Captain Toad GLB was produced by the existing game exporter; the
`bytes` recipe independently proves where its original BCH input lives, and does
not pretend to regenerate the animated GLB. Decoder implementation changes that
alter output semantics require a new registry version.

## Limits

- Package schema: 8 MiB, 4,096 locations/assets, 16 location/dependency depth.
- Media identity: at most 16 GiB total across the selected roles; streaming hash.
- A resolver session: 2 GiB of logical reads after initial identity hashing,
  65,536 metadata/dependency visits, and 128 MiB cumulative retained decoded/asset
  allocations. Reopen a session to inspect a separate large asset collection.
- One read/transform/asset: 64 MiB. Metadata tables/directories: 16 MiB.
- Paths: 1,024 characters, filesystem descent at most 32 components.
- At most 16 source fragments and 64 dependencies per asset; dependency validation
  also has a 65,536-visit limit. No runtime arbitrary loops.
- Indexed PNG: dimensions at most 4,096 each. Curated artifacts at most 256 MiB;
  PNG previews must fit 4,096 × 4,096.

Read buffers and PNG encoding use additional bounded scratch memory; the retained
allocation limit is not a promise that total process RSS stays below 128 MiB.
Resolvers are single-owner, synchronous offline tools. They do not run on the
emulator worker and do not claim interactive cancellation latency.

## Authoring workflow

1. Pin the exact original media and release. Derive offsets, directory names and
   formats using repository tools and the original image. Preserve uncertainty;
   never fill a plausible field just to complete the schema.
2. Write reusable location nodes from image to filesystem/container/transform.
   Hash each selected file/member and decoded stage. Record relationships only
   when the resource's own data or existing research establishes them.
3. Add semantic assets with fragment lengths, hashes and evidence. A BCH holding
   many resources remains a whole BCH until internal subresource ranges are known.
4. Add curated output hashes only for existing, reproducible exporter products.
   Keep the source chain separate from output filenames and output hashes.
5. Run `knowledgecheck`, `knowledgeasset` on matching media, negative/synthetic
   tests, and `knowledgecheck -artifacts site`. Keep private media/output scratch
   in ignored `image/` or `work/`, not in the knowledge package.
6. Increment revision when changing published facts. Update the narrative,
   regenerate the catalog/browser data, and commit source, tests and metadata
   together. Public tests regenerate synthetic ROM, raw-sector, filesystem,
   archive, malformed stream and preview fixtures without commercial game data.

`STANDARDS.md` now makes a knowledge-package update part of completing new or
revised game research. Existing games migrate as research resumes; an old game
without a package does not justify inventing a bulk conversion.
