// Reparent existing panels instead of recreating them: disclosure state, focus
// and selection belong to the panel, never to an execution owner.
export function createPanelLayout(panel){
 const hosts=new Map();let active=null,visible=true,placement='after';
 function mount(){if(!active)return;const {root,controls}=hosts.get(active);root.hidden=!visible;controls.querySelector('input').checked=visible;controls.querySelector('select').value=placement;if(visible)root.append(panel);const parent=controls.parentElement;if(placement==='before')parent.insertBefore(controls,parent.firstChild);else parent.append(controls);controls.after(root);}
 return {
  register(id,parent){
   const controls=document.createElement('div');controls.className='code-toolbar';
   const label=document.createElement('label'),input=document.createElement('input');input.type='checkbox';input.checked=visible;label.append(input,' Game state');
   const select=document.createElement('select');select.setAttribute('aria-label','Game state panel position');for(const [value,text] of [['after','Below workspace'],['before','Above workspace']]){const o=document.createElement('option');o.value=value;o.textContent=text;select.append(o);}
   const root=document.createElement('section');root.className='state-panel-host';parent.append(controls,root);controls.append(label,select);
   input.onchange=()=>{visible=input.checked;mount();};select.onchange=()=>{placement=select.value;mount();};hosts.set(id,{root,controls});
  },
  suggest(show){visible=show;mount();},
  activate(id){if(!hosts.has(id))return;active=id;mount();},
 };
}
