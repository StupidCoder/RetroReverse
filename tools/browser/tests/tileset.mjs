import assert from 'node:assert/strict';
import {decodeTileset,tilesetBytes} from '../../../site/emulators/tileset-decode.js';
const data=p=>new Uint8Array(tilesetBytes[p]);
const pixel=(s,x,y)=>[...s.pixels.slice((y*s.width+x)*4,(y*s.width+x)*4+4)];
// C64: single/multicolor is chosen per character cell, not just D016.
const c=data('c64');for(let i=0;i<16;i++)c.set([i,i,i,255],80+i*4);
c[0x21]=1;c[0x22]=2;c[0x23]=3;c[65]=0x20;c[67]=4;c[144]=10;c[145]=7;c[1169]=1;c[2192]=0x1b;c[2200]=0x80;
let s=decodeTileset('c64',c);assert.deepEqual(pixel(s,0,0),[1,1,1,255]);assert.deepEqual(pixel(s,3,0),[10,10,10,255]);
c[0x16]=0x10;s=decodeTileset('c64',c);assert.deepEqual([0,2,4,6].map(x=>pixel(s,x,0)[0]),[1,2,3,2]);assert.deepEqual(pixel(s,8,0),[7,7,7,255]);assert.deepEqual(pixel(s,9,0),[1,1,1,255]);
s=decodeTileset('c64',c,{context:'cell',cell:1});assert.deepEqual(pixel(s,3,0),[7,7,7,255]);
c[0x16]=0;c[0x11]=0x40;c[0x22]=5;c[2192]=0;s=decodeTileset('c64',c);assert.equal(pixel(s,0,32)[0],5);assert.match(s.describe(64),/\$2000/); // ECM code64 reuses glyph0, BG1.
const before=decodeTileset('c64',c).pixels.slice();c[2192]=255;assert.notDeepEqual(decodeTileset('c64',c).pixels,before);
// DMG two bitplanes, palette remapping, signed IDs, transparent object zero.
const gb=data('gb');gb[0]=0x80;gb[1]=0x40;gb[8192+0x47]=0xe4;gb[8192+0x48]=0x1b;
s=decodeTileset('gb',gb);assert.equal(s.count,384);assert.deepEqual([0,1,2].map(x=>pixel(s,x,0)[0]),[170,85,255]);assert.match(s.describe(128),/\$80/);s=decodeTileset('gb',gb,{palette:1});assert.equal(pixel(s,2,0)[3],0);assert.equal(pixel(s,0,0)[0],85);
// GG: planar significance, 12-bit RGB, both captured palette banks.
const gg=data('gg');gg.set([0x80,0x40,0x20,0x10]);for(const [idx,c]of [[1,0x00f],[2,0x0f0],[4,0xf00],[8,0xfff],[17,0x123]]){gg[16384+idx*2]=c;gg[16385+idx*2]=c>>8;}
s=decodeTileset('gg',gg);assert.deepEqual([0,1,2,3].map(x=>pixel(s,x,0)),[[255,0,0,255],[0,255,0,255],[0,0,255,255],[255,255,255,255]]);assert.deepEqual(pixel(decodeTileset('gg',gg,{palette:1}),0,0),[51,34,17,255]);
// GBA: low nibble first, charbase selection, 8bpp/affine, object bank.
const gba=data('gba');gba[8]=4;gba[8192+0x4000]=0x21;gba[4096+3*32+2]=31;gba[4096+3*32+4]=0xe0;gba[4096+3*32+5]=3;
s=decodeTileset('gba',gba,{palette:3});assert.equal(s.count,512);assert.deepEqual(pixel(s,0,0),[255,0,0,255]);assert.deepEqual(pixel(s,1,0),[0,255,0,255]);assert.equal(pixel(s,2,0)[3],0);assert.match(s.describe(0),/\$06004000/);
gba[8]|=128;gba[4096+0x21*2+1]=0x7c;s=decodeTileset('gba',gba,{palette:3});assert.equal(s.count,256);assert.deepEqual(pixel(s,0,0),[0,0,255,255]);
gba[0]=2;gba[12]=0;gba[8192]=0x21;s=decodeTileset('gba',gba,{source:2});assert.equal(s.count,256);assert.deepEqual(pixel(s,0,0),[0,0,255,255]);
gba[8192+0x10000]=1;gba[4096+512+2]=31;s=decodeTileset('gba',gba,{source:4});assert.deepEqual(pixel(s,0,0),[255,0,0,255]);assert.match(s.describe(0),/\$06010000/);
assert.throws(()=>decodeTileset('gb',new Uint8Array(2)),/unavailable/);
console.log('Tileset decoding: C64 per-cell modes/ECM, GB palettes/signed IDs, GG bitplanes/palettes, GBA banks/depth/affine/OBJ pass');
