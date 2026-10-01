#pragma once
#include <cstdint>
// Existing browser execution/debug protocol, implemented by the owned core.
// Capture is observational and uses the owned VIC for live output and previews.
extern "C" {
uint8_t* rr_input();
int rr_init(int basic_size,int kernal_size,int chars_size);
int rr_tape(int size);
int rr_prepare(int pc,int pulse); // input: synthetic 64 KiB RAM; never authentic boot
int rr_run(int ticks,int stop_kind,int target);
int rr_stop_reason();
void rr_play(int down);
void rr_key(int code,int down);
void rr_joystick(int port,int mask);
int rr_checkpoint(int slot);
int rr_restore(int slot);
int rr_corrupt(int pulse,int duration); // edited current pulse restarts; invalidates identities
void rr_trace(int enabled,int pc_lo,int pc_hi);
void rr_trace_mask(int mask);
void rr_trace_reads(int enabled);
const char* rr_status();
const char* rr_events();
const char* rr_error();
const char* rr_profile();
uint8_t* rr_ram();
uint32_t rr_bus();
int rr_bus_flags();
uint32_t* rr_frame();
int rr_width();
int rr_height();
float* rr_audio();
int rr_audio_count();
double rr_cycle();
double rr_previous();
// Browser jobs: normalize, instruction, address, over, out; results retain the
// existing debugger protocol's interrupt-entry/mapping/return/watch distinctions.
int rr_debug_begin(int mode,int target,int next_match);
int rr_debug_run(int cycles);
int rr_debug_watch(int address,int value,int mask,int writer);
const char* rr_debug_event();
const char* rr_debug_snapshot(int address);
int rr_debug_edit(int count);
const char* rr_capabilities();
int rr_drive_rom(int size); // input: optional 16 KiB 1541 firmware; before running
int rr_disk(int size,int write_protected); // input: D64/G64; requires drive ROM
const char* rr_drive_status();
int rr_state_bind(); // input: seven SHA256 digests + u32le configuration (228 bytes)
uint8_t* rr_state_input(uint32_t size);
uint32_t rr_state_save();
uint8_t* rr_state_data();
int rr_state_load(uint32_t size);
const char* rr_inspect_regions();
const uint8_t* rr_inspect_data(uint32_t region);
void rr_activity_begin(uint32_t mask);
void rr_activity_end();
uint32_t rr_activity_count();
uint32_t rr_activity_dropped();
const uint8_t* rr_activity_data();
}

extern "C" {
int rr_capture_begin();int rr_capture_end();const char* rr_capture_info();
const char* rr_pixel(int x,int y);const char* rr_raster_info();const char* rr_raster_seek(int line);uint32_t* rr_raster_frame(int panel);const char* rr_raster_pixel(int panel,int x,int y);
const char* rr_replay_info();const char* rr_replay_begin();int rr_replay_seek(unsigned count);uint32_t* rr_replay_frame();uint32_t rr_replay_for_write(uint32_t y);
int rr_tileset_size();uint8_t* rr_tileset_data();const char* rr_memory(int space,int address);
}
extern "C" {
const char* rr_drive_debug_snapshot(int address);int rr_drive_debug_begin(int mode,int target,int next);int rr_drive_debug_run(int edges);double rr_drive_cycle();
}
