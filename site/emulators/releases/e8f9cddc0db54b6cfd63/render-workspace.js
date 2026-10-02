import {localElement} from './panel-dom.js';
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

export function createRenderWorkspace({platform,presentation,send,resume,playCanvas,workspaces,root:panelRoot,beforeCapture=()=>{}}) {
  const root=panelRoot||document.getElementById('render-workspace'),$=id=>localElement(root,id)||document.getElementById(id);
  const views=workspaces||createWorkspaces({navigation:$('workspace-nav'),onChange:id=>document.body.dataset.workspace=id});
  if(!workspaces)views.register({id:'play',label:'Play',panel:$('play-workspace')});
  views.register({id:'render',label:'Render',panel:root,enabled:false});
  const output=createBufferView($('render-output-slot'),{id:'render-output',title:'Output buffer',width:presentation.width,height:presentation.height,aspect:presentation.aspect,output:true});
  const ui={root,output,position:$('render-position'),auxiliary:$('render-auxiliary'),toolbar:$('render-toolbar'),sidebar:$('render-details'),timeline:createTimeline($('render-timeline')),
    buffer(parent,options){return createBufferView(parent,{width:presentation.width,height:presentation.height,aspect:presentation.aspect,...options});}};
  const adapter=adapters[presentation.render]({platform,send,ui});
  let capture=null,pending=false,available=false,supported=true;
  $('render-resume').onclick=resume;
  $('render-capture').onclick=open;
  function ready(value=true){value=!!value&&supported;available=value;$('render-capture').disabled=!value||pending;views.enable('render',value);}
  function open(){if(!available||pending)return;pending=true;$('render-capture').disabled=true;beforeCapture();ui.position.textContent='Capturing the next complete display interval…';$('cancelcapture').hidden=false;send('capture-render');}
  function reset({starting=false,unload=false}={}){capture=null;adapter.reset();if(starting)return;pending=false;$('render-capture').disabled=!available;if(unload)available=false;views.enable('render',available);}
  function progress(text){ui.position.textContent=text;}

  function setCapture(c){const requested=pending;pending=false;$('render-capture').disabled=!available;capture=c;output.copy(playCanvas);adapter.setCapture(c);views.enable('render',true);if(!requested)views.select('render',{focus:true});}
  function result(m){if(capture&&m.capture===capture.id)adapter.result(m);}
  function capabilities(c){const wasSupported=supported;supported=c.renderCapture!==false;if(!supported){ready(false);ui.position.textContent='Rendering capture is unavailable in this development core.';}else if(!wasSupported)ui.position.textContent='Capture next display to inspect rendering.';}
  reset();return {reset,ready,capabilities,progress,setCapture,result,isInspecting:()=>!!capture};
}
