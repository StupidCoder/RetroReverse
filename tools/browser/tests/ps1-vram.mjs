import assert from 'node:assert/strict';
import {decodeVRAM} from '../../../site/emulators/ps1-vram-decode.js';
const b=new Uint16Array(1024*512),s={depth:0,clut:400<<6,pageX:64,pageY:256,window:[0,0,0,0]};
b[400*1024+1]=31;b[400*1024+2]=0x3e0;b[400*1024+3]=0x7c00;b[400*1024+4]=0x8000;b[256*1024+64]=0x4321;
const pixel=(bytes,n)=>[...bytes.slice(n*4,n*4+4)];
let d=decodeVRAM(b.buffer,s);assert.equal(d.width,4096);assert.deepEqual([0,1,2,3].map(n=>pixel(d.page,n)),[[248,0,0,255],[0,248,0,255],[0,0,248,255],[0,0,0,255]]);assert.equal(d.pageSample(3,0).address,(256*1024+64)*2);assert.equal(d.pageSample(3,0).paletteAddress,(400*1024+4)*2);assert.equal(d.pageSample(4,0).transparent,true);
assert.deepEqual(d.overviewSample(64*4+2,256).rgba,d.pageSample(2,0).rgba);
s.depth=1;b[256*1024+64]=0x0201;d=decodeVRAM(b.buffer,s);assert.equal(d.width,2048);assert.deepEqual(pixel(d.page,1),[0,248,0,255]);
s.depth=2;b[256*1024+64]=0x7c00;d=decodeVRAM(b.buffer,s);assert.equal(d.width,1024);assert.deepEqual(pixel(d.page,0),[0,0,248,255]);
// Window mask remaps UV bits; direct sampling wraps horizontally at VRAM edge.
s.window=[1,1,1,1];b[(256+8)*1024+64+8]=31;d=decodeVRAM(b.buffer,s);assert.deepEqual(pixel(d.page,0),[248,0,0,255]);assert.equal(d.pageSample(0,0).u,8);assert.equal(d.pageSample(0,0).v,8);
d=decodeVRAM(b.buffer,s,{window:false});assert.deepEqual(pixel(d.page,0),[0,0,248,255]);
s.window=[0,0,0,0];s.pageX=960;b[256*1024]=0x3e0;d=decodeVRAM(b.buffer,s);assert.equal(d.pageSample(64,0).x,0);assert.deepEqual(pixel(d.page,64),[0,248,0,255]);
// Index 0 is not intrinsically transparent on PS1: the resolved color decides.
s.depth=0;s.clut=400<<6;s.pageX=64;b[256*1024+64]=0;b[400*1024]=31;d=decodeVRAM(b.buffer,s);assert.equal(d.pageSample(0,0).transparent,false);assert.deepEqual(pixel(d.page,0),[248,0,0,255]);
s.clut=-1;b[256*1024+64]=15;d=decodeVRAM(b.buffer,s);assert.equal(d.hasPalette,false);assert.deepEqual(pixel(d.page,0),[255,255,255,255]);assert.equal(d.pageSample(0,0).paletteAddress,null);
d=decodeVRAM(b.buffer,s,{mode:'2'});assert.equal(d.depth,2);assert.equal(d.width,1024);
assert.throws(()=>decodeVRAM(new ArrayBuffer(1),s),/captured VRAM/);
console.log('PS1 VRAM decode: packed 4/8-bit indices, CLUT resolution, RGB555, transparency, UV window, wrap and explicit color overrides pass');
