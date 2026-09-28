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

const mouse=new InputQueue(50);mouse.enqueue({buttons:32,mouse:{x:300,y:-200}},0);mouse.enqueue({buttons:0},0);
assert.deepEqual(mouse.mouseForFrame(0),{x:127,y:-127});assert.deepEqual(mouse.mouseForFrame(0),{x:0,y:0});assert.deepEqual(mouse.mouseForFrame(1),{x:127,y:-73});assert.deepEqual(mouse.mouseForFrame(2),{x:46,y:0});assert.equal(mouse.drain(0).buttons,32);assert.equal(mouse.drain(.021).buttons,0);
assert.throws(()=>mouse.enqueue({mouse:{x:NaN,y:0}},0),/Invalid mouse/);

const clicks=new InputQueue(50);clicks.mouseMask=96;
for(const buttons of [32,0,32,0])clicks.enqueue({buttons},0);
assert.equal(clicks.drain(0).buttons,32);assert.equal(clicks.drain(.021).buttons,0);assert.equal(clicks.drain(.041).buttons,32);assert.equal(clicks.drain(.061).buttons,0);
