// Run in a real browser: serve the repository, then open ui-workspace.html.
// Synthetic captures exercise the UI protocol without copyrighted media.
import {tilesetBytes} from '../../../site/emulators/tileset-decode.js';
import {mountShell} from '../../../site/emulators/ui-shell.js';
import {presentation} from '../../../site/emulators/ui-platforms.js';
import {createRenderWorkspace} from '../../../site/emulators/render-workspace.js';
const $=id=>document.getElementById(id),assert=(value,message)=>{if(!value)throw Error(message);};
const nextPaint=()=>new Promise(resolve=>requestAnimationFrame(()=>requestAnimationFrame(resolve)));
function pixels(w,h,color=0xff332211){const out=new Uint32Array(w*h);out.fill(color);return out.buffer;}
function pixel(canvas){return [...canvas.getContext('2d').getImageData(0,0,1,1).data].join(',');}
function tiles(platform){if(!tilesetBytes[platform])return null;const b=new Uint8Array(tilesetBytes[platform]);b.fill(0xaa);if(platform==='c64')for(let i=0;i<16;i++)b.set([i*16,i*16,i*16,255],80+i*4);return b.buffer;}
const results=[];
try {
 for(const [platform,view] of Object.entries(presentation)){
  document.body.dataset.platform=platform;mountShell(platform);
  const play=$('screen');play.getContext('2d').fillStyle='#123456';play.getContext('2d').fillRect(0,0,play.width,play.height);const original=pixel(play);
  let sequence=0,resumed=0;const sent=[];
  const render=createRenderWorkspace({platform,presentation:view,playCanvas:play,resume:()=>resumed++,send:(type,data)=>{const m={type,...data,request:++sequence};sent.push(m);return m.request;}});
  const ids=[...document.querySelectorAll('[id]')].map(n=>n.id);assert(new Set(ids).size===ids.length,platform+': unique DOM IDs');
  assert($('view-render').disabled,platform+': Render disabled before capture');
  const c={id:7,tileset:tiles(platform),width:view.width,height:view.height,start:{frames:1},end:{frames:4},info:{producerPixels:platform==='dos'?100:0},replay:{count:3,complete:true},raster:{lines:[{line:0,changes:2},{line:2,changes:1}],blits:[{index:0,kind:'Copy A',width:32,height:2},{index:1,kind:'Cookie cut',width:32,height:2}]}};
  render.setCapture(c);assert(document.body.dataset.workspace==='render'&&!$('view-render').disabled,platform+': capture opens Render');
  await nextPaint();
  assert(document.documentElement.scrollWidth<=innerWidth,platform+': no horizontal overflow');
  const output=$('render-output');assert(output!==play,platform+': independent output canvas');
  if(view.render==='commands'){
   assert(pixel(output)===original,platform+': final capture initialized');
   document.querySelector('[data-render-nav=first]').click();const first=sent.at(-1);assert(first.type==='seek'&&first.step===0,platform+': First seeks initial state');
   render.result({type:'seek',capture:7,request:first.request,info:{cursor:0,cacheBytes:0},tileset:tiles(platform),pixels:pixels(view.width,view.height),elapsedMs:1});
   assert(pixel(output)==='17,34,51,255',platform+': scrub paints output');
   document.querySelector('[data-render-nav=last]').click();const last=sent.at(-1);assert(last.step===3,platform+': Last seeks final step');
   render.result({type:'seek',capture:7,request:first.request,info:{cursor:0,cacheBytes:0},tileset:tiles(platform),pixels:pixels(view.width,view.height,0xff00ff00),elapsedMs:1});
   assert(pixel(output)==='17,34,51,255',platform+': stale seek ignored');
   $('replay-cancel').click();assert(sent.at(-1).type==='cancel-seek',platform+': cancel request');
   render.result({type:'seek',capture:7,request:last.request,info:{cursor:3,cacheBytes:0},pixels:pixels(view.width,view.height,0xff00ff00),elapsedMs:1});
   assert(pixel(output)==='17,34,51,255',platform+': cancelled response ignored');
   if(platform==='dos')assert($('replay-reveal').checked&&!$('replay-reveal').parentElement.hidden,'DOS reveal survives refactor');
   const r=output.getBoundingClientRect();output.dispatchEvent(new MouseEvent('click',{clientX:r.left+r.width*(.25+1/view.width),clientY:r.top+r.height*(.75+1/view.height),bubbles:true}));
   assert(sent.at(-1).type==='pixel'&&Math.abs(sent.at(-1).x-Math.floor(view.width*.25))<=1&&Math.abs(sent.at(-1).y-Math.floor(view.height*.75))<=1,platform+': aspect-correct output pixel');
  }else{
   const initial=sent.at(-1);assert(initial.type==='raster-seek'&&initial.line===0,platform+': initial line');
   const info={line:0,raster:44,frame:4,complete:true,registers:[],changes:[],pointers:[],palette:Array(32).fill(0),planes:4,mode:platform==='c64'?0:'Indexed color'};
   render.result({type:'raster-seek',capture:7,request:initial.request,info,tileset:tiles(platform),layers:Array.from({length:3},()=>pixels(view.width,view.height))});
   document.querySelector('[data-render-nav=next]').click();assert(sent.at(-1).line===2,platform+': sparse scanline navigation');
   if(view.render==='amiga'){
    $('amiga-blitter').click();const blit=sent.at(-1);assert(blit.type==='blit-seek'&&blit.index===1,'Amiga masked lens');
    render.result({type:'blit-seek',capture:7,request:blit.request,info:{index:1,count:2,width:32,height:2,frame:3,line:80,kind:'Cookie cut',con0:0xfca,con1:0,firstMask:65535,lastMask:65535,pointers:[0,0,0,0],modulos:[0,0,0,0],complete:true},layers:Array.from({length:6},(_,i)=>pixels(i<4?32:640,i<4?2:256))});
    assert(pixel(output)==='17,34,51,255','Amiga uses common output for destination preview');
    $('amiga-filter').value='copy';$('amiga-filter').dispatchEvent(new Event('change'));assert(sent.at(-1).index===0,'Amiga filtered operations');
   }
  }
  if(tilesetBytes[platform]){
   const mode=$('tileset-auxiliary-view');if(mode){mode.value='tiles';mode.dispatchEvent(new Event('change'));}
   assert(!$('tileset-panel').hidden,platform+': tileset shares auxiliary slot');
   const atlas=$('tileset'),r=atlas.getBoundingClientRect();atlas.dispatchEvent(new MouseEvent('click',{clientX:r.left+r.width/32,clientY:r.top+r.height/32,bubbles:true}));
   assert(/\$[0-9a-f]+/.test($('tileset-details').textContent),platform+': tile inspection identifies memory');
   await nextPaint();
  }
  if(innerWidth>1000){const aux=$('render-auxiliary');assert(aux.hidden||aux.scrollHeight<=aux.clientHeight+1,platform+': auxiliary buffers remain fully visible');assert(document.documentElement.scrollHeight<=innerHeight,platform+': desktop workspace fits viewport');}
  assert(pixel(play)===original,platform+': scrub preserves Play pixels');
  $('view-play').click();assert(!$('play-workspace').hidden&&$('render-workspace').hidden,platform+': Play navigation');
  $('view-render').click();assert(!$('render-workspace').hidden,platform+': Render navigation retains capture');
  $('render-resume').click();assert(resumed===1,platform+': shared resume');
  render.reset();assert($('view-render').disabled&&document.body.dataset.workspace==='play',platform+': reset invalidates capture');
  results.push(platform+' passed');$('results').textContent=results.join('\n');
 }
 $('results').textContent='PASS: all sixteen UI adapters, shared timeline, stale/cancelled seeks, pixel coordinates, separate Play pixels and capture lifecycle.\n'+results.join('\n');
 document.title='PASS — Shared emulator UI checks';
}catch(error){$('results').textContent='FAIL: '+error.stack+'\n'+results.join('\n');document.title='FAIL — Shared emulator UI checks';throw error;}
