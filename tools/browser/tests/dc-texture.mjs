import assert from 'node:assert/strict';
import fs from 'node:fs';
import {decodeTexture,textureState,twiddle,color} from '../../../site/emulators/dc-texture-decode.js';
assert.equal(twiddle(1,0),2);assert.equal(twiddle(0,1),1);
const memory=new Uint8Array(8388608),word=(at,v)=>{memory[at]=v;memory[at+1]=v>>8;};
const info={known:true,pcw:8,isp:7<<29,tsp:0,tcw:0x1000/8,registersKnown:true,paletteFormat:3,palette:Array.from({length:256},(_,i)=>(0xff000000|i<<16|i<<8|i)>>>0),uv:[]};
word(0x1000+4,0xffff);let d=decodeTexture(memory,info);assert.deepEqual(d.sample(1,0).rgba,[248,248,248,255]);assert.deepEqual(d.sample(0,1).rgba,[0,0,0,0]);
info.tcw=(5<<27)|0x1000/8;memory[0x1000]=0x21;d=decodeTexture(memory,info);assert.equal(d.sample(0,0).index,1);assert.equal(d.sample(0,1).index,2);assert.equal(d.sample(0,1).paletteAddress,0x5f9008);assert.deepEqual(d.auxiliary().rgba.slice(4,8),new Uint8ClampedArray([1,1,1,255]));
info.tcw=(6<<27)|0x1000/8|(2<<25);d=decodeTexture(memory,info);assert.equal(d.state.paletteBase,512);assert.equal(d.sample(0,0).paletteAddress,0x5f9000+(512+33)*4);
info.tcw=(1<<30)|(1<<27)|0x1000/8;memory[0x1800]=7;word(0x1000+7*8+4,0xf800);d=decodeTexture(memory,info);assert.equal(d.sample(1,0).entry,7);assert.equal(d.sample(1,0).indexAddress,0x1800);assert.deepEqual(d.sample(1,0).rgba,[248,0,0,255]);assert.deepEqual(d.dictionarySample(15,0).rgba,[248,0,0,255]);assert.equal(d.indexSample(0,0).entry,7);assert.equal(d.storage().width,4);
assert.equal(textureState({...info,tcw:info.tcw|0x80000000}).indexOffset,6);
assert.deepEqual(color(0x0123,2),[17,34,51,0]);assert.deepEqual(color(0x80402010,3),[64,32,16,128]);
assert.equal(decodeTexture(memory,{...info,tcw:0x1fffff|(1<<27)}).sample(7,7).valid,false);
assert.equal(decodeTexture(memory,{...info,tcw:3<<27}).state.supported,false);
assert.equal(decodeTexture(memory,{...info,tcw:5<<27,registersKnown:false}).sample(0,0).valid,false);
if(process.argv[2]){
 for(let i=0;i<memory.length;i++)memory[i]=(i*37+(i>>8)*13)^((i>>4)&255);
 let cases=0,samples=0;for(const line of fs.readFileSync(process.argv[2],'utf8').trim().split('\n')){const i=JSON.parse(line),d=decodeTexture(memory,i);for(const [x,y,...rgba]of i.expected){assert.deepEqual(d.sample(x,y).rgba,rgba,JSON.stringify({case:cases,state:d.state,x,y}));samples++;}cases++;}console.log(`Dreamcast JS decoder matches native sampler: ${cases} formats/layouts, ${samples} texels`);
}
console.log('Dreamcast texture addressing, palette, VQ, mip offsets and alpha tests pass');
