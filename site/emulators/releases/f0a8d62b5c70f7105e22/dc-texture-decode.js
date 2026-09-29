// Match the current PowerVR software sampler. All addresses use the core's
// 64-bit VRAM layout, not the CPU framebuffer aperture's bank interleave.
export const textureFormats=['ARGB1555','RGB565','ARGB4444','YUV422','Bump map','4-bit indexed','8-bit indexed','Reserved'];
export function twiddle(x,y){let n=0;for(let b=0;b<10;b++)n|=((x>>b)&1)<<(2*b+1)|((y>>b)&1)<<(2*b);return n;}
export function color(word,format){switch(format){case 0:return [(word>>10&31)<<3,(word>>5&31)<<3,(word&31)<<3,(word>>>15&1)*255];case 1:return [(word>>11&31)<<3,(word>>5&63)<<2,(word&31)<<3,255];case 2:return [(word>>8&15)*17,(word>>4&15)*17,(word&15)*17,(word>>12&15)*17];case 3:return [word>>>16&255,word>>>8&255,word&255,word>>>24];default:return [255,0,255,255];}}
export function textureState(info){
 const {pcw=0,isp=0,tsp=0,tcw=0}=info,format=tcw>>>27&7,paletted=format===5||format===6,mipmap=!!(tcw>>>31),vq=!paletted&&!!(tcw&0x40000000);
 const width=8<<(tsp>>>3&7),height=mipmap?width:8<<(tsp&7),base=(tcw&0x1fffff)*8;
 let address=base,indexOffset=0;if(mipmap){if(vq){for(let n=width>>1;n;n>>=1)indexOffset+=Math.max(1,n*n/4);}else{const texels=(width*width-1)/3+3;address+=format===5?texels/2:format===6?texels:texels*2;}}
 return {known:!!info.known,textured:!!(pcw&8),width,height,base,address,indexOffset,format,paletted,vq,mipmap,twiddled:paletted||vq||!(tcw&0x4000000),paletteBase:format===5?(tcw>>>21&63)*16:(tcw>>>25&3)*256,shade:tsp>>>6&3,depth:isp>>>29&7,zWrite:!(isp&0x4000000),list:pcw>>>24&7,gouraud:!!(pcw&2),filter:tsp>>>13&3,strided:!paletted&&!!(tcw&0x2000000),supported:format<3||paletted};
}
function mortonIndex(x,y,w,h){const side=Math.min(w,h);return (Math.floor(y/side)*Math.floor(w/side)+Math.floor(x/side))*side*side+twiddle(x%side,y%side);}
export function decodeTexture(memory,info){
 const bytes=memory instanceof Uint8Array?memory:new Uint8Array(memory),s=textureState(info),palette=info.palette||[];
 const byte=a=>a>=0&&a<bytes.length?bytes[a]:null;
 const word=a=>a>=0&&a+1<bytes.length?bytes[a]|bytes[a+1]<<8:null;
 function paletteSample(index){const value=info.registersKnown&&index<palette.length?palette[index]:null;return {index,paletteAddress:0x5f9000+(s.paletteBase+index)*4,value,rgba:value===null?[128,128,128,255]:color(value,info.paletteFormat)};}
 function dictionarySample(x,y){const entry=Math.floor(y/2)*16+Math.floor(x/2),address=s.base+entry*8+((x&1)*2+(y&1))*2,value=word(address);return {entry,address,value,rgba:value===null?[0,0,0,0]:s.supported?color(value,s.format):[255,0,255,255]};}
 function sample(x,y,storage=false){
  x=Math.max(0,Math.min(s.width-1,Math.trunc(x)));y=Math.max(0,Math.min(s.height-1,Math.trunc(y)));
  let address,index=null,indexAddress=null,entry=null,paletteAddress=null,value=null,nibble=null;
  if(s.vq){
   const bx=x>>1,by=y>>1,bw=s.width/2,bh=s.height/2;
   indexAddress=s.base+2048+s.indexOffset+(storage?by*bw+bx:mortonIndex(bx,by,bw,bh));entry=byte(indexAddress);
   address=entry===null?null:s.base+entry*8+((x&1)*2+(y&1))*2;value=address===null?null:word(address);
  }else{
   const at=storage||!s.twiddled?y*s.width+x:mortonIndex(x,y,s.width,s.height);
   address=s.address+(s.format===5?Math.floor(at/2):s.format===6?at:at*2);
   if(s.paletted){const v=byte(address);nibble=s.format===5?at&1:null;index=v===null?null:s.format===5?(v>>((at&1)*4)&15):v;
    if(index!==null){const p=paletteSample(index);value=p.value;paletteAddress=p.paletteAddress;}
   }else value=word(address);
  }
  const valid=value!==null&&s.supported;
  return {x,y,address,index,indexAddress,entry,paletteAddress,nibble,value,valid,rgba:value===null?[0,0,0,0]:s.supported?color(value,s.paletted?info.paletteFormat:s.format):[255,0,255,255]};
 }
 function image(width,height,fn){const rgba=new Uint8ClampedArray(width*height*4);for(let y=0;y<height;y++)for(let x=0;x<width;x++)rgba.set(fn(x,y).rgba,(y*width+x)*4);return {width,height,rgba};}
 // Only construct the requested view; no decoding runs during normal Play.
 const cache={};
 return {state:s,sample,paletteSample,dictionarySample,
  texture:()=>cache.texture??=image(s.width,s.height,sample),
  storage:()=>cache.storage??=(s.vq?image(s.width/2,s.height/2,(x,y)=>{const v=byte(s.base+2048+s.indexOffset+y*(s.width/2)+x);return {rgba:v===null?[0,0,0,0]:[v,v,v,255]};}):image(s.width,s.height,(x,y)=>sample(x,y,true))),
  auxiliary:()=>cache.aux??=(s.paletted?image(16,s.format===5?1:16,(x,y)=>paletteSample(y*16+x)):s.vq?image(32,32,dictionarySample):null),
  indexSample:(x,y)=>{const address=s.base+2048+s.indexOffset+y*(s.width/2)+x;return {address,entry:byte(address)};}
 };
}
