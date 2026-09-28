import {sha256File,selectMedia} from '../../../site/emulators/media.js';
import {createHash,randomBytes} from 'node:crypto';
import assert from 'node:assert/strict';
for(const n of [0,1,55,56,63,64,65,1048577]){const b=randomBytes(n);assert.equal(await sha256File(new Blob([b])),createHash('sha256').update(b).digest('hex'));}
const bin=new File([new Uint8Array(2048*20)],'track.bin');
assert.equal((await selectMedia([bin])).size,bin.size);
const cue=new File(['FILE "track.bin" BINARY\n TRACK 01 MODE1/2048\n INDEX 01 00:00:00'],'test.cue');
assert.equal((await selectMedia([cue,bin])).size,bin.size);
await assert.rejects(selectMedia([cue]),/Missing/);
await assert.rejects(selectMedia([bin,new File(['x'],'other.bin')]),/Select one/);
console.log('SHA-256 boundary vectors and CUE validation passed');
