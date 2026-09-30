const root=new URL('../../../',import.meta.url),out=document.getElementById('result');
const assert=(value,message)=>{if(!value)throw Error(message);};
async function snapshot(release,file) {
 const worker=new Worker(new URL(`site/emulators/releases/${release}/worker.js`,root),{type:'module'});
 try {
  return await new Promise((resolve,reject)=>{
   const timer=setTimeout(()=>reject(Error('Worker timeout')),15000);
   const done=(fn,value)=>{clearTimeout(timer);fn(value);};
   worker.onerror=e=>done(reject,Error(e.message));
   worker.onmessage=({data:m})=>{
    if(m.type==='error'||m.type==='memory-error')done(reject,Error(m.text));
    if(m.type==='ready')worker.postMessage({type:'memory-snapshot',session:1,request:2});
    if(m.type==='memory-overview')done(resolve,m.overview);
   };
   worker.postMessage({type:'load',session:1,platform:'c64',files:[file]});
  });
 }finally{worker.terminate();}
}
try {
 const {id}=await (await fetch(new URL('site/emulators/release.json',root))).json();
 const {knowledgePackages}=await import(new URL(`site/emulators/releases/${id}/knowledge-data.js`,root));
 const pkg=knowledgePackages.find(p=>p.id==='fort-apocalypse-c64');assert(pkg,'package missing from release');
 const tape=new Uint8Array(21);tape.set(new TextEncoder().encode('C64-TAPE-RAW'));tape[16]=1;tape[20]=48;
 const generic=await snapshot(id,new File([tape],'synthetic.tap'));
 assert(generic.labels.length===0,'unknown image received labels');
 assert(generic.regions.some(r=>r.id==='ram'),'unknown image lost RAM inspector');
 const report={pass:true,release:id,packageSHA256:pkg.sourceSHA256,unknown:{labels:0,regions:generic.regions.length},fort:'skipped; use ?fort=1 with local reference tape'};
 if(new URL(location.href).searchParams.has('fort')){
  const response=await fetch(new URL('games/fort-apocalypse-c64/Fort_Apocalypse.tap',root));assert(response.ok,'local Fort tape unavailable');
  const known=await snapshot(id,new File([await response.arrayBuffer()],'Fort_Apocalypse.tap'));
  assert(known.labels.length===13,'expected 13 Fort labels');
  assert(known.labels.find(l=>l.id==='terrain')?.start===0x503,'wrong terrain location');
  report.fort={labels:known.labels.length,regions:known.regions.length,names:known.labels.map(l=>l.name)};
 }
 out.textContent=JSON.stringify(report,null,2);document.title='PASS — M1 knowledge';
}catch(e){out.textContent=String(e.stack||e);document.title='FAIL — M1 knowledge';}
