import {createDCTexture} from './dc-texture.js';
import {createPS1VRAM} from './ps1-vram.js';
import {createTileset} from './tileset.js';
import {createWorkspaces} from './workspaces.js';
import {createBufferView} from './buffer-view.js';
import {createTimeline} from './render-timeline.js';
import {createRaster} from './raster.js';
import {createAmigaRaster} from './amiga-raster.js';
import {createReplay} from './replay.js';
import {createInspector} from './inspector.js';

const adapters={raster:createRaster,amiga:createAmigaRaster,commands:createCommands};
function createCommands({platform,send,ui}) {
  let tileset=null,vram=null;
  const replay=createReplay({send,ui,onPosition:m=>{tileset?.update(m.tileset);vram?.update(m.vram);}});
  const inspector=createInspector({platform,canvas:ui.output.canvas,buffer:ui.output,root:ui.sidebar,send,jump:replay.seek});
  if(platform==='ps1')vram=createPS1VRAM({ui});
  if(platform==='dc')vram=createDCTexture({ui,seek:replay.seek});
  if(platform==='gba')tileset=createTileset({platform,ui});
  return {reset(){replay.reset();inspector.reset();tileset?.reset();vram?.reset();},setCapture(c){inspector.setCapture(c);replay.setCapture(c);tileset?.update(c.tileset);vram?.update(c.vram);},result(m){replay.result(m);inspector.result(m);}};
}

export function createRenderWorkspace({platform,presentation,send,resume,playCanvas,workspaces,beforeCapture=()=>{}}) {
  const $=id=>document.getElementById(id),root=$('render-workspace');
  const views=workspaces||createWorkspaces({navigation:$('workspace-nav'),onChange:id=>document.body.dataset.workspace=id});
  if(!workspaces)views.register({id:'play',label:'Play',panel:$('play-workspace')});
  views.register({id:'render',label:'Render',panel:root,enabled:false,onSelect:open});
  const output=createBufferView($('render-output-slot'),{id:'render-output',title:'Output buffer',width:presentation.width,height:presentation.height,aspect:presentation.aspect,output:true});
  const ui={root,output,position:$('render-position'),auxiliary:$('render-auxiliary'),toolbar:$('render-toolbar'),sidebar:$('render-details'),timeline:createTimeline($('render-timeline')),
    buffer(parent,options){return createBufferView(parent,{width:presentation.width,height:presentation.height,aspect:presentation.aspect,...options});}};
  const adapter=adapters[presentation.render]({platform,send,ui});
  let capture=null,pending=false,available=false;
  $('render-resume').onclick=resume;
  function ready(value=true){available=value;views.enable('render',value);}
  function open(){if(!available||capture||pending)return;pending=true;beforeCapture();ui.position.textContent='Capturing the next complete display interval…';$('cancelcapture').hidden=false;send('capture-render');}
  function reset({starting=false,unload=false}={}){capture=null;adapter.reset();if(starting)return;pending=false;if(unload)available=false;if(views.current()==='render')views.select('play');views.enable('render',available);}
  function progress(text){ui.position.textContent=text;}

  function setCapture(c){const requested=pending;pending=false;capture=c;output.copy(playCanvas);adapter.setCapture(c);views.enable('render',true);if(!requested)views.select('render',{focus:true});}
  function result(m){if(capture&&m.capture===capture.id)adapter.result(m);}
  reset();return {reset,ready,progress,setCapture,result,isInspecting:()=>!!capture};
}
