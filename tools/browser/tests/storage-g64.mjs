import assert from 'node:assert/strict';
import {openStorage} from '../../../site/emulators/storage-media.js';
import {diskPoint,diskTracks} from '../../../site/emulators/disk-atlas.js';
const bytes=new Uint8Array(48),v=new DataView(bytes.buffer);bytes.set(new TextEncoder().encode('GCR-1541'));bytes[9]=3;v.setUint16(10,4,true);v.setUint32(12,36,true);v.setUint32(20,42,true);v.setUint16(36,4,true);v.setUint16(42,4,true);bytes.set([0,64,128,255],38);bytes.set([255,128,64,0],44);
const image=await openStorage(new File([bytes],'authored.g64'),'c64');assert.deepEqual(image.diskGeometry.counts,[4,0,4]);assert.equal(image.sectorCount,8);assert.equal(image.diskGeometry.trackStep,.5);assert.deepEqual([...await image.diskBytes()],[0,64,128,255,255,128,64,0]);assert.equal(image.describeSector(4),'Track 2 · encoded byte 0');assert.deepEqual([...await image.readSector(7)],[0]);
assert.equal(diskTracks(image.diskGeometry)[2].start,4);assert.equal(diskPoint(image.diskGeometry,0,256,106),null);
const bad=bytes.slice();new DataView(bad.buffer).setUint32(12,12,true);await assert.rejects(openStorage(new File([bad],'bad.g64'),'c64'),/overlaps/);
console.log('PASS G64 storage: bounded raw tracks, empty half-tracks, byte geometry and malformed extents');
