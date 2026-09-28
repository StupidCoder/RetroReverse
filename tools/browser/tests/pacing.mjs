import assert from 'node:assert/strict';import {FrameClock} from '../../../site/emulators/pacing.js';
for(const hz of [50,60]){
 let wall=0,emu=0,clock=new FrameClock(wall,emu),presentations=[];
 for(let i=0;i<hz*10;i++){wall+=1;emu+=1/hz;presentations.push(wall);wall+=clock.delay(wall,emu,false);}
 assert(Math.abs(wall-10000)<.001);assert(presentations.slice(1).every((x,i)=>Math.abs(x-presentations[i]-1000/hz)<.001));
}
const c=new FrameClock(0,0);assert.equal(c.delay(10,1,true),0);assert.equal(c.delay(20,2,false),0);assert.equal(c.delay(1000,2.1,false),0);assert(c.delay(1001,2.12,false)<20);
console.log('50/60 Hz pacing, turbo transitions and bounded slow-frame catch-up passed');
