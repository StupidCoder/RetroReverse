import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import os from 'node:os';
import {execFileSync} from 'node:child_process';
import {createHash} from 'node:crypto';
const {chromium}=await import(process.env.PLAYWRIGHT_MODULE||'playwright');
const [media,checkpoint,out,fieldsArg='300',trialsArg='5']=process.argv.slice(2);
assert(media&&checkpoint&&out,'Usage: 3ds-demo-browser.mjs MEDIA RAW_CHECKPOINT REPORT [FIELDS] [TRIALS]');
const fields=Number(fieldsArg),trials=Number(trialsArg);assert(Number.isInteger(fields)&&fields>0&&fields<=10000);assert(Number.isInteger(trials)&&trials>0&&trials<=10);
const sha=bytes=>createHash('sha256').update(bytes).digest('hex');
const privateDir=process.env.PRIVATE_OUT;assert(privateDir,'PRIVATE_OUT is required for private scene verification images');fs.mkdirSync(privateDir,{recursive:true});
const root=process.cwd(),worker=fs.readFileSync('site/emulators/worker.js','utf8')+'\n'+fs.readFileSync('tools/browser/tests/3ds-demo-worker.js','utf8');
const builds=(process.env.BUILDS||'shipped:site/emulators/cores/3ds,candidate:tools/platform/n3ds/browser/web').split(',').map(v=>{const [name,dir]=v.split(':');return {name,dir:path.resolve(dir)};});
const browser=await chromium.launch({channel:'chrome',headless:process.env.HEADLESS==='1'}),pages=[];
const report={schema:1,method:'Production worker pump, idle input, both-screen presentation; fixed display boundaries; end-state serialization and PNG encoding excluded; periodic pixel-copy observer cost included and reported.',checkpointSha256:sha(fs.readFileSync(checkpoint)),fields,trials,browser:await browser.version(),headless:process.env.HEADLESS==='1',builds:[],measurements:[]};
report.release=JSON.parse(fs.readFileSync('site/emulators/release.json')).id;
report.workerSha256=sha(worker);report.rasterSha256=sha(fs.readFileSync('site/emulators/raster-3ds.js'));report.transportSha256=sha(fs.readFileSync('site/emulators/graphics-live-3ds.js'));
report.host={platform:os.platform(),architecture:os.arch(),cpus:os.cpus().length,cpu:os.cpus()[0]?.model,memoryBytes:os.totalmem()};
if(os.platform()==='darwin'){report.host.model=execFileSync('sysctl',['-n','hw.model'],{encoding:'utf8'}).trim();report.host.acPower=execFileSync('pmset',['-g','batt'],{encoding:'utf8'}).includes("'AC Power'");report.host.lowPowerModes=[...execFileSync('pmset',['-g','custom'],{encoding:'utf8'}).matchAll(/lowpowermode\s+(\d+)/g)].map(m=>Number(m[1]));}
function summarize(r){
 const sorted=[...r.times].sort((a,b)=>a-b),sum=r.times.reduce((a,b)=>a+b,0);let worst60=0;
 for(let i=0;i+60<=r.times.length;i++)worst60=Math.max(worst60,r.times.slice(i,i+60).reduce((a,b)=>a+b,0));
 const middle=Math.floor(sorted.length/2),median=sorted.length%2?sorted[middle]:(sorted[middle-1]+sorted[middle])/2;
 return {intervalsPerSecond:r.fields*1000/r.elapsedMs,nominalPercent:r.fields*1000/r.elapsedMs/60*100,meanMs:r.elapsedMs/r.fields,medianMs:median,p95Ms:sorted[Math.ceil(sorted.length*.95)-1],worst60MeanMs:worst60?worst60/60:null,unattributedMs:r.elapsedMs-Object.values(r.profile).reduce((a,b)=>a+b,0),intervalSumMs:sum};
}
try{
 for(const build of builds){
  const context=await browser.newContext({viewport:{width:1400,height:1000}}),page=await context.newPage(),errors=[];
  page.setDefaultTimeout(120000);
  page.on('pageerror',e=>errors.push(e.message));
  await page.route('**/site/emulators/worker.js',route=>route.fulfill({contentType:'text/javascript',body:worker}));
  for(const asset of ['core.js','core.wasm'])await page.route('**/site/emulators/cores/3ds/'+asset,route=>route.fulfill({contentType:asset.endsWith('.wasm')?'application/wasm':'text/javascript',body:fs.readFileSync(path.join(build.dir,asset))}));
  const manifest=JSON.parse(fs.readFileSync('site/emulators/build-manifest.json'));manifest['3ds/core.wasm']=sha(fs.readFileSync(path.join(build.dir,'core.wasm')));
  await page.route('**/site/emulators/build-manifest.json',route=>route.fulfill({contentType:'application/json',body:JSON.stringify(manifest)}));
  await page.exposeFunction('demoProgress',p=>console.log(JSON.stringify({build:build.name,...p})));
  await page.addInitScript(()=>{const Original=Worker;window.demo={};window.Worker=class extends Original{constructor(...args){super(...args);window.demo.worker=this;this.addEventListener('message',({data:m})=>{window.demo.session=m.session;if(m.type==='demo-progress')window.demoProgress(m);if(m.type==='demo-ready')window.demo.ready=true;if(m.type==='demo-result')window.demo.result=m.result;if(m.type==='demo-error'||m.type==='error')window.demo.error=m.text;});}};});
  await page.goto('http://127.0.0.1:8790/tools/browser/tests/3ds-acceleration.html');
  await page.locator('#open-media').click();await page.locator('#files').setInputFiles(media);await page.locator('#load').click();
  await page.waitForFunction(()=>!document.getElementById('run').disabled||window.demo.error);assert.equal(await page.evaluate(()=>window.demo.error??null),null);if(await page.locator('#media-dialog').isVisible())await page.locator('#close-media').click();
  await page.evaluate(()=>{const i=document.createElement('input');i.type='file';i.id='demo-checkpoint';document.body.append(i);});await page.locator('#demo-checkpoint').setInputFiles(checkpoint);
  await page.evaluate(async()=>{const bytes=await document.getElementById('demo-checkpoint').files[0].arrayBuffer();window.demo.worker.postMessage({type:'demo-initial',session:window.demo.session,bytes},[bytes]);});await page.waitForFunction(()=>window.demo.ready);
  const adapter=await page.evaluate(async()=>{const a=await navigator.gpu.requestAdapter();return a?{vendor:a.info.vendor,architecture:a.info.architecture,device:a.info.device,description:a.info.description}:null;});
  report.builds.push({name:build.name,wasmSha256:sha(fs.readFileSync(path.join(build.dir,'core.wasm'))),adapter});pages.push({page,errors,build});
 }
 const run=async (item,label,count,diagnostics=0,timestamps=false)=>{
  const {page,build,errors}=item;await page.bringToFront();
  if(await page.locator('#performance-panel').evaluate(e=>e.open)!==timestamps)await page.locator('#performance-panel > summary').click();
  await page.evaluate(({fields,diagnostics,timestamps})=>{window.demo.result=null;window.demo.error=null;window.demo.worker.postMessage({type:'demo-run',session:window.demo.session,mode:'experimental',fields,diagnostics,timestamps});},{fields:count,diagnostics,timestamps});
  await page.waitForFunction(()=>window.demo.result||window.demo.error,null,{timeout:1800000});
  const error=await page.evaluate(()=>window.demo.error);assert.equal(error,null);assert.deepEqual(errors,[]);
  const r=await page.evaluate(()=>{const r=window.demo.result;for(const s of r.samples){let text='';for(let i=0;i<s.png.length;i+=8192)text+=String.fromCharCode(...s.png.subarray(i,i+8192));s.png=btoa(text);}return r;});
  for(const sample of r.samples){fs.writeFileSync(path.join(privateDir,`${build.name}-${label}-${sample.interval}.png`),Buffer.from(sample.png,'base64'));delete sample.png;}
  assert.equal(r.end.frames-r.start.frames,count);assert(r.gpu.accelerated[3]>0,'No accelerated raster work');assert.deepEqual(r.traffic.refusals,{});
  if(diagnostics){assert.equal(r.diagnostics.programOverflow,0);assert.equal(r.diagnostics.fallbackOverflow,0);}
  return {build:build.name,label,...r,...summarize(r)};
 };
 for(const item of pages)report.measurements.push(await run(item,'cold',Math.min(30,fields)));
 for(let trial=0;trial<trials;trial++)for(const item of trial%2?[...pages].reverse():pages){
  const r=await run(item,`trial-${trial+1}`,fields);report.measurements.push(r);fs.writeFileSync(out,JSON.stringify(report,null,2)+'\n');
 }
 if(process.env.DIAGNOSTICS==='1'){
  const candidate=pages.at(-1);report.measurements.push(await run(candidate,'timing',fields,0,true));report.measurements.push(await run(candidate,'diagnostics',fields,16,true));
 }
 const comparable=report.measurements.filter(r=>r.fields===fields&&r.label!=='cold');
 for(const r of comparable){assert.equal(r.stateHash,comparable[0].stateHash,'Canonical continuation differs');assert.deepEqual(r.samples.map(s=>({interval:s.interval,hash:s.hash,status:s.status})),comparable[0].samples.map(s=>({interval:s.interval,hash:s.hash,status:s.status})),'Scene samples differ');}
 report.result='PASS';fs.writeFileSync(out,JSON.stringify(report,null,2)+'\n');console.log(JSON.stringify({result:report.result,runs:report.measurements.map(r=>({build:r.build,label:r.label,meanMs:r.meanMs,p95Ms:r.p95Ms,nominalPercent:r.nominalPercent}))}));
}finally{await browser.close();}
