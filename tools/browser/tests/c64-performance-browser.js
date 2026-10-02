import {benchmark} from './c64-performance.js';
try{
 const bytes=async p=>new Uint8Array(await(await fetch(p)).arrayBuffer()),release=await(await fetch('/site/emulators/release.json')).json();
 const roms=await Promise.all(['basic','kernal','chargen'].map(n=>bytes('/site/emulators/firmware/c64/'+n+'.rom'))),tape=await bytes('/games/fort-apocalypse-c64/Fort_Apocalypse.tap'),pkg=await(await fetch('/games/fort-apocalypse-c64/knowledge.json')).json(),reports=[];
 for(const [label,id]of [['owned',release.id],['legacy',release.rollback]]){const base='/site/emulators/releases/'+id+'/cores/c64/';reports.push(await benchmark({factory:(await import(base+'core.js')).default,wasm:await bytes(base+'core.wasm'),roms,tape,pkg,label}));postMessage({reports});}
 postMessage({result:'PASS',reports});
}catch(e){postMessage({result:'FAIL',error:e.stack});}
