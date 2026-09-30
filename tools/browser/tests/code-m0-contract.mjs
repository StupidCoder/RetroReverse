// Executable specification vectors. This does not validate a production knowledge package.
import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
const v=JSON.parse(fs.readFileSync(new URL('./fixtures/code-m0/contracts.json',import.meta.url)));
const sha=b=>crypto.createHash('sha256').update(b).digest('hex');
function canonical(path,policy){
 assert(['exact','ascii-insensitive'].includes(policy));
 path=path.replaceAll('\\','/').normalize('NFC');
 assert(!/^[A-Za-z]:/.test(path)&&!/[\x00-\x1f\x7f]/.test(path));
 assert(path.split('/').every(s=>s&&s!=='.'&&s!=='..'));
 return policy==='ascii-insensitive'?path.replace(/[A-Z]/g,c=>c.toLowerCase()):path;
}
function manifest(members){
 const rows=members.map(m=>{const b=Buffer.from(m.hex,'hex');assert.equal(sha(b),m.sha256);return [canonical(m.path,'ascii-insensitive'),b.length,m.sha256];});
 rows.sort((a,b)=>Buffer.compare(Buffer.from(a[0]),Buffer.from(b[0])));
 assert.equal(new Set(rows.map(r=>r[0])).size,rows.length);
 return 'rr-media-set-v1\n'+JSON.stringify(rows);
}
for(const p of v.identity.validPaths)assert.equal(canonical(p.input,p.policy),p.canonical);
for(const p of v.identity.invalidPaths)assert.throws(()=>canonical(p,'exact'));
assert.equal(new Set(v.identity.collision.map(p=>canonical(p,'ascii-insensitive'))).size,1);
assert.throws(()=>manifest([v.identity.members[0],v.identity.members[0]]));
const serialized=manifest(v.identity.members);
assert.equal(serialized,manifest([...v.identity.members].reverse()));
const expected='rr-media-set-v1\n[["a.bin",3,"ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad"],["b.bin",0,"e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"]]';
assert.equal(serialized,expected);
const bytes=Buffer.from(v.state.hex,'hex'),slots=[];
for(let i=0;i<v.state.maxSlots&&bytes[v.state.typeOffset+i];i++){
 const address=bytes.readUInt16LE(v.state.pointerOffset+2*i);
 assert(address+v.state.recordBytes<=bytes.length);
 slots.push({index:i,type:bytes[i],address,value:bytes.readUInt16LE(address),flags:bytes[address+2]});
}
assert.deepEqual(slots,v.state.expectedSlots);
assert.equal(v.activity.initial!==v.activity.final,v.activity.endpointChanged);
assert.equal(v.activity.writes.length,v.activity.writeCount);
console.log(JSON.stringify({pass:true,fixture:'code-m0/contracts.json',manifest:serialized,manifestSHA256:sha(serialized),slots},null,2));
