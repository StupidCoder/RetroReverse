import assert from 'node:assert/strict';
import {shellMarkup} from '../../../site/emulators/ui-shell.js';
import {platforms} from '../../../site/emulators/platforms.js';
import {presentation} from '../../../site/emulators/ui-platforms.js';
assert.deepEqual(Object.keys(presentation).sort(),Object.keys(platforms).sort());
for(const platform of Object.keys(platforms)){
 const html=shellMarkup(platform),ids=[...html.matchAll(/id="([^"]+)"/g)].map(m=>m[1]);
 assert.equal(new Set(ids).size,ids.length,platform+' duplicate IDs');
 for(const id of ['screen','files','load','run','pause','reset','step','turbo','fullscreen','save','statefile','profile','cancelcapture','play-workspace','render-workspace','render-output-slot','render-timeline','render-auxiliary','render-details'])assert(ids.includes(id),platform+': '+id);
 for(const [p,extras] of Object.entries({c64:['basic','kernal','chargen'],amiga:['kickstart','mouse-speed'],ps2:['bios'],dos:['program']}))for(const id of extras)assert.equal(ids.includes(id),p===platform,platform+': '+id);
 assert.equal(ids.includes('compatprofile'),['3do','dos'].includes(platform));
 assert.equal(html.includes('webkitdirectory'),platform==='dos');
 assert.equal(html.includes('legacy.html'),platform==='c64');
 assert.equal(ids.includes('execution-mode'),platform==='3ds');
 if(platform==='3ds'){assert(ids.includes('execution-note'));assert.match(html,/<option value="experimental" disabled>/);}
 assert.match(presentation[platform].aspect,/^\d+\/\d+$/);
}
console.log('Shared shell covers all sixteen systems and preserves media/firmware controls');
