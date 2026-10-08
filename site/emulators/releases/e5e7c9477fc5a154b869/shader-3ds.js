// Bounded PICA vertex specialization. Generated WASM shares the core's memory;
// only an entire successful batch is allowed to reach canonical draw outputs.
// Arithmetic follows gpu_shader.go, including ordered min/max and binary64 RSQ.
const I32=0x7f,F32=0x7d,VOID=0x40;
const uleb=n=>{const a=[];do{const b=n&127;n=Math.floor(n/128);a.push(b|(n?128:0));}while(n);return a;};
const sleb=n=>{const a=[];for(;;){const b=n&127;n>>=7;const end=(n===0&&!(b&64))||(n===-1&&(b&64));a.push(b|(end?0:128));if(end)return a;}};
const bytes=s=>[...uleb(s.length),...new TextEncoder().encode(s)];
const section=(id,a)=>[id,...uleb(a.length),...a];
const floatBytes=(x,wide=false)=>{const b=new ArrayBuffer(wide?8:4),v=new DataView(b);wide?v.setFloat64(0,x,true):v.setFloat32(0,x,true);return [...new Uint8Array(b)];};

export function compilePicaShader(code,descriptors,entry){
 if(code.length!==4096||descriptors.length!==128||!Number.isInteger(entry)||entry<0||entry>=4096)throw Error('Invalid PICA program');
 // Parameters: input, output, count, float uniforms, integer uniforms, bool mask.
 const types=[],body=[],local=t=>{types.push(t);return 6+types.length-1;};
 const out=Array.from({length:64},()=>local(F32)),tmp=Array.from({length:64},()=>local(F32));
 const result=Array.from({length:4},()=>local(F32));
 const ax=local(I32),ay=local(I32),al=local(I32),cx=local(I32),cy=local(I32),vertex=local(I32),fuel=local(I32);
 const emit=(...a)=>body.push(...a),get=n=>emit(0x20,...uleb(n)),set=n=>emit(0x21,...uleb(n));
 const icon=n=>emit(0x41,...sleb(n)),fcon=n=>emit(0x43,...floatBytes(n));
 const reject=()=>{icon(0);emit(0x0f);};
 const load=(ptr,offset,op=0x2a,align=2)=>{get(ptr);emit(op,align,...uleb(offset));};
 const integerGuard=n=>{get(n);fcon(-2147483648);emit(0x60);get(n);fcon(2147483648);emit(0x5d,0x71,0x45,0x04,VOID);reject();emit(0x0b);};
 const rejectNaN=n=>{get(n);get(n);emit(0x5c,0x04,VOID);reject();emit(0x0b);}; // NaN payloads stay in Reference.
 function source(reg,idx,desc,slot,c){
  const shift=[4,13,22][slot],sw=(desc>>>(shift+1+6-2*c))&3,neg=(desc>>>shift)&1;
  if(reg<16)load(0,reg*16+sw*4);
  else if(reg<32)get(tmp[(reg-16)*4+sw]);
  else if(!idx)load(3,(reg-32)*16+sw*4);
  else{
   const r=reg-32,a=[0,ax,ay,al][idx];
   // Test before addition: the Reference adds an int32 index in int64 space.
   get(a);icon(-r);emit(0x4e);get(a);icon(96-r);emit(0x48,0x71,0x04,F32);
   get(3);get(a);icon(r);emit(0x6a);icon(4);emit(0x74,0x6a,0x2a,2,...uleb(sw*4));
   if(neg)emit(0x8c);emit(0x05);fcon(0);emit(0x0b);return;
  }
  if(neg)emit(0x8c);
 }
 function condition(word,bool=false,invert=false){
  if(bool){get(5);icon((word>>>22)&15);emit(0x76);icon(1);emit(0x71);if(invert)emit(0x45);return;}
  const mode=(word>>>22)&3;
  const x=()=>{get(cx);icon((word>>>25)&1);emit(0x46);};
  const y=()=>{get(cy);icon((word>>>24)&1);emit(0x46);};
  if(mode===2)x();else if(mode===3)y();else{x();y();emit(mode===0?0x72:0x71);}
 }
 function arithmetic(word){
  const op=word>>>26,mad=op>=0x30,cmp=(op>>>1)===0x17,desc=descriptors[word&(mad?31:127)];
  const dst=mad?(word>>>24)&31:(word>>>21)&31,src=[];
  if(mad){src.push([(word>>>17)&31,0]);if(op>=0x38)src.push([(word>>>10)&127,(word>>>22)&3],[(word>>>5)&31,0]);else src.push([(word>>>12)&31,0],[(word>>>5)&127,(word>>>22)&3]);}
  else if([0x18,0x19,0x1a,0x1b].includes(op))src.push([(word>>>14)&31,0],[(word>>>7)&127,(word>>>19)&3]);
  else src.push([(word>>>12)&127,(word>>>19)&3],[(word>>>7)&31,0]);
  const s=(slot,c)=>source(...src[slot],desc,slot,c);
  if(cmp){for(let c=0;c<2;c++){const comparison=(word>>>(c?21:24))&7;if(comparison>5)icon(1);else{s(0,c);s(1,c);emit([0x5b,0x5c,0x5d,0x5f,0x5e,0x60][comparison]);}set(c?cy:cx);}return;}
  if(!mad&&![0,1,2,3,8,9,10,11,12,13,14,15,18,19,24,26,27].includes(op))throw Error('Unsupported PICA arithmetic');
  if(op===18){
   // Capture both before writing either address register (relative sources).
   for(let c=0;c<2;c++)if(desc&(8>>c)){s(0,c);set(result[c]);integerGuard(result[c]);}
   for(let c=0;c<2;c++)if(desc&(8>>c)){get(result[c]);emit(0xa8);set(c?ay:ax);}return;
  }
  const scalar=!mad&&[1,2,3,24,14,15].includes(op);
  for(let c=0;c<(scalar?1:4);c++){
   if(!scalar&&!(desc&(8>>c)))continue;
   if(mad){s(0,c);s(1,c);emit(0x94);s(2,c);emit(0x92);}
   else if([1,2,3,24].includes(op)){
    for(let k=0;k<(op===2?4:3);k++){s(0,k);s(1,k);emit(0x94);if(k)emit(0x92);}
    if(op===3||op===24){s(1,3);emit(0x92);}
   }else if(op===14){fcon(1);s(0,0);emit(0x95);}
   else if(op===15){
    s(0,0);fcon(0);emit(0x5f,0x04,F32); // Reference returns +infinity for f <= 0.
    fcon(Infinity);emit(0x05,0x44,...floatBytes(1,true));s(0,0);emit(0xbb,0x9f,0xa3,0xb6,0x0b);
   }else if(op===11){
    s(0,c);set(result[c]);integerGuard(result[c]);
    get(result[c]);emit(0xa8,0xb2);const rounded=local(F32);set(rounded);
    get(result[c]);fcon(0);emit(0x5d);get(result[c]);get(rounded);emit(0x5c,0x71,0x04,F32);
    get(rounded);fcon(1);emit(0x93,0x05);get(rounded);emit(0x0b);
   }else if(op===12||op===13){s(0,c);s(1,c);s(0,c);s(1,c);emit(op===12?0x5e:0x5d,0x1b);}
   else if([9,10,26,27].includes(op)){s(0,c);s(1,c);emit(op===9||op===26?0x60:0x5d,0xb2);}
   else if(op===19)s(0,c);
   else{s(0,c);s(1,c);emit(op===0?0x92:0x94);}
   set(result[c]);if(op!==19)rejectNaN(result[c]);
  }
  // All operands/results precede writes, even with aliased sources/destinations.
  for(let c=0;c<4;c++)if(desc&(8>>c)){get(result[scalar?0:c]);set((dst<16?out:tmp)[(dst&15)*4+c]);}
 }
 let nodes=0;const active=new Set();
 function range(start,end,depth=0,loopDepth=0){
  if(depth>16||start<0||end>4096||start>end)throw Error('Unbounded PICA flow');
  const key=start+':'+end;if(active.has(key))throw Error('Recursive PICA call');active.add(key);
  for(let pc=start;pc<end;){
   if(++nodes>4096)throw Error('PICA compiler budget');
   const word=code[pc],op=word>>>26,dst=(word>>>10)&4095,num=word&255;
   if(op===0x21){pc++;continue;}if(op===0x22)break;
   if([0x24,0x25,0x26].includes(op)){
    if(op!==0x24){condition(word,op===0x26);emit(0x04,VOID);}
    range(dst,dst+num,depth+1,loopDepth);if(op!==0x24)emit(0x0b);pc++;continue;
   }
   if(op===0x27||op===0x28){
    if(dst<=pc)throw Error('Backward PICA IF');
    condition(word,op===0x27);emit(0x04,VOID);range(pc+1,dst,depth+1,loopDepth);emit(0x05);range(dst,dst+num,depth+1,loopDepth);emit(0x0b);pc=dst+num;continue;
   }
   if(op===0x29){
    if(dst<pc||loopDepth>=2)throw Error('Unbounded PICA LOOP');
    const n=local(I32),limit=local(I32),step=local(I32),ir=((word>>>22)&3)*4;
    load(4,ir+1,0x2d,0);set(al);load(4,ir,0x2d,0);set(limit);load(4,ir+2,0x2d,0);set(step);icon(0);set(n);
    emit(0x03,VOID);get(fuel);emit(0x45,0x04,VOID);reject();emit(0x0b);get(fuel);icon(1);emit(0x6b);set(fuel);
    range(pc+1,dst+1,depth+1,loopDepth+1);get(al);get(step);emit(0x6a);set(al);get(n);icon(1);emit(0x6a,0x22,...uleb(n));get(limit);emit(0x4d,0x0d,0,0x0b);pc=dst+1;continue;
   }
   if(op===0x2c||op===0x2d){
    if(dst<=pc)throw Error('Backward PICA jump');condition(word,op===0x2d,!!(word&1)&&op===0x2d);emit(0x04,VOID);
    if(dst<end)range(dst,end,depth+1,loopDepth);emit(0x05);range(pc+1,end,depth+1,loopDepth);emit(0x0b);break;
   }
   arithmetic(word);pc++;
  }
  active.delete(key);
 }
 get(2);emit(0x45,0x04,VOID);icon(1);emit(0x0f,0x0b);
 emit(0x03,VOID); // One compiled invocation for the entire batch.
 for(const n of [...out,...tmp]){fcon(0);set(n);}for(const n of [ax,ay,al,cx,cy]){icon(0);set(n);}icon(65536);set(fuel);
 range(entry,4096);
 for(let i=0;i<64;i++){get(1);get(out[i]);emit(0x38,2,...uleb(i*4));}
 for(const p of [0,1]){get(p);icon(256);emit(0x6a);set(p);}
 get(vertex);icon(1);emit(0x6a,0x22,...uleb(vertex));get(2);emit(0x49,0x0d,0,0x0b);icon(1);emit(0x0b);
 if(body.length>256*1024)throw Error('PICA generated code budget');
 const locals=[...uleb(types.length),...types.flatMap(t=>[1,t])],fn=[...locals,...body];
 return new Uint8Array([0,97,115,109,1,0,0,0,
  ...section(1,[1,0x60,6,I32,I32,I32,I32,I32,I32,1,I32]),
  ...section(2,[1,...bytes('env'),...bytes('memory'),2,0,0]),
  ...section(3,[1,0]),...section(7,[1,...bytes('run'),0,0]),
  ...section(10,[1,...uleb(fn.length),...fn])]);
}

