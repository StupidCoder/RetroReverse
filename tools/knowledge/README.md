# Game knowledge tools — M1–M5 profile

`games/<slug>/knowledge.json` is the hand-maintained source. The initial Fort
package supplies Memory's existing annotations and records documented state and
function candidates. Generated browser data is never edited by hand.

## Commands (repository root)

```sh
go run ./tools/cmd/knowledgecheck games/fort-apocalypse-c64/knowledge.json

go run ./tools/cmd/knowledgecheck \
  -media tape=games/fort-apocalypse-c64/Fort_Apocalypse.tap \
  -release reference-pal games/fort-apocalypse-c64/knowledge.json

go run ./tools/cmd/knowledgeexport \
  -out site/emulators/knowledge-data.js games/fort-apocalypse-c64/knowledge.json

go run ./tools/cmd/knowledgeexport -check \
  -out site/emulators/knowledge-data.js games/fort-apocalypse-c64/knowledge.json

go test ./tools/knowledge ./tools/cmd/knowledgecheck ./tools/cmd/knowledgeexport
node tools/browser/tests/knowledge.mjs
```

`knowledgecheck` without media proves structural/semantic validity, not the truth
of the research. `-media role=path` verifies exact raw-image hashes and sizes.
Repeat it for multiple media roles. For logical file-set releases use repeated
`-file normalized/name=local/path`; raw and logical identity modes cannot be mixed.
`-release` asserts the unique expected match. Files are streamed through SHA-256;
no game bytes are written into packages or exports.

A package declares either `media` or `fileSet` per release. File-set canonical
serialization, path policy, member scope and NFC rules are specified in the
[M0 contract](../browser/docs/code-m0/CONTRACTS.md). Unicode normalization uses
pinned `golang.org/x/text`; checksums are recorded in `tools/go.sum`. Role matching
never infers disc order from input order. Multiple matches fail closed. Extra
logical files do not authorize annotations for those extras.

## Schema and validation

[game-knowledge-v1.schema.json](schema/game-knowledge-v1.schema.json) is a standard
Draft 2020-12 schema for the implemented profile. Go embeds it and evaluates the
exact keywords it uses, followed by semantic validation. This is a deliberately
small private evaluator, not a general-purpose JSON Schema library or an API for
user-supplied schemas. A schema keyword audit rejects future keywords unless
support is explicitly added and tested. Do not remove that fail-closed check.

Validation rejects unknown fields/kinds, unsupported nonempty operations, invalid
hashes, duplicate JSON keys, invalid UTF-8, excess depth/size, dangling references,
cyclic/deep locations and types, inconsistent file-set hashes, duplicate media
roles, ambiguous raw releases, overflowing addresses, out-of-space spans, bad enum
widths, overlapping struct/bitfield layouts and undersized array strides.

Implemented:

- Exact raw media identity and logical file-set canonicalization/matching.
- Sources/evidence with confirmed/inferred/hypothesis status and limitations.
- CPU, physical, image and relative locations; checked static bounds.
- Explicit integer types (width, endian, signedness), enum, bitfield, fixed-point,
  struct and fixed-count array definitions. Integer/enum scalar decoding preserves
  exact values as strings, including 64-bit values. Unknown enums stay numeric.
- Regions, state definitions, function entries and instruction-relative commentary.
- Deterministic browser data export carrying source SHA-256 and package revision.

The current profile does not implement general pointer/filesystem/sector/transform resolution,
sign-magnitude types, dynamic arrays, live struct decoding, runtime signatures,
general phase predicates, experiments, or executable asset recipes. Reserved
`assets` and `experiments` must be empty; unsupported location/type kinds
are rejected. Those features have later milestone gates rather than inert objects
that look executable. Future v1 extensions must remain fail-closed in older tools.

All M1 locations must resolve in each declared release; per-release relocation and
location overrides are not implemented. Use explicit release applicability for
read-only definitions; do not infer relocation from image similarities. CPU spans
remain logical and are never silently treated as physical RAM. A Memory label
requires an explicit physical location and `memoryLabel:true` on a region.
`length:0` is a symbol with unknown extent, not an empty known memory area.

