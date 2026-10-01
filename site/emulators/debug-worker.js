import {createStateSampler} from './structured-state.js';
// All core calls here are synchronous; cooperative jobs yield between bounded
// slices. The worker's common execution gate excludes other machine owners.
export function createDebugService({core,send,sleep,busy,paint,applyInputs,knowledge,generation,onStart=()=>{},platform="c64",resolveLocation,decorate=s=>s,prepareTarget=()=>()=>{}}) {
 const limit=core._rr_ram_size?.()||65536;
 const sampleState=createStateSampler(knowledge,(location,size)=>{
  const address=location.address;
  if(!Number.isSafeInteger(address)||address<0||address+size>limit)throw Error('State range is outside backing storage');
  if(location.kind==='ram'){const p=core._rr_ram();return {bytes:Array.from(core.HEAPU8.slice(p+address,p+address+size)),mapping:Array(size).fill(0)};}
  const bytes=[],mapping=[];
  for(let offset=0;offset<size;offset+=256){const s=JSON.parse(core.UTF8ToString(core._rr_debug_snapshot(address+offset)));bytes.push(...s.bytes.slice(0,size-offset));mapping.push(...s.mapping.slice(0,size-offset));}
  return {bytes,mapping};
 },resolveLocation);
 const snapshots=new Map(),clients=new Map();
 let job=null,sequence=0,last=null,mappingGeneration=0,lastBank=null;
 const snapshot=(address=-1,client="default")=>{
  const s=decorate(JSON.parse(core.UTF8ToString(core._rr_debug_snapshot(address))));
  if((s.contextKey??s.bank)!==lastBank){lastBank=s.contextKey??s.bank;mappingGeneration++;}
  last={...s,snapshotId:++sequence,generation,mappingGeneration,state:sampleState(s)};snapshots.delete(clients.get(client));clients.delete(client);clients.set(client,last.snapshotId);snapshots.set(last.snapshotId,last);if(clients.size>64){const oldest=clients.keys().next().value;snapshots.delete(clients.get(oldest));clients.delete(oldest);}return last;
 };
 function reply(m,type,data={}){send(type,{protocol:1,generation,request:m.request,jobId:m.request,...data});}
 function reject(m,text){reply(m,'debug-result',{reason:'rejected',text});}
 async function request(m){
  if(!Number.isSafeInteger(m.request)||m.request<1||m.protocol!==1||m.generation!==generation){reject(m,'Stale debugger generation or unsupported protocol.');return;}
  const client=m.client??'default';if(typeof client!=='string'||!/^[a-zA-Z0-9-]{1,100}$/.test(client)){reject(m,'Invalid inspector identity.');return;}
  if(m.type==='debug-capabilities'){reply(m,'debug-capabilities',{capabilities:{instructionStep:true,stepOver:platform==='c64',stepOut:platform==='c64',writeWatchpoint:platform==='c64',runToAddress:true,safePeek:true,normalizeBoundary:['c64','1541'].includes(platform),liveSampleHz:5,architecture:platform==='dos'?'x86':platform==='3do'?'arm60':platform==='1541'?'6502':'6510'},knowledge});return;}
  if(m.type==='debug-cancel'){if(job)job.cancel=true;reply(m,'debug-cancelled',{activeJob:job?.id??null});return;}
  if(m.type==='debug-snapshot'){
   if(busy()&&!job){reply(m,'debug-result',{reason:'rejected',text:'Another inspection job owns the machine.',retryable:true});return;}
   const address=m.address??-1;if(!Number.isInteger(address)||address < -1||address>=limit){reject(m,'Invalid CPU address.');return;}
   reply(m,'debug-snapshot',{snapshot:snapshot(address,client)});return;
  }
  if(!['debug-normalize','debug-step','debug-until','debug-over','debug-out','debug-watch'].includes(m.type)){reject(m,'Unsupported debugger command.');return;}
  if(job||busy(true)){reject(m,'Machine is busy. Pause execution before starting a debugger job.');return;}
  if(!snapshots.has(m.snapshotId)||clients.get(client)!==m.snapshotId){reject(m,'Snapshot is stale. Refresh before executing.');return;}
  const expected=snapshots.get(m.snapshotId),now=snapshot(-1,client);
  if(now.cycle!==expected.cycle||now.bank!==expected.bank||now.contextKey!==expected.contextKey||m.cycle!==expected.cycle||m.bank!==expected.bank){reject(m,'Machine changed since the displayed snapshot. Refresh first.');return;}
  const mode=({'debug-step':1,'debug-until':2,'debug-over':3,'debug-out':4,'debug-watch':5})[m.type]||0,target=m.target??0;
  if(mode>=3&&platform!=='c64'){reject(m,'This core does not support advanced stepping or watchpoints.');return;}
  if(!Number.isInteger(target)||target<0||(mode===2&&target>=limit)||!['stop-if-current','next-match'].includes(m.resume??'stop-if-current')){reject(m,'Invalid target or resume policy.');return;}
  let validateTarget;try{validateTarget=m.type==='debug-until'?prepareTarget(m,now):()=>{};}catch(e){reject(m,e.message);return;}
  const w=m.watch;
  if(mode===5&&(!w||![w.address,w.value,w.mask,w.writer].every(Number.isInteger)||w.address<2||w.address>65535||w.value<0||w.value>255||w.mask<0||w.mask>255||(w.value&w.mask)!==w.value||w.writer < -1||w.writer>65535)){reject(m,'Invalid masked write condition.');return;}
  if(!(mode===5?core._rr_debug_watch(w.address,w.value,w.mask,w.writer):core._rr_debug_begin(mode,target,+(m.resume==='next-match')))){reject(m,'The core cannot start this job in the current execution context.');return;}
  validateTarget.bindTarget?.();
  const owner=job={id:m.request,cancel:false};
  let reason='budget',text='',cycles=0,retired=mode===1?0:undefined;
  let lastPaint=performance.now();
  const began=performance.now(),budget=Math.min(985248*30,Math.max(1,Number(m.cycleBudget)||985248*10)),wallBudget=Math.min(30000,Math.max(1,Number(m.wallBudget)||10000));
  try{
   onStart();reply(m,'debug-started');paint();
   while(!owner.cancel&&cycles<budget&&performance.now()-began<wallBudget){
    validateTarget();applyInputs();const before=core._rr_cycle();
    const result=core._rr_debug_run(Math.min(1000,budget-cycles));
    cycles+=core._rr_cycle()-before;
    validateTarget();if(result<0){reason='error';text=core.UTF8ToString(core._rr_error());break;}
    if(result){reason=({1:'boundary',2:'instruction',3:'interrupt-entry',4:'target',5:'mapping-invalid',6:'halted',7:'unsupported-or-atomic-budget',8:'hle-service',9:'task-switch',10:'instruction-with-hle',11:'display-boundary',12:'return',13:'write-watchpoint',14:'return-context-invalid'})[result];if(mode===1)retired=result===2||result===10?1:0;break;}
    if(performance.now()-lastPaint>=200){paint();lastPaint=performance.now();}
    await sleep(0);
   }
   if(owner.cancel)reason='cancelled';
  }catch(e){reason='error';text=String(e);}
  finally{if(job===owner)job=null;reply(m,'debug-result',{reason,text,retired,cycles,event:mode===5?JSON.parse(core.UTF8ToString(core._rr_debug_event())):null,elapsedMs:performance.now()-began,snapshot:snapshot(-1,client)});paint();}
 }
 return {snapshot,request,active:()=>!!job,cancel:()=>{if(job)job.cancel=true;}};
}
