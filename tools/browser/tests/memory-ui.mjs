import {createWorkspaces} from '../../../site/emulators/workspaces.js';
import {createMemoryWorkspace} from '../../../site/emulators/memory-workspace.js';
const $=id=>document.getElementById(id),assert=(x,s)=>{if(!x)throw Error(s);};
try{
 const sent=[];let id=0;
 const views=createWorkspaces({navigation:$('nav')});views.register({id:'play',label:'Play',panel:$('play')});
 const ui=createMemoryWorkspace({root:$('memory'),views,resume:()=>{},send:(type,args)=>{sent.push({type,...args,request:++id});return id;}});
 ui.ready();views.select('memory');ui.open();const snap=sent.at(-1);
 $('memory-scale').value='1';$('memory-scale').dispatchEvent(new Event('change'));
 assert(sent.length===1,'a scale change during an async snapshot must not invalidate the snapshot request');
 const region={id:'ram',index:0,name:'Work RAM',kind:'ram',size:512,base:0xc000,aliases:[0xc000],scale:1,bitmap:new Uint8Array(512),activityMap:new Uint8Array(512),mapSize:512,units:'bytes'};
 ui.result({type:'memory-overview',request:snap.request,overview:{id:10,regions:[region],labels:[{region:'ram',start:32,length:16,name:'Known region',description:'Fixture'}],activity:false,recording:0,count:0,position:0}});
 const page=sent.findLast(m=>m.type==='memory-page');
 ui.result({type:'memory-page',request:999,page:{id:10,region:'ram',offset:0,bytes:Uint8Array.of(255)}});
 assert(!$('memory-bytes').children.length,'stale pages must be ignored');
 ui.result({type:'memory-page',request:page.request,page:{id:10,region:'ram',offset:0,bytes:Uint8Array.of(65,66,67,68)}});
 assert($('memory-bytes').textContent.includes('41')&&$('memory-bytes').textContent.includes('ABCD'),'hex and ASCII show actual page values');
 $('memory-search').value='Known';$('memory-search').dispatchEvent(new Event('input'));
 assert($('memory-labels').children.length===1,'label search filters hardware and annotation entries');$('memory-labels').firstChild.click();
 assert(sent.findLast(m=>m.type==='memory-page').offset===32,'label selects its precise offset');
 const canvas=$('memory-atlas').querySelector('canvas'),box=canvas.getBoundingClientRect();canvas.dispatchEvent(new MouseEvent('click',{clientX:box.left+box.width/2,clientY:box.top+box.height/4}));
 assert(sent.at(-1).type==='memory-map','bitmap click uses the shared map protocol');
 const before=sent.length;ui.open();assert(sent.length===before,'returning to Memory retains the paused inspection');
 ui.reset();assert($('view-memory').disabled,'reset disables obsolete session navigation');
 $('results').textContent='PASS: async snapshot race, stale pages, hex/ASCII, label filtering, bitmap navigation and session reset.';document.title='PASS — Memory UI checks';
}catch(e){$('results').textContent='FAIL: '+e.stack;document.title='FAIL — Memory UI checks';throw e;}
