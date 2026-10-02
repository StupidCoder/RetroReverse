// One sampler per session, irrespective of the number of visible inspectors.
// Historical Memory recording is an explicit shared timeline; navigation is local.
export function createInspectionFeed({send,platform}){
 const clients=new Set();let loaded=false,state=null,timer=null,pending=0,requestedAt=0,overview=null,history=false,legacy=false,dirty=true,alive=true;
 function interested(){return [...clients].some(c=>c.active);}
 function blocked(){return state&&(state.capturing||state.saving||state.memoryRecording||state.debugBusy||state.experimentOwned);}
 function cancel(){clearTimeout(timer);timer=null;}
 // Frame updates must not postpone a pending sample: they arrive faster than 200ms.
 function schedule(){if(!alive||!loaded||!interested()){cancel();return;}if(timer===null)timer=setTimeout(tick,200);}
 function tick(){timer=null;if(!alive)return;if(pending&&performance.now()-requestedAt>1500)pending=0;if(!legacy&&!pending&&!history&&!blocked()&&(state?.running||dirty)){pending=send('memory-live-snapshot',{scale:0,fetches:platform==='c64'});requestedAt=performance.now();dirty=false;}schedule();}
 return {
  subscribe(client){clients.add(client);if(overview)client.overview(overview);schedule();return()=>{clients.delete(client);if(!interested()){cancel();send('memory-live-end');}};},
  visibility(){if(!interested()){cancel();send('memory-live-end');}else{dirty=true;schedule();}},
  legacy(value){legacy=value;schedule();},
  ready(){loaded=true;dirty=true;schedule();},
  reset(){cancel();loaded=false;state=null;pending=0;overview=null;history=false;dirty=true;for(const c of clients)c.reset?.();},
  state(m){const before=state?.state?.cycle??state?.state?.steps;state=m;if(m.running)history=false;if(before!==(m.state.cycle??m.state.steps))dirty=true;schedule();},
  result(m){if(m.request===pending){pending=0;if(m.type!=='memory-overview')dirty=true;}if(m.type==='memory-overview'&&m.overview){overview=m.overview;history=!!overview.recording;for(const c of clients)if(c.active)c.overview(overview);}else for(const c of clients)if(c.active)c.result?.(m);schedule();},
  refresh(){history=false;dirty=true;schedule();},
  dispose(){alive=false;cancel();clients.clear();}
 };
}
