// CUE parsing is deliberately strict: one Mode 1/2 data track, optional separate
// audio files. Audio playback is not implemented. Never guess a companion file.
export async function selectMedia(files) {
 const cues=files.filter(f=>/\.cue$/i.test(f.name));
 if(cues.length>1)throw Error('Select only one CUE sheet');
 if(!cues.length){if(files.length!==1)throw Error('Select one image, or its CUE and all companion tracks');return files[0];}
 let current,tracks=[],track;
 for(const raw of (await cues[0].text()).split(/\r?\n/)){
  const line=raw.trim();let m;
  if(m=/^FILE\s+"([^"]+)"\s+(\S+)$/i.exec(line)){
   if(m[2].toUpperCase()!=='BINARY')throw Error('Only BINARY CUE files are supported');
   const name=m[1].replaceAll('\\','/').split('/').pop();
   const matches=files.filter(f=>f.name.toLowerCase()===name.toLowerCase());
   if(matches.length!==1)throw Error('Missing or ambiguous CUE companion: '+name);
   current=matches[0];
  }else if(m=/^TRACK\s+(\d+)\s+(\S+)$/i.exec(line)){
   if(!current)throw Error('CUE track without FILE');
   track={file:current,mode:m[2].toUpperCase(),index:null};tracks.push(track);
  }else if(m=/^INDEX\s+01\s+(\d+):(\d+):(\d+)$/i.exec(line)){
   if(!track||+m[2]>=60||+m[3]>=75)throw Error('Invalid CUE index');
   track.index=(+m[1]*60 + +m[2])*75 + +m[3];
  }else if(/^PREGAP|^POSTGAP/i.test(line))throw Error('CUE PREGAP/POSTGAP is not supported; supply an extracted data track');
 }
 const data=tracks.filter(t=>t.mode!=='AUDIO');
 if(data.length!==1||!['MODE1/2048','MODE1/2352','MODE2/2352'].includes(data[0].mode))throw Error('CUE requires exactly one supported Mode 1/2 data track');
 const t=data[0];
 if(t.index===null)throw Error('Data track is missing INDEX 01');
 if(tracks.filter(x=>x.file===t.file).length!==1)throw Error('Combined data/audio BIN files are not supported; supply separate tracks');
 const stride=+t.mode.split('/')[1],offset=t.index*stride;
 if(offset>=t.file.size||(t.file.size-offset)%stride)throw Error('CUE data extent does not match sector geometry');
 return new File([t.file.slice(offset)],t.file.name);
}

// Incremental SHA-256: fixed 1 MiB file reads, no full-disc allocation.
const K=new Uint32Array([0x428a2f98,0x71374491,0xb5c0fbcf,0xe9b5dba5,0x3956c25b,0x59f111f1,0x923f82a4,0xab1c5ed5,0xd807aa98,0x12835b01,0x243185be,0x550c7dc3,0x72be5d74,0x80deb1fe,0x9bdc06a7,0xc19bf174,0xe49b69c1,0xefbe4786,0x0fc19dc6,0x240ca1cc,0x2de92c6f,0x4a7484aa,0x5cb0a9dc,0x76f988da,0x983e5152,0xa831c66d,0xb00327c8,0xbf597fc7,0xc6e00bf3,0xd5a79147,0x06ca6351,0x14292967,0x27b70a85,0x2e1b2138,0x4d2c6dfc,0x53380d13,0x650a7354,0x766a0abb,0x81c2c92e,0x92722c85,0xa2bfe8a1,0xa81a664b,0xc24b8b70,0xc76c51a3,0xd192e819,0xd6990624,0xf40e3585,0x106aa070,0x19a4c116,0x1e376c08,0x2748774c,0x34b0bcb5,0x391c0cb3,0x4ed8aa4a,0x5b9cca4f,0x682e6ff3,0x748f82ee,0x78a5636f,0x84c87814,0x8cc70208,0x90befffa,0xa4506ceb,0xbef9a3f7,0xc67178f2]);
const rot=(x,n)=>(x>>>n)|(x<<(32-n));
export async function sha256File(file,progress=()=>{}){
 const h=new Uint32Array([0x6a09e667,0xbb67ae85,0x3c6ef372,0xa54ff53a,0x510e527f,0x9b05688c,0x1f83d9ab,0x5be0cd19]),w=new Uint32Array(64);
 const compress=bytes=>{const v=new DataView(bytes.buffer,bytes.byteOffset,bytes.byteLength);for(let p=0;p<bytes.length;p+=64){for(let i=0;i<16;i++)w[i]=v.getUint32(p+i*4);for(let i=16;i<64;i++){const a=w[i-15],b=w[i-2];w[i]=w[i-16]+(rot(a,7)^rot(a,18)^(a>>>3))+w[i-7]+(rot(b,17)^rot(b,19)^(b>>>10));}let[a,b,c,d,e,f,g,j]=h;for(let i=0;i<64;i++){const t=(j+(rot(e,6)^rot(e,11)^rot(e,25))+((e&f)^(~e&g))+K[i]+w[i])|0,u=((rot(a,2)^rot(a,13)^rot(a,22))+((a&b)^(a&c)^(b&c)))|0;j=g;g=f;f=e;e=(d+t)|0;d=c;c=b;b=a;a=(t+u)|0;}[a,b,c,d,e,f,g,j].forEach((n,i)=>h[i]+=n);}};
 const end=file.size-file.size%64;
 for(let p=0;p<end;p+=1048576){compress(new Uint8Array(await file.slice(p,Math.min(end,p+1048576)).arrayBuffer()));progress(Math.min(end,p+1048576)/file.size);}
 const tail=new Uint8Array(file.size-end<56?64:128);tail.set(new Uint8Array(await file.slice(end).arrayBuffer()));tail[file.size-end]=128;new DataView(tail.buffer).setBigUint64(tail.length-8,BigInt(file.size)*8n);compress(tail);
 return [...h].map(n=>n.toString(16).padStart(8,'0')).join('');
}
