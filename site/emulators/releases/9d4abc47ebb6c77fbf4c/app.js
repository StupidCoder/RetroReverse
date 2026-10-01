import {mountEmulatorSession} from './emulator-session.js';
import {platforms} from './platforms.js';
// File objects retain locally selected media; only compact checkpoints are copied.
const suspended=new Map();let active,switching=false;
const maxBytes=256*1024*1024;
function mount(platform){active=mountEmulatorSession(platform,{onSystem:switchSystem});document.title=platforms[platform].name+' — RetroReverse';}
async function switchSystem(platform){
 if(switching||!platforms[platform]||active.platform===platform)return;switching=true;
 try{
  const old=active.platform,saved=await active.suspend();
  const used=[...suspended].filter(([id])=>id!==old&&id!==platform).reduce((n,[,v])=>n+v.stateFile.size,0);
  if(saved&&used+saved.stateFile.size>maxBytes)throw Error('Suspended sessions exceed 256 MiB. Save your game to a file and reload the page to release suspended sessions.');
  if(saved)suspended.set(old,saved);const resume=suspended.get(platform);
  active.dispose();mount(platform);if(resume)active.restore(resume);
  // Keep direct system URLs shareable; embedded acceptance pages retain their URL.
  if(window===window.top)history.replaceState(null,'',new URL('../../'+platform+'/',import.meta.url));
 }catch(e){active.status(e.message);}finally{switching=false;}
}
mount(document.body.dataset.platform);