export function createPicaShaders(){
 const programs=new Map();let disposed=false;
 const stats={requested:0,ready:0,unsupported:0,failed:0,batches:0,vertices:0,refused:0,compileLatencyMs:0};
 return {
  prepare(id,code,descriptors,entry,memory){
   if(disposed||programs.has(id)||programs.size>=128)return;
   const p={run:null};programs.set(id,p);stats.requested++;
   const start=performance.now(),words=code.slice(),operands=descriptors.slice();
   // Even binary construction runs after this guest call yields. Uploaded code
   // may change meanwhile, so the task owns immutable copies of both arrays.
   Promise.resolve().then(async()=>{
    if(disposed)return;let binary;
    try{binary=compilePicaShader(words,operands,entry);}catch{stats.unsupported++;return;}
    try{
     const {instance}=await WebAssembly.instantiate(binary,{env:{memory}});
     if(!disposed){p.run=instance.exports.run;stats.ready++;stats.compileLatencyMs+=performance.now()-start;}
    }catch{stats.failed++;}
   });
  },
  ready(id){return !!programs.get(id)?.run;},
  run(id,...args){
   const p=programs.get(id);if(!p?.run||disposed)return 0;
   try{const ok=p.run(...args);if(ok){stats.batches++;stats.vertices+=args[2];}else stats.refused++;return ok;}
   catch{p.run=null;stats.failed++;return 0;}
  },
  snapshot(){return {...stats};},
  dispose(){disposed=true;programs.clear();}
 };
}
