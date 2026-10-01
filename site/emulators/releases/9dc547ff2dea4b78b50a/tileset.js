import {decodeTileset} from './tileset-decode.js';
const make=(tag,text)=>{const e=document.createElement(tag);if(text)e.textContent=text;return e;};
// Shares the auxiliary slot with layer previews, leaving the output and scrubber
// in place. Color changes are local interpretations of the captured bytes.
export function createTileset({platform,ui,layerPanels=[]}) {
 const c64=platform==='c64',panel=make('div');panel.id='tileset-panel';panel.hidden=layerPanels.length>0;
 ui.auxiliary.append(panel);ui.auxiliary.hidden=false;ui.toolbar.hidden=false;
 const controls=make('div');controls.className='tileset-controls';ui.toolbar.append(controls);
 function select(name,entries,change){const label=make('label',name),input=make('select');input.id='tileset-'+name.toLowerCase().replaceAll(' ','-');for(const [value,text]of entries){const o=make('option',text);o.value=value;input.append(o);}label.append(input);controls.append(label);input.onchange=change;return input;}
 const mode=layerPanels.length?select('Auxiliary view',[['layers','Layers'],['tiles',c64?'Character set':'Tileset']],()=>{const tiles=mode.value==='tiles';panel.hidden=!tiles;for(const p of layerPanels)p.hidden=tiles;for(const c of choices)c.hidden=!tiles;}):null;
 const choices=[],options={};let snapshot=null,decoded=null,selected=-1;
 function option(name,key,entries){const s=select(name,entries,()=>{options[key]=s.value;draw();});choices.push(s.parentElement);return s;}
 if(c64){option('Color from','context',[['auto','First matching screen cell'],['cell','Chosen screen cell']]);const label=make('label','Fallback / chosen cell'),input=make('input');input.id='tileset-cell';input.type='number';input.min=0;input.max=999;input.value=0;input.oninput=()=>{options.cell=input.value;draw();};label.append(input);controls.append(label);choices.push(label);}
 else if(platform==='gb')option('Palette','palette',[[0,'Background (BGP)'],[1,'Objects (OBP0)'],[2,'Objects (OBP1)']]);
 else if(platform==='gg')option('Palette','palette',[[0,'Palette 0'],[1,'Palette 1 / sprites']]);
 else {option('Source','source',[[0,'BG0'],[1,'BG1'],[2,'BG2'],[3,'BG3'],[4,'Objects · 4-bit'],[5,'Objects · 8-bit']]);option('Palette bank','palette',Array.from({length:16},(_,i)=>[i,String(i)]));}
 if(mode)for(const c of choices)c.hidden=true;
 const buffer=ui.buffer(panel,{id:'tileset',title:c64?'Character set':'Tileset',width:128,height:128,aspect:'1/1'});
 const details=make('section');details.id='tileset-details';details.append(make('h3',c64?'Character memory':'Tile memory'));
 const context=make('p'),swatches=make('div'),selection=make('p');swatches.className='tileset-swatches';details.append(context,swatches,selection);ui.sidebar.append(details);
 const selectedText=()=>selection.textContent=decoded&&selected>=0?decoded.describe(Math.min(selected,decoded.count-1)):'Select a character or tile to see its code and memory address.';
 buffer.setInspect((x,y)=>{if(!decoded)return;selected=(y>>3)*decoded.columns+(x>>3);selectedText();details.scrollIntoView({block:'nearest'});});
 function draw(){
  if(!snapshot)return;decoded=decodeTileset(platform,snapshot,options);buffer.labels(decoded.title);buffer.picture.style.aspectRatio=`${decoded.width}/${decoded.height}`;panel.style.setProperty('--tileset-ratio',decoded.width/decoded.height);
  buffer.draw(decoded.pixels,decoded.width,decoded.height);context.textContent=decoded.description;swatches.replaceChildren();
  for(const [i,color]of decoded.colors.entries()){const s=make('span');s.style.backgroundColor=`rgba(${color[0]},${color[1]},${color[2]},${color[3]/255})`;s.title=`${i}: ${color.slice(0,3).join(', ')}${color[3]?'':' (transparent)'}`;swatches.append(s);}selectedText();
 }
 function reset(){snapshot=decoded=null;selected=-1;buffer.clear();context.textContent='Graphics memory at the selected render position.';swatches.replaceChildren();selectedText();}
 function update(data){if(!data){reset();return;}snapshot=data;draw();}
 reset();return {reset,update};
}
