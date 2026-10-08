// Test-only suffix appended by Playwright to the production worker. Execution,
// pacing, input application, GPU transport and presentation remain production
// functions. No cartridge, checkpoint or pixel data belongs in public reports.
let demoInitial=null,demoTrial=null;
const demoMessage=onmessage,demoPaint=paint;
const demoHash=async bytes=>Array.from(new Uint8Array(await crypto.subtle.digest('SHA-256',bytes)),v=>v.toString(16).padStart(2,'0')).join('');
const demoPixels=()=>{const p=core._rr_frame();return core.HEAPU8.slice(p,p+400*480*4);};
paint=function(){
 demoPaint();const t=demoTrial;if(!t)return;
 const frame=status().frames;if(frame===t.lastFrame)return;
 if(frame!==t.lastFrame+1)throw Error('Benchmark skipped a display boundary');
 const now=performance.now();t.times.push(now-t.lastTime);t.lastTime=now;t.lastFrame=frame;
 const elapsed=frame-t.startFrame;
 if(elapsed%t.sampleEvery===0||elapsed===t.fields){
  const start=performance.now();t.samples.push({interval:elapsed,status:status(),pixels:demoPixels()});t.observerMs+=performance.now()-start;
 }
 if(elapsed%30===0)send('demo-progress',{interval:elapsed,fields:t.fields,ms:now-t.start});
 if(elapsed>=t.fields){t.end=performance.now();running=false;}
};
onmessage=async event=>{
 const m=event.data;if(!m.type.startsWith('demo-'))return demoMessage(event);
 let ownership=null;
 try{
  if(!loaded||platform!=='3ds'||running||executionGate.owner)throw Error('Benchmark requires a paused loaded 3DS machine');
  if(m.type==='demo-initial'){demoInitial=new Uint8Array(m.bytes);send('demo-ready');return;}
  if(m.type!=='demo-run')throw Error('Unknown benchmark operation');
  if(!demoInitial||!Number.isInteger(m.fields)||m.fields<1||m.fields>10000)throw Error('Invalid benchmark fixture');
  ownership=executionGate.acquire('demo-benchmark');
  await execution3DS.select('reference');
  const at=core._rr_state_input(demoInitial.length);check(at);core.HEAPU8.set(demoInitial,at);check(core._rr_state_load(demoInitial.length));
  // Idle input is deliberate: button pulses dismiss the attract/demo sequence.
  inputQueue.length=0;appliedKeys.clear();inputs=new InputQueue(60);lastButtons=-1;lastX=lastY=0;inputSequence=lastInputStep=0;
  core._rr_pad(0,0,0);core._rr_touch(0,0,0);turbo=false;
  await execution3DS.select(m.mode);graphics3DS.measureTiming(!!m.timestamps);
  if(m.diagnostics&&!core._rr_perf_enable)throw Error('Core has no diagnostic API');
  if(core._rr_perf_enable)check(core._rr_perf_enable(m.diagnostics||0));
  const beforeShaders=core.picaShaders?.snapshot();
  const beforeProfile=json('_rr_profile'),beforeGPU=graphics3DS.stats(),initialStatus=status();
  const transfer=core.graphicsTransfer,batchTransfer=core.graphicsBatch,traffic={batchSizes:{},sparseDraws:0,workgroups:0,fullWorkgroups:0,deviceCopyBytes:0,specializedMaterials:0,specializedLighting:0,operations:0,submissions:0,batches:0,inputBytes:0,beforeBytes:0,outputBytes:0,kinds:{},refusals:{}};
  const record=async(packets,batch)=>{
   traffic.submissions++;if(batch){traffic.batches++;traffic.batchSizes[packets.length]=(traffic.batchSizes[packets.length]||0)+1;}
   for(const packet of packets){traffic.operations++;traffic.fullWorkgroups+=Math.ceil(packet.before.length/(packet.kind>=7?512:256));traffic.inputBytes+=packet.input.length;traffic.kinds[packet.kind]=(traffic.kinds[packet.kind]||0)+1;}
   traffic.beforeBytes+=packets[0].before.length;
   const result=await(batch?batchTransfer(packets):transfer(packets[0]));
   if(result.supported){for(const [i,c] of (batch?result.counts:[result]).entries()){traffic.sparseDraws+=Number(!!c.sparse);traffic.workgroups+=c.workgroups??Math.ceil(packets[i].before.length/(packets[i].kind>=7?512:256));traffic.deviceCopyBytes+=c.sparse&&!c.inPlace?packets[i].before.length:0;}traffic.specializedMaterials+=batch?result.counts.filter(c=>c.materialSpecialized).length:Number(!!result.materialSpecialized);traffic.specializedLighting+=batch?result.counts.filter(c=>c.specialized).length:Number(!!result.specialized);traffic.outputBytes+=result.bytes.length;}
   else traffic.refusals[result.reason]=(traffic.refusals[result.reason]||0)+1;return result;
  };
  core.graphicsTransfer=packet=>record([packet],false);core.graphicsBatch=packets=>record(packets,true);
  const t={fields:m.fields,sampleEvery:Math.max(1,Math.floor(m.fields/4)),times:[],samples:[{interval:0,status:initialStatus,pixels:demoPixels()}],observerMs:0,startFrame:initialStatus.frames,lastFrame:initialStatus.frames};
  maxCall=runMs=paintMs=0;t.start=t.lastTime=performance.now();demoTrial=t;running=true;
  try{await pump(++epoch);}finally{running=false;demoTrial=null;core.graphicsTransfer=transfer;core.graphicsBatch=batchTransfer;}
  if(t.times.length!==m.fields)throw Error('Benchmark ended before the requested boundary');
  const elapsedMs=t.end-t.start,afterGPU=graphics3DS.stats(),diagnostics=m.diagnostics?json('_rr_perf_stats'):null;
  if(core._rr_perf_enable)check(core._rr_perf_enable(0));
  const state=coreState(),stateHash=await demoHash(state);
  const samples=[];
  for(const sample of t.samples){
   const hash=await demoHash(sample.pixels);const canvas=new OffscreenCanvas(400,480);canvas.getContext('2d').putImageData(new ImageData(new Uint8ClampedArray(sample.pixels.buffer),400,480),0,0);
   const png=new Uint8Array(await(await canvas.convertToBlob({type:'image/png'})).arrayBuffer());
   samples.push({interval:sample.interval,status:sample.status,hash,png});
  }
  const diff=(a,b)=>Object.fromEntries(Object.entries(a).filter(([,v])=>typeof v==='number').map(([k,v])=>[k,v-(b[k]||0)]));
  send('demo-result',{result:{mode:m.mode,fields:m.fields,diagnosticStride:m.diagnostics||0,timestamps:!!m.timestamps,elapsedMs,times:t.times,observerMs:t.observerMs,maxCallMs:maxCall,runMs,paintMs,heapBytes:core.HEAPU8.length,
   shaders:core.picaShaders?diff(core.picaShaders.snapshot(),beforeShaders):null,shaderCache:core.picaShaders?.snapshot(),start:initialStatus,end:status(),stateHash,stateBytes:state.length,samples,diagnostics,traffic,
   gpu:{materialCompilation:afterGPU.materialCompilation,lightingCompilation:afterGPU.lightingCompilation,allocatedBytes:afterGPU.allocatedBytes,timing:diff(afterGPU.timing,beforeGPU.timing),accelerated:afterGPU.accelerated.map((v,i)=>v-beforeGPU.accelerated[i])},
   profile:Object.fromEntries(json('_rr_profile').buckets.map(b=>[b.name,b.ms-(beforeProfile.buckets.find(a=>a.name===b.name)?.ms||0)]))}});
 }catch(e){demoTrial=null;running=false;send('demo-error',{text:String(e)});}
 finally{if(ownership)executionGate.release(ownership);}
};
