#include "../../platform/threedo/browser/core/runtime.h"
#include <cassert>
int main(){
 auto a=Slice<int>{1,2,3,4,5,6};auto b=append(sub(a,0,2),sub(a,1,4));
 assert(b.n==5&&b[0]==1&&b[1]==2&&b[2]==2&&b[3]==3&&b[4]==4);
 auto c=Slice<int>{7,8};auto d=append(c,c);assert(d.n==4&&d[0]==7&&d[2]==7&&c[0]==7);
 auto e=Slice<std::string>{"a","b","c","d","e"};auto f=append(sub(e,0,2),sub(e,1,4));
 assert(f.n==5&&f[2]=="b"&&f[3]=="c"&&f[4]=="d");
}
