import {createWorkspaces} from './workspaces.js';
import {pixelCoordinates} from './inspector.js';
const hex=(n,digits=4)=>'0x'+(Number(n)>>>0).toString(16).padStart(digits,'0');
const el=(tag,text,cls)=>{const n=document.createElement(tag);if(text!==undefined)n.textContent=text;if(cls)n.className=cls;return n;};
export function createGameBoyRaster({send,resume}) {
  const $=id=>document.getElementById(id),root=$('render-workspace');
  const views=createWorkspaces({navigation:$('workspace-nav'),onChange:id=>document.body.dataset.workspace=id});
  views.register({id:'play',label:'Play',panel:$('play-workspace')});
  views.register({id:'render',label:'Render',panel:root,enabled:false});
  const canvases=[$('raster-background'),$('raster-sprites'),$('raster-output')];
  const slider=$('raster-position'),position=$('raster-line'),note=$('raster-note'),detail=$('raster-pixel'),history=$('raster-history');
  let capture=null,latest=0,query=0,line=0,valid=[],layers=null,pending=false;
  const paint=()=>{
    if(!layers)return;
    for(let i=0;i<3;i++){
      const ctx=canvases[i].getContext('2d');ctx.putImageData(new ImageData(new Uint8ClampedArray(layers[i]),160,144),0,0);
      if($('raster-guide').checked){ctx.fillStyle='rgba(23,111,158,.85)';ctx.fillRect(0,line,160,1);}
    }
  };
  function message(parent,text){parent.replaceChildren(el('p',text));}
  function reset(){capture=null;latest=query=0;valid=[];layers=null;pending=false;views.enable('render',false);views.select('play');detail.replaceChildren();history.replaceChildren();}
  function seek(target){
    if(!capture||!valid.length)return;
    target=valid.reduce((best,n)=>Math.abs(n-target)<Math.abs(best-target)?n:best,valid[0]);
    pending=true;query=0;slider.value=target;position.textContent='Line '+target;
    note.textContent='Reconstructing state at line '+target+'…';
    latest=send('raster-seek',{capture:capture.id,line:target});
  }
  function setCapture(c){
    layers=null;pending=false;latest=query=0;detail.replaceChildren();history.replaceChildren();
    capture=c;valid=(c.raster?.lines||[]).map(n=>n.line);views.enable('render',true);views.select('render',{focus:true});
    $('raster-markers').replaceChildren();
    for(const row of c.raster?.lines||[])if(row.changes){const tick=el('button');tick.type='button';tick.style.left=(row.line/143*100)+'%';tick.title=`Line ${row.line}: ${row.changes} video changes`;tick.setAttribute('aria-label',tick.title);tick.onclick=()=>seek(row.line);$('raster-markers').append(tick);}
    for(const b of root.querySelectorAll('[data-line-nav]'))b.disabled=!valid.length;slider.disabled=!valid.length;
    for(const canvas of canvases)canvas.getContext('2d').clearRect(0,0,160,144);
    if(valid.length)seek(valid[0]);
    else {position.textContent='No recorded lines';$('raster-state-line').textContent='No visible scanlines';$('raster-registers').replaceChildren();$('raster-changes').replaceChildren();$('raster-window-line').textContent='';note.textContent='No visible scanlines were captured. The LCD may be disabled. Return to Play and advance the game.';message(detail,'Capture another frame to inspect its layers.');}
  }
  function sourceButton(parent,label,address,size,before,expected){
    const b=el('button',label);b.type='button';b.onclick=()=>{if(pending||!capture)return;query=send('source',{capture:capture.id,address,size,before,expected});message(history,'Reading historical writes…');};parent.append(b);
  }
  function field(parent,label,value){const p=el('p');p.append(el('strong',label+': '),document.createTextNode(value));parent.append(p);}
  function inspect(panel,x,y){
    if(!capture||pending||!layers)return;
    query=send('raster-pixel',{capture:capture.id,panel,x,y});message(detail,'Inspecting pixel…');history.replaceChildren();
  }
  for(const [panel,canvas]of canvases.entries())canvas.addEventListener('click',e=>{const p=pixelCoordinates(canvas.getBoundingClientRect(),160,144,e.clientX,e.clientY);if(p)inspect(panel,p.x,p.y);});
  $('raster-inspect').onclick=()=>inspect(Number($('raster-panel').value),Number($('raster-x').value),Number($('raster-y').value));
  $('raster-resume').onclick=resume;slider.oninput=()=>seek(Number(slider.value));
  $('raster-guide').onchange=paint;
  for(const b of root.querySelectorAll('[data-line-nav]'))b.onclick=()=>{const i=valid.indexOf(line);seek(b.dataset.lineNav==='first'?valid[0]:b.dataset.lineNav==='last'?valid.at(-1):valid[Math.max(0,Math.min(valid.length-1,i+(b.dataset.lineNav==='next'?1:-1)))]);};
  function showState(info){
    line=info.line;slider.value=line;position.textContent='Line '+line+' / 143';$('raster-state-line').textContent='State at line '+line;
    note.textContent=`Screen contains captured lines through ${line}. Source panels hold this line’s settings fixed.`;
    if(!info.complete||valid.length!==144)note.textContent+=' This capture is incomplete; missing lines remain transparent.';
    const regs=$('raster-registers');regs.replaceChildren();
    for(const r of info.registers){const entry=el('div');entry.append(el('dt',r.name),el('dd',`${r.value} (${hex(r.value,2)})`));regs.append(entry);}
    $('raster-window-line').textContent='Window row counter: '+info.windowLine;
    const changes=$('raster-changes');changes.replaceChildren();
    for(const c of info.changes){const p=el('p');p.append(el('strong',c.name+' '),document.createTextNode(`${c.before} → ${c.after}`),el('span','PC '+hex(c.pc),'raster-writer'));changes.append(p);}
    const counts=[];if(info.tileWrites)counts.push(info.tileWrites+' tile-data writes');if(info.mapWrites)counts.push(info.mapWrites+' tilemap writes');if(info.objectWrites)counts.push(info.objectWrites+' sprite-attribute writes');
    if(counts.length)changes.append(el('p',counts.join(', ')));
    if(!info.changes.length&&!counts.length)changes.append(el('p','No video changes since the previous rendered line.'));
    if(info.changeCount>info.changes.length)changes.append(el('p','Register change list truncated.'));
    message(detail,'Select a pixel in any panel to follow its tile, sprite attributes or palette.');history.replaceChildren();
  }
  function showPixel(p){
    detail.replaceChildren();history.replaceChildren();if(p.error){message(detail,p.error);return;}
    $('raster-panel').value=p.panel;$('raster-x').value=p.x;$('raster-y').value=p.y;
    const title=['Background / window','Sprites','Screen so far'][p.panel];
    field(detail,title,`(${p.x}, ${p.y}), state at line ${p.stateLine}`);
    if(p.panel<2)detail.append(el('p','Layer preview with settings frozen at this line. Other rows are illustrative.','raster-context'));
    else field(detail,'Reconstruction',p.complete?'Layer decisions match the captured pixel.':'Incomplete or mismatched evidence.');
    for(const c of p.candidates){
      const group=el('div',undefined,'raster-candidate');const verdict=c.flags&4?'Transparent color':c.flags&8?'Hidden by another sprite':c.flags&2?'Behind background':c.kind==='Sprite'?'Composited sprite':'Base layer';
      group.append(el('strong',c.kind+(c.object>=0?' '+c.object:'')),el('p',verdict));
      if(c.sourceAddress)sourceButton(group,'Tile row '+hex(c.sourceAddress),c.sourceAddress,2,p.sourceBefore,c.sourceValue);
      if(c.mapAddress)sourceButton(group,'Tilemap entry '+hex(c.mapAddress),c.mapAddress,1,p.sourceBefore,c.mapValue);
      if(c.descriptor)sourceButton(group,'Sprite attributes '+hex(c.descriptor),c.descriptor,4,p.sourceBefore,c.descriptorValue);
      sourceButton(group,'Palette '+hex(c.paletteAddress),c.paletteAddress,1,p.sourceBefore,c.paletteValue);detail.append(group);
    }
    if(!p.candidates.length)detail.append(el('p','No selected sprite covers this position.'));
  }
  function result(m){
    if(!capture||m.capture!==capture.id)return;
    if(m.type==='raster-seek'&&m.request===latest){
      pending=false;if(m.info.error){message(detail,m.info.error);note.textContent=m.info.error;return;}
      layers=m.layers;showState(m.info);paint();return;
    }
    if(m.request!==query||pending)return;
    if(m.type==='raster-pixel')showPixel(m.evidence);
    else if(m.type==='source'){
      const p=m.evidence;history.replaceChildren();if(p.error){message(history,p.error);return;}
      field(history,'Memory',hex(p.address));field(history,'At this scanline',p.complete?'Historical bytes match the sampled source.':'History is incomplete.');
      for(const w of p.contributors||[])field(history,'CPU write',`${hex(w.before,w.size*2)} → ${hex(w.after,w.size*2)}, PC ${hex(w.pc)}`);
      if(!p.contributors?.length)history.append(el('p','These contents were already present when capture began.'));
      if(p.truncated||p.overflow)history.append(el('p','The displayed writer history is truncated.'));
    }
  }
  reset();return {reset,setCapture,result,seek,isInspecting:()=>!!capture};
}
