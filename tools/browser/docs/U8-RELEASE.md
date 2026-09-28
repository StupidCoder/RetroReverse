# U8 — static Pages release and validation

The existing site remains the Pages output directory (`site`). Its extracted
assets, explanation text and routes are unchanged. The emulator gallery is
linked from the Studio welcome view. Production Pages host: `https://aireverseengineering.pages.dev/` (confirmed from
the repository's Cloudflare deployment check). Direct routes are `/emulators/c64/`,
`/emulators/ps1/`, `/emulators/n64/`, and `/emulators/3do/`.

## Reproducible packaging

`release.json` pins Emscripten 4.0.16, C++20 and the build/deploy contract.
Run `python3 tools/browser/build-release.py --emcc /path/to/em++` to rebuild
all four checked-in C++ cores and package them. Go regeneration is optional and
not part of the release build. The Pages integration can continue serving the
checked-in artifacts with no emulator build; its optional validation build
command is `python3 tools/browser/check-release.py`, output directory `site`.
No Worker/Pages Function or other server process is required.

Packaging creates a content-addressed executable bundle under
`site/emulators/releases/<hash>/`. Each emulator HTML page loads that bundle's
app and CSS; its worker, helper modules, WASM/JS and firmware resolve inside the
same bundle. The manifest hashes every file. WASM and firmware bytes are checked
before use. Packaging retains the previous bundle for already-open pages; tabs
older than that may need a reload. The emulator subtree uses `Cache-Control:
no-cache`, COOP/COEP and nosniff through the existing `_headers`. There is no
service worker or stale offline cache. Native Pages content types are used for
JavaScript, WASM and HTML. Deployment header/MIME checks are recorded separately.

`check.py` runs the public media-free JS/native tests, syntax checks and release
integrity checks. `.github/workflows/emulators.yml` runs it on emulator changes.
Local reference-game tests require user-supplied media and are not uploaded to CI.
A portable state remains bound to the exact core build and selected media; this
release does not migrate old states.

Cloudflare's [static HTML guidance](https://developers.cloudflare.com/pages/framework-guides/deploy-anything/),
[build configuration](https://developers.cloudflare.com/pages/configuration/build-configuration/)
and [headers documentation](https://developers.cloudflare.com/pages/configuration/headers/)
were checked for this packaging arrangement. The existing Git integration is
preserved rather than introducing a second deployment project.

## Network and compatibility contract

The emulator's application code has no remote emulator API, fixture endpoint,
localhost dependency, upload, analytics report, socket or beacon. Disc bytes
are read from mounted local Files using worker-local synchronous slices; media
hashes are streamed in 1 MiB chunks. Known-image identification reuses the media
hash for state validation rather than hashing the same disc twice. Fetches are
for executable assets, manifests and hosted C64 firmware. Generated module
loaders fetch only their matching local WASM bundle. Artwork is hosted locally.
The existing Studio's CDN/asset requests are intentionally unaffected.

The user-facing format/game/model/browser matrix is published at
`/emulators/compatibility.html`. Unknown images remain selectable; unsupported
layouts return an error. Browser capability and game compatibility are separate.

## Validation status

See U4–U7 acceptance documents for state/capture/replay correctness and limitations.
The current public check suite passes. Four 30-minute Node/WASM soak sessions completed successfully, each with 60
single-interval capture/replay checks, 35 state restores and five fresh core
replacements. Results are in `results/u8-soak-*.jsonl`. These are core longevity
tests, not browser end-to-end tests. PS1's extended four-field browser policy is
also stress-tested separately: 70 four-field capture/replay cycles in one
instance passed, with a 155 MiB WASM high-water mark. WASM memory high-water marks were 128 MiB C64,
90.75 MiB PS1, 110.63 MiB N64 and 232.25 MiB 3DO in the 30-minute runs. Repeated
cycles returned to comparable ranges; linear memory does not shrink after peak
allocation. Node RSS includes a complete disc buffer supplied by the test
harness, unlike the browser's bounded local File reads.

Chromium browser checks cover local media, hosted C64 firmware, state import,
controls, capture, pixel inspection and replay. Normal C64 presentation was
observed at about 50 frames/s and Ridge Racer at about 60 presented frames/s after
the frame-pacing fix. See `FRAME-PACING.md` for the cause and measurements. PS1
capture now includes four synthetic fields of double-buffer producer context. Browser-specific
results, deployed checks and the soak outcome are recorded in the final update.

## Hosted and browser results

Cloudflare Pages and GitHub's `static-and-core` CI check passed for the release.
`check-deployed.py` verified all 23 executable/firmware asset hashes, all four
direct routes, JavaScript/WASM MIME types and COOP/COEP/no-cache headers on the
production host. Results are in `results/u8-deployed.json`.

Hosted Chromium: Pilotwings initialized from a local cartridge in about 0.27 s,
ran to display 400, captured in about 0.23 s, inspected a reconstructed pixel
and sought First/Last (cached end seek about 0.6 ms). Save produced an actual
4,214,072-byte `.rrstate` file in Downloads; its container integrity passed and
it restored display 400. A full page reload, cartridge reselection and import of that actual downloaded
file again restored display 400 (about 0.28 s initialization/verification).
Local Chromium: Ridge Racer captured four fields in about 0.23 s and followed
a texture source's historical bytes. A contributor jump took about 15.3 ms.
C64 capture/replay and normal 50 Hz presentation also passed.

Safari, using native application control on the deployed URL: C64 initialized
with automatically hosted firmware in about 0.35 s, restored a portable Fort
state in about 0.19 s, ran at 100% PAL / about 50.2 presented frames/s, captured
in about 0.04 s and inspected/replayed the raster evidence (about 0.2 ms seek).
This establishes that tested Safari path, not all four Safari cores.

**Remaining release-validation gates:** Firefox was unavailable as an installed
controllable application; remaining Safari/game combinations, full four-platform
hosted workflows, physical gamepad input and the complete multi-file-disc browser
matrix have not all been exercised. Media-free CUE tests pass. These gaps are
explicitly published rather than inferred from native/WASM tests. U8's packaging
and deployment implementation is complete; its exhaustive browser release gate
is still partial.

State saving also blocks concurrent Run/Step requests while media hashing and
serialization finish; the UI restores its controls on completion or failure.
This prevents a long first save from racing a resumed machine.
