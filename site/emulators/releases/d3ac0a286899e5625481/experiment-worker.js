import {matches} from './tour-worker.js';
import {digest} from './state.js';
// Explicit experiments own execution from preparation through return/release.
// Every mutation is package-declared and checked; source media is never written.
export function createExperimentService({core,knowledge,generation,send,sleep,busy,paint,snapshot,captureHost=()=>({}),restoreHost=()=>{},clearInput=()=>{},onStart=()=>{}}){
 let anchor=null,host=null,prepared=null,definition=null,phase='idle',run=null,serial=0,lastRequest=0,experimentId=null,branches=[],deterministic=false,released=false;
 const ram=()=>{const p=core._rr_ram();return core.HEAPU8.subarray(p,p+65536);};
 const raw=()=>JSON.parse(core.UTF8ToString(core._rr_debug_snapshot(-1)));
 const save=()=>{const n=core._rr_state_save();if(n<=0||n>4*1024*1024)throw Error('Experiment checkpoint exceeds 4 MiB or is unavailable');return core.HEAPU8.slice(core._rr_state_data(),core._rr_state_data()+n);};
 const load=b=>{const p=core._rr_state_input(b.length);if(!p)throw Error('Checkpoint allocation failed');core.HEAPU8.set(b,p);if(!core._rr_state_load(b.length))throw Error('Checkpoint restore failed');};
 function emit(request,text=''){send('experiment-state',{generation,request,protocol:1,phase,experimentId,branches,deterministic,canReturn:!!anchor,released,text,snapshot:snapshot()});}
 function assertConditions(list){const s=raw(),b=ram();if(!list.every(c=>matches(c,s,b)))throw Error('Experiment invariant failed');}
 function guards(){for(const g of definition.guards){const s=JSON.parse(core.UTF8ToString(core._rr_debug_snapshot(g.address)));if(!g.bytes.every((v,i)=>s.mapping[i]===0&&s.bytes[i]===v))throw Error('Experiment code signature or mapping differs');}}
 function edit(list){if(!list.length)return;const b=[];for(const e of list)for(let i=0;i<e.before.length;i++)b.push((e.address+i)&255,(e.address+i)>>8,e.before[i],e.after[i]);core.HEAPU8.set(b,core._rr_input());if(!core._rr_debug_edit(b.length/4))throw Error('Checked edit rejected; no bytes were changed');}
 function resetSession(){load(anchor);restoreHost(structuredClone(host));anchor=prepared=null;host=null;released=false;}
 function cancelled(){if(run.cancel)throw Error('cancelled');if(performance.now()-run.started>definition.wallBudget)throw Error('Experiment wall budget exceeded');}
 async function advance(cycles,pc=null,onHit=null,inputs=[]){
  let used=0,inputIndex=0,armed=false;
  while(used<cycles){cancelled();while(inputIndex<inputs.length&&inputs[inputIndex].cycle===used){core._rr_joystick(2,inputs[inputIndex++].buttons);}
   const count=Math.min(1000,cycles-used,inputIndex<inputs.length?inputs[inputIndex].cycle-used:1000),before=core._rr_cycle();let result;
   if(pc===null){if(core._rr_run(count,0,0)<0)throw Error('Core execution failed');}
   else{if(!armed){if(!core._rr_debug_begin(2,pc,1))throw Error('Breakpoint unavailable');armed=true;}result=core._rr_debug_run(count);if(result<0||result===5||result===6)throw Error('Execution or mapping failed');}
   used+=core._rr_cycle()-before;
   if(result===4){if(!onHit)return;onHit(used);armed=false;}
   if(performance.now()-run.paint>100){paint();run.paint=performance.now();}await sleep(0);
  }
  if(pc!==null&&!onHit)throw Error('Preparation breakpoint budget exceeded');
 }
 function readings(){const b=ram();return Object.fromEntries(definition.watches.map(w=>[w.id,b[w.address]]));}
 async function branch(label,patched){
  load(prepared);clearInput();if(patched)edit(definition.patch);
  const id=`${generation}:${serial}:${label}`,observations=[],start=String(core._rr_cycle());
  const take=offset=>{if(observations.length>=512)throw Error('Observation budget exceeded');const s=raw();observations.push({branchId:id,offset,pc:s.nextPC,values:readings(),predicates:definition.observations.filter(o=>matches(o.when,s,ram())).map(o=>o.id)});};
  take(0);phase='running-'+label;emit(run.request);
  await advance(definition.duration,definition.probePC,take,definition.inputs);
  const endpoint=save(),hash=await digest(endpoint);cancelled();
  const result={id,label,startCycle:start,endCycle:String(core._rr_cycle()),stateSHA256:hash,observations,final:readings()};branches.push(result);return result;
 }
 async function request(m){
  if(m.protocol!==1||m.generation!==generation||!Number.isSafeInteger(m.request)||m.request<=lastRequest){send('experiment-rejected',{generation,request:m.request,text:'Stale experiment request'});return;}lastRequest=m.request;
  if(m.type==='experiment-cancel'){if(run)run.cancel=true;else if(busy()){send('experiment-rejected',{generation,request:m.request,text:'Pause execution before returning the experiment session.'});}else if(anchor){resetSession();phase='returned';emit(m.request,'Cancelled; original session restored.');paint();}return;}
  if(run||busy()){send('experiment-rejected',{generation,request:m.request,text:'Pause and finish other execution jobs first.'});return;}
  try{
   if(m.type==='experiment-prepare'){
    if(anchor)throw Error('Return to the original session before starting another experiment');
    definition=knowledge?.data?.experiments?.[m.id];if(knowledge?.status!=='matched'||!definition?.releases.includes(knowledge.releaseId)||!core._rr_debug_edit)throw Error('No supported exact-image experiment');
    guards();assertConditions([definition.start]);host=structuredClone(captureHost());anchor=save();serial++;experimentId=m.id;branches=[];deterministic=false;released=false;phase='preparing';
   }else if(m.type==='experiment-return'){
    if(!anchor)throw Error('No original session checkpoint');resetSession();phase='returned';emit(m.request,'Original session and input queue restored.');paint();return;
   }else if(m.type==='experiment-keep'){
    if(phase!=='completed')throw Error('Complete the comparison first');released=true;phase='modified-session';clearInput();emit(m.request,'Continuing the modified branch. Return session remains available; source media is unchanged.');return;
   }else if(m.type==='experiment-original'){
    if(phase!=='prepared')throw Error('Prepare the experiment first');
   }else if(m.type==='experiment-modified'){
    if(phase!=='original-complete'||!deterministic)throw Error('Repeat original runs must first prove determinism');
   }else throw Error('Unsupported experiment request');
   run={request:m.request,cancel:false,started:performance.now(),paint:performance.now()};onStart();clearInput();emit(m.request);
   if(m.type==='experiment-prepare'){
    for(const stage of definition.setup){guards();assertConditions(stage.require);edit(stage.edits);if(stage.cycles)await advance(stage.cycles);await advance(stage.budget,stage.pc);assertConditions(stage.assertions);cancelled();}
    guards();assertConditions(definition.invariants);prepared=save();phase='prepared';emit(m.request,'Prepared checkpoint verified. Review the scene, then run the original twice.');
   }else if(m.type==='experiment-original'){
    const a=await branch('original-1',false),b=await branch('original-2',false);
    const comparable=r=>JSON.stringify(r.observations.map(({branchId,...o})=>o));deterministic=a.stateSHA256===b.stateSHA256&&comparable(a)===comparable(b);
    if(!deterministic)throw Error('Inconclusive: original replay was not deterministic');phase='original-complete';emit(m.request,'Both original replays match: complete machine state and observations.');
   }else{
    await branch('modified',true);phase='completed';emit(m.request,'Comparison complete. Observations are measured results, not a claim that the repair works universally.');
   }
  }catch(e){if(anchor&&!released){try{resetSession();}catch(restore){phase='restore-failed';emit(m.request,restore.message);return;}}phase=e.message==='cancelled'?'cancelled':'failed';emit(m.request,e.message+'; original session restored when an experiment anchor was present.');}
  finally{run=null;paint();}
 }
 return {request,active:()=>!!run,owns:()=>!!anchor&&!released,cancel:()=>{if(run)run.cancel=true;}};
}
