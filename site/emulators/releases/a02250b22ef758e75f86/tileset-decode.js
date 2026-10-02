// Captured-memory layouts are exported by rr_tileset_data in each core.
// Pure decoders: changing a palette here cannot modify the paused machine.
export const tilesetBytes={c64:4240,gb:8320,gg:16464,gba:106496};
const hex=(v,n=4)=>'$'+v.toString(16).padStart(n,'0');
const word=(b,i)=>b[i]|b[i+1]<<8;
const expand5=v=>(v<<3)|(v>>2);
const rgb15=c=>[expand5(c&31),expand5(c>>5&31),expand5(c>>10&31),255];
export function decodeTileset(platform,data,options={}) {
 const b=new Uint8Array(data);if(b.length!==tilesetBytes[platform])throw Error('Tileset snapshot is unavailable.');
 let count,columns=16,colors,decode,describe,title,description;
 const palette=Math.max(0,Math.min(15,Number(options.palette)||0));
 if(platform==='c64'){
  const mode=((b[0x11]>>4)&6)|((b[0x16]>>4)&1),ecm=mode===4,multi=mode===1,base=word(b,64),screen=word(b,66);
  const cell=Math.max(0,Math.min(999,Number(options.cell)||0)),auto=options.context!=='cell',cells=new Int16Array(256);cells.fill(-1);
  for(let i=0;i<1000;i++)if(cells[b[1168+i]]<0)cells[b[1168+i]]=i;
  const context=code=>auto&&cells[code]>=0?cells[code]:cell;
  colors=Array.from({length:16},(_,i)=>Array.from(b.slice(80+i*4,84+i*4)));
  count=256;title='Character set';description=`${hex(base)} · ${['Single-color text','Multicolor text','Bitmap mode: raw character memory','Multicolor bitmap: raw character memory','Extended-background text'][mode]||'Unsupported text mode: raw character memory'}. ${auto?'Each code uses its first screen cell; unused codes use cell '+cell+'.':'All codes use color RAM from cell '+cell+'.'}`;
  decode=(code,x,y)=>{const fg=b[144+context(code)]&15,v=b[2192+(ecm?code&63:code)*8+y],mc=multi&&!!(fg&8);return colors[mc?[b[0x21]&15,b[0x22]&15,b[0x23]&15,fg&7][v>>(6-(x&6))&3]:(v>>(7-x)&1)?fg:b[0x21+(ecm?code>>6:0)]&15];};
  describe=code=>{const c=context(code),fg=b[144+c]&15;return `Code ${hex(code,2)} · ${hex(base+(ecm?code&63:code)*8)} · screen cell ${c} (${hex(screen+c)}) · color ${fg}${multi?(fg&8?' · multicolor':' · single-color override'):''}${ecm?' · background '+(code>>6):''}. ${auto&&cells[code]<0?'Code absent from the screen; fallback color.':''}`;};
 }else if(platform==='gb'){
  const p=Math.min(palette,2),pal=b[8192+0x47+p];colors=Array.from({length:4},(_,i)=>{const c=255-85*(pal>>(i*2)&3);return [c,c,c,p&&i===0?0:255];});
  count=384;title='Tile memory';description=`$8000–$97ff · BGP / OBJ palettes from this line. BG uses ${b[8192+0x40]&16?'unsigned IDs at $8000':'signed IDs around $9000'}; sprites use $8000. Object color 0 is transparent.`;
  decode=(tile,x,y)=>{const at=tile*16+y*2;return colors[(b[at]>>(7-x)&1)|((b[at+1]>>(7-x)&1)<<1)];};
  describe=tile=>{const address=0x8000+tile*16,id=b[8192+0x40]&16?(tile<256?tile:null):(tile>=128?(tile-256)&255:null);return `Tile ${tile} · ${hex(address)} · BG code ${id===null?'outside current addressing range':hex(id,2)} · ${['BGP','OBP0','OBP1'][p]} = ${hex(pal,2)}.`;};
 }else if(platform==='gg'){
  const p=Math.min(palette,1);colors=Array.from({length:16},(_,i)=>{const c=word(b,16384+(p*16+i)*2);return [(c&15)*17,(c>>4&15)*17,(c>>8&15)*17,255];});
  count=512;title='Tile memory';description=`512 tiles · palette ${p}. Background cells select either palette; sprites use palette 1. Color 0 is shown as a swatch here; sprite color 0 is transparent.`;
  decode=(tile,x,y)=>{const at=tile*32+y*4;let c=0;for(let p=0;p<4;p++)c|=(b[at+p]>>(7-x)&1)<<p;return colors[c];};
  describe=tile=>`Tile ${tile} (${hex(tile,3)}) · VRAM ${hex(tile*32)} · palette ${p}.`;
 }else if(platform==='gba'){
  const source=Number(options.source)||0,obj=source>=4,bg=source&3,mode=word(b,0)&7,ctl=word(b,8+bg*2),affine=!obj&&((mode===1&&bg===2)||(mode===2&&bg>=2));
  const eight=obj?source===5:affine||!!(ctl&128),base=obj?0x10000:(ctl>>2&3)*0x4000,palBase=4096+(obj?512:0)+(eight?0:palette*32),tileSize=eight?64:32;
  colors=Array.from({length:eight?256:16},(_,i)=>{const c=rgb15(word(b,palBase+i*2));if(!i)c[3]=0;return c;});
  count=(obj?32768:16384)/tileSize;if(obj)columns=32;
  title=obj?'Object tiles':`BG${bg} character block`;
  const active=obj?!!(word(b,0)&0x1000):!!(word(b,0)&(0x100<<bg))&&(mode===0||mode===1&&bg<3||mode===2&&bg>=2);
  description=`${hex(0x6000000+base,8)} · ${obj?32:16} KiB · ${eight?'8-bit, 256 colors':`4-bit, palette bank ${palette}`}. ${active?'':'This layer is disabled or uses bitmap graphics; showing raw tile memory. '}${obj?'Object size and mapping determine tile arrangement.':'One character block; larger tile IDs may continue into the next block.'} Color 0 is transparent.`;
  decode=(tile,x,y)=>{const at=8192+base+tile*tileSize+y*(eight?8:4)+(eight?x:x>>1),v=b[at];return colors[eight?v:v>>((x&1)*4)&15];};
  describe=tile=>`Tile ${tile} · ${hex(0x6000000+base+tile*tileSize,8)} · ${eight?'256-color palette':`palette bank ${palette}`} · ${affine?'affine background':obj?'object data':'text background'}.`;
 }else throw Error('No tileset decoder for this platform.');
 const width=columns*8,height=Math.ceil(count/columns)*8,pixels=new Uint8ClampedArray(width*height*4);
 for(let tile=0;tile<count;tile++)for(let y=0;y<8;y++)for(let x=0;x<8;x++)pixels.set(decode(tile,x,y),(((tile/columns|0)*8+y)*width+(tile%columns)*8+x)*4);
 return {width,height,count,columns,pixels,colors,title,description,describe};
}
