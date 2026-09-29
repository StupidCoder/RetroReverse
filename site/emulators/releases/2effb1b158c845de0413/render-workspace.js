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
  if(platform==='gba')tileset=createTileset({platform,ui});
  return {reset(){replay.reset();inspector.reset();tileset?.reset();vram?.reset();},setCapture(c){inspector.setCapture(c);replay.setCapture(c);tileset?.update(c.tileset);vram?.update(c.vram);},result(m){replay.result(m);inspector.result(m);}};
}

export function createRenderWorkspace({platform,presentation,send,resume,playCanvas}) {
  const $=id=>document.getElementById(id),root=$('render-workspace');
  const views=createWorkspaces({navigation:$('workspace-nav'),onChange:id=>document.body.dataset.workspace=id});
  views.register({id:'play',label:'Play',panel:$('play-workspace')});
  views.register({id:'render',label:'Render',panel:root,enabled:false});
  const output=createBufferView($('render-output-slot'),{id:'render-output',title:'Output buffer',width:presentation.width,height:presentation.height,aspect:presentation.aspect,output:true});
  const ui={root,output,position:$('render-position'),auxiliary:$('render-auxiliary'),toolbar:$('render-toolbar'),sidebar:$('render-details'),timeline:createTimeline($('render-timeline')),
    buffer(parent,options){return createBufferView(parent,{width:presentation.width,height:presentation.height,aspect:presentation.aspect,...options});}};
  const adapter=adapters[presentation.render]({platform,send,ui});
  let capture=null;
  $('render-resume').onclick=resume;
  function reset(){capture=null;adapter.reset();views.enable('render',false);views.select('play');}
  function setCapture(c){capture=c;output.copy(playCanvas);adapter.setCapture(c);views.enable('render',true);views.select('render',{focus:true});}
  function result(m){if(capture&&m.capture===capture.id)adapter.result(m);}
  reset();return {reset,setCapture,result,isInspecting:()=>!!capture};
}
