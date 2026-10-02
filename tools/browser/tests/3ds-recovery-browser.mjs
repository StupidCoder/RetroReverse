import assert from 'node:assert/strict';
import fs from 'node:fs';
const {chromium}=await import(process.env.PLAYWRIGHT_MODULE||'playwright');
const [media,checkpoint,out]=process.argv.slice(2),browser=await chromium.launch({channel:'chrome',headless:true});
try{
 const page=await browser.newPage(),errors=[];page.on('pageerror',e=>errors.push(e.message));
 await page.goto('http://127.0.0.1:8790/tools/browser/tests/graphics-live-3ds.html');await page.locator('#media').setInputFiles(media);await page.locator('#checkpoint').setInputFiles(checkpoint);
 const result=await page.evaluate(async()=>{
  const {default:factory}=await import('/tools/platform/n3ds/browser/web/core.js');
  const {create3DSGraphics}=await import('/site/emulators/graphics-3ds.js');
  const {createLive3DSGraphics}=await import('/site/emulators/graphics-live-3ds.js');
  const {create3DSExecution}=await import('/site/emulators/execution-3ds.js');
  const core=await factory(),check=(ok,why)=>{if(!ok)throw Error(why||core.UTF8ToString(core._rr_error()));};
  let bytes=new Uint8Array(await document.querySelector('#media').files[0].arrayBuffer());const upload=core._rr_input(bytes.length);core.HEAPU8.set(bytes,upload);check(core._rr_init(bytes.length));bytes=null;
  const initial=new Uint8Array(await document.querySelector('#checkpoint').files[0].arrayBuffer()),trials=[];
  const hash=async bytes=>Array.from(new Uint8Array(await crypto.subtle.digest('SHA-256',bytes)),x=>x.toString(16).padStart(2,'0')).join('');
  for(const fault of ['reference','device-loss','timeout','delayed']){
   let calls=0,injected=false;
   const backend=await createLive3DSGraphics(core,{timeoutMs:fault==='timeout'?1000:5000,createGPU:async options=>{
    const gpu=await create3DSGraphics(options),execute=gpu.execute.bind(gpu);
    gpu.execute=packet=>{
     calls++;
     if(calls===4&&fault==='timeout'){injected=true;return new Promise(()=>{});}
     const pending=execute(packet);
     if(calls===4&&fault==='device-loss'){injected=true;gpu.destroy();}
     return fault==='delayed'&&calls>1?pending.then(async r=>{await new Promise(resolve=>setTimeout(resolve,3));return r;}):pending;
    };return gpu;
   }});
   const execution=create3DSExecution({reference:backend.reference,experimental:backend.experimental});
   try{
    const at=core._rr_state_input(initial.length);core.HEAPU8.set(initial,at);check(core._rr_state_load(initial.length));
    if(fault!=='reference')await execution.select('experimental');
    const frame=JSON.parse(core.UTF8ToString(core._rr_status())).frames;
    for(let i=0;i<6;i++){
     core._rr_pad(i%2,20,0);core._rr_touch(160,120,i%2);
     while(JSON.parse(core.UTF8ToString(core._rr_status())).frames<frame+i+1){
      check(await core.ccall('rr_run','number',['number'],[10000],{async:true})>=0);
      if(backend.failed)await execution.select('reference');
     }
    }
    const n=core._rr_state_save(),p=core._rr_state_data(),state=await hash(core.HEAPU8.slice(p,p+n));
    const fp=core._rr_frame(),pixels=await hash(core.HEAPU8.slice(fp,fp+400*480*4));
    trials.push({fault,state,pixels,injected,failed:backend.failed,reason:backend.reason,mode:execution.snapshot().effective,stats:backend.stats()});
   }finally{backend.dispose();}
  }
  return {schema:1,trials};
 });
 assert.deepEqual(errors,[]);const baseline=result.trials[0];
 for(const t of result.trials){assert.equal(t.state,baseline.state);assert.equal(t.pixels,baseline.pixels);if(['device-loss','timeout'].includes(t.fault)){assert(t.injected&&t.failed);assert.equal(t.mode,'reference');assert(t.stats.accelerated.some(n=>n>0));}}
 result.result='PASS';result.browser=await browser.version();fs.writeFileSync(out,JSON.stringify(result,null,2)+'\n');console.log(JSON.stringify({result:'PASS',faults:result.trials.map(t=>t.fault)}));
}finally{await browser.close();}
