import {decodeTexture,textureFormats} from './dc-texture-decode.js';
const make=(tag,text)=>{const e=document.createElement(tag);if(text!==undefined)e.textContent=text;return e;},hex=(v,n=6)=>v===null?'unavailable':'$'+(v>>>0).toString(16).padStart(n,'0');
export function createDCTexture({ui,seek}){
 ui.auxiliary.hidden=ui.toolbar.hidden=false;
 const controls=make('div');controls.className='tileset-controls';ui.toolbar.append(controls);
 const label=make('label','Texture view'),mode=make('select');mode.id='dc-texture-mode';
 for(const [value,text]of [['decoded','Decoded texture'],['storage','Memory order']]){const o=make('option',text);o.value=value;mode.append(o);}label.append(mode);controls.append(label);
 const overlayLabel=make('label'),overlay=make('input');overlay.type='checkbox';overlay.checked=true;overlay.id='dc-texture-outline';overlayLabel.append(overlay,make('span','Polygon coordinates'));controls.append(overlayLabel);
 const auxLabel=make('label'),showAux=make('input');showAux.type='checkbox';showAux.id='dc-texture-lookup';showAux.checked=true;auxLabel.append(showAux,make('span','Palette / dictionary'));controls.append(auxLabel);
 const previous=make('button','Previous texture'),next=make('button','Next texture');previous.id='dc-texture-prev';next.id='dc-texture-next';previous.onclick=()=>seek(snapshot.info.previousTextured);next.onclick=()=>seek(snapshot.info.nextTextured);controls.append(previous,next);
 const board=make('div');board.id='dc-texture-board';ui.auxiliary.append(board);
 const texture=ui.buffer(board,{id:'dc-texture',title:'Current texture',width:8,height:8,aspect:'1/1',fit:'contain'});
 const aux=ui.buffer(board,{id:'dc-texture-aux',title:'Palette / block dictionary',width:32,height:32,aspect:'1/1',fit:'contain'});
 const empty=make('p','Capture a frame to inspect its textures.');empty.className='dc-texture-empty';board.append(empty);
 const section=make('section');section.id='dc-gpu-state';section.append(make('h3','GPU state at this step'));
 const context=make('p'),fields=make('dl'),limits=make('details');fields.className='gpu-state-fields';
 limits.append(make('summary','How this preview works'),make('p','The preview follows the current software sampler: nearest texel, wrapping UVs and the largest mip level. Requested filtering, clamp/flip and stride controls are shown below but are not implemented by the core. YUV and bump textures are unsupported. Rendering follows submitted polygons; hardware tile processing and translucent sorting are not reproduced.'),make('p','Texture memory is shown after the selected step. If rendering overwrites its own texture, this preview can differ from earlier reads within that step. Pixel history currently records output writes, not the original texture samples.'));
 section.append(context,fields,limits);ui.sidebar.prepend(section);
 const source=make('section');source.id='dc-texture-source';source.append(make('h3','Texture sample'));const sampleText=make('p');source.append(sampleText);ui.sidebar.append(source);
 let snapshot=null,decoded=null,point=null;
 function field(name,value){const row=make('div');row.append(make('dt',name),make('dd',value));fields.append(row);}
 function display(view,img){view.picture.style.aspectRatio=img.width+'/'+img.height;view.draw(img.rgba,img.width,img.height);}
 function inspect(kind,x,y){if(!decoded)return;point={kind,x,y};draw();source.scrollIntoView({block:'nearest'});}
 texture.setInspect((x,y)=>inspect('texture',x,y));aux.setInspect((x,y)=>inspect('aux',x,y));
 function sample(){
  if(!point||!decoded)return;const s=decoded.state;let p;
  if(point.kind==='aux'){
   if(s.paletted){p=decoded.paletteSample(point.y*16+point.x);sampleText.textContent=`Palette index ${p.index} · register ${hex(p.paletteAddress)} = ${hex(p.value,8)}.`;}
   else{p=decoded.dictionarySample(point.x,point.y);sampleText.textContent=`Dictionary entry ${p.entry} · texel (${point.x&1}, ${point.y&1}) · VRAM ${hex(p.address)} = ${hex(p.value,4)}.`;}
  }else if(mode.value==='storage'&&s.vq){p=decoded.indexSample(point.x,point.y);sampleText.textContent=`Index byte at VRAM ${hex(p.address)} selects dictionary entry ${p.entry??'outside VRAM'}. Each entry supplies a 2 × 2 block.`;
  }else{
   p=decoded.sample(point.x,point.y,mode.value==='storage');
   sampleText.textContent=`${mode.value==='storage'?'Storage position':'Texel'} (${point.x}, ${point.y}) · VRAM ${hex(p.address)}${p.nibble===null?'':` · ${p.nibble?'high':'low'} nibble`}${p.index===null?'':` · index ${p.index} → palette register ${hex(p.paletteAddress)}`}${p.entry===null?'':` · index byte ${hex(p.indexAddress)} → dictionary entry ${p.entry}`} · color ${hex(p.value,s.paletted?8:4)}${p.valid?` · RGBA ${p.rgba.join(', ')}`:' · unsupported or unavailable'}.`;
  }
  if(!aux.figure.hidden){const c=aux.canvas.getContext('2d');c.strokeStyle='#da9825';c.lineWidth=.25;
   if(s.paletted&&p.index!==null&&p.index!==undefined)c.strokeRect(p.index%16,Math.floor(p.index/16),1,1);
   if(s.vq&&p.entry!==null&&p.entry!==undefined)c.strokeRect((p.entry%16)*2,Math.floor(p.entry/16)*2,2,2);
  }
 }
 function draw(){
  if(!snapshot)return;const i=snapshot.info;decoded??=decodeTexture(snapshot.memory,i);const s=decoded.state,active=s.known&&s.textured;
  previous.disabled=!(i.previousTextured>=0);next.disabled=!(i.nextTextured>=0);fields.replaceChildren();texture.figure.hidden=!active;aux.figure.hidden=true;empty.hidden=active;
  context.textContent=`Step ${i.cursor.toLocaleString()} · ${i.type===4?'Polygon header':i.type===5?'Sprite header':i.type===7?(i.draw?'Textured primitive':'Strip vertex'):(s.known?'No texture sampled · retained binding':'No active texture')}.${i.complete?'':' Capture is incomplete.'}`;
  if(!active){empty.textContent='No texture is bound at this step. Use Previous texture or Next texture to find a draw.';sampleText.textContent='Select a textured header or primitive with the render scrubber.';return;}
  field('Texture',`${s.width} × ${s.height} · ${textureFormats[s.format]}`);field('VRAM address',hex(s.base));
  field('Storage',s.vq?'VQ · 2 × 2 block dictionary':s.twiddled?'Twiddled · interleaved coordinates':'Row order');
  field('Mipmaps',s.mipmap?'Present · largest level sampled':'Off');if(s.mipmap)field('Level address',hex(s.vq?s.base+2048+s.indexOffset:s.address));
  if(s.paletted)field('Palette',`${s.paletteBase}–${s.paletteBase+(s.format===5?15:255)} · ${['ARGB1555','RGB565','ARGB4444','ARGB8888'][i.paletteFormat]}${i.registersKnown?'':' · not captured'}`);
  field('Shading',['Decal','Modulate','Decal alpha','Modulate alpha'][s.shade]+(s.gouraud?' · Gouraud':' · flat'));
  field('List',['Opaque','Opaque modifier','Translucent','Translucent modifier','Punch-through'][s.list]||s.list);
  field('Depth test',['Never','Less','Equal','Less or equal','Greater','Not equal','Greater or equal','Always'][s.depth]);field('Depth writes',s.zWrite?'On':'Off');
  field('Framebuffer address',hex(i.target)+' · CPU 32-bit layout');
  field('Filter requested',['Nearest','Bilinear','Trilinear pass 1','Trilinear pass 2'][s.filter]);
  field('UV requested',`Clamp ${i.tsp&0x10000?'U':''}${i.tsp&0x8000?'V':''}${!(i.tsp&0x18000)?'off':''} · flip ${i.tsp&0x40000?'U':''}${i.tsp&0x20000?'V':''}${!(i.tsp&0x60000)?'off':''}`);
  if(s.strided)field('Stride requested',`${i.stride} pixels · core uses ${s.width}`);
  if(!s.supported)field('Format support','Unsupported by the core · magenta placeholder');
  const img=mode.value==='storage'?decoded.storage():decoded.texture();display(texture,img);
  texture.labels(mode.value==='storage'?(s.vq?'VQ indices · memory order':'Texels · memory order'):'Current texture',mode.value==='storage'?'Consecutive addresses, laid out in rows':`${s.width} × ${s.height} · ${textureFormats[s.format]}`);
  const a=showAux.checked?decoded.auxiliary():null;if(a){aux.figure.hidden=false;display(aux,a);aux.labels(s.paletted?'Active palette':'VQ block dictionary',s.paletted?'Click a color to inspect its register':'256 entries · each contains 2 × 2 texels');}
  board.dataset.aux=a?'yes':'no';board.style.setProperty('--texture-ratio',img.width/img.height);board.style.setProperty('--aux-ratio',a?a.width/a.height:1);
  if(mode.value==='decoded'&&overlay.checked&&i.uv.length){
   const c=texture.canvas.getContext('2d');c.strokeStyle='#eabf32';c.lineWidth=Math.max(s.width,s.height)/180;
   c.beginPath();let count=0;for(const uv of i.uv){if(uv.some(v=>v===null||!Number.isFinite(v)))continue;const [u,v]=uv;if(Math.abs(u)>1e6||Math.abs(v)>1e6)continue;c[count++?'lineTo':'moveTo'](u*s.width,v*s.height);}if(count>2)c.closePath();c.stroke();
  }
  sample();
 }
 mode.onchange=()=>{point=null;sampleText.textContent='Click a texel to inspect its memory lookup.';draw();};overlay.onchange=showAux.onchange=draw;
 function reset(){snapshot=decoded=point=null;previous.disabled=next.disabled=true;texture.clear();aux.clear();texture.figure.hidden=aux.figure.hidden=true;empty.hidden=false;empty.textContent='Capture a frame to inspect its textures.';fields.replaceChildren();context.textContent='Texture settings follow the render position.';sampleText.textContent='Click a texel to inspect its memory lookup.';}
 function update(data){if(!data||data.info?.error){reset();return;}snapshot=data;decoded=null;point=null;sampleText.textContent='Click a texel to inspect its memory lookup.';draw();}
 reset();return {reset,update};
}
