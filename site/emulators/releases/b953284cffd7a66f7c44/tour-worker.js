// Read-only, bounded tours. Package predicates are data, never JavaScript.
export function matches(c,s,ram){
 if(c.mode!==undefined)return s.mode===c.mode;
 if(c.pc!==undefined)return s.boundary&&!s.interruptPending&&s.nextPC===c.pc&&s.mapping?.[0]===0;
 if(c.register!==undefined){const i=['ES','CS','SS','DS','FS','GS'].indexOf(c.register);return (i>=0?s.segments?.[i]?.selector:s.registers?.[c.register])===c.value;}
 if(c.address!==undefined)return (ram[c.address]&(c.mask??255))===c.value;
 if(c.all)return c.all.every(x=>matches(x,s,ram));
 if(c.any)return c.any.some(x=>matches(x,s,ram));
 throw Error('Unsupported tour condition');
}
function target(c){if(c.pc!==undefined)return c.pc;if(c.all){for(const x of c.all){const pc=target(x);if(pc!==null)return pc;}}return null;}
export function intervalEvidence(ranges,before,ram,events,dropped,limited=false){
 const counts=new Map(),details=[];let writes=0;const storageWrites={};
 for(const e of events){if(e.kind&2)storageWrites[e.region]=(storageWrites[e.region]||0)+1;if(e.region!==0||!(e.kind&2)||!ranges.some(r=>e.offset>=r.address&&e.offset<r.address+r.length))continue;writes++;counts.set(e.offset,(counts.get(e.offset)||0)+1);if(details.length<2048)details.push(e);}
 const changes=[];let changed=0;
 ranges.forEach((r,i)=>{for(let j=0;j<r.length;j++){const a=r.address+j,b=before[i][j],after=ram[a],n=counts.get(a)||0;if(b!==after)changed++;if(b!==after||n)changes.push({address:a,before:b,after,writes:n});}});
 return {storageWrites,otherWrites:Object.values(storageWrites).reduce((a,b)=>a+b,0)-writes,scope:'physical RAM writes in selected ranges; endpoint bytes at interval boundaries',complete:!dropped&&!limited,dropped,writes,changed,changes,events:details,detailsTruncated:writes>details.length};
}
export function createTourService({core,knowledge,generation,send,sleep,busy,paint,snapshot,onStart=()=>{},onRestore=()=>{},resolveTour=t=>t,validate=()=>{},maxCheckpointBytes=4*1024*1024}){
 let run=null,tour=null,index=-1,phase='idle',anchor=null,anchorEvidence=null,lastRequest=0,interval=null;
 const ram=()=>core.HEAPU8.subarray(core._rr_ram(),core._rr_ram()+(core._rr_ram_size?.()||65536));
 const raw=()=>JSON.parse(core.UTF8ToString(core._rr_debug_snapshot(-1)));
 function save(){const n=core._rr_state_save();if(n<=0||n>maxCheckpointBytes)throw Error('Tour checkpoint unavailable or over budget');return core.HEAPU8.slice(core._rr_state_data(),core._rr_state_data()+n);}
 function sameAnchor(){if(!anchor)return false;const b=save();return b.length===anchor.length&&b.every((v,i)=>v===anchor[i]);}
 function guards(){validate();for(const g of tour.guards){const s=JSON.parse(core.UTF8ToString(core._rr_debug_snapshot(g.address)));if(!g.bytes.every((b,i)=>s.mapping[i]===0&&s.bytes[i]===b))throw Error('Live code signature or RAM mapping differs from this tour.');}}
 function emit(request,text='',evidence=interval){send('tour-state',{protocol:1,generation,request,phase,index,tourId:tour?.id??null,text,evidence,canRestore:!!anchor,stop:index>=0?tour?.stops[index]:null,snapshot:snapshot()});}
 function invalidate(text='Free exploration changed the machine. Restore the tour stop before continuing.'){if(tour&&['paused-at-stop','completed'].includes(phase)){phase='diverged';emit(0,text);}}
 async function execute(request,next){
  const startCycle=String(core._rr_cycle()),stop=tour.stops[next],owner=run={cancel:false};interval=null;phase=next===0?'preparing':'continuing';emit(request);let recording=false,before=[],cycles=0,hits=0,reason='',began=performance.now(),limited=false,lastPaint=began;
  try{
   onStart();guards();before=stop.capture.map(r=>ram().slice(r.address,r.address+r.length));
   core._rr_inspect_regions();core._rr_activity_begin(2);recording=true;
   phase='running-to-stop';emit(request);
   const pc=target(stop.until);let skip=next>0;
   while(true){
    if(owner.cancel){reason='cancelled';break;}
    if(core._rr_activity_dropped()||core._rr_activity_count()>=stop.eventBudget){limited=true;reason='trace-overflow';break;}
    if(performance.now()-began>=stop.wallBudget||cycles>=stop.cycleBudget){reason='budget';break;}
    const s=raw();
    if(!skip&&s.boundary&&matches(stop.until,s,ram())){hits++;if(hits>=stop.hitCount){guards();if(!stop.assertions.every(c=>matches(c,s,ram())))throw Error('Stop assertion failed; explanation is not verified.');reason='stop';break;}}
    const mode=pc===null?(s.boundary?1:0):2;
    if(!core._rr_debug_begin(mode,pc??0,1))throw Error('Cannot start tour execution');
    // Continue an armed core job across slices; rearming mid-instruction would
    // lose next-visit semantics and could miss a target.
    let result=0;
    do{
     const old=core._rr_cycle();result=core._rr_debug_run(Math.min(1000,stop.cycleBudget-cycles));cycles+=core._rr_cycle()-old;
     if(result<0||result===5||result===6||result===7)throw Error('CPU halted or target mapping changed');
     if(performance.now()-lastPaint>=200){paint();lastPaint=performance.now();}
     validate();await sleep(0);
    }while(!result&&!owner.cancel&&cycles<stop.cycleBudget&&performance.now()-began<stop.wallBudget&&core._rr_activity_count()<stop.eventBudget&&!core._rr_activity_dropped());
    skip=false;
   }
  }catch(e){reason='failed: '+e.message;}
  finally{
   if(recording){core._rr_activity_end();limited ||= core._rr_activity_count()>=stop.eventBudget;const n=core._rr_activity_count(),p=core._rr_activity_data(),v=new DataView(core.HEAPU8.buffer,p,n*36);
    function* events(){for(let i=0;i<n;i++){const at=i*36;yield {cycle:String(BigInt(v.getUint32(at+4,true))*4294967296n+BigInt(v.getUint32(at,true))),region:v.getUint32(at+8,true),offset:v.getUint32(at+12,true),value:v.getUint32(at+16,true),kind:v.getUint32(at+24,true),pc:v.getUint32(at+28,true)};}}
    interval={...intervalEvidence(stop.capture,before,ram(),events(),core._rr_activity_dropped(),limited),stopId:stop.id,startCycle,endCycle:String(core._rr_cycle()),cycles,elapsedMs:performance.now()-began};
   }
   run=null;
   if(reason==='stop'){index=next;phase=next===tour.stops.length-1?'completed':'paused-at-stop';try{anchor=save();anchorEvidence=interval;}catch(e){phase='failed';reason=e.message;}}
   else phase=reason==='cancelled'?'cancelled':'failed';
   emit(request,reason==='stop'?stop.explanation:reason);paint();
  }
 }
 async function request(m){
  if(m.protocol!==1||m.generation!==generation||!Number.isSafeInteger(m.request)||m.request<=lastRequest){send('tour-rejected',{generation,request:m.request,text:'Stale tour request.'});return;}
  lastRequest=m.request;
  if(m.type==='tour-cancel'){if(run)run.cancel=true;else{phase='cancelled';emit(m.request,'Tour cancelled.');}return;}
  if(run||busy()){send('tour-rejected',{generation,request:m.request,text:'Pause the machine and finish other inspection jobs first.'});return;}
  try{
   if(m.type==='tour-start'){
    const t=knowledge?.data?.tours?.[m.id];if(knowledge?.status!=='matched'||!t?.releases.includes(knowledge.releaseId))throw Error('No exact supported media match for this tour.');
    tour={...resolveTour(t),id:m.id};index=-1;anchor=null;anchorEvidence=null;interval=null;phase='preparing';guards();
    if(!matches(tour.start,raw(),ram()))throw Error(tour.requirements);
    return await execute(m.request,0);
   }
   if(m.type==='tour-restore'){
    if(!anchor)throw Error('No tour stop checkpoint to restore.');onStart();const p=core._rr_state_input(anchor.length);if(!p)throw Error('Checkpoint allocation failed');core.HEAPU8.set(anchor,p);if(!core._rr_state_load(anchor.length))throw Error('Checkpoint restore failed');onRestore();guards();interval=anchorEvidence;phase=index===tour.stops.length-1?'completed':'paused-at-stop';emit(m.request,'Restored tour stop.');paint();return;
   }
   if(m.type==='tour-explore'){invalidate();return;}
   if(m.type==='tour-continue'){
    if(phase!=='paused-at-stop'||!sameAnchor())throw Error('Machine diverged from the tour stop. Restore it before continuing.');
    return await execute(m.request,index+1);
   }
   throw Error('Unsupported tour command');
  }catch(e){phase='failed';emit(m.request,e.message);}
 }
 return {request,active:()=>!!run,cancel:()=>{if(run)run.cancel=true;},invalidate};
}
