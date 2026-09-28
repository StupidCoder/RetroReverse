import {InputQueue} from '../../../site/emulators/input.js';
import assert from 'node:assert/strict';
const q=new InputQueue(60);
q.enqueue({buttons:8,request:1},0);q.enqueue({buttons:0,request:2},0);
assert.equal(q.drain(0).buttons,8);assert.equal(q.drain(1/60).buttons,0);
q.enqueue({buttons:0x80000000,request:3},1);assert.equal(q.drain(1).buttons,0x80000000);
q.enqueue({buttons:0,request:4},2);assert.equal(q.drain(2).buttons,0);
const k=new InputQueue(50);
for(const c of [76,79,65,68,13]){k.enqueue({buttons:0,keys:[[c,1]]},0);k.enqueue({buttons:0,keys:[[c,0]]},0);}
assert.deepEqual(k.drain(0).keys,[[76,1]]);
assert.deepEqual(k.drain(.06).keys,[[76,0]]);
assert.deepEqual(k.drain(.08).keys,[[79,1]]);
assert.equal(k.drain(1).keys.length,7);
console.log('Short controller taps, unsigned masks, and rapid C64 typing preserve ordered emulated-time edges.');
// A stylus tap must remain down for at least one emulated frame, even when both
// browser edges arrive before the slow emulator worker gets another turn.
const stylus=new InputQueue(60);
stylus.enqueue({buttons:0,touch:{x:120,y:80,down:true}},0);
stylus.enqueue({buttons:0,touch:{x:120,y:80,down:false}},0);
assert.equal(stylus.drain(0).touch.down,true);
assert.equal(stylus.drain(.02).touch.down,false);
// A 30 Hz game can poll only every second 60 Hz input sample. Handheld pulses
// must remain observable across that gap, even if press/release arrive together.
const handheld = new InputQueue(60);handheld.hold=3/60;
handheld.enqueue({buttons:1},0);handheld.enqueue({buttons:0},0);
assert.equal(handheld.drain(2/60).buttons,1);
assert.equal(handheld.drain(3/60).buttons,0);
