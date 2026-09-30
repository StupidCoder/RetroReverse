// Data-only, bounded decoding. Values remain decimal strings to preserve 64 bits.
export function typeSize(types,id,depth=0){
 if(depth>=16)throw Error('Type depth limit');const t=types[id];if(!t)throw Error('Unknown type');
 const n=t.kind==='integer'?t.bytes:['enum','fixed','bitfield'].includes(t.kind)?typeSize(types,t.storage,depth+1):t.kind==='struct'?t.size:t.kind==='array'?t.count*t.stride:NaN;
 if(!Number.isSafeInteger(n)||n<1||n>16384)throw Error('State byte budget exceeded');return n;
}
export function decodeStateType(types,id,bytes,{offset=0,path='',budget={nodes:4096},depth=0}={}){
 if(depth>=16||--budget.nodes<0)throw Error('State node/depth budget exceeded');
 const t=types[id],size=typeSize(types,id),raw=bytes.slice(0,size),node={type:id,kind:t.kind,path,offset,size,raw,status:'known'};
 const unavailable=raw.length!==size||raw.some(b=>b==null);
 if(unavailable)node.status='unavailable';
 const child=(type,at,key)=>decodeStateType(types,type,raw.slice(at),{offset:offset+at,path:path?path+'.'+key:String(key),budget,depth:depth+1});
 if(t.kind==='struct'){node.children=t.fields.map(f=>({...child(f.type,f.offset,f.name),label:f.name,description:f.description}));return node;}
 if(t.kind==='array'){
  if(t.count>256)throw Error('State slot budget exceeded');node.children=Array.from({length:t.count},(_,i)=>({...child(t.element,i*t.stride,i),label:'Slot '+i,index:i}));return node;
 }
 if(unavailable)return node;
 const base=t.kind==='integer'?t:types[t.storage];let n=0n;
 for(const b of base.endian==='little'?[...raw].reverse():raw)n=(n<<8n)|BigInt(b);
 const unsigned=n;if(base.signed&&(n&(1n<<BigInt(size*8-1))))n-=1n<<BigInt(size*8);
 node.value=n.toString();
 if(t.kind==='enum'){node.enumLabel=t.values[node.value];if(node.enumLabel===undefined)node.status='unknown';}
 if(t.kind==='fixed'){
  const denominator=1n<<BigInt(t.fractionBits),abs=n<0n?-n:n;
  const remainder=abs%denominator;
  node.value=(n<0n?'-':'')+(abs/denominator).toString()+(remainder?'.'+(remainder*5n**BigInt(t.fractionBits)).toString().padStart(t.fractionBits,'0').replace(/0+$/,''):'');
  node.storageValue=n.toString();
 }
 if(t.kind==='bitfield')node.children=Object.entries(t.bits).map(([key,b])=>({path:path+'.'+key,label:b.label,kind:'bits',offset,size,raw,value:((unsigned>>BigInt(b.offset))&((1n<<BigInt(b.width))-1n)).toString(),status:'known'}));
 return node;
}
export function resolveStateLocation(k,id,depth=0){
 if(depth>=16)throw Error('Location depth limit');const l=k.locations[id];if(!l)throw Error('Unknown location');
 if(l.kind==='relative'){const b=resolveStateLocation(k,l.base,depth+1);return {...b,address:b.address+Number(l.offset)};}
 const space=k.spaces[l.space];
 if(l.kind==='cpu')return {kind:'cpu',address:Number(l.address),space:l.space};
 if(l.kind==='physical'&&space?.region==='ram')return {kind:'ram',address:Number(l.offset),space:l.space};
 throw Error('Live location is unavailable for this adapter');
}
export function createStateSampler(knowledge,read){
 const history=new Map();
 return function sample(context){
  const k=knowledge?.data;if(!k)return {cycle:context.cycle,boundary:context.boundary,entries:[]};
  let remaining=16384;const budget={nodes:4096};
  const entries=Object.entries(k.state||{}).filter(([,d])=>d.releases.includes(knowledge.releaseId)).map(([id,d])=>{
   const out={id,label:d.label,applicability:d.applicability,evidence:d.evidence.map(e=>k.evidence[e]),relatedFunctions:d.relatedFunctions||[]};
   try{
    const location=resolveStateLocation(k,d.location),size=typeSize(k.types,d.type);remaining-=size;if(remaining<0)throw Error('Snapshot state byte budget exceeded');
    const span=read(location,size),node=decodeStateType(k.types,d.type,span.bytes,{budget});
    if(d.occupancy&&node.children)for(const slot of node.children){
     let value=slot;for(const field of d.occupancy.path)value=value?.children?.find(c=>c.label===field);
     slot.occupancy=!value||value.status==='unavailable'||value.value===undefined?'unavailable':value.value===BigInt(d.occupancy.notEquals).toString()?'empty':'occupied';
     const key=id+':'+slot.index,old=history.get(key);let incarnation=old?.incarnation||0;
     if(slot.occupancy==='occupied'&&old?.occupancy!=='occupied')incarnation++;
     slot.incarnation=incarnation;history.set(key,{occupancy:slot.occupancy,incarnation});
    }
    return {...out,location,mapping:span.mapping,node};
   }catch(e){history.forEach((_,key)=>{if(key.startsWith(id+':'))history.delete(key);});return {...out,error:String(e.message||e)};}
  });
  return {cycle:context.cycle,boundary:context.boundary,entries};
 };
}
