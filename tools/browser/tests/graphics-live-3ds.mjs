import assert from 'node:assert/strict';
import {createLive3DSGraphics} from '../../../site/emulators/graphics-live-3ds.js';
const core=()=>({_rr_graphics_enable(n){this.enabled=n;},_rr_graphics_stats:()=>'{"operations":[0,0,0,0],"accelerated":[0,0,0,0]}',UTF8ToString:s=>s});
const selfTest=()=>({supported:true,bytes:Uint8Array.from({length:4096},(_,i)=>[0x78,0x56,0x34,0x12][i%4])});
{
 const c=core(),b=await createLive3DSGraphics(c,{createGPU:async()=>{throw Error('No adapter');}});
 assert(!b.experimental.available);assert.match(b.reason,/No adapter/);assert.throws(()=>b.experimental.activate());b.reference.activate();assert.equal(c.enabled,0);b.dispose();
}
{
 let resolve,calls=0,destroyed=0;const c=core(),b=await createLive3DSGraphics(c,{createGPU:async()=>({execute(){return ++calls===1?selfTest():new Promise(r=>resolve=r);},destroy(){destroyed++;}})});
 b.experimental.activate();const pending=c.graphicsTransfer({});await assert.rejects(c.graphicsTransfer({}),/in flight/);
 let quiet=false;const q=b.experimental.quiesce().then(()=>quiet=true);await Promise.resolve();assert(!quiet);resolve({supported:true,bytes:new Uint8Array(4)});await pending;await q;assert(quiet);b.reference.activate();assert.equal(c.enabled,0);b.dispose();assert.equal(destroyed,1);
}
{
 let calls=0,destroyed=0;const c=core(),b=await createLive3DSGraphics(c,{timeoutMs:5,createGPU:async()=>({execute:()=>++calls===1?selfTest():new Promise(()=>{}),destroy(){destroyed++;}})});
 b.experimental.activate();const result=await c.graphicsTransfer({});assert.equal(result.supported,false);assert(b.failed);assert(!b.experimental.available);assert.match(b.reason,/timed out/);assert.equal(destroyed,1);b.dispose();
}
{
 let lose;const c=core(),b=await createLive3DSGraphics(c,{createGPU:async({onLost})=>{lose=onLost;return {execute:selfTest,destroy(){}};}});
 b.experimental.activate();lose('Lost during a frame');assert(!b.experimental.available);assert.equal((await c.graphicsTransfer({})).supported,false);b.reference.activate();assert.equal(c.enabled,0);b.dispose();
}
console.log('3DS live bridge: adapter failure, exclusive operation, quiescence, timeout and device loss pass');
