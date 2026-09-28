#include "../core/capture.h"
#include <cassert>
int main(){using namespace rrcapture;uint8_t b[8]={1,2,3,4,5,6,7,8};Recorder t;t.begin(b,8);auto e=t.event(12,0x1000,"{\"kind\":\"synthetic\"}");t.record(0,0x908,2,13,0x1000,e);b[0]=8;b[1]=9;t.record(0,0,2,14,0x1000,e,2);auto id=t.resource(b,8);assert(id==t.resource(b,8));t.end(b,8);auto s=t.pixel(0,2);assert(s.find("\"complete\":true")!=s.npos);assert(s.find("\"depthRejected\":true")!=s.npos);assert(t.pixel(2,2).find("\"contributors\":[]")!=s.npos);
 t.begin(b,8);b[0]=99;t.end(b,8);assert(t.pixel(0,1).find("\"complete\":false")!=s.npos);
 t.begin(b,8);for(size_t i=0;i<Recorder::WRITE_CAP+1;i++)t.record(0,99,1,i,0);t.end(b,8);assert(t.overflow==1);assert(t.writes.size()==Recorder::WRITE_CAP);
}
