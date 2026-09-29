// A directory selection retains paths. Ordinary multi-file selections form a flat folder.
export function dosFiles(files){
 if(!files.length||files.length>10000)throw Error('Select a game folder containing at most 10,000 files');
 const folder=files.every(f=>f.webkitRelativePath&&f.webkitRelativePath.includes('/'));
 const seen=new Set();return files.map((file,index)=>{
  let path=(folder?file.webkitRelativePath.split('/').slice(1).join('/'):file.name).replaceAll('\\','/').toLowerCase();
  if(!path||path.length>4096||path.split('/').some(p=>!p||p==='..'||p==='.')||path.includes(':')||path.includes('\0')||seen.has(path))throw Error('Invalid or duplicate DOS game path: '+path);
  if(file.size>4*1024**3)throw Error('A game file exceeds 4 GiB');seen.add(path);return {file,path,index};
 });
}
export function selectDOSMedia(files,entry){const entries=dosFiles(files),programs=entries.filter(e=>/\.exe$/i.test(e.path));if(!programs.length)throw Error('The folder contains no DOS EXE');const chosen=entry?programs.find(e=>e.path===entry):programs.length===1?programs[0]:null;if(!chosen)throw Error('Choose the DOS executable to run');return {entries,entry:chosen.path};}
export const dosKeys={Escape:1,Digit1:2,Digit2:3,Digit3:4,Digit4:5,Digit5:6,Digit6:7,Digit7:8,Digit8:9,Digit9:10,Digit0:11,Minus:12,Equal:13,Backspace:14,Tab:15,KeyQ:16,KeyW:17,KeyE:18,KeyR:19,KeyT:20,KeyY:21,KeyU:22,KeyI:23,KeyO:24,KeyP:25,BracketLeft:26,BracketRight:27,Enter:28,ControlLeft:29,ControlRight:29,KeyA:30,KeyS:31,KeyD:32,KeyF:33,KeyG:34,KeyH:35,KeyJ:36,KeyK:37,KeyL:38,Semicolon:39,Quote:40,Backquote:41,ShiftLeft:42,Backslash:43,KeyZ:44,KeyX:45,KeyC:46,KeyV:47,KeyB:48,KeyN:49,KeyM:50,Comma:51,Period:52,Slash:53,ShiftRight:54,AltLeft:56,Space:57,CapsLock:58,F1:59,F2:60,F3:61,F4:62,F5:63,F6:64,F7:65,F8:66,F9:67,F10:68,Home:71,ArrowUp:72,PageUp:73,ArrowLeft:75,ArrowRight:77,End:79,ArrowDown:80,PageDown:81,Insert:82,Delete:83};
