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
