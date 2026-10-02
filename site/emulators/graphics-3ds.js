import {create3DSPresentation} from './presentation-3ds.js';
import {fragmentWGSL} from './fragment-3ds.js';
import {rasterWGSL} from './raster-3ds.js';
// Immutable, versioned GX input packets. No private media is stored in this module.
export function decodeGraphicsStream(bytes){
 const data=bytes instanceof Uint8Array?bytes:new Uint8Array(bytes),v=new DataView(data.buffer,data.byteOffset,data.byteLength);
 if(data.length<8||v.getUint32(0,true)!==0x50475252||v.getUint32(4,true)!==1)throw Error('Unknown 3DS graphics stream');
 const records=[];let offset=8;
 while(offset<data.length){
  if(offset+32>data.length)throw Error('Truncated graphics record');
  const h=Array.from({length:8},(_,i)=>v.getUint32(offset+i*4,true)),[size,kind,np,ni,nb,ne,src,dst]=h;
  if(size<32||size%4||np>16384||size!==32+np*4+align(ni)+align(nb)+align(ne)||offset+size>data.length)throw Error('Invalid graphics record size');
  let at=offset+32;const params=Array.from({length:np},()=>{const n=v.getUint32(at,true);at+=4;return n;});
  const take=n=>{const b=data.slice(at,at+n);at+=align(n);return b;};
  records.push({kind,params,src,dst,input:take(ni),before:take(nb),expected:take(ne)});offset+=size;
 }
 return records;
}
const align=n=>Math.ceil(n/4)*4;
export function transferSupport(p){
 const a=p.params;if(![1,2,3,4,5,6].includes(p.kind))return 'Unsupported graphics operation';
 if(!(p.input instanceof Uint8Array)||!(p.before instanceof Uint8Array)||!p.before.length||p.before.length>16*1024*1024||p.input.length>16*1024*1024)return 'Invalid transfer buffers';
 if(!Array.isArray(a)||!a.every(n=>Number.isInteger(n)&&n>=0&&n<=0xffffffff))return 'Invalid transfer parameters';
 if(p.kind===1){if(a.length!==4||a[0]<2||a[0]>4||a[2]>p.before.length||a[2]%a[0])return 'Invalid fill';}
 if(p.kind===2){
  const [n,iw,ig,ow,og,overlap]=a;
  if(a.length!==6||!n||!iw||!ow||n%iw||n+Math.floor((n-1)/iw)*ig>p.input.length||n+Math.floor((n-1)/ow)*og>p.before.length)return 'Invalid copy';
  if(overlap)return 'Overlapping copy requires ordered reference writes';
 }
 if(p.kind===3){
  const [sw,sh,dw,dh,flags,overlap]=a,fmt=flags>>>12&7,bpp=fmt===0?4:fmt===1?3:2;
  if(a.length!==6||!sw||!sh||!dw||!dh||Math.max(sw,sh,dw,dh)>2048||fmt>4||(flags&0x03000702)||Math.ceil(sw/8)*8*Math.ceil(sh/8)*8*4>p.input.length||dw*dh*bpp>p.before.length)return 'Unsupported display transfer';
  if(overlap)return 'Overlapping display transfer requires reference execution';
 }
 if(p.kind===4){
  const [w,h,cfg,op,write,count]=a;
  if(a.length!==7+count*10||!count||count>1024||!w||!h||w>1024||h>1024||w%8||h%8||w*h*4!==p.before.length||!(cfg&1)||((cfg>>>4)&7)!==0||op>7)return 'Unsupported stencil draw';
  for(let i=6;i<a.length-1;i+=10)if(a[i]>a[i+1]||a[i+1]>w||a[i+2]>a[i+3]||a[i+3]>h||a.slice(i+4,i+10).some(x=>x>2048))return 'Invalid stencil coverage';
 }
 if(p.kind===5||p.kind===6){
  const [w,h,mask,,blend]=a,n=w*h;
  if(a.length!==(p.kind===5?36:52)||!w||!h||w>1024||h>1024||w%8||h%8||mask>15||p.before.length!==n*4||(p.kind===5&&p.input.length!==n*4+a[9]*20))return 'Invalid fragment packet';
  if((a[3]&256)&&((blend&7)>4||((blend>>>8)&7)>4||[16,20,24,28].some(s=>((blend>>>s)&15)>14)))return 'Unsupported blend';
  if(a.at(-1)||(a[34]!==0xffffffff&&a[34]>a[9])||(a[8]&~0xff71))return 'Invalid fragment metadata';
  const sources=[0,1,2,3,4,5,13,14,15],combines=[0,1,2,3,4,5,8,9];
  for(let i=10;i<34;i+=4){
   if(a[i]>>>24||a[i+1]>>>24||[0,8,16].some(s=>!sources.includes((a[i]>>>s)&15)||!sources.includes((a[i+1]>>>s)&15)||((a[i+1]>>>(s+4))&15)>7)||!combines.includes(a[i+2]&255)||!combines.includes((a[i+2]>>>8)&255)||(a[i+2]&~0x03330f0f))return 'Unsupported TEV stage';
  }
  if(p.kind===6)return rasterSupport(p);
  // Offline streams are untrusted too. Each link must identify a later record;
  // this bounds work and prevents cycles before submitting a compute dispatch.
  const bytes=p.input,words=new DataView(bytes.buffer,bytes.byteOffset,bytes.byteLength),end=bytes.length/4;
  const referenced=new Uint8Array(a[9]);
  const valid=(x,min)=>{if(x===0xffffffff)return true;if(x<min||x>=end||(x-n)%5)return false;const i=(x-n)/5;if(referenced[i])return false;referenced[i]=1;return true;};
  for(let i=0;i<n;i++)if(!valid(words.getUint32(i*4,true),n))return 'Invalid fragment head';
  for(let i=n;i<end;i+=5)if(!valid(words.getUint32(i*4,true),i+5))return 'Invalid fragment link';
 }
 return null;
}
// Validate coarse bins, ordered triangle references, bounded binary32 inputs and
// immutable decoded texture ranges before an untrusted replay reaches the GPU.
function rasterSupport(packet){
 const a=packet.params,[w,h]=a,[count,nx,ny,enabled]=a.slice(35,39),bytes=packet.input;
 if(!count||count>1024||nx!==Math.ceil(w/16)||ny!==Math.ceil(h/16)||enabled>7||bytes.length%4||a[9]>838860)return 'Invalid raster metadata';
 const data=new DataView(bytes.buffer,bytes.byteOffset,bytes.byteLength),words=bytes.length/4,header=nx*ny*2,triEnd=header+count*44;
 if(triEnd>words)return 'Truncated raster triangles';
 const get=i=>data.getUint32(i*4,true),float=i=>data.getFloat32(i*4,true);let work=0;
 for(let i=header;i<triEnd;i+=44){
  const x0=get(i),x1=get(i+1),y0=get(i+2),y1=get(i+3),area=float(i+4);
  if(x0>x1||x1>w||y0>y1||y1>h||!Number.isFinite(area)||area<2**-20||area>2**28)return 'Invalid raster bounds';work+=(x1-x0)*(y1-y0);
  for(let v=i+5;v<i+44;v+=13){
   for(let j=0;j<13;j++){const n=float(v+j),bound=j<2?4096:j===2?2**20:j<7?1e6:1024;if(!Number.isFinite(n)||Math.abs(n)>bound)return 'Invalid raster vertex';}
   const iw=float(v+2);if(iw<2**-20||iw>2**20)return 'Invalid raster reciprocal';
  }
 }
 if(work!==a[9])return 'Invalid raster workload';
 const references=new Uint32Array(count);let end=triEnd;
 for(let bin=0;bin<header;bin+=2){
  const start=get(bin),length=get(bin+1);if(start!==end||length>count||start+length>words)return 'Invalid raster bin';
  let previous=header-44;
  for(let i=start;i<start+length;i++){const t=get(i);if(t<=previous||t< header||t>=triEnd||(t-header)%44)return 'Invalid raster triangle link';previous=t;const bx=(bin/2)%nx,by=Math.floor(bin/2/nx);if(get(t)===get(t+1)||get(t+2)===get(t+3)||bx*16>=get(t+1)||(bx+1)*16<=get(t)||by*16>=get(t+3)||(by+1)*16<=get(t+2))return 'Raster bin misses triangle';references[(t-header)/44]++;}
  end=start+length;
 }
 for(let t=0;t<count;t++){const i=header+t*44;const expected=get(i)===get(i+1)||get(i+2)===get(i+3)?0:(Math.ceil(get(i+1)/16)-Math.floor(get(i)/16))*(Math.ceil(get(i+3)/16)-Math.floor(get(i+2)/16));if(references[t]!==expected)return 'Missing raster triangle bin';}
 for(let u=0;u<3;u++){
  const [offset,tw,th,wrap]=a.slice(39+u*4,43+u*4);
  if(!(enabled&(1<<u))){if(offset||tw||th||wrap)return 'Invalid disabled texture';continue;}
  if(offset!==end||tw>2047||th>2047||end+tw*th>words)return 'Invalid raster texture';end+=tw*th;
 }
 return end===words?null:'Trailing raster inputs';
}
export const transferWGSL=`
@group(0) @binding(0) var<storage,read> p:array<u32>;
@group(0) @binding(1) var<storage,read> src:array<u32>;
@group(0) @binding(2) var<storage,read> old:array<u32>;
@group(0) @binding(3) var<storage,read_write> dst:array<u32>;
fn source(i:u32)->u32{return (src[i/4u]>>((i%4u)*8u))&255u;}
fn previous(i:u32)->u32{return (old[i/4u]>>((i%4u)*8u))&255u;}
fn edge(a:vec2<i32>,b:vec2<i32>,v:vec2<i32>)->i32{return (b.x-a.x)*(v.y-a.y)-(b.y-a.y)*(v.x-a.x);}
fn stencil(i:u32)->u32{
 let w=p[4];let h=p[5];let tile=i/64u;let mo=i%64u;
 let x=(tile%(w/8u))*8u+(mo&1u)+((mo>>1u)&2u)+((mo>>2u)&4u);
 let ty=(tile/(w/8u))*8u+((mo>>1u)&1u)+((mo>>2u)&2u)+((mo>>3u)&4u);let y=h-1u-ty;
 let pos=vec2<i32>(i32(x*2u+1u),i32(y*2u+1u));var s=old[i]>>24u;let cfg=p[6];let op=p[7];let mask=(cfg>>8u)&255u;let reference=(cfg>>16u)&255u;
 for(var t=0u;t<p[9];t++){
  let b=10u+t*10u;if(x<p[b]||x>=p[b+1u]||y<p[b+2u]||y>=p[b+3u]){continue;}
  let a=vec2<i32>(i32(p[b+4u]),i32(p[b+5u]));let c=vec2<i32>(i32(p[b+6u]),i32(p[b+7u]));let d=vec2<i32>(i32(p[b+8u]),i32(p[b+9u]));
  if(edge(c,d,pos)<0||edge(d,a,pos)<0||edge(a,c,pos)<0){continue;}
  var n=s;switch(op){case 1u:{n=0u;}case 2u:{n=reference;}case 3u:{n=min(s+1u,255u);}case 4u:{n=select(s-1u,0u,s==0u);}case 5u:{n=s^255u;}case 6u:{n=(s+1u)&255u;}case 7u:{n=(s-1u)&255u;}default:{}}
  if(p[8]!=0u){s=(s&(~mask))|(n&mask);}
 }
 return (old[i]&0xffffffu)|(s<<24u);
}
fn byte(i:u32)->u32{
 if(i>=p[2]){return 0u;}
 if(p[0]==1u){if(i<p[6]){return (p[5]>>((i%p[4])*8u))&255u;}return previous(i);}
 if(p[0]==2u){
  let row=i/(p[7]+p[8]);let col=i%(p[7]+p[8]);let n=row*p[7]+col;
  if(col>=p[7]||n>=p[4]){return previous(i);}return source((n/p[5])*(p[5]+p[6])+n%p[5]);
 }
 let sw=p[4];let sh=p[5];let dw=p[6];let dh=p[7];let flags=p[8];let fmt=(flags>>12u)&7u;
 var bpp=2u;if(fmt==0u){bpp=4u;}else if(fmt==1u){bpp=3u;}
 let x=(i/bpp)%dw;var y=(i/bpp)/dw;let w=min(sw,dw);let h=min(sh,dh);
 if(x>=w||y>=h){return previous(i);}if((flags&1u)!=0u){y=h-1u-y;}
 let tile=(y/8u)*(sw/8u)+x/8u;
 let morton=(x&1u)|((y&1u)<<1u)|((x&2u)<<1u)|((y&2u)<<2u)|((x&4u)<<2u)|((y&4u)<<3u);
 let q=(tile*64u+morton)*4u;let a=source(q);let b=source(q+1u);let g=source(q+2u);let r=source(q+3u);
 if(fmt==0u){return source(q+i%4u);}if(fmt==1u){return source(q+1u+i%3u);}
 var v=0u;
 if(fmt==2u){v=((r>>3u)<<11u)|((g>>2u)<<5u)|(b>>3u);}
 if(fmt==3u){v=((r>>3u)<<11u)|((g>>3u)<<6u)|((b>>3u)<<1u)|(a>>7u);}
 if(fmt==4u){v=((r>>4u)<<12u)|((g>>4u)<<8u)|((b>>4u)<<4u)|(a>>4u);}
 return (v>>((i%2u)*8u))&255u;
}
${fragmentWGSL}
${rasterWGSL}
@compute @workgroup_size(64) fn main(@builtin(global_invocation_id) id:vec3<u32>){
 let index=id.x+id.y*4194240u;let i=index*4u;if(i>=p[2]){return;}
 if(p[0]==4u){dst[index]=stencil(index);return;}
 if(p[0]==5u){dst[index]=fragmentPixel(index);return;}
 if(p[0]==6u){dst[index]=rasterPixel(index);return;}
 dst[index]=byte(i)|(byte(i+1u)<<8u)|(byte(i+2u)<<16u)|(byte(i+3u)<<24u);
}`;

