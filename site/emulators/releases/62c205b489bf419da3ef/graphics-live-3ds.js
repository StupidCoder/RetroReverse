import {create3DSGraphics} from './graphics-3ds.js';

// Every operation finishes its readback before WASM resumes. This conservative
// bridge keeps RAM authoritative across *all* direct views and physical aliases.
// GPU storage is scratch, never the only copy of guest-visible data.
export async function createLive3DSGraphics(core,{createGPU=create3DSGraphics,timeoutMs=5000}={}){
 let gpu=null,reason='',closed=false,active=false,pending=null;
 let count=0,totalMs=0;
 const fail=e=>{reason=String(e?.message??e);active=false;};
 const bounded=async(work,label)=>{let timer;try{return await Promise.race([work,new Promise((_,reject)=>{timer=setTimeout(()=>reject(Error(label+' timed out')),timeoutMs);})]);}finally{clearTimeout(timer);}};
 let initializing=true;
 try{await bounded((async()=>{
  const candidate=await createGPU({onLost:fail,measureGPU:false});
  if(!initializing){candidate.destroy();return;}gpu=candidate;
  const check=await gpu.execute({kind:1,params:[4,0x12345678,4096,0],input:new Uint8Array(),before:new Uint8Array(4096)});
  if(!initializing)return;
  if(!check.supported||!(check.bytes instanceof Uint8Array)||check.bytes.length!==4096||check.bytes.some((b,i)=>b!==[0x78,0x56,0x34,0x12][i%4]))throw Error('WebGPU transfer self-test failed');
  await gpu.preparePresentation?.();
 })(),'WebGPU initialization');
 }catch(e){fail(e);gpu?.destroy();}finally{initializing=false;}
 core.graphicsTransfer=packet=>{
  if(!active||closed||reason)return Promise.resolve({supported:false,reason:reason||'Reference execution'});
  if(pending)return Promise.reject(Error('A GPU operation is already in flight'));
  const start=performance.now();
  pending=bounded(Promise.resolve().then(()=>gpu.execute(packet)),'GPU completion')
   .then(result=>{if(result.supported){count++;totalMs+=performance.now()-start;}return result;})
   .catch(e=>{fail(e);gpu?.destroy();return {supported:false,reason};})
   .finally(()=>{pending=null;});
  return pending;
 };
 const reference={activate(){active=false;core._rr_graphics_enable(0);}};
 const experimental={
  get available(){return !!gpu&&!reason&&!core.graphicsError&&!closed;},get reason(){return reason||core.graphicsError||'WebGPU transfers with reference PICA fallbacks';},
  label:'WebGPU + software PICA',stats:()=>result.stats(),
  prepare(){if(!this.available)throw Error(this.reason);},
  activate(){if(!this.available)throw Error(this.reason);active=true;core._rr_graphics_enable(1);},
  async quiesce(){await pending;}
 };
 const result={reference,experimental,get failed(){return !!(reason||core.graphicsError);},get reason(){return reason||core.graphicsError;},
  present(){if(!active||reason||closed)return null;try{return gpu.present(core);}catch(e){fail(e);return null;}},
  stats(){return {...JSON.parse(core.UTF8ToString(core._rr_graphics_stats())),hostOperations:count,hostMs:totalMs,allocatedBytes:gpu?.bytesAllocated??0};},
  dispose(){closed=true;active=false;gpu?.destroy();core.graphicsTransfer=null;}
 };return result;
}
