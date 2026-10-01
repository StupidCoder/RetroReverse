// Explicit development opt-in. Production URLs and checkpoint metadata remain
// unchanged unless the owned C64 backend is selected.
export function backendAssets(platform,backend='production'){
 if(backend!=='production'&&(platform!=='c64'||backend!=='owned'))throw Error('Unsupported emulator backend');
 const owned=backend==='owned',directory=owned?'c64-owned':platform;
 return {owned,backend,directory,module:`./cores/${directory}/${owned?'core.mjs':'core.js'}`,wasm:`./cores/${directory}/core.wasm`,manifest:owned?'./cores/c64-owned/manifest.json':'./build-manifest.json',key:directory+'/core.wasm'};
}
export function selectedC64Backend(search){
 const value=new URLSearchParams(search).get('c64Core')||'production';backendAssets('c64',value);return value;
}
export function bindOwnedIdentity(core,{core:build,firmware,tape}){
 if(!core._rr_state_bind)throw Error('Owned checkpoint identity API is missing');
 if(firmware.length!==3)throw Error('Owned C64 needs three firmware identities');
 const bytes=new Uint8Array(228);
 [build,...firmware,null,tape,null].forEach((hash,index)=>{
  if(hash===null)return;
  if(typeof hash!=='string'||!/^[0-9a-f]{64}$/.test(hash))throw Error('Invalid owned checkpoint identity');
  for(let i=0;i<32;i++)bytes[index*32+i]=parseInt(hash.slice(i*2,i*2+2),16);
 });
 // This integration enables standalone TAP sessions. Drive configuration bits
 // stay zero until the worker also owns a verified 1541 firmware/disk pair.
 core.HEAPU8.set(bytes,core._rr_input());
 if(!core._rr_state_bind())throw Error('Owned checkpoint identity binding failed');
}
export function matchingPreparedKnowledge(knowledge,identity){
 if(!knowledge?.data?.preparedStarts)return knowledge;
 const preparedStarts=Object.fromEntries(Object.entries(knowledge.data.preparedStarts).filter(([,r])=>r.core===identity.core&&JSON.stringify(r.firmware)===JSON.stringify(identity.firmware)));
 return {...knowledge,data:{...knowledge.data,preparedStarts}};
}
