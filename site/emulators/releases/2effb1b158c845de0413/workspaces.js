// Views share a machine session, but own their presentation and navigation.
// Registering a memory workspace later does not require changing the player.
export function createWorkspaces({navigation, onChange=()=>{}}) {
  const views=new Map();let active=null;
  navigation.setAttribute('role','tablist');navigation.setAttribute('aria-label','Workspace');
  function select(id,{focus=false}={}) {
    const next=views.get(id);if(!next||next.button.disabled)return;
    active=id;
    for(const [key,v] of views){const chosen=key===id;v.panel.hidden=!chosen;v.button.setAttribute('aria-selected',String(chosen));v.button.tabIndex=chosen?0:-1;}
    if(focus)next.button.focus();onChange(id);
  }
  function register({id,label,panel,enabled=true}) {
    const button=document.createElement('button');button.type='button';button.id='view-'+id;
    button.textContent=label;button.setAttribute('role','tab');button.setAttribute('aria-controls',panel.id);
    button.disabled=!enabled;button.tabIndex=-1;button.setAttribute('aria-selected','false');
    panel.setAttribute('role','tabpanel');panel.setAttribute('aria-labelledby',button.id);panel.hidden=true;
    button.onclick=()=>select(id);views.set(id,{button,panel});navigation.append(button);
    if(active===null&&enabled)select(id);
  }
  function enable(id,value){const v=views.get(id);if(!v)return;v.button.disabled=!value;if(!value&&active===id){const fallback=[...views].find(([,v])=>!v.button.disabled);if(fallback)select(fallback[0]);}}
  navigation.addEventListener('keydown',e=>{
    if(!['ArrowLeft','ArrowRight','Home','End'].includes(e.key))return;
    const ids=[...views].filter(([,v])=>!v.button.disabled).map(([id])=>id);if(!ids.length)return;
    const i=ids.indexOf(active),n=e.key==='Home'?0:e.key==='End'?ids.length-1:(i+(e.key==='ArrowRight'?1:-1)+ids.length)%ids.length;
    e.preventDefault();select(ids[n],{focus:true});
  });
  return {register,enable,select,current:()=>active};
}
