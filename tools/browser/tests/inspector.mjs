import assert from 'node:assert/strict';
import {pixelCoordinates,decodedColor} from '../../../site/emulators/inspector.js';
const rect={left:10,top:20,width:800,height:600};
assert.deepEqual(pixelCoordinates(rect,320,240,410,320),{x:160,y:120});
assert.equal(pixelCoordinates(rect,392,272,410,22),null); // C64 letterbox, not pixel 0
assert.deepEqual(pixelCoordinates(rect,392,272,410,320),{x:196,y:136});
assert.equal(pixelCoordinates(rect,320,240,810,320),null); // right edge is excluded
assert.deepEqual(decodedColor('ps1',31,2),[248,0,0,255]);
assert.deepEqual(decodedColor('n64',0x01f8,2),[248,0,0,255]);
assert.deepEqual(decodedColor('3do',0x007c,2),[255,0,0,255]);
assert.deepEqual(decodedColor('n64',0x12345678,4),[120,86,52,255]);
console.log('Coordinate scaling/letterboxing and modeled scanout colors passed');

assert.deepEqual(decodedColor('ds',0xff332211,4),[17,34,51,255]);
assert.deepEqual(decodedColor('3ds',0x11223344,4,0),[17,34,51,255]);
assert.deepEqual(decodedColor('3ds',0x112233,3,1),[17,34,51,255]);

assert.deepEqual(decodedColor('psp',0x03bf,2,0),[255,117,0,255]);
assert.deepEqual(decodedColor('psp',0x7c1f,2,1),[255,0,255,255]);
assert.deepEqual(decodedColor('psp',0x4321,2,2),[17,34,51,255]);
assert.deepEqual(decodedColor('psp',0xaa332211,4,3),[17,34,51,255]);

for(const platform of ['gb','gg'])assert.deepEqual(decodedColor(platform,0xff332211,4),[17,34,51,255]);

assert.deepEqual(pixelCoordinates({left:10,top:20,width:500,height:400},640,256,260,220,true),{x:320,y:128});
assert.deepEqual(decodedColor('amiga',0xff332211,4),[17,34,51,255]);

assert.deepEqual(decodedColor('ps2',0x80332211,4),[17,34,51,255]);
assert.equal(decodedColor('gc',0x11223344,4,'EFB RGBA'),null);
