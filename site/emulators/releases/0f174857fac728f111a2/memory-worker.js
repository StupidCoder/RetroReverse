import {summarize,applyEvents,EVENT_WORDS,tapIndex} from './memory-model.js';
import {memoryLabels} from './memory-labels.js';

// Owns copies only: inspecting and scrubbing never invokes device read handlers.
export function createMemoryService({core,platform,files,status}){
 let snapshot=null,initial=null,events=null,position=0,serial=0,recording=0,labels=[],tape=null,recordStart=null,recordEnd=null,live=false,asset=null,labelCache=null;
 const immutable=new Map(),bitmaps=new WeakMap();
 let tapeBitmap=null;
 const id=()=>++serial;
 async function take(){
  if(!core._rr_inspect_regions)return {id:id(),regions:[],activity:false,note:'Physical memory inspection is not yet available for this core.'};
  const info=JSON.parse(core.UTF8ToString(core._rr_inspect_regions()));
  const regions=info.regions.flatMap(r=>{const p=core._rr_inspect_data(r.index);if(r.kind==='disk')return Array.from({length:r.size/5632},(_,track)=>({...r,index:r.index+track,id:'disk-'+track,name:`ADF · cylinder ${Math.floor(track/2)} / side ${track%2} · sectors 0–10`,size:5632,base:track*5632,bytes:core.HEAPU8.slice(p+track*5632,p+(track+1)*5632)}));return [{...r,bytes:(r.kind==='rom'&&immutable.has(r.id))?immutable.get(r.id):core.HEAPU8.slice(p,p+r.size)}];});
  for(const r of regions)if(r.kind==='rom')immutable.set(r.id,r.bytes);
  if(platform==='c64'){
   const bytes=asset??=new Uint8Array(await files[0].arrayBuffer());tape??=tapIndex(bytes);
   regions.push({id:'tape',index:regions.length,name:'Tape · pulse stream',kind:'tape',size:bytes.length,base:0,aliases:[],bytes});
  }
  labels=labelCache??=await memoryLabels(platform,files[0]);
  return {id:id(),regions,activity:info.activity,fetches:platform==='c64',coverage:({gg:'CPU reads include fetches; RAM and video-port writes. PPU reads are not traced.',c64:'CPU RAM/ROM accesses and tape pulses. I/O and VIC reads are not traced.',amiga:'CPU reads include fetches; chip DMA accesses and encoded disk payload consumption. I/O is not traced.'})[platform],state:status()};
 }
 function overview(scale=0,window=20000){
  if(!snapshot)return null;
  const total=snapshot.regions.reduce((n,r)=>n+r.size,0);
  const bytesPerPixel=scale===1?1:Math.max(1,2**Math.ceil(Math.log2(Math.max(1,total/65536))));
  const regions=snapshot.regions.map(r=>{
   const data=r.kind==='tape'?(tapeBitmap??=Uint8Array.from(tape.durations,n=>Math.min(255,Math.round(Math.log2(n+1)*18)))):r.bytes;
   const s=r.kind==='tape'?Math.max(1,Math.ceil(data.length/65536)):Math.max(bytesPerPixel,2**Math.ceil(Math.log2(Math.max(1,data.length/1048576))));
   let cached=bitmaps.get(data);if(!cached){cached=new Map();bitmaps.set(data,cached);}if(!cached.has(s))cached.set(s,summarize(data,s));
   return {...r,bytes:undefined,bitmap:cached.get(s),scale:s,units:r.kind==='tape'?'pulses':'bytes',mapSize:data.length,activityMap:new Uint8Array(Math.ceil(data.length/s))};
  });
  if(events){
   const start=window===0?0:Math.max(0,position-20000);
   for(let i=start;i<position;i++){const j=i*EVENT_WORDS,r=regions[events[j+2]];if(r){const a=Math.floor(events[j+3]/r.scale),b=Math.min(r.activityMap.length,Math.ceil((events[j+3]+events[j+5])/r.scale));for(let p=a;p<b;p++)r.activityMap[p]|=events[j+6]&7;}}
  }
  return {...snapshot,regions,labels,recording,live,count:events?events.length/EVENT_WORDS:0,position,start:recordStart,end:recordEnd,tapeCursor:snapshot.state?.pulse??null};
 }
 return {
  stopLive(){if(live)core._rr_activity_end?.();live=false;},
  async liveSnapshot(scale=0,window=20000,fetches=false){
   if(live&&snapshot?.activity){core._rr_activity_end();const n=core._rr_activity_count(),p=core._rr_activity_data();events=new Uint32Array(core.HEAPU8.slice(p,p+n*EVENT_WORDS*4).buffer);if(platform==='amiga')for(let i=0;i<n;i++){const j=i*EVENT_WORDS;if(events[j+2]===3){events[j+2]+=Math.floor(events[j+3]/5632);events[j+3]%=5632;}}position=n;}else{events=null;position=0;}
   const dropped=live?core._rr_activity_dropped?.()||0:0;initial=null;recording=0;snapshot=await take();snapshot.dropped=dropped;live=true;if(snapshot.activity)core._rr_activity_begin(fetches?7:3);return overview(scale,0);
  },
  async snapshot(){this.stopLive();initial=events=null;recording=0;position=0;snapshot=await take();return overview();},
  overview,
  page(region,offset){const r=snapshot?.regions.find(r=>r.id===region);if(!r)return null;offset=Math.max(0,Math.min(Math.floor(offset/16)*16,Math.max(0,Math.ceil(r.size/16)*16-256)));return {id:snapshot.id,region,offset,bytes:r.bytes.slice(offset,offset+256)};},
  async begin(fetches=false){this.stopLive();snapshot=await take();if(!snapshot.activity)throw Error('This core supports snapshots only.');initial=snapshot;events=null;recording++;recordStart=snapshot.state;core._rr_activity_begin(fetches?7:3);},
  finish(){core._rr_activity_end();const n=core._rr_activity_count(),p=core._rr_activity_data();events=new Uint32Array(core.HEAPU8.slice(p,p+n*EVENT_WORDS*4).buffer);if(platform==='amiga')for(let i=0;i<n;i++){const j=i*EVENT_WORDS;if(events[j+2]===3){events[j+2]+=Math.floor(events[j+3]/5632);events[j+3]%=5632;}}position=0;recordEnd=status();return this.seek(n);},
  seek(n){if(!events||!initial)return overview();n=Math.max(0,Math.min(Math.floor(n),events.length/EVENT_WORDS));
   if(position===0||n<position)snapshot={...initial,id:id(),regions:initial.regions.map(r=>({...r,bytes:r.bytes.slice()})),state:{...initial.state}};
   if(n<position)position=0;
   applyEvents(snapshot.regions,events,position,n);position=n;snapshot.id=id();
   if(platform==='c64'){let pulse=initial.state.pulse;for(let i=n-1;i>=0;i--){const j=i*EVENT_WORDS;if(events[j+6]===17){pulse=events[j+3]+1;break;}}snapshot.state.pulse=pulse;}
   snapshot.dropped=core._rr_activity_dropped();snapshot.limited=events.length/EVENT_WORDS>=524288;snapshot.cursorCycle=n?events[(n-1)*EVENT_WORDS]+events[(n-1)*EVENT_WORDS+1]*4294967296:initial.state.cycle??initial.state.steps;return overview();
  },
  detail(region,offset){const r=snapshot?.regions.find(r=>r.id===region);if(!r)return null;const out=[];
   let pulse;if(r.kind==='tape'){let lo=0,hi=tape.offsets.length;while(lo<hi){const mid=(lo+hi)>>>1;if(tape.offsets[mid]<=offset)lo=mid+1;else hi=mid;}pulse=Math.max(0,lo-1);}
   const at=pulse??offset;
   if(events)for(let i=position-1;i>=0&&out.length<4;i--){const j=i*EVENT_WORDS;if(events[j+2]===r.index&&at>=events[j+3]&&at<events[j+3]+events[j+5])out.push({index:i,cycle:events[j]+events[j+1]*4294967296,kind:events[j+6],pc:events[j+7],address:events[j+8],value:events[j+4]});}
   return {region,offset,events:out,...(pulse===undefined?{}:{pulse,duration:tape.durations[pulse]})};
  },
  mapOffset(region,pixel,scale){const r=snapshot?.regions.find(r=>r.id===region);const off=pixel*scale;return r?.kind==='tape'?tape.offsets[Math.min(off,tape.offsets.length-1)]:off;},
  pulses(start,count){if(!snapshot||!tape)return null;start=Math.max(0,Math.min(tape.durations.length-1,Math.floor(Number(start)||0)));count=Math.max(1,Math.min(256,Math.floor(Number(count)||128)));return {id:snapshot.id,start,total:tape.durations.length,cursor:snapshot.state?.pulse??0,durations:tape.durations.slice(start,start+count)};},
  pulseForOffset(offset){let i=0;while(i+1<tape.offsets.length&&tape.offsets[i+1]<=offset)i++;return i;}
 };
}
