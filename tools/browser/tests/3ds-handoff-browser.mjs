// Private cartridges/checkpoints stay local. Reports contain hashes and counts.
import assert from 'node:assert/strict';
import fs from 'node:fs';
const {chromium}=await import(process.env.PLAYWRIGHT_MODULE||'playwright');
const [media,checkpoint,out,fields='520']=process.argv.slice(2);
const browser=await chromium.launch({channel:'chrome',headless:true});
try{
 const page=await browser.newPage(),errors=[];page.on('pageerror',e=>errors.push(e.message));
 await page.exposeFunction('progress',message=>console.log(JSON.stringify(message)));
 await page.goto('http://127.0.0.1:8790/tools/browser/tests/graphics-live-3ds.html');
 await page.locator('#media').setInputFiles(media);await page.locator('#checkpoint').setInputFiles(checkpoint);
 const result=await page.evaluate(async fields=>{
  const {default:factory}=await import('/tools/platform/n3ds/browser/web/core.js');
  const {createLive3DSGraphics}=await import('/site/emulators/graphics-live-3ds.js');
  const {create3DSExecution}=await import('/site/emulators/execution-3ds.js');
  const core=await factory();let media=new Uint8Array(await document.querySelector('#media').files[0].arrayBuffer());
  const check=(ok,why)=>{if(!ok)throw Error(why||core.UTF8ToString(core._rr_error()));};let p=core._rr_input(media.length);check(p);core.HEAPU8.set(media,p);check(core._rr_init(media.length));media=null;
  const initial=new Uint8Array(await document.querySelector('#checkpoint').files[0].arrayBuffer());
  const backend=await createLive3DSGraphics(core),execution=create3DSExecution({reference:backend.reference,experimental:backend.experimental});check(backend.experimental.available,backend.reason);
  const restore=bytes=>{const at=core._rr_state_input(bytes.length);core.HEAPU8.set(bytes,at);check(core._rr_state_load(bytes.length));};
  const save=()=>{let n=core._rr_state_save();check(n);return core.HEAPU8.slice(core._rr_state_data(),core._rr_state_data()+n);};
  const json=fn=>JSON.parse(core.UTF8ToString(core[fn]())),status=()=>json('_rr_status');
  const hash=async bytes=>Array.from(new Uint8Array(await crypto.subtle.digest('SHA-256',bytes)),x=>x.toString(16).padStart(2,'0')).join('');
  const frame=()=>{const p=core._rr_frame();return core.HEAPU8.slice(p,p+400*480*4);};
  const trials=[];
  try{for(const mixed of [false,true]){
   await execution.select('reference');restore(initial);let switches=0,capturing=false;
   const checkpoints=[],captures=[],memory=[],allocations=[],startFrame=status().frames,start=performance.now(),statsBefore=backend.stats();
   const select=async mode=>{const previous=execution.snapshot().effective;await execution.select(mode);if(execution.snapshot().effective!==previous)switches++;};
   for(let i=0;i<fields;i++){
    if(mixed&&!capturing&&i%2===0)await select(i%4===0?'experimental':'reference');
    if(i%100===40){await select('reference');check(core._rr_capture_begin());capturing=true;}
    core._rr_pad(i%13===0?512:0,(i%7-3)*8,(i%5-2)*8);core._rr_touch(i%8,i%8,+(i%17===0));
    while(status().frames<startFrame+i+1)check(await core.ccall('rr_run','number',['number'],[10000],{async:true})>=0);
    if(i%100===42){
     check(core._rr_capture_end());capturing=false;
     const before=await hash(save()),pixels=frame(),info=json('_rr_capture_info'),replay=json('_rr_replay_begin');check(replay.complete,'Incomplete capture replay');
     while(!core._rr_replay_seek(replay.count)){}const rp=core._rr_replay_frame();check(pixels.every((v,j)=>v===core.HEAPU8[rp+j]),'Capture replay differs');
     const evidence=[];for(const [x,y] of [[100,100],[160,120],[160,360]]){const q=JSON.parse(core.UTF8ToString(core._rr_pixel(x,y)));check(!q.error&&q.complete!==false,'Missing pixel evidence');evidence.push(q);}
     while(!core._rr_replay_seek(0)){}check(before===await hash(save()),'Replay or pixel inspection mutated guest state');
     captures.push({interval:i+1,pixels:await hash(pixels),evidence:await hash(new TextEncoder().encode(JSON.stringify(evidence))),writes:info.writes});
    }
    if(i%50===49||i===fields-1){
     const state=save();checkpoints.push({interval:i+1,state:await hash(state),bytes:state.length,status:status()});
     // Memory inspection must be non-invasive. Force the reference handoff,
     // then restore a state and resume the selected engine at the next interval.
     const mode=execution.snapshot().effective;await select('reference');
     const regions=json('_rr_inspect_regions');memory.push({interval:i+1,regions:regions.regions?.length});
     check(await hash(save())===checkpoints.at(-1).state,'Memory inspection mutated state');
     restore(state);check(await hash(save())===checkpoints.at(-1).state,'State round trip differs');
     if(mixed)await select(mode);
     allocations.push({interval:i+1,wasm:core.HEAPU8.length,gpu:backend.stats().allocatedBytes,accelerated:backend.stats().accelerated.map((n,i)=>n-statsBefore.accelerated[i])});
     await window.progress({mixed,interval:i+1,fields,switches,seconds:(performance.now()-start)/1000});
    }
   }
   const stats=backend.stats();trials.push({mixed,fields,switches,ms:performance.now()-start,checkpoints,captures,memory,allocations,accelerated:stats.accelerated.map((n,i)=>n-statsBefore.accelerated[i])});
  }
  return {schema:1,fields,trials};
  }finally{backend.dispose();}
 },Number(fields));
 assert.deepEqual(errors,[]);const [reference,mixed]=result.trials;
 assert.deepEqual(mixed.checkpoints,reference.checkpoints,'Switched continuation differs');assert.deepEqual(mixed.captures,reference.captures,'Capture pixels or histories differ');assert(mixed.switches>=Number(fields)/2);assert(mixed.accelerated.some(n=>n>0));
 assert(mixed.allocations.every(a=>a.gpu<=32*1024*1024),'GPU scratch exceeds fixture budget');
 for(let i=1;i<mixed.allocations.length;i++)assert(mixed.allocations[i].accelerated[3]>mixed.allocations[i-1].accelerated[3],'Scene stopped producing accelerated draws');
 result.result='PASS';result.browser=await browser.version();fs.writeFileSync(out,JSON.stringify(result,null,2)+'\n');console.log(JSON.stringify({result:'PASS',fields:result.fields,switches:mixed.switches,captures:mixed.captures.length}));
}finally{await browser.close();}
