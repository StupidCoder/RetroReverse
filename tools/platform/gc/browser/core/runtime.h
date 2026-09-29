#pragma once
#include "../../../../browser/core/console-runtime.h"
struct gc_Machine;struct gc_workPool{};
struct gc_dspBus{gc_Machine*m{};explicit operator bool()const{return m!=nullptr;}};
inline constexpr int64_t gc_maxWorkers=1;

#define RR_CAPTURE_WRITE_CAP 8388608
#define RR_CAPTURE_EVENT_CAP 131072
#include "../../../../browser/core/console-trace.h"
