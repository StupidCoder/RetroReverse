#pragma once
#include "../../../../browser/core/console-runtime.h"
struct LocalSource;
template<class T>void clear(Slice<T>s){std::fill(s.begin(),s.end(),T{});}
inline double go_math_Atan(double x){return std::atan(x);}inline double go_math_Atan2(double y,double x){return std::atan2(y,x);}inline double go_math_Round(double x){return std::round(x);}

#define RR_CAPTURE_WRITE_CAP 1048576
#define RR_CAPTURE_EVENT_CAP 131072
#include "../../../../browser/core/console-trace.h"
