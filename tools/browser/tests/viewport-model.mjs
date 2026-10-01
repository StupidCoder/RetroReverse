import assert from 'node:assert/strict';
import {leaf,split,leaves,replace,close,validLayout,preset,migrateLayouts} from '../../../site/emulators/viewport-model.js';
const types=['storage','session','game','atlas','hex','recording','render','render-sources','render-details','code','state','lesson','tape'];
for(const id of ['play','memory','render','code','loader','guided']){const p=preset(id);assert(validLayout(p,types));assert(leaves(p).length<=4);}
let tree=split('x',leaf('a','code'),split('y',leaf('b','atlas'),leaf('c','tape')));
tree=replace(tree,'b',leaf('b','hex'));assert.equal(leaves(tree)[1].kind,'hex');
tree=close(tree,'b');assert.deepEqual(leaves(tree).map(n=>n.id),['a','c']);assert.equal(tree.b.kind,'tape');
assert(!validLayout(split('x',leaf('a','code'),leaf('a','hex')),types));assert(!validLayout({...tree,ratio:NaN},types));assert(!validLayout(leaf('a','unknown'),types));
console.log('PASS viewport presets, independent identity, split replacement, merge and persisted-layout validation');

const legacy={loader:split('x',leaf('loader-code','code'),split('y',leaf('loader-ram','atlas'),leaf('loader-tape','tape')),.55),custom:split('x',leaf('mine','tape'),leaf('other','game'),.7)};
migrateLayouts(legacy);assert.deepEqual(legacy.loader,preset('loader'));assert.equal(legacy.custom.ratio,.7);assert.equal(legacy.custom.a.kind,'storage');assert.equal(legacy.custom.a.id,'mine');
