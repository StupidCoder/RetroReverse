import {createDrivePanel} from './drive-panel.js';
import {createStoragePanel,createStorageSource} from './storage-panel.js';
import {createRenderWorkspace} from './render-workspace.js';
import {localElement,namespacePanel} from './panel-dom.js';
import {createCodeWorkspace} from './code-workspace.js';
import {createStatePanel} from './state-panel.js';
import {createMemoryPanel} from './memory-panel.js';
export function createViewportPanels({views,platform,send,transport,feed,canvas,legacy,warehouse,bindGameInput,recording,presentation,renderTemplate}){
 const storageSource=createStorageSource(platform);
 const renderPanels=new Set();let capture=null,capabilities={};
 const drivePanels=new Set();const codePanels=new Set(),statePanels=new Set(),games=new Set(),borrowed=new Map();let generation=0,knowledge=null,last=null,machine=null;
 const inspect=(at,id)=>{views.reveal('code').inspect?.(at,id);};
 const stateNavigation={code:inspect,memory:(region,offset)=>views.reveal('hex').navigate?.(region,offset)};
 const noViews={register(){},enable(){},select(){},current(){return null;}};
 function borrow(key,root){const node=legacy[key];if(!node){root.textContent='Unavailable for this system.';return {};}
  const previous=borrowed.get(key);if(previous){root.append(node);previous.textContent='This shared capture/control panel moved to another viewport.';}else root.append(node);borrowed.set(key,root);node.hidden=false;
  return {setActive(v){if(v){borrowed.set(key,root);root.replaceChildren(node);node.hidden=false;}},dispose(){if(borrowed.get(key)===root){warehouse.append(node);borrowed.delete(key);}}};
 }
 const api={
  make(kind,id,root){
   if(kind==='game'){const holder=document.createElement('div');holder.className='viewport-game';const c=document.createElement('canvas');c.width=canvas.width;c.height=canvas.height;c.tabIndex=0;c.setAttribute('aria-label','Game output and keyboard controls');holder.append(c);root.append(holder);const unbind=bindGameInput(c);const ratio=presentation.aspect.split('/').map(Number).reduce((a,b)=>a/b);const observer=new ResizeObserver(([entry])=>{const {width,height}=entry.contentRect,w=Math.min(width,height*ratio);c.style.width=w+'px';c.style.height=w/ratio+'px';});observer.observe(holder);c.style.objectFit='fill';const g={canvas:c,active:false};games.add(g);const draw=()=>{if(c.width!==canvas.width)c.width=canvas.width;if(c.height!==canvas.height)c.height=canvas.height;c.getContext('2d').drawImage(canvas,0,0);};draw();return {setActive(v){g.active=v;if(v)draw();},dispose(){games.delete(g);observer.disconnect();unbind?.();}};}
   if(kind==='drive'){const p=createDrivePanel({root,id,send});drivePanels.add(p);if(generation)p.ready(generation);if(machine)p.state(machine);return {...p,dispose(){p.dispose();drivePanels.delete(p);}};}
   if(kind==='storage')return createStoragePanel({root,source:storageSource,platform,send,feed});
   if(['atlas','hex','tape'].includes(kind))return createMemoryPanel({root,kind,send,feed,onCode:['c64','dos','3do'].includes(platform)?inspect:null});
   if(kind.startsWith('render')&&({render:'render-render','render-sources':'render-sources','render-details':'render-details'})[kind]!==id){
    const host=document.createElement('div');host.className='render-embedded';host.dataset.renderView=kind;host.innerHTML=renderTemplate;root.append(host);
    const panel=createRenderWorkspace({platform,presentation,send,resume:()=>transport('run'),playCanvas:canvas,workspaces:noViews,root:host});
    panel.capabilities(capabilities);localElement(host,'cancelcapture').onclick=()=>send('cancel-capture');namespacePanel(host,id);const entry={panel,host,id};renderPanels.add(entry);if(generation)panel.ready(true);if(capture){panel.setCapture(capture);namespacePanel(host,id);}
    return {setActive(){},dispose(){renderPanels.delete(entry);}};
   }
   if(kind==='state'){const p=createStatePanel(stateNavigation);p.active=false;statePanels.add(p);root.append(p.root);p.setKnowledge(knowledge);if(last)p.update(last);return {setActive(v){p.active=v;if(v&&last)p.update(last);},dispose(){statePanels.delete(p);},...stateNavigation};}
   if(kind==='code'){const p=createCodeWorkspace({root,views:noViews,send,transport,platform,idPrefix:'code-'+id,embedded:true,lessons:false,statePanel:{update:s=>api.snapshot(s),setKnowledge(){},setRunning(){},reset(){}}});codePanels.add(p);if(generation){p.ready(generation);if(knowledge)p.result({type:'debug-capabilities',generation,knowledge});}if(machine)p.state(machine);return {...p,dispose(){p.dispose();codePanels.delete(p);}};}
   if(kind==='recording'){const panel=borrow(kind,root),client={active:false,overview:o=>recording.acceptOverview({type:'memory-overview',request:-1,overview:o})},off=feed.subscribe(client);return {setActive(v){panel.setActive(v);client.active=v;feed.visibility();},dispose(){off();panel.dispose();}};}
   return borrow(kind,root);
  },
  mediaReady(files){storageSource.loaded(files[0]);},
  media(files){storageSource.set(files);},
  capabilities(c){capabilities=c;for(const r of renderPanels)r.panel.capabilities(c);},
  ready(g){generation=g;for(const p of drivePanels)p.ready(g);for(const p of codePanels)p.ready(g);},
  state(m){machine=m;for(const p of drivePanels)p.state(m);for(const r of renderPanels)r.panel.ready(!!generation&&!m.running&&!m.capturing&&!m.saving&&!m.debugBusy&&!m.memoryRecording&&!m.experimentOwned);for(const p of codePanels)p.state(m);for(const p of statePanels)p.setRunning(m.running||m.debugBusy||m.capturing||m.memoryRecording);},
  driveResult(m){for(const p of drivePanels)p.result(m);},
  result(m){if(m.type==='debug-capabilities'){knowledge=m.knowledge;for(const p of statePanels)p.setKnowledge(knowledge);}for(const p of codePanels)p.result(m);if(m.snapshot)api.snapshot(m.snapshot);},
  snapshot(s){last=s;for(const p of statePanels)if(p.active)p.update(s);},
  present(){for(const g of games)if(g.active){const c=g.canvas;if(c.width!==canvas.width)c.width=canvas.width;if(c.height!==canvas.height)c.height=canvas.height;c.getContext('2d').drawImage(canvas,0,0);}},
  capture(c){capture=c;for(const r of renderPanels){r.panel.setCapture(c);namespacePanel(r.host,r.id);}},
  renderResult(m){for(const r of renderPanels){r.panel.result(m);namespacePanel(r.host,r.id);}},
  clearCapture(m){capture=null;for(const r of renderPanels)r.panel.reset(m);},
  reset(){generation=0;for(const p of drivePanels)p.reset();capture=null;for(const r of renderPanels)r.panel.reset({unload:true});knowledge=null;last=null;for(const p of codePanels)p.reset();for(const p of statePanels)p.reset();},
  inspect,
 };
 return api;
}
