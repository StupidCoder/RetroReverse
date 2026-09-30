// node tools/browser/tests/code-m0-audit.mjs > tools/browser/results/code-m0/core-audit.json
import fs from 'node:fs';
import crypto from 'node:crypto';
import os from 'node:os';
import {execFileSync} from 'node:child_process';
import {probeC64} from './code-m0-probe.mjs';
const root = new URL('../../../', import.meta.url);
const sha = b => crypto.createHash('sha256').update(b).digest('hex');
const platforms = ['c64','gb','gg','gba','amiga','ps1','n64','3do','ds','3ds','psp','ps2','gc','dc','dos','xbox'];
const exportsByPlatform = {};
for (const p of platforms) {
  const bytes = fs.readFileSync(new URL(`site/emulators/cores/${p}/core.wasm`, root));
  const names = WebAssembly.Module.exports(await WebAssembly.compile(bytes)).map(x => x.name);
  const glue = fs.readFileSync(new URL(`site/emulators/cores/${p}/core.js`, root), 'utf8');
  const has = n => names.includes(n) || names.includes('_' + n) || glue.includes('Module["_' + n + '"]');
  exportsByPlatform[p] = {sha256: sha(bytes), bytes: bytes.length,
    glueSHA256: sha(glue), rawExportCount: names.length,
    exports: Object.fromEntries(['rr_status','rr_run','rr_run_slice','rr_state_save','rr_state_load',
      'rr_inspect_regions','rr_inspect_data','rr_activity_begin','rr_bus','rr_bus_flags','rr_stop_reason']
      .map(n => [n, has(n)]))};
}
const factory = (await import(new URL('site/emulators/cores/c64/core.js', root))).default;
const core = await factory({wasmBinary: fs.readFileSync(new URL('site/emulators/cores/c64/core.wasm', root))});
console.log(JSON.stringify({schema: 1, date: new Date().toISOString(),
  revision: execFileSync('git',['rev-parse','HEAD'],{encoding:'utf8'}).trim(),
  environment: {node: process.version, platform: process.platform, arch: process.arch, cpu: os.cpus()[0].model},
  note: 'Export presence proves linkage only, not semantic support or game compatibility.',
  exportsByPlatform, c64: probeC64(core)}, null, 2));
