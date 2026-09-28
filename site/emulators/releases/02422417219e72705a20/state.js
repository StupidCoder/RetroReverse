// Portable container: magic, JSON metadata length, metadata, core bytes, SHA-256.
// The digest covers both metadata and core bytes. Immutable media stays outside.
const magic = new TextEncoder().encode('RRSTATE1');
export const MAX_STATE = 129 * 1024 * 1024;
const hex = b => [...new Uint8Array(b)].map(x=>x.toString(16).padStart(2,'0')).join('');
export const digest = async b => hex(await crypto.subtle.digest('SHA-256', b));
export async function packState(metadata, payload) {
  const meta = new TextEncoder().encode(JSON.stringify(metadata));
  if(meta.length > 1024*1024 || payload.length > 128*1024*1024) throw Error('State exceeds size limit');
  const b = new Uint8Array(12+meta.length+payload.length+32);
  b.set(magic);new DataView(b.buffer).setUint32(8,meta.length,true);
  b.set(meta,12);b.set(payload,12+meta.length);
  b.set(new Uint8Array(await crypto.subtle.digest('SHA-256',b.subarray(0,-32))),b.length-32);
  return b;
}
export async function unpackState(file) {
  if(file.size<44||file.size>MAX_STATE)throw Error('Invalid state file size');
  const b=new Uint8Array(await file.arrayBuffer());
  if(!magic.every((v,i)=>b[i]===v))throw Error('Not a RetroReverse state file');
  const n=new DataView(b.buffer).getUint32(8,true);
  if(n>1024*1024||n>b.length-44)throw Error('Invalid state metadata length');
  const actual=await crypto.subtle.digest('SHA-256',b.subarray(0,-32));
  if(hex(actual)!==hex(b.subarray(-32)))throw Error('State integrity check failed');
  const meta=JSON.parse(new TextDecoder('utf-8',{fatal:true}).decode(b.subarray(12,12+n)));
  if(meta.format!==1||!['c64','ps1','n64','3do','ds','3ds','psp','gb','gg','amiga'].includes(meta.platform))throw Error('Unsupported state format');
  return {meta,payload:b.slice(12+n,-32)};
}
