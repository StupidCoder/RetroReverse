export function createReplay({canvas,send}) {
 const root=document.createElement('section');root.className='replay';root.hidden=true;root.setAttribute('aria-label','Rendering replay');
 root.innerHTML=`<div class="replay-buttons"><strong>Rendering</strong><button data-nav="first">First</button><button data-nav="previous">Previous</button><button data-nav="next">Next</button><button data-nav="last">Last</button><label><input type="checkbox" id="replay-highlight"> Highlight changes</label><button id="replay-cancel" hidden>Cancel seek</button></div><label class="replay-slider">Rendering progress <input type="range" min="0" value="0" step="1" id="replay-position"></label><p id="replay-note" aria-live="polite"></p>`;
 canvas.closest('.monitor').after(root);
 const slider=root.querySelector('#replay-position'),note=root.querySelector('#replay-note'),cancel=root.querySelector('#replay-cancel'),highlight=root.querySelector('#replay-highlight');
 let capture=null,cursor=0,count=0,latest=0,shown=null,previous=null;
 const draw=()=>{if(!shown)return;const pixels=new Uint8ClampedArray(shown);if(highlight.checked&&previous)for(let i=0;i<pixels.length;i+=4)if(pixels[i]!==previous[i]||pixels[i+1]!==previous[i+1]||pixels[i+2]!==previous[i+2]){pixels[i]=255;pixels[i+1]=195;pixels[i+2]=30;}canvas.getContext('2d').putImageData(new ImageData(pixels,canvas.width,canvas.height),0,0);};
 function reset(){capture=null;latest=0;shown=previous=null;root.hidden=true;cancel.hidden=true;}
 function setCapture(c){capture=c.id;count=c.replay.count;cursor=count;slider.max=count;slider.value=cursor;root.hidden=false;shown=canvas.getContext('2d').getImageData(0,0,canvas.width,canvas.height).data;previous=null;note.textContent=`Final captured display · ${count.toLocaleString()} rendering steps. First shows pre-existing buffer contents. Pixel inspection refers to the final capture.`;if(!c.replay.complete)note.textContent+=' Evidence is incomplete; replay may not reach the final pixels.';}
 function seek(step){if(capture===null)return;step=Math.max(0,Math.min(count,step));latest=send('seek',{capture,step});cancel.hidden=false;note.textContent=`Seeking to rendering step ${step.toLocaleString()}…`;}
 root.querySelectorAll('[data-nav]').forEach(b=>b.onclick=()=>seek({first:0,previous:cursor-1,next:cursor+1,last:count}[b.dataset.nav]));
 slider.oninput=()=>seek(Number(slider.value));
 highlight.onchange=draw;
 cancel.onclick=()=>{latest=0;send('cancel-seek');cancel.hidden=true;note.textContent='Seek cancelled. The paused machine is unchanged.';slider.value=cursor;};
 function result(m){if(m.capture!==capture||m.request!==latest)return;if(m.type==='seek-progress'){note.textContent='Replaying recorded rendering effects…';return;}cancel.hidden=true;cursor=m.info.cursor;slider.value=cursor;previous=shown;shown=new Uint8ClampedArray(m.pixels);draw();note.textContent=`Step ${cursor.toLocaleString()} / ${count.toLocaleString()} · ${cursor?m.info.command?.kind||'Rendering event':'Pre-existing buffer contents'} · ${m.elapsedMs.toFixed(1)} ms seek · ${(m.info.cacheBytes/1048576).toFixed(1)} MiB replay cache. Pixel inspection refers to the final capture.`;}
 return {reset,setCapture,seek,result};
}
