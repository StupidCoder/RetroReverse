//go:build race

package n3ds

// raceBuild is true when the race detector is compiled in. It is not a statement
// about threads: it is a statement about ARITHMETIC. The race build inhibits FMA
// contraction, so the shader's MAD rounds differently and the frame differs in its
// last bits from the one the pinned hashes were taken on — even single-threaded. See
// parallel_test.go.
const raceBuild = true
