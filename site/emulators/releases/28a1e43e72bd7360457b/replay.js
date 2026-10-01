export function createReplay({ui,send,onPosition=()=>{}}) {
  const {output,timeline,position}=ui,canvas=output.canvas;
  let capture=null,cursor=0,count=0,latest=0,shown=null,previous=null;
  const highlight=timeline.toggle('replay-highlight','Highlight changes',{onChange:draw}).input;
  const revealOption=timeline.toggle('replay-reveal','Reveal buffer writes',{hidden:true,onChange:()=>seek(cursor)});
  const reveal=revealOption.input,revealLabel=revealOption.wrapper,note=timeline.note;
  const cancel=timeline.button('replay-cancel','Cancel seek',()=>{latest=0;send('cancel-seek');cancel.hidden=true;note.textContent='Seek cancelled. The paused machine is unchanged.';timeline.value(cursor,`Step ${cursor.toLocaleString()} / ${count.toLocaleString()}`);});
  function draw(){
    if(!shown)return;
    const pixels=new Uint8ClampedArray(shown);
    if(highlight.checked&&previous)for(let i=0;i<pixels.length;i+=4)if(pixels[i]!==previous[i]||pixels[i+1]!==previous[i+1]||pixels[i+2]!==previous[i+2]){pixels[i]=255;pixels[i+1]=195;pixels[i+2]=30;}
    output.draw(pixels);
  }
  function reset(){capture=null;latest=0;shown=previous=null;cancel.hidden=true;}
  function setCapture(c){
    revealLabel.hidden=!c.info.producerPixels;reveal.checked=!!c.info.producerPixels;
    capture=c.id;count=c.replay.count;cursor=count;timeline.range({max:count});timeline.value(cursor,`Step ${count.toLocaleString()} / ${count.toLocaleString()}`);
    shown=canvas.getContext('2d').getImageData(0,0,canvas.width,canvas.height).data;previous=null;
    position.textContent=`Captured display ${c.start.frames}–${c.end.frames}`;
    output.labels('Output buffer','Rendering at the selected step');
    note.textContent=`Final captured display · ${count.toLocaleString()} rendering steps. First shows pre-existing buffer contents. Pixel inspection refers to the final capture.`;
    if(c.replay.surface)note.textContent+=' '+c.replay.surface+'. This preview projects RAM through the final copies; it is not historical monitor output. Reveal buffer writes dims pixels until their first recorded write.';
    if(!c.replay.complete)note.textContent+=' Evidence is incomplete; replay may not reach the final pixels.';
  }
  function seek(step){if(capture===null)return;step=Math.max(0,Math.min(count,step));latest=send('seek',{capture,step,reveal:!revealLabel.hidden&&reveal.checked});cancel.hidden=false;note.textContent=`Seeking to rendering step ${step.toLocaleString()}…`;}
  timeline.onSeek(seek);
  function result(m){
    if(!['seek','seek-progress'].includes(m.type)||m.capture!==capture||m.request!==latest)return;
    if(m.type==='seek-progress'){note.textContent='Replaying recorded rendering effects…';return;}
    cancel.hidden=true;cursor=m.info.cursor;timeline.value(cursor,`Step ${cursor.toLocaleString()} / ${count.toLocaleString()}`);previous=shown;shown=new Uint8ClampedArray(m.pixels);draw();onPosition(m);
    note.textContent=`${cursor?m.info.command?.kind||'Rendering event':'Pre-existing buffer contents'}${m.info.writerPC&&cursor?' · writer PC 0x'+m.info.pc.toString(16).padStart(8,'0'):''}${m.info.surface?' · '+m.info.surface:''} · ${m.elapsedMs.toFixed(1)} ms seek · ${(m.info.cacheBytes/1048576).toFixed(1)} MiB replay cache. Pixel inspection refers to the final capture.`;
  }
  return {reset,setCapture,seek,result};
}
