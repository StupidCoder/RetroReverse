import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import {createHash} from 'node:crypto';
import {packState,unpackState} from '../../../site/emulators/state.js';
import {InputQueue} from '../../../site/emulators/input.js';
const {chromium}=await import(process.env.PLAYWRIGHT_MODULE||'playwright');
const [media,checkpoint,out]=process.argv.slice(2),dir=path.dirname(out);
const hash=b=>createHash('sha256').update(b).digest('hex');
const mediaHash=createHash('sha256');for await(const chunk of fs.createReadStream(media))mediaHash.update(chunk);
const manifest=JSON.parse(fs.readFileSync('site/emulators/build-manifest.json'));
const graphics=['graphics-3ds.js','graphics-live-3ds.js','presentation-3ds.js','fragment-3ds.js'].map(n=>hash(fs.readFileSync('site/emulators/'+n)));
const identity=hash(JSON.stringify({wasm:manifest['3ds/core.wasm'],graphics}));
const input={...new InputQueue(60),pulses:[],down:[],pending:[],appliedKeys:[],lastButtons:-1,lastX:0,lastY:0,inputSequence:0,lastInputStep:0};
const statePath=path.join(dir,'3ds-handoff-ui-initial.rrstate'),savedPath=path.join(dir,'3ds-handoff-ui-saved.rrstate');
fs.writeFileSync(statePath,await packState({format:1,platform:'3ds',media:[{name:path.basename(media),size:fs.statSync(media).size,sha256:mediaHash.digest('hex')}],firmware:[],core:identity,configuration:{compatibility:true,customFirmware:false},input},fs.readFileSync(checkpoint)));
const browser=await chromium.launch({channel:'chrome',headless:true});
try{
 const page=await browser.newPage({viewport:{width:1400,height:1000}}),errors=[];page.on('pageerror',e=>errors.push(e.message));
 await page.addInitScript(()=>{const Original=window.Worker;window.handoff={messages:[]};window.Worker=class extends Original{constructor(...args){super(...args);window.handoff.worker=this;this.addEventListener('message',({data:m})=>{const h=window.handoff;h.session=m.session;if(m.type==='state')h.state={running:m.running,capturing:m.capturing,execution:m.execution,frames:m.state.frames};if(m.type==='capture')h.capture={id:m.id,replay:m.replay,frameHash:m.frameHash};if(m.type==='memory-overview')h.memory=m.request;if(m.type==='capture-progress')h.captureRunning=true;if(m.type==='capture-cleared')h.captureRunning=false;if(m.type==='pixel')h.pixel=m.evidence;h.messages.push({type:m.type,text:m.text});if(h.messages.length>300)h.messages.shift();});}};});
 await page.goto('http://127.0.0.1:8790/tools/browser/tests/3ds-acceleration.html');
 assert.equal(await page.locator('#execution-mode').inputValue(),'reference');
 await page.locator('#open-media').click();await page.locator('#files').setInputFiles(media);await page.locator('#statefile').setInputFiles(statePath);
 await page.waitForFunction(()=>!document.getElementById('run').disabled,{timeout:120000});
 // State loading need not close the media dialog; use its normal close control.
 if(await page.locator('#media-dialog').isVisible())await page.locator('#close-media').click();
 await page.locator('#execution-mode').selectOption('experimental');
 await page.waitForFunction(()=>window.handoff.state?.execution.effective==='experimental');
 await page.locator('#run').click();await page.waitForFunction(()=>window.handoff.state?.running&&window.handoff.state.execution.graphics.accelerated[3]>0,{timeout:60000});
 await page.locator('#pause').click();await page.waitForFunction(()=>!window.handoff.state?.running);
 await page.locator('#workspace-nav').selectOption('render');await page.locator('#render-capture').click();
 await page.waitForFunction(()=>!!window.handoff.capture,{timeout:120000});
 assert.equal(await page.locator('#execution-mode').inputValue(),'experimental');
 assert.match(await page.locator('#execution-note').textContent(),/Reference inspection; Resume uses Experimental/);
 assert.equal(await page.evaluate(()=>window.handoff.state.execution.effective),'reference');
 assert.equal(await page.evaluate(()=>window.handoff.capture.replay.complete),true);
 await page.locator('#inspect-pixel').click();await page.waitForFunction(()=>window.handoff.pixel?.complete===true);
 await page.screenshot({path:path.join(dir,'3ds-handoff-render.png'),fullPage:true});
 await page.locator('#render-resume').click();await page.waitForFunction(()=>window.handoff.state?.running&&window.handoff.state.execution.effective==='experimental');
 await page.locator('#pause').click();await page.waitForFunction(()=>!window.handoff.state.running);
 // Explicit worker requests exercise the same gate used by memory panels and
 // rapid user actions without depending on the current viewport arrangement.
 await page.evaluate(()=>{const h=window.handoff;h.worker.postMessage({type:'memory-snapshot',session:h.session,request:9901});});
 await page.waitForFunction(()=>window.handoff.memory===9901,{timeout:60000});
 assert.equal(await page.evaluate(()=>window.handoff.state.execution.effective),'reference');
 assert.equal(await page.evaluate(()=>window.handoff.state.execution.preferred),'experimental');
 await page.locator('#workspace-nav').selectOption('play');
 const download=page.waitForEvent('download');await page.locator('#save').click();await(await download).saveAs(savedPath);
 const stored=fs.readFileSync(savedPath),decoded=await unpackState({size:stored.length,arrayBuffer:async()=>stored.buffer.slice(stored.byteOffset,stored.byteOffset+stored.length)});assert.equal(decoded.meta.playMode,'experimental');
 await page.locator('#statefile').setInputFiles(savedPath);
 await page.waitForFunction(()=>!document.getElementById('run').disabled&&window.handoff.state?.execution.effective==='experimental',{timeout:120000});
 // A Pause queued behind a switch must suppress its automatic run continuation.
 await page.locator('#run').click();await page.waitForFunction(()=>window.handoff.state.running);
 await page.evaluate(()=>{const h=window.handoff;h.worker.postMessage({type:'execution-mode',mode:'reference',session:h.session});h.worker.postMessage({type:'pause',session:h.session});});
 await page.waitForFunction(()=>!window.handoff.state.running&&!window.handoff.state.execution.busy);
 const paused=await page.evaluate(()=>window.handoff.state.frames);await page.waitForTimeout(300);assert.equal(await page.evaluate(()=>window.handoff.state.frames),paused);
 await page.locator('#execution-mode').selectOption('reference');await page.waitForFunction(()=>window.handoff.state.execution.preferred==='reference');
 // Cancel an actual reference capture, then ensure it releases execution ownership.
 await page.evaluate(()=>{const h=window.handoff;h.worker.postMessage({type:'capture-render',session:h.session});});
 await page.waitForFunction(()=>window.handoff.captureRunning===true);
 await page.evaluate(()=>{const h=window.handoff;h.worker.postMessage({type:'cancel-capture',session:h.session});});
 await page.waitForFunction(()=>window.handoff.captureRunning===false&&!window.handoff.state.capturing);
 const cancelledAt=await page.evaluate(()=>window.handoff.state.frames);
 await page.locator('#step').click();await page.waitForFunction(old=>window.handoff.state.frames>old&&!window.handoff.state.running,cancelledAt,{timeout:60000});
 await page.locator('#reset').click();await page.waitForFunction(()=>!document.getElementById('run').disabled&&window.handoff.state?.frames===0,{timeout:120000});
 assert.equal(await page.locator('#execution-mode').inputValue(),'reference');
 const workerErrors=await page.evaluate(()=>window.handoff.messages.filter(m=>['error','memory-error'].includes(m.type)));assert.deepEqual(workerErrors,[]);assert.deepEqual(errors,[]);
 await page.close();
 const fallback=await browser.newPage({viewport:{width:1400,height:1000}});fallback.on('pageerror',e=>errors.push(e.message));fallback.setDefaultTimeout(120000);
 await fallback.addInitScript(()=>{const Original=window.Worker;window.Worker=class extends Original{constructor(url,options){const base=new URL(url,location.href).href;const source="import "+JSON.stringify(base)+";Object.defineProperty(navigator,'gpu',{value:undefined});const originalFetch=fetch;globalThis.fetch=(input,...args)=>originalFetch(typeof input==='string'?new URL(input,"+JSON.stringify(base)+"):input,...args);";const blob=URL.createObjectURL(new Blob([source],{type:'text/javascript'}));super(blob,options);this.addEventListener('message',()=>URL.revokeObjectURL(blob),{once:true});}};});
 await fallback.goto('http://127.0.0.1:8790/tools/browser/tests/3ds-acceleration.html');await fallback.locator('#open-media').click();await fallback.locator('#files').setInputFiles(media);await fallback.locator('#statefile').setInputFiles(savedPath);
 await fallback.waitForFunction(()=>!document.getElementById('run').disabled,null,{timeout:120000}).catch(async e=>{console.log('Fallback status:',await fallback.locator('#status').textContent(),errors);throw e;});
 assert.equal(await fallback.locator('#execution-mode').inputValue(),'reference');assert.equal(await fallback.locator('#execution-mode option[value=experimental]').isDisabled(),true);assert.match(await fallback.locator('#execution-note').textContent(),/WebGPU is unavailable/);assert.deepEqual(errors,[]);
 const result={result:'PASS',browser:await browser.version(),checks:['Reference default','Experimental live GPU draws','Pause boundary','Reference Render capture with complete replay and pixel evidence','Play preference retained','Resume Experimental','Memory Reference handoff','Preference saved and restored','Pause cancels automatic mode-switch resume','Capture cancellation releases ownership','Reference frame stepping','Reset defaults to Reference','Saved Experimental preference loads in Reference without WebGPU']};
 fs.writeFileSync(out,JSON.stringify(result,null,2)+'\n');console.log(JSON.stringify(result));
}finally{await browser.close();}
