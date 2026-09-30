import {hex,parseAddress,pixelRange} from './memory-model.js';

export function createMemoryWorkspace({root,views,send,transport,platform}){
 root.innerHTML=`<div class="memory-heading"><div><h2>Memory</h2><p id="memory-note" role="status">Load a game to inspect its memory.</p></div><button id="memory-refresh">Snapshot now</button></div>
 <div class="memory-player"><canvas id="memory-screen" tabindex="0" width="320" height="240" aria-label="Live game output"></canvas><div><div class="memory-transport"><button id="memory-play">Play</button><button id="memory-pause" disabled>Pause</button><button id="memory-step">Next frame</button></div><p id="memory-live-status">Paused</p><div id="memory-tape" hidden><button id="memory-tape-play">Play tape</button><button id="memory-tape-stop">Stop tape</button></div><p>Live memory updates up to five times per second. Click the preview for keyboard controls. Access highlights cover the latest update.</p></div></div>
 <div class="memory-toolbar"><form id="memory-go"><label>Address or bank:offset <input id="memory-address" placeholder="$D200 or rom-7:2E30" spellcheck="false"></label><button>Go</button></form><label><input id="memory-reads" type="checkbox" checked> Reads</label><label><input id="memory-writes" type="checkbox" checked> Writes</label><label><input id="memory-fetches" type="checkbox"> Fetches</label><label>Activity window <select id="memory-window"><option value="20000">Recent 20,000 events</option><option value="0">Whole recording to cursor</option></select></label><label>Map scale <select id="memory-scale"><option value="0">Fit overview</option><option value="1">Detailed (up to 1 byte / pixel)</option></select></label></div>
 <div class="memory-layout"><section class="memory-hex"><h3 id="memory-title">Hex view</h3><p id="memory-aliases"></p><div id="memory-bytes" tabindex="0" aria-label="Hex bytes. Arrow keys select, Page Up and Page Down scroll."></div><div class="memory-paging"><button id="memory-prev">Previous page</button><span id="memory-range"></span><button id="memory-next">Next page</button></div><p id="memory-detail" aria-live="polite"></p></section>
 <section class="memory-map"><h3>Memory atlas</h3><p>Byte brightness · <span class="memory-read-key">reads</span> · <span class="memory-write-key">writes</span> · <span class="memory-both-key">both</span><br>Outline = hex viewport. Click to jump.</p><div id="memory-atlas"></div></section>
 <aside class="memory-regions"><h3>Regions & labels</h3><input id="memory-search" type="search" placeholder="Search regions or labels" aria-label="Search regions or labels"><div id="memory-labels"></div><p id="memory-source"></p></aside></div>
 <section class="memory-record"><div><button id="memory-record">Record & run</button><label>Maximum duration <select id="memory-duration"><option value="0.1">0.1 seconds</option><option value="1" selected>1 second</option><option value="5">5 seconds</option></select></label><button id="memory-stop" disabled>Stop recording</button><span id="memory-capability"></span></div><div id="memory-history" hidden><button id="memory-replay">Replay</button><button id="memory-event-prev">Previous event</button><input id="memory-cursor" type="range" min="0" max="0" value="0" aria-label="Recorded event position"><button id="memory-event-next">Next event</button><span id="memory-position"></span></div></section>`;
 const $=id=>root.querySelector('#memory-'+id);
 let data=null,selected=null,selectedLabel=null,offset=0,page=null,pending=0,detailPending=0,enabled=false,busy=false,playing=false,timer=null,active=false,running=false,liveTimer=null,livePending=0,finalRefresh=false;
 const canvases=new Map();
 function stopReplay(){playing=false;clearTimeout(timer);$('replay').textContent='Replay';}
 function request(type,args={}){if(busy&&!['memory-snapshot','memory-record'].includes(type))return pending;pending=send(type,args);return pending;}
 function snapshot(){if(busy)return;stopReplay();busy=true;$('record').disabled=$('refresh').disabled=$('play').disabled=true;$('note').textContent='Pausing and reading physical storage…';request('memory-snapshot');}
 views.register({id:'memory',label:'Memory',panel:root,enabled:false});
 function select(region,at=0,reveal=true){
  if(!data)return;selected=data.regions.find(r=>r.id===region);if(!selected)return;
  offset=Math.max(0,Math.min(at,selected.size-1));$('title').textContent=selected.name;
  $('aliases').textContent=`${selected.size.toLocaleString()} bytes · physical base ${hex(selected.base,selected.base>65535?8:4)}${selected.mappings?' · CPU windows in this snapshot: '+(selected.mappings.map(m=>hex(m.cpu)+'–'+hex(m.cpu+m.size-1)).join(', ')||'unmapped'):selected.aliases.length?' · CPU aliases in this snapshot: '+selected.aliases.map(x=>hex(x)).join(', '):' · no CPU alias'}`;
  if(document.activeElement!==$('address'))$('address').value=selected.id+':'+offset.toString(16).toUpperCase();
  request('memory-page',{region:selected.id,offset,snapshot:data.id});
  detailPending=send('memory-detail',{region:selected.id,offset,snapshot:data.id});
  if(reveal){const c=canvases.get(region),scroller=$('atlas');if(c){const box=c.getBoundingClientRect(),view=scroller.getBoundingClientRect();if(box.bottom>view.bottom)scroller.scrollTop+=box.bottom-view.bottom;if(box.top<view.top)scroller.scrollTop+=box.top-view.top;}}
 }
 function drawMap(r){
  const c=canvases.get(r.id);if(!c)return;const ctx=c.getContext('2d'),n=r.bitmap.length;
  c.width=256;c.height=Math.max(1,Math.ceil(n/256));const im=ctx.createImageData(c.width,c.height);
  for(let i=0;i<n;i++){const v=r.bitmap[i];im.data.set([v,v,v,255],i*4);}ctx.putImageData(im,0,0);
  const filter=($('reads').checked?1:0)|($('writes').checked?2:0)|($('fetches').checked?4:0);
  for(let p=0;p<r.activityMap.length;p++){const kind=r.activityMap[p]&filter;if(!kind)continue;ctx.fillStyle=(kind&3)===3?'#cb7dcc':(kind&2)?'#ef883b':(kind&1)?'#38a5e6':'#af78cd';ctx.fillRect(p%256,Math.floor(p/256),1,1);}
  if(selectedLabel?.region===r.id&&selectedLabel.length){ctx.fillStyle='rgba(235,193,75,.4)';for(let p=Math.floor(selectedLabel.start/r.scale);p<Math.ceil((selectedLabel.start+selectedLabel.length)/r.scale);p++)ctx.fillRect(p%256,Math.floor(p/256),1,1);}
  if(r.kind==='tape'&&data.tapeCursor!==null){const p=Math.floor(data.tapeCursor/r.scale);ctx.fillStyle='#ffc541';ctx.fillRect(p%256,Math.floor(p/256),2,3);}
  if(selected?.id===r.id&&page){
   // Tape coordinates are pulse indices; worker supplies the selected pulse range.
   const start=r.kind==='tape'?page.pulseStart:page.offset,end=r.kind==='tape'?page.pulseEnd:page.offset+page.bytes.length;
   if(start!==undefined){let a=Math.floor(start/r.scale),b=Math.max(a+1,Math.ceil(end/r.scale));ctx.strokeStyle='#ffffff';ctx.lineWidth=1;
    for(let p=a;p<b;){const stop=Math.min(b,(Math.floor(p/256)+1)*256);ctx.strokeRect(p%256+.5,Math.floor(p/256)+.5,Math.max(1,stop-p)-1,1);ctx.fillStyle='#4d578f';ctx.fillRect(p%256,Math.floor(p/256)-1,stop-p,1);p=stop;}
   }
  }
 }
 function atlas(){
  $('atlas').replaceChildren();canvases.clear();
  for(const r of data.regions){const figure=document.createElement('figure'),caption=document.createElement('figcaption'),c=document.createElement('canvas');
   caption.textContent=`${r.name} · ${r.size.toLocaleString()} bytes · ${r.scale} ${r.units}/pixel`;
   c.dataset.mapSize=r.mapSize;c.dataset.mapScale=r.scale;c.tabIndex=0;c.setAttribute('aria-label',r.name+'. Click to inspect; Enter selects the start.');
   c.onkeydown=e=>{if(e.key==='Enter'){e.preventDefault();select(r.id);}};
   c.onclick=e=>{const box=c.getBoundingClientRect(),pixel=Math.floor((e.clientY-box.top)*c.height/box.height)*256+Math.floor((e.clientX-box.left)*256/box.width);if(!pixelRange(pixel,r.scale,r.mapSize))return;request('memory-map',{region:r.id,pixel,scale:r.scale,snapshot:data.id});};
   figure.append(caption,c);$('atlas').append(figure);canvases.set(r.id,c);drawMap(r);
  }
 }
 function labels(){
  const q=$('search').value.toLowerCase();$('labels').replaceChildren();
  const items=[...data.regions.map(r=>({region:r.id,name:r.name,start:0,description:r.kind})),...(data.labels||[])];
  for(const l of items.filter(l=>(l.name+' '+l.region+' '+hex(l.start)+' '+l.description).toLowerCase().includes(q))){const b=document.createElement('button');b.textContent=l.name;b.title=l.description;b.onclick=()=>{selectedLabel=l;select(l.region,l.start);$('source').replaceChildren(document.createTextNode(`${l.description||''}${l.length?' · '+l.length+' bytes':''} `));if(l.source){const a=document.createElement('a');a.href=new URL(l.source,location.origin);a.textContent='Reference';a.target='_blank';a.rel='noopener';$('source').append(a);}};$('labels').append(b);}
 }
 function showPage(m){
  page=m;const box=$('bytes');box.replaceChildren();
  for(let row=0;row<m.bytes.length;row+=16){const line=document.createElement('div');line.className='memory-row';const address=document.createElement('span');address.className='memory-row-address';address.textContent=hex((selected.kind==='rom'?selected.base:(selected.aliases[0]??selected.base))+m.offset+row,selected.base>65535?8:4);line.append(address);
   for(let i=row;i<Math.min(row+16,m.bytes.length);i++){const b=document.createElement('button');b.type='button';b.textContent=m.bytes[i].toString(16).toUpperCase().padStart(2,'0');b.tabIndex=-1;b.className=m.offset+i===offset?'selected':'';b.setAttribute('aria-label',hex(m.offset+i)+' = '+b.textContent);b.onclick=()=>select(selected.id,m.offset+i,false);line.append(b);}
   const ascii=document.createElement('span');ascii.className='memory-ascii';ascii.textContent=Array.from(m.bytes.slice(row,row+16),x=>x>=32&&x<127?String.fromCharCode(x):'·').join('');line.append(ascii);box.append(line);
  }
  $('range').textContent=hex(m.offset)+'–'+hex(m.offset+m.bytes.length-1);$('prev').disabled=m.offset===0;$('next').disabled=m.offset+m.bytes.length>=selected.size;
  data.regions.forEach(drawMap);
 }
 function result(m){
  if(m.type==='memory-progress'){$('note').textContent=m.text;return;}
  if(m.type==='memory-error'){if(m.request===livePending){livePending=0;schedule();$('note').textContent=m.text;return;}if(m.request!==pending)return;busy=false;$('refresh').disabled=$('play').disabled=false;stopReplay();$('note').textContent=m.text;$('record').disabled=!data?.activity;$('stop').disabled=true;return;}
  if(m.type==='memory-detail'){if(m.request!==detailPending||m.snapshot!==data?.id)return;const d=m.detail;$('detail').textContent=d?`${d.pulse===undefined?(selected?.kind==='disk'?`Sector ${Math.floor(d.offset/512)} · byte ${d.offset%512}. Media events refer to encoded MFM payload consumption. `:''):`Pulse ${d.pulse} · ${d.duration} cycles. `}${d.events.map(e=>`${e.kind&2?'Write':e.kind&4?'Fetch':'Read'} ${hex(e.value)} at cycle ${e.cycle}, PC ${hex(e.pc)}`).join(' · ')||'No recorded accesses at this byte.'}`:'';return;}
  const liveUpdate=m.type==='memory-overview'&&m.request===livePending;
  if(liveUpdate)livePending=0;
  if(!liveUpdate&&m.request!==pending)return;
  if(m.type==='memory-page'){if(m.page&&m.page.id===data?.id)showPage(m.page);schedule();return;}
  if(m.type==='memory-map'){select(m.region,m.offset,false);return;}
  if(m.type!=='memory-overview')return;
  data=m.overview;if(!data)return;busy=false;$('refresh').disabled=$('play').disabled=running;
  const old=selected?.id;$('note').textContent=data.note||`${data.live?'Live memory':data.recording?'Historical recording':'Paused snapshot'} · ${data.regions.length} physical regions${data.dropped||data.limited?' · Activity limit reached; some accesses omitted':''}`;
  $('capability').textContent=data.activity?data.coverage:'Live snapshots · access recording unavailable for this core';
  $('record').disabled=running||!data.activity;$('stop').disabled=true;$('fetches').disabled=!data.fetches;$('fetches').parentElement.title=data.fetches?'Record instruction fetches too':(data.coverage||'No access recording');$('history').hidden=!data.recording;$('window').disabled=!!data.live;
  $('cursor').max=data.count;$('cursor').value=data.position;$('position').textContent=`${data.position.toLocaleString()} / ${data.count.toLocaleString()} events${data.cursorCycle!==undefined?' · cycle/step '+data.cursorCycle.toLocaleString():''}`;
  page=null;const reusable=liveUpdate&&data.regions.length===canvases.size&&data.regions.every(r=>{const c=canvases.get(r.id);return c&&+c.dataset.mapSize===r.mapSize&&+c.dataset.mapScale===r.scale;});if(reusable)data.regions.forEach(drawMap);else{atlas();labels();}if(data.regions.length)select(data.regions.some(r=>r.id===old)?old:data.regions[0].id,offset,false);
  if(liveUpdate&&finalRefresh){finalRefresh=false;livePending=send('memory-live-snapshot',{scale:+$('scale').value,fetches:$('fetches').checked});}
  if(!data.regions.length)schedule();
  if(playing){if(data.position>=data.count)stopReplay();else timer=setTimeout(()=>request('memory-seek',{position:Math.min(data.count,data.position+Math.max(1,Math.ceil(data.count/100))),scale:+$('scale').value,window:+$('window').value}),80);}
 }
 $('refresh').onclick=snapshot;
 function poll(){clearTimeout(liveTimer);liveTimer=null;if(!active||busy||livePending||!running)return;livePending=send('memory-live-snapshot',{scale:+$('scale').value,window:+$('window').value,fetches:$('fetches').checked});}
 function schedule(){clearTimeout(liveTimer);liveTimer=null;if(active&&running)liveTimer=setTimeout(poll,200);}
 for(const id of ['play','pause','step'])$(id).onclick=()=>{stopReplay();if(id!=='pause'){busy=false;data=null;livePending=send('memory-live-snapshot',{scale:+$('scale').value,fetches:$('fetches').checked});}transport?.(id==='play'?'run':id);};
 $('tape').hidden=platform!=='c64';$('tape-play').onclick=()=>send('tape',{down:true});$('tape-stop').onclick=()=>send('tape',{down:false});

 $('go').onsubmit=e=>{e.preventDefault();const found=parseAddress($('address').value,data?.regions||[],selected);if(found)select(found.region.id,found.offset);else $('detail').textContent='Address is outside the selected region. Use region-id:offset for a bank.';};
 $('prev').onclick=()=>select(selected.id,Math.max(0,page.offset-256),false);$('next').onclick=()=>select(selected.id,page.offset+256,false);
 $('bytes').onkeydown=e=>{const delta={ArrowLeft:-1,ArrowRight:1,ArrowUp:-16,ArrowDown:16,PageUp:-256,PageDown:256}[e.key];if(delta&&selected){e.preventDefault();select(selected.id,offset+delta,false);}};
 $('bytes').onwheel=e=>{if(selected){e.preventDefault();select(selected.id,offset+(e.deltaY>0?48:-48),false);}};
 $('search').oninput=()=>data&&labels();$('window').onchange=$('scale').onchange=()=>request('memory-overview',{scale:+$('scale').value,window:+$('window').value});
 for(const name of ['reads','writes','fetches'])$(name).onchange=()=>data?.regions.forEach(drawMap);
 $('record').onclick=()=>{stopReplay();busy=true;$('refresh').disabled=$('play').disabled=true;$('record').disabled=true;$('stop').disabled=false;request('memory-record',{duration:+$('duration').value,fetches:$('fetches').checked});};
 $('stop').onclick=()=>send('memory-stop');
 $('cursor').oninput=()=>{stopReplay();request('memory-seek',{position:+$('cursor').value,scale:+$('scale').value,window:+$('window').value});};
 $('event-prev').onclick=()=>{stopReplay();request('memory-seek',{position:data.position-1,scale:+$('scale').value,window:+$('window').value});};
 $('event-next').onclick=()=>{stopReplay();request('memory-seek',{position:data.position+1,scale:+$('scale').value,window:+$('window').value});};
 $('replay').onclick=()=>{if(playing){stopReplay();return;}playing=true;$('replay').textContent='Stop replay';request('memory-seek',{position:data.position>=data.count?0:data.position,scale:+$('scale').value,window:+$('window').value});};
 function setActive(value){active=value;clearTimeout(liveTimer);liveTimer=null;if(!value){livePending=0;send('memory-live-end');}else schedule();}
 function reset(){setActive(false);stopReplay();data=null;selected=null;selectedLabel=null;page=null;pending=detailPending=0;enabled=false;busy=false;views.enable('memory',false);$('atlas').replaceChildren();$('bytes').replaceChildren();$('labels').replaceChildren();$('detail').textContent='';$('source').textContent='';$('history').hidden=true;}
 return {result,reset,setActive,present(canvas){if(active){const c=$('screen');if(c.width!==canvas.width)c.width=canvas.width;if(c.height!==canvas.height)c.height=canvas.height;c.getContext('2d').drawImage(canvas,0,0);}},state(m){const was=running;running=m.running;const blocked=!!(m.capturing||m.saving||m.memoryRecording);$('play').disabled=running||blocked;$('pause').disabled=!running;$('step').disabled=running||blocked;$('refresh').disabled=running||blocked;$('record').disabled=running||blocked||!data?.activity;$('live-status').textContent=`${running?'Running':'Paused'} · display ${m.state.frames.toLocaleString()}${m.state.pulse!==undefined?' · tape pulse '+m.state.pulse.toLocaleString():''}`;if(active&&!blocked){if(was&&!running&&livePending)finalRefresh=true;else if(was&&!running)livePending=send('memory-live-snapshot',{scale:+$('scale').value,fetches:$('fetches').checked});else if(running&&!liveTimer&&!livePending)schedule();}},ready(){enabled=true;views.enable('memory',true);},open(){if(enabled&&!data){if(running)poll();else snapshot();}},invalidate(){stopReplay();data=null;pending=detailPending=0;}};
}
