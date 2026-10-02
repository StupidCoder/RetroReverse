// Execution engines share one machine. Preparation/quiescence may await host
// work; activation is synchronous and must not execute guest instructions. The
// caller holds the worker execution gate until select() settles, even if cancelled.
export function create3DSExecution({reference={},experimental=null,onChange=()=>{}}={}) {
  const engines={reference,experimental};
  let effective='reference',requested='reference',revision=0,job=null,disposed=false;
  const unavailable='Experimental rendering is not available in this build.';
  function snapshot(){return {requested,effective,transition:job?.target??null,busy:!!job,
    cpu:'ARM interpreter',renderer:effective==='reference'?'Software PICA':(engines[effective]?.label??'Experimental'),
    experimental:{available:!!experimental?.available,reason:experimental?.reason??unavailable}};}
  function publish(){if(!disposed)onChange(snapshot());}
  async function select(target){
    if(disposed)throw Error('3DS execution session ended');
    if(!Object.hasOwn(engines,target))throw Error('Unknown 3DS execution mode');
    if(target!=='reference'&&!engines[target]?.available)throw Error(engines[target]?.reason??unavailable);
    if(job)throw Error('An execution transition is already pending');
    requested=target;
    if(effective===target){publish();return snapshot();}
    const token={target,id:++revision};job=token;publish();
    const valid=()=>!disposed&&revision===token.id;
    try{
      // Preparation must not mutate guest state. A cancelled preparation is
      // allowed to finish; it cannot activate a backend in the old session.
      await engines[target].prepare?.();
      if(!valid())return snapshot();
      await engines[effective].quiesce?.();
      if(!valid())return snapshot();
      // Backend activation must be transactional on failure and must not await.
      engines[target].activate?.();
      effective=target;
    }finally{
      if(job===token){job=null;requested=effective;publish();}
    }
    return snapshot();
  }
  function cancel(){revision++;requested=effective;publish();}
  return {snapshot,select,cancel,dispose(){cancel();disposed=true;}};
}
