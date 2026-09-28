// Generated declaration-field traversal; pointer callbacks/maps are rebound.
#pragma once
inline void stateFields(rrstate::Archive&a,c64_t&v){
a(v.cpu.IR);
a(v.cpu.PC);
a(v.cpu.AD);
a(v.cpu.A);
a(v.cpu.X);
a(v.cpu.Y);
a(v.cpu.S);
a(v.cpu.P);
a(v.cpu.PINS);
a(v.cpu.irq_pip);
a(v.cpu.nmi_pip);
a(v.cpu.brk_flags);
a(v.cpu.bcd_enabled);
a(v.cpu.io_ddr);
a(v.cpu.io_inp);
a(v.cpu.io_out);
a(v.cpu.io_pins);
a(v.cpu.io_pullup);
a(v.cpu.io_floating);
a(v.cpu.io_drive);
a(v.cia_1.pa.reg);
a(v.cia_1.pa.ddr);
a(v.cia_1.pa.inp);
a(v.cia_1.pa.pins);
a(v.cia_1.pb.reg);
a(v.cia_1.pb.ddr);
a(v.cia_1.pb.inp);
a(v.cia_1.pb.pins);
a(v.cia_1.ta.latch);
a(v.cia_1.ta.counter);
a(v.cia_1.ta.cr);
a(v.cia_1.ta.t_bit);
a(v.cia_1.ta.t_out);
a(v.cia_1.ta.t_load);
a(v.cia_1.ta.pip);
a(v.cia_1.tb.latch);
a(v.cia_1.tb.counter);
a(v.cia_1.tb.cr);
a(v.cia_1.tb.t_bit);
a(v.cia_1.tb.t_out);
a(v.cia_1.tb.t_load);
a(v.cia_1.tb.pip);
a(v.cia_1.intr.imr);
a(v.cia_1.intr.imr1);
a(v.cia_1.intr.icr);
a(v.cia_1.intr.pip);
a(v.cia_1.intr.flag);
a(v.cia_1.intr.irq);
a(v.cia_1.intr.ir_now);
a(v.cia_1.intr.ir_block);
a(v.cia_1.pins);
a(v.cia_2.pa.reg);
a(v.cia_2.pa.ddr);
a(v.cia_2.pa.inp);
a(v.cia_2.pa.pins);
a(v.cia_2.pb.reg);
a(v.cia_2.pb.ddr);
a(v.cia_2.pb.inp);
a(v.cia_2.pb.pins);
a(v.cia_2.ta.latch);
a(v.cia_2.ta.counter);
a(v.cia_2.ta.cr);
a(v.cia_2.ta.t_bit);
a(v.cia_2.ta.t_out);
a(v.cia_2.ta.t_load);
a(v.cia_2.ta.pip);
a(v.cia_2.tb.latch);
a(v.cia_2.tb.counter);
a(v.cia_2.tb.cr);
a(v.cia_2.tb.t_bit);
a(v.cia_2.tb.t_out);
a(v.cia_2.tb.t_load);
a(v.cia_2.tb.pip);
a(v.cia_2.intr.imr);
a(v.cia_2.intr.imr1);
a(v.cia_2.intr.icr);
a(v.cia_2.intr.pip);
a(v.cia_2.intr.flag);
a(v.cia_2.intr.irq);
a(v.cia_2.intr.ir_now);
a(v.cia_2.intr.ir_block);
a(v.cia_2.pins);
a(v.vic.debug_vis);
for(size_t i0=0;i0<64;++i0){
a(v.vic.reg.regs[i0]);
}
for(size_t i0=0;i0<8;++i0){
for(size_t i1=0;i1<2;++i1){
a(v.vic.reg.mxy[i0][i1]);
}
}
a(v.vic.reg.mx8);
a(v.vic.reg.ctrl_1);
a(v.vic.reg.raster);
for(size_t i0=0;i0<2;++i0){
a(v.vic.reg.lightpen_xy[i0]);
}
a(v.vic.reg.me);
a(v.vic.reg.ctrl_2);
a(v.vic.reg.mye);
a(v.vic.reg.mem_ptrs);
a(v.vic.reg.int_latch);
a(v.vic.reg.int_mask);
a(v.vic.reg.mdp);
a(v.vic.reg.mmc);
a(v.vic.reg.mxe);
a(v.vic.reg.mcm);
a(v.vic.reg.mcd);
a(v.vic.reg.ec);
for(size_t i0=0;i0<4;++i0){
a(v.vic.reg.bc[i0]);
}
for(size_t i0=0;i0<2;++i0){
a(v.vic.reg.mm[i0]);
}
for(size_t i0=0;i0<8;++i0){
a(v.vic.reg.mc[i0]);
}
for(size_t i0=0;i0<17;++i0){
a(v.vic.reg.unused[i0]);
}
a(v.vic.crt.x);
a(v.vic.crt.y);
a(v.vic.crt.vis_x0);
a(v.vic.crt.vis_y0);
a(v.vic.crt.vis_x1);
a(v.vic.crt.vis_y1);
a(v.vic.crt.vis_w);
a(v.vic.crt.vis_h);
a(v.vic.brd.left);
a(v.vic.brd.right);
a(v.vic.brd.top);
a(v.vic.brd.bottom);
a(v.vic.brd.main);
a(v.vic.brd.vert);
a(v.vic.brd.bc);
a(v.vic.rs.h_count);
a(v.vic.rs.v_count);
a(v.vic.rs.v_irqline);
a(v.vic.rs.vc);
a(v.vic.rs.next_vc);
a(v.vic.rs.vc_base);
a(v.vic.rs.rc);
a(v.vic.rs.display_state);
a(v.vic.rs.badline);
a(v.vic.rs.frame_badlines_enabled);
a(v.vic.mem.c_addr_or);
a(v.vic.mem.g_addr_and);
a(v.vic.mem.g_addr_or);
a(v.vic.mem.i_addr);
a(v.vic.mem.p_addr_or);
a(v.vic.gunit.enabled);
a(v.vic.gunit.mode);
a(v.vic.gunit.count);
a(v.vic.gunit.shift);
a(v.vic.gunit.outp);
a(v.vic.gunit.outp2);
a(v.vic.gunit.c_data);
for(size_t i0=0;i0<4;++i0){
a(v.vic.gunit.bg[i0]);
}
a(v.vic.sunit.disp_enabled);
a(v.vic.sunit.expand);
a(v.vic.sunit.dma_enabled);
for(size_t i0=0;i0<8;++i0){
a(v.vic.sunit.h_first[i0]);
}
for(size_t i0=0;i0<8;++i0){
a(v.vic.sunit.h_last[i0]);
}
for(size_t i0=0;i0<8;++i0){
a(v.vic.sunit.h_offset[i0]);
}
for(size_t i0=0;i0<8;++i0){
a(v.vic.sunit.p_data[i0]);
}
for(size_t i0=0;i0<8;++i0){
a(v.vic.sunit.mc[i0]);
}
for(size_t i0=0;i0<8;++i0){
a(v.vic.sunit.mc_base[i0]);
}
for(size_t i0=0;i0<8;++i0){
a(v.vic.sunit.delay_count[i0]);
}
for(size_t i0=0;i0<8;++i0){
a(v.vic.sunit.outp2_count[i0]);
}
for(size_t i0=0;i0<8;++i0){
a(v.vic.sunit.xexp_count[i0]);
}
for(size_t i0=0;i0<8;++i0){
a(v.vic.sunit.shift[i0]);
}
for(size_t i0=0;i0<8;++i0){
a(v.vic.sunit.outp[i0]);
}
for(size_t i0=0;i0<8;++i0){
a(v.vic.sunit.outp2[i0]);
}
for(size_t i0=0;i0<8;++i0){
for(size_t i1=0;i1<4;++i1){
a(v.vic.sunit.colors[i0][i1]);
}
}
a(v.vic.vm.vmli);
a(v.vic.vm.next_vmli);
for(size_t i0=0;i0<64;++i0){
a(v.vic.vm.line[i0]);
}
a(v.vic.pins);
a(v.sid.sound_hz);
a(v.sid.bus_value);
a(v.sid.bus_decay);
for(size_t i0=0;i0<3;++i0){
a(v.sid.voice[i0].muted);
a(v.sid.voice[i0].freq);
a(v.sid.voice[i0].pulse_width);
a(v.sid.voice[i0].ctrl);
a(v.sid.voice[i0].sync);
a(v.sid.voice[i0].noise_shift);
a(v.sid.voice[i0].wav_accum);
a(v.sid.voice[i0].wav_output);
a(v.sid.voice[i0].env_state);
a(v.sid.voice[i0].env_attack_add);
a(v.sid.voice[i0].env_decay_sub);
a(v.sid.voice[i0].env_sustain_level);
a(v.sid.voice[i0].env_release_sub);
a(v.sid.voice[i0].env_cur_level);
a(v.sid.voice[i0].env_counter);
a(v.sid.voice[i0].env_exp_counter);
a(v.sid.voice[i0].env_counter_compare);
}
a(v.sid.filter.cutoff);
a(v.sid.filter.resonance);
a(v.sid.filter.voices);
a(v.sid.filter.mode);
a(v.sid.filter.volume);
a(v.sid.filter.nyquist_freq);
a(v.sid.filter.resonance_coeff_div_1024);
a(v.sid.filter.w0);
a(v.sid.filter.v_hp);
a(v.sid.filter.v_bp);
a(v.sid.filter.v_lp);
a(v.sid.sample_period);
a(v.sid.sample_counter);
a(v.sid.sample_accum);
a(v.sid.sample_accum_count);
a(v.sid.sample_mag);
a(v.sid.sample);
a(v.sid.pins);
a(v.pins);
a(v.joystick_type);
a(v.io_mapped);
a(v.cas_port);
a(v.iec_port);
a(v.cpu_port);
a(v.kbd_joy1_mask);
a(v.kbd_joy2_mask);
a(v.joy_joy1_mask);
a(v.joy_joy2_mask);
a(v.vic_bank_select);
a(v.kbd.cur_time);
a(v.kbd.sticky_time);
a(v.kbd.active_columns);
a(v.kbd.active_lines);
for(size_t i0=0;i0<256;++i0){
a(v.kbd.key_masks[i0]);
}
for(size_t i0=0;i0<4;++i0){
a(v.kbd.mod_masks[i0]);
}
for(size_t i0=0;i0<4;++i0){
a(v.kbd.key_buffer[i0].key);
a(v.kbd.key_buffer[i0].mask);
a(v.kbd.key_buffer[i0].pressed_time);
a(v.kbd.key_buffer[i0].released);
}
for(size_t i0=0;i0<12;++i0){
a(v.kbd.scanout_column_masks[i0]);
}
for(size_t i0=0;i0<12;++i0){
a(v.kbd.scanout_line_masks[i0]);
}
a(v.kbd.cur_column_mask);
a(v.kbd.cur_scanout_line_mask);
a(v.kbd.cur_line_mask);
a(v.kbd.cur_scanout_column_mask);
a(v.valid);
a(v.audio.num_samples);
a(v.audio.sample_pos);
for(size_t i0=0;i0<1024;++i0){
a(v.audio.sample_buffer[i0]);
}
for(size_t i0=0;i0<1024;++i0){
a(v.color_ram[i0]);
}
for(size_t i0=0;i0<65536;++i0){
a(v.ram[i0]);
}
for(size_t i0=0;i0<157248;++i0){
a(v.fb[i0]);
}
}
