// Adapters supply the unit and seek operation; keyboard, range and navigation
// behavior are identical for scanlines, GPU commands and blits.
export function createTimeline(root) {
  root.className='render-timeline';
  root.innerHTML=`<div class="render-transport"><strong id="render-cursor"></strong><button data-render-nav="first">First</button><button data-render-nav="previous">Previous</button><button data-render-nav="next">Next</button><button data-render-nav="last">Last</button><div class="render-options"></div></div><div class="render-track"><label class="sr-only" for="render-slider">Render position</label><input id="render-slider" type="range" min="0" max="0" value="0" step="1"><div id="render-markers" aria-label="Rendering events"></div></div><p id="render-note" aria-live="polite"></p>`;
  const slider=root.querySelector('input'),label=root.querySelector('strong'),note=root.querySelector('#render-note'),markers=root.querySelector('#render-markers'),options=root.querySelector('.render-options'),buttons=[...root.querySelectorAll('[data-render-nav]')];
  let current=0,min=0,max=0,stops=null,onSeek=()=>{};
  function choose(value){value=Math.max(min,Math.min(max,value));return stops?.length?stops.reduce((best,n)=>Math.abs(n-value)<Math.abs(best-value)?n:best,stops[0]):value;}
  function request(value){if(!slider.disabled)onSeek(choose(value));}
  slider.oninput=()=>request(Number(slider.value));
  buttons.forEach(button=>button.onclick=()=>{
    const action=button.dataset.renderNav;
    if(action==='first')request(stops?.[0]??min);
    else if(action==='last')request(stops?.at(-1)??max);
    else {const delta=action==='next'?1:-1;request(stops?.length?stops[Math.max(0,Math.min(stops.length-1,stops.indexOf(current)+delta))]:current+delta);}
  });
  function range({min:low=0,max:high=0,positions=null,disabled=false}={}){min=low;max=high;stops=positions;slider.min=min;slider.max=max;slider.disabled=disabled||positions?.length===0;buttons.forEach(b=>b.disabled=slider.disabled);}
  function value(n,text){current=n;slider.value=n;label.textContent=text;}
  function toggle(id,text,{checked=false,hidden=false,onChange=()=>{}}={}) {
    const wrapper=document.createElement('label'),input=document.createElement('input'),caption=document.createElement('span');input.type='checkbox';input.id=id;input.checked=checked;caption.textContent=text;wrapper.hidden=hidden;wrapper.append(input,caption);options.append(wrapper);input.onchange=()=>onChange(input.checked);
    return {input,wrapper,caption};
  }
  function button(id,text,onClick){const b=document.createElement('button');b.id=id;b.textContent=text;b.onclick=onClick;options.append(b);return b;}
  function setMarkers(entries){markers.replaceChildren();for(const entry of entries){const b=document.createElement('button');b.type='button';b.style.left=(max===min?0:(entry.value-min)/(max-min)*100)+'%';b.title=entry.label;b.setAttribute('aria-label',entry.label);b.onclick=()=>request(entry.value);markers.append(b);}}
  return {slider,label,note,markers,range,value,toggle,button,setMarkers,onSeek(fn){onSeek=fn;}};
}
