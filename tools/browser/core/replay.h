#pragma once
#include "capture.h"
#include <unordered_set>
// Scratch replay of captured effects. Every write contains its historical value;
// replay never re-reads live textures, calls HLE, or changes a guest clock.
namespace rrreplay {
struct Step {uint32_t end=0,event=0,pc=0;uint64_t clock=0;};
struct Replay {
 std::vector<Step> steps;
 std::vector<uint8_t> memory;
 std::vector<std::pair<uint32_t,std::vector<uint8_t>>> checkpoints;
 uint32_t cursor=0,writeCursor=0;size_t stride=1;
 void reset(){steps.clear();memory.clear();checkpoints.clear();cursor=writeCursor=0;}
 void begin(){
  reset();auto&t=rrcapture::trace;memory=t.initial;
  for(uint32_t i=0;i<t.writes.size();i++){
   const auto&w=t.writes[i];
   if(steps.empty()||w.event!=steps.back().event||(!w.event&&(w.pc!=steps.back().pc||w.clock!=steps.back().clock)))steps.push_back({i+1,w.event,w.pc,w.clock});
   else steps.back().end=i+1;
  }
  // Include commands which produce no writes, preserving submission order.
  std::unordered_set<uint32_t> seen;for(auto&s:steps)seen.insert(s.event);
  for(uint32_t i=0;i<t.events.size();i++){
   if(!seen.count(i+1))steps.push_back({t.events[i].sourceBefore,i+1,t.events[i].pc,t.events[i].clock});
  }
  std::stable_sort(steps.begin(),steps.end(),[](const Step&a,const Step&b){return a.end<b.end;});
  const size_t slots=std::max<size_t>(1,(16*1024*1024)/std::max<size_t>(1,memory.size()));
  stride=std::max<size_t>(1,(steps.size()+slots-1)/slots);
 }
 uint32_t forWrite(uint32_t id)const{if(!id)return 0;auto it=std::lower_bound(steps.begin(),steps.end(),id,[](const Step&s,uint32_t n){return s.end<n;});return it==steps.end()?steps.size():it-steps.begin()+1;}
 bool seek(uint32_t target,uint32_t budget=32768){
  target=std::min<uint32_t>(target,steps.size());auto&t=rrcapture::trace;
  uint32_t end=target?steps[target-1].end:0;
  if(end<writeCursor){memory=t.initial;writeCursor=cursor=0;for(auto&[at,b]:checkpoints)if(at<=target&&at>cursor){memory=b;cursor=at;writeCursor=steps[at-1].end;}}
  const uint32_t stop=std::min<uint32_t>(end,writeCursor+budget);
  while(writeCursor<stop){auto&w=t.writes[writeCursor++];if(w.flags&1)for(unsigned b=0;b<w.size;b++)memory[w.address-t.base+b]=w.after>>(8*b);}
  if(writeCursor<end)return false;
  cursor=target;
  if(target&&std::none_of(checkpoints.begin(),checkpoints.end(),[&](auto&x){return x.first==target;})){
   if((checkpoints.size()+1)*memory.size()>16*1024*1024&&!checkpoints.empty())checkpoints.erase(checkpoints.begin());
   if(memory.size()<=16*1024*1024)checkpoints.emplace_back(target,memory);
  }
  return true;
 }
 std::string info()const{
  auto&t=rrcapture::trace;std::ostringstream s;s<<"{\"count\":"<<steps.size()<<",\"cursor\":"<<cursor<<",\"cacheBytes\":"<<checkpoints.size()*memory.size()<<",\"complete\":"<<(!t.overflow?"true":"false");
  if(cursor){const auto&x=steps[cursor-1];s<<",\"clock\":"<<x.clock<<",\"pc\":"<<x.pc<<",\"command\":"<<(x.event?t.events[x.event-1].detail:"{\"kind\":\"CPU / HLE memory write\"}");}
  return s.str()+"}";
 }
};
inline Replay replay;
}
