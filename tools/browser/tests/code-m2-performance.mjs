import fs from 'node:fs';
import {probeC64} from './code-m0-probe.mjs';
const release=JSON.parse(fs.readFileSync(new URL('../../../site/emulators/release.json',import.meta.url)));
const bases={before:`../../../site/emulators/releases/${release.previous}/cores/c64/`,after:'../../../site/emulators/cores/c64/'};
const cores={},hashes={};
const {createHash}=await import('node:crypto');
for(const [key,path] of Object.entries(bases)){
 const base=new URL(path,import.meta.url),binary=fs.readFileSync(new URL('core.wasm',base));
 hashes[key]=createHash('sha256').update(binary).digest('hex');cores[key]=await (await import(new URL('core.js',base))).default({wasmBinary:binary});
}
const samples={before:[],after:[]};
for(let i=0;i<6;i++)for(const key of i%2?['after','before']:['before','after']){
 const result=probeC64(cores[key]);if(i)samples[key].push(...result.throughput.map(s=>s.cyclesPerSecond));
}
const median=a=>[...a].sort((a,b)=>a-b)[Math.floor(a.length/2)];
const before=median(samples.before),after=median(samples.after);
const result={schema:1,date:new Date().toISOString(),node:process.version,arch:process.arch,hashes,samples,medianCyclesPerSecond:{before,after},disabledOverheadPercent:(before/after-1)*100,limitations:'Interleaved synthetic INC/JMP core-only samples; disabled debugger. Compilation/code layout and system noise affect results. Not a game or worker benchmark.'};
fs.writeFileSync(new URL('../results/code-m2/performance.json',import.meta.url),JSON.stringify(result,null,2)+'\n');console.log(JSON.stringify(result,null,2));
