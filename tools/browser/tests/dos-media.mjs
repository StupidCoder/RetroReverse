import assert from 'node:assert/strict';import {dosFiles,selectDOSMedia,dosKeys} from '../../../site/emulators/dos-media.js';
const file=(name,path='Game/'+name,size=10)=>({name,webkitRelativePath:path,size});
const files=[file('QUAKE.EXE'),file('PAK0.PAK','Game/ID1/PAK0.PAK')];
assert.equal(selectDOSMedia(files).entry,'quake.exe');assert.equal(dosFiles(files)[1].path,'id1/pak0.pak');
assert.throws(()=>selectDOSMedia([...files,file('SETUP.EXE')]),/Choose/);assert.equal(selectDOSMedia([...files,file('SETUP.EXE')],'setup.exe').entry,'setup.exe');
for(const bad of [[file('x.exe','Game/../x.exe')],[file('x.exe'),file('X.EXE')],[file('x.exe','Game/c:x.exe')],[file('x.exe','Game/x.exe',5*1024**3)]])assert.throws(()=>dosFiles(bad));
assert.equal(dosKeys.ArrowUp,0x48);assert.equal(dosKeys.Enter,0x1c);console.log('DOS directory paths, entry selection and input map pass');
