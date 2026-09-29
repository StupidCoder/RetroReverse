// Wrap a private raw checkpoint for the public browser state-import workflow.
// The image, state and output must remain local; no game data belongs in Git.
import fs from 'node:fs';
import path from 'node:path';
import {createHash} from 'node:crypto';
import assert from 'node:assert/strict';
import {packState} from '../../../site/emulators/state.js';
import {InputQueue} from '../../../site/emulators/input.js';
import {platforms} from '../../../site/emulators/platforms.js';
const [platform,image,state,output]=process.argv.slice(2);
assert(platforms[platform]&&image&&state&&output,'platform image raw-state output.rrstate');
const paths=[image];
if(platform==='dc'){
 const match=/(?:DATAFILE|FILE)\s+"([^"]+)"/.exec(fs.readFileSync(image,'utf8'));assert(match);
 paths.push(path.join(path.dirname(image),match[1]));
}
const media=[];
for(const file of paths){const hash=createHash('sha256');for await(const b of fs.createReadStream(file))hash.update(b);media.push({name:path.basename(file),size:fs.statSync(file).size,sha256:hash.digest('hex')});}
media.sort((a,b)=>a.name.localeCompare(b.name));
const manifest=JSON.parse(fs.readFileSync(new URL('../../../site/emulators/build-manifest.json',import.meta.url)));
const q=new InputQueue(platforms[platform].hz),input={...q,pulses:[],down:[],pending:[],appliedKeys:[],lastButtons:-1,lastX:0,lastY:0,inputSequence:0,lastInputStep:0};
fs.writeFileSync(output,await packState({format:1,platform,media,firmware:[],core:manifest[platform+'/core.wasm'],configuration:{compatibility:true,customFirmware:false},input},fs.readFileSync(state)));
console.log('Private browser state written: '+output);
