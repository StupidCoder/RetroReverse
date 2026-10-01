// Read-only, bounded views of local image bytes. Format offsets follow the
// repository's ISO 9660, AmigaDOS and Nintendo 3DS parsers; no core execution.
export const formatSize=n=>{const i=n>=1e9?3:n>=1e6?2:1;return (n/1000**i).toFixed(1)+' '+['B','KB','MB','GB'][i];};
const ascii=b=>new TextDecoder('latin1').decode(b).replace(/\0.*$/s,'').trimEnd();
const u32=(b,p,le=true)=>new DataView(b.buffer,b.byteOffset,b.byteLength).getUint32(p,le);
const u64=(b,p)=>{const v=new DataView(b.buffer,b.byteOffset,b.byteLength).getBigUint64(p,true);if(v>BigInt(Number.MAX_SAFE_INTEGER))throw Error('Image offset exceeds safe integer range');return Number(v);};
const range=(at,n,size)=>{if(!Number.isSafeInteger(at)||!Number.isSafeInteger(n)||at<0||n<0||at+n>size)throw Error('Image extent is out of bounds');};
const fileNode=(name,size,offset)=>({name,size,offset,type:'file'});
const directory=(name,children,offset=0)=>({name,type:'directory',offset,children});
const bounded=(seen,key)=>{if(seen.has(key)||seen.size>=4096)throw Error('Cyclic or oversized directory');seen.add(key);};
export async function openStorage(file,platform=''){
 const read=async(at,n)=>{range(at,n,file.size);if(n>2*1024*1024)throw Error('Metadata read exceeds 2 MB inspection limit');return new Uint8Array(await file.slice(at,at+n).arrayBuffer());};
 const head=await read(0,Math.min(file.size,512));
 const result={file,format:'Raw image',unit:'Block',sectorSize:256,sectorCount:Math.ceil(file.size/256),note:'Filesystem not recognized. Showing original image bytes in 256-byte blocks.',roots:[]};
 const geometry=(format,size,count=Math.ceil(file.size/size))=>Object.assign(result,{format,sectorSize:size,sectorCount:count,unit:'Sector',note:'Original selected image bytes; this view does not represent live drive reads or guest writes.'});
 result.describeSector=n=>`${result.unit} ${n}`;
 result.readSector=async n=>{if(!Number.isSafeInteger(n)||n<0||n>=result.sectorCount)throw Error('Sector is outside the image');return read(n*result.sectorSize,Math.min(result.sectorSize,file.size-n*result.sectorSize));};
 if(ascii(head.slice(0,12))==='C64-TAPE-RAW'){result.format='C64 TAP';result.tape=true;result.note='Tape pulse stream';return result;}
 if(platform==='c64'&&/\.d64$/i.test(file.name)){
  const sectors=tracks=>Array.from({length:tracks},(_,i)=>i<17?21:i<24?19:i<30?18:17),tracks=[35,40,42].find(t=>{const n=sectors(t).reduce((a,b)=>a+b,0);return file.size===n*256||file.size===n*257;});
  if(!tracks)throw Error('Unsupported D64 track geometry');const counts=sectors(tracks),total=counts.reduce((a,b)=>a+b,0),raw=await read(0,total*256);
  const index=(t,s)=>{if(t<1||t>tracks||s<0||s>=counts[t-1])throw Error('Invalid D64 track/sector link');return counts.slice(0,t-1).reduce((a,b)=>a+b,0)+s;};
  geometry('C64 D64',256,total);result.describeSector=n=>{let t=1;while(n>=counts[t-1])n-=counts[t++-1];return `Track ${t} · sector ${n}`;};
  const name=b=>Array.from(b).map(v=>v===160?' ':v>=32&&v<=126?String.fromCharCode(v):v>=193&&v<=218?String.fromCharCode(v-128):'�').join('').trimEnd();
  result.roots=[directory('Disk directory',async()=>{const entries=[],seen=new Set();let t=18,s=1;while(t){const sector=index(t,s);bounded(seen,sector);const b=raw.subarray(sector*256,sector*256+256);for(let p=0;p<256;p+=32){const type=b[p+2]&7;if(!type)continue;let ft=b[p+3],fs=b[p+4],size=0;const chain=new Set();while(ft){const at=index(ft,fs);bounded(chain,at);const next=raw.subarray(at*256,at*256+256);if(!next[0]&&next[1]<1)throw Error('Invalid D64 final sector length');size+=next[0]?254:next[1]-1;ft=next[0];fs=next[1];}entries.push({...fileNode(name(b.slice(p+5,p+21)),size,b[p+3]?index(b[p+3],b[p+4])*256:0),note:['DEL','SEQ','PRG','USR','REL'][type]||'Unknown type'});}t=b[0];s=b[1];}return entries;},index(18,1)*256)];return result;
 }
 if(platform==='amiga'&&/\.adf$/i.test(file.name)){
  if(![901120,1802240].includes(file.size))throw Error('Unsupported ADF geometry');geometry('Amiga ADF',512);const perTrack=file.size===901120?11:22;
  result.describeSector=n=>`Cylinder ${Math.floor(n/(perTrack*2))} · side ${Math.floor(n/perTrack)%2} · sector ${n%perTrack}`;
  if(ascii(head.slice(0,3))!=='DOS'){result.note+=' No AmigaDOS filesystem signature (custom loader disk).';return result;}
  const root=file.size/1024,block=async n=>{range(n*512,512,file.size);return read(n*512,512);},name=b=>{if(b[432]>30)throw Error('Invalid AmigaDOS name');return ascii(b.slice(433,433+b[432]));};
  const list=async(n,ancestors=new Set())=>{if(ancestors.has(n)||ancestors.size>=32)throw Error('Cyclic/deep AmigaDOS directory');const parents=new Set([...ancestors,n]),b=await block(n);if(u32(b,0,false)!==2)throw Error('Invalid AmigaDOS directory header');const out=[],seen=new Set();for(let i=0;i<72;i++){let next=u32(b,24+4*i,false);while(next){bounded(seen,next);const e=await block(next),type=u32(e,508,false),at=next;if(u32(e,0,false)!==2)throw Error('Invalid AmigaDOS entry header');if(type===2)out.push(directory(name(e),()=>list(at,parents),at*512));else if(type===0xfffffffd)out.push({...fileNode(name(e),u32(e,324,false),at*512),note:'File header block'});next=u32(e,496,false);}}return out;};
  const b=await block(root);if(u32(b,508,false)!==1)throw Error('Invalid AmigaDOS root');result.roots=[directory(name(b)||'AmigaDOS',()=>list(root),root*512)];return result;
 }
 // NCSD/NCCH partitions: metadata is read lazily; encrypted contents stay opaque.
 if(['NCSD','NCCH'].includes(ascii(head.slice(256,260)))){
  geometry('Nintendo 3DS',512);result.unit='Media block';
  const partition=async(base,size)=>{range(0,512,size);const h=await read(base,512);if(ascii(h.slice(256,260))!=='NCCH')throw Error('Partition is not NCCH');if(!(h[399]&4))return [{name:'Encrypted partition — filesystem unavailable',type:'note'}];if(h[398]>8)throw Error('Unsupported NCCH media unit');const unit=512*2**h[398],out=[];
   const region=at=>{const off=u32(h,at)*unit,n=u32(h,at+4)*unit;range(off,n,size);return [base+off,n];};
   const [exe,en]=region(416);if(en){out.push(directory('ExeFS',async()=>{if(en<512)throw Error('Truncated ExeFS');const b=await read(exe,512),out=[];for(let i=0;i<10;i++){const name=ascii(b.slice(i*16,i*16+8));if(!name)continue;const at=512+u32(b,i*16+8),size=u32(b,i*16+12);range(at,size,en);out.push(fileNode(name,size,exe+at));}return out;},exe));}
   const [rom,rn]=region(432);if(rn)out.push(directory('RomFS',async()=>romfs(rom,rn),rom));return out;};
  async function romfs(base,size){range(0,96,size);const iv=await read(base,96);if(ascii(iv.slice(0,4))!=='IVFC'||u32(iv,4)!==65536)throw Error('Invalid RomFS IVFC header');const exponent=u32(iv,76);if(exponent>20)throw Error('Unsupported IVFC block size');const unit=2**exponent,start=Math.ceil((96+u32(iv,8))/unit)*unit,levelSize=u64(iv,68);range(start,levelSize,size);range(0,40,levelSize);const h=await read(base+start,40);if(u32(h,0)!==40)throw Error('Invalid RomFS level-3 header');const dirs=u32(h,12),dirSize=u32(h,16),files=u32(h,28),fileSize=u32(h,32),data=u32(h,36);range(dirs,dirSize,levelSize);range(files,fileSize,levelSize);range(data,0,levelSize);
   const entry=async(off,isDir)=>{const header=isDir?24:32,table=isDir?dirs:files,limit=isDir?dirSize:fileSize;range(off,header,limit);const b=await read(base+start+table+off,header),len=u32(b,header-4);if(len>1024||len%2)throw Error('Invalid RomFS name length');range(off+header,len,limit);const text=new TextDecoder('utf-16le',{fatal:true}).decode(await read(base+start+table+off+header,len));return {b,name:text};};
   const walk=async(off,ancestors=new Set())=>{if(ancestors.has(off)||ancestors.size>=32)throw Error('Cyclic/deep RomFS directory');const parents=new Set([...ancestors,off]),{b}=await entry(off,true),out=[];for(const isDir of [true,false]){const seen=new Set();let p=u32(b,isDir?8:12);while(p!==0xffffffff){bounded(seen,p);const at=p,e=await entry(p,isDir);if(isDir)out.push(directory(e.name,()=>walk(at,parents),base+start+dirs+at));else{const rel=u64(e.b,8),n=u64(e.b,16);range(data+rel,n,levelSize);out.push(fileNode(e.name,n,base+start+data+rel));}p=u32(e.b,4);}}return out;};return walk(0);
  }
  if(ascii(head.slice(256,260))==='NCCH')result.roots=[directory('Partition',()=>partition(0,file.size))];else{if(head[398]>8)throw Error('Unsupported NCSD media unit');const unit=512*2**head[398];for(let i=0;i<8;i++){const off=u32(head,288+i*8)*unit,n=u32(head,292+i*8)*unit;if(n){range(off,n,file.size);result.roots.push(directory('Partition '+i,()=>partition(off,n),off));}}}return result;
 }
 for(const [stride,dataOffset]of [[2048,0],[2352,16],[2352,24],[2448,16],[2448,24],[2448,0],[2336,8],[2336,0]]){
  const at=16*stride+dataOffset;if(at+2048>file.size)continue;const pvd=await read(at,2048);if(pvd[0]!==1||ascii(pvd.slice(1,6))!=='CD001'||pvd[6]!==1)continue;
  if((pvd[128]|pvd[129]<<8)!==2048)throw Error('Unsupported ISO logical block size');const volumeSize=u32(pvd,80)*stride;range(0,volumeSize,file.size);geometry('ISO 9660',stride,volumeSize/stride);result.note+=` Logical data begins at byte ${dataOffset} in each ${stride}-byte sector.`;
  const extent=(lba,size)=>{range(lba*2048,size,result.sectorCount*2048);};
  const list=async(lba,size,parents=new Set())=>{extent(lba,size);if(size>1024*1024||parents.size>=32||parents.has(lba))throw Error('Oversized/cyclic ISO directory');const ancestors=new Set([...parents,lba]),out=[];for(let block=0;block<Math.ceil(size/2048);block++){const b=await read((lba+block)*stride+dataOffset,2048);for(let p=0;p<Math.min(2048,size-block*2048);){const n=b[p];if(!n)break;if(n<34||p+n>2048||p+n>size-block*2048||33+b[p+32]>n)throw Error('Malformed ISO directory record');const e=b.subarray(p,p+n);p+=n;if(e[32]===1&&(e[33]===0||e[33]===1))continue;if(e[1]||e[26]||e[27]||e[25]&128)throw Error('Extended, interleaved or multi-extent ISO entries are not supported');const start=u32(e,2),length=u32(e,10),name=ascii(e.slice(33,33+e[32])).replace(/;\d+$/,'');extent(start,length);out.push(e[25]&2?directory(name,()=>list(start,length,ancestors),start*stride):fileNode(name,length,start*stride));if(out.length>4096)throw Error('ISO directory exceeds entry limit');}}return out;};
  const rootLBA=u32(pvd,158),size=u32(pvd,166);extent(rootLBA,size);result.roots=[directory(ascii(pvd.slice(40,72))||'ISO 9660',()=>list(rootLBA,size),rootLBA*stride)];return result;
 }
 return result;
}
export function selectedFileTree(files){
 const root={children:new Map()};for(const f of files){const parts=(f.webkitRelativePath||f.name).replaceAll('\\','/').split('/').filter(Boolean);let d=root;for(const name of parts.slice(0,-1)){if(!d.children.has(name))d.children.set(name,{name,children:new Map()});d=d.children.get(name);if(!d.children)throw Error('Conflicting local paths');}const name=parts.at(-1);if(d.children.has(name))throw Error('Duplicate local path');d.children.set(name,{...fileNode(name,f.size,0),file:f});}
 const nodes=d=>[...d.children.values()].map(e=>e.children?directory(e.name,async()=>nodes(e)) :e);return nodes(root);
}
