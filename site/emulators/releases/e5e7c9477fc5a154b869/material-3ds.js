// Compile a validated packet's fixed fragment program. Colors, texture/LUT
// contents, geometry and addresses remain runtime inputs, never cache identity.
function modes(packet){
 const fields=[[2,0xffffffff],[3,256],[4,0xffffffff],[5,15],[8,0x71]];
 if(packet.kind>=6)fields.push([38,7]);
 if(packet.kind>=7)fields.push([51,0x77]);
 if([8,10].includes(packet.kind))for(const i of [56,57,58,59,60,62])fields.push([i,0xffffffff]);
 if(packet.kind>=9){const at=packet.kind===10?84:56;fields.push([at,0x71],[at+1,0x777],[at+2,1]);}
 return fields.map(([i,mask])=>[i,mask,(packet.params[i]&mask)>>>0]);
}
export function materialKey(packet){
 const stages=[];for(let i=10;i<34;i+=4)stages.push(...packet.params.slice(i,i+3));
 return [packet.kind,...modes(packet).map(([, ,value])=>value),...stages].join('/');
}
function fetch(sel){return ['vertex','primary','secondary','tex[0]','tex[1]','tex[2]'][sel]||({13:'buf',14:'konst'}[sel])||'prev';}
function colorOperand(source,op){
 switch(op){
  case 1:return `vec3<i32>(255)-${source}.rgb`;
  case 2:return `vec3<i32>(${source}.a)`;case 3:return `vec3<i32>(255-${source}.a)`;
  case 4:return `vec3<i32>(${source}.r)`;case 5:return `vec3<i32>(255-${source}.r)`;
  case 8:return `vec3<i32>(${source}.g)`;case 9:return `vec3<i32>(255-${source}.g)`;
  case 12:return `vec3<i32>(${source}.b)`;case 13:return `vec3<i32>(255-${source}.b)`;
  default:return `${source}.rgb`;
 }
}
function alphaOperand(source,op){const value=`${source}.${['a','r','g','b'][op>>1]}`;return op&1?`255-${value}`:value;}
function combine(op,a,b,c){
 switch(op){case 0:return a;case 1:return `(${a}*${b})/255`;case 2:return `${a}+${b}`;case 3:return `${a}+${b}-128`;case 4:return `(${a}*${c}+${b}*(255-${c}))/255`;case 5:return `${a}-${b}`;case 8:return `(${a}*${b})/255+${c}`;default:return `(clamp(${a}+${b},0,255)*${c})/255`;}
}
export function specializeMaterialWGSL(source,packet){
 const lines=['fn litTev(vertex:vec4<i32>,primary:vec4<i32>,secondary:vec4<i32>,tex:array<vec4<i32>,3>)->vec4<i32>{','var prev=vertex;var buf=vec4<i32>(0);var next=rgba(p[11]);'];
 for(let stage=0;stage<6;stage++){
  const b=10+stage*4,[cs,as,ops]=packet.params.slice(b,b+3);
  lines.push('{',`let konst=rgba(p[${b+7}]);`);
  for(let j=0;j<3;j++){
   const color=cs>>>(j*8)&255,alpha=as>>>(j*8)&255;
   lines.push(`let c${j}=${colorOperand(fetch(color&15),color>>>4)};let a${j}=${alphaOperand(fetch(alpha&15),alpha>>>4)};`);
  }
  lines.push(`let combined=vec4<i32>(${['r','g','b'].map(c=>combine(ops&255,`c0.${c}`,`c1.${c}`,`c2.${c}`)).join(',')},${combine(ops>>>8&255,'a0','a1','a2')});`,
   `let value=clamp(combined<<vec4<u32>(vec3<u32>(${ops>>>16&3}u),${ops>>>20&3}u),vec4<i32>(0),vec4<i32>(255));`,
   'buf=next;');
  if(ops&0x1000000)lines.push('next=vec4<i32>(value.rgb,next.a);');
  if(ops&0x2000000)lines.push('next.a=value.a;');
  lines.push('prev=value;}');
 }
 lines.push('return prev;}');
 const start=source.indexOf('fn litTev('),end=source.indexOf('fn tev(',start);
 if(start<0||end<0)throw Error('Missing fragment specialization boundary');
 let code=source.slice(0,start)+lines.join('\n')+'\n'+source.slice(end);
 for(const [i,mask,value] of modes(packet)){
  const name=`p[${i+4}]`,remaining=(~mask)>>>0;
  code=code.replaceAll(name,remaining?`((${name}&${remaining}u)|${value}u)`:`${value}u`);
 }
 return code;
}
