// Async inspection operations keep ownership across await (hashing/replay).
// Tokens make completion of a cancelled predecessor harmless to its successor.
export function createExecutionGate(){
 let owner=null,sequence=0;
 return {
  get owner(){return owner?.kind??null;},
  cancel(kinds){if(owner&&kinds.includes(owner.kind))owner=null;},
  acquire(kind){if(owner)return null;return owner={kind,id:++sequence};},
  release(token){if(owner===token)owner=null;}
 };
}
