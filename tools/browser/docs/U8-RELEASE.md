# U8 — static Pages release and validation

The existing site remains the Pages output directory (`site`). Its extracted
assets, explanation text and routes are unchanged. The emulator gallery is
linked from the Studio welcome view. Direct routes are `/emulators/c64/`,
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
The current public check suite passes. Four 30-minute Node/WASM soak sessions
exercise continuous execution, capture/replay, periodic state loads and fresh
core replacement. Their final results will be recorded after completion; they
are core longevity tests, not substitutes for browser end-to-end acceptance.

Chromium browser checks cover local media, hosted C64 firmware, state import,
controls, capture, pixel inspection and replay. Normal C64 presentation was
observed at about 50 frames/s and Ridge Racer at about 60 presented frames/s after
the frame-pacing fix. See `FRAME-PACING.md` for the cause and measurements. PS1
capture now includes four synthetic fields of double-buffer producer context. Browser-specific
results, deployed checks and the soak outcome are recorded in the final update.

**Release gates not yet established:** Firefox and Safari workflows, a physical
standard gamepad session, and download-to-disk plus import from that actual
browser-produced file across reload. Available automation exposes Chromium;
its download/file-chooser restrictions are documented in U4. These gaps are not
silently counted as passing. Automated state-container round trips and native/
WASM continuation do pass, but establish a narrower result.
