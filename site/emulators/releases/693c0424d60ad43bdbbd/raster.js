import {localElement} from './panel-dom.js';
import {createTileset} from './tileset.js';
const hex=(n,digits=4)=>'0x'+(Number(n)>>>0).toString(16).padStart(digits,'0');
const el=(tag,text,cls)=>{const n=document.createElement(tag);if(text!==undefined)n.textContent=text;if(cls)n.className=cls;return n;};
export function createRaster({platform,send,ui}) {
  const c64=platform==='c64',gg=platform==='gg',width=c64?392:160,height=c64?272:144;
  const labels=c64?['Graphics / border','Sprites','Screen so far']:[gg?'Background':'Background / window','Sprites','Screen so far'];
  const $=id=>localElement(ui.root,id),timeline=ui.timeline;
  ui.auxiliary.hidden=false;
  const buffers=[
    ui.buffer(ui.auxiliary,{id:'raster-background',title:labels[0],description:c64?'Current character or bitmap memory':'Current tilemap and scrolling'}),
    ui.buffer(ui.auxiliary,{id:'raster-sprites',title:'Sprites',description:'Before background priority'}),
    ui.output
  ];
  ui.output.labels('Screen so far','Completed lines keep their original settings');
  ui.sidebar.innerHTML=`<section><h3>Video settings</h3><dl id="raster-registers"></dl><p id="raster-window-line"></p></section>
    <section><h3>Changes before this line</h3><div id="raster-changes"></div></section>
    <section><h3>Pixel and source</h3><div class="raster-pixel-controls"><label>Layer<select id="raster-panel"><option value="2">Screen</option><option value="0">${labels[0]}</option><option value="1">Sprites</option></select></label><label>X<input id="raster-x" type="number" min="0" max="${width-1}" value="${width/2}"></label><label>Y<input id="raster-y" type="number" min="0" max="${height-1}" value="${height/2}"></label><button id="raster-inspect">Inspect</button></div><div id="raster-pixel" aria-live="polite"></div><div id="raster-history" aria-live="polite"></div></section>`;
  const tileset=createTileset({platform,ui,layerPanels:buffers.slice(0,2).map(b=>b.figure)});
  const canvases=buffers.map(b=>b.canvas),slider=timeline.slider,position=timeline.label,note=timeline.note,detail=$('raster-pixel'),history=$('raster-history');
  const guide=timeline.toggle('raster-guide','Show scanline',{checked:true,onChange:()=>paint()}).input;
  let capture=null,latest=0,query=0,line=0,valid=[],layers=null,pending=false;
  const paint=()=>{
    if(!layers)return;
    for(let i=0;i<3;i++){
      buffers[i].draw(layers[i],width,height);const ctx=canvases[i].getContext('2d');
      if(guide.checked){ctx.fillStyle='rgba(23,111,158,.85)';ctx.fillRect(0,line,width,1);}
    }
  };
  function message(parent,text){parent.replaceChildren(el('p',text));}
  function reset(){tileset.reset();capture=null;latest=query=0;valid=[];layers=null;pending=false;detail.replaceChildren();history.replaceChildren();}
  function seek(target){
    if(!capture||!valid.length)return;
    target=valid.reduce((best,n)=>Math.abs(n-target)<Math.abs(best-target)?n:best,valid[0]);
    pending=true;query=0;slider.value=target;position.textContent='Line '+target;
    note.textContent='Reconstructing state at line '+target+'…';
    latest=send('raster-seek',{capture:capture.id,line:target});
  }
  function setCapture(c){
    tileset.reset();layers=null;pending=false;latest=query=0;detail.replaceChildren();history.replaceChildren();
    capture=c;valid=(c.raster?.lines||[]).map(n=>n.line);
    timeline.range({max:height-1,positions:valid});
    timeline.setMarkers((c.raster?.lines||[]).filter(row=>row.changes).map(row=>({value:row.line,label:`Line ${row.line}: ${row.changes} video changes`})));
    buffers.forEach(buffer=>buffer.clear());
    if(valid.length)seek(valid[0]);
    else {position.textContent='No recorded lines';ui.position.textContent='No visible scanlines';$('raster-registers').replaceChildren();$('raster-changes').replaceChildren();$('raster-window-line').textContent='';note.textContent='No visible scanlines were captured. The LCD may be disabled. Return to Play and advance the game.';message(detail,'Capture another frame to inspect its layers.');}
  }
  function memoryLabel(a){return gg?(a>=0x14040?'VDP register '+(a-0x14040):a>=0x14000?'CRAM '+hex(a-0x14000,2):a>=0x10000?'VRAM '+hex(a-0x10000):hex(a)):hex(a);}
  function sourceButton(parent,label,address,size,before,expected){
    const b=el('button',label);b.type='button';b.onclick=()=>{if(pending||!capture)return;query=send('source',{capture:capture.id,address,size,before,expected});message(history,'Reading historical writes…');};parent.append(b);
  }
  function field(parent,label,value){const p=el('p');p.append(el('strong',label+': '),document.createTextNode(value));parent.append(p);}
  function inspect(panel,x,y){
    if(!capture||pending||!layers)return;
    query=send('raster-pixel',{capture:capture.id,panel,x,y});message(detail,'Inspecting pixel…');history.replaceChildren();
  }
  buffers.forEach((buffer,panel)=>buffer.setInspect((x,y)=>inspect(panel,x,y)));
  $('raster-inspect').onclick=()=>inspect(Number($('raster-panel').value),Number($('raster-x').value),Number($('raster-y').value));
  timeline.onSeek(seek);
  function showState(info){
    line=info.line;timeline.value(line,'Line '+line+' / '+(height-1));ui.position.textContent='State at line '+line+(c64?' (VIC raster '+info.raster+')':gg?' (VDP line '+info.raster+')':'');
    note.textContent=`Screen contains captured lines through ${line}. Source panels hold ${c64?'memory and registers from the start of this line':'this line’s settings'} fixed.`;
    if(!info.complete||valid.length!==height)note.textContent+=' Some capture evidence is incomplete.';
    const regs=$('raster-registers');regs.replaceChildren();
    for(const r of info.registers){const entry=el('div');entry.append(el('dt',r.name),el('dd',`${r.value} (${hex(r.value,2)})`));regs.append(entry);}
    $('raster-window-line').textContent=c64?`VIC bank ${hex(info.vicBank)} · screen ${hex(info.screenBase)} · ${(info.mode&2)?'bitmap '+hex(info.bitmapBase):'characters '+hex(info.charsetBase)} · ${['Text','Multicolor text','Bitmap','Multicolor bitmap','Extended-background text'][info.mode]||'Invalid mode '+info.mode}`:gg?`Effective scroll X / Y: ${info.scrollX} / ${info.scrollY}. Line counter: ${info.lineCounter}. Pending interrupts: ${[info.linePending?'line':'',info.framePending?'frame':''].filter(Boolean).join(', ')||'none'}. Tilemap ${hex(info.tilemapBase)}, sprites ${hex(info.spriteBase)} in VRAM.`:'Window row counter: '+info.windowLine;
    const changes=$('raster-changes');changes.replaceChildren();
    let more=null;
    for(const [index,c]of info.changes.entries()){
      const p=el(c64?'button':'p');if(c64){p.type='button';p.className='raster-write';p.onclick=()=>showC64Source({role:'Video write',address:c.address,space:c.space,value:c.after,writer:c.writer});}
      p.append(el('strong',c.name+' '),document.createTextNode(`${c.before} → ${c.after}`),el('span','PC '+hex(c.pc),'raster-writer'));
      if(c64&&index>=4){if(!more){more=el('details',undefined,'raster-more');more.append(el('summary',`${info.changes.length-4} more writes`));changes.append(more);}more.append(p);}else changes.append(p);
    }
    const counts=[];if(info.tileWrites)counts.push(info.tileWrites+(c64?' graphics-byte writes':gg?' VRAM writes':' tile-data writes'));if(info.mapWrites)counts.push(info.mapWrites+(c64?' screen-memory writes':' tilemap writes'));if(info.objectWrites)counts.push(info.objectWrites+(c64?' sprite-data/pointer writes':' sprite-attribute writes'));
    if(info.colorWrites)counts.push(info.colorWrites+(gg?' palette color writes':' color RAM writes'));
    if(counts.length)changes.append(el('p',counts.join(', ')));
    if(!info.changes.length&&!counts.length)changes.append(el('p','No video writes since the previous rendered line.'));
    if(info.changeCount>info.changes.length)changes.append(el('p','Change list truncated.'));
    message(detail,c64?'Select a pixel to follow character/bitmap bytes, screen memory, sprites and CPU writers.':'Select a pixel in any panel to follow its tile, sprite attributes or palette.');history.replaceChildren();
    ui.sidebar.scrollTop=0;
  }
  function showPixel(p){
    detail.replaceChildren();history.replaceChildren();if(p.error){message(detail,p.error);return;}
    $('raster-panel').value=p.panel;$('raster-x').value=p.x;$('raster-y').value=p.y;
    if(c64){showC64Pixel(p);return;}
    const title=labels[p.panel];
    field(detail,title,`(${p.x}, ${p.y}), state at line ${p.stateLine}`);
    if(p.panel<2)detail.append(el('p','Layer preview with settings frozen at this line. Other rows are illustrative.','raster-context'));
    else field(detail,'Reconstruction',p.complete?'Layer decisions match the captured pixel.':'Incomplete or mismatched evidence.');
    for(const c of p.candidates){
      const group=el('div',undefined,'raster-candidate');const verdict=c.flags&4?'Transparent color':c.flags&8?'Hidden by another sprite':c.flags&2?'Behind background':c.kind==='Sprite'?'Composited sprite':'Base layer';
      group.append(el('strong',c.kind+(c.object>=0?' '+c.object:'')),el('p',verdict));
      if(gg){for(const source of c.sources||[])sourceButton(group,source.role+' '+memoryLabel(source.address),source.address,source.size,p.sourceBefore,source.value);detail.append(group);continue;}
      if(c.sourceAddress)sourceButton(group,'Tile row '+hex(c.sourceAddress),c.sourceAddress,2,p.sourceBefore,c.sourceValue);
      if(c.mapAddress)sourceButton(group,'Tilemap entry '+hex(c.mapAddress),c.mapAddress,1,p.sourceBefore,c.mapValue);
      if(c.descriptor)sourceButton(group,'Sprite attributes '+hex(c.descriptor),c.descriptor,4,p.sourceBefore,c.descriptorValue);
      sourceButton(group,'Palette '+hex(c.paletteAddress),c.paletteAddress,1,p.sourceBefore,c.paletteValue);detail.append(group);
    }
    if(!p.candidates.length)detail.append(el('p','No selected sprite covers this position.'));
  }
  function showC64Source(c){
    history.replaceChildren();field(history,c.role,`${hex(c.address)} = ${hex(c.value,2)} (${['RAM','color RAM','I/O','ROM'][c.space]})`);
    if(c.fetchCycle)field(history,'Actual fetch cycle',c.fetchCycle);
    if(c.writer){const w=c.writer;field(history,'CPU writer',`PC ${hex(w.pc)}, cycle ${w.cycle}`);field(history,'Byte change',`${hex(w.oldValue,2)} → ${hex(w.value,2)}`);field(history,'Instruction bytes',w.code.map(n=>hex(n,2)).join(' '));if(w.interrupt)field(history,'Context','Interrupt handler');}
    else field(history,'Writer',c.space===3?'Immutable character ROM':'No earlier writer is recorded.');
    history.scrollIntoView({block:'nearest'});
  }
  function showC64Pixel(p){
    $('raster-panel').value=p.panel;$('raster-x').value=p.x;$('raster-y').value=p.y;
    field(detail,labels[p.panel],`(${p.x}, ${p.y}), state at line ${p.stateLine}`);
    detail.append(el('p',p.preview?'Preview using frozen memory and registers. Actual fetching can differ when the game changes settings during the frame.':`Recorded VIC output at raster ${p.raster}, cycle ${p.cycle}.`,'raster-context'));
    if(!p.complete)detail.append(el('p','Some evidence is incomplete.'));
    if(!p.preview)field(detail,'VIC decision',p.border?'Border/background':p.spriteMask?'Graphics and sprite candidates '+hex(p.spriteMask,2):'Graphics');
    let controls=null;
    for(const c of p.contributors||[]){
      const b=el('button',c.role+(c.missing?' — unavailable':' '+hex(c.address)));b.type='button';b.disabled=!!c.missing;b.onclick=()=>showC64Source(c);
      if(c.space===2){if(!controls){controls=el('details',undefined,'raster-more');controls.append(el('summary','VIC registers and bank controls'));}controls.append(b);}else detail.append(b);
    }
    if(controls)detail.append(controls);
    if(p.panel===1&&!(p.contributors||[]).some(c=>/sprite.*pixel byte/.test(c.role)))detail.append(el('p','No sprite covers this position in the frozen preview.'));
    const graphics=(p.contributors||[]).find(c=>!c.missing&&/graphics byte|bitmap byte|sprite.*pixel byte/.test(c.role));
    if(graphics)showC64Source(graphics);
    else detail.scrollIntoView({block:'nearest'});
  }
  function result(m){
    if(!capture||m.capture!==capture.id)return;
    if(m.type==='raster-seek'&&m.request===latest){
      pending=false;if(m.info.error){message(detail,m.info.error);note.textContent=m.info.error;return;}
      layers=m.layers;showState(m.info);tileset.update(m.tileset);paint();return;
    }
    if(m.request!==query||pending)return;
    if(m.type==='raster-pixel')showPixel(m.evidence);
    else if(m.type==='source'){
      const p=m.evidence;history.replaceChildren();if(p.error){message(history,p.error);return;}
      field(history,'Memory',memoryLabel(p.address));field(history,'At this scanline',p.complete?'Historical bytes match the sampled source.':'History is incomplete.');
      for(const w of p.contributors||[])field(history,'CPU write',`${hex(w.before,w.size*2)} → ${hex(w.after,w.size*2)}, PC ${hex(w.pc)}`);
      if(!p.contributors?.length)history.append(el('p','These contents were already present when capture began.'));
      if(p.truncated||p.overflow)history.append(el('p','The displayed writer history is truncated.'));
    }
  }
  reset();return {reset,setCapture,result,seek,isInspecting:()=>!!capture};
}
