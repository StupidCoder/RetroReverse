import assert from 'node:assert/strict';
import {mock} from 'node:test';
import {createInspectionFeed} from '../../../site/emulators/inspection-feed.js';
mock.timers.enable({apis:['setTimeout']});
let sequence=0,requests=[],seen=0;
const feed=createInspectionFeed({platform:'c64',send(type){requests.push(type);return ++sequence;}});
const client={active:true,overview(){seen++;}};
const unsubscribe=feed.subscribe(client);feed.ready();
// Continuous 60Hz frame and unrelated result traffic used to starve sampling.
for(let i=0;i<120;i++){
 feed.state({running:true,state:{cycle:i}});feed.result({type:'unrelated'});
 mock.timers.tick(16);
 if(requests.at(-1)==='memory-live-snapshot'){
  feed.result({type:'memory-overview',request:sequence,overview:{id:sequence}});requests.push('answered');
 }
}
assert(seen>=8,'live snapshots continue during frame traffic');
assert(seen<=10,'sampling remains throttled');
client.active=false;feed.visibility();const before=sequence;mock.timers.tick(1000);assert.equal(sequence,before);
client.active=true;feed.visibility();mock.timers.tick(200);assert.equal(sequence,before+1,'sampling resumes after hiding');
feed.reset();const reset=sequence;mock.timers.tick(1000);assert.equal(sequence,reset);
feed.ready();mock.timers.tick(200);assert.equal(sequence,reset+1,'sampling resumes after reset');
unsubscribe();const removed=sequence;mock.timers.tick(1000);assert.equal(sequence,removed);
feed.dispose();mock.timers.reset();
console.log('PASS inspection feed: continuous frame traffic, throttling, hide/resume, reset, unsubscribe');
