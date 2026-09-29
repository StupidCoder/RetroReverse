import {decodeVRAM} from './ps1-vram-decode.js';
const make=(tag,text)=>{const e=document.createElement(tag);if(text!==undefined)e.textContent=text;return e;},hex=(v,n=4)=>'$'+(v>>>0).toString(16).padStart(n,'0');
export function createPS1VRAM({ui}) {
 ui.auxiliary.hidden=false;ui.toolbar.hidden=false;
 const controls=make('div');controls.className='tileset-controls';ui.toolbar.append(controls);
 const label=make('label','VRAM colors'),mode=make('select');mode.id='vram-mode';
 for(const [value,text]of [['auto','Follow GPU'],['0','4-bit indexed'],['1','8-bit indexed'],['2','16-bit direct']]){const o=make('option',text);o.value=value;mode.append(o);}label.append(mode);controls.append(label);
 const viewLabel=make('label','Auxiliary view'),view=make('select');view.id='vram-view';
 for(const [value,text]of [['all','VRAM + texture'],['vram','Full VRAM'],['texture','Texture page']]){const o=make('option',text);o.value=value;view.append(o);}viewLabel.append(view);controls.append(viewLabel);
 const toggle=make('label'),window=make('input');window.type='checkbox';window.id='vram-window';window.checked=true;toggle.append(window,make('span','Apply texture window'));controls.append(toggle);
 const board=make('div');board.id='ps1-vram-board';board.dataset.view='all';ui.auxiliary.append(board);
 const overview=ui.buffer(board,{id:'ps1-vram',title:'VRAM',width:1024,height:512,aspect:'2/1'});
 const texture=ui.buffer(board,{id:'ps1-texture',title:'Texture page',width:256,height:256,aspect:'1/1'});
 view.onchange=()=>{board.dataset.view=view.value;overview.figure.hidden=view.value==='texture';texture.figure.hidden=view.value==='vram';};
 const section=make('section');section.id='ps1-gpu-state';section.append(make('h3','GPU state at this step'));
 const context=make('p'),fields=make('dl'),limits=make('details'),limitText=make('p');fields.className='gpu-state-fields';limits.append(make('summary','Emulation limits'),limitText);section.append(context,fields,make('p','Blend, dither and mask effects are not emulated.'),limits);ui.sidebar.prepend(section);
 const source=make('section');source.id='ps1-vram-source';source.append(make('h3','Texture sample'));const sampleText=make('p');source.append(sampleText);ui.sidebar.append(source);
 let snapshot=null,decoded=null,point=null;
 const formats=['4-bit indexed','8-bit indexed','16-bit direct'];
 function field(name,value){const row=make('div');row.append(make('dt',name),make('dd',value));fields.append(row);}
 function inspect(kind,x,y){if(!decoded)return;point={kind,x,y};paintSelection();source.scrollIntoView({block:'nearest'});}
 overview.setInspect((x,y)=>inspect('vram',x,y));texture.setInspect((x,y)=>inspect('texture',x,y));
 function paintSelection(){
  if(!point||!decoded)return;
  const s=point.kind==='texture'?decoded.pageSample(point.x,point.y):decoded.overviewSample(Math.min(decoded.width-1,point.x),point.y);
  sampleText.textContent=`${point.kind==='texture'?`UV (${point.x}, ${point.y}) → sampled (${s.u}, ${s.v}) · `:''}VRAM word (${s.x}, ${s.y}), byte ${hex(s.address,6)} = ${hex(s.word)}${s.index===null?'':` · index ${s.index}${s.paletteAddress===null?' · CLUT unknown':` → palette byte ${hex(s.paletteAddress,6)}`}`}${s.color===null?'':` · color ${hex(s.color)}${s.transparent?' (transparent)':s.color&0x8000?' (bit 15 set)':''}`}. Values are from the completed rendering step.`;
 }
 function rect(ctx,x,y,w,h,color){ctx.strokeStyle=color;ctx.lineWidth=3*decoded.factor;ctx.strokeRect(x,y,w,h);}
 function draw(){
  if(!snapshot)return;const s=snapshot.info;decoded=decodeVRAM(snapshot.memory,s,{mode:mode.value,window:window.checked});
  overview.labels('VRAM · '+formats[decoded.depth]);overview.draw(decoded.overview,decoded.width,decoded.height);
  texture.labels('Texture page'+(s.textured?' · current command':' · current settings'));texture.draw(decoded.page,256,256);
  const ctx=overview.canvas.getContext('2d'),f=decoded.factor;
  // Keep physical word coordinates in the overview. Split wrapped pages at VRAM edges.
  for(const dx of [0,-decoded.width])for(const dy of [0,-512])rect(ctx,s.pageX*f+dx,s.pageY+dy,256,256,'#eabf32');
  if(decoded.depth<2&&s.clut>=0){const begin=((s.clut>>6)&511)*1024+(s.clut&63)*16,count=decoded.depth===0?16:256;for(let i=0;i<count;){const at=(begin+i)&524287,n=Math.min(count-i,1024-(at&1023));rect(ctx,(at&1023)*f,at>>10,n*f,1,'#25b8e8');i+=n;}}
  const t=texture.canvas.getContext('2d');if(s.uv.length){t.strokeStyle='#eabf32';t.lineWidth=1.5;const order=s.uv.length===4&&s.opcode<0x40?[0,1,3,2]:s.uv.map((_,i)=>i);t.beginPath();order.forEach((i,k)=>t[k?'lineTo':'moveTo'](...s.uv[i]));t.closePath();t.stroke();}
  context.textContent=`Step ${s.cursor.toLocaleString()} · ${s.cursor?'GP0 '+hex(s.opcode,2):'Initial state'}. ${s.textured?'Yellow outlines show this primitive’s texture coordinates.':'No texture sampled by this command; showing retained sampler settings.'} Yellow = page, blue = palette. Overview keeps VRAM’s word layout; the page shows square texels. Indexed colors use one CLUT across the overview.${decoded.depth<2&&!decoded.hasPalette?' CLUT unknown until a textured primitive is captured; showing grayscale indices.':''}${!s.complete?' Capture is incomplete.':''}`;
  fields.replaceChildren();field('Texture format',formats[Math.min(s.depth,2)]+(mode.value==='auto'?'':' (view override)'));field('Page origin',`${s.pageX}, ${s.pageY} words`);field('CLUT',s.clut<0?'Not recorded yet':`${(s.clut&63)*16}, ${(s.clut>>6)&511} words${s.textured?'':' · last textured primitive'}`);
  field('Drawing area',`${s.area[0]}, ${s.area[1]} – ${s.area[2]-1}, ${s.area[3]-1}`);field('Drawing offset',s.offset.join(', '));field('Texture window',`Mask ${s.window[0]}, ${s.window[1]} · offset ${s.window[2]}, ${s.window[3]} (×8)`);
  if((s.opcode>=0x20&&s.opcode<0x40)||(s.opcode>=0x60&&s.opcode<0x80)){field('Shading',s.gouraud?'Gouraud':'Flat');if(s.textured)field('Texture color',s.raw?'Raw requested':'Modulated');field('Semi-transparency',s.semi?'Requested':'Off');}
  const requested=v=>v<0?'Not observed in capture':v?'On (requested)':'Off';field('Blend formula',s.blend<0?'Not observed in capture':['B/2 + F/2','B + F','B − F','B + F/4'][s.blend]+' (requested)');field('Dithering',requested(s.dither));field('Set / test mask',requested(s.forceMask)+' / '+requested(s.checkMask));
  limitText.textContent='The core currently ignores semi-transparency blending, dithering and mask tests. Raw-texture polygons still use modulation; rectangles honor raw texture. Values above describe recorded requests. UV outlines are guides, not exact coverage. VRAM is shown after the step: self-modifying texture reads can differ from this preview; pixel provenance retains the original sampled values.';
  paintSelection();
 }
 mode.onchange=()=>{point=null;draw();};window.onchange=draw;
 function reset(){snapshot=decoded=point=null;overview.clear();texture.clear();fields.replaceChildren();context.textContent='Capture a display to inspect VRAM and GPU state.';sampleText.textContent='Select a texel or a VRAM word to see its address and palette lookup.';}
 function update(data){if(!data||data.info?.error){reset();return;}snapshot=data;draw();}
 reset();return {reset,update};
}
