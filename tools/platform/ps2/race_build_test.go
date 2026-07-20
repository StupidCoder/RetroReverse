//go:build race

package ps2

// raceBuild is true when the race detector is compiled in. It is a statement about
// ARITHMETIC, not threads: the race build inhibits FMA contraction, and the EE FPU and the
// vector units fold multiply-adds, so a float result can differ in its last bits from the
// one the pinned hashes were taken on — even single-threaded, where no race can exist. The
// frame gate skips under -race for that reason; see bench_test.go.
const raceBuild = true
