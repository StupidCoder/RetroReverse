import assert from 'node:assert/strict';
import {selectDreamcastMedia} from '../../../site/emulators/dc-media.js';
const raw=new Uint8Array(2352*40);function header(offset,lba){raw.set([0,...Array(10).fill(255),0],offset);let fad=lba+150;const bcd=n=>(Math.floor(n/10)<<4)|(n%10);raw.set([bcd(Math.floor(fad/4500)),bcd(Math.floor(fad/75)%60),bcd(fad%75),1],offset+12);}
header(0,0);header(2352*20,45000);const bin=new File([raw],'game.bin');
const cue=new File(['CD_ROM\nTRACK MODE1_RAW\nDATAFILE "game.bin" 00:00:20\nTRACK AUDIO\nDATAFILE "game.bin" #23520\nTRACK MODE1_RAW\nDATAFILE "game.bin" #47040'],'game.cue');
const d=await selectDreamcastMedia([cue,bin]);assert.deepEqual(d.tracks.map(t=>[t.offset,t.length,t.lba]),[[0,23520,0],[23520,23520,-1],[47040,47040,45000]]);assert.equal(d.file,bin);
const standard=new File(['FILE "game.bin" BINARY\nTRACK 01 MODE1/2352\nINDEX 01 00:00:00\nTRACK 02 AUDIO\nINDEX 01 00:00:10\nTRACK 03 MODE1/2352\nINDEX 01 00:00:20'],'game.cue');assert.deepEqual((await selectDreamcastMedia([standard,bin])).tracks,d.tracks);
await assert.rejects(selectDreamcastMedia([cue]),/companion/);await assert.rejects(selectDreamcastMedia([bin]),/CUE/);await assert.rejects(selectDreamcastMedia([cue,new File([new Uint8Array(raw.length)],'game.bin')]),/header/);await assert.rejects(selectDreamcastMedia([new File(['CD_ROM\nTRACK MODE1_RAW\nDATAFILE "game.bin" #999999'],'bad.cue'),bin]),/extent/);
console.log('Dreamcast media bounds, track mapping and companion checks pass');