Applicability is required **documentation text** in M1, not an evaluated predicate.
Exported labels retain this caveat and evidence status. Functions are research
entries only: M1 exports no runtime activation flag or breakpoint operation.
No code executes because an annotation exists.

Package size is capped at 8 MiB; nesting at 64, reference depth at 16; collections
at 4,096 entries, releases at 64, media roles at 32, fields per struct at 256.
Sizes/counts fit JSON's exact safe-integer range. Addresses use canonical lowercase
hex strings without leading zeroes. Bounds and multiplication are checked before
use, and type graph sizing memoizes shared subgraphs.

Source `path` is a reference, never fetched by the validator. Scheme/host URLs and
unsafe paths are rejected. Existing `/public/...` paths link to curated site
references for Memory; the Fort mechanics source identifies the repository
write-up and is not emitted as a Memory link. A future Code viewer must resolve
repository references appropriately rather than treating them as hosted files.

## Export and browser integration

`knowledgeexport` validates every package, sorts output deterministically and
emits data via `JSON.parse` of a quoted JSON string. Package strings cannot become
JavaScript expressions or object-literal prototype setters. The compiled output
contains the original structured knowledge and release-specific physical labels.
The content hash identifies the exact source bytes, including formatting.

`site/emulators/knowledge-model.js` currently recognizes only exact single-image
releases; multi-disc and file-set browser activation await a complete browser
identity flow. `memory-labels.js` uses this data for Fort and retains Sonic's
existing descriptor-based implementation until Sonic is migrated. An unknown,
wrong-sized, wrong-hash or ambiguous image gets no game-specific labels; hardware
regions remain available. The cache is keyed by both File and platform.

`tools/browser/package.py` regenerates knowledge before packaging, pins the data
with the release and keeps the previous executable bundle. Use `--site-only` to
reuse the already packaged core binaries after verifying their manifest hashes.
This avoids copying unrelated local core builds during annotation-only changes.
The public check runner tests knowledge and rejects stale generated exports.

## Updating research

1. Edit the package's definitions, exact release identity and evidence; increment
   `revision` for a published content change. Uncertain claims stay marked inferred
   or hypothesis. Do not create an invented field to fit a panel.
2. Run `knowledgecheck`, then optional matching-media verification. Add semantic
   negative cases or real derivation fixtures for new supported behavior.
3. Export and test. Keep generated data with its source changes; `-check` verifies
   byte-for-byte reproducibility. Package the site to update direct emulator routes.
4. Update the narrative write-up when findings change. Passing schema validation
   is never a substitute for checking the binary or reproducing game behavior.

Acceptance details: [M1 report](../browser/docs/code-m1/README.md).

## M3 live state additions

The C64 browser decodes bounded arrays/records, integer/enum/bitfield/fixed values,
and preserves raw bytes. State definitions can now include `relatedFunctions`
(release-compatible function IDs) and array `occupancy`, for example
`{"path":["type"],"notEquals":"0"}`. The path is relative to each element;
an empty path compares the scalar element. This is a bounded storage comparison,
not an expression or script. The authoring validator checks its type and range.
See [M3 implementation and limits](../browser/docs/code-m3/README.md).

## M4 read-only tours

`tours` supports validated sequential C64 investigations with byte signatures,
PC/masked-RAM conditions, assertions, budgets, layout hints and interval write
evidence. See the [M4 specification and acceptance report](../browser/docs/code-m4/README.md).
Tour conditions are data; applicability prose is still never evaluated.

## M5 DOS runtime knowledge

DOS file-set releases can pin an `executable`. Verified MZ/overlay modules,
module-relative and relocation-derived segment locations, mode/register/symbolic
tour conditions, and bounded contextual pixel buffers are now implemented.
See the [M5 profile and acceptance report](../browser/docs/code-m5/README.md).
Protected-mode decoding uses the core’s cached segment bases; runtime knowledge
modules currently support the verified real-mode Underworld profile.
