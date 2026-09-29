// Private-media A/B benchmark. Each sample starts in a fresh Node process and
// restores the same checkpoint; variants run sequentially in alternating order.
import fs from 'node:fs';
import path from 'node:path';
import {fileURLToPath} from 'node:url';
import {execFileSync} from 'node:child_process';
import {createHash} from 'node:crypto';
import assert from 'node:assert/strict';

const [platform,image,state,before,after,fields='60',repeats='3']=process.argv.slice(2);
assert(platform&&image&&state&&before&&after,'platform image raw-state before-core-dir after-core-dir [fields] [repeats]');
const samples=[],hashes={};let proof;
for(const [variant,dir] of Object.entries({before,after}))
  hashes[variant]=createHash('sha256').update(fs.readFileSync(path.join(dir,'core.wasm'))).digest('hex');
for(let repeat=0;repeat<Number(repeats);repeat++){
  for(const variant of repeat%2?['after','before']:['before','after']){
    const env={...process.env,CORE_DIR:path.resolve(variant==='before'?before:after)};
    for(const name of ['RR_CPU_PROFILE','RR_AUTO','RR_SAVE','RR_PPM'])delete env[name];
    const output=execFileSync(process.execPath,[fileURLToPath(new URL('./console-bench.mjs',import.meta.url)),platform,image,state,fields],{env,encoding:'utf8'});
    const sample={variant,repeat,...JSON.parse(output.trim())};
    if(proof)assert.deepEqual(sample.proof,proof,'A/B execution or rendered output differs');else proof=sample.proof;
    samples.push(sample);
    process.stderr.write(`${platform} ${path.basename(state)} ${variant} ${repeat+1}: ${sample.ms.toFixed(1)} ms\n`);
  }
}
const median=values=>{values.sort((a,b)=>a-b);const n=values.length;return (values[(n-1)>>1]+values[n>>1])/2;};
const beforeMs=median(samples.filter(s=>s.variant==='before').map(s=>s.ms));
const afterMs=median(samples.filter(s=>s.variant==='after').map(s=>s.ms));
console.log(JSON.stringify({platform,state:path.basename(state),node:process.version,arch:process.arch,fields:Number(fields),repeats:Number(repeats),hashes,beforeMs,afterMs,speedup:beforeMs/afterMs,samples}));
