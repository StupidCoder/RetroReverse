import {opcodes} from './opcodes6502.js';
export const hex=(n,width=4)=>n.toString(16).toUpperCase().padStart(width,'0');
export function decode6502(bytes,pc) {
 const op=opcodes[bytes[0]];
 if(bytes[0]==null)return {length:0,text:'Unavailable mapped memory',supported:false};
 if(!op)return {length:1,text:`.byte $${hex(bytes[0],2)} — undocumented / unsupported`,supported:false};
 if(bytes.length<op.length||bytes.slice(0,op.length).some(b=>b==null))return {length:0,text:'Incomplete mapped instruction',supported:false};
 const b=bytes[1],word=b|(bytes[2]<<8),v='$'+hex(b??0,2),w='$'+hex(word);
 const operands={imp:'',acc:'A',imm:'#'+v,zp:v,zpx:v+',X',zpy:v+',Y',izx:'('+v+',X)',izy:'('+v+'),Y',rel:'$'+hex((pc+2+(b<128?b:b-256))&65535),abs:w,abx:w+',X',aby:w+',Y',ind:'('+w+')'};
 return {length:op.length,text:op.mnemonic+(operands[op.mode]?' '+operands[op.mode]:''),supported:true};
}
export function disassemble6502(bytes,address,count=48){
 const rows=[];let offset=0;
 while(offset<bytes.length&&rows.length<count){const pc=(address+offset)&65535,op=decode6502(bytes.slice(offset,offset+3),pc);rows.push({address:pc,bytes:bytes.slice(offset,offset+Math.max(1,op.length)),...op});if(!op.supported)break;offset+=op.length;}
 return rows;
}
