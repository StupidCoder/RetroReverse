// Layout contains presentation only. No machine addresses or execution state.
export const leaf=(id,kind)=>({id,kind});
export const split=(axis,a,b,ratio=.5)=>({axis,ratio,a,b});
export function leaves(node){return node.axis?[...leaves(node.a),...leaves(node.b)]:[node];}
export function replace(node,id,value){if(!node.axis)return node.id===id?value:node;return {...node,a:replace(node.a,id,value),b:replace(node.b,id,value)};}
export function close(node,id){if(!node.axis)return node.id===id?null:node;const a=close(node.a,id),b=close(node.b,id);return !a?b:!b?a:{...node,a,b};}
export function validLayout(n,types,ids=new Set(),depth=0){if(!n||depth>3)return false;if(n.axis)return ['x','y'].includes(n.axis)&&Number.isFinite(n.ratio)&&n.ratio>=.15&&n.ratio<=.85&&validLayout(n.a,types,ids,depth+1)&&validLayout(n.b,types,ids,depth+1)&&ids.size<=4;if(typeof n.id!=='string'||!/^[a-z0-9-]{1,80}$/.test(n.id)||ids.has(n.id)||!types.includes(n.kind))return false;ids.add(n.id);return ids.size<=4;}
export function preset(name,debug=true,tape=false){const l=(id,kind)=>leaf(name+'-'+id,kind);
 if(name==='play')return split('x',l('game','game'),l('session','session'),.7);
 if(name==='code')return split('x',split('y',l('game','game'),l('state','state')),split('y',l('code','code'),l('lesson','lesson'),.65));
 if(name==='loader')return split('x',l('code','code'),split('y',l('ram','atlas'),l('tape','tape')),.55);
 if(name==='memory')return split('x',split('y',l('game','game'),l('atlas','atlas')),split('y',l('hex','hex'),l('record','recording')));
 return split('x',l('render','render'),split('y',l('sources','render-sources'),l('details','render-details')),.62);
}
