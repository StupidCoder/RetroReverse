export const hex=(v,n=4)=>'$'+Number(v).toString(16).toUpperCase().padStart(n,'0');
export const EVENT_WORDS=9;
export function parseAddress(text,regions,current){
 const parts=text.trim().split(':');
 let region=current,raw=parts.at(-1).replace(/^\$|^0x/i,'');
 if(!/^[\da-f]+$/i.test(raw))return null;
 const value=parseInt(raw,16);
 if(parts.length===2){region=regions.find(r=>r.id===parts[0]);return region&&value<region.size?{region,offset:value}:null;}
 if(parts.length!==1)return null;
 // Plain addresses belong to the selected physical segment; a bank ID removes ambiguity.
 const base=region?.kind==='rom'?region.base:(region?.aliases?.[0]??region?.base??0);
 if(region&&value>=base&&value<base+region.size)return {region,offset:value-base};
 return null;
}
export function pixelRange(pixel,bytesPerPixel,size){const start=pixel*bytesPerPixel;return start>=0&&start<size?{start,end:Math.min(start+bytesPerPixel,size)}:null;}
export function summarize(bytes,scale){const out=new Uint8Array(Math.ceil(bytes.length/scale));for(let i=0;i<out.length;i++){let sum=0,end=Math.min(bytes.length,(i+1)*scale);for(let j=i*scale;j<end;j++)sum+=bytes[j];out[i]=Math.round(sum/(end-i*scale));}return out;}
export function applyEvents(regions,events,from,to){
 for(let i=from;i<to;i++){const j=i*EVENT_WORDS,r=regions[events[j+2]];if(!r||!(events[j+6]&2))continue;
  for(let b=0;b<events[j+5]&&b<4;b++){const at=events[j+3]+b;if(at<r.bytes.length)r.bytes[at]=events[j+4]>>>(b*8)&255;}
 }
}
export function tapIndex(bytes){
 const offsets=[],durations=[];
 for(let i=20;i<bytes.length;){offsets.push(i);let n=bytes[i++]*8;if(!n){if(bytes[12]===0)n=2048;else{if(i+3>bytes.length)throw Error('Truncated TAP pulse');n=bytes[i]|bytes[i+1]<<8|bytes[i+2]<<16;i+=3;}}durations.push(n);}
 return {offsets:Uint32Array.from(offsets),durations:Uint32Array.from(durations)};
}
