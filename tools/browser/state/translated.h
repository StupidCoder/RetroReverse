#pragma once
#include "archive.h"
template<class T> void stateFields(rrstate::Archive&a,Slice<T>&s){
 auto n=a.count(s.n);if(a.reading){a.charge(uint64_t(n)*sizeof(T));s=Slice<T>::make(n);}for(int64_t i=0;i<s.n;i++)a(s[i]);
}
template<class K,class V>void stateFields(rrstate::Archive&a,Map<K,V>&m){
 auto n=a.count(m.size());if(a.reading){a.charge(uint64_t(n)*(sizeof(K)+sizeof(V)+32));m=Map<K,V>();for(uint32_t i=0;i<n;i++){K k{};V v{};a(k,v);if(m.p->contains(k))throw std::runtime_error("Duplicate map key");(*m.p)[k]=v;}}
 else{std::vector<K> keys;for(auto&[k,v]:*m.p)keys.push_back(k);std::sort(keys.begin(),keys.end());for(auto k:keys)a(k,(*m.p)[k]);}
}
