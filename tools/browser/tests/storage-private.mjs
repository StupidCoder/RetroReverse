// Private-media acceptance using bounded file reads, never a full-disc buffer.
import fs from 'node:fs/promises';import path from 'node:path';import assert from 'node:assert/strict';
import {openStorage} from '../../../site/emulators/storage-media.js';
const [platform,filename]=process.argv.slice(2),handle=await fs.open(filename,'r'),stat=await handle.stat();let totalRead=0,maxRead=0,entries=0,dirs=0;
try{const image={name:path.basename(filename),size:stat.size,slice:(a,b)=>({async arrayBuffer(){const bytes=new Uint8Array(b-a);const {bytesRead}=await handle.read(bytes,0,bytes.length,a);assert.equal(bytesRead,bytes.length);totalRead+=bytesRead;maxRead=Math.max(maxRead,bytesRead);return bytes.buffer;}})},volume=await openStorage(image,platform);
 async function visit(nodes,depth=0){assert(depth<=32);for(const n of nodes){assert(++entries<=20000);if(n.type==='directory'){dirs++;await visit(await n.children(),depth+1);}else if(n.type==='file')assert(Number.isSafeInteger(n.size)&&n.size>=0);}}
 await visit(volume.roots);assert.equal((await volume.readSector(volume.sectorCount-1)).length,Math.min(volume.sectorSize,stat.size-(volume.sectorCount-1)*volume.sectorSize));
 console.log(JSON.stringify({result:'PASS',platform,format:volume.format,imageBytes:stat.size,entries,directories:dirs,totalRead,maxRead,sectors:volume.sectorCount},null,2));
}finally{await handle.close();}
