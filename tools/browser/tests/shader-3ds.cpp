// Test-only WASM oracle: the actual generated interpreter and compiler use the
// same memory, input words, uniforms and browser floating-point implementation.
#include "../../platform/n3ds/browser/core/host.h"
#include <emscripten.h>
static n3ds_Machine m;
static arm_CPU cpu;
static n3ds_GPU gpu;
static rrvertex::Attributes input[16],reference[16],compiled[16];
EM_JS(void,exposeMemory,(),{Module.testMemory=wasmMemory;});
extern "C" {
uintptr_t shader_memory(int slot){exposeMemory();switch(slot){
 case 0:return uintptr_t(gpu.Code.data());case 1:return uintptr_t(gpu.Opdesc.data());
 case 2:return uintptr_t(input);case 3:return uintptr_t(reference);case 4:return uintptr_t(compiled);
 case 5:return uintptr_t(gpu.Float.data());case 6:return uintptr_t(gpu.Int.data());default:return 0;
}}
int shader_reference(uint32_t entry,uint32_t count,uint32_t booleans){
 if(entry>=4096||count>16)return 0;gpu.m=&m;m.CPU=&cpu;gpu.Bool=booleans;cpu.Halted=false;n3ds_GPU_invalidateShaders(&gpu);
 for(uint32_t i=0;i<count;i++)if(!n3ds_GPU_shaderRun(&gpu,&input[i],&reference[i],entry))return 0;
 return !cpu.Halted;
}
}
