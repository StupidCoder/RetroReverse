/* Explicit 68000 state fields; no host pointers, padding or FPU state. */
#include "../vendor/musashi/m68kcpu.h"
unsigned rr_cpu_state(unsigned *words, int load) {
  unsigned i = 0, j;
#define FIELD(f)                                                               \
  do {                                                                         \
    if (load)                                                                  \
      m68ki_cpu.f = words[i++];                                                \
    else                                                                       \
      words[i++] = m68ki_cpu.f;                                                \
  } while (0)
  for (j = 0; j < 16; j++) {
    FIELD(dar[j]);
    FIELD(dar_save[j]);
  }
  for (j = 0; j < 7; j++)
    FIELD(sp[j]);
  FIELD(ppc);
  FIELD(pc);
  FIELD(ir);
  FIELD(t1_flag);
  FIELD(t0_flag);
  FIELD(s_flag);
  FIELD(m_flag);
  FIELD(x_flag);
  FIELD(n_flag);
  FIELD(not_z_flag);
  FIELD(v_flag);
  FIELD(c_flag);
  FIELD(int_mask);
  FIELD(int_level);
  FIELD(stopped);
  FIELD(pref_addr);
  FIELD(pref_data);
  FIELD(instr_mode);
  FIELD(run_mode);
  FIELD(reset_cycles);
  FIELD(virq_state);
  FIELD(nmi_pending);
#undef FIELD
  return i;
}
