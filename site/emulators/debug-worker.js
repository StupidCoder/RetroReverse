// All core calls here are synchronous; cooperative jobs yield between bounded
// slices. The worker's common execution gate excludes other machine owners.
export function createDebugService({core,send,sleep,busy,paint,applyInputs,knowledge,generation,onStart=()=>{}}) {
 let job=null,sequence=0,last=null,mappingGeneration=0,lastBank=null;
 const snapshot=(address=-1)=>{
  const s=JSON.parse(core.UTF8ToString(core._rr_debug_snapshot(address)));
  if(s.bank!==lastBank){lastBank=s.bank;mappingGeneration++;}
  last={...s,snapshotId:++sequence,generation,mappingGeneration};return last;
 };
 function reply(m,type,data={}){send(type,{protocol:1,generation,request:m.request,jobId:m.request,...data});}
 function reject(m,text){reply(m,'debug-result',{reason:'rejected',text});}
 async function request(m){
  if(!Number.isSafeInteger(m.request)||m.request<1||m.protocol!==1||m.generation!==generation){reject(m,'Stale debugger generation or unsupported protocol.');return;}
  if(m.type==='debug-capabilities'){reply(m,'debug-capabilities',{capabilities:{instructionStep:true,runToAddress:true,safePeek:true,normalizeBoundary:true,liveSampleHz:5},knowledge});return;}
  if(m.type==='debug-cancel'){if(job)job.cancel=true;reply(m,'debug-cancelled',{activeJob:job?.id??null});return;}
  if(m.type==='debug-snapshot'){
   if(busy()&&!job){reject(m,'Another inspection job owns the machine.');return;}
   const address=m.address??-1;if(!Number.isInteger(address)||address < -1||address>65535){reject(m,'Invalid CPU address.');return;}
   reply(m,'debug-snapshot',{snapshot:snapshot(address)});return;
  }
  if(!['debug-normalize','debug-step','debug-until'].includes(m.type)){reject(m,'Unsupported debugger command.');return;}
  if(job||busy(true)){reject(m,'Machine is busy. Pause execution before starting a debugger job.');return;}
  if(!last||m.snapshotId!==last.snapshotId){reject(m,'Snapshot is stale. Refresh before executing.');return;}
  const expected=last,now=snapshot();
  if(now.cycle!==expected.cycle||now.bank!==expected.bank||m.cycle!==expected.cycle||m.bank!==expected.bank){reject(m,'Machine changed since the displayed snapshot. Refresh first.');return;}
  const mode=m.type==='debug-step'?1:m.type==='debug-until'?2:0,target=m.target??0;
  if(!Number.isInteger(target)||target<0||target>65535||!['stop-if-current','next-match'].includes(m.resume??'stop-if-current')){reject(m,'Invalid target or resume policy.');return;}
  if(!core._rr_debug_begin(mode,target,+(m.resume==='next-match'))){reject(m,'Normalize to a fetch boundary before instruction stepping.');return;}
  const owner=job={id:m.request,cancel:false};
  let reason='budget',text='',cycles=0,retired=mode===1?0:undefined;
  let lastPaint=performance.now();
  const began=performance.now(),budget=Math.min(985248*30,Math.max(1,Number(m.cycleBudget)||985248*10)),wallBudget=Math.min(30000,Math.max(1,Number(m.wallBudget)||10000));
  try{
   onStart();reply(m,'debug-started');paint();
   while(!owner.cancel&&cycles<budget&&performance.now()-began<wallBudget){
    applyInputs();const before=core._rr_cycle();
    const result=core._rr_debug_run(Math.min(1000,budget-cycles));
    cycles+=core._rr_cycle()-before;
    if(result<0){reason='error';text=core.UTF8ToString(core._rr_error());break;}
    if(result){reason=({1:'boundary',2:'instruction',3:'interrupt-entry',4:'target',5:'mapping-invalid',6:'halted'})[result];if(mode===1)retired=result===2?1:0;break;}
    if(performance.now()-lastPaint>=200){paint();lastPaint=performance.now();}
    await sleep(0);
   }
   if(owner.cancel)reason='cancelled';
  }catch(e){reason='error';text=String(e);}
  finally{if(job===owner)job=null;reply(m,'debug-result',{reason,text,retired,cycles,elapsedMs:performance.now()-began,snapshot:snapshot()});paint();}
 }
 return {request,active:()=>!!job,cancel:()=>{if(job)job.cancel=true;}};
}
