import assert from 'node:assert/strict';
import {openStorage,formatSize,selectedFileTree} from '../../../site/emulators/storage-media.js';
const file=(b,name)=>new File([b],name),set=(b,p,n,le=true)=>new DataView(b.buffer).setUint32(p,n,le),txt=(b,p,s)=>b.set(new TextEncoder().encode(s),p);
assert.equal(formatSize(1234),'1.2 KB');assert.equal(formatSize(1234567),'1.2 MB');assert.equal(formatSize(1234567890),'1.2 GB');
// D64 directory + two-sector file: exact payload size excludes link bytes.
const d64=new Uint8Array(174848),directorySector=358;d64[directorySector*256+2]=130;d64[directorySector*256+3]=1;txt(d64,directorySector*256+5,'HELLO');d64[0]=1;d64[1]=1;d64[256]=0;d64[257]=11;
let v=await openStorage(file(d64,'test.d64'),'c64');assert.equal(v.sectorCount,683);assert.equal(v.describeSector(358),'Track 18 · sector 1');assert.equal((await v.roots[0].children())[0].size,264);
d64[256]=1;d64[257]=0;v=await openStorage(file(d64,'cyclic.d64'),'c64');await assert.rejects(v.roots[0].children(),/Cyclic/);
// AmigaDOS root hash chain with a file and child folder.
const adf=new Uint8Array(901120);txt(adf,0,'DOS');const header=(block,type,name)=>{const at=block*512;set(adf,at,2,false);set(adf,at+508,type,false);adf[at+432]=name.length;txt(adf,at+433,name);};header(880,1,'TEST');header(900,0xfffffffd,'one');header(901,2,'folder');set(adf,880*512+24,900,false);set(adf,900*512+496,901,false);set(adf,900*512+324,1234,false);
v=await openStorage(file(adf,'test.adf'),'amiga');let entries=await v.roots[0].children();assert.equal(entries[0].size,1234);assert.equal(entries[1].type,'directory');assert.equal((await entries[1].children()).length,0);assert.equal(v.describeSector(22),'Cylinder 1 · side 0 · sector 0');
set(adf,901*512+496,900,false);v=await openStorage(file(adf,'cycle.adf'),'amiga');await assert.rejects(v.roots[0].children(),/Cyclic/);
// ISO logical directory records, cooked and Mode 1 raw layouts.
const iso=new Uint8Array(24*2048),pvd=16*2048;iso[pvd]=1;txt(iso,pvd+1,'CD001');iso[pvd+6]=1;iso[pvd+128]=0;iso[pvd+129]=8;set(iso,pvd+80,24);txt(iso,pvd+40,'TEST');
function record(at,name,lba,size,dir=false){const n=34+name.length;iso[at]=n;set(iso,at+2,lba);set(iso,at+10,size);iso[at+25]=dir?2:0;iso[at+32]=name.length;txt(iso,at+33,name);return n;}
record(pvd+156,'\0',20,2048,true);record(20*2048,'HELLO.BIN;1',21,1234);
for(const stride of [2048,2352]){const b=new Uint8Array(24*stride);for(let i=0;i<24;i++)b.set(iso.slice(i*2048,(i+1)*2048),i*stride+(stride===2352?16:0));v=await openStorage(file(b,'disc.bin'));assert.equal(v.format,'ISO 9660');entries=await v.roots[0].children();assert.equal(entries[0].name,'HELLO.BIN');assert.equal(entries[0].size,1234);assert.equal(entries[0].offset,21*stride);assert.equal((await v.readSector(20)).length,stride);}
// Minimal decrypted NCCH with RomFS root, folder and file. No full-ROM reads.
const rom=new Uint8Array(0x5000);txt(rom,256,'NCCH');rom[399]=4;set(rom,432,8);set(rom,436,32);txt(rom,0x1000,'IVFC');set(rom,0x1004,65536);set(rom,0x1008,32);set(rom,0x1044,0x3000);set(rom,0x104c,12);
const l3=0x2000;set(rom,l3,40);set(rom,l3+12,40);set(rom,l3+16,64);set(rom,l3+28,104);set(rom,l3+32,48);set(rom,l3+36,160);
set(rom,l3+40+8,24);set(rom,l3+40+12,0xffffffff);const dir=l3+64;set(rom,dir+4,0xffffffff);set(rom,dir+8,0xffffffff);set(rom,dir+12,0);set(rom,dir+20,6);rom.set([100,0,105,0,114,0],dir+24);
const f=l3+104;set(rom,f+4,0xffffffff);set(rom,f+16,1234);set(rom,f+28,8);rom.set([116,0,101,0,115,0,116,0],f+32);
v=await openStorage(file(rom,'test.cxi'),'3ds');let roots=await v.roots[0].children();let dirs=await roots[0].children();assert.equal(dirs[0].name,'dir');entries=await dirs[0].children();assert.equal(entries[0].name,'test');assert.equal(entries[0].size,1234);
rom[399]=0;v=await openStorage(file(rom,'encrypted.cxi'),'3ds');assert.match((await v.roots[0].children())[0].name,/Encrypted/);
const local=file(new Uint8Array(12),'file');Object.defineProperty(local,'webkitRelativePath',{value:'folder/nested/file'});const tree=selectedFileTree([local]);assert.equal((await(await tree[0].children())[0].children())[0].size,12);
console.log('PASS Storage: decimal sizes, D64 sectors/chains, AmigaDOS hierarchy, cooked/raw ISO files, decrypted/encrypted 3DS RomFS, local folders and corrupt cycles');

// Circular disk mapping uses track-specific sector counts and alternating ADF sides.
const {diskTracks,diskPoint}=await import('../../../site/emulators/disk-atlas.js');
const d64Geometry={cylinders:35,sides:1,counts:Array.from({length:35},(_,i)=>i<17?21:i<24?19:i<30?18:17)};
assert.equal(diskTracks(d64Geometry).at(-1).start,666);
assert.equal(diskPoint(d64Geometry,0,256,11).sector,0);
assert.equal(diskPoint(d64Geometry,0,256,256),null);
assert.equal(diskPoint(d64Geometry,0,256,0),null);
const adfGeometry={cylinders:80,sides:2,counts:Array(160).fill(11)};
assert.equal(diskPoint(adfGeometry,1,256,11).sector,11);
assert.equal(diskTracks(adfGeometry).at(-1).start,1749);
console.log('PASS circular geometry: D64 variable sectors, ADF side ordering, outer track and hub bounds');
