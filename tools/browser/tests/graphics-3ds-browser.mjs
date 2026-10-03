import assert from 'node:assert/strict';
import fs from 'node:fs';
const {chromium}=await import(process.env.PLAYWRIGHT_MODULE||'playwright');
const [base='http://127.0.0.1:8790',stream='/tools/platform/n3ds/browser/work/graphics-public.bin',out]=process.argv.slice(2);
const browser=await chromium.launch({channel:'chrome',headless:true});
try{
 const page=await browser.newPage(),errors=[];page.on('pageerror',e=>errors.push(e.message));
 await page.goto(base+'/tools/browser/tests/graphics-3ds.html');
 const result=await page.evaluate(async path=>{
  const {decodeGraphicsStream,create3DSGraphics}=await import('/site/emulators/graphics-3ds.js');
  const {rasterProbe}=await import('./raster-probe-3ds.js');const coverage=await rasterProbe();
  const records=decodeGraphicsStream(await(await fetch(path)).arrayBuffer()),gpu=await create3DSGraphics(),results=[];
  try{for(let i=0;i<records.length;i++){
   const p=records[i],r=await gpu.execute(p);if(!r.supported){results.push({index:i,kind:p.kind,supported:false,reason:r.reason,expectedFallback:[2,3].includes(p.kind)&&p.params.at(-1)===1});continue;}
   let differences=0,first=-1;for(let j=0;j<p.expected.length;j++)if(p.expected[j]!==r.bytes[j]){differences++;if(first<0)first=j;}
   results.push({index:i,kind:p.kind,supported:true,bytes:r.bytes.length,differences,first,drawn:r.drawn,expectedDrawn:p.kind>=5?p.params[34]:undefined,depthKilled:r.depthKilled,expectedDepthKilled:p.kind===7?p.params[55]:undefined,...r.timing});
  }
  const arithmeticFallbacks={};
  for(const kind of [6,7]){const raster=records.find(p=>p.kind===kind);if(!raster)continue;
   const input=raster.input.slice(),view=new DataView(input.buffer),header=raster.params[36]*raster.params[37]*2;
   const stride=kind===7?14:13,words=5+stride*3;
   for(let t=0;t<raster.params[35];t++){const at=header+t*words;view.setFloat32((at+4)*4,1,true);for(let v=0;v<3;v++){view.setFloat32((at+5+v*stride)*4,.5,true);view.setFloat32((at+6+v*stride)*4,.5,true);}}
   const result=await gpu.execute({...raster,input});arithmeticFallbacks[kind]=!result.supported&&result.reason==='Raster arithmetic requires Reference';
  }
  return {schema:1,arithmeticFallback:Object.values(arithmeticFallbacks).every(Boolean),arithmeticFallbacks,adapter:gpu.info,warmupMs:gpu.warmupMs,allocatedBytes:gpu.bytesAllocated,coverage,records:results};}finally{gpu.destroy();}
 },stream);
 assert.deepEqual(errors,[]);assert(result.records.some(r=>r.supported));assert(result.records.every(r=>r.supported||r.expectedFallback),'Unexpected GPU fallback');assert(result.arithmeticFallback!==false,'Invalid interpolation committed output');assert(result.records.every(r=>!r.supported||r.differences===0),JSON.stringify(result.records.filter(r=>r.differences)));
 assert(result.records.every(r=>r.expectedDrawn===undefined||r.drawn===r.expectedDrawn),'Fragment statistics differ');
 assert(result.records.every(r=>r.expectedDepthKilled===undefined||r.depthKilled===r.expectedDepthKilled),'Depth rejection statistics differ');
 result.result='PASS';result.browser=await browser.version();if(out)fs.writeFileSync(out,JSON.stringify(result,null,2)+'\n');
 console.log(JSON.stringify({result:'PASS',operations:result.records.length,supported:result.records.filter(r=>r.supported).length,unsupported:result.records.filter(r=>!r.supported).length,warmupMs:result.warmupMs}));
}finally{await browser.close();}
