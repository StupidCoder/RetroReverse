// Only local generated checkpoint bytes. Failure/private browsing falls back to regeneration.
const LIMIT=32*1024*1024,ITEM=4*1024*1024;
export async function checkpointCache(operation,key,bytes){
 if(!globalThis.indexedDB)return null;
 let db;
 try{
  db=await new Promise((resolve,reject)=>{const r=indexedDB.open('rr-prepared-v1',1);r.onupgradeneeded=()=>r.result.createObjectStore('states',{keyPath:'key'});r.onsuccess=()=>resolve(r.result);r.onerror=()=>reject(r.error);r.onblocked=()=>reject(Error('Cache blocked'));});
  return await new Promise((resolve,reject)=>{const tx=db.transaction('states','readwrite'),s=tx.objectStore('states');let answer=null;tx.oncomplete=()=>resolve(answer);tx.onerror=tx.onabort=()=>reject(tx.error);
   if(operation==='clear'){s.clear();return;}
   if(operation==='get'){const r=s.get(key);r.onsuccess=()=>{if(r.result){answer=r.result.bytes;r.result.used=Date.now();s.put(r.result);}};return;}
   if(!(bytes instanceof ArrayBuffer)||bytes.byteLength>ITEM)return;
   const r=s.getAll();r.onsuccess=()=>{let total=bytes.byteLength;const keep=r.result.filter(e=>e.key!==key).sort((a,b)=>b.used-a.used);for(const e of keep){total+=e.bytes.byteLength;if(total>LIMIT)s.delete(e.key);}s.put({key,bytes,used:Date.now()});};
  });
 }catch{return null;}finally{db?.close();}
}
