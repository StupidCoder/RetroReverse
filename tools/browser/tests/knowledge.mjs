import assert from 'node:assert/strict';
import fs from 'node:fs';
import crypto from 'node:crypto';
import {knowledgePackages} from '../../../site/emulators/knowledge-data.js';
import {identifySingleImage,hasSingleImageCandidate} from '../../../site/emulators/knowledge-model.js';
import {memoryLabels} from '../../../site/emulators/memory-labels.js';
import {createMemoryService} from '../../../site/emulators/memory-worker.js';

const pkg=knowledgePackages.find(p=>p.id==='fort-apocalypse-c64');
assert(pkg);
const source=fs.readFileSync(new URL('../../../games/fort-apocalypse-c64/knowledge.json',import.meta.url));
assert.equal(pkg.sourceSHA256,crypto.createHash('sha256').update(source).digest('hex'));
assert.deepEqual(pkg.knowledge,JSON.parse(source));
const media=pkg.releases[0].media[0];
const match=identifySingleImage('c64',media);
assert.equal(match.status,'matched');assert.equal(match.labels.length,13);
const terrain=match.labels.find(l=>l.id==='terrain');
assert.equal(terrain.region,'ram');assert.equal(terrain.start,0x0503);assert.equal(terrain.length,0x2800);
assert.equal(match.labels.find(l=>l.id==='rle-symbol').start,0x8cdb);
assert.equal(match.labels.find(l=>l.id==='tape-pulses').start,20);
assert(match.labels.every(l=>l.description.includes('Applicability:')));
terrain.start=0;assert.equal(identifySingleImage('c64',media).labels.find(l=>l.id==='terrain').start,0x0503);
assert.equal(identifySingleImage('c64',{...media,sha256:'0'.repeat(64)}).status,'unknown');
assert.equal(identifySingleImage('c64',{...media,size:media.size-1}).status,'unknown');
assert.equal(identifySingleImage('gg',media).status,'unknown');
assert.equal(identifySingleImage('c64',media,[pkg,pkg]).status,'ambiguous');
assert.deepEqual(identifySingleImage('c64',media,[pkg,pkg]).labels,[]);
const multi={...pkg,releases:[{...pkg.releases[0],media:[media,{...media,role:'other'}]}]};
assert.equal(identifySingleImage('c64',media,[multi]).status,'unknown');
assert(hasSingleImageCandidate('c64',media.size));
// Unknown same-sized image must still be hashed and fail recognition.
const unknown=new File([new Uint8Array(media.size)],'unknown.tap');
assert.deepEqual(await memoryLabels('c64',unknown),[]);
assert.deepEqual(await memoryLabels('gg',unknown),[]);
// Recognition failure leaves generic physical snapshots functional.
const heap=Uint8Array.of(10,20,30,40);
const core={HEAPU8:heap,UTF8ToString:x=>x,
 _rr_inspect_regions:()=>JSON.stringify({activity:false,regions:[{id:'ram',name:'RAM',kind:'ram',index:0,size:4,base:0,aliases:[0]}]}),
 _rr_inspect_data:()=>0};
const service=createMemoryService({core,platform:'c64',files:[unknown],status:()=>({})});
await service.snapshot();assert.deepEqual(service.overview().labels,[]);assert.equal(service.overview().regions[0].name,'RAM');
assert.deepEqual([...service.page('ram',0).bytes],[10,20,30,40]);
// Optional private-media path: validates actual hashing through Memory's entry point.
if(process.argv[2]) {
 const bytes=fs.readFileSync(process.argv[2]);
 const file=new File([bytes],'reference.tap');
 const labels=await memoryLabels('c64',file);
 assert.equal(labels.length,13,'supplied tape must match the exact Fort reference');
 assert.equal(labels.find(l=>l.id==='enemy-mode'),undefined,'M1 does not invent live state UI');
 console.log('PASS actual Fort tape recognition through memoryLabels');
}
console.log('PASS knowledge: source integrity, label migration, exact identity, ambiguity, defensive copies, unknown-image generic Memory');
