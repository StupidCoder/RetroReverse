// Private cartridges/checkpoints stay local. Reports contain hashes and counts.
import assert from 'node:assert/strict';
import fs from 'node:fs';
const {chromium}=await import(process.env.PLAYWRIGHT_MODULE||'playwright');
const [media,checkpoint,out,fields='520']=process.argv.slice(2);
assert(['','reject','counter','length'].includes(process.env.BATCH_FAULT||''),'Unknown batch fault');
assert(['full','none'].includes(process.env.CAPTURE_MODE||'full'),'Unknown capture mode');
const browser=await chromium.launch({channel:'chrome',headless:true});
try{
 const page=await browser.newPage(),errors=[];page.on('pageerror',e=>errors.push(e.message));
 await page.exposeFunction('progress',message=>console.log(JSON.stringify(message)));
 await page.goto('http://127.0.0.1:8790/tools/browser/tests/graphics-live-3ds.html');
 await page.locator('#media').setInputFiles(media);await page.locator('#checkpoint').setInputFiles(checkpoint);
 const result=await page.evaluate(async ({fields,captureMode,requiredKind,batchFault})=>{
  const {default:factory}=await import('/tools/platform/n3ds/browser/web/core.js');
  const {createLive3DSGraphics}=await import('/site/emulators/graphics-live-3ds.js');
  const {create3DSExecution}=await import('/site/emulators/execution-3ds.js');
  const {createPicaShaders}=await import('/site/emulators/shader-3ds.js');
  const core=await factory({picaShaders:createPicaShaders()});let media=new Uint8Array(await document.querySelector('#media').files[0].arrayBuffer());
  const check=(ok,why)=>{if(!ok)throw Error(why||core.UTF8ToString(core._rr_error()));};let p=core._rr_input(media.length);check(p);core.HEAPU8.set(media,p);check(core._rr_init(media.length));media=null;
  const initial=new Uint8Array(await document.querySelector('#checkpoint').files[0].arrayBuffer());
  const backend=await createLive3DSGraphics(core),execution=create3DSExecution({reference:backend.reference,experimental:backend.experimental});check(backend.experimental.available,backend.reason);
  const supportedKinds={},transfer=core.graphicsTransfer,batchTransfer=core.graphicsBatch;core.graphicsTransfer=async packet=>{const result=await transfer(packet);if(result.supported)supportedKinds[packet.kind]=(supportedKinds[packet.kind]||0)+1;return result;};
  let injectedBatchFaults=0;
  core.graphicsBatch=async packets=>{const result=await batchTransfer(packets);
   if(result.supported&&packets.length>1&&batchFault&&!injectedBatchFaults++){
    if(batchFault==='reject')return {supported:false,reason:'Injected batch rejection'};
    if(batchFault==='counter')return {...result,counts:result.counts.map((c,i)=>i===result.counts.length-1?{...c,drawn:-1}:c)};
    if(batchFault==='length')return {...result,bytes:result.bytes.slice(1)};
   }
   if(result.supported)for(const packet of packets)supportedKinds[packet.kind]=(supportedKinds[packet.kind]||0)+1;return result;};
  const restore=bytes=>{const at=core._rr_state_input(bytes.length);core.HEAPU8.set(bytes,at);check(core._rr_state_load(bytes.length));};
  const save=()=>{let n=core._rr_state_save();check(n);return core.HEAPU8.slice(core._rr_state_data(),core._rr_state_data()+n);};
  const json=fn=>JSON.parse(core.UTF8ToString(core[fn]())),status=()=>json('_rr_status');
  const hash=async bytes=>Array.from(new Uint8Array(await crypto.subtle.digest('SHA-256',bytes)),x=>x.toString(16).padStart(2,'0')).join('');
  const frame=()=>{const p=core._rr_frame();return core.HEAPU8.slice(p,p+400*480*4);};
  const trials=[];
  try{for(const mixed of [false,true]){
   await execution.select('reference');restore(initial);const shaderBefore=core.picaShaders.snapshot();let switches=0,capturing=false;
   const checkpoints=[],captures=[],memory=[],allocations=[],startFrame=status().frames,start=performance.now(),statsBefore=backend.stats();
   const kindsBefore={...supportedKinds},kindCounts=()=>Object.fromEntries(Object.entries(supportedKinds).map(([kind,n])=>[kind,n-(kindsBefore[kind]||0)]));
   const select=async mode=>{const previous=execution.snapshot().effective;await execution.select(mode);if(execution.snapshot().effective!==previous)switches++;};
   for(let i=0;i<fields;i++){
    if(mixed&&!capturing&&i%2===0)await select(i%4===0?'experimental':'reference');
    if(captureMode==='full'&&i%100===40){await select('reference');check(core._rr_capture_begin());capturing=true;}
    core._rr_pad(i%13===0?512:0,(i%7-3)*8,(i%5-2)*8);core._rr_touch(i%8,i%8,+(i%17===0));
    while(status().frames<startFrame+i+1)check(await core.ccall('rr_run','number',['number'],[10000],{async:true})>=0);
    if(captureMode==='full'&&i%100===42){
     check(core._rr_capture_end());capturing=false;
     const before=await hash(save()),pixels=frame(),info=json('_rr_capture_info'),replay=json('_rr_replay_begin');check(replay.complete,'Incomplete capture replay: '+JSON.stringify(info));
     while(!core._rr_replay_seek(replay.count)){}const rp=core._rr_replay_frame(),replayPixels=await hash(core.HEAPU8.slice(rp,rp+400*480*4));if(replay.complete)check(pixels.every((v,j)=>v===core.HEAPU8[rp+j]),'Capture replay differs');
     const evidence=[];for(const [x,y] of [[100,100],[160,120],[160,360]]){const q=JSON.parse(core.UTF8ToString(core._rr_pixel(x,y)));check(!q.error&&q.complete===replay.complete&&q.overflow===info.overflow,'Missing or incorrectly labelled pixel evidence');evidence.push(q);}
     while(!core._rr_replay_seek(0)){}check(before===await hash(save()),'Replay or pixel inspection mutated guest state');
     captures.push({interval:i+1,complete:replay.complete,overflow:info.overflow,pixels:await hash(pixels),replayPixels,evidence:await hash(new TextEncoder().encode(JSON.stringify(evidence))),writes:info.writes,events:info.events});
    }
    if(i%50===49||i===fields-1){
     const state=save();checkpoints.push({interval:i+1,state:await hash(state),pixels:await hash(frame()),bytes:state.length,status:status()});
     // Memory inspection must be non-invasive. Force the reference handoff,
     // then restore a state and resume the selected engine at the next interval.
     const mode=execution.snapshot().effective;await select('reference');
     const regions=json('_rr_inspect_regions');memory.push({interval:i+1,regions:regions.regions?.length});
     check(await hash(save())===checkpoints.at(-1).state,'Memory inspection mutated state');
     restore(state);check(await hash(save())===checkpoints.at(-1).state,'State round trip differs');
     if(mixed)await select(mode);
     allocations.push({interval:i+1,wasm:core.HEAPU8.length,gpu:backend.stats().allocatedBytes,accelerated:backend.stats().accelerated.map((n,i)=>n-statsBefore.accelerated[i]),supportedKinds:kindCounts()});
     await window.progress({mixed,interval:i+1,fields,switches,seconds:(performance.now()-start)/1000});
    }
   }
   const stats=backend.stats();trials.push({shaders:Object.fromEntries(Object.entries(core.picaShaders.snapshot()).map(([k,v])=>[k,v-shaderBefore[k]])),mixed,fields,switches,ms:performance.now()-start,checkpoints,captures,memory,allocations,supportedKinds:kindCounts(),accelerated:stats.accelerated.map((n,i)=>n-statsBefore.accelerated[i])});
  }
  return {schema:1,fields,captureMode,requiredKind,batchFault,injectedBatchFaults:Math.min(1,injectedBatchFaults),trials};
  }finally{core.picaShaders.dispose();backend.dispose();}
 },{fields:Number(fields),captureMode:process.env.CAPTURE_MODE||'full',requiredKind:Number(process.env.RASTER_KIND||6),batchFault:process.env.BATCH_FAULT||''});
 assert.deepEqual(errors,[]);if(result.batchFault)assert.equal(result.injectedBatchFaults,1);const [reference,mixed]=result.trials;
 assert.deepEqual(mixed.checkpoints,reference.checkpoints,'Switched continuation differs');assert.deepEqual(mixed.captures,reference.captures,'Capture pixels or histories differ');assert(mixed.switches>=Number(fields)/2);assert(mixed.accelerated.some(n=>n>0));
 assert(mixed.allocations.every(a=>a.gpu<=32*1024*1024),'GPU scratch exceeds fixture budget');
 for(let i=1;i<mixed.allocations.length;i++)assert(mixed.allocations[i].accelerated[3]>mixed.allocations[i-1].accelerated[3],'Scene stopped producing accelerated draws');
 assert(mixed.supportedKinds[result.requiredKind]>0,'Required GPU draw kind did not run');
 assert(mixed.shaders.vertices>0,'Compiled vertex path did not run');assert.equal(reference.shaders.vertices,0,'Reference executed compiled shaders');
 result.result='PASS';result.browser=await browser.version();fs.writeFileSync(out,JSON.stringify(result,null,2)+'\n');console.log(JSON.stringify({result:'PASS',fields:result.fields,switches:mixed.switches,captures:mixed.captures.length}));
}finally{await browser.close();}
