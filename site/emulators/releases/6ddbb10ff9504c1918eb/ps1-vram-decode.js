// Mirrors GPU::texel, using only VRAM at the completed replay cursor.
export function decodeVRAM(data,state,{mode='auto',window=true}={}) {
 const bytes=new Uint8Array(data);if(bytes.length!==1048576)throw Error('No captured VRAM.');
 const words=new Uint16Array(bytes.buffer,bytes.byteOffset,524288),depth=mode==='auto'?Math.min(state.depth,2):Number(mode),factor=depth===0?4:depth===1?2:1,bits=16/factor;
 const hasPalette=state.clut>=0,clut=hasPalette?((state.clut>>6)&511)*1024+(state.clut&63)*16:0;
 const rgba=v=>(0xff000000|((v&31)<<3)|((v>>5&31)<<11)|((v>>10&31)<<19))>>>0;
 const palette=new Uint32Array(256);for(let i=0;i<256;i++){const v=words[(clut+i)&524287],gray=i*(depth===0?17:1);palette[i]=hasPalette?rgba(v):(0xff000000|gray*0x10101)>>>0;}
 const width=1024*factor,height=512,overview=new Uint32Array(width*height),page=new Uint32Array(256*256);
 const indexMask=(1<<bits)-1;
 for(let at=0;at<words.length;at++){const v=words[at];for(let sub=0;sub<factor;sub++)overview[at*factor+sub]=depth===2?rgba(v):palette[(v>>(sub*bits))&indexMask];}
 function sample(at,part){const value=words[at],index=depth===2?null:(value>>(part*bits))&indexMask,paletteAddress=index!==null&&hasPalette?((clut+index)&524287)*2:null,color=index===null?value:hasPalette?words[paletteAddress/2]:null;return {x:at&1023,y:at>>10,address:at*2,word:value,part,index,paletteAddress,color,rgba:color===null?palette[index]:rgba(color),transparent:color===0};}
 function overviewSample(x,y){return sample(y*1024+Math.floor(x/factor),x%factor);}
 function pageSample(u,v){
  u&=255;v&=255;
  if(window){const [mx,my,ox,oy]=state.window;u=(u&~(mx*8))|((ox&mx)*8);v=(v&~(my*8))|((oy&my)*8);}
  const at=((state.pageY+v)&511)*1024+((state.pageX+Math.floor(u/factor))&1023);
  return {...sample(at,u%factor),u,v};
 }
 for(let v=0;v<256;v++)for(let u=0;u<256;u++){const s=pageSample(u,v);page[v*256+u]=s.transparent?0:s.rgba;}
 return {width,height,factor,depth,hasPalette,overview:new Uint8ClampedArray(overview.buffer),page:new Uint8ClampedArray(page.buffer),overviewSample,pageSample};
}