export async function create3DSGraphics({gpu=globalThis.navigator?.gpu,onLost=()=>{},measureGPU=true}={}){
 if(!gpu)throw Error('WebGPU is unavailable in this browser');
 const adapter=await gpu.requestAdapter();if(!adapter)throw Error('No WebGPU adapter is available');
 const timestamps=measureGPU&&adapter.features.has('timestamp-query');
 const device=await adapter.requestDevice({requiredFeatures:timestamps?['timestamp-query']:[]});let lost=null,busy=false,closed=false,allocation=null;
 const queries=timestamps?device.createQuerySet({type:'timestamp',count:2}):null;
 device.lost.then(info=>{lost=info.message||'WebGPU device lost';if(!closed)onLost(lost);});
 const warmStart=performance.now();device.pushErrorScope('validation');
 const module=device.createShaderModule({code:transferWGSL});
 let pipeline;
 try{pipeline=await device.createComputePipelineAsync({layout:'auto',compute:{module,entryPoint:'main'}});}
 catch(e){const info=await module.getCompilationInfo();device.destroy();throw Error(info.messages.map(m=>`${m.lineNum}:${m.linePos} ${m.message}`).join('\n')||e.message);}
 const error=await device.popErrorScope();if(error){device.destroy();throw Error(error.message);}
 const warmupMs=performance.now()-warmStart;
 let presentation=null;
 function release(){if(allocation)for(const b of allocation.buffers)b.destroy();allocation=null;}
 function buffers(n){
  if(allocation?.capacity>=n)return allocation;
  release();const capacity=2**Math.ceil(Math.log2(Math.max(n,256))),U=GPUBufferUsage;
  const b=[device.createBuffer({size:65536,usage:U.STORAGE|U.COPY_DST}),...Array.from({length:2},()=>device.createBuffer({size:capacity,usage:U.STORAGE|U.COPY_DST})),device.createBuffer({size:capacity,usage:U.STORAGE|U.COPY_SRC}),device.createBuffer({size:capacity,usage:U.COPY_DST|U.MAP_READ})];
  const counter=device.createBuffer({size:4,usage:U.STORAGE|U.COPY_DST|U.COPY_SRC});
  const bind=device.createBindGroup({layout:pipeline.getBindGroupLayout(0),entries:[...b.slice(0,4).map((buffer,binding)=>({binding,resource:{buffer}})),{binding:4,resource:{buffer:counter}}]});
  if(timestamps){b.push(device.createBuffer({size:16,usage:U.QUERY_RESOLVE|U.COPY_SRC}));b.push(device.createBuffer({size:16,usage:U.COPY_DST|U.MAP_READ}));}
  b.push(counter);return allocation={capacity,buffers:b,bind,counter};
 }
 const padded=b=>{if(b.length%4===0&&b.length)return b;const out=new Uint8Array(Math.max(4,align(b.length)));out.set(b);return out;};
 async function execute(packet){
  const reason=transferSupport(packet);if(reason)return {supported:false,reason};
  if(closed||lost)throw Error(lost||'Graphics backend closed');if(busy)throw Error('Graphics operation already in flight');busy=true;
  const start=performance.now();let scoped=false;
  try{
   device.pushErrorScope('validation');scoped=true;
   const a=buffers(Math.max(packet.input.length,align(packet.before.length)+4)),[params,src,old,out,read]=a.buffers;
   device.queue.writeBuffer(params,0,new Uint32Array([packet.kind,packet.input.length,packet.before.length,0,...packet.params]));
   device.queue.writeBuffer(src,0,padded(packet.input));device.queue.writeBuffer(old,0,padded(packet.before));
   const uploaded=performance.now(),encoder=device.createCommandEncoder();encoder.clearBuffer(a.counter);const pass=encoder.beginComputePass(timestamps?{timestampWrites:{querySet:queries,beginningOfPassWriteIndex:0,endOfPassWriteIndex:1}}:{});
   pass.setPipeline(pipeline);pass.setBindGroup(0,a.bind);const groups=Math.ceil(packet.before.length/256);pass.dispatchWorkgroups(Math.min(groups,65535),Math.ceil(groups/65535));pass.end();encoder.copyBufferToBuffer(out,0,read,0,align(packet.before.length));encoder.copyBufferToBuffer(a.counter,0,read,align(packet.before.length),4);
   if(timestamps){encoder.resolveQuerySet(queries,0,2,a.buffers[5],0);encoder.copyBufferToBuffer(a.buffers[5],0,a.buffers[6],0,16);}
   device.queue.submit([encoder.finish()]);const submitted=performance.now();
   await Promise.all([read.mapAsync(GPUMapMode.READ,0,align(packet.before.length)+4),...(timestamps?[a.buffers[6].mapAsync(GPUMapMode.READ)]:[])]);const mapped=performance.now();
   let gpuMs=null;if(timestamps){const t=new BigUint64Array(a.buffers[6].getMappedRange());gpuMs=Number(t[1]-t[0])/1e6;a.buffers[6].unmap();}
   const mappedBytes=read.getMappedRange(0,align(packet.before.length)+4),drawn=new DataView(mappedBytes).getUint32(align(packet.before.length),true);
   const result=new Uint8Array(mappedBytes).slice(0,packet.before.length);read.unmap();
   const err=await device.popErrorScope();scoped=false;if(err)throw Error(err.message);if(lost)throw Error(lost);
   if(packet.kind===6&&(drawn&0x80000000))return {supported:false,reason:'Raster arithmetic requires Reference'};
   return {supported:true,bytes:result,drawn,timing:{uploadMs:uploaded-start,submitMs:submitted-uploaded,gpuMs,queueAndMapMs:mapped-submitted,readbackMs:performance.now()-mapped,totalMs:performance.now()-start}};
  }finally{if(scoped)await device.popErrorScope();busy=false;}
 }
 return {execute,warmupMs,info:{vendor:adapter.info?.vendor,architecture:adapter.info?.architecture,timestamps},async preparePresentation(){presentation??=await create3DSPresentation(device);},present(core){if(closed||lost)return null;return presentation?.present(core)??null;},get bytesAllocated(){return (allocation?65536+allocation.capacity*4+4+(timestamps?32:0):0)+(presentation?.bytesAllocated??0);},get busy(){return busy;},destroy(){closed=true;presentation?.destroy();presentation=null;release();queries?.destroy();device.destroy();}};
}
