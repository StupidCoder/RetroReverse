import {sha256File} from './media.js';
const cached=new WeakMap();
const sonicHash='dd605c17c8d24b2db9928beff6d8294fd84184b31c2fa92760ae691db84fd569';
const fortHash='9e444c4576bac52ba691f0ffe2c0a7efb0f62f3fe2be7cbe78dba08672dda00b';
const sonicSource='https://github.com/StupidCoder/RetroReverse/blob/main/games/sonic-gg/sonic-gg.md#4-level-maps-how-a-zone-is-stored-and-drawn';
// Source paths resolve through the site's checked-in reference pages.
export function sonicLabels(rom){
 const out=[],word=p=>rom[p]|rom[p+1]<<8;
 const add=(name,start,length,description)=>{
  if(start<0||start+Math.max(1,length)>rom.length)throw Error('Annotation exceeds ROM');
  let n=Math.max(1,length),part=0;while(n){const bank=Math.floor(start/16384),offset=start%16384,size=Math.min(n,16384-offset);out.push({name:name+(part?' (continued)':''),region:'rom-'+bank,start:offset,length:length?size:0,description,source:sonicSource});start+=size;n-=size;part++;}
 };
 const names=['Green Hills','Bridge','Jungle','Labyrinth','Scrap Brain','Sky Base'];
 add('Act descriptor pointers',0x15600,36,'18 relative pointers. Reference: Sonic Part IV, act descriptor field map.');
 for(let i=0;i<18;i++){
  const name=names[Math.floor(i/3)]+' Act '+(i%3+1),d=0x15600+word(0x15600+i*2);
  add(name+' descriptor',d,37,'Verified 37-byte act resource descriptor.');
  add(name+' compressed map',0x14000+word(d+15),word(d+17),'RLE input; expands into the 4096-byte RAM map.');
  add(name+' block table',0x10000+word(d+19),0,'Block-to-tile table symbol; extent not inferred.');
  add(name+' tile stream',0x30000+word(d+21),0,'Compressed tile stream symbol; extent not inferred.');
 }
 out.push({name:'Expanded level map (gameplay)',region:'ram',start:0,length:4096,description:'Gameplay map window at CPU $C000; loading phases may use this memory differently.'},
  {name:'Mapper shadows',region:'ram',start:0x122f,length:2,description:'CPU $D22F/$D230: saved slot-1 and slot-2 ROM banks.'},
  {name:'Map stride',region:'ram',start:0x1232,length:2,description:'CPU $D232: map row stride.'});
 return out;
}
export function fortLabels(){
 const ranges=[['Expanded terrain (gameplay)',0x0503,0x2d02],['Scanner bitmap (gameplay)',0x2e00,0x343f],['Sprite blocks',0x4000,0x43ff],['Screen RAM (gameplay)',0x4400,0x47e7],['HUD charset',0x5000,0x57ff],['Playfield charset',0x5800,0x5fff],['Level 0 compressed map',0x7000,0x762a],['Level 1 compressed map',0x762b,0x7d36],['Level 0 compressed scanner',0x8000,0x81e8],['Level 1 compressed scanner',0x81e9,0x84ee],['Packed sprite shapes',0x870f,0x8906]];
 return [...ranges.map(([name,start,end])=>({name,region:'ram',start,length:end-start+1,description:'Fort reference: documented game memory; contents depend on boot/loading phase.',source:'/public/fort-apocalypse-c64/docs/architecture.html'})),
 {name:'RLE decompressor',region:'ram',start:0x8cdb,length:0,description:'Routine entry; reads compressed streams and writes the selected output region.',source:'/public/fort-apocalypse-c64/docs/playfield.html'},
 {name:'Tape pulse stream',region:'tape',start:20,length:0,description:'Raw TAP pulse encodings, not decoded game bytes.',source:'/public/fort-apocalypse-c64/docs/tape.html'}];
}
export async function memoryLabels(platform,file){
 if(!file||!['gg','c64'].includes(platform))return [];
 if(cached.has(file))return cached.get(file);
 const task=(async()=>{if(platform==='gg'&&file.size===262144&&await sha256File(file)===sonicHash)return sonicLabels(new Uint8Array(await file.arrayBuffer()));if(platform==='c64'&&file.size===225817&&await sha256File(file)===fortHash)return fortLabels();return [];})();cached.set(file,task);return task;
}
