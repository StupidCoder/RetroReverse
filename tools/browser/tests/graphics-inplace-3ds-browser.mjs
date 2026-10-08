import assert from 'node:assert/strict';
import fs from 'node:fs';
const {chromium}=await import(process.env.PLAYWRIGHT_MODULE||'playwright');
const [out]=process.argv.slice(2),browser=await chromium.launch({channel:'chrome',headless:true});
try{
 const page=await browser.newPage(),errors=[];page.on('pageerror',e=>errors.push(e.message));
 await page.goto('http://127.0.0.1:8790/tools/browser/tests/graphics-3ds.html');
 const rows=await page.evaluate(async()=>{
  const {decodeGraphicsStream,create3DSGraphics}=await import('/site/emulators/graphics-3ds.js');
  const read=async name=>decodeGraphicsStream(await(await fetch('/tools/platform/n3ds/browser/work/'+name)).arrayBuffer());
  const records=await read('graphics-public.bin'),raster=await read('graphics-sparse-public.bin');
  const transfers=[1,2,3,4,5].map(kind=>records.find(p=>p.kind===kind&&p.params.at(-1)===0));
  const equal=(a,b)=>a.length===b.length&&a.every((v,i)=>v===b[i]);
  const check=(ok,why)=>{if(!ok)throw Error(why);};
  const matches=(r,p)=>r.supported&&equal(r.bytes,p.expected)&&(p.kind<5||r.drawn===p.params[34])&&(p.kind<7||r.depthKilled===p.params[55]);
  const rows=[];
  for(const options of [{sparseRaster:true,inPlaceRaster:true},{sparseRaster:false,inPlaceRaster:true},{sparseRaster:true,inPlaceRaster:false}]){
   const gpu=await create3DSGraphics({...options,earlyRejection:true,specializeMaterials:false,specializeLighting:false});let singles=0,batches=0,recoveries=0,sparseDraws=0,inPlaceDraws=0;
   try{
    // Alternate paths on one backend so slot-zero bind groups and resized
    // allocations cannot accidentally retain the other path's source/target.
    for(const p of raster)for(const q of transfers){
     const r=await gpu.execute(p);check(matches(r,p),'Raster after transfer differs');
     check(r.inPlace===options.inPlaceRaster,'Wrong raster binding mode');
     check(options.sparseRaster||!r.sparse,'Unexpected sparse dispatch');
     sparseDraws+=Number(r.sparse);inPlaceDraws+=Number(r.inPlace);
     const transfer=await gpu.execute(q);check(matches(transfer,q)&&!transfer.inPlace,'Transfer after raster differs');singles+=2;
    }
    for(let i=0;i<raster.length;){
     const group=[raster[i++]];while(i<raster.length&&group.length<8&&raster[i].before.length===group[0].before.length)group.push(raster[i++]);
     const before=group.map(p=>p.before.slice());
     const compare=r=>r.supported&&equal(r.bytes,group.at(-1).expected)&&r.counts.every((c,j)=>c.inPlace===options.inPlaceRaster&&c.drawn===group[j].params[34]&&c.depthKilled===(group[j].kind>=7?group[j].params[55]:0));
     check(compare(await gpu.executeBatch(group)),'Ordered batch differs');batches++;
     // Fail after an earlier pass may already have modified the GPU target.
     // Recovery must reinitialize the entire target from canonical input.
     if(group.length<2)continue;
     const original=group[1],bad={...original,input:original.input.slice()},view=new DataView(bad.input.buffer),header=bad.params[36]*bad.params[37]*2,stride=[8,10].includes(bad.kind)?21:bad.kind>=7?14:13;
     for(let t=0;t<bad.params[35];t++){const at=header+t*(5+stride*3);view.setFloat32((at+4)*4,1,true);for(let v=0;v<3;v++){view.setFloat32((at+5+v*stride)*4,.5,true);view.setFloat32((at+6+v*stride)*4,.5,true);}}
     const failed=await gpu.executeBatch([group[0],bad]);check(!failed.supported&&!failed.bytes&&failed.reason==='Raster arithmetic requires Reference','Failed batch exposed partial output');
     check(compare(await gpu.executeBatch(group)),'Recovery reused partial GPU writes');recoveries++;
     check(group.every((p,j)=>equal(p.before,before[j])),'Canonical input changed');
    }
    check(!options.sparseRaster||sparseDraws>0,'Sparse path never executed');
    rows.push({...options,singles,batches,recoveries,sparseDraws,inPlaceDraws});
   }finally{gpu.destroy();}
  }
  return rows;
 });
 assert.deepEqual(errors,[]);assert(rows.every(r=>r.recoveries>0));
 const result={schema:1,result:'PASS',browser:await browser.version(),rows};if(out)fs.writeFileSync(out,JSON.stringify(result,null,2)+'\n');console.log(JSON.stringify(result));
}finally{await browser.close();}
