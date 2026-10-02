import assert from 'node:assert/strict';
import fs from 'node:fs';
const {chromium}=await import(process.env.PLAYWRIGHT_MODULE||'playwright');
const [media,checkpoint,out,fields='6']=process.argv.slice(2);
const browser=await chromium.launch({channel:'chrome',headless:true});
try{
 const page=await browser.newPage(),errors=[];page.on('pageerror',e=>errors.push(e.message));page.on('console',m=>{if(m.type()==='error')console.error(m.text());});
 await page.goto('http://127.0.0.1:8790/tools/browser/tests/graphics-live-3ds.html');
 await page.locator('#media').setInputFiles(media);await page.locator('#checkpoint').setInputFiles(checkpoint);
 const result=await page.evaluate(async fields=>{
  const {default:factory}=await import('/tools/platform/n3ds/browser/web/core.js');
  const {createLive3DSGraphics}=await import('/site/emulators/graphics-live-3ds.js');
  const {create3DSExecution}=await import('/site/emulators/execution-3ds.js');
  const core=await factory();let media=new Uint8Array(await document.querySelector('#media').files[0].arrayBuffer());
  const check=ok=>{if(!ok)throw Error(core.UTF8ToString(core._rr_error()));};let p=core._rr_input(media.length);check(p);core.HEAPU8.set(media,p);check(core._rr_init(media.length));media=null;
  const initial=new Uint8Array(await document.querySelector('#checkpoint').files[0].arrayBuffer());
  const backend=await createLive3DSGraphics(core),execution=create3DSExecution({reference:backend.reference,experimental:backend.experimental});
  if(!backend.experimental.available)throw Error(backend.reason);
  const restore=()=>{core.HEAPU8.set(initial,core._rr_state_input(initial.length));check(core._rr_state_load(initial.length));};
  const save=()=>{let n=core._rr_state_save();check(n);return core.HEAPU8.slice(core._rr_state_data(),core._rr_state_data()+n);};
  const status=()=>JSON.parse(core.UTF8ToString(core._rr_status()));
  const hash=async bytes=>Array.from(new Uint8Array(await crypto.subtle.digest('SHA-256',bytes)),x=>x.toString(16).padStart(2,'0')).join('');
  const run=async(mode,delay=0)=>{
   await execution.select('reference');restore();await execution.select(mode);const before=backend.stats(),start=performance.now(),frame=status().frames;
   const original=core.graphicsTransfer;if(delay)core.graphicsTransfer=async packet=>{await new Promise(r=>setTimeout(r,delay));return original(packet);};
   for(let i=0;i<fields;i++){
    core._rr_pad(i%5===0?1:0,(i%3-1)*20,0);core._rr_touch(160,120,+(i%7===0));
    while(status().frames<frame+i+1)check(await core.ccall('rr_run','number',['number'],[10000],{async:true})>=0);
   }
   core.graphicsTransfer=original;const ms=performance.now()-start,end=save(),stats=backend.stats();
   let presentationDifferences=null;
   if(mode==='experimental'){
    const bitmap=backend.present();if(!bitmap)throw Error('GPU scanout presentation unavailable');
    const canvas=new OffscreenCanvas(400,480),ctx=canvas.getContext('2d');ctx.drawImage(bitmap,0,0);bitmap.close();
    const actual=ctx.getImageData(0,0,400,480).data,fp=core._rr_frame();presentationDifferences=0;
    for(let i=0;i<actual.length;i++)presentationDifferences+=actual[i]!==core.HEAPU8[fp+i];
   }
   return {ms,state:await hash(end),bytes:end.length,presentationDifferences,operations:stats.operations.map((n,i)=>n-before.operations[i]),accelerated:stats.accelerated.map((n,i)=>n-before.accelerated[i]),gpuHostMs:stats.hostMs-before.hostMs,allocatedBytes:stats.allocatedBytes};
  };
  try{
   const reference=await run('reference'),experimental=await run('experimental'),delayed=await run('experimental',3);
   await execution.select('reference');return {schema:1,fields,reference,experimental,delayed,mode:execution.snapshot().effective};
  }finally{backend.dispose();}
 },Number(fields));
 console.log(JSON.stringify(result));if(out)fs.writeFileSync(out,JSON.stringify(result,null,2)+'\n');
 assert.deepEqual(errors,[]);assert.equal(result.experimental.state,result.reference.state,'GPU continuation differs');assert.equal(result.delayed.state,result.reference.state,'Host delay changed guest order');assert(result.experimental.accelerated.some(n=>n>0),'No live GPU operation ran');assert.equal(result.mode,'reference');
 assert.equal(result.experimental.presentationDifferences,0,'GPU scanout differs from reference image');assert.equal(result.delayed.presentationDifferences,0);
 result.result='PASS';result.browser=await browser.version();if(out)fs.writeFileSync(out,JSON.stringify(result,null,2)+'\n');console.log(JSON.stringify(result));
}finally{await browser.close();}
