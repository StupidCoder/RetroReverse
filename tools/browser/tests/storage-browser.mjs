const release=(await(await fetch('/site/emulators/release.json')).json()).id;
const style=document.createElement('link');style.rel='stylesheet';style.href='/site/emulators/releases/'+release+'/style.css';document.head.append(style);
const {createStorageSource,createStoragePanel}=await import('/site/emulators/releases/'+release+'/storage-panel.js');
const assert=(v,t)=>{if(!v)throw Error(t);},sleep=ms=>new Promise(r=>setTimeout(r,ms));async function until(fn){for(let i=0;i<400&&!fn();i++)await sleep(20);assert(fn(),'Timed out');}
let report;try{
 const clients=[];const feed={subscribe(c){clients.push(c);return()=>{};},visibility(){}};
 const source=createStorageSource('amiga'),left=document.querySelector('#left'),right=document.querySelector('#right');for(const root of [left,right])createStoragePanel({root,source,platform:'amiga',feed}).setActive(true);
 const b=new Uint8Array(901120),v=new DataView(b.buffer),text=(at,s)=>b.set(new TextEncoder().encode(s),at);text(0,'DOS');for(const [block,type,name]of [[880,1,'Volume'],[900,2,'Folder'],[901,0xfffffffd,'File.bin']]){v.setUint32(block*512,2);v.setUint32(block*512+508,type);b[block*512+432]=name.length;text(block*512+433,name);}v.setUint32(880*512+24,900);v.setUint32(900*512+24,901);v.setUint32(901*512+324,1234);
 const medium=new File([b],'fixture.adf');source.set([medium]);await until(()=>left.querySelector('summary'));
 left.querySelector('summary').click();await until(()=>left.querySelectorAll('summary').length===2);[...left.querySelectorAll('summary')][1].click();await until(()=>left.textContent.includes('File.bin'));assert(left.textContent.includes('1.2 KB'),'one-decimal file size');assert(right.querySelectorAll('summary').length===1,'independent tree expansion');
 const file=[...left.querySelectorAll('.storage-entry button')].find(e=>e.textContent==='File.bin');file.click();await until(()=>left.querySelector('pre').textContent);assert(left.querySelector('[aria-label="Storage sector"]').value==='901','file navigates to header sector');assert(right.querySelector('[aria-label="Storage view"]').value==='files','other view unchanged');
 const view=right.querySelector('[aria-label="Storage view"]');view.value='sectors';view.dispatchEvent(new Event('change'));await until(()=>right.querySelector('pre').textContent);assert(right.querySelector('[aria-label="Storage sector"]').value==='0','independent sector');
 assert(right.querySelectorAll('.disk-surfaces canvas').length===2,'both ADF sides');
 const diskCanvas=right.querySelector('.disk-surfaces canvas'),rect=diskCanvas.getBoundingClientRect();
 diskCanvas.dispatchEvent(new MouseEvent('click',{clientX:rect.left+rect.width/2,clientY:rect.top+rect.height*11/512,bubbles:true}));
 await until(()=>right.querySelector('[aria-label="Storage sector"]').value==='0');
 assert(right.querySelector('.disk-legend').textContent.includes('unavailable'),'unloaded image has no live head');
 source.loaded(medium);await until(()=>right.querySelector('summary'));view.value='sectors';view.dispatchEvent(new Event('change'));
 await until(()=>clients.some(c=>c.active));for(const c of clients)if(c.active)c.overview({state:{track:35}});
 assert(right.querySelector('.disk-legend').textContent.includes('track 17, side 1'),'bound live track and side');
 const side1=right.querySelectorAll('.disk-surfaces canvas')[1],pixels=side1.getContext('2d').getImageData(0,0,512,512).data;
 let orange=0;for(let i=0;i<pixels.length;i+=4)if(pixels[i]===255&&pixels[i+1]===123&&pixels[i+2]===36)orange++;
 assert(orange>100,'live head ring is rendered');
 source.set([new File([b],'other.adf')]);await until(()=>right.querySelector('.storage-status').textContent.includes('other.adf'));view.value='sectors';view.dispatchEvent(new Event('change'));
 assert(!clients.some(c=>c.active),'different disk cannot borrow live head');
 assert(right.querySelector('.disk-legend').textContent.includes('unavailable'),'old head cleared on image change');
 report={result:'PASS' ,release,checks:['expandable filesystem tree','human-readable sizes','file-to-sector navigation','independent panel navigation','two-sided ADF atlas and click navigation','live head ring bound to loaded medium']};
}catch(e){report={result:'FAIL',release,error:e.stack||String(e)};}
document.querySelector('#result').textContent=JSON.stringify(report,null,2);document.title=report.result+' — Storage';const id=new URLSearchParams(location.search).get('report');if(id)await fetch('/__viewport_result__?id='+encodeURIComponent(id),{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify(report)});
