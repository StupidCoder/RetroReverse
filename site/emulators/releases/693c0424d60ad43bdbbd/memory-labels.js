import {sha256File} from './media.js';
import {identifySingleImage,hasSingleImageCandidate} from './knowledge-model.js';
const cached=new WeakMap();
const sonicHash='dd605c17c8d24b2db9928beff6d8294fd84184b31c2fa92760ae691db84fd569';
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
export async function memoryLabels(platform,file){
 if(!file)return [];
 let byPlatform=cached.get(file);
 if(!byPlatform){byPlatform=new Map();cached.set(file,byPlatform);}
 if(byPlatform.has(platform))return byPlatform.get(platform);
 const task=(async()=>{
  const sonic=platform==='gg'&&file.size===262144;
  if(!sonic&&!hasSingleImageCandidate(platform,file.size))return [];
  const sha256=await sha256File(file);
  if(sonic&&sha256===sonicHash)return sonicLabels(new Uint8Array(await file.arrayBuffer()));
  return identifySingleImage(platform,{size:file.size,sha256}).labels;
 })();
 byPlatform.set(platform,task);
 try{return await task;}catch(error){byPlatform.delete(platform);throw error;}
}
