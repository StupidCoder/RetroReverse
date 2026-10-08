import assert from 'node:assert/strict';
import fs from 'node:fs';
const {chromium}=await import(process.env.PLAYWRIGHT_MODULE||'playwright');
const [out]=process.argv.slice(2),browser=await chromium.launch({channel:'chrome',headless:true});
try{
 const page=await browser.newPage(),errors=[];page.on('pageerror',e=>errors.push(e.message));
 await page.exposeFunction('materialProgress',p=>console.log(JSON.stringify(p)));
 await page.goto('http://127.0.0.1:8790/tools/browser/tests/graphics-3ds.html');
 const result=await page.evaluate(async()=>{
  const {decodeGraphicsStream,create3DSGraphics}=await import('/site/emulators/graphics-3ds.js');
  const read=async name=>decodeGraphicsStream(await(await fetch('/tools/platform/n3ds/browser/work/'+name)).arrayBuffer());
  const records=(await read('graphics-public.bin')).filter(p=>p.kind>=5),sparse=await read('graphics-sparse-public.bin');
  const equal=(a,b)=>a.length===b.length&&a.every((v,i)=>v===b[i]);
  const rows=[],batches=[],empty=[],arithmetic=[];
  const check=(ok,why)=>{if(!ok)throw Error(why);};
  const ready=async gpu=>{const end=performance.now()+60000;while(gpu.materialCompilation().pending||gpu.lightingCompilation().pending){check(performance.now()<end,'Compilation timeout');await new Promise(r=>setTimeout(r,5));}check(!gpu.materialCompilation().failed&&!gpu.lightingCompilation().failed,'Specialization failed to compile');};
  const matches=(r,p)=>r.supported&&equal(r.bytes,p.expected)&&r.drawn===p.params[34]&&r.depthKilled===(p.kind>=7?p.params[55]:0);
  // Recreate the bounded cache in chunks so EVERY Reference fixture executes a
  // ready material shader, even though the full corpus exceeds the cache cap.
  for(let start=0;start<records.length;start+=48){
   const gpu=await create3DSGraphics({sparseRaster:true});
   try{for(let i=start;i<Math.min(start+48,records.length);i++){
    const p=records[i],warm=await gpu.execute(p);check(matches(warm,p),'Warm output differs at '+i);await ready(gpu);
    const r=await gpu.execute(p);check(r.materialSpecialized&&matches(r,p),'Material output differs at '+i);
    rows.push({index:i,kind:p.kind,bytes:r.bytes.length,drawn:r.drawn,depthKilled:r.depthKilled,material:r.materialSpecialized});
   }}finally{gpu.destroy();}
   await window.materialProgress({records:rows.length,total:records.length});
  }
  const gpu=await create3DSGraphics({sparseRaster:true});try{
   for(const p of sparse){check(matches(await gpu.execute(p),p),'Sparse warm output differs');await ready(gpu);const r=await gpu.execute(p);check(r.materialSpecialized&&matches(r,p),'Sparse specialized output differs');rows.push({kind:p.kind,sparse:r.sparse,groups:r.workgroups,bytes:r.bytes.length,drawn:r.drawn,depthKilled:r.depthKilled,material:r.materialSpecialized});}
   for(const size of [1,2,3,4,8])for(let i=0;i<sparse.length;){
    const group=[sparse[i++]];while(i<sparse.length&&group.length<size&&sparse[i].before.length===group[0].before.length)group.push(sparse[i++]);
    for(let j=1;j<group.length;j++)check(equal(group[j].before,group[j-1].expected),'Discontinuous fixture');
    const before=group.map(p=>p.before.slice()),r=await gpu.executeBatch(group);
    check(r.supported&&equal(r.bytes,group.at(-1).expected),'Sparse batch bytes differ');
    check(r.counts.every((c,j)=>c.drawn===group[j].params[34]&&c.depthKilled===(group[j].kind>=7?group[j].params[55]:0)),'Sparse batch counters differ');
    check(group.every((p,j)=>equal(p.before,before[j])),'Mutated input snapshot');
    batches.push({size:group.length,sparse:r.counts.filter(c=>c.sparse).length,material:r.counts.filter(c=>c.materialSpecialized).length});
   }
   // Zero-area bounds have no bin links or writes. Repack the validated packet
   // without links; the entire initialized surface pair must survive unchanged.
   for(const kind of [6,7]){
    const original=sparse.find(p=>p.kind===kind),p={...original,params:[...original.params],expected:original.before};
    const old=new Uint32Array(original.input.buffer,original.input.byteOffset,original.input.length/4),header=p.params[36]*p.params[37]*2,stride=kind===6?13:14,end=header+p.params[35]*(5+3*stride),tail=old[header-2]+old[header-1],words=new Uint32Array(old.length-(tail-end));
    words.set(old.subarray(0,end));words.set(old.subarray(tail),end);for(let b=0;b<header;b+=2){words[b]=end;words[b+1]=0;}
    for(let t=header;t<end;t+=5+3*stride){words[t+1]=words[t];words[t+3]=words[t+2];}
    for(let u=0;u<3;u++)if(p.params[38]&(1<<u))p.params[39+u*4]-=tail-end;
    p.input=new Uint8Array(words.buffer);p.params[9]=p.params[34]=0;if(kind>=7)p.params[55]=0;
    const r=await gpu.execute(p);check(r.sparse&&r.workgroups===1&&matches(r,p),'Empty draw changed a surface');empty.push(kind);
   }
   for(const kind of [6,7,8,9,10]){
    const original=sparse.find(p=>p.kind===kind),p={...original,input:original.input.slice()},view=new DataView(p.input.buffer),header=p.params[36]*p.params[37]*2,stride=[8,10].includes(kind)?21:kind>=7?14:13;
    for(let t=0;t<p.params[35];t++){const at=header+t*(5+stride*3);view.setFloat32((at+4)*4,1,true);for(let v=0;v<3;v++){view.setFloat32((at+5+v*stride)*4,.5,true);view.setFloat32((at+6+v*stride)*4,.5,true);}}
    const r=await gpu.execute(p);check(!r.supported&&!r.bytes&&r.reason==='Raster arithmetic requires Reference','Invalid arithmetic was committed');arithmetic.push(kind);
   }
   return {schema:1,records:rows,batches,empty,arithmetic,materialCompilation:gpu.materialCompilation(),lightingCompilation:gpu.lightingCompilation(),adapter:gpu.info};
  }finally{gpu.destroy();}
 });
 assert.deepEqual(errors,[]);assert(result.records.some(r=>r.sparse));assert(result.batches.some(r=>r.size>1&&r.sparse>0&&r.sparse<r.size));assert.equal(result.empty.length,2);assert.equal(result.arithmetic.length,5);
 result.result='PASS';result.browser=await browser.version();if(out)fs.writeFileSync(out,JSON.stringify(result,null,2)+'\n');console.log(JSON.stringify({result:'PASS',records:result.records.length,batches:result.batches.length}));
}finally{await browser.close();}
