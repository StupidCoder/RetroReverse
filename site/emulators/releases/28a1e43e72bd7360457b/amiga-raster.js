import {localElement} from './panel-dom.js';
const hex=(n,d=4)=>'$'+(Number(n)>>>0).toString(16).padStart(d,'0');
const swap=n=>((n&255)<<8)|(n>>>8&255);
const el=(tag,text,cls)=>{const n=document.createElement(tag);if(text!==undefined)n.textContent=text;if(cls)n.className=cls;return n;};
const memory=a=>a>=0x100000?hex(0xdff000+a-0x100000,6):'Chip RAM '+hex(a,5);
const byteText=(n,size)=>Array.from({length:size},(_,i)=>hex(n>>>(8*i)&255,2)).join(' ');
export function createAmigaRaster({send,ui}){
 const timeline=ui.timeline;
 ui.toolbar.hidden=ui.auxiliary.hidden=false;
 ui.toolbar.innerHTML=`<div role="group" aria-label="Render view"><button id="amiga-scan" aria-pressed="true">Scanlines &amp; Copper</button><button id="amiga-blitter" aria-pressed="false">Blitter &amp; masks</button></div><label id="amiga-plane-label">Playfield <select id="amiga-plane"><option value="-1">Combined</option>${Array.from({length:6},(_,i)=>`<option value="${i}">Bitplane ${i+1}</option>`).join('')}</select></label><label id="amiga-filter-label" hidden>Operations <select id="amiga-filter"><option value="all">All blits</option><option value="mask">Cookie-cut masks</option><option value="copy">Copies</option></select></label><label id="amiga-operation-label" hidden>Blit <select id="amiga-operation"></select></label>`;
 ui.auxiliary.innerHTML='<div id="amiga-scan-board" class="buffer-stack"></div><div id="amiga-blit-board" class="buffer-stack" hidden><div class="amiga-channel-grid"></div><div id="amiga-before-slot"></div></div>';
 ui.sidebar.innerHTML=`<section><h3 id="amiga-state-heading">Video settings</h3><div id="amiga-state"></div><div id="amiga-palette" aria-label="Palette"></div></section>
   <section><h3 id="amiga-changes-heading">Copper and register writes</h3><div id="amiga-changes"></div></section>
   <section><h3>Pixel and source</h3><div class="raster-pixel-controls"><label id="amiga-panel-label">Layer <select id="amiga-panel"><option value="2">Screen</option><option value="0">Playfield</option><option value="1">Sprites</option></select></label><label>X<input id="amiga-x" type="number" min="0" value="320"></label><label>Y<input id="amiga-y" type="number" min="0" value="128"></label><button id="amiga-inspect">Inspect</button></div><div id="amiga-pixel"></div><div id="amiga-history" aria-live="polite"></div></section>`;
 const guide=timeline.toggle('amiga-guide','Show scanline',{checked:true,onChange:()=>{paint();highlightChange();}});
 const shared={position:ui.position,cursor:timeline.label,timeline:timeline.slider,markers:timeline.markers,note:timeline.note,guide:guide.input,'guide-text':guide.caption};
 const $=id=>shared[id]||localElement(ui.root,'amiga-'+id);
 const scanBuffers=[ui.buffer($('scan-board'),{id:'amiga-playfield',title:'Playfield memory',description:'Includes objects drawn by the blitter'}),ui.buffer($('scan-board'),{id:'amiga-sprites',title:'Hardware sprites',description:'Before playfield priority'}),ui.output];
 const channels=[['a','A · mask','After edge masks and shift'],['b','B · image','After shift'],['c','C · background','Read before writing'],['d','D · result','The Boolean function combines A, B and C']];
 const blitBuffers=channels.map(([id,title,description])=>ui.buffer(ui.auxiliary.querySelector('.amiga-channel-grid'),{id:'amiga-'+id,title,description,fit:'contain',aspect:'2/1'}));
 blitBuffers.push(ui.buffer($('before-slot'),{id:'amiga-before',title:'Before blit',description:'Final frame layout, memory before this blit'}),ui.output);
 const blitCanvases=blitBuffers.map(b=>b.canvas);
 let capture=null,mode='scan',line=0,blit=0,blitIDs=[],latest=0,query=0,pending=false,layers=null,info=null;
 const message=(parent,text)=>parent.replaceChildren(el('p',text));
 function field(parent,name,value){const p=el('p');p.append(el('strong',name+': '),document.createTextNode(value));parent.append(p);}
 function reset(){capture=null;query=latest=0;pending=false;layers=info=null;$('pixel').replaceChildren();$('history').replaceChildren();}
 function blitFilter(){blitIDs=(capture?.raster?.blits||[]).filter(b=>$('filter').value==='all'||($('filter').value==='mask'?b.kind==='Cookie cut':b.kind.startsWith('Copy'))).map(b=>b.index);$('operation').replaceChildren();for(const i of blitIDs){const b=capture.raster.blits[i],option=el('option',`${i+1} · ${b.kind} · ${b.width}×${b.height}`);option.value=i;$('operation').append(option);}}
 function request(){
  if(!capture)return;pending=true;query=0;layers=null;$('pixel').replaceChildren();$('history').replaceChildren();
  if(mode==='scan'){latest=send('raster-seek',{capture:capture.id,line,plane:Number($('plane').value)});timeline.range({max:255,positions:(capture.raster?.lines||[]).map(l=>l.line)});timeline.value(line,`Line ${line} / 255`);}
  else if(blitIDs.length){if(!blitIDs.includes(blit))blit=blitIDs[0];latest=send('blit-seek',{capture:capture.id,index:blit});timeline.range({max:capture.raster.blits.length-1,positions:blitIDs});timeline.value(blit,`Blit ${blit+1} / ${capture.raster.blits.length}`);$('operation').value=blit;}
  else {pending=false;$('cursor').textContent='No matching blits';message($('state'),'No matching operation was recorded in these display intervals.');$('changes').replaceChildren();$('note').textContent='Choose another filter or capture another frame.';for(const c of blitCanvases)c.getContext('2d').clearRect(0,0,c.width,c.height);}
  if(mode==='blit'&&!blitIDs.length)timeline.range({disabled:true});
 }
 function setMode(next){mode=next;$('scan-board').hidden=mode!=='scan';$('blit-board').hidden=mode!=='blit';$('scan').setAttribute('aria-pressed',mode==='scan');$('blitter').setAttribute('aria-pressed',mode==='blit');$('plane-label').hidden=mode!=='scan';$('filter-label').hidden=mode!=='blit';$('operation-label').hidden=mode!=='blit';$('guide-text').textContent=mode==='scan'?'Show scanline':'Highlight change';$('panel-label').hidden=mode!=='scan';$('markers').hidden=mode!=='scan';$('palette').hidden=mode!=='scan';ui.output.labels(mode==='scan'?'Screen so far':'Playfield after',mode==='scan'?'Each completed row keeps its original colors':'Final frame layout, memory after this blit');ui.output.setInspect(mode==='scan'?(x,y)=>inspect(2,x,y):null);request();}
 function setCapture(c){capture=c;line=(c.raster?.lines||[])[0]?.line||0;blit=0;blitFilter();
  timeline.range({max:255});timeline.setMarkers((c.raster?.lines||[]).filter(row=>row.changes).map(row=>({value:row.line,label:`Line ${row.line}: ${row.changes} video writes`})));
  setMode('scan');
 }
 function paint(){if(!layers||!info)return;const buffers=mode==='scan'?scanBuffers:blitBuffers;for(let p=0;p<buffers.length;p++){const w=mode==='blit'&&p<4?info.width:640,h=mode==='blit'&&p<4?info.height:256;buffers[p].draw(layers[p],w,h);if(mode==='scan'&&$('guide').checked){const ctx=buffers[p].canvas.getContext('2d');ctx.fillStyle='rgba(23,111,158,.85)';ctx.fillRect(0,line,640,1);}}}
 function highlightChange(){
  if(mode!=='blit'||!layers||!$('guide').checked)return;
  const before=new Uint32Array(layers[4]),after=new Uint32Array(layers[5]);let left=640,top=256,right=-1,bottom=-1;
  for(let i=0;i<before.length;i++)if(before[i]!==after[i]){const x=i%640,y=Math.floor(i/640);left=Math.min(left,x);top=Math.min(top,y);right=Math.max(right,x);bottom=Math.max(bottom,y);}
  if(right>=0)for(const c of blitCanvases.slice(4)){const ctx=c.getContext('2d');ctx.strokeStyle='#e37b32';ctx.lineWidth=3;ctx.strokeRect(Math.max(1,left-2),Math.max(1,top-2),right-left+5,bottom-top+5);}
 }
 function sourceButton(parent,s){const b=el('button',s.role+' · '+memory(s.address));b.type='button';b.onclick=()=>{if(pending)return;query=send('source',{capture:capture.id,address:s.address,size:s.size,before:s.before,expected:s.value});message($('history'),'Reading historical writes…');};parent.append(b);}
 function showScan(s){
  line=s.line;$('position').textContent=`Final captured frame ${s.frame}, row ${line} (PAL line ${s.raster})`;$('cursor').textContent=`Line ${line} / 255`;
  $('note').textContent='Source panels hold memory and settings from this line fixed. Completed screen rows retain their actual colors.'+(!s.complete?' Some capture evidence is missing.':'');
  $('state-heading').textContent='Video settings';$('changes-heading').textContent='Copper and register writes';$('state').replaceChildren();
  field($('state'),'Playfield',`${s.planes} bitplanes · ${s.mode}`);
  const regs=el('details');regs.append(el('summary','Registers and plane addresses'));for(const r of s.registers)field(regs,r.name,hex(r.value));s.pointers.slice(0,s.planes).forEach((p,i)=>field(regs,'Bitplane '+(i+1),hex(p,5)));$('state').append(regs);
  $('palette').replaceChildren();s.palette.forEach((c,i)=>{const b=el('button',String(i));b.type='button';const css='#'+c.toString(16).padStart(3,'0').slice(-3);b.style.background=css;b.style.color=((c>>8&15)*3+(c>>4&15)*6+(c&15))>75?'#111':'#fff';b.title=`COLOR${i} = ${hex(c,3)}`;b.onclick=()=>{const matching=s.changes.filter(v=>v.register===0x180+i*2).at(-1);if(matching){query=send('source',{capture:capture.id,address:0x100180+i*2,size:2,before:matching.cutoff,expected:swap(c)});message($('history'),'Reading palette writers…');}else message($('history'),`COLOR${i} = ${hex(c,3)}. No write to this color on the selected line. Select a pixel to search earlier history.`);};$('palette').append(b);});
  $('changes').replaceChildren();for(const c of s.changes){const b=el('button',`${c.name} ${hex(c.before)} → ${hex(c.value)}`,'raster-write');b.type='button';b.append(el('span',`${c.copperPC?'Copper '+hex(c.copperPC,5):'CPU '+hex(c.pc,6)} · PAL ${c.line}, H ${hex(c.beam/2,2)}`,'raster-writer'));b.onclick=()=>{query=send('source',{capture:capture.id,address:0x100000+c.register,size:2,before:c.cutoff,expected:swap(c.value)});message($('history'),'Reading register history…');};$('changes').append(b);}
  if(!s.changes.length)message($('changes'),'No video-register writes on this line. Markers below the slider identify changes.');
  message($('pixel'),'Select a pixel to follow its plane words or palette to CPU, Copper and blitter writers.');
 }
 function showBlit(s){
  blit=s.index;$('position').textContent=`${s.kind} in frame ${s.frame}, PAL line ${s.line}`;$('cursor').textContent=`Blit ${s.index+1} / ${s.count}`;
  $('note').textContent='One-bit inputs and result, after masking and shifts. Playfield previews use the final frame’s layout; writes to other buffers may not appear.'+(!s.complete?' Some capture evidence is missing.':'');
  $('state-heading').textContent='Blitter operation';$('changes-heading').textContent='Shifts, masks and channels';$('state').replaceChildren();$('changes').replaceChildren();
  const op=s.con0&255;
  $('a').closest('figure').querySelector('strong').textContent=op===0xca?'A · mask':'A · source';
  $('b').closest('figure').querySelector('strong').textContent=op===0xca?'B · image':'B · source';
  $('c').closest('figure').querySelector('strong').textContent=op===0xca?'C · background':'C · source';
  field($('state'),'Function',op===0xca?'D = (A & B) | (~A & C)':op===0xf0?'D = A':op===0xcc?'D = B':op===0?'D = 0':`Minterm ${hex(op,2)}`);
  if(op===0xca)$('state').append(el('p','Mask bit 1 selects the image from B. Mask bit 0 keeps the background from C.'));
  field($('state'),'Size',`${s.width} bits × ${s.height} rows`);field($('state'),s.copperPC?'Started by Copper':'Started by CPU',hex(s.copperPC||s.pc,6));
  field($('changes'),'A / B shift',`${s.con0>>>12} / ${s.con1>>>12} bits`);field($('changes'),'A edge masks',`${hex(s.firstMask)} / ${hex(s.lastMask)}`);field($('changes'),'Direction',s.con1&1?'Line mode (rows show successive steps)':s.con1&2?'Descending':'Ascending');
  if(!(s.con1&1)&&(s.con1&0x18))field($('changes'),'Area fill',`${s.con1&0x10?'Exclusive':'Inclusive'} fill modifies the Boolean result; initial carry ${s.con1&4?1:0}`);
  for(let p=0;p<4;p++)field($('changes'),String.fromCharCode(65+p),`${s.con1&1&&p<2?'Line pattern data register':s.con0&(0x800>>p)?memory(s.pointers[p]):p===3?'Destination DMA disabled':'Constant data register'}; modulo ${s.modulos[p]}`);
  const table=el('details');table.append(el('summary','Boolean truth table (before optional fill)'));const text=el('p');text.textContent=Array.from({length:8},(_,i)=>`${i>>2&1}${i>>1&1}${i&1} → ${op>>i&1}`).join('  ·  ');table.append(el('p','A B C → D'),text);$('changes').append(table);
  message($('pixel'),'Click a bit in any channel to inspect the fetched words, shift carry and destination write.');$('x').value=Math.min(Number($('x').value),s.width-1);$('y').value=Math.min(Number($('y').value),s.height-1);
 }
 function inspect(panel,x,y){if(!capture||pending||!layers)return;query=send(mode==='scan'?'raster-pixel':'blit-pixel',{capture:capture.id,panel,x,y});message($('pixel'),'Inspecting…');$('history').replaceChildren();}
 scanBuffers.slice(0,2).forEach((buffer,p)=>buffer.setInspect((x,y)=>inspect(p,x,y)));
 blitBuffers.slice(0,4).forEach(buffer=>buffer.setInspect((x,y)=>inspect(0,x,y)));
 function showPixel(p){$('pixel').replaceChildren();$('history').replaceChildren();if(p.error){message($('pixel'),p.error);return;}$('x').value=p.x;$('y').value=p.y;
  if(mode==='blit'){field($('pixel'),'Word / bit',`${p.word} / ${p.bit}`);field($('pixel'),'A B C → D',`${p.a>>>p.bit&1} ${p.b>>>p.bit&1} ${p.c>>>p.bit&1} → ${p.d>>>p.bit&1}`);field($('pixel'),'Destination',`${hex(p.oldD)} → ${hex(p.d)}${p.written?'':' (write suppressed)'}`);for(const s of p.sources)sourceButton($('pixel'),s);}
  else {$('panel').value=p.panel;field($('pixel'),'Pixel',`(${p.x}, ${p.y}), state at line ${p.stateLine}`);if(p.preview)$('pixel').append(el('p','Preview using frozen memory and settings.'));else field($('pixel'),'Reconstruction',p.complete?'Matches the recorded output.':'Evidence is incomplete or differs.');if(p.ham)$('pixel').append(el('p','HAM colors can also depend on preceding pixels; that chain is not expanded here.'));
   for(const c of p.candidates){const g=el('div',undefined,'raster-candidate');g.append(el('strong',c.kind+(c.object>=0?' '+c.object:'')));if(c.object>=0)g.append(el('p',c.flags&4?'Transparent':c.flags&8?'Behind another sprite':c.flags&2?'Behind playfield':'Visible sprite'));for(const s of c.sources)sourceButton(g,s);$('pixel').append(g);}
  }
 }
 function showHistory(p){$('history').replaceChildren();if(p.error){message($('history'),p.error);return;}field($('history'),'Memory',memory(p.address));field($('history'),'Reconstruction',p.complete?'Historical bytes match the sampled word.':'History is incomplete.');
  for(const w of p.contributors||[]){const c=w.command||{},g=el('div',undefined,'raster-candidate');field(g,c.origin==='copper'?'Copper MOVE':c.origin==='blitter'?'Blitter write':c.origin==='disk'?'Disk DMA':'CPU write',`${byteText(w.before,w.size)} → ${byteText(w.after,w.size)}`);
   if(c.origin==='copper')field(g,'Copper instruction',hex(c.copperPC,5));else if(c.origin==='blitter'&&c.copperPC)field(g,'Started by Copper',hex(c.copperPC,5));else field(g,c.origin==='blitter'?'Submission CPU PC':c.origin?'Concurrent CPU PC':'CPU PC',hex(w.pc,6));
   if(c.origin==='blitter'&&Number.isInteger(c.blit)){const b=el('button',`Inspect blit ${c.blit+1}`);b.type='button';b.onclick=()=>{$('filter').value='all';blitFilter();blit=c.blit;setMode('blit');};g.append(b);}$('history').append(g);
  }
  if(!p.contributors?.length)$('history').append(el('p','These bytes were already present when this capture began.'));if(p.truncated||p.overflow)$('history').append(el('p','Writer history is truncated.'));
  $('history').scrollIntoView({block:'nearest'});
 }
 function result(m){if(!capture||m.capture!==capture.id)return;
  if((m.type==='raster-seek'||m.type==='blit-seek')&&m.request===latest){pending=false;info=m.info;if(info.error){message($('pixel'),info.error);return;}layers=m.layers;mode==='scan'?showScan(info):showBlit(info);paint();highlightChange();ui.sidebar.scrollTop=0;return;}
  if(m.request!==query||pending)return;if(m.type==='raster-pixel'||m.type==='blit-pixel')showPixel(m.evidence);else if(m.type==='source')showHistory(m.evidence);
 }
 $('scan').onclick=()=>setMode('scan');$('blitter').onclick=()=>{if(mode==='scan'&&blit===0)blit=capture?.raster?.blits.find(b=>b.kind==='Cookie cut')?.index||0;setMode('blit');};$('filter').onchange=()=>{blitFilter();request();};$('operation').onchange=()=>{blit=Number($('operation').value);request();};$('plane').onchange=request;
 $('inspect').onclick=()=>inspect(Number($('panel').value),Number($('x').value),Number($('y').value));
 timeline.onSeek(value=>{if(mode==='scan')line=value;else blit=value;request();});
 reset();return{reset,setCapture,result,isInspecting:()=>!!capture};
}
