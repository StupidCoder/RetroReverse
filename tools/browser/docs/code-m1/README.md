# M1 — knowledge schema, validator and first package

Completed 2026-09-30 for the planned basic profile. No instruction debugger or
live structured-state UI is claimed. M2 is the next milestone.

## Delivered

- [Fort knowledge package](../../../../games/fort-apocalypse-c64/knowledge.json):
  exact tape identity, evidence, address spaces/locations, 13 migrated Memory
  labels, six documented state values, five function entries, and a cautionary
  annotation on the upward probe. Tours/assets/experiments are explicitly empty.
- [Schema and authoring tools](../../../knowledge/README.md): embedded strict
  schema plus semantic validation, canonical file-set identity, static resolution,
  type bounds, safe enum fallback and deterministic export/check commands.
- Browser knowledge data and exact single-image recognition. Fort's addresses
  were removed from hand-authored Memory JavaScript. Sonic remains on its existing
  descriptor-based path; it was not silently converted to incomplete knowledge.
- Packager regeneration and stale-export CI checks. A `--site-only` mode verifies
  existing binaries and repackages code/data without rebuilding/copying local cores.
- Production tests consuming M0 identity vectors, negative schema/semantic cases,
  unknown-image generic Memory regression, and an actual-browser worker fixture.

## Validation

Run from the repository root:

```sh
go test ./tools/knowledge ./tools/cmd/knowledgecheck ./tools/cmd/knowledgeexport
go vet ./tools/knowledge ./tools/cmd/knowledgecheck ./tools/cmd/knowledgeexport

go run ./tools/cmd/knowledgecheck -media tape=games/fort-apocalypse-c64/Fort_Apocalypse.tap \
  -release reference-pal games/fort-apocalypse-c64/knowledge.json

go run ./tools/cmd/knowledgeexport -check -out site/emulators/knowledge-data.js \
  games/fort-apocalypse-c64/knowledge.json

node tools/browser/tests/knowledge.mjs games/fort-apocalypse-c64/Fort_Apocalypse.tap
node tools/browser/tests/memory.mjs
python3 tools/browser/package.py --site-only
python3 tools/browser/check-release.py
```

All completed successfully. The Go test package contains semantic tests; the two
CLI packages compile under `go test` and are smoke-tested through the commands
above. No commercial bytes are needed for the Go tests or the default Node test.
The private-media commands were also run with the available reference tape.

Serve the repository on loopback and open
`/tools/browser/tests/knowledge-ui.html?fort=1`. The **packaged release worker**
returned 13 Fort labels and six physical regions; a synthetic unknown tape
returned zero labels and retained six regions. The page displayed PASS. Without
`?fort=1` this fixture is media-free and skips the local reference tape.

[Recorded acceptance](../../results/code-m1/acceptance.json) pins the package hash
and release. The generated bundle is `a164c618f7aa82e20dbb`, with 73 pinned assets.
All existing WASM hashes were retained. No deployment or git push was performed.

## M1 exit review

- [x] Schema, strict JSON handling and semantic reference/bounds validation.
- [x] Initial Fort package validates and matches the exact supplied local image.
- [x] M0 normalization vectors pass through the production Go implementation.
- [x] Bad hashes, missing references, unknown operations, invalid layouts,
  applicability omissions and ambiguity are rejected.
- [x] Unknown enum values remain numeric; unknown images retain generic inspection.
- [x] Browser export is reproducible and a stale export fails checking.
- [x] Existing applicable Fort Memory labels come from the single package.

## Scope and remaining work

M1 supports CPU/physical/image/relative static locations and basic fixed-layout
metadata. Unsupported complex types, asset location chains and executable tour
operations fail closed; they are not partially activated. See the tools README
for the exact profile and resource limits. The private schema evaluator supports
only the keywords used by the embedded standard schema and rejects unimplemented
keywords. A separate system Python JSON Schema implementation was unavailable;
no independent conformance-library result is claimed.

Applicability is explanatory text, not a detected phase. Fort research remains
inferred and its bug explanation requires reconciliation against live bytes.
No claims were upgraded to confirmed merely by migrating labels or matching the
image hash. Multi-disc/file-set identities are supported in authoring tools;
the browser currently activates single-image packages only.

M2 must implement accurate instruction boundaries, stepping and stop conditions.
M3 consumes structured state definitions in live panels. M4/M8 add actual tours
and mutations. The repository's all-platform native suite and cross-browser matrix
were not rerun for this annotation/tooling change; targeted Memory, knowledge and
release-integrity checks are the validation scope.
