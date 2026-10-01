import {leaf,split,leaves,replace,close,validLayout,preset} from './viewport-model.js';
// Stable pane instances survive presets, resizing and maximization. Only an
// explicit content replacement/close disposes an instance.
export function createViewportWorkspace({root,navigation,platform,onChange=()=>{}}){
 const registered=new Map(),instances=new Map(),layouts=new Map(),types=new Map();let active='play',tree=preset('play'),maximized=null,mobileId=null,sequence=0,factory=null;
 const mobile=matchMedia('(max-width:700px)');
 const storageKey='rr.viewport.v1.'+platform;let saved={};try{saved=JSON.parse(localStorage.getItem(storageKey)||'{}');}catch{}
 navigation.setAttribute('aria-label','Layout');navigation.onchange=()=>select(navigation.value);
 // Upgrade the former default without discarding customized Play layouts.
 if(saved.play?.id==='play-game'&&saved.play.kind==='game'&&!saved.play.axis)saved.play=preset('play');
 function save(){try{localStorage.setItem(storageKey,JSON.stringify(Object.fromEntries(layouts)));}catch{}}
 function activate(){for(const n of leaves(tree)){const select=instances.get(n.id)?.element.querySelector('.viewport-header select');if(select)for(const o of select.options)o.disabled=['lesson','recording','session'].includes(o.value)&&leaves(tree).some(v=>v.id!==n.id&&v.kind===o.value);}const shown=new Set([...root.querySelectorAll('.viewport')].map(n=>n.dataset.pane));for(const [id,v]of instances)v.panel.setActive?.(shown.has(id));onChange(active);}
 function pane(n){let v=instances.get(n.id);if(v)return v.element;
  const element=document.createElement('section');element.className='viewport';element.dataset.pane=n.id;element.dataset.kind=n.kind;
  const header=document.createElement('header'),select=document.createElement('select'),actions=document.createElement('details'),summary=document.createElement('summary'),menu=document.createElement('div'),body=document.createElement('div');body.className='viewport-body';header.className='viewport-header';select.setAttribute('aria-label','Panel content');
  for(const [id,t]of types){const o=document.createElement('option');o.value=id;o.textContent=t.label;select.append(o);}select.value=n.kind;summary.textContent='Layout';menu.className='viewport-menu';actions.append(summary,menu);header.append(select,actions);element.append(header,body);
  const panel=factory(n.kind,n.id,body);v={element,panel};instances.set(n.id,v);
  const button=(text,fn)=>{const b=document.createElement('button');b.textContent=text;b.onclick=()=>{actions.open=false;fn();};menu.append(b);};
  select.onchange=()=>{const kind=select.value;if(['lesson','recording','session'].includes(kind)&&leaves(tree).some(v=>v.id!==n.id&&v.kind===kind)){select.value=n.kind;return;}v.panel.dispose?.();instances.delete(n.id);tree=replace(tree,n.id,leaf(n.id,kind));mobileId=n.id;maximized=null;commit();};
  button('Maximize / restore',()=>{maximized=maximized?null:n.id;draw();});
  for(const [axis,label]of [['x','Split side by side'],['y','Split above / below']])button(label,()=>{if(leaves(tree).length>=4)return;tree=replace(tree,n.id,split(axis,n,leaf('pane-'+Date.now()+'-'+(++sequence),'hex')));maximized=null;commit();});
  button('Close / merge with neighbor',()=>{if(leaves(tree).length===1)return;tree=close(tree,n.id);v.panel.dispose?.();instances.delete(n.id);maximized=null;commit();});
  return element;
 }
 function build(n){if(!n.axis)return pane(n);const box=document.createElement('div');box.className='viewport-split';box.dataset.axis=n.axis;const a=build(n.a),b=build(n.b),bar=document.createElement('div');bar.className='viewport-divider';bar.tabIndex=0;bar.setAttribute('role','separator');bar.setAttribute('aria-label',n.axis==='x'?'Resize columns':'Resize rows');bar.setAttribute('aria-orientation',n.axis==='x'?'vertical':'horizontal');bar.setAttribute('aria-valuemin','15');bar.setAttribute('aria-valuemax','85');
  const size=()=>{box.style.setProperty('--split',n.ratio*100+'%');bar.setAttribute('aria-valuenow',String(Math.round(n.ratio*100)));};size();
  const move=e=>{const r=box.getBoundingClientRect();n.ratio=Math.max(.15,Math.min(.85,n.axis==='x'?(e.clientX-r.left)/r.width:(e.clientY-r.top)/r.height));size();};
  bar.onpointerdown=e=>{bar.setPointerCapture(e.pointerId);e.preventDefault();};bar.onpointermove=e=>{if(bar.hasPointerCapture(e.pointerId))move(e);};bar.onpointerup=e=>{bar.releasePointerCapture(e.pointerId);save();};
  bar.onkeydown=e=>{const d={ArrowLeft:-.025,ArrowUp:-.025,ArrowRight:.025,ArrowDown:.025}[e.key];if(d){e.preventDefault();n.ratio=Math.max(.15,Math.min(.85,n.ratio+d));size();save();}};box.append(a,bar,b);return box;
 }
 function draw(){if(!factory)return;const parking=document.getElementById('panel-warehouse');for(const v of instances.values())parking.append(v.element);if(mobile.matches){const choices=leaves(tree);if(!choices.some(n=>n.id===mobileId))mobileId=(active==='play'?choices[0]:(choices.find(n=>n.kind!=='game')||choices[0])).id;const picker=document.createElement('select');picker.className='mobile-pane-picker';picker.setAttribute('aria-label','Visible viewport');for(const [i,n]of choices.entries()){const o=document.createElement('option');o.value=n.id;o.textContent=(i+1)+'. '+types.get(n.kind).label;picker.append(o);}picker.value=mobileId;picker.onchange=()=>{mobileId=picker.value;draw();};root.replaceChildren(picker,pane(choices.find(n=>n.id===mobileId)));}else root.replaceChildren(maximized?pane(leaves(tree).find(n=>n.id===maximized)):build(tree));navigation.value=active;activate();}
 function commit(){layouts.set(active,tree);save();draw();}
 function select(id){const v=registered.get(id);if(!v||v.button.disabled)return;active=id;maximized=null;mobileId=null;tree=layouts.get(id)||preset(id);layouts.set(id,tree);draw();}
 mobile.addEventListener('change',draw);
 return {
  register({id,label,panel,enabled=true}){panel.hidden=true;const button=document.createElement('option');button.id='view-'+id;button.value=id;button.textContent=label;button.disabled=!enabled;navigation.append(button);registered.set(id,{button,panel});},
  enable(id,value){const v=registered.get(id);if(v)v.button.disabled=!value;},select,current:()=>active,
  configure(entries,make){for(const [id,label]of entries)types.set(id,{label});factory=make;for(const [id,n]of Object.entries(saved))if(validLayout(n,[...types.keys()]))layouts.set(id,n);tree=layouts.get(active)||preset(active);layouts.set(active,tree);draw();},
  visible(kind){return [...root.querySelectorAll('.viewport')].some(p=>p.dataset.kind===kind);},
  each(fn){for(const v of instances.values())fn(v.panel);},
  reveal(kind){const found=leaves(tree).find(n=>n.kind===kind);if(found){maximized=null;mobileId=found.id;draw();return instances.get(found.id).panel;}const n=leaves(tree).at(-1);instances.get(n.id)?.panel.dispose?.();instances.delete(n.id);tree=replace(tree,n.id,leaf(n.id,kind));mobileId=n.id;maximized=null;commit();return instances.get(n.id).panel;},
  tourLayout(){active='loader';maximized=null;mobileId='loader-lesson';tree=split('x',split('y',leaf('loader-code','code'),leaf('loader-lesson','lesson'),.65),split('y',leaf('loader-ram','atlas'),leaf('loader-tape','tape')),.55);for(const n of leaves(tree)){const v=instances.get(n.id);if(v&&v.element.dataset.kind!==n.kind){v.panel.dispose?.();instances.delete(n.id);}}commit();},
  restorePreset(){for(const n of leaves(tree)){instances.get(n.id)?.panel.dispose?.();instances.delete(n.id);}tree=preset(active);maximized=null;commit();},
  dispose(){mobile.removeEventListener('change',draw);for(const v of instances.values())v.panel.dispose?.();instances.clear();}
 };
}
