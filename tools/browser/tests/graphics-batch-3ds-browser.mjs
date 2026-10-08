import assert from 'node:assert/strict';
import fs from 'node:fs';
const {chromium}=await import(process.env.PLAYWRIGHT_MODULE||'playwright');
const [out]=process.argv.slice(2);
const browser=await chromium.launch({channel:'chrome',headless:true});
try{
 const page=await browser.newPage(),errors=[];page.on('pageerror',e=>errors.push(e.message));
 await page.goto('http://127.0.0.1:8790/tools/browser/tests/graphics-3ds.html');
 const result=await page.evaluate(async()=>{
  const {decodeGraphicsStream,create3DSGraphics}=await import('/site/emulators/graphics-3ds.js');
  const records=decodeGraphicsStream(await(await fetch('/tools/platform/n3ds/browser/work/graphics-batch-public.bin')).arrayBuffer());
  const gpu=await create3DSGraphics(),runs=[];
  const equal=(a,b)=>a.length===b.length&&a.every((v,i)=>v===b[i]);
  try{
   for(const size of [1,2,3,4,8])for(let i=0;i<records.length;){
    const group=[records[i++]];while(i<records.length&&group.length<size&&records[i].before.length===group[0].before.length){
     if(!equal(records[i].before,group.at(-1).expected))throw Error('Native fixture is not a continuous surface history');group.push(records[i++]);
    }
    const before=group.map(p=>p.before.slice()),r=await gpu.executeBatch(group);
    runs.push({size:group.length,first:i-group.length,supported:r.supported,bytesMatch:r.supported&&equal(r.bytes,group.at(-1).expected),countsMatch:r.supported&&r.counts.every((c,j)=>c.drawn===group[j].params[34]&&c.depthKilled===(group[j].kind>=7?group[j].params[55]:0)),inputsUnchanged:group.every((p,j)=>equal(p.before,before[j])),timing:r.timing});
   }
   const original=records.slice(0,3),bad={...original[1],input:original[1].input.slice()},v=new DataView(bad.input.buffer),stride=bad.kind===8||bad.kind===10?21:bad.kind>=7?14:13,header=bad.params[36]*bad.params[37]*2;
   for(let t=0;t<bad.params[35];t++){const at=header+t*(5+stride*3);v.setFloat32((at+4)*4,1,true);for(let j=0;j<3;j++){v.setFloat32((at+5+j*stride)*4,.5,true);v.setFloat32((at+6+j*stride)*4,.5,true);}}
   const rejected=await gpu.executeBatch([original[0],bad,original[2]]);
   const recovered=await gpu.executeBatch(original);
   const malformed=[];for(const packets of [[],Array(9).fill(records[0]),[records[0],records.at(-1)],[records[0],{...records[1],params:[]}]] )malformed.push(!(await gpu.executeBatch(packets)).supported);
   return {schema:1,records:records.length,runs,arithmeticRejection:!rejected.supported&&!rejected.bytes,recovery:recovered.supported&&equal(recovered.bytes,original.at(-1).expected),malformed,allocatedBytes:gpu.bytesAllocated,adapter:gpu.info};
  }finally{gpu.destroy();}
 });
 assert.deepEqual(errors,[]);assert(result.runs.every(r=>r.supported&&r.bytesMatch&&r.countsMatch&&r.inputsUnchanged));assert(result.arithmeticRejection&&result.recovery);assert(result.malformed.every(Boolean));
 result.result='PASS';result.browser=await browser.version();if(out)fs.writeFileSync(out,JSON.stringify(result,null,2)+'\n');console.log(JSON.stringify({result:'PASS',records:result.records,batches:result.runs.length}));
}finally{await browser.close();}
