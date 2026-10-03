import assert from 'node:assert/strict';
import fs from 'node:fs';
const {chromium}=await import(process.env.PLAYWRIGHT_MODULE||'playwright');
const [media,checkpoint,out,fields='6',trials='1',delay='3']=process.argv.slice(2);
const browser=await chromium.launch({channel:'chrome',headless:true});
try{
 const page=await browser.newPage(),errors=[];page.on('pageerror',e=>errors.push(e.message));page.on('console',m=>{if(m.type()==='error')console.error(m.text());});
 await page.goto('http://127.0.0.1:8790/tools/browser/tests/graphics-live-3ds.html');
 await page.locator('#media').setInputFiles(media);await page.locator('#checkpoint').setInputFiles(checkpoint);
 const result=await page.evaluate(async ({fields,trials,delay,coreBase,graphicsBase,measureGPU,idleInput})=>{
  const {default:factory}=await import(coreBase+'/core.js');
  const {createLive3DSGraphics}=await import(graphicsBase+'/graphics-live-3ds.js');
  const {create3DSExecution}=await import(graphicsBase+'/execution-3ds.js');
  const core=await factory();let media=new Uint8Array(await document.querySelector('#media').files[0].arrayBuffer());
  const check=ok=>{if(!ok)throw Error(core.UTF8ToString(core._rr_error()));};let p=core._rr_input(media.length);check(p);core.HEAPU8.set(media,p);check(core._rr_init(media.length));media=null;
  const initial=new Uint8Array(await document.querySelector('#checkpoint').files[0].arrayBuffer());
  const backend=await createLive3DSGraphics(core,{measureGPU}),execution=create3DSExecution({reference:backend.reference,experimental:backend.experimental});
  if(!backend.experimental.available)throw Error(backend.reason);
  const restore=()=>{const at=core._rr_state_input(initial.length);core.HEAPU8.set(initial,at);check(core._rr_state_load(initial.length));};
  const save=()=>{let n=core._rr_state_save();check(n);return core.HEAPU8.slice(core._rr_state_data(),core._rr_state_data()+n);};
  const status=()=>JSON.parse(core.UTF8ToString(core._rr_status()));
  const profile=()=>Object.fromEntries(JSON.parse(core.UTF8ToString(core._rr_profile())).buckets.map(b=>[b.name,b.ms]));
  const hash=async bytes=>Array.from(new Uint8Array(await crypto.subtle.digest('SHA-256',bytes)),x=>x.toString(16).padStart(2,'0')).join('');
  const presentationCanvas=new OffscreenCanvas(400,480),presentationContext=presentationCanvas.getContext('2d');
  const run=async(mode,delay=0,count=fields)=>{
   await execution.select('reference');restore();await execution.select(mode);const before=backend.stats(),profileBefore=profile(),start=performance.now(),frame=status().frames;
   const submittedKinds={},supportedKinds={},unsupported={},inputBytes={};
   const original=core.graphicsTransfer;core.graphicsTransfer=async packet=>{submittedKinds[packet.kind]=(submittedKinds[packet.kind]||0)+1;inputBytes[packet.kind]=(inputBytes[packet.kind]||0)+packet.input.length;if(delay)await new Promise(r=>setTimeout(r,delay));const result=await original(packet);if(result.supported)supportedKinds[packet.kind]=(supportedKinds[packet.kind]||0)+1;else unsupported[result.reason]=(unsupported[result.reason]||0)+1;return result;};
   for(let i=0;i<count;i++){
    core._rr_pad(idleInput?0:i%5===0?1:0,idleInput?0:(i%3-1)*20,0);core._rr_touch(160,120,idleInput?0:+(i%7===0));
    while(status().frames<frame+i+1)check(await core.ccall('rr_run','number',['number'],[10000],{async:true})>=0);
    if(mode==='experimental'){const image=backend.present();if(!image)throw Error('GPU presentation unavailable');presentationContext.drawImage(image,0,0);image.close();}
    else {const at=core._rr_frame();presentationContext.putImageData(new ImageData(new Uint8ClampedArray(core.HEAPU8.slice(at,at+400*480*4)),400,480),0,0);}
   }
   core.graphicsTransfer=original;const ms=performance.now()-start,end=save(),stats=backend.stats(),profileMs=Object.fromEntries(Object.entries(profile()).map(([name,ms])=>[name,ms-(profileBefore[name]||0)]));
   let presentationDifferences=null;
   if(mode==='experimental'){
    const bitmap=backend.present();if(!bitmap)throw Error('GPU scanout presentation unavailable');
    const canvas=new OffscreenCanvas(400,480),ctx=canvas.getContext('2d');ctx.drawImage(bitmap,0,0);bitmap.close();
    const actual=ctx.getImageData(0,0,400,480).data,fp=core._rr_frame();presentationDifferences=0;
    for(let i=0;i<actual.length;i++)presentationDifferences+=actual[i]!==core.HEAPU8[fp+i];
   }
   return {ms,submittedKinds,supportedKinds,unsupported,inputBytes,state:await hash(end),bytes:end.length,presentationDifferences,operations:stats.operations.map((n,i)=>n-before.operations[i]),accelerated:stats.accelerated.map((n,i)=>n-before.accelerated[i]),gpuHostMs:stats.hostMs-before.hostMs,timestamps:stats.timestamps,timing:stats.timing?Object.fromEntries(Object.entries(stats.timing).map(([key,n])=>[key,n-(before.timing?.[key]||0)])):null,allocatedBytes:stats.allocatedBytes,profileMs};
  };
  try{
   await run('reference',0,2);await run('experimental',0,2);
   const measurements=[];let reference,experimental;
   for(let i=0;i<trials;i++){
    // Alternate order after warming both engines to avoid a cold-reference bias.
    if(i%2){experimental=await run('experimental');reference=await run('reference');}
    else {reference=await run('reference');experimental=await run('experimental');}
    if(reference.state!==experimental.state)throw Error('Trial state mismatch');
    measurements.push({reference,experimental,speedup:reference.ms/experimental.ms});
   }
   const delayed=delay?await run('experimental',delay):null;
   await execution.select('reference');return {schema:1,fields,measureGPU,idleInput,coreBase,graphicsBase,includesPresentation:true,measurements,reference,experimental,delayed,mode:execution.snapshot().effective};
  }finally{backend.dispose();}
 },{fields:Number(fields),trials:Number(trials),delay:Number(delay),coreBase:process.env.CORE_BASE||'/tools/platform/n3ds/browser/web',graphicsBase:process.env.GRAPHICS_BASE||'/site/emulators',measureGPU:process.env.MEASURE_GPU==='1',idleInput:process.env.IDLE_INPUT==='1'});
 console.log(JSON.stringify(result));if(out)fs.writeFileSync(out,JSON.stringify(result,null,2)+'\n');
 assert.deepEqual(errors,[]);assert.equal(result.experimental.state,result.reference.state,'GPU continuation differs');if(result.delayed)assert.equal(result.delayed.state,result.reference.state,'Host delay changed guest order');assert(result.experimental.accelerated.some(n=>n>0),'No live GPU operation ran');assert.equal(result.mode,'reference');assert(result.experimental.supportedKinds[Number(process.env.RASTER_KIND||6)]>0,'No GPU coverage/interpolation/sampling ran');assert.deepEqual(result.experimental.unsupported,{},'GPU packets unexpectedly fell back');
 assert.equal(result.experimental.presentationDifferences,0,'GPU scanout differs from reference image');if(result.delayed)assert.equal(result.delayed.presentationDifferences,0);
 for(const {reference,experimental} of result.measurements){
  assert.equal(reference.profileMs['WebGPU upload / wait / readback']||0,0);
  assert.equal(reference.profileMs['PICA coverage / sampling / GPU inputs']||0,0);
  assert(experimental.profileMs['WebGPU upload / wait / readback']>0,'GPU wait is missing from the profile');
  assert(Object.values(experimental.profileMs).every(ms=>ms>=0),'Exclusive profiling became negative across Asyncify');
  if(experimental.accelerated[3]>0)assert(experimental.profileMs['PICA coverage / sampling / GPU inputs']>0,'Hybrid preparation is missing from the profile');
 }
 result.result='PASS';result.browser=await browser.version();if(out)fs.writeFileSync(out,JSON.stringify(result,null,2)+'\n');console.log(JSON.stringify(result));
}finally{await browser.close();}
