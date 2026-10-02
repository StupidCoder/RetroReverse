import fs from 'node:fs';import {pathToFileURL} from 'node:url';import path from 'node:path';
import {benchmark} from './c64-performance.js';
const [current,legacy,tapePath]=process.argv.slice(2);
const root=new URL('../../../',import.meta.url),roms=['basic','kernal','chargen'].map(n=>fs.readFileSync(new URL('site/emulators/firmware/c64/'+n+'.rom',root))),tape=fs.readFileSync(tapePath),pkg=JSON.parse(fs.readFileSync(new URL('games/fort-apocalypse-c64/knowledge.json',root)));
for(const [label,module]of [['owned',current],['legacy',legacy]])console.log(JSON.stringify(await benchmark({factory:(await import(pathToFileURL(path.resolve(module)))).default,wasm:fs.readFileSync(path.join(path.dirname(module),'core.wasm')),roms,tape,pkg,label})));
