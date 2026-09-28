#pragma once
#include <stdint.h>
extern "C" {
uint8_t* rr_input();
int rr_init(int basic_size,int kernal_size,int chars_size); // input: BASIC,KERNAL,chars
int rr_tape(int size);
int rr_prepare(int pc,int pulse); // input: 64KiB prepared RAM, explicit synthetic start
int rr_stop_reason();
int rr_run(int ticks,int stop_kind,int target); // 0 budget; 1 edge; 2 PC; 3 address; 4 writer; 5 pulse; 6 next opcode; 7 raster frame
void rr_play(int down);
void rr_key(int code,int down);
void rr_joystick(int port,int mask);
int rr_checkpoint(int slot);
int rr_restore(int slot);
void rr_trace(int enabled,int pc_lo,int pc_hi);
const char* rr_status();
const char* rr_events();
const char* rr_error();
uint8_t* rr_ram();
uint32_t rr_bus();
int rr_bus_flags();
uint32_t* rr_frame();
int rr_width(); int rr_height();
float* rr_audio(); int rr_audio_count();
void rr_trace_mask(int mask);
void rr_trace_reads(int enabled);
double rr_cycle();
double rr_previous();
const char* rr_memory(int space,int address);
void rr_capture_begin();
void rr_capture_end();
const char* rr_pixel(int x,int y);
int rr_corrupt(int pulse,int duration);
}
